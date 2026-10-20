package tui

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

func TestQueueDropsOldestAndNeverCoalesces(t *testing.T) {
	q := newQueue[int](3)
	dropped := 0
	for i := 1; i <= 7; i++ {
		if q.push(i) {
			dropped++
		}
	}
	if dropped != 4 {
		t.Fatalf("dropped %d pushes, want 4 (cap 3, 7 pushed)", dropped)
	}
	var got []int
	for len(q.ch) > 0 {
		got = append(got, <-q.ch)
	}
	if len(got) != 3 || got[0] != 5 || got[1] != 6 || got[2] != 7 {
		t.Fatalf("queue kept %v, want the newest three untouched [5 6 7]", got)
	}
}

func TestQueueDrainEmpties(t *testing.T) {
	q := newQueue[int](4)
	q.push(1)
	q.push(2)
	q.drain()
	if len(q.ch) != 0 {
		t.Fatal("drain left items")
	}
}

// ---- feed against a real hub ----

type sinkRig struct {
	gate chan struct{} // closed to release a stalled consumer
	out  chan tea.Msg
}

func newSinkRig(stalled bool) *sinkRig {
	s := &sinkRig{gate: make(chan struct{}), out: make(chan tea.Msg, 1<<15)}
	if !stalled {
		close(s.gate)
	}
	return s
}

func (s *sinkRig) sink(msg tea.Msg) {
	<-s.gate // a stalled renderer: program.Send blocking
	s.out <- msg
}

func (s *sinkRig) release() { close(s.gate) }

func startFeed(t *testing.T, r *rig, queue, hubBuf int, s *sinkRig) *Feed {
	t.Helper()
	r.c.Close() // the rig's own TUI client: only one may attach
	f := NewFeed(FeedOptions{Dial: HubDialer(r.h, hubBuf), Sink: s.sink, QueueSize: queue})
	ctx, cancel := context.WithCancel(context.Background())
	if err := f.Start(ctx); err != nil {
		cancel()
		t.Fatalf("start: %v", err)
	}
	t.Cleanup(func() {
		select { // never leave a stalled sink behind Close
		case <-s.gate:
		default:
			close(s.gate)
		}
		cancel()
		f.Close()
	})
	return f
}

// converge applies sink output to a model until the entry "s" is complete with
// want as its text, or fails after testWait of no convergence.
func converge(t *testing.T, s *sinkRig, want string) Model {
	t.Helper()
	m := newModel(t, &fakeIngress{}, 80, 12)
	deadline := time.After(testWait)
	for {
		select {
		case msg := <-s.out:
			m, _ = step(m, msg)
			for _, e := range m.entries {
				if e.ID == "s" && e.Status == "complete" && e.Text == want {
					return m
				}
			}
		case <-deadline:
			t.Fatalf("model never converged on the canonical state; entries=%+v", m.entries)
		}
	}
}

func streamBurst(t *testing.T, r *rig, n int) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), testWait)
	defer cancel()
	for i := 0; i < n; i++ {
		// A stalled renderer must never make the hub's ingress stop draining:
		// each Send would time out if it did.
		if err := r.h.Send(ctx, delta("s", "x")); err != nil {
			t.Fatalf("hub blocked on a stalled TUI after %d events: %v", i, err)
		}
	}
	want := strings.Repeat("x", n)
	if err := r.h.Send(ctx, message("s", want)); err != nil {
		t.Fatal(err)
	}
	return want
}

func TestStalledRendererNeverBlocksHubAndStateConverges(t *testing.T) {
	r := newRig(t, hub.Config{})
	s := newSinkRig(true)
	f := startFeed(t, r, 8, 16, s)

	want := streamBurst(t, r, 3000)

	deadline := time.Now().Add(testWait)
	for f.Stats().Rehydrates == 0 {
		if time.Now().After(deadline) {
			t.Fatal("overflow never forced a re-hydrate")
		}
		time.Sleep(time.Millisecond)
	}
	s.release()
	m := converge(t, s, want)
	if len(m.entries) != 1 {
		t.Fatalf("converged state has %d entries, want 1: %+v", len(m.entries), m.entries)
	}
}

func TestOverflowDropsOldestAndTriggersRehydrate(t *testing.T) {
	r := newRig(t, hub.Config{})
	s := newSinkRig(true)
	f := startFeed(t, r, 4, 0, s)

	want := streamBurst(t, r, 200)

	deadline := time.Now().Add(testWait)
	for {
		st := f.Stats()
		if st.Dropped > 0 && st.Rehydrates > 0 {
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("stats after overflow: %+v, want drops and a re-hydrate", st)
		}
		time.Sleep(time.Millisecond)
	}
	s.release()
	converge(t, s, want)
}

func TestFeedDeliversInOrderWithoutOverflow(t *testing.T) {
	r := newRig(t, hub.Config{})
	s := newSinkRig(false)
	f := startFeed(t, r, 0, 0, s)
	r.send(message("a", "1"), message("b", "2"))
	var kinds []hub.Kind
	for len(kinds) < 3 {
		select {
		case msg := <-s.out:
			if fm, ok := msg.(FrameMsg); ok {
				kinds = append(kinds, fm.Frame.Kind())
			}
		case <-time.After(testWait):
			t.Fatalf("got %v", kinds)
		}
	}
	if kinds[0] != hub.KindSessionHydrate || kinds[1] != hub.KindAgentMessage || kinds[2] != hub.KindAgentMessage {
		t.Fatalf("order: %v", kinds)
	}
	if st := f.Stats(); st.Dropped != 0 || st.Rehydrates != 0 {
		t.Fatalf("unexpected overflow: %+v", st)
	}
}

func TestFeedReportsDisconnectWhenHubStops(t *testing.T) {
	r := newRig(t, hub.Config{})
	s := newSinkRig(false)
	startFeed(t, r, 0, 0, s)
	r.h.Stop()
	for {
		select {
		case msg := <-s.out:
			if _, ok := msg.(DisconnectedMsg); ok {
				return
			}
		case <-time.After(testWait):
			t.Fatal("no DisconnectedMsg after the hub stopped")
		}
	}
}

func TestFeedDoGoesThroughHubIngress(t *testing.T) {
	r := newRig(t, hub.Config{}) // no agent attached: the hub refuses prompts
	s := newSinkRig(false)
	f := startFeed(t, r, 0, 0, s)
	ctx, cancel := context.WithTimeout(context.Background(), testWait)
	defer cancel()
	res, err := f.Do(ctx, hub.OperatorAction{Type: hub.OpPrompt, Prompt: "hi"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Accepted || res.Reason != hub.ReasonAgentUnavailable {
		t.Fatalf("the hub's own verdict must come back unchanged, got %+v", res)
	}
}

func TestFeedDoWithoutConnection(t *testing.T) {
	f := NewFeed(FeedOptions{Dial: func(context.Context) (Conn, error) { return nil, errors.New("no") }, Sink: func(tea.Msg) {}})
	if _, err := f.Do(context.Background(), hub.OperatorAction{Type: hub.OpAbortTurn}); !errors.Is(err, ErrNotConnected) {
		t.Fatalf("err = %v, want ErrNotConnected", err)
	}
	if err := f.Start(context.Background()); err == nil {
		t.Fatal("Start must surface a failed first dial")
	}
}

func TestFeedClosesCleanlyAndStopsDelivering(t *testing.T) {
	r := newRig(t, hub.Config{})
	s := newSinkRig(false)
	f := startFeed(t, r, 0, 0, s)
	f.Close()
	f.Close() // idempotent
	// Close waited for both goroutines, so no sink call can be in flight or
	// start later: the count after Close is final.
	n := len(s.out)
	r.send(agent.Event{Kind: agent.KindTurnStart, TurnID: "t"})
	if got := len(s.out); got != n {
		t.Fatalf("feed delivered after Close: %d -> %d", n, got)
	}
}

func TestFeedCloseBeforeStartAndDoubleStart(t *testing.T) {
	var dials atomic.Int32
	dial := func(context.Context) (Conn, error) {
		dials.Add(1)
		return nil, errors.New("unused")
	}
	f := NewFeed(FeedOptions{Dial: dial, Sink: func(tea.Msg) {}})
	f.Close()
	if err := f.Start(context.Background()); !errors.Is(err, ErrFeedClosed) {
		t.Fatalf("Start after Close = %v, want ErrFeedClosed", err)
	}
	if dials.Load() != 0 {
		t.Fatal("Start after Close must not dial")
	}

	r := newRig(t, hub.Config{})
	s := newSinkRig(false)
	f = startFeed(t, r, 0, 0, s)
	if err := f.Start(context.Background()); !errors.Is(err, ErrFeedStarted) {
		t.Fatalf("second Start = %v, want ErrFeedStarted", err)
	}
}

func TestFeedReleasesHubSlotWhenContextCancelledWithoutClose(t *testing.T) {
	r := newRig(t, hub.Config{})
	r.c.Close()
	s := newSinkRig(false)
	f := NewFeed(FeedOptions{Dial: HubDialer(r.h, 0), Sink: s.sink})
	ctx, cancel := context.WithCancel(context.Background())
	if err := f.Start(ctx); err != nil {
		t.Fatal(err)
	}
	cancel() // no Close
	deadline := time.After(testWait)
	for {
		c, err := HubDialer(r.h, 0)(context.Background())
		if err == nil {
			c.Close()
			break
		}
		select {
		case <-deadline:
			t.Fatalf("the single TUI slot was never released: %v", err)
		case <-time.After(time.Millisecond):
		}
	}
	f.Close()
}

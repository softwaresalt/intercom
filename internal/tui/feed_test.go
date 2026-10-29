package tui

import (
	"context"
	"errors"
	"strings"
	"sync"
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

// countingDial wraps a Dial and counts how many connections it opened.
func countingDial(inner Dial, n *atomic.Int32) Dial {
	return func(ctx context.Context) (Conn, error) {
		n.Add(1)
		return inner(ctx)
	}
}

func TestOverflowRehydratesInPlaceWithoutAutoDenyingPendingPermission(t *testing.T) {
	r := newRig(t, hub.Config{PermissionTimeout: time.Hour})
	r.c.Close() // only one TUI may attach
	s := newSinkRig(true)
	var dials atomic.Int32
	f := NewFeed(FeedOptions{Dial: countingDial(HubDialer(r.h, 0), &dials), Sink: s.sink, QueueSize: 4})
	ctx, cancel := context.WithCancel(context.Background())
	if err := f.Start(ctx); err != nil {
		cancel()
		t.Fatalf("start: %v", err)
	}
	t.Cleanup(func() {
		select {
		case <-s.gate:
		default:
			close(s.gate)
		}
		cancel()
		f.Close()
	})

	a := newPermAsk(agent.PermissionKindShell, "ls")
	a.post(t, r.h)
	want := streamBurst(t, r, 200)

	deadline := time.Now().Add(testWait)
	for f.Stats().Rehydrates == 0 {
		if time.Now().After(deadline) {
			t.Fatal("overflow never forced a re-hydrate")
		}
		time.Sleep(time.Millisecond)
	}
	if n := dials.Load(); n != 1 {
		t.Fatalf("dialled %d times, want 1: overflow must rehydrate in place, not close and redial", n)
	}
	sctx, scancel := context.WithTimeout(context.Background(), testWait)
	defer scancel()
	st, err := r.h.Stats(sctx)
	if err != nil {
		t.Fatal(err)
	}
	if st.PendingPermissions != 1 {
		t.Fatalf("pending permissions = %d, want 1 kept alive", st.PendingPermissions)
	}
	select {
	case d := <-a.reply:
		t.Fatalf("pending permission was resolved (%v) by the overflow recovery", d.Kind)
	default:
	}
	s.release()
	converge(t, s, want)
}

// fakeConn is a scriptable Conn.
type fakeConn struct {
	frames chan hub.Frame
	done   chan struct{}
	once   sync.Once
}

func newFakeConn() *fakeConn {
	return &fakeConn{frames: make(chan hub.Frame, 4), done: make(chan struct{})}
}

func (c *fakeConn) Frames() <-chan hub.Frame { return c.frames }
func (c *fakeConn) Done() <-chan struct{}    { return c.done }
func (c *fakeConn) Do(context.Context, hub.OperatorAction) (hub.ActionResult, error) {
	return hub.ActionResult{Accepted: true}, nil
}
func (c *fakeConn) Rehydrate(context.Context) error { return nil }
func (c *fakeConn) Close()                          { c.once.Do(func() { close(c.done) }) }

func TestFeedRedialRetriesTransientFailureAndRecoversDroppedClient(t *testing.T) {
	first, second := newFakeConn(), newFakeConn()
	var dials atomic.Int32
	dial := func(context.Context) (Conn, error) {
		switch dials.Add(1) {
		case 1:
			return first, nil
		case 2, 3:
			return nil, errors.New("transient dial failure")
		default:
			return second, nil
		}
	}
	s := newSinkRig(false)
	f := NewFeed(FeedOptions{Dial: dial, Sink: s.sink, RedialRetries: 3, RedialBackoff: time.Millisecond})
	ctx, cancel := context.WithCancel(context.Background())
	if err := f.Start(ctx); err != nil {
		cancel()
		t.Fatalf("start: %v", err)
	}
	t.Cleanup(func() { cancel(); f.Close() })

	first.Close() // the hub dropped this client: no server.shutdown frame

	deadline := time.After(testWait)
	for resynced := false; !resynced; {
		select {
		case msg := <-s.out:
			switch msg.(type) {
			case ResyncMsg:
				resynced = true
			case DisconnectedMsg:
				t.Fatal("feed gave up on a dropped client instead of redialling")
			}
		case <-deadline:
			t.Fatalf("no resync after a dropped client; dials=%d", dials.Load())
		}
	}
	if n := dials.Load(); n != 4 {
		t.Fatalf("dials = %d, want 4 (initial + two transient failures + success)", n)
	}
	if st := f.Stats(); st.Rehydrates != 1 {
		t.Fatalf("rehydrates = %d, want 1", st.Rehydrates)
	}
	res, err := f.Do(context.Background(), hub.OperatorAction{Type: hub.OpAbortTurn})
	if err != nil || !res.Accepted {
		t.Fatalf("Do after recovery = %+v, %v", res, err)
	}
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

// blockingRehydrateConn is a fakeConn whose in-place Rehydrate blocks until
// its context ends or release is closed.
type blockingRehydrateConn struct {
	*fakeConn
	started chan struct{}
	release chan struct{}
	ctxErr  atomic.Value // error seen by Rehydrate, if its ctx ended
}

func (c *blockingRehydrateConn) Rehydrate(ctx context.Context) error {
	close(c.started)
	select {
	case <-ctx.Done():
		c.ctxErr.Store(ctx.Err())
		return ctx.Err()
	case <-c.release:
		return nil
	}
}

func TestFeedRehydrateCallerCancelDoesNotTearDownFeed(t *testing.T) {
	conn := &blockingRehydrateConn{fakeConn: newFakeConn(), started: make(chan struct{}), release: make(chan struct{})}
	var dials atomic.Int32
	dial := func(context.Context) (Conn, error) {
		dials.Add(1)
		return conn, nil
	}
	s := newSinkRig(false)
	f := NewFeed(FeedOptions{Dial: dial, Sink: s.sink, RedialRetries: -1})
	ctx, cancel := context.WithCancel(context.Background())
	if err := f.Start(ctx); err != nil {
		cancel()
		t.Fatalf("start: %v", err)
	}
	t.Cleanup(func() { cancel(); f.Close() })

	rctx, rcancel := context.WithCancel(context.Background())
	errc := make(chan error, 1)
	go func() { errc <- f.Rehydrate(rctx) }()
	select {
	case <-conn.started:
	case <-time.After(testWait):
		t.Fatal("rehydrate never reached the connection")
	}
	rcancel()
	if err := <-errc; !errors.Is(err, context.Canceled) {
		t.Fatalf("caller got %v, want context.Canceled", err)
	}
	close(conn.release) // recovery completes under the feed's own lifetime
	deadline := time.After(testWait)
	for f.Stats().Rehydrates != 1 {
		select {
		case <-deadline:
			t.Fatalf("rehydrate did not complete (ctx err seen: %v)", conn.ctxErr.Load())
		case <-time.After(time.Millisecond):
		}
	}
	select {
	case <-conn.Done():
		t.Fatal("a cancelled caller closed the live connection")
	default:
	}
	if _, err := f.Do(context.Background(), hub.OperatorAction{Type: hub.OpAbortTurn}); err != nil {
		t.Fatalf("feed unusable after a cancelled Rehydrate: %v", err)
	}
	if n := dials.Load(); n != 1 {
		t.Fatalf("dials = %d, want 1", n)
	}
}

func TestFeedShutdownFrameOnFramesPathIsNotRedialled(t *testing.T) {
	r := newRig(t, hub.Config{})
	r.frame() // hydrate
	ctx, cancel := context.WithTimeout(context.Background(), testWait)
	defer cancel()
	if err := r.h.AnnounceShutdown(ctx, "test"); err != nil {
		t.Fatal(err)
	}
	sd := r.frame()
	if sd.Kind() != hub.KindServerShutdown {
		t.Fatalf("got %s, want server.shutdown", sd.Kind())
	}

	conn := newFakeConn()
	var dials atomic.Int32
	dial := func(context.Context) (Conn, error) {
		dials.Add(1)
		return conn, nil
	}
	s := newSinkRig(false)
	f := NewFeed(FeedOptions{Dial: dial, Sink: s.sink, RedialRetries: 2, RedialBackoff: time.Millisecond})
	fctx, fcancel := context.WithCancel(context.Background())
	if err := f.Start(fctx); err != nil {
		fcancel()
		t.Fatalf("start: %v", err)
	}
	t.Cleanup(func() { fcancel(); f.Close() })

	conn.frames <- sd
	deadline := time.After(testWait)
	for got := false; !got; { // the Frames() path has forwarded it
		select {
		case msg := <-s.out:
			if fm, ok := msg.(FrameMsg); ok && fm.Frame.Kind() == hub.KindServerShutdown {
				got = true
			}
		case <-deadline:
			t.Fatal("shutdown frame not delivered")
		}
	}
	conn.Close() // the hub goes away after announcing shutdown
	for {
		select {
		case msg := <-s.out:
			switch msg.(type) {
			case DisconnectedMsg:
				if n := dials.Load(); n != 1 {
					t.Fatalf("dials = %d, want 1: intentional shutdown must not redial", n)
				}
				return
			case ResyncMsg:
				t.Fatal("redialled after an intentional shutdown")
			}
		case <-deadline:
			t.Fatal("no DisconnectedMsg after shutdown")
		}
	}
}

func TestFeedLateBroadcastsAfterShutdownDoNotEvictShutdownFrame(t *testing.T) {
	r := newRig(t, hub.Config{})
	r.frame() // hydrate
	ctx, cancel := context.WithTimeout(context.Background(), testWait)
	defer cancel()
	if err := r.h.AnnounceShutdown(ctx, "test"); err != nil {
		t.Fatal(err)
	}
	sd := r.frame()
	if sd.Kind() != hub.KindServerShutdown {
		t.Fatalf("got %s, want server.shutdown", sd.Kind())
	}
	r.send(message("a", "1"), message("b", "2"), message("c", "3"))
	late := []hub.Frame{r.frame(), r.frame(), r.frame()}

	conn := newFakeConn()
	var dials atomic.Int32
	dial := func(context.Context) (Conn, error) {
		dials.Add(1)
		return conn, nil
	}
	s := newSinkRig(true) // stalled: the queue is the only buffer
	f := NewFeed(FeedOptions{Dial: dial, Sink: s.sink, QueueSize: 2, RedialRetries: 1, RedialBackoff: time.Millisecond})
	fctx, fcancel := context.WithCancel(context.Background())
	if err := f.Start(fctx); err != nil {
		fcancel()
		t.Fatalf("start: %v", err)
	}
	t.Cleanup(func() {
		select {
		case <-s.gate:
		default:
			close(s.gate)
		}
		fcancel()
		f.Close()
	})

	conn.frames <- sd
	for _, fr := range late {
		select {
		case conn.frames <- fr:
		case <-time.After(testWait):
			t.Fatal("could not queue late frame")
		}
	}
	time.Sleep(50 * time.Millisecond) // let the pump consume what it will
	s.release()

	deadline := time.After(testWait)
	gotShutdown := false
	for {
		select {
		case msg := <-s.out:
			switch m := msg.(type) {
			case FrameMsg:
				if m.Frame.Kind() == hub.KindServerShutdown {
					gotShutdown = true
				}
			case ResyncMsg:
				t.Fatal("late broadcasts after shutdown triggered overflow recovery")
			case DisconnectedMsg:
				if !gotShutdown {
					t.Fatal("shutdown frame was evicted before delivery")
				}
				if n := dials.Load(); n != 1 || f.Stats().Rehydrates != 0 {
					t.Fatalf("dials=%d rehydrates=%d, want 1 and 0", n, f.Stats().Rehydrates)
				}
				return
			}
		case <-deadline:
			t.Fatal("no DisconnectedMsg after shutdown")
		}
	}
}

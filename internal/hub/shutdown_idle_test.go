package hub

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

// ---- 067.007-T: shutdown steps 2-4 (abort, positive idle, stop) ----

func idleEvent() agent.Event { return agent.Event{Kind: agent.KindIdle, Aborted: true} }
func turnStart() agent.Event { return agent.Event{Kind: agent.KindTurnStart, TurnID: "t1"} }

func waitIdleAsync(h *Hub, ctx context.Context) <-chan error {
	res := make(chan error, 1)
	go func() { res <- h.WaitIdle(ctx) }()
	return res
}

func TestWaitIdleReturnsAtOnceWhenTheSessionNeverRan(t *testing.T) {
	h := startHub(t, Config{})
	ctx, cancel := context.WithTimeout(context.Background(), testWait)
	defer cancel()
	if err := h.WaitIdle(ctx); err != nil {
		t.Fatalf("WaitIdle = %v", err)
	}
}

func TestWaitIdleBlocksWhileATurnIsActiveUntilAnIdleEvent(t *testing.T) {
	h := startHub(t, Config{})
	send(t, h, turnStart())
	settle(t, h, 1)

	ctx, cancel := context.WithTimeout(context.Background(), testWait)
	defer cancel()
	res := waitIdleAsync(h, ctx)
	stats(t, h) // the waiter is registered by now
	select {
	case err := <-res:
		t.Fatalf("WaitIdle returned (%v) while a turn was active", err)
	case <-time.After(50 * time.Millisecond):
	}
	send(t, h, idleEvent())
	select {
	case err := <-res:
		if err != nil {
			t.Fatalf("WaitIdle = %v", err)
		}
	case <-time.After(testWait):
		t.Fatal("WaitIdle did not return after the idle event")
	}
}

// The positive-idle rule (design §3.1 step 3): a flag read while events are
// still queued is stale. Idle holds only when the last observed state is idle
// AND ingress has been drained. White-box: the hub is driven step by step.
func TestWaitIdleDoesNotTrustAStaleFlagWhileEventsAreQueued(t *testing.T) {
	h, err := New(Config{SessionID: "s", IngressBuffer: 4})
	if err != nil {
		t.Fatal(err)
	}
	reply := make(chan ctlResp, 1)
	h.ingress <- turnStart() // queued, not yet folded: the flag still says idle
	h.handleCtl(&ctlReq{op: ctlIdle, reply: reply})
	h.releaseIdleWaiters()
	select {
	case <-reply:
		t.Fatal("idle was reported while a turn.start sat in ingress")
	default:
	}
	h.handleEvent(<-h.ingress) // fold turn.start
	h.releaseIdleWaiters()
	select {
	case <-reply:
		t.Fatal("idle was reported during an active turn")
	default:
	}
	h.handleEvent(idleEvent())
	h.releaseIdleWaiters()
	select {
	case r := <-reply:
		if r.err != nil {
			t.Fatal(r.err)
		}
	default:
		t.Fatal("idle was not reported after the idle event with ingress drained")
	}
}

func TestWaitIdleHonoursItsContextAndHubStop(t *testing.T) {
	h := startHub(t, Config{})
	send(t, h, turnStart())
	settle(t, h, 1)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if err := h.WaitIdle(ctx); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("WaitIdle = %v, want the deadline", err)
	}
	res := waitIdleAsync(h, context.Background())
	stats(t, h)
	h.Stop()
	select {
	case err := <-res:
		if !errors.Is(err, ErrStopped) {
			t.Fatalf("WaitIdle after stop = %v", err)
		}
	case <-time.After(testWait):
		t.Fatal("WaitIdle hung past hub stop")
	}
}

// ---- executor: the agent stage ----

func TestAgentStageAbortsWaitsForPositiveIdleThenStops(t *testing.T) {
	r := newShutdownRig(t, Config{})
	send(t, r.h, turnStart())
	settle(t, r.h, 1)

	idleRelease := make(chan struct{})
	r.ag.abortHook = func() {
		go func() { // the SDK answers an abort with an idle event, asynchronously
			<-idleRelease
			send(t, r.h, idleEvent())
		}()
	}
	s := r.newShutdown(t, func(c *ShutdownConfig) { c.Timeouts.Idle = 10 * time.Second })
	s.Start(context.Background())

	waitUntil(t, "Abort", func() bool { return slices.Contains(r.rec.events(), "agent.abort") })
	// Stop must not run on a stale flag: the turn is still active.
	time.Sleep(100 * time.Millisecond)
	if n := r.ag.stops.Load(); n != 0 {
		t.Fatalf("Stop ran %d times before the positive idle signal", n)
	}
	close(idleRelease)
	waitShutdown(t, s)
	if err := s.Err(); err != nil {
		t.Fatalf("shutdown error: %v", err)
	}
	ev := r.rec.events()
	a, i, st, sd := indexOf(t, ev, "agent.abort"), indexOf(t, ev, "frame:agent.idle"), indexOf(t, ev, "agent.stop"), indexOf(t, ev, "frame:server.shutdown")
	if st < 0 || a < 0 || i < 0 || sd < 0 || sd >= a || a >= i || i >= st {
		t.Fatalf("order must be server.shutdown < abort < idle < stop, got %v", ev)
	}
}

func TestIdleTimeoutIsAggregatedAndStopStillRuns(t *testing.T) {
	r := newShutdownRig(t, Config{})
	send(t, r.h, turnStart())
	settle(t, r.h, 1)
	s := r.newShutdown(t, func(c *ShutdownConfig) { c.Timeouts.Idle = 50 * time.Millisecond }) // never goes idle
	s.Start(context.Background())
	waitShutdown(t, s)
	if !errors.Is(s.Err(), context.DeadlineExceeded) {
		t.Fatalf("Err() = %v, want the idle deadline", s.Err())
	}
	if n := r.ag.stops.Load(); n != 1 {
		t.Fatalf("Stop ran %d times, want 1 (the runtime must still be torn down)", n)
	}
}

func TestAbortFailureIsAggregatedSkipsTheIdleWaitAndStopStillRuns(t *testing.T) {
	r := newShutdownRig(t, Config{})
	send(t, r.h, turnStart())
	settle(t, r.h, 1)
	boom := errors.New("abort boom")
	r.ag.abortErr = boom
	start := time.Now()
	s := r.newShutdown(t, func(c *ShutdownConfig) { c.Timeouts.Idle = 10 * time.Second })
	s.Start(context.Background())
	waitShutdown(t, s)
	if !errors.Is(s.Err(), boom) {
		t.Fatalf("Err() = %v lacks the abort error", s.Err())
	}
	if time.Since(start) > 5*time.Second {
		t.Fatal("a failed abort still waited out the idle deadline")
	}
	if n := r.ag.stops.Load(); n != 1 {
		t.Fatalf("Stop ran %d times", n)
	}
}

func TestAbortOfAnAbsentSessionIsNotAnError(t *testing.T) {
	r := newShutdownRig(t, Config{})
	r.ag.abortErr = agent.ErrNotRunning
	s := r.newShutdown(t, nil)
	s.Start(context.Background())
	waitShutdown(t, s)
	if err := s.Err(); err != nil {
		t.Fatalf("Err() = %v: nothing to abort is not a failure", err)
	}
}

func TestStopThatIgnoresItsContextIsBoundedByTheStopDeadline(t *testing.T) {
	r := newShutdownRig(t, Config{})
	r.ag.stopBlock = make(chan struct{})
	r.ag.stopIgnoresCtx = true
	t.Cleanup(func() { close(r.ag.stopBlock) })
	s := r.newShutdown(t, func(c *ShutdownConfig) { c.Timeouts.Stop = 50 * time.Millisecond })
	s.Start(context.Background())
	waitShutdown(t, s)
	if !errors.Is(s.Err(), context.DeadlineExceeded) {
		t.Fatalf("Err() = %v, want the stop deadline", s.Err())
	}
	select {
	case <-r.h.Done():
	default:
		t.Fatal("the hub must still be stopped last")
	}
}

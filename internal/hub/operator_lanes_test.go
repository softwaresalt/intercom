package hub

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

// ---- 074.003: aborted-turn permission asks ----

// drainFrames returns every frame currently queued to c without blocking.
func drainFrames(c *Client) []Frame {
	var out []Frame
	for {
		select {
		case f := <-c.send:
			out = append(out, f)
		default:
			return out
		}
	}
}

func turnAsk(turnID string) *testAsk {
	a := newTestAsk(agent.PermissionKindShell, "cmd")
	a.ask.Request.TurnID = turnID
	return a
}

func replyKind(t *testing.T, a *testAsk) agent.DecisionKind {
	t.Helper()
	select {
	case d := <-a.reply:
		return d.Kind
	default:
		t.Fatal("ask was not answered")
		return 0
	}
}

func TestAbortedIdleEventSettlesDecidedResolvesPendingAndRejectsLateAsks(t *testing.T) {
	h := idleHub(t, Config{})
	c := newClient(1, ClientTUI, 128)
	h.register(c)
	h.handleEvent(agent.Event{Kind: agent.KindTurnStart, TurnID: "t1"})

	// (a) one decided-but-unacknowledged and one pending ask of the turn.
	decided, pending := turnAsk("t1"), turnAsk("t1")
	h.intakeAsk(decided.ask)
	h.intakeAsk(pending.ask)
	if h.perm.len() != 2 {
		t.Fatalf("setup: %d registered", h.perm.len())
	}
	h.decide(h.perm.order[0], ActorLocalTUI, allowOnce, "")
	<-decided.reply // the allow sits in the handler's reply buffer; the handler has not delivered

	// (b) asks buffered in h.asks: one of the aborted turn, one of the next.
	buffered, next := turnAsk("t1"), turnAsk("t2")
	h.asks <- buffered.ask
	h.asks <- next.ask

	h.handleEvent(agent.Event{Kind: agent.KindIdle, TurnID: "t1", Aborted: true})

	if _, ok := decided.outcome.Delivered(); ok {
		t.Fatal("decided ask delivered after abort")
	}
	if replyKind(t, pending) != agent.DecisionUnavailable {
		t.Fatal("pending ask of the aborted turn must resolve to unavailable")
	}
	if e := h.perm.order[0]; e.ask.Request.TurnID != "t1" || e.state != permDecided || e.reason != permReasonTurnAborted {
		t.Fatalf("pending entry = %+v, want decided with %s", e, permReasonTurnAborted)
	}
	if replyKind(t, buffered) != agent.DecisionUnavailable {
		t.Fatal("buffered ask of the aborted turn must be answered unavailable")
	}
	if len(h.asks) != 0 {
		t.Fatalf("%d asks left in the buffer", len(h.asks))
	}
	// The next turn's buffered ask was passed through intake: registered.
	if h.perm.len() != 2 { // pending (awaiting ack) + next
		t.Fatalf("registry has %d entries, want 2", h.perm.len())
	}
	var sawAborted, sawRequested bool
	for _, f := range drainFrames(c) {
		switch f.Kind() {
		case KindPermissionAborted:
			sawAborted = sawAborted || contains(f, permReasonTurnAborted)
		case KindPermissionRequested:
			sawRequested = true
		}
	}
	if !sawAborted || !sawRequested {
		t.Fatalf("aborted=%v requested=%v, want both broadcast", sawAborted, sawRequested)
	}

	// (c) idle-before-ask: a late ask carrying the aborted turn's id.
	before := h.perm.len()
	late := turnAsk("t1")
	h.intakeAsk(late.ask)
	if replyKind(t, late) != agent.DecisionUnavailable || h.perm.len() != before {
		t.Fatalf("late aborted-turn ask registered (len %d, was %d)", h.perm.len(), before)
	}
}

func TestAbortedIdleThenNextTurnAskBeforeTurnStartIsAccepted(t *testing.T) {
	h := idleHub(t, Config{})
	c := newClient(1, ClientTUI, 128)
	h.register(c)
	h.handleEvent(agent.Event{Kind: agent.KindTurnStart, TurnID: "t1"})
	h.handleEvent(agent.Event{Kind: agent.KindIdle, TurnID: "t1", Aborted: true})
	drainFrames(c)

	// (d) the next turn's ask is dequeued before its TurnStart.
	a := turnAsk("t2")
	h.intakeAsk(a.ask)
	if h.perm.len() != 1 {
		t.Fatalf("next-turn ask not registered (len %d)", h.perm.len())
	}
	var requested bool
	for _, f := range drainFrames(c) {
		requested = requested || f.Kind() == KindPermissionRequested
	}
	if !requested {
		t.Fatal("next-turn ask did not broadcast permission.requested")
	}
	h.handleEvent(agent.Event{Kind: agent.KindTurnStart, TurnID: "t2"})
	if h.activeTurnID != "t2" {
		t.Fatalf("activeTurnID = %q", h.activeTurnID)
	}
}

func contains(f Frame, s string) bool { return strings.Contains(string(f.Bytes()), s) }

// ---- review cycle 2 N3: lanes context and agentIdle bookkeeping ----

// Step 0 cancels the lanes' context SYNCHRONOUSLY: the instant BeginShutdown
// returns, an in-flight Send's context is already done, so a prompt cannot be
// handed to the SDK with a live context after step 0 (an async watcher would
// leave a window).
func TestBeginShutdownCancelsTheLanesContextBeforeItReturns(t *testing.T) {
	for i := 0; i < 50; i++ {
		ag := newFakeAgent()
		ag.block = make(chan struct{})
		h := startHub(t, Config{Agent: ag, OperatorActionTimeout: time.Minute})
		tui := connectRole(t, h, ClientTUI, 0)
		if res := do(t, tui, prompt("in-flight")); !res.Accepted {
			t.Fatalf("prompt = %+v", res)
		}
		awaitPrompt(t, ag)
		var sendCtx context.Context
		select {
		case sendCtx = <-ag.sendCtx:
		case <-time.After(testWait):
			t.Fatal("Send never captured its context")
		}
		ctx, cancel := context.WithTimeout(context.Background(), testWait)
		err := h.BeginShutdown(ctx)
		cancel()
		if err != nil {
			t.Fatalf("BeginShutdown: %v", err)
		}
		select {
		case <-sendCtx.Done():
		default:
			t.Fatalf("iteration %d: the in-flight Send's context was still live after step 0 returned", i)
		}
	}
}

// A Send that returns an error after it was invoked proves nothing about
// delivery (the runtime may have started a turn before the call timed out), so
// the prompt keeps the session busy: only the bounded idle wait covers it.
func TestFailedPromptSendKeepsTheSessionBusy(t *testing.T) {
	ag := newFakeAgent()
	ag.sendErr = errors.New("sdk send failed")
	h := startHub(t, Config{Agent: ag})
	tui := connectRole(t, h, ClientTUI, 0)
	if res := do(t, tui, prompt("go")); !res.Accepted {
		t.Fatalf("prompt = %+v", res)
	}
	awaitPrompt(t, ag)
	stats(t, h)
	time.Sleep(50 * time.Millisecond) // let any failure report reach the hub
	short, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if err := h.WaitIdle(short); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("WaitIdle = %v: a delivery-uncertain failed Send must not release the idle wait", err)
	}
}

// lateFailAgent fails its first Send only after gate is closed (a call that
// times out after the runtime already started the turn); later Sends succeed.
type lateFailAgent struct {
	gate  chan struct{}
	sent  chan string
	mu    sync.Mutex
	calls int
}

func (a *lateFailAgent) Send(_ context.Context, p string) error {
	a.mu.Lock()
	a.calls++
	first := a.calls == 1
	a.mu.Unlock()
	a.sent <- p
	if first {
		<-a.gate
		return errors.New("send timed out after the turn started")
	}
	return nil
}

func (a *lateFailAgent) Abort(context.Context) error { return nil }

// Review: A's turn.start consumes A's pending count, B is accepted, and then
// A's Send reports a late failure. The failure must not consume B's count, or
// A's idle would release WaitIdle before B's turn.start.
func TestLateSendFailureOfAStartedPromptDoesNotConsumeALaterPromptsCount(t *testing.T) {
	ag := &lateFailAgent{gate: make(chan struct{}), sent: make(chan string, 8)}
	h := startHub(t, Config{Agent: ag, OperatorActionTimeout: time.Minute})
	tui := connectRole(t, h, ClientTUI, 0)
	if res := do(t, tui, prompt("A")); !res.Accepted {
		t.Fatalf("prompt A = %+v", res)
	}
	<-ag.sent
	send(t, h, turnStart()) // A's turn starts (consumes A's count)
	if res := do(t, tui, prompt("B")); !res.Accepted {
		t.Fatalf("prompt B = %+v", res)
	}
	close(ag.gate) // A's Send now reports failure, late
	<-ag.sent      // the lane moved on to B: A's failure has been reported
	stats(t, h)
	time.Sleep(50 * time.Millisecond) // let the report reach the hub
	send(t, h, idleEvent())           // A ends; B's turn.start has not arrived
	short, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	err := h.WaitIdle(short)
	cancel()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("WaitIdle = %v: A's late failure consumed B's count and A's idle released the wait", err)
	}
	send(t, h, turnStart()) // B's turn
	send(t, h, idleEvent())
	ctx, cancel2 := context.WithTimeout(context.Background(), testWait)
	defer cancel2()
	if err := h.WaitIdle(ctx); err != nil {
		t.Fatalf("WaitIdle = %v after B's turn finished, want idle", err)
	}
}

// ... but while a turn IS running, a failed second prompt must not claim idle.
func TestFailedPromptSendDoesNotRestoreIdleWhileATurnRuns(t *testing.T) {
	ag := newFakeAgent()
	ag.sendErr = errors.New("sdk send failed")
	h := startHub(t, Config{Agent: ag})
	tui := connectRole(t, h, ClientTUI, 0)
	send(t, h, turnStart())
	if res := do(t, tui, prompt("go")); !res.Accepted {
		t.Fatalf("prompt = %+v", res)
	}
	awaitPrompt(t, ag)
	stats(t, h)
	time.Sleep(50 * time.Millisecond) // let the failure reach the hub
	short, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if err := h.WaitIdle(short); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("WaitIdle = %v: a turn is running, so a failed prompt must not release the idle wait", err)
	}
}

// Final review P3: promptsPending is a count, not a flag. Turn A is running and
// prompt B has been accepted; A's idle must not release the idle wait because
// B's turn has not started yet. Only B's turn.start (consuming the pending
// prompt) followed by B's idle makes the session idle.
func TestIdleOfARunningTurnDoesNotReleaseWaitIdleBeforeAPendingPromptsTurnStart(t *testing.T) {
	ag := newFakeAgent()
	h := startHub(t, Config{Agent: ag})
	tui := connectRole(t, h, ClientTUI, 0)
	send(t, h, turnStart()) // turn A
	if res := do(t, tui, prompt("B")); !res.Accepted {
		t.Fatalf("prompt = %+v", res)
	}
	awaitPrompt(t, ag)
	send(t, h, idleEvent()) // A ends; B's turn.start has not arrived
	short, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	err := h.WaitIdle(short)
	cancel()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("WaitIdle = %v: A's idle released the wait while B's turn was still pending", err)
	}
	send(t, h, turnStart()) // turn B
	send(t, h, idleEvent())
	ctx, cancel2 := context.WithTimeout(context.Background(), testWait)
	defer cancel2()
	if err := h.WaitIdle(ctx); err != nil {
		t.Fatalf("WaitIdle = %v after B's turn finished, want idle", err)
	}
}

// Prompts dropped at step 0 never ran, so their counts are retracted; the
// in-flight prompt already entered Agent.Send (delivery-uncertain) and keeps
// its count. So the session stays busy until that prompt's turn starts and
// ends, and then it is idle: the two dropped prompts left no stale count.
func TestStep0DroppedPromptsAreRetractedButTheInFlightPromptStaysCounted(t *testing.T) {
	ag := newFakeAgent()
	ag.block = make(chan struct{})
	h := startHub(t, Config{Agent: ag, OperatorActionTimeout: time.Minute})
	tui := connectRole(t, h, ClientTUI, 0)
	for _, p := range []string{"in-flight", "queued-1", "queued-2"} {
		if res := do(t, tui, prompt(p)); !res.Accepted {
			t.Fatalf("prompt %q = %+v", p, res)
		}
	}
	awaitPrompt(t, ag)
	ctx, cancel := context.WithTimeout(context.Background(), testWait)
	defer cancel()
	if err := h.BeginShutdown(ctx); err != nil {
		t.Fatalf("BeginShutdown: %v", err)
	}
	short, cancel2 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	err := h.WaitIdle(short)
	cancel2()
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("WaitIdle = %v: the in-flight prompt may have started a turn, so the session must stay busy", err)
	}
	send(t, h, turnStart()) // the in-flight prompt's turn
	send(t, h, idleEvent())
	if err := h.WaitIdle(ctx); err != nil {
		t.Fatalf("WaitIdle = %v: the dropped prompts' counts were not retracted", err)
	}
}

// A dropped prompt must not clear the busy state while a turn is running.
func TestStep0DroppedPromptsDoNotRestoreIdleWhileATurnRuns(t *testing.T) {
	ag := newFakeAgent()
	ag.block = make(chan struct{})
	h := startHub(t, Config{Agent: ag, OperatorActionTimeout: time.Minute})
	tui := connectRole(t, h, ClientTUI, 0)
	send(t, h, turnStart())
	for _, p := range []string{"in-flight", "queued-1"} {
		do(t, tui, prompt(p))
	}
	awaitPrompt(t, ag)
	ctx, cancel := context.WithTimeout(context.Background(), testWait)
	defer cancel()
	if err := h.BeginShutdown(ctx); err != nil {
		t.Fatalf("BeginShutdown: %v", err)
	}
	// stopLanes (and its retraction) ran on the hub goroutine before
	// BeginShutdown returned, so no settling delay is needed.
	short, cancel2 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel2()
	if err := h.WaitIdle(short); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("WaitIdle = %v: a turn is running, so dropping prompts must not report idle", err)
	}
}

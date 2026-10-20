package hub

import (
	"context"
	"errors"
	"testing"
	"time"
)

// ---- review finding B: operator lanes versus shutdown step 0 ----

// Step 0 stops the operator lanes: the in-flight Send is cancelled and nothing
// still queued reaches the agent, so no operator prompt can start a turn after
// the shutdown abort (or while the runtime is being stopped).
func TestBeginShutdownCancelsInFlightSendAndDropsQueuedPrompts(t *testing.T) {
	ag := newFakeAgent()
	ag.block = make(chan struct{}) // Send hangs until its context ends
	h := startHub(t, Config{Agent: ag, OperatorActionTimeout: time.Minute})
	tui := connectRole(t, h, ClientTUI, 0)

	for _, p := range []string{"in-flight", "queued-1", "queued-2"} {
		if res := do(t, tui, prompt(p)); !res.Accepted {
			t.Fatalf("prompt %q = %+v", p, res)
		}
	}
	if got := awaitPrompt(t, ag); got != "in-flight" {
		t.Fatalf("first prompt = %q", got)
	}

	ctx, cancel := context.WithTimeout(context.Background(), testWait)
	defer cancel()
	if err := h.BeginShutdown(ctx); err != nil {
		t.Fatalf("BeginShutdown: %v", err)
	}
	select {
	case err := <-ag.sendCtxE:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("in-flight Send ended with %v, want context.Canceled", err)
		}
	case <-time.After(testWait):
		t.Fatal("the in-flight Send was not cancelled by shutdown step 0")
	}
	stats(t, h)
	time.Sleep(50 * time.Millisecond) // a lane that kept draining would have taken queued-1 by now
	if got := ag.promptsSeen(); len(got) != 1 {
		t.Fatalf("agent saw %v: queued prompts must not be delivered after step 0", got)
	}
}

func TestBeginShutdownDropsAQueuedAbort(t *testing.T) {
	ag := newFakeAgent()
	ag.abortBlock = make(chan struct{})
	h := startHub(t, Config{Agent: ag, OperatorActionTimeout: time.Minute})
	tui := connectRole(t, h, ClientTUI, 0)

	do(t, tui, OperatorAction{Type: OpAbortTurn})
	awaitAbort(t, ag)                             // the lane is inside Abort
	do(t, tui, OperatorAction{Type: OpAbortTurn}) // queued behind it

	ctx, cancel := context.WithTimeout(context.Background(), testWait)
	defer cancel()
	if err := h.BeginShutdown(ctx); err != nil {
		t.Fatalf("BeginShutdown: %v", err)
	}
	stats(t, h)
	close(ag.abortBlock) // let the in-flight abort return so the lane would take the queued one
	time.Sleep(50 * time.Millisecond)
	ag.mu.Lock()
	n := ag.aborts
	ag.mu.Unlock()
	if n != 1 {
		t.Fatalf("agent saw %d operator aborts, want the queued one dropped", n)
	}
}

// A dispatched prompt means a turn is about to start: until the hub sees its
// turn.start (and then idle) the session is not positively idle, so the
// shutdown's idle wait must not be released by the stale pre-prompt flag.
func TestWaitIdleIsNotReleasedByAStaleFlagAfterAPromptIsDispatched(t *testing.T) {
	ag := newFakeAgent()
	h := startHub(t, Config{Agent: ag})
	tui := connectRole(t, h, ClientTUI, 0)
	if res := do(t, tui, prompt("go")); !res.Accepted {
		t.Fatalf("prompt = %+v", res)
	}
	awaitPrompt(t, ag)

	short, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if err := h.WaitIdle(short); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("WaitIdle = %v, want it to wait: a prompt was dispatched and no idle has been seen", err)
	}
	send(t, h, turnStart())
	send(t, h, idleEvent())
	ctx, cancel2 := context.WithTimeout(context.Background(), testWait)
	defer cancel2()
	if err := h.WaitIdle(ctx); err != nil {
		t.Fatalf("WaitIdle after turn.start and idle = %v", err)
	}
}

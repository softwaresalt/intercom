package hub

import (
	"context"
	"errors"
	"testing"
	"time"
)

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

// A Send that fails while no turn is known to be running means the prompt
// started nothing: the session is idle again, so the shutdown idle wait must
// not burn its whole deadline on the stale "prompt dispatched" flag.
func TestFailedPromptSendRestoresIdleWhenNoTurnStarted(t *testing.T) {
	ag := newFakeAgent()
	ag.sendErr = errors.New("sdk send failed")
	h := startHub(t, Config{Agent: ag})
	tui := connectRole(t, h, ClientTUI, 0)
	if res := do(t, tui, prompt("go")); !res.Accepted {
		t.Fatalf("prompt = %+v", res)
	}
	awaitPrompt(t, ag)
	ctx, cancel := context.WithTimeout(context.Background(), testWait)
	defer cancel()
	if err := h.WaitIdle(ctx); err != nil {
		t.Fatalf("WaitIdle = %v: a failed Send must restore the idle flag when no turn started", err)
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

// Prompts dropped at step 0 never ran: with no turn known running the session
// is idle, so the executor's idle wait is satisfied by a positive state rather
// than a stale "prompt dispatched" flag.
func TestStep0DroppedPromptsRestoreIdleWhenNoTurnRuns(t *testing.T) {
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
	if err := h.WaitIdle(ctx); err != nil {
		t.Fatalf("WaitIdle = %v: dropped prompts must not leave the session marked busy", err)
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
	time.Sleep(50 * time.Millisecond)
	short, cancel2 := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel2()
	if err := h.WaitIdle(short); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("WaitIdle = %v: a turn is running, so dropping prompts must not report idle", err)
	}
}

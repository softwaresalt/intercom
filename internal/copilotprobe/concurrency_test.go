package copilotprobe

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	copilot "github.com/github/copilot-sdk/go"
	"github.com/github/copilot-sdk/go/rpc"
)

// TestSQBReentrancyAndSQCDispatchIndependence answers SQ-b (callback
// re-entrancy) and SQ-c (dispatch independence). Depends on B2 (harness) and
// B3 (shared DriveTurn fixture / event instrumentation pattern).
//
// SQ-b: session.On's callback is instrumented with an atomic in-flight
// counter (incremented on entry, held briefly, decremented on exit) across a
// token-streaming turn, recording the observed maximum concurrent callback
// count and the number of turns sampled.
//
// SQ-c: a PermissionHandlerFunc is held blocked (B2's harness, same
// technique as B4) while the turn streams, observing whether session events
// continue to arrive during the block -- i.e. whether permission dispatch
// is independent of event dispatch, or head-of-line blocks it.
func TestSQBReentrancyAndSQCDispatchIndependence(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), probeDeadline)
	defer cancel()

	blockCh := make(chan struct{})
	handlerEntered := make(chan struct{})
	var enteredOnce int32

	harness := NewPermissionHarness(func(n int, _ copilot.PermissionRequest, _ copilot.PermissionInvocation) rpc.PermissionDecision {
		if atomic.CompareAndSwapInt32(&enteredOnce, 0, 1) {
			close(handlerEntered)
		}
		<-blockCh
		return &rpc.PermissionDecisionApproveOnce{}
	})

	client := newProbeClient(t, ctx)
	session := newProbeSession(t, ctx, client, &copilot.SessionConfig{
		OnPermissionRequest: harness.Handler(),
	})

	var inFlight int32
	var maxConcurrent int32
	var callbackInvocations int32
	var eventsAfterBlock int32
	var blockedSignaled int32

	onEvent := func(event copilot.SessionEvent) {
		n := atomic.AddInt32(&inFlight, 1)
		atomic.AddInt32(&callbackInvocations, 1)
		for {
			cur := atomic.LoadInt32(&maxConcurrent)
			if n <= cur || atomic.CompareAndSwapInt32(&maxConcurrent, cur, n) {
				break
			}
		}
		// Hold briefly so genuinely concurrent SDK dispatch would be
		// observable as an overlapping in-flight count > 1.
		time.Sleep(15 * time.Millisecond)

		if atomic.LoadInt32(&blockedSignaled) == 1 {
			atomic.AddInt32(&eventsAfterBlock, 1)
		}

		atomic.AddInt32(&inFlight, -1)
	}

	// Watch for the handler-entered signal on a side goroutine and flip
	// blockedSignaled so subsequent event callbacks can be attributed to
	// the post-block window (SQ-c).
	go func() {
		select {
		case <-handlerEntered:
			atomic.StoreInt32(&blockedSignaled, 1)
		case <-ctx.Done():
		}
	}()

	prompt := "First write exactly two short sentences about the color blue. Then, in the same turn, use the shell tool to run exactly this command and nothing else: echo B5TEST"

	unsubscribe := session.On(onEvent)
	defer unsubscribe()

	if _, err := session.Send(ctx, copilot.MessageOptions{Prompt: prompt}); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	select {
	case <-handlerEntered:
		t.Log("permission handler is now blocked; observing whether session events continue to arrive")
	case <-time.After(60 * time.Second):
		t.Skip("permission handler never entered within 60s; recording UNPROVEN for SQ-b/SQ-c")
		return
	}

	// Observe for a bounded window while the handler remains blocked.
	preBlockEvents := atomic.LoadInt32(&callbackInvocations)
	time.Sleep(10 * time.Second)
	postWindowEvents := atomic.LoadInt32(&callbackInvocations)
	afterBlock := atomic.LoadInt32(&eventsAfterBlock)

	close(blockCh) // release the blocked handler so the client can shut down cleanly

	t.Logf("SQ-b evidence: total_callback_invocations=%d max_observed_concurrent=%d", atomic.LoadInt32(&callbackInvocations), atomic.LoadInt32(&maxConcurrent))
	if atomic.LoadInt32(&maxConcurrent) <= 1 {
		t.Log("SQ-b ANSWER: SERIALISED -- across every sampled callback invocation, the observed maximum concurrent in-flight count was 1; no overlapping invocation was observed in this sample.")
	} else {
		t.Logf("SQ-b ANSWER: NOT SERIALISED -- observed up to %d concurrent callback invocations.", atomic.LoadInt32(&maxConcurrent))
	}

	t.Logf("SQ-c evidence: events_observed_before_block_window=%d events_observed_after_10s_window=%d events_strictly_after_block_signal=%d", preBlockEvents, postWindowEvents, afterBlock)
	if postWindowEvents > preBlockEvents || afterBlock > 0 {
		t.Log("SQ-c ANSWER: YES -- session events continued to arrive while the PermissionHandlerFunc was blocked; permission dispatch is independent of event dispatch (no head-of-line blocking observed).")
	} else {
		t.Log("SQ-c ANSWER: NO -- no further session events were observed while the handler was blocked; evidence is consistent with head-of-line blocking behind the pending permission decision.")
	}
}

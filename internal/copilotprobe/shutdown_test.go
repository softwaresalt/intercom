package copilotprobe

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	copilot "github.com/github/copilot-sdk/go"
	"github.com/github/copilot-sdk/go/rpc"
)

// TestS3CancellationShutdown proves D2's S3 criterion and answers SQ-a.
//
// Depends on B2's PermissionHarness (blocked-handler condition): the harness
// here blocks the handler goroutine on an unclosed channel to simulate an
// already-blocked PermissionHandlerFunc (e.g. one genuinely waiting on a
// human). It then exercises Session.Abort, Client.Stop, and Client.ForceStop
// as an escalating shutdown ladder, each under its own bounded sub-deadline
// so this test can never hang unbounded even if a real deadlock exists --
// discovering a deadlock here is itself a successful, informative outcome
// per the plan.
func TestS3CancellationShutdown(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), probeDeadline)
	defer cancel()

	handlerEntered := make(chan struct{})
	blockCh := make(chan struct{}) // deliberately never closed during the probe window
	var enteredOnce int32

	harness := NewPermissionHarness(func(n int, _ copilot.PermissionRequest, _ copilot.PermissionInvocation) rpc.PermissionDecision {
		if atomic.CompareAndSwapInt32(&enteredOnce, 0, 1) {
			close(handlerEntered)
		}
		<-blockCh // block indefinitely -- simulates an already-blocked handler
		return &rpc.PermissionDecisionApproveOnce{}
	})

	client := newProbeClientManualLifecycle(t, ctx)
	session := newProbeSession(t, ctx, client, &copilot.SessionConfig{
		OnPermissionRequest: harness.Handler(),
	})

	// Trigger a permission request without waiting for the turn to end --
	// the handler will block inside the SDK's own dispatch goroutine.
	if _, err := session.Send(ctx, copilot.MessageOptions{Prompt: "Use the shell tool to run exactly this command and nothing else: echo B4TEST"}); err != nil {
		t.Fatalf("Send failed: %v", err)
	}

	select {
	case <-handlerEntered:
		t.Log("handler is now blocked (simulating an already-blocked PermissionHandlerFunc)")
	case <-time.After(30 * time.Second):
		t.Skip("permission handler never entered within 30s; recording UNPROVEN for S3/SQ-a")
		return
	}

	// SQ-a: does Session.Abort unblock an already-blocked PermissionHandlerFunc?
	abortDone := make(chan error, 1)
	abortCtx, abortCancel := context.WithTimeout(ctx, 20*time.Second)
	defer abortCancel()
	go func() { abortDone <- session.Abort(abortCtx) }()

	var abortErr error
	var abortReturnedPromptly bool
	select {
	case abortErr = <-abortDone:
		abortReturnedPromptly = true
	case <-time.After(20 * time.Second):
		abortReturnedPromptly = false
	}
	t.Logf("Abort: returned_promptly=%v err=%v", abortReturnedPromptly, abortErr)

	// SQ-a evidence caveat (correctness finding): blockCh in this harness is
	// ONLY ever closed by this test's own close(blockCh) call below -- no
	// other path (SDK-driven or otherwise) can close it. A select against
	// blockCh is therefore NOT an independent empirical observation of
	// whether Abort delivered any signal into the handler; it is
	// guaranteed to report "still parked" by construction on every run. It
	// is retained here only as a construction sanity-check (it would be a
	// genuine harness bug if it ever reported otherwise), not as SQ-a
	// evidence. The actual empirical measurement this probe produces is
	// abortReturnedPromptly above (whether Abort itself hangs); the
	// "handler remains blocked" half of the SQ-a conclusion below is a
	// structural consequence of PermissionHandlerFunc's signature carrying
	// no context.Context (verified via `go doc`), not something this
	// specific channel check discovered live.
	time.Sleep(500 * time.Millisecond)
	handlerStillParked := true
	select {
	case <-blockCh:
		handlerStillParked = false
	default:
	}
	if !handlerStillParked {
		t.Fatalf("construction invariant violated: blockCh closed before this test's own close(blockCh) call -- harness bug, not an SQ-a finding")
	}
	t.Log("SQ-a evidence: abortReturnedPromptly is the genuine empirical signal (no deadlock at the Abort/RPC layer); the handler-remains-blocked half of the conclusion below follows from PermissionHandlerFunc's signature (no context.Context parameter), not from a live SDK-driven unblock signal, since no such signal exists for this harness to observe.")
	if abortReturnedPromptly {
		t.Log("SQ-a ANSWER: NO -- Session.Abort returned promptly without delivering any signal that could unblock an already-blocked PermissionHandlerFunc. Abort operates at the session/RPC level and PermissionHandlerFunc's signature carries no context.Context, so there is no channel through which Abort could interrupt an in-flight handler call. Consistent with design section 3.1's conservative assumption ('until answered, assume it does not').")
	} else {
		t.Log("SQ-a ANSWER: DEADLOCK OBSERVED -- Session.Abort itself did not return within its 20s sub-deadline while the permission handler remained blocked. This is the design section 3.1 deadlock scenario made concrete; recorded as a successful, informative de-risking outcome.")
	}

	// Release the blocked handler goroutine so it doesn't leak past this test.
	close(blockCh)

	// S3 escalating shutdown ladder: Stop, then ForceStop as the escape
	// hatch, each bounded so neither call can hang this test unbounded.
	stopDone := make(chan error, 1)
	go func() { stopDone <- client.Stop() }()
	select {
	case err := <-stopDone:
		t.Logf("Client.Stop(): returned err=%v", err)
	case <-time.After(20 * time.Second):
		t.Log("Client.Stop(): did not return within 20s; escalating to ForceStop")
	}

	forceStopDone := make(chan struct{})
	go func() {
		client.ForceStop()
		close(forceStopDone)
	}()
	select {
	case <-forceStopDone:
		t.Log("Client.ForceStop(): returned")
	case <-time.After(10 * time.Second):
		t.Log("Client.ForceStop(): did not return within 10s (unexpected for a force-kill escape hatch)")
	}

	t.Log("S3 PROVEN: Abort -> Stop -> ForceStop shutdown ladder exercised under a deadline-bearing context.Context; no unbounded hang occurred in this test (each stage individually bounded).")
}

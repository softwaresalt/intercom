package hub

import (
	"context"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

// ---- review finding C: a decision is revoked unless the handler took it ----

func decidedEntry(t *testing.T, r *permRig, tui *Client) (string, *testAsk) {
	t.Helper()
	a := newTestAsk(agent.PermissionKindShell, "cmd")
	a.post(t, r.h)
	id := requestID(t, recvKind(t, tui, KindPermissionRequested))
	if res := resolveAs(t, r.h, tuiP(), id, allowOnce); !res.Accepted {
		t.Fatalf("resolve = %+v", res)
	}
	return id, a
}

// Ack deadline, decision still in the buffer: the handler never took it, so it
// must not be delivered later (an allow could reach the SDK after the surfaces
// were told "aborted"). The ask is revoked: the late handler's Deliver fails
// and it must return UserNotAvailable.
func TestAckTimeoutRevokesABufferedDecisionAndPublishesAborted(t *testing.T) {
	r := newPermRig(t, Config{PermissionTimeout: time.Hour, PermissionAckTimeout: 5 * time.Second})
	tui := connect(t, r.h, ClientTUI, 16)
	id, a := decidedEntry(t, r, tui)

	r.clock.Advance(5 * time.Second)
	r.fire(t)
	p := abortedPayload(t, recvKind(t, tui, KindPermissionAborted))
	if p.RequestID != id || p.Reason != "ack_timeout" {
		t.Fatalf("aborted = %+v", p)
	}
	late := a.consumeOnly(t) // a late handler still finds the stale allow
	if a.outcome.Deliver(late) {
		t.Fatalf("a late handler delivered %v after the surfaces were told aborted", late.Kind)
	}
}

// Ack deadline, handler delivered: the SDK got the decision, so the truthful
// outcome is resolved, not aborted.
func TestAckTimeoutAfterTheHandlerTookTheDecisionPublishesResolved(t *testing.T) {
	r := newPermRig(t, Config{PermissionTimeout: time.Hour, PermissionAckTimeout: 5 * time.Second})
	tui := connect(t, r.h, ClientTUI, 16)
	id, a := decidedEntry(t, r, tui)
	a.takeReply(t) // consumed and delivered, but the handler never returns

	r.clock.Advance(5 * time.Second)
	r.fire(t)
	p := resolvedPayload(t, recvKind(t, tui, KindPermissionResolved))
	if p.RequestID != id || p.Decision != "allow_once" || p.Actor != ActorLocalTUI {
		t.Fatalf("resolved = %+v", p)
	}
	noFrame(t, r.h, tui, "no permission.aborted may accompany a delivered decision")
}

// Review cycle 2 N2: the window. The handler has taken the allow off the reply
// buffer but has not yet recorded what it returns (the buffer is empty, the
// outcome is not recorded). The hub must NOT infer delivery from the empty
// buffer: it revokes, publishes aborted, and the handler's later Deliver fails
// so the SDK never receives the allow.
func TestAckTimeoutInTheConsumedButUndeliveredWindowRevokesAndAborts(t *testing.T) {
	r := newPermRig(t, Config{PermissionTimeout: time.Hour, PermissionAckTimeout: 5 * time.Second})
	tui := connect(t, r.h, ClientTUI, 16)
	id, a := decidedEntry(t, r, tui)
	taken := a.consumeOnly(t) // consumed; not yet delivered

	r.clock.Advance(5 * time.Second)
	r.fire(t)
	p := abortedPayload(t, recvKind(t, tui, KindPermissionAborted))
	if p.RequestID != id || p.Reason != "ack_timeout" {
		t.Fatalf("aborted = %+v", p)
	}
	if a.outcome.Deliver(taken) {
		t.Fatal("the handler delivered the allow after the hub published aborted: the SDK would get a decision the surfaces were told was abandoned")
	}
	noFrame(t, r.h, tui, "no permission.resolved may follow an abort")
}

func TestBeginShutdownInTheConsumedButUndeliveredWindowRevokesAndAborts(t *testing.T) {
	r := newPermRig(t, Config{PermissionTimeout: time.Hour})
	tui := connect(t, r.h, ClientTUI, 16)
	id, a := decidedEntry(t, r, tui)
	taken := a.consumeOnly(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := r.h.BeginShutdown(ctx); err != nil {
		t.Fatalf("BeginShutdown: %v", err)
	}
	p := abortedPayload(t, recvKind(t, tui, KindPermissionAborted))
	if p.RequestID != id || p.Reason != "shutdown" {
		t.Fatalf("aborted = %+v", p)
	}
	if a.outcome.Deliver(taken) {
		t.Fatal("an allow was delivered to the SDK after shutdown step 0 aborted it")
	}
}

// The handler downgrades an approved allow because the adapter stopped: the SDK
// got UserNotAvailable, not the operator's allow, so the surfaces must see
// aborted, never resolved(allow_once).
func TestHandlerDowngradeOfAnAllowIsPublishedAsAbortedNotResolved(t *testing.T) {
	r := newPermRig(t, Config{PermissionTimeout: time.Hour})
	tui := connect(t, r.h, ClientTUI, 16)
	id, a := decidedEntry(t, r, tui)
	a.consumeOnly(t)
	a.outcome.Deliver(agent.PermissionDecision{Kind: agent.DecisionUnavailable}) // what the handler returned
	close(a.resolved)
	p := abortedPayload(t, recvKind(t, tui, KindPermissionAborted))
	if p.RequestID != id || p.Reason != "handler_gone" {
		t.Fatalf("aborted = %+v", p)
	}
	noFrame(t, r.h, tui, "no permission.resolved may follow an abort")
}

// A hub-issued UserNotAvailable the handler returned unchanged (even on its
// stop path) is the truth: resolved with the hub's reason, never aborted.
func TestHubIssuedUnavailableDeliveredIsPublishedAsResolvedWithItsReason(t *testing.T) {
	r := newPermRig(t, Config{PermissionTimeout: time.Minute})
	tui := connect(t, r.h, ClientTUI, 16)
	a := newTestAsk(agent.PermissionKindShell, "cmd")
	a.post(t, r.h)
	id := requestID(t, recvKind(t, tui, KindPermissionRequested))
	r.clock.Advance(time.Minute)
	r.fire(t)
	a.consumeOnly(t)
	a.outcome.Deliver(agent.PermissionDecision{Kind: agent.DecisionUnavailable}) // the stop path returns unavailable too
	close(a.resolved)
	p := resolvedPayload(t, recvKind(t, tui, KindPermissionResolved))
	if p.RequestID != id || p.Decision != "unavailable" || p.Reason != "timeout" {
		t.Fatalf("resolved = %+v", p)
	}
}

// Closing Resolved without recording an outcome is a protocol violation:
// nothing proves the SDK got the decision, so the hub fails closed (aborted).
func TestAckWithoutARecordedOutcomeIsAbortedNotResolved(t *testing.T) {
	r := newPermRig(t, Config{PermissionTimeout: time.Hour})
	tui := connect(t, r.h, ClientTUI, 16)
	id, a := decidedEntry(t, r, tui)
	a.consumeOnly(t)
	close(a.resolved)
	p := abortedPayload(t, recvKind(t, tui, KindPermissionAborted))
	if p.RequestID != id || p.Reason != "handler_gone" {
		t.Fatalf("aborted = %+v", p)
	}
}

func TestAskWithoutAnOutcomeIsFailedClosed(t *testing.T) {
	r := newPermRig(t, Config{PermissionTimeout: time.Hour})
	tui := connect(t, r.h, ClientTUI, 16)
	a := newTestAsk(agent.PermissionKindShell, "cmd")
	a.ask.Outcome = nil
	a.post(t, r.h)
	if d := a.consumeOnly(t); d.Kind != agent.DecisionUnavailable {
		t.Fatalf("handler got %v, want unavailable", d.Kind)
	}
	if n := stats(t, r.h).PendingPermissions; n != 0 {
		t.Fatalf("pending = %d: a malformed ask must not be registered", n)
	}
	noFrame(t, r.h, tui, "a malformed ask is never broadcast")
}

func TestBeginShutdownRevokesADecidedButUnackedBufferedAllow(t *testing.T) {
	r := newPermRig(t, Config{PermissionTimeout: time.Hour})
	tui := connect(t, r.h, ClientTUI, 16)
	id, a := decidedEntry(t, r, tui)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := r.h.BeginShutdown(ctx); err != nil {
		t.Fatalf("BeginShutdown: %v: a decided entry must be settled by step 0, not wait for an ack", err)
	}
	late := a.consumeOnly(t)
	if a.outcome.Deliver(late) {
		t.Fatalf("handler delivered %v after step 0 (no allow after shutdown began)", late.Kind)
	}
	p := abortedPayload(t, recvKind(t, tui, KindPermissionAborted))
	if p.RequestID != id || p.Reason != "shutdown" {
		t.Fatalf("aborted = %+v", p)
	}
}

func TestBeginShutdownPublishesResolvedForADecisionTheHandlerAlreadyTook(t *testing.T) {
	r := newPermRig(t, Config{PermissionTimeout: time.Hour})
	tui := connect(t, r.h, ClientTUI, 16)
	id, a := decidedEntry(t, r, tui)
	a.takeReply(t)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := r.h.BeginShutdown(ctx); err != nil {
		t.Fatalf("BeginShutdown: %v", err)
	}
	p := resolvedPayload(t, recvKind(t, tui, KindPermissionResolved))
	if p.RequestID != id || p.Decision != "allow_once" {
		t.Fatalf("resolved = %+v", p)
	}
}

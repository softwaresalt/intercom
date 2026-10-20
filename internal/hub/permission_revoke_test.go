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
// were told "aborted"). The buffer is swapped for UserNotAvailable.
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
	if d := a.takeReply(t); d.Kind != agent.DecisionUnavailable {
		t.Fatalf("a late handler would receive %v, want the allow revoked to unavailable", d.Kind)
	}
}

// Ack deadline, buffer empty: the handler took the decision (the SDK got it),
// so the truthful outcome is resolved, not aborted.
func TestAckTimeoutAfterTheHandlerTookTheDecisionPublishesResolved(t *testing.T) {
	r := newPermRig(t, Config{PermissionTimeout: time.Hour, PermissionAckTimeout: 5 * time.Second})
	tui := connect(t, r.h, ClientTUI, 16)
	id, a := decidedEntry(t, r, tui)
	a.takeReply(t) // consumed, but the handler never returns

	r.clock.Advance(5 * time.Second)
	r.fire(t)
	p := resolvedPayload(t, recvKind(t, tui, KindPermissionResolved))
	if p.RequestID != id || p.Decision != "allow_once" || p.Actor != ActorLocalTUI {
		t.Fatalf("resolved = %+v", p)
	}
	noFrame(t, r.h, tui, "no permission.aborted may accompany a delivered decision")
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
	if d := a.takeReply(t); d.Kind != agent.DecisionUnavailable {
		t.Fatalf("handler would receive %v after step 0, want unavailable (no allow after shutdown began)", d.Kind)
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

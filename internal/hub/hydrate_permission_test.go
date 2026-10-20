package hub

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

// ---- 067.003-T: session.hydrate carries unresolved pending permissions ----

type hydrateView struct {
	Entries            []json.RawMessage            `json:"entries"`
	PendingPermissions []PermissionRequestedPayload `json:"pending_permissions"`
}

func hydrateOf(t *testing.T, f Frame) hydrateView {
	t.Helper()
	if f.Kind() != KindSessionHydrate {
		t.Fatalf("frame kind = %s, want hydrate", f.Kind())
	}
	var v hydrateView
	if err := json.Unmarshal(decode(t, f).Payload, &v); err != nil {
		t.Fatal(err)
	}
	return v
}

// hydrateFor connects a fresh remote client and returns its first frame.
func hydrateFor(t *testing.T, h *Hub) hydrateView {
	t.Helper()
	c, err := h.Connect(t.Context(), ConnectOptions{Kind: ClientRemote, Buffer: 16})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	return hydrateOf(t, recv(t, c))
}

func TestHydrateCarriesUnresolvedPermissionWithRemainingDeadline(t *testing.T) {
	r := newPermRig(t, Config{PermissionTimeout: time.Minute})
	tui := connect(t, r.h, ClientTUI, 16)
	send(t, r.h, msgEvent("m1", "hello"))
	a := newTestAsk(agent.PermissionKindShell, "git push")
	a.post(t, r.h)
	id := requestID(t, recvKind(t, tui, KindPermissionRequested))

	r.clock.Advance(20 * time.Second)
	v := hydrateFor(t, r.h)
	if len(v.Entries) != 1 {
		t.Fatalf("fold entries lost: %d", len(v.Entries))
	}
	if len(v.PendingPermissions) != 1 {
		t.Fatalf("pending_permissions = %+v, want the modal state", v.PendingPermissions)
	}
	p := v.PendingPermissions[0]
	if p.RequestID != id || p.ToolCallID != "call-1" || p.PermissionKind != "shell" || p.Summary != "git push" {
		t.Fatalf("pending = %+v", p)
	}
	if want := (40 * time.Second).Milliseconds(); p.RemainingMS != want {
		t.Fatalf("remaining_ms = %d, want %d", p.RemainingMS, want)
	}
}

func TestHydrateHasNoPendingSectionContentWhenNothingIsPending(t *testing.T) {
	h := startHub(t, Config{})
	c, err := h.Connect(t.Context(), ConnectOptions{Kind: ClientRemote, Buffer: 8})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	f := recv(t, c)
	if !strings.Contains(string(decode(t, f).Payload), `"pending_permissions":[]`) {
		t.Fatalf("payload %s must carry an explicit empty pending_permissions list", decode(t, f).Payload)
	}
}

func TestHydrateOmitsResolvedAndDecidedRequests(t *testing.T) {
	r := newPermRig(t, Config{PermissionTimeout: time.Minute})
	tui := connect(t, r.h, ClientTUI, 32)
	first := newTestAsk(agent.PermissionKindShell, "one")
	first.post(t, r.h)
	id1 := requestID(t, recvKind(t, tui, KindPermissionRequested))
	second := newTestAsk(agent.PermissionKindRead, "two")
	second.post(t, r.h)
	id2 := requestID(t, recvKind(t, tui, KindPermissionRequested))

	// Decided but not yet acknowledged: already resolved from the operator's
	// point of view, so it is not offered again.
	resolveAs(t, r.h, tuiP(), id1, allowOnce)
	v := hydrateFor(t, r.h)
	if len(v.PendingPermissions) != 1 || v.PendingPermissions[0].RequestID != id2 {
		t.Fatalf("pending = %+v, want only %s", v.PendingPermissions, id2)
	}

	// Fully resolved (acknowledged) as well.
	first.handlerAcks(t)
	recvKind(t, tui, KindPermissionResolved)
	resolveAs(t, r.h, tuiP(), id2, denyDecision)
	second.handlerAcks(t)
	recvKind(t, tui, KindPermissionResolved)
	if v := hydrateFor(t, r.h); len(v.PendingPermissions) != 0 {
		t.Fatalf("pending = %+v after everything resolved", v.PendingPermissions)
	}
}

func TestForcedRehydrateOfOverflowedTUICarriesPendingPermissions(t *testing.T) {
	r := newPermRig(t, Config{PermissionTimeout: time.Minute})
	tui := connect(t, r.h, ClientTUI, 2)
	newTestAsk(agent.PermissionKindShell, "cmd").post(t, r.h)
	// Do not read: flood so the TUI overflows and is forced to re-hydrate.
	for i := 0; i < 8; i++ {
		send(t, r.h, msgEvent("m", "x"))
	}
	deadline := time.Now().Add(testWait)
	for stats(t, r.h).Rehydrations == 0 {
		if time.Now().After(deadline) {
			t.Fatal("TUI was never force-re-hydrated")
		}
		time.Sleep(time.Millisecond)
	}
	var got hydrateView
	for len(got.PendingPermissions) == 0 {
		f := recv(t, tui)
		if f.Kind() == KindSessionHydrate {
			got = hydrateOf(t, f)
			break
		}
	}
	if len(got.PendingPermissions) != 1 {
		t.Fatalf("forced re-hydrate lost the modal state: %+v", got.PendingPermissions)
	}
}

func TestOversizedPermissionSummaryIsDeniedNotRegistered(t *testing.T) {
	h := startHub(t, Config{})
	a := newTestAsk(agent.PermissionKindShell, strings.Repeat("a", MaxPermissionSummaryBytes+1))
	a.post(t, h)
	d := a.takeReply(t)
	if d.Kind != agent.DecisionDeny {
		t.Fatalf("decision = %v, want deny (a request that cannot be reviewed in full is never approvable)", d.Kind)
	}
	if n := stats(t, h).PendingPermissions; n != 0 {
		t.Fatalf("oversized request was registered (%d)", n)
	}
}

// Review finding G: a request the adapter says was truncated cannot be reviewed
// in full, so the hub denies it itself (fail closed) rather than trusting every
// adapter to have denied it first; and a client-visible identifier is bounded.
func TestTruncatedPermissionRequestIsDeniedByTheHubNotRegistered(t *testing.T) {
	h := startHub(t, Config{})
	tui := connect(t, h, ClientTUI, 16)
	a := newTestAsk(agent.PermissionKindShell, "echo ok…")
	a.ask.Request.Truncated, a.ask.Request.FullLen = true, 9999
	a.post(t, h)
	if d := a.takeReply(t); d.Kind != agent.DecisionDeny {
		t.Fatalf("decision = %v, want deny", d.Kind)
	}
	if n := stats(t, h).PendingPermissions; n != 0 {
		t.Fatalf("a truncated request was registered (%d)", n)
	}
	noFrame(t, h, tui, "a truncated request must never be broadcast as approvable")
}

func TestPermissionToolCallIDIsBoundedInTheBroadcastAndHydrate(t *testing.T) {
	h := startHub(t, Config{})
	tui := connect(t, h, ClientTUI, 16)
	a := newTestAsk(agent.PermissionKindRead, "x")
	a.ask.Request.ToolCallID = strings.Repeat("c", 100000)
	a.post(t, h)
	var p PermissionRequestedPayload
	if err := json.Unmarshal(decode(t, recvKind(t, tui, KindPermissionRequested)).Payload, &p); err != nil {
		t.Fatal(err)
	}
	if len(p.ToolCallID) == 0 || len(p.ToolCallID) > maxIdentBytes {
		t.Fatalf("broadcast tool_call_id is %d bytes, want 1..%d", len(p.ToolCallID), maxIdentBytes)
	}
	v := hydrateFor(t, h)
	if len(v.PendingPermissions) != 1 || len(v.PendingPermissions[0].ToolCallID) > maxIdentBytes {
		t.Fatalf("hydrate carries an unbounded tool_call_id: %+v", v.PendingPermissions)
	}
}

func TestHydratePermissionSectionIsBounded(t *testing.T) {
	const max = 4
	h := startHub(t, Config{MaxPendingPermissions: max})
	tui := connect(t, h, ClientTUI, 64)
	nasty := strings.Repeat("<", MaxPermissionSummaryBytes) // 6x JSON escaping
	for i := 0; i < max+2; i++ {                            // two beyond the cap are refused
		newTestAsk(agent.PermissionKindShell, nasty).post(t, h)
	}
	for i := 0; i < max; i++ {
		recvKind(t, tui, KindPermissionRequested)
	}
	c, err := h.Connect(t.Context(), ConnectOptions{Kind: ClientRemote, Buffer: 8})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	f := recv(t, c)
	v := hydrateOf(t, f)
	if len(v.PendingPermissions) != max {
		t.Fatalf("pending = %d, want the capped %d", len(v.PendingPermissions), max)
	}
	if bound := SnapshotOverhead + DefaultMaxSnapshotBytes + max*PermissionEntryHydrateBytes; f.Len() > bound+frameHeaderBytes {
		t.Fatalf("hydrate frame %d bytes exceeds the declared bound %d", f.Len(), bound+frameHeaderBytes)
	}
}

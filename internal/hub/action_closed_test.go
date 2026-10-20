package hub

import (
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

// ---- review finding F: an action queued by a client that is gone ----

// The in-process TUI is held to the same rule: an action queued by a TUI
// client that then closed is refused at the hub step, not applied on behalf of a
// surface that is gone (it has no credential, so the reason is client_closed).
func TestActionQueuedByTheTUIThatThenClosedIsRefused(t *testing.T) {
	ag := newFakeAgent()
	h := startHub(t, Config{Agent: ag})
	tui := connectRole(t, h, ClientTUI, 0)
	connectRole(t, h, ClientRemote, RoleObserver) // keeps the broadcast set non-empty once the TUI closes
	id, a := pendingOf(t, h, tui, agent.PermissionKindRead)

	tui.Close() // closed after Do's check, before the hub's step
	for _, act := range []OperatorAction{answer(id, allowOnce), prompt("late"), {Type: OpAbortTurn}} {
		req := &actionReq{
			from:   principal{surface: surfaceTUI, role: RoleApprover},
			client: tui, act: act, reply: make(chan ActionResult, 1),
		}
		select {
		case h.actions <- req:
		case <-time.After(testWait):
			t.Fatal("hub did not take the action")
		}
		select {
		case res := <-req.reply:
			if res.Accepted || res.Reason != ReasonClientClosed {
				t.Fatalf("%v from a closed TUI = %+v, want client_closed", act.Type, res)
			}
		case <-time.After(testWait):
			t.Fatal("no verdict")
		}
	}
	select {
	case d := <-a.reply:
		t.Fatalf("handler got %v from a closed TUI's action", d.Kind)
	default:
	}
	if len(ag.promptsSeen()) != 0 {
		t.Fatal("a prompt from a closed TUI reached the agent")
	}
}

// Client.Do checks the client once, then queues the action; the client can be
// closed (credential revoked, dropped) before the hub processes it. The hub
// must re-check at processing time.
func TestActionQueuedByARemoteClientThatThenClosedIsRefusedAsRevoked(t *testing.T) {
	ag := newFakeAgent()
	h := startHub(t, Config{Agent: ag})
	tui := connectRole(t, h, ClientTUI, 0)
	appr := connectRole(t, h, ClientRemote, RoleApprover)
	id, a := pendingOf(t, h, tui, agent.PermissionKindRead)

	appr.Close() // closed after Do's check, before the hub's step
	for _, act := range []OperatorAction{answer(id, allowOnce), prompt("late"), {Type: OpAbortTurn}} {
		req := &actionReq{
			from:   principal{surface: surfaceRemote, role: RoleApprover},
			client: appr, act: act, reply: make(chan ActionResult, 1),
		}
		select {
		case h.actions <- req:
		case <-time.After(testWait):
			t.Fatal("hub did not take the action")
		}
		select {
		case res := <-req.reply:
			if res.Accepted || res.Reason != ReasonCredentialRevoked {
				t.Fatalf("%v from a closed client = %+v, want credential_revoked", act.Type, res)
			}
		case <-time.After(testWait):
			t.Fatal("no verdict")
		}
	}
	// Nothing took effect: the request is still pending and the agent saw nothing.
	if n := stats(t, h).PendingPermissions; n != 1 {
		t.Fatalf("pending = %d: a refused action must not consume the request", n)
	}
	select {
	case d := <-a.reply:
		t.Fatalf("handler got %v from a closed client's action", d.Kind)
	default:
	}
	if len(ag.promptsSeen()) != 0 {
		t.Fatal("a prompt from a closed client reached the agent")
	}
	select {
	case <-ag.abortCh:
		t.Fatal("an abort from a closed client reached the agent")
	case <-time.After(50 * time.Millisecond):
	}
}

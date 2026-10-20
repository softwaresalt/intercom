package hub

import (
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

// ---- 067.008-T: empty broadcast set resolves immediately ----
//
// Membership rule (design §5.4): the local TUI is a permanent member of the
// broadcast set, so "empty" means the last remote surface is gone AND the TUI
// is not attached. An overflowed (stale) TUI is still a member.

func wantUnavailable(t *testing.T, a *testAsk) {
	t.Helper()
	select {
	case d := <-a.reply:
		if d.Kind != agent.DecisionUnavailable {
			t.Fatalf("decision = %v, want unavailable", d.Kind)
		}
	case <-time.After(testWait):
		t.Fatalf("no decision within %v (the request waited for its deadline)", testWait)
	}
}

// settleMembership polls hub steps until the registered client counts match,
// so a following assertion cannot race the hub noticing a closed client.
func settleMembership(t *testing.T, h *Hub, remote int, tui bool) {
	t.Helper()
	deadline := time.Now().Add(testWait)
	for {
		s := stats(t, h)
		if s.Remote == remote && s.TUIAttached == tui {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("membership = %d remote, tui=%v; want %d, %v", s.Remote, s.TUIAttached, remote, tui)
		}
		time.Sleep(time.Millisecond)
	}
}

func assertStillPending(t *testing.T, h *Hub, a *testAsk) {
	t.Helper()
	if n := stats(t, h).PendingPermissions; n != 1 {
		t.Fatalf("pending = %d, want the request to stay pending", n)
	}
	select {
	case d := <-a.reply:
		t.Fatalf("request was resolved (%v) while a member was still present", d.Kind)
	default:
	}
}

func TestAskWithNoSurfaceAtAllResolvesUnavailableImmediately(t *testing.T) {
	h := startHub(t, Config{PermissionTimeout: time.Hour})
	a := newTestAsk(agent.PermissionKindShell, "cmd")
	a.post(t, h)
	wantUnavailable(t, a)
}

func TestLastRemoteDisconnectWithoutTUIResolvesWithoutWaitingForDeadline(t *testing.T) {
	h := startHub(t, Config{PermissionTimeout: time.Hour})
	rem := connect(t, h, ClientRemote, 16)
	a := newTestAsk(agent.PermissionKindShell, "cmd")
	a.post(t, h)
	recvKind(t, rem, KindPermissionRequested)
	assertStillPending(t, h, a) // a remote surface is still present

	rem.Close() // the last surface goes away
	wantUnavailable(t, a)
}

func TestTUIPresenceKeepsRequestPendingAfterLastRemoteLeaves(t *testing.T) {
	h := startHub(t, Config{PermissionTimeout: time.Hour})
	tui := connect(t, h, ClientTUI, 16)
	rem := connect(t, h, ClientRemote, 16)
	a := newTestAsk(agent.PermissionKindShell, "cmd")
	a.post(t, h)
	recvKind(t, rem, KindPermissionRequested)

	rem.Close()
	settleMembership(t, h, 0, true)
	assertStillPending(t, h, a)

	tui.Close() // now the set is empty by the permanent-TUI rule
	wantUnavailable(t, a)
}

func TestOverflowedTUIStillCountsAsAMember(t *testing.T) {
	h := startHub(t, Config{PermissionTimeout: time.Hour})
	tui := connect(t, h, ClientTUI, 2) // never read: it will overflow
	for i := 0; i < 6; i++ {
		send(t, h, msgEvent("m", "x"))
	}
	deadline := time.Now().Add(testWait)
	for stats(t, h).TUIOverflows == 0 {
		if time.Now().After(deadline) {
			t.Fatal("TUI never overflowed")
		}
		time.Sleep(time.Millisecond)
	}
	a := newTestAsk(agent.PermissionKindShell, "cmd")
	a.post(t, h)
	settleMembership(t, h, 0, true)
	assertStillPending(t, h, a) // dropped-for-backpressure would have emptied the set
	_ = tui
}

func TestEmptySetFastResolveIsRaceCleanUnderChurn(t *testing.T) {
	h := startHub(t, Config{PermissionTimeout: time.Hour, MaxPendingPermissions: 64})
	for i := 0; i < 40; i++ {
		rem := connect(t, h, ClientRemote, 16)
		a := newTestAsk(agent.PermissionKindRead, "p")
		a.post(t, h)
		recvKind(t, rem, KindPermissionRequested)
		rem.Close()
		wantUnavailable(t, a)
	}
}

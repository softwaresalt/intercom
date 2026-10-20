package transport

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/hub"
	"github.com/softwaresalt/intercom-go/internal/transport/wsclient"
)

// 070.002-T: Remote is the hub.Gate the shutdown executor closes at step 0.

func (rr *remoteRig) dial(role hub.Role) (*wsclient.Client, Issued) {
	rr.t.Helper()
	iss, err := rr.remote.Mint(role)
	if err != nil {
		rr.t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, err := wsclient.Dial(ctx, wsclient.Options{URL: rr.url(), Credential: iss.Token.Reveal()})
	if err != nil {
		rr.t.Fatalf("dial: %v", err)
	}
	rr.t.Cleanup(c.Abort)
	return c, iss
}

func TestGateRevokeCredentialsRefusesNewAndKeepsLiveSocketsOpen(t *testing.T) {
	rr := newRemoteRig(t, defaultRemote())
	live, iss := rr.dial(hub.RoleApprover)
	eventually(t, "session registered", 5*time.Second, func() bool { return rr.remote.handler.ActiveSessions() == 1 })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rr.remote.RevokeCredentials(ctx); err != nil {
		t.Fatalf("RevokeCredentials: %v", err)
	}
	if err := rr.remote.RevokeCredentials(ctx); err != nil {
		t.Fatalf("RevokeCredentials is not idempotent: %v", err)
	}

	// A new upgrade with the (now revoked) credential is refused.
	_, err := wsclient.Dial(ctx, wsclient.Options{URL: rr.url(), Credential: iss.Token.Reveal()})
	var he *wsclient.HandshakeError
	if !errors.As(err, &he) || he.Status != 503 {
		t.Fatalf("dial after revoke: %v; want a 503 (shutting down) refusal", err)
	}
	// Nothing can be minted during shutdown (re-issue is refused).
	if _, err := rr.remote.Mint(hub.RoleApprover); err == nil {
		t.Fatal("a credential was minted after shutdown revoke")
	}

	// The established socket is NOT closed by the revoke (ADR MVP-3: pumps
	// close after the shutdown broadcast and drain); its actions are refused.
	if n := rr.remote.handler.ActiveSessions(); n != 1 {
		t.Fatalf("revoke closed the live session (active = %d)", n)
	}
	if err := live.Prompt(ctx, "hi"); err != nil {
		t.Fatalf("the live socket is closed after revoke: %v", err)
	}
	for {
		m, err := live.Next(ctx)
		if err != nil {
			t.Fatalf("live socket closed instead of refusing the action: %v", err)
		}
		if m.Reply != nil {
			if m.Reply.Accepted() || (m.Reply.Reason != "credential_revoked" && m.Reply.Reason != "shutting_down") {
				t.Fatalf("reply = %+v; want a revoked/shutting-down refusal", *m.Reply)
			}
			break
		}
	}
}

func TestGateCloseListenerStopsAcceptingButKeepsSessions(t *testing.T) {
	rr := newRemoteRig(t, defaultRemote())
	rr.dial(hub.RoleObserver)
	eventually(t, "session registered", 5*time.Second, func() bool { return rr.remote.handler.ActiveSessions() == 1 })

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	addr := rr.remote.Addr().String()
	if err := rr.remote.CloseListener(ctx); err != nil {
		t.Fatalf("CloseListener: %v", err)
	}
	if err := rr.remote.CloseListener(ctx); err != nil {
		t.Fatalf("CloseListener is not idempotent: %v", err)
	}
	if c, err := net.DialTimeout("tcp", addr, time.Second); err == nil {
		_ = c.Close()
		t.Fatal("the listener still accepts after CloseListener")
	}
	if n := rr.remote.handler.ActiveSessions(); n != 1 {
		t.Fatalf("CloseListener tore a session down (active = %d)", n)
	}
}

func TestGateCloseListenerHonoursCancelledContextBeforeClosing(t *testing.T) {
	rr := newRemoteRig(t, defaultRemote())
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if err := rr.remote.CloseListener(cancelled); err == nil {
		t.Fatal("CloseListener ignored a cancelled context")
	}
	addr := rr.remote.Addr().String()
	c, err := net.DialTimeout("tcp", addr, time.Second)
	if err != nil {
		t.Fatalf("a cancelled CloseListener still closed the listener: %v", err)
	}
	_ = c.Close()
}

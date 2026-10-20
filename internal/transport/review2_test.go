package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

// serveSession serves ONE hand-built session over a throwaway HTTP server so a
// test can pre-arrange the session's internal state (queued replies, a
// replaced read pump) before its pumps start. setup runs before run().
func (e *env) serveSession(t *testing.T, grant *Grant, setup func(s *session)) *websocket.Conn {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client, err := e.hub.Connect(ctx, hub.ConnectOptions{Kind: hub.ClientRemote, Role: grant.Role()})
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer close(done)
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			client.Close()
			return
		}
		s := newSession(e.handler, conn, client, grant)
		setup(s)
		s.run()
	}))
	t.Cleanup(func() {
		client.Close()
		srv.Close()
	})
	dctx, dcancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer dcancel()
	c, _, err := websocket.Dial(dctx, "ws://"+srv.Listener.Addr().String(), nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.CloseNow() })
	return c
}

// readToClose reads until the close frame, returning every action reply seen
// before it and the close status and reason.
func readToClose(t *testing.T, c *websocket.Conn, within time.Duration) (replies []result, code websocket.StatusCode, reason string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), within)
	defer cancel()
	for {
		_, data, err := c.Read(ctx)
		if err != nil {
			var ce websocket.CloseError
			if !errors.As(err, &ce) {
				t.Fatalf("read error %v; want a close frame", err)
			}
			return replies, ce.Code, ce.Reason
		}
		if r, ok := parseResult(data); ok {
			replies = append(replies, r)
		}
	}
}

func parseResult(data []byte) (result, bool) {
	var r result
	if err := json.Unmarshal(data, &r); err != nil || !strings.HasPrefix(r.Type, "action.") {
		return result{}, false
	}
	return r, true
}

// Item 5: on revocation the pump flushes already queued refusals before the
// 1008 close, deterministically (both signals are ready when the pump starts).
func TestPumpFlushesQueuedRefusalBeforeRevocationClose(t *testing.T) {
	e := newEnv(t, envOpts{})
	act := hub.OperatorAction{Type: hub.OpPrompt}
	for i := 0; i < 12; i++ {
		is, err := e.creds.Mint(hub.RoleApprover)
		if err != nil {
			t.Fatal(err)
		}
		g, err := e.creds.Verify(is.Token.Reveal(), e.hub.Gen())
		if err != nil {
			t.Fatal(err)
		}
		c := e.serveSession(t, g, func(s *session) {
			s.replies <- refusalFor(act, string(hub.ReasonCredentialRevoked))
			e.creds.Revoke(is.ID) // Revoked() is ready as the pump starts
		})
		replies, code, reason := readToClose(t, c, 5*time.Second)
		if len(replies) != 1 || replies[0].Reason != string(hub.ReasonCredentialRevoked) {
			t.Fatalf("iteration %d: replies before close = %+v; want the credential_revoked refusal", i, replies)
		}
		if code != websocket.StatusPolicyViolation || reason != closeReasonEnded {
			t.Fatalf("close = %d %q; want 1008 %q", code, reason, closeReasonEnded)
		}
	}
}

// Item 5: likewise on credential expiry.
func TestPumpFlushesQueuedRefusalBeforeExpiryClose(t *testing.T) {
	clk := newFakeClockAt(time.Now())
	e := newEnv(t, envOpts{clock: clk})
	act := hub.OperatorAction{Type: hub.OpPrompt}
	for i := 0; i < 12; i++ {
		is, err := e.creds.Mint(hub.RoleApprover)
		if err != nil {
			t.Fatal(err)
		}
		g, err := e.creds.Verify(is.Token.Reveal(), e.hub.Gen())
		if err != nil {
			t.Fatal(err)
		}
		clk.Advance(2 * time.Hour) // the grant is now past its expiry: the timer fires at once
		c := e.serveSession(t, g, func(s *session) {
			s.replies <- refusalFor(act, ReasonCredentialExpired)
		})
		replies, code, reason := readToClose(t, c, 5*time.Second)
		if len(replies) != 1 || replies[0].Reason != ReasonCredentialExpired {
			t.Fatalf("iteration %d: replies before close = %+v; want the credential_expired refusal", i, replies)
		}
		if code != websocket.StatusPolicyViolation || reason != closeReasonEnded {
			t.Fatalf("close = %d %q", code, reason)
		}
	}
}

// Item 9: the pump's expiry timer and Validate share the store's clock, so a
// skewed Config.Now cannot stop (or hasten) expiry.
func TestPumpExpiryFollowsStoreClockNotConfigNow(t *testing.T) {
	e := newEnv(t, envOpts{lifetime: 1500 * time.Millisecond, cfgTweak: func(c *Config) {
		c.Now = func() time.Time { return time.Unix(0, 0) } // wildly skewed handler clock
	}})
	c := dialWS(t, e, e.obs.Token.Reveal())
	readEnvelope(t, c, 3*time.Second)
	_, code, _ := readToClose(t, c, 10*time.Second)
	if code != websocket.StatusPolicyViolation {
		t.Fatalf("close code = %d; want 1008 at the credential's expiry", code)
	}
}

func TestStoreRemainingUsesStoreClock(t *testing.T) {
	clk := newFakeClockAt(time.Unix(5000, 0))
	s := newTestStore(t, clk) // lifetime 1h
	is := mustMint(t, s, hub.RoleApprover)
	g, err := s.Verify(is.Token.Reveal(), testGen)
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Remaining(g); got != time.Hour {
		t.Fatalf("Remaining = %v; want 1h", got)
	}
	clk.Advance(40 * time.Minute)
	if got := s.Remaining(g); got != 20*time.Minute {
		t.Fatalf("Remaining = %v; want 20m", got)
	}
	clk.Advance(time.Hour)
	if got := s.Remaining(g); got > 0 {
		t.Fatalf("Remaining = %v after expiry; want <= 0", got)
	}
	if got := s.Remaining(nil); got > 0 {
		t.Fatalf("Remaining(nil) = %v; want <= 0", got)
	}
}

// Item 10: a panic in the read pump closes the client; the process survives.
func TestReadPumpPanicClosesClientWithoutCrashing(t *testing.T) {
	e := newEnv(t, envOpts{})
	g, err := e.creds.Verify(e.app.Token.Reveal(), e.hub.Gen())
	if err != nil {
		t.Fatal(err)
	}
	c := e.serveSession(t, g, func(s *session) {
		s.read = func(context.Context) { panic("boom") }
	})
	_, code, _ := readToClose(t, c, 5*time.Second)
	if code != websocket.StatusGoingAway {
		t.Fatalf("close code = %d; want 1001 after the read pump died", code)
	}
	eventually(t, "session gone", 3*time.Second, func() bool { return e.handler.ActiveSessions() == 0 })
}

// Item 8: CloseListener only stops accepting; sessions stay alive until
// Shutdown, which remains the final step.
func TestCloseListenerStopsAcceptingButKeepsSessions(t *testing.T) {
	e := newEnv(t, envOpts{})
	c := dialWS(t, e, e.app.Token.Reveal())
	readEnvelope(t, c, 2*time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := e.srv.CloseListener(ctx); err != nil {
		t.Fatalf("CloseListener: %v", err)
	}
	if err := e.srv.CloseListener(ctx); err != nil {
		t.Fatalf("second CloseListener: %v", err)
	}
	if conn, err := net.DialTimeout("tcp", e.hostPort(), 500*time.Millisecond); err == nil {
		_ = conn.Close()
		t.Fatal("listener still accepts after CloseListener")
	}
	// the established session is untouched and still receives events
	e.event("after-close-listener")
	got := readEnvelope(t, c, 3*time.Second)
	if got.Kind != hub.KindAgentMessage {
		t.Fatalf("frame = %s; want the event on a live session", got.Kind)
	}
	if n := e.handler.ActiveSessions(); n != 1 {
		t.Fatalf("ActiveSessions = %d after CloseListener; want 1", n)
	}
	// Shutdown is still the final step and tears the session down.
	sctx, scancel := context.WithTimeout(context.Background(), MinShutdownBudget+2*time.Second)
	defer scancel()
	if err := e.srv.Shutdown(sctx); err != nil {
		t.Fatalf("Shutdown after CloseListener: %v", err)
	}
	if n := e.handler.ActiveSessions(); n != 0 {
		t.Fatalf("%d sessions outlived Shutdown", n)
	}
}

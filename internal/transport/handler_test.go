package transport

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/coder/websocket"

	"github.com/softwaresalt/intercom-go/internal/hub"
)

// wsKey returns a fresh, valid Sec-WebSocket-Key (16 random bytes, base64),
// generated at run time so no key literal appears in the source.
func wsKey(t *testing.T) string {
	t.Helper()
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}

// upgradeRequest sends an authenticated /ws upgrade request with the given
// Sec-WebSocket-Version and key through a plain HTTP client, so the response
// of a refused handshake can be read.
func upgradeRequest(t *testing.T, e *env, version, key string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, "http://"+e.hostPort()+WSPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+e.app.Token.Reveal())
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", version)
	req.Header.Set("Sec-WebSocket-Key", key)
	req.Header.Set("Sec-WebSocket-Protocol", SubprotocolV1)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func TestHandlerRejectAccountingCovers404AcceptFailureAndExcludesDraining503(t *testing.T) {
	rec := &rejectLog{}
	e := newEnv(t, envOpts{maxConns: 1, cfgTweak: func(c *Config) { c.OnReject = rec.record }})
	// maxConns is 1 so the over-cap case below is reachable; every request in
	// this test closes its connection (the server sets Connection: close).
	wantCount := func(n uint64, what string) {
		t.Helper()
		eventually(t, what, 2*time.Second, func() bool { return e.handler.RejectedUpgrades() == n })
		if got := e.handler.RejectedUpgrades(); got != n {
			t.Fatalf("%s: RejectedUpgrades = %d; want %d", what, got, n)
		}
	}
	last := func() rejection {
		l := rec.snapshot()
		if len(l) == 0 {
			t.Fatal("OnReject was never called")
		}
		return l[len(l)-1]
	}

	// 1. Wrong path: counted as wrong_path.
	resp, err := http.Get("http://" + e.hostPort() + "/nope")
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("wrong path status = %d; want 404", resp.StatusCode)
	}
	wantCount(1, "after wrong-path 404")
	if got := last(); got.status != http.StatusNotFound || got.reason != rejectNotFound {
		t.Fatalf("404 rejection = %+v", got)
	}

	// 2. Handshake the library refuses (unsupported version): counted once as
	// invalid_handshake with the library's single response, no hub client.
	resp = upgradeRequest(t, e, "99", wsKey(t))
	if resp.StatusCode < 400 {
		t.Fatalf("bad-version handshake status = %d; want a refusal", resp.StatusCode)
	}
	wantCount(2, "after invalid handshake")
	if got := last(); got.reason != rejectHandshake || got.status != resp.StatusCode {
		t.Fatalf("handshake rejection = %+v (response %d)", got, resp.StatusCode)
	}
	resp = upgradeRequest(t, e, "13", "not-a-16-byte-key")
	if resp.StatusCode < 400 {
		t.Fatalf("bad-key handshake status = %d; want a refusal", resp.StatusCode)
	}
	wantCount(3, "after invalid key")
	if e.handler.ActiveSessions() != 0 {
		t.Fatal("a refused handshake created a session")
	}

	// 3. Over the connection cap: the listener's 503 is counted.
	hold, err := net.Dial("tcp", e.hostPort())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = hold.Close() }()
	eventually(t, "cap slot held", 2*time.Second, func() bool { return e.srv.ActiveConnections() == 1 })
	over, err := http.Get("http://" + e.hostPort() + WSPath)
	if err == nil {
		_ = over.Body.Close()
		if over.StatusCode != http.StatusServiceUnavailable {
			t.Fatalf("over-cap status = %d; want 503", over.StatusCode)
		}
	}
	wantCount(4, "after over-cap 503")
	if got := last(); got.status != http.StatusServiceUnavailable || got.reason != rejectOverCap {
		t.Fatalf("over-cap rejection = %+v", got)
	}
	_ = hold.Close()
	eventually(t, "cap slot released", 2*time.Second, func() bool { return e.srv.ActiveConnections() == 0 })

	// 4. A benign draining 503 is NOT counted and does not call OnReject.
	e.creds.RevokeAll()
	before := len(rec.snapshot())
	resp = upgradeRequest(t, e, "13", wsKey(t))
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("draining status = %d; want 503", resp.StatusCode)
	}
	time.Sleep(50 * time.Millisecond)
	if got := e.handler.RejectedUpgrades(); got != 4 {
		t.Fatalf("draining 503 changed RejectedUpgrades to %d; want 4", got)
	}
	if after := len(rec.snapshot()); after != before {
		t.Fatalf("draining 503 invoked OnReject (%d -> %d calls)", before, after)
	}
}

// The credential can die between ServeHTTP's Verify and the hub step that
// registers the client. Connect carries a Validate hook, so the hub refuses the
// connect inside its step instead of upgrading a revoked credential.
func TestHandlerConnectOptionsValidateClosesConnectTOCTOU(t *testing.T) {
	e := newEnv(t, envOpts{})
	g, err := e.creds.Verify(e.app.Token.Reveal(), e.hub.Gen())
	if err != nil {
		t.Fatal(err)
	}
	// grantValidator maps each failure to a hub reason.
	v := e.handler.grantValidator(g)
	if r := v(e.hub.Gen()); r != "" {
		t.Fatalf("fresh grant validator = %q; want valid", r)
	}
	if r := v("other-gen"); r != hub.ReasonGenerationMismatch {
		t.Fatalf("wrong-gen validator = %q; want %q", r, hub.ReasonGenerationMismatch)
	}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verified earlier (g), revoked since: exactly the TOCTOU window.
		e.creds.Revoke(g.CredentialID())
		e.handler.upgrade(w, r, g)
	}))
	defer srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	c, resp, err := websocket.Dial(ctx, "ws://"+srv.Listener.Addr().String()+WSPath, nil)
	if err == nil {
		_ = c.CloseNow()
		t.Fatal("a credential revoked before the hub step was upgraded")
	}
	if resp == nil || resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("response = %v; want 401", resp)
	}
	if got := e.handler.RejectedUpgrades(); got != 1 {
		t.Fatalf("RejectedUpgrades = %d; want 1", got)
	}

	// Shutdown revocation maps to the draining reason.
	g2, err := e.creds.Verify(e.obs.Token.Reveal(), e.hub.Gen())
	if err != nil {
		t.Fatal(err)
	}
	e.creds.RevokeAll()
	if r := e.handler.grantValidator(g2)(e.hub.Gen()); r != hub.ReasonShuttingDown {
		t.Fatalf("post-RevokeAll validator = %q; want %q", r, hub.ReasonShuttingDown)
	}
}

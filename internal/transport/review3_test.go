package transport

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/coder/websocket"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

// F-5: the http.Server's ReadTimeout/ReadHeaderTimeout must not end an
// upgraded session. net/http clears the connection deadline on Hijack, so a
// session outlives HandshakeTimeout; this proves it end to end.
func TestSessionOutlivesHandshakeTimeout(t *testing.T) {
	const hs = 300 * time.Millisecond
	fa := &fakeAgent{}
	e := newEnv(t, envOpts{
		hubCfg:   hub.Config{Agent: fa},
		srvTweak: func(sc *ServerConfig) { sc.HandshakeTimeout = hs },
	})
	c := dialWS(t, e, e.app.Token.Reveal())
	readEnvelope(t, c, 2*time.Second)

	time.Sleep(2*hs + 200*time.Millisecond)

	e.event("after-handshake-timeout")
	if got := readEnvelope(t, c, 3*time.Second); got.Kind != hub.KindAgentMessage {
		t.Fatalf("frame = %s; want the event on a session older than HandshakeTimeout", got.Kind)
	}
	sendText(t, c, `{"op":"prompt","prompt":"still alive","request_id":"r9"}`)
	if r := nextResult(t, c); r.Type != ReplyAccepted {
		t.Fatalf("action after 2x HandshakeTimeout = %+v; want accepted", r)
	}
	if n := e.handler.ActiveSessions(); n != 1 {
		t.Fatalf("ActiveSessions = %d; want 1", n)
	}
}

// F-6: a credential revoked by shutdown is refused with a generic 503, not 401.
func TestVerifyShutdownRevokedCredentialIs503(t *testing.T) {
	e := newEnv(t, envOpts{})
	tok := e.app.Token.Reveal()
	e.creds.RevokeAll()

	if _, err := e.creds.Verify(tok, e.hub.Gen()); !errors.Is(err, ErrShuttingDown) {
		t.Fatalf("Verify = %v; want ErrShuttingDown", err)
	}
	resp := e.rawUpgrade(bearer(tok), "")
	wantStatus(t, resp, http.StatusServiceUnavailable, "shutdown-revoked credential")
	if !resp.Close { // net/http folds "Connection: close" into Response.Close
		t.Fatalf("Connection header = %q; want close", resp.Header.Get("Connection"))
	}
	if resp.Header.Get("WWW-Authenticate") != "" {
		t.Fatal("503 must not carry WWW-Authenticate")
	}
	if b := bodyOf(resp); b != "" {
		t.Fatalf("503 body = %q; want empty", b)
	}
	// An unknown credential stays an indistinguishable 401.
	// change the LAST character to one that differs, so the value is always a different (same-shape) credential\n	last := byte('A')\n	if tok[len(tok)-1] == last {\n		last = 'B'\n	}\n	wantStatus(t, e.rawUpgrade(bearer(tok[:len(tok)-1]+string(last)), ""), http.StatusUnauthorized, "wrong credential during shutdown")
}

// F-11: the 404 for a non-/ws path has an empty body like every rejection.
func TestNotFoundHasEmptyBody(t *testing.T) {
	e := newEnv(t, envOpts{})
	resp, err := http.Get("http://" + e.hostPort() + "/other")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	wantStatus(t, resp, http.StatusNotFound, "other path")
	if b := bodyOf(resp); b != "" {
		t.Fatalf("404 body = %q; want empty", b)
	}
}

// F-8: request_id is truncated on a rune boundary.
func TestEchoRequestIDIsUTF8Safe(t *testing.T) {
	for _, unit := range []string{"a", "é", "世", "😀"} {
		id := strings.Repeat(unit, 200)
		got := echoRequestID(id)
		if len(got) > maxEchoedRequestID {
			t.Fatalf("%q: len %d > %d", unit, len(got), maxEchoedRequestID)
		}
		if !utf8.ValidString(got) {
			t.Fatalf("%q: truncated id is not valid UTF-8", unit)
		}
		if len(got) < maxEchoedRequestID-utf8.UTFMax+1 {
			t.Fatalf("%q: truncated too much: %d", unit, len(got))
		}
	}
	if got := echoRequestID("short"); got != "short" {
		t.Fatalf("short id changed: %q", got)
	}
}

// F-8: transport refusals echo the offending action's request_id.
func TestRefusalForEchoesRequestID(t *testing.T) {
	r, ok := parseResult(refusalFor(hub.OperatorAction{Type: hub.OpPrompt, RequestID: "req-7"}, ReasonRateLimited))
	if !ok || r.RequestID != "req-7" || r.Reason != ReasonRateLimited {
		t.Fatalf("refusal = %+v", r)
	}
	long := strings.Repeat("世", 100)
	r, _ = parseResult(refusalFor(hub.OperatorAction{Type: hub.OpPrompt, RequestID: long}, ReasonRateLimited))
	if !utf8.ValidString(r.RequestID) || len(r.RequestID) > maxEchoedRequestID {
		t.Fatalf("echoed id invalid or too long: %d bytes", len(r.RequestID))
	}
}

func TestRateLimitedAndRevokedRefusalsEchoRequestID(t *testing.T) {
	e := newEnv(t, envOpts{
		hubCfg:   hub.Config{Agent: &fakeAgent{}},
		cfgTweak: func(c *Config) { c.ActionRate = 0.001; c.ActionBurst = 1; c.MaxRateViolations = 50 },
	})
	c := dialWS(t, e, e.app.Token.Reveal())
	readEnvelope(t, c, 2*time.Second)
	sendText(t, c, `{"op":"prompt","prompt":"one","request_id":"a"}`)
	if r := nextResult(t, c); r.Type != ReplyAccepted || r.RequestID != "a" {
		t.Fatalf("first = %+v", r)
	}
	sendText(t, c, `{"op":"prompt","prompt":"two","request_id":"b"}`)
	if r := nextResult(t, c); r.Reason != ReasonRateLimited || r.RequestID != "b" {
		t.Fatalf("rate-limited reply = %+v; want request_id b", r)
	}
	// an unknown op keeps its request id too
	e2 := newEnv(t, envOpts{})
	c2 := dialWS(t, e2, e2.app.Token.Reveal())
	readEnvelope(t, c2, 2*time.Second)
	e2.creds.RevokeAll()
	sendText(t, c2, `{"op":"abort_turn","request_id":"rv"}`)
	if r := nextResult(t, c2); r.Reason != string(hub.ReasonCredentialRevoked) || r.RequestID != "rv" {
		t.Fatalf("revoked reply = %+v; want request_id rv", r)
	}
}

// F-12: during the shutdown-revoked state refusals pass through the limiter,
// so a flood is rate limited and eventually closed.
func TestRevokedStateFloodIsRateLimitedAndClosed(t *testing.T) {
	e := newEnv(t, envOpts{
		hubCfg:   hub.Config{Agent: &fakeAgent{}},
		cfgTweak: func(c *Config) { c.ActionRate = 0.001; c.ActionBurst = 2; c.MaxRateViolations = 3 },
	})
	c := dialWS(t, e, e.app.Token.Reveal())
	readEnvelope(t, c, 2*time.Second)
	e.creds.RevokeAll()

	for i := 0; i < 8; i++ {
		sendText(t, c, `{"op":"abort_turn"}`)
	}
	replies, code, reason := readToClose(t, c, 5*time.Second)
	if code != websocket.StatusPolicyViolation || reason != closeReasonRate {
		t.Fatalf("close = %d %q; want 1008 %q", code, reason, closeReasonRate)
	}
	var revoked, limited int
	for _, r := range replies {
		switch r.Reason {
		case string(hub.ReasonCredentialRevoked):
			revoked++
		case ReasonRateLimited:
			limited++
		}
	}
	if revoked != 2 || limited < 3 {
		t.Fatalf("revoked=%d rate_limited=%d; want 2 within burst, then >=3 limited (replies %+v)", revoked, limited, replies)
	}
}

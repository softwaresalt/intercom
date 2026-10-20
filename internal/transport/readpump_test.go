package transport

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

type fakeAgent struct {
	mu      sync.Mutex
	prompts []string
	aborts  int
}

func (f *fakeAgent) Send(_ context.Context, p string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.prompts = append(f.prompts, p)
	return nil
}

func (f *fakeAgent) Abort(context.Context) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.aborts++
	return nil
}

func (f *fakeAgent) snapshot() ([]string, int) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.prompts...), f.aborts
}

func sendText(t *testing.T, c *websocket.Conn, msg string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := c.Write(ctx, websocket.MessageText, []byte(msg)); err != nil {
		t.Fatalf("write: %v", err)
	}
}

// result is the transport-owned reply to one inbound action.
type result struct {
	V         int    `json:"v"`
	Type      string `json:"type"`
	Op        string `json:"op"`
	Reason    string `json:"reason"`
	Winner    string `json:"winner"`
	RequestID string `json:"request_id"`
}

// nextResult reads frames until a transport reply (type action.*) arrives,
// skipping hub envelopes.
func nextResult(t *testing.T, c *websocket.Conn) result {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_, data, err := c.Read(ctx)
		cancel()
		if err != nil {
			t.Fatalf("waiting for an action reply: %v", err)
		}
		var r result
		if json.Unmarshal(data, &r) == nil && strings.HasPrefix(r.Type, "action.") {
			return r
		}
	}
	t.Fatal("no action reply")
	return result{}
}

func expectClose(t *testing.T, c *websocket.Conn, want websocket.StatusCode, within time.Duration) string {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), within)
	defer cancel()
	for {
		_, _, err := c.Read(ctx)
		if err == nil {
			continue
		}
		var ce websocket.CloseError
		if !errors.As(err, &ce) {
			t.Fatalf("read error %v; want a close frame with code %d", err, want)
		}
		if ce.Code != want {
			t.Fatalf("close code = %d (%q); want %d", ce.Code, ce.Reason, want)
		}
		return ce.Reason
	}
}

func TestOversizedMessageClosesConnection(t *testing.T) {
	e := newEnv(t, envOpts{cfgTweak: func(c *Config) { c.MaxMessageBytes = 512 }})
	c := dialWS(t, e, e.app.Token.Reveal())
	readEnvelope(t, c, 2*time.Second)
	sendText(t, c, `{"op":"prompt","prompt":"`+strings.Repeat("a", 2048)+`"}`)
	if reason := expectClose(t, c, websocket.StatusMessageTooBig, 5*time.Second); reason == "" {
		t.Log("library close carried no reason")
	}
	eventually(t, "session ended", 3*time.Second, func() bool { return e.handler.ActiveSessions() == 0 })
}

func TestIdleConnectionClosedAtDeadline(t *testing.T) {
	e := newEnv(t, envOpts{cfgTweak: func(c *Config) { c.IdleTimeout = 250 * time.Millisecond }})
	c := dialWS(t, e, e.app.Token.Reveal())
	readEnvelope(t, c, 2*time.Second)
	begin := time.Now()
	if reason := expectClose(t, c, websocket.StatusPolicyViolation, 3*time.Second); reason != closeReasonIdle {
		t.Fatalf("idle close reason = %q; want %q", reason, closeReasonIdle)
	}
	if d := time.Since(begin); d < 150*time.Millisecond || d > 2*time.Second {
		t.Fatalf("idle close after %v; want near the 250ms deadline", d)
	}
	eventually(t, "session ended", 3*time.Second, func() bool { return e.handler.ActiveSessions() == 0 })
}

func TestActivityRefreshesIdleDeadline(t *testing.T) {
	e := newEnv(t, envOpts{cfgTweak: func(c *Config) { c.IdleTimeout = 300 * time.Millisecond }})
	c := dialWS(t, e, e.app.Token.Reveal())
	readEnvelope(t, c, 2*time.Second)
	for i := 0; i < 5; i++ { // 750ms total: alive only because messages refresh the deadline
		time.Sleep(150 * time.Millisecond)
		sendText(t, c, `{"op":"abort_turn"}`)
		nextResult(t, c)
	}
	if e.handler.ActiveSessions() != 1 {
		t.Fatal("an active connection was closed as idle")
	}
}

func TestReconnectReplaysWithValidCursorOtherwiseHydrates(t *testing.T) {
	e := newEnv(t, envOpts{hubCfg: hub.Config{ReplayRing: 4}})
	first := dialWS(t, e, e.obs.Token.Reveal())
	if h := readEnvelope(t, first, 2*time.Second); h.Kind != hub.KindSessionHydrate {
		t.Fatalf("first connect frame = %s", h.Kind)
	}
	for i := 0; i < 3; i++ {
		e.event("a")
	}
	var last hub.Envelope
	for i := 0; i < 3; i++ {
		last = readEnvelope(t, first, 2*time.Second)
	}
	_ = first.CloseNow()
	gen := e.hub.Gen()

	e.event("b") // missed while disconnected
	e.event("c")

	dialCursor := func(g string, seq uint64) *websocket.Conn {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		u := e.wsURL() + "?gen=" + g + "&seq=" + strconv.FormatUint(seq, 10)
		c, _, err := websocket.Dial(ctx, u, &websocket.DialOptions{HTTPHeader: authHeader(e.obs.Token.Reveal())})
		if err != nil {
			t.Fatalf("dial with cursor: %v", err)
		}
		t.Cleanup(func() { _ = c.CloseNow() })
		return c
	}

	// valid cursor, ring covers it: replay seq+1...HEAD with no hydrate.
	c := dialCursor(gen, last.Seq)
	r1 := readEnvelope(t, c, 2*time.Second)
	r2 := readEnvelope(t, c, 2*time.Second)
	if r1.Kind == hub.KindSessionHydrate || r1.Seq != last.Seq+1 || r2.Seq != last.Seq+2 {
		t.Fatalf("expected replay seq %d,%d; got %s/%d then %d", last.Seq+1, last.Seq+2, r1.Kind, r1.Seq, r2.Seq)
	}
	_ = c.CloseNow()

	// wrong generation: full hydrate.
	c = dialCursor("00000000-0000-7000-8000-000000000000", last.Seq)
	if h := readEnvelope(t, c, 2*time.Second); h.Kind != hub.KindSessionHydrate {
		t.Fatalf("wrong gen got %s; want hydrate", h.Kind)
	}
	_ = c.CloseNow()

	// the ring (4) no longer covers an old cursor: full hydrate.
	for i := 0; i < 10; i++ {
		e.event("d")
	}
	c = dialCursor(gen, last.Seq)
	if h := readEnvelope(t, c, 2*time.Second); h.Kind != hub.KindSessionHydrate {
		t.Fatalf("uncovered cursor got %s; want hydrate", h.Kind)
	}
}

func TestMalformedCursorIsRejectedBeforeUpgrade(t *testing.T) {
	e := newEnv(t, envOpts{})
	hdr := bearer(e.app.Token.Reveal())
	for _, q := range []string{"gen=abc", "seq=1", "gen=abc&seq=x", "gen=abc&seq=-1", "gen=abc&seq=1&seq=2", "gen=&seq=1", "gen=" + strings.Repeat("a", 200) + "&seq=1"} {
		wantStatus(t, e.rawUpgrade(hdr, q), http.StatusBadRequest, q)
	}
}

func TestDecodeActionIsStrict(t *testing.T) {
	good := map[string]hub.OperatorAction{
		`{"op":"prompt","prompt":"hi"}`: {Type: hub.OpPrompt, Prompt: "hi"},
		`{"op":"abort_turn"}`:           {Type: hub.OpAbortTurn},
		`{"op":"resolve_permission","request_id":"r1","decision":"allow_once"}`:           {Type: hub.OpResolvePermission, RequestID: "r1", Decision: agent.PermissionDecision{Kind: agent.DecisionAllowOnce}},
		`{"op":"resolve_permission","request_id":"r1","decision":"deny","feedback":"no"}`: {Type: hub.OpResolvePermission, RequestID: "r1", Decision: agent.PermissionDecision{Kind: agent.DecisionDeny, Feedback: "no"}},
		`{"op":"launch_missiles"}`: {Type: hub.OpUnknown},
	}
	for in, want := range good {
		got, err := decodeAction([]byte(in))
		if err != nil || got != want {
			t.Errorf("decodeAction(%s) = %+v, %v; want %+v", in, got, err, want)
		}
	}
	for _, in := range []string{
		``, `null`, `[]`, `"x"`, `{`, `{}`, `{"prompt":"x"}`, `{"op":""}`, `{"op":1}`,
		`{"op":"prompt","prompt":"x","role":"approver"}`,   // unknown field
		`{"op":"prompt","prompt":"x"} {"op":"abort_turn"}`, // trailing data
		`{"op":"prompt","op":"abort_turn"}`,
		`{"op":"resolve_permission","request_id":"r","decision":"allow_always"}`,
		`{"op":"resolve_permission","request_id":"r","decision":"allow_once","kind":"shell"}`,
	} {
		if _, err := decodeAction([]byte(in)); err == nil {
			t.Errorf("decodeAction(%q) accepted; want a protocol violation", in)
		}
	}
}

func TestProtocolViolationsCloseWithPolicyViolation(t *testing.T) {
	for name, msg := range map[string]string{
		"not json":      `hello`,
		"unknown field": `{"op":"abort_turn","admin":true}`,
		"missing op":    `{"prompt":"x"}`,
	} {
		t.Run(name, func(t *testing.T) {
			e := newEnv(t, envOpts{})
			c := dialWS(t, e, e.app.Token.Reveal())
			readEnvelope(t, c, 2*time.Second)
			sendText(t, c, msg)
			expectClose(t, c, websocket.StatusPolicyViolation, 3*time.Second)
		})
	}
	t.Run("binary", func(t *testing.T) {
		e := newEnv(t, envOpts{})
		c := dialWS(t, e, e.app.Token.Reveal())
		readEnvelope(t, c, 2*time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := c.Write(ctx, websocket.MessageBinary, []byte{1, 2, 3}); err != nil {
			t.Fatal(err)
		}
		expectClose(t, c, websocket.StatusPolicyViolation, 3*time.Second)
	})
}

func TestActionsDispatchThroughHubAndReplyToSenderOnly(t *testing.T) {
	fa := &fakeAgent{}
	e := newEnv(t, envOpts{hubCfg: hub.Config{Agent: fa}})
	app := dialWS(t, e, e.app.Token.Reveal())
	obs := dialWS(t, e, e.obs.Token.Reveal())
	readEnvelope(t, app, 2*time.Second)
	readEnvelope(t, obs, 2*time.Second)

	sendText(t, app, `{"op":"prompt","prompt":"do it"}`)
	if r := nextResult(t, app); r.Type != "action.accepted" || r.Op != "prompt" {
		t.Fatalf("approver prompt reply = %+v", r)
	}
	prompts, _ := fa.snapshot()
	if len(prompts) != 1 || prompts[0] != "do it" {
		t.Fatalf("agent prompts = %v", prompts)
	}

	// unknown op is refused with action_unknown and the connection stays open
	sendText(t, app, `{"op":"launch_missiles"}`)
	if r := nextResult(t, app); r.Type != "action.refused" || r.Reason != string(hub.ReasonActionUnknown) {
		t.Fatalf("unknown op reply = %+v", r)
	}
	// unknown request id
	sendText(t, app, `{"op":"resolve_permission","request_id":"nope","decision":"deny"}`)
	if r := nextResult(t, app); r.Reason != string(hub.ReasonRequestUnknown) || r.RequestID != "nope" {
		t.Fatalf("unknown request reply = %+v", r)
	}
	// the observer never saw any of those replies (refusals go to the sender only)
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	if _, data, err := obs.Read(ctx); err == nil {
		var r result
		if json.Unmarshal(data, &r) == nil && strings.HasPrefix(r.Type, "action.") {
			t.Fatalf("observer received another client's action reply: %s", data)
		}
	}
}

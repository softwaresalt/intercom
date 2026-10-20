package transport

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

func TestRevalidateMapsEachFailureToStableReason(t *testing.T) {
	clk := newFakeClockAt(time.Now())
	e := newEnv(t, envOpts{clock: clk})
	g, err := e.creds.Verify(e.app.Token.Reveal(), e.hub.Gen())
	if err != nil {
		t.Fatal(err)
	}
	if v := e.handler.revalidate(g, hub.RoleApprover); v.reason != "" || v.closeConn {
		t.Fatalf("fresh grant: %+v", v)
	}
	// role captured at connect is cross-checked against the credential record
	if v := e.handler.revalidate(g, hub.RoleObserver); v.reason == "" || !v.closeConn {
		t.Fatalf("role mismatch must refuse and close: %+v", v)
	}
	clk.Advance(2 * time.Hour)
	if v := e.handler.revalidate(g, hub.RoleApprover); v.reason != ReasonCredentialExpired || !v.closeConn {
		t.Fatalf("expired: %+v", v)
	}
	clk.Advance(-2 * time.Hour)
	e.creds.Revoke(g.CredentialID())
	if v := e.handler.revalidate(g, hub.RoleApprover); v.reason != string(hub.ReasonCredentialRevoked) || !v.closeConn {
		t.Fatalf("revoked: %+v", v)
	}
	if v := e.handler.revalidate(nil, hub.RoleApprover); v.reason == "" || !v.closeConn {
		t.Fatalf("nil grant must fail closed: %+v", v)
	}
}

func TestRevalidateWrongGenerationAndShutdown(t *testing.T) {
	e := newEnv(t, envOpts{})
	stale, err := NewCredentialStore(CredentialConfig{Gen: "stale-gen", Lifetime: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	is, _ := stale.Mint(hub.RoleApprover)
	g, err := stale.Verify(is.Token.Reveal(), "stale-gen")
	if err != nil {
		t.Fatal(err)
	}
	if v := e.handler.revalidate(g, hub.RoleApprover); v.reason != ReasonGenerationMismatch || !v.closeConn {
		t.Fatalf("wrong gen: %+v", v)
	}
	// shutdown revocation refuses WITHOUT closing, so the drain can finish
	g2, _ := e.creds.Verify(e.app.Token.Reveal(), e.hub.Gen())
	e.creds.RevokeAll()
	if v := e.handler.revalidate(g2, hub.RoleApprover); v.reason != string(hub.ReasonCredentialRevoked) || v.closeConn {
		t.Fatalf("shutdown revoke: %+v", v)
	}
}

func TestExpiredCredentialActionRefusedAndConnectionClosed(t *testing.T) {
	clk := newFakeClockAt(time.Now())
	fa := &fakeAgent{}
	e := newEnv(t, envOpts{clock: clk, hubCfg: hub.Config{Agent: fa}})
	c := dialWS(t, e, e.app.Token.Reveal())
	readEnvelope(t, c, 2*time.Second)

	sendText(t, c, `{"op":"prompt","prompt":"ok"}`)
	if r := nextResult(t, c); r.Type != ReplyAccepted {
		t.Fatalf("valid action: %+v", r)
	}
	clk.Advance(2 * time.Hour) // credential expires (the pump's timer is on the fake clock's horizon, so only re-validation can catch this)
	sendText(t, c, `{"op":"prompt","prompt":"after expiry"}`)
	r := nextResult(t, c)
	if r.Type != ReplyRefused || r.Reason != ReasonCredentialExpired {
		t.Fatalf("expired action reply = %+v", r)
	}
	expectClose(t, c, websocket.StatusPolicyViolation, 3*time.Second)
	if prompts, _ := fa.snapshot(); len(prompts) != 1 {
		t.Fatalf("agent received %v; the expired action must never reach it", prompts)
	}
}

func TestRevokedCredentialActionNeverReachesAgentAndConnectionCloses(t *testing.T) {
	fa := &fakeAgent{}
	e := newEnv(t, envOpts{hubCfg: hub.Config{Agent: fa}})
	c := dialWS(t, e, e.app.Token.Reveal())
	readEnvelope(t, c, 2*time.Second)
	e.creds.Revoke(e.app.ID)
	sendText(t, c, `{"op":"prompt","prompt":"too late"}`)
	// Either the pump's revoke signal or the per-action re-check wins the
	// race. The pump's revoke signal can legitimately close before the action
	// is even read, so a refusal is not guaranteed here (the deterministic
	// pump-level guarantee is TestPumpFlushesQueuedRefusalBeforeRevocationClose);
	// what IS guaranteed: any reply that does arrive is the credential_revoked
	// refusal, it precedes the 1008 "session ended" close, and the agent is
	// never reached.
	replies, code, reason := readToClose(t, c, 5*time.Second)
	for _, r := range replies {
		if r.Type != ReplyRefused || r.Reason != string(hub.ReasonCredentialRevoked) {
			t.Fatalf("reply before close = %+v; want only the credential_revoked refusal", r)
		}
	}
	if code != websocket.StatusPolicyViolation || reason != closeReasonEnded {
		t.Fatalf("close = %d %q; want 1008 %q", code, reason, closeReasonEnded)
	}
	if prompts, _ := fa.snapshot(); len(prompts) != 0 {
		t.Fatalf("revoked credential reached the agent: %v", prompts)
	}
}

func TestShutdownRevocationRefusesActionsButKeepsConnectionOpen(t *testing.T) {
	fa := &fakeAgent{}
	e := newEnv(t, envOpts{hubCfg: hub.Config{Agent: fa}})
	c := dialWS(t, e, e.app.Token.Reveal())
	readEnvelope(t, c, 2*time.Second)
	e.creds.RevokeAll()
	sendText(t, c, `{"op":"abort_turn"}`)
	r := nextResult(t, c)
	if r.Type != ReplyRefused || r.Reason != string(hub.ReasonCredentialRevoked) {
		t.Fatalf("reply = %+v", r)
	}
	if e.handler.ActiveSessions() != 1 {
		t.Fatal("shutdown revocation must not close the connection (drain must finish)")
	}
	if _, aborts := fa.snapshot(); aborts != 0 {
		t.Fatal("abort reached the agent")
	}
}

func TestObserverPrivilegedActionsRefusedWithReasonAndStayConnected(t *testing.T) {
	fa := &fakeAgent{}
	e := newEnv(t, envOpts{hubCfg: hub.Config{Agent: fa}})
	obs := dialWS(t, e, e.obs.Token.Reveal())
	app := dialWS(t, e, e.app.Token.Reveal())
	readEnvelope(t, obs, 2*time.Second)
	readEnvelope(t, app, 2*time.Second)

	// a real pending permission, so the refusal is the role check and not an
	// unknown-request lookup
	reply := make(chan agent.PermissionDecision, 1)
	resolved := make(chan struct{})
	ask := agent.PermissionAsk{
		Request:  agent.PermissionRequest{ToolCallID: "c1", Kind: agent.PermissionKindRead, RawKind: "read", Summary: "read x"},
		Reply:    reply,
		Outcome:  &agent.PermissionOutcome{},
		Resolved: resolved,
	}
	e.hub.Asks() <- ask
	var reqID string
	for reqID == "" {
		env := readEnvelope(t, obs, 3*time.Second)
		if env.Kind == hub.KindPermissionRequested {
			var p struct {
				RequestID string `json:"request_id"`
			}
			if err := json.Unmarshal(env.Payload, &p); err != nil || p.RequestID == "" {
				t.Fatalf("permission.requested payload: %v %s", err, env.Payload)
			}
			reqID = p.RequestID
		}
	}

	for _, msg := range []string{
		`{"op":"resolve_permission","request_id":"` + reqID + `","decision":"allow_once"}`,
		`{"op":"prompt","prompt":"x"}`,
		`{"op":"abort_turn"}`,
	} {
		sendText(t, obs, msg)
		r := nextResult(t, obs)
		if r.Type != ReplyRefused || r.Reason != string(hub.ReasonRoleNotPermitted) {
			t.Fatalf("%s => %+v; want role_not_permitted", msg, r)
		}
	}
	if prompts, aborts := fa.snapshot(); len(prompts) != 0 || aborts != 0 {
		t.Fatal("an observer action reached the agent")
	}
	select {
	case d := <-reply:
		t.Fatalf("observer resolved the permission: %+v", d)
	default:
	}
	if e.handler.ActiveSessions() != 2 {
		t.Fatal("a role refusal must not close the connection")
	}
	// the approver still can
	sendText(t, app, `{"op":"resolve_permission","request_id":"`+reqID+`","decision":"allow_once"}`)
	if r := nextResult(t, app); r.Type != ReplyAccepted {
		t.Fatalf("approver resolve reply = %+v", r)
	}
}

func TestInboundFloodIsRateLimitedThenClosed(t *testing.T) {
	clk := newFakeClockAt(time.Now())
	fa := &fakeAgent{}
	e := newEnv(t, envOpts{clock: clk, hubCfg: hub.Config{Agent: fa}, cfgTweak: func(c *Config) {
		c.ActionRate = 1
		c.ActionBurst = 3
		c.MaxRateViolations = 5
	}})
	c := dialWS(t, e, e.app.Token.Reveal())
	readEnvelope(t, c, 2*time.Second)

	accepted, limited := 0, 0
	for i := 0; i < 5; i++ {
		sendText(t, c, `{"op":"prompt","prompt":"p"}`)
		switch r := nextResult(t, c); {
		case r.Type == ReplyAccepted:
			accepted++
		case r.Reason == ReasonRateLimited:
			limited++
		default:
			t.Fatalf("unexpected reply %+v", r)
		}
	}
	if accepted != 3 || limited != 2 {
		t.Fatalf("accepted=%d limited=%d; want the burst of 3 then limiting", accepted, limited)
	}
	if prompts, _ := fa.snapshot(); len(prompts) != 3 {
		t.Fatalf("agent saw %d prompts; rate-limited ones must not reach it", len(prompts))
	}
	// refill: after a second one more is allowed again
	clk.Advance(1100 * time.Millisecond)
	sendText(t, c, `{"op":"prompt","prompt":"again"}`)
	if r := nextResult(t, c); r.Type != ReplyAccepted {
		t.Fatalf("after refill: %+v", r)
	}
	// persistent violation closes the connection
	for i := 0; i < 8; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		err := c.Write(ctx, websocket.MessageText, []byte(`{"op":"abort_turn"}`))
		cancel()
		if err != nil {
			break
		}
	}
	expectClose(t, c, websocket.StatusPolicyViolation, 3*time.Second)
	eventually(t, "session ended", 3*time.Second, func() bool { return e.handler.ActiveSessions() == 0 })
}

func TestTokenBucket(t *testing.T) {
	clk := newFakeClockAt(time.Unix(1000, 0))
	b := newTokenBucket(2, 4, clk.Now)
	for i := 0; i < 4; i++ {
		if !b.allow() {
			t.Fatalf("burst token %d denied", i)
		}
	}
	if b.allow() {
		t.Fatal("bucket must be empty after the burst")
	}
	clk.Advance(500 * time.Millisecond) // 1 token at 2/s
	if !b.allow() || b.allow() {
		t.Fatal("expected exactly one refilled token")
	}
	clk.Advance(time.Hour) // never exceeds burst
	n := 0
	for b.allow() {
		n++
	}
	if n != 4 {
		t.Fatalf("refill capped at burst: got %d", n)
	}
	// a clock going backwards must not mint tokens
	clk.Advance(-time.Hour)
	if b.allow() {
		t.Fatal("backwards clock minted a token")
	}
}

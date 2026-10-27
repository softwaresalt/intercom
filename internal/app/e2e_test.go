package app

// Headless end-to-end scenario tests (070.003-T): a fake agent behind the REAL
// hub and the REAL WebSocket server (assembled by Start), the Go reference
// client, and the real Bubble Tea TUI attached to the same hub through
// injected pipes. Scenario steps are those of the primary-scenario
// deliberation (MVP-D1) "The primary scenario":
//
//	step 4  the local TUI renders the streamed transcript; a remote client
//	        authenticates and sees the same stream
//	step 5  the operator sends a prompt (steer); the agent requests a
//	        permission; first responder wins (design 5.4); destructive remote
//	        approval is refused unless opted in (ADR MVP-3)
//	step 6  the remote client drops and reconnects with its (gen, seq) cursor
//	        and replays without gaps or duplicates (design 5.3)
//	step 7  the operator quits; shutdown runs the design 3.1 order and no CLI
//	        child remains
//
// Every wait is a bounded channel/frame read or a bounded poll of recorded
// state (waitFor); there are no fixed sleeps. The credentials used are the
// ones the process minted for itself; nothing here is a real secret.

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/hub"
	"github.com/softwaresalt/intercom-go/internal/transport/wsclient"
)

// ---- the fake agent with an observable prompt stream ----

// e2eAgent is the injected fake: no CLI child, every call recorded, and every
// prompt it receives is also published on sent so tests wait on a channel
// instead of sleeping.
type e2eAgent struct {
	*fakeAgent
	sent chan string
}

func (a *e2eAgent) Send(ctx context.Context, p string) error {
	err := a.fakeAgent.Send(ctx, p)
	select {
	case a.sent <- p:
	default: // 64 buffered prompts are far more than any scenario sends
	}
	return err
}

// runTurn plays the turn each accepted prompt starts: turn.start, turn.end,
// idle. Without it the hub (rightly) considers the session busy and shutdown
// would wait out its idle bound (design 3.1 step 3).
func (a *e2eAgent) runTurn(id string) {
	a.emit(agent.Event{Kind: agent.KindTurnStart, TurnID: id})
	a.emit(agent.Event{Kind: agent.KindTurnEnd, TurnID: id})
	a.emit(agent.Event{Kind: agent.KindIdle})
}

// awaitPrompt blocks (bounded) for the next prompt the agent received.
func (a *e2eAgent) awaitPrompt(t *testing.T) string {
	t.Helper()
	select {
	case p := <-a.sent:
		return p
	case <-time.After(waitFor):
		t.Fatal("the agent never received the prompt")
		return ""
	}
}

// ---- the permission request the fake agent raises ----

// e2ePerm plays the adapter's permission handler: it waits for the hub's one
// decision, delivers it to the outcome and acknowledges it (closing Resolved),
// which is what lets the hub publish permission.resolved (design 5.4 item 7).
type e2ePerm struct {
	ask     agent.PermissionAsk
	reply   chan agent.PermissionDecision
	outcome *agent.PermissionOutcome
	decided chan agent.PermissionDecision
	// gate, when non-nil, holds the handler's acknowledgement: the agent has
	// the decision (awaitDecision returns) but the hub has not yet published
	// permission.resolved until release. In that window the request is
	// decided-but-unacknowledged, which is where a loser's late action gets
	// already_resolved with the winner named.
	gate chan struct{}
}

func newE2EPerm(rec *recorder, callID string, kind agent.PermissionKind, raw, summary string, hold bool) *e2ePerm {
	p := &e2ePerm{
		reply:   make(chan agent.PermissionDecision, 1),
		outcome: &agent.PermissionOutcome{},
		decided: make(chan agent.PermissionDecision, 1),
	}
	if hold {
		p.gate = make(chan struct{})
	}
	resolved := make(chan struct{})
	p.ask = agent.PermissionAsk{
		Request:  agent.PermissionRequest{ToolCallID: callID, Kind: kind, RawKind: raw, Summary: summary},
		Reply:    p.reply,
		Outcome:  p.outcome,
		Resolved: resolved,
	}
	go func() {
		d := <-p.reply
		if p.gate != nil {
			p.decided <- d
			<-p.gate
		}
		rec.add("permission.decided")
		p.outcome.Deliver(d)
		close(resolved)
		if p.gate == nil {
			p.decided <- d // after the ack, so assertExactlyOne sees the outcome
		}
	}()
	return p
}

// release lets a held handler acknowledge its decision.
func (p *e2ePerm) release() { close(p.gate) }

// awaitDecision blocks (bounded) for the decision the agent received.
func (p *e2ePerm) awaitDecision(t *testing.T) agent.PermissionDecision {
	t.Helper()
	select {
	case d := <-p.decided:
		return d
	case <-time.After(waitFor):
		t.Fatal("the agent never received a permission decision")
		return agent.PermissionDecision{}
	}
}

// assertExactlyOne fails if the hub gave the agent more than the one decision
// already taken: nothing is queued on Reply and the outcome recorded `want`.
func (p *e2ePerm) assertExactlyOne(t *testing.T, want agent.DecisionKind) {
	t.Helper()
	if n := len(p.reply); n != 0 {
		t.Errorf("%d further decision(s) queued for the agent after the first (double resolve)", n)
	}
	d, ok := p.outcome.Delivered()
	if !ok || d.Kind != want {
		t.Errorf("outcome = (%v, %v); want exactly %v delivered", d.Kind, ok, want)
	}
}

// ---- the rig ----

type rigConfig struct {
	tables string           // extra config tables, e.g. "[remote]\n..."
	noTUI  bool             // run without the local TUI
	now    func() time.Time // injected clock
}

type e2eRig struct {
	t    *testing.T
	app  *App
	ag   *e2eAgent
	rec  *recorder
	sig  chan os.Signal
	out  *syncBuf
	keys io.Writer
	logs *logBuf
}

func newE2ERig(t *testing.T, rc rigConfig) *e2eRig {
	t.Helper()
	f := newFactory()
	o, ws := baseOptions(t, f)
	o.ConfigPath = writeConfig(t, ws, "", rc.tables)
	r := &e2eRig{t: t, rec: f.rec, sig: make(chan os.Signal, 2), out: &syncBuf{}, logs: &logBuf{}}
	o.NewAgent = func(d AgentDeps) (agent.Agent, error) {
		r.ag = &e2eAgent{fakeAgent: &fakeAgent{rec: r.rec, deps: d}, sent: make(chan string, 64)}
		return r.ag, nil
	}
	o.Signals = r.sig
	o.Logger = newTestLogger(r.logs)
	if rc.now != nil {
		o.Now = rc.now
	}
	if !rc.noTUI {
		in, inW := io.Pipe()
		t.Cleanup(func() { _ = inW.Close(); _ = in.Close() }) // registered before startApp: runs after its shutdown
		r.keys = inW
		o.TUI = TUIOptions{Input: in, Output: r.out, Width: 100, Height: 30}
	}
	r.app = startApp(t, o)
	if !rc.noTUI {
		waitOutput(t, r.out, "shown once") // the TUI is up and attached
		r.press("\x1b")                    // dismiss the credential panel so the transcript shows
	}
	return r
}

// press sends keys to the local TUI. io.Pipe is synchronous, so successive
// presses are never coalesced into one read.
func (r *e2eRig) press(s string) {
	r.t.Helper()
	if _, err := r.keys.Write([]byte(s)); err != nil {
		r.t.Fatalf("send keys: %v", err)
	}
}

// ask raises a permission request through the agent's real Asks channel.
func (r *e2eRig) ask(callID string, kind agent.PermissionKind, raw, summary string) *e2ePerm {
	return r.raise(newE2EPerm(r.rec, callID, kind, raw, summary, false))
}

// askHeld is ask with the handler's acknowledgement held until release.
func (r *e2eRig) askHeld(callID string, kind agent.PermissionKind, raw, summary string) *e2ePerm {
	return r.raise(newE2EPerm(r.rec, callID, kind, raw, summary, true))
}

func (r *e2eRig) raise(p *e2ePerm) *e2ePerm {
	r.ag.deps.Asks <- p.ask
	return p
}

// ---- the remote surface (reference client with a frame queue) ----

type surface struct {
	t     *testing.T
	c     *wsclient.Client
	seen  []wsclient.Message // every frame, in arrival order
	queue []wsclient.Message // frames not yet taken
	want  string             // what the current take waits for (failure message)
}

// connect dials the app as role and waits for the first frame (the hydrate on
// a first connect, the first replayed envelope on a resume).
func connect(t *testing.T, a *App, role hub.Role, resume *hub.Position) *surface {
	t.Helper()
	tok := credentialFor(t, a, role)
	ctx, cancel := context.WithTimeout(context.Background(), waitFor)
	defer cancel()
	c, err := wsclient.Dial(ctx, wsclient.Options{URL: a.WSURL(), Credential: tok, Resume: resume})
	if err != nil {
		t.Fatalf("dial as role %v: %v", role, err)
	}
	t.Cleanup(c.Abort)
	return &surface{t: t, c: c}
}

func credentialFor(t *testing.T, a *App, role hub.Role) string {
	t.Helper()
	for _, c := range a.Credentials() {
		if c.Role == role {
			return c.Token.Reveal()
		}
	}
	t.Fatalf("no credential minted for role %v", role)
	return ""
}

func (s *surface) read() wsclient.Message {
	s.t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), waitFor)
	defer cancel()
	m, err := s.c.Next(ctx)
	if err != nil {
		s.t.Fatalf("waiting for %s: %v (seen so far: %v)", s.want, err, s.kinds())
	}
	s.seen = append(s.seen, m)
	return m
}

// take returns the first frame (queued or newly read) satisfying pred; frames
// that do not are kept for later takes, so no frame is ever lost.
func (s *surface) take(what string, pred func(wsclient.Message) bool) wsclient.Message {
	s.t.Helper()
	for i, m := range s.queue {
		if pred(m) {
			s.queue = append(s.queue[:i:i], s.queue[i+1:]...)
			return m
		}
	}
	for {
		m := s.read()
		if pred(m) {
			return m
		}
		s.queue = append(s.queue, m)
		_ = what
	}
}

func (s *surface) kinds() []string {
	var out []string
	for _, m := range s.seen {
		switch {
		case m.Envelope != nil:
			out = append(out, string(m.Envelope.Kind))
		case m.Reply != nil:
			out = append(out, "reply:"+m.Reply.Type+"/"+m.Reply.Reason)
		}
	}
	return out
}

func (s *surface) hydrate() {
	s.take("session.hydrate", isKind(hub.KindSessionHydrate))
}

func isKind(k hub.Kind) func(wsclient.Message) bool {
	return func(m wsclient.Message) bool { return m.Envelope != nil && m.Envelope.Kind == k }
}

func isReply(m wsclient.Message) bool { return m.Reply != nil }

func decode[T any](t *testing.T, m wsclient.Message) T {
	t.Helper()
	var v T
	if m.Envelope == nil {
		t.Fatalf("frame %s is not an envelope", m.Raw)
	}
	if err := json.Unmarshal(m.Envelope.Payload, &v); err != nil {
		t.Fatalf("decode payload %s: %v", m.Envelope.Payload, err)
	}
	return v
}

func isMessageText(text string) func(wsclient.Message) bool {
	return func(m wsclient.Message) bool {
		if m.Envelope == nil || m.Envelope.Kind != hub.KindAgentMessage {
			return false
		}
		var p hub.MessagePayload
		return json.Unmarshal(m.Envelope.Payload, &p) == nil && p.Text == text
	}
}

func isRequestFor(callID string) func(wsclient.Message) bool {
	return func(m wsclient.Message) bool {
		if m.Envelope == nil || m.Envelope.Kind != hub.KindPermissionRequested {
			return false
		}
		var p hub.PermissionRequestedPayload
		return json.Unmarshal(m.Envelope.Payload, &p) == nil && p.ToolCallID == callID
	}
}

func ctxWait() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), waitFor)
}

// resolve sends a permission decision and returns the hub's reply to it.
func (s *surface) resolve(requestID string, allow bool) wsclient.Reply {
	s.t.Helper()
	ctx, cancel := ctxWait()
	defer cancel()
	if err := s.c.ResolvePermission(ctx, requestID, allow, ""); err != nil {
		s.t.Fatalf("send resolve_permission: %v", err)
	}
	return *s.take("action reply", isReply).Reply
}

func (s *surface) prompt(text string) wsclient.Reply {
	s.t.Helper()
	ctx, cancel := ctxWait()
	defer cancel()
	if err := s.c.Prompt(ctx, text); err != nil {
		s.t.Fatalf("send prompt: %v", err)
	}
	return *s.take("action reply", isReply).Reply
}

func (s *surface) abortTurn() wsclient.Reply {
	s.t.Helper()
	ctx, cancel := ctxWait()
	defer cancel()
	if err := s.c.AbortTurn(ctx); err != nil {
		s.t.Fatalf("send abort_turn: %v", err)
	}
	return *s.take("action reply", isReply).Reply
}

// drain reads until the connection ends and returns the close code.
func (s *surface) drain() (code int, reason string) {
	s.t.Helper()
	ctx, cancel := ctxWait()
	defer cancel()
	for {
		m, err := s.c.Next(ctx)
		if err != nil {
			code, reason, _ = wsclient.CloseStatus(err)
			return code, reason
		}
		s.seen = append(s.seen, m)
	}
}

// requestID returns the id of the permission request raised for callID.
func (s *surface) requestID(callID string) string {
	s.t.Helper()
	m := s.take("permission.requested "+callID, isRequestFor(callID))
	return decode[hub.PermissionRequestedPayload](s.t, m).RequestID
}

func wantReply(t *testing.T, got wsclient.Reply, accepted bool, reason hub.Reason) {
	t.Helper()
	if got.Accepted() != accepted || got.Reason != string(reason) {
		t.Fatalf("reply = %+v; want accepted=%v reason=%q", got, accepted, reason)
	}
}

// ---- (1) + (2): events and permission requests reach BOTH surfaces ----

// TestE2EStreamAndPermissionReachBothSurfaces covers scenario step 4 (the TUI
// renders the stream and the remote client sees it) and the permission
// request reaching every surface, observers included (ADR MVP-3).
func TestE2EStreamAndPermissionReachBothSurfaces(t *testing.T) {
	r := newE2ERig(t, rigConfig{})
	approver := connect(t, r.app, hub.RoleApprover, nil)
	observer := connect(t, r.app, hub.RoleObserver, nil)
	approver.hydrate()
	observer.hydrate()

	// Streamed deltas, then the completed message.
	deltas := []string{"alpha-", "beta-", "gamma"}
	for _, d := range deltas {
		r.ag.emit(agent.Event{Kind: agent.KindDelta, MessageID: "m1", Text: d})
	}
	full := strings.Join(deltas, "")
	r.ag.emit(agent.Event{Kind: agent.KindAgentMessage, MessageID: "m1", Text: full})

	for name, s := range map[string]*surface{"approver": approver, "observer": observer} {
		var got string
		for range deltas {
			m := s.take(name+" delta", isKind(hub.KindAgentMessageDelta))
			if m.Verdict != wsclient.VerdictApply {
				t.Errorf("%s: delta verdict %v; want apply", name, m.Verdict)
			}
			got += decode[hub.MessagePayload](t, m).Text
		}
		if got != full {
			t.Errorf("%s: streamed deltas = %q; want %q", name, got, full)
		}
		s.take(name+" message", isMessageText(full))
	}
	waitOutput(t, r.out, full) // the TUI rendered the same stream

	// A permission request reaches the TUI and both remote roles.
	perm := r.ask("call-both", agent.PermissionKindRead, "read", "read-notes-both.txt")
	var reqID string
	for name, s := range map[string]*surface{"approver": approver, "observer": observer} {
		m := s.take(name+" request", isRequestFor("call-both"))
		p := decode[hub.PermissionRequestedPayload](t, m)
		if p.Summary != "read-notes-both.txt" || p.RequestID == "" {
			t.Errorf("%s: request payload = %+v", name, p)
		}
		if name == "approver" {
			reqID = p.RequestID
		}
	}
	waitOutput(t, r.out, "Permission request")
	waitOutput(t, r.out, "read-notes-both.txt")

	// Finish cleanly: the approver answers.
	wantReply(t, approver.resolve(reqID, true), true, "")
	if d := perm.awaitDecision(t); d.Kind != agent.DecisionAllowOnce {
		t.Fatalf("decision = %v; want allow_once", d.Kind)
	}
}

// ---- (3): first responder wins ----

// TestE2EFirstResponder covers design 5.4: the first resolution wins, the
// loser's late action is refused (never double-resolving) and the agent sees
// exactly one decision. Losers are exercised deterministically: while the
// decision is held (before permission.resolved is published) a late action
// (a second remote connection, or the TUI arming and confirming its own
// approval) is refused with already_resolved naming the winner.
func TestE2EFirstResponder(t *testing.T) {
	t.Run("remote wins, TUI loses", func(t *testing.T) {
		r := newE2ERig(t, rigConfig{})
		first := connect(t, r.app, hub.RoleApprover, nil)
		second := connect(t, r.app, hub.RoleApprover, nil) // a second connection on the approver credential
		first.hydrate()
		second.hydrate()

		perm := r.askHeld("call-remote-wins", agent.PermissionKindRead, "read", "read-notes-remote.txt")
		id := first.requestID("call-remote-wins")
		waitOutput(t, r.out, "read-notes-remote.txt") // pending on the TUI too

		wantReply(t, first.resolve(id, true), true, "")
		if d := perm.awaitDecision(t); d.Kind != agent.DecisionAllowOnce {
			t.Fatalf("decision = %v; want allow_once", d.Kind)
		}
		// The agent holds the decision but has not acknowledged it: the request
		// is decided, so the loser's late action names the winner.
		late := second.resolve(id, false)
		wantReply(t, late, false, hub.ReasonAlreadyResolved)
		if late.Winner != "remote-approver" {
			t.Errorf("late action winner = %q; want remote-approver", late.Winner)
		}
		// The TUI is a loser too: while the decision is still held (before
		// permission.resolved is published) it arms and confirms its own late
		// approval, a real OpResolvePermission against the hub. The hub refuses
		// it with already_resolved naming the winner, and the TUI drops the
		// modal on that refusal, not on the (not yet published) resolution.
		waitOutput(t, r.out, "y approve (then Enter to confirm)") // modal fully seen: approvable
		r.press("y")
		waitOutput(t, r.out, "APPROVE? press Enter to confirm")
		r.press("\r")
		waitOutput(t, r.out, "request already answered by remote-approver")

		perm.release()
		for name, s := range map[string]*surface{"first": first, "second": second} {
			p := decode[hub.PermissionResolvedPayload](t, s.take(name+" resolved", isKind(hub.KindPermissionResolved)))
			if p.Actor != "remote-approver" || p.Decision != "allow_once" {
				t.Errorf("%s: resolved = %+v; want remote-approver/allow_once", name, p)
			}
		}
		waitOutput(t, r.out, "by remote-approver") // the TUI's modal cleared with the winner named

		// After the resolution is published the request is gone: still refused.
		wantReply(t, second.resolve(id, false), false, hub.ReasonRequestUnknown)
		r.press("y") // nothing pending any more: must not resolve anything
		r.press("\r")
		perm.assertExactlyOne(t, agent.DecisionAllowOnce)
	})

	t.Run("TUI wins, remote loses", func(t *testing.T) {
		r := newE2ERig(t, rigConfig{})
		remote := connect(t, r.app, hub.RoleApprover, nil)
		remote.hydrate()

		perm := r.askHeld("call-tui-wins", agent.PermissionKindRead, "read", "read-notes-tui.txt")
		id := remote.requestID("call-tui-wins")
		waitOutput(t, r.out, "y approve (then Enter to confirm)")
		r.press("y")
		waitOutput(t, r.out, "APPROVE? press Enter to confirm")
		r.press("\r")

		if d := perm.awaitDecision(t); d.Kind != agent.DecisionAllowOnce {
			t.Fatalf("decision = %v; want allow_once", d.Kind)
		}
		late := remote.resolve(id, true)
		wantReply(t, late, false, hub.ReasonAlreadyResolved)
		if late.Winner != "local-tui" {
			t.Errorf("late action winner = %q; want local-tui", late.Winner)
		}
		perm.release()
		p := decode[hub.PermissionResolvedPayload](t, remote.take("resolved", isKind(hub.KindPermissionResolved)))
		if p.Actor != "local-tui" || p.Decision != "allow_once" {
			t.Errorf("resolved = %+v; want local-tui/allow_once", p)
		}
		wantReply(t, remote.resolve(id, true), false, hub.ReasonRequestUnknown)
		perm.assertExactlyOne(t, agent.DecisionAllowOnce)
	})
}

// ---- (4): destructive remote approval ----

// TestE2EDestructiveRemoteApproval covers ADR MVP-3 with
// remote_approve_destructive = false (the default): shell, write, URL, other
// and unclassified kinds cannot be approved remotely (a remote DENY is always
// allowed), the request stays pending, and the TUI can still approve.
func TestE2EDestructiveRemoteApproval(t *testing.T) {
	kinds := []struct {
		name string
		kind agent.PermissionKind
		raw  string
	}{
		{"shell", agent.PermissionKindShell, "shell"},
		{"write", agent.PermissionKindWrite, "write"},
		{"url", agent.PermissionKindURL, "url"},
		{"other", agent.PermissionKindOther, "mcp"},
		{"unknown", agent.PermissionKindUnknown, "mystery"},
	}
	t.Run("remote approve refused, remote deny allowed", func(t *testing.T) {
		r := newE2ERig(t, rigConfig{noTUI: true})
		remote := connect(t, r.app, hub.RoleApprover, nil)
		remote.hydrate()
		for _, k := range kinds {
			call := "call-" + k.name
			perm := r.ask(call, k.kind, k.raw, "do-"+k.name)
			id := remote.requestID(call)
			wantReply(t, remote.resolve(id, true), false, hub.ReasonDestructiveRemoteRefused)
			// Still pending: a deny now resolves it and is the agent's only decision.
			wantReply(t, remote.resolve(id, false), true, "")
			if d := perm.awaitDecision(t); d.Kind != agent.DecisionDeny {
				t.Fatalf("%s: decision = %v; want deny", k.name, d.Kind)
			}
			perm.assertExactlyOne(t, agent.DecisionDeny)
		}
	})

	t.Run("TUI can still approve a destructive request", func(t *testing.T) {
		r := newE2ERig(t, rigConfig{})
		remote := connect(t, r.app, hub.RoleApprover, nil)
		remote.hydrate()
		perm := r.ask("call-tui-shell", agent.PermissionKindShell, "shell", "run-shell-e2e")
		id := remote.requestID("call-tui-shell")
		wantReply(t, remote.resolve(id, true), false, hub.ReasonDestructiveRemoteRefused)

		waitOutput(t, r.out, "y approve (then Enter to confirm)")
		r.press("y")
		waitOutput(t, r.out, "APPROVE? press Enter to confirm")
		r.press("\r")
		if d := perm.awaitDecision(t); d.Kind != agent.DecisionAllowOnce {
			t.Fatalf("decision = %v; want allow_once from the TUI", d.Kind)
		}
		perm.assertExactlyOne(t, agent.DecisionAllowOnce)
	})

	t.Run("opt-in permits remote approval", func(t *testing.T) {
		r := newE2ERig(t, rigConfig{noTUI: true, tables: "[remote]\nremote_approve_destructive = true\n"})
		remote := connect(t, r.app, hub.RoleApprover, nil)
		remote.hydrate()
		perm := r.ask("call-optin", agent.PermissionKindShell, "shell", "run-shell-optin")
		wantReply(t, remote.resolve(remote.requestID("call-optin"), true), true, "")
		if d := perm.awaitDecision(t); d.Kind != agent.DecisionAllowOnce {
			t.Fatalf("decision = %v; want allow_once", d.Kind)
		}
	})
}

// ---- (5): prompts from either surface reach the agent in order ----

// TestE2EPromptsReachAgentInOrder covers scenario step 5: the operator's
// prompt (steer) from the remote client and from the TUI reaches the agent,
// in the order sent; an abort reaches it too.
func TestE2EPromptsReachAgentInOrder(t *testing.T) {
	r := newE2ERig(t, rigConfig{})
	remote := connect(t, r.app, hub.RoleApprover, nil)
	remote.hydrate()

	wantReply(t, remote.prompt("remote-one"), true, "")
	if got := r.ag.awaitPrompt(t); got != "remote-one" {
		t.Fatalf("first prompt = %q; want remote-one", got)
	}
	r.ag.runTurn("t1")

	r.press("tui-two")
	waitOutput(t, r.out, "tui-two") // the typed text is on the input line
	r.press("\r")
	if got := r.ag.awaitPrompt(t); got != "tui-two" {
		t.Fatalf("second prompt = %q; want tui-two", got)
	}
	r.ag.runTurn("t2")

	wantReply(t, remote.prompt("remote-three"), true, "")
	if got := r.ag.awaitPrompt(t); got != "remote-three" {
		t.Fatalf("third prompt = %q; want remote-three", got)
	}
	r.ag.runTurn("t3")

	r.ag.mu.Lock()
	sends := append([]string(nil), r.ag.sends...)
	r.ag.mu.Unlock()
	if want := []string{"remote-one", "tui-two", "remote-three"}; !equalStrings(sends, want) {
		t.Fatalf("agent received %v; want %v", sends, want)
	}

	wantReply(t, remote.abortTurn(), true, "")
	r.rec.waitEvent(t, "agent.abort")
}

// ---- (6): reconnect with the stored cursor ----

// TestE2EReconnectResumesWithoutGapsOrDuplicates covers scenario step 6: the
// remote client drops, events are emitted while it is away, and a reconnect
// with its (gen, seq) cursor replays exactly those events once, in order,
// then continues live with no duplicate and no hydrate.
func TestE2EReconnectResumesWithoutGapsOrDuplicates(t *testing.T) {
	r := newE2ERig(t, rigConfig{noTUI: true})
	dropper := connect(t, r.app, hub.RoleApprover, nil)
	witness := connect(t, r.app, hub.RoleObserver, nil) // stays connected: proves the hub sequenced the events
	dropper.hydrate()
	witness.hydrate()

	msg := func(id, text string) {
		r.ag.emit(agent.Event{Kind: agent.KindAgentMessage, MessageID: id, Text: text})
	}
	msg("m1", "one")
	dropper.take("one", isMessageText("one"))
	pos, ok := dropper.c.Position()
	if !ok {
		t.Fatal("the client has no cursor after applying frames")
	}
	dropper.c.Abort() // the network drops

	for _, text := range []string{"two", "three", "four"} {
		msg("m-"+text, text)
	}
	witness.take("four", isMessageText("four")) // all three are in the hub's ring now

	back := connect(t, r.app, hub.RoleApprover, &pos)
	back.take("four", isMessageText("four"))
	msg("m5", "five")
	back.take("five", isMessageText("five"))

	if len(back.seen) == 0 || back.seen[0].Envelope == nil || back.seen[0].Envelope.Kind == hub.KindSessionHydrate {
		t.Fatalf("resume began with %v; want a replay of the missed envelopes, not a hydrate", back.kinds())
	}
	var texts []string
	next := pos.Seq + 1
	for _, m := range back.seen {
		if m.Envelope == nil {
			continue
		}
		if m.Verdict != wsclient.VerdictApply {
			t.Errorf("seq %d verdict %v; want apply (a gap or duplicate)", m.Envelope.Seq, m.Verdict)
		}
		if m.Envelope.Gen != pos.Gen || m.Envelope.Seq != next {
			t.Errorf("got gen/seq %s/%d; want %s/%d (gap or duplicate)", m.Envelope.Gen, m.Envelope.Seq, pos.Gen, next)
		}
		next = m.Envelope.Seq + 1
		if m.Envelope.Kind == hub.KindAgentMessage {
			texts = append(texts, decode[hub.MessagePayload](t, m).Text)
		}
	}
	if want := []string{"two", "three", "four", "five"}; !equalStrings(texts, want) {
		t.Fatalf("replayed+live messages = %v; want %v exactly once each", texts, want)
	}
}

// ---- (7): authentication rejections ----

// upgrade performs a raw WebSocket upgrade request with full control of Host
// and headers, returning the HTTP status and WWW-Authenticate header.
func upgrade(t *testing.T, a *App, host string, hdr map[string]string) (int, string) {
	t.Helper()
	ctx, cancel := ctxWait()
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://"+a.Addr().String()+"/ws", nil)
	if err != nil {
		t.Fatal(err)
	}
	if host != "" {
		req.Host = host
	}
	req.Header.Set("Connection", "Upgrade")
	req.Header.Set("Upgrade", "websocket")
	req.Header.Set("Sec-WebSocket-Version", "13")
	req.Header.Set("Sec-WebSocket-Key", "AAAAAAAAAAAAAAAAAAAAAA==")
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	resp, err := (&http.Client{Transport: &http.Transport{DisableKeepAlives: true}}).Do(req)
	if err != nil {
		t.Fatalf("raw upgrade: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	return resp.StatusCode, resp.Header.Get("WWW-Authenticate")
}

func wantRefusal(t *testing.T, err error, status int) {
	t.Helper()
	he, ok := err.(*wsclient.HandshakeError) //nolint:errorlint // the reference client returns it unwrapped
	if !ok || he.Status != status {
		t.Fatalf("dial error = %v; want a %d handshake refusal", err, status)
	}
}

// TestE2EAuthRejections covers the documented rejections (wsclient package
// documentation; design 6): each credential failure is the same bare 401,
// bad Host/Origin are 403, an observer's privileged actions are refused with
// role_not_permitted, and an expired or re-issued credential ends the session.
func TestE2EAuthRejections(t *testing.T) {
	dial := func(r *e2eRig, o wsclient.Options) error {
		ctx, cancel := ctxWait()
		defer cancel()
		o.URL = r.app.WSURL()
		c, err := wsclient.Dial(ctx, o)
		if err == nil {
			c.Abort()
		}
		return err
	}

	t.Run("no credential", func(t *testing.T) {
		r := newE2ERig(t, rigConfig{noTUI: true})
		status, www := upgrade(t, r.app, "", nil)
		if status != http.StatusUnauthorized || www != "Bearer" {
			t.Fatalf("no credential: status %d, WWW-Authenticate %q; want 401 Bearer", status, www)
		}
	})

	t.Run("empty and wrong credentials", func(t *testing.T) {
		r := newE2ERig(t, rigConfig{noTUI: true})
		real := credentialFor(t, r.app, hub.RoleApprover)
		flipped := []byte(real)
		if flipped[len(flipped)-1] == 'A' {
			flipped[len(flipped)-1] = 'B'
		} else {
			flipped[len(flipped)-1] = 'A'
		}
		for name, tok := range map[string]string{
			"empty":        "",
			"wrong":        strings.Repeat("A", len(real)),
			"one char off": string(flipped),
			"truncated":    real[:len(real)-1],
			"not a token":  "not a credential",
			"doubled":      real + real,
		} {
			wantRefusal(t, dial(r, wsclient.Options{Credential: tok}), 401)
			t.Logf("%s credential refused with 401", name)
		}
		wantRefusal(t, dial(r, wsclient.Options{Credential: strings.Repeat("A", len(real)), Carriage: wsclient.CarriageSubprotocol}), 401)
	})

	t.Run("bad Origin", func(t *testing.T) {
		r := newE2ERig(t, rigConfig{noTUI: true})
		tok := credentialFor(t, r.app, hub.RoleApprover)
		// allowed_origins is empty by default: every browser origin is refused
		// even with a valid credential.
		wantRefusal(t, dial(r, wsclient.Options{Credential: tok, Origin: "https://evil.example"}), 403)
		wantRefusal(t, dial(r, wsclient.Options{Credential: tok, Origin: "https://evil.example", Carriage: wsclient.CarriageSubprotocol}), 403)
		// A native client (no Origin) with the same credential is accepted.
		if err := dial(r, wsclient.Options{Credential: tok}); err != nil {
			t.Fatalf("valid native dial refused: %v", err)
		}
	})

	t.Run("bad Host", func(t *testing.T) {
		r := newE2ERig(t, rigConfig{noTUI: true})
		tok := credentialFor(t, r.app, hub.RoleApprover)
		status, _ := upgrade(t, r.app, "evil.example", map[string]string{"Authorization": "Bearer " + tok})
		if status != http.StatusForbidden {
			t.Fatalf("bad Host with a valid credential: status %d; want 403", status)
		}
		status, _ = upgrade(t, r.app, "", map[string]string{"Authorization": "Bearer " + tok})
		if status != http.StatusSwitchingProtocols {
			t.Fatalf("control: the loopback Host with the same credential: status %d; want 101", status)
		}
	})

	t.Run("observer cannot approve, prompt or abort", func(t *testing.T) {
		r := newE2ERig(t, rigConfig{noTUI: true})
		approver := connect(t, r.app, hub.RoleApprover, nil)
		observer := connect(t, r.app, hub.RoleObserver, nil)
		approver.hydrate()
		observer.hydrate()
		perm := r.ask("call-observer", agent.PermissionKindRead, "read", "read-observer.txt")
		id := observer.requestID("call-observer") // observers do see the request

		wantReply(t, observer.resolve(id, true), false, hub.ReasonRoleNotPermitted)
		wantReply(t, observer.prompt("observer-prompt"), false, hub.ReasonRoleNotPermitted)
		wantReply(t, observer.abortTurn(), false, hub.ReasonRoleNotPermitted)

		// Nothing reached the agent, the request is still pending, and the
		// observer's socket is still open (a refusal is not a violation).
		r.ag.mu.Lock()
		sends := len(r.ag.sends)
		r.ag.mu.Unlock()
		if sends != 0 {
			t.Errorf("the agent received %d prompt(s) from an observer", sends)
		}
		for _, e := range r.rec.snapshot() {
			if e == "agent.abort" || e == "permission.decided" {
				t.Errorf("observer action had an effect: %v", r.rec.snapshot())
			}
		}
		wantReply(t, approver.resolve(id, true), true, "")
		if d := perm.awaitDecision(t); d.Kind != agent.DecisionAllowOnce {
			t.Fatalf("decision = %v; want the approver's allow_once", d.Kind)
		}
	})

	t.Run("revoked after re-issue", func(t *testing.T) {
		r := newE2ERig(t, rigConfig{noTUI: true})
		old := credentialFor(t, r.app, hub.RoleApprover)
		live := connect(t, r.app, hub.RoleApprover, nil)
		live.hydrate()

		ctx, cancel := ctxWait()
		defer cancel()
		if _, err := r.app.Reissue(ctx); err != nil {
			t.Fatalf("Reissue: %v", err)
		}
		code, reason := live.drain()
		if code != 1008 || reason != "session ended" {
			t.Errorf("the revoked credential's socket closed with %d %q; want 1008 \"session ended\"", code, reason)
		}
		wantRefusal(t, dial(r, wsclient.Options{Credential: old}), 401)
		if err := dial(r, wsclient.Options{Credential: credentialFor(t, r.app, hub.RoleApprover)}); err != nil {
			t.Fatalf("the re-issued credential was refused: %v", err)
		}
	})

	t.Run("expired credential", func(t *testing.T) {
		var offset atomic.Int64
		base := time.Now()
		clock := func() time.Time { return base.Add(time.Duration(offset.Load())) }
		r := newE2ERig(t, rigConfig{noTUI: true, now: clock})
		live := connect(t, r.app, hub.RoleApprover, nil)
		live.hydrate()
		tok := credentialFor(t, r.app, hub.RoleApprover)

		life := time.Duration(r.app.Config().Remote.TokenLifetimeSeconds) * time.Second //nolint:gosec // validated range
		offset.Store(int64(life + time.Second))

		wantRefusal(t, dial(r, wsclient.Options{Credential: tok}), 401) // a new dial: the same bare 401
		// The already-open socket: its next action is refused as expired and
		// the session ends; the prompt never reaches the agent.
		wantReply(t, live.prompt("late-prompt"), false, hub.Reason("credential_expired"))
		if code, reason := live.drain(); code != 1008 || reason != "session ended" {
			t.Errorf("expired socket closed with %d %q; want 1008 \"session ended\"", code, reason)
		}
		r.ag.mu.Lock()
		defer r.ag.mu.Unlock()
		if len(r.ag.sends) != 0 {
			t.Errorf("an expired credential's prompt reached the agent: %v", r.ag.sends)
		}
	})
}

// ---- (8): clean shutdown ----

// shutdownScenario runs the design 3.1 shutdown with a TUI attached, a remote
// client connected and a permission request pending, started by trigger.
func shutdownScenario(t *testing.T, trigger func(*e2eRig)) {
	baseline := ownGoroutines()
	r := newE2ERig(t, rigConfig{})
	var cur atomic.Pointer[App]
	var pr probes
	r.ag.abortHook = func() { pr.set("abort", takeProbe(cur.Load())) }
	r.ag.stopHook = func(context.Context) { pr.set("stop", takeProbe(cur.Load())) }
	cur.Store(r.app)
	watchHubStop(r.rec, r.ag.deps.HubDone)

	remote := connect(t, r.app, hub.RoleApprover, nil)
	remote.hydrate()
	perm := r.ask("call-shutdown", agent.PermissionKindRead, "read", "read-shutdown.txt")
	remote.requestID("call-shutdown")
	waitOutput(t, r.out, "read-shutdown.txt") // pending on both surfaces

	trigger(r)

	select {
	case <-r.app.Done():
	case <-time.After(waitFor):
		t.Fatal("the process did not finish shutting down")
	}
	if err := r.app.Err(); err != nil {
		t.Fatalf("shutdown reported errors: %v", err)
	}
	if d := perm.awaitDecision(t); d.Kind != agent.DecisionUnavailable {
		t.Errorf("pending permission resolved to %v; want UserNotAvailable first", d.Kind)
	}
	r.rec.waitEvent(t, "hub.stopped")
	want := []string{"agent.start", "permission.decided", "agent.abort", "agent.stop", "hub.stopped"}
	if got := r.rec.snapshot(); !equalStrings(got, want) {
		t.Fatalf("shutdown order = %v; want %v", got, want)
	}
	for _, label := range []string{"abort", "stop"} {
		p, ok := pr.get(label)
		if !ok || !p.listenerRefused || !p.shuttingDown || !p.hubRunning {
			t.Errorf("at agent %s: %+v (taken=%v); want listener closed, hub shutting down and still running", label, p, ok)
		}
	}

	remote.drain()
	kinds := remote.kinds()
	if r, s := indexOf(kinds, string(hub.KindPermissionResolved)), indexOf(kinds, string(hub.KindServerShutdown)); r < 0 || s < 0 || r > s {
		t.Errorf("remote client saw %v; want permission.resolved before server.shutdown", kinds)
	}
	select {
	case <-r.app.ui.done:
	default:
		t.Error("the TUI program is still running after shutdown")
	}
	// The fake stands in for the CLI child: started once, stopped once.
	if starts, stops := r.ag.counts(); starts != 1 || stops != 1 {
		t.Errorf("agent starts/stops = %d/%d; want 1/1 (no stranded runtime)", starts, stops)
	}
	assertNoLeaks(t, baseline)
}

// TestE2EShutdown covers scenario step 7: the operator quits (SIGINT, or Ctrl+C
// in the TUI) and shutdown runs the design 3.1 order; the pending permission
// is resolved UserNotAvailable first, server.shutdown is broadcast, the agent
// is aborted and stopped, the hub stops last, and nothing leaks.
func TestE2EShutdown(t *testing.T) {
	t.Run("SIGINT", func(t *testing.T) {
		shutdownScenario(t, func(r *e2eRig) { r.sig <- os.Interrupt })
	})
	t.Run("TUI quit", func(t *testing.T) {
		shutdownScenario(t, func(r *e2eRig) { r.press("\x03") })
	})
}

var _ sync.Locker = (*sync.Mutex)(nil) // keep the sync import honest if helpers move

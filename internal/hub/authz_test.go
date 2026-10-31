package hub

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

// ---- 067.004-T: authorization policy enforced in the hub's arbitration ----

var allKinds = []agent.PermissionKind{
	agent.PermissionKindUnknown, agent.PermissionKindShell, agent.PermissionKindWrite,
	agent.PermissionKindRead, agent.PermissionKindURL, agent.PermissionKindOther,
	agent.PermissionKind(200), // a value no adapter defines
}

func TestAuthorizeMatrixRows(t *testing.T) {
	tui := principal{surface: surfaceTUI, role: RoleApprover}
	appr := principal{surface: surfaceRemote, role: RoleApprover}
	obs := principal{surface: surfaceRemote, role: RoleObserver}
	none := principal{surface: surfaceRemote, role: RoleNone}
	allow, deny := agent.DecisionAllowOnce, agent.DecisionDeny
	rows := []struct {
		name string
		in   authzInput
		ok   bool
		why  Reason
	}{
		{"tui prompt", authzInput{p: tui, op: OpPrompt}, true, ""},
		{"tui abort", authzInput{p: tui, op: OpAbortTurn}, true, ""},
		{"tui approves shell with remote opt-in off", authzInput{p: tui, op: OpResolvePermission, kind: agent.PermissionKindShell, decision: allow}, true, ""},
		{"approver prompt", authzInput{p: appr, op: OpPrompt}, true, ""},
		{"approver abort", authzInput{p: appr, op: OpAbortTurn}, true, ""},
		{"approver approves read", authzInput{p: appr, op: OpResolvePermission, kind: agent.PermissionKindRead, decision: allow}, true, ""},
		{"approver denies shell", authzInput{p: appr, op: OpResolvePermission, kind: agent.PermissionKindShell, decision: deny}, true, ""},
		{"approver approves shell, default", authzInput{p: appr, op: OpResolvePermission, kind: agent.PermissionKindShell, decision: allow}, false, ReasonDestructiveRemoteRefused},
		{"approver approves write, default", authzInput{p: appr, op: OpResolvePermission, kind: agent.PermissionKindWrite, decision: allow}, false, ReasonDestructiveRemoteRefused},
		{"approver approves url, default", authzInput{p: appr, op: OpResolvePermission, kind: agent.PermissionKindURL, decision: allow}, false, ReasonDestructiveRemoteRefused},
		{"approver approves unknown, default", authzInput{p: appr, op: OpResolvePermission, kind: agent.PermissionKindUnknown, decision: allow}, false, ReasonDestructiveRemoteRefused},
		{"approver approves other, default", authzInput{p: appr, op: OpResolvePermission, kind: agent.PermissionKindOther, decision: allow}, false, ReasonDestructiveRemoteRefused},
		{"approver approves shell, opt-in", authzInput{p: appr, op: OpResolvePermission, kind: agent.PermissionKindShell, decision: allow, remoteApproveDestructive: true}, true, ""},
		// ADR MVP-3 (074-S amendment): unknown/unmodeled kinds are local-TUI-only
		// for approvals; the remote_approve_destructive opt-in never unlocks
		// them. Remote deny stays allowed.
		{"approver approves unknown, opt-in", authzInput{p: appr, op: OpResolvePermission, kind: agent.PermissionKindUnknown, decision: allow, remoteApproveDestructive: true}, false, ReasonDestructiveRemoteRefused},
		{"approver approves other, opt-in", authzInput{p: appr, op: OpResolvePermission, kind: agent.PermissionKindOther, decision: allow, remoteApproveDestructive: true}, false, ReasonDestructiveRemoteRefused},
		{"approver approves url, opt-in", authzInput{p: appr, op: OpResolvePermission, kind: agent.PermissionKindURL, decision: allow, remoteApproveDestructive: true}, true, ""},
		{"approver approves write, opt-in", authzInput{p: appr, op: OpResolvePermission, kind: agent.PermissionKindWrite, decision: allow, remoteApproveDestructive: true}, true, ""},
		{"approver approves undefined kind, opt-in", authzInput{p: appr, op: OpResolvePermission, kind: agent.PermissionKind(200), decision: allow, remoteApproveDestructive: true}, false, ReasonDestructiveRemoteRefused},
		{"approver approves undefined kind, default", authzInput{p: appr, op: OpResolvePermission, kind: agent.PermissionKind(200), decision: allow}, false, ReasonDestructiveRemoteRefused},
		{"approver denies unknown, default", authzInput{p: appr, op: OpResolvePermission, kind: agent.PermissionKindUnknown, decision: deny}, true, ""},
		{"approver denies other, opt-in", authzInput{p: appr, op: OpResolvePermission, kind: agent.PermissionKindOther, decision: deny, remoteApproveDestructive: true}, true, ""},
		{"approver denies undefined kind, opt-in", authzInput{p: appr, op: OpResolvePermission, kind: agent.PermissionKind(200), decision: deny, remoteApproveDestructive: true}, true, ""},
		{"tui approves unknown, opt-in off", authzInput{p: tui, op: OpResolvePermission, kind: agent.PermissionKindUnknown, decision: allow}, true, ""},
		{"tui approves undefined kind, opt-in off", authzInput{p: tui, op: OpResolvePermission, kind: agent.PermissionKind(200), decision: allow}, true, ""},
		{"observer approves unknown, opt-in", authzInput{p: obs, op: OpResolvePermission, kind: agent.PermissionKindUnknown, decision: allow, remoteApproveDestructive: true}, false, ReasonRoleNotPermitted},
		{"observer prompt", authzInput{p: obs, op: OpPrompt}, false, ReasonRoleNotPermitted},
		{"observer abort", authzInput{p: obs, op: OpAbortTurn}, false, ReasonRoleNotPermitted},
		{"observer approves read", authzInput{p: obs, op: OpResolvePermission, kind: agent.PermissionKindRead, decision: allow}, false, ReasonRoleNotPermitted},
		{"observer denies read", authzInput{p: obs, op: OpResolvePermission, kind: agent.PermissionKindRead, decision: deny}, false, ReasonRoleNotPermitted},
		{"observer approves shell, opt-in", authzInput{p: obs, op: OpResolvePermission, kind: agent.PermissionKindShell, decision: allow, remoteApproveDestructive: true}, false, ReasonRoleNotPermitted},
		{"no-role remote prompt", authzInput{p: none, op: OpPrompt}, false, ReasonRoleNotPermitted},
		{"unknown op, tui", authzInput{p: tui, op: OpUnknown}, false, ReasonActionUnknown},
		{"unknown op, approver", authzInput{p: appr, op: OpUnknown}, false, ReasonActionUnknown},
		{"garbage op", authzInput{p: tui, op: OpType(250)}, false, ReasonActionUnknown},
	}
	for _, r := range rows {
		ok, why := authorize(r.in)
		if ok != r.ok || why != r.why {
			t.Errorf("%s: authorize = (%v, %q), want (%v, %q)", r.name, ok, why, r.ok, r.why)
		}
	}
}

// Invariants over the whole input space, so a kind or role nobody listed in a
// row cannot slip through.
func TestAuthorizeRefusesRemoteApproveOfClippedWrite(t *testing.T) {
	tui := principal{surface: surfaceTUI, role: RoleApprover}
	appr := principal{surface: surfaceRemote, role: RoleApprover}
	obs := principal{surface: surfaceRemote, role: RoleObserver}
	allow, deny := agent.DecisionAllowOnce, agent.DecisionDeny
	for _, optIn := range []bool{false, true} {
		rows := []struct {
			name string
			p    principal
			dec  agent.DecisionKind
			ok   bool
			why  Reason
		}{
			{"remote approver allows", appr, allow, false, ReasonDestructiveRemoteRefused},
			{"remote approver denies", appr, deny, true, ""},
			{"local tui allows", tui, allow, true, ""},
			{"remote observer allows", obs, allow, false, ReasonRoleNotPermitted},
			{"remote observer denies", obs, deny, false, ReasonRoleNotPermitted},
		}
		for _, r := range rows {
			in := authzInput{p: r.p, op: OpResolvePermission, kind: agent.PermissionKindWrite, decision: r.dec, remoteApproveDestructive: optIn, previewClipped: true}
			if ok, why := authorize(in); ok != r.ok || why != r.why {
				t.Errorf("%s (optIn=%v): authorize = (%v, %q), want (%v, %q)", r.name, optIn, ok, why, r.ok, r.why)
			}
		}
	}
}

func TestRemoteClippedWriteApprovalRefusedEvenWhenOptedIn(t *testing.T) {
	h := startHub(t, Config{RemoteApproveDestructive: true})
	tui := connectRole(t, h, ClientTUI, 0)
	appr := connectRole(t, h, ClientRemote, RoleApprover)
	a := newTestAsk(agent.PermissionKindWrite, "f.go\n+x\n[write preview truncated]")
	a.ask.Request.PreviewClipped = true
	a.post(t, h)
	id := requestID(t, recvKind(t, tui, KindPermissionRequested))
	if res := do(t, appr, answer(id, allowOnce)); res.Accepted || res.Reason != ReasonDestructiveRemoteRefused {
		t.Fatalf("remote approval of a clipped write = %+v, want destructive_remote_refused", res)
	}
	select {
	case d := <-a.reply:
		t.Fatalf("a refused approval reached the agent: %v", d.Kind)
	default:
	}
	if res := do(t, tui, answer(id, allowOnce)); !res.Accepted {
		t.Fatalf("local tui approval = %+v", res)
	}
}

func TestAuthorizeInvariantsOverTheWholeInputSpace(t *testing.T) {
	surfaces := []surface{0, surfaceRemote, surfaceTUI, 9}
	roles := []Role{RoleNone, RoleObserver, RoleApprover, 9}
	ops := []OpType{OpUnknown, OpResolvePermission, OpPrompt, OpAbortTurn, 77}
	decisions := []agent.DecisionKind{0, agent.DecisionAllowOnce, agent.DecisionDeny, agent.DecisionUnavailable, 99}
	for _, sf := range surfaces {
		for _, role := range roles {
			for _, op := range ops {
				for _, kind := range allKinds {
					for _, dec := range decisions {
						for _, flag := range []bool{false, true} {
							in := authzInput{p: principal{surface: sf, role: role}, op: op, kind: kind, decision: dec, remoteApproveDestructive: flag}
							ok, why := authorize(in)
							if ok && why != "" || !ok && why == "" {
								t.Fatalf("%+v: ok=%v reason=%q are inconsistent", in, ok, why)
							}
							known := op == OpResolvePermission || op == OpPrompt || op == OpAbortTurn
							switch {
							case !known:
								if ok {
									t.Fatalf("%+v: unknown action allowed", in)
								}
							case sf == surfaceTUI:
								if !ok {
									t.Fatalf("%+v: the local TUI must always be an approver", in)
								}
							case sf != surfaceRemote || role != RoleApprover:
								if ok {
									t.Fatalf("%+v: a non-approver was allowed", in)
								}
							case op == OpResolvePermission && dec != agent.DecisionDeny && kind != agent.PermissionKindRead &&
								(!flag || (kind != agent.PermissionKindShell && kind != agent.PermissionKindWrite && kind != agent.PermissionKindURL)):
								if ok {
									t.Fatalf("%+v: remote destructive approval allowed without opt-in or for an unknown kind", in)
								}
								if why != ReasonDestructiveRemoteRefused {
									t.Fatalf("%+v: reason %q", in, why)
								}
							default:
								if !ok {
									t.Fatalf("%+v: approver action refused (%s)", in, why)
								}
							}
						}
					}
				}
			}
		}
	}
}

// connectRole registers a client with an explicit role and discards its
// hydrate.
func connectRole(t *testing.T, h *Hub, kind ClientKind, role Role) *Client {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), testWait)
	defer cancel()
	c, err := h.Connect(ctx, ConnectOptions{Kind: kind, Role: role, Buffer: 64})
	if err != nil {
		t.Fatalf("Connect: %v", err)
	}
	recvKind(t, c, KindSessionHydrate)
	return c
}

func do(t *testing.T, c *Client, act OperatorAction) ActionResult {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), testWait)
	defer cancel()
	res, err := c.Do(ctx, act)
	if err != nil {
		t.Fatalf("Do: %v", err)
	}
	return res
}

func answer(id string, d agent.PermissionDecision) OperatorAction {
	return OperatorAction{Type: OpResolvePermission, RequestID: id, Decision: d}
}

// pendingOf posts an ask of kind and returns its request id (read from the
// observer's broadcast, as a real client would).
func pendingOf(t *testing.T, h *Hub, watch *Client, kind agent.PermissionKind) (string, *testAsk) {
	t.Helper()
	a := newTestAsk(kind, "x")
	a.post(t, h)
	return requestID(t, recvKind(t, watch, KindPermissionRequested)), a
}

func TestObserverCannotApproveOrDenyAndDoesNotConsumeTheRequest(t *testing.T) {
	h := startHub(t, Config{RemoteApproveDestructive: true})
	tui := connectRole(t, h, ClientTUI, 0)
	obs := connectRole(t, h, ClientRemote, RoleObserver)
	id, a := pendingOf(t, h, tui, agent.PermissionKindRead)

	for _, d := range []agent.PermissionDecision{allowOnce, denyDecision} {
		res := do(t, obs, answer(id, d))
		if res.Accepted || res.Reason != ReasonRoleNotPermitted {
			t.Fatalf("observer %v = %+v, want role_not_permitted", d.Kind, res)
		}
	}
	select {
	case d := <-a.reply:
		t.Fatalf("a refused action reached the agent: %v", d.Kind)
	default:
	}
	// The refusal did not resolve, consume or time out the request: the local
	// TUI can still answer it (ADR MVP-3).
	if res := do(t, tui, answer(id, allowOnce)); !res.Accepted {
		t.Fatalf("tui after an observer refusal = %+v", res)
	}
}

func TestRemoteShellAndWriteApprovalRefusedByDefaultWithReason(t *testing.T) {
	for _, kind := range []agent.PermissionKind{agent.PermissionKindShell, agent.PermissionKindWrite} {
		h := startHub(t, Config{})
		tui := connectRole(t, h, ClientTUI, 0)
		appr := connectRole(t, h, ClientRemote, RoleApprover)
		id, a := pendingOf(t, h, tui, kind)

		res := do(t, appr, answer(id, allowOnce))
		if res.Accepted || res.Reason != ReasonDestructiveRemoteRefused {
			t.Fatalf("kind %v: remote approval = %+v, want destructive_remote_refused", kind, res)
		}
		select {
		case d := <-a.reply:
			t.Fatalf("a refused approval reached the agent: %v", d.Kind)
		default:
		}
		// Denial is always safe: the same remote approver may deny.
		if res := do(t, appr, answer(id, denyDecision)); !res.Accepted {
			t.Fatalf("kind %v: remote deny = %+v", kind, res)
		}
		if d := a.takeReply(t); d.Kind != agent.DecisionDeny {
			t.Fatalf("agent got %v", d.Kind)
		}
	}
}

func TestRemoteDestructiveApprovalWhenOptedIn(t *testing.T) {
	h := startHub(t, Config{RemoteApproveDestructive: true})
	tui := connectRole(t, h, ClientTUI, 0)
	appr := connectRole(t, h, ClientRemote, RoleApprover)
	id, a := pendingOf(t, h, tui, agent.PermissionKindShell)
	if res := do(t, appr, answer(id, allowOnce)); !res.Accepted {
		t.Fatalf("opted-in remote approval = %+v", res)
	}
	if d := a.takeReply(t); d.Kind != agent.DecisionAllowOnce {
		t.Fatalf("agent got %v", d.Kind)
	}
}

func TestRemoteApproverMayApproveReadByDefault(t *testing.T) {
	h := startHub(t, Config{})
	tui := connectRole(t, h, ClientTUI, 0)
	appr := connectRole(t, h, ClientRemote, RoleApprover)
	id, _ := pendingOf(t, h, tui, agent.PermissionKindRead)
	if res := do(t, appr, answer(id, allowOnce)); !res.Accepted {
		t.Fatalf("remote read approval = %+v", res)
	}
}

func TestUnknownPermissionKindsAreDeniedToRemoteApprovers(t *testing.T) {
	for _, optIn := range []bool{false, true} {
		for _, kind := range []agent.PermissionKind{agent.PermissionKindUnknown, agent.PermissionKindOther, agent.PermissionKind(200)} {
			h := startHub(t, Config{RemoteApproveDestructive: optIn})
			tui := connectRole(t, h, ClientTUI, 0)
			appr := connectRole(t, h, ClientRemote, RoleApprover)
			id, _ := pendingOf(t, h, tui, kind)
			if res := do(t, appr, answer(id, allowOnce)); res.Accepted || res.Reason != ReasonDestructiveRemoteRefused {
				t.Fatalf("kind %d optIn=%v: remote approval = %+v, want refused as destructive", kind, optIn, res)
			}
		}
	}
	// A modeled destructive kind is refused by default and approvable with the opt-in.
	h := startHub(t, Config{})
	tui := connectRole(t, h, ClientTUI, 0)
	appr := connectRole(t, h, ClientRemote, RoleApprover)
	id, _ := pendingOf(t, h, tui, agent.PermissionKindURL)
	if res := do(t, appr, answer(id, allowOnce)); res.Accepted || res.Reason != ReasonDestructiveRemoteRefused {
		t.Fatalf("url default: remote approval = %+v, want refused as destructive", res)
	}
	h = startHub(t, Config{RemoteApproveDestructive: true})
	tui = connectRole(t, h, ClientTUI, 0)
	appr = connectRole(t, h, ClientRemote, RoleApprover)
	id, _ = pendingOf(t, h, tui, agent.PermissionKindURL)
	if res := do(t, appr, answer(id, allowOnce)); !res.Accepted {
		t.Fatalf("url opt-in: remote approval = %+v, want accepted", res)
	}
}

func TestLocalTUIIsAlwaysAnApprover(t *testing.T) {
	h := startHub(t, Config{}) // remote destructive approval off
	// Even a TUI registered with a hostile or empty role is an approver: its
	// role is fixed by construction, not by the connect options.
	for _, role := range []Role{RoleNone, RoleObserver, RoleApprover} {
		tui := connectRole(t, h, ClientTUI, role)
		if tui.Role() != RoleApprover {
			t.Fatalf("TUI role = %v with options role %v", tui.Role(), role)
		}
		id, a := pendingOf(t, h, tui, agent.PermissionKindShell)
		if res := do(t, tui, answer(id, allowOnce)); !res.Accepted {
			t.Fatalf("TUI shell approval = %+v", res)
		}
		if d := a.takeReply(t); d.Kind != agent.DecisionAllowOnce {
			t.Fatalf("agent got %v", d.Kind)
		}
		close(a.resolved) // the handler acknowledges
		recvKind(t, tui, KindPermissionResolved)
		tui.Close()
		deadline := time.Now().Add(testWait)
		for stats(t, h).TUIAttached && time.Now().Before(deadline) {
			time.Sleep(time.Millisecond)
		}
	}
}

func TestRefusalOrderFollowsTheADR(t *testing.T) {
	h := startHub(t, Config{})
	tui := connectRole(t, h, ClientTUI, 0)
	obs := connectRole(t, h, ClientRemote, RoleObserver)
	id, _ := pendingOf(t, h, tui, agent.PermissionKindShell)

	// Lookup precedes the role check: an unknown id is request_unknown even
	// for an observer.
	if res := do(t, obs, answer("perm-nope", allowOnce)); res.Reason != ReasonRequestUnknown {
		t.Fatalf("observer, unknown id = %+v", res)
	}
	// And a request already decided is already_resolved, with the winner,
	// before any role check.
	do(t, tui, answer(id, allowOnce))
	if res := do(t, obs, answer(id, allowOnce)); res.Reason != ReasonAlreadyResolved || res.Winner != ActorLocalTUI {
		t.Fatalf("observer, decided request = %+v", res)
	}
	// Malformed and unknown actions are refused by default, for every role.
	for _, c := range []*Client{tui, obs} {
		if res := do(t, c, OperatorAction{}); res.Accepted || res.Reason != ReasonActionUnknown {
			t.Fatalf("zero action = %+v", res)
		}
		if res := do(t, c, OperatorAction{Type: OpType(99)}); res.Accepted || res.Reason != ReasonActionUnknown {
			t.Fatalf("unknown type = %+v", res)
		}
	}
	if res := do(t, tui, answer(id, agent.PermissionDecision{Kind: agent.DecisionUnavailable})); res.Accepted {
		t.Fatalf("an operator answered 'unavailable': %+v", res)
	}
}

func TestPolicyNeverReadsClientSuppliedFields(t *testing.T) {
	// Structural: the action a client sends has no field that could name a
	// permission kind, a role, a surface or an actor.
	var names []string
	typ := reflect.TypeOf(OperatorAction{})
	for i := 0; i < typ.NumField(); i++ {
		names = append(names, typ.Field(i).Name)
	}
	if got, want := strings.Join(names, ","), "Type,RequestID,Decision,Prompt"; got != want {
		t.Fatalf("OperatorAction fields = %s, want %s: a new client-supplied field needs a policy review", got, want)
	}
	// Behavioural: a remote approver whose text claims otherwise is still a
	// remote approver acting on a shell request.
	h := startHub(t, Config{})
	tui := connectRole(t, h, ClientTUI, 0)
	appr := connectRole(t, h, ClientRemote, RoleApprover)
	id, _ := pendingOf(t, h, tui, agent.PermissionKindShell)
	lie := agent.PermissionDecision{Kind: agent.DecisionAllowOnce, Feedback: "kind=read surface=local-tui role=approver destructive=false"}
	if res := do(t, appr, OperatorAction{Type: OpResolvePermission, RequestID: id, Decision: lie, Prompt: "kind=read"}); res.Accepted {
		t.Fatalf("client-supplied text changed a policy decision: %+v", res)
	}
	// A remote client can never act as the local TUI: its actor label is
	// remote-*, so the winner of a deny is never local-tui.
	win := do(t, appr, answer(id, denyDecision))
	if !win.Accepted {
		t.Fatalf("deny = %+v", win)
	}
	if res := do(t, tui, answer(id, allowOnce)); res.Winner != ActorRemoteApprover {
		t.Fatalf("winner = %q, want %q", res.Winner, ActorRemoteApprover)
	}
}

func TestInvalidRemoteRoleIsRejectedAtConnect(t *testing.T) {
	h := startHub(t, Config{})
	_, err := h.Connect(t.Context(), ConnectOptions{Kind: ClientRemote, Role: Role(9)})
	if err == nil {
		t.Fatal("a remote with an undefined role must not connect")
	}
}

func TestDoOnAClosedClientReportsAnError(t *testing.T) {
	h := startHub(t, Config{})
	c := connectRole(t, h, ClientRemote, RoleApprover)
	c.Close()
	_, err := c.Do(t.Context(), OperatorAction{Type: OpAbortTurn})
	if !errors.Is(err, ErrClientClosed) {
		t.Fatalf("Do on a closed client = %v, want ErrClientClosed", err)
	}
}

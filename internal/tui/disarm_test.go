package tui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

// armedModal returns a model with an approvable pending request that the
// operator has armed with y (the next Enter would approve it).
func armedModal(t *testing.T, ing *fakeIngress, kind agent.PermissionKind, summary string) Model {
	t.Helper()
	m, _, _ := modalModel(t, ing, kind, summary, 70, 16)
	m = scrollModalToEnd(t, m)
	m, _ = step(m, keyRunes("y"))
	if m.armed == "" {
		t.Fatal("y did not arm the approval")
	}
	return m
}

// assertEnterDoesNotApprove presses Enter and fails if it issued a decision.
func assertEnterDoesNotApprove(t *testing.T, m Model, ing *fakeIngress) {
	t.Helper()
	if m.armed != "" {
		t.Fatalf("approval still armed after a non-confirming key: %q", m.armed)
	}
	_, cmd := step(m, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		cmd()
	}
	if calls := ing.snapshot(); len(calls) != 0 {
		t.Fatalf("Enter approved after a non-confirming key: %+v", calls)
	}
}

func TestNonConfirmingKeysDisarmApproval(t *testing.T) {
	reissue := func(context.Context) (CredentialsMsg, error) { return credsMsg(fakeApprover2, fakeObserver2), nil }
	keys := []struct {
		name  string
		creds bool // open the credential panel behind the modal first
		setup func(Model) Model
		key   tea.KeyMsg
	}{
		{name: "esc dismisses credentials", creds: true, key: tea.KeyMsg{Type: tea.KeyEsc}},
		{name: "ctrl+r reissue available", setup: func(m Model) Model { m.reissue = reissue; return m }, key: tea.KeyMsg{Type: tea.KeyCtrlR}},
		{name: "ctrl+r reissue unavailable", setup: func(m Model) Model { m.reissue = nil; return m }, key: tea.KeyMsg{Type: tea.KeyCtrlR}},
		{name: "ctrl+r reissue in flight", setup: func(m Model) Model { m.reissue = reissue; m.reissuing = true; return m }, key: tea.KeyMsg{Type: tea.KeyCtrlR}},
		{name: "up", key: tea.KeyMsg{Type: tea.KeyUp}},
		{name: "down", key: tea.KeyMsg{Type: tea.KeyDown}},
		{name: "pgup", key: tea.KeyMsg{Type: tea.KeyPgUp}},
		{name: "pgdown", key: tea.KeyMsg{Type: tea.KeyPgDown}},
		{name: "scroll with credentials open", creds: true, key: tea.KeyMsg{Type: tea.KeyDown}},
		{name: "other rune", key: keyRunes("x")},
		{name: "esc without credentials", key: tea.KeyMsg{Type: tea.KeyEsc}},
	}
	kinds := []struct {
		name    string
		kind    agent.PermissionKind
		summary string
	}{
		{"read", agent.PermissionKindRead, "read file"},
		{"destructive", agent.PermissionKindShell, "rm -rf build"},
	}
	for _, kd := range kinds {
		for _, tc := range keys {
			t.Run(kd.name+"/"+tc.name, func(t *testing.T) {
				ing := &fakeIngress{}
				m, _, _ := modalModel(t, ing, kd.kind, kd.summary, 70, 16)
				if tc.creds {
					m, _ = step(m, credsMsg(fakeApprover, fakeObserver))
				}
				m = scrollModalToEnd(t, m)
				m, _ = step(m, keyRunes("y"))
				if m.armed == "" {
					t.Fatal("y did not arm the approval")
				}
				if tc.setup != nil {
					m = tc.setup(m)
				}
				m, cmd := step(m, tc.key)
				if cmd != nil && tc.key.Type != tea.KeyCtrlR {
					cmd()
				}
				assertEnterDoesNotApprove(t, m, ing)
			})
		}
	}
}

// TestConfirmingEnterStillApprovesWhenArmed guards the happy path.
func TestConfirmingEnterStillApprovesWhenArmed(t *testing.T) {
	ing := &fakeIngress{}
	m := armedModal(t, ing, agent.PermissionKindShell, "rm -rf build")
	_, cmd := step(m, tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("y then Enter must still approve")
	}
	cmd()
	if calls := ing.snapshot(); len(calls) != 1 || calls[0].Decision.Kind != agent.DecisionAllowOnce {
		t.Fatalf("want one allow, got %+v", calls)
	}
}

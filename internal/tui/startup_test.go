package tui

import (
	"context"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func countingReissuer(calls *int) Reissuer {
	return func(context.Context) (CredentialsMsg, error) {
		*calls++
		return credsMsg(fakeApprover2, fakeObserver2), nil
	}
}

// TestCtrlRBeforeAnyCredentialsIsANoOp: until the initial credentials have
// been delivered there is nothing to revoke, and a re-issue could be
// overwritten by the late startup delivery.
func TestCtrlRBeforeAnyCredentialsIsANoOp(t *testing.T) {
	var calls int
	m := New(Options{Renderer: asciiRenderer(), Reissue: countingReissuer(&calls)})
	m, _ = step(m, tea.WindowSizeMsg{Width: 100, Height: 24})
	m, cmd := step(m, tea.KeyMsg{Type: tea.KeyCtrlR})
	if cmd != nil || calls != 0 {
		t.Fatal("Ctrl+R before the credentials were delivered started a re-issue")
	}
	if m.reissuing {
		t.Fatal("a no-op Ctrl+R marked a re-issue in flight")
	}
	// Once delivered, Ctrl+R works.
	m, _ = step(m, credsMsg(fakeApprover, fakeObserver))
	if _, cmd := step(m, tea.KeyMsg{Type: tea.KeyCtrlR}); cmd == nil {
		t.Fatal("Ctrl+R stayed disabled after the credentials were delivered")
	}
}

// TestStartupDeliveryDuringReissueKeepsGuard: a plain (startup) delivery that
// arrives while a re-issue is pending must not clear the in-flight guard.
func TestStartupDeliveryDuringReissueKeepsGuard(t *testing.T) {
	var calls int
	m := New(Options{Renderer: asciiRenderer(), Reissue: countingReissuer(&calls), Credentials: credsMsg(fakeApprover, fakeObserver)})
	m, _ = step(m, tea.WindowSizeMsg{Width: 100, Height: 24})
	m, first := step(m, tea.KeyMsg{Type: tea.KeyCtrlR})
	if first == nil {
		t.Fatal("Ctrl+R returned no Cmd")
	}
	m, _ = step(m, credsMsg(fakeApprover, fakeObserver)) // late startup delivery
	if _, second := step(m, tea.KeyMsg{Type: tea.KeyCtrlR}); second != nil {
		t.Fatal("a startup delivery cleared the in-flight guard; a second re-issue launched")
	}
}

// TestStartupDeliveryAfterReissueResultCannotRestoreRevokedSet: both orders of
// startup delivery and re-issue result end on the new set.
func TestStartupDeliveryInterleavedWithReissue(t *testing.T) {
	for _, startupFirst := range []bool{true, false} {
		var calls int
		m := New(Options{Renderer: asciiRenderer(), Reissue: countingReissuer(&calls), Credentials: credsMsg(fakeApprover, fakeObserver)})
		m, _ = step(m, tea.WindowSizeMsg{Width: 100, Height: 24})
		m, cmd := step(m, tea.KeyMsg{Type: tea.KeyCtrlR})
		res := cmd()
		if startupFirst {
			m, _ = step(m, credsMsg(fakeApprover, fakeObserver))
			m, _ = step(m, res)
		} else {
			m, _ = step(m, res)
			m, _ = step(m, credsMsg(fakeApprover, fakeObserver))
		}
		v := view(m)
		if !strings.Contains(v, fakeApprover2) || strings.Contains(v, fakeApprover+" ") || strings.Contains(v, fakeApprover+"\n") {
			t.Errorf("startupFirst=%v: panel not on the re-issued set:\n%s", startupFirst, v)
		}
	}
}

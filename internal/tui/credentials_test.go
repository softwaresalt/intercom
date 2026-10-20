package tui

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

// Obviously fake, non-key-shaped placeholders.
const (
	fakeApprover  = "fake-approver-credential-one"
	fakeObserver  = "fake-observer-credential-one"
	fakeApprover2 = "fake-approver-credential-two"
	fakeObserver2 = "fake-observer-credential-two"
	fakeURL       = "ws://127.0.0.1:4242/ws"
)

func credsMsg(approver, observer string) CredentialsMsg {
	exp := time.Date(2026, 10, 20, 16, 0, 0, 0, time.UTC)
	return CredentialsMsg{URL: fakeURL, Credentials: []Credential{
		{Role: "approver", Secret: approver, Fingerprint: "aaaa1111", ExpiresAt: exp},
		{Role: "observer", Secret: observer, Fingerprint: "bbbb2222", ExpiresAt: exp},
	}}
}

func TestCredentialPanelShownOnceThenGone(t *testing.T) {
	ing := &fakeIngress{}
	m := newModel(t, ing, 100, 24)
	m, _ = step(m, credsMsg(fakeApprover, fakeObserver))

	v := view(m)
	for _, want := range []string{fakeApprover, fakeObserver, fakeURL, "aaaa1111", "bbbb2222", "approver", "observer"} {
		if !strings.Contains(v, want) {
			t.Errorf("credential panel lacks %q:\n%s", want, v)
		}
	}

	m, _ = step(m, tea.KeyMsg{Type: tea.KeyEsc})
	v = view(m)
	if strings.Contains(v, fakeApprover) || strings.Contains(v, fakeObserver) {
		t.Fatalf("credentials still on screen after dismissal:\n%s", v)
	}
	if dump := fmt.Sprintf("%+v", m); strings.Contains(dump, fakeApprover) || strings.Contains(dump, fakeObserver) {
		t.Fatal("the model still holds the plaintext after the panel was dismissed")
	}
	if len(ing.snapshot()) != 0 {
		t.Fatal("showing credentials must not touch the hub ingress")
	}
}

func TestCredentialsNeverEnterTheTranscript(t *testing.T) {
	m := newModel(t, &fakeIngress{}, 100, 24)
	m, _ = step(m, credsMsg(fakeApprover, fakeObserver))
	if len(m.entries) != 0 {
		t.Fatalf("a credential message created %d transcript entries", len(m.entries))
	}
	if strings.Contains(m.renderTranscript(), fakeApprover) || strings.Contains(m.vp.View(), fakeApprover) {
		t.Fatal("credential leaked into the transcript viewport")
	}
}

func TestEscDismissesOnlyThePanelNotTheTypedPrompt(t *testing.T) {
	m := newModel(t, &fakeIngress{}, 100, 24)
	m = typeText(m, "draft")
	m, _ = step(m, credsMsg(fakeApprover, fakeObserver))
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyEsc})
	if got := string(m.input); got != "draft" {
		t.Fatalf("Esc on the credential panel cleared the prompt: %q", got)
	}
}

func TestReissueKeyMintsNewAndReplacesPanel(t *testing.T) {
	var calls int
	reissue := func(context.Context) (CredentialsMsg, error) {
		calls++
		return credsMsg(fakeApprover2, fakeObserver2), nil
	}
	ing := &fakeIngress{}
	m := New(Options{Ingress: ing, Renderer: asciiRenderer(), Reissue: reissue})
	m, _ = step(m, tea.WindowSizeMsg{Width: 100, Height: 24})
	m, _ = step(m, credsMsg(fakeApprover, fakeObserver))

	m, cmd := step(m, tea.KeyMsg{Type: tea.KeyCtrlR})
	if cmd == nil {
		t.Fatal("Ctrl+R returned no Cmd")
	}
	if calls != 0 {
		t.Fatal("Update itself minted (side effects belong in the Cmd)")
	}
	res := cmd()
	if calls != 1 {
		t.Fatalf("reissuer called %d times; want 1", calls)
	}
	m, _ = step(m, res)
	v := view(m)
	if !strings.Contains(v, fakeApprover2) || !strings.Contains(v, fakeObserver2) {
		t.Fatalf("panel does not show the re-issued credentials:\n%s", v)
	}
	if strings.Contains(v, fakeApprover) || strings.Contains(v, fakeObserver) {
		t.Fatalf("panel still shows the revoked credentials:\n%s", v)
	}
	if len(ing.snapshot()) != 0 {
		t.Fatal("re-issue must not go through the hub ingress")
	}
}

func TestReissueWorksWhenThePanelWasDismissed(t *testing.T) {
	reissue := func(context.Context) (CredentialsMsg, error) { return credsMsg(fakeApprover2, fakeObserver2), nil }
	m := New(Options{Renderer: asciiRenderer(), Reissue: reissue})
	m, _ = step(m, tea.WindowSizeMsg{Width: 100, Height: 24})
	m, cmd := step(m, tea.KeyMsg{Type: tea.KeyCtrlR})
	if cmd == nil {
		t.Fatal("Ctrl+R returned no Cmd")
	}
	m, _ = step(m, cmd())
	if !strings.Contains(view(m), fakeApprover2) {
		t.Fatal("re-issue from a clean screen did not open the panel")
	}
}

func TestReissueFailureKeepsOldPanelAndShowsNoSecret(t *testing.T) {
	boom := errors.New("entropy unavailable")
	m := New(Options{Renderer: asciiRenderer(), Reissue: func(context.Context) (CredentialsMsg, error) { return CredentialsMsg{}, boom }})
	m, _ = step(m, tea.WindowSizeMsg{Width: 100, Height: 24})
	m, _ = step(m, credsMsg(fakeApprover, fakeObserver))
	m, cmd := step(m, tea.KeyMsg{Type: tea.KeyCtrlR})
	if cmd == nil {
		t.Fatal("Ctrl+R returned no Cmd")
	}
	m, _ = step(m, cmd())
	v := view(m)
	if !strings.Contains(v, "re-issue failed") || !strings.Contains(v, "entropy unavailable") {
		t.Errorf("failure not reported:\n%s", v)
	}
	if !strings.Contains(v, fakeApprover) {
		t.Errorf("the still-valid credentials vanished after a failed re-issue:\n%s", v)
	}
}

func TestReissueWithoutReissuerSaysSo(t *testing.T) {
	m := newModel(t, &fakeIngress{}, 100, 24)
	m, cmd := step(m, tea.KeyMsg{Type: tea.KeyCtrlR})
	if cmd != nil {
		t.Fatal("Ctrl+R without a reissuer returned a Cmd")
	}
	if !strings.Contains(view(m), "re-issue unavailable") {
		t.Errorf("no notice:\n%s", view(m))
	}
}

func TestPermissionModalTakesPriorityOverCredentialPanel(t *testing.T) {
	m, _, _ := modalModel(t, &fakeIngress{}, agent.PermissionKindRead, "read file", 100, 24)
	m, _ = step(m, credsMsg(fakeApprover, fakeObserver))
	v := view(m)
	if !strings.Contains(v, "Permission request") {
		t.Fatalf("an open credential panel hid the permission request:\n%s", v)
	}
	if strings.Contains(v, fakeApprover) {
		t.Fatalf("credentials drawn behind a permission request:\n%s", v)
	}
}

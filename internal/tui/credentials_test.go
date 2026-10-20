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
	m, _ = step(m, credsMsg(fakeApprover, fakeObserver))
	m, _ = step(m, tea.KeyMsg{Type: tea.KeyEsc})
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

// TestEscDismissesCredentialsBeforeAnsweringAPendingPermission: with the panel
// open behind a permission modal, Esc drops the plaintext first and never
// answers the permission.
func TestEscDismissesCredentialsBeforeAnsweringAPendingPermission(t *testing.T) {
	ing := &fakeIngress{}
	m, _, _ := modalModel(t, ing, agent.PermissionKindRead, "read file", 100, 24)
	m, _ = step(m, credsMsg(fakeApprover, fakeObserver))
	if !m.credOpen {
		t.Fatal("credential panel did not open")
	}
	m, cmd := step(m, tea.KeyMsg{Type: tea.KeyEsc})
	if m.credOpen || m.creds.Credentials != nil {
		t.Fatal("Esc left the credential plaintext in the model while a permission was pending")
	}
	if cmd != nil || len(ing.snapshot()) != 0 {
		t.Fatal("Esc answered the permission")
	}
	if len(m.pending) != 1 {
		t.Fatalf("pending permissions = %d; want 1 (Esc must not resolve it)", len(m.pending))
	}
}

// TestPartialReissueFailureShowsTheLiveSet: a re-issue that replaced some
// roles before failing reports the authoritative live set, so the panel never
// keeps showing a credential that was revoked.
func TestPartialReissueFailureShowsTheLiveSet(t *testing.T) {
	boom := errors.New("entropy unavailable")
	live := credsMsg(fakeApprover2, fakeObserver) // approver replaced, observer kept
	m := New(Options{Renderer: asciiRenderer(), Reissue: func(context.Context) (CredentialsMsg, error) { return live, boom }})
	m, _ = step(m, tea.WindowSizeMsg{Width: 100, Height: 24})
	m, _ = step(m, credsMsg(fakeApprover, fakeObserver))
	m, cmd := step(m, tea.KeyMsg{Type: tea.KeyCtrlR})
	if cmd == nil {
		t.Fatal("Ctrl+R returned no Cmd")
	}
	m, _ = step(m, cmd())
	v := view(m)
	if strings.Contains(v, fakeApprover) {
		t.Errorf("the revoked approver credential is still shown:\n%s", v)
	}
	if !strings.Contains(v, fakeApprover2) || !strings.Contains(v, fakeObserver) {
		t.Errorf("the live credentials are not shown:\n%s", v)
	}
	if !strings.Contains(v, "re-issue failed") {
		t.Errorf("failure not reported:\n%s", v)
	}
}

// ---- Copilot round 1: re-issue ordering and panel scrolling ----

// TestRepeatedCtrlRKeepsOneReissueInFlight: a second Ctrl+R while one is
// pending launches nothing, so two sets cannot race each other.
func TestRepeatedCtrlRKeepsOneReissueInFlight(t *testing.T) {
	var calls int
	reissue := func(context.Context) (CredentialsMsg, error) {
		calls++
		return credsMsg(fakeApprover2, fakeObserver2), nil
	}
	m := New(Options{Renderer: asciiRenderer(), Reissue: reissue})
	m, _ = step(m, tea.WindowSizeMsg{Width: 100, Height: 24})
	m, _ = step(m, credsMsg(fakeApprover, fakeObserver))
	m, first := step(m, tea.KeyMsg{Type: tea.KeyCtrlR})
	if first == nil {
		t.Fatal("first Ctrl+R returned no Cmd")
	}
	m, second := step(m, tea.KeyMsg{Type: tea.KeyCtrlR})
	if second != nil {
		t.Fatal("second Ctrl+R while one is in flight launched another re-issue")
	}
	m, _ = step(m, first())
	if calls != 1 {
		t.Fatalf("reissuer called %d times; want 1", calls)
	}
	// Once the result is processed, Ctrl+R works again.
	if _, again := step(m, tea.KeyMsg{Type: tea.KeyCtrlR}); again == nil {
		t.Fatal("Ctrl+R stayed blocked after the re-issue completed")
	}
}

// TestCtrlRUnblocksAfterFailedReissue: a failure also ends the in-flight state.
func TestCtrlRUnblocksAfterFailedReissue(t *testing.T) {
	boom := errors.New("entropy unavailable")
	m := New(Options{Renderer: asciiRenderer(), Reissue: func(context.Context) (CredentialsMsg, error) { return CredentialsMsg{}, boom }})
	m, _ = step(m, tea.WindowSizeMsg{Width: 100, Height: 24})
	m, _ = step(m, credsMsg(fakeApprover, fakeObserver))
	m, cmd := step(m, tea.KeyMsg{Type: tea.KeyCtrlR})
	m, _ = step(m, cmd())
	if _, again := step(m, tea.KeyMsg{Type: tea.KeyCtrlR}); again == nil {
		t.Fatal("Ctrl+R stayed blocked after a failed re-issue")
	}
}

// TestDelayedResultsCannotReplaceNewerSet: with results delivered late, the
// panel ends on the latest set, and a failure's live set still wins.
func TestDelayedResultsCannotReplaceNewerSet(t *testing.T) {
	boom := errors.New("entropy unavailable")
	sets := []CredentialsMsg{credsMsg(fakeApprover2, fakeObserver2)}
	var i int
	m := New(Options{Renderer: asciiRenderer(), Reissue: func(context.Context) (CredentialsMsg, error) {
		s := sets[i%len(sets)]
		i++
		return s, nil
	}})
	m, _ = step(m, tea.WindowSizeMsg{Width: 100, Height: 24})
	m, _ = step(m, credsMsg(fakeApprover, fakeObserver))
	m, c1 := step(m, tea.KeyMsg{Type: tea.KeyCtrlR})
	m, c2 := step(m, tea.KeyMsg{Type: tea.KeyCtrlR})
	if c2 != nil {
		t.Fatal("overlapping re-issue was launched")
	}
	m, _ = step(m, c1())
	if v := view(m); !strings.Contains(v, fakeApprover2) || strings.Contains(v, fakeApprover+" ") {
		t.Fatalf("panel not on the newest set:\n%s", v)
	}
	// A late failure carrying the authoritative live set is still shown.
	m, _ = step(m, ReissueFailedMsg{Err: boom, Live: credsMsg(fakeApprover2, fakeObserver)})
	if v := view(m); !strings.Contains(v, fakeObserver) || strings.Contains(v, fakeObserver2) {
		t.Fatalf("live set after failure not authoritative:\n%s", v)
	}
}

func scrolledAll(t *testing.T, m Model, want ...string) {
	t.Helper()
	seen := map[string]bool{}
	collect := func() {
		flat := strings.NewReplacer("\n", "", " ", "").Replace(view(m))
		for _, w := range want {
			if strings.Contains(flat, w) {
				seen[w] = true
			}
		}
	}
	collect()
	for range 40 {
		m, _ = step(m, tea.KeyMsg{Type: tea.KeyDown})
		collect()
	}
	for _, w := range want {
		if !seen[w] {
			t.Errorf("%q was never reachable by scrolling the panel", w)
		}
	}
}

// TestCredentialPanelScrollsAtSmallSizes: both credentials stay reachable at
// the 80x8 and smallest accepted sizes; nothing is silently clipped.
func TestCredentialPanelScrollsAtSmallSizes(t *testing.T) {
	for name, size := range map[string][2]int{"80x8": {80, 8}, "min": {minWidth, minHeight}} {
		t.Run(name, func(t *testing.T) {
			m := newModel(t, &fakeIngress{}, size[0], size[1])
			m, _ = step(m, credsMsg("fake-a-1", "fake-o-1"))
			v := view(m)
			if got := len(strings.Split(v, "\n")); got != size[1] {
				t.Fatalf("view is %d rows; want %d", got, size[1])
			}
			scrolledAll(t, m, "approver:fake-a-1", "observer:fake-o-1", "Escdismiss")
			// Scrolling back up returns to the top.
			for range 40 {
				m, _ = step(m, tea.KeyMsg{Type: tea.KeyUp})
			}
			if !strings.Contains(view(m), "Credentials") {
				t.Fatalf("panel did not scroll back to the top:\n%s", view(m))
			}
		})
	}
}

func TestCredentialPanelShowsScrollHintWhenClipped(t *testing.T) {
	m := newModel(t, &fakeIngress{}, 80, 8)
	m, _ = step(m, credsMsg(fakeApprover, fakeObserver))
	if v := view(m); !strings.Contains(v, "scroll") {
		t.Fatalf("clipped panel gives no scroll hint:\n%s", v)
	}
}

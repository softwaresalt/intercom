package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Credential is one minted credential, for display to the LOCAL operator only.
type Credential struct {
	// Role names the role the credential grants ("approver", "observer").
	Role string
	// Secret is the plaintext bearer credential. It is held only while the
	// credential panel is open and is never logged, put in the transcript or
	// sent to the hub.
	Secret string
	// Fingerprint identifies the credential without revealing it.
	Fingerprint string
	// ExpiresAt is the credential's fixed expiry.
	ExpiresAt time.Time
}

// CredentialsMsg opens the credential panel. It reaches the model only by
// tea.Program.Send (startup, or a re-issue Cmd's result), never as a hub
// envelope, so it cannot enter the transcript, a hydrate payload or the
// feed's queue.
type CredentialsMsg struct {
	// URL is the WebSocket endpoint (it carries no credential).
	URL         string
	Credentials []Credential
}

// ReissueFailedMsg reports a failed re-issue. Live is the authoritative set of
// credentials that are live after the failure: a re-issue that replaced some
// roles before failing revoked their old credentials, so the panel must show
// Live rather than keep the revoked ones. A zero Live means nothing changed.
type ReissueFailedMsg struct {
	Err  error
	Live CredentialsMsg
}

// Reissuer mints a fresh credential set and revokes the previous one. It runs
// inside a tea.Cmd, off the Update goroutine. On error it may still return the
// live set (a partial re-issue); a zero message means nothing changed.
type Reissuer func(ctx context.Context) (CredentialsMsg, error)

// showCredentials opens the panel with msg, replacing any set already shown
// (a re-issue revokes the old one, so it must not stay on screen).
func (m Model) showCredentials(msg CredentialsMsg) Model {
	m.creds = msg
	m.credOpen = len(msg.Credentials) > 0
	if m.credOpen {
		m.notice = "credentials are shown once: copy them now (Esc dismisses)"
	}
	return m
}

// dismissCredentials drops the plaintext from the model.
func (m Model) dismissCredentials() Model {
	m.creds = CredentialsMsg{}
	m.credOpen = false
	m.notice = ""
	return m
}

// reissueCmd re-mints the credentials in a Cmd. The result arrives as a
// CredentialsMsg (or ReissueFailedMsg), never through the hub.
func (m Model) reissueCmd() (tea.Model, tea.Cmd) {
	if m.reissue == nil {
		m.notice = "re-issue unavailable"
		return m, nil
	}
	m.notice = "re-issuing credentials (the previous ones are revoked)..."
	reissue, timeout := m.reissue, m.timeout
	return m, func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		msg, err := reissue(ctx)
		if err != nil {
			return ReissueFailedMsg{Err: err, Live: msg}
		}
		return msg
	}
}

// credentialView renders the panel in exactly height rows.
func (m Model) credentialView(height int) string {
	rows := []string{
		m.st.bold.Render("Credentials (shown once - copy them now)"),
		"Endpoint: " + SanitizeLine(m.creds.URL),
	}
	for _, c := range m.creds.Credentials {
		rows = append(rows,
			"",
			SanitizeLine(c.Role)+": "+SanitizeLine(c.Secret),
			fmt.Sprintf("  fingerprint %s, expires %s", SanitizeLine(c.Fingerprint), c.ExpiresAt.UTC().Format(time.RFC3339)),
		)
	}
	rows = append(rows, "", "Esc dismiss   Ctrl+R re-issue (revokes these)")
	for i, r := range rows {
		rows[i] = m.wrap(r, m.width)
	}
	return padRows(strings.Join(rows, "\n"), height)
}

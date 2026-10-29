package tui

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/softwaresalt/intercom-go/internal/agent"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

// permission is one pending permission request as shown to the operator. Its
// text has been sanitized but never shortened: the hub denies any request over
// its summary bound, so what arrives here is the whole request.
type permission struct {
	RequestID  string
	ToolCallID string
	Kind       string
	Summary    string
	// Empty is set when the request text is blank: nothing was shown to
	// approve, so the request is deny-only.
	Empty bool
	// Deadline is the envelope ts plus the hub-reported remaining time (design
	// §5.3: anchored to the envelope, not to local receipt). Zero if unknown.
	Deadline time.Time
}

// emptyRequestText stands in for a blank summary.
const emptyRequestText = "(empty request - deny only)"

func unmarshal(raw []byte, v any) error { return json.Unmarshal(raw, v) }

func newPermission(p hub.PermissionRequestedPayload, ts time.Time) permission {
	out := permission{
		RequestID:  p.RequestID, // opaque hub id: echoed back verbatim, never displayed raw
		ToolCallID: SanitizeLine(p.ToolCallID),
		Kind:       SanitizeLine(p.PermissionKind),
		Summary:    Sanitize(p.Summary),
	}
	switch {
	case strings.TrimSpace(p.Summary) == "":
		out.Summary, out.Empty = emptyRequestText, true
	case isBlankSummary(p.Summary):
		// Invisible-only text: keep the visible markers and flag it deny-only.
		out.Summary, out.Empty = out.Summary+"\n"+emptyRequestText, true
	}
	if p.RemainingMS > 0 && !ts.IsZero() {
		out.Deadline = ts.Add(time.Duration(p.RemainingMS) * time.Millisecond)
	}
	return out
}

// appendPermission adds p unless that request id is already pending.
func appendPermission(list []permission, p hub.PermissionRequestedPayload, ts time.Time) []permission {
	for _, e := range list {
		if e.RequestID == p.RequestID {
			return list
		}
	}
	return append(append([]permission(nil), list...), newPermission(p, ts))
}

func (m Model) applyPermissionRequested(raw []byte, ts time.Time) Model {
	var p hub.PermissionRequestedPayload
	if err := unmarshal(raw, &p); err != nil || p.RequestID == "" {
		// Never present a request we could not read: it must not be approvable.
		m.notice = "received an undecodable permission request; it cannot be answered here"
		return m
	}
	m.pending = appendPermission(m.pending, p, ts)
	return m
}

func (m Model) applyPermissionResolved(raw []byte) Model {
	var p hub.PermissionResolvedPayload
	if unmarshal(raw, &p) != nil {
		m.notice = "ignored an undecodable permission.resolved"
		return m
	}
	m = m.dropPending(p.RequestID)
	by := SanitizeLine(p.Actor)
	if p.Reason != "" {
		by += ", " + SanitizeLine(p.Reason)
	}
	m.notice = fmt.Sprintf("permission %s by %s", SanitizeLine(p.Decision), by)
	return m
}

func (m Model) applyPermissionAborted(raw []byte) Model {
	var p hub.PermissionAbortedPayload
	if unmarshal(raw, &p) != nil {
		m.notice = "ignored an undecodable permission.aborted"
		return m
	}
	m = m.dropPending(p.RequestID)
	m.notice = "permission request aborted: " + SanitizeLine(p.Reason)
	return m
}

// dropPending removes a request (answered elsewhere, aborted, or unknown) and
// everything tied to it, so the modal clears.
func (m Model) dropPending(id string) Model {
	kept := make([]permission, 0, len(m.pending))
	for _, p := range m.pending {
		if p.RequestID != id {
			kept = append(kept, p)
		}
	}
	m.pending = kept
	m.inflight = without(m.inflight, id)
	if m.armed == id {
		m.armed = ""
	}
	return m
}

func without(set map[string]bool, id string) map[string]bool {
	if !set[id] {
		return set
	}
	out := make(map[string]bool, len(set))
	for k := range set {
		if k != id {
			out[k] = true
		}
	}
	return out
}

func with(set map[string]bool, id string) map[string]bool {
	out := make(map[string]bool, len(set)+1)
	for k := range set {
		out[k] = true
	}
	out[id] = true
	return out
}

// keepOnly drops in-flight markers for requests that are no longer pending.
func keepOnly(set map[string]bool, pending []permission) map[string]bool {
	var out map[string]bool
	for _, p := range pending {
		if set[p.RequestID] {
			out = with(out, p.RequestID)
		}
	}
	return out
}

// ---- modal geometry and viewing ----

// modalBodyRows is the number of body rows available for the request text. A
// box needs 2 border rows, 2 header rows and 2 hint rows around the body.
func (m Model) modalBodyRows() int { return m.areaHeight() - 6 }

func (m Model) modalInnerWidth() int {
	if w := m.width - 4; w > 10 {
		return w
	}
	return 10
}

// body is the scrollable text of the request. The kind and tool call id sit in
// it, not in a fixed header row, so a long one is wrapped and has to be
// scrolled through like the rest before approval unlocks. Every summary line
// (newline-separated or soft-wrapped) is indented two spaces so none can
// start at column 0 and pass for a "kind:" or "tool call:" header.
func (m Model) permissionBody(p permission) string {
	inner := m.modalInnerWidth()
	lines := strings.Split(m.wrap(p.Summary, max(inner-2, 1)), "\n")
	for i, l := range lines {
		lines[i] = "  " + l
	}
	return m.wrap("kind: "+p.Kind, inner) + "\n" + m.wrap("tool call: "+p.ToolCallID, inner) + "\n\n" + strings.Join(lines, "\n")
}

// syncModal keeps the modal viewport in step with the head pending request.
func (m *Model) syncModal(reset bool) {
	if len(m.pending) == 0 {
		m.modalFor, m.armed = "", ""
		m.modalReady = false
		return
	}
	p := m.pending[0]
	if p.RequestID != m.modalFor || reset {
		rows := m.modalBodyRows()
		if rows < 0 {
			rows = 0
		}
		m.modal = viewport.New(m.modalInnerWidth(), rows)
		m.modal.SetContent(m.permissionBody(p))
		m.modalFor, m.modalSeen, m.armed = p.RequestID, 0, ""
		m.modalGen++
		m.modalReady = m.approvalDelay <= 0
		m.timerNeeded = !m.modalReady
	}
	m.noteSeen()
}

// noteSeen records how far into the body the operator has scrolled.
func (m *Model) noteSeen() {
	if end := m.modal.YOffset + m.modal.Height; end > m.modalSeen {
		m.modalSeen = end
	}
}

// approvable reports whether the operator may approve the head request: every
// line of it has been on screen. A request that cannot be shown in full (the
// terminal is too small for any body row) is deny-only. This is the guard that
// keeps a truncated or partly seen request from being approved.
func (m Model) approvable() bool {
	if len(m.pending) == 0 || m.tooSmall() || m.modal.Height < 1 || m.modalFor != m.pending[0].RequestID || m.pending[0].Empty {
		return false
	}
	return m.modalSeen >= m.modal.TotalLineCount() && m.modalReady && !m.inflight[m.pending[0].RequestID]
}

func (m Model) decisionCmd(kind agent.DecisionKind) (Model, tea.Cmd) {
	id := m.pending[0].RequestID
	m.inflight = with(m.inflight, id)
	m.armed = ""
	m.notice = "sending decision..."
	return m, m.act(hub.OperatorAction{
		Type:      hub.OpResolvePermission,
		RequestID: id,
		Decision:  agent.PermissionDecision{Kind: kind},
	})
}

func (m Model) onModalKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	id := m.pending[0].RequestID
	wasArmed := m.armed == id
	m.armed = ""
	switch k.String() {
	case "enter":
		if wasArmed && m.approvable() {
			return m.decisionCmd(agent.DecisionAllowOnce)
		}
	case "y":
		switch {
		case m.inflight[id]:
		case !m.approvable():
			m.notice = "scroll to the end of the request (PgDn) before approving"
			if m.pending[0].Empty {
				m.notice = "an empty request cannot be approved"
			} else if !m.modalReady && m.modalSeen >= m.modal.TotalLineCount() {
				m.notice = "wait a moment before approving"
			}
		default:
			// Approval takes y then Enter, so typing ahead ("yes\n") when the
			// modal appears cannot approve a command by accident.
			m.armed = id
			m.notice = ""
		}
	case "n":
		if !m.inflight[id] {
			return m.decisionCmd(agent.DecisionDeny)
		}
	case "ctrl+x":
		return m.abortCmd()
	// Contiguous scrolling only: no Home/End jump, so "seen" means every line
	// really was on screen.
	case "up":
		m.modal.ScrollUp(1)
	case "down":
		m.modal.ScrollDown(1)
	case "pgup":
		m.modal.PageUp()
	case "pgdown":
		m.modal.PageDown()
	}
	m.noteSeen()
	return m, nil
}

func (m Model) modalView(height int) string {
	p := m.pending[0]
	inner := m.modalInnerWidth()
	title := "Permission request"
	if n := len(m.pending); n > 1 {
		title += fmt.Sprintf(" (1 of %d)", n)
	}
	expires := "expires: unknown"
	if !p.Deadline.IsZero() {
		expires = "expires: " + p.Deadline.UTC().Format("15:04:05Z")
	}

	if m.modal.Height < 1 {
		msg := "Permission request: terminal too small to show it. Deny only (n)."
		return padRows(m.fit(msg, m.width), height)
	}

	var hint string
	switch id := p.RequestID; {
	case m.inflight[id]:
		hint = "decision sent; waiting for the agent to acknowledge..."
	case m.armed == id:
		hint = "APPROVE? press Enter to confirm, any other key cancels"
	case p.Empty:
		hint = "empty request: approve disabled   n deny"
	case m.approvable():
		hint = "y approve (then Enter to confirm)   n deny"
	case m.modalSeen >= m.modal.TotalLineCount() && !m.modalReady:
		hint = "approval unlocks in a moment   n deny"
	default:
		hint = "scroll to the end (PgDn) to enable approve   n deny"
	}
	first := m.modal.YOffset + 1
	last := m.modal.YOffset + m.modal.VisibleLineCount()
	pos := fmt.Sprintf("lines %d-%d of %d", first, last, m.modal.TotalLineCount())

	rows := []string{
		m.st.bold.Render(m.fit(title, inner)),
		m.fit(expires, inner),
		m.modal.View(),
		m.fit(hint, inner),
		m.fit(pos, inner),
	}
	box := m.st.modal.Width(m.width - 2).Render(strings.Join(rows, "\n"))
	return padRows(box, height)
}

// padRows pads or trims s to exactly n rows.
func padRows(s string, n int) string {
	lines := strings.Split(s, "\n")
	for len(lines) < n {
		lines = append(lines, "")
	}
	return strings.Join(lines[:n], "\n")
}

// Package tui is the local Bubble Tea operator surface (ADR MVP-3: the TUI is
// co-equal with the remote surface and is always an approver).
//
// The model is a pure Elm-style tea.Model. Update and View touch no terminal,
// no clock, no shared state and no agent: every side effect is a tea.Cmd, and
// the only Cmd that acts on the session submits a [hub.OperatorAction] through
// the [Ingress] (design §4.1). The model learns about the session solely from
// hub envelopes (hydrate, then live deltas) fed by a [Feed], and treats the
// hub's session.hydrate as authoritative state: applying one replaces the
// transcript and the pending-permission set, so applying it twice, or after a
// gap, can never duplicate content.
//
// Everything sourced from the agent is passed through [Sanitize] before it is
// stored, so no later render path can emit terminal escapes or bidi controls.
package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/softwaresalt/intercom-go/internal/hub"
)

// DefaultActionTimeout bounds one ingress call made by an action Cmd.
const DefaultActionTimeout = 5 * time.Second

// Layout constants: header, status and input take one row each.
const (
	chromeRows = 3
	minWidth   = 24
	minHeight  = 5
)

// Options configures [New].
type Options struct {
	// Ingress receives every operator action. Nil means actions fail with
	// [ErrNoIngress].
	Ingress Ingress
	// Renderer is the lipgloss renderer (colour profile). Nil selects the
	// default renderer; tests pass a colourless one so frames are stable.
	Renderer *lipgloss.Renderer
	// ActionTimeout bounds one ingress call. Zero selects
	// [DefaultActionTimeout].
	ActionTimeout time.Duration
}

type styles struct {
	header, status, modal, dim, bold lipgloss.Style
	r                                *lipgloss.Renderer
}

func newStyles(r *lipgloss.Renderer) styles {
	if r == nil {
		r = lipgloss.DefaultRenderer()
	}
	return styles{
		r:      r,
		header: r.NewStyle().Bold(true),
		status: r.NewStyle().Reverse(true),
		modal:  r.NewStyle().Border(lipgloss.RoundedBorder()).Padding(0, 1),
		dim:    r.NewStyle().Faint(true),
		bold:   r.NewStyle().Bold(true),
	}
}

// Model is the TUI state. It is a value type: Update returns the next state.
type Model struct {
	ingress Ingress
	timeout time.Duration
	st      styles

	cursor hub.Cursor

	// Session projection, rebuilt from hydrate and advanced by live envelopes.
	entries    []entry
	turnActive bool
	evicted    int

	// Connection and degradation state shown on the status line.
	degraded     bool
	shutdown     bool
	resync       bool
	disconnected error
	notice       string

	// Permission modal state.
	pending   []permission
	inflight  map[string]bool // request ids whose decision Cmd was issued
	armed     string          // request id whose approval awaits Enter
	modal     viewport.Model
	modalFor  string
	modalSeen int // furthest body line the operator has scrolled into view

	promptBusy bool
	input      []rune

	vp            viewport.Model
	width, height int
}

// New builds an empty model.
func New(o Options) Model {
	if o.ActionTimeout <= 0 {
		o.ActionTimeout = DefaultActionTimeout
	}
	return Model{ingress: o.Ingress, timeout: o.ActionTimeout, st: newStyles(o.Renderer), vp: viewport.New(0, 0)}
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd { return nil }

// Update implements tea.Model. It performs no I/O: actions are returned as
// Cmds that run on Bubble Tea's own goroutines.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.layout(true)
		return m, nil
	case FrameMsg:
		return m.onFrame(msg.Frame)
	case ResyncMsg:
		m.resync = true
		return m, nil
	case DisconnectedMsg:
		m.disconnected = msg.Err
		if m.disconnected == nil {
			m.disconnected = ErrNotConnected
		}
		return m, nil
	case ActionResultMsg:
		return m.onActionResult(msg), nil
	case tea.KeyMsg:
		return m.onKey(msg)
	}
	return m, nil
}

// act builds the one Cmd that touches the session: it submits act through the
// ingress and reports the hub's verdict back as an ActionResultMsg.
func (m Model) act(act hub.OperatorAction) tea.Cmd {
	ing, timeout := m.ingress, m.timeout
	return func() tea.Msg {
		if ing == nil {
			return ActionResultMsg{Action: act, Err: ErrNoIngress}
		}
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		res, err := ing.Do(ctx, act)
		return ActionResultMsg{Action: act, Result: res, Err: err}
	}
}

func (m Model) onKey(k tea.KeyMsg) (tea.Model, tea.Cmd) {
	if k.String() == "ctrl+c" {
		return m, tea.Quit
	}
	if len(m.pending) > 0 {
		return m.onModalKey(k)
	}
	switch k.Type { //nolint:exhaustive // other keys are ignored
	case tea.KeyEnter:
		return m.submitPrompt()
	case tea.KeyBackspace:
		if n := len(m.input); n > 0 {
			m.input = m.input[:n-1]
		}
	case tea.KeyEsc:
		m.input = nil
	case tea.KeyCtrlX:
		m.notice = "abort requested"
		return m, m.act(hub.OperatorAction{Type: hub.OpAbortTurn})
	case tea.KeyPgUp:
		m.vp.PageUp()
	case tea.KeyPgDown:
		m.vp.PageDown()
	case tea.KeyUp:
		m.vp.ScrollUp(1)
	case tea.KeyDown:
		m.vp.ScrollDown(1)
	case tea.KeyHome:
		m.vp.GotoTop()
	case tea.KeyEnd:
		m.vp.GotoBottom()
	case tea.KeySpace:
		m.input = append(m.input, ' ')
	case tea.KeyRunes:
		for _, r := range k.Runes {
			if r >= 0x20 && r != 0x7f && (r < 0x80 || r >= 0xa0) {
				m.input = append(m.input, []rune(Sanitize(string(r)))...)
			}
		}
	}
	return m, nil
}

func (m Model) submitPrompt() (tea.Model, tea.Cmd) {
	text := strings.TrimSpace(string(m.input))
	if text == "" || m.promptBusy {
		return m, nil
	}
	if m.disconnected != nil {
		m.notice = "not connected: prompt not sent"
		return m, nil
	}
	m.promptBusy = true
	m.notice = "sending prompt..."
	return m, m.act(hub.OperatorAction{Type: hub.OpPrompt, Prompt: text})
}

func (m Model) onActionResult(msg ActionResultMsg) Model {
	switch msg.Action.Type { //nolint:exhaustive // only the TUI's own actions are issued
	case hub.OpPrompt:
		m.promptBusy = false
		switch {
		case msg.Err != nil:
			m.notice = "prompt failed: " + SanitizeLine(msg.Err.Error())
		case msg.Result.Accepted:
			m.input = nil
			m.notice = "prompt sent"
		default:
			m.notice = "prompt refused: " + SanitizeLine(string(msg.Result.Reason))
		}
	case hub.OpAbortTurn:
		switch {
		case msg.Err != nil:
			m.notice = "abort failed: " + SanitizeLine(msg.Err.Error())
		case msg.Result.Accepted:
			m.notice = "abort sent"
		default:
			m.notice = "abort refused: " + SanitizeLine(string(msg.Result.Reason))
		}
	case hub.OpResolvePermission:
		m = m.onResolveResult(msg)
	}
	return m
}

func (m Model) onResolveResult(msg ActionResultMsg) Model {
	id := msg.Action.RequestID
	switch {
	case msg.Err != nil:
		m.inflight = without(m.inflight, id)
		m.notice = "decision failed: " + SanitizeLine(msg.Err.Error())
	case msg.Result.Accepted:
		// Stay on the modal: permission.resolved is authoritative and arrives
		// only after the agent adapter acknowledged the decision (§5.4 item 7).
		m.notice = "decision sent; waiting for the agent to acknowledge"
	case msg.Result.Reason == hub.ReasonAlreadyResolved:
		m = m.dropPending(id)
		m.notice = "request already answered by " + SanitizeLine(msg.Result.Winner)
	case msg.Result.Reason == hub.ReasonRequestUnknown:
		m = m.dropPending(id)
		m.notice = "request no longer pending"
	default:
		m.inflight = without(m.inflight, id)
		m.notice = "decision refused: " + SanitizeLine(string(msg.Result.Reason))
	}
	return m
}

// ---- envelopes ----

func (m Model) onFrame(f hub.Frame) (tea.Model, tea.Cmd) {
	switch m.cursor.Apply(f) {
	case hub.ActionDiscard:
		return m, nil
	case hub.ActionRehydrate:
		// A gap, generation change or missing baseline: nothing may be applied
		// until the next hydrate (hub.Cursor contract). The feed forces that
		// hydrate; the model just says so.
		m.resync = true
		return m, nil
	}
	env, err := decodeEnvelope(f)
	if err != nil {
		m.notice = "ignored an undecodable envelope"
		return m, nil
	}
	ts := parseTS(env.TS)
	atBottom := m.vp.AtBottom()
	switch env.Kind {
	case hub.KindSessionHydrate:
		m = m.applyHydrate(env.Payload, ts)
	case hub.KindAgentMessageDelta:
		m = m.applyMessage(env.Payload, true)
	case hub.KindAgentMessage:
		m = m.applyMessage(env.Payload, false)
	case hub.KindAgentTurnStart:
		m.turnActive = true
	case hub.KindAgentTurnEnd, hub.KindAgentIdle:
		m.turnActive = false
	case hub.KindToolStarted:
		m = m.applyToolStarted(env.Payload)
	case hub.KindToolEnded:
		m = m.applyToolEnded(env.Payload)
	case hub.KindError:
		m = m.applyError(env.Payload)
	case hub.KindPermissionRequested:
		m = m.applyPermissionRequested(env.Payload, ts)
	case hub.KindPermissionResolved:
		m = m.applyPermissionResolved(env.Payload)
	case hub.KindPermissionAborted:
		m = m.applyPermissionAborted(env.Payload)
	case hub.KindServerShutdown:
		m.shutdown = true
	}
	m.refreshTranscript(atBottom)
	m.syncModal(false)
	return m, nil
}

func (m Model) applyHydrate(raw []byte, ts time.Time) Model {
	var body hydrateBody
	if err := unmarshal(raw, &body); err != nil {
		m.resync = true
		m.notice = "could not decode the hydrate snapshot"
		return m
	}
	m.entries = make([]entry, 0, len(body.Entries))
	m.degraded = false
	for _, e := range body.Entries {
		e = cleanEntry(e)
		if e.Kind == kindError && e.ID == hub.ErrCodeDegraded {
			m.degraded = true
		}
		m.entries = append(m.entries, e)
	}
	m.turnActive = body.TurnActive
	m.evicted = body.EvictedCount
	m.resync = false
	m.shutdown = false
	// The hydrate is authoritative for pending permissions: replace, never merge.
	m.pending = nil
	for _, p := range body.PendingPermissions {
		m.pending = appendPermission(m.pending, p, ts)
	}
	m.inflight = keepOnly(m.inflight, m.pending)
	return m
}

func cleanEntry(e entry) entry {
	e.ID = SanitizeLine(e.ID)
	e.Tool = SanitizeLine(e.Tool)
	e.Text = Sanitize(e.Text)
	e.Kind = SanitizeLine(e.Kind)
	e.Status = SanitizeLine(e.Status)
	return e
}

// find returns the index of the newest entry matching kind, id and status, or -1.
func (m Model) find(kind, id, status string) int {
	for i := len(m.entries) - 1; i >= 0; i-- {
		e := m.entries[i]
		if e.Kind == kind && e.ID == id && e.Status == status {
			return i
		}
	}
	return -1
}

func (m Model) applyMessage(raw []byte, isDelta bool) Model {
	var p hub.MessagePayload
	if unmarshal(raw, &p) != nil {
		m.notice = "ignored an undecodable message envelope"
		return m
	}
	id, text := SanitizeLine(p.MessageID), Sanitize(p.Text)
	i := m.find(kindMessage, id, statusStreaming)
	// Entries are copied on write so a previous Model value is never mutated.
	m.entries = append([]entry(nil), m.entries...)
	switch {
	case isDelta && i >= 0:
		m.entries[i].Text += text
		m.entries[i].Truncated = m.entries[i].Truncated || p.Truncated
	case isDelta:
		m.entries = append(m.entries, entry{Kind: kindMessage, ID: id, Text: text, Status: statusStreaming, Truncated: p.Truncated})
	case i >= 0:
		m.entries[i] = entry{Kind: kindMessage, ID: id, Text: text, Status: statusComplete, Truncated: p.Truncated}
	default:
		m.entries = append(m.entries, entry{Kind: kindMessage, ID: id, Text: text, Status: statusComplete, Truncated: p.Truncated})
	}
	return m
}

func (m Model) applyToolStarted(raw []byte) Model {
	var p hub.ToolStartedPayload
	if unmarshal(raw, &p) != nil {
		return m
	}
	m.entries = append(append([]entry(nil), m.entries...),
		entry{Kind: kindTool, ID: SanitizeLine(p.ToolCallID), Tool: SanitizeLine(p.ToolName), Status: statusRunning, Truncated: p.Truncated})
	return m
}

func (m Model) applyToolEnded(raw []byte) Model {
	var p hub.ToolEndedPayload
	if unmarshal(raw, &p) != nil {
		return m
	}
	id, name := SanitizeLine(p.ToolCallID), SanitizeLine(p.ToolName)
	status := statusFailed
	if p.Success {
		status = statusSucceeded
	}
	m.entries = append([]entry(nil), m.entries...)
	if i := m.find(kindTool, id, statusRunning); i >= 0 {
		m.entries[i].Status = status
		return m
	}
	m.entries = append(m.entries, entry{Kind: kindTool, ID: id, Tool: name, Status: status, Truncated: p.Truncated})
	return m
}

func (m Model) applyError(raw []byte) Model {
	var p hub.ErrorPayload
	if unmarshal(raw, &p) != nil {
		return m
	}
	if p.Code == hub.ErrCodeDegraded {
		m.degraded = true
	}
	m.entries = append(append([]entry(nil), m.entries...),
		entry{Kind: kindError, ID: SanitizeLine(p.Code), Text: Sanitize(p.Message)})
	return m
}

// ---- layout and rendering ----

func (m *Model) areaHeight() int {
	h := m.height - chromeRows
	if h < 1 {
		return 1
	}
	return h
}

// layout recomputes sizes; resetModal restarts the permission body at the top
// (a resize reflows it, so the operator must read it again).
func (m *Model) layout(resetModal bool) {
	atBottom := m.vp.AtBottom()
	m.vp.Width, m.vp.Height = m.width, m.areaHeight()
	m.refreshTranscript(atBottom)
	m.syncModal(resetModal)
}

// refreshTranscript re-renders the transcript. The scroll offset is kept; only
// a viewport that was already at the bottom follows new output.
func (m *Model) refreshTranscript(wasAtBottom bool) {
	m.vp.SetContent(m.renderTranscript())
	if wasAtBottom {
		m.vp.GotoBottom()
	}
}

func (m Model) wrap(text string, width int) string {
	if width < 1 {
		return text
	}
	out := m.st.r.NewStyle().Width(width).Render(text)
	lines := strings.Split(out, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimRight(l, " ")
	}
	return strings.Join(lines, "\n")
}

func entryText(e entry) string {
	var s string
	switch e.Kind {
	case kindTool:
		state := map[string]string{statusRunning: "running", statusSucceeded: "ok", statusFailed: "FAILED"}[e.Status]
		if state == "" {
			state = e.Status
		}
		s = fmt.Sprintf("tool> %s [%s]", e.Tool, state)
	case kindError:
		s = fmt.Sprintf("error> %s: %s", e.ID, e.Text)
	default:
		s = "agent> " + e.Text
		if e.Status == statusStreaming {
			s += " [streaming]"
		}
	}
	if e.Truncated {
		s += " [truncated]"
	}
	return s
}

func (m Model) renderTranscript() string {
	if len(m.entries) == 0 {
		return "No activity yet. Type a prompt and press Enter."
	}
	parts := make([]string, 0, len(m.entries)+1)
	if m.evicted > 0 {
		parts = append(parts, fmt.Sprintf("[%d earlier entries evicted]", m.evicted))
	}
	for _, e := range m.entries {
		parts = append(parts, m.wrap(entryText(e), m.vp.Width))
	}
	return strings.Join(parts, "\n")
}

func (m Model) statusLine() string {
	var parts []string
	switch {
	case m.disconnected != nil:
		parts = append(parts, "DISCONNECTED: "+SanitizeLine(m.disconnected.Error())+" (restart to reconnect)")
	case m.shutdown:
		parts = append(parts, "hub shutting down")
	}
	if m.degraded {
		parts = append(parts, "DEGRADED: agent adapter fault, transcript may be incomplete")
	}
	if m.resync {
		parts = append(parts, "resync: re-hydrating from the hub")
	}
	if m.notice != "" {
		parts = append(parts, m.notice)
	}
	if len(m.pending) == 0 && !m.vp.AtBottom() {
		parts = append(parts, "scrolled up (End to follow)")
	}
	if len(parts) == 0 {
		return "connected"
	}
	return strings.Join(parts, " | ")
}

func (m Model) fit(s string, width int) string {
	return m.st.r.NewStyle().MaxWidth(width).Render(s)
}

// View implements tea.Model.
func (m Model) View() string {
	if m.width < minWidth || m.height < minHeight {
		return "terminal too small"
	}
	turn := "idle"
	if m.turnActive {
		turn = "turn active"
	}
	header := m.st.header.Render(m.fit("intercom | "+turn, m.width))

	var area string
	if len(m.pending) > 0 {
		area = m.modalView(m.areaHeight())
	} else {
		area = m.vp.View()
	}

	prompt := "> " + string(m.input) + "_"
	if m.disconnected != nil {
		prompt = "> (disconnected)"
	} else if len(m.pending) > 0 {
		prompt = "> (answer the permission request first)"
	}
	if rs := []rune(prompt); len(rs) > m.width {
		prompt = "> ..." + string(rs[len(rs)-(m.width-5):])
	}
	return strings.Join([]string{header, area, m.st.status.Render(m.fit(m.statusLine(), m.width)), prompt}, "\n")
}

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
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/softwaresalt/intercom-go/internal/hub"
)

// DefaultActionTimeout bounds one ingress call made by an action Cmd.
const DefaultActionTimeout = 5 * time.Second

// DefaultApprovalDelay is how long a freshly shown (or reset) permission modal
// refuses approval, so typed-ahead keys cannot approve a request unseen.
const DefaultApprovalDelay = 150 * time.Millisecond

// ApprovalReadyMsg unlocks approval for a modal once its delay elapsed. It is
// ignored unless both fields still match the live modal.
type ApprovalReadyMsg struct {
	RequestID string
	ModalGen  uint64
}

// ResyncFailedMsg reports that the model-initiated re-hydrate failed.
type ResyncFailedMsg struct{ Err error }

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
	// Reissue re-mints the session credentials (Ctrl+R). Nil disables the
	// key.
	Reissue Reissuer
	// Credentials is the initial credential set, shown from the first frame.
	// Passing it here (rather than sending it after start) means no startup
	// message can race a Ctrl+R re-issue.
	Credentials CredentialsMsg
	// ApprovalDelay is the approval lock-out after a modal opens or resets.
	// Zero selects [DefaultApprovalDelay].
	ApprovalDelay time.Duration
	// NoApprovalDelay (or a negative ApprovalDelay) disables the lock-out;
	// synchronous tests use it.
	NoApprovalDelay bool
	// MaxEntries caps the live transcript; the oldest entries beyond it are
	// evicted. Zero or negative selects hub.DefaultMaxEntries.
	MaxEntries int
	// NoMouse leaves mouse reporting off. It is for programs fed from a pipe
	// rather than a terminal: on Windows, enabling mouse mode restarts
	// Bubble Tea's input reader, which cannot be cancelled on a pipe and
	// would swallow the next keystroke.
	NoMouse bool
	// After is the timer source for the approval delay. Nil selects time.After.
	After func(time.Duration) <-chan time.Time
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

// SecurityNoticeMsg reports the running count of refused WebSocket upgrades
// and a refusal. Reason is a fixed non-secret code; it is sanitized again when
// rendered. A message with a lower Total than one already applied is stale and
// ignored. The app's mailbox keeps the notice with the highest Total.
type SecurityNoticeMsg struct {
	Status int
	Reason string
	Total  uint64
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

	// Rejected WebSocket upgrades reported by the transport (SecurityNoticeMsg).
	rejectedUpgrades uint64
	lastRejectStatus int
	lastRejectReason string
	// Permission modal state.
	pending   []permission
	inflight  map[string]bool // request ids whose decision Cmd was issued
	armed     string          // request id whose approval awaits Enter
	modal     viewport.Model
	modalFor  string
	modalSeen int // furthest body line the operator has scrolled into view

	approvalDelay time.Duration
	after         func(time.Duration) <-chan time.Time
	modalReady    bool
	modalGen      uint64

	resyncRequested bool
	noMouse         bool
	maxEntries      int
	timerNeeded     bool // syncModal opened/reset a modal that needs its delay timer

	promptBusy bool
	input      []rune

	// Credential panel: the plaintext lives here only while it is shown, and
	// is dropped when the operator dismisses it (shown once, local only).
	reissue  Reissuer
	creds    CredentialsMsg
	credOpen bool
	// credScroll is the first visible row of the panel; reissuing is true while
	// a re-issue Cmd is in flight (cleared when its result is processed).
	credScroll int
	reissuing  bool
	// credsKnown is true once the initial credentials were set (Options or the
	// first CredentialsMsg); Ctrl+R is a no-op before that.
	credsKnown bool

	vp            viewport.Model
	width, height int
}

// New builds an empty model.
func New(o Options) Model {
	if o.ActionTimeout <= 0 {
		o.ActionTimeout = DefaultActionTimeout
	}
	m := Model{ingress: o.Ingress, timeout: o.ActionTimeout, st: newStyles(o.Renderer), vp: viewport.New(0, 0), reissue: o.Reissue, after: o.After}
	switch {
	case o.NoApprovalDelay || o.ApprovalDelay < 0:
		m.approvalDelay = 0
	case o.ApprovalDelay == 0:
		m.approvalDelay = DefaultApprovalDelay
	default:
		m.approvalDelay = o.ApprovalDelay
	}
	if m.after == nil {
		m.after = time.After
	}
	m.maxEntries = o.MaxEntries
	m.noMouse = o.NoMouse
	if m.maxEntries <= 0 {
		m.maxEntries = hub.DefaultMaxEntries
	}
	if len(o.Credentials.Credentials) > 0 {
		m = m.showCredentials(o.Credentials)
	}
	return m
}

// Init implements tea.Model.
func (m Model) Init() tea.Cmd {
	if m.noMouse {
		return nil
	}
	return tea.EnableMouseCellMotion
}

// Update implements tea.Model. It performs no I/O: actions are returned as
// Cmds that run on Bubble Tea's own goroutines.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	next, cmd := m.update(msg)
	nm, ok := next.(Model)
	if !ok || !nm.timerNeeded {
		return next, cmd
	}
	nm.timerNeeded = false
	timer := nm.approvalTimerCmd(nm.modalFor, nm.modalGen)
	if cmd == nil {
		return nm, timer
	}
	return nm, tea.Batch(cmd, timer)
}

// approvalTimerCmd waits out the approval delay, then reports the modal it was
// started for.
func (m Model) approvalTimerCmd(id string, gen uint64) tea.Cmd {
	after, d := m.after, m.approvalDelay
	return func() tea.Msg {
		<-after(d)
		return ApprovalReadyMsg{RequestID: id, ModalGen: gen}
	}
}

// wheelStep is how many rows one wheel notch scrolls.
const wheelStep = 3

// onMouse handles wheel scrolling. Any mouse event disarms a pending approval;
// the modal only ever scrolls contiguously so "seen" stays honest.
func (m Model) onMouse(msg tea.MouseMsg) Model {
	m.armed = ""
	dir := 0
	switch msg.Button { //nolint:exhaustive // only the wheel scrolls
	case tea.MouseButtonWheelUp:
		dir = -1
	case tea.MouseButtonWheelDown:
		dir = 1
	default:
		return m
	}
	switch {
	case len(m.pending) > 0:
		n := min(wheelStep, max(m.modal.Height, 1))
		if dir < 0 {
			m.modal.ScrollUp(n)
		} else {
			m.modal.ScrollDown(n)
		}
		m.noteSeen()
	case m.credOpen:
		m = m.scrollCredentials(dir * wheelStep)
	case dir < 0:
		m.vp.ScrollUp(wheelStep)
	default:
		m.vp.ScrollDown(wheelStep)
	}
	return m
}

// requestRehydrate asks the connection to re-serve a hydrate once per stuck
// resync, bounded by the action timeout.
func (m Model) requestRehydrate() (Model, tea.Cmd) {
	r, ok := m.ingress.(Rehydrater)
	if !ok || m.resyncRequested {
		return m, nil
	}
	m.resyncRequested = true
	timeout := m.timeout
	return m, func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		if err := r.Rehydrate(ctx); err != nil {
			return ResyncFailedMsg{Err: err}
		}
		return nil
	}
}

func (m Model) update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.layout(true)
		return m, nil
	case FrameMsg:
		return m.onFrame(msg.Frame)
	case tea.MouseMsg:
		return m.onMouse(msg), nil
	case ApprovalReadyMsg:
		if len(m.pending) > 0 && m.pending[0].RequestID == msg.RequestID && m.modalFor == msg.RequestID && m.modalGen == msg.ModalGen {
			m.modalReady = true
		}
		return m, nil
	case ResyncFailedMsg:
		m.resyncRequested = false
		if msg.Err != nil {
			m.notice = "re-hydrate failed: " + SanitizeLine(msg.Err.Error())
		}
		return m, nil
	case ResyncMsg:
		m.resync = true
		m.armed = ""
		return m, nil
	case DisconnectedMsg:
		m.disconnected = msg.Err
		if m.disconnected == nil {
			m.disconnected = ErrNotConnected
		}
		return m, nil
	case ActionResultMsg:
		return m.onActionResult(msg), nil
	case CredentialsMsg:
		// An initial delivery only: once credentials are known a plain message
		// is stale (a re-issue revoked them) and must not replace the live set
		// nor clear an in-flight re-issue.
		if m.credsKnown {
			return m, nil
		}
		return m.showCredentials(msg), nil
	case reissuedMsg:
		m.reissuing = false
		return m.showCredentials(msg.CredentialsMsg), nil
	case ReissueFailedMsg:
		m.reissuing = false
		if len(msg.Live.Credentials) > 0 {
			// A partial failure replaced some roles: show what is really live.
			m = m.showCredentials(msg.Live)
		}
		m.notice = "re-issue failed: " + SanitizeLine(msg.Err.Error())
		return m, nil
	case SecurityNoticeMsg:
		// Notices are coalesced and may arrive out of order: never go back.
		if msg.Total >= m.rejectedUpgrades {
			m.rejectedUpgrades = msg.Total
			m.lastRejectStatus = msg.Status
			m.lastRejectReason = msg.Reason
		}
		return m, nil
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
	// Single choke point: an armed approval survives only until the very next
	// key, and only Enter may confirm it. Every other key (shortcuts, scrolling,
	// credential dismissal) disarms before any handler can return early.
	if k.Type != tea.KeyEnter {
		m.armed = ""
	}
	if k.String() == "ctrl+c" {
		return m, tea.Quit
	}
	if k.Type == tea.KeyCtrlR {
		return m.reissueCmd()
	}
	if k.Type == tea.KeyEsc && m.credOpen {
		// Dismissing the credential panel has priority, even behind a permission
		// modal: the plaintext must not linger, and Esc never answers a request.
		return m.dismissCredentials(), nil
	}
	if len(m.pending) > 0 {
		return m.onModalKey(k)
	}
	if m.credOpen {
		switch k.Type { //nolint:exhaustive // other keys fall through
		case tea.KeyUp:
			return m.scrollCredentials(-1), nil
		case tea.KeyDown:
			return m.scrollCredentials(1), nil
		case tea.KeyPgUp:
			return m.scrollCredentials(-m.areaHeight()), nil
		case tea.KeyPgDown:
			return m.scrollCredentials(m.areaHeight()), nil
		}
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
		return m.abortCmd()
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
		// What the operator typed is sent as typed (ZWJ emoji and other format
		// runes included); only control runes are dropped. Display safety is
		// applied when the input line is rendered, not here.
		for _, r := range k.Runes {
			if r >= 0x20 && r != 0x7f && (r < 0x80 || r >= 0xa0) && r != '\u2028' && r != '\u2029' && r != utf8.RuneError {
				m.input = append(m.input, r)
			}
		}
	}
	return m, nil
}

// abortCmd requests an abort of the running turn through the ingress.
func (m Model) abortCmd() (tea.Model, tea.Cmd) {
	m.notice = "abort requested"
	return m, m.act(hub.OperatorAction{Type: hub.OpAbortTurn})
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
	// A dropped head must advance the modal content and the decision target
	// together, not wait for the next frame or resize.
	m.syncModal(false)
	return m
}

// ---- envelopes ----

func (m Model) onFrame(f hub.Frame) (tea.Model, tea.Cmd) {
	action := m.cursor.Apply(f)
	if f.Kind() == hub.KindServerShutdown {
		// Terminal frame: even across a gap (deliverTerminal may drop the
		// oldest queued frame) the hub is going away, so never resync.
		m.shutdown = true
		m.resync = false
		if action != hub.ActionApply {
			return m, nil
		}
	}
	switch action {
	case hub.ActionDiscard:
		return m, nil
	case hub.ActionRehydrate:
		// A gap, generation change or missing baseline: nothing may be applied
		// until the next hydrate (hub.Cursor contract). The feed forces that
		// hydrate; the model asks for one too, in case the feed cannot.
		m.resync = true
		m.armed = ""
		return m.requestRehydrate()
	}
	env, err := decodeEnvelope(f)
	if err != nil {
		m.notice = "ignored an undecodable envelope"
		return m, nil
	}
	ts := parseTS(env.TS)
	atBottom := m.vp.AtBottom()
	var cmd tea.Cmd
	switch env.Kind {
	case hub.KindSessionHydrate:
		m = m.applyHydrate(env.Payload, ts)
		if m.resync {
			m, cmd = m.requestRehydrate()
		}
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
	return m, cmd
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
	m.capEntries()
	m.resync = false
	m.resyncRequested = false
	m.armed = ""
	// Force syncModal to rebuild the modal (fresh viewport, read progress and
	// approval delay) even when the head request ID is unchanged.
	m.modalFor = ""
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
		m.entries[i].wrapped, m.entries[i].wrapWidth = "", 0
	case isDelta:
		m.entries = append(m.entries, entry{Kind: kindMessage, ID: id, Text: text, Status: statusStreaming, Truncated: p.Truncated})
	case i >= 0:
		m.entries[i] = entry{Kind: kindMessage, ID: id, Text: text, Status: statusComplete, Truncated: p.Truncated}
	default:
		m.entries = append(m.entries, entry{Kind: kindMessage, ID: id, Text: text, Status: statusComplete, Truncated: p.Truncated})
	}
	m.capEntries()
	return m
}

// capEntries evicts the oldest entries beyond maxEntries.
func (m *Model) capEntries() {
	if m.maxEntries <= 0 || len(m.entries) <= m.maxEntries {
		return
	}
	n := len(m.entries) - m.maxEntries
	m.entries = append([]entry(nil), m.entries[n:]...)
	m.evicted += n
}

func (m Model) applyToolStarted(raw []byte) Model {
	var p hub.ToolStartedPayload
	if unmarshal(raw, &p) != nil {
		return m
	}
	m.entries = append(append([]entry(nil), m.entries...),
		entry{Kind: kindTool, ID: SanitizeLine(p.ToolCallID), Tool: SanitizeLine(p.ToolName), Status: statusRunning, Truncated: p.Truncated})
	m.capEntries()
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
		if m.entries[i].Tool == "" && name != "" {
			m.entries[i].Tool = name
		}
		m.entries[i].Truncated = m.entries[i].Truncated || p.Truncated
		m.entries[i].wrapped, m.entries[i].wrapWidth = "", 0
		return m
	}
	m.entries = append(m.entries, entry{Kind: kindTool, ID: id, Tool: name, Status: status, Truncated: p.Truncated})
	m.capEntries()
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
	m.capEntries()
	return m
}

// ---- layout and rendering ----

// tooSmall reports whether the terminal is below the minimum View renders in.
// View and the approval guard both use it so they cannot diverge.
func (m Model) tooSmall() bool {
	return m.width < minWidth || m.height < minHeight
}

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

func (m *Model) renderTranscript() string {
	if len(m.entries) == 0 {
		return "No activity yet. Type a prompt and press Enter."
	}
	parts := make([]string, 0, len(m.entries)+1)
	if m.evicted > 0 {
		parts = append(parts, fmt.Sprintf("[%d earlier entries evicted]", m.evicted))
	}
	w := m.vp.Width
	for i := range m.entries {
		e := &m.entries[i]
		if e.wrapped == "" || e.wrapWidth != w {
			e.wrapped, e.wrapWidth = m.wrap(entryText(*e), w), w
		}
		parts = append(parts, e.wrapped)
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
	if m.rejectedUpgrades > 0 {
		parts = append(parts, fmt.Sprintf("rejected upgrades: %d (last: %d %s)",
			m.rejectedUpgrades, m.lastRejectStatus, SanitizeLine(m.lastRejectReason)))
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
	if m.tooSmall() {
		return "terminal too small"
	}
	turn := "idle"
	if m.turnActive {
		turn = "turn active"
	}
	header := m.st.header.Render(m.fit("intercom | "+turn, m.width))

	var area string
	switch {
	case len(m.pending) > 0:
		area = m.modalView(m.areaHeight())
	case m.credOpen:
		area = m.credentialView(m.areaHeight())
	default:
		area = m.vp.View()
	}

	prompt := "> " + SanitizeLine(string(m.input)) + "_"
	if m.disconnected != nil {
		prompt = "> (disconnected)"
	} else if len(m.pending) > 0 {
		prompt = "> (answer the permission request first)"
	}
	if pw := lipgloss.Width(prompt); pw > m.width {
		// Linear, grapheme-aware: drop whole cells from the left in one pass.
		keep := max(m.width-lipgloss.Width("> ..."), 0)
		prompt = "> ..." + ansi.TruncateLeft(prompt, pw-keep, "")
	}
	return strings.Join([]string{header, area, m.st.status.Render(m.fit(m.statusLine(), m.width)), prompt}, "\n")
}

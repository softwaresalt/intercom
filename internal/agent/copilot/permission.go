package copilot

import (
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	copilot "github.com/github/copilot-sdk/go"
	"github.com/github/copilot-sdk/go/rpc"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

// This file is the ONLY place intercom's PermissionDecision is converted to
// and from rpc.PermissionDecision (design §7.1). rpc.PermissionDecision is
// documented "Experimental: ... may change or be removed", so its blast
// radius is confined here.

// toSDKDecision converts an intercom decision to the SDK type. Fail-closed:
// any invalid decision (including the zero value) becomes a rejection, never
// an approval.
func toSDKDecision(d agent.PermissionDecision) rpc.PermissionDecision {
	switch d.Kind {
	case agent.DecisionAllowOnce:
		return &rpc.PermissionDecisionApproveOnce{}
	case agent.DecisionDeny:
		return rejection(d.Feedback)
	case agent.DecisionUnavailable:
		return &rpc.PermissionDecisionUserNotAvailable{}
	default:
		return rejection("invalid permission decision")
	}
}

func rejection(feedback string) rpc.PermissionDecision {
	r := &rpc.PermissionDecisionReject{}
	if feedback != "" {
		r.Feedback = &feedback
	}
	return r
}

// fromSDKDecision converts an SDK decision back to the intercom type. It
// reports ok=false for SDK variants intercom does not model.
func fromSDKDecision(d rpc.PermissionDecision) (agent.PermissionDecision, bool) {
	switch v := d.(type) {
	case *rpc.PermissionDecisionApproveOnce:
		return agent.PermissionDecision{Kind: agent.DecisionAllowOnce}, true
	case *rpc.PermissionDecisionReject:
		out := agent.PermissionDecision{Kind: agent.DecisionDeny}
		if v.Feedback != nil {
			out.Feedback = *v.Feedback
		}
		return out, true
	case *rpc.PermissionDecisionUserNotAvailable:
		return agent.PermissionDecision{Kind: agent.DecisionUnavailable}, true
	default:
		return agent.PermissionDecision{}, false
	}
}

// maxSummaryRunes bounds PermissionRequest.Summary so an adversarial or
// accidental megabyte command line cannot flow unbounded into the hub.
const maxSummaryRunes = 2048

// writePreviewTruncatedSuffix marks a write preview clipped to fit the summary.
const writePreviewTruncatedSuffix = "\n[write preview truncated]"

// maxIdentRunes bounds single-line identifiers (RawKind, ToolCallID).
const maxIdentRunes = 128

// sanitizeIdent returns s as a single-line identifier: control, format, and
// line/paragraph-separator runes (including \n and \t) become U+FFFD and the
// result is capped at max runes with an ellipsis.
func sanitizeIdent(s string, max int) string {
	var b strings.Builder
	n := 0
	for _, r := range s {
		n++
		if n > max {
			break
		}
		if isInvisibleOrBlankRune(r) {
			b.WriteRune(utf8.RuneError)
			continue
		}
		b.WriteRune(r)
	}
	if n > max {
		b.WriteString("…")
	}
	return b.String()
}

// sanitizeBound returns s with control and bidi/format runes replaced by
// U+FFFD (newline and tab are kept), truncated to max runes with an ellipsis.
// It is a single pass that never allocates proportionally to an oversized
// input. fullLen is the rune count of the original string.
func sanitizeBound(s string, max int) (out string, truncated bool, fullLen int) {
	var b strings.Builder
	for _, r := range s {
		fullLen++
		if fullLen > max {
			continue // keep counting, stop copying
		}
		switch {
		case r == '\n' || r == '\t':
			b.WriteRune(r)
		case isInvisibleOrBlankRune(r):
			b.WriteRune(utf8.RuneError)
		default:
			b.WriteRune(r)
		}
	}
	if fullLen > max {
		b.WriteString("…")
		truncated = true
	}
	return b.String(), truncated, fullLen
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}

// isInvisibleOrBlankRune reports whether r is a control, format, private-use,
// variation-selector, or visually blank filler rune that must never reach an
// operator surface as-is.
func isInvisibleOrBlankRune(r rune) bool {
	switch r {
	case '\u2028', '\u2029', '\u2800', '\u3164', '\u115f', '\u1160', '\uffa0', '\u034f', '\u17b4', '\u17b5':
		return true
	}
	return unicode.IsControl(r) || unicode.Is(unicode.Cf, r) || unicode.Is(unicode.Co, r) ||
		unicode.Is(unicode.Variation_Selector, r)
}

// foldURLsLookalikeRune folds r to the ASCII byte of the "urls:" character it
// imitates, or reports false when r imitates none of them.
func foldURLsLookalikeRune(r rune) (byte, bool) {
	if r >= 0xff01 && r <= 0xff5e {
		r -= 0xfee0 // fullwidth ASCII
	}
	switch r {
	case ':', '\ufe55', '\ufe13', '\u2236', '\ua789', '\u02d0', '\u0589':
		return ':', true
	case 'u', 'U', '\u03c5', '\u0443', '\u0423', '\u1d1c', '\u057d':
		return 'u', true
	case 'r', 'R', '\u0433', '\u0280', '\u0440', '\u03c1':
		return 'r', true
	case 'l', 'L', '1', '|', 'I', '\u0131', '\u04cf', '\u04c0', '\u0456', '\u0406', '\u03b9', '\u029f', '\u2113':
		return 'l', true
	case 's', 'S', '5', '\u0455', '\u0405', '\ua731', '\u03c2', '\u03c3':
		return 's', true
	}
	return 0, false
}

func isCombiningMark(r rune) bool { return unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) }

// hasURLsLookalikePrefix reports whether line starts, after blank and
// combining runes, with something that renders as "urls:".
func hasURLsLookalikePrefix(s string) bool {
	var buf [5]byte
	n := 0
	lead := true
	for _, r := range s {
		if lead {
			if unicode.IsSpace(r) || isInvisibleOrBlankRune(r) || r == utf8.RuneError || isCombiningMark(r) {
				continue
			}
			lead = false
		}
		if isCombiningMark(r) {
			continue
		}
		b, ok := foldURLsLookalikeRune(r)
		if !ok {
			return false
		}
		buf[n] = b
		if n++; n == len(buf) {
			break
		}
	}
	return n == len(buf) && string(buf[:]) == "urls:"
}

// sanitizeShellURL returns u as a single URL token: whitespace, commas (the
// list separator), control, format, and line/paragraph-separator runes become
// U+FFFD so a possible-URL cannot forge extra list entries or lines.
func sanitizeShellURL(u string) string {
	var b strings.Builder
	for _, r := range u {
		if unicode.IsSpace(r) || r == ',' || isInvisibleOrBlankRune(r) {
			b.WriteRune(utf8.RuneError)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// indentURLsFraming prefixes two spaces to every line of the command text whose
// leading-whitespace-trimmed form starts with "urls:" (case-insensitive), so no
// command line can begin at column 0 and imitate the adapter's own urls line.
func indentURLsFraming(text string) string {
	lines := strings.Split(text, "\n")
	for i, l := range lines {
		if hasURLsLookalikePrefix(l) {
			lines[i] = "  " + l
		}
	}
	return strings.Join(lines, "\n")
}

// shellSummary renders a shell request: the command text followed by any
// URLs the command may access, so the operator sees network targets. Each URL
// is sanitized and command lines imitating the urls line are indented, so the
// "\nurls: " framing is only ever produced by the adapter.
func shellSummary(r *rpc.PermissionRequestShell) string {
	text := indentURLsFraming(r.FullCommandText)
	urls := make([]string, 0, len(r.PossibleURLs))
	for _, u := range r.PossibleURLs {
		if u.URL != "" {
			urls = append(urls, sanitizeShellURL(u.URL))
		}
	}
	if len(urls) > 0 {
		text += "\nurls: " + strings.Join(urls, ", ")
	}
	return text
}

// writeSummary renders a write request: the file name followed by the diff
// (or the new file contents when there is no diff). A body too large for the
// summary is clipped with an explicit notice and sets previewClipped instead
// of truncated; only a file name that leaves no room for the preview sets
// truncated. fullLen is the unclipped rune length of the file name, separator,
// and body.
func writeSummary(r *rpc.PermissionRequestWrite) (summary string, truncated, previewClipped bool, fullLen int) {
	fnSan, fnTrunc, fnLen := sanitizeBound(r.FileName, maxSummaryRunes)
	if fnTrunc {
		return fnSan, true, false, fnLen
	}
	body := r.Diff
	if body == "" {
		body = deref(r.NewFileContents)
	}
	if body == "" {
		return fnSan, false, false, fnLen
	}
	suffixLen := utf8.RuneCountInString(writePreviewTruncatedSuffix)
	avail := maxSummaryRunes - fnLen - 1
	if avail < suffixLen+1 {
		return fnSan, true, false, fnLen
	}
	bodySan, bodyTrunc, bodyLen := sanitizeBound(body, avail)
	totalLen := fnLen + 1 + bodyLen
	if !bodyTrunc {
		return fnSan + "\n" + bodySan, false, false, totalLen
	}
	preview, _, _ := sanitizeBound(body, avail-suffixLen-1)
	return fnSan + "\n" + preview + writePreviewTruncatedSuffix, false, true, totalLen
}

// toDomainRequest converts an SDK permission request to the intercom type.
func toDomainRequest(req copilot.PermissionRequest, inv copilot.PermissionInvocation) agent.PermissionRequest {
	out := agent.PermissionRequest{SessionID: inv.SessionID}
	if req == nil {
		return out
	}
	out.RawKind = string(req.Kind())
	rawKind := out.RawKind
	var text string
	done := false // write requests build their own summary
	switch r := req.(type) {
	case *rpc.PermissionRequestShell:
		out.Kind, text, out.ToolCallID = agent.PermissionKindShell, shellSummary(r), deref(r.ToolCallID)
	case *rpc.PermissionRequestWrite:
		out.Kind, out.ToolCallID = agent.PermissionKindWrite, deref(r.ToolCallID)
		out.Summary, out.Truncated, out.PreviewClipped, out.FullLen = writeSummary(r)
		done = true
	case *rpc.PermissionRequestRead:
		out.Kind, text, out.ToolCallID = agent.PermissionKindRead, r.Path, deref(r.ToolCallID)
	case *rpc.PermissionRequestURL:
		out.Kind, text, out.ToolCallID = agent.PermissionKindURL, r.URL, deref(r.ToolCallID)
	default:
		// Kinds not yet modelled carry only their discriminator; the live-SDK
		// test records which kinds reach the handler (ADR MVP-1). Hub policy
		// treats Other as the most restrictive class.
		out.Kind, text = agent.PermissionKindOther, rawKind
	}
	if !done {
		out.Summary, out.Truncated, out.FullLen = sanitizeBound(text, maxSummaryRunes)
	}
	// Identifiers are single-line and capped independently of the summary.
	out.RawKind = sanitizeIdent(rawKind, maxIdentRunes)
	out.ToolCallID = sanitizeIdent(out.ToolCallID, maxIdentRunes)
	return out
}

// permissionBridge adapts the SDK's blocking PermissionHandlerFunc to the
// hub's ask/reply channels.
type permissionBridge struct {
	asks    chan<- agent.PermissionAsk
	hubDone <-chan struct{}
	// stop is closed when the adapter shuts down; pending requests then
	// resolve to UserNotAvailable even if the hub is still alive. Nil means
	// "never" (a nil channel blocks forever in select).
	stop <-chan struct{}
	log  *slog.Logger

	// Turn scope: asks are stamped with the turn that issued them, and an
	// abort cancels the turn's parked asks through turnStop. mu guards the
	// fields below.
	mu          sync.Mutex
	turnID      string
	turnGen     uint64
	turnAborted bool
	turnStop    chan struct{}
}

func newPermissionBridge(asks chan<- agent.PermissionAsk, hubDone <-chan struct{}, log *slog.Logger) *permissionBridge {
	if log == nil {
		log = slog.Default()
	}
	return &permissionBridge{asks: asks, hubDone: hubDone, log: log, turnStop: make(chan struct{})}
}

// setTurn records the start of a turn and returns its id (generated when id is
// empty). The first turn start after an abort replaces the cancelled
// generation with a fresh open one.
func (b *permissionBridge) setTurn(id string) string {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.turnGen++
	if id == "" {
		id = fmt.Sprintf("turn-%d", b.turnGen)
	}
	b.turnID = id
	if b.turnAborted || b.turnStop == nil {
		b.turnStop = make(chan struct{})
		b.turnAborted = false
	}
	return b.turnID
}

// abortTurn cancels the current turn's asks and returns the turn's id. The
// cancelled, identifiable generation is retained (turnID kept, turnStop left
// closed) until the next setTurn, so a callback that reaches currentTurn()
// after the abort still sees a closed turnStop and is refused.
func (b *permissionBridge) abortTurn() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.turnID == "" {
		b.turnGen++
		b.turnID = fmt.Sprintf("aborted-turn-%d", b.turnGen)
	}
	if !b.turnAborted {
		b.turnAborted = true
		if b.turnStop == nil {
			b.turnStop = make(chan struct{})
		}
		close(b.turnStop)
	}
	return b.turnID
}

// clearTurn ends a turn that was not aborted; an aborted generation is kept
// until the next turn starts.
func (b *permissionBridge) clearTurn() {
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.turnAborted {
		b.turnID = ""
	}
}

// currentTurn returns the current turn id and the channel closed when it is
// aborted.
func (b *permissionBridge) currentTurn() (string, <-chan struct{}) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.turnID, b.turnStop
}

// handler returns the SDK handler value for SessionConfig.OnPermissionRequest.
func (b *permissionBridge) handler() copilot.PermissionHandlerFunc { return b.handle }

// handle blocks until the hub replies or the hub is gone. It returns only a
// value the hub sent, or UserNotAvailable when the hub-done or stop channel
// closes first (shutdown always wins over a simultaneously ready approval) or
// the hub revoked the ask. Whatever it returns it first records in the ask's
// Outcome, which is how the hub learns what the SDK received. It never
// approves on its own, and it denies without asking when the request cannot be
// shown in full.
func (b *permissionBridge) handle(req copilot.PermissionRequest, inv copilot.PermissionInvocation) (rpc.PermissionDecision, error) {
	unavailable := func() (rpc.PermissionDecision, error) {
		return &rpc.PermissionDecisionUserNotAvailable{}, nil
	}
	turnID, turnStop := b.currentTurn()
	gone := func() bool {
		select {
		case <-b.hubDone:
			return true
		case <-b.stop:
			return true
		case <-turnStop:
			return true
		default:
			return false
		}
	}
	if gone() {
		return unavailable()
	}

	domain := toDomainRequest(req, inv)
	domain.TurnID = turnID
	if domain.Truncated {
		// An operator cannot approve text they cannot see.
		return toSDKDecision(agent.PermissionDecision{
			Kind:     agent.DecisionDeny,
			Feedback: fmt.Sprintf("request is %d characters and cannot be reviewed in full; denied", domain.FullLen),
		}), nil
	}

	reply := make(chan agent.PermissionDecision, 1)
	resolved := make(chan struct{})
	outcome := &agent.PermissionOutcome{}
	defer close(resolved) // runs after the outcome below is recorded
	ask := agent.PermissionAsk{Request: domain, Reply: reply, Outcome: outcome, Resolved: resolved}
	// deliver records the decision about to be returned to the SDK, so the hub
	// knows exactly what the SDK got (see agent.PermissionOutcome). A revoked
	// ask (the hub gave up waiting) delivers nothing but UserNotAvailable.
	deliver := func(d agent.PermissionDecision) (rpc.PermissionDecision, error) {
		if !outcome.Deliver(d) {
			d = agent.PermissionDecision{Kind: agent.DecisionUnavailable}
		}
		return toSDKDecision(d), nil
	}
	notAvailable := agent.PermissionDecision{Kind: agent.DecisionUnavailable}
	// Shutdown or an abort may have begun while the request was converted;
	// re-check so a ready send never wins a random select over it.
	if gone() {
		return unavailable()
	}
	select {
	case b.asks <- ask:
	case <-b.hubDone:
		return unavailable()
	case <-b.stop:
		return unavailable()
	case <-turnStop:
		return unavailable()
	}
	select {
	case d := <-reply:
		if d.Kind == agent.DecisionAllowOnce && gone() {
			// Never deliver an approval after shutdown began or the turn was
			// aborted. Only an allow is downgraded: a deny or a hub-issued
			// UserNotAvailable is already safe and is returned (and recorded)
			// unchanged.
			d = notAvailable
		}
		return deliver(d)
	case <-b.hubDone:
		return deliver(notAvailable)
	case <-b.stop:
		return deliver(notAvailable)
	case <-turnStop:
		return deliver(notAvailable)
	}
}

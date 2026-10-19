package copilot

import (
	"fmt"
	"log/slog"
	"strings"
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
		case unicode.IsControl(r) || unicode.Is(unicode.Cf, r):
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

// toDomainRequest converts an SDK permission request to the intercom type.
func toDomainRequest(req copilot.PermissionRequest, inv copilot.PermissionInvocation) agent.PermissionRequest {
	out := agent.PermissionRequest{SessionID: inv.SessionID}
	if req == nil {
		return out
	}
	out.RawKind = string(req.Kind())
	var text string
	switch r := req.(type) {
	case *rpc.PermissionRequestShell:
		out.Kind, text, out.ToolCallID = agent.PermissionKindShell, r.FullCommandText, deref(r.ToolCallID)
	case *rpc.PermissionRequestWrite:
		out.Kind, text, out.ToolCallID = agent.PermissionKindWrite, r.FileName, deref(r.ToolCallID)
	case *rpc.PermissionRequestRead:
		out.Kind, text, out.ToolCallID = agent.PermissionKindRead, r.Path, deref(r.ToolCallID)
	case *rpc.PermissionRequestURL:
		out.Kind, text, out.ToolCallID = agent.PermissionKindURL, r.URL, deref(r.ToolCallID)
	default:
		// Kinds not yet modelled carry only their discriminator; the live-SDK
		// test records which kinds reach the handler (ADR MVP-1). Hub policy
		// treats Other as the most restrictive class.
		out.Kind, text = agent.PermissionKindOther, out.RawKind
	}
	out.Summary, out.Truncated, out.FullLen = sanitizeBound(text, maxSummaryRunes)
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
}

func newPermissionBridge(asks chan<- agent.PermissionAsk, hubDone <-chan struct{}, log *slog.Logger) *permissionBridge {
	if log == nil {
		log = slog.Default()
	}
	return &permissionBridge{asks: asks, hubDone: hubDone, log: log}
}

// handler returns the SDK handler value for SessionConfig.OnPermissionRequest.
func (b *permissionBridge) handler() copilot.PermissionHandlerFunc { return b.handle }

// handle blocks until the hub replies or the hub is gone. It returns only a
// value the hub sent, or UserNotAvailable when the hub-done or stop channel
// closes first (shutdown always wins over a simultaneously ready approval). It
// never approves on its own, and it denies without asking when the request
// cannot be shown in full.
func (b *permissionBridge) handle(req copilot.PermissionRequest, inv copilot.PermissionInvocation) (rpc.PermissionDecision, error) {
	unavailable := func() (rpc.PermissionDecision, error) {
		return &rpc.PermissionDecisionUserNotAvailable{}, nil
	}
	gone := func() bool {
		select {
		case <-b.hubDone:
			return true
		case <-b.stop:
			return true
		default:
			return false
		}
	}
	if gone() {
		return unavailable()
	}

	domain := toDomainRequest(req, inv)
	if domain.Truncated {
		// An operator cannot approve text they cannot see.
		return toSDKDecision(agent.PermissionDecision{
			Kind:     agent.DecisionDeny,
			Feedback: fmt.Sprintf("request is %d characters and cannot be reviewed in full; denied", domain.FullLen),
		}), nil
	}

	reply := make(chan agent.PermissionDecision, 1)
	resolved := make(chan struct{})
	defer close(resolved)
	ask := agent.PermissionAsk{Request: domain, Reply: reply, Resolved: resolved}
	select {
	case b.asks <- ask:
	case <-b.hubDone:
		return unavailable()
	case <-b.stop:
		return unavailable()
	}
	select {
	case d := <-reply:
		if d.Kind != agent.DecisionDeny && gone() {
			return unavailable() // never deliver an approval after shutdown began
		}
		return toSDKDecision(d), nil
	case <-b.hubDone:
		return unavailable()
	case <-b.stop:
		return unavailable()
	}
}

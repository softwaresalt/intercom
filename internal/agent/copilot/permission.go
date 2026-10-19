package copilot

import (
	"log/slog"
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

func truncateRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n]) + "…"
}

// toDomainRequest converts an SDK permission request to the intercom type.
func toDomainRequest(req copilot.PermissionRequest, inv copilot.PermissionInvocation) agent.PermissionRequest {
	out := agent.PermissionRequest{SessionID: inv.SessionID}
	if req == nil {
		return out
	}
	out.Kind = string(req.Kind())
	switch r := req.(type) {
	case *rpc.PermissionRequestShell:
		out.Summary = truncateRunes(r.FullCommandText, maxSummaryRunes)
		if r.ToolCallID != nil {
			out.ToolCallID = *r.ToolCallID
		}
	default:
		// Other kinds carry only their kind until a consumer needs more;
		// the live-SDK test records which kinds reach the handler (ADR MVP-1).
		out.Summary = out.Kind
	}
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
// value the hub sent, or UserNotAvailable when the hub-done channel closes
// first. It never approves on its own.
func (b *permissionBridge) handle(req copilot.PermissionRequest, inv copilot.PermissionInvocation) (rpc.PermissionDecision, error) {
	unavailable := func() (rpc.PermissionDecision, error) {
		return &rpc.PermissionDecisionUserNotAvailable{}, nil
	}
	select {
	case <-b.hubDone:
		return unavailable()
	case <-b.stop:
		return unavailable()
	default:
	}

	reply := make(chan agent.PermissionDecision, 1)
	ask := agent.PermissionAsk{Request: toDomainRequest(req, inv), Reply: reply}
	select {
	case b.asks <- ask:
	case <-b.hubDone:
		return unavailable()
	case <-b.stop:
		return unavailable()
	}
	select {
	case d := <-reply:
		return toSDKDecision(d), nil
	case <-b.hubDone:
		return unavailable()
	case <-b.stop:
		return unavailable()
	}
}

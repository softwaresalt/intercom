// Package agent defines the intercom-owned domain types and the Agent
// interface that sit in front of the Copilot SDK.
//
// It is the inside of the anti-corruption layer (design §7.1): nothing in
// this package may import github.com/github/copilot-sdk. The single adapter
// that does lives in internal/agent/copilot and converts SDK values to and
// from the types declared here.
package agent

import "context"

// EventKind is the intercom-owned classification of an agent event.
type EventKind string

// Event kinds produced by the adapter's normalization table (design §4.3,
// §5.2). KindUnknown is an explicit result for SDK event variants the
// adapter does not map; it is never a silent drop.
const (
	KindAgentMessage EventKind = "agent.message"
	KindDelta        EventKind = "agent.delta"
	KindTurnStart    EventKind = "turn.start"
	KindTurnEnd      EventKind = "turn.end"
	KindIdle         EventKind = "session.idle"
	KindToolStarted  EventKind = "tool.started"
	KindToolEnded    EventKind = "tool.ended"
	KindUnknown      EventKind = "unknown"
)

// Event is a normalized agent event. It carries no SDK types. Fields not
// relevant to a Kind are left at their zero value.
type Event struct {
	Kind EventKind
	// MessageID identifies the assistant message (agent.message, agent.delta).
	MessageID string
	// Text is message content or a streaming delta.
	Text string
	// TurnID identifies the agent-loop turn (turn.start, turn.end).
	TurnID string
	// ToolCallID and ToolName identify a tool invocation (tool.*).
	ToolCallID string
	ToolName   string
	// Success reports tool completion success (tool.ended).
	Success bool
	// Aborted reports that an idle followed an abort (session.idle).
	Aborted bool
	// RawType is the SDK event type name; set for KindUnknown so the hub can
	// see what was not understood.
	RawType string
}

// SessionHandle identifies the single agent session owned by a process
// (ADR MVP-D7).
type SessionHandle struct {
	ID string
}

// PermissionRequest is an intercom-owned description of a tool permission
// request. It deliberately carries a bounded summary instead of the SDK
// request value.
type PermissionRequest struct {
	SessionID  string
	ToolCallID string
	// Kind is the SDK's request discriminator (shell, write, read, ...).
	Kind string
	// Summary is a human-readable description for the operator (for a shell
	// request, the command text). It is bounded: adapters truncate it with an
	// ellipsis at a fixed rune limit before it crosses into the hub.
	Summary string
}

// DecisionKind is the intercom-owned permission outcome. The zero value is
// deliberately invalid so an unset decision can never be interpreted as an
// approval (fail-closed).
type DecisionKind uint8

const (
	// DecisionAllowOnce approves exactly this request.
	DecisionAllowOnce DecisionKind = iota + 1
	// DecisionDeny rejects the request.
	DecisionDeny
	// DecisionUnavailable reports that no operator could be asked.
	DecisionUnavailable
)

// Valid reports whether k is a defined decision kind.
func (k DecisionKind) Valid() bool {
	return k >= DecisionAllowOnce && k <= DecisionUnavailable
}

// PermissionDecision is the operator's answer to a PermissionRequest.
type PermissionDecision struct {
	Kind DecisionKind
	// Feedback is optional operator text explaining a denial.
	Feedback string
}

// PermissionAsk is sent to the hub for each permission request. The hub
// answers by sending exactly one decision on Reply (capacity 1, so the hub
// never blocks on it).
type PermissionAsk struct {
	Request PermissionRequest
	Reply   chan<- PermissionDecision
}

// Agent is the interface the hub uses to drive an agent runtime. One Agent
// owns one session (ADR MVP-D7). Events and permission asks flow out through
// channels supplied at construction, not through this interface.
type Agent interface {
	// Start launches the runtime, asserts its version, and creates the
	// single session.
	Start(ctx context.Context) (SessionHandle, error)
	// Send submits a prompt. ctx must carry a deadline or a default applies.
	Send(ctx context.Context, prompt string) error
	// Abort cancels the in-flight turn. ctx must carry a deadline or a
	// default applies.
	Abort(ctx context.Context) error
	// Stop quiesces callbacks and shuts the runtime down within ctx.
	Stop(ctx context.Context) error
}

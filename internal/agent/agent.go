// Package agent defines the intercom-owned domain types and the Agent
// interface that sit in front of the Copilot SDK.
//
// It is the inside of the anti-corruption layer (design §7.1): nothing in
// this package may import github.com/github/copilot-sdk. The single adapter
// that does lives in internal/agent/copilot and converts SDK values to and
// from the types declared here.
package agent

import (
	"context"
	"errors"
)

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

// PermissionKind is the intercom-owned classification of a permission
// request. Hub authorization policy keys on this value, never on SDK
// discriminator strings (design §7.1). The zero value is Unknown and policy
// must treat Unknown and Other as the most restrictive class.
type PermissionKind uint8

const (
	// PermissionKindUnknown is the zero value: unclassified.
	PermissionKindUnknown PermissionKind = iota
	PermissionKindShell
	PermissionKindWrite
	PermissionKindRead
	PermissionKindURL
	// PermissionKindOther is any SDK kind the adapter does not model.
	PermissionKindOther
)

// PermissionRequest is an intercom-owned description of a tool permission
// request. It deliberately carries a bounded, sanitized summary instead of the
// SDK request value.
type PermissionRequest struct {
	SessionID  string
	ToolCallID string
	// Kind is the intercom classification; RawKind is the SDK discriminator,
	// kept for diagnostics only.
	Kind    PermissionKind
	RawKind string
	// Summary is a human-readable description for the operator (shell: the
	// command text; write/read: the path; url: the target). It is sanitized
	// (control and bidi-format runes replaced) and bounded: adapters truncate
	// it with an ellipsis at a fixed rune limit before it crosses into the hub.
	Summary string
	// Truncated reports that Summary does not show the whole request; FullLen
	// is the original length in runes. A truncated request cannot be reviewed
	// in full, so the adapter denies such requests itself rather than asking.
	Truncated bool
	FullLen   int
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
//
// Reply is bidirectional for one reason: when a decision the handler never
// took must be withdrawn (its acknowledgement is overdue, or shutdown began),
// the hub receives it back out of the buffer and replaces it with
// UserNotAvailable, so an approval can never be delivered after the surfaces
// were told it was abandoned. The hub is the only party that does so; the
// adapter only ever receives.
//
// Resolved is closed when the adapter no longer needs an answer (the request
// was answered, or shutdown or hub-done abandoned it). A hub that dequeues an
// ask whose Resolved is already closed must discard it instead of prompting
// the operator. The hub owns any approval timeout; the adapter applies none.
//
// Ordering: asks travel on a different channel from events, so the relative
// order of an ask and the events around it is undefined; the hub must not
// infer order across the two.
type PermissionAsk struct {
	Request  PermissionRequest
	Reply    chan PermissionDecision
	Resolved <-chan struct{}
}

// Lifecycle sentinel errors returned (wrapped) by Agent implementations so
// callers can use errors.Is to tell an orderly shutdown from a real fault.
var (
	ErrStopped        = errors.New("agent stopped")
	ErrNotRunning     = errors.New("agent session not running")
	ErrAlreadyStarted = errors.New("agent already started")
)

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
	// Stop quiesces callbacks and shuts the runtime down within ctx. Calling
	// order (design §3.1): Abort, wait for the idle event on ingress, then
	// Stop; events still in flight when Stop begins may be discarded (logged
	// at warn), so Stop must follow an observed idle. Stop also interrupts a
	// Start in progress and is idempotent.
	Stop(ctx context.Context) error
}

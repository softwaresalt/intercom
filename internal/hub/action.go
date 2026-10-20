package hub

import (
	"context"

	"github.com/softwaresalt/intercom-go/internal/agent"
)

// Role is the authorization role of a surface (ADR MVP-3). It is fixed when a
// client is registered, from trusted state (the credential record that
// authenticated a remote connection, or construction for the in-process TUI),
// and is never read from an inbound message. The zero value carries no
// privilege: a remote client registered without a role can only observe.
type Role uint8

const (
	// RoleNone is the zero value: observe only. Fail-closed default.
	RoleNone Role = iota
	// RoleObserver may receive every envelope and perform no privileged
	// action.
	RoleObserver
	// RoleApprover may steer the agent and answer permission requests, subject
	// to the destructive-approval policy for remote surfaces.
	RoleApprover
)

// surface distinguishes the in-process TUI from remote clients.
type surface uint8

const (
	surfaceRemote surface = iota + 1
	surfaceTUI
)

// principal is the identity an action is performed as. It is built by the hub
// from a registered [Client]'s own fields (kind and role), never from the
// action's content.
type principal struct {
	surface surface
	role    Role
}

// Actor labels carried in permission.resolved and action.refused (ADR MVP-3:
// "the winning actor as the role and surface", never a credential).
const (
	// ActorLocalTUI is the in-process Bubble Tea surface.
	ActorLocalTUI = "local-tui"
	// ActorRemoteApprover is a remote client holding the approver role.
	ActorRemoteApprover = "remote-approver"
	// ActorRemoteObserver is a remote client holding the observer role (it
	// can never win an arbitration; the label exists for refusals).
	ActorRemoteObserver = "remote-observer"
	// ActorHub is the hub itself (timeout, shutdown, empty broadcast set).
	ActorHub = "hub"
)

// label is the actor label for p.
func (p principal) label() string {
	if p.surface == surfaceTUI {
		return ActorLocalTUI
	}
	if p.role == RoleApprover {
		return ActorRemoteApprover
	}
	return ActorRemoteObserver
}

// Reason is a stable refusal code (ADR MVP-3). A refused action returns one
// to the refused surface only; it is never broadcast.
type Reason string

// Stable reason codes. The last two are hub extensions beyond the ADR's list,
// needed for conditions the ADR does not cover (see the 067-F report).
const (
	ReasonRoleNotPermitted         Reason = "role_not_permitted"
	ReasonDestructiveRemoteRefused Reason = "destructive_remote_refused"
	ReasonAlreadyResolved          Reason = "already_resolved"
	ReasonRequestUnknown           Reason = "request_unknown"
	ReasonActionUnknown            Reason = "action_unknown"
	ReasonCredentialRevoked        Reason = "credential_revoked"
	// ReasonAgentUnavailable: the hub cannot hand the action to the agent now
	// (no agent configured, or its operator queue is full).
	ReasonAgentUnavailable Reason = "agent_unavailable"
	// ReasonShuttingDown: the hub is shutting down (design §3.1 step 0) and no
	// longer accepts privileged actions from the credential-less local TUI.
	ReasonShuttingDown Reason = "shutting_down"
)

// OpType is the kind of operator action. The zero value is unknown and is
// refused (default deny).
type OpType uint8

const (
	// OpUnknown is the zero value and any unrecognised or malformed type.
	OpUnknown OpType = iota
	// OpResolvePermission answers a pending permission request.
	OpResolvePermission
	// OpPrompt sends a prompt to the agent.
	OpPrompt
	// OpAbortTurn aborts the in-flight turn.
	OpAbortTurn
)

// OperatorAction is one inbound operator action. Everything in it is
// client-supplied and therefore untrusted: authorization never reads the
// permission kind, the role or the surface from here (ADR MVP-3).
type OperatorAction struct {
	Type OpType
	// RequestID names the pending permission for OpResolvePermission.
	RequestID string
	// Decision is the answer for OpResolvePermission: allow_once or deny.
	Decision agent.PermissionDecision
	// Prompt is the text for OpPrompt.
	Prompt string
}

// ActionResult is the hub's verdict on one action. A refusal is explicit and
// goes only to the acting surface.
type ActionResult struct {
	// Accepted reports that the action took effect (or was handed off).
	Accepted bool
	// Reason is set when the action was refused.
	Reason Reason
	// Winner is the actor that won the arbitration, set with
	// ReasonAlreadyResolved.
	Winner string
}

// AgentControl is the narrow slice of the agent the hub drives on behalf of an
// operator. agent.Agent satisfies it; the hub never calls it on its own
// goroutine.
type AgentControl interface {
	// Send submits a prompt.
	Send(ctx context.Context, prompt string) error
	// Abort cancels the in-flight turn.
	Abort(ctx context.Context) error
}

// actionReq is the hub-internal message carrying an action and the identity
// it is performed as.
type actionReq struct {
	from principal
	// client is the registered client the action came from (nil for in-package
	// callers with no client). The hub re-checks it when it processes the
	// action, not only when Do queued it, so an action queued by a client that
	// was closed or revoked in between is refused.
	client *Client
	act    OperatorAction
	reply  chan ActionResult // capacity 1: the hub never blocks on it
}

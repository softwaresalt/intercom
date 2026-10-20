package hub

import (
	"github.com/softwaresalt/intercom-go/internal/agent"
)

// authzInput is everything the authorization decision may depend on. Every
// field is hub-owned state: p is built from the registered Client's own kind
// and role, kind is read from the pending-request record the adapter's
// classification created, and the flag is hub configuration. Nothing here is
// copied from an inbound message (ADR MVP-3).
type authzInput struct {
	p        principal
	op       OpType
	kind     agent.PermissionKind // the pending entry's kind; resolve only
	decision agent.DecisionKind   // the answer; resolve only
	// remoteApproveDestructive is [remote].remote_approve_destructive.
	remoteApproveDestructive bool
}

// destructive reports whether approving a request of kind k needs the remote
// opt-in. Only read is non-destructive; shell, write, URL, every other kind and
// any kind the hub does not recognise are destructive (ADR MVP-3: unknown
// kinds are default-deny for remote surfaces). Moving a kind out of this set
// needs an ADR amendment, never a silent code change.
func destructive(k agent.PermissionKind) bool { return k != agent.PermissionKindRead }

// authorize is the ADR MVP-3 action-by-role matrix as a pure function. It is
// fail-closed: anything it does not recognise is refused.
func authorize(in authzInput) (bool, Reason) {
	switch in.op {
	case OpResolvePermission, OpPrompt, OpAbortTurn:
	default:
		return false, ReasonActionUnknown
	}
	if in.p.surface == surfaceTUI {
		return true, "" // the local TUI is always an approver, with no credential
	}
	if in.p.surface != surfaceRemote || in.p.role != RoleApprover {
		return false, ReasonRoleNotPermitted // observers, role-less and corrupt principals
	}
	if in.op == OpResolvePermission && in.decision != agent.DecisionDeny &&
		destructive(in.kind) && !in.remoteApproveDestructive {
		return false, ReasonDestructiveRemoteRefused
	}
	return true, ""
}

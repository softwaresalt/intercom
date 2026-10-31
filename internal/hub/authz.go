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
	// previewClipped is set for a write permission whose preview was clipped.
	// Approving it is local-TUI-only and is never unlocked by
	// remoteApproveDestructive: the approver cannot see the whole change.
	previewClipped bool
}

// knownDestructive reports whether k is a modeled destructive kind (shell,
// write, URL). A remote approver may approve these only with the
// [remote].remote_approve_destructive opt-in. Moving a kind in or out of this
// set needs an ADR amendment, never a silent code change.
func knownDestructive(k agent.PermissionKind) bool {
	switch k {
	case agent.PermissionKindShell, agent.PermissionKindWrite, agent.PermissionKindURL:
		return true
	}
	return false
}

// unknownPermissionKind reports whether k is neither read nor a modeled
// destructive kind: PermissionKindUnknown, PermissionKindOther, and any
// undefined value. Such requests carry only a raw discriminator, so approving
// them is local-TUI-only and is never unlocked by the remote opt-in (ADR MVP-3,
// 074-S amendment).
func unknownPermissionKind(k agent.PermissionKind) bool {
	return k != agent.PermissionKindRead && !knownDestructive(k)
}

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
	if in.op == OpResolvePermission && in.decision != agent.DecisionDeny {
		if unknownPermissionKind(in.kind) || (in.previewClipped && in.kind == agent.PermissionKindWrite) {
			return false, ReasonDestructiveRemoteRefused // local-TUI-only; never unlocked by the opt-in
		}
		if knownDestructive(in.kind) && !in.remoteApproveDestructive {
			return false, ReasonDestructiveRemoteRefused
		}
	}
	return true, ""
}

// Package apperr defines the AppError taxonomy ported from the
// agent-intercom behavioral oracle (softwaresalt/agent-intercom @ 41df772,
// src/errors.rs). It provides a fixed, 14-variant error kind taxonomy with a
// pinned, lowercase-prefixed display contract that later phases and the
// oracle parity corpus depend on.
package apperr

// Kind identifies the taxonomy variant of an *Error. The set of kinds is a
// frozen, operator-visible contract (Protected Invariant I4): adding,
// removing, or renaming a rendered prefix is a breaking change requiring a
// new deliberation.
type Kind int

const (
	// KindConfig indicates a configuration-loading or -validation failure.
	KindConfig Kind = iota
	// KindDB indicates a database-layer failure.
	KindDB
	// KindSlack indicates a Slack-integration failure.
	KindSlack
	// KindMCP indicates an MCP (Model Context Protocol) surface failure.
	KindMCP
	// KindDiff indicates a diff-generation or diff-application failure.
	KindDiff
	// KindPolicy indicates a policy-evaluation failure.
	KindPolicy
	// KindIPC indicates an inter-process-communication failure.
	KindIPC
	// KindPathViolation indicates a workspace path-containment violation.
	KindPathViolation
	// KindPatchConflict indicates a patch application conflict.
	KindPatchConflict
	// KindNotFound indicates a requested resource was not found.
	KindNotFound
	// KindUnauthorized indicates an authorization failure.
	KindUnauthorized
	// KindAlreadyConsumed indicates a single-use resource was already consumed.
	KindAlreadyConsumed
	// KindIO indicates a generic input/output failure.
	KindIO
	// KindACP indicates an Agent Client Protocol failure.
	KindACP
)

// prefix returns the exact, lowercase display prefix for the kind, pinned to
// the oracle's src/errors.rs display contract. The returned string never
// changes across kinds (see the uniqueness test in apperr_test.go).
func (k Kind) prefix() string {
	switch k {
	case KindConfig:
		return "config"
	case KindDB:
		return "db"
	case KindSlack:
		return "slack"
	case KindMCP:
		return "mcp"
	case KindDiff:
		return "diff"
	case KindPolicy:
		return "policy"
	case KindIPC:
		return "ipc"
	case KindPathViolation:
		return "path violation"
	case KindPatchConflict:
		return "patch conflict"
	case KindNotFound:
		return "not found"
	case KindUnauthorized:
		return "unauthorized"
	case KindAlreadyConsumed:
		return "already consumed"
	case KindIO:
		return "io"
	case KindACP:
		return "acp"
	default:
		return "unknown"
	}
}

// Error is the taxonomy's concrete error type. Fields are unexported so a
// caller cannot re-categorize an error after construction (finding GO-8 —
// this is a security-relevant taxonomy). All methods use pointer receivers
// and every constructor returns *Error (finding GO-9).
type Error struct {
	kind  Kind
	msg   string
	cause error
}

// Kind returns the taxonomy variant of the error.
func (e *Error) Kind() Kind {
	return e.kind
}

// Message returns the error's message text, excluding the kind prefix.
func (e *Error) Message() string {
	return e.msg
}

// Error renders the error as "{prefix}: {msg}". The cause, if any, is never
// appended: the oracle's display strings are a pinned, operator-visible
// contract.
func (e *Error) Error() string {
	return e.kind.prefix() + ": " + e.msg
}

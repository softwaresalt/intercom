// Package apperr defines the AppError taxonomy ported from the
// agent-intercom behavioral oracle (softwaresalt/agent-intercom @ 41df772,
// src/errors.rs). It provides a fixed, 11-variant error kind taxonomy with a
// pinned, lowercase-prefixed display contract that later phases and the
// oracle parity corpus depend on.
package apperr

import "fmt"

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
	// KindMCP indicates an MCP (Model Context Protocol) surface failure.
	KindMCP
	// KindDiff indicates a diff-generation or diff-application failure.
	KindDiff
	// KindPolicy indicates a policy-evaluation failure.
	KindPolicy
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
	case KindMCP:
		return "mcp"
	case KindDiff:
		return "diff"
	case KindPolicy:
		return "policy"
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
	default:
		return "unknown"
	}
}

// String returns the Kind's exported name (e.g. "NotFound"), satisfying
// fmt.Stringer. This is a debugging/observability affordance and is
// deliberately separate from the unexported prefix() display-contract
// accessor used by Error.Error(): collapsing the two would couple log
// formatting to the pinned, wire-visible error text.
func (k Kind) String() string {
	switch k {
	case KindConfig:
		return "Config"
	case KindDB:
		return "DB"
	case KindMCP:
		return "MCP"
	case KindDiff:
		return "Diff"
	case KindPolicy:
		return "Policy"
	case KindPathViolation:
		return "PathViolation"
	case KindPatchConflict:
		return "PatchConflict"
	case KindNotFound:
		return "NotFound"
	case KindUnauthorized:
		return "Unauthorized"
	case KindAlreadyConsumed:
		return "AlreadyConsumed"
	case KindIO:
		return "IO"
	default:
		return "Unknown"
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

// Unwrap returns only the cause, if any. Is (kind) and Unwrap (cause) are
// deliberately separate mechanisms: a single Unwrap cannot serve both roles
// (stash finding P1-j). Unlike the oracle, the cause is retained here so
// errors.Is/errors.As can traverse it — a recorded improvement that does not
// alter the display contract.
func (e *Error) Unwrap() error {
	return e.cause
}

// New constructs an *Error of the given kind with the given message.
func New(kind Kind, msg string) *Error {
	return &Error{kind: kind, msg: msg}
}

// Newf constructs an *Error of the given kind with a formatted message.
//
// Newf's format string is not checked by go vet's printf analyzer, since
// custom "…f" functions outside the fmt package are not recognized by that
// analyzer (finding GO-11). Call sites should be reviewed manually.
func Newf(kind Kind, format string, args ...any) *Error {
	return &Error{kind: kind, msg: fmt.Sprintf(format, args...)}
}

// Wrap constructs an *Error of the given kind whose message is cause.Error(),
// reproducing the oracle's three From impls (toml::de::Error, sqlx::Error,
// std::io::Error), which stringify the source. The cause is retained for
// Unwrap even though it is not rendered by Error().
//
// Wrap guards cause == nil and falls back to an empty message rather than
// panicking on a nil-interface method call (finding GO-3).
func Wrap(kind Kind, cause error) *Error {
	msg := ""
	if cause != nil {
		msg = cause.Error()
	}
	return &Error{kind: kind, msg: msg, cause: cause}
}

// Wrapf constructs an *Error of the given kind with a formatted message,
// while also preserving cause for Unwrap()/errors.Is/errors.As traversal.
// Neither Wrap (no message parameter; derives msg from cause.Error()) nor
// New (accepts a message but drops the cause) can both preserve a cause and
// retain human-readable context — Wrapf fills that gap.
//
// Like Newf, Wrapf's format string is not checked by go vet's printf
// analyzer unless explicitly registered (see .golangci.yml's
// linters.settings.govet.settings.printf.funcs); Wrapf is registered
// alongside Newf for exactly this reason.
//
// Wrapf guards cause == nil, matching Wrap's existing nil-cause handling.
func Wrapf(kind Kind, cause error, format string, args ...any) *Error {
	return &Error{kind: kind, msg: fmt.Sprintf(format, args...), cause: cause}
}

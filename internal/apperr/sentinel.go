package apperr

// kindSentinel is an unexported, immutable value type implementing error,
// used solely to identify a Kind for errors.Is comparisons. It is a value
// type (not *Error) so that sentinels cannot be mutated after construction —
// a *Error sentinel would be a mutable global hazard on a security-relevant
// taxonomy (finding GO-1 — P0).
type kindSentinel Kind

// Error satisfies the error interface for kindSentinel. Its own rendering is
// not part of any pinned display contract; kindSentinel values are only ever
// compared via errors.Is, never displayed directly.
func (k kindSentinel) Error() string {
	return Kind(k).prefix()
}

// Exported kind sentinels, one per taxonomy variant. Use with errors.Is,
// e.g. errors.Is(err, apperr.ErrNotFound).
var (
	ErrConfig          error = kindSentinel(KindConfig)
	ErrDB              error = kindSentinel(KindDB)
	ErrSlack           error = kindSentinel(KindSlack)
	ErrMCP             error = kindSentinel(KindMCP)
	ErrDiff            error = kindSentinel(KindDiff)
	ErrPolicy          error = kindSentinel(KindPolicy)
	ErrIPC             error = kindSentinel(KindIPC)
	ErrPathViolation   error = kindSentinel(KindPathViolation)
	ErrPatchConflict   error = kindSentinel(KindPatchConflict)
	ErrNotFound        error = kindSentinel(KindNotFound)
	ErrUnauthorized    error = kindSentinel(KindUnauthorized)
	ErrAlreadyConsumed error = kindSentinel(KindAlreadyConsumed)
	ErrIO              error = kindSentinel(KindIO)
	ErrACP             error = kindSentinel(KindACP)
)

// Is reports whether target is the kind sentinel matching e's kind. It
// performs a direct type assertion to kindSentinel and never calls errors.Is
// internally, so no recursion is possible. Matching is by kind: because
// (*Error).Unwrap is also implemented (unit A3), errors.Is(err, ErrX)
// matches if KindX appears anywhere in err's cause chain, not only at the
// top level (finding GO-2).
func (e *Error) Is(target error) bool {
	k, ok := target.(kindSentinel)
	if !ok {
		return false
	}
	return e.kind == Kind(k)
}

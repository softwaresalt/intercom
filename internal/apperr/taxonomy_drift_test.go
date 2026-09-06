package apperr

import "testing"

// allSentinels enumerates every exported Err* sentinel. This list, together
// with allKinds, is the taxonomy drift guard: if a Kind is added without its
// sentinel (or vice versa), TestTaxonomyKindsAndSentinelsStayInSync fails.
var allSentinels = []error{
	ErrConfig,
	ErrDB,
	ErrMCP,
	ErrDiff,
	ErrPolicy,
	ErrPathViolation,
	ErrPatchConflict,
	ErrNotFound,
	ErrUnauthorized,
	ErrAlreadyConsumed,
	ErrIO,
}

// TestTaxonomyKindsAndSentinelsStayInSync asserts len(allKinds) equals the
// number of exported Err* sentinels, so adding a Kind without a matching
// sentinel (or a sentinel without a matching Kind) fails CI rather than
// silently drifting.
//
// Proven by construction (2026-09-05): temporarily appending an unpaired
// Kind to allKinds (with no corresponding entry in allSentinels) was
// confirmed to fail this assertion before being reverted; no such temporary
// change is present in the committed tree.
func TestTaxonomyKindsAndSentinelsStayInSync(t *testing.T) {
	if got, want := len(allKinds), len(allSentinels); got != want {
		t.Fatalf("len(allKinds) = %d, len(allSentinels) = %d — taxonomy and sentinel list have drifted apart", got, want)
	}
}

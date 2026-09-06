package apperr

import (
	"fmt"
	"testing"
)

// kindNames maps each retained Kind to its expected String() name, used for
// exhaustive table verification.
var kindNames = map[Kind]string{
	KindConfig:          "Config",
	KindDB:              "DB",
	KindMCP:             "MCP",
	KindDiff:            "Diff",
	KindPolicy:          "Policy",
	KindPathViolation:   "PathViolation",
	KindPatchConflict:   "PatchConflict",
	KindNotFound:        "NotFound",
	KindUnauthorized:    "Unauthorized",
	KindAlreadyConsumed: "AlreadyConsumed",
	KindIO:              "IO",
}

// TestKindSatisfiesStringer verifies Kind implements fmt.Stringer and that
// %v renders the Kind's name rather than its underlying integer value.
func TestKindSatisfiesStringer(t *testing.T) {
	var _ fmt.Stringer = KindConfig

	if got, want := fmt.Sprintf("%v", KindNotFound), "NotFound"; got != want {
		t.Fatalf("fmt.Sprintf(%%v, KindNotFound) = %q, want %q", got, want)
	}
}

// TestKindStringExhaustive asserts String() covers all 11 retained Kinds by
// name.
func TestKindStringExhaustive(t *testing.T) {
	if got, want := len(allKinds), len(kindNames); got != want {
		t.Fatalf("allKinds has %d entries, kindNames table has %d — table is not exhaustive", got, want)
	}
	for _, k := range allKinds {
		want, ok := kindNames[k]
		if !ok {
			t.Fatalf("kind %v has no expected name in kindNames table", k)
		}
		if got := k.String(); got != want {
			t.Fatalf("Kind(%d).String() = %q, want %q", int(k), got, want)
		}
	}
}

// TestKindStringOutOfRangeFallback verifies String() has a defined fallback
// for an out-of-range Kind value rather than panicking or returning empty.
func TestKindStringOutOfRangeFallback(t *testing.T) {
	invalid := Kind(999)
	if got := invalid.String(); got == "" {
		t.Fatalf("Kind(999).String() = %q, want a non-empty fallback", got)
	}
}

// TestErrorRenderingUnchangedByStringer is the golden test for U2 AC-2: adding
// Stringer must not alter Error()'s pinned, prefix-based display contract for
// any retained Kind, since Error() calls prefix() explicitly and never routes
// through %v/String().
func TestErrorRenderingUnchangedByStringer(t *testing.T) {
	want := map[Kind]string{
		KindConfig:          "config: x",
		KindDB:              "db: x",
		KindMCP:             "mcp: x",
		KindDiff:            "diff: x",
		KindPolicy:          "policy: x",
		KindPathViolation:   "path violation: x",
		KindPatchConflict:   "patch conflict: x",
		KindNotFound:        "not found: x",
		KindUnauthorized:    "unauthorized: x",
		KindAlreadyConsumed: "already consumed: x",
		KindIO:              "io: x",
	}
	for _, k := range allKinds {
		e := &Error{kind: k, msg: "x"}
		if got := e.Error(); got != want[k] {
			t.Fatalf("kind %v: Error() = %q, want %q (Stringer must not affect the display contract)", k, got, want[k])
		}
	}
}

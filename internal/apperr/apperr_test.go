package apperr

import (
	"strings"
	"testing"
)

// allKinds enumerates every taxonomy member for exhaustive table tests.
var allKinds = []Kind{
	KindConfig,
	KindDB,
	KindMCP,
	KindDiff,
	KindPolicy,
	KindPathViolation,
	KindPatchConflict,
	KindNotFound,
	KindUnauthorized,
	KindAlreadyConsumed,
	KindIO,
}

// TestPrefixUniquenessAndNoTrailingPeriod asserts, across all 11 kinds, that
// each rendered prefix is unique and that no rendered message ends in '.'.
func TestPrefixUniquenessAndNoTrailingPeriod(t *testing.T) {
	if got := len(allKinds); got != 11 {
		t.Fatalf("expected 11 kinds, got %d", got)
	}

	seen := make(map[string]Kind, len(allKinds))
	for _, k := range allKinds {
		p := k.prefix()
		if other, ok := seen[p]; ok {
			t.Fatalf("prefix %q is not unique: shared by %v and %v", p, other, k)
		}
		seen[p] = k

		e := &Error{kind: k, msg: "example message"}
		if strings.HasSuffix(e.Error(), ".") {
			t.Fatalf("kind %v rendered message %q ends in a trailing period", k, e.Error())
		}
	}
}

// TestPathViolationErrorDisplay pins the exact display contract for a
// representative multi-word-prefix kind.
func TestPathViolationErrorDisplay(t *testing.T) {
	e := &Error{kind: KindPathViolation, msg: "stream closed"}
	const want = "path violation: stream closed"
	if got := e.Error(); got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

// TestKindAndMessageAccessors verifies the read-only accessors surface the
// unexported fields without allowing mutation.
func TestKindAndMessageAccessors(t *testing.T) {
	e := &Error{kind: KindIO, msg: "disk full"}
	if e.Kind() != KindIO {
		t.Fatalf("Kind() = %v, want %v", e.Kind(), KindIO)
	}
	if e.Message() != "disk full" {
		t.Fatalf("Message() = %q, want %q", e.Message(), "disk full")
	}
}

// TestErrorSatisfiesErrorInterface confirms *Error implements the error interface.
func TestErrorSatisfiesErrorInterface(t *testing.T) {
	var err error = &Error{kind: KindDB, msg: "connection refused"}
	if err.Error() != "db: connection refused" {
		t.Fatalf("Error() = %q, want %q", err.Error(), "db: connection refused")
	}
}

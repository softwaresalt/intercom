package apperr

import (
	"errors"
	"testing"
)

// TestErrorsIsMatchesOwnKindOnly verifies errors.Is matches the sentinel
// corresponding to an Error's own kind, and does not match an unrelated
// sentinel. Constructed via a package-internal struct literal because the
// constructors (New/Newf/Wrap) do not exist until unit A3 (finding CORR-1).
func TestErrorsIsMatchesOwnKindOnly(t *testing.T) {
	e := &Error{kind: KindNotFound, msg: "x"}

	if !errors.Is(e, ErrNotFound) {
		t.Fatalf("errors.Is(e, ErrNotFound) = false, want true")
	}
	if errors.Is(e, ErrDB) {
		t.Fatalf("errors.Is(e, ErrDB) = true, want false")
	}
}

// sentinelForKind maps each Kind to its exported sentinel for table-driven
// verification.
func sentinelForKind(k Kind) error {
	switch k {
	case KindConfig:
		return ErrConfig
	case KindDB:
		return ErrDB
	case KindMCP:
		return ErrMCP
	case KindDiff:
		return ErrDiff
	case KindPolicy:
		return ErrPolicy
	case KindPathViolation:
		return ErrPathViolation
	case KindPatchConflict:
		return ErrPatchConflict
	case KindNotFound:
		return ErrNotFound
	case KindUnauthorized:
		return ErrUnauthorized
	case KindAlreadyConsumed:
		return ErrAlreadyConsumed
	case KindIO:
		return ErrIO
	default:
		return nil
	}
}

// TestAllSentinelsMatchOwnKindOnly is a table test asserting each of the 11
// sentinels matches its own kind and no other.
func TestAllSentinelsMatchOwnKindOnly(t *testing.T) {
	for _, k := range allKinds {
		e := &Error{kind: k, msg: "probe"}
		want := sentinelForKind(k)
		if !errors.Is(e, want) {
			t.Fatalf("kind %v: errors.Is did not match its own sentinel", k)
		}
		for _, other := range allKinds {
			if other == k {
				continue
			}
			otherSentinel := sentinelForKind(other)
			if errors.Is(e, otherSentinel) {
				t.Fatalf("kind %v: errors.Is unexpectedly matched sentinel for kind %v", k, other)
			}
		}
	}
}

// Note: the deep-cause-chain matching behavior described in this unit's
// third acceptance criterion (finding GO-2) requires (*Error).Unwrap, which
// is introduced by unit A3. That specific scenario is exercised by
// TestIsMatchesKindDeepInCauseChain in wrap_test.go once Unwrap exists; this
// file covers same-kind and cross-kind sentinel matching at the top level.

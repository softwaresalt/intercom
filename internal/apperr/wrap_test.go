package apperr

import (
	"errors"
	"fmt"
	"os"
	"testing"
)

// TestWrapPreservesCauseForErrorsIs verifies Wrap(KindIO, os.ErrNotExist):
// errors.Is(err, os.ErrNotExist) is true and err.Error() equals
// "io: " + os.ErrNotExist.Error() — the cause is reachable but invisible in
// the display.
func TestWrapPreservesCauseForErrorsIs(t *testing.T) {
	err := Wrap(KindIO, os.ErrNotExist)

	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("errors.Is(err, os.ErrNotExist) = false, want true")
	}
	want := "io: " + os.ErrNotExist.Error()
	if got := err.Error(); got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}

// TestWrapNilCauseDoesNotPanic verifies Wrap(kind, nil) does not panic and
// renders "io: " (finding GO-3).
func TestWrapNilCauseDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Wrap(KindIO, nil) panicked: %v", r)
		}
	}()

	err := Wrap(KindIO, nil)
	if got, want := err.Error(), "io: "; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
	if err.Unwrap() != nil {
		t.Fatalf("Unwrap() = %v, want nil", err.Unwrap())
	}
}

// TestErrorsAsRecoversOutermostAndWrapped verifies errors.As recovers a
// *Error and its Kind() both when *Error is outermost and when wrapped by
// fmt.Errorf("...: %w", ...) (finding GO-10).
func TestErrorsAsRecoversOutermostAndWrapped(t *testing.T) {
	base := New(KindNotFound, "widget missing")

	var outermost *Error
	if !errors.As(base, &outermost) {
		t.Fatalf("errors.As did not recover outermost *Error")
	}
	if outermost.Kind() != KindNotFound {
		t.Fatalf("outermost.Kind() = %v, want %v", outermost.Kind(), KindNotFound)
	}

	wrapped := fmt.Errorf("lookup failed: %w", base)
	var recovered *Error
	if !errors.As(wrapped, &recovered) {
		t.Fatalf("errors.As did not recover *Error through fmt.Errorf wrapping")
	}
	if recovered.Kind() != KindNotFound {
		t.Fatalf("recovered.Kind() = %v, want %v", recovered.Kind(), KindNotFound)
	}
}

// TestIsMatchesKindDeepInCauseChain exercises the deep-cause-chain matching
// scenario described by unit A2's third acceptance criterion (finding GO-2),
// now that Unwrap exists.
func TestIsMatchesKindDeepInCauseChain(t *testing.T) {
	inner := New(KindIO, "disk full")
	outer := Wrap(KindConfig, inner)

	if !errors.Is(outer, ErrConfig) {
		t.Fatalf("errors.Is(outer, ErrConfig) = false, want true")
	}
	if !errors.Is(outer, ErrIO) {
		t.Fatalf("errors.Is(outer, ErrIO) = false, want true (kind found via cause chain)")
	}
}

// TestNewAndNewf cover the two simple constructors.
func TestNewAndNewf(t *testing.T) {
	e := New(KindPolicy, "denied")
	if e.Kind() != KindPolicy || e.Message() != "denied" {
		t.Fatalf("New() = %+v, want kind=%v msg=%q", e, KindPolicy, "denied")
	}

	f := Newf(KindPolicy, "denied for user %s", "alice")
	if got, want := f.Message(), "denied for user alice"; got != want {
		t.Fatalf("Newf().Message() = %q, want %q", got, want)
	}
}

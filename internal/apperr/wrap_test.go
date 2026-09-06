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

// TestWrapfSetsFormattedMessageAndPreservesCause verifies Wrapf both
// formats a human-readable message AND preserves cause for errors.Is/As
// through Unwrap() — the gap neither Wrap (no message param) nor New (drops
// cause) can fill (finding: 010-F U1 blocker).
func TestWrapfSetsFormattedMessageAndPreservesCause(t *testing.T) {
	err := Wrapf(KindIO, os.ErrNotExist, "opening %s", "config.toml")

	const wantMsg = "opening config.toml"
	if got := err.Message(); got != wantMsg {
		t.Fatalf("Message() = %q, want %q", got, wantMsg)
	}
	const wantRendered = "io: opening config.toml"
	if got := err.Error(); got != wantRendered {
		t.Fatalf("Error() = %q, want %q", got, wantRendered)
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("errors.Is(err, os.ErrNotExist) = false, want true — cause must be reachable via Unwrap")
	}
}

// TestWrapfKindDiscriminationUnaffected verifies errors.Is against the Kind
// sentinel still matches for a Wrapf-constructed error (Kind discrimination
// must not regress).
func TestWrapfKindDiscriminationUnaffected(t *testing.T) {
	err := Wrapf(KindNotFound, os.ErrNotExist, "widget %d", 42)

	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("errors.Is(err, ErrNotFound) = false, want true")
	}
	if errors.Is(err, ErrDB) {
		t.Fatalf("errors.Is(err, ErrDB) = true, want false")
	}
}

// TestWrapfNilCauseDoesNotPanic verifies Wrapf(kind, nil, ...) does not
// panic, matching Wrap's existing nil-cause handling, and still renders the
// formatted message.
func TestWrapfNilCauseDoesNotPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Wrapf(KindIO, nil, ...) panicked: %v", r)
		}
	}()

	err := Wrapf(KindIO, nil, "opening %s", "config.toml")
	if got, want := err.Error(), "io: opening config.toml"; got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
	if err.Unwrap() != nil {
		t.Fatalf("Unwrap() = %v, want nil", err.Unwrap())
	}
}

// TestWrapAndNewUnchangedByWrapf is a regression guard: adding Wrapf must
// not alter Wrap or New's existing behaviour.
func TestWrapAndNewUnchangedByWrapf(t *testing.T) {
	w := Wrap(KindIO, os.ErrNotExist)
	if got, want := w.Error(), "io: "+os.ErrNotExist.Error(); got != want {
		t.Fatalf("Wrap().Error() = %q, want %q", got, want)
	}

	n := New(KindNotFound, "widget missing")
	if got, want := n.Error(), "not found: widget missing"; got != want {
		t.Fatalf("New().Error() = %q, want %q", got, want)
	}
}

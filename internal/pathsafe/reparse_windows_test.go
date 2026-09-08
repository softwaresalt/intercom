//go:build windows

package pathsafe

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// canonicalForComparison normalizes want through filepath.EvalSymlinks so
// test assertions are not defeated by an incidental Windows 8.3 short-name
// alias in a t.TempDir() path (e.g. "DEWILL~1" vs "dewilliams") -- a naming
// artifact of the test environment, not a canonicalizeReparse defect.
// filepath.EvalSymlinks performs the same long-name normalization Windows
// applies internally, giving an independent oracle for the expected value.
func canonicalForComparison(t *testing.T, want string) string {
	t.Helper()
	resolved, err := filepath.EvalSymlinks(want)
	if err != nil {
		t.Fatalf("filepath.EvalSymlinks(%q) returned error: %v", want, err)
	}
	return stripUNCPrefix(resolved)
}

// TestCanonicalizeReparsePlainDirectory verifies an ordinary directory with
// no reparse point resolves to itself (014.003-T AC6, case 1).
func TestCanonicalizeReparsePlainDirectory(t *testing.T) {
	dir := t.TempDir()
	want := canonicalForComparison(t, dir)

	got, err := canonicalizeReparse(dir)
	if err != nil {
		t.Fatalf("canonicalizeReparse(%q) returned unexpected error: %v", dir, err)
	}
	got = stripUNCPrefix(got)
	if !strings.EqualFold(got, want) {
		t.Fatalf("canonicalizeReparse(%q) = %q, want %q", dir, got, want)
	}
}

// TestCanonicalizeReparseInRootJunction verifies a live directory junction
// whose target lies inside the workspace resolves to the real (in-root)
// target rather than the junction's own lexical path (014.003-T AC6, case
// 2).
func TestCanonicalizeReparseInRootJunction(t *testing.T) {
	rootDir := t.TempDir()
	root, err := NewRoot(rootDir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", rootDir, err)
	}

	targetDir := filepath.Join(root.Path(), "real-target")
	if err := os.Mkdir(targetDir, 0o755); err != nil {
		t.Fatalf("failed to create in-root junction target: %v", err)
	}

	linkPath := filepath.Join(root.Path(), "junction-in-root")
	createDirectoryJunction(t, linkPath, targetDir)

	got, err := canonicalizeReparse(linkPath)
	if err != nil {
		t.Fatalf("canonicalizeReparse(%q) returned unexpected error: %v", linkPath, err)
	}
	got = stripUNCPrefix(got)
	if !hasPathPrefix(got, root.Path()) {
		t.Fatalf("canonicalizeReparse(%q) = %q, want a path inside root %q", linkPath, got, root.Path())
	}
	want := canonicalForComparison(t, targetDir)
	if !strings.EqualFold(got, want) {
		t.Fatalf("canonicalizeReparse(%q) = %q, want %q", linkPath, got, want)
	}
}

// TestCanonicalizeReparseOutOfRootJunction verifies a live directory
// junction whose target lies outside the workspace resolves to the true,
// outside-root target -- the core defect this helper closes (014.003-T
// AC6, case 3).
func TestCanonicalizeReparseOutOfRootJunction(t *testing.T) {
	outsideRoot := t.TempDir()

	rootDir := t.TempDir()
	root, err := NewRoot(rootDir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", rootDir, err)
	}

	linkPath := filepath.Join(root.Path(), "junction-out")
	createDirectoryJunction(t, linkPath, outsideRoot)

	got, err := canonicalizeReparse(linkPath)
	if err != nil {
		t.Fatalf("canonicalizeReparse(%q) returned unexpected error: %v", linkPath, err)
	}
	got = stripUNCPrefix(got)
	if hasPathPrefix(got, root.Path()) {
		t.Fatalf("canonicalizeReparse(%q) = %q, want a path outside root %q", linkPath, got, root.Path())
	}
	want := canonicalForComparison(t, outsideRoot)
	if !strings.EqualFold(got, want) {
		t.Fatalf("canonicalizeReparse(%q) = %q, want %q", linkPath, got, want)
	}
}

// TestCanonicalizeReparseNonExistentPathErrors verifies a non-existent path
// returns an error rather than a zero-value success (014.003-T AC6, case
// 4).
func TestCanonicalizeReparseNonExistentPathErrors(t *testing.T) {
	rootDir := t.TempDir()
	missing := filepath.Join(rootDir, "does-not-exist")

	_, err := canonicalizeReparse(missing)
	if err == nil {
		t.Fatalf("canonicalizeReparse(%q) = nil error, want an error for a non-existent path", missing)
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("canonicalizeReparse(%q) error = %v, want it to satisfy errors.Is(err, fs.ErrNotExist)", missing, err)
	}
}

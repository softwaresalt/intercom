package pathsafe

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestNewRootOnExistingDirIsAbsoluteAndSymlinkResolved verifies NewRoot on an
// existing directory returns an absolute, symlink-resolved path with no
// \\?\ UNC prefix.
func TestNewRootOnExistingDirIsAbsoluteAndSymlinkResolved(t *testing.T) {
	dir := t.TempDir()

	r, err := NewRoot(dir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", dir, err)
	}
	if !filepath.IsAbs(r.Path()) {
		t.Fatalf("Root.Path() = %q, want absolute path", r.Path())
	}
	if strings.HasPrefix(r.Path(), `\\?\`) {
		t.Fatalf("Root.Path() = %q, want no \\\\?\\ UNC prefix", r.Path())
	}
}

// TestNewRootOnMissingDirReturnsPathViolation verifies NewRoot on a
// non-existent directory returns a PathViolation whose message begins
// "path violation: workspace root invalid".
func TestNewRootOnMissingDirReturnsPathViolation(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "does-not-exist")

	_, err := NewRoot(dir)
	if err == nil {
		t.Fatalf("NewRoot(%q) = nil error, want PathViolation", dir)
	}
	if !strings.HasPrefix(err.Error(), "path violation: workspace root invalid") {
		t.Fatalf("err.Error() = %q, want prefix %q", err.Error(), "path violation: workspace root invalid")
	}
}

// TestNewRootIsIdempotent verifies NewRoot(r.Path()) yields the same path.
func TestNewRootIsIdempotent(t *testing.T) {
	dir := t.TempDir()

	r1, err := NewRoot(dir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", dir, err)
	}

	r2, err := NewRoot(r1.Path())
	if err != nil {
		t.Fatalf("NewRoot(%q) (idempotent call) returned error: %v", r1.Path(), err)
	}

	if r1.Path() != r2.Path() {
		t.Fatalf("NewRoot is not idempotent: %q != %q", r1.Path(), r2.Path())
	}
}

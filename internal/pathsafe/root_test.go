package pathsafe

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/softwaresalt/intercom-go/internal/apperr"
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

// TestNewRootPreservesEvalSymlinksNotExistCause verifies a missing workspace
// root preserves both the path-violation classification and the underlying
// os.ErrNotExist cause from filepath.EvalSymlinks.
func TestNewRootPreservesEvalSymlinksNotExistCause(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "does-not-exist")

	_, err := NewRoot(dir)
	if err == nil {
		t.Fatalf("NewRoot(%q) = nil error, want wrapped PathViolation", dir)
	}
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("errors.Is(err, os.ErrNotExist) = false, want true; err = %v", err)
	}
	if !errors.Is(err, apperr.ErrPathViolation) {
		t.Fatalf("errors.Is(err, apperr.ErrPathViolation) = false, want true; err = %v", err)
	}
	if !strings.Contains(err.Error(), "workspace root invalid") {
		t.Fatalf("err.Error() = %q, want to contain %q", err.Error(), "workspace root invalid")
	}
}

// TestNewRootRejectsFilePathAsWorkspaceRoot verifies NewRoot rejects an
// existing file path rather than accepting it as a valid workspace root.
func TestNewRootRejectsFilePathAsWorkspaceRoot(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "not-a-directory.txt")
	if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
		t.Fatalf("failed to create file root candidate: %v", err)
	}

	root, err := NewRoot(filePath)
	if err == nil {
		t.Fatalf("NewRoot(%q) = (%v, nil error), want PathViolation", filePath, root)
	}
	if !errors.Is(err, apperr.ErrPathViolation) {
		t.Fatalf("errors.Is(err, apperr.ErrPathViolation) = false, want true; err = %v", err)
	}
}

// TestNewRootAcceptsSymlinkToDirectory verifies NewRoot resolves symlinks
// before checking the canonical target type, so a symlink to a directory is
// accepted as a valid workspace root.
func TestNewRootAcceptsSymlinkToDirectory(t *testing.T) {
	requireSymlinkOrFailClosed(t)

	dir := t.TempDir()
	targetDir := filepath.Join(dir, "target")
	if err := os.Mkdir(targetDir, 0o755); err != nil {
		t.Fatalf("failed to create target directory: %v", err)
	}
	linkDir := filepath.Join(dir, "link")
	if err := os.Symlink(targetDir, linkDir); err != nil {
		t.Fatalf("failed to create directory symlink: %v", err)
	}

	root, err := NewRoot(linkDir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", linkDir, err)
	}

	resolved, err := filepath.EvalSymlinks(linkDir)
	if err != nil {
		t.Fatalf("filepath.EvalSymlinks(%q) returned error: %v", linkDir, err)
	}
	want := stripUNCPrefix(resolved)
	if root.Path() != want {
		t.Fatalf("Root.Path() = %q, want %q", root.Path(), want)
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

// TestStripUNCPrefixWindowsTransformVerdicts verifies stripUNCPrefix's pure
// string transforms for Windows extended-path inputs without relying on
// host-dependent filepath helpers.
func TestStripUNCPrefixWindowsTransformVerdicts(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{name: "unc share root", in: `\\?\UNC\server\share`, want: `\\server\share`},
		{name: "unc share subpath", in: `\\?\UNC\server\share\ws\file`, want: `\\server\share\ws\file`},
		{name: "drive path", in: `\\?\C:\path`, want: `C:\path`},
		{name: "lowercase unc prefix", in: `\\?\unc\server\share`, want: `\\server\share`},
		{name: "unc near miss", in: `\\?\UNCfoo`, want: `\\?\UNCfoo`},
		{name: "bare unc marker", in: `\\?\UNC`, want: `\\?\UNC`},
		{name: "volume guid", in: `\\?\Volume{GUID}\path`, want: `\\?\Volume{GUID}\path`},
		{name: "globalroot", in: `\\?\GLOBALROOT\Device\HarddiskVolumeShadowCopy1`, want: `\\?\GLOBALROOT\Device\HarddiskVolumeShadowCopy1`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := stripUNCPrefix(tc.in); got != tc.want {
				t.Fatalf("stripUNCPrefix(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

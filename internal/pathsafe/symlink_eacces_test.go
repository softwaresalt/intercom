//go:build !windows

package pathsafe

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

// TestResolveRejectsChildBeneathUnsearchableDirectory closes finding F11
// (014.007-T AC2): a dedicated test for the "else if
// !errors.Is(err, fs.ErrNotExist)" probe-error branch in checkSymlinkEscape
// triggered by EACCES rather than ENOTDIR (contain_posix_test.go covers the
// ENOTDIR trigger for the same branch under F4) or a dangling target
// (014.006-T's canonicalizeReparse-failure branch). Removing the execute
// (search) permission bit from an ancestor directory makes the kernel
// refuse to traverse into it while resolving a longer path, which is
// reported as EACCES -- a distinct OS cause from both ENOTDIR and ENOENT,
// and one that a superuser is exempt from (root bypasses the directory
// permission check entirely), so this test skips itself under root
// (matching the existing os.Geteuid()-based skip pattern used for
// privilege-dependent POSIX tests in this package).
func TestResolveRejectsChildBeneathUnsearchableDirectory(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("skipping: running as root bypasses directory search-permission checks (EACCES is unreachable)")
	}

	dir := t.TempDir()
	root, err := NewRoot(dir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", dir, err)
	}

	blockedDir := filepath.Join(root.Path(), "blocked")
	if err := os.Mkdir(blockedDir, 0o755); err != nil {
		t.Fatalf("failed to create ancestor directory: %v", err)
	}
	// Remove the execute (search) bit so the kernel refuses to traverse
	// into blockedDir while resolving a path beneath it. Restore
	// permissions during cleanup so t.TempDir() can remove the tree.
	if err := os.Chmod(blockedDir, 0o644); err != nil {
		t.Fatalf("failed to remove search permission on ancestor directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(blockedDir, 0o755)
	})

	candidate := filepath.Join("blocked", "child")
	_, err = root.Resolve(candidate)
	if err == nil {
		t.Fatalf("Resolve(%q) = nil error, want a KindPathViolation rejection", candidate)
	}
	if !strings.Contains(err.Error(), symlinkUnverifiableMsg) {
		t.Fatalf("Resolve(%q) error = %q, want to contain %q", candidate, err.Error(), symlinkUnverifiableMsg)
	}
	if !errors.Is(err, syscall.EACCES) {
		t.Fatalf("Resolve(%q) error = %v, want errors.Is(err, syscall.EACCES) -- the OS cause must be preserved "+
			"through apperr.Wrapf for errors.Is/errors.As inspection (014.006-T AC3)", candidate, err)
	}
}

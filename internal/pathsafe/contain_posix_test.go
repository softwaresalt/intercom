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

// TestResolveRejectsChildBeneathRegularFilePOSIXTakesProbeErrorBranch closes
// finding F4 (014.007-T AC1). TestResolveRejectsChildBeneathRegularFile
// (contain_test.go) is named for the ancestor-guard branch in
// checkSymlinkEscape (the "!finalProbe && !info.IsDir() && !ModeSymlink"
// check), but on POSIX it never actually reaches that branch: the kernel's
// path-resolution algorithm reports ENOTDIR -- not ENOENT -- the instant it
// must treat an existing non-directory component (file.txt) as a directory
// in order to keep resolving the remainder of the path, regardless of how
// many more components follow or whether they exist. Since ENOTDIR does not
// satisfy errors.Is(err, fs.ErrNotExist) (syscall.Errno.Is only matches
// ENOENT for that sentinel; see syscall/syscall_unix.go), the walk in
// checkSymlinkEscape never climbs far enough to Lstat "file.txt" in
// isolation and observe it as a non-directory ancestor -- it takes the
// "else if !errors.Is(err, fs.ErrNotExist)" probe-error branch on its very
// first iteration instead. This was verified empirically (a Python
// os.lstat repro against a real Linux kernel, and confirmed against
// syscall_unix.go's Errno.Is) before writing this test, rather than
// asserted from the plan's Windows-centric assumption alone.
//
// This test pins down that real, platform-specific routing explicitly: it
// asserts the observed OS cause is ENOTDIR (via errors.Is), documenting
// that the ancestor-guard branch is unreachable for this scenario on
// POSIX by construction -- not a coverage gap to chase further. The
// observable outcome (rejection, symlinkUnverifiableMsg,
// apperr.KindPathViolation) is identical to the Windows-routed ancestor-
// guard branch; only the internal path and the wrapped OS cause differ.
func TestResolveRejectsChildBeneathRegularFilePOSIXTakesProbeErrorBranch(t *testing.T) {
	dir := t.TempDir()
	root, err := NewRoot(dir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", dir, err)
	}

	filePath := filepath.Join(root.Path(), "file.txt")
	if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
		t.Fatalf("failed to create regular file ancestor: %v", err)
	}

	candidate := filepath.Join("file.txt", "sub")
	_, err = root.Resolve(candidate)
	if err == nil {
		t.Fatalf("Resolve(%q) = nil error, want a KindPathViolation rejection", candidate)
	}
	if !strings.Contains(err.Error(), symlinkUnverifiableMsg) {
		t.Fatalf("Resolve(%q) error = %q, want to contain %q", candidate, err.Error(), symlinkUnverifiableMsg)
	}
	if !errors.Is(err, syscall.ENOTDIR) {
		t.Fatalf("Resolve(%q) error = %v, want errors.Is(err, syscall.ENOTDIR) -- the probe-error branch must "+
			"preserve the OS cause so this platform routing stays provable rather than assumed", candidate, err)
	}
}

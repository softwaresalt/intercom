//go:build windows

package pathsafe

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

// TestResolveDanglingReparsePointErrorSurfaceIsPathError is the 017.001-T
// RED/GREEN lock: checkSymlinkEscape's canonicalizeReparse error branch
// (pathsafe.go) must expose an *fs.PathError via errors.As, exactly like
// NewRoot's own failure branch already does (root_windows_test.go's
// TestNewRootErrorSurfaceIsPathError, C2/AC3 of 016.007-T). Before this
// task, checkSymlinkEscape returned canonicalizeReparse's raw
// syscall.Errno wrapped only by apperr.Wrapf -- no *fs.PathError re-forming
// -- so a caller doing errors.As(err, &pathErr) matched on POSIX (where
// canonicalizeReparse already returns *fs.PathError via
// filepath.EvalSymlinks) but silently stopped matching on Windows. Routing
// the same call site through wrapPathError (root.go, no build tag) closes
// that parity gap.
//
// Uses a DANGLING JUNCTION, not a dangling symlink, deliberately
// (privilege-independent fixture per operator directive): junction
// creation does not require SeCreateSymbolicLinkPrivilege, so this test
// runs on unprivileged CI where a symlink-based equivalent would only
// exercise the enumerated skip path (requireSymlinkOrFailClosed,
// symlink_test.go) rather than the real assertion. This exercises the
// identical checkSymlinkEscape code path as
// TestResolveRejectsDanglingJunctionAsExactFinalComponent
// (junction_windows_test.go) -- final path component, Lstat succeeds,
// canonicalizeReparse fails -- adding only the *fs.PathError surface
// assertion that test does not make.
func TestResolveDanglingReparsePointErrorSurfaceIsPathError(t *testing.T) {
	outsideRoot := t.TempDir()
	targetPath := filepath.Join(outsideRoot, "target")
	if err := os.Mkdir(targetPath, 0o755); err != nil {
		t.Fatalf("failed to create junction target: %v", err)
	}

	rootDir := t.TempDir()
	root, err := NewRoot(rootDir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", rootDir, err)
	}

	linkPath := filepath.Join(root.Path(), "junction-dangling")
	createDirectoryJunction(t, linkPath, targetPath)

	if err := os.Remove(targetPath); err != nil {
		t.Fatalf("failed to remove junction target and leave a dangling junction: %v", err)
	}

	_, err = root.Resolve("junction-dangling")
	if err == nil {
		t.Fatalf(`Resolve("junction-dangling") = nil error, want %q`, symlinkUnverifiableMsg)
	}

	var pathErr *fs.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("errors.As(err, &pathErr) = false, want true (wrapPathError parity with NewRoot); err = %v", err)
	}
	if pathErr.Op != "open" {
		t.Fatalf("pathErr.Op = %q, want %q (matching NewRoot's identical wrapPathError call site, root.go)", pathErr.Op, "open")
	}
}

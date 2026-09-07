package pathsafe

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// canSymlink probes for symlink-creation privilege at runtime (not
// runtime.GOOS), so elevated Windows environments are still covered
// (finding GO-12). It returns the probe error so callers can t.Skip with an
// explicit reason rather than passing vacuously.
func canSymlink(t *testing.T) error {
	t.Helper()
	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatalf("failed to create symlink probe target: %v", err)
	}
	link := filepath.Join(dir, "link")
	return os.Symlink(target, link)
}

// requireSymlinkOrFailClosed is the 011.010-T durable assertion: on Windows,
// a symlink-dependent test's unmet precondition must FAIL rather than skip,
// except for the single enumerated privilege case (ERROR_PRIVILEGE_NOT_HELD
// -- junction/symlink creation requiring elevation or Developer Mode), which
// still skips with an explicit reason. This replaces a one-time log
// inspection with a mechanically enforced assertion: any other Windows
// failure reason (a genuinely broken environment, not merely an unprivileged
// one) now fails the test instead of silently skipping it. Non-Windows
// platforms are unaffected by this tightening and keep the original
// unconditional skip.
func requireSymlinkOrFailClosed(t *testing.T) {
	t.Helper()
	err := canSymlink(t)
	if err == nil {
		return
	}
	if runtime.GOOS != "windows" {
		t.Skipf("skipping: no symlink privilege in this environment: %v", err)
		return
	}
	if isSymlinkPrivilegeError(err) {
		t.Skipf("skipping (enumerated privilege case, 011.010-T): ERROR_PRIVILEGE_NOT_HELD -- SeCreateSymbolicLinkPrivilege not held; enable Developer Mode or run elevated: %v", err)
		return
	}
	t.Fatalf("symlink precondition failed on windows for a reason OTHER than the enumerated privilege case (011.010-T durable assertion demands failure, not a silent skip): %v", err)
}

// TestResolveRejectsSymlinkEscapingRoot verifies a symlink inside the root
// whose target is outside the root is rejected with
// "symlink target escapes workspace".
func TestResolveRejectsSymlinkEscapingRoot(t *testing.T) {
	requireSymlinkOrFailClosed(t)

	outsideDir := t.TempDir()
	outsideTarget := filepath.Join(outsideDir, "secret.txt")
	if err := os.WriteFile(outsideTarget, []byte("secret"), 0o644); err != nil {
		t.Fatalf("failed to create outside target: %v", err)
	}

	rootDir := t.TempDir()
	root, err := NewRoot(rootDir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", rootDir, err)
	}

	linkPath := filepath.Join(root.Path(), "escape-link")
	if err := os.Symlink(outsideTarget, linkPath); err != nil {
		t.Fatalf("failed to create escaping symlink: %v", err)
	}

	_, err = root.Resolve("escape-link")
	if err == nil {
		t.Fatalf("Resolve(%q) = nil error, want %q", "escape-link", symlinkEscapeMsg)
	}
	if !strings.Contains(err.Error(), symlinkEscapeMsg) {
		t.Fatalf("Resolve(%q) error = %q, want to contain %q", "escape-link", err.Error(), symlinkEscapeMsg)
	}
}

// TestResolveAllowsSymlinkInsideRoot verifies a symlink inside the root
// whose target is also inside the root resolves successfully.
func TestResolveAllowsSymlinkInsideRoot(t *testing.T) {
	requireSymlinkOrFailClosed(t)

	rootDir := t.TempDir()
	root, err := NewRoot(rootDir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", rootDir, err)
	}

	insideTarget := filepath.Join(root.Path(), "real.txt")
	if err := os.WriteFile(insideTarget, []byte("real"), 0o644); err != nil {
		t.Fatalf("failed to create inside target: %v", err)
	}

	linkPath := filepath.Join(root.Path(), "inside-link")
	if err := os.Symlink(insideTarget, linkPath); err != nil {
		t.Fatalf("failed to create in-root symlink: %v", err)
	}

	if _, err := root.Resolve("inside-link"); err != nil {
		t.Fatalf("Resolve(%q) returned unexpected error: %v", "inside-link", err)
	}
}

// TestResolveRejectsSymlinkedIntermediateDirEscapingRoot verifies the
// checkSymlinkEscape ancestor-walk fix: a symlinked *intermediate* directory
// pointing outside the root, combined with a non-existent leaf component (the
// dominant new-file-write shape), is still rejected with
// "symlink target escapes workspace" rather than silently falling into the
// non-existent-leaf accept branch.
func TestResolveRejectsSymlinkedIntermediateDirEscapingRoot(t *testing.T) {
	requireSymlinkOrFailClosed(t)

	outsideDir := t.TempDir()

	rootDir := t.TempDir()
	root, err := NewRoot(rootDir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", rootDir, err)
	}

	linkDir := filepath.Join(root.Path(), "link-dir")
	if err := os.Symlink(outsideDir, linkDir); err != nil {
		t.Fatalf("failed to create escaping directory symlink: %v", err)
	}

	// The leaf file does not exist yet under the symlinked directory — this
	// is the shape that a naive os.Stat(resolved)-only check would miss.
	_, err = root.Resolve("link-dir/new-file.txt")
	if err == nil {
		t.Fatalf("Resolve(%q) = nil error, want %q", "link-dir/new-file.txt", symlinkEscapeMsg)
	}
	if !strings.Contains(err.Error(), symlinkEscapeMsg) {
		t.Fatalf("Resolve(%q) error = %q, want to contain %q", "link-dir/new-file.txt", err.Error(), symlinkEscapeMsg)
	}
}

// TestResolveAllowsDanglingSymlinkAtFinalComponent verifies the documented
// lexical-only accept branch for a dangling symlink at the final path
// component.
func TestResolveAllowsDanglingSymlinkAtFinalComponent(t *testing.T) {
	requireSymlinkOrFailClosed(t)

	rootDir := t.TempDir()
	root, err := NewRoot(rootDir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", rootDir, err)
	}

	danglingTarget := filepath.Join(t.TempDir(), "missing-target.txt")
	linkPath := filepath.Join(root.Path(), "dangling-link")
	if err := os.Symlink(danglingTarget, linkPath); err != nil {
		t.Fatalf("failed to create dangling symlink: %v", err)
	}

	resolved, err := root.Resolve("dangling-link")
	if err != nil {
		t.Fatalf("Resolve(%q) returned unexpected error: %v", "dangling-link", err)
	}
	if resolved != linkPath {
		t.Fatalf("Resolve(%q) = %q, want %q", "dangling-link", resolved, linkPath)
	}
}

// path under the root resolves successfully (oracle step 7).
func TestResolveAllowsNonExistentRelativePath(t *testing.T) {
	rootDir := t.TempDir()
	root, err := NewRoot(rootDir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", rootDir, err)
	}

	resolved, err := root.Resolve("does/not/exist.txt")
	if err != nil {
		t.Fatalf("Resolve(%q) returned unexpected error: %v", "does/not/exist.txt", err)
	}
	if !hasPathPrefix(resolved, root.Path()) {
		t.Fatalf("Resolve(%q) = %q, want path inside root %q", "does/not/exist.txt", resolved, root.Path())
	}
}

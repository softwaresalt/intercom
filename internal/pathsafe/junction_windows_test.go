package pathsafe

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const junctionCreateTimeout = time.Minute

func createDirectoryJunction(t *testing.T, linkPath, targetPath string) {
	t.Helper()

	if _, err := exec.LookPath("pwsh"); err != nil {
		// Fail loud, not skip (adversarial-review finding, 013-S): this
		// helper backs the C1/C2/C3 regression locks that ARE this
		// shipment's core verified deliverable (700B41CE closure, plan
		// Sec.9 "GREEN proof"). A silent t.Skip on missing pwsh would let
		// a future environment lacking pwsh report these tests as PASS
		// without ever exercising them -- exactly the kind of masked,
		// falsely-green evidence plan Sec.7 already warns against for the
		// CI gate. GitHub's windows-latest runners and this shipment's
		// own verified local dev machine both have pwsh on PATH by
		// default; its absence here signals a genuine environment
		// misconfiguration, not a legitimately-inapplicable test (unlike,
		// e.g., symlink_eacces_test.go's documented root-privilege skip,
		// where the scenario truly cannot occur under root).
		t.Fatalf("pwsh not available on PATH: %v -- required to create a real directory junction for this test; install PowerShell 7+ or run on a Windows image that provides it (GitHub's windows-latest runner does)", err)
	}

	helperScript := filepath.Join(t.TempDir(), "create-junction.ps1")
	helperContent := strings.Join([]string{
		"param(",
		"    [Parameter(Mandatory)][string]$LinkPath,",
		"    [Parameter(Mandatory)][string]$TargetPath",
		")",
		"$ErrorActionPreference = 'Stop'",
		"New-Item -ItemType Junction -Path $LinkPath -Target $TargetPath | Out-Null",
		"",
	}, "\n")
	if err := os.WriteFile(helperScript, []byte(helperContent), 0o600); err != nil {
		t.Fatalf("writing junction helper script: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), junctionCreateTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pwsh", "-NoProfile", "-File", helperScript, "-LinkPath", linkPath, "-TargetPath", targetPath)
	out, err := cmd.CombinedOutput()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("junction creation exceeded %s timeout; output so far:\n%s", junctionCreateTimeout, out)
	}
	if err != nil {
		t.Fatalf("creating junction %q -> %q: %v\noutput:\n%s", linkPath, targetPath, err, out)
	}
	t.Cleanup(func() {
		if err := os.Remove(linkPath); err != nil && !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("removing junction %q: %v", linkPath, err)
		}
	})
}

// TestResolveRejectsDanglingIntermediateJunctionOutsideRoot verifies an
// intermediate directory junction that ends up dangling is rejected with
// "symlink target escapes workspace".
func TestResolveRejectsDanglingIntermediateJunctionOutsideRoot(t *testing.T) {
	outsideRoot := t.TempDir()
	targetPath := filepath.Join(outsideRoot, "target")
	if err := os.Mkdir(targetPath, 0o755); err != nil {
		t.Fatalf("failed to create outside junction target: %v", err)
	}

	rootDir := t.TempDir()
	root, err := NewRoot(rootDir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", rootDir, err)
	}

	linkPath := filepath.Join(root.Path(), "junction-out")
	createDirectoryJunction(t, linkPath, targetPath)

	if err := os.Remove(targetPath); err != nil {
		t.Fatalf("failed to remove junction target and leave a dangling junction: %v", err)
	}
	if _, err := os.Lstat(linkPath); err != nil {
		t.Fatalf("os.Lstat(%q) returned error after target removal, want the dangling junction entry to remain: %v", linkPath, err)
	}
	if _, err := os.Stat(linkPath); err == nil {
		t.Fatalf("os.Stat(%q) succeeded after target removal, want a dangling junction", linkPath)
	}

	// A dangling junction's target cannot be resolved at all, so
	// containment is unverifiable rather than a proven escape (014.006-T).
	candidate := filepath.Join("junction-out", "new-file.txt")
	_, err = root.Resolve(candidate)
	if err == nil {
		t.Fatalf("Resolve(%q) = nil error, want %q", candidate, symlinkUnverifiableMsg)
	}
	if !strings.Contains(err.Error(), symlinkUnverifiableMsg) {
		t.Fatalf("Resolve(%q) error = %q, want to contain %q", candidate, err.Error(), symlinkUnverifiableMsg)
	}
}

// TestResolveRejectsDanglingJunctionAsExactFinalComponent covers a gap
// identified by 013-S's adversarial review (finding U4): the existing
// dangling-junction test above uses the junction as an INTERMEDIATE
// component (a child beneath it), and the existing dangling-symlink final-
// component test (TestResolveRejectsDanglingSymlinkAtFinalComponent,
// symlink_test.go) covers a SYMLINK, not a JUNCTION. root.go's GO-14 entry
// claims the dangling-target rejection now applies "at ANY position,
// including the terminal component" for reparse points generally -- this
// test locks in that claim specifically for a directory junction (not a
// symlink) used as the exact literal final Resolve() argument, mirroring
// TestResolveRejectsLiveJunctionAsFinalComponent's "deliberately uses the
// junction itself, not a subpath" pattern above.
func TestResolveRejectsDanglingJunctionAsExactFinalComponent(t *testing.T) {
	outsideRoot := t.TempDir()
	targetPath := filepath.Join(outsideRoot, "target")
	if err := os.Mkdir(targetPath, 0o755); err != nil {
		t.Fatalf("failed to create outside junction target: %v", err)
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
	if _, err := os.Lstat(linkPath); err != nil {
		t.Fatalf("os.Lstat(%q) returned error after target removal, want the dangling junction entry to remain: %v", linkPath, err)
	}

	_, err = root.Resolve("junction-dangling")
	if err == nil {
		t.Fatalf("Resolve(%q) = nil error, want %q", "junction-dangling", symlinkUnverifiableMsg)
	}
	if !strings.Contains(err.Error(), symlinkUnverifiableMsg) {
		t.Fatalf("Resolve(%q) error = %q, want to contain %q", "junction-dangling", err.Error(), symlinkUnverifiableMsg)
	}
}

// TestResolveRejectsLiveJunctionAsFinalComponent verifies case C1 (014.001-T):
// a LIVE (non-dangling) directory junction whose real target lies outside
// the workspace root is rejected when the junction itself is the literal
// final path component. Deliberately uses Resolve("junction-live"), NOT
// Resolve("junction-live/child") -- a subpath exercises the already-fixed
// ancestor branch and would false-negative (finding A1/R3).
func TestResolveRejectsLiveJunctionAsFinalComponent(t *testing.T) {
	outsideRoot := t.TempDir()

	rootDir := t.TempDir()
	root, err := NewRoot(rootDir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", rootDir, err)
	}

	linkPath := filepath.Join(root.Path(), "junction-live")
	createDirectoryJunction(t, linkPath, outsideRoot)

	_, err = root.Resolve("junction-live")
	if err == nil {
		t.Fatalf("Resolve(%q) = nil error, want %q", "junction-live", symlinkEscapeMsg)
	}
	if !strings.Contains(err.Error(), symlinkEscapeMsg) {
		t.Fatalf("Resolve(%q) error = %q, want to contain %q", "junction-live", err.Error(), symlinkEscapeMsg)
	}
}

// TestResolveAllowsLiveJunctionAsIntermediateWithExistingLeafInRoot locks in
// a claim made in root.go's 700B41CE risk-register entry (corrected during
// 013-S adversarial review, finding "in-root C2/C3 acceptance is untested
// at the Resolve() level"): when a live directory junction used as an
// INTERMEDIATE path component has a real target INSIDE the workspace root,
// and the leaf beyond it EXISTS, Resolve() must ACCEPT it -- the mirror
// image of TestResolveRejectsLiveJunctionAsIntermediateWithExistingLeaf
// above (same code shape, opposite direction). Before this test, only
// canonicalizeReparse itself (TestCanonicalizeReparseInRootJunction,
// reparse_windows_test.go) was verified to resolve this shape correctly in
// isolation; nothing exercised the full Resolve() path end-to-end to prove
// the accept side of the containment check, not just canonicalizeReparse's
// own internal resolution.
func TestResolveAllowsLiveJunctionAsIntermediateWithExistingLeafInRoot(t *testing.T) {
	rootDir := t.TempDir()
	root, err := NewRoot(rootDir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", rootDir, err)
	}

	targetDir := filepath.Join(root.Path(), "real-target")
	if err := os.Mkdir(targetDir, 0o755); err != nil {
		t.Fatalf("failed to create in-root junction target: %v", err)
	}
	existingLeaf := filepath.Join(targetDir, "existing.txt")
	if err := os.WriteFile(existingLeaf, []byte("x"), 0o644); err != nil {
		t.Fatalf("failed to create existing leaf inside root target: %v", err)
	}

	linkPath := filepath.Join(root.Path(), "junction-in-root")
	createDirectoryJunction(t, linkPath, targetDir)

	candidate := filepath.Join("junction-in-root", "existing.txt")
	got, err := root.Resolve(candidate)
	if err != nil {
		t.Fatalf("Resolve(%q) returned unexpected error: %v", candidate, err)
	}
	if !hasPathPrefix(got, root.Path()) {
		t.Fatalf("Resolve(%q) = %q, want a path inside root %q", candidate, got, root.Path())
	}
}

// TestResolveRejectsLiveJunctionAsIntermediateWithExistingLeaf verifies case
// C2 (014.001-T): a live directory junction used as an INTERMEDIATE path
// component, with an EXISTING leaf beyond it, is rejected. os.Lstat/os.Stat
// on the full candidate transparently traverses the junction and succeeds on
// the first ancestor-walk iteration, so the reparse branch never fires
// without whole-prefix canonicalization (finding A1).
func TestResolveRejectsLiveJunctionAsIntermediateWithExistingLeaf(t *testing.T) {
	outsideRoot := t.TempDir()
	existingLeaf := filepath.Join(outsideRoot, "existing.txt")
	if err := os.WriteFile(existingLeaf, []byte("x"), 0o644); err != nil {
		t.Fatalf("failed to create existing leaf outside root: %v", err)
	}

	rootDir := t.TempDir()
	root, err := NewRoot(rootDir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", rootDir, err)
	}

	linkPath := filepath.Join(root.Path(), "junction-live")
	createDirectoryJunction(t, linkPath, outsideRoot)

	candidate := filepath.Join("junction-live", "existing.txt")
	_, err = root.Resolve(candidate)
	if err == nil {
		t.Fatalf("Resolve(%q) = nil error, want %q", candidate, symlinkEscapeMsg)
	}
	if !strings.Contains(err.Error(), symlinkEscapeMsg) {
		t.Fatalf("Resolve(%q) error = %q, want to contain %q", candidate, err.Error(), symlinkEscapeMsg)
	}
}

// TestResolveRejectsLiveJunctionAsIntermediateWithNonExistentLeaf covers a
// gap identified by 013-S's adversarial review (finding U2): the dominant
// real-world use case for this package is creating a NEW file (a
// non-existent leaf), and no prior test combined that shape with a LIVE,
// out-of-root junction intermediate -- only the existing-leaf variant
// (TestResolveRejectsLiveJunctionAsIntermediateWithExistingLeaf, above) was
// covered. Unlike that test, this one is expected to reject via the
// strict-ancestor-block branch (symlinkUnverifiableMsg, F3's blanket
// "any live junction ancestor is rejected" guard in checkSymlinkEscape),
// NOT via canonicalizeReparse's verified-escape path (symlinkEscapeMsg):
// with a non-existent leaf, the ancestor-climb loop never reaches
// canonicalizeReparse at all -- it climbs straight to the junction itself
// as the first EXISTING entry found, and F3's blanket ancestor-block
// rejects it there before any target-resolution occurs. This test locks
// in that this out-of-root case stays rejected for the right reason
// (F3's guard incidentally also covers this correct-rejection shape, not
// just the false-rejection shape F3 itself names) as a landmine warning
// for any future shipment that relaxes F3: relaxing the strict-ancestor
// guard to permit an in-root junction ancestor must not accidentally
// also permit this out-of-root shape to slip through unverified.
func TestResolveRejectsLiveJunctionAsIntermediateWithNonExistentLeaf(t *testing.T) {
	outsideRoot := t.TempDir()

	rootDir := t.TempDir()
	root, err := NewRoot(rootDir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", rootDir, err)
	}

	linkPath := filepath.Join(root.Path(), "junction-live")
	createDirectoryJunction(t, linkPath, outsideRoot)

	candidate := filepath.Join("junction-live", "nonexistent.txt")
	_, err = root.Resolve(candidate)
	if err == nil {
		t.Fatalf("Resolve(%q) = nil error, want %q", candidate, symlinkUnverifiableMsg)
	}
	if !strings.Contains(err.Error(), symlinkUnverifiableMsg) {
		t.Fatalf("Resolve(%q) error = %q, want to contain %q", candidate, err.Error(), symlinkUnverifiableMsg)
	}
}

// TestResolveRejectsLiveJunctionAncestorWithExistingDirBeyond verifies case
// C3 (014.001-T): a live directory junction whose real target contains an
// EXISTING directory, with a non-existent leaf below that, is rejected. The
// ancestor walk terminates on the existing directory BEYOND the junction --
// a position filepath.EvalSymlinks (which does not resolve mount points)
// cannot see as escaping, because it never redirects through the junction
// in the first place.
func TestResolveRejectsLiveJunctionAncestorWithExistingDirBeyond(t *testing.T) {
	outsideRoot := t.TempDir()
	existingDir := filepath.Join(outsideRoot, "existing-dir")
	if err := os.Mkdir(existingDir, 0o755); err != nil {
		t.Fatalf("failed to create existing dir outside root: %v", err)
	}

	rootDir := t.TempDir()
	root, err := NewRoot(rootDir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", rootDir, err)
	}

	linkPath := filepath.Join(root.Path(), "junction-live")
	createDirectoryJunction(t, linkPath, outsideRoot)

	candidate := filepath.Join("junction-live", "existing-dir", "missing.txt")
	_, err = root.Resolve(candidate)
	if err == nil {
		t.Fatalf("Resolve(%q) = nil error, want %q", candidate, symlinkEscapeMsg)
	}
	if !strings.Contains(err.Error(), symlinkEscapeMsg) {
		t.Fatalf("Resolve(%q) error = %q, want to contain %q", candidate, err.Error(), symlinkEscapeMsg)
	}
}

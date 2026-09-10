//go:build windows

package pathsafe

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestStripUNCPrefixReturnsWindowsAbsolutePaths verifies the stripped result
// remains Windows-absolute for the cases whose postcondition requires it.
func TestStripUNCPrefixReturnsWindowsAbsolutePaths(t *testing.T) {
	cases := []struct {
		name       string
		in         string
		want       string
		wantVolume string
	}{
		{name: "unc share root", in: `\\?\UNC\server\share`, want: `\\server\share`, wantVolume: `\\server\share`},
		{name: "unc share subpath", in: `\\?\UNC\server\share\ws\file`, want: `\\server\share\ws\file`, wantVolume: `\\server\share`},
		{name: "drive path", in: `\\?\C:\path`, want: `C:\path`, wantVolume: `C:`},
		{name: "lowercase unc prefix", in: `\\?\unc\server\share`, want: `\\server\share`, wantVolume: `\\server\share`},
		{name: "volume guid", in: `\\?\Volume{GUID}\path`, want: `\\?\Volume{GUID}\path`, wantVolume: `\\?\Volume{GUID}`},
		{name: "globalroot", in: `\\?\GLOBALROOT\Device\HarddiskVolumeShadowCopy1`, want: `\\?\GLOBALROOT\Device\HarddiskVolumeShadowCopy1`, wantVolume: `\\?\GLOBALROOT`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := stripUNCPrefix(tc.in)
			if got != tc.want {
				t.Fatalf("stripUNCPrefix(%q) = %q, want %q", tc.in, got, tc.want)
			}
			if !filepath.IsAbs(got) {
				t.Fatalf("filepath.IsAbs(stripUNCPrefix(%q)) = false, want true; got %q", tc.in, got)
			}
			if volume := filepath.VolumeName(got); volume != tc.wantVolume {
				t.Fatalf("filepath.VolumeName(stripUNCPrefix(%q)) = %q, want %q", tc.in, volume, tc.wantVolume)
			}
		})
	}
}

// TestNewRootOnJunctionRootedWorkspaceResolvesExistingDescendant is the C1
// RED regression lock (016.005-T, availability direction, transition (a)):
// when the workspace root ITSELF is a live directory junction (not a
// subpath through one -- the exact reported shape per the 011-S compound
// learning, since a structurally different repro would exercise an
// already-fixed branch and false-negative), an EXISTING descendant must
// still resolve successfully.
//
// RED against current main: NewRoot canonicalizes the root via
// filepath.EvalSymlinks, which does not resolve a directory junction (Go
// reports it as os.ModeIrregular, not a symlink), so root.path retains the
// junction's own unresolved lexical path. Resolve's existing-descendant
// path reaches canonicalizeReparse, which DOES follow the junction via
// CreateFile and returns the real (different) target path -- so
// hasPathPrefix(real, root.path) compares two differently-canonicalized
// strings and spuriously reports an escape for a workspace that never
// escaped (this shipment's §2 root cause, a fail-closed regression
// introduced by 013-S).
func TestNewRootOnJunctionRootedWorkspaceResolvesExistingDescendant(t *testing.T) {
	targetDir := t.TempDir()
	existingLeaf := filepath.Join(targetDir, "existing.txt")
	if err := os.WriteFile(existingLeaf, []byte("x"), 0o644); err != nil {
		t.Fatalf("failed to create existing leaf in junction target: %v", err)
	}

	junctionParent := t.TempDir()
	junctionRootPath := filepath.Join(junctionParent, "workspace-root")
	createDirectoryJunction(t, junctionRootPath, targetDir)

	root, err := NewRoot(junctionRootPath)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", junctionRootPath, err)
	}

	got, err := root.Resolve("existing.txt")
	if err != nil {
		t.Fatalf("Resolve(\"existing.txt\") returned unexpected error: %v -- want the existing descendant of a junction-rooted workspace to resolve successfully", err)
	}
	if !strings.EqualFold(filepath.Base(got), "existing.txt") {
		t.Fatalf("Resolve(\"existing.txt\") = %q, want a path ending in existing.txt", got)
	}
}

// TestNewRootOnJunctionRootedWorkspaceRejectsEscapeViaInnerJunction is the
// C1b containment-preservation lock (016.006-T, security review addition,
// confidence 0.72): C1 alone proves only that a junction-rooted workspace
// does not OVER-reject. On a NON-NEGOTIABLE containment control, this test
// proves the complementary property -- that it does not UNDER-reject either.
// Under a junction-rooted workspace, an in-root reparse point whose target
// lies OUTSIDE the resolved root must still be rejected.
//
// Unlike C1 (016.005-T), this is NOT a RED->GREEN unit: it is written to
// PASS both before and after 016.007-T's (C2) GREEN change, because
// containment is preserved on both sides of that change (pre-C2, the escape
// is still caught by canonicalizeReparse's own out-of-root resolution
// compared against the unresolved junction root path; post-C2, the same
// escape is caught against the now-resolved root path). A regression here at
// any point is a STOP condition (plan H3b).
func TestNewRootOnJunctionRootedWorkspaceRejectsEscapeViaInnerJunction(t *testing.T) {
	trulyOutside := t.TempDir()

	targetDir := t.TempDir()
	escapeLinkPath := filepath.Join(targetDir, "escape-link")
	createDirectoryJunction(t, escapeLinkPath, trulyOutside)

	junctionParent := t.TempDir()
	junctionRootPath := filepath.Join(junctionParent, "workspace-root")
	createDirectoryJunction(t, junctionRootPath, targetDir)

	root, err := NewRoot(junctionRootPath)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", junctionRootPath, err)
	}

	_, err = root.Resolve("escape-link")
	if err == nil {
		t.Fatalf("Resolve(\"escape-link\") = nil error, want %q -- an in-root reparse point targeting outside the resolved root must still be rejected", symlinkEscapeMsg)
	}
	if !strings.Contains(err.Error(), symlinkEscapeMsg) {
		t.Fatalf("Resolve(\"escape-link\") error = %q, want to contain %q", err.Error(), symlinkEscapeMsg)
	}
}

// regression lock (016.005-T, availability direction, transition (b)): a
// NON-EXISTENT leaf under a junction-rooted workspace must resolve to an
// in-root candidate path (the ordinary "create a new file" case), not be
// rejected.
//
// RED against current main: the non-existent leaf causes the ancestor-climb
// loop in checkSymlinkEscape to ascend to the junction root itself as the
// first existing entry. os.Lstat on that live junction reports
// os.ModeIrregular (verified: IsDir=false, ModeSymlink unset), which trips
// F3's unconditional "!finalProbe && !info.IsDir() && !ModeSymlink" ancestor
// rejection (symlinkUnverifiableMsg) before canonicalizeReparse is ever
// reached -- a DIFFERENT failing code path from transition (a) above, which
// is why both transitions are covered rather than just one (plan C1/AC4).
func TestNewRootOnJunctionRootedWorkspaceResolvesNonExistentLeaf(t *testing.T) {
	targetDir := t.TempDir()

	junctionParent := t.TempDir()
	junctionRootPath := filepath.Join(junctionParent, "workspace-root")
	createDirectoryJunction(t, junctionRootPath, targetDir)

	root, err := NewRoot(junctionRootPath)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", junctionRootPath, err)
	}

	got, err := root.Resolve("new-file.txt")
	if err != nil {
		t.Fatalf("Resolve(\"new-file.txt\") returned unexpected error: %v -- want a non-existent leaf under a junction-rooted workspace to resolve (the ordinary create-a-new-file case)", err)
	}
	if !strings.EqualFold(filepath.Base(got), "new-file.txt") {
		t.Fatalf("Resolve(\"new-file.txt\") = %q, want a path ending in new-file.txt", got)
	}
}

// TestNewRootErrorSurfaceIsPathError pins C2/AC3 (016.007-T, Go review P1):
// on a missing workspace root, NewRoot's returned error must still expose an
// *fs.PathError via errors.As, even though canonicalizeReparse's Windows
// implementation returns a raw syscall.Errno rather than *fs.PathError (what
// filepath.EvalSymlinks returned pre-C2). wrapPathError re-forms the error
// at NewRoot's boundary so any caller doing errors.As(err, &pathErr) keeps
// matching on Windows exactly as it did before this shipment.
func TestNewRootErrorSurfaceIsPathError(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")

	_, err := NewRoot(missing)
	if err == nil {
		t.Fatalf("NewRoot(%q) = nil error, want an error for a missing root", missing)
	}
	var pathErr *fs.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("errors.As(err, &pathErr) = false, want true; err = %v", err)
	}
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("errors.Is(err, fs.ErrNotExist) = false, want true; err = %v", err)
	}
}

// TestNewRootResolvesEndToEndUnderVolumeGUIDRoot is the C2/AC6 end-to-end
// volume-GUID coverage (016.007-T, correctness review P2): the existing
// "volume guid" case in TestStripUNCPrefixWindowsTransformVerdicts is a
// synthetic table test that never exercises NewRoot/Resolve/hasPathPrefix.
// Because hasPathPrefix is a plain string comparison, a root/candidate
// namespace mismatch under a GUID-volume root would be invisible to it.
// This test constructs a REAL `\\?\Volume{GUID}\...` root (via mountvol's
// reported GUID for the volume hosting t.TempDir()) and proves Resolve
// still correctly accepts an in-root candidate and rejects an escape.
func TestNewRootResolvesEndToEndUnderVolumeGUIDRoot(t *testing.T) {
	dir := t.TempDir()
	guidRoot := volumeGUIDPathForTesting(t, dir)

	root, err := NewRoot(guidRoot)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", guidRoot, err)
	}
	// Measured finding: GetFinalPathNameByHandleW's default VOLUME_NAME_DOS
	// behavior normalizes a Volume{GUID} input back to its mounted
	// drive-letter form whenever the volume HAS one (as this host's does),
	// so root.Path() here is the ordinary "C:\..." form, NOT the
	// \\?\Volume{GUID}\... form. D-6's non-stripping guarantee is a
	// property of stripUNCPrefix as a pure function (pinned by
	// TestStripUNCPrefixWindowsTransformVerdicts) and matters for a volume
	// with NO drive-letter mapping, which this fixture (backed by the
	// drive hosting t.TempDir()) cannot construct. What this test proves
	// instead is the AC6 gap it was added to close: end-to-end
	// NewRoot/Resolve/hasPathPrefix correctness when NewRoot's INPUT is a
	// Volume{GUID}-form path, regardless of what internal form the
	// canonicalized root ends up in.
	if !filepath.IsAbs(root.Path()) {
		t.Fatalf("NewRoot(%q).Path() = %q, want an absolute path", guidRoot, root.Path())
	}

	if err := os.WriteFile(filepath.Join(dir, "in-root.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("failed to create in-root file: %v", err)
	}

	got, err := root.Resolve("in-root.txt")
	if err != nil {
		t.Fatalf("Resolve(\"in-root.txt\") returned unexpected error: %v", err)
	}
	if !strings.EqualFold(filepath.Base(got), "in-root.txt") {
		t.Fatalf("Resolve(\"in-root.txt\") = %q, want a path ending in in-root.txt", got)
	}

	if _, err := root.Resolve(`..\..\escape.txt`); err == nil {
		t.Fatalf(`Resolve("..\..\escape.txt") = nil error, want rejection`)
	}
}

// volumeGUIDPathForTesting re-forms dir (an absolute drive-letter path) as
// its `\\?\Volume{GUID}\...` equivalent, using the live volume GUID that
// mountvol reports for dir's drive. Skips (t.Skip, not t.Fatal: this is an
// environment capability probe, not the behavior under test) if mountvol is
// unavailable or reports no GUID for dir's drive -- both legitimately
// possible on a non-NTFS or restricted CI image, unlike the RED-lock
// junction tests, whose capability (pwsh) is expected to always be present.
func volumeGUIDPathForTesting(t *testing.T, dir string) string {
	t.Helper()

	drive := filepath.VolumeName(dir) // e.g. "C:"
	if drive == "" {
		t.Skip("cannot determine drive volume name for GUID lookup")
	}

	out, err := exec.Command("cmd", "/c", "mountvol", drive+`\`, "/L").CombinedOutput()
	if err != nil {
		t.Skipf("mountvol unavailable or failed: %v; output: %s", err, out)
	}
	guidPath := strings.TrimSpace(string(out))
	if !strings.HasPrefix(guidPath, uncPrefix+"Volume{") {
		t.Skipf("mountvol did not report a Volume{GUID} path: %q", guidPath)
	}

	rest := strings.TrimPrefix(dir, drive)
	rest = strings.TrimPrefix(rest, `\`)
	return guidPath + rest
}

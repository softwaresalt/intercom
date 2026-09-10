package pathsafe

import (
	"os"
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

// TestNewRootOnJunctionRootedWorkspaceResolvesNonExistentLeaf is the C1 RED
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

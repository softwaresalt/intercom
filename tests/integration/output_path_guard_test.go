// Package integration exercises repository scripts directly without running the full build matrix.
package integration

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

const outputPathGuardTimeout = time.Minute

func runOutputPathGuard(t *testing.T, testRepoRoot, outputDir string) ([]byte, error) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), outputPathGuardTimeout)
	defer cancel()

	// GuardPath is deliberately decoupled from RepoRoot (011.011-T): the
	// actual OutputPathGuard.ps1 script always lives at the real checkout
	// root, but RepoRoot is a fixture value a test may set to something
	// else entirely (e.g. a filesystem root like "C:\" or "/", which would
	// be privileged/impractical to actually place the script at).
	checkoutRoot := repoRoot(t)

	helperScript := filepath.Join(t.TempDir(), "invoke-output-path-guard.ps1")
	helperContent := strings.Join([]string{
		"param(",
		"    [Parameter(Mandatory)][string]$GuardPath,",
		"    [Parameter(Mandatory)][string]$RepoRoot,",
		"    [Parameter(Mandatory)][string]$OutputDir,",
		"    [Parameter(Mandatory)][string]$BasePath",
		")",
		"$ErrorActionPreference = 'Stop'",
		". $GuardPath",
		"$resolved = Resolve-OutputPathWithinRoot -RepoRoot $RepoRoot -OutputDir $OutputDir -BasePath $BasePath",
		"Write-Output $resolved",
		"",
	}, "\n")
	if err := os.WriteFile(helperScript, []byte(helperContent), 0o600); err != nil {
		t.Fatalf("writing PowerShell helper script: %v", err)
	}

	guardPath := filepath.Join(checkoutRoot, "scripts", "lib", "OutputPathGuard.ps1")
	cmd := exec.CommandContext(ctx, "pwsh", "-NoProfile", "-File", helperScript, "-GuardPath", guardPath, "-RepoRoot", testRepoRoot, "-OutputDir", outputDir, "-BasePath", checkoutRoot)
	cmd.Dir = checkoutRoot
	out, err := cmd.CombinedOutput()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("output-path guard exceeded %s timeout; output so far:\n%s", outputPathGuardTimeout, out)
	}
	return out, err
}

func inRepoGuardDir(t *testing.T, root string) string {
	t.Helper()

	distRoot := filepath.Join(root, "dist")
	if err := os.MkdirAll(distRoot, 0o755); err != nil {
		t.Fatalf("creating dist/ root: %v", err)
	}
	dir, err := os.MkdirTemp(distRoot, "output-path-guard-")
	if err != nil {
		t.Fatalf("creating in-repo guard fixture dir: %v", err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(dir); err != nil && !os.IsNotExist(err) {
			t.Fatalf("cleaning in-repo guard fixture dir %q: %v", dir, err)
		}
	})
	return dir
}

func createDirectorySymlink(t *testing.T, linkPath, targetPath string) {
	t.Helper()

	if err := os.Symlink(targetPath, linkPath); err != nil {
		t.Skipf("directory symlink creation unavailable (missing symbolic-link privilege / Developer Mode?): %v", err)
	}
	t.Cleanup(func() {
		if err := os.Remove(linkPath); err != nil && !os.IsNotExist(err) {
			t.Fatalf("removing directory symlink %q: %v", linkPath, err)
		}
	})
}

func createDirectoryJunction(t *testing.T, linkPath, targetPath string) {
	t.Helper()

	if runtime.GOOS != "windows" {
		t.Skip("Windows junction creation requires Windows")
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

	ctx, cancel := context.WithTimeout(context.Background(), outputPathGuardTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "pwsh", "-NoProfile", "-File", helperScript, "-LinkPath", linkPath, "-TargetPath", targetPath)
	out, err := cmd.CombinedOutput()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("junction creation exceeded %s timeout; output so far:\n%s", outputPathGuardTimeout, out)
	}
	if err != nil {
		t.Fatalf("creating junction %q -> %q: %v\noutput:\n%s", linkPath, targetPath, err, out)
	}
	t.Cleanup(func() {
		if err := os.Remove(linkPath); err != nil && !os.IsNotExist(err) {
			t.Fatalf("removing junction %q: %v", linkPath, err)
		}
	})
}

func TestResolveOutputPathWithinRoot(t *testing.T) {
	if _, err := exec.LookPath("pwsh"); err != nil {
		t.Skip("pwsh not available on PATH")
	}

	root := repoRoot(t)

	t.Run("rejects directory symlink escape", func(t *testing.T) {
		fixtureRoot := inRepoGuardDir(t, root)
		outsideRoot := filepath.Join(t.TempDir(), "outside")
		if err := os.MkdirAll(outsideRoot, 0o755); err != nil {
			t.Fatalf("creating outside root: %v", err)
		}

		linkPath := filepath.Join(fixtureRoot, "symlink-out")
		createDirectorySymlink(t, linkPath, outsideRoot)

		out, err := runOutputPathGuard(t, root, filepath.Join(linkPath, "nested", "missing"))
		if err == nil {
			t.Fatalf("expected symlink escape to be rejected, got success with output:\n%s", out)
		}
		if !strings.Contains(string(out), "invariant I5") {
			t.Fatalf("expected invariant I5 in refusal output, got:\n%s", out)
		}
	})

	t.Run("rejects dangling symlink escape", func(t *testing.T) {
		// The symlink target intentionally does NOT exist. Existence checks
		// that follow a reparse point to its target (rather than detecting
		// the reparse point itself) can report a dangling link as "does not
		// exist" and skip resolution entirely, falling through to a lexical
		// join that can pass the in-root check even though the eventual
		// target is outside the repo root. This guards against that gap.
		fixtureRoot := inRepoGuardDir(t, root)
		outsideRoot := filepath.Join(t.TempDir(), "outside-not-created-yet")

		linkPath := filepath.Join(fixtureRoot, "dangling-symlink-out")
		createDirectorySymlink(t, linkPath, outsideRoot)

		out, err := runOutputPathGuard(t, root, filepath.Join(linkPath, "nested", "missing"))
		if err == nil {
			t.Fatalf("expected dangling symlink escape to be rejected, got success with output:\n%s", out)
		}
		if !strings.Contains(string(out), "invariant I5") {
			t.Fatalf("expected invariant I5 in refusal output, got:\n%s", out)
		}
	})

	t.Run("rejects junction escape", func(t *testing.T) {
		fixtureRoot := inRepoGuardDir(t, root)
		outsideRoot := filepath.Join(t.TempDir(), "outside")
		if err := os.MkdirAll(outsideRoot, 0o755); err != nil {
			t.Fatalf("creating outside root: %v", err)
		}

		linkPath := filepath.Join(fixtureRoot, "junction-out")
		createDirectoryJunction(t, linkPath, outsideRoot)

		out, err := runOutputPathGuard(t, root, filepath.Join(linkPath, "nested", "missing"))
		if err == nil {
			t.Fatalf("expected junction escape to be rejected, got success with output:\n%s", out)
		}
		if !strings.Contains(string(out), "invariant I5") {
			t.Fatalf("expected invariant I5 in refusal output, got:\n%s", out)
		}
	})

	t.Run("accepts in-root symlink", func(t *testing.T) {
		fixtureRoot := inRepoGuardDir(t, root)
		targetRoot := filepath.Join(fixtureRoot, "actual-target")
		if err := os.MkdirAll(targetRoot, 0o755); err != nil {
			t.Fatalf("creating in-repo symlink target: %v", err)
		}

		linkPath := filepath.Join(fixtureRoot, "symlink-in")
		createDirectorySymlink(t, linkPath, targetRoot)

		outputDir := filepath.Join(linkPath, "nested", "missing")
		out, err := runOutputPathGuard(t, root, outputDir)
		if err != nil {
			t.Fatalf("expected in-root symlink to be accepted: %v\noutput:\n%s", err, out)
		}
		got := strings.TrimSpace(string(out))
		want := filepath.Clean(outputDir)
		if got != want {
			t.Fatalf("expected accepted path %q, got %q", want, got)
		}
	})

	t.Run("rejects lexical escape without links", func(t *testing.T) {
		outputDir := filepath.Join(root, "..", "..", "..", "outside")

		out, err := runOutputPathGuard(t, root, outputDir)
		if err == nil {
			t.Fatalf("expected lexical escape to be rejected, got success with output:\n%s", out)
		}
		if !strings.Contains(string(out), "invariant I5") {
			t.Fatalf("expected invariant I5 in refusal output, got:\n%s", out)
		}
	})

	// J12 mandatory negative test (011.011-T, resolves 707FE72B): a
	// sibling directory named "<root>-evil" must be REJECTED, not
	// accepted as a false in-root match. The separator suffix appended
	// to resolvedRepoRoot is the SOLE mechanism preventing this; a naive
	// fix for the filesystem-root false-rejection bug below could easily
	// reintroduce this as a genuine containment bypass while still
	// passing the pre-existing suite cleanly.
	t.Run("rejects sibling-prefix directory at a nested root", func(t *testing.T) {
		parent := t.TempDir()
		goodRoot := filepath.Join(parent, "ws")
		evilDir := filepath.Join(parent, "ws-evil")
		if err := os.MkdirAll(goodRoot, 0o755); err != nil {
			t.Fatalf("creating nested root: %v", err)
		}
		if err := os.MkdirAll(evilDir, 0o755); err != nil {
			t.Fatalf("creating sibling-prefix dir: %v", err)
		}

		out, err := runOutputPathGuard(t, goodRoot, evilDir)
		if err == nil {
			t.Fatalf("expected sibling-prefix directory to be rejected, got success with output:\n%s", out)
		}
		if !strings.Contains(string(out), "invariant I5") {
			t.Fatalf("expected invariant I5 in refusal output, got:\n%s", out)
		}
	})

	// Same hazard class as above, exercised at a directory immediately
	// under the filesystem root (e.g. C:\ws vs C:\ws-evil) rather than a
	// deeply nested temp directory. ACCEPTED RESIDUAL: creating a
	// directory directly under the filesystem root requires elevated
	// privileges on some CI runners (notably Linux); when creation fails
	// for a permission-shaped reason, this subtest skips rather than
	// fails -- the nested-root subtest above already exercises the
	// identical code path and carries the same hazard class.
	t.Run("rejects sibling-prefix directory at a top-level directory root", func(t *testing.T) {
		fsRoot := filepath.VolumeName(root) + string(filepath.Separator)
		if fsRoot == string(filepath.Separator) {
			fsRoot = string(filepath.Separator)
		}
		marker := "outputpathguard-toplevel-011011t"
		goodRoot := filepath.Join(fsRoot, marker)
		evilDir := filepath.Join(fsRoot, marker+"-evil")

		// Refuse to proceed if either path already exists (adversarial
		// review finding, verified: os.MkdirAll succeeds silently on an
		// already-existing directory, and the unconditional
		// os.RemoveAll cleanup below would then delete pre-existing,
		// unrelated content at a deterministic, guessable path). This
		// test may only create and later remove paths it is certain it
		// itself created.
		if _, err := os.Stat(goodRoot); err == nil {
			t.Skipf("refusing to use pre-existing path %q (would risk deleting unrelated content on cleanup)", goodRoot)
		}
		if _, err := os.Stat(evilDir); err == nil {
			t.Skipf("refusing to use pre-existing path %q (would risk deleting unrelated content on cleanup)", evilDir)
		}

		if err := os.MkdirAll(goodRoot, 0o755); err != nil {
			t.Skipf("cannot create a top-level directory at the filesystem root (privilege-shaped failure, accepted residual): %v", err)
		}
		t.Cleanup(func() { _ = os.RemoveAll(goodRoot) })
		if err := os.MkdirAll(evilDir, 0o755); err != nil {
			t.Skipf("cannot create a top-level sibling directory at the filesystem root (privilege-shaped failure, accepted residual): %v", err)
		}
		t.Cleanup(func() { _ = os.RemoveAll(evilDir) })

		out, err := runOutputPathGuard(t, goodRoot, evilDir)
		if err == nil {
			t.Fatalf("expected top-level sibling-prefix directory to be rejected, got success with output:\n%s", out)
		}
		if !strings.Contains(string(out), "invariant I5") {
			t.Fatalf("expected invariant I5 in refusal output, got:\n%s", out)
		}
	})

	// Positive regression (011.011-T): a legitimate in-root path at a
	// drive root (Windows: "C:\") is accepted. This case is UNSATISFIABLE
	// as a sibling-prefix rejection test (candidate "C:\-evil" is a
	// legitimate DESCENDANT of root "C:\", not an escape) -- it gets the
	// positive case only. Fails before the 011.011-T fix (double
	// separator from the unconditional "+ separator" bug rejects every
	// legitimate drive-root descendant); passes after.
	if runtime.GOOS == "windows" {
		t.Run("accepts legitimate path at a drive root", func(t *testing.T) {
			driveRoot := filepath.VolumeName(root) + string(filepath.Separator)
			outputDir := filepath.Join(driveRoot, "outputpathguard-driveroot-011011t-missing")

			out, err := runOutputPathGuard(t, driveRoot, outputDir)
			if err != nil {
				t.Fatalf("expected legitimate drive-root descendant to be accepted: %v\noutput:\n%s", err, out)
			}
			got := strings.TrimSpace(string(out))
			want := filepath.Clean(outputDir)
			if got != want {
				t.Fatalf("expected accepted path %q, got %q", want, got)
			}
		})
	}

	// Positive regression (011.011-T): a legitimate in-root path at a
	// POSIX root ("/") is accepted. Same rationale as the drive-root case
	// above; gets the positive case only. Fails before the fix, passes
	// after.
	if runtime.GOOS != "windows" {
		t.Run("accepts legitimate path at a POSIX root", func(t *testing.T) {
			posixRoot := string(filepath.Separator)
			outputDir := filepath.Join(posixRoot, "outputpathguard-posixroot-011011t-missing")

			out, err := runOutputPathGuard(t, posixRoot, outputDir)
			if err != nil {
				t.Fatalf("expected legitimate POSIX-root descendant to be accepted: %v\noutput:\n%s", err, out)
			}
			got := strings.TrimSpace(string(out))
			want := filepath.Clean(outputDir)
			if got != want {
				t.Fatalf("expected accepted path %q, got %q", want, got)
			}
		})
	}
}

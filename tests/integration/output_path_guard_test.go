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

func runOutputPathGuard(t *testing.T, root, outputDir string) ([]byte, error) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), outputPathGuardTimeout)
	defer cancel()

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

	guardPath := filepath.Join(root, "scripts", "lib", "OutputPathGuard.ps1")
	cmd := exec.CommandContext(ctx, "pwsh", "-NoProfile", "-File", helperScript, "-GuardPath", guardPath, "-RepoRoot", root, "-OutputDir", outputDir, "-BasePath", root)
	cmd.Dir = root
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
}

// Package integration contains cross-module tests that exercise the build
// tooling end to end.
package integration

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// buildScriptTimeout bounds every build-script invocation in this file so a
// hung cross-compile toolchain fetch or interactive prompt cannot block the
// test binary (and therefore CI) indefinitely.
const buildScriptTimeout = 5 * time.Minute

// repoRoot walks up from this test file's directory to the repository root
// (identified by the presence of go.mod).
func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate repository root (go.mod not found)")
		}
		dir = parent
	}
}

func targetCount(t *testing.T, root string) int {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(root, "scripts", "targets.json"))
	if err != nil {
		t.Fatalf("reading scripts/targets.json: %v", err)
	}
	var manifest struct {
		Targets []struct {
			GOOS   string `json:"goos"`
			GOARCH string `json:"goarch"`
		} `json:"targets"`
	}
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatalf("parsing scripts/targets.json: %v", err)
	}
	return len(manifest.Targets)
}

func runBuildScript(t *testing.T, root string, outputDir string) ([]byte, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), buildScriptTimeout)
	defer cancel()

	scriptPath := filepath.Join(root, "scripts", "build.ps1")
	cmd := exec.CommandContext(ctx, "pwsh", "-NoProfile", "-File", scriptPath, "-OutputDir", outputDir)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		t.Fatalf("build script exceeded %s timeout; output so far:\n%s", buildScriptTimeout, out)
	}
	return out, err
}

// inRepoTempDir creates a unique, gitignored scratch directory under the
// repository's dist/ tree (a legitimate in-repo output location) and
// registers its removal on test cleanup.
func inRepoTempDir(t *testing.T, root string) string {
	t.Helper()
	distRoot := filepath.Join(root, "dist")
	if err := os.MkdirAll(distRoot, 0o755); err != nil {
		t.Fatalf("creating dist/ root: %v", err)
	}
	dir, err := os.MkdirTemp(distRoot, "build-test-")
	if err != nil {
		t.Fatalf("creating in-repo scratch dir: %v", err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) }) // best-effort cleanup; leftover dirs are gitignored under dist/
	return dir
}

func binaryCount(t *testing.T, root string) int {
	t.Helper()
	entries, err := os.ReadDir(filepath.Join(root, "cmd"))
	if err != nil {
		t.Fatalf("reading cmd/ directory: %v", err)
	}
	count := 0
	for _, e := range entries {
		if e.IsDir() {
			count++
		}
	}
	return count
}

// TestBuildScript_ProducesArtifactsForEveryTarget asserts scripts/build.ps1
// produces one artifact per binary per target.
func TestBuildScript_ProducesArtifactsForEveryTarget(t *testing.T) {
	if _, err := exec.LookPath("pwsh"); err != nil {
		t.Skip("pwsh not available on PATH")
	}

	root := repoRoot(t)
	wantTargets := targetCount(t, root)
	wantBinaries := binaryCount(t, root)
	outputDir := inRepoTempDir(t, root)

	out, err := runBuildScript(t, root, outputDir)
	if err != nil {
		t.Fatalf("build script failed: %v\noutput:\n%s", err, out)
	}

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("reading output dir: %v", err)
	}

	wantCount := wantTargets * wantBinaries
	if len(entries) != wantCount {
		t.Fatalf("expected %d artifacts (%d binaries x %d targets), got %d: %v", wantCount, wantBinaries, wantTargets, len(entries), entries)
	}
}

// TestBuildScript_ProducesCGOFreeBinaries asserts a produced artifact reports
// CGO_ENABLED=0 via 'go version -m'.
func TestBuildScript_ProducesCGOFreeBinaries(t *testing.T) {
	if _, err := exec.LookPath("pwsh"); err != nil {
		t.Skip("pwsh not available on PATH")
	}

	root := repoRoot(t)
	outputDir := inRepoTempDir(t, root)

	out, err := runBuildScript(t, root, outputDir)
	if err != nil {
		t.Fatalf("build script failed: %v\noutput:\n%s", err, out)
	}

	// Find a produced artifact matching the host GOOS/GOARCH so it can be
	// inspected directly.
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		t.Fatalf("reading output dir: %v", err)
	}
	var artifact string
	suffix := runtime.GOOS
	for _, e := range entries {
		if strings.Contains(e.Name(), suffix) {
			artifact = filepath.Join(outputDir, e.Name())
			break
		}
	}
	if artifact == "" && len(entries) > 0 {
		artifact = filepath.Join(outputDir, entries[0].Name())
	}
	if artifact == "" {
		t.Fatal("no artifacts produced to inspect")
	}

	cmd := exec.Command("go", "version", "-m", artifact)
	verOut, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("go version -m failed: %v\noutput:\n%s", err, verOut)
	}
	if !strings.Contains(string(verOut), "CGO_ENABLED=0") {
		t.Fatalf("expected build info to report CGO_ENABLED=0, got:\n%s", verOut)
	}
}

// TestBuildScript_RefusesOutputPathOutsideRepoRoot is the negative test for
// invariant I5: the script must exit non-zero when the resolved output path
// escapes the repository root.
func TestBuildScript_RefusesOutputPathOutsideRepoRoot(t *testing.T) {
	if _, err := exec.LookPath("pwsh"); err != nil {
		t.Skip("pwsh not available on PATH")
	}

	root := repoRoot(t)
	scriptPath := filepath.Join(root, "scripts", "build.ps1")
	if _, statErr := os.Stat(scriptPath); statErr != nil {
		t.Fatalf("scripts/build.ps1 must exist for this negative test to be meaningful: %v", statErr)
	}

	outsideDir := filepath.Join(t.TempDir(), "escaped-dist")

	// t.TempDir() is outside the repo root by construction.
	out, err := runBuildScript(t, root, outsideDir)
	if err == nil {
		t.Fatal("expected the build script to fail for an out-of-repo output path, got nil error")
	}
	if _, statErr := os.Stat(outsideDir); statErr == nil {
		t.Fatalf("build script created the out-of-repo output path %q; it must refuse before writing", outsideDir)
	}
	if !strings.Contains(string(out), "repo") && !strings.Contains(string(out), "root") {
		t.Fatalf("expected the refusal message to reference the repo-root guard, got:\n%s", out)
	}
}

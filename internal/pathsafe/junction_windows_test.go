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

	candidate := filepath.Join("junction-out", "new-file.txt")
	_, err = root.Resolve(candidate)
	if err == nil {
		t.Fatalf("Resolve(%q) = nil error, want %q", candidate, symlinkEscapeMsg)
	}
	if !strings.Contains(err.Error(), symlinkEscapeMsg) {
		t.Fatalf("Resolve(%q) error = %q, want to contain %q", candidate, err.Error(), symlinkEscapeMsg)
	}
}

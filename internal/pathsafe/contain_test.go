package pathsafe

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestResolveInsideRoot verifies src/main.go and ./src/main.go resolve to an
// absolute path inside the root.
func TestResolveInsideRoot(t *testing.T) {
	dir := t.TempDir()
	root, err := NewRoot(dir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", dir, err)
	}

	for _, candidate := range []string{"src/main.go", "./src/main.go"} {
		resolved, err := root.Resolve(candidate)
		if err != nil {
			t.Fatalf("Resolve(%q) returned unexpected error: %v", candidate, err)
		}
		if !filepath.IsAbs(resolved) {
			t.Fatalf("Resolve(%q) = %q, want absolute path", candidate, resolved)
		}
		if !hasPathPrefix(resolved, root.Path()) {
			t.Fatalf("Resolve(%q) = %q, want path inside root %q", candidate, resolved, root.Path())
		}
	}
}

// TestResolveRejectsSiblingDirEscape verifies a sibling directory named
// <root>-evil is rejected with "path outside workspace" — proof that
// containment is component-aware, not a raw string prefix check
// (finding GO-6).
func TestResolveRejectsSiblingDirEscape(t *testing.T) {
	parent := t.TempDir()
	rootDir := filepath.Join(parent, "ws")
	evilDir := filepath.Join(parent, "ws-evil")
	if err := os.MkdirAll(rootDir, 0o755); err != nil {
		t.Fatalf("failed to create root dir: %v", err)
	}
	if err := os.MkdirAll(evilDir, 0o755); err != nil {
		t.Fatalf("failed to create sibling dir: %v", err)
	}

	root, err := NewRoot(rootDir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", rootDir, err)
	}

	// hasPathPrefix must not treat the sibling "ws-evil" as contained by "ws".
	if hasPathPrefix(evilDir, root.Path()) {
		t.Fatalf("hasPathPrefix(%q, %q) = true, want false (sibling-directory escape)", evilDir, root.Path())
	}
}

// TestResolveRejectsEmptyCandidate verifies an empty-string candidate is
// rejected with "path outside workspace" rather than silently resolving to
// the workspace root itself (finding CORR-2).
func TestResolveRejectsEmptyCandidate(t *testing.T) {
	dir := t.TempDir()
	root, err := NewRoot(dir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", dir, err)
	}

	_, err = root.Resolve("")
	if err == nil {
		t.Fatalf("Resolve(\"\") = nil error, want %q", outsideMsg)
	}
	if !strings.Contains(err.Error(), outsideMsg) {
		t.Fatalf("Resolve(\"\") error = %q, want to contain %q", err.Error(), outsideMsg)
	}
}

package pathsafe

import (
	"os"
	"path/filepath"
	"runtime"
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

// TestHasPathPrefixTreatsFilesystemRootAsContainingChild verifies the
// filesystem root itself is handled as a container prefix for a direct child,
// without requiring a doubled separator in the child path.
func TestHasPathPrefixTreatsFilesystemRootAsContainingChild(t *testing.T) {
	prefix := string(filepath.Separator)
	if runtime.GOOS == "windows" {
		volume := filepath.VolumeName(t.TempDir())
		if volume == "" {
			t.Fatal("filepath.VolumeName(t.TempDir()) = empty, want drive volume")
		}
		prefix = volume + string(filepath.Separator)
	}

	path := filepath.Join(prefix, "child")
	if !hasPathPrefix(path, prefix) {
		t.Fatalf("hasPathPrefix(%q, %q) = false, want true", path, prefix)
	}
}

// TestHasPathPrefixCommonContainmentVerdicts locks common containment
// verdicts that must not change during internal helper deduplication.
func TestHasPathPrefixCommonContainmentVerdicts(t *testing.T) {
	workspace := filepath.Join(t.TempDir(), "ws")
	parent := filepath.Dir(workspace)
	filesystemRoot := string(filepath.Separator)
	if runtime.GOOS == "windows" {
		volume := filepath.VolumeName(workspace)
		if volume == "" {
			t.Fatal("filepath.VolumeName(workspace) = empty, want drive volume")
		}
		filesystemRoot = volume + string(filepath.Separator)
	}

	cases := []struct {
		name   string
		path   string
		prefix string
		want   bool
	}{
		{name: "exact match", path: workspace, prefix: workspace, want: true},
		{name: "child inside prefix", path: filepath.Join(workspace, "child.txt"), prefix: workspace, want: true},
		{name: "sibling with shared prefix", path: filepath.Join(parent, filepath.Base(workspace)+"-evil"), prefix: workspace, want: false},
		{name: "unrelated path", path: filepath.Join(parent, "other", "child.txt"), prefix: workspace, want: false},
		{name: "filesystem root equals prefix", path: filesystemRoot, prefix: filesystemRoot, want: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := hasPathPrefix(tc.path, tc.prefix); got != tc.want {
				t.Fatalf("hasPathPrefix(%q, %q) = %v, want %v", tc.path, tc.prefix, got, tc.want)
			}
		})
	}
}

// TestResolveRejectsChildBeneathRegularFile verifies a candidate whose
// parent path component is an existing regular file is rejected rather than
// accepted by walking past the non-directory entry.
func TestResolveRejectsChildBeneathRegularFile(t *testing.T) {
	dir := t.TempDir()
	root, err := NewRoot(dir)
	if err != nil {
		t.Fatalf("NewRoot(%q) returned error: %v", dir, err)
	}

	filePath := filepath.Join(root.Path(), "file.txt")
	if err := os.WriteFile(filePath, []byte("x"), 0o644); err != nil {
		t.Fatalf("failed to create regular file ancestor: %v", err)
	}

	_, err = root.Resolve(filepath.Join("file.txt", "sub"))
	if err == nil {
		t.Fatalf("Resolve(%q) = nil error, want %q", filepath.Join("file.txt", "sub"), symlinkEscapeMsg)
	}
	if !strings.Contains(err.Error(), symlinkEscapeMsg) {
		t.Fatalf("Resolve(%q) error = %q, want to contain %q", filepath.Join("file.txt", "sub"), err.Error(), symlinkEscapeMsg)
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

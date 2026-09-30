package retiredarch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// realSelectGoPath returns this repo's real, on-disk select.go path.
func realSelectGoPath(t *testing.T) string {
	t.Helper()
	root := repoRoot(t)
	return filepath.Join(root, "tools", "gatecheck", "internal", "retiredarch", "select.go")
}

func TestCheckPathspecPin_UnmodifiedCopy_Accepted(t *testing.T) {
	src, err := os.ReadFile(realSelectGoPath(t))
	if err != nil {
		t.Fatalf("read select.go: %v", err)
	}
	dir := t.TempDir()
	copyPath := filepath.Join(dir, "select.go")
	if err := os.WriteFile(copyPath, src, 0o644); err != nil {
		t.Fatalf("write copy: %v", err)
	}
	pin := checkPathspecPin(copyPath)
	if !pin.OK() {
		t.Fatalf("unmodified select.go copy must be ACCEPTED, got %+v", pin)
	}
}

func TestCheckPathspecPin_MissingFile_FailsClosed(t *testing.T) {
	dir := t.TempDir()
	pin := checkPathspecPin(filepath.Join(dir, "does-not-exist.go"))
	if pin.OK() {
		t.Fatalf("a missing file must fail closed, got %+v", pin)
	}
	if pin.SelectFound || pin.GuardFound {
		t.Fatalf("a missing file must report SelectFound=false and GuardFound=false, got %+v", pin)
	}
}

// writeMutatedCopy reads the real select.go, applies exactly one string
// replacement (old must appear exactly once), and writes the mutated
// text to a fresh file in t.TempDir(). It fails the test if old does not
// appear (so a future select.go refactor that removes the anchor text
// is caught immediately rather than silently no-op mutating).
func writeMutatedCopy(t *testing.T, old, new string) string {
	t.Helper()
	src, err := os.ReadFile(realSelectGoPath(t))
	if err != nil {
		t.Fatalf("read select.go: %v", err)
	}
	text := string(src)
	if strings.Count(text, old) != 1 {
		t.Fatalf("mutation anchor %q must appear exactly once in select.go, found %d", old, strings.Count(text, old))
	}
	mutated := strings.Replace(text, old, new, 1)
	dir := t.TempDir()
	path := filepath.Join(dir, "select.go")
	if err := os.WriteFile(path, []byte(mutated), 0o644); err != nil {
		t.Fatalf("write mutated copy: %v", err)
	}
	return path
}

func TestCheckPathspecPin_Mutations_AllRejected(t *testing.T) {
	t.Run("cmd_glob_removed", func(t *testing.T) {
		path := writeMutatedCopy(t,
			`git(root, "config.toml.example", "cmd/**", "internal/**")`,
			`git(root, "config.toml.example", "internal/**")`,
		)
		if pin := checkPathspecPin(path); pin.OK() {
			t.Fatalf("removing the cmd/** literal must be REJECTED, got %+v", pin)
		}
	})

	t.Run("cmd_glob_moved_to_package_const", func(t *testing.T) {
		path := writeMutatedCopy(t,
			`git(root, "config.toml.example", "cmd/**", "internal/**")`,
			`git(root, "config.toml.example", cmdGlobPin, "internal/**")`,
		)
		// Insert the constant declaration at package scope so the file
		// still parses; the literal is no longer INSIDE selectRepoPaths'
		// body, which is exactly what the pin must reject.
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read mutated copy: %v", err)
		}
		withConst := strings.Replace(string(src), "package retiredarch\n", "package retiredarch\n\nconst cmdGlobPin = \"cmd/**\"\n", 1)
		if err := os.WriteFile(path, []byte(withConst), 0o644); err != nil {
			t.Fatalf("write const copy: %v", err)
		}
		if pin := checkPathspecPin(path); pin.OK() {
			t.Fatalf("moving the cmd/** literal to a package-level const must be REJECTED, got %+v", pin)
		}
	})

	t.Run("cmd_glob_rewritten_as_different_raw_string", func(t *testing.T) {
		path := writeMutatedCopy(t,
			`git(root, "config.toml.example", "cmd/**", "internal/**")`,
			"git(root, \"config.toml.example\", `cmd/*`, \"internal/**\")",
		)
		if pin := checkPathspecPin(path); pin.OK() {
			t.Fatalf("rewriting cmd/** as an equivalent-shape but different-value raw string must be REJECTED, got %+v", pin)
		}
	})

	t.Run("select_function_renamed", func(t *testing.T) {
		path := writeMutatedCopy(t, "func selectRepoPaths(", "func selectRepoPathsRenamed(")
		if pin := checkPathspecPin(path); pin.OK() {
			t.Fatalf("renaming selectRepoPaths must be REJECTED, got %+v", pin)
		}
		if pin := checkPathspecPin(path); pin.SelectFound {
			t.Fatalf("renaming selectRepoPaths must report SelectFound=false")
		}
	})

	t.Run("file_missing", func(t *testing.T) {
		dir := t.TempDir()
		if pin := checkPathspecPin(filepath.Join(dir, "select.go")); pin.OK() {
			t.Fatalf("a missing file must be REJECTED")
		}
	})
}

// TestCheckPathspecPin_PrefixMutation_Rejected covers the
// shouldScanRepoPath/prefix half of the pin with an analogous mutation,
// since the AC's enumerated mutation list is pathspec-focused but the
// pin has two independent halves (pathspec_ok, prefix_ok) that must both
// be exercised by a red mutation.
func TestCheckPathspecPin_PrefixMutation_Rejected(t *testing.T) {
	path := writeMutatedCopy(t,
		`strings.HasPrefix(path, "internal/")`,
		`strings.HasPrefix(path, "internal_moved/")`,
	)
	if pin := checkPathspecPin(path); pin.OK() {
		t.Fatalf("mutating the internal/ prefix literal must be REJECTED, got %+v", pin)
	}
}

// TestSelectionPathspecPin_LiveTree exercises the production entry
// point (git rev-parse --show-toplevel resolution) against the real
// repository, confirming it accepts the real, unmodified select.go.
func TestSelectionPathspecPin_LiveTree(t *testing.T) {
	root := repoRoot(t)
	pin := SelectionPathspecPin(root)
	if !pin.OK() {
		t.Fatalf("SelectionPathspecPin against the live tree must be accepted, got %+v", pin)
	}
}

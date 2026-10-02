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
			`{pathspec: "cmd/**", prefix: "cmd/", includeTests: true},`,
			`{prefix: "cmd/", includeTests: true},`,
		)
		if pin := checkPathspecPin(path); pin.OK() {
			t.Fatalf("removing the cmd/** literal must be REJECTED, got %+v", pin)
		}
	})

	t.Run("cmd_glob_moved_to_package_const", func(t *testing.T) {
		path := writeMutatedCopy(t,
			`{pathspec: "cmd/**", prefix: "cmd/", includeTests: true},`,
			`{pathspec: cmdGlobPin, prefix: "cmd/", includeTests: true},`,
		)
		// Insert the constant declaration at package scope so the file
		// still parses; the literal is no longer INSIDE the scanScope
		// declaration, which is exactly what the pin must reject.
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
			`{pathspec: "cmd/**", prefix: "cmd/", includeTests: true},`,
			"{pathspec: `cmd/*`, prefix: \"cmd/\", includeTests: true},",
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
		`{pathspec: "internal/**", prefix: "internal/", includeTests: false},`,
		`{pathspec: "internal/**", prefix: "internal_moved/", includeTests: false},`,
	)
	if pin := checkPathspecPin(path); pin.OK() {
		t.Fatalf("mutating the internal/ prefix literal must be REJECTED, got %+v", pin)
	}
}

// TestPinLiterals_NonEmpty is the AC-A2.2 non-vacuity guard: containsAll
// is vacuously true for an empty wanted list, so emptying either pin list
// would turn the pin into one that asserts nothing.
func TestPinLiterals_NonEmpty(t *testing.T) {
	if len(pathspecPinLiterals) == 0 {
		t.Fatalf("pathspecPinLiterals must be non-empty")
	}
	if len(prefixPinLiterals) == 0 {
		t.Fatalf("prefixPinLiterals must be non-empty")
	}
}

// TestCheckPathspecPin_NarrowedScope_Rejected is the AC-A2.3 negative
// control: a select.go whose scanScope omits the cmd/** arm (a narrowed
// scan scope) must make the pin FAIL.
func TestCheckPathspecPin_NarrowedScope_Rejected(t *testing.T) {
	path := writeMutatedCopy(t,
		"\t{pathspec: \"cmd/**\", prefix: \"cmd/\", includeTests: true},\n",
		"",
	)
	pin := checkPathspecPin(path)
	if pin.OK() {
		t.Fatalf("a scanScope without the cmd/** arm must be REJECTED, got %+v", pin)
	}
	if pin.PathspecOK || pin.PrefixOK {
		t.Fatalf("a scanScope without the cmd/** arm must fail both pin halves, got %+v", pin)
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

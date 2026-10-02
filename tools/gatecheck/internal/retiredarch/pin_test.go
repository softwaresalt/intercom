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

// TestContainsAll_EmptyInput_False is the AC-A3a.1 pure-predicate test:
// containsAll must not be vacuously true. An empty wanted list or an empty
// haystack both yield false, so a pin whose literal list or collected
// literal set was emptied fails closed.
func TestContainsAll_EmptyInput_False(t *testing.T) {
	cases := []struct {
		name     string
		haystack []string
		wanted   []string
	}{
		{name: "empty_wanted", haystack: []string{"cmd/**"}, wanted: nil},
		{name: "empty_wanted_non_nil", haystack: []string{"cmd/**"}, wanted: []string{}},
		{name: "empty_haystack", haystack: nil, wanted: []string{"cmd/**"}},
		{name: "both_empty", haystack: nil, wanted: nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if containsAll(tc.haystack, tc.wanted) {
				t.Fatalf("containsAll(%q, %q) = true, want false", tc.haystack, tc.wanted)
			}
		})
	}
	if !containsAll([]string{"a", "b"}, []string{"b"}) {
		t.Fatalf("containsAll must still accept a non-empty subset")
	}
}

// TestCheckPathspecPin_FrozenDecls_RejectTable is the AC-A3b.1 reject
// table (one scenario: a single token-equality / shared-rule predicate over
// data rows). Every row keeps scanScope's literals present, so the
// ALP-1 + A-T3a presence pin accepts it (red-phase rule, rev 6). Each row
// asserts SelectFound==true AND PathspecOK==false (and PrefixOK==false for
// shared-rule rows), so a parse error cannot satisfy it.
func TestCheckPathspecPin_FrozenDecls_RejectTable(t *testing.T) {
	rows := []struct {
		name, old, new string
		shared         bool
	}{
		{"git_call_in_dead_code",
			"\tout, err := git(root, pathspecs...)\n",
			"\tvar out []byte\n\tvar err error\n\tif false {\n\t\tout, err = git(root, pathspecs...)\n\t}\n", false},
		{"discarded_git_call_second_call_produces_out",
			"\tout, err := git(root, pathspecs...)\n",
			"\t_, _ = git(root, pathspecs...)\n\tout, err := git(root)\n", false},
		{"root_reassigned",
			"\tout, err := git(root, pathspecs...)\n",
			"\troot = \"/tmp\"\n\tout, err := git(root, pathspecs...)\n", false},
		{"second_runner_exec_command",
			"\tout, err := git(root, pathspecs...)\n",
			"\tout, err := exec.Command(\"git\", \"ls-files\").Output()\n", false},
		{"extra_filter",
			"if shouldScanRepoPath(path) {",
			"if shouldScanRepoPath(path) && !strings.HasPrefix(path, \"cmd/\") {", false},
		{"pathspecs_resliced",
			"git(root, pathspecs...)",
			"git(root, pathspecs[1:]...)", false},
		{"default_git_runner_drops_pathspecs",
			"args := append([]string{\"ls-files\", \"--\"}, pathspecs...)",
			"args := []string{\"ls-files\", \"--\", \"internal/**\"}", false},
		{"extra_exclude_cmd_arm",
			"\t{pathspec: \"internal/**\", prefix: \"internal/\", includeTests: false},\n",
			"\t{pathspec: \"internal/**\", prefix: \"internal/\", includeTests: false},\n\t{pathspec: \":(exclude)cmd/**\", exact: \":(exclude)cmd/**\"},\n", false},
		{"duplicated_cmd_arm_first",
			"var scanScope = []scanArm{\n",
			"var scanScope = []scanArm{\n\t{pathspec: \"cmd/**\", prefix: \"cmd/\", includeTests: false},\n", false},
		{"init_func",
			"\treturn selected, nil\n}\n",
			"\treturn selected, nil\n}\n\nfunc init() {}\n", true},
		{"go_build_constraint",
			"package retiredarch\n",
			"//go:build linux\n\npackage retiredarch\n", true},
		{"aliased_import",
			"\t\"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem\"\n",
			"\tps \"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem\"\n", false},
		{"engine_for_path_references_scan_scope",
			"\tif filepath.Base(path) == \"config.toml.example\" {\n",
			"\tif len(scanScope) == 0 {\n\t\treturn \"\"\n\t}\n\tif filepath.Base(path) == \"config.toml.example\" {\n", true},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			pin := checkPathspecPin(writeMutatedCopy(t, r.old, r.new))
			if !pin.SelectFound || pin.PathspecOK {
				t.Fatalf("want SelectFound=true and PathspecOK=false, got %+v", pin)
			}
			if r.shared && pin.PrefixOK {
				t.Fatalf("shared-rule violation must also clear PrefixOK, got %+v", pin)
			}
		})
	}
}

// TestCheckPathspecPin_FrozenDecls_PositiveControls is AC-A3b.2 (one
// scenario; green on arrival, declared as regression guards): rewritten
// comments, a //line directive and re-indented bodies do not change any
// frozen declaration's tokens, and the live select.go is accepted.
func TestCheckPathspecPin_FrozenDecls_PositiveControls(t *testing.T) {
	rows := []struct {
		name string
		pin  func(t *testing.T) PathspecPin
	}{
		{"comment_and_whitespace_insensitive", func(t *testing.T) PathspecPin {
			src, err := os.ReadFile(realSelectGoPath(t))
			if err != nil {
				t.Fatalf("read select.go: %v", err)
			}
			text := strings.ReplaceAll(string(src), "\n\t", "\n  \t ")
			text = strings.ReplaceAll(text, "// selectRepoPaths ports select_repo_paths.", "// selectRepoPaths: a REWRITTEN doc comment.")
			text = strings.Replace(text, "  \t sort.Strings(selected)\n", "//line elsewhere.go:900\n  \t /* inline */ sort.Strings(selected) // trailing\n", 1)
			path := filepath.Join(t.TempDir(), "select.go")
			if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
				t.Fatalf("write rewritten copy: %v", err)
			}
			return checkPathspecPin(path)
		}},
		{"live_tree", func(t *testing.T) PathspecPin {
			return SelectionPathspecPin(repoRoot(t))
		}},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			if pin := r.pin(t); !pin.OK() {
				t.Fatalf("must be ACCEPTED, got %+v", pin)
			}
		})
	}
}

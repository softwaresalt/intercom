package retiredarch

import (
	"go/parser"
	"go/scanner"
	"go/token"
	"go/types"
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

// TestPinLiterals_NonEmpty is the AC-A2.2 non-vacuity guard. containsAll
// already rejects an empty wanted list; this test independently asserts
// that neither pin list is emptied, so the pin always names the literals
// it guards.
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
		{"should_scan_repo_path_reformatted", func(t *testing.T) PathspecPin {
			return checkPathspecPin(writeMutatedCopy(t,
				"\t\tif strings.HasPrefix(path, a.prefix) {\n",
				"\t\t// A-T3c: comments and spacing inside shouldScanRepoPath only.\n\t\tif strings.HasPrefix( path, a.prefix ) { /* still the prefix arm */\n"))
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

// TestCheckPathspecPin_PrefixFrozen_RejectTable is the AC-A3c.1 reject
// table (one scenario: a single token-equality predicate over data rows).
// Every row changes only shouldScanRepoPath, so the A-T3b pin (presence
// for the prefix half) accepts it (red-phase rule). Each row asserts
// GuardFound==true AND PrefixOK==false, so a parse error cannot satisfy it.
func TestCheckPathspecPin_PrefixFrozen_RejectTable(t *testing.T) {
	const head = "func shouldScanRepoPath(path string) bool {\n"
	const prefixIf = "\t\tif strings.HasPrefix(path, a.prefix) {\n"
	rows := []struct{ name, old, new string }{
		{"early_return_before_loop", head, head + "\tif len(path) > 0 {\n\t\treturn false\n\t}\n"},
		{"goto", head, head + "\tgoto scan\nscan:\n"},
		{"second_loop", head, head + "\tfor range scanScope {\n\t}\n"},
		{"defer", head, head + "\tdefer func() {}()\n"},
		{"and_false_operand", "\t\t\treturn strings.HasSuffix(path, \".go\")\n", "\t\t\treturn strings.HasSuffix(path, \".go\") && false\n"},
		{"if_returns_false_for_cmd", prefixIf, "\t\tif strings.HasPrefix(path, \"cmd/\") {\n\t\t\treturn false\n\t\t}\n" + prefixIf},
		{"negated_has_prefix", prefixIf, "\t\tif !strings.HasPrefix(path, a.prefix) {\n"},
		{"swapped_has_prefix_args", prefixIf, "\t\tif strings.HasPrefix(a.prefix, path) {\n"},
		{"continue_keyed_on_cmd_prefix", prefixIf, "\t\tif a.prefix == \"cmd/\" {\n\t\t\tcontinue\n\t\t}\n" + prefixIf},
		{"path_reassigned", head, head + "\tpath = strings.ToUpper(path)\n"},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			pin := checkPathspecPin(writeMutatedCopy(t, r.old, r.new))
			if !pin.GuardFound || pin.PrefixOK {
				t.Fatalf("want GuardFound=true and PrefixOK=false, got %+v", pin)
			}
		})
	}
}

// TestScopeDataOK_Table is AC-A3c.2 (one scenario; compile-red until
// scopeDataOK existed, declared): scopeDataOK is called directly on
// []armValues. Every mutated set is rejected; the §A-CANON values pass.
func TestScopeDataOK_Table(t *testing.T) {
	canon := func() []armValues {
		return []armValues{
			{pathspec: "config.toml.example", exact: "config.toml.example"},
			{pathspec: "cmd/**", prefix: "cmd/", includeTests: true},
			{pathspec: "internal/**", prefix: "internal/"},
		}
	}
	with := func(extra armValues) []armValues { return append(canon(), extra) }
	edit := func(i int, f func(*armValues)) []armValues {
		arms := canon()
		f(&arms[i])
		return arms
	}
	rows := []struct {
		name string
		arms []armValues
		want bool
	}{
		{"canonical", canon(), true},
		{"exclude_magic", with(armValues{pathspec: ":(exclude)cmd/**", exact: ":(exclude)cmd/**"}), false},
		{"bang_magic", with(armValues{pathspec: ":!cmd/**", exact: ":!cmd/**"}), false},
		{"caret_magic", with(armValues{pathspec: ":^cmd/**", exact: ":^cmd/**"}), false},
		{"duplicate_pathspec", with(armValues{pathspec: "internal/**", prefix: "internal/"}), false},
		{"duplicate_prefix", with(armValues{pathspec: "cmd/x/**", prefix: "cmd/", includeTests: true}), false},
		{"overlapping_prefix", with(armValues{pathspec: "c**", prefix: "c"}), false},
		{"exact_under_prefix", with(armValues{pathspec: "cmd/x.go", exact: "cmd/x.go"}), false},
		{"pathspec_not_prefix_glob", func() []armValues {
			arms := canon()
			arms[1].pathspec, arms[2].pathspec = arms[2].pathspec, arms[1].pathspec
			return arms
		}(), false},
		{"missing_expected_pathspec", canon()[1:], false},
		{"extra_pathspec", with(armValues{pathspec: "README.md", exact: "README.md"}), false},
		{"cmd_include_tests_false", edit(1, func(a *armValues) { a.includeTests = false }), false},
		{"internal_include_tests_true", edit(2, func(a *armValues) { a.includeTests = true }), false},
		{"empty", nil, false},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			if got := scopeDataOK(r.arms); got != r.want {
				t.Fatalf("scopeDataOK(%+v) = %v, want %v", r.arms, got, r.want)
			}
		})
	}
}

// TestCheckPathspecPin_PackageClosure_RejectTable is AC-A3d.1 (033.007-T):
// each row writes the live select.go plus ONE extra non-test entry into a
// temp dir; package closure must clear both flags. Rows other than the
// unparseable and unreadable ones also assert SelectFound && GuardFound and
// that the extra .go file parses, so no row can pass for the wrong reason.
// A select.go-only fixture is accepted (green on arrival).
func TestCheckPathspecPin_PackageClosure_RejectTable(t *testing.T) {
	src, err := os.ReadFile(realSelectGoPath(t))
	if err != nil {
		t.Fatalf("read select.go: %v", err)
	}
	const hdr = "package retiredarch\n\n"
	rows := []struct {
		name, file, body string
		dir, broken      bool
	}{
		{name: "init mutates scanScope", file: "mutate.go", body: hdr + "func init() { scanScope = scanScope[1:] }\n"},
		{name: "mentions scanArm", file: "mention.go", body: hdr + "var _ scanArm\n"},
		{name: "func append", file: "append.go", body: hdr + "func append(xs []string, _ ...string) []string { return xs }\n"},
		{name: "var len", file: "len.go", body: hdr + "var len = 0\n"},
		{name: "type string", file: "string.go", body: hdr + "type string = []byte\n"},
		{name: "var nil true", file: "nil.go", body: hdr + "var nil, true = 0, 1\n"},
		{name: "import unsafe", file: "unsafe.go", body: hdr + "import _ \"unsafe\"\n"},
		{name: "import C", file: "cgo.go", body: hdr + "import \"C\"\n"},
		{name: "go:linkname", file: "link.go", body: hdr + "//go:linkname x runtime.x\nvar x int\n"},
		{name: "os.Setenv in init", file: "env.go", body: hdr + "import \"os\"\n\nfunc init() { _ = os.Setenv(\"GIT_INDEX_FILE\", \"x\") }\n"},
		{name: "assembly .s", file: "asm.s", body: "TEXT ·f(SB),0,$0-0\n\tRET\n"},
		{name: "object .syso", file: "blob.syso", body: "\x00\x01\x02"},
		{name: "swig .swig", file: "iface.swig", body: "%module retiredarch\n"},
		{name: "closed-world name", file: "cw.go", body: hdr + "func engineForPath() {}\n"},
		{name: "unparseable", file: "broken.go", body: hdr + "func {\n", broken: true},
		{name: "unreadable directory x.go", file: "x.go", dir: true, broken: true},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			dir := t.TempDir()
			selectPath := filepath.Join(dir, "select.go")
			if err := os.WriteFile(selectPath, src, 0o644); err != nil {
				t.Fatalf("write select.go: %v", err)
			}
			extra := filepath.Join(dir, r.file)
			if r.dir {
				if err := os.Mkdir(extra, 0o755); err != nil {
					t.Fatalf("mkdir: %v", err)
				}
			} else if err := os.WriteFile(extra, []byte(r.body), 0o644); err != nil {
				t.Fatalf("write extra: %v", err)
			}
			if !r.broken && strings.HasSuffix(r.file, ".go") {
				if _, err := parser.ParseFile(token.NewFileSet(), extra, r.body, parser.ParseComments); err != nil {
					t.Fatalf("row fixture must parse: %v", err)
				}
			}
			pin := checkPathspecPin(selectPath)
			if pin.PathspecOK || pin.PrefixOK {
				t.Fatalf("package closure must clear both flags, got %+v", pin)
			}
			if !r.broken && (!pin.SelectFound || !pin.GuardFound) {
				t.Fatalf("row must fail on closure only (SelectFound && GuardFound), got %+v", pin)
			}
		})
	}
	t.Run("select.go only accepted", func(t *testing.T) {
		dir := t.TempDir()
		selectPath := filepath.Join(dir, "select.go")
		if err := os.WriteFile(selectPath, src, 0o644); err != nil {
			t.Fatalf("write select.go: %v", err)
		}
		if pin := checkPathspecPin(selectPath); !pin.OK() {
			t.Fatalf("select.go-only fixture must be accepted, got %+v", pin)
		}
	})
}

// TestUniverseDeclNames_CoverCanonicalTexts is AC-A3d.2 (033.007-T): every
// identifier token of the §A-CANON texts that types.Universe resolves must
// be in pin.go's authored universeDeclNames list. Toolchain-stable: a new
// builtin the frozen code does not mention cannot redden it.
func TestUniverseDeclNames_CoverCanonicalTexts(t *testing.T) {
	fset := token.NewFileSet()
	var s scanner.Scanner
	s.Init(fset.AddFile("canon", -1, len(canonicalDecls)), []byte(canonicalDecls), nil, 0)
	seen := 0
	for {
		_, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		if tok != token.IDENT || types.Universe.Lookup(lit) == nil {
			continue
		}
		seen++
		if !universeDeclNames[lit] {
			t.Errorf("universe identifier %q used by the canonical texts is missing from universeDeclNames", lit)
		}
	}
	if seen == 0 {
		t.Fatal("canonical texts must mention at least one universe identifier (non-vacuity)")
	}
}

// TestCheckPathspecPin_GitRunnerIsolation_Rejected (U5 scenario 2;
// D7BF9F74): weakening DefaultGitRunner's git-environment isolation is a
// pin violation. Each row mutates one literal of a whole-package copy and
// must leave SelectFound, GuardFound and PrefixOK true while clearing
// PathspecOK, so a parse-error all-false result cannot pass (G2-2).
// Row (i) is the gitRunnerEnv freeze; row (ii) is characterization, already
// rejected through the frozen DefaultGitRunner text.
func TestCheckPathspecPin_GitRunnerIsolation_Rejected(t *testing.T) {
	rows := []struct{ name, old, new string }{
		{"i_git_config_nosystem_weakened", `"GIT_CONFIG_NOSYSTEM=1"`, `"GIT_CONFIG_NOSYSTEM=0"`},
		{"ii_cmd_env_assignment_dropped", "cmd.Env = gitRunnerEnv(os.Environ())", "_ = gitRunnerEnv(os.Environ())"},
	}
	for _, r := range rows {
		t.Run(r.name, func(t *testing.T) {
			pin := checkPathspecPin(writeMutatedCopy(t, r.old, r.new))
			if !pin.SelectFound || !pin.GuardFound || !pin.PrefixOK || pin.PathspecOK {
				t.Fatalf("want SelectFound, GuardFound and PrefixOK true with PathspecOK false, got %+v", pin)
			}
		})
	}
}

// TestGitShowToplevelIgnoresGitEnv (U6 scenario 1, AC-1; D7BF9F74): a
// decoy GIT_DIR/GIT_WORK_TREE in the gate's environment must not redirect
// the pin's repository-root resolution. The expected value is derived by
// gitShowToplevel itself before the decoy env is set, which sidesteps 8.3
// short-path, symlinked-tmp and case differences between t.TempDir() and
// git's own output (G2-3).
func TestGitShowToplevelIgnoresGitEnv(t *testing.T) {
	repoR, repoD := t.TempDir(), t.TempDir()
	sub := filepath.Join(repoR, "sub")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	fixtureGit(t, repoR, "init", "-q")
	fixtureGit(t, repoD, "init", "-q")
	want, err := gitShowToplevel(sub)
	if err != nil {
		t.Fatalf("gitShowToplevel(R/sub) without decoy env: %v", err)
	}
	decoy, err := gitShowToplevel(repoD)
	if err != nil {
		t.Fatalf("gitShowToplevel(D) without decoy env: %v", err)
	}
	if want == decoy {
		t.Fatalf("fixture error: R and D resolve to the same root %q", want)
	}
	t.Setenv("GIT_DIR", filepath.Join(repoD, ".git"))
	t.Setenv("GIT_WORK_TREE", repoD)
	got, err := gitShowToplevel(sub)
	if err != nil || got != want {
		t.Fatalf("gitShowToplevel(R/sub) with decoy GIT_DIR/GIT_WORK_TREE = (%q, %v), want (%q, nil); decoy root is %q", got, err, want, decoy)
	}
}

// TestGitShowToplevelRefusesRelativeGit (U6 scenario 2, AC-2): when PATH
// lookup resolves git to a relative path, gitShowToplevel refuses it before
// launching anything. t.Chdir is required because gitShowToplevel sets no
// cmd.Dir.
func TestGitShowToplevelRefusesRelativeGit(t *testing.T) {
	fakeRoot := t.TempDir()
	dir, marker := setupRelativeFakeGit(t, fakeRoot+"\n")
	got, err := gitShowToplevel(dir)
	if err == nil || !strings.Contains(err.Error(), "non-absolute") {
		t.Fatalf("gitShowToplevel = (%q, %v), want an error containing \"non-absolute\"", got, err)
	}
	assertFakeGitNotRun(t, marker)
}

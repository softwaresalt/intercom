package retiredarch

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"sort"
	"strings"
	"testing"
)

func TestShouldScanRepoPath_Table(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"config.toml.example", true},
		{"cmd/gatecheck/main.go", true},
		{"cmd/gatecheck/main_test.go", true}, // cmd/ includes tests (asymmetry)
		{"cmd/gatecheck/testdata/fixture.go", true},
		{"cmd/gatecheck/readme.md", false}, // cmd/ still requires .go
		{"internal/foo/bar.go", true},
		{"internal/foo/bar_test.go", false},     // internal/ excludes tests
		{"internal/foo/testdata/bar.go", false}, // internal/ excludes testdata
		{"internal/foo/bar.py", false},
		{"scripts/testdata/retiredgo/x.go", false},
		{"other/file.go", false},
		{"", false},
	}
	for _, c := range cases {
		c := c
		t.Run(c.path, func(t *testing.T) {
			if got := shouldScanRepoPath(c.path); got != c.want {
				t.Fatalf("shouldScanRepoPath(%q) = %v, want %v", c.path, got, c.want)
			}
		})
	}
}

// TestShouldScanRepoPath_CmdIncludesInternalExcludes_Asymmetry explicitly
// names and asserts the AC-required asymmetry: cmd/**'s predicate does
// NOT exclude *_test.go or testdata/ paths, unlike internal/**'s.
func TestShouldScanRepoPath_CmdIncludesInternalExcludes_Asymmetry(t *testing.T) {
	if !shouldScanRepoPath("cmd/x/y_test.go") {
		t.Fatalf("cmd/**'s predicate must INCLUDE *_test.go paths")
	}
	if !shouldScanRepoPath("cmd/x/testdata/z.go") {
		t.Fatalf("cmd/**'s predicate must INCLUDE testdata/ paths")
	}
	if shouldScanRepoPath("internal/x/y_test.go") {
		t.Fatalf("internal/**'s predicate must EXCLUDE *_test.go paths")
	}
	if shouldScanRepoPath("internal/x/testdata/z.go") {
		t.Fatalf("internal/**'s predicate must EXCLUDE testdata/ paths")
	}
}

func TestEngineForPath_Table(t *testing.T) {
	cases := []struct {
		path string
		want string
	}{
		{"config.toml.example", "toml"},
		{"/abs/path/config.toml.example", "toml"},
		{"foo.go", "go"},
		{"foo.toml", "toml"},
		{"foo.md", ""},
		{"foo", ""},
	}
	for _, c := range cases {
		c := c
		t.Run(c.path, func(t *testing.T) {
			if got := engineForPath(c.path); got != c.want {
				t.Fatalf("engineForPath(%q) = %q, want %q", c.path, got, c.want)
			}
		})
	}
}

func TestScanPath_UnmappedExtension_SyntheticFinding(t *testing.T) {
	findings := scanPath("foo.md")
	if len(findings) != 1 {
		t.Fatalf("scanPath(unmapped) = %v, want exactly one synthetic finding", findings)
	}
	if !strings.Contains(findings[0], "no scan engine registered") {
		t.Fatalf("scanPath(unmapped) = %q, want a fail-closed synthetic finding message", findings[0])
	}
}

func TestScanPath_ReadError_SyntheticFinding_NeverPanics(t *testing.T) {
	// A .go path that does not exist: scanGo's pysem.ReadText call fails,
	// and scanPath must convert that into a synthetic finding, never
	// panic or propagate a Go error.
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("scanPath must never panic on a read error, got: %v", r)
		}
	}()
	findings := scanPath("does-not-exist-xyz.go")
	if len(findings) != 1 {
		t.Fatalf("scanPath(missing file) = %v, want exactly one synthetic finding", findings)
	}
}

func TestSelectRepoPaths_GitError_ReturnsError(t *testing.T) {
	fakeErr := errors.New("git exploded")
	git := func(root string, pathspecs ...string) ([]byte, error) {
		return nil, fakeErr
	}
	if _, err := selectRepoPaths("/does/not/matter", git); !errors.Is(err, fakeErr) {
		t.Fatalf("selectRepoPaths git error = %v, want wrapping of %v", err, fakeErr)
	}
}

func TestSelectRepoPaths_FiltersAndSorts(t *testing.T) {
	git := func(root string, pathspecs ...string) ([]byte, error) {
		if len(pathspecs) != 3 || pathspecs[0] != "config.toml.example" || pathspecs[1] != "cmd/**" || pathspecs[2] != "internal/**" {
			return nil, fmt.Errorf("unexpected pathspecs: %v", pathspecs)
		}
		return []byte("internal/z/a.go\ninternal/z/a_test.go\ncmd/x/y_test.go\nconfig.toml.example\ninternal/a/b.go\n"), nil
	}
	got, err := selectRepoPaths("/root", git)
	if err != nil {
		t.Fatalf("selectRepoPaths: %v", err)
	}
	want := []string{"cmd/x/y_test.go", "config.toml.example", "internal/a/b.go", "internal/z/a.go"}
	if len(got) != len(want) {
		t.Fatalf("selectRepoPaths = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("selectRepoPaths[%d] = %q, want %q (full: %v)", i, got[i], want[i], got)
		}
	}
}

// TestSelectRepoPaths_MatchesGolden_LiveTree is M2-T7's evidence AC. The
// parent-commit Python selected_path_set golden was captured ONCE, at
// evidence-collection time (see m2.md §13.2), and is compared here only
// as non-fatal diagnostic information (t.Logf) -- never a hard equality
// assertion -- because freezing an exact file list into a durable
// `go test ./...` test would fail on every future legitimate addition or
// removal of a source file under cmd/**/internal/**, even when selection
// behavior itself remains perfectly correct (Copilot review round 6;
// this test previously hard-failed on drift). The real pass/fail
// assertions below are dynamically-derived structural invariants plus an
// independent second enumeration of the internal/ subset via
// filepath.WalkDir over the real filesystem (deliberately NOT git
// ls-files, and NOT shouldScanRepoPath/expectedInternalRepoPaths, both of
// which already share the same git-based implementation), so this test
// still catches a real selection regression without depending on a
// point-in-time snapshot.
func TestSelectRepoPaths_MatchesGolden_LiveTree(t *testing.T) {
	root := repoRoot(t)
	g := loadGolden(t)
	got, err := selectRepoPaths(root, DefaultGitRunner)
	if err != nil {
		t.Fatalf("selectRepoPaths(live tree): %v", err)
	}

	// Evidence-only, non-fatal: log (never fail) any drift from the
	// frozen parent-commit golden snapshot recorded in m2.md.
	want := g.RepoModeEvidence.SelectedPathSet
	switch {
	case len(got) != len(want):
		t.Logf("live selected set has %d paths, golden snapshot (m2.md) has %d -- informational only: got=%v want=%v", len(got), len(want), got, want)
	default:
		for i := range want {
			if got[i] != want[i] {
				t.Logf("live selected set[%d] = %q differs from golden snapshot (m2.md) %q -- informational only", i, got[i], want[i])
				break
			}
		}
	}

	if len(got) == 0 {
		t.Fatal("selectRepoPaths(live tree) returned zero paths")
	}
	for i := 1; i < len(got); i++ {
		if got[i-1] >= got[i] {
			t.Fatalf("live selected set not strictly sorted/unique at index %d: %q >= %q", i, got[i-1], got[i])
		}
	}
	for _, p := range got {
		if strings.HasPrefix(p, "tools/") {
			t.Fatalf("selected path %q must never start with tools/ (gatecheck's own source is never self-scanned)", p)
		}
		switch {
		case p == "config.toml.example":
		case strings.HasPrefix(p, "cmd/"):
			if !strings.HasSuffix(p, ".go") {
				t.Fatalf("selected cmd/ path %q must end in .go", p)
			}
		case strings.HasPrefix(p, "internal/"):
			if strings.Contains(p, "/testdata/") || strings.HasSuffix(p, "_test.go") {
				t.Fatalf("selected internal/ path %q must exclude testdata/_test.go", p)
			}
			if !strings.HasSuffix(p, ".go") {
				t.Fatalf("selected internal/ path %q must end in .go", p)
			}
		default:
			t.Fatalf("selected path %q is outside the cmd/**, internal/**, config.toml.example pathspecs", p)
		}
	}

	// Independent second enumeration: walk the real filesystem (not git
	// ls-files) for the internal/ subset and require it to match exactly.
	wantInternal := walkInternalGoFiles(t, root)
	var gotInternal []string
	for _, p := range got {
		if strings.HasPrefix(p, "internal/") {
			gotInternal = append(gotInternal, filepath.ToSlash(p))
		}
	}
	if len(gotInternal) != len(wantInternal) {
		t.Fatalf("selectRepoPaths internal/ subset has %d entries, filesystem walk found %d: got=%v want=%v", len(gotInternal), len(wantInternal), gotInternal, wantInternal)
	}
	for i := range wantInternal {
		if gotInternal[i] != wantInternal[i] {
			t.Fatalf("selectRepoPaths internal/ subset[%d] = %q, filesystem walk = %q", i, gotInternal[i], wantInternal[i])
		}
	}
}

// walkInternalGoFiles independently enumerates internal/**.go files
// (excluding _test.go and testdata/) directly from the filesystem via
// filepath.WalkDir, deliberately not reusing
// shouldScanRepoPath/expectedInternalRepoPaths, so
// TestSelectRepoPaths_MatchesGolden_LiveTree has a genuinely independent
// oracle rather than a second copy of the same selection-logic filter.
//
// Copilot review round 8 (PR #83, HEAD 660d271) found that a raw
// filepath.WalkDir enumeration also picks up untracked files (e.g. a
// local scratch/editor temp file under internal/), which selectRepoPaths
// never selects (it enumerates only git-tracked paths via `git
// ls-files`) -- so an untracked stray .go file would make this test fail
// even though selection behavior is correct. Fixed by cross-checking
// each walked candidate directly against `git ls-files --error-unmatch`
// (a raw exec.Command invocation, deliberately NOT going through
// selectRepoPaths/DefaultGitRunner/the GitRunner type under test) and
// dropping any untracked file before comparison -- this keeps the
// oracle's filesystem-presence check and its selection-logic filtering
// both independent of the code under test, while now also independently
// verifying trackedness instead of assuming every walked file is tracked.
func walkInternalGoFiles(t *testing.T, root string) []string {
	t.Helper()
	var out []string
	internalRoot := filepath.Join(root, "internal")
	err := filepath.WalkDir(internalRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasSuffix(path, "_test.go") || !strings.HasSuffix(path, ".go") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		relSlash := filepath.ToSlash(rel)
		if !isGitTrackedFile(t, root, relSlash) {
			return nil
		}
		out = append(out, relSlash)
		return nil
	})
	if err != nil {
		t.Fatalf("walkInternalGoFiles: %v", err)
	}
	sort.Strings(out)
	return out
}

// isGitTrackedFile reports whether relPath is tracked in git, via a raw
// `git ls-files --error-unmatch` invocation independent of
// selectRepoPaths/DefaultGitRunner/the GitRunner type under test (round
// 8 fix, see walkInternalGoFiles doc comment above).
func isGitTrackedFile(t *testing.T, root, relPath string) bool {
	t.Helper()
	cmd := exec.Command("git", "ls-files", "--error-unmatch", "--", relPath)
	cmd.Dir = root
	if err := cmd.Run(); err != nil {
		return false
	}
	return true
}

// TestMain doubles as a natively launchable fake git (U4/U6 helper; plan
// S2-1, AS-2/G-1). copyFakeGit copies this test binary to dir/git(.exe);
// when that copy is launched git-style (os.Args[1] is "ls-files" or "-C")
// with RETIREDARCH_FAKE_GIT=1, it writes the RETIREDARCH_FAKE_GIT_MARKER
// file, prints RETIREDARCH_FAKE_GIT_OUT and exits 0 before any test flag
// parsing. A stray exported RETIREDARCH_FAKE_GIT=1 seen with test-style
// arguments exits 2 instead, so it can never turn the package suite into a
// silent no-op.
func TestMain(m *testing.M) {
	if os.Getenv("RETIREDARCH_FAKE_GIT") == "1" {
		if len(os.Args) > 1 && (os.Args[1] == "ls-files" || os.Args[1] == "-C") {
			if marker := os.Getenv("RETIREDARCH_FAKE_GIT_MARKER"); marker != "" {
				if err := os.WriteFile(marker, []byte("fake git ran\n"), 0o644); err != nil {
					fmt.Fprintf(os.Stderr, "fake git: write marker: %v\n", err)
					os.Exit(3)
				}
			}
			fmt.Print(os.Getenv("RETIREDARCH_FAKE_GIT_OUT"))
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, "retiredarch tests: RETIREDARCH_FAKE_GIT=1 is set but the binary was invoked test-style; refusing to run the suite as a silent no-op")
		os.Exit(2)
	}
	os.Exit(m.Run())
}

// copyFakeGit copies the running test binary to dir/git (git.exe on
// Windows) with mode 0o755 and returns the copy's path (plan G2-1).
func copyFakeGit(t *testing.T, dir string) string {
	t.Helper()
	self, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable: %v", err)
	}
	data, err := os.ReadFile(self)
	if err != nil {
		t.Fatalf("read test binary: %v", err)
	}
	name := "git"
	if runtime.GOOS == "windows" {
		name = "git.exe"
	}
	dst := filepath.Join(dir, name)
	if err := os.WriteFile(dst, data, 0o755); err != nil {
		t.Fatalf("write fake git: %v", err)
	}
	return dst
}

// setupRelativeFakeGit puts a fake git in a fresh directory, makes it the
// working directory, and sets PATH="." with GODEBUG=execerrdot=0 so that
// exec.LookPath("git") resolves to a RELATIVE path with a nil error (plan
// G-3; precondition asserted). It returns the directory and the absolute
// marker path the fake writes if it is ever launched.
func setupRelativeFakeGit(t *testing.T, out string) (dir, marker string) {
	t.Helper()
	dir = t.TempDir()
	copyFakeGit(t, dir)
	marker = filepath.Join(t.TempDir(), "fake-git-ran")
	t.Chdir(dir)
	t.Setenv("PATH", ".")
	t.Setenv("GODEBUG", "execerrdot=0")
	t.Setenv("RETIREDARCH_FAKE_GIT", "1")
	t.Setenv("RETIREDARCH_FAKE_GIT_MARKER", marker)
	t.Setenv("RETIREDARCH_FAKE_GIT_OUT", out)
	if p, err := exec.LookPath("git"); err != nil || filepath.IsAbs(p) {
		t.Fatalf("precondition: exec.LookPath(\"git\") = (%q, %v), want a relative path and a nil error", p, err)
	}
	return dir, marker
}

// assertFakeGitNotRun fails the test if the fake git wrote its marker.
func assertFakeGitNotRun(t *testing.T, marker string) {
	t.Helper()
	if _, err := os.Stat(marker); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("the relative fake git was launched (marker stat err = %v)", err)
	}
}

// fixtureGit runs real git for test-fixture construction only, with every
// ambient GIT_* variable removed and global/system config isolated, so the
// fixture is identical on every machine. It deliberately does not use the
// code under test.
func fixtureGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	var env []string
	for _, kv := range os.Environ() {
		name, _, _ := strings.Cut(kv, "=")
		if strings.HasPrefix(strings.ToUpper(name), "GIT_") {
			continue
		}
		env = append(env, kv)
	}
	env = append(env, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull)
	base := []string{"-c", "user.name=retiredarch-test", "-c", "user.email=retiredarch-test@example.invalid", "-c", "commit.gpgsign=false"}
	cmd := exec.Command("git", append(base, args...)...)
	cmd.Dir = dir
	cmd.Env = env
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("fixture git %v: %v: %s", args, err, out)
	}
}

// TestDefaultGitRunnerIgnoresGitEnv (U4 scenario 2, AC-2): GIT_*
// variables in the gate's own environment must not change which files
// DefaultGitRunner selects. The baseline records 4537B2F6's known
// quoted-path skip: internal/<e-acute>.go is printed quoted by git and is
// therefore not selected. When 4537B2F6 lands it must update this
// baseline and replace vectors (c)-(e).
func TestDefaultGitRunnerIgnoresGitEnv(t *testing.T) {
	root := t.TempDir()
	for rel, body := range map[string]string{
		"cmd/x/main.go":      "package main\n\nfunc main() {}\n",
		"internal/\u00e9.go": "package internal\n",
	} {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	fixtureGit(t, root, "init", "-q")
	fixtureGit(t, root, "add", "-A")
	fixtureGit(t, root, "commit", "-q", "--no-verify", "-m", "fixture")

	want := []string{"cmd/x/main.go"}
	baseline, err := selectRepoPaths(root, DefaultGitRunner)
	if err != nil || !slices.Equal(baseline, want) {
		t.Fatalf("baseline selection = (%q, %v), want (%q, nil)", baseline, err, want)
	}

	globalCfg := filepath.Join(t.TempDir(), "gitconfig")
	if err := os.WriteFile(globalCfg, []byte("[core]\n\tquotePath = false\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	vectors := []struct {
		name string
		env  map[string]string
	}{
		{"a_GIT_INDEX_FILE_nonexistent", map[string]string{"GIT_INDEX_FILE": filepath.Join(t.TempDir(), "missing-index")}},
		{"b_GIT_LITERAL_PATHSPECS", map[string]string{"GIT_LITERAL_PATHSPECS": "1"}},
		{"c_GIT_CONFIG_COUNT_quotePath", map[string]string{"GIT_CONFIG_COUNT": "1", "GIT_CONFIG_KEY_0": "core.quotePath", "GIT_CONFIG_VALUE_0": "false"}},
		{"d_GIT_CONFIG_PARAMETERS_quotePath", map[string]string{"GIT_CONFIG_PARAMETERS": "'core.quotepath'='false'"}},
		{"e_GIT_CONFIG_GLOBAL_quotePath", map[string]string{"GIT_CONFIG_GLOBAL": globalCfg}},
	}
	for _, v := range vectors {
		t.Run(v.name, func(t *testing.T) {
			// Set only now, after the fixture repo is fully built (G2-7).
			for k, val := range v.env {
				t.Setenv(k, val)
			}
			got, err := selectRepoPaths(root, DefaultGitRunner)
			if err != nil || !slices.Equal(got, want) {
				t.Fatalf("selection with %v = (%q, %v), want (%q, nil)", v.env, got, err, want)
			}
		})
	}
}

// TestDefaultGitRunnerRefusesRelativeGit (U4 scenario 3, AC-3): when PATH
// lookup resolves git to a relative path (here PATH="." with
// GODEBUG=execerrdot=0), DefaultGitRunner refuses it before launching
// anything.
func TestDefaultGitRunnerRefusesRelativeGit(t *testing.T) {
	dir, marker := setupRelativeFakeGit(t, "cmd/x/main.go\n")
	out, err := DefaultGitRunner(dir, "cmd/")
	if err == nil || !strings.Contains(err.Error(), "non-absolute") {
		t.Fatalf("DefaultGitRunner = (%q, %v), want an error containing \"non-absolute\"", out, err)
	}
	assertFakeGitNotRun(t, marker)
}

// TestDefaultGitRunnerMissingGitIsNotFound (U4 AC-4, characterization;
// SB2-2): with no git reachable at all, cmd.Err is reported first, so the
// error is exec.ErrNotFound and never the non-absolute refusal.
func TestDefaultGitRunnerMissingGitIsNotFound(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("PATH", "")
	_, err := DefaultGitRunner(dir, "cmd/")
	if !errors.Is(err, exec.ErrNotFound) || strings.Contains(fmt.Sprint(err), "non-absolute") {
		t.Fatalf("DefaultGitRunner err = %v, want exec.ErrNotFound without \"non-absolute\"", err)
	}
}

// TestGitRunnerEnv (U4 scenario 1, AC-1): the pure environment filter.
func TestGitRunnerEnv(t *testing.T) {
	in := []string{
		"PATH=/usr/bin",
		"git_dir=/decoy/.git",
		"Git_Index_File=/decoy/index",
		"GIT_CEILING_DIRECTORIES=/ceiling",
		`=C:=C:\x`,
		"GITX=1",
		"GIT_CONFIG_GLOBAL=/decoy/gitconfig",
		"HOME=/home/u",
		"GIT_CONFIG_NOSYSTEM=0",
	}
	want := []string{
		"PATH=/usr/bin",
		"GIT_CEILING_DIRECTORIES=/ceiling",
		`=C:=C:\x`,
		"GITX=1",
		"HOME=/home/u",
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_CONFIG_GLOBAL=" + os.DevNull,
		"GIT_CONFIG_SYSTEM=" + os.DevNull,
	}
	if got := gitRunnerEnv(in); !slices.Equal(got, want) {
		t.Fatalf("gitRunnerEnv =\n%q\nwant\n%q", got, want)
	}
	if got := gitRunnerEnv(nil); !slices.Equal(got, want[len(want)-3:]) {
		t.Fatalf("gitRunnerEnv(nil) = %q, want only the three appended entries", got)
	}
}

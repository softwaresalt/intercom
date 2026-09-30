package retiredarch

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
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
// filepath.WalkDir, deliberately not reusing git ls-files or
// shouldScanRepoPath/expectedInternalRepoPaths, so
// TestSelectRepoPaths_MatchesGolden_LiveTree has a genuinely independent
// oracle rather than a second copy of the same git-based filter.
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
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		t.Fatalf("walkInternalGoFiles: %v", err)
	}
	sort.Strings(out)
	return out
}

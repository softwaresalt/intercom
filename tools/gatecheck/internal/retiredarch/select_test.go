package retiredarch

import (
	"errors"
	"fmt"
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

// TestSelectRepoPaths_MatchesGolden_LiveTree is M2-T7's evidence AC: the
// live selected set (real git, real tree) must equal the parent-commit
// Python selected_path_set golden exactly.
func TestSelectRepoPaths_MatchesGolden_LiveTree(t *testing.T) {
	root := repoRoot(t)
	g := loadGolden(t)
	got, err := selectRepoPaths(root, DefaultGitRunner)
	if err != nil {
		t.Fatalf("selectRepoPaths(live tree): %v", err)
	}
	want := g.RepoModeEvidence.SelectedPathSet
	if len(got) != len(want) {
		t.Fatalf("live selected set has %d paths, golden has %d: got=%v want=%v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("live selected set[%d] = %q, golden = %q (full got=%v want=%v)", i, got[i], want[i], got, want)
		}
	}
}

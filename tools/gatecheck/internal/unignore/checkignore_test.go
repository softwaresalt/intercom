package unignore

import (
	"testing"
)

func TestIsIgnored(t *testing.T) {
	git := newIsolatedGitRunner(t)
	root := t.TempDir()
	scratchDir, err := makeScratchGitignore(root, "repo", "foo\n", git)
	if err != nil {
		t.Fatalf("makeScratchGitignore: %v", err)
	}

	ignored, err := isIgnored(git, scratchDir, "foo")
	if err != nil {
		t.Fatalf("isIgnored(foo): %v", err)
	}
	if !ignored {
		t.Fatal("expected foo to be ignored")
	}

	ignored, err = isIgnored(git, scratchDir, "bar")
	if err != nil {
		t.Fatalf("isIgnored(bar): %v", err)
	}
	if ignored {
		t.Fatal("expected bar to NOT be ignored")
	}
}

func TestIsIgnoredBatch(t *testing.T) {
	git := newIsolatedGitRunner(t)
	root := t.TempDir()
	// "*.secret" matches BOTH keep.secret and other.secret at the SAME
	// directory level (not nested under an excluded directory, where git
	// cannot re-include a path at all); "!keep.secret" re-includes
	// exactly one of them -- a negation, which must NOT be
	// misclassified as "ignored" by the verbose-line parser.
	scratchDir, err := makeScratchGitignore(root, "repo", "*.secret\n!keep.secret\n", git)
	if err != nil {
		t.Fatalf("makeScratchGitignore: %v", err)
	}

	paths := []string{"keep.secret", "other.secret", "clean.txt"}
	got, err := isIgnoredBatch(git, scratchDir, paths)
	if err != nil {
		t.Fatalf("isIgnoredBatch: %v", err)
	}
	want := []bool{false, true, false}
	if len(got) != len(want) {
		t.Fatalf("got %d results, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("paths[%d]=%q: got %v, want %v", i, paths[i], got[i], want[i])
		}
	}
}

func TestIsIgnoredBatch_EmptyInput(t *testing.T) {
	git := newIsolatedGitRunner(t)
	root := t.TempDir()
	scratchDir, err := makeScratchGitignore(root, "repo", "foo\n", git)
	if err != nil {
		t.Fatalf("makeScratchGitignore: %v", err)
	}
	got, err := isIgnoredBatch(git, scratchDir, nil)
	if err != nil {
		t.Fatalf("isIgnoredBatch(nil): %v", err)
	}
	if got != nil {
		t.Fatalf("expected nil result for empty input, got %v", got)
	}
}

func TestIsIgnored_NonBinaryExitIsError(t *testing.T) {
	// A fake GitRunner that returns an exit code outside {0, 1} must
	// surface as a hard error, fail-closed, never silently treated as
	// either "ignored" or "not ignored".
	fake := func(dir string, stdin []byte, args ...string) ([]byte, []byte, error) {
		return nil, []byte("fatal: something went very wrong"), &fakeExitError{code: 128}
	}
	_, err := isIgnored(fake, "unused", "some/path")
	if err == nil {
		t.Fatal("expected an error for a non-{0,1} exit code")
	}
}

func TestIsIgnoredBatch_LineCountMismatchIsHardError(t *testing.T) {
	// A fake GitRunner that returns fewer lines than input paths must be a
	// hard error (cannot safely align results), matching the Python
	// engine's own equivalent fail-closed check.
	fake := func(dir string, stdin []byte, args ...string) ([]byte, []byte, error) {
		return []byte("::\tone\n"), nil, nil
	}
	_, err := isIgnoredBatch(fake, "unused", []string{"one", "two", "three"})
	if err == nil {
		t.Fatal("expected a line-count mismatch error")
	}
}

func TestParseCheckIgnoreVerboseLine(t *testing.T) {
	cases := []struct {
		name string
		line string
		want bool
	}{
		{"non-matching", "::\tbaz", false},
		{"positive-match", ".gitignore:1:foo\tfoo", true},
		{"negation-match", ".gitignore:2:!foo/bar.secret\tfoo/bar.secret", false},
		{"nested-pattern-with-colon-in-path", ".gitignore:3:build/\tbuild/output:artifact", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := parseCheckIgnoreVerboseLine(tc.line); got != tc.want {
				t.Errorf("parseCheckIgnoreVerboseLine(%q) = %v, want %v", tc.line, got, tc.want)
			}
		})
	}
}

// fakeExitError is a minimal error whose presence lets exitCode's default
// branch (an error that is not an *exec.ExitError) be exercised
// deterministically without needing to shell out to a real failing git
// invocation.
type fakeExitError struct {
	code int
}

func (e *fakeExitError) Error() string { return "fake exit error" }

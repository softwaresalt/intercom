package unignore

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestSplitNulTerminated(t *testing.T) {
	got := splitNulTerminated([]byte("a\x00b\x00c\x00"))
	want := []string{"a", "b", "c"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("index %d: got %q, want %q", i, got[i], want[i])
		}
	}
}

func TestSplitNulTerminated_Empty(t *testing.T) {
	if got := splitNulTerminated(nil); len(got) != 0 {
		t.Fatalf("expected an empty result for nil input, got %v", got)
	}
	if got := splitNulTerminated([]byte{}); len(got) != 0 {
		t.Fatalf("expected an empty result for empty input, got %v", got)
	}
}

func TestRunDenylistCheck_AllIgnored(t *testing.T) {
	git := newIsolatedGitRunner(t)
	repoDir := t.TempDir()
	scratchRoot := t.TempDir()

	var b strings.Builder
	for _, entry := range Denylist {
		b.WriteString(entry)
		b.WriteString("\n")
	}
	if err := writeTestGitignore(repoDir, b.String()); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}

	evaluated, failures, err := runDenylistCheck(git, repoDir, scratchRoot, "HEAD")
	if err != nil {
		t.Fatalf("runDenylistCheck: %v", err)
	}
	if evaluated != len(Denylist) {
		t.Fatalf("evaluated = %d, want %d", evaluated, len(Denylist))
	}
	if len(failures) != 0 {
		t.Fatalf("expected no failures, got %v", failures)
	}
}

func TestRunDenylistCheck_SomeMissing(t *testing.T) {
	git := newIsolatedGitRunner(t)
	repoDir := t.TempDir()
	scratchRoot := t.TempDir()

	// Cover every denylist entry EXCEPT the first one.
	var b strings.Builder
	for _, entry := range Denylist[1:] {
		b.WriteString(entry)
		b.WriteString("\n")
	}
	if err := writeTestGitignore(repoDir, b.String()); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}

	evaluated, failures, err := runDenylistCheck(git, repoDir, scratchRoot, "HEAD")
	if err != nil {
		t.Fatalf("runDenylistCheck: %v", err)
	}
	if evaluated != len(Denylist) {
		t.Fatalf("evaluated = %d, want %d", evaluated, len(Denylist))
	}
	if len(failures) != 1 || failures[0] != Denylist[0] {
		t.Fatalf("failures = %v, want [%s]", failures, Denylist[0])
	}
}

// The following three scenarios exercise runDifferentialCheck directly
// (rather than through the full Run/self-test path), covering the exact
// accept/reject cases the plan's M3-T5 AC requires: an existing ignored
// file un-ignored by a negation is rejected; a negation for a
// nonexistent path is accepted; and a negation that ALSO tracks the file
// in the same change is still rejected.

func TestRunDifferentialCheck_RejectsExistingFileUnignored(t *testing.T) {
	git := newIsolatedGitRunner(t)
	tmpRoot := t.TempDir()
	repoDir, baseRef, err := makeScenarioRepo(
		git, tmpRoot, "reject-existing",
		"foo/bar.secret\n", "foo/bar.secret\n!foo/bar.secret\n",
		[]string{"foo/bar.secret"}, nil,
	)
	if err != nil {
		t.Fatalf("makeScenarioRepo: %v", err)
	}
	scratchRoot := filepath.Join(tmpRoot, "scratch1")
	evaluated, failures, err := runDifferentialCheck(git, repoDir, scratchRoot, baseRef, "HEAD")
	if err != nil {
		t.Fatalf("runDifferentialCheck: %v", err)
	}
	if evaluated == 0 {
		t.Fatal("expected at least one candidate evaluated")
	}
	if len(failures) != 1 || failures[0] != "foo/bar.secret" {
		t.Fatalf("failures = %v, want [foo/bar.secret]", failures)
	}
}

func TestRunDifferentialCheck_AcceptsNonexistentNegation(t *testing.T) {
	git := newIsolatedGitRunner(t)
	tmpRoot := t.TempDir()
	repoDir, baseRef, err := makeScenarioRepo(
		git, tmpRoot, "accept-nonexistent",
		"foo/\n", "foo/\n!foo/keep.txt\n",
		nil, nil,
	)
	if err != nil {
		t.Fatalf("makeScenarioRepo: %v", err)
	}
	scratchRoot := filepath.Join(tmpRoot, "scratch2")
	_, failures, err := runDifferentialCheck(git, repoDir, scratchRoot, baseRef, "HEAD")
	if err != nil {
		t.Fatalf("runDifferentialCheck: %v", err)
	}
	if len(failures) != 0 {
		t.Fatalf("expected no failures, got %v", failures)
	}
}

func TestRunDifferentialCheck_RejectsTrackedUnignoredInSameChange(t *testing.T) {
	git := newIsolatedGitRunner(t)
	tmpRoot := t.TempDir()
	repoDir, baseRef, err := makeScenarioRepo(
		git, tmpRoot, "reject-tracked",
		"foo/bar.secret\n", "foo/bar.secret\n!foo/bar.secret\n",
		nil, []string{"foo/bar.secret"},
	)
	if err != nil {
		t.Fatalf("makeScenarioRepo: %v", err)
	}
	scratchRoot := filepath.Join(tmpRoot, "scratch3")
	_, failures, err := runDifferentialCheck(git, repoDir, scratchRoot, baseRef, "HEAD")
	if err != nil {
		t.Fatalf("runDifferentialCheck: %v", err)
	}
	if len(failures) != 1 || failures[0] != "foo/bar.secret" {
		t.Fatalf("failures = %v, want [foo/bar.secret]", failures)
	}
}

func TestRunDifferentialCheck_NoCandidates(t *testing.T) {
	git := newIsolatedGitRunner(t)
	tmpRoot := t.TempDir()
	// The .gitignore itself changes at head (so there IS a real commit
	// diff to evaluate), but the change only ADDS an unrelated pattern --
	// no candidate path is ignored at base, so the differential check has
	// nothing to evaluate.
	repoDir, baseRef, err := makeScenarioRepo(
		git, tmpRoot, "no-candidates",
		"foo\n", "foo\nbar\n",
		nil, nil,
	)
	if err != nil {
		t.Fatalf("makeScenarioRepo: %v", err)
	}
	scratchRoot := filepath.Join(tmpRoot, "scratch4")
	evaluated, failures, err := runDifferentialCheck(git, repoDir, scratchRoot, baseRef, "HEAD")
	if err != nil {
		t.Fatalf("runDifferentialCheck: %v", err)
	}
	if evaluated != 0 || len(failures) != 0 {
		t.Fatalf("expected a vacuous clean result, got evaluated=%d failures=%v", evaluated, failures)
	}
}

func TestRunLsFilesCheck_TrackedIgnoredFileIsReported(t *testing.T) {
	// Built manually (not via makeScenarioRepo) because real git refuses
	// a plain `git add` of a path that is ALREADY ignored at the time of
	// the add -- exactly the "should never happen on a healthy checkout"
	// state run_ls_files_check exists to catch requires committing the
	// file BEFORE the ignoring pattern exists, then adding the pattern in
	// a later, separate commit.
	git := newIsolatedGitRunner(t)
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		if _, stderr, err := git(dir, nil, args...); err != nil {
			t.Fatalf("git %v failed: %v: %s", args, err, stderr)
		}
	}
	run("init", "-q", "-b", "main")
	if err := os.WriteFile(filepath.Join(dir, "tracked.secret"), []byte("fixture content\n"), 0o644); err != nil {
		t.Fatalf("write tracked.secret: %v", err)
	}
	run("add", "tracked.secret")
	run("commit", "-q", "-m", "track before ignoring")

	if err := os.WriteFile(filepath.Join(dir, ".gitignore"), []byte("tracked.secret\n"), 0o644); err != nil {
		t.Fatalf("write .gitignore: %v", err)
	}
	run("add", ".gitignore")
	run("commit", "-q", "-m", "ignore after tracking")

	got, err := runLsFilesCheck(git, dir)
	if err != nil {
		t.Fatalf("runLsFilesCheck: %v", err)
	}
	sort.Strings(got)
	if len(got) != 1 || got[0] != "tracked.secret" {
		t.Fatalf("got %v, want [tracked.secret]", got)
	}
}

func TestRunLsFilesCheck_CleanRepoIsEmpty(t *testing.T) {
	git := newIsolatedGitRunner(t)
	tmpRoot := t.TempDir()
	repoDir, _, err := makeScenarioRepo(
		git, tmpRoot, "clean",
		"foo\n", "foo\nbaz\n",
		nil, nil,
	)
	if err != nil {
		t.Fatalf("makeScenarioRepo: %v", err)
	}

	got, err := runLsFilesCheck(git, repoDir)
	if err != nil {
		t.Fatalf("runLsFilesCheck: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty result, got %v", got)
	}
}

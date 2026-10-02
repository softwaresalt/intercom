package retiredarch

import (
	"errors"
	"strings"
	"testing"
)

var errGitFailed = errors.New("git exploded (test double)")

func TestRunRepoSelectionSelfTest_AssertionNamesAndOrder_MatchGolden(t *testing.T) {
	g := loadGolden(t)
	// The selection self-test assertions are the LAST 10 entries of the
	// self_test stream's assertion list (everything before them belongs
	// to the fixture self-test suites).
	all := g.StreamCaptures["self_test"].Assertions
	if len(all) < 10 {
		t.Fatalf("golden self_test assertions too short: %d", len(all))
	}
	wantTail := all[len(all)-10:]

	root := repoRoot(t)
	out, err := runRepoSelectionSelfTest(root, DefaultGitRunner)
	if err != nil {
		t.Fatalf("runRepoSelectionSelfTest: %v", err)
	}
	if out.failed {
		t.Fatalf("runRepoSelectionSelfTest reported a failure on the live tree:\n%s", out.stdout.String())
	}
	if len(out.assertions) != len(wantTail) {
		t.Fatalf("assertion count mismatch: got %d, want %d\ngot: %+v", len(out.assertions), len(wantTail), out.assertions)
	}
	for i, want := range wantTail {
		got := out.assertions[i]
		if got.Name != want.Name {
			t.Errorf("assertion %d name = %q, want %q", i, got.Name, want.Name)
		}
		if got.Outcome != want.Outcome {
			t.Errorf("assertion %d (%s) outcome = %s, want %s", i, got.Name, got.Outcome, want.Outcome)
		}
	}
}

// TestRunRepoSelectionSelfTest_NoToolsPathSelected is a NEW, test-only
// assertion (M2-T9's AC): it does not appear in the M4-deleted retired_arch module's own
// self-test stdout at all -- it lives purely here, guarding the invariant
// that the ported Go package's own home (tools/) is never itself
// accidentally swept into the selection set select_repo_paths() returns
// (tools/gatecheck's exclusion from every gate pathspec is a load-bearing
// C-1 property; this test independently re-checks it from the selection
// self-test's own real output rather than only from select_test.go).
func TestRunRepoSelectionSelfTest_NoToolsPathSelected(t *testing.T) {
	root := repoRoot(t)
	relPaths, err := selectRepoPaths(root, DefaultGitRunner)
	if err != nil {
		t.Fatalf("selectRepoPaths: %v", err)
	}
	for _, p := range relPaths {
		if strings.HasPrefix(p, "tools/") {
			t.Fatalf("selected path %q starts with tools/ -- gatecheck's own source must never be self-scanned", p)
		}
	}
}

func TestExpectedInternalRepoPaths_NonEmptyAndSorted(t *testing.T) {
	root := repoRoot(t)
	got, err := expectedInternalRepoPaths(root, DefaultGitRunner)
	if err != nil {
		t.Fatalf("expectedInternalRepoPaths: %v", err)
	}
	if len(got) == 0 {
		t.Fatal("expectedInternalRepoPaths returned zero paths on the live tree")
	}
	for _, p := range got {
		if !strings.HasPrefix(p, "internal/") {
			t.Errorf("expectedInternalRepoPaths returned a non internal/ path: %q", p)
		}
		if strings.Contains(p, "/testdata/") || strings.HasSuffix(p, "_test.go") {
			t.Errorf("expectedInternalRepoPaths returned an excluded path: %q", p)
		}
	}
	for i := 1; i < len(got); i++ {
		if got[i-1] > got[i] {
			t.Fatalf("expectedInternalRepoPaths not sorted at index %d: %q > %q", i, got[i-1], got[i])
		}
	}
}

func TestExpectedInternalRepoPaths_GitError_Propagates(t *testing.T) {
	failing := func(root string, pathspecs ...string) ([]byte, error) {
		return nil, errGitFailed
	}
	if _, err := expectedInternalRepoPaths("/nonexistent", failing); err == nil {
		t.Fatal("expected an error when the injected GitRunner fails")
	}
}

func TestRunRepoSelectionSelfTest_GitError_Propagates(t *testing.T) {
	failing := func(root string, pathspecs ...string) ([]byte, error) {
		return nil, errGitFailed
	}
	if _, err := runRepoSelectionSelfTest("/nonexistent", failing); err == nil {
		t.Fatal("expected an error when the injected GitRunner fails")
	}
}

// TestCmdProbesHold_DropsNonTestCmdGo_False is the AC-A4.1 pure-predicate
// test: a predicate that keeps only cmd/ test and testdata files (dropping
// ordinary non-test cmd/ Go files) must fail the probe set, while the real
// shouldScanRepoPath must pass it.
func TestCmdProbesHold_DropsNonTestCmdGo_False(t *testing.T) {
	dropping := func(p string) bool {
		return strings.HasSuffix(p, "_test.go") || strings.Contains(p, "/testdata/")
	}
	if cmdProbesHold(dropping) {
		t.Fatalf("cmdProbesHold(predicate dropping non-test cmd/ Go files) = true, want false")
	}
	if !cmdProbesHold(shouldScanRepoPath) {
		t.Fatalf("cmdProbesHold(shouldScanRepoPath) = false, want true")
	}
}

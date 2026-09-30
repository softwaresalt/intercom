package retiredarch

import (
	"os"
	"strings"
	"testing"
)

// splitAssertionLines extracts every "PASS name: ..." / "FAIL name: ..."
// line from stdout into an assertionRecord, in order, mirroring the
// M2-T1 generator's own {name, outcome} extraction so a Go-produced
// stream can be compared against the golden's assertions list without
// depending on live corpus-count text (C-5).
func splitAssertionLines(stdout string) []assertionRecord {
	var out []assertionRecord
	for _, line := range strings.Split(stdout, "\n") {
		var outcome, rest string
		switch {
		case strings.HasPrefix(line, "PASS "):
			outcome, rest = "PASS", strings.TrimPrefix(line, "PASS ")
		case strings.HasPrefix(line, "FAIL "):
			outcome, rest = "FAIL", strings.TrimPrefix(line, "FAIL ")
		default:
			continue
		}
		idx := strings.Index(rest, ": ")
		name := rest
		if idx >= 0 {
			name = rest[:idx]
		}
		out = append(out, assertionRecord{Name: name, Outcome: outcome})
	}
	return out
}

// nonAssertionLines returns every stdout line that is NOT a "PASS "/"FAIL "
// assertion line (the notice line and the trailing success banner), so
// those CAN be compared byte-for-byte against the golden (they never
// embed a live corpus count).
func nonAssertionLines(stdout string) []string {
	var out []string
	for _, line := range strings.Split(stdout, "\n") {
		if strings.HasPrefix(line, "PASS ") || strings.HasPrefix(line, "FAIL ") {
			continue
		}
		out = append(out, line)
	}
	return out
}

func assertStreamMatchesGolden(t *testing.T, label string, gotStdout string, gotCode int, want streamRow) {
	t.Helper()
	if gotCode != want.ExitCode {
		t.Errorf("%s: exit code = %d, want %d", label, gotCode, want.ExitCode)
	}

	gotAssertions := splitAssertionLines(gotStdout)
	if len(gotAssertions) != len(want.Assertions) {
		t.Fatalf("%s: assertion count = %d, want %d\ngot: %+v", label, len(gotAssertions), len(want.Assertions), gotAssertions)
	}
	for i, wantA := range want.Assertions {
		gotA := gotAssertions[i]
		if gotA.Name != wantA.Name {
			t.Errorf("%s: assertion %d name = %q, want %q", label, i, gotA.Name, wantA.Name)
		}
		if gotA.Outcome != wantA.Outcome {
			t.Errorf("%s: assertion %d (%s) outcome = %s, want %s", label, i, gotA.Name, gotA.Outcome, wantA.Outcome)
		}
	}

	gotNonAssertion := nonAssertionLines(gotStdout)
	wantNonAssertion := nonAssertionLines(want.Stdout)
	if len(gotNonAssertion) != len(wantNonAssertion) {
		t.Fatalf("%s: non-assertion stdout line count = %d, want %d\ngot: %q\nwant: %q", label, len(gotNonAssertion), len(wantNonAssertion), gotNonAssertion, wantNonAssertion)
	}
	for i := range wantNonAssertion {
		if gotNonAssertion[i] != wantNonAssertion[i] {
			t.Errorf("%s: non-assertion stdout line %d = %q, want %q", label, i, gotNonAssertion[i], wantNonAssertion[i])
		}
	}
}

func TestRun_SelfTest_MatchesGolden(t *testing.T) {
	g := loadGolden(t)
	root := repoRoot(t)
	var stdout, stderr strings.Builder
	code := Run("--self-test", root, DefaultGitRunner, &stdout, &stderr)
	want := g.StreamCaptures["self_test"]
	assertStreamMatchesGolden(t, "--self-test", stdout.String(), code, want)
	if stderr.Len() != 0 {
		t.Errorf("--self-test: unexpected stderr on the live (green) tree: %q", stderr.String())
	}
}

func TestRun_SelfTestIntegrity_MatchesGolden(t *testing.T) {
	g := loadGolden(t)
	root := repoRoot(t)
	var stdout, stderr strings.Builder
	code := Run("--self-test-integrity", root, DefaultGitRunner, &stdout, &stderr)
	want := g.StreamCaptures["self_test_integrity"]
	assertStreamMatchesGolden(t, "--self-test-integrity", stdout.String(), code, want)
	if stderr.Len() != 0 {
		t.Errorf("--self-test-integrity: unexpected stderr on the live (green) tree: %q", stderr.String())
	}
}

func TestRun_UnknownMode_MatchesGolden(t *testing.T) {
	g := loadGolden(t)
	root := repoRoot(t)
	var stdout, stderr strings.Builder
	code := Run("--bogus", root, DefaultGitRunner, &stdout, &stderr)
	want := g.StreamCaptures["unknown_mode"]
	if code != want.ExitCode {
		t.Errorf("unknown mode: exit code = %d, want %d", code, want.ExitCode)
	}
	if stdout.String() != want.Stdout {
		t.Errorf("unknown mode: stdout = %q, want %q", stdout.String(), want.Stdout)
	}
	if stderr.String() != want.Stderr {
		t.Errorf("unknown mode: stderr = %q, want %q", stderr.String(), want.Stderr)
	}
}

func TestRun_Repo_MatchesGolden(t *testing.T) {
	g := loadGolden(t)
	root := repoRoot(t)
	var stdout, stderr strings.Builder
	code := Run("", root, DefaultGitRunner, &stdout, &stderr)
	want := g.StreamCaptures["repo"]
	if code != want.ExitCode {
		t.Errorf("repo mode: exit code = %d, want %d (live tree finding count may legitimately differ if new tokens were introduced)", code, want.ExitCode)
	}
	if stdout.String() != want.Stdout {
		t.Errorf("repo mode: stdout = %q, want %q", stdout.String(), want.Stdout)
	}
}

// TestRun_RepoScan_GitError_FailsClosed proves a git failure during the
// repo-mode scan surfaces as a fail-closed ::error:: line and exit 1,
// never a panic (ED-2).
func TestRun_RepoScan_GitError_FailsClosed(t *testing.T) {
	failing := func(root string, pathspecs ...string) ([]byte, error) { return nil, errGitFailed }
	var stdout, stderr strings.Builder
	code := Run("", t.TempDir(), failing, &stdout, &stderr)
	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.HasPrefix(stderr.String(), "::error::") {
		t.Fatalf("stderr = %q, want an ::error:: prefixed line", stderr.String())
	}
}

// TestRun_UnknownMode_NoNoticeLine confirms the unknown-mode branch never
// prints a "::notice::" line (bash's own case default branch never
// invoked the Python engine at all, so there was never a notice for it).
func TestRun_UnknownMode_NoNoticeLine(t *testing.T) {
	var stdout, stderr strings.Builder
	code := Run("--nope", t.TempDir(), DefaultGitRunner, &stdout, &stderr)
	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if strings.Contains(stdout.String(), "::notice::") {
		t.Fatalf("unexpected notice line in unknown-mode stdout: %q", stdout.String())
	}
}

func TestLoadFixtureManifest_MissingFile_Errors(t *testing.T) {
	if _, err := loadFixtureManifest("/does/not/exist.json"); err == nil {
		t.Fatal("expected an error for a missing manifest file")
	}
}

func TestLoadFixtureManifest_InvalidShape_Errors(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/manifest.json"
	if err := os.WriteFile(path, []byte("[1,2,3]"), 0o644); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}
	if _, err := loadFixtureManifest(path); err == nil {
		t.Fatal("expected an error for a non-object manifest shape")
	}
}

// TestConfigTomlExample_DualEngineAgreement_LiveCorpus is M2-T5's live
// corpus check (ED-6 disclosure): config.toml.example is tracked, selected
// by selectRepoPaths, and dispatched to the TOML engine in production, but
// it is not one of the scripts/testdata/retired-*.toml fixtures the golden
// captures per-fixture, so its dual-engine agreement is checked directly
// here as an explicit live-tree ED-6 corpus assertion (see m2.md M2-T5
// evidence: no tracked .toml file, including this one, uses TOML
// 1.1-only syntax).
func TestConfigTomlExample_DualEngineAgreement_LiveCorpus(t *testing.T) {
	root := repoRoot(t)
	path := root + "/config.toml.example"
	primary := scanTomlPrimary(path)
	fallback, err := scanTomlFallback(path)
	if err != nil {
		t.Fatalf("scanTomlFallback(config.toml.example): %v", err)
	}
	if (len(primary) > 0) != (len(fallback) > 0) {
		t.Fatalf("dual-engine disagreement on config.toml.example: primary=%v fallback=%v", primary, fallback)
	}
}

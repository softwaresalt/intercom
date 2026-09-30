package retiredarch

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
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

// TestLoadFixtureManifest_TopLevelNull_Errors is a Copilot-review fix
// (PR #83, HEAD c37b3a3 review round 1): encoding/json decodes a
// top-level JSON `null` into a nil map[string]interface{} with NO error
// -- a well-known Go json quirk -- but Python's `isinstance(data, dict)`
// guard rejects `None` (the result of `json.loads("null")`) as not a
// dict, confirmed against the real CPython 3.14 interpreter
// (`isinstance(None, dict)` -> `False`, and
// `raise SystemExit(f"invalid fixture manifest shape: {p}")` fires).
// Without the raw == nil check, a top-level-null manifest would silently
// behave as an empty manifest instead of failing closed with the shape
// error, diverging from Python's actual behavior.
func TestLoadFixtureManifest_TopLevelNull_Errors(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")
	if err := os.WriteFile(path, []byte("null"), 0o644); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}
	manifest, err := loadFixtureManifest(path)
	if err == nil {
		t.Fatalf("expected a manifestShapeError for a top-level JSON null manifest, got manifest=%v, err=nil", manifest)
	}
	var shapeErr *manifestShapeError
	if !errors.As(err, &shapeErr) {
		t.Fatalf("loadFixtureManifest(null) error = %v (%T), want a *manifestShapeError", err, err)
	}
}

// TestLoadFixtureManifest_NonStringValue_Accepted is F-2's core fix
// assertion (adversarial review
// docs/closure/2026-09-30-gate-engine-m2-retired-arch-adversarial-
// review.md): unlike the prior map[string]string-based port, a manifest
// whose top level IS a genuine JSON object but contains a non-string
// value (e.g. an int) must NOT be hard-rejected at load time -- Python's
// load_fixture_manifest only requires isinstance(data, dict) and defers
// any per-value type mismatch to the per-fixture "unknown expectation"
// failure branch (confirmed against the real CPython 3.14 interpreter;
// see this task's evidence-file addendum for the exact transient-script
// invocation and captured output).
func TestLoadFixtureManifest_NonStringValue_Accepted(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")
	if err := os.WriteFile(path, []byte(`{"foo.toml": 1, "bar.toml": "accept"}`), 0o644); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}
	manifest, err := loadFixtureManifest(path)
	if err != nil {
		t.Fatalf("loadFixtureManifest with a non-string value must not hard-reject (F-2), got error: %v", err)
	}
	exp := lookupExpectation(manifest, "foo.toml")
	if exp.isString || exp.isNone {
		t.Fatalf("lookupExpectation(foo.toml) = %+v, want a non-string/non-none 'other' value", exp)
	}
	// Confirmed against real CPython: f"unknown expectation {1!r} in x" ==
	// "unknown expectation 1 in x" (Python's repr(1) == "1").
	if got, want := exp.repr(), "1"; got != want {
		t.Fatalf("manifestExpectation.repr() = %q, want %q", got, want)
	}
}

// TestManifestExpectation_AbsentVsExplicitEmptyString is F-2's second
// fix assertion: an absent manifest key and an explicit JSON empty string
// value must stay distinguishable (Python's manifest.get(name) already
// distinguishes them: None vs ”), unlike the prior port's map[string]string
// lookup, which conflated both into the same "" sentinel.
func TestManifestExpectation_AbsentVsExplicitEmptyString(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")
	if err := os.WriteFile(path, []byte(`{"present-empty.toml": ""}`), 0o644); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}
	manifest, err := loadFixtureManifest(path)
	if err != nil {
		t.Fatalf("loadFixtureManifest: %v", err)
	}

	present := lookupExpectation(manifest, "present-empty.toml")
	if !present.isString || present.str != "" || present.isNone {
		t.Fatalf("lookupExpectation(present-empty.toml) = %+v, want isString=true str=\"\" isNone=false", present)
	}
	absent := lookupExpectation(manifest, "absent.toml")
	if !absent.isNone || absent.isString {
		t.Fatalf("lookupExpectation(absent.toml) = %+v, want isNone=true", absent)
	}
	if present == absent {
		t.Fatalf("an explicit empty-string expectation must not compare equal to an absent key")
	}
}

// TestRunFixtureSelfTest_ManifestShapeError_BareMessage is F-6's fix
// assertion: loadFixtureManifest's genuinely-non-object hard-reject must
// surface as Python's own SystemExit(str)-style BARE message + "\n" on
// stderr (no "::error::" prefix -- that prefix is reserved for ED-2's
// traceback-replacement class), confirmed against the real CPython 3.14
// interpreter's f"invalid fixture manifest shape: {path.as_posix()}"
// (see this task's evidence-file addendum).
func TestRunFixtureSelfTest_ManifestShapeError_BareMessage(t *testing.T) {
	root := t.TempDir()
	manifestDir := filepath.Join(root, "scripts", "testdata")
	if err := os.MkdirAll(manifestDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	manifestPath := filepath.Join(manifestDir, "retired-manifest.json")
	if err := os.WriteFile(manifestPath, []byte("[1,2,3]"), 0o644); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}

	res := runFixtureSelfTest(root)
	if res.Code != 1 {
		t.Fatalf("Code = %d, want 1", res.Code)
	}
	if strings.Contains(res.Stderr, "::error::") {
		t.Fatalf("manifest-shape reject must NOT use the \"::error::\" prefix (F-6); got %q", res.Stderr)
	}
	want := fmt.Sprintf("invalid fixture manifest shape: %s\n", filepath.ToSlash(manifestPath))
	if res.Stderr != want {
		t.Fatalf("Stderr = %q, want %q", res.Stderr, want)
	}
}

// TestRunFixtureSelfTest_UnknownExpectation_NonStringValue is F-2's
// end-to-end fix assertion: a manifest entry whose value is a non-string
// JSON scalar must reach the SAME "unknown expectation %s in %s" failure
// text Python's run_fixture_self_test produces for that branch (confirmed
// against the real CPython 3.14 interpreter), rather than being rejected
// at manifest-load time.
func TestRunFixtureSelfTest_UnknownExpectation_NonStringValue(t *testing.T) {
	root := t.TempDir()
	testdataDir := filepath.Join(root, "scripts", "testdata")
	if err := os.MkdirAll(testdataDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	// "toml" suite: one fixture, manifest expectation is the JSON int 1
	// (never a legal 'accept'/'reject' string).
	if err := os.WriteFile(filepath.Join(testdataDir, "retired-crafted.toml"), []byte("x = 1\n"), 0o644); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testdataDir, "retired-manifest.json"), []byte(`{"retired-crafted.toml": 1}`), 0o644); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}
	// "go" / "go-differential" suites: minimal valid empty manifests so
	// loadFixtureManifest never hard-errors on those (missing-file errors
	// return early and would prevent the "toml" suite's own failures from
	// ever reaching the final aggregated stderr dump).
	if err := os.WriteFile(filepath.Join(testdataDir, "retiredgo-manifest.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testdataDir, "retiredgo-differential-manifest.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}

	res := runFixtureSelfTest(root)
	if res.Code != 1 {
		t.Fatalf("Code = %d, want 1 (an all-1-int manifest can never satisfy accept/reject coverage)", res.Code)
	}
	for _, label := range []string{"tomllib", "fallback"} {
		want := fmt.Sprintf("retired-crafted.toml [%s]: unknown expectation 1 in retired-manifest.json", label)
		if !strings.Contains(res.Stderr, want) {
			t.Fatalf("Stderr = %q, want it to contain %q", res.Stderr, want)
		}
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

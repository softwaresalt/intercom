package retiredarch

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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

// TestRun_SelfTestIntegrity_ForeignCWD_MatchesGolden ports the retired
// test_integrity_mode_is_cwd_independent: with the process working directory
// set to an unrelated temp dir, integrity mode must still resolve everything
// from the explicit root and produce the golden green stream.
func TestRun_SelfTestIntegrity_ForeignCWD_MatchesGolden(t *testing.T) {
	g := loadGolden(t)
	root := repoRoot(t)
	t.Chdir(t.TempDir())
	var stdout, stderr strings.Builder
	code := Run("--self-test-integrity", root, DefaultGitRunner, &stdout, &stderr)
	want := g.StreamCaptures["self_test_integrity"]
	assertStreamMatchesGolden(t, "--self-test-integrity (foreign cwd)", stdout.String(), code, want)
	for _, line := range []string{"PASS selection pathspec pin (AC-6/AG-1)", "PASS selection cmd/ coverage (AG-5/D4)"} {
		if !strings.Contains(stdout.String(), line) {
			t.Errorf("foreign cwd: stdout missing %q", line)
		}
	}
	if strings.Contains(stdout.String(), "FAIL") {
		t.Errorf("foreign cwd: stdout contains FAIL: %q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Errorf("foreign cwd: unexpected stderr: %q", stderr.String())
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
	manifest, err := loadFixtureManifest(path)
	if err == nil {
		t.Fatal("expected an error for a non-object manifest shape")
	}
	// A syntactically-valid JSON array is exactly the shape-mismatch
	// class Python's `isinstance(data, dict)` guard rejects via its own
	// SystemExit -- it must stay a *manifestShapeError (bare message, no
	// "::error::" prefix -- see F-6 and
	// TestRunFixtureSelfTest_ManifestShapeError_BareMessage), not the
	// plain-decode-error class added by the round-2 Copilot-review fix
	// below.
	var shapeErr *manifestShapeError
	if !errors.As(err, &shapeErr) {
		t.Fatalf("loadFixtureManifest([1,2,3]) error = %v (%T), want a *manifestShapeError, got manifest=%v", err, err, manifest)
	}
}

// TestLoadFixtureManifest_MalformedJSON_PlainError is a Copilot-review
// fix (PR #83, HEAD 2c5fe8e review round 2): a genuine JSON syntax error
// (truncated/malformed input) must NOT surface as a *manifestShapeError.
// Python's load_fixture_manifest calls `json.loads` unguarded -- a
// malformed document raises an UNCAUGHT json.JSONDecodeError, which is
// NOT load_fixture_manifest's own `raise SystemExit(f"invalid fixture
// manifest shape: ...")`; it propagates as an ordinary Python exception
// (a traceback + exit 1), exactly ED-2's "today a Python traceback with
// exit 1; after, a one-line ::error:: message with exit 1" delta. A
// prior port revision decoded straight into map[string]interface{} and
// funneled BOTH failure classes (genuine syntax errors and valid-but-
// wrong-shape values) into the same *manifestShapeError bucket, which
// would have made a malformed-JSON manifest print the bare
// "invalid fixture manifest shape" message instead of an "::error::"-
// prefixed one -- diverging from Python's actual traceback-class
// behavior.
func TestLoadFixtureManifest_MalformedJSON_PlainError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")
	if err := os.WriteFile(path, []byte(`{"a": `), 0o644); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}
	manifest, err := loadFixtureManifest(path)
	if err == nil {
		t.Fatalf("expected a decode error for malformed JSON, got manifest=%v, err=nil", manifest)
	}
	var shapeErr *manifestShapeError
	if errors.As(err, &shapeErr) {
		t.Fatalf("loadFixtureManifest(malformed JSON) error = %v (%T), want a plain decode error (ED-2 class), NOT a *manifestShapeError", err, err)
	}
}

// TestLoadFixtureManifest_TrailingData_PlainError is the companion to
// TestLoadFixtureManifest_MalformedJSON_PlainError: trailing
// non-whitespace content after the top-level JSON value is Python's
// "Extra data" json.JSONDecodeError -- the SAME uncaught-exception class
// as a syntax error, not a shape mismatch -- so it must also stay a
// plain error routed to the caller's "::error::" (ED-2) branch, matching
// ED-8's explicit "trailing data after the top-level value stays a SKIP
// [via the decode-error path]" rule for the sibling merge-strategy JSON
// evaluator.
func TestLoadFixtureManifest_TrailingData_PlainError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")
	if err := os.WriteFile(path, []byte(`{"a": 1} garbage`), 0o644); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}
	manifest, err := loadFixtureManifest(path)
	if err == nil {
		t.Fatalf("expected a decode error for trailing data after the top-level value, got manifest=%v, err=nil", manifest)
	}
	var shapeErr *manifestShapeError
	if errors.As(err, &shapeErr) {
		t.Fatalf("loadFixtureManifest(trailing data) error = %v (%T), want a plain decode error (ED-2 class), NOT a *manifestShapeError", err, err)
	}
}

// TestLoadFixtureManifest_TrailingClosingDelimiter_PlainError is a
// Copilot-review fix (PR #83, HEAD 0c3aa94 review round 3):
// json.Decoder.More() is NOT a top-level EOF check -- it is intended for
// array/object element iteration and reports false whenever the NEXT
// byte is a closing delimiter (`]` or `}`), so a prior fix revision
// (round 2, using dec.More()) silently ACCEPTED a manifest with trailing
// `]`/`}` bytes, even though Python's `json.loads` rejects both as
// "Extra data" (confirmed against the real CPython 3.14 interpreter:
// `json.loads('{"a":"accept"}]')` and `json.loads('{"a":"accept"}}')`
// both raise JSONDecodeError). The fix requires a second Decode call and
// asserting io.EOF -- exactly ED-8's prescribed pattern for the sibling
// merge-strategy JSON evaluator's identical rule.
func TestLoadFixtureManifest_TrailingClosingDelimiter_PlainError(t *testing.T) {
	for _, tc := range []string{
		`{"a":"accept"}]`,
		`{"a":"accept"}}`,
	} {
		dir := t.TempDir()
		path := filepath.Join(dir, "manifest.json")
		if err := os.WriteFile(path, []byte(tc), 0o644); err != nil {
			t.Fatalf("os.WriteFile: %v", err)
		}
		manifest, err := loadFixtureManifest(path)
		if err == nil {
			t.Fatalf("loadFixtureManifest(%q): expected a decode error for trailing closing-delimiter data, got manifest=%v, err=nil", tc, manifest)
		}
		var shapeErr *manifestShapeError
		if errors.As(err, &shapeErr) {
			t.Fatalf("loadFixtureManifest(%q) error = %v (%T), want a plain decode error (ED-2 class), NOT a *manifestShapeError", tc, err, err)
		}
	}
}

// TestLoadFixtureManifest_InvalidUTF8_PlainError is a Copilot-review fix
// (PR #83, HEAD 3006994, review round 7): Python's
// Path.read_text(encoding='utf-8') decodes in STRICT mode by default, so
// a manifest file containing invalid UTF-8 byte sequences raises an
// uncaught UnicodeDecodeError -- the same ED-2 "decode error" class as a
// JSON syntax error. Go's string(data) conversion performs no such
// validation, and encoding/json's decoder silently substitutes U+FFFD
// for invalid sequences inside JSON string literals instead of failing,
// so without an explicit utf8.Valid check a corrupted manifest could
// continue through the self-test instead of taking the required ED-2
// error path.
func TestLoadFixtureManifest_InvalidUTF8_PlainError(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.json")
	// 0xFF is never valid as any byte of a well-formed UTF-8 sequence.
	invalid := []byte(`{"a": "`)
	invalid = append(invalid, 0xFF)
	invalid = append(invalid, []byte(`"}`)...)
	if err := os.WriteFile(path, invalid, 0o644); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}
	manifest, err := loadFixtureManifest(path)
	if err == nil {
		t.Fatalf("expected a decode error for invalid UTF-8, got manifest=%v, err=nil", manifest)
	}
	var shapeErr *manifestShapeError
	if errors.As(err, &shapeErr) {
		t.Fatalf("loadFixtureManifest(invalid UTF-8) error = %v (%T), want a plain decode error (ED-2 class), NOT a *manifestShapeError", err, err)
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

// TestRunFixtureSelfTest_RejectOnly_NullExpectationBreaksExemption is a
// Copilot round-12 finding (PR #83): Python's reject-only check is
// `expectations <= {'reject'}`, where `expectations` is a SET that
// includes a `None` member whenever any discovered fixture's manifest
// entry is absent or explicit JSON null (`manifest.get(name)`'s
// sentinel). `{'reject', None} <= {'reject'}` is False, so Python FAILS
// the "suite is a named reject-only exemption" assertion for a
// go-differential manifest that mixes a 'reject' fixture with a
// null/missing one. The prior port silently dropped None entries from
// BOTH expectationStrs and otherPresent, so it reported PASS for this
// exact mixed manifest -- a parity break this test pins down.
func TestRunFixtureSelfTest_RejectOnly_NullExpectationBreaksExemption(t *testing.T) {
	root := t.TempDir()
	testdataDir := filepath.Join(root, "scripts", "testdata")
	differentialDir := filepath.Join(testdataDir, "retiredgo-differential")
	if err := os.MkdirAll(differentialDir, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(differentialDir, "fixture_a.go"), []byte("package retiredgo\n"), 0o644); err != nil {
		t.Fatalf("os.WriteFile fixture_a.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(differentialDir, "fixture_b.go"), []byte("package retiredgo\n"), 0o644); err != nil {
		t.Fatalf("os.WriteFile fixture_b.go: %v", err)
	}
	// fixture_a.go: 'reject' (a string member). fixture_b.go: explicit
	// JSON null -- present as a manifest KEY (so the separate "missing
	// from manifest" check does not also fire) but collapsing to
	// isNone, exactly like an absent key would.
	manifest := `{"fixture_a.go": "reject", "fixture_b.go": null}`
	if err := os.WriteFile(filepath.Join(testdataDir, "retiredgo-differential-manifest.json"), []byte(manifest), 0o644); err != nil {
		t.Fatalf("os.WriteFile retiredgo-differential-manifest.json: %v", err)
	}
	// "toml" / "go" suites: minimal valid empty manifests so
	// loadFixtureManifest never hard-errors on those.
	if err := os.WriteFile(filepath.Join(testdataDir, "retired-manifest.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("os.WriteFile retired-manifest.json: %v", err)
	}
	if err := os.WriteFile(filepath.Join(testdataDir, "retiredgo-manifest.json"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("os.WriteFile retiredgo-manifest.json: %v", err)
	}

	res := runFixtureSelfTest(root)
	if res.Code != 1 {
		t.Fatalf("Code = %d, want 1 (a null-mixed manifest must break the reject-only exemption, matching Python's `expectations <= {'reject'}` set check)", res.Code)
	}
	want := "suite go-differential is declared reject-only but manifest expectations are ['reject']"
	if !strings.Contains(res.Stderr, want) {
		t.Fatalf("Stderr = %q, want it to contain %q", res.Stderr, want)
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

// u3StubGit returns a GitRunner that ignores the pathspecs and reports
// exactly the given repo-relative paths, so a U3 containment scenario
// needs no real git index (plan U3: containment is index-independent).
func u3StubGit(paths ...string) GitRunner {
	return func(root string, pathspecs ...string) ([]byte, error) {
		return []byte(strings.Join(paths, "\n") + "\n"), nil
	}
}

// u3CleanGo is a retired-token-free Go source body: read through a link,
// it scans clean, which is exactly why the parent's Code 0 is the red.
const u3CleanGo = "package x\n\nfunc Clean() {}\n"

// u3Junction creates a Windows directory junction via plain argv (no
// embedded quotes; plan G-2). It fails the test if mklink fails.
func u3Junction(t *testing.T, link, target string) {
	t.Helper()
	out, err := exec.Command("cmd", "/c", "mklink", "/J", link, target).CombinedOutput()
	if err != nil {
		t.Fatalf("mklink /J %s %s: %v: %s", link, target, err, out)
	}
}

// u3MustContain asserts a repo scan failed closed with the U3 synthetic
// finding for rel.
func u3MustContain(t *testing.T, res Result, root, rel string) {
	t.Helper()
	want := filepath.ToSlash(filepath.Join(root, filepath.FromSlash(rel))) + ": not a contained regular file: "
	if res.Code != 1 {
		t.Fatalf("Code = %d, want 1 (stderr=%q)", res.Code, res.Stderr)
	}
	if !strings.Contains(res.Stderr, want) || !strings.Contains(res.Stderr, "(fail-closed synthetic finding)") {
		t.Fatalf("stderr = %q, want the U3 finding prefix %q", res.Stderr, want)
	}
}

// TestRunRepoScan_FinalComponentSymlink_FailsClosed (U3 AC-1): a selected
// internal/x.go that is a symlink to a clean file outside root must not be
// read through.
func TestRunRepoScan_FinalComponentSymlink_FailsClosed(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	target := filepath.Join(outside, "x.go")
	if err := os.WriteFile(target, []byte(u3CleanGo), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "internal"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(target, filepath.Join(root, "internal", "x.go")); err != nil {
		if runtime.GOOS == "windows" {
			t.Skipf("os.Symlink unavailable on Windows without privilege: %v", err)
		}
		t.Fatalf("os.Symlink: %v", err)
	}
	u3MustContain(t, runRepoScan(root, u3StubGit("internal/x.go")), root, "internal/x.go")
}

// TestRunRepoScan_FinalComponentJunction_FailsClosed (U3 AC-2, Windows):
// a selected internal/x.go that is a junction to an outside directory
// yields the U3 text instead of the parent's directory read-error text.
func TestRunRepoScan_FinalComponentJunction_FailsClosed(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("directory junctions are Windows-only")
	}
	root, outside := t.TempDir(), t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "internal"), 0o755); err != nil {
		t.Fatal(err)
	}
	u3Junction(t, filepath.Join(root, "internal", "x.go"), outside)
	u3MustContain(t, runRepoScan(root, u3StubGit("internal/x.go")), root, "internal/x.go")
}

// TestRunRepoScan_IntermediateLink_FailsClosed (U3 AC-3): an intermediate
// internal/d that is a symlink (Unix) or a junction (Windows) to an
// outside directory holding clean y.go must not be traversed.
func TestRunRepoScan_IntermediateLink_FailsClosed(t *testing.T) {
	root, outside := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(outside, "y.go"), []byte(u3CleanGo), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(root, "internal"), 0o755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(root, "internal", "d")
	if runtime.GOOS == "windows" {
		u3Junction(t, link, outside)
	} else if err := os.Symlink(outside, link); err != nil {
		t.Fatalf("os.Symlink: %v", err)
	}
	u3MustContain(t, runRepoScan(root, u3StubGit("internal/d/y.go")), root, "internal/d/y.go")
}

// TestContainedRegularFile_RejectsMalformedComponents (U3 AC-4): these
// rel shapes never pass shouldScanRepoPath, so they are direct unit rows.
func TestContainedRegularFile_RejectsMalformedComponents(t *testing.T) {
	root := t.TempDir()
	rows := []string{"a//b.go", "./a.go", "../x.go", "a/", ""}
	if runtime.GOOS == "windows" {
		rows = append(rows, `a\b.go`, "c:x.go", "a/b:c.go")
	}
	for _, rel := range rows {
		if ok, reason := containedRegularFile(root, rel); ok || reason == "" {
			t.Errorf("containedRegularFile(%q) = (%v, %q), want (false, non-empty reason)", rel, ok, reason)
		}
	}
}

// TestContainedRegularFile_AcceptsRegularAndMissing pins the two ok=true
// outcomes: a real regular file under real directories, and a path that
// verifiably does not exist (fs.ErrNotExist falls through to scanPath).
func TestContainedRegularFile_AcceptsRegularAndMissing(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "internal", "d"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "internal", "d", "y.go"), []byte(u3CleanGo), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, rel := range []string{"internal/d/y.go", "internal/gone.go", "internal/nodir/gone.go"} {
		if ok, reason := containedRegularFile(root, rel); !ok {
			t.Errorf("containedRegularFile(%q) = (false, %q), want ok", rel, reason)
		}
	}
}

// TestRunRepoScan_DeletedFile_KeepsReadErrorText (U3 AC-4
// characterization): a selected-but-deleted file keeps the existing
// scanPath read-error text byte-for-byte (SB-3, ED-11 unchanged).
func TestRunRepoScan_DeletedFile_KeepsReadErrorText(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "internal"), 0o755); err != nil {
		t.Fatal(err)
	}
	res := runRepoScan(root, u3StubGit("internal/gone.go"))
	want := strings.Join(scanPath(filepath.Join(root, "internal", "gone.go")), "\n") + "\n"
	if res.Code != 1 || res.Stderr != want {
		t.Fatalf("got (Code=%d, stderr=%q), want (1, %q)", res.Code, res.Stderr, want)
	}
	if strings.Contains(res.Stderr, "not a contained regular file") {
		t.Fatalf("deleted file must not produce the U3 text: %q", res.Stderr)
	}
}

// TestContainedRegularFile_UnreadableIntermediate_FailsClosed (U3 AC-4b,
// non-Windows): a 0o000 intermediate directory makes the child Lstat fail
// with a permission error; that is not fs.ErrNotExist, so the component
// type was never verified and the check fails closed with an lstat reason.
func TestContainedRegularFile_UnreadableIntermediate_FailsClosed(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("POSIX permission bits do not gate Lstat on Windows")
	}
	if os.Geteuid() == 0 {
		t.Skip("root bypasses directory permission checks")
	}
	root := t.TempDir()
	locked := filepath.Join(root, "internal")
	if err := os.MkdirAll(filepath.Join(locked, "d"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(locked, 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(locked, 0o755) })
	ok, reason := containedRegularFile(root, "internal/d/y.go")
	if ok || !strings.HasPrefix(reason, "lstat internal/d: ") {
		t.Fatalf("containedRegularFile = (%v, %q), want (false, \"lstat internal/d: ...\")", ok, reason)
	}
}

// This file (retiredarch.go) ports load_fixture_manifest, run_fixture_self_test,
// run_repo_scan, report_assertion and main's mode dispatch from
// scripts/lib/retired_arch.py, and additionally absorbs the notice-line
// and success-banner text that scripts/check-retired-architecture.sh's
// bash case statement used to print around each Python invocation
// (M2-T11): Run is the single entry point the reduced bash wrapper now
// calls once per invocation, exactly mirroring how writepath.Run already
// absorbs its own wrapper's self-test/repo-scan/banner composition.
package retiredarch

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem"
)

// Result carries the ordered stdout/stderr text and process-style exit
// code a single mode-scoped operation produced (mirrors
// writepath.Result's shape/contract exactly).
type Result struct {
	Stdout string
	Stderr string
	Code   int
}

// selfTestEngine names one (label, scan function) pair a fixture suite
// exercises, mirroring self_test_engines_for_name's list-of-tuples shape.
type selfTestEngine struct {
	label string
	scan  func(path string) ([]string, error)
}

// fixtureSuite ports one entry of the fixture_suites list.
type fixtureSuite struct {
	name           string
	glob           string // relative to the manifest's own directory
	manifestPath   string // repo-relative, forward-slashed
	engines        []string
	engineOverride []selfTestEngine // nil unless the suite overrides (go-differential)
}

// fixtureSuites ports the fixture_suites list verbatim (name, glob,
// manifest path, declared engines, and the go-differential suite's
// engine_override to the unmasked scan).
var fixtureSuites = []fixtureSuite{
	{
		name:         "toml",
		glob:         "retired-*.toml",
		manifestPath: "scripts/testdata/retired-manifest.json",
		engines:      []string{"toml"},
	},
	{
		name:         "go",
		glob:         "retiredgo/*.go",
		manifestPath: "scripts/testdata/retiredgo-manifest.json",
		engines:      []string{"go"},
	},
	{
		name:         "go-differential",
		glob:         "retiredgo-differential/*.go",
		manifestPath: "scripts/testdata/retiredgo-differential-manifest.json",
		engines:      []string{"go"},
		engineOverride: []selfTestEngine{
			{label: "go-unmasked", scan: func(path string) ([]string, error) { return scanGo(path, false) }},
		},
	},
}

// reject-only suites: go-differential is the ONLY named exemption (named
// explicitly, not a general escape hatch -- see run_fixture_self_test's
// Python comment).
var rejectOnlySuites = map[string]bool{"go-differential": true}

// selfTestEnginesForName ports self_test_engines_for_name. Go always has
// the BurntSushi TOML engine available (no optional-dependency branch,
// unlike Python's `if tomllib is not None`), so the 'toml' case always
// returns both engines.
func selfTestEnginesForName(engineName string) []selfTestEngine {
	switch engineName {
	case "go":
		return []selfTestEngine{{label: "go", scan: func(path string) ([]string, error) { return scanGo(path, true) }}}
	case "toml":
		return []selfTestEngine{
			{label: "tomllib", scan: func(path string) ([]string, error) { return scanTomlPrimary(path), nil }},
			{label: "fallback", scan: scanTomlFallback},
		}
	default:
		return nil
	}
}

// manifestShapeError is the error type loadFixtureManifest returns for a
// genuinely non-object (or null) top-level JSON value -- the ONLY case
// this type is used for. Every read failure, invalid-UTF-8 rejection,
// JSON syntax error, and trailing-data-after-top-level-value error is
// returned as a plain error instead (Copilot review finding, PR #83,
// round 9: an earlier version of this comment described this type as
// also covering "a read/decode failure it cannot otherwise attribute",
// which stopped being true once those paths gained their own explicit
// plain-error returns above). The caller checks for this type
// specifically so it can emit the BARE message Python's
// `raise SystemExit(f"invalid fixture manifest shape: ...")` prints
// (no "::error::" prefix -- see F-6), while any other loadFixtureManifest
// error keeps the existing "::error::"-prefixed ED-2 treatment.
type manifestShapeError struct {
	path string
}

func (e *manifestShapeError) Error() string {
	return fmt.Sprintf("invalid fixture manifest shape: %s", filepath.ToSlash(e.path))
}

// manifestExpectation models one fixture-manifest lookup result the way
// Python's `manifest.get(name)` does: Python's dict.get cannot distinguish
// an absent key from an explicit JSON `null` value (both yield None), so
// this port collapses both into isNone; but unlike a bare
// map[string]string, an explicit JSON empty string "" stays
// distinguishable from either, and a present-but-non-string JSON value
// (number, bool, list, object) is carried through as `other` instead of
// being rejected at load time (see loadFixtureManifest and F-2).
type manifestExpectation struct {
	isNone   bool
	isString bool
	str      string      // valid when isString
	other    interface{} // the decoded JSON value when present, non-string, non-null
}

// repr mirrors Python's {expectation!r} formatting used in the
// "unknown expectation" failure text: None for an absent key or an
// explicit null, the quoted string for a string value, or a best-effort
// Python-repr rendering of the raw JSON value for anything else.
func (e manifestExpectation) repr() string {
	switch {
	case e.isNone:
		return "None"
	case e.isString:
		return pysem.Repr(e.str)
	default:
		return reprJSONValue(e.other)
	}
}

// lookupExpectation looks up name in a manifest decoded by
// loadFixtureManifest, mirroring Python's manifest.get(name).
func lookupExpectation(manifest map[string]interface{}, name string) manifestExpectation {
	v, ok := manifest[name]
	if !ok || v == nil {
		return manifestExpectation{isNone: true}
	}
	if s, ok := v.(string); ok {
		return manifestExpectation{isString: true, str: s}
	}
	return manifestExpectation{other: v}
}

// reprJSONValue renders a decoded encoding/json value (as produced by
// loadFixtureManifest's json.Number-preserving decode) the way Python's
// repr() would render the equivalent json.loads() value. This is a
// best-effort rendering for the non-string manifest-expectation values
// F-2 requires loadFixtureManifest to defer (rather than hard-reject) to
// the per-fixture "unknown expectation" failure branch; nested
// object/array values are rare in practice for a fixture manifest and are
// rendered with sorted keys (Go maps have no ordering to preserve, unlike
// Python's dict), which is a known, accepted narrower-fidelity trade-off
// for this MINOR-severity divergence class.
func reprJSONValue(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return "None"
	case bool:
		if t {
			return "True"
		}
		return "False"
	case string:
		return pysem.Repr(t)
	case json.Number:
		return t.String()
	case []interface{}:
		parts := make([]string, len(t))
		for i, elem := range t {
			parts[i] = reprJSONValue(elem)
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case map[string]interface{}:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		parts := make([]string, len(keys))
		for i, k := range keys {
			parts[i] = pysem.Repr(k) + ": " + reprJSONValue(t[k])
		}
		return "{" + strings.Join(parts, ", ") + "}"
	default:
		return fmt.Sprintf("%v", t)
	}
}

// loadFixtureManifest ports load_fixture_manifest: reads and json-decodes
// manifestPath, requiring a JSON object (map) shape. Unlike Python, which
// only requires `isinstance(data, dict)` at the top level and defers any
// per-value type mismatch to the per-fixture "unknown expectation" failure
// branch inside run_fixture_self_test, an earlier version of this port
// unmarshalled directly into map[string]string, which hard-rejected the
// WHOLE manifest the moment any single value was non-string -- a false
// divergence from Python's actual failure surface (F-2). This port now
// decodes with json.Number preserved (so an int-valued expectation reprs
// as "1", not "1e+00") and only hard-rejects a genuinely non-object
// top-level shape (or a read/decode failure), leaving non-string
// per-value shapes for lookupExpectation/manifestExpectation to carry
// through to the per-fixture loop.
//
// This function decodes into a bare interface{} FIRST, deliberately
// separating two Python-distinguishable failure classes that a prior
// port revision (see F-6/F-3 history above) conflated into one
// *manifestShapeError bucket (a Copilot-review finding on PR #83, HEAD
// 2c5fe8e, review round 2):
//
//   - A genuine JSON syntax error (malformed/truncated input) or trailing
//     data after the top-level value is, in Python, an UNCAUGHT
//     json.JSONDecodeError from `json.loads` -- load_fixture_manifest
//     never catches it, so it propagates as an ordinary Python exception
//     (a traceback + exit 1), NOT the function's own
//     `raise SystemExit(f"invalid fixture manifest shape: ...")`. That
//     traceback class is exactly ED-2's "today a Python traceback with
//     exit 1; after, a one-line ::error:: message with exit 1" delta, so
//     this port returns a PLAIN error here, letting the caller's
//     "::error::"-prefixed branch (the ED-2 path) handle it -- not the
//     bare-message manifestShapeError branch.
//   - A value that decodes as syntactically valid JSON but is not a JSON
//     object at the top level (an array, string, number, bool, or the
//     `null` literal, which decodes to a nil interface{} and therefore
//     also fails the map type-assertion below) IS what Python's
//     `isinstance(data, dict)` guard rejects with its own
//     SystemExit(str) -- a BARE message (C-2's usage/infrastructure-error
//     class), no "::error::" prefix (see F-6). That is the ONLY case
//     that still produces a *manifestShapeError here.
//
// Trailing non-whitespace content after the top-level value is Python's
// "Extra data" JSONDecodeError -- the SAME uncaught-exception class as a
// syntax error, not a shape mismatch -- so it also returns a plain error,
// matching ED-8's explicit "trailing data after the top-level value
// stays a SKIP [via the decode-error path]" rule for the sibling
// merge-strategy JSON evaluator. This check requires a SECOND Decode
// call to return io.EOF (ED-8's own prescribed pattern), NOT
// json.Decoder.More(): More() is intended for array/object element
// iteration and reports false whenever the next byte is a closing
// delimiter (`]` or `}`), so it silently misses trailing data shaped
// like `{"a":"accept"}]` or `{"a":"accept"}}` (confirmed empirically) --
// a Copilot-review finding (PR #83, HEAD 0c3aa94, review round 3).
func loadFixtureManifest(manifestPath string) (map[string]interface{}, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	// Python's Path.read_text(encoding='utf-8') decodes in STRICT mode by
	// default: a manifest file containing invalid UTF-8 byte sequences
	// raises an uncaught UnicodeDecodeError (the SAME ED-2 "decode error"
	// class as a JSON syntax error, not a shape mismatch). Go's
	// string(data) conversion performs no such validation, and
	// encoding/json's decoder silently substitutes U+FFFD for invalid
	// byte sequences inside JSON string literals instead of failing --
	// so, uncaught, a corrupted manifest could continue through the
	// self-test instead of taking the required ED-2 error path (Copilot
	// review finding, PR #83, HEAD 3006994, round 7). Reject invalid
	// UTF-8 up front, before it ever reaches the JSON decoder.
	if !utf8.Valid(data) {
		return nil, fmt.Errorf("decode fixture manifest %s: invalid UTF-8", manifestPath)
	}
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.UseNumber()
	var raw interface{}
	if err := dec.Decode(&raw); err != nil {
		// A decode failure here is a genuine JSON syntax error (or EOF on
		// empty input): in Python this is an uncaught json.JSONDecodeError,
		// not the function's own SystemExit -- route it to the caller's
		// "::error::" (ED-2) branch via a plain error, not
		// *manifestShapeError.
		return nil, fmt.Errorf("decode fixture manifest %s: %w", manifestPath, err)
	}
	var trailing interface{}
	if err := dec.Decode(&trailing); err != io.EOF {
		// dec.More() is NOT a top-level EOF check -- it is intended for
		// array/object element iteration and reports false whenever the
		// NEXT byte is a closing delimiter (`]` or `}`), so it silently
		// missed trailing data like `{"a":"accept"}]` or
		// `{"a":"accept"}}` (confirmed empirically: dec.More() == false
		// for both, even though a byte remains unconsumed) -- a Copilot
		// review finding (PR #83, HEAD 0c3aa94, review round 3). A
		// second Decode call requiring io.EOF is the correct
		// whole-document-exhaustion check (and is exactly the pattern
		// ED-8 already prescribes for the sibling merge-strategy JSON
		// evaluator's identical "trailing data stays a SKIP" rule):
		// io.EOF means nothing remains; any other result (a
		// successfully-decoded second value, or a genuine second syntax
		// error) means something remains, matching Python's "Extra data"
		// JSONDecodeError for all three cases (confirmed against the
		// real CPython 3.14 interpreter).
		return nil, fmt.Errorf("decode fixture manifest %s: trailing data after top-level JSON value", manifestPath)
	}
	manifest, ok := raw.(map[string]interface{})
	if !ok {
		// raw is either a non-object JSON value (array, string, number,
		// bool) or the `null` literal (which decodes to a nil
		// interface{}, also failing this assertion): Python's
		// `isinstance(data, dict)` guard rejects all of these the same
		// way, via its own SystemExit(str) -- verified against a real
		// `python -c` run for the null case specifically. Without this
		// check, a top-level-null manifest would silently behave as an
		// empty manifest instead of failing closed with the shape error,
		// diverging from Python's actual behavior (same class of gap as
		// F-2/F-6).
		return nil, &manifestShapeError{path: manifestPath}
	}
	return manifest, nil
}

// discoverFixtures globs fixtureDir/glob and returns the matched paths
// sorted by their forward-slashed form (path.as_posix() in Python).
func discoverFixtures(fixtureDir, glob string) ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(fixtureDir, glob))
	if err != nil {
		return nil, err
	}
	sort.Slice(matches, func(i, j int) bool {
		return filepath.ToSlash(matches[i]) < filepath.ToSlash(matches[j])
	})
	return matches, nil
}

// runFixtureSelfTest ports run_fixture_self_test verbatim: framework
// invariants first (suite non-empty, accept/reject coverage or the
// go-differential reject-only exemption), then manifest/disk diffing,
// then a per-fixture-per-engine verdict loop. PASS lines from
// report_assertion AND from the per-fixture loop print to stdout
// immediately; every failure (from report_assertion's FAIL branch or a
// raw failures.append equivalent below) is collected and, if non-empty,
// dumped as "FAIL <failure>" lines to STDERR at the very end (mirroring
// Python's `print('\n'.join(...), file=sys.stderr)`), then the caller
// returns exit 1. Every report_assertion FAIL therefore appears TWICE --
// once on stdout immediately, once again on stderr at the end -- exactly
// as it does in the Python source.
func runFixtureSelfTest(root string) Result {
	var stdout strings.Builder
	var failures []string

	report := func(name string, ok bool, success, failure string) {
		if ok {
			fmt.Fprintf(&stdout, "PASS %s: %s\n", name, success)
			return
		}
		fmt.Fprintf(&stdout, "FAIL %s: %s\n", name, failure)
		failures = append(failures, fmt.Sprintf("%s: %s", name, failure))
	}

	for _, suite := range fixtureSuites {
		manifestPath := filepath.Join(root, filepath.FromSlash(suite.manifestPath))
		manifest, err := loadFixtureManifest(manifestPath)
		if err != nil {
			var shapeErr *manifestShapeError
			if errors.As(err, &shapeErr) {
				// F-6: Python's raise SystemExit(f"invalid fixture
				// manifest shape: ...") prints the BARE message + "\n"
				// (C-2's usage/infrastructure-error class) -- no
				// "::error::" prefix. That prefix is reserved for ED-2's
				// traceback-replacement class (git failures, read
				// failures, invalid UTF-8), which is what the OTHER
				// loadFixtureManifest error path (a genuine read
				// failure, e.g. missing file) still uses below.
				return Result{Stdout: stdout.String(), Stderr: fmt.Sprintf("%v\n", err), Code: 1}
			}
			return Result{Stdout: stdout.String(), Stderr: fmt.Sprintf("::error::%v\n", err), Code: 1}
		}
		fixtureDir := filepath.Dir(manifestPath)
		discoveredPaths, err := discoverFixtures(fixtureDir, suite.glob)
		if err != nil {
			return Result{Stdout: stdout.String(), Stderr: fmt.Sprintf("::error::%v\n", err), Code: 1}
		}
		discovered := make([]string, len(discoveredPaths))
		for i, p := range discoveredPaths {
			discovered[i] = filepath.Base(p)
		}

		report(
			fmt.Sprintf("suite %s non-empty", suite.name),
			len(discovered) > 0,
			fmt.Sprintf("suite discovered %d fixture(s)", len(discovered)),
			fmt.Sprintf("suite %s discovered zero fixtures via %s", suite.name, suite.glob),
		)

		// expectationStrs collects the distinct STRING-typed expectation
		// values seen across this suite's discovered fixtures (mirroring
		// Python's `expectations = {manifest.get(name) for name in
		// discovered}` restricted to the string members that
		// sortedNonEmptyKeys/pyStrList can faithfully repr). An explicit
		// empty string "" is a real member and IS included -- unlike the
		// prior map[string]string-based port, which conflated "" already
		// keying it OR a missing manifest entry into the same ""
		// sentinel (F-2). Python's `expectations` set itself still
		// contains a `None` member whenever any fixture's manifest entry
		// is absent or explicit JSON null (dict.get's sentinel), and
		// Python's reject-only check is `expectations <= {'reject'}` --
		// a None member fails that subset test exactly like any other
		// non-'reject' member would. So this port must track None's mere
		// presence (nonePresent) exactly as it tracks otherPresent: a
		// manifest with fixtures {"a": "reject", "b": null} is NOT
		// reject-only-clean in Python (None ⊄ {'reject'}), and folding
		// None silently out of the check here would let such a manifest
		// pass in Go while Python fails it. The display text this port
		// derives from expectationStrs deliberately mirrors Python's own
		// `sorted(e for e in expectations if e is not None)` filtering,
		// so None/otherPresent are excluded from the rendered failure
		// list even though they still gate the pass/fail verdict.
		// otherPresent additionally tracks whether any discovered
		// fixture's expectation was a non-string, non-null JSON value
		// (number/bool/array/object): such a value can never equal
		// 'accept' or 'reject', so its mere presence must still be able
		// to break the reject-only "all expectations are exactly
		// 'reject'" assertion below, even though this port does not
		// attempt to reproduce Python's exact mixed-type sorted() display
		// text for that narrow edge (see reprJSONValue's doc comment).
		expectationStrs := map[string]bool{}
		otherPresent := false
		nonePresent := false
		for _, name := range discovered {
			exp := lookupExpectation(manifest, name)
			switch {
			case exp.isString:
				expectationStrs[exp.str] = true
			case exp.isNone:
				nonePresent = true
			default:
				otherPresent = true
			}
		}
		if rejectOnlySuites[suite.name] {
			onlyReject := len(expectationStrs) > 0 && !otherPresent && !nonePresent
			for e := range expectationStrs {
				if e != "reject" {
					onlyReject = false
				}
			}
			report(
				fmt.Sprintf("suite %s reject-only exemption", suite.name),
				onlyReject,
				fmt.Sprintf("suite is a named reject-only exemption and all %d fixture(s) are 'reject'", len(discovered)),
				fmt.Sprintf("suite %s is declared reject-only but manifest expectations are %s", suite.name, pyStrList(sortedKeys(expectationStrs))),
			)
		} else {
			report(
				fmt.Sprintf("suite %s has accept and reject coverage", suite.name),
				expectationStrs["accept"] && expectationStrs["reject"],
				"suite has at least one 'accept' and one 'reject' fixture",
				fmt.Sprintf("suite %s is missing accept and/or reject coverage (found %s)", suite.name, pyStrList(sortedKeys(expectationStrs))),
			)
		}

		discoveredSet := map[string]bool{}
		for _, n := range discovered {
			discoveredSet[n] = true
		}
		manifestSet := map[string]bool{}
		for n := range manifest {
			manifestSet[n] = true
		}
		var missingManifest, extraManifest []string
		for n := range discoveredSet {
			if !manifestSet[n] {
				missingManifest = append(missingManifest, n)
			}
		}
		for n := range manifestSet {
			if !discoveredSet[n] {
				extraManifest = append(extraManifest, n)
			}
		}
		sort.Strings(missingManifest)
		sort.Strings(extraManifest)
		for _, name := range missingManifest {
			failures = append(failures, fmt.Sprintf("%s: discovered by %s but missing from %s", name, suite.glob, filepath.Base(manifestPath)))
		}
		for _, name := range extraManifest {
			failures = append(failures, fmt.Sprintf("%s: listed in %s but not found on disk", name, filepath.Base(manifestPath)))
		}

		for _, path := range discoveredPaths {
			name := filepath.Base(path)
			exp := lookupExpectation(manifest, name)
			engineName := engineForPath(path)

			engineDeclared := false
			for _, e := range suite.engines {
				if e == engineName {
					engineDeclared = true
					break
				}
			}
			if !engineDeclared {
				failures = append(failures, fmt.Sprintf("%s: suite %s expected engines %s but engine_for_path returned %s", name, suite.name, pyStrList(suite.engines), pysem.Repr(engineName)))
				continue
			}

			engines := suite.engineOverride
			if engines == nil {
				engines = selfTestEnginesForName(engineName)
			}
			if len(engines) == 0 {
				failures = append(failures, fmt.Sprintf("%s: no self-test engines configured for %s", name, pysem.Repr(engineName)))
				continue
			}

			// expectationRepr mirrors Python's {expectation!r}: an absent
			// key or an explicit JSON null both repr as "None"; a string
			// value reprs as its own quoted repr; any other JSON value
			// (F-2: no longer hard-rejected at manifest-load time) reprs
			// via reprJSONValue, so it still reaches this exact
			// "unknown expectation %s in %s" failure branch below rather
			// than a manifest-load-time hard reject.
			expectationRepr := exp.repr()

			for _, engine := range engines {
				findings, err := engine.scan(path)
				if err != nil {
					return Result{Stdout: stdout.String(), Stderr: fmt.Sprintf("::error::%s: %v\n", filepath.ToSlash(path), err), Code: 1}
				}
				rejected := len(findings) > 0
				label := fmt.Sprintf("%s [%s]", name, engine.label)

				switch {
				case exp.isString && exp.str == "accept":
					if rejected {
						failures = append(failures, fmt.Sprintf("%s: expected clean, got findings: %s", label, strings.Join(findings, "; ")))
					} else {
						fmt.Fprintf(&stdout, "PASS %s: clean as expected\n", label)
					}
				case exp.isString && exp.str == "reject":
					if rejected {
						fmt.Fprintf(&stdout, "PASS %s: rejected as expected\n", label)
					} else {
						failures = append(failures, fmt.Sprintf("%s: expected rejection, got clean", label))
					}
				default:
					failures = append(failures, fmt.Sprintf("%s: unknown expectation %s in %s", label, expectationRepr, filepath.Base(manifestPath)))
				}
			}
		}
	}

	if len(failures) > 0 {
		var b strings.Builder
		for _, f := range failures {
			b.WriteString("FAIL " + f + "\n")
		}
		return Result{Stdout: stdout.String(), Stderr: b.String(), Code: 1}
	}
	return Result{Stdout: stdout.String(), Code: 0}
}

// sortedKeys returns the sorted keys of set. Since expectationStrs (the
// only caller) is now built purely from genuine STRING-typed manifest
// expectations (lookupExpectation's isString branch), an absent key or an
// explicit JSON null never reaches this map at all -- unlike the prior
// "" sentinel scheme, an explicit empty-string expectation "" IS a real
// member here and is no longer dropped (F-2: absent-key vs
// explicit-empty-string ambiguity).
func sortedKeys(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// runRepoScan ports run_repo_scan: select every repo path, scan it, and
// fail closed (an ::error::-free plain findings dump to stderr, exit 1,
// matching Python's `print(..., file=sys.stderr); raise SystemExit(1)`
// exactly -- no ::error:: prefix here, unlike a git/read error) if any
// finding turns up.
func runRepoScan(root string, git GitRunner) Result {
	relPaths, err := selectRepoPaths(root, git)
	if err != nil {
		return Result{Stderr: fmt.Sprintf("::error::git ls-files: %v\n", err), Code: 1}
	}

	var findings []string
	for _, rel := range relPaths {
		findings = append(findings, scanPath(filepath.Join(root, filepath.FromSlash(rel)))...)
	}

	if len(findings) > 0 {
		return Result{Stderr: strings.Join(findings, "\n") + "\n", Code: 1}
	}
	return Result{Code: 0}
}

// Run dispatches the top-level CLI-style modes exactly like
// scripts/check-retired-architecture.sh's case statement used to (before
// M2-T11 reduced that wrapper to argument-parsing plus the shared
// build/invoke/cleanup runner):
//
//	flag == ""                      -> notice mode=repo, then repo scan
//	flag == "--self-test"           -> notice mode=self-test, then fixture
//	                                    self-test, then the selection
//	                                    self-test, then (if both passed) a
//	                                    repo scan, then a trailing banner
//	flag == "--self-test-integrity" -> notice mode=self-test-integrity,
//	                                    then fixture self-test, then the
//	                                    selection self-test only (repo scan
//	                                    deliberately skipped), then a
//	                                    trailing banner
//	anything else                   -> usage to stderr, exit 2, NO notice
//	                                    line (bash's own case default
//	                                    branch never invoked the engine)
func Run(flag, root string, git GitRunner, stdout, stderr io.Writer) int {
	switch flag {
	case "":
		_, _ = fmt.Fprintln(stdout, "::notice::retired-arch gate mode=repo")
		res := runRepoScan(root, git)
		_, _ = io.WriteString(stdout, res.Stdout)
		_, _ = io.WriteString(stderr, res.Stderr)
		return res.Code

	case "--self-test":
		_, _ = fmt.Fprintln(stdout, "::notice::retired-arch gate mode=self-test")
		if code := runSelfTestAssertions(root, git, stdout, stderr); code != 0 {
			return code
		}
		repoRes := runRepoScan(root, git)
		_, _ = io.WriteString(stdout, repoRes.Stdout)
		_, _ = io.WriteString(stderr, repoRes.Stderr)
		if repoRes.Code != 0 {
			return repoRes.Code
		}
		_, _ = fmt.Fprintln(stdout, "self-test passed: fixtures matched expectations and the tracked tree is clean")
		return 0

	case "--self-test-integrity":
		_, _ = fmt.Fprintln(stdout, "::notice::retired-arch gate mode=self-test-integrity")
		if code := runSelfTestAssertions(root, git, stdout, stderr); code != 0 {
			return code
		}
		// 015.001-T: the old task ID here is preserved VERBATIM (not
		// updated to a current task ID) -- see the banner text in the
		// pre-M2-T11 wrapper, which this line reproduces exactly.
		_, _ = fmt.Fprintln(stdout, "self-test-integrity passed: fixtures matched expectations (repo scan skipped, 015.001-T)")
		return 0

	default:
		_, _ = fmt.Fprintln(stderr, "usage: scripts/check-retired-architecture.sh [--self-test|--self-test-integrity]")
		return 2
	}
}

// runSelfTestAssertions runs the fixture self-test then the repo
// selection self-test (both "self-test" and "self-test-integrity" run
// exactly this pair; whether a repo scan and/or which banner follows is
// Run's concern), writing their combined stdout/stderr and returning 0 or
// 1.
func runSelfTestAssertions(root string, git GitRunner, stdout, stderr io.Writer) int {
	fixtureRes := runFixtureSelfTest(root)
	_, _ = io.WriteString(stdout, fixtureRes.Stdout)
	_, _ = io.WriteString(stderr, fixtureRes.Stderr)
	if fixtureRes.Code != 0 {
		return fixtureRes.Code
	}

	selectionOut, err := runRepoSelectionSelfTest(root, git)
	_, _ = io.WriteString(stdout, selectionOut.stdout.String())
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "::error::%v\n", err)
		return 1
	}
	if selectionOut.failed {
		return 1
	}
	return 0
}

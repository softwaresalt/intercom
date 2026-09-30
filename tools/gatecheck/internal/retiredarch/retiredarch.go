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
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

// loadFixtureManifest ports load_fixture_manifest: reads and json-decodes
// manifestPath, requiring a JSON object (map) shape. A read or decode
// error, or a non-object top-level shape, fails closed with an error the
// caller reports as an ::error:: line + exit 1 (matching Python's
// SystemExit(str) -> exit 1 with the message on stderr).
func loadFixtureManifest(manifestPath string) (map[string]string, error) {
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}
	var raw map[string]string
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("invalid fixture manifest shape: %s", filepath.ToSlash(manifestPath))
	}
	return raw, nil
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

		expectationSet := map[string]bool{}
		for _, name := range discovered {
			if v, ok := manifest[name]; ok {
				expectationSet[v] = true
			} else {
				expectationSet[""] = true
			}
		}
		if rejectOnlySuites[suite.name] {
			onlyReject := len(expectationSet) > 0
			for e := range expectationSet {
				if e != "reject" {
					onlyReject = false
				}
			}
			report(
				fmt.Sprintf("suite %s reject-only exemption", suite.name),
				onlyReject,
				fmt.Sprintf("suite is a named reject-only exemption and all %d fixture(s) are 'reject'", len(discovered)),
				fmt.Sprintf("suite %s is declared reject-only but manifest expectations are %s", suite.name, pyStrList(sortedNonEmptyKeys(expectationSet))),
			)
		} else {
			report(
				fmt.Sprintf("suite %s has accept and reject coverage", suite.name),
				expectationSet["accept"] && expectationSet["reject"],
				"suite has at least one 'accept' and one 'reject' fixture",
				fmt.Sprintf("suite %s is missing accept and/or reject coverage (found %s)", suite.name, pyStrList(sortedNonEmptyKeys(expectationSet))),
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
			expectation, hasExpectation := manifest[name]
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

			// expectationRepr mirrors Python's {expectation!r}: manifest.get(name)
			// yields None (repr "None") when the key is absent, otherwise the
			// stored string's own repr.
			expectationRepr := "None"
			if hasExpectation {
				expectationRepr = pysem.Repr(expectation)
			}

			for _, engine := range engines {
				findings, err := engine.scan(path)
				if err != nil {
					return Result{Stdout: stdout.String(), Stderr: fmt.Sprintf("::error::%s: %v\n", filepath.ToSlash(path), err), Code: 1}
				}
				rejected := len(findings) > 0
				label := fmt.Sprintf("%s [%s]", name, engine.label)

				switch {
				case hasExpectation && expectation == "accept":
					if rejected {
						failures = append(failures, fmt.Sprintf("%s: expected clean, got findings: %s", label, strings.Join(findings, "; ")))
					} else {
						fmt.Fprintf(&stdout, "PASS %s: clean as expected\n", label)
					}
				case hasExpectation && expectation == "reject":
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

// sortedNonEmptyKeys returns the sorted keys of set, excluding the ""
// sentinel used above to stand in for a missing manifest entry (Python's
// manifest.get(name) yields None there, and sorted(e for e in expectations
// if e is not None) drops it the same way).
func sortedNonEmptyKeys(set map[string]bool) []string {
	var out []string
	for k := range set {
		if k != "" {
			out = append(out, k)
		}
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
		fmt.Fprintln(stdout, "::notice::retired-arch gate mode=repo")
		res := runRepoScan(root, git)
		_, _ = io.WriteString(stdout, res.Stdout)
		_, _ = io.WriteString(stderr, res.Stderr)
		return res.Code

	case "--self-test":
		fmt.Fprintln(stdout, "::notice::retired-arch gate mode=self-test")
		if code := runSelfTestAssertions(root, git, stdout, stderr); code != 0 {
			return code
		}
		repoRes := runRepoScan(root, git)
		_, _ = io.WriteString(stdout, repoRes.Stdout)
		_, _ = io.WriteString(stderr, repoRes.Stderr)
		if repoRes.Code != 0 {
			return repoRes.Code
		}
		fmt.Fprintln(stdout, "self-test passed: fixtures matched expectations and the tracked tree is clean")
		return 0

	case "--self-test-integrity":
		fmt.Fprintln(stdout, "::notice::retired-arch gate mode=self-test-integrity")
		if code := runSelfTestAssertions(root, git, stdout, stderr); code != 0 {
			return code
		}
		// 015.001-T: the old task ID here is preserved VERBATIM (not
		// updated to a current task ID) -- see the banner text in the
		// pre-M2-T11 wrapper, which this line reproduces exactly.
		fmt.Fprintln(stdout, "self-test-integrity passed: fixtures matched expectations (repo scan skipped, 015.001-T)")
		return 0

	default:
		fmt.Fprintln(stderr, "usage: scripts/check-retired-architecture.sh [--self-test|--self-test-integrity]")
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
		fmt.Fprintf(stderr, "::error::%v\n", err)
		return 1
	}
	if selectionOut.failed {
		return 1
	}
	return 0
}

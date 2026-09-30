// This file (selftest_selection.go) ports expected_internal_repo_paths and
// run_repo_selection_self_test from scripts/lib/retired_arch.py: every
// assertion, in the SAME order, with the SAME name, so
// retiredarch_test.go's stream comparison against the M2-T1 goldens can
// verify assertion identity even though the exact corpus-count numbers
// embedded in some assertions' text are, per C-5, evidence rather than a
// byte-parity requirement (the live tree's tracked-file counts can
// legitimately differ from the parent-commit capture).
package retiredarch

import (
	"fmt"
	"sort"
	"strings"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem"
)

// assertionOutcome records one report_assertion call's name and PASS/FAIL
// outcome, mirroring the M2-T1 golden's per-assertion shape.
type assertionOutcome struct {
	Name    string
	Outcome string // "PASS" or "FAIL"
}

// selfTestOutput accumulates one self-test run's stdout text, the ordered
// assertion outcomes, and whether any assertion failed.
type selfTestOutput struct {
	stdout     strings.Builder
	assertions []assertionOutcome
	failed     bool
}

// reportAssertion ports report_assertion: on ok, writes "PASS name:
// success\n" to stdout; otherwise writes "FAIL name: failure\n" to
// stdout and records failure text for the caller's own SystemExit(1)
// decision. Both branches print to STDOUT (Python's print() default),
// never stderr -- this matches report_assertion's own behavior exactly,
// which is distinct from run_fixture_self_test's separate stderr-summary
// print on exit.
func (o *selfTestOutput) reportAssertion(name string, ok bool, success, failure string) {
	if ok {
		fmt.Fprintf(&o.stdout, "PASS %s: %s\n", name, success)
		o.assertions = append(o.assertions, assertionOutcome{name, "PASS"})
		return
	}
	fmt.Fprintf(&o.stdout, "FAIL %s: %s\n", name, failure)
	o.assertions = append(o.assertions, assertionOutcome{name, "FAIL"})
	o.failed = true
}

// pyBool formats a bool the way an f-string interpolates a Python bool
// (str(True) == "True", str(False) == "False"), used only inside
// assertion FAILURE text that mirrors Python's own f-string formatting.
func pyBool(b bool) string {
	if b {
		return "True"
	}
	return "False"
}

// pyStrList formats a []string the way an f-string interpolates a Python
// list of str (str(['a', 'b']) == "['a', 'b']"), via pysem.Repr per
// element (Python's list repr calls repr() on each element).
func pyStrList(items []string) string {
	parts := make([]string, len(items))
	for i, s := range items {
		parts[i] = pysem.Repr(s)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// expectedInternalRepoPaths ports expected_internal_repo_paths: an
// INDEPENDENTLY re-derived expected internal/** selection set (its own
// `git ls-files -- internal/**` call plus its own inline filter), used by
// the "selection structural inclusion" assertion below to prove
// selectRepoPaths' real behavior against a second, independently-written
// implementation of the same three filter rules -- not a self-comparison.
func expectedInternalRepoPaths(root string, git GitRunner) ([]string, error) {
	out, err := git(root, "internal/**")
	if err != nil {
		return nil, err
	}
	listing, err := pysem.GitText(out)
	if err != nil {
		return nil, err
	}
	var expected []string
	for _, path := range pysem.SplitLines(listing) {
		if !strings.HasPrefix(path, "internal/") {
			continue
		}
		if strings.Contains(path, "/testdata/") || strings.HasSuffix(path, "_test.go") {
			continue
		}
		if strings.HasSuffix(path, ".go") {
			expected = append(expected, path)
		}
	}
	sort.Strings(expected)
	return expected, nil
}

func stringSetDiff(a, b []string) []string {
	inB := make(map[string]bool, len(b))
	for _, s := range b {
		inB[s] = true
	}
	var diff []string
	for _, s := range a {
		if !inB[s] {
			diff = append(diff, s)
		}
	}
	sort.Strings(diff)
	return diff
}

func stringSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// runRepoSelectionSelfTest ports run_repo_selection_self_test verbatim,
// assertion-for-assertion and in the same order. It returns the
// accumulated stdout, the ordered assertion outcomes and whether any
// assertion failed (the caller decides the exit code, mirroring how
// run_repo_selection_self_test's own `raise SystemExit(1)` is Run's
// concern here, not this function's -- see writepath.Run's analogous
// split for precedent).
func runRepoSelectionSelfTest(root string, git GitRunner) (selfTestOutput, error) {
	var out selfTestOutput

	relPaths, err := selectRepoPaths(root, git)
	if err != nil {
		return out, err
	}

	var internalActual []string
	for _, p := range relPaths {
		if strings.HasPrefix(p, "internal/") {
			internalActual = append(internalActual, p)
		}
	}
	sort.Strings(internalActual)

	internalExpected, err := expectedInternalRepoPaths(root, git)
	if err != nil {
		return out, err
	}

	missingInternal := stringSetDiff(internalExpected, internalActual)
	extraInternal := stringSetDiff(internalActual, internalExpected)

	out.reportAssertion(
		"selection internal non-empty",
		len(internalExpected) > 0,
		fmt.Sprintf("independently-derived expected internal/** set is non-empty (%d paths)", len(internalExpected)),
		"independently-derived expected internal/** set was empty -- git ls-files -- internal/** "+
			"likely returned nothing; the structural inclusion assertion below would pass vacuously",
	)

	missingDisplay := missingInternal
	if len(missingDisplay) == 0 {
		missingDisplay = []string{"none"}
	}
	extraDisplay := extraInternal
	if len(extraDisplay) == 0 {
		extraDisplay = []string{"none"}
	}
	out.reportAssertion(
		"selection structural inclusion",
		stringSliceEqual(internalActual, internalExpected),
		fmt.Sprintf("selected every tracked internal non-test, non-testdata Go file (%d paths)", len(internalActual)),
		fmt.Sprintf("missing=%s extra=%s", pyStrList(missingDisplay), pyStrList(extraDisplay)),
	)

	internalTestOrTestdata := false
	for _, p := range relPaths {
		if strings.HasPrefix(p, "internal/") && (strings.HasSuffix(p, "_test.go") || strings.Contains(p, "/testdata/")) {
			internalTestOrTestdata = true
			break
		}
	}
	out.reportAssertion(
		"selection internal exclusions",
		!internalTestOrTestdata,
		"excluded internal test files and internal testdata paths",
		"selected an internal test or testdata path",
	)

	predicateGuard := !shouldScanRepoPath("scripts/testdata/retiredgo/x.go")
	selectedScripts := false
	for _, p := range relPaths {
		if strings.HasPrefix(p, "scripts/") {
			selectedScripts = true
			break
		}
	}
	out.reportAssertion(
		"selection self-scan guard",
		!selectedScripts && predicateGuard,
		"did not select scripts/ paths and predicate rejects scripts/testdata/retiredgo/x.go",
		"self-scan guard failed for scripts/ selection or predicate probe",
	)

	out.reportAssertion(
		"selection non-empty",
		len(relPaths) > 0,
		fmt.Sprintf("selected %d tracked repo paths", len(relPaths)),
		"selected zero repo paths",
	)

	out.reportAssertion(
		"dispatch config.toml.example",
		engineForPath("config.toml.example") == "toml",
		"engine_for_path routes config.toml.example to the TOML engine",
		fmt.Sprintf("engine_for_path returned %s", pysem.Repr(engineForPath("config.toml.example"))),
	)

	includesConfigExample := false
	for _, p := range relPaths {
		if p == "config.toml.example" {
			includesConfigExample = true
			break
		}
	}
	out.reportAssertion(
		"selection includes config.toml.example",
		includesConfigExample,
		"real selection set includes config.toml.example end-to-end",
		"config.toml.example was not present in the real selection set",
	)

	var unmapped []string
	for _, p := range relPaths {
		if engineForPath(joinRoot(root, p)) == "" {
			unmapped = append(unmapped, p)
		}
	}
	sort.Strings(unmapped)
	out.reportAssertion(
		"selection dispatch coverage",
		len(unmapped) == 0,
		"every selected repo path resolves to a known scan engine",
		fmt.Sprintf("selected path(s) with no known engine (would hit the fail-closed synthetic finding): %s", pyStrList(unmapped)),
	)

	cmdTestProbe := shouldScanRepoPath("cmd/x/y_test.go")
	cmdTestdataProbe := shouldScanRepoPath("cmd/x/testdata/z.go")
	var cmdTestFilesSelected []string
	for _, p := range relPaths {
		if strings.HasPrefix(p, "cmd/") && strings.HasSuffix(p, "_test.go") {
			cmdTestFilesSelected = append(cmdTestFilesSelected, p)
		}
	}
	out.reportAssertion(
		"selection cmd/ coverage (AG-5/D4)",
		cmdTestProbe && cmdTestdataProbe && len(cmdTestFilesSelected) > 0,
		fmt.Sprintf("cmd/ deliberately includes *_test.go and testdata/ paths (%d tracked cmd/**/*_test.go file(s) selected)", len(cmdTestFilesSelected)),
		fmt.Sprintf(
			"AG-5/D4 violation: this assertion guards the DELIBERATE decision that cmd/ "+
				"selection coverage includes test and testdata paths (unlike internal/**) -- "+
				"should_scan_repo_path('cmd/x/y_test.go')=%s, "+
				"should_scan_repo_path('cmd/x/testdata/z.go')=%s, "+
				"tracked cmd/**/*_test.go selected=%d",
			pyBool(cmdTestProbe), pyBool(cmdTestdataProbe), len(cmdTestFilesSelected),
		),
	)

	pin := SelectionPathspecPin(root)
	out.reportAssertion(
		"selection pathspec pin (AC-6/AG-1)",
		pin.PathspecOK && pin.PrefixOK,
		"select_repo_paths pathspec and should_scan_repo_path prefix set pinned via inspect.getsource() source-text extraction",
		fmt.Sprintf(
			"AC-6/AG-1 pin failed -- source extraction failed (fail-closed) or an expected "+
				"literal is missing: select_repo_paths region found=%s, "+
				"should_scan_repo_path region found=%s, "+
				"pathspec literals present=%s, prefix literals present=%s",
			pyBool(pin.SelectFound), pyBool(pin.GuardFound), pyBool(pin.PathspecOK), pyBool(pin.PrefixOK),
		),
	)

	return out, nil
}

// joinRoot mirrors Python's `root / p` (pathlib's join operator) closely
// enough for engineForPath's suffix-only dispatch: it need not produce a
// byte-identical OS path, only one whose filepath.Ext/filepath.Base match
// what `root / p` would have.
func joinRoot(root, p string) string {
	if root == "" {
		return p
	}
	return root + "/" + p
}

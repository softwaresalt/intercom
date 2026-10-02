// This file (unignore.go) ports do_self_test_landing_precondition,
// make_scenario_repo, run_scenario and do_self_test_scenarios from the
// check-unignore-regression.sh Python heredoc (plan §6, M3-T6), and
// additionally absorbs the bash wrapper's own argument parsing and mode
// dispatch as Run (plan §6, M3-T7) -- mirroring exactly how
// retiredarch.Run and writepath.Run already absorb their own wrappers'
// argument parsing.
//
// This package documents an accepted, scoped Principle III exception, not an
// undiscovered hole: unignore self-test scenario and scratch repos are created
// by os.MkdirTemp("", ...) outside the repo root. The retired Python used
// tempfile.TemporaryDirectory() for the same purpose, so this preserves its
// temporary-repository behavior.
package unignore

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem"
)

// pyListRepr reproduces Python's repr() of a list of strings (e.g.
// "['a', 'b']"), exactly as used by the f-string interpolations of
// failures/regressed-paths lists throughout this engine.
func pyListRepr(items []string) string {
	if len(items) == 0 {
		return "[]"
	}
	parts := make([]string, len(items))
	for i, s := range items {
		parts[i] = pysem.Repr(s)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

// doSelfTestLandingPrecondition reproduces do_self_test_landing_precondition:
// the landing-precondition guard that every denylist entry is ignored at
// HEAD. A failure prints to stderr; a pass prints to stdout.
func doSelfTestLandingPrecondition(git GitRunner, repoDir, scratchRoot string, stdout, stderr io.Writer) bool {
	evaluated, failures, err := runDenylistCheck(git, repoDir, scratchRoot)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "%v\n", err)
		return false
	}
	if len(failures) > 0 {
		_, _ = fmt.Fprintf(stderr, "FAIL landing-precondition: the following denylist entries are NOT ignored at HEAD: %s\n", pyListRepr(failures))
		return false
	}
	_, _ = fmt.Fprintf(stdout, "PASS landing-precondition: all %d denylist entries are ignored at HEAD\n", evaluated)
	return true
}

// makeScenarioRepo reproduces make_scenario_repo: builds a fresh, throwaway
// git repository under tmpRoot/name with a base commit (old_gitignore) and
// a head commit (new_gitignore, plus any existing/tracked fixture paths),
// returning the repo directory and the base commit's resolved ref.
func makeScenarioRepo(git GitRunner, tmpRoot, name, oldGitignore, newGitignore string, existingPaths, trackedPaths []string) (repoDir, baseRef string, err error) {
	repoDir = filepath.Join(tmpRoot, name)
	if mkErr := os.MkdirAll(repoDir, 0o755); mkErr != nil {
		return "", "", mkErr
	}
	run := func(args ...string) error {
		_, stderr, runErr := git(repoDir, nil, args...)
		if runErr != nil {
			return fmt.Errorf("git %s in %s: %w: %s", strings.Join(args, " "), repoDir, runErr, strings.TrimSpace(string(stderr)))
		}
		return nil
	}

	if err = run("init", "-q", "-b", "main"); err != nil {
		return "", "", err
	}
	if err = run("config", "user.email", "fixture@example.invalid"); err != nil {
		return "", "", err
	}
	if err = run("config", "user.name", "fixture"); err != nil {
		return "", "", err
	}

	if err = os.WriteFile(filepath.Join(repoDir, ".gitignore"), []byte(oldGitignore), 0o644); err != nil {
		return "", "", err
	}
	if err = run("add", ".gitignore"); err != nil {
		return "", "", err
	}
	if err = run("commit", "-q", "-m", "base"); err != nil {
		return "", "", err
	}
	revOut, revErr, gitErr := git(repoDir, nil, "rev-parse", "HEAD")
	if gitErr != nil {
		return "", "", fmt.Errorf("git rev-parse HEAD in %s: %w: %s", repoDir, gitErr, strings.TrimSpace(string(revErr)))
	}
	revText, convErr := pysem.GitText(revOut)
	if convErr != nil {
		return "", "", convErr
	}
	baseRef = pysem.Strip(revText)

	if err = os.WriteFile(filepath.Join(repoDir, ".gitignore"), []byte(newGitignore), 0o644); err != nil {
		return "", "", err
	}
	for _, rel := range existingPaths {
		if err = writeFixtureFile(repoDir, rel); err != nil {
			return "", "", err
		}
	}
	addArgs := []string{"add", ".gitignore"}
	for _, rel := range trackedPaths {
		if err = writeFixtureFile(repoDir, rel); err != nil {
			return "", "", err
		}
		addArgs = append(addArgs, rel)
	}
	if err = run(addArgs...); err != nil {
		return "", "", err
	}
	if err = run("commit", "-q", "-m", "head"); err != nil {
		return "", "", err
	}

	return repoDir, baseRef, nil
}

// runScenario reproduces run_scenario: builds the scenario repo, runs the
// differential check against it, and asserts the expected accept/reject
// outcome, printing a PASS/FAIL line exactly as the Python engine does.
func runScenario(git GitRunner, tmpRoot, name, oldGitignore, newGitignore string, existingPaths []string, expectReject bool, trackedPaths []string, stdout, stderr io.Writer) bool {
	repoDir, baseRef, err := makeScenarioRepo(git, tmpRoot, name, oldGitignore, newGitignore, existingPaths, trackedPaths)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "::error::%v\n", err)
		return false
	}
	scratchRoot := filepath.Join(tmpRoot, name+"-scratch")
	evaluated, failures, err := runDifferentialCheck(git, repoDir, scratchRoot, baseRef, "HEAD")
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "::error::%v\n", err)
		return false
	}
	rejected := len(failures) > 0
	if expectReject {
		if rejected {
			_, _ = fmt.Fprintf(stdout, "PASS %s: rejected as expected (regressed paths: %s)\n", name, pyListRepr(failures))
			return true
		}
		_, _ = fmt.Fprintf(stderr, "FAIL %s: expected rejection, got clean (evaluated=%d)\n", name, evaluated)
		return false
	}
	if rejected {
		_, _ = fmt.Fprintf(stderr, "FAIL %s: expected clean, got rejection: %s\n", name, pyListRepr(failures))
		return false
	}
	_, _ = fmt.Fprintf(stdout, "PASS %s: clean as expected (evaluated=%d)\n", name, evaluated)
	return true
}

// doSelfTestScenarios reproduces do_self_test_scenarios: the three named
// synthetic accept/reject scenarios (A, B, C), each built as an ephemeral,
// throwaway git repo under a fresh os.MkdirTemp-style temp directory.
func doSelfTestScenarios(git GitRunner, stdout, stderr io.Writer) bool {
	tmpRoot, err := os.MkdirTemp("", "unignore-selftest-")
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "::error::%v\n", err)
		return false
	}
	defer func() { _ = os.RemoveAll(tmpRoot) }()

	ok := true

	// Scenario A: an EXISTING, previously-ignored file is un-ignored by a
	// negation added at head. Must be REJECTED.
	ok = runScenario(
		git, tmpRoot, "reject-existing-file-unignored",
		"foo/bar.secret\n", "foo/bar.secret\n!foo/bar.secret\n",
		[]string{"foo/bar.secret"}, true, nil, stdout, stderr,
	) && ok

	// Scenario B: a negation for a path that does NOT exist (the
	// stowaway's actual case) is added at head. Must be ACCEPTED.
	ok = runScenario(
		git, tmpRoot, "accept-nonexistent-negation",
		"foo/\n", "foo/\n!foo/keep.txt\n",
		nil, false, nil, stdout, stderr,
	) && ok

	// Scenario C: a negation un-ignores a previously-ignored file, AND
	// that file is `git add`ed in the SAME change -- making it TRACKED at
	// head, not untracked. Must still be REJECTED.
	ok = runScenario(
		git, tmpRoot, "reject-tracked-file-unignored-in-same-change",
		"foo/bar.secret\n", "foo/bar.secret\n!foo/bar.secret\n",
		nil, true, []string{"foo/bar.secret"}, stdout, stderr,
	) && ok

	return ok
}

// Run is unignore's top-level entry point, absorbing the bash wrapper's
// own argument parsing (mode=check|self-test, --base-ref, --head-ref,
// -h/--help, and the unrecognized-argument usage/exit-2 path) so
// scripts/check-unignore-regression.sh (M3-T7) is reduced to the shared
// build/invoke/cleanup runner, exactly like write-path and retired-arch
// before it.
func Run(args []string, root string, git GitRunner, stdout, stderr io.Writer) int {
	mode := "check"
	baseRef := ""
	headRef := "HEAD"

	i := 0
	for i < len(args) {
		switch args[i] {
		case "--self-test":
			mode = "self-test"
			i++
		case "--base-ref":
			if i+1 >= len(args) {
				baseRef = ""
				i++
			} else {
				baseRef = args[i+1]
				i += 2
			}
		case "--head-ref":
			if i+1 >= len(args) {
				headRef = ""
				i++
			} else {
				headRef = args[i+1]
				i += 2
			}
		case "-h", "--help":
			_, _ = fmt.Fprintln(stderr, "Usage: check-unignore-regression.sh --self-test | --base-ref <ref> [--head-ref <ref>]")
			return 0
		default:
			_, _ = fmt.Fprintf(stderr, "::error::unrecognized argument: %s\n", args[i])
			return 2
		}
	}

	if mode == "self-test" {
		return runSelfTest(root, git, stdout, stderr)
	}

	if baseRef == "" {
		_, _ = fmt.Fprintln(stderr, "::error::--base-ref was not supplied. FAIL-CLOSED: pass --base-ref explicitly (CI passes github.event.pull_request.base.sha).")
		return 1
	}
	return runCheck(root, baseRef, headRef, git, stdout, stderr)
}

// runSelfTest reproduces the `mode_arg == "self-test"` driver block:
// landing precondition, then the three scenarios, then a SEPARATE
// real-repo denylist check (same evaluation, different message text --
// faithfully preserved duplication from the Python engine), then the
// tracked-ignored ls-files check.
func runSelfTest(root string, git GitRunner, stdout, stderr io.Writer) int {
	scratchRoot, err := os.MkdirTemp("", "unignore-scratch-")
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "::error::%v\n", err)
		return 1
	}
	defer func() { _ = os.RemoveAll(scratchRoot) }()

	overallOK := true
	overallOK = doSelfTestLandingPrecondition(git, root, scratchRoot, stdout, stderr) && overallOK
	overallOK = doSelfTestScenarios(git, stdout, stderr) && overallOK

	denylistEvaluated, denylistFailures, err := runDenylistCheck(git, root, scratchRoot)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "%v\n", err)
		overallOK = false
	} else {
		if denylistEvaluated == 0 {
			_, _ = fmt.Fprintln(stderr, "FAIL non-vacuity: denylist evaluated count is zero")
			overallOK = false
		}
		if len(denylistFailures) > 0 {
			_, _ = fmt.Fprintf(stderr, "FAIL real-repo denylist check: %s\n", pyListRepr(denylistFailures))
			overallOK = false
		} else {
			_, _ = fmt.Fprintf(stdout, "PASS real-repo denylist check: %d entries all ignored\n", denylistEvaluated)
		}
	}

	trackedIgnored, err := runLsFilesCheck(git, root)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "%v\n", err)
		overallOK = false
	} else if len(trackedIgnored) > 0 {
		_, _ = fmt.Fprintf(stderr, "FAIL: tracked files are ignored (git ls-files -i -c): %s\n", pyListRepr(trackedIgnored))
		overallOK = false
	} else {
		_, _ = fmt.Fprintln(stdout, "PASS: git ls-files -i -c --exclude-standard is empty")
	}

	if !overallOK {
		return 1
	}
	_, _ = fmt.Fprintln(stdout, "check-unignore-regression --self-test: PASS")
	return 0
}

// runCheck reproduces the `mode_arg == "check"` driver block. The
// differential check always evaluates against the literal "HEAD" ref
// (matching the Python engine's own hardcoded head_ref="HEAD" argument to
// run_differential_check, independent of the --head-ref flag, which is
// used only for the sanity-check assertion below) -- this checker
// evaluates the CURRENT working tree as head-ref, matching how CI checks
// out the PR head.
func runCheck(root, baseRef, headRef string, git GitRunner, stdout, stderr io.Writer) int {
	if headRef != "HEAD" {
		currentOut, currentErrBytes, currentErr := git(root, nil, "rev-parse", "HEAD")
		if currentErr != nil {
			_, _ = fmt.Fprintf(stderr, "::error::git rev-parse HEAD failed: %s\n", strings.TrimSpace(string(currentErrBytes)))
			return 1
		}
		targetOut, targetErrBytes, targetErr := git(root, nil, "rev-parse", headRef)
		if targetErr != nil {
			_, _ = fmt.Fprintf(stderr, "::error::git rev-parse %s failed: %s\n", headRef, strings.TrimSpace(string(targetErrBytes)))
			return 1
		}
		currentText, _ := pysem.GitText(currentOut)
		targetText, _ := pysem.GitText(targetOut)
		current := pysem.Strip(currentText)
		target := pysem.Strip(targetText)
		if current != target {
			_, _ = fmt.Fprintf(
				stderr,
				"::error::--head-ref %s does not match the current checkout HEAD (%s); this checker evaluates the CURRENT working tree as head-ref, matching how CI checks out the PR head.\n",
				pysem.Repr(headRef), current,
			)
			return 1
		}
	}

	scratchRoot, err := os.MkdirTemp("", "unignore-scratch-")
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "::error::%v\n", err)
		return 1
	}
	defer func() { _ = os.RemoveAll(scratchRoot) }()

	denylistEvaluated, denylistFailures, err := runDenylistCheck(git, root, scratchRoot)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}
	diffEvaluated, diffFailures, err := runDifferentialCheck(git, root, scratchRoot, baseRef, "HEAD")
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "%v\n", err)
		return 1
	}

	_, _ = fmt.Fprintf(stdout, "check-unignore-regression: denylist_evaluated=%d differential_evaluated=%d\n", denylistEvaluated, diffEvaluated)

	ok := true
	if denylistEvaluated == 0 {
		_, _ = fmt.Fprintln(stderr, "::error::non-vacuity violation: denylist evaluated count is zero")
		ok = false
	}
	if len(denylistFailures) > 0 {
		_, _ = fmt.Fprintf(stderr, "::error::denylist regression -- no longer ignored at head-ref: %s\n", pyListRepr(denylistFailures))
		ok = false
	}
	if len(diffFailures) > 0 {
		_, _ = fmt.Fprintf(stderr, "::error::differential regression -- ignored at base-ref but not at head-ref: %s\n", pyListRepr(diffFailures))
		ok = false
	}

	trackedIgnored, err := runLsFilesCheck(git, root)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "%v\n", err)
		ok = false
	} else if len(trackedIgnored) > 0 {
		_, _ = fmt.Fprintf(stderr, "::error::tracked files report as ignored (git ls-files -i -c --exclude-standard): %s\n", pyListRepr(trackedIgnored))
		ok = false
	}

	if !ok {
		_, _ = fmt.Fprintln(stderr, "::error::un-ignore regression detected. No waiver mechanism exists (deliberate, YAGNI) -- escalate to the operator for explicit review rather than bypassing this gate.")
		return 1
	}
	_, _ = fmt.Fprintln(stdout, "check-unignore-regression: PASS")
	return 0
}

// writeFixtureFile writes the fixed "fixture content\n" payload to
// repoDir/rel, creating parent directories as needed, matching the
// Python engine's identical fixture-content convention.
func writeFixtureFile(repoDir, rel string) error {
	full := filepath.Join(repoDir, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return err
	}
	return os.WriteFile(full, []byte("fixture content\n"), 0o644)
}

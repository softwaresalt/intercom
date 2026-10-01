package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// M4-T1/T2 (047.001-T, 047.002-T): text-level replacements for the retired
// Python wrapper-source assertions (the R-13 class (b) rows recorded in
// docs/plans/evidence/2026-09-28-gate-engine-go-migration/m4.md §2). Each
// assertion is a pure function over wrapper text returning the list of
// problems found, so the live wrappers must produce zero problems and
// deliberately mutated copies must produce at least one.
//
// Forbidden tokens are assembled from fragments so this file can never
// satisfy (or trip) its own patterns when scanned by the repo-wide greps.

var (
	wrapperInterpreterWord = regexp.MustCompile(`(?i)\b` + "py" + "thon" + `3?\b`)
	wrapperHeredoc         = regexp.MustCompile(`<` + `<`)
)

// readWrapperText returns the LF-normalized text of a repo-relative script.
func readWrapperText(t *testing.T, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(gatecheckRepoRoot(t), filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return strings.ReplaceAll(string(b), "\r\n", "\n")
}

// wrapperCodeLines returns the non-blank lines whose trimmed form does not
// start with '#' (shell comments, including the shebang).
func wrapperCodeLines(text string) []string {
	var out []string
	for _, line := range strings.Split(text, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		out = append(out, line)
	}
	return out
}

func codeContains(lines []string, needle string) bool {
	for _, l := range lines {
		if strings.Contains(l, needle) {
			return true
		}
	}
	return false
}

// retiredArchDecisionProblems asserts the 032.001-T decision record survives
// in the wrapper header.
func retiredArchDecisionProblems(text string) []string {
	var problems []string
	for _, s := range []string{
		"CANONICAL GO MASKER DECISION (032.001-T)",
		"retired-architecture SUPERSET",
		"SUPERSEDED by 032.001-T",
		"EXPECTED WRITE-PATH VERDICT DELTAS",
	} {
		if !strings.Contains(text, s) {
			problems = append(problems, "missing decision-record string: "+s)
		}
	}
	return problems
}

// retiredArchNoEngineProblems asserts no interpreter-hosted engine remains in
// the wrapper's executable lines.
func retiredArchNoEngineProblems(text string) []string {
	var problems []string
	for _, l := range wrapperCodeLines(text) {
		if wrapperInterpreterWord.MatchString(l) {
			problems = append(problems, "interpreter invocation in code line: "+l)
		}
		if wrapperHeredoc.MatchString(l) {
			problems = append(problems, "heredoc in code line: "+l)
		}
		if strings.Contains(l, "retired"+"_arch") {
			problems = append(problems, "retired engine module reference in code line: "+l)
		}
		if strings.Contains(l, "ENGINE"+"=") {
			problems = append(problems, "embedded engine variable in code line: "+l)
		}
	}
	return problems
}

// retiredArchDispatchProblems asserts the wrapper dispatches to the Go engine.
func retiredArchDispatchProblems(text string) []string {
	lines := wrapperCodeLines(text)
	var problems []string
	for _, s := range []string{
		`SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"`,
		`source "${SCRIPT_DIR}/lib/gatecheck-run.sh"`,
		"gatecheck_build",
		"gatecheck_invoke retired-arch",
		"--self-test",
		"--self-test-integrity",
	} {
		if !codeContains(lines, s) {
			problems = append(problems, "missing dispatch code: "+s)
		}
	}
	return problems
}

// writePathPureBashProblems asserts the write-path wrapper is a pure-bash
// dispatcher to the Go engine with no embedded masker.
func writePathPureBashProblems(text string) []string {
	var problems []string
	for _, s := range []string{"mask_go_" + "non_code", "from go" + "mask import"} {
		if strings.Contains(text, s) {
			problems = append(problems, "embedded masker reference: "+s)
		}
	}
	lines := wrapperCodeLines(text)
	for _, l := range lines {
		if wrapperInterpreterWord.MatchString(l) {
			problems = append(problems, "interpreter invocation in code line: "+l)
		}
		if wrapperHeredoc.MatchString(l) {
			problems = append(problems, "heredoc in code line: "+l)
		}
	}
	for _, s := range []string{"gatecheck_build", "gatecheck_invoke write-path"} {
		if !codeContains(lines, s) {
			problems = append(problems, "missing dispatch code: "+s)
		}
	}
	return problems
}

// buildAnchor is a unique code-line anchor present in both wrappers.
const buildAnchor = "gatecheck_build\nbuild_rc=$?"

// mustMutate applies a replacement and fails the test if it was a no-op, so a
// mutation case can never silently pass against unchanged text.
func mustMutate(t *testing.T, text, old, repl string) string {
	t.Helper()
	if !strings.Contains(text, old) {
		t.Fatalf("mutation anchor not found: %q", old)
	}
	return strings.Replace(text, old, repl, 1)
}

func TestRetiredArchWrapperText_DecisionRecord(t *testing.T) {
	live := readWrapperText(t, "scripts/check-retired-architecture.sh")
	if p := retiredArchDecisionProblems(live); len(p) != 0 {
		t.Fatalf("live wrapper: %v", p)
	}
	mutated := mustMutate(t, live, "SUPERSEDED by 032.001-T", "superseded")
	if p := retiredArchDecisionProblems(mutated); len(p) == 0 {
		t.Fatal("mutated wrapper (decision string removed) unexpectedly passed")
	}
}

func TestRetiredArchWrapperText_NoEmbeddedEngine(t *testing.T) {
	live := readWrapperText(t, "scripts/check-retired-architecture.sh")
	if p := retiredArchNoEngineProblems(live); len(p) != 0 {
		t.Fatalf("live wrapper: %v", p)
	}
	for name, inject := range map[string]string{
		"interpreter": "py" + "thon3 -c 'pass'\n",
		"heredoc":     "cat <" + "<'EOF'\nEOF\n",
		"module":      "echo retired" + "_arch\n",
		"engine var":  "ENGINE" + "=x\n",
	} {
		mutated := mustMutate(t, live, buildAnchor, inject+buildAnchor)
		if p := retiredArchNoEngineProblems(mutated); len(p) == 0 {
			t.Errorf("mutation %q unexpectedly passed", name)
		}
	}
	// Comment-only mentions are permitted (decision-record prose).
	commented := mustMutate(t, live, buildAnchor, "# py"+"thon3 retired"+"_arch\n"+buildAnchor)
	if p := retiredArchNoEngineProblems(commented); len(p) != 0 {
		t.Errorf("comment-only mention flagged: %v", p)
	}
}

func TestRetiredArchWrapperText_DispatchesToGoEngine(t *testing.T) {
	live := readWrapperText(t, "scripts/check-retired-architecture.sh")
	if p := retiredArchDispatchProblems(live); len(p) != 0 {
		t.Fatalf("live wrapper: %v", p)
	}
	for name, m := range map[string][2]string{
		"no invoke": {"gatecheck_invoke retired-arch \"${mode}\"", "true"},
		"no build":  {buildAnchor, "true\nbuild_rc=$?"},
		"no source": {`source "${SCRIPT_DIR}/lib/gatecheck-run.sh"`, "true"},
	} {
		mutated := mustMutate(t, live, m[0], m[1])
		if name == "no invoke" {
			mutated = mustMutate(t, mutated, "gatecheck_invoke retired-arch\n", "true\n")
		}
		if p := retiredArchDispatchProblems(mutated); len(p) == 0 {
			t.Errorf("mutation %q unexpectedly passed", name)
		}
	}
}

func TestWritePathWrapperText_PureBashDispatch(t *testing.T) {
	live := readWrapperText(t, "scripts/check-write-path-precondition.sh")
	if p := writePathPureBashProblems(live); len(p) != 0 {
		t.Fatalf("live wrapper: %v", p)
	}
	for name, m := range map[string][2]string{
		"masker":      {buildAnchor, "# from go" + "mask import x\n" + buildAnchor},
		"interpreter": {buildAnchor, "py" + "thon3 -c 'pass'\n" + buildAnchor},
		"heredoc":     {buildAnchor, "cat <" + "<'EOF'\nEOF\n" + buildAnchor},
		"no invoke":   {`gatecheck_invoke write-path "$@"`, "true"},
	} {
		mutated := mustMutate(t, live, m[0], m[1])
		if p := writePathPureBashProblems(mutated); len(p) == 0 {
			t.Errorf("mutation %q unexpectedly passed", name)
		}
	}
}

package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// M4-T3 (047.003-T): retirement assertions. Once the gate-engine Python is
// gone, three things must stay true:
//
//	(i)   outside job topology-check, no line installs Python via
//	      actions/setup-python;
//	(ii)  outside job topology-check, no `run:` line invokes a Python
//	      interpreter;
//	(iii) no Python file is tracked under scripts/ or tools/.
//
// topology-check is exempt because its setup-python step installs the
// autoharness Python package, not a gate engine.
//
// Every pattern is assembled from string fragments so this file never
// matches its own patterns (compound 2026-09-06).

const ciTopologyJob = "topology-check"

var (
	pyWord = "py" + "thon"

	// (i) `^\s*(-\s*)?uses:\s*actions/setup-python@`
	ciSetupPythonUse = regexp.MustCompile(`(?m)^\s*(-\s*)?uses:\s*actions/setup-` + pyWord + `@`)

	// (ii) `(^|[\s;&|(])python3?(\s|$)`
	ciPythonInvoke = regexp.MustCompile(`(^|[\s;&|(])` + pyWord + `3?(\s|$)`)

	// A `run:` key at any indent, optionally as a list item.
	ciAnyRunKey = regexp.MustCompile(`^(\s*)(?:-\s+)?run:[ \t]*(.*?)[ \t]*$`)
)

// ciWithoutJob returns text with the span of job `job` removed. A missing job
// leaves text unchanged.
func ciWithoutJob(text, job string) string {
	s, e, ok := ciJobSpan(text, job)
	if !ok {
		return text
	}
	return text[:s] + text[e:]
}

// ciRunLines returns every line of every `run:` value in text: the inline
// value, or each line of a block scalar (`|` / `>`). A block scalar continues
// while lines are blank or more indented than the `run:` key.
func ciRunLines(text string) []string {
	lines := strings.Split(text, "\n")
	var out []string
	for i := 0; i < len(lines); i++ {
		m := ciAnyRunKey.FindStringSubmatch(lines[i])
		if m == nil {
			continue
		}
		keyIndent := len(m[1])
		val := m[2]
		if !strings.HasPrefix(val, "|") && !strings.HasPrefix(val, ">") {
			out = append(out, val)
			continue
		}
		for j := i + 1; j < len(lines); j++ {
			l := lines[j]
			if strings.TrimSpace(l) == "" {
				continue
			}
			if len(l)-len(strings.TrimLeft(l, " \t")) <= keyIndent {
				break
			}
			out = append(out, l)
			i = j
		}
	}
	return out
}

// ciSetupPythonProblems implements assertion (i).
func ciSetupPythonProblems(text string) []string {
	var probs []string
	for _, l := range strings.Split(ciWithoutJob(text, ciTopologyJob), "\n") {
		if ciSetupPythonUse.MatchString(l) {
			probs = append(probs, "setup-"+pyWord+" outside "+ciTopologyJob+": "+strings.TrimSpace(l))
		}
	}
	return probs
}

// ciPythonInvokeProblems implements assertion (ii).
func ciPythonInvokeProblems(text string) []string {
	var probs []string
	for _, l := range ciRunLines(ciWithoutJob(text, ciTopologyJob)) {
		if ciPythonInvoke.MatchString(l) {
			probs = append(probs, pyWord+" invocation in run outside "+ciTopologyJob+": "+strings.TrimSpace(l))
		}
	}
	return probs
}

// isolatedGitEnv returns the plan §6 git-isolation environment. It also
// removes the four GIT_*_PATHSPECS variables, because any of them could make
// assertion (iii) pass while matching nothing.
func isolatedGitEnv() []string {
	blocked := map[string]bool{
		"GIT_DIR":               true,
		"GIT_WORK_TREE":         true,
		"GIT_INDEX_FILE":        true,
		"GIT_COMMON_DIR":        true,
		"GIT_OBJECT_DIRECTORY":  true,
		"GIT_LITERAL_PATHSPECS": true,
		"GIT_GLOB_PATHSPECS":    true,
		"GIT_NOGLOB_PATHSPECS":  true,
		"GIT_ICASE_PATHSPECS":   true,
	}
	var env []string
	for _, kv := range os.Environ() {
		key, _, _ := strings.Cut(kv, "=")
		if blocked[strings.ToUpper(key)] {
			continue
		}
		env = append(env, kv)
	}
	return append(env,
		"GIT_CONFIG_GLOBAL="+os.DevNull,
		"GIT_CONFIG_SYSTEM="+os.DevNull,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_TERMINAL_PROMPT=0",
	)
}

// runIsolatedGit runs git in dir under the isolation environment and returns
// stdout. Skips the test when git is not on PATH.
func runIsolatedGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found on PATH")
	}
	full := append([]string{
		"-c", "user.name=gatecheck",
		"-c", "user.email=gatecheck@invalid",
		"-c", "core.autocrlf=false",
		"-c", "init.defaultBranch=main",
	}, args...)
	cmd := exec.Command("git", full...)
	cmd.Dir = dir
	cmd.Env = isolatedGitEnv()
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, stderr.String())
	}
	return stdout.String()
}

// trackedPythonProblems implements assertion (iii) for the repository at dir.
//
// The pathspecs deliberately carry no `:(glob)` magic. Without it git's `*`
// also matches `/`, so `scripts/*.py` covers both scripts/x.py and
// scripts/lib/tests/y.py. The form `scripts/**/*.py` would miss the
// top-level file.
func trackedPythonProblems(t *testing.T, dir string) []string {
	t.Helper()
	ext := "." + "py"
	out := runIsolatedGit(t, dir, "ls-files", "-z", "--", "scripts/*"+ext, "tools/*"+ext)
	var probs []string
	for _, p := range strings.Split(out, "\x00") {
		if p != "" {
			probs = append(probs, "tracked "+pyWord+" file: "+p)
		}
	}
	return probs
}

func TestCIWiringRetire_Live(t *testing.T) {
	text := readCIWorkflow(t)
	var probs []string
	probs = append(probs, ciSetupPythonProblems(text)...)
	probs = append(probs, ciPythonInvokeProblems(text)...)
	probs = append(probs, trackedPythonProblems(t, gatecheckRepoRoot(t))...)
	for _, p := range probs {
		t.Error(p)
	}
}

// TestCIWiringRetire_TopologyExemptionIsReal checks that the exemption covers
// live content: the topology-check job's own setup-python step is flagged by
// assertion (i) once the job is renamed. Its run steps call pip, not the
// interpreter, so assertion (ii) has no live exempt content to prove.
func TestCIWiringRetire_TopologyExemptionIsReal(t *testing.T) {
	text := readCIWorkflow(t)
	renamed := mustMutate(t, text, "\n  "+ciTopologyJob+":\n", "\n  "+ciTopologyJob+"-renamed:\n")
	if got, base := len(ciSetupPythonProblems(renamed)), len(ciSetupPythonProblems(text)); got <= base {
		t.Errorf("assertion (i): unexempting %s should add a problem (got %d, base %d)", ciTopologyJob, got, base)
	}

}

// ciRetireSyntheticWorkflow is a minimal workflow that is green for (i) and
// (ii): its only Python lives inside the exempt topology-check job. Mutation
// cases inject into the lint job, so they are independent of the live file.
func ciRetireSyntheticWorkflow() string {
	return "jobs:\n" +
		"  lint:\n" +
		"    runs-on: ubuntu-latest\n" +
		"    steps:\n" +
		"      - name: Set up Go\n" +
		"        uses: actions/setup-go@0000000000000000000000000000000000000000\n" +
		"      - name: Gate\n" +
		"        run: bash scripts/check-x.sh --" + pyWord + "-free\n" +
		"  " + ciTopologyJob + ":\n" +
		"    runs-on: ubuntu-latest\n" +
		"    steps:\n" +
		"      - name: Set up Python\n" +
		"        uses: actions/setup-" + pyWord + "@0000000000000000000000000000000000000000\n" +
		"      - name: Install\n" +
		"        run: |\n" +
		"          " + pyWord + " -m pip install --require-hashes -r req.txt\n" +
		"  ci-gate:\n" +
		"    runs-on: ubuntu-latest\n" +
		"    steps:\n" +
		"      - run: echo done\n"
}

func TestCIWiringRetire_MutatedInputsAreRed(t *testing.T) {
	base := ciRetireSyntheticWorkflow()
	if p := append(ciSetupPythonProblems(base), ciPythonInvokeProblems(base)...); len(p) != 0 {
		t.Fatalf("synthetic base must be green (topology-check exempt), got %q", p)
	}
	anchor := "      - name: Gate\n"

	cases := []struct {
		name  string
		check func(string) []string
		repl  string
	}{
		{
			name:  "setup-python step in lint",
			check: ciSetupPythonProblems,
			repl:  "      - name: Injected\n        uses: actions/setup-" + pyWord + "@0000000000000000000000000000000000000000\n",
		},
		{
			name:  "setup-python as first key of a list item",
			check: ciSetupPythonProblems,
			repl:  "      - uses: actions/setup-" + pyWord + "@0000000000000000000000000000000000000000\n",
		},
		{
			name:  "inline run invokes python3",
			check: ciPythonInvokeProblems,
			repl:  "      - name: Injected\n        run: " + pyWord + "3 x\n",
		},
		{
			name:  "list-item run invokes python",
			check: ciPythonInvokeProblems,
			repl:  "      - run: " + pyWord + " -V\n",
		},
		{
			name:  "block run invokes python after a command separator",
			check: ciPythonInvokeProblems,
			repl:  "      - name: Injected\n        run: |\n          set -e\n          cd scripts && " + pyWord + " -m pytest\n",
		},
		{
			name:  "folded run invokes python in a subshell",
			check: ciPythonInvokeProblems,
			repl:  "      - name: Injected\n        run: >\n          (" + pyWord + "3 -c 1)\n",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			mutated := mustMutate(t, base, anchor, tc.repl+anchor)
			if len(tc.check(mutated)) == 0 {
				t.Errorf("mutation %q not detected", tc.name)
			}
		})
	}

	// Non-invocations must stay green: a hyphenated flag, a longer word, and
	// a step name (not a run line) that mentions Python.
	greens := []string{
		"      - name: Injected\n        run: tool --" + pyWord + "-version 3\n",
		"      - name: Injected\n        run: echo " + pyWord + "ic\n",
		"      - name: Uses no " + pyWord + " at all\n        run: echo ok\n",
	}
	for i, g := range greens {
		mutated := mustMutate(t, base, anchor, g+anchor)
		if p := ciPythonInvokeProblems(mutated); len(p) != 0 {
			t.Errorf("green case %d flagged: %q", i, p)
		}
	}
}

func TestCIWiringRetire_TrackedPythonInTempRepoIsRed(t *testing.T) {
	dir := t.TempDir()
	runIsolatedGit(t, dir, "init", "-q")
	ext := "." + "py"
	files := []string{
		filepath.Join("scripts", "x"+ext),
		filepath.Join("scripts", "a", "b"+ext),
		filepath.Join("tools", "c", "d"+ext),
		filepath.Join("scripts", "keep.sh"),
		filepath.Join("docs", "e"+ext),
	}
	for _, f := range files {
		p := filepath.Join(dir, f)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runIsolatedGit(t, dir, "add", "-A")

	got := trackedPythonProblems(t, dir)
	want := []string{"scripts/a/b" + ext, "scripts/x" + ext, "tools/c/d" + ext}
	if len(got) != len(want) {
		t.Fatalf("tracked problems = %q, want exactly %q", got, want)
	}
	for i, w := range want {
		if !strings.HasSuffix(got[i], w) {
			t.Errorf("problem[%d] = %q, want suffix %q", i, got[i], w)
		}
	}
}

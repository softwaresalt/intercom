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
//	      interpreter and no `shell:` key (step or defaults) selects one;
//	(iii) no Python file is tracked under scripts/ or tools/.
//
// topology-check is exempt because its setup-python step installs the
// autoharness Python package, not a gate engine.
//
// (i) and (ii) are line-based scans of block-style YAML. To fail closed, both
// also reject any YAML flow-style mapping outside topology-check (see
// ciFlowMappingProblems), because a flow-style step would hide its keys from
// the line-anchored patterns.
//
// Every pattern is assembled from string fragments so this file never
// matches its own patterns (compound 2026-09-06).

const ciTopologyJob = "topology-check"

var (
	pyWord = "py" + "thon"

	// (i) `(?im)^\s*(-\s*)?["']?uses["']?:\s*["']?actions/setup-python@`
	//
	// A superset of the plan's M4-T3 form: YAML may quote the key or the
	// action value, and GitHub resolves action owner and repository names
	// case-insensitively.
	ciSetupPythonUse = regexp.MustCompile(`(?im)^\s*(-\s*)?["']?uses["']?:\s*["']?actions/setup-` + pyWord + `@`)

	// (ii) `(?i)(^|[\s;&|("'/\\`])(pythonw?[0-9.]*|py|pytest)(\.exe)?(\s|$|[;&|)"'`])`
	//
	// This is a strict superset of the plan's M4-T3 form
	// `(^|[\s;&|(])python3?(\s|$)`. It also catches versioned interpreters
	// (python3.12), absolute POSIX and Windows paths (/usr/bin/python3,
	// C:\Python\Python314\python.exe), the Windows `.exe` suffix, the
	// Windows GUI interpreter (pythonw.exe), quoted invocations
	// (sh -c "python ..."), the `py` launcher, pytest, and any letter case.
	ciPythonInvoke = regexp.MustCompile(`(?i)(^|[\s;&|("'/\\` + "`" + `])(` +
		pyWord + `w?[0-9.]*|` + pyWord[:2] + `|` + pyWord[:2] + `test)(\.exe)?(\s|$|[;&|)"'` + "`" + `])`)

	// A `run:` key at any indent, optionally as a list item and optionally
	// quoted ('run': / "run":), which YAML parses identically.
	ciAnyRunKey = regexp.MustCompile(`^(\s*)(?:-\s+)?["']?run["']?:[ \t]*(.*?)[ \t]*$`)

	// A `shell:` key at any indent, optionally as a list item and optionally
	// quoted. Actions runs the step body with this program, so `shell: python`
	// executes Python even when the `run:` body names no interpreter.
	ciAnyShellKey = regexp.MustCompile(`^\s*(?:-\s+)?["']?shell["']?:[ \t]*(.*?)[ \t]*$`)

	// Any mapping key, optionally as a list item and optionally quoted, whose
	// value opens a block scalar (`|` / `>` with optional indicators).
	ciBlockScalarKey = regexp.MustCompile(`^(\s*)(?:-\s+)?(?:"[^"]*"|'[^']*'|[^\s"'#][^:#]*?):[ \t]*[|>][-+0-9]*[ \t]*(?:#.*)?$`)

	// Text a flow-mapping scan must ignore: workflow expressions, quoted
	// scalars, and comments (a `#` at line start or after whitespace).
	ciFlowNoise = regexp.MustCompile(`\$\{\{.*?\}\}|"(?:[^"\\]|\\.)*"|'(?:[^']|'')*'|(?:^|\s)#.*$`)
)

// ciFlowMappingProblems rejects YAML flow-style mappings (`{...}`) outside
// job topology-check. The other checks are line-anchored on block-style keys,
// so a step such as `- {name: Gate, run: <interpreter> -V}` would otherwise
// bypass them. Rejecting every flow mapping fails closed for `run:`, `uses:`,
// `shell:`, and any key added later. Block scalar bodies are skipped, so shell
// braces inside a `run: |` script are not flagged.
func ciFlowMappingProblems(text string) []string {
	lines := strings.Split(ciWithoutJob(text, ciTopologyJob), "\n")
	var probs []string
	for i := 0; i < len(lines); i++ {
		l := lines[i]
		if strings.ContainsAny(ciFlowNoise.ReplaceAllString(l, ""), "{}") {
			probs = append(probs, "flow-style YAML mapping outside "+ciTopologyJob+": "+strings.TrimSpace(l))
		}
		m := ciBlockScalarKey.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		keyIndent := len(m[1])
		for j := i + 1; j < len(lines); j++ {
			b := lines[j]
			if strings.TrimSpace(b) != "" && len(b)-len(strings.TrimLeft(b, " \t")) <= keyIndent {
				break
			}
			i = j
		}
	}
	return probs
}

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
	probs := ciFlowMappingProblems(text)
	for _, l := range strings.Split(ciWithoutJob(text, ciTopologyJob), "\n") {
		if ciSetupPythonUse.MatchString(l) {
			probs = append(probs, "setup-"+pyWord+" outside "+ciTopologyJob+": "+strings.TrimSpace(l))
		}
	}
	return probs
}

// ciPythonInvokeProblems implements assertion (ii).
func ciPythonInvokeProblems(text string) []string {
	probs := ciFlowMappingProblems(text)
	scoped := ciWithoutJob(text, ciTopologyJob)
	for _, l := range ciRunLines(scoped) {
		if ciPythonInvoke.MatchString(l) {
			probs = append(probs, pyWord+" invocation in run outside "+ciTopologyJob+": "+strings.TrimSpace(l))
		}
	}
	for _, l := range strings.Split(scoped, "\n") {
		if m := ciAnyShellKey.FindStringSubmatch(l); m != nil && ciPythonInvoke.MatchString(m[1]) {
			probs = append(probs, pyWord+" shell outside "+ciTopologyJob+": "+strings.TrimSpace(l))
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
// scripts/a/b/y.py. The form `scripts/**/*.py` would miss the
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
	// Separate subtests so a missing git (which skips (iii)) cannot hide a
	// failure in (i) or (ii).
	t.Run("setup-"+pyWord, func(t *testing.T) {
		for _, p := range ciSetupPythonProblems(text) {
			t.Error(p)
		}
	})
	t.Run(pyWord+"-invocation", func(t *testing.T) {
		for _, p := range ciPythonInvokeProblems(text) {
			t.Error(p)
		}
	})
	t.Run("tracked-"+pyWord+"-files", func(t *testing.T) {
		for _, p := range trackedPythonProblems(t, gatecheckRepoRoot(t)) {
			t.Error(p)
		}
	})
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
			name:  "setup-python single-quoted",
			check: ciSetupPythonProblems,
			repl:  "      - name: Injected\n        uses: 'actions/setup-" + pyWord + "@0000000000000000000000000000000000000000'\n",
		},
		{
			name:  "setup-python double-quoted mixed case",
			check: ciSetupPythonProblems,
			repl:  "      - name: Injected\n        uses: \"Actions/Setup-" + strings.ToUpper(pyWord[:1]) + pyWord[1:] + "@0000000000000000000000000000000000000000\"\n",
		},
		{
			name:  "setup-python with single-quoted uses key",
			check: ciSetupPythonProblems,
			repl:  "      - name: Injected\n        'uses': actions/setup-" + pyWord + "@0000000000000000000000000000000000000000\n",
		},
		{
			name:  "setup-python with double-quoted uses key as first list key",
			check: ciSetupPythonProblems,
			repl:  "      - \"uses\": actions/setup-" + pyWord + "@0000000000000000000000000000000000000000\n",
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
			name:  "single-quoted run key invokes python",
			check: ciPythonInvokeProblems,
			repl:  "      - name: Injected\n        'run': " + pyWord + " -V\n",
		},
		{
			name:  "double-quoted run key with block scalar",
			check: ciPythonInvokeProblems,
			repl:  "      - \"run\": |\n          " + pyWord + " -V\n",
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
		{
			name:  "versioned interpreter",
			check: ciPythonInvokeProblems,
			repl:  "      - run: " + pyWord + "3.12 -m unittest\n",
		},
		{
			name:  "absolute interpreter path",
			check: ciPythonInvokeProblems,
			repl:  "      - run: /usr/bin/" + pyWord + "3 x\n",
		},
		{
			name:  "quoted invocation",
			check: ciPythonInvokeProblems,
			repl:  "      - run: sh -c \"" + pyWord + " x\"\n",
		},
		{
			name:  "py launcher",
			check: ciPythonInvokeProblems,
			repl:  "      - run: " + pyWord[:2] + " -m unittest\n",
		},
		{
			name:  "pytest",
			check: ciPythonInvokeProblems,
			repl:  "      - run: " + pyWord[:2] + "test -q\n",
		},
		{
			name:  "uppercase interpreter",
			check: ciPythonInvokeProblems,
			repl:  "      - run: " + strings.ToUpper(pyWord) + " x\n",
		},
		{
			name:  "windows exe interpreter",
			check: ciPythonInvokeProblems,
			repl:  "      - run: " + pyWord + ".exe -V\n",
		},
		{
			name:  "windows path-qualified interpreter",
			check: ciPythonInvokeProblems,
			repl:  "      - run: C:\\Python\\Python314\\" + pyWord + ".exe x\n",
		},
		{
			name:  "windows gui interpreter shell",
			check: ciPythonInvokeProblems,
			repl:  "      - shell: " + pyWord + "w.exe {0}\n        run: print('x')\n",
		},
		{
			name:  "windows gui interpreter run",
			check: ciPythonInvokeProblems,
			repl:  "      - run: " + pyWord + "w.exe script.py\n",
		},
		{
			name:  "windows py launcher exe",
			check: ciPythonInvokeProblems,
			repl:  "      - shell: " + pyWord[:2] + ".exe {0}\n        run: print('x')\n",
		},
		{
			name:  "step shell selects python",
			check: ciPythonInvokeProblems,
			repl:  "      - name: Injected\n        shell: " + pyWord + "\n        run: print('x')\n",
		},
		{
			name:  "custom shell template selects python3",
			check: ciPythonInvokeProblems,
			repl:  "      - shell: '" + pyWord + "3 {0}'\n        run: print('x')\n",
		},
		{
			name:  "double-quoted shell key selects python",
			check: ciPythonInvokeProblems,
			repl:  "      - name: Injected\n        \"shell\": " + pyWord + "\n        run: print('x')\n",
		},
		{
			name:  "single-quoted shell key as first list key",
			check: ciPythonInvokeProblems,
			repl:  "      - 'shell': " + pyWord + "\n        run: print('x')\n",
		},
		{
			name:  "defaults run shell selects python",
			check: ciPythonInvokeProblems,
			repl:  "    defaults:\n      run:\n        shell: " + pyWord + "\n",
		},
		{
			name:  "flow-style step runs python",
			check: ciPythonInvokeProblems,
			repl:  "      - {name: Injected, run: " + pyWord + " -V}\n",
		},
		{
			name:  "flow-style step uses setup-python",
			check: ciSetupPythonProblems,
			repl:  "      - {uses: actions/setup-" + pyWord + "@0000000000000000000000000000000000000000}\n",
		},
		{
			name:  "flow-style step shell selects python",
			check: ciPythonInvokeProblems,
			repl:  "      - {name: Injected, shell: " + pyWord + ", run: echo}\n",
		},
		{
			name:  "flow-style step with quoted keys",
			check: ciPythonInvokeProblems,
			repl:  "      - {\"name\": Injected, 'run': " + pyWord + " -V}\n",
		},
		{
			name:  "flow-style uses with quoted key",
			check: ciSetupPythonProblems,
			repl:  "      - {'uses': \"actions/setup-" + pyWord + "@0000000000000000000000000000000000000000\"}\n",
		},
		{
			name:  "multi-line flow-style step",
			check: ciPythonInvokeProblems,
			repl:  "      - {\n          name: Injected,\n          run: " + pyWord + " -V\n        }\n",
		},
		{
			name:  "flow-style defaults shell",
			check: ciPythonInvokeProblems,
			repl:  "    defaults: {run: {shell: " + pyWord + "}}\n",
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

	// Non-invocations must stay green: a hyphenated flag, longer words, a
	// file extension, and a step name (not a run line) that mentions Python.
	greens := []string{
		"      - name: Injected\n        run: tool --" + pyWord + "-version 3\n",
		"      - name: Injected\n        run: echo " + pyWord + "ic\n",
		"      - name: Injected\n        run: echo " + pyWord[:2] + "project.toml copy\n",
		"      - name: Injected\n        run: ls x." + pyWord[:2] + "\n",
		"      - name: Uses no " + pyWord + " at all\n        run: echo ok\n",
		"      - name: Injected\n        shell: bash\n        run: echo ok\n",
		// Braces that are not YAML flow mappings: a workflow expression, a
		// quoted scalar, a comment, and shell braces in a block scalar body.
		"      - name: Injected\n        if: ${{ github.event_name == 'push' }}\n        run: echo ok\n",
		"      - name: \"Injected {x}\"\n        run: echo 'a {b}'\n",
		"      - name: Injected # {run: " + pyWord + "}\n        run: echo ok\n",
		"      - name: Injected\n        run: |\n          if [ -n \"${X:-}\" ]; then { echo ok; }; fi\n",
	}
	for i, g := range greens {
		mutated := mustMutate(t, base, anchor, g+anchor)
		if p := append(ciSetupPythonProblems(mutated), ciPythonInvokeProblems(mutated)...); len(p) != 0 {
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

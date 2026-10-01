package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// M4-T2 (047.002-T): text-level CI-wiring assertions that stay true after the
// Python gate engines are retired. ci.yml is read as text (no YAML library);
// jobs are located by their two-space-indented key and steps by their
// `name:` text, never by line number. Every assertion is a pure function over
// the workflow text returning problems, so the live file must produce none
// and each mutated copy must produce at least one.

const ciWorkflowRel = ".github/workflows/ci.yml"

var (
	ciJobKeyLine = regexp.MustCompile(`(?m)^  [A-Za-z0-9_-]+:[ \t]*$`)
	ciStepSplit  = regexp.MustCompile(`(?m)^      - `)
	// ciCOEKey matches a continue-on-error mapping key at any indentation,
	// whether it is a step's first key (`- continue-on-error:`), a sibling
	// key, a quoted key or an explicit `? ` key. Group 1 is the line prefix and
	// group 2 the inline value. Comment lines never match.
	ciCOEKey      = regexp.MustCompile(`(?m)^([ \t]*(?:-[ \t]+)?(?:\?[ \t]+)?)["']?continue-on-error["']?[ \t]*(?::[ \t]*(.*?))?[ \t]*$`)
	ciStepRunLine = regexp.MustCompile(`(?m)^        run:[ \t]*(.*?)[ \t]*$`)
	ciStepName    = regexp.MustCompile(`^(?:        )?name:[ \t]*(.*?)[ \t]*$`)
)

// ciStep is one `- ` list item of a job's `steps:` sequence.
type ciStep struct {
	name string
	body string
}

// readCIWorkflow returns the LF-normalized live ci.yml text.
func readCIWorkflow(t *testing.T) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(gatecheckRepoRoot(t), filepath.FromSlash(ciWorkflowRel)))
	if err != nil {
		t.Fatalf("read %s: %v", ciWorkflowRel, err)
	}
	return strings.ReplaceAll(string(b), "\r\n", "\n")
}

// ciJobSpan returns the [start, end) byte span of job `job` in text: from its
// `  job:` key line to the next two-space-indented key line or EOF.
func ciJobSpan(text, job string) (int, int, bool) {
	re := regexp.MustCompile(`(?m)^  ` + regexp.QuoteMeta(job) + `:[ \t]*$`)
	loc := re.FindStringIndex(text)
	if loc == nil {
		return 0, 0, false
	}
	end := len(text)
	if next := ciJobKeyLine.FindStringIndex(text[loc[1]:]); next != nil {
		end = loc[1] + next[0]
	}
	return loc[0], end, true
}

// ciJobBlock returns the text of job `job`.
func ciJobBlock(text, job string) (string, bool) {
	s, e, ok := ciJobSpan(text, job)
	if !ok {
		return "", false
	}
	return text[s:e], true
}

// ciSteps splits a job block into its steps.
func ciSteps(block string) []ciStep {
	parts := ciStepSplit.Split(block, -1)
	var steps []ciStep
	for _, p := range parts[1:] {
		st := ciStep{body: p}
		for _, line := range strings.Split(p, "\n") {
			if m := ciStepName.FindStringSubmatch(line); m != nil {
				st.name = strings.Trim(m[1], `"'`)
				break
			}
		}
		steps = append(steps, st)
	}
	return steps
}

// ciStepRun returns a step's `run:` text: the inline value, or for a block
// scalar (`|` / `>`), the following more-indented lines joined with "\n".
func ciStepRun(st ciStep) (string, bool) {
	loc := ciStepRunLine.FindStringSubmatchIndex(st.body)
	if loc == nil {
		return "", false
	}
	val := st.body[loc[2]:loc[3]]
	if val != "|" && val != ">" && !strings.HasPrefix(val, "|") && !strings.HasPrefix(val, ">") {
		return val, true
	}
	var lines []string
	for _, l := range strings.Split(st.body[loc[1]:], "\n")[1:] {
		if strings.TrimSpace(l) != "" && !strings.HasPrefix(l, "          ") {
			break
		}
		lines = append(lines, strings.TrimSpace(l))
	}
	return strings.TrimSpace(strings.Join(lines, "\n")), true
}

// ciStepCOEValues returns every continue-on-error value in a step, so a
// duplicated or reformatted key cannot hide behind the first match.
func ciStepCOEValues(st ciStep) []string {
	var vals []string
	for _, m := range ciCOEKey.FindAllStringSubmatch(st.body, -1) {
		vals = append(vals, m[2])
	}
	return vals
}

// ciJobCOEProblems rejects a continue-on-error key in job `job` that is not a
// step-level key: anything before the first step, or a later key that is
// neither a step's first key nor an eight-space step sibling.
func ciJobCOEProblems(text, job string) []string {
	block, ok := ciJobBlock(text, job)
	if !ok {
		return nil
	}
	first := ciStepSplit.FindStringIndex(block)
	var problems []string
	for _, loc := range ciCOEKey.FindAllStringSubmatchIndex(block, -1) {
		prefix := strings.TrimSuffix(block[loc[2]:loc[3]], "? ")
		inSteps := first != nil && loc[0] >= first[0]
		if inSteps && (prefix == "      - " || prefix == "        ") {
			continue
		}
		line := strings.TrimSpace(block[loc[0]:loc[1]])
		problems = append(problems, fmt.Sprintf("job %q: job-level continue-on-error %q", job, line))
	}
	return problems
}

func ciStepIndex(steps []ciStep, name string) int {
	for i, st := range steps {
		if st.name == name {
			return i
		}
	}
	return -1
}

// ciGateExpectation describes one gate or self-test step.
type ciGateExpectation struct {
	name        string
	runExact    string // exact inline run text, or "" to use runContains
	runContains string
	coe         string // required continue-on-error expression; "" means none allowed
}

// ciJobExpectation describes the Go-wiring of one job.
type ciJobExpectation struct {
	job   string
	gates []ciGateExpectation
}

var ciWiringExpectations = []ciJobExpectation{
	{job: "lint", gates: []ciGateExpectation{
		{name: "Run retired-architecture gate", runExact: "bash scripts/check-retired-architecture.sh", coe: "${{ vars.RETIRED_ARCH_GATE_ADVISORY == 'true' }}"},
		{name: "Run retired-architecture self-test", runExact: "bash scripts/check-retired-architecture.sh --self-test-integrity"},
		{name: "Run write-path-precondition gate", runExact: "bash scripts/check-write-path-precondition.sh", coe: "${{ vars.WRITE_PATH_GATE_ADVISORY == 'true' }}"},
		{name: "Run write-path-precondition self-test", runExact: "bash scripts/check-write-path-precondition.sh --self-test-integrity"},
	}},
	{job: "gitignore-append-only", gates: []ciGateExpectation{
		{name: "Self-test the un-ignore regression checker's own logic", runExact: "bash scripts/check-unignore-regression.sh --self-test"},
		{name: "Check un-ignore regression (behavioural differential)", runContains: "bash scripts/check-unignore-regression.sh \\"},
	}},
	{job: "merge-strategy", gates: []ciGateExpectation{
		{name: "Self-test the checker's own logic (regression guard)", runExact: "bash scripts/check-merge-strategy.sh --self-test"},
		{name: "Run merge-strategy gate", runExact: "bash scripts/check-merge-strategy.sh", coe: "${{ vars.MERGE_STRATEGY_GATE_REQUIRED != 'true' }}"},
	}},
}

// ciTopologyPythonPin is the SHA-pinned setup-python action that the
// topology-check job (and only that job) keeps after M4.
const ciTopologyPythonPin = "uses: actions/setup-" + "python@a309ff8b426b58ec0e2a45f0f869d46889d02405"

// ciTopologyPinLine matches ciTopologyPythonPin only as an active step-level
// key (a step's first key or a sibling key), optionally followed by a trailing
// comment. A commented-out or otherwise inert occurrence does not match.
var ciTopologyPinLine = regexp.MustCompile(`(?m)^(?:        |      - )` +
	regexp.QuoteMeta(ciTopologyPythonPin) + `(?:[ \t]+#.*)?[ \t]*$`)

// ciGoWiringProblems checks the Set-up-Go ordering and the gate steps of the
// lint, gitignore-append-only and merge-strategy jobs.
func ciGoWiringProblems(text string) []string {
	var problems []string
	for _, je := range ciWiringExpectations {
		block, ok := ciJobBlock(text, je.job)
		if !ok {
			problems = append(problems, fmt.Sprintf("job %q not found", je.job))
			continue
		}
		steps := ciSteps(block)
		goIdx := ciStepIndex(steps, "Set up Go")
		if goIdx < 0 {
			problems = append(problems, fmt.Sprintf("job %q: no step named %q", je.job, "Set up Go"))
		}
		for _, g := range je.gates {
			idx := ciStepIndex(steps, g.name)
			if idx < 0 {
				problems = append(problems, fmt.Sprintf("job %q: missing step %q", je.job, g.name))
				continue
			}
			if goIdx >= 0 && goIdx > idx {
				problems = append(problems, fmt.Sprintf("job %q: %q precedes %q", je.job, g.name, "Set up Go"))
			}
			run, ok := ciStepRun(steps[idx])
			switch {
			case !ok:
				problems = append(problems, fmt.Sprintf("job %q step %q: no run", je.job, g.name))
			case g.runExact != "" && run != g.runExact:
				problems = append(problems, fmt.Sprintf("job %q step %q: run %q, want %q", je.job, g.name, run, g.runExact))
			case g.runContains != "" && !strings.Contains(run, g.runContains):
				problems = append(problems, fmt.Sprintf("job %q step %q: run lacks %q", je.job, g.name, g.runContains))
			}
			coes := ciStepCOEValues(steps[idx])
			switch {
			case g.coe == "" && len(coes) > 0:
				problems = append(problems, fmt.Sprintf("job %q step %q: must not carry continue-on-error (has %q)", je.job, g.name, coes))
			case g.coe != "" && len(coes) == 0:
				problems = append(problems, fmt.Sprintf("job %q step %q: missing continue-on-error %q", je.job, g.name, g.coe))
			case g.coe != "" && (len(coes) != 1 || coes[0] != g.coe):
				problems = append(problems, fmt.Sprintf("job %q step %q: continue-on-error %q, want exactly %q", je.job, g.name, coes, g.coe))
			}
		}
		problems = append(problems, ciJobCOEProblems(text, je.job)...)
	}
	return problems
}

// ciTopologyInstallSteps are topology-check's two hash-pinned install steps,
// in their load-bearing order: the bootstrap step pins the installer toolchain
// that then performs the autoharness install. Each is resolved by name and
// must run exactly its own lock-file command with no continue-on-error.
var ciTopologyInstallSteps = []ciGateExpectation{
	{name: "Bootstrap pip/setuptools/wheel (hash-pinned, --require-hashes)",
		runExact: "pip install --require-hashes --only-binary=:all: --no-deps -r .github/constraints/pip-bootstrap-lock.txt"},
	{name: "Install autoharness (hash-pinned, --require-hashes)",
		runExact: "pip install --require-hashes --only-binary=:all: -r .github/constraints/autoharness-lock.txt"},
}

// ciTopologyProblems checks that topology-check keeps its SHA-pinned
// setup-python inside the step named "Set up Python" and, after that step, its
// two named hash-pinned install steps in order, each running its own lock file.
func ciTopologyProblems(text string) []string {
	block, ok := ciJobBlock(text, "topology-check")
	if !ok {
		return []string{`job "topology-check" not found`}
	}
	var problems []string
	steps := ciSteps(block)
	prev, prevName := ciStepIndex(steps, "Set up Python"), "Set up Python"
	switch {
	case prev < 0:
		problems = append(problems, `topology-check: no step named "Set up Python"`)
	case !ciTopologyPinLine.MatchString("      - " + steps[prev].body):
		// The pin must sit in the named step itself: a pin moved to another
		// step would leave the ordering anchor a no-op.
		problems = append(problems, `topology-check: SHA-pinned setup-python missing from step "Set up Python"`)
	}
	for _, g := range ciTopologyInstallSteps {
		idx := ciStepIndex(steps, g.name)
		if idx < 0 {
			problems = append(problems, fmt.Sprintf("topology-check: missing step %q", g.name))
			continue
		}
		if prev >= 0 && idx < prev {
			problems = append(problems, fmt.Sprintf("topology-check: %q precedes %q", g.name, prevName))
		}
		prev, prevName = idx, g.name
		if run, ok := ciStepRun(steps[idx]); !ok || run != g.runExact {
			problems = append(problems, fmt.Sprintf("topology-check step %q: run %q, want %q", g.name, run, g.runExact))
		}
		if coes := ciStepCOEValues(steps[idx]); len(coes) > 0 {
			problems = append(problems, fmt.Sprintf("topology-check step %q: must not carry continue-on-error (has %q)", g.name, coes))
		}
	}
	return append(problems, ciJobCOEProblems(text, "topology-check")...)
}

func ciWiringProblems(text string) []string {
	return append(ciGoWiringProblems(text), ciTopologyProblems(text)...)
}

// ciMoveSetUpGoAfter returns text with job `job`'s `Set up Go` step moved to
// just after step `after`, to model a reordering mutation.
func ciMoveSetUpGoAfter(t *testing.T, text, job, after string) string {
	t.Helper()
	return ciMoveStepAfter(t, text, job, "Set up Go", after)
}

// ciMoveStepAfter returns text with job `job`'s step `name` moved to just
// after the later step `after`, to model a reordering mutation.
func ciMoveStepAfter(t *testing.T, text, job, name, after string) string {
	t.Helper()
	s, e, ok := ciJobSpan(text, job)
	if !ok {
		t.Fatalf("job %q not found", job)
	}
	block := text[s:e]
	head := ciStepSplit.Split(block, -1)[0]
	steps := ciSteps(block)
	goIdx, afterIdx := ciStepIndex(steps, name), ciStepIndex(steps, after)
	if goIdx < 0 || afterIdx < 0 || goIdx > afterIdx {
		t.Fatalf("job %q: cannot move %q after %q", job, name, after)
	}
	var order []ciStep
	for i, st := range steps {
		if i != goIdx {
			order = append(order, st)
		}
		if i == afterIdx {
			order = append(order, steps[goIdx])
		}
	}
	var b strings.Builder
	b.WriteString(head)
	for _, st := range order {
		b.WriteString("      - ")
		b.WriteString(st.body)
	}
	return text[:s] + b.String() + text[e:]
}

// ciRemoveStep returns text with the named step of job `job` removed.
func ciRemoveStep(t *testing.T, text, job, name string) string {
	t.Helper()
	s, e, ok := ciJobSpan(text, job)
	if !ok {
		t.Fatalf("job %q not found", job)
	}
	block := text[s:e]
	head := ciStepSplit.Split(block, -1)[0]
	steps := ciSteps(block)
	if ciStepIndex(steps, name) < 0 {
		t.Fatalf("job %q: step %q not found", job, name)
	}
	var b strings.Builder
	b.WriteString(head)
	for _, st := range steps {
		if st.name == name {
			continue
		}
		b.WriteString("      - ")
		b.WriteString(st.body)
	}
	return text[:s] + b.String() + text[e:]
}

func TestCIWiring_Live(t *testing.T) {
	if p := ciWiringProblems(readCIWorkflow(t)); len(p) != 0 {
		t.Fatalf("live %s:\n  %s", ciWorkflowRel, strings.Join(p, "\n  "))
	}
}

func TestCIWiring_ParserLocatesStepsByName(t *testing.T) {
	block, ok := ciJobBlock(readCIWorkflow(t), "lint")
	if !ok {
		t.Fatal("lint job not found")
	}
	steps := ciSteps(block)
	if ciStepIndex(steps, "Checkout") != 0 {
		t.Errorf("lint step 0 = %q, want Checkout", steps[0].name)
	}
	if ciStepIndex(steps, "no such step") != -1 {
		t.Error("unknown step name resolved")
	}
}

func TestCIWiring_MutatedInputsAreRed(t *testing.T) {
	live := readCIWorkflow(t)
	cases := map[string]func() string{
		"lint self-test step removed": func() string {
			return ciRemoveStep(t, live, "lint", "Run retired-architecture self-test")
		},
		"lint Set up Go removed": func() string {
			return ciRemoveStep(t, live, "lint", "Set up Go")
		},
		"merge-strategy gate removed": func() string {
			return ciRemoveStep(t, live, "merge-strategy", "Run merge-strategy gate")
		},
		"lint Set up Go after a gate": func() string {
			return ciMoveSetUpGoAfter(t, live, "lint", "Run write-path-precondition gate")
		},
		"gitignore Set up Go after self-test": func() string {
			return ciMoveSetUpGoAfter(t, live, "gitignore-append-only", "Self-test the un-ignore regression checker's own logic")
		},
		"merge-strategy Set up Go after self-test": func() string {
			return ciMoveSetUpGoAfter(t, live, "merge-strategy", "Self-test the checker's own logic (regression guard)")
		},
		"continue-on-error added to integrity self-test": func() string {
			return mustMutate(t, live,
				"        run: bash scripts/check-retired-architecture.sh --self-test-integrity\n",
				"        continue-on-error: true\n        run: bash scripts/check-retired-architecture.sh --self-test-integrity\n")
		},
		"continue-on-error added to merge-strategy self-test": func() string {
			return mustMutate(t, live,
				"        run: bash scripts/check-merge-strategy.sh --self-test\n",
				"        continue-on-error: true\n        run: bash scripts/check-merge-strategy.sh --self-test\n")
		},
		"advisory expression inverted": func() string {
			return mustMutate(t, live, "${{ vars.WRITE_PATH_GATE_ADVISORY == 'true' }}", "${{ vars.WRITE_PATH_GATE_ADVISORY != 'true' }}")
		},
		"gate runs the wrong wrapper": func() string {
			return mustMutate(t, live, "        run: bash scripts/check-write-path-precondition.sh\n", "        run: bash scripts/other.sh\n")
		},
		"topology pin changed": func() string {
			return strings.ReplaceAll(live, ciTopologyPythonPin, ciTopologyPythonPin[:len(ciTopologyPythonPin)-1]+"0")
		},
		"topology pin commented out": func() string {
			return strings.ReplaceAll(live, "        "+ciTopologyPythonPin, "        # "+ciTopologyPythonPin)
		},
		"topology pin relocated out of Set up Python": func() string {
			m := mustMutate(t, live, "- name: Set up Python\n        "+ciTopologyPythonPin,
				"- name: Set up Python\n        run: echo no-op\n        #")
			_, e, ok := ciJobSpan(m, "topology-check")
			if !ok {
				t.Fatal("topology-check not found")
			}
			pre := m[:e]
			if !strings.HasSuffix(pre, "\n") {
				pre += "\n"
			}
			return pre + "      - name: Relocated setup\n        " + ciTopologyPythonPin + "\n" + m[e:]
		},
		"topology hash-pinned pip step removed": func() string {
			return ciRemoveStep(t, live, "topology-check", "Install autoharness (hash-pinned, --require-hashes)")
		},
		"topology bootstrap step removed": func() string {
			return ciRemoveStep(t, live, "topology-check", "Bootstrap pip/setuptools/wheel (hash-pinned, --require-hashes)")
		},
		"topology bootstrap runs the autoharness lock": func() string {
			return mustMutate(t, live,
				"--no-deps -r .github/constraints/pip-bootstrap-lock.txt\n",
				"--no-deps -r .github/constraints/autoharness-lock.txt\n")
		},
		"topology install drops --only-binary": func() string {
			return mustMutate(t, live,
				"pip install --require-hashes --only-binary=:all: -r .github/constraints/autoharness-lock.txt\n",
				"pip install --require-hashes -r .github/constraints/autoharness-lock.txt\n")
		},
		"topology bootstrap after autoharness install": func() string {
			return ciMoveStepAfter(t, live, "topology-check",
				"Bootstrap pip/setuptools/wheel (hash-pinned, --require-hashes)",
				"Install autoharness (hash-pinned, --require-hashes)")
		},
		"topology bootstrap before setup-python": func() string {
			return ciMoveStepAfter(t, live, "topology-check", "Set up Python",
				"Bootstrap pip/setuptools/wheel (hash-pinned, --require-hashes)")
		},
		"continue-on-error added to topology install": func() string {
			return mustMutate(t, live,
				"        run: |\n          pip install --require-hashes --only-binary=:all: -r .github/constraints/autoharness-lock.txt\n",
				"        continue-on-error: true\n        run: |\n          pip install --require-hashes --only-binary=:all: -r .github/constraints/autoharness-lock.txt\n")
		},
		"continue-on-error as first key of integrity self-test": func() string {
			return mustMutate(t, live,
				"      - name: Run retired-architecture self-test\n",
				"      - continue-on-error: true\n        name: Run retired-architecture self-test\n")
		},
		"double-quoted continue-on-error on merge-strategy self-test": func() string {
			return mustMutate(t, live,
				"        run: bash scripts/check-merge-strategy.sh --self-test\n",
				"        \"continue-on-error\": true\n        run: bash scripts/check-merge-strategy.sh --self-test\n")
		},
		"single-quoted continue-on-error on topology install": func() string {
			return mustMutate(t, live,
				"        run: |\n          pip install --require-hashes --only-binary=:all: -r .github/constraints/autoharness-lock.txt\n",
				"        'continue-on-error' : true\n        run: |\n          pip install --require-hashes --only-binary=:all: -r .github/constraints/autoharness-lock.txt\n")
		},
		"explicit-key continue-on-error on write-path self-test": func() string {
			return mustMutate(t, live,
				"        run: bash scripts/check-write-path-precondition.sh --self-test-integrity\n",
				"        ? continue-on-error\n        : true\n        run: bash scripts/check-write-path-precondition.sh --self-test-integrity\n")
		},
		"duplicate continue-on-error on advisory gate": func() string {
			return mustMutate(t, live,
				"        continue-on-error: ${{ vars.RETIRED_ARCH_GATE_ADVISORY == 'true' }}\n",
				"        continue-on-error: ${{ vars.RETIRED_ARCH_GATE_ADVISORY == 'true' }}\n        continue-on-error: true\n")
		},
		"job-level continue-on-error in lint": func() string {
			return mustMutate(t, live, "\n  lint:\n", "\n  lint:\n    continue-on-error: true\n")
		},
		"quoted job-level continue-on-error in gitignore-append-only": func() string {
			return mustMutate(t, live, "\n  gitignore-append-only:\n", "\n  gitignore-append-only:\n    'continue-on-error': true\n")
		},
		"job-level continue-on-error in topology-check": func() string {
			return mustMutate(t, live, "\n  topology-check:\n", "\n  topology-check:\n    \"continue-on-error\": true\n")
		},
		"job-level continue-on-error after merge-strategy steps": func() string {
			_, e, ok := ciJobSpan(live, "merge-strategy")
			if !ok {
				t.Fatal("merge-strategy not found")
			}
			pre := live[:e]
			if !strings.HasSuffix(pre, "\n") {
				pre += "\n"
			}
			return pre + "    continue-on-error: true\n" + live[e:]
		},
		"job removed": func() string {
			return mustMutate(t, live, "\n  merge-strategy:\n", "\n  merge-strategy-renamed:\n")
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			mutated := mutate()
			if mutated == live {
				t.Fatal("mutation was a no-op")
			}
			if p := ciWiringProblems(mutated); len(p) == 0 {
				t.Fatal("mutated ci.yml unexpectedly passed")
			}
		})
	}
}

// TestCIWiring_EquivalentCOEFormsAreGreen checks that a required
// continue-on-error written as the step's first key, or a step-level key in
// an unprotected step, is still accepted.
func TestCIWiring_EquivalentCOEFormsAreGreen(t *testing.T) {
	live := readCIWorkflow(t)
	coe := "        continue-on-error: ${{ vars.RETIRED_ARCH_GATE_ADVISORY == 'true' }}\n"
	cases := map[string]string{
		"advisory gate continue-on-error as first key": mustMutate(t, mustMutate(t, live, coe, ""),
			"      - name: Run retired-architecture gate\n",
			"      - continue-on-error: ${{ vars.RETIRED_ARCH_GATE_ADVISORY == 'true' }}\n        name: Run retired-architecture gate\n"),
		"step-level continue-on-error on lint Checkout": mustMutate(t, live,
			"      - name: Checkout\n        uses: actions/checkout@df4cb1c069e1874edd31b4311f1884172cec0e10 # v6.0.3\n        with:\n          persist-credentials: false\n      - name: Set up Go\n        uses: actions/setup-go@d35c59abb061a4a6fb18e82ac0862c26744d6ab5 # v5.5.0\n        with:\n          go-version: '1.26.x'\n          cache: true\n      - name: Run retired-architecture gate\n",
			"      - continue-on-error: false\n        name: Checkout\n        uses: actions/checkout@df4cb1c069e1874edd31b4311f1884172cec0e10 # v6.0.3\n        with:\n          persist-credentials: false\n      - name: Set up Go\n        uses: actions/setup-go@d35c59abb061a4a6fb18e82ac0862c26744d6ab5 # v5.5.0\n        with:\n          go-version: '1.26.x'\n          cache: true\n      - name: Run retired-architecture gate\n"),
	}
	for name, mutated := range cases {
		t.Run(name, func(t *testing.T) {
			if mutated == live {
				t.Fatal("mutation was a no-op")
			}
			if p := ciWiringProblems(mutated); len(p) != 0 {
				t.Fatalf("equivalent form flagged:\n  %s", strings.Join(p, "\n  "))
			}
		})
	}
}

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
	ciJobKeyLine  = regexp.MustCompile(`(?m)^  [A-Za-z0-9_-]+:[ \t]*$`)
	ciStepSplit   = regexp.MustCompile(`(?m)^      - `)
	ciStepCOE     = regexp.MustCompile(`(?m)^        continue-on-error:[ \t]*(.*?)[ \t]*$`)
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

// ciStepCOEValue returns a step's continue-on-error value, if any.
func ciStepCOEValue(st ciStep) (string, bool) {
	m := ciStepCOE.FindStringSubmatch(st.body)
	if m == nil {
		return "", false
	}
	return m[1], true
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
			coe, has := ciStepCOEValue(steps[idx])
			switch {
			case g.coe == "" && has:
				problems = append(problems, fmt.Sprintf("job %q step %q: must not carry continue-on-error (has %q)", je.job, g.name, coe))
			case g.coe != "" && !has:
				problems = append(problems, fmt.Sprintf("job %q step %q: missing continue-on-error %q", je.job, g.name, g.coe))
			case g.coe != "" && coe != g.coe:
				problems = append(problems, fmt.Sprintf("job %q step %q: continue-on-error %q, want %q", je.job, g.name, coe, g.coe))
			}
		}
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
// setup-python and, after it, its two named hash-pinned install steps in
// order, each running its own lock file.
func ciTopologyProblems(text string) []string {
	block, ok := ciJobBlock(text, "topology-check")
	if !ok {
		return []string{`job "topology-check" not found`}
	}
	var problems []string
	if !ciTopologyPinLine.MatchString(block) {
		problems = append(problems, "topology-check: SHA-pinned setup-python missing")
	}
	steps := ciSteps(block)
	prev, prevName := ciStepIndex(steps, "Set up Python"), "Set up Python"
	if prev < 0 {
		problems = append(problems, `topology-check: no step named "Set up Python"`)
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
		if coe, has := ciStepCOEValue(steps[idx]); has {
			problems = append(problems, fmt.Sprintf("topology-check step %q: must not carry continue-on-error (has %q)", g.name, coe))
		}
	}
	return problems
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

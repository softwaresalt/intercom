package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// readSurface reads a fixed repo-relative instruction file, slices it from
// startAnchor (inclusive) to endAnchor (exclusive), and returns the
// whitespace-collapsed text. An empty anchor leaves that side unsliced. Any
// read, path, or anchor error is a t.Fatalf, never the not-implemented marker.
func readSurface(t *testing.T, relpath, startAnchor, endAnchor string) string {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repoRoot(t), filepath.FromSlash(relpath)))
	if err != nil {
		t.Fatalf("read surface %s: %v", relpath, err)
	}
	text := strings.Join(strings.Fields(string(raw)), " ")
	start, end := strings.Join(strings.Fields(startAnchor), " "), strings.Join(strings.Fields(endAnchor), " ")
	if start != "" {
		i := strings.Index(text, start)
		if i < 0 {
			t.Fatalf("start anchor %q not found in %s", start, relpath)
		}
		text = text[i:]
	}
	if end != "" {
		i := strings.Index(text[len(start):], end)
		if i < 0 {
			t.Fatalf("end anchor %q not found after start anchor in %s", end, relpath)
		}
		text = text[:len(start)+i]
	}
	return text
}

// pin asserts that the whitespace-collapsed clause is present or absent in
// subject. A mismatch is the wording-contract red marker.
func pin(t *testing.T, subject, clause, id string, wantPresent bool) {
	t.Helper()
	clause = strings.Join(strings.Fields(clause), " ")
	if (countOccurrences(subject, clause) > 0) == wantPresent {
		return
	}
	state := "still present"
	if wantPresent {
		state = "absent"
	}
	panic(fmt.Sprintf("not implemented: pinned wording %s: %s", state, id))
}

const harnessArchitectSkill = ".github/skills/harness-architect/SKILL.md"

func TestHarnessArchitect_ExplicitActiveTaskInScope(t *testing.T) {
	whole := readSurface(t, harnessArchitectSkill, "", "")
	desc := readSurface(t, harnessArchitectSkill, `description: "Comma-separated task IDs`, "required:")
	intro := readSurface(t, harnessArchitectSkill, "# Harness Architect Skill", "## Purpose")
	broadcast := readSurface(t, harnessArchitectSkill, "| Tasks loaded", "| Codebase analyzed")
	inputs := readSurface(t, harnessArchitectSkill, "* `${input:tasks}`", "## Workflow")
	step1 := readSurface(t, harnessArchitectSkill, "### Step 1:", "### Step 2:")

	pin(t, desc, "defaults to all `queued` tasks", "PW-6 description", true)
	pin(t, intro, "a feature's or chore's in-scope work items", "PW-6 line 18", true)
	pin(t, broadcast, "Tasks in scope: {task_count}", "PW-6 broadcast", true)
	pin(t, inputs, "When omitted, use all `queued` tasks under the feature. "+
		"An `active` task is in scope only when it is listed explicitly in `${input:tasks}`.", "PW-6 line 50", true)
	pin(t, step1, "### Step 1: Select the in-scope task set", "PW-6 step 1 heading", true)
	pin(t, step1, "Load the feature or chore and its in-scope descendants (`queued`, plus explicitly listed `active`)", "PW-6 step 1 item 1", true)
	pin(t, step1, "Exclude blocked, done, and archived work items. "+
		"An explicitly listed `active` task is in scope and MUST NOT be excluded as non-ready.", "PW-6 step 1 item 3", true)

	pin(t, whole, "Exclude blocked, done, or otherwise non-ready work items", "retired non-ready exclusion", false)
}

const shipAgent = ".github/agents/_ship.agent.md"

func TestStep2_PreambleDiscriminatorAndSingleDerivation(t *testing.T) {
	step2 := readSurface(t, shipAgent, "### Step 2: Harness Generation", "### Step 3: Build Ready Queue")
	step3Item1 := readSurface(t, shipAgent, "1. **Shipment-manifest executable-set derivation", "2. List all tasks with")

	pin(t, step2, "Ensure every task in the selected batch (item 1) has a passing test harness before that task's implementation begins. "+
		"On the shipment route this step runs just-in-time per task, invoked from Step 4.1 (item 3): "+
		"at this position execute item 1 only (halt if the batch is empty) and items 2-4 are executed per task from Step 4.1; "+
		"on any other route it runs once, up front, not in a loop.", "PW-5a preamble", true)
	pin(t, step2, "Select the batch. **Route discriminator**: the *shipment route* is any run in which a `shipment_id` is recorded "+
		"as the session scope (Step 0.5 primary item 5 or fallback item 4) or in which a task reaches Step 4.1 already `active`, "+
		"and any ambiguity is treated as the shipment route; on that route the Step 0.5 claim has already activated manifest tasks. "+
		"On the shipment route the batch is the Step 3 item 1 derived executable set, admitting both `queued` and `active` members; "+
		"Ship evaluates that single read-only derivation FIRST, once, and Step 3 reuses it.", "PW-5b route discriminator", true)
	pin(t, step2, "On any other route, list all tasks for the target feature or chore that are in `queued` status.", "PW-5b legacy sentence", true)
	pin(t, step3Item1, "(This is the single derivation Step 2 item 1 evaluates first on the shipment route; it is not recomputed.)", "PW-5g single derivation note", true)

	pin(t, step2, "This step runs once, up front — not in a loop.", "retired once-up-front sentence", false)
}

func TestStep2_JustInTimeTaskPartitionGuardAndClosingParagraph(t *testing.T) {
	step2 := readSurface(t, shipAgent, "### Step 2: Harness Generation", "### Step 3: Build Ready Queue")

	pin(t, step2, "Partition the selected batch (on the shipment route: the current task):", "PW-5c partition", true)
	pin(t, step2, "On the shipment route, this item is invoked from Step 4.1: scaffold ONE task at a time in dependency order, "+
		"immediately before that task's Step 4.1, passing that task's ID explicitly as `tasks` "+
		"(a whole-batch up-front scaffold would leave several red harnesses at once and break the Step 4.3 full-suite gate after the first task); "+
		"do not proceed until Compilation PASS, Red Phase CONFIRMED and the `harness-ready` label are recorded. "+
		"On any other route, invoke the skill for the batch as before.", "PW-5d per-task invocation", true)
	pin(t, step2, "On the shipment route confirm the current task now carries the `harness-ready` label; "+
		"on any other route confirm every task in the batch does.", "PW-5e confirmation", true)
	pin(t, step2, "An empty selected batch on the shipment route is a halt, never a vacuous pass.", "PW-5e empty batch halt", true)
	pin(t, step2, "broadcast `[SHIP] Invoking harness-architect skill` when invoking the skill (per task from Step 4.1 on the shipment route).", "PW-5f broadcast", true)
	pin(t, step2, "On the shipment route the batch comes from the Step 3 item 1 derivation, never from a queued-only queue operation.", "PW-5f2 closing paragraph", true)

	pin(t, step2, "confirm every queued task now carries", "retired vacuous confirmation", false)
}

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

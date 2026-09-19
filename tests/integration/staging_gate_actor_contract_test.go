package integration

// Harness for 027.001-T ("Bind Step 1.5 item 3 to actors and remove
// direct-push arm"), scoped by shipment 024-S. This is a CHARACTERIZATION
// harness (per the shipped precedent tests/integration/stage_branch_gate_test.go,
// 018.008-T, PR #58): it reads the installed contract surface directly --
//   - .github/agents/_orchestrator.agent.md, Step 1.5 ("Staging Artifact
//     Merge Gate") through the start of Step 2 ("Route to Ship")
//
// and asserts the corrected properties named in AC-1 through AC-9 per plan
// docs/plans/2026-09-18-intercom-go-staging-artifact-handoff-actor-plan.md
// section 5. There is no Go production code implementing this contract (the
// contract is prompt/policy text), so no companion production stub is
// created outside this single test file, per harness-architect Step 4 item
// 3's own rationale.
//
// P-004 RED PHASE: the sole expected marker lives in the unexported helper
// below and is invoked unconditionally by TestStagingGateActorContract.
// Zero skips, zero build tags, zero env/platform gates. All assertions are
// bounded to the Step 1.5 slice, delimited (as substrings, not line
// equality -- the installed heading carries a "(NON-NEGOTIABLE)" suffix and
// the Step 2 heading carries a "(when a queued shipment is ready)" suffix
// the delimiters below omit) by "### Step 1.5: Staging Artifact Merge Gate"
// and "### Step 2: Route to Ship", so a future Amendment Log row or
// provenance note quoting retired text cannot false-positive.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sliceBetween returns the substring of content strictly between the first
// occurrence of startNeedle and the first occurrence of endNeedle found
// after startNeedle. It fails the test if either delimiter is absent, per
// the plan's slice-bounding requirement (H16: delimiters matched as
// substrings, not line equality).
func sliceBetween(t *testing.T, content, startNeedle, endNeedle, row string) string {
	t.Helper()
	startIdx := strings.Index(content, startNeedle)
	if startIdx < 0 {
		t.Fatalf("row %s: slice start delimiter not found: %q", row, startNeedle)
	}
	endIdx := strings.Index(content[startIdx:], endNeedle)
	if endIdx < 0 {
		t.Fatalf("row %s: slice end delimiter not found after start: %q", row, endNeedle)
	}
	return content[startIdx : startIdx+endIdx]
}

// assertStagingGateActorContract carries the full 14-row AC-1..AC-9
// characterization assertion set (9 red rows, 1 fence row, 4 guard rows)
// over the installed Step 1.5 slice of .github/agents/_orchestrator.agent.md,
// per plan section 5.
func assertStagingGateActorContract(t *testing.T) {
	t.Helper()

	root := repoRoot(t)
	orchestratorPath := filepath.Join(root, ".github", "agents", "_orchestrator.agent.md")

	orchestratorBytes, err := os.ReadFile(orchestratorPath)
	if err != nil {
		t.Fatalf("reading %s: %v", orchestratorPath, err)
	}
	orchestrator := string(orchestratorBytes)

	slice := sliceBetween(t, orchestrator,
		"### Step 1.5: Staging Artifact Merge Gate",
		"### Step 2: Route to Ship",
		"slice-bound")

	// --- Row 1 (AC-2): the direct-push-first arm is absent from the slice ---
	assertAbsent(t, slice, "Attempt a direct push to `main` first", "1")

	// --- Row 2 (AC-2): the branch-protection-handling sub-step is absent ---
	assertAbsent(t, slice, "Branch protection handling", "2")

	// --- Row 3 (AC-3): the stale chore/stage-{shipment_id} branch literal is absent ---
	assertAbsent(t, slice, "chore/stage-{shipment_id}", "3")

	// --- Row 4 (AC-1): item 3 is bound to named actors with no default-branch write ---
	assertPresent(t, slice, "ACTOR-BOUND, NO DEFAULT-BRANCH WRITE", "4")

	// --- Row 5 (AC-4): the uncommitted-artifact case halts with a tokenized message ---
	assertPresent(t, slice, "STAGING_GATE_UNCOMMITTED", "5")

	// --- Row 6 (AC-5): the local-default-branch-ahead case halts with a tokenized message ---
	assertPresent(t, slice, "STAGING_GATE_LOCAL_MAIN_AHEAD", "6")

	// --- Row 7 (AC-6): the operator arm emits an awaiting-operator token ---
	assertPresent(t, slice, "STAGING_GATE_AWAITING_OPERATOR", "7")

	// --- Row 8 (AC-6): the operator wait is bounded by a timeout token ---
	assertPresent(t, slice, "STAGING_GATE_OPERATOR_TIMEOUT", "8")

	// --- Row 9 (AC-7): the no-route terminal is tokenized ---
	assertPresent(t, slice, "STAGING_GATE_NO_ROUTE", "9")

	// --- Row 10 (AC-9, fence): handback-record fields are not defined here ---
	assertAbsent(t, slice, "stage_artifact_paths", "10")

	// --- Guard 1 (AC-8): item 1's uncommitted-artifact detection command is preserved ---
	assertPresent(t, slice, "git status --short -- .backlogit/", "G1")

	// --- Guard 2 (AC-8): item 2's unpushed-commit detection command is preserved ---
	assertPresent(t, slice, "git log origin/main..main --oneline", "G2")

	// --- Guard 3 (AC-8): item 4's remote-manifest verification command is preserved ---
	assertPresent(t, slice, "git show origin/main:.backlogit/queue/{shipment_id}.md", "G3")

	// --- Guard 4 (AC-8): item 4's fail-closed halt token is preserved ---
	assertPresent(t, slice, "STAGING_GATE_FAIL: shipment manifest", "G4")
}

// TestStagingGateActorContract is the sole exported test function for this
// release unit's harness footprint (027-F has exactly one queued task,
// 027.001-T). It invokes assertStagingGateActorContract unconditionally --
// no t.Run, no t.Skip, no build tag, no env-var/platform gate, no
// activation manifest, no selector flag.
func TestStagingGateActorContract(t *testing.T) {
	assertStagingGateActorContract(t)
}

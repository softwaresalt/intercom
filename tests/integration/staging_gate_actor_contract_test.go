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

// TestDarkStagingPublicationContract checks the installed authorization boundary,
// not a runtime enforcement engine: these agents execute the policy prose.
func TestDarkStagingPublicationContract(t *testing.T) {
	files := make(map[string]string)
	for name, path := range map[string]string{
		"policy":       filepath.Join(".github", "policies", "workflow-policies.md"),
		"stage":        filepath.Join(".github", "agents", "_stage.agent.md"),
		"orchestrator": filepath.Join(".github", "agents", "_orchestrator.agent.md"),
	} {
		data, err := os.ReadFile(filepath.Join(repoRoot(t), path))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		files[name] = strings.ReplaceAll(string(data), "\r\n", "\n")
	}
	gate := sliceBetween(t, files["orchestrator"],
		"### Step 1.5: Staging Artifact Merge Gate", "### Step 2: Route to Ship", "dark-gate")
	p010 := policySection(t, files["policy"], "## P-010:")
	p017 := policySection(t, files["policy"], "## P-017:")

	t.Run("recorded-dark-only-owned-committed-artifacts", func(t *testing.T) {
		grant := sliceBetween(t, p010, "**Orchestrator MAY", "**Ship MUST NOT**", "dark-grant")
		dark := sliceBetween(t, gate, "**Recorded-dark publication (actor: Orchestrator).**",
			"d. **Merge", "dark-publication")
		for surface, text := range map[string]string{"P-010 grant": grant, "handoff": dark} {
			for _, needle := range []string{
				"DARK_MODE_ACTIVE", "`merge_approval_pre_authorized=true`", "recorded scope",
				"already committed", "Stage Step 1.9", "ownership", "source, test, or config",
			} {
				assertPresent(t, text, needle, surface)
			}
		}
		for _, needle := range []string{
			"publish that Stage artifact branch", "open its staging PR",
			"MUST NOT commit Stage artifacts", "MUST NOT claim shipments",
			"NO DIRECT DEFAULT-BRANCH PUSH", "STAGING_GATE_SCOPE_FAIL",
			"a clean worktree and synchronized local default do not prove",
		} {
			assertPresent(t, gate, needle, "bounded-publication")
		}
		assertAbsent(t, p010, "the Orchestrator's role statement above is unchanged", "retired-conflict")
		assertAbsent(t, files["orchestrator"], "You do NOT write code or create PRs yourself.", "retired-conflict")
	})

	t.Run("normal-mode-operator-and-stage-still-forbidden", func(t *testing.T) {
		normal := sliceBetween(t, gate, "**Normal/non-dark publication (actor: OPERATOR).**",
			"**Recorded-dark publication (actor: Orchestrator).**", "normal-publication")
		for _, needle := range []string{
			"without recorded dark merge authorization", "STAGING_GATE_AWAITING_OPERATOR",
			"STAGING_GATE_OPERATOR_TIMEOUT", "never act in the operator's place",
		} {
			assertPresent(t, normal, needle, "normal-operator")
		}
		row := sliceBetween(t, files["stage"], "| PR |", "\n\n", "stage-pr-row")
		assertPresent(t, row, "never Stage", "stage-pr")
		assertPresent(t, row, "Create, push, or merge pull requests |", "stage-pr")
		assertPresent(t, row, "P-010 recorded-dark staging publication exception", "stage-pr")
		assertAbsent(t, row, "no Orchestrator authority to do so", "stage-pr")
		stageForbidden := sliceBetween(t, p010, "**Stage MUST NOT**", "**Stage MAY**", "stage-forbidden")
		assertPresent(t, stageForbidden, "Create, push, or merge pull requests", "stage-forbidden")
		assertPresent(t, stageForbidden, "Run build systems, test suites, or linters", "stage-forbidden")
	})

	t.Run("approval-source-and-independent-safety-gates", func(t *testing.T) {
		activation := sliceBetween(t, files["orchestrator"],
			"| `merge_approval_pre_authorized`", "| `stop_conditions`", "activation")
		for _, needle := range []string{
			"explicit P-017 trigger", "scope clarification", "set `true`",
			"explicitly withholds", "ambiguous", "set `false`",
			"`admin_fallback_pre_authorized`",
		} {
			assertPresent(t, activation, needle, "activation")
		}
		for _, needle := range []string{
			"current-HEAD local review", "P-009", "P-014", "P-018",
			"required CI/checks", "autoharness gate copilot-review",
			"no redundant per-PR approval", "admin fallback", "explicitly authorized",
		} {
			assertPresent(t, gate, needle, "merge-gates")
		}
		merge := sliceBetween(t, gate, "d. **Merge", "e. **After merge", "merge")
		for _, needle := range []string{
			"independent current-HEAD local review", "require PASS for the current HEAD",
			"any change invalidates readiness",
		} {
			assertPresent(t, merge, needle, "current-head-gates")
		}
		remote := gate[indexOf(t, gate, "4. Verify the scoped staging artifacts", "remote-versions"):]
		for _, needle := range []string{
			"compare Git blob IDs", "{reviewed_head}:{path}", "origin/{default_branch}:{path}",
			"deleted paths must be absent", "Missing or mismatched content halts",
			"Manifest existence alone never authorizes Ship routing",
		} {
			assertPresent(t, remote, needle, "remote-versions")
		}
		for _, needle := range []string{
			"explicit dark-mode activation or scope clarification",
			"no redundant per-PR approval", "current-HEAD", "P-018",
		} {
			assertPresent(t, p017, needle, "P-017-approval")
		}
		assertPresent(t, p010, "per-closure operator approval, not waived in dark mode", "destructive-gate")
		assertPresent(t, p017, "Destructive actions still require separate per-action approval", "destructive-gate")
	})
}

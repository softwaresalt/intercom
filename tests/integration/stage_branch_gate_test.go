package integration

// Harness for 018.008-T ("Gate Stage artifact mutation behind a dedicated
// branch"), scoped by shipment 017-S. This is a CHARACTERIZATION harness
// (per the task's HARNESS FOOTPRINT note): it reads the two installed
// contract surfaces directly --
//   - .github/policies/workflow-policies.md (P-010)
//   - .github/agents/_stage.agent.md (Role Boundary + Step 1.9 gate)
//
// and asserts the corrected properties named in AC-1 through AC-12. There is
// no Go production code implementing this contract (the contract is
// prompt/policy text), so no companion production stub is created outside
// this single test file, per harness-architect Step 4 item 3's own
// rationale (a stub exists only so the module compiles while a test calls
// not-yet-existing production code; a test that only reads files needs no
// such stub).
//
// P-004 RED PHASE (H0, AC-13): the sole expected marker lives in the
// unexported helper below and is invoked unconditionally by
// TestStageBranchGate_ContractCorrected. Zero skips, zero build tags, zero
// env/platform gates (AC-14).

import (
	"os"
	"path/filepath"
	"testing"
)

// assertStageBranchGateContract carries the full AC-1..AC-12 characterization
// assertions over the two installed contract surfaces.
func assertStageBranchGateContract(t *testing.T) {
	t.Helper()

	root := repoRoot(t)
	policyPath := filepath.Join(root, ".github", "policies", "workflow-policies.md")
	stagePath := filepath.Join(root, ".github", "agents", "_stage.agent.md")

	policyBytes, err := os.ReadFile(policyPath)
	if err != nil {
		t.Fatalf("reading %s: %v", policyPath, err)
	}
	stageBytes, err := os.ReadFile(stagePath)
	if err != nil {
		t.Fatalf("reading %s: %v", stagePath, err)
	}
	policy := string(policyBytes)
	stage := string(stageBytes)

	// --- AC-1: no default-branch commit grant remains anywhere in P-010 ---
	assertAbsent(t, policy, "Commit backlog and planning artifacts (on the default branch or a dedicated chore/admin branch)", "AC-1")
	assertAbsent(t, policy, "on the default branch or a dedicated chore/admin branch", "AC-1")
	assertPresent(t, policy, "Commit backlog and planning artifacts on that dedicated Stage artifact branch only — never on the default branch", "AC-1")

	// --- AC-2: actor-named Stage MUST NOT bullet symmetric with Ship's; Stage MAY create/checkout grant ---
	assertPresentAtLeast(t, policy, "- Commit or push directly to `main`", 2, "AC-2")
	assertPresent(t, policy, "- Create and check out a dedicated Stage artifact branch (`chore/stage-{scope-slug}`)", "AC-2")

	// --- AC-3: P-016/P-001 clarifying sentence; no Orchestrator authority granted; Amendment Log 1.25.0 ---
	assertPresent(t, policy, "it is therefore neither an implementation branch under P-016 nor a release unit under P-001", "AC-3")
	assertPresent(t, policy, "This clarification does not grant, imply, or restate any Orchestrator authority to commit Stage artifacts or push the default branch", "AC-3")
	assertAbsent(t, policy, "Orchestrator MAY commit Stage artifacts", "AC-3")
	assertAbsent(t, policy, "Orchestrator MAY push the default branch", "AC-3")
	assertPresent(t, policy, "| 1.25.0", "AC-3")
	assertPresent(t, policy, "Corrected P-010", "AC-3")

	// --- AC-4: P-016 spike/research exception paragraph byte-for-byte unchanged ---
	const p016ExceptionParagraph = "**Allowed exception (Stage spike/research only)**: Stage MAY create or use a separate worktree only for an explicit, time-boxed spike or research investigation during staging. That worktree MUST NOT be used for implementation, template/source/config mutation, shipment claim, PR preparation, or Ship execution. Stage MUST record the spike/research context and clean up the worktree or hand off findings before Ship execution begins."
	assertPresentExactly(t, policy, p016ExceptionParagraph, "AC-4")

	// --- AC-5: Role Boundary Git row + PR row ---
	assertPresent(t, stage, "Create and check out a dedicated Stage artifact branch (`chore/stage-{scope-slug}`); commit backlog/planning artifacts on that branch only", "AC-5")
	assertPresent(t, stage, "create/use an explicit, time-boxed spike/research worktree only for staging investigation", "AC-5")
	assertPresent(t, stage, "Create or checkout feature/chore branches for code execution; create/use parallel implementation branches or worktrees", "AC-5")
	assertPresent(t, stage, "The operator — never Stage — pushes the Stage artifact branch and opens/approves the resulting merge-commit staging PR", "AC-5")
	assertPresent(t, stage, "no Orchestrator authority to do so is granted or assumed by this release unit", "AC-5")
	assertPresent(t, stage, "| PR | The operator", "AC-5")
	assertPresent(t, stage, "Create, push, or merge pull requests |", "AC-5")

	// --- AC-6: Step Sequence Contract checklist line, anchored between Step 1.8 and Step 2 ---
	assertOrderedBefore(t, stage, "[ ] Step 1.8 — Learnings retrieval", "[ ] Step 1.9 — Stage artifact branch gate (fail-closed)", "AC-6")
	assertOrderedBefore(t, stage, "[ ] Step 1.9 — Stage artifact branch gate (fail-closed)", "[ ] Step 2   — Deliberation", "AC-6")

	// --- AC-7: new section positioned strictly after Step 1.8 and before Step 2 ---
	assertOrderedBefore(t, stage, "### Step 1.8: Learnings Retrieval", "### Step 1.9: Stage Artifact Branch Gate (NON-NEGOTIABLE)", "AC-7")
	assertOrderedBefore(t, stage, "### Step 1.9: Stage Artifact Branch Gate (NON-NEGOTIABLE)", "### Step 2: Deliberation", "AC-7")

	// --- AC-8: worktree topology precheck ---
	assertPresent(t, stage, "git worktree list --porcelain", "AC-8")
	assertPresent(t, stage, "Class (2) is permitted and passes the gate", "AC-8")
	assertPresent(t, stage, "and records a P-016/P-005 violation, before any branch operation", "AC-8")

	// --- AC-9: slug source fields + normalization + 48-byte bound + empty fallback ---
	assertPresent(t, stage, "Proposed covering feature title", "AC-9")
	assertPresent(t, stage, "truncate to at most 48 bytes", "AC-9")
	assertPresent(t, stage, "max_slug_length` of 60", "AC-9")
	assertPresent(t, stage, "use the literal slug\n   `stage-session`", "AC-9")

	// --- AC-10: base_ref resolution + create-or-select/resume + discriminator ---
	assertPresent(t, stage, "git symbolic-ref --short refs/remotes/origin/HEAD", "AC-10")
	assertPresent(t, stage, "The gate must not\n   assume the literal name `main`", "AC-10")
	assertPresent(t, stage, "require a clean worktree (`git status --porcelain=v1` produces no", "AC-10")
	assertPresent(t, stage, "Content ownership", "AC-10")
	assertPresent(t, stage, ".backlogit/`,\n      `docs/plans/`, `docs/decisions/`, `docs/memory/`", "AC-10")

	// --- AC-11: symbolic HEAD equality + default-branch inequality ---
	assertPresent(t, stage, "`git symbolic-ref --short HEAD` must\n   equal the expected", "AC-11")
	assertPresent(t, stage, "The expected branch must differ from `base_ref`", "AC-11")

	// --- AC-12: fail-closed halt tokens + P-005 + no tracked mutation + categorical deferral rule ---
	assertPresent(t, stage, "STAGE_BRANCH_GATE_FAIL", "AC-12")
	assertPresent(t, stage, "STAGE_BRANCH_GATE_ANCHOR_FAIL", "AC-12")
	assertPresent(t, stage, "records a\n   P-005 policy-violation event, and performs no tracked mutation", "AC-12")
	assertPresent(t, stage, "The rule is\n   categorical; the following are illustrative, not a closed list", "AC-12")
	assertPresent(t, stage, "the write half of the\n   Step 1 late-identifier reconciliation", "AC-12")
	assertPresent(t, stage, "the archival half of the Step 1 unconditional duplicate-detection scan", "AC-12")
}

// TestStageBranchGate_ContractCorrected is the sole generated function for
// this release unit's harness footprint (018-F has exactly one queued task).
// It invokes assertStageBranchGateContract unconditionally -- no t.Skip,
// no build tag, no env-var/platform gate, no activation manifest, no
// selector flag (AC-14).
func TestStageBranchGate_ContractCorrected(t *testing.T) {
	assertStageBranchGateContract(t)
}

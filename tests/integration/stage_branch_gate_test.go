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
	"strings"
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

	// --- AC-2: actor-named Stage MUST NOT bullet symmetric with Ship's; Stage MAY create/checkout grant.
	// U-6 fix: the withdrawn grant phrase must also be absent from _stage.agent.md itself (not only the
	// policy file), and the "Commit or push directly to `main`" count is bounded to the region between
	// the two role-actor headings so a duplicate elsewhere in the same file cannot mask a missing bullet
	// for either actor specifically. ---
	assertPresentAtLeast(t, policy, "- Commit or push directly to `main`", 2, "AC-2")
	assertPresent(t, policy, "- Create and check out a dedicated Stage artifact branch (`chore/stage-{scope-slug}`)", "AC-2")
	assertAbsent(t, stage, "on the default branch or a dedicated chore/admin branch", "AC-2/U-6")
	stageMustNotIdx := indexOf(t, policy, "Stage MUST NOT", "AC-2/U-6")
	stageMayIdx := indexOf(t, policy, "Stage MAY", "AC-2/U-6")
	if stageMayIdx <= stageMustNotIdx {
		t.Fatalf("AC-2/U-6: expected 'Stage MAY' to appear after 'Stage MUST NOT' in policy")
	}
	stageBlock := policy[stageMustNotIdx:stageMayIdx]
	assertPresent(t, stageBlock, "- Commit or push directly to `main`", "AC-2/U-6")

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

	// --- AC-5: Role Boundary Git row + PR row + Planning row (U-3 fix: Planning row no longer grants an
	// unqualified, branch-agnostic commit; it defers to the Git row) ---
	assertPresent(t, stage, "Create and check out a dedicated Stage artifact branch (`chore/stage-{scope-slug}`); commit backlog/planning artifacts on that branch only", "AC-5")
	assertPresent(t, stage, "create/use an explicit, time-boxed spike/research worktree only for staging investigation", "AC-5")
	assertPresent(t, stage, "Create or checkout feature/chore branches for code execution; create/use parallel implementation branches or worktrees", "AC-5")
	assertPresent(t, stage, "The operator — never Stage — pushes the Stage artifact branch and opens/approves the resulting merge-commit staging PR", "AC-5")
	assertPresent(t, stage, "no Orchestrator authority to do so is granted or assumed by this release unit", "AC-5")
	assertPresent(t, stage, "| PR | The operator", "AC-5")
	assertPresent(t, stage, "Create, push, or merge pull requests |", "AC-5")
	assertPresent(t, stage, "commit them to the dedicated Stage artifact branch only (see Git row)", "AC-5/U-3")
	assertAbsent(t, stage, "commit them to the repo", "AC-5/U-3")

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

	// --- AC-9: slug source fields + normalization order (all 5 steps, in order; U-4 fix) + 48-byte bound
	// + max_slug_length + empty fallback. U-5 fix: anchored to the Step-1.9-specific phrase so Step 1.5's
	// own unrelated, pre-existing use of "Proposed covering feature title" cannot tautologically satisfy
	// this assertion after a deletion of the Step 1.9 clause. ---
	assertPresent(t, stage, `the "Proposed covering feature title" string of the Step 1.5`, "AC-9/U-5")
	assertOrderedBefore(t, stage, "NFC-normalize the source string", "lowercase by ASCII case folding only", "AC-9/U-4")
	assertOrderedBefore(t, stage, "lowercase by ASCII case folding only", "replace every maximal run of characters outside", "AC-9/U-4")
	assertOrderedBefore(t, stage, "replace every maximal run of characters outside", "trim leading and trailing `-`", "AC-9/U-4")
	assertOrderedBefore(t, stage, "trim leading and trailing `-`", "truncate to at most 48 bytes", "AC-9/U-4")
	assertPresent(t, stage, "truncate to at most 48 bytes", "AC-9")
	assertPresent(t, stage, "max_slug_length` of 60", "AC-9")
	assertPresent(t, stage, "use the literal slug\n   `stage-session`", "AC-9")

	// --- AC-10: base_ref resolution (primary + fallback command + both-fail halt; U-4 fix) +
	// create-or-select/resume + discriminator (exact merge-base command, triple-dot diff syntax, and the
	// corrected/broadened content-ownership allow-list from U-1/U-2) ---
	assertPresent(t, stage, "git symbolic-ref --short refs/remotes/origin/HEAD", "AC-10")
	assertPresent(t, stage, "git rev-parse --abbrev-ref origin/HEAD", "AC-10/U-4")
	assertPresent(t, stage, "If both fail, halt with `STAGE_BRANCH_GATE_FAIL`", "AC-10/U-4")
	assertPresent(t, stage, "The gate must not\n   assume the literal name `main`", "AC-10")
	assertPresent(t, stage, "require a clean worktree (`git status --porcelain=v1` produces no", "AC-10")
	assertPresent(t, stage, "Content ownership", "AC-10")
	assertPresent(t, stage, "git merge-base refs/heads/{expected_branch} refs/remotes/origin/{base_ref}", "AC-10/U-4")
	assertPresent(t, stage, "refs/remotes/origin/{base_ref}...refs/heads/{expected_branch}", "AC-10/U-4")
	assertPresent(t, stage, "`.backlogit/queue/`,\n      `.backlogit/archive/`, `.backlogit/checkpoints/`, `.backlogit/reconcile/`,\n      `.backlogit/stash.jsonl`, `docs/plans/`, `docs/decisions/`, `docs/memory/`,\n      `docs/compound/`, `docs/archive/`", "AC-10/U-1")
	assertPresent(t, stage, "`.backlogit/config.yaml`, `.backlogit/header-def.yaml`,\n      `.backlogit/hooks.yaml`, `.backlogit/migration.yaml`, `.backlogit/registry.yaml`, or", "AC-10/U-2")
	assertPresent(t, stage, "Scope identity: `git log -1 --format=%B refs/heads/{expected_branch}` must contain", "AC-10/M-1")
	assertPresent(t, stage, "48-byte truncation of two distinct\n      covering-feature titles", "AC-10/M-1")
	assertPresent(t, stage, "fall back to the identical literal `stage-session`", "AC-10/M-1")
	assertPresent(t, stage, "Branch\n   resume is permitted only when all three conditions match", "AC-10/M-1")
	assertPresent(t, stage, "Operator escape:", "AC-10/M-1")

	// --- AC-11: symbolic HEAD equality + default-branch inequality + detached-HEAD failure clause (U-4 fix) ---
	assertPresent(t, stage, "`git symbolic-ref --short HEAD` must\n   equal the expected", "AC-11")
	assertPresent(t, stage, "A detached HEAD (non-zero exit\n   from that command) fails the gate", "AC-11/U-4")
	assertPresent(t, stage, "The expected branch must differ from `base_ref`", "AC-11")

	// --- AC-12: fail-closed halt tokens + P-005 + no tracked mutation + categorical deferral rule ---
	assertPresent(t, stage, "STAGE_BRANCH_GATE_FAIL", "AC-12")
	assertPresent(t, stage, "STAGE_BRANCH_GATE_ANCHOR_FAIL", "AC-12")
	assertPresent(t, stage, "records a\n   P-005 policy-violation event, and performs no tracked mutation", "AC-12")
	assertPresent(t, stage, "The rule is\n   categorical; the following are illustrative, not a closed list", "AC-12")
	assertPresent(t, stage, "the write half of the\n   Step 1 late-identifier reconciliation", "AC-12")
	assertPresent(t, stage, "the archival half of the Step 1 unconditional duplicate-detection scan", "AC-12")
}

// indexOf locates needle in content and fails the test if it is absent, returning the byte
// offset of the first occurrence. Used to bound a sub-region of content for a scoped assertion
// (U-6 fix: prevents a global, unscoped count from being satisfied by an unrelated duplicate
// elsewhere in the same file).
func indexOf(t *testing.T, content, needle, row string) int {
	t.Helper()
	idx := strings.Index(content, needle)
	if idx < 0 {
		t.Fatalf("row %s: ordering anchor not found: %q", row, needle)
	}
	return idx
}

// TestStageBranchGate_ContractCorrected is the sole generated function for
// this release unit's harness footprint (018-F has exactly one queued task).
// It invokes assertStageBranchGateContract unconditionally -- no t.Skip,
// no build tag, no env-var/platform gate, no activation manifest, no
// selector flag (AC-14).
func TestStageBranchGate_ContractCorrected(t *testing.T) {
	assertStageBranchGateContract(t)
}

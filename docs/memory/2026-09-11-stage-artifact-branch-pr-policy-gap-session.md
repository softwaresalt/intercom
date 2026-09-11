# Stage session memory — Stage artifact branch/PR policy gap correction

**Date**: 2026-09-11
**Agent**: Stage
**Branch**: `chore/stage-pipeline-policy-gap`
**Phase**: complete — shipment queued, handed off to Ship
**Session outcome**: exactly one reviewed, queued shipment — `017-S`

---

## 1. Entry point

Direct operator request (not an existing stash entry): correct the pipeline policy gap that
allowed Stage artifacts to be pushed directly to `main`, evidenced by commit `fdff9e4`. Captured
through the configured backlog intake as stash entry **`638A410B`** (kind `feature` — `--kind
chore` is rejected by this backlogit build), then run through the normal
deliberate → plan → harden → review → harvest pipeline. No ad hoc tracker was created.

## 2. Tool state

- **`TOOL_DEGRADED`** — no backlogit MCP tools were exposed in this session. Fell back to the
  registry-declared `backlogit` CLI (v1.10.1). All operations below used the CLI fallback.
- `INDEX_SYNC_OK` at session start (199 artifacts) and at session end (221 artifacts).
- Checkpoint recovery: zero `stage`-owned active checkpoints, no validation/quarantine anomalies
  → normal startup, not a failure.

## 3. Artifacts produced

| Path | Role |
|---|---|
| `docs/decisions/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-deliberation.md` | Source document / P-003 lineage root (§8 rev-2 addendum, §9 rev-3 addendum) |
| `docs/plans/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-plan.md` | Implementation plan, revision 3, `status: reviewed` (§10a hardening, §10 Constitution Check, §12 review record) |
| `docs/memory/2026-09-11-stage-artifact-branch-pr-policy-gap-session.md` | This file |

## 4. Decision summary

Chose **Option 2**: prose correction of every authorization surface **plus** a deterministic CI
regression gate. Rejected: prose-only (D4 — unreliable), a new policy ID (scope), and a
generalized branch framework (operator simplicity constraint).

**Root cause — four authorization surfaces** (the count was revised twice; see §5):

1. `.github/agents/_stage.agent.md` Role Boundary **Git row** — the *operative* surface
2. `.github/policies/workflow-policies.md` **P-010 Stage MAY** bullet
3. `.github/agents/_orchestrator.agent.md` Step 1.5 **sub-step 3e**
4. `.github/agents/_orchestrator.agent.md` Step 1.5 **preamble**

**D2 (template coupling)** — the `.tmpl` sources under
`.copilot/installed-plugins/autoharness/autoharness/templates/` are **gitignored and untracked**
(verified with `git check-ignore -v` / `git ls-files --error-unmatch`). Editing them is
uncommittable, unreviewable, and erased on reinstall, and no CI check can enforce
template↔artifact equality. Operator scope item 1 was therefore narrowed: the gate asserts the
**invariant over the installed artifacts**, so a regeneration that restores the language turns CI
red. The Scope Boundary Auditor independently reviewed and endorsed this narrowing.

## 5. What the review gate actually caught

This is the most reusable part of the session. Three of my own framing claims were **falsified by
review and independently re-verified by me** before being withdrawn:

- "Exactly one push-shaped hit in the corpus" — **false**. The search was push-shaped; the
  *operative* surface is **commit-shaped** (`_stage.agent.md` Git row). Round 1, P0.
- The rev-1 P-010 wording was **self-deadlocking**: it required Stage to merge a PR while Stage's
  own Role Boundary forbids creating, pushing, or merging PRs. Fixed by naming the **Orchestrator**
  as the PR actor. Round 1, P0.
- "Deleting 3e loses no behaviour" — **false**. 3e carried the only branch-point instruction
  ("from the current commit") for the already-committed-but-unpushed path. Folded into 3a. Round 1, P0.

Then round 2 found a **fourth** surface (the Step 1.5 preamble) that survived *both* the
push-shaped and commit-shaped sweeps because it reads as a verification instruction while stating
the gate's postcondition as artifacts being committed *to* the default branch.

**Lesson (recorded in deliberation §9.2)**: manual enumeration of contract surfaces is not
reliable at this scale — which is itself the argument for D4. The plan now requires the detector,
not a reader, to establish the surface count (AC-A3.4 records the first full-corpus scan before
any correction begins; risk R9 tracks a possible fifth).

## 6. Highest-risk design element

The detector must reject imperative direct-push/commit-to-default *instructions* while accepting
*prohibitions about the same topic*. Prior art
`docs/compound/2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md` records a CI
grep that permanently self-failed by matching its own describing text — the exact failure mode here.

Rev-3 design decisions that came out of review:

- **Per-construct verdicts, never block-level** — a normalized Step 1.5 block contains both an
  accept (3b `create a PR to \`main\``) and a reject (3e); block-level accept would mask it.
- **Bold lead-in, not ATX heading**, for prohibition context — the real files use
  `**Ship MUST NOT**:` lead-ins, so an ATX-only reading would flag correct text and make
  zero-findings unreachable.
- **Cell-scoped, not row-scoped** table matching — a Role Boundary row's Allowed cell holds push
  verbs and its Forbidden cell holds `main`; row-scoping would pair them.
- **Guarded default-branch resolution** — `git symbolic-ref refs/remotes/origin/HEAD` is not
  created by `actions/checkout`, so under `set -euo pipefail` an unguarded call exits non-zero in
  CI and is indistinguishable from a real violation.
- `pull`/`checkout` added to the read-verb exclusions; `rather than`/`instead of` **removed** from
  the prohibition-marker set (directional, not prohibitive → fail-open).

**Ordering keeps `main` green while staying test-first**: the gate enters CI only *after* the
violations are removed (sub-epic C follows B), so **no advisory toggle is needed** — strictly
safer and simpler than the bounded-advisory-window pattern prior art required.

## 7. Backlog structure created

- Feature **`018-F`** — top-level release unit (chore-classified via label; backlogit exposes no
  `chore` type)
- Tasks **`018.004-T` … `018.011-T`** — 8 atomic units
- Shipment **`017-S`** — queued, 9 explicit items

**Structural finding**: this workspace's backlogit schema stores `size` on the `task` type only
and defines **no `complexity` field on any type** (`backlogit metadata types` → task fields =
`priority`, `size`, `status`), despite `backlog-registry.yaml` declaring `features.sizing: true`.
Consequences, both handled:

1. A first-pass feature → sub-epic(`task`) → task(`subtask`) hierarchy could not carry sizing at
   the atomic level. The 11 provisional items (`018.001-T`–`018.003-T`,
   `018.001.001-ST`–`018.003.002-ST`) were **archived, not deleted** (no destructive operation was
   authorized) and the 8 atomic units were recreated as `task` type. Sub-epic grouping is now
   carried on `sub-epic-a/b/c` labels plus the `blocks` dependency graph.
2. `complexity` is recorded as **enum-validated prose** in each task's `implementation-notes`,
   per the Stage two-axis sizing contract's degradation path, with the degradation flagged
   in-line on every task.

**Known benign artifact**: shipment `017-S`'s derived `size_composition` roll-up reports
`unsized: 3` and 11 members, because it walks all children of `018-F` including the three archived
provisional sub-epic containers. The authoritative membership is the explicit `items` array
(9 IDs). This is documented in `018-F`'s goals section so Ship is not misled.

## 8. Next actor

**Ship** — claim shipment `017-S` and execute `018.004-T` first (the failing regression harness).
Do not start at `018.007-T`; the `blocks` edges enforce harness-before-correction.

The Orchestrator's Step 1.5 manifest verification must use the **staging-branch** form, not the
`origin/main` form — this work exists precisely to forbid the latter for Stage artifacts:

```sh
git show chore/stage-pipeline-policy-gap:.backlogit/queue/017-S.md
```

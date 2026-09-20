---
title: "029.001-T — CASCADE contract-surface matrix (required before any edit)"
date: 2026-09-20
shipment: 026-S
feature: 029-F
task: 029.001-T
status: complete
---

## Purpose

Per `docs/compound/workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md`
(026-F unblock step O-3), this matrix enumerates all six live surfaces of the CASCADE
contract this unit touches — **Re-point the CASCADE branch at `shipment-reconcile
mode:safe-close`** (`_ship.agent.md:830` and `:861-869`) — before any edit lands. Any empty
cell below is an open edge that blocks the edit. A producer-scoped or single-file measurement
is explicitly not accepted as contract closure.

Bound contract snapshot: commit `14d44e3c1f28b321e8db24ab8cafcba346996133`
(verified via `git diff 14d44e3c1f28b321e8db24ab8cafcba346996133 HEAD --
.github/agents/_ship.agent.md` → empty, i.e. no drift since the plan froze PIN-830/PIN-861).

## Surface Matrix

| # | Surface | Current evidence | Intended change | Executable verification | Depends on |
|---|---|---|---|---|---|
| 1 | **Producer / API** — `classify-close-path` (`.github/skills/shipment-reconcile/SKILL.md`, "Classify-Close-Path Mode" section) | Already emits the three-token value set verbatim: `` CLOSE_PATH_VERDICT: CASCADE ``, `` CLOSE_PATH_VERDICT: SAFE_CLOSE ``, `` CLOSE_PATH_VERDICT: BLOCK ``, each with `VERDICT_REASON` / `VERDICT_EVIDENCE` / `CLASSIFICATION_BINDING`. Read-only, no lock, no mutation ("performs no archive, no status transition and no record mutation"). | **No change required.** This unit edits only the consumer (`_ship.agent.md`); the producer's value set and binding digest are unaffected and were pinned by a prior unit (029-F's predecessor `028-F`). | `go test ./tests/integration/... -run TestShipFeatureCompletionContract33Rows/row_21` and `/row_22` (rows 21a-22c cover producer section heading, no-mutation clause, three verdict tokens) — **PASS pre-repair (verified above)**; must remain PASS post-repair since this surface is untouched. | none |
| 2 | **Consumer** — `_ship.agent.md` CASCADE clauses (`:830`, `:832-836`, `:861-869`) | `:830` currently reads *"Do NOT call `backlogit shipment ship` / `backlogit_ship_shipment` unless `mode: classify-close-path` has already returned `CLOSE_PATH_VERDICT: CASCADE`…"* — a **conditional authorization** of the direct call. `:861-869` currently reads *"invoke the cascade `backlogit shipment ship` / `backlogit_ship_shipment` operation as this shipment's closure instead of the safe-close sequence…"* — a **direct-cascade branch**. Both measured verbatim against the live file at the bound snapshot (line numbers confirmed via `Get-Content .github\agents\_ship.agent.md` indexed access: line 830, lines 861-869). | Apply PIN-830 (plan §3.3) verbatim to `:830`, replacing the conditional authorization with a designation of `shipment-reconcile mode: safe-close` (its Cascade Close Sub-Procedure) as the CASCADE close path, carrying `CLASSIFICATION_BINDING` into that call, and prohibiting any direct `backlogit shipment ship` / `backlogit_ship_shipment` call for any verdict. Apply PIN-861 verbatim to `:861-869`, replacing the direct-cascade branch with an invocation of `shipment-reconcile mode: safe-close` carrying the binding, preserving the requeue/detach semantics. | New contract test (029.005-T, `tests/integration/`) asserting PRESENT ``designates `shipment-reconcile` `mode: safe-close` `` and ABSENT ``invoke the cascade`` against `_ship.agent.md` content — **red pre-repair, green post-repair** (AC-2.6). Existing suite rows 25/26/31 (`ship_feature_completion_contract_test.go`) must stay green before and after (AC-2.9) since they read the frozen `:832-836` span this edit does not touch. | Surface 5 (test authored against this edit); frozen spans `:832-836` remain content-identical (verified separately, no dependency on this edit). |
| 3 | **Authority grant** — P-010 (`.github/policies/workflow-policies.md:211-250`) | Ship's authorized action is stated generically: *"Claim shipments, move tasks to active/done, … close shipments, archive completed items."* No specific close-path mechanism (direct cascade vs. safe-close delegation) is named at this surface — the grant is mechanism-agnostic. | **No change required.** The generic grant already covers closing a shipment via whichever close path `_ship.agent.md`'s consumer clauses designate; changing the mechanism at the consumer surface does not require re-wording the authority grant. | `go test ./tests/integration/... -run TestShipFeatureCompletionContract33Rows/row_20` and `/row_28` (policy-surface rows for the covering-feature completion grant) — **PASS pre-repair**; unaffected by this edit since it targets a different clause (covering-feature completion, not CASCADE routing) and remains PASS post-repair. | none |
| 4 | **Refusal / fallback sink** — BLOCK and safe-close-as-default (`.github/skills/shipment-reconcile/SKILL.md`, "Classify-Close-Path Mode" fail-closed clause and "Safe-Close Mode" step 0) | *"every unsupported, mixed, ambiguous, torn, missing or incompletely enumerated case returns `CLOSE_PATH_VERDICT: BLOCK`"*; *"safe-close is the default"* (Safe-Close Mode step 0 preamble). Both already correct and unrelated to which mechanism the CASCADE consumer branch invokes. | **No change required.** The refusal/default sink is independent of how the CASCADE consumer branch is wired; PIN-830/PIN-861 do not alter BLOCK handling or the safe-close-is-default posture. | `go test ./tests/integration/... -run TestShipFeatureCompletionContract33Rows/row_22c` and `/row_23` (fail-closed BLOCK clause, bound-refusal-and-branch-selects, default-gone-bound-only-present, unbound/invalid refusal tokens) — **PASS pre-repair**; remain PASS post-repair (untouched surface). | none |
| 5 | **Tests / pinned needles** — `tests/integration/` | Existing `TestShipFeatureCompletionContract33Rows` rows 25, 26, 31 already assert against the **frozen** spans (`:832-836` `HALT on any result token…`, `carrying the returned CLASSIFICATION_BINDING…`) which this edit preserves byte-identical, and against the **absence** of stale literals (e.g. ``skill with `mode: safe-close`, `shipment_id`, and the`` at row 31) which this edit must not reintroduce. No test yet asserts the PRESENT/ABSENT red→green pair for PIN-830's new designation wording. | Author one new test (029.005-T) with the single red→green assertion pair pinned in plan §3.3: PRESENT (post-repair only) ``designates `shipment-reconcile` `mode: safe-close` `` ; ABSENT (post-repair) ``invoke the cascade``. Needles cut verbatim from PIN-830/PIN-861 per the compound rule — never authored from prose. | `go test ./tests/integration/...` run twice: once against the pre-repair tree (new test FAILS, 33-row suite PASSES) and once against the post-repair tree (new test PASSES, 33-row suite still PASSES) — both runs recorded in the 029.005-T task completion note. | Surface 2 (the consumer edit must land before the new test can go green). |
| 6 | **Tooling probes** — `backlogit shipment ship` CLI contract | `backlogit shipment ship --help` (probed live, this session): accepts only `<id>` plus `--sha` / `--message` / `--author`. No `--classification-binding` or equivalent flag exists on this CLI surface. | **No change required.** PIN-830/PIN-861 do not alter this CLI's contract; they change which *instruction-file clause* is permitted to invoke it (only from inside `shipment-reconcile mode: safe-close`'s Cascade Close Sub-Procedure, never directly from `_ship.agent.md`). The CLI itself is unaffected. | Verbatim command + output already captured this session: `backlogit shipment ship --help` → `Usage: backlogit shipment ship <id> [flags]` with `--sha`/`--message`/`--author`/`--help` only. Re-run post-repair to confirm no drift (expected: identical output, since this edit never touches the CLI). | none |

## Missing-edge check (mechanical)

* Every producer row (1) has a consumer row (2) that acts on its value set. ✓
* Every consumer row (2) that acts on an outcome has an authority row (3) permitting the action. ✓ (P-010's generic "close shipments" grant covers it; verified mechanism-agnostic, no re-wording needed.)
* Every outcome value in the producer's value set (`CASCADE`, `SAFE_CLOSE`, `BLOCK`) has a refusal row (4) or a consumer row (2). ✓ (`BLOCK` → refusal row 4; `CASCADE`/`SAFE_CLOSE` → consumer row 2, both now routed through the same `safe-close` delegation per PIN-830/861.)
* Every surface row has ≥1 test row (5) whose needle resolves inside that surface's own block, never borrowed from a neighbouring site. ✓ (Row 25/26/31 needles are already scoped to `_ship.agent.md`'s own consumer block; the new 029.005-T needle is scoped to the same file's rewritten `:830`/`:861-869` span.)
* Every claim about installed tooling has a probe row (6) with command + verbatim output. ✓

**No open edges.** The matrix is closed; 029.003-T / 029.004-T / 029.005-T may proceed.

## Sources

* `docs/plans/2026-09-18-intercom-go-ship-pipeline-contract-repair-plan.md` §3.2, §3.3
* `docs/compound/workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md`
* `.github/agents/_ship.agent.md` (measured lines 439-441, 830-836, 861-869, 265, 322, 338-358, 362-367)
* `.github/skills/shipment-reconcile/SKILL.md` (Classify-Close-Path Mode, Safe-Close Mode step 0)
* `.github/policies/workflow-policies.md` (P-010)
* `tests/integration/ship_feature_completion_contract_test.go` (`TestShipFeatureCompletionContract33Rows`)
* Live probe: `backlogit shipment ship --help`

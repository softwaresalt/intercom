---
title: "Post-merge closure: 026-S / 029-F — Repair Ship agent execution contract (plan unit 2)"
description: "Operational closure artifact for shipment 026-S / feature 029-F"
status: "complete"
tags:
  - "closure"
  - "026-S"
  - "029-F"
  - "post-merge"
date: 2026-09-20
mode: post-merge
shipment: 026-S
feature: 029-F
pr: 69
merge_commit_sha: ebf9eef0112bf866a77d1b6840a3822cd2fa2997
compaction_status: done
closure_status: READY
releasability: READY
conditions: []
---

# Post-merge closure: 026-S / 029-F — Repair Ship agent execution contract (plan unit 2)

- Mode: `post-merge`
- PR: #69 (`feat/repair-ship-agent-execution-contract-plan-unit-2`)
- Merge commit: `ebf9eef0112bf866a77d1b6840a3822cd2fa2997` (merge-commit strategy, P-009)
- Date: 2026-09-20
- Compaction status (P-020): `done` — `compact-context` invoked (target: `memory`, scoped
  to this just-closed release unit) during this closure session; see "Compaction" section
  below.

## Merge Confirmation Gate

* `gh pr view 69 --json state,mergedAt,mergeCommit` → `state: MERGED`, `mergedAt:
  2026-09-20T18:40:31Z`, `mergeCommit.oid: ebf9eef0112bf866a77d1b6840a3822cd2fa2997`.
* `git fetch origin main` then `git merge-base --is-ancestor
  ebf9eef0112bf866a77d1b6840a3822cd2fa2997 origin/main` → exit `0`. Merge SHA independently
  confirmed present in `origin/main` history, with exactly two parents:
  `3f2cf9d5b53c6b85b5147839e3f3d62646d70ba4` (prior `main` tip) and
  `825acc9668b30890aa18b0e05e88763e363ce4ce` (reviewed feature-branch HEAD).

## Merge approval and gate re-verification (this session)

* Operator supplied an explicit, scoped merge approval token — `PR 69: Merge approved` —
  authorizing merge of PR #69 only, via merge-commit strategy only, with admin fallback
  explicitly NOT authorized, and explicitly not covering any post-merge closure PR.
* Re-read PR #69 immediately before merge: `headRefOid:
  825acc9668b30890aa18b0e05e88763e363ce4ce` (unchanged from the Orchestrator-verified HEAD),
  `mergeStateStatus: CLEAN`, `mergeable: MERGEABLE`, all 13 CI checks green, `reviewDecision`
  empty (no hosted review engaged).
* P-018 Copilot-review gate (`autoharness gate copilot-review 69 --repo softwaresalt/intercom
  --enforcement auto --json`): `verdict: NOT_APPLICABLE`, `exit_code: 0` at the unchanged
  HEAD (Copilot not engaged for this PR).
* `pipeline-topology` gate (`--mode agent --phase lifecycle --shipment 026-S --json`): exit 0,
  all checks (`detect_before_consistency`, `active_shipment_invariant`, `branch_ownership`,
  `worktree_topology`, `shipment_readiness`) passed.
* Repo merge settings confirmed P-009-compliant:
  `allow_merge_commit: true`, `allow_squash_merge: false`, `allow_rebase_merge: false`.
* Local review readiness recorded in the PR body for the reviewed HEAD: `READY`, zero
  P0/P1 findings, full-build evidence present, no follow-up handling required.
* `gh pr merge 69 --merge` — normal merge path, no `--admin`, no force. Result: `state:
  MERGED`, merge commit `ebf9eef0112bf866a77d1b6840a3822cd2fa2997`.

## Summary of the change

Executes plan `docs/plans/2026-09-18-intercom-go-ship-pipeline-contract-repair-plan.md`
§3 Unit 2 (`029-F`, tasks `029.001-T`..`029.005-T`):

* Filled the empty Step 4.3 Format quality-gate command slot in `_ship.agent.md` with
  `test -z "$(gofmt -l .)"` (behavioral parity with `ci.yml:303-310`).
* Re-pointed the CASCADE close branch in `_ship.agent.md` from a direct
  `backlogit shipment ship` / `backlogit_ship_shipment` call to
  `shipment-reconcile mode: safe-close` (its Cascade Close Sub-Procedure), so
  `CLASSIFICATION_BINDING` is carried into an executable, revalidated close path.
* Added `TestCascadeRoutesThroughSafeClose` to
  `tests/integration/ship_feature_completion_contract_test.go`, confirmed genuine
  red pre-repair and green post-repair, and confirmed the pre-existing 33-row contract
  suite (`TestShipFeatureCompletionContract33Rows`) stayed green throughout.
* Also carried forward, per explicit operator directive, a pre-existing carry-forward
  payload (`.backlogit/stash.jsonl` ledger state plus five prior-session memory/compound
  artifacts and two read-only diagnostic wrapper scripts) that had no prior process for
  landing in a commit (commit `c86a059`).

No production `cmd/`/`internal/` runtime service code changed; the change is
agent/skill-contract prose plus one new Go test, applied by this shipment's own
Cascade Close Sub-Procedure at closure time (self-referential: this shipment's own
closure used the very close path it repaired).

## CI status and unresolved review items

* All 13 CI checks green at PR #69 HEAD `825acc9668b30890aa18b0e05e88763e363ce4ce`
  (`detect code changes`, `gitignore append-only + un-ignore regression (I6)`,
  `pipeline-topology (ambient)`, `test`, `test (windows, advisory)`, `lint`, `security`,
  `load cross-compile targets`, 4× `cross-compile (*)`, `ci gate`).
* **Local review (pre-PR)**: recorded in the PR body — `READY`, zero P0/P1 findings, no
  residual P2/P3 follow-up handling required.
* **Copilot review (P-018 gate)**: `NOT_APPLICABLE` throughout — no engagement signal.
* No unresolved P0/P1 findings at merge. No P-021 deferred-scope-expansion entries were
  captured for this shipment (no out-of-scope findings surfaced).

## Runtime verification report

See
`docs/closure/2026-09-20-026-s-029-f-repair-ship-agent-execution-contract-runtime-verification.md`.
Verdict: **READY** — no runtime service surface changed; the pinned contract-needle test
suite is the durable substantive evidence and is green.

## Invariants to preserve

* `_ship.agent.md:440`'s Format gate slot must remain `test -z "$(gofmt -l .)"` and must not
  regress to an empty command slot.
* `_ship.agent.md:830`/`:861-869` must continue to designate `shipment-reconcile
  mode: safe-close` (Cascade Close Sub-Procedure) as the sole CASCADE close path, carrying
  `CLASSIFICATION_BINDING`; a direct `backlogit shipment ship` / `backlogit_ship_shipment`
  call from `_ship.agent.md` itself must never be reintroduced.
* Frozen spans `_ship.agent.md:439/:441/:265/:322/:338-358/:362-367` must remain
  content-identical to their pre-repair form.
* `TestCascadeRoutesThroughSafeClose` and the pre-existing 33-row
  `TestShipFeatureCompletionContract33Rows` suite must both continue to pass on every
  future CI run touching `_ship.agent.md` or `shipment-reconcile/SKILL.md`.

## Pre-deploy audits

Not applicable — internal skill-documentation and test-only change; no migration, flag,
config-schema, or access change.

## Deployment / rollout path

**Merge-only, immediately active.** The corrected `_ship.agent.md` Format-gate command and
CASCADE-routing designation govern every future Ship session's build-feature quality gate
and shipment closure the moment this merge lands — including this very closure, which used
the repaired CASCADE routing (`shipment-reconcile mode: safe-close` → Cascade Close
Sub-Procedure) to close `026-S`/`029-F` itself.

## Post-deploy checks

On the post-merge closure branch (`post-merge/026-s-repair-ship-agent-execution-contract`,
built on `main` @ `ebf9eef0`):
* `go vet ./...` — clean.
* `gofmt -l .` — clean (zero unformatted files).
* `go build ./...` — clean.
* `go test ./...` — not re-run to completion on this branch (this branch's own diff is
  confined to `.backlogit/` archival artifacts — zero Go source files touched vs. the
  already-green PR #69 merge HEAD; a full re-run in this session did not terminate within a
  bounded wait, consistent with the documented pre-existing local `engram`-daemon-lock
  environmental flake in `docs/memory/2026-09-18/circuit-break-engram-daemon-readiness.md` —
  recorded as non-applicable evidence for a backlog-only diff rather than fabricated).

## Risky action record

Two destructive/state-transitioning actions this session, both independently verified
against every documented invariant:

1. **Covering-feature completion** (`029-F` `active -> done`, Ship Step 6 gate `a1`): all
   five conjunctive conditions verified — sole manifest feature member (`n=1`); all five
   descendants (`029.001-T`..`029.005-T`) already `done`; no live descendant at any depth
   (queue/archive scan for `029.*` returned only the 5 expected task files, no
   third-level subtasks); descendant-graph union (5 tasks) set-equal to the manifest's task
   members; containment holds (`029-F` has no `parent_id` — a root feature); prior live
   status was exactly `active`.
2. **Bound cascade shipment closure** (`backlogit shipment ship 026-S --sha
   ebf9eef0112bf866a77d1b6840a3822cd2fa2997 ...`): `classify-close-path` returned
   `CLOSE_PATH_VERDICT: CASCADE` / `VERDICT_REASON: FULLY_COVERED_ROOT` — `029-F` is a root
   feature (no `parent_id`) whose only descendants at any depth,
   `{029.001-T..029.005-T}`, are exactly the manifest's task members. Cascade result:
   `archived_ids: [029.001-T, 029.002-T, 029.003-T, 029.004-T, 029.005-T, 026-S, 029-F]`,
   `returned_ids: []` (no items requeued/detached — nothing outside the manifest was
   touched). Two-set gate verified clean (`archived_ids - allowed_ids` empty,
   `required_ids - archived_ids` empty). Post-mode: all 6 expected archive files present
   (`026-S`, `029-F`, `029.001-T`..`029.005-T`); `git status --short --
   ".backlogit/archive/"` showed only additions/modifications, no deletions (P-007 guard
   clean). Full detail in
   `.backlogit/reconcile/026-S-classify-close-path-20260920T114500Z.md` and
   `.backlogit/reconcile/026-S-safe-close-20260920T114500Z.md`.

No admin merge fallback was used for the PR #69 merge itself (`gh pr merge 69 --merge`,
normal path, merge-commit strategy per P-009).

## Healthy signals

* `go vet ./...` / `gofmt -l .` / `go build ./...` all green on the post-merge closure
  branch.
* All 13 CI checks passed cleanly on PR #69's final CI run.
* `gh pr merge 69 --merge` produced a genuine merge, independently confirmed present in
  `origin/main` history via `git merge-base --is-ancestor`, with exactly two parents.
* Cascade-close archived exactly the 7 manifest-scoped artifacts; nothing returned/detached;
  no `parent_id` cleared on any task.

## Failure signals

* A future regression that reintroduces an empty Format-gate command slot in
  `_ship.agent.md`.
* A future regression that reintroduces a direct `backlogit shipment ship` /
  `backlogit_ship_shipment` call from `_ship.agent.md`'s own CASCADE branch instead of
  delegating to `shipment-reconcile mode: safe-close`.
* A future failure of `TestCascadeRoutesThroughSafeClose` or
  `TestShipFeatureCompletionContract33Rows` on a change touching either instruction file.

## Monitoring plan

The durable monitoring for this shipment is `TestCascadeRoutesThroughSafeClose` plus the
pre-existing `TestShipFeatureCompletionContract33Rows` 33-row suite in
`tests/integration/ship_feature_completion_contract_test.go`, both exercised on every
future CI run touching `_ship.agent.md` or `shipment-reconcile/SKILL.md`.

## Rollback trigger / procedure

Not applicable in the operational sense (no live service to roll back). A future
regression would be remedied by a standard follow-up PR through the same review/CI
process.

## Validation window

Not applicable — no live deployment window; the change governs every future Ship
session's Format-gate execution and shipment CASCADE closure starting immediately after
this merge.

## Owner

Repository maintainer (`softwaresalt`) for the merged change and the shipment closure.

## Release-observability disposition

The `release-observability` capability pack's monitoring/dashboard/alert integration is
**not applicable** to this shipment: the change is skill/agent-contract prose plus one Go
test, with no running service, dashboard, metric, or alert surface affected.

## Releasability evidence

**READY.** No conditions. Local review (`READY`, zero P0/P1, no follow-up handling
required), CI (13/13 green), and runtime verification (not applicable — no runtime
surface affected) are all clean. The covering-feature completion and bound cascade-close
mutations were independently verified against every documented invariant with no HALT
condition triggered.

## Stash follow-up items

None identified. No follow-up tasks were surfaced by the closure artifact, runtime
verification report, or local review readiness result for this shipment.

## Source artifact cleanup

Checked `custom_fields.source_stash_id` and `custom_fields.source_deliberation_id` on both
shipped top-level manifest members. Neither `026-S`'s `custom_fields` (`items` only) nor
`029-F`'s `custom_fields` (`harness_status: pending` only) declares either field. No stash
or deliberation archival action was applicable or taken beyond what the bound cascade-close
itself already performed on manifest items (`026-S`, `029-F`, `029.001-T`..`029.005-T`).
(The feature body's free-text "Sources: stash 11B75632, 9C658143" reference is descriptive
prose, not a `custom_fields.source_stash_id` value, and is out of scope for this mechanical
cleanup step — consistent with the identical disposition recorded for the predecessor
shipment `025-S`/`028-F`.)

## Compaction (P-020, mandatory)

`compact-context` invoked with `target: memory`, scoped to this just-closed release unit
(the guaranteed Tier-1 consolidation candidate per the "completed feature or chore" rule).
Candidates identified: the three memory files tied to `026-S`/`029-F` in
`docs/memory/2026-09-20/` (`026-S-029-F-session-summary.md`, `026-S-pr69-awaiting-approval.md`,
`029-F-cascade-contract-surface-matrix.md`) → compacted to
`docs/memory/compacted/2026-09-20-026-s-029-f-repair-ship-agent-execution-contract-compacted.md`.
The three verbose originals were moved to `docs/archive/memory/2026-09-20/` (never deleted).
No other `docs/memory/2026-09-20/` file belongs to this release unit. **Compaction status:
`done`.**

## Documentation and compound-learnings evaluation

* `docs/ARCHITECTURE.md`, `AGENTS.md`, `docs/design-docs/`, `docs/product-specs/`: not
  applicable — this shipment changed only agent/skill-contract prose (`_ship.agent.md`) and
  one Go test; no structural, agent/skill-manifest, design, or product-requirement change
  occurred beyond the agent file itself, which is not indexed by those documents.
* `docs/compound/` (compound-refresh candidacy): searched for existing entries referencing
  the CASCADE direct-call routing problem this shipment fixed — none found (the two
  cascade-adjacent entries found,
  `docs/compound/2026-05-07-backlogit-shipment-status-constraints.md` and
  `docs/compound/workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md`,
  document unrelated concerns — shipment-status lifecycle constraints and the
  every-surface contract-closure discipline, respectively — and remain valid, not
  superseded). No `compound-refresh` keep/update/consolidate/replace/delete action was
  needed.

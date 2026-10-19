---
shipment_id: 017-S
feature_id: 018-F
pr_implementation: 58
merge_commit_sha: 5d69c727fc532eb61139f3e52cd3b64ab96574ac
cascade_commit_sha: ce5126709db61cf3adb431245fc1817f081e4f5b
pre_cascade_sha: 24d3acf3eb8456db036da5dda8289184c1ee1ebb
closure_branch: post-merge/017-s-stage-artifact-branch-pr-policy-gap-correction
compaction: done
releasability: READY
date: 2026-09-17
---

# 017-S / 018-F Post-Merge Closure

## Summary of the change

`018.008-T` — "Gate Stage artifact mutation behind a dedicated branch" — corrected a
P-010 policy gap: the Stage agent's artifact commits (backlog/planning artifacts)
were previously unguarded by a dedicated branch-and-PR discipline analogous to
Ship's implementation branches. This closure delivers:

* `.github/policies/workflow-policies.md` — P-010 amended (section: Stage MAY /
  MUST NOT) to require Stage artifact commits on a dedicated
  `chore/stage-{scope-slug}` branch, never on `main` directly.
* `.github/agents/_stage.agent.md` — Step 1.9 branch-gate contract added,
  requiring Stage to create/checkout its dedicated artifact branch before any
  commit, with a documented `base_ref` fallback and detached-HEAD handling.
* `tests/integration/stage_branch_gate_test.go` — new harness/test
  (`TestStageBranchGate_ContractCorrected`) proving the corrected contract text is
  present and internally consistent.

Scope was exactly 3 files (per `018.008-T`'s recorded `FILE COUNT: 3 files total …
0 new production files`) — no application/runtime source under `internal/` or
`cmd/` was touched. This is a governance/policy and test-harness change to the
agent workflow tooling, not a product runtime change.

## CI status and review

* PR #58 (`feat/017-s-stage-artifact-branch-pr-policy-gap-correction` →
  `main`): all 13 CI checks green (`detect code changes`, `gitignore
  append-only + un-ignore regression (I6)`, `pipeline-topology (ambient)`,
  `test`, `test (windows, advisory)`, `lint`, `security`, `load cross-compile
  targets`, 4× `cross-compile`, `ci gate`).
* Adversarial review (3 reviewers, `mode: report-only`): initial verdict
  `BLOCKED` (4 confirmed MAJOR in-scope findings: `U-3`, `U-1`, `U-2`, `M-1` —
  all classified P-021-C1 same-contract-surface completions of `018.008-T`
  and fixed directly, along with batched MINOR fixes `U-4`/`U-5`/`U-6`).
  **Review-fix cycle count: 1 of 3.** Post-fix re-verification (`go vet`,
  full test suite) PASS; PR proceeded to creation at reviewed HEAD `55ded63b…`.
  Four findings were deferred as advisory-only, not scope expansions —
  `C-1` (pre-existing `diagnostic.ps1` CWD-relative paths, carried forward
  verbatim by explicit operator instruction), `U-7` (remote-tracking-ref
  branch existence check), `U-8` (forward reference from Step 1 to the
  Step 1.9 deferral rule), `U-9` (`STAGE_BRANCH_GATE_ANCHOR_FAIL` trigger
  condition undefined) — no P-021 stash capture required for these per the
  reviewing session's own disposition.
* P-018 Copilot-review gate: `NOT_APPLICABLE` (no Copilot engagement signal).
* P-009 merge-strategy: repository confirmed `allow_merge_commit=true`; merge
  executed with `gh pr merge 58 --merge` (explicit merge-commit method).
  Verified two-parent shape at `5d69c727f…`: parent 1 `b04dd64` (mainline),
  parent 2 `55ded63b` (reviewed PR head).

## Risky/destructive action record — `PA-017-CASCADE`

* **Action**: cascade close of shipment `017-S` via `backlogit shipment ship`
  (the P-015 verified fully-covered-root exception path).
* **ActionRisk**: destructive. **Approval**: recorded in escrow
  2026-09-13T11:35:43-07:00 (prerequisite plan §11.1); release condition
  satisfied by the governing closure plan (revision 15) present on
  `origin/main` with `OPERATOR_ACCEPTED_RESIDUALS` disposition (contract
  D-L7).
* **Re-measurement at invocation (this session, all 9 escrow conditions,
  never inherited)**:
  1. Escrow released — governing plan on `origin/main`, `OPERATOR_ACCEPTED_RESIDUALS` — PASS.
  2. `classify-close-path` returned `CASCADE`/`FULLY_COVERED_ROOT` before any close call — PASS.
  3. `safe-close` (Cascade Close Sub-Procedure) revalidated the binding before mutation — PASS (digest match, no drift).
  4. `021-S` has shipped — PASS (verified prior session, unchanged).
  5. Parent-completion capability (Ship `_ship.agent.md` Step 6.1 a0/a1) merged and live — PASS (delivered by this same PR #58 / 018.008-T's prerequisite capability, already installed).
  6. Manifest unchanged at 13 members — PASS (re-measured this session).
  7. Pre-cascade baseline + pre-invocation path inventory exist and verified — PASS (`{pre_cascade_sha}` = `24d3acf3`, `{pre_paths}` bound at S-15).
  8. Linked-deliberation set verified EMPTY for this manifest, this invocation — PASS (`018-F` carries no `source_deliberation_id` and no `DL` pattern match).
  9. No scope expansion — PASS (manifest unchanged; §6.1 D1/D2/P6 all held; cascade-scope assertion confined to the 13 manifest members + shipment record; `019.007-T` byte-identical to its S-15 baseline).
* **ActionResult**: **executed successfully**. `backlogit shipment ship 017-S
  --sha 5d69c727… --message … --author …` returned `shipment_status: shipped`,
  `archived_ids: [018.008-T, 017-S, 018-F]`, `returned_ids: []`. Two-set gate
  (`allowed_ids`/`required_ids`) satisfied exactly; `parent_id` preserved on
  all 13 members; `019.007-T` (outside manifest, `dependencies: [018.008-T]`)
  verified untouched at the tree level, matching its D3 carve-out (its own
  status/dependency fields may legitimately change later as a consequence of
  `018.008-T` reaching `done`, but no such change occurred in this cascade
  invocation itself).
* Full evidence: `.backlogit/reconcile/017-S-cascade-close-20260917-220250.md`,
  `.backlogit/reconcile/017-S-post-20260917-220600.md`,
  `docs/plans/evidence/2026-09-16-017-S-closure/run-log.md`.

## Invariants preserved

* Single-active-release-unit (P-001): `017-S` was the sole active shipment
  throughout; now `archived`/`shipped` — the slot is free.
* Manifest closure discipline (P-015): only the 13 manifest members + the
  shipment record itself were archived/mutated; no out-of-manifest artifact
  (including `019.007-T`, its dependent) was touched.
* No unclaim / no `abandoned`: `017-S` moved `active → shipped → archived`
  only, per the only valid exits from `active` (M-10).
* Merge-commit discipline (P-009): both the implementation merge (`5d69c727`)
  and this closure's own eventual PR merge use the two-parent merge-commit
  shape, never squash/rebase.

## Validator evidence / runtime verification

**Not applicable.** This change touches no product runtime surface: no file
under `internal/` or `cmd/`, no `go.mod`/`go.sum` change, no non-test `*.go`
file. The change is confined to agent-workflow governance
(`.github/policies/`, `.github/agents/`) and a Go integration test
(`tests/integration/stage_branch_gate_test.go`) that exercises policy-text
contract shape, not runtime behavior of the `intercom-go` SDK. `runtime-verification`
is therefore skipped per its own "required when runtime surfaces were changed"
scoping; there is no validator manifest entry that applies.

## Pre-deploy audits

None required — no migration, flag, config, or access change. Deployment path
is merge-only (this is a source-controlled agent-tooling/governance change with
no separate deploy step).

## Post-deploy checks

* Confirm the next Stage session invocation creates/checks-out a
  `chore/stage-{scope-slug}` branch before any artifact commit (smoke-check on
  next Stage run).
* Confirm `go test ./tests/integration/ -run TestStageBranchGate_ContractCorrected`
  continues to pass on `main` post-merge (already re-verified locally before
  merge; CI `test` job covers it going forward).

## Healthy signals

* Next Stage session commits artifacts only on a dedicated `chore/stage-*`
  branch, never on `main`.
* No regression in `TestStageBranchGate_ContractCorrected` on `main`.

## Failure signals

* A future Stage session commits directly to `main` (policy-gap regression).
* `TestStageBranchGate_ContractCorrected` starts failing on `main`.

## Monitoring plan

No dashboards/alerts applicable (agent-workflow governance, not a running
service). Monitoring is the `test` CI job on `main` plus operator observation
of the next Stage session's branch behavior.

## Rollback trigger

If a future Stage session is observed committing artifacts directly to `main`,
or `TestStageBranchGate_ContractCorrected` regresses on `main`.

## Rollback procedure

Revert the implementation merge commit `5d69c727f…` via a new operator-approved
PR (plain revert — do not rewrite `main` history), then re-open `018.008-T`
(or a successor task) for Stage to re-plan the fix. This closure's own cascade
work (`ce5126709…`) would separately require the quarantine-then-revert
procedure in §8.2 of the governing closure plan if it, specifically, needed
unwinding — no such need arose in this run since every S-20 postcheck passed.

## Validation window

7 days of ordinary Stage session usage (next several planning sessions) to
confirm the branch-gate contract holds in practice.

## Owner

Ship/Stage agent operator (repository maintainer) — the same operator who
authorized `AUTHORIZE MERGE PR 58` and the governing closure plan's
`OPERATOR_ACCEPTED_RESIDUALS` disposition.

## Releasability evidence

**READY.** Merge-only release path; no runtime surface changed; no pre-deploy
audit required; monitoring is CI + next-session observation; rollback path
is a straightforward operator-approved revert PR. No conditions outstanding.

## Compaction status (P-020)

`done`. `compact-context` invoked with `target: all` immediately after this closure
artifact's creation. Candidate selection (threshold-gated): the just-shipped release
unit's 11 memory checkpoints (`017-S`/`018-F`, spanning Stage's plan-review lineage
2026-09-16 through Ship's merge/closure 2026-09-17) qualified under the
completed-work rule. Consolidated into
`docs/memory/compacted/2026-09-17-017-S-018-F-compacted.md` (decisions, rationale,
failed approaches, measured facts, deviations, file scope). All 11 verbose originals
moved to `docs/archive/memory/` (git renames, traceable). `markdownlint` clean on the
compacted summary (exit 0). No active-status checkpoint or unrelated release unit's
memory was touched; the six other active Stage-owned checkpoints were left untouched
per role boundary.

## Follow-up items stashed

The four deferred review findings (`C-1`, `U-7`, `U-8`, `U-9` — see CI status
and review section above) are advisory-only observations on pre-existing or
already-corrected text, not scope expansions, per the reviewing session's own
P-021 disposition recorded in the run log — no new stash capture is required
for them.

Two genuine out-of-plan file-disposition expansions were already captured as
stash entries earlier in this session (prior to merge): `29DC2014` (relocate
`diagnostic.ps1` → `scripts/diagnostic.ps1`) and `775E4A35` (contain the local
OOM crash report in an ignored location — see `.autoharness/staging/`, left
untouched per explicit operator instruction). No additional follow-up stash
entries were identified during this post-merge closure pass.

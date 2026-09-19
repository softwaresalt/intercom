# Compacted: 024-S / 027-F — Staging artifact handoff actor and push authority (B2)

**Release unit**: shipment `024-S` (feature `027-F`, task `027.001-T`)
**Status**: shipped and archived (`024-S` → `shipped`/archived, `027-F` → `done`/archived,
`027.001-T` → `done`/archived)
**PR**: [#64](https://github.com/softwaresalt/intercom/pull/64), merged via merge-commit
`748852a1d29921a444325145162945e718278e4f`
**Compacted from**: 4 verbose checkpoints in `docs/memory/2026-09-18/` (moved to
`docs/archive/memory/2026-09-18/`)

## What shipped

Rewrote item 3 of the Staging Artifact Merge Gate in
`.github/agents/_orchestrator.agent.md` (Step 1.5) to bind every sub-step to a named actor
(Stage, Orchestrator, OPERATOR) exercising only authority it already holds under P-010, and
removed the direct-push-to-`main` fallback arm entirely — replaced by
`STAGING_GATE_NO_ROUTE`. New tokenized halts: `STAGING_GATE_UNCOMMITTED`,
`STAGING_GATE_LOCAL_MAIN_AHEAD`, `STAGING_GATE_AWAITING_OPERATOR`,
`STAGING_GATE_OPERATOR_TIMEOUT`, `STAGING_GATE_NO_ROUTE`. Items 1, 2, 4 preserved unchanged;
`.github/policies/workflow-policies.md` byte-identical. Added characterization harness
`tests/integration/staging_gate_actor_contract_test.go` (one exported test fn, one helper, no
subtests/skips/build tags).

## Key decisions and rationale

* **Stage**: withdrew a claimed dependency-cycle finding (F-6) after the Scope Boundary Auditor
  refuted it and independent `backlogit dep list` verification confirmed 019.004-T's
  reverse-edge set was empty. Re-grounded the 019.004-T/027-F split on
  sequencing/independence instead. Lesson: verify a claimed cycle against recorded edges before
  building a decision on it.
* **Topology gate**: sole blocker at every phase (`pre_claim`, `post_claim`, `lifecycle`) was
  `PREDECESSOR_NOT_SHIPPED` naming numeric predecessor `023-S`. Per explicit operator directive
  (numeric shipment ordering deprecated for gating; `023-S` intentionally stays
  `archived_status: queued`), the audited `--force` override was applied narrowly to this one
  blocker at each phase — never any other gate class. Audit trail:
  `.autoharness/gates/pipeline-topology-force-audit.log`. `023-S` was never mutated.
  A fifth force was applied at the post-merge `lifecycle` phase for the same reason during
  closure.
* **Flaky test**: `TestInvokeStartScriptMainPropagatesNonZeroExitCode` intermittently exceeded
  its 30s timeout under environment-level engram/graphtor indexing load; confirmed
  environment-timing flakiness unrelated to the diff via isolated reruns on clean HEAD and
  HEAD+diff, both passing.
* **CRLF gofmt finding**: `tests/integration/staging_gate_actor_contract_test.go` shows CRLF in
  the local working-tree checkout only; the committed git blob is pure LF per
  `.gitattributes`. Classified as pre-existing, out-of-diff, left untouched (P-021 C1).
* Three unrelated pre-existing untracked files were preserved byte-for-byte throughout every
  branch/session boundary: `docs/memory/2026-09-17/017-S-018-F-post-merge-closure-pr59-awaiting-approval.md`,
  `run_all_commands.sh`, `run_commands.ps1`.

## Merge and closure

* Local review: `READY_WITH_FOLLOWUPS`, `P0=0/P1=0/P2=4/P3=9`, all residual findings classified
  non-blocking with explicit follow-up handling (see PR #64 body). One P2 (harness coverage gap
  on AC-1/AC-5/AC-8/AC-9 needles) captured as stash follow-up `3289EB69` for a future fast-follow
  alongside `019.004-T`.
* P-018 Copilot-review gate: `NOT_APPLICABLE` throughout (never engaged on this PR).
* CI: all 13 checks green at merge HEAD.
* Merge: `gh pr merge 64 --merge` (merge-commit strategy, P-009 compliant; repo has
  squash/rebase disabled).
* Post-merge covering-feature completion gate (Step 6 a1): all five conditions verified;
  `027-F` `active -> done`.
* Close path: `classify-close-path` → `CASCADE` / `FULLY_COVERED_ROOT` (027-F is a root feature
  whose sole descendant `027.001-T` is a manifest member); cascade `backlogit shipment ship
  024-S` invoked with the classification binding. All cascade verification gates passed
  (`returned_ids` empty, `archived_ids == allowed_ids == required_ids`, `parent_id` preserved).
* **Process correction during this closure**: the first attempt at Step 6 item 1 (covering
  feature transition + reconciliation + cascade close) was committed directly to `main` in
  error. Remediated before push: created `post-merge/024-s-staging-artifact-handoff-actor-authority`
  from the errant commit, reset local `main` back to `origin/main`, and continued all
  subsequent closure work on the branch. No push to `main` occurred at any point.
* Full closure detail: `docs/closure/2026-09-18-024-s-027-f-staging-artifact-handoff-actor-runtime-verification.md`,
  `docs/closure/024-S-027-F-post-merge-closure.md`,
  `.backlogit/reconcile/024-S-pre-20260919T003834Z.md`,
  `.backlogit/reconcile/024-S-post-20260919T003945Z.md`.

## Downstream status (informational only, not mutated)

`019.004-T` depends on `027-F` (now `done`); its own blocking predecessor `025.001-T` (blockers
B3/B4/B5) is unrelated and untouched by this closure. `019.004-T` was not claimed or modified.

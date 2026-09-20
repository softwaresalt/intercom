---
title: "026-S / 029-F — Repair Ship agent execution contract (plan unit 2) — compacted"
date: 2026-09-20
shipment: 026-S
feature: 029-F
pr: 69
merge_commit_sha: ebf9eef0112bf866a77d1b6840a3822cd2fa2997
status: shipped
---

## Mandate

Operator directive: execute shipment `026-S` and carry forward a pre-existing, previously
uncommitted carry-forward payload (`.backlogit/stash.jsonl` ledger plus 5 prior-session
memory/compound artifacts and 2 diagnostic wrapper scripts) via `git stash push -u` /
`apply` (never `pop`), landed in commit `c86a059`. `stash@{0}` retained as rollback
evidence throughout, per operator directive — not dropped without a separate explicit
approval.

## Shipment intake and execution

* `026-S` claimed cleanly (`queued -> active`, cascaded to manifest); topology gates
  (`pre_claim`, `post_claim`) both passed.
* 5 tasks executed in dependency order (`029.001-T` → `029.003-T` → `029.004-T` →
  `029.005-T`; `029.002-T` independent):
  * `029.001-T` — CASCADE contract-surface matrix (6 surfaces, no open edges).
  * `029.002-T` — filled `_ship.agent.md:440`'s empty Format gate slot with
    `test -z "$(gofmt -l .)"`.
  * `029.003-T` — read-only plan-admissibility verification (zero mutation).
  * `029.004-T` — applied PIN-830/PIN-861 verbatim to `_ship.agent.md:830`/`:861-869`,
    re-pointing the CASCADE close branch from a direct `backlogit shipment ship` call to
    `shipment-reconcile mode: safe-close` (Cascade Close Sub-Procedure).
  * `029.005-T` — added `TestCascadeRoutesThroughSafeClose`; confirmed genuine red
    pre-repair, green post-repair; confirmed the pre-existing 33-row
    `TestShipFeatureCompletionContract33Rows` suite stayed green throughout.
* Final quality gates: `go vet`, `gofmt -l .`, `go test ./...` all clean pre-PR.
* PR #69 opened, all 13 CI checks green, local review `READY` (zero P0/P1, no follow-up
  handling required), P-018 gate `NOT_APPLICABLE`, P-009 merge-commit-only confirmed.

## Merge and post-merge closure (this session)

* Operator approval: `PR 69: Merge approved` (scoped to PR #69 only, merge-commit only, no
  admin fallback, not covering any closure PR).
* Independently re-verified HEAD unchanged, CI green, P-018/topology gates passed, merge
  settings compliant → `gh pr merge 69 --merge` → merge commit
  `ebf9eef0112bf866a77d1b6840a3822cd2fa2997` (2 parents, confirmed ancestor of
  `origin/main`).
* Post-merge closure branch: `post-merge/026-s-repair-ship-agent-execution-contract`.
* Covering-feature completion gate (a1): all five conjunctive conditions verified;
  `029-F` `active -> done`.
* Shipment close-path classification: `CLOSE_PATH_VERDICT: CASCADE` /
  `FULLY_COVERED_ROOT` (029-F is a root, fully covered by exactly its 5 manifest tasks at
  every depth). Cascade Close Sub-Procedure invoked
  (`backlogit shipment ship 026-S --sha ebf9eef0...`); two-set gate clean
  (`archived_ids - allowed_ids` empty, `required_ids - archived_ids` empty), all
  `parent_id` preserved, P-007 archive-integrity clean.
* `026-S`: `active -> shipped -> archived`. `029-F` and 5 tasks: archived.
* Full detail in `.backlogit/reconcile/026-S-{pre,classify-close-path,safe-close,post}-*.md`
  and `docs/closure/026-S-029-F-post-merge-closure.md`.

## Outcome

Shipped. No open follow-ups, no deferred P-021 entries, no source-artifact cleanup
applicable (no `custom_fields.source_stash_id`/`source_deliberation_id` on either manifest
top-level member). `stash@{0}` remains retained pending separate explicit operator
approval to drop.

## Key learning

Self-referential closure: this shipment repaired the Ship agent's CASCADE-routing
instruction (`_ship.agent.md:830`/`:861-869`), and its own post-merge closure was the
first execution of that repaired routing — `shipment-reconcile mode: safe-close`
delegating to the Cascade Close Sub-Procedure rather than a direct
`backlogit shipment ship` call from `_ship.agent.md` itself. No discrepancy found between
the merged contract and the in-session instructions at closure time (diff-verified
identical).

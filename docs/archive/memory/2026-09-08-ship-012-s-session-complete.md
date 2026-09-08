---
title: "Ship session — shipment 012-S (013-F, gate scope realignment) complete"
date: 2026-09-08
shipment: 012-S
feature: 013-F
pr: 36
status: complete
---

# Ship Session Summary — Shipment 012-S (013-F: retire D6a gate narrowing)

## Items Completed

- `013.001-T` — extracted `select_repo_paths()`; added 4 structural `--self-test`
  assertions (RED confirmed pre-fix, captured in
  `docs/memory/012-s-013.001-red-evidence.md`). Commit `b464979`.
- `013.002-T` — added Go fixture-suite descriptor, per-suite manifest
  reconciliation, shared `engine_for_path()` dispatch helper. Commit `a6bdd32`.
- `013.003-T` — added 2 Go fixtures (1 reject, 1 accept), secret-scanner-safe.
  Commit `f93eceb`.
- `013.004-T` — broadened scope `internal/config/** -> internal/**`; verified
  zero findings on the broadened live tree. Commit `0b2e3b3`.
- `013.005-T` — fixed P0 dead TOML dispatch for `config.toml.example` via
  `engine_for_path()`. Commit `8f34ec6`.
- `013.006-T` — added D6c amendment (decision doc) + historical-note
  annotations (plan doc), exactly 2 files touched. Commit `aaa15a1`.
- Feature `013-F` and all 6 tasks marked `done`. Commit `c79de78`.
- Adversarial review (3 reviewers, report-only): 1 HIGH + 1 MEDIUM finding
  remediated in-scope (commit `ef74888`); 1 MEDIUM finding independently
  verified accurate; 4 new + 1 reused deferred-scope-expansion stash entries
  captured for out-of-scope LOW findings. Report:
  `docs/closure/2026-09-08-012-s-retire-d6a-gate-narrowing-adversarial-review.md`.
- PR #36 created, Copilot review requested; round 1 found 1 actionable comment
  (stale usage docstring), fixed in-scope (commit `c403e5e`), replied,
  resolved via GraphQL; round 2 `SATISFIED`, 0 unresolved threads.
- All CI checks green at merge HEAD `c403e5e`.
- Merged via merge commit `9d8937ae2bc721cdeb15b74963acb92221843f06`
  (dark-mode pre-authorized, P-009/P-014/P-016/P-018 all satisfied).
- Post-merge: shipment 012-S closed via the P-015 verified fully-covered-root
  CASCADE path (classifier-confirmed); all archive/parent_id/two-set gate
  checks passed. Report:
  `.backlogit/reconcile/012-S-cascade-close-20260908T194800Z.md`.
- Operational closure artifact produced (runtime verification: not
  applicable — CI-tooling/docs-only change). Releasability: READY.
- 1 new compound learning captured (PR review-comment reply REST endpoint
  requires the `pull_number` path segment).

## Items Blocked

None. All 6 tasks, review, CI, merge, and closure completed without a hard
stop.

## Branch State

- Feature branch `feat/012-s-retire-the-d6a-gate-narrowing-broaden-u-e1a-to-internal`
  merged via PR #36 (merge commit `9d8937a`), not deleted (cleanup not
  requested).
- Post-merge closure branch `post-merge/012-s-retire-d6a-gate-narrowing`
  active, carrying: shipment closure commit `7736406`, closure artifact +
  compound learning commit `381c8e0`. Pending: this compaction commit, then
  push + closure PR.
- `main` at `9d8937a` (pre-closure-PR).

## Decisions with Rationale

- Selected the P-015 CASCADE close path over manual safe-close after
  independently running the workspace's own
  `classify_shipment_close_path()` classifier and confirming
  `013-F` is a verified fully-covered root (all 6 children in-manifest, no
  further descendants). This matches the precedent set by shipment 006-S's
  cascade close.
- Remediated 2 of 3 review findings (C-1 HIGH, M-1 MEDIUM) directly as
  in-scope same-contract-surface completions (P-021 C1/C3) rather than
  deferring, since both strengthened self-test assertions this exact
  shipment introduced.
- Deferred 4 clusters of pre-existing LOW findings (token-matching gaps,
  dispatch fail-open generalization, self-test framework hardening,
  residual evasion blind spots) via P-021 C2 stash capture rather than
  fixing, since none were touched by the 6 authorized tasks.
- Fixed the Copilot-flagged usage-docstring staleness directly (in-scope,
  same Usage block this shipment already edits) rather than treating it as
  a duplicate of the already-deferred U-6 finding — the live PR-scoped
  finding on this shipment's own diff takes priority classification over a
  prior local-review bundling decision.
- Did not archive stash `A92E3FA0`/`4989A42D` during source-artifact
  cleanup: no `custom_fields.source_stash_id` field is set on `013-F`, and
  the feature's own description explicitly records both stashes as
  "RETAINED, not archived" pending other completion conditions.

## Next Steps

- Push `post-merge/012-s-retire-d6a-gate-narrowing`, open the post-merge
  closure PR, run local review + §1.9 gate for that PR, obtain fresh
  operator/dark-mode approval, and merge with a merge commit.
- No further work items are pending against shipment 012-S; it is fully
  closed (archived, `archived_status: shipped`).

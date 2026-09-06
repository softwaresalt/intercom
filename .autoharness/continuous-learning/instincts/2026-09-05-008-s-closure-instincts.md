---
generated: 2026-09-05
scope: recent
source_observations: .autoharness/continuous-learning/observations/2026-09-05.jsonl
release_context: 008-S / 009-F (apperr taxonomy correction)
---

# Instincts — 008-S closure session

Clustering of the 3 observations captured during 008-S/009-F build and
closure. This is the first `learn` invocation in this workspace — no prior
instinct store existed. All three patterns below are **single-occurrence**
(evidence count = 1); none meet the configured promotion threshold
(`continuous_learning.promotion_threshold: 3`). None are promoted to
`evolve` at this time — recorded as low-confidence instincts to watch for
repetition in future sessions.

## Instinct 1: backlogit auto-relocates done items to archive/ (terminal relocation, not archived-status)

* **Pattern**: `backlogit move <id> --status done` immediately relocates the
  artifact's file from `.backlogit/queue/` to `.backlogit/archive/`,
  declaring `status: done` (not `status: archived`). This happens per-item,
  independent of any shipment-level close operation.
* **Evidence count**: 1 (this session)
* **Confidence**: low (single occurrence, but high certainty within that
  occurrence — directly observed and verified against file content)
* **Workflow phase**: build / closure
* **Suggested next action**: keep observing. If this recurs in a future
  shipment's closure, consider evolving into an addition to the
  `shipment-reconcile` skill's "pre-archived" classification notes (it
  already tolerates this state, but an explicit callout of *why* it occurs
  would help future sessions recognize it faster instead of re-deriving it
  from file inspection each time).

## Instinct 2: P-015 classifier is importable standalone via PYTHONPATH

* **Pattern**: `autoharness.gates.shipment_closure.classify_shipment_close_path`
  is directly importable by setting `PYTHONPATH` to the installed plugin's
  `src/` directory, even though the `autoharness gate` CLI does not yet
  expose a `shipment-closure` subcommand.
* **Evidence count**: 1 (this session)
* **Confidence**: low
* **Workflow phase**: closure
* **Suggested next action**: keep observing. If this manual-import pattern
  recurs across shipments, consider proposing a `autoharness gate
  shipment-closure` CLI subcommand (mirroring `pipeline-topology` and
  `copilot-review`) so Ship does not need to hand-invoke Python via
  PYTHONPATH during every cascade-eligible closure.

## Instinct 3: batch-then-commit erodes per-task commit granularity

* **Pattern**: Implementing several TDD units back-to-back before the first
  `git add`/commit collapses them into a single commit, requiring a
  post-hoc `git commit --amend` to correct the message (safe only while the
  branch is unpushed).
* **Evidence count**: 1 (this session)
* **Confidence**: low
* **Workflow phase**: build
* **Suggested next action**: keep observing. If this recurs, consider
  evolving into an explicit reminder in `build-feature`/Ship Step 4 to
  commit immediately after each unit's own red/green cycle completes,
  before starting the next unit.

## Promotion status

No instinct in this batch reaches the promotion threshold (3 corroborating
observations). No `evolve` proposal generated this session.

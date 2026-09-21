# Compacted: 027-S/030-F closure-evidence repair (PR #73)

**Shipment:** N/A (no shipment claimed — standalone repair chore) |
**Feature:** N/A (repairs already-archived `030-F` closure evidence) |
**PR:** #73 | **Merged SHA:** `57fc976bf16e8f16b88790af21f45cf7e33dcd38`

## Outcome

Merged via merge-commit strategy on explicit operator approval
("PR 73: merge approved"). Post-merge closure completed on branch
`post-merge/027-s-030-f-closure-evidence-repair`.

## What shipped

Added exactly two frontmatter keys to
`docs/closure/027-S-030-F-post-merge-closure.md`:
`closure_status: READY` and `conditions: []`. Zero narrative changes,
zero other files touched (2 insertions, 0 deletions on the original
diff).

## Root cause

`autoharness.gates.topology._closure_artifact_complete` requires both
`compaction_status` in `{done, degraded}` AND a machine-readable
`closure_status` key. The original closure artifact had
`compaction_status: done` and a "READY" narrative but was missing
`closure_status`, so the predicate fail-closed and the
`pipeline-topology` gate blocked `028-S`..`033-S` with
`PREDECESSOR_CLOSURE_INCOMPLETE`. Fourth recurrence of the
producer/consumer contract-drift class documented in
`docs/bugs/2026-09-06-closure-evidence-producer-consumer-contract-mismatch.md`
(previously repaired for `001-S`, `005-S`, `008-S`) — no new learning,
existing bug doc guidance followed exactly.

## Quality gates

`go vet ./...` clean, `gofmt -l .` clean. Full local build recorded
non-applicable (docs-only diff). Local review: `READY`, P0=0/P1=0. All
required CI checks green at merge HEAD. P-009 (merge-commit-only)
satisfied, no admin fallback. P-018 Copilot-review gate:
`NOT_APPLICABLE` at both pre-merge and last-mile re-verification.
Last-mile HEAD re-check confirmed no drift from reviewed HEAD before
merge.

## Verification effect

`FilesystemTopologyReaders.closure_complete('027-S')` now returns
`True`; the `pipeline-topology` `pre_claim` gate for `028-S` no longer
returns `PREDECESSOR_CLOSURE_INCOMPLETE` on account of `027-S`.

## Post-merge closure

* Post-merge closure branch: `post-merge/027-s-030-f-closure-evidence-repair`
* Closure artifact: `docs/closure/027-S-030-F-closure-evidence-repair-post-merge-closure.md`
  (releasability: READY)
* No shipment claim, no backlog archival mutation this session (`027-S`
  was already archived before the session began).
* No new P-021 deferred-scope-expansion follow-ups identified.
* `028-S`..`033-S` remain untouched/unclaimed per explicit scope
  boundary — starting any of them is a separate, later invocation.

## Session history compacted

Original superseded:
`2026-09-20-ship-027-s-030-f-closure-evidence-repair-pr73-awaiting-approval.md`
(moved to `docs/archive/memory/`).

---
title: "Closure Records Compacted — Band 1 (2026-09-03 to 2026-09-07) Index"
description: "Index summary for 14 closure records archived to docs/archive/closure/ under task 063.001-T"
status: "complete"
tags:
  - "closure"
  - "compaction"
  - "archive"
date: 2026-09-07
mode: compaction
shipment: 055-S
feature: 063-F
task: 063.001-T
compaction_status: done
closure_status: READY
releasability: READY
---

# Closure Records Compacted — Band 1

Fourteen closure records (frozen list D1 in
`docs/plans/2026-10-18-intercom-go-stage-batch-bcd-plan.md`) were moved with
`git mv` from `docs/closure/` to `docs/archive/closure/`. Contents are
byte-identical; nothing was deleted. This file is an index, not a re-narration.
Read the archived record for detail.

## Index

All records below were closed `READY` unless noted; the archive path is
`docs/archive/closure/<name>`.

| Original path (`docs/closure/`) | Finding |
|---|---|
| `001-S-001-F-post-merge-closure.md` | 001-S foundation (PR 4): READY; review `READY_WITH_FOLLOWUPS` with P0=0/P1=0; 7 out-of-scope findings stashed under P-021 C2. |
| `002-S-002-F-post-merge-closure.md` | 002-S error taxonomy and path containment (PR 8): READY at reviewed HEAD; deferred machine-checkable oracle-drift gate follow-up. |
| `003-S-003-F-post-merge-closure.md` | 003-S configuration package (PR 11): READY; review findings remediated in-session, re-review READY with no residual. |
| `004-S-005-F-post-merge-closure.md` | 004-S architecture correction (PR 15): READY; review `READY_WITH_FOLLOWUPS`; copilot-review gate verified twice. |
| `005-S-006-F-post-merge-closure.md` | 005-S design reconciliation and Copilot SDK spike (PR 17): READY; BLOCKED verdict on 9 findings, all remediated and re-verified. |
| `2026-09-04-007-f-ci-supply-chain-hardening-pass2-adversarial-review.md` | 007-F pass-2 adversarial review: FAIL verdict; job-level `continue-on-error` masked all step failures; blocking set #1-#3. |
| `006-S-007-F-post-merge-closure.md` | 006-S CI supply-chain hardening (PR 20): READY; pass-2 FAIL findings remediated (continue-on-error scoped to the verdict step). |
| `007-S-008-F-post-merge-closure.md` | 007-S build and scan script hardening (PR 22): READY; multi-line TOML fixtures and verdict manifest authored before implementation. |
| `008-S-009-F-post-merge-closure.md` | 008-S apperr taxonomy correction (PR 24): READY; review READY with 0 findings. |
| `009-S-010-F-post-merge-closure.md` | 009-S pathsafe and config containment hardening (PR 27): READY; scope-expansion follow-ups captured. |
| `2026-09-06-010-s-011-f-residual-hardening-round2-adversarial-review.md` | 010-S round-2 adversarial review (4 reviewers, consensus); some platform coverage (darwin) left pending verification. |
| `010-S-011-F-post-merge-closure.md` | 010-S residual hardening round 2 (PR 29): READY; deferred follow-ups listed in the record. |
| `011-S-012-F-post-merge-closure.md` | 011-S pathsafe containment correctness (PR 34): READY; initial review BLOCKED on one escalated CRITICAL finding, then remediated. |
| `2026-09-07-011-s-012-f-pathsafe-containment-adversarial-review.md` | 011-S adversarial review: verdict BLOCKED; junction-following containment defect; pre-existing accepted fail-open risk noted. |

## Archive locations

Each original `docs/closure/<name>` now lives at `docs/archive/closure/<name>`
with the same file name.

## Inbound references to moved paths

Found with `git grep -F <name>` across the repository at compaction time.
Not rewritten (history and plans are left as written); resolve via the archive
path. Every moved file name is also named in the shipment plan's Appendix D1
and in the harness test `tests/integration/closure_band1_compaction_test.go`,
which are omitted from the per-file lists below.

* `001-S-001-F-post-merge-closure.md`: `.backlogit/archive/stash.jsonl`, `docs/bugs/2026-09-06-closure-evidence-producer-consumer-contract-mismatch.md`, `docs/decisions/2026-09-06-intercom-go-residual-hardening-round2-deliberation.md`
* `002-S-002-F-post-merge-closure.md`: `.backlogit/archive/stash.jsonl`, `docs/archive/memory/2026-09-03-ship-intercom-go-apperr-pathsafe.md`, `docs/archive/memory/2026-09-03-ship-resume-merge-post-merge-closure-002-s.md`, `docs/decisions/2026-09-03-intercom-go-p2-config-deliberation-addendum.md`, `docs/memory/compacted/2026-09-03-002-s-intercom-go-apperr-pathsafe-compacted.md`, `docs/plans/2026-09-03-intercom-go-p2-config-plan.md`
* `003-S-003-F-post-merge-closure.md`: `.backlogit/archive/stash.jsonl`, `docs/archive/memory/2026-09-03-ship-intercom-go-p2-config.md`, `docs/memory/compacted/2026-09-04-003-s-intercom-go-p2-config-compacted.md`, `docs/plans/2026-09-04-intercom-go-architecture-correction-plan.md`
* `004-S-005-F-post-merge-closure.md`: `.backlogit/archive/stash.jsonl`, `docs/archive/memory/2026-09-04-ship-004-S-pre-merge-checkpoint.md`, `docs/archive/memory/2026-09-04-stage-first-round-c2-reconciliation.md`, `docs/decisions/2026-09-04-intercom-go-implementation-design-reconciliation-deliberation.md`
* `005-S-006-F-post-merge-closure.md`: none beyond the plan and test
* `2026-09-04-007-f-ci-supply-chain-hardening-pass2-adversarial-review.md`: none beyond the plan and test
* `006-S-007-F-post-merge-closure.md`: `docs/memory/compacted/2026-09-05-006-s-ci-supply-chain-hardening-compacted.md`
* `007-S-008-F-post-merge-closure.md`: none beyond the plan and test
* `008-S-009-F-post-merge-closure.md`: `.backlogit/archive/stash.jsonl`, `docs/archive/memory/2026-09-06-004136-ship-pr26-closure-repair-merged.md`, `docs/bugs/2026-09-06-closure-evidence-producer-consumer-contract-mismatch.md`
* `009-S-010-F-post-merge-closure.md`: `docs/archive/memory/2026-09-06-214700-ship-009-s-post-merge-closure-complete.md`, `docs/memory/compacted/2026-09-06-009-s-compacted.md`
* `2026-09-06-010-s-011-f-residual-hardening-round2-adversarial-review.md`: none beyond the plan and test
* `010-S-011-F-post-merge-closure.md`: `docs/decisions/2026-09-08-intercom-go-retired-arch-gate-detection-quality-deliberation.md`
* `011-S-012-F-post-merge-closure.md`: `.backlogit/archive/stash.jsonl`, `docs/archive/memory/2026-09-08-ship-011-s-session-complete.md`, `docs/decisions/2026-09-08-intercom-go-pathsafe-reparse-containment-deliberation.md`
* `2026-09-07-011-s-012-f-pathsafe-containment-adversarial-review.md`: `.backlogit/archive/stash.jsonl`, `docs/decisions/2026-09-08-intercom-go-pathsafe-reparse-containment-deliberation.md`, `docs/memory/compacted/2026-09-08-011-s-pathsafe-containment-compacted.md`, `docs/plans/2026-09-08-intercom-go-pathsafe-reparse-containment-plan.md`, `internal/pathsafe/root.go` (code comment; left unchanged)

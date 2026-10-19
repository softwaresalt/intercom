---
title: "Post-merge closure: 011-S -- pathsafe containment correctness (PR #34)"
description: "Operational closure artifact for shipment 011-S / feature 012-F"
status: "complete"
tags:
  - "closure"
  - "011-S"
  - "012-F"
  - "post-merge"
date: 2026-09-08
mode: post-merge
shipment: 011-S
feature: 012-F
pr: 34
merge_commit_sha: 54d27ba953a068cea539a1f691e5c4ad1231d2ba
compaction_status: done
closure_status: READY
releasability: READY
---

# Post-Merge Closure — 011-S / 012-F: Pathsafe Containment Correctness and Cross-Platform Coverage

## Incident context

This shipment's staging artifacts were originally committed directly to
`main` by Stage as `f6d5879` (P-001/P-016 violation). Recovery sequence,
all merged to `main` via reviewed PRs with merge commits:

1. PR [#31](https://github.com/softwaresalt/intercom/pull/31) — corrective
   revert of `f6d5879` (merge commit `68d0f1659d4a8d327251aca74d105d4b86f8aa47`).
   Preservation branch `origin/chore/stage-011-s-preserve` retained the
   exact SHA throughout.
2. PR [#32](https://github.com/softwaresalt/intercom/pull/32) — unblocked
   the `011-S` pre-claim gate by fixing a pre-existing, unrelated,
   upstream-owned closure-artifact filename contract-drift defect for
   `010-S` (merge commit `64fdee7af0361fd046dd7baa7532467d3d7fe398`).
3. PR [#33](https://github.com/softwaresalt/intercom/pull/33) — proper
   reapplication of the preserved staging artifacts via branch → PR →
   review → CI → Copilot → merge commit (merge commit
   `42b9530c6f19a09350199c2102a1a98f307ac331`).
4. PR [#34](https://github.com/softwaresalt/intercom/pull/34) — this
   shipment's implementation (below).

## Summary

Shipment `011-S` (feature `012-F`, "pathsafe containment correctness and
cross-platform coverage") merged to `main` via PR
[#34](https://github.com/softwaresalt/intercom/pull/34), merge commit
`54d27ba953a068cea539a1f691e5c4ad1231d2ba`. All 9 units
(`012.001-T`–`012.009-T`) completed, reviewed, and shipped.

## Dispositions

| Task | Disposition | Commit(s) |
|---|---|---|
| 012.001-T | Done (RED: ancestor-walk regression tests) | `f04b332` |
| 012.002-T | Done (RED: Windows junction containment test) | `f04b332` |
| 012.003-T | Done (GREEN: depth-scoped probe + ENOENT-only ascent) | `f04b332`, `9158a05` (review-driven redundant-code removal) |
| 012.004-T | Done (docs: GO-14 risk-register boundary) | `4605076` |
| 012.005-T | Done (RED: stripUNCPrefix regression tests) | `0340a7c`, `902c05f` (2 additional share-less regression cases) |
| 012.006-T | Done (GREEN: stripUNCPrefix fail-closed postcondition) | `0680667`, `902c05f` (Copilot-review fix: share-less UNC now correctly rejected) |
| 012.007-T | Done (docs: case-folding risk-register entry) | `451fe4c` |
| 012.008-T | Done (fail-closed integration symlink helper) | `a1a7c34`, `902c05f` (Copilot-review fix: pwsh LookPath skip guard) |
| 012.009-T | Done (truncated test doc-comment repair) | `2e5f661` |

Plus bookkeeping commits `f8ac28d` (task archival + session memory) and
`165704a` (adversarial-review findings + risk-register entry for the
newly-discovered live-junction gap, see below).

## Review outcomes

**Local multi-persona review** (before PR): Security Reviewer
(READY_WITH_FOLLOWUPS — 1 P3 redundant-code finding, fixed), Correctness
Reviewer (READY — all 9 tasks' acceptance criteria verified), Go Reviewer
(READY_WITH_FOLLOWUPS — idiom nits, fixed), Scope Boundary Auditor
(READY — anti-goals honored, zero scope creep), Maintainability Reviewer
(READY_WITH_FOLLOWUPS — redundant-code finding with verified-safe
simplification, applied).

**Adversarial multi-model review** (4-model consensus, report-only):
initial verdict **BLOCKED** on one escalated CRITICAL finding — a live
(non-dangling) Windows directory junction at the final path component
bypasses `checkSymlinkEscape` entirely. Ship independently verified this
empirically (live reproduction on the build host, confirmed reading an
outside-root file through an accepted path) before determining
disposition, and confirmed via `git show main:internal/pathsafe/pathsafe.go`
that the exact same bypass mechanism is **pre-existing in `main`, not
introduced or worsened by this shipment**. `012-F`'s chartered scope
(stash entries `ECE3DAB7`/`CA469B3D`/`5FE4A7BE`/`5158769F`/`805248F7`) is
explicitly the **dangling** intermediate case only; the live-junction case
was never named in the feature's scope or anti-goals and requires its own
red-phase-first design work. Disposed per **P-021 C1/C2**: documented in
`internal/pathsafe/root.go`'s risk register (entry `700B41CE`), full trace
recorded in
`docs/closure/2026-09-07-011-s-012-f-pathsafe-containment-adversarial-review.md`,
and captured as deferred-scope stash entries `700B41CE` (critical,
requires Stage deliberation) and `AD0D9D1F` (medium, 7 bundled advisory
findings). Two of the bundled advisory findings (share-less UNC edge case,
optional-tooling skip-vs-fail convention) were independently raised again
by Copilot review on PR #34 and fixed directly in commit `902c05f` since
they were same-surface, mechanical, and low-risk.

**Copilot review** (PR #34): 2 comments, both fixed, replied to, and
threads resolved (commit `902c05f`). P-018 gate: SATISFIED for final HEAD
`902c05f`.

**CI**: all required and advisory jobs green (cross-compile ×4 platforms,
lint 0 issues, security, test, test-windows-advisory, gitignore regression
check, pipeline-topology ambient).

## Quality gates (independently re-verified by Ship, not solely delegated)

* `go build ./...` — pass
* `go vet ./...` — pass (0 issues)
* `golangci-lint run ./...` — pass (0 issues)
* `go test ./...` — pass (all packages)
* `go test -race ./internal/pathsafe/... ./tests/integration/...` — pass

## Runtime verification

No runtime surface adapters were touched (pure internal library security
fix, `internal/pathsafe` + its tests; no CLI flag, config schema, or wire
protocol change). Per `.autoharness/workspace-profile.yaml`'s required
`cli-smoke` probe, ran the CLI cold-start check as defense-in-depth:

```
$ go run ./cmd/intercom --help       # exit 0, usage printed, no panic
$ go run ./cmd/intercom-ctl --help   # exit 0, usage printed, no panic
```

Both processes start and print usage without panic. Full validator-manifest
pass beyond this smoke probe is not applicable — no other runtime surface
in the manifest is touched by this shipment's diff.

## Releasability

**READY.** No monitoring changes required (no new runtime surface). No
rollback beyond standard `git revert` of the merge commit if a regression
is later discovered (none is expected — all quality gates green, zero
regressions relative to `main` confirmed for the newly-discovered residual
risk). No new operational owner assignment needed (existing
`internal/pathsafe` ownership unchanged). No validation window required
beyond the CI/review gates already completed pre-merge.

## Deferred follow-ups (stashed for Stage)

* `700B41CE` (critical, requires deliberation) — live Windows junction at
  final path component bypasses containment; pre-existing, not introduced
  by this shipment; needs its own red-phase-first design.
* `AD0D9D1F` (medium, requires deliberation for 3 of 7 bundled items) —
  remaining adversarial-review advisory findings (POSIX ENOTDIR test-
  coverage gap, fail-open defensive-branch direction, `GetFinalPathNameByHandle`
  doc misattribution, risk-register entry clarity, untested
  permission-denied/cycle branch). 2 of the original 7 (share-less UNC,
  pwsh skip-vs-fail) were fixed directly in PR #34 per Copilot review and
  are no longer open.
* Pre-existing stash entry `F133AB7E` (GO-14 reconsideration) and
  `2362BBB5` (P3 pathsafe residuals) remain open from this shipment's own
  plan review — unchanged, not newly introduced.

## Source artifact cleanup

`012-F`'s `custom_fields.source_stash_id` / `source_deliberation_id` were
not populated on the queue record at claim time (Stage-authored staging
predates this convention for this artifact); the covering deliberation
(`docs/decisions/2026-09-07-intercom-go-pathsafe-containment-correctness-deliberation.md`)
and plan (`docs/plans/2026-09-07-intercom-go-pathsafe-containment-correctness-plan.md`)
are retained as historical record, not archived, consistent with how
`010-S`'s equivalent artifacts were handled.

## Compaction status

**done.** Mandatory P-020 `compact-context` invocation performed in this
closure session: this shipment's own fresh session memory (the build-phase
checkpoint written under `docs/memory/2026-09-07/` during implementation)
was the intended Tier-1 candidate (eligible under the completed-work rule),
compacted to
`docs/memory/compacted/2026-09-08-011-s-pathsafe-containment-compacted.md`,
and the verbose original relocated to its current, sole location at
`docs/archive/memory/2026-09-07-ship-pathsafe-containment-shipment-011-memory.md`
(the only path at which this file now exists in the repository).
No other `docs/memory/`, `docs/plans/`, or `docs/closure/` artifacts met
the threshold/completed-work candidate criteria this cycle. A durable
compound-learnings entry was also captured:
`docs/compound/2026-09-08-adversarial-review-empirical-verification-and-scope-check.md`.
No existing `docs/compound/` entries referenced pathsafe/symlink/junction
topics, so no compound-refresh consolidation was needed.

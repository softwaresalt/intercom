---
title: "Post-merge closure: 030-S / 033-F — Unify retired-architecture scan scope declaration"
description: "Post-merge closure artifact for shipment 030-S / feature 033-F"
status: "complete"
tags:
  - "operational-closure"
  - "030-S"
  - "033-F"
  - "post-merge"
date: 2026-10-02
mode: post-merge
shipment: 030-S
feature: 033-F
pr: 95
merge_commit_sha: e1d61a7bb60c89f503622a88c0ae0f437631e7e6
compaction_status: done
closure_status: READY
releasability: READY
---

# Post-merge closure: 030-S / 033-F — Unify retired-architecture scan scope declaration

## Summary

Shipment `030-S` delivered feature `033-F`, "Unify retired-architecture scan scope
declaration" (post-M4 plan unit 6, Unit A rev 6). PR #95 merged to `main` at
`e1d61a7bb60c89f503622a88c0ae0f437631e7e6` on 2026-10-02.

The merge hardens the retired-architecture gate's scan-scope pin in
`tools/gatecheck/internal/retiredarch/`:

* **ALP-1 (`15b0aeb`, 033.001-T + 033.002-T as one commit):** the scan scope is anchored
  in a single `scanScope` declaration and repinned. The commit body carries the AC-A1.9
  §A-CANON conformance checklist.
* **033.004-T (`f48135c`):** `containsAll` is non-vacuous on empty input.
* **033.008-T (`2bfbec5`):** the selection self-test probes a non-test `cmd/x/main.go`.
* **033.005-T (`205e1d7`):** the selection pin freezes the pathspec declarations and shared
  rules.
* **033.006-T (`825da4b`):** the selection pin freezes the prefix surface and adds
  `scopeDataOK`.
* **033.007-T (`9bfcffb`):** package closure (A-T3d).
* **033.003-T (`ac22deb`):** authorized verified no-op closure (D-000-2, item-scoped P-002
  skip D-000-3), closed `queued -> active -> done` after 033.007-T.
  `.backlogit/hooks.yaml` is unchanged.

Every hardening narrows toward fail-closed (INV-2). No gate verdict changes except where an
acceptance criterion names it (INV-1). No destructive write primitive was added (INV-3), and
lint stays clean (INV-4).

The change is internal CI and developer gate tooling only. It does not touch Intercom's
product CLI, TUI, WebSocket/API or deployment surfaces, so no production deployment or data
migration is required.

## Invariants to preserve

* **IVL-1:** the bounded INV-2 interval existed only on the 030-S branch. It closed before
  merge under AC-A3d.6, because 033.004-T through 033.008-T were all `done`. `main` never
  observed it.
* The retired-architecture scan scope has a single declaration. The selection pin fails
  closed if that declaration, its pathspecs, its prefix surface or its package closure drift.
* `containsAll` never accepts an empty required set as satisfied.

## Validator evidence

| Gate | Result |
|---|---|
| `gofmt -l .` (verified on LF content) | Clean |
| `go vet ./...` | Passed |
| `golangci-lint run ./...` | 0 issues |
| `staticcheck ./...` | Passed |
| `go test ./...` | Passed (Git for Windows Bash first on `PATH`) |
| `go test -race ./tools/gatecheck/...` | Passed |
| `go build ./...` | Passed (full local build) |
| Full gate script | `GATE_RESULT fail=0` |
| PR #95 CI | All 14 checks passed |
| First post-merge `main` CI | Passed; run 37059148788 at the merge SHA |

### Runtime validation applicability

Runtime verification was not run, because this shipment changes only internal gatecheck
tooling. The workspace profile's product runtime probes and manual checkpoints do not apply.
The relevant validator evidence is the retiredarch unit tests, the selection self-test and the
CI gate jobs.

## Review, CI, and merge evidence

* **Local adversarial review: one cycle** at pinned HEAD `ac22deb`, from the pinned patch
  file. Each reviewer printed `REVIEWED_HEAD ac22deb`.
  * gpt-5.6-sol (anchor) raised one P1. It is a pre-existing gap on `main@698e2bc`: package
    closure does not constrain sibling `init` process side effects. On re-adjudication the
    anchor ruled `ACCEPT_DEFERRAL` (P-021 C1 out of scope, needs a new rule). It was captured
    as `3750C37C`.
  * claude-opus-5.5 returned `READY_WITH_FOLLOWUPS` with one P2: the SEC4-3 env rule misses
    Windows `syscall.SetEnvironmentVariable`. It was captured as `0ECC1895`.
  * gemini-3.8-flash returned `READY` with no findings.
  * Final result: P0 = 0, P1 = 0 unresolved, outcome `READY_WITH_FOLLOWUPS`.
* The delta after review was re-reviewed locally, with gates re-run green. It was `1cbedc9`
  (two stash captures only) and `d40da16` (doc comments only).
* **Copilot on PR #95: two iterations.**
  * Iteration 1 raised two comment-accuracy threads. Both were fixed in `d40da16`, then
    replied to citing the SHA and resolved.
  * Iteration 2 had no new findings.
  * The P-018 gate returned `SATISFIED`.
* The PR body recorded `READY_WITH_FOLLOWUPS` for current HEAD `d40da16`, with full-build
  evidence and follow-ups `3750C37C` and `0ECC1895`.
* P-016 topology passed (pre_claim, post_claim, lifecycle before build and before the PR, and
  `a0`).
* P-009 settings allow merge commits only.
* P-014 approval came from the scoped `DARK_MODE_ACTIVE` activation
  (`merge_approval_pre_authorized: true`).
* The merge used the normal `gh pr merge 95 --merge --delete-branch` path. Admin fallback was
  neither authorized nor used.

## Shipment reconciliation and backlog closure

**Feature completion gate (a1).**

* S0 found no anomalies.
* There was exactly one feature member (`n = 1`).
* All five S4 conditions held:
  * all eight tasks were done;
  * no live descendant remained;
  * the descendant set plus `033-F` equalled the manifest;
  * containment held;
  * `033-F` was live `active`.
* Ship moved `033-F` from `active` to `done` and verified the re-read.

**Pre-mode.** `shipment-reconcile` pre-mode (`expected_status: done`) returned `PROCEED`.
The record was active and record-consistent, every manifest item was `done`, and there
were no orphans. The eight tasks were already in `.backlogit/archive/`. They were relocated
by `f67fc3d` and by the 033.005-T through 033.008-T feature commits. `033-F` was still in
`.backlogit/queue/` with `status: done` after the a1 transition.

**Location-defect disclosure (found in post-close review).** The helper that encoded the
classification and safe-close snapshots recorded the `033-F` location as `archive`. The
real pre-close location was `queue`. Both snapshots came from the same helper, so the
Step 0 match held on every field except location. The binding recomputed with the correct
`033-F` encoding (`done` / `queue`) is
`72ce320618fa8abce9525e3b66928b9f41111a56bea4390f7013173cd7bf0deb`.

Location is not a coverage input, and the root and descendant topology are unchanged. So
the `CASCADE` / `FULLY_COVERED_ROOT` verdict and the `CLOSED` outcome stay valid. A review
claim that `033-F` was still `active` at pre-mode is refuted by `.backlogit/logs/033-F.jsonl`:
the a1 transition to `done` was recorded at 20:14:52Z, and pre-mode ran at 20:16:29Z.

**Classification.** Classify-close-path returned `CASCADE` / `FULLY_COVERED_ROOT` with
`CLASSIFICATION_BINDING: 6464e6ed004517c40fae3b1b4ed393f061b36fcd786089b53bb3013c83b37f8f`.
The binding engine line is `backlogit version 1.11.0`.

**Safe-close.** Safe-close recomputed the binding, found it identical, and entered the Cascade
Close Sub-Procedure. It returned `CLOSED`:

* `returned_ids` was `[]`.
* `archived_ids` = `allowed_ids` = `required_ids` = {`030-S`, `033-F`,
  `033.001-T` … `033.008-T`}.
* Both set differences were empty.
* Every task kept `parent_id: 033-F`.
* `030-S` is archived with `archived_status: shipped`, with `commit` set to the merge SHA.
  `033-F` is archived with `archived_status: done`.

**Post-mode.** Post-mode returned `PROCEED`. All archive files were present, and the P-007
check found no archive deletions.

**Lock.** The `file-lock` skill's bundled script
(`.github/skills/file-lock/scripts/acquire_lock.ps1`) held `.backlogit/queue/030-S.md` from
pre-mode through post-mode, and `release_lock.ps1` then released it. This resolves the 032-S hand-rolled-lock deviation pattern for this shipment. The
agent template's root `scripts/acquire_lock.ps1` path is still missing; that conflict stays
Stage-owned under `9F824B64`.

**Index engine drift.** The first pre-mode attempt found a backlogit index that 1.11.0 could
not read, because a stale 1.5.0 binary earlier on `PATH` had rebuilt it. Ship put the 1.11.0
binary first on `PATH` and re-ran `backlogit sync` before classification. The index is a
disposable cache, and the Markdown records were unaffected. The learning is captured in
`docs/compound/2026-10-02-backlogit-stale-binary-on-path-breaks-index.md`.

**Reports.**

* `.backlogit/reconcile/030-S-pre-2026-10-02T20-16-29Z.md`
* `.backlogit/reconcile/030-S-safe-close-2026-10-02T20-18-49Z.md`
* `.backlogit/reconcile/030-S-post-2026-10-02T20-18-49Z.md`

The backlog closure was committed on the post-merge branch as `d2e256c`
(`chore(ci): archive 030-S backlog artifacts`), never directly to `main`.

## Source artifact cleanup

`033-F` has neither `custom_fields.source_stash_id` nor
`custom_fields.source_deliberation_id`. Under the manifest-derived cleanup rule, no source
stash entry or deliberation was archived. Stash IDs that appear only in plan or feature prose
were not inferred or touched.

## Pre-deploy audits and rollout path

No data migration, feature flag, access change or production deployment is required.
Rollout is merge-only: the change takes effect in CI and in developer gate runs on `main`.

## Post-merge checks and monitoring

* First `main` CI for merge SHA `e1d61a7bb60c89f503622a88c0ae0f437631e7e6` passed
  (run 37059148788).
* Monitor the retired-architecture gate and the gatecheck test jobs on the next two or three
  PRs that touch `tools/gatecheck/`, `cmd/` or `internal/`.

### Healthy signals

* The retiredarch unit tests, the selection self-test and the CI gate jobs stay green.
* Ordinary Go changes under `cmd/` and `internal/` do not trip the scan-scope pin.

### Failure signals

* The scan-scope pin passes after a scope declaration, pathspec, prefix or package-closure
  edit that should have been caught.
* The pin fails on an unrelated change that leaves the scan-scope surface untouched (a false
  positive).

## Risky action record

* **Shipment closure:** the bound P-015 `CASCADE` path was selected only after full-root and
  full-descendant coverage was verified. Its evidence is recorded above: the two-set gate,
  empty `returned_ids`, preserved parent links and shipped provenance.
* **Merge:** the normal merge-commit path under the scoped P-017 authorization. No `--admin`
  attempt was made.
* **Index rebuild:** `backlogit sync` rebuilt a disposable cache only. No Markdown record was
  rewritten by it.

## Rollback trigger and procedure

Trigger a rollback if the retired-architecture gate starts failing on changes that leave the
scan-scope surface untouched, or if the pin is shown to accept a scope drift it should reject.
Open a merge-commit revert PR for PR #95 and rerun the retiredarch tests. Do not revert
directly on `main`.

## Validation window and owner

Watch the next two or three PRs that exercise `tools/gatecheck/`, `cmd/` or `internal/`.
Owner: Ship and the repository maintainer (`softwaresalt/intercom`).

## Follow-ups (Stage-owned)

* `3750C37C`: deferred scope expansion (P-021 C2). Package closure does not constrain
  sibling `init` process side effects. Pre-existing; needs deliberation.
* `0ECC1895`: deferred scope expansion (P-021 C2). The SEC4-3 env rule misses Windows
  `syscall.SetEnvironmentVariable`. Needs deliberation.
* Both were captured threadless during local review, and both reached `main` in PR #95
  (`1cbedc9`).
* `DC921AF6` (R-A1) and `D7BF9F74` (R-A2) were already captured before this shipment.
  They were deliberately not implemented.
* No other follow-up was identified by local review, Copilot or this closure.

## Knowledge graduation

No architecture or product-spec change is required.

* **New learning:** captured in
  `docs/compound/2026-10-02-backlogit-stale-binary-on-path-breaks-index.md`.
* **Compound refresh:** no existing `docs/compound/` entry covers the retiredarch scan-scope
  pin or backlogit binary drift, so none became stale.

## Compaction status (P-020)

**`done`.** Ship invoked `compact-context` with `target: all`, and it succeeded without
degradation.

* **Memory:** the 030-S session memory
  (`docs/archive/memory/2026-10-02/030-s-ship-memory.md`) was compacted into
  `docs/memory/compacted/2026-10-02-030-s-033-f-compacted.md`. The original was archived, not
  deleted.
* **Plans:** the 2026-10-01 post-M4 re-plan was skipped as a decided-plan candidate, because
  later units (031-S, 039-S) are still open.
* **Closure records:** none have passed the threshold age.
* **Checkpoints:** no active checkpoint was compacted.
* **Residual pool:** two completed-unit addenda remain in `docs/memory/2026-09-30/` (034-S
  and 035-S session-end notes). They were disclosed as residuals by their own closures and are
  left for a dedicated compaction pass.

## Closure PR gate status

The closure PR is opened from branch
`post-merge/030-s-unify-retired-architecture-scan-scope-declaration` after archival.

* **Topology:** `a0` passed before safe-close. Following the Orchestrator ruling from the
  PR #92 and PR #94 precedents, the closure PR's topology evidence is
  `autoharness gate pipeline-topology --mode manual --phase ambient --json`, not the
  lifecycle phase.
* **Review:** the closure diff gets a local three-model review before the PR.
* **Gates:** P-018, required CI and the §1.9 readiness gate apply. The merge uses a normal merge
  commit; admin fallback is not authorized.

## Releasability evidence

**READY.** This is internal CI and developer tooling, with no change to the product runtime or
deployment surface.

* The multi-model adversarial review ran at pinned `ac22deb` and returned
  `READY_WITH_FOLLOWUPS` (P0 = 0, P1 = 0 unresolved), with both follow-ups captured.
  * The post-review delta was re-reviewed locally with gates green, rather than by a second
    multi-model cycle. That delta was `1cbedc9` (stash captures) and `d40da16` (doc comments).
  * The PR readiness block recorded `d40da16` as the current HEAD.
* Copilot review completed and P-018 returned `SATISFIED`.
* All PR checks and post-merge CI checks passed.
* P-009 and P-016 passed.
* The merge used the authorized merge-commit path.
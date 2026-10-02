---
title: "Post-merge closure: 031-S / 034-F — Harden write-path gate, masked-text increment"
description: "Post-merge closure artifact for shipment 031-S / feature 034-F"
status: "complete"
tags:
  - "operational-closure"
  - "031-S"
  - "034-F"
  - "post-merge"
date: 2026-10-02
mode: post-merge
shipment: 031-S
feature: 034-F
pr: 97
merge_commit_sha: c04d9751d0ce98d0e10bf5d498985d2ddf8b73ee
compaction_status: done
closure_status: READY
releasability: READY
conditions: []
---

# Post-merge closure: 031-S / 034-F — Harden write-path gate, masked-text increment

## Summary

Shipment `031-S` delivered feature `034-F`, "Harden write-path gate, masked-text increment"
(post-M4 plan Unit D). PR #97 merged to `main` at
`c04d9751d0ce98d0e10bf5d498985d2ddf8b73ee` on 2026-10-02 (22:24:10Z).

The merge hardens the write-path precondition gate in `tools/gatecheck`:

* **034.010-T (`c1bebe5`, D-T1a):** a frozen differential oracle for write-path detection
  parity.
* **034.002-T (`fbf03f0`, D-T1):** an occurrence cursor plus a call-extent extractor over
  the masked text.
* **034.003-T (`14dbe67`, D-T2):** an access-mode allowance predicate for metadata-only
  `syscall.CreateFile` calls.
* **034.004-T (`758a11c`, D-T3):** six new decidable write-primitive selectors (26 in total).
* **034.007-T (`0862525`, D-T4):** documents the residual evasion surface.
* **034.008-T (`358e470`, D-T5a):** verdict-boundary fixtures. Mutation-tested: forcing
  `occurrenceAllowed` to false fails the accept fixture.
* **034.011-T (`fdf0de6`, D-T5b):** presence fixtures for the new selectors (B72E9715).
* **034.009-T (`fe8c284`, D-T6):** characterization regression fixtures for the
  pre-existing `os` selectors, including `os.Chtimes` (D-031-6, 8E9F8E55).
* **Review fixes:** `34386ad` (tighter fail-closed cursor and masked tag-row tests) and
  `2144414` (a corrected stale comment).

Every hardening narrows toward fail-closed (INV-2). No gate verdict changes except where an
acceptance criterion names it (INV-1). Write primitives appear only in gate testdata fixtures,
and no destructive write primitive was introduced into `internal/**` or `cmd/**` (INV-3). Lint
stays clean (INV-4). ALP-1 was not touched (INV-5).

The change is internal CI and developer gate tooling only. It does not touch Intercom's
product CLI, TUI, WebSocket/API or deployment surfaces, so no production deployment or data
migration is required.

## 034.001-T authorized no-op closure

`034.001-T` (D-T0) was closed inside the shipment as an authorized verified no-op (D-000-2,
item-scoped P-002 skip D-000-3).

* It moved `queued -> active -> done` in commit `7863eae`.
* The pre-task-completion gate evidence (`pre_task_completion_gate_passed`) was taken at
  `8a517bb`, the `main` HEAD the shipment branched from.
* `.backlogit/hooks.yaml` is unchanged.

## Invariants to preserve

* The write-path gate locates every selector occurrence through the occurrence cursor over
  masked text. A string or comment occurrence never counts as a call.
* The `CreateFile` access-mode allowance accepts only metadata-only access. Any write-capable
  or undecidable access mode fails closed.
* The frozen differential oracle pins the detection set. A selector that disappears or
  changes verdict fails the oracle.

## Validator evidence

| Gate | Result |
|---|---|
| `gofmt -l` (verified on LF content) | Clean |
| `go vet ./...` | Passed |
| `golangci-lint run ./...` | 0 issues |
| `staticcheck ./...` | Passed |
| `go test ./... -count=1` | Passed (Git for Windows Bash first on `PATH`) |
| `go test -race ./tools/gatecheck/... -count=1` | Passed |
| `go build ./...` | Passed (full local build) |
| `gatecheck write-path --root .` | Passed bare, with `--self-test` and with `--self-test-integrity` |
| `scripts/check-write-path-precondition.sh` | Passed bare and with `--self-test` |
| PR #97 CI | All 14 checks passed |
| First post-merge `main` CI | Passed; run 37072295122 at the merge SHA |

### Runtime validation applicability

Runtime verification was not run, because this shipment changes only internal gatecheck
tooling. The workspace profile's product runtime probes and manual checkpoints do not apply.
The relevant validator evidence is the gatecheck unit tests, the gate self-tests and the CI
gate jobs.

## Review, CI, and merge evidence

* **Local adversarial review: three cycles.** Each cycle used three reviewers on three
  models: anchor gpt-5.6-sol, claude-opus-5.5 and gemini-3.8-flash. Each reviewer read a
  pinned patch file and printed `REVIEWED_HEAD`.

  | Cycle | Reviewed HEAD | Outcome | Findings and fixes |
  |---|---|---|---|
  | 1 | `f3f4a87` | READY / READY / READY_WITH_FOLLOWUPS | 2 × P3, both in scope. Fixed in `34386ad`; each fix was verified by killing its mutation. |
  | 2 | `34386ad` | READY / READY / READY_WITH_FOLLOWUPS | 1 × P3 (a stale comment), in scope. Fixed in `2144414`. |
  | 3 | `2144414` | READY ×3 | Confirmation run with no findings. |

  * Final result: P0 = 0, P1 = 0, outcome `READY`. No finding was out of scope, so no P-021
    deferred entry was captured.
* **Copilot on PR #97: one iteration.** The review returned no findings and opened no
  threads, and the P-018 gate returned `SATISFIED`.
* The PR body recorded `READY` for current HEAD `2144414`, with full-build evidence and no
  follow-ups.
* P-016 topology passed at every phase: pre_claim, post_claim, lifecycle before the build,
  lifecycle before the PR, and `a0`.
* P-009 settings allow merge commits only.
* P-014 approval came from the scoped `DARK_MODE_ACTIVE` activation
  (`merge_approval_pre_authorized: true`).
* The merge used the normal `gh pr merge 97 --merge --delete-branch` path. Admin fallback was
  neither authorized nor used.

## Shipment reconciliation and backlog closure

**Feature completion gate (a1).**

* S0 found no anomalies.
* There was exactly one feature member (`n = 1`).
* All five S4 conditions held:
  * all nine tasks were done;
  * no live descendant remained;
  * the descendant set plus `034-F` equalled the manifest;
  * containment held;
  * `034-F` was live `active`.
* Ship moved `034-F` from `active` to `done` (`backlogit move 034-F --status done`, exit 0)
  and verified the re-read.
* In backlogit 1.11.0, that move also relocated `034-F` from `.backlogit/queue/` to
  `.backlogit/archive/` while it kept `status: done`. This is the valid relocated-done
  intermediate state.

**Pre-mode.** `shipment-reconcile` pre-mode (`expected_status: done`) returned `PROCEED`.

* The record was active and record-consistent.
* Every manifest item was `done`, and all ten members were already in `.backlogit/archive/`.
* There were no orphans. Queue mentions of `031-S` were prose only, plus the `blocks`
  dependency from `039-S`.

**Classification.** Classify-close-path returned `CASCADE` / `FULLY_COVERED_ROOT` with
`CLASSIFICATION_BINDING: d3fe146612514f1fb4e3547c714530936dc3b228f508e0f6b9167542620c95bb`.

* The binding engine line is `backlogit version 1.11.0`.
* The serialization follows the skill's canonical binding exactly: `skill=` is the SHA-256 of the
  LF `SKILL.md` bytes (`1c7d523d…3ff2`), and `deps=` is the sorted shipment `dependencies`
  field (`027-S,029-S,037-S`). The shipment status is `active`. Recomputing with those inputs
  reproduces `d3fe1466…95bb`. A `deps=` line carrying only `037-S` yields a different digest
  (`14d62f24…`). That was the source of a closure-review P1, declined on this evidence.
* **Location came from a filesystem probe, not an assumption.** This applies the D10D3AFC
  lesson from 030-S. The helper checked queue and archive file presence for each member, and
  fails closed if a member is in both roots or in neither.
* All ten members were encoded `done` / `archive`.
* This differs from 030-S, where the covering feature stayed in `queue` after a1. An assumed
  location would have been wrong in one of the two shipments.

**Safe-close.** Safe-close Step 0 recomputed the binding from a fresh probe and reported
`BINDING_MATCH`. It then entered the Cascade Close Sub-Procedure and returned `CLOSED`.

* The cascade primitive was invoked from inside that sub-procedure, with the merge SHA,
  message and author.
* `returned_ids` was `[]`.
* `archived_ids` = `allowed_ids` = `required_ids` = {`031-S`, `034-F`, `034.001-T`,
  `034.002-T`, `034.003-T`, `034.004-T`, `034.007-T`, `034.008-T`, `034.009-T`,
  `034.010-T`, `034.011-T`}.
* Both set differences were empty.
* Every task kept `parent_id: 034-F`.
* `031-S` is archived with `archived_status: shipped`, with `commit` set to the merge SHA.
  `034-F` is archived with `archived_status: done`.

**Post-mode.** Post-mode returned `PROCEED`. All archive files were present, and the P-007
check found no archive deletions.

**Lock.** The `file-lock` skill's bundled script
(`.github/skills/file-lock/scripts/acquire_lock.ps1`) held `.backlogit/queue/031-S.md` from
pre-mode through post-mode. `release_lock.ps1` then released it, and the lock file is gone.
The agent template's root `scripts/acquire_lock.ps1` path is still missing; that conflict
stays Stage-owned under `9F824B64`.

**Reports.**

* `.backlogit/reconcile/031-S-pre-2026-10-02T22-26-15Z.md`
* `.backlogit/reconcile/031-S-safe-close-2026-10-02T22-28-06Z.md`
* `.backlogit/reconcile/031-S-post-2026-10-02T22-28-06Z.md`

The backlog closure was committed on the post-merge branch as `f734fd8`
(`chore(ci): archive 031-S backlog artifacts`), never directly to `main`.

## Source artifact cleanup

`034-F` has neither `custom_fields.source_stash_id` nor
`custom_fields.source_deliberation_id`. Under the manifest-derived cleanup rule, no source
stash entry or deliberation was archived. Stash IDs that appear only in plan or feature prose,
such as `B72E9715` and `8E9F8E55`, were not inferred or touched.

## Pre-deploy audits and rollout path

No data migration, feature flag, access change or production deployment is required.
Rollout is merge-only: the change takes effect in CI and in developer gate runs on `main`.

## Post-merge checks and monitoring

* First `main` CI for merge SHA `c04d9751d0ce98d0e10bf5d498985d2ddf8b73ee` passed
  (run 37072295122).
* Monitor the write-path precondition gate and the gatecheck test jobs on the next two or
  three PRs that touch `tools/gatecheck/`, `cmd/` or `internal/`.

### Healthy signals

* The gatecheck unit tests, the write-path self-tests and the CI gate jobs stay green.
* Ordinary Go changes under `cmd/` and `internal/` do not trip the write-path gate.

### Failure signals

* The gate passes a change that adds a destructive write primitive under `cmd/` or
  `internal/`.
* The gate fails on a change that adds no write primitive (a false positive, for example from
  a string or comment occurrence).

## Risky action record

* **Shipment closure:** the bound P-015 `CASCADE` path was selected only after full-root and
  full-descendant coverage was verified. Its evidence is recorded above: the binding match,
  the two-set gate, empty `returned_ids`, preserved parent links and shipped provenance.
* **Merge:** the normal merge-commit path under the scoped P-017 authorization. No `--admin`
  attempt was made.

## Rollback trigger and procedure

Trigger a rollback if the write-path gate starts failing on changes that add no write
primitive, or if it is shown to accept a write primitive it should reject. Open a merge-commit
revert PR for PR #97 and rerun the gatecheck tests. Do not revert directly on `main`.

## Validation window and owner

Watch the next two or three PRs that exercise `tools/gatecheck/`, `cmd/` or `internal/`.
Owner: Ship and the repository maintainer (`softwaresalt/intercom`).

## Follow-ups (Stage-owned)

* No new follow-up was identified by local review, Copilot or this closure, and no stash
  entry was created.
* Deferred residuals that were captured before this shipment were deliberately not
  implemented: `458F9385`, alias tracking, `WRITE_PATH_GATE_ADVISORY` and D-2′ widening.
* `D10D3AFC` (existing, still active): an invokable `safe-close` CLI or gate. This closure
  avoided the same-helper location blind spot by probing location, but it still relied on a
  session-local helper. No new entry was created.
* `9F824B64` (existing): the root lock-script path conflict.

## Knowledge graduation

No architecture or product-spec change is required.

* **Compound refresh:** refreshed
  `docs/compound/2026-09-30-shipment-reconcile-skill-bypass-manual-substitution-risk.md`
  with the 031-S data point. That data point shows that a covering feature's post-a1 location
  is not stable across shipments, which is why location must be probed.
* No other compound entry covers the write-path gate's masked-text cursor, so none became
  stale.

## Compaction status (P-020)

**`done`.** Ship invoked `compact-context` with `target: all`, and it succeeded without
degradation.

* **Memory:** the 031-S session memory
  (`docs/archive/memory/2026-10-02/031-s-ship-memory.md`) was compacted into
  `docs/memory/compacted/2026-10-02-031-s-034-f-compacted.md`. The original was archived, not
  deleted.
* **Plans:** the 2026-10-01 post-M4 re-plan was skipped as a decided-plan candidate, because
  a later unit (039-S) is still open.
* **Closure records:** none have passed the threshold age.
* **Checkpoints:** no active checkpoint was compacted.

## Closure PR gate status

The closure PR is opened from branch `post-merge/031-s-harden-write-path-gate` after
archival.

* **Topology:** `a0` passed before safe-close. Following the Orchestrator ruling from the
  PR #92, PR #94 and PR #96 precedents, the closure PR's topology evidence is
  `autoharness gate pipeline-topology --mode manual --phase ambient --json`, not the
  lifecycle phase.
* **Review:** the closure diff gets a local three-model review before the PR.
* **Gates:** P-018, required CI and the §1.9 readiness gate apply. The merge uses a normal merge
  commit; admin fallback is not authorized.

## Releasability evidence

**READY** (closure status `READY`, with no conditions). This is internal CI and developer
tooling, with no change to the product runtime or deployment surface.

* The multi-model adversarial review ran three cycles and ended `READY` ×3 at pinned
  `2144414` (P0 = 0, P1 = 0).
* The PR readiness block recorded `2144414` as the current HEAD.
* Copilot review completed with no findings, and P-018 returned `SATISFIED`.
* All PR checks and post-merge CI checks passed.
* P-009 and P-016 passed.
* The merge used the authorized merge-commit path.
* The safe-close binding was computed from probed locations and matched at Step 0.

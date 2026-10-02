---
title: "Post-merge closure: 032-S / 035-F — Repair unignore-regression checker ref resolution"
description: "Post-merge closure artifact for shipment 032-S / feature 035-F"
status: "complete"
tags:
  - "operational-closure"
  - "032-S"
  - "035-F"
  - "post-merge"
date: 2026-10-02
mode: post-merge
shipment: 032-S
feature: 035-F
pr: 93
merge_commit_sha: 532cf525e0b5bd99a6b1ec4ea8335c7d4d911775
compaction_status: done
closure_status: READY
releasability: READY
conditions: []
---

# Post-merge closure: 032-S / 035-F — Repair unignore-regression checker ref resolution

## Summary

Shipment `032-S` delivered feature `035-F`, "Repair unignore-regression checker ref
resolution" (post-M4 plan Unit B / S-8). PR #93 merged to `main` at
`532cf525e0b5bd99a6b1ec4ea8335c7d4d911775` on 2026-10-02.

The merge fixes a latent defence-in-depth defect in the merge-blocking
unignore-regression check's baseline helper:

* **Before:** `rootGitignoreTextAt` treated any `git show <ref>:.gitignore` failure as
  "no `.gitignore`". At the helper level, an unresolvable ref therefore became an
  empty baseline.
  * This was not a live fail-open at the entry point: `runDifferentialCheck`'s
    upstream `git diff <base> <head>` already rejects an ordinary invalid ref.
  * During review, option-like and empty refs (`--format=x`, `""`, `:`) were found
    to reach `git show` and succeed before any validation ran.
* **After:**
  * The ref is validated: an empty ref, or one starting with `-`, is rejected before
    any git call.
  * It is resolved with `rev-parse --verify --quiet <ref>^{tree}` and must be a hex
    object ID.
  * `.gitignore` is read with `git show <treeOID>:.gitignore`.
  * A `show` failure is classified with an exact
    `ls-tree --full-tree -z <treeOID>` listing.
  * `("", nil)` is returned only for a valid ref whose root tree has no `.gitignore`.
    Every other non-`HEAD` resolve, `show` or `ls-tree` failure is a `::error::`
    diagnostic naming the ref.
  * The `HEAD` working-tree read path and text-conversion errors are unchanged.

The change is internal CI and developer gate tooling only. It does not touch
Intercom's product CLI, TUI, WebSocket/API or deployment surfaces, so no production
deployment or data migration is required.

## Invariants to preserve

* An invalid or unresolvable baseline ref is an error. It never degrades to an
  empty baseline.
* A valid ref whose root tree has no `.gitignore` is a legitimate empty baseline,
  not an error (the named stop condition: over-strictness was rejected).
* For a valid ref that has a `.gitignore`, the returned text is unchanged (INV-1).
* Text the caller supplies as a ref never reaches the `git show` or `git ls-tree`
  arguments. Only the resolved tree OID does.

## Validator evidence

| Gate | Result |
|---|---|
| `gofmt -l .` / `goimports -l tools/gatecheck` | Clean |
| `go vet ./...` | Passed |
| `golangci-lint run ./...` | 0 issues |
| `staticcheck ./...` | Passed |
| `go test -race ./tools/gatecheck/internal/unignore/` | Passed |
| `go test -count=1 ./...` | Passed (Git for Windows Bash first on `PATH`) |
| `go build ./...` | Passed (full local build) |
| `check-unignore-regression.sh --self-test` / `--base-ref origin/main` | PASS / PASS |
| PR #93 CI | All 14 checks passed |
| First post-merge `main` CI | Passed; run 37043580641 at the merge SHA |

### Runtime validation applicability

Runtime verification was not run, because this shipment changes only internal
gatecheck tooling. The workspace profile's product runtime probes and manual
checkpoints do not apply. The relevant validator evidence is the CI gitignore/unignore
job (I6), the wrapper self-test and the differential run against `origin/main`.

## Review, CI, and merge evidence

* **Local review: four rounds.** Each round used three models (gpt-5.6-sol as
  anchor, claude-opus-5.5, gemini-3.8-flash), all reviewing the same pinned HEAD. Each
  reviewer printed `REVIEWED_HEAD`.
  * Round 1 (`90ab453`): BLOCKED. Anchor P1: an option-like or empty ref could make
    `git show` succeed before validation. Fixed in `20b5b21`.
  * Round 2 (`20b5b21`): READY, READY_WITH_FOLLOWUPS (4 × P3, all fixed in
    `0540220`) and READY.
  * Round 3 (`0540220`): READY × 3.
  * Round 4 (`d84a3ff`, after the Copilot fix): READY × 3, with one out-of-scope P3
    captured as `F5958BBC`.
  * Final result: P0 = 0, P1 = 0.
* **Copilot on PR #93: two iterations.**
  * Iteration 1 (`0540220`) raised one thread: use the isolated git runner in
    `realExitError`. It was fixed in `d84a3ff`, then replied to citing the SHA and
    resolved.
  * Iteration 2 (`d84a3ff`) had no findings.
  * The P-018 gate returned `SATISFIED`, including at the last-mile re-check.
* The PR body recorded `READY` for current HEAD `d84a3ff`, with full-build evidence and
  follow-ups `B29A565E` and `F5958BBC`.
* P-016 topology passed (pre_claim, post_claim, lifecycle before build and before the
  PR, and `a0`).
* P-009 settings allow merge commits only.
* P-014 approval came from the scoped `DARK_MODE_ACTIVE` activation
  (`merge_approval_pre_authorized: true`).
* The merge used the normal `gh pr merge 93 --merge --delete-branch` path. Admin
  fallback was neither authorized nor used.

## Shipment reconciliation and backlog closure

**Feature completion gate (a1).**
* S0 found no anomalies.
* There was exactly one feature member (`n = 1`).
* All five S4 conditions held:
  * all three tasks were done;
  * no live descendant remained;
  * the descendant set plus `035-F` equalled the manifest;
  * containment held;
  * `035-F` was live `active`.
* Ship moved `035-F` from `active` to `done` and verified the re-read.

**Pre-mode.** `shipment-reconcile` pre-mode (`expected_status: done`) returned
`PROCEED`. The record was active and record-consistent, every manifest item was
pre-archived `done`, and there were no orphans.

**Classification.** Classify-close-path returned `CASCADE` / `FULLY_COVERED_ROOT` with
`CLASSIFICATION_BINDING: 8ea1b40ad0f37fa69e30a90407bc84f8d076bb37fe69fc2f0cf88a64e6d5a96a`.
The binding engine line is `backlogit version 1.11.0`.

**Safe-close.** Safe-close recomputed the binding, found it identical, and entered the
Cascade Close Sub-Procedure. It returned `CLOSED`:

* `returned_ids` was `[]`.
* `archived_ids` = `allowed_ids` = `required_ids` = {`032-S`, `035-F`, `035.001-T`,
  `035.002-T`, `035.003-T`}.
* Both set differences were empty.
* Every task kept `parent_id: 035-F`.
* `032-S` is archived with `archived_status: shipped`, with `commit` set to the merge
  SHA. `035-F` is archived with `archived_status: done`.

**Post-mode.** Post-mode returned `PROCEED`. All archive files were present, and the
P-007 check found no archive deletions.

**Lock.** As in 037-S and 038-S, the repository has no `scripts/acquire_lock.ps1`. Ship
used the degraded manual substitute: an atomic create-new
`.backlogit/queue/.032-S.md.lock`, released after post-mode. Existing stash `9F824B64`
tracks this gap.

**Reports.**
* `.backlogit/reconcile/032-S-pre-2026-10-02T17-06-13Z.md` (intake)
* `.backlogit/reconcile/032-S-pre-2026-10-02T17-53-48Z.md`
* `.backlogit/reconcile/032-S-safe-close-2026-10-02T17-55-12Z.md`
* `.backlogit/reconcile/032-S-post-2026-10-02T17-55-12Z.md`

The backlog closure was committed on the post-merge branch as `30ab027`
(`chore(ci): archive 032-S backlog artifacts`), never directly to `main`.

## Source artifact cleanup

`035-F` has neither `custom_fields.source_stash_id` nor
`custom_fields.source_deliberation_id`. Under the manifest-derived cleanup rule, no
source stash entry or deliberation was archived. The stash IDs mentioned in the
feature description (`2787DA56`, `4C5BEC23`) are prose, not custom fields, and were
not inferred or touched.

## Pre-deploy audits and rollout path

No data migration, feature flag, access change or production deployment is
required. Rollout is merge-only: the change takes effect in CI and in developer gate
runs on `main`.

## Post-merge checks and monitoring

* First `main` CI for merge SHA `532cf525e0b5bd99a6b1ec4ea8335c7d4d911775` passed
  (run 37043580641).
* Monitor the `gitignore append-only + un-ignore regression (I6)` CI job on the next
  two or three PRs that touch `.gitignore` or gatecheck.

### Healthy signals

* The I6 job and the wrapper `--self-test` stay green.
* A valid base ref with no root `.gitignore` still passes, rather than erroring.

### Failure signals

* The unignore unit tests (`TestRootGitignoreTextAt_*`) show `rootGitignoreTextAt`
  returning `("", nil)` for an unresolvable or option-like ref. The CLI cannot
  show this on its own while the upstream `git diff` guard rejects invalid refs.
* A repository whose baseline legitimately lacks a root `.gitignore` starts
  failing I6.

## Risky action record

* **Shipment closure:** the bound P-015 `CASCADE` path was selected only after
  full-root and full-descendant coverage was verified. Its evidence is recorded
  above: the two-set gate, empty `returned_ids`, preserved parent links and shipped
  provenance.
* **Merge:** the normal merge-commit path under the scoped P-017 authorization. No
  `--admin` attempt was made.
* **Locking:** a degraded manual lock-file substitute (`9F824B64`).

## Rollback trigger and procedure

Trigger a rollback if the I6 gate begins failing for a valid baseline that has no
`.gitignore`. Also trigger one if a function-level test shows `rootGitignoreTextAt`
accepting an unresolvable or option-like ref. Open a merge-commit revert PR for
PR #93 and rerun the unignore tests. Do not revert directly on `main`.

## Validation window and owner

Watch the next two or three PRs that exercise `.gitignore` or gatecheck. Owner: Ship
and the repository maintainer (`softwaresalt/intercom`).

## Follow-ups (Stage-owned)

* `B29A565E`: deferred scope expansion (P-021 C2). Harden the `checks.go` `git diff`
  and `runCheck` `rev-parse` against option-like refs.
* `F5958BBC`: deferred scope expansion (P-021 C2). Isolate the existing
  `TestExitCode_ExitError` from user git aliases and `git-*` executables on PATH.
* Both were captured threadless during local review. Their source refs are recorded
  in the entries, and both reach `main` with this closure PR. No other follow-up was
  identified by local review, Copilot or this closure.
* Closure review noted that `B29A565E` cites `unignore.go:321`; the call is at line
  320. The entry was not edited: Ship's P-021 C2 single-write invariant forbids
  amending a captured entry. The line drift is minor, and Stage can correct it at
  triage.

## Knowledge graduation

No architecture or product-spec change is required.

* **New learning:** captured in
  `docs/compound/2026-10-02-git-show-ref-failure-ambiguity-resolve-before-show.md`.
* **Compound refresh:** no existing `docs/compound/` entry mentions
  `rootGitignoreTextAt` or the old degrade-to-empty behaviour, so none became stale.

## Compaction status (P-020)

**`done`.** Ship invoked `compact-context` with `target: all`, and it succeeded
without degradation.

* **Memory:** two completed-unit memory records were compacted. Their originals were
  archived, not deleted, under `docs/archive/memory/2026-10-02/`.
  * `032-S-unit-b-progress.md` became
    `docs/memory/compacted/2026-10-02-032-s-035-f-compacted.md`.
  * `038-S-closure-pr-halt-memory.md` became
    `docs/memory/compacted/2026-10-02-038-s-closure-pr-halt-supplement-compacted.md`.
* **Plans:** the 2026-10-01 post-M4 re-plan was skipped as a decided-plan candidate,
  because later units (030-S, 031-S, 039-S) are still open.
* **Closure records:** none have passed the threshold age.
* **Checkpoints:** no active checkpoint was compacted.

## Closure PR gate status

The closure PR is opened from branch
`post-merge/035-f-repair-unignore-regression-checker-ref-resolution` after archival.

* **Topology:** `a0` passed before safe-close (`BRANCH_POST_MERGE_CLOSURE_ELIGIBLE`).
  Following the Orchestrator ruling from the 038-S / PR #92 precedent, the closure
  PR's topology evidence is
  `autoharness gate pipeline-topology --mode manual --phase ambient --json`, not the
  lifecycle phase.
* **Review:** the closure diff gets a local three-model review before the PR.
* **Gates:** P-018, required CI and the §1.9 readiness gate apply. The merge uses a
  normal merge commit; admin fallback is not authorized.

## Releasability evidence

**READY.** This is internal CI and developer tooling, with no change to the product
runtime or deployment surface.

* Exact-HEAD local review was `READY` (P0 = 0, P1 = 0).
* Copilot review completed and P-018 returned `SATISFIED`.
* All PR checks and post-merge CI checks passed.
* P-009 and P-016 passed.
* The merge used the authorized merge-commit path.

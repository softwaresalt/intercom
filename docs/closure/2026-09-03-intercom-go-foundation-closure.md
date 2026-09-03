---
title: "Operational Closure — intercom-go Foundation (001-S)"
date: 2026-09-03
mode: post-merge
shipment: 001-S
feature: 001-F
pr: 4
merge_commit_sha: d85eaeaf4c9ee36e2fbc9e3b3a407ffc8ea2052a
compaction_status: done
releasability: READY
---

# Operational Closure — intercom-go Foundation (001-S)

## Summary of the Change

Shipment `001-S` — "intercom-go Foundation: Module Skeleton and Hardened CI
Gates" — establishes the first shippable, fork-independent slice of the
agent-intercom Rust-to-Go port:

* Go module skeleton (`go.mod`, module `github.com/softwaresalt/intercom-go`,
  language floor `go 1.22`, toolchain pinned `go1.26.5` for GO-2025-3750).
* Two stub binary entrypoints (`cmd/intercom`, `cmd/intercom-ctl`) with cobra
  root commands, `log/slog` JSON logging to stderr, and a not-implemented
  sentinel — no real operator-facing behavior ships in this slice.
* A canonical CGO-free 4-target cross-compile build script
  (`scripts/build.ps1` + `scripts/targets.json`) with a repo-root escape guard.
* A hardened CI pipeline (`build`/`vet`/`race`/`mod-verify`/`mod-tidy`, `lint`,
  `security` [staticcheck, govulncheck, gitleaks], 4-target `cross-compile`
  matrix, `pipeline-topology` ambient check, `ci gate` aggregation).

Config schema, credential resolution, and the `AppError` taxonomy are
deliberately out of scope (deferred to stash `037B1552`).

PR: https://github.com/softwaresalt/intercom/pull/4
Merge commit: `d85eaeaf4c9ee36e2fbc9e3b3a407ffc8ea2052a` (merge-commit strategy,
per P-009).

## CI Status and Review Items

* Hosted CI (final run `33810635425`, post-revert HEAD `08d0c4a`): green end
  to end — `ci gate`, `detect code changes`, `pipeline-topology (ambient)`,
  `test`, `lint`, `security`, `load cross-compile targets`, 4/4
  `cross-compile` legs.
* Local review readiness: `READY_WITH_FOLLOWUPS`, `P0=0, P1=0`, both at the
  original reviewed HEAD (`8907367`) and re-affirmed by tree-identity at the
  final merged HEAD (`08d0c4a`) after a mechanical revert (see "Pre-Merge
  Remediation" below).
* P-018 copilot-review gate: `NOT_APPLICABLE` at both the pre-revert and
  final HEAD (no engagement signal; enforcement `auto`).
* Unresolved review threads: none (0 PR review comments, 0 formal reviews
  requested — 4 persona reviewers ran in report-only local mode instead).
* Follow-ups: 7 out-of-scope findings captured to stash under P-021 C2
  (`EFFAA358`, `90350C9A`, `F7C6420D`, `BEDD2E70`, `35D76D5E`, `EF9352FB`,
  `90EE7758`) — none of them blocking; all are documentation/hardening
  hygiene items for a future shipment.

### Pre-Merge Remediation (this session)

On resumption, PR #4's `headRefOid` had advanced past the recorded reviewed
HEAD (`8907367` → `0e82e2b`, commit "Configs"). Investigation found this
stray commit introduced four machine-local artifacts never in scope for
`001-S`: `.mcp.json`, `intercom-go.code-workspace`,
`.backlogit/hooks_queue.jsonl`, and a `.gitignore` edit distinct from the
one intentionally parked in `stash@{0}` for this session. None contained
secret values, but none belonged in this PR.

Remediation: `git revert --no-edit 0e82e2b` → commit `08d0c4a`. Verified
mechanically that `HEAD^{tree}` for `08d0c4a` is byte-identical
(`b8767344b8c73e4da2e8b77633d6eb20c121ba7b`) to the tree already covered by
the `READY_WITH_FOLLOWUPS` local review — so the review's findings remain
fully valid with zero new surface introduced. Re-ran full local
build/vet/test and confirmed hosted CI green on the new HEAD before
re-checking mergeability and merging.

## Runtime Verification / Validator Evidence

No `runtime-verification` invocation was required for this shipment. Both
`cmd/intercom` and `cmd/intercom-ctl` are stub entrypoints that parse flags,
configure logging, and print a not-implemented sentinel — there is no
operator-facing runtime behavior, external integration, data path, or
migration in this slice for a validator to probe. The change's operational
surface is the **CI pipeline itself** (build/lint/security/cross-compile
gates), which is exercised directly by the hosted CI evidence above rather
than a separate runtime validator.

## Pre-Deploy Audits

* No config, flags, credentials, or migrations are introduced by this slice
  (deliberately deferred — see Summary above).
* `go.mod` toolchain pin (`go1.26.5`) verified compatible with the CI
  `setup-go` matrix (all four `go-version` values bumped to `1.26.x` in
  commit `d9add30`, remediating `GO-2025-3750`).
* Cross-compile artifacts (`scripts/build.ps1 -OutputDir dist`) verified
  8/8 produced locally with `CGO_ENABLED=0` confirmed via `go version -m`.

## Deployment / Rollout Path

Merge-only. There is no deployable artifact or release process for this
slice — the built binaries are not-implemented stubs, not a shippable
product surface. No canary, phased rollout, or maintenance window applies.

## Post-Deploy Checks

* Confirm `main`'s CI pipeline stays green on the next PR that targets this
  module (first real signal that the hardened gate set holds under new
  content).
* Confirm `govulncheck`/toolchain-pin maintenance is picked up as documented
  in follow-up `90EE7758` (deferred, non-blocking).

## Healthy Signals

* CI gate aggregation (`ci gate` check) stays green on subsequent PRs.
* `go.mod`/`go.sum` remain stable (no unexpected drift) until the next
  intentional dependency change.
* Cross-compile matrix continues producing 4/4 artifacts.

## Failure Signals

* Any hosted CI run regressing the `security` job (staticcheck / govulncheck
  / gitleaks) without an explicit, reviewed cause.
* `go.mod` toolchain pin falling behind a disclosed Go CVE (tracked via
  follow-up `90EE7758`).

## Monitoring Plan

* GitHub Actions run history for `.github/workflows/ci.yml` on `main`.
* No dashboards/alerts apply — this is a foundation/scaffolding shipment
  with no running service.

## Rollback Trigger / Procedure

* Trigger: a subsequent PR reveals the module skeleton, build script, or CI
  gate set itself is structurally broken (e.g. cross-compile matrix
  silently stops verifying a target).
* Procedure: `git revert` the merge commit `d85eaeaf4c9ee36e2fbc9e3b3a407ffc8ea2052a`
  on `main` via a standard PR (no direct-to-main mutation), since no
  deployed artifact exists to roll back independently of source control.

## Validation Window

One subsequent shipment cycle (first real feature PR built on this
foundation) — the CI pipeline is considered validated once it gates a
non-scaffolding change successfully.

## Owner

Repository maintainer (`softwaresalt`) — no separate on-call/runtime owner
applies; this is a build/CI foundation, not a running service.

## Risky Action Record

None. No destructive, high-blast-radius, or irreversible actions were taken
during this shipment's implementation or closure. The stray-commit
remediation (revert `0e82e2b`) was a low-risk, tree-verified mechanical
correction, not a risky action requiring pre-approval.

## Source Artifact Cleanup

* `001-F.custom_fields.source_stash_id`: not present — no stash entry to
  archive.
* `001-F.custom_fields.source_deliberation_id`: not present — no
  deliberation artifact to archive. (The plan/deliberation docs referenced
  by `001-F.references` are durable design documents, not backlogit
  deliberation artifacts, and are retained.)
* Archived: 0 stash, 0 deliberations.

## Shipment Closure Evidence

* Reconciliation: `mode: pre` (`expected_status: done`) — all 11 non-shipment
  manifest items classified `pre-archived` (already archived prior to this
  session's resumption); shipment record `active` → `record-consistent`; no
  orphans.
* Close path: **P-015 verified fully-covered-root cascade** — `001-F` is a
  root feature (no `parent_id`) and its full descendant set at every depth
  (`001.001-T`, `001.002-T`, and all 8 subtasks) is exactly the shipment
  manifest; no other descendants exist. Verified via
  `backlogit shipment ship 001-S --sha d85eaeaf4c9ee36e2fbc9e3b3a407ffc8ea2052a`:
  `returned_ids: []`; `archived_ids` (12: shipment + feature + 2 tasks + 8
  subtasks) exactly matches the computed `allowed_ids`/`required_ids` sets;
  every `parent_id` preserved against the pre-close snapshot; `git status`
  showed only modifications/rename under `.backlogit/`, no unexpected
  deletions.
* Shipment record: `status: shipped`, `archived_status: shipped`,
  `commit_sha: d85eaeaf4c9ee36e2fbc9e3b3a407ffc8ea2052a`.
* Post-mode: all 12 archive files present, no deletions detected —
  `recommendation: PROCEED`.
* Backlog archival committed on `post-merge/001-f-intercom-go-foundation`
  (commit `7134523`), not on `main` (P-020/branch-per-release-unit).

## Compaction Status (P-020)

`done`. `compact-context` (`target: memory`) ran in this closure session: the
two session memory files for this release unit
(`docs/memory/2026-09-03-stage-intercom-go-foundation.md`,
`docs/memory/2026-09-03-ship-intercom-go-foundation.md`) were consolidated
into
`docs/memory/compacted/2026-09-03-001-s-intercom-go-foundation-compacted.md`
and the verbose originals moved to `docs/archive/memory/`. No other
`docs/memory/` files exceeded the threshold-gated candidate criteria (file
count and total size are both well under the `max_files: 40` /
`max_size_kb: 500` defaults), so this was a bounded, per-release-unit Tier-1
consolidation, not a broad sweep.

## Releasability Evidence

**READY.** No blocking findings, no runtime risk, no deployment/rollback
complexity beyond a standard `git revert` of the merge commit. Follow-ups are
all non-blocking, deferred, stash-captured items for a future shipment.

---
title: "Post-merge closure: 012-S -- retire the D6a gate narrowing (PR #36)"
description: "Operational closure artifact for shipment 012-S / feature 013-F"
status: "complete"
tags:
  - "closure"
  - "012-S"
  - "013-F"
  - "post-merge"
date: 2026-09-08
mode: post-merge
shipment: 012-S
feature: 013-F
pr: 36
merge_commit_sha: 9d8937ae2bc721cdeb15b74963acb92221843f06
compaction_status: done
closure_status: READY
releasability: READY
---

# Post-Merge Operational Closure — Shipment 012-S (Feature 013-F)

- Mode: `post-merge`
- Context: PR [#36](https://github.com/softwaresalt/intercom/pull/36), merge commit `9d8937ae2bc721cdeb15b74963acb92221843f06`
- Feature: `013-F` — Retire the D6a gate narrowing: broaden U-E1a to `internal/**`
- Tasks: `013.001-T` .. `013.006-T` (all `done`, archived)
- Shipment: `012-S` (`status: archived`, `archived_status: shipped`)

## Summary of Change

Retired the vestigial D6a narrowing of the U-E1a retired-architecture CI gate script
(`scripts/check-retired-architecture.sh`). Broadened the enforced scan scope from
`internal/config/**` to `internal/**`, added structural self-test assertions pinning
that scope against silent re-narrowing, fixed a P0 dead-dispatch bug where
`config.toml.example` was never routed to the TOML scan engine, added a Go fixture
suite alongside the existing TOML suite via a shared `engine_for_path()` dispatch
helper, and amended governing decision D6 with dated amendment D6c. Two governance
docs were annotated (not rewritten) to reflect the historical narrower scope.

## CI Status and Unresolved Review Items

- All required CI checks passed at merge (`ci gate`, `lint`, `test`,
  `test (windows, advisory)`, `security` [gitleaks], 4x `cross-compile`,
  `pipeline-topology (ambient)`, `gitignore append-only + un-ignore regression (I6)`,
  `detect code changes`, `load cross-compile targets`).
- Copilot review: `SATISFIED` at merge HEAD `c403e5e`, 0 unresolved threads
  (1 round-trip: 1 actionable comment on the `--self-test` usage docstring, fixed
  in-scope, replied, and resolved via `resolveReviewThread`).
- Local adversarial review (3 reviewers, report-only):
  1 HIGH consensus finding + 1 MEDIUM majority finding remediated in-scope
  pre-PR; 1 MEDIUM finding independently verified accurate (commit-hash claim);
  remaining LOW/pre-existing findings captured as deferred-scope-expansion stash
  entries (`8988120D`, `9A14D3B7`, `E403E3C5`, `09CD6ACF`, plus reuse of `F4F4A959`).
  Full report: `docs/closure/2026-09-08-012-s-retire-d6a-gate-narrowing-adversarial-review.md`.
- No unresolved review items remain open against this PR.

## Runtime Verification

**Not applicable.** This shipment changes a CI gate script
(`scripts/check-retired-architecture.sh`), its test fixtures, and governance
documentation only. It does not touch the CLI or API runtime surfaces described in
`.autoharness/workspace-profile.yaml` (no changes under `cmd/`, `internal/config`,
`internal/copilotprobe`, or any runtime-facing package). `runtime_validation` from
the workspace profile (CLI smoke, WebSocket upgrade/keepalive, manual checkpoints)
is out of scope for this change; no validator evidence was produced or required.

## Risky Actions

None. No destructive data actions, no backfill, no irreversible steps. The gate
script broadening was verified against the live tracked tree before merge
(`scripts/check-retired-architecture.sh` exits 0 with zero findings across the
broadened `internal/**` scope).

## Affected Runtime Surfaces

None (CI tooling + documentation only).

## Deployment / Rollout Path

Merge-only. No deploy, canary, or phased rollout applies to a CI gate script
change. The gate takes effect immediately for the next CI run on `main` and any
subsequent PR.

## Invariants Preserved

- `internal/**/*_test.go` and `internal/**/testdata/**` remain accepted-unenforced
  surface (excluded from the gate scan), preserving invariant I-B from the original
  gate design.
- The gate's threat model is unchanged: D6b's anti-accident-only characterization
  still applies; broadening scan coverage is a hygiene precondition for C3, not an
  authorization or anti-adversary control, and does not discharge H5 (recorded
  explicitly in amendment D6c).
- `--self-test` continues to exit 0 only when all three check classes (fixture
  suites, structural selection assertions, repo scan) pass.

## Post-Deploy / Post-Merge Checks

- Confirmed `bash scripts/check-retired-architecture.sh --self-test` passes at
  merge HEAD (already run pre-merge and re-verified in this closure pass).
- Confirmed the gate's `lint` job step is wired into `ci.yml` and ran successfully
  on the merge commit.

## Healthy Signals

- `scripts/check-retired-architecture.sh` (no args) exits 0 on `main` post-merge.
- `scripts/check-retired-architecture.sh --self-test` exits 0 on `main` post-merge.
- No new CI failures on subsequent PRs attributable to the broadened scope.

## Failure Signals

- A future PR's CI fails the `lint` job's retired-architecture gate step or
  `--self-test` step in a way traceable to this broadening (e.g. a legitimate
  `internal/**` file outside `internal/config/` now flagged unexpectedly) would
  indicate a false-positive introduced by this change.
- Note (deferred, non-blocking): stash `F4F4A959` records that this gate step
  currently runs with `continue-on-error: true` in `ci.yml`'s `lint` job, so the
  bare repo-scan step alone cannot fail CI — real enforcement currently depends on
  the blocking `--self-test` step (which internally re-invokes the repo scan).
  This is pre-existing wiring, unchanged by this shipment.

## Monitoring Plan

- Standard CI monitoring: watch the `lint` job and `--self-test` step results on
  subsequent PRs touching `internal/**`.
- No dashboards, alerts, or additional monitoring infrastructure required for a
  CI-tooling-only change.

## Rollback Trigger

A confirmed false-positive gate failure on `main` or an open PR, traceable to the
`internal/config/** -> internal/**` scope broadening, blocking otherwise-valid work.

## Rollback Procedure

Revert the merge commit (`9d8937ae2bc721cdeb15b74963acb92221843f06`) via a
standard `git revert` PR, or narrow `should_scan_repo_path()`'s `internal/`
prefix check back to a specific subpath while preserving the task-001/005
self-test hardening. No data migration or state rollback is required — this is a
pure code/CI-config change.

## Validation Window

7 days of normal CI activity on `main` (subsequent PRs touching `internal/**`)
is sufficient to observe whether the broadened scope produces any false positive.

## Owner

Repository maintainer (`softwaresalt`) — the sole committer/reviewer context for
this dark-mode autonomous session.

## Releasability Evidence

- Status: **READY**
- Required evidence: not applicable (no runtime surfaces changed; `runtime_validation.releasability` from the workspace profile applies to CLI/API surfaces not touched here).
- CI: all required checks green at merge.
- Review: local adversarial review + Copilot review both clean at merge HEAD.
- Follow-ups: 4 new deferred-scope-expansion stash entries + 1 reused, all
  provisional-priority, none blocking.

## Compaction Status (P-020)

`done` — `compact-context` invoked at post-merge closure, consolidated this
release unit's memory (Stage's queuing session, task 001's RED evidence, and
Ship's session summary) into
`docs/memory/compacted/2026-09-08-012-s-013-f-compacted.md`, with verbose
originals moved to `docs/archive/memory/`. Overall `docs/memory/` was well
under both compaction thresholds (23 files / 92.7 KB vs. 40 files / 500 KB),
so this was the intended single bounded Tier-1 consolidation of the
just-closed release unit's memory, not a broader sweep.

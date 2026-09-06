---
title: "Post-merge closure: 009-S -- pathsafe and config containment hardening (PR #27)"
description: "Operational closure artifact for shipment 009-S / feature 010-F"
status: "complete"
tags:
  - "closure"
  - "009-S"
  - "010-F"
  - "post-merge"
date: 2026-09-06
mode: post-merge
shipment: 009-S
feature: 010-F
pr: 27
merge_commit_sha: d8ca53f59871158d007acdd86188b2c0b47c2bc8
compaction_status: done
closure_status: READY
releasability: READY
---

# 2026-09-06 — 009-S / 010-F pathsafe and config containment hardening — closure

- shipment: 009-S
- feature: 010-F
- tasks: 010.001-T, 010.002-T, 010.003-T, 010.004-T, 010.005-T
- PR: [#27](https://github.com/softwaresalt/intercom/pull/27)
- merge commit: `d8ca53f59871158d007acdd86188b2c0b47c2bc8`
- reviewed HEAD: `6476ab1430ba0d41d685d83ff78361c55143d624`
- mode: post-merge
- compaction status (P-020): done (see Compaction section below)

## Summary of change

Hardening pass over `internal/pathsafe` and `internal/config`: `NewRoot`'s
two failure branches (`filepath.Abs`, `filepath.EvalSymlinks`) now preserve
the underlying cause via `apperr.Wrapf` (U1) while retaining
dual-discriminability against `apperr.ErrPathViolation`; `NewRoot` rejects
non-directory roots via an added `os.Stat` check (U2); the filesystem-root
doubled-separator containment bug in `hasPathPrefix` is fixed (U3); the
dangling-symlink final-component accept behavior is pinned as a permanent
regression test (U4, GO-14); and `internal/config/validate.go`'s
`containsDotDotSegment` was relocated as a pure, zero-verdict-change move to
the new exported `pathsafe.ContainsDotDotSegment` (U5/010.005-T), consumed by
rule 7. All five 010-F task ACs pass; the original normalize-based
unification design (which would have regressed containment behavior) was
rejected during adversarial review and replaced with the narrower pure-move
approach actually shipped.

## CI status and unresolved review items

* CI: all 12 GitHub Actions checks passed on PR #27 (cross-compile ×4, lint,
  security, test, detect-code-changes, gitignore append-only,
  pipeline-topology (ambient), ci gate, load cross-compile targets) at
  reviewed/merged HEAD `6476ab1430ba0d41d685d83ff78361c55143d624`.
* Local review (Step 4.4): 5-persona report-only review. Outcome:
  `READY_WITH_FOLLOWUPS`. Blocking findings: `P0=0, P1=0`.
* P-018 Copilot review gate: `NOT_APPLICABLE` (no engagement signal, mode
  `auto`) — not held, re-verified immediately before merge and again at the
  unconditional last-mile re-check.
* P-009 merge-strategy guardrail: repository allows merge-commit strategy
  (`allow_merge_commit: true`); PR #27 merged via `gh pr merge --merge`
  (merge commit, not squash/rebase).
* Unresolved review items: none blocking. Five P-021 deferred
  scope-expansion follow-ups were captured (see Follow-ups below);
  all threadless (local review, pre-PR), all cited by ID in this record per
  P-021 C3.

## Runtime verification

**Not applicable.** This change does not touch any runtime surface
(`cli`/`public_api`/`web_ui`/`background_jobs` per `workspace-profile.yaml`
`runtime_surfaces`). The full changed-file set is `internal/config/validate.go`,
`internal/pathsafe/{contain_test.go,lexical.go,lexical_test.go,pathsafe.go,
root.go,root_test.go,symlink_test.go}` — internal path-safety/config-validation
library code with no `cmd/` entrypoint, CLI command handler, or WebSocket/API
server file touched. No validator evidence was produced because no runtime
probe applies to this diff.

## Invariants to preserve

* `errors.Is(err, os.ErrNotExist)` / `errors.Is(err, os.ErrPermission)` both
  continue to work through `NewRoot`'s wrapped errors (U1 AC-1).
* `errors.Is(err, apperr.ErrPathViolation)` still matches for both `NewRoot`
  failure branches — dual discriminability, neither regresses (U1 AC-2).
* `hasPathPrefix`'s filesystem-root doubled-separator fix does not change any
  other containment verdict (U3, regression-tested).
* The dangling-symlink final-component accept behavior is now permanently
  pinned by `TestResolveAllowsDanglingSymlinkAtFinalComponent` (U4, GO-14) —
  any future change altering this behavior will fail CI, not silently drift.
* `pathsafe.ContainsDotDotSegment` is a byte-for-byte behavior-preserving move
  of the prior `internal/config` logic — zero verdict change (010.005-T AC-4),
  confirmed by all three pre-existing `internal/config/paths_test.go` rule-7
  tests passing unmodified.

## Pre-deploy audits

* No migrations, flags, config schema, or access changes — pure Go source
  change plus new/updated regression tests.
* No rollout prerequisites; `internal/pathsafe` and `internal/config` ship as
  part of whatever binary imports them (no independent deployment path).

## Deployment / rollout path

Merge-only. No deploy/canary/phased rollout applicable — this repository has
no independent runtime deployment pipeline for these internal packages
outside of whatever downstream binary/service consumes them.

## Post-deploy checks

* `go build ./...` — success (post-merge, on `main`-derived closure branch).
* `go vet ./...` — 0 issues (post-merge).
* `go test ./...` — all packages pass, including
  `internal/config` (2.577s) and `tests/integration` (49.655s), post-merge.
* Full local build/test evidence was also captured pre-merge in the PR #27
  local review readiness block (`go build ./... && go test ./...` —
  successful, all packages pass).

## Risky action record

None beyond ordinary shipment closure. No destructive, high-blast-radius, or
approval-gated action was taken during implementation. The one notable design
correction (rejecting the normalize-based unification approach after
adversarial review flagged a containment regression risk) occurred entirely
pre-implementation/pre-merge and is recorded in the PR history and in deferred
entry `798002CB`.

## Healthy signals

* Full repo build/vet/test green, confirmed both pre-merge (CI, 12/12 checks)
  and post-merge (local, this session).
* No new lint/security findings introduced (CI `lint`/`security` jobs green).
* Regression tests added for every hardened behavior (U1–U4, plus the pure
  move in U5) — the invariant list above is now mechanically enforced.

## Failure signals

* Any future change causing `TestResolveAllowsDanglingSymlinkAtFinalComponent`,
  `TestNewRootOnMissingDirReturnsPathViolation`, or the `internal/config`
  rule-7 tests to fail would indicate an undetected containment regression.
* Loss of dual `errors.Is` discriminability (`os.ErrNotExist`/`os.ErrPermission`
  vs `apperr.ErrPathViolation`) on `NewRoot`'s wrapped errors.

## Monitoring plan

No dedicated runtime monitoring required — this is a compile-time and
test-verified internal library change with no live traffic/runtime signal to
observe. Standard CI on `main` going forward is the ongoing safety net; the
new regression tests (U1–U4) act as permanent drift guards for the exact
behaviors this shipment hardened.

## Rollback trigger

A regression discovered in any downstream consumer of `internal/pathsafe` or
`internal/config` (compile failure, containment-verdict change, or any of the
new/updated regression tests failing) traceable to this change.

## Rollback procedure

`git revert d8ca53f59871158d007acdd86188b2c0b47c2bc8` on `main` (single merge
commit, cleanly revertible — no follow-on commits depend on it yet beyond
this post-merge closure branch, which only touches backlog/docs artifacts).

## Validation window

Standard: next `main` CI run plus normal PR-review cadence of any future
consumer touching `internal/pathsafe`/`internal/config`. No time-boxed
observation window needed — no live/runtime surface to observe.

## Owner

Ship agent (this session) / repository maintainer for any post-merge issue
triage.

## Follow-ups

Five P-021 deferred scope-expansion entries captured during the 009-S local
review gate, all threadless (pre-PR local review, no review thread existed),
all cited here per P-021 C3 (source refs: task=N/A or specific task,
feature=010-F, shipment=009-S, PR=N/A at capture time, thread=N/A):

* `BF5DE670` (medium, requires deliberation) — consolidate tracking for the
  residual write-through-dangling-symlink primitive (GO-14), the
  Resolve-to-filesystem-use TOCTOU window, and the EvalSymlinks-does-not-
  resolve-hardlinks gap (SEC-5) before pathsafe gates any live destructive
  file-write path.
* `E89E5D00` (medium, requires deliberation) — pathsafe's Windows-specific
  security branches (SEC-1/SEC-2/GO-7) are never exercised by `pull_request`
  CI (ubuntu-only runner); needs a cross-cutting infra decision on adding a
  Windows CI job.
* `798002CB` (low, no deliberation needed) — `pathsafe.ContainsDotDotSegment`
  is exported from the same namespace as the security-critical
  Root/Resolve containment API but is materially weaker; risks a future
  caller mistaking it for full containment.
* `8D953C4B` (medium, requires deliberation) — `internal/config` rule 7
  permits an absolute `database.path` with no `pathsafe.NewRoot`/`Resolve`
  containment check (pre-existing "divergence V8", not introduced by
  010.005-T), textually in tension with constitution Principle III without a
  recorded exception.
* `8472E0A1` (low, no deliberation needed) — consolidated cluster of 8 minor
  code-quality/test-coverage follow-ups (helper extraction, additional
  dual-discriminability test, config-layer regression test, long-path
  concern, ancestor-walk regression test, GOOS-gating consistency, doc note
  on env-var expansion, package-organization question).

All five are recorded in `.backlogit/stash.jsonl` and committed on this
closure branch. None require action within 009-S's own scope; each names its
P-021 C1 out-of-scope rationale in the stash entry itself.

## Source artifact cleanup

Checked `010-F.custom_fields` for `source_stash_id` / `source_deliberation_id`
— **neither present** (only `harness_status`). No stash or deliberation
artifact to archive for this shipment's covering-feature scope. This is
expected, not an error: 010-F's description references stash IDs `4A91CA81`,
`3D61B5A8`, `7028FBE7`, `7774C9CA` in free text, but the structured
`custom_fields` linkage was never populated, so there is nothing for this
automated cleanup step to act on. (The five new P-021 entries captured
during this shipment's own review gate are follow-up captures, not source
artifacts being retired, and are intentionally left active in the stash for
Stage triage.)

## Backlog closure

Closed via the P-015 **verified fully-covered-root exception**: `010-F` is a
root feature (no `parent_id`) fully covered by its 5 manifest tasks (no
non-manifest descendants, no linked deliberations), so `backlogit shipment
ship` was the machine-verified permitted close path rather than manual
safe-close (which is also no longer available as a non-cascading path in the
installed `backlogit` 1.10.1 CLI for shipment records). Verification
(`returned_ids` empty; `archived_ids` exactly matched the allowed/required
two-set gate; all `parent_id` values preserved) recorded in
`.backlogit/reconcile/009-S-safe-close-20260906-214047.md`.

Archived: `009-S` (status `archived`, `archived_status: shipped`), `010-F`,
`010.001-T`..`010.005-T` (all `archived` / `archived_status: done`).

Backlog artifacts committed on `post-merge/009-s-pathsafe-and-config-
containment-hardening` at commit `a155dd8` ("chore: archive 009-S backlog
artifacts").

## Compaction (P-020)

Invoked `compact-context` with `target: all` (mandatory per merge). The
just-closed release unit's own two session memory checkpoints
(`2026-09-06-193000-ship-009-s-pre-pr-checkpoint.md`,
`2026-09-06-193500-ship-009-s-merge-approval-halt.md`) qualified as
candidates under the Phase 2 completed-work rule (all 009-S tasks/feature
now `done`/archived). No `docs/plans/` entries or `docs/closure/` records
were old enough (`threshold_days` default 14) or otherwise eligible for
compaction this cycle. Result:

* Files compacted: 2 (memory)
* Compacted summary: `docs/memory/compacted/2026-09-06-009-s-compacted.md`
* Verbose originals moved to: `docs/archive/memory/`
* Plans consolidated: 0 (none eligible)
* Closure records compacted: 0 (none eligible — this shipment's own closure
  artifact is freshly created, not a compaction candidate)
* `compaction_status`: `done`

## Releasability evidence

* **Status**: `READY`
* Required evidence per `runtime_validation.releasability` (healthy-signal:
  CLI/TUI smoke, healthy-signal: WebSocket connect/hydrate/tool-approval) —
  **not applicable** to this change; no runtime surface touched. Marking
  `READY` rather than `READY_WITH_CONDITIONS` because the releasability
  policy's required evidence is scoped to changes that touch `cli`/`api`
  surfaces, and this diff touches neither.

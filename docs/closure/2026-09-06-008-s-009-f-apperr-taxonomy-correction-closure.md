# 2026-09-06 — 008-S / 009-F apperr taxonomy correction — closure

- shipment: 008-S
- feature: 009-F
- tasks: 009.001-T, 009.002-T, 009.003-T, 009.004-T, 009.005-T, 009.006-T
- PR: [#24](https://github.com/softwaresalt/intercom/pull/24)
- merge commit: `1334c5b14958a4755ee590d24f580b4b43701dec`
- reviewed HEAD: `9acb40b8b8cb04a0199eec59b68b2b2f13d31d2a`
- mode: post-merge
- compaction status (P-020): done

## Summary of change

Pure internal correction of the `internal/apperr` taxonomy: retired 3 dead
Kinds (`KindSlack`/`KindIPC`/`KindACP`, 14→11), added an exported
`Kind.String()` (`fmt.Stringer`), added a taxonomy drift-guard test tying
`Kind` count to sentinel count, added a cause-preserving `Wrapf` constructor
(consumed by the upcoming `010-F`), and registered `Newf`/`Wrapf` with
golangci-lint's govet printf analyzer. No public API surface removed outside
the already-dead Kinds (zero external callers verified pre-implementation);
`Error()` output is byte-identical pre/post change (golden test
`TestErrorRenderingUnchangedByStringer`).

## CI status and unresolved review items

* CI: all GitHub Actions checks passed (cross-compile ×4, lint, security,
  test, detect-code-changes, gitignore append-only, pipeline-topology, ci
  gate) — see PR #24 checks.
* Local review (Step 4.4): dispatched independent `code-review` sub-agent,
  verdict **READY**, 0 P0/P1/P2/P3 findings (one trivial doc-comment wording
  nuance noted, explicitly not requiring follow-up).
* P-018 Copilot review gate: `NOT_APPLICABLE` (no engagement signal, mode
  `auto`) — not held.
* No unresolved review items.

## Runtime verification

**Not applicable.** This change does not touch any runtime surface
(`cli`/`api`/`web_ui`/`background_jobs` per `workspace-profile.yaml`
`runtime_surfaces`). `internal/apperr` is an internal error-taxonomy package;
the change set adds/removes internal error constructors and a `Stringer`
method with byte-identical `Error()` output (golden-tested) and zero
behavior change to any CLI, WebSocket, or Dev Tunnel surface. No validator
evidence was produced because no runtime probe applies to this diff.

## Invariants to preserve

* `Error()` string output for every retained `Kind` is unchanged
  (golden-tested: `TestErrorRenderingUnchangedByStringer`).
* `allKinds`/sentinel counts stay in lockstep going forward
  (`TestTaxonomyKindsAndSentinelsStayInSync` — drift guard, U3).
* `Kind` has no persistence/serialization surface (verified pre-implementation
  via `git grep`; U1b AC-5 precondition) — removing 3 Kind values could not
  break any serialized/stored representation.

## Pre-deploy audits

* No migrations, flags, config, or access changes — pure Go source + lint
  config change.
* No rollout prerequisites; this is a library-level change with no
  independent deployment path (it ships as part of whatever binary imports
  `internal/apperr`).

## Deployment / rollout path

Merge-only. No deploy/canary/phased rollout applicable — this repository has
no independent runtime deployment pipeline for `internal/apperr` outside of
whatever downstream binary/service consumes it.

## Post-deploy checks

* `go build ./...`, `go vet ./...`, `go test ./...` (all packages, not just
  `internal/apperr`) confirmed no downstream compile/test breakage from the
  taxonomy trim, prior to merge.
* `go test -race ./internal/apperr/...` confirmed no data races introduced.
* `golangci-lint run ./...` confirmed 0 issues, including verified printf
  registration effectiveness for `Newf`/`Wrapf` (proved via a temporary probe
  file with deliberate format-arg mismatches, deleted before commit).

## Risky action record

None. No destructive, high-blast-radius, or approval-gated action was taken.
Removal of the 3 retired Kinds was preceded by a `git grep` blast-radius
verification confirming zero external callers outside `internal/apperr`
before any code was deleted.

## Healthy signals

* Full repo build/vet/test/lint green (confirmed pre-merge and via CI
  post-push).
* No new golangci-lint findings anywhere in the repo.

## Failure signals

* Any compile failure in a package importing `internal/apperr` (none
  observed; would indicate an undiscovered caller of a retired Kind/sentinel).
* Any diff in `Error()` string output for a retained Kind (guarded by the
  golden test going forward).

## Monitoring plan

No dedicated monitoring required — this is a compile-time-verified internal
library change with no live traffic/runtime signal to observe. Standard CI on
`main` going forward is the ongoing safety net (the new drift-guard test
`TestTaxonomyKindsAndSentinelsStayInSync` will fail CI if the taxonomy drifts
out of sync again).

## Rollback trigger

A regression discovered in any downstream consumer of `internal/apperr`
(compile failure, behavior change in error message text, or drift-guard test
failure) traceable to this change.

## Rollback procedure

`git revert 1334c5b14958a4755ee590d24f580b4b43701dec` on `main` (single
merge commit, cleanly revertible — no follow-on commits depend on it yet).

## Validation window

Standard: next `main` CI run plus normal PR-review cadence of any consumer
touching `internal/apperr` (e.g. the upcoming `010-F` which consumes `Wrapf`).
No time-boxed observation window needed — no live/runtime surface to observe.

## Owner

Ship agent (this session) / repository maintainer for any post-merge issue
triage.

## Follow-ups

None identified beyond the already-scoped, not-yet-started `010-F` (which
consumes the new `Wrapf` constructor) — that is future planned work already
tracked outside this shipment, not a new follow-up generated by this
closure.

## Source artifact cleanup

Checked `009-F.custom_fields` for `source_stash_id` / `source_deliberation_id`
— **neither present** (only `harness_status`). No stash or deliberation
artifact to archive for this shipment's scope. This is expected, not an
error: 009-F's description references stash IDs `8C2D578D`, `9A4C8749`,
`7774C9CA` in free text, but the structured `custom_fields` linkage was never
populated, so there is nothing for this automated cleanup step to act on.

## Backlog closure

Closed via the P-015 **verified fully-covered-root exception**: `009-F` is a
root feature (no `parent_id`) fully covered by its 6 manifest tasks (no
non-manifest descendants), so `backlogit shipment ship` was the machine-
verified permitted close path rather than manual safe-close. Verification
(returned_ids empty; archived_ids exactly matched the allowed/required
two-set gate; all `parent_id` values preserved) recorded in
`.backlogit/reconcile/008-S-safe-close-20260906-024344.md`.

Archived: `008-S` (status `archived`, `archived_status: shipped`), `009-F`,
`009.001-T`..`009.006-T` (all `archived` / `archived_status: done`).

## Releasability evidence

* **Status**: `READY`
* Required evidence per `runtime_validation.releasability` (healthy-signal:
  CLI/TUI smoke, healthy-signal: WebSocket connect/hydrate/tool-approval) —
  **not applicable** to this change; no runtime surface touched. Marking
  `READY` rather than `READY_WITH_CONDITIONS` because the releasability
  policy's required evidence is scoped to changes that touch `cli`/`api`
  surfaces, and this diff touches neither.

---
title: "Operational Closure — intercom-go P1: Error Taxonomy and Workspace Path Containment (002-S)"
date: 2026-09-03
mode: post-merge
shipment: 002-S
feature: 002-F
pr: 8
merge_commit_sha: 9efb156839e679fd7c29484fb99acd35fb9acd8b
compaction_status: done
closure_status: READY
releasability: READY
---

# Operational Closure — intercom-go P1 (002-S)

## Summary of the Change

Shipment `002-S` — "intercom-go P1: error taxonomy and workspace path
containment" — ports the AppError 14-variant taxonomy and the workspace
path-containment security control from the `agent-intercom` Rust behavioral
oracle (`softwaresalt/agent-intercom` @ `41df772`) into `intercom-go`:

* `internal/apperr`: 14 error `Kind`s, a display contract, sentinel errors
  with `errors.Is` support, constructors, and cause-preserving `Unwrap`.
* `internal/pathsafe`: a validated `Root` canonicalization type, lexical
  rejection and component normalization, component-aware containment
  enforcement in `Root.Resolve`, and symlink-escape detection for existing
  paths — the NON-NEGOTIABLE workspace-isolation control called out in the
  port brief §10.
* `docs/oracle-pin.md`: behavioral-oracle pin (commit `41df772`), the
  verification method, rejection of `references/herdr` as a mis-mount, and
  a 16-test oracle parity traceability table.

Phase P1 of the port roadmap; dependency root for all later phases. Config
schema, credential resolution, and later port phases remain deferred to
stash successors (`037B1552`/`4989A42D` rewrites, per the Stage
deliberation).

PR: https://github.com/softwaresalt/intercom/pull/8
Merge commit: `9efb156839e679fd7c29484fb99acd35fb9acd8b` (merge-commit
strategy, per P-009).

## CI Status and Review Items

* Hosted CI (final run, reviewed HEAD `093945b66b1bf8c4e07027110db36ec1d91ef553`):
  11/11 checks green — `ci gate`, `detect code changes`,
  `pipeline-topology (ambient)`, `test`, `lint`, `security`, `load
  cross-compile targets`, 4/4 `cross-compile` legs.
* Local review readiness: `READY` at reviewed HEAD `093945b66b1bf8c4e07027110db36ec1d91ef553`.
  This HEAD is tree-identical for source content to the last content commit
  (`a95e7c5`) recorded during the prior Ship session; the only commit between
  them (`093945b`) is a memory-only docs commit, so the local review remains
  fully valid — no re-review was required.
* Five parallel structured local reviews (Go Reviewer, Correctness Reviewer,
  Security Reviewer, Architecture Strategist, Constitution Reviewer),
  report-only mode: converged on 1×P1 (symlinked-intermediate-directory
  containment bypass, convergent across 3 independent reviewers) and 1×P2
  (`isRooted` not `GOOS`-gated). Both remediated in-session (commit
  `f3a7fda`) before PR creation; full validation suite re-confirmed green
  afterward.
* P-018 copilot-review gate: `NOT_APPLICABLE` (no engagement signal, no
  Copilot review requested, enforcement `auto`) — verified at merge time via
  `autoharness gate copilot-review 8 --repo softwaresalt/intercom
  --enforcement auto --json`, `exit_code: 0`.
* Unresolved review threads/comments: none (0 PR comments, 0 formal GitHub
  reviews, 0 review threads — confirmed via GraphQL immediately before
  merge).
* Repository merge-strategy check (P-009): `allow_merge_commit: true`;
  merged explicitly via `gh pr merge 8 --merge` (repo also permits
  squash/rebase, so the strategy was specified, not assumed from a
  default).
* Pipeline-topology gate: `lifecycle` phase passed (branch ownership,
  worktree topology, active-shipment invariant, shipment readiness) prior to
  both the pre-merge check and the shipment safe-close mutation.

### Merge Confirmation

* `gh pr view 8 --json state,mergedAt,mergeCommit`: `state: MERGED`,
  `mergedAt: 2026-09-04T00:03:34Z`, merge SHA
  `9efb156839e679fd7c29484fb99acd35fb9acd8b`.
* `git merge-base --is-ancestor 9efb156839e679fd7c29484fb99acd35fb9acd8b
  origin/main`: exit 0 — confirmed in `origin/main` history.
* Feature branch
  `feat/002-s-intercom-go-p1-error-taxonomy-and-workspace-path-containment`
  deleted from origin post-merge.

## Runtime Verification / Validator Evidence

No `runtime-verification` invocation was required for this shipment.
`internal/apperr` and `internal/pathsafe` are primitives-only library
packages with zero importers in this slice — neither `cmd/intercom` nor
`cmd/intercom-ctl` references them yet (confirmed via source scan). There is
no operator-facing runtime behavior, external integration, data path, or
migration in this slice for a validator to probe. The change's operational
surface is fully exercised by the standard Go validation suite and hosted CI
evidence above (identical rationale to the `001-S` foundation closure).

## Pre-Deploy Audits

* No config, flags, credentials, or migrations are introduced by this slice.
* `go.mod`/`go.sum` unchanged — zero new dependencies (stdlib only).
* `gofmt -l .` clean; `go vet ./...` clean; `golangci-lint run ./...` 0
  issues; `staticcheck ./...` clean; `govulncheck ./...` 0 vulnerabilities
  affecting this code (pre-existing package/module-level advisories not
  called by our code, consistent with `001-S`'s CVE baseline).
* `go test -race ./...` — full pass, including the symlink-privilege-gated
  suite in `internal/pathsafe` (privilege available in this environment;
  tests did not skip vacuously).
* `CGO_ENABLED=0 go build ./...` — succeeds (CGO-free mandate preserved).

## Deployment / Rollout Path

Merge-only. `internal/apperr` and `internal/pathsafe` are unconsumed library
primitives in this slice — there is no deployable artifact, operator-facing
behavior change, or release process triggered by this shipment. No canary,
phased rollout, or maintenance window applies.

## Post-Deploy Checks

* Confirm the next port-roadmap phase (which will consume
  `internal/apperr`/`internal/pathsafe`) continues to pass hosted CI without
  regressing the taxonomy or containment contracts pinned here.
* Confirm the deferred machine-checkable oracle-drift gate (follow-up
  `8ACF7110`) is picked up before dark-factory (unattended) execution of
  later phases, per the plan review's residual ARCH-5 finding.

## Healthy Signals

* CI gate aggregation (`ci gate` check) stays green on subsequent PRs.
* `go.mod`/`go.sum` remain stable until the next intentional dependency
  change.
* `internal/apperr` kind count (14) and `internal/pathsafe` containment
  regression suite remain stable as later phases begin consuming them.

## Failure Signals

* Any hosted CI run regressing the `security` job (staticcheck /
  govulncheck / gitleaks) without an explicit, reviewed cause.
* A future phase importing `internal/pathsafe` without exercising the
  symlinked-intermediate-directory containment regression test
  (`TestResolveRejectsSymlinkedIntermediateDirEscapingRoot`).
* Oracle drift: the pinned oracle commit (`docs/oracle-pin.md`, `41df772`)
  is currently documentary-only (out-of-tree, no CI-enforced check) — see
  follow-up `8ACF7110`.

## Monitoring Plan

* GitHub Actions run history for `.github/workflows/ci.yml` on `main`.
* No dashboards/alerts apply — this is a library-primitives shipment with no
  running service and no importers yet.

## Rollback Trigger / Procedure

* Trigger: a subsequent phase reveals the `apperr` taxonomy or `pathsafe`
  containment contract is structurally wrong (e.g. a real containment
  bypass discovered once a consumer exists).
* Procedure: `git revert` the merge commit
  `9efb156839e679fd7c29484fb99acd35fb9acd8b` on `main` via a standard PR (no
  direct-to-main mutation) — no deployed artifact exists to roll back
  independently of source control, since nothing yet imports these
  packages.

## Validation Window

One subsequent shipment cycle (the first later port phase that imports
`internal/apperr` and/or `internal/pathsafe`) — the taxonomy/containment
contracts are considered validated once a real consumer exercises them in
production-shaped code paths.

## Owner

Repository maintainer (`softwaresalt`) — no separate on-call/runtime owner
applies; this is a library-primitives shipment, not a running service.

## Risky Action Record

None. No destructive, high-blast-radius, or irreversible actions were taken
during this shipment's implementation or closure. The shipment-record close
required the P-015 cascade path (`backlogit shipment ship`) because the
generic `backlogit move --status shipped` path explicitly refuses direct
shipment status mutation in this backlogit version ("shipment must be
shipped via ShipShipment, not a direct status update") — this is the
intended, verified-precondition mechanism, not an ad hoc workaround; see
Shipment Closure Evidence below for full verification.

## Source Artifact Cleanup

* `002-F.custom_fields.source_stash_id`: not present — no stash entry to
  archive.
* `002-F.custom_fields.source_deliberation_id`: not present — no
  deliberation artifact to archive. (The plan/deliberation docs referenced
  by `002-F.references` — `docs/plans/2026-09-03-intercom-go-apperr-pathsafe-plan.md`,
  `docs/decisions/2026-09-03-intercom-go-port-continuation-deliberation.md`
  — are durable design documents, not backlogit deliberation artifacts, and
  are retained.)
* Archived: 0 stash, 0 deliberations.

## Deferred Stash Entries (P-021 C2, pre-existing at merge time)

Captured during this shipment's implementation/review (threadless/pre-PR
path, per source refs recorded in each entry):

| ID | Priority | Summary |
|---|---|---|
| `4A91CA81` | low | `NewRoot` drops cause via `apperr.New` (plan-mandated mechanism; deliberation required) |
| `7774C9CA` | low | Go-idiom P3 advisories (`splitPath` wrapper, `Kind` `Stringer`, `kindSentinel` bounds/drift test) |
| `3D61B5A8` | low | `hasPathPrefix` filesystem-root edge case; `NewRoot` directory-type check; GO-14 regression gap |
| `9A4C8749` | medium | `Newf` missing from `.golangci.yml` printf allowlist (shared lint config — deliberation required) |
| `2130906D` | medium | Plan's Constitution Check table principle-number mismatch (Stage-owned artifact correction; deliberation required) |
| `8ACF7110` | medium | Machine-checkable oracle-drift gate (deferred from plan review, depends on `002.003.001-ST`) |

No new deferred-scope-expansion captures were required during this
post-merge closure session itself (last-mile recheck found 0 open
P0/P1, 0 unresolved threads).

## Shipment Closure Evidence

* Reconciliation baseline: on session resumption, all 11 non-shipment
  manifest items (`002-F` + 3 tasks + 8 subtasks) were already present in
  `.backlogit/archive/` with declared `status: done` (relocated by the prior
  session's per-task completion flow, not yet transitioned to
  `status: archived`); the shipment record `002-S` remained `active` in
  `.backlogit/queue/`. `backlogit doctor` reported "No issues found" against
  this state. No orphans found; no other `002.*`-prefixed artifacts existed
  outside the manifest in either `queue/` or `archive/`.
* Close-path classification (P-015): `002-F` is a root feature (no
  `parent_id`) and its full descendant set at every depth — `002.001-T`,
  `002.001.001-ST`, `002.001.002-ST`, `002.001.003-ST`, `002.002-T`,
  `002.002.001-ST`, `002.002.002-ST`, `002.002.003-ST`, `002.002.004-ST`,
  `002.003-T`, `002.003.001-ST` — is exactly the shipment manifest; no other
  descendants exist; `002-F` carries no linked-deliberation ID. **Verified
  fully-covered-root cascade** selected.
* Executed `backlogit shipment ship 002-S --sha
  9efb156839e679fd7c29484fb99acd35fb9acd8b --message "merge: shipment 002-S
  — intercom-go P1: error taxonomy and workspace path containment (PR #8)"
  --author "Derek Williams <42183845+softwaresalt@users.noreply.github.com>"`:
  * `shipment_status: "shipped"`; `returned_ids: []` (no
    classifier/engine mismatch).
  * `archived_ids` (13: shipment + feature + 3 tasks + 8 subtasks) exactly
    matches the computed `allowed_ids`/`required_ids` sets (all 11 non-shipment
    manifest items were declared `status: done`, not `archived`, pre-close, so
    all were required-and-archived; no unexpected artifact archived, no
    required artifact left unarchived).
  * Every task's `parent_id` re-verified unchanged post-close against the
    pre-close snapshot.
* Shipment record: `status: archived`, `archived_status: shipped`,
  `commit: 9efb156839e679fd7c29484fb99acd35fb9acd8b` (verified via
  `backlogit get 002-S --json` and the raw archive frontmatter).
* Post-mode: `.backlogit/archive/002-S.md` present; `git status --short --
  ".backlogit/archive/"` showed only modifications/one rename, no
  unexpected deletions — `recommendation: PROCEED`.
* Backlog archival committed on
  `post-merge/002-s-intercom-go-p1-error-taxonomy-and-workspace-path-containment`
  (commit `72ea4d6`), not on `main` (P-020/branch-per-release-unit).
* File lock on `.backlogit/queue/002-S.md` acquired before the safe-close
  sequence and released immediately after (protocol compliance; no
  concurrent agents were active, single-agent session).

## Compaction Status (P-020)

`done`. `compact-context` (`target: memory`) ran in this closure session:
the two session memory files for this release unit
(`docs/memory/2026-09-03-stage-intercom-go-apperr-pathsafe.md`,
`docs/memory/2026-09-03-ship-intercom-go-apperr-pathsafe.md`) were
consolidated into
`docs/memory/compacted/2026-09-03-002-s-intercom-go-apperr-pathsafe-compacted.md`
and the verbose originals moved to `docs/archive/memory/`. A separate
leftover completed-work memory file from the already-closed `001-S`
release unit
(`docs/memory/2026-09-03-ship-post-merge-closure-001-s.md`, describing the
now-merged PR #5 closure) was also found eligible under the Phase 2
completed-work criterion; it was folded as an addendum into the existing
`docs/memory/compacted/2026-09-03-001-s-intercom-go-foundation-compacted.md`
and the verbose original archived. No other `docs/memory/` files exceeded
the threshold-gated candidate criteria (3 files / ~17 KB total, well under
the `max_files: 40` / `max_size_kb: 500` defaults) — this was a bounded,
per-release-unit Tier-1 consolidation.

## Releasability Evidence

**READY.** No blocking findings, no runtime risk (zero importers of the new
packages), no deployment/rollback complexity beyond a standard `git revert`
of the merge commit. All 6 stash-captured follow-ups are non-blocking,
deferred, and correctly out of this shipment's P-021 C1 scope.

## Dark Mode

Not applicable — this session ran in standard sequential mode.
`DARK_MODE_ACTIVE` was not present. All last-mile gate re-checks (§1.9,
P-018, P-009, pipeline-topology) were run as independent verifications
immediately before merge, and the operator's explicit continuation
directive ("Keep working autonomously until the task is truly finished,
then call task_complete") was scoped by the operator to PR #8 merge plus
its dependent post-merge closure work — not a dark-mode activation record.

---
title: "Post-merge closure: 007-S -- Build and scan script hardening (PR #22)"
description: "Operational closure artifact for shipment 007-S / feature 008-F"
status: "complete"
tags:
  - "closure"
  - "007-S"
  - "008-F"
  - "post-merge"
date: 2026-09-06
mode: post-merge
shipment: 007-S
feature: 008-F
pr: 22
merge_commit_sha: c521b2cf07555493ad35a1d20844acbba3f12045
compaction_status: done
closure_status: READY
releasability: READY
---

# Post-merge closure: 007-S / 008-F

**Mode:** post-merge
**PR:** #22, merged `2026-09-06T00:55:33Z`, merge commit `c521b2cf07555493ad35a1d20844acbba3f12045`
**Shipment:** 007-S (shipped, archived)
**Feature:** 008-F (done, archived)

## Summary of the change

Build and scan script hardening — resolves deferred stash findings `35D76D5E`
(build.ps1's I5 repo-root-escape guard is lexical-only, blind to symlinks/
junctions), `78D13775` (the retired-architecture scanner's TOML comment
masking is line-oriented, blind to multi-line strings), and `B6203CCA` (wire
`--self-test` into CI, relocated here from 007-F). No Go source (non-test)
changes. Five units: (U2) symlink/junction escape tests authored first; (U1)
`build.ps1`'s guard extracted into `scripts/lib/OutputPathGuard.ps1` and made
symlink/junction-aware while preserving the Ordinal case-sensitive comparison;
(U4) multi-line TOML fixtures + a verdict manifest authored before U3; (U3)
`check-retired-architecture.sh`'s masking rewritten to a real `tomllib` parse,
failing closed on parse errors, with `--self-test` now discovering fixtures by
glob + manifest; (U5) `--self-test` wired into CI as a second, blocking step
alongside the existing advisory bare scan.

## CI status and review

* CI: all checks green on the merged HEAD (`ci gate`, `pipeline-topology
  (ambient)`, `test`, `lint` (now including the new blocking self-test step),
  `security`, `cross-compile` x4, `gitignore append-only (I6)`, `detect code
  changes`).
* Standard multi-persona review (Security/Correctness/Maintainability/
  Constitution) — found and remediated a P1 (the defensive TOML-masking
  fallback lexer failed *open* on an unterminated multi-line string at EOF,
  reintroducing the exact R2 fail-open masking risk this shipment exists to
  close). One P2 advisory (the Windows-junction escape test has no Windows CI
  runner to execute on) captured as deferred entry, not fixed.
* Adversarial multi-model review (3 reviewers) — unanimous HIGH-confidence
  finding: the P1 fix above was itself untested dead code in CI (tomllib
  always wins on Python >=3.11). Fixed by running both the tomllib and
  fallback engines against every self-test fixture explicitly. A
  single-reviewer LOW-confidence finding, escalated given direct I5/
  Constitution-IV stakes, found the guard's ancestor walk used `Test-Path`,
  which can treat a *dangling* reparse point as "does not exist" and skip
  resolution — fixed by switching to `Get-Item -Force`; a regression test
  was added. Post-remediation re-review then caught a narrower gap in that
  fix itself (`-ErrorAction SilentlyContinue` folded all Get-Item failures,
  not just genuine absence, into the same branch) — narrowed to only catch
  `ItemNotFoundException`, failing closed on any other error.
* Copilot review (explicitly requested via `gh api .../requested_reviewers`
  since this repository has no collaborator-level Copilot bot by default;
  each subsequent HEAD advance required an explicit re-request since GitHub
  does not auto-re-review every push here): 3 rounds, 3 threads total.
  * Round 1 (HEAD `d5e79e1`): (1) PEP 585 generic-subscript type annotations
    (`list[str]`, `dict[str, bool]`) in the embedded Python fallback lexer
    crash at def-time on Python <3.9 -- exactly the interpreters the fallback
    exists to support. Fixed: `from __future__ import annotations` (PEP 563)
    defers all annotation evaluation to strings. (2) `$repoRootWithSep`
    double-separator false-rejection when the repo root is itself a
    filesystem/UNC root. Classified out of scope per P-021 C1 (pre-existing
    in the original `build.ps1` before this shipment touched it; fails safe,
    never a bypass; cannot manifest for this repository) -- captured as
    deferred entry `707FE72B`, replied to, thread resolved without a code
    change.
  * Round 2 (HEAD `2a68e50`): (3) `createDirectoryJunction` ran `pwsh`
    without a context timeout, unlike every other invocation in the test
    file, risking an indefinite stall. Fixed: bounded with the existing
    `outputPathGuardTimeout`, matching `runOutputPathGuard`'s pattern.
  * Round 3 (HEAD `6e3ca0c`): clean -- no new threads. `autoharness gate
    copilot-review 22` returned `SATISFIED: PASS` for the final merged HEAD.
  * All 3 threads replied to citing the fixing commit SHA (or, for the
    deferred one, the stash entry ID) and resolved via GraphQL
    `resolveReviewThread` before merge.
* No unresolved review items at merge time.

## Runtime verification

**Not applicable to the product's CLI/TUI surface.** This shipment touches no
`cmd/intercom`, `cmd/intercom-ctl`, WebSocket API, TUI, or config code path --
all changes are a build-tooling PowerShell script, a CI scanner shell/Python
script, CI workflow wiring, and Go integration tests for that tooling. The
`runtime_validation.validator_manifest` CLI-surface probe in
`.autoharness/workspace-profile.yaml` targets the product binaries themselves,
which are unchanged.

**Build-tooling self-verification (performed in lieu, since this shipment IS
the build tooling):** a live `pwsh -File scripts/build.ps1` sanity build
produced all 8 expected artifacts (2 binaries x 4 targets) after every code
change round; `bash scripts/check-retired-architecture.sh --self-test`
(dual-engine) and the bare repo scan both passed after every round; the full
cross-compile matrix job passed on the real `ubuntu-latest`/`windows`/
`macos` CI runners for the merged HEAD.

## Invariants to preserve

* Invariant I5 (`build.ps1` never writes outside the repo root) is now
  symlink/junction-aware, not just lexical -- any future edit to
  `scripts/lib/OutputPathGuard.ps1` must re-run
  `go test ./tests/integration/... -run TestResolveOutputPathWithinRoot` and
  keep the Ordinal (case-sensitive) comparison; do not loosen to
  case-insensitive.
* `check-retired-architecture.sh --self-test` must keep discovering
  `scripts/testdata/retired-*.toml` fixtures via `retired-manifest.json` and
  running BOTH the `tomllib` and fallback engines against every fixture --
  the fallback engine having its own test coverage (not just the primary
  path) is the specific invariant the adversarial-review remediation added;
  do not revert to single-engine self-test.
* The embedded Python in `check-retired-architecture.sh` needs
  `from __future__ import annotations` to stay compatible with the fallback
  path's own reason for existing (pre-3.11 Python); do not remove it while
  PEP 585 generic-subscript annotations remain in that code.
* `--self-test`'s bare-scan re-run at the end (`scan_with_mode repo`) means
  the `lint` job's new self-test step is effectively blocking on the real
  tracked tree too, even though the dedicated bare-scan step stays
  `continue-on-error: true` -- this net posture change is intentional
  (U5 AC-2) but easy to miss reading the advisory step in isolation.

## Pre-deploy audits

Not applicable -- no deployment, no migration, no config/flag change
consumed by production code, no access change.

## Deployment or rollout path

Merge-only. The hardened guard and scanner take effect on the next
`scripts/build.ps1` invocation and the next CI run against `main`
(already exercised live on this PR -- see CI status above).

## Post-deploy checks

The merge-commit CI run on `main` fast-forwarding to `c521b2cf` is itself the
post-deploy check for this build/scan-tooling shipment: `ci gate` and every
job it aggregates, including `lint`'s new blocking self-test step, passed on
the real runners.

## Risky action record

| Action | Risk | Outcome |
|---|---|---|
| Rewriting the I5 containment guard's ancestor-walk existence check | High (a security containment control; a wrong existence check could silently reopen a bypass) | Caught in review, not silently merged: adversarial review found the `Test-Path`-based check could miss a dangling reparse point; re-review then found the fix's own error-suppression was too broad. Both narrowed and re-verified with regression tests before merge. |
| Rewriting the TOML masking from a hand-rolled lexer to a real parser | High (R2 in the plan: an over-masking bug would make the scanner report clean on genuinely contaminated config, the worst failure mode for an unattended gate) | Positive-control fixtures (U4 AC-1) caught this class directly; the fallback lexer's own EOF fail-closed logic was found to be untested dead code in CI and is now exercised by every self-test run against both engines. |
| Wiring a new blocking CI step (`--self-test`) into the `lint` job | Low-medium (could newly block otherwise-mergeable PRs) | Landed last (U5), only after U3/U4 were green; the existing advisory bare-scan step's posture is explicitly unchanged. |
| Requesting Copilot review on a repository where it is not a default collaborator | Low (informational uncertainty, not a delivery risk) | REST `requested_reviewers` POST with the `[bot]`-suffixed login succeeded where the GraphQL-style unsuffixed login and `gh pr edit --add-reviewer` both failed; documented in the PR body for future shipments on this repo. |

## Healthy signals

* `go build -mod=readonly ./...`, `go vet ./...`, `gofmt -l .`,
  `goimports -l .`, `golangci-lint run ./...` (0 issues), and
  `go test -race -mod=readonly ./...` all green, re-verified after every
  remediation round.
* `bash scripts/check-retired-architecture.sh --self-test` passes both
  engines against all 6 fixtures (old + 5 new); bare repo scan clean.
* Live `pwsh -File scripts/build.ps1` sanity builds (8/8 artifacts) after
  every round.
* All 12 CI checks green on the final merged HEAD, including the new
  blocking self-test step.
* `autoharness gate copilot-review 22` -> `SATISFIED: PASS` for the final
  merged HEAD; all 3 review threads resolved.

## Failure signals

* A future regression in `Resolve-ReparseAwarePath`'s ancestor walk would be
  caught by `TestResolveOutputPathWithinRoot`, including its
  `rejects_dangling_symlink_escape` case -- investigate that test before
  assuming the guard itself is fine.
* A future over-masking regression in the TOML scanner would be caught by
  `retired-positive-after-multiline-close.toml` / `retired-positive-
  escaped-delimiter-close.toml` failing `--self-test` under EITHER engine --
  the dual-engine self-test is the load-bearing signal here.
* The Windows-junction escape test (`rejects_junction_escape`) has no
  Windows CI runner to execute on (deferred entry `DA945722`) -- a
  regression in junction handling specifically would only be caught by a
  human running `go test` locally on Windows until that gap is closed.

## Monitoring plan

* No dashboards/alerts affected -- this is build/CI-internal tooling, not a
  deployed service.
* Watch the first several ordinary PRs exercising the new blocking
  `--self-test` CI step for false positives.

## Rollback trigger / rollback procedure

If this shipment's changes need to be rolled back:
`git revert -m 1 c521b2cf07555493ad35a1d20844acbba3f12045` (reverting a merge
commit requires the mainline parent via `-m 1`). Single merge commit revert;
all 5 units (U1-U5) landed in one shipment with no other consumers depending
on `scripts/lib/OutputPathGuard.ps1` or the new fixture-discovery convention
yet.

## Validation window

Through the next several ordinary PRs exercising the new blocking
`check-retired-architecture.sh --self-test` CI step and any future edit to
`scripts/build.ps1` or `scripts/check-retired-architecture.sh` under real
traffic.

## Owner

Ship agent (this session) / repository maintainer for: (a) disposition of
the two P-021 deferred entries below; (b) any future decision to add a
Windows CI runner (would close the junction-test coverage gap noted above).

## Compaction status (P-020)

`done`. `compact-context` invoked with `target: all` -- see the accompanying
session memory and compaction commit for the specific artifacts consolidated
this cycle.

## Releasability evidence

**READY.** No runtime surface changed; `runtime_validation.releasability`
requirements in the workspace profile are not applicable to this shipment.
All CI, local review, adversarial review, and Copilot review gates passed
for the merged HEAD, including three full remediation-and-re-review cycles
with zero residual blocking findings.

## Follow-up items (P-021 deferred stash entries)

* `DA945722` (low, requires deliberation) -- the Windows-junction escape
  test (`TestResolveOutputPathWithinRoot/rejects_junction_escape`) never
  runs in CI because `.github/workflows/ci.yml` has no Windows runner.
  Symlink-escape coverage is continuously exercised on Linux CI, so this is
  a coverage-completeness gap, not an unexercised guard entirely.
* `707FE72B` (low, requires deliberation) -- `OutputPathGuard.ps1`'s
  `repoRootWithSep` prefix comparison double-separates (and therefore
  falsely rejects legitimate in-root paths) when the repo root itself
  resolves to a filesystem/UNC root. Pre-existing in the original
  `build.ps1` before this shipment; fails safe, not a containment bypass;
  cannot manifest for this repository. Found by Copilot review.

## Source artifact cleanup

* `008-F.custom_fields.source_stash_id`: not present on this feature record
  (008-F's description names the three resolved stash findings --
  `35D76D5E`, `78D13775`, `B6203CCA` -- as plain text, not as a single
  literal `custom_fields.source_stash_id`) -- no `backlogit_stash_archive`
  action applicable via that mechanism. Independently confirmed all 3 named
  stash IDs are **no longer present** in the active stash (`backlogit stash
  get <id>` returns `not found` for each) -- already consumed by Stage's
  prior harvest/triage before this session began -- so no discretionary
  stash archival was performed by Ship, consistent with Ship's role
  boundary (discretionary stash archival is Stage-only).
* `008-F.custom_fields.source_deliberation_id`: not present as a single
  literal field; the feature references
  `docs/decisions/2026-09-04-intercom-go-residual-hardening-triage-deliberation.md`
  as a plain markdown `references` entry, not as a `backlogit` artifact with
  its own `artifact_type: deliberation` record -- no `backlogit_archive_item`
  action applicable via this mechanism.

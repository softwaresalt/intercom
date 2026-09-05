---
title: "Post-merge closure: 006-S -- CI supply-chain hardening (PR #20)"
description: "Operational closure artifact for shipment 006-S / feature 007-F"
status: "complete"
tags:
  - "closure"
  - "006-S"
  - "007-F"
  - "post-merge"
date: 2026-09-05
mode: post-merge
shipment: 006-S
feature: 007-F
pr: 20
merge_commit_sha: c627a9d3e28354b10323bf1d6605529fc679c267
compaction_status: pending
closure_status: READY
releasability: READY
---

# Post-merge closure: 006-S / 007-F

**Mode:** post-merge
**PR:** #20, merged `2026-09-05T22:56:33Z`, merge commit `c627a9d3e28354b10323bf1d6605529fc679c267`
**Shipment:** 006-S (shipped, archived)
**Feature:** 007-F (done, archived)

## Summary of the change

CI supply-chain hardening — the trust root for all later dark-factory
shipments. No Go source changes. Five units: (U1) hash-pinned,
`--only-binary=:all:` `autoharness` install via a committed lock file
(replacing a floating `pip install`); (U2) a new scheduled full-history
gitleaks scan, redacted and deduped, routed to a GitHub-issue operator
handoff on a finding; (U3) advisory-only `CODEOWNERS` for CI-critical
paths, gated by a blocking AC-0 branch-protection/ruleset pre-flight; (U4a/
U4b) test-first fixtures plus the real subsequence-predicate implementation
for the `.gitignore` append-only invariant (I6); (U5) wiring that checker
into `ci.yml` as a required, always-fail-closed gate.

## CI status and review

* CI: all checks green on the merged HEAD (`ci gate`, `pipeline-topology
  (ambient)`, `test`, `lint`, `security`, `cross-compile` x4, `gitignore
  append-only (I6)`, `detect code changes`).
* Standard review (Security/Correctness/Maintainability/Constitution/Scope)
  — found and remediated a P0 (missing `typing-extensions` transitive
  dependency, would have broken `--require-hashes` on every CI run).
* Pass-2 adversarial multi-model review (4-reviewer panel) — FAIL verdict
  on 5 findings, all remediated: CODEOWNERS self-contradiction,
  `--only-binary=:all:` hardening, CRLF/first-ever-file edge cases in the
  gitignore checker, gitleaks exit-code conflation, handoff-issue dedupe.
  Two findings genuinely out of scope captured as P-021 deferred entries
  (`11ECB954`, `AFE0E95C`) rather than fixed. Full report:
  `docs/closure/2026-09-04-007-f-ci-supply-chain-hardening-pass2-adversarial-review.md`.
* Post-remediation re-review (independent verification) caught a regression
  in the first fix pass itself (a markdown-table-escaping bug, and a
  base-ref-absence-check ambiguity) — both fixed and re-verified; final
  independent verify pass returned PASS.
* **CI-caught version regression** (real GitHub Actions run, not caught in
  local review): the hash-pinned lock file initially pinned
  `autoharness==1.4.11`, resolved against this sandbox's internal package
  proxy whose "latest" view was stale relative to real public PyPI. CI
  failed with `Unknown gate subcommand: pipeline-topology` (that subcommand
  doesn't exist in 1.4.11). Corrected to `1.5.0` (confirmed via
  `https://pypi.org/pypi/autoharness/json` directly and this repo's own
  last-known-good CI log) — re-verified fully green on the real runner.
* Copilot review (auto-triggered on PR open, then explicitly re-requested
  via GraphQL `requestReviews` for the fix commit since GitHub does not
  auto-re-review every push): 3 comments — job-level `continue-on-error`
  masking genuine install failures (scoped down to just the verdict step);
  `grep -Ev` exit-1-on-empty-result mishandling; `mapfile`-via-process-
  substitution error-swallowing risk. All 3 fixed, replied to citing the
  fixing commit, threads resolved via GraphQL. Fresh review on the fix
  commit returned clean. `autoharness gate copilot-review` → `SATISFIED:
  PASS` for the final merged HEAD.
* No unresolved review items at merge time.

## Runtime verification

**Not applicable.** This shipment touches no runtime surface: no changes to
`cmd/intercom`, `cmd/intercom-ctl`, the WebSocket API, the TUI, or any
config surface consumed by production code. All changes are CI workflow
configuration, a hash-pin lock file, ownership metadata, and a standalone
shell script plus its test fixtures. The `runtime_validation` block in
`.autoharness/workspace-profile.yaml` does not apply to this change set.

## Invariants to preserve

* No floating (`==`-less) `pip install autoharness` remains in any
  workflow; the hash-pinned install always targets the committed lock
  file with `--require-hashes --only-binary=:all:`.
* `secret-scan-history.yml`'s gitleaks version/sha256 stay byte-identical
  to `ci.yml`'s per-PR scan's pinned values.
* `CODEOWNERS` remains advisory-only until a deliberate, separate operator
  decision provisions an unattended-PR approval path and flips branch
  protection (see Risk R1 below) — this shipment never modifies
  branch-protection/ruleset settings itself.
* The `.gitignore` append-only checker (`scripts/check-gitignore-append-only.sh`)
  remains fail-closed on an unresolvable `--base-ref` (AC-4); the only
  skip path is the explicit `--allow-no-merge-base` flag, which CI never
  passes.

## Pre-deploy audits

Not applicable — no deployment, no migration, no config/flag change
consumed by production code, no access change (`CODEOWNERS` is advisory
metadata, not an access grant).

## Deployment or rollout path

Merge-only. The new/changed CI workflow behavior takes effect on the next
push/PR to `main` (already exercised live on this PR itself — see CI status
above) and on the next scheduled Monday 06:00 UTC run for
`secret-scan-history.yml`.

## Post-deploy checks

The merge-commit CI run on `main` fast-forwarding to `c627a9d3` is itself
the post-deploy check for this CI-config-only shipment: `ci gate` and every
job it aggregates (including the newly-added `gitignore append-only (I6)`
and the corrected `pipeline-topology (ambient)`) passed on the real
`ubuntu-latest` runner.

## Risky action record

| Action | Risk | Outcome |
|---|---|---|
| Hash-pinning `autoharness`'s install using a version resolved from a stale internal proxy | High (initially shipped a wrong pin that broke CI's own topology gate) | Caught immediately by CI on this PR (not silently merged); corrected to the version confirmed via public PyPI and this repo's own real CI log; re-verified fully green before merge |
| Shipping `CODEOWNERS` (ownership metadata) on a CI-critical path set | Medium (R1: could deadlock unattended merges if branch protection is ever set to require code-owner review) | AC-0 blocking pre-flight verified no such rule is live today (`gh api` against both branch protection and repository rulesets); file ships advisory-only; flipping enforcement is an explicit, separate, future operator decision |
| Full-history secret scan surfacing a pre-existing historical finding | Medium (first-ever scan of full git history) | Job is schedule/dispatch only, never PR-required; a finding routes to a deduped GitHub-issue operator handoff, redacted (`--redact=100`), metadata-only; does not block delivery |
| Wiring a new required gate (`gitignore-append-only`) into `ci-gate`'s aggregation | Low-medium (could newly block otherwise-mergeable PRs) | Test-first fixtures (insertion/deletion/reorder) plus a `--self-test` regression guard in CI; fail-closed base-ref resolution verified against real repo history, a synthetic first-ever-file commit, and an invalid-ref case |

## Healthy signals

* `go build ./...`, `go vet ./...`, `gofmt -l .`, `go test ./...` all green
  on `main` post-merge (no Go files were touched by this shipment).
* `actionlint` clean on both modified/new workflow files.
* `bash scripts/check-gitignore-append-only.sh --self-test` passes (3/3
  fixtures) — wired into CI as a standing regression guard.
* `pip install --require-hashes --only-binary=:all: -r
  .github/constraints/autoharness-lock.txt` succeeds on the real
  `ubuntu-latest` runner (confirmed in the merged PR's own CI run).

## Failure signals

* A future `pip install --require-hashes` failure on `topology-check` (now
  never maskable by `continue-on-error`, which is scoped to only the
  topology-verdict step) — investigate the lock file's hashes/versions
  against the actual CI runtime before assuming compromise.
* `secret-scan-history.yml` turning red on its weekly schedule — expected
  signal, not a bug; triage via the deduped GitHub issue it opens.
* Any future PR attempting to modify `.gitignore` non-append-only would be
  caught by `gitignore-append-only (I6)`; a false-positive block there
  should be investigated against
  `scripts/check-gitignore-append-only.sh`'s subsequence predicate before
  assuming the invariant itself is wrong.

## Monitoring plan

* Watch the first scheduled run of `secret-scan-history.yml` (next Monday
  06:00 UTC) for a clean result or a legitimate historical finding.
* No dashboards/alerts affected — this is CI-internal tooling, not a
  deployed service.

## Rollback trigger / rollback procedure

If this shipment's CI changes need to be rolled back:
`git revert -m 1 c627a9d3e28354b10323bf1d6605529fc679c267` (reverting a
merge commit requires the mainline parent via `-m 1`). This is a single
merge commit revert; all 5 units (U1-U5) landed in one shipment with no
other consumers, so a revert is clean. `CODEOWNERS` and
`.github/constraints/autoharness-lock.txt` have no other shipment
depending on their presence yet.

## Validation window

Through the first scheduled `secret-scan-history.yml` run (next Monday
06:00 UTC) and the next few ordinary PRs exercising `gitignore-append-only
(I6)` and the hash-pinned install under real traffic.

## Owner

Ship agent (this session) / repository maintainer for: (a) any future
decision to flip `CODEOWNERS` to enforced (R1, requires provisioning an
unattended-PR approval path first); (b) triage of any
`secret-scan-history.yml` finding; (c) disposition of the two P-021
deferred entries below.

## Compaction status (P-020)

Set by the `compact-context` invocation immediately following this
artifact's creation (see the session's final memory/compaction note for
the outcome).

## Releasability evidence

**READY.** No runtime surface changed; `runtime_validation.releasability`
requirements in the workspace profile are not applicable to this shipment.
All CI, review, and quality gates passed for the merged HEAD, including a
real GitHub Actions verification of the corrected hash-pinned install.

## Follow-up items (P-021 deferred stash entries)

* `11ECB954` (medium, requires deliberation) — extend the `.gitignore`
  append-only checker to understand gitignore pattern semantics (e.g. a
  `!negation` line bypass), not just line-ordering. Flagged
  LOW-confidence/CRITICAL by the pass-2 adversarial review as the sharpest
  conceptual gap in the I6 invariant as currently defined.
* `AFE0E95C` (medium) — pin the installer itself (pip/setuptools/wheel,
  Python patch version), not just `autoharness`'s own resolved closure.
  Flagged MEDIUM-confidence/MAJOR by the pass-2 adversarial review.

## Source artifact cleanup

* `007-F.custom_fields.source_stash_id`: not present on this feature record
  (007-F was harvested via Stage's standard triage/plan flow from four
  stash entries — `EFFAA358`, `90350C9A`, `F7C6420D`, `BEDD2E70` — named in
  its own description text, not recorded as a single literal
  `custom_fields.source_stash_id`) — no `backlogit_stash_archive` action
  applicable via that mechanism. Independently confirmed all 4 named stash
  IDs are **no longer present** in the active stash (`backlogit stash
  list`) — already consumed by Stage's prior harvest/triage — so no
  discretionary stash archival was performed by Ship, consistent with
  Ship's role boundary (discretionary stash archival is Stage-only).
* `007-F.custom_fields.source_deliberation_id`: not present as a single
  literal field; the feature references
  `docs/decisions/2026-09-04-intercom-go-residual-hardening-triage-deliberation.md`
  as a plain markdown `references` entry, not as a `backlogit` artifact
  with its own `artifact_type: deliberation` record — no
  `backlogit_archive_item` action applicable via this mechanism.

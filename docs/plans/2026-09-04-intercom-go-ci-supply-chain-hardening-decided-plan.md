---
title: "Decided Plan — CI supply-chain and gate integrity hardening (007-F)"
date: 2026-09-04
consolidated: 2026-09-05
status: shipped
feature: 007-F
shipment: 006-S
pr: 20
merge_commit_sha: c627a9d3e28354b10323bf1d6605529fc679c267
source_deliberation: "docs/decisions/2026-09-04-intercom-go-residual-hardening-triage-deliberation.md"
resolved_stash: [EFFAA358, 90350C9A, F7C6420D]
partially_resolved_stash: [BEDD2E70]
---

# Decided Plan — CI supply-chain and gate integrity (007-F)

Verbose original with full deliberation/remediation history archived at
`docs/archive/plans/2026-09-04-intercom-go-ci-supply-chain-hardening-plan.md`.

## Problem

Five deferred findings all targeted the CI trust root: the pipeline that
validates every later shipment under dark-factory (unattended) operation.

## Scope fence (Non-Goals, held throughout)

* No Go source changes.
* Does NOT flip `PIPELINE_TOPOLOGY_GATE_REQUIRED` to required, and does NOT
  modify branch-protection/ruleset settings.
* Does not alter `check-retired-architecture.sh`'s scanner logic (that is
  008-F's scope; this shipment only invokes it, unchanged).

## Final implementation units (all shipped)

* **U1** — Hash-pinned, `--require-hashes --only-binary=:all:` `autoharness`
  install via a committed lock file (`.github/constraints/autoharness-lock.txt`),
  replacing a floating `pip install`. Final pin: `autoharness==1.5.0` (corrected
  post-merge-attempt from an initial `1.4.11` mis-resolution — see the
  compound learning at
  `docs/compound/2026-09-05-pip-index-proxy-staleness-vs-real-ci.md`).
* **U2** — Scheduled (weekly) + `workflow_dispatch` full-history gitleaks
  scan (`secret-scan-history.yml`), reusing the exact pinned gitleaks
  version/hash as the per-PR scan. `--redact=100`, no raw report upload,
  deduped GitHub-issue operator handoff on a finding. Not PR-required.
* **U3** — Advisory-only `CODEOWNERS` for CI-critical paths. Ships
  ownership metadata only; explicitly does **not** close the
  self-modifying-PR gap (see Residual risk below). AC-0 blocking pre-flight
  (probe effective branch-protection/rulesets before creating the file)
  verified clean at implementation time.
* **U4a/U4b** — `.gitignore` append-only invariant (I6): "the old file's
  non-comment/non-blank lines are an ordered subsequence of the new file's."
  Fixtures + `--self-test` driver shipped test-first, against an
  intentionally-failing stub, before the real subsequence-predicate logic.
* **U5** — Wired the checker into `ci.yml` as a new `gitignore-append-only`
  job, added to `ci-gate`'s required aggregation.
* **U6** (originally B6203CCA / `check-retired-architecture.sh --self-test`
  wiring) — **relocated to 008-F** as its final unit; landing it here would
  have made 008-F's own test-first fixtures unmergeable (see Rejected
  alternatives below).

## Key constraints that survived review

* U1 must be hermetic: exact-version + SHA-256 hash for the full transitive
  closure, `--require-hashes` AND `--only-binary=:all:` (the latter added
  during PR review to close the unhashed-sdist-build-isolation fallback
  `--require-hashes` alone leaves open).
* U3 must never risk deadlocking unattended merges: AC-0's pre-flight probe
  is blocking and must abort the unit if code-owner review is already
  enforced anywhere in the effective branch-protection/ruleset set.
* U4b must fail closed on an unresolvable comparison base; skip only via an
  explicit `--allow-no-merge-base` flag CI never passes.
* U5's CI wiring must scope `fetch-depth: 0` to only the one job that needs
  it, and must name its target job explicitly (not left to implementer
  judgment).

## Rejected alternatives

* **U1 as an `==` version pin alone** (no hashes) — rejected: still permits
  arbitrary code execution via an unhashed distribution or a compromised
  mutable index; not actually hermetic.
* **U3 claimed as closing the self-modifying-PR gap** — rejected: a
  `pull_request`-triggered workflow runs from the PR's own head, so a PR
  can edit the very workflow/scripts that validate it; advisory CODEOWNERS
  has no enforcement effect against that. Narrowed to "ownership metadata"
  only.
* **U6 landing in 007-F ahead of 008-F's masking-logic change** — rejected:
  would make 008-F's own test-first fixtures (which must fail against the
  current masking implementation) unmergeable without immediately
  reddening a gate 007-F had just installed. Relocated to 008-F as its
  final unit instead.
* **U4's original fail-open base-resolution behavior** (skip silently on an
  unresolvable merge-base) — rejected: defeats the very invariant the gate
  exists to enforce. Made fail-closed.

## Residual risk (accepted, carried to operator)

**R1 (highest risk in this plan) — CODEOWNERS could deadlock unattended
operation** if branch protection is ever set to "Require review from Code
Owners": every dark-factory PR touching `.github/workflows/**` would block
awaiting a review that isn't coming (`merge_approval_pre_authorized` does
not satisfy a GitHub branch-protection code-owner-review requirement).
Mitigated by shipping the file advisory-only and verifying no such rule is
live today; flipping enforcement remains a deliberate, separate, future
operator decision requiring an unattended-PR approval path first.

**R5 — Version-pin staleness** is an accepted, deliberate tradeoff (U1
AC-2's bump-expectation comment); U1's own maintenance procedure documents
the re-resolution steps for future bumps, now informed by the pip-index-
proxy-staleness lesson from this shipment's own remediation.

## Deferred, out of this shipment's scope (P-021 stash entries)

* `11ECB954` — extend the I6 checker to understand gitignore-pattern
  semantics (e.g. `!negation` bypass), not just line-ordering.
* `AFE0E95C` — pin the installer itself (pip/setuptools/wheel), not just
  `autoharness`'s own resolved closure.

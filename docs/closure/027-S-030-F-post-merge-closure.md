---
title: "Post-merge closure: 027-S / 030-F — Write-path gate kill switch (plan unit 3)"
description: "Post-merge closure artifact for shipment 027-S / feature 030-F"
status: "complete"
tags:
  - "operational-closure"
  - "027-S"
  - "030-F"
  - "post-merge"
date: 2026-09-20
mode: post-merge
shipment: 027-S
feature: 030-F
pr: 71
merge_commit_sha: 5208085379da4b945eae6d341089f0f985adc09c
compaction_status: done
releasability: READY
---

# Post-merge closure: 027-S / 030-F — Write-path gate kill switch (plan unit 3)

## Summary

Shipment `027-S` implements `030-F` ("Write-path gate kill switch"), plan
unit 3 of the intercom-go gate-reliability plan. It splits
`scripts/check-write-path-precondition.sh` into a fixtures-only
`--self-test-integrity` mode (unconditionally blocking) and the existing
combined `--self-test` mode, and splits the single CI step in
`.github/workflows/ci.yml` into an unconditionally-blocking integrity step
plus a separately `WRITE_PATH_GATE_ADVISORY`-toggled verdict step, mirroring
the existing `check-retired-architecture.sh` / `RETIRED_ARCH_GATE_ADVISORY`
split. **Zero production `cmd/`/`internal/` runtime Go code was added or
modified** — this is a CI-gate/tooling change only (shell script + workflow
YAML).

## Invariants to preserve

* `--self-test-integrity` runs fixtures only and never the tracked-tree
  repo scan (AC-0.1).
* The new CI integrity step is unconditionally blocking; the advisory
  toggle cannot disable it (AC-0.2).
* The new CI verdict step honours `WRITE_PATH_GATE_ADVISORY` via
  step-level `continue-on-error`, case-insensitive `'true'` opt-in,
  fail-closed default (AC-0.3).
* The LOCAL DIVERGENCE tracking comment enumerates both new steps and the
  new variable name so `autoharness tune` does not silently drop them
  (AC-0.4).
* Verdicts at the bound snapshot are unchanged — pure additive mode, no
  selector/masking drift (AC-0.5).

## Validator manifest application

Per `.autoharness/workspace-profile.yaml` `runtime_validation.validator_manifest`,
the only declared runtime surface for this repository is `cli` (`go run
./cmd/... --help` smoke). This shipment does not touch `cmd/intercom` or
`cmd/intercom-ctl`; the CLI smoke probe would exercise general
process-startup health only, not the changed surface (a CI gate shell
script and its workflow wiring). Recorded as **not applicable to this
shipment's diff** rather than fabricated evidence.

## Substantive evidence

The actual "runtime" surface exercised by this shipment is the CI gate
script itself, run directly in both its pre-existing and new modes:

* PR #71 (branch `feat/027-s-write-path-gate-kill-switch`, reviewed/merged
  HEAD `341f760ae5e92005ad532c375364fd4513dd2cbb`):
  * `go build ./...`, `go vet ./...`, `gofmt -l .`, `go test ./...`
    (integration suite, 160.5s) — all green, recorded in the PR's Local
    Review Readiness block.
  * `bash scripts/check-write-path-precondition.sh --self-test` — pass
    (regression, unchanged mode)
  * `bash scripts/check-write-path-precondition.sh --self-test-integrity` — pass (new mode)
  * `bash scripts/check-retired-architecture.sh --self-test-integrity` — pass (sibling gate, unaffected)
  * `python -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))"` — pass
  * `actionlint .github/workflows/ci.yml` — pass
  * All 13 CI checks green at merge HEAD `341f760ae5e92005ad532c375364fd4513dd2cbb`
    (cross-compile ×4, lint, security, test, test (windows, advisory),
    pipeline-topology, ci gate, gitignore regression, detect-changes,
    load-cross-compile-targets)
  * Merged via merge-commit strategy: `5208085379da4b945eae6d341089f0f985adc09c`
    (repository is merge-commit-only: `allow_merge_commit=true`,
    `allow_squash_merge=false`, `allow_rebase_merge=false` — P-009 satisfied)
  * P-018 Copilot-review gate: `NOT_APPLICABLE` (no Copilot engagement signal
    on PR #71; enforcement mode `auto`)

* Post-merge closure branch (`post-merge/030-write-path-gate-kill-switch`):
  this branch's own diff prior to closure work was confined to the merge
  commit itself; closure-branch-added diff is `.backlogit/` archival
  artifacts (backlog-only). Re-verified on this branch's HEAD:
  * `go vet ./...` — clean
  * `gofmt -l .` — clean (zero unformatted files)
  * `bash scripts/check-write-path-precondition.sh --self-test` — pass
  * `bash scripts/check-write-path-precondition.sh --self-test-integrity` — pass
  * `go test ./...` not re-run to completion on this branch — no Go source
    changed relative to the already-green PR #71 merge HEAD; recorded as
    non-applicable evidence for a backlog-only diff rather than fabricated.

## Pre-deploy audits

None applicable — no migration, flag, config, or access change. The change
is confined to CI tooling that runs on every future PR.

## Deployment / rollout path

Merge-only. The new CI steps take effect on the next PR run against `main`;
there is no separate deploy or release step for this repository's CI
tooling.

## Post-deploy checks

The next PR touching write-path-relevant Go source will exercise both new
CI steps (`--self-test-integrity` unconditionally, and the advisory-toggled
verdict step per `WRITE_PATH_GATE_ADVISORY`). Confirm on that PR's CI run
that:
* the integrity step is present and blocking (no `continue-on-error`)
* the verdict step honors the advisory toggle when set

## Risky action record

None. No destructive, irreversible, or elevated-privilege action was taken.
The covering-feature completion gate transitioned `030-F` `active -> done`
via `backlogit move` after independently verifying all five required
conditions (topology intact, containment holds, no live descendant at any
depth, descendant graph set-equal to the manifest, feature status exactly
`active`). The shipment closure used the P-015 verified fully-covered-root
`CASCADE` path (`backlogit shipment ship`), gated by a two-set
`allowed_ids`/`required_ids` verification against the pre-close
declared-status snapshot; see
`.backlogit/reconcile/027-S-safe-close-20260920T180734Z.md` for the full
verification record. No cascade anomaly was detected.

## Healthy signals

* CI continues green on subsequent PRs with the new steps present.
* `WRITE_PATH_GATE_ADVISORY` toggle behaves symmetrically to
  `RETIRED_ARCH_GATE_ADVISORY`.

## Failure signals

* A future PR's integrity step fails to block despite no advisory flag set
  (regression of AC-0.2).
* The LOCAL DIVERGENCE enumeration drops the new steps after an
  `autoharness tune` run (regression of AC-0.4), silently reverting this
  shipment's CI wiring.

## Monitoring plan

Standard CI dashboard observation on `main` for the next several PRs that
touch `scripts/check-write-path-precondition.sh` or its CI step wiring.

## Rollback trigger

If the new integrity step false-positives and blocks unrelated PRs
(including its own revert), or the verdict-step advisory toggle fails to
suppress a false positive.

## Rollback procedure

Revert merge commit `5208085379da4b945eae6d341089f0f985adc09c` on `main`
via a standard merge-commit revert PR. No data or state migration is
associated with this change, so revert is a pure code/workflow rollback.

## Validation window

Through the next 2–3 PRs on `main` that trigger the write-path gate CI
steps (expected within the current planning cycle, plan units 4+).

## Owner

Ship agent / repository maintainer (`softwaresalt/intercom`), consistent
with existing gate-reliability plan unit ownership.

## Compaction status (P-020)

`done`. Ship Step 6 item 8's mandatory `compact-context` invocation ran with
`target: all` and compacted the two 027-S-scoped session memory checkpoints
(`2026-09-20-ship-027-s-pr-ready-awaiting-merge-approval.md`,
`2026-09-21-ship-027-s-halted-at-merge-approval-gate.md`) into
`docs/memory/compacted/2026-09-21-027-s-030-f-compacted.md`, moving the
verbose originals to `docs/archive/memory/`. No plan consolidation applied
(the parent gate-reliability plan still has Units 1/2/3/4 outstanding — not
complete). No closure-record compaction applied (this artifact is fresh, 0
days old).
## Releasability evidence

**READY.** All required evidence is satisfied:
* Local review readiness: `READY_WITH_FOLLOWUPS` (P0=0, P1=0) at reviewed
  HEAD `341f760ae5e92005ad532c375364fd4513dd2cbb`, unchanged through merge.
* All 13 required CI checks green at merge HEAD.
* P-009 (merge-commit-only) satisfied.
* P-018 (Copilot review) `NOT_APPLICABLE`.
* P-016 (pipeline topology) `passed` at every gate invocation (pre-claim,
  post-claim, lifecycle ×2, pre-closure).
* No runtime service surface affected; CI-gate substantive evidence (self-test
  and self-test-integrity fixture runs) green on both the feature branch and
  the post-merge closure branch.
* 4 pre-existing P-021 deferred-scope-expansion follow-ups remain
  untriaged/unplanned in this run per this session's DARK_MODE_ACTIVE scope
  contract (Stage-owned triage, out of scope for Ship).

Not applicable: no manual checkpoint, external service, or deployment gate
applies to this CI-tooling-only shipment.


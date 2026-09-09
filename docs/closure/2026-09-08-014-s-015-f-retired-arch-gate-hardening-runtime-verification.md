---
title: "Runtime verification: 014-S / 015-F — retired-architecture gate detection-quality hardening"
description: "Runtime/black-box verification report for shipment 014-S / feature 015-F"
status: "complete"
date: 2026-09-08
mode: post-merge
shipment: 014-S
feature: 015-F
pr: 46
merge_commit_sha: d7be7e883f1d309a22037aae97ec769cb647b020
verdict: READY
---

# Runtime verification: 014-S / 015-F — retired-architecture gate detection-quality hardening

## Surface under change

`scripts/check-retired-architecture.sh`, invoked live from
`.github/workflows/ci.yml`'s `lint` job on every pull request and push to
`main`:

- Verdict step: `bash scripts/check-retired-architecture.sh` (repo scan;
  `continue-on-error` toggled by `vars.RETIRED_ARCH_GATE_ADVISORY`, now
  fail-closed-by-default per 015.013-T).
- Integrity step: `bash scripts/check-retired-architecture.sh --self-test-integrity`
  (fixture + selection self-tests; unconditionally blocking).

Unlike a library change with no wired caller, this shipment's surface **is**
its own live runtime entrypoint — the script runs as a real GitHub Actions
step against the real repository tree on every CI invocation, not only in a
local simulation.

## Verdict: READY

The exact surface this shipment hardens executed successfully, live, in
GitHub Actions, against the real merged tree, before merge:

- PR #46, HEAD `eb6a816c742475120580df61fcbb5efe81a2d7b2`, `lint` job
  (run `34305860586`, job `102322314424`): **pass** (1m20s) — both the
  verdict step and the `--self-test-integrity` step executed as part of this
  job and did not fail.
- This is the same job/step wiring that will execute on every future PR
  against `main`, so no additional "first activation" event remains
  outstanding — the fail-closed re-arm (015.013-T) is already live and has
  already been exercised at least once (this PR's own CI run) with a clean
  repo scan.

## Compensating/prior evidence (in addition to the live CI run above)

- `bash scripts/check-retired-architecture.sh --self-test`: dual-engine TOML
  suite (14 fixtures), Go suite (19 fixtures, including the new struct-tag
  and fused-plural regression fixtures), differential suite (1 fixture) —
  all pass, including a 60-second hang-detection timeout guard proving the
  M-1 fix actually terminates.
- `go build ./...`, `go vet ./...`, `gofmt -l .`, `go test ./...` (including
  `tests/integration`) — all green, both pre-merge and re-verified on `main`
  post-merge (see the post-merge closure artifact).
- Full CI check suite on PR #46 at merge HEAD: all 13 checks pass
  (cross-compile ×4, lint, security, test, test (windows, advisory),
  pipeline-topology, ci gate, gitignore regression, detect-changes,
  load-cross-compile-targets).

## Manual checkpoints / blocked prerequisites

None. No manual checkpoint or blocked prerequisite applies — the CI
environment itself is the live runtime surface, and it already exercised
the change end-to-end before merge.

## Follow-up

None required for runtime verification. The advisory window this shipment
closes (015.013-T) is now permanently fail-closed; no further activation
event is pending.

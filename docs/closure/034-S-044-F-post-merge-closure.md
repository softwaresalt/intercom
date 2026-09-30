---
title: "Post-merge closure: 034-S / 044-F — Gate-engine Go migration M1 (gatecheck scaffold, pysem, masker, write-path)"
description: "Post-merge closure artifact for shipment 034-S / feature 044-F"
status: "complete"
tags:
  - "operational-closure"
  - "034-S"
  - "044-F"
  - "post-merge"
date: 2026-09-30
mode: post-merge
shipment: 034-S
feature: 044-F
pr: 81
merge_commit_sha: 2a334e06f1e675f69c9882e228fb15aadfb4372e
compaction_status: done
closure_status: READY
releasability: READY
conditions: []
---

# Post-merge closure: 034-S / 044-F — Gate-engine Go migration M1 (gatecheck scaffold, pysem, masker, write-path)

## Summary

Shipment `034-S` delivers `044-F`, "Scaffold gatecheck Go tool and port masker and write-path gate
(M1)". This is plan unit M1 (of M1–M4) in
`docs/plans/2026-09-28-intercom-go-gate-engine-go-migration-plan.md`, per the deliberation in
`docs/decisions/2026-09-28-intercom-go-gate-engine-python-to-go-migration-deliberation.md`
(Option B, parity-preserving port first). What it did:

* Scaffolded the root-module tool `tools/gatecheck` (stdlib plus the existing BurntSushi TOML
  dependency): `main.go` dispatch, exit-code contract, and per-sub-command stubs.
* Added `internal/pysem`, a Python-semantics compatibility layer: character classes, case mapping
  (including `Final_Sigma`), string helpers, and I/O/lookaround helpers — 14 exported helpers plus
  `ErrInvalidUTF8`, all golden-verified against the Python engine.
* Ported the rune-level masker onto `internal/gomask`, and the write-path engine (scan / self-test /
  repo-scan) onto `internal/writepath`.
* Added the shared bash runner `scripts/lib/gatecheck-run.sh` (C-3: `gatecheck_build` /
  `gatecheck_invoke` / `gatecheck_cleanup`), driven by a Go test suite
  (`tools/gatecheck/runner_test.go`) with OS-specific process-group signal tests.
* Switched `scripts/check-write-path-precondition.sh` to the Go engine. No Python interpreter
  dependency remains in that wrapper.

Behaviour is preserved: exit codes, verdicts, and self-test modes are unchanged; only the
enumerated deltas ED-1..ED-9 apply (per the plan). D-4 was honored — the masker's
`multiline-tag-shaped-raw-string` golden was pinned as-is (stash `C312BD4C` is out of scope, not
"fixed"). Parity evidence is in
`docs/plans/evidence/2026-09-28-gate-engine-go-migration/m1.md`.

Task commits:

| Task | Commit |
|---|---|
| 044.001-T | `1f11d00` |
| 044.002-T | `83d99c9` |
| 044.003-T | `83d99c9` |
| 044.004-T | `65655f7`, `86922b6`, `3b13012` |
| 044.005-T | `637f924` |
| 044.006-T | `cfe0ec9` |
| 044.007-T | `2d9562f` |
| 044.008-T | `6e130f7` |
| 044.009-T | `5a44cc6` |
| 044.010-T | `f0d2915` |
| 044.011-T | `5558dd6` |
| 044.012-T | `b352754` |
| fix-ci cycle 1 | `9df7dfb`, `de005ff` |
| docs (evidence) | `715063d` |

## Invariants to preserve

* `tools/gatecheck`'s exit contract: bash wrappers own exit 2 (usage/verdict), `main` owns exit 1 via
  panic recover; sub-command dispatch stays table-driven.
* `internal/pysem` is the single source of Python-semantics compatibility (character classes, case
  mapping, string helpers, I/O/lookaround) for every Go-ported gate engine. Future ports (M2's
  retired-arch, per the plan) must reuse it rather than re-deriving equivalents.
* The masker's `multiline-tag-shaped-raw-string` golden stays pinned to its current (pre-fix)
  behaviour until stash `C312BD4C` is explicitly triaged and implemented by Stage — it is not an
  invitation to "fix" the behaviour incidentally in a later shipment.
* `scripts/lib/gatecheck-run.sh` is the one shared bash runner for every Go-backed gate wrapper (C-3).
  The retired-arch wrapper still uses the Python `scripts/lib/gomask.py`/`retired_arch.py` engines
  until M2 — this is a documented, temporary parity bridge, not a regression.
* `git grep -n -i -E '\bpython3?\b|heredoc' -- scripts/check-write-path-precondition.sh` must stay
  empty (already true as of this shipment).

## Validator evidence

`.autoharness/workspace-profile.yaml` `runtime_validation.validator_manifest` declares one runtime
surface for this repository: `cli`, checked by a `go run ./cmd/... --help` smoke run. This shipment
does not touch `cmd/intercom` or `cmd/intercom-ctl`; that probe does not apply and no probe evidence
is recorded for it.

The surfaces that did change are `tools/gatecheck/...`, `internal/pysem`, `internal/gomask`,
`internal/writepath`, and the write-path bash wrapper + shared runner. Evidence at reviewed HEAD
`aa79b54`:

* **Go**: `gofmt -l .` (verified against an LF-normalized / `git archive`-extracted tree, since this
  checkout is CRLF) produced no output. `go vet ./...` and `go build ./...` exited 0. `go test ./...`
  (full repo suite) and `go test -race ./tools/gatecheck/...` both reported every package `ok`.
* **Parity evidence**: `docs/plans/evidence/2026-09-28-gate-engine-go-migration/m1.md` records
  byte-reproducible goldens for all three ports (masker, pysem, write-path), M1-T10's ordered
  stdout/stderr/exit-code identity across repo mode, `--self-test`, `--self-test-integrity`, and a
  bogus-flag case (zero ED-3/ED-5 delta), and masked-output SHA-256 parity across all 83 tracked
  `.go` files under `cmd/**`/`internal/**`/`scripts/testdata/**`.
* **`git grep`**: `git grep -n -i -E '\bpython3?\b|heredoc' -- scripts/check-write-path-precondition.sh`
  is empty.
* **CI**: all 14 checks were green at PR head `aa79b54`: `detect code changes`,
  `gitignore append-only + un-ignore regression (I6)`, `pipeline-topology (ambient)`,
  `merge-strategy structural verification`, `test`, `test (windows, advisory)`, `lint`, `security`,
  `load cross-compile targets`, `cross-compile (linux/amd64|windows/amd64|darwin/amd64|darwin/arm64)`,
  `ci gate`.
* **Local review readiness**: `READY` (0 P0, 0 P1) across the full 18-commit branch diff, recorded
  in the PR body's `## Local Review Readiness` block, unchanged through the merge.
* **Merge**: `gh pr merge 81 --repo softwaresalt/intercom --merge` (last-mile re-check confirmed
  `headRefOid` still `aa79b54`, all checks green, `mergeStateStatus: CLEAN` immediately before). The
  merge commit is `2a334e06f1e675f69c9882e228fb15aadfb4372e`. The repository allows merge commits
  only (`allow_merge_commit=true`, `allow_squash_merge=false`, `allow_rebase_merge=false`), which
  satisfies P-009.
* **P-018 Copilot-review gate**: `autoharness gate copilot-review 81 --repo softwaresalt/intercom
  --enforcement auto --max-wait 0` returned `NOT_APPLICABLE: PASS` (exit 0), both after CI went
  green and again at the last mile immediately before merge.
* **Merge confirmation**: `gh pr view 81` reports `MERGED` at `2026-09-30T03:55:28Z`, and
  `git merge-base --is-ancestor 2a334e06 origin/main` exits 0.

Verdict: `PASS`. No follow-ups.

## Pre-deploy audits

None apply. The shipment involves no migration, feature flag, runtime config change, or access
change — it is internal CI gate tooling.

## Deployment / rollout path

Merge-only. `scripts/check-write-path-precondition.sh` now runs the Go engine on every future PR
and every push to `main`; `tools/gatecheck` builds as part of that invocation via the new shared
runner. There is no separate deploy step.

## Post-deploy checks

* On the next few PRs, confirm the `lint` job's write-path-precondition gate and self-test steps
  still pass and still exit with the documented codes.
* Confirm no stray `tools/gatecheck` build artifacts or `__pycache__`/`*.pyc` files appear in the
  working tree after a local gate run.

## Risky action record

No destructive, irreversible, or elevated-privilege action was taken.

* **Merge**: the operator approved it under P-014 ("PR 81: Merge approved", 2026-09-30T03:54:46Z).
  It used a merge commit only and was preceded by an unconditional last-mile re-check (HEAD,
  checks, P-018, P-009) immediately before the merge command.
* **Covering-feature completion (a1, S4)**: all five conditions held. `backlogit move 044-F
  --status done` moved `044-F` from `active` to `done`; it exited 0, and a re-read showed `done`.
* **Shipment closure**: `shipment-reconcile` `classify-close-path` returned `CASCADE` /
  `FULLY_COVERED_ROOT` with binding `a36a46c5ad8ca807e2701645e33bc424dca8f6a1a8f7c1d866d0dccc9e920837`.
  The binding was recomputed from live state immediately before use and matched. Safe-close then
  delegated to the Cascade Close Sub-Procedure, which ran `backlogit shipment ship 034-S --sha
  2a334e06…`. Results:
  * `returned_ids=[]`.
  * `archived_ids`, `allowed_ids` and `required_ids` are the same set: {034-S, 044-F,
    044.001-T..044.012-T}.
  * `parent_id` was preserved on all twelve tasks.
  * `034-S` has `archived_status: shipped`; `044-F` has `archived_status: done`.

  See `.backlogit/reconcile/034-S-safe-close-20260930T040019Z.md` and the accompanying
  `034-S-classify-close-path-20260930T040019Z.md`, `034-S-pre-20260930T040019Z.md`, and
  `034-S-post-20260930T040019Z.md` reports.

## Healthy signals

* The `lint` job's write-path-precondition gate and self-test verdicts stay green on `main`.
* `tools/gatecheck` builds cleanly as part of the `go build ./...` / `go vet ./...` sequence on
  every PR.

## Failure signals

* The write-path gate's verdict or output changes on a PR that does not touch the scanned surfaces
  — that would be a behaviour-preservation regression in the Go port.
* `tools/gatecheck` fails to build, or the shared runner (`gatecheck-run.sh`) fails to invoke it,
  on any supported platform (Linux CI, Windows advisory CI).
* A future `internal/pysem` consumer (M2's retired-arch port) rediscovers a Unicode
  category-vs-derived-property gap already fixed here (see the compound capture below) —
  that would indicate the fix did not generalize as intended.

## Monitoring plan

Watch the standard CI dashboard on `main`, in particular the `lint` job's write-path-precondition
steps and the new `cross-compile`/`test`/`test (windows, advisory)` jobs' `tools/gatecheck` build,
over the next several PRs (035-S onward, and M2 when it starts).

## Rollback trigger

Revert if any of these happens:

* A false-positive or false-negative write-path verdict that the Go port caused.
* `tools/gatecheck` breaking the build or any cross-compile target.
* The shared bash runner failing on `ubuntu-latest` CI (as opposed to only the Windows advisory
  lane, where a known signal-test SKIP is expected and documented).

## Rollback procedure

Open a standard merge-commit revert PR for merge commit
`2a334e06f1e675f69c9882e228fb15aadfb4372e`. That restores the Python-heredoc-based write-path
wrapper exactly. There is no data or state migration.

For a narrower fix, `scripts/check-write-path-precondition.sh` can temporarily fall back to
invoking the pre-existing Python engine while the Go-side defect is fixed, since
`scripts/lib/gomask.py` still exists (retired by M4, not this shipment).

## Validation window

The next 2–3 PRs on `main`: 035-S onward, through the start of M2 (retired-architecture Go port,
which will be the first real consumer of `internal/pysem`/`internal/gomask` beyond write-path).

## Owner

The Ship agent and the repository maintainer (`softwaresalt/intercom`).

## Follow-ups (stash; Stage-owned triage)

No new follow-ups. Local review returned `READY` (0 P0/P1) on the first pass; the two fix-ci
commits (`9df7dfb`, `de005ff`) were both in-scope same-contract-surface completions, not scope
expansions, so no new P-021 `DEFERRED SCOPE EXPANSION` stash entries were captured during this
shipment's build/review/CI-fix cycles.

The following out-of-scope items were named in the operator's scope guards and confirmed **not**
implemented (per D-4 and the review pass): `C312BD4C` (masker multiline struct-tag semantics —
golden pinned as-is), `8E9F8E55`, `B72E9715`, `8E18CCF5`, `56B16321` (write-path entries). These
remain Stage-owned and untouched by this closure.

## Source artifact cleanup

* `044-F` has no `custom_fields.source_stash_id`. Its description cites "stash C44E2C1F; also folds
  8387758F", but Step 7 keys only on the structured field, so it skipped both entries and left them
  for Stage. Because the referenced stash (`C44E2C1F`) governs the full M1–M4 migration plan (not
  just M1), archiving it on M1's shipment alone would be premature.
* `044-F` has no `custom_fields.source_deliberation_id`. Skipped. Its references cite
  `docs/decisions/2026-09-28-intercom-go-gate-engine-python-to-go-migration-deliberation.md`, which
  is a file artifact, not a backlogit deliberation ID.
* Archived: none. Skipped: `C44E2C1F`, `8387758F` (both no structured link).

## Knowledge graduation

* No updates were needed to `docs/ARCHITECTURE.md`, `AGENTS.md`, `docs/design-docs/`, or
  `docs/product-specs/`. The change is internal CI gate tooling; neither doc currently mentions
  `gatecheck`/`pysem`/`gomask`, and the migration plan does not call for graduating those docs until
  M4 (full Python retirement), if at all.
* Compound capture:
  `docs/compound/2026-09-30-go-msys-bash-bridging-and-unicode-case-mapping-gap.md`. It covers the
  MSYS/Git-Bash path-translation pitfalls (`toBashPath()` off-by-slice, `pwd -W` +
  `os.SameFile` identity comparison), the bare-`bash`-on-PATH WSL-vs-Git-Bash ambiguity, and the
  Go `unicode.IsUpper`/`IsLower` vs CPython `Other_Uppercase`/`Other_Lowercase` derived-property gap
  — all discovered during 044.006-T/044.011-T.
* `compound-refresh`: the one related entry,
  `docs/compound/2026-09-28-extracting-bash-heredoc-python-engines-into-importable-modules.md`,
  describes 029-S's own historical extraction of the Python engines into `scripts/lib/`. It remains
  accurate as a record of that shipment and its lessons (heredoc-to-module extraction pitfalls)
  still apply to the still-Python retired-arch wrapper pending M2. It is kept, unchanged, and
  cross-referenced from the new entry.

## Compaction status (P-020)

`done`. Ship Step 6 item 8 ran `compact-context` with `target: all`, performing the bounded
per-merge Tier-1 floor.

* **Memory**: the three 034-S session checkpoints (`docs/memory/2026-09-29/034-s-gate-engine-go-migration-m1-memory.md`,
  `docs/memory/2026-09-30/034-s-gate-engine-go-migration-m1-memory.md`,
  `docs/memory/2026-09-30/034-s-gate-engine-go-migration-m1-memory-2.md`) were compacted into
  `docs/memory/compacted/2026-09-30-034-s-044-f-compacted.md`. All three originals were moved to
  `docs/archive/memory/2026-09-29/` and `docs/archive/memory/2026-09-30/` respectively.
* **Plans**: not consolidated. `docs/plans/2026-09-28-intercom-go-gate-engine-go-migration-plan.md`
  still has outstanding units (M2, M3, M4).
* **Closure records**: this pass compacted none.
* **Residual, disclosed and not actioned in this closure**: `docs/memory/` still holds 40
  uncompacted files after this pass (unchanged net count: 3 removed by this compaction, 1 new
  compacted summary added, plus this session's own final checkpoint). Several files in
  `docs/memory/2026-09-18/` and loose root-level files dated 2026-09-12 through 2026-09-28 are
  Stage-authored and predate the 14-day threshold. `docs/closure/` holds 42 records, several of
  which are also aged. As with the 029-S closure before it, this residual pool spans Stage-authored
  memory and many long-closed shipments, and is better handled by a dedicated compaction chore than
  folded into a per-shipment closure PR. That decision belongs to the operator or Stage.

## Releasability evidence

**READY.**

* Local review readiness: `READY` (P0=0, P1=0) at reviewed HEAD `aa79b54`, unchanged through the
  merge.
* All 14 CI checks were green at the merge HEAD, and the merge used a merge commit (P-009).
* P-018 `NOT_APPLICABLE: PASS`, P-016 pipeline-topology passed, and the P-014 operator approval is
  recorded ("PR 81: Merge approved", 2026-09-30T03:54:46Z).
* No runtime service surface changed. `tools/gatecheck` build, the parity evidence, and the full
  local + CI test suites are all green.
* There are no residual follow-ups from this shipment. The disclosed aged-memory-pool note above is
  advisory workspace hygiene, not a release condition.

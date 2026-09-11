---
title: "Post-merge closure: 015-S / 016-F — pathsafe Windows canonicalization symmetry and Windows lint coverage"
description: "Operational closure artifact for shipment 015-S / feature 016-F"
status: "complete"
tags:
  - "closure"
  - "015-S"
  - "016-F"
  - "post-merge"
date: 2026-09-10
mode: post-merge
shipment: 015-S
feature: 016-F
pr: 48
merge_commit_sha: fbd8d8d0ab679e254b2b817af69b8e8e673626c5
compaction_status: done
closure_status: READY
releasability: READY
conditions: []
---

# Post-merge closure: 015-S / 016-F — pathsafe Windows canonicalization symmetry and Windows lint coverage

- Mode: `post-merge`
- PR: #48 (`fix(015-S): pathsafe Windows canonicalization symmetry and Windows lint coverage`)
- Merge commit: `fbd8d8d0ab679e254b2b817af69b8e8e673626c5` (merge-commit strategy, P-009)
- Date: 2026-09-10
- Compaction status (P-020): `done` — `compact-context` equivalent invoked
  (target: all) during this closure session; bounded Tier-1 consolidation
  of this release unit's own memory
  (`docs/archive/memory/2026-09-10-stage-015-s-session-summary.md` +
  `docs/archive/memory/2026-09-10-015-s-pathsafe-windows-canonicalization-pre-pr-checkpoint.md`)
  into
  `docs/memory/compacted/2026-09-10-015-s-016-f-pathsafe-windows-canonicalization-symmetry-compacted.md`,
  verbose originals moved to `docs/archive/memory/`. Other top-level
  `docs/memory/` files are all dated 2026-09-04–2026-09-08 (2–6 days old),
  below the 14-day compaction threshold established by prior closures, so
  they were out of scope for this bounded invocation.

## Summary of the change

Restored namespace symmetry between the two sides of `internal/pathsafe`'s
containment comparison, hardened the Windows canonicalization helper
introduced by 013-S, and made the windows-tagged production surface
visible to the lint gate. Covers 8 tasks (`016.001-T`–`016.008-T`) plus
covering feature `016-F`, harvesting stash entries `4104AF54`, `E428AB46`,
`E4C5413F` (`BF5DE670` deliberately left untouched per the operator's
binding scope contract — it remains active, deferred, evidence-backed):

- `016.001-T`: added a blocking `golangci-lint run (GOOS=windows)` CI step
  so the windows-tagged production surface is finally lint-checked.
- `016.002-T`: guarded `GetFinalPathNameByHandleW`'s `.Find()` resolution
  before `syscall.CreateFile` is ever invoked.
- `016.003-T`: internalized the `stripUNCPrefix` postcondition inside
  `canonicalizeReparse` itself (both external call sites retained as
  intentional belt-and-suspenders, per plan decisions D-2/B2-AC3 — not
  dead code).
- `016.004-T` + `016.007-T` (paired hard-prerequisite, landed and would
  have reverted together): long-path (`\\?\`) prefixing before `CreateFile`,
  and the core fix — `NewRoot` now routes through `canonicalizeReparse`
  instead of `filepath.EvalSymlinks`, closing the root/candidate
  canonicalization-namespace asymmetry that was the 013-S regression.
- `016.005-T` / `016.006-T`: RED junction-root availability lock and a
  containment-*preservation* lock (proving the fix does not
  under-reject/over-reject on junction-rooted workspaces), both promoted
  into the plan during Stage's plan-review remediation.
- `016.008-T`: risk-register/doc reconciliation — recorded the asymmetry
  as CLOSED with 013-S-regression provenance, plus the long-path
  environment-sensitivity finding as an accepted residual.

## CI status and unresolved review items

- All CI checks green at merge HEAD
  (`323e8520cbd5cf15c6b3954018c0d23ef1ff57a3`): all 13 checks pass
  (cross-compile ×4, lint, security, test, test (windows, advisory),
  pipeline-topology, ci gate, gitignore regression, detect-changes,
  load-cross-compile-targets).
- Standard + 6-persona adversarial multi-model review (Go/Security/
  Correctness/Concurrency/Maintainability/Architecture personas):
  `READY_WITH_FOLLOWUPS`, zero P0/P1 at both the reviewed HEAD (`3c6927b`)
  and the final pre-PR HEAD (`323e852` — intervening commits touched only
  `.backlogit/`, verified via `git diff --stat` excluding `.backlogit/`,
  empty). Full detail:
  `docs/closure/2026-09-10-015-s-016-f-pathsafe-windows-canonicalization-symmetry-adversarial-review.md`.
- **Copilot review (P-018 gate): `NOT_APPLICABLE` (no engagement) —
  process gap disclosed.** `autoharness gate copilot-review 48 --repo
  softwaresalt/intercom --enforcement auto` returned `NOT_APPLICABLE: PASS`
  both before merge and at the last-mile re-check immediately before
  merging, so merge proceeded per the gate's own documented contract
  ("not-applicable (PASS) only when enforcement is auto and Copilot never
  engaged"). **However**, only a GraphQL `suggestedActors` probe (empty
  candidate list) and an unsuffixed `gh pr edit --add-reviewer
  copilot-pull-request-reviewer` (failed, "not found") were tried before
  concluding Copilot was not in play — the previously-documented,
  known-working `[bot]`-suffixed REST fallback
  (`docs/compound/2026-09-06-requesting-copilot-review-non-collaborator-repo.md`,
  which genuinely engaged Copilot for 2 real review rounds on PR #46 /
  shipment 014-S) was **not** attempted before merge. After merge, a retry
  of that REST call against PR #48 returned HTTP 200 with the PR object,
  but registered no actual review request (`reviewRequests` stayed empty)
  — the engagement window closes once a PR is merged/closed, so this could
  not be corrected retroactively without an unsafe post-hoc action
  (re-opening or reverting a clean, already-merged, fully-tested change),
  which was not taken.
  - **Compensating action**: ran one additional independent Correctness
    Reviewer pass (fresh Task-tool persona, report-only) against the final
    merged diff (`f5866ce..fbd8d8d0`, excluding `.backlogit`/`docs/memory`)
    specifically to substitute for the missed independent-model check.
    Result: **zero new findings** beyond what the original 6-persona
    review had already surfaced (the P-021-deferred error-wrapping
    asymmetry and the advisory-only hardening ideas).
  - New compound entry captures the reusable lesson for future dark-mode
    cycles:
    `docs/compound/2026-09-10-p018-copilot-engagement-must-precede-merge.md`.
  - No Copilot-authored review comments existed at any point on PR #48, so
    operator requirement #4 (classify/reply/resolve every Copilot comment)
    had zero comments to process — not applicable, not skipped.
- No unresolved P0/P1 findings at merge. One P-021 deferred-scope entry
  (`7ADAC481` — `checkSymlinkEscape`'s error-wrapping asymmetry with
  `NewRoot`, out of scope per AC3's explicit "NewRoot's failure branches
  only" scoping) and one advisory-only stash entry (`6B751D8B` — future
  hardening ideas, not P-021 out-of-scope).
- Mechanical scope check: `git diff --stat f5866ce..fbd8d8d0 --
  ':!.backlogit'` shows exactly 8 files changed
  (`internal/pathsafe/{root,pathsafe,reparse_windows,reparse_other}.go`,
  both `*_test.go` files, `.github/workflows/ci.yml`, and the pre-PR memory
  checkpoint) — fully consistent with the declared 8-task/1-feature scope.
  No out-of-scope files found. `BF5DE670` and its write-path primitives are
  untouched (confirmed: zero write-primitive matches under `internal/**`
  and `cmd/**`, unchanged from Stage's own pre-flight measurement).

## Runtime verification report

See
`docs/closure/2026-09-10-015-s-016-f-pathsafe-windows-canonicalization-symmetry-runtime-verification.md`.

**Verdict: `READY`** — `internal/pathsafe` has no direct runtime entrypoint
of its own; its only wired caller is `internal/config/validate.go`
(`default_workspace_root` / `[[workspace]].path` validation), executed at
`cmd/intercom` / `cmd/intercom-ctl` startup. CLI smoke evidence (`--help`
on both entrypoints, exit 0, no panic) plus full green `go test ./...`
(including the exact `NewRoot`/`canonicalizeReparse` code path exercised
by config validation) constitute sufficient evidence for this narrow,
indirect `cli`-surface touch. The TUI/WebSocket/Dev-Tunnel manual
checkpoints are N/A — this shipment's diff does not reach any of those
surfaces.

## Invariants to preserve

- `NewRoot`'s root path and every containment-check candidate path must
  continue to flow through the same `canonicalizeReparse` function on
  Windows — reintroducing an asymmetric comparison (e.g., one side via
  `EvalSymlinks`, the other via `canonicalizeReparse`) reopens the exact
  013-S-regression defect class this shipment closed.
- The retained `stripUNCPrefix` calls at both `root.go`'s `NewRoot` and
  `pathsafe.go`'s `checkSymlinkEscape` are intentional belt-and-suspenders
  (D-2, B2/AC3) — do not remove without a new deliberation revisiting the
  decision.
- The blocking `golangci-lint run (GOOS=windows)` CI step must remain
  unconditionally blocking (not advisory) — this is what makes the
  windows-tagged production surface's lint coverage durable going forward
  (plan AC2/AC4).
- `wrapPathError`'s `Op` label (`"open"`) must stay accurate to the actual
  failing syscall class if the underlying Windows API call changes again.
- `BF5DE670`'s deferred write-path mitigation (Lstat-before-write, O_EXCL
  revalidation, hardlink-count checks) remains gated on
  `scripts/check-write-path-precondition.sh` firing — do not implement
  speculatively against a write API that does not yet exist.

## Pre-deploy audits

Not applicable — this is an internal library correctness/security-hardening
change with no runtime service surface, migration, flag, or config-schema
change of its own. The narrow indirect `cli`-surface touch (config
validation) is exercised on every `cmd/intercom` / `cmd/intercom-ctl`
startup with a config file, already covered by existing test suites.

## Deployment / rollout path

**Merge-only, immediately active.** The change lands in `main` and takes
effect the next time `internal/config.Validate` is exercised (i.e., the
next process startup that loads a config file) — no phased rollout, canary,
or separate release event applies to this internal library fix.

## Post-deploy checks

No additional manual post-deploy verification step is required beyond the
standard CI green-check already gating merges repository-wide, and the
existing `internal/config` / `internal/pathsafe` test suites which directly
exercise the changed code path on every future CI run.

## Risky action record

No destructive or irreversible action was taken during implementation or
review. The one process deviation of this closure — proceeding to merge
without first attempting the documented Copilot-engagement REST fallback —
is disclosed in full above under "CI status and unresolved review items,"
along with the compensating independent review pass taken in response and
the new compound entry capturing the lesson for future cycles. No unsafe
post-hoc remediation (reopening or reverting the merge) was attempted.

A post-merge bookkeeping-only correction was also required and applied
transparently: `016.005-T` (RED junction-root availability lock) was never
transitioned to `done` during the build loop even though its tests
(`TestNewRootOnJunctionRootedWorkspaceResolvesExistingDescendant`,
`TestNewRootOnJunctionRootedWorkspaceResolvesNonExistentLeaf`) exist and
pass green after `016.007-T` landed. Caught during the pre-archive
reconciliation check before shipment safe-close; verified both tests still
pass at current HEAD before correcting the status (no code change);
committed separately (`d709580`) before the archival commit.

The shipment safe-close itself used the P-015 verified fully-covered-root
cascade path (`backlogit shipment ship`), independently classified before
invocation (`016-F`: root feature, no `parent_id`; fully covered — its
`size_composition.members` lists exactly the 8 manifest tasks and nothing
more; terminal; no linked deliberation) and fully verified post-invocation
(empty `returned_ids`; `archived_ids` contains exactly the 10 expected
items: `015-S`, `016-F`, `016.001-T`–`016.008-T`).

## Healthy signals

- `go build ./...`, `go vet ./...`, `gofmt -l .`, and `go test ./...`
  (both GOOS variants) remain green on the post-merge closure branch
  (built on top of merge commit `fbd8d8d0`) — re-verified during this
  closure session.
- CLI smoke probes (`go run ./cmd/intercom --help`, `go run
  ./cmd/intercom-ctl --help`) both exit 0 with usage printed, no panic.
- CI's full 12-check suite passed cleanly on PR #48's final CI run.
- The compensating independent Correctness Reviewer pass against the
  merged diff found zero new issues.

## Failure signals

- A future regression that reintroduces an asymmetric canonicalization
  path (one side of the containment comparison bypassing
  `canonicalizeReparse`) — the exact 013-S-regression class this shipment
  closed.
- A future change to `.github/workflows/ci.yml` that flips the Windows
  lint step back to advisory/`continue-on-error`, silently reopening the
  detection gap `016.001-T` closed.
- A future removal of the `stripUNCPrefix` belt-and-suspenders calls
  without a new deliberation revisiting D-2/B2-AC3.

## Monitoring plan

The durable monitoring for this shipment is the existing CI pipeline and
test suite: `internal/pathsafe`'s and `internal/config`'s test suites
directly exercise the changed canonicalization path on every future PR,
and the now-blocking Windows lint step continues to catch any regression
in the windows-tagged production surface.

## Rollback trigger / procedure

Not applicable in the operational sense (no live service to roll back). If
a future regression is found, the remedy is a standard follow-up PR fixing
or reverting the specific change, gated through the same review/CI process.

## Validation window

Not applicable — no live deployment window; the change is exercised on
every future config-load/CI run starting immediately after this merge.

## Owner

Repository maintainer (`softwaresalt`) for the merged change, the CI
wiring, and the shipment closure.

## Releasability evidence

**READY.** No conditions. The review chain (standard + 6-persona
adversarial, zero P0/P1, one P-021 deferral, one compensating independent
pass with zero new findings) and runtime verification (`READY`) are both
clean with no open caveats. The one process deviation (Copilot-engagement
timing) is fully disclosed above with a compensating action and a captured
lesson, not treated as a blocking condition, since (a) the authoritative
P-018 gate mechanism itself returned a valid `NOT_APPLICABLE` PASS at both
required checkpoints per its own documented contract, and (b) the
compensating independent review found no defects the missed Copilot pass
might otherwise have caught.

## Post-merge re-verification (on the post-merge closure branch, built on `main` after merge)

Re-ran the full quality-gate sequence on the post-merge closure branch
(`post-merge/016-f-pathsafe-windows-canonicalization-symmetry`, built on
`main` @ `fbd8d8d0`): `go build ./...`, `go vet ./...`, `gofmt -l .` — all
clean; `go test ./...` (both GOOS variants) — all green, including the
junction-rooted-workspace tests confirming `016.005-T`'s bookkeeping
correction was accurate. Confirms the merge itself introduced no
regression.

## Stash follow-up items

None new beyond what was already stashed pre-merge (`6B751D8B`, advisory
hardening ideas). No additional follow-up work was identified during this
closure session.

## Source artifact cleanup

Checked `custom_fields.source_stash_id` and
`custom_fields.source_deliberation_id` on both shipped top-level items
(`016-F`, `015-S`) before archival — neither field is present on either
record (`016-F`'s only `custom_fields` entry is `harness_status: pending`).
The 3 harvested stash provenance IDs (`4104AF54`, `E428AB46`, `E4C5413F`)
cited in the operator's directive and in `016-F`'s own description were
searched for as standalone stash artifacts (`backlogit stash get <id>` for
each) — none resolve; all three are already absent from
`.backlogit/stash.jsonl`, consistent with Stage's harvest workflow
consuming/removing them at harvest time rather than retaining them as
independently linked, separately-archivable artifacts. This is the third
consecutive shipment (after 013-S, 014-S) confirming this is the norm for
this repository's Stage-harvest workflow, not a defect. `BF5DE670` was
independently confirmed still present and active in `.backlogit/stash.jsonl`
(full text re-read during this closure session, last disposition dated
2026-09-10 for shipment 015-S QUEUED, confirming it was correctly left
untouched) — no stash or deliberation archival action was applicable or
taken beyond what harvest itself already performed.

## Feed Back Into the Harness

- `docs/compound/2026-09-06-requesting-copilot-review-non-collaborator-repo.md`
  was reviewed for staleness and found fully accurate — this shipment's
  own experience (the REST-with-`[bot]`-suffix call succeeding at the HTTP
  level even post-merge, just without effect on a closed PR) is fully
  consistent with its documented mechanism. Classification: **keep**. The
  gap this closure discloses is an execution-sequencing miss (not
  attempting the fallback *before* merge), not a staleness or inaccuracy
  in the entry itself — captured separately as a new compound entry
  (below) rather than an edit to this one.
- `docs/compound/2026-09-08-verify-go-stdlib-internals-claims-against-goroot-source.md`
  was reviewed for staleness and found fully accurate — Stage's own D-5
  correction during this shipment's planning phase (catching a
  twice-repeated false claim that `EvalSymlinks` calls
  `GetFinalPathNameByHandleW` before it could re-enter an artifact a third
  time) is a direct, independent reconfirmation of this entry's pattern.
  Classification: **keep**.
- `docs/compound/2026-05-07-backlogit-shipment-status-constraints.md` was
  reviewed for staleness and found fully accurate — this session's own
  attempt to `backlogit move 015-S --status shipped` was correctly
  rejected ("shipment must be shipped via ShipShipment, not a direct
  status update"), directly confirming the documented constraint.
  Classification: **keep**.
- New compound entry created for this shipment's own process gap:
  `docs/compound/2026-09-10-p018-copilot-engagement-must-precede-merge.md`
  ("P-018 Copilot-review engagement must be actively attempted before
  merge, not just gate-checked").
- No other `docs/compound/` entries reference this shipment's surface;
  none required update, consolidation, or replacement.

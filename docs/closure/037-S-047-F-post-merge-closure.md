---
title: "Post-merge closure: 037-S / 047-F — Gate-engine Go migration M4 (retire gate-engine Python)"
description: "Post-merge closure artifact for shipment 037-S / feature 047-F"
status: "complete"
tags:
  - "operational-closure"
  - "037-S"
  - "047-F"
  - "post-merge"
date: 2026-10-01
mode: post-merge
shipment: 037-S
feature: 047-F
pr: 87
merge_commit_sha: ec6d1d9938559292363508134ddbf462a62e564d
compaction_status: done
closure_status: READY
releasability: READY
conditions: []
---

# Post-merge closure: 037-S / 047-F — Gate-engine Go migration M4 (retire gate-engine Python)

## Summary

Shipment `037-S` delivers `047-F`, "Retire gate-engine Python, CI steps and docs (M4)". This is
the last plan unit (M4 of M1–M4) in
`docs/plans/2026-09-28-intercom-go-gate-engine-go-migration-plan.md`, per the deliberation in
`docs/decisions/2026-09-28-intercom-go-gate-engine-python-to-go-migration-deliberation.md`. It
follows M3 (036-S / 046-F, PR #85). With M4 merged, every blocking CI gate runs on the Go
`tools/gatecheck` engine. The only Python left in CI is the pinned autoharness install in
`topology-check`. What it did:

* Mapped all 38 assertions in the retiring Python suite to Go (R-13 coverage map, M4-T1). The
  final totals are 30 (a) direct Go equivalents and 8 (b) no longer applicable, with no assertion
  lost.
* Ported the retained CI-wiring and wrapper-text assertions to Go (M4-T2). Added red-first
  retirement assertions (M4-T3) that fail if gate-engine Python, a Python CI step, or an
  interpreter call comes back.
* Recorded the M4-T4 retirement approval before any removal commit. It is backed by the P-017
  dark-mode activation record for 037-S (evidence §5).
* Removed the gate-engine Python steps from the `lint` job (M4-T5) and deleted the retired Python
  modules and tests (M4-T6).
* Rewrote the wrapper headers and gatecheck provenance comments to describe the Go engines
  (M4-T7, M4-T8).
* Recorded consolidated per-switch parity (M4-T9). Across the four migrated gates, the M4 pair
  (`f221023` against `b51d1f0`) is byte-identical: 15 of 15 exit codes and 30 of 30 streams
  match (evidence §8.2).
* All 9 tasks (`047.001-T`..`047.009-T`) have passing acceptance criteria. Evidence is in
  `docs/plans/evidence/2026-09-28-gate-engine-go-migration/m4.md`.

The diff covers 69 files, with 5,941 insertions and 1,822 deletions. `cmd/` is untouched.

| Commit | Description |
|---|---|
| `c50971f` | chore(ci): claim shipment 037-S |
| `cfb8852` | docs(docs): record M4-T1 R-13 coverage map |
| `87d4970` | test(ci): port retained CI-wiring and wrapper-text assertions to Go (M4-T2) |
| `f8b346a` | test(ci): add red-first gate-engine Python retirement assertions (M4-T3) |
| `939e00d` | docs(docs): record M4-T4 retirement approval before removal |
| `d2430cf` | ci(ci): remove gate-engine Python steps from lint job (M4-T5) |
| `d1509c9` | chore(ci): delete retired gate-engine Python modules and tests (M4-T6) |
| `b022276`, `d748a91`, `b51d1f0` | docs(ci): wrapper header and provenance rewrites (M4-T7, M4-T8) |
| `e4924f3` | docs(docs): record M4-T7..T9 rewrite and consolidated parity evidence |
| `d9eefab`, `99b8d00` | Pre-PR adversarial review remediation |
| `75f107e`..`5ff8b19` | Copilot review remediation, rounds 1–22 (22 commits; see evidence §10) |

## Invariants to preserve

* Every blocking gate keeps its verdicts and exit codes. M1–M3 established parity; M4 only
  removes code that was already dead.
* No Python interpreter, Python CI step, or `scripts/lib` Python module comes back.
  `TestCIWiringRetire_*` and the wrapper-text guards enforce this and fail closed.
* The pinned, hash-locked autoharness install in `topology-check` stays exactly as pinned.

## Validator evidence

* Locally, `go build ./...`, `go vet ./...` and `go test ./...` pass, as do the four migrated
  wrappers' self-test modes, and `actionlint` is clean (evidence §6 and §8).
* CI run URLs, which evidence §8.5 defers to this record. Both concluded `success`:

  | Unit | PR | PR-head CI | First post-merge `main` CI |
  |---|---|---|---|
  | M4 | #87 (merge `ec6d1d9`) | https://github.com/softwaresalt/intercom/actions/runs/36904296835 (head `5ff8b19`) | https://github.com/softwaresalt/intercom/actions/runs/36905546227 |

* The Windows test job is advisory. Linux CI is authoritative.
* No CLI or runtime probe applies, because `cmd/` is untouched and no service surface changed.

## Pre-deploy audits

None required. The change is CI-only and reverts in a single step.

## Deployment / rollout path

Merge only. The change takes effect on the next CI run on `main`.

## Post-deploy checks

* The `lint`, `gitignore-append-only`, `topology-check`, `merge-strategy` and `ci-gate` jobs
  stay green on `main`.
* No job attempts to invoke Python other than the pinned autoharness install.

## Risky action record

No data loss or elevated-privilege action was taken.

* **Code deletion (M4-T6)**: the retired Python modules and tests were deleted only after the
  R-13 coverage map (M4-T1) and the M4-T4 approval record. Approval came from the P-017
  dark-mode activation for 037-S. No `agent-intercom` broadcast was possible because the pack is
  not configured. This deviation is disclosed in evidence §5 and §9. The deleted files remain in
  git history.
* **Merge**: performed under P-017 with `merge_approval_pre_authorized: true` (operator AFK,
  scoped to 037-S and its closure PR). The merge used a merge commit. It was preceded by the
  P-018 `autoharness gate copilot-review` gate (`SATISFIED`) and a green `gh pr checks`.
  `admin_fallback_pre_authorized: false`; admin fallback was never needed.
* **Copilot review rounds beyond the 3-cycle limit (disclosed)**: PR #87 went through 23 Copilot
  reviews. Rounds 1–22 raised findings, either as one of the 18 inline threads or in the review
  overview text; round 23 was clean. That exceeds the circuit breaker's 3-cycle review-fix
  limit. Each finding passed the P-021 C1 same-contract-surface test, being
  a gap in M4's own new guards or evidence. The operator's activation instruction explicitly
  required waiting for Copilot and resolving every comment before merging, so continuing was the
  disposition the operator had already chosen. All 18 review threads were replied to after the
  fix was pushed, then resolved through GraphQL. None remain unresolved.
* **Task status lapse (disclosed)**: the 9 tasks stayed `active` throughout the build. They were
  moved to `done` only at closure, just before reconciliation, instead of as each task finished.
  The end state is correct. Each task's acceptance criteria were met and are evidenced in m4.md
  before the merge. The lapse affected only backlog visibility during the run.
* **Covering-feature completion (a1)**: all conditions held: every manifest descendant was
  done, no descendant was live, the descendant set equalled the manifest, and the shipment
  status was `active`. `backlogit move 047-F --status done` exited 0, and a re-read confirmed
  `done`.
* **Shipment closure: binding revalidation performed, but not tool-enforced (residual
  disclosed)**: `mode: classify-close-path` was run manually, because this workspace has no CLI.
  It returned `CASCADE` / `FULLY_COVERED_ROOT` with `CLASSIFICATION_BINDING`
  `0f9be8a39309d70cd01b641871f84de68fc0abe5a80bce03fe394a0c230792d9` at 18:20:09Z.
  The 035-S and 036-S closures delegated to the cascade primitive without the safe-close Step 0
  recompute-and-compare. This closure performed that step: it took a fresh snapshot at
  18:20:12Z, recomputed the binding independently, and found an identical value. The cascade
  ran only after that match. What remains is the gap Stage holds as stash `D10D3AFC`:
  * There is still no invokable `safe-close`.
  * The recompute and the comparison were agent-executed by a session-local helper, so they
    are not a tool-enforced atomic call.
  * Under two minutes separated the fresh snapshot (18:20:12Z) from the mutation. The mutation
    landed at 18:20:42Z per artifact `updated_at`. backlogit's slog clock reads about one
    minute later; both postdate the snapshot. One intervening attempt failed during argument
    parsing with `unknown flag: --json` and changed nothing. This was a single-agent session
    with no concurrent backlog writers.

  `backlogit shipment ship 037-S --sha ec6d1d99…` exited 0:
  * `archived_ids` equals `allowed_ids` and `required_ids`: the 11 items `{037-S, 047-F,
    047.001-T..047.009-T}`.
  * `returned_ids=[]`.
  * `parent_id: 047-F` is preserved on all 9 tasks.
  * `037-S` has `archived_status: shipped`, and `047-F` has `archived_status: done`.

  See `.backlogit/reconcile/037-S-pre-2026-10-01T18-19-59Z.md`,
  `037-S-classify-close-path-2026-10-01T18-20-09Z.md`,
  `037-S-cascade-close-2026-10-01T18-21-40Z.md`, and `037-S-post-2026-10-01T18-23-04Z.md`.
* **P-007 archive-integrity guard**: `git status --short -- ".backlogit/archive/"` showed no
  deletions.

## Healthy signals

* All blocking CI jobs and `ci-gate` stay green on `main`.
* `tools/gatecheck` builds, and its retirement and CI-wiring guard tests pass on every PR.

## Failure signals

* A retirement guard (`TestCIWiringRetire_*`, the wrapper-text guards, or the mask-flow guards)
  goes red on a PR that did not add Python. That would mean a false positive in the textual
  YAML scanning.
* A gate's verdict changes on a PR that does not touch its scanned surface.

## Monitoring plan

Watch the standard CI dashboard on `main` over the next several PRs. Pay particular attention
to any PR that edits `.github/workflows/ci.yml`. The retire guards fail closed on YAML forms they
cannot parse: flow mappings, aliases, explicit keys, tags and anchors.

## Rollback trigger

* The retirement guards block a legitimate workflow change and cannot be satisfied with an
  equivalent block-style YAML form.
* A gate verdict regresses because of a removed step.

## Rollback procedure

Open a standard merge-commit revert PR for `ec6d1d9938559292363508134ddbf462a62e564d`. That
restores the Python modules, tests and CI steps exactly. There is no data or state migration.

## Validation window

The next 2–3 PRs on `main`, especially the first one that edits `ci.yml`.

## Owner

The Ship agent and the repository maintainer (`softwaresalt/intercom`).

## Follow-ups (Stage-owned triage)

* Stash `44F8CC48` (low priority; captured during pre-PR adversarial review per P-021 C1/C2):
  `.gitignore` lines 125–126 still mention gate engines under `scripts/lib/`. The file is
  append-only and gated, so it was not edited in this shipment.
* Stash `C44E2C1F` (the M1–M4 migration umbrella) was already archived as consumed by Stage on
  2026-09-28, when it harvested M1–M4 (`.backlogit/archive/stash.jsonl`). It needs no action.
* The held shipments `030-S`, `031-S` and `032-S` were held pending this migration and now need
  re-planning onto Go, per 047-F.
* The masker multiline-tag semantics stash entry (`C312BD4C`) is marked "decide after 037-S (M4) ships,
  alongside the 034-F re-plan". That decision is now unblocked.
* Earlier follow-ups are unchanged and still with Stage: `7223218F`, `5A8EC1BC`, `978D2946`,
  `50E6F22C` and `D10D3AFC` (the invokable `safe-close`).

The Windows-only `tools/gatecheck` `runner_test.go` `TestGatecheckCleanup_*` failures are a
known, environment-specific issue and not a regression. Linux CI is authoritative.

## Source artifact cleanup

* `047-F` has no `custom_fields.source_stash_id`, so Step 7 skipped it. Its description cites
  `C44E2C1F`, which Stage had already archived on 2026-09-28.
* `047-F` has no `custom_fields.source_deliberation_id`, so it was skipped. Its references cite
  the migration deliberation as a file artifact.
* Archived: none.

## Knowledge graduation

* `docs/ARCHITECTURE.md`, `AGENTS.md`, `docs/design-docs/` and `docs/product-specs/` need no
  updates. The change retires internal CI tooling, and the gate docs were updated in the PR.
* For compound-refresh, see `docs/closure/2026-10-01-037-s-compound-refresh.md`.

## Compaction status (P-020)

**`done`.** The mandatory post-merge `compact-context` call ran with `target: all`. It produced
`docs/memory/compacted/2026-10-01-037-s-047-f-compacted.md` from two sources:

* the 037-S session memory
* the C44E2C1F migration staging memory (`2026-09-28-stage-c44e2c1f-python-to-go-migration-staging.md`)

The staging memory qualified under the completed-work rule, because all four migration shipments
have now shipped. The originals were moved to `docs/archive/memory/` and can be traced through
the compacted file's `compacted_from` field. The migration plan stays in place as the record of
the completed M1–M4 program.

## Releasability evidence

**READY.**

* Local review readiness was `READY_WITH_FOLLOWUPS` (P0=0, P1=0) at final reviewed HEAD
  `5ff8b19`, unchanged through the merge.
* A pre-PR adversarial review ran with four reviewers: an anchor plus Tier 1, Tier 2 and Tier 3.
  It found no P0 or P1 issues. The remediation is recorded in evidence §9.
* At merge:
  * Copilot round 23 was clean.
  * P-018 `autoharness gate copilot-review` returned `SATISFIED`, with no unresolved threads.
  * All CI checks were green.
  * The merge used a merge commit (P-009).
  * P-016 topology passed.
  * P-014 approval came from the DARK_MODE_ACTIVE `merge_approval_pre_authorized: true` record
    for 037-S.
* The follow-ups above are Stage-owned triage items, not release conditions.

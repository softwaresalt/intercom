---
title: "Stage session: Gatecheck Batch B1 write-path input hardening (shipment 042-S queued)"
date: 2026-10-10
agent: stage
status: complete
branch: chore/stage-gatecheck-batch-b1-write-path-input-hardening
base: "origin/main@1b130a2"
deliberation: docs/decisions/2026-10-10-intercom-go-gatecheck-batch-b1-write-path-input-hardening-deliberation.md
plan: docs/plans/2026-10-10-intercom-go-gatecheck-batch-b1-write-path-input-hardening-plan.md
feature_id: 052-F
shipment_id: 042-S
---

# Stage session memory: Gatecheck Batch B1 write-path input hardening

## Status

**Complete. Shipment 042-S is queued** and covers feature 052-F plus five tasks. The next step
belongs to the operator (push the Stage branch, open and approve the merge-commit staging PR) and
then to Ship (claim 042-S after the staging PR is merged).

## Session inputs

* **Orchestrator `stage next`, batch B1.** Sequential, non-dark. Route: `claude-sonnet-5.5` /
  anthropic / xhigh (`config.model_routing.stage`).
* **Batch B1 (operator-confirmed):** exactly B83F53BB (high, bug) and 05E12A6F (low, task). Both
  carry the literal DEFERRED SCOPE EXPANSION marker, so P-021 C6 forced the `deliberate` route.
* **Explicitly excluded:** the 33 other write-path entries, retiredarch entries (DC921AF6,
  3750C37C, 0ECC1895 held by D-RA-7; 4A3851F0, A85D25F7, D17D5C5A, 31F33EFE, 4995F8C3, C8827920),
  5A8EC1BC (note-only), and every unignore, CI/script and docs entry.

## Steps

| Step | Outcome |
|---|---|
| 0.0 Tool gate | `TOOL_DEGRADED`: no MCP tools are exposed in this session; every backlogit operation used the registry-declared CLI fallback (backlogit v1.11.0). agent-engram and graphtor-docs MCP tools are not exposed either, so code context came from direct file reads and `rg`. agent-intercom is not installed (no broadcasts) |
| 0.1 Index sync | `INDEX_SYNC_OK` (CLI) at start; end-of-session sync run before the summary |
| Recovery | 71 checkpoints, all `resolved`, zero anomalies; zero-candidate normal startup |
| Hook events | 81 historical lifecycle events, no `feature_review_ready` or `blocked_stale` signal; acked through seq 1137 |
| 1 Triage | Both entries task-shaped with the DEFERRED marker: route deliberate. (A) Duplicate scan CLEAN over 74 active entries. (B) PR #116 recovered for both (Ship residual records cite them as reused deferrals); review-thread `N/A` stands (PR #116 has 3 threads, none about these findings) |
| 1.5 Grouping | Operator-selected single group; no alternative groupings generated (two entries) |
| 1.8 Learnings | learnings-researcher: confidence medium (no entry covers writepath `-z` or the fixture self-test directly) |
| 1.9 Branch gate | `STAGE_BRANCH_GATE_OK`: one worktree; base `main` resolved from `origin/HEAD`; branch created from `origin/main` @ `1b130a2` (2 docs-only commits ahead of the Orchestrator's `3f92d7e`); upstream tracking unset so a bare push cannot target `main` |
| 2 Deliberation | Decisions D-BW-1, D-BW-1a, D-BW-2, D-BW-3, D-BW-4, D-BW-5 |
| 3 Plan | impl-plan written; plan-harden run (3 signals present, strict-safety action record) |
| 4 Plan review | Attempt 1: **FAIL** (3 P1). Attempt 2: **ADVISORY** (no P0/P1; every P2 resolved in the plan text). `dispatch_mode: multi-agent`; anchor route `gpt-6.1-sol` dispatchable. ADVISORY authorization recorded in the plan (relayed `stage next` instruction) |
| 5 Harvest | Feature 052-F and tasks 052.001-T to 052.005-T; dependency edges recorded; sizes set as structured fields; complexity as prose |
| 5.5 Shipment | 042-S created with 052-F first, then the five tasks in dependency order; manifest read back (6 items) |
| 5.6 Archive | B83F53BB and 05E12A6F archived with a promotion note |

## Premise verification (against origin/main @ 1b130a2)

Both premises HOLD. Stash line numbers had drifted (the plan carries the current ones). The
staging probe (git 2.55, scratch repo under `logs/`, deleted) showed: git C-quotes `internal/e-acute.go`;
`GIT_INDEX_FILE` to a partial index yields a non-empty partial listing, so ED-7 does not fire
(**false-clean**); `GIT_LITERAL_PATHSPECS`, `GIT_DIR`, malformed config vectors fail closed. This
host's own session exports `GIT_CONFIG_COUNT=3` and `KEY_0..2`. 051.004-T had already added
`containedRegularFile` to writepath, which U4 reuses. No pin refreeze is needed: the retiredarch
writepath pin inspects only `scanFile`/`scanSource`. One frozen file must change: the oracle
`writepath_oracle_test.go` decodes the real runner output (authorized as PA-2 by the staging-PR
approval).

## Decisions (Stage-recommended; operator override at staging-PR review)

* **D-BW-1:** `-z` with a strict unexported NUL-record helper (`scannedPathsFromListing`).
* **D-BW-1a:** UTF-8 validity and control/separator runes are judged only for paths that will be
  scanned (diverges from retiredarch until `31F33EFE` is decided).
* **D-BW-2:** in-package copy of `gitRunnerEnv`; no relative-`git` refusal; three env vectors.
* **D-BW-3:** reuse `containedRegularFile` per fixture (leaf and ancestors).
* **D-BW-4:** prep task, single authorized oracle edit, no pin refreeze.
* **D-BW-5:** one feature, five tasks, one shipment.

## Backlog created

| ID | Title | Size / complexity | Depends on |
|---|---|---|---|
| 052-F | Gatecheck Batch B1 write-path input hardening | | |
| 052.001-T | Add listing test helpers and adapt oracle decode | XS / trivial | |
| 052.002-T | Switch writepath git listing to NUL-safe selection | M / medium | 052.001-T |
| 052.003-T | Isolate git environment for writepath ls-files runner | S / medium | 052.002-T |
| 052.004-T | Refuse linked fixtures in writepath self-test | S / low | |
| 052.005-T | Correct stale writepath listing comment in retiredarch | XS / trivial | 052.002-T |
| 042-S | shipment (queued) | | |

Degradation flagged: the workspace task type defines no `complexity` field (probe confirmed:
`artifact type "task" does not define a complexity field`), so complexity is recorded as labelled
prose in implementation-notes. `size` is structured (`size_source: agent`,
`size_ruleset_version: 2h-rule-v1`).

## Stash dispositions

* **B83F53BB, 05E12A6F:** consumed and archived (promotion targets recorded in the text).
* **New DEFERRED SCOPE EXPANSION entries (active):**
  * `4372BAD4` (low): writepath relative-`git` path refusal (retiredarch parity).
  * `DBE25DF5` (medium, bug): wrapper `ROOT` derivation is unscrubbed and ED-7 detects only an empty
    selection.
  * `EDC18D59` (low): pre-existing CI-log echo sinks (fixture names, `errorLine` tail, `ReadText`
    path).
* **5A8EC1BC:** retained, deferred; evidence-only note appended (allowed by the operator).
* **31F33EFE, C8827920, 4995F8C3:** retained untouched.
* All other active entries untouched.

## Deviations and disclosures

* **Compound learning captured:** `docs/compound/workflow-issues/staging-verification-commands-and-ambient-git-env-2026-10-10.md`
  (unprobed verification commands without the mandatory `--root`; ambient `GIT_*` skews git-env
  probes; an empty-selection guard is not a completeness guard).
* **Continuous-learning observe/learn skipped:** the observation store
  `.autoharness/continuous-learning/` is a tracked path outside Stage's artifact roots (the Step 1.9
  ownership discriminator and the Role Boundary), so Stage did not write to it; the recurring patterns
  are in the compound document above.
* **Principle IV (minor):** early in the session, a scratch copy of the HEAD `stash.jsonl` blob was
  written to the OS temp directory (`%TEMP%\head-stash.jsonl`) while checking that the working copy was
  byte-identical to HEAD. It contains only tracked, non-secret content. It was left in place (a delete
  outside the workspace would be a second deviation); the operator may remove it. All other scratch work
  used the git-ignored `logs/` directory and was deleted (`logs/scratch-b1-probe`), plus
  `logs/stage-b1/` (empty).
* **Stat-only `M .backlogit/stash.jsonl`:** the working copy was byte-identical to HEAD (SHA-256
  verified); `git add -- .backlogit/stash.jsonl` refreshed the index stat entry so the Step 1.9 clean
  worktree check could pass. No content changed.
* **Plan edited after the first re-review:** the attempt-2 recommendations were applied in place
  (text-only; no new design). An ADVISORY verdict does not require a third review.

## Operator actions still required

1. Review the staging PR content, including the Stage-recommended options (D-BW-1a, D-BW-2), the
   exceptions EX-1 to EX-5, and **PA-2** (the single authorized edit to the header-frozen oracle).
2. Push `chore/stage-gatecheck-batch-b1-write-path-input-hardening` and open and approve the
   merge-commit staging PR (Stage does not push).
3. After the merge, hand `042-S` to Ship.
4. Optional: delete `%TEMP%\head-stash.jsonl`.

## Resume instructions

Nothing is pending in Stage for this batch. If the operator overrides a decision at staging-PR
review, Stage re-enters at Step 2 for that decision, revises the plan and the affected task text,
and re-assembles (the shipment is `queued`, so tasks can still be edited).

## Adversarial review remediation (2026-10-10)

A multi-model adversarial review (anchor `gpt-5.6-sol` plus three tiers) returned
READY_WITH_FOLLOWUPS: no P0/P1. Because the operator is AFK and Ship runs unattended, the Orchestrator
invoked a narrow Stage pass (dark factory, scope 042-S) to fold the follow-ups into the Stage-owned
text before the staging PR is opened. Text-only: no design change; shipment 042-S membership
(052-F, 052.001-T to 052.005-T), task IDs, titles and dependency edges are unchanged; no scope added.
Each claim was re-verified against the code on this branch (source files equal `main`) before editing.

| Finding | Disposition | Evidence and files changed |
|---|---|---|
| **M-1** (P2) U4 red evidence incoherent | **Applied** | Verified: `u4CreateSymlink` skips on `Errno 1314` under `GOOS == windows` (`writepath_test.go:49-57`); `ci.yml:101-103` sets `cancel-in-progress: true`. Plan U4 AC1/AC4 and posture paragraph, Blocked-path handling (U4 has no CI fallback; U2/U3 fallback waits for the red run and is two commits), both Rollback sections (revert both commits), Runtime Verification U4 row, RK-4, prechecks bullet, risk-signal bullet, CON-F2 history row; task `052.004-T` AC1/AC4; tasks `052.002-T` and `052.003-T` Rollback lines (two-commit exception); deliberation D-BW-3 |
| **M-2** (P2) `DBE25DF5` missing from operator checkpoints | **Applied** | Verified: wrapper `check-write-path-precondition.sh:150` unscrubbed `rev-parse`; ED-7 empty-only (`writepath.go:850-855`); `GATECHECK_SRC` is derived by the sourced `scripts/lib/gatecheck-run.sh:32`; retiredarch's wrapper has the same unscrubbed pattern (`check-retired-architecture.sh:198`). Plan operator checkpoint 6 (separate future task override; deferral kept under P-021 C1; not moved into 042-S) |
| **O-1** (P2) P-021 C2(4) explicit `N/A` IDs | **Applied** | `backlogit stash edit --text` on `4372BAD4`, `DBE25DF5`, `EDC18D59`: appended `task ID N/A; feature ID N/A at capture (052-F assigned at harvest); shipment ID N/A at capture (042-S)` to source refs; field-level diff versus HEAD shows only `text` changed on those three; archived `B83F53BB` untouched. `.backlogit/stash.jsonl` |
| **P-1** (P3) COUNT/KEY_n coverage claim | **Applied** | Verified the e2 row list has no COUNT/KEY row. Reworded to "covered by the generic `GIT_` prefix rule, not independently observed"; no table row added. Plan U3 "Why three vectors"; deliberation D-BW-2 |
| **L-1** (P3) green-on-arrival precedent | **Applied** | Verified: EX-2 is recorded in the Batch A **plan** (`2026-10-08-...-batch-a-correctness-plan.md`, lines 100-104); the Batch A deliberation states the same labelling rule (lines 97, 226); archived `051.001-T` and `051.006-T` carry `harness-ready` and "no red claimed" harness results. Plan EX-2; tasks `052.001-T`, `052.005-T` |
| **L-2** (info) empty-linked-ancestor row | **Not applied** | An ancestor linked to an empty directory returns Code 1 `no fixtures discovered` at the parent and after (`writepath.go:919-921`; the loop never runs), so the row is green on arrival and adds a sixth sub-case to a unit already over its bound (EX-5). The finding's own condition (no enlargement) is not met. Recorded in the plan's remediation table |
| **L-3** (P3) stub sentinels | **Applied** | Plan U2 Posture step 1; task `052.002-T` AC2 |
| **L-4** (P3) wrap invalid-UTF-8 error | **Applied** | Plan U2 helper now `fmt.Errorf("path %q: %w", p, pysem.ErrInvalidUTF8)` (`errors.Is`, INV-D and the planned rows unaffected); deliberation D-BW-1 failure modes; task `052.002-T` description |
| **L-6** (P3) three text fixes | **Applied** | EX-3 count 9 to 10 (adds `hermeticGitEnv`); deliberation D-BW-1a "plus the SplitLines boundaries" reworded to "equals retiredarch's set exactly" (U+001C..U+001E already inside `< 0x20`); U3 e1 baseline pathspec order now `internal/** cmd/**` (the runner's) |
| **L-7** (P3) cross-reference comment | **Applied** | Verified retiredarch `select.go:225` judges every listed record while the writepath helper judges scanned records only. Plan U2 helper notes; task `052.002-T` implementation notes (cite stash `31F33EFE`, not a unit tag) |
| **L-5** (P3), **L-8** | No change | L-5 left as-is (tracked by `EDC18D59`); L-8 was a false positive (operator instruction) |

**Files changed in this pass:** `docs/plans/2026-10-10-intercom-go-gatecheck-batch-b1-write-path-input-hardening-plan.md`,
`docs/decisions/2026-10-10-intercom-go-gatecheck-batch-b1-write-path-input-hardening-deliberation.md`,
`.backlogit/queue/052.001-T.md`, `052.002-T.md`, `052.003-T.md`, `052.004-T.md`, `052.005-T.md`, `.backlogit/stash.jsonl`,
and this memory file. Not changed: `042-S`, `052-F`, every
source, test and config file.

**Tool gate for this pass:** `TOOL_DEGRADED` as above (no MCP tools exposed; backlogit v1.11.0 CLI used
for `sync`, `stash edit` and `stash list`). **Citation correction:** the operator's instruction named
"EX-2 of the Batch A deliberation"; EX-2 lives in the Batch A plan, and both documents are now cited.

---
title: "Stage memory — rev-18 adversarial remediation cycle (017-S task-only manifest, shipment-free deferred work)"
date: 2026-09-12
agent: Stage
session_type: bounded-remediation-cycle
cycle: 2 of max 3
branch: chore/stage-pipeline-policy-gap
base_commit: 9e44bfe
---

# Stage memory — rev-18 adversarial remediation cycle

Second bounded Stage remediation cycle against the adversarial re-review of `9e44bfe`. Verdict
**MUST_REMEDIATE**, not `MUST_REPLAN`. The one-task `017-S` decomposition and the two cohesive
future feature strands were **preserved**; only mechanics were corrected. No generalized
fixture/platform work was added to the current unit.

## Tool status

`TOOL_DEGRADED: backlogit MCP — CLI fallback in use (backlogit 1.10.1)`. All backlog mutations went
through registry-declared `cli_command` fallbacks. `INDEX_SYNC_OK` at session start and end.
P-016 verified: single worktree, `chore/stage-pipeline-policy-gap`, default branch `main`.

A **bounded, disposable probe workspace** outside the repository (`$TEMP/bkl-probe-1`, since
deleted) was used for read-only-with-respect-to-this-repo behavioural verification of backlogit
semantics. No repository worktree was created; no repository file was mutated by the probe.

## Current release unit — unchanged in scope

`018-F` / `017-S`, executable set still exactly `018.008-T` (M / medium). No task added, none split.

## Blockers remediated

1. **`017-S` closure — the rev-17 repair was itself unreachable.** The 13-entry `CASCADE` manifest
   made `018-F` a manifest **item**, and Ship Step 6 pre-mode (`expected_status: done`) compares any
   queue-resident member's declared status with no artifact-type exemption. `018-F` is live and
   `active` at that point (claim moves the covering feature `queued → active`) and **no Ship step
   moves a covering feature to `done`** — verified by execution that backlogit does not
   auto-complete it either. Result: `status-mismatch` → `HALT` before closure ever began.
   **Repaired** by a **12-entry task-only manifest** (`018.008-T` + the 11 pre-archived legacy
   descendants), with `018-F` removed via the **registry-declared supported operation**
   `backlogit shipment return-blocked` and restored with `backlogit move 018-F --status queued`
   (which also cleared the stamped `blocked_reason`). No frontmatter hand-edit.
2. **Full pre/post trace recorded.** Classification `SAFE_CLOSE`; protected set exactly `{018-F}`;
   baseline integrity gate passes; archive loop mutates nothing (all 12 members pre-archived);
   executable set `[018.008-T]` with `pre_archived_skipped [018.001-T, 018.002-T, 018.003-T]`;
   post-close leaves `018-F` live and protected.
3. **Residual P0 escalated, not papered over.** Safe-close **step 8**'s
   `backlogit move <shipment_id> --status shipped` is **refused by backlogit 1.10.1 with exit 9**
   (verified from both `queued` and `active`). Workspace-wide, shape-independent, pre-existing,
   already recorded in the `009-S` report and the compound doc. Both substitutes are contract-
   forbidden. Stashed as `A10EF3D0` (critical, deliberation) with the observed—but not acted on—
   evidence that the cascade op is non-destructive for a **task-only** manifest.
4. **Interim persistence is now OPERATOR-ONLY.** Two rev-17 claims withdrawn as factually wrong:
   that Orchestrator Step 1.5 is "naturally fail-closed" for this route (it is not — a dirty
   `.backlogit/` routes control into item 3, the commit/push arm), and that item 3 supplies
   pre-existing authority (it does not — 3(a) commits Stage artifacts and **3(e) attempts a direct
   push to `main` first**, both against the Orchestrator's own orchestration-only Role statement).
   Contract now: Stage halts at `STAGE_ARTIFACTS_UNCOMMITTED`; **the Orchestrator MUST HALT and MUST
   NOT execute Step 1.5 item 3**; only the operator may verify/commit/push/open-approve the
   merge-commit staging PR/update default/re-verify. **Stated plainly: the dark run cannot finish
   autonomously past that checkpoint while the operator is AFK.**
5. **Harness made P-004-compliant unconditionally.** The "if harness-architect insists" fallback
   framing is **withdrawn**. One generated function, one mandatory expected marker
   (`panic("not implemented: stage branch gate contract")` in an unexported helper inside the single
   test file), all-red before implementation with four explicit captured H0 checks (literal 1-of-1),
   then full-suite green after the sole task.
6. **Step 1.9 algorithm strengthened**: byte-exact em-dash checklist anchors + normalized locator +
   `STAGE_BRANCH_GATE_ANCHOR_FAIL` on zero/multiple; P-016 `git worktree list --porcelain`
   classification of every attached worktree before create/select/resume with the spike/research
   exception explicitly preserved; exact slug source fields per intake shape, five-step ordered
   normalization, explicit 48-byte max, literal `stage-session` empty fallback; base-ref resolution
   from `origin/HEAD` (never assume `main`); branch ownership/base-identity/collision discriminator
   with resume **only** on identity match; symbolic-HEAD equality, expected ≠ default, categorical
   deferral and named halt all retained.
7. **Invalid shipment lock retired.** `blocked` is not a valid backlogit shipment status
   (`queued|active|shipped|abandoned`). Verified by execution: `move <shipment> --status blocked`
   exits **0** (generic-mover defect), `--status shipped` exits **9**, `--status abandoned` from
   `queued` is rejected. Removed the `018-S → 020-S` and `019-S → 020-S` edges, then **archived**
   `018-S`, `019-S`, `020-S` with the official non-destructive `backlogit archive`. Nothing deleted;
   manifests and titles preserved; `archived_status: blocked` retained as honest provenance.
8. **Deferred work is now shipment-free.** `019-F`, `020-F`, `021-F` and all their tasks are
   `blocked` with **no shipment**. Orchestrator Step 2 and Ship Step 0.5 both operate on shipments,
   so there is no eligibility surface at all. This **supersedes** the earlier request for a queued
   future shipment: the adversarial gate proved no valid safe queued representation exists, and
   reliability/safety take precedence. Verified by execution that archival costs nothing —
   `backlogit archive <shipment>` does not cascade, and members of an archived shipment are freely
   re-assemblable into a new shipment.
9. **`021-F` / `021.001-T` retained, not archived.** They do not exist only to support the invalid
   lock: `021.001-T` carries the three-option harness execution-model analysis that the P-006 plan
   review depends on. Only the lock-token machinery was stripped.
10. **Adoption provenance corrected.** Structured `origin_feature: 018-F` is the **original/root**
    origin and is preserved across successive adoptions; the intermediate `018-F → 019-F → 020-F`
    re-parent history is **prose only**. The rev-17 table asserting `origin_feature: 019-F`
    contradicted the stored data and is withdrawn, as is the claim that `adopt` "records
    `origin_feature: 019.001-T -> 020.001-T`" (that was an ID remap, not a field value).
11. **Executable filtering corrected to `artifact_type: task`, never ID suffix** — now load-bearing,
    because the manifest deliberately contains eight `-ST` subtasks whose IDs end in `T`.
12. **Dependency hygiene**: no edge references an archived shipment; no cycle; no `current → future`
    dependency. `018.008-T`'s `related_to` link to `019.007-T` is a semantic link, not a `blocks`
    edge, and does not appear in `dep list`.

## Verification

`backlogit sync` OK · `backlogit doctor` "No issues found" · `backlogit queue view`: `017-S` is the
**only** shipment and the only live queued shipment; `018-F`/`018.008-T` queued;
`019-F`/`020-F`/`021-F`/`021.001-T` blocked; **no invalid shipment status remains in the queue** ·
`backlogit shipment list` returns `017-S` only · `markdownlint-cli2` clean on changed docs ·
dependency graph acyclic, future → current only · every in-scope task has ≥4 ACs and a structured
size.

## Strict scope

P-017 scope remains **`017-S` only**. Future work (`019-F`, `020-F`, `021-F`) is outside it and has
no shipment.

## Next actions (not Stage's)

**OPERATOR ONLY.** The commit produced by this cycle is **local**. Only the operator may push
`chore/stage-pipeline-policy-gap`, update PR #54, and merge. Stage did not push, did not touch
PR #54, did not invoke Copilot, did not reply to or resolve any review thread, did not claim, ship
or abandon a shipment, and did not invoke Ship.

## Open questions

* `021.001-T` option selection (single-task features vs. serialization vs. contract amendment).
* **OQ-2 / stash `A10EF3D0` — the safe-close step-8 closure conflict.** Highest-priority open item:
  `017-S` merges normally but pauses at Ship Step 6 closure until this is resolved.
* Whether `019.004-T` should also repair the Orchestrator Step 1.5 item 3 internal contradiction.
* Whether the generic-`move`-accepts-invalid-shipment-status defect should be reported upstream.

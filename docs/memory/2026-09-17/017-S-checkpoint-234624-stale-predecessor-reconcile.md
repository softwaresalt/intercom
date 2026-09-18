# 017-S — Stale Predecessor Checkpoint Restore / Reconcile / Resolve (Stage)

**Session:** stage-2026-09-17-017-S-scoped-final-verification (resumed)
**Restored checkpoint:** `checkpoint-20260917-234624.json` (stale, pre-merge)
**Reconciled against:** `checkpoint-20260917-235351.json` (resolved successor) + current `origin/main`
**Recorded at:** 2026-09-17T18:45-07:00
**Operation:** Stage-owned crash-resumption recovery (get → validate → restore → prune → reconcile → resume → resolve)
**Scope:** DARK_MODE_ACTIVE, exactly `017-S`

## Tool availability (P-012)

MCP surface unavailable this session. Registry `.autoharness/backlog-registry.yaml` declares
CLI fallbacks for every checkpoint operation → `TOOL_DEGRADED` (not silent filesystem reads).

- `get_checkpoint` → `backlogit checkpoint get {{filename}}`
- `list_checkpoints` → `backlogit checkpoint list`
- `resolve_checkpoint` → `backlogit checkpoint resolve {{filename}}`

`backlogit` 1.10.1. `INDEX_SYNC_OK` (246 artifacts indexed).

## Recovery gate record

- **Operator confirmation:** explicit required token supplied, naming the single checkpoint
  filename plus matching agent, session and phase. No auto-pick.
- **Enumeration anomaly scan (fail-closed, FULL unfiltered list before partitioning):**
  30 records, `needs_quarantine=0`, `quarantined=0`, 0 missing/malformed required fields,
  30/30 returned. Clean → proceed. No `status`/`agent` filter applied at the API call.
- **Partition:** 6 Stage-owned `active` checkpoints; 1 `ship`-owned record present and
  `resolved` (never touched — cross-role handling prohibited, P-001).
- **Owner validation:** `agent: stage` exact match; `valid: true`; `schema_version: 1`;
  `session_id` and `phase` both identical to the operator token.
- **Engram substrate reachable at restore:** daemon v0.3.0-rc.1, uptime 4375s, branch `main`,
  89 code files, 456 edges, `stale_files: false`, scan 4176/4176 complete. Prune gate
  therefore executes rather than failing closed.

## Bounded prune-on-restore

Sequence honored: **restore → prune/gate → reconcile → resume** (never restore → resume → prune).
Superseded action-observation history summarized rather than replayed verbatim: the pre-merge
"awaiting PR #57 merge authorization" deliberation turns, plan-review attempts 1–10 traces and
per-persona finding transcripts — all already synthesized into recorded state and Git/PR history.

### Preserved under the prune allowlist (never pruned)

1. **Active cursor** — shipment `017-S`, covering feature `018-F`; next action
   "claim 017-S after PF-0..PF-12 preflight"; `next_owner: ship`.
2. **Unresolved-checkpoint pointer** — `checkpoint-20260917-234624.json` itself, retained
   through the prune and resolved only after confirmed resume.
3. **Gate verdicts** — retained intact: `OPERATOR_ACCEPTED_RESIDUALS` @ revision 15,
   `reviewing_sha 4b92d52` (contract D-L7 PASS-equivalent; NOT a reviewer PASS); attempts
   8/9/10 zero P0; PF-2 conditions (a) and (b); four eligibility grounds; Probe 25 / Probe 26
   PASS; PA-017-CASCADE approved-but-escrowed with nine conditions re-measured at invocation
   (D-L2 no claim authority, D-L3 no scope widening, D-L4 no inheritance); P-021 deferrals
   DB12DA37, 11B75632, 3F546E63, B6EF23CC.

## Reconciliation — stale phase vs. successor + origin/main

Restored phase `017-S-ground-3-condition-a-discharged-awaiting-pr57-merge-authorization` is
**STALE**. Its open item — condition (b), plan presence on `origin/main` — is now discharged.

| Recorded in 234624 (stale) | Measured now | Result |
|---|---|---|
| `head: a7f0271` | `a7f0271` is merge parent 2 of `5d0c0d2` | superseded, consistent |
| `next_action: authorize PR #57 merge` | PR #57 merged at `5d0c0d2` | **discharged** |
| `next_owner: operator` | now `ship` per successor | superseded |
| condition (b) open | plan + contract resolve on `origin/main` | **discharged** |
| `pr: 57`, `reviewing_sha: 4b92d52`, `plan_revision: 15` | unchanged | match |

Successor/origin evidence verified independently:

- `5d0c0d2` has **exactly two parents** (`f2d4cf9` mainline, `a7f0271` reviewed PR head) →
  P-009 conformant merge commit; `git merge-base --is-ancestor 5d0c0d2 origin/main` = 0.
- `origin/main` = `b04dd64` (matches successor `origin_main`), 3 commits after the merge.
- **PF-2 byte-identity measured under the gate's own scope.** The plan's PF-2 text compares
  "everything above the `## Plan Review` heading", not the whole file. Heading at line 1088 in
  both revisions; pre-heading region = 108,268 bytes, sha256
  `015c8daad946234574161e612f4d2d65c5320e6ce729ffaac499469ef46c0862` at **both** `4b92d52`
  and `origin/main` → **identical**. Whole-file blobs differ only below the heading (the
  revision-15 disposition row + status prose), which PF-2 does not compare. Canonical contract
  is whole-file identical (`feaeeb3b`).
- Winning ledger row on `origin/main`: `| — | 15 | operator-disposition | — | OPERATOR_ACCEPTED_RESIDUALS | 4b92d52 |`,
  unsuperseded by any later attempt.
- Shipment `017-S`: status **`queued`** (unclaimed), 13-member manifest intact,
  dependencies `[021-S]`, record already reads CLAIMABLE.

Note: the plan prose on `origin/main` still reads "not claimable until that merge lands" —
self-referential text authored pre-merge. Its condition is satisfied by the file's own presence
on `origin/main`; it is not a live blocker.

## Resume outcome

Resumed at the reconciled phase: all four grounds discharged, `017-S` CLAIMABLE, Stage work
COMPLETE. The stale phase's only open item is closed. No Stage pipeline steps
(triage/deliberate/plan/harvest/shipment-assembly) are outstanding.

Stage did **not** claim `017-S`, did not cascade, performed no Ship work, created no branch or
worktree (single worktree, on `main`, HEAD `b04dd64`).

`checkpoint-20260917-234624.json` resolved after confirmed resume. The other **five** active
Stage checkpoints were NOT touched, and the `ship`-owned checkpoint was NOT touched.

---
title: "Session memory — Post-M4 stash triage and Go re-plan"
agent: Stage
date: 2026-10-01
branch: chore/stage-post-m4-followups-and-go-replan
base: main @ 9b299c8
deliberation: docs/decisions/2026-10-01-intercom-go-post-m4-stash-triage-and-go-replan-deliberation.md
plan: docs/plans/2026-10-01-intercom-go-post-m4-go-replan-and-containment-hardening-plan.md
---

# Session memory — Post-M4 stash triage and Go re-plan

Stage session invoked by the Orchestrator in sequential, non-dark mode. Two
mandates: triage 8 active stash entries, and re-plan the three held shipments
(030-S, 031-S, 032-S) from the M4-retired Python gate engines onto the Go
engines under `tools/gatecheck`.

## Gate and tool state

| Gate | Outcome |
|---|---|
| Step 0.0 tool availability | `ALL_TOOLS_OK` — backlogit 1.11.0, registry features `sizing`/`shipments`/`dependencies`/`checkpoints` all true |
| Step 0.1 index sync | `INDEX_SYNC_OK` (377 artifacts at start, 382 at end) |
| Checkpoint scan | 35 total, 0 quarantined, 0 unresolved |
| Step 1.9 branch gate | PASSED — single worktree, `base_ref=main`, branch created from `refs/remotes/origin/main` |
| P-021 C5 duplicate scan | CLEAN for all 8 targets (unconditional scan, 27 entries, 8 theme patterns) |
| P-021 C6 late-id reconciliation | PR #85 recovered for 7223218F / 5A8EC1BC / 978D2946; PR #87 for 44F8CC48 |
| Step 4 plan review | Round 1 FAIL → round 2 FAIL → **round 3 ADVISORY** (no P1/P2) |

**Degraded tool (must carry forward):** this workspace's `.backlogit/header-def.yaml`
defines `size` for `task` but **does not define `complexity`**, even though the
registry advertises a `complexity` param on `update_task`. `backlogit update
<id> --complexity <v>` fails with a validation error. Complexity is therefore
preserved as labeled prose (`Size: X | Complexity: y`) in every task
description, per the Stage structured-emission capability gate.

## Stash dispositions (all 8 carried the `DEFERRED SCOPE EXPANSION` marker)

Per P-021 C6 precedence, all 8 were forced onto the `deliberate` route
regardless of shape or size.

| ID | Pri | Group | Disposition | Result |
|---|---|---|---|---|
| 7223218F | high | G1 | HARVEST | 048-F / 048.001-T–048.004-T; archived |
| 50E6F22C | med | G1 | HARVEST | 048.005-T; archived |
| 44F8CC48 | low | G6 | HARVEST (fold-in) | 048.006-T; archived |
| D10D3AFC | high | G2 | DEFER | needs its own harness deliberation |
| 9F824B64 | med | G2 | DEFER | real contract conflict, verified |
| 5A8EC1BC | low | G3 | DEFER | own stated trigger not fired |
| C312BD4C | low | G4 | DEFER + RE-SCOPE | fix the comment, not the regex |
| 978D2946 | med | G5 | DEFER | standalone CI hygiene |

Every claim was verified against current code before acceptance. Notable
confirmations: `scripts/acquire_lock.*` genuinely absent (9F824B64 correct);
`check-write-path-precondition.sh:84` forwards raw `"$@"` unguarded (7223218F
is a **live** gap); the gomask multiline-tag contradiction survived the port.

## Re-plan outcomes

### 030-S / 033-F — HOLD LIFTED
* `033.001-T` → A-T1 (scanScope + rewire both consumers in `select.go`)
* `033.002-T` → A-T2 (re-anchor pin in `pin.go` + committed negative control), S/medium
* `033.003-T` → **DROPPED** (archived status unavailable — see limitation below)
* **ALP-1 atomic landing pair declared**: A-T1 is not independently landable
  (moving the literals necessarily reddens the pin by design), so Ship must land
  033.001-T and 033.002-T as a **single commit**.

### 032-S / 035-F — HOLD LIFTED
* `035.001-T` → B-T1 (**test task**, inverts `git_test.go:103-121`)
* `035.002-T` → B-T2 (implementation in `git.go`)
* `035.003-T` → B-T3 (comment correction), XS/low
* Dependency direction corrected to genuine test-first: added edge
  `035.003-T → 035.002-T`; chain is now 001 → 002 → 003.

### 031-S / 034-F — HOLD LIFTED after the operator SPLIT (D-031-2)
First session: held on the **Option A vs Option B fork** (D-031-1). Follow-up
session (2026-10-01/02): the operator chose to **SPLIT** 031-S. See
"Follow-up session — D-031-2 split" below.

### 048-F / 038-S — NEW
Six tasks, 3×S + 3×XS, 0 unsized. Dependency edges: `048.001-T → 048.002-T →
048.003-T`, `048.001-T → 048.004-T`; 048.005-T and 048.006-T independent.

## Key corrections made during the session

1. **Inverted hazard narrative (plan-review round 1).** My rev-1 claim that the
   pin "silently stops pinning while still passing" was factually backwards.
   `containsAll` returns **false** when literals move → the pin **fails loudly**
   (`select.go:5-9` documents this as deliberate). The real unguarded hazard is
   `containsAll(_, []) == true` at `pin.go:127` — emptying the literal lists
   yields a **vacuously green pin**. Corrected in the plan, the deliberation
   (D-030-2) and the task descriptions.
2. **Bare gate invocations (round 1, P1).** `check-unignore-regression.sh` and
   `check-gitignore-append-only.sh` always exit 1 when invoked bare. Several ACs
   and the Verification block were unsatisfiable; corrected to the real CI entry
   points (`ci.yml:464`, `:486`).
3. **Unused-linter landability (round 1, P1).** `.golangci.yml` `default:
   standard` includes `unused`; a declaration-only A-T1 would redden CI. Merged
   declaration + rewiring; added INV-4.
4. **Not independently landable (round 2, P1).** The merge above still left
   A-T1 breaking the pin. Resolved with INV-5 / ALP-1 rather than a further
   merge (which would have breached the 2-hour rule at 3 files / ~5 functions).

## Known limitations and follow-ups

* **033.003-T could not be given a terminal status.** `.backlogit/hooks.yaml`
  allows only `queued → active|blocked`; there is no `queued → archived|rejected`
  transition. The task is retitled `DROPPED (post-M4 Go re-plan) …` with full
  rationale, but it remains `queued` and a member of 030-S. **Operator action
  needed** to give it a terminal state. *Correction (follow-up session):*
  `backlogit shipment return-blocked` **is** a member-removal path, but it
  leaves the item `blocked`/orphaned. Removal is therefore not used for
  satisfied or dropped items. 034.001-T has the same status.
* Satisfied `blocks` edges were **retained** as historical record (D-000-1).
* Plan hardening was authored **inline during impl-plan**, not as a discrete
  `plan-harden` skill pass.
* Recommended fold-ins outside the assigned 8, flagged not harvested:
  **DD0BB60F** (weakens the pin reasoning → belongs with 030-S; **still not
  approved, untouched**), **8E9F8E55** and **B72E9715** (→ 031-S; **harvested
  and archived in the follow-up session**).

## Follow-up session — D-031-2 split (2026-10-01/02)

**Operator decision D-031-2 (explicit):** SPLIT 031-S.
* **031-S** keeps the Option A masked-text hardening increment.
* The work that `go/ast` replaces moves to a new feature and shipment.

Stage added **D-031-3** to re-anchor D-1/D-2 onto the Go engine as D-1′/D-2′.
D-031-3's operator acknowledgment is **PENDING**.

**Plan-review of rev 4 (031-S Unit D + Unit E):**
* **Round 1 FAIL** (1 P0, 6 P1). Remediation:
  * D-1′/D-2′ amendments;
  * an oracle split-out (034.010-T);
  * a fixture split-out (034.011-T);
  * residual item 6.
* **Round 2 FAIL** (1 P1: the oracle can't survive the Unit E boundary).
  Remediation:
  * a single authorised oracle adaptation (AC-E2.6);
  * rule 3b, and rule 7 narrowed to exactly `FILE_FLAG_BACKUP_SEMANTICS`;
  * residual item 7 (split selector).
* **Round 3 ADVISORY**, 0 P0/P1, on the anchor plus Security, Go and
  Constitution personas. Advisories applied. **031-S STAGE HOLD LIFTED.**
* Plan markers: `<!-- plan-review-attempt: 2 -->`, `<!-- plan-review-attempt: 3 -->`.

**ID map** (rollback record; nothing deleted):

| Plan | ID | Note |
|---|---|---|
| D-T0 | 034.001-T | SATISFIED; kept queued in 031-S with no live edges |
| D-T1a | 034.010-T | NEW (frozen differential oracle) |
| D-T1/T2/T3/T4/T5a | 034.002/003/004/007/008-T | rewritten in place |
| D-T5b | 034.011-T | NEW (presence fixtures) |
| D-T6 | 034.009-T | NEW (8E9F8E55 fixtures) |
| E | 049-F (blocked = STAGE HOLD) / 039-S | NEW |
| E-T1/T2/T3/T6 | 049.001/002/003/006-T | NEW |
| E-T4 | 049.004-T | **was 034.006-T**: `shipment return-blocked` → `move queued` → `adopt` |
| E-T5 | 049.005-T | **was 034.005-T**: same path |

**Manifests:**
* 031-S: 034-F, 034.001, 002, 003, 004, 007, 008, 010, 011, 009. Sizes 8×S +
  1×XS.
* 039-S: 049-F, 049.001…006. Sizes 3×M + 1×S + 2×XS.

**Edges added:**
* 034.002→034.010, 034.007→034.004, 034.008→034.004, 034.011→034.008,
  034.009→034.011;
* the 049.00x chain;
* 049-F→034-F, **039-S→031-S (blocks)**, **038-F→049-F**.

Six edges were removed. The full list is in the plan's
`## Harvest Record — Revision 4`.

**038-F consequence:** its trigger was the tripwire "delivered by 034.005-T",
which is now 049.005-T. Stage added the edge to 049-F and a marked description
note.

**Stash:** 8E9F8E55 → 034.009-T and B72E9715 → 034.004-T. Both had their text
edited to carry `PR=#71` (C6 reconciliation) and the harvest target, then were
**archived**. The C5 duplicate scan was CLEAN, and review-thread N/A stands.

**Open operator decisions:**
1. Acknowledge D-031-3.
2. Approve the H-5 per-session re-measure; declining the CI tripwire rests on it.
3. Confirm `os.Chtimes` in 034.009-T (the D-031-2 summary named three
   selectors; the stash entry lists four).
4. Terminal state for 034.001-T and 033.003-T, a closure prerequisite for 031-S
   and 030-S.
5. Schedule the Stage-executed `go/ast` spike session (P-016 worktree). Lifting
   the 049-F hold depends on it.
6. Selector-set widening (residual item 6), a stash candidate.
7. DD0BB60F remains pending.

## Recommended execution order

**038-S (Unit C) → 032-S (Unit B) → 030-S (Unit A) → 031-S (Unit D) → 039-S
(Unit E).**
* Unit C goes first because it is the only unit closing a live gap, and
  `048.001-T` changes shared CLI dispatch that the other units run through.
* 039-S is hard-gated: it `blocks`-depends on 031-S and is held by the 049-F
  STAGE HOLD until the spike re-plan passes plan-review.

*Generated by Copilot*

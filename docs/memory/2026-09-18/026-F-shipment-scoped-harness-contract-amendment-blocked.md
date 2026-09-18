---
date: 2026-09-18
agent: stage
session: shipment-scoped-harness-contract-amendment
branch: chore/stage-shipment-scoped-harness-contract-amendment
base: 101e974a2890e7ba82aacb22d734866fddaf973b
outcome: BLOCKED — plan-review FAIL attempt 1, P-005 recorded, no shipment created
---

# Stage session — B1/B2/DB12DA37 scoped Option 3 amendment → BLOCKED

## Outcome in one line

The contract amendment was deliberated, planned, hardened and reviewed. **Plan-review FAILED**
(attempt 1, 2× P0, 14× P1). **No shipment was created.** A precise blocked decision is persisted
and two blocked features carry the work forward.

## What was measured and is now settled (do not re-derive)

| Claim | Status | Evidence |
|---|---|---|
| `DB12DA37` root cause | **LOCALIZED** | `_ship.agent.md:362-367` — the verb *"replace"* discards **both** the `queued` predicate (intended) and the `harness-ready` predicate (unintended). Step 3 item 1 (`:339-361`) has **zero** occurrences of `harness-ready`. |
| P-002 Enforcement already exists | **TRUE** | `workflow-policies.md:50` — *"Filter ready queue to only tasks carrying the `harness-ready` label"*. So this is a **conformance defect, not a policy conflict**. No P-002/P-004 amendment needed. |
| **B1** ("no legal pre-claim execution slot") | **REFUTED** | `_ship.agent.md:189` fourth arm — *"If already on a branch matching this shipment … `BRANCH_OK` … proceed"*. A legal in-PR slot. B1 was enumerated over only 3 routes. Downgrades to a **specification gap**. |
| **B2** ("handoff is circular") | **SEPARABLE, not universally circular** | `_orchestrator.agent.md:270-297` already defines a working Staging Artifact Merge Gate; PRs #59/#60/#61 all reached `main` through it. B2's real defects are narrower: item 3 names **no actor**; item 3(e) directs a **direct push to `main`** (conflicts with P-009/P-010). |
| Suppressing shipment auto-activation | **UNREACHABLE** | `backlogit` 1.10.1 exposes no `unclaim`/`release` and no flag to suppress descendant auto-activation. The claim is a one-way door. Do not pursue this arm. |
| Retired rev-11…rev-15 mechanisms | **NONE resurrected** | Verified including the `harness_status` custom field (the natural disguise vector). |

## Answer to the operator's separability question

**Not one inseparable unit.** B1 + `DB12DA37` are inseparable *from each other* → Unit 1 / **`026-F`**.
B2 is separable → Unit 2 / **`027-F`**, and is **not** gated on Unit 1 (sequencing preference only,
per the Architecture persona; the source decision does not make it a hard edge).

## Why it is blocked (the decisive pair)

* **C3/C4 — no admissible acceptance surface.** An "unchanged" assertion is **green-on-arrival**, so it
  can never satisfy P-004's *"expected failure markers for every test function"*; harness-architect must
  refuse the label and the bootstrap self-blocks. A behavioural IV-5 harness (required by the adopted
  `022.001-T` resolution) has **no engine to drive** — Ship Steps 2–3 are agent-interpreted prose, not
  code. Declaring IV-5 non-applicable is an **operator determination Stage may not self-grant**.
* **C1/C2/C5/C6 — the unit is larger than it can legally be.** Step 2 forward-references a value Step 3
  has not yet computed; two surfaces both claim final execution-boundary authority; the real surface
  inventory is ~8–9 clause sites across ≥4 files (including `harness-architect/SKILL.md:50` and `:56`,
  two further `ready`-scoped doors); the `manifest ⊆ feature` cardinality proof is invalid. Closing these
  breaches the 2-hour rule, and the unit **cannot be split** without reproducing the G1 deadlock (D4) —
  the same trap that terminated the Option 1 decision.

Against the operator's own admission gate, **2 of 4 conditions fail**: harness-ready timing is
unadjudicated against P-002's Statement clause, and the claim transition is not repairable at all.

## Unblock path (ordered)

1. **O-1 — OPERATOR.** Decide the acceptance surface for agent-interpreted contract text: is a
   structural/substring contract test admissible as the IV-5 surface when no executable engine exists,
   or must a contract-linter engine be built first? **Blocks everything else.**
2. **O-2 — Stage.** Adjudicate P-002's Statement clause (`workflow-policies.md:44`) against post-claim
   in-session harnessing. One-question deliberation.
3. **O-3 — Stage.** Rebuild the surface matrix **whole** (compound stop rule): all ~8–9 sites, four
   columns, verbatim current text, literal pinned replacements, probe rows, explicit `Depends on`.
4. **O-4 — Stage.** Re-scope Unit 1 against a closed matrix; resolve C1 by making **Step 2 the producer**
   of the authoritative selected batch and Step 3 its consumer.
5. **O-5.** `027-F` may proceed independently.

## Mutations

* Created `026-F` (blocked, high) and `027-F` (blocked, medium).
* Dependencies: `025-F → 026-F (blocks)`, `025-F → 027-F (blocks)`.
* `DB12DA37` — **updated in place, kept ACTIVE** (not consumed; no shipment delivered it), priority
  medium → high, superseding paragraph added.
* New stash `9C658143` (low) — empty `**Format**:` gate at `_ship.agent.md:440`.
* `019-F`, `020-F`, `023-F`, `024-F`, `025-F` untouched, all still `blocked`.

## Next session must NOT

* Re-deliberate `DB12DA37` from scratch — start from the measured position above.
* Form a shipment over `026-F` or route it to Ship until **O-1 and O-3** are closed.
* Pursue a P-002/P-004 amendment, or pursue suppressing shipment auto-activation.
* Attempt to split Unit 1 into two queued test-bearing tasks under one covering feature (G1).

## Inherited residual (unremediated)

`019-F`, `019.004-T`, `019.008-T`, `019.009-T`, `020.004-T` still carry prose references to the retired
ID `019.007-T`. Structured dependency edges are correct. For whoever opens the `019-F` strand.

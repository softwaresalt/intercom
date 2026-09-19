---
title: "Deliberation — Shipment-claim reversibility and the CASCADE binding route (3F546E63 / 11B75632)"
date: 2026-09-18
status: decided — routing repair AUTHORIZED (ships in S-2); claim-reversibility BLOCKED on operator determination O-5
agent: Stage
governs: stash 3F546E63; residual execution-model half of stash 11B75632
supersedes: nothing
base_commit: 14d44e3c1f28b321e8db24ab8cafcba346996133
stage_branch: chore/stage-ship-pipeline-contract-repair
measured_against: installed .github/agents/_ship.agent.md and .github/skills/shipment-reconcile/SKILL.md at 14d44e3c; backlogit 1.10.1
outcome: SPLIT_BY_AUTHORITY — executable routing repair proceeds; execution-model change withheld
separability: two units — one shippable, one blocked
revision: 1
---

## 1. Problem frame

Two stash entries describe what looks like one problem and is in fact two, separated by an
authority boundary Stage must not cross.

`11B75632` reports that `_ship.agent.md:827–832` **mandates** carrying a `CLASSIFICATION_BINDING`
into the CASCADE close, while `:861–864` **routes** that close through a direct
`backlogit_ship_shipment` call. Measured: the `ship_shipment` operation accepts exactly
`shipment_id`, `sha`, `message`, `author` — on the CLI surface and in
`.autoharness/backlog-registry.yaml` alike. It has **no** binding parameter. The only surface that
accepts `classification_binding` is `shipment-reconcile` `mode:safe-close`, which routes a bound
CASCADE verdict into its Cascade Close Sub-Procedure.

`3F546E63` reports that the shipment claim is an irreversible one-way door opened four steps before
any build work: Step 0.5 item 4 claims the shipment, Step 2 generates the harness, Step 4 executes
tasks, Step 5 opens the PR, Step 6 closes. Measured against backlogit 1.10.1: there is **no**
`unclaim`, no `release`, and no flag to suppress descendant auto-activation. `return-blocked` is
single-item; `abandoned` is destructive and unapproved. So any post-claim halt strands the shipment
`active` in the P-001 single-active slot with no in-band exit.

The framing question is whether these are one problem or two.

## 2. They are two problems, and the boundary is authority

* `11B75632`'s defect is that **one document contradicts itself** about which operation realizes a
  mandate it already imposes. Repairing it changes *no* policy, adds *no* authority, and removes
  *no* gate. It makes the installed text agree with the only executable route that already exists.
  This is a **conformance repair**, structurally identical to the class 026-F's deliberation
  correctly re-classified.
* `3F546E63`'s defect is that **the contract's shape is expensive**. Repairing it means either
  inverting the claim against Steps 2–5 or giving backlogit a reversible claim. Both are
  execution-model changes to a production contract surface.

Stage may perform the first. Stage may **not** perform the second: it is the same class as 026-F's
blocker O-1 and as the determination 021-F was created to hold. Bundling them would drag a
shippable repair behind an operator decision that has no scheduled resolution.

## 3. Options for the executable half (`11B75632`)

| # | Option | Assessment |
|---|---|---|
| **O1-a** | Re-point the CASCADE branch at `shipment-reconcile mode:safe-close` | **ADOPTED.** Minimal diff; makes the text describe the route that is already the *sole* executable realization of the existing mandate. Independently corroborated: the 017-S closure plan pinned exactly this route at S-17/S-18, and its PF-10 proved it is the only executable realization. Precedent exists in a shipped plan. |
| O1-b | Add a `classification_binding` parameter to the `ship_shipment` op | Rejected here. This is an **engine** change (backlogit CLI + MCP + registry), far outside a markdown conformance repair, and it would create a *second* binding-capable surface — the same two-surfaces-claim-authority defect 026-F's C1/C2/C6 blockers already record against this file. |
| O1-c | Delete the binding mandate at `:827–832` | Rejected. Removes a real control to resolve a documentation contradiction. Strictly worse than making the document consistent. |

**Decision D-A**: adopt **O1-a**. The CASCADE branch is re-pointed at `mode:safe-close`, and the
direct-cascade branch is removed so no future executor can follow the unexecutable path. A contract
test asserts the binding route is singular — this gives the unit a **behavioural** acceptance
surface, deliberately avoiding the green-on-arrival "unchanged assertion" trap that produced 026-F's
C3/C4 blockers. That trap is the single most important lesson available from the prior cycle and it
is designed around here rather than rediscovered.

## 4. Options for the blocked half (`3F546E63`)

| # | Option | Assessment |
|---|---|---|
| O2-a | Claim-late / claim-on-merge mode for closure-shaped shipments | Requires re-ordering `_ship.agent.md` Steps 0.5/2/4/5 and re-specifying P-001's single-active semantics. Operator determination required. |
| O2-b | Reversible claim (`unclaim`/`release`) in backlogit | Engine feature. Cleanest conceptually — it makes a post-claim halt *recoverable* rather than avoided — but it is a new engine capability with its own concurrency semantics against the single-active slot. Operator determination required. |
| O2-c | Status quo + documented stranding-recovery procedure | Cheapest, and it is what 017-S actually did. But it manages the symptom; the one-way door remains and every future claim route re-inherits it. |
| O2-d | Do nothing | Rejected as a *recorded* outcome — the exposure is real and repeatedly re-encountered. |

**Decision D-B**: **BLOCKED, no option adopted.** Each of O2-a, O2-b, O2-c changes who may do what
and when, on a production contract surface. Stage selecting among them would be self-authorizing an
execution-model change (P-017). Recorded as **operator determination O-5**.

**Decision D-C**: `3F546E63` is **not** archived and **not** downgraded. It lands as a blocked
feature carrying the measured evidence above, so the question is visible in the backlog rather than
buried in stash text. Its priority stays medium: it blocks nothing today — every shipment in this
cycle proceeds with the claim at its latest policy-valid point — but every claim route keeps
inheriting the stranding exposure.

## 5. Interaction with this cycle's own shipments

Nine shipments are queued this cycle, so the stranding exposure is inherited nine times. That is an
argument for raising O-5's urgency, **not** for Stage resolving it unilaterally. Two mitigations are
available without any contract change and are applied:

1. **Each unit is small.** The smaller the post-claim work, the narrower the window in which a halt
   can strand a shipment. This is a direct, concrete benefit of rejecting the one-big-shipment
   option in the cycle deliberation.
2. **The chain is ordered so the most contract-fragile work ships first**, while the backlog is
   otherwise empty and a stranded shipment would contend with nothing.

## 6. Open questions

* **O-5** *(operator, new)* — Should the shipment claim become reversible (`unclaim`/`release`), or
  deferrable (claim-late / claim-on-merge) for closure-shaped shipments, or should the status quo be
  formalized with a documented stranding-recovery procedure? Blocks `3F546E63`.
* Cross-reference: 026-F's **O-1**/**O-3** and 021-F's execution-model determination are adjacent but
  distinct. O-5 concerns claim *reversibility*; O-1 concerns the admissible *acceptance surface* for
  an agent-prose repair. They should not be merged.

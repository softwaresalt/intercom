---
title: "Implementation Plan — Stage Artifact Persistence Route Execution (Strand A)"
date: 2026-09-12
revision: 2
status: scope-defined-not-implementation-ready
agent: Stage
governs: feature 019-F, shipment 018-S
source: docs/decisions/2026-09-12-intercom-go-stage-policy-gap-re-plan-decision.md
split_decision: docs/decisions/2026-09-12-intercom-go-deferred-split-and-readiness-lock-decision.md
predecessor_plan: docs/plans/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-plan.md
sibling_plan: docs/plans/2026-09-12-intercom-go-stage-contract-enforcement-platform-plan.md
---

# Implementation Plan — Stage Artifact Persistence Route Execution (Strand A)

**Feature**: `019-F` · **Shipment**: `018-S` (**blocked**, readiness-locked, **not claimed**)
**Requires plan hardening**: **yes**
**Status**: scope and dependencies defined; **NOT implementation-ready**. See §5.

> **Revision 2 (2026-09-12).** This plan was narrowed to **Strand A only**. The reusable
> mechanical-enforcement platform it previously carried moved to `020-F` / `019-S`, planned in
> `docs/plans/2026-09-12-intercom-go-stage-contract-enforcement-platform-plan.md`. The rev-1 text
> is in git at `14d64e4`.

## 1. Why this feature exists

`018-F` delivers the **prohibition** half of the Stage default-branch policy gap: Stage may not
commit on the default branch, may create `chore/stage-{scope-slug}`, and must pass a fail-closed
Step 1.9 gate before any tracked mutation. It deliberately delivers **no persistence execution**.

Until this feature lands:

* a Stage session terminates with bounded artifacts **uncommitted** on the Stage branch;
* the installed Orchestrator Step 1.5 **cannot discover that branch** — its step 2 evaluates
  `git log origin/main..main`, which is empty whenever Stage worked on a side branch;
* the pipeline **fails closed** at Step 1.5 step 4 with `STAGING_GATE_FAIL` unless an operator or
  Orchestrator performs the manual sequence recorded in §R16.5 of the predecessor plan.

This feature replaces that manual sequence with a specified, executable one.

## 2. Why Strand A is separate from Strand B

The former single deferred feature mixed **persistence semantics** — what Stage and the
Orchestrator *do* at handback — with **reusable enforcement infrastructure** — detector, fixtures,
event parsing, CI, regeneration proof. They have different consumers, different blast radius and
different review needs, and bundling them is what let the original release unit grow to eight
tasks. After the split, **neither strand depends on the other**: each depends only on
`018.008-T` and its own internal chain.

## 3. Dependency on current work

`019-F` **depends on** `018-F`; `018-F` depends on nothing here. Verified: no `current → future`
edge and no cycle anywhere in the graph.

| Task | Size | Blocks on | Rationale |
|---|---|---|---|
| `019.007-T` | M | `018.008-T` | The handback record **names the Step 1.9 branch** that `018.008-T` creates |
| `019.008-T` | M | `019.007-T` | The no-shipment terminal emits the same handback record |
| `019.009-T` | M | `018.008-T`, `019.007-T` | The commit step runs **on** that branch and must match the handback's path-set derivation |
| `019.004-T` | M | `018.008-T`, `019.007-T` | Orchestrator branch discovery **consumes** the handback record |

```text
018.008-T (current unit — external prerequisite)
   └─> 019.007-T ─┬─> 019.008-T
                  ├─> 019.009-T
                  └─> 019.004-T
```

The rev-1 edge `019.004-T → 019.003-T` (now `020.003-T`, the Strand B detector) was **removed** in
the rev-17 remediation: Orchestrator branch discovery semantically requires the handback record
and the corrected contract, not the detector.

## 4. Scope

| Task | Summary |
|---|---|
| `019.007-T` | Step 6 staging-handback emission with concrete artifact paths and the named derivation command |
| `019.008-T` | Make the `no-shipment` terminal outcome reachable — **failures still halt** |
| `019.009-T` | Step 5.7 Stage artifact commit with out-of-root path safety |
| `019.004-T` | Orchestrator Step 1.5 authorization, branch discovery, contradiction repair, concrete post-merge verification |

**Out of scope**: everything in `020-F` (detector, fixtures, event parser, CI wiring, CODEOWNERS,
divergence ledger, regeneration proof), and any change to `018-F`'s reduced contract.

## 5. BLOCKING PREREQUISITE — the harness execution model

This feature has **four test-bearing tasks**. P-004 requires every generated test function RED
before implementation; Ship Step 4.3 requires the full `go test ./...` GREEN after every task; and
**Ship Step 2 item 1 harnesses every `queued` task of the covering FEATURE in one batch**.

The rev-1 recommendation to "split into a sequence of single-task **shipment manifests**" was
**invalid and is withdrawn**: the harness batch is scoped to the covering **feature**, not to the
manifest, so several single-task manifests over one multi-task feature still deadlock.

The decision is tracked by **`021.001-T`**, which enumerates the three admissible options
(single-task features/shipments; machine-enforced one-queued-child serialization proven against
Ship intake and P-015; or a reviewed contract amendment making harness selection shipment-scoped).
**No mechanism is selected here.**

**Do not resurrect the rev-11..15 manifest / terminal-witness / `-gatetask` / MP0–MP6 machinery
without explicit re-deliberation.** It is non-governing provenance in the predecessor plan's
historical appendix.

## 6. Mechanical readiness lock

`018-S` is **not claimable**, and not merely labelled so:

* `018-S` is `status: blocked` — Orchestrator Step 2 triggers on `queued` shipments only, and
  Ship Step 0.5 item 1b rejects a shipment that is neither `queued` nor `active`.
* `018-S` carries a `blocks` dependency edge to readiness-gate shipment `020-S`, which is itself
  `blocked` and never executed. Orchestrator Step 2's explicit pre-claim re-check treats an
  unshipped blocking predecessor as a **hard** eligibility gate regardless of queue position.

Verified: `backlogit queue view --type shipment` lists only `017-S`.

**Unlock procedure (Stage only).** After `021.001-T` records a reviewed decision:
`backlogit dep remove 018-S 020-S` → restructure into the shipment shape that decision requires →
`backlogit move 018-S --status queued` and re-queue member tasks → run `impl-plan`,
`plan-harden` (**required**) and `plan-review`. Removing the edge without the review is a P-006
violation. Ship and the Orchestrator must not remove these edges.

## 7. Risks

| ID | Risk | Mitigation |
|---|---|---|
| R1 | Harness execution model unresolved → repeat of the revs 10–15 failure | §5 is a hard gate; `020-S` enforces it mechanically |
| R2 | `019.008-T` converts a loud failure into a silent success | Load-bearing constraint in the task: P-003 violations, errored harvests and unformable shipments must still **halt** and must be provably non-recordable as `no-shipment` |
| R3 | `019.004-T` silently widens the Orchestrator role boundary | AC-2 forbids creating new authority; any widening is a separate deliberated P-010 amendment |
| R4 | Interim fail-closed halt mistaken for a defect | §R16.5 of the predecessor plan states it is the intended terminal and names the manual completion sequence |

## 8. Next steps

1. Resolve `021.001-T` (harness execution model) under deliberation and plan review.
2. Restructure `019-F` / `018-S` per that decision.
3. Run `impl-plan`, then `plan-harden` (**required**) and `plan-review`.
4. Only then remove the `018-S → 020-S` edge and re-queue.
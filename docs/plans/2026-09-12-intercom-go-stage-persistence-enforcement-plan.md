---
title: "Implementation Plan — Stage Persistence Route Execution and Mechanical Policy Enforcement"
date: 2026-09-12
revision: 1
status: scope-defined-not-implementation-ready
agent: Stage
governs: feature 019-F, shipment 018-S
source: docs/decisions/2026-09-12-intercom-go-stage-policy-gap-re-plan-decision.md
predecessor_plan: docs/plans/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-plan.md
---

# Implementation Plan — Stage Persistence Route Execution and Mechanical Policy Enforcement

**Feature**: `019-F` · **Shipment**: `018-S` (queued, **not claimed**)
**Source decision**: `docs/decisions/2026-09-12-intercom-go-stage-policy-gap-re-plan-decision.md`
**Requires plan hardening**: **yes**
**Status**: scope and dependencies defined; **NOT implementation-ready**. See §5.

> This plan deliberately stops at **reviewed scope**. Per the re-plan mandate, the deferred
> feature "need not be implementation-ready beyond a reviewed plan if it is intrinsically
> complex; it must have a clear scope and dependency on current work."

## 1. Why this feature exists

`018-F` / `017-S` accumulated a verification platform plus the end-to-end persistence route and
failed adversarial plan review at revisions 10–15. The 2026-09-12 re-plan reduced `018-F` to the
**prohibition** half of the Stage default-branch policy gap — one task — and moved everything
separable here.

Detailed root-cause analysis is in the source decision §2 and in the predecessor plan's §R16.1.

## 2. Dependency on current work

`019-F` **depends on** `018-F`; `018-F` depends on nothing here.

| Task | Blocks on | Rationale |
|---|---|---|
| `019.001-T` | `018.008-T` | A detector's fixture corpus must encode the **corrected** contract; building it first would freeze the defective shape |
| `019.007-T` | `018.008-T` | The handback record **names the Step 1.9 branch** that `018.008-T` creates |
| `019.009-T` | `018.008-T`, `019.007-T` | The commit step runs **on** that branch and must match the handback's path-set derivation |

Only the reduced `017-S` proceeds after the staging PR merges. **`018-S` is future work.**

## 3. Scope — two strands

### Strand A — Persistence route execution

Makes the corrected policy **runnable end to end**, including when no shipment is formed.

| Task | Size | Summary |
|---|---|---|
| `019.007-T` | M | Step 6 staging-handback emission with concrete artifact paths and the named derivation command |
| `019.008-T` | M | Make the `no-shipment` terminal outcome reachable — **failures still halt** |
| `019.009-T` | M | Step 5.7 Stage artifact commit with out-of-root path safety |
| `019.004-T` | M | Orchestrator Step 1.5 authorization, branch discovery, concrete post-merge verification |

### Strand B — Mechanical enforcement platform

Makes regeneration of the autoharness-generated contract files unable to silently reopen the gap.

| Task | Size | Summary |
|---|---|---|
| `019.001-T` | S | Direct-push fixture corpus and manifest |
| `019.002-T` | S | Fixture-suite driver over a stubbed detector |
| `019.003-T` | M | Structure-aware direct-push detector |
| `019.005-T` | S | CI lint-job wiring, divergence ledger, CODEOWNERS |
| `019.006-T` | S | Coupling verification with executed mutation proof |

## 4. Execution order

```text
018.008-T (current unit — external prerequisite)
   ├─> 019.007-T ─> 019.008-T
   │        └────> 019.009-T
   └─> 019.001-T ─> 019.002-T ─> 019.003-T ─> 019.004-T ─> 019.005-T ─> 019.006-T
```

`019.005-T` additionally blocks on `019.007-T` and `018.008-T`.

## 5. BLOCKING PREREQUISITE — the harness lifecycle must be decided first

This feature has **nine tasks**. The structural impossibility identified in the re-plan applies
with full force:

* **P-004** requires *every* generated test function RED before implementation.
* **Ship Step 4.3** requires the *full* `go test ./...` GREEN after *every* task.

These are **jointly unsatisfiable for any shipment with more than one test-bearing task**.
**`018-S` must therefore not be claimed as a single nine-task shipment.**

Before execution, a reviewed plan must choose one of:

1. **Split `018-S` into a sequence of single-task shipments**, each with exactly one red→green
   transition. **Recommended default** — already proven by the reduced `017-S`, requires no
   contract amendment.
2. **Amend the installed harness contract** (`P-002`/`P-004`, `_ship.agent.md` Step 2 and 4.3,
   `harness-architect`) to legitimately support a multi-task upfront-red batch. This was
   **expressly out of scope** for `018-F` and is a deliberate safety-contract change requiring its
   own deliberation.
3. **An equivalent mechanism** that does **not** reintroduce surface-predicate activation.

**Do not resurrect the rev-11 manifest / terminal-witness / `-gatetask` mechanism without
explicitly re-deliberating it.** Its retired defects are documented in the predecessor plan's
historical appendix (revisions 10–15) and must be read before any such proposal.

## 6. Risks

| ID | Risk | Mitigation |
|---|---|---|
| R1 | Harness lifecycle unresolved → repeat of the revs 10–15 failure | §5 is a hard gate before claiming `018-S` |
| R2 | Detector over-generalization — the original scope driver | Constrain to the corrected contract corpus; prefer targeted assertions over a generalized platform |
| R3 | `019.008-T` converts a loud failure into a silent success | AC-5 regression guard: P-003 violations, errored harvests, and unformable shipments must still **halt** and must be provably non-recordable as `no-shipment` |
| R4 | Interim regeneration vulnerability while `019-F` is deferred | Accepted and tracked; see decision §6 |

## 7. Out of scope

Anything already delivered by `018-F` (P-010 grants, Role Boundary rows, the Step 1.9 gate), and
any change to `018-F`'s reduced contract.

## 8. Next steps

1. Deliberate the §5 harness-lifecycle question.
2. Run `impl-plan` for the chosen option, then `plan-harden` (required) and `plan-review`.
3. Only then restructure/claim `018-S`.

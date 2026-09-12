---
title: "Implementation Plan — Mechanical Enforcement of the Stage Default-Branch Contract (Strand B)"
date: 2026-09-12
revision: 1
status: scope-defined-not-implementation-ready
agent: Stage
governs: feature 020-F, shipment 019-S
source: docs/decisions/2026-09-12-intercom-go-deferred-split-and-readiness-lock-decision.md
predecessor_plan: docs/plans/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-plan.md
sibling_plan: docs/plans/2026-09-12-intercom-go-stage-persistence-enforcement-plan.md
---

# Implementation Plan — Mechanical Enforcement of the Stage Default-Branch Contract (Strand B)

**Feature**: `020-F` · **Shipment**: `019-S` (**blocked**, readiness-locked, **not claimed**)
**Requires plan hardening**: **yes**
**Status**: scope and dependencies defined; **NOT implementation-ready**. See §5.

> **Created by the 2026-09-12 rev-17 remediation.** This strand was carved out of the former
> `019-F`, which mixed persistence semantics with reusable enforcement infrastructure. All five
> tasks were moved with backlogit's supported `adopt` operation, which re-IDs and records
> `origin_feature`. Nothing was destructively removed.

## 1. Why this feature exists

The contract corrected by `018-F` lives in **autoharness-generated files** —
`.github/policies/workflow-policies.md` and `.github/agents/_stage.agent.md`. A regeneration can
silently revert it. `018-F` accepts that as an explicit, tracked interim residual risk
(§R16.9 of the predecessor plan). This feature removes the risk by making the corrected contract
**mechanically enforced and regeneration-resistant**.

## 2. Why Strand B is separate from Strand A

Strand A (`019-F` / `018-S`) is about **what Stage and the Orchestrator do at handback** —
handback emission, the commit step, the no-shipment terminal, Orchestrator branch discovery.
Strand B is about **reusable infrastructure that proves the contract stays corrected** — detector,
fixture corpus, event parsing, CI wiring, regeneration proof. Different consumers, different blast
radius, different review needs. After the split **neither strand depends on the other**: this one
depends only on `018.008-T` and its own internal chain.

## 3. Dependency on current work

`020-F` **depends on** `018-F`; `018-F` depends on nothing here. Verified: no `current → future`
edge and no cycle anywhere in the graph.

| Task | Size | Blocks on | Rationale |
|---|---|---|---|
| `020.001-T` | S | `018.008-T` | The corpus must encode the **corrected** contract; building it first would freeze the defective shape |
| `020.002-T` | S | `020.001-T` | The driver consumes that corpus and manifest |
| `020.003-T` | M | `020.002-T` | The detector is implemented against a red driver suite |
| `020.004-T` | S | `020.003-T`, `018.008-T` | There must be a working detector to wire, and a corrected contract for the ledger entry to describe |
| `020.005-T` | S | `020.004-T` | There must be a wired blocking gate to mutate against |

```text
018.008-T (current unit — external prerequisite)
   └─> 020.001-T ─> 020.002-T ─> 020.003-T ─> 020.004-T ─> 020.005-T
                                                   ^
                                            018.008-T ─┘
```

The rev-1 cross-strand edges `020.004-T → 019.004-T` and `020.004-T → 019.007-T` were **removed**
in the rev-17 remediation: CI wiring semantically requires the detector and the corrected
contract, not the Strand A persistence tasks.

## 4. Scope

| Task | Summary |
|---|---|
| `020.001-T` | Direct-push fixture corpus and manifest — **test data only** |
| `020.002-T` | Table-driven fixture-suite driver over a **stubbed** detector |
| `020.003-T` | Structure-aware direct-push detector over a **closed** construct set |
| `020.004-T` | Blocking CI lint-job wiring, divergence ledger, CODEOWNERS |
| `020.005-T` | Coupling verification with an **executed** mutation proof |

**Out of scope**: everything in `019-F` (handback, commit step, no-shipment terminal, Orchestrator
arms), and any change to `018-F`'s reduced contract.

### 4.1 Scope discipline — the original failure mode

Detector over-generalization is what inflated the retired eight-task unit. Two standing
constraints:

* The detector's construct set is **closed and enumerated**, and every member maps to a **named
  property** of the corrected `018-F` contract. No general-purpose policy-linting platform.
* The fixture corpus is sized to **prove the corrected contract**, not to build a fixture
  platform. A fixture that does not map to a named contract property does not belong in it.

Detection must be **structure-aware** — walking headings, table rows, list items and fenced blocks
— never a regex sweep over raw prose, which produces false verdicts on quoted, historical or
example text.

## 5. BLOCKING PREREQUISITE — the harness execution model

This feature has **five test-bearing tasks**. P-004 requires every generated test function RED
before implementation; Ship Step 4.3 requires the full `go test ./...` GREEN after every task; and
**Ship Step 2 item 1 harnesses every `queued` task of the covering FEATURE in one batch**. These
are jointly unsatisfiable here.

The decision is tracked by **`021.001-T`**, which enumerates the three admissible options
(single-task features/shipments; machine-enforced one-queued-child serialization proven against
Ship intake and P-015; or a reviewed contract amendment making harness selection shipment-scoped).
**No mechanism is selected here.**

**Do not resurrect the rev-11..15 manifest / terminal-witness / `-gatetask` / MP0–MP6 machinery
without explicit re-deliberation.** It is non-governing provenance in the predecessor plan's
historical appendix.

## 6. Mechanical readiness lock

`019-S` is **not claimable**, and not merely labelled so:

* `019-S` is `status: blocked` — Orchestrator Step 2 triggers on `queued` shipments only, and
  Ship Step 0.5 item 1b rejects a shipment that is neither `queued` nor `active`.
* `019-S` carries a `blocks` dependency edge to readiness-gate shipment `020-S`, which is itself
  `blocked` and never executed. Orchestrator Step 2's explicit pre-claim re-check treats an
  unshipped blocking predecessor as a **hard** eligibility gate regardless of queue position.

Verified: `backlogit queue view --type shipment` lists only `017-S`.

**Unlock procedure (Stage only).** After `021.001-T` records a reviewed decision:
`backlogit dep remove 019-S 020-S` → restructure into the shipment shape that decision requires →
`backlogit move 019-S --status queued` and re-queue member tasks → run `impl-plan`,
`plan-harden` (**required**) and `plan-review`. Removing the edge without the review is a P-006
violation. Ship and the Orchestrator must not remove these edges.

## 7. Risks

| ID | Risk | Mitigation |
|---|---|---|
| R1 | Harness execution model unresolved → repeat of the revs 10–15 failure | §5 is a hard gate; `020-S` enforces it mechanically |
| R2 | Detector over-generalization — the original scope driver | §4.1: closed construct set, corrected-contract corpus, targeted assertions |
| R3 | Regex-over-prose detection produces false verdicts on quoted or historical text | `020.003-T` AC-2/AC-4 require structure-aware walking and a corpus fixture proving quoted text does not trigger |
| R4 | A wiring test that never observes a failure proves nothing | `020.005-T` requires an **executed** revert→fail→restore→pass pair, parsed from a machine-readable event stream, over a workspace copy with byte-identical before/after tree snapshots |
| R5 | Interim regeneration vulnerability while this strand is deferred | Accepted and tracked; see §R16.9 of the predecessor plan |

## 8. Task provenance

| Former ID | New ID | `origin_feature` |
|---|---|---|
| `019.001-T` | `020.001-T` | `019-F` |
| `019.002-T` | `020.002-T` | `019-F` |
| `019.003-T` | `020.003-T` | `019-F` |
| `019.005-T` | `020.004-T` | `019-F` |
| `019.006-T` | `020.005-T` | `019-F` |

Each task body was **fully rewritten** in the rev-17 remediation. The previous bodies carried the
retired rev-11..15 machinery as **active acceptance criteria** plus a stale parent (`018-F`),
stale predecessor IDs and a stale `017-S` composition. All of that is withdrawn and
**non-governing**; it is preserved unedited in git at
`git show 14d64e4:.backlogit/queue/019.00N-T.md`.

## 9. Next steps

1. Resolve `021.001-T` (harness execution model) under deliberation and plan review.
2. Restructure `020-F` / `019-S` per that decision.
3. Run `impl-plan`, then `plan-harden` (**required**) and `plan-review`.
4. Only then remove the `019-S → 020-S` edge and re-queue.

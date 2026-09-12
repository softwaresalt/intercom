---
title: "Implementation Plan — Mechanical Enforcement of the Stage Default-Branch Contract (Strand B)"
date: 2026-09-12
revision: 2
status: scope-defined-not-implementation-ready
agent: Stage
governs: feature 020-F (no shipment)
source: docs/decisions/2026-09-12-intercom-go-deferred-split-and-readiness-lock-decision.md
predecessor_plan: docs/plans/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-plan.md
sibling_plan: docs/plans/2026-09-12-intercom-go-stage-persistence-enforcement-plan.md
---

# Implementation Plan — Mechanical Enforcement of the Stage Default-Branch Contract (Strand B)

**Feature**: `020-F` (**blocked**) · **Shipment**: **none** — the rev-17 candidate `019-S` was archived in rev 18
**Requires plan hardening**: **yes**
**Status**: scope and dependencies defined; **NOT implementation-ready**. See §5.

> **Created by the 2026-09-12 rev-17 remediation; corrected by rev 18.** This strand was carved out
> of the former `019-F`, which mixed persistence semantics with reusable enforcement infrastructure.
> All five tasks were moved with backlogit's supported `adopt` operation, which re-IDs and re-parents
> the task. Nothing was destructively removed. **Rev 18 retires the invalid shipment-status readiness
> lock** — `blocked` is not a valid backlogit shipment status, so candidate shipment `019-S` and
> lock-token shipment `020-S` were archived and this strand now has **no shipment at all** — and
> **corrects the adoption-provenance wording** (§8). The rev-1 text is in git at `9e44bfe`.

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

## 6. No live shipment — this is the enforcement (rev 2)

`020-F` is **`status: blocked` and belongs to no shipment at all.**

**Why the rev-1 lock was invalid and was retired.** Rev 1 made `019-S` `status: blocked` and gave it
a `blocks` edge to lock-token shipment `020-S`. **`blocked` is not a valid backlogit shipment
status** — the enum is exactly `queued | active | shipped | abandoned`
(`.backlogit/header-def.yaml`; `docs/compound/2026-05-07-backlogit-shipment-status-constraints.md`).
Verified by execution against backlogit 1.10.1: `backlogit move <shipment> --status blocked` exits
**0** (the known generic-mover defect that wrote those values), `--status shipped` exits **9**, and
`--status abandoned` from `queued` is rejected by `validate_status_transition`. A lock token that can
exist only in an invalid state is not enforcement.

**There is no safe valid queued-shipment readiness lock** in current backlogit semantics unless a
genuine predecessor shipment is actually intended to ship: a `queued` shipment is by definition
claimable, and `blocked → queued` is not a supported shipment unlock transition. **This supersedes
the earlier request for a queued future shipment** — the adversarial gate proved no valid safe queued
representation exists, and reliability and safety take precedence.

**What replaces it.** The `019-S → 020-S` edge was removed, then `019-S` and `020-S` were
**archived** with the official non-destructive `backlogit archive` operation. **Nothing was
deleted**: the archived records keep their IDs, titles and manifests, and carry
`archived_status: blocked` as honest provenance of the defect-written value.

With no shipment, there is nothing for Orchestrator Step 2 to select and nothing for Ship Step 0.5
to claim: both operate on shipments. Verified: `backlogit queue view` lists `017-S` as the only
shipment in the workspace. Verified by execution that archival costs no future capability —
`backlogit archive <shipment>` does not cascade to members, and members of an archived shipment are
freely re-assemblable into a new shipment.

**Reconstitution (Stage only).** After `021.001-T` records a reviewed decision: restructure `020-F`
into the shape that decision requires → run `impl-plan`, `plan-harden` (**required**) and
`plan-review` → **only then create a fresh shipment** over the reviewed shape and move its member
tasks to `queued`. Creating a shipment before plan review is a P-006 violation.

## 7. Risks

| ID | Risk | Mitigation |
|---|---|---|
| R1 | Harness execution model unresolved → repeat of the revs 10–15 failure | §5 is a hard gate; no shipment exists for this feature, so nothing is claimable |
| R2 | Detector over-generalization — the original scope driver | §4.1: closed construct set, corrected-contract corpus, targeted assertions |
| R3 | Regex-over-prose detection produces false verdicts on quoted or historical text | `020.003-T` AC-2/AC-4 require structure-aware walking and a corpus fixture proving quoted text does not trigger |
| R4 | A wiring test that never observes a failure proves nothing | `020.005-T` requires an **executed** revert→fail→restore→pass pair, parsed from a machine-readable event stream, over a workspace copy with byte-identical before/after tree snapshots |
| R5 | Interim regeneration vulnerability while this strand is deferred | Accepted and tracked; see §R16.9 of the predecessor plan |

## 8. Task provenance (corrected rev 18)

| Former ID | New ID | Structured `origin_feature` | Prose-only re-parent history |
|---|---|---|---|
| `019.001-T` | `020.001-T` | `018-F` | `018-F` → `019-F` → `020-F` |
| `019.002-T` | `020.002-T` | `018-F` | `018-F` → `019-F` → `020-F` |
| `019.003-T` | `020.003-T` | `018-F` | `018-F` → `019-F` → `020-F` |
| `019.005-T` | `020.004-T` | `018-F` | `018-F` → `019-F` → `020-F` |
| `019.006-T` | `020.005-T` | `018-F` | `018-F` → `019-F` → `020-F` |

**`origin_feature` records the ORIGINAL/ROOT origin feature and is preserved across successive
adoptions** — it is **not** rewritten to the immediate predecessor. The stored value `018-F` is
therefore **correct**, and the **intermediate** re-parent step through `019-F` is carried in
**prose only**, here and in each task body, because no structured field records it.

**Two rev-17 provenance claims are WITHDRAWN as inaccurate.** (1) The statement that `adopt`
"records `origin_feature: 019.001-T -> 020.001-T`" conflated an **ID remap** with a **field value**.
(2) The rev-17 table asserting `origin_feature` = `019-F` for these five tasks **contradicted the
stored frontmatter**, which reads `018-F`.

Each task body was **fully rewritten** in the rev-17 remediation. The previous bodies carried the
retired rev-11..15 machinery as **active acceptance criteria** plus a stale parent (`018-F`),
stale predecessor IDs and a stale `017-S` composition. All of that is withdrawn and
**non-governing**; it is preserved unedited in git at
`git show 14d64e4:.backlogit/queue/019.00N-T.md`.

## 9. Next steps

1. Resolve `021.001-T` (harness execution model) under deliberation and plan review.
2. Restructure `020-F` per that decision.
3. Run `impl-plan`, then `plan-harden` (**required**) and `plan-review`.
4. **Only then** create a fresh shipment over the reviewed shape and move member tasks to `queued`.

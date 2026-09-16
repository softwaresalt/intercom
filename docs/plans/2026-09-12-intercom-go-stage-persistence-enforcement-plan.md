---
title: "Implementation Plan — Stage Artifact Persistence Route Execution (Strand A)"
date: 2026-09-12
revision: 3
status: scope-defined-not-implementation-ready
agent: Stage
governs: feature 019-F (no shipment)
source: docs/decisions/2026-09-12-intercom-go-stage-policy-gap-re-plan-decision.md
split_decision: docs/decisions/2026-09-12-intercom-go-deferred-split-and-readiness-lock-decision.md
predecessor_plan: docs/plans/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-plan.md
sibling_plan: docs/plans/2026-09-12-intercom-go-stage-contract-enforcement-platform-plan.md
---

# Implementation Plan — Stage Artifact Persistence Route Execution (Strand A)

**Feature**: `019-F` (**blocked**) · **Shipment**: **none** — the rev-17 candidate `018-S` was archived in rev 18
**Requires plan hardening**: **yes**
**Status**: scope and dependencies defined; **NOT implementation-ready**. See §5.

> **Revision 3 (2026-09-12).** Rev 2 narrowed this plan to **Strand A only**; the reusable
> mechanical-enforcement platform moved to `020-F`, planned in
> `docs/plans/2026-09-12-intercom-go-stage-contract-enforcement-platform-plan.md`. **Rev 3 retires the
> invalid shipment-status readiness lock**: `blocked` is not a valid backlogit shipment status, so
> candidate shipment `018-S` and lock-token shipment `020-S` were archived and this strand now has
> **no shipment at all**. The rev-1 text is in git at `14d64e4`, the rev-2 text at `9e44bfe`.

## 1. Why this feature exists

`018-F` delivers the **prohibition** half of the Stage default-branch policy gap: Stage may not
commit on the default branch, may create `chore/stage-{scope-slug}`, and must pass a fail-closed
Step 1.9 gate before any tracked mutation. It deliberately delivers **no persistence execution**.

Until this feature lands:

* a Stage session terminates at a **named fail-closed handback**, `STAGE_ARTIFACTS_UNCOMMITTED`,
  with bounded artifacts **uncommitted** on the Stage branch — uncommitted *because* commit and
  handback automation is deferred to `019.009-T` / `019.007-T`;
* the **Orchestrator must halt** there and **must not** execute its Step 1.5 item 3
  uncommitted/unpushed arm. That arm commits Stage artifacts and its sub-item **3(e) attempts a
  direct push to `main` first**, both of which contradict the Orchestrator's own installed
  orchestration-only Role statement (*"You do NOT triage stash entries yourself. You do NOT write
  code or create PRs yourself."*) and the very prohibition `018-F` installs;
* completion is **operator-only**: only the operator may verify branch/worktree/artifact state,
  commit the four allowed Stage roots, push the side branch, open and approve a **merge-commit**
  staging PR (P-009), update the default branch, and re-verify the manifest on `origin/main`;
* **the current dark run therefore pauses at that operator checkpoint and cannot finish
  autonomously** while the operator is away.

This feature replaces the operator-only sequence with a specified, executable one. `019.004-T` is
where any Orchestrator authority would be **deliberated, scoped and reviewed** — it does not exist
today, and `018-F` creates none.

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

## 6. No live shipment — this is the enforcement (rev 3)

`019-F` is **`status: blocked` and belongs to no shipment at all.**

**Why the rev-2 lock was invalid and was retired.** Rev 2 made `018-S` `status: blocked` and gave it
a `blocks` edge to lock-token shipment `020-S`. **`blocked` is not a valid backlogit shipment
status** — the enum is exactly `queued | active | shipped | abandoned`
(`.backlogit/header-def.yaml`; `docs/compound/2026-05-07-backlogit-shipment-status-constraints.md`).
Verified by execution against backlogit 1.10.1: `backlogit move <shipment> --status blocked` exits
**0** (the known generic-mover defect that wrote those values), `--status shipped` exits **9**, and
`--status abandoned` from `queued` is rejected by `validate_status_transition`. A lock token that can
exist only in an invalid state is not enforcement.

**There is no safe valid queued-shipment readiness lock** in current backlogit semantics unless a
genuine predecessor shipment is actually intended to ship: a `queued` shipment is by definition
claimable, and `blocked → queued` is not a supported shipment unlock transition. Leaving an
implementation-unready future shipment claimable is unacceptable. **This supersedes the earlier
request for a queued future shipment** — the adversarial gate proved no valid safe queued
representation exists, and reliability and safety take precedence.

**What replaces it.** The `018-S → 020-S` edge was removed, then `018-S` and `020-S` were
**archived** with the official non-destructive `backlogit archive` operation. **Nothing was
deleted**: the archived records keep their IDs, titles and manifests, and carry
`archived_status: blocked` as honest provenance of the defect-written value — a value the
`shipment-reconcile` contract already enumerates and handles.

With no shipment, there is nothing for Orchestrator Step 2 to select and nothing for Ship Step 0.5
to claim: both operate on shipments. Verified: `backlogit queue view` lists `017-S` as the only
shipment in the workspace. Verified by execution that archival costs no future capability —
`backlogit archive <shipment>` does not cascade to members, and members of an archived shipment are
freely re-assemblable into a new shipment.

**Reconstitution (Stage only).** After `021.001-T` records a reviewed decision: restructure `019-F`
into the shape that decision requires → run `impl-plan`, `plan-harden` (**required**) and
`plan-review` → **only then create a fresh shipment** over the reviewed shape and move its member
tasks to `queued`. Creating a shipment before plan review is a P-006 violation. Ship and the
Orchestrator never perform any of this.

## 7. Risks

| ID | Risk | Mitigation |
|---|---|---|
| R1 | Harness execution model unresolved → repeat of the revs 10–15 failure | §5 is a hard gate; no shipment exists for this feature, so nothing is claimable |
| R2 | `019.008-T` converts a loud failure into a silent success | Load-bearing constraint in the task: P-003 violations, errored harvests and unformable shipments must still **halt** and must be provably non-recordable as `no-shipment` |
| R3 | `019.004-T` silently widens the Orchestrator role boundary | AC-2 forbids creating new authority; any widening is a separate deliberated P-010 amendment. **`018-F` creates none, and rev 18 withdrew the claim that Step 1.5 item 3 already supplies it** — that arm's direct-`main` attempt and commit action sit outside the Orchestrator's orchestration-only Role statement |
| R4 | Interim operator-only halt mistaken for a defect | §1 and §R16.5 of the predecessor plan state it is the intended terminal, name the actor (operator only), and state plainly that the dark run cannot finish autonomously past it |

## 8. Next steps

1. Resolve `021.001-T` (harness execution model) under deliberation and plan review.
2. Restructure `019-F` per that decision.
3. Run `impl-plan`, then `plan-harden` (**required**) and `plan-review`.
4. **Only then** create a fresh shipment over the reviewed shape and move member tasks to `queued`.
---
title: "Decision — Re-plan of the Stage Artifact Branch/PR Policy Gap release unit"
date: 2026-09-12
status: superseded-in-part
superseded_in_part_by: docs/decisions/2026-09-13-intercom-go-a-only-simplification-decision.md
superseded_decision: "4.2 — 017-S manifest shape; the live 13-member fully-covered-root shape governs"
agent: Stage
governs: feature 018-F, shipment 017-S
supersedes_scope_of: docs/plans/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-plan.md
plan: docs/plans/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-plan.md
deferred_plan: docs/plans/2026-09-12-intercom-go-stage-persistence-enforcement-plan.md
---

# Decision — Re-plan of the Stage Artifact Branch/PR Policy Gap release unit

## 1. Problem frame

Feature `018-F` / shipment `017-S` had grown to **eight executable tasks** (one `L`, two `M`,
five `S`) plus a covering feature. Its governing plan reached **432 KB across 15 revisions**, and
its deliberation **112 KB**. PR #54's remote is at rev 10; local planning history through rev 15
is unpushed and **failed adversarial review repeatedly**.

The operator's diagnosis — that fixture/harness infrastructure and the `L`-sized B3 task had
accumulated too much scope — is confirmed, but the investigation found a **deeper root cause**
than scope creep alone.

## 2. Root cause — a structural impossibility, not a drafting defect

The installed contract contains two requirements that are **jointly unsatisfiable** for any
shipment containing more than one test-bearing task:

| Source | Requirement |
|---|---|
| `P-004` (`.github/policies/workflow-policies.md`) | `go test ./...` exits non-zero **with expected failure markers for every test function** before `harness-ready` is applied |
| `_ship.agent.md` Step 4.3 | After **every** task, the **full** `go test ./...` must pass; on failure Ship returns to `build-feature` for a fix iteration |

With N > 1 test-bearing tasks: H0 makes all N red; task 1 goes green; functions 2..N remain red by
design; Step 4.3 fails; `build-feature` exhausts its five attempts attempting to implement future
work. **No task can complete.** No ordering of the tasks escapes this — it is arithmetic, not
sequencing.

Revisions 10–15 were successive attempts to *bridge* this rather than remove it:

* **rev 10** — eight surface predicates. Defective on two counts (self-disarming predicate;
  contract breach). Confirmed P0 unanimously by a four-model adversarial panel.
* **rev 11** — harness-state manifest + `-gatetask` selector + MP0–MP6 mutation proofs.
* **rev 12** — MP5 retired (it re-created the very Step 4.3 deadlock), terminal witness added.
* **rev 13** — two-stage selector precedence; control-state exemption.
* **rev 14** — H0 eligibility re-argued; citation corrected.
* **rev 15** — fixture snapshot equality; self-contained Git fixture.

Every one of those artifacts existed **only** to make an eight-task batch survive Step 4.3.

**Decision: remove the cause, not the symptom.** Reduce the release unit to **exactly one queued
task**. One task ⇒ one red function ⇒ one red→green transition ⇒ full suite green immediately.
The entire manifest/witness/selector/MP-series platform becomes **unnecessary rather than
merely simplified**.

## 3. Options considered

| # | Option | Verdict |
|---|---|---|
| A | Keep 8 tasks, simplify the manifest | **Rejected** — does not address the impossibility; any N>1 still deadlocks |
| B | Amend `P-002`/`P-004`/Step 4.3 to permit multi-task red batches | **Rejected for now** — expressly out of scope for `018-F`; amending the safety contract to fit a plan inverts the control direction. Recorded as a legitimate future option for `019-F` |
| C | **Reduce to one task; defer the rest to a new feature/shipment** | **CHOSEN** |
| D | Reduce to one task and *delete* the deferred work | **Rejected** — the enforcement platform is genuinely valuable; deleting it discards reviewed analysis |

### Why C over B

Option B would change the rule that *detects* the problem in order to accommodate the plan that
*triggered* it. Option C keeps every safety gate intact and proves the reduced unit under the
contract **exactly as installed**. B remains available to `019-F` as an explicit, separately
deliberated decision.

## 4. Decisions

> ### ⚠️ SUPERSEDED IN PART — decision 2 no longer describes the live manifest
>
> **Decision 2 below (`017-S` manifest becomes `[018-F, 018.008-T]`) is SUPERSEDED on
> MEMBERSHIP and MUST NOT be used as the manifest of record — but its core principle,
> that the covering feature belongs INSIDE the manifest, is now VINDICATED.**
>
> **Rev 2 of this note (2026-09-13).** The intervening "12-member task-only" shape is
> itself **withdrawn**. Governing authority is now
> `docs/decisions/2026-09-13-intercom-go-a-only-simplification-decision.md`
> (verdict `SIMPLIFICATION_VALID`).
>
> | | This decision (stale) | Intervening task-only shape (WITHDRAWN) | Live / governing |
> |---|---|---|---|
> | `017-S` manifest | `[018-F, 018.008-T]` (fully-covered root, 2 members) | 12-member TASK-ONLY, `018-F` excluded | **13-member FULLY-COVERED ROOT**: `018-F` + `018.008-T` + `018.001-T` + 3 `-ST` + `018.002-T` + 3 `-ST` + `018.003-T` + 2 `-ST` |
> | Member states | both live | 11 pre-archived, 1 live | **11 pre-archived, 1 live task (`018.008-T`), 1 live root (`018-F`)** |
> | Covering feature | `018-F` **inside** the manifest | `018-F` **outside**; resolved via `parent_id` | **`018-F` INSIDE the manifest.** `backlogit shipment get 017-S` populates `covering_feature: 018-F` |
> | Close path | `CASCADE` (fully-covered root) | `TASK_ONLY_FINALIZE` | **`CASCADE`** — the **existing** P-015 exception, enabled by shipment A's Step 6.1(a1) |
>
> **Why the membership differs from decision 2.** Decision 2's 2-member manifest omitted
> the 11 pre-archived legacy descendants. A fully-covered root requires the manifest to
> contain the root **and its complete descendant set at every depth**, so the correct
> shape is **13**, not 2. Decision 2 had the right principle and the wrong membership.
>
> **Why the task-only shape was withdrawn.** It rested on the clause *"no Ship step ever
> moves a covering FEATURE to done"*. Shipment A (`021-S`) introduces exactly that step,
> so the exclusion built on it falls. `017-S` never needs `TASK_ONLY_FINALIZE`, which is
> now **deferred generalized platform work** (`023-F`/`024-F` `blocked`; shipments
> `022-S`/`023-S` archived).
>
> Superseding authorities, in order: the **A-only simplification decision**, plan A rev 3
> (`docs/plans/2026-09-12-intercom-go-ship-feature-completion-foundation-plan.md`), and
> the umbrella evidence document at rev 8. Read decision 2 as a record of what was decided
> on 2026-09-12 — not as a current instruction.
>
> Decisions **1, 3, 4, 5 and 6 below remain in force.** Only the manifest shape in
> decision 2 changed. Nothing here alters `018-F`'s reduction to a single executable task,
> the `019-F`/`018-S` deferral, or the dependency direction.

1. **`018-F` is reduced to one executable task**, `018.008-T` (M / medium), covering P-010,
   the Stage Role Boundary rows, and the new Step 1.9 branch gate.
2. **`017-S` manifest becomes `[018-F, 018.008-T]`.** — **SUPERSEDED, see the note above.**
3. **New feature `019-F` + new shipment `018-S`** carry the deferred persistence route and the
   mechanical enforcement platform (nine tasks).
4. **Dependency direction is future → current.** `019.001-T`, `019.007-T`, `019.009-T` block on
   `018.008-T`. `018-F` depends on nothing in `019-F`.
5. **`018-S` is not claimed in the current run** and is **not implementation-ready**: it must
   first resolve the harness-lifecycle question recorded in its plan.
6. **No generalized fixture/manifest/witness machinery enters the current unit.**

## 5. Scope boundary of the reduced unit

**In scope** — the *prohibition* half of the gap:

* Stage may not commit or push on the default branch (policy **and** agent contract).
* Stage may create and check out `chore/stage-{scope-slug}`.
* A **Step 1.9** gate requires that branch to exist before any tracked artifact mutation, under a
  **categorical** deferral rule.
* Orchestrator/operator — never Stage — pushes and opens/merges the staging PR.

**Deferred** — the *execution* half and the enforcement platform: handback emission,
`stage_artifact_paths` derivation, no-shipment terminal reachability, the Step 5.7 artifact-commit
step, out-of-root path safety, Orchestrator branch discovery and post-merge verification, the
direct-push detector, fixture corpus, CI wiring, CODEOWNERS/ledger, and mutation proofs.

### Coherence of the reduced unit

The reduced unit is a **complete, independently valuable safety property**: *Stage cannot mutate
the default branch, and a dedicated artifact branch must exist before any tracked write.* It does
not depend on any deferred item to be true or testable.

## 6. Accepted residual risk

| Risk | Assessment |
|---|---|
| No mechanical CI enforcement until `019.005-T` | Contract is regeneration-vulnerable in the interim. **Accepted** — the prohibition still binds the agent contract, which is what role enforcement reads at mutation time |
| No automated persistence route until `019-F` | Operator/Orchestrator pushes and opens the PR manually. **Accepted** — identical to the interim behaviour already in use, and strictly safer than the gap being closed |

Neither risk is a regression: both describe capability **not yet added**, not protection removed.

## 7. Task disposition and traceability

Re-parenting used backlogit's supported `adopt` operation, which **re-IDs the item and rewrites
cross-references**, recording `origin_feature`. Nothing was destructively removed.

| Original | Disposition | New ID |
|---|---|---|
| `018.001-T`–`018.003-T` | Already **archived** before this session (provisional sub-epic containers). Left archived | — |
| `018.004-T` | Adopted to `019-F` | `019.001-T` |
| `018.005-T` | Adopted to `019-F` | `019.002-T` |
| `018.006-T` | Adopted to `019-F` | `019.003-T` |
| `018.007-T` | Adopted to `019-F` | `019.004-T` |
| `018.008-T` | **Retained and re-scoped** as the single current task | `018.008-T` |
| `018.009-T` (L) | Adopted to `019-F`, then **split three ways** | `019.007-T` + new `019.008-T`, `019.009-T` |
| `018.010-T` | Adopted to `019-F` | `019.005-T` |
| `018.011-T` | Adopted to `019-F` | `019.006-T` |

`018.009-T`'s eight items were split by contract boundary: items **(1)(2)(3)(7a)** — role
boundary, branch grant, PR actor note, Step 1.9 gate — folded into `018.008-T` (current);
item **(4)(5)** → `019.007-T`; item **(6)** → `019.008-T`; item **(7b)** → `019.009-T`.

No live obligation is duplicated: each item text exists in exactly one active task.

## 8. Defects found and repaired during the re-plan

1. **Inverted dependency.** `018.008-T` blocked on `018.006-T`, which the re-plan moved to the
   future feature — making the *current* unit depend on *deferred* work. Edge **removed**.
2. **Dangling shipment manifest.** `adopt` re-IDs items but does **not** rewrite shipment
   `custom_fields.items`. After adoption `017-S` still listed seven dead IDs and backlogit emitted
   `size composition: skipping unresolved member` for each. No supported operation repairs this
   (`shipment return-blocked` fails with `not found` on an unresolved ID; there is no
   `shipment remove`). Repaired by the operator-authorised fallback — direct `custom_fields.items`
   edit followed by immediate `backlogit sync`. **Recorded as a backlogit gap.**
3. **Registry/schema mismatch on `complexity`.** The backlog registry declares a `complexity`
   param and the CLI exposes `--complexity`, but `.backlogit/header-def.yaml` defines **no
   complexity field on the `task` type**, so the call fails validation. The original `018-F`
   SIZING NOTE was therefore **correct** and stands; complexity is carried as enum-validated
   prose. Adding the field would mutate a tooling config file **outside Stage's role boundary**
   and was not done. **Reporting defect only.**
4. **Misleading `size_composition` roll-up.** `017-S` reports three `unsized` members
   (`018.001-T`–`018.003-T`) that are **not** in its `items` array, because that view rolls up
   *feature children* rather than shipment membership. Cosmetic; documented in `018-F`.

## 9. Open questions for review

* **OQ-1** — Is the one-task-per-shipment consequence acceptable as a standing constraint, or
  should `019-F` pursue option B (amending `P-002`/`P-004`/Step 4.3)? This decision defers the
  question but does not foreclose it.
* **OQ-2** — Does `018.008-T` fit the 2-hour rule at three files? Assessed **yes** (pre-specified
  textual edits, one test scenario group, single skill domain), but it is the reduced unit's
  largest single judgement call.

---
title: "Implementation Plan — Stage handback record emission contract (025-F / 025.001-T)"
date: 2026-09-18
status: hardened
agent: Stage
source_document: docs/decisions/2026-09-18-intercom-go-harness-execution-model-decision.md
governs: feature 025-F; task 025.001-T
base_commit: 1bef819843771e028ce8f1cff4dfffdae920a43e
stage_branch: chore/stage-harness-execution-model-readiness-decision-defer
revision: 2
revision_note: "Rev 2 remediates plan-review attempt 1 (FAIL). F1: restructure performed and measured; adopt remapped 019.007-T → 025.001-T; blocked → queued recorded as owed action A1. F2: closure path corrected to CASCADE / FULLY_COVERED_ROOT; H-A4 and C3 replaced. F3: R1 mitigation mechanism corrected. F4: P-001 added to prechecks and operator decisions. F5: AC-4/AC-5 un-asserted. F6: executor named. F7: destructive signal PRESENT, H-A4 reclassified destructive. F8–F18 resolved."
---

<!-- plan-review-attempt: 2 -->

# Stage handback record emission contract

Source decision: `docs/decisions/2026-09-18-intercom-go-harness-execution-model-decision.md`
(**OPTION 1 — single-task features and shipments**).
This plan covers exactly one release unit: covering feature **`025-F`** holding exactly one
task, **`025.001-T`** (adopted from `019.007-T`; `backlogit adopt` remapped the ID).

## Problem Frame

`.github/agents/_orchestrator.agent.md` Step 1.5 (Staging Artifact Merge Gate) is specified to
consume a **staging handback record** produced by Stage. No producer exists. The installed
`.github/agents/_stage.agent.md` Step 6 (`Summary`, line 813, with its Pre-Summary Verification
Gate at line 815) emits a prose session summary only — no `stage_branch`, no commit pair, no
artifact path set, no machine-consumable `stage_outcome`.

Consequences measured on the installed contract:

* The Orchestrator's Step 1.5 has **no producer** for the record it is written to consume, so
  every pipeline run degrades to operator-mediated handoff.
* `019.004-T` (Orchestrator Step 1.5 branch discovery) cannot be specified until the record's
  shape exists — it is a declared dependency (`019.004-T → 025.001-T`).
* `019.009-T` (Stage artifact commit step) must emit `stage_head_commit` into a record whose
  field set is not yet defined — also a declared dependency (`019.009-T → 025.001-T`).

The unit is **contract text in one file, plus its harness**. It changes no Go production
source, no schema, no CI, no runtime surface.

## Backlog state — measured, not assumed

Read back from the live backlog on 2026-09-18 after the Stage restructure:

| Artifact | Type | Status | Parent | Note |
|---|---|---|---|---|
| `025-F` | feature | `queued` | *(root, `parent_id: null`)* | created by this session |
| `025.001-T` | task | `blocked` | `025-F` | adopted from `019.007-T`; **ID remapped by the tool** |
| `019-F` | feature | `blocked` | root | retains 3 tasks, unchanged |
| `019.004-T`, `019.008-T`, `019.009-T` | task | `blocked` | `019-F` | each `→ 025.001-T (blocks)`, tool-rewritten |
| `020-F` (+5), `023-F` (+1), `024-F` (+1) | — | `blocked` | — | untouched |
| `018.008-T` | task | archived `done` | — | `025.001-T`'s only upstream dependency — **satisfied** |

**`backlogit adopt` ID semantics — measured (resolves review finding F8).** The operation
returned `rewritten_artifact_ids: [018.008-T, 019.004-T, 019.008-T, 019.009-T]` and renumbered
the adopted artifact `019.007-T → 025.001-T`. All four dependency edges survived and were
re-pointed automatically; none was lost. Because the task ID now shares the covering feature's
prefix, `shipment-reconcile`'s ID-prefix covering-feature derivation and P-015's `parent_id`
walk **agree on `025-F`** — the divergence risk recorded in rev 1 as R3 is **eliminated**, not
tolerated.

**Sizing capability degradation (recorded per the structured-emission capability gate).**
`size: M` is written as a structured field with `size_source: agent` and
`size_ruleset_version: 2h-rule-v1`. **`complexity` could not be written as a structured field**:
`backlogit update 025.001-T --complexity medium` failed with *"artifact type `task` does not
define a complexity field"*. `complexity: medium` is therefore carried as enum-validated prose
in the task's implementation notes. Both axes are assigned; neither is derived from the other.

## Requirements Trace

| Source AC (`025.001-T`) | Implementation action | File |
|---|---|---|
| **AC-1** — Step 6 emits a required staging handback line carrying `stage_branch`, `stage_base_commit`, `stage_head_commit`, `stage_outcome`, `stage_artifact_paths`, and `shipment_id` when formed | Add a **"Staging Handback Record (REQUIRED OUTPUT)"** subsection to Step 6, after the Pre-Summary Verification Gate and before the summary bullet list, specifying the six fields as a required emission | `.github/agents/_stage.agent.md` |
| **AC-2** — `stage_artifact_paths` is a NON-EMPTY set of concrete repository-relative FILE paths (**no directory prefixes, no trailing slash**) spanning `base..head`, not only the tip commit | State the shape rule with **all four** prohibitions named explicitly: non-empty; file paths not directory prefixes; **no trailing slash**; `base..head` span not tip-only | same |
| **AC-3** — the contract names the exact derivation command including `--diff-filter=d` and the four `STAGE_ARTIFACT_ROOTS` pathspecs | Embed the literal command `git diff --name-only -z --diff-filter=d {stage_base_commit} {stage_head_commit} -- .backlogit/ docs/plans/ docs/decisions/ docs/memory/` | same |
| **AC-4** — Stage **ECHOES** `stage_base_commit`; the Orchestrator's retained pre-invocation value is authoritative | State the echo rule and the authority rule, and state that a **direct** Stage invocation (no Orchestrator bracket) still emits a usable base | same |
| **AC-5** — emission is documented as authority-neutral | State that emitting a branch name, a commit pair and a file list transfers **no** authority: Stage does not push, open, or merge anything | same |

Every AC maps to at least one action; every action maps to at least one AC. No action is
untraced. AC-2's trailing-slash prohibition is asserted by harness scenario **S3** (rev 1 omitted
it — review finding F11).

## Implementation Units

### U0 — Stage backlog restructure (**PERFORMED**, owner: Stage)

Recorded as a unit so the plan's preconditions are auditable rather than implied
(review finding F1). U0 is **not** Ship work and is already complete except for **A1**.

| Step | Action | State |
|---|---|---|
| U0.1 | Create root covering feature `025-F`, `status: queued` | **done** |
| U0.2 | `backlogit adopt 019.007-T --parent 025-F` | **done** — ID remapped to `025.001-T` |
| U0.3 | Verify all four dependency edges survived | **done** — verified present |
| U0.4 | Withdraw the task's stale `DEFERRED / BLOCKED / NO SHIPMENT / Do not execute` header; carry the pre-claim precondition and scope fence on the record | **done** |
| **A1** | `025.001-T` `blocked → queued`, **then** create the shipment over `[025-F, 025.001-T]` | **OWED — post-review only** |

**A1 is the sole action that satisfies invariant I2.** It is deliberately sequenced *after* the
review gate because `021.001-T` AC-5 requires it: *"Only after AC-1 through AC-4, Stage CREATES A
FRESH SHIPMENT over the reviewed shape and moves its member tasks to queued."* The task reaches
`queued` **before** it is ever a manifest member, so no manifest member is ever `blocked`.

### U1 — Add the Staging Handback Record subsection to Stage Step 6 (owner: Ship)

* **Files affected**: **2** — `.github/agents/_stage.agent.md` (contract text) and **one new
  harness test file** (added by `harness-architect`; rev 1 under-counted this — review finding F9).
  Both are under the 2-hour rule's "fewer than 3 files" bound.
* **Production Go functions touched**: 0.
* **Position**: inside `### Step 6: Summary`, strictly **after**
  `#### Pre-Summary Verification Gate (NON-NEGOTIABLE)` (line 815) and strictly **before** the
  existing summary bullet list. Step 5.6 (line 790) and the Pre-Summary gate are **not** modified.
* **Heading level**: use `####` (peer of the Pre-Summary gate). The new subsection must be placed
  so the pre-existing summary bullets remain under the heading that owns them today and are **not**
  silently re-parented (review finding F17); verify by reading the rendered section order after the
  edit.
* **Changes**:
  1. New subsection heading `#### Staging Handback Record (REQUIRED OUTPUT)`.
  2. The six-field record definition (AC-1), with `shipment_id` explicitly conditional on a
     shipment having been formed and `stage_outcome ∈ {shipment, no-shipment}`.
  3. The `stage_artifact_paths` shape rule (AC-2) with **all four** prohibitions.
  4. The literal derivation command (AC-3).
  5. The echo/authority rule for `stage_base_commit` (AC-4) including the direct-invocation case.
  6. The authority-neutrality statement (AC-5).
* **Execution posture**: **test-first**. The harness asserts the contract text's required
  properties and is red before the text exists.
* **Atomic milestone**: `go test ./...` green with the new harness passing, and the five ACs each
  provable by a distinct assertion.
* **Width isolation**: single domain — contract/documentation text plus its own assertions. No Go
  production code.
* **2-hour rule**: 2 files, 0 production functions, **3 harness scenarios** (strictly fewer than
  4, per the constitution's bound — rev 1's "≤ 4 … Satisfied" was drift; review finding F18).
  **Satisfied.**

**Scope fence (non-negotiable).** U1 must **not** touch Step 5.5, Step 5.6, or the Pre-Summary
Verification Gate's existing `shipment_id` requirement. Relaxing that gate for the `no-shipment`
outcome is `019.008-T`'s scope and is explicitly **out of scope here**. A change to Step 5.5/5.6
semantics in this unit is a scope-fence breach and must fail review.

## Dependency Graph

```text
018.008-T (archived, done)
      │
      ▼
025.001-T  ── U1 ──►  025-F shipped
      │
      ├────────────► 019.008-T   (blocked, stays in 019-F)
      ├────────────► 019.009-T   (blocked, stays in 019-F)
      └────────────► 019.004-T   (blocked, stays in 019-F)
```

Single executable unit ⇒ no internal ordering. No cycles. `025.001-T`'s only upstream dependency,
`018.008-T`, is **archived done** (merge `5d69c727`, PR #58), so the unit has zero unfinished
dependencies. The three downstream edges were rewritten by the tool during U0.2 and verified.

## Decisions and Rationale

| # | Decision | Rationale |
|---|---|---|
| D1 | One task per covering feature (`025-F` ⊃ `025.001-T` only) | Instantiates Option 1. Bounds the tree to **at most one** harnessed-but-unimplemented test function, which is what Step 4.3's full-suite gate needs to go green after the single task. Decision §3.1. |
| D2 | Adopt rather than create a new task | Preserves the ACs, sizing, `origin_feature` provenance and all inbound edges. `backlogit adopt` is the tool's designated re-parent operation; it additionally renumbered the ID, which removed the rev-1 R3 derivation risk. Creating a duplicate would have broken traceability and orphaned the edges. |
| D3 | `019-F` is **not** dissolved | Only this task passes the Option 1 coherence test today. The other three remain genuinely blocked on it; dissolving `019-F` now would produce wrapper features around still-blocked tasks — exactly what Option 1 forbids. |
| D4 | `020-F` untouched | Its five tasks are fragments of one mechanism (fixture corpus → driver → detector → CI wiring → mutation proof); none is independently releasable. Decision §4. |
| D5 | Manifest is `[025-F, 025.001-T]`, not task-only | A fully-covered root **avoids** manufacturing the genuinely-task-only shipment that is `023-F`/`024-F`'s FIRED TRIGGER, preserving their intentional block, and yields the strongest closure guarantee (no protected set arises). Decision §3.2(c), §3.1. |
| D6 | Harness asserts contract **properties**, not byte-equality | Byte-equality over agent markdown is brittle against unrelated edits and would false-red on every future Stage template change. Property assertions are stable and still falsifiable. |
| D7 | **Closure is the `CASCADE` path, accepted explicitly** *(new in rev 2, review finding F2)* | `shipment-reconcile` SKILL.md:378 maps *"Every feature member is a root, fully covered at every depth, set-equal to the manifest"* → `CASCADE` / `FULLY_COVERED_ROOT`, and SKILL.md:576-577: *"the bound verdict alone selects the branch."* Prescribing safe-close would prescribe a branch the classifier does not return. The plan therefore adopts the cascade path and its two-set post-condition gate, rather than asserting a path around it. |

## Risks and Caveats

| ID | Risk | Severity | Mitigation |
|---|---|---|---|
| R1 | **Silent harness bypass on claim** (`DB12DA37`). The shipment claim auto-activates `025.001-T` `queued → active`; Ship Step 2 item 1 then lists **zero** `queued` tasks, scaffolds nothing, and item 4's confirmation is vacuously true over the empty set — so Step 4 executes an **unharnessed** task, a P-004 red-phase bypass. | **HIGH — blocking** | **H-A1 pre-claim precondition**: `harness-architect` is run over `025.001-T` and `harness-ready` applied **before** the claim. *Corrected mechanism (review finding F3):* this does **not** route the task into Step 2's "Already harnessed" partition — on the claim route the task is absent from item 1's list entirely and is never partitioned. It works by making Step 2 **irrelevant**: a real red harness exists on the artifact before execution, so the property P-004 protects holds regardless of Step 2's bookkeeping. Step 2's vacuity is **not** repaired and is routed to the Option 3 amendment item. |
| R2 | **Residual P-002 deviation** persists even with R1 mitigated, because the claim still transitions the task before Ship's Step 2 runs. | Medium | Disclosed, not authorised here. If the operator elects to proceed without H-A1, the `D-D6` precedent applies: **explicit recorded operator authorisation**, P-005 event with `violation_policy: P-002`, `skip_policy: P-002`; **absence, ambiguity, or silence ⇒ HALT, zero mutation.** Stage grants nothing. |
| R3 | ~~ID-prefix vs `parent_id` derivation~~ — **CLOSED**. | — | Eliminated by measurement: `backlogit adopt` remapped the ID to `025.001-T`, so both derivations resolve to `025-F`. The underlying contract divergence is no longer reachable from this manifest; carried as Option 3 candidate scope. |
| R4 | **Scope-fence breach** — implementer relaxes the Pre-Summary gate's `shipment_id` requirement while adding the subsection, silently absorbing `019.008-T`. | Medium | U1's scope fence; harness scenario **S3** asserts the Pre-Summary gate's `shipment_id` requirement is unchanged (negative assertion). |
| R5 | **Brittle harness** — over-specified text assertions false-red on unrelated Stage edits. | Low | D6: property assertions only; explicitly no byte-equality over the file or whole section. |
| R6 | **Self-modification.** The unit edits `.github/agents/_stage.agent.md` — the contract the Stage agent itself executes. | Medium | Ship, not Stage, implements it (P-010). The change is additive, does not alter Step 1.9 / 5.5 / 5.6 / the Pre-Summary gate, and takes effect only on the **next** Stage session. |
| R7 | **P-001 halt after the one-way-door claim** *(new in rev 2, review finding F4)*. `021-F` / `021.001-T` are `active`. P-001's gate point is Ship **Step 1** (pre-flight), which runs **after** the Step 0.5.4 claim — so a stale `active` top-level unit lets Ship claim and *then* halt. | **HIGH — blocking** | Stage moves `021.001-T` and `021-F` to a terminal state in this same session, before the shipment is claimable. Added as **environment precheck 1** and as operator decision **O3**. |
| R8 | **Closure is destructive and irreversible** *(new in rev 2, review findings F2/F7)*. The `CASCADE` path invokes the destructive `backlogit_ship_shipment` engine operation over the manifest subtree; the shipment claim is itself a one-way door. | **HIGH** | H-A4 reclassified `ActionRisk: destructive` with recorded operator approval, matching the `017-S` precedent for the identical operation. Backlog-state unwind procedure specified separately from `git revert`. Two-set `allowed_ids`/`required_ids` post-condition gate is the required check. |
| R9 | **Second, independent harness-bypass surface** *(new in rev 2, review finding F12)*. P-002's enforcement filters the ready queue to `harness-ready` tasks, but Ship Step 3 item 2's substitution replaces that queued-only membership with item 1's derived executable set, which selects on **status**, not on the label. | Low | Fully mitigated by H-A1 as a side effect: with `harness-ready` present pre-claim, both surfaces agree. Recorded so the mitigation's coverage is explicit rather than incidental; the contract inconsistency itself is Option 3 scope. |

## Plan Hardening Signals (REQUIRED)

| Signal | Present? | Justification |
|---|---|---|
| public API, schema, or contract change | **PRESENT** | The unit defines a new **inter-agent contract**: the staging handback record produced by Stage Step 6 and consumed by Orchestrator Step 1.5. Its field set becomes a compatibility surface for `019.004-T` and `019.009-T`. |
| security, auth, permission, or compliance-sensitive behavior | **PRESENT** | AC-5 is an explicit **authority-neutrality** assertion and AC-4 assigns authority over `stage_base_commit`. The unit sits inside the P-010 role-boundary surface `018-F` established; a wording slip could read as widening Stage or Orchestrator authority. |
| migration, backfill, destructive data/config action, or irreversible step | **PRESENT** *(corrected in rev 2 — review finding F7; rev 1 wrongly marked this Absent)* | Two irreversible actions: (a) the **shipment claim** is a one-way door that auto-transitions the task; (b) **closure runs the `CASCADE` path**, invoking the destructive `backlogit_ship_shipment` engine operation which archives the manifest subtree. `git revert` of the U1 commit does **not** unwind either; a distinct backlog-state unwind is required. |
| external integration, operator checkpoint, or external dependency | **PRESENT** | Three operator checkpoints: H-A1 (pre-claim `harness-ready`), H-A2 (P-002 authorisation if H-A1 is not taken), H-A4 (destructive closure approval). Plus the P-001 terminal-state precondition (R7). |
| high runtime, rollout, or rollback risk | Absent | No runtime surface is changed. Rollback of the *text* is a single-commit revert; the agent template is re-read from disk each session, so revert is immediately effective. Backlog-state rollback is covered under the destructive signal above, not here. |

**Requires plan hardening: yes**

## Runtime Verification and Closure

**Runtime surface changed: NONE.** No CLI, API, browser UI, or background job is touched. The
changed surface is an **agent-contract** surface whose consumer is the next Stage session.

| ID | Verification | Proves |
|---|---|---|
| C1 | `go vet ./...` and `go test ./...` green at HEAD after U1 | No regression; the new harness passes |
| C2 | Re-read `.github/agents/_stage.agent.md` and confirm `#### Staging Handback Record (REQUIRED OUTPUT)` sits strictly between the Pre-Summary Verification Gate and the summary bullet list, and that the pre-existing summary bullets were not re-parented | AC-1 positioning; Step 5.6 and the gate untouched (R4, F17) |
| C3 | **At closure, before any archive**: invoke `shipment-reconcile` in `mode: classify-close-path` and record the emitted `CLOSE_PATH_VERDICT`, `VERDICT_REASON`, `VERDICT_EVIDENCE` and `CLASSIFICATION_BINDING`. Assert the verdict is `CASCADE` / `FULLY_COVERED_ROOT`. **Any other verdict ⇒ HALT** — the manifest has drifted from the reviewed shape. *(Rev 1's C3 asked for a protected-set derivation check, which is unexecutable on the cascade path — review finding F2.)* | D7 holds; closure path is measured, not assumed |
| C4 | At closure, apply the cascade two-set post-condition gate: fail on `archived_ids − allowed_ids` (unexpected artifact) and, separately, on `required_ids − archived_ids` (required artifact not archived), where `required_ids = {025-F, 025.001-T, shipment record}` | R8's destructive operation is bounded and verified |
| C5 | Confirm `019.004-T`, `019.008-T`, `019.009-T` remain `blocked` in `019-F` and their `→ 025.001-T` edges still resolve after closure | D2/D3 held; no orphaning |
| C6 | Confirm `020-F`, `023-F`, `024-F` remain `blocked`; confirm no task-only shipment was created | Intentional blocks preserved; `023-F` trigger did not fire |

**Operational closure**: the unit closes when U1 is merged, C1–C6 are recorded, and
`[025-F, 025.001-T]` plus the shipment record are archived by the cascade path under C4's gate.
**Owner**: Ship (execution and closure); Stage (backlog shape only). **Validation window**: the
next Stage session, which exercises the new Step 6 emission for real.

---

## Plan Hardening

**Hardening required: YES** — four signals present (contract change; security/authority-sensitive
text; destructive/irreversible actions; operator checkpoints). Hardening was performed in full per
P-006; this plan was **not** passed to review on a pre-hardened body. Rev 2 re-hardens against the
attempt-1 review findings.

**Sources consulted**

* `.github/policies/workflow-policies.md` — P-001, P-002, P-004, P-006, P-010, P-015 (incl. the
  verified fully-covered-root exception and Amendment Log rows 1.19.0–1.24.0)
* `.github/agents/_ship.agent.md` — Step 0.5 item 1a, Step 1, Step 2, Step 3, Step 4.3
* `.github/skills/shipment-reconcile/SKILL.md` — Step 0(c) classification and the exhaustive
  condition → verdict mapping (:378), verdict binding (:576-577), Cascade Close Sub-Procedure,
  protected-set computation (:605-645)
* `docs/plans/2026-09-16-017-S-closure-canonical-contract.md` — rows **D-D6**, **D-M8**
* `.backlogit/archive/018.008-T.md` — the `harness-ready`-at-claim precedent (line 17)
* `.backlogit/stash.jsonl` — `DB12DA37`
* `docs/decisions/2026-09-18-intercom-go-stage-readiness-reconciliation-and-blocking-matrix.md`
  — G1/G2 blocking matrix
* Retired rev-11…rev-15 historical appendix in
  `docs/plans/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-plan.md` — read and
  confirmed NON-GOVERNING; **nothing from it is reintroduced**, in explicit or disguised form

### High-risk triggers and invariants to preserve

| Invariant | Must hold |
|---|---|
| **I1** | `025-F` holds **exactly one** task for the whole lifetime of this shipment. A second queued test-bearing task re-creates the G1 deadlock and voids the Option 1 proof. |
| **I2** | No manifest member is ever in `blocked` status — Ship Step 3 item 1's fail-closed rule would HALT the run. **Satisfied by A1's ordering**: `025.001-T` reaches `queued` before it becomes a manifest member. |
| **I3** | The manifest stays `[025-F, 025.001-T]`. It is the **closure membership record** and is **never mutated to make execution proceed**. Drift is caught by C3. |
| **I4** | Step 5.5 / Step 5.6 / the Pre-Summary gate's `shipment_id` requirement are **unchanged** by this unit (`019.008-T`'s scope). |
| **I5** | `019-F`, `020-F`, `023-F`, `024-F` remain `blocked`. The `023-F`/`024-F` FIRED TRIGGER must **not** fire — guaranteed by D5's fully-covered-root manifest. |
| **I6** | No harness function is ever executed for a task lacking a real red harness. Enforced pre-claim by H-A1, since Step 2 cannot enforce it on the claim route. |

### ProposedAction / ActionRisk entries carried forward

| Action | ActionRisk | Approval required |
|---|---|---|
| **H-A1** — Run `harness-architect` over `025.001-T` and apply `harness-ready` **before** the shipment claim | **HIGH** — omitting it yields a silent P-004 red-phase bypass (R1, R9) | **YES — operator.** *Executor named (review finding F6):* the **operator** invokes the `harness-architect` skill directly, as the always-available substitute actor. No installed surface defines a pre-claim harness slot for an agent (P-004 applies to `ship` via Step 2, which is post-claim; P-010 forbids Stage) — that gap is Option 3 candidate scope. Stage neither performs nor waives it. If it cannot be satisfied, **do not claim the shipment.** |
| **H-A2** — Claim shipment `{shipment_id}` (auto-activates `025.001-T`; residual P-002 deviation, R2; one-way door, R8) | **HIGH / irreversible** | **YES — explicit recorded operator authorisation** per the `D-D6` precedent, with a P-005 record (`violation_policy: P-002`, `skip_policy: P-002`). D-D6 records the gate label as `S-3` for `017-S`; the equivalent gate here is the Step 0.5.4 shipment claim — the label is route-specific and is recorded as such rather than copied (review finding F16). Absence / ambiguity / silence ⇒ **HALT, zero mutation**. |
| **H-A3** — Edit `.github/agents/_stage.agent.md` (self-modifying contract surface, R6) | Medium | No separate approval; constrained by U1's scope fence and harness scenario **S3**. |
| **H-A4** — **Cascade closure** of `[025-F, 025.001-T]` via the Cascade Close Sub-Procedure (`backlogit_ship_shipment`) | **DESTRUCTIVE / irreversible** *(reclassified in rev 2 — review findings F2/F7; rev 1 wrongly prescribed safe-close at `ActionRisk: Medium` with no approval)* | **YES — recorded operator approval**, matching the `017-S` precedent for the identical operation. Gated on **C3** returning `CASCADE` / `FULLY_COVERED_ROOT` and bounded by **C4**'s two-set gate. |

### Deepened runtime verification

**Environment prechecks (before U1 begins)**

1. **P-001** *(new in rev 2 — review finding F4)*: no top-level work item other than `025-F` is
   `active`. Specifically, `021-F` and `021.001-T` must be terminal. **If not, HALT before the
   claim** — P-001's own gate at Ship Step 1 fires only *after* the one-way-door claim, so it
   cannot be relied on as the guard.
2. `025.001-T` carries `harness-ready` from a real `harness-architect` run (**H-A1**). If not,
   **do not claim** (R1).
3. `git symbolic-ref --short HEAD` resolves to the Ship feature/chore branch — **not** `main`,
   **not** a Stage artifact branch.
4. **Green baseline** *(new in rev 2 — review finding F10)*: `go vet ./...` **and**
   `go test ./...` both exit 0 **before** the harness is added. Without a proven-green baseline a
   later red cannot be attributed to the new harness. *Ordering note (review finding F13): this
   baseline is taken **before** H-A1's harness files are written; "clean tree" in rev 1 precheck 2
   contradicted rev 1 precheck 4 and is replaced by this explicit ordering.*
5. `.github/agents/_stage.agent.md` contains `### Step 6: Summary` and
   `#### Pre-Summary Verification Gate (NON-NEGOTIABLE)`. If either anchor is missing, **HALT** —
   the file has drifted and U1's positioning rule is unsatisfiable.

**Target harness scenarios (3 — strictly fewer than 4; red before implementation per P-004)**

| # | Asserts | AC |
|---|---|---|
| **S1** | `#### Staging Handback Record (REQUIRED OUTPUT)` exists in `_stage.agent.md`, positioned strictly after the Pre-Summary Verification Gate heading and strictly before the summary bullet list; and all six field names — `stage_branch`, `stage_base_commit`, `stage_head_commit`, `stage_outcome`, `stage_artifact_paths`, `shipment_id` — appear in that subsection, with `shipment_id` marked conditional and `stage_outcome` constrained to `shipment`/`no-shipment` | AC-1 |
| **S2** | The subsection contains the literal derivation command including `--diff-filter=d` **and** all four pathspecs `.backlogit/ docs/plans/ docs/decisions/ docs/memory/`; **and** states all **four** `stage_artifact_paths` prohibitions — non-empty, file paths not directory prefixes, **no trailing slash**, `base..head` not tip-only | AC-2, AC-3 |
| **S3** | The subsection states the `stage_base_commit` **echo** rule with Orchestrator authority and the **authority-neutrality** sentence; **and** the Pre-Summary gate's existing `shipment_id` requirement is byte-unchanged (negative assertion, R4/I4) | AC-4, AC-5 |

**Blocked-path handling**

* Any harness scenario that cannot be made red before implementation ⇒ **HALT** and report; do not
  implement first and retrofit the assertion.
* `go test ./...` red at Step 4.3 after U1 ⇒ back to `build-feature`; **never** narrow the test
  command to make it pass (that is Option 3's amendment and is out of scope).
* Missing Step 6 anchor at precheck 5 ⇒ **HALT** to Stage for a re-plan; do not relocate the
  subsection by judgement.
* C3 returning any verdict other than `CASCADE` / `FULLY_COVERED_ROOT` ⇒ **HALT**; do not archive.

### Deepened operational closure

* **Monitoring signals**: the next Stage session emits a complete six-field handback record;
  `019.004-T` becomes specifiable against a real record shape; `backlogit doctor --check-orphans`
  reports no orphan after the adoption and after closure.
* **Rollback triggers**: a Stage session that cannot emit a complete record; an Orchestrator
  Step 1.5 that rejects a well-formed record; C3 returning a non-`CASCADE` verdict; C4's two-set
  gate failing either condition.
* **Rollback procedure — two distinct unwinds** *(rev 2, review finding F7)*:
  1. **Text rollback**: `git revert` the U1 commit. No migration, no state; the agent template is
     re-read per session, so revert is immediately effective.
  2. **Backlog-state unwind** (required if rollback occurs after closure — `git revert` does **not**
     cover this): `025-F` and `025.001-T` are already archived by the destructive cascade. Recover
     by un-archiving the **existing** records and re-adopting — **never** by creating duplicates,
     which would orphan the three inbound edges from `019-F`. If the cascade archived any artifact
     outside `required_ids`, the P-015 Violation Action applies: `git restore -- .backlogit/queue/
     .backlogit/archive/`, re-verify, and halt with `HALT — cascade detected, revert required`.
* **Owner**: Ship (execution, quality gates, PR, closure). Stage owns backlog shape only.
* **Validation window**: the next Stage session after merge.

### Human checkpoints

1. **Before claim** — P-001 terminal state for `021-F` / `021.001-T` (**O3**). Hard gate.
2. **Before claim** — H-A1 (`harness-ready` present, operator-run). Hard gate. Stage cannot satisfy it.
3. **Before claim** — H-A2 (recorded P-002 authorisation), only if H-A1's ordering cannot be
   achieved. Fail-closed on silence.
4. **Before closure** — H-A4 destructive-cascade approval, gated on C3.
5. **Before merge** — normal P-014 merge approval. Unchanged by this plan.

### Review-gate capability risk and fallback declaration

The `adversarial-review` capability pack is installed, so Ship's Step 4.4 gate dispatches
`adversarial-review` (`mode: report-only`, `reviewers: 3`) in place of the standard review skill.
**Carried-forward requirement**: plan review for this unit MUST emit literal `dispatch_mode:` and
`decision:` markers. If multi-reviewer dispatch is unavailable, the review MUST declare the
fallback explicitly (`dispatch_mode: single-reviewer-fallback`) **before** harvest rather than
silently degrading. A review that emits no `dispatch_mode:` marker is treated as **FAIL**, not as
PASS-by-default. *Attempt-1 note: dispatch was `multi-reviewer`, but one of the three reviewers
returned an empty findings array and upheld claims that primary text refutes; its signal was
discounted as non-substantive. Re-review should treat a zero-finding reviewer as non-substantive
rather than as a concurring PASS.*

### Unresolved operator decisions that still block safe execution

| # | Decision | Blocks |
|---|---|---|
| **O1** | Satisfy H-A1 (operator runs `harness-architect`, applies `harness-ready` to `025.001-T`) **before** the shipment is claimed | The claim. Without it, R1/R9's P-004 bypass is live. |
| **O2** | If O1 is not taken: grant or withhold the explicit recorded P-002 authorisation (H-A2) | The claim. Silence ⇒ HALT. |
| **O3** | Confirm `021-F` / `021.001-T` are terminal so P-001 holds at Ship Step 1 | The claim (R7). Stage completes these in the staging session; the operator confirms before Ship runs. |
| **O4** | Approve the destructive cascade closure (H-A4) | Closure only, not the claim (R8). |

O1, O2 and O4 are **outside Stage's authority**. The shipment is created `queued` and carries them
as recorded claim/closure preconditions; it is **not** claimable until O1 (or O2) and O3 are
resolved.

## Plan Review Record — GATE NOT PASSED

<!-- plan-review-attempt: 2 -->
<!-- plan-review-final: FAIL -->

This plan consumed **both permitted plan-review re-entry cycles** (Stage Step 4: max 2) and
did **not** obtain a passing or acceptable verdict. Per Stage Step 4 option (c) the FAIL is
recorded as a **P-005** event. **No shipment was created** — AC-5 is gated on AC-4.

| Attempt | Dispatch mode | Decision | Findings |
|---|---|---|---|
| 1 | `multi-reviewer` | **FAIL** | P0: F1 (restructure asserted in past tense but never performed), F2 (manifest classifies CASCADE not safe-close; H-A4/C3 prescribed an unreachable closure path), F3 (Option 1 Step 2 proof refuted by the artifact's own §5.2; R1 "already harnessed partition" mechanism factually false). P1: F4–F7. Advisory: F8–F18. |
| 2 | `multi-reviewer` | **FAIL** | 9 of attempt 1's findings verified remediated; reviewers independently re-measured and confirmed the live backlog restructure. **16 new findings, 5 at P0**: N1, N2, N3, N4, N6, N7. P1: N5, N8, N10, N11, N13, N14. |

### Attempt-2 findings and disposition

| ID | Sev | Finding | Disposition |
|---|---|---|---|
| **N1** | P0 | `.backlogit/queue/025-F.md` body still named the nonexistent pre-adoption ID `019.007-T` and manifest `[025-F, 019.007-T]` ⇒ `RECONCILE_FAIL_SNAPSHOT_MISSING` / fail-closed HALT. | **FIXED** — live backlog corruption, remediated immediately regardless of terminal path. `025-F` body rewritten to `025.001-T` / `[025-F, 025.001-T]`. |
| **N2** | P0 | Decision §3.1 "Proof against Step 3 item 1" still on the refuted `[025-F, 019.007-T]` at `queued`. | **FIXED** in decision rev 3. |
| **N15** | adv | Decision §3.2(c) wrote verdict set `CASCADE, SAFE_CLOSE, HALT`; installed token is `BLOCK` (`shipment-reconcile/SKILL.md:375-376`). | **FIXED** in decision rev 3. |
| **N16** | adv | Decision §6 AC-4 deferred to a "Plan Review section" that did not exist. | **FIXED** — this section, plus AC-4 now carries a concrete verdict. |
| **N3** | P0 | Unbound cascade: C3 only *records* `CLASSIFICATION_BINDING`; H-A4 names `backlogit_ship_shipment` directly ⇒ `RECONCILE_FAIL_CASCADE_UNBOUND`. | **NOT REMEDIATED** — fixable, but moot: no shipment exists to close. Carried to the unblock path as **B4**. |
| **N4** | P0 | C4 omits the mandatory `returned_ids == []` assertion and never defines `allowed_ids`. | **NOT REMEDIATED** — same; **B4**. |
| **N6** | P0 | **No physical execution slot for the pre-claim harness.** Every route trips a gate: uncommitted on `main` → Ship Step 0.5 3a halt; committed to `main` → red harness on the default branch outside any PR gate; operator branch → `BRANCH_MISMATCH`. | **NOT REMEDIABLE IN THIS UNIT** — installed-contract gap. **B1**, decisive. |
| **N7** | P0 | Wrong policy authorised: R2/H-A2/O2 offer a `P-002`-scoped deviation, but the consequence is a **`P-004`** red-phase bypass; the route also contradicts declared invariant I6. | **NOT REMEDIABLE BY STAGE** — re-scoping to P-004 would have Stage authorise a red-phase bypass. **B3**. |
| **N14** | P1 | **No step delivers Stage's `.backlogit/` mutations from `chore/stage-*` to `main`.** Ship Step 0.5 3a branches from `main`, so the shipment record and `queued` status are absent from Ship's branch. | **CIRCULAR — NOT REMEDIABLE IN THIS UNIT**: this is exactly the defect `025.001-T` exists to make specifiable. **B2**, decisive. |
| **N8** | P1 | `_stage.agent.md`: `#### Pre-Summary Verification Gate` is immediately followed by the summary bullets with no intervening heading, so the required `####` insertion necessarily re-parents them. The placement rule is unsatisfiable without a restoring heading. | **NOT REMEDIATED** — needs a scope change the task's fence forbids. **B5**. |
| **N13** | P1 | Stage Step 5.5 item 3 scope guard restricts `add_to_shipment` to IDs the immediately preceding harvest returned; `025.001-T` is an **adopted** pre-existing queue item. | **NOT REMEDIATED** — Stage's own contract forbids assembling the designed manifest from an adopted item. **B6**. |
| **N5** | P1 | No `harness_cmd` deliverable defined. | Deferred to re-plan. |
| **N10** | P1 | `025.001-T` carries no `<!-- BEGIN:acceptance-criteria -->` block, yet the plan traces against "Source AC"s. | Deferred to re-plan. |
| **N11** | P1 | No executable close-path classifier exists in this repo; classification is agent-driven from skill markdown, so a self-computed digest gives no drift protection. | **NOT REMEDIABLE** — property of the installed skill. **B4**. |
| N9, N12 | adv | Advisory. | Deferred to re-plan. |

### Verdict

**The plan is not approved and the unit is not safely claimable.** Blockers **B1** and **B2**
are gaps in the installed execution surface, not defects of this plan — B2 is strictly
circular. Escalation under P-013.6 (see the decision artifact §9) rather than a third attempt.

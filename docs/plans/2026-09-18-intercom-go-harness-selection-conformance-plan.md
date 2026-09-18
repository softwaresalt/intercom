---
title: "Implementation Plan — Option 3-S harness selection and enforcement conformance (026-F)"
date: 2026-09-18
status: planned
agent: Stage
source_document: docs/decisions/2026-09-18-intercom-go-shipment-scoped-harness-contract-amendment-deliberation.md
governs: feature 026-F; task 026.001-T
bound_snapshot: 101e974a2890e7ba82aacb22d734866fddaf973b
stage_branch: chore/stage-shipment-scoped-harness-contract-amendment
revision: 1
---

## Problem Frame

On the shipment-claim route, `backlogit shipment claim` auto-transitions manifest tasks
`queued → active` at `.github/agents/_ship.agent.md:265` (Step 0.5 item 4) — before Ship Step 2
runs. Step 2 item 1 (`_ship.agent.md:322`) selects *"all tasks for the target feature or chore that
are in `queued` status"*, which is now the empty set. Step 2 item 3 scaffolds nothing and item 4's
*"confirm every queued task now carries the `harness-ready` label"* passes vacuously over `∅`.

Step 3 item 1 (`_ship.agent.md:338-358`) then derives the executable set with the positive status
rule KEEP `{queued, active}`, so the unharnessed task is retained. Step 3 item 2
(`_ship.agent.md:362-367`) normally applies P-002's Enforcement filter — *"List all tasks with
`harness-ready` label and `queued` status"* — but on the shipment path it is instructed to
*"**replace** this queued-only membership with item 1's derived executable set"*. That replacement
discards **both** predicates: the `queued` status filter (intended) and the `harness-ready` label
filter (unintended). Item 1's rule never mentions `harness-ready`.

Net effect: Step 4 executes a task that was never harnessed. `.github/policies/workflow-policies.md:50`
already requires *"**Enforcement** (ship): Filter ready queue to only tasks carrying the `harness-ready`
label"* with gate point *"Queue building (Step 2) and task claiming (Step 3)"*. Ship is therefore
**non-conformant with an installed policy it already carries**. This is a conformance repair, not a
policy amendment.

A third surface participates: `.github/skills/harness-architect/SKILL.md:60` excludes *"blocked, done,
or otherwise non-ready work items"*, which is ambiguous for `active` and would re-vacuate a repaired
Step 2 through a different door.

## Requirements Trace

| Source requirement (deliberation §) | Implementation action | Unit |
|---|---|---|
| §6 S1 — Step 2 item 1 manifest-scoped, status-inclusive on the shipment path; feature-scoped `queued` retained on the non-shipment path | Edit `_ship.agent.md` Step 2 item 1 | U1 |
| §6 S2 — Step 3 item 2 substitution restores the `harness-ready` conjunct; missing label ⇒ fail-closed halt citing P-002 | Edit `_ship.agent.md` Step 3 item 2 | U1 |
| §6 S3 — harness-architect exclusion rule disambiguated for explicitly-listed `active` tasks; harness artifacts authored on the current feature/chore branch | Edit `harness-architect/SKILL.md` Step 1 item 3 (+ a branch-locus note) | U1 |
| §6 S4(a) — shipment-route selector includes an `active` manifest member | Test | U1 |
| §6 S4(b) — derived member lacking `harness-ready` halts fail-closed | Test | U1 |
| §6 S4(c) — non-shipment path unchanged | Test | U1 |
| §6 S4(d) — Step 4.3 full-suite text unmodified (anti-regression pin) | Test | U1 |
| §6 Non-goals — P-002/P-004/Step 4.3 text unamended | Test S4(d) + review assertion | U1 |
| §2.4 / F-2 — name the legal pre-claim harness slot | Covered by S3's branch-locus note | U1 |

**Deliberately unmapped (out of scope, recorded):** B2 / Orchestrator Step 1.5 → Unit 2 (`027-F`).
B4–B6 → remediable inside a re-planned `025.001-T` after Unit 2. `3F546E63`, `11B75632` → remain
stashed.

## Implementation Units

### U1 — Harness selection and enforcement conformance (single atomic unit)

**Why one unit and not three.** Cardinality `N = 1` is mandatory. A second queued test-bearing task
under the same covering feature reproduces the G1 deadlock: Step 2 would emit `N = 2` red functions,
task 1 would go green, function 2 would remain red by design, and Step 4.3's full-suite gate would
fail after every task. Splitting is therefore unsafe, not merely inconvenient. The three surfaces are
also mutually dependent (deliberation §4): S1 without S2 leaves the bypass reachable; S2 without S1
halts every shipment route; either without S3 can be re-vacuated by the skill's exclusion rule.

**Files affected (3):**

1. `.github/agents/_ship.agent.md` — Step 2 item 1; Step 3 item 2 substitution clause
2. `.github/skills/harness-architect/SKILL.md` — Step 1 item 3 (+ branch-locus note)
3. `tests/integration/harness_selection_conformance_test.go` — new

**2-hour rule assessment.** 3 files (one over the `<3` heuristic — justified below and in
deliberation R-4), 0 production functions changed, 4 test scenarios. Two of the three files are
instruction surfaces with surgical clause-level edits. Precedent: `018.008-T` shipped a 3-file
contract task (2 instruction surfaces + 1 test) successfully in `017-S`. Estimated 1.5–2.0h.

**Width isolation.** Single domain: **contract text + its verifying tests**. No production Go source
is modified. This matches the established workspace pattern for instruction-surface tasks
(`022.001-T`, `018.008-T`).

**Execution posture: test-first (RED mandatory).** Write all four scenarios first and observe them
fail against the unmodified surfaces, then edit the clauses to green. A green-on-arrival harness is
itself a failure (compound residual N-5/N-6).

**Changes required:**

* **S1 (`_ship.agent.md` Step 2 item 1).** Replace the single-sentence selector with a two-branch
  rule: when operating under a Stage-prepared shipment (Step 0.5 recorded a `shipment_id`), the
  batch is the Step 3 item 1 manifest-derived executable set, admitting both `queued` and `active`;
  otherwise the existing feature-scoped `queued`-only behaviour applies **verbatim**. Item 4's
  confirmation must be restated over *the selected batch* rather than over *"every queued task"*, so
  it cannot pass vacuously. Cross-reference Step 3 item 1 as the single authoritative derivation —
  do not restate the status rule (avoid a second divergent copy).
* **S2 (`_ship.agent.md` Step 3 item 2).** In the substitution clause, after *"include every task in
  that set regardless of whether its status is `queued` or `active`"*, add the conjunct: the set is
  further filtered to members carrying the `harness-ready` label; any derived member lacking it is a
  **fail-closed halt** — never a silent skip and never a proceed — citing P-002's Violation Action
  (*"Halt and suggest running the harness-architect"*). State explicitly that this restores
  `workflow-policies.md:50`'s Enforcement clause on the shipment route.
* **S3 (`harness-architect/SKILL.md` Step 1 item 3).** Amend to: exclude `blocked`, `done`, and
  archived items; when `${input:tasks}` is supplied explicitly, an `active` task **is in scope** and
  must not be excluded as "non-ready". Add a short locus note: harness artifacts are authored on the
  invoking session's current feature/chore branch, never on `main` (deliberation F-2).

**Tests (4 scenarios, `tests/integration/harness_selection_conformance_test.go`):**

| # | Name | Asserts |
|---|---|---|
| T1 | `TestStep2_ShipmentRoute_IncludesActiveManifestTask` | The Step 2 selector clause admits an `active` manifest member on the shipment path (no longer `queued`-only) |
| T2 | `TestStep3Item2_DerivedMemberWithoutHarnessReady_HaltsFailClosed` | The substitution clause carries the `harness-ready` conjunct and a fail-closed halt, citing P-002 |
| T3 | `TestStep2_NonShipmentRoute_FeatureScopedQueuedUnchanged` | The direct-invocation path retains feature-scoped `queued`-only selection verbatim |
| T4 | `TestStep43_FullSuiteGate_Unmodified` | Step 4.3's `go vet ./...` / `go test ./...` full-suite text is byte-unchanged (anti-regression pin proving the scope fence held) |

Needles MUST be derived from the **post-edit pinned wording** and each assertion must resolve inside
its own subject; no assertion may be credited by another's output (compound residuals N-5/N-6).

**Atomic milestone.** `go vet ./...` exits 0 and `go test ./...` is fully green with all four
scenarios passing, having first been observed red.

## Dependency Graph

```text
U1 (026.001-T)  — no unfinished upstream dependencies
   │
   └── blocks ──▶ 027-F / Unit 2 (B2, Orchestrator Step 1.5)   [blocked, not in this shipment]
                     │
                     └── blocks ──▶ 025-F / 025.001-T           [remains blocked]
```

No cycles. U1 is the single executable node this cycle. `026.001-T → DB12DA37` is a traceability
link, not a blocking edge (the stash entry is consumed, not a prerequisite).

## Decisions and Rationale

| # | Decision | Rationale |
|---|---|---|
| D1 | Repair Ship's conformance rather than amend P-002/P-004 | `workflow-policies.md:50` already mandates the `harness-ready` filter. The defect is that Step 3 item 2's substitution drops it. Amending policy to describe broken behaviour inverts the control direction (precedent: `stage-policy-gap-re-plan-decision` §2). |
| D2 | Do **not** narrow Step 4.3 | It is the workspace's broadest regression detector; narrowing it to the manifest silently stops detecting cross-task regressions (Option 1 decision §3.3 item 1). Scoping Step 2 alone is sufficient. |
| D3 | Manifest scope can only shrink the batch | For a Stage-assembled shipment the manifest task set is a subset of the covering feature's tasks, so `N` is `≤` today's. G1 is never worsened; under Option 1 it stays `N = 1`. |
| D4 | Single task, three files | `N = 1` is mandatory (D3); a second task reintroduces G1. Three surfaces are mutually dependent. Precedent `018.008-T`. |
| D5 | Fail-closed halt, not skip, on missing `harness-ready` | P-002's Violation Action is literally *"Halt and suggest running the harness-architect."* A skip would silently shrink the executable set and could empty it, colliding with Step 3 item 1's empty-set HALT. |
| D6 | Cross-reference Step 3 item 1 rather than restate the status rule in Step 2 | A second copy of the positive status rule would be a new divergence seam — precisely the class of defect being repaired. |
| D7 | Bootstrap via the `017-S` pre-claim precedent | U1 must travel the unrepaired route once. A real pre-claim harness on the shipment branch makes Step 2's vacuity harmless rather than authorized. |

## Risks and Caveats

| # | Risk | Mitigation |
|---|---|---|
| R-1 | **Bootstrap depends on a manual operator step.** If the shipment is claimed without the pre-claim harness, Ship executes an unharnessed task — the exact defect being fixed. | Hard pre-claim precondition recorded on the shipment record; Stage halts there and does not authorize deviation. Self-eliminating after U1 merges. |
| R-2 | **CASCADE closure is destructive and irreversible.** Manifest `[026-F, 026.001-T]` classifies `CASCADE`/`FULLY_COVERED_ROOT`, invoking `backlogit_ship_shipment`. | Classify `ActionRisk: destructive`; require recorded operator approval; document a backlog-state unwind distinct from `git revert`. |
| R-3 | **Cross-artifact closure risk (7 prior FAILs).** A producer-scoped fix that leaves a consumer surface open. | Build the surface matrix (producer, every consumer, authority grant, refusal path, tests, probes) against the bound snapshot `101e974` **before** editing. F-4's third surface is already caught. Probe the tool, not the prose. |
| R-4 | **Editing `_ship.agent.md` could disturb adjacent clauses**, as B5 showed for `_stage.agent.md`. | Edits are additive-conjunct and two-branch restatements at clause level; T4 pins Step 4.3 as an anti-regression witness; verify heading/list structure after each edit. |
| R-5 | **Substring-only tests cannot observe behaviour** (022.001-T / IV-5 precedent). | T1–T3 assert clause structure AND the presence of the decisive conjuncts/branches; where feasible, drive a fixture manifest through the derivation logic rather than asserting prose alone. Needles taken from post-edit pinned wording, never from soft-wrapped source. |
| R-6 | **Repaired Step 2 could surface a latent G1** for any future multi-task manifest. | Out of scope and unchanged by this plan (manifest ⊆ feature, D3). Recorded as the standing reason the Option 1 single-task shape remains in force. |

## Plan Hardening Signals (REQUIRED)

| Signal | Present | Justification |
|---|---|---|
| Public API, schema, or **contract** change | **YES** | Three installed contract surfaces change: Ship Step 2 item 1, Ship Step 3 item 2, harness-architect Step 1 item 3. |
| Security, auth, permission, or compliance-sensitive behavior | **YES** | The change governs a **TDD safety gate** (P-002 Enforcement / P-004 red phase). A mis-scoped edit re-opens a silent red-phase bypass. Also touches P-010 role-boundary language. |
| Migration, backfill, **destructive** or **irreversible** step | **YES** | Two one-way doors: the shipment claim (no `unclaim`/`release` exists — measured) and `CASCADE` closure via `backlogit_ship_shipment`. |
| External integration, **operator checkpoint**, or external dependency | **YES** | Mandatory operator checkpoints: the staging PR merge, the pre-claim `harness-architect` run, and destructive-closure approval. |
| High runtime, rollout, or rollback risk | **YES** | Rollback of a merged contract change requires both `git revert` and a distinct backlog-state unwind; the bootstrap route is single-shot. |

**Requires plan hardening: yes**

## Runtime Verification and Closure

**Runtime surface changed:** No CLI, API, browser UI, or background-job surface changes. The changed
surfaces are **agent-contract instruction files**, whose "runtime" is agent execution. Full-build
non-applicability does **not** apply, because a new Go test file is added — `go build ./...`,
`go vet ./...`, and `go test ./...` must all be run and recorded.

**Runtime verification that should prove absorption:**

1. `go vet ./...` exits 0; `go test ./...` fully green with T1–T4 passing, each observed red first.
2. T4 proves the scope fence: Step 4.3's full-suite text is byte-unchanged.
3. A read-back of the edited clauses confirms: Step 2 item 1 has both branches; Step 3 item 2 carries
   the `harness-ready` conjunct plus the fail-closed halt; harness-architect item 3 admits an
   explicitly-listed `active` task.
4. Structural integrity: heading levels and list numbering in both instruction files unchanged
   outside the edited clauses.

**Operational closure artifact:** `docs/closure/026-S-026-F-post-merge-closure.md` recording —
*monitoring*: the next shipment-claim route run must show Step 2 emitting a **non-empty** batch (the
direct observable proving the repair); *rollback trigger*: any Step 2 batch that is empty while the
manifest has an executable member, or any Step 4 execution of a task lacking `harness-ready`;
*ownership*: Ship for execution, operator for the destructive-closure approval; *validation window*:
the first subsequent shipment-claim route through Ship; *follow-ups*: Unit 2 (`027-F`) for B2,
then re-planned `025.001-T` for B4–B6.

## Plan Hardening

**Hardening required: YES.** All five impl-plan signals fired. The decisive ones are
*contract change* (three installed agent/skill surfaces), *safety-sensitive behavior* (the change
governs the P-002/P-004 TDD gate itself), and *irreversible step* (two one-way doors: the shipment
claim, for which no `unclaim`/`release` exists — measured against `backlogit` 1.10.1 — and `CASCADE`
closure).

### Learnings and instruction files consulted

| Source | Carried into this plan |
|---|---|
| `docs/compound/workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md` | Seven consecutive plan-review FAILs on producer-scoped fixes that left a consumer open. Mandates the **six-surface matrix before any edit** (H-1) and the **stop rule** (H-7). |
| `docs/compound/2026-05-07-backlogit-shipment-status-constraints.md` | No shipment `blocked` status; closure only via `shipment ship` (cascade) or safe-close; archive location and declared status are independent facts. |
| `docs/compound/2026-09-08-adversarial-review-empirical-verification-and-scope-check.md` | Reproduce installed-behaviour claims before asserting them — applied in H-1's probe column. |
| `docs/plans/2026-09-16-017-S-closure-canonical-contract.md` (D-D6, D-D7, PF-12, S-3) | `harness-ready` may be applied **only** by the installed harness-architect skill; producer admission must be pre-claim; the claim is a one-way door. |
| `docs/decisions/2026-09-18-intercom-go-harness-execution-model-decision.md` §3.3, §5.2, §9.2 | Option 3's standing cautions; the measured bypass chain; blockers B1–B7. |
| `.github/instructions/strict-safety.instructions.md` | `ProposedAction` / `ActionRisk` / `ActionResult` classification below. |
| `.github/instructions/concurrency.instructions.md` | Single-agent, single-branch, single-worktree — no file locks required this session. |

### H-1 — MANDATORY surface matrix (build BEFORE any edit; blocking)

Ship MUST materialize this table as the first build artifact and confirm every edge is closed
against the single bound snapshot `101e974`. An open edge is a stop condition, not a note.

| # | Surface role | File:line (bound snapshot) | Current evidence | Pinned replacement | Executable verification |
|---|---|---|---|---|---|
| 1 | Producer — harness batch selector | `_ship.agent.md:322` | `queued`-only, feature-scoped | two-branch rule (S1) | T1, T3 |
| 2 | Producer — vacuity guard | `_ship.agent.md:329` | *"every queued task"* | restated over the selected batch | T1 |
| 3 | Consumer — executable set | `_ship.agent.md:338-358` | KEEP `{queued, active}` | **unchanged** (authoritative) | T1 (cross-ref only) |
| 4 | Consumer — enforcement | `_ship.agent.md:362-367` | substitution drops `harness-ready` | conjunct + fail-closed halt (S2) | T2 |
| 5 | Consumer — skill selection | `harness-architect/SKILL.md:60` | *"otherwise non-ready"* ambiguous | explicit `active`-in-scope (S3) | T1 |
| 6 | Authority grant | `workflow-policies.md:50` P-002 Enforcement | already mandates the filter | **unchanged — no amendment** | T2 asserts the citation |
| 7 | Refusal / fallback path | P-002 Violation Action, `workflow-policies.md:52` | *"Halt and suggest running the harness-architect"* | cited verbatim in S2 | T2 |
| 8 | Anti-regression witness | `_ship.agent.md:435-447` Step 4.3 | full-suite `go test ./...` | **byte-unchanged** | T4 |

**Stop rule (H-7).** If a finding reappears at an adjacent surface during build or review, treat it
as a design defect: stop, rebuild the matrix, do not patch locally.

### H-2 — Protected invariants (each must be provable at closure)

| ID | Invariant | Proof |
|---|---|---|
| **I1** | Step 4.3's full-suite gate text is byte-unchanged | T4 |
| **I2** | P-002 and P-004 policy text are byte-unchanged | Diff review: `workflow-policies.md` must not appear in the PR diff |
| **I3** | The non-shipment/direct-invocation path is behaviourally unchanged | T3 |
| **I4** | No derived member without `harness-ready` can reach Step 4 | T2 |
| **I5** | `N = 1` holds — exactly one harnessed-but-unimplemented function exists at any time | Manifest is `[026-F, 026.001-T]`; one task |
| **I6** | No rev-11…rev-15 mechanism is reintroduced in explicit or disguised form | Review assertion: zero new manifest/witness/selector-flag/predicate constructs |
| **I7** | The `023-F`/`024-F` task-only trigger is not fired | Manifest is a fully-covered root ⇒ `CASCADE`, never task-only |
| **I8** | Structural integrity of both instruction files outside the edited clauses | Heading/list-level read-back (R-4) |

**I1, I2 and I6 govern over any conflicting statement elsewhere in this plan.** They are the scope
fence; a green suite that violates any of them is a false green.

### H-3 — Risky actions (`ProposedAction` / `ActionRisk`)

| ID | ProposedAction | ActionRisk | Approval | Expected ActionResult |
|---|---|---|---|---|
| **PA-1** | Operator pushes `chore/stage-*` and merges the merge-commit staging PR | `moderate` — reversible by revert | Operator (P-009 merge-commit only) | Backlog + shipment artifacts present on `origin/main` |
| **PA-2** | Operator creates `feat/{shipment-slug}` from fresh `main` and invokes **harness-architect** for `026.001-T` | `moderate` — red harness committed on the shipment branch, inside the PR gate | Operator (P-004 producer duty; P-010 forbids Stage) | `026.001-T` carries `harness-ready`; red phase CONFIRMED |
| **PA-3** | Ship claims shipment `026-S` | **`irreversible`** — one-way door; no `unclaim`/`release` exists | Operator, and **only after PA-2 is verified** | Shipment `active`, sole active shipment; task auto-activated |
| **PA-4** | Ship edits three installed contract surfaces | `elevated` — contract blast radius | Covered by PR review + P-014 gate | I1–I8 all provable |
| **PA-5** | `CASCADE` closure via `backlogit_ship_shipment` | **`destructive`** — irreversible archive of the manifest subtree | **Explicit recorded operator approval, mandatory** | `archived_ids == required_ids`; `returned_ids == []` |

**PA-3 and PA-5 must not be executed on an unverified precondition.** Absence, ambiguity, or silence
on approval ⇒ **HALT, zero mutation**.

### H-4 — Pre-claim environment prechecks (all must pass before PA-3)

1. `026.001-T` carries `harness-ready` **and** a real red harness exists on the branch (PA-2 done).
2. No other top-level release unit is `active` (**P-001**). `021-F`/`021.001-T` are `done`; `027-F`
   is created `blocked` precisely so it is not a competing active unit.
3. Current branch is `feat/{shipment-slug}` ⇒ Step 0.5 3a logs `BRANCH_OK` (never `BRANCH_MISMATCH`).
4. `git worktree list --porcelain` shows exactly one worktree (**P-016**).
5. `git show origin/main:.backlogit/queue/026-S.md` resolves (**PA-1 verified on the remote**).
6. Shipment status is `queued`; `026.001-T` status is `queued` — no `SHIPMENT_STATE_INCONSISTENT`.

**Blocked-path handling:** if any precheck fails, do **not** claim. Report the failing precheck and
halt. A failed precheck is never remediated by proceeding.

### H-5 — Rollback coupling (two independent state domains)

A merged contract change cannot be rolled back by `git revert` alone — the backlog state is a
separate domain.

| Phase reached | Git rollback | Backlog-state unwind |
|---|---|---|
| Before PA-3 (claim) | Delete the feature branch | None — nothing mutated |
| After PA-3, before merge | Reset/abandon the branch | **No `unclaim` exists.** Shipment is stranded `active` in the P-001 slot. Exit is operator-directed only (`return-blocked` is single-item; `abandoned` is destructive and requires approval). **This is the point of no cheap return — the reason PA-3 is gated on H-4.** |
| After merge, before PA-5 | `git revert` the merge commit (merge-commit-only, P-009) | Shipment remains `active`; re-plan required |
| After PA-5 (cascade) | `git revert` the merge commit | **Irreversible archive.** Unwind = restore `026-F`/`026.001-T` records from `.backlogit/archive/` to `.backlogit/queue/` and reset the shipment record — a manual, operator-approved backlog surgery, explicitly **not** `git revert`. |

### H-6 — Monitoring signals and validation window

* **Primary observable (repair proof):** on the next shipment-claim route, Step 2 emits a
  **non-empty** batch. An empty batch with an executable manifest member means the repair did not
  take — immediate rollback trigger.
* **Secondary:** no Step 4 execution of a task lacking `harness-ready` (S2 must halt first).
* **Anti-regression:** T4 stays green in CI; `workflow-policies.md` never appears in this PR's diff.
* **Validation window:** the first subsequent shipment-claim route through Ship after merge.
* **Owner:** Ship for execution and monitoring; operator for PA-1, PA-2, PA-3 and PA-5 approvals.

### H-7 — Review-gate capability risks (carried into plan-review)

* Plan review MUST emit literal `dispatch_mode:` and `decision:` markers. If multi-persona dispatch
  is unavailable, review MUST declare the fallback explicitly (**P-012**: no silent degradation to
  ad hoc reading) and record it as a degraded-review condition before harvest.
* **Known degraded capability in this workspace:** `4029DABB` records that `shipment-reconcile`
  cites a classifier path that does not exist here, so closure classification is agent-computed
  rather than tool-executed. Review must treat the `CASCADE` classification as **asserted, not
  tool-verified**, and require PA-5's two-set post-condition (`allowed_ids` / `required_ids`, plus
  `returned_ids == []`) to be evaluated as separately-labelled, independently-failing conditions.
* This session runs in **CLI-fallback mode** (no MCP tool surface exposed); all backlog mutations use
  the `backlogit` CLI per the registry's `cli_command` fallbacks. Recorded per P-012 — a declared
  degraded mode, not a silent one.

### H-8 — Unresolved operator decisions that still block safe execution

1. **PA-2 executor confirmation.** The pre-claim `harness-architect` run is an operator action;
   P-010 forbids Stage and Ship's only invocation point (Step 2) is post-claim and currently vacuous.
   Ship MUST NOT claim until the operator confirms PA-2 is complete.
2. **PA-5 destructive-closure approval** must be recorded before cascade closure.
3. **Unit 2 (`027-F`) scope** — whether it absorbs `3F546E63` (claim-last expressibility) is deferred
   to Unit 2's own deliberation and is explicitly not decided here.

None of these block **planning or harvest**; all three block **execution** and are recorded on the
shipment record as hard gates.

## Plan Review — attempt 1

```text
dispatch_mode: multi-agent
decision: FAIL
```

Personas dispatched: Constitution Reviewer, Scope Boundary Auditor, Learnings Researcher
(always-on) and Architecture Strategist (cross-model anchor, `gpt-5.6-sol`). Go Reviewer was
**not** dispatched: the unit modifies no production Go source, and its rubric (type signatures,
error handling, package boundaries) has no subject here. Agent-Native Parity and Security Lens were
not triggered. Coverage of every *selected* persona is complete, so the multi-agent dispatch mode
stands.

| Persona | Verdict | P0 | P1 |
|---|---|---|---|
| Constitution Reviewer | **FAIL** | 1 | 2 |
| Scope Boundary Auditor | ADVISORY | 0 | 3 |
| Learnings Researcher | **FAIL** | 1 | 3 |
| Architecture Strategist (anchor) | **FAIL** | 0 | 6 |
| **Aggregate** | **FAIL** | **2** | **14** |

### What the review UPHELD (re-measured independently, not merely accepted)

* **The root-cause localization is correct.** `workflow-policies.md:50` P-002 Enforcement exists as
  quoted; `_ship.agent.md:362-367`'s substitution discards it; Step 3 item 1 (`:339-361`) contains
  **zero** occurrences of `harness-ready`. Conformance-repair framing verified true.
* **B1 is genuinely refuted.** `_ship.agent.md:189` fourth arm confirmed; prior art did **not**
  consider-and-reject it. B1 downgrades to a specification gap.
* **P-004 is not bypassed or re-scoped**; P-001, P-009, P-010, P-015 preserved; the
  `023-F`/`024-F` `TASK_ONLY_FINALIZE` trigger is provably unfired; **I6 holds** — no
  rev-11…rev-15 mechanism is present, including via the `harness_status` custom field that would
  have been the natural disguise vector.
* Keeping Step 4.3 un-narrowed is confirmed correct by prior art.

### Blocking findings (C-series) — why they are not locally patchable

| ID | Finding | Source | Why it blocks |
|---|---|---|---|
| **C1** | **S1 is temporally incoherent.** Step 2 cannot consume "Step 3 item 1's derived executable set": Step 3 runs *after* Step 2, and item 1 is explicitly scoped *"Before assembling the queue in item 2 below"*. The cross-reference points at a not-yet-computed value. | Architecture P1 | The only sound fixes are (a) move the authoritative derivation into Step 2 and restate Step 3 item 1 as its consumer — restructuring a clause carrying dense C1–C6 / 139-F provenance, far beyond Option 3-S's declared scope; or (b) give Step 2 its own manifest read — creating the second divergent copy of the status rule that decision **D6** forbids, i.e. re-creating the exact defect class being repaired. |
| **C2** | **Two conflicting declarations of the final execution boundary.** Step 3 item 1 declares its status-only derived set *"the actual task-membership boundary that Step 4 executes"* (`:356-358`). Adding the `harness-ready` filter only to item 2 leaves two surfaces each claiming final authority. | Architecture P1 | Resolving it requires editing item 1 as well — again outside the declared three-surface scope, and compounding C1. |
| **C3** | **The acceptance surface cannot satisfy P-004.** T4 asserts Step 4.3 is *unchanged*, so it is **green-on-arrival by construction**; P-004 requires `go test ./...` to fail *"for every test function"* before `harness-ready` may be applied. harness-architect must therefore **refuse the label**, which fails H-4 precheck 1 and makes the bootstrap **self-blocking at PA-3**. T3 has the same defect. | Constitution **P0**, Scope P1 | Dropping T4/T3 fixes the red-phase problem but collides with C4. |
| **C4** | **No executable engine exists to produce a behavioural harness.** The adopted `022.001-T` / IV-5 resolution requires fixture-backed behavioural tests and demotes substring needles to *supplemental*. Ship Steps 2–3 are **agent-interpreted prose**, not code — unlike `shipment-reconcile`, which the IV-5 harness could drive through the real `backlogit` binary. There is nothing to drive. | Learnings **P0**/P1 | Either the acceptance surface is substring-only (contradicting an adopted decision), or a contract-linter engine must be built first — a substantial separate unit. Declaring IV-5 non-applicable is an **operator** determination Stage cannot self-grant. |
| **C5** | **The consumer surface inventory is materially incomplete** — the seven-FAIL defect class reproduced inside the plan that cites it. Missing: `harness-architect/SKILL.md:50` (*"all **ready** tasks"*) and `:56` (*"its **ready** descendants"*) — two further `ready`-scoped doors that re-vacuate a repaired Step 2; `_ship.agent.md:318` and `:336` preambles (*"every task in the target feature"*, *"all tasks are harnessed"*); Step 4.1 `:372-379`; `build-feature/SKILL.md:8`. | Learnings **P0**, Architecture P1 | Real surface count is ~8–9 clause sites across ≥4 files, not 3. **Both** the Learnings and Architecture personas independently invoked the compound learning's **stop rule** (a finding reappearing at an adjacent surface is a design defect ⇒ rebuild the matrix, do not patch locally). |
| **C6** | **D3's cardinality proof is invalid.** `manifest ∩ {queued, active}` is *not* a subset of `feature ∩ {queued}` — the new selector deliberately adds `active` members, so the post-claim batch goes from empty to non-empty. The single-covering-feature premise is also unenforced: intake requires only that each task have *some* parent (`_ship.agent.md:166-167`). | Architecture P1 | The G1-safety argument must be rebuilt from scratch, not reworded. |
| **C7** | **No route discriminator.** "Stage-prepared shipment" is equated with *"Step 0.5 recorded a `shipment_id`"*, but the direct-invocation fallback (`:274-299`) also creates, claims and records one — so the "unchanged direct path" promise is not met. | Architecture P1 | Requires an origin discriminator plus a third tested route. |
| **C8** | **H-1 assigns plan-artifact authorship to Ship** (*"Ship MUST materialize this table as the first build artifact"*). This re-runs recorded finding **A-2**: Ship's P-010 boundary forbids creating or modifying plan artifacts, and it previously deadlocked a gate. | Learnings P1 | A P-010 defect authored by Stage into its own plan. Fixable, but it invalidates the matrix's delivery mechanism. |

### Non-blocking findings carried forward (not remediated this session)

Constitution: missing `## Constitution Check` section (P1); P-002 **Statement** row
(`workflow-policies.md:44`, *"may only **claim** and implement"*) never adjudicated against
post-claim in-session harnessing (P1); three drifted line citations (P2); `archived_ids ==
required_ids` propagates a post-condition withdrawn by P-015 Amendment 1.21.0 (P2); the empty
`**Format**: ` ` `` gate at `_ship.agent.md:440` needs a P-021 C2 capture (P2); Orchestrator
Step 1.5 item 3(e)'s direct-push arm not explicitly fenced for PA-1 (P2); width-isolation and
file-count deviations under-disclosed (P2/P3). Scope: I1 unprovable by a needle test (P1); R-5's
undecided fixture escape hatch (P1); I5 unit drift, function-count vs task-count (P2); `027-F`
placeholder created for undecided scope (P3). Architecture: U1→Unit 2 modelled as a hard `blocks`
edge when the source decision calls it a sequencing preference (P2).

### Gate disposition

**FAIL, attempt 1.** Stage Step 4 offers (a) re-plan, (b) operator-supplied revised plan, or
(c) halt and record P-005. **Option (c) is taken**, for a reason recorded here rather than
deferred to a second attempt:

C1+C2+C5+C6 are not wording defects. Closing them expands the unit from 3 clause sites to ~8–9
across ≥4 files and requires restructuring Step 3 item 1's densely-provenanced C1–C6 clause. That
breaches the 2-hour rule decisively, and the unit **cannot be split** — a second queued
test-bearing task under one covering feature reproduces the G1 deadlock (decision **D4**). This is
the identical structural trap that terminated the Option 1 decision.

C3+C4 are worse: they show the unit has **no admissible acceptance surface**. An "unchanged"
assertion can never be red, so it cannot satisfy P-004; and a behavioural harness has no engine to
drive because the subject is agent-interpreted prose. Resolving this needs an operator
determination on IV-5 applicability, which **Stage may not self-grant**.

Against the operator's own admission gate — adopt a shipment-scoped harness contract *only if*
execution slot, branch/PR ownership, harness-ready timing and claim transition are **all** fully
executable — the measured result is **2 of 4 fail**: harness-ready timing is unadjudicated against
P-002's Statement clause (Constitution P1), and the claim transition is **not repairable at all**
(`backlogit` 1.10.1 exposes no `unclaim`/`release` and no flag to suppress descendant
auto-activation). Adoption is therefore not permitted this cycle.

**P-005 recorded.** No second attempt is opened; a re-plan would land in the same place while
spending a cycle. No shipment is created.


<!-- plan-review-attempt: 1 -->

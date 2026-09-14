---
title: "Decided Plan — Ship covering-feature completion and close-path delegation (022-F)"
date: 2026-09-13
status: planned
agent: Stage
revision: 10
feature: 022-F
task: 022.001-T
shipment: 021-S
releases: 017-S
deliberation: docs/decisions/2026-09-13-intercom-go-a-only-simplification-decision.md
archived_full_plan: docs/archive/plans/2026-09-12-intercom-go-ship-feature-completion-foundation-plan.md
evidence: docs/plans/evidence/2026-09-12-task-only-shipment-finalization/probe25-root-included-cascade-fixture.txt
---

# Decided Plan — Ship covering-feature completion and close-path delegation (022-F)

> **This is the SOLE GOVERNING ARTIFACT for `021-S`.** It states the current, executable
> contract and nothing else. It carries **no inline review narrative**: lineage, findings and
> superseded wording are reachable through the audit pointers in §13.2, which resolve to Git
> history, PR #54, the archived plan, `docs/memory/` and `.backlogit/checkpoints/`.
>
> **Reauthored, not patched.** Revision 10 replaces revision 9 wholesale under explicit
> operator direction to reauthor rather than append. Lineage is preserved by Git commits and
> PR history; it is deliberately not restated inline.

## 1. Objective

Grant Ship one narrow authority — complete a covering feature `active -> done` — and replace
Ship's **duplicated copy** of the close-path classification with **delegation** to the
installed policy and skill.

**Why it is required**: pre-mode runs `expected_status: done` over **every** manifest item, so
a fully-covered-root manifest cannot pass while its covering feature is still `active`. Ship's
Role Boundary today permits only task transitions. Independently, Ship's duplicated copy has
**already drifted unsafe** — it says *"every one of its children"* where the installed policy
requires descendants **at every depth**.

## 2. Surface — exactly 3 implementation files

| File | Change |
|---|---|
| `.github/agents/_ship.agent.md` | 9 clause sites (**CS1–CS9**, all mandatory) + 2 additive (**ADD-1**, **ADD-2**) |
| `.github/policies/workflow-policies.md` | 1 clause site (**CS10**) — the P-010 `Ship MAY` grant |
| `tests/integration/ship_feature_completion_contract_test.go` | **new** — 20-row table-driven contract test |

**0 new production files.** Stage-authored planning artifacts (this plan, the Probe-25
evidence, `docs/memory/`, `docs/decisions/`, `.backlogit/` records) are committed separately
and are **not** counted against this 3-file surface.

**Why the third file is in scope.** `_ship.agent.md`'s Role Boundary table and P-010's
`Ship MAY` list are **two halves of one authority contract**; P-010 is the policy registry that
the agent table is required to conform to. Granting the covering-feature transition in only one
half lands Ship simultaneously **authorized** by its agent table and **unauthorized** by the
policy registry. The expansion is **surgical**: exactly one bullet in P-010's `Ship MAY` list
(**CS10**). No other policy, no other clause, no other file.

> **Label note (traceability):** clause sites are **CS1–CS10** here. The archived plan called
> them `C1`–`C8`; `CSn` ≡ `Cn` for n ≤ 8. Renamed because the Probe-25 criterion tokens are
> also named `C1…C8` and the collision was a live mis-read hazard.

## 3. Authority model (normative)

### 3.1 Lifecycle is declared status; location is integrity only

* **Declared `status`**, read from the record's own frontmatter, **determines lifecycle
  class**. Nothing infers lifecycle from which directory holds a record.
* **Location has exactly two jobs, neither of which is a lifecycle decision:**
  1. **Containment (general, every member, every status)** — present in `.backlogit/queue/`
     **XOR** `.backlogit/archive/`. Both ⇒ torn duplicate. Neither ⇒ missing.
  2. **Archive routing integrity (specific)** — applied **only after** a declared
     `status: archived` has already been established, requiring that the single copy be the
     `archive/` one.
* Consequence: a descendant declaring `status: done` is **completed**, even when registry
  routing stores it under `archive/`. It is **not** pre-archived and is never tested against
  the §3.3 allowlist.

### 3.2 The grant

Ship may transition **one** covering feature `active -> done`, at **one** position
(Step 6.1(a1)), scoped to the **active shipment's manifest**, via
**`backlogit_move_item`** (MCP) / **`backlogit move {id} --status done`** (CLI fallback).
It may **not** archive, reparent, create or delete. All five conditions are conjunctive:

1. every manifest descendant is **complete** — declares `status: done`, or is exempt under §3.3;
2. **no live descendant remains** (`queued` or `active`);
3. **topology intact** — every descendant's `parent_id` resolves to this feature;
4. **containment holds** for every member (§3.1);
5. the feature's **live status is exactly `active`**.

**Any condition unmet ⇒ HALT, fail closed, no mutation.**

#### 3.2.1 Descendant scope — the full live graph, not the manifest (normative)

Conditions 1–3 quantify over the **full live `parent_id` descendant graph of the feature at
every depth**, enumerated by walking `parent_id` from the feature across
`.backlogit/queue/` ∪ `.backlogit/archive/` — **never** over the manifest list alone.

**Additionally, condition 3 requires set equality:** the enumerated descendant graph ∪ `{the
feature}` **MUST equal** the manifest exactly. A descendant found in the live graph but absent
from the manifest ⇒ **HALT**. A manifest member absent from the graph ⇒ **HALT**.

This is the same quantifier P-015's **VERIFIED FULLY-COVERED-ROOT EXCEPTION** clause 1 uses,
and the same one the installed `shipment-reconcile` skill applies at `mode: safe-close`
Step 0(c). A manifest-only reading is the exact defect P-015 clause 1 names: a manifest
`[feature, task]` whose task has an out-of-manifest subtask would pass a manifest-only check
and be destroyed by the cascade before any postcondition gate observes it.

**Zero-descendant case.** A feature enumerating to zero descendants is accepted **only** when
that childlessness is **positively verified** against the live workspace. An enumeration that
errored, was incomplete, or encountered a malformed record is **not** verified childless ⇒
**HALT**.

### 3.3 Pre-archived exemption (narrow, fail-closed)

A descendant is exempt from condition 1 **only when all five** hold:

1. it **declares `status: archived`**;
2. **containment** — exactly one physical copy;
3. **archive routing integrity** — that copy is the `archive/` one;
4. **valid archive provenance** — well-formed `archived_from`, present `archived_status`,
   resolvable `artifact_type`;
5. **it is on the exact allowlist below** — a recorded Stage disposition naming it.

**Condition 5 is load-bearing.** 1–4 prove the record is stably archived; only a Stage
decision proves it is **not owed**. `archived_status: queued` never proves completion.

**`017-S` allowlist — exactly these 11 descendants of `018-F`, and no others:**

`018.001-T`, `018.001.001-ST`, `018.001.002-ST`, `018.001.003-ST`,
`018.002-T`, `018.002.001-ST`, `018.002.002-ST`, `018.002.003-ST`,
`018.003-T`, `018.003.001-ST`, `018.003.002-ST`

**Any descendant declaring `status: archived` that is not on this list ⇒ HALT.**
The list is a **release-instance disposition and belongs to this plan** — it is **never**
written into `_ship.agent.md` or into P-010.

**`021-S` has no pre-archived descendants at any point.** Its manifest is
`[022-F, 022.001-T]`; `022.001-T` reaches `done` at step 4 and is archive-located by routing,
which makes it **completed**, not pre-archived.

### 3.4 Delegation and output validation

**Authority split:** the **installed `shipment-reconcile` skill is the RUNTIME authority**;
**P-015 is the STATICALLY REVIEWED authority**. Ship never adjudicates between them, never
re-derives a classification, and performs **no** runtime `policy_id`, policy-version or
source-disagreement check.

#### 3.4.1 The delegation target is `mode: safe-close` Step 0(c) — there is no separate classify API

**Measured against the installed skill.** `.github/skills/shipment-reconcile/SKILL.md` exposes
exactly four modes — `pre`, `post`, `safe-close`, `detect-mixed-role` — and **no public
classification entry point**. The close-path classification is **internal to `mode:
safe-close`, at Step 0(c)**, which selects `CASCADE` or falls the whole manifest back to the
safe-close steps 1–10. The **Cascade Close Sub-Procedure** runs *"only when Step 0 of Safe-Close
Mode above selects `CASCADE`"*.

**Therefore the executable delegation is:**

1. Ship invokes the skill with **`mode: safe-close`**, `shipment_id` and `merge_commit_sha`.
   This is the **existing, installed** call already prescribed at `_ship.agent.md` step 1.b.
   **Ship does not call, and this plan does not require, any classification API.**
2. **Step 0(c) selects the close path internally.** Ship supplies no verdict, no predicate and
   no hint, and never pre-computes the selection.
3. Ship **reads the selected verdict back from the skill's own report** — the
   reconciliation report at `.backlogit/reconcile/{shipment_id}-{mode}-{timestamp}.md`, whose
   cascade-close report step 5 *"records the classifier's verdict"*. This is a **read of a
   produced artifact**, not a second decision.
4. Ship **validates that reported verdict against a fixed allowlist** and does nothing else
   with it.

**`_ship.agent.md` MAY contain an output-validation allowlist of the installed CLOSE VERDICT
tokens — `CASCADE` and `SAFE_CLOSE` — and nothing more.** It **MUST NOT** contain any
classification predicate: no root/`parent_id` test, no coverage test, no per-member rule, no
fallback rule. Comparing a reported token against a list is validation; deciding which token
applies is classification, and stays inside Step 0(c).

* **`HALT` and `RECONCILE_FAIL` are non-close procedural outcomes**, not close verdicts. They
  authorize no close path.
* **`BLOCKED` is not an installed token** and is handled as unknown.
* **The skill's `recommendation` field is a separate axis** (`PROCEED`, `CLOSED`, `HALT — …`)
  and is **never** read as a close verdict.

**Ship HALTs, no mutation, on any of:** a reported token outside the close-verdict allowlist
(including `BLOCKED` and any future token); an absent, empty, unparseable or ambiguous report;
a skill that is not installed or not readable; or a non-close procedural outcome.

The allowlist is bound to the **currently installed** skill. Widening it requires its own
review; this plan authorizes **no new verdict and no new close path**.

#### 3.4.2 Classifier integrity (byte-identity, pinned)

Before the close call, Ship verifies the installed skill file is the reviewed one:

```text
skill_blob_expected := 2026-09-13 pinned value in §8.1
skill_blob_current  := git hash-object -- .github/skills/shipment-reconcile/SKILL.md
classifier identity <=> skill_blob_current == skill_blob_expected
```

A mismatch ⇒ **HALT**, no close mutation, return to Stage. A skill that changed after review is
an unreviewed classifier, regardless of whether the change looks benign.

### 3.5 Release-instance execution contract (this plan only)

`_ship.agent.md` carries **no shipment IDs and no expected-verdict rule** — only the general
delegation above. The expectation below is a **precondition of this plan and its records**:

> For **`021-S`** and **`017-S`**, Step 0(c) is **expected to select `CASCADE`**, which enters
> the **pre-existing** Cascade Close Sub-Procedure. **A reported `SAFE_CLOSE`, an unknown
> token, an unavailable or unreadable skill, an empty or ambiguous report, `HALT`, or
> `RECONCILE_FAIL` ⇒ HALT BEFORE ANY CLOSE MUTATION** and return to Stage.
> **There is no `SAFE_CLOSE` fallback on these routes.**

A `SAFE_CLOSE` here means the live manifest is not the shape this plan asserts — a planning
error only Stage can reconcile.

### 3.6 Authority-escalation rule

A context reload MUST NOT widen the authority envelope of the session performing it. The
session that merges this change **may not use the authority it introduces**; it checkpoints
and ends.

## 4. Clause inventory — 10 sites (all mandatory) + 2 additive

**Apply by ANCHOR TEXT. Line ordinals are NOT part of this contract.** Every site below is
located by a unique anchor string measured in the current tree. Ordinals appear only in the
dated observation table at the end of this section, and an executor **MUST NOT** use them to
locate, bound or range a site. This is deliberate: a site expressed as an ordinal *range*
creates untouched gaps between ranges, and one such gap is exactly how a stale claim survived
a prior sweep.

| Site | File | Anchor text (unique) | Change |
|---|---|---|---|
| **CS1** | `_ship.agent.md` | ``Mandatory pre-self-close context reload`` | Also reload **P-015**; verify merged tokens **and the merge commit**; add the §3.6 carve-out, scoped to **Role Boundary** changes only |
| **CS2** | `_ship.agent.md` | ``<shipment_id> --status shipped`` | Delete the prescription (the engine refuses it for shipments, exit 9). Replace with a pointer to the skill's authoritative close sequence |
| **CS3** | `_ship.agent.md` | ``Do NOT call `backlogit shipment ship``` | Replace the named single exception with: do not call it unless **the verdict reported by `mode: safe-close` Step 0(c)** authorizes it |
| **CS4** | `_ship.agent.md` | ``P-015 verified fully-covered-root exception (select the close path from the`` | **Delete the re-derived classification prose.** Replace with generic delegation per §3.4.1 |
| **CS5** | `_ship.agent.md` | ``in place of the safe-close sequence above for this shipment's`` | Remove the single-exception framing (subsumed by CS4) |
| **CS6** | `_ship.agent.md` | *(new bullet, inserted after CS4's replacement)* | **NEW** delegation bullet: Ship invokes whichever close path the reported verdict names, and no path absent from that verdict |
| **CS7** | `_ship.agent.md` | ``every manifest item is present in `.backlogit/queue/` with the`` | **Pointer-only delegation** to the installed `mode: pre` contract at the `expected_status` already named one line above; continue **only** on an authoritative `PROCEED`. **PRESERVE** the orphan scan and the adjacent `Scope note (139-F/139.001-T)`; **ADD** the `record-consistent` conjunct |
| **CS8** | `_ship.agent.md` | ``queue with `status: done`, and scans for orphan items.`` | Same pointer-only delegation at `expected_status: done`. **PRESERVE** the single-writer lock clause and the orphan scan; **ADD** the `record-consistent` conjunct |
| **CS9** | `_ship.agent.md` | ``proves the protected set and halts fail-closed on any cascade or provenance`` | **Scope the claim to the path on which it is true.** The installed skill records a protected set on the **safe-close** path only; a manifest qualifying for `CASCADE` **has no protected set by construction**. Rewrite the bullet to say so explicitly |
| **CS10** | `workflow-policies.md` | ``Claim shipments, move tasks to active/done, close shipments, archive completed items`` | **P-010 `Ship MAY` list.** Extend this bullet with the §3.2 narrow covering-feature grant, in wording that matches ADD-1 |
| **ADD-1** | `_ship.agent.md` | ``\| Backlog \| Claim shipments, move tasks to active/done`` | Role Boundary row — the §3.2 narrow grant |
| **ADD-2** | `_ship.agent.md` | ``a0. **TOPOLOGY_GATE: lifecycle`` | **Step 6.1(a1)** — the covering-feature completion gate (§5), inserted **after** `a0` and **before** `a` |

**CS7/CS8 MUST NOT state any per-item classification predicate — locational or
declared-status.** This plan asserts nothing about the installed pre-mode's internal taxonomy;
it depends only on the **outcome** `PROCEED`.

**CS9 closes a gap, and the gap is the lesson.** The protected-set bullet sits *between* CS2's
and CS3's anchors. Under an ordinal-range inventory it belonged to neither site and would have
survived H1 unmodified, shipping a claim that is false on the cascade path. It is now its own
anchored site. **This is why §4 is anchor-addressed and not ordinal-ranged.**

**CS10 is bounded.** It edits exactly one bullet of P-010's `Ship MAY` list. It adds no policy,
no clause, no `Ship MUST NOT` change, and no release-instance ID. The 11-ID allowlist of §3.3
stays in this plan and never enters P-010 (asserted negatively by test row 20b).

### 4.1 Pinned replacement wording (normative — reachable by the test literals)

* **CS7** → *"This delegates the per-item check to the `shipment-reconcile` skill's
  `mode: pre` contract at the `expected_status` above and **defers every per-item status
  decision to that skill's `mode: pre` classification**, continuing **only** on an
  authoritative `PROCEED`; it also requires a `record-consistent` shipment record and **scans
  for orphan items**."*
* **CS8** → *"This acquires the single-writer lock on `.backlogit/queue/{shipment_id}.md`
  (via the `file-lock` skill) and **defers every per-item status decision to the `mode: pre`
  classification at `expected_status: done`**, continuing **only** on an authoritative
  `PROCEED`, requires that the shipment-record-status classification is `record-consistent`,
  and **scans for orphan items**."*
* **CS9** → *"**proves the protected set on the safe-close path**, where a protected set
  exists; a manifest qualifying for the cascade close path has **no protected set by
  construction**, and the skill halts fail-closed on any cascade or provenance ambiguity on
  either path."*
* **CS10** → *"Claim shipments, move tasks to active/done, **complete one covering feature
  `active -> done` when every one of its live descendants at every depth is complete and all
  five conditions in the Ship Role Boundary hold**, close shipments, archive completed items"*

### 4.2 Dated ordinal observation (NON-NORMATIVE)

Measured 2026-09-13 against `.github/agents/_ship.agent.md` (1111 lines) and
`.github/policies/workflow-policies.md`. **These ordinals are an observation, not an
invariant, and MUST NOT be used to locate a site.** They exist so a reviewer can spot-check
that each anchor resolves.

| Site | Observed line(s) |
|---|---|
| CS1 | 756 |
| CS2 | 790 |
| CS3 | 794 |
| CS4 | 802 |
| CS5 | 822 |
| CS7 | 255 |
| CS8 | 777 |
| CS9 | 792 |
| CS10 | 241 (`workflow-policies.md`) |
| ADD-1 | 38 |
| ADD-2 | 767 |

## 5. Step 6.1(a1) — ordered, total, non-overlapping selector

Positioned **after `a0`** and **immediately before `a`**. Runs **BEFORE** pre-mode and is
**self-sufficient** — it reads declared status, containment, topology and the §3.3 allowlist
directly. It **must not** require pre-mode to have run, and must not invoke it.
It answers exactly one question: *is this feature's obligation discharged?* It computes no
close path and emits no verdict.

**Read-only during evaluation; one authorized mutation.** All gathering and all guard
evaluation (S0–S3, S5) are **reads**. The **only** write a1 may perform is **S4's single
`active -> done` transition on one covering feature**, and only when all five §3.2 conditions
hold. a1 performs no other write of any kind.

Execute **S0 → S5 in order**; stop at the first that applies. No input satisfies two steps.

| Step | Guard | Action |
|---|---|---|
| **S0** | **Anomaly gate, before `n` is computed.** Any member with missing/unresolvable `artifact_type`; present in **both** queue and archive; present in **neither**; **declaring `status: archived`** with malformed provenance; **declaring `status: archived`** and **not** on the §3.3 allowlist; or a descendant-graph enumeration that errored or was incomplete (§3.2.1) | **HALT**, no mutation |
| **S1** | `n == 0` (no feature member) | Record `A1_NOT_APPLICABLE`; proceed to `a`. **A success, not a skip** |
| **S2** | `n > 1` | **HALT**, no mutation. Record `A1_MULTIPLE_FEATURE_MEMBERS: {ids}` and return to Stage |
| **S3** | `n == 1`, live status **`done`** | Re-evaluate conditions 1–4. All hold ⇒ record `A1_ALREADY_DONE`, proceed to `a` without re-issuing the transition. Any fail ⇒ **HALT** |
| **S4** | `n == 1`, live status **`active`** | All five §3.2 conditions hold ⇒ `active -> done` via **`backlogit_move_item`** (CLI fallback `backlogit move {id} --status done`), record the transition. Any unmet ⇒ **HALT**, naming the unmet condition |
| **S5** | `n == 1`, any other status | **HALT**, record the observed status, return to Stage |

`n` = count of manifest members whose `artifact_type` is `feature`, **never** by ID suffix.

**S2 is a HALT, aligned with P-015.** P-015's fully-covered-root exception is quantified over
**every** feature member, and a manifest failing the quantifier disqualifies the **whole**
manifest. §3.2 grants exactly **one** transition at **one** position, so `n > 1` is outside the
grant entirely. Recording and continuing would leave an unsatisfiable state in flight; the
fail-closed reading is the only one consistent with both.

**S0 explicitly does NOT halt on** an archive-located member declaring `status: done`. That is
a normal completed member: no allowlist check, no provenance requirement.

**Executable surface (corrected — measured against the installed engine):**

| Property | Mechanism | Note |
|---|---|---|
| Containment (torn duplicate / missing) | **`backlogit doctor`** | The **only** operation that detects duplicate IDs across queue and archive. `backlogit doctor --help`: *"duplicate IDs across queue and archive directories."* |
| Orphans | **`backlogit doctor`** | Same scan reports child types with no parent |
| Declared `status`, `artifact_type`, `parent_id`, provenance | `backlogit_get_item` | Per-record frontmatter reads |
| Full descendant graph (§3.2.1) | Repeated `backlogit_get_item` over `parent_id` from `.backlogit/queue/` ∪ `.backlogit/archive/` | Walked live, at every depth |

**`backlogit_query_sql` MUST NOT be used for containment.** Measured: the index `items` table
carries **no path or location column**, and `backlogit get --json` exposes no path, so a torn
duplicate is **invisible** to it. `backlogit_get_item` alone is equally insufficient — it
returns one record and cannot reveal a second copy. **Only `backlogit doctor` closes this.**

## 6. Go contract test — 20 rows

One table-driven function, ≤4 helpers, **stdlib assertions only** (`testing`, `strings`), no
`testify`, reusing the package-level `repoRoot(t)` from `tests/integration/build_script_test.go`.
Rows 1–19 read `.github/agents/_ship.agent.md`; row 20 reads
`.github/policies/workflow-policies.md`.

| # | Assertion | Kind |
|---|---|---|
| 1 | Role Boundary row contains the narrow feature-completion grant | positive |
| 2 | The grant names all five conditions | positive |
| 3 | Step 6.1(a1) exists, positioned **after** `a0` and **before** `a` | ordering |
| 4 | The delegation block is present and names P-015 + `shipment-reconcile` | positive |
| 5 | CS1's reload clause also names P-015 and the merge commit | positive |
| 6 | The authority-escalation rule is present | positive |
| 7 | The drifted `children`-only coverage rule is gone | negative |
| 8 | `_ship.agent.md` does **not** name `TASK_ONLY_FINALIZE` | negative |
| 9 | No classification **predicate** — a close-verdict allowlist is permitted, a selection condition is not | compound +/− |
| 10 | The `backlogit move <shipment_id> --status shipped` prescription is gone | compound +/− |
| 11 | CS1's carve-out is scoped to **Role Boundary** changes only | consistency |
| 12 | a1 states the `n == 0` no-op (S1) **and** that an unmet guard halts (S4) | consistency |
| 13 | a1 states idempotent resume after conditions 1–4 revalidate (S3) | consistency |
| 14 | **(CS7)** delegation phrase **PRESENT** **and** old queue-only literal **ABSENT** | compound +/− |
| 15 | **(CS8)** delegation phrase **PRESENT** **and** old queue-only literal **ABSENT** | compound +/− |
| 16 | a1 states `n > 1` halts with no mutation (S2) **and** the anomaly-gate-before-`n` rule (S0) | compound ++ |
| 17 | CS8's sentence **PRESERVES** the single-writer lock **and** the orphan scan | preservation |
| 18 | CS7's adjacent `Scope note (139-F/139.001-T)` **survives** | preservation |
| 19 | **(CS9)** the protected-set claim is **scoped to safe-close** **and** the unscoped literal is **ABSENT** | compound +/− |
| 20 | **(CS10)** P-010's `Ship MAY` list carries the grant (20a) **and** `workflow-policies.md` contains no `018.` literal (20b) | compound +/− |

**Kinds**: `compound +/−` = {9, 10, 14, 15, 19, 20} (**6**); `compound ++` = {16};
`negative` = {7, 8}.

### 6.1 Pinned literals — ALL 20 rows (normative; single-line, CRLF-safe)

**Every row is pinned.** No row derives its needle from prose, and no acceptance criterion
claims verification for a site without a needle here. Counts are measured at the current tree
(H0) over the file named in the row.

| Row | Site | MUST BE ABSENT | MUST BE PRESENT | H0 |
|---|---|---|---|---|
| 1 | ADD-1 | — | ``complete one covering feature `active -> done``` | 0 → RED |
| 2 | ADD-1 | — | ``all five conditions are conjunctive`` | 0 → RED |
| 3 | ADD-2 | — | ``a1. **Covering-feature completion gate**`` (index **>** `a0. **TOPOLOGY_GATE: lifecycle`, **<** `a. **Pre-archive reconciliation gate`) | 0 → RED |
| 4 | CS4/CS6 | — | ``the `shipment-reconcile` skill's `mode: safe-close` Step 0 selects the close path`` | 0 → RED |
| 5 | CS1 | — | ``also re-read `P-015` and verify the merge commit`` | 0 → RED |
| 6 | CS1 | — | ``A context reload MUST NOT widen the authority envelope`` | 0 → RED |
| 7 | CS4 | ``it is fully covered (every one of its children,`` (×1) | — | 1 → RED |
| 8 | guard | ``TASK_ONLY_FINALIZE`` — already ×0; **must stay ×0** | — | GREEN |
| 9 | CS4 | ``qualification is never per-member, and no feature ID is ever special-cased.`` (×1) | ``HALT on any result token the installed classification did not name`` | 1 / 0 → RED |
| 10 | CS2 | ``<shipment_id> --status shipped`` (×1) | ``the authoritative close sequence defined by the `shipment-reconcile` skill`` | 1 / 0 → RED |
| 11 | CS1 | — | ``scoped to Role Boundary changes only`` | 0 → RED |
| 12 | ADD-2 | — | ``A1_NOT_APPLICABLE`` **and** ``HALT, naming the unmet condition`` | 0 / 0 → RED |
| 13 | ADD-2 | — | ``A1_ALREADY_DONE`` | 0 → RED |
| 14 | CS7 | ``every manifest item is present in `.backlogit/queue/` with the`` (×1) | ``defers every per-item status decision to that skill's `mode: pre` classification`` | 1 / 0 → RED |
| 15 | CS8 | ``queue with `status: done`, and scans for orphan items.`` (×1) | ``defers every per-item status decision to the `mode: pre` classification at `expected_status: done``` | 1 / 0 → RED |
| 16 | ADD-2 | — | ``evaluated before `n` is computed`` **and** ``halts with no feature mutation`` | 0 / 0 → RED |
| 17 | guard | — | ``scans for orphan items`` — already ×2 (L256, L777); **must stay ≥2** | GREEN |
| 18 | guard | — | ``Scope note (139-F/139.001-T)`` — already ×1 (L259); **must stay ×1** | GREEN |
| 19 | CS9 | ``proves the protected set and halts fail-closed on any cascade or provenance`` (×1) | ``proves the protected set on the safe-close path`` **and** ``no protected set by construction`` | 1 / 0,0 → RED |
| 20a | CS10 | — | ``complete one covering feature `active -> done``` in **`workflow-policies.md`** | 0 → RED |
| 20b | guard | ``018.`` in **`workflow-policies.md`** — already ×0; **must stay ×0** | — | GREEN |
| — | guard | ``018.`` in `_ship.agent.md` — already ×0; **must stay ×0** | — | GREEN |

**Two classes, and only one is red → green:**

| Class | Rows | H0 state |
|---|---|---|
| **Transition rows** — must change | 1–7, 9–16, 19, 20a (**16**) | **RED.** None can be vacuously green |
| **Guard rows** — must NOT change | 8, 17, 18, 20b, and the `_ship.agent.md` `018.` guard (**5**) | **GREEN today, by design.** They assert a property that already holds and must survive |

**The guard rows are green at H0 and that is correct** — their purpose is to fail if the
implementation *breaks* something, not to record a transition. Row 3's `a1` anchor is absent
pre-implementation **by design**; that absence is the H0 readiness guard.

**Only the transition rows constitute the red → green evidence** for AC-3/AC-8/AC-24.
Do not read the guard rows as proof the contract changed.

**Independently confirmed at attempt 5 and re-measured at revision 10** (§13.2 audit pointers):
every ABSENT needle above resolves ×1 and every PRESENT needle ×0 in the live tree;
`TASK_ONLY_FINALIZE` ×0; `018.` ×0 in both files; `scans for orphan items` ×2;
`Scope note (139-F/139.001-T)` ×1; `record-consistent` ×0 in `_ship.agent.md` (it is **added**,
not preserved). `repoRoot(t)` exists at package scope in
`tests/integration/build_script_test.go`; `testify` is absent from `go.mod`.

## 7. Ordering

**Test-first, `main` stays green.** H0: write the 20-row test → **red**. H1: apply ADD-1,
ADD-2 and **CS1–CS10** → **green**. All ten clause sites are required to reach H1.

## 8. Pre-claim evidence — Probe 25 (DISCHARGED)

Authored, executed and **committed by Stage before `021-S` is claimed**. **Ship never authors,
modifies, re-runs or regenerates it.**

`docs/plans/evidence/2026-09-12-task-only-shipment-finalization/probe25-root-included-cascade-fixture.{ps1,txt}`

**Result**: `PROBE25_RESULT=PASS`, `FAILED_CRITERIA=0`, `DIGEST_GATE=PASS`, over the exact
13-member root-included manifest with 11 members declaring `status: archived`
(`archived_status: queued`): `TOPOLOGY_MEMBERS=13`, `ROOT_INCLUDED=True`, `PRE_ARCHIVED=11`,
`ARCHIVED_STATUS_FLAVOUR=queued`, `PREMODE_RESULT=PROCEED`.

**The nine STEP 11 aggregate criterion tokens** (these exact names — **not** the eight inline
`CRITERION_n` tokens): `C1_DIGEST_GATE`, `C2_PREMODE_PROCEED_11_PRE_ARCHIVED`,
`C3_CLASSIFICATION_CASCADE`, `C4_RETURNED_IDS_EMPTY`, `C5A_ARCHIVED_MINUS_ALLOWED_EMPTY`,
`C5B_REQUIRED_MINUS_ARCHIVED_EMPTY`, `C6_PARENT_ID_PRESERVED`, `C7_SHIPMENT_ARCHIVED_SHIPPED`,
`C8_NOTHING_OUTSIDE_MANIFEST_TOUCHED`.

**What each criterion measures** — criteria **2 and 3 are PowerShell-derived** simulations of
the pre-mode and close classification; criterion **1** is a real digest; the **authorized
cascade call** is the transcript's `### STEP 8: AUTHORIZED CALL`, and criteria **4–8** observe
that call's real output and post-call state. The central claim — the cascade archives exactly
the manifest and preserves every `parent_id` — rests on the **engine-observed** criteria.

**What it does not prove**: it measures the engine on a faithful fixture. It is not a claim
about the live records at closure time, which is why §9 step 12a is an **exact** topology match
that HALTs on mismatch.

### 8.1 Evidence identity — plan-pinned, not transcript-derived

Identity is established by **Git blob object IDs**, which are byte-exact, tool-native and
reproducible.

```text
committed_blob(f) := git rev-parse {evidence_commit_sha}:{f}
current_blob(f)   := git hash-object -- {f}
file identity holds <=> committed_blob(f) == current_blob(f) == the PINNED value below
```

**PINNED CONSTANTS (measured 2026-09-13; part of this reviewed contract):**

| Object | Pinned value |
|---|---|
| `evidence_commit_sha` | `e36d853639899492c49ff3f9b1f2c7879bc41807` |
| blob — `…probe25-root-included-cascade-fixture.ps1` | `81d85454d920b87b9270d62a06b3f031dd85e190` |
| blob — `…probe25-root-included-cascade-fixture.txt` | `334cf3be007c15f1fec28760dcc4ee8568768b43` |
| `ENGINE_SHA256_EXPECTED` (backlogit executable) | `1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98` |
| blob — `.github/skills/shipment-reconcile/SKILL.md` (§3.4.2) | `312f29b6e393bc9cffec411c386f2f8fc4454e50` |

#### 8.1.1 Why the pinning is in THIS file and not only in the transcript

An ancestry check alone proves only *"last touch is not after the merge"*. A **same-PR**
replacement of both Probe-25 files satisfies ancestry; and if the expected engine digest is
read **out of the replaced transcript**, the forgery is self-consistent and passes every check.
**The transcript cannot be the sole authority for its own identity.**

The pinned table above breaks that loop. The values live in the **governing plan**, which is
itself a reviewed artifact under a plan-review gate. Forging the evidence now additionally
requires editing this table, which is a **visible, reviewed diff** rather than a silent
substitution.

**Three independent anchors must agree, and all three are checked (§9.2):**

1. **Ancestry** — `evidence_commit_sha` is an ancestor of the merge commit.
2. **Plan-pinned identity** — both blob OIDs equal the values pinned above, **read from this
   plan**, never from the transcript.
3. **External engine identity** — the resolved engine digest equals the pinned
   `ENGINE_SHA256_EXPECTED` **read from this plan**. It is compared against the transcript's
   own `ENGINE_SHA256_EXPECTED` as a **cross-check**; a divergence between the two ⇒ **HALT**,
   because it means one of the two authorities was altered.

`evidence_commit_sha` remains the **LAST commit touching either file**, not the introducing
one — that is what closes post-merge replacement, since replaced evidence moves the sha to a
**descendant** of the merge.

## 9. Closure sequence

`021-S` = `[022-F, 022.001-T]` — a fully-covered root (`022-F` has no `parent_id`; its complete
live descendant graph at every depth is exactly `{022.001-T}`; the manifest contains nothing
else). Verified against §3.2.1's set-equality rule.

| Step | Session | Action |
|---|---|---|
| 0a | **Stage** | Probe 25 authored, executed, committed. **DISCHARGED.** `021-S` not claimable until PASSing |
| 1 | **S1** | **Step 0.5 Shipment Intake.** Verify on `main`; **P-011** branch-before-mutation + **P-016** single-worktree check; pre-claim topology gate; **claim `021-S`** |
| **1a** | S1 | **POST-CLAIM CONDITION — `022-F` must be `active` (checked, not assumed).** Re-read `022-F` after the claim and require its live status to be **exactly `active`**. **Any other value ⇒ HALT** and return to Stage. **Ship is granted NO authority to transition a feature `queued -> active`** — §3.2 grants exactly one transition, `active -> done`. *(Live state at planning time is `queued`; the shipment claim is what is expected to activate it. If it does not, that is a planning-state error only Stage can reconcile.)* |
| **1b** | S1 | **Step 2 Harness Generation (P-002 / P-004).** The **harness-architect** produces the 20-row harness: `go vet ./...` exits 0, `go test ./...` exits non-zero with expected failure markers ⇒ **H0 RED CONFIRMED** ⇒ `harness-ready` applied to `022.001-T`. **This precedes the TASK claim** — P-002 permits claiming a task only after red-phase confirmation |
| **1c** | S1 | **Step 3 Build Ready Queue** — filtered to tasks carrying `harness-ready` |
| **1d** | S1 | **Step 4.1 Claim Task** — `022.001-T -> active` |
| 2 | S1 | **Step 4.2 implement** → apply ADD-1, ADD-2 and **CS1–CS10** → **H1 green** |
| 3 | S1 | Quality gates; review gate |
| 4 | S1 | Complete Task — commit the implementation, then **`022.001-T -> done`**, then **commit the resulting `.backlogit/` change** and verify a clean worktree. **Both commits must be in the PR that merges at step 5** — otherwise merged `main` still shows the task `queued` and a1 halts at step 13 |
| 5 | S1 | PR lifecycle — gates, build, push, PR, **P-018 Copilot engagement**, operator approval, **merge** |
| 6 | S1 | Merge Confirmation Gate — `gh pr view` `MERGED`; `git fetch origin main`; `git merge-base --is-ancestor {merge_sha} origin/main` |
| 7 | S1 | Reload merged `main`; detect the Role Boundary and P-010 changes; per §3.6 the new authority is **not** available to this session |
| 8 | S1 | **Write checkpoint** (§9.1) and **END**. Record the emitted filename **and** `session_id` |
| 9 | **S2** | **Fresh session.** Run the **installed** crash-resumption protocol (§9.1). **All reads and in-memory work; NO mutation.** The **resolve is deferred to step 11a** |
| 10 | S2 | Verify merge SHA is an ancestor of `origin/main`; validate the same active shipment. **Reads only** |
| 11 | S2 | **Post-Merge Branch Protocol (P-011)** — `git checkout main`, pull, `git checkout -b post-merge/022-ship-feature-completion`. **This PRECEDES EVERY S2 MUTATION**, including the checkpoint resolve at 11a. `main` is never the active branch for any mutation |
| **11a** | S2 | **Resolve the checkpoint** (§9.1) — the **first** S2 mutation, performed **on the closure branch**, and only after the resume at step 9 was confirmed successful |
| 12 | S2 | `--phase lifecycle` topology gate — **while `021-S` is still `active`**; never re-run after the close |
| **12a** | S2 | **READ-ONLY re-verification** of Stage's committed evidence (§9.2). FAIL ⇒ HALT, `021-S` stays `active`, no a1 mutation |
| 13 | S2 | **Step 6.1(a1)** — §5 **S4**. `022.001-T` declares `done` (archive-located = completed, not pre-archived) ⇒ all five conditions hold ⇒ `022-F active -> done` via `backlogit_move_item` |
| **13a** | S2 | **PRE-CASCADE BASELINE COMMIT (mandatory).** Commit a1's `.backlogit/` mutation **on the existing closure branch** — no new branch, no new worktree. **`{pre_cascade_sha}` := this commit.** Then satisfy §10.3 step 1 **against this commit**: verify `git status --porcelain -- .backlogit/queue/ .backlogit/archive/` is **empty**, and capture `{pre_paths}` from that clean baseline per §10.3.1. **Without 13a the trees are dirty from step 13 and the cascade MUST NOT be invoked** |
| 14 | S2 | **Pre-mode**, `expected_status: done`. **Requires an authoritative `PROCEED`**; anything else HALTs |
| 15 | S2 | **Invoke `mode: safe-close`** (§3.4.1). Step 0(c) **must select `CASCADE`** (§3.5) ⇒ the pre-existing Cascade Close Sub-Procedure runs; read the verdict back from the produced report and validate it against the allowlist; P-007 archive-integrity verify; post-mode. **Capture `{post_paths}` immediately on return, before any commit or cleanup** (§10.3.1). **The cascade output is committed ONLY after every postcheck passes** (§10.3) |
| 16 | S2 | Operational closure → `docs/closure/`; P-020 compact-context — on the closure branch, before the closure PR is pushed |
| 17 | S2 | Sync — **MCP `backlogit_sync_index` first**, CLI `backlogit sync` as declared fallback — then push, closure PR, **P-018 Copilot engagement**, local review, operator approval, merge |
| 18 | S2 | Return to `main`; pull |
| 19 | — | Orchestrator routes **`017-S`**. **No `017-S` execution step exists in this plan** — see §11.1 |

> **Step ordering is derived from the installed Ship contract, not invented.** `_ship.agent.md`
> orders **Step 0.5 Shipment Intake** → **Step 1 Pre-Flight** → **Step 2 Harness Generation
> (P-002/P-004)** → **Step 3 Build Ready Queue** → **Step 4.1 Claim Task**. The **shipment**
> claim therefore correctly precedes harness generation; only the **task** claim must follow
> red-phase confirmation. Steps 1 → 1d above reproduce that order exactly.

**P-018 engagement is mandatory before BOTH merges** (steps 5 and 17) and the window closes
permanently at merge. Use the workspace-verified `[bot]`-suffixed REST fallback; judge blocking
state from the paginated GraphQL `reviewThreads`, never from a review body's "Suppressed
comments". See `docs/compound/2026-09-06-requesting-copilot-review-non-collaborator-repo.md`,
`docs/compound/2026-09-10-p018-copilot-engagement-must-precede-merge.md`, and
`docs/compound/2026-09-11-copilot-suppressed-comments-vs-review-threads.md`.

### 9.1 S2 checkpoint recovery — the INSTALLED protocol governs (pointer-only)

**This plan states NO recovery algorithm of its own.** S2 executes the
**Crash-Resumption / Startup Recovery Protocol as installed in
`.github/agents/_ship.agent.md`**, verbatim and unmodified, including:

* `backlogit_list_checkpoints` called **with `consumer_id: "ship"` and NO `status` or `agent`
  filter** — exactly as the installed step 1 prescribes;
* the **validation/quarantine anomaly gate FIRST**, over the full enumeration;
* **EXPLICIT OPERATOR SELECTION** of a single checkpoint by filename — never auto-picked, even
  when exactly one candidate exists;
* **OWNER VALIDATION** that `agent` is exactly `ship`;
* **EXPLICIT OPERATOR CONFIRMATION** before any restore or prune;
* **bounded prune-on-restore** per the backlogit-pack overlay instruction, which **never
  prunes** any of the **three** never-prune classes: (1) the **active-shipment / active-task
  cursor**, (2) the **unresolved-checkpoint pointer** itself, (3) **recorded gate verdicts**.
  Engram unreachable ⇒ **FAIL CLOSED**, no prune, no resume;
* **OWNER-SCOPED RESOLUTION** only after a confirmed successful resume, for that one
  checkpoint only, never a bulk sweep and never a `stage`-owned record;
* **FAIL CLOSED — no fresh-start fallback.**

**A prior revision of this plan stated a competing algorithm** (unfiltered enumeration, no
operator selection, no confirmation). It is **withdrawn in full**. Where this plan and the
installed protocol could differ, **the installed protocol governs**; this plan is not an
override and must not be read as one.

**This plan adds exactly two release-instance obligations, and nothing else:**

1. **S1 records** the emitted `checkpoint-YYYYMMDD-HHMMSS.json` filename **and** its
   `session_id` in the handoff. **S2 carries both as the EXPECTED identity** and presents them
   to the operator alongside the candidate list.
2. **If the operator-selected checkpoint does not match that expected identity on BOTH
   filename and `session_id` ⇒ HALT** and return to Stage. This is an *additional* fail-closed
   check layered on top of the installed protocol; it never substitutes for operator selection
   and never permits auto-selection.

**Resolve timing (P-011 overlay).** `backlogit_resolve_checkpoint` is a **backlog mutation**,
so on this route it runs at **§9 step 11a**, **on the closure branch**, never on `main`. The
installed protocol's ordering constraint — *resolve only after a confirmed successful resume* —
is **unchanged**; only the *position* of the write moves later.

**Checkpoint payload**: `schema_version: 1`, `agent: ship`, `session_id`, `phase`,
`resume_hint` naming `021-S`, `022-F`, `022.001-T` and the merge SHA; domain data under
`context`. Write it with the official create operation, passing the state dump **inline**.

### 9.2 S2 read-only re-verification — seven checks

**S2 authors nothing, modifies nothing, re-runs nothing.** Every check is a read
(`git log`, `git rev-parse`, `git hash-object`, `git merge-base`, `git diff --quiet`,
`Get-FileHash`, text inspection).

1. **Exact topology** — `TOPOLOGY_MEMBERS=13`, `ROOT_INCLUDED=True`, `PRE_ARCHIVED=11`,
   `ARCHIVED_STATUS_FLAVOUR=queued`. Exact match; any other topology means the measurement is
   not about `017-S`.
2. **PASS fields** — `PROBE25_RESULT=PASS`, `FAILED_CRITERIA=0`, and all **nine pinned STEP 11
   tokens** (§8) read `PASS`.
3. **Ancestry** — both files tracked, and `evidence_commit_sha` equals the §8.1 pinned value
   **and** is an **ancestor of the merge commit** verified at step 6.
4. **Engine identity** — the currently-resolved engine digest equals the §8.1 **plan-pinned**
   `ENGINE_SHA256_EXPECTED`, **and** the transcript's own `ENGINE_SHA256_EXPECTED` equals the
   same pinned value. Divergence between the two authorities ⇒ **HALT** (§8.1.1). No absolute
   path is required.
5. **File identity** — both committed blobs equal their working-tree files **and** equal the
   §8.1 **plan-pinned** OIDs, hashed **separately**; neither file missing, empty or modified.
6. **Internal consistency** — `PREMODE_MATCHED + PREMODE_PRE_ARCHIVED + PREMODE_MISSING +
   PREMODE_STATUS_MISMATCH + PREMODE_DUPLICATE_TORN == MANIFEST_COUNT`;
   `PRE_ARCHIVED_COUNT == PREMODE_PRE_ARCHIVED`; `TOPOLOGY_MEMBERS == MANIFEST_COUNT`;
   `FAILED_CRITERIA` agrees with the per-criterion tokens. *(Current: `2+11+0+0+0 == 13`,
   `11 == 11`, `13 == 13`, `0`.)*

   > **Scope of check 6 (explicit, so it is not read as a taxonomy claim).** These
   > `PREMODE_*` names are **fields Stage's own Probe-25 transcript emits**, and §8 records
   > that criteria 2 and 3 are **PowerShell-derived simulations**. Check 6 therefore asserts
   > only that **Stage's committed evidence file is internally self-consistent**. It is
   > **not** an assertion about how the installed pre-mode allocates any live record, and
   > nothing in this plan may be read as claiming the installed pre-mode's internal taxonomy
   > — per §3.4, this plan depends only on the **outcome** `PROCEED`.

7. **Classifier byte-identity** — `git hash-object -- .github/skills/shipment-reconcile/SKILL.md`
   equals the §8.1 pinned skill blob OID (§3.4.2). Mismatch ⇒ **HALT**.

> **Why the skill blob is safely pinnable.** `shipment-reconcile/SKILL.md` is **not** in the
> §2 three-file surface, so H1 does not modify it and the pin survives the merge unchanged.
> `_ship.agent.md` and `workflow-policies.md` **are** modified by H1 and are therefore
> deliberately **not** pinned.

## 10. Recovery

### 10.1 HALTs before the CLOSE — nothing to unwind, but a1 may already have run

a1 S0–S5; pre-mode `RECONCILE_FAIL`; a reported verdict other than `CASCADE` on these routes;
§9.2 failure. In every case **no CLOSE mutation has occurred**: `021-S` stays `active`, is not
shipped and not archived.

**Precise state, by position in §9:**

| HALT | a1 (step 13) already ran? | Feature state | Resume |
|---|---|---|---|
| §9.2 failure (step **12a**) | **No** — 12a precedes a1 | `022-F` still `active` | §5 **S4** |
| a1 S0–S5 (step **13**) | halts **within** a1, before its own S4 write | `022-F` still `active` | §5 **S4** |
| pre-mode `RECONCILE_FAIL` (step **14**) | **YES** | `022-F` is **`done`** | §5 **S3** |
| verdict not `CASCADE` (step **15**) | **YES** | `022-F` is **`done`** | §5 **S3** |

**The completed feature is PRESERVED, not rolled back.** a1's S4 transition is a correct,
independently-guarded act: it fired only because all five §3.2 conditions held, and a later
pre-mode or verdict failure does not falsify them. Reverting it would discard a valid result
and re-open a gate that already passed. **S3 is the designated resume path** — it revalidates
conditions 1–4 and proceeds without re-issuing the transition, and HALTs if they no longer
hold. `021-S` therefore remains cleanly resumable in every row above.

### 10.2 §9.2 failure — recovery depends on WHICH class failed

| Class | Checks | Evidence still valid? | Recovery |
|---|---|---|---|
| **Transient — toolchain/executable drift** | **4** (engine half only) | **YES** — only the environment moved | **Correct the drift and re-verify read-only, in place.** A clean re-verification **clears it** and A proceeds |
| **Committed evidence / identity / ancestry / classifier** | **1, 2, 3, 5, 6, 7**, and check 4's *plan-vs-transcript divergence* half | **NO** — the artifact is absent, altered, inconsistent or misplaced | **Cannot be cleared in place.** Requires **`git revert` of the merge** (after holding `017-S`) **and a re-plan** |

Re-running Probe 25 and re-committing **after** the merge moves `evidence_commit_sha` to a
**descendant** of the merge SHA, so check 3 still fails. That restores a valid posture for a
**re-planned** A or a future shipment — never for an in-flight post-merge A.

The gate is **non-bypassable** (no actor closes A without a PASS) and **recoverable** (a
failure has a named in-role exit). There is no `--force`, no "proceed with a recorded
residual", and no self-authorized waiver.

### 10.3 Post-mutation failure — mandatory VERIFIED restore

The installed cascade verifies **after** it mutates. **Every** post-CASCADE failure carries the
**same** obligation:

| Cascade check | Failure |
|---|---|
| step 2 | `returned_ids` **not empty** |
| step 3 | `archived_ids − allowed_ids` **not empty** (archived outside the manifest) |
| step 3 | `required_ids − archived_ids` **not empty** (failed to archive a required member) |
| step 4 | any `parent_id` **altered or cleared** vs the pre-call snapshot |

#### 10.3.1 The two path inventories — BOTH are defined and BOTH are captured

Every variable consumed by the restore is defined here, with the exact command and the exact
moment of capture. **A variable that is consumed but never captured is an unexecutable
rollback**, so both are stated explicitly.

| Variable | Definition | Captured WHEN | Captured HOW |
|---|---|---|---|
| **`{pre_paths}`** | The full **sorted** list of every path under `.backlogit/queue/` and `.backlogit/archive/`, tracked and untracked, as of the clean pre-cascade baseline | At **§9 step 13a**, immediately after the baseline commit and after the clean-tree check passes — **before** the cascade is invoked | `Get-ChildItem -Recurse -File .backlogit/queue/, .backlogit/archive/ \| ForEach-Object FullName \| Sort-Object` (equivalently `git ls-files` ∪ untracked, both trees) |
| **`{post_paths}`** | The same sorted list, re-enumerated over the **same two trees** by the **same command**, after the cascade returns | At **§9 step 15**, **immediately on return from the cascade call** — before any commit, cleanup, postcheck-driven restore, or further mutation | **identical command to `{pre_paths}`** |

**Both captures MUST use the identical enumeration command over the identical two roots.** A
difference in command, root set or sort order makes the set difference meaningless.

**`{post_paths}` capture is unconditional.** It is taken on **every** return from the cascade,
success or failure, because at the moment of return it is not yet known which postcheck will
fail. Capturing it only after a failure is detected is too late — intervening steps may have
already changed the trees.

**`{post_paths} − {pre_paths}`** is the deterministic set the cascade **created**.

#### 10.3.2 Obligation, in order

1. **Capture is a precondition, not a reaction — and it is §9 step 13a.** The baseline is
   established by the **named pre-cascade baseline commit at §9 step 13a**, on the **existing**
   closure branch. **If any of the following is unavailable, do not invoke the cascade.**
   a. **`{pre_cascade_sha}`** — the step-13a commit. It exists specifically because §9 step 13
      (a1's `active -> done`) dirties `.backlogit/`, so **`HEAD` before 13a is NOT the
      pre-cascade state** and restoring from it would discard a1's valid result.
   b. **Clean trees, verified against that commit** —
      `git status --porcelain -- .backlogit/queue/ .backlogit/archive/` returns **empty** at
      `{pre_cascade_sha}`. A dirty baseline makes "what the cascade created" undecidable.
   c. **`{pre_paths}`** — captured per §10.3.1, alongside the skill's in-memory Step 0(b)
      snapshot of declared `status`, `parent_id` and location for every watched record.
2. **Restore**, in this order:
   * `git restore --source {pre_cascade_sha} -- .backlogit/queue/ .backlogit/archive/`
     — restores **tracked** content; and
   * **delete exactly `{post_paths} − {pre_paths}`** — the paths present after the call and
     absent from the baseline inventory. This is the deterministic set the cascade created
     (most notably the newly written archived shipment record), which `git restore` alone
     leaves behind. **Delete nothing else**; step 1b guarantees this set contains no
     pre-existing untracked file.
3. **VERIFY the restore — both halves:**
   * **Tree equivalence** — the current sorted path list, re-enumerated with the **same
     command** as §10.3.1, equals `{pre_paths}`, and `git status --porcelain` over both trees
     is **empty** (content identical to `{pre_cascade_sha}`); **and**
   * **Record equivalence** — every record in the Step 0(b) snapshot matches it on declared
     `status`, `parent_id` and location.

   **An unverified restore is not a restore.**
4. **HALT.** No retry, no second close path, no re-invocation of the cascade.

> **The cascade result is NEVER committed before its postchecks pass (bounded state machine).**
> The authorized flow is strictly:
>
> ```text
> 13a commit baseline  ->  {pre_cascade_sha}, clean trees, {pre_paths}
> 15  invoke cascade   ->  mutates the WORKTREE only
> 15  on return        ->  capture {post_paths} IMMEDIATELY (unconditional)
> 15  postchecks       ->  returned_ids / two-set / parent_id / P-007 / post-mode
>        pass  -> commit the cascade result (first commit after 13a)
>        fail  -> restore (step 2) -> verify (step 3) -> HALT. NOTHING is committed
> ```
>
> **A committed cascade awaiting postchecks is therefore unreachable in this plan**, which is
> why §10.3 is worktree-level and sufficient. P-015's `git revert` remedy addresses the
> general case where a cascade was already committed; **on this route that state cannot
> arise**, and §10.4's `git revert` is reserved for the *merge*, not the cascade.
> **If an executor ever finds the cascade result already committed before its postchecks
> completed, that is an out-of-contract state: HALT immediately, do not restore, and return to
> the operator** — the §10.3 worktree restore is not valid against committed history.

**If the restore cannot be verified**, HALT with the backlog flagged
**operator-reconcile-required**.

This detect-after-mutate ordering is an **inherited P-015 residual**. This plan adds **no**
cascade path, **no** verdict and **no** rollback mechanism — but it **does** make that
destructive branch **reachable for `017-S`**, which previously failed pre-mode first. That new
reachability is a genuine consequence and is disclosed, not argued away. Remediating the
ordering itself is a recorded follow-up against P-015 and the skill, not charged here.

### 10.4 Rollback

`git revert` of the merge commit — never a history rewrite, never a force-push, never an amend.
`017-S` must be held first.

## 11. Strict safety — approved destructive actions

Operator authorization timestamp for both: **`2026-09-13T11:35:43-07:00`**.

> **This approval covers exactly the two actions below and nothing else.** It does **not**
> authorize admin fallback, force-merge, force-push, history rewrite, `--force`, any other
> destructive operation, any other shipment, or a re-scoped version of either action. **Any
> mismatch between a recorded condition and live state invalidates the approval and HALTs.**
> **Neither record is a claim authorization** — claiming remains gated on §13.

| Field | `PA-021-CASCADE` | `PA-017-CASCADE` |
|---|---|---|
| **Summary** | Close `021-S` via the pre-existing P-015 fully-covered-root cascade | Close `017-S` the same way |
| **Targets** | Shipment `021-S` **and its own record**; manifest exactly `[022-F, 022.001-T]` | Shipment `017-S` **and its own record**; the exact 13 members below |
| **change_kind** | **Destructive** — archives the manifest subtree and transitions the shipment | same |
| **Execution site** | **§9 steps 13a–15 of THIS plan** | **NOT IN THIS PLAN — see §11.1** |
| **Rollback** | §10.3 — pre-cascade commit + both path inventories + snapshot restore, **verified**, then HALT | §10.3, **applied within `017-S`'s own closure plan** (§11.1) |
| **approval_required** | **YES** (non-negotiable) | **YES** |
| **ActionRisk** | **`destructive`** | **`destructive`** |
| **ActionResult** | **`approved`** | **`approved`** |

**`017-S` manifest (13)**: `018-F`, `018.008-T`, `018.001-T`, `018.001.001-ST`,
`018.001.002-ST`, `018.001.003-ST`, `018.002-T`, `018.002.001-ST`, `018.002.002-ST`,
`018.002.003-ST`, `018.003-T`, `018.003.001-ST`, `018.003.002-ST`.

**Conditions — all must hold at invocation:**

1. `mode: safe-close` Step 0(c) selects **`CASCADE`** and the produced report reports it;
   every other outcome (§3.5) **voids the approval and HALTs**;
2. **all preflight gates pass** — §9.2's seven checks, the §5 S0–S5 selector, and pre-mode
   `PROCEED`;
3. the **manifest and topology are exactly as recorded**, unchanged, and satisfy §3.2.1's
   set-equality rule — `021-S`: 2 members, no dependencies; `017-S`: the 13 members above,
   `018-F` a root, all 12 descendants resolving to it at every depth, dependency
   `[{id: 021-S, type: blocks}]`;
4. **`PA-017-CASCADE` only** — the dependency is **satisfied by a shipped `021-S`**. This
   approval is **void while `021-S` is unshipped**;
5. the **clean trees, pre-cascade commit, `{pre_paths}`, `{post_paths}` capture procedure and
   Step 0(b) snapshot are all defined and available** (§10.3.1, §10.3.2 step 1);
6. **no scope expansion** — nothing outside the manifest **and the shipment's own record** is
   archived, reparented, created or deleted.

**Post-mutation failure ⇒ restore, VERIFY, HALT** (§10.3).

### 11.1 `PA-017-CASCADE` has NO execution site in this plan (explicit, non-dangling)

**This plan executes `021-S` only.** §9 steps 1–18 are entirely `021-S`; step 19 merely
**routes** `017-S` to the Orchestrator. There is therefore **no `017-S` a1 step, no `017-S`
pre-cascade baseline commit, and no `017-S` cascade invocation anywhere in this plan.**

A prior revision cross-referenced `PA-017-CASCADE` to §9 step 13a. That reference resolved
syntactically but **not to an applicable action**, because step 13a is a `021-S` step. It is
**withdrawn**.

**Corrected binding:**

* `PA-017-CASCADE` is a **recorded operator approval held in escrow**. It authorizes the
  destructive close of `017-S` **when, and only when, a Stage-authored `017-S` closure plan
  exists** that supplies the `017-S` equivalents of §9 step 13a (pre-cascade baseline commit),
  §10.3.1 (`{pre_paths}` / `{post_paths}`) and §10.3.2 (verified restore).
* **Until that plan exists and passes its own plan-review gate, `PA-017-CASCADE` is NOT
  exercisable.** This is a **third** independent ground for ineligibility, alongside (a) this
  plan's own gate state and (b) the unsatisfied `021-S` dependency.
* Authoring that `017-S` closure plan is **Stage work, out of scope here** (§16), and is
  triggered by §9 step 19.
* **The approval does not expire and is not re-litigated** by this correction; only its
  dangling execution-site reference is repaired.

## 12. Acceptance criteria

1. **AC-1** Role Boundary carries the §3.2 grant with all five conditions, fail-closed.
2. **AC-2** Step 6.1(a1) exists, positioned after `a0` and immediately before `a`.
3. **AC-3** All **ten** clause sites CS1–CS10 are applied; **every one of them has at least one
   pinned needle in §6.1** (CS1 → rows 5/6/11; CS2 → row 10; CS3 → row 9; CS4 → rows 4/7/9;
   CS5 → row 9; CS6 → row 4; CS7 → row 14; CS8 → rows 15/17; CS9 → row 19; CS10 → row 20a);
   the 20-row test passes; H0 was red.
4. **AC-4** `_ship.agent.md` states **no classification predicate**, and **no `018.` literal**.
5. **AC-5** The close-verdict allowlist is exactly `CASCADE` / `SAFE_CLOSE`; `HALT` and
   `RECONCILE_FAIL` are non-close outcomes; unknown tokens (incl. `BLOCKED`), absent, empty,
   ambiguous or unreadable reports **HALT**. **No runtime `policy_id`, version or
   source-disagreement check exists.**
6. **AC-6** No lifecycle class is determined by location (§3.1); containment is preserved in
   full; archive placement applies only after declared `status: archived`.
7. **AC-7** The §3.3 exemption is triggered only by declared `status: archived`, applies only
   to the 11 enumerated IDs, requires all five gates, and lives **only in this plan** — not in
   `_ship.agent.md` and not in `workflow-policies.md`.
8. **AC-8** CS7/CS8 are pointer-only; they **preserve** the single-writer lock, the orphan
   scan and the `Scope note` (all three already present in the file), **add** the
   `record-consistent` conjunct (not previously present), and state no per-item predicate.
   Rows 14/15/17/18 pass.
9. **AC-9** a1 runs **before** pre-mode and is self-sufficient; all evaluation (S0–S3, S5) is
   read-only and its **only** write is S4's single `active -> done` transition; the S0–S5
   selector is total and disjoint. Row 16 passes.
10. **AC-10** Probe 25 is committed and PASSing **before `021-S` is claimed**; §9.2's seven
    checks pass at step 12a; every check is a read; S2 writes nothing to `docs/plans/`.
11. **AC-11** Exactly **3 implementation files**, **0 new production files**, **0 new module
    dependencies**.
12. **AC-12** The two sessions are distinct: S1 never performs a1 or the close.
13. **AC-13** S2's recovery is the **installed** `_ship.agent.md` crash-resumption protocol,
    executed unmodified (§9.1).
14. **AC-14** For `021-S`/`017-S` a non-`CASCADE` reported verdict **HALTs before any close
    mutation**; no `SAFE_CLOSE` fallback.
15. **AC-15** Every post-CASCADE failure performs the §10.3 **verified** restore — tree
    equivalence **and** record equivalence — then HALTs. The cascade is not invoked unless
    §10.3.2 step 1's clean trees, pre-cascade commit, `{pre_paths}` and Step 0(b) snapshot all
    exist.
16. **AC-16** Both `PA-*` records are present with `ActionRisk: destructive`,
    `ActionResult: approved`, the operator timestamp, and all applicable conditions.
17. **AC-17** The lifecycle order of §9 holds; the `--phase lifecycle` gate never runs after
    the archive.
18. **AC-18** P-018 engagement precedes **both** merges.
19. **AC-19** After claiming `021-S`, `022-F`'s live status is **re-read** and required to be
    **exactly `active`**; any other value **HALTs**. Ship performs no `queued -> active`
    transition on a feature.
20. **AC-20** **(P-002/P-004)** The **task** claim `022.001-T -> active` occurs **after**
    harness-architect has confirmed the red phase and applied `harness-ready` — `go vet ./...`
    exits 0 and `go test ./...` fails with expected markers. *(The **shipment** claim correctly
    precedes harness generation, per the installed Step 0.5 → Step 2 → Step 4.1 order.)*
21. **AC-21** **(P-011)** The post-merge closure branch exists **before any S2 mutation**,
    including `backlogit_resolve_checkpoint`, which runs at §9 step 11a on that branch. No S2
    mutation occurs on `main`. Resume-before-resolve semantics are unchanged.
22. **AC-22** §9 step **13a** commits a1's mutation on the **existing** closure branch,
    designates that commit `{pre_cascade_sha}`, verifies clean `.backlogit/queue/` +
    `.backlogit/archive/` trees and captures `{pre_paths}` **against that commit**, before
    pre-mode or the cascade runs. No second branch and no second worktree is created. §10.3's
    restore uses **exactly** this baseline.
23. **AC-23** The cascade result is **not committed until every postcheck passes**; a
    committed-cascade-awaiting-postchecks state is unreachable, and if observed is an
    out-of-contract HALT to the operator rather than a §10.3 restore.
24. **AC-24** **(J-4)** `workflow-policies.md`'s **P-010 `Ship MAY`** list carries the §3.2
    grant in wording that matches ADD-1, so Ship is never simultaneously authorized by its
    agent table and unauthorized by the policy registry. Row 20a passes. **P-010 gains no
    other change** and carries **no release-instance ID** — row 20b passes.
25. **AC-25** **(J-3)** **Both** `{pre_paths}` and `{post_paths}` are defined in §10.3.1 with
    an explicit capture command and capture moment; `{post_paths}` is captured
    **unconditionally on return from the cascade**; both use the **identical** enumeration
    command over the identical two roots. No variable is consumed by the destructive rollback
    without a definition.
26. **AC-26** **(J-6, J-11)** §9.1 states **no competing recovery algorithm**; it defers to the
    installed `_ship.agent.md` protocol including `consumer_id: "ship"`, operator selection and
    operator confirmation, and names **all three** never-prune allowlist elements. Its only
    additions are the expected-identity carry and the identity-mismatch HALT.
27. **AC-27** **(J-5)** The delegation target is the installed **`mode: safe-close` Step 0(c)**
    classification, reached through the existing `mode: safe-close` invocation; the verdict is
    **read back from the produced report**. **No plan text requires an API the installed skill
    does not expose**, and no Ship runtime code is added.
28. **AC-28** **(J-9)** Containment and orphan detection are specified via **`backlogit doctor`
    only**. `backlogit_query_sql` is explicitly excluded for containment, with the reason
    recorded.
29. **AC-29** **(J-8, J-15)** Evidence identity is verified against **plan-pinned** blob OIDs
    and a plan-pinned engine digest, cross-checked against the transcript, with divergence
    HALTing; and the classifier's own byte identity is verified against a pinned blob OID
    (§9.2 check 7). A same-PR transcript+script replacement cannot pass without an additional
    visible edit to this reviewed plan.
30. **AC-30** **(J-16)** §3.2 conditions 1–3 quantify over the **full live `parent_id`
    descendant graph at every depth**, and condition 3 additionally requires **set equality**
    between that graph ∪ `{feature}` and the manifest. Zero-descendant cases are positively
    verified.
31. **AC-31** **(J-7)** `PA-017-CASCADE` cites **no execution step of this plan**. §11.1 binds
    it explicitly to a future Stage-authored `017-S` closure plan and records that as a third
    independent ineligibility ground. No cross-reference in §11 resolves to a non-applicable
    action.
32. **AC-32** **(J-18)** §5 **S2** (`n > 1`) is a **HALT with no mutation**, consistent with
    P-015's whole-manifest disqualification and with §3.2's one-transition grant.

## 13. Gate state

```text
plan-review-attempt: 6
dispatch_mode: multi-agent
decision: FAIL
```

> **Attempt 6 returned FAIL (P0 = 1).** Per operator direction, **the plan was not patched in
> response.** The verdict, the deduplicated blocker set `K-1`…`K-7` and the per-persona counts
> are recorded in
> `docs/reviews/2026-09-13-true-lineage-attempt-6-verdict.md`.
> **No attempt 7 is authorized.** `021-S` remains `queued` and not claimable.

### 13.1 Current state

| Field | Value |
|---|---|
| **Plan `status:`** | `planned` |
| **Revision** | **10** — reauthored, not patched |
| **True-lineage attempt** | **6** (the counter continues across the revision-9 → 10 reauthoring; it is **not** reset) |
| **Gate decision** | **FAIL** — P0 = 1. See `docs/reviews/2026-09-13-true-lineage-attempt-6-verdict.md` |
| **Harvest-ready** | **NO** |
| **`021-S`** | `queued` — **not claimable**; the gate did not return P0 = 0, P1 = 0 |
| **`017-S`** | `queued`, dependency `[021-S]` — ineligible until `021-S` ships, **and** separately ineligible until an `017-S` closure plan exists (§11.1) |
| **`PA-021-CASCADE`** | recorded, **unexercised**; gated on this state, which is now FAIL |
| **`PA-017-CASCADE`** | recorded, **unexercised**, **held in escrow** (§11.1) |
| **Authorization** | Explicit operator authorization to **reauthor** the artifact and run **exactly one** further gate as true-lineage attempt 6. **That authorization is now spent. No attempt 7 is authorized, and no further patch cycle was performed.** |

**Lineage (compact — narrative deliberately not restated inline):**

| Attempt | Revision | Verdict |
|---|---|---|
| 1 | 4 | FAIL |
| 2 | 5 | FAIL |
| 3 | 6 | FAIL (re-entry budget exhausted, escalated) |
| 4 | 7 | FAIL (`H-1`…`H-9`) |
| 5 | 9 | FAIL (`J-1`…`J-18`) |
| **6** | **10** | **FAIL** (`K-1` P0 + `K-2`…`K-7` P1) |

> The counter reflects **plan-review lineage, which follows the plan, not the filename or the
> revision number.** Revisions 1–3 and 8 carried no gate of their own. A counter reset was
> corrected at attempt 5 and is **not** re-reset by this reauthoring.

### 13.2 Audit pointers (lineage lives in Git, not in this file)

This plan carries **no inline review history**. Everything below is reachable and complete.

| What | Where |
|---|---|
| Superseded revisions 1–8, attempts 1–4, withdrawn wording | `docs/archive/plans/2026-09-12-intercom-go-ship-feature-completion-foundation-plan.md` |
| Attempt-4 findings `H-1`…`H-9` and dispositions | commit `4abea02`, and revision 9 of this file in Git history |
| Attempt-5 findings `J-1`…`J-18`, full text and per-persona counts | commit `e9019d4` (§13.4 of revision 9) |
| Attempt-5 session record | `docs/memory/2026-09-13-stage-true-lineage-attempt-5-fail.md` |
| Attempt-6 reauthoring session record | `docs/memory/2026-09-13-stage-true-lineage-attempt-6-reauthor.md` |
| Recovery / torn-state halt record | `docs/memory/2026-09-13-stage-recovery-torn-state-halt.md`, commit `3e0cf8f` |
| Checkpoints | `.backlogit/checkpoints/checkpoint-20260913-231952.json` (attempt-5 terminal state) and its siblings |
| Change lineage for this file | `git log --follow -- docs/plans/2026-09-13-intercom-go-ship-feature-completion-decided-plan.md`; PR #54 |

**Reauthoring does not discharge findings.** `H-1`…`H-9` were disposed at attempt 5.
`J-1`…`J-18` are disposed in §13.3 below, each pointing at the section that resolves it.
**A finding is closed by a landing site, never by the disappearance of the text that carried
it.**

### 13.3 Attempt-5 findings — dispositions at revision 10

**P1 (`J-1`…`J-9`) — all REMEDIATED:**

| Ref | Landing site |
|---|---|
| **J-1** incomplete rev sweep inside the lineage section | **Structurally eliminated.** §13 carries no inline lineage narrative and no per-revision prose; the revision appears in frontmatter and §13.1 only. Verified by §13.4 measurement. |
| **J-2** occurrence claims scoped to the plan while literals survived in 4 records | §13.4 — every measurement is declared over **{plan} ∪ {4 backlog records}**, with the exact reproducible command. The records are reauthored, so no defect literal survives to be excluded. |
| **J-3** `{post_paths}` consumed but never defined | **§10.3.1** — both `{pre_paths}` and `{post_paths}` defined with command, roots and capture moment; `{post_paths}` captured unconditionally on cascade return. AC-25. |
| **J-4** P-010 grants task transitions only; conflicts with the Ship role table | **§2 + CS10 + §4.1 + AC-24** — surgical third-file scope expansion, exactly one bullet of P-010's `Ship MAY` list, pinned wording matching ADD-1, asserted by test row 20a/20b. |
| **J-5** delegation targets a classify API the skill does not expose | **§3.4.1** — delegation retargeted to the installed **`mode: safe-close` Step 0(c)**, reached via the existing invocation; verdict read back from the produced report. No new API, no Ship runtime code. AC-27. |
| **J-6** plan recovery algorithm conflicts with installed `_ship.agent.md` | **§9.1** — the competing algorithm is **withdrawn in full**; the installed protocol governs verbatim (`consumer_id: "ship"`, operator selection, operator confirmation). AC-26. |
| **J-7** `PA-017-CASCADE` cites a non-applicable action | **§11.1** — the §9 step 13a reference is withdrawn; the approval is bound to a future Stage-authored `017-S` closure plan and recorded as a third ineligibility ground. AC-31. |
| **J-8** evidence ancestry defeatable by same-PR replacement | **§8.1 + §8.1.1 + §9.2 checks 3/4/5** — blob OIDs and the engine digest are **pinned in this reviewed plan**, cross-checked against the transcript, divergence HALTs. AC-29. |
| **J-9** `query_sql` cannot detect duplicate paths | **§5 executable-surface table** — containment and orphans via **`backlogit doctor`** only; `query_sql` explicitly excluded with the measured reason. AC-28. |

**P2 (`J-10`…`J-18`) — all addressed:**

| Ref | Landing site |
|---|---|
| **J-10** protected-set claim in the CS2→CS3 gap | **CS9** (new anchored site) + **§4**'s switch to anchor-addressed sites, which removes ordinal gaps as a class. Test row 19. |
| **J-11** never-prune allowlist lost 1 of 3 elements | **§9.1** names all three explicitly. AC-26. |
| **J-12** AC-3 claimed test coverage for sites without needles | **AC-3** now enumerates the needle for **every** CS1–CS10 site; §6.1 pins all 20 rows. |
| **J-13** only 9 of 18 rows pinned | **§6.1** pins **all 20 rows**, both polarities, with H0 counts. |
| **J-14** backlog records append-only | All four records **reauthored to clean current state** in the same commit as this revision. |
| **J-15** classifier byte-identity unverified | **§3.4.2** + **§9.2 check 7**, pinned OID in §8.1. AC-29. |
| **J-16** condition-2 manifest vs full-graph ambiguity | **§3.2.1** — full live descendant graph at every depth, plus set equality with the manifest. AC-30. |
| **J-17** uncited prior art | §9 (P-018 trio), §14 (`…sizing-is-wit-gated`), §9 note (`…shipment-status-constraints`), and §13.5 (`…adversarial-review-empirical-verification-and-scope-check`, `…verify-go-stdlib-internals-claims-against-goroot-source`). |
| **J-18** `a1 n>1` halt narrower than P-015 | **§5 S2** is now an explicit **HALT, no mutation**, with the P-015 alignment stated. AC-32. |

### 13.4 Measurement — two scopes, declared separately

`J-2` was caused by stating one count and implying a wider scope. There are **two distinct
scopes** here and a count is meaningless without naming which one it is.

* **Scope I — the implementation surface**: `_ship.agent.md` ∪ `workflow-policies.md`.
  This is what the 20-row test asserts over. **All Scope-I counts live in §6.1**, pinned per
  row, and are not restated here.
* **Scope A — the artifact set**: `{this plan}` ∪ `{021-S, 017-S, 022-F, 022.001-T}` — five
  files. This is what the reauthoring had to clean.

**A defect literal may legitimately appear in Scope A while being required to be absent from
Scope I.** `TASK_ONLY_FINALIZE` is the canonical example: this plan must *name* it to pin
row 4's negative needle, and it is measured **4× in Scope A** for exactly that reason (three
pin/prescription sites plus this sentence), while its required Scope-I count is **0**.
Conflating those two numbers is the `J-2` error itself.

#### Scope-A measurement (reproduce exactly)

Needles are written as **regexes with a bracketed character class** so that this table does
**not** match itself. `re[v]ision: 9` matches the string it describes; the cell does not.

```powershell
$files = @(
  'docs/plans/2026-09-13-intercom-go-ship-feature-completion-decided-plan.md',
  '.backlogit/queue/021-S.md', '.backlogit/queue/017-S.md',
  '.backlogit/queue/022-F.md', '.backlogit/queue/022.001-T.md')
foreach ($rx in $patterns) {
  ($files | ForEach-Object { Get-Content $_ } | Where-Object { $_ -match $rx }).Count
}
```

| Regex | Scope-A count | What a non-zero would mean |
|---|---|---|
| `re[v]ision: 9` | **0** | a stale frontmatter revision survived |
| `re[v] 8` | **0** | a stale short-form revision survived (`J-1`) |
| `re[v] 9` | **0** | same |
| `AUTHORITATIVE GATE [S]TATE` | **0** | the dangling `H-1` heading survived in plan **or** records |
| `SUPERSEDES EVERY [B]LOCK` | **0** | append-only stacking survived in a record (`J-14`) |
| `2 [F]ILES` | **0** | the superseded 2-file surface survived a record (`J-2`, `J-4`) |
| `6 [S]ITES` | **0** | the superseded clause count survived a record (`J-2`) |
| `MUST NOT enum[e]rate` | **0** | withdrawn `H-6` wording survived |
| `without touching this [f]ile again` | **0** | withdrawn `H-5` wording survived |

**Bounded, not zero — and deliberately so:**

| Regex | Scope-A count | Every occurrence is |
|---|---|---|
| `re[v]ision 9` | **3** | one supersession statement (§ header banner) + two §13.2 audit-pointer rows. **No normative use.** |
| `backlogit_query[_]sql` | **2** | §5's prohibition and AC-28's restatement of it. **Never a prescription** (`J-9`). |
| `{post[_]paths}` | **≥ 2** | §10.3.1's definition and its §10.3/§10.4 consumption. Definition precedes consumption (`J-3`). |
| `re[v]ision 10` | **7** | three plan sites (supersession banner, §4.2 re-measurement note, §13.3 heading) + **one governing-plan pointer in each of the four records**. Every record names the current revision; none names a superseded one. |

**Rule for any future claim over these artifacts: name the scope, give the regex, give the
number.** A bare occurrence claim is not evidence.

### 13.5 Prior art relied upon

* `docs/compound/2026-09-08-adversarial-review-empirical-verification-and-scope-check.md` —
  the discipline applied throughout §4.2, §5, §6.1 and §13.4: **measure, then claim**, and
  declare the scope of every count. The `H-3`/`J-2`/`J-9` failures were all instances of
  inference presented as measurement.
* `docs/compound/2026-09-08-verify-go-stdlib-internals-claims-against-goroot-source.md` —
  verify a tool/API claim against the installed source, not against memory. Directly applied
  to `J-5` (skill mode inventory) and `J-9` (index schema).
* `docs/compound/2026-05-07-backlogit-shipment-status-constraints.md` — shipment status enum,
  which is why CS2 deletes the `--status shipped` prescription.
* `docs/compound/2026-09-04-backlogit-sizing-is-wit-gated.md` — §14.
* The P-018 trio cited in §9.

## 14. Sizing

**Size `M` · Complexity `high`** (complexity is enum-validated **prose** — this workspace's
`header-def.yaml` defines no `complexity` field on the task type; see
`docs/compound/2026-09-04-backlogit-sizing-is-wit-gated.md`. Do **not** invoke
`backlogit update --complexity`; it fails with exit 1).

| Work item | Est. |
|---|---|
| ADD-1 Role Boundary grant | 8 |
| ADD-2 Step 6.1(a1) incl. S0–S5 | 18 |
| CS1 reload + carve-out | 4 |
| CS2 close-sequence pointer | 4 |
| CS3–CS6 delegation block | 12 |
| CS7 intake pointer | 4 |
| CS8 pre-archive pointer | 4 |
| CS9 protected-set scoping | 3 |
| CS10 P-010 grant bullet | 3 |
| Manifest/topology verification | 4 |
| 20-row Go test | 32 |
| H0/H1 cycle + gates | 10 |
| **Total** | **106 min** (margin **14** under the 2-hour rule) |

## 15. Constitution

| Principle | Assessment |
|---|---|
| **I. Safety-First Go** | One table-driven contract test. No `unsafe`, no goroutines, no I/O beyond reading two files; errors via `t.Errorf` |
| **II. Test-First (NON-NEGOTIABLE)** | H0 red → H1 green; the test lands before any `_ship.agent.md` or `workflow-policies.md` edit. **Deviation D1** |
| **III. Workspace Isolation** | Probe 25 ran in a disposable, gitignored, workspace-contained directory; `FIXTURE_ISOLATED_FROM_LIVE=True`, `LIVE_BACKLOG_UNMUTATED=True` measured |
| **IV. CLI Containment (NON-NEGOTIABLE)** | Every probe call `--cwd $ws`; engine resolved via the registered command, never an absolute path |
| **V. Structured Observability** | a1 emits `A1_NOT_APPLICABLE`, `A1_ALREADY_DONE`, `A1_MULTIPLE_FEATURE_MEMBERS`; Probe 25 emits one `KEY=VALUE` per criterion |
| **VI. Single Responsibility** | **Dependency discipline**: zero module dependencies added; stdlib-only assertions; `testify` absent from `go.mod`; reuses `repoRoot(t)`. The change **removes** a dependency direction — Ship stops depending on a local copy of the classification. **Deviation D2** |
| **VII. Destructive Approval (NON-NEGOTIABLE)** | The cascade is destructive and **live**. Gated by P-015, pre-mode `PROCEED`, the §3.5 no-fallback HALT, **and** the two §11 approved-action records with operator authorization |
| **VIII. Explicit Safety Modes** | **Deviation D3** |
| **IX. Git-Friendly Persistence** | Line-oriented Markdown/Go; the backlog mutation is committed on the closure branch, never on `main` |
| **X. Agent Context Efficiency** | Net **removal** of re-derived classification prose from Ship's hot path; this plan itself is reauthored rather than appended, removing ~180 lines of inline review history |
| **XI. Merge History (NON-NEGOTIABLE)** | Rollback is `git revert`; no rewrite, no force-push, no amend. **The reauthoring adds commits; it never amends or rewrites `3e0cf8f`, `3423581` or `e9019d4`** |

**D1 — 20 scenarios vs "fewer than 4".** These are **rows in one table-driven function** over
**two** files, sharing one fixture and ≤4 helpers; each is a substring or index assertion, not
an independent scenario. Collapsing them reduces falsifiability. *Assertion density, not scope.*

**D2 — Width-Isolation: one task spans Markdown and Go.** `_ship.agent.md` and
`workflow-policies.md` are **not documentation** — they are the agent's **executable
instruction surface** and its **policy registry**, read and acted on at run time. The Go test
is a table of substring assertions over those two files, derived mechanically from §6.1. They
are **the contract and its compiler-check**. Splitting them produces a task that cannot be
verified and a task that cannot compile green, breaking the mandatory H0-red → H1-green
ordering, which requires both halves in one atomic change.
Precedent: `tests/integration/build_script_test.go` asserts over build text the same way.

**D3 — Principle VIII: global-closure blast radius without a mode directive.** Accepted with
compensating controls: the change *removes* authority duplication **while also adding a narrow
permanent authority and new cascade reachability for `017-S`** — both stated plainly; five
negative widening guards (rows 7, 8, 9, 10, 20b plus the `018.` guards) fail if anything new is
authorized; the grant is bounded by five conjunctive conditions, one position, one scope and
one transition; and A closes by the **pre-existing** path.
**Rejected simpler alternative: scope the grant to `021-S`/`017-S` by ID inside
`_ship.agent.md`.** Rejected because (a) it re-introduces release-instance IDs into the global
contract, the exact coupling §3.3 and §3.5 remove; (b) it relocates rather than reduces risk —
each new shipment would need another edit to the same global file; (c) it cannot be verified
negatively, since the widening guards assert the files contain **no** release-specific logic.
Its safety intent is preserved instead by §3.5 and the two §11 approved-action records.
**Is VIII satisfied via freeze-scope rather than deviated?** **No** — the grant is general and
permanent from merge; the approved-action records bound the **cascade invocations**, not the
grant. It remains a **documented deviation with compensating controls**.

## 16. Out of scope

The deferred B/C platform plans; `shipment-reconcile/SKILL.md` edits; Orchestrator repair; the
pre-existing `SAFE_CLOSE`/exit-9 tool conflict; the P-015 detect-after-mutate **ordering**,
which is an inherited residual recorded as a follow-up against P-015 and the skill; and the
**`017-S` closure plan** (§11.1), which is future Stage work triggered by §9 step 19.

**Explicitly out of scope for the P-010 expansion**: any P-010 clause other than the single
`Ship MAY` bullet named by CS10; any change to P-010's `Ship MUST NOT`, `Stage MAY` or
`Stage MUST NOT` lists; and any other policy in `workflow-policies.md`.

## 17. Full history

Superseded revisions 1–9, plan-review attempts **1–5**, and all withdrawn language are
preserved in Git history, in PR #54, and in
`docs/archive/plans/2026-09-12-intercom-go-ship-feature-completion-foundation-plan.md`.
**No content was deleted from history; this file was reauthored, not rewritten in place.**
Where the archive and this plan disagree, **this plan governs**.

**Findings remain live across reauthoring.** Supersession, compaction and reauthoring all
transfer the contract *and* its outstanding findings. `H-1`…`H-9` were disposed at attempt 5;
`J-1`…`J-18` are disposed in §13.3 with landing sites. **Do not read "the history is
non-governing" as "the history's findings are closed."**




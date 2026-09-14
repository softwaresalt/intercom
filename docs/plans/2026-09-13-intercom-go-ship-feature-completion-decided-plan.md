---
title: "Decided Plan — Ship covering-feature completion and close-path delegation (022-F)"
date: 2026-09-13
status: planned
agent: Stage
revision: 11
feature: 022-F
task: 022.001-T
shipment: 021-S
releases: 017-S
deliberation: docs/decisions/2026-09-13-intercom-go-a-only-simplification-decision.md
archived_full_plan: docs/archive/plans/2026-09-12-intercom-go-ship-feature-completion-foundation-plan.md
requires_plan_hardening: "no"
evidence: docs/plans/evidence/2026-09-12-task-only-shipment-finalization/probe25-root-included-cascade-fixture.txt
---

# Decided Plan — Ship covering-feature completion and close-path delegation (022-F)

> **This is the SOLE GOVERNING ARTIFACT for `021-S`.** It states the current, executable
> contract and nothing else. It carries **no inline review narrative**: lineage, findings and
> superseded wording are reachable through the audit pointers in §13.2, which resolve to Git
> history, PR #54, the archived plan, `docs/memory/` and `.backlogit/checkpoints/`.
>
> **Reauthored, not patched.** Revision 11 replaces revision 10 under explicit operator
> direction to reauthor rather than append. Lineage is preserved by Git commits and PR
> history; it is deliberately not restated inline.
>
> **Revision 11 adopts OPTION A (prevention-first).** The operator selected, on
> 2026-09-13, adding a **read-only pre-mutation classification boundary** to the installed
> `shipment-reconcile` skill over the alternative of rewriting the safety model as
> detect-and-restore. The implementation surface is therefore **four files**, not three.
> See §2, §3.4.3 and §4 (**CS11–CS14**).

## 1. Objective

Grant Ship one narrow authority — complete a covering feature `active -> done` — and replace
Ship's **duplicated copy** of the close-path classification with **delegation** to the
installed policy and skill.

**Why it is required**: pre-mode runs `expected_status: done` over **every** manifest item, so
a fully-covered-root manifest cannot pass while its covering feature is still `active`. Ship's
Role Boundary today permits only task transitions. Independently, Ship's duplicated copy has
**already drifted unsafe** — it says *"every one of its children"* where the installed policy
requires descendants **at every depth**.

## 2. Surface — exactly 4 implementation files

| File | Change |
|---|---|
| `.github/agents/_ship.agent.md` | 9 clause sites (**CS1–CS9**, all mandatory) + 2 additive (**ADD-1**, **ADD-2**) |
| `.github/policies/workflow-policies.md` | 1 clause site (**CS10**) — the P-010 `Ship MAY` grant |
| `.github/skills/shipment-reconcile/SKILL.md` | 4 clause sites (**CS11–CS14**) — the read-only pre-mutation classification boundary (§3.4.3) |
| `tests/integration/ship_feature_completion_contract_test.go` | **new** — 24-row table-driven contract test over the three instruction-surface files above |

**0 new production files.** Stage-authored planning artifacts (this plan, the Probe-25
evidence, `docs/memory/`, `docs/decisions/`, `.backlogit/` records) are committed separately
and are **not** counted against this 4-file surface.

**Why the third file is in scope.** `_ship.agent.md`'s Role Boundary table and P-010's
`Ship MAY` list are **two halves of one authority contract**; P-010 is the policy registry that
the agent table is required to conform to. Granting the covering-feature transition in only one
half lands Ship simultaneously **authorized** by its agent table and **unauthorized** by the
policy registry. The expansion is **surgical**: exactly one bullet in P-010's `Ship MAY` list
(**CS10**). No other policy, no other clause, no other file.

**Why the fourth file is in scope (OPTION A — the `K-1` resolution).** This plan's central
safety guarantee is *a non-`CASCADE` close path must be refusable **before** any close
mutation*. Measured against the installed skill, that guarantee was **unreachable**: the
close-path classification lives **inside** `mode: safe-close` Step 0(c), and on a non-`CASCADE`
selection the skill *"continue[s] to step 1 below"* — entering the mutating safe-close steps
1–10 directly (`SKILL.md` L509–512) — while the verdict is recorded only in the **cascade**
report, at Cascade Close Sub-Procedure step 5, i.e. **after** the destructive
`backlogit_ship_shipment` call (L821). A caller therefore cannot observe a verdict before a
mutation has already occurred, and `SAFE_CLOSE` is never an observable value at all.

A plan cannot assert a halt its tooling cannot perform. Two repairs were possible: **(A)** give
the skill a read-only pre-mutation classification boundary, or **(B)** rewrite §3.5 / §10.1 /
AC-14 and both `PA-*` condition-1 clauses from *prevent* to *detect-and-restore*. **The
operator selected (A)** on 2026-09-13. (B) is rejected in this plan because it converts a
preventable destructive execution into an expected one and makes the restore path — itself the
riskiest machinery here — load-bearing for the common case.

The expansion is **surgical and additive**: one new **read-only** mode, one new Behavioral
Constraint, one new optional `safe-close` input, and one scoping edit to the existing Step 0
dispatch. **The four pre-existing public modes (`pre`, `post`, `safe-close`,
`detect-mixed-role`) keep their current inputs, outputs and behaviour when invoked exactly as
they are today**, asserted negatively by test rows 23b and 24b.

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

#### 3.4.1 The delegation target is the read-only `mode: classify-close-path` boundary

**Measured against the installed skill (Scope T, §13.6).** `.github/skills/shipment-reconcile/SKILL.md`
today exposes exactly four modes — `pre`, `post`, `safe-close`, `detect-mixed-role` — and **no
public classification entry point**. The close-path classification is internal to
`mode: safe-close` **Step 0(c)**, and on a non-`CASCADE` selection the skill continues into the
**mutating** safe-close steps 1–10. **CS11–CS14 add the missing boundary** (§3.4.3): a new
**read-only** public mode, `classify-close-path`, that runs the *existing* Step 0(a)/(b)/(c)
logic and **stops** at the dispatch, returning the verdict instead of acting on it.

**The executable delegation is therefore:**

1. Ship invokes the skill with **`mode: classify-close-path`** and `shipment_id`. This call
   **mutates nothing** and is safe to issue unconditionally.
2. **Step 0(c) selects the close path internally.** Ship supplies no verdict, no predicate and
   no hint, and never pre-computes the selection.
3. Ship reads the returned **`CLOSE_PATH_VERDICT`** — one of `CASCADE`, `SAFE_CLOSE`, `BLOCK` —
   together with its `VERDICT_REASON`, `VERDICT_EVIDENCE` and `CLASSIFICATION_BINDING`.
4. Ship **validates that token against a fixed allowlist** and does nothing else with it.
   **On any value other than `CASCADE` on these routes, Ship HALTs — and at that point no
   close call has been made at all**, so there is nothing to unwind (§3.5, §10.1).
5. Only on `CASCADE` does Ship invoke **`mode: safe-close`**, passing the
   `CLASSIFICATION_BINDING` back. Safe-close **revalidates** the binding against a freshly
   computed snapshot before any mutation (§3.4.3 (E)), so the verdict Ship acted on is the
   verdict the mutation executes under.

**`_ship.agent.md` MAY contain an output-validation allowlist of the installed CLOSE VERDICT
tokens — `CASCADE`, `SAFE_CLOSE` and `BLOCK` — and nothing more.** It **MUST NOT** contain any
classification predicate: no root/`parent_id` test, no coverage test, no per-member rule, no
fallback rule. Comparing a returned token against a list is validation; deciding which token
applies is classification, and stays inside Step 0(c).

* **`HALT` and `RECONCILE_FAIL` are non-close procedural outcomes**, not close verdicts. They
  authorize no close path.
* **`BLOCK` is the fail-closed verdict** for an unsupported, mixed, ambiguous, torn, missing or
  incompletely-enumerated manifest. It is **not** a close path and authorizes no mutation.
* **The skill's `recommendation` field is a separate axis** (`PROCEED`, `CLOSED`, `HALT — …`)
  and is **never** read as a close verdict.

**Ship HALTs, no mutation, on any of:** a returned token outside the close-verdict allowlist;
`BLOCK`; an absent, empty, unparseable or ambiguous classification result; a skill that is not
installed or not readable; or a non-close procedural outcome.

The allowlist is bound to the **post-`H1`** skill, whose identity is established by §9.2
check 7. Widening it requires its own review; this plan authorizes **no new close path** —
`BLOCK` is a **refusal**, not a path.

#### 3.4.2 Classifier integrity — established by contract, not by a pre-pinned blob

Revision 10 verified the classifier by pinning `shipment-reconcile/SKILL.md`'s blob OID,
which was sound **only while that file sat outside the implementation surface**. Under
Option A the file **is** modified by `H1`, so a pre-merge blob OID is unpinnable by
construction and that check is **withdrawn**.

Identity is instead established by three independent, reproducible facts, all verified at
§9.2 check 7 against the **merge commit** confirmed at §9 step 6:

```text
1. contract identity  := the 24-row test at {merge_sha} passes, including rows 21-24
                         (CS11-CS14) over .github/skills/shipment-reconcile/SKILL.md
2. surface identity   := `git diff --name-only {merge_base}..{merge_sha}` intersected with
                         the tracked tree equals exactly the four files of section 2
3. working identity   := `git diff --quiet {merge_sha} -- .github/skills/shipment-reconcile/SKILL.md`
                         (the installed file is byte-identical to the reviewed merge commit)
```

A classifier that changed after the reviewed merge fails fact 3; a classifier that changed
*within* the merge in an unreviewed way fails fact 1 or fact 2. **Any of the three failing
⇒ HALT**, no close call, return to Stage.

#### 3.4.3 The read-only pre-mutation classification boundary (CS11–CS14, normative contract)

This is the **fourth file's** entire contract. It is stated as an executable, test-first
specification: every clause below is either pinned by a test row in §6.1 or measured in
§13.6.

**(A) New public mode — read-only.** `mode: classify-close-path` joins the mode enum.
Required input `shipment_id`. It accepts **no** `merge_commit_sha` and **no**
`expected_status`.

**(B) Classification runs before any mutation, and the mode cannot reach one.** The mode
executes the **existing** Safe-Close Step 0(a) manifest load, Step 0(b) snapshot and
Step 0(c) classification **verbatim and unmodified**, then **stops at the dispatch** — it
never takes the `CASCADE selected →` or `SAFE_CLOSE selected →` branch. It **performs no
archive, no status transition and no record mutation**, acquires **no** single-writer lock
(there is nothing to serialise), and writes only its own additive report, exactly as
`mode: detect-mixed-role` already does.

**(C) Machine-consumable result.** The mode returns, and records at
`.backlogit/reconcile/{shipment_id}-classify-close-path-{timestamp}.md`, exactly these
fields:

| Field | Values | Meaning |
|---|---|---|
| `CLOSE_PATH_VERDICT` | `CASCADE` \| `SAFE_CLOSE` \| `BLOCK` | The selected close path, or a refusal |
| `VERDICT_REASON` | one stable token | e.g. `FULLY_COVERED_ROOT`, `PARTIAL_FEATURE`, `MANIFEST_MEMBER_NOT_DESCENDANT`, `SNAPSHOT_AMBIGUOUS`, `SNAPSHOT_MISSING`, `ENUMERATION_INCOMPLETE`, `MIXED_QUALIFICATION` |
| `VERDICT_EVIDENCE` | the per-member evidence set | For every manifest member: its ID, `artifact_type`, declared `status`, `parent_id`, resolved location, and which precondition it satisfied or failed |
| `CLASSIFICATION_BINDING` | a digest | §3.4.3 (D) |
| `CLASSIFIED_AT` | timestamp | Capture moment of the snapshot |

**(D) Canonical snapshot and binding — the TOCTOU control.** `CLASSIFICATION_BINDING` is a
SHA-256 over a **canonical serialisation** of exactly the state the classification consumed:
the manifest item IDs **sorted**; the shipment record's own declared `status` and its
`dependencies` list, sorted; and, for every member of the combined pre-close snapshot that
Step 0(b)/(c) already builds (manifest tasks, qualifying feature members, and qualifying
features' validated linked deliberations), the tuple `(id, artifact_type, declared status,
parent_id, resolved location)` — each tuple rendered on one line, the lines sorted, joined
with `\n`. **Nothing else enters the digest**, so the binding is reproducible by any party
reading the same tree.

**(E) Mutation revalidates or consumes the binding.** `mode: safe-close` gains one
**optional** input, `classification_binding`. When supplied, safe-close recomputes the
binding from its own freshly loaded Step 0(a)/(b)/(c) snapshot **before any mutation** and:

* **binding matches and the bound verdict is `CASCADE`** ⇒ proceed to the Cascade Close
  Sub-Procedure as today;
* **binding matches and the bound verdict is not `CASCADE`** ⇒ **HALT before any close
  mutation**, returning `RECONCILE_FAIL_CLASSIFICATION_REFUSED`. Safe-close does **not** fall
  through into its mutating steps 1–10 under a bound non-`CASCADE` verdict;
* **binding does not match** ⇒ **HALT before any close mutation**, returning
  `RECONCILE_FAIL_CLASSIFICATION_DRIFT`. The tree moved between classification and mutation;
  the caller must reclassify.

**(F) Backward compatibility is total.** When `classification_binding` is **absent**,
`mode: safe-close` behaves **exactly as it does today**, including its existing default
fall-through to steps 1–10. The three pre-existing close-path behaviours are unchanged for
every existing caller; Option A adds a **stricter** path, it does not narrow the existing
one. This is asserted negatively by row 23b.

**(G) Independent invocability.** `mode: classify-close-path` is invocable by any caller —
Ship, an operator, or a future gate — **without entering the mutating path**, exactly as
`mode: detect-mixed-role` is today. It has no prerequisite mode, no lock and no ordering
constraint.

**(H) Fail closed.** Every unsupported, mixed, ambiguous, torn, missing or incompletely
enumerated case returns **`BLOCK`** with a reason token, **never** a silent `SAFE_CLOSE`.
This is the one behavioural difference from the internal Step 0(c) default, and it exists
because a *refusal* and a *chosen safe path* must not share a token when a caller is
expected to halt on one of them. Inside `mode: safe-close` with no binding supplied, the
legacy default is preserved unchanged (F).

**(I) Test-first obligations (`H0` red before `H1`).** The contract above is falsifiable by
rows **21–24** of §6.1 and by the three `AC-33`…`AC-38` criteria:

* **no mutation on a non-`CASCADE`/`BLOCK` result** — rows 22b and 23a pin the
  no-mutation clause and the bound-refusal HALT; `AC-35`;
* **fully-covered-root `CASCADE`** — row 21b pins the `CASCADE` emission path against the
  unchanged Step 0(c) predicate; `AC-34`;
* **safe-close case** — row 23b pins the preserved legacy fall-through; `AC-36`.

### 3.5 Release-instance execution contract (this plan only)

`_ship.agent.md` carries **no shipment IDs and no expected-verdict rule** — only the general
delegation above. The expectation below is a **precondition of this plan and its records**:

> For **`021-S`** and **`017-S`**, `mode: classify-close-path` is **expected to return
> `CLOSE_PATH_VERDICT: CASCADE`**, after which `mode: safe-close` is invoked **with the
> returned binding** and enters the **pre-existing** Cascade Close Sub-Procedure.
> **A returned `SAFE_CLOSE`, `BLOCK`, an unknown token, an unavailable or unreadable skill,
> an empty or ambiguous result, `HALT`, or `RECONCILE_FAIL` ⇒ HALT AT THE CLASSIFICATION
> BOUNDARY** — **before any close call is made** — and return to Stage.
> **There is no `SAFE_CLOSE` fallback on these routes.**

**This halt is now reachable, and that is the whole point of Option A.** It fires at §9
step 14a, after a **read-only** call. At that moment the skill has archived nothing,
transitioned nothing and taken no lock, so §10.1's *"no CLOSE mutation has occurred"* is a
verified property of the call sequence rather than an assertion about unreachable behaviour.
A second, independent guard covers the case where the tree moves between classification and
close: §3.4.3 (E)'s binding revalidation HALTs inside safe-close, still before any mutation.

A `SAFE_CLOSE` or `BLOCK` here means the live manifest is not the shape this plan asserts — a
planning error only Stage can reconcile.

### 3.6 Authority-escalation rule

A context reload MUST NOT widen the authority envelope of the session performing it. The
session that merges this change **may not use the authority it introduces**; it checkpoints
and ends.

## 4. Clause inventory — 14 sites (all mandatory) + 2 additive

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
| **CS11** | `shipment-reconcile/SKILL.md` | ``\| `mode` \| yes \| `pre` \| `post` \| `safe-close` \| `detect-mixed-role` \|`` | Add **`classify-close-path`** to the mode enum; add the optional `classification_binding` input row for `safe-close` (§3.4.3 A, E) |
| **CS12** | `shipment-reconcile/SKILL.md` | ``### Safe-Close Mode`` *(insertion point — the new section is inserted **immediately before** this heading)* | **NEW section** `### Classify-Close-Path Mode` — the read-only pre-mutation boundary (§3.4.3 B–D, G, H) |
| **CS13** | `shipment-reconcile/SKILL.md` | ``* **SAFE_CLOSE selected** (default, including any classifier error,`` | **Scope the dispatch.** Insert the binding revalidation ahead of it and make the fall-through conditional on **no bound verdict** (§3.4.3 E, F). The unbound behaviour is preserved verbatim |
| **CS14** | `shipment-reconcile/SKILL.md` | ``* **`mode: detect-mixed-role` is strictly READ-ONLY.**`` | **NEW** Behavioral Constraint bullet, parallel to this one: `mode: classify-close-path` is strictly read-only and never reaches a close path |

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
* **ADD-1** (Role Boundary row, `_ship.agent.md`) → *"Claim shipments, move tasks to
  active/done, **complete one covering feature `active -> done` when every live descendant at
  every depth is done and all five conditions below are conjunctive and fail-closed**, close
  shipments, archive completed items"*, immediately followed, in the same row, by the five
  conditions rendered so that the quantifier and the set-equality conjunct are both literal:
  *"(1) every manifest descendant is complete; (2) **no live descendant remains at every
  depth**; (3) topology is intact and **the descendant graph union the feature is set-equal to
  the manifest**; (4) containment holds; (5) the feature's live status is exactly `active`."*
* **CS10** → *"Claim shipments, move tasks to active/done, **complete one covering feature
  `active -> done` when every live descendant at every depth is done and all five conditions
  below are conjunctive and fail-closed**, close shipments, archive completed items"*
* **CS11** → the mode cell becomes *"``pre`` \| ``post`` \| ``safe-close`` \|
  ``classify-close-path`` \| ``detect-mixed-role``"*, plus a new input row *"`classification_binding`
  | no | safe-close only | digest returned by `mode: classify-close-path`; when present,
  safe-close revalidates it **before any mutation** and halts on drift or on a bound
  non-`CASCADE` verdict"*.
* **CS12** → the new section opens *"### Classify-Close-Path Mode (READ-ONLY, pre-mutation
  boundary)"* and states *"This mode **performs no archive, no status transition and no record
  mutation**, acquires no single-writer lock, and **stops at the Step 0 dispatch** — it never
  enters the cascade path and never continues into safe-close steps 1–10."* It emits
  ``CLOSE_PATH_VERDICT: CASCADE``, ``CLOSE_PATH_VERDICT: SAFE_CLOSE`` or
  ``CLOSE_PATH_VERDICT: BLOCK`` with ``VERDICT_REASON``, ``VERDICT_EVIDENCE`` and
  ``CLASSIFICATION_BINDING`` (§3.4.3 C/D), and *"every unsupported, mixed, ambiguous, torn,
  missing or incompletely enumerated case returns `CLOSE_PATH_VERDICT: BLOCK`"*.
* **CS13** → immediately before the preserved dispatch bullets, insert *"**Bound-classification
  revalidation (before any mutation).** When `classification_binding` is supplied, recompute the
  binding from the Step 0(a)/(b)/(c) snapshot just taken. On mismatch, halt with
  `RECONCILE_FAIL_CLASSIFICATION_DRIFT`. On match, **HALT before any close mutation when the
  bound verdict is not `CASCADE`**, with `RECONCILE_FAIL_CLASSIFICATION_REFUSED`."* The existing
  ``* **SAFE_CLOSE selected** (default, …`` bullet is **retained verbatim** and re-scoped by a
  leading clause *"When no `classification_binding` was supplied —"*.
* **CS14** → *"**`mode: classify-close-path` is strictly READ-ONLY.** It NEVER archives, NEVER
  transitions any status, NEVER calls `backlogit_ship_shipment`, and requires NO `file-lock`
  acquisition because no backlog/shipment artifact is ever mutated — its classification report
  is an additive-only write to a non-backlog-state location. DEGRADED (backlogit unreachable)
  is REPORTED as `CLOSE_PATH_VERDICT: BLOCK` and the mode HALTS."*

### 4.2 Dated ordinal observation (NON-NORMATIVE)

Measured 2026-09-13 against `.github/agents/_ship.agent.md` (1111 lines),
`.github/policies/workflow-policies.md` and `.github/skills/shipment-reconcile/SKILL.md`
(1082 lines). **These ordinals are an observation, not an
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
| CS11 | 56 (`SKILL.md`) |
| CS12 | 354 (`SKILL.md`, insertion point) |
| CS13 | 509 (`SKILL.md`) |
| CS14 | 245 (`SKILL.md`) |

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

## 6. Go contract test — 24 rows

One table-driven function, ≤4 helpers, **stdlib assertions only** (`testing`, `strings`), no
`testify`, reusing the package-level `repoRoot(t)` from `tests/integration/build_script_test.go`.
**Every row names its own file** — there is no implicit default file — because the suite now
spans three instruction surfaces and a needle legitimately present in one must be assertable
absent in another. Rows 1–19 read `.github/agents/_ship.agent.md`; row 20 reads
`.github/policies/workflow-policies.md`; rows 21–24 read
`.github/skills/shipment-reconcile/SKILL.md`.

| # | File | Assertion | Kind |
|---|---|---|---|
| 1 | ship | Role Boundary row contains the narrow feature-completion grant | positive |
| 2 | ship | The grant names all five conditions, the **at-every-depth** quantifier and the **set-equality** conjunct | positive ×3 |
| 3 | ship | Step 6.1(a1) exists, positioned **after** `a0` and **before** `a` | ordering |
| 4 | ship | The delegation block is present and names P-015 + the read-only `classify-close-path` boundary | positive |
| 5 | ship | CS1's reload clause also names P-015 and the merge commit | positive |
| 6 | ship | The authority-escalation rule is present | positive |
| 7 | ship | **(CS4)** the drifted `children`-only coverage rule is gone | negative |
| 8 | ship | `_ship.agent.md` does **not** name `TASK_ONLY_FINALIZE` | negative |
| 9 | ship | **(CS4)** No classification **predicate** — all four re-derived predicate sentences are **ABSENT**; a close-verdict allowlist is permitted | compound +/− |
| 10 | ship | **(CS2)** The `backlogit move <shipment_id> --status shipped` prescription is gone | compound +/− |
| 11 | ship | CS1's carve-out is scoped to **Role Boundary** changes only | consistency |
| 12 | ship | a1 states the `n == 0` no-op (S1) **and** that an unmet guard halts (S4) | consistency |
| 13 | ship | a1 states idempotent resume after conditions 1–4 revalidate (S3) | consistency |
| 14 | ship | **(CS7)** delegation phrase **PRESENT** **and** old queue-only literal **ABSENT** | compound +/− |
| 15 | ship | **(CS8)** delegation phrase **PRESENT** **and** old queue-only literal **ABSENT** | compound +/− |
| 16 | ship | a1 states `n > 1` halts with no mutation (S2) **and** the anomaly-gate-before-`n` rule (S0) | compound ++ |
| 17 | ship | **(CS8)** sentence **PRESERVES** the single-writer lock **and** the orphan scan | preservation |
| 18 | ship | **(CS7)** adjacent `Scope note (139-F/139.001-T)` **survives** | preservation |
| 19 | ship | **(CS9)** the protected-set claim is **scoped to safe-close** **and** the unscoped literal is **ABSENT** | compound +/− |
| 20 | policy | **(CS10)** P-010's `Ship MAY` list carries the grant (20a) **and** `workflow-policies.md` contains no `018.` literal (20b) | compound +/− |
| **21** | skill | **(CS11/CS12)** the mode enum carries `classify-close-path` (21a) **and** the section emits all three verdict tokens incl. the fully-covered-root `CASCADE` emission (21b) | compound ++ |
| **22** | skill | **(CS12)** the read-only clause is **PRESENT** (22a) **and** the no-mutation clause is **PRESENT** (22b) | compound ++ |
| **23** | skill | **(CS13)** the bound-refusal HALT is **PRESENT** (23a) **and** the legacy unbound fall-through bullet **SURVIVES** (23b) | compound + / preservation |
| **24** | skill | **(CS14)** the drift token and the read-only Behavioral Constraint are **PRESENT** (24a) **and** the four pre-existing public mode names **SURVIVE** (24b) | compound + / preservation |

**Kinds**: `compound +/−` = {9, 10, 14, 15, 19, 20} (**6**); `compound ++` = {16, 21, 22};
`negative` = {7, 8}; `preservation` = {17, 18, 23b, 24b}.

### 6.1 Pinned literals — ALL 24 rows (normative; single-line, CRLF-safe)

**Every row is pinned, every needle names its file, and every mandatory clause site has at
least one needle that resolves inside its own block.** No row derives its needle from prose,
and no acceptance criterion claims verification for a site without a needle here. Counts are
measured at the current tree (H0) over the file named in the row. **All H0 counts below were
re-measured on 2026-09-13 by the Scope-I and Scope-T commands in §13.4 / §13.6.**

| Row | File | Site | MUST BE ABSENT | MUST BE PRESENT | H0 |
|---|---|---|---|---|---|
| 1 | ship | ADD-1 | — | ``complete one covering feature `active -> done``` | 0 → RED |
| 2a | ship | ADD-1 | — | ``all five conditions below are conjunctive and fail-closed`` | 0 → RED |
| 2b | ship | ADD-1 | — | ``no live descendant remains at every depth`` | 0 → RED |
| 2c | ship | ADD-1 | — | ``the descendant graph union the feature is set-equal to the manifest`` | 0 → RED |
| 3 | ship | ADD-2 | — | ``a1. **Covering-feature completion gate**`` (index **>** `a0. **TOPOLOGY_GATE: lifecycle`, **<** `a. **Pre-archive reconciliation gate`) | 0 → RED |
| 4 | ship | CS4/CS6 | — | ``the `shipment-reconcile` skill's read-only `mode: classify-close-path` boundary selects the close path`` | 0 → RED |
| 5 | ship | CS1 | — | ``also re-read `P-015` and verify the merge commit`` | 0 → RED |
| 6 | ship | CS1 | — | ``A context reload MUST NOT widen the authority envelope`` | 0 → RED |
| 7 | ship | CS4 | ``it is fully covered (every one of its children,`` (×1) | — | 1 → RED |
| 8 | ship | guard | ``TASK_ONLY_FINALIZE`` — already ×0; **must stay ×0** | — | GREEN |
| 9a | ship | CS4 | ``permitted **only** when, for **every** feature member of the manifest: it is a`` (×1) | — | 1 → RED |
| 9b | ship | CS4 | ``enumerates to zero children, that childlessness is **positively verified**`` (×1) | — | 1 → RED |
| 9c | ship | CS4 | ``The manifest must contain nothing beyond the`` (×1) | — | 1 → RED |
| 9d | ship | CS4 | ``qualification is never per-member, and no feature ID is ever special-cased.`` (×1) | ``HALT on any result token the installed classification did not name`` | 1 / 0 → RED |
| 10 | ship | CS2 | ``<shipment_id> --status shipped`` (×1) | ``the authoritative close sequence defined by the `shipment-reconcile` skill`` | 1 / 0 → RED |
| 11 | ship | CS1 | — | ``scoped to Role Boundary changes only`` | 0 → RED |
| 12 | ship | ADD-2 | — | ``A1_NOT_APPLICABLE`` **and** ``HALT, naming the unmet condition`` | 0 / 0 → RED |
| 13 | ship | ADD-2 | — | ``A1_ALREADY_DONE`` | 0 → RED |
| 14 | ship | CS7 | ``every manifest item is present in `.backlogit/queue/` with the`` (×1) | ``defers every per-item status decision to that skill's `mode: pre` classification`` | 1 / 0 → RED |
| 15 | ship | CS8 | ``queue with `status: done`, and scans for orphan items.`` (×1) | ``defers every per-item status decision to the `mode: pre` classification at `expected_status: done``` | 1 / 0 → RED |
| 16 | ship | ADD-2 | — | ``evaluated before `n` is computed`` **and** ``halts with no feature mutation`` | 0 / 0 → RED |
| 17 | ship | CS8 | — | ``scans for orphan items`` — already ×2 (L256, L777); **must stay ≥2** | GREEN |
| 18 | ship | CS7 | — | ``Scope note (139-F/139.001-T)`` — already ×1 (L259); **must stay ×1** | GREEN |
| 19 | ship | CS9 | ``proves the protected set and halts fail-closed on any cascade or provenance`` (×1) | ``proves the protected set on the safe-close path`` **and** ``no protected set by construction`` | 1 / 0,0 → RED |
| 20a | policy | CS10 | — | ``complete one covering feature `active -> done``` in **`workflow-policies.md`** | 0 → RED |
| 20b | policy | guard | ``018.`` in **`workflow-policies.md`** — already ×0; **must stay ×0** | — | GREEN |
| **21a** | skill | CS11 | — | ``classify-close-path`` in **`SKILL.md`** | 0 → RED |
| **21b** | skill | CS12 | — | ``CLOSE_PATH_VERDICT: CASCADE`` **and** ``CLOSE_PATH_VERDICT: SAFE_CLOSE`` **and** ``CLOSE_PATH_VERDICT: BLOCK`` | 0 / 0 / 0 → RED |
| **22a** | skill | CS12 | — | ``### Classify-Close-Path Mode`` | 0 → RED |
| **22b** | skill | CS12 | — | ``performs no archive, no status transition and no record mutation`` | 0 → RED |
| **23a** | skill | CS13 | — | ``HALT before any close mutation when the bound verdict is not `CASCADE``` **and** ``RECONCILE_FAIL_CLASSIFICATION_REFUSED`` | 0 / 0 → RED |
| **23b** | skill | CS13 | — | ``* **SAFE_CLOSE selected** (default, including any classifier error,`` — already ×1; **must stay ×1** (the legacy unbound path is preserved verbatim) | GREEN |
| **24a** | skill | CS14 | — | ``RECONCILE_FAIL_CLASSIFICATION_DRIFT`` **and** ``**`mode: classify-close-path` is strictly READ-ONLY.**`` | 0 / 0 → RED |
| **24b** | skill | guard | — | ``mode: detect-mixed-role`` ×≥1, ``mode: safe-close`` ×≥1, ``mode: pre`` ×≥1, ``mode: post`` ×≥1 — **all four pre-existing public modes must survive** | GREEN |
| — | ship | guard | ``018.`` in `_ship.agent.md` — already ×0; **must stay ×0** | — | GREEN |
| — | ship | guard | ``classify-close-path`` in `_ship.agent.md` is permitted; ``CLOSE_PATH_VERDICT: BLOCK`` etc. are **verdict tokens, not predicates** — row 9a–9d already asserts no predicate survives | — | — |

**Independent falsifiability — the `K-2` resolution.** Every mandatory site now has at least
one needle that resolves **inside its own block**, so no site can be credited by an edit to a
different site:

| Site | Own needle | Located in |
|---|---|---|
| CS1 | rows 5, 6, 11 | the reload clause |
| CS2 | row 10 | `<shipment_id> --status shipped` |
| **CS3** | **row 9d PRESENT needle** | the `Do NOT call backlogit shipment ship` bullet — the only place a result-token allowlist can sit |
| CS4 | rows 7, 9a, 9b, 9c | four distinct survivor sentences of the CS4 block |
| **CS5** | **row 19's ABSENT needle is CS9's; CS5's own is row 10's PRESENT needle**, which replaces the `in place of the safe-close sequence above` framing | the close-sequence pointer |
| **CS6** | **row 4** | the new delegation bullet — its needle exists in no other block |
| CS7 | rows 14, 18 | the intake clause |
| CS8 | rows 15, 17 | the pre-archive clause |
| CS9 | row 19 | the protected-set bullet |
| CS10 | row 20a | `workflow-policies.md` |
| **CS11** | **row 21a** | the Inputs mode enum |
| **CS12** | **rows 21b, 22a, 22b** | the new section |
| **CS13** | **rows 23a, 23b** | the Step 0 dispatch |
| **CS14** | **rows 24a, 24b** | Behavioral Constraints |

**An `H1` that edits only the CS4 block now leaves rows 4, 10, 19, 20a and 21–24 red.**
Rows 9a–9d additionally make AC-4's *"no classification predicate"* claim **directly
falsifiable**: all four re-derived predicate sentences measured in the CS4 block are pinned
ABSENT, so a partial deletion fails the suite. Those same four sentences remain legitimately
**present** in `SKILL.md`, which is why every row names its file.

**Two classes, and only one is red → green:**

| Class | Rows | H0 state |
|---|---|---|
| **Transition rows** — must change | 1, 2a–2c, 3–7, 9a–9d, 10–16, 19, 20a, 21a, 21b, 22a, 22b, 23a, 24a (**26 assertions**) | **RED.** None can be vacuously green |
| **Guard rows** — must NOT change | 8, 17, 18, 20b, 23b, 24b, and the `_ship.agent.md` `018.` guard (**7**) | **GREEN today, by design.** They assert a property that already holds and must survive |

**The guard rows are green at H0 and that is correct** — their purpose is to fail if the
implementation *breaks* something, not to record a transition. Row 3's `a1` anchor is absent
pre-implementation **by design**; that absence is the H0 readiness guard. Rows 23b and 24b are
the **backward-compatibility guards for Option A**: they fail if the fourth-file edit narrows
or removes any pre-existing public mode or the legacy unbound fall-through (§3.4.3 F).

**Only the transition rows constitute the red → green evidence** for AC-3/AC-8/AC-24/AC-33.
Do not read the guard rows as proof the contract changed.

**Independently re-measured at revision 11** (commands in §13.4 and §13.6): every ABSENT
needle above resolves ×1 and every PRESENT needle ×0 in the live tree; `TASK_ONLY_FINALIZE`
×0; `018.` ×0 in both `_ship.agent.md` and `workflow-policies.md`; `scans for orphan items`
×2; `Scope note (139-F/139.001-T)` ×1; `record-consistent` ×0 in `_ship.agent.md` (it is
**added**, not preserved); `classify-close-path` ×0 and `CLOSE_PATH_VERDICT` ×0 in **all
three** instruction surfaces; `* **SAFE_CLOSE selected** (default, including any classifier
error,` ×1 in `SKILL.md`. `repoRoot(t)` exists at package scope in
`tests/integration/build_script_test.go`; `testify` is absent from `go.mod`.

## 7. Ordering

**Test-first, `main` stays green.** H0: write the 24-row test → **red**. H1: apply ADD-1,
ADD-2 and **CS1–CS14** → **green**. All fourteen clause sites are required to reach H1.

**Within H1, the fourth file lands first.** `CS11–CS14` (the classification boundary) are
applied before `CS3–CS6` (the delegation that consumes it), so `_ship.agent.md` never names a
mode the installed skill does not expose — the exact defect class `J-5`/`K-1` recorded twice.
This is an **ordering constraint inside one task**, not a task split; it is also the plan's
named mid-task checkpoint boundary (§14).

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

> **The `shipment-reconcile/SKILL.md` blob pin is WITHDRAWN at revision 11.** Under Option A
> that file is part of the §2 implementation surface and is modified by `H1`, so a pre-merge
> blob OID cannot survive the merge and is unpinnable by construction. Classifier identity is
> established instead by §3.4.2's three post-merge facts, verified at §9.2 check 7. This is a
> **substitution of mechanism, not a relaxation**: the withdrawn pin proved *"the file did not
> change"*; the replacement proves *"the file changed only as reviewed, and has not changed
> since"*, which is strictly what is needed once the file is legitimately edited.

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
| 1 | **S1** | **Step 0.5 Shipment Intake.** Verify on `main`; **P-011** branch-before-mutation + **P-016** single-worktree check; pre-claim topology gate; **intake `mode: pre` with `expected_status: queued` — run BEFORE the claim, while the manifest is uniformly `queued` and therefore representable (§9.3)**; then **claim `021-S`** |
| **1a** | S1 | **POST-CLAIM CONDITION — `022-F` must be `active` (checked, not assumed).** Re-read `022-F` **by a single `backlogit_get_item` record read** and require its live status to be **exactly `active`**. **This is NOT a manifest-wide `mode: pre` invocation** — see §9.3. **Any other value ⇒ HALT** and return to Stage. **Ship is granted NO authority to transition a feature `queued -> active`** — §3.2 grants exactly one transition, `active -> done`. *(Live state at planning time is `queued`; the shipment claim is what is expected to activate it. If it does not, that is a planning-state error only Stage can reconcile.)* |
| **1b** | S1 | **Step 2 Harness Generation (P-002 / P-004).** The **harness-architect** produces the 24-row harness: `go vet ./...` exits 0, `go test ./...` exits non-zero with expected failure markers ⇒ **H0 RED CONFIRMED** ⇒ `harness-ready` applied to `022.001-T`. **This precedes the TASK claim** — P-002 permits claiming a task only after red-phase confirmation |
| **1c** | S1 | **Step 3 Build Ready Queue** — filtered to tasks carrying `harness-ready` |
| **1d** | S1 | **Step 4.1 Claim Task** — `022.001-T -> active` |
| 2 | S1 | **Step 4.2 implement** → apply **CS11–CS14 first** (§7), then ADD-1, ADD-2 and **CS1–CS10** → **H1 green** |
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
| 14 | S2 | **Pre-mode**, `expected_status: done`. **Every manifest member declares `done` at this point** (`022-F` by step 13, `022.001-T` by step 4), so the single `expected_status` is representable (§9.3). **Requires an authoritative `PROCEED`**; anything else HALTs |
| **14a** | S2 | **CLASSIFICATION BOUNDARY (read-only, pre-mutation).** Invoke **`mode: classify-close-path`** with `shipment_id: 021-S`. **This call mutates nothing.** Read `CLOSE_PATH_VERDICT`. **Anything other than `CASCADE` ⇒ HALT HERE, with no close call made at all** (§3.5); record `VERDICT_REASON` + `VERDICT_EVIDENCE` and return to Stage. On `CASCADE`, carry `CLASSIFICATION_BINDING` forward to step 15 |
| 15 | S2 | **Invoke `mode: safe-close`** (§3.4.1) **passing `classification_binding` from 14a**. Safe-close **revalidates the binding before any mutation** (§3.4.3 E) and halts on drift (`RECONCILE_FAIL_CLASSIFICATION_DRIFT`) or on a bound non-`CASCADE` verdict (`RECONCILE_FAIL_CLASSIFICATION_REFUSED`); otherwise the pre-existing Cascade Close Sub-Procedure runs. P-007 archive-integrity verify; post-mode. **Capture `{post_paths}` immediately on return, before any commit or cleanup** (§10.3.1). **The cascade output is committed ONLY after every postcheck passes** (§10.3) |
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

7. **Classifier identity (§3.4.2)** — the three post-merge facts, all verified against the
   merge commit confirmed at step 6: (a) the 24-row contract test passes at `{merge_sha}`,
   **including rows 21–24 over `.github/skills/shipment-reconcile/SKILL.md`**; (b)
   `git diff --name-only {merge_base}..{merge_sha}` intersected with the tracked tree equals
   exactly the four files of §2; (c)
   `git diff --quiet {merge_sha} -- .github/skills/shipment-reconcile/SKILL.md` exits 0.
   **Any of the three failing ⇒ HALT.**

> **Why check 7 is no longer a blob pin.** Revision 10 pinned `SKILL.md`'s blob OID and noted
> it was *"safely pinnable"* precisely because the file sat **outside** the implementation
> surface. Option A moves it **inside** that surface, so the premise no longer holds and the
> pin is **withdrawn** (§8.1). `_ship.agent.md` and `workflow-policies.md` were already
> deliberately unpinned for the same reason; `SKILL.md` now joins them, and all three are
> covered by facts (a)+(b)+(c), which bind them to the reviewed merge rather than to a
> pre-merge byte image.

### 9.3 Pre-mode invocation sites — exactly two, both representable (the `K-6` resolution)

The installed intake check (`_ship.agent.md` L253–262, the **CS7** site) runs `mode: pre` with
a **single** `expected_status` over **every** manifest item, and its own adjacent Scope note
states outright that a manifest whose members hold **different** statuses necessarily yields a
`status-mismatch` ⇒ `RECONCILE_FAIL`. A plan that sequences a manifest-wide pre-mode run while
`021-S` is `[022-F active, 022.001-T queued]` therefore prescribes a call the installed tooling
**cannot represent**. Revision 10 did exactly that.

**This plan invokes `mode: pre` at exactly two sites, and no others:**

| Site | When | Manifest state | `expected_status` | Representable? |
|---|---|---|---|---|
| §9 step **1** | **BEFORE** the claim | `[022-F queued, 022.001-T queued]` | `queued` | **YES** — uniform |
| §9 step **14** | after step 13's a1 transition | `[022-F done, 022.001-T done]` | `done` | **YES** — uniform |

**The mixed window is `[step 1a, step 13)`**, during which the manifest is
`[022-F active, 022.001-T queued → done]`. **No manifest-wide `mode: pre` invocation occurs
anywhere in that window.** Step 1a is a **single-record `backlogit_get_item` read of `022-F`**
— one record, one status, no aggregate — and a1's §5 selector likewise reads records directly
(§5's executable-surface table) and explicitly **must not invoke pre-mode**.

**Consequences, stated so they are checkable:**

* Moving the intake pre-mode run to **before** the claim matches the installed skill's own
  documented intake semantics — *"`queued` for fresh intake; `active` when the shipment was
  already claimed in a prior session"* — and this is a **fresh** intake.
* The CS7 edit is **pointer-only** and does **not** change the single-`expected_status`
  contract. This plan does not ask the installed pre-mode to represent a mixed manifest; it
  **avoids creating one at an invocation site**.
* `AC-37` asserts the two-site inventory and the empty mixed window.

## 10. Recovery

### 10.1 HALTs before the CLOSE — nothing to unwind, but a1 may already have run

a1 S0–S5; pre-mode `RECONCILE_FAIL`; a non-`CASCADE` verdict at the **step 14a classification
boundary**; a binding drift or bound-refusal HALT inside safe-close; §9.2 failure. In every
case **no CLOSE mutation has occurred**: `021-S` stays `active`, is not shipped and not
archived.

**This is now a verified property of the call sequence, not an assertion** (`K-1`). The
step-14a boundary is a **read-only** call (§3.4.3 B), and the only path from it to a mutation
runs through safe-close's binding revalidation (§3.4.3 E), which halts **before** its first
write. There is no call ordering in this plan under which a non-`CASCADE` verdict is observed
after a close mutation.

**Precise state, by position in §9:**

| HALT | a1 (step 13) already ran? | Feature state | Resume |
|---|---|---|---|
| §9.2 failure (step **12a**) | **No** — 12a precedes a1 | `022-F` still `active` | §5 **S4** |
| a1 S0–S5 (step **13**) | halts **within** a1, before its own S4 write | `022-F` still `active` | §5 **S4** |
| pre-mode `RECONCILE_FAIL` (step **14**) | **YES** | `022-F` is **`done`** | §5 **S3** |
| verdict not `CASCADE` (step **14a**) | **YES** | `022-F` is **`done`** | §5 **S3** |
| binding drift / bound refusal (step **15**) | **YES** | `022-F` is **`done`** | re-run step **14a**, then §5 **S3** |

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
2. **Restore — NON-DESTRUCTIVE, by quarantine-then-revert (the `K-5` resolution).**
   Revision 10 restored by `git restore --source {pre_cascade_sha}` (an **unbacked
   overwrite**) plus **deletion** of `{post_paths} − {pre_paths}` — two operations Principle
   VII names explicitly as destructive, executed with **no approval gate**, while §11 states
   the operator authorization *"covers exactly the two actions below and nothing else."* The
   plan therefore declared its own mandatory recovery path unapproved. **That restore is
   withdrawn.** It is replaced by a sequence containing **no destructive primitive at all**:

   a. **Quarantine the post-cascade state.** `git add -A -- .backlogit/queue/ .backlogit/archive/`
      then commit as `{quarantine_sha}` on the **existing** closure branch, message
      `chore(quarantine): post-cascade state of {shipment_id}, postcheck FAILED`. This
      **captures**, rather than discards, everything the cascade produced — including every
      path in `{post_paths} − {pre_paths}`, which becomes tracked content of this commit.
   b. **Revert it.** `git revert --no-edit {quarantine_sha}`. Reverting a commit that
      *added* those paths **removes them through ordinary merge history**, and restores every
      modified tracked file to its `{pre_cascade_sha}` content — achieving exactly the
      revision-10 end state with **no overwrite of uncommitted work and no file deletion**,
      because nothing is uncommitted at that point (step a committed it) and nothing is
      deleted (the revert commit supersedes it).
   c. **Nothing else is touched.** The `-- .backlogit/queue/ .backlogit/archive/` pathspec in
      step a bounds the quarantine to the two trees; step b's revert is bounded by step a's
      commit.

   **Why this is in-approval.** `git revert` is the mechanism §10.4 and Principle XI
   (NON-NEGOTIABLE) already designate as this plan's sanctioned, non-destructive unwind, and
   §15 records it as such. The restore therefore needs **no third approved-action record and
   no new operator authorization**, and §11's *"exactly the two actions and nothing else"*
   scope is respected rather than exceeded. The `{pre_paths}` / `{post_paths}` inventories
   are **retained** — not as delete targets, but as the **verification** inputs of step 3.
3. **VERIFY the restore — both halves:**
   * **Tree equivalence** — the current sorted path list, re-enumerated with the **same
     command** as §10.3.1, equals `{pre_paths}`, and `git status --porcelain` over both trees
     is **empty** (content identical to `{pre_cascade_sha}`); **and**
   * **Record equivalence** — every record in the Step 0(b) snapshot matches it on declared
     `status`, `parent_id` and location.

   **An unverified restore is not a restore.**
4. **HALT.** No retry, no second close path, no re-invocation of the cascade.

> **The cascade result is NEVER committed AS A RESULT before its postchecks pass (bounded
> state machine).** The authorized flow is strictly:
>
> ```text
> 13a commit baseline  ->  {pre_cascade_sha}, clean trees, {pre_paths}
> 15  invoke cascade   ->  mutates the WORKTREE only
> 15  on return        ->  capture {post_paths} IMMEDIATELY (unconditional)
> 15  postchecks       ->  returned_ids / two-set / parent_id / P-007 / post-mode
>        pass  -> commit the cascade result (first RESULT commit after 13a)
>        fail  -> quarantine-commit (step 2a) -> revert (step 2b) -> verify (step 3) -> HALT
> ```
>
> **The quarantine commit is not a result commit.** It is an explicitly-labelled evidence
> capture whose only purpose is to make the unwind non-destructive, and it is immediately
> superseded by its own revert. The invariant that matters is preserved exactly: **no
> postcheck-failed cascade state is ever left standing as the branch tip.**
>
> **A committed cascade awaiting postchecks is therefore unreachable in this plan**, which is
> why §10.3 is worktree-level and sufficient. P-015's `git revert` remedy addresses the
> general case where a cascade was already committed; **on this route that state cannot
> arise**, and §10.4's `git revert` is reserved for the *merge*, not the cascade.
> **If an executor ever finds the cascade result already committed as a result before its
> postchecks completed, that is an out-of-contract state: HALT immediately, do not restore,
> and return to the operator** — the §10.3 restore is not valid against a committed result.

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
| **Execution site** | **§9 steps 13a–15 of THIS plan** (13a pre-cascade baseline, 14 pre-mode, **14a classification boundary**, 15 the bound cascade invocation) | **NOT IN THIS PLAN — see §11.1** |
| **Rollback** | §10.3 — pre-cascade commit + both path inventories + snapshot, unwound **non-destructively by quarantine-then-revert**, **verified**, then HALT | §10.3, **applied within `017-S`'s own closure plan** (§11.1) |
| **approval_required** | **YES** (non-negotiable) | **YES** |
| **ActionRisk** | **`destructive`** | **`destructive`** |
| **ActionResult** | **`approved`** | **`approved`** |

**`017-S` manifest (13)**: `018-F`, `018.008-T`, `018.001-T`, `018.001.001-ST`,
`018.001.002-ST`, `018.001.003-ST`, `018.002-T`, `018.002.001-ST`, `018.002.002-ST`,
`018.002.003-ST`, `018.003-T`, `018.003.001-ST`, `018.003.002-ST`.

**Conditions — all must hold at invocation:**

1. **`mode: classify-close-path` returns `CLOSE_PATH_VERDICT: CASCADE`** at §9 step 14a —
   a **read-only** call made **before** any close call — **and** `mode: safe-close`
   revalidates the returned `CLASSIFICATION_BINDING` successfully at step 15. Every other
   outcome (§3.5), including `SAFE_CLOSE`, `BLOCK`, a binding drift and a bound refusal,
   **voids the approval and HALTs before any close mutation**;
2. **all preflight gates pass** — §9.2's seven checks, the §5 S0–S5 selector, and pre-mode
   `PROCEED` at the two representable invocation sites of §9.3;
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
3. **AC-3** All **fourteen** clause sites CS1–CS14 are applied; **every one of them has at
   least one pinned needle in §6.1 that resolves inside its own block** (CS1 → rows 5/6/11;
   CS2 → row 10; CS3 → row 9d's PRESENT needle; CS4 → rows 7/9a/9b/9c; CS5 → row 10's PRESENT
   needle; CS6 → row 4; CS7 → rows 14/18; CS8 → rows 15/17; CS9 → row 19; CS10 → row 20a;
   CS11 → row 21a; CS12 → rows 21b/22a/22b; CS13 → rows 23a/23b; CS14 → rows 24a/24b);
   the 24-row test passes; H0 was red.
4. **AC-4** `_ship.agent.md` states **no classification predicate** — all four re-derived
   predicate sentences measured in the CS4 block are **ABSENT** (rows 7, 9a, 9b, 9c) — and
   contains **no `018.` literal**.
5. **AC-5** The close-verdict allowlist is exactly `CASCADE` / `SAFE_CLOSE` / `BLOCK`, where
   `BLOCK` is a **refusal and not a close path**; `HALT` and `RECONCILE_FAIL` are non-close
   outcomes; unknown tokens, absent, empty, ambiguous or unreadable results **HALT**. **No
   runtime `policy_id`, version or source-disagreement check exists.**
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
11. **AC-11** Exactly **4 implementation files**, **0 new production files**, **0 new module
    dependencies**.
12. **AC-12** The two sessions are distinct: S1 never performs a1 or the close.
13. **AC-13** S2's recovery is the **installed** `_ship.agent.md` crash-resumption protocol,
    executed unmodified (§9.1).
14. **AC-14** For `021-S`/`017-S` a non-`CASCADE` verdict **HALTs at the §9 step 14a
    classification boundary, before any close call is made**; the halt is reachable because
    that call is read-only (§3.4.3 B); no `SAFE_CLOSE` fallback.
15. **AC-15** Every post-CASCADE failure performs the §10.3 **verified, non-destructive**
    unwind — quarantine-commit, revert, then tree equivalence **and** record equivalence —
    then HALTs. **No `git restore --source` overwrite and no file deletion is prescribed
    anywhere in the recovery path**, so no third approved-action record is required and §11's
    two-action scope is not exceeded. The cascade is not invoked unless §10.3.2 step 1's clean
    trees, pre-cascade commit, `{pre_paths}` and Step 0(b) snapshot all exist.
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
27. **AC-27** **(J-5, K-1)** The delegation target is the installed **read-only
    `mode: classify-close-path`** boundary added by CS11–CS14; the verdict is **returned by
    that call**, not scraped from a post-mutation report. **No plan text requires an API the
    post-`H1` skill does not expose**, and no Ship runtime code is added.
28. **AC-28** **(J-9)** Containment and orphan detection are specified via **`backlogit doctor`
    only**. `backlogit_query_sql` is explicitly excluded for containment, with the reason
    recorded.
29. **AC-29** **(J-8, J-15)** Evidence identity is verified against **plan-pinned** blob OIDs
    and a plan-pinned engine digest, cross-checked against the transcript, with divergence
    HALTing; and the classifier's identity is verified by §3.4.2's **three post-merge facts**
    at §9.2 check 7 (contract, surface, working identity). The withdrawn blob pin is recorded
    as withdrawn in §8.1 with its reason. A same-PR transcript+script replacement cannot pass
    without an additional visible edit to this reviewed plan.
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
33. **AC-33** **(K-1, Option A)** `.github/skills/shipment-reconcile/SKILL.md` exposes a
    **read-only** public mode `classify-close-path` that runs the existing Step 0(a)/(b)/(c)
    logic, **stops at the dispatch**, and returns `CLOSE_PATH_VERDICT` ∈
    {`CASCADE`, `SAFE_CLOSE`, `BLOCK`} with `VERDICT_REASON`, `VERDICT_EVIDENCE` and
    `CLASSIFICATION_BINDING`. Rows 21a, 21b, 22a pass.
34. **AC-34** **(fully-covered-root CASCADE case)** A manifest satisfying §3.2.1 — a root
    feature whose full live descendant graph at every depth is set-equal to the manifest minus
    the feature — classifies as `CLOSE_PATH_VERDICT: CASCADE`, and that is the **only**
    verdict that permits the Cascade Close Sub-Procedure. Row 21b pins the emission; the
    selecting predicate is the unchanged installed Step 0(c).
35. **AC-35** **(no mutation on non-`CASCADE`/`BLOCK`)** The classification call archives
    nothing, transitions nothing and takes no lock (rows 22a, 22b); and safe-close, given a
    matching binding whose verdict is not `CASCADE`, **HALTs before any close mutation** with
    `RECONCILE_FAIL_CLASSIFICATION_REFUSED` rather than falling through to steps 1–10
    (row 23a). Both are independently falsifiable.
36. **AC-36** **(safe-close case + backward compatibility)** With **no** `classification_binding`
    supplied, `mode: safe-close` behaves exactly as today, including its legacy default
    fall-through, and all four pre-existing public modes survive unchanged. Rows 23b and 24b
    pass. No existing caller is broken and no public mode is narrowed or removed.
37. **AC-37** **(K-6)** `mode: pre` is invoked at exactly **two** sites (§9 steps 1 and 14),
    each over a **uniform** manifest with a representable single `expected_status`; **no
    manifest-wide pre-mode invocation occurs in the mixed window `[1a, 13)`**; step 1a is a
    single-record read. The intake run is moved **before** the claim, matching the installed
    skill's documented `queued`-for-fresh-intake semantics.
38. **AC-38** **(TOCTOU)** `CLASSIFICATION_BINDING` is a digest over a canonical serialisation
    of exactly the manifest + dependency + per-record snapshot the classification consumed
    (§3.4.3 D), and `mode: safe-close` **recomputes and compares it before any mutation**,
    halting with `RECONCILE_FAIL_CLASSIFICATION_DRIFT` on mismatch (row 24a). Classification
    and mutation therefore execute against the same bound snapshot or not at all.

## 13. Gate state

```text
plan-review-attempt: 7
dispatch_mode: multi-agent
decision: FAIL
P0: 0
P1: 27
personas: 7/7
anchor: architecture-strategist gpt-5.6-sol high
authorization: SPENT
```

> **Revision 11 resolves `K-1`…`K-7`** under explicit operator selection of **Option A**.
> The attempt-6 verdict, its per-persona counts and the `K-1`…`K-7` blocker set remain
> recorded in `docs/reviews/2026-09-13-true-lineage-attempt-6-verdict.md`; the attempt-7
> verdict is recorded in `docs/reviews/2026-09-13-true-lineage-attempt-7-verdict.md`.
> **Attempt 7 has been RUN and its authorization is SPENT.** It returned **FAIL** with **P0 = 0** (the attempt-6 P0 is closed) and **P1 = 27 raw / 15 deduplicated**. **No attempt 8 is authorized and no post-review patch cycle was performed.** A further gate requires fresh operator authorization.

### 13.1 Current state

| Field | Value |
|---|---|
| **Plan `status:`** | `planned` |
| **Revision** | **11** — reauthored, not patched |
| **True-lineage attempt** | **7** (the counter continues across the revision-10 → 11 reauthoring; it is **not** reset) |
| **Gate decision** | **FAIL** — P0 = 0, P1 = 27 raw / 15 deduplicated, 7/7 personas. See `docs/reviews/2026-09-13-true-lineage-attempt-7-verdict.md` |
| **Harvest-ready** | **NO** — the §13.1 rule requires P0 = 0 **and** P1 = 0. P0 is met; P1 is not. No harvest and no shipment assembly were performed. |
| **`021-S`** | `queued` — **not claimable** until the gate returns P0 = 0 and P1 = 0 |
| **`017-S`** | `queued`, dependency `[021-S]` — ineligible until `021-S` ships, **and** separately ineligible until an `017-S` closure plan exists (§11.1) |
| **`PA-021-CASCADE`** | recorded, **unexercised**; **manifest unchanged at `[022-F, 022.001-T]`**, so the approval is **not** re-scoped by revision 11 and needs no re-authorization |
| **`PA-017-CASCADE`** | recorded, **unexercised**, **held in escrow** (§11.1) |
| **Authorization** | Explicit operator selection of **Option A** on 2026-09-13, authorizing the fourth-file scope expansion and **exactly one** further gate as true-lineage attempt 7. |

**Lineage (compact — narrative deliberately not restated inline):**

| Attempt | Revision | Verdict |
|---|---|---|
| 1 | 4 | FAIL |
| 2 | 5 | FAIL |
| 3 | 6 | FAIL (re-entry budget exhausted, escalated) |
| 4 | 7 | FAIL (`H-1`…`H-9`) |
| 5 | 9 | FAIL (`J-1`…`J-18`) |
| 6 | 10 | FAIL (`K-1` P0 + `K-2`…`K-7` P1) |
| **7** | **11** | **FAIL** — P0 = **0** (K-1 P0 closed; K-4, K-5 closed), P1 = 27 raw / 15 deduped. Option A applied to the skill side only; CS3-CS6, CS10 and the unbound sink were not brought along. |

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
| **Attempt-6 verdict, `K-1`…`K-7` full text and per-persona counts** | **`docs/reviews/2026-09-13-true-lineage-attempt-6-verdict.md`, commit `8066483`** |
| **Attempt-7 verdict (Option A)** | **`docs/reviews/2026-09-13-true-lineage-attempt-7-verdict.md`** |
| **Attempt-7 reauthoring session record** | **`docs/memory/2026-09-13-stage-true-lineage-attempt-7-option-a.md`** |
| Recovery / torn-state halt record | `docs/memory/2026-09-13-stage-recovery-torn-state-halt.md`, commit `3e0cf8f` |
| Checkpoints | `.backlogit/checkpoints/checkpoint-20260913-231952.json` (attempt-5 terminal state) and its siblings |
| Change lineage for this file | `git log --follow -- docs/plans/2026-09-13-intercom-go-ship-feature-completion-decided-plan.md`; PR #54 |

**Reauthoring does not discharge findings.** `H-1`…`H-9` were disposed at attempt 5.
`J-1`…`J-18` and `K-1`…`K-7` are disposed in §13.3 below, each pointing at the section that
resolves it.
**A finding is closed by a landing site, never by the disappearance of the text that carried
it.**

### 13.3 Findings dispositions at revision 11

**Attempt-6 blockers (`K-1`…`K-7`) — all RESOLVED:**

| Ref | Sev | Landing site |
|---|---|---|
| **K-1** the installed skill exposes no pre-mutation verdict boundary | **P0** | **§2 (fourth file) + §3.4.3 + CS11–CS14 + §9 step 14a + rows 21–24.** OPTION A, operator-selected: a new **read-only** `mode: classify-close-path` returns the verdict **before any close call is made**, and safe-close revalidates the bound classification **before any mutation**. §10.1's *"no CLOSE mutation has occurred"* is now a property of the call sequence, not an assertion. AC-33, AC-34, AC-35, AC-36, AC-38. |
| **K-2** CS3/CS5/CS6 not independently falsifiable; AC-4 undetectable | P1 | **§6.1** — every mandatory site now has a needle **inside its own block** (the per-site table); CS3 → row 9d PRESENT, CS5 → row 10 PRESENT, CS6 → row 4; and all **four** measured CS4 survivor sentences are pinned ABSENT (rows 7, 9a, 9b, 9c), making AC-4 directly falsifiable. AC-3, AC-4. |
| **K-3** row 2's needle unsatisfiable by the plan's own pinned wording | P1 | **§4.1 ADD-1 is now pinned verbatim**, and rows 2a/2b/2c draw their needles **from that pinned text** — including the `at every depth` quantifier (2b) and the set-equality conjunct (2c), both asserted **in `_ship.agent.md`**. Row 2a's literal `all five conditions below are conjunctive and fail-closed` appears verbatim in the pinned ADD-1 and CS10 wording. |
| **K-4** `PA-021-CASCADE` names the wrong execution site | P1 | **`.backlogit/queue/021-S.md`** corrected to *"plan section 9 steps 13a-15"*, matching §11's `Execution site` cell, which additionally enumerates 13a/14/14a/15. |
| **K-5** the rollback is itself an unapproved destructive operation | P1 | **§10.3.2 step 2** — the `git restore --source` overwrite and the file deletion are **withdrawn** and replaced by **quarantine-commit → `git revert`**, which contains **no destructive primitive**. `git revert` is the mechanism §10.4 and Principle XI already designate. **No third approval is required** and §11's two-action scope is respected. AC-15. |
| **K-6** §9 steps 1/1a produce a manifest the installed intake check cannot represent | P1 | **§9.3** — pre-mode is invoked at exactly **two** sites, both over a **uniform** manifest; the intake run moves **before** the claim (`expected_status: queued`); step 1a is a **single-record read**, not a manifest-wide run; the mixed window `[1a, 13)` contains **no** pre-mode invocation. AC-37. |
| **K-7** the measurement discipline omits the scope where every failure occurred | P1 | **§13.6 — Scope T (installed tooling)** is added with command, verbatim output and pinned digest for every tooling claim, covering `SKILL.md`, the backlogit engine and its index schema. §13.5's bullet 1 citation is corrected to the lesson the source actually teaches. |

**Attempt-5 findings — dispositions carried forward (all remain landed at revision 11):**

**P1 (`J-1`…`J-9`) — all REMEDIATED:**

| Ref | Landing site |
|---|---|
| **J-1** incomplete rev sweep inside the lineage section | **Structurally eliminated.** §13 carries no inline lineage narrative and no per-revision prose; the revision appears in frontmatter and §13.1 only. Verified by §13.4 measurement. |
| **J-2** occurrence claims scoped to the plan while literals survived in 4 records | §13.4 — every measurement is declared over **{plan} ∪ {4 backlog records}**, with the exact reproducible command. The records are reauthored, so no defect literal survives to be excluded. |
| **J-3** `{post_paths}` consumed but never defined | **§10.3.1** — both `{pre_paths}` and `{post_paths}` defined with command, roots and capture moment; `{post_paths}` captured unconditionally on cascade return. AC-25. |
| **J-4** P-010 grants task transitions only; conflicts with the Ship role table | **§2 + CS10 + §4.1 + AC-24** — surgical third-file scope expansion, exactly one bullet of P-010's `Ship MAY` list, pinned wording matching ADD-1, asserted by test row 20a/20b. |
| **J-5** delegation targets a classify API the skill does not expose | **§3.4.1 + §3.4.3** — at revision 11 the delegation targets the **read-only `mode: classify-close-path`** boundary that CS11–CS14 **add to the installed skill**, so the API the plan names is the API the post-`H1` skill exposes. (Revision 10's retarget to `safe-close` Step 0(c) removed the missing-API defect but left the verdict unreadable before mutation — that residue was `K-1`, resolved here.) AC-27. |
| **J-6** plan recovery algorithm conflicts with installed `_ship.agent.md` | **§9.1** — the competing algorithm is **withdrawn in full**; the installed protocol governs verbatim (`consumer_id: "ship"`, operator selection, operator confirmation). AC-26. |
| **J-7** `PA-017-CASCADE` cites a non-applicable action | **§11.1** — the §9 step 13a reference is withdrawn; the approval is bound to a future Stage-authored `017-S` closure plan and recorded as a third ineligibility ground. AC-31. |
| **J-8** evidence ancestry defeatable by same-PR replacement | **§8.1 + §8.1.1 + §9.2 checks 3/4/5** — blob OIDs and the engine digest are **pinned in this reviewed plan**, cross-checked against the transcript, divergence HALTs. AC-29. |
| **J-9** `query_sql` cannot detect duplicate paths | **§5 executable-surface table** — containment and orphans via **`backlogit doctor`** only; `query_sql` explicitly excluded with the measured reason. AC-28. |

**P2 (`J-10`…`J-18`) — all addressed:**

| Ref | Landing site |
|---|---|
| **J-10** protected-set claim in the CS2→CS3 gap | **CS9** (new anchored site) + **§4**'s switch to anchor-addressed sites, which removes ordinal gaps as a class. Test row 19. |
| **J-11** never-prune allowlist lost 1 of 3 elements | **§9.1** names all three explicitly. AC-26. |
| **J-12** AC-3 claimed test coverage for sites without needles | **AC-3** now enumerates the needle for **every** CS1–CS14 site; §6.1 pins all **24** rows. |
| **J-13** only 9 of 18 rows pinned | **§6.1** pins **all 24 rows**, both polarities, with H0 counts. |
| **J-14** backlog records append-only | All four records **reauthored to clean current state** in the same commit as this revision. |
| **J-15** classifier byte-identity unverified | **§3.4.2** + **§9.2 check 7**. The revision-10 blob pin is **withdrawn** (§8.1) because Option A puts the file inside the implementation surface; identity is now established by contract identity, surface identity and working identity against the reviewed merge commit. AC-29. |
| **J-16** condition-2 manifest vs full-graph ambiguity | **§3.2.1** — full live descendant graph at every depth, plus set equality with the manifest. AC-30. |
| **J-17** uncited prior art | §9 (P-018 trio), §14 (`…sizing-is-wit-gated`), §9 note (`…shipment-status-constraints`), and §13.5 (`…adversarial-review-empirical-verification-and-scope-check`, `…verify-go-stdlib-internals-claims-against-goroot-source`). |
| **J-18** `a1 n>1` halt narrower than P-015 | **§5 S2** is now an explicit **HALT, no mutation**, with the P-015 alignment stated. AC-32. |

### 13.4 Measurement — three scopes, declared separately

`J-2` was caused by stating one count and implying a wider scope. `K-7` was caused by declaring
only two scopes while asserting facts about a **third**. There are **three distinct scopes**
here and a count is meaningless without naming which one it is.

* **Scope I — the implementation surface (instruction files)**: `_ship.agent.md` ∪
  `workflow-policies.md` ∪ `shipment-reconcile/SKILL.md`. This is what the 24-row test asserts
  over. **All Scope-I counts live in §6.1**, pinned per row, and are not restated here.
* **Scope A — the artifact set**: `{this plan}` ∪ `{021-S, 017-S, 022-F, 022.001-T}` — five
  files. This is what the reauthoring had to clean.
* **Scope T — the installed tooling**: the `shipment-reconcile` skill **as installed**, the
  backlogit engine and its index schema. **Every claim this plan makes about installed
  behaviour is measured here, in §13.6, with command + verbatim output + digest.** `H-3`,
  `H-6`, `J-5`, `J-9` and `K-1` all failed in this scope; it now has a declared scope and a
  reproduction rule like the other two.

> `SKILL.md` is in **both** Scope I and Scope T at revision 11, and that is not a conflict: as
> Scope I it is a file this plan **changes** and pins by test row; as Scope T it is the tool
> whose **current installed behaviour** this plan makes assertions about. §13.6 measures the
> pre-change (`H0`) tool; §6.1 rows 21–24 pin the post-change (`H1`) contract.

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
| `re[v]ision: 10` | **0** | the superseded frontmatter revision survived |
| `re[v] 8` | **0** | a stale short-form revision survived (`J-1`) |
| `re[v] 9` | **0** | same |
| `AUTHORITATIVE GATE [S]TATE` | **0** | the dangling `H-1` heading survived in plan **or** records |
| `SUPERSEDES EVERY [B]LOCK` | **0** | append-only stacking survived in a record (`J-14`) |
| `2 [F]ILES` | **0** | the superseded 2-file surface survived a record (`J-2`, `J-4`) |
| `3 [F]ILES` | **0** | the **revision-10** 3-file surface survived a record (`K-1`) |
| `3 [I]MPLEMENTATION` | **0** | same, long form |
| `6 [S]ITES` | **0** | the superseded clause count survived a record (`J-2`) |
| `10 [A]NCHOR` | **0** | the revision-10 clause count survived a record (`K-1`) |
| `20 [r]ows` | **0** | the revision-10 row count survived anywhere (`K-1`) |
| `step [1]6` | **0** | the mis-anchored `PA-021-CASCADE` execution site survived (`K-4`) |
| `attempt [6]` | **0** | a stale gate-lineage claim survived (`K-1`) |
| `MUST NOT enum[e]rate` | **0** | withdrawn `H-6` wording survived |
| `without touching this [f]ile again` | **0** | withdrawn `H-5` wording survived |

*Measured 2026-09-13 at revision 11 over the five Scope-A files; every row above returned
exactly the stated **0**.*

**Bounded, not zero — and deliberately so:**

| Regex | Scope-A count | Every occurrence is |
|---|---|---|
| `re[v]ision 9` | **2** | both are §13.2 audit-pointer rows (L1365, L1366). **No normative use.** |
| `re[v]ision 10` | **6** | the §0 supersession banner, four explicit "what the superseded revision did and why it changed" contrast notes (§3.4.2, §9.2 check 7, §9.3, §10.3.2), and one §13.3 disposition cell. **Every one is a historical contrast; none is a governing pointer.** |
| `re[v]ision 11` | **24** | the current governing revision, named in the plan and in **all four** records. |
| `attempt [7]` | **6** | the current gate lineage, in the plan and in all four records. |
| `24 [r]ows` | **6** | the current row count: four plan sites (§5, §6, §13.3, §17-adjacent §14) + `022-F` + `022.001-T`. |
| `backlogit_query[_]sql` | **3** | §5's prohibition, AC-28's restatement, and the §13.6 `T-6` measurement that proves the prohibition. **Never a prescription** (`J-9`). |
| `{post[_]paths}` | **≥ 2** | §10.3.1's definition and its §10.3/§10.4 consumption. Definition precedes consumption (`J-3`). |

**Rule for any future claim over these artifacts: name the scope, give the regex, give the
number.** A bare occurrence claim is not evidence.

### 13.5 Prior art relied upon

* `docs/compound/2026-09-08-adversarial-review-empirical-verification-and-scope-check.md` —
  **independently reproduce a claim about installed behaviour before asserting it**; measuring
  the artifact you are editing is not a substitute for measuring the tool you are depending
  on. This is the lesson the source actually teaches, and citing the weaker *"declare your
  count scope"* reading is how the stronger one was evaded at `H-3`, `J-2`, `J-9` and `K-1`.
  §13.6 (Scope T) is the direct application; §4.2, §5, §6.1 and the Scope-A table apply the
  weaker count-scope discipline as well.
* `docs/compound/2026-09-08-verify-go-stdlib-internals-claims-against-goroot-source.md` —
  verify a tool/API claim against the installed source, not against memory. Directly applied
  to `J-5` / `K-1` (skill mode inventory and dispatch behaviour) and `J-9` (index schema).
* `docs/compound/2026-05-07-backlogit-shipment-status-constraints.md` — shipment status enum,
  which is why CS2 deletes the `--status shipped` prescription.
* `docs/compound/2026-09-04-backlogit-sizing-is-wit-gated.md` — §14.
* The P-018 trio cited in §9.

### 13.6 Scope-T measurement — installed tooling (the `K-7` resolution)

**Rule: every assertion this plan makes about installed tooling appears below with the exact
command, the verbatim output and, where the artifact is a file, its digest.** No tooling claim
elsewhere in this plan is admissible unless it resolves here.

**Capture moment**: 2026-09-13, on branch `chore/stage-pipeline-policy-gap` at the tree that
carries revision 11, **before any `H1` edit**.

```powershell
# T-1..T-4  installed skill: identity, mode inventory, dispatch behaviour
git hash-object -- .github/skills/shipment-reconcile/SKILL.md
(Get-Content .github/skills/shipment-reconcile/SKILL.md).Count
Select-String -Path .github/skills/shipment-reconcile/SKILL.md -Pattern '^\#\#\# .*Mode'
Select-String -Path .github/skills/shipment-reconcile/SKILL.md `
  -Pattern ([regex]::Escape('* **SAFE_CLOSE selected** (default, including any classifier error,'))
Select-String -Path .github/skills/shipment-reconcile/SKILL.md -Pattern "records the classifier"
# T-5  engine identity
(Get-FileHash (Get-Command backlogit).Source -Algorithm SHA256).Hash
backlogit --version
# T-6  index schema: no path/location column (the J-9 claim)
#      NOTE: `backlogit query` rejects non-SELECT input ("Only SELECT statements are
#      permitted"), so PRAGMA is NOT available. The column set is read from a SELECT *.
backlogit query "SELECT * FROM items LIMIT 1;"
# T-7..T-8  installed intake contract and the read-only-mode precedent
(Get-Content .github/agents/_ship.agent.md)[252..261]
(Get-Content .github/skills/shipment-reconcile/SKILL.md)[244]
```

| ID | Tooling claim made by this plan | Where claimed | Measured value (verbatim) |
|---|---|---|---|
| **T-1** | The installed skill file is the one reviewed at revision 11's `H0` | §3.4.2, §9.2 ck 7 | `git hash-object` ⇒ `312f29b6e393bc9cffec411c386f2f8fc4454e50`; **1082 lines** |
| **T-2** | The installed skill exposes exactly four public **mode sections** — `pre`, `post`, `safe-close`, `detect-mixed-role` — and **no** classification entry point | §2, §3.4.1 | The command returns **five** `### …Mode` headings: L151 `### Mixed-Role Detection **Classification**` (a classification heading, **not** a mode section), then the four mode sections at L257 `### Pre-Mode`, L327 `### Post-Mode`, L354 `### Safe-Close Mode`, L839 `### Mixed-Role Detection Mode (…, READ-ONLY)`. `classify-close-path` ⇒ **×0** |
| **T-3** | On a non-`CASCADE` selection the skill **continues into the mutating** safe-close steps 1–10 | §2, §3.4.3 F, `K-1` | `* **SAFE_CLOSE selected** (default, including any classifier error,` ⇒ **×1 at L509**; L510 reads verbatim *"ambiguity, or unresolved precondition) → continue to step 1 below"* |
| **T-4** | The verdict is recorded **only** in the cascade report, **after** the destructive call | §2, `K-1` | Cascade Close Sub-Procedure begins L618; **L821** reads verbatim *"5. **Produce cascade-close report** recording the classifier's verdict,"* — step 5, after the `backlogit_ship_shipment` invocation |
| **T-5** | The resolved engine identity matches the §8.1 pin | §8.1, §9.2 ck 4 | `SHA256` ⇒ `1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98`; `backlogit version 1.10.1-0.20260823032255-b07729386a31` |
| **T-6** | The index `items` table carries **no** path/location column, so `backlogit_query_sql` cannot detect a torn duplicate | §5, AC-28 | `SELECT *` returns exactly: `artifact_type, assigned_to, branch, commit, created_at, custom_fields, dependencies, description, harness_status, hierarchy_path, id, items, labels, level, owner, parent_id, priority, references, severity, size, source_branch, sprint, status, title, updated_at` — **no path, no location, no directory column**. Hence **`backlogit doctor` only** |
| **T-7** | The installed intake check runs `mode: pre` with a **single** `expected_status` over every manifest item, so a mixed manifest necessarily yields `RECONCILE_FAIL` | §9.3, `K-6` | `_ship.agent.md` **L253–254** prescribe `mode: pre` + `expected_status: queued` *"(or `active` if already claimed)"*; **L259–262** state verbatim *"this single-`expected_status` check applies to true session-start intake, where every manifest task still shares one uniform status (all `queued` pre-claim, or all `active` immediately after this session's own claim…)"* |
| **T-8** | `mode: detect-mixed-role` is an existing **read-only, no-lock** public mode — the precedent Option A's new mode follows | §3.4.3 B/G | **L245** reads verbatim *"* **`mode: detect-mixed-role` is strictly READ-ONLY.** It NEVER mutates any"*, and its bullet continues *"requires NO `file-lock` acquisition because no backlog/shipment artifact is ever mutated"* |

**Every `T-n` above is reproducible by the commands in the block above, on the stated tree.**
A future claim about installed tooling that does not appear in this table is **not evidence**,
regardless of how confidently it is stated — that is the `K-1` lesson in one sentence.

## 14. Sizing

**Size `M` · Complexity `high`** (complexity is enum-validated **prose** — this workspace's
`header-def.yaml` defines no `complexity` field on the task type; see
`docs/compound/2026-09-04-backlogit-sizing-is-wit-gated.md`. Do **not** invoke
`backlogit update --complexity`; it fails with exit 1).

| Work item | Session | Est. |
|---|---|---|
| ADD-1 Role Boundary grant | S1 | 8 |
| ADD-2 Step 6.1(a1) incl. S0–S5 | S1 | 15 |
| CS1 reload + carve-out | S1 | 4 |
| CS2 close-sequence pointer | S1 | 4 |
| CS3–CS6 delegation block | S1 | 8 |
| CS7 intake pointer | S1 | 4 |
| CS8 pre-archive pointer | S1 | 4 |
| CS9 protected-set scoping | S1 | 3 |
| CS10 P-010 grant bullet | S1 | 3 |
| **CS11 mode enum + `classification_binding` input** | S1 | 2 |
| **CS12 `### Classify-Close-Path Mode` section** | S1 | 9 |
| **CS13 Step 0 binding revalidation + dispatch scoping** | S1 | 4 |
| **CS14 Behavioral Constraint bullet** | S1 | 2 |
| 24-row Go test (3 file fixtures) | S1 | 36 |
| **S1 implementation stretch** | | **106** (margin **14**) |
| Manifest/topology verification | S2 | 4 |
| H0/H1 cycle + gates | S1/S2 | 10 |
| **Total across both sessions** | | **120** |

**Why this is within the 2-hour rule.** The rule bounds a **single uninterrupted agent
stretch**, and this plan already mandates a **two-session split** (§9: S1 implements and
merges; S2 closes). The **S1 implementation stretch is 106 minutes with 14 minutes of
margin**; the remaining 14 minutes are S2 verification and gate work in a separate session
with a checkpoint between them (§9 step 8). Neither stretch exceeds 120 minutes.

**Named mid-task checkpoint boundary.** Within S1, the plan-mandated ordering (§7) places
`CS11–CS14` + rows 21–24 **first**. That point — fourth file applied, `H0` still red on the
`_ship.agent.md`/`workflow-policies.md` rows — is the designated resumption point if S1 is
interrupted. It is a **verifiable** boundary (`go test` output distinguishes it), not a
narrative one.

**Alternative considered and rejected: split `022.001-T` into two tasks.** Rejected because
any second task under `022-F` becomes a live descendant at depth 1, which §3.2.1's
set-equality rule forces into the `021-S` manifest. That changes the manifest from
`[022-F, 022.001-T]` to a three-member list, which **live-mismatches `PA-021-CASCADE`
condition 3** and, under §11, **voids the recorded operator approval** — §11 states the
authorization does not cover *"a re-scoped version of either action."* Re-scoping an approved
destructive action to buy 14 minutes of estimate margin is a strictly worse trade than the
two-session split the plan already performs, and Stage cannot self-authorize the replacement
approval. **The manifest is therefore unchanged at revision 11** (§13.1).

## 15. Constitution

| Principle | Assessment |
|---|---|
| **I. Safety-First Go** | One table-driven contract test. No `unsafe`, no goroutines, no I/O beyond reading two files; errors via `t.Errorf` |
| **II. Test-First (NON-NEGOTIABLE)** | H0 red → H1 green; the test lands before any `_ship.agent.md` or `workflow-policies.md` edit. **Deviation D1** |
| **III. Workspace Isolation** | Probe 25 ran in a disposable, gitignored, workspace-contained directory; `FIXTURE_ISOLATED_FROM_LIVE=True`, `LIVE_BACKLOG_UNMUTATED=True` measured |
| **IV. CLI Containment (NON-NEGOTIABLE)** | Every probe call `--cwd $ws`; engine resolved via the registered command, never an absolute path |
| **V. Structured Observability** | a1 emits `A1_NOT_APPLICABLE`, `A1_ALREADY_DONE`, `A1_MULTIPLE_FEATURE_MEMBERS`; Probe 25 emits one `KEY=VALUE` per criterion |
| **VI. Single Responsibility** | **Dependency discipline**: zero module dependencies added; stdlib-only assertions; `testify` absent from `go.mod`; reuses `repoRoot(t)`. The change **removes** a dependency direction — Ship stops depending on a local copy of the classification. **Deviation D2** |
| **VII. Destructive Approval (NON-NEGOTIABLE)** | The cascade is destructive and **live**. Gated by P-015, pre-mode `PROCEED`, the **read-only §9 step 14a classification boundary** whose non-`CASCADE` result halts **before any close call**, the safe-close binding revalidation, **and** the two §11 approved-action records with operator authorization. **The recovery path contains no destructive primitive** (§10.3.2 step 2 is quarantine-then-`git revert`), so no unapproved destructive operation exists anywhere in this plan |
| **VIII. Explicit Safety Modes** | **Deviation D3** |
| **IX. Git-Friendly Persistence** | Line-oriented Markdown/Go; the backlog mutation is committed on the closure branch, never on `main` |
| **X. Agent Context Efficiency** | Net **removal** of re-derived classification prose from Ship's hot path; this plan itself is reauthored rather than appended, removing ~180 lines of inline review history |
| **XI. Merge History (NON-NEGOTIABLE)** | Rollback is `git revert`; no rewrite, no force-push, no amend. **The reauthoring adds commits; it never amends or rewrites `3e0cf8f`, `3423581` or `e9019d4`** |

**D1 — 24 rows vs "fewer than 4".** These are **rows in one table-driven function** over
**three** files, sharing one fixture helper and ≤4 helpers; each is a substring or index
assertion, not an independent scenario. Collapsing them reduces falsifiability. *Assertion
density, not scope.*

**D2 — Width-Isolation: one task spans Markdown and Go.** `_ship.agent.md`,
`workflow-policies.md` and `shipment-reconcile/SKILL.md` are **not documentation** — they are
the agent's **executable instruction surface**, its **policy registry** and its **runtime
classifier**, read and acted on at run time. The Go test is a table of substring assertions
over those three files, derived mechanically from §6.1. They are **the contract and its
compiler-check**. Splitting them produces a task that cannot be verified and a task that
cannot compile green, breaking the mandatory H0-red → H1-green ordering, which requires all
parts in one atomic change. §14 records why a backlog-level split is additionally blocked by
the approved-action manifest binding.
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

The deferred B/C platform plans; Orchestrator repair; the pre-existing `SAFE_CLOSE`/exit-9
tool conflict; the P-015 detect-after-mutate **ordering** inside the **cascade sub-procedure**,
which is an inherited residual recorded as a follow-up against P-015 and the skill; and the
**`017-S` closure plan** (§11.1), which is future Stage work triggered by §9 step 19.

**`shipment-reconcile/SKILL.md` is IN scope at revision 11**, narrowly: exactly the four
**additive** sites CS11–CS14 (§4). Explicitly still out of scope within that file: any change
to `mode: pre`, `mode: post`, `mode: detect-mixed-role`, the Cascade Close Sub-Procedure, the
protected-set computation, the linked-deliberation snapshot extension, or the **unbound**
safe-close fall-through behaviour — all asserted preserved by rows 23b and 24b.

**Explicitly out of scope for the P-010 expansion**: any P-010 clause other than the single
`Ship MAY` bullet named by CS10; any change to P-010's `Ship MUST NOT`, `Stage MAY` or
`Stage MUST NOT` lists; and any other policy in `workflow-policies.md`.

## 17. Full history

Superseded revisions 1–10, plan-review attempts **1–6**, and all withdrawn language are
preserved in Git history, in PR #54, and in
`docs/archive/plans/2026-09-12-intercom-go-ship-feature-completion-foundation-plan.md`.
**No content was deleted from history; this file was reauthored, not rewritten in place.**
Where the archive and this plan disagree, **this plan governs**.

**Findings remain live across reauthoring.** Supersession, compaction and reauthoring all
transfer the contract *and* its outstanding findings. `H-1`…`H-9` were disposed at attempt 5;
`J-1`…`J-18` and `K-1`…`K-7` are disposed in §13.3 with landing sites. **Do not read "the
history is non-governing" as "the history's findings are closed."**




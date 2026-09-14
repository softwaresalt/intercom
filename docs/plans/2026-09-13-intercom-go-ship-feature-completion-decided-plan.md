---
title: "Decided Plan — Ship covering-feature completion and close-path delegation (022-F)"
date: 2026-09-13
status: planned
agent: Stage
revision: 12
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
> **Reauthored, not patched.** Revision 12 replaces revision 11 under explicit operator
> direction to reauthor rather than append. Lineage is preserved by Git commits and PR
> history; it is deliberately not restated inline.
>
> **This plan adopts OPTION A (prevention-first).** The operator selected, on
> 2026-09-13, adding a **read-only pre-mutation classification boundary** to the installed
> `shipment-reconcile` skill over the alternative of rewriting the safety model as
> detect-and-restore. The implementation surface is therefore **four files**, not three.
> See §2, §3.4.3 and §4 (**CS11–CS14**).
>
> **Revision 12 closes the contract across ALL of its surfaces, not only the producer.**
> Revision 11 added the read-only boundary to the skill and left the clauses that *consume*
> it, the policy that *authorizes* its outcome, the sink that must *refuse* an unbound
> cascade, and the probes that *measure* the tooling on the previous contract. That is the
> single defect class that lost six of seven gates; it is analysed and generalised in
> `docs/compound/workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md`.
> **§4.0 is the structural remedy**: one contract-surface matrix that must be closed before
> any site is claimed complete. Every clause site now has pinned replacement wording, and
> every test needle is a **verbatim substring of that pinned wording**.
>
> **GATE STATE — read this before acting on the plan.** The last plan-review gate that **RAN**
> is **true-lineage attempt 7**, and it **FAILED** (P0 = 0, P1 = 27 raw / 15 deduplicated).
> **Revision 12 is a PREPARED CANDIDATE for attempt 8. The attempt-8 gate has NOT been
> invoked, no attempt-8 verdict exists, and attempt 8 is NOT authorized.** The attempt counter
> is **not** reset by this reauthoring. `021-S` remains `queued` and **not claimable**, and
> `017-S` remains blocked. See §13 and §13.1. Revision 12 also carries one **open operator
> decision, `D-1`** (§3.4.3 J, §13.1).

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
| `.github/agents/_ship.agent.md` | 10 clause sites (**CS1–CS9**, **CS15**, all mandatory) + 2 additive (**ADD-1**, **ADD-2**) |
| `.github/policies/workflow-policies.md` | 1 clause site (**CS10**) — the P-010 `Ship MAY` grant |
| `.github/skills/shipment-reconcile/SKILL.md` | 4 clause sites (**CS11–CS14**) — the read-only pre-mutation classification boundary (§3.4.3) |
| `tests/integration/ship_feature_completion_contract_test.go` | **new** — 30-row table-driven contract test over the three instruction-surface files above |

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
   `CLASSIFICATION_BINDING` back. **Ship never invokes `mode: safe-close` without that
   binding** (CS6) — an unbound invocation by Ship is a contract violation and HALTs.
   Safe-close **revalidates** the binding against a freshly computed snapshot before any
   mutation (§3.4.3 (E)), so the verdict Ship acted on is the verdict the mutation executes
   under.

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
1. contract identity  := the 30-row test at {merge_sha} passes, including rows 21-24
                         (CS11-CS14) over .github/skills/shipment-reconcile/SKILL.md
2. surface identity   := `git diff --name-only {merge_base}..{merge_sha}` intersected with
                         the tracked tree equals exactly the four files of section 2
2b. review identity   := `git diff {merge_base}..{merge_sha} --
                         .github/skills/shipment-reconcile/SKILL.md` is byte-identical to the
                         diff recorded in the merged PR's reviewed head commit, AND the added
                         lines are confined to the four CS11-CS14 anchor blocks
3. working identity   := `git diff --quiet {merge_sha} -- .github/skills/shipment-reconcile/SKILL.md`
                         (the installed file is byte-identical to the reviewed merge commit)
```

A classifier that changed after the reviewed merge fails fact 3. **Fact 2b exists because
facts 1 and 2 alone do not close the in-merge case:** `SKILL.md` is *one of the four files*,
so fact 2 — a **file-set** check — is satisfied by **any** change inside it, and fact 1 only
asserts that the pinned needles resolve, not that nothing else was added. An unreviewed edit
made *within* the merge, in a region no needle covers, therefore passes both. Fact 2b bounds
the change to the reviewed diff and to the four anchor blocks, which is the only one of the
four facts that actually constrains *unpinned* content. **Any of the four failing ⇒ HALT**, no
close call, return to Stage.

#### 3.4.3 The read-only pre-mutation classification boundary (CS11–CS14, normative contract)

This is the **fourth file's** entire contract. It is stated as an executable, test-first
specification: every clause below is either pinned by a test row in §6.1 or measured in
§13.6.

**(A) New public mode — read-only.** `mode: classify-close-path` joins the mode enum.
Required input `shipment_id`. It accepts **no** `merge_commit_sha` and **no**
`expected_status`.

**(B) Classification runs before any mutation, and the mode cannot reach one.** The mode
executes the **existing** Safe-Close Step 0(a) manifest load, Step 0(b) snapshot and
Step 0(c) classification **with the single, exhaustively specified divergence recorded in
(H)** — the default-case token — and **stops at the dispatch**: it never takes the
`CASCADE selected →` or `SAFE_CLOSE selected →` branch. **The selecting predicates are
unchanged**; only the token emitted on the *default* branch differs. It **performs no
archive, no status transition and no record mutation**, acquires **no** single-writer lock
(there is nothing to serialise), and writes only its own additive report, exactly as
`mode: detect-mixed-role` already does.

**(B1) Exhaustive condition → verdict mapping (normative).** (B) and (H) are reconciled by
this table and by nothing else; there is no unmapped case and no residual "verbatim"
claim to interpret.

| Step 0(c) condition | Installed `safe-close` token | `classify-close-path` token | Same predicate? |
|---|---|---|---|
| Every feature member is a root, fully covered at every depth, set-equal to the manifest | `CASCADE` | `CASCADE` | **yes — unchanged** |
| A qualifying-but-partial feature member (genuine partial-feature shipment) | `SAFE_CLOSE` | `SAFE_CLOSE` | **yes — unchanged** |
| Manifest member is not a descendant of any qualifying feature | `SAFE_CLOSE` (default) | **`BLOCK`** (`MANIFEST_MEMBER_NOT_DESCENDANT`) | predicate same; **token differs** |
| Snapshot ambiguous / torn (queue **and** archive) | `SAFE_CLOSE` (default) | **`BLOCK`** (`SNAPSHOT_AMBIGUOUS`) | predicate same; **token differs** |
| Snapshot missing (neither location) | `SAFE_CLOSE` (default) | **`BLOCK`** (`SNAPSHOT_MISSING`) | predicate same; **token differs** |
| Descendant enumeration errored or incomplete | `SAFE_CLOSE` (default) | **`BLOCK`** (`ENUMERATION_INCOMPLETE`) | predicate same; **token differs** |
| Mixed qualification across feature members | `SAFE_CLOSE` (default) | **`BLOCK`** (`MIXED_QUALIFICATION`) | predicate same; **token differs** |
| Classifier error of any other kind | `SAFE_CLOSE` (default) | **`BLOCK`** (`CLASSIFIER_ERROR`) | predicate same; **token differs** |

**Every row whose token differs is a default-branch row, and every one becomes *more*
restrictive**, never less: a case the installed mode would have silently treated as a chosen
safe path becomes an explicit refusal. No row moves a case *into* `CASCADE`.

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
SHA-256 over a **canonical serialisation** of exactly the state the classification consumed
**plus the identity of the classification itself**. The serialisation is fully specified so
that any party reading the same tree reproduces the same digest byte-for-byte:

```text
line 1      : "v1"                                  # binding format version
line 2      : "shipment=" + shipment_id             # WHICH shipment was classified
line 3      : "verdict="  + CLOSE_PATH_VERDICT      # WHICH verdict is bound
line 4      : "reason="   + VERDICT_REASON
line 5      : "skill="    + sha256(SKILL.md bytes)  # WHICH classifier produced it
line 6      : "engine="   + backlogit --version     # WHICH engine resolved the records
line 7      : "manifest=" + join(",", sort(manifest item IDs))
line 8      : "deps="     + join(",", sort(shipment.dependencies))
line 9      : "status="   + shipment declared status
line 10..n  : one line per member of the combined pre-close snapshot Step 0(b)/(c) already
              builds - manifest tasks, qualifying feature members, and qualifying features'
              validated linked deliberations - rendered as
                 id + "\x1f" + artifact_type + "\x1f" + declared_status + "\x1f" +
                 (parent_id or "-") + "\x1f" + resolved_location
              with the lines sorted bytewise by id
final       : the lines joined with "\n", NO trailing newline; SHA-256 of the UTF-8 bytes
```

**`\x1f` (ASCII Unit Separator) is the intra-tuple delimiter** and is not a legal character in
any of the five fields, so no field value can forge a tuple boundary. **Nothing else enters
the digest.**

**Why `shipment_id`, the verdict and the two identities are inside the digest.** A binding
that covers only the record snapshot is satisfiable by a *different* shipment with an
identical member shape, and cannot distinguish *which* verdict was bound — so a matching
binding would not actually prove that the mutation executes under the classification Ship
read. Lines 2–6 close all four substitutions: wrong shipment, wrong verdict, a classifier
edited between classification and mutation, and an engine swapped underneath.

**Residual window, stated rather than argued away.** A binding proves the tree was
*identical at two moments*; it cannot prove nothing changed *between* them and reverted. The
window between §9 step 14a and step 15 is bounded by: both calls occurring in the same S2
session on the same closure branch, no mutation being prescribed between them (§9 has none),
and safe-close acquiring its single-writer lock **before** recomputing. The revalidation
therefore runs under the same lock that guards the mutation, which is the strongest guarantee
available without an engine-level transaction. **That remaining window is an inherited P-015
residual, disclosed here and charged to the follow-up recorded in §10.3, not to this plan.**

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

**(F) Backward compatibility — total for every existing caller, with one disclosed residual.**
When `classification_binding` is **absent**, `mode: safe-close` behaves **exactly as it does
today**, including its existing default fall-through to steps 1–10. No pre-existing public mode
is narrowed, renamed or removed; no existing input or output changes. Asserted negatively by
rows 23b and 24b.

**The residual, stated plainly:** because the unbound path is preserved unchanged, a caller
that invokes `mode: safe-close` **without** a binding on a fully-covered-root manifest can
still reach the destructive `CASCADE` branch, exactly as it can today. **This plan closes that
sink for the only caller it authorizes** — **CS6** forbids Ship from ever invoking
`mode: safe-close` unbound, and row 4b makes that falsifiable. For any *other* caller the
reachability is **pre-existing and unworsened by this plan**, which is the load-bearing
scope test recorded in
`docs/compound/2026-09-08-adversarial-review-empirical-verification-and-scope-check.md`.
It is therefore disclosed here and carried as decision **`D-1`** below — not silently fixed,
and not silently ignored.

**(J) `D-1` — OPERATOR DECISION REQUIRED (not resolved by this revision).**

> **Question:** should `mode: safe-close` **refuse** the `CASCADE` branch when no
> `classification_binding` was supplied (`RECONCILE_FAIL_CASCADE_UNBOUND`), making the
> read-only boundary **mandatory** at the destructive sink for *every* caller?
>
> **Why it is not decided here.** Doing so **narrows the behaviour of a pre-existing public
> mode**. §2 records the operator-approved Option-A scope as *"surgical and additive"*, and the
> attempt-7 Scope Boundary Auditor passed the fourth file on exactly that basis. Adopting the
> narrowing is a **scope change to an operator-selected option**, and Stage cannot
> self-authorize it.
>
> **Cost of `D-1 = yes`:** one additional clause at CS13, one additional test row, and a
> disclosed behaviour change for any unbound caller of `safe-close` anywhere in the workspace.
> **Cost of `D-1 = no` (the revision-12 position):** the sink stays reachable for non-Ship
> callers, unchanged from today, and Security's `SEC-11-03` remains **open as a disclosed
> residual** rather than closed.
>
> **This is a named, tracked blocker on the attempt-8 candidate**, not an omission. See §13.1.

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

## 4. Clause inventory — 15 sites (all mandatory) + 2 additive

### 4.0 Contract-surface matrix (NORMATIVE — the closure gate for this plan)

**No site in §4 may be claimed complete, and no acceptance criterion in §12 may be claimed
satisfied, while any row below carries an open edge.** This section exists because six of this
plan's seven gate failures shared one cause: a correction applied to the artifact where a
finding was reported, then declared closed by a measurement scoped to that same artifact,
while the consuming, authorizing, refusing, testing and probing surfaces of the same contract
stayed on the previous contract revision. See
`docs/compound/workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md`.

**The invariant this section enforces:**

> A cross-artifact contract change is complete only when the **producer**, **every consumer**,
> the **authority grant**, **every refusal/fallback path**, the **tests**, and the
> **measurement probes** are all updated and verified against the **same bound contract
> snapshot**. A local zero-occurrence result, or any check scoped only to the producer or only
> to the artifact where the finding was reported, **cannot prove closure** and must not be
> cited as closure evidence.

**Contract C-α — "a non-`CASCADE` close path is refusable before any close mutation."**

| # | Surface | Site | Current evidence (measured H0) | Intended change | Executable verification | Depends on |
|---|---|---|---|---|---|---|
| α1 | **producer / API** | CS11, CS12 (`SKILL.md`) | Modes are `pre`\|`post`\|`safe-close`\|`detect-mixed-role` (L56); no classification entry point (T-2) | §4.2 CS11, CS12 | rows 21a, 21b, 21c, 22a, 22b, 22c, 30 | — |
| α2 | **consumer** | CS3 (`_ship.agent.md` L794) | *"unless the P-015 **VERIFIED FULLY-COVERED-ROOT EXCEPTION** below applies"* | §4.2 **CS3** | row **25** | α1 |
| α3 | **consumer** | CS4 (L802–819) | 4 re-derived predicate sentences | §4.2 **CS4** | rows 7, 9a–9d | α1 |
| α4 | **consumer** | CS5 (L822) | *"in place of the safe-close sequence above for this shipment's closure"* | §4.2 **CS5** | row **26** | α1, α3 |
| α5 | **consumer** | CS6 (new bullet after CS4) | absent | §4.2 **CS6** | row **27** | α1, α3 |
| α6 | **consumer** | CS2 (L789–791) | `backlogit move <shipment_id> --status shipped` | §4.2 **CS2** | row 10 | α1 |
| α7 | **authority grant** | CS10 (`workflow-policies.md` L241) | *"Claim shipments, move tasks to active/done, close shipments, archive completed items"* | §4.2 **CS10** | rows 20a, **28** | α2–α5 |
| α8 | **authority grant** | ADD-1 (`_ship.agent.md` L38) | same bullet, Role Boundary row | §4.2 **ADD-1** | rows 1, 2a–2d | α7 (must match) |
| α9 | **refusal — bound non-`CASCADE`** | CS13 (`SKILL.md` L509) | `SAFE_CLOSE selected` is the unconditional default (T-3) | §4.2 **CS13** clause 1 | row 23a | α1 |
| α10 | **refusal — binding drift** | CS13 | absent | §4.2 **CS13** clause 1 | row 24a | α1 |
| α11 | **refusal — unbound at the destructive sink** | **CS6** (Ship consumer) | unbound `safe-close` can reach `CASCADE` (L505–507) | §4.2 **CS6** — Ship MUST NOT invoke `safe-close` unbound. Skill-side closure is **deferred to decision `D-1`** (§3.4.3 J) | rows **27**, 23b | α1, α5 |
| α12 | **refusal — unbound legacy safe path (PRESERVED)** | CS13 | `SAFE_CLOSE selected (default…)` ×1 | preserved **verbatim** | row 23b (preservation) | α9 |
| α13 | **refusal — mixed/ambiguous/torn** | CS12 (H) | Step 0(c) defaults these to `SAFE_CLOSE` | §4.2 **CS12** `BLOCK` clause | row 22c | α1 |
| α14 | **consumer — scoping** | CS9 (L792) | *"proves the protected set and halts fail-closed…"* — false on the cascade path | §4.2 **CS9** | row 19 | α3 |
| α15 | **tests** | §6.1 | 8 PRESENT needles satisfied by no pinned wording | every needle a **verbatim substring** of §4.2 | §6.1 derivation check | α1–α14 |
| α16 | **probes** | §13.6 | T-1…T-8 | **T-9…T-12** added | §13.6 commands | — |

**Contract C-β — "Ship may complete one covering feature `active -> done`."**

| # | Surface | Site | Current evidence (measured H0) | Intended change | Executable verification | Depends on |
|---|---|---|---|---|---|---|
| β1 | **producer / mechanism** | `backlogit move` / `backlogit_move_item` | gate broker may exit **6/7/8** (T-10) — unhandled by revision 11 | §5 S4 exit-code contract | §13.6 T-10 | — |
| β2 | **consumer** | ADD-2 / §5 a1 | absent | §4.2 **ADD-2** | rows 3, 12, 13, 16a, 16b | β1 |
| β3 | **authority grant** | ADD-1 + CS10 | task transitions only | §4.2 ADD-1 / CS10 | rows 1, 2a–2d, 20a | β2 |
| β4 | **refusal** | §5 S0, S2, S5 | absent | §4.2 **ADD-2** | rows 12, 16a, 16b | β2 |
| β5 | **containment probe** | `backlogit doctor` | named as sole mechanism, never probed | §13.6 **T-9** | §13.6 T-9 | — |
| β6 | **tests** | §6.1 | rows 3/12/16a needles unsatisfiable | needles drawn from §4.2 ADD-2 | §6.1 derivation check (rows 3, 12, 13, 16a, 16b) | β2 |

**Contract C-γ — "intake `mode: pre` is invoked only over a uniform manifest."**

| # | Surface | Site | Current evidence (measured H0) | Intended change | Executable verification | Depends on |
|---|---|---|---|---|---|---|
| γ1 | **producer** | installed `mode: pre` | single `expected_status` only (T-7) | **unchanged** | §13.6 T-7 | — |
| γ2 | **consumer — ordering** | **CS15** (`_ship.agent.md` Step 0.5 item 6) | intake runs at item **6**, **after** the claim at item **4** | §4.2 **CS15** — move intake **before** the claim | row **29** (anchor **and** index ordering) | γ1 |
| γ3 | **consumer — wording** | CS7 | queue-only literal | §4.2 CS7 | rows 14, 18 | γ1 |
| γ4 | **consumer — wording** | CS8 | queue-only literal | §4.2 CS8 | rows 15, 17 | γ1 |
| γ5 | **plan sequencing** | §9 step 1, §9.3 | prose only — **contradicted** the installed order | §9 step 1 now cites CS15 | `AC-37` | γ2 |

**Closure rule — mechanically checkable, not read for.** The matrix is closed when **all** of
the following hold. Each is a countable check, not a judgement:

1. every **producer** row (α1, β1, γ1) has **≥1 consumer** row depending on it;
2. every **consumer** row that acts on an outcome has an **authority** row permitting that
   action (α2–α6 → α7/α8; β2 → β3);
3. every value in the producer's value set — `CASCADE`, `SAFE_CLOSE`, `BLOCK`, *unbound*,
   *drifted* — has a **refusal or consumer** row (α9–α13);
4. every row has **≥1 test row** whose PRESENT needle resolves **inside that row's own block**
   and is a **verbatim substring** of that row's pinned §4.2 wording — **never borrowed** from
   a neighbouring site;
5. every claim about installed tooling has a **probe row** in §13.6 with command **and**
   verbatim output.

**Status at revision 12: CLOSED.** All 16 + 6 + 5 rows carry all four columns, and checks 1–5 were executed (not read for) against the final §6.1 table: **40 transition-row PRESENT needles, 40 verbatim substrings of §4.2, 0 failures**; 6 anchor/guard needles correctly excluded by their Provenance cell.
The per-site needle map in §6.1 is the machine-readable form of check 4.

### 4.1 Clause site inventory (apply by anchor text)

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
| **CS3** | `_ship.agent.md` | ``Do NOT call `backlogit shipment ship``` | Replace the named single exception with: do not call it unless **`mode: classify-close-path` has already returned `CLOSE_PATH_VERDICT: CASCADE`** and its binding is carried into the close call |
| **CS4** | `_ship.agent.md` | ``P-015 verified fully-covered-root exception (select the close path from the`` | **Delete the re-derived classification prose.** Replace with generic delegation to the **read-only pre-mutation boundary** per §3.4.1 |
| **CS5** | `_ship.agent.md` | ``in place of the safe-close sequence above for this shipment's`` | Remove the single-exception framing and replace it with the **bound** close-sequence pointer (subsumed by CS4/CS6) |
| **CS6** | `_ship.agent.md` | *(new bullet, inserted after CS4's replacement)* | **NEW** delegation bullet: Ship invokes whichever close path the **read-only** boundary's verdict names, never a path absent from that verdict, and **never invokes `mode: safe-close` without carrying the binding** |
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
| **CS15** | `_ship.agent.md` | ``6. **Intake reconciliation check**: Invoke `shipment-reconcile` with `mode: pre` and`` | **ORDERING.** Move the intake reconciliation check so it runs **before** the shipment claim, while the manifest is still uniformly `queued`. Revision 11 stated this ordering in plan prose only (§9 step 1) while the installed file kept the check at item **6**, after the claim at item **4** — so the plan prescribed a sequence the installed agent did not perform. **PRESERVE** the adjacent `Scope note (139-F/139.001-T)` and the orphan scan verbatim (§4.2 CS15) |

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

### 4.2 Pinned replacement wording (normative — the SOLE source of every PRESENT needle)

**Derivation rule (NON-NEGOTIABLE).** Every **transition-row** PRESENT needle in §6.1 is a
**verbatim substring of the wording pinned below**, character for character, including
backticks and emphasis markers. A needle authored from prose, from intent, or from a
neighbouring site's text is **invalid**. This rule is what makes §6.1 a proof of application
rather than a restatement of it, and its absence for `CS1`–`CS6` at revision 11 is what left
eight transition rows unsatisfiable by this plan's own normative text.

**Two needle kinds are outside this rule, and §6.1's Provenance column names which is which:**

| Kind | Source of truth | Provenance cell reads |
|---|---|---|
| **Transition needle** — text this plan *adds* | **§4.2 below** | the site name (`CS4`, `ADD-1`, …) |
| **Anchor needle** — pre-existing text used only to fix a **position** (rows 3, 29) | the installed file, measured at §13.6 | the site name + `(anchor)` |
| **Guard needle** — pre-existing text that must **survive** (rows 17, 18, 23b, 24b) | the installed file, measured at §13.6 | the site name + `(preserve)`, or `—` |

**An anchor or guard needle asserted against §4.2 would be a category error** — §4.2 pins what
changes, and those needles exist precisely because something must *not* change. They are
verified against the `H0` measurement in §13.6 instead, and the executable check below
enforces the split rather than papering over it:

```powershell
# §4.0 closure check 5 — run before claiming any revision complete.
# Every transition-row PRESENT needle must be a literal substring of §4.2.
# Anchor/guard rows are excluded BY THEIR PROVENANCE CELL, never by hand.
```

* **CS1** → *"Mandatory pre-self-close context reload: re-read the Role Boundary, and **also re-read `P-015` and verify the merge commit** before trusting any merged token. This reload is **scoped to Role Boundary changes only** and **A context reload MUST NOT widen the authority envelope** of the session performing it."*
* **CS2** → *"closes only the shipment record through **the authoritative close sequence defined by the `shipment-reconcile` skill**, never by a direct shipment status prescription."*
* **CS3** → *"**Do NOT call `backlogit shipment ship` / `backlogit_ship_shipment`** unless **`mode: classify-close-path` has already returned `CLOSE_PATH_VERDICT: CASCADE`** for this shipment. **HALT on any result token the installed classification did not name**, and HALT on `SAFE_CLOSE`, on `BLOCK`, and on any absent, empty, unparseable or ambiguous result."*
* **CS4** → *"Close-path selection is **delegated**: **the `shipment-reconcile` skill's read-only `mode: classify-close-path` boundary selects the close path**, and Ship performs **no classification of its own** — no root or `parent_id` test, no coverage test, no per-member rule and no fallback rule. Ship validates the returned token against the fixed allowlist `CASCADE` / `SAFE_CLOSE` / `BLOCK` and does nothing else with it."*
* **CS5** → *"Invoke the close path the boundary named, **carrying the returned `CLASSIFICATION_BINDING` into that call**, in place of any locally selected sequence for this shipment's closure."*
* **CS6** → *"Ship invokes **only** the close path the returned verdict names, and **never a path absent from that verdict**. **Ship MUST NOT invoke `mode: safe-close` without a `classification_binding`**; an unbound safe-close invocation by Ship is a contract violation and HALTs."*
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
  shipments, archive completed items"*, immediately followed, in the same row, by the
  **scoping clause** *"**scoped to the manifest of the shipment this session has claimed and
  whose live status is exactly `active`**"* and then the five conditions rendered so that the
  quantifier and the set-equality conjunct are both literal:
  *"(1) every manifest descendant is complete; (2) **no live descendant remains at every
  depth**; (3) topology is intact and **the descendant graph union the feature is set-equal to
  the manifest**; (4) containment holds; (5) the feature's live status is exactly `active`."*
* **CS10** → the P-010 `Ship MAY` bullet becomes *"Claim shipments, move tasks to active/done,
  **complete one covering feature `active -> done` when every live descendant at every depth is
  done and all five conditions below are conjunctive and fail-closed**, close shipments,
  archive completed items"* — and, **in the same bullet**, the scoping clause and the five
  conditions **rendered verbatim as in ADD-1**: *"**scoped to the manifest of the shipment this
  session has claimed and whose live status is exactly `active`**"*, then *"(1) every manifest
  descendant is complete; (2) **no live descendant remains at every depth**; (3) topology is
  intact and **the descendant graph union the feature is set-equal to the manifest**; (4)
  containment holds; (5) the feature's live status is exactly `active`."*

  > **CS10 is self-contained by construction.** At revision 11 the pinned CS10 wording ended
  > at *"all five conditions below"* while the conditions themselves existed only in
  > `_ship.agent.md`'s Role Boundary row — a cross-file dangling reference that resolves to
  > nothing inside the policy registry. The authority grant must be readable and checkable at
  > its **own** site (§4.0 check 2). Row **28** asserts the set-equality conjunct **and** the
  > scoping clause are present in `workflow-policies.md` itself; row **2d** asserts the
  > scoping clause in `_ship.agent.md`. The `021-S`/`017-S` release-instance IDs and the §3.3
  > 11-ID allowlist still **never** enter P-010 (rows 20b, 20c).
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
* **CS13** → immediately before the preserved dispatch bullets, insert **two** clauses.
  **Clause 1 (bound path):** *"**Bound-classification revalidation (before any mutation).**
  When `classification_binding` is supplied, recompute the binding from the Step 0(a)/(b)/(c)
  snapshot just taken. On mismatch, halt with `RECONCILE_FAIL_CLASSIFICATION_DRIFT`. On match,
  **HALT before any close mutation when the bound verdict is not `CASCADE`**, with
  `RECONCILE_FAIL_CLASSIFICATION_REFUSED`."*
  **Clause 2 — NOT PART OF THIS REVISION'S NORMATIVE CONTRACT.** Closing the unbound
  destructive sink *inside the skill* would narrow a pre-existing public mode and is therefore
  outside the "surgical and additive" Option-A scope the operator approved. It is recorded as
  **decision `D-1`** in §3.4.3 (J) and is **not** applied here. At revision 12 the sink is
  closed **for the only consumer this plan authorizes** — Ship — by **CS6**.
  The existing ``* **SAFE_CLOSE selected** (default, …`` bullet is **retained verbatim** and
  re-scoped by a leading clause *"When no `classification_binding` was supplied —"*.
* **CS15** → the Step 0.5 intake item is **relocated to run before the shipment claim**, and
  its opening sentence becomes *"**Intake reconciliation check — run BEFORE the claim.** Invoke
  `shipment-reconcile` with `mode: pre` and `expected_status: queued`, while every manifest
  member still shares the uniform pre-claim status."* The adjacent
  ``Scope note (139-F/139.001-T)`` paragraph and the ``scans for orphan items`` clause are
  **carried with the item unchanged**, and the claim item that previously preceded it is
  renumbered without any other edit.
* **ADD-2** (Step 6.1(a1), `_ship.agent.md`) → the step opens *"a1. **Covering-feature
  completion gate**"* and states, as literal text: *"S0 is an anomaly gate **evaluated before
  `n` is computed**; any anomaly **halts with no feature mutation**."*; *"S1: when `n == 0`,
  record `A1_NOT_APPLICABLE` and proceed — this is a success, not a skip."*; *"S2: when
  `n > 1`, record `A1_MULTIPLE_FEATURE_MEMBERS` and **halt with no feature mutation**."*;
  *"S3: when the single feature member already declares `done`, re-evaluate conditions 1–4 and
  record `A1_ALREADY_DONE` without re-issuing the transition."*; *"S4: when all five conditions
  hold, transition `active -> done`; on any unmet condition **HALT, naming the unmet
  condition**."*; and *"**A non-zero move exit code is a HALT**, never a retry: exit 6 blocked,
  exit 7 configuration, exit 8 retryable-but-not-retried-here."*
* **CS14** → *"**`mode: classify-close-path` is strictly READ-ONLY.** It NEVER archives, NEVER
  transitions any status, NEVER calls `backlogit_ship_shipment`, and requires NO `file-lock`
  acquisition because no backlog/shipment artifact is ever mutated — its classification report
  is an additive-only write to a non-backlog-state location. DEGRADED (backlogit unreachable)
  is REPORTED as `CLOSE_PATH_VERDICT: BLOCK` and the mode HALTS."*

### 4.3 Dated ordinal observation (NON-NORMATIVE)

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

**S4's mutation has a non-zero exit contract, and it is handled (measured — §13.6 `T-10`).**
`backlogit move` routes completions through a **gate broker** and documents four non-zero
outcomes: **exit 6** (blocked by a gate refusal), **exit 7** (configuration/setup error),
**exit 8** (retryable — lock contention or timeout), plus the ordinary non-zero failure. The
plan's revision-11 wording named the command and handled **none** of them.

| Exit | Meaning | a1 action |
|---|---|---|
| **0** | transition applied | Verify by re-reading the record; declared `status` must be exactly `done`. A zero exit with a non-`done` re-read ⇒ **HALT** |
| **6** | gate refusal | **HALT**, record `A1_MOVE_BLOCKED`, return to Stage. **Never** `--force-gates` — that flag is operator-only and §11's authorization does not cover it |
| **7** | configuration/setup error | **HALT**, record `A1_MOVE_CONFIG_ERROR`, return to Stage |
| **8** | retryable (lock/timeout) | **HALT**, record `A1_MOVE_BUSY`. **No retry inside a1** — a retry loop around the grant's only mutation is outside the one-transition scope of §3.2 |
| other | any other failure | **HALT**, record the exit code, return to Stage |

**`--force-gates` and `--force-reason` are forbidden on this route.** They are operator-only
overrides; §11's approval covers the cascade and nothing else, and §3.6 forbids a session from
widening its own authority envelope.

*(Measured caveat, recorded rather than assumed: `backlogit move --help` scopes the gate broker
to **"task/subtask completions"**, so a `feature` move may not invoke it at all. The table above
is therefore **fail-closed against a superset** of the outcomes that can actually occur — the
correct posture when the tool's own documentation does not state the feature case explicitly.)*

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

## 6. Go contract test — 30 rows

One table-driven function, ≤4 helpers, **stdlib assertions only** (`testing`, `strings`), no
`testify`, reusing the package-level `repoRoot(t)` from `tests/integration/build_script_test.go`.
**Every row names its own file** — there is no implicit default file — because the suite now
spans three instruction surfaces and a needle legitimately present in one must be assertable
absent in another. Rows 1–19 and 25–27, 29 read `.github/agents/_ship.agent.md`; rows 20 and 28
read `.github/policies/workflow-policies.md`; rows 21–24 and 30 read
`.github/skills/shipment-reconcile/SKILL.md`.

**Needle provenance rule (NON-NEGOTIABLE).** Every PRESENT needle in §6.1 is a **verbatim
substring of the pinned replacement wording in §4.2** — not of §3's prose, not of §5's tables,
and not of another site's text. The §6.1 **Provenance** column names the §4.2 bullet each
needle was cut from, so the derivation is checkable by inspection rather than trusted. At
revision 11, eight transition rows (3, 4, 5, 9d, 10, 11, 12, 16) pinned needles that **no**
pinned wording satisfied, which would have left them red at `H1` while the plan claimed a
red→green transition.

| # | File | Assertion | Kind |
|---|---|---|---|
| 1 | ship | Role Boundary row contains the narrow feature-completion grant | positive |
| 2 | ship | The grant names all five conditions, the **at-every-depth** quantifier, the **set-equality** conjunct and the **shipment-active scoping clause** | positive ×4 |
| 3 | ship | Step 6.1(a1) exists, positioned **after** `a0` and **before** `a` | ordering |
| 4 | ship | **(CS4)** the delegation block names the read-only `classify-close-path` boundary as the selector | positive |
| 5 | ship | CS1's reload clause also names P-015 and the merge commit | positive |
| 6 | ship | The authority-escalation rule is present | positive |
| 7 | ship | **(CS4)** the drifted `children`-only coverage rule is gone | negative |
| 8 | ship | `_ship.agent.md` does **not** name `TASK_ONLY_FINALIZE` | negative |
| 9 | ship | **(CS4)** No classification **predicate** — all four re-derived predicate sentences are **ABSENT** | negative ×4 |
| 10 | ship | **(CS2)** The `backlogit move <shipment_id> --status shipped` prescription is gone, replaced by the skill-defined close sequence | compound +/− |
| 11 | ship | CS1's carve-out is scoped to **Role Boundary** changes only | consistency |
| 12 | ship | a1 states the `n == 0` no-op (S1) **and** that an unmet guard halts (S4) | consistency |
| 13 | ship | a1 states idempotent resume after conditions 1–4 revalidate (S3) | consistency |
| 14 | ship | **(CS7)** delegation phrase **PRESENT** **and** old queue-only literal **ABSENT** | compound +/− |
| 15 | ship | **(CS8)** delegation phrase **PRESENT** **and** old queue-only literal **ABSENT** | compound +/− |
| 16 | ship | a1 states the anomaly-gate-before-`n` rule (S0) **and** that `n > 1` halts (S2) | compound ++ |
| 17 | ship | **(CS8)** sentence **PRESERVES** the single-writer lock **and** the orphan scan | preservation |
| 18 | ship | **(CS7/CS15)** adjacent `Scope note (139-F/139.001-T)` **survives** | preservation |
| 19 | ship | **(CS9)** the protected-set claim is **scoped to safe-close** **and** the unscoped literal is **ABSENT** | compound +/− |
| 20 | policy | **(CS10)** P-010's `Ship MAY` list carries the grant (20a), no `018.` literal (20b), no `021-`/`022-` literal (20c) | compound +/− |
| **21** | skill | **(CS11/CS12)** the mode enum carries `classify-close-path` **in the enum cell** (21a) **and** the section emits all three verdict tokens (21b) | compound ++ |
| **22** | skill | **(CS12)** the read-only clause (22a), the no-mutation clause (22b) **and** the fail-closed `BLOCK` clause (22c) are **PRESENT** | compound ++ |
| **23** | skill | **(CS13)** the bound-refusal HALT is **PRESENT** (23a) **and** the legacy unbound fall-through bullet **SURVIVES** (23b) | compound + / preservation |
| **24** | skill | **(CS14)** the drift token and the read-only Behavioral Constraint are **PRESENT** (24a) **and** the four pre-existing public mode names **SURVIVE** (24b) | compound + / preservation |
| **25** | ship | **(CS3)** the unknown-result-token HALT is **PRESENT** inside the `Do NOT call` bullet | positive |
| **26** | ship | **(CS5)** the binding is **carried into** the close call | positive |
| **27** | ship | **(CS6)** Ship **never** invokes `safe-close` unbound | positive |
| **28** | policy | **(CS10)** the five conditions **and** the scoping clause are present **in `workflow-policies.md` itself** — the grant is self-contained at its own site | compound ++ |
| **29** | ship | **(CS15)** intake reconciliation is ordered **before** the shipment claim | ordering |
| **30** | skill | **(CS12)** the fully-covered-root `CASCADE` emission is pinned against the **unchanged** Step 0(c) predicate | positive |

**Kinds**: `compound +/−` = {10, 14, 15, 19, 20} (**5**); `compound ++` = {16, 21, 22, 28};
`negative` = {7, 8, 9}; `preservation` = {17, 18, 23b, 24b}; `ordering` = {3, 29}.

### 6.1 Pinned literals — ALL 30 rows (normative; single-line, CRLF-safe)

**Every row is pinned, every needle names its file, every mandatory clause site has at
least one needle that resolves inside its own block, and every PRESENT needle names the §4.2
bullet it was cut from.** No row derives its needle from prose. Counts are measured at the
current tree (H0) over the file named in the row.

| Row | File | Site | MUST BE ABSENT | MUST BE PRESENT | Provenance (§4.2) | H0 |
|---|---|---|---|---|---|---|
| 1 | ship | ADD-1 | — | ``complete one covering feature `active -> done``` | ADD-1 | 0 → RED |
| 2a | ship | ADD-1 | — | ``all five conditions below are conjunctive and fail-closed`` | ADD-1 | 0 → RED |
| 2b | ship | ADD-1 | — | ``no live descendant remains at every depth`` | ADD-1 | 0 → RED |
| 2c | ship | ADD-1 | — | ``the descendant graph union the feature is set-equal to the manifest`` | ADD-1 | 0 → RED |
| 2d | ship | ADD-1 | — | ``scoped to the manifest of the shipment this session has claimed and whose live status is exactly `active``` | ADD-1 | 0 → RED |
| 3 | ship | ADD-2 | — | ``a1. **Covering-feature completion gate**`` (index **>** `a0. **TOPOLOGY_GATE: lifecycle`, **<** `a. **Pre-archive reconciliation gate`) | ADD-2 + **(anchor)** for the two bracketing literals | 0 → RED |
| 4 | ship | CS4 | — | ``the `shipment-reconcile` skill's read-only `mode: classify-close-path` boundary selects the close path`` | CS4 | 0 → RED |
| 5 | ship | CS1 | — | ``also re-read `P-015` and verify the merge commit`` | CS1 | 0 → RED |
| 6 | ship | CS1 | — | ``A context reload MUST NOT widen the authority envelope`` | CS1 | 0 → RED |
| 7 | ship | CS4 | ``it is fully covered (every one of its children,`` (×1) | — | — | 1 → RED |
| 8 | ship | guard | ``TASK_ONLY_FINALIZE`` — already ×0; **must stay ×0** | — | — | GREEN |
| 9a | ship | CS4 | ``permitted **only** when, for **every** feature member of the manifest: it is a`` (×1) | — | — | 1 → RED |
| 9b | ship | CS4 | ``enumerates to zero children, that childlessness is **positively verified**`` (×1) | — | — | 1 → RED |
| 9c | ship | CS4 | ``The manifest must contain nothing beyond the`` (×1) | — | — | 1 → RED |
| 9d | ship | CS4 | ``qualification is never per-member, and no feature ID is ever special-cased.`` (×1) | — | — | 1 → RED |
| 10 | ship | CS2 | ``<shipment_id> --status shipped`` (×1) | ``the authoritative close sequence defined by the `shipment-reconcile` skill`` | CS2 | 1 / 0 → RED |
| 11 | ship | CS1 | — | ``scoped to Role Boundary changes only`` | CS1 | 0 → RED |
| 12 | ship | ADD-2 | — | ``A1_NOT_APPLICABLE`` **and** ``HALT, naming the unmet condition`` | ADD-2 | 0 / 0 → RED |
| 13 | ship | ADD-2 | — | ``A1_ALREADY_DONE`` | ADD-2 | 0 → RED |
| 14 | ship | CS7 | ``every manifest item is present in `.backlogit/queue/` with the`` (×1) | ``defers every per-item status decision to that skill's `mode: pre` classification`` | CS7 | 1 / 0 → RED |
| 15 | ship | CS8 | ``queue with `status: done`, and scans for orphan items.`` (×1) | ``defers every per-item status decision to the `mode: pre` classification at `expected_status: done``` | CS8 | 1 / 0 → RED |
| 16a | ship | ADD-2 | — | ``evaluated before `n` is computed`` | ADD-2 | 0 → RED |
| 16b | ship | ADD-2 | — | ``A1_MULTIPLE_FEATURE_MEMBERS`` **and** ``halt with no feature mutation`` | ADD-2 | 0 / 0 → RED |
| 17 | ship | CS8 | — | ``scans for orphan items`` — already ×2 (L256, L777); **must stay ≥2** | **(preserve)** | GREEN |
| 18 | ship | CS7/CS15 | — | ``Scope note (139-F/139.001-T)`` — already ×1 (L259); **must stay ×1** | CS15 (preserve) | GREEN |
| 19 | ship | CS9 | ``proves the protected set and halts fail-closed on any cascade or provenance`` (×1) | ``proves the protected set on the safe-close path`` **and** ``no protected set by construction`` | CS9 | 1 / 0,0 → RED |
| 20a | policy | CS10 | — | ``complete one covering feature `active -> done``` in **`workflow-policies.md`** | CS10 | 0 → RED |
| 20b | policy | guard | ``018.`` in **`workflow-policies.md`** — already ×0; **must stay ×0** | — | — | GREEN |
| 20c | policy | guard | ``021-`` and ``022-`` in **`workflow-policies.md`** — already ×0; **must stay ×0** (no release-instance ID enters P-010) | — | — | GREEN |
| **21a** | skill | CS11 | — | ``digest returned by `mode: classify-close-path``` — the **new `classification_binding` input row**, which is CS11-specific and cannot be satisfied by the enum cell alone | CS11 | 0 → RED |
| **21b** | skill | CS12 | — | ``CLOSE_PATH_VERDICT: CASCADE`` **and** ``CLOSE_PATH_VERDICT: SAFE_CLOSE`` **and** ``CLOSE_PATH_VERDICT: BLOCK`` | CS12 | 0 / 0 / 0 → RED |
| **21c** | skill | CS11 | — | ``classify-close-path`` ×**≥2** in `SKILL.md` (the mode-enum cell **and** the input row); at H0 it is ×0, so this row is red and cannot be satisfied by a single mention | CS11 **(count)** | 0 → RED |
| **22a** | skill | CS12 | — | ``### Classify-Close-Path Mode`` | CS12 | 0 → RED |
| **22b** | skill | CS12 | — | ``performs no archive, no status transition and no record mutation`` | CS12 | 0 → RED |
| **22c** | skill | CS12 | — | ``every unsupported, mixed, ambiguous, torn, missing or incompletely enumerated case returns `CLOSE_PATH_VERDICT: BLOCK``` | CS12 | 0 → RED |
| **23a** | skill | CS13 | — | ``HALT before any close mutation when the bound verdict is not `CASCADE``` **and** ``RECONCILE_FAIL_CLASSIFICATION_REFUSED`` | CS13 | 0 / 0 → RED |
| **23b** | skill | CS13 | — | ``* **SAFE_CLOSE selected** (default, including any classifier error,`` — already ×1; **must stay ×1** (the legacy unbound path is preserved verbatim) | CS13 **(preserve)** | GREEN |
| **24a** | skill | CS14 | — | ``RECONCILE_FAIL_CLASSIFICATION_DRIFT`` **and** ``**`mode: classify-close-path` is strictly READ-ONLY.**`` | CS14 | 0 / 0 → RED |
| **24b** | skill | guard | — | ``mode: detect-mixed-role`` ×≥1, ``mode: safe-close`` ×≥1, ``mode: pre`` ×≥1, ``mode: post`` ×≥1 — **all four pre-existing public modes must survive** | **(preserve)** | GREEN |
| **25** | ship | **CS3** | — | ``HALT on any result token the installed classification did not name`` | **CS3** | 0 → RED |
| **26** | ship | **CS5** | ``in place of the safe-close sequence above for this shipment's`` (×1) | ``carrying the returned `CLASSIFICATION_BINDING` into that call`` | **CS5** | 1 / 0 → RED |
| **27** | ship | **CS6** | — | ``Ship MUST NOT invoke `mode: safe-close` without a `classification_binding``` | **CS6** | 0 → RED |
| **28** | policy | **CS10** | — | ``the descendant graph union the feature is set-equal to the manifest`` **and** ``scoped to the manifest of the shipment this session has claimed and whose live status is exactly `active``` — both **in `workflow-policies.md`** | **CS10** | 0 / 0 → RED |
| **29** | ship | **CS15** | — | ``Intake reconciliation check — run BEFORE the claim`` **and** index(``Intake reconciliation check — run BEFORE the claim``) **<** index(``Record `shipment_id` as the session scope``) | **CS15** + **(anchor)** for the claim literal | 0 → RED |
| **30** | skill | CS12 | — | ``CLOSE_PATH_VERDICT: CASCADE`` appears within the ``### Classify-Close-Path Mode`` section block (index-bounded, not whole-file) | CS12 | 0 → RED |
| — | ship | guard | ``018.`` in `_ship.agent.md` — already ×0; **must stay ×0** | — | — | GREEN |

**Independent falsifiability — every mandatory site owns at least one needle that no other
site can satisfy.** No site is credited by an edit to a different site, and **no needle is
borrowed**:

| Site | Own needle(s) | Located in | Borrowed? |
|---|---|---|---|
| CS1 | rows 5, 6, 11 | the reload clause | no |
| CS2 | row 10 | `<shipment_id> --status shipped` | no |
| **CS3** | **row 25** | the `Do NOT call backlogit shipment ship` bullet | **no** — was row 9d (a CS4 row) at rev 11 |
| CS4 | rows 4, 7, 9a–9d | the delegation block and its four survivor sentences | no |
| **CS5** | **row 26** | the close-sequence pointer | **no** — was row 10 (a CS2 row) at rev 11 |
| **CS6** | **row 27** | the new delegation bullet | no |
| CS7 | rows 14, 18 | the intake clause | no |
| CS8 | rows 15, 17 | the pre-archive clause | no |
| CS9 | row 19 | the protected-set bullet | no |
| CS10 | rows 20a, 28 | `workflow-policies.md` | no |
| CS11 | rows 21a, 21c | the Inputs `classification_binding` row and the mode enum cell | no |
| CS12 | rows 21b, 22a, 22b, 22c, 30 | the new section | no |
| CS13 | rows 23a, 23b | the Step 0 dispatch | no |
| CS14 | rows 24a, 24b | Behavioral Constraints | no |
| **CS15** | **row 29** | Step 0.5 intake ordering | no |
| ADD-1 | rows 1, 2a–2d | the Role Boundary row | no |
| ADD-2 | rows 3, 12, 13, 16a, 16b | Step 6.1(a1) | no |

**An `H1` that edits only the CS4 block now leaves rows 4… no — leaves rows 10, 19, 20a, 21–24
and 25–30 red.** Rows 7 and 9a–9d additionally make the *"no classification predicate"* claim
directly falsifiable: all four re-derived predicate sentences measured in the CS4 block are
pinned ABSENT, so a partial deletion fails the suite. Those same four sentences remain
legitimately **present** in `SKILL.md`, which is why every row names its file.

**Two classes, and only one is red → green:**

| Class | Rows | H0 state |
|---|---|---|
| **Transition rows** — must change | 1, 2a–2d, 3–7, 9a–9d, 10–16b, 19, 20a, 21a, 21b, 21c, 22a–22c, 23a, 24a, 25–30 (**36 assertions**) | **RED.** None can be vacuously green |
| **Guard rows** — must NOT change | 8, 17, 18, 20b, 20c, 23b, 24b, and the `_ship.agent.md` `018.` guard (**8**) | **GREEN today, by design.** They assert a property that already holds and must survive |

**The guard rows are green at H0 and that is correct** — their purpose is to fail if the
implementation *breaks* something, not to record a transition. Row 3's `a1` anchor is absent
pre-implementation **by design**; that absence is the H0 readiness guard. Rows 23b and 24b are
the **backward-compatibility guards for Option A**: they fail if the fourth-file edit narrows
or removes any pre-existing public mode or the legacy unbound fall-through (§3.4.3 F).

**Rows 23b/24b are substring-survival guards, and that is their stated limit.** They prove the
mode *names* and the legacy dispatch *bullet* survive; they do **not** prove the surrounding
semantics are unchanged. Semantic backward compatibility is carried instead by §3.4.3 (F)'s
normative statement and by §3.4.3 (B1)'s exhaustive mapping, which shows every divergent case
is a **default-branch token change only**, never a predicate change and never a case moving
*into* `CASCADE`. Claiming more from a substring guard than it can deliver is the defect class
this plan exists to stop.

**Only the transition rows constitute the red → green evidence** for AC-3/AC-8/AC-24/AC-33.
Do not read the guard rows as proof the contract changed.

**Independently re-measured at revision 12** (commands in §13.4 and §13.6): every ABSENT
needle above resolves ×1 and every PRESENT needle ×0 in the live tree; `TASK_ONLY_FINALIZE`
×0; `018.` ×0 in both `_ship.agent.md` and `workflow-policies.md`; `021-` and `022-` ×0 in
`workflow-policies.md`; `scans for orphan items` ×2; `Scope note (139-F/139.001-T)` ×1;
`record-consistent` ×0 in `_ship.agent.md` (it is **added**, not preserved);
`classify-close-path` ×0 and `CLOSE_PATH_VERDICT` ×0 in **all three** instruction surfaces;
`* **SAFE_CLOSE selected** (default, including any classifier error,` ×1 in `SKILL.md`;
`Intake reconciliation check` ×1 in `_ship.agent.md`, currently at item 6, **after** the
claim at item 4 (the CS15 defect, measured). `repoRoot(t)` exists at package scope in
`tests/integration/build_script_test.go`; `testify` is absent from `go.mod`.

> **What this measurement does and does not prove (§4.0).** It is an `H0` state capture over
> Scope I. It proves each needle's pre-change count, and **nothing about `H1`**. The claim
> that each PRESENT needle becomes satisfiable at `H1` rests on the **Provenance** column
> above — each needle being a verbatim substring of §4.2 — not on this count. At revision 11
> this distinction was collapsed, and an identical-looking measurement certified eight rows
> that no pinned wording could satisfy.

## 7. Ordering

**Test-first, `main` stays green.** H0: write the 30-row test → **red**. H1: apply ADD-1,
ADD-2 and **CS1–CS15** → **green**. All fifteen clause sites are required to reach H1.

**Within H1, the fourth file lands first.** `CS11–CS14` (the classification boundary) are
applied before `CS3–CS6` (the delegation that consumes it), so `_ship.agent.md` never names a
mode the installed skill does not expose — the exact defect class `J-5`/`K-1` recorded twice.
This is also the **S1a/S1b session boundary** (§14), not merely a mid-task checkpoint.

**Then the consumers, then the authority, then the ordering.** The §4.0 dependency column
gives the total order for the rest of H1: α1 (producer) → α2–α6, α14 (consumers) → α7, α8
(authority) → α9–α13 (refusal) → γ2 (CS15 ordering) → α15 (tests green). **No surface may be
marked complete before the surfaces it depends on**, and §4.0's closure rule is re-evaluated
at the end of H1 rather than assumed from the earlier check.

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
| 1 | **S1** | **Step 0.5 Shipment Intake.** Verify on `main`; **P-011** branch-before-mutation + **P-016** single-worktree check; pre-claim topology gate; **intake `mode: pre` with `expected_status: queued` — run BEFORE the claim, while the manifest is uniformly `queued` and therefore representable (§9.3)**; then **claim `021-S`**. **This ordering is enforced by the `CS15` clause edit (§4.1/§4.2), not by this row.** The installed `_ship.agent.md` runs intake at Step 0.5 item **6**, *after* the claim at item **4**; a plan row alone cannot change what Ship executes, so `CS15` physically relocates the intake bullet above the claim bullet and row 29 of §6.1 asserts the resulting index ordering |
| **1a** | S1 | **POST-CLAIM CONDITION — `022-F` must be `active` (checked, not assumed).** Re-read `022-F` **by a single `backlogit_get_item` record read** and require its live status to be **exactly `active`**. **This is NOT a manifest-wide `mode: pre` invocation** — see §9.3. **Any other value ⇒ HALT** and return to Stage. **Ship is granted NO authority to transition a feature `queued -> active`** — §3.2 grants exactly one transition, `active -> done`. *(Live state at planning time is `queued`; the shipment claim is what is expected to activate it. If it does not, that is a planning-state error only Stage can reconcile.)* |
| **1b** | S1 | **Step 2 Harness Generation (P-002 / P-004).** The **harness-architect** produces the 30-row harness: `go vet ./...` exits 0, `go test ./...` exits non-zero with expected failure markers ⇒ **H0 RED CONFIRMED** ⇒ `harness-ready` applied to `022.001-T`. **This precedes the TASK claim** — P-002 permits claiming a task only after red-phase confirmation |
| **1c** | S1 | **Step 3 Build Ready Queue** — filtered to tasks carrying `harness-ready` |
| **1d** | S1 | **Step 4.1 Claim Task** — `022.001-T -> active` |
| 2 | S1 | **Step 4.2 implement** → apply **CS11–CS14 first** (§7, end of **S1a**), then ADD-1, ADD-2, **CS1–CS10 and CS15** in the §4.0 dependency order → **H1 green** (**S1b**) |
| 3 | S1 | Quality gates; review gate |
| 4 | S1 | Complete Task — commit the implementation, then **`022.001-T -> done`**, then **commit the resulting `.backlogit/` change** and verify a clean worktree. **Both commits must be in the PR that merges at step 5** — otherwise merged `main` still shows the task `queued` and a1 halts at step 13 |
| 5 | S1 | PR lifecycle — gates, build, push, PR, **P-018 Copilot engagement**, operator approval, **merge** |
| 6 | S1 | Merge Confirmation Gate — `gh pr view` `MERGED`; `git fetch origin main`; `git merge-base --is-ancestor {merge_sha} origin/main`; **and `git rev-list --parents -n 1 {merge_sha}` returns three fields (true merge commit).** A single-parent result means the PR was squashed or rebased, which invalidates §10.4's `git revert -m 1` rollback path ⇒ **HALT** (§15 Principle XI, `M-13`) |
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
| 17 | S2 | Sync — **MCP `backlogit_sync_index` first**, CLI `backlogit sync` as declared fallback — then push, closure PR, **P-018 Copilot engagement**, local review, operator approval, **merge**. **Verify the merge strategy here too** — `git rev-list --parents -n 1` must return three fields — for the same §10.4 reason as step 6 (`M-13`) |
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
   merge commit confirmed at step 6: (a) the 30-row contract test passes at `{merge_sha}`,
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
* **The pre-claim ordering is carried by `CS15`, not by §9's prose.** At revision 11 this
  section and §9 step 1 both *described* intake-before-claim while the installed file kept
  intake at item 6 and the claim at item 4 — the plan asserted an ordering no clause site
  established (`M-12`). `CS15` is the missing producer edit; §6.1 row 29 is its probe.

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

**`git revert -m 1 {merge_sha}`** — the mainline-parent form, because both merges in this plan
(§9 steps 5 and 17) produce **merge commits**, and `git revert` refuses a commit with more than
one parent unless `-m` names the mainline. A bare `git revert {merge_sha}` **fails** on a merge
commit; prescribing it would leave this plan with a rollback path that does not execute
(**`M-13`**).

**Preconditions, verified before the revert is attempted:**

1. `git rev-list --parents -n 1 {merge_sha}` returns **three** fields (commit + two parents),
   confirming a true merge commit. **One parent ⇒ the PR was squashed or rebased**, `-m 1` is
   invalid, and this rollback path does not apply — HALT to the operator rather than
   improvising a different revert. §15 Principle XI requires this check at **both** merges,
   which is why the failure is detected at merge time and not first discovered here.
2. `-m 1` selects the **first** parent, which for a GitHub merge commit is the base branch
   (`main`). Reverting relative to parent 1 therefore removes the merged topic-branch content.

**Never a history rewrite, never a force-push, never an amend.** `017-S` must be held first.

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
7. **`PA-021-CASCADE` only** — the validated linked-deliberation set for the `021-S` manifest
   is **empty**, verified by measurement (§13.6 **T-11**) rather than assumed. This condition
   exists because `mode: safe-close`'s Cascade Close Sub-Procedure also archives linked
   deliberation artifacts; if that set were non-empty the cascade would touch records outside
   the manifest and condition 6 would be violated. **Measured at revision 12: `backlogit link
   list` returns `[]` for all three manifest members, the engine's linked-deliberation regex
   matches zero times across `022-F`, `022.001-T` and `018-F`, no record carries
   `source_deliberation_id`, and the index holds zero `artifact_type='deliberation'` rows.**
   A non-empty set at invocation time **voids the approval and HALTs** — it does **not**
   license a re-scope, because a re-scoped action is explicitly outside this approval.

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
3. **AC-3** All **fifteen** clause sites CS1–CS15 are applied; **every one of them has at
   least one pinned needle in §6.1 that resolves inside its own block, and no site is credited
   by a needle belonging to another site** (CS1 → rows 5/6/11; CS2 → row 10; **CS3 → row 25**;
   CS4 → rows 4/7/9a–9d; **CS5 → row 26**; **CS6 → row 27**; CS7 → rows 14/18;
   CS8 → rows 15/17; CS9 → row 19; CS10 → rows 20a/**28**; CS11 → rows 21a/**21c**;
   CS12 → rows 21b/22a/22b/**22c**/**30**; CS13 → rows 23a/23b; CS14 → rows 24a/24b;
   **CS15 → row 29**); the **30-row** test passes; H0 was red. **Every PRESENT needle is a
   verbatim substring of the §4.2 pinned wording for its own site** (the §6.1 Provenance
   column), so every transition row is satisfiable by applying §4.2 alone.
4. **AC-4** `_ship.agent.md` states **no classification predicate** — all four re-derived
   predicate sentences measured in the CS4 block are **ABSENT** (rows 7, 9a, 9b, 9c) — and
   contains **no `018.` literal**.
5. **AC-5** The close-verdict allowlist is exactly `CASCADE` / `SAFE_CLOSE` / `BLOCK`, where
   `BLOCK` is a **refusal and not a close path**; `HALT` and `RECONCILE_FAIL` are non-close
   outcomes; unknown tokens, absent, empty, ambiguous or unreadable results **HALT**. **No
   runtime `policy_id`, version or source-disagreement check exists.** The mapping from every
   classifier input condition to its verdict is **exhaustive and total** (§3.4.3 B1): every
   unsupported, mixed, ambiguous, torn, missing or incompletely enumerated case returns
   `BLOCK`, and **no case is described as "verbatim today's behaviour" while the table assigns
   it a different token** — the two statements are reconciled, not merely adjacent
   (row 22c).
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
27. **AC-27** **(J-5, K-1, M-1)** The delegation target is the installed **read-only
    `mode: classify-close-path`** boundary added by CS11–CS14; the verdict is **returned by
    that call**, not scraped from a post-mutation report. **No plan text requires an API the
    post-`H1` skill does not expose**, and no Ship runtime code is added. **Every one of the
    four consumer clause sites CS3–CS6 names that boundary** — none still routes through
    `mode: safe-close` Step 0(c)'s post-mutation verdict — and each carries its own pinned
    replacement wording in §4.2 (rows 25, 4, 26, 27).
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
     `CLASSIFICATION_BINDING`. Rows 21a, 21b, 21c, 22a pass.
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
    and mutation therefore execute against the same bound snapshot or not at all. **The
    serialisation is fully specified** — field order, the `\x1f` intra-tuple delimiter, the
    `shipment_id`, the verdict, the reason, the skill digest and the engine version — so two
    independent implementations produce the same digest, and the **residual drift window**
    (between recomputation and the first mutation) is disclosed rather than denied.
39. **AC-39** **(M-15, the closure invariant)** **§4.0's contract-surface matrix is CLOSED**:
    for each of the three contracts C-α, C-β and C-γ every row carries current evidence, an
    intended change, an executable verification and its dependency; and all five mechanical
    closure checks pass — (1) every producer row has ≥1 consumer row, (2) every consumer row
    names a producer row, (3) every contract has ≥1 authority row and ≥1 refusal row, (4)
    every row's executable verification resolves to a numbered §6.1 row or a §13.6 probe, and
    (5) no row depends on a surface absent from the matrix. **A revision that adds a clause
    site without adding its matrix row and its §6.1 row is incomplete by construction.**
40. **AC-40** **(M-12)** `CS15` relocates the Step 0.5 intake bullet **above** the claim bullet
    in `_ship.agent.md`; row 29 asserts both the new anchor text and the index ordering
    `index(intake) < index(claim)`. **No part of the pre-claim ordering rests on plan prose
    alone.**
41. **AC-41** **(M-6)** `workflow-policies.md`'s P-010 entry renders **all five conditions
    inline** and is **self-contained**: it contains no unresolvable pointer such as "the five
    conditions below", and its scoping clause limits the grant to the manifest of the
    shipment the session has claimed and whose live status is exactly `active` (rows 2d, 28).
42. **AC-42** **(M-5, D-1)** The **in-scope** half of the unbound-sink defect is closed: CS6
    forbids Ship from invoking `mode: safe-close` without a `classification_binding`
    (row 27). The **skill-side** narrowing — making an unbound `safe-close` refuse rather than
    fall through — is recorded in §3.4.3 (J) as operator decision **`D-1`**, **explicitly
    NOT adopted**, because it would narrow a pre-existing public mode and contradict §2's
    operator-approved additive scope. **This plan does not claim the sink is closed at the
    skill.**
43. **AC-43** **(M-16, Scope T)** §5's S4 handles the **full `backlogit move` exit-code
    contract** measured at §13.6 T-10 — `0` proceed, `6` `A1_MOVE_BLOCKED`, `7`
    `A1_MOVE_CONFIG_ERROR`, `8` `A1_MOVE_BUSY`, any other value HALT — and **never** passes
    `--force-gates` / `--force-reason`. Where the broker's applicability to `artifact_type:
    feature` is undocumented, the plan fail-closes against the **superset**.
44. **AC-44** **(M-7)** The work is sized as **three sessions S1a / S1b / S2** (§14), each
    within the 2-hour rule, with **no change to the `021-S` manifest** and therefore no
    disturbance of `PA-021-CASCADE` condition 3. **No second task is created**, because a
    second task would become a depth-1 live descendant of `022-F` and break the recorded
    set-equality.

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

next-attempt: 8
next-attempt-state: CANDIDATE PREPARED — NOT RUN
next-attempt-authorization: NOT GRANTED
```

> **Read the two blocks above as one fact: the last gate that RAN is attempt 7 and it
> FAILED.** Revision 12 is a **prepared candidate** for attempt 8. **No attempt-8 gate has
> been invoked, no attempt-8 verdict exists, and no attempt-8 authorization has been
> granted.** The counter is **not** reset and revision 12 does **not** carry a PASS, a
> PENDING or any other marker implying a gate ran.
>
> **Revision 11 resolved `K-1`…`K-7`** under explicit operator selection of **Option A**, but
> closed only the **producer** surface; attempt 7 then returned 27 raw / 15 deduplicated P1
> findings, all on the consumer, authority, refusal, test and measurement surfaces that
> revision 11 left stale. **Revision 12 is the root remediation of that defect class**, not a
> further patch: see §4.0 and
> `docs/compound/workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md`.
>
> The attempt-6 verdict, its per-persona counts and the `K-1`…`K-7` blocker set remain
> recorded in `docs/reviews/2026-09-13-true-lineage-attempt-6-verdict.md`; the attempt-7
> verdict is recorded in `docs/reviews/2026-09-13-true-lineage-attempt-7-verdict.md`.

### 13.1 Current state

| Field | Value |
|---|---|
| **Plan `status:`** | `planned` |
| **Revision** | **12** — reauthored, not patched |
| **True-lineage attempt (last RUN)** | **7** — the counter continues across every reauthoring and is **not** reset |
| **Gate decision (last RUN)** | **FAIL** — P0 = 0, P1 = 27 raw / 15 deduplicated, 7/7 personas. See `docs/reviews/2026-09-13-true-lineage-attempt-7-verdict.md` |
| **Attempt 8** | **CANDIDATE PREPARED, NOT RUN.** Revision 12 is review-ready in Stage's judgement; **the gate has not been invoked and is not authorized** |
| **Harvest-ready** | **NO** — the rule requires P0 = 0 **and** P1 = 0 from a gate that actually ran. The last run had P1 = 27. No harvest and no shipment assembly were performed |
| **`021-S`** | `queued` — **not claimable**. Unchanged by revision 12 |
| **`017-S`** | `queued`, dependency `[021-S]` — ineligible until `021-S` ships, **and** separately ineligible until an `017-S` closure plan exists (§11.1) |
| **`PA-021-CASCADE`** | recorded, **unexercised**; **manifest unchanged at `[022-F, 022.001-T]`**, so the approval is **not** re-scoped by revision 12 and needs no re-authorization. §11 condition 7 is satisfied **by measurement** (§13.6 T-11), which is why `M-11` required no re-scope |
| **`PA-017-CASCADE`** | recorded, **unexercised**, **held in escrow** (§11.1) |
| **Open blocker `D-1`** | **OPEN — operator decision required.** Whether `mode: safe-close` invoked **without** a `classification_binding` should refuse instead of falling through. Narrowing it would contradict §2's operator-approved additive Option-A scope, so this plan does **not** adopt it (§3.4.3 J, AC-42). Ship-side exposure is closed by CS6; the skill-side sink remains reachable by any **other** caller |
| **Authorization (revision 12)** | Operator instruction of 2026-09-13/14 authorizing a **compound-learning and root-cause pass before attempt 8**, the durable learning under `docs/compound/workflow-issues/`, root remediation of this plan and the coupled backlog records, and any surgical planning-surface expansion needed to describe Option A coherently. **It explicitly does NOT authorize spending true-lineage attempt 8.** |

**Lineage (compact — narrative deliberately not restated inline):**

| Attempt | Revision | Verdict |
|---|---|---|
| 1 | 4 | FAIL |
| 2 | 5 | FAIL |
| 3 | 6 | FAIL (re-entry budget exhausted, escalated) |
| 4 | 7 | FAIL (`H-1`…`H-9`) |
| 5 | 9 | FAIL (`J-1`…`J-18`) |
| 6 | 10 | FAIL (`K-1` P0 + `K-2`…`K-7` P1) |
| 7 | 11 | FAIL — P0 = **0** (K-1 P0 closed; K-4, K-5 closed), P1 = 27 raw / 15 deduped (`M-1`…`M-15`). Option A applied to the skill side only; CS3–CS6, CS10, the unbound sink, the tests and the measurement probes were not brought along |
| **8** | **12** | **NOT RUN.** Candidate prepared under the §4.0 closure invariant. Awaiting explicit operator authorization |

> The counter reflects **plan-review lineage, which follows the plan, not the filename or the
> revision number.** Revisions 1–3 and 8 carried no gate of their own. A counter reset was
> corrected at attempt 5 and is **not** re-reset by this reauthoring. **Revision 12 adds no
> attempt.**

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
| **Root-cause / compound learning for the seven-attempt loop** | **`docs/compound/workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md`** |
| **Revision-12 root-remediation session record** | **`docs/memory/2026-09-13-stage-compound-root-remediation-attempt-8-candidate.md`** |
| Recovery / torn-state halt record | `docs/memory/2026-09-13-stage-recovery-torn-state-halt.md`, commit `3e0cf8f` |
| Checkpoints | `.backlogit/checkpoints/checkpoint-20260913-231952.json` (attempt-5 terminal state) and its siblings |
| Change lineage for this file | `git log --follow -- docs/plans/2026-09-13-intercom-go-ship-feature-completion-decided-plan.md`; PR #54 |

**Reauthoring does not discharge findings.** `H-1`…`H-9` were disposed at attempt 5.
`J-1`…`J-18` and `K-1`…`K-7` are disposed in §13.3 below, each pointing at the section that
resolves it.
**A finding is closed by a landing site, never by the disappearance of the text that carried
it.**

### 13.3 Findings dispositions at revision 12

**Attempt-7 findings (`M-1`…`M-15`, the 15 deduplicated themes) — dispositions.** The exact
source is `docs/reviews/2026-09-13-true-lineage-attempt-7-verdict.md`. **Every row names a
landing site or an explicit blocker; none is closed by assertion.**

| Ref | Theme | Disposition at revision 12 |
|---|---|---|
| **M-1** | CS3–CS6 still delegate to the mutation-coupled `safe-close` Step 0(c) verdict, so the Option-A boundary has no consumer | **ROOT-FIXED.** §4.1 retargets all four sites to `mode: classify-close-path`; §4.2 pins replacement wording for each (previously **none** of CS1–CS6 was pinned); §6.1 rows 4, 25, 26, 27 give each its own needle; §4.0 C-α rows α2–α5 make the producer→consumer edge mechanical. AC-27. |
| **M-2** | `CLASSIFICATION_BINDING` under-specified — two implementations could disagree; residual window undisclosed | **ROOT-FIXED.** §3.4.3 (D) now gives the full canonical serialisation — field order, `\x1f` intra-tuple delimiter, `shipment_id`, verdict, reason, skill digest, engine version — and **discloses** the recompute→mutate residual window instead of denying it. AC-38. |
| **M-3** | Needle coverage incomplete; several sites creditable by an edit elsewhere | **ROOT-FIXED.** §6.1 is 30 rows; the independent-falsifiability table shows **no borrowed needles** (CS3 moved off row 9d, CS5 off row 10); every mandatory site owns ≥1 needle in its own block. AC-3. |
| **M-4** | §6 assertion table and §6.1 literal table disagreed on row count and content | **ROOT-FIXED.** Both are now **30 rows** and generated against the same inventory; §4.0 closure check 4 requires every matrix row's verification to resolve to a numbered §6.1 row. |
| **M-5** | Unbound `mode: safe-close` still falls through to the destructive path — the sink is open | **PARTIALLY FIXED + NAMED BLOCKER `D-1`.** The **in-scope** half is closed: CS6 forbids **Ship** from invoking `safe-close` without a binding (row 27). The **skill-side** narrowing is **not adopted** — it would narrow a pre-existing public mode and exceed §2's operator-approved additive scope. Recorded as operator decision `D-1` in §3.4.3 (J) and §13.1. AC-42. |
| **M-6** | CS10's P-010 wording pointed at "the five conditions below", which do not exist in `workflow-policies.md` | **ROOT-FIXED.** §4.2 CS10 now renders **all five conditions inline** in `workflow-policies.md` and adds the shipment-active scoping clause; rows 2d and 28 assert both, **in that file**. AC-41. |
| **M-7** | `022.001-T` exceeds the 2-hour rule at 15 clause sites + 30 test rows | **ROOT-FIXED WITHOUT A MANIFEST CHANGE.** §14 restructures the work into **S1a / S1b / S2**. A second *task* was rejected deliberately: it would become a depth-1 live descendant of `022-F`, break §3.2.1 set-equality and **void `PA-021-CASCADE` condition 3**. AC-44. |
| **M-8** | Scope T never probed `doctor`, feature `active -> done`, or classifier behaviour | **ROOT-FIXED BY MEASUREMENT.** §13.6 adds **T-9** (`doctor`), **T-10** (`move` exit codes), **T-11** (linked-deliberation set), **T-12** (mode/classifier surface), each with command and verbatim output. |
| **M-9** | §13.5 attributes to a cited compound doc a "weaker reading" that document does not contain | **ROOT-FIXED.** The source was re-read; it teaches *reproduce installed-behaviour claims* and *scope-check against `main`*, and contains **no** count-scope reading at all. §13.5 bullet 1 is corrected to cite what it actually says. |
| **M-10** | §3.4.3 (B) claimed the classifier reproduces Step 0(c) "verbatim" while also assigning `BLOCK` to cases Step 0(c) handles differently | **ROOT-FIXED.** §3.4.3 (B) is rewritten and **(B1)** adds an exhaustive condition→verdict mapping showing every divergence is a **default-branch token change only** — never a predicate change, never a case moving *into* `CASCADE`. AC-5, row 22c. |
| **M-11** | The cascade also archives linked deliberation artifacts, so §11 condition 6 may be violated | **CLOSED BY MEASUREMENT — no re-scope.** §13.6 **T-11**: `backlogit link list` → `[]` for all three manifest members; the engine's linked-deliberation regex matches **0** times; no `source_deliberation_id`; the index holds **zero** `artifact_type='deliberation'` rows. Recorded as §11 **condition 7**, checked at invocation. A re-scope was deliberately avoided because it would void the operator approval. |
| **M-12** | §9 asserts intake-before-claim, but the installed file runs intake at item 6, *after* the claim at item 4 | **ROOT-FIXED.** **`CS15`** is added — a fifteenth clause site inside the already-in-scope `_ship.agent.md` — physically relocating the intake bullet above the claim bullet. Row 29 asserts the anchor **and** the index ordering. AC-40. |
| **M-13** | §10.4 and Principle XI do not name the merge-commit revert form, and verify the merge strategy at only one of the two merges | **ROOT-FIXED.** §10.4 specifies `git revert -m 1 {merge_sha}`; §15 Principle XI requires merge-strategy verification at **both** merges (§9 steps 5 and 17). |
| **M-14** | §3.4.2 claims an in-merge unreviewed change "fails fact 1 or fact 2", which is false | **ROOT-FIXED.** §3.4.2 adds **fact 2b (review identity)** and withdraws the false claim; the detection now rests on a stated fact rather than an unsound inference. |
| **M-15** | Eight PRESENT needles were not verbatim substrings of any pinned wording — unsatisfiable at `H1` | **ROOT-FIXED, AND THIS IS THE ROOT.** §4.2 now pins **every** site including CS1–CS6, ADD-2 and CS15; §6.1 adds a **Provenance** column requiring each PRESENT needle to be a verbatim substring of its own site's §4.2 bullet; §4.0 makes the producer/consumer/authority/refusal/test/probe edge set mechanically checkable. AC-3, AC-39. |

> **Why these fifteen are one defect, not fifteen.** Every one of them is an instance of the
> same failure: a correction applied to one artifact or section, then asserted complete on the
> strength of a measurement scoped only to that artifact. The full derivation, the recurrence
> chains (`J-5`→`K-1`→`M-1`, `J-12`→`K-2`→`M-3`, `J-2`/`J-7`→`K-4`→`M-12`,
> `H-3`/`J-2`/`J-9`→`K-7`→`M-8`, `K-3`→`M-15`) and the invariant that closes them are in
> `docs/compound/workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md`.
> **§4.0 is that invariant made executable.**

**Attempt-6 blockers (`K-1`…`K-7`) — all RESOLVED at revision 11, re-verified at 12:**

| Ref | Sev | Landing site |
|---|---|---|
| **K-1** the installed skill exposes no pre-mutation verdict boundary | **P0** | **§2 (fourth file) + §3.4.3 + CS11–CS14 + §9 step 14a + rows 21–24.** OPTION A, operator-selected: a new **read-only** `mode: classify-close-path` returns the verdict **before any close call is made**, and safe-close revalidates the bound classification **before any mutation**. §10.1's *"no CLOSE mutation has occurred"* is now a property of the call sequence, not an assertion. AC-33, AC-34, AC-35, AC-36, AC-38. |
| **K-2** CS3/CS5/CS6 not independently falsifiable; AC-4 undetectable | P1 | **§6.1** — every mandatory site has a needle **inside its own block**, and all **four** measured CS4 survivor sentences are pinned ABSENT (rows 7, 9a, 9b, 9c), making AC-4 directly falsifiable. **Revision 11's fix was incomplete** — it satisfied CS3 with a CS4-block needle (row 9d) and CS5 with a CS2-block needle (row 10), i.e. borrowed needles in the very table certifying no borrowing. That residue was `M-3`; revision 12 gives CS3, CS5 and CS6 rows 25, 26 and 27 in their **own** blocks. AC-3, AC-4. |
| **K-3** row 2's needle unsatisfiable by the plan's own pinned wording | P1 | **§4.2 ADD-1 is pinned verbatim**, and rows 2a/2b/2c/2d draw their needles **from that pinned text** — including the `at every depth` quantifier (2b), the set-equality conjunct (2c) and the shipment-active scoping clause (2d), all asserted **in `_ship.agent.md`**. **Revision 11 closed this for ADD-1 only**; eight other PRESENT needles remained unsatisfiable because CS1–CS6 carried no pinned wording at all. That residue was `M-15`; revision 12 pins **every** site and adds the §6.1 **Provenance** column so the defect is structurally unrepeatable. |
| **K-4** `PA-021-CASCADE` names the wrong execution site | P1 | **`.backlogit/queue/021-S.md`** corrected to *"plan section 9 steps 13a-15"*, matching §11's `Execution site` cell, which additionally enumerates 13a/14/14a/15. |
| **K-5** the rollback is itself an unapproved destructive operation | P1 | **§10.3.2 step 2** — the `git restore --source` overwrite and the file deletion are **withdrawn** and replaced by **quarantine-commit → `git revert`**, which contains **no destructive primitive**. `git revert` is the mechanism §10.4 and Principle XI already designate. **No third approval is required** and §11's two-action scope is respected. AC-15. |
| **K-6** §9 steps 1/1a produce a manifest the installed intake check cannot represent | P1 | **§9.3** — pre-mode is invoked at exactly **two** sites, both over a **uniform** manifest; the intake run moves **before** the claim (`expected_status: queued`); step 1a is a **single-record read**, not a manifest-wide run; the mixed window `[1a, 13)` contains **no** pre-mode invocation. AC-37. |
| **K-7** the measurement discipline omits the scope where every failure occurred | P1 | **§13.6 — Scope T (installed tooling)** is added with command, verbatim output and pinned digest for every tooling claim, covering `SKILL.md`, the backlogit engine and its index schema. §13.5's bullet 1 citation is corrected to the lesson the source actually teaches. |

**Attempt-5 findings — dispositions carried forward (all remain landed at revision 12):**

**P1 (`J-1`…`J-9`) — all REMEDIATED:**

| Ref | Landing site |
|---|---|
| **J-1** incomplete rev sweep inside the lineage section | **Structurally eliminated.** §13 carries no inline lineage narrative and no per-revision prose; the revision appears in frontmatter and §13.1 only. Verified by §13.4 measurement. |
| **J-2** occurrence claims scoped to the plan while literals survived in 4 records | §13.4 — every measurement is declared over **{plan} ∪ {4 backlog records}**, with the exact reproducible command. The records are reauthored, so no defect literal survives to be excluded. |
| **J-3** `{post_paths}` consumed but never defined | **§10.3.1** — both `{pre_paths}` and `{post_paths}` defined with command, roots and capture moment; `{post_paths}` captured unconditionally on cascade return. AC-25. |
| **J-4** P-010 grants task transitions only; conflicts with the Ship role table | **§2 + CS10 + §4.2 + AC-24** — surgical third-file scope expansion, exactly one bullet of P-010's `Ship MAY` list, pinned wording matching ADD-1, asserted by test row 20a/20b. |
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
| **J-12** AC-3 claimed test coverage for sites without needles | **AC-3** now enumerates the needle for **every** CS1–CS15 site; §6.1 pins all **30** rows. |
| **J-13** only 9 of 18 rows pinned | **§6.1** pins **all 30 rows**, both polarities, with H0 counts. |
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
  `workflow-policies.md` ∪ `shipment-reconcile/SKILL.md`. This is what the **30-row** test
  asserts over. **All Scope-I counts live in §6.1**, pinned per row, and are not restated here.
* **Scope A — the artifact set**: `{this plan}` ∪ `{021-S, 017-S, 022-F, 022.001-T}` — five
  files. This is what the reauthoring had to clean.
* **Scope T — the installed tooling**: the `shipment-reconcile` skill **as installed**, the
  backlogit engine and its index schema. **Every claim this plan makes about installed
  behaviour is measured here, in §13.6, with command + verbatim output + digest.** `H-3`,
  `H-6`, `J-5`, `J-9` and `K-1` all failed in this scope; it now has a declared scope and a
  reproduction rule like the other two.

> `SKILL.md` is in **both** Scope I and Scope T from revision 11 onward, and that is not a
> conflict: as Scope I it is a file this plan **changes** and pins by test row; as Scope T it
> is the tool whose **current installed behaviour** this plan makes assertions about. §13.6
> measures the pre-change (`H0`) tool; §6.1 rows 21–24 and 30 pin the post-change (`H1`)
> contract.

> **Scope is necessary but not sufficient — the revision-12 correction.** Declaring three
> scopes prevents the `J-2` error (a count stated over one scope and read as another). It does
> **not** prevent the `M-15` error, where a count is correct over the right scope and still
> proves nothing, because it measures **absence at `H0`** for a needle no pinned wording can
> ever produce at `H1`. Satisfiability is a **provenance** property, not a count: it is
> discharged by §6.1's Provenance column and by §4.0's closure checks, never by a number in
> this section. **A green Scope-I measurement is not evidence of closure.**

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
| `14 [s]ites` | **0** | the **revision-11** clause count survived anywhere (`M-12` added CS15) |
| `14 [c]lause` | **0** | same, long form |
| `20 [r]ows` | **0** | the revision-10 row count survived anywhere (`K-1`) |
| `24 [r]ows` | **0** | the **revision-11** row count survived anywhere (`M-3`, `M-4`) |
| `step [1]6` | **0** | the mis-anchored `PA-021-CASCADE` execution site survived (`K-4`) |
| `attempt [6]` | **0** | a stale gate-lineage claim survived (`K-1`) |
| `§4.[1] Pinned` | **0** | a pre-renumber pointer to the pinned-wording section survived (revision 12 moved it to §4.2) |
| `MUST NOT enum[e]rate` | **0** | withdrawn `H-6` wording survived |
| `without touching this [f]ile again` | **0** | withdrawn `H-5` wording survived |

*Measured 2026-09-13 at revision 12 over the five Scope-A files; every row above returned
exactly the stated **0**.*

**Bounded, not zero — and deliberately so:**

| Regex | Scope-A count | Every occurrence is |
|---|---|---|
| `re[v]ision 9` | **2** | both are §13.2 audit-pointer rows. **No normative use.** |
| `re[v]ision 10` | **bounded** | the §0 supersession banner and the explicit "what the superseded revision did and why it changed" contrast notes (§3.4.2, §9.2 check 7, §9.3, §10.3.2, §13.3). **Every one is a historical contrast; none is a governing pointer.** |
| `re[v]ision 11` | **bounded** | the immediately superseded revision, named only in supersession banners, §13.1's lineage table and §13.3's disposition cells explaining what revision 11 left open. **Never a governing pointer.** |
| `re[v]ision 12` | **≥ 6** | the current governing revision, named in the plan and in **all four** records. |
| `attempt [7]` | **bounded** | the last gate that **ran**, in the plan and in all four records. |
| `attempt [8]` | **≥ 6** | the **prepared, unauthorized** candidate. **Every occurrence must be qualified by `NOT RUN`, `candidate`, or `awaiting authorization`** — an unqualified occurrence would misrepresent gate state and is itself a defect. |
| `30 [r]ows` | **≥ 6** | the current row count: plan sites (§5, §6, §6.1, §7, §13.3, §14) + `022-F` + `022.001-T`. |
| `15 [s]ites` | **≥ 6** | the current clause-site count, in the plan and in the two work records. |
| `backlogit_query[_]sql` | **3** | §5's prohibition, AC-28's restatement, and the §13.6 `T-6` measurement that proves the prohibition. **Never a prescription** (`J-9`). |
| `{post[_]paths}` | **≥ 2** | §10.3.1's definition and its §10.3/§10.4 consumption. Definition precedes consumption (`J-3`). |

**Rule for any future claim over these artifacts: name the scope, give the regex, give the
number — and then say what the number cannot prove.** A bare occurrence claim is not evidence,
and neither is a correctly-scoped one when the property at stake is satisfiability rather than
presence (§4.0, `M-15`).

### 13.5 Prior art relied upon

* `docs/compound/2026-09-08-adversarial-review-empirical-verification-and-scope-check.md` —
  **independently reproduce a claim about installed behaviour before asserting it**; measuring
  the artifact you are editing is not a substitute for measuring the tool you are depending
  on. It also teaches **scope-checking a change against `main`**. §13.6 (Scope T) is the
  direct application to `H-3`, `J-2`, `J-9` and `K-1`.

  > **Miscitation withdrawn (`M-9`).** Revision 11 attributed to this document a weaker
  > *"declare your count scope"* reading, and said the stronger reading was evaded by citing
  > the weaker one. **The document contains no count-scope reading at all.** The count-scope
  > discipline in §4.3, §5, §6.1 and §13.4 is this plan's own, developed in response to `J-2`;
  > it is not derived from this source and is no longer presented as such.

* **`docs/compound/workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md`**
  — **this plan's own harvested learning**, authored after attempt 7 from the primary
  verdict artifacts of attempts 1–7. It supplies the closure invariant, the contract-surface
  matrix technique, the before-and-after probe requirement and the reappearance stop rule.
  **§4.0 is its direct application, and §13.3's `M-1`…`M-15` table is its evidence base.**
* `docs/compound/2026-09-08-verify-go-stdlib-internals-claims-against-goroot-source.md` —
  verify a tool/API claim against the installed source, not against memory. Directly applied
  to `J-5` / `K-1` (skill mode inventory and dispatch behaviour), `J-9` (index schema) and
  `M-8` / `M-16` (`doctor` and the `move` gate-broker exit codes).
* `docs/compound/2026-05-07-backlogit-shipment-status-constraints.md` — shipment status enum,
  which is why CS2 deletes the `--status shipped` prescription.
* `docs/compound/2026-09-04-backlogit-sizing-is-wit-gated.md` — §14.
* The P-018 trio cited in §9.

### 13.6 Scope-T measurement — installed tooling (the `K-7` resolution)

**Rule: every assertion this plan makes about installed tooling appears below with the exact
command, the verbatim output and, where the artifact is a file, its digest.** No tooling claim
elsewhere in this plan is admissible unless it resolves here.

**Capture moment**: 2026-09-13, on branch `chore/stage-pipeline-policy-gap` at the tree that
carries revision 12, **before any `H1` edit**. `T-1`–`T-8` were captured at the revision-11 tree and **re-verified unchanged** at revision 12 — no implementation edit has occurred, so the `H0` tool tree is byte-identical; `T-9`–`T-12` are **new at revision 12** and are the `M-8` remediation.

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
# T-9   containment / duplicate-ID detection surface
backlogit doctor --help
backlogit doctor
# T-10  the `move` gate-broker exit-code contract consumed by a1's S4
backlogit move --help
# T-11  the linked-deliberation set for the 021-S manifest (three independent methods)
backlogit link list --id 022-F ; backlogit link list --id 022.001-T ; backlogit link list --id 018-F
Select-String -Path .backlogit/queue/022-F.md,.backlogit/queue/022.001-T.md,.backlogit/queue/018-F.md `
  -Pattern '\b(?:DL\d+|[0-9]+(?:\.[0-9]+)*-DL)\b'
Select-String -Path .backlogit/queue/022-F.md,.backlogit/queue/022.001-T.md -Pattern 'source_deliberation_id'
backlogit query "SELECT COUNT(*) FROM items WHERE artifact_type='deliberation';"
# T-12  intake ordering actually installed in Step 0.5 (the CS15 defect)
Select-String -Path .github/agents/_ship.agent.md -Pattern 'Intake reconciliation check','Record `shipment_id` as the session scope'
```

| ID | Tooling claim made by this plan | Where claimed | Measured value (verbatim) |
|---|---|---|---|
| **T-1** | The installed skill file is the one reviewed at the current `H0` | §3.4.2, §9.2 ck 7 | `git hash-object` ⇒ `312f29b6e393bc9cffec411c386f2f8fc4454e50`; **1082 lines** |
| **T-2** | The installed skill exposes exactly four public **mode sections** — `pre`, `post`, `safe-close`, `detect-mixed-role` — and **no** classification entry point | §2, §3.4.1 | The command returns **five** `### …Mode` headings: L151 `### Mixed-Role Detection **Classification**` (a classification heading, **not** a mode section), then the four mode sections at L257 `### Pre-Mode`, L327 `### Post-Mode`, L354 `### Safe-Close Mode`, L839 `### Mixed-Role Detection Mode (…, READ-ONLY)`. `classify-close-path` ⇒ **×0** |
| **T-3** | On a non-`CASCADE` selection the skill **continues into the mutating** safe-close steps 1–10 | §2, §3.4.3 F, `K-1` | `* **SAFE_CLOSE selected** (default, including any classifier error,` ⇒ **×1 at L509**; L510 reads verbatim *"ambiguity, or unresolved precondition) → continue to step 1 below"* |
| **T-4** | The verdict is recorded **only** in the cascade report, **after** the destructive call | §2, `K-1` | Cascade Close Sub-Procedure begins L618; **L821** reads verbatim *"5. **Produce cascade-close report** recording the classifier's verdict,"* — step 5, after the `backlogit_ship_shipment` invocation |
| **T-5** | The resolved engine identity matches the §8.1 pin | §8.1, §9.2 ck 4 | `SHA256` ⇒ `1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98`; `backlogit version 1.10.1-0.20260823032255-b07729386a31` |
| **T-6** | The index `items` table carries **no** path/location column, so `backlogit_query_sql` cannot detect a torn duplicate | §5, AC-28 | `SELECT *` returns exactly: `artifact_type, assigned_to, branch, commit, created_at, custom_fields, dependencies, description, harness_status, hierarchy_path, id, items, labels, level, owner, parent_id, priority, references, severity, size, source_branch, sprint, status, title, updated_at` — **no path, no location, no directory column**. Hence **`backlogit doctor` only** |
| **T-7** | The installed intake check runs `mode: pre` with a **single** `expected_status` over every manifest item, so a mixed manifest necessarily yields `RECONCILE_FAIL` | §9.3, `K-6` | `_ship.agent.md` **L253–254** prescribe `mode: pre` + `expected_status: queued` *"(or `active` if already claimed)"*; **L259–262** state verbatim *"this single-`expected_status` check applies to true session-start intake, where every manifest task still shares one uniform status (all `queued` pre-claim, or all `active` immediately after this session's own claim…)"* |
| **T-8** | `mode: detect-mixed-role` is an existing **read-only, no-lock** public mode — the precedent Option A's new mode follows | §3.4.3 B/G | **L245** reads verbatim *"* **`mode: detect-mixed-role` is strictly READ-ONLY.** It NEVER mutates any"*, and its bullet continues *"requires NO `file-lock` acquisition because no backlog/shipment artifact is ever mutated"* |
| **T-9** | **`backlogit doctor` is the sole containment / duplicate-ID detector**, and its exits are gate-stable | §5, AC-28, `M-8` | `doctor --help` documents `--target {file}` with exit codes **0** = no issues, **1** = issues found, **2** = usage error, **3** = config error, **4** = internal error. Live run on this tree ⇒ **`No issues found`**, exit **0**. It is the only command that compares `.backlogit/queue/` against `.backlogit/archive/` for a duplicate ID |
| **T-10** | **`backlogit move` routes through a gate broker with a non-binary exit contract** that a1's S4 must handle | §5 S4, AC-43, `M-8`/`M-16` | `move --help` documents: exit **0** success, **6** `blocked by gates`, **7** `configuration/setup error`, **8** `retryable error`. `--force-gates` exists but **requires `--force-reason`** and is documented operator-only. `--json` emits a machine-readable gate outcome. **The help text scopes the broker to *"task/subtask completions"*; behaviour for `artifact_type: feature` is undocumented**, so §5 S4 fail-closes against the **superset** and handles all four codes for the `022-F` move |
| **T-11** | **The validated linked-deliberation set for the `021-S` manifest is EMPTY**, so the cascade archives nothing outside the manifest | §11 cond. 7, `M-11` | Three independent measurements agree. (1) `backlogit link list` ⇒ **`[]`** for `022-F`, `022.001-T` **and** `018-F`. (2) The engine's own linked-deliberation pattern `\b(?:DL\d+\|[0-9]+(?:\.[0-9]+)*-DL)\b` ⇒ **0 matches** across all three records. (3) No record carries `source_deliberation_id`, and `SELECT COUNT(*) … artifact_type='deliberation'` ⇒ **null / zero rows — the workspace contains no deliberation artifacts at all**. **This closes `M-11` by measurement and required no re-scope of `PA-021-CASCADE`** |
| **T-12** | **The installed Step 0.5 runs the intake check AFTER the claim**, contradicting §9 step 1's prose — the `CS15` defect | §9, §9.3, AC-40, `M-12` | In `_ship.agent.md` Step 0.5: `Record \`shipment_id\` as the session scope` (the claim) is item **4**; `Intake reconciliation check` is item **6**. `index(intake) > index(claim)` — **the ordering the plan asserted was never established by any clause edit.** `CS15` relocates it; §6.1 row 29 asserts the corrected index relation |

**Every `T-n` above is reproducible by the commands in the block above, on the stated tree.**
A future claim about installed tooling that does not appear in this table is **not evidence**,
regardless of how confidently it is stated — that is the `K-1` lesson in one sentence.

> **`T-9`…`T-12` are the `M-8` remediation and they changed the plan.** They were not
> confirmations of what revision 11 already believed. `T-10` produced a four-way exit contract
> §5 did not handle; `T-12` produced a fifteenth clause site; `T-11` retired `M-11` without a
> re-scope that would have voided an operator approval. **A measurement scope that is declared
> but never exercised is indistinguishable from one that does not exist** — revision 11
> declared Scope T and ran no probe in it.

## 14. Sizing

**Size `M` · Complexity `high`** (complexity is enum-validated **prose** — this workspace's
`header-def.yaml` defines no `complexity` field on the task type; see
`docs/compound/2026-09-04-backlogit-sizing-is-wit-gated.md`. Do **not** invoke
`backlogit update --complexity`; it fails with exit 1).

| Work item | Session | Est. (min) |
|---|---|---|
| **CS11** mode enum + `classification_binding` input | **S1a** | 2 |
| **CS12** `### Classify-Close-Path Mode` section | **S1a** | 9 |
| **CS13** Step 0 binding revalidation + dispatch scoping | **S1a** | 4 |
| **CS14** Behavioral Constraint bullet | **S1a** | 2 |
| **30-row Go test** (3 file fixtures, generated from §6.1) | **S1a** | 40 |
| H0 red confirmation + `harness-ready` + task claim | **S1a** | 6 |
| **S1a stretch** | | **63** (margin **57**) |
| ADD-1 Role Boundary grant | **S1b** | 8 |
| ADD-2 Step 6.1(a1) incl. S0–S5 + move exit-code handling | **S1b** | 18 |
| CS1 reload + carve-out | **S1b** | 4 |
| CS2 close-sequence pointer | **S1b** | 4 |
| **CS3–CS6** delegation block (retargeted to the classifier) | **S1b** | 10 |
| CS7 intake pointer | **S1b** | 4 |
| CS8 pre-archive pointer | **S1b** | 4 |
| CS9 protected-set scoping | **S1b** | 3 |
| CS10 P-010 grant bullet (five conditions rendered inline) | **S1b** | 5 |
| **CS15** Step 0.5 intake-before-claim relocation | **S1b** | 3 |
| H1 green + §4.0 closure re-check + quality gates | **S1b** | 12 |
| **S1b stretch** | | **75** (margin **45**) |
| Checkpoint recovery, merge verification, §9.2 seven checks | **S2** | 14 |
| a1 (§5 S0–S5), 13a baseline commit, 14/14a/15, postchecks | **S2** | 22 |
| Closure artifacts, sync, closure PR | **S2** | 14 |
| **S2 stretch** | | **50** (margin **70**) |
| **Total across three sessions** | | **188** |

**Why this satisfies the 2-hour rule (`M-7`).** The rule bounds a **single uninterrupted agent
stretch**, not a backlog record. Revision 11 sized the work as one 106-minute S1 stretch and
then grew to 15 clause sites and 30 test rows without resizing — the estimate and the scope had
drifted apart, which is `M-7`. **Revision 12 restructures execution into three stretches of
63 / 75 / 50 minutes, each with ≥45 minutes of margin.**

**The S1a/S1b boundary is the §7 ordering boundary, promoted to a session boundary.** §7
already requires `CS11–CS14` (the producer) to land before `CS3–CS6` (the consumers). That
point — fourth file applied, `H0` still red on every `_ship.agent.md` and
`workflow-policies.md` row — is now the **end of S1a**, a checkpointed session boundary rather
than a mid-task note. It is **verifiable**: `go test` output distinguishes it, because rows
21a–24b and 30 are green while rows 1–20c, 25–29 are still red. S1b resumes from that state.

**No backlog record changes.** S1a / S1b / S2 are **execution stretches inside `022.001-T`**,
not new work items. `022.001-T` keeps `size: M`, `complexity: high` and its single parent.

**Alternative considered and rejected: split `022.001-T` into two tasks.** Rejected because
any second task under `022-F` becomes a live descendant at depth 1, which §3.2.1's
set-equality rule forces into the `021-S` manifest. That changes the manifest from
`[022-F, 022.001-T]` to a three-member list, which **live-mismatches `PA-021-CASCADE`
condition 3** and, under §11, **voids the recorded operator approval** — §11 states the
authorization does not cover *"a re-scoped version of either action."* Re-scoping an approved
destructive action to buy estimate margin is a strictly worse trade than the session split,
and **Stage cannot self-authorize the replacement approval**. **The manifest is therefore
unchanged at revision 12** (§13.1).

> **If the operator prefers a genuine task split**, it is available — but it is **not** a Stage
> decision: it requires re-authorizing `PA-021-CASCADE` against a three-member manifest. That
> is recorded here as an **available option with a named prerequisite**, not as a blocker,
> because the session split already satisfies the 2-hour rule without it.

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
| **XI. Merge History (NON-NEGOTIABLE)** | Rollback is `git revert`; no rewrite, no force-push, no amend. **The reauthoring adds commits; it never amends or rewrites `3e0cf8f`, `3423581` or `e9019d4`.** **Merge-strategy verification is required at BOTH merges** — §9 step 5 (the implementation PR) and §9 step 17 (the closure PR) — not only the first. Each merge must be confirmed a **true merge commit** (`git rev-list --parents -n 1 {merge_sha}` returns three fields) before the session proceeds, because the revert form that §10.4 prescribes is parent-relative: a squash or rebase merge makes `git revert -m 1` fail and silently removes the rollback path this plan depends on. **`M-13`** |

**D1 — 30 rows vs "fewer than 4".** These are **rows in one table-driven function** over
**three** files, sharing one fixture helper and ≤4 helpers; each is a substring or index
assertion, not an independent scenario. Collapsing them reduces falsifiability — and §4.0
closure check 4 requires **every** matrix row to resolve to a numbered §6.1 row, so the row
count is derived from the contract surface rather than chosen. *Assertion density, not scope.*

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

**`shipment-reconcile/SKILL.md` is IN scope**, narrowly: exactly the four
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




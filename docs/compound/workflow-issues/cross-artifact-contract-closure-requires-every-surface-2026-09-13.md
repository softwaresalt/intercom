---
title: "A cross-artifact contract change is closed only when every surface is updated and verified against one bound contract snapshot"
description: "Seven consecutive plan-review gates failed on one defect class: a correction was applied to the artifact where a finding was reported, then declared closed by a measurement scoped to that same artifact, while the consuming, authorizing, refusing, testing and tooling surfaces of the same contract stayed on the previous contract revision and re-raised the identical defect under a new label"
source: "docs/compound/workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md"
doc_type: "learning"
problem_type: "recurring_review_failure"
category: "workflow-issues"
component: "plan-review gate; docs/plans decided-plan authoring; clause-inventory + contract-test pinning"
root_cause: "Closure of a contract spanning producer, consumers, authority grant, refusal path, tests and tooling probes was asserted from a measurement scoped only to the producer or only to the reporting artifact; a local zero-occurrence or producer-only count is structurally incapable of detecting a stale consumer, authority, refusal, test-needle or probe surface, so each gate re-reported the same defect at whichever surface was still stale"
resolution_type: "design_change"
severity: "high"
message: "FAIL P1 recurrence: 'X recurs', 'same defect class as Y relocated to Z', 'defect class surviving its own remediation', 'applied to one half of the contract'"
file_path: "docs/plans/2026-09-13-intercom-go-ship-feature-completion-decided-plan.md"
citations:
  - "docs/reviews/2026-09-13-true-lineage-attempt-7-verdict.md"
  - "docs/reviews/2026-09-13-true-lineage-attempt-6-verdict.md"
  - "docs/plans/2026-09-13-intercom-go-ship-feature-completion-decided-plan.md"
  - "docs/memory/2026-09-13-stage-true-lineage-attempt-7-option-a.md"
  - "docs/memory/2026-09-13-stage-true-lineage-attempt-5-fail.md"
  - ".backlogit/queue/021-S.md"
  - ".backlogit/queue/017-S.md"
  - ".backlogit/queue/022-F.md"
  - ".backlogit/queue/022.001-T.md"
  - ".github/agents/_ship.agent.md"
  - ".github/policies/workflow-policies.md"
  - ".github/skills/shipment-reconcile/SKILL.md"
  - "docs/compound/2026-09-08-adversarial-review-empirical-verification-and-scope-check.md"
tags:
  - "compound"
  - "plan-review"
  - "contract-surface"
  - "circular-remediation"
  - "measurement-scope"
  - "stage"
  - "021-S"
  - "022-F"
---

## Problem

Seven consecutive `plan-review` gates (true-lineage attempts 1–7) over one decided plan
returned **FAIL**. The surface counts moved — attempt 6 was `P0 = 1 / P1 = 7`, attempt 7 was
`P0 = 0 / P1 = 27 raw / 15 deduplicated` — but the *defect class* did not. Six of the seven
gates were lost to the same shape, and the verdict artifacts name the recurrence explicitly:

| Chain | Attempt-n finding | Attempt-n+1 finding | Verdict's own words |
|---|---|---|---|
| Delegation target | `J-5` delegation targets an API the skill does not expose | `K-1` the verdict is unreadable before mutation | *"the **same defect class as `J-5`** … relocated from the classification API to the mutation boundary. `J-5`'s remediation corrected **where** the verdict comes from but not **when** it becomes readable"* |
| Delegation target | `K-1` (P0) | `M-1` | *"Option A was applied to **one half of the contract**"* — `CS3`–`CS6` still authorize the cascade on *"the verdict reported by `mode: safe-close` Step 0(c)"* |
| Needle coverage | `J-12` AC claimed coverage for sites without needles | `K-2` `CS3`/`CS5`/`CS6` not independently verifiable | *"`J-12`'s defect class surviving its own remediation"* |
| Needle coverage | `K-2` | `M-3` | shared/borrowed rows, unanchored assertions |
| Scope of sweep | `J-2`/`J-7` literals survived in coupled records | `K-4` `PA-021-CASCADE` names the wrong execution site | *"`J-7`'s defect class reproduced in the coupled record while being fixed in the plan — i.e. a **scope-of-sweep** failure, the `J-2` pattern"* |
| Measurement scope | `H-3`/`J-2`/`J-9` | `K-7` measurement omits the scope where every failure occurred | `M-8` Scope T still omits `doctor`, the feature `active → done` move, and the classifier |
| Needle satisfiability | `K-3` needle unsatisfiable by the plan's own pinned wording | `M-15` | PRESENT needles still conflict with §4.1 |

**The loop is circular, not merely repetitive.** A repetitive loop re-fails on *new* ground
each cycle. Here each cycle's remediation *created* the next cycle's finding by moving the
contradiction to an adjacent surface rather than closing it, and each cycle's closure evidence
was a measurement that could not see the surface the contradiction moved to.

## Root Cause

**The contracts in this plan span multiple artifacts, but closure was proven with
single-artifact measurements.**

One contract — "a non-`CASCADE` close path must be refusable before any close mutation" — has
**six** live surfaces:

1. **producer / API** — the classifier that emits the verdict (`shipment-reconcile/SKILL.md`);
2. **consumer** — the clauses that read and act on it (`_ship.agent.md` `CS3`–`CS6`);
3. **authority grant** — the policy that permits the resulting action (`P-010` `CS10`);
4. **refusal / fallback path** — what happens on unbound, drifted, mixed, ambiguous or
   non-`CASCADE` input (the safe-close default sink);
5. **tests** — the pinned needle that makes each site independently falsifiable;
6. **measurement probes** — the commands that establish what the installed tooling actually
   does.

Each remediation cycle updated surface (1) or whichever surface the reviewer had *reported*,
then declared the contract closed on the strength of a count taken over that same surface. The
count always returned the expected number, because the count was scoped to the part that had
just been fixed.

**Mechanically verified in the attempt-7 tree (revision 11), by direct measurement:**

* §4.1 "Pinned replacement wording (normative)" pins replacement text for **9 of 14** clause
  sites (`CS7`, `CS8`, `CS9`, `ADD-1`, `CS10`, `CS11`–`CS14`). **`CS1`–`CS6` have none.** The
  producer sites (`CS11`–`CS14`) were pinned in the same revision that adopted Option A; the
  consumer sites were not.
* §4's own clause table still describes `CS3` as *"do not call it unless **the verdict reported
  by `mode: safe-close` Step 0(c)** authorizes it"* — the **pre-Option-A, mutation-coupled**
  source that `K-1` rejected, sitting three sections away from §3.4.3, which replaced it.
* **8 of the 26 transition-row PRESENT needles** in §6.1 (rows 3, 4, 5, 9d, 10, 11, 12, 16)
  are **not verbatim substrings of any pinned normative wording anywhere in the plan**. Row 11
  pins `scoped to Role Boundary changes only` while the only governing text reads
  `scoped to **Role Boundary** changes only` (emphasis markers inside the span). An executor
  applying the plan's own wording leaves those rows **red at H1**.
* §6.1's "independent falsifiability" table — the artifact whose entire purpose is to prove
  per-site coverage — assigns **`CS3`** a needle located in the **`CS4`** block (row 9d) and
  **`CS5`** a needle located in the **`CS2`** block (row 10). Two of fourteen sites have no
  own needle *in the table that certifies that every site has one*.
* §6.1 nevertheless states: *"Independently re-measured at revision 11 … every ABSENT needle
  resolves ×1 and every PRESENT needle ×0 in the live tree."* **That measurement is true and
  proves nothing.** It is an `H0`-absence count over the pre-change tree. It cannot detect an
  unpinned consumer clause, a needle that no pinned wording satisfies, a borrowed needle, a
  stale authority grant, an unrefused sink, or an unprobed tool.

That last bullet is the root cause in one line: **the measurement that certified closure was
structurally incapable of detecting the failure mode that kept recurring.**

Two aggravating conditions, both now removed:

* **Append-only artifact drift** — superseded wording accumulated in the plan and in the four
  coupled backlog records, so a "fixed" claim and its superseded contradiction coexisted.
  Largely solved by the operator-directed reauthoring at revisions 10–11, which is why `J-1`
  and `J-14` did not recur.
* **Zero harvested learnings** — the attempt-7 Learnings Researcher recorded that **six failed
  gates had produced zero compound entries** (`L-9`). With nothing written down, each cycle
  re-derived the remedy from the last verdict alone, which describes a *symptom location*, not
  a *contract*. This document is the correction.

**What the loop was NOT.** Three candidate explanations were tested against the artifacts and
rejected: the plan does **not** assume a nonexistent API any more (`J-5`/`K-1` are closed at
the producer — `mode: classify-close-path` is added by the plan itself); the reviewers did
**not** merely relabel — each `M-n` cites a distinct, independently reproducible defect; and
context drift is **not** the driver post-reauthoring, since revisions 10 and 11 carry no
inline review history at all and the class still recurred.

## Resolution

**The invariant (normative):**

> **A cross-artifact contract change is complete only when the producer, every consumer, the
> authority grant, every refusal/fallback path, the tests, and the measurement probes are all
> updated and verified against the *same bound contract snapshot*. A local zero-occurrence
> result, or any check scoped only to the producer or only to the artifact where the finding
> was reported, cannot prove closure and must not be cited as closure evidence.**

Applied as a three-part technique:

### 1. Contract-surface matrix — build it before editing anything

For every contract the change touches, enumerate **all six surfaces** as rows before the first
edit. Each row states four things, and a row with any of them missing is an **open edge**:

| Column | Requirement |
|---|---|
| **Current evidence** | The exact text, line and file as measured *now*, quoted verbatim |
| **Intended change** | The pinned replacement wording — the literal an executor will paste |
| **Executable verification** | A command or test row that returns a number, whose PRESENT needle is a **verbatim substring of the pinned replacement wording** |
| **Depends on** | The other surface(s) that must land first, making ordering explicit |

Then **mechanically detect missing edges** rather than reading for them:

* every producer row has ≥1 consumer row;
* every consumer row that acts on an outcome has an authority row permitting that action;
* every outcome value in the producer's value set has a refusal row or a consumer row;
* every surface row has ≥1 test row whose needle resolves **inside that surface's own block** —
  never borrowed from a neighbouring site;
* every claim about installed tooling has a probe row with command + verbatim output.

**The matrix is closed only when no row has an open edge.** Until then, the change is
incomplete by construction, regardless of how many local counts return zero.

### 2. Before-and-after executable probes, on both sides of the boundary

For each surface, the verification must be runnable **now** (establishing the `H0` state) and
**after** (establishing the `H1` state), over the **same** named scope. Separate and name every
scope — the artifacts being edited, the coupled records, and the **installed tooling** — and
never let a count taken in one stand as evidence about another. A tooling claim without a
command and its verbatim output is **not evidence**, however confidently stated. (This is the
stronger reading of
`docs/compound/2026-09-08-adversarial-review-empirical-verification-and-scope-check.md`:
*independently reproduce a claim about installed behaviour before asserting it*. That source
teaches reproduction and scope-classification-against-`main`; it does **not** teach
"declare your count scope", and citing the weaker reading is how the stronger one was evaded
at `H-3`, `J-2`, `J-9`, `K-1` and `M-8`.)

Probing the previously-unprobed surfaces immediately produced two live defects that seven
prose-only gates had not: `backlogit move <id> --status done` can exit **6 / 7 / 8** through a
gate broker the plan's transition step did not handle at all, and the cascade's
linked-deliberation set is derived by an engine-internal regex that no plan section had ever
measured.

### 3. Stop rule — a reappearing finding is a design defect, not a patch target

> **When a finding reappears under a new label at an adjacent surface, stop patching. The
> next action is not another local fix; it is to build the contract-surface matrix for that
> contract and close every open edge in one revision.**

Concretely: two consecutive gates in which any finding's disposition reads *"recurs"*,
*"same class as"*, *"relocated"*, *"partially closed"* or *"improved, recurs deeper"* triggers
a mandatory compound-learning pass **before** the next gate is spent. Attempt 7's disposition
table contained four such entries (`K-2` recurs, `K-3` recurs, `K-6` partially closed, `K-7`
improved/recurs deeper) — the trigger had been met twice over and was not acted on, because
no rule existed to act on it.

## Prevention

* **Never cite a producer-scoped or reporting-artifact-scoped count as contract closure.** Name
  the surface the count covers. If the contract has six surfaces and the count covers one, say
  so, and treat the other five as unverified.
* **Pin the replacement wording for every clause site before writing any test needle**, and
  derive each needle **mechanically** from that pinned wording as a verbatim substring. Never
  author a needle from prose or from memory of intent. Emphasis markers, backticks and
  punctuation inside the span are part of the literal.
* **Never let one site's needle certify another site.** A borrowed needle means an
  implementation that edits only the lending site leaves the borrowing site unapplied with a
  fully green suite.
* **Enumerate the producer's full value set and require every value to be consumed or
  refused.** `BLOCK` being a refusal rather than a path is a contract fact that must appear in
  the consumer *and* the authority *and* the refusal surfaces, not only in the producer.
* **Check the authority grant is self-contained at its own site.** A policy bullet that says
  *"all five conditions below"* while the five conditions live in a different file is a
  dangling reference that reads correct in review and resolves to nothing at run time.
* **Probe the tool, not the prose.** Any step naming a CLI/MCP operation must have run
  `--help` (or the operation itself, read-only) and recorded its exit-code contract and output
  shape in the plan.
* **Harvest a learning at the second consecutive same-class failure, not the seventh.** Six
  failed gates producing zero compound entries is itself the strongest predictor that the
  seventh will fail the same way.

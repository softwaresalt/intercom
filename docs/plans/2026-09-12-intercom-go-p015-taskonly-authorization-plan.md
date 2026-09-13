---
title: "Plan — Dormant token-gated TASK_ONLY_FINALIZE authorization in P-015 (Policy gate)"
date: 2026-09-12
status: planned
agent: Stage
revision: 1
feature: 023-F
task: 023.001-T
shipment: 022-S
depends_on_shipment: 021-S
deliberation: docs/decisions/2026-09-12-intercom-go-three-shipment-bootstrap-deliberation.md
umbrella_evidence: docs/plans/2026-09-12-intercom-go-task-only-shipment-finalization-plan.md
---

# Plan — Dormant token-gated TASK_ONLY_FINALIZE authorization in P-015 (Policy gate)

> **Requires plan hardening**: **yes — applied in rev 1** (see `## Plan Hardening`).

- **Shipment**: `022-S` (B, Policy gate) — second of three
- **Depends on**: `021-S` (A) — B's closure requires A's covering-feature completion authority
- **Closes by**: A's parent-completion authority + the **existing** P-015 `CASCADE`
- **Activates**: nothing. The branch it authorizes is **dormant** until C ships

## 1. Objective

Authorize a **dormant, exact-version** `TASK_ONLY_FINALIZE` policy branch in P-015,
ordered:

```text
CASCADE  →  TASK_ONLY_FINALIZE  →  SAFE_CLOSE / HALT
```

selectable **only** when **both**:

1. the **installed** reconciliation skill advertises the required **capability/schema
   token**, and
2. **every** policy topology and safety precondition holds.

Until C ships the token, condition 1 is **false by absence**, the branch can never be
selected, and behavior is byte-for-byte today's behavior. Also reconcile P-015's blanket
and sole-exception clauses and add the Amendment Log row.

## 2. Why dormant-by-token is the right construction

The alternative — land the authorization live and rely on the skill "not implementing it
yet" — is unsafe: a partially-drafted classifier, or a reviewer reading P-015 alone,
could conclude the path is available. Token gating inverts the default:

| Property | Consequence |
|---|---|
| Absence of the token is the **default state** | No action is required to keep the path off |
| The token is advertised by the **installed skill**, not by config | It cannot be enabled by a flag, an operator, or an agent |
| The gate is a **precondition of selection**, not a warning | A classifier that ignored it would fail P-015, not merely deviate |
| The token is **exact-version** | A future incompatible skill revision **de-authorizes** the branch rather than silently keeping it |

This is fail-closed **by construction**, not by discipline. It is what makes B safe to
land as its own shipment.

## 3. Surface — exactly 2 files

| # | File | Change |
|---|---|---|
| 1 | `.github/policies/workflow-policies.md` | P-015 clause reconciliation + new dormant exception block + Amendment Log row |
| 2 | `tests/integration/p015_taskonly_authorization_contract_test.go` | one table-driven test fn + ≤4 helpers |

**New production code files: 0.** No agent file, no skill file. **A's `_ship.agent.md`
delegation means Ship needs no edit** — that is precisely what A bought.

## 4. Exact inventory — 6 clause sites + 2 additive

Line anchors **re-verified against the installed file at `e7e981d`**; all match rev 6
Inventory Table §6.1-A exactly.

| # | Line | Clause | Class | Change |
|---|---|---|---|---|
| **A1** | 410 | heading *"Single-Artifact Shipment Closure (**No Cascade Ship**)"* | CONSISTENCY | parenthetical now describes neither exception |
| **A2** | 420 | *"It MUST NOT call the cascade `backlogit_ship_shipment` for closure"* | **MUST** | the blanket prohibition **is** the authorization gate; scope it to the default branch |
| **A3** | 426 | Postcondition: *"the cascade … was never called"* | **MUST** | unreachable once a second exception exists; scope to the safe-close branch |
| **A4** | 436 | *"**VERIFIED FULLY-COVERED-ROOT EXCEPTION** … remains the DEFAULT"* | **MUST** | asserts exclusivity; amend per the preserved-region rule (§4.1) |
| **A5** | 443 | item 6 *"IS permitted … in place of the single-artifact safe-close"* | **MUST** | scope the permission to the `CASCADE` branch |
| **A6** | 448 | *"This exception is defined entirely in terms of the general shape above … a pure classification function … that Ship consults to **select the close path**"* | **MUST** | the selector sentence; binary → **ternary** |
| **A7** | new block after 448 | — | **MUST** | **NEW** dormant `TASK_ONLY_FINALIZE` exception block |
| **A8** | after 774 | — | **MUST** | Amendment Log row naming `TASK_ONLY_FINALIZE` and its dormancy |

**6 clause sites (5 MUST + 1 CONSISTENCY) + 2 additive.** This is B's **entire** share of
the rev-6 28-site inventory. No skill site and no Ship site appears here.

### 4.1 Preserved region (mandatory)

The existing `CASCADE` exception's **items 1–5 and item 7, and the SUPERSESSION NOTE**, are
**byte-preserved**.

**Item 6 is the one exception and is amended by A5.** Item 6 is the *permission* sentence
(*"the cascade … IS permitted … in place of the single-artifact safe-close"*), so scoping
it to the `CASCADE` branch is unavoidable — a blanket permission cannot coexist with a
second verdict. A4/A6 amend only framing sentences that assert *exclusivity* or
*selection*.

**No other precondition text is touched.** Test row 9 asserts byte-preservation of items
1–5 + 7 + the SUPERSESSION NOTE; test row 12 asserts item 6's amendment is **scoping
only** — it must still permit the cascade for `CASCADE`, and must not grant any
permission to a second verdict.

Rationale: items 1–5 and 7 encode the `allowed_ids`/`required_ids` gate and the three
rounds of 155-S corrections. Re-drafting them is out of scope and would silently re-open
closed review findings.

## 5. What A7 contains — and what it does not

**A7 contains** (the policy-level authorization):

1. The **capability/schema token** requirement — a **compatibility-level** token of the
   form `task-only-finalize/v1`, advertised by the **installed** `shipment-reconcile`
   skill. Absence, mismatch, or ambiguity → **not selected**.

   **Compatibility-level, not revision-level (review correction).** The token binds to the
   *contract version* of the classification, **not** to an incidental skill file revision.
   A routine editorial change to `SKILL.md` must **not** de-authorize the branch; only an
   **incompatible** change to the classification contract does, by bumping to `/v2`, which
   P-015 then does not authorize until amended. This keeps the gate a genuine compatibility
   handshake rather than a build-identity coupling.
2. The **classifier order**: `CASCADE` → `TASK_ONLY_FINALIZE` → `SAFE_CLOSE`/`HALT`.
3. The **topology assertions T1–T5** (policy-level safety preconditions):
   - **T1** task-only live manifest;
   - **T2** exactly one live member declaring `status: done`;
   - **T3** zero live feature and zero live subtask members (archived task/subtask members
     **are** permitted — proven by the EXACT `017-S` fixture, Probe 15,
     `INVARIANCE_FAILURES=0`);
   - **T4** exactly one root parent feature, **live and OUTSIDE** the manifest;
   - **T5** no omitted live descendants at any depth.
4. **Mutual exclusivity by construction**: `CASCADE` requires ≥1 qualifying feature
   member; `TASK_ONLY_FINALIZE` requires **zero** (T3). No manifest satisfies both, so the
   ordering resolves no ambiguity and creates none.
5. **Fail-closed fallback**: any guard failure, ambiguity, or query error → not classified
   → `SAFE_CLOSE`/`HALT`, exactly as today.

**A7 does NOT contain** (C's obligations, deliberately):

* the guards `G0–G13` and post-guards `P1–P10`;
* the classifier/selector implementation;
* CLI digest-bound invocation;
* the two-path failure split and bounded recovery;
* the self-host ordering.

This division is the layering the deliberation §4 option 2 protects: **P-015 authorizes;
the skill procedures.** It is also why B's budget closes.

## 6. Ordering (test-first; `main` stays green)

1. **H0 — red.** One table-driven test function. Marker:
   `not implemented: p015 task-only authorization contract`. Red count **1 of 1**;
   `go vet` clean; `go test ./...` non-zero.
2. **H1 — green.** Apply A1–A8. Re-run: `go vet` clean, `go test ./...` green,
   `markdownlint` clean.

Exactly **one** generated test function — the `021.001-T` one-queued-task constraint.

## 7. Verification obligations

### 7.1 What the Go test proves

**Contract text and ordering on `workflow-policies.md` only.** 11 rows:

| # | Assertion | Kind |
|---|---|---|
| 1 | P-015 names `TASK_ONLY_FINALIZE` | positive |
| 2 | The classifier order appears as `CASCADE` → `TASK_ONLY_FINALIZE` → `SAFE_CLOSE`/`HALT` | **ordering** |
| 3 | A7 states the capability/schema token requirement, with an exact version | positive |
| 4 | A7 states that token **absence** means **not selected** | positive |
| 5 | A7 states T1–T5 | positive |
| 6 | A7 states mutual exclusivity via the zero-feature-member rule | positive |
| 7 | A7 states the fail-closed fallback to `SAFE_CLOSE`/`HALT` | positive |
| 8 | Amendment Log row present, naming the dormancy | positive |
| 9 | The existing exception's **items 1–5, item 7 and the SUPERSESSION NOTE are byte-unchanged** | **negative** |
| 10 | P-015 does **not** state `G0`–`G13` or `P1`–`P10` (C's surface) | **negative** |
| 11 | P-015 does **not** authorize `TASK_ONLY_FINALIZE` unconditionally — the token clause is **not** absent | **negative** |
| 12 | Item 6's amendment is **scoping only**: it still permits the cascade for `CASCADE` and grants **no** permission to a second verdict | **negative** |

Rows 9–12 are the widening guards.

### 7.2 What it does **not** prove

It does **not** execute a closure, does **not** prove the branch is unreachable at
runtime (that is a property of the absent token, argued in §2 and confirmed by C's
activation probe), and does **not** prove the skill's guards are sufficient — those are
C's obligations. Honest scoping: **B proves policy text and ordering, nothing more.**

## 8. Self-hosting closure sequence

`022-S` manifest is `[023-F, 023.001-T]` — a **fully-covered root**. `023-F` is a root
(no `parent_id`), its complete descendant set at every depth is exactly `{023.001-T}`,
and the manifest contains nothing else.

**B closes by the EXISTING `CASCADE` exception — not by the branch it just authorized.**
That branch is dormant (no token), so it is not even selectable at B's own closure. This
is the bootstrap property.

| Step | Session | Action |
|---|---|---|
| 1 | S1 | Orchestrator routes `022-S` **only after** `021-S` is archived `shipped`; claim; `023-F -> active`, `023.001-T -> active` |
| 2 | S1 | H0 red → implement A1–A8 → H1 green; commit; push; PR; operator approval; merge |
| 3 | S1 | Verify merge SHA; `023.001-T -> done` |
| 4 | S1 | Reload merged `main`. **No Role Boundary change in this merge**, so §6 of plan A does **not** force a session end |
| 5 | S1 | Step 6.1(a1) — A's authority: verify conditions; `023-F active -> done` |
| 6 | S1 | Pre-mode `expected_status: done` → both members `done` → `PROCEED` |
| 7 | S1 | Classification → **`CASCADE`** (fully-covered root; `TASK_ONLY_FINALIZE` not selectable — token absent); close via cascade; P-007; post-mode; sync |
| 8 | S1 | Operational closure + P-020; closure PR; operator merge |
| 9 | — | Orchestrator routes **C** only |

### 8.1 Why B may close in one session, unlike A

A changed **Ship's own Role Boundary**, which plan A §6 forbids taking effect
mid-session. B changes **policy text only** — it grants Ship no new authority. The
authority B's closure relies on (`023-F active -> done`) was merged and session-loaded in
**A**, one shipment earlier. So B's single-session closure is consistent with the
authority-escalation rule rather than an exception to it.

This asymmetry is deliberate and is itself evidence that the rule is scoped correctly:
it binds **authority** changes, not all contract changes.

## 9. Intermediate state after B merges (the coherence proof)

| Surface | State | Consequence |
|---|---|---|
| P-015 | authorizes `TASK_ONLY_FINALIZE` **conditionally on the token** | branch exists |
| `shipment-reconcile` | unchanged — advertises **no** token | condition **false** |
| `_ship.agent.md` (post-A) | delegates to the machine verdict | invokes only what is named |
| **Selectable verdicts** | **`CASCADE`, `SAFE_CLOSE`, `HALT`** | **identical to today** |

No reachable state exists in which `TASK_ONLY_FINALIZE` is selected between B's merge and
C's merge. The intermediate contract is **coherent and fail-closed**.

## 10. Risks, residuals, rollback

| ID | Risk | Disposition |
|---|---|---|
| **R-1** | An implementer drafts the token clause as advisory ("should advertise") | Test row 4 asserts the **not selected** consequence of absence; row 11 asserts the clause is present |
| **R-2** | A4/A5/A6 edits accidentally alter items 1–7 | Test row 9 asserts byte-preservation of the region |
| **R-3** | C's token name/version drifts from A7's | **Coupling obligation CO-1** (§10.1) |
| **R-4** | B cannot close because A's a1 gate failed (plan A R-1/PO-1) | B is **blocked** by dependency until `021-S` is archived `shipped`. B is never reached if A halted |
| **R-5** | Reviewer reads A7 alone and believes the path is live | A7 states dormancy explicitly; A8's Amendment Log row records it |

### 10.1 CO-1 — token identity coupling (blocking for C)

A7 fixes the token's **exact name and version**. C MUST advertise **that exact string**.
C's plan carries the reciprocal obligation and its Go test asserts parity against P-015.
If C needs a different token, the change is made **in B's surface by Stage**, re-reviewed,
and only then implemented — never adjusted unilaterally inside C.

### 10.2 Rollback

Two files on a feature branch. `git revert` the merge commit. Because the branch is
dormant, reverting B changes **no runtime behavior** — the strongest possible rollback
property, and a direct consequence of token gating.

### 10.3 Failure paths

| Failure | Response |
|---|---|
| H1 not reachable in budget | HALT, return to Stage |
| A5/A6 cannot be scoped without touching items 1–7 | HALT, return to Stage — do **not** re-draft the preserved region |
| Classification returns `SAFE_CLOSE` at B's closure | HALT — the manifest is not the fully-covered root asserted. Return to Stage |
| a1 gate unavailable (A not actually live) | HALT — dependency ordering was violated |

## 11. Acceptance criteria

- **AC-1** A1–A8 all applied; 6 clause sites + 2 additive.
- **AC-2** A7 states the exact-version capability/schema token and that absence ⇒ not selected.
- **AC-3** A7 states T1–T5 and the classifier order.
- **AC-4** A7 states mutual exclusivity by the zero-feature-member rule.
- **AC-5** A7 states the fail-closed fallback.
- **AC-6** The existing exception's items **1–5, item 7 and the SUPERSESSION NOTE** are **byte-unchanged**; item 6 is amended by **A5 for scoping only** and still permits the cascade for `CASCADE`.
- **AC-7** P-015 states **no** `G0`–`G13` and **no** `P1`–`P10`.
- **AC-8** Amendment Log row present, naming dormancy.
- **AC-9** Exactly **one** Go test function; H0 red 1 of 1; H1 green; `go vet` clean.
- **AC-10** `markdownlint` clean.
- **AC-11** `_ship.agent.md` and `shipment-reconcile/SKILL.md` are **unchanged** by this shipment.
- **AC-12** Closure follows §8 and classifies **`CASCADE`**, never `TASK_ONLY_FINALIZE`.
- **AC-13** Exactly 2 files changed. 0 new production files.

## 12. Sizing

**Size `M` · Complexity `high`.** Complexity carried as **enum-validated prose** —
`header-def.yaml` defines `size` on tasks but no `complexity` field.

| Work item | Estimate |
|---|---|
| A1 heading (CONSISTENCY) | ~1.5 min |
| A2 + A3 mechanical MUST | ~3 min |
| A4 + A5 + A6 structural MUST | ~12 min |
| **A7 dormant exception block** (token + order + T1–T5 + exclusivity + fallback) | **~20 min** |
| A8 Amendment Log row | ~1 min |
| Go table-driven test (11 rows) + ≤4 helpers | ~18 min |
| H0→H1, `go vet`, `go test`, `markdownlint` | ~10 min |
| **Total** | **~66 min ≈ 1.1 h** |

Margin to the 2-hour rule: **~54 min**. A7 is ~20 min rather than rev 6's ~15 because it
adds the token gate; it is far below rev 6's ~30 min `B18` because **T1–T5 only** land
here and `G0–G13`/`P1–P10` go to C.

**Contingency.** Exceeding 2 h mid-execution → **HALT and return to Stage.**

## Plan Hardening

**Blast radius.** P-015 governs closure for every shipment in the workspace. But B's
change is **dormant**, so its runtime blast radius at merge is **zero** — the only live
effects are the A1–A6 framing edits, which scope existing statements rather than widen
them. This is the lowest-risk of the three shipments despite touching the highest-stakes
file.

**Hardening 1 — "dormant" must be a property, not a promise.** The token is advertised by
the **installed skill**, so dormancy is a fact about the filesystem, not a configuration
choice. There is no flag to misset, no operator to trust, no ordering to remember. Test
rows 4 and 11 fail if the clause degrades to advisory.

**Hardening 2 — the preserved region is the real hazard.** A4/A5/A6 sit adjacent to items
1–7, which carry the `allowed_ids`/`required_ids` gate and three rounds of 155-S
corrections. An implementer "tidying while nearby" could silently re-open closed findings.
§4.1 makes preservation normative and names the **one** unavoidable exception — item 6,
the permission sentence, which A5 scopes to the `CASCADE` branch. Test row 9 makes
preservation of items 1–5 + 7 + the SUPERSESSION NOTE mechanical; **test row 12 bounds
item 6's amendment to scoping only**. §10.3 makes "cannot scope without touching the
preserved region" a **halt**, not a judgment call.

*(Rev-1 review correction: an earlier draft claimed items **1–7** byte-preserved while A5
edits item 6 — an internal contradiction that made the plan unimplementable as written.
Resolved above.)*

**Hardening 3 — ordering claims must be exact.** `CASCADE` is evaluated **first**, and
`TASK_ONLY_FINALIZE` is unreachable for any manifest with a live feature member (T3).
Mutual exclusivity is **by construction**, so the ordering is belt-and-braces, not the
safety mechanism. Test row 2 asserts order; row 6 asserts the construction. Both, because
either alone would be a weaker claim than the policy needs.

**Hardening 4 — dependency on A is real, not stylistic.** B's manifest is a fully-covered
root, so B's closure **requires** A's Step 6.1(a1). If B were routed before A, it would
implement cleanly and then **fail at closure** with no live authority to complete `023-F`.
The `022-S depends_on 021-S` edge is therefore load-bearing, and §10.3 halts rather than
improvises.

**Hardening 5 — coupling to C is the one cross-shipment obligation.** CO-1 is the only
place B and C must agree. It is narrow (one token string), asserted from **both** sides
(B's row 3, C's parity row), and changes to it route through **Stage**, not through C.
Unilateral adjustment inside C is forbidden precisely because it would silently
de-authorize or over-authorize the branch.

**Hardening 6 — what could make this plan wrong.** If the skill cannot advertise a
machine-readable token at all, A7's gate is unimplementable and the whole three-shipment
ordering must be revisited. That is a **C** discovery risk, surfaced early by CO-1 rather
than at C's closure. Stage judges the risk low — the skill already carries structured
frontmatter — but the plan does not assume it, and C's plan carries the falsification
step.

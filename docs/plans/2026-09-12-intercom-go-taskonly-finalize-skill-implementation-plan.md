---
title: "Plan — TASK_ONLY_FINALIZE classifier, guards and recovery in shipment-reconcile (Skill implementation)"
date: 2026-09-12
status: planned
agent: Stage
revision: 1
feature: 024-F
task: 024.001-T
shipment: 023-S
depends_on_shipment: 022-S
deliberation: docs/decisions/2026-09-12-intercom-go-three-shipment-bootstrap-deliberation.md
umbrella_evidence: docs/plans/2026-09-12-intercom-go-task-only-shipment-finalization-plan.md
evidence_dir: docs/plans/evidence/2026-09-12-task-only-shipment-finalization/
---

# Plan — TASK_ONLY_FINALIZE classifier, guards and recovery in shipment-reconcile

> **Requires plan hardening**: **yes — applied in rev 1** (see `## Plan Hardening`).

- **Shipment**: `023-S` (C, Skill implementation) — third of three
- **Depends on**: `022-S` (B), which depends on `021-S` (A)
- **Closes by**: the **new** `TASK_ONLY_FINALIZE` path — **C is the self-proof**
- **Manifest**: **task-only** `[024.001-T]`

## 1. Objective

Implement, in `shipment-reconcile/SKILL.md`, the procedure that B's dormant policy branch
authorizes:

1. the **capability/schema token** (activating B's gate);
2. the **narrow classifier/selector** for `TASK_ONLY_FINALIZE`;
3. the **`G0–G13` pre-call guards** and **`P1–P10` post-call guards**;
4. **CLI digest-bound invocation**;
5. the **two-path failure split** and **bounded recovery**;
6. the **self-host ordering**.

On merge, A's generic delegation + B's dormant authorization **activate together**, and C
closes itself **task-only**. That closure is the durable runtime proof the umbrella plan
could never obtain.

## 2. Surface — exactly 2 files

| # | File | Change |
|---|---|---|
| 1 | `.github/skills/shipment-reconcile/SKILL.md` | 16 clause sites + `B11` + `B18` + token in frontmatter |
| 2 | `tests/integration/taskonly_finalize_contract_test.go` | one table-driven test fn + ≤4 helpers |

**New production code files: 0.**

**Test-file decision (explicit).** A **separate** test file is used rather than extending
an existing integration file. Rationale: task isolation stays unambiguous, the H0 red
count is trivially `1 of 1`, and the file is the natural home for the token-parity
assertion against P-015. Still exactly two files.

**Evidence (§7.3) is not a third surface.** `docs/plans/evidence/…/clause-inventory.md`
is an evidence artifact, as rev 6 §6.1-E established. A and B may **cite** shared
observations from this directory; **C owns it.**

### 2.1 Token feasibility — falsified in advance

Plan B hardening 6 flagged "the skill may be unable to advertise a machine-readable
token" as C's discovery risk. **Resolved during planning**: `SKILL.md` already carries
YAML frontmatter (`name`, `description`) at lines 1–4, so a `capabilities:` key is
directly addable. Risk closes **low** before C starts.

### 2.2 Execution boundary — the skill markdown IS the classifier here (verified)

`SKILL.md` names `src/autoharness/gates/shipment_closure.py` as *"this self-hosting
repository's own `classify_shipment_close_path` implementation"* at **line 413** (inside
`B9`) and **line 1075** (`B17`). That is template text inherited from the autoharness
source repository.

**Verified at `e7e981d`: `intercom-go` has NO `src/` directory and no executable close-path
classifier.** `.autoharness/gates` holds only a log; Ship Step 6 invokes only
`autoharness gate pipeline-topology`. Classification in this workspace is therefore
**agent-driven from the skill markdown**, which is what makes the "0 source changes" claim
true and what makes C's markdown edits actually activate the path.

Both stale references fall **inside C's existing inventory** (`B9`, `B17`) and are
reconciled there — scoping them to *"workspaces with a Python implementation installed;
in this workspace the skill text is authoritative"*. No new site, no budget change.

Had the executable classifier actually been present, C's markdown-only edits would **not**
activate the path and §8's runtime proof would fail. Recording the verification is what
distinguishes a checked assumption from a lucky one. Asserted by **AC-16**.

## 3. Exact inventory — 16 clause sites + 2 additive

Line anchors from rev 6 Inventory Table §6.1-B; frontmatter and lines 10–32 spot-verified
against the installed file at `e7e981d`.

| # | Line | Clause | Class |
|---|---|---|---|
| **B1** | 3 | frontmatter *"**except for** the narrow, machine-verified P-015 fully-covered-root case"* | **MUST** |
| **B2** | 10–11 | *"run `mode: safe-close` **in place of** the destructive cascade"* | CONSISTENCY |
| **B3** | 14–15 | *"runs the P-015 … classification and, **only when** every precondition holds, delegates to the Cascade Close Sub-Procedure instead"* | **MUST** |
| **B4** | 28 | *"**instead of** the cascade … call, then `mode: post`"* | CONSISTENCY |
| **B5** | 30–32 | *"never calls the cascade op directly **unless** that classification confirms the narrow verified fully-covered-root exception"* | **MUST** |
| **B6** | 232–233 | *"It **never calls** the cascade `backlogit_ship_shipment`"* | CONSISTENCY |
| **B7** | 324 | pre-mode `RECONCILE_FAIL`: *"Do NOT call `backlogit_ship_shipment`"* | CONSISTENCY |
| **B8** | 356–358 | Safe-Close preamble: *"… Step 0 below, where the cascade op is **the** permitted close path"* | **MUST** |
| **B9** | 408–415 | **Step 0(c) "Classify the close path"** — classifier body | **MUST** |
| **B10** | 505–512 | **Step 0 close-path SELECTOR** — binary → **ternary** | **MUST** |
| **B12** | 618–620 | Cascade Sub-Procedure heading *"(… exception **ONLY**)"* | CONSISTENCY |
| **B13** | 698–709 | No-substitution rule — must also bind the **third** verdict | **MUST** |
| **B14** | 998 | lock-failure: *"Do NOT call `backlogit_ship_shipment`"* | CONSISTENCY |
| **B15** | 1015–1016 | Safety invariants — must enumerate the new path | **MUST** |
| **B16** | 1022 | *"… **never via the cascade op**"* | CONSISTENCY |
| **B17** | 1074 | references line naming P-015 as *"cascade prohibition"* | **MUST** |

| # | Site | Content |
|---|---|---|
| **B11** | 591–605 | Safe-Close **step 8** — fail-closed cross-reference; step 8's own text otherwise unchanged |
| **B18** | new section | **NEW** `TASK_ONLY_FINALIZE` section: token, classifier, `G0–G13`, `P1–P10`, failure split, bounded recovery, self-host ordering |

**16 clause sites (9 MUST + 7 CONSISTENCY) + `B11` + `B18`.** This is C's entire share.

**`B17` classification.** Rev 6 classes `B17` CONSISTENCY in Table §6.1-B but budgets it
as MUST-mechanical in §10.1. This plan follows **§10.1 (MUST)** — the stricter reading —
consistent with the deliberation §6 resolution.

### 3.1 Partition check

`A` takes `C1–C6` (6). `B` takes `A1–A6` (6). `C` takes `B1–B10`, `B12–B17` (16).
**6 + 6 + 16 = 28.** Exact. No duplicated obligation; no missing site.

## 4. What B18 contains

| Part | Content |
|---|---|
| **Token** | `capabilities:` frontmatter key advertising the exact-version token P-015 A7 requires (**CO-1**) |
| **Classifier** | Evaluates after `CASCADE`, before `SAFE_CLOSE`. Any failure/ambiguity/query error → not classified → falls through |
| **Selector** | Step 0 selector becomes ternary (`B10`) |
| **`G0–G13`** | Conjunctive, fail-closed pre-call guards |
| **`P1–P10`** | Post-call guards on the **RAW post-call state, BEFORE any archive restoration** (P-007 ordering, umbrella §8.2) |
| **Failure split** | Two separate paths (umbrella §6.5) |
| **Bounded recovery** | Enumerated-set-only restore; approval before any destructive step; byte-prefix append-only preservation; out-of-bounds paths reported-not-moved; sync-only rehydration (Probe 19) |
| **Digest binding** | §5 |
| **Self-host ordering** | §8 |

T1–T5 are **not** restated — they are B's surface. B18 **cites** P-015.

## 5. Engine digest binding (unchanged from rev 6)

Engine identity is **binding, not advisory**:

```text
backlogit 1.10.1-0.20260823032255-b07729386a31+dirty
SHA-256 1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98
```

`+dirty` is not a unique build identity, so **the digest binds**. **Mismatch → HALT +
full probe refresh.** The rev-4.1 advisory fallback stays **withdrawn**.

Invocation is **CLI-only**. No MCP equivalence is claimed; the registry advertises MCP
tools, and that claim is explicitly **not** made.

## 6. Ordering (test-first; `main` stays green)

1. **H0 — red.** One table-driven test function. Marker:
   `not implemented: task-only finalize contract`. Red count **1 of 1**; `go vet` clean;
   `go test ./...` non-zero.
2. **H1 — green.** Apply `B1–B18`. Re-run: `go vet` clean, `go test ./...` green,
   `markdownlint` clean.

Exactly **one** generated test function — the `021.001-T` constraint.

## 7. Verification obligations

### 7.1 What the Go test proves — text, parity, tokens

12 rows over `SKILL.md` (+ P-015 for parity):

| # | Assertion | Kind |
|---|---|---|
| 1 | Frontmatter advertises the capability token | positive |
| 2 | **Token parity**: the token string in `SKILL.md` **exactly equals** the one P-015 A7 requires | **parity (CO-1)** |
| 3 | Step 0 selector is **ternary**, ordered `CASCADE` → `TASK_ONLY_FINALIZE` → `SAFE_CLOSE` | **ordering** |
| 4 | `B18` states `G0`–`G13` | positive |
| 5 | `B18` states `P1`–`P10` | positive |
| 6 | `P1`–`P10` are stated to evaluate on the **raw post-call state before archive restoration** | **ordering** |
| 7 | `B18` states the two-path failure split | positive |
| 8 | `B18` states bounded recovery (enumerated-set-only; approval-before-destructive) | positive |
| 9 | `B18` states the digest binding and CLI-only invocation | positive |
| 10 | `B11` step-8 fail-closed cross-reference present; step 8's own text otherwise unchanged | **negative** |
| 11 | The **existing** fully-covered-root exception's skill-side handling is unmodified | **negative** |
| 12 | `SKILL.md` does **not** restate T1–T5 as its own authority (cites P-015 instead) | **negative** |

Rows 10–12 are the widening guards. Row 2 discharges **CO-1** from C's side.

### 7.2 What the Go test does **not** prove

It proves **contract text, parity and tokens**. It does **not** prove runtime behavior:
not that the classifier selects correctly, not that guards hold against the live engine,
not that recovery restores correctly. Those are **§7.3 durable runtime probes**. The
separation is stated exactly so no acceptance criterion conflates them.

### 7.3 Durable runtime probes — behavior

Retained from rev 6, committed and reproducible under
`docs/plans/evidence/2026-09-12-task-only-shipment-finalization/`:

| Probe | Proves | Status |
|---|---|---|
| **14** | dependency lifecycle end-to-end | committed, reused |
| **15** | EXACT `017-S` 12-member fixture, `INVARIANCE_FAILURES=0` | committed, reused |
| **16** | injected partial failure + bounded recovery | committed, reused |
| **18** | live-config inbound-edge byte-invariance | committed, reused |
| **19** | bounded recovery v2, both branches | committed, reused |
| **20** | post-archive closure route (O-6) | committed, reused |
| **21** | **NEW — activation probe** | see below |
| — | `clause-inventory.md` re-scan (§6.1-E) | regenerated by C |

Rev 4.1's `%TEMP%`-and-deleted probes 5/5b/10/12/13 remain **WITHDRAWN** as evidence.

**Probe 21 (new, and nearly free).** Before C's implementation merges, confirm the
classifier returns **not-selected** with the token absent; after merge, confirm it returns
`TASK_ONLY_FINALIZE`. The **post** half is discharged by **C's own closure** (§8), so
Probe 21 costs only the pre-merge negative observation. This is the activation evidence
for the whole three-shipment chain.

## 8. Self-hosting closure sequence — C is the proof

`023-S` manifest is **task-only** `[024.001-T]`. Covering feature `024-F` is live and
**outside** the manifest — exactly **T4**.

| Step | Session | Action |
|---|---|---|
| 1 | S1 | Orchestrator routes `023-S` **only after** `022-S` is archived `shipped`; claim; `024.001-T -> active` |
| 2 | S1 | **Probe 21 (pre)**: token absent ⇒ classifier does **not** select `TASK_ONLY_FINALIZE` |
| 3 | S1 | H0 red → implement `B1–B18` → H1 green; commit; push; PR; operator approval; merge |
| 4 | S1 | Verify merge SHA; `024.001-T -> done` |
| 5 | S1 | **Mandatory pre-self-close reload** of merged `main`: Ship instructions, `shipment-reconcile`, **and P-015** (A's `C1`). Positively verify merged tokens **and** the merge commit |
| 6 | S1 | **No Role Boundary change in this merge** ⇒ plan A §6 does **not** force a session end |
| 7 | S1 | **Step 6.1(a1) is NOT run** — `024-F` is **not** a manifest member (T4). No feature completion occurs |
| 8 | S1 | Pre-mode `expected_status: done` over the **task-only** manifest → `PROCEED` |
| 9 | S1 | Classification: `CASCADE` **not** applicable (zero feature members, T3) → `TASK_ONLY_FINALIZE` evaluated → token now **present** → `G0–G13` → **selected** |
| 10 | S1 | Digest check (§5). Mismatch ⇒ **HALT** |
| 11 | S1 | Authorized call; then `P1–P10` on the **raw** post-call state **before** archive restoration; then P-007; post-mode; sync |
| 12 | S1 | **Probe 21 (post)** discharged by this closure. Record the verdict as durable evidence |
| 13 | S1 | Operational closure + P-020; closure PR; operator merge |
| 14 | **Stage** | **Dispose of `024-F` (§8.3) — MANDATORY before `017-S` routes** |
| 15 | — | Orchestrator routes **`017-S`** |

### 8.3 Disposition of `024-F` — required, not a vague follow-up

After C closes task-only, `024-F` remains **live** with all descendants archived. Ship
Step 1's **P-001 gate** (line 308) reads: *"at most one top-level release unit may be in
flight. Check that no other top-level release units…"*. A live `024-F` can therefore be
read as an in-flight release unit and **block `017-S` from routing** — the chain would
close C successfully and then stall.

This is **not** deferred to an unscheduled follow-up. It is an ordered, owned step:

| Field | Value |
|---|---|
| **Owner** | **Stage** (feature lifecycle outside Ship's narrow a1 grant is Stage's) |
| **When** | Immediately after `023-S` is archived `shipped`, **before** `017-S` routes |
| **Operations** | `backlogit move 024-F --status done`, then `backlogit archive 024-F` — both already-authorized, non-destructive |
| **Precondition** | `024.001-T` archived; `024-F` has no live descendant |
| **P-001** | Discharges the gate by leaving **zero** live top-level units before `017-S` |

Ship never performs this: `024-F` is outside C's manifest, so A's a1 gate does not and
must not touch it (running a1 over `024-F` would violate **T4**, which requires the root
parent **live and outside** the manifest at classification time). Asserted by **AC-17**.

### 8.1 Why C may close in one session

Like B, C changes **no** Role Boundary. The authority-escalation rule binds **authority**
changes; C changes procedure. A's reload clause (`C1`) requires reloading the merged skill
and P-015 and **positively verifying the merged tokens**, which is exactly what makes the
newly-merged token visible at step 9.

### 8.2 If C's first real self-close fails after the implementation merged

The implementation is on `main`; only the **closure** failed. Do **not** revert. Follow
the umbrella plan §8.3 contingency and §6.5 bounded recovery: `P1–P10` evaluate on the raw
post-call state, recovery is enumerated-set-only with approval before any destructive step,
and out-of-bounds paths are **reported, never moved**. If recovery cannot complete within
its bound, **HALT** and return to Stage with the raw state preserved.

## 9. Intermediate and final state

| After | P-015 | Skill | Ship | Selectable verdicts |
|---|---|---|---|---|
| A | unchanged | unchanged | delegates | `CASCADE`, `SAFE_CLOSE`, `HALT` |
| B | authorizes, token-gated | unchanged (no token) | delegates | `CASCADE`, `SAFE_CLOSE`, `HALT` |
| **C** | authorizes, token-gated | **advertises token** | delegates | `CASCADE`, **`TASK_ONLY_FINALIZE`**, `SAFE_CLOSE`, `HALT` |

Every intermediate row is coherent and fail-closed. Activation happens **once**, at C.

## 10. Risks, residuals, rollback

| ID | Risk | Disposition |
|---|---|---|
| **R-1** | **Budget is tight (~106.5 min, ~13.5 min margin — smaller than `B18`'s own estimating variance)** | §12.1 valve, **honestly bounded**: it sheds only ~7.5 min and **cannot** rescue a `B18` or test overrun. That case is an explicit **HALT to Stage**. Risk is real and bounded to one shipment |
| **R-2** | `B18` sprawls beyond its allotment | `B18` **cites** T1–T5 rather than restating them (test row 12); the failure split and recovery are **carried from rev 6**, not re-derived |
| **R-3** | Token string drifts from P-015 (**CO-1**) | Test row 2 asserts **exact parity** against P-015. Any change routes through **Stage** into B's surface — never adjusted unilaterally here |
| **R-4** | C's closure fails, leaving the chain half-activated | §8.2. Implementation stays on `main`; only closure retries |
| **R-5** | `024-F` left live after task-only closure blocks `017-S` via Ship's P-001 gate | **§8.3** — an ordered, Stage-owned disposition step with named operations and preconditions, **before** `017-S` routes. Asserted by AC-17. *(Upgraded from a vague follow-up after review.)* |
| **R-6** | Engine digest mismatch at closure | **HALT + full probe refresh** (§5). Not advisory |
| **R-7** | Guard drift between `B18` and P-015 over time | Residual **accepted**; a CI coupling gate is out of scope. Test row 2 covers the token, not the full guard set |

### 10.1 Rollback — reverse dependency order is MANDATORY after activation

An earlier draft described each merge as independently revertible. **That is false once
the chain is activated**, and the correction matters:

| Revert | While installed | Result |
|---|---|---|
| **C** | — | Token disappears ⇒ B's gate returns dormant ⇒ `CASCADE`/`SAFE_CLOSE`/`HALT`. **Clean.** Does **not** undo an already-completed closure |
| **B** | C still installed | Skill advertises **and implements** `TASK_ONLY_FINALIZE` with **no P-015 authorization**. **PROHIBITED** |
| **A** | B and/or C installed | Restores Ship's stale binary restatement and the `children`-only drift, contradicting both newer surfaces. **PROHIBITED** |

**Rule: after activation, roll back in `C → B → A` order only.** Reverting B requires C
already reverted; reverting A requires B and C already reverted. Reverting A alone is
additionally undesirable on its own merits — it reintroduces the §2.2 coverage drift that
A fixed.

Before activation (i.e. between merges) each shipment is independently revertible, because
the downstream surfaces do not yet exist. The claim is **phase-relative**, and this table
states which phase it holds in.

### 10.2 Failure paths

| Failure | Response |
|---|---|
| MUST set alone cannot close in 2 h | **HALT, return to Stage.** Deferring a MUST site is **not** authorized |
| Token cannot be advertised machine-readably | HALT — falsifies B's A7 gate. Return to Stage; B's surface is re-planned |
| Digest mismatch | HALT + full probe refresh |
| Classifier selects `TASK_ONLY_FINALIZE` for a non-conforming manifest during Probe 21 | HALT — guards are insufficient. Return to Stage |
| `P1–P10` detect an out-of-bounds artifact post-call | Bounded recovery; approval before any destructive step; HALT if unbounded |

## 11. Acceptance criteria

- **AC-1** `B1–B18` all applied: 16 clause sites + `B11` + `B18`.
- **AC-2** Frontmatter advertises the exact-version token; **parity with P-015 A7 asserted**.
- **AC-3** Step 0 selector is ternary and correctly ordered.
- **AC-4** `B18` states `G0–G13`, `P1–P10`, the failure split, bounded recovery, digest binding, CLI-only invocation, self-host ordering.
- **AC-5** `P1–P10` evaluate on the **raw post-call state before** archive restoration.
- **AC-6** `B18` **cites** T1–T5 from P-015 and does **not** restate them as its own authority.
- **AC-7** `B11` cross-reference present; step 8's own text otherwise unchanged.
- **AC-8** The existing fully-covered-root exception's skill-side handling is unmodified.
- **AC-9** Exactly **one** Go test function; H0 red 1 of 1; H1 green; `go vet` clean.
- **AC-10** `markdownlint` clean.
- **AC-11** `workflow-policies.md` and `_ship.agent.md` are **unchanged** by this shipment.
- **AC-12** `clause-inventory.md` (§6.1-E) regenerated and committed. A site found by the re-scan and absent from §3 is **added, never skipped**; a new **MUST** site invalidates the budget and returns the unit to Stage.
- **AC-13** Probe 21 pre-half executed and recorded before merge.
- **AC-14** Closure classifies **`TASK_ONLY_FINALIZE`** and completes task-only; the verdict is recorded as durable evidence.
- **AC-15** Exactly 2 files changed (+ the evidence directory). 0 new production files.
- **AC-16** The absence of an executable close-path classifier in this workspace is **verified and recorded**; the stale `src/autoharness/...` references at `B9`/`B17` are reconciled (§2.2).
- **AC-17** `024-F` is disposed of per §8.3 by **Stage**, before `017-S` routes.
- **AC-18** The 7 CONSISTENCY edits are executed **last**; `B7` and `B12` are treated as **non-deferrable** (§12.1).

## 12. Sizing

**Size `M` · Complexity `high`.** Complexity carried as **enum-validated prose** —
`header-def.yaml` defines `size` on tasks but no `complexity` field.

| Work item | Estimate |
|---|---|
| 8 MUST clause edits (`B1 B3 B5 B8 B13 B15` + selectors `B9 B10`) | ~28.5 min |
| `B17` mechanical MUST | ~1.5 min |
| 7 CONSISTENCY edits (`B2 B4 B6 B7 B12 B14 B16`) | ~10.5 min |
| `B11` step-8 cross-reference | ~3 min |
| **`B18` new classification section** | **~28 min** |
| §6.1-E committed inventory re-scan | ~5 min |
| Go table-driven test (12 rows) + ≤4 helpers | ~20 min |
| H0→H1, `go vet`, `go test`, `markdownlint` | ~10 min |
| **Total** | **~106.5 min ≈ 1.8 h** |

Under the 2-hour rule with **~13 min** margin. This is the **tightest** of the three and
the plan says so rather than rounding it away.

`B18` is ~28 min rather than rev 6's ~30 because **T1–T5 moved to B**. The rest of rev 6's
`B18` content is carried, not re-derived.

### 12.1 Declared overflow valve (pre-authorized, and honestly bounded)

**Mandatory execution order.** The 7 CONSISTENCY edits are sequenced **LAST**, after
`B18` and after the Go test. The §12 cost table lists them earlier for *costing* clarity
only; it is **not** the execution order. A valve that sheds already-completed work is no
valve at all.

**Trigger.** At **100 minutes**, if the remaining CONSISTENCY edits are not yet done,
defer the **deferrable subset** to a follow-up hygiene stash entry.

**Deferrable subset — 5 clauses, ~7.5 min** (not 7):

| Clause | Why deferrable |
|---|---|
| `B2` (10–11) | describes what safe-close does *in place of* the cascade — branch-local, still true |
| `B4` (28) | same framing, "instead of the cascade … then `mode: post`" — branch-local |
| `B6` (232–233) | *"It never calls the cascade"* — explicitly scoped to safe-close mode |
| `B14` (998) | lock-failure prohibition — **no** close path is authorized without the lock, on any verdict |
| `B16` (1022) | *"never via the cascade op"* — scoped to safe-close's own archival |

**NOT deferrable (review correction):**

* **`B7` (324)** — pre-mode `RECONCILE_FAIL` prohibition. It asserts *sufficiency* of a
  single prohibition at a point that now precedes **three** verdicts. If a third verdict
  exists and `B7` still reads as a one-path statement, a reader could infer `RECONCILE_FAIL`
  blocks only the cascade.
* **`B12` (618–620)** — *"Cascade Sub-Procedure (… exception **ONLY**)"*. The literal word
  **ONLY** asserts **exclusivity**, which the amendment falsifies.

Both assert exclusivity or sufficiency rather than describing a branch, so deferring them
would leave the contract **incorrect**, not merely uneven — which violates the valve's own
safety predicate. They are reclassified as **non-deferrable** and treated as MUST for
scheduling.

**What the valve can and cannot rescue — stated plainly.** It sheds **~7.5 min**. It
therefore rescues an overrun in the *cheap early edits* only. **It cannot rescue a `B18`
or Go-test overrun**, because neither can be shed and both are followed by ~30 min of test
and verification. If `B18` is materially incomplete at 100 minutes, the valve is **not**
the answer: **HALT and return to Stage.**

**Deferring any MUST site is NOT authorized.** If the MUST set alone cannot close in 2 h,
**HALT and return to Stage**.

This valve is **reviewed in advance** so the implementer executes a decision rather than
improvising one under time pressure — and its **limits** are stated so it is not mistaken
for cover it does not provide.

## Plan Hardening

**Blast radius.** C is the **only** shipment of the three that changes runtime behavior.
A removes duplication; B lands dormant; **C activates**. All three shipments' risk
concentrates here — which is exactly why it is last, why it is smallest in surface (2
files), and why its activation is observable in one place (the token).

**Hardening 1 — the tight budget is the top risk, and the valve's limits are named.**
~13.5 min of margin is thinner than the estimating variance of `B18` (28 min) alone.
Review found the first draft's valve **mis-targeted** (it triggered on a `B18` overrun it
could not shed), **inoperable** (CONSISTENCY was costed before `B18`, so it would already
be done at the trigger), and **over-broad** (it deferred `B7`/`B12`, which assert
sufficiency and exclusivity rather than describing a branch). §12.1 now fixes all three:
CONSISTENCY is sequenced **last**, the deferrable subset is **5 clauses / ~7.5 min** with
per-clause justification, and a `B18` overrun is an explicit **HALT to Stage** rather than
a shed. Rev 6 failed by discovering its overrun **after** drafting; C names its overflow
behavior **and that behavior's limits** before starting.

**Hardening 5 — T4's live orphan is a design consequence with an owned disposition.** After
C closes task-only, `024-F` remains live with no live descendants. T4 **requires** the root
parent live and outside the manifest, so this is the contract working as specified — but
review established that a live `024-F` can trip Ship's P-001 single-in-flight gate (line
308) and **stall `017-S`**. §8.3 therefore upgrades it from a vague follow-up to an
**ordered, Stage-owned step** with named operations, preconditions and an acceptance
criterion, executed **before** `017-S` routes. It is deliberately **not** absorbed into
C's own scope: doing so would put a second live member in the manifest and destroy the
very T2/T4 shape C exists to prove.

**Hardening 2 — C closes by the path it ships, which is the strongest and riskiest proof.**
Strongest: no amount of contract text substitutes for an executed closure. Riskiest: a
guard defect surfaces at the moment of a destructive engine call. Mitigations are ordered
so detection precedes repair — `P1–P10` on the **raw** post-call state **before** archive
restoration (`AC-5`), bounded recovery with approval before any destructive step, and
out-of-bounds paths **reported, never moved**. §8.2 keeps the implementation on `main` and
retries only closure, so a closure failure never strands the merged contract.

**Hardening 3 — CO-1 is the only cross-shipment coupling, asserted from both sides.**
B's row 3 asserts the token requirement exists; C's row 2 asserts **exact parity**. A
mismatch fails C's harness at H1, **before** any closure attempt. Unilateral adjustment
inside C is forbidden: a token change is a **policy** change and routes through Stage into
B's surface. This prevents the failure where C "fixes" the mismatch by weakening B's gate.

**Hardening 4 — the guard set is inherited, not invented.** `G0–G13`, `P1–P10`, the failure
split and bounded recovery are **carried from the umbrella plan** (rev 6 §6.2–§6.5), which
survived six revisions and multiple adversarial rounds, with Probes 16 and 19 executed
against both recovery branches. C's job is **transcription into the skill**, not
re-derivation. That is what makes ~28 min for `B18` credible where inventing it would not be.

**Hardening 5 — T4's live orphan is a design consequence, stated not hidden.** After C
closes task-only, `024-F` remains live with no live children. T4 **requires** the root
parent live and outside the manifest, so this is the contract working as specified. It is
recorded as residual **R-5** with a **Stage** follow-up rather than quietly absorbed into
C's scope — absorbing it would put a second live member in the manifest and destroy the
very T2/T4 shape C is proving.

**Hardening 6 — what could make this plan wrong.** (a) If the token cannot be advertised
machine-readably, B's A7 gate is unimplementable and the chain re-plans — **falsified in
advance** at §2.1, risk **low**. (b) If the re-scan (`AC-12`) finds a **MUST** site absent
from §3, the budget is invalidated and the unit returns to Stage — the same tripwire rev 6
set, retained deliberately. (c) If the engine digest has moved, §5 halts before any call.
(d) If an **executable** classifier existed in this workspace, markdown-only edits would
not activate the path — **verified false** at §2.2 (no `src/`, no classifier), which is why
"0 source changes" holds here. Each failure mode has a **detector** and a **named owner**;
none degrades silently.

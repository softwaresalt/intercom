---
title: "Decided Plan — Ship covering-feature completion and close-path delegation (022-F)"
date: 2026-09-13
status: planned
agent: Stage
revision: 9
feature: 022-F
task: 022.001-T
shipment: 021-S
releases: 017-S
deliberation: docs/decisions/2026-09-13-intercom-go-a-only-simplification-decision.md
archived_full_plan: docs/archive/plans/2026-09-12-intercom-go-ship-feature-completion-foundation-plan.md
evidence: docs/plans/evidence/2026-09-12-task-only-shipment-finalization/probe25-root-included-cascade-fixture.txt
---

# Decided Plan — Ship covering-feature completion and close-path delegation (022-F)

> **This is the SOLE GOVERNING ARTIFACT for `021-S`.** It contains only the final executable
> contract. All review history, superseded revisions and withdrawn language live in the
> archived full plan named in `archived_full_plan:` and are **non-governing**.
>
> **It inherits the archive's plan-review LINEAGE in full.** Supersession transfers the
> contract *and* its review history: the pending gate here is **true-lineage attempt 5**, and
> the attempt-4 findings `H-1`…`H-9` are **carried and disposed in §13.2**. Superseding an
> artifact never discharges findings raised against it.

## 1. Objective

Grant Ship one narrow authority — complete a covering feature `active -> done` — and replace
Ship's **duplicated copy** of the close-path classification with **delegation** to the
installed policy and skill.

**Why it is required**: pre-mode runs `expected_status: done` over **every** manifest item, so
a fully-covered-root manifest cannot pass while its covering feature is still `active`. Ship's
Role Boundary today permits only task transitions. Independently, Ship's duplicated copy has
**already drifted unsafe** — it says *"every one of its children"* where the installed policy
requires descendants **at every depth**.

## 2. Surface — exactly 2 implementation files

| File | Change |
|---|---|
| `.github/agents/_ship.agent.md` | 8 clause sites (**CS1–CS8**, all mandatory) + 2 additive (**ADD-1**, **ADD-2**) |
| `tests/integration/ship_feature_completion_contract_test.go` | **new** — 18-row table-driven contract test |

**0 new production files.** Stage-authored planning artifacts (this plan, the Probe-25
evidence, `docs/memory/`, `docs/decisions/`, `.backlogit/` records) are committed separately
and are **not** counted against this 2-file surface.

> **Label note (traceability):** clause sites are **CS1–CS8** here. The archived plan called
> them `C1`–`C8`; `CSn` ≡ `Cn`. Renamed because the Probe-25 criterion tokens are also named
> `C1…C8` and the collision was a live mis-read hazard.

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
written into `_ship.agent.md`.

**`021-S` has no pre-archived descendants at any point.** Its manifest is
`[022-F, 022.001-T]`; `022.001-T` reaches `done` at step 4 and is archive-located by routing,
which makes it **completed**, not pre-archived.

### 3.4 Delegation and output validation

**Authority split:** the **installed `shipment-reconcile` skill is the RUNTIME authority**;
**P-015 is the STATICALLY REVIEWED authority**. Ship never adjudicates between them, never
re-derives a classification, and performs **no** runtime `policy_id`, policy-version or
source-disagreement check.

**Installed contract — linked, not restated:**

| Subject | Location |
|---|---|
| Close-path policy | `.github/policies/workflow-policies.md` → `## P-015: Single-Artifact Shipment Closure (No Cascade Ship)` (L410) |
| Pre-mode contract | `.github/skills/shipment-reconcile/SKILL.md` → `### Pre-Mode` (L257) |
| Cascade close procedure | same file → `### Cascade Close Sub-Procedure` (L618) |

**`_ship.agent.md` MAY contain an output-validation allowlist of the installed CLOSE VERDICT
tokens — `CASCADE` and `SAFE_CLOSE` — and nothing more.** It **MUST NOT** contain any
classification predicate: no root/`parent_id` test, no coverage test, no per-member rule, no
fallback rule. Comparing a returned token against a list is validation; deciding which token
applies is classification, and stays delegated.

* **`HALT` and `RECONCILE_FAIL` are non-close procedural outcomes**, not close verdicts. They
  authorize no close path.
* **`BLOCKED` is not an installed token** and is handled as unknown.

**Ship HALTs, no mutation, on any of:** a token outside the close-verdict allowlist (including
`BLOCKED` and any future token); an absent, empty, unparseable or ambiguous result; a
classifier that is not installed or not readable; or a non-close procedural outcome.

The allowlist is bound to the **currently installed** skill. Widening it requires its own
review; this plan authorizes **no new verdict and no new close path**.

### 3.5 Release-instance execution contract (this plan only)

`_ship.agent.md` carries **no shipment IDs and no expected-verdict rule** — only the general
delegation above. The expectation below is a **precondition of this plan and its records**:

> For **`021-S`** and **`017-S`**, the installed classification is **expected to return
> `CASCADE`**. `CASCADE` enters the **pre-existing** cascade branch. **`SAFE_CLOSE`, an
> unknown token, an unavailable or unreadable classifier, an empty or ambiguous result,
> `HALT`, or `RECONCILE_FAIL` ⇒ HALT BEFORE ANY CLOSE MUTATION** and return to Stage.
> **There is no `SAFE_CLOSE` fallback on these routes.**

A `SAFE_CLOSE` here means the live manifest is not the shape this plan asserts — a planning
error only Stage can reconcile.

### 3.6 Authority-escalation rule

A context reload MUST NOT widen the authority envelope of the session performing it. The
session that merges this change **may not use the authority it introduces**; it checkpoints
and ends.

## 4. Clause inventory — 8 sites (all mandatory) + 2 additive

Apply by **anchor text**, never by line ordinal.

| Site | Line | Change |
|---|---|---|
| **CS1** | 756–764 | Pre-self-close reload: also reload **P-015**; verify merged tokens **and the merge commit**; add the §3.6 carve-out, scoped to **Role Boundary** changes only |
| **CS2** | 789–791 | Delete the `backlogit move <shipment_id> --status shipped` prescription (the engine refuses it for shipments, exit 9). Replace with a pointer to the skill's authoritative close sequence |
| **CS3** | 794–795 | Replace the named single exception with: do not call it unless **the machine-selected verdict** authorizes it |
| **CS4** | 802–823 | **Delete the re-derived classification prose.** Replace with generic delegation |
| **CS5** | 821–823 | Remove the *"in place of the safe-close sequence"* single-exception framing (subsumed by CS4) |
| **CS6** | new bullet | **NEW** delegation bullet: Ship invokes whichever close path the verdict names, and no path absent from the verdict |
| **CS7** | 255–256 | **Pointer-only delegation.** Replace the queue-only summary with a pointer to the installed `mode: pre` contract at the `expected_status` already named one line above; continue **only** on an authoritative `PROCEED`. **PRESERVE** the orphan scan (already present) and the adjacent `Scope note (139-F/139.001-T)`; **ADD** the `record-consistent` conjunct (**not** currently in the file) |
| **CS8** | 775–777 | Same pointer-only delegation at `expected_status: done`. **PRESERVE** the single-writer lock clause and the orphan scan (both already present); **ADD** the `record-consistent` conjunct |
| **ADD-1** | 38 | Role Boundary row — the §3.2 narrow grant |
| **ADD-2** | new | **Step 6.1(a1)** — the covering-feature completion gate (§5) |

**CS7/CS8 MUST NOT state any per-item classification predicate — locational or
declared-status.** This plan asserts nothing about the installed pre-mode's internal taxonomy;
it depends only on the **outcome** `PROCEED`.

### 4.1 Pinned replacement wording (so the test literals are reachable)

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
| **S0** | **Anomaly gate, before `n` is computed.** Any member with missing/unresolvable `artifact_type`; present in **both** queue and archive; present in **neither**; **declaring `status: archived`** with malformed provenance; or **declaring `status: archived`** and **not** on the §3.3 allowlist | **HALT**, no mutation |
| **S1** | `n == 0` (no feature member) | Record `A1_NOT_APPLICABLE`; proceed to `a`. **A success, not a skip** |
| **S2** | `n > 1` | Record `A1_MULTIPLE_FEATURE_MEMBERS: {ids}`. **Evaluate no condition, mutate nothing.** Return to Stage |
| **S3** | `n == 1`, live status **`done`** | Re-evaluate conditions 1–4. All hold ⇒ record `A1_ALREADY_DONE`, proceed to `a` without re-issuing the transition. Any fail ⇒ **HALT** |
| **S4** | `n == 1`, live status **`active`** | All five §3.2 conditions hold ⇒ `active -> done` via **`backlogit_move_item`** (CLI fallback `backlogit move {id} --status done`), record the transition. Any unmet ⇒ **HALT**, naming the condition |
| **S5** | `n == 1`, any other status | **HALT**, record the observed status, return to Stage |

`n` = count of manifest members whose `artifact_type` is `feature`, **never** by ID suffix.

**S0 explicitly does NOT halt on** an archive-located member declaring `status: done`. That is
a normal completed member: no allowlist check, no provenance requirement.

**Executable surface**: containment via `backlogit doctor` (`check_duplicates`,
`check_orphans`) or `backlogit_query_sql` — **never** `get_item` alone, which cannot reveal a
torn duplicate. Declared status / `artifact_type` / `parent_id` / provenance via
`backlogit_get_item`.

## 6. Go contract test — 18 rows

One table-driven function, ≤4 helpers, **stdlib assertions only** (`testing`, `strings`), no
`testify`, reusing the package-level `repoRoot(t)` from `tests/integration/build_script_test.go`.

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

**Kinds**: `compound +/−` = {9, 10, 14, 15} (**4**); `compound ++` = {16}; `negative` = {7, 8}.

### 6.1 Pinned literals (normative — all single-line, CRLF-safe)

| Row | Class | MUST BE ABSENT | MUST BE PRESENT |
|---|---|---|---|
| 14 | transition | ``every manifest item is present in `.backlogit/queue/` with the`` (L255, ×1) | ``defers every per-item status decision to that skill's `mode: pre` classification`` (×0 today) |
| 15 | transition | ``queue with `status: done`, and scans for orphan items.`` (L777, ×1) | ``defers every per-item status decision to the `mode: pre` classification at `expected_status: done``` (×0 today) |
| 9 | transition | ``qualification is never per-member, and no feature ID is ever special-cased.`` (L819, ×1) | ``HALT on any result token the installed classification did not name`` (×0 today) |
| 10 | transition | ``<shipment_id> --status shipped`` (L790, ×1) | ``the authoritative close sequence defined by the `shipment-reconcile` skill`` (×0 today) |
| 7 | transition | ``it is fully covered (every one of its children,`` (L810, ×1) | — (ABSENT-only) |
| 16 | transition | — | ``evaluated before `n` is computed`` **and** ``halts with no feature mutation`` (both ×0 today) |
| 8 | **guard** (regression) | ``TASK_ONLY_FINALIZE`` — **already ×0; must stay ×0** | — |
| 17 | **guard** (preservation) | — | ``scans for orphan items`` — **already ×2 (L256, L777); must stay ≥2** |
| 18 | **guard** (preservation) | — | ``Scope note (139-F/139.001-T)`` — **already ×1 (L259); must stay ×1** |

**Measured at the current tree — the rows fall into TWO classes, and only one is red → green:**

| Class | Rows | Current counts | H0 state |
|---|---|---|---|
| **Transition rows** — must change | **9, 10, 14, 15** (`compound +/−`); **7** (ABSENT-only); **16** (`compound ++`) | 9/10/14/15: ABSENT = **1** each, PRESENT = **0** each · 7: ABSENT = **1** · 16: both PRESENT needles = **0** | **RED.** None can be vacuously green |
| **Guard rows** — must NOT change | **8** (regression), **17** (preservation), **18** (preservation), plus the `018.` decoupling guard | 8: `TASK_ONLY_FINALIZE` = **0** · 17: `scans for orphan items` = **2** (L256, L777) · 18: `Scope note (139-F/139.001-T)` = **1** (L259) · `018.` = **0** | **GREEN today, by design.** They assert a property that already holds and must survive |

**The guard rows are green at H0 and that is correct** — their purpose is to fail if the
implementation *breaks* something, not to record a transition. Row 3's `a1` anchor is absent
pre-implementation **by design**; that absence is the H0 readiness guard.

**Only the transition rows constitute the red → green evidence** for AC-3/AC-8. Do not read
the guard rows as proof the contract changed.

**Negative decoupling guard**: the test also asserts `_ship.agent.md` contains **no `018.`
literal**, so the release-instance allowlist of §3.3 cannot silently migrate into the global
contract.

## 7. Ordering

**Test-first, `main` stays green.** H0: write the 18-row test → **red**. H1: apply ADD-1,
ADD-2 and **CS1–CS8** → **green**. All eight clause sites are required to reach H1.

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
the pre-mode and close-path decisions; the **real engine cascade invocation begins at
transcript `### STEP 8: AUTHORIZED CALL`**, and criteria **4–8** observe that call's real
output and post-call state. **C1** is a real digest of the resolved executable. The central
claim — the cascade archives exactly the manifest and preserves every `parent_id` — rests on
the **engine-observed** criteria.

**What it does not prove**: it measures the engine on a faithful fixture. It is not a claim
about the live records at closure time, which is why §9 step 12a is an **exact** topology match
that HALTs on mismatch.

### 8.1 Evidence identity — ONE scheme, non-self-referential

Identity is established by **Git blob object IDs**, which are byte-exact, tool-native and
reproducible. **No digest is stored inside the file it identifies.**

```text
evidence_commit_sha := git log -1 --format=%H -- <ps1> <txt>
committed_blob(f)   := git rev-parse {evidence_commit_sha}:{f}
current_blob(f)     := git hash-object -- {f}
file identity holds <=> committed_blob(f) == current_blob(f)   for BOTH files
engine_expected     := ENGINE_SHA256_EXPECTED read from the committed <txt>
engine_current      := SHA-256 of (Get-Command backlogit).Source
engine identity     <=> engine_expected == engine_current
```

The engine digest **is** read from the transcript, correctly: it asserts something about an
**external** object, so no self-reference arises.

**Measured**: `evidence_commit_sha = e36d853`; both blobs match
(`81d85454d920`, `334cf3be007c`); `git diff --quiet HEAD` clean; engine
`1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98` matches.

**`evidence_commit_sha` is the LAST commit touching either file** — not the introducing one.
That is what closes post-merge replacement: replaced evidence moves the sha to a **descendant**
of the merge, failing the ancestry check.

## 9. Closure sequence

`021-S` = `[022-F, 022.001-T]` — a fully-covered root (`022-F` has no `parent_id`; its complete
descendant set is exactly `{022.001-T}`; the manifest contains nothing else).

| Step | Session | Action |
|---|---|---|
| 0a | **Stage** | Probe 25 authored, executed, committed. **DISCHARGED.** `021-S` not claimable until PASSing |
| 1 | **S1** | **Step 0.5 Shipment Intake.** Verify on `main`; **P-011** branch-before-mutation + **P-016** single-worktree check; pre-claim topology gate; **claim `021-S`** |
| **1a** | S1 | **POST-CLAIM CONDITION — `022-F` must be `active` (checked, not assumed).** Re-read `022-F` after the claim and require its live status to be **exactly `active`**. **Any other value ⇒ HALT** and return to Stage. **Ship is granted NO authority to transition a feature `queued -> active`** — §3.2 grants exactly one transition, `active -> done`. *(Live state at planning time is `queued`; the shipment claim is what is expected to activate it. If it does not, that is a planning-state error only Stage can reconcile.)* |
| **1b** | S1 | **Step 2 Harness Generation (P-002 / P-004).** The **harness-architect** produces the 18-row harness: `go vet ./...` exits 0, `go test ./...` exits non-zero with expected failure markers ⇒ **H0 RED CONFIRMED** ⇒ `harness-ready` applied to `022.001-T`. **This precedes the TASK claim** — P-002 permits claiming a task only after red-phase confirmation |
| **1c** | S1 | **Step 3 Build Ready Queue** — filtered to tasks carrying `harness-ready` |
| **1d** | S1 | **Step 4.1 Claim Task** — `022.001-T -> active` |
| 2 | S1 | **Step 4.2 implement** → apply ADD-1, ADD-2 and **CS1–CS8** → **H1 green** |
| 3 | S1 | Quality gates; review gate |
| 4 | S1 | Complete Task — commit the implementation, then **`022.001-T -> done`**, then **commit the resulting `.backlogit/` change** and verify a clean worktree. **Both commits must be in the PR that merges at step 5** — otherwise merged `main` still shows the task `queued` and a1 halts at step 13 |
| 5 | S1 | PR lifecycle — gates, build, push, PR, **P-018 Copilot engagement**, operator approval, **merge** |
| 6 | S1 | Merge Confirmation Gate — `gh pr view` `MERGED`; `git fetch origin main`; `git merge-base --is-ancestor {merge_sha} origin/main` |
| 7 | S1 | Reload merged `main`; detect the Role Boundary change; per §3.6 the new authority is **not** available to this session |
| 8 | S1 | **Write checkpoint** (§9.1) and **END**. Record the emitted filename |
| 9 | **S2** | **Fresh session.** §9.1 items **1–6** — enumerate, anomaly gate, MATCH-set filter, restore, Engram prune/gate, **resume**. **All reads and in-memory work; NO mutation.** The **resolve is deferred to step 11a** |
| 10 | S2 | Verify merge SHA is an ancestor of `origin/main`; validate the same active shipment. **Reads only** |
| 11 | S2 | **Post-Merge Branch Protocol (P-011)** — `git checkout main`, pull, `git checkout -b post-merge/022-ship-feature-completion`. **This PRECEDES EVERY S2 MUTATION**, including the checkpoint resolve at 11a. `main` is never the active branch for any mutation |
| **11a** | S2 | **Resolve the checkpoint** (§9.1 item 7) — the **first** S2 mutation, performed **on the closure branch**, and only after the resume at step 9 was confirmed successful |
| 12 | S2 | `--phase lifecycle` topology gate — **while `021-S` is still `active`**; never re-run after the close |
| **12a** | S2 | **READ-ONLY re-verification** of Stage's committed evidence (§9.2). FAIL ⇒ HALT, `021-S` stays `active`, no a1 mutation |
| 13 | S2 | **Step 6.1(a1)** — §5 **S4**. `022.001-T` declares `done` (archive-located = completed, not pre-archived) ⇒ all five conditions hold ⇒ `022-F active -> done` via `backlogit_move_item` |
| **13a** | S2 | **PRE-CASCADE BASELINE COMMIT (mandatory).** Commit a1's `.backlogit/` mutation **on the existing closure branch** — no new branch, no new worktree. **`{pre_cascade_sha}` := this commit.** Then satisfy §10.3 step 1 **against this commit**: verify `git status --porcelain -- .backlogit/queue/ .backlogit/archive/` is **empty**, and capture `{pre_paths}` from that clean baseline. **Without 13a the trees are dirty from step 13 and the cascade MUST NOT be invoked** |
| 14 | S2 | **Pre-mode**, `expected_status: done`. **Requires an authoritative `PROCEED`**; anything else HALTs |
| 15 | S2 | Classification **must return `CASCADE`** (§3.5) ⇒ close via the **pre-existing** cascade op; P-007 archive-integrity verify; post-mode. **The cascade output is committed ONLY after every postcheck passes** (§10.3) |
| 16 | S2 | Operational closure → `docs/closure/`; P-020 compact-context — on the closure branch, before the closure PR is pushed |
| 17 | S2 | Sync — **MCP `backlogit_sync_index` first**, CLI `backlogit sync` as declared fallback — then push, closure PR, **P-018 Copilot engagement**, local review, operator approval, merge |
| 18 | S2 | Return to `main`; pull |
| 19 | — | Orchestrator routes **`017-S`** |

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

### 9.1 S2 checkpoint recovery — exact identity

**S1 records** the emitted `checkpoint-YYYYMMDD-HHMMSS.json` filename **and** its `session_id`
in the handoff. **S2 carries both as the expected identity.**

1. **Enumerate ALL** via `backlogit_list_checkpoints` — **pass NO parameters**. **Do not pass
   `consumer_id`, and do not use the CLI `--agent` flag: both are FILTERS.** The tool manifest
   describes `consumer_id` as *"Filter by consumer/agent ID"*. A filtered enumeration hides
   records and silently disables the anomaly gate and the leftover warning below.

   **The durable rule is "pass no parameters"; the counts below are a dated observation, not
   an invariant.** Checkpoint totals grow every session, so no absolute count may be asserted
   as a live expectation or used as an identity.

   | Observed at | unfiltered | `--agent ship` | `--agent stage` |
   |---|---|---|---|
   | 2026-09-13 (attempt-4 review) | 15 | 1 | 14 |
   | 2026-09-13 (attempt-5 remediation, re-measured) | **17** | **1** | **16** |

   The filter behaviour — *strictly fewer records returned than unfiltered* — is the invariant;
   **an executor MUST re-derive it at run time and MUST NOT compare against either row.**
2. **Anomaly gate FIRST**, over the full enumeration, before any filtering or counting.
   The signal is the **top-level counters** — there is no per-summary quarantine flag:
   `needs_quarantine > 0`, `quarantined > 0`, or `total != checkpoints.length` ⇒ **FAIL
   CLOSED**. Additionally, **any record with a missing or malformed required field** (notably
   an empty `agent` or `status`) ⇒ **FAIL CLOSED**.
3. **Filter to the MATCH set** — exact filename **AND** `agent: ship` **AND** the exact
   `session_id` (all three are exposed in list summaries).
4. **Require exactly ONE MATCH.** **Zero ⇒ FAIL CLOSED** (S2 presupposes S1's checkpoint; the
   installed "zero candidates is normal startup" continuation does not apply). **More than one
   ⇒ FAIL CLOSED** — ambiguity is never resolved by picking.
5. **Non-matching active records are a WARNING**, surfaced in the handoff report and otherwise
   ignored. They never strand a valid handoff, and S2 **never** resolves, prunes or touches
   them — least of all a `stage`-owned record.
6. **Restore → Engram prune/gate → resume.** Engram unreachable ⇒ **FAIL CLOSED**, no prune,
   no resume. A file-based prune degradation is not permitted. **Everything through this item
   is reads and in-memory work — no mutation.**
7. **Resolve — deferred until AFTER the closure branch exists (P-011).**
   `backlogit_resolve_checkpoint` is a **backlog mutation**, so it runs at **§9 step 11a**,
   **on the closure branch**, never on `main` and never before §9 step 11 creates that branch.
   It still runs **only after a confirmed successful resume** (item 6), for **that one
   checkpoint only** — the resume cursor and checkpoint identity semantics are unchanged; only
   the *position* of the write moved.

**Checkpoint payload**: `schema_version: 1`, `agent: ship`, `session_id`, `phase`,
`resume_hint` naming `021-S`, `022-F`, `022.001-T` and the merge SHA; domain data under
`context`. Write it with the official create operation, passing the state dump **inline**.

### 9.2 S2 read-only re-verification — six checks

**S2 authors nothing, modifies nothing, re-runs nothing.** Every check is a read
(`git log`, `git rev-parse`, `git hash-object`, `git merge-base`, `git diff --quiet`,
`Get-FileHash`, text inspection).

1. **Exact topology** — `TOPOLOGY_MEMBERS=13`, `ROOT_INCLUDED=True`, `PRE_ARCHIVED=11`,
   `ARCHIVED_STATUS_FLAVOUR=queued`. Exact match; any other topology means the measurement is
   not about `017-S`.
2. **PASS fields** — `PROBE25_RESULT=PASS`, `FAILED_CRITERIA=0`, and all **nine pinned STEP 11
   tokens** (§8) read `PASS`.
3. **Ancestry** — both files tracked, and `evidence_commit_sha` is an **ancestor of the merge
   commit** verified at step 6.
4. **Engine identity** — the currently-resolved engine digest equals `ENGINE_SHA256_EXPECTED`
   from the committed transcript. **No absolute path is required.**
5. **File identity** — both committed blobs equal their working-tree files (§8.1), hashed
   **separately**; neither file missing, empty or modified.
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
   > — per §4, this plan depends only on the **outcome** `PROCEED`.

## 10. Recovery

### 10.1 HALTs before the CLOSE — nothing to unwind, but a1 may already have run

a1 S0–S5; pre-mode `RECONCILE_FAIL`; a verdict other than `CASCADE` on these routes; §9.2
failure. In every case **no CLOSE mutation has occurred**: `021-S` stays `active`, is not
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
| **Transient — toolchain/executable drift** | **4** | **YES** — only the environment moved | **Correct the drift and re-verify read-only, in place.** A clean re-verification **clears it** and A proceeds |
| **Committed evidence / identity / ancestry** | **1, 2, 3, 5, 6** | **NO** — the artifact is absent, altered, inconsistent or misplaced | **Cannot be cleared in place.** Requires **`git revert` of the merge** (after holding `017-S`) **and a re-plan** |

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

**Obligation, in order:**

1. **Capture is a precondition, not a reaction — and it is §9 step 13a.** The baseline is
   established by the **named pre-cascade baseline commit at §9 step 13a**, on the **existing**
   closure branch. **If any of the following is unavailable, do not invoke the cascade.**
   a. **`{pre_cascade_sha}`** — the step-13a commit. It exists specifically because §9 step 13
      (a1's `active -> done`) dirties `.backlogit/`, so **`HEAD` before 13a is NOT the
      pre-cascade state** and restoring from it would discard a1's valid result.
   b. **Clean trees, verified against that commit** —
      `git status --porcelain -- .backlogit/queue/ .backlogit/archive/` returns **empty** at
      `{pre_cascade_sha}`. A dirty baseline makes "what the cascade created" undecidable.
   c. **A complete path inventory `{pre_paths}`** — the full sorted path list of both trees as
      of `{pre_cascade_sha}`, captured alongside the skill's in-memory Step 0(b) snapshot of
      declared `status`, `parent_id` and location for every watched record.
2. **Restore**, in this order:
   * `git restore --source {pre_cascade_sha} -- .backlogit/queue/ .backlogit/archive/`
     — restores **tracked** content; and
   * **delete exactly `{post_paths} − {pre_paths}`** — the paths present after the call and
     absent from the inventory. This is the deterministic set the cascade created (most
     notably the newly written archived shipment record), which `git restore` alone leaves
     behind. **Delete nothing else**; step 1a guarantees this set contains no pre-existing
     untracked file.
3. **VERIFY the restore — both halves:**
   * **Tree equivalence** — the current sorted path list equals `{pre_paths}`, and
     `git status --porcelain` over both trees is **empty** (content identical to
     `{pre_cascade_sha}`); **and**
   * **Record equivalence** — every record in the Step 0(b) snapshot matches it on declared
     `status`, `parent_id` and location.

   **An unverified restore is not a restore.**
4. **HALT.** No retry, no second close path, no re-invocation of the cascade.

> **The cascade result is NEVER committed before its postchecks pass (F6 — bounded state
> machine).** The authorized flow is strictly:
>
> ```text
> 13a commit baseline  ->  {pre_cascade_sha}, clean trees, {pre_paths}
> 15  invoke cascade   ->  mutates the WORKTREE only
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
| **Rollback** | §10.3 — pre-cascade commit + snapshot restore, **verified**, then HALT | same |
| **approval_required** | **YES** (non-negotiable) | **YES** |
| **ActionRisk** | **`destructive`** | **`destructive`** |
| **ActionResult** | **`approved`** | **`approved`** |

**`017-S` manifest (13)**: `018-F`, `018.008-T`, `018.001-T`, `018.001.001-ST`,
`018.001.002-ST`, `018.001.003-ST`, `018.002-T`, `018.002.001-ST`, `018.002.002-ST`,
`018.002.003-ST`, `018.003-T`, `018.003.001-ST`, `018.003.002-ST`.

**Conditions — all must hold at invocation:**

1. the installed classification returns **`CASCADE`**; every other outcome (§3.5) **voids the
   approval and HALTs**;
2. **all preflight gates pass** — §9.2's six checks, the §5 S0–S5 selector, and pre-mode
   `PROCEED`;
3. the **manifest and topology are exactly as recorded**, unchanged — `021-S`: 2 members, no
   dependencies; `017-S`: the 13 members above, `018-F` a root, all 12 descendants resolving to
   it, dependency `[{id: 021-S, type: blocks}]`;
4. **`PA-017-CASCADE` only** — the dependency is **satisfied by a shipped `021-S`**. This
   approval is **void while `021-S` is unshipped**;
5. the **clean trees, pre-cascade commit, path inventory and Step 0(b) snapshot are all
   captured and available** (§10.3 step 1);
6. **no scope expansion** — nothing outside the manifest **and the shipment's own record** is
   archived, reparented, created or deleted.

**Post-mutation failure ⇒ restore, VERIFY, HALT** (§10.3).

## 12. Acceptance criteria

1. **AC-1** Role Boundary carries the §3.2 grant with all five conditions, fail-closed.
2. **AC-2** Step 6.1(a1) exists, positioned after `a0` and immediately before `a`.
3. **AC-3** All **eight** clause sites CS1–CS8 are applied; the 18-row test passes; H0 was red.
4. **AC-4** `_ship.agent.md` states **no classification predicate**, and **no `018.` literal**.
5. **AC-5** The close-verdict allowlist is exactly `CASCADE` / `SAFE_CLOSE`; `HALT` and
   `RECONCILE_FAIL` are non-close outcomes; unknown tokens (incl. `BLOCKED`), absent, empty,
   ambiguous or unreadable results **HALT**. **No runtime `policy_id`, version or
   source-disagreement check exists.**
6. **AC-6** No lifecycle class is determined by location (§3.1); containment is preserved in
   full; archive placement applies only after declared `status: archived`.
7. **AC-7** The §3.3 exemption is triggered only by declared `status: archived`, applies only
   to the 11 enumerated IDs, requires all five gates, and lives **only in this plan**.
8. **AC-8** CS7/CS8 are pointer-only; they **preserve** the single-writer lock, the orphan
   scan and the `Scope note` (all three already present in the file), **add** the
   `record-consistent` conjunct (not previously present), and state no per-item predicate.
   Rows 14/15/17/18 pass.
9. **AC-9** a1 runs **before** pre-mode and is self-sufficient; all evaluation (S0–S3, S5) is
   read-only and its **only** write is S4's single `active -> done` transition; the S0–S5
   selector is total and disjoint. Row 16 passes.
10. **AC-10** Probe 25 is committed and PASSing **before `021-S` is claimed**; §9.2's six
    checks pass at step 12a; every check is a read; S2 writes nothing to `docs/plans/`.
11. **AC-11** Exactly **2 implementation files**, **0 new production files**, **0 new module
    dependencies**.
12. **AC-12** The two sessions are distinct: S1 never performs a1 or the close.
13. **AC-13** S2's recovery follows §9.1 exactly — unfiltered enumeration, anomaly gate first
    on the top-level counters, then MATCH-set filtering requiring exactly one.
14. **AC-14** For `021-S`/`017-S` a non-`CASCADE` outcome **HALTs before any close mutation**;
    no `SAFE_CLOSE` fallback.
15. **AC-15** Every post-CASCADE failure performs the §10.3 **verified** restore — tree
    equivalence **and** record equivalence — then HALTs. The cascade is not invoked unless
    §10.3 step 1's clean trees, pre-cascade commit, path inventory and Step 0(b) snapshot all
    exist.
16. **AC-16** Both `PA-*` records are present with `ActionRisk: destructive`,
    `ActionResult: approved`, the operator timestamp, and all applicable conditions.
17. **AC-17** The lifecycle order of §9 holds; the `--phase lifecycle` gate never runs after
    the archive.
18. **AC-18** P-018 engagement precedes **both** merges.
19. **AC-19** **(F3)** After claiming `021-S`, `022-F`'s live status is **re-read** and
    required to be **exactly `active`**; any other value **HALTs**. Ship performs no
    `queued -> active` transition on a feature.
20. **AC-20** **(F4, P-002/P-004)** The **task** claim `022.001-T -> active` occurs **after**
    harness-architect has confirmed the red phase and applied `harness-ready` — `go vet ./...`
    exits 0 and `go test ./...` fails with expected markers. *(The **shipment** claim correctly
    precedes harness generation, per the installed Step 0.5 → Step 2 → Step 4.1 order.)*
21. **AC-21** **(F5, P-011)** The post-merge closure branch exists **before any S2 mutation**,
    including `backlogit_resolve_checkpoint`, which runs at §9 step 11a on that branch. No S2
    mutation occurs on `main`. Resume-before-resolve semantics are unchanged.
22. **AC-22** **(F1)** §9 step **13a** commits a1's mutation on the **existing** closure
    branch, designates that commit `{pre_cascade_sha}`, and verifies clean
    `.backlogit/queue/` + `.backlogit/archive/` trees and captures `{pre_paths}` **against that
    commit**, before pre-mode or the cascade runs. No second branch and no second worktree is
    created. §10.3's restore uses **exactly** this baseline.
23. **AC-23** **(F6)** The cascade result is **not committed until every postcheck passes**; a
    committed-cascade-awaiting-postchecks state is unreachable, and if observed is an
    out-of-contract HALT to the operator rather than a §10.3 restore.

## 13. Gate state

```text
plan-review-attempt: 5
dispatch_mode: multi-agent
decision: PENDING
```

### 13.1 Review-lineage correction (TRUE LINEAGE — do not renumber)

**This gate is TRUE-LINEAGE ATTEMPT 5.** The counter above is corrected, not reset.

When the append-only foundation plan was compacted into this decided plan, the
`plan-review-attempt` marker **restarted at 1** and a subsequent edit bumped it to **2**. That
was a **counter reset across an artifact rename**, not a new lineage: the plan-review history
follows the *plan*, not the *filename*. This plan declares itself the **sole governing
artifact** and makes the archive non-governing, so it inherits the archive's review lineage
**in full**.

| True attempt | Revision | Artifact | Decision |
|---|---|---|---|
| 1 | rev 4 | foundation plan (now archived) | **FAIL** |
| 2 | rev 5 | foundation plan | **FAIL** |
| 3 | rev 6 | foundation plan | **FAIL** — re-entry cycles exhausted, **ESCALATED** |
| 4 | rev 7 | foundation plan | **FAIL** — operator-authorized exceptional cycle, budget **SPENT** (`H-1`…`H-9`) |
| **5** | **rev 8** (this plan) | **this decided plan** | **PENDING** — this gate |

The archived plan's markers read `1, 2, 3, 3, 4` (attempt 3 appears twice: the gate marker and
its remediation-disposition echo). The two values `1` and `2` that appeared in this file were
**never independent attempts** and are superseded by the corrected counter above.

**Authorization for this attempt.** Attempt 5 runs under an **explicit operator
authorization** naming it as true-lineage attempt 5 and authorizing the carry-and-remediate of
`H-1`…`H-9`, the counter correction and the backlog resynchronization. **This authorization
covers exactly ONE exceptional attempt.** A further remediation-and-review cycle is **not**
self-authorized: if this gate returns FAIL, Stage halts and returns to the operator.

### 13.2 Attempt-4 findings `H-1`…`H-9` — CARRIED and DISPOSED

Attempt 4 returned **0 P0 / 9 P1** against rev 7. Those findings were raised against the
foundation plan, which this plan supersedes; **relocation is not remediation**, so all nine are
**carried here verbatim** with an auditable disposition. Wording is the attempt-4 review's own.

Dispositions are one of: **REMEDIATED-BY-REWRITE** (the defective text no longer exists in the
governing contract — verified by measurement, not assumed), **REMEDIATED-THIS-PASS** (the
defect survived the rewrite and was corrected now), or **CARRIED-OPEN**.

| Ref | Finding (attempt-4 wording, abridged to the defect) | Disposition | Evidence / landing site |
|---|---|---|---|
| **H-1** | **Claimed-but-unapplied.** The rev-7 header asserted the literal `## Plan Review — AUTHORITATIVE GATE STATE` heading was *"inserted before the operative marker"* — **it did not exist at review time**. Three live cross-references dangled, including one **inside a destructive-action approval record**. | **REMEDIATED-THIS-PASS** | The heading does **not** exist in this plan (measured: **0** occurrences) — the operative marker lives in **§13** and nowhere else. The dangling references survived in the **backlog records** (`021-S` cited the non-existent heading inside `PA-021-CASCADE`; `022-F` cited it as "OPERATIVE MARKER LOCATION"); both were repointed to **§13** in the same commit as this revision. |
| **H-2** | **Backlog split-brain, FOURTH consecutive recurrence (`A-14` → `B-7` → `E-9` → `H-2`).** Records carried superseded revision/attempt state and no `PA-*` cross-reference. The standing process fix mandates the resync travel in the **same commit** as the revision bump — and it did not. | **REMEDIATED-THIS-PASS** (recurrence **5**) | `022.001-T` still identified the plan as *"Plan rev 3: docs/plans/2026-09-12-…-foundation-plan.md"* — the **archived** path at a **five-revision-stale** revision. All four records (`022-F`, `022.001-T`, `021-S`, `017-S`) resynced to this plan / rev 8 / **attempt 5**, **in this same commit**. |
| **H-3** | **Rev 7's `F-8` correction is FACTUALLY FALSE — a NEW defect.** Rev 7 asserted `consumer_id: "ship"` is *"access identity, not a filter."* **Measured live: it IS a filter.** Passing it **hides all 14 stage records**, defeating the full-enumeration anomaly gate. **Correct instruction: OMIT `consumer_id` entirely.** | **REMEDIATED-BY-REWRITE**, hardened this pass | §9.1 item 1 already instructs **pass NO parameters** and names both `consumer_id` and CLI `--agent` as filters. **Hardened now**: the attempt-4 counts (15/1/14) had already drifted to **17/1/16**, so pinned counts were demoted to a **dated observation** and the invariant restated as *"strictly fewer than unfiltered, re-derived at run time"* — a pinned count is the same drifting-identity class as `F-8`. |
| **H-4** | **§4.1's own opening still contradicts the rev-7 allowlist.** The operative first line still read *"The replacement text MUST NOT enumerate the authorized verdicts"*, while the rev-7 block and `AC-25` **permit** an explicit `{CASCADE, SAFE_CLOSE}` allowlist. The *"future close-path work lands without touching this file again"* claim also contradicted the bound-to-installed-contract rule. | **REMEDIATED-BY-REWRITE** | Measured in this plan: `MUST NOT enumerate` = **0**; `without touching this file again` = **0**. §3.4 states the permission positively and bounds it to the **currently installed** skill; §4.1 is now pinned replacement wording only. |
| **H-5** | **§13 still instructs reviewers to validate SUPERSEDED rev-6 behavior.** Item 5 asked for rows 14/15's *rev-6 declared-status* phrases (since replaced); item 7 asked whether the *P-015/skill-disagreement HALT* is correct (explicitly **deleted** by rev 7). | **REMEDIATED-BY-REWRITE** | The reviewer-instruction section does not exist in this plan; §13 is **gate state only**. The deleted disagreement check is stated as deleted in §3.4 (*"performs **no** runtime `policy_id`, policy-version or source-disagreement check"*). |
| **H-6** | **The pointer-only decision was not swept into the execution path.** §9 step 14, §9.3 step 5 and §9.4.1 still asserted per-item `matched` / `pre-archived` allocations of the installed pre-mode — exactly what `E-2`/`E-3` forbade. The aggregate `PROCEED` outcome is correct and measured; the **per-item allocation** is not the plan's to assert. | **REMEDIATED-BY-REWRITE**, scoped this pass | §9 step 14 now requires only *"an authoritative `PROCEED`"*; §9.3/§9.4.1 no longer exist. The one surviving `PREMODE_*` use (§9.2 check 6) is a consistency check on **Stage's own probe transcript**; an explicit scoping note was added so it cannot be misread as a taxonomy claim about live records. |
| **H-7** | **Evidence identity is internally inconsistent — NEW in rev 7.** Hardening 9a and `AC-20a` named **`_sha256`** fields, but the normative derivation compared **Git blob object IDs**, and the "measured" values shown were **abbreviated blob IDs**, not SHA-256 digests. An executor cannot satisfy the stated SHA-256 contract using the prescribed algorithm. | **REMEDIATED-BY-REWRITE** | §8.1 is **ONE scheme**: file identity is **Git blob OIDs** throughout. The only SHA-256 remaining is `ENGINE_SHA256_EXPECTED`, which identifies an **external** object (the resolved engine binary) — no self-reference and no algorithm mismatch. Measured: no `_sha256` **file-identity** field exists. |
| **H-8** | **The mandatory "verified snapshot restore" is not executable as specified.** The only restore command was `git restore -- .backlogit/queue/ .backlogit/archive/`, which restores **tracked paths from Git HEAD**, not the skill's in-memory snapshot, and does not remove newly-created untracked archive files. Because a1 sets the feature `done` **before** the cascade, a HEAD restore can disagree with the pre-call snapshot. The `PA-*` records additionally required verifying a *"protected set"* the installed skill says **does not exist**. | **REMEDIATED-THIS-PASS** | §10.3 is now executable: a **named pre-cascade baseline commit** (§9 step **13a**, so a1's result is inside the baseline), `git restore --source {pre_cascade_sha}`, **plus** deletion of exactly `{post_paths} − {pre_paths}` for untracked creations, **plus** a two-halves verification. `protected set` = **0** occurrences. The **stale unexecutable command survived verbatim in `PA-021-CASCADE`'s backlog record** and was replaced there this pass. |
| **H-9** | **Inventory vs implementation sequence mismatch.** §4 made **C1–C8** mandatory, but §3 described the change as **C1–C6** and §7's H1 instructed applying only ADD-1, ADD-2 and **C1–C6**. Following §7 omits C7/C8, which rows 14–15 and `AC-22` require — the documented H1 state is unreachable from the documented action. | **REMEDIATED-BY-REWRITE** | §2, §4, §7, §9 step 2, `AC-3` and the 18-row table now all read **CS1–CS8** (8 sites, all mandatory). The `Cn` → `CSn` rename also removed the `C1…C8` collision with the Probe-25 criterion tokens that made the mismatch easy to miss. |

**Open after this pass: NONE.** `P0 = 0`, `P1 = 0` carried open. Every disposition above is
either measured in this file or landed in the backlog records in the same commit.

**Why the rewrite closed six of nine without editing them.** `H-4`, `H-5`, `H-6`, `H-7` and
`H-9` were all defects of **stale surviving prose in an append-only artifact** — the exact
failure mode the attempt-4 review named as the binding constraint (*"a plan that is ~3,800
lines with four stacked review histories cannot be swept consistently by any single pass"*).
Compaction to a single governing contract removed the contradictory text rather than patching
around it. **This is recorded, not assumed**: each row above cites a measurement.

**The three that did NOT close by rewrite are the lesson.** `H-1`, `H-2` and `H-8` all
survived because they lived in the **backlog records**, not the plan — the compaction swept
the plan and left the records pointing at a heading that no longer exists, a plan revision
five versions stale, and a restore command the plan had already replaced. **A split-brain that
spans two artifact classes is not closed by fixing one of them.**

### 13.3 Gate state

| Field | Value |
|---|---|
| **Plan `status:`** | `planned` |
| **True-lineage attempt** | **5** (see §13.1) |
| **Harvest-ready** | **NO** until this gate returns PASS / ADVISORY with P0=0 and P1=0 |
| **`021-S`** | `queued` — **not claimable** until the gate passes |
| **`017-S`** | `queued`, dependency `[021-S]` — ineligible until `021-S` ships |
| **`PA-*`** | recorded, **unexercised**; gated on this state |
| **Re-entry budget** | **EXHAUSTED.** A FAIL here halts to the operator; attempt 6 is not self-authorized |

**Correction pass `F1`–`F8` applied before this gate** (adjudicated against the installed
contracts, not self-directed). **This is a SEPARATE, EARLIER finding set from `H-1`…`H-9`** —
`F1`–`F8` came from the internal correction pass over the compacted plan; `H-1`…`H-9` are the
carried attempt-4 gate findings disposed in §13.2. The two sets do **not** overlap and neither
renumbers the other:

| Ref | Disposition | Landed in |
|---|---|---|
| **F1** (P0) | **CONFIRMED.** Step 13 dirtied `.backlogit/` with no commit before the §10.3 clean-tree precondition | §9 step **13a**, §10.3 step 1, AC-22 |
| **F2** (P1) | **CONFIRMED.** The red/green summary was false for the guard rows | §6.1 two-class table + measured counts |
| **F3** (P1) | **CONFIRMED.** Live `022-F` is `queued`; the plan asserted `active` | §9 step **1a**, AC-19 |
| **F4** (P1) | **CONFIRMED, narrowly.** Installed order is Step 0.5 intake → Step 2 harness (P-002/P-004) → Step 4.1 task claim. The **shipment** claim before harness is the established contract and was **NOT** reordered; only the **task** activation was moved after red-phase confirmation | §9 steps **1b–1d**, AC-20 |
| **F5** (P1) | **CONFIRMED.** `resolve_checkpoint` is a mutation and ran on `main` before branch creation (P-011) | §9.1 item 7, §9 step **11a**, AC-21 |
| **F6** (P2) | **CONFIRMED as a clarification only.** The cascade is never committed before postchecks, so a committed-cascade recovery is unreachable; scope not broadened | §10.3 state machine, AC-23 |
| **F7** (P3) | **CONFIRMED.** `record-consistent` occurs **0×** in `_ship.agent.md` — it is **added**, not preserved | §4 CS7/CS8, AC-8 |
| **F8** (P3) | **CONFIRMED.** Records identified the plan by a stale line count | Records now use **path + revision** |

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
| Manifest/topology verification | 4 |
| 18-row Go test | 29 |
| H0/H1 cycle + gates | 10 |
| **Total** | **97 min** (margin **23** under the 2-hour rule) |

## 15. Constitution

| Principle | Assessment |
|---|---|
| **I. Safety-First Go** | One table-driven contract test. No `unsafe`, no goroutines, no I/O beyond reading one file; errors via `t.Errorf` |
| **II. Test-First (NON-NEGOTIABLE)** | H0 red → H1 green; the test lands before any `_ship.agent.md` edit. **Deviation D1** |
| **III. Workspace Isolation** | Probe 25 ran in a disposable, gitignored, workspace-contained directory; `FIXTURE_ISOLATED_FROM_LIVE=True`, `LIVE_BACKLOG_UNMUTATED=True` measured |
| **IV. CLI Containment (NON-NEGOTIABLE)** | Every probe call `--cwd $ws`; engine resolved via the registered command, never an absolute path |
| **V. Structured Observability** | a1 emits `A1_NOT_APPLICABLE`, `A1_ALREADY_DONE`, `A1_MULTIPLE_FEATURE_MEMBERS`; Probe 25 emits one `KEY=VALUE` per criterion |
| **VI. Single Responsibility** | **Dependency discipline**: zero module dependencies added; stdlib-only assertions; `testify` absent from `go.mod`; reuses `repoRoot(t)`. The `_ship.agent.md` change **removes** a dependency direction — Ship stops depending on a local copy of the classification. **Deviation D2** |
| **VII. Destructive Approval (NON-NEGOTIABLE)** | The cascade is destructive and **live**. Gated by P-015, pre-mode `PROCEED`, the §3.5 no-fallback HALT, **and** the two §11 approved-action records with operator authorization |
| **VIII. Explicit Safety Modes** | **Deviation D3** |
| **IX. Git-Friendly Persistence** | Line-oriented Markdown/Go; the backlog mutation is committed on the closure branch, never on `main` |
| **X. Agent Context Efficiency** | Net **removal** of ~20 lines of re-derived classification prose from Ship's hot path |
| **XI. Merge History (NON-NEGOTIABLE)** | Rollback is `git revert`; no rewrite, no force-push, no amend |

**D1 — 18 scenarios vs "fewer than 4".** These are **rows in one table-driven function** over a
**single** file, sharing one fixture and ≤4 helpers; each is a substring or index assertion, not
an independent scenario. Collapsing them reduces falsifiability. *Assertion density, not scope.*

**D2 — Width-Isolation: one task spans Markdown and Go.** `_ship.agent.md` is **not
documentation** — it is the agent's **executable instruction surface**, read and acted on at
run time. The Go test is a table of substring assertions over that one file, derived
mechanically from §6.1. They are **the contract and its compiler-check**. Splitting them
produces a task that cannot be verified and a task that cannot compile green, breaking the
mandatory H0-red → H1-green ordering, which requires both halves in one atomic change.
Precedent: `tests/integration/build_script_test.go` asserts over build text the same way.

**D3 — Principle VIII: global-closure blast radius without a mode directive.** Accepted with
compensating controls: the change *removes* authority duplication **while also adding a narrow
permanent authority and new cascade reachability for `017-S`** — both stated plainly; four
negative widening guards (rows 7, 8, 9, 10 plus the `018.` guard) fail if anything new is
authorized; the grant is bounded by five conjunctive conditions, one position, one scope and
one transition; and A closes by the **pre-existing** path.
**Rejected simpler alternative: scope the grant to `021-S`/`017-S` by ID inside
`_ship.agent.md`.** Rejected because (a) it re-introduces release-instance IDs into the global
contract, the exact coupling §3.3 and §3.5 remove; (b) it relocates rather than reduces risk —
each new shipment would need another edit to the same global file; (c) it cannot be verified
negatively, since the widening guards assert the file contains **no** release-specific logic.
Its safety intent is preserved instead by §3.5 and the two §11 approved-action records.
**Is VIII satisfied via freeze-scope rather than deviated?** **No** — the grant is general and
permanent from merge; the approved-action records bound the **cascade invocations**, not the
grant. It remains a **documented deviation with compensating controls**.

## 16. Out of scope

The deferred B/C platform plans; `shipment-reconcile/SKILL.md` edits; Orchestrator repair; the
pre-existing `SAFE_CLOSE`/exit-9 tool conflict; and the P-015 detect-after-mutate **ordering**,
which is an inherited residual recorded as a follow-up against P-015 and the skill.

## 17. Full history

Superseded revisions 1–8, plan-review attempts **1–4**, and all withdrawn language are
preserved **verbatim and non-governing** in
`docs/archive/plans/2026-09-12-intercom-go-ship-feature-completion-foundation-plan.md`.
Git history is preserved; no content was deleted. **Where the archive and this plan disagree,
this plan governs.**

**Findings are the ONE exception to "non-governing".** The archive's *contract wording* is
non-governing, but **review findings raised against it remain live until disposed here**. The
attempt-4 findings `H-1`…`H-9` are carried into **§13.2** with per-finding dispositions and
evidence. **Do not read "the archive is non-governing" as "the archive's findings are
closed"** — that reading is what allowed nine unremediated P1s to become invisible when the
plan was compacted, and §13.2 exists to prevent it recurring.

---
title: "Decided Plan — Ship covering-feature completion and close-path delegation (022-F)"
date: 2026-09-13
status: planned
agent: Stage
revision: 8
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
| **CS7** | 255–256 | **Pointer-only delegation.** Replace the queue-only summary with a pointer to the installed `mode: pre` contract at the `expected_status` already named one line above; continue **only** on an authoritative `PROCEED`. **Preserve** the orphan scan, the `record-consistent` conjunct, and the adjacent `Scope note (139-F/139.001-T)` |
| **CS8** | 775–777 | Same pointer-only delegation at `expected_status: done`. **Preserve** the single-writer lock clause, the orphan scan and the `record-consistent` conjunct |
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

| Row | MUST BE ABSENT (exact, occurs once) | MUST BE PRESENT (absent today) |
|---|---|---|
| 14 | ``every manifest item is present in `.backlogit/queue/` with the`` (L255) | ``defers every per-item status decision to that skill's `mode: pre` classification`` |
| 15 | ``queue with `status: done`, and scans for orphan items.`` (L777) | ``defers every per-item status decision to the `mode: pre` classification at `expected_status: done``` |
| 9 | ``qualification is never per-member, and no feature ID is ever special-cased.`` (L819) | ``HALT on any result token the installed classification did not name`` |
| 10 | ``<shipment_id> --status shipped`` (L790) | ``the authoritative close sequence defined by the `shipment-reconcile` skill`` |
| 7 | ``it is fully covered (every one of its children,`` (L810) | — (ABSENT-only) |
| 16 | — | ``evaluated before `n` is computed`` **and** ``halts with no feature mutation`` |
| 8 | ``TASK_ONLY_FINALIZE`` | — (ABSENT-only) |
| 17 | — | ``scans for orphan items`` (survival; occurs twice — L256, L777) |
| 18 | — | ``Scope note (139-F/139.001-T)`` (survival) |

**Measured at the current tree**: all ABSENT literals occur **exactly once**; all PRESENT
literals occur **zero** times, so every `compound +/−` row is a genuine **red → green** and
none can be vacuously green at H0. Row 3's `a1` anchor is absent pre-implementation **by
design** — that absence is the H0 readiness guard.

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
| 1 | **S1** | Verify on `main`; pre-claim topology gate; **claim `021-S`**; `022.001-T -> active`; `022-F` left `active` |
| 2 | S1 | H0 red → implement → H1 green |
| 3 | S1 | Quality gates; review gate |
| 4 | S1 | Complete Task — commit the implementation, then **`022.001-T -> done`**, then **commit the resulting `.backlogit/` change** and verify a clean worktree. **Both commits must be in the PR that merges at step 5** — otherwise merged `main` still shows the task `queued` and a1 halts at step 13 |
| 5 | S1 | PR lifecycle — gates, build, push, PR, **P-018 Copilot engagement**, operator approval, **merge** |
| 6 | S1 | Merge Confirmation Gate — `gh pr view` `MERGED`; `git fetch origin main`; `git merge-base --is-ancestor {merge_sha} origin/main` |
| 7 | S1 | Reload merged `main`; detect the Role Boundary change; per §3.6 the new authority is **not** available to this session |
| 8 | S1 | **Write checkpoint** (§9.1) and **END**. Record the emitted filename |
| 9 | **S2** | **Fresh session.** Full checkpoint recovery (§9.1) |
| 10 | S2 | Verify merge SHA is an ancestor of `origin/main`; validate the same active shipment |
| 11 | S2 | Post-Merge Branch Protocol — `git checkout main`, pull, `git checkout -b post-merge/022-ship-feature-completion` **before** any backlog mutation |
| 12 | S2 | `--phase lifecycle` topology gate — **while `021-S` is still `active`**; never re-run after the close |
| **12a** | S2 | **READ-ONLY re-verification** of Stage's committed evidence (§9.2). FAIL ⇒ HALT, `021-S` stays `active`, no a1 mutation |
| 13 | S2 | **Step 6.1(a1)** — §5 **S4**. `022.001-T` declares `done` (archive-located = completed, not pre-archived) ⇒ all five conditions hold ⇒ `022-F active -> done` via `backlogit_move_item` |
| 14 | S2 | **Pre-mode**, `expected_status: done`. **Requires an authoritative `PROCEED`**; anything else HALTs |
| 15 | S2 | Classification **must return `CASCADE`** (§3.5) ⇒ close via the **pre-existing** cascade op; P-007 archive-integrity verify; post-mode; commit `.backlogit/` on the closure branch |
| 16 | S2 | Operational closure → `docs/closure/`; P-020 compact-context — on the closure branch, before the closure PR is pushed |
| 17 | S2 | Sync — **MCP `backlogit_sync_index` first**, CLI `backlogit sync` as declared fallback — then push, closure PR, **P-018 Copilot engagement**, local review, operator approval, merge |
| 18 | S2 | Return to `main`; pull |
| 19 | — | Orchestrator routes **`017-S`** |

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
   `consumer_id`, and do not use the CLI `--agent` flag: both are FILTERS.** Measured:
   `consumer_id=ship` → `total=1`; `--agent ship` → 1; `--agent stage` → 14; unfiltered → 15.
   A filtered enumeration hides records and silently disables the anomaly gate and the
   leftover warning below.
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
6. **Restore → Engram prune/gate → resume**, then **resolve only after a confirmed successful
   resume**, for that one checkpoint. Engram unreachable ⇒ **FAIL CLOSED**, no prune, no
   resume. A file-based prune degradation is not permitted.

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

1. **Capture is a precondition, not a reaction.** Before the cascade runs, S2 must establish
   **all three** of the following. **If any is unavailable, do not invoke the cascade.**
   a. **Clean trees** — `git status --porcelain -- .backlogit/queue/ .backlogit/archive/`
      returns **empty** (no modified and no untracked paths). A dirty baseline makes
      "what the cascade created" undecidable.
   b. **A real pre-cascade commit** — commit the closure branch so the pre-cascade state is
      restorable: `{pre_cascade_sha}`.
   c. **A complete path inventory** — record the full sorted path list of both trees,
      `{pre_paths}`, alongside the skill's in-memory Step 0(b) snapshot of declared `status`,
      `parent_id` and location for every watched record.
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
8. **AC-8** CS7/CS8 are pointer-only, preserve lock / orphan scan / `record-consistent` /
   `Scope note`, and state no per-item predicate. Rows 14/15/17/18 pass.
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

## 13. Gate state

```text
plan-review-attempt: 1
dispatch_mode: multi-agent
decision: PENDING
```

| Field | Value |
|---|---|
| **Plan `status:`** | `planned` |
| **Harvest-ready** | **NO** until this gate returns PASS / ADVISORY with P0=0 and P1=0 |
| **`021-S`** | `queued` — **not claimable** until the gate passes |
| **`017-S`** | `queued`, dependency `[021-S]` — ineligible until `021-S` ships |
| **`PA-*`** | recorded, **unexercised**; gated on this state |

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

Superseded revisions 1–7, four plan-review attempts, and all withdrawn language are preserved
**verbatim and non-governing** in
`docs/archive/plans/2026-09-12-intercom-go-ship-feature-completion-foundation-plan.md`.
Git history is preserved; no content was deleted. **Where the archive and this plan disagree,
this plan governs.**

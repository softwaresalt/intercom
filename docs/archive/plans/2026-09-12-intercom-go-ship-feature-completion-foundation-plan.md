---
title: "Plan — Ship covering-feature completion and close-path delegation (Foundation)"
date: 2026-09-12
status: superseded
agent: Stage
revision: 7
superseded_by: docs/plans/2026-09-13-intercom-go-ship-feature-completion-decided-plan.md
superseded_date: 2026-09-13
feature: 022-F
task: 022.001-T
shipment: 021-S
deliberation: docs/decisions/2026-09-13-intercom-go-a-only-simplification-decision.md
superseded_deliberation: docs/decisions/2026-09-12-intercom-go-three-shipment-bootstrap-deliberation.md
umbrella_evidence: docs/plans/2026-09-12-intercom-go-task-only-shipment-finalization-plan.md
---

# Plan — Ship covering-feature completion and close-path delegation (Foundation)

> ## ⚠️ SUPERSEDED — ARCHIVED HISTORY, NON-GOVERNING
>
> **This file is retained verbatim for audit and traceability ONLY. Nothing in it governs
> execution.** It accumulated seven revisions and four plan-review attempts in a single
> append-only artifact, which is itself the defect that made it non-convergent: later
> reviewers repeatedly treated superseded prose as current contract.
>
> **The SOLE GOVERNING ARTIFACT is:**
> **`docs/plans/2026-09-13-intercom-go-ship-feature-completion-decided-plan.md`**
>
> | | |
> |---|---|
> | **Superseded on** | 2026-09-13 |
> | **Last revision here** | rev 7 |
> | **Last gate recorded here** | attempt 4 — `decision: FAIL` (0 P0, 9 P1: `H-1`…`H-9`) |
> | **Successor revision** | rev 8 (decided plan) |
> | **Feature / task / shipment** | `022-F` / `022.001-T` / `021-S` (unchanged) |
> | **Releases** | `017-S` (unchanged) |
> | **Evidence** | Probe 25 — unchanged, still committed, still PASSing |
>
> **Where this file and the decided plan disagree, the DECIDED PLAN GOVERNS.** Do not apply
> any clause wording, acceptance criterion, recovery path or review finding from this file.
> The decided plan carries the final executable contract; every `E-*` and `H-*` finding
> recorded below was either resolved into that contract or is obsolete because the construct
> it described no longer exists.
>
> **Label mapping**: clause sites `C1`–`C8` here are **`CS1`–`CS8`** in the decided plan
> (renamed to end the collision with the Probe-25 criterion tokens `C1…C8`).


> **Requires plan hardening**: **yes — applied in rev 2, re-affirmed in rev 3, re-stated in
> rev 6** (see `## Plan Hardening`, including **Hardening 9a**).

> **Rev 3 — A is now the SOLE prerequisite.**
> `docs/decisions/2026-09-13-intercom-go-a-only-simplification-decision.md`
> (`SIMPLIFICATION_VALID`) retires the three-shipment chain for the current scope.
> Shipments **B** (`022-S`) and **C** (`023-S`) are **deferred generalized platform
> work** — records **archived non-destructively** (`archived_status: queued`), **not**
> current prerequisites. *(Rev 6, finding D-9: the rev-3 wording called them `blocked`,
> which is **not a valid shipment status** — the vocabulary is
> `queued|active|shipped|abandoned`. Their actual state is **archived** with
> `archived_status: queued`.)*
> **A's own scope, surface, inventory, budget and closure sequence are UNCHANGED.**
> What changed is only what comes *after* A: the route now goes **A → `017-S`**, and
> `017-S` is a **feature-present** `CASCADE` case, not the zero-feature example rev 2
> used.

> **Rev 4 — A-only final-review remediation cycle 1. Five substantive corrections.**
> Rev 3 passed internal review but the A-only final review found five blockers. All are
> fixed here, and **none of them widens A's surface beyond its two files**.
>
> 1. **Inventory reopened from 6 to 8 clause sites (§4).** Rev 3 deliberately scoped the
>    stale queue-only pre-mode summary at L776–777 *out* of the inventory to protect a
>    pinned budget. That was the wrong trade: the summary is **operative instruction text
>    Ship reads at the gate**, it is **literally false on `017-S`'s route**, and a *second*
>    instance of the same defect exists at **L255–256** (Step 0.5 item 6) which rev 3 never
>    detected at all. Both are now **C7** and **C8**. The budget is **recalculated, not
>    preserved**: ~78 → **~92 min**, margin ~42 → **~28 min**. Test rows 13 → **16**.
> 2. **Step 6.1(a1)'s selector is now exhaustive over member multiplicity (§5.1).** Rev 3's
>    three cases silently assumed *at most one* feature member. A **two-feature manifest
>    reached an unspecified branch.** Case **(iv)** now fails closed with **no** feature
>    mutation.
> 3. **S2's checkpoint recovery now cites the FULL installed protocol (§9.1, AC-9).** Rev 3
>    listed 8 steps and **omitted the `restore → prune/gate → resume` ordering, the Engram
>    reachability gate, and the never-prune allowlist** — all three of which are installed,
>    mandatory, and fail-closed in this workspace.
> 4. **The root-included 13-member fixture is now a NON-BYPASSABLE A closure gate (§9.4,
>    AC-20).** Rev 3 recorded it as a follow-up "before `017-S`, not before A". That left
>    the only unproven step in the chain guarded by nothing. It is now a **gate on A's own
>    close**.
> 5. **Formal `plan-review` gate executed (`## Plan Review`).** Rev 3 carried an internal
>    prose review only. The real gate ran as **attempt 1**, dispatching all seven selected
>    personas (including the `gpt-5.6-sol` anchor route) with machine-readable
>    `dispatch_mode:` / `decision:` markers. **It returned `decision: FAIL`.** Rev 4 is
>    therefore **NOT harvest-ready** and `021-S` must not be claimed; see `## Plan Review`
>    for the full finding set and the remediation queue.

> **Rev 5 — remediation cycle 2 of the attempt-1 `FAIL`. The Probe-25 hybrid is APPLIED.**
> Rev 4's own `## Plan Review` returned `FAIL` with 1 P0 and 12 P1s. Cycle 1 closed five
> (A-4, A-5, A-14, A-15, A-16). **Rev 5 closes the remaining eleven** — A-1, A-2, A-3,
> A-6…A-13 — plus the cheap same-surface P2/P3 queue. The single structural change:
>
> 1. **The §9.4 fixture is re-seated as a Stage pre-claim gate with a read-only Ship
>    re-verification (§9.4, §9.4a).** This is the **hybrid** the attempt-1 review
>    recommended and the operator selected. **Stage has already authored, executed and
>    committed Probe 25** — `PROBE25_RESULT=PASS`, all eight criteria, measured at `4fd21c5`
>    **before** `021-S` is claimed. Ship's fresh S2 session now only **re-verifies that
>    committed evidence read-only**. This closes **A-1** (no post-merge no-override
>    deadlock) and **A-2** (Ship never authors plan evidence, so no P-010 violation).
> 2. **Engine identity is portable (§9.4, A-3).** The contract resolves `backlogit` at
>    runtime via the registered command and compares its SHA-256 to the committed digest.
>    No runtime requirement hardcodes an absolute path; the observed path is recorded as
>    telemetry only.
> 3. **a1 is now a strictly-ordered, non-overlapping selector (§5.1, A-6)** with the
>    pre-archived descendant anomaly gate evaluated **before** `n` is computed (**A-10**).
> 4. **Test rows 14–16 are compound and non-vacuous (§8.1, A-7)** with all four old literals
>    and all four new phrases pinned verbatim. Row count **16 → 18**; budget **~92 → ~99 min**.
> 5. **A `## Constitution Check` section is added (A-13)**, and §9.1/AC-9 gain top-level
>    `session_id` (**A-8**) and an S2-specific fail-closed zero-candidate rule (**A-9**).
>
> **Rev 5 is submitted to a FRESH `plan-review` as attempt 2 — and it returned
> `decision: FAIL`.** The gate closed every attempt-1 P0/P1, but found **1 new P0 and 7 new
> P1s** in rev 5's own constructs. The decisive one: **§5.1's new S0 anomaly gate HALTs on
> A's own manifest**, because `022.001-T` is `done` (and therefore archive-located) by the
> time a1 runs — two personas converged on it and the committed Probe-25 transcript proves
> it. **Rev 5 is NOT harvest-ready and `021-S` must NOT be claimed.** See
> `## Plan Review — attempt 2` for the full finding set and the cycle-3 remediation queue.

> **Rev 6 — remediation cycle 3 (FINAL) of the attempt-2 `FAIL`. DECLARED STATUS replaces
> FILE LOCATION everywhere.** Rev 5's `## Plan Review — attempt 2` returned `FAIL` with 1 P0
> and 7 P1s in rev 5's own constructs. **Rev 6 closes B-1 through B-8 and the
> `C-*`/`D-*` same-surface items enumerated below — but NOT all of them.** *(Rev 7, finding
> **E-10**: the rev-6 header originally claimed it closed *"the same-surface P2/P3 queue"*
> outright. That was an **overclaim**. Closed in rev 6: **C-1…C-10** and **D-1, D-2, D-5,
> D-6, D-8, D-9**. **NOT** addressed in rev 6 and carried forward explicitly: **D-3**
> (Governance-required rejected simpler alternative / Principle VIII freeze-scope framing),
> **D-4** (Principle VI assessed as cohesion rather than dependency discipline;
> Width-Isolation undocumented), **D-7** (S4 transition carries no tool pin), and **D-10**
> (B-12 `--complexity` prohibition claimed closed but addressed nowhere). **All four are
> CLOSED IN REV 7** — see the rev-7 block below.)* The one root-cause correction, and six
> localized ones:
>
> 1. **B-1/B-2 — every lifecycle decision is re-based on DECLARED `status`; file location
>    proves exactly-one-copy containment ONLY (§4.4, §5, §5.0, §5.1 S0, rows 14/15, AC-22,
>    AC-24, AC-27).** The installed skill is explicit (L394–399): declared status is *"never
>    inferred from, nor substituted by, which of `queue/`/`archive/` currently holds the
>    record."* **A descendant declaring `status: done` is a COMPLETED descendant even when
>    registry routing stores it under `archive/`; a PRE-ARCHIVED descendant means declared
>    `status: archived`**, which alone triggers the §5.0 11-ID allowlist and the
>    provenance/immutability checks. **`022.001-T`, `done`-in-archive at the a1 gate,
>    therefore PASSES §5 condition 1 and is NOT an unmapped pre-archived member** — the P0
>    stall is gone. The committed transcript (`MEMBERS_QUEUE_LOCATED=0` with
>    `PREMODE_MATCHED=2`/`PREMODE_PRE_ARCHIVED=11`) was always consistent with the declared-
>    status rule; rev 5's restatement of the rule was the defect. **No re-measurement.**
> 2. **B-3 — the enumeration contradiction is resolved by the SAFE ALLOWLIST CONTRACT (§4.1,
>    AC-5, AC-25, row 9).** `_ship.agent.md` MAY name the **result tokens the installed skill
>    can currently emit**, purely to validate output; it MUST NOT restate any classification
>    **predicate**. Unknown/absent/ambiguous token ⇒ **HALT**; skill unavailable or unreadable
>    ⇒ **HALT**; bound to the **currently installed** contract, never a future capability.
> 3. **B-4 — Hardening 9's withdrawn no-override language is DELETED, and four distinct
>    fields replace the overloaded `evidence_sha` (Hardening 9a, §9.4a):**
>    `evidence_commit_sha` (the **LAST** commit touching either file), `fixture_script_sha256`,
>    `fixture_result_sha256`, `backlogit_executable_sha256`. Stage verifies them pre-claim;
>    **S2 re-verifies them read-only**. A post-merge mismatch **HALTs active A** with named
>    operator recovery — **no rollback or override fiction**.
> 4. **B-5 — the pre-mode circularity is removed (§5.1).** **a1 runs BEFORE pre-mode and MUST
>    NOT require it**; its read-only guard directly checks declared status, exactly-one
>    physical copy, topology/parent, and the exact 017 allowlist. This is a **narrow
>    completion precondition**, not duplicate close-path classification.
> 5. **B-6 — the CASCADE reachability and fallback overclaims are deleted (§9.3, §9.4b).**
>    The **authoritative** mode selection runs **after** a1; A/017's exact root-inclusive
>    manifests **EXPECT `CASCADE`**, and **any other verdict is a topology discrepancy that
>    HALTs — there is NO `SAFE_CLOSE` fallback**. The pre-existing cascade branch is invoked
>    **only** on an authoritative `CASCADE`. The inherited detect-after-mutate residual is
>    stated honestly, including that **A adds new reachability** to it.
> 6. **B-8 — S2 checkpoint selection order is corrected (§9.1 step 2a, AC-29).** Enumerate
>    all with **no filter** → anomaly gate over the **full** enumeration → **then** filter to
>    filename + `agent: ship` + `session_id` → require **exactly one MATCH**. **Unrelated
>    active checkpoints no longer strand the handoff.**
> 7. **§12's arithmetic is fixed (C-1).** Rev 5's rows summed to **97** while claiming ~99.
>    Rows now sum **exactly** to **97 min**, margin **23 min**. Rev 6 adds **no** new work.
>
> **Rev 6 was submitted to a FRESH `plan-review` as attempt 3 — the LAST cycle before the
> Escalation Protocol applied.** It returned `decision: FAIL` (0 P0, 10 P1 — `E-1`…`E-10`),
> the re-entry budget was exhausted, and Stage's circuit **opened**.

> **Rev 7 — OPERATOR-AUTHORIZED EXCEPTIONAL REMEDIATION CYCLE (adjudicated, not
> self-directed).** Rev 7 exists **only** because the operator explicitly authorized one
> bounded additional Stage remediation cycle after the circuit-breaker escalation and an
> independent 5-model adversarial decision of **`READY for one operator-authorized
> exceptional remediation cycle`** (authorization timestamp **2026-09-13T11:35:43-07:00**).
> **Stage did not re-enter remediation on its own authority**, and the attempt-3 escalation
> record stands unaltered. The adversarial decisions below were **handed down, not
> designed by Stage** — rev 7 applies them and does not redesign them.
>
> 1. **`E-1` — declared status determines LIFECYCLE CLASS; placement is a separate
>    condition (§5.0 condition 2, AC-27).** Rev 6 re-based the *trigger* on declared status
>    but left §5.0 condition 2 requiring `archive/` **specifically**. Rev 7 separates the two
>    concerns cleanly: the **general** torn/duplicate check is **exactly one physical copy
>    (`queue/` XOR `archive/`)** and applies to every member regardless of status; **once
>    condition 1 has established a declared `status: archived`**, archive-only placement is a
>    **separate routing/integrity condition**, not a lifecycle inference. **Lifecycle class is
>    never inferred from location.**
> 2. **`E-2`/`E-3` — C7/C8 become POINTER-ONLY DELEGATION (§4, §4.4, rows 14/15, AC-22).**
>    Rev 6 replaced a false queue-only *predicate* with a **different restated predicate**,
>    and additionally promoted the skill's **cascade Step 0(b)** snapshot rule (L394–399)
>    into the **pre-mode** API, which the installed pre-mode does not have. Rev 7 stops
>    restating the classification entirely: both sentences become a **pointer to the
>    installed `shipment-reconcile` `mode: pre` contract at the required `expected_status`**,
>    proceeding **only** on an authoritative `PROCEED` and otherwise halting. The lock,
>    orphan scan, record-consistency conjunct and the adjacent `Scope note` are **preserved**.
>    **Rev 7 does NOT claim the installed pre-mode uses this plan's declared-status
>    taxonomy.**
> 3. **`E-4`/`E-5`/`E-6` — the delegation contract is made SINGLE-AUTHORITY (§4.1, AC-25,
>    AC-28, Hardening 2).** **a1 runs first**; Ship **must then invoke the installed
>    `shipment-reconcile`**. Output validation **may allowlist the installed close verdict
>    tokens `CASCADE` and `SAFE_CLOSE`** without restating any predicate. `HALT` and
>    `RECONCILE_FAIL` are **non-close procedural outcomes**, not close verdicts; **`BLOCKED`
>    is not installed and is an unknown token**. The **runtime `policy_id`/version comparison
>    and the independent P-015 re-computation/disagreement check are REMOVED** — the
>    installed skill is the **runtime** authority and P-015 is the **statically reviewed**
>    authority. Expected-`CASCADE` for `021-S`/`017-S` lives in **this plan's execution
>    contract**, never as global ID-specific logic in `_ship.agent.md`.
> 4. **`E-7` — the mandatory verified snapshot restore covers EVERY post-CASCADE check
>    (§9.4b, §10.3).** Not only out-of-`allowed_ids` archival: **non-empty `returned_ids`**,
>    **either two-set difference**, and an **altered or cleared `parent_id` at cascade
>    step 4** each halt **after** mutation and each now carry the same mandatory verified
>    restore, mirrored in both tables.
> 5. **`E-8` — transient drift and committed-evidence failure are separated (§9.4a,
>    AC-20a, §10.3).** **Transient toolchain/executable drift** may be corrected and
>    **re-verified read-only**, in place. A **committed evidence / hash / ancestry failure
>    CANNOT be cleared in place** — it requires `git revert` per §10.2 and a re-plan. The
>    blanket "only revert clears" contradiction is removed.
> 6. **`E-10` — the header overclaim is corrected** (see the rev-6 block above), with the
>    four genuinely-deferred same-surface items **named by ID**. **D-3, D-4, D-7 and D-10
>    are closed in rev 7.**
> 7. **Evidence identity is derived NON-SELF-REFERENTIALLY (Hardening 9a, §9.4a check 2).**
>    `evidence_commit_sha` comes from `git log -1` over **both** files; the committed
>    **blobs read from that commit** are compared to the current files; the **expected engine
>    digest is read from the committed transcript** and compared to the currently-resolved
>    command. **`fixture_result_sha256` is NOT embedded inside the file it hashes.** The
>    **exact nine STEP 11 aggregate criterion names are pinned** (closing `F-7`).
> 8. **Checkpoint wording is corrected to the installed tool surface (§9.1).**
>    `consumer_id: "ship"` is **access identity, not a filter**; the anomaly gate reads the
>    **actual top-level `needs_quarantine` / `quarantined` / `total` counters**. Exact
>    filename + `agent` + `session_id` matching and unrelated-candidate tolerance are
>    **preserved unchanged**.
> 9. **Low findings closed.** Hardening 1's overclaim corrected; `backlogit_move_item` /
>     `backlogit move {id} --status done` **pinned** in §5.1 **S4** and §9 step 13
>     (**D-7**/`F-10`); Probe 25 criteria **2 and 3** disclosed as **PowerShell-derived**,
>     with the actual engine cascade invocation beginning at `CASCADE` (`F-4`).
>     The literal `## Plan Review — AUTHORITATIVE GATE STATE` heading **now exists at the end
>     of this file**, created as part of recording the attempt-4 gate.
>     **HONESTY NOTE (finding `H-1`):** an earlier draft of this header asserted that heading
>     had already been inserted **while it did not yet exist**, and four personas flagged the
>     dangling cross-references — including one inside a `PA-*` approval record. That is the
>     same overclaim class as `A-16` and `E-10`. The claim is true **only as of this commit**,
>     and `H-1` is recorded as a **blocking P1**, not silently cleared.
> 10. **Strict-safety approved-action records added for BOTH destructive operations**
>     (`## Strict Safety — Approved Destructive Actions`): **`PA-021-CASCADE`** and
>     **`PA-017-CASCADE`**. Principle **VI** is completed with dependency discipline, the
>     **Width-Isolation** explanation, and **D4's rejected simpler alternative** (**D-4**,
>     **D-3**, `F-11`, `G-3`, `G-2`).
>
> **Rev 7 was submitted to a FRESH `plan-review` as ATTEMPT 4** — the one operator-authorized
> exceptional gate run. **It returned `decision: FAIL`** (0 P0, **9 P1** — `H-1`…`H-9`),
> 7/7 personas, anchor `gpt-5.6-sol`/high. See `## Plan Review — attempt 4` and the
> `## Plan Review — AUTHORITATIVE GATE STATE` section at the **end of this file** for the
> operative markers. **The plan is NOT harvest-ready; `021-S` is NOT claimable; neither
> approved destructive action is exercisable.** The exceptional re-entry budget is **spent**.

- **Shipment**: `021-S` (A, Foundation) — **the sole prerequisite**; the only eligible shipment
- **Closes by**: the **existing** P-015 `VERIFIED FULLY-COVERED-ROOT EXCEPTION` (`CASCADE`)
- **Authorizes**: nothing new in the close-path verdict set — **a result-token allowlist is
  output validation, not authorization (§4.1 rev 6)**
- **Releases**: `017-S` (13-member fully-covered root), which then closes by the **same existing** `CASCADE`

## 1. Objective

Make the fully-covered-root manifest shape **reachable** by Ship, and stop Ship from
restating a classification it should delegate.

Three deliveries, one surface plus one test:

1. **Narrow Role Boundary authority** for Ship to complete a covering feature
   `active -> done`, fail-closed.
2. **An operative post-task / pre-reconcile parent-completion step** that exercises it.
3. **Replace the duplicated close-path exception wording** in `_ship.agent.md` with a
   generic delegation to the authoritative P-015 + `shipment-reconcile` machine-selected
   verdict — **without authorizing any new verdict**.

Current behavior after this lands remains exactly `CASCADE` or `SAFE_CLOSE`/`HALT`.

## 2. Why this is the sole prerequisite

Two independent reasons, both verified at `e7e981d`:

**2.1 Pre-mode makes the shape unreachable today.** `_ship.agent.md` Step 6.1(a) invokes
`shipment-reconcile` with `expected_status: done` and verifies *"every manifest item is
present in queue with `status: done`"*. A manifest `[022-F, 022.001-T]` therefore needs
the **feature** `done`. Ship's Role Boundary permits only *"move **tasks** to
active/done"* (line 38). No feature-completion authority exists, so no fully-covered-root
manifest can ever reach closure.

**`017-S` has exactly this shape**, now that its root `018-F` has been restored to the
manifest (13 members). So `017-S` cannot close until A is live — which is precisely why A
is the sole prerequisite, and why nothing else is needed between them.

**2.2 The duplicate has already drifted unsafe.** This is a pre-existing defect found
during re-planning, independent of the bootstrap:

| Surface | Coverage rule |
|---|---|
| P-015 item 1 (v1.24.0), `workflow-policies.md:438` | *"every one of its **descendants — at every depth, not only direct children**"*, and: *"A check limited to direct children is insufficient"* |
| `_ship.agent.md:810` | *"every one of its **children**, enumerated live from `.backlogit/queue/` + `.backlogit/archive/`"* |

A manifest `[feature, task]` whose task owns an out-of-manifest subtask **passes Ship's
copy and fails the policy** — precisely the corruption v1.24.0 was written to prevent.
Delegation removes the drift **and** the class of defect.

## 3. Surface — exactly 2 files

| # | File | Change |
|---|---|---|
| 1 | `.github/agents/_ship.agent.md` | Role Boundary row; new Step 6.1(a1); C1–C6 |
| 2 | `tests/integration/ship_feature_completion_contract_test.go` | one table-driven test fn + ≤4 helpers |

**New production code files: 0.** No policy file, no skill file, no source change.

## 4. Exact inventory — 8 clause sites (all MUST) + 2 additive

Line anchors **re-verified against the installed file at `d19b954`**. C1–C6 match rev 6
Inventory Table §6.1-C exactly. **C7 and C8 are new in rev 4** and are *not* in rev 6's
table — C8 was recorded there but deliberately excluded, and C7 was never detected.

> **Label order is allocation order, not line order.** Sorted by line the sites run
> **C7 (255) → C1 (756) → C8 (776) → C2 (789) → C3 (794) → C4 (802) → C5 (821)**. C1–C6
> keep their rev-6 labels so cross-revision traceability survives; the two new sites take
> the next free labels. An implementer MUST apply by **anchor text**, never by label order
> and never by line ordinal.

| # | Line | Clause | Change |
|---|---|---|---|
| **C1** | 756–764 | Mandatory pre-self-close context reload — names only Ship instructions + `shipment-reconcile` | Also reload **P-015**; positively verify merged tokens **and the merge commit**; add the §6 authority-escalation carve-out with the **pinned reconciled wording** of §4.2 |
| **C2** | 789–791 | safe-close summary prescribing `backlogit move <shipment_id> --status shipped` | Correct: the engine **refuses** this for shipments (exit 9, Probe 18). Replace with a pointer to the skill's authoritative step 8 |
| **C3** | 794–795 | *"**Do NOT call** `backlogit shipment ship` … **unless** the P-015 VERIFIED FULLY-COVERED-ROOT EXCEPTION applies"* | Replace the named single exception with: do not call it unless **the machine-selected verdict** authorizes it |
| **C4** | 802–823 | the re-derived classification prose (binary selector + drifted `children` rule) | **Delete the restatement.** Replace with generic delegation |
| **C5** | 821–823 | *"invoke the cascade … **in place of** the safe-close sequence above"* | Subsumed by C4's delegation block; single-exception framing removed |
| **C6** | new bullet | — | **NEW** delegation bullet: Ship invokes whichever close path the verdict names, and **no** path absent from the verdict |
| **C7** | 255–256 | **(rev 4)** Step 0.5 item 6 intake pre-mode summary: *"verifies every manifest item is present in `.backlogit/queue/` with the expected status"* — **queue-only** | **(rev 7, finding E-2)** **POINTER-ONLY DELEGATION (§4.4).** Replace the restated per-item predicate with a pointer to the **installed `shipment-reconcile` `mode: pre` contract at the `expected_status` already named one line above**, proceeding **only** on an authoritative `PROCEED` and otherwise halting. **Preserve** the orphan scan, the `record-consistent` conjunct, and the adjacent `Scope note`. **State no classification predicate of any kind** |
| **C8** | 776–777 | **(rev 4)** Step 6.1(a) pre-archive gate summary: *"verifies that every manifest item is present in queue with `status: done`"* — **queue-only** | **(rev 7, finding E-2)** Same **pointer-only delegation** as C7, at `expected_status: done`. **Preserve** the single-writer lock clause, the orphan scan and the `record-consistent` conjunct |

| Additive | Content |
|---|---|
| **ADD-1** | Role Boundary table row (line 38) — narrow feature-completion authority |
| **ADD-2** | New **Step 6.1(a1)** — covering-feature completion gate |

### 4.0 Why C7 and C8 were reopened (rev 4 — reverses a rev-3 decision)

Rev 3 recorded L776–777 as *"deliberately OUT OF A's INVENTORY"* on two stated grounds —
budget, and *"not load-bearing … descriptive drift in a summary, not an operative gate"* —
and instructed that *"a reviewer MUST NOT fail A for its absence."* **Rev 4 withdraws that
instruction and both grounds.**

1. **It is operative, not descriptive.** `_ship.agent.md` is the agent's **instruction
   surface**. A sentence telling Ship what a gate verifies *is* what Ship acts on when it
   reads the step. The rev-3 argument — that the skill computes the right answer anyway —
   proves the *runtime* is safe; it does not make a false instruction harmless. An agent
   reading *"every manifest item is present in queue with `status: done`"* against
   `017-S`, where **11 of 13 members are not in queue at all**, has every reason to
   conclude the manifest is broken and halt, or to "help" by restoring archived members to
   the queue. That is a **fabricated** failure on the exact route A exists to enable.
2. **The same defect exists twice, and rev 3 found only one instance.** C7 (L255–256) is
   the intake-side twin. Rev 3's inventory was therefore not merely under-budgeted, it was
   **incomplete** — which is precisely the §2.2 duplication-drift failure mode this plan
   was written to eliminate, reproduced inside the plan's own inventory.
3. **Budget is an output, not a constraint.** Pinning an inventory to protect a number
   inverts the relationship. The correct response is to **recalculate** (§12): ~92 min,
   margin ~28 min, still inside the 2-hour rule with room. Had it not fit, the correct
   response would have been HALT-and-return-to-Stage (§10.3), never silent exclusion.

### 4.4 C7/C8 reconciled pre-mode wording (normative)

> **Rev 7 — C7/C8 ARE POINTER-ONLY DELEGATION. THE PLAN NO LONGER RESTATES THE PRE-MODE
> CLASSIFICATION (plan-review findings E-2/E-3, NORMATIVE, supersedes rev 6's table).**
>
> Rev 6 correctly identified that the rev-4 wording was **locational and false**, but then
> fixed it by **writing a different predicate** — a declared-status restatement — into
> Ship's instruction surface. That reproduced the very duplication §2.2 exists to eliminate,
> and it did so **inaccurately**: rev 6 sourced its taxonomy from the installed skill's
> **L394–399**, which is the **Cascade Close Sub-Procedure's Step 0(b) snapshot rule**, and
> presented it as the **pre-mode API**. The installed pre-mode's own `## Classification`
> table and its step-3 operational sketch are **location-first**. **Rev 7 therefore makes no
> claim about which taxonomy the installed pre-mode uses internally**, because Ship does not
> need to know and this plan must not assert it.
>
> **The reconciliation is a POINTER, not a paraphrase.** C7 and C8 delegate to the installed
> contract and state only what Ship must *do* with the result:
>
> | Element | C7 (L255–256) | C8 (L775–777) |
> |---|---|---|
> | **Delegate to** | the installed `shipment-reconcile` `mode: pre` contract | the same, at `expected_status: done` |
> | **Required expected status** | the `expected_status` already named one line above (`queued`, or `active` if already claimed) | `expected_status: done` |
> | **Proceed condition** | **only** an authoritative `PROCEED` | **only** an authoritative `PROCEED` |
> | **Otherwise** | **HALT** — surface the report; no mutation | **HALT** — surface the report; do not proceed to step 1.b |
> | **Preserved verbatim** | the **orphan scan**, the **`record-consistent`** conjunct, and the adjacent **`Scope note (139-F/139.001-T)`** | the **single-writer lock** clause, the **orphan scan**, and the **`record-consistent`** conjunct |
> | **Forbidden** | any per-item classification predicate — locational **or** declared-status | same |
>
> **Why this is strictly safer than rev 6's wording.** A pointer cannot drift out of sync
> with the skill, because it asserts nothing the skill could contradict. Rev 6's
> declared-status sentence could have become false the moment the installed pre-mode changed
> — and, per E-3, **was already describing semantics the delegated pre-mode does not
> document**. The pointer is correct under **every** internal taxonomy the skill may use.

**The classification itself is the installed skill's, and this plan does not restate it.**
For the two manifests A actually closes, the only property this plan depends on is the
**outcome**: `PROCEED`. That outcome is **measured**, not assumed — the committed Probe-25
transcript records `PREMODE_RESULT=PROCEED` for the exact 13-member root-included shape
(`PREMODE_MATCHED=2`, `PREMODE_PRE_ARCHIVED=11`, `PREMODE_MISSING=0`,
`PREMODE_STATUS_MISMATCH=0`, `PREMODE_DUPLICATE_TORN=0`).

> **Safety note — the outcome is stable under BOTH readings of the installed skill.** The
> skill's normative L394–399 (declared-status) and its location-first step-3 sketch are in
> tension with each other, and that tension is **disclosed, not papered over**. It is
> **not load-bearing for either live route**: `PROCEED` accepts `matched` **and**
> `pre-archived` alike, so **both** readings return `PROCEED` for `021-S` and for `017-S`.
> Aligning the skill's sketch with its own normative rule is a **recorded follow-up against
> the skill**; `shipment-reconcile/SKILL.md` is **outside A's 2-file surface and is NOT
> edited here**.

`PROCEED` carries two further conjuncts beyond per-item acceptance, and the replacement
sentences MUST preserve both. Asserted by Go test rows **14**, **15** and **17**.

> **Rev 6 — DECLARED STATUS IS THE SOLE LIFECYCLE AUTHORITY *FOR THIS PLAN'S OWN GATES*
> (plan-review findings B-1/B-2, NORMATIVE, load-bearing; SCOPE CORRECTED in rev 7, E-3).**
> Rev 5's table defined `matched` as *"**Queue file present** AND declared status matches"* —
> a **location-derived** predicate. That is the single root cause of both B-1 and B-2, and it
> is **falsified by this plan's own committed measurement**.
>
> > **Rev 7 scope correction (E-3).** The rule below governs **this plan's own gates** —
> > §5, §5.0, §5.1's a1 selector, and §9's step descriptions — **all of which read records
> > directly and are this plan's to define**. It is **NOT** a description of the installed
> > **pre-mode API**, and rev 7 does **not** assert that pre-mode classifies this way
> > internally. The quotation below is the skill's **Cascade Close Sub-Procedure Step 0(b)
> > snapshot rule**; it is cited here as the **precedent and authority for a1's own reading
> > discipline**, which is exactly the job a1 does. C7/C8 carry **none** of this taxonomy —
> > they are pointer-only (see the rev-7 block at the top of §4.4).
>
> **The rule, quoted verbatim from the installed `shipment-reconcile/SKILL.md` (L394–399,
> Cascade Close Sub-Procedure Step 0(b)):**
>
> > **Declared `status` is read from the record's own frontmatter `status` field — never
> > inferred from, nor substituted by, which of `queue/`/`archive/` currently holds the
> > record.** A record residing in `.backlogit/archive/` while declaring `status: done` is
> > **not** truly archived; only a declared `status: archived` counts as truly archived …
> > location alone is never sufficient.
>
> **Two consequences, both binding on every gate THIS PLAN defines:**
>
> 1. **A descendant declaring `status: done` is a COMPLETED descendant — full stop** — even
>    when registry-managed storage has relocated it under `.backlogit/archive/`. It satisfies
>    §5 condition 1, it is **not** live for condition 2, and it is **NOT** a pre-archived
>    member. The live `registry.yaml` routes `done|accepted|rejected|archived` to `archive/`
>    and only `queued|active|blocked|review` to `queue/`, so **correct completion itself moves
>    a record out of the queue**. Treating that relocation as evidence of anything is the
>    defect.
> 2. **A PRE-ARCHIVED descendant means declared `status: archived` — never merely
>    archive-located.** Only such a record is subject to §5.0's exact 11-ID allowlist and its
>    immutable-metadata/provenance/parent checks.
>
> **Physical-location reads are retained, with exactly one narrowed job:** proving
> **exactly-one-copy containment** (present in `queue/` XOR `archive/`; both ⇒ torn duplicate,
> neither ⇒ missing). They **never** classify lifecycle state. This preserves the
> duplicate/containment guarantee in full while removing every status inference from it.
>
> **The measurement that forces this (committed, `PROBE25_RESULT=PASS`).** In one transcript
> block the fixture records `MEMBERS_QUEUE_LOCATED=0` **and** `MEMBERS_ARCHIVE_LOCATED=13`
> **and** `PREMODE_MATCHED=2 PREMODE_PRE_ARCHIVED=11 PREMODE_RESULT=PROCEED`. Under rev 5's
> locational table those are contradictory (zero queue-located ⇒ `MATCHED=0`). Under the
> declared-status rule they are **exactly consistent**: the 2 members declaring `status: done`
> are `matched`, the 11 declaring `status: archived` are `pre-archived`. **The transcript was
> never wrong; rev 5's restatement of the rule was.** No re-measurement is required, and none
> was performed.
>
> **Disclosed tension inside the skill, recorded not papered over (SCOPE CORRECTED rev 7,
> E-3).** The skill's *operational sketch* at its pre-mode step 3 is written location-first
> (*"attempt to locate the file at `.backlogit/queue/{id}.*` … if NOT found in queue, check
> `.backlogit/archive/{id}.*` — … classify as `pre-archived`"*), and its `## Classification`
> table reads the same way. That sketch predates registry-managed `done` relocation and is in
> tension with the skill's own L394–399 rule **in the cascade sub-procedure**.
> **Rev 6 resolved that tension by declaring L394–399 governing over the pre-mode API. Rev 7
> WITHDRAWS that resolution as out-of-scope overreach**: which rule governs *inside* the
> installed skill is **the skill's business, not this plan's**, and A does not edit
> `shipment-reconcile/SKILL.md` (outside the 2-file surface). Aligning the sketch is a
> **recorded follow-up against the skill**.
> **Why A does not need the tension resolved:** `PROCEED` accepts `matched` **and**
> `pre-archived` alike, so **both** readings return `PROCEED` for `021-S` and for `017-S` —
> as the committed transcript measures. The B-1 stall was never in pre-mode; it was in
> **this plan's own** a1/S0 gate, which is what §5.0/§5.1 correct, and which C7/C8's
> pointer-only wording now keeps entirely out of Ship's instruction surface.

> **Rev 5 — the two PROCEED conditions rev 4 dropped (plan-review finding B-11, normative).**
> `PROCEED` is **not** "every item is `matched` or `pre-archived`" alone. The authoritative
> rule carries two further conjuncts that the rev-4 summary silently lost, and a reconciled
> sentence that omits them would be a *new* inaccuracy replacing the old one:
>
> * **no orphans** — the orphan scan must find none; and
> * **`record-consistent`** — the shipment-record-status classification must be
>   `record-consistent`; any other value is `HALT — operator reconcile required`.
>
> **C8 additionally carries the lock acquisition and the orphan scan** in the same sentence
> region. The reconciliation MUST preserve those clauses intact. An implementer who rewrites
> the sentence wholesale and drops the single-writer lock or the orphan scan has introduced a
> **worse** defect than the queue-only wording being fixed. Asserted by the ABSENT/PRESENT
> literals pinned in §8.1 rows 14–15 and by new row **17**.

> **Rev 5 — C7 must COEXIST with the adjacent scope note (plan-review finding B-14).**
> Immediately after C7, at L257+, `_ship.agent.md` carries a
> `Scope note (139-F/139.001-T)` which scopes Step 0.5 item 6 to **uniform-status** intake.
> `017-S`'s intake manifest is **not** uniform-status — it is 2 live + 11 pre-archived. C7's
> replacement MUST be applied **without deleting or contradicting** that scope note: the note
> narrows *when item 6's check applies*, while C7 corrects *what that check accepts when it
> does apply*. These are orthogonal. Asserted by new row **18** (the scope note survives).

> **Scope guard.** C7 and C8 correct a **summary of an existing gate**. They change no
> threshold, add no classification, and grant no authority — the skill's semantics are
> already what they are. This is A's `_ship.agent.md` file, so both sites are inside A's
> declared 2-file surface; `shipment-reconcile/SKILL.md` is **not** touched.

> **Rev 2 — additive-item label rename, recorded for traceability.** Rev 1 labelled these
> two additive items `A-1` and `A-2`. That collided visually with plan **B**'s clause-site
> labels `A1`–`A8` (B's sites are named `A*` because rev 6's Inventory Table §6.1-A covers
> `workflow-policies.md`). Because the three plans are reviewed together, `A-1` and `A1`
> appearing in the same review were a genuine mis-read hazard. They are renamed
> **`ADD-1`/`ADD-2`** here. **Traceability:** rev 1 `A-1` ≡ rev 2 `ADD-1`; rev 1 `A-2` ≡
> rev 2 `ADD-2`. No content changed — label only. The cross-plan label map is:
>
> | Plan | Surface | Clause-site labels | Origin |
> |---|---|---|---|
> | **A** (this plan) | `_ship.agent.md` | `C1`–`C6` + `ADD-1`/`ADD-2` | rev 6 Inventory Table §6.1-C |
> | **B** | `workflow-policies.md` | `A1`–`A8` | rev 6 Inventory Table §6.1-A |
> | **C** | `shipment-reconcile/SKILL.md` | `B1`–`B18` | rev 6 Inventory Table §6.1-B |
>
> The letters track the **rev-6 inventory section**, not the shipment. That is confusing
> but it is the traceable naming, so it is documented rather than silently re-lettered.

### 4.1 Delegation wording constraint (normative)

The replacement text MUST NOT enumerate the authorized verdicts. It delegates:

> Ship MUST invoke the close path named by the machine-checkable classification defined
> in **P-015** and implemented by the **`shipment-reconcile`** skill, and MUST NOT invoke
> any close path the classification did not name. Ship does not re-derive, restate,
> extend or narrow that classification.

This is what lets **future** close-path work land **without touching this file again** —
including the deferred `TASK_ONLY_FINALIZE` platform work, if a fired trigger ever brings
it back. It also forbids Ship
from authorizing a verdict on its own — A adds **no** verdict.

> **Rev 7 — SINGLE RUNTIME AUTHORITY. The disagreement and version checks are REMOVED
> (plan-review findings E-4/E-5/E-6, NORMATIVE, supersedes the rev-6 fail-closed list).**
>
> **The authority split, stated once:** the **installed `shipment-reconcile` skill is the
> RUNTIME authority** — it is what Ship actually invokes and whose result Ship acts on.
> **P-015 is the STATICALLY REVIEWED authority** — it is what governs whether that skill's
> classification was correct *at review time*, and it is enforced by review, not by Ship at
> run time. **Ship never adjudicates between them.**
>
> **Consequently, three rev-6 requirements are DELETED, not softened:**
>
> * the runtime **`policy_id` check** — the installed classifier emits no such field;
> * the runtime **policy-version comparison** — likewise unemitted, and a version skew is a
>   *review* concern, not a Ship-time one; and
> * the **P-015/skill disagreement check** (`E-5`) — P-015 supplies **no independent machine
>   result** to disagree with. Detecting *semantic* disagreement would force Ship to **re-run
>   the predicates**, which is precisely the duplication **AC-5** forbids. Rev 6 required
>   Ship to compute the thing this plan exists to stop it computing.
>
> **The ordering is fixed: a1 FIRST, then the installed skill (`E-6`).** a1 (§9 step 13) is a
> narrow completion precondition and runs **before** any classification. **Ship MUST then
> invoke the installed `shipment-reconcile`** (§9 step 14–15). There is no path in which Ship
> skips the invocation.
>
> **Output validation — what `_ship.agent.md` MAY name.** Ship must distinguish a valid
> result from garbage, so the file MAY carry an **output-validation allowlist of the
> installed CLOSE VERDICT tokens** — and **nothing more**:
>
> | Token | Kind | Ship's action |
> |---|---|---|
> | `CASCADE` | **close verdict** | invoke the close path the classification named |
> | `SAFE_CLOSE` | **close verdict** | invoke the close path the classification named |
>
> **`HALT` and `RECONCILE_FAIL` are NON-CLOSE PROCEDURAL OUTCOMES, not close verdicts.**
> They authorize no close path at all: Ship reports and stops. Listing them alongside the
> close verdicts (as rev 6 did for `HALT`) blurs a distinction that matters, because a
> procedural outcome can never be "the path the classification named".
>
> **`BLOCKED` is NOT an installed token.** It appears nowhere in the installed skill's result
> vocabulary. It is therefore an **unknown token** and takes the unknown-token HALT path
> below. It must not be invented, honored, or documented as a result.
>
> **What `_ship.agent.md` MUST NOT contain — any classification PREDICATE.** Unchanged from
> rev 6 and still load-bearing: no root/`parent_id` test, no full-coverage test, no
> nothing-beyond-the-root test, no per-member-versus-whole-manifest rule, no fallback rule.
> **The allowlist has no conditions in it** — comparing a returned token against a list of
> tokens that exist is **input validation**; deciding *which* token applies is
> **classification**, and remains wholly delegated.
>
> **Fail-closed behavior — Ship MUST HALT, no mutation, on ANY of:**
>
> * a result token **outside** the installed close-verdict allowlist — including `BLOCKED`
>   and including any token a **future** skill version introduces. **An unknown token is
>   never honored, never guessed, never mapped to a neighbour**; it HALTs and returns the
>   shipment to **Stage**;
> * an **absent**, **empty**, **unparseable** or **ambiguous** classifier result;
> * a classifier that is **not installed**, **not readable**, or otherwise unavailable —
>   **HALT**; there is no default, no assumed verdict, and no "proceed without
>   classification" path; or
> * a **non-close procedural outcome** (`HALT`, `RECONCILE_FAIL`) — no close path is
>   authorized; report and stop.
>
> **Where "expected `CASCADE`" lives (`E-6`, and this is the whole resolution).** Rev 6 put
> §4.1's *"MUST invoke the named path"* against AC-28's *"HALT on a valid `SAFE_CLOSE`"* and
> made the same result simultaneously mandatory and forbidden. **They are not in conflict
> once the two obligations are seated in different documents:**
>
> | Obligation | Lives in | Scope |
> |---|---|---|
> | **Invoke the path the classification names; never one it did not name** | `_ship.agent.md` (**general, ID-free**) | every shipment, forever |
> | **For `021-S` and `017-S` specifically, EXPECT `CASCADE`; halt before any close mutation on anything else** | **this plan's and the backlog records' execution contract** (§9.3, AC-28, `PA-021-CASCADE`, `PA-017-CASCADE`) | these **two exact release instances only** |
>
> **`_ship.agent.md` therefore contains NO global ID-specific close logic** — no `021-S`, no
> `017-S`, no `018.*`, no expected-verdict rule. It states only the general delegation. The
> expectation is a **release-instance precondition** carried by the plan and the shipment
> records, checked by the executing session against **this** plan, exactly as every other
> per-shipment precondition in §9 is. This is the same decoupling rev 6 applied to the §5.0
> 11-ID allowlist (finding C-5), applied to the verdict expectation.
>
> **What the expectation means operationally for these two instances:** `CASCADE` enters the
> **pre-existing** cascade branch; **`SAFE_CLOSE`, an unknown token, an unavailable or
> unreadable classifier, an empty or ambiguous result, `HALT`, or `RECONCILE_FAIL` ⇒ HALT
> BEFORE ANY CLOSE MUTATION** and return to Stage. A `SAFE_CLOSE` here is not "close a
> different way" — it means the live manifest is not the shape this plan asserts, which only
> **Stage** can reconcile.
>
> **BOUND TO THE CURRENT INSTALLED CONTRACT — no unversioned future capability.** The
> allowlist is the close-verdict token set of the **skill installed and reviewed at this
> plan's revision**. A future version that adds a token gains **nothing** from A: its token
> lands outside the installed allowlist and **HALTs**. Widening the allowlist requires that
> change to carry **its own review and its own authorization**. A authorizes **no new verdict
> and no new close path**.

### 4.2 Pinned reconciled C1 wording (normative)

The existing clause says *"Close under the just-merged contract, not a stale in-context
copy."* The §6 carve-out says a reload MUST NOT widen the authority envelope. Left
side-by-side and unreconciled, these two sentences are in direct tension **for A's own
merge**, which changes both procedure and authority. The implementer MUST NOT leave the
reconciliation to inference. C1's reconciled text is pinned:

> Close under the just-merged contract, not a stale in-context copy — **with one
> exception: if the merged change alters this agent's own Role Boundary, the altered
> authority does NOT apply to the session performing the merge.** That session MUST write
> a checkpoint and end; a fresh session loads the new Role Boundary and completes closure.
> This carve-out is scoped to **Role Boundary** changes only; every other merged change —
> procedure, algorithm, gate ordering — takes effect on reload exactly as before.

Go test row 11 asserts the scoping clause is present, so the two sentences cannot ship
unreconciled.

### 4.3 Stage precondition — manifest re-shape (owned, and now COMPLETED)

**Rev 2 status: DONE.** `021-S`'s live manifest is the fully-covered root
`[022-F, 022.001-T]`, verified at this plan's revision. Rev 1 described the re-shape in the
present/future tense (*"`021-S`'s live manifest is task-only `[022.001-T]` … is a required
re-shape"*), which was accurate when drafted but became stale the moment the re-shape was
performed in the same re-plan session. An implementer reading rev 1 would have gone looking
for work that no longer existed.

The re-shape was performed by **Stage** (Ship's Role Boundary forbids editing shipment
planning fields), **before** Ship claims `021-S`.

If it had been skipped, A's manifest would have stayed task-only, the classifier would
return `SAFE_CLOSE` (no feature member ⇒ no `CASCADE`; `TASK_ONLY_FINALIZE` not yet
authorized), and A could not close. That outcome is **fail-closed** (§10.3). **AC-12** is
now a *verification* that the shape is present at claim, not an instruction to create it.

## 5. The Role Boundary grant (normative, narrow, fail-closed)

Authorized transition: a **covering feature** `active -> done`, and **only** when **all**
hold:

1. every **non-pre-archived executable descendant** (at any depth) **declares `status: done`**
   — read from the record's own frontmatter, **never** from which directory holds it
   (§4.4, rev 6);
2. **no** `queued` or `active` live descendant remains;
3. parent/child topology is intact — every descendant resolves to this feature by
   `parent_id`, and no descendant is orphaned or reparented;
4. it is **immediately before** `shipment-reconcile mode: pre` for the shipment whose
   manifest contains that feature;
5. the feature's live status is exactly `active`.

**Any** condition unmet → **HALT**, fail closed. No inference, no fallback, no partial
grant. The authority is **not** general feature lifecycle management: it may not move a
feature to `done` outside Step 6.1(a1), and it grants nothing over features outside the
active shipment's manifest.

### 5.0 What "pre-archived descendant" means for THIS grant (rev 5; **re-based on DECLARED STATUS in rev 6**, normative, fail-closed)

**Plan-review findings A-10, B-1, B-5.** Condition 1 exempts *pre-archived* descendants from
the completion check. Rev 4 inherited pre-mode's *sketch* definition — **archive-located,
accepted without a status check** — which is **unsafe as a completion exemption**. Under that
reading, an unfinished descendant that is merely archive-located stops counting for
condition 1, is not live for condition 2, and can keep an intact `parent_id` for condition 3
— letting Ship complete a covering feature whose work is **not done**.

**Rev 5 tightened the gates but left the TRIGGER locational, which is the B-1 P0.** Rev 5
still keyed the exemption, and §5.1's S0 anomaly gate, off **archive location**. Because the
live registry routes `done` to `archive/`, a **correctly completed** descendant trips a
gate built for **archived** ones. That is what made `021-S` unclosable.

**Rev 6 removes location from the trigger entirely.** "Pre-archived" means **declared
`status: archived`**, and nothing else. Being archive-**located** proves nothing; being
`archived_status: queued` **does not prove completion** and is never treated as proving it.

A manifest descendant is **excluded from condition 1** only when **all five** hold:

1. **it DECLARES `status: archived`** in its own frontmatter — the sole definition of
   "pre-archived" for this grant. **Read directly by a1** from the record, as a **narrow
   completion precondition**; a1 does **not** wait for, depend on, or re-run any pre-mode
   classification (rev 6, finding B-5). A record that is archive-**located** while declaring
   `status: done` is **NOT** pre-archived — it is a **completed** descendant that **satisfies**
   condition 1 normally;
2. **exactly one physical copy** — the record is present in `.backlogit/queue/` **XOR**
   `.backlogit/archive/`. **This is the GENERAL containment check and it applies to every
   member at every status** — a record in **both** is a torn/duplicate state, a record in
   **neither** is missing. It proves **containment and uniqueness ONLY**, never lifecycle
   state (§4.4). **Separately, and only because condition 1 has already established a
   declared `status: archived`**, that single copy is additionally required to be the
   **archive-located** one — a **routing/integrity** condition (the live `registry.yaml`
   routes `archived` to `archive/`, so a declared-`archived` record sitting in `queue/` is a
   routing anomaly), **never a lifecycle inference**;

   > **Rev 7 — WHY THIS IS TWO CONDITIONS AND NOT ONE (plan-review finding E-1,
   > normative).** Rev 6 wrote condition 2 as *"the record exists in `.backlogit/archive/`
   > and **not** in `.backlogit/queue/`"* — which silently required `archive/`
   > **specifically** and so re-derived a lifecycle-relevant fact from placement, the exact
   > B-1 root cause, inside the gate rev 6 was fixing. The two concerns are now separated:
   >
   > | Concern | Predicate | Applies to |
   > |---|---|---|
   > | **Containment / torn-duplicate** (general) | present in `queue/` **XOR** `archive/` | **every** member, **every** status |
   > | **Archive routing integrity** (specific) | the single copy is the `archive/` one | **only** a record that **already declared `status: archived`** at condition 1 |
   >
   > **Order matters and is one-directional.** Condition 1 (**declared status**) determines
   > the **lifecycle class**. Only *after* that class is `archived` does placement become
   > relevant at all, and then only as an **integrity** check on correct routing. **Nothing
   > anywhere reads placement to decide what lifecycle class a record is in.** A
   > `status: done` record is never subjected to the archive-placement condition, which is
   > why `022.001-T` — `done` and archive-located — passes cleanly.
3. **immutable across the gate** — its content hash, declared `status` and `parent_id` are
   unchanged from the pre-gate snapshot;
4. **valid archive provenance** — a well-formed `archived_from`, a present `archived_status`,
   and a resolvable `artifact_type`. **Required only because condition 1 established a
   declared `status: archived`**; it is never demanded of a `status: done` record, whose
   relocation is a registry routing effect that carries no provenance by design; and
5. **it is on the §5.0 exact allowlist below** — a recorded Stage decision mapping it as a
   superseded / non-obligatory closure member: an explicit, named disposition that this
   descendant is **not** part of the obligation the covering feature's completion asserts.

**Condition 5 is the load-bearing one.** Conditions 1–4 establish that the record is
*genuinely and stably archived* **by its own declaration**; only a **Stage decision**
establishes that it is *not owed*. Without it the exemption would still be inferring
completion from a status the descendant did not earn.

**`017-S` — the exact 11 IDs, mapped.** Stage's disposition covers precisely these eleven
manifest descendants of `018-F` that **declare `status: archived`**, and no others:

| # | ID | Kind | Parent |
|---|---|---|---|
| 1 | `018.001-T` | task | `018-F` |
| 2 | `018.001.001-ST` | subtask | `018.001-T` |
| 3 | `018.001.002-ST` | subtask | `018.001-T` |
| 4 | `018.001.003-ST` | subtask | `018.001-T` |
| 5 | `018.002-T` | task | `018-F` |
| 6 | `018.002.001-ST` | subtask | `018.002-T` |
| 7 | `018.002.002-ST` | subtask | `018.002-T` |
| 8 | `018.002.003-ST` | subtask | `018.002-T` |
| 9 | `018.003-T` | task | `018-F` |
| 10 | `018.003.001-ST` | subtask | `018.003-T` |
| 11 | `018.003.002-ST` | subtask | `018.003-T` |

**Any manifest descendant DECLARING `status: archived` that is NOT on this list ⇒ HALT.**
The list is the authorization; membership is not inferable. If `017-S`'s live manifest at
its own gate presents a descendant declaring `status: archived` that is absent here, a1
**halts and returns to Stage**. **Being archive-located is NOT the trigger** — a descendant
declaring `status: done` is never tested against this list, no matter where it is stored.

**`021-S` (A itself) has NO pre-archived descendants — and this remains TRUE AT THE MOMENT
a1 RUNS (rev 6, the B-1 correction).** Its manifest is `[022-F, 022.001-T]`. At §9 step 4
Ship moves **`022.001-T -> done`**, and the live registry relocates it to `.backlogit/archive/`
as a routing consequence. By §9 step 13, when a1 runs, `022.001-T` is therefore
**archive-located while declaring `status: done`**.

Under the rev-5 locational rule that record was an *"archive-located member not on the
§5.0 list"* ⇒ **S0 HALT** ⇒ `022-F` never reaches `done` ⇒ `021-S` unclosable **after A's
irreversible merge**. That was the P0.

Under the rev-6 declared-status rule the same record is unambiguously a **completed
descendant that SATISFIES §5 condition 1**:

| Check at a1 | `022.001-T` | Result |
|---|---|---|
| Declared `status` | `done` | **completed** ⇒ condition 1 **satisfied** |
| Pre-archived? (declares `status: archived`?) | **no** | **not** a pre-archived member; §5.0's allowlist is **never consulted** |
| §5.0 allowlist membership required? | **no** | absence from the 11-ID list is **irrelevant** |
| Archive provenance required? | **no** | registry relocation on `done` carries none **by design**; demanding it was the second, independent S0 HALT — also removed |
| Exactly one physical copy | `archive/` only | containment **satisfied** |
| Live (`queued`/`active`)? | no | condition 2 **satisfied** |
| `parent_id` → `022-F` | intact | condition 3 **satisfied** |

The exemption is therefore **never exercised on A's own route** — not because A's
descendants stay in the queue, but because **none of them ever declares `status: archived`**.
It exists for `017-S` and is gated fail-closed for everything else.

**The general authority stays narrow.** This is an exemption from *one* condition, for
*enumerated* IDs, under a *recorded* Stage decision, triggered **only** by a declared
`status: archived`. It is not a general rule that archived descendants don't count, and it
must not be generalized into one.

> **Rev 6 — where the 11 IDs LIVE (finding C-5).** The enumerated list is a **release-instance
> disposition and belongs to THIS PLAN**, not to the global agent contract. `_ship.agent.md`
> receives only the **general rule** — *"a descendant declaring `status: archived` is exempt
> from condition 1 only under a recorded Stage disposition that names it; absent such a
> disposition, HALT"* — and **never the `018.*` IDs themselves**. This keeps the reusable
> policy decoupled from one release, and it is why no Go test row pins an `018.*` literal.
> Stage owns the disposition and re-states it per shipment.

> **Rev 6 — HOW S0 and §5.0's containment checks are actually executed (finding C-10).**
> `backlogit_get_item` returns a single resolved record and **cannot reveal a torn duplicate**
> (the same ID present in both `queue/` and `archive/`), so containment must not be pinned to
> it. The executable surface:
>
> | Check | Tool |
> |---|---|
> | **Exactly-one-copy containment** (torn duplicate / missing) | `backlogit doctor` (`check_duplicates`, `check_orphans`) and/or `backlogit_query_sql` over the index — **never** `get_item` alone |
> | **Declared `status`, `artifact_type`, `parent_id`** | `backlogit_get_item` / `backlogit get {id}` — the record's own frontmatter |
> | **Archive provenance** (`archived_from`, `archived_status`) on a declared `status: archived` record | `backlogit_get_item` |
> | **Topology at every depth** | `backlogit_query_sql` over `parent_id`, or `backlogit_get_dependencies` / `get_links` as applicable |
>
> **Doctor was run clean this session** (`No issues found.`, 240 artifacts indexed,
> `parse_failures=0`), which is the containment evidence for `021-S`'s two members.

### 5.1 Step 6.1(a1) applicability — a STRICTLY ORDERED, non-overlapping selector (rev 2; re-scoped rev 3; made total in rev 4; **ordered and disambiguated in rev 5**)

Rev 1 specified what a1 does when it applies, but never said what a1 does when the
manifest has **no** feature member. An unspecified step invites an implementer to either
skip the gate silently or invent a transition. Both are wrong.

**Rev 3's three cases were still not total.** They branched on *"the manifest contains a
feature member"* — a predicate that is equally true of a manifest with **one** feature
member and a manifest with **three**. Rev 3 therefore left the multi-feature manifest on an
**unspecified branch**, and the two readings an implementer could reach from that text are
both unsafe: complete *the first* feature member found (silently leaving the others
`active`, and making the outcome depend on manifest ordering), or complete *all* of them
(a bulk feature-lifecycle power §5 explicitly refuses to grant).

a1 MUST therefore **select on multiplicity first**, and the selector MUST be exhaustive
over `n = |{manifest members whose artifact_type is feature}|`:

> **Rev 5 — the selector is a strictly ordered sequence, not a set of predicates
> (plan-review finding A-6).** Rev 4's table let an already-`done` feature member match
> **two** rows at once: case (iii) fires on *"any §5 condition unmet"* and condition 5 is
> *"live status is exactly `active`"*, so a `done` member is told to **HALT**; while §5.2
> **simultaneously** tells it to proceed via idempotent resume. That is the same
> dual-action defect rev 4 fixed for `n > 1`, left unfixed one row above it.
>
> **The fix is ordering, not re-wording.** a1 executes steps **S0 → S4 in order** and stops
> at the first that applies. §5.2 is **not nested inside** any case; it is reached only at
> **S3**. **No input can satisfy two steps** — each step's guard is disjoint from every
> earlier one by construction, because every step after S0 assumes all earlier guards failed.

**S0 — ANOMALY GATE (before `n` is computed).** Evaluated over every manifest member,
**first**, because a malformed record makes the multiplicity count itself untrustworthy.
**Every trigger below is keyed on DECLARED STATUS or on physical CONTAINMENT — never on
which directory holds a record (rev 6, finding B-1).**
**HALT, fail closed, no mutation** on any of:

* a member whose `artifact_type` is **missing, empty, or unresolvable**;
* a member present in **both** `.backlogit/queue/` and `.backlogit/archive/` — a
  **containment** violation (torn/duplicate), not a status inference;
* a member present in **neither** (missing);
* a member **declaring `status: archived`** with **malformed archive provenance** (absent or
  ill-formed `archived_from` or `archived_status`). **Scoped strictly to declared
  `status: archived`** — a `status: done` record carries no provenance by design, and
  demanding it was one of the two rev-5 S0 HALTs that made `021-S` unclosable;
* a member **declaring `status: archived`** that is **not on the §5.0 recorded Stage
  disposition list**.

> **Rev 6 — what S0 explicitly DOES NOT halt on (normative).** A member that is
> **archive-located while declaring `status: done`** is a **normal completed member**. It
> raises **no** S0 trigger, is **not** matched against §5.0's allowlist, and is **not**
> required to carry archive provenance. This is the exact shape `022.001-T` presents at
> §9 step 13 (§5.0), and the exact shape rev 5 halted on.

Only if S0 raises nothing does a1 compute `n`.

> **Rev 6 — THE PRE-MODE CIRCULARITY IS REMOVED (plan-review finding B-5, normative).**
> Rev 5's §5.0 gate 1 required that *"P-015 pre-mode classifies it `pre-archived`"*, taken
> from the authoritative scan. But §9 runs **a1 at step 13** and
> **`shipment-reconcile mode: pre` at step 14**. Pre-mode **has not run** when a1 executes,
> and an earlier intake scan is a different snapshot, so the input had **no defined
> provider** on the `017-S` route.
>
> **The fixed ordering is strictly one-directional, and there is no back-edge:**
>
> | # | Step | Runs | Reads |
> |---|---|---|---|
> | 1 | **a1** (§9 step 13) | **BEFORE** pre-mode | its **own** direct, read-only reads of each manifest member's record |
> | 2 | **feature completion** | inside a1's S4 | — |
> | 3 | **`shipment-reconcile mode: pre`** (§9 step 14) | **AFTER** a1 | the authoritative scan, unchanged |
>
> **a1's completion guard is self-sufficient and READ-ONLY.** It directly checks, per member:
> **(a)** the **declared `status`** in the record's own frontmatter; **(b)** **exactly one
> physical copy** (`queue/` XOR `archive/`); **(c)** topology — `parent_id` resolves to this
> feature, nothing orphaned or reparented; and **(d)** for a member declaring
> `status: archived` **only**, the exact §5.0 11-ID allowlist plus its provenance and
> immutable-metadata checks. Nothing else. **a1 MUST NOT require pre-mode to have already
> run**, and MUST NOT invoke pre-mode itself.
>
> **This is a NARROW COMPLETION PRECONDITION, not a duplicate close-path classification.**
> a1 answers exactly one question — *"is this covering feature's obligation discharged, such
> that `active -> done` is authorized?"* It does **not** compute a close path, does **not**
> emit or consume a verdict, and does **not** decide `PROCEED`/`RECONCILE_FAIL`. The
> **authoritative** pre-mode classification and the P-015 close-path selection run **after**
> a1, **once**, and remain **solely** authoritative for everything they own (§4.1, §9.3).
> Reading a record's own declared status to decide whether an authority's precondition holds
> is not re-deriving a classification — it is the ordinary evidence any guard must gather.

**S1 — `n == 0` ⇒ NOT APPLICABLE.** Record `A1_NOT_APPLICABLE: no feature member in
manifest` and proceed to Step 6.1(a). A **successful** outcome — not a skip, not a failure,
and **no** transition of any kind.

**S2 — `n > 1` ⇒ AMBIGUOUS, HALT.** Record `A1_MULTIPLE_FEATURE_MEMBERS: {ids}`. **Evaluate
no §5 condition and mutate no feature**, not even one that would individually qualify.
Return to **Stage** for manifest disposition. *(Placed before the `n == 1` branches so the
single-member steps below may assume a unique referent.)*

**S3 — `n == 1` and that member's live status is exactly `done` ⇒ IDEMPOTENT NO-OP —
but only after revalidation.** Read the member's live status per **PO-1b** (§10.1).
Re-evaluate §5 conditions **1, 2, 3 and 4** — every non-status guard. Then:

* **all four hold** ⇒ record `A1_ALREADY_DONE: {feature_id}` and proceed to Step 6.1(a)
  **without error and without re-issuing the transition**;
* **any of the four fails** ⇒ **HALT** (this is the S4 failure disposition), because a
  feature that is `done` while its descendants are *not* is a corruption, not a resume.

Condition 5 is **not** evaluated at S3 — reaching S3 *is* the status determination. This is
the only step that may observe a `done` member, so no other step can give a second answer.

**S4 — `n == 1` and live status is exactly `active` ⇒ evaluate the full §5 guard.**

* **all five conditions hold** ⇒ perform `active -> done` on that member **via
  `backlogit_move_item` (MCP) — CLI fallback `backlogit move {id} --status done`** — and
  record the transition;
* **any condition unmet** ⇒ **HALT**, fail closed. Surface which condition failed. Do **not**
  proceed to Step 6.1(a). Do **not** widen any condition in-flight.

> **Rev 7 — the transition tool is PINNED here, not only in §10.1's PO-1b box (finding D-7 /
> `F-10`).** Rev 6 stated the `active -> done` transition abstractly at S4 and at §9 step 13,
> leaving the tool pin to a distant section. **The registry-declared operation is
> `move_task` → MCP `backlogit_move_item`, CLI `backlogit move {{id}} --status {{status}}`.**
> An implementer must not reach for `backlogit update --status`, a direct file edit, or any
> other path. The same pin is repeated at §9 step 13 so both statements of the transition
> carry it.

**S5 — `n == 1` and live status is anything else** (`queued`, `blocked`, `review`, or any
other value) ⇒ **HALT**, fail closed. Record the observed status. Return to **Stage**.

| Step | Guard (assumes every earlier guard failed) | Outcome |
|---|---|---|
| **S0** | any member anomaly — declared-status or containment keyed, **never** location | **HALT** — no `n`, no mutation |
| **S1** | `n == 0` | explicit no-op ⇒ proceed to `a` |
| **S2** | `n > 1` | **HALT** — no condition evaluated, no mutation |
| **S3** | `n == 1`, status `done` | conditions 1–4 revalidate ⇒ idempotent no-op; else **HALT** |
| **S4** | `n == 1`, status `active` | all five ⇒ `active -> done`; else **HALT** |
| **S5** | `n == 1`, any other status | **HALT** |

**Totality and disjointness.** S1/S2/S3/S4/S5 partition the input exactly: `n` is either
`0`, `>1`, or `==1`; and in the `==1` case the live status is either `done`, `active`, or
something else. Every input reaches **exactly one** step, and **no input reaches two**.

> **Legacy case-label map (rev 2–4 → rev 5).** Prior revisions, the acceptance criteria and
> the appended review all refer to cases (i)–(iv). The labels are retained for traceability
> and map onto the ordered steps as follows. Where a rev-4 label was ambiguous, the map
> records which step now owns the input:
>
> | Legacy label | Rev-5 step | Note |
> |---|---|---|
> | case (i) — not applicable | **S1** | unchanged |
> | case (ii) — applies, proceed | **S4** (all five hold) | `done`-member sub-case **moved out** to S3 |
> | case (iii) — applies, halt | **S4** (guard unmet), **S5** (wrong status), **S3** (revalidation failure) | rev 4's single label covered three distinct inputs |
> | case (iv) — ambiguous, halt | **S2** | unchanged; now evaluated **before** the `n == 1` steps |
> | §5.2 idempotent resume | **S3** | no longer nested inside case (ii) |
> | *(new in rev 5)* | **S0** | the anomaly gate — no legacy label existed |

**`n` is computed by `artifact_type`, never by ID suffix.** Subtask IDs end `-ST`, whose
last character is also `T`; feature IDs end `-F`. A suffix test is the same class of defect
§R16.3.3 of the policy-gap plan records for the executable-set filter. Read
`artifact_type` from each manifest member's record.

> **Rev 3 — `017-S` is case (ii), NOT case (i).** Rev 2 cited `017-S` as the live example
> of case (i), on the basis that it was a **12-member task-only** manifest. That is no
> longer true. Per
> `docs/decisions/2026-09-13-intercom-go-a-only-simplification-decision.md`, `017-S`'s
> root `018-F` has been **restored to the manifest**, giving **13** members and making it
> a **fully-covered root**. At `017-S`'s gate, a1 therefore takes **case (ii)**: `018-F`
> is a present feature member, the five §5 conditions hold, and a1 moves it
> `active -> done` — after which pre-mode passes and classification returns **`CASCADE`**.
>
> **Case (i) is retained regardless**, and is still asserted by Go test row 12. It is now
> a **defensive completeness** requirement rather than a case this chain is known to
> exercise: the step must be total over its input, and a future task-only manifest must
> not be able to reach an unspecified branch. Retaining it costs one recorded no-op and
> removes an entire class of ambiguity.

> **Rev 4 — both live manifests are `n == 1`; case (iv) is defensive.** `021-S` is
> `[022-F, 022.001-T]` ⇒ `n == 1`. `017-S` is the 13-member root ⇒ `n == 1` (`018-F`; the
> other 12 members are tasks and subtasks). **No current shipment reaches case (iv)** —
> it exists so that the step is **total**, exactly as case (i) does.

> **DISCLOSED NARROWING — case (iv) restricts a manifest shape P-015 explicitly authorizes
> (rev 4, plan-review finding A-5).** This is recorded honestly rather than claimed away.
>
> P-015's fully-covered-root exception is *"quantified over **every feature member of the
> manifest**"* (item 1) and item 4 speaks of *"the qualifying root feature **member(s)**"* —
> **plural is authorized by policy**. A manifest with two qualifying root features, each
> fully covered, with nothing extra, **satisfies P-015 today**.
>
> Case (iv)'s unconditional `n > 1` HALT therefore makes such a manifest **unclosable by
> Ship**, because a1 halts before pre-mode ever runs. That is a **real narrowing of
> reachability**, and it sits in tension with §4.1's constraint that Ship *"does not
> re-derive, restate, extend or **narrow** that classification."*
>
> **Both statements are kept, and the tension is resolved explicitly rather than by
> deletion:** §4.1 governs the **verdict selection** — Ship still may not re-derive or
> narrow *which close path the classifier names*. Case (iv) is an **upstream mutation
> guard**: it declines to exercise the §5 grant when the grant has no unique referent, and
> it halts **instead of** mutating. A halt that returns the shipment to Stage is not a
> re-classification.
>
> **This is nonetheless a narrowing, and the claim that A "narrows nothing" is
> WITHDRAWN.** The accepted cost: a legitimate multi-root manifest must be dispositioned by
> **Stage** (split into one shipment per covering feature, or re-planned) rather than closed
> by Ship. **Recorded follow-up**, not charged to A: make a1 set-based — validate *every*
> feature member against the same five §5 conditions under one lock and transition all of
> them — which would restore full P-015 reachability. Deliberately **not** done here: it
> widens the grant from one feature to N, which is a materially larger authority increase
> than A is scoped to make, and it has no current consumer.

**The distinction between S1 and S4/S5 is load-bearing.** S1 is "there is nothing
here for this gate to do". S4/S5 are "there is something here and it is not in the
required state". Collapsing them — treating an unmet guard on a real feature member as a
benign no-op — would let a shipment close with a covering feature left `active` or with
descendants unfinished, which is precisely the corruption the gate exists to prevent. A
zero-feature-member manifest must never be reported as a guard failure either: that would
make any task-only shipment un-closable at the gate.

**S2 is fail-closed for a different reason than S4.** S4 halts because a known
member is in the wrong state. S2 halts because **a1 cannot determine which member the
grant is about.** §5 grants completion of *"a covering feature"* — singular, and scoped to
"the active shipment's manifest". With two feature members there is no unique referent, so
**the grant does not apply at all** and a1 has no authority to mutate either one. Halting
without evaluating conditions is deliberate: evaluating them first would invite an
implementer to proceed when exactly one of the several happens to qualify, which is the
manifest-ordering-dependent behavior this step exists to forbid.

Asserted by **AC-15** (S1), **AC-16** (S4), **AC-17** (S3) and **AC-21** (the full ordering).

### 5.2 Idempotent resume (rev 2; **re-seated as selector step S3 in rev 5**)

A1 MUST be **idempotent**. If the feature member's live status is already `done` when a1
runs, a1 reaches **selector step S3** (§5.1) — **not** case (ii), and **not** case (iii).

> **Rev 5 — the ordering is now explicit, and the no-op is CONDITIONAL.** Rev 4 stated the
> resume as an unconditional proceed, which (a) collided with case (iii)'s halt and (b)
> would have let a1 wave through a feature that is `done` while its descendants are not.
> **S3 records `A1_ALREADY_DONE: {feature_id}` and proceeds ONLY after §5 conditions 1, 2,
> 3 and 4 — every non-status guard — are RE-EVALUATED and all hold.** If any of the four
> fails, S3 **HALTs**. An already-`done` feature is a resume only when everything that made
> it completable is still true.

This is not hypothetical tidiness — it is required by A's own two-session closure. If S2
completes `022-F -> done` and then fails anywhere between a1 and the close (a
`RECONCILE_FAIL`, an evidence re-verification failure, an operator halt), the recovery
session re-enters at a1 with the feature already `done`. Without idempotent resume,
condition 5 (*"live status is exactly `active`"*) would fail closed and A could **never** be
closed by any session — a permanently stuck shipment created by its own partial success.

Idempotent resume is scoped tightly: it recognises the **already-satisfied terminal state**
of the transition a1 itself performs. It does **not** relax conditions 1–4 — they are
**re-verified on every entry, including this one** — and it does **not** authorize any
transition from any status other than `active`. A feature in `queued`, `blocked`, or any
other status reaches **S5** — **HALT**. Asserted by **AC-17**.

## 6. The authority-escalation rule (load-bearing)

A merged change to Ship's **Role Boundary** MUST NOT take effect mid-session.

* A reload MAY update the **procedure** a session follows.
* A reload MUST NOT widen the **authority envelope** of the session performing it.

Role Boundary is a **session-identity** constraint fixed at session start. Without this
rule, the existing pre-self-close reload would let the session that merged A immediately
use the authority A introduced — an agent granting itself authority by merging a change
to its own contract.

Therefore: **the session that merges A checkpoints and ends.**

## 7. Ordering (test-first; `main` stays green)

1. **H0 — red.** Add the single table-driven test function. The readiness guard finds the
   `a1` anchor **absent** and fails once with the message
   `not implemented: ship feature-completion delegation contract` (§8.1.1). Red count
   **1 of 1**; `go vet` clean; `go test ./...` non-zero.
2. **H1 — green.** Apply ADD-1, ADD-2, C1–C6. Re-run: `go vet` clean, `go test ./...` green,
   `markdownlint` clean.

Exactly **one** generated test function, so Ship Step 2's feature-scoped batch yields one
function and Step 4.3 cannot deadlock (the `021.001-T` constraint).

## 8. Verification obligations

### 8.1 What the Go test proves

**Contract text and ordering on `_ship.agent.md` only.** **18 rows** (rev 5: 16 → 18).

| # | Assertion | Kind |
|---|---|---|
| 1 | Role Boundary row contains the narrow feature-completion grant | positive |
| 2 | The grant names all five conditions of §5 | positive |
| 3 | Step 6.1(a1) exists and is positioned **after** `a0` and **before** `a` | **ordering** |
| 4 | The delegation block is present and names P-015 + `shipment-reconcile` | positive |
| 5 | The reload clause (C1) also names P-015 and the merge commit | positive |
| 6 | The authority-escalation rule (§6) is present | positive |
| 7 | `_ship.agent.md` no longer contains the drifted `children`-only coverage rule | **negative** |
| 8 | `_ship.agent.md` does **not** name `TASK_ONLY_FINALIZE` | **negative** |
| 9 | `_ship.agent.md` states **no classification PREDICATE** — a result-token allowlist is permitted, a selection condition is not (§4.1 rev 6) | **compound +/−** |
| 10 | The `backlogit move <shipment_id> --status shipped` prescription (C2) is gone | **compound +/−** |
| 11 | C1's reconciled wording scopes the carve-out to **Role Boundary** changes only (§4.2) | **ordering/consistency** |
| 12 | Step 6.1(a1) states the **zero-feature-member explicit no-op** (S1) and that an unmet guard on a present feature member **halts** (S4) | **ordering/consistency** |
| 13 | Step 6.1(a1) states **idempotent resume** — a feature member already `done` proceeds without error and without re-issuing the transition, **after conditions 1–4 revalidate** (S3) | **ordering/consistency** |
| 14 | **(C7)** Step 0.5 item 6's pre-mode summary **PRESENT**: the reconciled **declared-status** phrase (§8.1.1); **AND ABSENT**: the exact old queue-only literal | **compound +/−** |
| 15 | **(C8)** Step 6.1(a)'s pre-archive gate summary **PRESENT**: the reconciled **declared-status** phrase (§8.1.1); **AND ABSENT**: the exact old queue-only literal | **compound +/−** |
| 16 | Step 6.1(a1)'s selector states **`n > 1` halts with no feature mutation** (S2) **AND** the anomaly-gate-before-`n` rule (S0) | **compound ++** |
| 17 | **(rev 5, B-11)** C8's reconciled sentence **PRESERVES** the single-writer lock clause **and** the orphan scan clause | **preservation** |
| 18 | **(rev 5, B-14)** C7's adjacent `Scope note (139-F/139.001-T)` **survives** the C7 replacement | **preservation** |

> **Rev 6 — Kind labelling is now consistent (finding D-2).** Rev 5 called rows 14/15/16
> "compound" while marking rows 7/10 "negative", even though §8.1.1 pinned **both** a PRESENT
> and an ABSENT literal for 7 and 10. Three distinct kinds are now named explicitly:
>
> * **`compound +/−`** — one PRESENT literal **and** one ABSENT literal on the same site:
>   rows **9, 10, 14, 15** = **4 rows**. This is the count AC-23 and §12 track.
> * **`compound ++`** — two PRESENT assertions, no ABSENT literal: row **16** only.
> * **`negative`** — ABSENT literal only: rows **7** and **8**. Row 7 is ABSENT-only by
>   deliberate rev-6 correction (finding C-2 — its rev-5 PRESENT needle was itself a
>   forbidden predicate); row 8's bare-token absence is correct for `TASK_ONLY_FINALIZE`, a
>   string the file must never contain in any form.
>
> **The compound count is unchanged at 4** — row 7 left the set and row 9 joined it — so
> **§12 carries no test-row delta for this change** (§12 rev-6 note).

#### 8.1.1 Pinned literals (rev 5, normative — plan-review finding A-7; **re-pinned in rev 6 for B-1/B-2/B-3**)

Rows 14–15 were **vacuously satisfiable** in rev 4: §8.1 said they *"reuse the same
substring-assertion helper"*, i.e. a whole-file `strings.Contains(content, "pre-archived")`.
That passes if the token appears **anywhere** in the ~94 KB file while L255–256 and
L776–777 keep their exact queue-only text — and it may be **green at H0**, never
transitioning red→green. Since rows 14–15 (AC-22) are the **entire mechanical
justification** for reopening the inventory, that would let the plan go all-green with its
headline defect intact.

**All FOUR `compound +/−` rows below assert a long distinctive PRESENT phrase that only the
reconciled wording can contain, AND assert the site's exact old literal is ABSENT. Row 7 is
ABSENT-only** (see the rev-6 note after the table). All needles are **single-line** (C-6: the
target file may be CRLF on Windows), are **verified present exactly once** in the installed
file at the line shown, and are pinned **verbatim** here:

| Row | MUST BE ABSENT (exact current literal) | MUST BE PRESENT (distinctive reconciled phrase) |
|---|---|---|
| 14 (C7) | ``every manifest item is present in `.backlogit/queue/` with the`` | ``defers every per-item status decision to that skill's `mode: pre` classification`` |
| 15 (C8) | ``queue with `status: done`, and scans for orphan items.`` | ``defers every per-item status decision to the `mode: pre` classification at `expected_status: done``` |
| **9** | ``qualification is never per-member, and no feature ID is ever special-cased.`` | ``HALT on any result token the installed classification did not name`` |
| 10 | ``<shipment_id> --status shipped`` | ``the authoritative close sequence defined by the `shipment-reconcile` skill`` |
| 7 | ``it is fully covered (every one of its children,`` | **— none (ABSENT-only; see below)** |

> **Rev 6 — row 7 loses its PRESENT needle, which was self-contradictory (finding C-2).**
> Rev 5 pinned row 7's PRESENT phrase as ``descendants — at every depth, not only direct
> children``. That phrase is **itself a coverage PREDICATE** — precisely what §4.1's rev-6
> allowlist contract and row 9 forbid `_ship.agent.md` from restating. Satisfying row 7 would
> have **forced the implementer to re-add a predicate** in order to pass a row whose purpose
> is removing one, and it appeared in **no** pinned replacement text.
> **Row 7 is therefore ABSENT-only, and that is sufficient**: its ABSENT literal is the
> **full drifted single-line clause**, verified to occur **exactly once** in the installed
> file (L810), so the row is **non-vacuous on its own** and cannot be green at H0. The
> delegation that replaces the deleted predicate is asserted positively by rows **4** and
> **10**, so nothing goes unchecked.

> **Rev 7 — rows 14/15 are re-pinned to DELEGATION wording (findings E-2/E-3).** Rev 6's
> PRESENT literals (``never inferring it from `queue/` or `archive/` location`` and
> ``either declares `status: done` or declares `status: archived```) encoded a **declared-status
> PREDICATE**. Satisfying them would have forced the implementer to write a per-item
> classification rule into Ship's instruction surface — the same duplication §2.2 exists to
> remove, and (per E-3) a rule the installed pre-mode does not document. **The rev-7 literals
> above can only be satisfied by DELEGATION wording**, and are unsatisfiable by any restated
> predicate. **Rows 14/15 therefore now test DELEGATION; row 17 tests PRESERVATION** (lock and
> orphan scan) **and row 18 tests SCOPE-NOTE SURVIVAL** — three distinct properties, no
> duplicated predicate anywhere. **The ABSENT literals are unchanged and re-verified**
> (row 14 at L255, row 15 at L777).
>
> **Row 9 is COMPOUND (finding B-3).** Its ABSENT literal is a **classification
> predicate** — the per-member/whole-manifest fallback rule at L819, verified present and
> **unique** in the installed file — so the row goes green only when the predicate block is
> genuinely replaced by delegation. Its PRESENT literal asserts the **fail-closed validation
> behavior**, not a verdict list. **Row 9 therefore does not contradict AC-25**: both require
> the same thing — close-verdict tokens may be listed, predicates may not.
>
> **Reconciled target wording (normative, so the literals are reachable). POINTER-ONLY —
> rev 7.**
>
> * **C7** (L255–256) becomes: *"This delegates the per-item check to the
>   `shipment-reconcile` skill's `mode: pre` contract at the `expected_status` above and
>   **defers every per-item status decision to that skill's `mode: pre` classification**,
>   continuing **only** on an authoritative `PROCEED`; it also requires a `record-consistent`
>   shipment record and **scans for orphan items**."*
> * **C8** (L775–777) becomes: *"This acquires the single-writer lock on
>   `.backlogit/queue/{shipment_id}.md` **(via the `file-lock` skill)** and **defers every
>   per-item status decision to the `mode: pre` classification at `expected_status: done`**,
>   continuing **only** on an authoritative `PROCEED`, requires that the
>   shipment-record-status classification is `record-consistent`, and **scans for orphan
>   items**."*
>
> **Neither sentence states a classification predicate of any kind.** Both preserve the
> `PROCEED` conjuncts (**no orphans**, **`record-consistent`**) and C8 preserves the
> **single-writer lock** clause, so rows 14/15 and row 17 are jointly satisfiable. Row 18's
> `Scope note (139-F/139.001-T)` at L259 is untouched by C7's sentence.

> **Rev 6 — ABSENT-needle uniqueness MEASURED, not assumed (finding D-5; RE-MEASURED at the
> current tree in rev 7).** Each ABSENT literal was counted across the whole installed file
> again this session; every one occurs **exactly once**, so no row can pass or fail on an
> unrelated match:
>
> | Row | ABSENT literal occurrences | Line |
> |---|---|---|
> | 14 | 1 | L255 |
> | 15 | 1 | L777 |
> | 7 | 1 | L810 |
> | 9 | 1 | L819 |
> | 10 | 1 | L790 |
>
> **Rev 7 additionally MEASURED the PRESENT needles at the current tree**: all four
> (rows 14, 15, 9, 10) occur **zero** times today, so every `compound +/−` row is a genuine
> **red → green** transition and none can be vacuously green at H0.
>
> Row 3's three ordering anchors were re-counted the same way (L767, L773, and the `a1`
> anchor which is absent pre-implementation **by design** — that absence is the H0 readiness
> guard, not a row failure). Row 17's ``scans for orphan items`` occurs **twice** (L256,
> L777) — correct and intended for a **preservation** needle, which asserts survival, not
> uniqueness. Row 18's ``Scope note (139-F/139.001-T)`` occurs **once** (L259).

**Negative needles are pinned to the FULL drifted clause, never a bare token (B-3).** A bare
`children` needle can **never** go green — `_ship.agent.md` legitimately uses the word
elsewhere — and would permanently block H1. Row 7 therefore asserts absence of the
**full single-line clause** above, not the word.

**Row 3's ordering anchors are unique (B-2).** Rev 4 anchored on `"Step 6.1(a1)"`, which
**does not exist** in the installed file — it uses `a0.`/`a.`/`b.` list markers, so a naive
`strings.Index("a.")` matches the first `a.` in a 94 KB document. Pinned anchors:

| Anchor | Literal |
|---|---|
| `a0` | ``TOPOLOGY_GATE: lifecycle (before closure/safe-close)`` |
| `a` | ``Pre-archive reconciliation gate (mandatory)`` |
| `a1` (new, unique) | ``a1. **Covering-feature completion gate (P-015 grant)**`` |

Row 3 asserts `index(a0) < index(a1) < index(a)`, all three found, each exactly once.

**Rows 17–18 assert PRESERVATION, not replacement.** Row 17 pins ``(via the `file-lock`
skill)`` and ``scans for orphan items`` as still-present after C8's rewrite; row 18 pins
``Scope note (139-F/139.001-T)`` as still-present after C7's. These exist because the most
likely implementation error is a wholesale sentence replacement that silently deletes
surrounding clauses.

**H0 red is `1 of 1` by construction (B-4; LOGIC CORRECTED in rev 6, finding C-3).** An
18-row table naturally yields many reds. A single **readiness guard runs BEFORE the row
loop**, and rev 5 stated its polarity **backwards** — it described failing when a marker was
*absent* and then claimed the rows execute once that same marker *is gone*, which taken
literally means the rows never run and H1 never greens. The guard is pinned correctly here:

| Phase | Readiness anchor ``a1. **Covering-feature completion gate (P-015 grant)**`` | Guard | Rows |
|---|---|---|---|
| **H0** (pre-implementation) | **absent** | `t.Fatalf("not implemented: ship feature-completion delegation contract")` — fails **once** and returns | **do not execute** |
| **H1** (post-implementation) | **present** | passes | **all 18 execute** |

The **readiness anchor is the file content** (the unique `a1` heading of §8.1.1's anchor
table, which exists only after implementation); `not implemented: ship feature-completion
delegation contract` is the **failure message**, never a string searched for in
`_ship.agent.md`. Conflating the two is what produced the inverted description. H0 is
therefore **exactly one** failure, and H1 executes the full table. Each row runs under its
own `t.Run(row.name, …)` for traceability (C-2), so the table is **real assertions from the
start**, not a stub.

**Assertions are stdlib only (B-5).** `technology-go.instructions.md` recommends testify,
but **testify is not in `go.mod`**; importing it would add a dependency and break AC-11's
2-file surface. Use `strings.Contains` / `strings.Index` + `t.Errorf`. **Reuse the existing
package-level `repoRoot(t)` from `build_script_test.go`** — redefining it is a
duplicate-declaration **compile failure** at H0 (C-5).

Rows 7–10 are the **widening guards**: they fail if this change authorized anything new.
Row 11 is the **coherence guard**: it fails if C1 ships with the carve-out and the
"close under the just-merged contract" directive unreconciled.
**Rows 12–13 are the applicability guards** (rev 2): they fail if a1 ships without an
explicit not-applicable outcome — which would leave any task-only manifest ambiguous at
the gate — or without idempotent resume, which would make a partially-completed A
permanently unclosable.
**Rows 14–15 are the truthfulness guards** (rev 4, made non-vacuous in rev 5): they fail if
either pre-mode summary ships still asserting that every manifest item lives in
`.backlogit/queue/` — an assertion Probe 25 **measured false** for **all 13** members of a
completed manifest (`MEMBERS_QUEUE_LOCATED=0`).
**Row 16 is the totality guard**: it fails if a1's selector ships without a specified
`n > 1` branch or without the S0 anomaly gate.
**Rows 17–18 are the preservation guards** (rev 5).

> **Rev 3 note (superseded by rev 4, and again by rev 5).** Rev 3 recorded the row count as
> **unchanged at 13**. **Rev 4** added rows 14–16 (→ 16). **Rev 5** adds rows 17–18 (→ **18**)
> and makes rows 14–15 and 7/10 **compound** with verbatim-pinned literals. Rows 1–13 keep
> their numbering and assertion text, so rev 2/rev 3 traceability is intact.

**Helper budget.** Still **one** table-driven test function and **≤4 helper functions**
(§3). The helpers are: `repoRoot` *(reused, not redefined)*, `mustContain`, `mustNotContain`,
and `indexOfUnique`. Compound rows call `mustContain` **and** `mustNotContain` on the same
row — **two calls, not a new helper**. Ordering rows call `indexOfUnique`. **Helper count is
unchanged at ≤4**, and the function count is unchanged at **1**, so Ship Step 4.3's
one-function constraint still holds.

### 8.2 What it does **not** prove

It does **not** execute a shipment closure, does **not** prove the engine's cascade
behavior, and does **not** prove that shipment claim leaves a covering feature `active`.
That last point is §10 risk R-1 and is discharged by **probe obligation PO-1**, not by
the test. Claims stay honest.

## 9. Self-hosting closure sequence

`021-S` manifest becomes `[022-F, 022.001-T]` — a **fully-covered root**: `022-F` is a
root (no `parent_id`), its complete descendant set at every depth is exactly
`{022.001-T}` (verified: no `022.001.*-ST` exists), and the manifest contains nothing
else. It therefore qualifies under the **existing, already-reviewed** P-015 exception.

| Step | Session | Action |
|---|---|---|
| 0 | **Stage** | **Pre-execution readiness check (PO-1a)** — **DISCHARGED**, see §10.1. Probe 22 measured the claim shape; no pre-implementation blocker remains |
| **0a** | **Stage** | **PROBE 25 — root-included cascade fixture (§9.4). DISCHARGED.** Authored, executed and **committed** by Stage **before `021-S` is claimed**. `PROBE25_RESULT=PASS`, 8/8 criteria (9 tokens). Stage records the four Hardening-9a identity fields — `evidence_commit_sha`, `fixture_script_sha256`, `fixture_result_sha256`, `backlogit_executable_sha256` — and confirms both files are **tracked and committed**. **`021-S` is not claimable until this evidence is committed and PASSing.** Ship never authors this artifact (P-010) |
| 1 | S1 | Verify on `main`; Step 0.5 pre-claim topology gate; **claim `021-S`**. Ship moves **`022.001-T -> active`**. `022-F` is left **`active`** — **measured, not assumed** (Probe 22 ARM AB) |
| 2 | S1 | H0 red → implement → H1 green |
| 3 | S1 | **Step 4.3 quality gates; Step 4.4 review gate** |
| 4 | S1 | **Step 4.5 Complete Task — commit, then `022.001-T -> done`.** *(Installed order: Step 4.5 (L517) precedes Step 5 (L554). The task reaches `done` BEFORE the implementation PR merges.)* |
| 5 | S1 | **Step 5 PR lifecycle** — full gate sequence, `--phase lifecycle` topology gate, build, push, PR, **P-018 Copilot review engagement (rev 6, finding C-9)**, operator approval, **merge** |
| 6 | S1 | **Step 6 Merge Confirmation Gate** — `gh pr view` state `MERGED`; `git fetch origin main`; `git merge-base --is-ancestor {merge_sha} origin/main` |
| 7 | S1 | **Reload merged `main`.** Detect that the merged change alters **Ship's own Role Boundary**. Per §6 the new authority is **not** available to this session |
| 8 | S1 | **Write checkpoint** (§9.1 payload contract: `schema_version: 1`, `agent: ship`, **`session_id`**, `phase: awaiting-fresh-session-parent-completion`, `resume_hint` naming `021-S`, `022-F`, `022.001-T` and the merge SHA; domain data under `context`) and **END**. **Record the emitted checkpoint filename** — S2 must match it exactly |
| 9 | **S2** | **Fresh session.** Full checkpoint-recovery protocol (§9.1, thirteen steps) — enumerate all with no filter, **anomaly gate over the full enumeration FIRST**, then **filter to the MATCH set** (filename + `agent: ship` + `session_id`) and require **exactly one MATCH**; unrelated active checkpoints are a warning only. Explicit owner selection, ownership validation, operator confirmation, **restore → prune/gate → resume** (Engram-reachability fail-closed; never-prune allowlist honored), same cursor, resolve-after-confirmed-resume. Loads the new Role Boundary + Step 6.1(a1) from merged `main` |
| 10 | S2 | Verify the merge SHA is an ancestor of `origin/main`; validate the **same** active shipment `021-S` |
| 11 | S2 | **Step 6.0 Post-Merge Branch Protocol** — `git checkout main`, `git pull`, `git checkout -b post-merge/022-ship-feature-completion`. **Created BEFORE any post-merge backlog mutation**, because Step 6.1(e) commits `.backlogit/` and those commits must not land on `main` |
| 12 | S2 | **Step 6.1(a0)** `--phase lifecycle` topology gate — run **while `021-S` is still `active`**. *(Probe 20 tripwire: post-archive this gate fails closed on every route; it must never be re-run after the close.)* |
| **12a** | **S2** | **READ-ONLY RE-VERIFICATION of Stage's committed Probe-25 evidence (§9.4a).** All **six** checks: exact topology, PASS fields (9 tokens), **`evidence_commit_sha`** ancestry of the merge SHA, **currently-resolved `backlogit_executable_sha256`**, per-file `fixture_script_sha256`/`fixture_result_sha256` against committed blobs, and transcript internal consistency. **S2 authors, modifies and re-runs NOTHING.** FAIL ⇒ HALT naming the failed field, with the §9.4a recovery path; `021-S` stays `active`, no a1 mutation |
| 13 | S2 | **Step 6.1(a1)** — §5.1 step **S4** (`n == 1`, status `active`). **Runs BEFORE pre-mode** and reads declared status directly (§5.1 rev 6). `022-F` is a manifest member; `022.001-T` declares `status: done` (archive-located by registry routing — a **completed** descendant, **not** pre-archived, §5.0); all five §5 conditions hold → `022-F active -> done` **via `backlogit_move_item` (MCP); CLI fallback `backlogit move 022-F --status done` (rev 7, D-7)** |
| 14 | S2 | **Step 6.1(a)** pre-mode, `expected_status: done` — **both** members declare `done` → **`matched`** → `PROCEED` |
| 15 | S2 | **Authoritative** classification returns **`CASCADE`** (fully-covered root); **any other verdict is a topology discrepancy ⇒ HALT to Stage, no `SAFE_CLOSE` fallback** (§9.3 rev 6). On `CASCADE`: **6.1(b)** close via the **pre-existing** cascade op; **6.1(c)** P-007 archive-integrity verify; **6.1(d)** post-mode; **6.1(e)** commit `.backlogit/` **on the closure branch** |
| 16 | S2 | **Step 6 item 2** `operational-closure mode=post-merge` → `docs/closure/`; **P-020** compact-context finalizes the compaction status — **all on the closure branch, before the closure PR is pushed** |
| 17 | S2 | **Step 6 item 9** sync — **MCP `backlogit_sync_index` first, CLI `backlogit sync` only as the declared fallback (rev 6, finding D-6)** — then push the closure branch, closure PR, **P-018 Copilot review engagement (rev 6, C-9)**, P-014 local review, operator approval, merge |
| 18 | S2 | **Step 6 item 10** return to `main`; `git pull` |
| 19 | — | Orchestrator routes **`017-S`** only (rev 3 — A is the sole prerequisite; there is no B) |

**Why the order in steps 4–5 and 11–17 is exactly this.** Read from the **installed**
`_ship.agent.md`, not inferred:

| Installed anchor | Consequence |
|---|---|
| Step 4.5 (L517) precedes Step 5 (L554) | task `-> done` **before** the implementation PR merges |
| Step 6.0 (L726) items 2–3 — branch from fresh `main`; *"All subsequent Step 6 work happens on this branch"* | closure branch exists **before** the 6.1(e) backlog commit |
| Step 6.1(a0) — *"the shipment-scoped check immediately preceding the safe-close mutation itself"* | lifecycle gate runs while `021-S` is still `active` |
| Step 6.1(a)→(b)→(c)→(d)→(e) | pre-mode → close → P-007 → post-mode → commit |
| Step 6 item 2 and item 8 (P-020) precede item 9 (`sync`) and the 6.0 item 4 push | closure artifacts + P-020 land on the closure branch **before** the closure PR is pushed |

Rev 1 placed `022.001-T -> done` **after** the merge, and placed operational closure + P-020
as a single late step without establishing the branch first. Both contradicted the
installed contract. Corrected above.

### 9.1 S2's checkpoint recovery is the FULL protocol, not a resume shortcut (rev 2; **completed in rev 4**)

Rev 1 said only that S2 *"restores the checkpoint"*. That is the step most likely to be
short-circuited, so it is specified in full.

> **Rev 4 — rev 3's list was itself incomplete.** Rev 3 enumerated eight steps and
> **omitted three mandatory, installed, fail-closed elements**: the
> `restore → prune/gate → resume` **ordering**, the **Engram reachability** gate, and the
> **never-prune allowlist**. All three are installed in this workspace
> (`agent-engram` is in `capability_packs`), so rev 3's list was not a safe abbreviation —
> it was a specification that, if followed literally and completely, would still **violate
> the installed protocol**. The authoritative sources are
> `.github/instructions/backlogit.instructions.md` §**Checkpoint-Recovery / Prune-on-Restore
> Protocol** and `_ship.agent.md`'s own **Crash-Resumption / Startup Recovery Protocol**.
> S2 MUST execute the installed protocol as written; the list below is a **pointer and a
> completeness checklist**, never a substitute for reading it.

S2 MUST, **in this order**:

0. **Know the expected checkpoint before enumerating. (rev 5 — plan-review findings A-8/A-9.)**
   S1's step-8 checkpoint is written through `backlogit_create_checkpoint` (MCP) **or**
   `backlogit checkpoint create --state-dump` (CLI) — **both are valid**; the payload MUST
   carry, **top-level**, all of `schema_version: 1`, **`agent: ship`**, **`session_id`**,
   `phase`, and `resume_hint`, with domain data nested under `context`.
   **`session_id` is REQUIRED top-level** by the installed Checkpoint Payload Contract; rev 4
   omitted it, so a plan-following agent would have written a **non-conforming** checkpoint
   that step 2's own anomaly gate then rejects. S1 **records the emitted filename**
   (`checkpoint-YYYYMMDD-HHMMSS.json`) and the `session_id` in its handoff. **S2 carries both
   as the expected identity** and matches them **exactly**.
1. **Enumerate** via `backlogit_list_checkpoints`, passing `consumer_id: "ship"` and **no**
   `status`/`agent` filter — a quarantined or schema-invalid record must not be silently
   excluded by the query itself.

   > **Rev 7 — `consumer_id` is ACCESS IDENTITY, NOT A FILTER (finding F-8, normative).**
   > `consumer_id: "ship"` identifies **the caller** to the MCP surface; it does **not**
   > restrict which records are returned, and it is **not** the `agent` field. The
   > enumeration is genuinely unfiltered — verified this session against the live tool, where
   > an unfiltered list returned **all 15** records including **14 `stage`-owned** ones.
   > **Do not substitute the CLI's `--agent` flag for it.** `--agent ship` *is* a real filter
   > and **would drop every `stage` record**, defeating the unfiltered-enumeration
   > requirement and silently disabling **AC-29**'s leftover warnings. The identity parameter
   > and the forbidden filter are different things that rev 6's wording ran together.

2. **Run the anomaly gate FIRST** — evaluated over the **full enumeration**, **before** the
   candidate-count check, and **fail closed** to operator handoff on any anomaly.

   > **Rev 7 — the anomaly signal is the TOP-LEVEL COUNTERS, not a per-summary flag
   > (finding F-9, normative).** Rev 6 spoke of a *"quarantine flag"* on each summary. **No
   > such per-record field exists.** The installed `backlogit_list_checkpoints` response
   > exposes the anomaly signal as **top-level counters** alongside the `checkpoints` array:
   >
   > | Field | Meaning | Gate |
   > |---|---|---|
   > | `needs_quarantine` | records detected as requiring quarantine | **any value > 0 ⇒ FAIL CLOSED** |
   > | `quarantined` | records already quarantined | **any value > 0 ⇒ FAIL CLOSED** |
   > | `total` | total records enumerated | must equal `checkpoints.length` — **a mismatch means records were dropped ⇒ FAIL CLOSED** |
   >
   > **In addition**, each enumerated summary is inspected for a **missing or malformed
   > required field** (notably an empty `agent` or `status`, the shape a parse-failed record
   > presents as). **Any such record ⇒ FAIL CLOSED**, regardless of the counters.
   > *(Measured this session: `needs_quarantine: 0`, `quarantined: 0`, `total: 15`,
   > `checkpoints.length = 15`, no empty required field — gate clean.)*
2a. **S2-SPECIFIC CANDIDATE RULE — FILTER FIRST, THEN COUNT THE MATCHES. (rev 5; ORDER
   CORRECTED in rev 6, finding B-8.)**
   Rev 5 fail-closed on **any** `>1` active `ship` candidate **before** applying the identity
   filter. But the installed protocol **deliberately leaves fail-closed handoffs `active`
   regardless of age**, so unrelated leftovers are **expected**, not exceptional — one
   leftover plus a perfectly valid S1 checkpoint would have **bricked S2**. The count is now
   taken over the **MATCH set**, never over the raw enumeration:

   | # | Operation | Rule |
   |---|---|---|
   | i | **Enumerate ALL** — pass `consumer_id: "ship"` (**access identity, not a filter**; never the CLI `--agent` flag), **no** `status`/`agent` filter | a quarantined or schema-invalid record must not be hidden by the query |
   | ii | **Anomaly gate over the FULL enumeration** (step 2 above) — top-level `needs_quarantine`/`quarantined` > 0, a `total` vs `checkpoints.length` mismatch, or any missing/malformed required field in **any** record | ⇒ **FAIL CLOSED**, before any filtering |
   | iii | **FILTER to MATCHES** — exact expected checkpoint filename/reference **AND** `agent: ship` **AND** the exact `session_id` S1 recorded | all three conjunctive |
   | iv | **COUNT the MATCHES — require exactly ONE** | **0 MATCHES ⇒ FAIL CLOSED** (S2 presupposes S1's checkpoint; the installed "zero candidates is normal startup" continuation **does not apply here**). **>1 MATCH ⇒ FAIL CLOSED** (ambiguity is never resolved by picking) |
   | v | **Non-matching active records are a WARNING, not a failure** | they are **surfaced in the handoff report** and otherwise ignored. **They never strand a valid S1 handoff**, and S2 **never** resolves, prunes or touches them — least of all a `stage`-owned record (P-001) |

   **Identity mismatch is still fail-closed.** If the single MATCH's filename/reference,
   `agent`, or `session_id` disagrees with S1's recorded values at any later validation
   point, S2 **FAILS CLOSED**. Malformed or unreadable records **FAIL CLOSED** at step ii.

   > **`session_id` availability (rev 6, finding B-8 second half).** The installed
   > `backlogit_list_checkpoints` summary **does** expose `session_id` — verified this session
   > against the live tool, whose summaries carry `filename`, `agent`, `session_id`, `phase`,
   > `status`, `created_at`, `shipment_id`, `feature_id` and `resume_hint`. The filter at
   > step iii is therefore executable **at enumeration time**. S2 nonetheless
   > **re-confirms `agent` and `session_id` after `backlogit_get_checkpoint`**, because the
   > CheckpointV1 record — not the summary — is the authoritative source.

   Exactly **one MATCH**, satisfying filename, `agent: ship` **and** `session_id`, may
   proceed to step 3. Anything else halts to the operator.
3. **Require explicit owner selection** — never auto-pick, **even when exactly one
   candidate is returned** and even when it matches. The operator selects one checkpoint by
   filename; ambiguity fails closed.
4. **Validate ownership** — CheckpointV1 `agent` MUST be exactly `ship`. A `stage`-owned
   checkpoint is never selectable here (P-001 role separation).
5. **Obtain operator confirmation before restore** — present `resume_hint` and recorded
   state. There is no automatic resume: schema V1 carries no heartbeat or lease, so elapsed
   time can never prove S1 dead.
6. **RESTORE** — load the selected checkpoint's `state_dump` via
   `backlogit_get_checkpoint`. **(rev 4)**
7. **PRUNE/GATE — before resuming, never after. (rev 4, mandatory here.)** The installed
   owner sequence is **`restore → prune/gate → resume`**, explicitly *"never restore →
   resume → prune"*. `agent-engram` **is installed and active** in this workspace, so this
   step is **not** the supported no-prune no-op available to backlogit-only installs — it
   is a required gate. S2 performs a bounded **read-select-summarize** over the restored
   `state_dump` and bound engram state, dropping **only** superseded action-observation
   history already synthesized into recorded state.
   * **Engram reachability is fail-closed. (rev 4)** If the bound engram substrate is
     **unreachable** at restore, S2 **FAILS CLOSED to operator handoff — NO prune and NO
     resume.** A bounded file-based prune degradation is explicitly **not** permitted. This
     is the installed degraded-fallback rule and is distinct from the static
     `agent-engram`-not-installed case, which does not apply here.
   * **Never-prune allowlist. (rev 4)** These three classes survive pruning **intact**,
     regardless of age or verbosity: **(a)** the active-shipment / active-task cursor — the
     single-active resumption point, here `021-S`; **(b)** the **unresolved-checkpoint
     pointer itself** — the checkpoint being resumed; **(c)** **recorded gate verdicts** —
     quality gates, review verdicts, CI/PR gate outcomes. Pruning any of these is a
     protocol violation, not an optimization.
8. **Resume the same cursor** — and only **after** step 7 completes. The **same** single
   active shipment `021-S`. No parallel resume, no new worktree (P-001/P-016).
9. **Resolve only after a confirmed successful resume** — `backlogit_resolve_checkpoint`
   for that **one** selected checkpoint only. No bulk sweep, never a `stage`-owned record.
10. **Fail closed with no fresh-start fallback** — an invalid, torn, or unreadable
    checkpoint halts to the operator. S2 MUST NOT discard it and begin fresh work.
11. **No `cleanup_checkpoints` against an unresolved record. (rev 4)** A fail-closed
    operator handoff performs no resolve and deliberately leaves the checkpoint `active`.
    That is **neither** resolution **nor** an explicit archival decision, so the record
    stays **excluded from `cleanup_checkpoints` eligibility regardless of age** until it is
    either resolved after a confirmed resume or explicitly archived/abandoned by a named
    operator decision.

### 9.2 What this proves

* **Checkpoint/resume across a session boundary** — steps 8→9, executed, not asserted.
* **No authority gained mid-session** — S1 never uses feature-completion authority;
  it is used first by S2, which did not perform the merge. This is the §6 property
  demonstrated by execution.
* **A closes by an existing path.** At step 15, P-015 still has exactly one exception.
  A introduces no verdict and closes by the pre-existing one.

### 9.3 What A releases — `017-S` (rev 3)

A's full post-merge closure (step 15) and its P-020 operational closure (step 16) are the
release condition for **`017-S`**, which is the **only** dependent. `017-S.dependencies`
is exactly `[021-S]`.

`017-S` then closes **by the same existing path A used** — no new verdict, no new gate:

| Step | Action |
|---|---|
| 1 | `017-S` becomes eligible once `021-S` is archived shipped **and** A's post-merge closure is complete (P-020 compaction status `done`/`degraded`, not `pending`) |
| 2 | Ship claims `017-S`. Manifest is the **13-member fully-covered root**: root `018-F` + `018.008-T` (live) + **11 legacy descendants that DECLARE `status: archived`** |
| 3 | Step 4.5 moves `018.008-T -> done` *(after which registry routing relocates it to `archive/` — it remains a **completed**, `matched` member, **not** pre-archived; §4.4 rev 6)* |
| 4 | **Step 6.1(a1)** — §5.1 step **S4** (`n == 1`, status `active`), **running BEFORE pre-mode**: `018-F` is a present feature member; conditions hold; `018-F active -> done`. The 11 descendants **declaring `status: archived`** are exempt from condition 1 **only** under the §5.0 enumerated-ID disposition; `018.008-T`, declaring `status: done`, **satisfies condition 1 normally** |
| 5 | **Step 6.1(a)** pre-mode `expected_status: done`: `018-F` declares `done` (**`matched`**), `018.008-T` declares `done` (**`matched`**), 11 legacy members declare `status: archived` (**`pre-archived`**, accepted without an `expected_status` check) → **`PROCEED`** |
| 6 | The **authoritative** classification returns **`CASCADE`** — the **existing** exception. **Any other verdict is a topology discrepancy ⇒ HALT to Stage; no `SAFE_CLOSE` fallback** (AC-28) |

**Why the rev-18 `SAFE_CLOSE` exit-9 blocker is not reached.** Rev 18 recorded that a
`SAFE_CLOSE`-classified shipment halts at skill step 8 because
`backlogit move <shipment_id> --status shipped` is refused (exit 9, Probe 18), and argued
no manifest shape could avoid it *"because the only shape that reaches the permitted
cascade path is one containing `018-F`, which fails pre-mode first."* **Step 6.1(a1) is
exactly what repeals that clause** — with a1 live, a manifest containing `018-F` no longer
fails pre-mode first. `017-S` classifies `CASCADE`, never `SAFE_CLOSE`, so step 8 is never
reached. The underlying tool/contract conflict is **not fixed** here; it is simply **off
this route**, and remains a recorded follow-up.

> **Verification scope of this claim — stated honestly (rev 3, internal review P2;
> ESCALATED TO A GATE in rev 4; **FALLBACK OVERCLAIM DELETED in rev 6, finding B-6**).**
> The `CASCADE` result above is established at the **classification** level: the manifest
> shape provably satisfies P-015's fully-covered-root preconditions, and pre-mode provably
> accepts both `matched` and `pre-archived` members. It is **not** established at the
> **engine-execution** level *for this exact member shape* by anything other than Probe 25's
> fixture measurement — which is why that measurement is a pre-claim gate (§9.4).
>
> The path is **fail-closed** at two independent gates — the skill's cascade step 2 HALTs
> on non-empty `returned_ids`, and step 3's two-set gate HALTs on any required/returned
> mismatch — so a mis-derivation cannot silently corrupt the backlog.
>
> **THE "FALLS BACK TO `SAFE_CLOSE` AT NO DATA COST" CLAIM IS WITHDRAWN IN FULL.** Rev 3
> asserted that a HALT on this route *"falls back to `SAFE_CLOSE`, re-exposing the exit-9
> blocker, returning `017-S` to Stage at no data cost."* **All three parts were wrong**, and
> the sentence is deleted rather than qualified:
>
> 1. **P-015 forbids CASCADE→SAFE_CLOSE substitution once `CASCADE` is selected.** The
>    skill's generic *"any classifier error … falls back to safe-close"* rule governs
>    **classification-time** failure, **not** a HALT raised by the cascade sub-procedure's own
>    post-mutation gates. Re-routing a halted cascade into safe-close would run a second close
>    path over a partially-mutated backlog.
> 2. **A HALT after the cascade has begun is NOT "no data cost."** The cascade's verification
>    is **detect-after-mutate** (§9.4b): by the time step 2 or step 3 halts, the mutation has
>    already occurred. The cost is a backlog requiring operator restore — §10.3's
>    `git restore -- .backlogit/queue/ .backlogit/archive/` — not zero.
> 3. **There is no automatic fallback of any kind on this route.** See the rev-6 verdict
>    contract immediately below.

> **Rev 6 — THE AUTHORITATIVE VERDICT CONTRACT FOR A AND `017-S` (finding B-6, normative).**
>
> **Ordering is fixed and one-directional** (§5.1 rev 6): **a1 runs first** (§9 step 13),
> then the **authoritative installed `shipment-reconcile` / P-015 mode selection runs after
> it** (§9 steps 14–15). a1 neither performs nor anticipates that selection.
>
> **Both A's and `017-S`'s manifests are exact root-inclusive fully-covered roots, so the
> authoritative classification is EXPECTED to return `CASCADE`.**
>
> **Any other verdict is a TOPOLOGY DISCREPANCY and HALTs. There is NO `SAFE_CLOSE` fallback
> for these two shipments.** If the classifier returns `SAFE_CLOSE`, that does not mean
> "close a different way" — it means **the live manifest is not the shape this plan asserts**,
> which is a planning-state error only **Stage** can reconcile. Ship HALTs, mutates nothing,
> and returns the shipment to Stage. `HALT` likewise halts. An unknown token halts (§4.1).
>
> **Invocation of the cascade branch is CONDITIONAL on the authoritative `CASCADE` verdict.**
> Ship does not "take the cascade path because the plan says so" — it takes it **if and only
> if** the classification it did not compute names `CASCADE`. The branch itself is the
> **already-existing** one; **A adds no cascade path, no verdict, and no new close
> procedure.**
>
> **What A genuinely changes, stated without softening: REACHABILITY.** Before A, `017-S`'s
> root-inclusive manifest failed pre-mode first, so the destructive cascade branch was
> **unreachable** for it. After A, a1 completes the covering feature, pre-mode passes, and the
> branch becomes **reachable**. A therefore **inherits that branch's detect-after-mutate
> residual for a route that previously could not reach it** (§9.4b). That is a real, disclosed
> consequence of A — **not** a defect A introduces into the cascade, and **not** something
> Probe 25 or any part of A repairs.

**Why the rev-18 `SAFE_CLOSE` exit-9 blocker is not reached.** Rev 18 recorded that a
>
> **Rev 4 — the rev-3 disposition is WITHDRAWN.** Rev 3 recorded this as *"not a current
> blocker … deliberately not charged to A"*, to be run *"before `017-S` is routed, **not**
> before A is."* That disposition is unsound **as a matter of sequencing**: A's entire
> justification for existing is that it makes `017-S`'s shape closable (§2.1, §9.3). If
> the root-included cascade does **not** behave as classified, A has shipped an
> irreversible contract change — a permanent Role Boundary authority grant — in service of
> a route that does not work, and the discovery happens only *after* A is closed and
> `017-S` is claimed. The measurement is **cheap, disposable, and available now**; the
> rev-3 ordering deferred it past the one commit that makes it expensive to act on.
> It is therefore promoted to a **non-bypassable gate on A's own close** (§9.4, **AC-20**).

> **Rev 4 — the `_ship.agent.md` pre-mode summaries are now IN A's inventory.**
> Rev 3 carried a block here scoping the L776–777 pre-mode summary *"deliberately OUT OF
> A's INVENTORY"* on budget and not-load-bearing grounds, and instructed that *"a reviewer
> MUST NOT fail A for its absence."* **That block is withdrawn in full, and so is that
> instruction.** Both stale summaries — L255–256 **and** L776–777 — are now clause sites
> **C7** and **C8**; see **§4.0** for why both rev-3 grounds fail, and **§4.4** for the
> reconciled wording. A reviewer **SHOULD** fail A if either site ships unreconciled; Go
> test rows **14–15** enforce it mechanically.

### 9.4 Root-included cascade fixture — STAGE PRE-CLAIM EVIDENCE (rev 5, hybrid)

> **Rev 5 — this section is RE-SEATED. Rev 4's placement is WITHDRAWN.**
> Rev 4 made this an S2-run, no-override gate at §9 step 12a — *after* A's Role Boundary
> change had already merged to `main` at step 5. The attempt-1 review found two blocking
> defects in that placement, and both are structural rather than editorial:
>
> * **A-1 (P0)** — a **no-override** gate placed **after an irreversible merge** creates an
>   **unrecoverable stall**: on FAIL the contract change is already deployed to every future
>   Ship session while `021-S` is stuck `active`/in-flight, and no in-role actor can clear it
>   (Stage may not close shipments; Ship is halted by the gate). It also contradicted this
>   plan's own §10.2 `git revert` rollback — precisely the override §9.4 forbade.
> * **A-2 (P1)** — it assigned **authoring and committing** `docs/plans/evidence/probe25-*`
>   to **Ship**, which Ship's Role Boundary **forbids** under **P-010**. Executed literally,
>   S2 would record a P-010 violation and halt, and **A could never close**.
>
> **The hybrid resolves both.** **Stage authors, executes and commits the evidence BEFORE
> `021-S` is claimed. Ship re-verifies it READ-ONLY.** Nothing has merged when the
> measurement happens, so a FAIL costs a re-plan and nothing else; and Ship never touches a
> planning artifact, so P-010 is never engaged.

**Ownership — unambiguous.**

| Actor | Does | Never does |
|---|---|---|
| **Stage** (§9 step **0a**, pre-claim) | authors, executes and **commits** `probe25-*.{ps1,txt}` | — |
| **Ship S2** (§9 step **12a**, post-merge) | **read-only re-verification** of the committed evidence (§9.4a) | author, modify, re-run or regenerate planning evidence |

**STATUS: EXECUTED AND COMMITTED.** Stage ran Probe 25 at `4fd21c5`, **before** `021-S` is
claimable. `PROBE25_RESULT=PASS`, `FAILED_CRITERIA=0`, **all eight criteria PASS across nine
emitted tokens** — criterion 5 emits `CRITERION_5A_NO_UNEXPECTED_ARCHIVE` and
`CRITERION_5B_NO_REQUIRED_LEFT_UNARCHIVED` as two independent, never-merged set equations, so
8 criteria ↔ 9 tokens is **correct, not a discrepancy** (rev 6, finding C-4 second half).

> **Rev 6 — the shipment record IS verified, by a separate criterion (finding C-4).** The
> transcript's `REQUIRED_IDS=001-F,001.004-T` omits the shipment record, and P-015 makes that
> record an **unconditional** required member — so on its own, criterion **5B** could pass
> while the engine silently failed to archive it. **It cannot pass unnoticed here, because
> criterion 7 checks the shipment record directly and independently**:
> `SHIPMENT_LOC=archive`, `SHIPMENT_STATUS=archived`, `SHIPMENT_ARCHIVED_STATUS=shipped`,
> `CRITERION_7_SHIPMENT_ARCHIVED_SHIPPED=PASS`. The engine's own `archived_ids` also lists it
> (`ARCHIVED_IDS_TRANSITION_LOG_ENGINE=001-F,001-S,001.004-T`), and
> `ALLOWED_IDS_INCL_SHIPMENT_COUNT=14` shows it inside the allowed set for equation A.
> **Coverage is therefore complete**, but it is delivered by criterion 7 rather than by 5B.
> **Recorded follow-up (not charged to A, and not a re-run):** a future revision of the
> fixture should add the shipment record to `required_ids` so equation B is
> self-sufficient. The current evidence is **not** invalidated — nothing it asserts is
> false, and no gap is left unchecked.

**Gate definition.**

| Property | Value |
|---|---|
| **Who runs it** | **Stage**, at §9 step **0a** — pre-claim, pre-merge, pre-implementation |
| **When** | **Before `021-S` is claimed.** A plan/PR readiness gate: `021-S` is **not claimable** until this evidence is committed and PASSing |
| **Fixture shape** | The **exact** `017-S` shape: a **13-member** manifest = **root feature included** + 12 descendants, of which **11 are genuinely archived** and 1 is a live task. Not an approximation, not the Probe-15 root-excluded inverse |
| **Workspace** | A **disposable, isolated, live-config-seeded** backlogit workspace — **never** the live one. Same pattern as Probes 18/20/22/24 |
| **Engine binding** | **Portable identity check.** The engine is resolved **at runtime** from the **registered command** (`Get-Command backlogit`); its SHA-256 is compared to the **committed digest** `1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98`. **No runtime requirement hardcodes an absolute path.** The evidence **records** the observed path as telemetry. **Mismatch ⇒ HALT** |
| **Evidence location** | `docs/plans/evidence/2026-09-12-task-only-shipment-finalization/probe25-root-included-cascade-fixture.ps1` **and** `…/probe25-root-included-cascade-fixture.txt` — **committed**, alongside probes 14–24 |
| **Authorization** | **CLI-only**, matching probes 18–24. No MCP equivalence is claimed |

> **Rev 5 — portability (plan-review finding A-3), and the Hardening 5 reconciliation.**
> Rev 4 pinned the gate to the absolute path `C:\Tools\backlogit.exe`, so any second
> operator, CI runner or toolchain move produced `DIGEST_GATE=FAIL` → HALT → (under rev 4's
> no-override rule) a bricked A. **That path binding is removed from the contract.** What
> the contract requires is an **identity** assertion, not a **location** assertion:
>
> * **resolve** `backlogit` at runtime via the registered command / `Get-Command`;
> * **compare** the resolved executable's SHA-256 to the committed digest;
> * **record** the observed path (`ENGINE_PATH_OBSERVED`) as evidence telemetry only;
> * on **mismatch**, **HALT** — the measurement cannot be attributed to a known engine build.
>
> A HALT here is now cheap and recoverable: it happens **pre-claim**, so it returns the work
> to Stage with nothing merged and nothing in flight. **Hardening 5 is reconciled**
> accordingly below: A's *execution* path still introduces no digest-bound engine call; A's
> *verification* path is digest-gated, and that gate now sits **before** the irreversible
> step rather than after it.

#### 9.4.1 The MEASURED live shape the fixture must reproduce (rev 4 — verified at `d19b954`)

Stage traced all 13 live `017-S` members this session. The fixture is only faithful if it
reproduces **this** shape; a fixture that archives its members *some other way* measures a
different case and its PASS is worthless.

| Member class | Count | Location | `artifact_type` | `status` | `archived_status` | `parent_id` |
|---|---|---|---|---|---|---|
| Root feature `018-F` | 1 | `queue/` | `feature` | `queued` | — | **none (root)** |
| Live task `018.008-T` | 1 | `queue/` | `task` | `queued` | — | `018-F` |
| Legacy tasks `018.001-T`, `018.002-T`, `018.003-T` | 3 | `archive/` | `task` | **`archived`** | **`queued`** | `018-F` |
| Legacy subtasks (3 + 3 + 2) | 8 | `archive/` | `subtask` | **`archived`** | **`queued`** | their `018.00N-T` |

Verified: **13/13 present**, **0 missing**, **0 duplicates** (no id in both `queue/` and
`archive/`), `018-F` declares **no `parent_id`**, and every one of the 12 descendants resolves
to `018-F` by the `parent_id` chain at depth 1 or 2. The manifest contains **nothing extra**.
P-015 preconditions 1 and 4 therefore hold on the live records.

> **The load-bearing detail: `archived_status: queued`, NOT `done`.** All 11 pre-archived
> members are `status: archived` with `archived_status: **queued**`. A fixture built with
> `archived_status: done` would be a **different and easier** case, because it would satisfy
> the skill's `archived-completed(done)` role, which these records **do not**.
>
> **Why this does not block the route** (checked, not assumed): under the skill's per-task
> **ROLE** table these 11 records classify as the **`any-other-archived-status` anomaly**
> (*"an archive record whose status is NEITHER `done` NOR `archived`-with-`archived_status:
> done`"*), and that table's anomaly list says *"REPORT and HALT"*. **But that table is scoped
> to `mode: detect-mixed-role` only**, which the skill states is *"used **ONLY to DESCRIBE**
> the inconsistency in a report — **NEVER** to gate a mutation"* (013-DL Addendum G, Copilot
> PR #304 finding 1). **Pre-mode** uses the separate four-value classification
> (`matched`/`pre-archived`/`missing`/`status-mismatch`), under which each of these 11 members
> is **`pre-archived` BECAUSE IT DECLARES `status: archived`** — valid, accepted without an
> `expected_status` check — so `017-S` reaches
> `PROCEED`. *(Rev 6, B-1/B-2: the qualifying fact is the **declared status**, not the archive
> location. §4.4's rev-6 table governs. The distinction is load-bearing here: `018.008-T` will
> declare `status: done` and is therefore **`matched`, not `pre-archived`** — exactly as
> `022.001-T` is on A's own route.)* The two classifications are explicitly different scans;
> conflating them would produce a false halt.
>
> **Recorded residual**: if any future gate were to adopt the role table as a *gating* check,
> all 11 members would halt `017-S`. That is not the case today and is not charged to A, but
> the fixture MUST carry `archived_status: queued` so this exact condition is exercised rather
> than accidentally avoided.

**PASS criteria — all of them, conjunctive. MEASURED RESULT: 8/8 PASS.**

| # | Criterion | Measured |
|---|---|---|
| 1 | `DIGEST_GATE=PASS` against the committed digest, engine resolved via registered command | **PASS** |
| 2 | pre-mode with `expected_status: done` returns **`PROCEED`**, with the 11 archived members classified **`pre-archived`**, the 2 completed members **`matched`**, and **no** `status-mismatch`, `missing` or torn/duplicate record | **PASS** |
| 3 | classification returns **`CASCADE`** (not `SAFE_CLOSE`, not `HALT`) | **PASS** |
| 4 | the cascade close returns **`returned_ids: []`** — empty | **PASS** |
| 5a | **`archived_ids − allowed_ids` is EMPTY** — nothing unexpected was archived | **PASS** |
| 5b | **`required_ids − archived_ids` is EMPTY** — nothing required was left unarchived | **PASS** |
| 6 | every surviving record's **`parent_id` is preserved** | **PASS** |
| 7 | the shipment record reaches **`shipped`** and archives with `archived_status: shipped` | **PASS** |
| 8 | **nothing outside the manifest** is archived, moved or deleted | **PASS** |

Criteria **5a and 5b are two separately-labelled, independently-failing conditions**
(P-015 amendment 1.21.0), never a single reconciliation. `allowed_ids` is the manifest
**plus the shipment record itself** — the close archives its own record and the engine
reports it. `archived_ids` is a **transition log, not a manifest echo**: a genuinely
`status: archived` member has no transition to report and is correctly absent, which is
exactly why full-set equality is the wrong test and why the 1.19.0 equality claim was
withdrawn. The probe evaluates 5a over the **union** of the observed pre/post snapshot log
**and the engine's own reported `archived_ids`**, so a mutation visible to one but not the
other cannot pass vacuously.

**Measured highlights** (full transcript in the committed `.txt`):

```text
MEMBERS_QUEUE_LOCATED=0          MEMBERS_ARCHIVE_LOCATED=13
C7_C8_QUEUE_ONLY_SUMMARY_FALSIFIED=True
PREMODE_MATCHED=2  PREMODE_PRE_ARCHIVED=11  PREMODE_RESULT=PROCEED
CLASSIFICATION_VERDICT=CASCADE
archived_ids=[001.004-T, 001-F, 001-S]   returned_ids=[]
SET_EQ_A__archived_minus_allowed=EMPTY   SET_EQ_B__required_minus_archived=EMPTY
LIVE_BACKLOG_UNMUTATED=True              PROBE25_RESULT=PASS
```

> **An unplanned but load-bearing measurement: `MEMBERS_QUEUE_LOCATED=0`.** At the pre-mode
> gate, **zero of the 13 members are in `.backlogit/queue/`** — the live registry routes
> `done|accepted|rejected|archived` to `archive/`, so even the two *correctly completed*
> members leave the queue. This **independently falsifies the queue-only pre-mode summaries**
> at C7 and C8 by measurement rather than by argument, and it strengthens §4.0: the stale
> sentences are wrong not only for the 11 legacy members but for **every** member of a
> completed manifest.

**Failure semantics — fail closed, and RECOVERABLE (rev 5).**

* **Pre-claim (Stage, the normal case).** A FAIL, or a digest mismatch, **HALTs before
  `021-S` is claimed**. Nothing has merged, nothing is in flight, and the cost is a return
  to Stage for re-planning. `021-S` stays **unclaimed**; `017-S → 021-S` stays unsatisfied.
  **This is the gate that actually protects the route**, and it has already run: **PASS**.
* **Evidence missing or not committed ⇒ `021-S` is NOT CLAIMABLE.** This is a plan/PR
  readiness gate, not a closure-time gate.
* **Post-merge (Ship S2 re-verification) — see §9.4a for the recovery path.** Rev 4's
  *"no operator override, no `--force`, no proceed-with-residual"* clause is **WITHDRAWN**:
  as written it contradicted §10.2's `git revert` rollback and left no in-role actor able
  to clear a stalled `021-S`.

#### 9.4a Ship S2 re-verification — READ-ONLY (rev 5, normative)

At §9 step **12a**, the fresh S2 session performs a **read-only** re-verification of the
**already-committed** evidence. **S2 authors nothing, modifies nothing, and re-runs nothing.**

**What S2 checks — all six, conjunctive (rev 6: field names disambiguated per B-4; check 6
added per C-8):**

1. **Exact topology.** The evidence records `TOPOLOGY_MEMBERS=13`, `ROOT_INCLUDED=True`,
   `PRE_ARCHIVED=11`, `ARCHIVED_STATUS_FLAVOUR=queued`. **An exact-match check** — any other
   topology means the measurement is not about `017-S`'s shape.
2. **PASS fields.** `PROBE25_RESULT=PASS` and `FAILED_CRITERIA=0`, and each of the **nine
   STEP 11 aggregate criterion tokens** reads `PASS`. **The nine names are PINNED here
   verbatim (rev 7, finding F-7)** — they are the `C*` aggregate names emitted in the
   transcript's `### STEP 11: AGGREGATE VERDICT` block, **not** the `CRITERION_n` tokens
   emitted inline earlier in the run:

   | # | Pinned STEP 11 aggregate token |
   |---|---|
   | 1 | `C1_DIGEST_GATE` |
   | 2 | `C2_PREMODE_PROCEED_11_PRE_ARCHIVED` |
   | 3 | `C3_CLASSIFICATION_CASCADE` |
   | 4 | `C4_RETURNED_IDS_EMPTY` |
   | 5 | `C5A_ARCHIVED_MINUS_ALLOWED_EMPTY` |
   | 6 | `C5B_REQUIRED_MINUS_ARCHIVED_EMPTY` |
   | 7 | `C6_PARENT_ID_PRESERVED` |
   | 8 | `C7_SHIPMENT_ARCHIVED_SHIPPED` |
   | 9 | `C8_NOTHING_OUTSIDE_MANIFEST_TOUCHED` |

   > **Why this pin exists.** Rev 6 said only *"the **nine** criterion tokens (eight
   > criteria; criterion 5 emits `5A` and `5B`)"*. The transcript body emits inline
   > `CRITERION_2…8` tokens — **eight**, not nine — so an S2 agent grepping for nine
   > `CRITERION_n` tokens finds eight and **HALTs on a fabricated failure**. The nine-token
   > set exists **only** in the STEP 11 aggregate, under the `C*` names above. S2 checks
   > **these exact nine names**.

   > **What criteria 2 and 3 actually measure (rev 7, finding F-4 — disclosed, not
   > softened).** The eight criteria are **not** uniformly engine-observed:
   >
   > | Criterion | Source | What it establishes |
   > |---|---|---|
   > | **C2** (`PREMODE_PROCEED_11_PRE_ARCHIVED`) | **PowerShell-derived** (transcript STEP 5) | the fixture's own re-derivation of the pre-mode outcome for the shape |
   > | **C3** (`CLASSIFICATION_CASCADE`) | **PowerShell-derived** (transcript STEP 6) | the fixture's own re-derivation of the P-015 close-path selection |
   > | **C1, C4, C5A, C5B, C6, C7, C8** | **engine-observed** | measured against the real `backlogit` engine |
   >
   > **The actual engine cascade invocation begins at the `CASCADE` call** — transcript
   > `### STEP 8: AUTHORIZED CALL`. Criteria **4 through 8** observe that call's real output
   > and the real post-call backlog state; **C1** is a real digest of the resolved
   > executable. **C2 and C3 are simulations of the decision that precedes it**, and the
   > plan does not claim otherwise. This matters because the **central empirical claim of
   > Probe 25 — that the root-included cascade archives exactly the manifest and preserves
   > every `parent_id` — rests on the engine-observed criteria**, which are unaffected by
   > the two simulated ones.
3. **Evidence commit ancestry — on `evidence_commit_sha`.** Both evidence files are
   **present and tracked**, and `evidence_commit_sha` — **the LAST commit touching either
   file**, per Hardening 9a — is an **ancestor of the merge commit** S2 verified at §9 step 6
   (`git merge-base --is-ancestor {evidence_commit_sha} {merge_sha}`). Because the field is
   the *last* commit and not the *introducing* one, **post-merge replacement of the evidence
   fails this check** rather than slipping past it.
4. **Current resolved engine digest — `backlogit_executable_sha256`.** S2 resolves
   `backlogit` **the same portable way** (registered command / `Get-Command`) and compares
   its **current** SHA-256 to the digest recorded in the evidence. This detects toolchain
   drift **between** measurement and close. **No absolute path is required by this check.**
5. **No stale or missing record — `fixture_script_sha256` / `fixture_result_sha256`.**
   Neither evidence file is absent, empty, or modified relative to its committed blob
   (`git diff --quiet HEAD -- {paths}`), and each file's current SHA-256 equals the value
   recorded for it at `evidence_commit_sha`. **The two files are hashed separately** so a
   swap or a partial edit cannot be masked by a single combined digest.
6. **Internal consistency of the transcript (rev 6, finding C-8).** A **read-only**
   arithmetic check on the evidence itself: `PREMODE_MATCHED + PREMODE_PRE_ARCHIVED +
   PREMODE_MISSING + PREMODE_STATUS_MISMATCH + PREMODE_DUPLICATE_TORN == MANIFEST_COUNT`,
   `PRE_ARCHIVED_COUNT == PREMODE_PRE_ARCHIVED`, and `FAILED_CRITERIA == 0` agrees with the
   per-criterion tokens. Checks 1–5 verify **provenance** only and would accept an internally
   contradictory transcript; this one catches the B-2 class of defect at re-verification
   time. **Current values satisfy it**: `2 + 11 + 0 + 0 + 0 == 13`, `11 == 11`, `0` failures.

**S2's role boundary is preserved absolutely.** Every one of the six is a **read**:
`git log`, `git merge-base`, `git diff --quiet`, `Get-FileHash`, and text inspection. S2
never writes to `docs/plans/`. This is what closes **A-2**.

**If S2 re-verification FAILS — the recovery path (closes A-1):**

1. **`021-S` remains `active`.** It is **not** marked `shipped`, **not** archived, and the
   a1 mutation is **not** performed. `017-S → 021-S` stays unsatisfied.
2. **S2 HALTs and returns to the operator**, recording **which named field or check** failed
   (`evidence_commit_sha`, `fixture_script_sha256`, `fixture_result_sha256`,
   `backlogit_executable_sha256`, topology, PASS fields, or consistency).

> **Rev 7 — THE RECOVERY PATH DEPENDS ON *WHICH CLASS* OF CHECK FAILED (plan-review finding
> E-8, normative; supersedes rev 6's flat three-option list).** Rev 6 offered "re-verify and
> continue" (option 1) and simultaneously asserted that **only** `git revert` clears a
> post-merge mismatch (option 3 / AC-20a). Both cannot be true of the same failure. They are
> true of **different** failures, and rev 7 separates them:
>
> | Failure class | Checks | Is the committed evidence still valid? | Recovery |
> |---|---|---|---|
> | **TRANSIENT — toolchain/executable drift** | **4** (`backlogit_executable_sha256`) | **YES** — the evidence is untouched; only the *environment* moved | **Correct the drift and RE-VERIFY READ-ONLY, in place.** A clean re-verification **clears the failure** and A proceeds. No revert, no re-plan, no new commit |
> | **COMMITTED-EVIDENCE / HASH / ANCESTRY** | **1, 2, 3, 5, 6** | **NO** — the evidence itself is absent, altered, internally inconsistent, or no longer an ancestor | **CANNOT be cleared in place.** Requires **`git revert` per §10.2** (after the `017-S` hold) **and a re-plan** |
>
> **Why the second class cannot be repaired forward.** Stage re-running Probe 25 and
> re-committing after the merge moves `evidence_commit_sha` to a **descendant** of the merge
> SHA, so **check 3 still fails**. That option therefore restores a valid pre-claim posture
> for a **re-planned** A or a future shipment — it is **not** a way to green an in-flight
> post-merge A (**finding C-7**, preserved).
>
> **The distinction is evidentiary, not procedural.** Check 4 compares the *current
> environment* against a recorded measurement; a mismatch says the **environment** changed,
> and re-resolving a correct engine restores the very state the evidence describes. Checks
> 1/2/3/5/6 interrogate the **committed artifact and its position in history**; a mismatch
> says the **evidence** is wrong or misplaced, and nothing S2 may do read-only can make a
> wrong artifact right.

1. **Recovery options, by class:**
   * **Transient drift (check 4 only)** — **re-verify after resolving toolchain drift**,
     read-only and in place. This is the common benign cause, and it genuinely clears.
   * **Committed-evidence/hash/ancestry (checks 1, 2, 3, 5, 6)** — **revert A** per §10.2,
     after holding `017-S` per §10.2's obligation, then **re-plan**. **This is the only
     option that clears this class on A itself.**
   * **Either class, forward-looking only** — direct **Stage** to re-run Probe 25 pre-claim
     and re-commit fresh evidence. Stage owns this artifact, so it is always available, but
     **it restores a valid posture only for a re-planned A or a future shipment**, never for
     an in-flight post-merge A.
2. **`021-S` is left claimable-and-active for a future Ship session**, not stranded: because
   no a1 mutation and no close occurred, a later S2 re-entry is a clean resume (§5.1 **S4**,
   or **S3** if a prior attempt had already completed the feature).

**No rollback fiction.** Nothing here claims an automatic restore, an automatic revert, or a
"proceed with a recorded residual" path. A post-merge mismatch **HALTs active A** and hands
the operator the **class-appropriate** recovery above: a **transient toolchain drift** is
cleared by an in-place read-only re-verification, and a **committed-evidence/hash/ancestry**
failure has exactly one true unwind — `git revert` — which is explicit, manual, and gated on
the `017-S` hold.

**The impossible contradiction is gone.** There is no longer a no-override clause sitting
against a rollback that requires an override, **and rev 7 removes the second contradiction —
"re-verify and continue" is no longer offered for failures it cannot actually clear
(finding E-8)**. No step assigns Ship a P-010-forbidden write. The gate is still
**non-bypassable** — S2 may not close A without a PASS — but a failure is now **recoverable
by a named actor, along the path that matches the failure** instead of deadlocked.

**What this gate does NOT claim.** It measures the **engine** on a faithful fixture; it is
not a proof about the live `017-S` records at the moment `017-S` is later claimed. The
live-record re-read obligations at `017-S`'s own gate are unchanged and still fail-closed —
the same relationship PO-1a/PO-1b have (§10.1).

#### 9.4b Inherited CASCADE residual — correctly classified (rev 5)

**Plan-review finding A-12** observed that the cascade's `returned_ids` / two-set /
`parent_id` checks run **after** the cascade has already mutated the backlog, so a
destructive mutation is *detected*, not *prevented*.

**This is a pre-existing P-015 residual, and A neither creates nor removes it.** Stating
that plainly matters, because the alternative — claiming Probe 25 prevents it — would be
an overclaim:

* **No new cascade path — but NEW REACHABILITY, and that distinction is stated in both
  directions (rev 6, finding B-6).** A introduces **no** verdict, **no** new close path and
  **no** change to the cascade procedure: the `CASCADE` invocation A rides is the identical
  call every fully-covered-root closure already makes today, and the detect-after-mutate
  ordering is a property of **P-015 and the engine**, inherited unchanged. **What A DOES do is
  make that destructive path REACHABLE for `017-S`, which previously could not reach it** —
  before a1, `017-S`'s root-inclusive manifest failed pre-mode first. Rev 5 asserted *"A
  neither creates nor removes it"* and *"not a new cascade path"* as if that settled the
  matter; both are true at the **implementation-definition** level and **incomplete** at the
  **reachability** level. **A inherits this residual for a new route. That is a genuine
  consequence of A**, and it is disclosed here rather than argued away.
* **What Probe 25 actually does** is act as a **pre-claim gate**: it measures, on an exact
  topology match, that this cascade shape behaves as classified. **Measured: `returned_ids`
  empty, both set equations EMPTY, all `parent_id`s preserved, nothing outside the manifest
  touched, and the live backlog unmutated.**
* **What Probe 25 does NOT do.** It does **not** prevent a future mutation, does **not**
  install a rollback, and does **not** make the detect-after-mutate window smaller. If the
  live topology at closure differs from the measured topology, the fixture's PASS does not
  transfer — which is exactly why §9.4a check 1 is an **exact topology match that HALTs on
  mismatch**.
* **Rollback is not claimed — and §10.3 is reconciled with this (rev 6, finding B-6).**
  Rev 4's implied restore-on-halt is **not** asserted here, and **no new rollback mechanism is
  introduced by A**. Rev 5 left this sitting against §10.3's `git restore` prescription
  without saying how they relate. They are reconciled explicitly:

  | HALT path | Mutation occurred? | Restore obligation |
  |---|---|---|
  | a1 **S0–S5** halts (§5.1) | **no** — a1 halts before any mutation | **none**; nothing to unwind |
  | pre-mode `RECONCILE_FAIL` (§9 step 14) | **no** — pre-mode is a check | **none** |
  | verdict is not `CASCADE` for A/`017-S` (§9.3 rev 6) | **no** — HALT precedes the close | **none**; return to **Stage** |
  | S2 §9.4a re-verification fails | **no** — 12a precedes a1 and the close | **none**; §9.4a's recovery paths |
  | **cascade step 2 — `returned_ids` NOT empty** | **YES — detect-after-mutate** | **MANDATORY verified restore** (below) |
  | **cascade step 3 — `archived_ids − allowed_ids` NOT empty** (archived something outside the manifest) | **YES — detect-after-mutate** | **MANDATORY verified restore** (below) |
  | **cascade step 3 — `required_ids − archived_ids` NOT empty** (failed to archive a required member) | **YES — detect-after-mutate** | **MANDATORY verified restore** (below) |
  | **cascade step 4 — any `parent_id` ALTERED or CLEARED vs the Step 0(b) snapshot** | **YES — detect-after-mutate** | **MANDATORY verified restore** (below) |

  > **Rev 7 — EVERY post-CASCADE check carries the SAME mandatory verified restore
  > (plan-review finding E-7, normative).** Rev 6's table collapsed the cascade's
  > post-mutation gates into one row (*"step 2 / step 3"*) and **omitted step 4's
  > `parent_id` preservation check entirely**, while §10.3 named only the
  > out-of-`allowed_ids` case. A valid-ID cascade that **cleared a `parent_id`** therefore
  > halted with **mutated state and no mandatory restore at all** — the reparenting/orphaning
  > failure mode, unguarded. **All four post-CASCADE failure modes are now enumerated
  > separately and every one carries the identical obligation**, mirrored verbatim in §10.3:
  >
  > 1. **`git restore -- .backlogit/queue/ .backlogit/archive/`** — restore from the
  >    pre-cascade state;
  > 2. **VERIFY the restore** — re-read the protected set and confirm declared `status`,
  >    `parent_id` and location match the **Step 0(b) pre-call snapshot** for every watched
  >    record. **An unverified restore is not a restore**;
  > 3. **HALT** and return to the operator. **No retry, no second close path, no
  >    "proceed with a recorded residual".**
  >
  > **If the restore itself cannot be verified, HALT with the backlog flagged as
  > operator-reconcile-required** — never continue, and never re-run the cascade.
  > **This is the inherited P-015 residual**, not something A adds; A adds only
  > **reachability** to it for `017-S`. Remediating the *ordering* (a pre-cascade restorable
  > snapshot and an unconditional restore) remains a **recorded follow-up against P-015 and
  > the skill** — **not charged to A**, whose 2-file surface touches neither.

  **Four rows carry a mandatory verified restore**, and all four are **inherited** P-015
  cascade gates — not anything A adds.

## 10. Risks, residuals, rollback

| ID | Risk | Disposition |
|---|---|---|
| **R-1** | `022-F` may be `queued`, not `active`, at the a1 gate, so §5 condition 5 fails closed and A cannot self-close | **(rev 2 — DISCHARGED by measurement.)** Probe 22 ARM AB measured `022-F`'s analogue at **`active`** after claim of a fully-covered-root manifest. PO-1a is closed; **PO-1b** remains as the at-the-gate re-read (§10.1). Fail-closed HALT is still correct if the live read disagrees |
| **R-2** | S1 might use the new authority anyway, defeating the proof | §6 is stated normatively **and** asserted by Go test row 6 |
| **R-3** | Delegation could be read as authorizing any verdict the skill invents | §4.1 binds Ship to verdicts the **classification** names; P-015 remains the authorization surface. Go test row 9 guards it |
| **R-4** | C2's correction touches the safe-close summary, which the deferred C plan would also edit in its own file | **No overlap, and now moot**: C2 is in `_ship.agent.md`; the skill's step 8 is C's file — disjoint surfaces. **Rev 3**: C is deferred and lands nothing, so there is no concurrent editor at all |
| **R-5** | Removing Ship's restatement loses reviewer-visible context | Accepted. The classification stays fully specified in P-015 and the skill; duplication is what caused the §2.2 drift |

### 10.1 PO-1 — covering-feature status at claim (**PO-1a discharged in rev 2**)

**PO-1a — before implementation begins (Stage / §9 step 0). STATUS: DISCHARGED.**
Rev 1 required Stage to determine what status shipment claim leaves a covering feature in,
**before** implementing, so a claim-shape mismatch could not be discovered only after A's
implementation merged. That measurement has now been executed:

> **Probe 22** (`probe22-parent-status-at-claim.ps1`, executed at `7bc0318`,
> `DIGEST_GATE=PASS`, live-config-seeded, CLI-only):
> **ARM AB** — fully-covered-root manifest `[001-F, 001.001-T]`, the exact shape of
> `021-S`:
> `ARM_AB_FEATURE_STATUS_PRE_CLAIM=queued`,
> **`ARM_AB_FEATURE_STATUS_AFTER_CLAIM=active`**,
> `ARM_AB_FEATURE_ACTIVE_AT_CLAIM=True`,
> `ARM_AB_FEATURE_DONE_MOVE_ACCEPTED=True`.

Claim leaves the covering feature **`active`**, which is exactly what §5 condition 5
requires, and the engine accepts the `active -> done` move a1 performs. **A may proceed to
implementation.** Neither of rev 1's contingency branches — amending the claim step, or
widening the authorized transition — is needed, and neither is authorized.

**PO-1b — at the gate (S2, before §9 step 13). RETAINED.** Read `022-F`'s live status with
**`backlogit get 022-F`** and record the `status` field.

> **Read by status, never by path (rev 2 correction).** An earlier draft said *"re-read
> `.backlogit/queue/022-F.md`"*. That is wrong and would have made the `done` branch below
> **unreachable**: the live `registry.yaml` routes `done|accepted|rejected|archived` to
> `archive/` and only `queued|active|blocked|review` to `queue/`, and `shipment-reconcile`
> independently classifies a live `status: done` record found in `queue/` as `conflicting`.
> A literal implementer reading the queue path for a completed feature gets file-not-found,
> falls into "anything else", and **HALTs** — producing exactly the permanently-stuck
> shipment §5.2 exists to prevent.

* `active` → proceed (**S4**).
* `done` → proceed via **idempotent resume** (**S3**), and only after §5 conditions 1–4
  revalidate; if any fails, **HALT**.
* anything else → **HALT** (**S5**). Do not widen §5 condition 5 in-flight. Return to
  **Stage**.

> **Rev 5 — PO-1b reads by ID, and the tool is pinned (finding B-10/B-13).** `backlogit get
> 022-F` **or** the MCP `backlogit_get_item` are **both** acceptable — the load-bearing rule
> is *read by status/ID, never by path*, not CLI-exclusivity, and rev 4 mandated the CLI with
> no justification. The **a1 transition itself** is pinned: it MUST be performed with
> `backlogit_move_item` / `backlogit move {id} --status done` — **never** a direct file edit.
> Ship MUST NOT invoke `backlogit update --complexity`, which is advertised but fails (exit 1)
> in this workspace (finding B-12).

PO-1b is retained even though PO-1a measured the general behavior, because a probe
measures the **engine**, not this specific live record at that specific moment. Widening an
authority envelope mid-execution remains forbidden; PO-1b keeps the gate fail-closed
regardless of what the probe found.

### 10.1.1 Probe 18 travels with A

C2's correction rests on the claim that the engine **refuses** `backlogit move
<shipment_id> --status shipped` for shipment artifacts (exit 9). That is **Probe 18**, and
it is A's justification, not C's. A cites
`docs/plans/evidence/2026-09-12-task-only-shipment-finalization/probe18-liveconfig-inbound-*`
and the compound note
`docs/compound/2026-05-07-backlogit-shipment-status-constraints.md`. Asserted by **AC-13**.

> **Rev 3 — evidence directory is RETAINED.** The evidence directory was nominally
> *"owned by C"*. C is deferred, but the directory is **not** deferred with it: A cites
> **Probe 18** here and **Probe 22** in §10.1, and the A-only decision relies on **Probe
> 24** for `017-S`'s 13-member shape. The directory stays in place and untouched.

### 10.2 Rollback

Two files on a feature branch. `git revert` the merge commit; Ship's closure behavior
returns to today's state (including, knowingly, the §2.2 drift). No backlog mutation is
performed by the implementation commit itself, so no backlog rollback is required.

> **Rev 5 — the rev-4 contradiction is RESOLVED (plan-review finding A-1).** Rev 4's §9.4
> declared *"no operator override … no `--force`, and no 'proceed with a recorded residual'
> path"* while this section **requires an operator `git revert`** — precisely the override
> §9.4 forbade. Two clauses of the same plan directly contradicted each other, and the
> practical consequence was an **unrecoverable stall**. The no-override clause is
> **withdrawn** (§9.4). Rollback per this section is now an **explicitly available** operator
> action at every stage, and §9.4a lists it as one of three named recovery options.
>
> **Rollback also covers the evidence files (finding B-1).** `probe25-*.{ps1,txt}` are
> committed by **Stage pre-claim**, in a **separate commit** from A's 2-file implementation.
> Reverting A's merge commit therefore does **not** disturb them, and it does not need to:
> they are planning evidence, not implementation surface. If the plan is abandoned outright,
> Stage archives the evidence with the plan.

**Rev 3 — the rollback chain is now A-only.** Rev 2 required a mandatory `C → B → A`
revert order because B and C would have layered on top of A's delegation. B and C are
**deferred and archived** and land on nothing, so **no downstream contract surface depends
on A's delegation**. A can be reverted **on its own**, with one obligation below.

**The one remaining obligation: hold `017-S` before reverting A.** Reverting A removes
Step 6.1(a1), which is the sole reason `017-S`'s 13-member fully-covered root can close.
A revert therefore changes the contract `017-S` was planned against. Before reverting A:

1. confirm `017-S` is **not** `active` (it must not be mid-execution);
2. explicitly **hold `017-S`** — a **Stage** action, recorded in the backlog, taken
   **before** the revert;
3. either restore `017-S` to its former task-only 12-member shape **or** re-plan it —
   do **not** re-route it on this plan once a1 is gone.

This is the general rule *"a revert changes the contract queued work was planned against"*
applied to the only dependent that now exists. It is stated here normatively; rev 2
delegated it to plan C §10.1, which is now **deferred** and must not be routed from.

### 10.3 Failure paths

| Failure | Response |
|---|---|
| H1 cannot be reached in budget | HALT, return to Stage. Never split — A is already minimal |
| a1: a **present** feature member fails any §5 condition 1–5 | **HALT** fail-closed (§5.1 **S4**); surface which condition; no close |
| a1: manifest has **zero** feature members | **Explicit no-op** (§5.1 **S1**) — record `A1_NOT_APPLICABLE` and proceed to `a`. **Not** a failure |
| a1: feature member already `done`, conditions 1–4 revalidate | **Idempotent resume** (§5.1 **S3**) — record `A1_ALREADY_DONE` and proceed to `a`. **Not** a failure |
| a1: feature member already `done`, but any of conditions 1–4 **fails** | **HALT** fail-closed (§5.1 **S3**) — a `done` feature over unfinished descendants is corruption, not a resume |
| a1: feature member in any other status (`queued`, `blocked`, `review`, …) | **HALT** fail-closed (§5.1 **S5**); record the observed status; return to Stage |
| a1: manifest has **more than one** feature member | **HALT** fail-closed (§5.1 **S2**) — record `A1_MULTIPLE_FEATURE_MEMBERS: {ids}`; evaluate **no** §5 condition and mutate **no** feature; return to Stage |
| **(rev 5; re-based rev 6)** a1: member anomaly — missing/unresolvable `artifact_type`, record in **both** queue and archive (containment), record in **neither**, or a member **declaring `status: archived`** with malformed provenance or **not on the §5.0 disposition list** | **HALT** fail-closed (§5.1 **S0**) — evaluated **before** `n` is computed; no mutation. **An archive-located member declaring `status: done` is a normal completed member and raises NO trigger** |
| **(rev 5)** Probe 25 **FAILS** or digest-mismatches at Stage §9 step **0a** | **HALT pre-claim.** `021-S` is **not claimable**; nothing has merged and nothing is in flight. Return to Stage to re-plan. *(Already executed: `PROBE25_RESULT=PASS`.)* |
| **(rev 5; SPLIT in rev 7, finding E-8)** S2 §9.4a **read-only re-verification** fails on **transient toolchain/executable drift only** (check 4, `backlogit_executable_sha256`) | **HALT.** `021-S` stays `active`, **no** a1 mutation, **not** shipped, **not** archived. **The operator MAY correct the drift and RE-VERIFY READ-ONLY, in place** — the committed evidence is untouched and still valid, so a clean re-verification clears this and A proceeds |
| **(rev 5; SPLIT in rev 7, finding E-8)** S2 §9.4a **read-only re-verification** fails on **committed evidence, hash or ancestry** (checks 1, 2, 3, 5 or 6) | **HALT. This CANNOT be cleared in place.** `021-S` stays `active`, **no** a1 mutation. Re-committing evidence post-merge moves `evidence_commit_sha` to a **descendant** of the merge SHA, so check 3 **still fails** — **`git revert` per §10.2 (holding `017-S` first) and a re-plan are required** |
| **(rev 5; ORDER CORRECTED rev 6)** S2 checkpoint **MATCH set** (exact filename/reference **AND** `agent: ship` **AND** `session_id`) has **zero** or **more than one** member; or any enumerated record is malformed/quarantined; or identity mismatches at re-confirmation | **FAIL CLOSED** to operator handoff (§9.1 step **2a**). The installed "zero candidates is normal startup" continuation **does not apply to S2**. **Unrelated active `stage`/`ship` checkpoints are a WARNING only and never strand the handoff** |
| **(rev 4)** `agent-engram` unreachable at S2 restore | **FAIL CLOSED** to operator handoff (§9.1 step 7) — **no prune and no resume**. A file-based prune degradation is not permitted |
| Pre-mode returns `RECONCILE_FAIL` | HALT; do not proceed to close; surface the report. **No mutation has occurred; no restore** |
| Classification returns `SAFE_CLOSE`, a non-close procedural outcome (`HALT`/`RECONCILE_FAIL`), an unknown token (incl. `BLOCKED`), or an absent/empty/ambiguous/unreadable result, for A or `017-S` | **HALT BEFORE ANY CLOSE MUTATION** — indicates the manifest is not the fully-covered root this plan asserts, or the classifier is unusable. Return to Stage. **No `SAFE_CLOSE` fallback** (AC-28) |
| **(rev 7, finding E-7)** Cascade **step 2** returns a **non-empty `returned_ids`** | **Detect-after-mutate.** P-015 violation action: **`git restore -- .backlogit/queue/ .backlogit/archive/`**, **VERIFY the restore against the Step 0(b) pre-call snapshot**, then **HALT**. No retry, no second close path |
| Cascade **step 3** archives anything outside `allowed_ids` (`archived_ids − allowed_ids` non-empty) | **Detect-after-mutate.** Same **mandatory verified restore** then **HALT** |
| **(rev 7, finding E-7)** Cascade **step 3** fails to archive a required member (`required_ids − archived_ids` non-empty) | **Detect-after-mutate.** Same **mandatory verified restore** then **HALT** |
| **(rev 7, finding E-7)** Cascade **step 4** finds any `parent_id` **altered or cleared** vs the Step 0(b) snapshot | **Detect-after-mutate.** Same **mandatory verified restore** then **HALT**. *(Rev 6 omitted this row entirely — a valid-ID cascade that cleared a `parent_id` halted with mutated state and no restore obligation.)* |
| **(rev 7, finding E-7)** The mandatory restore itself **cannot be verified** | **HALT** with the backlog flagged **operator-reconcile-required**. **Never continue and never re-run the cascade** |
| S1 attempts the a1 gate or the close after its own merge | **HALT** — §6 forbids it; S1 must checkpoint and end. Return to Stage; the two-session property has been violated |
| S2 recovery hits a validation/quarantine anomaly, ambiguity, or a `stage`-owned record | **FAIL CLOSED** to operator handoff (§9.1). No restore, no resume, no resolve. The checkpoint stays `active` and is **excluded from `cleanup_checkpoints`** regardless of age |
| `--phase lifecycle` gate invoked after the archive | **HALT** — Probe 20 tripwire; it fails closed with zero active shipments and nothing remains to re-claim |

## 11. Acceptance criteria

- **AC-1** Role Boundary carries the narrow grant with all five §5 conditions, fail-closed.
- **AC-2** Step 6.1(a1) exists, positioned after `a0` and immediately before `a`.
- **AC-3** The §6 authority-escalation rule is stated normatively in `_ship.agent.md`.
- **AC-4** C1–C6 are all applied; the delegation block names P-015 + `shipment-reconcile`.
- **AC-5** `_ship.agent.md` contains **no classification predicate** — no restatement,
  re-implementation or paraphrase of the conditions under which a close path is selected —
  and **no** occurrence of `TASK_ONLY_FINALIZE`. **(rev 6, B-3: a result-token allowlist used
  solely to validate the classifier's output IS permitted; a selection condition is not.
  See §4.1.)**
- **AC-6** The drifted `children`-only coverage rule is **absent**.
- **AC-7** Exactly **one** Go test function; H0 red 1 of 1; H1 green; `go vet` clean.
- **AC-8** `markdownlint` clean on the changed file.
- **AC-9** Closure follows §9 including the S1 checkpoint-and-end and the S2 fresh-session
  parent completion. AC-9 is **not** satisfied if one session performs both. The
  checkpoint MUST be written through **`backlogit_create_checkpoint` (MCP) *or*
  `backlogit checkpoint create --state-dump` (CLI) — both are valid (rev 5, C-9)** —
  tool-timestamped, with **top-level** `schema_version: 1`, `agent: ship`,
  **`session_id` (rev 5, A-8 — REQUIRED by the installed Checkpoint Payload Contract)**,
  a `phase`, and a `resume_hint` naming `021-S`, `022-F`, `022.001-T` and the merge SHA,
  with domain data under `context`; its `created_at` must precede S2's resume. S1 records
  the **emitted filename** and the `session_id`; S2 matches **both exactly**.
  **S2 executes the full §9.1 recovery protocol — all thirteen steps (0, 1, 2, 2a, 3–11), in
  order.** AC-9 is **not** satisfied unless S2 demonstrably performed **every** element
  below; rev 3's shorter list is superseded and is **not** sufficient evidence:
  * enumeration via `backlogit_list_checkpoints` passing `consumer_id: "ship"` (**access
    identity, not a filter** — rev 7, F-8) and **no** `status`/`agent` filter;
  * **anomaly gate evaluated FIRST**, over the **full enumeration**, reading the **top-level
    `needs_quarantine` / `quarantined` / `total` counters** (rev 7, F-9) plus per-record
    required-field validity, before any filtering or counting;
  * **the S2-SPECIFIC MATCH rule (rev 5, A-9; ORDER CORRECTED rev 6, B-8)** — **filter
    first** to the MATCH set (exact filename/reference **AND** `agent: ship` **AND** the exact
    `session_id`), **then** require **exactly one MATCH**. **Zero MATCHES is FAIL CLOSED**
    (the installed "normal startup" continuation does not apply); **more than one MATCH is
    FAIL CLOSED**; **unrelated active checkpoints are a WARNING only and never strand the
    handoff**, and are never resolved, pruned or touched by S2;
  * **explicit owner selection by filename, even for a single candidate** — never auto-pick;
  * `agent: ship` ownership validation (a `stage`-owned record is never selectable);
  * **explicit operator confirmation before restore** (no automatic resume; no age-based
    liveness inference);
  * **the `restore → prune/gate → resume` ORDERING, executed in that order** — never
    restore → resume → prune **(rev 4)**;
  * **the Engram reachability gate** — `agent-engram` is installed here, so the prune step
    is mandatory and **not** the backlogit-only no-op; if the engram substrate is
    unreachable, S2 **FAILS CLOSED to operator handoff with no prune and no resume**, and a
    file-based prune degradation is **not** an acceptable substitute **(rev 4)**;
  * **the never-prune allowlist honored intact** — the active-shipment/active-task cursor
    (`021-S`), the unresolved-checkpoint pointer itself, and all recorded gate verdicts
    **(rev 4)**;
  * **same cursor** — the same single active shipment `021-S`, no parallel resume and no
    new worktree;
  * **resolve only after a confirmed successful resume**, for that one selected checkpoint
    only, with no bulk sweep;
  * **fail closed with no fresh-start fallback**; and a checkpoint left `active` by a
    fail-closed handoff is **excluded from `cleanup_checkpoints` regardless of age**
    **(rev 4)**.

  *Residual, stated honestly:* the artifact is not tamper-proof — the two-session property
  ultimately rests on process integrity, and the checkpoint is evidence, not enforcement.
- **AC-10** PO-1a is **discharged by Probe 22** and cited (§10.1); PO-1b executed before the a1 gate and recorded.
- **AC-11** Exactly **2 implementation files** changed (`_ship.agent.md` + the one Go test) and **0 new production files**. **(rev 6, finding D-1/B-1: planning artifacts — this plan, `docs/plans/evidence/probe25-*`, `docs/memory/`, `docs/decisions/` and `.backlogit/` records — are authored by Stage in separate commits and are NOT counted against this 2-file implementation surface.)**
- **AC-12** `021-S`'s manifest is verified to be `[022-F, 022.001-T]` **at claim** (§4.3 — the re-shape is already complete; this is a verification, not a task).
- **AC-13** Probe 18 is cited as C2's justification (§10.1.1).
- **AC-14** C1 ships with the §4.2 pinned reconciled wording; Go test row 11 passes.
- **AC-15** **(rev 2; re-labelled rev 6)** Step 6.1(a1) states the **zero-feature-member explicit no-op** (§5.1 **S1**, legacy case i); Go test row 12 passes.
- **AC-16** **(rev 2; re-labelled rev 6)** Step 6.1(a1) states that an unmet §5 condition on a **present** feature member **HALTs** (§5.1 **S4**, legacy case iii) and is not treated as a no-op.
- **AC-17** **(rev 2)** Step 6.1(a1) states **idempotent resume** for a feature member already `done` (§5.2); Go test row 13 passes.
- **AC-18** **(rev 2)** The lifecycle order of §9 is followed: `022.001-T -> done` **before** the implementation PR merges; the closure branch created from fresh `main` **before** any post-merge backlog mutation; closure artifacts and P-020 on the closure branch **before** the closure PR is pushed; `sync` and post-mode in installed order.
- **AC-19** **(rev 2)** The `a0` `--phase lifecycle` gate runs **while `021-S` is still active**; **no** `--phase lifecycle` invocation occurs after the archive (Probe 20 tripwire).
- **AC-20** **(rev 5, RE-SEATED — supersedes rev 4's S2-run gate)** The §9.4 root-included cascade fixture (Probe 25) is **authored, executed and committed by STAGE at §9 step 0a, BEFORE `021-S` is claimed**. **`021-S` is not claimable** until the evidence is committed and PASSing. The evidence lives at `docs/plans/evidence/2026-09-12-task-only-shipment-finalization/probe25-root-included-cascade-fixture.{ps1,txt}`, records `DIGEST_GATE=PASS` against SHA-256 `1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98` **with the engine resolved at runtime via the registered command (no hardcoded path is a requirement of any check)**, and satisfies **all eight** PASS criteria — **emitted as nine tokens** (criterion 5 splits into `5A`/`5B`) — over the **exact 13-member root-included manifest with 11 members genuinely declaring `status: archived` (`archived_status: queued`)**. **STATUS: DISCHARGED** — `PROBE25_RESULT=PASS`, `FAILED_CRITERIA=0`, measured at `4fd21c5`. **Ship never authors, modifies, re-runs or regenerates this artifact (P-010).**
- **AC-20a** **(rev 5; extended rev 6; RECOVERY SPLIT + CRITERION NAMES PINNED in rev 7, E-8/F-7)** At §9 step 12a the fresh S2 session performs the §9.4a **read-only** re-verification and **all six** checks pass: (1) exact topology `TOPOLOGY_MEMBERS=13` / `ROOT_INCLUDED=True` / `PRE_ARCHIVED=11` / `ARCHIVED_STATUS_FLAVOUR=queued`; (2) `PROBE25_RESULT=PASS`, `FAILED_CRITERIA=0`, and **all nine PINNED STEP 11 aggregate tokens** (`C1_DIGEST_GATE`, `C2_PREMODE_PROCEED_11_PRE_ARCHIVED`, `C3_CLASSIFICATION_CASCADE`, `C4_RETURNED_IDS_EMPTY`, `C5A_ARCHIVED_MINUS_ALLOWED_EMPTY`, `C5B_REQUIRED_MINUS_ARCHIVED_EMPTY`, `C6_PARENT_ID_PRESERVED`, `C7_SHIPMENT_ARCHIVED_SHIPPED`, `C8_NOTHING_OUTSIDE_MANIFEST_TOUCHED`) reading `PASS` — **not** the eight inline `CRITERION_n` tokens; (3) **`evidence_commit_sha`** — the **LAST** commit touching either evidence file — is an **ancestor of the verified merge SHA**; (4) the **currently-resolved** engine digest matches **`ENGINE_SHA256_EXPECTED` read from the committed transcript**, with **no absolute path required**; (5) **`fixture_script_sha256`** and **`fixture_result_sha256`** each match, established by comparing the **blob read from `evidence_commit_sha`** to the current file, hashed **separately**, with neither file missing, empty or modified — **and neither digest is embedded inside the file it hashes**; (6) the transcript is **internally consistent**. **Every check is a read.** On failure S2 HALTs **naming the failed field**, `021-S` stays `active` with **no** a1 mutation, and the recovery path is **class-appropriate (rev 7, E-8)**: a **transient toolchain/executable drift (check 4)** may be corrected and **re-verified read-only in place**, which genuinely clears it; a **committed-evidence/hash/ancestry failure (checks 1, 2, 3, 5, 6)** **cannot be cleared in place** and requires **`git revert` per §10.2** plus a re-plan. AC-20a is **not** satisfied if S2 writes to `docs/plans/`.
- **AC-21** **(rev 5; S0 RE-BASED in rev 6, B-1/B-5)** Step 6.1(a1)'s selector is the **strictly ordered, non-overlapping** sequence **S0 → S5** of §5.1, computing `n` by `artifact_type` and never by ID suffix: **S0** anomaly gate **before** `n`, with **every trigger keyed on declared status or physical containment and NONE on archive location** — an archive-located member declaring `status: done` raises **no** trigger; **S1** `n == 0` ⇒ explicit no-op; **S2** `n > 1` ⇒ HALT with `A1_MULTIPLE_FEATURE_MEMBERS` and **no** feature mutation whatsoever; **S3** `n == 1` + `done` ⇒ idempotent no-op **only after conditions 1–4 revalidate**, else HALT; **S4** `n == 1` + `active` ⇒ full five-condition guard then `active -> done`; **S5** `n == 1` + any other status ⇒ HALT. **No input may satisfy two steps.** **a1 runs BEFORE `shipment-reconcile mode: pre` and MUST NOT require pre-mode to have run** — its completion guard reads declared status, containment, topology and the §5.0 allowlist directly. Go test row 16 passes.
- **AC-22** **(rev 4; assertions made NON-VACUOUS in rev 5; re-based rev 6; POINTER-ONLY in rev 7, findings E-2/E-3)** Both stale queue-only pre-mode summaries are reconciled at `_ship.agent.md` — **C7** (L255–256, Step 0.5 item 6) and **C8** (L775–777, Step 6.1(a)) — to **POINTER-ONLY DELEGATION** per §4.4. Each reconciled sentence **delegates the per-item classification to the installed `shipment-reconcile` `mode: pre` contract at the required `expected_status`** (C7: the `expected_status` already named one line above; C8: `expected_status: done`), **proceeds only on an authoritative `PROCEED`**, and **halts otherwise**. Each **preserves** the `PROCEED` conjuncts **no orphans** and **`record-consistent`**; C8 additionally preserves the **single-writer lock** clause; C7 leaves the adjacent **`Scope note (139-F/139.001-T)`** intact. **Neither sentence may remain queue-only, and NEITHER MAY STATE ANY PER-ITEM CLASSIFICATION PREDICATE — locational OR declared-status.** Go test rows **14** and **15** are **`compound +/−`** — each asserts the §8.1.1 **delegation** PRESENT phrase **and** the ABSENT exact old literal — so neither can pass while its site keeps the old text, **and neither is satisfiable by restating a predicate**. The rev-3 instruction that *"a reviewer MUST NOT fail A for its absence"* is **withdrawn**. **Rev 7 asserts nothing about which taxonomy the installed pre-mode uses internally.**
- **AC-23** **(rev 5; RECALCULATED in rev 6)** The inventory is **8 clause sites + 2 additive**, the recalculated budget is **97 min** with a **23 min** margin to the 2-hour rule, **and §12's rows sum exactly to that total (`8+18+4+4+12+4+4+4+29+10 = 97`)**, and the Go test carries **18 rows** — of which **4 are `compound +/−`**, **1 is `compound ++`** and **2 are negative** — in **one** function with **≤4 helpers**, using **stdlib assertions only** and **reusing** `repoRoot(t)`. No site is excluded from the inventory in order to preserve a previously-pinned budget, and no row is adjusted to preserve a previously-pinned total.
- **AC-24** **(rev 5; RE-BASED ON DECLARED STATUS in rev 6, B-1)** §5.0's pre-archived-descendant exemption is triggered **only** by a descendant **declaring `status: archived`** — **never** by archive location — is applied **only** to the enumerated **11 `017-S` IDs**, and **only** when all five §5.0 gates hold. `archived_status: queued` is **never** treated as proving completion. Any descendant **declaring `status: archived`** that is not on the list ⇒ **HALT**. **A descendant declaring `status: done` is a COMPLETED descendant that satisfies §5 condition 1 even when stored under `archive/`**, is never tested against the allowlist, and is never required to carry archive provenance. `021-S` has **no** pre-archived descendants **at the moment a1 runs** — `022.001-T` is `done`, not `archived`.
- **AC-25** **(rev 5; resolved rev 6; SINGLE-AUTHORITY in rev 7, findings E-4/E-5/E-6)** `_ship.agent.md` implements §4.1's delegation contract: it MAY name the **installed CLOSE VERDICT tokens** — **`CASCADE` and `SAFE_CLOSE`, and only these** — **solely as an output-validation allowlist**, and it **MUST NOT restate or re-implement any classification predicate**. **`HALT` and `RECONCILE_FAIL` are non-close procedural outcomes, not close verdicts**, and authorize no close path. **`BLOCKED` is not an installed token and is treated as unknown.** Ship **HALTs, fail closed, no mutation** on: a token outside the installed close-verdict allowlist (**including `BLOCKED` and any future token**); an absent, empty, unparseable or ambiguous result; a classifier that is **not installed or not readable**; or a non-close procedural outcome. **The rev-6 runtime `policy_id` check, the runtime policy-version comparison, and the P-015/skill disagreement check are DELETED** — the installed skill is the **runtime** authority, P-015 is the **statically reviewed** authority, and requiring Ship to detect semantic disagreement would force it to re-run the predicates in violation of **AC-5**. The allowlist is bound to the **currently installed** skill — **no unversioned future capability is authorized**, and widening it requires its own review. **AC-25, AC-5 and row 9 are mutually satisfiable**: close-verdict tokens may be listed, predicates may not.
- **AC-26** **(rev 5)** Rows **17** and **18** pass: C8's reconciled sentence preserves the **single-writer lock** and **orphan scan** clauses; C7's replacement leaves the adjacent `Scope note (139-F/139.001-T)` intact.
- **AC-27** **(rev 6, finding B-1/B-2; RECONCILED in rev 7, finding E-1)** **No LIFECYCLE CLASS anywhere in A is determined by file location.** Every completion, exemption and anomaly decision in §4.4, §5, §5.0, §5.1 and §9 reads the record's **own declared `status`** to determine what the record *is*. Physical-location reads survive in exactly **two** narrowly-scoped jobs, and neither is a lifecycle inference: **(a) GENERAL containment** — exactly one physical copy, `queue/` **XOR** `archive/`, applied to **every** member at **every** status (both ⇒ torn, neither ⇒ missing), **preserved in full**; and **(b) ARCHIVE ROUTING INTEGRITY** — applied **only after** declared `status: archived` has already been established at §5.0 condition 1, requiring that the single copy be the `archive/` one. **(b) is downstream of the lifecycle determination and can never produce one.** **`022.001-T`, `done` and archive-located at the a1 gate, passes §5 condition 1, is never subjected to (b), and raises no S0 trigger.** The rev-6 formulation — which required `archive/` *specifically* inside condition 2 — is **withdrawn**.
- **AC-28** **(rev 6, finding B-6; SEATED AS A RELEASE-INSTANCE CONTRACT in rev 7, finding E-6)** **a1 runs FIRST; Ship MUST then invoke the installed `shipment-reconcile`.** For the **two exact release instances** `021-S` and `017-S`, this plan's execution contract **EXPECTS `CASCADE`**: a `CASCADE` verdict enters the **pre-existing** cascade branch; **`SAFE_CLOSE`, an unknown token (including `BLOCKED`), an unavailable or unreadable classifier, an empty or ambiguous result, `HALT`, or `RECONCILE_FAIL` ⇒ HALT BEFORE ANY CLOSE MUTATION** and return the shipment to Stage. **There is NO `SAFE_CLOSE` fallback on these routes.** **This expectation is a RELEASE-INSTANCE precondition carried by this plan and the `021-S`/`017-S` records (and by `PA-021-CASCADE`/`PA-017-CASCADE`) — it is NOT global ID-specific logic in `_ship.agent.md`, which carries only the general, ID-free delegation rule.** A adds **no** cascade path, **no** verdict and **no** rollback; it discloses that it adds **reachability** to an inherited detect-after-mutate residual.
- **AC-29** **(rev 6, finding B-8; TOOL SURFACE CORRECTED in rev 7, F-8/F-9)** S2's checkpoint recovery **enumerates all records with no filter** — passing `consumer_id: "ship"` as **access identity, never as a filter, and never via the CLI `--agent` flag** — runs the **anomaly gate over the full enumeration first**, reading the **top-level `needs_quarantine` / `quarantined` / `total` counters** plus per-record required-field validity, **then** filters to the MATCH set (**exact filename/reference AND `agent: ship` AND `session_id`**) and requires **exactly one MATCH**. Zero or more than one MATCH, a malformed record, or an identity mismatch ⇒ **FAIL CLOSED**. **Unrelated active Stage- or Ship-owned checkpoints are surfaced as a warning and never strand the handoff**, and are never resolved, pruned or touched. Order is **restore → prune/gate → resume**, and `backlogit_resolve_checkpoint` runs **only after a confirmed successful resume**, for that one checkpoint only.
- **AC-30** **(rev 7, Constitution Principle VII / finding `G-2`)** Both live destructive cascade invocations carry an explicit **approved-action record** in `## Strict Safety — Approved Destructive Actions` — **`PA-021-CASCADE`** (target: `021-S`, manifest exactly `[022-F, 022.001-T]`) and **`PA-017-CASCADE`** (target: `017-S`, the exact 13-member manifest, dependency satisfied by a **shipped** `021-S`). Each record carries a ProposedAction summary/targets/`change_kind`/rollback/`approval_required`, **`ActionRisk: destructive`**, **`ActionResult: approved`**, the operator authorization timestamp **`2026-09-13T11:35:43-07:00`**, and its conditions: **authoritative `CASCADE` verdict**, **all preflight gates pass**, **exact unchanged manifest and topology**, **snapshot available**, **no scope expansion**. **Any mismatch invalidates the approval and HALTs before mutation.** **Post-mutation failure restores from the snapshot, VERIFIES the restore, then HALTs.** The approval authorizes **only these two exact actions** — **not** admin fallback, force-push, history rewrite, any other destructive operation, any other shipment, or a re-scoped manifest — and is **not** a claim authorization.

## Constitution Check

**(rev 5 — plan-review finding A-13.)** The constitution's Governance section requires every
implementation plan to map its work against Principles I–XI and to document justified
deviations. Rev 4 mapped exhaustively against *workflow policies* but never against the
*constitutional principles*, leaving genuinely justified deviations undocumented.

| Principle | Applies | Assessment |
|---|---|---|
| **I. Safety-First Go** | yes | The only Go artifact is one table-driven contract test. No `unsafe`, no goroutines, no I/O beyond reading one file. Errors surface via `t.Errorf`, never panics or ignored returns |
| **II. Test-First Development (NON-NEGOTIABLE)** | yes | **Honored literally.** §7 mandates H0 red → H1 green; the test lands **before** any `_ship.agent.md` edit. **Deviation (justified) — see D1 below** on scenario count |
| **III. Workspace Isolation and Security Boundaries** | yes | Probe 25 runs in a disposable, gitignored, workspace-contained directory seeded from live config; `FIXTURE_ISOLATED_FROM_LIVE=True` and `LIVE_BACKLOG_UNMUTATED=True` are **measured**, not asserted |
| **IV. CLI Workspace Containment (NON-NEGOTIABLE)** | yes | Every probe call is `--cwd $ws`. The engine is resolved via the **registered command**, not an absolute path (§9.4). No write escapes the repo |
| **V. Structured Observability** | yes | a1 emits named, greppable tokens (`A1_NOT_APPLICABLE`, `A1_ALREADY_DONE`, `A1_MULTIPLE_FEATURE_MEMBERS`); Probe 25 emits one `KEY=VALUE` line per criterion |
| **VI. Single Responsibility** | yes | **(rev 7, findings D-4 / `F-11` — the rev-6 row assessed COHESION; VI's body is DEPENDENCY DISCIPLINE.)** **Dependency discipline:** the Go artifact adds **zero** module dependencies — assertions are **stdlib only** (`testing` + `strings`), **`testify` is confirmed ABSENT from `go.mod`** and is not introduced, and the test **reuses** the existing package-level `repoRoot(t)` helper rather than adding a fixture framework. The `_ship.agent.md` change **removes** a dependency direction: Ship stops depending on a *local copy* of the classification and depends only on the **delegated** skill/P-015 surface. Net dependency edges introduced: **none**. **Cohesion (also satisfied):** A does exactly two things — grant one narrow authority, and replace a duplicated classification with delegation — and §2.2's drift is the cost of the duplication this removes. **See the Width-Isolation deviation D5 below** for why one task legitimately spans a Markdown contract and a Go file |
| **VII. Destructive Command Approval (NON-NEGOTIABLE)** | yes | The cascade is destructive. It is invoked **only** on a machine-selected `CASCADE` verdict from the authoritative classification, and only after a1 (§9.3 rev 6). **Correction (rev 6, finding C-6): the rev-5 claim that it runs "never against the live backlog by this plan" was FALSE** — §9 step 15 and §9.3 step 6 both invoke it **live**. What is true: the **measurement** (Probe 25) runs only inside the disposable fixture, and every **live** invocation remains gated by P-015, by the pre-mode `PROCEED`, and by the no-fallback HALT rule of AC-28. **A adds reachability to this destructive path for `017-S`** (§9.4b). **(rev 7, finding `G-2`: operator-approval routing is no longer machine-gates-only.)** Both live destructive invocations now carry an **explicit recorded approved-action entry** — **`PA-021-CASCADE`** and **`PA-017-CASCADE`** — in `## Strict Safety — Approved Destructive Actions`, each with `ActionRisk: destructive`, `ActionResult: approved`, the operator authorization timestamp, and the exact conditions whose violation invalidates the approval and halts |
| **VIII. Explicit Safety Modes for Elevated Risk** | yes | **Deviation (justified) — see D4 below** |
| **IX. Git-Friendly Persistence** | yes | All artifacts are line-oriented Markdown/PowerShell/Go. The backlog mutation at 6.1(e) is committed on the closure branch, never on `main` |
| **X. Agent Context Efficiency** | yes | A **removes** ~20 lines of re-derived classification prose from `_ship.agent.md` and replaces them with a delegation pointer — a net context reduction on Ship's hot path |
| **XI. Merge Commit History Preservation (NON-NEGOTIABLE)** | yes | §10.2 rollback is `git revert` of the merge commit — never a history rewrite, never a force-push. No amend anywhere in this plan |

**Documented deviations — justified, not waived.**

* **D1 — Task granularity: 18 test scenarios vs *"fewer than 4"*.** The granularity heuristic
  targets *behavioral* scenarios. These 18 are **rows in one table-driven function** over a
  **single** file, sharing one fixture and ≤4 helpers; each is a substring or index assertion,
  not an independent scenario. Collapsing them would **reduce** falsifiability — and rev 4's
  attempt to economize is exactly what produced the vacuous rows 14–15 (finding A-7).
  **Justification: assertion density, not scope creep.** The §12 budget carries them at ~29 min.
* **D2 — *"fewer than 5 functions"*: 1 test function + ≤4 helpers = 5 symbols.** The
  constraint is read as **one function under test per task**; helpers are shared machinery.
  This is also a hard Ship constraint (Step 4.3's one-function batch rule, the `021.001-T`
  deadlock), so **1 function is mandatory** and the helpers cannot be split into a second
  function without breaking it. **Justification: an externally imposed constraint.**
* **D3 — `tests/integration/` rather than an inline `_test.go` or `tests/contract/`.** The
  instructions' taxonomy names `tests/contract/`, which **does not exist** in this repo. The
  established precedent for text-contract tests here is `tests/integration/`
  (`build_script_test.go`), whose package-level `repoRoot(t)` this test **reuses**.
  **Justification: existing precedent; creating a new top-level test tree is out of scope.**
* **D4 — Principle VIII: a global-closure blast radius without an explicit safety-mode
  directive** *(finding C-1)*. A changes closure behavior for **every** future shipment.
  **Accepted with compensating controls rather than a mode declaration**: the change
  *removes* authority duplication rather than adding reach; four negative widening guards
  (rows 7–10) fail if anything new is authorized; the new grant is bounded by five
  conjunctive conditions, one position, one scope and one transition; and A closes by the
  **pre-existing** path, so a defect in the new wording cannot corrupt A's own closure.
  **Recorded as a residual**, not as a satisfied principle.

  > **Rev 7 — the Governance-required REJECTED SIMPLER ALTERNATIVE, stated as such
  > (findings D-3 / `G-3`).** Governance requires a deviation to name the simpler
  > alternative that was considered and rejected. Rev 6 argued for D4 without ever
  > identifying one.
  >
  > **The simpler alternative considered and REJECTED: scope the grant to `021-S` and
  > `017-S` by ID** — write the two shipment IDs directly into `_ship.agent.md`'s Role
  > Boundary so the new authority literally cannot apply to any other shipment, reducing the
  > blast radius from global to two records and removing the Principle VIII question
  > entirely.
  >
  > **Why it was rejected — three independent grounds, any one sufficient:**
  >
  > 1. **It contradicts finding C-5's decoupling**, which this plan already accepted:
  >    release-instance identifiers belong to **Stage's disposition and this plan**, never to
  >    the global agent contract. Rev 6 removed the `018.*` IDs from `_ship.agent.md` for
  >    exactly this reason; re-adding `021-S`/`017-S` would reintroduce the defect one
  >    section later.
  > 2. **It does not reduce risk, it relocates it.** The next shipment needing the gate
  >    would require **another edit to the same global file** — the identical blast radius,
  >    paid repeatedly, with a growing ID list that is itself the duplication §2.2 exists to
  >    remove.
  > 3. **It cannot be verified negatively.** The widening guards (rows 7–10) assert the file
  >    contains **no** release-specific logic; an ID-scoped grant would require *inverting*
  >    those guards, destroying the mechanical check that the delegation stayed general.
  >
  > **What was taken from the alternative instead.** Its safety intent — *bound the
  > destructive reach to two known instances* — is preserved **without** putting IDs in the
  > global contract, by seating the expectation in this plan and the backlog records
  > (**AC-28**) and by the two explicit approved-action records **`PA-021-CASCADE`** and
  > **`PA-017-CASCADE`**, which name the exact manifests and halt on any mismatch.
  >
  > **Is VIII *satisfied via freeze-scope* rather than deviated?** Engaged honestly, and the
  > answer is **no**. A freeze-scope argument would require A's reach to be *structurally*
  > limited to the frozen set. It is not: the Role Boundary grant is **general and permanent**
  > from the moment it merges, and the two approved-action records bound the **destructive
  > cascade invocations**, not the **grant**. **VIII therefore remains a documented deviation
  > with compensating controls — not a satisfied principle** — and rev 7 declines the more
  > flattering framing.

* **D5 — Width-Isolation: one task spans a Markdown file and a Go file** *(rev 7, findings
  D-4 / `F-11`)*. The task-granularity rule asks for a **single skill domain per task** (code
  **or** docs **or** tests **or** config). `022.001-T` edits `_ship.agent.md` (Markdown) and
  `tests/integration/ship_feature_completion_contract_test.go` (Go), which reads as two
  domains and was never explained.
  **Justification: this is ONE domain — an executable contract and its verification.**
  `_ship.agent.md` is **not documentation**. It is the **agent's executable instruction
  surface**: Ship reads it at run time and acts on the sentences in it, which is precisely
  why §4.0 ruled that a false summary in it is *operative, not descriptive*. The Go test is
  not an independent test-authoring effort either — it is a **table of substring assertions
  over that one file**, with no logic of its own, whose every row is derived mechanically
  from §8.1.1's pinned literals.
  **They are therefore the contract and its compiler-check, and splitting them is strictly
  worse on three counts:** (1) it would produce a task that **cannot be verified** (contract
  edited, nothing asserting it) and a task that **cannot compile green** (assertions for text
  that does not exist yet), breaking §7's mandatory **H0 red → H1 green** ordering, which
  requires both halves in one atomic change; (2) the red→green transition **is** the
  acceptance evidence for AC-22/AC-25 — separated, neither half has an acceptance criterion
  it can satisfy alone; and (3) it would double the Ship sessions for a **97-minute** task
  that is already inside the 2-hour rule with a 23-minute margin.
  **Precedent:** this workspace already treats text-contract-plus-Go-assertion as one unit —
  `tests/integration/build_script_test.go` asserts over shell/build text the same way, and
  this test reuses its `repoRoot(t)`.
  **Boundary preserved:** the task still touches **exactly 2 implementation files** (AC-11),
  adds **0 production files**, and introduces **0 dependencies** (Principle VI).

## Strict Safety — Approved Destructive Actions

**(rev 7, operator-authorized; Constitution Principle VII, finding `G-2`.)** Both live
destructive cascade invocations reachable from this plan now carry an **explicit recorded
approved-action entry**. Rev 6 mapped Principle VII to machine gates only (P-015, pre-mode
`PROCEED`, the AC-28 no-fallback HALT) and never recorded an **operator approval** for the
destructive operation itself. These two records close that gap.

**Operator authorization timestamp (both records): `2026-09-13T11:35:43-07:00`.**

> **SCOPE OF THIS APPROVAL — read before relying on it.** This approval authorizes
> **exactly the two actions recorded below, and nothing else**. It does **NOT** authorize
> admin fallback, force-merge, force-push, history rewrite, `--force` on any command, any
> other destructive operation, any other shipment, or a re-scoped version of either action.
> **Any mismatch between a recorded condition and the live state INVALIDATES the approval
> for that action and HALTs** — the approval is not re-derivable by the executing agent and
> must not be "interpreted" toward a near-miss.

### `PA-021-CASCADE`

| Field | Value |
|---|---|
| **Action ID** | `PA-021-CASCADE` |
| **ProposedAction — summary** | Close shipment `021-S` via the **pre-existing** P-015 fully-covered-root cascade (`backlogit_ship_shipment`), archiving exactly the manifest members |
| **ProposedAction — targets** | Shipment **`021-S`**; manifest **exactly `[022-F, 022.001-T]`** — 2 members, no dependencies |
| **ProposedAction — change_kind** | **Destructive** — archives backlog records and transitions the shipment to `shipped`/`archived` |
| **ProposedAction — rollback** | `git restore -- .backlogit/queue/ .backlogit/archive/` from the **Step 0(b) pre-call snapshot**, **followed by a verified re-read of the protected set** (§9.4b, §10.3). Unverified restore ⇒ HALT, operator-reconcile-required |
| **ProposedAction — approval_required** | **YES** — destructive, non-negotiable (Principle VII) |
| **ActionRisk** | **`destructive`** |
| **ActionResult** | **`approved`** |
| **Operator authorization** | `2026-09-13T11:35:43-07:00` |

**Conditions — ALL must hold at invocation time. Any mismatch invalidates this approval and
HALTs before mutation:**

1. the **authoritative installed classification returns `CASCADE`** (§9 step 15). `SAFE_CLOSE`,
   an unknown token (incl. `BLOCKED`), a non-close procedural outcome (`HALT`,
   `RECONCILE_FAIL`), an unavailable/unreadable classifier, or an empty/ambiguous result ⇒
   **approval void, HALT**;
2. **all preflight gates pass** — §9.4a's six read-only re-verification checks at step 12a,
   the a1 S0–S5 selector at step 13, and pre-mode `PROCEED` at step 14;
3. the manifest is **exactly `[022-F, 022.001-T]`**, unchanged, with `021-S` carrying **no
   dependencies** — the topology recorded at authorization time;
4. a **pre-call snapshot is captured and available** for restore (skill Step 0(b));
5. **no scope expansion** — nothing outside the two manifest IDs is archived, reparented,
   created or deleted.

**Post-mutation failure handling:** any post-CASCADE check failing (non-empty `returned_ids`;
either two-set difference; an altered or cleared `parent_id` at step 4) ⇒ **restore from the
snapshot, VERIFY the restore, then HALT** (§9.4b, §10.3). **No retry and no second close
path.**

### `PA-017-CASCADE`

| Field | Value |
|---|---|
| **Action ID** | `PA-017-CASCADE` |
| **ProposedAction — summary** | Close shipment `017-S` via the **pre-existing** P-015 fully-covered-root cascade, archiving exactly the manifest members |
| **ProposedAction — targets** | Shipment **`017-S`**; the exact **13-member** manifest: `018-F`, `018.008-T`, `018.001-T`, `018.001.001-ST`, `018.001.002-ST`, `018.001.003-ST`, `018.002-T`, `018.002.001-ST`, `018.002.002-ST`, `018.002.003-ST`, `018.003-T`, `018.003.001-ST`, `018.003.002-ST` |
| **ProposedAction — change_kind** | **Destructive** — archives backlog records and transitions the shipment to `shipped`/`archived` |
| **ProposedAction — rollback** | Identical to `PA-021-CASCADE`: snapshot restore **plus verified re-read** of the protected set; unverified restore ⇒ HALT |
| **ProposedAction — approval_required** | **YES** — destructive, non-negotiable (Principle VII) |
| **ActionRisk** | **`destructive`** |
| **ActionResult** | **`approved`** |
| **Operator authorization** | `2026-09-13T11:35:43-07:00` |

**Conditions — ALL must hold at invocation time. Any mismatch invalidates this approval and
HALTs before mutation:**

1. the **authoritative installed classification returns `CASCADE`**; every other outcome ⇒
   **approval void, HALT** (same enumeration as `PA-021-CASCADE` condition 1);
2. **all preflight gates pass**, including the a1 gate with §5.0's **exact 11-ID
   disposition allowlist** and its provenance/immutability checks, and pre-mode `PROCEED`;
3. the manifest is the **exact 13 members above**, unchanged, with `018-F` a **root** (no
   `parent_id`) and all 12 descendants resolving to it — the topology recorded at
   authorization time;
4. **the dependency is SATISFIED by a SHIPPED `021-S`** — `017-S` carries
   `dependencies: [021-S] (type: blocks)`, so this approval is **void while `021-S` is
   unshipped**. `017-S` must not be claimed or closed ahead of it;
5. a **pre-call snapshot is captured and available** for restore;
6. **no scope expansion** — nothing outside the 13 manifest IDs is touched.

**Post-mutation failure handling:** identical to `PA-021-CASCADE` — restore, **verify**, then
**HALT**.

### What these records do NOT authorize

* **No admin fallback**, no merge-protection bypass, no `--force`, no force-push, no history
  rewrite, no amend.
* **No other destructive operation** — not `backlogit delete`, not a destructive stash
  removal, not an archive purge.
* **No other shipment**, and **no re-scoped version** of either action. A changed manifest is
  a **different action** requiring **new** authorization.
* **Not a claim authorization.** These records govern the **close mutation** only. Claiming
  `021-S` remains gated by the plan review verdict and the readiness state recorded in
  `## Plan Review — AUTHORITATIVE GATE STATE`.

## 12. Sizing

**Size `M` · Complexity `high`.** Complexity is carried as **enum-validated prose**: this
workspace's `header-def.yaml` defines `size` on the task type but **no** `complexity`
field (confirmed at `.backlogit/header-def.yaml:144–166`, and in
`docs/compound/2026-09-04-backlogit-sizing-is-wit-gated.md`). Size is structured.

| Work item | Estimate |
|---|---|
| Role Boundary grant (ADD-1) | ~8 min |
| Step 6.1(a1) parent-completion step (ADD-2), incl. §5.1's ordered **S0–S5** selector + §5.2 idempotent resume | ~18 min |
| C1 reload extension | ~4 min |
| C2 safe-close summary correction | ~4 min |
| C3 + C4 + C5 generic delegation (incl. §4.1's result-token allowlist and fail-closed HALT rules) | ~12 min |
| C6 delegation bullet | ~4 min |
| **C7** Step 0.5 item 6 pre-mode summary reconciliation (**rev 4**) | ~4 min |
| **C8** Step 6.1(a) pre-mode summary reconciliation (**rev 4**) | ~4 min |
| Go table-driven test (**18** rows — **4 `compound +/−`**, 1 `compound ++`, 2 negative) + ≤4 helpers | ~29 min |
| H0→H1, `go vet`, `go test`, `markdownlint` | ~10 min |
| **Total** | **97 min ≈ 1.62 h** |

**Arithmetic check (rev 6, finding C-1): `8 + 18 + 4 + 4 + 12 + 4 + 4 + 4 + 29 + 10 = 97`.**
The rows sum **exactly** to the stated total. Rates are rev 6 §10.1's own, so this is directly
comparable to the measurement that produced `MUST_REPLAN`.
**Margin to the 2-hour rule: 23 min.**

> **Rev-6 delta: 0 min of new work; the total changes only because the ARITHMETIC was wrong.**
>
> **The rev-5 table never summed to its stated total.** Its rows added to **97**, not the
> pinned **~99** — a two-minute discrepancy that AC-23 and §13.6 then propagated as if it
> were measured. That is the C-1 finding, and it is fixed by **recomputing the total from the
> rows**, never by adjusting a row to protect a headline number (Hardening 7).
> **97 was always the real figure**; rev 6 simply states it.
>
> **Every rev-6 change is a re-wording or a re-pinning at equal cost — no row moves:**
>
> | Rev-6 change | Surface effect | Delta |
> |---|---|---|
> | B-1/B-2 declared-status rebasing (§4.4, §5.0, §5.1 S0, rows 14/15) | same sentences, different predicate — and **simpler** to write: a1 reads one frontmatter field instead of reasoning over location + provenance + allowlist for every member | **0** |
> | B-3 allowlist contract (§4.1, C3–C5) | **reshapes** the trust-boundary text rev 5 already required under AC-25 | **0** |
> | B-3 row 9 becomes `compound +/−`; C-2 row 7 becomes ABSENT-only | **one needle added, one removed**; the compound count holds at **4** | **0** |
> | B-4 four identity fields (Hardening 9a, §9.4a) | **Stage pre-claim** + **S2 read-only** obligations — outside the implementation task, like every other §9.4a read | **0** |
> | B-5 pre-mode de-circularization (§5.1) | **removes** a dependency; writes no new contract text into `_ship.agent.md` | **0** |
> | B-6 verdict contract (§9.3, §9.4b) | plan-level honesty corrections; the `_ship.agent.md` surface is the delegation block already priced in C3–C5 | **0** |
> | B-8 S2 MATCH-set ordering (§9.1) | an **S2 session** obligation against the installed protocol; not `_ship.agent.md` surface | **0** |
>
> **No row was reduced to buy margin, and no site was dropped from the inventory.** The
> margin improves from a *claimed* ~21 to a *computed* **23** solely because the base was
> corrected. **Helper count unchanged at ≤4** and function count unchanged at **1**, so Ship
> Step 4.3's one-function constraint still holds.

*(Rev-2 delta: ADD-2 **+5** for the §5.1/§5.2 applicability and idempotence text; Go test
**+2** for rows 12–13. Total ~71 → ~78 min, margin ~49 → ~42 min.)*

*(Rev-3 delta: **none**. The A-only simplification changes no surface, no clause site and
no test row, so the budget stood at **~78 min** with a **~42 min** margin. Rev 2's
comparative remark that A was *"the second-most comfortable of the three"* is withdrawn —
there is no longer a set of three to compare against.)*

*(**Rev-4 delta: +14 min, recalculated rather than preserved.** C7 **+4** and C8 **+4** —
the two pre-mode summary sites rev 3 excluded (§4.0). ADD-2 **+3** for §5.1 case (iv), the
multiplicity-first selector. Go test **+3** for rows 14–16. Total ~78 → **~92 min**; margin
~42 → **~28 min**. **Helper count unchanged at ≤4** and the function count unchanged at 1,
so Ship Step 4.3's one-function constraint still holds.)*

> **The margin is real and it is now COMPUTED, not asserted.** **23 min** of headroom against
> the 2-hour rule, and §12's contingency is unchanged: if actual work exceeds 2 h
> mid-execution, **HALT and return to Stage**. Rev 3 protected a ~42 min margin by leaving two
> defects out of the inventory; rev 4 spent 14 of those minutes fixing them, rev 5 spent 5
> more making the resulting assertions **actually falsifiable**, and **rev 6 spends none** —
> it corrects the arithmetic and re-points needles at equal cost. Trading margin for
> completeness is the correct direction; **mis-adding the column is not a way to buy it
> back**, which is why C-1 is fixed by recomputation.

> **Rev-5 delta: +5 min (as recomputed in rev 6). Total ~92 → 97 min; margin ~28 → 23 min.**
> Go test **+3** for rows 17–18 and for making rows 7, 10, 14 and 15 compound with
> verbatim-pinned literals (§8.1.1) — the rev-4 versions were vacuously satisfiable, so this
> is the cost of the assertions being real. ADD-2 **+2** for the §5.1 ordered S0–S5 selector
> and the §5.0 pre-archived-descendant gate. **Rev 5's own note claimed +7 and a ~99 total;
> both were arithmetic errors (finding C-1) — the rows it actually wrote sum to 97.**
> **Helper count unchanged at ≤4** and function count unchanged at **1**, so Ship Step 4.3's
> one-function constraint still holds.

> **Not charged to the task budget: Probe 25 (§9.4).** Probe 25 is run by **Stage**, at §9
> step **0a**, **pre-claim** — it is a Stage planning-time measurement against a disposable
> workspace, outside the implementation task's 2-hour envelope. **It has already been
> executed and committed** (`PROBE25_RESULT=PASS`). S2's §9.4a obligation is a **read-only
> re-verification** of that committed evidence — **six** read operations, ~4 min of S2 session
> time, and likewise not part of the harnessed task. Neither consumes the **23 min** margin.

*(Rev-1 review: +1 min for the C1 coherence row 11. PO-1a is a Stage/pre-execution read
and is not charged to the task budget.)*

**Contingency.** If actual work exceeds 2 h mid-execution: **HALT and return to Stage.**
Never split, never rush contract prose, never mis-size.

## Plan Hardening

**Blast radius.** One agent instruction file governing closure for every shipment. A
defect here affects all future closures, not just this chain. Mitigated by: the change
**removes** authority duplication rather than adding authority; four negative test rows;
and A closing by the pre-existing path, so a defect in the new wording cannot corrupt A's
own closure.

**Hardening 1 — the grant is a genuine authority increase.** Ship gains a power it did
not have. It is bounded by five conjunctive conditions, one position (Step 6.1(a1)), one
scope (the active shipment's manifest), and one transition (`active -> done`). It cannot
archive, reparent, create or delete.

> **Rev 7 — the residual containment claim is CORRECTED (finding: Hardening 1 overclaim).**
> Rev 6 asserted the residual — *a Ship defect completing a feature whose descendants are not
> all done* — *"is caught by pre-mode's `expected_status: done` sweep and by P-015's
> protected-set gate, both downstream and both fail-closed."* **That is only true of a
> descendant that would be classified against `expected_status`.** It is **false for exactly
> the case the grant exists to permit**: a descendant declaring `status: archived` is accepted
> by pre-mode **as `pre-archived`, WITHOUT a status check**, so pre-mode **cannot** catch an
> unfinished pre-archived descendant. Claiming a downstream gate catches what it structurally
> does not inspect is the kind of false assurance this plan exists to remove.
>
> **The accurate statement of what contains the residual:**
>
> | Descendant shape | Caught downstream by pre-mode? | What actually contains it |
> |---|---|---|
> | declares a **non-`done`, non-`archived`** status | **YES** — `status-mismatch` ⇒ no `PROCEED` | pre-mode's `expected_status: done` sweep, fail-closed |
> | declares **`status: archived`** (the exempt case) | **NO** — accepted without a status check | **§5.0's five conjunctive gates**, and decisively **condition 5**: the exact recorded Stage disposition naming that ID. Absent it ⇒ **S0 HALT** |
> | declares **`status: done`** | n/a — genuinely complete | nothing to contain |
>
> **Condition 5 is therefore the real containment for the exempt case, not pre-mode.**
> P-015's protected-set gate remains a genuine downstream check on *what the cascade
> archives*, but it is **not** a completion check and was never evidence that descendants were
> done. Both claims are now scoped to what they actually cover.

**Rev 2 adds the two cases rev 1 left unstated**: a manifest with **zero** feature members
is an explicit **no-op** (§5.1 case i), and a feature member already `done` is an
**idempotent resume** (§5.2). Neither widens the grant — the no-op performs no transition
at all, and idempotent resume recognises the terminal state of the *same* transition
without relaxing conditions 1–4. Both are asserted negatively-by-construction in Go test
rows 12–13.

**Hardening 2 — delegation must not become a blank cheque.** §4.1's wording binds Ship to
the verdict **named by the classification**, and P-015 remains the sole **statically
reviewed** authorization surface while the installed skill is the **runtime** authority.

> **Rev 7 — row 9's obligation is stated correctly (plan-review finding E-4).** Rev 6 said
> *"Go test row 9 fails if `_ship.agent.md` ever re-enumerates verdicts"* — which **directly
> contradicted AC-25**, whose whole content is that naming the installed **close-verdict
> tokens** as an output-validation allowlist is **permitted**. The two could not both be
> satisfied by any honest implementation. **Row 9 fails on a restated classification
> PREDICATE, never on the presence of a result token.** Its ABSENT literal is a predicate
> (the per-member/whole-manifest fallback rule at L819); its PRESENT literal asserts
> fail-closed output validation. **Listing `CASCADE` and `SAFE_CLOSE` does not fail row 9;
> restating *when* either is selected does.**

Ship cannot authorize `TASK_ONLY_FINALIZE` — and after rev 3 **nothing currently authorizes
it at all**: the policy branch (B) and its classifier (C) are deferred, so the selectable
**close-verdict** set stays exactly `CASCADE`, `SAFE_CLOSE` (with `HALT` and
`RECONCILE_FAIL` as non-close procedural outcomes, and `BLOCKED` not installed at all —
§4.1 rev 7). Go test row 8 independently asserts `_ship.agent.md` never names the
`TASK_ONLY_FINALIZE` token.

**Hardening 3 — the self-hosting boundary is the highest-risk step.** The S1→S2 handoff is
where an impatient implementer would "just finish it". Three independent defenses: §6 is
normative contract text; AC-9 explicitly fails if one session does both; the Go test
asserts the rule is present in the merged file. The checkpoint `resume_hint` names
`021-S`, `022-F`, `022.001-T` and the merge SHA so S2 can validate rather than trust.
**Rev 2 specifies S2's side of that handoff in full (§9.1)** — rev 1's *"restore the
checkpoint"* was exactly the kind of one-line instruction that gets satisfied by a
plausible-looking shortcut. Auto-picking a sole candidate, filtering the enumeration by
status, or resolving before a confirmed resume would each defeat the property the handoff
exists to establish.

**Hardening 4 — R-1 was a real possibility and is now MEASURED, not merely planned for.**
Rev 1 correctly refused to assume claim leaves a covering feature `active`, and made PO-1a
a pre-execution read. Rev 2 **executed** it: Probe 22 ARM AB measured
`ARM_AB_FEATURE_STATUS_AFTER_CLAIM=active` on the exact `[feature, task]` shape, and
measured that the engine accepts the subsequent `active -> done` move. The expensive
discovery has been made cheaply and has come back **favourable**, so A proceeds. **PO-1b
is retained anyway** — a probe measures the engine, not this particular record at that
particular moment — and both non-`active`, non-`done` branches still halt and return to
Stage rather than widening the grant in-flight.

**Hardening 5 — engine coupling.** A performs **no** engine call that depends on the
1.10.1 digest beyond what every closure already does. The digest binding
(`1E106F5F…959A98`) was **C's** obligation and is **deferred with C** — A never needed it.
A's cascade invocation is the same call every
fully-covered-root closure already makes today. **Rev 3**: this is now strictly simpler —
with C deferred, **no** digest-bound engine invocation is introduced on any current route.

> **Rev 5 correction — A-3 is now CLOSED, and the reconciliation is stable.** §9.4's gate
> does bind a digest, and that remains true. What changed is **where the gate sits** and
> **how the engine is identified**:
>
> * **A's own execution path** introduces no digest-bound engine call — the paragraph above
>   remains true of the *implementation* and of the `CASCADE` invocation itself.
> * **A's verification path** is digest-gated, so a measurement cannot be silently
>   attributed to a different engine build. The gate now runs **pre-claim** (Stage, §9 step
>   0a), so a mismatch costs a re-plan, not a bricked shipment.
> * **The machine-specific binding is REMOVED.** The contract resolves `backlogit` at runtime
>   from the **registered command** and compares **SHA-256 to the committed digest**. The
>   absolute path `C:\Tools\backlogit.exe` is **no longer a requirement of any runtime check**
>   — the evidence merely *records* the observed path as telemetry
>   (`ENGINE_PATH_OBSERVED`). A second operator, a CI runner, or a relocated toolchain
>   therefore passes, provided the binary is the same build.
> * **A mismatch HALTs**, and that HALT is now recoverable: pre-claim it returns to Stage;
>   at S2 re-verification it follows the §9.4a operator recovery path. The rev-4
>   no-override clause that made a mismatch unrecoverable is **withdrawn**.
>
> **Residual, recorded**: a point-release upgrade of backlogit still requires Stage to re-run
> and re-commit Probe 25 so the digest matches. That is a cheap, in-role, pre-claim action —
> not a blocker, and no longer a way to strand A.

**Hardening 6 — what could make this plan wrong.** If `expected_status: done` is not
actually enforced over feature members, the a1 gate is unnecessary (harmless but
pointless). If claim already sets features `active`, PO-1 passes trivially. Neither
falsifies the delegation fix, which stands on §2.2 alone. The plan is therefore
**falsifiable in both directions** and its two deliveries are **independently
justified**.

**Hardening 7 — the inventory itself was a blast-radius surface (rev 4).** Rev 3 closed
the inventory at 6 sites to protect a budget, and in doing so reproduced — inside the
plan's own inventory — the exact duplication-drift failure mode §2.2 exists to eliminate:
one instance of the defect was known and excluded, a second (C7, L255–256) was never
looked for. The generalized lesson is recorded as a constraint: **an inventory is closed by
exhaustive search over the surface, never by a budget target.** When the recalculated
budget does not fit, the response is HALT-and-return-to-Stage (§10.3), never silent
exclusion. AC-23 pins the recalculated numbers so a future revision cannot quietly
re-shrink them.

**Hardening 8 — totality of the a1 selector (rev 4).** A guard that is not **total** over
its input is not a guard. Rev 2 made a1 total over *state* (cases i/ii/iii + §5.2
idempotence); rev 4 makes it total over *multiplicity* (case iv). The residual risk of
case (iv) is nil-by-construction — it performs no mutation and evaluates no condition — and
its cost is one halt on a manifest shape no current shipment has. The asymmetry is
deliberate: the cost of the branch existing is a recorded halt; the cost of it not existing
is a feature completed by manifest ordering, or a bulk feature-lifecycle power the §5 grant
explicitly refuses.

**Hardening 9 — the fixture gate converts a confidence residual into a blocking
measurement (rev 4; **RE-STATED in rev 6, plan-review finding B-4**).** Rev 3 classified the
unproven root-included cascade as *"a confidence residual, not a safety one"* and deferred it
past A. That classification was defensible in isolation and wrong in **sequence**: A is an
**irreversible contract change** (a permanent Role Boundary authority grant) whose sole
release value is making `017-S` closable, so deferring the measurement past A's close means
discovering a mis-classification only once it is expensive to act on. §9.4 moves it
**pre-claim**, **before** A's merge and **before** a1's mutation, so a FAIL costs nothing but
a halt.

> **Rev 6 — THE NO-OVERRIDE LANGUAGE IS DELETED (finding B-4).** Rev 5 rewrote §9.4a and
> §10.2 to **provide** an operator recovery path, but this item still read *"The gate
> deliberately carries **no override path**"* — a direct self-contradiction, and the exact
> clause whose withdrawal was the substance of closing A-1. **It is removed, not softened.**
>
> **The gate remains NON-BYPASSABLE — that is a different property from having no override.**
> Non-bypassable means **no actor may close A without a PASS**. Recoverable means **a FAIL
> has a named in-role exit**. Rev 4 conflated them and produced a deadlock; rev 6 keeps the
> first and provides the second. The authoritative recovery paths are §9.4a's
> **class-appropriate** options (rev 7, E-8) and §10.2's `git revert`. **There is no
> `--force`, no "proceed with a recorded residual", and no self-authorized waiver** — every
> exit is an explicit operator action.

Its placement — pre-claim, riding the pre-existing `017-S → 021-S` edge — means it adds
**no shipment, no backlog item and no dependency edge**, so it cannot itself perturb the
topology it is protecting.

**Hardening 9a — the four evidence identity fields are DISTINCT and UNAMBIGUOUS (rev 6,
finding B-4).** Rev 5 used a single overloaded name, `evidence_sha`, for two different
things — *the commit that **introduced** the evidence* (§9.4a check 3) and *the content
compared against `HEAD`* (check 5). A post-merge commit could therefore **replace** the
evidence while the introducing commit remained a valid ancestor, and both checks would still
pass. Four separately-named fields replace it; **no field is ever used for another's job**:

| Field | Definition | Verified by |
|---|---|---|
| `evidence_commit_sha` | the **LAST** commit that modified **either** evidence file (`git log -1 --format=%H -- {both paths}`) — **not** the introducing commit | Stage at §9 step 0a; **re-verified read-only** by S2 at 12a |
| `fixture_script_sha256` | content identity of `probe25-root-included-cascade-fixture.ps1`, established by comparing the **blob read from `evidence_commit_sha`** to the current working-tree file | Stage; re-verified by S2 |
| `fixture_result_sha256` | content identity of `probe25-root-included-cascade-fixture.txt`, established the same way | Stage; re-verified by S2 |
| `backlogit_executable_sha256` | **expected** value **read from the committed transcript** (`ENGINE_SHA256_EXPECTED`), compared to the SHA-256 of the `backlogit` executable **resolved at run time via the registered command**. **No absolute path is a requirement of any check** | Stage; re-resolved and re-verified by S2 |

> **Rev 7 — THE DERIVATION IS NON-SELF-REFERENTIAL (normative).** Rev 6's definitions were
> circular in two places, and a circular identity check is not a check at all:
>
> 1. **`fixture_result_sha256` must NOT be embedded inside the file it hashes.** The `.txt`
>    transcript **cannot contain its own SHA-256** — writing the digest changes the digest.
>    The field is therefore **a derived comparison, not a stored literal in the transcript**:
>    it is computed by Stage, recorded in **this plan, the session memory artifact and the
>    checkpoint**, and re-derived by S2. **No step reads it out of the `.txt`.**
> 2. **The comparison basis is the COMMITTED BLOB, not a value the artifact asserts about
>    itself.** Both file identities are established by reading the blob **as committed at
>    `evidence_commit_sha`** and comparing it to the current working-tree file.
>
> **The executable digest is the one field that IS read from the transcript — and that is
> correct**, because it is an assertion about an **external** object (the engine), not about
> the transcript itself. The transcript records `ENGINE_SHA256_EXPECTED`; S2 reads that
> expected value **from the committed transcript** and compares it to the **currently
> resolved** command. No self-reference arises.
>
> **Executable derivation, exactly (reproducible, and run this session):**
>
> ```text
> evidence_commit_sha  := git log -1 --format=%H -- <ps1> <txt>
> committed_blob(f)    := git rev-parse {evidence_commit_sha}:{f}
> current_blob(f)      := git hash-object -- {f}
> identity holds       <=> committed_blob(f) == current_blob(f)   for BOTH files
> engine_expected      := ENGINE_SHA256_EXPECTED read from the committed <txt>
> engine_current       := SHA-256 of (Get-Command backlogit).Source
> engine identity      <=> engine_expected == engine_current
> ```
>
> **Measured this session at the current tree:** `evidence_commit_sha = e36d853`; both
> committed blobs **MATCH** their working-tree files (`81d85454d920`, `334cf3be007c`);
> `git diff --quiet HEAD` over both paths is **clean**; `ENGINE_SHA256_EXPECTED` =
> `1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98` **equals** the
> currently-resolved engine digest. **Git blob-OID comparison is used as the byte-exact
> basis** because a text-mode extract-then-hash round trip can differ on line endings
> without any real drift — an artifact of the *method*, not of the evidence.

**Using the LAST commit — not the introducing one — is what closes the replacement hole.**
If evidence is replaced after the merge, `evidence_commit_sha` **moves to that new commit**,
which is a **descendant** of the merge SHA, so the ancestry check **fails**. Under rev 5's
"introducing commit" reading it would have passed.

**Who verifies what, and when:**

* **Stage — pre-claim gate (§9 step 0a).** Stage computes all four fields, confirms the two
  evidence files are **tracked and committed** (not merely present in the working tree), and
  confirms `evidence_commit_sha` is an ancestor of the branch tip. **`021-S` is not claimable
  until this holds.**
* **Ship S2 — post-merge re-verification (§9.4a).** S2 **re-reads** the same four fields and
  re-confirms them. **Every operation is a read** (`git log`, `git merge-base`,
  `git diff --quiet`, `Get-FileHash`, text inspection). S2 authors nothing, modifies nothing,
  re-runs nothing (P-010).

**Post-merge mismatch HALTs active A — with named operator recovery, and NO rollback
fiction.** If any of the four mismatches at S2: `021-S` **stays `active`**, **no** a1
mutation occurs, nothing is shipped or archived, and S2 HALTs to the operator naming the
field that failed. The operator's recovery options are exactly §9.4a's **class-appropriate**
paths (rev 7, E-8): an in-place read-only re-verification for **transient toolchain drift**
(check 4), and `git revert` per §10.2 for a **committed-evidence/hash/ancestry** failure
(checks 1/2/3/5/6), which cannot be cleared in place. **No automatic
rollback, no automatic revert, and no "restore and continue" is claimed** — because none
exists. The only state to unwind is the merge itself, and unwinding it is the explicit
operator `git revert` of §10.2, subject to its `017-S` hold obligation.

## 13. Independent re-review scope (rev 7)

**All prior review scopes are SUPERSEDED**, including rev 6's §13 and every
three-shipment-era scope. Review **rev 7 of this plan only**;
the deferred B/C plans are **not** in scope and must not be routed from.

> **Rev 7 scope note.** Items **1–19** below are carried forward from rev 6 and remain in
> scope — they cover constructs rev 7 did not touch. **Items 20–27 are the rev-7 additions**
> and target exactly the adjudicated `E-*` remediations. **Item 14 and item 15 are
> RE-SCOPED** because rev 7 changed the answers they ask about.

| # | Item | Why it is in scope |
|---|---|---|
| 1 | **Adjudicate the §9.4/§9.4a HYBRID.** Is Stage-authors / Ship-verifies-read-only the correct resolution of A-1 (post-merge no-override deadlock) and A-2 (P-010)? Does §9.4a's **six-check** re-verification remain genuinely non-bypassable while being purely read-only? Is the recovery path complete and in-role? *(Rev 7: corrected from "five-check" — C-8 added the sixth, `F-6`.)* | The structural change of rev 5; closes the only P0 |
| 2 | **Verify the committed Probe-25 evidence.** Does the transcript actually establish the 8 criteria on the exact 13-member root-included shape with `archived_status: queued`? Is the two-set gate evaluated over the engine's own `archived_ids`, not vacuously? | The plan's central empirical claim |
| 3 | **Audit §5.1's S0–S5 ordering.** Is it genuinely total *and* disjoint over (`n`, status)? Can any input still reach two steps? Is S0 correctly placed before `n` is computed? | Closes A-6; rev-4's table was ambiguous |
| 4 | **Audit §5.0's pre-archived exemption.** Is the five-gate definition fail-closed? Is the 11-ID enumeration correct for `017-S`? Is the general authority narrow enough that it cannot be read as "archived descendants don't count"? | Closes A-10 — the completion-proof gap |
| 5 | **Adjudicate §8.1.1's pinned literals.** Are the **five** ABSENT literals (rows 7/9/10/14/15) verbatim-correct and **singly-occurring** against the installed file? Are rows 14/15's rev-6 **declared-status** PRESENT phrases reachable only by declared-status wording? Is row 7's ABSENT-only form non-vacuous (C-2)? Can any row still be green at H0 under the corrected readiness guard (C-3)? | Closes A-7/B-1/B-2/C-2/C-3; rows 14–15 are the whole justification for the inventory reopening |
| 6 | **Adjudicate the recalculated budget** (**97 min, 23 min margin, rows summing exactly**) and the ≤4-helper / 1-function claim at 18 rows | The rev-6 `MUST_REPLAN` verdict was a budget verdict; this is the same axis, and C-1 was an arithmetic defect |
| 7 | **Audit §4.1's delegation trust boundary.** Does pinning the closed verdict set in Ship's own file contradict §4.1's "do not enumerate verdicts" constraint? Is the P-015/skill-disagreement HALT correct? | Closes A-11 without re-introducing the duplication A removes |
| 8 | **Audit §9.1's S2 candidate rule and payload.** Is `session_id` top-level correct per the installed contract? Is overriding the "zero candidates is normal startup" continuation safe and correctly scoped to S2 only? | Closes A-8/A-9 |
| 9 | **Verify §9.4b's residual classification.** Is detect-after-mutate correctly attributed to inherited P-015 rather than to A? Is the plan free of prevention/rollback overclaims? | Closes A-12 honestly rather than by assertion |
| 10 | **Audit the `## Constitution Check`.** Are D1–D4 genuine justifications rather than waivers? Is any principle mis-assessed? | Closes A-13 |
| 11 | **Confirm A still authorizes no new verdict** and that the §6 authority-escalation property survives rev 5's larger surface | Widening guard (rows 7–10); core self-hosting property |
| 12 | **Adjudicate the §R16 correction** in `docs/plans/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-plan.md` — is R16.3 still the *sole operative* manifest contract, with withdrawn text clearly non-governing? | Governing-contract coherence; a contradiction there re-blocks `017-S`. **Read-only coherence check** (C-8) |
| 13 | **Confirm the operator-only PR boundary survives**, and that the fail-closed stance on the contradictory Orchestrator Step 1.5 (3a/3e) is correctly recorded **without** modifying the Orchestrator in this cycle | P-010/role-separation boundary. **Read-only coherence check** (C-8) |
| **14** | **Adjudicate the DECLARED-STATUS re-basing AND ITS REV-7 RE-SCOPING (rev 6 B-1/B-2; rev 7 E-1/E-3).** Is any **lifecycle class** anywhere in §4.4/§5/§5.0/§5.1/§9 still determined by file location? Is §5.0 condition 2's split — **general `queue` XOR `archive` containment** for every member, **archive-placement as a separate routing/integrity condition** only after declared `status: archived` — genuinely free of lifecycle inference? Is the declared-status rule now correctly scoped to **this plan's own gates** rather than asserted of the installed pre-mode API? | **The attempt-2 P0 and its twin P1, plus rev 7's E-1/E-3 correction** |
| **15** | **Adjudicate the DELEGATION CONTRACT as rev 7 states it (rev 6 B-3; rev 7 E-4/E-5/E-6).** Are AC-5, AC-25 and row 9 mutually satisfiable now that row 9 fails on **predicates** and not on **result tokens**? Is the close-verdict allowlist (`CASCADE`, `SAFE_CLOSE`) correctly distinguished from the non-close procedural outcomes (`HALT`, `RECONCILE_FAIL`) and from the uninstalled `BLOCKED`? Is removing the runtime `policy_id`/version and P-015-disagreement checks correct, given the installed classifier emits no such fields and AC-5 forbids re-running predicates? | Rev 6 traded one contradiction for three; rev 7 must not have traded them for an ambiguity |
| **20** | **(rev 7, E-6) Adjudicate WHERE "expected `CASCADE`" is seated.** Is the split — general ID-free delegation in `_ship.agent.md`, release-instance expectation in the plan/backlog/`PA-*` records — genuinely free of the §4.1-vs-AC-28 contradiction? **Confirm `_ship.agent.md` carries NO global ID-specific close logic** (no `021-S`, `017-S`, `018.*`, no expected-verdict rule). Is the halt enumeration complete and fail-closed? | The sharpest attempt-3 contradiction; the fix must not relocate it |
| **21** | **(rev 7, E-2) Adjudicate C7/C8 as POINTER-ONLY.** Do the §8.1.1 rows 14/15 PRESENT needles assert **delegation** rather than any restated predicate? Are they **unsatisfiable** by a declared-status or locational sentence? Are the lock, orphan scan, `record-consistent` conjunct and the `Scope note` demonstrably preserved by rows 17/18? Are the ABSENT literals still unique and the PRESENT literals still absent at the current tree? | Rows 14/15 are the whole justification for reopening the inventory — twice now re-pinned |
| **22** | **(rev 7, E-7) Adjudicate the post-CASCADE restore coverage.** Do §9.4b's table and §10.3 **both** enumerate all four post-mutation failure modes — non-empty `returned_ids`, **each** two-set difference, and an altered/cleared `parent_id` at step 4 — and does **each** carry the identical mandatory **verified** restore? Is the unverifiable-restore path fail-closed? | Rev 6 left the reparenting failure mode with mutated state and no restore obligation |
| **23** | **(rev 7, E-8) Adjudicate the recovery split.** Is the transient-drift / committed-evidence division **evidentiary and correct**? Does "re-verify and continue" now appear **only** for check 4? Is the claim that checks 1/2/3/5/6 cannot be cleared in place accurate, and consistent across §9.4a, AC-20a and §10.3? | Rev 6's option 1 and option 3/AC-20a were flatly contradictory |
| **24** | **(rev 7, item 7) Adjudicate the evidence-identity derivation.** Is it genuinely **non-self-referential** — `evidence_commit_sha` from `git log -1` over both files, committed blobs compared to current files, engine digest read from the committed transcript and compared to the resolved command? Is **`fixture_result_sha256` demonstrably NOT embedded in the file it hashes**? Are the **nine STEP 11 aggregate criterion names** correct and complete against the committed transcript (closing `F-7`)? | A circular identity check is not a check |
| **25** | **(rev 7, item 8) Adjudicate the checkpoint tool surface.** Is `consumer_id` correctly described as **access identity, not a filter**, with the CLI `--agent` flag correctly forbidden? Does the anomaly gate read the **actual top-level `needs_quarantine`/`quarantined`/`total`** fields? Are exact filename + `agent` + `session_id` matching and unrelated-candidate tolerance **preserved unchanged**? | `--agent ship` would silently disable AC-29's warnings |
| **26** | **(rev 7, item 10) Audit the STRICT-SAFETY approved-action records.** Do `PA-021-CASCADE` and `PA-017-CASCADE` name the **exact current** manifests and topology? Are `ActionRisk: destructive` / `ActionResult: approved` / the operator timestamp / the five–six conditions / the restore-verify-halt rule all present? Is the **"authorizes only these two exact actions, not admin fallback"** boundary unambiguous? Is `PA-017-CASCADE` correctly **void while `021-S` is unshipped**? | Principle VII was previously mapped to machine gates only (`G-2`) |
| **27** | **(rev 7, item 10) Audit the completed Constitution items.** Does Principle **VI** now assess **dependency discipline** (stdlib-only, no testify, no new edges) and not merely cohesion? Is **D5 Width-Isolation** — executable Markdown contract + its Go verification as **one** domain — a genuine justification rather than a waiver? Does **D4** name a real **rejected simpler alternative** and honestly decline the freeze-scope framing? | Closes `D-3`, `D-4`, `F-11`, `G-3` |
| **16** | **Adjudicate Hardening 9a's four identity fields (rev 6, B-4).** Does `evidence_commit_sha` = *last* commit genuinely close the post-merge replacement hole that the *introducing* commit left open? Are the two fixture hashes separable and both checked? Is the no-override language fully gone, with **non-bypassable** and **recoverable** kept distinct and no rollback fiction? Is the C-7 caveat on recovery option 2 accurate? | Closes the last self-contradiction rev 5 carried |
| **17** | **Adjudicate a1's ordering and self-sufficiency (rev 6, B-5).** Does a1 genuinely run before pre-mode with **no** back-edge? Is its guard truly a narrow completion precondition rather than a second close-path classification? Is the authoritative pre-mode/P-015 selection left solely authoritative for what it owns? | The circular dependency had no defined provider on the `017-S` route |
| **18** | **Adjudicate the no-fallback verdict contract (rev 6, B-6, AC-28).** Is "EXPECT `CASCADE`; any other verdict HALTs to Stage" correct against P-015 and the installed skill for these two exact manifests? Is the reachability admission complete and free of prevention/rollback overclaims? Is §9.4b's HALT/restore table consistent with §10.3? | Rev 5's `SAFE_CLOSE`-at-no-data-cost claim was wrong in all three parts |
| **19** | **Adjudicate the S2 MATCH-set ordering (rev 6, B-8, AC-29).** Does anomaly-gate-over-full-enumeration → filter → exactly-one-MATCH preserve fail-closed safety while no longer stranding on expected leftovers? Is the `session_id`-in-summary claim correct against the installed tool? | The rev-5 rule bricked S2 on a state the protocol deliberately produces |

**Out of scope, explicitly**: the deferred B/C plans; `shipment-reconcile/SKILL.md`
(including its stale `src/autoharness/...` classifier paths — recorded as a follow-up, not
a rev-5 edit); Orchestrator implementation repair (deferred to `019.004-T`); the
pre-existing `SAFE_CLOSE`/exit-9 tool conflict, which remains off-route and recorded; and
the P-015 detect-after-mutate ordering (§9.4b), which is an inherited residual recorded as a
follow-up against P-015 and the skill, not against A.

## Plan Review

<!-- plan-review-attempt: 1 -->

```text
dispatch_mode: multi-agent
decision: FAIL
```

- **Reviewed revision**: rev 4
- **Reviewed at commit**: `d19b954` (working tree, pre-commit)
- **Gate run**: attempt 1
- **Plan hardening required**: **yes** (declared in the plan header). **Present**: yes — a
  `## Plan Hardening` section exists with 9 hardening items. **Materially complete**: yes
  for the authority grant; **NO** for rev 4's new constructs (see A-3).
- **Verdict**: **FAIL** — gate rule *"Any P0 or P1 findings ⇒ FAIL"*. The plan MUST be
  revised before `harvest`. **No plan status advance**: `status:` stays `planned`.

### Dispatch capability (P-012)

| Capability | Status |
|---|---|
| reviewer-subagent-dispatch | `TOOL_OK` |
| model-specific-review-routing | `TOOL_OK` — anchor route `model_routing.anchor_review` = `openai` / `gpt-5.6-sol` / `high` dispatched |
| indexed knowledge retrieval (`agent-engram`) | `TOOL_OK` — `engram stats` reachable, 297 symbols, 100% embedding coverage |
| `agent-native-parity-reviewer` identity file | `TOOL_DEGRADED: agent-native-parity-reviewer-identity — declared fallback: rubric applied via dispatched general-purpose subagent on a cross-model route.` The persona was **covered, not skipped**; `.github/agents/subagents/agent-native-parity-reviewer.agent.md` is not installed in this workspace |

`dispatch_mode: multi-agent` is correct: subagent dispatch **and** model-specific routing
were both available and used. The single missing identity file is recorded as a declared
degradation of that persona's *identity source*, not of the dispatch surface.

### Persona coverage

| Persona | Trigger | Mode | Model/route | Findings (P0/P1/P2/P3) |
|---|---|---|---|---|
| Constitution Reviewer | always-on | subagent | `claude-opus-4.8` | 1 / 3 / 4 / 3 |
| Go Reviewer | always-on | subagent | `claude-opus-4.8` | 0 / 1 / 5 / 4 |
| Scope Boundary Auditor | always-on | subagent | `claude-opus-4.8` | 0 / 1 / 1 / 5 |
| Learnings Researcher | always-on | subagent | default | 1 / 3 / 8 / 3 |
| Architecture Strategist | always (cross-model) | subagent — **ANCHOR** | `gpt-5.6-sol` / high | 0 / 7 / 3 / 1 |
| Security Lens Reviewer | triggered — the plan is an **authorization** change | subagent (cross-model) | `gpt-5.5` / high | 0 / 4 / 1 / 0 |
| Agent-Native Parity Reviewer | triggered — MCP tools + agent-facing contract | subagent (cross-model), **degraded identity** | `grok-4.6` / high | 0 / 3 / 4 / 3 |

All seven selected personas were covered. No persona was skipped.

### Gate decision rationale

Rev 4 fixed the five blockers it set out to fix, and three independent personas
**affirmatively cleared** its most contested moves: the C7/C8 inventory reopening is
correct and the line anchors were verified against the installed file; the §9.1/AC-9
recovery expansion is **faithful** to the installed Prune-on-Restore protocol; and the
§12 budget arithmetic sums exactly to **92 min / 28 min margin**.

The plan nonetheless **FAILS**, because rev 4's *new* constructs — principally §9.4/AC-20
and §5.1 case (iv) — introduced defects of their own, and because three pre-existing
issues were surfaced that prior cycles missed. **Four separate personas independently
converged on §9.4/AC-20**, which is the strongest signal in this review.

### P0 findings — blocking

**A-1 · §9.4/AC-20 is a no-override gate placed AFTER the irreversible merge → unrecoverable stall.**
*(Constitution Reviewer P0-1; independently corroborated by Architecture Strategist, Scope
Boundary Auditor, and the P-010 finding A-2.)*
Sections: §9.4 "Failure semantics", §9 step 12a, AC-20, §10.3.
A's Role-Boundary change **merges to `main` at §9 step 5**, *before* S2 runs the fixture at
step 12a. So on FAIL the contract change is **already deployed to every future Ship
session**, while `021-S` is stuck `active`/in-flight — and §9.4 states there is *"no
operator override … no `--force`, and no 'proceed with a recorded residual' path."* No
in-role actor can clear that state: Stage may not close shipments (P-010), and Ship is
halted by the gate. It also **directly contradicts the plan's own §10.2 rollback**, which
requires an operator `git revert` — precisely the override §9.4 forbids.
**Recommendation (two viable resolutions; operator adjudication required):**
**(R1)** Move the fixture to a **Stage pre-execution gate** (the PO-1a/Probe-22 slot, §10.1)
so a FAIL costs nothing and nothing has merged; or **(R2)** keep an S2-side gate but make it
a **pre-merge** CI/quality gate on A's branch, and add an explicit operator-escalation path
(abandon `021-S` + operator-approved revert per §10.2). A **hybrid** preserves the operator's
stated intent at lowest risk: **Stage runs and commits Probe 25 pre-claim**, and **S2
re-verifies the committed evidence (presence + digest + PASS tokens) read-only** before A may
be marked/archived `shipped` — still non-bypassable, but with no deadlock and no P-010 issue.

### P1 findings — blocking

**A-2 · §9.4/§12/AC-20 assign Probe-25 authoring *and commit* to Ship (S2) → P-010 violation.**
*(Constitution Reviewer P1-2; Architecture Strategist.)* `docs/plans/evidence/…/probe25-*`
are plan/spike evidence artifacts under `docs/plans/`. Ship's Role Boundary **forbids**
creating or modifying deliberation, spike, plan, or review artifacts. Executed literally,
S2 records a P-010 violation and halts at step 12a — **A can never close**. The plan even
calls it *"a Stage/closure-time measurement"* (§12) while assigning it to Ship. Note the
plan's own precedent is the opposite: Probe 22 (PO-1a) is explicitly *"Stage / §9 step 0."*
**Fix**: resolved together with A-1 by the hybrid — Stage authors/commits, S2 verifies.

**A-3 · §9.4's digest binding is machine-specific and contradicts Hardening 5.**
*(Constitution Reviewer P1-3; Agent-Native Parity P3-10.)* A **non-bypassable, no-override**
gate is pinned to `C:\Tools\backlogit.exe` and one exact SHA-256. Any toolchain drift, CI
runner, or second operator yields `DIGEST_GATE=FAIL` → HALT → per A-1, no override → A
bricked. It also flatly contradicts **Hardening 5**, which asserts *"A performs **no** engine
call that depends on the 1.10.1 digest … A never needed it."*
**Fix**: replace the absolute path with workspace tool discovery; treat the digest as
recorded telemetry with an explicit operator-acknowledged mismatch path rather than a
no-override hard FAIL; and reconcile Hardening 5 with §9.4 — one of the two is wrong.

**A-4 · §9.4 PASS criterion 5 does not match the authoritative cascade API.**
*(Architecture Strategist — VERIFIED against P-015 amendment 1.21.0.)* §9.4 says
`required_ids` and `allowed_ids` must reconcile with *"no required/returned mismatch."*
The real gate is **two separately-labelled, independently-failing** conditions:
`archived_ids − allowed_ids` (unexpected artifact archived) and `required_ids − archived_ids`
(required artifact not archived). `returned_ids` is a **separate** step-2 check. A literal
Probe-25 implementation could satisfy criterion 5 while asserting neither real invariant.
**Fix**: restate criterion 5 as those two exact set equations and keep `returned_ids == []`
as its own criterion.

**A-5 · §5.1 case (iv) narrows a shape P-015 explicitly authorizes.**
*(Architecture Strategist — VERIFIED.)* P-015 item 1 is *"quantified over **every feature
member of the manifest**"* and item 4 says *"the qualifying root feature **member(s)**"* —
**plural is authorized by policy**. An unconditional `n > 1` HALT makes every legitimate
multi-root manifest permanently unclosable by Ship, which **falsifies the plan's repeated
claim that A "narrows nothing"** and that its blast radius is harmless.
**Note**: case (iv) itself was an explicit operator directive and is *safe* (it mutates
nothing). The defect is the **undisclosed narrowing**, not the halt.
**Fix**: either make a1 set-based (validate every feature member against the same five
conditions and transition all of them), or **keep the halt but disclose it honestly** as an
accepted narrowing with a recorded follow-up, and delete the "narrows nothing" claims.

**A-6 · a1 gives two different answers for an already-`done` feature member.**
*(Agent-Native Parity P1-2.)* Case (iii) fires on *"any §5 condition unmet"*, and condition 5
is *"live status is exactly `active`."* An already-`done` member **fails (iii) → HALT** and is
**simultaneously** told by §5.2 to **proceed via idempotent resume**. This is exactly the
dual-action class of defect rev 4 fixed for `n > 1`, left unfixed one row above it.
**Fix**: order the selector explicitly — compute `n`; then branch on status: `done` +
conditions 1–4 ⇒ §5.2 resume; `active` + all five ⇒ (ii); anything else ⇒ (iii). Do not nest
§5.2 inside (ii).

**A-7 · Go test rows 14–15 are vacuously satisfiable — the C7/C8 verification is hollow.**
*(Go Reviewer P1-1.)* §8.1 says rows 14–15 *"reuse the same substring-assertion helper"*, i.e.
a whole-file `strings.Contains(content, "pre-archived")`. That passes if the token appears
**anywhere** in the ~94 KB file while L255–256 and L776–777 keep their exact queue-only text —
and it may be **green at H0**, never transitioning red→green. It also contradicts §4.4's own
*"Neither sentence may be left in the queue-only form."* Since rows 14–15 (AC-22) are the
**entire mechanical justification** for reopening the inventory and spending +14 min, this
would let rev 4 go all-green with its headline defect intact.
**Fix**: make each row **compound** — assert PRESENT a long distinctive phrase only the
reconciled wording can contain, **AND** assert ABSENT that site's exact old literal. Pin all
four literals verbatim in §8.1.

**A-8 · Checkpoint payload omits the required top-level `session_id`.**
*(Agent-Native Parity P1-1.)* §9 step 8 and AC-9 enumerate `schema_version`, `agent`, `phase`,
`resume_hint` and `context` — but **not** `session_id`, which the installed Checkpoint Payload
Contract requires top-level. An agent following the plan rather than the overlay writes a
non-conforming checkpoint, which the §9.1 anomaly gate would then reject.
**Fix**: add `session_id` to §9 step 8 and AC-9; permit MCP `backlogit_create_checkpoint`
**or** CLI `backlogit checkpoint create --state-dump`.

**A-9 · S2 with zero checkpoint candidates is ambiguous — continue vs fail closed.**
*(Agent-Native Parity P1-3.)* The installed protocol treats zero active `ship` candidates as
**normal startup — explicitly not a failure**. But S2's entire purpose depends on S1's
checkpoint existing. As written, an agent may legitimately read "zero candidates" as
"proceed with fresh work", skipping parent completion.
**Fix**: specialize in §9.1 — for S2, zero active `ship` candidates is **FAIL CLOSED**, not
the ZERO-CANDIDATE continuation.

**A-10 · `pre-archived` descendant exemption can satisfy a1 without proving completion.**
*(Security Lens Reviewer.)* §5 condition 1 checks only *non-pre-archived* descendants for
`done`, and §4.4 defines `pre-archived` as **archive-located, accepted without a status
check**. An unfinished descendant that is merely archive-located therefore stops counting for
condition 1, is not live for condition 2, and can keep an intact `parent_id` for condition 3 —
letting Ship complete a covering feature whose work is not done. Missing/malformed
`artifact_type` also has no fail-closed rule before `n` is computed.
**Fix**: define "pre-archived executable descendant" for the **grant** as a *proven terminal*
descendant (`status: done` terminal relocation, or `status: archived` + valid
`archived_status`/`archived_from`), per the skill's own `archived-completed(done)` role. Treat
duplicate queue+archive records, archive-located `queued`/`active`, malformed provenance, and
missing `artifact_type` as a1 HALT conditions, evaluated **before** computing `n`.

**A-11 · Delegation (C3–C6) hands authorization to an unversioned skill boundary.**
*(Security Lens Reviewer; Architecture Strategist P2.)* Ship must invoke *"the close path named
by the machine-checkable classification"* and may not enumerate verdicts — but the implementing
surface, `shipment-reconcile/SKILL.md`, is **outside this plan's 2-file scope and outside its
review**. A future change there could name a new destructive verdict and Ship would obey with
no local provenance check. The plan also never says what Ship does when P-015 and the skill
**disagree**, or when the classifier returns nothing/an unknown path.
**Fix**: keep verdict names out of Ship, but require the classifier result to carry
`policy_id: P-015` + a policy version/commit, and require Ship to **HALT** on an absent,
unknown, ambiguous, or policy-mismatched verdict.

**A-12 · CASCADE guardrails detect some destructive mutations only *after* they occur.**
*(Security Lens Reviewer.)* `returned_ids`/two-set/`parent_id` checks run **after** the cascade
has already mutated the backlog; a non-empty `returned_ids` HALT does not explicitly restore
`.backlogit/queue/` and `.backlogit/archive/`. Worst case for `017-S` is the 13-member subtree
requeued/detached with `parent_id` cleared before the halt fires.
**Fix**: require a clean `.backlogit/` baseline + restorable snapshot before any CASCADE, and
make **every** cascade HALT path run an unconditional `git restore -- .backlogit/queue/
.backlogit/archive/` and re-verify before halting.

**A-13 · Missing mandatory `## Constitution Check` section.**
*(Constitution Reviewer P1-1.)* The constitution's Governance section requires **every**
implementation plan to include a "Constitution Check" mapping the work against Principles
I–XI and documenting justified violations. The plan maps exhaustively against *workflow
policies* but never against the *constitutional principles*, so its genuinely justified
deviations (task-granularity heuristics — 16 test scenarios vs "fewer than 4"; 1 fn + ≤4
helpers vs "fewer than 5 functions"; `tests/integration/` vs inline `_test.go`) are
undocumented.
**Fix**: add `## Constitution Check` with those deviations and their rationale.

**A-14 · Backlog records are a rev-3 split-brain contract.**
*(Architecture Strategist; Scope Boundary Auditor P3.)* `022.001-T` still states *"Plan rev 3"*,
*"6 SITES"*, *"13 ROWS"*, *"~78 min … ~42 min margin"*, *"AC-1..AC-19"*; `022-F` echoes rev 3.
Ship consumes **both** the queue record and the plan, so it would execute the rev-3 contract
and silently omit C7, C8, case (iv), rows 14–16 and AC-20–AC-23.
**Fix**: Stage resyncs both records to rev 4 **before** A is claimed. *(Remediated in this
session — see the Remediation Cycle 1 note below.)*

**A-15 · §R16 of the governing policy-gap plan is internally contradictory.**
*(Architecture Strategist.)* R16.2 declares the 13-member fully-covered root current, while
R16.3/R16.3.1/R16.3.1a/R16.3.2/R16.3.3 still operatively assert the manifest *"now contains 12
entries and is task-only"*, that the 13-entry `CASCADE` shape is *"WITHDRAWN"* and *"unreachable
by construction"*, and that *"neither is remediable."* None of that sits in the historical
appendix, and §R16 is declared *"the whole governing contract."* This directly contradicts the
route Plan A depends on.
**Fix**: demote the withdrawn analysis to clearly non-governing historical labels and leave one
unambiguous operative statement of the 13-member CASCADE route. *(Remediated in this session.)*

**A-16 · Rev-4 header claimed an executed `plan-review` gate (§14) that did not exist.**
*(Constitution P2-2; Scope Boundary P3; Architecture Strategist.)* The rev-4 banner asserted
*"Formal plan-review gate executed (§14) … now appended"* while the file ended at §13.
**Fix**: this section is that gate. The banner claim is corrected in the same pass.

### P2 findings — advisory

| # | Finding | Persona |
|---|---|---|
| B-1 | §3/AC-11 *"exactly 2 files changed"* is contradicted by AC-20's mandated `probe25-*.{ps1,txt}` commit. Qualify as "2 implementation files + N closure-evidence files" | Constitution, Scope Boundary |
| B-2 | Row 3's ordering anchors (`"Step 6.1(a1)"`) **do not exist** in `_ship.agent.md` — the file uses `a0.`/`a.`/`b.` list markers. A naive `strings.Index("a.")` matches the first `a.` in a 94 KB doc. Pin unique anchors (`TOPOLOGY_GATE: lifecycle`, `Pre-archive reconciliation gate (mandatory)`, and a new unique a1 heading) | Go Reviewer |
| B-3 | Negative needles (rows 7–10) are unpinned; a bare `children` needle can **never** go green and would permanently block H1. Pin the full drifted clause and the full `backlogit move <shipment_id> --status shipped` literal | Go Reviewer |
| B-4 | *"H0 red 1 of 1"* is not automatic — a 16-row table yields many reds. Requires a single marker-fatal guard **before** the row loop | Go Reviewer |
| B-5 | `technology-go.instructions.md` says "use testify", but testify is **not** in `go.mod`; importing it breaks AC-11's 2-file surface. Mandate stdlib assertions explicitly | Go Reviewer |
| B-6 | ~24 min for 16 rows is optimistic while **zero** of the 16 needles/anchors are pinned; the derivation churn can consume the ~28 min margin | Go Reviewer |
| B-7 | §9.4 lacks machine checks for **live-backlog isolation** (prove `FIXTURE_BACKLOG_ROOT` ≠ / not nested under live `.backlogit`; before/after `git status` on live `.backlogit`) and for **evidence redaction** (absolute user paths, emails, tokens in the committed `.txt`) | Security Lens |
| B-8 | a1's descendant traversal duplicates graph-walking already owned by P-015 + `shipment-reconcile` — recreating, for mutation eligibility, the same duplication pattern C4 removes for verdict selection | Architecture Strategist |
| B-9 | The a1 mutation occurs **before** pre-mode acquires the single-writer lock, and the ~20–30 min fixture sits between `a0` and the close, so topology can go stale inside the closure critical section (TOCTOU) | Architecture Strategist |
| B-10 | PO-1b mandates CLI `backlogit get` with no justification — `backlogit_get_item` reads by ID too. "Read by status, never by path" is right; CLI-exclusivity is not | Agent-Native Parity |
| B-11 | §4.4 drops two authoritative PROCEED conditions: **no orphans** and `record-consistent`. C8's sentence also carries the lock acquisition and orphan scan, which a careless rewrite would delete | Agent-Native Parity |
| B-12 | The complexity degradation claim is accurate but incomplete — the plan never **forbids** the advertised-but-failing `backlogit update --complexity` call (exit 1) | Agent-Native Parity |
| B-13 | a1 does not pin the move tool; specify `backlogit_move_item` / `backlogit move {id} --status done` only, never a file edit | Agent-Native Parity |
| B-14 | C7's replacement must coexist with the existing `Scope note (139-F/139.001-T)` at L257+, which scopes item 6 to *uniform-status* intake — `017-S`'s intake manifest is **not** uniform-status | Scope Boundary |
| B-15 | §6's self-hosting property is enforced only by contract text + a text-presence test; nothing mechanically stops S1 continuing. Honestly disclosed, but worth explicit operator sign-off at the S1→S2 boundary | Constitution, Security Lens, Architecture Strategist |
| B-16 | Task-granularity heuristics (16 scenarios vs "<4"; 5 functions vs "<5") breached without documented justification — ties to A-13 | Constitution |
| B-17 | The Go test is text-only and provides **no** executable blast-radius coverage for task-only/partial-feature/multi-feature manifests on the shared closure path | Architecture Strategist |

### P3 findings — noted

| # | Finding | Persona |
|---|---|---|
| C-1 | Principle VIII: a global-closure blast radius is never paired with an explicit careful/freeze-scope safety-mode directive | Constitution |
| C-2 | H0's `not implemented:` marker risks the stub anti-pattern; prefer running the 16 real assertions and emitting the marker from the first failure, with `t.Run` per row for traceability | Constitution, Go Reviewer |
| C-3 | `_contract_test.go` in `tests/integration/` deviates from the instructions' `tests/contract/` taxonomy (justified by precedent; no `tests/contract/` exists) | Constitution, Go Reviewer |
| C-4 | Helper taxonomy is imprecise — only row 3 is positional; rows 11–13/16 are multi-substring presence checks. Budget conclusion (≤4) still holds | Go Reviewer |
| C-5 | Reuse the existing package-level `repoRoot(t)` from `build_script_test.go`; redefining it is a duplicate-declaration compile failure at H0 | Go Reviewer |
| C-6 | Keep needles single-line — the target file may be CRLF on Windows | Go Reviewer |
| C-7 | Case (iv)'s justifying prose ("what Stage produces if a grouping session ever…") is speculative; rest the case on **totality** alone | Scope Boundary |
| C-8 | §13 items 10–11 reach into two adjacent governing documents; keep them strictly read-only coherence checks | Scope Boundary |
| C-9 | AC-9 names only the MCP create; the CLI create is equally valid. §9 step 17 names CLI `backlogit sync`; MCP is `backlogit_sync_index` | Agent-Native Parity |

### Affirmative clearances (recorded as evidence, not findings)

* **C7/C8 reopening is correct.** Both anchors verified in the installed file: L255–256
  reads *"present in `.backlogit/queue/` with the expected status"*; L776–777 reads
  *"present in queue with `status: done`"*. Neither contains `pre-archived`. Both are inside
  A's declared file #1, so the **file surface does not widen**. Rev 3's budget-driven
  exclusion was itself the scope error. *(Scope Boundary, Agent-Native Parity, Go Reviewer.)*
* **§9.1/AC-9 expansion is faithful.** The installed `backlogit.instructions.md`
  Prune-on-Restore protocol contains verbatim the `restore → prune/gate → resume` ordering,
  the three-class never-prune allowlist, the Engram-unreachable fail-closed rule, and the
  `cleanup_checkpoints` exclusion. `agent-engram` **is** installed, so these are mandatory,
  not a no-op. Nothing was invented. *(Scope Boundary, Agent-Native Parity.)*
* **Budget arithmetic is correct**: 8+18+4+4+12+4+4+4+24+10 = **92**; 120−92 = **28**.
  *(Scope Boundary.)*
* **All C1–C8 line anchors verified** against the installed file: ADD-1 L38; C7 255–256;
  C1 756–764; C8 776–777; C2 789–791; C3 794–795; C4 802–823; C5 821–823. *(Architecture
  Strategist.)*
* **The §2.2 drift is real**: Ship's copy checks direct `children`; P-015 requires
  descendants at every depth and says a direct-children check *"is insufficient"*.
  *(Architecture Strategist.)*
* **Compound library confirms** the exit-9 shipment-move refusal, the shipment lifecycle
  triple, and the `complexity`-not-settable degradation. *(Learnings Researcher.)*
* **No P-016 violation**: S1→S2 is sequential on one branch; §9.1 forbids a new worktree;
  the probe workspace is a scratch data dir, not a git worktree. *(Constitution.)*
* **A authorizes no new verdict**; the selectable set remains `CASCADE`, `SAFE_CLOSE`,
  `HALT`. *(Constitution, Architecture Strategist.)*

### Runtime verification and operational closure gaps

* **Runtime verification**: the only executable artifact is a **text-contract** test. No
  runtime behavior of the closure path is exercised by A itself (B-17). §9.4 was rev 4's
  attempt to close this gap and is itself the P0 (A-1).
* **Operational closure**: §9 steps 16–18 correctly place `operational-closure
  mode=post-merge`, P-020 compaction, `backlogit sync`, and the closure PR on the closure
  branch before push. **No gap.**
* **Rollback**: §10.2 is sound for a 2-file revert but is **contradicted by §9.4's
  no-override clause** (A-1) and does not cover the AC-20 evidence files (B-1).

### Required before re-review (P0/P1 remediation queue)

A-1 and A-2 are **jointly decisive** and require **operator adjudication** of the fixture
placement (R1 / R2 / hybrid) before any further remediation of §9.4 is meaningful.
A-3 → A-13 are remediable in-plan. A-14, A-15 and A-16 were remediated in this session.

## Plan Review — Remediation Cycle 1 disposition

This session is **remediation cycle 1** of the attempt-1 `FAIL`. Dispositions:

| Finding | Disposition |
|---|---|
| **A-14** | **CLOSED** — `022-F` and `022.001-T` resynced to rev 4 (8 sites, 16 rows, case (iv), ~92 min, AC-1…AC-23) |
| **A-15** | **CLOSED** — §R16 corrected; the withdrawn 12-entry/task-only/unreachable analysis is demoted to explicitly non-governing historical labels |
| **A-16** | **CLOSED** — this `## Plan Review` section is the executed gate; the rev-4 banner claim is corrected |
| **A-4** | **CLOSED** — §9.4 criterion 5 restated as the two independent set equations |
| **A-5** | **CLOSED (disclosure)** — the "narrows nothing" claim is withdrawn; case (iv)'s narrowing of P-015's authorized multi-root shape is now disclosed and recorded as a follow-up. The operator-directed halt itself is retained |
| **A-3** | **PARTIAL** — the Hardening 5 / §9.4 digest contradiction is reconciled in-plan; the machine-specific path binding remains pending A-1 adjudication |
| **A-1, A-2** | **OPEN — operator adjudication required.** The gate is retained exactly as directed; both findings are recorded unmodified |
| **A-6 … A-13, B-*, C-*** | **OPEN** — queued for remediation cycle 2, which **requires a fresh `plan-review` (attempt 2)** |

**The gate verdict for this plan remains `decision: FAIL`.** Remediation cycle 1 does not
convert a FAIL into a PASS: rev 5 must be re-reviewed. `status:` stays `planned`, the plan is
**not** harvest-ready, and `021-S` must **not** be claimed.

## Plan Review — attempt 2

<!-- plan-review-attempt: 2 -->

```text
dispatch_mode: multi-agent
decision: FAIL
```

- **Reviewed revision**: rev 5
- **Reviewed at commit**: `4fd21c5` (working tree, pre-commit)
- **Gate run**: attempt 2 (remediation cycle 2 of max 3)
- **Plan hardening required**: **yes** (declared in the header). **Present**: yes — `## Plan
  Hardening` with 9 items. **Materially complete**: **NO** — Hardening 9 still asserts the
  withdrawn no-override stance (see B-4 below).
- **Verdict**: **FAIL** — gate rule *"Any P0 or P1 findings ⇒ FAIL"*. `status:` stays
  `planned`. **`021-S` must NOT be claimed.**

### Dispatch capability (P-012)

| Capability | Status |
|---|---|
| reviewer-subagent-dispatch | `TOOL_OK` |
| model-specific-review-routing | `TOOL_OK` — anchor route `model_routing.anchor_review` = `openai` / `gpt-5.6-sol` / `high` dispatched |
| `agent-native-parity-reviewer` identity file | `TOOL_DEGRADED: agent-native-parity-reviewer-identity — declared fallback: rubric applied via dispatched subagent on a cross-model route.` Persona **covered, not skipped** |

### Persona coverage

| Persona | Mode | Model/route | P0/P1/P2/P3 |
|---|---|---|---|
| Constitution Reviewer | subagent | `claude-opus-4.8` | 0 / 1 / 2 / 2 |
| Go Reviewer | subagent | `claude-opus-4.8` | 0 / 0 / 2 / 2 |
| Scope Boundary Auditor | subagent | `claude-opus-4.8` | 0 / 0 / 1 / 3 |
| Learnings Researcher | subagent | default | 1 / 2 / 2 / 2 |
| Architecture Strategist | subagent — **ANCHOR** | `gpt-5.6-sol` / high | 1 / 4 / 3 / 0 |
| Security Lens Reviewer | subagent (cross-model) | `gpt-5.5` / high | 0 / 0 / 0 / 0 |
| Agent-Native Parity Reviewer | subagent (cross-model), degraded identity | `grok-4.6` / high | 0 / 1 / 1 / 3 |

All seven selected personas were covered. No persona was skipped.

### Gate decision rationale

**Rev 5 closed every attempt-1 blocker it targeted.** A-1, A-2, A-3, A-6 (partially),
A-7, A-8, A-9, A-10, A-11, A-12 and A-13 are affirmatively cleared below by the personas
that originally filed them — the **Security Lens Reviewer returned a clean sheet (0/0/0/0)**,
explicitly clearing all four of its attempt-1 P1s.

The plan nonetheless **FAILS**, on a defect **rev 5 itself introduced**. Two personas —
the **anchor** and the **Learnings Researcher** — independently converged on the same P0,
and **the plan's own committed Probe-25 transcript proves it**. That convergence, with
mechanical evidence, is the strongest signal in this review.

### P0 findings — blocking

**B-1 · §5.1's new S0 anomaly gate HALTs on A's own manifest — `021-S` becomes unclosable.**
*(Architecture Strategist (anchor) P0; independently corroborated by Learnings Researcher
P0-1; **confirmed by the plan's own §9.4 evidence**.)*
Sections: §5.0 gate 2, §5.1 **S0**, §9 steps 4 and 13.
S0 HALTs on *"an archive-located member that is not on the §5.0 recorded Stage disposition
list"* — a list containing only the eleven `018.*` IDs. But §9 step 4 moves **`022.001-T ->
done`** in S1, and the live `registry.yaml` routes `done|accepted|rejected|archived` to
`archive/`. By the time a1 runs at step 13, **`022.001-T` is archive-located and is not on
the list** ⇒ **S0 HALTs before `n` is computed**, `022-F` never reaches `done`, pre-mode
never returns `PROCEED`, and `021-S` stalls `active` — **after** A's irreversible Role
Boundary change merged at step 5. A second S0 bullet fires independently: a registry
relocation on `done` is not a close-archive, so `archived_status`/`archived_from` are absent
⇒ *"malformed archive provenance"* ⇒ HALT.
**This reintroduces the attempt-1 A-1 unrecoverable-stall shape through a different door**,
and it falsifies §5.0's claim that *"`021-S` has NO pre-archived descendants"* — true today,
**false at the moment a1 actually runs**.
**The plan's own evidence proves it**: `probe25-…txt` records `MEMBERS_QUEUE_LOCATED=0` and
`001.004-T loc=archive status=done`.
**Fix**: Re-base §5.0 gate 2 and every S0 archive-location trigger on **declared `status`**,
never on file location — the authoritative rule the `shipment-reconcile` skill already
states (*"never inferred from … which of `queue/`/`archive/` currently holds the record"*)
and which `docs/compound/2026-05-07-backlogit-shipment-status-constraints.md` records. Apply
the §5.0 list and provenance gates **only** to records declaring `status: archived`; accept
a uniquely-located record declaring `status: done` as a normal completed member.

### P1 findings — blocking

**B-2 · §4.4's classification table is location-first and contradicts the plan's own measurement.**
*(Learnings Researcher P1-1.)* §4.4 defines `matched` as *"**Queue file present** AND declared
status matches"*. The committed transcript reports, in one block, `MEMBERS_QUEUE_LOCATED=0`
**and** `PREMODE_MATCHED=2 PREMODE_PRE_ARCHIVED=11`. Under §4.4 those cannot both hold (zero
queue-located ⇒ `MATCHED=0`, `PRE_ARCHIVED=13`). §4.4 is the **normative source for C7/C8**,
and §8.1.1 pins row 15's PRESENT literal to that locational wording — **encoding the error
permanently into Ship's instruction surface** and letting rows 14–15 go green on wording that
is still false. That is exactly the *"new inaccuracy replacing the old one"* §4.4's own rev-5
note warns against.
**Fix**: Restate §4.4 from **declared status**; reconcile the transcript narrative; re-pin the
row 14/15 PRESENT literals.

**B-3 · §4.1/AC-25 require `_ship.agent.md` to enumerate the verdict set that AC-5/row 9 require it NOT to enumerate.**
*(Architecture Strategist P1; corroborated by Scope Boundary F-4 (P3) and Go P2-1.)*
AC-5 and Go test row 9 assert the file *"does **not** enumerate the authorized verdict set"*;
rev-5's AC-25 requires it to *"bind Ship to the **closed** verdict set `{CASCADE, SAFE_CLOSE,
HALT}`"*. **Both cannot pass under an honest test.** It also conflates `HALT` — a control
outcome with no close mutation — with the two close-path classification values.
**Fix**: Keep the no-enumeration rule in Ship's file and express the trust boundary as
**behavior** (HALT on any verdict the classification did not name; HALT on absent/ambiguous/
policy-mismatched results; HALT on P-015/skill disagreement), holding the closed set in
**this plan** as a review-time constraint. Pin row 9's needle to re-derivation phrasing, not
bare verdict tokens.

**B-4 · Hardening 9 still asserts the withdrawn no-override stance; `evidence_sha` is ambiguous.**
*(Architecture Strategist P1.)* `## Plan Hardening` item 9 still reads *"The gate deliberately
carries **no override path**"* — directly contradicting §9.4a and §10.2, which rev 5 rewrote
to **provide** one. Separately, §9.4a check 3 says the commit that *"introduced"* the evidence
must pre-date the merge, while check 5 compares current files only to `HEAD`: a post-merge
commit could **replace** the evidence while the introducing commit remains an ancestor.
**Fix**: Delete/supersede Hardening 9's no-override language. Define `evidence_sha` as the
**last** commit affecting both evidence files and require current blobs to equal that commit's.

**B-5 · §5.0's exemption has a circular dependency on pre-mode.**
*(Architecture Strategist P1.)* §5.0 gate 1 requires *"P-015 pre-mode classifies it
`pre-archived`"*, taken from the authoritative scan rather than re-derived — but §9 runs a1 at
step **13** and `shipment-reconcile mode: pre` at step **14**. Pre-mode cannot have run, and an
earlier intake scan does not satisfy the same-snapshot/immutability requirement. The `017-S`
route has **no defined provider** for a load-bearing a1 input.
**Fix**: Remove the cycle — define the exemption from the Stage disposition plus explicit
declared-status/location/provenance reads, reserving the authoritative pre-mode call for its
existing post-a1 gate.

**B-6 · The cascade reachability claim is still an overclaim, and §9.3/§9.4b/§10.3 disagree.**
*(Architecture Strategist P1.)* §9.4b's *"A neither creates nor removes it"* and *"not a new
cascade path"* are true at the **implementation-definition** level but not at the
**reachability** level: A is precisely what makes the destructive path reachable for `017-S`.
Meanwhile §9.3 still claims a HALT *"falls back to `SAFE_CLOSE`"* at *"no data cost"*, §9.4b
disclaims rollback, and §10.3 prescribes `git restore` — and **P-015 forbids
CASCADE→SAFE_CLOSE substitution once CASCADE is selected**.
**Fix**: State that A introduces new **reachability** to an inherited destructive path and
inherits its detect-after-mutate risk; delete the `SAFE_CLOSE`-fallback and "no data cost"
claims; reconcile §9.4b with §10.3 on which HALT paths carry a mandatory verified restore.

**B-7 · Backlog records `022-F` / `022.001-T` are a rev-4 split-brain against rev 5.**
*(Constitution Reviewer P1; corroborated by Learnings Researcher P1-2.)* Both records carry a
`=== REV 4 RECONCILIATION … THIS BLOCK WINS ===` block stating **16 rows**, **~92 min**,
**AC-1..AC-23**, a **four-way** selector, and *"AC-20 = the non-bypassable root-included
13-member cascade fixture gate"* — i.e. **the rev-4 S2-run post-merge gate that was the A-1
P0**. Rev 5 is **18 rows / ~99 min / AC-1..AC-26 / S0–S5 / Stage-pre-claim hybrid**. Ship
consumes both the plan and the queue record; a status flip without a content resync would
make Ship execute the superseded rev-4 contract and **re-instantiate A-1 and A-2**.
*(Learnings reported the records as "still rev 3"; that is inaccurate — the superseding rev-4
block is present. The split-brain conclusion stands.)*
**Fix**: Stage resyncs both records to rev 5 **before** `021-S` is claimable.

**B-8 · S2 step 2a's `>1` rule can strand a valid S1 checkpoint.**
*(Agent-Native Parity P1.)* Step 2a FAIL-CLOSES on **any** `>1` active `ship` candidate
**before** the identity filter, so one leftover unresolved record plus a valid S1 checkpoint
bricks S2 — and the installed protocol deliberately leaves fail-closed handoffs `active`
regardless of age, so leftovers are expected. Also, `session_id` is required for matching but
is not among the fields `backlogit_list_checkpoints` presents.
**Fix**: Filter to filename+`session_id` **first**; FAIL CLOSED only if the **matching** set
≠ 1; treat non-matching leftovers as a warning. Verify `session_id` after
`backlogit_get_checkpoint`, or document that list returns it.

### P2 findings — advisory

| # | Finding | Persona |
|---|---|---|
| C-1 | §12's table sums to **97**, not the pinned ~99 — the ADD-2 row was never raised 18→20 and still says "four-case" (rev-4 wording). AC-23/§13.6 pin 99/21 | Scope Boundary, Go |
| C-2 | Row 7's PRESENT needle (`descendants — at every depth…`) appears in **no** pinned replacement text; satisfying it forces re-adding the restatement §4.1 forbids. Its ABSENT needle is already non-vacuous alone | Go |
| C-3 | §8.1.1's marker guard is described with **inverted logic** ("fails if marker absent" + "at H1 the marker is gone ⇒ rows execute"). Taken literally the rows never execute and H1 never greens. Fails safe, but B-4 is not cleanly closed | Go |
| C-4 | Probe 25's `required_ids` omits the shipment record, which P-015 makes an unconditional member; criterion 5b could pass if the engine archived it but omitted it from `archived_ids`. Also: 8 criteria but **9** emitted tokens (5a/5b) | Architecture |
| C-5 | §5.0 hardcodes 11 `017-S` IDs into a global agent contract, coupling reusable policy to one release instance | Architecture, Scope Boundary |
| C-6 | Constitution Check's Principle VII row claims the cascade is *"never against the live backlog by this plan"* — false; §9 step 15 and §9.3 step 6 both run it live | Constitution |
| C-7 | §9.4a recovery option 2 (Stage re-runs and re-commits post-merge) is **structurally void** — the new evidence commit is a descendant of the merge SHA and fails check 3 | Constitution |
| C-8 | §9.4a's five checks verify **provenance only** and cannot detect the internally inconsistent transcript of B-2; add a read-only consistency check | Learnings |
| C-9 | P-018 Copilot engagement is absent from **both** merge points (§9 steps 5 and 17); the window closes permanently at merge | Learnings |
| C-10 | S0/§5.0's dual-location and provenance checks are not expressible on the registry surface (`get_item` cannot show torn duplicates); pin to `backlogit_doctor` / `backlogit_query_sql` | Agent-Native Parity |

### P3 findings — noted

| # | Finding | Persona |
|---|---|---|
| D-1 | AC-11 still reads literally *"Exactly 2 files changed"* without B-1's recommended qualifier | Scope Boundary |
| D-2 | "4 compound rows" (§12/AC-23) vs §8.1's Kind column marking {14,15,16} compound and {7,10} negative — inconsistent labelling | Scope Boundary, Go |
| D-3 | D4 omits the Governance-required *"simpler alternative that was rejected"*; Principle VIII arguably satisfied via freeze-scope rather than deviated | Constitution |
| D-4 | Principle VI row assesses cohesion but VI's body is dependency discipline; cite the stdlib-only/no-testify fact. Width-Isolation (markdown + Go in one task) unaddressed | Constitution |
| D-5 | Rows 10/14 ABSENT-needle single-occurrence across the 94 KB file not verified | Go |
| D-6 | §9 step 17 names CLI `backlogit sync` only; installed order is MCP `backlogit_sync_index` first | Agent-Native Parity |
| D-7 | S4 states the transition with no tool pin; the pin lives only in the PO-1b box | Agent-Native Parity |
| D-8 | AC-9 says *"all twelve steps"*; the checklist now has 13 items (0, 1, 2, 2a, 3–11) | Agent-Native Parity |
| D-9 | Rev-3 banner calls B/C shipments `blocked`, which is not a valid shipment status (`queued\|active\|shipped\|abandoned`); actual state is `archived`/`archived_status: queued` | Learnings |
| D-10 | B-12 (`backlogit update --complexity` never explicitly forbidden) is claimed closed by the rev-5 banner but is addressed nowhere in §12 | Learnings |

### Affirmative clearances (recorded as evidence, not findings)

* **Security Lens Reviewer returned 0/0/0/0** — all four of its attempt-1 P1s cleared:
  **A-10** (§5.0's five gates are fail-closed; `archived_status: queued` no longer proves
  completion), **A-11** (the closed verdict set + HALT rules prevent a future skill from
  naming a new destructive verdict), **A-12** (inherited-residual classification is honest),
  and **B-7** (isolation measured: `FIXTURE_ISOLATED_FROM_LIVE=True`,
  `LIVE_BACKLOG_UNMUTATED=True`; **no** user-profile paths, emails, tokens or secrets in the
  committed `.txt`). It also cleared the **A-1/A-2 hybrid**, noting the evidence-commit
  ancestry requirement defeats post-merge regeneration/replay.
* **A-1 (P0) RESOLVED.** *(Constitution, Architecture.)* The gate is re-seated pre-claim; a
  FAIL costs a re-plan with nothing merged; the post-merge path has an in-role escape; the
  no-override clause is withdrawn. **The rev-4 deadlock is genuinely broken.**
* **A-2 (P-010) CLOSED.** *(Constitution, Security Lens, Architecture.)* S2's five checks are
  pure reads (`git log`, `git merge-base`, `git diff --quiet`, `Get-FileHash`, text
  inspection); Stage authors and commits. Confirmed in the transcript: `PROBE_RUN_BY=Stage`.
* **A-7 (vacuousness) CLOSED.** *(Go Reviewer — verified against the installed file.)* All
  four pinned ABSENT literals are **verbatim-correct, single-line, and genuinely present** at
  their target sites: row 14 at **L255**, row 15 at **L777**, row 7 at **L810**, row 10 at
  **L790**. No compound row can pass vacuously.
* **B-2/B-3/B-5/C-5/C-6 (attempt 1) CLOSED.** *(Go Reviewer.)* Anchors
  `TOPOLOGY_GATE: lifecycle (before closure/safe-close)` (L788) and
  `Pre-archive reconciliation gate (mandatory)` (L792) both exist and are **unique**;
  needles are full single-line clauses; **testify is absent from `go.mod`** so stdlib is
  correctly mandated; `repoRoot(t)` **is** package-level in `build_script_test.go` and reuse
  avoids a duplicate-declaration compile failure.
* **A-8 CLOSED, A-9 (zero-candidate half) CLOSED, B-10/B-12/B-13 CLOSED, C-9 (create half)
  CLOSED.** *(Agent-Native Parity — verified against the installed Checkpoint Payload
  Contract.)* `session_id` **is** required top-level; both MCP and CLI create are valid; the
  S2-only zero ⇒ FAIL-CLOSED override is correct and correctly scoped.
* **No scope expansion.** *(Scope Boundary.)* All eight rev-5 additions map 1:1 to attempt-1
  findings; **no third implementation file**; `## Constitution Check` ← A-13, §9.4a ← A-1/A-2,
  §5.0 ← A-10, §4.1 ← A-11, §9.4b ← A-12, rows 17–18 ← B-11/B-14. **C-7 and C-8 (attempt 1)
  CLEARED** — speculative case-(iv) prose is gone; §13 items 12–13 are labelled read-only.
* **A-23 helper budget CLEARED** *(Go)* — `repoRoot` (reused) + `mustContain` +
  `mustNotContain` + `indexOfUnique` = ≤4 across 18 rows; one function holds; ~29 min is
  realistic **because** the literals are now pinned.
* **A-13 substantially CLOSED** *(Constitution)* — D1–D3 are genuine justifications, not
  disguised waivers. **P-016: no violation. Principle XI: compliant** (revert, never rewrite;
  no amend). **Principle II honored** (H0 red → H1 green is now a real transition).

### Required before re-review (P0/P1 remediation queue — cycle 3)

**B-1** is decisive and must be fixed first: **re-base every location-derived check on
declared `status`** (§5.0 gate 2, §5.1 S0). **B-2** is the same root cause in §4.4 and
should be fixed in the same pass — together they are one coherent correction. **B-3**
requires choosing one side of the enumeration contract. **B-4**, **B-5**, **B-6** and
**B-8** are localized text corrections. **B-7** is a Stage backlog action, **performed in
this session** (see the cycle-2 disposition below).

**Max re-entry cycles**: 2 remain consumed of 2; **cycle 3 is the last** before the
Escalation Protocol applies.

## Plan Review — Remediation Cycle 2 disposition

This session is **remediation cycle 2**. It closed **every attempt-1 P0/P1** and re-ran the
formal gate as **attempt 2**, which returned **`decision: FAIL`** on defects rev 5
introduced.

| Attempt-1 finding | Disposition |
|---|---|
| **A-1** (P0) | **CLOSED** — hybrid applied: Stage authors/commits Probe 25 pre-claim (§9 step 0a); S2 re-verifies read-only (§9.4a); no-override clause withdrawn; recovery path named |
| **A-2** | **CLOSED** — Ship never authors plan evidence; all five S2 checks are reads |
| **A-3** | **CLOSED** — engine resolved at runtime via registered command; digest compared; path recorded as telemetry only; Hardening 5 reconciled |
| **A-6** | **CLOSED** — §5.1 is a strictly ordered, non-overlapping S0–S5 selector; §5.2 re-seated as S3 |
| **A-7** | **CLOSED** — rows 7/10/14/15 compound with four verbatim-pinned ABSENT literals (§8.1.1), verified against the installed file |
| **A-8** | **CLOSED** — top-level `session_id`; MCP **or** CLI create |
| **A-9** | **CLOSED** (zero-candidate rule); residual `>1` handling re-opened as **B-8** |
| **A-10** | **CLOSED** — §5.0's five gates; `archived_status: queued` never proves completion; 11 IDs enumerated; `021-S` has none |
| **A-11** | **CLOSED** in substance; the *representation* is re-opened as **B-3** |
| **A-12** | **CLOSED** — §9.4b classifies the residual as inherited; reachability wording re-opened as **B-6** |
| **A-13** | **CLOSED** — `## Constitution Check` added with D1–D4 |
| **A-4, A-5, A-14, A-15, A-16** | **CLOSED in cycle 1** (unchanged) |
| **B-2, B-3, B-5, B-7, B-10, B-11, B-12, B-13, B-14, C-5, C-6, C-7, C-8, C-9** | **CLOSED** — same-surface P2/P3 queue absorbed into rev 5 |
| **B-1, B-4, B-6, B-9, B-15, B-16, B-17, C-1, C-2, C-3, C-4** | **DEFERRED with rationale** — see the deferral register below |

**Deferral register (explicit, not silent).**

| ID | Why deferred |
|---|---|
| **B-1** (AC-11 "2 files" vs evidence) | Substantively reconciled (§9.4/§10.2/§12); the literal AC wording remains — re-filed as **D-1** |
| **B-4** ("H0 red 1 of 1") | Addressed by the marker-fatal guard, but the guard's logic is inverted — re-filed as **C-3** |
| **B-6** (24 min optimistic) | Superseded: literals are now pinned, which removes the derivation churn B-6 priced |
| **B-9** (a1/lock TOCTOU) | Inherited P-015 ordering, not A's surface — same class as §9.4b |
| **B-15** (self-hosting rests on process integrity) | Honestly disclosed in AC-9's residual; no mechanical enforcement exists to add within 2 files |
| **B-16** (task-granularity heuristics) | **CLOSED** by `## Constitution Check` D1/D2 |
| **B-17** (no executable blast-radius coverage) | Out of A's 2-file surface; recorded against P-015/skill follow-up |
| **C-1** (Principle VIII safety mode) | **CLOSED** as deviation D4; completeness nit re-filed as **D-3** |
| **C-2** (stub anti-pattern / `t.Run` per row) | **CLOSED** — §8.1.1 mandates `t.Run` per row and real assertions |
| **C-3** (`tests/contract/` taxonomy) | **CLOSED** as deviation D3 |
| **C-4** (helper taxonomy imprecise) | Cosmetic; the ≤4 conclusion holds and was re-verified by the Go Reviewer |

**Stage action taken this cycle (closes B-7):** `022-F` and `022.001-T` are resynced to
**rev 5** and re-banner-ed **NOT CLAIMABLE — attempt 2 FAIL**.

**The gate verdict is `decision: FAIL`.** `status:` stays `planned`. The plan is **not**
harvest-ready. **`021-S` must not be claimed**, and `017-S` behind it stays ineligible.
**No backlog items are created from this plan.**

## Plan Review — attempt 3

<!-- plan-review-attempt: 3 -->

```text
dispatch_mode: multi-agent
decision: FAIL
```

- **Reviewed revision**: rev 6
- **Reviewed at commit**: `f57703d` + working tree (pre-commit)
- **Gate run**: attempt 3 (remediation cycle 3 of max 3 — **the last**)
- **Plan hardening required**: **yes** (declared in the header). **Present**: yes — `## Plan
  Hardening` with 9 items **plus the new Hardening 9a**. **Materially complete**: **yes** —
  the withdrawn no-override language is deleted and the four identity fields are defined.
- **Verdict**: **FAIL** — gate rule *"Any P0 or P1 findings ⇒ FAIL"*. `status:` stays
  `planned`. **`021-S` must NOT be claimed.**

### Dispatch capability (P-012)

| Capability | Status |
|---|---|
| reviewer-subagent-dispatch | `TOOL_OK` |
| model-specific-review-routing | `TOOL_OK` — anchor route `model_routing.anchor_review` = `openai` / `gpt-5.6-sol` / `high` dispatched |
| `agent-native-parity-reviewer` identity file | `TOOL_DEGRADED: agent-native-parity-reviewer-identity — declared fallback: rubric applied via dispatched subagent on a cross-model route.` Persona **covered, not skipped** |

### Persona coverage

| Persona | Mode | Model/route | P0/P1/P2/P3 |
|---|---|---|---|
| Architecture Strategist | subagent — **ANCHOR** | `gpt-5.6-sol` / high | **0 / 6 / 2 / 0** |
| Constitution Reviewer | subagent | `claude-opus-4.8` | 0 / 1 / 1 / 3 |
| Go Reviewer | subagent | `claude-opus-4.8` | **0 / 0 / 0 / 1** |
| Scope Boundary Auditor | subagent | `claude-opus-4.8` | 0 / 1 / 1 / 2 |
| Learnings Researcher | subagent | default | 0 / 2 / 4 / 4 |
| Security Lens Reviewer | subagent (cross-model) | `gpt-5.5` / high | **0 / 0 / 0 / 0** |
| Agent-Native Parity Reviewer | subagent (cross-model), degraded identity | `grok-4.6` / high | 0 / 0 / 3 / 1 |

All seven selected personas were covered. No persona was skipped.
**Totals: P0 = 0 · P1 = 10 · P2 = 11 · P3 = 11.**

### Gate decision rationale

**The attempt-2 P0 (B-1) is GENUINELY CLOSED, and every persona that could adjudicate it
agrees.** The declared-status re-basing is correct, it matches the recorded institutional
knowledge that originally diagnosed the defect
(`docs/compound/2026-05-07-backlogit-shipment-status-constraints.md`: *"Treat 'location in
archive/' and 'declared `status: archived`' as two independent facts"*), and it is
corroborated by the committed Probe-25 transcript. **`022.001-T`, `done`-in-archive at the a1
gate, now passes §5 condition 1 and raises no S0 trigger.** The **Go Reviewer** mechanically
re-verified every pinned literal and returned a clean sheet on its axis; the **Security Lens
Reviewer** returned **0/0/0/0** for the second consecutive attempt.

**The plan nonetheless FAILS on 10 P1s.** They fall into three groups:

1. **The declared-status sweep is INCOMPLETE (anchor P1-1, Learnings P1-1).** §5.0 condition 2
   still requires the record to be in `archive/` specifically, which is stronger than the
   containment-only contract AC-27 asserts; and **§4's clause-inventory table (L222–223) still
   prescribes the rev-4 LOCATIONAL C7/C8 wording** — the exact predicate §4.4 names as the
   root cause. AC-27 and §13 item 14 both enumerate *"§4.4, §5, §5.0, §5.1, §9"* and **omit
   §4**, which is how it survived.
2. **Three new internal contradictions introduced by rev 6's own corrections (anchor
   P1-2/3/4/5/6).** Hardening 2 still fails row 9 on any verdict enumeration, contradicting
   AC-25's permitted allowlist; **AC-28's "HALT on a valid `SAFE_CLOSE`" contradicts §4.1's
   "MUST invoke the close path the classification named"**; AC-25 requires detecting
   P-015/skill disagreement that the installed classifier emits no fields to support;
   §9.4b's restore table omits the cascade's post-mutation `parent_id` check; and §9.4a's
   option 1 (re-verify and continue) contradicts option 3/AC-20a (only revert clears a
   post-merge mismatch).
3. **The backlog-record split-brain has recurred for the THIRD consecutive cycle
   (Constitution P1, Scope P1, Learnings P1-2).** `022-F`/`022.001-T` still carry the rev-5 /
   attempt-2 banner and — critically — `022.001-T`'s most recent C7/C8 statement is the
   **rev-4 LOCATIONAL** one. **This is closed by the Stage action recorded below**; the
   standing fix is recorded as a process change.

### P1 findings — blocking

| # | Finding | Persona | Section |
|---|---|---|---|
| **E-1** | §5.0 condition 2 still makes a lifecycle decision from physical location (requires `archive/` specifically, not `queue/` XOR `archive/`), contradicting §4.4/AC-27's containment-only contract | Architecture (anchor) | §5.0 |
| **E-2** | §4's clause-inventory table still prescribes the rev-4 **locational** C7/C8 replacement wording — the B-1/B-2 root-cause predicate, in the table an implementer actually reads. AC-27 and §13 item 14 omit §4 from the sweep | Learnings | §4 (L222–223) |
| **E-3** | §4.4 promotes the skill's safe-close Step 0(b) snapshot rule (L394–399) into the *pre-mode* API, but the skill's Classification table and pre-mode step 3 are explicitly location-first. Non-load-bearing for these two manifests, but C7/C8 would describe semantics the delegated skill does not have | Architecture (anchor) | §4.4 |
| **E-4** | Hardening 2 still says row 9 fails whenever `_ship.agent.md` re-enumerates verdicts — directly contradicting AC-25's permitted result-token allowlist | Architecture (anchor) | Hardening 2 / AC-25 |
| **E-5** | AC-25 requires detecting `P-015`/skill **disagreement** and a `policy_id`/version match, but the installed classifier emits no such fields and P-015 supplies no independent machine result. Detecting semantic disagreement would force Ship to re-run the predicates — violating AC-5 | Architecture (anchor) | §4.1 / AC-25 |
| **E-6** | **§4.1 and AC-28 prescribe incompatible actions for a valid `SAFE_CLOSE`.** §4.1: Ship MUST invoke the named close path. AC-28: Ship MUST HALT without invoking it. The result is simultaneously mandatory and forbidden | Architecture (anchor) | §4.1 / AC-28 |
| **E-7** | §9.4b's HALT/restore table omits the installed cascade's **step 4 `parent_id` preservation check**, which also fails *after* mutation; §10.3 covers only out-of-`allowed_ids` archival. A valid-ID cascade that clears a `parent_id` halts with mutated state and no mandatory verified restore | Architecture (anchor) | §9.4b / §10.3 |
| **E-8** | §9.4a's recovery contract is internally inconsistent: option 1 allows the same A execution to continue after re-verification, while option 3 / AC-20a claim only `git revert` clears a post-merge mismatch. The committed-evidence-replacement case has no specified abandon/reset/re-merge sequence | Architecture (anchor) | §9.4a |
| **E-9** | `022-F` / `022.001-T` are a rev-5 / attempt-2 split-brain against a rev-6 plan; `022.001-T`'s latest C7/C8 statement is the **rev-4 locational** wording. Third consecutive recurrence (A-14 → B-7 → E-9) | Constitution, Scope, Learnings | backlog records |
| **E-10** | The rev-6 header claims it closes *"the same-surface P2/P3 queue"*, but D-3, D-4, D-7 and D-10 are neither addressed nor dispositioned; no cycle-3 disposition table existed | Scope, Learnings | header / disposition |

### P2 findings — advisory

| # | Finding | Persona |
|---|---|---|
| F-1 | Hardening 9a does not define a durable owner/storage location for `evidence_commit_sha` / the two fixture hashes; the transcript records `PROBE_HEAD` and engine hashes but not those three fields | Architecture |
| F-2 | The reachability disclosure attributes new CASCADE reachability only to `017-S`; **`021-S`'s own manifest also could not reach it before a1** and now invokes the destructive branch at §9 step 15 | Architecture |
| F-3 | Row 16 (the a1/S0 selector guard) is **not** in §8.1.1's pinned-literals table, so a1 could ship with a location-keyed S0 and still pass green — the A-7 vacuousness class reproduced on the headline defect | Scope |
| F-4 | §9.4's "8/8 PASS" does not distinguish the two **probe-simulated** criteria (2 = pre-mode `PROCEED`, 3 = `CASCADE`, both re-derived in PowerShell) from the six **engine-observed** ones | Learnings |
| F-5 | §9 steps 5/17 name P-018 engagement but not the workspace's **verified `[bot]`-suffixed REST fallback** and poll cadence, which the compound entry records as mandatory before merge | Learnings |
| F-6 | Stale **"five checks"** survives in §10.3 and §13 item 1 after C-8 added the sixth | Learnings |
| F-7 | §9.4a check 2's "nine criterion tokens" are unpinned; the transcript body emits `CRITERION_2…8` and the nine-token set exists only in the STEP 11 aggregate. An S2 agent grepping for nine `CRITERION_n` tokens finds eight and HALTs — a fabricated post-merge failure | Learnings |
| F-8 | §9.1 2a.i still passes `consumer_id: "ship"` while claiming unfiltered enumeration; the CLI has `--agent` (not `--consumer_id`), and `--agent ship` would drop the 13 stage records so AC-29's leftover warnings cannot fire | Agent-Native Parity |
| F-9 | §9.1 step 2's "quarantine flag" is not a per-summary field; the live list exposes top-level `needs_quarantine` / `quarantined` / `total` | Agent-Native Parity |
| F-10 | **D-7 NOT FIXED** — §5.1 S4 and §9 step 13 still state `active -> done` with no tool pin; the `backlogit_move_item` pin lives only in §10.1's PO-1b box | Agent-Native Parity |
| F-11 | Constitution Principle VI row assesses cohesion, not dependency discipline; Width-Isolation (markdown + Go in one task) remains undocumented | Constitution |

### P3 findings — noted

| # | Finding | Persona |
|---|---|---|
| G-1 | Row 10's ABSENT needle is a command **fragment**, not the "full drifted clause" §8.1.1's own rule prescribes — verified unique, so it functions, but the rule and the needle disagree | Go |
| G-2 | Principle VII marks "satisfied" via machine gates only; operator-approval routing for the now-reachable destructive cascade is not mapped | Constitution |
| G-3 | Deviation D4 still does not label the rejected simpler alternative as such, nor engage whether VIII is *satisfied via freeze-scope* rather than deviated | Constitution |
| G-4 | C-5's decoupling is prose-only: no negative row asserts `_ship.agent.md` contains no `018.` literal, so the release-instance boundary can silently regress | Scope |
| G-5 | §9.4a check 6 does not tie the topology tokens to the classification arithmetic (`TOPOLOGY_MEMBERS == MANIFEST_COUNT`, etc.) | Learnings |
| G-6 | Dangling duplicated sentence fragment in §9.3 (*"Rev 18 recorded that a"*) | Learnings |
| G-7 | §4's subsections run 4 → 4.0 → 4.4 → 4.1 → 4.2 → 4.3; the load-bearing normative block is out of sequence — the same locality problem that let E-2 survive | Learnings |
| G-8 | `query_sql` cannot show torn queue+archive copies (`items.id` is PK); containment must be `backlogit_doctor --check-duplicates` only | Agent-Native Parity |

### Affirmative clearances (recorded as evidence, not findings)

* **Security Lens Reviewer returned 0/0/0/0 for the second consecutive attempt.** The
  declared-status exemption remains fail-closed; the five §5.0 gates still hold conjunctively;
  `archived_status: queued` still never proves completion; the allowlist contract still
  prevents a future skill naming a new destructive verdict; `evidence_commit_sha` = LAST commit
  genuinely defeats post-merge replacement/replay; the two fixture hashes are separable; the
  MATCH-set reordering does not weaken the fail-closed handoff; the committed `.txt` carries
  **no** secrets, tokens, emails or user-profile paths; and **no hardcoded absolute executable
  path is a runtime requirement of any check**.
* **Go Reviewer returned 0/0/0/1 — a clean mechanical sheet.** Independently re-measured, at
  the current tree rather than at `d19b954`: all five ABSENT literals present **exactly once**
  and single-line (rows 14/L255, 15/L777, 9/L819, 10/L790, 7/L810); row 3's anchors unique at
  L767/L773 with the `a1` anchor correctly **absent** pre-implementation; all four reconciled
  PRESENT phrases **absent today**, so every `compound +/−` row is a real red→green.
  **C-1, C-2, C-3, D-2 and D-5 are mechanically verified FIXED.** `8+18+4+4+12+4+4+4+29+10 =
  97` re-added independently; margin **23 min**. `repoRoot(t)` confirmed package-level in
  `build_script_test.go` and already reused by two other tests; the three new helpers collide
  with nothing; **testify confirmed absent from `go.mod`**.
* **B-1 (P0) RESOLVED — unanimous among the personas that could adjudicate it.** The
  Learnings Researcher independently re-derived the transcript: the `.ps1`'s STEP 5 classifies
  by declared status (`$st -eq 'archived' → pre-archived`, `$st -eq 'done' → matched`) with
  location consulted only for existence/uniqueness, so `MEMBERS_QUEUE_LOCATED=0` with
  `PREMODE_MATCHED=2`/`PREMODE_PRE_ARCHIVED=11` is **exactly consistent**. *"The transcript was
  never wrong; rev 5's restatement of the rule was"* is confirmed accurate. **No
  re-measurement was performed or required.**
* **B-5, B-8, C-1, C-2, C-3, C-6, C-7, C-8, C-9, C-10, D-1, D-2, D-5, D-6, D-8, D-9 CLOSED.**
  Verified by the Architecture, Go, Constitution, Learnings and Agent-Native Parity personas
  respectively. `session_id` **is** exposed by `backlogit checkpoint list` summaries —
  empirically confirmed against the live tool (14 records), closing B-8's second half.
* **No scope expansion.** *(Scope Boundary.)* Rev 6 stays inside the 2-file surface; no third
  implementation file, no new shipment, backlog item or dependency edge; **every** rev-6 change
  maps 1:1 to an attempt-2 finding; AC-27/28/29 are finding-justified, not invented.

### Required before re-review (P0/P1 remediation queue — cycle 4)

**Max re-entry cycles are EXHAUSTED.** Attempt 3 was the last cycle available under the
plan-review gate's 2-re-entry limit. **The Escalation Protocol applies** — see the cycle-3
disposition below. Stage does **not** self-remediate these findings.

## Plan Review — Remediation Cycle 3 disposition

This session is **remediation cycle 3**. It closed the attempt-2 **P0 (B-1)** and **B-2, B-5,
B-8** outright, closed **B-3, B-4, B-6** in substance while introducing internal
contradictions in their *representation*, and closed the same-surface P2/P3 queue.
The re-run gate returned **`decision: FAIL`** on 10 P1s.

| Attempt-2 finding | Disposition |
|---|---|
| **B-1** (P0) | **CLOSED** — declared-status re-basing; unanimous across personas; corroborated by the committed transcript and by recorded institutional knowledge |
| **B-2** | **CLOSED in §4.4**; **residual in §4's inventory table** re-opened as **E-2** |
| **B-3** | **CLOSED in substance** (allowlist contract); *representation* re-opened as **E-4**, **E-5**, **E-6** |
| **B-4** | **CLOSED** — no-override language deleted; four distinct identity fields defined; `evidence_commit_sha` = LAST commit closes the replacement hole (Security Lens confirmed). Recovery-path consistency re-opened as **E-8** |
| **B-5** | **CLOSED** — a1 runs before pre-mode with no back-edge; narrow completion precondition |
| **B-6** | **CLOSED in substance** (no-fallback contract, honest reachability); *representation* re-opened as **E-6**, **E-7** |
| **B-7** | **RE-OPENED as E-9** — third consecutive recurrence. **Record-state half CLOSED this session** (see below); the **process** fix is recorded |
| **B-8** | **CLOSED** — MATCH-set ordering; `session_id` availability empirically confirmed |
| **C-1, C-2, C-3, C-6, C-7, C-8, C-9, C-10, D-1, D-2, D-5, D-6, D-8, D-9** | **CLOSED** — verified by the owning personas |
| **D-3, D-4, D-7, D-10** | **DEFERRED, explicitly** — recorded as **E-10**/**F-10**/**F-11**/**G-2**/**G-3**; D-7 is a one-line tool pin and should be taken in cycle 4 |

**Stage action taken this cycle (closes the record-state half of E-9):** `022-F` and
`022.001-T` are resynced to **rev 6** and re-banner-ed **NOT CLAIMABLE — attempt 3 FAIL**,
with the rev-4 locational C7/C8 wording and the invalid `blocked` shipment status explicitly
withdrawn. **Standing process fix recorded:** the backlog resync MUST travel in the same commit
as any plan revision bump, rather than being rediscovered as a finding each cycle — three
cycles have now each re-found it.

**The gate verdict is `decision: FAIL`.** `status:` stays `planned`. The plan is **not**
harvest-ready. **`021-S` must not be claimed**, and `017-S` behind it stays ineligible.
**No backlog items are created from this plan. No shipment is claimed.**



**This section is the operative provenance record.** It is the **last** `## Plan Review`
section in this file and restates the markers of the **most recent** gate run, so a
`harvest` provenance check that reads the latest section cannot accidentally bind to
**attempt 1**. Attempt 1's section is retained **only** for audit history and is
**superseded**.

<!-- plan-review-attempt: 3 -->

```text
dispatch_mode: multi-agent
decision: FAIL
```

| Field | Value |
|---|---|
| **Latest attempt** | **3** (of max 3 — **cycles EXHAUSTED; Escalation Protocol applies**) |
| **Reviewed revision** | rev 6 |
| **Decision** | **FAIL** |
| **Dispatch mode** | `multi-agent` (7/7 personas; anchor `openai`/`gpt-5.6-sol`/`high`) |
| **Anchor verdict** | **FAIL** (0 P0 / 6 P1) |
| **Harvest-ready** | **NO** |
| **Plan `status:`** | `planned` (unchanged — **not** advanced to `reviewed`) |
| **`021-S` claimable** | **NO** |
| **Shipment harvest provenance** | **NOT SET** — no shipment assembled, none claimed |
| **Blocking** | **0 P0** + **10 P1** (`E-1`…`E-10`) |
| **Attempt-2 P0 `B-1`** | **CLOSED** (unanimous; corroborated by committed evidence) |
| **Clean sheets** | Security Lens `0/0/0/0` (2nd consecutive) · Go Reviewer `0/0/0/1` |
| **Superseded** | attempt 1 (rev 4, FAIL) · attempt 2 (rev 5, FAIL) — their marker comments remain in their own sections above, earlier in the file |

**Harvest MUST halt on this plan.** `decision: FAIL` is the operative value.
**Stage has STOPPED remediating** — cycle 3 was the final permitted cycle, and the
Escalation Protocol (P-013.6) governs from here. Findings `E-1`…`E-10` are handed off, not
self-fixed.

> **SUPERSEDED BY ATTEMPT 4.** This attempt-3 block is retained as **audit history only**.
> The operative provenance record is
> `## Plan Review — AUTHORITATIVE GATE STATE` at the **end of this file**.

## Plan Review — attempt 4

**Operator-authorized EXCEPTIONAL gate run.** Attempt 4 exists only because the operator
explicitly authorized one bounded additional Stage remediation cycle after the attempt-3
circuit-breaker escalation, supported by an independent 5-model adversarial decision of
**`READY for one operator-authorized exceptional remediation cycle`**
(authorization timestamp **2026-09-13T11:35:43-07:00**). **Attempt history is NOT
renumbered** — attempts 1, 2 and 3 keep their own sections and markers above.

`dispatch_mode: multi-agent` · `decision: FAIL` · anchor `openai`/`gpt-5.6-sol`/`high`
· reviewed revision **rev 7**

### Persona coverage — 7/7

| Persona | Route | P0/P1/P2/P3 | Verdict |
|---|---|---|---|
| Architecture Strategist (**ANCHOR**) | `gpt-5.6-sol`/high | **0 / 8 / 1 / 0** | **`ANCHOR VERDICT: FAIL`** |
| Constitution Reviewer | `claude-opus-4.8` | 0 / 1 / 1 / 1 | FAIL |
| Go Reviewer | `claude-opus-4.8` | **0 / 0 / 0 / 0** | **PASS** (clean mechanical sheet) |
| Scope Boundary Auditor | `claude-opus-4.8` | 0 / 0 / 1 / 0 | ADVISORY |
| Security Lens Reviewer | `gpt-5.5`/high | **0 / 0 / 1 / 0** | ADVISORY |
| Learnings Researcher | default | 0 / 1 / 3 / 2 | FAIL |
| Agent-Native Parity | `grok-4.6`/high | 0 / 2 / 0 / 0 | FAIL |

**Deduplicated totals: P0 = 0 · P1 = 9 · P2 = 5 · P3 = 3.**

### Gate decision rationale

Rev 7 **closed** the adjudicated substance in several places — the Go Reviewer returned a
**fully clean mechanical sheet** (all five ABSENT literals unique at their cited lines, all
four rev-7 DELEGATION PRESENT literals absent at H0, so every `compound +/−` row is a genuine
red→green; §12 re-adds to exactly 97/23; no dependency added; no package-scope collision), and
Security Lens returned **no P0/P1 for the third consecutive attempt**. The Constitution
Reviewer independently **verified** Principle VI's dependency-discipline claims, both `PA-*`
records' required fields, D4's rejected-alternative and D5's Width-Isolation justification,
and confirmed both manifests against the live backlog.

**But the gate FAILS on 9 P1s, and the decisive pattern is the SAME ONE THAT KILLED REVS 5
AND 6: rev 7 fixed findings by introducing new inconsistencies elsewhere, and left its own
sweep incomplete.** Two P1s are outright **factual errors introduced by rev 7 itself**
(`H-3`, `H-7`), four are **incomplete sweeps** of rev 7's own decisions (`H-4`, `H-5`, `H-6`,
`H-9`), one is a **claimed-but-unapplied change** (`H-1`), one is an **unexecutable
mechanism** (`H-8`), and one is the **fourth consecutive recurrence** of the backlog
split-brain (`H-2`).

### P1 findings — blocking

| # | Finding | Persona(s) | Section |
|---|---|---|---|
| **H-1** | **Claimed-but-unapplied.** The rev-7 header (item 9) asserted the literal `## Plan Review — AUTHORITATIVE GATE STATE` heading was *"inserted before the operative marker"* — **it did not exist at review time**. Three live cross-references dangled, including one **inside a destructive-action approval record** (`PA-*` → *"readiness state recorded in …"*). Same overclaim class as `A-16` and `E-10`. *(The heading is created below as part of recording THIS gate; the finding stands — the header claimed it before it existed.)* | Anchor, Constitution, Scope, Learnings | header item 9 / file tail |
| **H-2** | **Backlog split-brain, FOURTH consecutive recurrence (`A-14` → `B-7` → `E-9` → `H-2`).** `022-F` and `022.001-T` still carried `REV 6 RECONCILIATION` / `ATTEMPT 3 FAIL` / the **rev-6 declared-status C7/C8 restatement** that rev 7 itself declares superseded, and no `AC-30` / `PA-*` cross-reference. Rev 7's own cycle-3 "standing process fix" mandates the resync travel in the **same commit** as the revision bump — and it did not. | Constitution, Agent-Native Parity, Anchor | backlog records |
| **H-3** | **Rev 7's `F-8` correction is FACTUALLY FALSE — a NEW defect.** Rev 7 asserted `consumer_id: "ship"` is *"access identity, not a filter."* **Measured live: it IS a filter.** MCP `backlogit_list_checkpoints` with `consumer_id=ship` returns `total=1`; unfiltered returns `total=15`. The tool manifest describes it as **"Filter by consumer/agent ID"**, and CLI `--agent ship`/`--agent stage`/unfiltered return **1 / 14 / 15**. Passing `consumer_id: "ship"` as §9.1 still instructs **hides all 14 stage records**, defeating the full-enumeration anomaly gate and making AC-29's leftover warnings unable to fire. **Correct instruction: OMIT `consumer_id` entirely.** | Agent-Native Parity | §9.1 steps 1/2/2a, AC-29, header item 8 |
| **H-4** | **§4.1's own opening still contradicts the rev-7 allowlist.** The operative first line of §4.1 still reads *"The replacement text MUST NOT enumerate the authorized verdicts"*, while the rev-7 block and AC-25 **permit** an explicit `{CASCADE, SAFE_CLOSE}` allowlist. The same section's *"future close-path work lands without touching this file again"* claim also contradicts the bound-to-installed-contract rule. | Anchor | §4.1 |
| **H-5** | **§13 still instructs reviewers to validate SUPERSEDED rev-6 behavior.** Item 5 asks for rows 14/15's *rev-6 declared-status* PRESENT phrases (replaced by delegation literals); item 7 asks whether the *P-015/skill-disagreement HALT* is correct (explicitly **deleted** by rev 7). Rev 7 re-scoped items 14/15 but did not sweep 5 and 7. | Anchor | §13 items 5, 7 |
| **H-6** | **The pointer-only decision was not swept into the execution path.** §9 step 14, §9.3 step 5 and §9.4.1 still assert per-item `matched` / `pre-archived` allocations of the installed pre-mode — exactly what E-2/E-3 forbade the plan from claiming. The aggregate `PROCEED` outcome is correct and measured; the **per-item allocation** is not the plan's to assert. | Anchor | §9 step 14, §9.3, §9.4.1 |
| **H-7** | **Evidence identity is internally inconsistent — NEW in rev 7.** Hardening 9a and AC-20a name **`_sha256`** fields, but the rev-7 normative derivation compares **Git blob object IDs** (`git rev-parse {commit}:{path}` vs `git hash-object`), and the "measured this session" values shown are **abbreviated blob IDs**, not SHA-256 digests. No full digest is pinned anywhere. An executor cannot satisfy the stated SHA-256 contract using the prescribed algorithm. | Anchor | Hardening 9a, §9.4a check 5, AC-20a |
| **H-8** | **The mandatory "verified snapshot restore" is not executable as specified.** The only restore command is `git restore -- .backlogit/queue/ .backlogit/archive/`, which restores **tracked paths from Git HEAD**, not from the skill's **Step 0(b) in-memory snapshot**, and does not remove newly-created untracked archive files. Because a1 sets the feature `done` **after** the branch HEAD and **before** the cascade, a HEAD restore can itself disagree with the pre-call snapshot. The `PA-*` records additionally require verifying a *"protected set"* that the installed skill says **does not exist** on a qualifying CASCADE path. | Anchor | §9.4b, §10.3, `PA-021`/`PA-017` |
| **H-9** | **Inventory vs implementation sequence mismatch.** §4 makes **C1–C8** mandatory, but §3 describes the change as **C1–C6** and §7's H1 instructs applying only ADD-1, ADD-2 and **C1–C6** before expecting green. Following §7 omits C7/C8, which rows 14–15 and AC-22 require — the documented H1 state is unreachable from the documented action. | Anchor | §3, §7 |

### P2 findings — advisory

| # | Finding | Persona |
|---|---|---|
| J-1 | `PA-*` condition "no scope expansion" defines scope as **manifest IDs only**, but a valid cascade necessarily also transitions the **shipment record itself** (`021-S`/`017-S`), which is outside both manifests. A literal reading voids the approval on every valid close. | Security Lens |
| J-2 | **`F-5` neither fixed nor dispositioned.** §9 steps 5/17 name P-018 engagement but not the workspace-verified **`[bot]`-suffixed REST fallback**, the **permanently-closed post-merge window**, or the **GraphQL `reviewThreads`** blocking-state rule. This plan has **two** merges, so the irreversible-window risk applies twice. | Learnings |
| J-3 | **No cycle-4 disposition table** — the `E-10` pattern reproduced one revision later. `F-3` (row 16 absent from §8.1.1, so a1 could ship a location-keyed S0 and still pass green), `G-4`, `G-5`, `G-7` are each **re-verified still open** and neither closed nor named-as-deferred. | Learnings, Scope |
| J-4 | D4 still uses *"removes authority duplication rather than adding reach"* as a compensating control, contradicting Hardening 1 (genuine authority increase) and §9.4b (new cascade reachability). | Anchor |
| J-5 | The §9.1 tool-surface claims were the only load-bearing empirical claims with **no committed evidence artifact** — self-reported *"verified this session"*. `H-3` is the direct consequence: one of them was wrong. | Learnings |

### P3 findings — noted

| # | Finding | Persona |
|---|---|---|
| K-1 | `PA-017-CASCADE` condition 4 writes `dependencies: [021-S] (type: blocks)`; the resolved record exposes `[{"id":"021-S","type":"blocks"}]` while the raw frontmatter is a bare list. Dependency existence and `type` value are correct; the inline representation is imprecise. | Constitution, Agent-Native Parity |
| K-2 | `D-10`'s closure is attributed to rev 7, but the only `--complexity` prohibition found is in a **rev-5-labelled** §10.1 block. Accounting only — the prohibition itself is correct. | Learnings |
| K-3 | Label collision: the pinned STEP-11 tokens `C7_SHIPMENT_ARCHIVED_SHIPPED` / `C8_NOTHING_OUTSIDE_MANIFEST_TOUCHED` are referred to as "C7"/"C8" beside this plan's **clause sites** C7/C8 — the exact hazard §4.4's rev-2 note renamed `A-1`/`A-2` to avoid. | Learnings |

### Affirmative clearances (recorded as evidence, not findings)

* **Go Reviewer: `0/0/0/0` — a fully clean mechanical sheet.** All five ABSENT literals occur
  **exactly once** at L255/L777/L819/L790/L810 and are single-line; **all four rev-7
  DELEGATION PRESENT literals occur ZERO times**, so no `compound +/−` row can be vacuously
  green at H0. Rows 14/15's near-prefix needles **diverge at the next token** and cannot
  cross-match. Row 17 = 2 (preservation, intended), row 18 = 1, the `a1` anchor correctly
  absent. `testify` **absent** from `go.mod`; `repoRoot(t)` package-level and already reused;
  ≤4 helpers / 1 function; **zero** package-scope collisions. §8.1 has **18** rows with Kind
  tallies matching AC-23. `8+18+4+4+12+4+4+4+29+10 = 97`, margin **23**.
* **Security Lens: no P0/P1 for the THIRD consecutive attempt.** The five §5.0 gates still
  hold conjunctively; the E-1 condition-2 split **did not open a hole**; `archived_status:
  queued` still never proves completion; the unknown-token HALT still prevents a future skill
  from silently gaining a destructive verdict; `BLOCKED` correctly unhonorable; no secrets,
  tokens, emails or user-profile paths in the committed evidence; **no check requires a
  hardcoded absolute executable path**.
* **Hardening 1's corrected containment claim is CONFIRMED ACCURATE.** Pre-mode structurally
  **cannot** catch an unfinished descendant declaring `status: archived` (it is accepted
  without a status check), so §5.0 **condition 5** — the recorded Stage disposition — is the
  real containment. Rev 7's withdrawal of the rev-6 overclaim is correct.
* **Constitution: Principle VI, VII, D4 and D5 all independently VERIFIED.** `testify` absent,
  `repoRoot` reusable, both `PA-*` records carry every required field including
  `ActionRisk: destructive` / `ActionResult: approved` / the operator timestamp, the "does NOT
  authorize" block is unambiguous, D4 names a genuinely rejected alternative and **declines**
  the flattering freeze-scope framing, and D5 is a real justification rather than a waiver.
* **Manifests VERIFIED against the live backlog**, independently by three personas:
  `021-S` = exactly `[022-F, 022.001-T]`, no dependencies; `017-S` = exactly the 13 named
  members, dependency `021-S` type `blocks`; `018-F` is a root. **`backlogit doctor`: clean.**
* **Probe-25 evidence VERIFIED.** `PROBE25_RESULT=PASS`, `FAILED_CRITERIA=0`, the **nine
  pinned STEP-11 names match the transcript exactly and in order**, topology tokens exact, and
  the rev-7 disclosure that **criteria 2 and 3 are PowerShell-derived** while the real engine
  cascade begins at `### STEP 8: AUTHORIZED CALL` is **confirmed accurate**.
* **No scope expansion.** *(Scope Boundary.)* Every substantive rev-7 change maps 1:1 to an
  adjudicated decision; surface still exactly 2 implementation files, 0 new production files;
  **no new shipment, backlog item or dependency edge**; the `PA-*` records smuggle **no** new
  authority; rev 7 net-**removes** machinery. Working tree touches planning artifacts only —
  the Go test file **does not exist**, confirming no Stage role-boundary violation.

### Required before re-review (P0/P1 remediation queue — cycle 5)

**Stage does NOT self-remediate these.** The operator authorized **one** exceptional cycle;
that cycle is now spent and its gate returned `FAIL`. `H-1`…`H-9` are handed off for operator
adjudication exactly as `E-1`…`E-10` were. **The plan-review re-entry budget remains
exhausted** and the Escalation Protocol (P-013.6) continues to govern.

## Plan Review — AUTHORITATIVE GATE STATE

**This heading is the literal, stable anchor the operator directed Stage to create.** It is
the **LAST** `## Plan Review` section in this file, so a `harvest` provenance check or
latest-marker parser that reads the final section binds **here**, to **attempt 4** — never to
attempt 1, 2 or 3, whose marker comments remain in their own sections above as audit history.

<!-- plan-review-attempt: 4 -->

```text
dispatch_mode: multi-agent
decision: FAIL
```

| Field | Value |
|---|---|
| **Latest attempt** | **4** — operator-authorized **exceptional** run (history **not** renumbered) |
| **Reviewed revision** | **rev 7** |
| **Decision** | **FAIL** |
| **Dispatch mode** | `multi-agent` (**7/7** personas; anchor `openai`/`gpt-5.6-sol`/`high`) |
| **Anchor verdict** | **FAIL** (0 P0 / 8 P1) |
| **Blocking** | **0 P0** + **9 P1** (`H-1`…`H-9`), deduplicated across personas |
| **Advisory / noted** | 5 P2 (`J-1`…`J-5`) · 3 P3 (`K-1`…`K-3`) |
| **Harvest-ready** | **NO** |
| **Plan `status:`** | `planned` (unchanged — **NOT** advanced to `reviewed`) |
| **`021-S` claimable** | **NO** |
| **`017-S`** | **INELIGIBLE** — `queued`, dependency `[021-S]` unsatisfied |
| **Shipment harvest provenance** | **NOT SET** — no shipment assembled, none claimed |
| **`PA-021-CASCADE` / `PA-017-CASCADE`** | **RECORDED but NOT EXERCISABLE** — both are gated on this readiness state, and both remain **unexercised**. No destructive operation was performed by this cycle |
| **Clean sheets** | Go Reviewer `0/0/0/0` · Security Lens `0/0/1/0` (no P0/P1, 3rd consecutive) |
| **Superseded** | attempt 1 (rev 4, FAIL) · attempt 2 (rev 5, FAIL) · attempt 3 (rev 6, FAIL) |
| **Re-entry budget** | **EXHAUSTED.** The one operator-authorized exceptional cycle is spent |

**Harvest MUST halt on this plan.** `decision: FAIL` is the operative value.
**`021-S` MUST NOT be claimed**, and `017-S` behind it stays ineligible. **No backlog items
are created from this plan. No shipment is assembled or claimed. Neither approved destructive
action is executed.** Findings `H-1`…`H-9` are **handed off, not self-fixed**.


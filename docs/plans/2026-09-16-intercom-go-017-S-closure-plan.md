---
title: "Implementation Plan — 017-S closure (Stage artifact branch/PR policy gap correction)"
date: 2026-09-16
revision: 4
status: draft
agent: Stage
shipment: 017-S
covering_feature: 018-F
source_decision: docs/decisions/2026-09-13-intercom-go-a-only-simplification-decision.md
prerequisite_plan: docs/plans/2026-09-13-intercom-go-ship-feature-completion-decided-plan.md
prerequisite_shipment: 021-S
requires_plan_hardening: yes
---

# Implementation Plan — `017-S` closure

> **This is the closure plan whose absence was `017-S`'s third independent ground for
> ineligibility.** Prerequisite plan §11.1 requires it to supply the `017-S` equivalents of
> §9 step 13a (pre-cascade baseline commit), §10.3.1 (`{pre_paths}` / `{post_paths}`) and
> §10.3.2 (verified restore), plus its own preflight gate set and its own linked-deliberation
> re-measurement. This plan supplies all five.

**Source document:** `docs/decisions/2026-09-13-intercom-go-a-only-simplification-decision.md`
(verdict `SIMPLIFICATION_VALID`, option **O-2**, §4.2 "How `017-S` then closes").
**Top-level work item:** shipment `017-S`, covering feature `018-F`.

## 1. Premise correction (divergence from the prior checkpoint)

The prior Stage checkpoint `checkpoint-20260916-193114.json` recorded `017-S` as *"blocked by
`021-S` on all three original grounds."* **Two of those three grounds are now discharged.**
This plan is authored against current measured truth, not the stale checkpoint premise.

| Ground (017-S record, 2026-09-14) | State at this plan's authoring | Evidence |
|---|---|---|
| **1. DEPENDENCY** — `021-S` has not shipped and is not claimed | **DISCHARGED** | `021-S` status `archived`; shipped at `74330d3`, closure merge `f2d4cf9`, PR #56 |
| **2. CAPABILITY** — parent-completion step not merged | **DISCHARGED** | `0cb4a42` added Ship Step 6.1(a1) + the §3.2 Role Boundary grant; both live on `main` |
| **3. NO CLOSURE PLAN** — no plan authorizes or sequences this shipment's closure | **DISCHARGED BY THIS DOCUMENT**, on its own plan-review PASS | this plan |
| **4. Decision §7 item 4** — no committed probe exercises the root-included, archived-sibling cascade shape | **DISCHARGED** (added at revision 2) | `docs/plans/evidence/2026-09-12-task-only-shipment-finalization/probe25-root-included-cascade-fixture.txt`, committed at `e36d853`: `PROBE25_RESULT=PASS`, `FAILED_CRITERIA=0`, `TOPOLOGY_MEMBERS=13`, `ROOT_INCLUDED=True`, `PRE_ARCHIVED=11`, `ARCHIVED_STATUS_FLAVOUR=queued` |

> **Revision-2 correction (self-applied).** Revision 1 quoted decision §7 item 4 as if its
> premise were still live and mandated authoring a **new** fixture test. That premise was
> **stale**: Probe 25 already exists, is committed, and measures this exact shape. Revision 1
> applied "measured truth, not stale premise" rigor to grounds 1–3 and **failed to apply it to
> its own preflight gate**. PF-6 is accordingly rewritten as **read-only re-verification**
> (§4.1), which also removes the unassignable test-authoring mandate revision 1 created.

**No ground is waived by assertion.** Ground 1 and 2 are re-verified mechanically at
**PF-8** and **PF-9**; ground 3 is discharged only when this plan passes its review gate.

## 2. Current-state measurements (Stage, 2026-09-16)

All values below were measured against the working tree at `f2d4cf9` with a clean
`git status` (no tracked modifications). Each is **re-measured by Ship at invocation**; a
Stage-time measurement is evidence about a different moment and is never inherited
(prerequisite plan §11 note).

| ID | Claim | Measured value (verbatim) |
|---|---|---|
| **P-1** | `017-S` manifest is 13 members | `items` = `018.008-T, 018.001-T, 018.001.001-ST, 018.001.002-ST, 018.001.003-ST, 018.002-T, 018.002.001-ST, 018.002.002-ST, 018.002.003-ST, 018.003-T, 018.003.001-ST, 018.003.002-ST, 018-F` ⇒ **13** |
| **P-2** | `017-S` dependencies are `[021-S]` exactly | `dependencies: [{id: 021-S, type: blocks}]` |
| **P-3** | `021-S` has shipped | `backlogit shipment get 021-S` ⇒ `status: archived`; `022-F` archive frontmatter carries `archived_status: done`, `commit: 74330d3…` |
| **P-4** | The live descendant graph of `018-F` ∪ `{018-F}` equals the manifest | Walking `parent_id` across `queue/` ∪ `archive/` yields exactly 12 descendants; ∪ `{018-F}` = **13** = manifest. **Set equality holds** |
| **P-5** | Exactly two manifest members are live | `018-F` = `queued` (queue), `018.008-T` = `queued` (queue). All other 11 = `status: archived` (archive) |
| **P-6** | The 11 archived members satisfy §3.3 conditions 1–4 | All 11: declare `status: archived`; exactly one physical copy (archive, not queue); well-formed `archived_from`; **present** `archived_status` (value `queued`); resolvable `artifact_type` |
| **P-7** | The 11 archived members are exactly the §3.3 allowlist | Measured set is byte-identical to the prerequisite plan §3.3 `017-S` allowlist. **No archived descendant is off-list** |
| **P-8** | The validated linked-deliberation set for the 13-member manifest is EMPTY | Three independent methods — see §2.1 |
| **P-9** | Workspace containment is clean | `backlogit doctor` ⇒ `No issues found.`, exit **0** |
| **P-10** | The topology gate passes for `017-S` at `pre_claim` | `autoharness gate pipeline-topology --mode agent --shipment 017-S --phase pre_claim --json` ⇒ `"exit_code": 0`, `"blocked": false`, `"message": "topology gate pass"`, exit **0** |
| **P-11** | The cascade close command surface | `backlogit shipment ship <id> --sha <s> --message <m> --author <a>` (`--help` verbatim) |
| **P-12** | Merge commits in this repo are 2-parent, mainline = parent 1 | `git cat-file -p f2d4cf9` ⇒ `parent 74330d3…` (mainline), `parent f88d413…`. **Corroborative only — repository convention.** It does **not** establish the shape of the future S-7/S-17 merges, which depend on the merge strategy chosen then; **S-7a** verifies those empirically |
| **P-13** | **The claim cascade over the exact `017-S` shape** (13 members, root included, 11 genuinely `status: archived` with `archived_status: queued`) | **Probe 26**, `docs/plans/evidence/2026-09-16-017-S-closure/probe26-claim-cascade-root-included.txt`. `FIXTURE_VALIDITY_GATE=PASS`, `CLAIM_EXIT_CODE=0`, `PROBE26_RESULT=PASS`, `FAILED_CRITERIA=0`. Root feature `queued -> active`; **all 11 archived members unchanged** (`loc=archive`, `status=archived`, `archived_status=queued`, `parent_id` intact); live task `queued -> active`; nothing outside the manifest touched |
| **P-14** | **`backlogit move <id> --status archived` is NOT an archive operation** | Probe 26 step 2: a loop using `move --status archived` left all 11 records `queued` and queue-located and surfaced no error a naive loop would catch, yielding `PRE_ARCHIVED=0`. The archive operation is **`backlogit archive <id>`**. No step in this plan archives via `move` |
| **P-15** | **There is NO claim-release / unclaim path for a shipment** | `backlogit shipment --help` ⇒ the complete subcommand set is **`add`, `claim`, `create`, `get`, `list`, `return-blocked`, `ship`**. There is **no `unclaim`, `release`, `return`, or `unshipment`**. `backlogit shipment return-blocked --help` ⇒ it returns a **single ITEM** from a shipment (`--shipment`/`--item`/`--reason`), i.e. it **mutates the manifest** — it is emphatically **not** a shipment-release path, and invoking it here would break the 13-member invariant and §3.2 condition 3 outright. Corroborated by the installed `_ship.agent.md` (shipment transitions are `queued -> active`, `active -> shipped`, `active -> abandoned` only) and by `docs/compound/2026-05-07-backlogit-shipment-status-constraints.md` (the generic `move` route **refuses** shipment status writes with **exit 9**). **⇒ `active -> queued` does not exist. A claimed shipment cannot be un-claimed.** |

### 2.1 P-8 — linked-deliberation re-measurement (three independent methods)

Re-measured over **all 13 members** (the prerequisite plan's T-11 measured only three records
and is **not inherited**):

1. `backlogit link list <id>` for each of the 13 ⇒ `"links": []` for **12 of 13**.
2. Engine linked-deliberation pattern `\b(?:DL\d+|[0-9]+(?:\.[0-9]+)*-DL)\b` across all 13
   records ⇒ **0 matches**.
3. `source_deliberation_id` across all 13 records ⇒ **0 matches**.
   `SELECT COUNT(*) … artifact_type='deliberation'` ⇒ **2** (`001-DL`, `002-DL`).

**Verdict: the validated linked-deliberation set for this manifest is EMPTY.** The
`COUNT(*) = 2` is **expected and disclosed** — those two deliberations belong to the `021-S`
lineage, are linked to no member of this manifest, and are not reachable by the cascade.

> **DISCLOSED NON-DELIBERATION LINK (new this cycle).** `018.008-T` carries exactly one
> outgoing link: `related_to → 019.007-T`. **It is not a deliberation link** and does not
> populate the linked-deliberation set. `019.007-T` is `status: blocked` with
> `parent_id: 019-F` — it is **outside this manifest and is not a descendant of `018-F`**, so
> it is outside the cascade's archival scope. **PF-5 records `019.007-T`'s pre-state for PROVENANCE ONLY; its value is INFORMATIONAL and is NOT an assertion baseline.** The untouched assertion is raw-byte identity between the **S-11** `{pre_paths}` capture and **S-15**/**R-6** — never against PF-5, because the S-6 dependency carve-out means `019.007-T` may legitimately change between PF-5 and S-11. **(Superseded prose kept for traceability:** PF-5 asserts `019.007-T` is untouched after
> the cascade.** This link did not exist in the prerequisite plan's T-11 scope and is recorded
> here rather than silently inherited.

### 2.2 Tooling drift recorded

The prerequisite plan's T-11 invocation form `backlogit link list --id <id>` **no longer
resolves** on backlogit `1.10.1` (`Error: unknown flag: --id`). The current surface is
positional: `backlogit link list <id>`. Any step re-running T-11 MUST use the positional form.

## 3. Authority model (inherited, restated, not re-litigated)

This plan **inherits** the prerequisite plan's §3.1 (declared status is lifecycle; location is
integrity only), §3.2 (the five conjunctive grant conditions) and §3.3 (the pre-archived
exemption and the `017-S` allowlist). It **does not amend them** and creates no new verdict,
no new token and no new policy surface.

### 3.1 The five §3.2 conditions, evaluated for `018-F`

| # | Condition | Status at claim time | Discharged by |
|---|---|---|---|
| 1 | every manifest descendant is **complete** (`status: done`, or exempt under §3.3) | 11 exempt under §3.3; `018.008-T` must reach `done` | **S-6** |
| 2 | **no live descendant remains** (`queued`/`active`) | `018.008-T` is the only live descendant | **S-6** |
| 3 | **topology intact** + descendant graph ∪ `{feature}` **set-equal** to manifest | **HOLDS NOW** (P-4) | re-verified at **PF-3** |
| 4 | **containment holds** for every member | **HOLDS NOW** (P-5, P-6, P-9) | re-verified at **PF-2/PF-3** |
| 5 | the feature's **live status is exactly `active`** | `018-F` is `queued` now; the **claim** transitions it | **S-2**, verified at **S-3** |

### 3.2 Condition 5 is satisfied by the claim, never by a Stage or Ship write

**`018-F` MUST be `queued` at PF/S-1 and `active` at S-10.** The transition is performed by
**`backlogit_claim_shipment` itself**, exactly as `022-F` was transitioned in the `021-S`
release. Prerequisite plan **AC-19** is normative and is inherited verbatim in shape:

> *"After claiming `021-S`, `022-F`'s live status is re-read and required to be exactly
> `active`; any other value HALTs. **Ship performs no `queued -> active` transition on a
> feature.**"*

**Consequences, both load-bearing:**

* **Stage MUST NOT pre-set `018-F` to `active`.** Step 0.5 item 3b runs
  `mode: pre` with `expected_status: queued` **before** the claim, "while every manifest
  member still shares the uniform pre-claim status." A pre-set `active` feature would
  classify `status-mismatch` ⇒ `RECONCILE_FAIL` ⇒ **HALT before the claim**. This plan
  therefore leaves `018-F` `queued`, deliberately.
* **Ship MUST NOT transition it either.** If the claim does not yield `active`, **S-3 HALTs
  and returns to Stage.** No agent may repair it by writing the status directly.

## 4. Preflight gate set (PF-1 … PF-9)

Run **in order**, **before** the claim, on the **pre-claim branch** defined in §5 S-0 (a
Stage-independent, Ship-created working branch; **NOT** the post-merge *closure* branch of
S-8, which does not exist yet at preflight time — these are two different branches and
revision 1 conflated them). **Any failure HALTs with no mutation.** These are this plan's own
gates; none is inherited by reference.

| Gate | Check | Pass criterion |
|---|---|---|
| **PF-0** | **Engine identity** | `backlogit version` ⇒ `1.10.1-…`; SHA-256 of the resolved `backlogit` binary equals `1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98`. **A mismatch invalidates every measurement in §2 and in PF-6 ⇒ HALT, return to Stage** for re-measurement. Resolve via `Get-Command backlogit`, never a hardcoded path |
| **PF-1** | This plan is present on `origin/main` | `git show origin/main:docs/plans/2026-09-16-intercom-go-017-S-closure-plan.md` resolves **and is byte-identical to the reviewed revision**; its `## Plan Review` section records a literal **`dispatch_mode:`** marker and a literal **`decision: PASS`** for **attempt 3**, with the attempt-3 persona coverage named. **No separate `docs/reviews/` artifact is required** — revision 3 demanded one that does not exist, making the first preflight gate unevaluable; the verdict lives in this plan's own `## Plan Review` section |
| **PF-2** | Workspace integrity | `backlogit doctor` ⇒ `No issues found.`, exit **0** |
| **PF-3** | Manifest + topology | manifest = 13; descendant graph ∪ `{018-F}` **set-equal** to manifest; every descendant's `parent_id` resolves to `018-F` transitively; enumeration completed without error |
| **PF-4** | §3.3 allowlist | the archived-declaring descendant set is **exactly** the 11-member allowlist; each satisfies §3.3 conditions 1–4. **Any off-list archived descendant ⇒ HALT** |
| **PF-5** | Linked-deliberation set | re-run §2.1's three methods (positional `link list`) ⇒ set **EMPTY**. Record `019.007-T`'s pre-state **for PROVENANCE ONLY — informational, NOT an assertion baseline** (the byte-identity assertion is anchored at **S-11**, per the S-6 carve-out) **and its `dependencies: [018.008-T]` edge** for the §7 R-6 assertion |
| **PF-6** | **Cascade-shape probe re-verification (read-only)** | see §4.1 |
| **PF-7** | Topology gate | `--phase pre_claim` ⇒ exit **0** |
| **PF-8** | Ground-1 re-verification | `021-S` ⇒ `status: archived` **and its own frontmatter `archived_status` is exactly `shipped`**, with `commit` provenance recorded. **`archived` alone is insufficient** — it is also the terminal state of an *abandoned* shipment, so `archived` does not by itself prove 021-S shipped. Cross-check `74330d3` / `f2d4cf9` / PR #56 |
| **PF-9** | Ground-2 re-verification (capability liveness) | the **installed** `_ship.agent.md` contains Step 6.1(a1) with the S0–S5 selector, and `.github/policies/workflow-policies.md` carries the §3.2 grant. Verified against the **installed** files, not against this plan's description of them |
| **PF-10** | **Close-path authority agreement** (revision 4) — makes §5 S-14's authority reconciliation falsifiable instead of asserted | All of: (a) `backlogit shipment ship --help` exposes **exactly** `--sha`, `--message`, `--author` and **no** binding parameter; (b) `.autoharness/backlog-registry.yaml` `ship_shipment.params` is **exactly** `shipment_id`, `sha`, `message`, `author` — **no** binding parameter; (c) `shipment-reconcile/SKILL.md` still accepts `classification_binding` for `safe-close` and still routes a bound `CASCADE` verdict into the Cascade Close Sub-Procedure; (d) installed `_ship.agent.md` still requires the binding to be carried into the close call. **If (a) or (b) gains a binding parameter, the direct-cascade route becomes executable and this plan's S-14 reconciliation is VOID ⇒ HALT, return to Stage for re-measurement.** If (c) or (d) has drifted ⇒ **HALT** |

### 4.1 PF-6 — cascade-shape probe re-verification (READ-ONLY)

> **Revision-2 rewrite.** Revision 1 made PF-6 a mandate to **author** a new behavioral test.
> That was wrong on two counts, both self-inflicted: (a) its premise was **stale** — decision
> §7 item 4's probe **already exists and is committed**; (b) it was **unassignable** — Stage
> may not write test files (Role Boundary) and Ship may not author new plan-mandated tests
> outside its shipment scope, so revision 1's own gate was executable by **no agent in the
> pipeline** and would have deadlocked the route at its first preflight.

PF-6 is now a **read-only re-verification of two committed probe artifacts**. It authors
nothing, mutates nothing, and adds no backlog item.

| Sub-gate | Artifact | Required values |
|---|---|---|
| **PF-6a** | `docs/plans/evidence/2026-09-12-task-only-shipment-finalization/probe25-root-included-cascade-fixture.txt` (committed `e36d853`) | `DIGEST_GATE=PASS`; `TOPOLOGY_MEMBERS=13`; `ROOT_INCLUDED=True`; `PRE_ARCHIVED=11`; `ARCHIVED_STATUS_FLAVOUR=queued`; C1–C8 all `PASS`; `FAILED_CRITERIA=0`; `PROBE25_RESULT=PASS` |
| **PF-6b** | `docs/plans/evidence/2026-09-16-017-S-closure/probe26-claim-cascade-root-included.txt` — **MUST resolve via `git show origin/main:<path>`**, with the same rigor PF-1 applies to this plan. The generating script `docs/plans/evidence/2026-09-16-017-S-closure/probe26.ps1` must likewise resolve on `origin/main`. **A working-tree-only copy does NOT satisfy this gate** — Probe 26 is the sole evidentiary basis for P-13, S-3, S-3b and S-4's pre-emption disclosure, and an unpinned artifact would let it be satisfied from a local, unreviewed file (or deadlock the route at first preflight if it never merges) | `DIGEST_GATE=PASS`; **`FIXTURE_VALIDITY_GATE=PASS`**; `TOPOLOGY_MEMBERS=13`; `ROOT_INCLUDED=True`; `PRE_ARCHIVED=11`; `ARCHIVED_STATUS_FLAVOUR=queued`; `CLAIM_EXIT_CODE=0`; `C1`–`C6` all `PASS`; `FAILED_CRITERIA=0`; `PROBE26_RESULT=PASS` |
| **PF-6c** | Engine agreement | both artifacts' `ENGINE_SHA256` equal PF-0's measured digest. **Any divergence ⇒ HALT** — the probes measured a different engine than the one about to run |

**Why two probes, not one.** They close **different** gaps and neither subsumes the other.
**Each is credited with exactly what its artifact records, and nothing more** — the discipline
this lineage has failed seven times:

* **Probe 25** covers the **structural/topology** half over the root-included, archived-sibling
  shape: `TOPOLOGY_MEMBERS=13`, `ROOT_INCLUDED=True`, `PRE_ARCHIVED=11`,
  `ARCHIVED_STATUS_FLAVOUR=queued`, and its criteria **C1–C8 as literally named in the file**.
* **Probe 26** covers the **claim** half — what `shipment claim` does to a manifest holding 11
  genuinely `status: archived` members. This was **entirely unmeasured** before revision 2: the
  only claim evidence was `021-S`'s **2-member, all-`queued`, zero-archived** manifest, which
  cannot speak for this shape. Probe 26 measured it directly: the archived members are
  **untouched** by the claim (`C3=PASS`), so §3.2 condition 2 and the §3.3 exemption survive
  it, and the root feature **does** go `queued -> active` (`C1=PASS`), which is the *measured*
  basis for S-3 — replacing revision 1's reliance on AC-19, a **normative requirement** rather
  than an observation of the installed engine.

> **DISCLOSURE — what neither probe measures (revision 3).** Revision 2's §4.1 credited Probe
> 25 with *"`classify-close-path` ⇒ `CASCADE`, zero-mutation classification, subtree-only
> archival, and the off-allowlist HALT."* **Three of those four are not in the artifact**, and
> asserting them was a relapse into the exact defect class above. Corrected:
> * Probe 25 ran at `PROBE_HEAD=4fd21c5` — **before `mode: classify-close-path` existed** (it
>   was added by `021-S`, merged `f2d4cf9`). Its STEP 6 computed the P-015 criteria **inside
>   the probe script**; it **never invoked the installed read-only boundary**.
> * There is **no zero-mutation measurement** around classification in either artifact.
> * There is **no off-allowlist arm** in either artifact — nothing was measured HALTing on an
>   off-list archived descendant.
> * Probe 26 records `AUTHORIZATION_SURFACE=CLI-only (MCP not probed, not authorized)`.
>
> **Consequence, stated plainly:** the bound `classify-close-path -> safe-close-revalidation ->
> cascade` route over the `017-S` shape is **UNMEASURED on the post-`021-S` engine**. It is
> protected **only** by S-13's fail-closed HALT (verdict must be exactly `CASCADE`; any other
> token, or an absent/unparseable one, HALTs) and S-14's binding-refusal tokens. That
> protection is real and fail-closed, but it is a **guard, not evidence** — and this plan does
> not claim otherwise. PF-4's allowlist check is likewise the *only* off-allowlist protection,
> and it is a **plan-level gate, not a measured engine behavior**.

**Failure disposition:** a PF-6 failure is a **HALT and return to Stage**, never a waiver. If
either artifact is missing, altered, or reports any non-`PASS` value, the evidentiary basis for
this plan's claim and cascade steps is void.

**Manifest invariant (NON-NEGOTIABLE):** PF-6 adds **no** backlog item. Probe artifacts are
evidence files, not backlog items. Adding any item under `018-F` would break the 13-member
manifest, break §3.2 condition 3 set equality, and invalidate `PA-017-CASCADE` condition 6.

## 5. Execution sequence — the exact `018-F → done` path

**This is the authorized sequence.** Steps are strictly ordered; each HALT is terminal for the
session and returns to Stage.

**Branch vocabulary (revision 2 — revision 1 used "closure branch" for two different
branches).** This plan names exactly two:

* **pre-claim branch** (`S-0`) — created by Ship before preflight; carries PF-0…PF-9, the
  claim, the `018.008-T` implementation, and the S-7 PR. It is the *implementation* branch.
* **closure branch** (`S-8`) — created **after** the S-7 merge, off the merged `main`; carries
  the a0/a1/cascade S2 mutations and the S-17 closure PR. It is the *record-mutation* branch.

No PF gate references the closure branch; no S2 mutation occurs on the pre-claim branch.

### Phase 1 — claim (establishes condition 5)

* **S-0** — Create the **pre-claim branch** off current `origin/main`. All of §4's preflight
  gates run here. **No S2 mutation on `main` at any point (P-011).**
* **S-1** — Step 0.5 item **3b** intake reconciliation, **pre-claim**: `shipment-reconcile`
  `mode: pre`, `expected_status: queued`. Expected: `018-F` and `018.008-T` `matched` at
  `queued`; the 11 allowlisted members classify **`pre-archived`** (accepted without a status
  check); no orphans; record classification `record-consistent`. **Continue only on
  `PROCEED`.**
  * **Known-stale disclosure (decision §7 item 5):** the installed `_ship.agent.md` summary at
    the Step 6.1 preamble describes `mode: pre` as queue-only. It is **stale**; the
    `shipment-reconcile` SKILL is the runtime authority and does classify archived members.
    Trust the skill's actual output, not the agent-file summary. If the two disagree in a way
    that changes a gate outcome ⇒ **HALT, return to Stage.**
* **S-2** — Claim `017-S` via the **CLI surface**: `backlogit shipment claim 017-S`.
  * **Surface pin (revision 3).** Probe 26 records `AUTHORIZATION_SURFACE=CLI-only (MCP not
    probed, not authorized)` and measured `backlogit … shipment claim`. Revision 2 prescribed
    MCP `backlogit_claim_shipment` as primary while citing P-13 as its "measured basis" —
    crediting one surface's measurement to a different surface. **S-2 is therefore pinned to
    the CLI surface Probe 26 actually measured.** If operational constraints force the MCP
    surface, that is a **known-unmeasured** path: S-3 and S-3b become the *only* guard, and
    that substitution must be **recorded in the run log**, not made silently.
  * **This is a one-way door (P-15): there is no unclaim.** Every §4 preflight gate must have
    passed before this line executes.
  * Record status ⇒ `active`.
* **S-3** — **Condition-5 verification (fail-closed, measured basis P-13).** Re-read `018-F`'s
  live status. **Pass criterion: exactly `active`.** Any other value ⇒ **HALT, return to
  Stage.** Ship performs **no** `queued -> active` write on a feature — the transition is
  performed by the claim itself, as **measured** in Probe 26 (`C1_ROOT_FEATURE_QUEUED_TO_ACTIVE
  =PASS`), not merely asserted by AC-19.
* **S-3b** — **Archived-member non-drift verification (fail-closed).** Re-read all **11**
  allowlisted members. **Pass criterion, per member:** still archive-located, `status` exactly
  `archived`, `archived_status` still present and unchanged, `parent_id` unchanged. **Any
  drift ⇒ HALT, return to Stage, invoke §7 rollback** — a claim that mutated archived members
  would have silently broken §3.2 condition 2 and the §3.3 exemption, and the next gate that
  would notice is a1 S0, *after* a merged PR. Probe 26 `C3=PASS` predicts no drift; S-3b makes
  the prediction falsifiable **at the cheapest possible unwind point**.
* **S-3a** — Topology gate `--phase post_claim` ⇒ exit 0 (`017-S` is the sole active shipment).

### Phase 2 — discharge conditions 1 and 2

* **S-4** — **Harness first (P-002/P-004).**
  * **Engine pre-emption disclosure (revision 2, measured P-13).** Probe 26
    (`C2_LIVE_TASK_QUEUED_TO_ACTIVE=PASS`) shows the **claim already transitioned
    `018.008-T` `queued -> active`**. Revision 1 required the red phase "before the task
    claim", which is **unachievable** — by S-4 the engine has already activated the task.
    The ordering prerequisite AC-20 protects is **red-before-implementation**, and that is
    preserved. **Restated requirement:** generate the failing harness for `018.008-T` ("Gate
    Stage artifact mutation behind a dedicated branch") and **observe and record the red
    phase before writing any implementation line (S-5)**. Ship performs no separate
    `queued -> active` write on `018.008-T`; if it re-reads as anything other than `active`
    here ⇒ **HALT**.
  * **Pass criterion (revision 2 — revision 1 had none):** the harness executes, **fails**,
    and the failure is attributable to the **absent `018.008-T` behavior** (not a compile
    error, missing fixture, or unrelated breakage). Record the failing test name and
    assertion message. A harness that passes on first run, or fails for an unrelated reason,
    ⇒ **HALT, return to Stage** — it does not establish red.
* **S-5** — Implement `018.008-T` to green. **Scope is exactly `018.008-T`'s recorded scope.**
  No scope expansion (PA-017 condition). Record the exact changed-path set as `{impl_paths}`.
* **S-6** — `018.008-T → done` (Step 4.5). **This discharges §3.2 conditions 1 and 2**: every
  descendant is now complete-or-exempt and no live descendant remains.
  * **`019.007-T` carve-out (revision 2).** `019.007-T` declares
    `dependencies: [018.008-T]` and is **outside** the manifest (parent `019-F`). Completing
    `018.008-T` may legitimately unblock it, so a **status change on `019.007-T` between S-6
    and S-11 is EXPECTED and is NOT drift.** The §7 R-6 untouched-assertion is therefore
    anchored to the **S-11 pre-cascade baseline**, never to the PF-5 preflight snapshot.
    Revision 1 anchored S-15 to PF-5 and R-6 to S-11 — two different baselines for the same
    assertion, which would have produced a false-positive unwind. **Both now use S-11.**
* **S-7** — PR lifecycle: open, review, operator-approved merge. P-018 engagement precedes the
  merge. **The merge MUST be a true merge commit (P-009): squash-merge and rebase-merge are
  FORBIDDEN.** Additionally assert the **exact changed-path allowlist**: the PR's changed-path
  set MUST equal `{impl_paths}` captured at S-5 — **no extra paths, no missing paths**. Any
  divergence is scope drift (PA-017 condition 9) ⇒ **HALT** per §7.1's S-7 row.
* **S-7a** — **Merge confirmation gate (revision 2 — revision 1 consumed `{merge_sha}` at S-14
  but never defined it).** After the S-7 merge completes, **bind the three values** that
  `mode: safe-close` and `backlogit shipment ship` later require, and verify the merge really
  landed:
  * `{merge_sha}` := **the PR-reported merge commit**, bound via
    `gh pr view <n> --json mergeCommit -q .mergeCommit.oid`. **NOT `git rev-parse
    origin/main`** — revision 2 used the remote tip, which equals the S-7 merge only if
    nothing else lands in between, and its pass criterion could never detect the error because
    **any later tip also contains the S-7 changes as an ancestor**. A wrong `{merge_sha}`
    propagates **irreversibly** into the destructive `shipment ship --sha`, into the archived
    `commit:` provenance checked at S-15.4, and into R-4's revert target.
  * **Verify, all three required:** (i) `git merge-base --is-ancestor {merge_sha} origin/main`
    succeeds; (ii) the S-7 branch head is an ancestor of `{merge_sha}`; (iii) every path in
    `{impl_paths}` (captured at S-5) appears in `git show --stat {merge_sha}`. This matches the
    installed contract's own Merge Confirmation Gate.
  * `{msg}` := the merge commit subject; `{author}` := the merge commit author.
  * **Measure the merge's parent shape — MERGE COMMIT ONLY (P-009, revision 4).**
    `git cat-file -p {merge_sha}` MUST report **exactly two parents**, with **parent 1 = the
    mainline (`origin/main`) commit** and **parent 2 = the reviewed PR head**. Verify both by
    `git merge-base --is-ancestor` against the pre-merge `origin/main` tip and against the S-7
    branch head respectively.
    **Any other shape — one parent (squash or rebase), three or more parents, or a parent
    order that does not match — is a P-009 VIOLATION ⇒ HALT** per §7.1's S-7a row.
    > **Revision-4 correction.** Revision 3 said a *"single-parent result is legitimate under a
    > squash/rebase strategy"*. **That is a direct P-009 violation** — P-009 is literally
    > *"Merge-Commit-Only (No Squash or Rebase Merge)"*. Tolerating a squashed merge here would
    > also have destroyed R-4's `-m 1` revert target. P-12 is repository *convention* and
    > remains corroborative only; **P-009 is the binding rule**, and S-7a enforces it.
  * **Pass criterion:** all of (i)–(iii) hold, the parent shape is exactly as specified, and all
    three values are recorded. Otherwise ⇒ **HALT** per §7.1's S-7a row.

### Phase 3 — post-merge closure (the a1 + cascade sequence)

* **S-8** — Post-merge branch protocol (P-011): create the closure branch **before any S2
  mutation**. No S2 mutation on `main`.
* **S-9** — **`a0`** topology gate `--phase lifecycle` ⇒ exit 0.
* **S-10** — **`a1` covering-feature completion gate.** Execute the **S0 → S5** selector in
  order, stopping at the first that applies:
  * **S0** (anomaly gate, evaluated **before** `n`): no member with missing/unresolvable
    `artifact_type`; none present in both queue and archive; none present in neither; no
    `status: archived` member with malformed provenance; **no `status: archived` member off
    the §3.3 allowlist**; enumeration complete and error-free. Any anomaly ⇒ **HALT, no
    feature mutation**. *(P-6/P-7 predict a clean S0; PF-4 re-verifies it.)*
  * `n` = count of manifest members whose **`artifact_type` is `feature`** — never by ID
    suffix. Here `n = 1` (`018-F`).
  * **S1** (`n == 0`) — not applicable. **S2** (`n > 1`) — not applicable.
  * **S3** — not applicable unless `018-F` already declares `done`.
  * **S4** — all five §3.2 conditions hold ⇒ transition **`018-F` `active -> done`** via
    `backlogit_move_item` (CLI fallback `backlogit move 018-F --status done`).
    **Exit-code contract (non-binary, fail-closed):** `0` = success; **`6` blocked,
    `7` configuration, `8` retryable-but-not-retried-here are each a HALT, never a retry.**
    A zero exit whose re-read is not exactly `done` is **also a HALT**. `--force-gates` /
    `--force-reason` are **forbidden** on this route.
  * **S5** — any other live status ⇒ **HALT**, record observed status, return to Stage.
* **S-11** — **Pre-cascade baseline commit (the §9 step 13a equivalent).** On the **existing**
  closure branch — **no second branch, no second worktree**:
  1. Stage **explicit paths only**: `.backlogit/queue/`, `.backlogit/archive/`.
  2. **Verify the stage is non-empty**: `git diff --cached --name-status` MUST list a1's
     `018-F` mutation. *(`git add` aborts the entire invocation on a single nonexistent
     pathspec, staging nothing; an unverified baseline commit silently destroys the rollback
     anchor.)*
  3. Commit. Designate that commit **`{pre_cascade_sha}`**.
  4. Verify clean `.backlogit/queue/` + `.backlogit/archive/` trees.
  5. Capture **`{pre_paths}`** — the full sorted path+blob-hash inventory of both directories
     **against `{pre_cascade_sha}`**, plus the raw frontmatter of `019.007-T`.
  **§7's restore uses exactly this baseline.**
* **S-12** — **Pre-archive reconciliation**, `mode: pre`, `expected_status: done`. Expected:
  `018-F` `done`, `018.008-T` `done`, the 11 `pre-archived`. **Continue only on `PROCEED`.**
* **S-13** — **`classify-close-path` (read-only, pre-mutation boundary).** Invoke
  `shipment-reconcile` `mode: classify-close-path` with `shipment_id: 017-S`. Read
  `CLOSE_PATH_VERDICT`, `VERDICT_REASON`, `CLASSIFICATION_BINDING`.
  * The verdict MUST be read **before any close call is made**.
  * Per D-3, this mode **MUST NOT** write a reconcile artifact or mutate any workspace/backlog
    state before returning its verdict; the five fields are returned as machine-consumable
    output.
  * Validate the token against the fixed allowlist `CASCADE` / `SAFE_CLOSE` / `BLOCK`.
    **HALT on `SAFE_CLOSE`, on `BLOCK`, and on any absent, empty, unparseable, or unnamed
    token.** Required verdict here: **`CASCADE`**.
* **S-14** — **Bound cascade close.** Invoke `mode: safe-close`, passing the
  `CLASSIFICATION_BINDING` from S-13.

> **AUTHORITY RECONCILIATION (revision 4) — the apparent contradiction is RESOLVED by an
> already-landed constraint, proven against current `origin/main` (`f2d4cf9`).**
>
> Revision 3 disclosed a divergence and simply *asserted* skill authority over the installed
> Ship contract. **That was not good enough** — a declaration of authority does not make
> contradictory installed instructions executable. Re-examined with hard evidence, the
> contradiction **dissolves**, and the resolution is the opposite of a Ship-file defect:
>
> | # | Fact | Evidence (current `origin/main`) |
> |---|---|---|
> | 1 | Installed Ship must invoke the named close path **"carrying the returned `CLASSIFICATION_BINDING` into that call"** | `_ship.agent.md` **L832**, restated for CASCADE at **L861–864** |
> | 2 | On `CASCADE` the named operation is `backlogit shipment ship` / `backlogit_ship_shipment` | `_ship.agent.md` **L861–864** |
> | 3 | **That operation accepts NO binding parameter — on EITHER surface** | CLI: `backlogit shipment ship --help` ⇒ flags are exactly `--sha`, `--message`, `--author`. MCP: `.autoharness/backlog-registry.yaml` **L187–194**, `ship_shipment.params` = `shipment_id`, `sha`, `message`, `author` |
> | 4 | ⇒ Read as a **direct** invocation, installed Ship's CASCADE branch is **internally unexecutable**: it mandates carrying a binding into a call that has no binding parameter | (1) ∧ (2) ∧ (3) |
> | 5 | The **only** surface in the installed system accepting `classification_binding` is `mode: safe-close` | `shipment-reconcile/SKILL.md` **L60** param table |
> | 6 | With a binding supplied, **"the bound verdict alone selects the branch: `CASCADE` enters the Cascade Close Sub-Procedure"** | `SKILL.md` **L595** |
> | 7 | On a `CASCADE` classification, safe-close steps 1–10 are **skipped entirely** and the Cascade Close Sub-Procedure runs instead — and that Sub-Procedure is what actually invokes the cascade op, under the `returned_ids`, two-set `allowed_ids`/`required_ids`, and `parent_id` gates | `SKILL.md` **L1108**, **L1111**, **L711** |
>
> **Conclusion:** `mode: safe-close` carrying a `CASCADE` binding is **not an alternative to**
> installed Ship's cascade branch — it is **the only executable realization of it**. The same
> `backlogit shipment ship` operation is invoked either way; the safe-close entry is precisely
> the binding-carrying wrapper that installed Ship **L832 itself requires**. This plan therefore
> **conforms to** the installed contract rather than overriding it, and **§6 condition 3 is
> satisfiable**, so `PA-017-CASCADE` is not invalidated.
>
> **No prerequisite Ship release unit is required.** Harmonizing the installed file would be a
> *clarity* improvement, not a correctness precondition — the binding-executability constraint
> already forces the single correct route. Ship attempting a **direct, unbound**
> `backlogit shipment ship` here is a **HALT**: it would discard the binding that L832 mandates
> and bypass the Sub-Procedure gates. **PF-10 makes this reconciliation falsifiable at preflight
> rather than trusting this prose** — if any of facts (1)–(7) has drifted, the route HALTs.

  Safe-close refuses unbound invocations (`RECONCILE_FAIL_CASCADE_UNBOUND`) and
  drift/invalid/ambiguous bindings. On a revalidated `CASCADE`, the Cascade Close
  Sub-Procedure performs the close via
  **`backlogit shipment ship 017-S --sha {merge_sha} --message {msg} --author {author}`**.
  * The generic shipment-status-move route is **not** the close path here and is not used
    (P-15: it refuses shipment status writes with **exit 9**).
  * Cascade step 2 HALTs on non-empty `returned_ids`; step 3's two-set gate HALTs on mismatch.
* **S-15** — **Postchecks**, all of which must pass **before** anything is committed:
  1. `mode: post` reconciliation ⇒ every manifest item present in archive.
  2. Capture **`{post_paths}`** (same inventory shape as `{pre_paths}`).
  3. **Cascade-scope assertion:** `{post_paths} △ {pre_paths}` is confined to the 13 manifest
     members and the `017-S` record. **`019.007-T` MUST be byte-identical to its `{pre_paths}`
     S-11 baseline capture** — *not* to its PF-5 preflight snapshot, which predates the S-6
     dependency-unblock carve-out. Any out-of-manifest artifact archived or deleted ⇒
     `HALT — cascade detected, revert required` ⇒ §7.
  4. Closure triple: `backlogit doctor` ⇒ `No issues found.` exit 0; `017-S` ⇒ `status:
     archived`; raw archive frontmatter ⇒ `archived_status: shipped` + `commit: {merge_sha}`.
* **S-16** — **Commit only after every postcheck passes.**
  * **Revision-2 reconciliation (revision 1 self-contradicted here).** Revision 1 declared the
    committed-cascade state *"unreachable"* while §7 R-2/R-4 simultaneously defined an
    **automatic** revert for exactly that state. Both cannot be true: an unreachable state
    needs no automatic handler, and an automatic handler implies a reachable state. The
    correct statement distinguishes **two different committed states**:
    * **Committed-with-postchecks-passed** — the **normal, reachable** result of S-16, and
      the state that persists into S-17. A defect discovered *after* this point (e.g. at the
      S-17 review, or after the closure merge) is a **genuine R-4 case**. R-4 is therefore a
      live, reachable path, not dead code.
    * **Committed-with-postchecks-failed-or-unrun** — **not produced by this sequence**,
      because S-16 is gated on S-15. If observed, it means something executed out of
      contract: **HALT to the operator**, apply **R-1 quarantine**, and take **no automatic
      revert** — an automatic unwind over state whose provenance is unknown is itself unsafe.
  * R-2's classify step now routes on this distinction explicitly.
* **S-17** — Closure PR, P-014 local review gate, operator-approved merge, verify on
  `origin/main`. **The merge MUST be a true merge commit (P-009)** — squash-merge and
  rebase-merge are FORBIDDEN here exactly as at S-7.

## 6. `PA-017-CASCADE` — escrow release and invocation conditions

`PA-017-CASCADE` is a **recorded operator approval held in escrow** (authorized
2026-09-13T11:35:43-07:00; `change_kind` **DESTRUCTIVE**; `ActionRisk` destructive;
`ActionResult` approved). **It does not expire and is not re-litigated here.** This plan
supplies the execution site that prerequisite plan §11.1 required, and **nothing else about
the approval is widened.**

**It remains NOT a claim authorization.** Claiming `017-S` is authorized by the shipment's own
eligibility, never by this approval.

**ALL conditions must hold at invocation time, re-measured by Ship — never inherited:**

| # | Condition | Gate |
|---|---|---|
| 1 | Escrow released by an existing, review-passed Stage closure plan supplying execution site, preflight gates, baseline and rollback | **this plan** + PF-1 |
| 2 | `classify-close-path` returns `CASCADE` **before any close call** | **S-13** |
| 3 | `safe-close` **revalidates the returned `CLASSIFICATION_BINDING` before any mutation** | **S-14** |
| 4 | `021-S` has shipped | **PF-8** |
| 5 | parent-completion capability merged and live | **PF-9** |
| 6 | manifest **unchanged at 13 members** | **PF-3** (and the PF-6 manifest invariant) |
| 7 | pre-cascade baseline **and the pre-invocation path inventory** exist and are verified | **S-11** (`{pre_cascade_sha}` + `{pre_paths}`) |
| 8 | validated linked-deliberation set for **this** manifest verified EMPTY **by this plan's own measurement** | **PF-5** (§2.1 re-measured over all 13) |
| 9 | **no scope expansion** | **S-5** scope lock; **PF-6** manifest invariant; §8 out-of-scope |

**Any mismatch INVALIDATES the approval and HALTs.**

> **Revision-2 correction to condition 7.** Revision 1 required *"both path inventories"* —
> including `{post_paths}`, which is captured at **S-15, after the S-14 invocation this very
> condition gates**. That made condition 7 **unsatisfiable at invocation time** and would have
> deadlocked the escrow release. Condition 7 is an **invocation-time** condition and now
> correctly names only what exists before S-14: `{pre_cascade_sha}` and `{pre_paths}`.
> `{post_paths}` remains **mandatory**, but as an **S-15 postcheck** (a condition of
> *committing* the result at S-16), not as a precondition of invoking it.

## 7. Rollback — verified, NON-DESTRUCTIVE, quarantine-then-revert

Per prerequisite plan §11.1 this plan defines **its own** unwind. The prerequisite plan's
rollback applies to `021-S` only and is **NOT reused by reference**.

**Prohibited unwind mechanisms (NON-NEGOTIABLE, scoped):** within **this plan's own rollback
procedure (R-1…R-7)**, no working-tree overwrite of backlog records from a prior commit, and no
file deletion. Restoration here is achieved **only** by adding commits, never by destroying
state. *(Needle scope: this prohibition names the mechanisms in order to forbid them; a
conformance check for their absence MUST exclude this plan file and MUST anchor to the
mechanism's invocation shape, not to a bare mention, or it will self-match this paragraph.)*

> **Revision-2 scope correction.** Revision 1 stated this ban **without scope**, which put it
> in direct conflict with the **installed** `_ship.agent.md` Step 6.1(b)/(c)/(d), whose
> **P-007-mandated** remedy for a reconcile-detected discrepancy is precisely
> `git restore .backlogit/archive/`. A blanket ban would have forbidden Ship from executing
> its own installed contract — and this plan has **no authority to override an installed
> policy**. The ban is therefore explicitly scoped to **R-1…R-7**, the procedure this plan
> owns. **Step 6.1(b)/(c)/(d) remains in force, unmodified**, and Ship executes it as
> installed; where that remedy applies, it takes precedence over this section. This plan's
> additive-only discipline governs **only** the post-cascade unwind it defines.

**Trigger for R-1…R-7 (cascade unwind only):** any **S-14/S-15** failure, any out-of-manifest
archival, or any postcheck failure. **Every HALT at S-0…S-13 is governed by §7.1 instead** —
R-1…R-7 presuppose cascade effects and a `{pre_cascade_sha}` anchor, and applying them earlier
would be incoherent.

### 7.1 Pre-baseline HALT disposition (revision 3 — rebuilt on a MEASURED release surface)

R-1…R-7 all unwind **to `{pre_cascade_sha}`**, which does not exist until **S-11**. Revision 1
defined **no rollback at all** for the claim-and-implement window. **Revision 2 filled that gap
with an operation that does not exist** — six of its seven rows said "release the claim via the
shipment's own unclaim/return path", an unmeasured surface. **Two independent reviewers flagged
this as the recidivist defect class, and the measurement (P-15) confirms it: there is no
unclaim path.**

**MEASURED GROUND TRUTH (P-15): a claimed shipment CANNOT be un-claimed.** The only transitions
out of `active` are `shipped` and `abandoned`.

> **`abandoned` IS FORBIDDEN ON THIS ROUTE (NON-NEGOTIABLE).** It is terminal and destructive,
> it is **not** covered by `PA-017-CASCADE` (whose scope is the cascade close, not shipment
> abandonment), and **no operator approval for abandoning `017-S` exists**. Improvising an
> abandon as a "rollback" would be an unapproved destructive act — strictly worse than the
> HALT it purports to remedy. Any abandon requires a **new, explicit operator approval** and is
> outside this plan.

**Therefore the correct disposition for every pre-baseline HALT is the same, and it is not a
rollback at all:** `017-S` **stays `active`**, nothing further is mutated, and the session
**HALTs to the operator**. This is honest rather than convenient — it means a HALT in this
window leaves `017-S` **occupying the single-active-release-unit slot (P-001)** until an
operator acts. **That cost is disclosed, not hidden**, and it is the price of a backlog engine
with no unclaim path. It is **not** grounds to improvise a destructive exit.

**Universal pre-baseline HALT procedure — TWO REGIMES, decided by ONE question:
is `017-S` live-status `active`?**

**Regime A — `017-S` is NOT `active`** (pre-claim, or a failed claim that left it `queued`).
No claim is held, so nothing is occupied:
1. Mutate nothing further. Record the failure token verbatim and the step ID.
2. **HALT to Stage.** No unwind is required and none is performed.

**Regime B — `017-S` IS `active`** (any HALT at S-3…S-13, or a failed claim that nonetheless
left it `active`). **The claim is a one-way door (P-15):**
1. **Stop.** Mutate no further backlog record; do not proceed to the next step.
2. Record the **failure token verbatim**, the step ID, and the live status of `017-S`, `018-F`,
   `018.008-T`, and the 11 allowlisted members.
3. **Leave `017-S` `active`. Perform NO backlog mutation of any kind.** Specifically
   **FORBIDDEN**: `shipment return-blocked` (returns an item ⇒ breaks the 13-member manifest),
   any `backlogit move` on the shipment (refused, exit 9), `abandoned` (terminal, destructive,
   **unapproved** — outside `PA-017-CASCADE`), and **R-1…R-7** (they presuppose cascade effects
   and a `{pre_cascade_sha}` anchor that may not exist).
4. **HALT to the operator** with the record from (2). Disposition of the claim is the
   **operator's** decision, informed by Stage re-measurement — never Ship's improvisation.
5. `PA-017-CASCADE` is **not** invoked and remains **unexercised and intact**.

**Disclosed cost:** in Regime B, `017-S` occupies the **P-001 single-active-release-unit slot**
until an operator acts. This is stated rather than engineered around, because every available
workaround is either manifest-breaking or an unapproved destructive act.

| HALT at | State | Regime + row-specific addition |
|---|---|---|
| **S-0 / PF-0…PF-10** | **no claim performed** | **Regime A.** Nothing to unwind. **HALT to Stage** — no live state is held. This is the only cheap-exit window, which is precisely why §4's gates are exhaustive |
| **S-1** (reconcile fail, pre-claim) | **no claim performed** | **Regime A.** S-1 runs strictly **before** S-2, so no claim can exist. **HALT to Stage** |
| **S-2** (the claim itself fails) | claim may be **partial** | **Re-read `017-S` live status — this is the regime-deciding question.** Still `queued` ⇒ **Regime A**, HALT to Stage. `active` ⇒ **Regime B**, plus run S-3b's non-drift check to characterize what the partial claim touched |
| **S-3** (`018-F` not `active`) | claim active; no code, no Ship record write | **Regime B.** Contradicts P-13 `C1` ⇒ **Stage re-measurement trigger** |
| **S-3b** (archived-member drift) | claim mutated archived members | **Regime B**, **plus**: quarantine the drifted records to `quarantine/017-S-<utc>` by **additive commit** before halting. **This is S-3b's own disposition — NOT §7 R-1…R-7**, which presuppose cascade effects that have not occurred. Contradicts P-13 `C3` ⇒ the plan's evidentiary basis is void; **Stage must re-measure before any re-route** |
| **S-3a** (post-claim topology gate) | claim active | **Regime B** |
| **S-4** (red not established) | claim active; no implementation written | Discard the harness working tree **on the pre-claim branch only** — ordinary pre-commit iteration, **not** a backlog-record unwind, and therefore outside §7's additive-only ban. Then **Regime B** |
| **S-5 / S-6** (implementation or `done` transition fails) | claim active; code on the pre-claim branch, **unmerged** | Leave the branch intact as evidence. **Do not merge.** Nothing is on `main`, so no revert is required. **Regime B** |
| **S-7** (PR not merged, or P-009 shape violation, or `{impl_paths}` allowlist mismatch) | claim active; PR open or improperly merged | Close or park the PR. A **squash/rebase merge is a P-009 violation** ⇒ escalate to the operator; do not proceed. **Regime B** |
| **S-7a** (`{merge_sha}` unbindable) | **merged** | **Do not proceed to S-8** — every later step consumes `{merge_sha}`. The merge itself **stands** (legitimate, reviewed work). **Regime B** |
| **S-8 / S-9** (branch or a0 gate fails) | merged impl on `main`; **`018-F` not yet mutated** | Park the closure branch. **Regime B** |
| **S-10** (a1 fails, possibly mid-mutation) | merged impl on `main`; `018-F` **possibly mutated, UNCOMMITTED** | **Revision-3 correction.** Revision 2 said "revert only that record mutation by additive commit" — **impossible**: the baseline commit is S-11, so at S-10 the a1 mutation is **uncommitted** and there is no commit to revert; and a backlogit `done -> active` write is **outside the §3.2 grant** (which authorizes exactly one `active -> done`) and outside Ship's Role Boundary. **Correct disposition:** the closure branch is **unmerged**, so **nothing needs reverting** — park the branch, leave the record as-is, **Regime B**. Where the discrepancy is one the **installed** Step 6.1(c) remedy addresses, Ship executes that remedy **as installed** (§7's ban is explicitly scoped away from it) |
| **S-11** (baseline commit or its staged-diff verification fails) | `018-F` moved to `done` by a1; **baseline anchor NOT established** | **The rollback anchor does not exist** — R-1…R-7 are unavailable from here. Do **not** proceed to S-12. **Regime B**, escalated: this is the HIGH-rated anchor-destroying case in §10 |
| **S-12** (`mode: pre` fail at `expected_status: done`) | claim active; `018-F` `done`; baseline committed | Baseline exists but **no cascade has run**, so R-1…R-7 do not apply. **Regime B** |
| **S-13** (verdict `SAFE_CLOSE`, `BLOCK`, absent, or unparseable) | claim active; `018-F` `done`; baseline committed; **zero mutation** (D-3) | `classify-close-path` is read-only, so **nothing has been mutated by it**. This is a realistic fail-closed outcome, **not** an error. **Regime B**. `PA-017-CASCADE` is **not** invoked |

**Invariant across the entire table:** `PA-017-CASCADE` has **not** been invoked at any point
above, **no destructive action is unwound**, and the escrow remains **unexercised and intact**.
**R-1…R-7 below apply only from S-14 onward**, once a cascade has actually run.

| Step | Action |
|---|---|
| **R-1 QUARANTINE** | **Stop. Mutate nothing further.** Do not commit the cascade result. Record the observed failure token verbatim |
| **R-2 CLASSIFY** | Route on the S-16 distinction: **(a) uncommitted** ⇒ R-3. **(b) committed with S-15 postchecks passed** (defect found later, e.g. at S-17) ⇒ R-4 — the reachable, automatic path. **(c) committed with postchecks failed or unrun** ⇒ **out of contract: HALT to the operator after R-1 quarantine, no automatic revert** |
| **R-3 UNCOMMITTED PATH** | The tree still carries only uncommitted cascade effects. Capture them to a **quarantine branch** (`quarantine/017-S-<utc>`) via an **additive commit** so the evidence survives, then return the closure branch to `{pre_cascade_sha}` **by adding a revert commit**, never by overwriting or deleting. Proceed to R-5 |
| **R-4 COMMITTED PATH** | **P-009 merge-commit-only (revision 4).** The revert target is a merge commit **by construction** — S-7a already HALTed the route if `{merge_sha}` was not a two-parent merge with parent 1 = mainline. Re-verify that shape (`git cat-file -p <sha>` ⇒ exactly two parents, parent 1 = mainline) and revert with **`git revert -m 1 <sha>`**. **Any other shape at this point is a P-009 violation and an out-of-contract state ⇒ HALT to the operator, no automatic revert.** *(Revision 3 carried a plain-revert branch for single-parent commits; that branch is **REMOVED** — under P-009 a squashed or rebased merge cannot legitimately exist on this route, so silently accommodating one would have masked a policy violation instead of surfacing it.)* |
| **R-5 TREE EQUIVALENCE** | Recompute the `{pre_paths}` inventory over `.backlogit/queue/` + `.backlogit/archive/` and require it **set-equal, path-for-path and blob-hash-for-blob-hash**, to `{pre_paths}` captured at S-11 |
| **R-6 RECORD EQUIVALENCE** | Tree equality is **not sufficient**. For all 13 members plus `017-S` plus `019.007-T`, re-read **raw frontmatter from the artifact files** — never a list view, never an index/query view — and require `status`, `archived_status`, `archived_from`, `parent_id`, `artifact_type` and `commit` to match their **S-11** baseline values exactly. **`019.007-T` is anchored to S-11, matching S-15** — both now use the same baseline (revision 1 used S-11 here and PF-5 at S-15, guaranteeing a false-positive unwind whenever S-6 legitimately unblocked it). Then `backlogit sync` and require `backlogit doctor` ⇒ `No issues found.`, exit 0 |
| **R-7 VERIFY-OR-ESCALATE** | **A restore is not complete until R-5 AND R-6 both pass.** If either fails, **HALT to the operator** with the quarantine branch name, `{pre_cascade_sha}`, and both inventories. Do not attempt a second automated unwind |

**Post-restore disposition:** the shipment returns to Stage. `PA-017-CASCADE` is **not**
re-exercised in the same session.

## 8. Out of scope (explicit)

* **Claiming `017-S`.** Stage does not claim; this plan authorizes and sequences only.
* **Implementing `018.008-T`.** Ship work, executed at S-5 under this plan's scope lock.
* **Widening `PA-017-CASCADE`**, or re-litigating the operator approval.
* **Amending P-015, P-010, `_ship.agent.md` or `shipment-reconcile/SKILL.md`.** This plan
  creates no new verdict, token, classifier or policy surface.
* **Adding any backlog item to `018-F` or `017-S`** — forbidden by the manifest invariant.
* **Decision §7 follow-ups 1, 2, 3 and 5** — recorded, off-route, and explicitly *"not current
  blockers"*. §7 item **4 alone** is pulled in, as **PF-6**.
* **Re-opening prerequisite plan attempt 9.** Revision 13 stays **FROZEN**; the attempt-8
  **FAIL STANDS**.

## 9. Acceptance criteria

1. **AC-1** **PF-0…PF-10** all pass before the claim; any failure halts with **no mutation**.
2. **AC-2** **PF-6 is a READ-ONLY re-verification gate**, not a probe-authoring mandate: it
   re-verifies two **already-committed** artifacts — Probe 25
   (`probe25-root-included-cascade-fixture.txt` @ `e36d853`) and Probe 26
   (`docs/plans/evidence/2026-09-16-017-S-closure/probe26-claim-cascade-root-included.txt`) —
   **plus the committed Probe 26 script**
   (`docs/plans/evidence/2026-09-16-017-S-closure/probe26.ps1`), and confirms their engine
   agreement. **This plan authors no probe, no test and no backlog item**, and PF-6 adds no
   manifest member.
3. **AC-3** `018-F` is `queued` at S-1, **exactly `active` at S-10 entry**, and **exactly
   `done` on S-10 success**; **no agent performs a `queued -> active` write on it** — the
   claim performs that transition (measured, P-13 `C1`).
4. **AC-4** S-10 executes S0→S5 in order; `n` is computed by `artifact_type`, never by ID
   suffix; `move` exit codes 6/7/8 each HALT without retry.
5. **AC-5** `{pre_cascade_sha}` is committed on the **existing** closure branch with a
   **verified non-empty** staged diff containing a1's mutation.
6. **AC-6** `CLOSE_PATH_VERDICT` is read **before** any close call, and `safe-close`
   revalidates the binding before any mutation.
7. **AC-7** The close is performed by the cascade command of P-11 — the generic
   shipment-status-move route is never invoked on `017-S`.
8. **AC-8** The cascade result is **not committed until every S-15 postcheck passes**.
9. **AC-9** `019.007-T` is **raw-byte identical between its S-11 `{pre_paths}` capture and the S-15 postcheck** (and again at R-6 if an unwind runs). **The baseline is S-11, never PF-5** — PF-5's earlier value is informational only, because the S-6 dependency carve-out permits `019.007-T` to change legitimately between PF-5 and S-11.
10. **AC-10** The rollback verifies merge-commit shape **empirically** before any `-m` revert,
    and completes only when **both** tree equivalence (R-5) and record equivalence (R-6) pass.
11. **AC-11** All nine `PA-017-CASCADE` conditions are re-measured at invocation; none is
    inherited from this plan's Stage-time measurements.
12. **AC-12** The manifest is **13 members at every point**; set equality holds at PF-3 and at
    S-10 S0.
13. **AC-13 (ONE-WAY DOOR).** No step attempts to un-claim, release, or return `017-S`, and no
    step invokes `shipment return-blocked`, a `backlogit move` on the shipment, or `abandoned`.
    On **every** HALT at S-3…S-13 the shipment is left **`active`** with **zero backlog
    mutation** and the session halts **to the operator** (§7.1 Regime B). A HALT before the
    claim, or after a failed claim that left `017-S` `queued`, halts **to Stage** (Regime A).
14. **AC-14 (P-009 MERGE-COMMIT-ONLY).** Both the S-7 and S-17 merges are **true merge
    commits** — squash-merge and rebase-merge are forbidden. `{merge_sha}` has **exactly two
    parents**, parent 1 = mainline and parent 2 = the reviewed PR head, verified at S-7a; any
    other shape HALTs as a P-009 violation, and R-4 carries **no** plain-revert branch.
15. **AC-15 (CLOSE-PATH AUTHORITY).** PF-10 confirms the cascade operation exposes **no**
    binding parameter on **either** the CLI or MCP surface, so `mode: safe-close` carrying a
    `CASCADE` binding remains the **only executable realization** of installed Ship's
    binding-carrying mandate. If a binding parameter appears on either surface, the S-14
    reconciliation is **VOID** and the route HALTs to Stage.
16. **AC-16 (SCOPE ALLOWLIST).** The S-7 PR's changed-path set **equals** `{impl_paths}`
    captured at S-5 — no extra paths, no missing paths.

## 10. Risk register

| Risk | Severity | Mitigation |
|---|---|---|
| Claim does not set `018-F` to `active` | **HIGH** — blocks condition 5 with no in-band repair | **S-3** fail-closed verification; HALT to Stage. Deliberately not "repaired" by a direct write, which would violate §3.2's grant shape |
| Root-included cascade shape never engine-verified | **MEDIUM** | **PF-6** converts the prediction into evidence before routing; route is fail-closed at two gates |
| Baseline commit silently stages nothing | **HIGH** — destroys the rollback anchor | **S-11 step 2** mandatory non-empty staged-diff verification |
| `git revert -m 1` applied to a non-merge commit | **MEDIUM** | **R-4** counts parents empirically first; `-m` forbidden on a 1-parent commit |
| Tree restored but records semantically drifted | **MEDIUM** | **R-6** record equivalence read from raw frontmatter, not list/index views |
| Stale `--id` flag form in inherited probes | **LOW** | §2.2 records the drift; PF-5 pins the positional form |
| Off-allowlist archived descendant appears later | **MEDIUM** | **PF-4** + **S-10 S0** both HALT on any off-list archived descendant |

## 11. Requires plan hardening

**Requires plan hardening: yes.** This plan supplies the execution site for a **DESTRUCTIVE**,
operator-approved cascade over a 13-member manifest, defines a net-new rollback for which the
compound library holds **no prior learning**, and depends on a claim-time state transition
(`018-F queued -> active`) that no agent is permitted to repair in band.

## Plan Hardening

**Hardening required: YES.** Triggers present: (a) an irreversible/destructive action
(`PA-017-CASCADE` archives a 13-member manifest subtree); (b) an operator-approval boundary
held in escrow; (c) high blast radius over backlog state that no build step can reconstruct;
(d) a rollback path with **zero prior compound learning**; (e) a claim-time state transition
with no in-band repair.

### H.1 Protected invariants

| # | Invariant | Violation consequence |
|---|---|---|
| **I-1** | Manifest is **exactly 13 members** from PF-3 through S-14 | Invalidates `PA-017-CASCADE` condition 6 and §3.2 condition 3 |
| **I-2** | The §3.3 allowlist is **exactly 11 IDs**; no archived descendant is off-list | S0 anomaly ⇒ HALT |
| **I-3** | `018-F` is `queued` at S-1, **`active`** at S-10, `done` only via S4 | Pre-claim `RECONCILE_FAIL`, or an ungranted feature write |
| **I-4** | No mutation precedes `CLOSE_PATH_VERDICT` being read | Re-opens the K-1 defect class (verdict recorded only after the destructive call) |
| **I-5** | `{pre_cascade_sha}` exists with a **verified non-empty** staged diff | Rollback anchor destroyed |
| **I-6** | Nothing outside the manifest is archived or deleted | `HALT — cascade detected, revert required` |
| **I-7** | Revision 13 of the prerequisite plan stays **FROZEN**; attempt 8 **FAIL STANDS** | Unauthorized attempt 9 |

### H.2 Risky actions (strict-safety vocabulary)

| ProposedAction | ActionRisk | ActionResult | Approval |
|---|---|---|---|
| Cascade close of `017-S` (archives the 13-member subtree) via the P-11 command at **S-14** | **destructive** | **approved** — `PA-017-CASCADE`, operator-authorized 2026-09-13T11:35:43-07:00 | Pre-approved in escrow; **released by this plan**. All nine §6 conditions re-measured at invocation. Any mismatch ⇒ approval **INVALIDATED**, HALT |
| `018-F` `active -> done` at **S-10/S4** | **moderate** — single reversible status write under a five-condition conjunctive gate | pending | Covered by the §3.2 Role Boundary grant; no separate approval |
| Revert/quarantine commits at **R-3/R-4** | **low** — purely additive; no overwrite, no deletion | pending | No approval needed; **R-7 escalates to the operator on any verification failure** |
| Implementation of `018.008-T` at **S-5** | **low** — scope-locked | pending | Ordinary task execution |

> **Note.** `strict_safety.enabled` is `false` in `.autoharness/config.yaml`, so no automated
> approval broker gates these actions. The classification is recorded explicitly anyway;
> `PA-017-CASCADE`'s approval is a **recorded operator decision**, not a broker artifact, and
> is unaffected by that flag.

### H.3 Learnings and instructions consulted

* `docs/compound/workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md`
  — **same shipment lineage**; documents **seven consecutive plan-review FAILs** on one defect
  class. Applied: this plan never cites a producer-scoped or reporting-artifact-scoped count as
  closure evidence; §2's probe table **names the surface each measurement covers** and quotes
  verbatim output. Supplied the `move` exit-6/7/8 fact hardened into **S-10/S4**.
* `docs/compound/2026-05-07-backlogit-shipment-status-constraints.md` — the cascade/safe-close
  split and the **exit-9 refusal** of the generic shipment-status-move route. Applied at
  **P-11**, **S-14** and **AC-7**. Also supplied the *location ≠ declared status* rule, which is
  why **PF-4 keys on declared `status`**, never on archive-directory membership.
* `docs/compound/2026-09-04-backlogit-sizing-is-wit-gated.md` — `git add` aborts the whole
  invocation on one bad pathspec, and `2>$null` manufactures fictional success. Applied at
  **S-11 step 2** (mandatory staged-diff verification) and **R-6** (read raw frontmatter, not
  list/index views).
* `docs/compound/2026-09-08-adversarial-review-empirical-verification-and-scope-check.md` —
  **strong reading applied**: load-bearing claims about installed behaviour are independently
  reproduced. Hence **P-12** measures merge-commit parentage empirically rather than assuming
  `-m 1`, and **PF-9** verifies capability liveness against the **installed** files.
* `docs/compound/2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md` — applied
  as the explicit **needle-scope caveat** in §7's prohibition paragraph.
* Instructions re-read: `backlogit.instructions.md`, `concurrency.instructions.md`,
  `strict-safety.instructions.md`, `agent-engram.instructions.md`.

**Gap carried forward (disclosed):** the compound library holds **no** prior learning on
`git revert -m 1` mainline selection, quarantine-then-revert, tree equivalence, or record
equivalence. §7 is **net-new ground**. It is therefore specified at step granularity with
dual verification (R-5 **and** R-6) and a fail-closed escalation (R-7), and a compound entry
**MUST** be harvested after first execution.

### H.4 Reinforced verification and operational closure

* **Environment prechecks:** PF-2 (`doctor` exit 0), PF-7 (topology `pre_claim`), PF-9
  (capability liveness on the installed files), plus the backlogit version pin — closure was
  designed against **1.10.1**; a different resolved engine identity is a **HALT**, since §2.2
  already records one flag-surface drift (`--id` removal) on this exact probe path.
* **Blocked-path handling:** every gate names its HALT token. `move` 6/7/8 never retries.
  `SAFE_CLOSE` / `BLOCK` / unparseable verdicts HALT rather than fall through — the
  unconditional safe-close default was removed by `0cb4a42` and MUST NOT be reintroduced.
* **Monitoring signals:** the S-15 closure triple (`doctor`; `017-S` ⇒ `archived`; frontmatter
  ⇒ `archived_status: shipped` + `commit`), plus the `{post_paths} △ {pre_paths}` scope
  assertion and the `019.007-T` raw-byte identity check **anchored at S-11** (never at PF-5).
* **Rollback trigger:** any S-14/S-15 failure, any out-of-manifest archival, any postcheck
  failure. **Owner: Ship**, escalating to the **operator** at R-7.
* **Validation window:** closure is not final until S-17 verifies the closure PR merged and
  present on `origin/main`.
* **Human checkpoints:** operator-approved merge at S-7 and S-17; operator escalation at R-7;
  and any PF/S HALT returns to **Stage**, not to an autonomous retry.

### H.5 Review-gate capability risks (carried into plan review)

* Plan review **MUST** emit literal `dispatch_mode:` and `decision:` markers.
* If review runs degraded (P-012 — unavailable persona/provider), the degradation **MUST** be
  declared in the verdict, and a degraded review **MUST NOT** be recorded as a clean PASS.
* This plan **inherits no gate credit** from the prerequisite plan. The attempt-8 **FAIL
  STANDS** against revision 13 and confers nothing here; this plan's gate is its own, at
  attempt 1.
* Given H.3's seven-FAIL history on this lineage, reviewers **MUST** check the specific defect
  class: every closure claim must name the surface its evidence covers, and no site may be
  credited by evidence belonging to another site.

---

## Plan Review

<!-- plan-review-attempt: 1 -->

### Attempt 1 — revision 1

dispatch_mode: multi-agent
reviewers: correctness-reviewer, scope-boundary-auditor
decision: FAIL

**Findings:** 36 total. Correctness — P0: 1, P1: 7, P2: 9, P3: 1. Scope — P0: 0, P1: 3,
P2: 11, P3: 4. The gate FAILED on the presence of a P0.

**P0 (Correctness F-1) — unmeasured claim-cascade behaviour over archived members.**
The plan's entire sequence depended on the claim transitioning `018-F` `queued -> active`,
but the only supporting evidence came from `021-S`, whose manifest was **2 members, all
`queued`, zero archived**. What the claim does to **11 `status: archived` members** was
never measured. Had it flipped any to `active`, §3.2 condition 2 and the §3.3 pre-archived
exemption would both break, leaving 11 `active` records inside `archive/` — and the first
gate that would notice is a1 S0, **after a merged PR**.

**Disposition — CLOSED by direct measurement.** Probe 26
(`docs/plans/evidence/2026-09-16-017-S-closure/probe26-claim-cascade-root-included.txt`)
built the exact 13-member root-included shape with 11 genuinely archived members
(`FIXTURE_VALIDITY_GATE=PASS`) and measured the real engine:
`C1_ROOT_FEATURE_QUEUED_TO_ACTIVE=PASS`, `C3_ARCHIVED_MEMBERS_UNCHANGED_BY_CLAIM=PASS`,
`FAILED_CRITERIA=0`, `PROBE26_RESULT=PASS`. Recorded as **P-13**; S-3's basis is now
measured rather than inherited from AC-19, and **S-3b** was added to make the non-drift
prediction falsifiable at the cheapest unwind point.

**A false FAIL was produced and discarded en route.** The first Probe 26 run archived its
fixtures with `backlogit move <id> --status archived`, which is **not an archive
operation**; it produced `PRE_ARCHIVED=0` and therefore measured a plain all-queued
manifest while appearing to measure the archived shape, emitting a spurious `C3=FAIL`.
That run was **discarded, not reported**. The correct command is `backlogit archive <id>`
(recorded as **P-14**), and a **fixture-validity gate** now aborts the probe before the
measurement whenever the fixture has not reached the target shape — so a broken fixture can
never again masquerade as a measurement.

**P1 cluster and disposition (all addressed in revision 2):**

| # | Finding | Revision-2 disposition |
|---|---|---|
| C F-2 | S-4's "red before task claim" is unachievable — the claim already activated `018.008-T`; S-3 had no pass criterion | S-4 restated to **red-before-implementation** with the engine pre-emption disclosed (P-13 `C2`); explicit pass criterion added |
| C F-3 | **No rollback defined for any HALT between S-2 (claim) and S-11 (baseline)** — R-1…R-7 all unwind to a `{pre_cascade_sha}` that does not yet exist | New **§7.1 pre-baseline HALT disposition table** covering S-1 → S-10, with the invariant that `PA-017-CASCADE` is unexercised throughout |
| C F-4 | `019.007-T` byte-identity anchored to **PF-5** at S-15 but to **S-11** at R-6 — two baselines for one assertion; `019.007-T` depends on `018.008-T`, so S-6 may legitimately change it ⇒ guaranteed false-positive unwind | Both anchored to **S-11**; explicit S-6 dependency carve-out added |
| C F-6 / S F-3 | "closure branch" named two different branches; PF gates referenced a branch not created until S-8 | **Branch vocabulary** section added: **pre-claim branch** (S-0, carries preflight + claim + impl) vs **closure branch** (S-8, carries S2 mutations) |
| C F-7 | §7's blanket ban on `git restore`/deletion **forbade the remedy installed Ship Step 6.1(b)/(c)/(d) mandates** under P-007 | Ban **scoped to R-1…R-7 only**; installed Step 6.1(b)/(c)/(d) explicitly preserved and given precedence where it applies |
| C F-8 | `{merge_sha}`/`{msg}`/`{author}` consumed at S-14 but never defined; no merge-confirmation gate; `mode: safe-close` requires `merge_commit_sha` | New **S-7a merge confirmation gate** binds all three from `origin/main` and measures the merge's real parent shape |
| C F-11 | §6 condition 7 unsatisfiable — required `{post_paths}`, captured at S-15, **after** the S-14 invocation it gates | Condition 7 restated to invocation-time artifacts only (`{pre_cascade_sha}` + `{pre_paths}`); `{post_paths}` retained as an S-15 postcheck gating S-16 |
| C F-13 / S F-8 | S-16 declared the committed-cascade state "unreachable" while R-2/R-4 defined an automatic revert for it — direct contradiction | S-16 now distinguishes **committed-postchecks-passed** (reachable, R-4 applies) from **committed-postchecks-failed/unrun** (out of contract, operator HALT, no automatic revert); R-2 routes on the distinction |
| C F-14 | PF-8 accepted `021-S` `status: archived`, but `archived` is also the terminal state of an **abandoned** shipment | PF-8 strengthened to require `021-S`'s own frontmatter `archived_status: shipped` plus commit/merge/PR cross-check |
| S F-1 / F-2 | PF-6 mandated **authoring** a new behavioural test whose premise was stale (Probe 25 already exists, committed `e36d853`) and which **no agent in the pipeline was permitted to execute** — Stage may not write tests, Ship may not author out-of-scope tests ⇒ the route would deadlock at its first preflight | PF-6 rewritten as **read-only re-verification** of two committed artifacts (PF-6a Probe 25, PF-6b Probe 26, PF-6c engine agreement). Ground 4 added to §1 and recorded as discharged |

**Self-criticism recorded.** Revision 1 applied "measured truth, not stale premise" rigor to
ineligibility grounds 1–3 and then **failed to apply it to its own preflight gate**, importing
decision §7 item 4's premise without checking whether the probe it demanded already existed.
It did. This is the same defect class that produced seven consecutive FAILs on this lineage
(H.3) — crediting or demanding a surface without measuring the surface itself.

### Attempt 2 — revision 2

<!-- plan-review-attempt: 2 -->

dispatch_mode: multi-agent
reviewers: correctness-reviewer, scope-boundary-auditor
decision: FAIL

**Scope dimension: PASS** — 0 P0, 2 P1, 7 P2, 6 P3. All three attempt-1 scope P1 dispositions
verified REAL (PF-6 genuinely read-only; branch vocabulary genuinely resolved; S-16/R-2
contradiction genuinely resolved). No manifest change, no backlog item added, no widening of
`PA-017-CASCADE`, no re-litigation of the frozen prerequisite.

**Correctness dimension: FAIL** — 1 P0, 7 P1, 7 P2, 6 P3.

**P0 — §7.1 discharged revision 1's rollback gap with an operation that DOES NOT EXIST.**
Six of its seven rows prescribed "release the claim via the shipment's own unclaim/return
path". No command was named and no measurement backed it. **Both reviewers flagged this
independently** (correctness P0, scope P1 #1) — convergence that made it unignorable.

**This was a relapse into the exact defect class that has FAILed this lineage seven times:**
a closure claim credited against a surface that was never measured. Revision 2's own Plan
Review section criticised revision 1 for precisely this, and then committed it again while
fixing it. Recorded without mitigation.

**MEASURED (now P-15):** `backlogit shipment --help` ⇒ the complete subcommand set is
`add | claim | create | get | list | return-blocked | ship`. **There is no `unclaim`,
`release`, or `return`.** `return-blocked` returns a single **ITEM** from a shipment, mutating
the manifest — using it here would break the 13-member invariant outright. Corroborated by the
installed `_ship.agent.md` (shipment transitions are `queued -> active`, `active -> shipped`,
`active -> abandoned` only) and by the compound entry recording that generic `move` refuses
shipment status writes with **exit 9**. **A claimed shipment cannot be un-claimed.**

**Revision-3 disposition:** §7.1 rebuilt on the measurement. Every pre-baseline HALT now leaves
`017-S` **`active`** and HALTs to the **operator**; `abandoned` is explicitly **FORBIDDEN**
without a new operator approval (it is terminal, destructive, and outside `PA-017-CASCADE`'s
scope). The real cost — a HALT in this window occupies the P-001 single-active slot until an
operator acts — is **disclosed rather than hidden**, because the alternative was improvising an
unapproved destructive exit. S-2 now carries a one-way-door warning.

**Correctness P1 cluster and revision-3 disposition:**

| # | Finding | Disposition |
|---|---|---|
| F-1b | §7.1 did not cover the window it claimed — no disposition for S-2, S-11, S-12, S-13 | Table extended to **S-0 through S-13**, every row given a universal procedure plus row-specific additions; §7's Trigger clause explicitly scoped to S-14 onward |
| F-2b | S-7a bound `{merge_sha}` as `git rev-parse origin/main` (the remote **tip**, not the merge); its criterion could never detect the error since any later tip also contains S-7 as an ancestor. Propagates irreversibly into the destructive `--sha`, the archived `commit:` provenance, and R-4's revert target | Bound from the **PR** (`gh pr view --json mergeCommit`), verified with `git merge-base --is-ancestor`, S-7 head ancestry, and `{impl_paths}` ⊆ `git show --stat` |
| F-3b | §4.1 **over-claimed Probe 25** — credited it with `classify-close-path ⇒ CASCADE`, zero-mutation classification, and an off-allowlist HALT. Probe 25 ran at `PROBE_HEAD=4fd21c5`, **before that mode existed**; it computed criteria inside the probe script and never invoked the installed boundary. Three of four claims unsupported | Coverage restated to exactly what each artifact records. Added an explicit **DISCLOSURE** that the bound `classify-close-path -> safe-close -> cascade` route is **UNMEASURED** on the post-`021-S` engine and is protected only by S-13's fail-closed HALT and S-14's refusal tokens — a **guard, not evidence** |
| F-4b | S-14's `mode: safe-close` route **contradicts the installed** Step 6.1(b), which calls the destructive op directly. Under the installed reading §6 condition 3 is unsatisfiable, which by §6's own rule **invalidates `PA-017-CASCADE`** | Explicit **divergence disclosure** added at S-14 pinning the SKILL/§3.4.1(5) route and stating that a direct `shipment ship` invocation is a **HALT** on this route |
| F-5b | P-13/S-3 surface mismatch — Probe 26 records `AUTHORIZATION_SURFACE=CLI-only`, but S-2 prescribed **MCP** as primary | S-2 **pinned to the CLI surface** Probe 26 measured; any MCP substitution is flagged known-unmeasured and must be recorded in the run log |
| F-6b | PF-6b had no provenance pin (PF-6a pins `e36d853`); satisfiable from a local unreviewed file | PF-6b now requires `git show origin/main:<path>` for both the artifact **and** the generating script |
| F-7b | §7.1's S-10 row said "revert the record mutation by additive commit" — **impossible** (the a1 mutation is uncommitted at S-10; baseline is S-11) and a `done -> active` write is outside the §3.2 grant | Corrected: the closure branch is unmerged, so **nothing needs reverting** — park it; where Step 6.1(c)'s installed remedy applies, Ship executes it as installed |

### Escalation (P-013.6) — revision 3 assessed, revision 4 authored

route: gpt-5.6-sol / openai / high (distinct from Stage's role route `claude-opus-5` ⇒ NOT degraded)
mutation_by_escalation: none
verdict: **attempt 3 NOT justified on revision 3** — close the findings below first.

**P0-1 — CASCADE authority contradiction. RESOLVED BY EVIDENCE, NOT BY HARMONIZATION.**
The escalation held that the installed Ship contract must be harmonized on `origin/main` before
any claim, and that a prerequisite Ship release unit might be required. **Assessed under Stage
authority against `origin/main` (`f2d4cf9`), that is not necessary** — an already-landed
constraint resolves the contradiction:

* `_ship.agent.md` **L832** / **L861–864**: the cascade op must be invoked *"carrying the
  returned `CLASSIFICATION_BINDING` into that call"*.
* **That op accepts no binding parameter on either surface.** CLI
  `backlogit shipment ship --help` ⇒ exactly `--sha`, `--message`, `--author`. MCP
  `.autoharness/backlog-registry.yaml` **L187–194** ⇒ `ship_shipment.params` = `shipment_id`,
  `sha`, `message`, `author`.
* ⇒ Read as a **direct** call, installed Ship's CASCADE branch is **internally unexecutable**.
* The only surface accepting `classification_binding` is `mode: safe-close` (`SKILL.md` **L60**),
  and **L595** routes a bound `CASCADE` verdict into the Cascade Close Sub-Procedure, with
  **L1108/L1111/L711** confirming that Sub-Procedure invokes the cascade op under the
  `returned_ids`, two-set, and `parent_id` gates.

**Conclusion:** `mode: safe-close` with a `CASCADE` binding is the **only executable realization
of installed Ship's own mandate**, not a competing route. §6 condition 3 is satisfiable;
`PA-017-CASCADE` is not invalidated; **no prerequisite Ship shipment is staged.** Revision 3's
error was asserting authority instead of proving executability — **PF-10** now makes the
reconciliation falsifiable at preflight, and voids the route if a binding parameter ever appears.

**P0-2 — P-009 violation. FIXED.** Revision 3 called a single-parent squash/rebase merge
"legitimate". `.github/policies/workflow-policies.md` **P-009 is literally "Merge-Commit-Only
(No Squash or Rebase Merge)"**. S-7 and S-17 now require true merge commits; S-7a requires
**exactly two parents**, parent 1 = mainline, parent 2 = reviewed PR head, any other shape
HALTing as a P-009 violation; **R-4's plain-revert branch is REMOVED** (it would have masked the
violation and destroyed the `-m 1` target). New **AC-14**.

**P1 remediations — all applied:**

| Finding | Disposition |
|---|---|
| Normalize one-way-door semantics | §7.1 rebuilt into **two regimes** keyed on one question (*is `017-S` `active`?*). Regime A ⇒ HALT to Stage. Regime B ⇒ leave `active`, **zero backlog mutation**, `return-blocked`/`move`/`abandoned`/R-1…R-7 all explicitly forbidden, HALT to operator. New **AC-13** |
| S-3b must reference its own disposition | S-3b row now states its quarantine disposition is **its own, NOT §7 R-1…R-7** |
| Probe 26 script path | corrected to `docs/plans/evidence/2026-09-16-017-S-closure/probe26.ps1` (`.backlogit/runtime/` is **gitignored**, so the original path could never resolve on `origin/main`) |
| PF-1 references a nonexistent `docs/reviews/` artifact | requirement **removed**; PF-1 now requires the exact plan on `origin/main` plus attempt-3 persona coverage and literal `dispatch_mode:` / `decision: PASS` markers |
| `019.007-T` PF-5 anchoring | §2.1, PF-5, **AC-9** and H.4 all re-anchored to **raw-byte identity between S-11 and S-15/R-6**; PF-5's earlier value is **informational only** |
| AC corrections | **AC-1** ⇒ PF-0…PF-10; **AC-2** ⇒ read-only re-verification of committed Probe 25/26 **and the Probe 26 script**; **AC-3** ⇒ `active` at S-10 entry, `done` on success |
| `017-S` record's impossible condition | *"both path inventories"* replaced with `pre_cascade_sha` + `pre_paths` before S-14; `post_paths` restated as an S-15 postcheck |
| `{impl_paths}` unconsumed | **Finding closed as inaccurate** — it *is* consumed at S-7a(iii). Strengthened anyway: **S-7 now asserts an exact changed-path allowlist** (PR set **equals** `{impl_paths}`, no extras, no omissions). New **AC-16** |

### Attempt 3 — revision 4

<!-- plan-review-attempt: 3 -->

dispatch_mode: multi-agent
reviewers: pending
decision: NOT YET RUN

All blocking escalation findings are closed on revision 4. Attempt 3 is now justified, but
**has not been run**, and this plan claims no PASS it has not earned. `017-S` remains
**NOT claimable**; ground 3 stays open until attempt 3 returns `decision: PASS` and this plan
is on `origin/main`.


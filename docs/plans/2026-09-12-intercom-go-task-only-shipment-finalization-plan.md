# Plan — Verified Task-Only Shipment Finalization

> **Requires plan hardening**: **yes — applied in rev 2, extended in rev 3, re-hardened in rev 4, rev 4.1, rev 5 and rev 6** (see `## Plan Hardening`).

## ROLE CHANGE (rev 7) — this document is now the UMBRELLA EVIDENCE AND DECISION SOURCE

**This document is NO LONGER AN EXECUTABLE PLAN.** Its rev-6 `MUST_REPLAN` verdict was
**accepted, not overturned**. Acting on it as a plan is a process error.

The work it scoped has been **structurally re-planned** as a dependency-ordered
**three-shipment bootstrap**. See
`docs/decisions/2026-09-12-intercom-go-three-shipment-bootstrap-deliberation.md`.

| Successor | Surface | Plan |
|---|---|---|
| **A** `021-S` / `022-F` / `022.001-T` | `_ship.agent.md` + 1 Go test | `docs/plans/2026-09-12-intercom-go-ship-feature-completion-foundation-plan.md` |
| **B** `022-S` / `023-F` / `023.001-T` | `workflow-policies.md` P-015 + 1 Go test | `docs/plans/2026-09-12-intercom-go-p015-taskonly-authorization-plan.md` |
| **C** `023-S` / `024-F` / `024.001-T` | `shipment-reconcile/SKILL.md` + 1 Go test | `docs/plans/2026-09-12-intercom-go-taskonly-finalize-skill-implementation-plan.md` |

**What this document REMAINS AUTHORITATIVE for**, and what A/B/C cite it as:

- **§3 / §3.0 / §3.1** — the probe evidence base (Probes 14–20), its fidelity table, and
  the withdrawal of rev-4.1 probes 5/5b/10/12/13.
- **§6.1 Inventory Tables §6.1-A/B/C/D** — the **closed 28-site clause inventory**, split
  across A/B/C **exactly** (`C1–C6` → A; `A1–A6` → B; `B1–B10`, `B12–B17` → C) with the
  additive rows `A7`, `A8`, `B11`, `B18` assigned to B, B, C, C respectively.
- **§6.2–§6.5** — topology assertions `T1–T5`, guards `G0–G13`, post-guards `P1–P10`, the
  TOCTOU boundary, the two-path failure split and bounded-recovery semantics. **T1–T5**
  land in **B**; the rest land in **C**.
- **§8.2** — P-007 ordering (`P1–P10` evaluate on the **raw** post-call state **before**
  any archive restoration).
- **§8.3** — the "first real self-close fails after the implementation merged" contingency.
- **§6.3** — the binding engine digest.
- **§11 / §14** — the PR #54 actor boundary and merge gates (**operator** is the sole
  push/merge actor; no agent push).

**What is SUPERSEDED**: §4 (surface), §5 (ordering), §7.1 (assertion set), §9 (acceptance
criteria), §10.1/§10.2 (budget and the no-split finding), and §13 (adversarial scope) —
all now carried per-shipment by the A/B/C plans. §10.2's "splitting is unavailable"
finding was correct **for the task-only close shape** and is dissolved by the bootstrap,
which closes A and B under the **existing** `CASCADE` exception.

- **Date:** 2026-09-12
- **Revision:** 7 — role changed to umbrella evidence/decision source (rev 1 → **FAIL**; rev 2 → **FAIL**; rev 3 → **ADVISORY**; rev 3 → independent adversarial **MUST_REPLAN**; rev 4 → internal review **FAIL**; rev 4.1 → internal re-review **ADVISORY**; rev 4.1 → independent adversarial **MUST_REPLAN**; rev 5 → independent adversarial **MUST_REMEDIATE**; rev 6 → **MUST_REPLAN on budget**, see §10.1; rev 7 → **re-planned as the three-shipment bootstrap**; see §12)
- **Source deliberation:** `docs/decisions/2026-09-12-intercom-go-task-only-shipment-finalization-deliberation.md`
- **Stash origin:** `A10EF3D0` (`kind: deliberation`, `priority: critical`)
- **Branch:** `chore/stage-pipeline-policy-gap` (planning artifacts carried over by cherry-pick; see §14)
- **Engine identity of record (binding):** `backlogit 1.10.1-0.20260823032255-b07729386a31+dirty`, **SHA-256 `1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98`** — the digest, not the version string, is the binding identity (§6.3)
- **Probe evidence (durable, in-repo):** `docs/plans/evidence/2026-09-12-task-only-shipment-finalization/`
- **Harvested backlog:** feature `022-F`, task `022.001-T`, shipment `021-S` — **rev 7: `021-S`'s manifest is re-shaped from task-only `[022.001-T]` to the FULLY-COVERED ROOT `[022-F, 022.001-T]`, and `022-F`/`022.001-T` are re-scoped to bootstrap shipment A.** Dependency chain becomes `017-S -> C(023-S) -> B(022-S) -> A(021-S)`; the direct `017-S -> 021-S` edge is removed only **after** the `017-S -> 023-S` edge exists, so no eligibility window opens.
- **Status:** **SUPERSEDED AS A PLAN (rev 7); RETAINED AS EVIDENCE.** The rev-6 **MUST_REPLAN** verdict on the §10.1 budget stands and was **accepted**: the closed clause inventory (§6.1) enumerates **28 normative edit sites** and does **not** close under the 2-hour rule as one atomic task. Rev 7 re-plans the work as the **three-shipment bootstrap** above, in which each shipment carries exactly one covering feature and one ≤2 h task. Separately still **BLOCKED** — PR #54 has no agent-executable route and requires the **operator** as PR actor (§11); this is **unchanged** by rev 7.

---

## 1. Objective

Make SAFE_CLOSE-classified shipments closable on backlogit 1.10.1 by authorizing
one **new, narrowly classified** close path — `TASK_ONLY_FINALIZE` — admitted
**only** after machine-verifying a precisely-bounded safe topology, and
fail-closed to today's behavior otherwise.

**Not in scope (explicit):** expanding Ship's Role Boundary; a generic
`move shipment → shipped`; bypassing Ship Step 6 pre-mode; any unrestricted
cascade; any weakening of the existing P-015 fully-covered-root exception.

## 2. Root cause

Safe-Close step 8 prescribes `backlogit move <shipment_id> --status shipped`,
which the 1.10.1 engine **refuses** (exit 9, "shipment must be shipped via
ShipShipment, not a direct status update"). Both documented substitutes are
contractually forbidden, and the CASCADE exception is unreachable because it
requires a covering-feature member that Ship Step 6 pre-mode then rejects
(`expected_status: done`, never set on a feature by any Ship step).

## 3. Evidence base — executed probes

**Evidence location (rev 5, corrected).** Rev 4.1 stated that probes ran "in
throwaway backlogit workspaces under `%TEMP%`" and that "every sandbox was
deleted and the repo tree verified clean afterwards". That left **no
reproducible artifact** behind, so none of those runs was independently
checkable. Rev 5 **re-executed** the load-bearing probes inside the workspace
and committed both the command script and the full transcript to
`docs/plans/evidence/2026-09-12-task-only-shipment-finalization/`.

Probes execute in `.backlogit/runtime/stage-probe-scratch/` (gitignored,
workspace-contained) and recovery artifacts in
`.autoharness/backups/stage-recovery/` (gitignored, workspace-contained).
Neither is under any compared queue/archive/log inventory. Scratch fixtures are
removed after transcripts are captured; durable evidence is never deleted; no
raw `.db`/`-wal`/`-shm` file is committed.

| Probe | Topology | Result |
|---|---|---|
| **1** | `move <shipment> --status shipped` | **exit 9**, refused. Blocker confirmed. |
| **2** | manifest `[T]`; parent `F` live, excluded; `T` has no descendants | `archived_ids=[T,S]`, `returned_ids=[]`; `F` **byte-identical**; `S` → `archived_status: shipped` |
| **3** | manifest `[T]`; `T` owns an **omitted `queued` subtask** | `archived_ids=[ST,T,S]` — omitted subtask **silently archived**, force-stamped `done` |
| **4** | manifest `[T1]`; **omitted sibling** `T2` + its subtask under the same parent | `archived_ids=[T1,S]`; `returned_ids=[T2, T2.ST]` — **non-empty** |
| **6** | manifest `[T_live]`; pre-archived task + subtask **outside** the manifest under the same parent | `archived_ids=[T_live,S]`, `returned_ids=[]`; out-of-manifest archived descendants **byte-identical** |
| **7** | create a 2nd shipment containing an already-assigned item | **refused**: `item already assigned to shipment 001-S` |
| **8** | `return_blocked` on the covering feature of a parent-first manifest | manifest reduced to task-only; feature → `blocked`; sibling byte-identical; `move --status queued` restores it |
| **9** | live **out-of-manifest sibling** under the same parent, owned by another shipment | `returned_ids=[sibling]`; the sibling's **`parent_id` was STRIPPED** (orphaned); collateral crossed a shipment boundary |
| **11** | `shipment ship` against a **`queued`** shipment | **refused**, exit 1, `shipment status conflict` — the record must be `active` |
| **14** | **executed, committed.** Two task-only shipments; successor `depends_on` predecessor `--type blocks` (the live `017-S -> 021-S` shape) | Predecessor `queued` → `ready_set=["001-S"]`, successor **suppressed**. Predecessor `active` → `ready_set=[]`, successor **still suppressed**. Predecessor closed → `archived` + `archived_status: shipped` + `commit` retained; after `sync`, `ready_set=["002-S"]` — successor **eligible**. All verdicts `status: ok`, `degraded_reason: null` |
| **15** | **executed, committed. EXACT `017-S` fixture**: 12-member task-only manifest = 1 live `done` task + **3 archived tasks + 8 archived subtasks (3/3/2)**, root covering feature live and excluded | `archived_ids=[T_live,S]`, `returned_ids=[]`; **all 11 pre-archived members and the root parent `byte=SAME`, `path=SAME`, `parent=SAME`; `INVARIANCE_FAILURES=0`** |
| **16** | **SUPERSEDED by Probe 19 — retained for history only.** Injected partial failure: archive destination of the shipment record occupied by a directory, then ship | **exit 1** *after* partial mutation (torn state observed). **Withdrawn as evidence for the §6.5.3 procedure**: it swept whole directories instead of an enumerated set, compared append-only streams by **length only**, never ran the approval-withheld branch, resolved its repo root to `docs\` (so the committed script does not reproduce as written), and ran on stock `init` defaults. See §6.5.3 and the evidence-integrity table below |
| **17** | **executed, committed.** Ship Step 5 `pipeline-topology` gate in a PR-lifecycle-only (unclaimed) session, all four routes | **Every route fails closed** — exit 2 `agent mode requires --shipment`, exit 1 `LIFECYCLE_NO_ACTIVE_SHIPMENT`, exit 2 `--phase lifecycle requires --shipment … in any mode`, exit 2 ambient. See §11 |
| **18** | **executed, committed. LIVE-CONFIG-SEEDED** exact prerequisite topology: root feature outside the manifest, one live `done` task, task-only shipment, **plus a second shipment holding the inbound `depends_on … --type blocks` edge** (the live `017-S -> 021-S` shape) | `DIGEST_GATE=PASS`; seeded from live `config.yaml`/`header-def.yaml`/`hooks.yaml`/`registry.yaml`/`migration.yaml`/`templates` with each seed file's SHA-256 recorded. **`INBOUND_EDGE_BYTE_IDENTICAL=True`**, `INBOUND_RECORD_PATH_CHANGED=False`, `PARENT_FEATURE_BYTE_IDENTICAL=True`, `returned_ids=[]`, archived record retains `archived_status: shipped` + `commit`. Root cause re-confirmed **under live config**: `MOVE_SHIPPED_EXIT=9`. Config delta confirmed material: `SEED_STATUS_ENUM_HAS_SHIPPED=False`, `SEED_HOOKS_VALIDATE_TRANSITION=True`, `SEED_HOOKS_PRE_TASK_GATE=True` |
| **19** | **executed, committed. Bounded recovery, rewritten to §6.5.3 and run on BOTH approval branches**, live-config-seeded, real errored invocation (`SHIP_EXIT=1`, torn state) | **Withheld:** quarantine 0, restore 0, torn state preserved, `TERMINAL_DISPOSITION=HALT`. **Granted:** 1 quarantined (moved, never deleted), 4 restored from the **enumerated** inventory, `MUTABLE_BYTE_EQUIVALENCE_FAILURES=0`, `MUTABLE_PATH_EQUIVALENCE_FAILURES=0`, `DEPENDENCY_EQUIVALENCE_OK=True`, `RESIDUAL_UNEXPECTED_PATHS_IN_BOUNDED_SET=0`. Both: `APPEND_ONLY_PREFIX_VIOLATIONS=0` (**full byte prefix**, not length), `APPEND_ONLY_DELETIONS=0`, out-of-bounds lock file **reported, not moved**, cache rehydrated only via official `sync` |
| **20** | **executed, committed. O-6 post-archive closure route**, live-config- **and** live-`.autoharness`-seeded, driven to the exact post-archive state | **Q1:** all **4** post-archive routes non-zero — `LIFECYCLE_NO_ACTIVE_SHIPMENT` (exit 1) on both shipment-bearing routes, exit 2 on the two invalid ones; `POST_ARCHIVE_ROUTES_NONZERO_EXIT=4`. **Q2:** the installed Ship contract mandates the lifecycle gate at L559/L571/L767 but **not** before the closure PR (L741–744). See §8.1.1 |

**Withdrawn probes.** Rev 4.1's Probes 5, 5b, 10, 12 and 13 are **superseded and
withdrawn as evidence**: they ran under `%TEMP%` and were deleted, so no
transcript exists. Their claims are now carried at full strength by the
re-executed Probes 14, 15 and 16, and by the engine-identity capture in the
evidence README.

**Evidence-degraded probes (rev 5 — honest accounting, resolves a rev-5
internal-review P1).** Probes **1, 2, 3, 4, 6, 7, 8, 9 and 11** are **retained**
but are *also* rev-4.1 `%TEMP%` runs with **no committed script or transcript**.
They are **not** reproducible and **cannot** be re-executed by §6.3's probe-refresh
rule after an engine upgrade, because no script survives. The following live
claims rest on them and are therefore **evidence-degraded**, not
evidence-committed:

| Claim | Rests on | Status |
|---|---|---|
| G11 (shared membership refused at creation) | Probe 7 | degraded |
| G12 (`queued` record refused) | Probe 11 | degraded |
| G13 / P1 (engine leaves dangling edges; return strips `parent_id`) | Probe 9 | degraded |
| P5 (parent preserved) | Probes 2, 6 | degraded — **but** independently re-proven by **Probe 15** (parent `byte=SAME`) |
| T5/G5 rationale (omitted descendants/siblings are damaging) | Probes 3, 4, 9 | degraded |
| §2 root cause (`move --status shipped` → exit 9) | Probe 1 | degraded — **but** independently re-proven **under live config** by **Probe 18** (`MOVE_SHIPPED_EXIT=9`) |

An earlier rev-5 draft asserted "No conclusion in this plan rests on a deleted
sandbox." **That was false and is WITHDRAWN.** The accurate statement is: no
conclusion rests on a *withdrawn* probe, and every conclusion resting on a
*retained-but-uncommitted* probe is tabulated above. Re-executing and committing
probes 2/3/4/6/7/9/11 is carried as a named residual (§10.3).

### 3.0 Evidence-integrity table (rev 6) — probe-vs-contract fidelity

Rev 5 treated "a probe exists and is committed" as sufficient. It is not: a
committed probe can still fail to implement the procedure it is cited for. Rev 6
records fidelity explicitly.

| Probe | Committed? | Live config? | Digest-gated? | Implements the cited contract? |
|---|---|---|---|---|
| 14 | yes | **no** (stock `init`) | no | yes, for dependency lifecycle |
| 15 | yes | **no** (stock `init`) | no | yes, for the exact `017-S` member shape |
| 16 | yes | **no** | no | **NO — superseded by 19** (5 divergences, §6.5.3) |
| 17 | transcript only | live workspace | no | yes, for gate feasibility |
| **18** | **yes** | **YES** | **YES** | **yes** |
| **19** | **yes** | **YES** | **YES** | **yes — both approval branches** |
| **20** | **yes** | **YES** | **YES** | **yes** |

**Digest binding is now enforced in-script, not by convention.** Probes 18, 19
and 20 each resolve the backlogit executable path, compute its SHA-256, compare
it to the recorded identity, print `DIGEST_GATE=PASS`/`FAIL`, record
`--version`, and **`exit 1` before performing any operation** on mismatch. The
transcripts capture all four fields. Probes 14/15/16 predate this rule.

**Committed artifacts for the rev-6 probes** (all under
`docs/plans/evidence/2026-09-12-task-only-shipment-finalization/`):

| Probe | Script | Transcript |
|---|---|---|
| 18 | `probe18-liveconfig-inbound-edge.ps1` | `probe18-liveconfig-inbound-edge.txt` |
| 19 | `probe19-bounded-recovery-v2.ps1` | `probe19-bounded-recovery-v2.txt` |
| 20 | `probe20-post-archive-closure-route.ps1` | `probe20-post-archive-closure-route.txt` |

**Authorization surface — CLI only.** All runtime authorization in this plan is
established **through the `backlogit` CLI**. The **MCP surface is neither probed
nor authorized**: no probe exercises `backlogit mcp`, no transcript records MCP
behavior, and **no MCP/CLI equivalence is claimed**. If the implementation or
Ship intends to close through MCP, that surface requires its own digest binding
and its own probes before any authorization transfers to it.

**Residual config delta.** Probes 14, 15 and 16 remain stock-`init`-based.
Probe 18 shows the delta is material — the live `config.yaml` status enum has
**no `shipped` value** and the live `hooks.yaml` enables `validate_transition`
and a `pre_task_completion_gate` — so their conclusions are re-derived under
live config where load-bearing (18 re-proves the parent-preservation, merge-SHA
retention and root-cause claims) and otherwise carried as a §10.3 residual.

### 3.1 What the probes establish

- **Descendants of a manifest item are ARCHIVED** (Probe 3) — a true correctness
  hazard: `releaseScopeItemIDs` recursively adds every descendant of each
  manifest item before archive-candidate collection.
- **Live descendants of the parent that are outside the manifest are RETURNED,
  and returning STRIPS `parent_id`** (Probes 4 + 9). Probe 9 read the returned
  record and found the `parent_id` field **absent** where it had been `001-F` —
  collateral **orphaning**, not merely a status touch, and the damage **crosses
  shipment boundaries**. This topology is **excluded by G5** and **detected by
  P1**; it is never authorized.
- **`017-S`'s exact pre-archived member set is inert** (Probe 15). Rev 4.1
  granted this support on a **three-member representative** fixture. Rev 5
  replaces that with the **exact** live shape — 3 archived tasks and 8 archived
  subtasks under one root feature, 12-member manifest — and measures every
  member: byte, path and `parent_id` all unchanged, `returned_ids=[]`. This
  **discharges** the adversarial precondition that pre-archived descendants may
  be supported only on an exact executed fixture. Support is **granted**; the
  contingent pre-`017-S`-claim manifest normalization is **not** required.
- **A ship-archived live task RETAINS its `parent_id`** (Probe 15) — the archived
  live task carries `status: archived`, `archived_status: done` and the merge
  `commit`. **P9** is therefore a satisfiable guard, and the *archive* path
  (lineage preserved) is distinguished from the *return* path (lineage stripped).
- **A partial failure leaves a TORN, self-inconsistent state, and bounded
  recovery restores it** (Probe 19, both approval branches). The engine applied
  mutations and *then* failed, leaving the shipment record in the queue declaring
  `status: shipped` — a state the engine itself refuses to create via `move`
  (Probe 18, exit 9, re-confirmed under live config).
  Recovery restored byte/path equivalence and the engine's view, **without**
  rewinding any append-only stream and **without** restoring any DB byte.
- **A `blocks` edge is satisfied by predecessor STATUS, not by edge removal**
  (Probe 14). After the predecessor archives, the successor **still carries the
  edge**, yet becomes eligible. The edge persists as durable lineage. This is the
  same rule the Orchestrator consumes via `autoharness gate dag-readiness`.
- **Shared shipment membership is structurally refused at creation** (Probe 7),
  but the manifest is a plain YAML list that hand-editing or a future API could
  still violate, so G11 verifies it independently rather than trusting the engine.
- **The shipment record must be `active`** (Probe 11). A `queued` record fails
  closed with `shipment status conflict`. Now **G12**.

## 4. Surface

| # | File | Kind | Change |
|---|---|---|---|
| 1 | `.github/policies/workflow-policies.md` | governing policy | Add a **second, narrower** named exception to P-015 authorizing `TASK_ONLY_FINALIZE`. The existing exception's **preserved region** (§4.1) is unchanged byte-for-byte. Amendment Log entry appended. |
| 2 | `.github/skills/shipment-reconcile/SKILL.md` | procedure contract | Add the `TASK_ONLY_FINALIZE` classification, its guards, and a step-8 fail-closed cross-reference. |
| 3 | `.github/agents/_ship.agent.md` | **executing-agent contract** | Reconcile the Step 6 closure clauses so a freshly-reloaded Ship **recognizes** `TASK_ONLY_FINALIZE`. Pointer-level only; the skill stays authoritative. |
| 4 | `tests/integration/taskonly_finalize_contract_test.go` | generated harness | One table-driven characterization test asserting all three contract files carry the required properties. |

**File count: 4.** **New production code files: 0.**

**Why surface 3 is mandatory, not optional (rev 5).** `_ship.agent.md` line 794
reads: *"**Do NOT call `backlogit shipment ship` / `backlogit_ship_shipment`**
unless the P-015 **VERIFIED FULLY-COVERED-ROOT EXCEPTION** below applies."* That
clause names **one** exception and is the instruction the executing agent
actually reloads. §8 step 5 requires Ship to check out a fresh `main`, reload the
merged contract, and **positively verify** the new tokens. If `_ship.agent.md` is
not amended, that reload returns an agent file that still authorizes exactly one
exception, and a compliant Ship would **refuse** the `TASK_ONLY_FINALIZE` call it
just merged — the self-hosting sequence in §8 would deadlock at step 7. Lines 803
(`safe-close remains the default`) and 822 (`in place of the safe-close sequence`)
are the two other clauses in the same file that assume a single exception.

P-015 states the prohibition *and* names its exceptions, so a skill-only edit
would leave a compliant Ship still obligated to refuse. All three instruction
surfaces are *directly coupled* and must become mutually consistent **in one
atomic change** — which is exactly why this remains **one task** (§5, AC-17).

### 4.1 Preserved region of the existing P-015 exception (rev 5 — resolves a rev-5 internal-review P0)

Rev 5's first draft said the fully-covered-root exception is "preserved
byte-for-byte" **while also** mandating edits to `workflow-policies.md` lines 436
and 443 — which are the exception's own header sentence and its item 6. Both
cannot hold. Internal review caught this as a **P0**; it is resolved by defining
the preserved region **precisely** instead of loosely:

| Region | Disposition |
|---|---|
| Preconditions **1–5**, **item 7**, and the **SUPERSESSION NOTE** (155-S, 2026-09-03) | **PRESERVED byte-for-byte.** Verified by diff. |
| Header sentence at line **436** ("…remains the DEFAULT for every shipment closure") | **AMENDED** — to acknowledge a second, narrower exception |
| Item **6** at line 443 ("…IS permitted … in place of the single-artifact safe-close") | **AMENDED** — to scope the permission to the **CASCADE** branch only |

The **safety content** of the existing exception — every precondition that
decides whether a cascade is allowed — is therefore untouched. Only the two
sentences that assert *exclusivity* are amended, which is precisely the
reconciliation §6.1 requires. AC-1 and §7.1 row 2 are stated against this
narrowed region, not against the whole block.

## 5. Ordering (test-first; `main` stays green)

1. **H0 — red.** Add the harness only.
   - `go vet ./...` exits **0**;
   - `go test ./...` exits **non-zero** (P-004's literal requirement) **and** the
     targeted `go test ./tests/integration/ -run TestTaskOnlyFinalize_ContractAuthorized`
     exits non-zero;
   - output carries the literal marker `not implemented: task-only finalize contract`;
   - red count over this unit's generated functions is literally **1 of 1**.
2. **H1 — green.** Apply all three contract edits. Marker gone; test passes;
   `go vet ./...` exits 0; full `go test ./...` exits **0**.

**This release unit contains exactly one task and exactly one generated
function.** A sequential split is **not** permitted under any circumstance: a
second task under this feature re-creates the Ship Step 4.3 full-suite-green
deadlock *and* breaks the §8 self-hosting topology, whose G5 proof depends on the
covering feature having exactly one live descendant. It is also **semantically**
indivisible — the three instruction files must be mutually consistent before the
unit can self-close, so a partial landing is an incoherent contract. If the work
is judged to exceed the 2-hour budget, the correct response is **HALT and return
to Stage** (§10), not to split the release unit.

## 6. The normative safety contract

### 6.1 Classification `TASK_ONLY_FINALIZE` and classifier order

The classifier evaluates, in this exact order:

1. the **existing** P-015 full-root `CASCADE` check (unchanged);
2. the **new** `TASK_ONLY_FINALIZE` check (all of G0–G13 below, machine-verified);
3. the **existing** `SAFE_CLOSE` / `HALT` fallback (unchanged).

`CASCADE` and `TASK_ONLY_FINALIZE` are mutually exclusive **by construction**:
`CASCADE` requires at least one qualifying **feature** member; `TASK_ONLY_FINALIZE`
requires **zero** feature members (G1). No manifest satisfies both, so the
ordering resolves no ambiguity and creates none.

Authorization derives from the **P-015 amendment** (surface 1), not from ordering.
Any guard failure, ambiguity, or query error → **not** classified → fall through
to `SAFE_CLOSE`/HALT exactly as today.

**Normative-clause reconciliation (mandatory, AC-13) — CLOSED inventory (rev 6).**
Rev 4.1 told the implementer to "locate and rewrite" exclusivity clauses. Rev 5
enumerated 11 rows / 14 sites and then recorded three *known* further sites as
**open items** (O-1, O-2, O-3) — which is the same open-endedness in a different
costume: a site you have already identified is not an open question, it is an
unbudgeted edit. Rev 6 **closes** the inventory. The scan below was re-run against
the installed files in this workspace at `1f55e6a`; it is **Inventory Table
§6.1-A**, it is referenced by name from **AC-13** and **AC-7**, and it is the
budget basis in §10.1.

Each site is classified:

- **MUST** — the clause is *incorrect or blocking* if left unedited (it either
  forbids the merged path, or asserts an exclusivity that the amendment falsifies).
- **CONSISTENCY** — the clause remains *literally true* after the amendment because
  it is already scoped to a branch that does not change, but it reads as
  single-exception framing and is reconciled for coherence.

Every entry must be reconciled to name both exceptions in the classifier order
CASCADE → `TASK_ONLY_FINALIZE` → `SAFE_CLOSE`/HALT.

**Inventory Table §6.1-A — `.github/policies/workflow-policies.md` (P-015)**

| # | Line | Clause (abbrev.) | Class | Why |
|---|---|---|---|---|
| A1 | 410 | heading "Single-Artifact Shipment Closure (**No Cascade Ship**)" | CONSISTENCY | parenthetical now describes neither exception |
| A2 | 420 | "It MUST NOT call the cascade `backlogit_ship_shipment` for closure" | **MUST** | the blanket prohibition *is* the authorization gate |
| A3 | 426 | Postcondition: "the cascade … was never called" | **MUST** | unreachable postcondition once a 2nd exception exists |
| A4 | 436 | "**VERIFIED FULLY-COVERED-ROOT EXCEPTION** … remains the DEFAULT" | **MUST** | asserts exclusivity; amended per §4.1 |
| A5 | 443 | item 6 "IS permitted … in place of the single-artifact safe-close" | **MUST** | must scope the permission to the CASCADE branch |
| A6 | 448 | "**This exception** is defined entirely in terms of the general shape above … a pure classification function … that Ship consults to **select the close path**" | **MUST** | the selector sentence; binary → ternary |
| A7 | new block after 448 | **NEW** second exception block authorizing `TASK_ONLY_FINALIZE` (T1–T5, G0–G13, P1–P10 pointers) | **MUST** | the authorization itself |
| A8 | 746 / after 774 | Amendment Log row naming `TASK_ONLY_FINALIZE` | **MUST** | required by AC-1 |

**Inventory Table §6.1-B — `.github/skills/shipment-reconcile/SKILL.md`**

| # | Line | Clause (abbrev.) | Class | Why |
|---|---|---|---|---|
| B1 | 3 | frontmatter: "**except for** the narrow, machine-verified P-015 fully-covered-root case" | **MUST** | literal sole-exception phrasing |
| B2 | 10–11 | "run `mode: safe-close` **in place of** the destructive cascade" | CONSISTENCY | true of safe-close; reads binary |
| B3 | 14–15 | "runs the P-015 verified fully-covered-root classification and, **only when** every precondition holds, delegates to the Cascade Close Sub-Procedure instead" | **MUST** | states the classifier as binary |
| B4 | 28 | "**instead of** the cascade … call, then `mode: post`" | CONSISTENCY | binary framing |
| B5 | 30–32 | "never calls the cascade op directly **unless** that classification confirms the narrow verified fully-covered-root exception" | **MUST** | sole-exception carve-out |
| B6 | 232–233 | "It **never calls** the cascade `backlogit_ship_shipment`" | CONSISTENCY | scoped to safe-close mode; stays true |
| B7 | 324 | pre-mode `RECONCILE_FAIL`: "Do NOT call `backlogit_ship_shipment`" | CONSISTENCY | pre-mode never closes on *any* path; already correct |
| B8 | 356–358 | Safe-Close preamble: "Runs **in place of** the destructive cascade … Step 0 below, where the cascade op is **the** *permitted* close path" | **MUST** | names one permitted alternative |
| B9 | 408–415 | **Step 0(c) "Classify the close path"** — classifier body | **MUST** | the classification function itself |
| B10 | 505–512 | **Step 0 close-path SELECTOR** — `CASCADE selected` / `SAFE_CLOSE selected (default, including any classifier error…)` | **MUST** | **the binary selector**; must become ternary with `TASK_ONLY_FINALIZE` evaluated between them |
| B11 | 591–605 | **Safe-Close step 8** ("Close the shipment record itself") | **MUST** | **AC-7 site.** Needs the fail-closed cross-reference; step 8's own text otherwise unchanged |
| B12 | 618–620 | Cascade Close Sub-Procedure heading "(P-015 verified fully-covered-root exception **ONLY**)" + "Runs **only** when Step 0 … selects `CASCADE`" | CONSISTENCY | still true; scoped to the cascade branch |
| B13 | 698–709 | No-substitution rule: "…grants no license to invoke cascade when Step 0 selects `SAFE_CLOSE`, which remains governed by the P-015 **default prohibition**" | **MUST** | the rule must also bind the third verdict |
| B14 | 998 | lock-failure: "Do NOT call `backlogit_ship_shipment`" | CONSISTENCY | no close path is authorized without the lock; already correct |
| B15 | 1015–1016 | Safety invariants: "runs **in place of** the cascade" + "Close-path selection is made **only** from the machine-checkable classification result (Step 0)" | **MUST** | invariant list must enumerate the new path |
| B16 | 1022 | "Safe-close archives the shipment record as its own single artifact and **never via the cascade op**" | CONSISTENCY | scoped to safe-close; stays true |
| B17 | 1074 | references line naming P-015 as "cascade prohibition" | CONSISTENCY | pointer text |
| B18 | new section | **NEW** `TASK_ONLY_FINALIZE` classification section: T1–T5, G0–G13, P1–P10, §6.5 failure split, §6.5.3 recovery | **MUST** | the normative contract body |

**Inventory Table §6.1-C — `.github/agents/_ship.agent.md`**

| # | Line | Clause (abbrev.) | Class | Why |
|---|---|---|---|---|
| C1 | 756–764 | **Mandatory pre-self-close context reload** — names only Ship instructions + `shipment-reconcile` | **MUST** | §8 step 6 must also reload **P-015** and positively verify the merged tokens **and the merge commit** |
| C2 | 789–791 | safe-close **summary** prescribing `backlogit move <shipment_id> --status shipped` | **MUST** | **O-1 site.** §2's root cause: the live engine **refuses** this (exit 9, re-confirmed under live config by **Probe 18**) |
| C3 | 794–795 | "**Do NOT call** `backlogit shipment ship` … **unless** the P-015 **VERIFIED FULLY-COVERED-ROOT EXCEPTION** applies" | **MUST** | **load-bearing** — the clause a reloaded Ship actually reads (§4) |
| C4 | 802–803 | "P-015 verified fully-covered-root exception (**select the close path** …): safe-close remains **the default**" | **MUST** | the agent-side selector; binary |
| C5 | 821–823 | "invoke the cascade … **in place of** the safe-close sequence above" | **MUST** | single-exception framing at the invocation site |
| C6 | new bullet | **NEW** `TASK_ONLY_FINALIZE` selector bullet naming the classification and its guards | **MUST** | so a fresh reload recognizes the path |

**Inventory Table §6.1-D — `tests/integration/taskonly_finalize_contract_test.go`**

| # | Site | Class | Why |
|---|---|---|---|
| D1 | one table-driven test function over the §7.1 table | **MUST** | AC-18 |
| D2 | ≤4 read/assert helpers | **MUST** | AC-18; counted separately from the "1 of 1" red count (see §7.1) |

**Closed totals: 28 normative edit sites across 4 files — 20 MUST + 8
CONSISTENCY — comprising 26 clause/site edits, 2 new normative blocks, 1 new
selector bullet, 1 Amendment Log row, 1 table-driven test function and ≤4
helpers.** The three rev-5 open items are **closed into this table**: O-1 →
**C2** and **B10**; O-2 → **B11**; O-3 → the inventory now has a name
(**Inventory Table §6.1-A/B/C/D**) and a committed home (§6.1-E below).

**§6.1-E — home of the committed inventory (closes O-3).** The implementing task
re-runs this scan and commits the regenerated inventory to
`docs/plans/evidence/2026-09-12-task-only-shipment-finalization/clause-inventory.md`.
That is an **evidence** artifact, not a fifth implementation surface, so the
"File count: 4" claim in §4 is unaffected. A clause found by the re-scan and
absent from §6.1-A/B/C/D is **added, never skipped**; if the re-scan finds a
**MUST** site not listed here, the budget in §10.1 is invalidated and the unit
returns to Stage.

The mandatory **pre-mode and post-mode reconciliation steps** of
`shipment-reconcile` are **retained unchanged** — `TASK_ONLY_FINALIZE` adds a
close path, never a reconciliation bypass.

### 6.2 Pre-call guards (ALL required, conjunctive, fail-closed)

#### 6.2.0 Authorized topology — the five explicit first-release assertions

Rev 4.1 described the authorized shape in a prose paragraph and then spread its
verification across **29 "precision points"** in §7.1, most of which asserted
that a sentence *mentions* a concept rather than that a topology *holds*. Rev 5
replaces that with **five explicit, machine-checkable topology assertions**. A
shipment classifies `TASK_ONLY_FINALIZE` **only if all five hold**; any failure,
ambiguity, or query error falls through to `SAFE_CLOSE`/HALT.

| # | Assertion | Rejects |
|---|---|---|
| **T1** | The shipment is **task-only**: every **live** manifest member has `artifact_type: task`. | any live `feature`, `subtask`, `epic` or `shipment` member |
| **T2** | **Exactly one** manifest member is live, and it declares `status: done`. | zero live members; two or more live members; a live member not `done` |
| **T3** | **Zero** `feature` members and **zero** `subtask` members are live in the manifest. Archived members MAY be `task` or `subtask` (Probe 15). | a live covering feature in the manifest (the Probe-3/9 cascade vector) |
| **T4** | **Exactly one** parent feature, **root** (`parent_id` absent), **live** by its frontmatter `status`, and **outside** the manifest. | non-root parent; parent inside the manifest; multiple parents; archived/missing parent |
| **T5** | **No omitted live descendants**: every non-archived descendant of that parent, at every depth, is a manifest member. | Probe 3 (omitted subtask) and Probes 4/9 (omitted live sibling) |

**T3 is scoped to LIVE members deliberately, and Probe 15 is why.** The
motivating case `017-S` has a manifest of one live task plus **eleven archived
members, eight of them `subtask`**. A blanket "any subtask member disqualifies"
rule would reject `017-S` outright, so `TASK_ONLY_FINALIZE` could never classify
the very shipment it exists to close. Probe 15 measured the **exact** `017-S`
member set through a real `ShipShipment` and found every archived member byte,
path and `parent_id` unchanged with `returned_ids=[]`. Archived members are
inert; **live** non-task members remain disqualifying.

Type filtering is by the frontmatter `artifact_type` field — **never** by ID
suffix (`-ST` ends in `T`). **Liveness is by the frontmatter `status` field —
never by queue-vs-archive location**, and T4 is stated that way in rev 6 for a
reason that is now *measured*, not merely asserted: this workspace's live
`.backlogit/registry.yaml` routes `done` artifacts to `archive/`, so **Probe 18**
observed the sole `done` task residing in `archive/` **while still being a live,
non-archived member**. A location-derived liveness test would have misclassified
it. Directory location is used **only** as a location/integrity cross-check —
that the record exists exactly once, at exactly one path, and that the path is
consistent with its declared status — and never as the source of the liveness
decision itself.

#### 6.2.1 Guards

- **G0 — bounded baseline.** Record the pre-call state of the **baseline scope**
  defined in §6.2.2. Every path that changes later must be attributable to this
  call. Pre-existing unrelated modifications inside the baseline scope → **HALT**;
  recovery cannot be bounded.
- **G1–G4 — topology.** Machine-verify **T1–T4** above from frontmatter.
- **G5 — complete parent-rooted live coverage (T5).** Enumerate the parent
  feature's **complete descendant set at every depth** by walking `parent_id`
  across **both** the queue and archive directories. Every descendant that is not
  truly `archived` MUST be a manifest member:
  `live_descendants(parent) == { m ∈ manifest : m not truly archived }`.
  This single guard subsumes the omitted-descendant hazard (Probe 3) and the
  omitted-live-sibling hazard (Probes 4, 9), and it is what makes post-guard
  **P1** (`returned_ids == []`) sound.
- **G6 — enumeration positively verified.** A zero/complete result must be
  positively confirmed against the live workspace, never inferred from a failed
  or partial scan. Any query error, unindexed directory or malformed record →
  not verified → disqualified.
- **G7 — pre-close snapshot.** Capture, from frontmatter and **before** any
  mutating call: each manifest task's `parent_id` + declared `status`; the parent
  feature's `parent_id`, declared `status`, and **full frontmatter bytes**; and
  the shipment record's declared `status`. Declared status is read from the
  `status` field only — **never** inferred from queue-vs-archive location.
- **G8 — orphan scan.** No record outside the manifest may declare
  `shipment_id: <this shipment>`. Any orphan disqualifies.
- **G9 — exactly-one-location precheck, AND destination-path availability.** Two
  conjuncts:
  1. Every manifest task, the shipment record, **and** the parent feature must
     resolve to **exactly one** of the queue or archive directories. Present in
     both (torn) or neither (missing) → **HALT** before any mutation.
  2. **For every record that this call will archive, its computed destination
     path under `archive/` MUST be absent, and MUST NOT be occupied by a
     directory or an unparseable file.** Probe 19's injected failure was exactly
     this: a **directory** named `001-S.md` at the shipment record's archive
     destination. That is *not* a record, so conjunct 1 alone would never see it.
     Without conjunct 2 the plan could not honestly claim §7.3's
     "destination conflict → excluded by G9".
- **G10 — WITHDRAWN in rev 5.** Superseded by **T2** (§6.2.0), which states the
  single-live-member rule as a topology assertion rather than a guard. The
  identifier is retired, **not reused**, so "G0–G13" denotes
  G0–G9 + G11–G13 (13 guards) plus T1–T5. Recorded explicitly because an
  implementer reconciling the skill text must be able to tell a deliberate
  withdrawal from an editing loss.
- **G11 — no shared shipment membership.** No manifest member is listed in any
  other shipment manifest, and no other shipment manifest lists the parent
  feature. Probe 7 shows the engine refuses this at creation, but the manifest is
  hand-editable, so this is verified independently.
- **G12 — shipment record is `active`.** Probe 11: a `queued` record fails closed
  with `shipment status conflict`. The record reaching `active` is the result of
  Ship's ordinary claim, **not** a new privilege granted by this path.
- **G13 — no unresolved blocking dependencies or reference-derived expansion.**
  No manifest member has an unresolved incoming **or outgoing** `blocks`
  dependency on a non-archived record outside the manifest, and no semantic link
  or reference causes the engine's release scope to expand beyond the manifest.
  **Inbound edges are explicitly in scope** — Probe 14 shows the engine does not
  enforce dependency state itself, and Probe 9 shows it leaves dangling edges
  behind, so this guard is the only protection. An edge whose counterpart is
  truly `archived` is **satisfied**, not unresolved (Probe 14).

#### 6.2.2 Baseline scope (normative)

The **baseline scope** for G0 capture, P4 comparison and §6.5 recovery is exactly:

- the shipment record's source-of-truth file;
- each manifest member's source-of-truth file;
- the parent feature's source-of-truth file;
- the source-of-truth file of **every** enumerated descendant of the parent at
  every depth (live and archived), per G5;
- the dependency, link and log records that reference any of the above —
  including records **outside** the parent subtree that hold an **inbound** edge,
  link, or shipment-membership reference to anything in it;
- the queue and archive directories themselves (path inventory, to detect
  additions and deletions, not only content edits).

**Explicitly excluded from mutation comparison (rev 5 — resolves a rev-5
internal-review P1):**

1. the expected report and lock outputs that pre-mode itself produces, **and all
   engine lock files** (`.locks/`, `*.lock`);
2. the engine's index/cache database (gitignored, disposable);
3. **the entire append-only class** — `logs/*.jsonl`, `hooks_queue.jsonl`,
   `stash.jsonl`, and telemetry streams.

Exclusion 3 is **load-bearing, not cosmetic.** Probe 19's transcript shows
per-artifact append-only logs grow on *every* operation. If the append-only class
were inside the mutation comparison, **P4 would fail on every successful close** —
the happy path would route straight to §6.5(B) bounded recovery and HALT, and
§6.5.2 forbids ever restoring those files, so recovery could never re-establish
equivalence either. The path would be permanently unusable.

These are **not** compared for mutation because they are *expected* to change;
they are **validated separately** — the report for well-formedness and expected
content, the lock for correct acquisition and release, the index by successful
rehydration (§6.5.3 step 5), and the append-only class by the
**monotonic-growth rule** of §6.5.2 (length must never decrease).

#### 6.2.3 TOCTOU isolation (rev 6, normative)

**Execution-mode precondition (relied upon, not assumed).** This procedure runs
only under the workspace's enforced **single-agent / single-worktree / P-001**
execution mode. The preconditions are, explicitly and conjunctively:

1. **exactly one agent** active in the workspace;
2. **exactly one worktree** (verified by `git worktree list --porcelain` at Ship
   Step 0.5 item 3a, which halts with `WORKTREE_TOPOLOGY_BLOCKED` on any
   prohibited or ambiguous worktree);
3. **no other active checkpoint or session** — verified by enumerating
   checkpoints and confirming no second `active` record for either agent; and
4. **no concurrent backlog mutation** in flight.

If **any** of the four is not verifiable, the procedure **HALTs** before the
call.

**What is claimed: DETECTION, not prevention.** Rev 6 states this without
hedging. This procedure does **not** and **cannot** prevent an external actor —
a human editor, a foreign process, a second agent started out of band — from
mutating a record. What the enforced execution mode plus the locking and
re-hashing below give is **reliable detection**: any in-scope divergence between
the G0 baseline and the immediately-pre-call re-scan is caught and fails closed.
Any claim of prevention against external actors is **withdrawn**.

1. **Bootstrap enumeration (unlocked).** Perform a first, unlocked G5 walk plus
   an inbound-reference scan to compute the **relation closure**: the shipment
   record, every manifest member, the parent feature, every descendant at every
   depth, **and every existing record holding an inbound or outbound dependency,
   link, or shipment-membership reference to any of those**. This is an
   *enumeration only*; no classification decision is made from it.
2. **Lock the entire relation closure** — via the repository's `file-lock` skill
   or the repository-approved equivalent. Existing outside relationship holders
   are **included**, not merely the parent subtree.
3. **Re-run G5 and the inbound scan under lock.** If the locked enumeration
   differs from step 1's result, release and restart **once**; a second mismatch
   → **fail closed**. This resolves the bootstrap ordering: the lock set cannot
   be computed by a guard that itself requires the locks.
4. Capture the G0 baseline and evaluate every guard entirely under lock.
5. **Immediately before** invoking `ShipShipment`, **re-scan and re-hash** the
   **complete relation closure** — every record enumerated in step 1, in both
   directions (inbound **and** outbound dependency edges, links, and shipment
   membership) — and recompute the classification.
6. If **any** hash, path, declared status, dependency edge, link, or membership
   in that closure differs from the G0 capture, **or any relevant new or changed
   record has appeared**, **fail closed**: release locks, make no destructive
   call, fall through to `SAFE_CLOSE`/HALT.
7. Release locks in a guaranteed-cleanup path on every exit, including failure.

**The `017-S -> 021-S` inbound edge is explicitly in scope.** `017-S` holds an
inbound `depends_on` edge naming this shipment. It is therefore a **relation
closure member**, is captured in the G0 baseline (§6.2.2), is **locked** at step
2, and is re-hashed at step 5 — even though it is not a manifest member, not a
descendant of the parent feature, and never mutated by the call.

**Probe 18 measures this end to end.** It builds the exact prerequisite shape
(root feature outside the manifest; one live `done` task; task-only shipment;
plus a second shipment holding `depends_on <prerequisite> --type blocks`),
baselines the inbound-edge holder, runs a real `ShipShipment`, and re-hashes:

- `INBOUND_EDGE_BYTE_IDENTICAL=True`
- `INBOUND_RECORD_PATH_CHANGED=False`
- `PARENT_FEATURE_BYTE_IDENTICAL=True`
- the edge itself persists after archival (`002-S → 001-S (blocks)`)

So the inbound-edge record is proven **byte-identical through the prerequisite
close**, which is exactly the invariant §6.2.2 places it in the baseline to
protect.

**Honest bound (no overclaim).** This does **not** claim absolute prevention. It
claims **detection**: every *existing* source-of-truth record in the relation
closure is locked and re-verified immediately before the call, and *new* record
creation is excluded by the execution-mode precondition rather than by a lock.
**If that precondition is violated or a new in-scope record is observed anyway,
the procedure HALTS** — and if it were somehow missed pre-call, it is caught
post-call by P1/P10 and routed to §6.5(B).

### 6.3 Authorized call

Only inside the `TASK_ONLY_FINALIZE` classification, and only after §6.2.3 step 5
re-verification passes, invoke the supported
`backlogit shipment ship <shipment_id>` / `ShipShipment` path, passing the same
merge-commit metadata the existing Cascade Close Sub-Procedure passes —
`--sha <merge_sha>` plus `--message` / `--author` where the registry's
`ship_shipment` operation declares them. Probes 14 and 15 confirm the resulting
archived shipment record retains `commit: <merge_sha>` alongside
`archived_status: shipped`, preserving **P-007** commit traceability. No other
classification may reach this call.

**Engine-identity binding (rev 5 — binding, not advisory).** Before the call,
Ship MUST assert **both**:

1. `backlogit version` matches the version string of record; **and**
2. the **SHA-256 of the executable actually being invoked** equals
   `1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98`.

The version string alone is insufficient: the `+dirty` suffix is **not** a unique
build identity — any build from an uncommitted tree at commit `b07729386a31`
reports the identical string. **The digest is therefore the binding identity.**

**A mismatch on either check is a HALT, and requires a probe refresh** — every
probe in `docs/plans/evidence/2026-09-12-task-only-shipment-finalization/` must be
re-executed against the new binary and the digest re-recorded before this path may
be used again. Rev 4.1's escape hatch — "if no digest is available the binding
must be documented as advisory rather than as a guarantee" — is **WITHDRAWN**.
There is no advisory fallback and no unversioned execution.

### 6.4 Post-call guards (ALL required)

Define, once and normatively:

- `allowed_ids` = (manifest tasks) ∪ {shipment record}
- `required_ids` = {shipment record} ∪ {manifest tasks **not** truly `archived`
  in the G7 snapshot}. **The shipment record is a `required_ids` member
  unconditionally, regardless of its own pre-close declared status.**

Guards:

- **P1** — `returned_ids == []`. Sound **because** G5 guarantees no live
  out-of-manifest descendant exists. Probe 9 shows a non-empty result means a
  record was orphaned — `parent_id` stripped — so this is a hard HALT.
- **P2** — `archived_ids ⊆ allowed_ids`. Any ID outside is a cascade.
- **P3** — `required_ids ⊆ archived_ids` (completeness).
- **P4 — independent filesystem verification, not engine self-report.** Diff the
  **mutable current-state class only** (§6.5.2) of the §6.2.2 baseline scope
  against the G0 capture; the excluded classes of §6.2.2 are validated separately
  and are **never** part of this comparison. The set of mutable paths that
  actually changed MUST equal exactly the transitions implied by
  **`required_ids`** (a pre-archived member implies no transition).
  `archived_ids` is never trusted as the sole detector.
- **P5 — parent preserved.** The parent feature is still in the queue directory
  and its **full frontmatter is byte-identical** to the G7 snapshot. Probes 2, 15
  and 6 establish this is achievable.
- **P6 — no torn records.** No manifest task, **nor the parent feature**, nor the
  shipment record exists in both directories; none is missing from both.
  Probe 19 shows a torn record is a real reachable state, not a theoretical one.
- **P7 — provenance.** The archived shipment record reports
  `archived_status: shipped` and retains the merge `commit`. Verified in Probes
  14 and 15; re-asserted after any P-007 restoration per §8.2 step 4.
- **P8 — pre-archived handling.** A manifest task truly `archived` pre-close is
  correctly **absent** from `archived_ids`; this is **not** an anomaly. Probe 15
  measured exactly this: an 12-member manifest yielded
  `archived_ids = [<live task>, <shipment>]` only.
- **P9 — per-task post-verification.** For **each** manifest task in
  `required_ids`, re-read its record and assert: declared `status: archived`;
  resolves to **exactly one** location, the archive directory; and its
  `parent_id` is **unchanged** from the G7 snapshot.
- **P10 — descendant integrity.** Every enumerated descendant outside
  `required_ids` — notably every pre-archived descendant — is **byte-identical**
  to its G0 capture, with `parent_id` intact. This is the direct post-check for
  the Probe 9 orphaning hazard, and Probe 15 confirms it holds across the exact
  `017-S` member set (`INVARIANCE_FAILURES=0`).

**Ordering note.** P1–P10 are evaluated on the **raw post-call state**, before
any P-007 archive restoration or other repair. See §8.2.

### 6.5 Failure semantics — two separate paths

They are distinct and must never be conflated.

**(A) Pre-call guard failure — no destructive call was made.**
Any topology (T1–T5) or guard (G0–G13) failure, any §6.2.3 re-verification
mismatch, any engine-identity mismatch: release locks and fall through to the
**existing** `SAFE_CLOSE` or `HALT` behavior exactly as today. **No recovery is
performed, because nothing was mutated.** This path is non-destructive by
construction, and falling through is safe **precisely because** it is
pre-mutation.

**(B) Post-call failure OR any errored/indeterminate invocation — mutation may
have occurred.**
Triggered by any post-guard (P1–P10) failure **or** a non-zero, timed-out,
partially-written, or unparseable `ShipShipment` result. Perform **bounded
recovery**, then **HALT**. **Never fall back to another close path, never re-run
the call, never continue the session.** The fall-through permitted in (A) is
**forbidden** here: once the engine may have mutated state, "try the other path"
compounds damage rather than avoiding it.

**Recovery is BOUNDED RECOVERY, not atomic rollback.** The engine offers no
transaction. What is specified is an evidence-based restoration of an enumerated
file set, executed end-to-end by **Probe 19** on a real errored/partial
invocation, on **both** approval branches. No claim of
atomicity is made anywhere in this plan.

#### 6.5.1 Recovery path (named, gitignored, workspace-contained)

The recovery root is **`.autoharness/backups/stage-recovery/`**, with per-run
subdirectories `snapshot/`, `quarantine/` and `inventory.json`.

It satisfies all four required properties, verified in Probe 19
(`RECOVERY_PATH_GITIGNORED=True`,
`RECOVERY_PATH_OUTSIDE_COMPARED_INVENTORY=True`):

- **workspace-contained** — never `%TEMP%`, never outside the repo root;
- **gitignored** — via `.autoharness/.gitignore` (`backups/`), so recovery
  artifacts never dirty the tree or enter a commit;
- **outside the compared inventory** — it is not under `.backlogit/queue/`,
  `.backlogit/archive/`, or `.backlogit/logs/`, so the recovery store can never
  be mistaken for, or collide with, the state being compared;
- **durable across the HALT** — it is the operator's diagnostic evidence.

#### 6.5.2 State classes (the discipline that makes recovery bounded)

Recovery treats workspace state as **three classes with different rules**.
Collapsing them is what makes naive "restore the backlog directory" unsound.

| Class | Members | Rule |
|---|---|---|
| **Mutable current-state** | `queue/*.md`, `archive/*.md` | **MAY be restored** byte-for-byte from the snapshot, by inventory only |
| **Append-only** | `logs/*.jsonl`, `hooks_queue.jsonl`, `stash.jsonl`, telemetry | **NEVER rewound, NEVER deleted, NEVER truncated.** The **full prior byte sequence MUST remain an exact prefix** of the current file — length comparison alone is insufficient. Receives an appended **recovery event** referencing the run |
| **Disposable cache** | `backlogit.db`, `-wal`, `-shm` | **NEVER byte-restored.** Rehydrated by the official `backlogit sync` |

Probe 19 measured this on both approval branches:
`APPEND_ONLY_PREFIX_VIOLATIONS=0` and `APPEND_ONLY_DELETIONS=0`, with the
verification comparing **every prior byte**, not merely the stream length, a
recovery event appended, and the engine's view restored from markdown alone with
`RECOVERY_STEP_5_DB_BYTE_RESTORED=False`.

#### 6.5.3 Procedure

Prerequisite, captured **before** the call: a durable pre-call snapshot of the
**mutable** source-of-truth files in the §6.2.2 baseline scope, plus an
**inventory manifest** recording each file's relative path and SHA-256, written
to the §6.5.1 recovery path.

1. **Enumerate** current paths in the baseline scope. *(non-destructive)*
2. **APPROVAL GATE — blocking.** In the live (non-probe) path, **explicit
   operator action-risk approval MUST be obtained BEFORE step 3 or step 4
   executes.** Enumeration (step 1) is non-destructive and runs first to preserve
   evidence; **quarantine is a destructive move and is gated together with
   restore.** If approval is withheld: **HALT** with no quarantine and no
   restore.
3. **Quarantine — bounded, move-never-delete (rev 5 — resolves a rev-5
   internal-review P1).** Quarantine is restricted to paths inside **the
   enumerated baseline-scope path set plus the destination paths implied by
   `required_ids`**. Engine lock files (`.locks/`, `*.lock`) and every other
   §6.2.2-excluded path are **never** quarantined. A path present now, absent
   from the inventory, and **inside** that bounded set is moved into the recovery
   path's `quarantine/` directory. An unexpected path **outside** that set is
   **reported in the recovery record and left in place — never moved.**
   - **Why this bound is mandatory.** The superseded `probe16-bounded-recovery.ps1`
     enumerated whole `queue/` + `archive/` directories and moved everything
     absent from its inventory; in that probe's 3-record sandbox it looked
     harmless (it relocated one engine lock file), but against the **live**
     `.backlogit/queue/`, which holds many unrelated records (`017-S`, `018-F`,
     `019-F`, …), the unbounded reading would relocate **unrelated live backlog
     artifacts out of the queue during recovery** — collateral damage strictly
     worse than the failure being recovered. **Probe 19** implements the bounded
     rule instead and demonstrates it: the out-of-scope engine lock file
     `.backlogit\queue\.001.001-T.md.lock` is **reported and left in place**
     (`RECOVERY_STEP_3_OUT_OF_BOUNDS_MOVED=0`) on **both** approval branches,
     while the single in-bounds unexpected path is quarantined by move.
4. **Restore only the enumerated mutable files** from the snapshot, by inventory.
   Never perform a blanket restore of the backlog directory. Never touch the
   append-only or disposable classes.
5. **Rehydrate the disposable index** by running the official `backlogit sync`.
6. **Verify equivalence** against the inventory: every enumerated mutable file
   present with a matching hash; **no unexpected residual paths *within the
   enumerated bounded set*** (the residual check is scoped to that set exactly —
   it is **not** a whole-directory assertion, because out-of-scope paths are
   legitimately present and deliberately untouched); and the **dependency, link
   and membership edges** of the baseline scope unchanged. Probe 19 records
   `MUTABLE_BYTE_EQUIVALENCE_FAILURES=0`, `MUTABLE_PATH_EQUIVALENCE_FAILURES=0`,
   `DEPENDENCY_EQUIVALENCE_OK=True` and
   `RESIDUAL_UNEXPECTED_PATHS_IN_BOUNDED_SET=0`.
7. **Append a recovery event** to the append-only stream — never a rewind.
8. **HALT** with a named token and a **P-005** event, **preserving** the
   quarantine directory and the full diagnostic evidence for operator review.

**Probe obligation — discharged by Probe 19 (rev 6).** Rev 5 rested this on
**Probe 16**, and claimed the "exact procedure" was proven. That claim is
**withdrawn**. Probe 16 diverged from this procedure in five ways, each now
recorded in the evidence-integrity table (§3): it swept whole `queue/` +
`archive/` directories instead of an enumerated set; it compared append-only
streams by **length only**, so any equal-length or prefix mutation was invisible;
it never executed the **approval-withheld** branch; it resolved its repo root by
`..\..\..` from a four-deep directory (landing on `docs\`), so the committed
script does not reproduce as written; and it ran against stock `backlogit init`
defaults rather than the live configuration.

**Probe 19** implements this procedure as specified and executes **both**
approval branches against a real errored invocation (`SHIP_EXIT=1`, destination
occupied by a directory, torn state observed: the shipment live in `queue/`
declaring `status: shipped`, a state the engine refuses to create directly):

| Assertion | Approval **withheld** | Approval **granted** |
|---|---|---|
| approval gate evaluated before any destructive step | **yes** | **yes** |
| files quarantined | **0** | **1** (moved, never deleted) |
| files restored | **0** | **4** (enumerated inventory only) |
| out-of-bounds path (`.001.001-T.md.lock`) | **reported, not moved** | **reported, not moved** |
| torn state preserved as evidence | **yes** | n/a (repaired) |
| mutable byte / path equivalence failures | n/a | **0 / 0** |
| dependency-edge equivalence | n/a | **True** |
| residual unexpected paths *in bounded set* | n/a | **0** |
| append-only **byte-prefix** violations | **0** | **0** |
| append-only deletions | **0** | **0** |
| cache rehydrated | n/a | official `backlogit sync` only; `DB_BYTE_RESTORED=False` |
| terminal disposition | **HALT** (no quarantine, no restore) | **HALT** (quarantine preserved) |

Both runs are seeded from the **live** `.backlogit` control files and gated on
the engine SHA-256 before any operation. Script and transcript are committed as
`probe19-bounded-recovery-v2.ps1` / `.txt`.

**Byte-prefix, not length.** §6.5.2's monotonic-growth rule is strengthened in
rev 6: an append-only stream satisfies the rule only if its **full prior byte
sequence remains an exact prefix** of the current file. Length comparison is
insufficient and is no longer accepted as evidence.

## 7. Verification obligations

### 7.1 Harness assertion set (reduced in rev 5)

Rev 4.1 required the harness to assert **29 "precision points"** — most of which
checked that a sentence *mentions* a concept ("must reject inference from a
failed/partial scan"). That is prose-shape testing: brittle, unfalsifiable, and a
large share of the unit's cost for little safety return.

Rev 5 replaces it with **one table-driven test** over a **12-row** assertion
table. Each row names a file, a required token, and the reason it is
load-bearing. The test fails if any token is absent from the installed file.

| # | File | Required token / property |
|---|---|---|
| 1 | `workflow-policies.md` | P-015 names a **second** exception, literally `TASK_ONLY_FINALIZE` |
| 2 | `workflow-policies.md` | the **§4.1 preserved region** of the fully-covered-root exception — preconditions **1–5**, **item 7**, and the **SUPERSESSION NOTE** — is present and **byte-for-byte unchanged**. (Scoped deliberately: lines 436, 443 and 448 **are** amended per §6.1 A4/A5/A6, so a whole-block "unmodified" assertion would contradict the intended edits and fail on a correct implementation.) |
| 3 | `workflow-policies.md` | Amendment Log has a new row naming `TASK_ONLY_FINALIZE` |
| 4 | `shipment-reconcile/SKILL.md` | classifier order `CASCADE` → `TASK_ONLY_FINALIZE` → `SAFE_CLOSE`/HALT |
| 5 | `shipment-reconcile/SKILL.md` | the five topology assertions **T1–T5** are stated |
| 6 | `shipment-reconcile/SKILL.md` | pre-mode and post-mode reconciliation retained (no bypass) |
| 7 | `shipment-reconcile/SKILL.md` | failure split: pre-call → fall-through; post-call/errored → recovery + HALT, **no fallback** |
| 8 | `shipment-reconcile/SKILL.md` | recovery names the three state classes and disclaims atomic rollback |
| 9 | `_ship.agent.md` | the Step 6 `Do NOT call … unless` clause names **both** exceptions |
| 10 | `_ship.agent.md` | `TASK_ONLY_FINALIZE` is recognized by name, so a fresh reload authorizes it |
| 11 | all three | no surviving "sole exception" / "only exception" phrasing **about the cascade being the single permitted alternative to safe-close**. The matcher is deliberately narrow: it targets the exclusivity claim, **not** the preserved block's legitimate "**only** when every one of the following preconditions holds" text, which must survive untouched (§4.1). |
| 12 | `_ship.agent.md` | Ship Role Boundary is **unchanged** (no new privilege) |

Rows 2 and 12 are **negative** assertions: they fail if the change widened
something it was not authorized to widen.

**The ≤4 helper functions are excluded from the "1 of 1" red count** (AC-8,
§6.1 D2). The red count is over *generated test functions*, of which there is
exactly one; helpers are not test functions and never report red.

### 7.2 What the harness does and does **not** verify — stated exactly (rev 6)

Rev 5's §7.2/AC-11 overstated what the Go harness covers (rev-5 open item O-4).
Rev 6 states the boundary exactly, and the acceptance criteria in §9 are
rewritten to match.

**What the ONE Go table-driven contract test actually proves**, and the only
things it proves:

1. **Presence** — each required token/phrase literally occurs in the installed
   file named by the row.
2. **Parity across the three files** — the same `TASK_ONLY_FINALIZE` identifier
   appears in all three instruction surfaces (rows 4, 5, 9, 10).
3. **Selector ordering as written text** — that the classifier order
   `CASCADE → TASK_ONLY_FINALIZE → SAFE_CLOSE`/HALT appears in the documented
   order (row 4). This is a **string-order** assertion over prose, not an
   executed classification.
4. **Required negative tokens** — that the §4.1 preserved region is byte-identical
   (row 2), that the narrow exclusivity phrasing is gone (row 11), and that the
   Ship Role Boundary is unchanged (row 12).

**What it does NOT prove, and must never be claimed to prove:**

- It does **not** execute G0–G13. No guard is evaluated; no topology is
  classified; no fixture is built.
- It does **not** execute P1–P10. No post-condition is evaluated.
- It does **not** execute `ShipShipment`, the §6.2.3 TOCTOU protocol, the §6.2.2
  baseline capture, the §6.5 bounded-recovery procedure, or any step of §8.
- It does **not** perform, simulate, or verify any destructive or recovery
  operation.

The harness reads installed markdown. It verifies **contract text**, not runtime
behavior. There are deliberately **no** executable negative cases in this unit.

**Engine behavior** is established separately and only by the **committed
runtime fixture probes** (§3), for the engine identity of record (§6.3). The
division of labour is:

| Claim | Established by |
|---|---|
| the contract *says* the right thing, in all three files | the one Go table-driven test |
| the engine *does* the right thing on the authorized topology | Probes 14, 15, **18** |
| an errored/partial call *recovers* boundedly, both approval branches | Probe **19** |
| the post-archive closure route's gate behavior | Probe **20** |
| Ship's PR-lifecycle-only route is not executable | Probe 17 |

Guards G0/G6/G7/G8/G11/G13, §6.2.3 isolation, §6.5 recovery and the §8 ordering
are **behavioral obligations on Ship**, verified by **review and diff** against
this plan — **not** by the harness and **not** by any automated gate in this
unit. Runtime enforcement (an executable classifier + negative fixtures) is
explicitly **deferred** to a separate release unit.

### 7.3 Probe obligations bound to supported topology

Supported runtime topology is bound **only** to shapes with an executed probe —
either **committed** (14–17, reproducible) or **retained-but-uncommitted**
(rev-4.1 `%TEMP%` runs, marked **degraded** per §3):

| Case | Probe | Supported? |
|---|---|---|
| one live task, no descendants | 2 *(degraded)* | **yes** |
| **exact `017-S` shape**: 1 live `done` task + 3 archived tasks + 8 archived subtasks | **15** *(committed)* | **yes — exact fixture** |
| pre-archived descendants outside manifest | 6 *(degraded)* | **yes** |
| `parent_id` retained on ship-archive of the live task | **15** *(committed)* | **verified** |
| merge SHA retained in the archived shipment record | **14, 15** *(committed)* | **verified** |
| blocks-edge suppression and release across the full lifecycle | **14** *(committed)* | **verified** |
| omitted queued subtask of a manifest item | 3 *(degraded)* | **no** — excluded by T5/G5 |
| omitted live sibling under parent | 4, 9 *(degraded)* | **no** — excluded by T5/G5, detected by P1/P10 |
| shared shipment membership | 7 *(degraded)* | **no** — excluded by G11 |
| unresolved blocking dependency / link expansion | 9 *(degraded)*, **14** *(committed)* | **no** — excluded by G13 |
| destination conflict / torn record | **16** *(committed)* | **no** — excluded by **G9 conjunct 2**, detected by P6 |
| `queued` shipment record | 11 *(degraded)* | **no** — excluded by G12 |
| **errored/partial invocation → bounded recovery** | **19** *(committed)* | procedure **executed on both approval branches** — see note |
| **post-archive closure route (O-6)** | **20** *(committed)* | lifecycle gate **blocks on all 4 routes**; §8.1.1 |
| **inbound-edge holder byte-invariance through close** | **18** *(committed)* | **verified** |
| engine-identity mismatch | — | **no** — excluded by §6.3 digest binding (HALT) |

Multi-**live**-member manifests have **no** probe and are therefore
**unsupported** (T2).

**Recovery-claim scoping (rev 6).** Rev 5 recorded the recovery row as "recovery
path **proven**". That phrasing is **withdrawn**: probe 16 exercised *a*
recovery, not *this* procedure (it swept whole directories rather than an
enumerated set, compared append-only streams by length only, and never ran the
approval-withheld branch). **Probe 19** implements and executes the procedure as
specified in §6.5.3, on **both** approval branches. The claim this plan makes is
therefore exactly: *the §6.5.3 procedure has been executed end-to-end against a
real errored invocation under live configuration, with byte/path/dependency
equivalence verified and both approval outcomes demonstrated* — not that the
procedure is proven correct for all failure modes.

## 8. Self-hosting closure sequence

### 8.1 The ordered sequence

The prerequisite shipment closes itself using the contract it merges. This
section is **self-contained** (rev 6): every step is stated here and aligned to
the **installed** `_ship.agent.md`, with the governing clause cited by line so
the sequence can be checked against the agent file without inference.

Rev 5's ordering was wrong in one material way: it opened and merged the closure
PR (step 10) **before** operational closure (11) and P-020 (12). The installed
Ship contract requires the opposite — `_ship.agent.md` **L741**: *"**After all
closure work is committed**, push the branch and create a PR"*. Closure work is
authored **on the closure branch first**, then pushed, then reviewed, then
merged. Rev 6 corrects this.

| # | Actor | Step | Installed-Ship basis |
|---|---|---|---|
| 1 | **Ship** | **Verify `021-S` is present on `main`** (read the merged `.backlogit/queue/021-S.md` on `main`, not on the branch). | §14.1 gate |
| 2 | **Ship** | **Claim `021-S`** — `queued -> active`. Required before any lifecycle gate can pass (**G12**; Probe 17/20 `LIFECYCLE_NO_ACTIVE_SHIPMENT`). | L169–207 pre_claim/post_claim gates |
| 3 | **Ship** | Harness (H0 red) → build → review → implementation PR on the implementation branch. | Step 5, L559 `TOPOLOGY_GATE: lifecycle (before build)`, L571 (before PR creation) |
| 4 | **Ship** | Merge the implementation PR (**P-009** merge commit). | Step 5 |
| 5 | **Ship** | **Verify the merge SHA** — two-parent merge commit confirmed via `merge-base --is-ancestor`; record it. | L731–733 Merge Confirmation Gate |
| 6 | **Ship** | Check out **fresh `main`**, cut the closure branch `post-merge/{feature_slug}` from it. | L734–738 |
| 7 | **Ship** | **Reload and positively verify**: re-read the merged **P-015**, **`shipment-reconcile`** and **`_ship.agent.md`**; assert the `TASK_ONLY_FINALIZE` tokens are present in **all three**, and that the step-5 merge commit is an ancestor of the checkout. | L756–764 (amended by **C1**, §6.1) |
| 8 | **Ship** | Run `shipment-reconcile` **pre-mode**. | L767 a0 `TOPOLOGY_GATE: lifecycle (before closure/safe-close)` — passes here because `021-S` is still **active** |
| 9 | **Ship** | Classify and close via **`TASK_ONLY_FINALIZE`** (§6.2 → §6.2.3 → §6.3 → §6.4), passing `--sha <merge_sha>` from step 5. | §6.3 |
| 10 | **Ship** | Evaluate **P1–P10 on the RAW post-call state**. No restoration of any kind first. | §8.2 |
| 11 | **Ship** | **P-007** archive-integrity restore **and re-verification**, in the order fixed by §8.2 — only if step 10 passed. | §8.2, P-007 |
| 12 | **Ship** | Run `shipment-reconcile` **post-mode**. | Step 6 |
| 13 | **Ship** | Run the official **`backlogit sync`**. Successor eligibility is conditional on it (Probe 14). | Probe 14 |
| 14 | **Ship** | **Verify** `021-S` is `archived` with `archived_status: shipped` and retains `commit: <merge_sha>`, and that `017-S` is now eligible. | Probes 14, 15, 18 |
| 15 | **Ship** | **Author operational closure and P-020 compaction ON the closure branch**, and commit them there. | **L741** — all closure work committed *before* the PR |
| 16 | **Ship** | `git push -u origin post-merge/{feature_slug}`; invoke `pr-lifecycle` to **open** the closure PR. | L741–744 |
| 17 | **Ship** | **Current-HEAD** local review, CI green, and **P-018** Copilot review on the closure PR; `--admin` may **not** bypass `COPILOT_REVIEW_BLOCK`. | L715, L718–722 |
| 18 | **Operator** | **Fresh explicit approval** — *"the prior main PR approval does not transfer."* | **L723** |
| 19 | **Ship** | Merge the closure PR as a **P-009 merge commit**. | L745–746 |
| 20 | **Ship** | Verify closure landed on `main`. | §14.1 |
| 21 | **Orchestrator** | **Only then** route the successor `017-S`. | §14.1 |

**Step 2 is new in rev 6 and closes rev-5 open item O-5 (claim).** Probe 17 and
Probe 20 both show every `--phase lifecycle` route fails closed with
`LIFECYCLE_NO_ACTIVE_SHIPMENT` unless exactly one shipment is **active**. Steps
3 and 8 both sit behind that gate, so the claim must precede them. **Step 13 is
the second half of O-5 (`backlogit sync`)**: Probe 14 makes successor
eligibility conditional on it, and rev 5's sequence never ran it.

**Step 7 must reload `_ship.agent.md` too.** This is why the file is surface 3
(§4): the clause a reloaded Ship actually reads (`_ship.agent.md` line 794) names
only the fully-covered-root exception. If it is not amended, step 7's positive
verification **fails** — and if the verification were skipped, step 9 would be
executed by an agent whose own contract forbids the call. Either way the sequence
deadlocks. Step 7 therefore verifies the token in **all three** instruction files
**and** verifies the merge commit, so a stale checkout cannot satisfy it.

#### 8.1.1 O-6 — the post-archive closure route, **probed** (rev 6)

Rev 5 recorded O-6 as the highest-risk open item: after step 9 archives the
shipment there are **zero active shipments**, so a topology gate firing on the
closure PR would deadlock **after** the destructive call. **Probe 20** measured
it rather than reasoning about it, and the answer has two separable halves.

**Q1 — does a lifecycle gate block post-archive? YES, on every route.** Probe 20
drove a live-config-seeded fixture to the exact post-archive state and ran all
four routes:

| Route | Exit | Token |
|---|---|---|
| `--mode agent --shipment <archived> --phase lifecycle` | **1** | `LIFECYCLE_NO_ACTIVE_SHIPMENT` |
| `--mode agent --phase ambient` | **2** | `agent mode requires --shipment` |
| `--mode manual --shipment <archived> --phase lifecycle` | **1** | `LIFECYCLE_NO_ACTIVE_SHIPMENT` |
| `--mode ci --phase lifecycle` | **2** | `--phase lifecycle requires --shipment` |

`POST_ARCHIVE_ROUTES_NONZERO_EXIT=4`. There is no active shipment left to
satisfy the gate, and re-claiming an archived shipment is not a supported
transition — so this deadlock, if entered, is **unrecoverable in-band**.

**Q2 — does the installed Ship contract actually run that gate before the
closure PR? NO.** Probe 20 step 5 cites the installed file directly: the
lifecycle topology gate is mandated at **L559** (before build), **L571** (before
*implementation* PR creation), and **L767** (before closure/safe-close, while the
shipment is still active). The closure-PR step at **L741–744** invokes
`pr-lifecycle` **without** a topology-gate precondition.

**Disposition — not a deadlock, but a one-clause-deep hazard.** The §8 sequence
as ordered above does not enter the deadlock, because no step between 15 and 19
runs a lifecycle-phase topology gate. This is **not** a comfortable margin: any
agent that defensively runs the lifecycle gate before the closure PR — a
plausible reading of "run the topology gate before `pr-lifecycle`" generalized
from L571 — **will** deadlock with the destructive call already committed.

Therefore, and normatively:

- **Steps 15–19 MUST NOT run `autoharness gate pipeline-topology --phase
  lifecycle`.** The gate's precondition is structurally unsatisfiable there.
- If a future harness change adds such a gate to the closure-PR path, this
  sequence is **invalidated** and must return to Stage before any close.
- The residual is recorded in §10.3 as **accepted-with-tripwire**, not as
  "proven safe".

### 8.2 P-007 ordering relative to P1–P10 (rev 5, normative)

Rev 4.1 said only that P-007 is "handled in explicit order" without saying
**where**. That ambiguity is dangerous, because P-007's remedy
(`git restore .backlogit/archive/`) is a **filesystem mutation that repairs
deletions** — exactly the evidence P1–P10 exist to detect. Running the repair
first would silently erase the signal.

The fixed order is:

1. **`ShipShipment` returns** (§6.3).
2. **Evaluate P1–P10 on the RAW post-call state.** No restoration, no `git
   restore`, no repair of any kind may run before this. Detection precedes
   remedy, always.
3. **Only if P1–P10 all pass**, run the P-007 archive-integrity check:
   `git status -- ".backlogit/archive/"`. If archive files appear as working-tree
   deletions, restore them (`git restore .backlogit/archive/`).
4. **Re-verify** P4/P6/P9/P10 after any P-007 restoration, so the restore itself
   is proven to have returned the tree to the expected state and nothing else.
5. If **any** of P1–P10 failed at step 2, do **not** run P-007 restoration. Go
   directly to §6.5(B) bounded recovery + HALT: the recovery procedure's own
   enumerate → approve → quarantine → restore sequence supersedes it, and it
   preserves the deletion as diagnostic evidence instead of quietly repairing it.

**Merge-metadata preservation and verification.** The merge SHA from **step 5**
is passed at §6.3 and **verified twice**: by **P7** (the archived shipment record
reports `archived_status: shipped` **and** retains `commit: <merge_sha>`) and
again at item 4 of the ordering list above after any P-007 restoration. Probes
14, 15 and **18** all confirm the archived shipment record carries
`status: archived`, `archived_status: shipped`, and the `commit` value passed to
`--sha`.

**Eligibility at step 9.** The harvested artifacts MUST satisfy:

- the covering feature has **exactly one** descendant at any depth — the sole
  task — so T5/G5 holds trivially, T2 holds, and P1 `returned_ids == []` follows;
- the feature is root and excluded from the manifest (T4);
- the task is `done` at closure (T2) and the shipment is `active` (G12);
- no record declares this `shipment_id` outside the manifest (G8);
- no shared membership (G11); and for G13, the `017-S depends_on <this shipment>`
  edge is an **inbound** edge on `017-S`, not an edge held by any member of this
  manifest. G13 explicitly scopes inbound edges (§6.2.1), and this one resolves
  to *this* shipment, which is being archived by the very call under evaluation —
  Probe 14 shows an archived predecessor **satisfies** the edge rather than
  leaving it unresolved. G13 is satisfied.

**Non-circular:** the authorization used at step 9 is the *merged* artifact from
step 4, not the in-flight working copy. Step 7 (reload + positive verification
across all three files, plus merge-commit ancestry) is the load-bearing ordering
constraint, bound as AC-9.

### 8.3 Contingency — first real self-close fails after the implementation merged

If step 9 fails after step 4 has already merged:

1. Execute §6.5 path (B): bounded recovery, then HALT.
2. Leave the prerequisite shipment **unresolved** — do not force it to any
   terminal state, and specifically do **not** attempt `move --status shipped`
   (the engine refuses it, exit 9) or the cascade path.
3. Leave `017-S` **blocked** by the dependency edge. Probe 14 confirms an
   unclosed predecessor keeps the successor out of the ready set, so this
   requires no extra action — the edge does it.
4. Halt for operator review with the recovery report and the preserved
   quarantine directory.
5. **No admin fallback. No cascade fallback. No alternate close path.** The
   merged contract stays on `main`; it is **inert** until a subsequent,
   separately reviewed session addresses the failure. An inert-but-merged
   contract is safe: it authorizes a path that is never entered, and every
   existing closure path is unchanged.
6. The failure is captured as a stash entry for Stage intake, carrying the
   recovery token, the quarantine path, and the engine digest in force.

## 9. Acceptance criteria (for the harvested task)

- **AC-1** P-015 gains a second named exception authorizing `TASK_ONLY_FINALIZE`; the **preserved region defined in §4.1** (preconditions 1–5, item 7, SUPERSESSION NOTE) is **byte-for-byte unchanged** (verified by diff), while the header sentence and item 6 are amended to remove exclusivity; Amendment Log entry appended.
- **AC-2** `shipment-reconcile` documents the classification, evaluated **after** `CASCADE` and **before** `SAFE_CLOSE`/HALT, with the mutual-exclusivity rationale (§6.1).
- **AC-3** The five topology assertions **T1–T5** and guards G0–G13 are **specified in contract text** (§6.2.0, §6.2.1). **Verification method: review/diff, plus §7.1 row 5 token presence only.** The Go harness asserts that T1–T5 are *stated*; it does **not** evaluate any of them and does **not** execute G0–G13.
- **AC-4** All of P1–P10 are **specified in contract text**, with `allowed_ids`/`required_ids` defined normatively. **Verification method: review/diff.** No automated verification exists in this unit; the Go harness does **not** execute P1–P10.
- **AC-5** P1 is justified by reference to G5; P4 requires independent filesystem diffing derived from `required_ids`; P9 post-compares each task's declared status, single location, and `parent_id`; P10 asserts descendant byte-identity. **Verification method: review/diff.**
- **AC-6** §6.5 specifies **two separate** failure paths: (A) pre-call guard failure → existing SAFE_CLOSE/HALT with **no** destructive call and **no** recovery; (B) post-call **or any errored/indeterminate** invocation → bounded recovery + HALT, **never** a fallback close path. **Verification: §7.1 row 7 (token presence) + Probe 19 (executed behavior).**
- **AC-7** **Safe-Close step 8** — enumerated as **§6.1 site B11** (`shipment-reconcile/SKILL.md` lines **591–605**) — carries a fail-closed cross-reference to the new path; step 8's existing text is otherwise unchanged. This site is in **Inventory Table §6.1-B** and is budgeted in §10.1.
- **AC-8** P-004: at H0 `go vet ./...` exits 0, `go test ./...` exits non-zero, the marker `not implemented: task-only finalize contract` is present, red count **1 of 1 over generated test functions** (the ≤4 helpers are **not** test functions and are excluded from the count). At H1 the marker is gone and full `go test ./...` exits 0.
- **AC-9** The task records the §8 step 6–7 fresh-`main` checkout, reload and **positive verification of merged tokens in all three instruction files** (including `_ship.agent.md`) **and of the step-5 merge-commit ancestry**, as an execution obligation on Ship.
- **AC-10** The generated function executes unconditionally — no `t.Skip`, build tag, env gate, or selector flag.
- **AC-11** No Ship Role Boundary change, no generic `move shipment → shipped`, no pre-mode bypass, no unrestricted cascade. Asserted **negatively, as token/text assertions only**, by harness rows 2 and 12 (§7.1). This is a *contract-text* guarantee; it is **not** a runtime guarantee, and §7.2 states the boundary exactly.
- **AC-12** The authorized call passes merge-commit metadata (`--sha`, and `--message`/`--author` where declared); **P7** re-asserts `archived_status: shipped` **and** `commit` retention; §8.2 fixes P-007 ordering **after** P1–P10 with a re-verification pass. **Verified by Probes 14, 15, 18.**
- **AC-13** All three instruction surfaces are reconciled against the **closed inventory in §6.1 — Inventory Tables §6.1-A, §6.1-B, §6.1-C** (28 sites; 20 MUST + 8 CONSISTENCY). The implementing task re-runs the scan and commits the resulting inventory to the path fixed in **§6.1-E**; any clause found by the re-scan but absent from the tables is added, never skipped, and a newly-found **MUST** site invalidates the §10.1 budget and returns the unit to Stage. Mandatory pre/post reconciliation steps are retained unchanged.
- **AC-14** §6.2.2 baseline scope is **specified in contract text**, including **inbound** edge/link/membership holders outside the parent subtree, and the explicit exclusion of expected pre-mode report/lock outputs from mutation comparison. **Verification method: review/diff for the specification; Probe 18 for the inbound-edge byte-invariance claim specifically.** The Go harness verifies neither.
- **AC-15** §6.2.3 TOCTOU isolation is **specified in contract text**: the four-part execution-mode precondition (one agent, one worktree, no other active checkpoint/session, no concurrent backlog mutation), bootstrap enumeration of the **relation closure** (including existing outside inbound/outbound relationship holders), locks over that closure, re-run under lock, **re-scan and re-hash immediately before the call**, fail-closed on any difference **or any relevant new or changed record**, and an honest statement that the claim is **detection, not prevention against external actors**, with HALT if the precondition is violated. **Verification method: review/diff.** The Go harness does **not** execute the protocol.
- **AC-16** §6.3 records the engine version **and the SHA-256 digest** of the executable, and **HALTs** on mismatch of either, requiring a probe refresh. No advisory fallback exists.
- **AC-17** The unit contains **exactly one** task and one generated (table-driven) function; no sequential split is permitted.
- **AC-18** The harness is **one table-driven test** over the 12-row assertion table in §7.1, including the two negative rows, plus **≤4** non-test helper functions.
- **AC-19** The committed probe evidence directory `docs/plans/evidence/2026-09-12-task-only-shipment-finalization/` is referenced by the contract as the **engine-behavior** basis, and the recorded digest matches the engine used at closure. **Scope statement:** that directory establishes *runtime engine behavior* via executed fixture probes (14, 15, 17, 18, 19, 20); it does **not** establish contract text, and the Go harness does **not** establish engine behavior. Neither substitutes for the other.

## 10. Sizing, risks, residuals

**Sizing.** `size: M`, `complexity: high` — recorded on the harvested task as
structured `size` (with `size_source: agent` and a `size_ruleset_version`) plus
`complexity` carried as **labeled prose**, because this workspace's
`.backlogit/header-def.yaml` defines **no** `complexity` field on the `task`
type. That partial degradation is flagged explicitly in the Stage report.

**Rev-6 note.** The recorded `size: M` is now **contradicted** by §10.1's
recalculated budget (~2.5–2.7 h). The task record is deliberately **not**
re-sized in this remediation cycle, because re-sizing would change harvested
state while the unit is returning **MUST_REPLAN**; the correct fix is made by
the re-planning cycle, not smuggled into a remediation commit.

### 10.1 Concrete edit budget (rev 6) — the 2-hour claim, **recalculated and NOT closing**

Rev 5 itemized a budget over a **14-site** inventory and simultaneously carried
three *known* further sites as open items (O-1/O-2/O-3). Rev 6 closed the
inventory (§6.1) to **28 sites**. The budget must be recalculated against the
closed inventory, and when it is, **it does not close**.

| Work item | Unit | Count | Basis |
|---|---|---|---|
| **MUST** clause edits — mechanical phrase reconciliation | small pre-located text edits | **9** | §6.1: A2, A3, B1, B5, C3, C4, C5, B15, B17-adjacent |
| **MUST** clause edits — structural rewrites | multi-sentence clause rewrites | **8** | §6.1: A4, A5, A6, B3, B8, B13, C1, C2 |
| **MUST** selector edits | ternary classifier surgery | **3** | §6.1: B9, B10, C6 |
| **CONSISTENCY** edits | coherence-only text edits | **8** | §6.1: A1, B2, B4, B6, B7, B12, B14, B16 |
| AC-7 step-8 cross-reference | one fail-closed pointer | **1** | §6.1 **B11** (was rev-5 open item O-2) |
| New P-015 exception block | one additive normative block | **1** | §6.1 A7 — mirrors the existing exception's shape |
| New `shipment-reconcile` classification section | one **large** additive normative block | **1** | §6.1 B18 — T1–T5 + G0–G13 + P1–P10 + failure split + recovery |
| Amendment Log row | one table row | **1** | §6.1 A8 |
| Committed clause inventory | one generated artifact | **1** | §6.1-E |
| Go harness | one table-driven test function | **1** | 12-row table (§7.1) |
| Go helpers | small file-read/assert helpers | **≤4** | `readContract`, `assertContains`, `assertAbsent`, `assertPreservedRegion` |

**Closed totals: 28 normative edit sites + 2 additive normative blocks + 1
Amendment Log row + 1 committed inventory + 1 table-driven test function + ≤4
helpers, across 4 files.**

#### Why the budget does not close (the honest arithmetic)

| Work item | Estimate |
|---|---|
| 9 mechanical clause edits @ ~1.5 min | ~14 min |
| 8 structural clause rewrites @ ~4 min | ~32 min |
| 3 selector edits @ ~6 min | ~18 min |
| 8 consistency edits @ ~1.5 min | ~12 min |
| B11 step-8 cross-reference | ~3 min |
| A7 new P-015 exception block | ~15 min |
| **B18 new classification section** | **~30 min** |
| A8 Amendment Log row | ~1 min |
| §6.1-E committed inventory | ~5 min |
| Go table-driven test + ≤4 helpers | ~22 min |
| H0→H1 verification, `go vet`, `go test`, markdownlint | ~10 min |
| **Total** | **~162 min ≈ 2.7 h** |

Even on the **MUST-only** reading (dropping all 8 CONSISTENCY edits, which
leaves the contract coherent but internally uneven) the total is **~150 min ≈
2.5 h**. Both readings exceed the 2-hour rule.

The single largest irreducible item is **B18**: the `shipment-reconcile`
classification section is the *normative home* of the whole contract — T1–T5,
G0–G13, P1–P10, the two-path failure split and the bounded-recovery procedure.
The plan can point at it, but the plan does not ship; the skill does. It cannot
be compressed to a pointer without moving the contract somewhere that Ship does
not reload.

#### Verdict: **MUST_REPLAN**

§10.2 establishes that this unit **cannot be split** — T2 admits exactly one
live manifest member at closure, so a second task under `022-F` breaks the §8
self-hosting topology. §5 and AC-17 restate the same constraint.

So the two constraints are jointly unsatisfiable as currently scoped:

- the work is **~2.5–2.7 h**, which violates the 2-hour rule; and
- the work **cannot be split**, which is the only in-band remedy.

Per §10's own stated contingency ("if the work is judged to exceed the 2-hour
budget, the correct response is **HALT and return to Stage**"), this plan
therefore returns **MUST_REPLAN**. Rev 6 does **not** paper over the overrun by
re-labelling known sites as open items, which is precisely what rev 5 did.

**The harvested topology is deliberately left untouched** (`022-F`,
`022.001-T`, task-only `021-S`, `017-S -> 021-S`). Re-scoping is a *future*
Stage decision, not a silent edit inside a remediation cycle. Candidate
directions for that decision, none of them taken here:

1. **Reduce the contract surface** — land `TASK_ONLY_FINALIZE` with the
   CONSISTENCY edits deferred to a follow-up hygiene unit, and re-measure.
2. **Move B18's bulk into P-015** so the skill carries a short normative
   pointer, accepting that the skill is then not self-contained.
3. **Accept a documented, operator-authorized 2-hour-rule exception** for this
   one structurally-indivisible unit, recorded as a P-005 deviation with
   rationale.
4. **Change the self-hosting approach** so the unit need not close itself,
   which dissolves the T2 one-live-member constraint and re-enables splitting.

**Honest limit.** This is a drafting task with pre-located edits, which is why
rev 5 believed it fit. What changed in rev 6 is not the work — it is the
*measurement*. Stage does not claim the estimate is beyond challenge; it claims
the estimate is now itemized against a **closed** inventory and therefore
falsifiable in both directions.

### 10.2 Why splitting is unavailable (structural, not stylistic)

**T2** admits **exactly one** live manifest member at closure. A task moved to
`done` declares `done` (not `archived`), so it remains a *live* member; two
completed tasks would present **two** live members and the prerequisite could not
classify itself under the very contract it ships (§8). Splitting breaks
self-hosting.

Independently, the three instruction files must be **mutually consistent before
the unit can self-close** (§4): a partial landing leaves P-015 authorizing a path
that `_ship.agent.md` still forbids. That is an incoherent contract, not merely
an inconvenient one.

### 10.3 Risks and residuals

- **Rollback of the release itself:** four files on a feature branch. Revert the
  merge commit; closure behavior returns to today's blocked state.
- **Failure mode — over-broad drafting.** Mitigated by conjunctive fail-closed
  guards, the five explicit topology assertions, and the two **negative** harness
  rows (§7.1 rows 2 and 12) that fail if the change widened anything.
- **Failure mode — guard drift.** Residual: **accepted**; a CI coupling gate is
  deferred to a separate unit.
- **Residual — runtime enforcement deferred** (§7.2). The harness proves contract
  *text*, not engine *behavior*; behavior is proven by the committed probes.
- **Residual — O-6 post-archive gate hazard (rev 6, accepted WITH TRIPWIRE).**
  Probe 20 proves a lifecycle-phase topology gate **blocks on all four routes**
  once the shipment is archived, and the state is not recoverable in band. The
  §8 sequence avoids it only because the installed closure-PR path (L741–744)
  does not invoke that gate. **Tripwire:** if any harness change adds a
  lifecycle-phase topology gate to the closure-PR path, §8 is invalidated and
  must return to Stage before any close.
- **Residual — authorization surface is CLI-only** (§3.0). The MCP surface is
  unprobed and unauthorized; no equivalence is claimed.
- **Residual — probes 14/15/16 remain stock-`init`-seeded.** Probe 18 shows the
  config delta is material; load-bearing claims are re-derived under live config
  by probes 18–20, the rest is carried here.
- **Residual — multi-live-member and multi-parent manifests unsupported** (T2/T4).
- **Residual — engine-identity coupling.** Safety claims bind to the SHA-256 in
  §6.3. An upgrade **HALTs** and requires re-running every probe in the evidence
  directory. This is now a hard gate, not a heuristic.
- **Residual — TOCTOU new-record creation.** Excluded by the single-agent /
  single-worktree / P-001 execution-mode precondition rather than by a lock. If
  that precondition is violated or a new in-scope record is observed, the
  procedure **HALTs** (§6.2.3). No claim of absolute prevention is made.
- **Residual — recovery is bounded, not atomic.** No transaction exists; §6.5
  restores an enumerated mutable file set, never rewinds append-only state, and
  never restores DB bytes.
- **Residual — `017-S` stays blocked** until this prerequisite ships. Probe 14
  confirms the edge enforces this without further action.
- **BLOCKER — PR #54 has no agent-executable route.** See §11. This is an open
  blocker returned to the operator, not a residual Stage has closed.
- **BLOCKER — the §10.1 budget does not close (rev 6).** The unit prices at
  ~2.5–2.7 h against the closed 28-site inventory and cannot be split (§10.2).
  Verdict **MUST_REPLAN**; harvested topology deliberately left unchanged.

## Plan Hardening

*(Literal section title, current as of rev 6. **Probes 14–20** referenced below
are committed with script and transcript under
`docs/plans/evidence/2026-09-12-task-only-shipment-finalization/`; **probes
18–20 are additionally live-config-seeded and digest-gated in-script** (§3.0).
**Probes 1, 3, 4 and 9** are retained rev-4.1 characterizations **without**
committed transcripts — see §3's evidence-degraded table. **Probe 16 is
superseded by Probe 19** and is no longer cited as evidence for the §6.5.3
procedure.)*

1. **Destructive-operation adjacency.** The authorized call mutates the backlog
   with no transaction. Mitigated by T1–T5 plus G0–G13, §6.2.3 isolation, P1–P10,
   and the §6.5(B) bounded recovery **executed on both approval branches** by
   **Probe 19**.
2. **Silent scope expansion is reachable and confirmed damaging.** Probe 3
   archives omitted descendants; Probes 4 and 9 return omitted live siblings —
   and Probe 9 proves the return **strips `parent_id`**, orphaning the record and
   doing so **across a shipment boundary**. T5/G5 and P1/P10 are correctness
   requirements derived from executed evidence.
3. **Recovery is bounded, not atomic rollback.** Stated plainly in §6.5. No
   DB/WAL/SHM byte restoration is claimed; the index is a disposable cache
   rehydrated by the official `backlogit sync`. **Probe 19 shows a partial
   failure is real**: the engine applied mutations and *then* errored, leaving the
   shipment record declaring `status: shipped` while still in the queue directory
   — a torn state the engine itself refuses to create directly (Probe 18, exit 9).
   Bounded recovery restored byte/path/dependency equivalence
   (`MUTABLE_BYTE_EQUIVALENCE_FAILURES=0`,
   `MUTABLE_PATH_EQUIVALENCE_FAILURES=0`, `DEPENDENCY_EQUIVALENCE_OK=True`) and
   the engine's own view, with `APPEND_ONLY_PREFIX_VIOLATIONS=0` verified over
   **every prior byte**, not merely stream length. **Live destructive recovery
   requires explicit operator action-risk approval BEFORE quarantine or restore
   — and Probe 19 executes the withheld branch to demonstrate it.**
4. **TOCTOU.** §6.2.3 locks the **entire relation closure** — including existing
   records outside the parent subtree that hold **inbound or outbound** edges,
   links, or shipment membership (explicitly including the `017-S -> 021-S`
   record, proven byte-identical through close by **Probe 18**) — re-runs
   enumeration under lock, and **re-scans and re-hashes immediately before the
   call**, HALTing on any relevant new or changed record. New-record creation is
   excluded by the enforced **one-agent / one-worktree / no-other-active-
   checkpoint-or-session / no-concurrent-mutation** execution mode, not by a
   lock. **The claim is DETECTION, not prevention against external actors.**
5. **Guard scoping is itself a hazard.** A blanket "any subtask member
   disqualifies" rule would have rendered the path unusable on `017-S`, the one
   shipment it exists to close. **T3** scopes task-only to **live** members on the
   executed evidence of **Probe 15**, which measured the **exact** `017-S` member
   set (3 archived tasks + 8 archived subtasks) and found every member byte, path
   and `parent_id` unchanged. Over-tightening a safety guard can destroy the
   objective as surely as under-tightening destroys safety.
6. **Self-hosting.** §8.1's **twenty-one-step** sequence is mandatory and
   ordered; step 2 (**claim**, required for every lifecycle gate), step 7 (reload
   + positive verification on fresh `main`, across **all three** instruction
   files **and** the merge commit, AC-9), step 13 (**official `backlogit sync`**)
   and step 15 (**operational closure + P-020 authored on the closure branch
   before the PR opens**, installed Ship L741) are all load-bearing. §8.2 fixes
   P-007 ordering **after** P1–P10 so a repair can never erase detection
   evidence. §8.1.1 records the measured post-archive gate hazard (O-6) and
   **prohibits** running a lifecycle-phase topology gate in steps 15–19. §8.3
   gives the no-fallback contingency.
7. **Contract/engine divergence.** `header-def.yaml` permits shipment
   `status: shipped` while the engine refuses the generic transition. The path
   depends on *engine* behavior established by the committed probes, never on the
   schema enum. §6.3 binds the engine by **SHA-256 digest**, and a mismatch is a
   **HALT** requiring a probe refresh — there is no advisory fallback.
8. **Blast radius.** P-015, `shipment-reconcile` and `_ship.agent.md` govern every
   shipment closure workspace-wide. Mitigated by additive-only drafting,
   byte-for-byte preservation of the existing exception, mutual exclusivity
   (§6.1), the **closed 28-site reconciliation inventory** that prevents a
   self-contradictory contract, and two **negative** harness assertions (§7.1
   rows 2 and 12) that fail if anything was widened.
9. **Verification honesty.** §7.2 states exactly what the one Go table-driven
   test proves and does not prove; §7.3 binds supported topology strictly to
   shapes with a **committed** probe, and §3.0 adds a probe-vs-contract fidelity
   table plus the **CLI-only** authorization statement. Rev 4.1's
   `%TEMP%`-and-deleted probes are **withdrawn as evidence** (§3).
10. **An executing agent that cannot execute is a planning defect, not a runtime
    surprise.** Rev 5 tested the actor assumption instead of asserting it, and
    found the PR-lifecycle-only Ship route **unexecutable** (Probe 17, §11). The
    blocker is returned to the operator rather than papered over.
11. **A budget that does not close is a planning defect, not an execution
    problem (rev 6).** Closing the §6.1 inventory raised the honest cost to
    ~2.5–2.7 h in a unit that structurally cannot be split (§10.2). Rev 6
    returns **MUST_REPLAN** rather than re-labelling known edit sites as open
    items to keep the estimate inside the 2-hour rule.

## 11. Authorized actors and PR boundary (rev 6, normative) — **BLOCKER**

This plan's artifacts ship as **one artifacts-only Stage PR**, PR #54 (§14).

### 11.1 Adjudication requested, and what testing found

Stage was directed to reconcile the actor artifacts so the governing planning
artifacts consistently permit **Ship OR operator** as PR actor, on the ground
that Ship's Role Boundary explicitly classifies Git branch/commit/push and PR
create/update/merge as **Allowed** — with the standing instruction to *return a
blocker rather than assert feasibility* if the installed Ship mandatory sequence
makes it impossible.

**It does make it impossible.** Stage tested it rather than reasoning about it.

Ship's Role Boundary table is necessary but **not sufficient**. It grants
*categories*; the **Required Steps** impose an additional mandatory gate that a
PR-lifecycle-only session cannot satisfy. Ship Step 5 items 1a and 5a require,
before invoking `pr-lifecycle`:

```text
autoharness gate pipeline-topology --mode agent --shipment {shipment_id} --phase lifecycle --json
```

with the explicit handling rule *"exit 1/2 halts immediately with the reported
token/message (never inferred, never fail-open)."*

The gate is installed and live in this workspace
(`.autoharness/gates/pipeline-topology-force-audit.log` records real operator
`--force` use on 2026-09-05). **Probe 17** exercised all four routes:

| Invocation | Exit | Token / message |
|---|---|---|
| `--mode agent --phase lifecycle` (no shipment) | **2** | `agent mode requires --shipment <shipment_id>` |
| `--mode agent --shipment 021-S --phase lifecycle` (queued, unclaimed) | **1** | `LIFECYCLE_NO_ACTIVE_SHIPMENT: expected exactly one active shipment` |
| `--mode manual --phase lifecycle` (no shipment) | **2** | `--phase lifecycle requires --shipment <shipment_id> in any mode` |
| `--mode agent --phase ambient` (no shipment) | **2** | `agent mode requires --shipment <shipment_id>` |

**Every route fails closed.** The gate passes only with **exactly one active
shipment**, and a shipment becomes `active` only by being **claimed** — precisely
what a PR-lifecycle-only session must not do, and what §12/§14 forbid before this
prerequisite merges. `--force` is documented *"Operator-only … Audited. **Never
reachable from an agent surface.**"*

### 11.2 Recorded actor boundary

- **The OPERATOR is the authorized PR actor for PR #54.** Only the operator can
  satisfy the gate (or use the audited operator-only `--force`).
- **Ship is NOT a viable PR actor for PR #54** on this installed harness. This is
  a *sequence* limitation, not a Role Boundary limitation — Ship's Role Boundary
  genuinely permits push and PR merge, which is why the claim was plausible
  enough to require testing.
- **Stage owns all planning edits.** A review comment requiring a planning or
  contract change returns to **Stage**. Ship, in any mode, is forbidden by its own
  Role Boundary from modifying planning content ("Planning | … | Create or modify
  deliberation, spike, plan, or review artifacts" = Forbidden).
- **Stage and Orchestrator never push and never merge.** Orchestrator is
  routing-only.

### 11.3 Consistency with `018-F` / `018.008-T`

The pre-existing operator-only language in `018-F` (items 4, 44, 48, 49) and
`018.008-T` is therefore **confirmed, not contradicted**, and requires no
weakening. Those artifacts already say the operator — "never Stage, and never the
Orchestrator" — pushes the Stage artifact branch and opens/approves the staging
PR. Probe 17 independently re-derives the same conclusion for Ship, closing the
gap that the earlier "no agent performs any of these steps" phrasing left
unproven. A cross-reference is recorded on both artifacts pointing here.

### 11.4 Unblocking options (operator decision, not Stage's to take)

1. **Operator carries PR #54** — available today, no change required.
2. **Amend Ship's Step 5 gate invocation** to admit an artifact-only,
   no-shipment PR mode — a *separate*, separately-reviewed release unit that
   changes an agent contract. Out of scope here and **not** something this plan
   may self-authorize.

Stage takes neither. It records the blocker and halts at the boundary.

## 12. Plan review record

- **Rev 1 → FAIL.** 5×P1 + 2×P2: skill-only edit leaves P-015's prohibition
  intact; G5 missed the omitted-sibling topology and the orphan scan;
  `required_ids` undefined and `archived_ids` over-trusted; blanket `git restore`
  unsound; self-hosting eligibility used the wrong descendant condition;
  characterization overclaimed; P-004 full-suite red omitted.
- **Rev 2 → FAIL** (narrowed). Remaining: no post-comparison of `parent_id`;
  torn-record check omitted the parent; rollback covered only post-guard failure
  and could not restore the gitignored cache; `live_descendants == manifest`
  contradicted G4/P8; sibling safety implied without verification; merge metadata
  omitted.
- **Rev 3 → ADVISORY** (attempt 3). All P1 closed; three P2 fixed in place.
- **Rev 3 → independent adversarial review: `MUST_REPLAN`.** Findings: the
  two-Stage-PR sequence creates an unsafe eligibility window and is unnecessary;
  ID allocation is branch-local so the prerequisite must be harvested on the
  policy branch; Ship — not Orchestrator — is the authorized PR actor; Stage
  Step 5.5 parent-first creation does **not** make task-only assembly impossible
  because `return_blocked` is a registered operation within Stage's Role Boundary.
- **Rev 4** discharges the adversarial findings. Sequencing collapses to **one**
  artifacts-only Stage PR (§14); harvest moves to the policy branch; §11 records
  the Ship PR-lifecycle-only boundary; Probe 8 verifies `return_blocked`
  directly; the topology is narrowed (G10–G13); baseline scope and TOCTOU
  isolation are specified (§6.2.1, §6.2.2); recovery is re-framed as bounded and
  non-atomic; failure semantics split into two paths; classifier order and clause
  reconciliation are bound (§6.1, AC-13); the self-hosting sequence is completed
  with a contingency (§8, §8.1); the two-task split allowance is **removed**
  (§5, AC-17); the literal `## Plan Hardening` section is added; and the engine
  version is recorded (§6.3).
- **Rev 4 → internal plan review: FAIL** (two independent reviewers, convergent).
  One **P1**: G1's blanket "any subtask member disqualifies" would have **rejected
  `017-S` itself** — whose manifest is one live task plus eleven archived members,
  eight of them `subtask` — so `TASK_ONLY_FINALIZE` could never classify the
  shipment it exists to close, contradicting §1, §3.1, G5 and §7.3. Verified
  directly against the live manifest. Four **P2**: Probe 10 did not exercise an
  errored/partial invocation and so could not discharge the §6.5 probe obligation;
  P9's `parent_id`-retention-on-ship-archive was an unverified assumption;
  the `+dirty` version string is not a unique build identity; §7.1's mapping
  omitted ten guards; and the M/2-hour estimate was questioned against AC-17's
  no-split rule.
- **Rev 4.1** closes all of them with executed evidence rather than redrafting:
  **G1/G2 are re-scoped to live members** (archived `task`/`subtask` members
  permitted, features and shipments still disqualifying unconditionally), with the
  `017-S` rationale stated inline; **Probe 5b** runs an `017-S`-representative
  fixture covering both archived member types
  including an archived subtask member and additionally **verifies `parent_id`
  retention on ship-archive**, making P9 satisfiable; **Probe 13** injects a real
  destination conflict, observes a non-zero exit *after* partial mutation leaving
  a torn record, and executes §6.5(B) to a byte/path-equivalence PASS with the
  engine view restored — discharging the partial-failure obligation; §6.3 records
  the `+dirty` limitation and requires a binary digest (AC-16); §6.2.2 is
  re-ordered to resolve the bootstrap-lock circularity and **downgrades its claim
  from "closes" to "narrows"** with the residual named; §7.1 gains rows for all
  twenty-nine guards and sections; AC-13 now requires an enumerated, committed
  clause inventory; and §10 records the budget tension as an accepted residual
  with a HALT-and-re-plan contingency, explaining why G10 makes splitting
  structurally unavailable.
- **Rev 4.1 → internal re-review: ADVISORY** (final local cycle). The P1 is
  confirmed **substantively closed**: the live-member rescoping admits `017-S`
  while G5/G10 still exclude the Probe 3/4/9 hazards, and permitting archived
  subtask members opens **no new cascade vector** (any live descendant beneath an
  archived node is still caught by G5's every-depth walk). All five P2 items
  confirmed closed. Three residual P2 accuracy defects were raised and **fixed in
  place**: Plan Hardening item 1 and §6.5 still credited Probe 10 with proving
  the recovery (now Probe 13, Probe 10 corroborating); Probe 5b was labelled the
  "exact `017-S` shape" when it is a three-member **representative** fixture
  covering both archived member types (relabelled, with the per-member
  generalisation argument stated); and §13's review scope carried superseded
  rev-4 framing (TOCTOU "close" → "narrow", probe count corrected). No P0/P1
  outstanding. Independent multi-model adversarial review scope is §13.

- **Rev 4.1 → independent adversarial review: `MUST_REPLAN`.** Ten findings,
  all addressed in rev 5 (§12.1).

### 12.1 Rev 5 — what changed and why

| # | Adversarial finding | Rev 5 disposition |
|---|---|---|
| 1 | Self-hosting surface omitted `_ship.agent.md`, so a fresh reload would not recognize `TASK_ONLY_FINALIZE` | **Fixed.** Surface is now **4 files** (§4). The load-bearing clause (`_ship.agent.md` L794) is quoted and its deadlock consequence stated. AC-9 extends positive verification to all three instruction files. Task/feature AC and file inventory updated. |
| 2 | 29 vague precision points; `017-S` pre-archived support rested on a *representative*, not exact, fixture | **Fixed.** §6.2.0 states **five explicit topology assertions (T1–T5)**; §7.1 reduced to a **12-row** table-driven set. **Probe 15** executes the **exact** 12-member `017-S` shape: `INVARIANCE_FAILURES=0`. Support granted on exact evidence; the contingent pre-claim manifest normalization is **not** required. |
| 3 | Engine identity not durable; probes in `%TEMP%` and deleted; digest binding only advisory | **Fixed.** SHA-256 computed and recorded (§6.3, evidence README). Mismatch is **HALT + probe refresh**; advisory fallback **withdrawn**. All probes re-executed in-workspace with script + transcript **committed**; `%TEMP%` probes 5/5b/10/12/13 **withdrawn as evidence**. No raw DB/WAL committed. |
| 4 | No executed end-to-end dependency lifecycle proof | **Fixed. Probe 14.** Two shipments, `blocks` edge matching `017-S -> 021-S`: successor suppressed while predecessor `queued` **and** `active`; predecessor closes to `archived` + `archived_status: shipped` + `commit`; after `sync`, successor eligible. All verdicts `status: ok`. Archived-predecessor handling clarified: the **edge persists**, eligibility derives from **status**. |
| 5 | TOCTOU claim under-specified; outside relationship holders unlocked | **Fixed.** §6.2.3 relies on the enforced single-agent/single-worktree/P-001 precondition, locks the **entire relation closure including existing outside inbound-edge/link/membership holders**, re-scans and re-hashes immediately before the call, and **HALTs** if the precondition is violated. **No absolute-prevention claim.** |
| 6 | Recovery path unnamed; class discipline absent | **Fixed.** §6.5.1 names **`.autoharness/backups/stage-recovery/`** (gitignored, workspace-contained, outside the compared inventory). §6.5.2 fixes three state classes. **Approval precedes quarantine AND restore.** Append-only never rewound; DB/WAL/SHM never byte-restored. **Probe 16** executes it end-to-end. |
| 7 | "never cascade"/"sole exception" occurrences not enumerated; failure order unclear | **Fixed.** §6.1 carries a **14-clause inventory** with file + line. Order is CASCADE → `TASK_ONLY_FINALIZE` → `SAFE_CLOSE`/HALT. §6.5: pre-call failure **may** fall through; **any** errored/post-call failure is recovery + HALT, **never** fallback. Pre/post reconcile retained. |
| 8 | P-007 order relative to P1–P10 unspecified; merge SHA verification thin; contingency incomplete | **Fixed.** §8.2 fixes the order: **P1–P10 evaluate on the raw post-call state first**; P-007 restoration runs **only after** they pass, then re-verification; on any post-guard failure the recovery procedure supersedes it so deletions are preserved as evidence. Merge SHA verified twice. §8.3 completes the contingency. |
| 9 | Stale rev references; lineage/marker/inbound-edge/contingency gaps | **Fixed.** Rev pointers updated throughout; G13 explicitly scopes **inbound** edges and §8's eligibility argument is corrected accordingly; §10.1 gives the >2h HALT-and-return contingency. |
| 10 | `## Plan Hardening` not literal/current | **Fixed.** Section is literal, renumbered to current probes, and gains item 10. |

**Additional rev-5 finding, discovered by testing rather than review:** the
**Ship PR-lifecycle-only route is not executable** on this harness (Probe 17).
Recorded as a **BLOCKER** in §11 rather than asserted as feasible.

### 12.2 Rev 5 → internal plan review: **ADVISORY with an open remediation queue**

An independent internal reviewer examined all 14 sections, the Plan Hardening
block, and all four committed transcripts, and **cross-checked every §6.1 line
citation against the installed contract files** (all 14 confirmed accurate).

**Closed in place during this cycle** — 1×P0, 4×P1, 3×P2, 3×P3:

| Sev | Finding | Resolution |
|---|---|---|
| **P0** | "byte-for-byte unchanged" contradicted §6.1's mandated edits to lines 436/443, which are *inside* the exception block | **§4.1** defines the **preserved region** precisely (preconditions 1–5, item 7, SUPERSESSION NOTE); AC-1 and harness row 2 restated against it |
| **P1** | `G10` silently vanished, leaving "G0–G13" unsatisfiable | explicit **G10 — WITHDRAWN** stub; identifier retired, not reused |
| **P1** | P4 would **fail on every successful close** — append-only logs are in the baseline scope but always grow | §6.2.2 exclusion 3 adds the whole append-only class; P4 scoped to the **mutable current-state class only** |
| **P1** | §6.5.3 quarantine was unbounded — would relocate unrelated live queue records during recovery | quarantine **bounded** to baseline scope + `required_ids` destinations; locks excluded; out-of-scope paths **reported, never moved** |
| **P1** | §7.3 claimed "only committed probes" while granting rows on uncommitted ones; §3's "no conclusion rests on a deleted sandbox" was false | §3 gains the **evidence-degraded table**; the false sentence is **withdrawn**; §7.3 rows marked *committed* / *degraded* |
| **P2** | Plan Hardening preamble claimed every cited probe was committed | corrected to name 14–17 as committed and 1/3/4/9 as degraded |
| **P2** | sizing said "14-row inventory"; the table has 11 rows / 14 sites | reconciled in §6.1 and §10.1 |
| **P2** | §7.3 claimed G9 excludes the destination conflict, which G9 did not test | **G9 conjunct 2** added: destination path must be absent and not a directory/unparseable file |
| **P3** | `§6.5 step 4` pointer resolved to "restore", not rehydration | → **§6.5.3 step 5** |
| **P3** | header Status showed an unblocked plan | now reads **BLOCKED** with the §11 pointer |
| **P3** | §8 had no `§8.1` | step table now carries an explicit **§8.1** heading |

**OPEN AT REV 5 — all eight CLOSED in rev 6 (§12.3).** Recorded here as the
rev-5 record; see §12.3 for each item's disposition. None was carried forward.

| # | Sev | Open item |
|---|---|---|
| O-1 | P2 | §6.1's inventory is likely **incomplete**: reviewer identified two further sites — `shipment-reconcile/SKILL.md` Safe-Close Step 0's **binary** close-path selector, which must become ternary, and `_ship.agent.md`'s safe-close summary still prescribing `move --status shipped`, which §2 says the engine **refuses** with exit 9. |
| O-2 | P2 | **AC-7's** Safe-Close step-8 cross-reference site is not in the §6.1 table and not in the §10.1 budget. |
| O-3 | P2 | The **committed clause inventory** artifact has no named path and no home in the 4-file surface. |
| O-4 | P2 | §7.2 / AC-11 **overstate harness coverage**: the 12-row table verifies none of G0–G13, the TOCTOU protocol, the baseline scope, or the §8 ordering, so AC-3/4/14/15/19 have **no** automated verification. |
| O-5 | P2 | §8 omits two steps its own evidence requires: an explicit **claim** step (G12 needs `active`) and an explicit **`backlogit sync`** step (Probe 14 makes successor eligibility conditional on it). |
| O-6 | P2 | The **post-archival closure route** was never probed against the same gate Probe 17 falsified. |
| O-7 | P2 | Probes ran against a default `backlogit init` workspace; the live workspace has custom `header-def.yaml`, `hooks.yaml`, `.locks/`. **Configuration is unbound.** |
| O-8 | P3 | Residual precision: T4 liveness-by-location wording; §7.1's row-11 matcher breadth; the ≤4 Go helpers vs the "1 of 1" red count. |

### 12.3 Rev 6 — adversarial remediation cycle 1 of 3 (against HEAD `1f55e6a`)

Independent adversarial re-review of rev 5 returned **MUST_REMEDIATE**, architecture
retained. **All eight open items are CLOSED in this revision** — none is carried
forward, and no known edit site is left labelled "open".

| # | Rev-6 disposition | Where |
|---|---|---|
| **O-1** | **CLOSED.** Both sites enumerated: the binary selector is **B10** (`SKILL.md` 505–512), the `move --status shipped` summary is **C2** (`_ship.agent.md` 789–791). Inventory closed at **28 sites**. | §6.1-B, §6.1-C |
| **O-2** | **CLOSED.** Safe-Close step 8 is **B11** (lines 591–605), in the table and in the budget. AC-7 rewritten to cite it by inventory ID. | §6.1-B, AC-7, §10.1 |
| **O-3** | **CLOSED.** The inventory has a name (**Inventory Table §6.1-A/B/C/D**) and a committed path (**§6.1-E**), recorded as *evidence*, so the 4-file surface is unchanged. | §6.1-E, AC-13 |
| **O-4** | **CLOSED.** §7.2 rewritten to state exactly what the one Go test proves (presence / parity / selector **text** ordering / required tokens) and exactly what it does not (no G0–G13, no P1–P10, no recovery, no `ShipShipment`). AC-3/4/11/14/15/19 restated with explicit verification methods. | §7.2, §9 |
| **O-5** | **CLOSED.** §8.1 gains **step 2 (claim)** and **step 13 (`backlogit sync`)**, both cited to the evidence that requires them. | §8.1 |
| **O-6** | **CLOSED by measurement.** **Probe 20** drove a live-config fixture to the post-archive state: **all 4 routes fail closed** (`POST_ARCHIVE_ROUTES_NONZERO_EXIT=4`). But the installed Ship contract does **not** run that gate before the closure PR (L741–744), so §8 is **not** deadlocked. Recorded as a normative prohibition on steps 15–19 plus a tripwire, not as "proven safe". | §8.1.1, §10.3 |
| **O-7** | **CLOSED.** Probes **18/19/20** are seeded from the **live** `.backlogit` control files with each seed file's SHA-256 recorded, and are **digest-gated in-script**. The delta is confirmed material (`SEED_STATUS_ENUM_HAS_SHIPPED=False`; hooks `validate_transition` + `pre_task_completion_gate` both present). | §3.0, §3 |
| **O-8** | **CLOSED.** T4 now reads liveness from frontmatter `status` with directory used only as a location/integrity cross-check — and Probe 18 **measured** why (live `registry.yaml` routes `done` to `archive/`). §7.1 row 11's matcher is narrowed to the exclusivity claim and explicitly spares the preserved block's "only when" text. Helpers explicitly excluded from the "1 of 1" red count. | §6.2.0, §7.1, AC-8 |

**Additional rev-6 findings, discovered by remediation rather than review:**

| Sev | Finding | Disposition |
|---|---|---|
| **P0** | **The §10.1 budget does not close.** The closed 28-site inventory prices at **~2.5–2.7 h**, and §10.2/§5/AC-17 establish the unit **cannot be split**. | **MUST_REPLAN** (§10.1). Topology deliberately left unchanged. |
| **P1** | Rev 5's §8 ordered the closure PR **before** operational closure and P-020, contradicting installed Ship **L741** ("after all closure work is committed, push … and create a PR"). | **Fixed** — §8.1 reordered; closure work is authored on the closure branch at step 15, PR opens at step 16. |
| **P1** | Rev 5 claimed the §6.5.3 recovery procedure was **proven** by Probe 16, which diverged from it in five ways and never ran the approval-withheld branch. | **Fixed** — claim withdrawn; **Probe 19** implements and executes the procedure on **both** branches. |
| **P1** | Rev 5 relied on an **unconditional** standing merge preauthorization for PR #54. | **Fixed** — §14 gate 7 now requires **fresh** approval after current-HEAD review/CI/P-018, with the bounded dark-mode exception named but **not assumed**. |
| **P2** | The committed `probe16-bounded-recovery.ps1` resolves its repo root by `..\..\..` from a 4-deep directory, landing on `docs\` — it does not reproduce as written. | **Recorded** in §3/§3.0; probes 18–20 resolve the root via `git rev-parse --show-toplevel`. |
| **P2** | No probe ever exercised the **MCP** surface, yet the plan spoke of tool operations generically. | **Fixed** — §3.0 states **CLI-only** authorization and explicitly disclaims MCP equivalence. |
| **P2** | TOCTOU claimed more than detection. | **Fixed** — §6.2.3 claims **detection, not prevention**; preconditions expanded to four; the inbound `017-S -> 021-S` record is explicitly in baseline/lock/re-hash and **proven byte-identical** by Probe 18. |

**Deferred, deliberately not widened:** the **P-017 blocked-status mismatch** remains
an unrelated finding outside this unit's scope and is **not** addressed here.

**Verdict: MUST_REPLAN** — on the §10.1 budget only. No P0 or P1 remains on
*correctness*; the blocking defect is that the honestly-measured work does not fit
the 2-hour rule in a unit that structurally cannot be split.

## 13. Scope for independent adversarial plan review

Reviewers should focus on, in priority order:

1. **The §10.1 budget verdict — highest priority.** Is the 28-site closed
   inventory in §6.1 genuinely complete against the installed files at
   `1f55e6a`, is the MUST/CONSISTENCY split defensible, and is
   **MUST_REPLAN** the correct verdict — or is there a scoping the plan has
   missed that closes the budget without breaking T2/§10.2?
2. **Inventory completeness.** Re-run the §6.1 scan independently. A **MUST**
   site present in the files but absent from Inventory Tables §6.1-A/B/C
   invalidates both the budget and AC-13.
3. **Guard sufficiency.** Can any topology satisfy **T1–T5** and G0–G13 yet cause
   `backlogit shipment ship` to archive, return, or reparent an artifact outside
   `allowed_ids`? Committed probes cover the shapes in §7.3.
4. **Detection completeness.** Can any violation escape P1–P10 given detection is
   post-destructive — and is §8.2's "detect before repair" ordering correct?
5. **Recovery soundness (now executed, both branches).** Is §6.5 sufficient given
   **Probe 19**'s evidence — enumerated-set-only restore, approval-before-any-
   destructive-step, byte-prefix append-only preservation, out-of-bounds paths
   reported-not-moved, sync-only rehydration, byte/path/dependency equivalence —
   and is the residual scoping (bounded set, not whole directory) correct?
6. **TOCTOU boundary.** §6.2.3 now claims **detection, not prevention**. Is the
   relation closure complete in both directions, is the four-part execution-mode
   precondition legitimate rather than unfalsifiable, and does Probe 18's
   inbound-edge byte-invariance result actually discharge the `017-S -> 021-S`
   obligation?
7. **§8 ordering and §8.1.1.** Is the corrected order (verify on main → **claim**
   → implement/merge → verify SHA → fresh main → reload+verify all three files
   and the merge commit → pre-mode → close → raw P1–P10 → P-007 → post-mode →
   **sync** → verify → **operational closure + P-020 on the branch** → push/open
   PR → current-HEAD review/CI/P-018 → **fresh** approval → merge → verify →
   route `017-S`) correct against the installed Ship agent? Is the §8.1.1
   disposition of O-6 — *gate blocks post-archive, but the closure-PR path does
   not invoke it, therefore prohibit invoking it and set a tripwire* — sound, or
   should §8 be gated harder?
8. **Harness honesty.** Does §7.2 now describe the single Go table-driven test
   accurately, and do AC-3/4/11/14/15/19 correctly separate *review/diff* from
   *runtime fixture probe* from *contract-text token assertion*?
9. **Evidence integrity and authorization surface.** Is §3.0's fidelity table
   right? Is the **CLI-only** authorization statement and the explicit refusal to
   claim MCP equivalence adequate, given the registry advertises MCP tools?
10. **§11/§14 actor and gates.** Is the conclusion that no agent can carry PR #54
    still correct, and is the **fresh-approval** requirement (gate 7) correctly
    stated, including the bounded dark-mode carve-out that Stage explicitly does
    **not** assume exists?

## 14. Sequencing — one artifacts-only Stage PR

### 14.0 REV 7 UPDATE — PR #54 now carries all THREE shipments atomically

The adopted single-PR sequencing below is **retained and extended**. PR #54 now carries
the **entire three-shipment bootstrap** and **every dependency edge** in **one atomic
merge**:

| Carried by PR #54 | Detail |
|---|---|
| Policy-gap backlog records | already on the branch (unchanged) |
| **A** | `022-F`, `022.001-T` (re-scoped), `021-S` re-shaped to `[022-F, 022.001-T]` |
| **B** | `023-F`, `023.001-T`, `022-S` = `[023-F, 023.001-T]` |
| **C** | `024-F`, `024.001-T`, `023-S` = `[024.001-T]` (task-only) |
| **Dependency chain** | `022-S -> 021-S`, `023-S -> 022-S`, `017-S -> 023-S`; stale `017-S -> 021-S` **removed** |
| Planning artifacts | the bootstrap deliberation + plans A/B/C + this rev-7 role change |

**Why atomic is still required, and now more so.** The rev-6 rationale — the
`017-S` edge must land in the same merge as `017-S`'s own records so no eligibility
window opens — applies with **three** edges instead of one. Landing A's records without
C's would leave `017-S` with a **dangling or absent** blocker. The edge surgery was
performed in the safe order (**add `017-S -> 023-S` first, remove `017-S -> 021-S`
second**), so at no point in the branch history is `017-S` unblocked.

**Gates 1–9 below apply unchanged.** Gate 3 (artifact-only diff) is satisfied: the diff is
limited to `docs/**` and `.backlogit/**`, with **zero** source, test, skill, agent or
policy implementation files.

**Actor: the OPERATOR, unchanged.** §11 and Probe 17 still hold — no agent can push or
merge PR #54. **No push is performed by Stage.** Output is local commits on
`chore/stage-pipeline-policy-gap`.

### 14.0.1 Post-merge routing (normative, rev 7)

After PR #54 merges and the records are verified present on `main`, the Orchestrator
routes **`021-S` (A) only**. Then, strictly sequentially, each under **P-001** (one
release unit in flight) and **P-016** (one branch/worktree):

```text
PR #54 merge
  -> route 021-S (A)  -> implement -> merge -> S1 checkpoint+end -> S2 closes via CASCADE
  -> route 022-S (B)  -> implement -> merge -> close via a1 + CASCADE
  -> route 023-S (C)  -> implement -> merge -> close via TASK_ONLY_FINALIZE (self-proof)
  -> Stage disposes of 024-F (plan C §8.3)
  -> route 017-S
```

**Never more than one active Ship shipment or PR at a time.** Each shipment completes
**P-020** and full closure requirements before the next is routed. `017-S` remains
`queued` and **unclaimed** throughout.

The rejected two-PR sequence is **withdrawn**. Its defect: landing the policy-gap
records first would leave `017-S` queued on `main` with **no** dependency edge
suppressing it, creating a window in which Ship could legitimately claim it before
the prerequisite existed.

**Adopted sequencing.** A single artifacts-only Stage PR — **PR #54**, on
`chore/stage-pipeline-policy-gap` — carries **both**:

- the policy-gap backlog records (already on the branch), and
- this prerequisite's feature, task, shipment and the
  `017-S depends_on <prerequisite shipment>` edge.

Because the dependency edge lands **atomically in the same merge** as `017-S`'s
own records, there is **no** eligibility window: at no point does a reachable
state exist where `017-S` is claimable without its blocker present.

**ID allocation** is branch-local, which is precisely why harvest must happen on
the policy branch: `018-F`, `017-S` and stash `A10EF3D0` exist **there**, so the
dependency edge and the stash archival are writable. The rev-3 blocker is
therefore dissolved rather than deferred.

**Merge gates for PR #54** (all required, **no admin fallback**):

| # | Gate | Detail |
|---|---|---|
| 1 | **No shipment claim** | PR #54 neither claims nor executes `021-S` or `017-S`. It carries **artifacts only**. |
| 2 | **Fast-forward push only** | Push the **existing** policy branch `chore/stage-pipeline-policy-gap`. No force, no rebase, no reset, no new branch. |
| 3 | **Artifact-only diff** | Zero source, test, skill, agent, or policy **implementation** files. Diff limited to `docs/**` and `.backlogit/**`. |
| 4 | **Exact-HEAD local adversarial READY** | The readiness record must cover the **exact** merge HEAD, not an ancestor. |
| 5 | **CI green** | Full pipeline green on that HEAD. |
| 6 | **P-018 Copilot review** | Completed against the **current HEAD**, with **zero** unresolved Copilot threads. |
| 7 | **Fresh merge approval** | **Rev 6 correction.** Approval MUST be obtained **after** gates 4–6 complete on the **current** HEAD. A pre-existing/standing merge preauthorization does **NOT** satisfy this gate. The only exception is a **currently valid, bounded dark-mode activation record that explicitly preauthorizes this exact PR number and this exact head SHA** — Stage has **not** verified that such a record exists and does **not** assume one. Absent that exact record, approval is fresh-or-nothing. Ship's own contract says the same thing for closure PRs: `_ship.agent.md` **L723**, *"the prior main PR approval does not transfer."* |
| 8 | **P-009 merge commit** | No squash, no rebase. |
| 9 | **Post-merge verification** | Two-parent merge commit confirmed; merged **artifact set** present on `main`; the **`017-S -> 021-S` dependency edge** verified present on `main`. |

**Actor for every gate above: the OPERATOR.** Per §11 and Probe 17, Ship cannot
execute this PR — its Step 5 topology gate fails closed without an active
shipment, and claiming one is forbidden here. This supersedes rev 4.1's
"Ship, in PR-lifecycle-only mode" designation.

**Rev 6 — the actor conclusion is retained unchanged and re-affirmed.** No agent
can push or merge PR #54 under the currently installed topology; the operator is
the only possible actor. **No push is performed at this stage**, by any actor.
Stage's output remains local commits on `chore/stage-pipeline-policy-gap`.

**Gate ordering is normative (rev 6).** Gates 4 → 5 → 6 → 7 → 8 run in that
order. Approval (gate 7) is the **last** gate before the merge mechanic (gate 8)
precisely because it must attest to a HEAD that has already passed local
adversarial review, CI, and P-018. Obtaining approval earlier and then pushing
further commits **invalidates** it and re-opens gate 7.

**Review-comment routing (unchanged and still binding).** If any review comment
requires a **planning or contract mutation**, the PR actor **pauses and routes to
Stage**. Stage makes the edit; no non-Stage actor modifies planning content. This
holds for the operator as PR actor exactly as it would have for Ship: it is a
Role-Boundary rule about *who authors planning artifacts*, not about who pushes.

### 14.1 Downstream sequencing gate (normative)

**No source implementation and no shipment claim may begin until BOTH hold:**

1. **PR #54 is merged**, and
2. **`021-S` is present on `main`** (verified by reading the merged
   `.backlogit/queue/021-S.md` on `main`, not on the branch).

Only then does the Orchestrator route **`021-S`** — and only `021-S`. Probe 14
confirms `017-S` remains suppressed while `021-S` is `queued` or `active`, and
becomes eligible only once `021-S` is `archived` with
`archived_status: shipped` and the index is synced. The live workspace already
reports exactly this: `ready_set: ["021-S"]`, `downstream_dependents: {"021-S":
["017-S"]}`, `cycle_detected: false`.

Planning artifacts from the sibling branch `chore/stage-task-only-shipment-finalization`
(commit `2701906`) were brought across by a clean `git cherry-pick`, preserving
both branch histories with no rebase, reset, or force operation. Provenance is
recorded in the session memory artifact.

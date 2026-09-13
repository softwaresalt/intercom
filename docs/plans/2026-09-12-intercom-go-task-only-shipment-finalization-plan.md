# Plan — Verified Task-Only Shipment Finalization

> **Requires plan hardening**: **yes — applied in rev 2, extended in rev 3, re-hardened in rev 4, rev 4.1 and rev 5** (see `## Plan Hardening`).

- **Date:** 2026-09-12
- **Revision:** 5 (rev 1 → **FAIL**; rev 2 → **FAIL**; rev 3 → **ADVISORY**; rev 3 → independent adversarial **MUST_REPLAN**; rev 4 → internal review **FAIL**; rev 4.1 → internal re-review **ADVISORY**; rev 4.1 → independent adversarial **MUST_REPLAN**; see §12)
- **Source deliberation:** `docs/decisions/2026-09-12-intercom-go-task-only-shipment-finalization-deliberation.md`
- **Stash origin:** `A10EF3D0` (`kind: deliberation`, `priority: critical`)
- **Branch:** `chore/stage-pipeline-policy-gap` (planning artifacts carried over by cherry-pick; see §14)
- **Engine identity of record (binding):** `backlogit 1.10.1-0.20260823032255-b07729386a31+dirty`, **SHA-256 `1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98`** — the digest, not the version string, is the binding identity (§6.3)
- **Probe evidence (durable, in-repo):** `docs/plans/evidence/2026-09-12-task-only-shipment-finalization/`
- **Harvested backlog:** feature `022-F`, task `022.001-T`, shipment `021-S` (task-only manifest); `017-S depends_on 021-S --type blocks`
- **Status:** planning complete; **BLOCKED** — PR #54 has no agent-executable route and requires the **operator** as PR actor (§11); harvest **COMPLETE** on the policy branch (§14)

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
| **16** | **executed, committed. Injected partial failure**: archive destination of the shipment record occupied by a directory, then ship | **exit 1** *after* partial mutation — task archived, shipment left in the queue declaring `status: shipped` (**torn**). Bounded recovery then executed: approval gate → quarantine (move, never delete) → enumerated restore → `backlogit sync` → `MUTABLE_EQUIVALENCE_FAILURES=0`, `RESIDUAL_UNEXPECTED_PATHS=0`, `APPEND_ONLY_REWIND_VIOLATIONS=0`, engine view restored, quarantine preserved |
| **17** | **executed, committed.** Ship Step 5 `pipeline-topology` gate in a PR-lifecycle-only (unclaimed) session, all four routes | **Every route fails closed** — exit 2 `agent mode requires --shipment`, exit 1 `LIFECYCLE_NO_ACTIVE_SHIPMENT`, exit 2 `--phase lifecycle requires --shipment … in any mode`, exit 2 ambient. See §11 |

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
| §2 root cause (`move --status shipped` → exit 9) | Probe 1 | degraded |

An earlier rev-5 draft asserted "No conclusion in this plan rests on a deleted
sandbox." **That was false and is WITHDRAWN.** The accurate statement is: no
conclusion rests on a *withdrawn* probe, and every conclusion resting on a
*retained-but-uncommitted* probe is tabulated above. Re-executing and committing
probes 1/2/3/4/6/7/9/11 is carried as a named residual (§10.3).

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
  recovery restores it** (Probe 16). The engine applied mutations and *then*
  failed, leaving the shipment record in the queue declaring `status: shipped` —
  a state the engine itself refuses to create via `move` (Probe 1, exit 9).
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

**Normative-clause reconciliation (mandatory, AC-13) — enumerated, not a sweep
instruction.** Rev 4.1 told the implementer to "locate and rewrite" exclusivity
clauses. Rev 5 **enumerates them**, because an unenumerated sweep is neither
falsifiable nor sizable. The inventory below was produced by scanning the three
surface files in this workspace at `e06b4d8`; each entry must be reconciled to
name **both** exceptions in the order CASCADE → `TASK_ONLY_FINALIZE` →
`SAFE_CLOSE`/HALT.

| File | Line | Clause (abbrev.) | Why it must change |
|---|---|---|---|
| `workflow-policies.md` | 420 | "MUST NOT call the cascade `backlogit_ship_shipment` for closure" | states the prohibition; must name both exceptions |
| `workflow-policies.md` | 426 | Postcondition: "the cascade `backlogit_ship_shipment` was never called" | unreachable postcondition once a 2nd exception exists |
| `workflow-policies.md` | 436 | "**VERIFIED FULLY-COVERED-ROOT EXCEPTION** … remains the DEFAULT" | must acknowledge a second, narrower exception |
| `workflow-policies.md` | 443 | item 6 "IS permitted … in place of the single-artifact safe-close" | must scope to the CASCADE branch only |
| `shipment-reconcile/SKILL.md` | 3 | frontmatter: "**except for** the narrow, machine-verified P-015 fully-covered-root case" | literal sole-exception phrasing |
| `shipment-reconcile/SKILL.md` | 10, 356, 1015 | "run `mode: safe-close` **in place of** the destructive cascade" | must admit the third path |
| `shipment-reconcile/SKILL.md` | 232 | "It **never calls** the cascade" | absolute; must be scoped |
| `shipment-reconcile/SKILL.md` | 324, 998 | "Do NOT call `backlogit_ship_shipment`" | must carry the classification carve-out |
| `_ship.agent.md` | **794** | "**Do NOT call** `backlogit shipment ship` … **unless** the P-015 **VERIFIED FULLY-COVERED-ROOT EXCEPTION** applies" | **load-bearing** — this is the clause a reloaded Ship reads; unamended, Ship refuses the merged path (§4) |
| `_ship.agent.md` | 803 | "safe-close remains the default" | single-exception framing |
| `_ship.agent.md` | 822 | "in place of the safe-close sequence above" | single-exception framing |

**Total: 11 inventory rows enumerating 14 clause sites across 3 files.** The
implementing task re-runs the same scan and commits the resulting inventory so the
reconciliation is falsifiable and re-runnable; a clause found by the re-scan that
is absent from this table is added, not silently skipped. **§12.2 carries known
candidate additions** already identified by internal review.

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
| **T4** | **Exactly one** parent feature, **root** (`parent_id` absent), **live** in the queue directory, and **outside** the manifest. | non-root parent; parent inside the manifest; multiple parents; archived/missing parent |
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
suffix (`-ST` ends in `T`). Liveness is by the frontmatter `status` field —
**never** by queue-vs-archive location.

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
     directory or an unparseable file.** Probe 16's injected failure was exactly
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

Exclusion 3 is **load-bearing, not cosmetic.** Probe 16's transcript shows
per-artifact logs grow on *every* operation (`001-S.jsonl` 341→1301,
`001.001-T.jsonl` 614→1031). If the append-only class were inside the mutation
comparison, **P4 would fail on every successful close** — the happy path would
route straight to §6.5(B) bounded recovery and HALT, and §6.5.2 forbids ever
restoring those files, so recovery could never re-establish equivalence either.
The path would be permanently unusable.

These are **not** compared for mutation because they are *expected* to change;
they are **validated separately** — the report for well-formedness and expected
content, the lock for correct acquisition and release, the index by successful
rehydration (§6.5.3 step 5), and the append-only class by the
**monotonic-growth rule** of §6.5.2 (length must never decrease).

#### 6.2.3 TOCTOU isolation (rev 5, normative)

**Execution-mode precondition (relied upon, not assumed).** This procedure runs
only under the workspace's enforced **single-agent / single-worktree / P-001**
execution mode: at most one top-level release unit in flight, no parallel
implementation branch or worktree (verified by `git worktree list --porcelain`
at Ship Step 0.5 item 3a, which halts with `WORKTREE_TOPOLOGY_BLOCKED` on any
prohibited or ambiguous worktree). Under that precondition **concurrent creation
of new backlog records by another agent is prohibited**, which is what bounds the
residual below.

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
   full relation closure and recompute the classification.
6. If **any** hash, path, declared status, dependency edge, link, or membership
   differs from the G0 capture, **fail closed**: release locks, make no
   destructive call, fall through to `SAFE_CLOSE`/HALT.
7. Release locks in a guaranteed-cleanup path on every exit, including failure.

**Honest bound (no overclaim).** This does **not** claim absolute prevention. It
claims: every *existing* source-of-truth record in the relation closure is locked
and re-verified immediately before the call, and *new* record creation is
excluded by the execution-mode precondition rather than by a lock. **If that
precondition is violated or a new in-scope record is observed anyway, the
procedure HALTS** — it is caught post-call by P1/P10 and routed to §6.5(B).

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
  Probe 16 shows a torn record is a real reachable state, not a theoretical one.
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
file set, proven by **Probe 16** (a real errored/partial invocation). No claim of
atomicity is made anywhere in this plan.

#### 6.5.1 Recovery path (named, gitignored, workspace-contained)

The recovery root is **`.autoharness/backups/stage-recovery/`**, with per-run
subdirectories `snapshot/`, `quarantine/` and `inventory.json`.

It satisfies all four required properties, verified in Probe 16
(`RECOVERY_PATH_GITIGNORED=True`):

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
| **Append-only** | `logs/*.jsonl`, `hooks_queue.jsonl`, `stash.jsonl`, telemetry | **NEVER rewound, NEVER deleted, NEVER truncated.** Receives an appended **recovery event** referencing the run |
| **Disposable cache** | `backlogit.db`, `-wal`, `-shm` | **NEVER byte-restored.** Rehydrated by the official `backlogit sync` |

Probe 16 measured this: `APPEND_ONLY_REWIND_VIOLATIONS=0` with per-artifact logs
growing across the failure and the recovery (341→1301 and 614→1031 bytes), a
recovery event appended, and the engine's view restored from markdown alone with
no DB byte restoration.

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
   - **Why this bound is mandatory.** The reference implementation in
     `probe16-bounded-recovery.ps1` enumerates whole `queue/` + `archive/`
     directories and moves everything absent from its inventory; in the probe's
     3-record sandbox that was harmless (it relocated one engine lock file), but
     against the **live** `.backlogit/queue/`, which holds many unrelated records
     (`017-S`, `018-F`, `019-F`, …), the unbounded reading would relocate
     **unrelated live backlog artifacts out of the queue during recovery** —
     collateral damage strictly worse than the failure being recovered. The probe
     script is evidence for the *procedure*, not a normative implementation; this
     clause governs.
4. **Restore only the enumerated mutable files** from the snapshot, by inventory.
   Never perform a blanket restore of the backlog directory. Never touch the
   append-only or disposable classes.
5. **Rehydrate the disposable index** by running the official `backlogit sync`.
6. **Verify equivalence** against the inventory: every enumerated mutable file
   present with a matching hash; **no unexpected residual paths**; and the
   **dependency, link and membership edges** of the baseline scope unchanged.
   Probe 16 records these as `MUTABLE_EQUIVALENCE_FAILURES=0` and
   `RESIDUAL_UNEXPECTED_PATHS=0`.
7. **Append a recovery event** to the append-only stream — never a rewind.
8. **HALT** with a named token and a **P-005** event, **preserving** the
   quarantine directory and the full diagnostic evidence for operator review.

**Probe obligation — discharged.** **Probe 16** injected a real destination
conflict, observed a **non-zero exit after partial mutation** leaving a torn
record (shipment in `queue/` declaring `status: shipped` — a state the engine
refuses to create directly), and executed this exact procedure to a verified
equivalence PASS with the engine view restored and the quarantine preserved. Its
script and transcript are committed under
`docs/plans/evidence/2026-09-12-task-only-shipment-finalization/`.

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
| 2 | `workflow-policies.md` | the existing fully-covered-root exception paragraph is present and **unmodified** |
| 3 | `workflow-policies.md` | Amendment Log has a new row naming `TASK_ONLY_FINALIZE` |
| 4 | `shipment-reconcile/SKILL.md` | classifier order `CASCADE` → `TASK_ONLY_FINALIZE` → `SAFE_CLOSE`/HALT |
| 5 | `shipment-reconcile/SKILL.md` | the five topology assertions **T1–T5** are stated |
| 6 | `shipment-reconcile/SKILL.md` | pre-mode and post-mode reconciliation retained (no bypass) |
| 7 | `shipment-reconcile/SKILL.md` | failure split: pre-call → fall-through; post-call/errored → recovery + HALT, **no fallback** |
| 8 | `shipment-reconcile/SKILL.md` | recovery names the three state classes and disclaims atomic rollback |
| 9 | `_ship.agent.md` | the Step 6 `Do NOT call … unless` clause names **both** exceptions |
| 10 | `_ship.agent.md` | `TASK_ONLY_FINALIZE` is recognized by name, so a fresh reload authorizes it |
| 11 | all three | no surviving "sole exception" / "only exception" phrasing about the cascade |
| 12 | `_ship.agent.md` | Ship Role Boundary is **unchanged** (no new privilege) |

Rows 2 and 12 are **negative** assertions: they fail if the change widened
something it was not authorized to widen.

### 7.2 What the harness does and does **not** verify

The Go harness reads installed markdown; it verifies **contract text**, not
runtime behavior. It cannot execute the engine, and there are deliberately **no**
executable negative cases in this unit.

Engine behavior is verified separately by the **committed characterization
probes** (§3). The plan claims only that:

- engine behavior is established by §3's executed probes, for the engine identity
  of record (§6.3), with script and transcript committed;
- contract text is established by the harness.

Guards G0/G6/G7/G8/G11/G13, §6.2.3 isolation, §6.5 recovery, and the §8 ordering
are **behavioral obligations on Ship**; the harness verifies they are *specified*,
not that they *executed*. Runtime enforcement (an executable classifier + negative
fixtures) is explicitly **deferred** to a separate release unit.

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
| **errored/partial invocation → bounded recovery** | **16** *(committed)* | recovery path **proven** |
| engine-identity mismatch | — | **no** — excluded by §6.3 digest binding (HALT) |

Multi-**live**-member manifests have **no** probe and are therefore
**unsupported** (T2).

## 8. Self-hosting closure sequence

### 8.1 The ordered sequence

The prerequisite shipment closes itself using the contract it merges. The full
ordered sequence, with actor, is:

| # | Actor | Step |
|---|---|---|
| 1 | **Ship** | Implement the sole task on an implementation branch; open the implementation PR. |
| 2 | **Ship** | Merge the implementation PR (P-009 merge commit). |
| 3 | **Ship** | **Verify the merge SHA** (two-parent merge commit confirmed) and record it. |
| 4 | **Ship** | Check out a **fresh `main`** and cut the closure branch from it. |
| 5 | **Ship** | **Reload** the merged P-015, `shipment-reconcile` skill **and `_ship.agent.md`**, and **positively verify** the `TASK_ONLY_FINALIZE` tokens and the merge commit are present in the working checkout. |
| 6 | **Ship** | Run `shipment-reconcile` **pre-mode**. |
| 7 | **Ship** | Classify and close via `TASK_ONLY_FINALIZE` (§6.2 → §6.2.3 → §6.3 → §6.4), passing `--sha <merge_sha>` from step 3. |
| 8 | **Ship** | **P-007 handling, in the explicit order fixed by §8.2 below.** |
| 9 | **Ship** | Run `shipment-reconcile` **post-mode**. |
| 10 | **Ship** | Open and merge the closure PR. |
| 11 | **Ship** | Operational closure. |
| 12 | **Ship** | **P-020** compaction. |
| 13 | — | **Only then** does the successor `017-S` become eligible. |

**Step 5 must reload `_ship.agent.md` too.** This is why the file is surface 3
(§4): the clause a reloaded Ship actually reads (`_ship.agent.md` line 794) names
only the fully-covered-root exception. If it is not amended, step 5's positive
verification **fails** — and if the verification were skipped, step 7 would be
executed by an agent whose own contract forbids the call. Either way the sequence
deadlocks. Step 5 therefore verifies the token in **all three** instruction files.

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

**Merge-metadata preservation and verification.** The merge SHA from step 3 is
passed at §6.3 and **verified twice**: by **P7** (the archived shipment record
reports `archived_status: shipped` **and** retains `commit: <merge_sha>`) and
again at step 4 above after any P-007 restoration. Probes 14 and 15 both confirm
the archived shipment record carries `status: archived`, `archived_status:
shipped`, and the `commit` value passed to `--sha`.

**Eligibility at step 7.** The harvested artifacts MUST satisfy:

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

**Non-circular:** the authorization used at step 7 is the *merged* artifact from
step 2, not the in-flight working copy. Step 5 (reload + positive verification
across all three files) is the load-bearing ordering constraint, bound as AC-9.

### 8.3 Contingency — first real self-close fails after the implementation merged

If step 7 fails after step 2 has already merged:

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
- **AC-3** The five topology assertions **T1–T5** and guards G0–G13 are specified (§6.2.0, §6.2.1).
- **AC-4** All of P1–P10 are specified, with `allowed_ids`/`required_ids` defined normatively.
- **AC-5** P1 is justified by reference to G5; P4 requires independent filesystem diffing derived from `required_ids`; P9 post-compares each task's declared status, single location, and `parent_id`; P10 asserts descendant byte-identity.
- **AC-6** §6.5 specifies **two separate** failure paths: (A) pre-call guard failure → existing SAFE_CLOSE/HALT with **no** destructive call and **no** recovery; (B) post-call **or any errored/indeterminate** invocation → bounded recovery + HALT, **never** a fallback close path.
- **AC-7** Safe-Close step 8 carries a fail-closed cross-reference to the new path; step 8's existing text is otherwise unchanged.
- **AC-8** P-004: at H0 `go vet ./...` exits 0, `go test ./...` exits non-zero, the marker `not implemented: task-only finalize contract` is present, red count 1 of 1. At H1 the marker is gone and full `go test ./...` exits 0.
- **AC-9** The task records the §8 step 4–5 fresh-`main` checkout, reload and **positive verification of merged tokens in all three instruction files**, including `_ship.agent.md`, as an execution obligation on Ship.
- **AC-10** The generated function executes unconditionally — no `t.Skip`, build tag, env gate, or selector flag.
- **AC-11** No Ship Role Boundary change, no generic `move shipment → shipped`, no pre-mode bypass, no unrestricted cascade. Asserted **negatively** by harness rows 2 and 12 (§7.1).
- **AC-12** The authorized call passes merge-commit metadata (`--sha`, and `--message`/`--author` where declared); **P7** re-asserts `archived_status: shipped` **and** `commit` retention; §8.2 fixes P-007 ordering **after** P1–P10 with a re-verification pass.
- **AC-13** All three surface files are reconciled against the **enumerated 14-clause inventory in §6.1**. The implementing task re-runs the scan and commits the resulting inventory (clause text + file + line); any clause found by the re-scan but absent from the table is added, never skipped. Mandatory pre/post reconciliation steps are retained unchanged.
- **AC-14** §6.2.2 baseline scope is specified, including **inbound** edge/link/membership holders outside the parent subtree, and the explicit exclusion of expected pre-mode report/lock outputs from mutation comparison.
- **AC-15** §6.2.3 TOCTOU isolation is specified: the single-agent/single-worktree/P-001 execution-mode precondition, bootstrap enumeration of the **relation closure** (including existing outside relationship holders), locks over that closure, re-run under lock, **re-scan and re-hash immediately before the call**, fail-closed on any difference, and an honest statement that new-record creation is excluded by the execution mode rather than by a lock — with HALT if that precondition is violated.
- **AC-16** §6.3 records the engine version **and the SHA-256 digest** of the executable, and **HALTs** on mismatch of either, requiring a probe refresh. No advisory fallback exists.
- **AC-17** The unit contains **exactly one** task and one generated (table-driven) function; no sequential split is permitted.
- **AC-18** The harness is **one table-driven test** over the 12-row assertion table in §7.1, including the two negative rows.
- **AC-19** The committed probe evidence directory `docs/plans/evidence/2026-09-12-task-only-shipment-finalization/` is referenced by the contract as the engine-behavior basis, and the recorded digest matches the engine used at closure.

## 10. Sizing, risks, residuals

**Sizing.** `size: M`, `complexity: high` — recorded on the harvested task as
structured `size` (with `size_source: agent` and a `size_ruleset_version`) plus
`complexity` carried as **labeled prose**, because this workspace's
`.backlogit/header-def.yaml` defines **no** `complexity` field on the `task`
type. That partial degradation is flagged explicitly in the Stage report.

### 10.1 Concrete edit budget (rev 5) — the 2-hour claim, itemized

Rev 4.1 asserted "upper bound of M" and left the estimate contestable. Rev 5
**itemizes** it. The file count (4) is a **heuristic, not a waiver**; the budget
below is the actual claim.

| Work item | Unit | Count | Basis |
|---|---|---|---|
| Clause reconciliations | small pre-located text edits | **14** | enumerated with file + line in §6.1 |
| New P-015 exception block | one additive block | **1** | mirrors the existing exception's shape |
| New `shipment-reconcile` classification section | one additive block | **1** | T1–T5 + guards + failure split |
| Amendment Log row | one table row | **1** | mechanical |
| Committed clause inventory | one generated table | **1** | output of re-running the §6.1 scan |
| Go harness | **one table-driven test function** | **1** | 12-row table (§7.1) |
| Go helpers | small file-read/assert helpers | **≤4** | e.g. `readContract`, `assertContains`, `assertAbsent`, `assertUnchanged` |

**Totals: 14 clause edits + 4 additive blocks/rows + 1 table-driven test function
+ ≤4 helpers. Zero new production files. Zero design work** — every edit location
is pre-identified in this plan, and the required token for each is stated.

Rev 5 **reduced** the budget relative to rev 4.1 in two material ways:

- §7.1 went from **29 prose-precision assertions to a 12-row table** driven by one
  function — the single largest cost reduction;
- §6.1 replaced a "locate and rewrite every equivalent paraphrase" **sweep** (open
  ended, unsizable) with a **closed 11-row inventory enumerating 14 clause sites**
  across the three *contract* files (the fourth surface is the test harness).

Those reductions are what make the budget defensible rather than asserted. The
added `_ship.agent.md` surface costs **3 of the 14** clause edits and no new
block, so the net change from rev 4.1 is a **decrease**.

**Honest limit.** This is a drafting task with pre-located edits, which is why it
fits. If Ship finds the actual work exceeding two hours **mid-execution**, the
required response is **HALT and return the unit to Stage for re-planning** —
never split it (see below), and never rush contract prose. Stage does **not**
claim the estimate is beyond challenge; it claims the estimate is now *itemized*
and therefore falsifiable.

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

## Plan Hardening

*(Literal section title, current as of rev 5. **Probes 14–17** referenced below
are committed with script and transcript under
`docs/plans/evidence/2026-09-12-task-only-shipment-finalization/`. **Probes 1, 3,
4 and 9** are retained rev-4.1 characterizations **without** committed
transcripts — see §3's evidence-degraded table.)*

1. **Destructive-operation adjacency.** The authorized call mutates the backlog
   with no transaction. Mitigated by T1–T5 plus G0–G13, §6.2.3 isolation, P1–P10,
   and the §6.5(B) bounded recovery proven by **Probe 16**.
2. **Silent scope expansion is reachable and confirmed damaging.** Probe 3
   archives omitted descendants; Probes 4 and 9 return omitted live siblings —
   and Probe 9 proves the return **strips `parent_id`**, orphaning the record and
   doing so **across a shipment boundary**. T5/G5 and P1/P10 are correctness
   requirements derived from executed evidence.
3. **Recovery is bounded, not atomic rollback.** Stated plainly in §6.5. No
   DB/WAL/SHM byte restoration is claimed; the index is a disposable cache
   rehydrated by the official `backlogit sync`. **Probe 16 proves a partial
   failure is real**: the engine applied mutations and *then* errored, leaving the
   shipment record declaring `status: shipped` while still in the queue directory
   — a torn state the engine itself refuses to create directly (Probe 1, exit 9).
   Bounded recovery restored byte/path equivalence
   (`MUTABLE_EQUIVALENCE_FAILURES=0`) and the engine's own view, with
   `APPEND_ONLY_REWIND_VIOLATIONS=0`. **Live destructive recovery requires
   explicit operator action-risk approval BEFORE quarantine or restore.**
4. **TOCTOU.** §6.2.3 locks the **entire relation closure** — including existing
   records outside the parent subtree that hold inbound edges, links, or
   shipment membership — re-runs enumeration under lock, and **re-scans and
   re-hashes immediately before the call**. New-record creation is excluded by
   the enforced **single-agent / single-worktree / P-001 execution mode**, not by
   a lock; if that precondition is violated or a new in-scope record is observed,
   the procedure **HALTs**. **No claim of absolute prevention is made.**
5. **Guard scoping is itself a hazard.** A blanket "any subtask member
   disqualifies" rule would have rendered the path unusable on `017-S`, the one
   shipment it exists to close. **T3** scopes task-only to **live** members on the
   executed evidence of **Probe 15**, which measured the **exact** `017-S` member
   set (3 archived tasks + 8 archived subtasks) and found every member byte, path
   and `parent_id` unchanged. Over-tightening a safety guard can destroy the
   objective as surely as under-tightening destroys safety.
6. **Self-hosting.** §8's thirteen-step sequence is mandatory and ordered; step 5
   (reload + positive verification on fresh `main`, across **all three**
   instruction files) is load-bearing (AC-9). §8.2 fixes P-007 ordering **after**
   P1–P10 so a repair can never erase detection evidence. §8.3 gives the
   no-fallback contingency.
7. **Contract/engine divergence.** `header-def.yaml` permits shipment
   `status: shipped` while the engine refuses the generic transition. The path
   depends on *engine* behavior established by the committed probes, never on the
   schema enum. §6.3 binds the engine by **SHA-256 digest**, and a mismatch is a
   **HALT** requiring a probe refresh — there is no advisory fallback.
8. **Blast radius.** P-015, `shipment-reconcile` and `_ship.agent.md` govern every
   shipment closure workspace-wide. Mitigated by additive-only drafting,
   byte-for-byte preservation of the existing exception, mutual exclusivity
   (§6.1), the enumerated 14-clause reconciliation that prevents a
   self-contradictory contract, and two **negative** harness assertions (§7.1
   rows 2 and 12) that fail if anything was widened.
9. **Verification honesty.** §7.2 states what is and is not verified; §7.3 binds
   supported topology strictly to shapes with a **committed** probe. Rev 4.1's
   `%TEMP%`-and-deleted probes are **withdrawn as evidence** (§3).
10. **An executing agent that cannot execute is a planning defect, not a runtime
    surprise.** Rev 5 tested the actor assumption instead of asserting it, and
    found the PR-lifecycle-only Ship route **unexecutable** (Probe 17, §11). The
    blocker is returned to the operator rather than papered over.

## 11. Authorized actors and PR boundary (rev 5, normative) — **BLOCKER**

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

**OPEN — carried into the independent adversarial re-review** (8 items, all P2/P3,
none blocking the artifact-only PR because none is reachable before implementation
begins):

| # | Sev | Open item |
|---|---|---|
| O-1 | P2 | §6.1's inventory is likely **incomplete**: reviewer identified two further sites — `shipment-reconcile/SKILL.md` ~366–368 (Safe-Close Step 0's **binary** close-path selector, which must become ternary) and `_ship.agent.md` 789–791 (the safe-close summary still prescribing `move --status shipped`, which §2 says the engine **refuses** with exit 9). Both plausibly raise the count to ~16 sites. |
| O-2 | P2 | **AC-7's** Safe-Close step-8 cross-reference site is not in the §6.1 table and not in the §10.1 budget. |
| O-3 | P2 | The **committed clause inventory** artifact has no named path and no home in the 4-file surface. |
| O-4 | P2 | §7.2 / AC-11 **overstate harness coverage**: the 12-row table verifies none of G0–G13, the TOCTOU protocol, the baseline scope, or the §8 ordering, so AC-3/4/14/15/19 have **no** automated verification. Either add rows (raising the budget) or state plainly that these are verified by review/diff. |
| O-5 | P2 | §8 omits two steps its own evidence requires: an explicit **claim** step (G12 needs `active`) and an explicit **`backlogit sync`** step (Probe 14 makes successor eligibility conditional on it). |
| O-6 | P2 | The **post-archival closure route** (§8 steps 9–12) was never probed against the same gate Probe 17 falsified. After step 7 archives the shipment there are **zero active shipments**, so a topology gate firing on the closure PR would deadlock **after** the destructive call. |
| O-7 | P2 | Probes ran against a default `backlogit init` workspace (`.backlog/`, stock config); the live workspace is `.backlogit/` with custom `header-def.yaml`, `hooks.yaml`, and `.locks/`. **Configuration is unbound** — configured hooks could mutate records the probes never observed. |
| O-8 | P3 | Residual precision: T4 states liveness by location while §6.2.0 forbids location-derived liveness; §7.1's semantics sentence omits negative row 11 and does not specify a matcher narrow enough to avoid colliding with the preserved block's legitimate "only when" text; and the ≤4 Go helpers are not explicitly excluded from the "1 of 1" red count. |

**Verdict: ADVISORY.** No P0 and no P1 outstanding. The eight open items are
recorded rather than silently closed, and they form part of the re-review scope
in §13.

## 13. Scope for independent adversarial plan review

Reviewers should focus on, in priority order:

1. **Guard sufficiency.** Can any topology satisfy **T1–T5** and G0–G13 yet cause
   `backlogit shipment ship` to archive, return, or reparent an artifact outside
   `allowed_ids`? Committed probes cover the shapes in §7.3.
2. **Detection completeness.** Can any violation escape P1–P10 given detection is
   post-destructive — and is §8.2's "detect before repair" ordering correct?
3. **Recovery soundness.** Is §6.5 sufficient — named gitignored recovery path,
   approval-before-destruction, three-class discipline, enumerated restore,
   `backlogit sync` rehydration, byte/path/dependency equivalence — for errored or
   partially-applied invocations, given that **no** DB restoration and **no**
   atomicity is claimed?
4. **TOCTOU boundary.** §6.2.3 relies on an *execution-mode precondition* for
   new-record exclusion and on *relation-closure locking* for existing records. Is
   that closure complete (inbound edges, links, membership, both directions), and
   is relying on the P-001 precondition legitimate rather than an unfalsifiable
   assumption?
5. **P-015 amendment drafting risk.** Does adding a second exception interact
   badly with the existing fully-covered-root exception or Ship Step 6 pre-mode,
   and is the **14-clause inventory** in §6.1 complete?
6. **Self-hosting ordering (§8) and contingency (§8.3).** Is merge → verify SHA →
   fresh `main` → reload+verify **all three files** → pre-mode → close → P-007
   (per §8.2) → post-mode → closure PR → operational closure → P-020 sound, and
   does the prerequisite's own topology provably satisfy T1–T5?
7. **§7.2/§7.3 honesty and §10.1 sizing.** Is the documentation-vs-behavior
   boundary stated accurately, is supported topology bound to **committed** probes
   only, and is the itemized edit budget (14 clause edits + 4 blocks + 1
   table-driven test + ≤4 helpers) genuinely ≤2 hours?
8. **§11 blocker.** Is the conclusion that Ship cannot execute PR #54 correct, and
   is recording it as an operator blocker the right disposition versus amending
   Ship's Step 5 gate in a separate unit?
9. **The eight OPEN items in §12.2 (O-1 … O-8).** These were raised by internal
   review and **deliberately left open** rather than closed under time pressure.
   Re-reviewers should treat them as in-scope and adjudicate each:
   - **O-1/O-2/O-3** — is the §6.1 clause inventory genuinely complete, and are
     AC-7's step-8 site and the clause-inventory artifact's path accounted for in
     both the inventory and the §10.1 budget?
   - **O-4** — does the 12-row harness actually verify what §7.2 and AC-11 claim,
     or must those claims be narrowed to "verified by review/diff"?
   - **O-5** — does §8 need explicit **claim** and **`backlogit sync`** steps?
   - **O-6** — **highest-risk open item.** The post-archival closure route
     (§8 steps 9–12) is unprobed against the gate Probe 17 falsified; after the
     shipment archives there are **zero active shipments**, so a deadlock there
     lands *after* the destructive call. Should §8 be gated on probing it first?
   - **O-7** — probes ran on a default `backlogit init` workspace, not the live
     `.backlogit` configuration (custom `header-def.yaml`, `hooks.yaml`,
     `.locks/`). Is the configuration delta a material threat to the probe
     conclusions, or a nameable residual?
   - **O-8** — residual precision items (T4 liveness wording, §7.1 row-11 matcher
     breadth, helper functions vs the "1 of 1" red count).

## 14. Sequencing — one artifacts-only Stage PR

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
| 7 | **User merge preauthorization** | The operator's existing merge preauthorization. |
| 8 | **P-009 merge commit** | No squash, no rebase. |
| 9 | **Post-merge verification** | Two-parent merge commit confirmed; merged **artifact set** present on `main`; the **`017-S -> 021-S` dependency edge** verified present on `main`. |

**Actor for every gate above: the OPERATOR.** Per §11 and Probe 17, Ship cannot
execute this PR — its Step 5 topology gate fails closed without an active
shipment, and claiming one is forbidden here. This supersedes rev 4.1's
"Ship, in PR-lifecycle-only mode" designation.

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

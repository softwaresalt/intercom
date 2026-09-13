# Plan — Verified Task-Only Shipment Finalization

> **Requires plan hardening**: **yes — applied in rev 2, extended in rev 3, re-hardened in rev 4 and rev 4.1** (see `## Plan Hardening`).

- **Date:** 2026-09-12
- **Revision:** 4.1 (rev 1 → **FAIL**; rev 2 → **FAIL**; rev 3 → **ADVISORY**; rev 3 → independent adversarial **MUST_REPLAN**; rev 4 → internal review **FAIL**; see §12)
- **Source deliberation:** `docs/decisions/2026-09-12-intercom-go-task-only-shipment-finalization-deliberation.md`
- **Stash origin:** `A10EF3D0` (`kind: deliberation`, `priority: critical`)
- **Branch:** `chore/stage-pipeline-policy-gap` (planning artifacts carried over by cherry-pick; see §14)
- **Engine version of record:** `backlogit 1.10.1-0.20260823032255-b07729386a31+dirty` (`go1.26.5`) — see §6.3 for the `+dirty` uniqueness limitation
- **Harvested backlog:** feature `022-F`, task `022.001-T`, shipment `021-S` (task-only manifest); `017-S depends_on 021-S --type blocks`
- **Status:** planning complete; harvest **COMPLETE** on the policy branch (§14)

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

## 3. Evidence base — fourteen executed probes

All probes ran in throwaway backlogit workspaces under `%TEMP%`, never in this
repository; every sandbox was deleted and the repo tree verified clean
afterwards. Probes 1–4 are from rev 1–3; probes 5–12 were executed in rev 4 to
discharge the adversarial `MUST_REPLAN` conditions; probes **5b** and **13** were
executed in rev 4.1 to close the internal-review findings.

| Probe | Topology | Result |
|---|---|---|
| **1** | `move <shipment> --status shipped` | **exit 9**, refused. Blocker confirmed. |
| **2** | manifest `[T]`; parent `F` live, excluded; `T` has no descendants | `archived_ids=[T,S]`, `returned_ids=[]`; `F` **byte-identical**; `S` → `archived_status: shipped` |
| **3** | manifest `[T]`; `T` owns an **omitted `queued` subtask** | `archived_ids=[ST,T,S]` — omitted subtask **silently archived**, force-stamped `done` |
| **4** | manifest `[T1]`; **omitted sibling** `T2` + its subtask under the same parent | `archived_ids=[T1,S]`; `returned_ids=[T2, T2.ST]` — **non-empty** |
| **5** | **`017-S`-matching**: manifest `[T_live(done), T_arch, ST_arch]`; parent `F` live/excluded/root | `archived_ids=[T_live,S]`, `returned_ids=[]`; parent **and both pre-archived members byte-identical**; `S` → `archived_status: shipped`, `commit` retained |
| **6** | manifest `[T_live]`; pre-archived task + subtask **outside** the manifest under the same parent | `archived_ids=[T_live,S]`, `returned_ids=[]`; out-of-manifest archived descendants **byte-identical** |
| **7** | create a 2nd shipment containing an already-assigned item | **refused**: `item already assigned to shipment 001-S` |
| **8** | `return_blocked` on the covering feature of a parent-first manifest | manifest reduced to task-only; feature → `blocked`; sibling byte-identical; `move --status queued` restores it |
| **9** | live **out-of-manifest sibling** under the same parent, owned by another shipment | `returned_ids=[sibling]`; the sibling's **`parent_id` was STRIPPED** (orphaned); collateral crossed a shipment boundary |
| **10** | recovery: snapshot → ship → quarantine → enumerated restore → `backlogit sync` | **byte/path equivalence PASS**; engine view restored to pre-call `active`; **no DB/WAL byte restoration required** |
| **11** | `shipment ship` against a **`queued`** shipment | **refused**, exit 1, `shipment status conflict` — the record must be `active` |
| **12** | engine version reporting | `backlogit version` line 1 is machine-readable: `1.10.1-0.20260823032255-b07729386a31+dirty` |
| **5b** | **`017-S`-representative**: manifest `[T_live(done), T_arch, ST_arch]` — covering **both** archived member types, including an **archived `subtask` member** | `archived_ids=[T_live,S]`, `returned_ids=[]`; archived task **and archived subtask** members **byte-identical**; parent byte-identical; **the ship-archived live task retained `parent_id`** |
| **13** | **injected partial failure**: archive destination of the shipment record occupied by an unparseable path, then ship | **exit 1** *after* partial mutation — task already archived+stamped, shipment left in the queue directory declaring `status: shipped` (**torn**). §6.5(B) recovery then executed: quarantine → enumerated restore → `backlogit sync` → **byte/path equivalence PASS**, shipment restored to `active` in file **and** engine view, quarantine preserved |

### 3.1 What the probes establish

- **Descendants of a manifest item are ARCHIVED** (Probe 3) — a true correctness
  hazard: `releaseScopeItemIDs` recursively adds every descendant of each
  manifest item before archive-candidate collection.
- **Live descendants of the parent that are outside the manifest are RETURNED,
  and returning STRIPS `parent_id`** (Probes 4 + 9). Rev 3 flagged this as an
  *uncharacterized* risk; **rev 4 confirms it is real**. Probe 9 read the
  returned record and found the `parent_id` field **absent** where it had been
  `001-F` — collateral **orphaning**, not merely a status touch. Probe 9 also
  shows the damage **crosses shipment boundaries**: the orphaned task belonged
  to a *different* shipment whose manifest still lists it. This topology is
  therefore **excluded by G5** and **detected by P1**; it is never authorized.
- **Pre-archived descendants and members are SAFE, inside and outside the
  manifest, of type `task` or `subtask`** (Probes 5, 5b, 6) — byte-identical, not
  returned, not moved, not rewritten, not reparented. Probe 5b is an
  **`017-S`-representative** fixture: it covers **both** archived member types
  (`task` and `subtask`) that `017-S` presents, though with three members rather
  than `017-S`'s twelve. The safety property established is **per-member and
  type-scoped** — archived members of either type are inert — and Probe 6 adds
  that a nested archived subtree is likewise inert, so the result generalises to
  `017-S`'s member count. This **discharges** the adversarial precondition that pre-archived descendants may be supported only
  after an executed fixture matching `017-S` proves exactly that. Support is
  therefore **granted**, bounded to the proven shape (§6.2 G1/G5/G10).
- **A ship-archived live task RETAINS its `parent_id`** (Probe 5b) — the archived
  record carried `parent_id: 001-F`, `status: archived`, `archived_status: done`
  and the merge `commit`. This makes **P9** a satisfiable guard rather than an
  unverified assumption, and distinguishes the *archive* path (lineage preserved)
  from the *return* path (lineage stripped, Probe 9).
- **A partial failure leaves a TORN, self-inconsistent state, and the §6.5(B)
  recovery restores it** (Probe 13). The engine applied mutations and *then*
  failed: the shipment record was left in the queue directory declaring
  `status: shipped` — a state the engine itself refuses to create via `move`
  (Probe 1, exit 9). Bounded recovery restored byte/path equivalence and the
  engine's own view of the record to pre-call `active`. This is the executed
  partial-failure/crash/rehydration probe §6.5 requires; Probe 10 alone (a
  *successful* call, then restore) did **not** establish it.
- **Shared shipment membership is structurally refused at creation** (Probe 7),
  but the manifest is a plain YAML list that hand-editing or a future API could
  still violate, so G11 verifies it independently rather than trusting the engine.
- **The shipment record must be `active`** (Probe 11). Rev 3's §6.3 named the
  call but never stated this precondition; a `queued` record fails closed with
  `shipment status conflict`. Now **G12**.
- **Recovery is achievable without restoring the engine cache** (Probe 10).

## 4. Surface

| # | File | Kind | Change |
|---|---|---|---|
| 1 | `.github/policies/workflow-policies.md` | governing policy | Add a **second, narrower** named exception to P-015 authorizing `TASK_ONLY_FINALIZE`. Existing fully-covered-root exception preserved **byte-for-byte**. Amendment Log entry appended. |
| 2 | `.github/skills/shipment-reconcile/SKILL.md` | procedure contract | Add the `TASK_ONLY_FINALIZE` classification, its guards, and a step-8 fail-closed cross-reference. |
| 3 | `tests/integration/taskonly_finalize_contract_test.go` | generated harness | One characterization test asserting both contract files carry the corrected properties. |

**File count: 3.** **New production code files: 0.**

P-015 states the prohibition *and* names its sole exception, so a skill-only edit
would leave a compliant Ship agent still obligated to refuse the call. P-015 is a
*directly coupled* surface and is therefore in charter. Adding a second, strictly
narrower exception does **not** weaken the first.

## 5. Ordering (test-first; `main` stays green)

1. **H0 — red.** Add the harness only.
   - `go vet ./...` exits **0**;
   - `go test ./...` exits **non-zero** (P-004's literal requirement) **and** the
     targeted `go test ./tests/integration/ -run TestTaskOnlyFinalize_ContractAuthorized`
     exits non-zero;
   - output carries the literal marker `not implemented: task-only finalize contract`;
   - red count over this unit's generated functions is literally **1 of 1**.
2. **H1 — green.** Apply both contract edits. Marker gone; test passes;
   `go vet ./...` exits 0; full `go test ./...` exits **0**.

**This release unit contains exactly one task and exactly one generated
function.** A sequential two-task split is **not** permitted under any
circumstance: a second task under this feature re-creates the Ship Step 4.3
full-suite-green deadlock *and* breaks the §8 self-hosting topology, whose G5
proof depends on the covering feature having exactly one live descendant. If the
work is judged to exceed the 2-hour budget, the correct response is to reduce
scope, not to split the release unit.

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

**Normative-clause reconciliation (mandatory, AC-13).** Every existing clause
that reads "never cascade", "the sole exception", "the only exception", or
equivalent must be located and rewritten to name **both** exceptions, in both
surface files. A second exception added while a "sole exception" sentence
survives elsewhere leaves the contract self-contradictory and the path legally
unexecutable. The mandatory pre-mode and post-mode reconciliation steps of
`shipment-reconcile` are **retained unchanged** — `TASK_ONLY_FINALIZE` adds a
close path, never a reconciliation bypass.

### 6.2 Pre-call guards (ALL required, conjunctive, fail-closed)

**Authorized topology (first authorization, deliberately minimal).** Exactly one
**live** manifest member, of `artifact_type: task`, declaring `status: done` at
finalization; **zero** feature and **zero** shipment members (archived or not);
exactly one live, protected, **root** parent feature outside the manifest; no
omitted **non-archived** task/subtask descendants anywhere under that parent.
Pre-archived members and descendants — inside or outside the manifest, of type
`task` **or** `subtask` — **are** supported, on the executed evidence of Probes 5
and 6. This is exactly the shape `017-S` presents: one live task plus eleven
archived members, eight of them subtasks. Everything else is unsupported and
falls through.

- **G0 — bounded baseline.** Record the pre-call state of the **baseline scope**
  defined in §6.2.1. Every path that changes later must be attributable to this
  call. Pre-existing unrelated modifications inside the baseline scope → **HALT**;
  recovery cannot be bounded.
- **G1 — task-only *live* manifest (corrected rev 4.1).** Every **non-archived**
  manifest member has `artifact_type: task`. Any **feature** or **shipment**
  member disqualifies **unconditionally**, archived or not. **Archived** members
  MAY be `task` or `subtask`: Probes 5 and 6 prove archived members are inert —
  byte-identical, not returned, moved, rewritten or reparented — so they cannot
  contribute to the cascade hazard the guard exists to prevent.
  Type filtering is by the frontmatter `artifact_type` field — **never** by ID
  suffix (`-ST` ends in `T`). Liveness is by the frontmatter `status` field —
  **never** by queue-vs-archive location (G7).
  - **Why this scoping is required, not a loosening.** The motivating case
    `017-S` has a manifest of one live task plus eleven **archived** members, of
    which eight are `subtask`. A blanket "any subtask member disqualifies" rule
    would reject `017-S` outright, so `TASK_ONLY_FINALIZE` could never classify
    the very shipment it exists to close — the objective in §1 would be
    unreachable. The hazard the guard guards against is *live* non-task members,
    which remain disqualifying.
- **G2 — single protected parent (live members).** All **non-archived** manifest
  members share one `parent_id` resolving to an `artifact_type: feature`.
  Archived members are exempt: an archived `subtask`'s parent is legitimately a
  `task`, not the feature, and Probes 5/6 prove those records are not touched.
- **G3 — parent excluded, live, and root.** The parent feature is **not** a
  manifest member, is present in the queue directory, and has **no** `parent_id`
  (a non-root parent disqualifies).
- **G4 — executables terminal.** Every manifest task declares `status: done` or
  is truly `status: archived`, read from frontmatter.
- **G5 — complete parent-rooted live coverage.** Enumerate the parent feature's
  **complete descendant set at every depth** by walking `parent_id` across
  **both** the queue and archive directories. Every descendant that is **not**
  truly `archived` MUST be a manifest member. Equivalently:
  `live_descendants(parent) == { m ∈ manifest : m not truly archived in G7 }`.
  The comparison is against the **non-archived manifest subset**, because a
  truly-`archived` manifest task is by definition excluded from
  `live_descendants` — the exact shape `017-S` presents (one live task plus
  eleven archived members).
  - This single guard subsumes both hazards: omitted descendants of manifest
    items (Probe 3) and omitted live siblings under the parent (Probes 4, 9).
  - It is what makes post-guard **P1** (`returned_ids == []`) sound.
- **G6 — enumeration positively verified.** A zero/complete result must be
  positively confirmed against the live workspace, never inferred from a failed
  or partial scan. Any query error, unindexed directory or malformed record →
  not verified → disqualified.
- **G7 — pre-close snapshot.** Capture, from frontmatter and **before** any
  mutating call: each manifest task's `parent_id` + declared `status`; the parent
  feature's `parent_id`, declared `status`, and **full frontmatter bytes**; and
  the shipment record's declared `status`. Declared status is read from the
  `status` field only — **never** inferred from queue-vs-archive location.
  *(Probes 5/10 reconfirm: `move --status done` relocates a record into the
  archive directory while it still declares `done`; location is not provenance.)*
- **G8 — orphan scan.** No record outside the manifest may declare
  `shipment_id: <this shipment>`. Any orphan disqualifies.
- **G9 — exactly-one-location precheck.** Every manifest task, the shipment
  record, **and** the parent feature must resolve to **exactly one** of the queue
  or archive directories. Present in both (torn) or neither (missing) → **HALT**
  before any mutation, mirroring `RECONCILE_FAIL_SNAPSHOT_AMBIGUOUS` /
  `RECONCILE_FAIL_SNAPSHOT_MISSING`.
- **G10 — exactly one live manifest member (rev 4).** Exactly **one** manifest
  member is non-archived in the G7 snapshot, and by G1 that member is a `task`.
  Multi-live-member manifests are **unsupported in this first authorization** and
  fall through — no probe establishes their safety.
- **G11 — no shared shipment membership (rev 4).** No manifest member is listed
  in any other shipment manifest, and no other shipment manifest lists the parent
  feature. Probe 7 shows the engine refuses this at creation, but the manifest is
  hand-editable, so this is verified independently.
- **G12 — shipment record is `active` (rev 4).** The shipment record declares
  `status: active`. Probe 11: a `queued` record fails closed with
  `shipment status conflict`. The record reaching `active` is the result of
  Ship's ordinary claim, **not** a new privilege granted by this path.
- **G13 — no unresolved blocking dependencies or reference-derived expansion
  (rev 4).** No manifest member has an unresolved incoming/outgoing `blocks`
  dependency on a non-archived record outside the manifest, and no semantic link
  or reference causes the engine's release scope to expand beyond the manifest.
  Probe 9 shows the engine does **not** enforce dependency state itself and
  leaves dangling edges behind, so this guard is the only protection.

#### 6.2.1 Baseline scope (rev 4, normative)

The **baseline scope** for G0 capture, P4 comparison and §6.5 recovery is exactly:

- the shipment record's source-of-truth file;
- each manifest member's source-of-truth file;
- the parent feature's source-of-truth file;
- the source-of-truth file of **every** enumerated descendant of the parent at
  every depth (live and archived), per G5;
- the dependency, link and log records that reference any of the above;
- the queue and archive directories themselves (path inventory, to detect
  additions and deletions, not only content edits).

**Explicitly excluded from mutation comparison:** the expected report and lock
outputs that pre-mode itself produces, and the engine's index/cache database
(gitignored, disposable). These are **not** compared for mutation because they
are *expected* to change; they are instead **validated separately** — the report
for well-formedness and expected content, the lock for correct acquisition and
release, and the index by successful rehydration (§6.5 step 4). Folding expected
outputs into the mutation set would guarantee a false positive on every run.

#### 6.2.2 TOCTOU isolation (rev 4, normative)

Guard evaluation and the authorized call must not be separated by an
unsynchronized window. This procedure **narrows** that window to a bounded,
fail-closed residual; it does **not** eliminate it (see the residual note below).

1. **Bootstrap enumeration (unlocked).** Perform a first, unlocked G5 walk to
   compute the candidate record set. This is an *enumeration only*; no
   classification decision is made from it.
2. Acquire locks — via the repository's file-lock skill or the
   repository-approved equivalent — for the shipment record, every manifest
   member, the parent feature, and every record produced by step 1.
3. **Re-run G5 under lock.** If the locked enumeration differs from step 1's
   result, release and restart once; a second mismatch → **fail closed**. This
   resolves the bootstrap ordering: the lock set cannot be computed by a guard
   that itself requires the locks.
4. Capture the G0 baseline and evaluate G0–G13 entirely under lock.
5. **Immediately before** invoking `ShipShipment`, **recompute** the full
   classification and re-hash the baseline scope.
6. If **any** hash, path, declared status, or dependency edge differs from the
   G0 capture, **fail closed**: release locks, make no destructive call, fall
   through to `SAFE_CLOSE`/HALT.
7. Release locks in a guaranteed-cleanup path on every exit, including failure.

**Residual (rev 4.1, honest bound).** A record **created** after step 5 and
before the engine's own scope collection cannot be locked or detected pre-call,
because it does not exist at enumeration time. That residual is **not** closed
here; it is caught **post-call** by P1/P10 and routed to §6.5(B) bounded
recovery + HALT. The guarantee this section provides is therefore "narrowed to a
fail-closed residual", not "closed".

### 6.3 Authorized call

Only inside the `TASK_ONLY_FINALIZE` classification, and only after §6.2.2 step 3
re-verification passes, invoke the supported
`backlogit shipment ship <shipment_id>` / `ShipShipment` path, passing the same
merge-commit metadata the existing Cascade Close Sub-Procedure passes —
`--sha <merge_sha>` plus `--message` / `--author` where the registry's
`ship_shipment` operation declares them. Probes 5 and 10 confirm the resulting
archived shipment record retains `commit: <merge_sha>`, preserving **P-007**
commit traceability. No other classification may reach this call.

**Engine-version binding.** Before the call, read `backlogit version` and assert
the engine matches the version of record in this plan's header. A mismatch →
**fail closed**, no call.

**Version-string limitation (rev 4.1).** The version of record carries a
`+dirty` suffix, which is **not** a unique build identity: any build from an
uncommitted tree at commit `b07729386a31` reports the identical string. The
version check is therefore a **necessary but not sufficient** control. The
implementing task MUST additionally record and compare a **content digest of the
engine binary** (or pin a clean, tagged release build) so the executing engine is
provably the one §3 characterized. Absent that, the binding must be documented as
advisory rather than as a guarantee (AC-16).

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
  §6.2.1 baseline scope against the G0 capture. The set of paths that actually
  changed MUST equal exactly the transitions implied by **`required_ids`** (a
  pre-archived member implies no transition). `archived_ids` is never trusted as
  the sole detector.
- **P5 — parent preserved.** The parent feature is still in the queue directory
  and its **full frontmatter is byte-identical** to the G7 snapshot. Probes 2, 5
  and 6 establish this is achievable.
- **P6 — no torn records.** No manifest task, **nor the parent feature**, nor the
  shipment record exists in both directories; none is missing from both.
- **P7 — provenance.** The archived shipment record reports
  `archived_status: shipped` and retains the merge `commit`.
- **P8 — pre-archived handling.** A manifest task truly `archived` pre-close is
  correctly **absent** from `archived_ids`; this is **not** an anomaly (Probe 5).
- **P9 — per-task post-verification.** For **each** manifest task in
  `required_ids`, re-read its record and assert: declared `status: archived`;
  resolves to **exactly one** location, the archive directory; and its
  `parent_id` is **unchanged** from the G7 snapshot.
- **P10 — descendant integrity (rev 4).** Every enumerated descendant outside
  `required_ids` — notably every pre-archived descendant — is **byte-identical**
  to its G0 capture, with `parent_id` intact. This is the direct post-check for
  the Probe 9 orphaning hazard.

### 6.5 Failure semantics — two separate paths (rev 4)

Rev 3 blurred these. They are now distinct and must never be conflated.

**(A) Pre-call guard failure — no destructive call was made.**
Any G0–G13 failure, any §6.2.2 re-verification mismatch, any engine-version
mismatch: release locks and fall through to the **existing** `SAFE_CLOSE` or
`HALT` behavior exactly as today. **No recovery is performed, because nothing was
mutated.** This path is non-destructive by construction.

**(B) Post-call failure or errored/indeterminate invocation — mutation may have
occurred.**
Triggered by any post-guard (P1–P10) failure **or** a non-zero, timed-out,
partially-written or unparseable `ShipShipment` result. Perform **bounded
recovery**, then **HALT**. Never fall back to another close path, never re-run
the call, never continue the session.

**Recovery is NOT atomic rollback.** The engine offers no transaction. What is
specified here is a bounded, evidence-based restoration of an enumerated file
set, proven by **Probe 13** (a real errored/partial invocation), with Probe 10
corroborating the rehydration mechanics on a successful call.

Prerequisites (must exist *before* the call):

1. A **durable, workspace-contained pre-call recovery snapshot**: byte copies of
   the exact mutable source-of-truth files in the §6.2.1 baseline scope, plus an
   **inventory manifest** recording each file's relative path and SHA-256.

Recovery procedure:

1. Enumerate current paths in the baseline scope.
2. **Quarantine** — do not delete — any path present now but absent from the
   inventory, moving it to a quarantine directory for operator inspection.
3. **Restore only the enumerated files** from the snapshot, by inventory. Never
   perform a blanket restore of the backlog directory.
4. **Rehydrate the disposable index** by running the official `backlogit sync`.
   The index database is gitignored and is a **disposable cache rebuilt from
   markdown**; this plan makes **no claim** of DB or WAL byte-level restoration,
   and none is required — Probe 10 restored the engine's view of shipment status
   from markdown alone.
5. **Verify byte/path/dependency equivalence** against the inventory: every
   enumerated file present with a matching hash, no unexpected residual paths,
   and the dependency/link edges of the baseline scope unchanged.
6. **HALT** with a named token and a **P-005** event, leaving the quarantine
   directory in place for operator review.

**Live-path approval gate.** In the live (non-probe) path, if destructive
recovery becomes necessary, it requires **explicit operator action-risk approval**
before step 3 executes. Steps 1–2 (enumerate, quarantine) are non-destructive and
may proceed immediately to preserve evidence.

**Probe obligation.** Step 4's rehydration and the whole procedure are not
theoretical: a **disposable partial-failure / crash / rehydration probe must have
been executed** and recorded before this path is authorized. **Probe 13**
discharges this obligation for the engine version of record — it injected a real
destination conflict, observed a **non-zero exit after partial mutation** leaving
a torn record, then executed this procedure to a verified byte/path equivalence
PASS with the engine's own view restored. Probe 10 (successful call, then
restore) is retained as corroboration but does **not**, by itself, discharge the
obligation.

## 7. Verification obligations

### 7.1 Guard → harness assertion mapping

The single characterization test asserts the **presence and precision** of each
guard's load-bearing clause in the installed contract files. Coarse substring
checks are insufficient for the precision points below.

| Guard | Asserted property | Precision point |
|---|---|---|
| G0 | bounded baseline capture | must reference the §6.2.1 scope, not "the backlog directory" |
| G1 | task-only **live** members + `artifact_type` filter | must scope task-only to **non-archived** members, must permit **archived** `subtask` members, must reject ID-suffix filtering, must reject location-derived liveness |
| G2 | single protected parent for **live** members | must exempt archived members explicitly |
| G3 | parent excluded, live, root | must require **no** `parent_id` on the parent |
| G4 | executables terminal | must accept `done` **or** truly `archived` |
| G5 | parent-rooted complete live coverage | must contain "at every depth", the parent-rooted framing, **and** the comparison against the **non-archived manifest subset** |
| G6 | enumeration positively verified | must reject inference from a failed/partial scan |
| G7 | declared status from frontmatter | must reject location-derived status |
| G8 | orphan scan | must name the `shipment_id` declaration test |
| G9 | exactly-one-location precheck | must halt **before** mutation, naming **task, shipment record, and parent feature** |
| G10 | exactly one live manifest member | must state multi-live-member manifests are unsupported |
| G11 | no shared shipment membership | must state it is verified independently of the engine's refusal |
| G12 | shipment record `active` | must cite the `shipment status conflict` failure mode |
| G13 | no unresolved blocking deps / reference expansion | must state the engine does not enforce this |
| §6.2.1 | baseline scope | must enumerate the scope **and** the excluded expected outputs |
| §6.2.2 | TOCTOU isolation | must require re-verification **immediately before** the call, **and** state the unclosed post-step-5 creation residual |
| P1 | `returned_ids == []` | must be tied to G5 as its justification |
| P2 | `archived_ids ⊆ allowed_ids` | must define `allowed_ids` normatively |
| P3 | shipment record unconditional | must contain "regardless of its own pre-close declared status" |
| P4 | independent FS diff | must not rely solely on `archived_ids`; expected set derived from `required_ids` |
| P5 | parent preserved | must require **full frontmatter** byte-identity |
| P6 | torn-record check | must explicitly include the **parent feature** |
| P7 | provenance | must require `archived_status: shipped` **and** merge `commit` retention |
| P8 | pre-archived handling | must state absence from `archived_ids` is **not** an anomaly |
| P9 | per-task post-verification | must name status **and** location **and** `parent_id` |
| P10 | descendant integrity | must name byte-identity **and** `parent_id` intact |
| §6.5 | two separate failure paths | must distinguish (A) non-destructive fall-through from (B) bounded recovery + HALT |
| §6.5 | recovery is not rollback | must disclaim DB/WAL byte restoration |
| §6.1 | classifier order + clause reconciliation | must name both exceptions and retain pre/post reconciliation |

### 7.2 What the harness does and does **not** verify

The Go harness reads installed markdown; it verifies **contract text**, not
runtime behavior. It cannot execute the engine, and there are deliberately **no**
executable negative cases in this unit.

Engine behavior is verified separately by the **disposable characterization
probes** (§3). The plan claims only that:

- engine behavior is established by §3's executed probes, for the engine version
  of record;
- contract text is established by the harness.

Guards G0/G6/G7/G8/G11/G13, §6.2.2 isolation, §6.5 recovery, and the §8 ordering
are **behavioral obligations on Ship**; the harness verifies they are *specified*,
not that they *executed*. Runtime enforcement (an executable classifier + negative
fixtures) is explicitly **deferred** to a separate release unit.

### 7.3 Probe obligations bound to supported topology

Supported runtime topology is bound **only** to shapes with an executed probe:

| Case | Probe | Supported? |
|---|---|---|
| one live task, no descendants | 2 | **yes** |
| one live task + pre-archived task/subtask members (**`017-S`-representative**) | 5, **5b** | **yes** |
| pre-archived descendants outside manifest | 6 | **yes** |
| `parent_id` retained on ship-archive of the live task | **5b** | **verified** |
| omitted queued subtask of a manifest item | 3 | **no** — excluded by G5 |
| omitted live sibling under parent | 4, 9 | **no** — excluded by G5, detected by P1/P10 |
| shared shipment membership | 7 | **no** — excluded by G11 |
| unresolved blocking dependency / link expansion | 9 | **no** — excluded by G13 |
| destination conflict / torn record | 9, **13** | **no** — excluded by G9, detected by P6 |
| `queued` shipment record | 11 | **no** — excluded by G12 |
| **errored/partial invocation → recovery + rehydration** | **13** (10 corroborating) | recovery path **proven** |
| engine-version mismatch | 12 | **no** — excluded by §6.3 version binding |

Multi-**live**-member manifests have **no** probe and are therefore
**unsupported** (G10).

## 8. Self-hosting closure sequence

The prerequisite shipment closes itself using the contract it merges. The full
ordered sequence, with actor, is:

| # | Actor | Step |
|---|---|---|
| 1 | **Ship** | Implement the sole task on an implementation branch; open the implementation PR. |
| 2 | **Ship** | Merge the implementation PR (P-009 merge commit). |
| 3 | **Ship** | **Verify the merge SHA** and record it for P-007 traceability. |
| 4 | **Ship** | Check out a **fresh `main`** and cut the closure branch from it. |
| 5 | **Ship** | **Reload** the merged P-015 + `shipment-reconcile` skill and **positively verify** the new tokens and the merge commit are present in the working checkout. |
| 6 | **Ship** | Run `shipment-reconcile` **pre-mode**. |
| 7 | **Ship** | Classify and close via `TASK_ONLY_FINALIZE` (§6.2 → §6.2.2 → §6.3 → §6.4). |
| 8 | **Ship** | Handle **P-007** commit traceability in explicit order: the verified merge SHA from step 3 is passed at §6.3 and re-asserted by P7. |
| 9 | **Ship** | Run `shipment-reconcile` **post-mode**. |
| 10 | **Ship** | Open and merge the closure PR. |
| 11 | **Ship** | Operational closure. |
| 12 | **Ship** | **P-020** compaction. |
| 13 | — | **Only then** does the successor `017-S` become eligible. |

**Eligibility at step 7.** The harvested artifacts MUST satisfy:

- the covering feature has **exactly one** descendant at any depth — the sole
  task — so G5 holds trivially, G10 holds, and P1 `returned_ids == []` follows;
- the feature is root and excluded from the manifest (G2/G3);
- the task is `done` at closure (G4) and the shipment is `active` (G12);
- no record declares this `shipment_id` outside the manifest (G8);
- no shared membership (G11) and no unresolved blocking dependency (G13) — note
  the `017-S depends_on <this shipment>` edge is **incoming to `017-S`**, not
  outgoing from a member of this manifest, so G13 is satisfied.

**Non-circular:** the authorization used at step 7 is the *merged* artifact from
step 2, not the in-flight working copy. Step 5 (reload + positive verification)
is the load-bearing ordering constraint, bound as AC-9.

### 8.1 Contingency — first real self-close fails after the implementation merged

If step 7 fails after step 2 has already merged:

1. Execute §6.5 path (B): bounded recovery, then HALT.
2. Leave the prerequisite shipment **unresolved** — do not force it to any
   terminal state.
3. Leave `017-S` **blocked** by the dependency edge.
4. Halt for operator review with the recovery report and quarantine directory.
5. **No admin fallback. No cascade fallback. No alternate close path.** The
   merged contract stays on `main`; it is inert until a subsequent, separately
   reviewed session addresses the failure.

## 9. Acceptance criteria (for the harvested task)

- **AC-1** P-015 gains a second named exception authorizing `TASK_ONLY_FINALIZE`; the existing fully-covered-root exception paragraph is **byte-for-byte unchanged** (verified by diff); Amendment Log entry appended.
- **AC-2** `shipment-reconcile` documents the classification, evaluated after `CASCADE` and before `SAFE_CLOSE`, with the mutual-exclusivity rationale (§6.1).
- **AC-3** All of G0–G13 are specified with their stated precision points (§7.1).
- **AC-4** All of P1–P10 are specified, with `allowed_ids`/`required_ids` defined normatively.
- **AC-5** P1 is justified by reference to G5; P4 requires independent filesystem diffing derived from `required_ids`; P9 post-compares each task's declared status, single location, and `parent_id`; P10 asserts descendant byte-identity.
- **AC-6** §6.5 specifies **two separate** failure paths: (A) pre-call guard failure → existing SAFE_CLOSE/HALT with **no** destructive call and **no** recovery; (B) post-call/errored → bounded recovery + HALT with no fallback close path.
- **AC-7** Safe-Close step 8 carries a fail-closed cross-reference to the new path; step 8's existing text is otherwise unchanged.
- **AC-8** P-004: at H0 `go vet ./...` exits 0, `go test ./...` exits non-zero, the marker `not implemented: task-only finalize contract` is present, red count 1 of 1. At H1 the marker is gone and full `go test ./...` exits 0.
- **AC-9** The task records the §8 step 4–5 fresh-`main` checkout, reload and **positive verification** of merged tokens as an execution obligation on Ship.
- **AC-10** The generated function executes unconditionally — no `t.Skip`, build tag, env gate, or selector flag.
- **AC-11** No Ship Role Boundary change, no generic `move shipment → shipped`, no pre-mode bypass, no unrestricted cascade.
- **AC-12** The authorized call passes merge-commit metadata (`--sha`, and `--message`/`--author` where declared), and P7 re-asserts `commit` retention, preserving **P-007**.
- **AC-13** Both surface files are swept against an **enumerated clause inventory** recorded in the task: every clause asserting exclusivity of the cascade exception — including but not limited to the literal phrases "never cascade", "sole exception", "the only exception", and any semantically equivalent paraphrase found by the sweep — is reconciled to name both exceptions. The inventory (clause text + file + line) is committed as part of the change so the sweep is falsifiable and re-runnable; mandatory pre/post reconciliation steps are retained unchanged.
- **AC-14** §6.2.1 baseline scope is specified, including the explicit exclusion of expected pre-mode report/lock outputs from mutation comparison and their separate validation.
- **AC-15** §6.2.2 TOCTOU isolation is specified: bootstrap enumeration, locks over shipment/member/parent/enumerated descendants, re-run of G5 under lock, re-verification immediately before the call, fail-closed on any hash/path/status/dependency change, **and** an explicit statement of the unclosed post-re-verification creation residual.
- **AC-16** §6.3 records the engine version of record, fails closed on mismatch, **and** records a binary content digest (or pins a clean tagged build); if no digest is available the binding is documented as advisory, never as a guarantee.
- **AC-17** The unit contains **exactly one** task and one generated function; no sequential split is permitted.

## 10. Sizing, risks, residuals

**Sizing.** `size: M`, `complexity: high` — recorded on the harvested task as
structured `size` (with `size_source: agent` and a `size_ruleset_version`) plus
`complexity` carried as **labeled prose**, because this workspace's
`.backlogit/header-def.yaml` defines **no** `complexity` field on the `task`
type. That partial degradation is flagged explicitly in the Stage report.

**The 2-hour-rule tension is real and is recorded, not hidden.** Rev 4.1 adds
guards G10–G13, §6.2.1, §6.2.2 and P10, and AC-13 now requires an enumerated
clause sweep. That places the unit at the **upper bound** of M, and review
reasonably questions whether it fits two hours.

Splitting is nevertheless **not available**, and the reason is structural rather
than stylistic: G10 admits **exactly one** live manifest member at closure. A
task moved to `done` declares `done` (not `archived`), so it remains a *live*
member; two completed tasks would therefore present **two** live members and the
prerequisite could not classify itself under the very contract it ships (§8).
Splitting would break self-hosting, not merely inconvenience it.

The resolution is therefore:

- the budget claim rests on every edit being a **pre-specified, pre-located
  textual insertion** into two known files plus one characterization harness —
  drafting, not design;
- if Ship finds the work exceeding the budget **mid-execution**, the required
  response is to **HALT and return the unit to Stage for re-planning** — never to
  split it, and never to rush the contract prose;
- this is recorded below as an **accepted, named residual** rather than an
  estimate asserted as safe.

- **Rollback of the release itself:** three files on a feature branch. Revert the
  merge commit; closure behavior returns to today's blocked state.
- **Failure mode — over-broad drafting.** Mitigated by conjunctive fail-closed
  guards and the §7.1 precision-point assertions.
- **Failure mode — guard drift.** Residual: **accepted**; a CI coupling gate is
  deferred to a separate unit.
- **Residual — runtime enforcement deferred** (§7.2).
- **Residual — multi-parent and multi-live-task manifests unsupported** (G10).
- **Residual — engine-version coupling.** Safety claims bind to the version of
  record; an upgrade requires re-running §3's probes. The `+dirty` version string
  is not a unique build identity (§6.3), so a binary digest is required to make
  the binding a guarantee rather than a heuristic.
- **Residual — 2-hour budget at the upper bound.** Accepted and named: the unit
  cannot be split without breaking G10/self-hosting, so the contingency is
  HALT-and-re-plan, not split.
- **Residual — TOCTOU creation window.** A record created after the §6.2.2
  re-verification is not pre-detectable; it is caught post-call by P1/P10 and
  routed to §6.5(B). Fail-closed, not eliminated.
- **Residual — `017-S` stays blocked** until this prerequisite ships.

## Plan Hardening

1. **Destructive-operation adjacency.** The authorized call mutates the backlog
   with no transaction. Mitigated by G0–G13, §6.2.2 isolation, P1–P10, and the
   §6.5(B) bounded recovery proven by **Probe 13** (Probe 10 corroborating).
2. **Silent scope expansion is reachable and confirmed damaging.** Probe 3
   archives omitted descendants; Probes 4 and 9 return omitted live siblings —
   and Probe 9 proves the return **strips `parent_id`**, orphaning the record and
   doing so **across a shipment boundary**. G5 and P1/P10 are correctness
   requirements derived from executed evidence.
3. **Recovery is not atomic rollback.** Stated plainly in §6.5. No DB/WAL byte
   restoration is claimed; the index is a disposable cache rehydrated by official
   `backlogit sync`. **Probe 13 proves a partial failure is real**: the engine
   applied mutations and *then* errored, leaving the shipment record declaring
   `status: shipped` while still in the queue directory — a torn state the engine
   itself refuses to create directly. Bounded recovery restored byte/path
   equivalence and the engine's own view. Live destructive recovery requires
   operator action-risk approval.
4. **TOCTOU.** Guard evaluation and the call are separated in time by default;
   §6.2.2 **narrows** the window with bootstrap enumeration, locks, a re-run of
   G5 under lock, and immediate-pre-call re-verification. It does **not** close
   it: a record created after re-verification is caught only post-call by
   P1/P10 and routed to §6.5(B). That residual is named, not papered over.
5. **Guard scoping is itself a hazard.** Rev 4's blanket "any subtask member
   disqualifies" would have rendered the path unusable on `017-S`, the one
   shipment it exists to close. Rev 4.1 scopes task-only to **live** members on
   the executed evidence of Probes 5b/6 that archived members are inert. Over-
   tightening a safety guard can destroy the objective as surely as
   under-tightening destroys safety.
6. **Self-hosting.** §8's thirteen-step sequence is mandatory and ordered; step 5
   (reload + positive verification on fresh `main`) is load-bearing (AC-9).
   §8.1 gives the no-fallback contingency.
7. **Contract/engine divergence.** `header-def.yaml` permits shipment
   `status: shipped` while the engine refuses the generic transition. The path
   depends on *engine* behavior established by §3 probes, never on the schema
   enum. §6.3 binds the engine version.
8. **Blast radius.** P-015 and `shipment-reconcile` govern every shipment closure
   workspace-wide. Mitigated by additive-only drafting, byte-for-byte
   preservation of the existing exception, mutual exclusivity (§6.1), and the
   AC-13 clause reconciliation that prevents a self-contradictory contract.
9. **Verification honesty.** §7.2 states what is and is not verified; §7.3 binds
   supported topology strictly to probed shapes.

## 11. Authorized actors and PR boundary (rev 4, normative)

This plan's artifacts ship as **one artifacts-only Stage PR** (§14). The actor
boundary for that PR is:

- **Ship, in PR-lifecycle-only mode**, is the authorized PR actor: it pushes,
  updates, requests review, replies to and resolves review threads, and merges
  Stage's artifacts. It does **not** claim a shipment, does not execute backlog
  work, and does not author planning content.
- **Stage** owns all planning edits. A review comment that requires a planning or
  contract change returns to **Stage**; Ship only pushes the resulting commit and
  replies/resolves the thread.
- **Stage and Orchestrator never push and never merge.** Orchestrator is
  routing-only.

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

## 13. Scope for independent adversarial plan review

Reviewers should focus on, in priority order:

1. **Guard sufficiency.** Can any topology satisfy G0–G13 yet cause
   `backlogit shipment ship` to archive, return, or reparent an artifact outside
   `allowed_ids`? Probes 2–13 cover twelve shapes.
2. **Detection completeness.** Can any violation escape P1–P10 given detection is
   post-destructive?
3. **Recovery soundness.** Is §6.5(B) genuinely sufficient — enumerated restore,
   quarantine, `backlogit sync` rehydration, byte/path/dependency equivalence —
   including for errored or partially-applied invocations, given that no DB
   restoration is claimed?
4. **TOCTOU residual.** §6.2.2 claims only to **narrow** the window to a
   fail-closed residual, via an unlocked bootstrap enumeration, locks, a re-run
   of G5 under lock, and immediate-pre-call re-verification. Is the bootstrap
   ordering genuinely non-circular, is the named post-re-verification creation
   residual the *only* residual, and is routing it to P1/P10 + §6.5(B)
   sufficient?
5. **P-015 amendment drafting risk.** Does adding a second exception interact
   badly with the existing fully-covered-root exception or Ship Step 6 pre-mode,
   and is AC-13's clause reconciliation complete?
6. **Self-hosting ordering (§8) and contingency (§8.1).** Is merge → verify SHA →
   fresh `main` → reload+verify → pre-mode → close → P-007 → post-mode → closure
   PR → operational closure → P-020 sound, and does the prerequisite's own
   topology provably satisfy G0–G13?
7. **§7.2/§7.3 honesty.** Is the documentation-vs-behavior boundary stated
   accurately, and is supported topology correctly bound to probed shapes only?

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

**Merge gates for PR #54** (all required, no admin fallback):

1. fast-forward branch push only;
2. artifact-only scope — zero source, test, or skill implementation files;
3. exact-HEAD local adversarial readiness review;
4. CI green;
5. **P-018**: Copilot review completed against the **current HEAD**, with zero
   unresolved Copilot threads;
6. the operator's existing merge preauthorization;
7. **P-009** merge commit (no squash, no rebase);
8. post-merge verification: two-parent merge commit confirmed, and the merged
   artifact set verified present on `main`.

Planning artifacts from the sibling branch `chore/stage-task-only-shipment-finalization`
(commit `2701906`) were brought across by a clean `git cherry-pick`, preserving
both branch histories with no rebase, reset, or force operation. Provenance is
recorded in the session memory artifact.

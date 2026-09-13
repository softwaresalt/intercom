# Plan — Verified Task-Only Shipment Finalization

> **Requires plan hardening**: **yes — applied in rev 2, extended in rev 3** (see §9).

- **Date:** 2026-09-12
- **Revision:** 3 (rev 1 → plan-review **FAIL**; rev 2 → **FAIL**, narrowed; see §12)
- **Source deliberation:** `docs/decisions/2026-09-12-intercom-go-task-only-shipment-finalization-deliberation.md`
- **Stash origin:** `A10EF3D0` (`kind: deliberation`, `priority: critical`)
- **Branch:** `chore/stage-task-only-shipment-finalization` (base `fdff9e4`, fresh `origin/main`)
- **Status:** planning complete; **backlog harvest BLOCKED** on deliberation §4.1

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

## 3. Evidence base — four executed probes

All probes ran in throwaway backlogit workspaces under `%TEMP%`, never in this
repository; each sandbox was deleted and the repo tree verified clean.

| Probe | Topology | Result |
|---|---|---|
| **1** | `move <shipment> --status shipped` | **exit 9**, refused. Blocker confirmed. |
| **2** | manifest `[T]`; parent `F` live, excluded; `T` has no descendants | `archived_ids=[T,S]`, `returned_ids=[]`; `F` **byte-identical** (`updated_at` unchanged); `S` → `archived_status: shipped` |
| **3** | manifest `[T]`; `T` owns an **omitted `queued` subtask** | `archived_ids=[ST,T,S]` — the omitted subtask was **silently archived** and force-stamped `archived_status: done` |
| **4** | manifest `[T1]`; **omitted sibling** `T2` + `T2`'s subtask, all under the same parent `F` | `archived_ids=[T1,S]` (no collateral **archive**; sibling files survived in `queue/`) but `returned_ids=[T2, T2.ST]` — **non-empty** |

### 3.1 What the probes establish (and correct)

Probes 3 and 4 identify **two structurally distinct** hazards:

- **Descendants of a manifest item are ARCHIVED** (Probe 3) — a true
  correctness hazard, because the engine's `releaseScopeItemIDs` recursively
  adds every descendant of each manifest item before archive-candidate
  collection.
- **Other live descendants of the parent feature are RETURNED, not archived**
  (Probe 4) — their files survive in `queue/`, but `returned_ids` becomes
  **non-empty**.
  - **Rev 3 caveat (review rev-2 F4).** Probe 4 verified only that the sibling
    **files survived**; it did **not** verify their `parent_id` was preserved.
    Backlogit's return path may clear the `parent_id` of a returned descendant,
    which would be collateral **reparenting** even though nothing was archived.
    This plan therefore makes **no safety claim** about the returned-sibling
    topology: it is **uncharacterized**. G5 **excludes** it outright and P1
    **detects** it, so the question never arises on the authorized path.

**Rev 1 defect corrected here.** Rev 1 asserted `returned_ids == []` as a
universal post-guard while defining the topology loosely enough to permit
omitted siblings. Probe 4 shows those two statements are inconsistent: the
guard would have fired on a manifest that rev 1's own guards would have
authorized. Rev 2 resolves this by **tightening the topology** (G5 below) so that
`returned_ids == []` becomes a true, meaningful assertion rather than a
coincidence of the minimal test case.

## 4. Surface

| # | File | Kind | Change |
|---|---|---|---|
| 1 | `.github/policies/workflow-policies.md` | governing policy | Add a **second, narrower** named exception to P-015 authorizing `TASK_ONLY_FINALIZE`. Existing fully-covered-root exception preserved **byte-for-byte**. Amendment Log entry appended. |
| 2 | `.github/skills/shipment-reconcile/SKILL.md` | procedure contract | Add the `TASK_ONLY_FINALIZE` classification, its guards, and a step-8 fail-closed cross-reference. |
| 3 | `tests/integration/taskonly_finalize_contract_test.go` | generated harness | One characterization test asserting both contract files carry the corrected properties. |

**File count: 3.** **New production code files: 0.**

**Rev 1 defect corrected (review F1).** Rev 1 changed only the skill. P-015 in
`workflow-policies.md` states the prohibition *and* names its sole exception, so
a skill-only edit would leave a compliant Ship agent still obligated to refuse
the call — the path would be unexecutable. P-015 is a *directly coupled* surface
and is therefore in charter. Adding a second, strictly narrower exception does
**not** weaken the first.

Following the reviewed `018.008-T` precedent (also 2 contract files + 1 test),
the contract is prompt/policy text, so the harness is a **characterization**
harness reading the installed artifacts. No production stub is needed.

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

**Rev 1 defect corrected (review F7).** Rev 1 specified only the targeted run at
H0; P-004 requires the full-suite red command.

Because the covering feature holds **exactly one** queued task, P-004 (all-red at
H0) and Ship Step 4.3 (full-suite green after the sole task) are simultaneously
satisfiable. **Adding a second queued task under this feature re-creates the Step
4.3 deadlock and is FORBIDDEN.**

## 6. The normative safety contract

### 6.1 Classification `TASK_ONLY_FINALIZE`

Evaluated **after** the existing P-015 `CASCADE` check and **before** falling
back to `SAFE_CLOSE`. The two are mutually exclusive by construction: `CASCADE`
requires at least one qualifying **feature** member; `TASK_ONLY_FINALIZE`
requires **zero** feature members (G1). No manifest can satisfy both, so
ordering creates no ambiguity.

Authorization derives from the **P-015 amendment** (surface 1), not from
ordering. Any guard failure, ambiguity, or query error → **not** classified →
fall through to `SAFE_CLOSE`/HALT exactly as today.

### 6.2 Pre-call guards (ALL required, conjunctive, fail-closed)

- **G0 — clean baseline.** `git status --porcelain -- .backlogit/` records the
  pre-call state. Every path that changes later must be attributable to this
  call. If unrelated `.backlogit/` modifications are already present, **HALT** —
  rollback cannot be bounded (review F4).
- **G1 — task-only manifest.** Every manifest member has
  `artifact_type: task`. Any feature, subtask or shipment member disqualifies.
  Filtering is by the frontmatter `artifact_type` field — **never** by ID suffix
  (`-ST` ends in `T`).
- **G2 — single protected parent.** All manifest tasks share one `parent_id`
  resolving to an `artifact_type: feature`.
- **G3 — parent excluded, live, and root-checked.** The parent feature is **not**
  a manifest member, is present in `.backlogit/queue/`, and its own `parent_id`
  chain is resolved and recorded (a non-root parent disqualifies).
- **G4 — executables terminal.** Every manifest task declares `status: done` or
  is truly `status: archived` (declared status from frontmatter).
- **G5 — complete parent-rooted live coverage (rev 2, replaces rev 1's G5).**
  Enumerate the parent feature's **complete descendant set at every depth** by
  walking `parent_id` live from `.backlogit/queue/` **and** `.backlogit/archive/`.
  Every descendant that is **not** truly `archived` MUST be a manifest member.
  Equivalently: `live_descendants(parent) == { m ∈ manifest : m not truly archived in G7 }`.
  - **Rev 3 correction.** Rev 2 wrote `live_descendants(parent) == manifest`, which
    is **inconsistent with G4/P8**: a truly-`archived` manifest task is by
    definition excluded from `live_descendants`, so the equality could never hold
    for a manifest containing pre-archived members — the very topology P8 promises
    to support. This is not hypothetical: `017-S`'s manifest is one live task plus
    eleven archived members. The comparison is against the **non-archived manifest
    subset**.
  - This single guard subsumes **both** hazards: omitted descendants of manifest
    items (Probe 3) and omitted siblings elsewhere under the parent (Probe 4).
  - It is what makes post-guard **P1** (`returned_ids == []`) sound.
  - Rev 1's G5 walked only downward from manifest tasks and therefore missed the
    Probe 4 sibling case entirely (review F2).
- **G6 — enumeration positively verified.** A zero/complete result must be
  positively confirmed against the live workspace, never inferred from a failed
  or partial scan. Any query error, unindexed directory or malformed record →
  not verified → disqualified.
- **G7 — pre-close snapshot.** Capture, from frontmatter and **before** any
  mutating call: each manifest task's `parent_id` + declared `status`; the parent
  feature's `parent_id`, declared `status`, and **full frontmatter bytes**; and
  the shipment record's declared `status`. Declared status is read from the
  `status` field only — **never** inferred from `queue/` vs `archive/` location.
  *(Probe finding: `move --status done` relocates a record into `archive/` while
  it still declares `done`; location is not provenance.)* Full-frontmatter
  capture is required so post-guard P4 can assert what it claims (review F3).
- **G8 — orphan scan.** Reuse pre-mode's orphan scan: no record outside the
  manifest may declare `shipment_id: <this shipment>`. Any orphan disqualifies
  (review F2).
- **G9 — exactly-one-location precheck (rev 3).** Every manifest task, the
  shipment record, **and** the parent feature must resolve to **exactly one** of
  `.backlogit/queue/` or `.backlogit/archive/`. Present in both (torn) or neither
  (missing) → **HALT** before any mutation, mirroring the existing
  `RECONCILE_FAIL_SNAPSHOT_AMBIGUOUS` / `RECONCILE_FAIL_SNAPSHOT_MISSING`
  semantics. Rev 2 only checked this *after* the call (review, rev-2 F1).

### 6.3 Authorized call

Only inside the `TASK_ONLY_FINALIZE` classification, invoke the supported
`backlogit shipment ship <shipment_id>` / `ShipShipment` path, passing the same
merge-commit metadata the existing Cascade Close Sub-Procedure passes —
`--sha <merge_sha>` plus `--message` / `--author` where the registry's
`ship_shipment` operation declares them. **Rev 3 correction:** rev 2 omitted this,
which would have closed the shipment with no merge association and violated
**P-007** commit traceability. No other classification may reach this call.

### 6.4 Post-call guards (ALL required)

Define, once and normatively (review F3):

- `allowed_ids` = (manifest tasks) ∪ {shipment record}
- `required_ids` = {shipment record} ∪ {manifest tasks **not** truly `archived`
  in the G7 snapshot}. **The shipment record is a `required_ids` member
  unconditionally, regardless of its own pre-close declared status.**

Guards:

- **P1** — `returned_ids == []`. Sound **because** G5 guarantees no live
  out-of-manifest descendant exists. A non-empty result means the topology was
  not what G5 certified → HALT.
- **P2** — `archived_ids ⊆ allowed_ids`. Any ID outside is a cascade.
- **P3** — `required_ids ⊆ archived_ids` (completeness).
- **P4 — independent filesystem verification, not engine self-report.** Diff
  `.backlogit/queue/` and `.backlogit/archive/` against the G0 baseline. The set
  of paths that actually changed MUST equal exactly the transitions implied by
  **`required_ids`** (i.e. derived from the G7 declared-status snapshot, **not**
  from all of `allowed_ids` — a pre-archived member implies no transition). This
  catches a collateral archive the engine failed to report (review F3):
  `archived_ids` is never trusted as the sole detector.
- **P5 — parent preserved.** The parent feature is still in `.backlogit/queue/`
  and its **full frontmatter is byte-identical** to the G7 snapshot (status,
  `parent_id`, and all other fields). Probe 2 establishes this is achievable:
  `updated_at` was unchanged.
- **P6 — no torn records.** No manifest task, **nor the parent feature**, nor the
  shipment record exists in **both** `queue/` and `archive/`; none is missing from
  both. (Rev 3 adds the parent feature, which rev 2 omitted.)
- **P9 — per-task post-verification (rev 3).** For **each** manifest task in
  `required_ids`, re-read its record and assert: declared `status: archived`;
  resolves to **exactly one** location, namely `archive/`; and its `parent_id` is
  **unchanged** from the G7 snapshot. Rev 2 captured `parent_id` in G7 but never
  compared it afterwards, so a task could be archived with cleared or rewritten
  lineage and still satisfy every stated guard (review, rev-2 F1).
- **P7 — provenance.** The archived shipment record reports
  `archived_status: shipped`.
- **P8 — pre-archived handling.** A manifest task truly `archived` pre-close is
  correctly **absent** from `archived_ids` (no transition to report); this is
  **not** an anomaly and is explicitly tested.

### 6.5 Failure handling (rev 3, review F4 + rev-2 F2)

Rollback is triggered by **either** of two conditions — rev 2 covered only the first:

- any post-guard (P1–P9) failure; **or**
- a **non-zero or indeterminate** `ShipShipment` invocation result (timeout,
  partial write, unparseable output). Mutation can occur before any post-guard is
  evaluable, so an errored invocation MUST roll back rather than be re-run.

Roll back from the **G0 baseline**, not blindly:

1. Enumerate exactly the `.backlogit/` paths that changed since G0.
2. For each: `git restore` tracked modifications/deletions, and **explicitly
   delete newly created untracked `archive/` files** (`git restore` alone does
   not remove them — rev 1's blanket `git restore` was unsound).
3. **Rehydrate the engine cache.** `.backlogit/backlogit.db` is **gitignored**, so
   no git operation can restore it, and `ShipShipment` mutates it. After the file
   restore, run `backlogit sync` (and verify it succeeds) so the index no longer
   reports state that the files no longer contain. Without this the tree is
   correct while the engine's view is stale — a silent divergence (rev-2 F2).
4. Re-verify the protected set and assert the tree matches the G0 baseline.
5. HALT with a named token and a **P-005** event.

Never auto-retry, never substitute another close path mid-procedure, never
commit a corrupt backlog.

## 7. Guard → harness assertion mapping (review, scope F1/F4)

The single characterization test asserts the **presence and precision** of each
guard's load-bearing clause in the installed contract files. Coarse substring
checks are insufficient for the three precision points below, which must be
matched on their distinguishing phrase:

| Guard | Asserted property | Precision point |
|---|---|---|
| G1 | task-only + `artifact_type` filter | must reject ID-suffix filtering explicitly |
| G5 | parent-rooted complete live coverage | must contain "at every depth", the parent-rooted framing, **and** the comparison against the **non-archived manifest subset** |
| G7 | declared status from frontmatter | must reject location-derived status |
| G9 | exactly-one-location precheck | must halt **before** mutation, and must name **task, shipment record, and parent feature** |
| P1 | `returned_ids == []` | must be tied to G5 as its justification |
| P3 | shipment record unconditional | must contain the "regardless of its own pre-close declared status" clause |
| P4 | independent FS diff | must not rely solely on `archived_ids`; expected set derived from `required_ids` |
| P6 | torn-record check | must explicitly include the **parent feature**, not only manifest tasks and the shipment |
| P9 | per-task post-verification | must name status **and** location **and** `parent_id` |
| others | presence of the named guard and its halt behavior | — |

### 7.1 What the harness does and does **not** verify (review F6)

**Rev 1 overclaimed.** The Go harness reads installed markdown; it verifies
**contract text**, not runtime behavior. It cannot execute the engine, and there
are deliberately **no** executable negative cases in this unit.

Engine behavior is verified separately by the **disposable characterization**
probes (§3) — which is exactly the evidence the charter scopes to this unit.
The plan claims only that:

- engine behavior is established by §3's executed probes;
- contract text is established by the harness.

Guards G0/G6/G7/G8, §6.5 rollback, and the §8 reload ordering are **behavioral
obligations on Ship**; the harness verifies they are *specified*, not that they
*executed*. Runtime enforcement (an executable classifier + negative fixtures) is
explicitly **deferred** to a separate release unit and is **not** claimed here.

## 8. Self-hosting closure proof

| Step | Actor | Effect |
|---|---|---|
| 1 | Ship | Implements the sole task on an implementation branch; opens PR. |
| 2 | Operator/Orchestrator | Merges the PR — corrected P-015 + skill are on `main`. |
| 3 | Ship | **Reloads** the freshly merged contract from the current checkout before closure. |
| 4 | Ship | Classifies this prerequisite's **own** shipment. |
| 5 | Ship | Closes it via §6.3; post-guards **P1–P9** verify. Shipment reaches `archived_status: shipped`. |
| 6 | — | `017-S` becomes eligible. |

**Eligibility at step 4 (rev 2, review F5).** Rev 1 justified this with "no
archived descendants", which is the **wrong** property — G5 forbids omitted
**non-archived** descendants. The correct construction, which the harvested
artifacts MUST satisfy:

- the covering feature has **exactly one** descendant at any depth: the sole task
  (so G5's `live_descendants(parent) == manifest` holds trivially, and
  P1 `returned_ids == []` follows);
- the feature is root and excluded from the manifest (G2/G3);
- the task is `done` at closure time (G4);
- no record declares this `shipment_id` outside the manifest (G8).

**Non-circular:** the authorization used at step 4 is the *merged* artifact from
step 2, not the in-flight working copy. Step 3 (reload) is the load-bearing
ordering constraint and is bound as an explicit acceptance criterion (§10 AC-9).

## 9. Plan hardening signals

1. **Destructive-operation adjacency.** Mitigated by G0–G9, P1–P9, and
   baseline-derived rollback with cache rehydration (§6.5).
2. **Probe 3 proves silent scope expansion is reachable**; **Probe 4 proves the
   sibling topology yields non-empty `returned_ids`** (and is *not* claimed safe —
   §3.1). G5 and P1 are correctness requirements derived from executed evidence,
   not defensive boilerplate.
3. **Self-hosting.** §8 step 3 reload ordering is mandatory (AC-9).
4. **Contract/engine divergence.** `header-def.yaml` permits shipment
   `status: shipped` while the engine refuses the generic transition. The path
   depends on *engine* behavior established by §3 probes, never on the schema enum.
5. **Blast radius.** P-015 and `shipment-reconcile` govern every shipment closure
   workspace-wide. Mitigated by additive-only drafting, byte-for-byte
   preservation of the existing exception, and mutual exclusivity (§6.1).
6. **Verification honesty.** §7.1 states plainly what is and is not verified;
   runtime enforcement is deferred, not claimed.
7. **Engine cache is outside git.** `.backlogit/backlogit.db` is gitignored and
   is mutated by the authorized call, so file-level rollback alone leaves a stale
   engine view. §6.5 step 3 is mandatory, not hygiene.

## 10. Acceptance criteria (for the harvested task)

- **AC-1** P-015 gains a second named exception authorizing `TASK_ONLY_FINALIZE`; the existing fully-covered-root exception paragraph is **byte-for-byte unchanged** (verified by diff); Amendment Log entry appended.
- **AC-2** `shipment-reconcile` documents the classification, evaluated after `CASCADE` and before `SAFE_CLOSE`, with the mutual-exclusivity rationale (§6.1).
- **AC-3** All of G0–G9 are specified with their stated precision points (§7 table).
- **AC-4** All of P1–P9 are specified, with `allowed_ids`/`required_ids` defined normatively.
- **AC-5** P1 is justified by reference to G5; P4 requires independent filesystem diffing derived from `required_ids` and does not rely solely on `archived_ids`; P9 post-compares each task's declared status, single location, and `parent_id`.
- **AC-6** §6.5 rollback is specified for **both** post-guard failure and errored/indeterminate invocation, includes explicit deletion of untracked `archive/` additions, and includes `backlogit sync` cache rehydration.
- **AC-12** The authorized call passes merge-commit metadata (`--sha`, and `--message`/`--author` where declared), preserving **P-007** commit traceability.
- **AC-7** Safe-Close step 8 carries a fail-closed cross-reference to the new path; step 8's existing text is otherwise unchanged.
- **AC-8** P-004: at H0 `go vet ./...` exits 0, `go test ./...` exits non-zero, the marker `not implemented: task-only finalize contract` is present, red count 1 of 1. At H1 the marker is gone and full `go test ./...` exits 0.
- **AC-9** The task records the mandatory reload ordering (§8 step 3) as an execution obligation on Ship.
- **AC-10** The generated function executes unconditionally — no `t.Skip`, build tag, env gate, or selector flag.
- **AC-11** No Ship Role Boundary change, no generic `move shipment → shipped`, no pre-mode bypass, no unrestricted cascade.

## 11. Sizing, risks, rollback, residuals

**Sizing.** `size: M`, `complexity: high` (two governing contract files with
precision-sensitive prose + one harness). Complexity is carried as
enum-validated **prose**: this workspace's `.backlogit/header-def.yaml` defines
**no** `complexity` field on the `task` type, so `--complexity` fails. The 2-hour
budget remains credible because every edit is a pre-specified, pre-located
textual insertion, but rev 3's added guards (G9, P9) push it to the **upper
bound**. Re-confirm the estimate at harvest; if the contract prose is judged to
exceed 2 hours, split the P-015 amendment (surface 1) from the skill procedure
(surface 2) into two sequential tasks **under the same feature, only one queued
at a time** — never two concurrently queued tasks (§5).

- **Rollback:** three files on a feature branch. Revert the merge commit; closure
  behavior returns to today's blocked state. No data migration, no schema change.
- **Failure mode — over-broad drafting.** Mitigated by conjunctive fail-closed
  guards and the §7 precision-point assertions.
- **Failure mode — guard drift.** A future edit could weaken a guard. Residual:
  **accepted**; a CI coupling gate is explicitly deferred to a separate unit.
- **Residual — runtime enforcement deferred** (§7.1): guards are specified and
  text-verified, not executable-verified, in this unit.
- **Residual — multi-parent manifests** unsupported (deliberation OQ-2).
- **Residual — `017-S` stays blocked** until this prerequisite ships.

## 12. Plan review record

- **Rev 1 → FAIL.** Correctness review returned 5×P1 + 2×P2: (F1) skill-only edit
  leaves P-015's prohibition intact so the path is unexecutable; (F2) G5 missed
  the omitted-sibling topology and the orphan scan; (F3) `required_ids` undefined,
  `archived_ids` trusted as sole detector, parent frontmatter not actually
  captured; (F4) blanket `git restore` is not a sound rollback; (F5) self-hosting
  eligibility used the wrong descendant condition; (F6) characterization
  overclaimed as verifying engine behavior; (F7) P-004 full-suite red omitted at H0.
  Scope audit returned ADVISORY: missing guard→assertion mapping, unbound reload
  AC, tight budget.
- **Rev 2 → FAIL** (narrowed). F1, F2, F5, F6, F7 confirmed **closed**. Remaining:
  (rev-2 F1, P1) G7 captured each task's `parent_id` but no post-guard compared it,
  and the torn-record check omitted the parent; (rev-2 F2, P1) rollback covered only
  post-guard failure, not an errored/indeterminate invocation, and could not restore
  the gitignored `backlogit.db`; (rev-2 F3, P2) `live_descendants(parent) == manifest`
  contradicted G4/P8 for pre-archived members; (rev-2 F4, P2) the Probe 4 narrative
  implied sibling safety without verifying `parent_id` preservation; (rev-2 F5, P2)
  the authorized call omitted merge-commit metadata required by P-007.
- **Rev 3** closes all five: G5 compares against the non-archived manifest subset;
  G9 adds an exactly-one-location precheck; P6 covers the parent; P9 adds per-task
  post-verification of status, location and `parent_id`; P4 derives expected
  transitions from `required_ids`; §6.3 passes merge metadata; §6.5 covers errored
  invocations and rehydrates the engine cache; §3.1 withdraws the sibling-safety claim.
- **Rev 3 → ADVISORY** (final local cycle, attempt 3). All P1 findings confirmed
  closed; three P2 items raised and **fixed in place**: the residual
  "otherwise-safe manifest" wording (now "uncharacterized"), the §7 mapping table
  not binding G5's non-archived-subset / G9's three-record coverage / P6's parent
  coverage, and a stale "P1–P8" cross-reference in §8 (now P1–P9). No P0/P1
  defects outstanding. Independent multi-model adversarial review is delegated to
  the Orchestrator (scope in §13).

## 13. Scope for independent adversarial plan review

Reviewers should focus on, in priority order:

1. **Guard sufficiency.** Can any topology satisfy G0–G9 yet cause
   `backlogit shipment ship` to archive, return, or reparent an artifact outside
   `allowed_ids`? Probes 2–4 cover three shapes only.
2. **Detection completeness.** Can any violation escape P1–P9 given detection is
   post-destructive?
3. **Rollback soundness.** Is §6.5 (baseline diff + untracked deletion +
   `backlogit sync` rehydration) genuinely sufficient, including for errored or
   partially-applied invocations?
4. **P-015 amendment drafting risk.** Does adding a second exception create any
   interaction with the existing fully-covered-root exception or with Ship Step 6
   pre-mode?
5. **Self-hosting ordering (§8).** Is the merge-then-reload-then-close sequence
   sound, and does the prerequisite's own topology provably satisfy G0–G9?
6. **§7.1 honesty.** Is the documentation-vs-behavior verification boundary
   stated accurately, and is deferring runtime enforcement acceptable?

## 14. BLOCKER — harvest deferred

Backlog artifacts were **not** created. On this `origin/main`-based branch the
next allocations are `018-F` and `017-S`, colliding with different artifacts of
those IDs on `chore/stage-pipeline-policy-gap`; and the required dependency
`017-S → prerequisite shipment` is unwritable because `017-S` does not exist
here. Full analysis and recommended sequencing: deliberation §4 / §4.1. Operator
decision required before harvest.

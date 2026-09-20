---
title: "Deliberation — Ship pipeline contract repair and merge-blocking gate reliability (22-entry autonomous staging cycle)"
date: 2026-09-18
status: decided — 9 queued shipments formed; 7 blocked features represented honestly; 2 entries discharged in-cycle under Stage artifact authority
agent: Stage
governs: >-
  stash entries A92E3FA0, BF5DE670, 4A01C53E, 37FAB8C2, 4029DABB, 11B75632, 3F546E63,
  B6EF23CC, EF9352FB, F47DB9A9, 2787DA56, 4C5BEC23, 9D45E62E, 1C6C3B46, 6C24E2E4,
  6B751D8B, 475E76D2, D8397D20, 775E4A35, 29DC2014, 9C658143, 3289EB69
  (22 eligible) plus excluded DB12DA37 (represented by blocked 026-F)
supersedes: nothing — extends the dispositions tables of the 2026-09-08 and 2026-09-11 cycle deliberations
base_commit: 14d44e3c1f28b321e8db24ab8cafcba346996133
stage_branch: chore/stage-ship-pipeline-contract-repair
measured_against: installed surfaces at 14d44e3c (main, post PR #64 / 024-S closure)
outcome: NINE_BOUNDED_SHIPMENTS_QUEUED — ship-contract repair first, gate reliability second, hygiene last
separability: NOT one unit — nine release units with an explicit, partially-ordered DAG
mode: autonomous staging (operator waived the Step 1.5 grouping-selection pause)
revision: 3
revision_note: >-
  rev 2 applied four premise corrections (C-1..C-4, §10) surfaced by plan-review attempt 1 FAIL.
  M-5's live-drift claim WITHDRAWN; M-3 re-characterized; masker clone shown non-identical;
  unignore checker shown latent-not-live and merge-blocking. Gate-chain ordering re-derived (§3.2).
  No §5 decision reversed.
  rev 3 (§11) records plan-review attempt 2 (ADVISORY/ADVISORY/FAIL), the resulting scope
  reductions — S-10 FOLDED into the documentation unit, 3289EB69 HELD as a blocked feature, a
  new Unit 0 kill switch PREPENDED — and the final harvested ID mapping. The shipment count
  therefore moves 10 → 9 and the blocked-feature count 4 → 7. Again, no §5 decision reversed:
  every change is a scope REDUCTION or an honesty correction, not a direction change.
---

## 1. Problem frame

The operator chartered a full autonomous staging cycle over **22 eligible stash entries** with an
explicit priority policy: reliability and security supersede feature work; composability /
interoperability / simplicity / reliability refactors precede feature additions; feature work
supersedes documentation-only work; within comparable importance, order bug → review → feature →
task → spike; group by contextual and code-surface similarity; encode real ordering constraints as
explicit dependency edges; and **do not create one oversized shipment**.

The backlog entering this cycle contains **zero queued work**, 6 blocked features and 11 blocked
tasks, and no active or queued shipment. The single largest risk in this cycle is therefore not
scheduling contention — it is **mis-ranking**: the stash is dominated by low-priority hygiene items
that are cheap and tempting, while the genuinely load-bearing defects are medium-priority entries
describing faults in the *shipping pipeline's own contract*. Those faults have already cost a real
shipment: `4A01C53E` records that 012-S's closure artifact was invisible to the pipeline-topology
gate and **blocked 013-S's `pre_claim` gate** with `PREDECESSOR_CLOSURE_INCOMPLETE`.

The framing question for this cycle is therefore:

> Which of these 22 entries describe defects in controls that *gate every future shipment*, and how
> do we sequence their repair so that each repair is verifiable against a control that is itself
> already trustworthy?

## 2. What was measured this cycle

Every claim below was verified directly against the working tree at `14d44e3c`. Several stash
entries carried stale or materially incorrect premises; those corrections are recorded here rather
than silently inherited, per the precedent set by the 2026-09-10 and 2026-09-11 cycles.

| # | Claim under test | Measured result |
|---|---|---|
| M-1 | `operational-closure/SKILL.md` documents `docs/closure/{YYYY-MM-DD}-{slug}-closure.md` | **CONFIRMED** at `SKILL.md:23`. All 19 real closure artifacts in `docs/closure/` use `{shipment_id}-{feature_id}-post-merge-closure.md`. The documented convention matches **zero** real artifacts. |
| M-2 | `shipment-reconcile/SKILL.md` cites `src/autoharness/**` paths that do not exist here | **CONFIRMED**. Citations at `SKILL.md:503`, `:1043`, `:1168`. `src/autoharness/` does not exist in this Go workspace. |
| M-3 | `_ship.agent.md` Step 4.3 declares `**Format**:` with no format | **CONFIRMED, BUT RE-CHARACTERIZED AT REVIEW (see §10, C-2).** The empty span is at `:440`, not `:442`. It is the **command slot of quality gate #2**, sibling to `1. **Lint**: `go vet ./...`` (`:439`) and `3. **Full Test Suite**: `go test ./...`` (`:441`) — *not* an unspecified evidence format for the full-suite gate. The stash entry's own framing ("the Step 4.3 full-suite gate") is wrong; the defect is a missing formatting command. |
| M-4 | `_ship.agent.md` CASCADE branch directs a direct `backlogit_ship_shipment` call that cannot carry `CLASSIFICATION_BINDING` | **CONFIRMED**. Binding mandate at `:827–832`; direct cascade call at `:861–864`. The `ship_shipment` op accepts only `shipment_id`/`sha`/`message`/`author` on both CLI and registry surfaces. The binding is executable **only** via `shipment-reconcile` `mode:safe-close`. |
| M-5 | `check-retired-architecture.sh` expresses scan scope twice | **PARTIALLY CONFIRMED — AND THE "LIVE DRIFT" CLAIM IS WITHDRAWN (see §10, C-1).** The duplication is real: a pathspec literal in `select_repo_paths` (`:823–831`) and a prefix test in `should_scan_repo_path` (`:188–197`). But they **do not disagree**. The real pathspec is `['git','ls-files','--','config.toml.example','cmd/**','internal/**']` (`:825`) — `cmd/**` **is** enumerated. The `internal/**`-only `ls-files` at `:836–837` belongs to `expected_internal_repo_paths` (`:834–852`), a **self-test oracle**, and is never on the repo-scan path. Furthermore the predicate is **not** a flat disjunction: `cmd/` deliberately *includes* `_test.go`/`testdata/` while `internal/` excludes them — a guarded decision (015.011-T, AG-5/D4) locked by the existing `selection cmd/ coverage` assertion (`:983–1000`). The duplication is additionally already pinned by a merge-blocking self-test, `selection pathspec pin (AC-6/AG-1)` (`:1013–1059`). |
| M-6 | The Python engine is sealed in a bash heredoc | **CONFIRMED**, and larger than reported: lines `123–1176`, ≈1054 lines (the entry said ≈435). `mask_go_non_code` is cloned into `check-write-path-precondition.sh` (`:92, :94, :188`). |
| M-7 | `check-write-path-precondition.sh` selector list is incomplete | **CONFIRMED**. Selectors at `:65–78` are 20 literal qualified identifiers. It does **not** detect import aliasing (`import fs "os"`), `syscall.CreateFile`/`syscall.Write`, Go 1.24 `os.Root` methods (`root.Create`, `root.OpenFile`, `root.Mkdir`, `root.Remove` — no `os.` qualifier, and a natural fit at this module's Go 1.24 floor), or DB drivers beyond `sql.Open`/`bbolt.Open`. It contains **no** check that fires when `pathsafe.Root.Resolve` gains a production caller. |
| M-8 | `pathsafe.Root.Resolve` has zero production callers | **CONFIRMED**. `.Resolve(` returns no matches in any non-test `.go` file under `internal/**` or `cmd/**`. |
| M-9 | `check-unignore-regression.sh` treats any `git show` failure as an empty baseline | **CONFIRMED** at `:146–160`; the `returncode != 0` arm returns `""` unconditionally (`:154–158`). Header comment at `:27` still describes a `git worktree add --detach` design the implementation abandoned in favour of `git show` (`:148`, `:153`). |
| M-10 | CI references two non-existent topology-gate docs | **CONFIRMED** at `ci.yml:512` and `:531`. Neither `docs/pipeline-topology-gate.md` nor `docs/pipeline-topology-gate-ci-rollout.md` exists. |
| M-11 | `staging_gate_actor_contract_test.go` under-covers its ACs | **CONFIRMED**. 123 lines, exactly **one** test function (`TestStagingGateActorContract`) against 027.001-T's multi-criterion AC set. |
| M-12 | Two Stage-authored docs lack frontmatter | **CONFIRMED**. The 2026-09-07 pathsafe containment deliberation and plan both open with an H1, no `---` block, while sibling artifacts carry a full YAML header. |
| M-13 | 5 archived stash entries lack forward-reference notes | **CONFIRMED**. `ECE3DAB7`, `CA469B3D`, `5FE4A7BE`, `5158769F`, `805248F7` are all present in `.backlogit/archive/stash.jsonl` (56 entries total) and none contains the string `STAGE DISPOSITION`. |
| M-14 | Root-level `diagnostic.ps1` needs relocating | **STALE — ALREADY DONE.** Root-level `diagnostic.ps1` does not exist; `scripts/diagnostic.ps1` exists **and is tracked** (`git ls-files` confirms). The relocation 29DC2014 requests has already occurred. |
| M-15 | The Copilot crash report needs containing | **STALE — ALREADY DONE.** The report lives at `.autoharness/staging/report.20260913.140024.22564.0.001.json` and is ignored via `.autoharness/.gitignore:3` (`staging/`). Containment is complete. |
| M-16 | Repository allows squash/rebase merging | **ACCEPTED AS REPORTED** (measured 2026-09-17 on `softwaresalt/intercom`: `allow_merge_commit`, `allow_squash_merge`, `allow_rebase_merge` all `true`). Not re-measured here — re-measuring requires a repo-admin API read, and the entry's own evidence is recent and specific. |

**Two entries are materially stale (M-14, M-15).** Their physical work is already done. This does
not make them no-ops: both entries explicitly ask Stage to *confirm the retention/layout policy*,
which is the part that was never recorded. They are re-scoped accordingly (§4, S-10) rather than
archived as spurious.

## 3. The ranking argument

### 3.1 Why ship-contract repair outranks everything else

Operator rule 1 puts reliability first, and rule 2 puts composability and interoperability
refactors ahead of feature work. Four entries — `4A01C53E`, `4029DABB`, `11B75632`, `9C658143` —
are not ordinary reliability work. They are defects in the **installed execution boundary**: the
agent and skill markdown that *is* the contract in this workspace. Their failure modes are:

* **`4A01C53E`** — a documented convention that matches no real artifact and disagrees with the
  gate's actual glob. Failure mode: **already observed**, blocked 013-S.
* **`4029DABB`** — three citations to a source tree that does not exist here. Failure mode: an
  executing agent looks for the cited implementation, finds nothing, and **halts or improvises**.
* **`11B75632`** — a mandate (carry `CLASSIFICATION_BINDING`) routed through an operation that
  cannot accept it. Failure mode: an executor that follows the direct-cascade branch **silently
  drops the binding**.
* **`9C658143`** — a gate whose required evidence shape is empty. Failure mode: the gate is
  **unverifiable**, so it cannot fail, so it is not a gate.

Each of these compromises the *verification substrate* itself. Repairing anything else first means
verifying that repair with instruments that are known-faulty. That is the decisive ordering
argument, and it is the same argument the 2026-09-08 cycle used to put the U-E1a gate realignment
ahead of feature phase C3.

### 3.2 Why gate-script reliability comes second, and in a strict internal order

`6C24E2E4` and the live residual half of `BF5DE670` both concern **merge-blocking controls** over
`internal/**` — the exact surface every future feature phase lands on. M-7 shows the write-path
precondition gate — protecting a control the Brief marks NON-NEGOTIABLE — can be trivially evaded
four different ways. M-5 shows the retired-architecture gate carries a genuine scope *duplication*,
though **not** the live drift an earlier revision of this deliberation asserted (§10, C-1).

These three units have a **real, non-invented ordering constraint** — but **not** the one originally
recorded. The corrected order is:

1. **Extract** the ≈1054-line Python engine to `scripts/lib/` and choose a single canonical Go
   masker. Must come **first**, because the engine is currently unimportable and untestable in
   isolation, and because every subsequent change is cheaper and safer inside a real module.
2. **Unify the scan scope** inside the extracted module. Must follow step 1: doing it first would
   force the merge-blocking `selection pathspec pin` assertion (`:1013–1059`) to be re-derived
   **twice** — once against the new `SCAN_SCOPE` declaration, again after extraction moves the
   functions out of the shell script the pin reads from disk.
3. **Harden** the write-path selector list and add a library-agnostic `Root.Resolve`-caller
   tripwire. Must follow step 1, because doing it before extraction would **re-clone** the masker
   that step 1 exists to unify — actively worsening the defect 6C24E2E4 names as the root cause.

This is a genuine chain and is encoded as `blocks` edges. Nothing else in this cycle is chained
except where an equally concrete constraint exists (§5).

### 3.3 Why the two "already done" entries still rank last

`775E4A35` and `29DC2014` are measured complete in-tree (M-14, M-15). What remains is a written
retention/layout policy — pure documentation, which operator rule 3 ranks below feature work and
therefore below everything else in this cycle.

### 3.4 Why documentation work is split rather than batched

`F47DB9A9` (topology-gate docs) is documentation, but it is **not independent**: the docs it must
author describe the very naming convention `4A01C53E` is correcting. Authoring them against the
current, wrong convention would produce documentation that is incorrect on the day it lands. It is
therefore chained behind the ship-contract repair, and only then ranked by its documentation
priority. `1C6C3B46` (frontmatter backfill) is also documentation but targets **Stage-authored
artifacts**, which Ship's role boundary forbids it to touch — so it cannot be shipped at all and is
discharged in-cycle instead (§6).

## 4. Options considered

### Option A — One large "reliability" shipment

Batch all ship-contract and gate work into a single release unit.

* **Pro**: one PR, one review cycle, one closure.
* **Con**: directly violates operator rule 5 ("do not create one oversized shipment"). Blast radius
  spans 2 skills, 1 agent template, 2 merge-blocking scripts, and a new `scripts/lib/` package.
  A single review finding anywhere blocks all of it. **Rejected.**

### Option B — One shipment per stash entry

22 shipments, maximal isolation.

* **Pro**: minimal blast radius each.
* **Con**: shatters genuine contextual cohesion. `2787DA56` and `4C5BEC23` are *the same file*;
  splitting them means two PRs touching `check-unignore-regression.sh` with a guaranteed conflict.
  Also ignores that several entries contribute **no task at all** (§6). **Rejected.**

### Option C — Surface-cohesive bounded units with a minimal real DAG *(ADOPTED)*

Group by the code surface actually touched; keep each unit inside the 2-hour rule per task; encode
`blocks` edges **only** where a concrete constraint was measured; and refuse to invent ordering
edges merely to produce a tidy linear sequence.

* **Pro**: satisfies rules 5 and 6 simultaneously. Each unit is independently reviewable and
  independently revertible. Units with no real constraint stay genuinely parallel, which is honest.
* **Con**: ten shipments is a lot of closure overhead — and closure overhead is exactly what
  `4A01C53E` shows is currently defective. This is *mitigated by the ordering*: the closure-contract
  repair is shipment 1, so the remaining nine close against a repaired convention.
* **ADOPTED.** The con is real but self-limiting under the chosen order, and it is strictly better
  than the alternative of leaving the closure contract broken across ten closures.

### Option D — Option C, plus decompose A92E3FA0's phase C3 this cycle

Also stage the agent-adapter feature phase.

* **Pro**: satisfies "fully stage all eligible work" in the most literal reading.
* **Con**: C3 lands new code under `internal/agent`, which is governed by **both** gates this cycle
  repairs. M-5 proves the retired-architecture gate currently mis-scopes that surface; M-7 proves
  the write-path gate would not detect a write primitive C3 plausibly introduces. Staging C3 now
  means planning a feature against two controls that are measurably broken — the precise mistake
  the 2026-09-08 cycle avoided. **Rejected on sequencing**, not on value. C3 is represented as a
  blocked feature with explicit prerequisite edges (§6) so it is visible but not falsely claimable.

## 5. Decisions

* **D-1** — Adopt **Option C**. Ten bounded queued shipments; one covering feature each.
* **D-2** — Rank order: ship-contract repair (1–2) → merge-strategy governance (3) → gate-script
  reliability chain (4–6) → checker/test correctness (7–8) → documentation and hygiene (9–10).
* **D-3** — Encode `blocks` edges **only** for the four measured constraints: S-1→S-2 (closure
  evidence must be discoverable before the cascade route that produces it is re-pointed),
  S-4→S-5→S-6 (§3.2), and S-1→S-9 (§3.4). All other units are genuinely unordered and are left so.
* **D-4** — **Observe D8397D20's interim rule.** Every shipment this cycle carries **exactly one**
  covering feature. The set-based P-015 multi-root trigger therefore does **not** fire, and
  `D8397D20` is correctly retained rather than actioned.
* **D-5** — **Split `B6EF23CC` along the authority boundary.** The GitHub repository-settings change
  is operator-only and is represented as a blocked item. The *standing verification* the entry also
  asks for is implementable and is shipped. It lands **advisory-first**, mirroring the existing
  `PIPELINE_TOPOLOGY_GATE_REQUIRED` advisory→required toggle precedent in this repo, because a
  required gate would fail on arrival until the operator flips the settings. Promotion to required
  is the operator's trigger, recorded in the unit.
* **D-6** — **Discharge `9D45E62E` and `1C6C3B46` in-cycle under Stage's own artifact authority.**
  Both target Stage-owned artifact roots (`.backlogit/archive/stash.jsonl`; `docs/decisions/`,
  `docs/plans/`) that Ship's role boundary explicitly forbids it to modify. Routing them to a
  shipment would create work no Ship agent is permitted to execute. They are performed here.
* **D-7** — **`BF5DE670` splits in two.** Its *gate-hardening* half (M-7) is live, measurable, and
  shipped as S-6. Its *mitigation* half (Lstat-before-write, `O_EXCL` revalidation, hardlink-count
  checks) remains blocked on the unchanged trigger — a live destructive write path — re-measured
  absent this cycle (M-7, M-8). The entry is **not** archived: it stays the sole tracker for the
  deferred mitigation. This preserves six prior cycles' consistent reasoning.
* **D-8** — **`37FAB8C2` stays blocked**, on a re-measured prerequisite: `Root.Resolve` has zero
  production callers (M-8). S-6's tripwire is the mechanism that will *detect* the trigger firing,
  so 37FAB8C2's blocked feature is given a dependency edge on S-6 — converting an un-watched wait
  into a monitored one.
* **D-9** — **`3F546E63` and the residual half of `11B75632` raise an execution-model question**
  (should the shipment claim be reversible or deferrable?) that Stage may not self-authorize; it is
  the same class as 026-F's O-1 and 021-F's execution-model determination. `11B75632`'s *executable*
  half — re-pointing the CASCADE branch at the one surface that accepts the binding — is shipped in
  S-2. The broader claim-reversibility question is deliberated separately in
  `docs/decisions/2026-09-18-intercom-go-shipment-claim-reversibility-deliberation.md` and lands as
  a blocked feature awaiting an operator determination.
* **D-10** — **`475E76D2` and `6B751D8B` re-affirmed ACCEPT-AS-IS / RETAINED**, contributing no
  task. Both triggers re-measured unfired: 475E76D2's fast path guards an already-allocation-free
  loop behind two Win32 syscalls with zero `Root.Resolve` callers (M-8); 6B751D8B's residual items
  (3) buffer pooling and (4) `GOOS=windows` lint scoping are pure performance/CI-runtime
  optimizations whose stated triggers ("call volume grows materially", "CI time becomes a concern")
  have not fired. Deferred **on value, never on safety** — 475E76D2's guard remains mathematically
  sound, as re-proved in the 2026-09-11 cycle.
* **D-11** — **`A92E3FA0` RETAINED** as the standing roadmap tracker and the sole index of five open
  operator decisions (Q3/H3 Dev Tunnel auth → C10; Q6/H4 sessions-per-process → C9; H5 permission
  authorization → C5; H6 session-pointer store semantics → C4; Q7 workspace tooling reach). All five
  re-confirmed open. Archiving it would destroy that index. Phase C3 is represented as a blocked
  feature per Option D's rejection rationale.
* **D-12** — **`EF9352FB` remains operator-only with zero repo-side work.** Re-affirmed without
  reading, printing, or transmitting any secret material; only metadata-level checks were used in
  prior cycles and none were needed here. Deliberately isolated into **no shipment** so it cannot
  stall the 10 implementable units.
* **D-13** — **`DB12DA37` excluded from staging**, already carried by blocked `026-F`. Reconciled in
  place (§7); **no duplicate feature or shipment created**.

## 6. Entries contributing no task this cycle

| Entry | Disposition | Why | Represented as |
|---|---|---|---|
| `A92E3FA0` | RETAINED (tracker) | Not implementable work; sole index of 5 open operator decisions | stays active; C3 → blocked feature |
| `BF5DE670` (mitigation half) | BLOCKED | Trigger unfired: zero destructive write paths, zero `Resolve` callers (M-7/M-8) | blocked feature, dep on S-6 |
| `37FAB8C2` | BLOCKED | Prerequisite unfired: zero production `Root.Resolve` callers (M-8) | blocked feature, dep on S-6 |
| `3F546E63` | BLOCKED | Needs an operator execution-model determination Stage cannot self-grant | blocked feature |
| `EF9352FB` | OPERATOR-ONLY | External credential console; zero repo-side remediation remains | blocked item, no shipment |
| `B6EF23CC` (settings half) | OPERATOR-ONLY | GitHub repo administration is outside Stage *and* Ship authority | blocked item beside S-3 |
| `D8397D20` | RETAINED | Trigger unfired by construction this cycle (D-4) | stays active |
| `475E76D2` | ACCEPT-AS-IS | Re-affirmed; deferred on value, not safety (D-10) | stays active |
| `6B751D8B` | RETAINED | Items (3)/(4) triggers unfired (D-10) | stays active |
| `9D45E62E` | DISCHARGED IN-CYCLE | Stage artifact authority; Ship forbidden (D-6) | done this session |
| `1C6C3B46` | DISCHARGED IN-CYCLE | Stage artifact authority; Ship forbidden (D-6) | done this session |
| `DB12DA37` | EXCLUDED | Already carried by blocked 026-F (D-13) | reconciled in place |

## 7. P-021 C5 / C6 compliance

**Duplicate detection (unconditional).** Run over all 23 active entries pairwise and against the
56 archived entries in `.backlogit/archive/stash.jsonl`. **CLEAN — no duplicate found, nothing
merged, nothing archived as a duplicate, nothing destroyed.** Nearest neighbours were examined
explicitly rather than dismissed on ID:

* `BF5DE670` ↔ `37FAB8C2` — same *gating shape* (both blocked on a missing caller) but disjoint
  deliverables (TOCTOU/hardlink mitigation vs. a black-box integration test). Not duplicates.
* `2787DA56` ↔ `4C5BEC23` — same file, but disjoint defects (fail-closed ref resolution vs. a stale
  header comment). Not duplicates; correctly **co-located in one shipment**, which is grouping, not
  merging.
* `6C24E2E4` ↔ `BF5DE670` — both touch the `mask_go_non_code` clone, but from opposite ends
  (extraction vs. selector hardening). Not duplicates; correctly **chained**.
* `11B75632` ↔ `3F546E63` — both concern the Ship claim/cascade contract, but one is a routing
  repair and the other an execution-model question. Not duplicates; **split by authority**.
* `9C658143` ↔ `026-F` — both touch `_ship.agent.md`, at disjoint clause sites (`:442` Step 4.3 vs.
  `:339–367` Steps 2–3). Not duplicates. Recorded as a *related* link, not a dependency.

**Late-identifier reconciliation (triggered by any `N/A` source ref).** Ship-owned residual and
closure records citing each entry were searched, joined on entry ID and shipment ID.

| Entry | `N/A` fields | Outcome |
|---|---|---|
| `4A01C53E` | PR, review-thread | **No late identifier found.** Recorded as a pre-PR local-review finding on the 012-S closure repair commit; no PR or thread ever existed. Terminal, truthful. |
| `37FAB8C2` | review-thread | **No late identifier found.** PR #40 already concrete. `DISCOVERY-STATUS: LOOKUP-UNAVAILABLE` token consumed as a scan accelerator; the scan ran and was clean. |
| `11B75632` | review-thread, task, feature | **No late identifier found.** Originated as a P-013.6 escalation finding, never a GitHub thread. Terminal. |
| `3F546E63` | review-thread, task, feature | **No late identifier found.** Stage feasibility assessment, never a thread. Terminal. |
| `B6EF23CC` | review-thread | **No late identifier found.** Surfaced by Stage's own pre-merge P-009 revalidation. Terminal. |
| `775E4A35`, `29DC2014` | task, PR, review-thread | **No late identifier found.** Pre-claim operator-instructed dispositions; no PR open at capture. Terminal. |
| `9C658143` | all | **No late identifier found.** Surfaced by Stage during 026-F contract measurement. Terminal. |
| `EF9352FB` | review-thread | **Idempotent no-op.** PR #4 already reconciled 2026-09-04; thread remains terminal `N/A` (governing closure record documents zero review comments). No concrete identifier overwritten. |
| `475E76D2` | review-thread | **Idempotent no-op.** PR #51 already reconciled 2026-09-11; thread terminal `N/A` (local Go Reviewer pass). |
| `BF5DE670` | task, review-thread | **Idempotent no-op.** PR #27 already reconciled; task genuinely cross-cutting. |
| `F47DB9A9`, `2787DA56`, `4C5BEC23`, `9D45E62E`, `1C6C3B46` | various | **No late identifier found** for the `N/A` fields; concrete fields (PR #33, threads `PRRC_kwDOTPuhps7rri5p`/`…59` on 1C6C3B46) left untouched. |
| `DB12DA37` | review-thread | **Idempotent no-op.** Already reconciled 2026-09-18 as a truthful terminal `N/A`. |

No `N/A` was written over a concrete value; no concrete value was rewritten. Every reconciliation
was applied **in place** under Stage's own stash authority — no Ship write was requested, and the
C5 capture-only carve-out and the single-write capture invariant are preserved unweakened. **No
missing late identifier gated deliberation, planning, or harvest.**

## 8. Dispositions table

| Entry | Pri | Group | Rank | Disposition | Target |
|---|---|---|---|---|---|
| `4A01C53E` | med | A — ship contract | 1 | HARVESTED | S-1 |
| `4029DABB` | med | A — ship contract | 1 | HARVESTED | S-1 |
| `11B75632` | med | A — ship contract | 2 | HARVESTED (executable half) | S-2 |
| `9C658143` | low | A — ship contract | 2 | HARVESTED | S-2 |
| `B6EF23CC` | med | B — security/governance | 3 | HARVESTED (verification half) | S-3 |
| `6C24E2E4` | low | D — gate reliability | 4, 5 | HARVESTED (split: scope, then extraction) | S-4, S-5 |
| `BF5DE670` | med | C — pathsafe/gate | 6 | HARVESTED (gate-hardening half) | S-6 |
| `2787DA56` | low | D — checker correctness | 7 | HARVESTED | S-7 |
| `4C5BEC23` | low | D — checker correctness | 7 | HARVESTED | S-7 |
| `3289EB69` | low | D — test coverage | 8 | HARVESTED | S-8 |
| `F47DB9A9` | low | E — documentation | 9 | HARVESTED | S-9 |
| `775E4A35` | low | E — hygiene | 10 | HARVESTED (re-scoped to policy, M-15) | S-10 |
| `29DC2014` | low | E — hygiene | 10 | HARVESTED (re-scoped to policy, M-14) | S-10 |
| `9D45E62E` | low | E — backlog hygiene | — | DISCHARGED IN-CYCLE | this session |
| `1C6C3B46` | low | E — docs hygiene | — | DISCHARGED IN-CYCLE | this session |
| `37FAB8C2` | med | C — pathsafe | — | BLOCKED (prereq unfired) | blocked feature |
| `BF5DE670` (mitigation) | med | C — pathsafe | — | BLOCKED (trigger unfired) | blocked feature |
| `3F546E63` | med | A — ship contract | — | BLOCKED (operator determination) | blocked feature |
| `EF9352FB` | low | B — security | — | OPERATOR-ONLY, no shipment | blocked item |
| `D8397D20` | low | A — ship contract | — | RETAINED (trigger unfired, D-4) | stays active |
| `475E76D2` | low | C — pathsafe | — | ACCEPT-AS-IS (D-10) | stays active |
| `6B751D8B` | low | C — pathsafe | — | RETAINED (D-10) | stays active |
| `A92E3FA0` | high | F — product direction | — | RETAINED (tracker, D-11) | stays active; C3 blocked |
| `DB12DA37` | high | A — ship contract | — | EXCLUDED (D-13) | 026-F, reconciled in place |

## 9. Open questions carried forward

* **Q-A** *(operator)* — Flip `allow_squash_merge` and `allow_rebase_merge` to `false` on
  `softwaresalt/intercom`, then promote S-3's advisory check to required. Blocks the settings half
  of `B6EF23CC`.
* **Q-B** *(operator)* — Should the shipment claim become reversible (`unclaim`/`release`) or
  deferrable (claim-late / claim-on-merge)? Blocks `3F546E63`. See the dedicated deliberation.
* **Q-C** *(operator)* — Rotate the Tavily API key. Blocks `EF9352FB`. Time-sensitive; unchanged.
* **Q-D..Q-H** *(operator)* — The five standing A92E3FA0 decisions (Q3/H3, Q6/H4, H5, H6, Q7),
  all re-confirmed open.
* **Q-I** *(operator)* — 026-F's O-1 and O-3 remain the gate on the harness-selection conformance
  repair. Untouched by this cycle.

## 10. Corrections applied at the plan-review gate (revision 2)

Plan-review attempt 1 returned **FAIL** on all three plans (independent reviewers: Scope Boundary
Auditor, Correctness Reviewer, Maintainability Reviewer). Four of the findings refuted premises
recorded in revision 1 of this deliberation. They are corrected here rather than silently dropped,
per this repository's standing practice.

* **C-1 — M-5's "live drift" claim is WITHDRAWN.** Revision 1 asserted that `run_repo_scan` globs
  only `internal/**` and therefore leaves `cmd/**` silently unscanned. **False.** The repo-scan
  pathspec is `['git','ls-files','--','config.toml.example','cmd/**','internal/**']`
  (`check-retired-architecture.sh:825`); the `internal/**`-only glob at `:836–837` belongs to the
  self-test oracle `expected_internal_repo_paths`. The existing assertion `selection cmd/ coverage
  (AG-5/D4)` (`:983–1000`) would already be red if the claim were true. The duplication 6C24E2E4
  reports is real; the drift is not. **Consequence:** the scope-unification unit is demoted from
  "fixes a live gate defect" to "removes a duplication that is already mechanically pinned", it is
  **re-ordered after** the extraction, and its ordering edge is re-derived (§3.2).
* **C-2 — M-3 is RE-CHARACTERIZED.** The empty `**Format**:` is at `_ship.agent.md:440` and is the
  **command slot of quality gate #2**, not an evidence-format declaration for the full-suite gate
  #3 at `:441`. The stash entry `9C658143` inherited the same misreading. **Consequence:** the
  repair is to supply the repository's canonical formatting command, and the full-suite text at
  `:441` is left byte-unchanged — which *satisfies* rather than violates blocked 026-F's declared
  "Step 4.3 full-suite text unmodified" anti-regression pin.
* **C-3 — The `mask_go_non_code` "clone" is NOT identical.** The retired-architecture version
  (`:432–548`) carries raw-string buffering, struct-tag unmasking, and an unterminated-raw-string
  fail-closed branch that the write-path version (`check-write-path-precondition.sh:91–175`) lacks;
  the divergence is documented as deliberate at `check-retired-architecture.sh:92–104`.
  **Consequence:** unification cannot be verdict-neutral for the write-path gate. The extraction
  unit must *choose and justify* a canonical implementation and re-baseline write-path verdicts,
  rather than assert a zero delta.
* **C-4 — `check-unignore-regression.sh` does NOT fail open end-to-end, and IS merge-blocking.**
  `run_differential_check` hard-errors on an unresolvable base ref via `git diff` before
  `root_gitignore_text_at` is reached, and `run_denylist_check` is only ever called with `HEAD`. The
  defect is **latent defence-in-depth**, not a live bypass. Separately, the checker runs in
  `ci.yml`'s `gitignore-append-only` job with no `continue-on-error`, and that job is in `ci-gate`'s
  `needs` list — so it *is* a merge-blocking control. **Consequence:** that unit's severity claim is
  corrected, its plan is re-declared as requiring hardening, and its red phase is moved to the
  function level.

**Three further structural findings were accepted and applied across all plans:** (i) no plan may
pre-stamp its own review verdict in frontmatter; (ii) every test-creating task must carry an
assertion that is demonstrably false pre-change — two were green-on-arrival and have been removed or
re-derived; (iii) the pipeline-topology gate lives in an **external** `autoharness` binary, so no
in-repo test can assert parity with its glob — that acceptance criterion has been re-scoped to
document↔artifact conformance, which is genuinely falsifiable.

**No decision in §5 is reversed by these corrections.** The ranking, the ten-unit split, the
authority splits (D-5, D-6, D-9), and every blocked/retained disposition stand. What changed is the
*internal ordering* of the gate chain (§3.2), the *characterization* of three defects, and the
acceptance surfaces of six tasks.

---

## 11. Plan-review attempt 2, adopted dispositions, and the final harvested mapping (revision 3)

### 11.1 Gate outcomes

The plan-review budget permits **two re-entry cycles**. Both were spent, so attempt 2 was the
final permitted gate and its verdicts are terminal for this cycle.

| Plan | Attempt 1 | Attempt 2 | P0 | P1 | Disposition adopted |
|---|---|---|---|---|---|
| `…ship-pipeline-contract-repair-plan.md` | FAIL | **ADVISORY** | 0 | 2 | Both attempt-1 P0s independently re-verified closed. P1s applied → rev 4. Shipped as units 1–2. |
| `…gate-reliability-plan.md` | FAIL | **ADVISORY** | 0 | 4 | All four attempt-1 P0s closed. P1s applied → rev 4. Shipped as units 3–7. |
| `…checker-test-docs-hygiene-plan.md` | FAIL | **FAIL** | 0 | 5 | Reviewer's own disposition adopted verbatim: *ship the two sound units, hold the rest.* → rev 4. |

Attempt 3 was **not** run, on two independent grounds, either of which alone is sufficient:

1. **Budget.** Step 4 permits a maximum of two re-entry cycles; attempt 3 would be a P-013.6
   escalation, not another attempt.
2. **STOP RULE fired.** The compound record
   `docs/compound/workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md`
   mandates that when two consecutive gates produce a finding worded *"recurs" / "same class as" /
   "relocated" / "partially closed"*, the correct next action is a **compound-learning pass, not
   another gate**. Attempt 2's reviewer of plan 1 wrote *"the same defect class as attempt-1's
   P0-2, now under-constrained"*. The rule fired on its literal trigger.

That compound-learning obligation is discharged this session (see §11.5).

### 11.2 Scope changes forced by attempt 2

* **Unit 0 PREPENDED (kill switch).** The write-path gate is merge-blocking, and units 5–7 modify
  it. A reviewer established that a mis-scoped change would wedge the pipeline with no escape
  hatch. A kill switch therefore becomes a **hard prerequisite** of the gate-modifying units, and
  is now shipment unit 3 (`030-F`), blocking units 5 and 7.
* **S-10 FOLDED into the documentation unit.** `775E4A35` and `29DC2014` were both re-measured as
  **already physically satisfied** at `14d44e3c`. A standalone shipment whose every task is
  green-on-arrival is a ceremony, not a release unit. They survive as verification-and-record tasks
  (`036.003-T`, `036.004-T`) inside unit 9. Shipment count 10 → 9.
* **`3289EB69` HELD, not shipped.** Its gap list was measured against the **wrong file**. Shipping
  it would have added assertions that pass on arrival and certify nothing. Held as blocked
  `037-F` with a five-item unblocking condition.
* **`034-F` task split.** The combined call-extent + allowance task was sized **L/high**, which the
  2-hour rule forbids. Split into a de-risking spike (`034.001-T`) plus two bounded `M` tasks.

### 11.3 Final mapping — stash entry → feature → shipment

| Unit | Shipment | Feature | Tasks | Stash entries | Blocked by |
|---|---|---|---|---|---|
| 1 | `025-S` | `028-F` | 3 | `4A01C53E`, `4029DABB` | — |
| 2 | `026-S` | `029-F` | 5 | `11B75632`, `9C658143` | `025-S` |
| 3 | `027-S` | `030-F` | 2 | *(kill switch, review-forced)* | — |
| 4 | `028-S` | `031-F` | 5 | `B6EF23CC` (verification half) | — |
| 5 | `029-S` | `032-F` | 7 | `6C24E2E4` (extraction half) | `027-S` |
| 6 | `030-S` | `033-F` | 3 | `6C24E2E4` (scope half) | `029-S` |
| 7 | `031-S` | `034-F` | 8 | `BF5DE670` (detection half) | `027-S`, `029-S` |
| 8 | `032-S` | `035-F` | 3 | `2787DA56`, `4C5BEC23` | — |
| 9 | `033-S` | `036-F` | 4 | `F47DB9A9`, `775E4A35`, `29DC2014` | `025-S` |

**Blocked features (no shipment — deliberately unclaimable).**

| Feature | Entry | Why blocked | Unblocks when |
|---|---|---|---|
| `037-F` | `3289EB69` | Gap list measured against the wrong file | 5-item re-measurement condition met |
| `038-F` | `37FAB8C2` | Zero production `Root.Resolve` callers — *unperformable*, not unscheduled | `034.005-T` tripwire fires |
| `039-F` | `BF5DE670` (mitigation) | Same unfired trigger; AC-4.5 names adding a write primitive a **stop condition** | a production write path appears |
| `040-F` | `3F546E63` | One-way-door execution-model choice | operator determination **O-5 / Q-B** |
| `041-F` | `EF9352FB` | External credential console | operator action (no agent authority) |
| `042-F` | `B6EF23CC` (settings) | GitHub repo administration | operator action **Q-A** |
| `043-F` | `A92E3FA0` (phase C3) | Feature work ranks below reliability/refactor work | units 1–9 land, then its own deliberation |

**Retained active in the stash (5):** `A92E3FA0` (tracker), `D8397D20`, `475E76D2`, `6B751D8B`
(all trigger-unfired or deferred on value), `DB12DA37` (excluded — carried by blocked `026-F`;
**no duplicate created**).

### 11.4 Honesty check on the shipped set

Three of the nine units contain work that is at least partly **green-on-arrival**, and each says so
in its own task text rather than claiming a red phase it does not have: unit 6's
derivation-agreement assertion, unit 9's verification tasks, and the characterization half of unit
5. Declaring this is the point — an undeclared green-on-arrival acceptance criterion is precisely
the defect pattern that produced two FAIL gates this cycle.

### 11.5 Compound-learning obligation

Discharged this session at
`docs/compound/workflow-issues/staging-premise-verification-before-planning-2026-09-18.md`.
This also discharges the **overdue O-3** obligation recorded against `026-F`.

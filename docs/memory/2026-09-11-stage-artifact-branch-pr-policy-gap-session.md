---
title: "Stage session: Stage artifact branch/PR policy gap correction (017-S)"
date: 2026-09-11
agent: stage
shipment: 017-S
feature: 018-F
branch: chore/stage-pipeline-policy-gap
status: complete
---

# Stage session memory — Stage artifact branch/PR policy gap correction

**Date**: 2026-09-11
**Agent**: Stage
**Branch**: `chore/stage-pipeline-policy-gap`
**Phase**: complete — shipment queued, handed off to Ship
**Session outcome**: exactly one reviewed, queued shipment — `017-S`

---

## 1. Entry point

Direct operator request (not an existing stash entry): correct the pipeline policy gap that
allowed Stage artifacts to be pushed directly to `main`, evidenced by commit `fdff9e4`. Captured
through the configured backlog intake as stash entry **`638A410B`** (kind `feature` — `--kind
chore` is rejected by this backlogit build), then run through the normal
deliberate → plan → harden → review → harvest pipeline. No ad hoc tracker was created.

## 2. Tool state

- **`TOOL_DEGRADED`** — no backlogit MCP tools were exposed in this session. Fell back to the
  registry-declared `backlogit` CLI (v1.10.1). All operations below used the CLI fallback.
- `INDEX_SYNC_OK` at three distinct phases. The index total is **backlog items + *active* stash
  entries** (verified against `.backlogit/backlogit.db`: `items` + `stash_entries`), so it moves
  with intake and stash archival as well as with harvest. Labelled by phase and UTC timestamp:

  | Phase | Timestamp (UTC) | Items | Active stash | Total |
  |---|---|---|---|---|
  | Step 0.1 session start — **before** stash intake | before `2026-09-11T16:51:55Z` | 186 | 13 | **199** |
  | Post-shipment-assembly, **before** Step 5.6 stash archive | `17:26:03Z`–`17:27:26Z` | 207 | 14 | **221** |
  | True session end — **after** Step 5.6 | from `2026-09-11T17:27:26Z` | 207 | 13 | **220** |

  The 221 reading is the transient peak in the ~83 s window after shipment `017-S` was created
  (`17:26:03Z`) and before stash entry `638A410B` was archived (`archived_at
  2026-09-11T17:27:26Z`); the 14th stash entry *is* `638A410B`. **220 is the steady-state count**
  and the one the PR description reports. Corroboration: the immediately prior session
  (`docs/memory/2026-09-11-stage-pathsafe-deferred-trigger-reassessment-no-shipment.md`) recorded
  the same 199 pre-session baseline, the 21 items created here (`018-F`, 11 archived provisional
  units, 8 atomic tasks, `017-S`) take 186 → 207, and `backlogit sync` re-observed 220 at cycle 2
  (§10.4) and again at cycle 3 (§11).
- Checkpoint recovery: zero `stage`-owned active checkpoints, no validation/quarantine anomalies
  → normal startup, not a failure.

## 3. Artifacts produced

| Path | Role |
|---|---|
| `docs/decisions/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-deliberation.md` | Source document / P-003 lineage root (§8 rev-2 addendum, §9 rev-3 addendum) |
| `docs/plans/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-plan.md` | Implementation plan, revision 3, `status: reviewed` (§10a hardening, §10 Constitution Check, §12 review record) |
| `docs/memory/2026-09-11-stage-artifact-branch-pr-policy-gap-session.md` | This file |

## 4. Decision summary

Chose **Option 2**: prose correction of every authorization surface **plus** a deterministic CI
regression gate. Rejected: prose-only (D4 — unreliable), a new policy ID (scope), and a
generalized branch framework (operator simplicity constraint).

**Root cause — four authorization surfaces** (the count was revised twice; see §5):

1. `.github/agents/_stage.agent.md` Role Boundary **Git row** — the *operative* surface
2. `.github/policies/workflow-policies.md` **P-010 Stage MAY** bullet
3. `.github/agents/_orchestrator.agent.md` Step 1.5 **sub-step 3e**
4. `.github/agents/_orchestrator.agent.md` Step 1.5 **preamble**

**D2 (template coupling)** — the `.tmpl` sources under
`.copilot/installed-plugins/autoharness/autoharness/templates/` are **gitignored and untracked**
(verified with `git check-ignore -v` / `git ls-files --error-unmatch`). Editing them is
uncommittable, unreviewable, and erased on reinstall, and no CI check can enforce
template↔artifact equality. Operator scope item 1 was therefore narrowed: the gate asserts the
**invariant over the installed artifacts**, so a regeneration that restores the language turns CI
red. The Scope Boundary Auditor independently reviewed and endorsed this narrowing.

## 5. What the review gate actually caught

This is the most reusable part of the session. Three of my own framing claims were **falsified by
review and independently re-verified by me** before being withdrawn:

- "Exactly one push-shaped hit in the corpus" — **false**. The search was push-shaped; the
  *operative* surface is **commit-shaped** (`_stage.agent.md` Git row). Round 1, P0.
- The rev-1 P-010 wording was **self-deadlocking**: it required Stage to merge a PR while Stage's
  own Role Boundary forbids creating, pushing, or merging PRs. Fixed by naming the **Orchestrator**
  as the PR actor. Round 1, P0.
- "Deleting 3e loses no behaviour" — **false**. 3e carried the only branch-point instruction
  ("from the current commit") for the already-committed-but-unpushed path. Folded into 3a. Round 1, P0.

Then round 2 found a **fourth** surface (the Step 1.5 preamble) that survived *both* the
push-shaped and commit-shaped sweeps because it reads as a verification instruction while stating
the gate's postcondition as artifacts being committed *to* the default branch.

**Lesson (recorded in deliberation §9.2)**: manual enumeration of contract surfaces is not
reliable at this scale — which is itself the argument for D4. The plan now requires the detector,
not a reader, to establish the surface count (AC-A3.4 records the first full-corpus scan before
any correction begins; risk R9 tracks a possible fifth).

## 6. Highest-risk design element

The detector must reject imperative direct-push/commit-to-default *instructions* while accepting
*prohibitions about the same topic*. Prior art
`docs/compound/2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md` records a CI
grep that permanently self-failed by matching its own describing text — the exact failure mode here.

Rev-3 design decisions that came out of review:

- **Per-construct verdicts, never block-level** — a normalized Step 1.5 block contains both an
  accept (3b `create a PR to \`main\``) and a reject (3e); block-level accept would mask it.
- **Bold lead-in, not ATX heading**, for prohibition context — the real files use
  `**Ship MUST NOT**:` lead-ins, so an ATX-only reading would flag correct text and make
  zero-findings unreachable.
- **Cell-scoped, not row-scoped** table matching — a Role Boundary row's Allowed cell holds push
  verbs and its Forbidden cell holds `main`; row-scoping would pair them.
- **Guarded default-branch resolution** — `git symbolic-ref refs/remotes/origin/HEAD` is not
  created by `actions/checkout`, so under `set -euo pipefail` an unguarded call exits non-zero in
  CI and is indistinguishable from a real violation.
- `pull`/`checkout` added to the read-verb exclusions; `rather than`/`instead of` **removed** from
  the prohibition-marker set (directional, not prohibitive → fail-open).

**Ordering keeps `main` green while staying test-first**: the gate enters CI only *after* the
violations are removed (sub-epic C follows B), so **no advisory toggle is needed** — strictly
safer and simpler than the bounded-advisory-window pattern prior art required.

## 7. Backlog structure created

- Feature **`018-F`** — top-level release unit (chore-classified via label; backlogit exposes no
  `chore` type)
- Tasks **`018.004-T` … `018.011-T`** — 8 atomic units
- Shipment **`017-S`** — queued, 9 explicit items

**Structural finding**: this workspace's backlogit schema stores `size` on the `task` type only
and defines **no `complexity` field on any type** (`backlogit metadata types` → task fields =
`priority`, `size`, `status`), despite `backlog-registry.yaml` declaring `features.sizing: true`.
Consequences, both handled:

1. A first-pass feature → sub-epic(`task`) → task(`subtask`) hierarchy could not carry sizing at
   the atomic level. The 11 provisional items (`018.001-T`–`018.003-T`,
   `018.001.001-ST`–`018.003.002-ST`) were **archived, not deleted** (no destructive operation was
   authorized) and the 8 atomic units were recreated as `task` type. Sub-epic grouping is now
   carried on `sub-epic-a/b/c` labels plus the `blocks` dependency graph.
2. `complexity` is recorded as **enum-validated prose** in each task's `implementation-notes`,
   per the Stage two-axis sizing contract's degradation path, with the degradation flagged
   in-line on every task.

**Known benign artifact**: shipment `017-S`'s derived `size_composition` roll-up reports
`unsized: 3` and 11 members, because it walks all children of `018-F` including the three archived
provisional sub-epic containers. The authoritative membership is the explicit `items` array
(9 IDs). This is documented in `018-F`'s goals section so Ship is not misled.

## 8. Review remediation — PR #54, cycle 1

Copilot reviewed `faf828b` and opened five threads; the review summary surfaced further
same-contract findings. All were remediated under **P-021 C1** (they complete the exact authorized
planning contract rather than expanding it) at PR head `c16f673`. No new shipment, no new task,
no source/template/CI file touched.

| Thread | Fix |
|---|---|
| `PRRT_kwDOTPuhps6hk-dz` | AC-A1.2 relocated from `018.004-T` (data-only) to `018.005-T` as **AC-A2.4** (harness owner). A1 keeps a static prefix/verdict-agreement check. |
| `PRRT_kwDOTPuhps6hk-ee` | Stale `check-direct-push-policy.sh` → `check-direct-push-language.sh` (deliberation L118, L276). |
| `PRRT_kwDOTPuhps6hk-fA` | "two CI steps" → **one blocking `lint` step** (deliberation L127, L238, L277). |
| `PRRT_kwDOTPuhps6hk-fa` | Repository-standard YAML frontmatter added to this file. |
| `PRRT_kwDOTPuhps6hk-f5` | B3 Role Boundary **Allowed** cell now carries the branch create/check-out grant, matching B2's P-010 grant; `018.009-T` aligned and retitled. |

**Suppressed-but-actionable findings, also fixed**: the eight
`System.Collections.Hashtable[018.xxx-T]` machine-readable AC placeholders (a PowerShell
interpolation defect) replaced with real criteria; `018.009-T`'s PR-row contradiction resolved by
explicitly **exempting** the PR row in AC-B3.3; deliberation scope widened to the four surfaces and
the now-vacuous separate testdata/self-exclusion assertion withdrawn in favour of the authoritative
pathspec; plan "all three surfaces" → **four**; the chained `gofmt && go vet && go test && go build`
recipe split into separate commands per terminal policy; the concrete no-shipment persistence route
documented with per-step actors; and this file's incorrect claim that Step 1.5 must use the
staging-branch form **instead of** `origin/main` withdrawn (see §9 — both checks stand).

**Dependency integrity after the AC move**: unchanged. The `018.004-T → 018.005-T` edge already
guaranteed fixtures land before the harness executes them, so relocating an execution criterion
onto the downstream task required no graph edit. `backlogit doctor` reports no issues.

**Tool hazard recorded**: `backlogit update --description` **destroys** template-backed body
sections (`implementation-notes`, `acceptance-criteria`) rather than preserving them. Only
`--section name=value` is body-preserving. The three tasks needing description edits had both
sections re-applied immediately; all eight tasks were then verified to carry intact, populated
blocks. Use `--section` for body content.

**Validation run** (planning-artifact scope only; no Go build/test, per the Stage role boundary):

- `backlogit doctor` → `No issues found.` (exit 0)
- `markdownlint` on the three changed docs → exit 0
- `backlogit docs lint` → 172 violations before **and** after, byte-identical findings — a
  pre-existing corpus-wide `doc_type`/`source` gap unrelated to this change, deliberately **not**
  fixed here to avoid scope creep. Flagged as a follow-up candidate.
- `backlogit list --type shipment` → exactly one shipment, `017-S`, 9 items unchanged.

## 9. Next actor

**Ship** — claim shipment `017-S` and execute `018.004-T` first (the failing regression harness).
Do not start at `018.007-T`; the `blocks` edges enforce harness-before-correction.

Step 1.5's verification uses **two distinct checks**; the second is **not** replaced by the first.

1. **Pre-merge, local evidence only** — confirm the manifest is committed on the staging branch:

   ```sh
   git show chore/stage-pipeline-policy-gap:.backlogit/queue/017-S.md
   ```

   This proves the artifact was committed. It proves **nothing** about the default branch.

2. **Post-merge, the authoritative gate — UNCHANGED and still required**:

   ```sh
   git show origin/main:.backlogit/queue/017-S.md
   ```

   Step 1.5 step 4 keeps this form byte-identical, and `STAGING_GATE_FAIL` still halts on it
   (plan task B1, AC-B1.4).

**Correction (review round 3)**: an earlier revision of this file claimed the manifest
verification "must use the staging-branch form, **not** the `origin/main` form". That was wrong
and is withdrawn. This work changes **where Stage may write** (a dedicated branch instead of the
default branch); it does **not** relax **where the Orchestrator must verify**. The post-merge
`origin/main` gate is the reason the branch route is safe, not a casualty of it. Both checks
stand, in the order above.

Push, PR creation, and merge between checks 1 and 2 are performed by the **Orchestrator**
(pipeline invocation) or the **operator** (direct Stage invocation) — never by Stage.

## 10. Review remediation — PR #54, cycle 2

Copilot's current-HEAD review of `a6cf863` opened two unresolved threads. Both are **P-021 C1
same-contract blockers** — they complete the authorized planning contract rather than expanding it
— so both were **fixed, not deferred**. No source, template, or CI file was touched; no shipment
was claimed; exactly one shipment (`017-S`, 9 items) still exists.

### 10.1 Thread `PRRT_kwDOTPuhps6hl8My` — the substitute harness was unexecutable

**Finding**: plan §10.1 (rev 3) declared a P-002/P-004 deviation and asked Ship to apply
`harness-ready` from a prose label over a bash `--self-test` suite. Ship cannot honour that.
`_ship.agent.md` Step 2 partitions the queue on the `harness-ready` label, invokes
`harness-architect` for every task lacking it, and halts unless *every* queued task carries it;
`harness-architect` emits only Go `_test.go` harnesses and applies the label only after
`go vet ./...` exits 0 and `go test ./...` is red. A Stage-applied label forges P-004's
postcondition; withholding it drives `harness-architect` to invent Go scaffolding for markdown and
bash tasks.

**Investigation performed before choosing**: `_ship.agent.md` Step 2 (L313–330) and Step 3 queue
derivation; `harness-architect/SKILL.md` Steps 4–6 and Guardrails; `build-feature/SKILL.md`
(`harness_cmd` is generic, so a non-`go test` command is already supported downstream);
`workflow-policies.md` P-002 (L36–52) and P-004 (L78–92); and the shell/script-only precedent —
shipment `015-S` shipped 13 bash/Python gate-script tasks with **no** `harness-ready` label on any
of them and no recorded P-002/P-004 disposition, i.e. an undocumented de facto bypass rather than
a contract Stage may cite.

**Resolution — restructure the release unit onto the *existing* harness path (no policy change)**:

- New plan **§5.0** defines the **H0 harness contract**: `harness-architect` produces
  `tests/integration/directpush_gate_test.go` (eight test functions, one per task) plus
  `scripts/check-direct-push-language.sh` as a **not-implemented structural stub** — the shell
  analogue of `panic("not implemented: <reason>")`. `go vet ./...` exits 0, `go test ./...` is red,
  label applied by the skill in the normal way.
- **Not invented scaffolding**: `tests/integration/build_script_test.go`, `start_script_test.go`,
  and `output_path_guard_test.go` are existing tests here whose whole subject is a non-Go script,
  and `internal/apperr/taxonomy_drift_test.go` is an existing Go drift guard over a non-runtime
  invariant. Go-test-over-non-Go-artifact is this repository's established pattern.
- Each task records its **harness function and `harness_cmd`**, giving `build-feature` a one-task
  boundary with exactly one observable red→green transition. B1/B2/B3 are **per-surface scoped**;
  a whole-corpus assertion would leave the first two red and trip the 5-attempt circuit breaker.
- **§5.5 (A2) retargeted to driver-only.** A task whose content is "create the harness" is
  structurally unreachable in Ship's flow, because Step 2 runs once up front for the whole queue.
  Retitled *Implement fixture-suite driver over stubbed detector*.
- **Fail-never-skip** on missing `bash` (§5.0), because a `t.Skip` would make the red phase
  unobservable and silently satisfy P-004 (P-012).
- §10.1 rewritten: **no deviation is declared**. §10 Constitution Check rows I and II updated;
  the stale "7 tasks" count corrected to 8.

**Explicitly rejected — amending the governing workflow.** A first-class non-Go harness route
needs edits to P-002/P-004, `_ship.agent.md` Step 2, and `harness-architect`. The latter two are
autoharness-generated from **gitignored** `.tmpl` sources (D2), so the change would be reverted by
the next render — failure mode R4/R5 — and amending P-002/P-004 is a **workspace-wide weakening of
test-first** binding every future shipment, outside both this release unit's branch-policy contract
and the governing deliberation. Recorded in plan §10.1 and added to §11 **Out of scope**.

### 10.2 Thread `PRRT_kwDOTPuhps6hl8Nd` — C2 contract contradiction

**Finding**: `018.011-T`'s AC mandated a **dirty-tree** reintroduction proof while plan §7.2 C2.2
mandated a **fixture-backed** one, and the two used different C2 numbering — so the PR's
"all 8 task contracts match" claim was false.

**Resolution**: plan §7.2 rewritten as the single C2 verification contract, renumbered
**AC-C2.1…AC-C2.7**, with the fixture-backed proof made explicit (both violation shapes are locked
by committed fixtures: `directpush-reject-attempt-first.md` for the push shape,
`directpush-reject-commit-on-default.md` for the commit-authorization shape), an explicit
**no-dirty-tree-edit** rule, and a `git status --porcelain` empty check before and after. The
dirty-tree form is **rejected inside the criterion text** so it cannot return by paraphrase.
`018.011-T` now carries that block **byte-identical**; its plain-text rendering is explicitly
marked non-authoritative. AC-A2.4's non-vacuity proof was aligned to the same rule (temporary
fixture copy under `t.TempDir()`).

### 10.3 Third instance found by verification sweep, not by review

An AC-ID parity sweep across all eight tasks found the same defect class again: plan §6.1's B1
criteria still carried the three-criterion rev-2 form while `018.007-T` carried the four-criterion
rev-3 form covering the preamble surface, and §6.3 already referenced a then-nonexistent
**AC-B1.4**. §6.1 corrected to the task's form.

### 10.4 Validation (planning-artifact scope only — no Go build/test, per the Stage role boundary)

- `backlogit sync` → `INDEX_SYNC_OK`, **220** artifacts at cycle-2 start and end — the
  post-Step-5.6 steady state (207 items + 13 active stash entries), unchanged from the harvest
  session's true end state. This does **not** conflict with §2's 221, which is the
  pre-stash-archive transient peak; see the phase/timestamp table in §2.
- `backlogit doctor` → `No issues found.` (exit 0)
- `markdownlint` on the plan and on `.backlogit/queue/*.md` → exit 0
- **Harness eligibility**: 0 of 8 tasks carry `harness-ready` (Stage forges no P-004 postcondition);
  `018-F.custom_fields.harness_status` remains `pending`
- **Harness-map bijection**: 8 plan functions ↔ 8 task functions, no orphan on either side
- **AC-ID parity**: 38 plan IDs ↔ 38 task IDs, exact bijection both directions
- **C2 verbatim check**: all 7 criteria byte-identical between plan §7.2 and `018.011-T`
- **Dependency graph**: `018.004 → 018.005 → 018.006 → {018.007, 018.008, 018.009} → 018.010 →
  018.011`, matching plan §4 and the per-task harness transitions
- **Scope**: exactly one shipment (`017-S`), 9 items, membership unchanged
- Stale-text sweep: no `Declared deviation`, `Substitute harness`, `label the harness tasks`,
  `No Go surface is touched`, or dirty-tree mutation language survives anywhere

### 10.5 Handoff delta for Ship

Unchanged: claim `017-S`, execute `018.004-T` first. **New**: Ship's Step 2 now has a real,
specified harness to build (plan §5.0) — produce `tests/integration/directpush_gate_test.go` and
the not-implemented stub script, confirm `go vet ./...` = 0 and `go test ./...` red, record
`Compilation: PASS` / `Red Phase: CONFIRMED`, and only then apply `harness-ready` and flip
`018-F.harness_status`. **Do not** amend P-002, P-004, `_ship.agent.md` Step 2, or the
`harness-architect` skill — explicitly out of scope.

## 11. Review remediation — PR #54, cycle 3 (final allowed cycle)

Copilot's current-HEAD review of `47ede9a` left exactly one unresolved thread
(`PRRT_kwDOTPuhps6hm1OQ`, comment 3992616381): the audit counts in this file conflicted — §2
reported 199 → 221 artifacts while §10.4 reported 220 at both start and end, and the PR
description reported 220. A **P-021 C1 same-contract** defect (it corrects the session record's
own accuracy rather than expanding scope), so it was fixed, not deferred. No source, template, or
CI file was touched; no shipment was claimed; exactly one shipment (`017-S`, 9 items) still exists.

### 11.1 Root cause — three distinct snapshots flattened into two labels

The counts were never wrong; they were **unlabelled**, so two of the three snapshots read as a
contradiction. The index total is `items + active stash_entries`, which means **stash intake and
stash archival move the number independently of harvest**. §2's "session end" label was applied to
a reading actually taken *before* Step 5.6 archived the intake stash entry, while §10.4's 220 was
the genuine post-Step-5.6 steady state.

### 11.2 Evidence used (no number was inferred or guessed)

| Claim | Evidence |
|---|---|
| Total = items + active stash entries | `.backlogit/backlogit.db` read-only: `items` = 207, `stash_entries` = 13, and `backlogit sync` → `Indexed 220 artifacts` |
| Pre-session items = 186 | `git ls-tree -r --name-only faf828b^ -- .backlogit/queue .backlogit/archive` → 0 + 186 `.md` |
| Post-harvest items = 207 | same command at `faf828b`/`d6c086c`/`47ede9a` → 10 + 197 `.md`; the 21-item delta matches the 21 IDs created in this session |
| Pre-session active stash = 13 | `git show faf828b^:.backlogit/stash.jsonl` → 13 non-blank lines |
| 199 baseline independently corroborated | `docs/memory/2026-09-11-stage-pathsafe-deferred-trigger-reassessment-no-shipment.md` L24 records `INDEX_SYNC_OK — 199 artifacts` for the immediately preceding session |
| The 14th stash entry is `638A410B` | `.backlogit/archive/stash.jsonl` → `created_at 2026-09-11T16:51:55Z`, `archived_at 2026-09-11T17:27:26Z`, `reason archived` |
| The 221 window is `17:26:03Z`–`17:27:26Z` | `items.created_at` for `017-S` = `2026-09-11T17:26:03Z`; `638A410B.archived_at` = `17:27:26Z` |
| 220 is current and stable | `backlogit sync` re-run at cycle 3 → `Indexed 220 artifacts` |

### 11.3 Fix applied

§2 replaced the two-label sentence with a **phase + UTC-timestamp table** carrying all three
snapshots and their `items` / `active stash` decomposition, plus an explicit note that 220 is the
steady state the PR description reports. §10.4's bullet now states the phase its 220 belongs to
and cross-references §2 so the two entries can no longer be read as rival claims about one state.
No count was changed; only the phase attribution that made them look contradictory.

### 11.4 Validation (planning-artifact scope only — no Go build/test, per the Stage role boundary)

- `backlogit sync` → `INDEX_SYNC_OK`, **220** artifacts (207 items + 13 active stash entries) —
  the figure now recorded as the steady state
- `backlogit doctor` → `No issues found.` (exit 0)
- `markdownlint` on this file → exit 0
- `backlogit list --type shipment` → exactly one shipment, `017-S`, 9 items, membership unchanged
- Count sweep: `199`, `220`, and `221` now appear in this file only inside phase-labelled context;
  no unlabelled "session end" count survives
- Scope: only `docs/memory/` was modified — no source, template, CI, plan, deliberation, or
  `.backlogit/` queue artifact changed

### 11.5 Handoff

Unchanged from §10.5. Ship claims `017-S` and starts at `018.004-T`.


---

## 12. PR #54 review remediation cycle 4 — rev 5 (operator-authorized fourth cycle)

**Authorization**: operator explicitly authorized ONE additional Stage review-fix cycle beyond the
three-cycle cap, scoped **only** to thread `PRRT_kwDOTPuhps6hnPkD`. No other scope reopened.
**Branch**: `chore/stage-pipeline-policy-gap`, base HEAD `7b5d3e3`. No push, no PR action, no
shipment claim, no branch switch, no source/template/CI mutation.

### 12.1 Finding

Copilot comment 3992775719 (plan L720): plan **R5** accepted the residual that a future autoharness
render could remove the job-`lint` gate step, leaving CI green with the detector never invoked.
That contradicted **§1** ("enforce it mechanically so regeneration ... cannot **silently** reopen
the gap") and the **018-F DoD**. The reviewer required either a tracked enforcement path surviving
regeneration, or an honest narrowing of the objective.

### 12.2 Investigation — render boundary established empirically

| Question | Evidence |
|---|---|
| Is `ci.yml` generated? | Yes — header `Generated by autoharness \| Template: ci/ci.yml.tmpl`; the gate step would be a LOCAL DIVERGENCE alongside five already enumerated |
| Which `scripts/*.sh` are generated? | Only 5 of 10 (`ci-topology-check`, `deploy-harness`, `pre-commit-markdownlint`, `pre-commit-pipeline-topology`, `pre-push-quality-gates`). The 5 gate scripts incl. the planned detector are **hand-authored** |
| Is `tests/**` generated? | No autoharness provenance anywhere under `tests/` |
| Is the `go test` invocation generated-baseline or divergence? | Job `expensive` (`name: test`), step `Test (race)` → `go test -race -mod=readonly ./...`. Listed in the header's **core four-job shape**, carries no task-ID comment, absent from the LOCAL DIVERGENCE list → **baseline; a render re-emits it** |
| Can a render skip that job? | No. `changes.code` is a **denylist** (`'**'` minus `docs/**`, `.backlogit/**`, `.backlog/**`, `.autoharness/**`); rewriting `ci.yml` is a `.github/**` change → `code == 'true'` |

**Alternatives rejected**: `secret-scan-history.yml` is non-generated and would survive, but its own
header declares it **NOT A PR-REQUIRED CHECK** (schedule/dispatch only) — cannot carry a blocking
gate without inverting its contract. `ci-topology-check.sh` is itself generated → same boundary.

### 12.3 Resolution — objective preserved, not narrowed

The regeneration-resistant path **already existed** in the release unit but was never claimed: the
§5.0 Go harness is non-generated and runs from a generated-baseline invocation. Rev 5 makes it
explicit, load-bearing, and self-verifying. Two criteria added (no new file/workflow/CI step/ledger
entry/CODEOWNERS line):

- **AC-C1.6** (018.010-T) — requirement **(b)**: step-presence assertion is **YAML-structural** over
  `jobs.lint.steps[]`, never a text grep (which would self-match C1's own ledger comment naming the
  script — the failure mode already recorded in the 2026-09-06 compound entry), with non-vacuity
  proven by execution against a `t.TempDir()` copy having the step removed. No dirty-tree edit.
- **AC-C2.8** (018.011-T) — requirement **(a)**: the full-corpus detector scan runs from
  `go test ./...`, so it keeps executing with the job-`lint` step gone.

R5 reclassified **accepted → mitigated**. AC-C1.5 reworded from "accepted residual" to "detected,
not accepted"; AC-C1.3 extended so the ledger records the same; §7.1 ledger prose updated.

### 12.4 Honest residual — no fabricated guarantee

The guarantee claimed is exactly §1's: *regeneration cannot **silently** reopen the gap*. It is
**not** a claim the control is irremovable. Deleting the harness remains possible via a reviewable
PR diff to a non-generated tracked file — recorded as new risk **R12**, explicitly not closed.
`CODEOWNERS` is **deliberately excluded** from the guarantee: its own header states it has no
enforcement effect until code-owner review is required on `main`, which is not enabled. R11 updated
to note both harness properties are now load-bearing rather than incidental.

### 12.5 Validation (Stage scope — no Go build/test)

- `backlogit sync` → `INDEX_SYNC_OK`, 220 artifacts (unchanged steady state)
- `backlogit doctor` → `No issues found.` (exit 0)
- `markdownlint` → exit 0 on all four changed files; `scripts/pre-commit-markdownlint.sh` → exit 0
- **AC-ID parity**: 40 plan IDs ↔ 40 task IDs, exact bijection, no orphan either side (was 38 ↔ 38)
- **Plan↔task verbatim parity**: plan §7.2 block byte-compared (`-ceq`) against `018.011-T`
  acceptance-criteria section → **PASS**, single-verification-contract rule preserved
- **Harness mapping**: all 8 task harness functions resolve into plan §5.0
- **Dependency order unchanged**: A1→A2→A3→{B1,B2,B3}→C1→C2
- **Shipment**: `017-S` still one shipment, 9 items, membership unchanged
- `scripts/pre-commit-pipeline-topology.sh` → **TOOL_DEGRADED** (P-012): `autoharness` not on PATH
  locally, gate self-skipped; recorded rather than silently passed
- Sizes/complexity unchanged (S/medium on both tasks) — rationale recorded in each task

### 12.6 Handoff

Unchanged. Ship claims `017-S` and starts at `018.004-T`. H0 (`harness-architect`) must now produce
`TestDirectPushGate_CIWiringIsBlocking` satisfying AC-C1.6's structural + non-vacuity shape.

---

## 13. PR #54 review remediation cycle 5 — rev 6 (operator-authorized fifth cycle)

**Scope**: thread `PRRT_kwDOTPuhps6hqWDe` (comment 3993998429, plan L517) ONLY. Thread
`PRRT_kwDOTPuhps6hqWD8` is an Orchestrator-owned incident-record correction and was deliberately
NOT touched. No push, no PR, no branch switch, no shipment claim, no source/template/CI edit.
Branch `chore/stage-pipeline-policy-gap`, single worktree, clean tree at start.

### 13.1 Finding

Step 1.5 discovers work through only two inputs — `git status --short -- .backlogit/` and a
default-branch-only commit range (`_orchestrator.agent.md:274-281`). Neither inspects the
checked-out branch or HEAD. Once B2/B3 require Stage to commit on `chore/stage-{chore-slug}` and
leave the worktree there, both inputs read "synchronized": the commit is not on the default branch,
and a no-shipment run's artifacts are not under `.backlogit/`. The gate skips step 3 and then fails
its `origin/main` manifest check — and a shipment-less run has no manifest to check at all. The
no-shipment route documented end to end in plan §6.2 was therefore **unexecutable as written**.

### 13.2 Root cause — the frame, not the wording

Revisions 1–5 all worked inside an **authorization** frame: find text that permits a Stage artifact
to reach the default branch without a PR, and correct it. That frame grew from two surfaces to four
and was internally consistent — but correcting authorization only *relocates* the commit. It never
told the gate that must verify the commit where to look. This is the §8.2 self-deadlock lesson one
layer down: a rule is sound only when the actor who must **verify** it can **observe** what it
verifies. Plan rev 2 had even declared Step 1.5 out of scope for shipment-less runs while §6.2's
own route table assigned the Orchestrator the push/PR/merge/verify duty for exactly that path —
a duty with no remaining home. That position is now explicitly withdrawn.

### 13.3 Resolution — four corrections, no scope expansion

1. **Explicit handback** — Orchestrator Step 1 requires `stage_branch`, `stage_head_commit`,
   `stage_outcome`, `stage_artifact_paths` (+ `shipment_id` when formed); Stage's Step 6 emits them.
   **Every** field has a deterministic default, so the degraded path **never halts**: branch/commit
   from `HEAD`, outcome from whether a `shipment_id` came back, paths from the step-1 default set,
   `STAGING_HANDBACK_DEGRADED` recorded once and surfaced. When the `HEAD` resolution yields the
   default branch, `stage_branch` stays **UNRESOLVED**, no branch assertion is applied, and steps 1
   and 2 still run unchanged — so any work they find routes to 3a's create-arm, which is
   exactly the pre-change behaviour — the default branch is never substituted and 3a's create-arm
   stays live rather than dead prescription. Defaults-with-degradation rather than hard halt,
   deliberately: a hard halt would wedge every Stage invocation predating B3 — the rev-2 lesson
   applied preemptively.
2. **Resolution hoisted above the dirty/clean split, dirtiness pathspec widened** — `$STAGE_BRANCH`
   resolution, `git rev-parse --verify`, and the HEAD-equality assertion (P-016 single-worktree
   gate) sit **above** step 1, because step 1 routes a dirty tree straight to step 3 and anything
   in step 2 is skipped on the very path that mutates the repo. The pathspec widens to the Stage
   artifact paths (default `.backlogit/ docs/plans/ docs/decisions/ docs/memory/`), so a docs-only
   no-shipment run is not reported clean.
3. **Branch-specific discovery** — step 2 compares
   `origin/main..HEAD`. The two range forms are provably the same range *because* of the hoisted
   HEAD assertion.
4. **Second verification arm** — `stage_outcome: no-shipment` verifies
   `git merge-base --is-ancestor {stage_head_commit} origin/main` plus `git show origin/main:{path}`
   over a **non-empty** resolved path set, halts with `STAGING_GATE_FAIL` on an empty set or either
   failure, and terminates without routing to Ship. Ancestry proves the commit landed; per-path
   `show` proves the artifacts are readable there rather than merely that some merge happened; the
   non-empty guard stops a zero-run loop from passing the gate having verified nothing. No
   `shipment_id` is named.

Actor sequence preserved without deadlock: **Stage** creates/checks out/commits and returns
branch+commit+outcome; **Orchestrator** pushes, opens the PR, waits for merge, verifies on
`origin/main`. Every command added is a read form (`status`, `rev-parse`, `fetch`, `log`, `show`,
`merge-base`) or a `checkout` — no direct default-branch push, P-009/P-011/P-016 textually intact.

### 13.4 Self-match hazard caught during drafting, not by review

The natural phrasing of correction 3 is a prohibition quoting the literal forbidden range. But
`_orchestrator.agent.md` **is** in the scanned corpus, and the B1 harness asserts that literal is
**absent** from Step 1.5 — so a prohibition quoting it would match itself and vacate the very
assertion that rejects the regression. Same failure class as
`docs/compound/2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md`, already
guarded at plan §5.0 and AC-C2.3. The prohibition is stated **descriptively**, and AC-B1.6 makes
that contractual rather than incidental. Worth a compound entry if it recurs a third time.

### 13.5 Deliberate non-expansions (the finding did not require them)

- **Detector NOT widened.** Main-only discovery is neither §5.3 violation construct. Adding it
  would open the closed construct set, add fixtures, and re-derive A1's 14/9/5 counts through
  A1→A2→A3→C2. Rejection is carried by the **existing** B1 harness function instead — already
  scoped to `_orchestrator.agent.md`, already B1's red→green boundary.
- **Fixture corpus unchanged.** `directpush-accept-read-and-pr-forms.md` keeps `origin/main..main`
  as an **accept** case: it proves the detector does not flag read verbs, which is independent of
  what the Orchestrator uses.
- **No new harness function.** A second function on the same file would not give an independent
  red→green transition (plan §5.0), so the assertions were added to the existing B1 function.
- **No new task.** Splitting B1 would add a 9th task whose transition could not be observed
  independently, and would perturb 017-S membership.
- **P-010 text unchanged.** The handback is a mechanism, not a permission — 018.008-T's ACs are
  untouched, with only a cross-reference note added.

### 13.6 AC-B1.4 narrowed, on the record

AC-B1.4 previously required step 4's whole block to be byte-identical. That is unsatisfiable once
step 4 must carry the no-shipment arm — and an unmodified block is exactly what left that path
unverified. It is narrowed to what it always protected: the shipment arm's lead sentence, `git show`
command, both bullets, and `STAGING_GATE_FAIL` string preserved verbatim with unchanged semantics;
the step may gain arm labels and the second arm, and nothing else. Recorded in the plan, the task,
and the feature DoD rather than applied silently — a weakened criterion with no stated reason is
indistinguishable from an abandoned one.

### 13.6b Four defects found by internal review before commit

The first draft of the rev-6 corrections was reviewed against the finding and four high-confidence
problems were fixed rather than shipped. Recorded because a cycle that silently repairs its own
draft teaches nothing:

1. **A halt that contradicted its own rationale.** The degraded handback halted when the HEAD
   resolution yielded the default branch — wedging every pre-B3 Stage invocation, the exact
   deadlock the paragraph three lines below claimed to avoid — and put that halt UPSTREAM of 3a's
   create-arm, making it dead prescription. Fixed: leave the branch UNRESOLVED and route to 3a.
2. **Partial defaults and a vacuous verification loop.** `stage_outcome` and
   `stage_artifact_paths` had no degraded default, leaving step 4's arm selection undefined; and
   an empty path set made the no-shipment arm pass having verified nothing (zero-run `show` loop
   plus a trivially-true ancestry check). Fixed: total defaults, empty path set is a hard fail.
3. **A guard placed on the unreached branch of a fork.** Step 1 routes a dirty tree straight to
   step 3, so the new verify/HEAD-equality guards sitting in step 2 were skipped on the one path
   that mutates the repo — falsifying AC-B1.9 as written. Fixed: resolution block hoisted ABOVE
   step 1.
4. **Prescribed text that could not satisfy its own AC.** An outer lead sentence replaced step 4's
   pre-change lead, and the original reappeared lowercased, failing AC-B1.4's verbatim clause.
   Fixed: arm label on its own line above the byte-verbatim original; AC-B1.4 now forbids an outer
   lead that replaces or rewords it.

The recurring shapes worth recognizing: a fallback that halts on the population it exists to
protect, a verification loop that can run zero times, and a guard installed downstream of the fork
that skips it.

### 13.7 Surfaces reconciled

| Surface | Change |
|---|---|
| Plan (rev 5→6) | Front matter; rev-6 header; §3 surface rows B1/B3; §5.0 harness rows B1/B3; §5.4 no-widening note; §6.1 restructured into §6.1.1/§6.1.2; §6.2 handoff note; §6.3 + AC-B3.5; §7.2 AC-C2.4; §11 R13; §12 rev-6 record |
| `018.007-T` (B1) | Title, description, 5 new ACs (B1.5–B1.9), B1.2 extended, B1.4 narrowed, size S→M, rev-6 reconciliation note |
| `018.009-T` (B3) | Title, description, AC-B3.5 added, AC-B3.3 scoped to the Role Boundary table, size XS→S, rev-6 note |
| `018.011-T` (C2) | AC-C2.4 extended in both renderings (handback contract agreement); rev-6 reconciliation note |
| `018.008-T` (B2) | Cross-reference note only — no AC, scope, or size change |
| `018-F` | DoD: no-shipment route executability, self-match discipline, AC-B1.4 narrowing; execution-order line |
| Deliberation | §10 Revision 6 addendum — frame extended from authorization to authorization+discovery |
| `017-S` | **Unchanged** — one shipment, 9 items, order A1→A2→A3→{B1,B2,B3}→C1→C2 |

### 13.8 Validation (Stage scope — no Go build/test, per the Stage role boundary)

AC-ID parity re-swept: **46 plan IDs ↔ 46 task IDs**, exact bijection (was 40↔40). Dependency
edges unchanged. Markdownlint clean on all changed markdown. Backlog sync + doctor clean.
Go build/test deliberately NOT run — Ship's responsibility (P-010).

### 13.9 Handoff

Unchanged: shipment **017-S**, `queued`, 9 items. Ship starts at Step 2 (harness-architect, H0),
which now produces a B1 function carrying the branch-discovery assertions and a B3 function
carrying the handback-emission assertion. Cycle budget for PR #54 is exhausted unless the operator
authorizes another.
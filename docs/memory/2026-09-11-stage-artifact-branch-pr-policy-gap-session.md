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
| `docs/decisions/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-deliberation.md` | Source document / P-003 lineage root (§8 rev-2 addendum, §9 rev-3 addendum, §10 rev-6 addendum, §11 rev-7 addendum, §12 rev-8 addendum, §13 rev-9 addendum, §14 rev-10 addendum) |
| `docs/plans/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-plan.md` | Implementation plan, **revision 10**, `status: reviewed` (§5.0.1 per-task activation, §9 Plan hardening signals, §10 Constitution Check, §12 plan review record) |
| `docs/memory/2026-09-11-stage-artifact-branch-pr-policy-gap-session.md` | This file |

**Cross-reference currency (rev 7, re-verified rev 8).** This table names the plan's **current**
revision and only sections that exist in it. The `§10a` label it previously carried was a rev-3
heading that later revisions renamed to **§9 Plan hardening signals**; the plan's own front matter
still carried the stale `§9/§10a` form and §5.3 still cited a nonexistent `§10a.1`, both corrected
in rev 7. A session record that points at an obsolete revision and a nonexistent section is not a
cosmetic defect — it is the artifact a future agent reads *first* to locate the authoritative
contract. Re-swept at rev 8 and again at rev 10: every `§N` reference in this file resolves against the plan's and the
deliberation's current heading lists, including the plan's new **§5.0.1** and the deliberation's new **§14**.

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

## 14. PR #54 review remediation cycle 6 — rev 7 (operator-authorized sixth cycle)

Operator authorized ONE bounded cycle scoped to the two then-unresolved threads, with an explicit
instruction not to pre-emptively expand scope. Both were fixed; nothing else was touched.

### 14.1 Findings

| Thread | Comment | Target | Finding |
|---|---|---|---|
| `PRRT_kwDOTPuhps6hrC0c` | 3994280334 | plan L626 | The no-shipment arm iterates `git show origin/main:{path}`, but the documented default path set is **directory prefixes**. Those directories already exist on `origin/main`, so the loop can succeed without proving this run's files persisted; combined with ancestry of a fallback `HEAD`, it can report a **false pass** |
| `PRRT_kwDOTPuhps6hrC0o` | 3994280351 | memory L59 | The artifacts table is stale — the plan is revision 6, and its hardening section is **§9**, not §10a. The record points readers at an obsolete revision and a nonexistent section |

### 14.2 Both findings confirmed by execution, not by reading

Neither was accepted on the reviewer's word. Both were reproduced against this repository before
any text changed:

* `git show origin/main:docs/plans/` → prints `tree origin/main:docs/plans/`, **exit 0**. The
  directory-prefix pass is real, not theoretical.
* `git cat-file -t origin/main:docs/plans/` → `tree`, **exit 0**; on a file → `blob`, exit 0; on an
  absent path → exit 128. So the exit code alone cannot discriminate a file from a directory — the
  **printed type** is the discriminator. This is why the arm switches command rather than adding a
  guard around `show`.
* `git diff-tree --no-commit-id --name-only -r --diff-filter=d <sha>` → blob paths only, zero
  entries with a trailing `/`, and deletions excluded (confirmed on this branch's `HEAD`, which
  renames a memory file: the deleted path is correctly absent from the derived set).
* `git log origin/main..<ancestor-of-origin/main>` → **0 commits**. This is why the derived set is
  taken from the commit's own changes rather than from a branch range: post-merge, the range form
  is empty by construction and would re-create the vacuous loop.
* Plan headings enumerated: the file has `## 9. Plan hardening signals` and `## 10. Constitution
  Check`; **no `§10a` exists**. Front matter read `revision: 6` at the time of this investigation
  (it reads `revision: 7` after this cycle's correction).

### 14.3 Resolution — structural rejection, not a bigger checklist

> **Superseded in part at rev 8 — see §15.1.** The concrete-file and `cat-file -t` decisions below
> **stand**. The scoping of the set to `stage_head_commit` alone does **not**: rev 8 replaced it
> with the aggregate `{stage_base_commit}`→`{stage_head_commit}` two-tree diff, and the arm now runs
> **six** ordered checks, not five. This block is retained as the rev-7 record, not as the current
> contract.

`stage_artifact_paths` is redefined as a **non-empty set of concrete repository-relative file paths
tied to `stage_head_commit`**, never a directory list. The no-shipment arm became five ordered
checks: commit existence, **ancestry**, derivation of a non-empty changed-file set, `VERIFY_SET`
resolution, and a per-file `git cat-file -t` requiring exit 0 **and** output exactly `blob`
(**readability + file-ness**). Directories, empty sets, out-of-root paths, and any mismatch with the
commit's changed-file set are all rejected.

Three deliberate choices, each rejecting a weaker alternative:

1. **`cat-file -t` replaces `show` in this arm.** Both exit 0 on a tree; only `cat-file -t` names
   the object type. Parsing `show`'s human-readable listing would be fragile in exactly the way a
   gate must not be.
2. **Set equality, not subset.** A subset check rejects an invented path but still accepts a
   handback that shrinks the verified set to one always-present file — the same defect through a
   different door.
3. **The commit's own changed files, not a branch range.** The range form is empty by construction
   post-merge (measured above). The tip-commit scope is sound because ancestry already proves every
   earlier commit reached `origin/main`; the per-file check closes the different gap.

Two second-order defects were found while drafting and fixed before commit, both of the
"correct rule installed on the unreached path" family this file has now recorded three times:

* **The handback commit predates the commit that carries the artifacts.** On the dirty-tree path,
  3a *creates* the artifact commit, but `stage_head_commit` was resolved before it existed —
  so the derivation would have read a commit that changed none of them and halted for the wrong
  reason on the one path Step 1.5 exists to repair. Fixed by correction 5: 3a re-records
  `stage_head_commit` and discards the stale reported paths.
* **`stage_artifact_paths` was doing two incompatible jobs.** It was both step 1's dirtiness
  pathspec (where directory prefixes are *correct* — step 1 runs before the commit exists) and
  step 4's verification target (where they are fatal). That conflation is *why* directories reached
  a loop with no way to reject them. Split: `STAGE_ARTIFACT_ROOTS` is now a **fixed constant** for
  the pathspec and the containment allow-list; `stage_artifact_paths` is always concrete files.
  Pinning the pathspec to the constant also closes a hole one step earlier — a handback naming one
  narrow path could previously have shrunk the dirtiness scan and hidden uncommitted artifacts.

### 14.4 Invariants explicitly preserved

The shipment arm is untouched and stays byte-verbatim (AC-B1.4). The handback-resolution
**never-halts** invariant (AC-B1.5) is intact: every new halt lives in step 4's *verification*, not
in field defaulting — a degraded resolution now yields a set that is either concretely verifiable or
provably empty, and an empty one fails loudly instead of passing silently. Clean-tree and
single-worktree gates (AC-B1.9), actor ownership (Orchestrator/operator push, open, merge — never
Stage), and P-009/P-011/P-016 are textually unchanged. Every command added is a read form
(`diff-tree`, `cat-file`, `rev-parse`, `merge-base`) — none is a push verb and none grants a commit
permission, so the detector's closed construct set and the fixture corpus are **not** widened.

### 14.5 Stale-reference sweep (finding 2 — file-wide, not line-59-only)

The cited line was the plan row of §3. A sweep of every plan-section and revision reference in this
file found the defect had **three** instances, two of them in the plan itself — which is why the fix
is not confined to the cited line:

| Location | Was | Now |
|---|---|---|
| memory §3 plan row | `revision 3`, `§10a hardening` | `revision 7`, `§9 Plan hardening signals` |
| memory §3 deliberation row | `§8`, `§9` addenda only | `§8`, `§9`, `§10`, `§11` |
| plan front matter L18 | `see §9/§10a` | `see §9` |
| plan §5.3 | `named in §10a.1` | `named in §5.0 and §9` |

Every remaining `§N` reference in this file was re-checked against the plan's actual heading list
and the deliberation's; all resolve.

### 14.6 Surfaces reconciled

| Surface | Change |
|---|---|
| Plan (rev 6→7) | Front matter `revision: 7` + stale `§10a` refs; rev-7 header; §3 surface rows B1/B3; §5.0 harness rows B1/B3; §6.1.1 3a re-record; §6.1.2 corrections 1, 2, 4, 5 + rationale; AC-B1.5/B1.7/B1.8 strengthened, AC-B1.10/AC-B1.11 added; §6.3 + AC-B3.5; §7.2 AC-C2.4; §8 verification commands; §11 R14; §12 rev-7 record |
| `018.007-T` (B1) | Description AXIS 2 addendum (corrections 8–10); AC-B1.5/B1.7/B1.8 strengthened, AC-B1.10/AC-B1.11 added in **both** renderings; rev-7 reconciliation note. Size **unchanged** (M) |
| `018.009-T` (B3) | Description item (5); AC-B3.5 extended in both renderings; rev-7 note. Size **unchanged** (S) |
| `018.011-T` (C2) | AC-C2.4 extended in both renderings (shape agreement, not name-only); rev-7 note. Remains plan §7.2 **verbatim** |
| `018-F` | DoD: new concrete-file bullet; rev-6 bullet corrected (pathspec constant, arm text); execution-order line |
| Deliberation | §11 Revision 7 addendum — the always-passing-verification lesson, generalized |
| `017-S` | **Unchanged** — one shipment, 9 items, order A1→A2→A3→{B1,B2,B3}→C1→C2 |
| `018.004-T`, `018.005-T`, `018.006-T`, `018.008-T`, `018.010-T` | **Untouched** |

### 14.7 Validation (Stage scope — no Go build/test, per the Stage role boundary)

Recorded in the cycle-6 reply on each thread. AC-ID parity re-swept to an exact bijection;
dependency edges, shipment membership and order unchanged; markdownlint clean on changed markdown;
backlog sync and doctor clean; no stale plan-section reference remains in this file. Go
build/test deliberately **not** run — Ship's responsibility (P-010).

### 14.8 Handoff

Unchanged: shipment **017-S**, `queued`, 9 items, Ship starts at Step 2 (harness-architect, H0).
The B1 harness function now additionally asserts the **absence** of the `git show origin/main:{path}`
loop and of any directory-prefix default in the no-shipment arm. Per the operator's standing
instruction, if this cycle does not achieve Copilot convergence the next step is an Orchestrator-run
adversarial review round, **not** another self-directed Stage cycle.

---

## 15. PR #54 adversarial review remediation — rev 8

**Trigger**: the §14.8 escalation condition fired. The Orchestrator convened an **adversarial
review** — anchor **GPT-5.6 Sol** with **GPT-5.4-mini**, **Claude Sonnet 5**, and **Claude Opus 5**;
**no route degradation** — and routed the verdict back to Stage for a scoped remediation pass on
this branch. All nine findings were classified **same-contract completion** under P-021; none
reopened the release unit's scope. Concurrent Copilot review `5184453325` (one visible inline
comment, six suppressed) raised no concern outside this set.

### 15.1 The two P1 blockers that mattered

**Blocker 1 — the verified set stopped at the tip commit.** Rev 7 derived the no-shipment arm's
path set from `stage_head_commit`'s **single-commit** diff. On a multi-commit Stage branch — the
ordinary case, and the shape of this very branch — artifacts written in earlier commits were never
read on `origin/main`; they rested on **commit ancestry**, which is exactly the property rev 7's own
§11.1 had judged insufficient when it introduced the per-file check. Rev 7 relocated the gap instead
of closing it.

Rev 8 introduces **`stage_base_commit`**, captured and retained by the **Orchestrator immediately
before it invokes Stage**, and derives the set from a **two-tree** diff:

```text
git diff --name-only -z --diff-filter=d {stage_base_commit} {stage_head_commit} -- {STAGE_ARTIFACT_ROOTS}
```

Two properties make this the right form, and both were checked rather than assumed:

* A two-tree diff is a property of the two trees, so it reads **identically before and after the
  merge** — the post-merge-stability requirement that made rev 7 reject the range form. The range
  form `origin/main..{stage_head_commit}` remains **refused**: it is empty by construction once the
  head is an ancestor of `origin/main`, which is precisely the state the arm runs in.
* It covers **every** commit of the session, not just the last.

The arm grows from five to **six** ordered checks (base resolution and base→head ancestry are new).
Step 3a **preserves** the base across its own commit — re-recording it would collapse the range, and
re-recording it to the pre-commit `HEAD` would silently drop already-committed unpushed commits, the
very population that path exists to handle.

**Blocker 2 — `no-shipment` was unreachable.** Rev 6 declared the no-shipment route "executable end
to end" and neither rev 6 nor rev 7 tested that claim against the **producer**. The installed
`_stage.agent.md` Step 5.5 declares shipment assembly mandatory, and Step 6's pre-summary gate halts
without a `shipment_id` — so `stage_outcome: no-shipment` could never be emitted, and B1's entire
second arm, the 018-F DoD bullet depending on it, and the 018.011-T coupling criterion checking it
were all gated on an outcome nothing could produce.

This is the **third instance of one failure class**: rev 1 obliged Stage to merge a PR it was
forbidden to create; rev 6 obliged the Orchestrator to find a branch nothing told it about; rev 8
found a verification keyed on an outcome nothing could emit. Each time a *consumer* obligation was
specified without confirming a permitted *producer* existed. Recorded as deliberation §12.2.

B3 gains **AC-B3.6** and two additional edit sites in a file it already owns. The load-bearing
constraint is stated three times across the surfaces, deliberately: **failures stay failures.**
P-003 lineage violations, harvest failures, and a missing required shipment after a **non-empty**
harvest still **halt** and are never recorded as `no-shipment`. Step 5.5's existing empty-harvest /
unresolved-P-003 guardrail is preserved **verbatim** because it is already the correct discriminator
between "nothing to ship" and "failed to ship". Without that, the fix would have traded an
unreachable verification for a silent failure channel — worse than the defect. Recorded as **R15**.

### 15.2 The other two P1s, and the five precision findings

| # | Finding | Fix |
|---|---|---|
| 3 | `018.007-T`'s description still stated the rev-6 algorithm (root-default path set, `git show` loop) as operative-looking instructions, with rev 7 layered on as an addendum — two contradictory contracts in one task body | Description **rewritten** to state the current contract once. Superseded algorithms are recorded as an explicit **FORBIDDEN FORMS** clause rather than as competing prose; all archaeology moved to implementation notes |
| 4 | Plan §6.2 and `018.008-T` carried **two divergent five-criterion** B2 lists — symmetry in one, the P-016/P-001 clarification in the other — with AC-B2.3/B2.4/B2.5 bound to **different** criteria on each surface | Reconciled to one identical **six**-criterion contract (AC-B2.1–AC-B2.6), carried verbatim on both surfaces. A count-only parity check would have passed this: both sides had five |
| 5 | Brittle `workflow-policies.md` **L238** locators for the Ship prohibition in plan §2, §5.3, §7.2/AC-C2.3, `018.011-T`, and deliberation §1.1 | Replaced with the structural anchor: P-010's **Ship MUST NOT** bullet "Commit or push directly to `main`". Deliberation §8.5 *announced* this switch at rev 2 and never applied it to §1.1; it is applied now. The ordinals were not merely brittle — **B2 appends an Amendment Log row to that file**, shifting them as a direct consequence of this shipment |
| 6 | "The `go test` step in job `expensive`" names no step in `ci.yml` | Cites job `expensive` (`name: test`) step **`Test (race)`**, command `go test -race -mod=readonly ./...`. AC-C2.7's constitutional `go test ./...` is stated as **separate and still required**; per-task `harness_cmd` values stay narrow and unchanged |
| 7 | AC-A1.2 named the prefixes `reject-*` / `accept-*`; the files are `directpush-reject-*` / `directpush-accept-*` | Corrected in plan §5.4 and `018.004-T`, both renderings |
| 8 | AC-A2.1 demanded the **bare** scan exit non-zero — unsatisfiable, because A2's stub classifies every construct `accept`, so a bare scan correctly exits **0** | Non-zero scoped to **`--self-test` only**; bare-mode red reserved for **AC-A3.3**, after a real detector exists. The A2→A3 dependency edge already orders them |
| 9 | Stale references: Constitution Check row XI cited AC-C2.5 (markdownlint) not AC-C2.6; §12's rev-7 row promised "(see below)" with no rev-7 block; the deliberation header claimed the branch had "no push, no PR" | All corrected. The rev-7 record block is added; the deliberation header now names its five addenda and PR #54 |

### 15.3 Scope held

No task, file, fixture, manifest entry, harness function, CI step, ledger entry, CODEOWNERS line,
policy ID, or dependency edge added. Two new criteria (**AC-B2.6**, **AC-B3.6**), six strengthened,
four corrected in place, one reconciled ID set, two new risk rows (**R15**, **R16**). Shipment
**017-S** and feature **018-F** are preserved exactly — 9 items, order
A1→A2→A3→{B1,B2,B3}→C1→C2. `018.009-T` re-sized **S→M** on volume alone (five edit sites, one file,
one domain, fully-specified text); every other size and complexity value is unchanged.

**Deferred, on the operator's instruction**: the pre-existing backlog-registry complexity/sizing
schema drift (panel finding F12) is **not** repaired here. The degradation note carried by every
task remains the accurate record of it.

**Not swept, recorded rather than hidden**: `_stage.agent.md` **L42** still appears as a
line-ordinal anchor in the plan's narrative (§2, §5.3, §5.4, §12 round-1 record) and the
deliberation's archaeology. Finding 5 was scoped to the **Ship prohibition** locators, and those
narrative uses are historical rather than operative — no acceptance criterion depends on them. They
are a candidate for a future sweep, not a defect this pass left in an executable surface.

### 15.4 Validation (Stage scope — no Go build/test, per the Stage role boundary)

`backlogit sync` + `backlogit doctor` clean; AC-ID parity re-swept to an exact plan↔task bijection;
shipment membership, item count, and dependency edges unchanged; markdownlint clean on every
changed markdown file; `git diff` reviewed file by file. Go build/test deliberately **not** run —
Ship's responsibility (P-010). No source, test, script, workflow, policy, or agent file was
modified by this pass; those remain Ship's execution surfaces.

### 15.5 Handoff

Unchanged: shipment **017-S**, `queued`, 9 items, Ship starts at Step 2 (harness-architect, H0).
The B1 harness function now additionally asserts the absence of a **single-commit** derivation in
the no-shipment arm; the B3 function additionally asserts the **reachable** `no-shipment` terminal
alongside the preserved halt clauses. Stage does not push, open, or update PR #54 — the Orchestrator
owns every GitHub operation on this branch.

---

## 16. PR #54 review remediation cycle 7 — rev 9

**Trigger**: one **visible** Copilot review comment on PR #54 (`3994429008`, thread
`PRRT_kwDOTPuhps6hrbAk`, on `.backlogit/queue/018.009-T.md:22`), routed by the Orchestrator to
Stage as a scoped same-contract remediation. Classified **same-contract completion** under P-021 —
it completes the B3 contract rather than reopening the release unit's scope.

### 16.1 The finding, and why it is correct

> This task updates Stage's authorization and declares a handback, but it does not add any operative
> Stage step to create/check out the dedicated branch or commit the artifacts. […] Implementing
> AC-B3.1–B3.5 as written can therefore still leave Stage on the default branch, making the reported
> `stage_head_commit` and concrete-path contract unverifiable.

Verified against the installed surface rather than accepted on assertion: `_stage.agent.md`'s
heading list runs Role Boundary → Step Sequence Contract → Steps 0.0, 0.1, 0, 1, 1.5, 1.8, 2, 3,
3.0–3.3, 4, 5, 5.0–5.3, 5.5, 5.6, 6 → Shipment Context → Session Continuity. **No step in that list
performs a branch operation or a commit.** The Step Sequence Contract checklist — which the file's
own preamble makes the definition of what a session MUST execute — lists none either. Revisions 1–8
gave Stage the *permission* (Role Boundary Git cell, P-010 grant) and the *reporting obligation*
(Step 6 handback), and stopped.

The consequence is not theoretical. An executor satisfying AC-B3.1–AC-B3.6 to the letter ends the
session on the default branch with every artifact written there — the `fdff9e4` shape — and then
emits `stage_branch`/`stage_head_commit`/`stage_artifact_paths` describing a branch never created
and commits never made. The Orchestrator's step-4 arm *rejects* that handback, so the gate still
fails closed; but the artifacts are already on the default branch by then, and **detection after
the fact is not prevention**.

### 16.2 Fourth instance of the producer/consumer failure class

| Rev | Consumer obligation | Missing half |
|---|---|---|
| 1 | Stage merges the staging PR | Stage was **forbidden** to create one |
| 6 | Orchestrator finds the Stage branch | nothing **announced** it |
| 8 | Orchestrator verifies `no-shipment` | nothing could **emit** it |
| **9** | Orchestrator consumes `stage_head_commit` | nothing **produced** it |

Deliberation §12.2's rule — *name the emitting surface and confirm it is permitted to emit* — was
satisfied here and still insufficient: the emitting surface was named (Step 6) and permitted, and
the value still had no origin. Deliberation §13.2 extends the rule with the two clauses that were
missing: confirm the surface is **instructed** to produce the value, and that the instruction is
**ordered** relative to the state the value describes.

### 16.3 Resolution — two operative steps, and one naming correction

Plan §6.3.1 specifies both, with full replacement text:

* **Step 1.9 — Stage Artifact Branch Gate**, between the learnings-retrieval and deliberation
  steps — a **fixed** position, with every mutation that would otherwise precede it (the Step 1
  deferred-expansion duplicate archival, the session's first memory checkpoint) **deferred until
  after** it. Derives `{scope-slug}`; records `stage_base_commit` (echoed from the Orchestrator in a
  pipeline invocation, captured from `HEAD` **before** branch creation in a direct one); **verifies
  and uses** an Orchestrator-supplied branch or **creates and checks out** `chore/stage-{scope-slug}`
  when none was supplied; refuses to write while `HEAD` is the default branch
  (`STAGE_BRANCH_GATE_FAIL`, P-010 recorded); switches the **single existing** worktree (P-016).
* **Step 5.7 — Stage Artifact Commit**, between consumed-stash archival and the summary. Asserts
  `HEAD`; stages only the four `STAGE_ARTIFACT_ROOTS`; commits conventionally; sets
  `stage_head_commit` from the **resulting** `HEAD`; derives the aggregate path set over the
  **preserved** base; and **stops** — no push, no PR create/update/merge.
* Both are **registered in the Step Sequence Contract checklist** and enforced by Step 6's
  pre-summary gate. A step section the checklist omits is advisory in this file, so registration is
  not bookkeeping.

**Naming, corrected as a direct consequence.** `chore/stage-{shipment_id}` is **underivable** once
the branch must precede the first artifact write, because Step 5.5 creates the shipment as the
session's *last* mutation. Following rev 8 literally would push an executor to defer the branch past
the writes it protects — a second route back to `fdff9e4`. The form is now
`chore/stage-{scope-slug}`, with a shipment ID kept as a **permitted** slug value rather than
deleted. This branch, `chore/stage-pipeline-policy-gap`, is a conforming example — and was already
conforming, which is why the defect was invisible to a reader checking only the current branch.

**Ordering, not presence, is the contract.** A gate after harvest leaves the artifacts on the
default branch; a commit after the summary emits a head that does not exist. Plan §6.3.2 therefore
specifies detection by **checklist-index and heading-position comparison** inside the existing
`TestDirectPushGate_StageAgentSurfaceClean` — never a text search, because `_stage.agent.md` names
its own steps in the Step 6 gate and a name search would pass with the sections deleted. That is the
AC-C1.6 self-match hazard one file over, and the same structural remedy.

### 16.4 One defect found by internal review before commit

The first draft of Step 1.9 gave it a **floating** position — "earliest point a slug is derivable,
latest immediately before the first artifact write, whichever comes first" — plus a provisional-slug
escape hatch so it could fire early when a Step 1 deferred-expansion reconciliation needed to
archive a duplicate stash entry. An adversarial pass over the draft (GPT-5.6 Sol, scoped to the
diff) rejected it, correctly: `_stage.agent.md`'s **Step Sequence Contract states that a session
MUST execute its steps in order**, so a step registered between Step 1.8 and Step 2 cannot also be
specified to run inside Step 1. The contradiction was not cosmetic — it made the gate's real
position unassertable (AC-B3.7 enforced the ordinal while the prose licensed an earlier run) and
left a tracked `.backlogit/` write legal on the default branch, which is the defect this whole
revision exists to close.

The fix inverts the relationship: the ordinal is **fixed**, and the two mutations that would
otherwise precede it — the Step 1 duplicate archival and the session's first memory checkpoint —
are **deferred** past it by an explicit rule. Deferral is sound: the gate is at most two steps
later, nothing downstream consumes either mutation before Step 2, and the archival's *analysis*
(the read-only half) still completes in Step 1. Recorded rather than silently repaired, per the
§13.6b precedent — the defect class, *a guard whose stated position contradicts the ordering
semantics of the contract it is inserted into*, is worth recognizing again.

### 16.5 Scope held

No task, file, fixture, manifest entry, harness function, CI step, ledger entry, CODEOWNERS line,
policy ID, shipment, or dependency edge added — the finding's "or split that work into an explicit
task" alternative was **declined** on operator instruction and on merit: the work is four more edit
sites in the one file `018.009-T` already owns, in the one domain it already works in, and a ninth
task would need a ninth harness function for assertions the existing B3 function's surface already
covers. One new criterion (**AC-B3.7**), three strengthened (AC-B2.2, AC-B3.2, AC-B3.5), two
re-scoped (AC-B3.3, AC-B3.6), one extended (AC-C2.4), two new risk rows (**R17**, **R18**).
`018.009-T` re-sized **M→L** on volume alone; every other size and complexity value unchanged.
Shipment **017-S** and feature **018-F** preserved exactly — 9 items, order
A1→A2→A3→{B1,B2,B3}→C1→C2.

**Rev-8 fixes preserved, verified not assumed**: the aggregate `{stage_base_commit}
{stage_head_commit}` two-tree derivation, the six-check no-shipment arm, the reachable
`no-shipment` terminal with its verbatim P-003 guardrail, and the refusal of both the
single-commit and range derivations are all intact; rev 9 adds the producing steps *behind* them
and changes none of them.

**Deliberately not swept, recorded rather than hidden**: `chore/stage-{shipment_id}` still appears
in historical revision narratives (the plan's Revision 6 paragraph, §6.1.1's quotation of the 3e
text B1 deletes, §12's rev-6 record), in the deliberation's archaeology, and in `018.007-T`'s rev-6
reconciliation note. Those record what earlier revisions said and what B1 removes; no acceptance
criterion depends on them. Every **operative** occurrence is corrected.

### 16.6 Validation (Stage scope — no Go build/test, per the Stage role boundary)

`backlogit sync` + `backlogit doctor` clean; AC-ID parity re-swept — **51 plan IDs ↔ 51 task IDs**,
exact bijection; shipment `017-S` membership (9 items), status, and dependency edges unchanged;
`markdownlint` clean on every changed markdown file; `git status --porcelain` empty after commit.
Go build/test deliberately **not** run — Ship's responsibility (P-010). No source, test, script,
workflow, policy, or agent file was modified by this pass.

### 16.7 Handoff

Unchanged: shipment **017-S**, `queued`, 9 items, Ship starts at Step 2 (harness-architect, H0).
The B3 harness function additionally asserts that the branch gate and the artifact commit are
**present and ordered** — checklist entry indices and heading positions, not text matches. Stage
does not push, open, update, or comment on PR #54, and does not reply to or resolve thread
`PRRT_kwDOTPuhps6hrbAk` — the Orchestrator owns every GitHub operation on this branch.

## 17. PR #54 second adversarial review remediation — rev 10

**Source**: operator-directed **second adversarial round** (anchor **GPT-5.6 Sol** with
**GPT-5.4-mini**, **Claude Sonnet 5**, and **Claude Opus 5**; **no route degradation**). Four
**visible** Copilot threads plus one adversarial-only finding, all classified **same-contract
completion** under P-021. Scope limited to those five; no other scope reopened.

### 17.1 The two defect classes this round separated

Rev 10 is the first round to find **two distinct classes at once**, and flattening them would have
lost the second.

**Class 1 — producer/consumer, fifth instance, and the worst form yet.** Thread
`PRRT_kwDOTPuhps6hsPRZ`. Rev 1/6/8/9 were all **absences**: a value nobody could create, announce,
emit, or produce. This one is a **substitution**. The consumer held a correct reported
`stage_outcome` and overwrote it with an unconditional derivation whose `otherwise` clause is a
**success** terminal. So a run that reported `shipment` but whose `shipment_id` did not arrive was
resolved to `no-shipment`, the shipment arm was skipped, the no-shipment arm **passed** (it
verifies *files*, not *intent*, and the artifacts genuinely were on `origin/main`), the gate
reported success, and **Ship was never routed**. An absence is loud at the point of use; a
substitution is silent by construction — and this one fired on exactly the input class most likely
to be degraded, since a handback missing its `shipment_id` is by definition one that was not fully
received.

**Class 2 — an obligation with no instrument.** Threads `PRRT_kwDOTPuhps6hsPRq`,
`PRRT_kwDOTPuhps6hsPR5`, `PRRT_kwDOTPuhps6hsPSJ`, and the adversarial harness finding. Each names
an actor, sits at the right ordinal, and is still inert because the operation it prescribes
**cannot observe the condition**:

| Obligation | Prescribed operation | Why it could never fire |
|---|---|---|
| halt on a change outside `STAGE_ARTIFACT_ROOTS` | root-scoped `git add` | a pathspec **ignores** what it excludes; nothing reports it |
| defer the mutations that precede the branch gate | an **enumeration of two** | the contract schedules five-plus tracked writes earlier; the rest stayed legal on the default branch |
| `go test ./...` red after H0, green per task | **all eight** red simultaneously | Ship's Step 4.3 re-reads the full suite after *every* task, so task 2 could never pass its own gate |

The discipline this adds, recorded in deliberation §14.1: *for every MUST, name the operation that
discharges it, and confirm that operation can observe the thing the MUST is about.*

### 17.2 What was verified by execution, not by reading

Four claims in this round are load-bearing and were **run** before being written down:

| Claim | Verification |
|---|---|
| A dotted test flag is shell-fragile; a dot-free one is not | `go test . -run '^T$' -directpush.task=alpha` → `flag provided but not defined: -directpush` (PowerShell splits on the dot); `-gatetask=alpha` parses unquoted and reaches the binary |
| "exits non-zero" alone cannot prove a test ran | `go test -run '^TestNoSuchTest$'` → `ok … [no tests to run]`, exit **0**. Hence AC-A2.5's `--- FAIL:` non-vacuity guard |
| `git status -z` rename shape | ` D keep.txt\0R  new name.txt\0old name.txt\0?? sub/deep.md\0?? untracked.md\0` — the ORIG_PATH is a **separate NUL record with no `XY` prefix** |
| `--untracked-files=all` is not optional | default `normal` emits `?? sub/` — a **directory** — which no per-file containment test can evaluate |
| The disposables really are gitignored | `git check-ignore` confirms `.backlogit/checkpoints/`; `.gitignore` carries `.backlogit/backlogit.db*`, `.backlogit/hooks_queue.jsonl`, `.backlogit/runtime/`. `.backlogit/queue/` and `.backlogit/archive/` are **tracked** (230 files), so stash writes are tracked and the deferral rule does bind them |

### 17.3 Resolution — four corrections and one mechanism

1. **Preserve, then validate** (§6.1.2 correction 6, AC-B1.12, AC-B3.5). The derivation is scoped
   to the **absent** case; the pair is validated fail-closed before arm selection. The halves sit
   at deliberately different points — preservation in **defaulting** (strictly weaker, so the
   never-halts invariant is untouched), validation in **verification** (where every other
   `STAGING_GATE_FAIL` lives). That placement is the rev-6 lesson applied on purpose rather than
   rediscovered.
2. **A categorical deferral rule** (§6.3.1 correction 1, AC-B3.7). "Every tracked Stage artifact
   mutation", with three classes as illustration. Read-only classification and grouping still run
   first — they derive the slug, so deferring them would make the gate's own position underivable.
   Gitignored disposables **expressly excluded**, not silently claimed.
3. **A detector for the out-of-root halt** (Step 5.7 sub-step 2, AC-B3.7). NUL-safe full-tree
   inspection before staging; both rename endpoints; halt fail-closed and **leave the dirt exactly
   as found** — no checkout, restore, stash, revert, or delete (Constitution VII).
4. **Two exactly-anchored invocations** for AC-A2.5, with the `--- FAIL:` guard.
5. **Per-task activation** (new §5.0.1): one dot-free selector flag, one eight-row table, an
   activation-independent A1 anchor that keeps H0's default-suite red literal, and four executed
   mutation-proof checks (MP0–MP3).

### 17.4 Honesty held, in three places

* **Harness-ready is not redefined.** It still means what `harness-architect` verifies —
  compilation clean, default suite red. §5.0 now *says* that instead of letting the label imply
  eight simultaneous reds. The per-task red stays a per-task obligation discharged by Ship at claim
  time from that task's `harness_cmd`. No label from prose — the rev-4 lesson, held.
* **The deferral rule does not overclaim.** Gitignored index/checkpoint/hook-queue operations are
  named as **outside** it. Stating what a control does not cover is part of the control.
* **R19's residual is recorded, not engineered away.** Between A1 and C1 a landed-then-regressed
  surface would re-skip rather than fail, because MP2 does not bind until C1's surface exists.
  Binding it earlier would make the suite red for every not-yet-started task — the deadlock rev 10
  exists to remove.

### 17.5 One latent defect found by internal review before commit

Requirement (d) of the harness redesign — *the final `go test ./...` runs all eight real assertions
green* — forced a check revisions 4–9 never made: **is every function's assertion still true at the
end state?** Seven were. **A2's was not.** Its specified assertion was that `--self-test` reports
every reject fixture failing **against the all-accept stub**, with a non-zero exit — a property
true only while A2's stub is the current implementation, and **false the moment A3 installs the
real detector**. That function would have gone red at A3 and **stayed** red, wedging
`go test ./...` for the remainder of the shipment.

Activation did not cause this. It made it **visible**, because "all eight green at the end" had
never been written down as a requirement before rev 10 had to state it. The fix: the function
asserts the **durable** driver contract A2 actually delivers (enumerate every manifest fixture by
name, emit a per-fixture verdict, fixtures before the real-tree scan), and the stub-specific
observations stay in **AC-A2.1**, read from the script's own output at A2's boundary. AC-A2.1 is
**not weakened** — it still demands the observed red for the detector, still binds `--self-test`
only, and still reserves bare-mode red for AC-A3.3; only its carrier is named.

Recorded rather than silently repaired, per the rev-6 and rev-9 precedent. The defect class —
*a regression assertion pinned to a transitional state* — generalizes: a harness written once, up
front, for a multi-task shipment must assert what is true **when the shipment is done**, not what
is true at the moment the assertion's task completes.

### 17.6 Scope held

No task, file, fixture, manifest entry, harness function, CI step, ledger entry, CODEOWNERS line,
policy ID, shipment, or dependency edge added. One new criterion (**AC-B1.12**), four strengthened
(AC-A2.5, AC-B1.5, AC-B3.5, AC-B3.7), one extended (AC-C2.4), one new plan subsection
(**§5.0.1**), one new B1 correction (**correction 6**), two new risk rows (**R19**, **R20**), and
one updated (R18). All eight `harness_cmd` values gain `-gatetask={ID}`; **no harness function is
added or renamed**. **No task re-sized** — rev 10 reworks clauses inside step sections rev 9
already specified. Shipment **017-S** and feature **018-F** preserved exactly — 9 items, order
A1→A2→A3→{B1,B2,B3}→C1→C2.

**Rev-8 and rev-9 fixes preserved, verified not assumed**: the aggregate two-tree derivation, the
six-check no-shipment arm, the reachable `no-shipment` terminal with its verbatim P-003 guardrail,
the refusal of the single-commit and range derivations, and the two operative ordered steps are all
intact. Rev 10 tightens the rules *inside* them and changes none of them.

### 17.7 Validation (Stage scope — no Go build/test, per the Stage role boundary)

`backlogit sync` + `backlogit doctor` clean; AC-ID parity re-swept — **52 plan IDs ↔ 52 task IDs**,
exact bijection; shipment `017-S` membership (9 items), status, and dependency edges unchanged;
`markdownlint` clean on every changed markdown file; `git status --porcelain` empty after commit.
Go build/test deliberately **not** run — Ship's responsibility (P-010). No source, test, script,
workflow, policy, or agent file was modified by this pass.

### 17.8 Handoff

Unchanged: shipment **017-S**, `queued`, 9 items, Ship starts at Step 2 (harness-architect, H0).
**New for Ship**: H0 must emit the §5.0.1 activation helper and table alongside the eight
functions, and every task's `harness_cmd` now carries `-gatetask={ID}`. The default suite is red at
H0 from the A1 anchor and green between tasks thereafter, so Step 4.3 no longer deadlocks. Stage
does not push, open, update, or comment on PR #54, and does **not** reply to or resolve threads
`PRRT_kwDOTPuhps6hsPRZ`, `PRRT_kwDOTPuhps6hsPRq`, `PRRT_kwDOTPuhps6hsPR5`, or
`PRRT_kwDOTPuhps6hsPSJ` — the Orchestrator owns every GitHub operation on this branch.
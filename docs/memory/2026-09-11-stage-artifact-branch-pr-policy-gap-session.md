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
- `INDEX_SYNC_OK` at session start (199 artifacts) and at session end (221 artifacts).
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

- `backlogit sync` → `INDEX_SYNC_OK`, 220 artifacts (start and end)
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

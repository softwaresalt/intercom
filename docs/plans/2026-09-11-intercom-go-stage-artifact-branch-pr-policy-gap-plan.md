---
title: "Implementation Plan — Stage Artifact Branch/PR Policy Gap Correction"
date: 2026-09-11
revision: 10
status: reviewed
agent: Stage
governs: stash 638A410B
source: docs/decisions/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-deliberation.md
---

# Implementation Plan — Stage Artifact Branch/PR Policy Gap Correction

**Source document**:
`docs/decisions/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-deliberation.md`
(including its **§8 Revision 2 addendum**, **§9 Revision 3 addendum**, **§10 Revision 6
addendum**, **§11 Revision 7 addendum**, **§12 Revision 8 addendum**, **§13 Revision 9
addendum**, and **§14 Revision 10 addendum**)
**Stash entry**: `638A410B`
**Requires plan hardening**: **yes** — see §9.

**Revision 2** incorporated the seven-persona plan review. Three rev-1 claims were falsified and
are corrected here: a **third** authorization surface exists (`_stage.agent.md` L42), the rev-1
P-010 wording was **self-deadlocking**, and deleting 3e **does** lose a behaviour. See
deliberation §8.

**Revision 3** incorporates the round-2 review. A **fourth** authorization surface was found — the
Step 1.5 **preamble**, which is commit-shaped and marker-free — and the detector specification is
tightened per-construct (block-level verdicts masked mixed blocks), structurally (bold lead-in, not
ATX heading; cell-scoped, not row-scoped), and defensively (guarded default-branch resolution,
`pull`/`checkout` exclusions, permissive-vs-verificational discrimination). Sub-epic A is split
A1/A2/A3 for granularity. See deliberation §9.

**Revision 4** incorporates the PR #54 current-HEAD review (two P-021 C1 same-contract blockers).
Rev 3's §10.1 declared a **P-002/P-004 deviation** and asked Ship to apply `harness-ready` from a
prose label — which Ship's Step 2 cannot honour, because it invokes `harness-architect` for every
queued task lacking that label and that skill only produces Go `_test.go` harnesses. Rev 4 removes
the deviation entirely: the release unit is restructured so the **existing** harness path is
genuinely satisfied by a real Go regression harness that drives the bash gate (§5.0), matching the
established `tests/integration/build_script_test.go` precedent of a Go test whose whole subject is
a non-Go script. No policy is amended, no test-first requirement is weakened, and no fake
scaffolding is added. Rev 4 also reconciles §7.2's C2 contract with task `018.011-T` on a single
**fixture-backed, no-dirty-tree-edit** mutation proof.

**Revision 5** incorporates the PR #54 current-HEAD review, cycle 4 (one finding, thread
`PRRT_kwDOTPuhps6hnPkD`). Rev 4's **R5 accepted the residual** that a future render could wipe the
job-`lint` gate step, leaving CI green with the detector never invoked — which directly contradicted
§1's "cannot **silently** reopen the gap" and the feature DoD. Rev 5 **does not narrow the
objective**. It instead identifies the tracked enforcement path that already survives the render
boundary — the §5.0 Go harness, invoked by the **generated-baseline** `Test (race)` step
(`go test -race -mod=readonly ./...`, job `expensive`) rather than by
a LOCAL DIVERGENCE step — and makes that path explicit, load-bearing, and mechanically
self-verifying (§5.0 "Regeneration-resistant enforcement", AC-C1.6, AC-C2.8). R5 is reclassified
from accepted residual to mitigated; the narrower, genuinely-visible residual is recorded as R12.

**Revision 6** incorporates the PR #54 current-HEAD review, cycle 5 (one finding, thread
`PRRT_kwDOTPuhps6hqWDe`, comment 3993998429, plan L517). Revisions 1–5 corrected *what* Step 1.5
**authorizes** but never touched *how* Step 1.5 **discovers** the work it must gate. Its two inputs
are `git status --short -- .backlogit/` and `git log origin/main..main` — a `.backlogit/`-only
pathspec and a **default-branch-only** commit range. Once B2/B3 require Stage to commit on
`chore/stage-{shipment_id}` / `chore/stage-{chore-slug}` and leave the worktree there, **both
inputs read "synchronized"**: the commit is not on the default branch and (for a no-shipment run)
the written paths are not under `.backlogit/`. Step 1.5 then falls through to its `origin/main`
manifest check and fails — so the no-shipment route B2 documents end to end was **unexecutable as
written**. Rev 6 makes the Stage artifact branch an **explicit handback** (`stage_branch`,
`stage_head_commit`, `stage_outcome`, `stage_artifact_paths`), makes discovery **branch-specific**
(`origin/main..HEAD`, proven equal to `origin/main..$STAGE_BRANCH` by a HEAD assertion that doubles
as the P-016 single-worktree gate), widens the dirtiness pathspec to the Stage artifact paths, and
gives step 4 a **second verification arm** keyed on `no-shipment` that needs no `shipment_id`. No
task, file, fixture, harness function, or dependency edge is added; 017-S stays one shipment.

**Revision 7** incorporates the PR #54 current-HEAD review, cycle 6 (two findings, threads
`PRRT_kwDOTPuhps6hrC0c` / comment 3994280334 and `PRRT_kwDOTPuhps6hrC0o` / comment 3994280351).
Rev 6's no-shipment arm iterated `git show origin/main:{path}` over a path set whose documented
default was the **directory list** `.backlogit/ docs/plans/ docs/decisions/ docs/memory/`. `git
show` on a tree **exits 0**, and those four directories already exist on `origin/main` in every
repository the gate will run in — so the loop could pass having read **no file this Stage run
wrote**, and paired with the degraded `stage_head_commit` default (`git rev-parse HEAD`, routinely
already an ancestor of `origin/main`) the arm reported a **false pass**. Rev 6 removed the
zero-iteration loop and left the always-satisfied one. Rev 7 makes `stage_artifact_paths` a
**non-empty set of concrete repository-relative file paths tied to `stage_head_commit`**, derived
deterministically from that commit's own changed-file set when the handback does not supply it,
validated for file-ness / root-containment / set-equality when it does, and verified with `git
cat-file -t` = `blob` rather than `git show` (§6.1.2 correction 4). Step 3 re-records the handback
commit after it creates one (correction 5), so the derived set is taken from the commit that
actually carries the artifacts. Step 1's dirtiness pathspec is pinned to the **fixed** allowed-root
constant rather than the handback field, closing the same class of hole one step earlier. Rev 7
also refreshes the stale plan-revision and section cross-references in the session memory record.
No task, file, fixture, harness function, or dependency edge is added; 017-S stays one shipment.

**Revision 8** incorporates the PR #54 **adversarial review** (anchor GPT-5.6 Sol, with
GPT-5.4-mini, Claude Sonnet 5, and Claude Opus 5; no route degradation), which classified every
finding below as **same-contract completion** under P-021. Four P1 blockers and five P2/P3
precision defects are remediated. **Blocker 1 — aggregate verification.** Rev 7's no-shipment arm
derived its verified file set from `stage_head_commit`'s **single-commit** diff, so artifacts
introduced by *earlier* commits on a multi-commit Stage branch were never read on `origin/main`: a
session that writes the deliberation in commit 1 and the plan in commit 2 had commit 1 covered by
ancestry alone. Rev 8 introduces `stage_base_commit` — captured by the **Orchestrator immediately
before it invokes Stage** and retained as the authoritative pre-invocation value — and derives the
artifact set from a stable **two-tree** diff over `{stage_base_commit} {stage_head_commit}`, which
reads identically before and after the merge (unlike the brittle `origin/main..{commit}` range
form, empty by construction post-merge). **Blocker 2 — reachable `no-shipment`.** Rev 6/7 built a
`no-shipment` verification arm that no producer could ever trigger: the installed Stage Step 5.5
and Step 6 pre-summary gate both **halt** without a `shipment_id`, so `stage_outcome: no-shipment`
was unreachable and the arm was dead prescription. B3 gains the narrowly-scoped Stage-side change
that lets a **reviewed, valid** empty-harvest terminal outcome emit the full handback and stop —
with P-003 failures preserved as failures, never re-labelled as success. **Blocker 3** rewrites
`018.007-T`'s description to a single authoritative contract (its rev-6 prose was still
operative-looking). **Blocker 4** reconciles §6.2 and `018.008-T` onto one identical **six**-criterion
B2 contract. **P2/P3**: line-ordinal locators for the Ship prohibition are replaced with structural
ones; generated-baseline evidence cites the exact `Test (race)` command; fixture-prefix wording is
corrected to the real `directpush-` names; A2's red obligation is scoped to `--self-test`; and
stale Constitution Check and deliberation-header references are corrected. No task, file, fixture,
harness function, or dependency edge is added; 017-S stays one shipment of nine items.

**Revision 9** incorporates the PR #54 current-HEAD review, cycle 7 (one visible Copilot finding,
thread `PRRT_kwDOTPuhps6hrbAk`, comment 3994429008, on `018.009-T:22`), classified **same-contract
completion** under P-021. Revisions 1–8 gave Stage the *permission* to create and commit on a
dedicated artifact branch (B2's P-010 grant, B3's Role Boundary cell) and the *obligation* to
report what it did (AC-B3.5's handback), but **never an operative workflow step that performs
either action**. The installed `_stage.agent.md` runs triage → grouping → deliberation → planning →
review → harvest → shipment → archive → summary and contains **no branch operation and no commit
operation anywhere**. An executor satisfying AC-B3.1–AC-B3.6 exactly as written could therefore
finish the session still on the default branch, having written every artifact there — the precise
`fdff9e4` shape this release unit exists to close — while emitting a handback whose
`stage_head_commit` and `stage_artifact_paths` describe commits that were never made on a branch
that was never created. Rev 9 closes it with **two operative steps** in the file B3 already owns:
a **pre-mutation branch gate** (Step 1.9) that runs once a stable scope slug is derivable and
before the session's first backlog/doc write, and a **post-mutation artifact commit** (Step 5.7)
that runs after every Stage mutation and before the Step 6 handback. Rev 9 also removes the
**shipment-ID naming dependency** that made the branch name underivable at gate time: the branch is
`chore/stage-{scope-slug}`, a shipment ID may *be* the slug when one is already known, and it is
never *required* — Step 5.5 creates the shipment long after the branch must exist. One acceptance
criterion is added (**AC-B3.7**) covering both orderings. No task, file, fixture, harness function,
shipment, or dependency edge is added; 017-S stays one shipment of nine items.

**Revision 10** incorporates the PR #54 **second adversarial review** (anchor GPT-5.6 Sol, with
GPT-5.4-mini, Claude Sonnet 5, and Claude Opus 5; no route degradation), which classified every
finding as **same-contract completion** under P-021. Four visible Copilot threads and one
adversarial-only finding are remediated. **Blocker 1 — the handback outcome pair could fail open.**
Rev 6–9 derived `stage_outcome` *unconditionally* from whether a `shipment_id` was returned, so a
Stage run that reported `stage_outcome: shipment` but whose `shipment_id` failed to reach the
Orchestrator was **silently reclassified** as a successful `no-shipment` — the loudest failure the
pipeline has, converted into a passing terminal. Rev 10 makes the field **preserved when reported**
and derived only when **absent**, and adds a fail-closed **pair validation** before arm selection
(`shipment` requires a non-empty `shipment_id`; `no-shipment` requires none; an unknown value or an
inconsistent pair halts `STAGING_GATE_FAIL`). **Blocker 2 — the pre-gate deferral rule was
under-inclusive.** Rev 9's Step 1.9 deferral named exactly two mutations, while the installed Stage
contract mandates a stash-classification checkpoint, a contextual-grouping/operator-selection
checkpoint, and the **write half** of late-identifier reconciliation — all of which precede
Step 1.9 and are all tracked. Rev 10 makes the rule **categorical and exhaustive** (Step 1.9
precedes *every* tracked Stage artifact mutation), enumerates the three classes explicitly, and
keeps gitignored disposable index/checkpoint/hook-queue operations **distinct** rather than
overclaiming them. **Blocker 3 — out-of-root detection was unimplementable.** Step 5.7 told Stage
to halt on a change outside `STAGE_ARTIFACT_ROOTS` while prescribing only a root-scoped `git add`,
which *ignores* such a change rather than detecting it. Rev 10 requires an explicit NUL-safe
full-tree status inspection **before** staging, with a concrete parse algorithm covering tracked
modifications, deletions, both rename endpoints, and recursed untracked files, and a fail-closed
halt that leaves pre-existing unrelated dirt **untouched**. **Blocker 4 — AC-A2.5 could not prove
what it claimed.** A single driver-scoped `-run` invocation cannot establish that another test is
red; rev 10 splits it into two exactly-anchored invocations with a non-vacuity guard. **Adversarial
finding — the harness lifecycle deadlocked against Ship's full-suite gate.** Rev 9's H0 left all
eight per-task assertions simultaneously red while Ship runs `go test ./...` after **every**
task, so no task could ever complete. Rev 10 adds one minimal **per-task activation** mechanism
(§5.0.1) that keeps the red phase honest and observable per task without wedging the default suite.
No task, file, fixture, shipment, or dependency edge is added; 017-S stays one shipment of nine
items.

**Artifact persistence note (P-010)**: this plan and its deliberation are themselves Stage
artifacts. They are committed on the dedicated branch `chore/stage-pipeline-policy-gap` and reach
`main` only via PR — the exact path this shipment installs. This shipment does not reproduce the
`fdff9e4` pattern it exists to close.

---

## 1. Objective

Make a dedicated Stage/admin branch plus a pull request the **only** path by which a Stage
artifact reaches the default branch — including when no shipment is formed — and enforce it
mechanically so regeneration of the autoharness-generated contract files cannot silently reopen
the gap.

## 2. Root cause — four authorization surfaces

| Surface | Authorizing text | Shape |
|---|---|---|
| `.github/agents/_stage.agent.md` Role Boundary table, **Git** row, Allowed cell | "Commit backlog/planning artifacts **on default or admin branch**" | **commit**, table cell |
| `.github/policies/workflow-policies.md` P-010 **Stage MAY** bullet | "Commit backlog and planning artifacts (**on the default branch** or a dedicated chore/admin branch)" | **commit**, list bullet |
| `.github/agents/_orchestrator.agent.md` Step 1.5 sub-step **3e** | "**Attempt a direct push to `main` first** … fall back to creating a staging PR" | **push**, instruction |
| `.github/agents/_orchestrator.agent.md` Step 1.5 **preamble** (rev 3) | "verify that all staging artifacts … **are committed to the default branch** and present on the remote" | **commit**, postcondition |

**Rev 3 — fourth surface.** The Step 1.5 preamble states the gate's postcondition as artifacts
being "committed to the default branch", with no PR in the stated end-state. Left as-is it both
(a) remains a compliant reading of direct commit-to-default, defeating §1's objective, and
(b) sits in the scanned corpus as a commit-shaped, marker-free construct — so it would either
trip the gate (making zero-findings unreachable) or force the detector to be written narrowly
enough to miss the `_stage.agent.md` L42 shape it exists to catch. It is corrected in B1.

`.github/instructions/role-enforcement.instructions.md` makes the **agent's own Role Boundary
table** the authoritative permission set at mutation time — so `_stage.agent.md` L42, not P-010,
is what the agent actually read before producing `fdff9e4`. Correcting only the other two would
produce a **false green**.

Ship's mirror prohibition already exists in the same policy — P-010's **Ship MUST NOT** bullet
"Commit or push directly to `main`", anchored by quoted text rather than by line ordinal, per the
ci.yml ledger's own "by description, not line number" rule. The fix makes Stage symmetric with it.

All three files carry `Generated by autoharness | Template: …` provenance — prose-only correction
is reversible by a future render with no signal.

## 3. Surface

| File | Change | Task |
|---|---|---|
| `tests/integration/directpush_gate_test.go` | **New.** Go regression harness driving the gate; 8 test functions, one per task, plus the **rev-10 per-task activation** helper (§5.0.1). Produced by `harness-architect` at **Ship Step 2**, not by a task (§5.0). | H0 |
| `scripts/testdata/directpush/*.md` + `directpush-manifest.json` | **New.** 14 fixtures + sorted verdict manifest. | A1 |
| `scripts/check-direct-push-language.sh` | **New.** Not-implemented stub (H0) → driver (A2) → detector (A3). | H0, A2, A3 |
| `.github/agents/_orchestrator.agent.md` | Reword Step 1.5 preamble; delete 3e; fold its branch-point into 3a. **Rev 6**: require the Stage handback record in Step 1; make Step 1.5 discovery branch-specific; widen the dirtiness pathspec; add step 4's no-shipment arm. **Rev 7**: make the no-shipment arm verify concrete files (`cat-file -t` = `blob`), pin step 1's pathspec to the fixed artifact-root constant, and re-record the handback commit in 3a. **Rev 8**: capture `stage_base_commit` before invoking Stage and derive the artifact set from the **aggregate** `{stage_base_commit} {stage_head_commit}` two-tree diff. **Rev 10**: preserve a **reported** `stage_outcome` and derive it only when absent, and validate the outcome/`shipment_id` **pair** fail-closed before selecting an arm. | B1 |
| `.github/policies/workflow-policies.md` | P-010 Stage bullets + Amendment Log **1.25.0**. | B2 |
| `.github/agents/_stage.agent.md` | Role Boundary Git row + PR row note. **Rev 6**: emit the staging handback record in the Step 6 summary contract. **Rev 7**: emit `stage_artifact_paths` as concrete file paths. **Rev 8**: emit `stage_base_commit` and the aggregate derivation, and make the `no-shipment` terminal outcome **reachable** in Step 5.5 / Step 6 without weakening P-003 failure handling. **Rev 9**: add the two **operative** steps the permission always presupposed — a pre-mutation **branch gate** (Step 1.9) and a post-mutation **artifact commit** (Step 5.7) — and register both in the Step Sequence Contract. **Rev 10**: make the Step 1.9 deferral rule **categorical and exhaustive** over every tracked pre-gate mutation, and give Step 5.7 an implementable out-of-root **detection** step before it stages anything. | B3 |
| `.github/workflows/ci.yml` | One step in job `lint`; LOCAL DIVERGENCE ledger. | C1 |
| `.github/CODEOWNERS` | Own the new gate + the three unowned gate scripts. | C1 |

**Name change (rev 2)**: `check-direct-push-language.sh`, not `…-policy.sh`. The gate detects
**contract language**, not a push event. The rev-1 name implied behavioural enforcement it cannot
deliver.

**Not touched**: `.copilot/installed-plugins/**` and `.autoharness/staging/**` — gitignored
(`.gitignore:60`, `.autoharness/.gitignore:3`), therefore uncommittable, unreviewable, and
overwritten on reinstall (deliberation D2).

## 4. Ordering (test-first; `main` stays green)

```text
H0  Ship Step 2: harness-architect                → Go harness + stub script; go vet 0, go test RED
A1  fixture corpus + manifest                    → test data only
A2  driver over the fixture corpus               → reject fixtures RED (detector still stubbed)
A3  detector implementation                      → fixtures GREEN, real-tree scan RED
B1..B3  contract correction (4 surfaces)         → real-tree scan GREEN
C1  CI wiring + ledger + CODEOWNERS              → gate blocking from first wiring
C2  coupling verification                        → zero-findings + fixture-backed mutation proof
```

**Rev 10 — the default suite stays green between tasks.** Ship's Step 4.3 runs `go test ./...`
after **every** task, so the harness must not leave later tasks' assertions failing in the default
suite once an earlier task completes. Per-task red is established by **targeted activation**
(§5.0.1); the default suite runs only the assertions whose surfaces exist, plus the
activation-independent anchor. H0's red phase is unchanged and still literal.

The gate enters CI only after the violations are removed, so `main` is never wedged and **no
advisory toggle is introduced** — strictly safer than a bounded advisory window, and simpler.
`blocks` edges: A1→A2→A3→{B1,B2,B3}→C1→C2. **H0 is not a task and carries no `blocks` edge** — it
is Ship's Step 2, which runs once up front for the whole queue and therefore precedes A1 by
construction (§5.0).

---

## 5. Sub-epic A — Failing regression harness

### 5.0 Ship Step 2 harness contract (P-002 / P-004) — rev 4

Rev 3 declared a P-002/P-004 **deviation** and asked Ship to label the tasks `harness-ready` from a
substitute bash harness. That is unexecutable. `_ship.agent.md` Step 2 partitions the queue on the
`harness-ready` label, invokes **`harness-architect`** for every task lacking it, and then halts
unless *every* queued task carries it. `harness-architect` only emits Go `_test.go` harnesses and
only applies the label after `go vet ./...` exits 0 and `go test ./...` is red. A Stage-applied
prose label would therefore be a **forged P-004 postcondition**, and withholding the label would
drive `harness-architect` to invent Go scaffolding for markdown and bash tasks. Rev 4 removes the
deviation instead of widening the policy.

**The harness is a real Go test whose subject is the bash gate.** This is not new scaffolding
invented for the label: `tests/integration/build_script_test.go`, `start_script_test.go`, and
`output_path_guard_test.go` are existing tests in this repository whose entire subject is a
non-Go script, and `internal/apperr/taxonomy_drift_test.go` is an existing Go **drift guard** over
a non-runtime invariant. This gate is the same shape.

**H0 deliverable** — produced by `harness-architect` at Ship Step 2, before any task is claimed:

1. `tests/integration/directpush_gate_test.go`, package `integration`, reusing the existing
   `repoRoot(t)` helper convention. **Eight** test functions, one per task (table below), plus the
   **per-task activation helper and table** of §5.0.1 (rev 10).
2. `scripts/check-direct-push-language.sh` as a **structural stub**: correct shebang,
   `set -euo pipefail`, argument dispatch present, and every mode exiting non-zero after printing
   the marker `not implemented: direct-push detector`. This is the shell analogue of
   `harness-architect`'s `panic("not implemented: <reason>")` production stub — the module
   "compiles" (the script parses and runs) while every test fails for the intended reason.

**Red-phase evidence (P-004 precondition, taken literally; rev 10):** after H0, `go vet ./...`
exits 0 and `go test ./...` exits **non-zero**, because the activation-independent anchor
`TestDirectPushGate_FixtureCorpusIsComplete` runs unconditionally and fails on the missing fixture
corpus. `Compilation: PASS`, `Red Phase: CONFIRMED`. That is the literal postcondition
`harness-architect` checks, and it is satisfied without leaving the other seven functions
simultaneously red (§5.0.1 — the change rev 10 makes, and the reason Ship's Step 4.3 full-suite
gate stops deadlocking task completion). **Each of the remaining seven tasks still has its own
observed red**, taken by Ship at task-claim time from that task's `harness_cmd`, which forces
activation for exactly that task (§5.0.1). Only after the default-suite red does
`harness-architect` apply `harness-ready`. **Stage does not apply that label and this plan does not
ask Ship to.** `018-F.custom_fields.harness_status` stays `pending` until H0 runs.

**Tool resolution — fail, never skip.** The harness resolves `bash` on `PATH` and **fails** the
test when it is absent; it MUST NOT `t.Skip`. A skip would make the red phase unobservable and
silently satisfy P-004 (P-012: record `TOOL_DEGRADED`, never silently skip). This diverges
deliberately from `build_script_test.go`'s `t.Skip("pwsh not available on PATH")`, because `pwsh`
is optional tooling there whereas `bash` is a hard prerequisite of every gate script in this repo.

**This rule is about tool absence, not about activation (rev 10).** §5.0.1's per-task activation
does use `t.Skip`, and the two are not in conflict: a **missing tool** means the assertion could
not be evaluated and its result is unknown — which must fail — whereas an **inactive task** means
the assertion is not yet applicable, because the surface it describes has not been built. The
discriminator is stated so an implementer cannot read one rule as licence to weaken the other: a
`bash` skip hides a failure that exists, an activation skip defers an assertion that has nothing
to assert against yet. MP2 (§5.0.1) is what stops the second from silently becoming the first, by
requiring **all eight** predicates to hold once the release unit has landed.

**Per-task harness map.** Each function is the `harness_cmd` boundary `build-feature` loops on, so
every task has exactly one observable red→green transition. **Rev 10**: every `harness_cmd` carries
the `-gatetask={ID}` activation selector (§5.0.1), which is what makes the transition observable
for a task whose surface does not yet exist.

| Task | Test function | Turns green when |
|---|---|---|
| A1 `018.004-T` | `TestDirectPushGate_FixtureCorpusIsComplete` | 14 fixtures + manifest exist; manifest keys sorted; bijection with the directory holds; 9 reject / 5 accept. **Rev 10**: this is the **activation-independent anchor** — it always runs, which is what keeps H0's default-suite red literal — and it also carries the activation table's static integrity and non-vacuity checks (§5.0.1, MP0/MP1) |
| A2 `018.005-T` | `TestDirectPushGate_SelfTestDriverEnumeratesFixtures` | `--self-test` **enumerates every manifest fixture by name and emits a per-fixture verdict**, and runs fixtures before the real-tree scan — the **driver** contract, which is what this task delivers. **Rev 10 correction**: the function asserts the driver contract and **not** the transient all-accept-stub behaviour, because a permanent assertion that the detector is still a stub would turn **red the moment A3 replaces it** and would wedge the suite from A3 onward. The stub-specific observations (every reject fixture reported failing; `--self-test` exit non-zero) stay **AC-A2.1**, taken from the script's own output at A2's boundary where they are true. The task's completion evidence is **two** exactly-anchored invocations, not one (AC-A2.5) |
| A3 `018.006-T` | `TestDirectPushGate_SelfTestPasses` | `--self-test` exits 0: actual == manifest verdict for all 14, plus both default-branch-resolution assertions |
| B1 `018.007-T` | `TestDirectPushGate_OrchestratorSurfaceClean` | scan scoped to `_orchestrator.agent.md` reports 0 findings **and** (rev 6) Step 1.5's discovery is branch-specific: Step 1 requires the handback fields, the unpushed-commit check resolves the reported branch and compares `origin/main..HEAD`, **no default-branch self-range literal remains in Step 1.5**, and step 4 carries both outcome arms. **Rev 7**: the no-shipment arm verifies concrete files with `cat-file -t` = `blob`; **no directory-prefix path set and no `git show`-over-`{path}` loop remains in that arm**; step 1's dirtiness pathspec is the fixed artifact-root constant; 3a re-records the handback commit. **Rev 8**: Step 1 captures `stage_base_commit` before invoking Stage, the arm asserts base→head ancestry and derives its set from the **aggregate two-tree** diff, and **no single-commit `diff-tree` derivation and no `origin/main..{commit}` range form remains** in that arm. **Rev 10**: a **reported** `stage_outcome` is preserved and derivation from `shipment_id` applies **only when it is absent**, and a fail-closed **pair validation** precedes arm selection |
| B2 `018.008-T` | `TestDirectPushGate_PolicySurfaceClean` | scan scoped to `workflow-policies.md` reports 0 findings **and** the Amendment Log carries row `1.25.0` |
| B3 `018.009-T` | `TestDirectPushGate_StageAgentSurfaceClean` | scan scoped to `_stage.agent.md` reports 0 findings **and** (rev 6) the Step 6 summary contract emits the staging handback record. **Rev 7**: that contract requires `stage_artifact_paths` to be concrete file paths, not directory prefixes. **Rev 8**: it also emits `stage_base_commit` and the aggregate derivation, and Step 5.5 / Step 6 admit a **reachable** `no-shipment` terminal outcome that requires `shipment_id` only for `stage_outcome: shipment` while still halting on P-003 failure. **Rev 9**: the **operative** pre-mutation branch gate and post-mutation artifact commit are both present and **ordered** — the Step Sequence Contract checklist carries a branch-gate entry ordinally **before** the deliberation step and a commit entry ordinally **after** stash archival and **before** the summary step, each backed by a step section carrying its mandatory clauses (§6.3.1). **Rev 10**: the branch-gate section's deferral clause is **categorical** over every tracked pre-gate mutation and enumerates its three classes, and the commit section carries the NUL-safe full-tree **out-of-root detection** clause positioned ordinally **before** its staging clause |
| C1 `018.010-T` | `TestDirectPushGate_CIWiringIsBlocking` | job `lint` carries the task-ID-suffixed step with no `continue-on-error` and no toggle; `ci-gate` still transitively needs `lint`; every script invoked by a `ci.yml` step has a `CODEOWNERS` owner line. **Rev 5**: the step-presence check is **YAML-structural** over `jobs.lint.steps[]` and its non-vacuity is proven (AC-C1.6) |
| C2 `018.011-T` | `TestDirectPushGate_FullCorpusClean` | bare full-corpus scan exits 0 with 0 findings and the three §5.3 marker-free constructs score **accept**. **Rev 5**: this runs the detector from `go test`, so it keeps executing even if the job-`lint` step is removed by a render (AC-C2.8). **Rev 10**: it also carries the terminal-completeness check (§5.0.1, MP2) — once C1's surface exists, **all eight** activation predicates must hold, so a regression can never degrade into a silent skip |

The per-surface scoping of B1/B2/B3 is what gives each of those three tasks an independent
transition; a whole-corpus assertion would only go green after all three landed and would trip
`build-feature`'s 5-attempt circuit breaker on the first two.

**Regeneration-resistant enforcement (rev 5 — R5 closure).** `.github/workflows/ci.yml` is
autoharness-generated, so the job-`lint` gate step added by C1 is a LOCAL DIVERGENCE sitting
**inside** the render replacement boundary, exactly like the five divergences already enumerated in
that file's own ledger. A render can drop it. The control is nevertheless **not** silently
disableable by regeneration, because enforcement also runs through this harness, and every link in
that path is either untouched by a render or **re-emitted** by one:

| Link | Artifact | Render behaviour |
|---|---|---|
| Detector | `scripts/check-direct-push-language.sh` | **Not generated** — hand-authored, like `check-retired-architecture.sh`, `check-write-path-precondition.sh`, `check-unignore-regression.sh`, `check-depguard-fixtures.sh` (only 5 of the 10 `scripts/*.sh` carry autoharness provenance). Survives |
| Assertion | `tests/integration/directpush_gate_test.go` | **Not generated** — no autoharness template covers `tests/**`. Survives |
| Invocation | step `Test (race)` → `go test -race -mod=readonly ./...`, in job `expensive` (`name: test`) | **Generated template baseline** — one of the core four jobs in this file's header, carrying **no** task-ID comment and absent from the LOCAL DIVERGENCE list. A render *re-emits* it |
| Trigger | job `expensive` runs when `changes.outputs.code == 'true'` | The `code` filter is a **denylist** — `'**'` minus `docs/**`, `.backlogit/**`, `.backlog/**`, `.autoharness/**`. `.github/workflows/**` is not excluded |

The fourth row is what makes this airtight rather than lucky: the *same* edit that removes the
job-`lint` step is itself a `.github/**` change, so it necessarily sets `code == 'true'` and
**cannot skip the job that runs the harness**.

Two harness functions carry the invariant, and the pair is precisely what R5 now rests on:

* `TestDirectPushGate_FullCorpusClean` (C2) re-runs the **detector** over the installed surfaces.
  Requirement **(a)** — forbidden direct-push language is absent — therefore still holds with the
  job-`lint` step gone.
* `TestDirectPushGate_CIWiringIsBlocking` (C1) asserts the job-`lint` step **still exists**.
  Requirement **(b)** — the enforcement invocation itself survives — therefore fails **red** on
  render removal instead of passing silently.

**The step-presence assertion MUST be YAML-structural, never a text search.** C1 also adds a ledger
entry to `ci.yml` that *names* `check-direct-push-language.sh` in prose, so a file-wide
`grep 'check-direct-push-language.sh' .github/workflows/ci.yml` would match that comment and keep
passing after the step itself was deleted. That is the exact self-matching failure already recorded
in `docs/compound/2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md` and warned
about at length in `ci.yml`'s own `011.002-T` step comment. The assertion parses the workflow and
walks `jobs.lint.steps[]`; its non-vacuity is proven mechanically (AC-C1.6).

**What this does NOT claim.** `CODEOWNERS` is **advisory metadata only** in this repository — its
own header records that it has no enforcement effect until code-owner review is required on `main`,
a deliberate operator action explicitly not taken here. It is therefore **not** part of this
guarantee (R6 continues to name it only as an ownership prerequisite). The guarantee is exactly and
only §1's: *a regeneration cannot **silently** reopen the gap.* Disabling the control still requires
editing a non-generated tracked file in a reviewable pull-request diff — a visible act, not a render
artifact. That narrower residual is recorded as **R12**, not hidden.

**Why not a separate workflow.** `secret-scan-history.yml` is the repo's one hand-authored,
non-generated workflow and would also survive a render — but it is `schedule`/`workflow_dispatch`
only and its header declares it **NOT A PR-REQUIRED CHECK** by design, so it cannot carry a blocking
merge gate without inverting its stated contract. `ci-topology-check.sh` is itself generated
(`ci/ci-topology-check.sh.tmpl`), so it sits inside the same boundary. The harness route adds **no
new file, no new workflow, no new CI step, no new ledger entry, and no new CODEOWNERS line** — it
only strengthens assertions in a file H0 already produces. It is the smallest sufficient mechanism.

**Why not the alternative.** Defining a first-class non-Go harness route would require amending
P-002 and P-004 in `workflow-policies.md`, Step 2 in `_ship.agent.md`, and the `harness-architect`
skill. Those last two are autoharness-generated with gitignored `.tmpl` sources (D2), so the change
would be silently reverted by the next render — the exact failure mode R4/R5 already document — and
amending P-002/P-004 is a **workspace-wide weakening of test-first** affecting every future
shipment, far outside this release unit's branch-policy contract and outside the governing
deliberation. Rejected on both counts.

#### 5.0.1 Per-task activation (rev 10) — the harness lifecycle vs Ship's full-suite gate

**The defect.** Rev 4–9's §5.0 required that after H0 **all eight** functions fail, and left them
failing until their task landed. `_ship.agent.md` Step 4.3 runs the **full test suite**
(`go test ./...`) as a quality gate after *every* task's `build-feature` success. So the moment A1
went green, the seven future-task functions were still intentionally red, Step 4.3 failed, and the
task could never be marked complete — the shipment deadlocks on its second gate. The harness was
correct per-task and unexecutable per-shipment.

**The mechanism — one selector flag and one activation table.** The harness declares a single
package-level test flag and a single table:

```go
var gateTask = flag.String("gatetask", "", "activate assertions for exactly one task ID")

// one row per task; consumed by gate(t, id) at the top of each test function
var gateSurfaces = map[string]gateSurface{ /* 8 rows, see table below */ }
```

`gate(t, "B3")` is the **first statement** of each test function and resolves one of three
outcomes:

| Condition | Outcome |
|---|---|
| `-gatetask` names this function's task ID | **activate** — assertions run regardless of surface state (this is the observed red before implementation) |
| `-gatetask` names a *different* valid task ID | `t.Skip` |
| `-gatetask` is empty (**default mode**) | activate **iff** this task's surface predicate holds; otherwise `t.Skip` with a reason naming the task ID and the missing surface |
| `-gatetask` names an ID absent from the table, or the function's own ID is absent | **fail** (never skip) |

`A1` is the one **activation-independent anchor**: its `gate` call always activates. That is what
keeps H0's `go test ./...` red literal rather than vacuously green, since at H0 no surface exists
and every other function skips.

**Why a Go test flag rather than an environment variable.** A flag is invoked identically from
`bash`, `pwsh`, and `cmd`; `VAR=value go test …` is a POSIX-shell-only construct and
`$env:VAR=…` is PowerShell-only, so an env-var selector would make `harness_cmd` shell-specific.
**Verified on this platform**: the flag name must be **dot-free** — `go test . -run '^T$'
-gatetask=A1` parses and reaches the test binary unquoted, whereas a dotted name such as
`-directpush.task=A1` is split by PowerShell's tokenizer and rejected by the binary as
`flag provided but not defined: -directpush` unless individually quoted. `-gatetask` is therefore
the portable form, and the plan names it explicitly rather than leaving it to the implementer.

**Activation table.** Each predicate is a narrow presence probe over the task's own deliverable —
deliberately weaker than the function's assertions, which is what lets a landed-but-regressed
surface still be *caught* (by MP2) rather than merely skipped:

| Task | Surface predicate (default mode activates when…) |
|---|---|
| A1 | *(activation-independent — always active)* |
| A2 | `scripts/check-direct-push-language.sh` no longer emits the H0 marker `not implemented: direct-push detector` for `--self-test` |
| A3 | that script's `--self-test` exits **0** |
| B1 | `.github/agents/_orchestrator.agent.md` contains the handback field token `stage_base_commit` |
| B2 | `.github/policies/workflow-policies.md` Amendment Log contains row `1.25.0` |
| B3 | `.github/agents/_stage.agent.md` Step Sequence Contract contains a `Step 1.9` entry |
| C1 | **YAML-structural**: `jobs.lint.steps[]` contains a step invoking `check-direct-push-language.sh` (never a text search — the AC-C1.6 self-matching-ledger hazard) |
| C2 | C1's predicate holds (C2 produces no file of its own; its deliverable is evidence) |

**Mutation-proof checks.** The risk this mechanism introduces is a harness that skips everything
forever. Four checks close it, and they are the whole of the added machinery:

* **MP0 — table integrity.** `gateSurfaces` has exactly **eight** rows, whose IDs are exactly the
  eight task IDs of the map above, and every function's own ID resolves to a row. A function whose
  ID is missing **fails**. Carried by the activation-independent A1 function, so it runs from H0
  onward and in every default-mode run.
* **MP1 — predicate non-vacuity, executed.** Every predicate is evaluated against an empty
  `t.TempDir()` root and MUST return **false**. A row rigged constant-true therefore fails rather
  than silently activating. Same no-dirty-tree-edit rule as AC-A2.4/AC-C2.2 — the probe runs
  against a scratch root, never against the real tree. Also carried by A1.
* **MP2 — terminal completeness.** Inside `TestDirectPushGate_FullCorpusClean` (C2), **all eight**
  predicates MUST hold. C2 activates once C1's surface exists, so from the moment the release unit
  has landed a regression in *any* surface turns a would-be skip into a **failure**. This is the
  check that makes "skipped" a build-time state only, never an end state.
* **MP3 — visible skips, strict selector.** Every skip is a `t.Skip` naming the task ID and the
  unmet predicate, so the state is greppable in `go test -v` rather than silent; and an unknown
  `-gatetask` value **fails** rather than skipping, so a typo in a `harness_cmd` cannot turn the
  whole suite green.

**What this preserves, point by point.**

| Requirement | How it holds |
|---|---|
| (a) harness-architect produces the **complete** file before implementation | Unchanged — H0 still emits all eight functions, the helper, and the table. No task creates a test |
| (b) every task has an observed targeted **red→green** transition before its implementation | Each `harness_cmd` carries `-gatetask={ID}`, which forces activation regardless of surface, so the function is red at claim time and green after implementation. Ship's `build-feature` boundary is unchanged |
| (c) future-task expected-red cases do not fail the **default** suite | Default mode skips a task whose surface is absent, so Step 4.3's `go test ./...` is green between tasks |
| (d) the final `go test ./...` runs all eight **real** assertions green | At the end state every surface exists, so all seven predicates hold and the anchor always runs — eight real assertions, and MP2 asserts exactly that |

**Harness-ready semantics stay honest.** `harness-ready` continues to mean what
`harness-architect` verifies: the file compiles (`go vet ./...` = 0) and the default suite is
**red** (`go test ./...` ≠ 0) before any implementation. It does **not** claim that all eight
assertions were simultaneously red — and rev 10 says so explicitly rather than letting the label
imply it. The per-task red is a **per-task** obligation, discharged by Ship at claim time from that
task's `harness_cmd`, and it is exactly the evidence `build-feature` already records.

**A latent defect this mechanism exposed, recorded rather than silently repaired.** Requirement (d)
forced a check that revisions 4–9 never made: *is every function's assertion still true at the end
state?* Seven are. **A2's was not.** Its map row asserted that `--self-test` reports every reject
fixture **failing against the all-accept stub** with a non-zero exit — a property that is true only
while A2's stub is the current implementation and becomes **false the moment A3 installs the real
detector**. The function would have gone red at A3 and stayed red, wedging `go test ./...` for the
rest of the shipment and failing (d) outright. This had nothing to do with activation; activation
merely made it visible, because "all eight green at the end" had never been stated as a requirement
before. The fix is to assert the **durable** property A2 actually delivers — the driver enumerates
every manifest fixture by name, emits a per-fixture verdict, and runs fixtures before the real-tree
scan — and to leave the stub-specific observations where they belong, as **AC-A2.1**, read from the
script's own output at A2's boundary. **AC-A2.1 is not weakened**: it still demands the observed red
for the detector, still binds `--self-test` only, and still reserves bare-mode red for AC-A3.3. What
changes is only that it is not *also* encoded as a permanent Go assertion that contradicts A3. The
defect class — *a regression assertion pinned to a transitional state* — is worth recognizing again.

**What was rejected.** (1) *Weakening Ship's Step 4.3* to a scoped `go test ./tests/integration
-run …` — that amends `_ship.agent.md`, which is autoharness-generated with a gitignored template
(D2), and weakens the full-suite quality gate for every future shipment; explicitly out of scope
(§11). (2) *Skipping every function until its task lands, with no anchor* — H0's `go test ./...`
would then exit **0** and P-004's red phase would be unobservable, which is the silent-skip failure
P-012 and §5.0's own `bash`-resolution rule exist to forbid. (3) *A build-tag or generated
per-phase file* — that is a new runtime framework, needs a file per phase, and moves state out of
the single table where MP0/MP1 can check it. (4) *Deriving activation from backlog item status* —
couples the test suite to `.backlogit/` mutable state, and a status edit would silently disable
assertions. (5) *Mode-scoped assertions* — letting targeted mode assert strictly more than default
mode would have preserved A2's stub-specific check at its boundary, but it adds a second behavioural
axis to the helper for one task's transitional property. The durable-assertion fix above achieves
the same outcome with no extra mechanism.

### 5.1 House-style contract

Matches `check-retired-architecture.sh` (closest precedent) on `::notice::` mode emission and the
fixture/manifest harness; matches retired-arch / write-path / unignore on the `python3`→`python`
probe, `ROOT="$(git rev-parse --show-toplevel)"`, and exit codes **0** pass / **1** violation /
**2** invalid invocation. Deliberately does **not** adopt `check-gitignore-append-only.sh`'s
`SCRIPT_DIR` + `--base-ref` shape, because this gate asserts a whole-tree invariant, not a diff.
`set -euo pipefail`. **No waiver mechanism.** Header states INVARIANT, detector scope, THREAT
MODEL, and ACCEPTED RESIDUALS.

**Two modes only** (rev 2): bare (repo scan) and `--self-test` (fixtures + selection assertions,
**then** real-tree scan). The rev-1 third mode is dropped — `--self-test-integrity` exists in
retired-arch solely because its verdict step is *toggled* by `RETIRED_ARCH_GATE_ADVISORY`; §7.1
has no toggle, so the split would be semantically identical to one `--self-test` call. Running
fixtures *before* the real scan preserves the `gitignore-append-only` rationale: a broken checker
reports "self-test failed", never a false violation.

### 5.2 Corpus (widened in rev 2)

`git ls-files` pathspec, markdown only:

`.github/agents/**`, `.github/policies/**`, `.github/skills/**`, `.github/instructions/**`,
`.github/prompts/**`, `.github/copilot-instructions.md`, `AGENTS.md`

`scripts/**` and `docs/**` are outside the pathspec **by construction** — there is no exclusion
predicate, so rev-1's "assert the exclusion" selection checks are dropped as vacuous. `docs/**`
being out of corpus is what lets the deliberation and this plan quote 3e freely.

**ACCEPTED RESIDUAL** (script header): `.github/workflows/**` and `.github/constraints/**` are not
scanned (the latter holds no markdown). R4 is therefore scoped to "a regeneration of the scanned
surfaces".

### 5.3 Detector (rev 3 — per-construct, structure-aware, closed construct set)

Rev 1's line-scoped in-line-marker heuristic was **unsound in both directions** and is replaced.

**Verdict granularity — per construct, never per block (rev 3).** Every matched construct is
scored and reported **individually**. A block is never accepted as a whole. This matters because
after normalization, Step 1.5's item 3 becomes one block containing *both* 3b (`create a PR to
\`main\`` — accept) and 3e (`Attempt a direct push to \`main\` first` — reject); a block-level
accept verdict would mask the reintroduced violation.

**Block normalization (rev 3).** Continuation lines are joined into logical blocks before
matching, covering **both** well-formed nested list items **and lazy paragraph continuations** —
the latter is the shape Step 1.5's a–e sub-steps actually use (see §11), i.e. precisely the file
the gate exists to protect. A re-render that wraps a construct across two lines cannot evade it.

**Violation constructs** (closed set — two shapes):

1. **Push** — an imperative push **verb-anchored on a push** whose *target* binds to a
   default-branch token. Excludes PR-base forms (`PR to`, `pull request to`, `--base`) and
   read/checkout verbs (`fetch`, `log`, `show`, `diff`, **`pull`**, **`checkout`** — rev 3; the
   corpus legitimately contains 3d's "pull `main`", P-011's `git checkout main; git pull`, and
   Ship Step 0.5's "Switch to the default branch"). This is the discrimination rev 1 never
   specified: 3b `Push the staging branch and create a PR to \`main\`` is **accept**
   (target = branch); 3e `Attempt a direct push to \`main\` first` is **reject** (target = `main`).
2. **Commit authorization** — text **permitting** artifacts to be committed/persisted *on* a
   default-branch token. **Verificational / postcondition phrasing is not a violation** (rev 3):
   "verify/confirm that X **has reached** the default branch" asserts a state; "artifacts may be
   committed **on** the default branch" grants a permission. Only the permissive form is a
   finding. This shape catches `_stage.agent.md` L42 and the P-010 bullet — which rev 1 missed
   entirely — without flagging B1's reworded preamble.

**Table handling is CELL-scoped, not row-scoped (rev 3).** A Role Boundary row is one line whose
Allowed cell holds verbs ("commit, push") and whose Forbidden cell holds the default-branch token.
Row-scoped matching would pair them and fire on every correct row. Constructs are matched **within
a single cell**, and the governing context is that cell's **column header**.

**Default-branch token** = alternation over: the configured default branch, the literal `main`,
the literal `{{DEFAULT_BRANCH}}`, and the prose "default branch". The placeholder form closes the
regeneration-evasion path named in §5.0 and §9.

**Default-branch resolution must not abort the gate (rev 3).** `git symbolic-ref
refs/remotes/origin/HEAD` is **not** created by `actions/checkout`, so under the mandated
`set -euo pipefail` an unguarded command substitution would exit non-zero in CI and be
indistinguishable from a real violation — wedging job `lint` the moment C1 wires it. Resolution
MUST be explicitly guarded (`|| true`, empty-result check) with a documented fallback to `main`,
and `--self-test` MUST assert both the resolution-failure fallback and a non-`main` default-branch
name.

**Prohibition resolution — structural, not in-line-only.** A construct is **accepted** when either:

- its **governing context** is prohibitive — (a) a **bold lead-in line** of the form
  `**<actor> MUST NOT**:` or `**Forbidden**` immediately preceding the list (rev 3 — this, not an
  ATX heading, is the shape `workflow-policies.md` actually uses for both the Ship
  "Commit or push directly to `main`" bullet and the Stage bullet B2 adds; reading "heading" as
  ATX-only would flag both and make zero-findings unreachable), (b) an ATX `MUST NOT` / `Forbidden`
  heading, or (c) a table cell under a `Forbidden` / `MUST NOT` **column header**; **or**
- a prohibition marker **precedes** the construct within the same clause:
  `never`, `must not`, `do not`, `don't`, `no`, `forbidden`, `prohibited`.

`rather than` and `instead of` are **removed** from the marker set — they are directional, not
prohibitive, and would have excused "push directly to `main` rather than opening a PR".
Marker-must-precede closes the "…first; do not create a PR unless rejected" fail-open.

Structural resolution is what makes the real corpus pass: P-010's **Ship MUST NOT** bullet
"Commit or push directly to `main`" in `workflow-policies.md` and `_ship.agent.md`'s Role
Boundary `| Git |` row carry **no in-line marker** — their prohibition
lives in the bold lead-in and the `Forbidden` column header respectively. Rev 1 would have flagged
both, making zero-findings unreachable.

### 5.4 Task A1 — fixture corpus + manifest (test data only)

**Files**: `scripts/testdata/directpush/` (14 fixtures), `scripts/testdata/directpush-manifest.json`.

| Fixture | Verdict | Locks |
|---|---|---|
| `directpush-reject-attempt-first.md` | reject | literal 3e construct |
| `directpush-reject-push-origin-main.md` | reject | `git push origin main` form |
| `directpush-reject-default-branch-prose.md` | reject | "default branch" phrasing |
| `directpush-reject-branch-placeholder.md` | reject | literal `{{DEFAULT_BRANCH}}` form |
| `directpush-reject-commit-on-default.md` | reject | **commit-shaped** permission (the L42 shape) |
| `directpush-reject-marker-bearing.md` | reject | violation with a **trailing** `do not` clause |
| `directpush-reject-rather-than.md` | reject | "push directly to `main` **rather than** opening a PR" — proves the pruned markers do not fail open |
| `directpush-reject-wrapped-construct.md` | reject | construct split across a line break **in lazy-paragraph-continuation shape** (the real Step 1.5 shape) |
| `directpush-reject-mixed-block.md` | reject | one normalized block holding an accepted **and** a violating construct — proves per-construct verdicts |
| `directpush-accept-prohibition-boldleadin.md` | accept | `**Stage MUST NOT**:` **bold lead-in** bullet, marker-free (the real P-010 shape) |
| `directpush-accept-prohibition-tablecell.md` | accept | `Forbidden` **column-header** table cell, marker-free (the real `_ship.agent.md` Git row shape, incl. an Allowed cell with push verbs to prove cell-scoping) |
| `directpush-accept-marker-precedes.md` | accept | "**never** commit directly to the default branch" — the in-line marker-precedes branch |
| `directpush-accept-read-and-pr-forms.md` | accept | `git fetch origin main`, `git log origin/main..main`, `git show origin/main:…`, `git checkout main`, `pull \`main\``, and 3b's `create a PR to \`main\`` |
| `directpush-accept-verification-form.md` | accept | "verify that artifacts **have reached** the default branch" — the postcondition shape B1 installs |

**Manifest is authoritative** for verdicts, with **sorted keys** (Principle IX). `--self-test`
asserts filename prefix (`directpush-reject-` / `directpush-accept-`) and manifest verdict
**agree**, and asserts **bijection** (every file on disk is in the manifest and vice versa),
closing rev 1's silently-unexercised-fixture hole.

**The corpus is deliberately NOT widened for the rev-6 defect.** `directpush-accept-read-and-pr-forms.md`
already carries `git log origin/main..main` as an **accept** case, and it stays exactly as it is:
its job is to prove the detector does not flag *read* verbs, and that property is unrelated to
whether the Orchestrator happens to use that range. The rev-6 defect — **main-only discovery** —
is a different defect class from the two §5.3 violation constructs (it is neither a push to the
default branch nor a permission to commit on it), so folding it into the detector would mean
opening the **closed construct set**, adding fixtures, and re-deriving A1's 14/9/5 counts through
A1→A2→A3→C2. That is a materially larger change than the finding requires. The rev-6 regression is
instead rejected mechanically by the **B1 harness function**, which already exists, is already
scoped to `_orchestrator.agent.md`, and is already B1's red→green boundary (§5.0). No fixture, no
manifest entry, and no construct-set change.

Each fixture carries a single leading H1 so **P-008 markdownlint** (`MD041`/`MD025`) passes.

**Acceptance criteria**

- **AC-A1.1** 14 fixtures exist; manifest keys are sorted and in bijection with the directory.
- **AC-A1.2** Every manifest verdict agrees with its fixture's filename prefix
  (`directpush-reject-*` / `directpush-accept-*` — the actual on-disk names, rev 8), and both
  classes are non-empty: 9 reject, 5 accept. **Static data check only** —
  A1 authors fixture data and owns no executable logic, so the *executed* non-vacuity proof of
  these assertions is **AC-A2.4 on task A2**, which owns the fixture-suite driver (rev 4 — review round 3).
- **AC-A1.3** `markdownlint "scripts/testdata/directpush/**"` exits 0.

### 5.5 Task A2 — fixture-suite driver over the stubbed detector

**Files**: `scripts/check-direct-push-language.sh` (replace the H0 stub's mode dispatch with the
real driver: mode dispatch, fixture/manifest harness, selection assertions, `::notice::` emission).
The **detector remains stubbed** — A2 owns no detection logic, so the stub is narrowed from
"everything fails with the not-implemented marker" to "every construct classified `accept`", and
A3 replaces it.

**Scope change (rev 4)**: A2 no longer creates the harness. The Go harness and the not-implemented
stub script are H0 deliverables produced by `harness-architect` at Ship Step 2 (§5.0), because
Step 2 runs once up front for the whole queue — a *task* whose content is "create the harness" is
structurally unreachable in Ship's flow. A2 keeps the driver work, which is genuine implementation.

- **AC-A2.1** Every **reject** fixture is reported **failing** against the all-accept stub, by
  name, and **`--self-test` exits non-zero**. This is the **observed red for the detector itself**
  — rev 1 never took the detector red. **Rev 8 scope correction**: the non-zero requirement binds
  **`--self-test` only**. Bare-scan exit is deliberately **not** constrained here, because A2's
  stub classifies every construct `accept`, so a bare scan correctly reports zero findings and
  exits **0** — demanding red there would make AC-A2.1 unsatisfiable by a correct A2
  implementation, or satisfiable only by a stub that lies about the real tree. Bare-mode red is
  A3's obligation once a real detector exists (**AC-A3.3**), and the A2→A3 dependency edge
  guarantees it is evaluated after this one. **Rev 10 clarification**: this criterion is verified
  from the **script's own `--self-test` output** at A2's boundary, and is deliberately **not**
  encoded as a permanent assertion in `TestDirectPushGate_SelfTestDriverEnumeratesFixtures` — a Go
  assertion that the detector is still a stub would go **red the moment A3 replaces it** and would
  wedge `go test ./...` for the rest of the shipment (§5.0.1). The criterion itself is unchanged
  and unweakened; only its carrier is named.
- **AC-A2.2** `--self-test` runs fixtures **before** the real-tree scan, so a broken checker
  reports "self-test failed", never a false violation.
- **AC-A2.3** `git diff --exit-code -- .github/workflows/ci.yml` is clean (no CI change yet).
- **AC-A2.4** Bijection and prefix/manifest-agreement assertions each **fail** when deliberately
  violated — non-vacuity proven by **execution**, not asserted. The proof is performed against a
  **temporary fixture copy under `t.TempDir()`**, never by editing the committed fixture tree
  (same no-dirty-tree-edit rule as AC-C2.2). Relocated here from A1 (rev 4 — review round 3): A1
  owns fixture data only and cannot execute an assertion, while this task owns the driver that
  runs it. The existing A1 → A2 dependency edge already guarantees the fixtures exist before this
  criterion is evaluated, so execution order is unaffected.
- **AC-A2.5** **Two separate, exactly-anchored invocations (rev 10).** A single driver-scoped
  `-run` command cannot establish anything about a test it does not run, so the evidence is split:
  1. `go test ./tests/integration -run '^TestDirectPushGate_SelfTestDriverEnumeratesFixtures$'
     -gatetask=A2` exits **0**.
  2. `go test ./tests/integration -run '^TestDirectPushGate_SelfTestPasses$' -gatetask=A3` exits
     **non-zero**, **and** its output carries `--- FAIL: TestDirectPushGate_SelfTestPasses` together
     with the expected **detector-stub** reason — A2's stub classifies every construct `accept`, so
     the reject fixtures' actual verdicts disagree with the manifest. An arbitrary failure does
     **not** satisfy this: a compile error, an absent `bash`, a missing fixture corpus, or a
     mis-typed `-gatetask` value is a harness defect, not the observed red, and each is rejected by
     name.

  Both regexes are **anchored** (`^…$`). The `--- FAIL:` requirement is the non-vacuity guard and is
  load-bearing: **verified on this platform**, `go test -run '^TestNoSuchTest$'` prints
  `ok … [no tests to run]` and exits **0**, so "exits non-zero" alone could be satisfied — or
  silently *unsatisfiable* — by a pattern that selects nothing. **Bare-mode red stays A3's
  obligation** (AC-A3.3) and is not asserted here; P-004 is not weakened in either direction — this
  criterion adds an observation, it removes none.

### 5.6 Task A3 — detector implementation

**Files**: `scripts/check-direct-push-language.sh` (replace stub with the §5.3 detector).

- **AC-A3.1** `--self-test` reports every fixture by name with actual == manifest verdict.
- **AC-A3.2** Guarded default-branch resolution is asserted for **both** the resolution-failure
  fallback to `main` and a non-`main` configured default branch.
- **AC-A3.3** Bare invocation exits **1** and names all four §2 surfaces — identified by **file and
  matched construct text**, not line ordinal.
- **AC-A3.4** The **first full-corpus scan output** — resolved corpus file count and the complete
  enumerated finding list — is recorded before B1 begins, so any additional surface is discovered
  now rather than at C2.
- **AC-A3.5** A `::notice::` names the resolved mode on every invocation.

---

## 6. Sub-epic B — Contract correction

### 6.1 Task B1 — Orchestrator Step 1.5

B1 corrects `.github/agents/_orchestrator.agent.md` on **two axes**. Revisions 1–5 corrected what
Step 1.5 **authorizes** (§6.1.1). Rev 6 corrects how Step 1.5 **discovers and verifies** the work
it gates (§6.1.2) — without which the authorization correction is unexecutable.

#### 6.1.1 Authorization surfaces (rev 1–5, unchanged)

Delete sub-step **3e** in full (the "Branch protection handling" block and its three nested
bullets plus trailing rationale).

**Fold its branch-point into 3a** — required, because 3e's "Create branch
`chore/stage-{shipment_id}` **from the current commit**" is the **only** instruction covering the
*already-committed-but-unpushed* path that Step 1.5 step 2 routes into step 3. 3a currently
specifies no branch point. Revised 3a (rev 6 extends this to reuse the reported branch — see
§6.1.2):

> a. Check out the Stage artifact branch `$STAGE_BRANCH` reported by the handback record — or,
>    when none was reported, create `chore/stage-{scope-slug}` **from the current commit** — and
>    commit any uncommitted backlog and
>    planning files to it. **Then re-record the handback commit** (rev 7 — correction 5, §6.1.2):
>    set `stage_head_commit` to `git rev-parse HEAD` and discard any `stage_artifact_paths` the
>    handback reported. **Preserve `stage_base_commit` unchanged** (rev 8), or — when it is still
>    UNRESOLVED — set it to the `HEAD` captured **before** this commit, so step 4 derives its file
>    set from the whole `{stage_base_commit}`→`{stage_head_commit}` range rather than from the tip
>    commit alone.

**Rewrite the Step 1.5 preamble** (rev 3 — the fourth surface, §2). It currently states the gate's
postcondition as artifacts being "committed to the default branch", which remains a compliant
reading of direct commit-to-default. Revised (rev 6 widens the scope clause so a shipment-less run
is inside the gate rather than outside it — see §6.1.2):

> After Stage completes — **whether or not a shipment was formed** — and before any routing to
> Ship, verify that all staging artifacts (backlog items, shipment manifests, and the plan,
> deliberation, and memory records Stage wrote) **have reached the default branch via a merged
> staging PR** and are present on the remote.

This is a **postcondition/verification** form, which §5.3 construct 2 explicitly does not treat as
a violation — so it satisfies the gate without narrowing the detector.

**Preserved unchanged**: 3b–3d; step 4's `git show origin/main:.backlogit/queue/{shipment_id}.md`
verification and `STAGING_GATE_FAIL` halt; the verified-on-origin broadcast. Step 1.5 remains
NON-NEGOTIABLE and **Orchestrator-owned**.

#### 6.1.2 Branch discovery and two-outcome verification (rev 6)

**The defect.** Step 1.5's two discovery inputs are both blind to the branch Stage actually commits
on, once B2/B3 require Stage to create, check out, and commit to `chore/stage-{scope-slug}` and
leave the worktree there:

| Step 1.5 input | Existing form | Why it misses the Stage branch |
|---|---|---|
| step 1 — dirtiness | `git status --short -- .backlogit/` | A no-shipment Stage run writes only `docs/plans/`, `docs/decisions/`, `docs/memory/` — the exact `fdff9e4` shape. The pathspec reports **clean**. |
| step 2 — unpushed commits | `git log origin/main..main` | Compares the **default branch** against its own remote-tracking ref. Stage's commit is on `chore/stage-…`, which is not the default branch, so the range is **empty** by construction. |

Both inputs report "synchronized", Step 1.5 skips step 3 entirely, and step 4 then fails its
`origin/main` manifest check — or, for a shipment-less run, has no manifest to name at all. The
**no-shipment persistence route documented end to end in §6.2 therefore cannot complete**. The
authorization fix and the discovery fix are not separable: correcting only the former relocates
Stage's commit to a branch the gate cannot see.

**Correction 1 — the Stage artifact branch becomes an explicit handback.** Step 1 (Route to Stage)
currently declares Stage's output as "a `shipment_id` in `queued` status" only. Revised bullet
under step 3, and revised step 4:

> * Stage's expected output — the **staging handback record**: `stage_branch` (the dedicated Stage
>   artifact branch Stage created and committed on), `stage_base_commit` (rev 8 — the commit the
>   Orchestrator captured **immediately before invoking Stage**; see below), `stage_head_commit`,
>   `stage_outcome` (`shipment` or `no-shipment`), `stage_artifact_paths` (the **concrete
>   repository-relative file paths** Stage wrote and committed between `stage_base_commit` and
>   `stage_head_commit` — file paths only, never directory prefixes; rev 7/8), and — when
>   `stage_outcome` is `shipment` — a `shipment_id` in `queued` status.

> **Before invoking Stage** (rev 8), capture and **retain** the pre-invocation tip:
> `stage_base_commit := git rev-parse HEAD`. This value is the Orchestrator's own, taken before
> Stage can mutate anything, and it is **authoritative**. Retaining it is what makes the
> verification in step 4 *aggregate*: every commit Stage creates during the invocation is a
> descendant of it, so one two-tree diff covers a Stage branch of **any** commit count.

> 4. Receive Stage's output: record the staging handback record, and the `shipment_id` when one
>    was returned. Apply these defaults for any field Stage did not report, recording
>    `STAGING_HANDBACK_DEGRADED: Stage did not report {fields} — defaulted` once and surfacing it
>    to the operator:
>    - `stage_branch` / `stage_head_commit`: resolve from the current checkout
>      (`git rev-parse --abbrev-ref HEAD`, `git rev-parse HEAD`). If that branch is the default
>      branch, leave `stage_branch` **UNRESOLVED** — **never** substitute the default branch for
>      the Stage artifact branch.
>    - `stage_base_commit` (rev 8): use the **retained pre-invocation value** — it wins over any
>      reported value. Use the reported value only when no pre-invocation value was retained (a
>      direct Stage invocation the Orchestrator did not bracket). When neither exists, leave it
>      **UNRESOLVED**; resolution still does not halt, and step 4's verification rejects the
>      UNRESOLVED case loudly.
>    - `stage_outcome`: **preserved when Stage reported one** (rev 10). Derive it from whether a
>      `shipment_id` was returned **only when the field is absent from the handback**. A reported
>      outcome is never overwritten by the derivation, and a reported `shipment` outcome is
>      **never** silently rewritten to `no-shipment` because no `shipment_id` reached this step —
>      that case is an inconsistent pair and is rejected by the validation below, not defaulted
>      away.
>    - `stage_artifact_paths`: leave **empty** here and derive it in step 4 from the
>      `stage_base_commit`→`stage_head_commit` aggregate (correction 4, check 4). It is **never**
>      defaulted to a directory list (rev 7).

**Two path notions, deliberately distinct (rev 7).** `STAGE_ARTIFACT_ROOTS` is a **fixed constant**
of the gate — `.backlogit/`, `docs/plans/`, `docs/decisions/`, `docs/memory/` — and is *not* a
handback field. It serves two roles and only two: it is step 1's dirtiness pathspec, and it is the
containment allow-list every concrete path in step 4 must lie under. `stage_artifact_paths` is a
handback field and is always a set of **concrete file paths**. Rev 6 conflated the two, using one
field for both jobs; that conflation is what let directory prefixes reach a verification loop that
had no way to reject them. Directory prefixes are correct in step 1 — it runs **before** the commit
exists, so there is no commit to derive files from — and are rejected in step 4, where the commits
do exist and the files they changed are knowable exactly.

**Two commit notions, also deliberately distinct (rev 8).** `stage_base_commit` is the
Orchestrator's **pre-invocation** tip; `stage_head_commit` is the Stage branch tip **after** the
invocation. Rev 7 carried only the latter and derived its file set from that one commit's own diff
against its parent, which silently assumed **a Stage session commits exactly once**. That
assumption is false in the general case and false in this very repository: the branch carrying this
plan has accumulated many commits. Under rev 7 an artifact written in an earlier commit and never
re-touched by the tip was covered by the **ancestry** check alone — which proves the *commit*
reached `origin/main` but never reads the *file* there, exactly the gap rev 7 said the per-file
check existed to close. The pair of commits closes it: the aggregate set is everything the Stage
session changed, not merely what its last commit changed.

**The degraded path defaults; it does not halt** (the rev-2 lesson, deliberation §8.2). Every field
has a deterministic fallback, so a Stage invocation predating B3 — which reports nothing and leaves
the worktree on the default branch — still completes: `stage_branch` stays **UNRESOLVED**, and
Step 1.5's resolution block routes that case straight to step 3, whose 3a create-arm builds the
artifact branch from the current commit. That is exactly the pre-change behaviour, so nothing
regresses, 3a's create-arm stays **live rather than dead prescription**, and the one substitution
that caused this defect is still refused. B3 supplies the producer, so the degraded path is the
exception rather than the steady state.

**Correction 2 — resolve the staging branch BEFORE the dirty/clean split, and widen the dirtiness
pathspec.** The resolution must sit **above step 1**, not inside step 2: step 1 routes a dirty tree
straight to step 3, so anything placed in step 2 is skipped on the very path that mutates the
repository. Add this block immediately after the preamble, and revise step 1:

> **Resolve the staging branch first.** Set `STAGE_BRANCH` from the Step 1 handback record.
> When it is resolved (not UNRESOLVED):
> `git rev-parse --verify "$STAGE_BRANCH"` — if this fails, halt with
> `STAGING_GATE_FAIL: staging branch {stage_branch} does not exist`
> `git rev-parse --abbrev-ref HEAD` — MUST equal `$STAGE_BRANCH`; if it does not, halt with
> `STAGING_GATE_FAIL: worktree is on {head} but Stage reported {stage_branch}` (this is also the
> P-016 single-worktree gate: exactly one worktree, on the branch Stage reported).
> When `STAGE_BRANCH` is UNRESOLVED, record that and apply **no** branch assertion; steps 1 and 2
> still run unchanged, and any work they find routes to step 3a's create-arm, which builds the
> artifact branch from the current commit.

> 1. Check `git status --short -- {STAGE_ARTIFACT_ROOTS}` — the **fixed** constant
>    `.backlogit/ docs/plans/ docs/decisions/ docs/memory/`, never a handback-supplied path list
>    (rev 7) — for uncommitted staging artifacts:
>    - If dirty: staging artifacts need to be committed first (proceed to step 3).
>    - If clean: proceed to step 2.

Both guards now execute on **every** path, including the dirty one, so 3a never checks out an
unverified branch name and the P-016 assertion is not bypassed in the one scenario where the
Orchestrator mutates the repo. Leaving the UNRESOLVED case inside the normal step-1/step-2 flow
(rather than short-circuiting it to step 3) is what preserves **pre-change semantics exactly** for a
legacy Stage run: with `HEAD` on the default branch, step 2's range is the same comparison Step 1.5
made before this change, so an already-committed-but-unpushed default-branch commit — the `fdff9e4`
shape — is still discovered and still routed to step 3 rather than silently skipped.

**Correction 3 — discovery becomes branch-specific.** Revised step 2:

> 2. Check for unpushed commits **on the checked-out staging branch** (`$STAGE_BRANCH`, resolved
>    and verified above):
>    `git fetch origin main`
>    `git log origin/main..HEAD --oneline`
>    - If output is empty: the staging branch carries nothing beyond the remote default branch.
>      Proceed to step 4.
>    - If output is non-empty: staging commits exist that are not on the remote. Proceed to step 3.
>
>    The comparison **must** be taken against the checked-out staging branch, never against the
>    default branch as a fixed endpoint — Stage commits on `$STAGE_BRANCH`, so a default-branch
>    self-range is empty by construction and reports a false "synchronized".

`origin/main..HEAD` and `origin/main..$STAGE_BRANCH` are the **same range** whenever `$STAGE_BRANCH`
is resolved, and that equality is *proven*, not assumed: the `git rev-parse --abbrev-ref HEAD`
assertion in the resolution block above is what makes `HEAD` a legitimate stand-in for the reported
branch. Writing the range against `HEAD` rather than against the branch name is what lets the one
command serve both populations — it is branch-specific discovery for a resolved Stage branch, and
it degrades to exactly the pre-change comparison for a legacy run whose `HEAD` is still the default
branch, so neither case is skipped. Either literal satisfies AC-B1.6; the `HEAD` form is written
because it composes with the clean-tree and single-worktree gates asserted on the same branch.

**Self-match hazard — the corrected text must NOT quote the forbidden range (rev 6).** The obvious
way to write correction 3 is a prohibition naming the literal `origin/main..main`. That would
reintroduce, in the very file this gate protects, exactly the failure documented in
`docs/compound/2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md` and already
guarded against at §5.0 and AC-C2.3: the B1 harness asserts the literal is **absent** from Step 1.5,
and a prohibition quoting it would match itself and make that assertion unsatisfiable — or force it
to be weakened into something that no longer rejects the regression. The prohibition is therefore
stated **descriptively** ("the default branch against its own remote-tracking ref"), and the
absence assertion stays a sound literal check. This constraint is load-bearing and is carried by
AC-B1.6.

**Correction 4 — step 4 gains a second, shipment-free verification arm.** Revised step 4:

> 4. **Validate the outcome/shipment pair before selecting an arm** (rev 10). Resolve
>    `stage_outcome` per step 1's rule — reported value preserved, derived only when absent — then,
>    **before** either arm runs:
>    - `stage_outcome` MUST be exactly `shipment` or `no-shipment`. Any other value, including an
>      empty one that no derivation could resolve, halts with `STAGING_GATE_FAIL: unknown
>      stage_outcome {stage_outcome}`.
>    - When `stage_outcome` is `shipment`, a **non-empty** `shipment_id` MUST be present. Otherwise
>      halt with `STAGING_GATE_FAIL: stage_outcome is shipment but no shipment_id was reported`.
>    - When `stage_outcome` is `no-shipment`, **no** `shipment_id` may be present. Otherwise halt
>      with `STAGING_GATE_FAIL: stage_outcome is no-shipment but shipment_id {shipment_id} was
>      reported`.
>
>    Only a validated pair selects an arm. A reported `shipment` outcome is **never** reclassified
>    as a successful `no-shipment` run.
>
>    **When `stage_outcome` is `shipment`:**
>    Verify the shipment manifest exists on the remote default branch:
>    `git show origin/main:.backlogit/queue/{shipment_id}.md`
>    - If the file exists: staging artifacts are confirmed on the remote. Proceed to Step 2.
>    - If the file does not exist: halt with `STAGING_GATE_FAIL: shipment manifest {shipment_id}
>      not found on origin/main`.
>
>    **When `stage_outcome` is `no-shipment`:**
>    There is no manifest to name, so verify the **commit range** and **every concrete file the
>    Stage session changed across it**. Run these six checks in order; any failure halts the gate.
>
>    1. `git rev-parse --verify {stage_head_commit}^{commit}` — halt with
>       `STAGING_GATE_FAIL: stage_head_commit {stage_head_commit} is not a commit in this
>       repository`.
>    2. `stage_base_commit` MUST be resolved and MUST be a commit:
>       `git rev-parse --verify {stage_base_commit}^{commit}` — halt with
>       `STAGING_GATE_FAIL: stage_base_commit is unresolved or is not a commit in this repository`.
>    3. `git merge-base --is-ancestor {stage_base_commit} {stage_head_commit}` — halt with
>       `STAGING_GATE_FAIL: stage_base_commit {stage_base_commit} is not an ancestor of
>       stage_head_commit {stage_head_commit}`. This is what makes the pair a **range** rather
>       than two unrelated commits.
>    4. `git merge-base --is-ancestor {stage_head_commit} origin/main` — halt with
>       `STAGING_GATE_FAIL: stage commit {stage_head_commit} has not reached origin/main`.
>    5. Derive the **aggregate** set of Stage artifact files the session changed:
>       `git diff --name-only -z --diff-filter=d {stage_base_commit} {stage_head_commit} --
>       {STAGE_ARTIFACT_ROOTS}`. Call the result `AGGREGATE`. If `AGGREGATE` is **empty**, halt
>       with `STAGING_GATE_FAIL: no Stage artifact files changed between {stage_base_commit} and
>       {stage_head_commit} under the allowed Stage artifact roots`.
>    6. Resolve `VERIFY_SET`, then read every member on the default branch:
>       - When the handback reported a **non-empty** `stage_artifact_paths`, every reported path
>         MUST be a repository-relative **file** path — no trailing `/`, not a bare
>         `STAGE_ARTIFACT_ROOTS` entry, no glob or wildcard — and MUST lie under
>         `STAGE_ARTIFACT_ROOTS`; otherwise halt with `STAGING_GATE_FAIL: stage_artifact_paths
>         contains a non-file or out-of-root path: {path}`. The reported set MUST then be
>         **equal** to `AGGREGATE` (same members, order-insensitive); otherwise halt with
>         `STAGING_GATE_FAIL: stage_artifact_paths does not match the files changed between
>         {stage_base_commit} and {stage_head_commit}`. `VERIFY_SET` is that set. When no paths
>         were reported — none were supplied, or step 3a discarded them after creating the
>         artifact commit — set `VERIFY_SET := AGGREGATE`.
>       - For **every** path in `VERIFY_SET`: `git cat-file -t "origin/main:{path}"` MUST exit 0
>         **and** print exactly `blob`. Halt with `STAGING_GATE_FAIL: Stage artifact {path} not
>         found on origin/main` on a non-zero exit, and with `STAGING_GATE_FAIL: Stage artifact
>         {path} is a {type}, not a file, on origin/main` on any other printed type.
>
>    When all six pass: every file the Stage session wrote is confirmed on the remote **as a
>    concrete file**. **Stop here — there is no shipment, so there is nothing to route to Ship.**

This is the **concrete artifact verification target that requires no shipment ID**: ancestry proves
the Stage commits reached the default branch, and the per-file type check proves the artifacts
themselves are readable **files** there rather than merely that some merge occurred.

**Why the path set must be concrete files, not directory prefixes (rev 7).** Rev 6 wrote this arm
as `git show origin/main:{path}` over `stage_artifact_paths`, whose documented default was the
directory list `.backlogit/ docs/plans/ docs/decisions/ docs/memory/`. `git show` on a **tree**
exits **0** and prints a directory listing — verified against this repository:
`git show origin/main:docs/plans/` prints `tree origin/main:docs/plans/` and returns status 0.
Those four directories already exist on `origin/main` in every repository this gate will ever run
in, so the loop passed **without reading a single file the Stage run wrote**. Paired with the
degraded `stage_head_commit` default — `git rev-parse HEAD`, which on a post-merge checkout is
routinely already an ancestor of `origin/main` — check 2 is trivially true as well, and the arm
reports a **false pass**. This is rev 6's own defect one layer down: rev 6 fixed *"the loop can run
zero times"* and left *"the loop can run only over things that are always there."* Three properties
close it, and all three are load-bearing:

1. **Concreteness is structural, not stylistic.** `git diff --name-only` between two trees emits
   **only blob paths** — a directory cannot appear in `AGGREGATE` by construction, so the derived
   set is concrete whether or not anyone remembers to check it. `-z` NUL-delimits the output so
   paths are never quoted or backslash-escaped, and `--diff-filter=d` excludes **deletions**, so a
   Stage artifact the session *removed* or renamed away is not demanded to be readable on
   `origin/main` (the `HEAD` of this very branch is a rename of that shape). The trailing
   `-- {STAGE_ARTIFACT_ROOTS}` pathspec applies the containment allow-list at the source rather
   than by post-filtering, so an out-of-root path cannot enter the set at all.
2. **`git cat-file -t` replaces `git show` in this arm.** Both exit 0 on a tree, so the exit code
   alone can never discriminate a file from a directory; the discriminator is the **printed object
   type**. Requiring exactly `blob` rejects a directory even if one reached `VERIFY_SET` through
   the reported-path branch, and `cat-file -t` still exits non-zero (128) when the path is absent,
   so presence and file-ness are proven by one command.
3. **Set equality, not subset.** Equality rejects a reported path the session never touched (an
   invented or copy-pasted path) **and** a reported set that silently drops files — which would
   otherwise let a handback shrink the verified set down to one always-present file. A subset check
   catches only the first.

**Why the set is the aggregate base→head diff, and why the range form is still refused (rev 8).**
A **two-tree** diff over `{stage_base_commit} {stage_head_commit}` is a property of the two commit
objects' trees: it reads **identically before and after the merge**, so it is recoverable when
step 4 runs. A **range** form such as `origin/main..{stage_head_commit}` is **empty by
construction** once that commit is an ancestor of `origin/main` — precisely the state step 4 runs
in — so a range-derived set would collapse to empty and re-create the vacuous-loop failure rev 7
exists to remove. The two-tree form has the post-merge stability of rev 7's single-commit
`diff-tree` **and** the session-wide coverage rev 7 lacked; that is the whole reason rev 8 changes
the derivation rather than merely widening it.

**Rev 7's "scoping to the tip loses nothing" argument is WITHDRAWN.** Rev 7 held that the per-file
check could safely read only `stage_head_commit`'s own changes, because check 2 already proved
every **ancestor** reached `origin/main`. That is a **category error**: commit ancestry proves the
*commit* landed, not that the *file* is readable on the default branch — which is exactly the gap
rev 7 introduced the per-file check to close. Applying the per-file check to the tip alone leaves
every artifact written in an earlier commit of the same session covered **only** by the property
rev 7 itself judged insufficient. A Stage session that writes its deliberation in commit 1 and its
plan in commit 2 therefore had the deliberation verified by nothing stronger than ancestry. The
aggregate set removes the asymmetry: every file the session changed is read on `origin/main`,
whichever commit changed it.

**Known widening, recorded not hidden (rev 8).** If a Stage session merges `origin/main` into its
artifact branch mid-flight, the two-tree diff also reports files that arrived from the default
branch rather than from Stage. The `-- {STAGE_ARTIFACT_ROOTS}` pathspec bounds that to the Stage
artifact roots, and every such file is by definition already on `origin/main`, so the per-file
check passes for it. The effect is therefore a few **extra** verified paths, never a missed one —
a strictly conservative failure direction. It is recorded as **R14** rather than engineered away,
because a first-parent or merge-aware derivation would trade this harmless over-coverage for the
under-coverage rev 8 exists to remove.

**The handback-resolution no-halt invariant is preserved (AC-B1.5).** Every new halt lives in
**step 4's verification**, not in the handback defaulting. All fields still resolve to a
deterministic default without halting; what changed is that a degraded resolution now yields a set
that is either concretely verifiable or **provably empty**, and an empty one fails loudly instead
of passing silently. A gate that halts because it could not prove the artifacts are on
`origin/main` is the gate working, not the degraded path wedging.

**No new command shape enters the corpus.** `diff`, `cat-file`, `rev-parse`, and `merge-base`
are read forms: none is a push verb, so §5.3 construct 1 cannot match them, and none grants
permission to commit anywhere, so construct 2 cannot either. The detector's closed construct set
and the fixture corpus are therefore **not** widened — the rev-6 non-expansion position is
unchanged. The shipment arm stays byte-verbatim (AC-B1.4), the resolution block's clean-tree and
single-worktree guards are untouched (AC-B1.9), and P-009, P-011, and P-016 remain textually
unchanged (AC-C2.6).

**Correction 5 (rev 7, extended rev 8) — step 3 re-records the head and PRESERVES the base.** When
step 1 finds a dirty tree, **3a creates the commit that carries the artifacts** — and the
handback's `stage_head_commit` was resolved *before* that commit existed. Deriving step 4's file
set from the stale value would read a commit that changed none of the artifacts, halting the gate
for the wrong reason on the one path Step 1.5 exists to repair. 3a therefore:

- re-records `stage_head_commit` as `git rev-parse HEAD` **after** committing;
- **preserves `stage_base_commit` unchanged** (rev 8). The base predates 3a's commit and is
  therefore still the correct lower bound; re-recording it to the post-commit `HEAD` would collapse
  the range to nothing, and re-recording it to the *pre-commit* `HEAD` would silently discard any
  already-committed unpushed Stage commits step 2 routed into step 3 — reproducing the tip-only
  blind spot rev 8 removes, on the one path most likely to carry several commits. When
  `stage_base_commit` is still **UNRESOLVED** at 3a — a direct invocation the Orchestrator never
  bracketed — 3a sets it to the `HEAD` captured **before** its own commit, which is the tightest
  sound lower bound available at that point;
- discards any reported `stage_artifact_paths` (they described the pre-commit intent, not the
  committed range) so check 6 falls through to `VERIFY_SET := AGGREGATE`, **recomputed from the
  preserved base to the new head** rather than from the tip commit alone.

When step 3 was entered for unpushed commits alone and nothing was uncommitted, `HEAD` is unchanged
and the re-record is a no-op. AC-B1.3's branch-point text (`from the current commit`) is preserved
verbatim.

**Correction 6 (rev 10) — the outcome field must not fail open.** Revisions 6–9 wrote the
`stage_outcome` default as an *unconditional derivation*: `shipment` when a `shipment_id` was
returned, `no-shipment` otherwise. That reads as a harmless fallback and is not one. It has no
`absent` branch, so it overwrites a **reported** value with a derived one, and the derivation's
`otherwise` clause is a **success** terminal. The failure is therefore silent and inverted:

| Reported by Stage | `shipment_id` reached the Orchestrator | Rev 6–9 resolution | Consequence |
|---|---|---|---|
| `shipment` | yes | `shipment` | correct |
| `shipment` | **no** (dropped, truncated, mis-parsed handback) | **`no-shipment`** | the shipment arm is **skipped**, the no-shipment arm runs and **passes** (the artifacts really are on `origin/main`), the gate reports success, and **Ship is never routed** — a formed shipment silently abandoned |
| `no-shipment` | n/a | `no-shipment` | correct |
| *(absent)* | either | derived | the intended use of the rule |

Row 2 is the whole finding. The no-shipment arm is *designed* to pass for a run whose artifacts
reached the default branch, so it cannot be relied on to notice that a shipment was expected; it
verifies files, not intent. The derivation converts the loudest failure the pipeline has — a
shipment that was formed but not handed off — into a clean terminal, and does it on the one input
class most likely to be degraded, since a handback that lost its `shipment_id` is by definition a
handback that was not fully received.

Two changes close it, and they are deliberately at **different** points in the flow:

1. **Preservation, in the defaulting phase.** The derivation is scoped to the **absent** case. A
   reported value wins. This is a strictly weaker rule than the one it replaces, so the
   never-halts invariant is untouched — resolution still cannot halt, and a handback that reports
   nothing still resolves exactly as before.
2. **Validation, in the verification phase.** The pair check above sits in **step 4**, alongside
   the other `STAGING_GATE_FAIL` halts, and runs **before** arm selection. Placing it here rather
   than in the defaulting block is what preserves AC-B1.5's "all halts live in verification, never
   in field defaulting" invariant — the same placement discipline rev 6 arrived at after its first
   draft hard-halted the degraded population it claimed not to wedge (§12, rev-6 internal defect 1).

**Why validation and not a third arm.** An inconsistent pair is not a third outcome needing its own
verification; it is evidence that the handback is **untrustworthy**, and every downstream check
reads that same handback. Continuing on either arm would verify a claim whose provenance has
already been contradicted. Halting is the fail-closed response, and it is loud: the operator sees
which half disagreed, and a re-invocation with a complete handback proceeds normally.

**No new command shape, no widened detector.** The pair validation is a comparison between two
already-declared fields. It adds no git command, no push verb, and no permission to commit
anywhere, so §5.3's closed construct set and the 14-fixture corpus are unchanged — the rev-6/7/8/9
non-expansion position, unchanged again.

**The rev-2 "no no-shipment clause here" position is WITHDRAWN.** Rev 2 argued that Step 1.5 is
scoped "After Stage completes and before routing to Ship" and that step 4 requires a
`{shipment_id}`, so a shipment-less run never enters the gate and the obligation belongs in B2/B3
instead. That reasoning was internally inconsistent with §6.2's own rev-4 route table, which names
the **Orchestrator (pipeline invocation)** as the actor for pushing, opening, merging, and
post-merge-verifying the no-shipment branch — and Step 1.5 is the **only** Orchestrator gate that
does any of those things. Rev 2 thus assigned the Orchestrator a duty and simultaneously declared
the one place it could discharge it out of scope. Corrections 1–5 give that duty a home; B2/B3 keep
the Stage-side obligations they already carry. Nothing is duplicated: B2/B3 say *where Stage may
write*, §6.1.2 says *how the Orchestrator finds and verifies it*.

**No direct default-branch push is introduced.** Corrections 1–5 add only read forms
(`status`, `rev-parse`, `fetch`, `log`, `show`, `merge-base`, `cat-file`, and — rev 8 — `diff`)
and one `checkout`; 3b's existing push-the-branch-and-open-a-PR form is untouched, and
3e stays deleted. P-009, P-011, and P-016 are textually unchanged (AC-C2.6).

**AC-B1.1** No direct-push-to-default-branch instruction remains anywhere in Step 1.5.
**AC-B1.2** The preamble states the postcondition as artifacts having **reached** the default
branch **via a merged staging PR** (rev 3's fourth surface — §2), and its scope clause binds
**whether or not a shipment was formed** (rev 6).
**AC-B1.3** 3a textually carries the branch point (`from the current commit`) for the
already-committed-but-unpushed path — textual, not structural, because the a–e sub-steps are lazy
paragraph continuations, not a real nested list (see §11 on the pre-existing list-rendering defect).
**AC-B1.4** Step 4's **shipment arm** preserves, verbatim, the pre-change lead sentence
(`Verify the shipment manifest exists on the remote default branch:`, capitalization included), the
`git show origin/main:.backlogit/queue/{shipment_id}.md` command, both outcome bullets, and the
`STAGING_GATE_FAIL: shipment manifest {shipment_id} not found on origin/main` halt string, with
pass/fail semantics unchanged. **Rev 6 scope refinement**: the criterion previously required the
whole step-4 block to be byte-identical. That form is no longer satisfiable, because step 4 must
gain the `no-shipment` arm (AC-B1.8) — and an unmodified block is precisely what leaves that path
unverified. The criterion is narrowed to what it was always protecting: the authoritative
post-merge gate is preserved and unrelaxed. The step may gain **outcome-arm label lines** and the
second arm, **and nothing else** — no outer lead sentence replaces or rewords the preserved one.
**AC-B1.5** **Staging handback recorded, with total defaults.** Step 1 declares and step 4 records
`stage_branch`, `stage_base_commit` (rev 8), `stage_head_commit`, `stage_outcome`, and
`stage_artifact_paths`. **Every** field has a deterministic default, so the degraded path **never
halts**: branch/head resolve from `HEAD`; `stage_base_commit` resolves to the value the
Orchestrator **captured and retained immediately before invoking Stage** — authoritative over any
reported value — falling back to the reported value only when nothing was retained, and otherwise
left UNRESOLVED without halting; `stage_outcome` is **preserved when reported** and derived from
whether a `shipment_id` was returned **only when it is absent** (rev 10 — the derivation never
overwrites a reported value);
`stage_artifact_paths` is left **empty** and derived in step 4 from the
`stage_base_commit`→`stage_head_commit` aggregate (rev 7/8 — it is **never** defaulted to a
directory list); `STAGING_HANDBACK_DEGRADED` is recorded once and surfaced. When the `HEAD`
resolution yields the default branch, `stage_branch` is left **UNRESOLVED** and the run routes to
step 3a's create-arm. The default branch is **never** substituted for the Stage artifact branch.
**AC-B1.6** **Branch-specific discovery.** Step 1.5's unpushed-commit check compares
`origin/main..HEAD` — equivalently `origin/main..$STAGE_BRANCH` whenever a branch is resolved,
proven equal by the `git rev-parse --abbrev-ref HEAD` assertion in the resolution block, and
degrading to exactly the pre-change comparison when it is UNRESOLVED, so neither population is
skipped. The literal default-branch self-range is **absent from the whole of Step 1.5**, including
from the prohibition sentence, which is phrased descriptively so the absence assertion cannot
self-match.
**AC-B1.7** **Dirtiness pathspec covers Stage artifacts.** Step 1's `git status` pathspec is the
**fixed** `STAGE_ARTIFACT_ROOTS` constant — `.backlogit/ docs/plans/ docs/decisions/
docs/memory/`, not `.backlogit/` alone — so a no-shipment run that writes only `docs/**` is not
reported clean. **Rev 7**: the pathspec is that constant and **not** the handback's
`stage_artifact_paths`, so a handback reporting one narrow path cannot shrink the dirtiness scan
and hide uncommitted artifacts. Directory prefixes are correct *here* — step 1 runs before the
artifact commit exists, so there is no commit to derive files from — and are rejected in step 4,
where a commit does exist.
**AC-B1.8** **No-shipment verification arm.** Step 4 carries a second arm keyed on
`stage_outcome: no-shipment` that runs **six** ordered checks and halts with `STAGING_GATE_FAIL`
on any failure: (1) `git rev-parse --verify {stage_head_commit}^{commit}`; (2) `git rev-parse
--verify {stage_base_commit}^{commit}`, rejecting an UNRESOLVED base; (3) `git merge-base
--is-ancestor {stage_base_commit} {stage_head_commit}`; (4) `git merge-base --is-ancestor
{stage_head_commit} origin/main`; (5) derivation of `AGGREGATE` from `git diff --name-only -z
--diff-filter=d {stage_base_commit} {stage_head_commit} -- {STAGE_ARTIFACT_ROOTS}`, halting when
it is **empty**; (6) resolution of `VERIFY_SET` per AC-B1.10 followed by `git cat-file -t
"origin/main:{path}"` for **every** path in it, requiring exit 0 **and** output exactly `blob`. It
terminates the gate without routing to Ship and names no `shipment_id`. **Rev 7**: `git show
origin/main:{path}` no longer appears in this arm — it exits 0 on a tree, so it could not reject
directory prefixes. **Rev 8**: the single-commit `git diff-tree` derivation no longer appears
either — it covered only the tip commit — and the `origin/main..{commit}` range form remains
refused, being empty by construction at the moment this arm runs.
**AC-B1.9** **Clean-tree and single-worktree gates preserved on every path.** The branch-resolution
block sits **above** step 1's dirty/clean split, so `git rev-parse --verify` and the
HEAD-equals-`$STAGE_BRANCH` assertion execute on the dirty path too — 3a never checks out an
unverified branch name. Step 1.5 halts when HEAD is not the resolved staging branch; no new branch
or worktree is created when the handback reports one; and P-009, P-011, and P-016 remain textually
unchanged.
**AC-B1.10** **Concrete-path contract (rev 7, aggregate at rev 8).** `stage_artifact_paths` is
defined as a **non-empty set of concrete repository-relative file paths changed between
`stage_base_commit` and `stage_head_commit`**. Step 4 rejects, by halting: a path with a trailing
`/`, a bare `STAGE_ARTIFACT_ROOTS` entry, or any glob/wildcard form (not a file path); a path not
under `STAGE_ARTIFACT_ROOTS` (outside the allowed Stage artifact roots); an **empty** resolved set;
any reported set that is not **equal** to `AGGREGATE`, the aggregate base→head changed-file set
(equality, not subset — a subset check would accept both an invented path and a silently shrunken
set); and any path that resolves on `origin/main` to an object type other than `blob`. No directory
prefix can satisfy completion on this arm, by construction: a two-tree `git diff --name-only` emits
blob paths only, the `-- {STAGE_ARTIFACT_ROOTS}` pathspec bounds containment at the source, and
`cat-file -t` must print exactly `blob`.
**AC-B1.11** **Handback commits re-recorded and preserved across step 3 (rev 7, extended rev 8).**
3a sets `stage_head_commit` to `git rev-parse HEAD` after committing and discards any reported
`stage_artifact_paths`, so step 4 derives its file set from the commits that actually carry the
artifacts rather than from the pre-commit `HEAD` the handback resolved. **`stage_base_commit` is
preserved unchanged** across 3a — or set to the `HEAD` captured *before* 3a's commit when it was
UNRESOLVED — so the recomputed set is the **aggregate over the preserved base**, never the tip
commit alone. The re-record is a no-op when step 3 was entered for unpushed commits alone.
AC-B1.3's `from the current commit` branch-point text is preserved verbatim.
**AC-B1.12** **Fail-closed outcome/shipment pair (rev 10).** `stage_outcome` is **preserved** when
the handback reports it and derived from `shipment_id` presence **only when it is absent**. Before
step 4 selects an arm, the resolved pair is validated and any failure halts with
`STAGING_GATE_FAIL`: an outcome that is neither `shipment` nor `no-shipment` (including an
unresolvable empty value) halts as an unknown outcome; `shipment` with an empty or missing
`shipment_id` halts; `no-shipment` with a `shipment_id` present halts. A reported `shipment`
outcome is **never** reclassified as a successful `no-shipment` run. The validation lives in
step 4's **verification**, not in the handback defaulting, so AC-B1.5's never-halts resolution
invariant is preserved intact.

**Rev 4 reconciliation**: this list previously carried only three criteria, worded before rev 3
added the preamble as the fourth authorization surface, while task `018.007-T` already carried the
four-criterion form and §6.3's preserved-form note already referenced **AC-B1.4**. The plan is
corrected to the task's form — the same plan↔task same-contract defect class as the C2 divergence
reconciled in §7.2, found by the rev-4 AC-ID parity sweep rather than by review.

**Rev 6 reconciliation**: AC-B1.5–AC-B1.9 are new and are mirrored into `018.007-T` in the same
change, so the plan↔task same-contract invariant holds. AC-B1.4's narrowing is recorded above
rather than silently applied, because a criterion that is weakened without a stated reason is
indistinguishable from one that was abandoned.

**Rev 7 reconciliation**: AC-B1.10 and AC-B1.11 are new, and AC-B1.5, AC-B1.7 and AC-B1.8 are
**strengthened** (never relaxed) — they now reject the directory-prefix path set rev 6 permitted.
All five are mirrored into `018.007-T` in the same change, preserving the plan↔task same-contract
invariant.

**Rev 8 reconciliation**: no criterion is added or removed — the B1 contract stays at
AC-B1.1–AC-B1.11. AC-B1.5, AC-B1.8, AC-B1.10 and AC-B1.11 are **strengthened** to the aggregate
base→head form; AC-B1.1–AC-B1.4, AC-B1.6, AC-B1.7 and AC-B1.9 are untouched. All four amended
criteria are mirrored into `018.007-T`, and that task's **description is
rewritten** so it states this contract once, in its current form, rather than layering a rev-7
addendum over operative-looking rev-6 prose (adversarial-review blocker 3).

**Rev 10 reconciliation**: one criterion is added — **AC-B1.12**, the fail-closed outcome/shipment
pair — and **AC-B1.5** is strengthened to preserve a reported `stage_outcome`. AC-B1.1–AC-B1.4 and
AC-B1.6–AC-B1.11 are untouched. Both are mirrored into `018.007-T` in the same change, so the
plan↔task same-contract invariant holds. The B1 contract is now AC-B1.1–AC-B1.12.

### 6.2 Task B2 — P-010 + Amendment Log

**Stage MAY** — replace the default-branch bullet (actor named, per deliberation §8.2):

> - Commit backlog and planning artifacts to a **dedicated Stage/admin branch** — never to the
>   default branch — **including when no shipment is formed**. Push, PR creation, and merge for
>   that branch are performed by the **Orchestrator under Step 1.5**, or by the **operator** in a
>   direct Stage invocation; **never by Stage**.

**Stage MUST NOT** — add, mirroring the existing Ship bullet:

> - Commit or push Stage artifacts directly to `main`

Stage's existing "Create, push, or merge pull requests" prohibition is **left intact**. The actor
clause is what makes the new rule satisfiable without weakening it.

Add a clarifying sentence that Step 1.5 staging orchestration is an **Orchestrator-owned gate**,
not delegated Ship work — so it does not collide with P-010's "Orchestrator must not perform Stage
or Ship work directly" — **and** that a Stage artifact branch/PR, carrying no source, test, or
config change, is **not an implementation branch** for P-016 or P-001 purposes (rev 3; see §10.2).

**Stage MAY** also gains an explicit branch-creation grant (rev 3), since no existing text names
who creates the branch in the no-shipment case:

> - Create and check out the dedicated Stage/admin artifact branch `chore/stage-{scope-slug}`,
>   where `{scope-slug}` is derived from the scope under staging per P-011's existing slug
>   convention. A shipment ID **may** be used as the slug when one is already known; it is **not
>   required**, and Stage never waits for one.

**Naming must not depend on a shipment ID (rev 9).** Revisions 1–8 wrote this grant as
`chore/stage-{shipment_id}`, with `chore/stage-{chore-slug}` as a shipment-less fallback. That form
is **underivable at the moment the branch is needed**: the branch must exist before Stage's first
artifact write (§6.3.1), and the shipment is not created until Step 5.5 — the *last* mutation of the
session. A shipment-forming run following the rev-8 wording literally therefore had no name to use
at gate time, and the wording pushed an executor toward deferring the branch until after the very
mutations it exists to protect. `{scope-slug}` is derivable from the selected grouping's covering
feature title (task-shaped intake) or the feature/epic/chore title (feature-shaped intake) as soon
as Step 1/1.5 completes, so the name is stable for the whole session and identical on both
invocation paths. The shipment-ID form is preserved as a **permitted** slug value rather than
deleted, so an Orchestrator that already holds a shipment ID may still name the branch after it
(`chore/stage-017-S`). The branch carrying this plan, `chore/stage-pipeline-policy-gap`, is a
conforming example of the `{scope-slug}` form.

**Concrete no-shipment persistence route** (rev 4 — review round 3). Stating the actors end to end,
because "never to the default branch" is only actionable if the alternative route is executable:

| Step | Actor | Action |
|---|---|---|
| 1 | **Stage** | Create and check out `chore/stage-{scope-slug}` **before its first artifact write** (grant above; mirrored into the `_stage.agent.md` Git row by B3, which is the cell role enforcement actually reads, and made operative by B3's Step 1.9 branch gate — §6.3.1) |
| 2 | **Stage** | Commit the backlog/planning artifacts to that branch (operative as B3's Step 5.7, after every Stage mutation and before the Step 6 handback — §6.3.1) |
| 3 | **Orchestrator** (pipeline invocation) or **operator** (direct Stage invocation) | Push the branch, open the PR, merge it |
| 4 | **Orchestrator** or **operator** | Post-merge verification |

Stage's role ends at step 2. Steps 3–4 are **never** Stage actions — this is what keeps the new
obligation satisfiable without weakening Stage's standing PR prohibition.

**How step 3's actor learns the branch (rev 6).** This table names the actors but not the handoff,
and a route whose step-3 actor cannot *find* the branch produced by step 1 is not executable. The
handoff is specified in B1 (the Orchestrator's Step 1 **staging handback record**) and produced in
B3 (Stage's Step 6 summary contract emits it). P-010's text is unchanged by rev 6 — the handback is
a mechanism, not a permission — but the route above is only complete when read together with
§6.1.2 and §6.3. No P-010 acceptance criterion changes.

**Verification evidence, pre- vs post-merge** — the two `git show` forms are **not**
interchangeable and neither replaces the other:

- **Pre-merge**, on the staging branch: `git show {branch}:{path}` is **local evidence only**. It
  proves the artifact was committed; it proves nothing about the default branch.
- **Post-merge**, the authoritative gate: `git show origin/main:{path}` **remains unchanged and
  remains required**. Step 1.5 step 4 keeps this form byte-identical (task B1, AC-B1.4).

Correcting *where Stage may write* does not relax *where the Orchestrator must verify*. The
post-merge `origin/main` gate is the reason the branch route is safe, not a casualty of it.

**Amendment Log**: append exactly one row, version **1.25.0** (the log stands at 1.24.0 — rev 1
said only "bump minor"), Change cell "Corrected P-010", body ending "Corrects, and does not delete
or edit, the 1.5.0 row above".

**Acceptance criteria (rev 8 — reconciled to a single six-criterion contract).** Revisions 1–7
carried **two** divergent five-criterion lists: this section asserted Stage/Ship symmetry but
omitted the P-016/P-001 clarification, `018.008-T` asserted the clarification but omitted symmetry,
and the shared IDs AC-B2.3/AC-B2.4/AC-B2.5 were bound to **different** criteria on each surface —
so an executor satisfying one list would have failed the other while both claimed to be the same
contract. The union is six criteria; both surfaces now carry them verbatim, in this order and with
these IDs.

**AC-B2.1** No P-010 text authorizes Stage to commit or push on the default branch — no
default-branch allowance remains anywhere in P-010's Stage block.
**AC-B2.2** The rule holds explicitly **even when no shipment is formed**, and the branch name is
derivable **without** one: the grant names `chore/stage-{scope-slug}` and treats a shipment ID as a
**permitted but not required** slug value (rev 9).
**AC-B2.3** P-010's **Stage and Ship columns are symmetric** on direct default-branch writes: the
Stage block carries a prohibition mirroring the existing **Ship MUST NOT** bullet
"Commit or push directly to `main`".
**AC-B2.4** **No actor is both prescribed and prohibited** on the same operation (no
self-deadlock): every obligation the new bullet creates names an authorized actor, with pushing the
branch and opening/merging the PR attributed to the **Orchestrator** (pipeline invocation) or the
**operator** (direct Stage invocation) — never to Stage, whose standing PR prohibition is
unweakened.
**AC-B2.5** The **P-016/P-001 non-implementation clarification** is present in the **policy text**,
not only in this plan: a Stage artifact branch/PR carries no source, test, or config change and is
therefore neither an implementation branch under P-016 nor a release unit under P-001.
**AC-B2.6** The Amendment Log gains **exactly one additive row, version 1.25.0**, following the
existing format; 1.24.0 and every earlier row are unmodified.

**Rev 9 reconciliation**: no criterion is added or removed — the B2 contract stays at
AC-B2.1–AC-B2.6. **AC-B2.2 is corrected in place** to the shipment-ID-independent naming form, and
the grant bullet and route table above are corrected with it. Both are mirrored into `018.008-T` in
the same change, preserving the plan↔task same-contract invariant. The correction **strengthens**
the criterion: the rev-8 form was satisfiable by a policy whose branch name could not be computed
at the moment the branch is required.

### 6.3 Task B3 — `_stage.agent.md` Role Boundary

Rewrite the **Git** row (deliberation §8.1 — the operative surface):

| Column | New text |
|---|---|
| Allowed | Commit backlog/planning artifacts to a **dedicated Stage/admin branch**; **create and check out that dedicated artifact branch** (`chore/stage-{scope-slug}`; a shipment ID may be used as the slug when one is already known, but is never required); create/use an explicit, time-boxed spike/research worktree only for staging investigation |
| Forbidden | **Commit or push directly to the default branch**; create or checkout feature/chore branches for code execution; create/use parallel implementation branches or worktrees |

The **branch creation/check-out grant is load-bearing, not decorative** (rev 4 — review round 3).
`role-enforcement.instructions.md` makes this table — not P-010 — the permission set consulted at
mutation time (§2). B2 grants Stage `Create and check out the dedicated Stage/admin artifact
branch`; if that grant appears only in P-010 and not in this cell, role enforcement **fails
closed** on the very first action the corrected policy requires, and the no-shipment persistence
route becomes unexecutable. The two surfaces must carry the grant in both places.

Add a parenthetical to the **PR** row recording that branch **push**, PR creation, and merge are
performed by the **Orchestrator** (pipeline invocation, Step 1.5) or the **operator** (direct Stage
invocation) on Stage's behalf — keeping the prohibition intact while removing the apparent
deadlock. The PR row is the **only** row other than Git that this task may touch, and it may gain
**only** that actor note.

**Staging handback emission (rev 6).** Add to Stage's **Step 6 (Summary)** session-output contract
one required line emitting the **staging handback record** that the Orchestrator's Step 1 now
requires (§6.1.2, correction 1): `stage_branch`, `stage_base_commit` (rev 8), `stage_head_commit`,
`stage_outcome` (`shipment` | `no-shipment`), `stage_artifact_paths`, and `shipment_id` when one
was formed.

**Outcome and shipment ID are emitted as a consistent pair (rev 10).** The same line states that
`stage_outcome` is **always emitted explicitly** — never omitted for the consumer to infer — and
that the pair is internally consistent: `shipment` is emitted together with a non-empty
`shipment_id`, and `no-shipment` is emitted with none. This is the producing half of AC-B1.12. The
consumer's validation is a fail-closed **agreement** check, and an agreement check needs two
parties: if the producer may omit the outcome and let the consumer derive it, the derivation is
back — and with it the silent reclassification of a reported `shipment` run into a successful
`no-shipment` one (§6.1.2, correction 6). Producer and consumer move together here for the same
reason they did at rev 7 and rev 8.

**Path shape is part of that contract (rev 7, aggregate at rev 8).** The same line states that
`stage_artifact_paths` is a **non-empty list of concrete repository-relative file paths** — the
files Stage committed between `stage_base_commit` and `stage_head_commit`, never directory prefixes
— and that Stage produces it with the same derivation the Orchestrator re-runs:
`git diff --name-only -z --diff-filter=d {stage_base_commit} {stage_head_commit} --
.backlogit/ docs/plans/ docs/decisions/ docs/memory/`. Naming the producing command here
is what makes the consumer's **set-equality** check (AC-B1.10) satisfiable by construction rather
than by coincidence: two surfaces computing the same set from the same commit pair with the same
command cannot disagree. A producer left free to emit directory prefixes, or free to report only
its tip commit's files, would either fail that check on every run or force it to be weakened back
into a form that verifies less than the session wrote. Stage echoes `stage_base_commit` rather than
inventing one: the Orchestrator's retained pre-invocation value is authoritative (AC-B1.5), and the
echo exists so a **direct** Stage invocation — which no Orchestrator bracketed — still hands the
operator a usable base.

This was the **first** edit B3 makes outside the Role Boundary table — rev 8 adds the Step 5.5 /
Step 6 edits below and rev 9 the two operative steps of §6.3.1; the enumerated set is fixed by
AC-B3.5, AC-B3.6, and AC-B3.7 and by nothing else. It is not cosmetic. Without
a producer, the Orchestrator's handback requirement has no source and every pipeline run falls into
B1's `STAGING_HANDBACK_DEGRADED` path. It also completes the symmetry the paragraph above rests on:
B3 must carry the branch-creation grant because role enforcement reads *this file* at mutation
time — and by the same argument the Orchestrator consumes *this file's* declared session output, so
the branch Stage created must be declared here too. Placing the emission only in the Orchestrator's
expectations would repeat, in the opposite direction, the grant-in-one-surface-only defect AC-B3.2
exists to prevent. It transfers **no** authority: emitting a branch name is not pushing, opening, or
merging anything, so the PR-row prohibition is untouched.

**Making `no-shipment` reachable (rev 8) — the second Step 5.5/Step 6 edit.** Everything above, and
the whole no-shipment arm of §6.1.2, is **dead prescription** against the installed
`_stage.agent.md` as it stands. Two clauses in that file make `stage_outcome: no-shipment`
unreachable:

| Clause | Installed behaviour | Consequence |
|---|---|---|
| **Step 5.5** | Shipment assembly is "MANDATORY — not optional" whenever the registry advertises `features.shipments: true`; ending the session without a `shipment_id` is declared a **P-005 violation** | A run that legitimately produces nothing to ship cannot terminate compliantly |
| **Step 6** pre-summary gate | "If no `shipment_id` exists, **HALT** and go back to Step 5.5" | The summary — which is where AC-B3.5's handback line lives — is never reached, so no handback is ever emitted for a shipment-less run |

So B1 waits on an outcome no producer can emit. B3 therefore also makes the terminal outcome
executable, with the **narrowest** change that achieves it:

1. **Step 5.5** is scoped to a harvest that produced items. Its existing guardrail — *do not
   assemble a shipment if the harvest step produced no items or produced items with unresolved
   P-003 violations; halt and report before creating an empty shipment* — is preserved **verbatim**
   and is precisely what identifies the two cases: a **valid empty harvest** (nothing to decompose)
   terminates as `no-shipment`, while **unresolved P-003 violations** still **halt**.
2. **Step 6**'s pre-summary verification gate requires a `shipment_id` **only when `stage_outcome`
   is `shipment`**. When it is `no-shipment`, the gate instead requires the **complete** handback
   record and an explicit terminal statement that there is nothing to route to Ship.
3. The summary for a `no-shipment` run **must not** direct the operator to Ship. This preserves,
   rather than contradicts, the existing prohibition on handing Ship a feature ID instead of a
   shipment ID: with no shipment, the correct handoff is *none*, and the Orchestrator's Step 1.5
   no-shipment arm is the terminal gate.

**Failures stay failures — this is the load-bearing constraint.** `no-shipment` is a **success**
terminal for a run that had nothing to ship. It is **never** a re-label for failure. A P-003
lineage violation, a harvest that errored, or a harvest that produced items but could not form the
required shipment all continue to **halt** exactly as they do today, and MUST NOT be recorded as
`no-shipment`. Without that sentence the edit would convert the pipeline's loudest failure mode
into a silent success — a strictly worse defect than the unreachable arm it fixes.

#### 6.3.1 Operative branch gate and artifact commit (rev 9)

**The defect.** Revisions 1–8 corrected *what Stage may do* and *what Stage must report*. Neither
is an instruction to **do** it. The installed `_stage.agent.md` contains no branch operation and no
commit operation at any step: its Required Steps run triage → grouping → learnings → deliberation →
planning → review → harvest → shipment assembly → stash archival → summary, and its Step Sequence
Contract checklist — the file's own ordered, machine-readable statement of what a session must
execute — names none either.

| Rev 1–8 surface | What it establishes | What it does not |
|---|---|---|
| Role Boundary Git row (AC-B3.1/B3.2) | Stage **may** create, check out, and commit on the artifact branch | that any step ever does |
| P-010 grant (AC-B2.2) | the same permission, in policy | the same gap |
| Step 6 handback (AC-B3.5) | Stage **must report** `stage_branch`, `stage_base_commit`, `stage_head_commit`, `stage_artifact_paths` | where those values come from |

An executor satisfying AC-B3.1–AC-B3.6 to the letter can therefore complete a session **on the
default branch**, having written every artifact there — the exact `fdff9e4` shape — and then emit a
handback naming a branch that was never created and commits that were never made. The Orchestrator's
step-4 arm would reject that handback (the fields resolve to the default branch, so `stage_branch`
is left UNRESOLVED and the aggregate diff is empty), which is the gate working; but the artifacts
are already on the default branch by then, and rejection after the fact is not prevention. This is
the **producer/consumer failure class a fourth time**, in its purest form: rev 1 obliged Stage to
merge a PR it could not create; rev 6 obliged the Orchestrator to find a branch nothing announced;
rev 8 keyed a verification on an outcome nothing could emit; rev 9 finds a **reported value with no
producing action**.

**Correction 1 — Step 1.9, the pre-mutation branch gate.** Add a new step section between the
existing Step 1.8 (Learnings Retrieval) and Step 2 (Deliberation):

> ### Step 1.9: Stage Artifact Branch Gate (NON-NEGOTIABLE)
>
> Stage MUST NOT write a backlog, plan, decision, or memory artifact while `HEAD` is the default
> branch. This gate occupies a **fixed position** in the Step Sequence Contract — after Step 1.8
> and before Step 2 — which is the first point at which a stable scope slug is derivable for both
> intake shapes, Step 1 classification having settled the feature-shaped case and Step 1.5
> grouping selection the task-shaped one.
>
> **Deferral rule — NOTHING TRACKED WRITES BEFORE THIS GATE (rev 10).** This gate precedes
> **every** tracked Stage artifact mutation of the session, without exception. Any such mutation
> that would otherwise occur at an earlier step is **deferred until after** it. The rule is
> categorical; the enumeration below names the classes the installed contract actually schedules
> earlier, and is illustrative of the rule rather than a closed list that a new earlier mutation
> could escape:
>
> 1. **The write half of the mandatory late-identifier reconciliation** (Step 1). The
>    *analysis* — searching Ship's residual-risk records for a late `review-thread ID` or PR
>    number — is **read-only** and completes in Step 1. The **in-place update** of the deferred
>    stash entry that records the recovered identifiers is a tracked write and moves here.
> 2. **The archival of a duplicate stash entry** produced by the unconditional duplicate-detection
>    scan (Step 1). Detection is read-only and completes in Step 1; the `stash archive` call moves.
> 3. **Every tracked `docs/memory/` checkpoint whose triggering milestone precedes this gate.**
>    That is specifically the **stash-classification** checkpoint (Step 1) and the
>    **contextual-grouping / operator-selection** checkpoint (Step 1.5), and any other checkpoint
>    a session would write before reaching this step. All are written **after** the gate succeeds,
>    against the same session state they would have recorded earlier.
>
> **Read-only work still runs first, and must.** Step 1 classification and Step 1.5 grouping
> analysis and operator selection are *reads plus a decision*; they are what make the scope slug
> derivable, so they necessarily precede the gate. Deferring them would make the gate's own
> position underivable. What is deferred is the **write**, never the analysis or the decision.
>
> **Disposable, gitignored operations are NOT tracked artifacts and are NOT deferred.** Index
> synchronization (`backlogit sync` / `backlogit_sync_index` → `.backlogit/backlogit.db*`),
> structured backlogit checkpoints (`.backlogit/checkpoints/`), hook-event polling and
> acknowledgement (`.backlogit/hooks_queue.jsonl`), and `.backlogit/runtime/` state are all
> **gitignored in this repository** — they never appear in `git status`, cannot be committed, and
> therefore cannot reach the default branch. They are outside this rule, and this plan does **not**
> claim they are covered by it. Conflating them with tracked artifacts would overstate the gate's
> reach and would wrongly forbid the Step 0.1 index sync that every later query depends on.
>
> Deferring the tracked writes is sound because this gate is at most two steps later, nothing
> downstream consumes any of them before Step 2, and each is a Stage artifact write of exactly the
> kind this gate exists to place on the correct branch. An artifact written before this step is a
> **P-005** step-order violation as well as a P-010 one.
>
> 1. **Derive the scope slug.** `{scope-slug}` is the lower-kebab-case reduction of the selected
>    grouping's covering feature title (task-shaped intake) or of the feature, epic, or chore
>    title (feature-shaped intake), restricted to `[a-z0-9-]` and truncated at 48 characters.
>    Record it once and do not rename it later in the session. When an authoritative shipment ID
>    for this scope is **already** known, it may be used as the slug instead. Never wait for a
>    shipment ID: Step 5.5 creates the shipment after every mutation this gate exists to protect.
> 2. **Capture `stage_base_commit`.** In a **pipeline invocation** the Orchestrator captured and
>    retains it immediately before invoking Stage; Stage echoes the value it was given and does
>    not invent one. In a **direct invocation** Stage captures `git rev-parse HEAD` **before**
>    sub-step 3 creates or checks out the branch. Either way it is recorded once and never
>    re-captured later in the session.
> 3. **Resolve the branch — verify, or create.**
>    - *Pipeline invocation with a branch supplied by the Orchestrator*:
>      `git rev-parse --verify "$STAGE_BRANCH"`, then `git checkout "$STAGE_BRANCH"` when
>      `git rev-parse --abbrev-ref HEAD` does not already name it. Stage **verifies and uses** the
>      supplied branch and does not create a second one.
>    - *Direct invocation, or no branch supplied*: `git checkout -b chore/stage-{scope-slug}` from
>      the current commit, or `git checkout chore/stage-{scope-slug}` when it already exists.
> 4. **Assert, then proceed.** `git rev-parse --abbrev-ref HEAD` MUST equal the resolved Stage
>    artifact branch and MUST NOT be the default branch. On failure, halt with
>    `STAGE_BRANCH_GATE_FAIL: refusing to write staging artifacts while HEAD is {head}` and record
>    P-010. No Stage mutation proceeds past a failed assertion.
> 5. **Single worktree (P-016).** The branch is entered by switching the **one existing** worktree.
>    Stage does not run `git worktree add` and does not open a second checkout for it. The
>    time-boxed spike/research worktree exception is unrelated and unchanged.
> 6. **Idempotent on re-entry.** A resumed session already on the resolved branch verifies and
>    proceeds: no second branch, no re-derived slug, no re-captured base.

**Correction 2 — Step 5.7, the post-mutation artifact commit.** Add a new step section between the
existing Step 5.6 (Archive Consumed Stash Entries) and Step 6 (Summary):

> ### Step 5.7: Stage Artifact Commit (NON-NEGOTIABLE)
>
> Runs after **every** Stage mutation of the session — backlog items, the shipment manifest, the
> plan, deliberation, and memory records, and consumed-stash archival — and **before** Step 6
> emits the handback.
>
> 1. **Assert HEAD.** `git rev-parse --abbrev-ref HEAD` MUST equal the branch resolved at Step 1.9
>    and MUST NOT be the default branch. Otherwise halt with
>    `STAGE_COMMIT_GATE_FAIL: HEAD is {head}, expected the Stage artifact branch {stage_branch}`.
>    Do not commit the session's artifacts anyway.
> 2. **Detect out-of-root changes BEFORE staging anything (rev 10).** A root-scoped `git add`
>    cannot discharge this obligation: it *ignores* paths outside its pathspec rather than
>    reporting them, so a change outside the allowed roots would be silently left behind, not
>    detected. Run an explicit full-tree inspection first:
>
>    ```bash
>    git status --porcelain=v1 -z --untracked-files=all
>    ```
>
>    `--untracked-files=all` is **load-bearing**: the default `normal` mode collapses an untracked
>    subtree to a **directory** entry (`?? sub/`), which no per-file containment test can evaluate;
>    `all` recurses to `?? sub/deep.md`. `-z` is likewise load-bearing: without it, paths
>    containing spaces are quoted and paths containing unusual bytes are backslash-escaped, and
>    line-oriented splitting is unsafe on any of them.
>
>    **Parse the NUL-delimited records, never lines.** Split the output on NUL and drop the
>    trailing empty element. Each record is `XY` + one space + `PATH`, where `XY` is the two-column
>    status. When either status column is `R` (rename) or `C` (copy), the **next** record is the
>    ORIG_PATH and carries no `XY` prefix — consume it as part of the same entry. Both endpoints of
>    a rename or copy are affected paths and both are checked. Verified output shape:
>
>    ```text
>     D keep.txt\0R  new name.txt\0old name.txt\0?? sub/deep.md\0?? untracked.md\0
>    ```
>
>    This one command covers tracked modifications, deletions, renames, copies, and recursed
>    untracked files — the full set of ways a change can be present outside the roots.
>
>    **Halt fail-closed on any out-of-root path.** If **any** affected path — including a rename's
>    origin — does not lie under the fixed `STAGE_ARTIFACT_ROOTS` constant, halt with
>    `STAGE_COMMIT_GATE_FAIL: change outside Stage artifact roots: {path}`, record a **P-010**
>    signal, and surface the full offending list to the operator. Stage MUST NOT widen the commit
>    to include it, and MUST NOT `checkout`, `restore`, `stash`, revert, or delete it — the change
>    may be a human operator's unrelated in-flight work, and discarding it would be both a
>    destructive act without approval (Constitution VII) and a data-loss risk. **Pre-existing
>    unrelated dirt is therefore a halt, not a cleanup**: Stage stops and hands the decision to the
>    operator, leaving the working tree exactly as it found it. Gitignored paths never appear in
>    this output and are out of scope by construction.
> 3. **Stage only the allowed roots.** Only after the inspection above passes:
>    `git add -- .backlogit/ docs/plans/ docs/decisions/ docs/memory/` — the same fixed
>    `STAGE_ARTIFACT_ROOTS` constant the Orchestrator's Step 1.5 uses. The pathspec is a
>    **bound**, not a detector; detection already happened in sub-step 2.
> 4. **Commit, conventionally.** One commit per Stage session, in Conventional Commits form
>    (`chore(stage): …` or `docs(plan): …`), whose subject names the staged scope. When nothing is
>    staged the commit is a **no-op** and the step proceeds — a session that already committed
>    incrementally on this branch is compliant.
> 5. **Record the head.** `stage_head_commit := git rev-parse HEAD`, taken **after** sub-step 4, so
>    it names the commit that actually carries the artifacts.
> 6. **Derive the paths.** `stage_artifact_paths := git diff --name-only -z --diff-filter=d
>    {stage_base_commit} {stage_head_commit} -- .backlogit/ docs/plans/ docs/decisions/
>    docs/memory/` — the aggregate two-tree derivation the Orchestrator re-runs, over the base
>    **preserved** from Step 1.9.
> 7. **Stop at the commit.** Stage MUST NOT push, and MUST NOT create, update, or merge a pull
>    request. The Orchestrator (pipeline invocation) or the operator (direct invocation) owns every
>    remote operation on this branch.

**Correction 3 — register both steps and extend the Step 6 gate.** Two further edits, both
required, because a step section the Step Sequence Contract does not list is advisory in a file
whose own preamble says the checklist is what a session MUST execute:

- The **Step Sequence Contract** checklist gains `[ ] Step 1.9 — Stage artifact branch gate
  (pre-mutation)` between its Step 1.8 and Step 2 entries, and `[ ] Step 5.7 — Stage artifact
  commit (pre-handback)` between its Step 5.6 and Step 6 entries.
- Step 6's **Pre-Summary Verification Gate** gains two checks alongside the ones it already runs:
  Step 1.9 completed and `HEAD` is still the Stage artifact branch; Step 5.7 completed and
  `stage_head_commit` was recorded from its result. A summary presented without both is the same
  **P-005** class the gate already enforces for shipment assembly.

**This is an ordering contract, not a presence contract.** Both failure modes the finding describes
survive mere presence. A branch gate placed *after* harvest leaves every artifact written on the
default branch; a commit step placed *after* the summary emits a `stage_head_commit` that does not
yet exist. The ordinal positions above are therefore load-bearing and are asserted as such
(§6.3.2, AC-B3.7) — the checklist is an **ordered** list, so "between Step 1.8 and Step 2" and
"between Step 5.6 and Step 6" are mechanically checkable facts, not editorial preferences.

**Detector-neutrality (§5.3).** None of the prescribed text is a violation construct, and this was
checked rather than assumed. `checkout` is an excluded verb in construct 1, and `add`, `commit`,
`rev-parse`, `status`, `diff`, and — rev 10 — `restore` and `stash` are not push verbs at all; the
only default-branch tokens introduced appear in **prohibitive** clauses (`MUST NOT be the default
branch`, `MUST NOT write … while HEAD is the default branch`) where the marker **precedes** the
construct, which construct 2 accepts by the marker-precedes rule. Rev 10's out-of-root clause adds
`MUST NOT checkout, restore, stash, revert, or delete it`, which is likewise marker-preceded and
names no default-branch token at all. No absence assertion is introduced that the text could
self-match, because AC-B3.7 asserts **presence and order**, not absence. The closed construct set
and the fixture corpus are therefore **not** widened — the rev-6/7/8/9 non-expansion position,
unchanged.

**Why this belongs to B3 and not to a new task.** The two steps are edits to
`.github/agents/_stage.agent.md`, the single file `018.009-T` already owns, in the single domain it
already works in (contract prose), with replacement text fully specified above. Splitting them out
would add a ninth task to a shipment whose harness map is already one function per task (§5.0), and
would require a ninth harness function for assertions that belong to the surface
`TestDirectPushGate_StageAgentSurfaceClean` already scans. `018.009-T` is re-sized **M→L** on volume
alone and stays inside the 2-hour rule.

#### 6.3.2 Mechanical detection of the two operative steps (rev 9)

The assertions live in the **existing** `TestDirectPushGate_StageAgentSurfaceClean` function and use
the **existing** parse-the-installed-surface technique, adding no function, fixture, manifest entry,
or task. Three structural facts are asserted over `_stage.agent.md`:

1. **Checklist ordinals.** Parse the fenced Step Sequence Contract block, which is an ordered list
   of `[ ] Step N — …` entries. Assert a branch-gate entry exists whose index is **greater than**
   the learnings-retrieval entry's and **less than** the deliberation entry's, and a commit entry
   whose index is **greater than** the stash-archival entry's and **less than** the summary
   entry's. Index comparison — not string search — is what makes this an ordering assertion.
2. **Step sections exist and are correspondingly ordered.** Assert an ATX step heading for each of
   the two steps, and that their heading positions in the document obey the same two inequalities
   against the Step 1.8, Step 2, Step 5.6, and Step 6 headings. A checklist entry with no section,
   or a section the checklist omits, fails.
3. **Mandatory clauses per step.** Within the branch-gate section: the default-branch refusal, the
   **deferral rule** placing every earlier-step **tracked** mutation after the gate — including its
   three enumerated classes, the **reconciliation write half**, the **duplicate archival**, and
   **every pre-gate `docs/memory/` checkpoint** (naming at least the stash-classification and
   contextual-grouping/operator-selection checkpoints) — the verify-supplied-branch arm, the
   create-or-checkout arm, the `stage_base_commit` capture, and the single-worktree sentence.
   Within the commit section: the HEAD assertion, the **full-tree out-of-root detection** clause
   (the NUL-safe `--porcelain=v1 -z --untracked-files=all` inspection, the rename-endpoint rule,
   the out-of-root halt, and the do-not-revert sentence), the `STAGE_ARTIFACT_ROOTS`-bounded
   staging, the conventional-commit requirement, the `stage_head_commit` assignment taken after
   the commit, and the no-push / no-PR sentence.
4. **Intra-section ordering of the commit step (rev 10).** The detection clause's position in the
   commit section MUST be **before** the staging clause's. This is the same index-comparison
   technique as assertion 1, applied within one section rather than across the checklist, and it
   is required for the same reason: a detection clause placed after `git add` documents an
   inspection that runs too late to prevent the commit it exists to gate. Presence alone would
   satisfy a section that stages first and looks afterwards.

Assertion 1 is deliberately **not** a whole-file text search. The same self-matching hazard recorded
at §5.0 and AC-C1.6 applies here: `_stage.agent.md` will mention its own step names in the Step 6
gate and in the Role Boundary discussion, so a bare `grep` for a step name would pass with the step
section deleted. Parsing the checklist block and comparing indices is the structural analogue of
AC-C1.6's `jobs.lint.steps[]` walk, and it is chosen for the same reason.

**Regeneration exposure.** `_stage.agent.md` is autoharness-generated, so a render can revert both
steps. That is the R13 shape, and it is carried by the same mechanism: this function lives in the
non-generated `tests/integration/directpush_gate_test.go` and runs from the generated-baseline
`Test (race)` step, so a render that drops the steps turns CI **red**. Recorded as **R17**.

**AC-B3.1** The Git row Allowed cell contains no default-branch allowance.
**AC-B3.2** The Git row Allowed cell grants creation and check-out of the dedicated artifact
branch, so the table and P-010 **agree** — no grant in one that the other omits or forbids. **Rev
9**: the branch is named `chore/stage-{scope-slug}`, with a shipment ID permitted but **not
required** as the slug, so the cell does not authorize a branch whose name is underivable at the
moment Step 1.9 needs it.
**AC-B3.3** The Git row Forbidden cell is unchanged; every other Role Boundary row is unchanged
**except** the PR row, which is explicitly exempted to carry the actor note. The note attributes
an actor only — the PR row's prohibition is unweakened and no authority transfers to Stage. This
criterion governs the **Role Boundary table**; the Step 6 edit required by AC-B3.5, the
Step 5.5 / Step 6 edits required by AC-B3.6, and the Step 1.9 / Step 5.7 / Step Sequence Contract /
Step 6-gate edits required by AC-B3.7 lie outside it.
**AC-B3.4** The file agrees with the corrected P-010 text — no surviving contradiction between
agent contract and policy, in either direction.
**AC-B3.5** **Staging handback emitted (rev 6).** Step 6's session-output contract requires
`stage_branch`, `stage_base_commit` (rev 8), `stage_head_commit`, `stage_outcome`,
`stage_artifact_paths`, and `shipment_id` (when formed) — every field the Orchestrator's Step 1
consumes under AC-B1.5, with none required there and absent here. **Rev 7/8**: the contract also
fixes the **shape** of `stage_artifact_paths` — a non-empty list of concrete repository-relative
**file** paths changed between `stage_base_commit` and `stage_head_commit`, never directory
prefixes and never only the tip commit's files — and names the derivation (`git diff --name-only
-z --diff-filter=d {stage_base_commit} {stage_head_commit} --` over the four Stage artifact roots),
so the emitted set satisfies the consumer's set-equality check (AC-B1.10) by construction.
**Rev 9**: the emitted values are **produced**, not asserted — `stage_branch` and
`stage_base_commit` come from the Step 1.9 branch gate and `stage_head_commit` and
`stage_artifact_paths` from the Step 5.7 artifact commit (AC-B3.7). **Rev 10**: the emitted
`stage_outcome` and `shipment_id` form a **consistent pair** — `shipment` is emitted with a
non-empty `shipment_id`, `no-shipment` is emitted with none, and the field is always emitted
explicitly rather than left for the consumer to infer, which is what makes the consumer's
fail-closed pair validation (AC-B1.12) a check on agreement rather than a substitute for a
missing producer.
**AC-B3.6** **The `no-shipment` terminal outcome is reachable (rev 8).** Step 5.5's mandatory
shipment assembly is scoped to a harvest that **produced items**; a reviewed, valid **empty**
harvest terminates with `stage_outcome: no-shipment` instead of halting, and Step 5.5's existing
empty-harvest / unresolved-P-003 guardrail is preserved **verbatim**. Step 6's pre-summary
verification gate requires `shipment_id` **only when `stage_outcome` is `shipment`**; for
`no-shipment` it requires the complete handback record (AC-B3.5) plus an explicit terminal
statement, and the summary does **not** route the operator to Ship. **P-003 failures, harvest
failures, and a missing required shipment after a non-empty harvest continue to HALT and are never
recorded as `no-shipment`** — the outcome is a success terminal, never a failure re-label.
**AC-B3.7** **Branch-before-mutation and commit-before-handback are operative and ordered
(rev 9).** `_stage.agent.md` carries **two operative steps**, each registered in the Step Sequence
Contract checklist and each backed by its own step section (§6.3.1):
(a) a **pre-mutation branch gate** whose checklist entry and section both sit **after** the
learnings-retrieval step and **before** the deliberation step — a **fixed** position, carrying a
**categorical** deferral rule that places **every** tracked Stage artifact mutation of the session
after the gate (rev 10), enumerating at least its three classes: the **write half** of the
mandatory late-identifier reconciliation, the **unconditional duplicate-entry archival**, and
**every tracked `docs/memory/` checkpoint whose triggering milestone precedes the gate**, naming
the stash-classification and contextual-grouping/operator-selection checkpoints explicitly. The
read-only classification and grouping analysis that derive the scope slug still run first, and
gitignored disposable operations (index sync, backlogit checkpoints, hook queue) are expressly
**outside** the rule rather than claimed by it — which derives a `{scope-slug}`
without requiring a `shipment_id`, records `stage_base_commit` (echoed from the Orchestrator in a
pipeline invocation, captured from `HEAD` **before** branch creation in a direct invocation),
**verifies and uses** an Orchestrator-supplied branch or **creates and checks out**
`chore/stage-{scope-slug}` when none was supplied, **refuses to write any artifact while `HEAD` is
the default branch**, and switches the single existing worktree rather than adding one (P-016);
and (b) a **post-mutation artifact commit** whose checklist entry and section both sit **after**
the consumed-stash-archival step and **before** the summary step, which asserts `HEAD` is the
Stage artifact branch and not the default branch, **detects out-of-root changes before staging
anything** (rev 10) via an explicit NUL-safe full-tree inspection
(`git status --porcelain=v1 -z --untracked-files=all`) whose records are parsed NUL-wise rather
than by line, whose rename/copy entries contribute **both** endpoints, and which halts fail-closed
with a P-010 signal on any path outside `STAGE_ARTIFACT_ROOTS` while leaving that pre-existing
change **untouched** (no checkout, restore, stash, revert, or delete), then stages only the four
`STAGE_ARTIFACT_ROOTS`,
commits in Conventional Commits form, sets `stage_head_commit` from the **resulting** `HEAD`, and
states that Stage neither pushes nor creates, updates, or merges a pull request. Step 6's
pre-summary verification gate requires both steps complete before any summary is presented.
**Ordering is asserted structurally**, by checklist-index and heading-position comparison rather
than by text search, and — rev 10 — the commit section's detection clause is asserted to precede
its staging clause by the same index comparison applied within the section (§6.3.2). AC-B3.5,
AC-B3.6 and AC-B3.7 are the **only** changes outside the Role Boundary table; no other section of
the file is modified.

These seven criteria are carried verbatim by task `018.009-T`. The post-correction **zero-findings**
gate run is deliberately **not** duplicated here: it is owned by **AC-C2.1** on `018.011-T`, which
scans the tree once after *all four* surfaces are corrected. Asserting it at B3 — when only three
of the four surfaces have landed — would be unsatisfiable.

---

## 7. Sub-epic C — CI coupling and verification

### 7.1 Task C1 — wiring, ledger, CODEOWNERS

**One step** in job `lint`, immediately after the write-path-precondition self-test step
(referenced by description, not line number):

```yaml
      - name: Run direct-push-language gate (0NN.00X-T)
        run: bash scripts/check-direct-push-language.sh --self-test
```

Unconditionally blocking — no `continue-on-error`, no toggle variable. The step name carries a
**task-ID suffix** so it satisfies the ledger's own membership rule; rev 1 omitted this.

**Ledger** (ci.yml header): add the new gate; **generalize** the membership sentence from the
literal `011.0xx-T` to "any `NNN.NNN-T` task-ID comment or task-ID-suffixed step name" (already
stale against the `015.0xx-T` group); **drop the "all five" magic count** in favour of the
structural membership rule; record that this gate deliberately has **no** toggle variable, so a
future reader does not add one for false symmetry; record that removal of this step is **caught,
not accepted** — `tests/integration/directpush_gate_test.go`
(`TestDirectPushGate_CIWiringIsBlocking`) is not autoharness-generated and runs from the
generated-baseline `Test (race)` step (`go test -race -mod=readonly ./...`) in job `expensive`, so
a render that drops this step turns CI
**red** rather than green and the re-apply obligation is discoverable at the point of divergence
(§5.0); and record the D2 residual — `.tmpl` sources
are gitignored, so template↔artifact equality is not CI-enforceable; the gate asserts the
*invariant* over installed artifacts instead.

**CODEOWNERS**: add `/scripts/check-direct-push-language.sh @softwaresalt`. R6 rests on CODEOWNERS
review, and the file's own header promises it "Covers ALL CI-critical scripts … or none" — shipping
the gate unowned would widen an already-false claim. Also add the three already-missing gate
scripts (`check-write-path-precondition.sh`, `check-unignore-regression.sh`,
`check-depguard-fixtures.sh`) so the all-or-none contract becomes true.

No `uses:` is added, so `ci-security.instructions.md`'s 40-hex-SHA pinning MUST is satisfied
vacuously; job `lint` already declares `permissions: contents: read` and
`persist-credentials: false`, so the least-privilege MUSTs need no new work.

**AC-C1.1** The step is present in job `lint`, **blocking** (no `continue-on-error`, no toggle
variable), and task-ID-suffixed.
**AC-C1.2** `.github/workflows/ci.yml` remains valid YAML
(`python -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))"` passes) **and**
`ci-gate` still transitively depends on job `lint`, so the new step is genuinely merge-blocking.
**AC-C1.3** Ledger names the gate, uses the generalized membership rule, carries no magic count,
and records both the no-toggle decision and the D2 residual.
**AC-C1.4** CODEOWNERS coverage is asserted **structurally, not by count** (rev 3): every script
invoked by a step in `.github/workflows/ci.yml` has an owner line in `.github/CODEOWNERS`. The
header's fixture-data carve-out is generalized from `scripts/testdata/gitignore/**` to
`scripts/testdata/**` so the new fixture tree is covered by the same stated rationale.
**AC-C1.5** The script header states **verbatim** that this is an **anti-accident control, not an
anti-adversary control** (CI runs it from the PR's own head), and discloses that the job-`lint`
wiring step sits **inside** the autoharness render boundary (deliberation D6) **and** that its
removal is **detected — not accepted — by** `TestDirectPushGate_CIWiringIsBlocking` (§5.0), so the
disclosure points a future reader at the surviving enforcement path rather than at a silent failure.
Added at rev 4 — review round 3 — to close a plan↔task drift: `018.010-T` already carried this
criterion while the plan did not state it. **Rev 5** replaced the "accepted residual" half of the
disclosure per the R5 reclassification.
**AC-C1.6** **Render-survival assertion — structural and non-vacuous (rev 5).**
`TestDirectPushGate_CIWiringIsBlocking` determines step presence by **parsing**
`.github/workflows/ci.yml` and walking `jobs.lint.steps[]` for a step whose `run` invokes
`scripts/check-direct-push-language.sh` — **never** by a file-wide text search, which would
self-match the LOCAL DIVERGENCE ledger comment naming the same script and pass vacuously after the
step itself was deleted (§5.0; prior art
`docs/compound/2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md`). Non-vacuity is
proven by **execution**: the test writes a copy of the current `ci.yml` with that step removed into
`t.TempDir()` and asserts the same check **fails** against the copy. **No edit is made to the real
tree** — `git status --porcelain` stays empty — the same no-dirty-tree-edit rule as AC-A2.4 and
AC-C2.2.

### 7.2 Task C2 — coupling verification

This numbered list is the **single verification contract** for C2. Task `018.011-T` carries it
verbatim; the two MUST NOT diverge (rev 4 reconciliation — the rev-3 task text asserted a
dirty-tree mutation proof that contradicted this section's fixture-backed one, and the two used
different numbering for the same criteria).

- **AC-C2.1** Full-corpus bare scan at HEAD exits **0** with **zero findings**; the resolved corpus
  file count and the findings count are recorded in the PR description as structured evidence.
- **AC-C2.2** **Mutation proof — fixture-backed, NO dirty-tree edit.** Reintroduction-detection for
  both violation shapes is proven by `--self-test` against the committed fixture corpus:
  `directpush-reject-attempt-first.md` locks the literal 3e **push** construct and
  `directpush-reject-commit-on-default.md` locks the **default-branch commit authorization**
  construct, and the run reports every fixture **by name** with actual == manifest verdict. **No
  temporary edit is made to the real tree at any point**; `git status --porcelain` is empty
  immediately before and after the proof, and the evidence recorded is the `--self-test` output,
  not a narrated local edit. A reintroduction proof performed by mutating the committed tree is
  **explicitly rejected**: it is unrecorded, unreproducible in CI, and leaves a window in which the
  repository contains the very construct this gate exists to forbid.
- **AC-C2.3** **Self-match negative proof, executed.** In the same bare scan, the three real
  marker-free prohibition constructs in the corrected tree — the post-change P-010 **Stage**
  prohibition bullet, P-010's **Ship MUST NOT** bullet "Commit or push directly to `main`" in
  `workflow-policies.md`, and `_ship.agent.md`'s Role Boundary `| Git |` Forbidden cell — all score
  **accept**. Each is located by **file plus quoted construct text**, never by line ordinal
  (rev 8), so the criterion survives any edit that shifts line numbers in those files. Rev 1's
  AC3.3 (ci.yml step names) was vacuous because `.github/workflows/**` is never in the corpus.
  Prior art:
  `docs/compound/2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md`.
- **AC-C2.4** **Four-surface agreement.** The installed `.github/**` artifacts and the corrected
  policy agree — no surviving contradiction across the four §2 authorization surfaces. **Rev 6**:
  the **staging handback contract** also agrees across its two surfaces — every field
  `_orchestrator.agent.md` Step 1 requires (AC-B1.5) is emitted by `_stage.agent.md` Step 6
  (AC-B3.5), with no field required in one and absent from the other. **Rev 7**: that agreement
  extends to the **shape** of `stage_artifact_paths` — the producer (`_stage.agent.md` Step 6,
  AC-B3.5) and the consumer (`_orchestrator.agent.md` Step 1.5 step 4, AC-B1.10) declare the same
  concrete-file contract and the same derivation command, and **neither** surface retains a
  directory-prefix default. A field agreeing in name but disagreeing in shape is the same
  defect class one level down, and passes a name-only comparison. **Rev 8**, two further
  agreements: (a) both surfaces carry `stage_base_commit` and the **same aggregate** two-tree
  derivation, so neither retains a single-commit form; and (b) the `no-shipment` **outcome value
  itself** is agreed — the consumer's arm is keyed on an outcome the producer can actually emit
  (AC-B3.6). A verification arm keyed on an outcome no producer can reach is not a passing check
  but an **unreachable** one, and it passes every agreement test that compares only field names.
  **Rev 9**, the producing-action agreement: every handback field the consumer reads has a
  **producing step** in the producer surface, and those steps are **ordered** correctly —
  `_stage.agent.md` carries an operative branch gate before its first artifact write and an
  operative artifact commit before its summary (AC-B3.7), so `stage_branch`, `stage_base_commit`,
  `stage_head_commit`, and `stage_artifact_paths` name a branch that was created and commits that
  were made. Both surfaces also agree on the **branch-naming form**: neither requires a
  `shipment_id` to name the branch (AC-B2.2, AC-B3.2). A field that is declared, shaped, and
  emitted but never *produced* is the same defect class a third level down, and it passes both a
  name-only and a shape-only comparison. **Rev 10**, two final agreements: (a) the
  outcome/`shipment_id` **pair** is agreed — the producer emits `stage_outcome` explicitly and
  consistently with `shipment_id` (AC-B3.5), and the consumer **preserves** a reported outcome,
  derives only when absent, and **halts fail-closed** on an unknown value or an inconsistent pair
  before selecting an arm (AC-B1.12); neither surface retains the unconditional
  derive-from-`shipment_id` rule that silently reclassified a reported `shipment` run as a
  successful `no-shipment` one. (b) The producer's **pre-gate deferral** and **out-of-root
  detection** clauses are present and correctly ordered (AC-B3.7) — the deferral rule is
  categorical over every tracked mutation and names its three classes while expressly excluding
  gitignored disposables, and the commit step's full-tree inspection is positioned **before** its
  staging clause, so the handback the consumer reads describes a commit that could not have
  silently omitted or silently widened its contents. A field whose value is agreed but whose
  *production order* admits a write to the wrong branch is the same defect class a fourth level
  down, and it passes name, shape, and producing-step comparisons alike.
- **AC-C2.5** `markdownlint` exits **0** on every changed markdown file (P-008).
- **AC-C2.6** Branch and PR merge-commit rules intact — P-009, P-011, P-016 textually unchanged.
- **AC-C2.7** Constitution Quality Gates run **in declared order, none skipped, as separate
  commands** (never chained with `&&`, which would short-circuit and mask which gate failed):
  `gofmt -l .` empty, then `go vet ./...`, then `go test ./...`, then `go build ./...` — each
  exits 0 and each result is recorded independently. **Rev 4**: `go test ./...` is no longer a pure
  regression check — it now includes the `tests/integration/directpush_gate_test.go` harness (§5.0),
  which must be **green** in full, so all eight per-task functions pass here.
- **AC-C2.8** **Regeneration-resistant enforcement, recorded (rev 5).** The evidence records that
  the full-corpus scan of AC-C2.1 is executed by `TestDirectPushGate_FullCorpusClean` from job
  `expensive` (`name: test`)'s **generated-baseline** `Test (race)` step, whose exact command is
  `go test -race -mod=readonly ./...` (rev 8 — cited precisely, because "the `go test` step" does
  not identify a step in this workflow), and **not solely** by the job-`lint` step added in C1.
  This is **separate from**, and does not replace, AC-C2.7's constitutional `go test ./...`
  requirement: that gate is run locally in declared order, this one names the CI step that survives
  a render. Both halves of the §5.0 invariant are asserted green in the same run: **(a)** zero
  findings over the installed surfaces, and **(b)** the job-`lint` step still present, structurally
  and non-vacuously per AC-C1.6. A future render that drops the C1 step therefore fails (b)
  **loudly** while (a) continues to execute — which is what makes §1's "cannot silently reopen the
  gap" true as written rather than aspirational.

## 8. Verification commands

```bash
bash scripts/check-direct-push-language.sh --self-test   # fixtures + selection + real-tree
bash scripts/check-direct-push-language.sh               # verdict; expect 0 findings
go test ./tests/integration -run TestDirectPushGate      # the P-002/P-004 harness (§5.0)
# R5 invariant (§5.0): (b) invocation survives the render boundary + (a) detector still runs
go test ./tests/integration -run 'TestDirectPushGate_(CIWiringIsBlocking|FullCorpusClean)'
markdownlint "**/*.md"
python -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))"
actionlint .github/workflows/ci.yml   # LOCAL, operator-run — no CI counterpart exists in this repo
git status --porcelain                # MUST be empty around AC-C2.2 — no dirty-tree mutation proof
# Rev 7/8 — the no-shipment arm's primitives, verified against this repository:
git show origin/main:docs/plans/           # exits 0 on a TREE — why `show` is insufficient (AC-B1.10)
git cat-file -t origin/main:docs/plans/    # prints `tree`  → rejected
# Rev 8 aggregate derivation (two-tree, stable across the merge). NOT the range form
# `origin/main..{head}`, which is empty by construction once the head is an ancestor of origin/main:
git diff --name-only -z --diff-filter=d {stage_base_commit} {stage_head_commit} -- \
  .backlogit/ docs/plans/ docs/decisions/ docs/memory/
git merge-base --is-ancestor {stage_base_commit} {stage_head_commit}   # base→head ancestry (check 3)
# Rev 9 — the producing side, exercised by the B3 harness function rather than by a shell command:
go test ./tests/integration -run TestDirectPushGate_StageAgentSurfaceClean   # branch gate + commit step, present AND ordered
# Rev 10 — per-task activation (§5.0.1). Default mode: surfaces present + the A1 anchor.
# Targeted mode: force exactly one task's assertions, which is how each task's red is observed.
go test ./tests/integration -run '^TestDirectPushGate_StageAgentSurfaceClean$' -gatetask=B3
# Rev 10 — AC-A2.5's two exactly-anchored invocations (the second MUST print `--- FAIL:`;
# a non-matching -run pattern exits 0 with `[no tests to run]`, so exit code alone proves nothing):
go test ./tests/integration -run '^TestDirectPushGate_SelfTestDriverEnumeratesFixtures$' -gatetask=A2
go test ./tests/integration -run '^TestDirectPushGate_SelfTestPasses$' -gatetask=A3   # expect non-zero
# Rev 10 — Step 5.7's out-of-root detection primitive. `--untracked-files=all` is load-bearing:
# the default `normal` mode emits a DIRECTORY entry (`?? sub/`) that no per-file check can evaluate.
git status --porcelain=v1 -z --untracked-files=all
gofmt -l .          # must print nothing
go vet ./...
go test ./...
go build ./...
# Constitution gate order: run as SEPARATE commands, in this order, none skipped.
# Do NOT chain with && — a chain masks which gate failed and short-circuits the rest,
# which is exactly what "in declared order, none skipped" forbids (AC-C2.7).
```

`actionlint` is **not** installed or invoked by any workflow here; AC-C1.2 (`yaml.safe_load`) is
the CI-enforced half. Rev 1 implied CI enforcement that does not exist. If a verification tool is
absent locally, record `TOOL_DEGRADED` rather than silently skipping (P-012).

## 9. Plan hardening signals

Edits a NON-NEGOTIABLE merge-gate contract (Step 1.5), the authoritative policy registry (P-010),
and the agent's operative Role Boundary table; adds a **blocking** CI gate; carries a documented
in-repo self-matching failure mode; touches autoharness-generated files with untracked templates.
**Rev 9** adds two **NON-NEGOTIABLE operative steps** to the Stage agent's own Step Sequence
Contract, which changes what every future Stage session must execute — a wider behavioural blast
radius than the prose corrections in revisions 1–8, and the reason AC-B3.7 constrains ordering
rather than presence alone. **Rev 10** adds a **fail-closed halt on a new input class** (the
outcome/shipment pair), makes a pre-mutation deferral rule **categorical** rather than enumerated,
and changes the **harness activation lifecycle** that every task's red→green boundary depends on —
the last of which interacts directly with Ship's Step 4.3 full-suite gate and is therefore the
rev-10 element most in need of the mutation-proof checks §5.0.1 specifies.

## 10. Constitution Check

Required by `.github/instructions/constitution.instructions.md` (Governance). Omitted in rev 1.

| Principle | Status | Note |
|---|---|---|
| I. Safety-First Go | **Satisfied** | Rev 4: one Go test file is added (§5.0) — no production Go surface; `go vet ./...` and job `expensive`'s `Test (race)` step (`go test -race -mod=readonly ./...`) cover it |
| II. Test-First (NON-NEGOTIABLE) | **Satisfied — no deviation** | Rev 4 removes rev 3's declared deviation. See §10.1 |
| III/IV. Workspace isolation, CLI containment | **N/A** | No runtime code |
| V. Structured Observability | **Satisfied** | `::notice::` mode emission; AC-C2.1 records gate verdict as PR evidence |
| VI. Single Responsibility | **Satisfied** | One gate, one invariant; 8 tasks each single-domain |
| VII. Destructive Command Approval | **N/A** | No destructive operation |
| VIII. Explicit Safety Modes | **N/A** | — |
| IX. Git-Friendly Persistence | **Satisfied** | Additive Amendment Log row 1.25.0; §11 records the stale header `Version` field |
| X. Agent Context Efficiency | **Satisfied** | Ledger duplication reduced (§7.1 drops the magic count; residual stated once in the script header, cross-referenced elsewhere) |
| XI. Merge Commit History (NON-NEGOTIABLE) | **Satisfied** | P-009 untouched; **AC-C2.6** (rev 8 — this row previously cited AC-C2.5, which is the markdownlint criterion; AC-C2.6 is the one asserting P-009/P-011/P-016 textually unchanged) |
| Task Granularity (NON-NEGOTIABLE) | **Satisfied** | Rev 3 splits sub-epic A into A1 (fixture data) / A2 (driver + stub) / A3 (detector) and B into B1/B2/B3 — 8 tasks, each single-domain and within 2 h. C1 combines ci.yml + CODEOWNERS + ledger as one **CI-configuration** domain (no code/doc mixing) |
| Development Workflow #2 (backlog-driven) | **Satisfied** | These tasks are not a static markdown list: they are harvested into `.backlogit/` under the covering chore with the full P-003 lineage chain before any execution |

### 10.1 P-002 / P-004 compliance — no deviation (rev 4)

**Rev 3 declared a deviation here and was wrong to.** It asked Ship to apply `harness-ready` from a
substitute bash harness by prose. `_ship.agent.md` Step 2 cannot honour that: it invokes
`harness-architect` for every queued task lacking the label and halts unless every one carries it,
and `harness-architect` applies the label only after a Go `_test.go` harness compiles under
`go vet ./...` and fails under `go test ./...`. A Stage-applied label would forge P-004's
postcondition; withholding it would drive `harness-architect` to invent Go scaffolding for markdown
and bash tasks. Both outcomes are P-002/P-004 failures.

**Rev 4 satisfies the existing harness path literally, with no policy change.**

- **Principle deviated**: **none**. P-004's precondition is met as written — `go vet ./...` exits 0
  and `go test ./...` exits non-zero with the expected `not implemented` marker across all eight
  harness functions after H0.
- **Harness**: `tests/integration/directpush_gate_test.go` (§5.0), a real Go regression harness
  whose subject is the bash gate. Not invented scaffolding: `tests/integration/build_script_test.go`,
  `start_script_test.go`, and `output_path_guard_test.go` are existing tests in this repository
  whose entire subject is a non-Go script, and `internal/apperr/taxonomy_drift_test.go` is an
  existing Go drift guard over a non-runtime invariant.
- **Red phase**: H0's structural stub (`scripts/check-direct-push-language.sh` printing
  `not implemented: direct-push detector` and exiting non-zero) is the shell analogue of
  `harness-architect`'s `panic("not implemented: <reason>")` stub. The detector is observed failing
  before the implementation that makes it pass — Principle II's requirement, met by execution.
  **Rev 10**: the *default-suite* red at H0 comes from the activation-independent A1 anchor
  (§5.0.1), and each remaining task's red is observed **per task** at claim time through that
  task's `-gatetask`-selected `harness_cmd`. Every task still has an executed red before its
  implementation; what changes is that the eight reds are no longer required to be **simultaneous**,
  which is what made the suite incompatible with Ship's Step 4.3 full-suite gate. No red is
  removed, none is asserted from prose, and Principle II is not relaxed.
- **Ship's P-002 precondition**: satisfied by `harness-architect` at Step 2 in the normal way.
  **Stage applies no `harness-ready` label and this plan asks Ship for no prose exemption.**
  `018-F.custom_fields.harness_status` stays `pending` until H0 records
  `Compilation: PASS` / `Red Phase: CONFIRMED`.
- **Rejected alternative 1 — amend the governing workflow.** A first-class non-Go harness route
  needs edits to P-002/P-004 in `workflow-policies.md`, Step 2 in `_ship.agent.md`, and the
  `harness-architect` skill. The latter two are autoharness-generated from gitignored `.tmpl`
  sources (D2), so the change would be reverted by the next render (R4/R5); and amending P-002/P-004
  is a workspace-wide weakening of test-first binding every future shipment, outside both this
  release unit's branch-policy contract and the governing deliberation.
- **Rejected alternative 2 — ship the detector and fixtures together** (rev 1's Task 1). It never
  observes the detector failing, so a detector that is right for the wrong reason ships undetected.
- **Rejected alternative 3 — rewrite the gate in Go** to avoid the shell subject entirely. It would
  discard the five-gate bash precedent, the shared `--self-test` house style (§5.1), and the CI
  step/CODEOWNERS conventions (§7.1), and put governance-document scanning inside the product
  module. The Go *harness* over a bash *gate* keeps both contracts.

### 10.2 P-016 / P-014 / P-018 disposition for the staging PR

P-016's Statement is scoped to **implementation** branches: "there must be exactly one agent-owned
**implementation branch/worktree** in use." A Stage artifact branch carries no source, test, or
config change, so it is not an implementation branch and does not consume the single-active
implementation slot during planning overlap; it is likewise **non-release-unit under P-001**.

Rev 3 does not leave this as an assertion: B2 adds the same clarification **to P-010's text**, so
the disposition is contract-visible at the point an agent reads it rather than inferred from a
plan. No P-016 amendment is needed — the clarification records an existing scope boundary rather
than creating an exception.

P-014 readiness and P-018 thread resolution apply to the staging PR as to any PR; the
**Orchestrator** produces the readiness record under Step 1.5, since it owns push/PR/merge.

## 11. Risks, residuals, out of scope

| # | Risk | Mitigation |
|---|---|---|
| R1 | Detector self-matches the corpus's own prohibition text | §5.3 structural prohibition resolution; AC-C2.3 **executes** the proof on three real marker-free constructs |
| R2 | Gate wired while violations present → `main` red | §4 ordering; CI wiring is C1, after B1–B3 |
| R3 | Detector too narrow → fails open | 9 reject fixtures across push/commit/prose/placeholder/wrapped-lazy/marker-bearing/`rather than`/mixed-block shapes; bijection + non-vacuity assertions |
| R4 | Regeneration restores a violation | Gate asserts the invariant over installed artifacts (D2) → CI red, scoped to the §5.2 corpus |
| R5 | Regeneration wipes the job-`lint` CI step → gate silently disabled | **MITIGATED at rev 5 — no longer an accepted residual.** Enforcement also runs through `tests/integration/directpush_gate_test.go` (not generated → survives a render), invoked by the **generated-baseline** `Test (race)` step — `go test -race -mod=readonly ./...` in job `expensive` (`name: test`), which a render *re-emits* — and any render that rewrites `ci.yml` sets `changes.code == 'true'` so that job cannot skip. `TestDirectPushGate_FullCorpusClean` keeps executing the detector (requirement **a**); `TestDirectPushGate_CIWiringIsBlocking` fails **red** when the step is gone (requirement **b**). See §5.0, AC-C1.6, AC-C2.8 |
| R6 | Gate mistaken for a security boundary | Header states verbatim: **anti-accident control, not an anti-adversary control** — CI runs it from the PR head. CODEOWNERS (AC-C1.4) is the named mitigation |
| R7 | Legitimate line trips the detector | Narrow the **pattern** or document-and-exclude a path; never reduce corpus coverage |
| R8 | 6th hand-copied gate scaffold (no `scripts/lib` bash helper exists) | **ACCEPTED RESIDUAL**, recorded here with the existing clone-divergence precedent (stash `6C24E2E4`); extracting a shared helper is out of scope under C1 |
| R9 | A **fifth** authorization surface exists outside the §5.2 corpus | AC-A3.4 records the first full-corpus scan (file count + complete finding list) **before** B1 begins, so an unknown surface surfaces at A3 rather than at C2 |
| R10 | The §5.0 Go harness makes `go test ./...` depend on `bash` | **ACCEPTED RESIDUAL** (rev 4). CI runs on `ubuntu-latest`; every existing gate script already requires `bash` locally. The harness **fails rather than skips** when `bash` is absent (§5.0), deliberately, so the P-004 red phase can never be satisfied by a silent skip (P-012) |
| R11 | The §5.0 harness is a sixth CI-relevant surface a future render could disturb | Low: it is a plain `_test.go` file under `tests/integration/`, not autoharness-generated, and `go test -race -mod=readonly ./...` already runs in job `test`. No new CI wiring, no ledger entry, no CODEOWNERS line needed. **Rev 5** makes both of those properties **load-bearing rather than incidental** — R5's mitigation depends on the harness being non-generated *and* on that `go test` step being generated-baseline, so §5.0 now states and justifies both explicitly |
| R12 | The harness itself is deleted or neutered, disabling both halves at once | **ACCEPTED — but VISIBLE, not silent (rev 5).** Requires editing non-generated tracked files (`directpush_gate_test.go` and/or the detector) in a reviewable pull-request diff; **no render can do it**. `CODEOWNERS` is deliberately **not** claimed as the mitigation here — its own header records it is advisory-only until code-owner review is required on `main`, an operator action not taken. This is strictly narrower than rev 4's R5: the failure mode moves from *silent regeneration* to *visible human/agent edit*, which is what §1 actually promises to exclude. **If a future autoharness version begins generating `tests/**`, the §5.0 invariant must be re-verified** |
| R13 | A render reverts the rev-6 handback/discovery contract, restoring main-only discovery — a defect the **detector cannot see**, since it is neither §5.3 construct | **MITIGATED, same mechanism as R5 (rev 6).** The two per-surface harness functions carry it: `TestDirectPushGate_OrchestratorSurfaceClean` asserts the branch-specific discovery, the handback requirement, both step-4 arms, and the **absence** of the default-branch self-range literal; `TestDirectPushGate_StageAgentSurfaceClean` asserts the Step 6 emission. Both live in the non-generated `directpush_gate_test.go` and run from the **generated-baseline** `Test (race)` step — `go test -race -mod=readonly ./...` in job `expensive` (§5.0) — so a render that reverts either surface fails CI **red**. The literal-absence assertion is sound only because the corrected text states its prohibition descriptively (§6.1.2, self-match hazard; AC-B1.6) — a prohibition quoting the forbidden range would self-match and silently vacate the assertion |
| R14 | A verification loop passes over targets that are **always present**, proving nothing — the rev-7 finding | **MITIGATED (rev 7, widened rev 8).** Structural, not procedural: a two-tree `git diff --name-only` cannot emit a directory, `cat-file -t` must print `blob`, the derived set must be non-empty, and a reported set must **equal** the aggregate changed-file set. **Rev 8** removes the residual rev 7 accepted rather than closed: the set is now the **aggregate** `{stage_base_commit}`→`{stage_head_commit}` diff, so a Stage session spread over several commits has **every** artifact read as a file on `origin/main`, not only its tip commit's. `TestDirectPushGate_OrchestratorSurfaceClean` additionally asserts the **absence** of a `git show origin/main:{path}` loop, of any directory-prefix default, and (rev 8) of any single-commit `diff-tree` derivation in the no-shipment arm, so a render or edit that restores either earlier form fails CI **red**. **Residual, recorded not hidden**: if a Stage session merges `origin/main` into its artifact branch mid-flight, the two-tree diff also reports files that arrived from the default branch. They are bounded to `STAGE_ARTIFACT_ROOTS` by the pathspec and are already on `origin/main`, so the per-file check passes — the effect is a few **extra** verified paths, never a missed one. Widening to a range form is **rejected**, not overlooked: `origin/main..{stage_head_commit}` is empty by construction post-merge, which is exactly the vacuous loop this row exists to close |
| R15 | The `no-shipment` terminal outcome becomes a dumping ground for genuine failures | **MITIGATED by construction (rev 8).** AC-B3.6 makes `no-shipment` a **success** terminal available only to a reviewed, valid **empty** harvest, and states explicitly that P-003 lineage violations, harvest failures, and a missing required shipment after a **non-empty** harvest continue to **halt**. Step 5.5's existing empty-harvest / unresolved-P-003 guardrail is preserved **verbatim** rather than rewritten, so the discriminator between "nothing to ship" and "failed to ship" is the one already in force. `TestDirectPushGate_StageAgentSurfaceClean` asserts the halt clauses are still present alongside the new terminal. **Residual**: an agent could still mis-classify a failed harvest as empty — a judgement error the contract narrows but cannot eliminate; the Orchestrator's step-4 no-shipment arm is the backstop, since a mis-classified run with no artifacts on `origin/main` fails check 5 or check 6 |
| R16 | `stage_base_commit` is captured wrongly — too late (missing early commits) or too early (over-wide set) | **BOUNDED (rev 8).** Capture is a single `git rev-parse HEAD` by the **Orchestrator immediately before it invokes Stage**, at a point where Stage has not yet run and cannot have committed — so "too late" requires the Orchestrator to reorder its own Step 1, which `TestDirectPushGate_OrchestratorSurfaceClean` asserts against. "Too early" is bounded by the `STAGE_ARTIFACT_ROOTS` pathspec and fails **conservatively** (extra paths verified, never fewer). The retained value is authoritative over any Stage-reported one, so a wrong report cannot shrink the set; and base→head ancestry (check 3) rejects an unrelated or rewritten base outright rather than silently producing a nonsense diff. **Rev 9** supplies the direct-invocation half the row previously left implicit: Step 1.9 captures the base from `HEAD` **before** it creates or checks out the branch, and never re-captures it, so the same "too late" argument holds on the unbracketed path |
| R17 | A render reverts the rev-9 operative steps, restoring a Stage agent that is *permitted* to branch and commit but never *instructed* to — a defect the detector cannot see, since it is neither §5.3 construct | **MITIGATED, same mechanism as R13 (rev 9).** `TestDirectPushGate_StageAgentSurfaceClean` asserts both steps present **and ordered**, by parsing the Step Sequence Contract checklist and comparing entry indices and heading positions rather than by text search (§6.3.2). It lives in the non-generated `directpush_gate_test.go` and runs from the generated-baseline `Test (race)` step (`go test -race -mod=readonly ./...`, job `expensive`), so a render that drops either step fails CI **red**. The structural form is load-bearing: `_stage.agent.md` names its own steps in the Step 6 gate, so a bare name search would pass with the step sections deleted — the AC-C1.6 hazard, one file over |
| R18 | An artifact is written before the branch gate runs, landing on the default branch | **CLOSED BY ORDERING, with a narrow residual (rev 9).** The gate occupies a **fixed** checklist position (after learnings retrieval, before deliberation), and every mutation that would otherwise precede it — the Step 1 deferred-expansion duplicate archival, the session's first memory checkpoint — is **deferred until after** it by an explicit deferral rule rather than by racing the gate earlier. **Rev 10 closes the under-inclusion**: the rule was an *enumeration* of two mutations, while the installed Stage contract also schedules a stash-classification checkpoint, a contextual-grouping/operator-selection checkpoint, and the **write half** of the mandatory late-identifier reconciliation before that ordinal — each of them tracked, and each therefore still legal on the default branch under the rev-9 wording. The rule is now **categorical** ("every tracked Stage artifact mutation"), with those classes enumerated as illustration rather than as the closed set, and with gitignored disposables expressly excluded rather than silently claimed. The rev-9 draft's alternative, letting the gate fire early against a provisional slug, was **rejected**: it contradicts the Step Sequence Contract's own "execute in order" semantics, leaves the real position unassertable, and re-admits the default-branch write it was meant to prevent. `TestDirectPushGate_StageAgentSurfaceClean` asserts both the ordinal and the deferral clause. **Residual, recorded not hidden**: an agent that writes before reaching Step 1.9 violates the contract — now a **P-005** step-order violation as well as P-010 — without tripping the harness, which reads the document, not the run. The Orchestrator's step-4 arm is the backstop: artifacts written on the default branch leave `stage_branch` UNRESOLVED and the aggregate diff empty, which halts the gate rather than passing it |
| R19 | The rev-10 activation mechanism degenerates — every task's assertions skip forever and the suite is green while proving nothing | **BOUNDED BY FOUR EXECUTED CHECKS (rev 10, §5.0.1).** MP0 fixes the table at exactly eight rows and fails a function whose ID is missing; MP1 evaluates every predicate against an empty `t.TempDir()` root and requires **false**, so a constant-true row fails rather than silently activating; MP2, inside `TestDirectPushGate_FullCorpusClean`, requires **all eight** predicates to hold once C1's surface exists, which converts any post-landing regression from a silent skip into a **failure**; MP3 makes every skip a `t.Skip` naming the task ID and the unmet predicate, and makes an unknown `-gatetask` value **fail** rather than skip. **Residual, recorded not hidden**: between A1 and C1 a landed-then-regressed surface would re-skip rather than fail, because MP2 does not bind until C1 exists. That window is inside one shipment's execution, Ship re-runs each task's `harness_cmd` with forced activation regardless, and the alternative — binding MP2 earlier — would make the suite red for every not-yet-started task, which is the deadlock rev 10 exists to remove |
| R20 | Step 5.7's out-of-root halt fires on a human operator's unrelated in-flight edit, blocking the Stage commit | **ACCEPTED AND DELIBERATE (rev 10).** Fail-closed is the correct direction: the alternative is a Stage commit that silently omits a change the operator believed was being saved, or one that silently widens beyond `STAGE_ARTIFACT_ROOTS`. The halt names every offending path, and Stage **leaves the change exactly as found** — no checkout, restore, stash, revert, or delete — so nothing is lost and the operator decides. Discarding it would be a destructive act without approval (Constitution VII). Gitignored paths never reach this check, so routine index/checkpoint/hook-queue churn cannot trip it |

**Known pre-existing defects, recorded and out of scope**: `workflow-policies.md` header
`**Version**: 1.0.0` is stale against Amendment Log 1.24.0; Step 1.5's a–e sub-steps are lazy
paragraph continuations rather than a real nested list (so AC-B1.3 is **textual**, not structural);
`_orchestrator.agent.md`'s provenance footer names `orchestrator.agent.md.tmpl` while the real
template is `_orchestrator.agent.md.tmpl`; `ci-topology-check.sh` cites two non-existent
`docs/pipeline-topology-gate*.md` files; job `lint` runs only when `changes.outputs.code == 'true'`
(benign — the `code` filter excludes only `docs/**`, `.backlogit/**`, `.backlog/**`,
`.autoharness/**`, so any `.github/**` edit triggers it).

**Out of scope**: reverting/rewriting `fdff9e4`; editing untracked `.copilot/` or
`.autoharness/staging/` templates; GitHub branch-protection configuration; a new policy ID;
**any amendment to P-002, P-004, `_ship.agent.md` Step 2, or the `harness-architect` skill**
(rev 4 — §10.1 rejected alternative 1); a shared `scripts/lib/gate-common.sh`; the missing compound
entry on generated-artifact divergence; the upstream template defect report.

---

## 12. Plan review record

**Gate**: `plan-review`, 7 personas (Architecture Strategist, Correctness, Scope Boundary Auditor,
Constitution, Maintainability, Template Integrity, Schema-CLI-Docs Coupling).

| Attempt | Revision reviewed | Verdict | Findings |
|---|---|---|---|
| 1 | rev 1 | **FAIL** | 3 x P0, 11 x P1 |
| 2 | rev 2 | **FAIL** | all round-1 P0s confirmed resolved; 1 new P0 (fourth surface), 5 P1 detector defects |
| — | rev 3 | **ADVISORY — accepted** | all P0/P1 remediated in rev 3; residual P2/P3 recorded below |
| — | rev 4 | **PR #54 current-HEAD review, cycle 2** | 2 x P-021 C1 same-contract blockers, both resolved in rev 4 (see below) |
| — | rev 5 | **PR #54 current-HEAD review, cycle 4** | 1 finding — R5 contradicted §1 and the feature DoD; resolved in rev 5 (see below) |
| — | rev 6 | **PR #54 current-HEAD review, cycle 5** | 1 finding — Step 1.5's discovery could not see the Stage branch; resolved in rev 6 (see below) |
| — | rev 7 | **PR #54 current-HEAD review, cycle 6** | 2 findings — the no-shipment arm verified directory prefixes that always exist (false pass); the session memory cited an obsolete plan revision and a nonexistent section. Both resolved in rev 7 (see below) |
| — | rev 8 | **PR #54 adversarial review** (anchor GPT-5.6 Sol + GPT-5.4-mini + Claude Sonnet 5 + Claude Opus 5; no route degradation) | 4 P1 blockers + 5 P2/P3 precision findings, all classified **same-contract completion** under P-021. All resolved in rev 8 (see below) |
| — | rev 9 | **PR #54 current-HEAD review, cycle 7** | 1 visible Copilot finding — B3 granted the branch permission and declared the handback but added no operative step that creates the branch or commits the artifacts. Resolved in rev 9 (see below) |
| — | rev 10 | **PR #54 second adversarial review** (anchor GPT-5.6 Sol + GPT-5.4-mini + Claude Sonnet 5 + Claude Opus 5; no route degradation) | 4 visible Copilot threads (3 P1, 1 P2) + 1 adversarial-only P1 harness-lifecycle finding, all classified **same-contract completion** under P-021. All resolved in rev 10 (see below) |

**Round-1 P0s (resolved in rev 2)**: missing third authorization surface `_stage.agent.md` L42;
self-deadlocking P-010 wording (mandated "merged via PR" against Stage's "must not create, push,
or merge pull requests"); false claim that deleting 3e loses no behaviour.

**Round-2 P0 (resolved in rev 3)**: Step 1.5 **preamble** is a fourth authorization surface —
commit-shaped, marker-free, in-corpus. Addressed by §2 (four surfaces) and B1's preamble rewrite
to a verification/postcondition form.

**Round-2 P1s (resolved in rev 3)**: block-level verdicts masked mixed blocks (→ per-construct);
ATX-heading-only prohibition context would have flagged the real bold lead-in shapes (→ bold
lead-in + column header); row-scoped table matching paired Allowed verbs with Forbidden tokens
(→ cell-scoped); unguarded `git symbolic-ref` aborts under `set -euo pipefail` (→ guarded with
`main` fallback); `pull`/`checkout` not excluded (→ added); Task A1 spanned 11 files (→ A1/A2/A3).

**Residual non-blocking findings (P2/P3), accepted**: the stale `**Version**: 1.0.0` header and
other pre-existing defects in §11; the `(0NN.00X-T)` ci.yml step-name suffix is a **placeholder**
to be resolved by Ship at C1 against the then-current ledger numbering; the LOCAL DIVERGENCE
membership rule's literal `011.0xx-T` wording may need a wildcard/range form; no compound entry
yet exists for generated-artifact divergence (out of scope, §11).

**Accepted at attempt 2 under the 2-cycle limit.** No P0 or P1 remains open.

**Rev 4 — PR #54 current-HEAD review, remediation cycle 2.** Two unresolved threads, both P-021 C1
same-contract blockers, both fixed rather than deferred:

| Thread | Finding | Resolution |
|---|---|---|
| `PRRT_kwDOTPuhps6hl8My` (plan L547) | Rev 3's §10.1 substitute harness does not satisfy Ship's actual execution gate; Step 2 invokes `harness-architect` for every task lacking `harness-ready`, and that skill requires Go `_test.go` harnesses plus `go vet`/`go test` red-phase evidence. Prose labels leave 017-S unable to pass P-002/P-004 without out-of-scope Go scaffolding. | **§10.1 deviation deleted.** New §5.0 defines the H0 harness contract: a real Go regression harness driving the bash gate, produced by `harness-architect` at Ship Step 2, with a not-implemented shell stub as the structural stub and eight per-task functions as `build-feature` boundaries. No policy amended, no test-first requirement weakened, no fake scaffolding. §5.5 (A2) retargeted to driver-only, since a task whose content is "create the harness" is unreachable in Ship's flow. Governing-workflow amendment explicitly rejected and recorded in §11 Out of scope. |
| `PRRT_kwDOTPuhps6hl8Nd` (`018.011-T:26`) | The task's AC contradicted §7.2 C2.2's fixture-backed mutation proof by mandating a dirty-tree reintroduction edit; C2 numbering also diverged, falsifying the "all 8 task contracts match" claim. | **§7.2 rewritten as the single C2 verification contract**, fixture-backed with an explicit no-dirty-tree-edit rule and a `git status --porcelain` empty check, renumbered AC-C2.1…AC-C2.7. `018.011-T` now carries that contract **verbatim**. The dirty-tree form is explicitly rejected in the criterion text so it cannot be reintroduced by paraphrase. AC-A2.4's non-vacuity proof was aligned to the same rule (temporary fixture copy under `t.TempDir()`). |

**Residual accepted at rev 4**: R10 (`go test` now depends on `bash`) and R11 (harness as an
additional surface), both recorded in §11.

**Found by the rev-4 verification sweep, not by review** — a third instance of the same
plan↔task same-contract defect class: §6.1's B1 acceptance criteria still carried the three-criterion
rev-2 form while `018.007-T` carried the four-criterion rev-3 form covering the preamble surface,
and §6.3 already referenced the then-nonexistent **AC-B1.4**. §6.1 is corrected to the task's form.
An AC-ID parity sweep across all eight tasks now shows an exact bijection (38 plan IDs ↔ 38 task
IDs, no orphan on either side), so the PR's "all 8 task contracts match" claim is true as stated.

**Rev 5 — PR #54 current-HEAD review, remediation cycle 4.** Operator-authorized fourth cycle
beyond the three-cycle cap, scope limited to thread `PRRT_kwDOTPuhps6hnPkD` (comment 3992775719,
plan L720). No other scope reopened.

| Thread | Finding | Resolution |
|---|---|---|
| `PRRT_kwDOTPuhps6hnPkD` (plan L720) | R5 contradicted §1's guarantee and the feature DoD. Because `ci.yml` is itself generated, a future render could both restore the scanned direct-push wording **and** remove the job-`lint` gate step, leaving CI green with the detector never invoked. The plan had to either add a tracked enforcement path surviving regeneration, or narrow the objective and admit silent disablement. | **Enforcement path added; objective NOT narrowed.** The tracked, regeneration-resistant path already existed in the release unit but was never claimed: the §5.0 Go harness is non-generated, and it is invoked by the **generated-baseline** `go test` step in job `expensive` — which a render *re-emits* rather than removes. New §5.0 subsection "Regeneration-resistant enforcement" proves the four-link chain (detector, assertion, invocation, trigger) and shows the trigger link is airtight: the same edit that removes the job-`lint` step is a `.github/**` change, so `changes.code == 'true'` and the harness job cannot skip. **AC-C1.6** makes the step-presence assertion YAML-structural (a text grep would self-match C1's own ledger comment) and proves its non-vacuity by execution against a `t.TempDir()` copy. **AC-C2.8** records that the detector still runs from `go test` with the C1 step gone. R5 reclassified accepted→mitigated; **AC-C1.5** disclosure reworded from "accepted residual" to "detected, not accepted"; ledger wording in §7.1 updated to match. |

**Honest residual, deliberately not fabricated (rev 5).** The guarantee established is exactly
§1's: *regeneration cannot **silently** reopen the gap*. It is **not** a claim that the control
cannot be removed at all — deleting the harness in a reviewable PR diff still disables it, recorded
as **R12**. `CODEOWNERS` is explicitly **not** claimed as mitigation, because this repo's
`CODEOWNERS` header states it has no enforcement effect until code-owner review is required on
`main`, which is deliberately not enabled. Two rejected alternatives are recorded in §5.0:
`secret-scan-history.yml` (non-generated and would survive, but declares itself NOT a PR-required
check by design) and `ci-topology-check.sh` (itself generated, same boundary).

**Scope discipline**: no new file, workflow, CI step, ledger entry, or CODEOWNERS line was added —
only two acceptance criteria (AC-C1.6, AC-C2.8), one reworded (AC-C1.5), and strengthened assertions
in a harness H0 already produces. AC-ID parity re-swept: **40 plan IDs ↔ 40 task IDs**, exact
bijection maintained. Shipment 017-S remains one shipment with unchanged membership and unchanged
dependency order (A1→A2→A3→{B1,B2,B3}→C1→C2).

**Rev 6 — PR #54 current-HEAD review, remediation cycle 5.** Operator-authorized fifth cycle,
scope limited to thread `PRRT_kwDOTPuhps6hqWDe` (comment 3993998429, plan L517). No other scope
reopened; the separate thread `PRRT_kwDOTPuhps6hqWD8` is an Orchestrator-owned incident-record
correction and is explicitly **not** touched here.

| Thread | Finding | Resolution |
|---|---|---|
| `PRRT_kwDOTPuhps6hqWDe` (plan L517) | Revisions 1–5 corrected what Step 1.5 **authorizes** but never its **discovery** logic. Step 1.5 inspects only `.backlogit/` dirtiness and `git log origin/main..main`, not the checked-out branch/HEAD. Once B2/B3 put Stage's commit on `chore/stage-{chore-slug}` and leave the worktree there, both inputs read "synchronized"; the gate skips its commit/push/PR step and then fails the `origin/main` manifest check — and for a shipment-less run there is no manifest to check at all. The **no-shipment route documented in §6.2 was unexecutable as written**. | **B1 expanded, no task added.** §6.1 restructured into §6.1.1 (authorization, rev 1–5) and §6.1.2 (discovery/verification, rev 6) with four corrections: (1) Step 1 requires an explicit **staging handback record** (`stage_branch`, `stage_head_commit`, `stage_outcome`, `stage_artifact_paths`) with a **total** set of deterministic defaults, so the degraded path never halts and never substitutes the default branch; (2) a **branch-resolution block above step 1's dirty/clean split** carries `rev-parse --verify` and the HEAD-equality assertion (P-016), and step 1's pathspec widens to the Stage artifact paths; (3) step 2 compares `origin/main..HEAD` against the already-resolved branch; (4) step 4 gains a `no-shipment` arm verifying `git merge-base --is-ancestor` plus per-path `git show origin/main:{path}` over a **non-empty** path set — a concrete target naming no `shipment_id`. Rev 2's "no no-shipment clause here" position is **explicitly withdrawn** with its inconsistency against §6.2's own route table recorded. B3 gains the producing half (Step 6 emission, AC-B3.5). Five new B1 criteria (AC-B1.5–B1.9) and one new B3 criterion. |

**Four defects found by the rev-6 internal review, before commit.** The first draft of these
corrections was reviewed against the finding and four high-confidence problems were fixed rather
than shipped:

1. **The degraded handback hard-halted on the population it claimed not to wedge.** Halting when
   the `HEAD` resolution yielded the default branch would have blocked every pre-B3 Stage
   invocation — the exact deadlock the paragraph three lines below it said it had avoided — and it
   made 3a's create-arm dead prescription, since the halt sat upstream of it. Now the branch is
   left **UNRESOLVED** and the run routes to 3a, restoring pre-change behaviour.
2. **`stage_outcome` and `stage_artifact_paths` had no degraded default**, leaving step 4's arm
   selection undefined; and an empty path set made the `no-shipment` arm **vacuous** — a zero-run
   `show` loop plus a trivially-true ancestry check would have passed the gate having verified
   nothing. Defaults are now total, and an empty path set is an explicit `STAGING_GATE_FAIL`.
3. **The new guards were unreachable on the dirty path.** Step 1 routes a dirty tree straight to
   step 3, so guards placed in step 2 were skipped on the one path that mutates the repository,
   falsifying AC-B1.9 as written. The resolution block was hoisted **above** step 1.
4. **Step 4's prescribed text could not satisfy its own AC-B1.4.** An outer lead sentence replaced
   the pre-change one and the original reappeared lowercased. The arm label is now a separate line
   above the byte-verbatim original, and AC-B1.4 states that no outer lead may replace or reword it.

Recorded because a review-fix cycle that silently repairs its own first draft teaches nothing; the
defect classes here (a halt contradicting its own rationale, a vacuous verification loop, a guard
placed on the unreached branch of a fork) are the ones worth recognizing again.

**Self-match hazard caught during rev 6, not by review.** The natural phrasing of correction 3 is a
prohibition quoting the literal `origin/main..main`. In `_orchestrator.agent.md` — which **is** in
the scanned corpus — that literal would match the very absence assertion that rejects the
regression, reproducing the failure documented in
`docs/compound/2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md` and already
guarded at §5.0/AC-C2.3. The prohibition is therefore phrased **descriptively**, and AC-B1.6 makes
that constraint contractual rather than incidental.

**AC-B1.4 narrowed, deliberately and on the record.** The criterion previously demanded that step
4's whole block be byte-identical. That is unsatisfiable once step 4 must carry the `no-shipment`
arm — and an unmodified block is exactly what leaves the no-shipment path unverified. AC-B1.4 is
narrowed to what it always protected: the shipment arm's lead sentence, `git show` command, both
bullets, and `STAGING_GATE_FAIL` string are preserved verbatim with unchanged semantics; the step
may gain arm labels and the second arm, and nothing else. The narrowing is recorded here rather
than applied silently, because a weakened criterion with no stated reason is indistinguishable from
an abandoned one.

**Detector deliberately not widened (rev 6).** Main-only discovery is neither §5.3 violation
construct, so catching it in the bash detector would mean opening the **closed construct set**,
adding fixtures, and re-deriving A1's 14/9/5 counts through A1→A2→A3→C2 — materially more than the
finding requires. It is rejected instead by the **existing** B1 harness function, already scoped to
`_orchestrator.agent.md` and already B1's red→green boundary. `directpush-accept-read-and-pr-forms.md`
keeps `git log origin/main..main` as an **accept** case unchanged: it proves the detector does not
flag read verbs, a property independent of what the Orchestrator uses. Recorded at §5.4 and §11 R13.

**Scope discipline (rev 6)**: no new task, file, fixture, manifest entry, harness function, CI step,
ledger entry, CODEOWNERS line, policy ID, or dependency edge. Six new acceptance criteria
(AC-B1.5–B1.9, AC-B3.5), one narrowed with rationale (AC-B1.4), two extended (AC-B1.2, AC-C2.4),
and one new risk row (R13). AC-ID parity re-swept: **46 plan IDs ↔ 46 task IDs**, exact bijection
maintained. Shipment **017-S remains one shipment** with unchanged membership (9 items) and
unchanged dependency order (A1→A2→A3→{B1,B2,B3}→C1→C2). B1 `018.007-T` is re-sized S→M and B3
`018.009-T` XS→S; both remain inside the 2-hour rule (single file, single domain, fully-specified
replacement text).

**Rev 7 — PR #54 current-HEAD review, remediation cycle 6.** Operator-authorized sixth cycle, scope
limited to threads `PRRT_kwDOTPuhps6hrC0c` (comment 3994280334, plan §6.1.2) and
`PRRT_kwDOTPuhps6hrC0o` (comment 3994280351, session memory §3).

| Thread | Finding | Resolution |
|---|---|---|
| `PRRT_kwDOTPuhps6hrC0c` | Rev 6's no-shipment arm iterated `git show origin/main:{path}` over a path set whose documented default was the **directory list** `.backlogit/ docs/plans/ docs/decisions/ docs/memory/`. `git show` exits **0** on a tree and those directories exist on `origin/main` in every repository, so the loop passed having read **no file the Stage run wrote**; paired with the degraded `stage_head_commit` default the ancestry check was trivially true too — a **false pass**. | `stage_artifact_paths` redefined as a **non-empty set of concrete repository-relative file paths**, derived from the commit when not supplied, validated for file-ness / root-containment / set-equality when supplied, and verified with `git cat-file -t` = `blob` rather than `git show`. `STAGE_ARTIFACT_ROOTS` split out as a fixed constant so step 1's pathspec cannot be shrunk by a handback. 3a re-records `stage_head_commit` after committing. AC-B1.10 and AC-B1.11 added; AC-B1.5/B1.7/B1.8 strengthened. |
| `PRRT_kwDOTPuhps6hrC0o` | The session-memory artifacts table cited plan **revision 6** and a hardening section **§10a** that no longer exists. | Corrected, and the stale-reference sweep widened file-wide: memory §3 plan and deliberation rows, plan front matter, and plan §5.3's `§10a.1` citation. Recorded in memory §14.5. |

**Rev 8 — PR #54 adversarial review.** Operator-directed adversarial round (anchor **GPT-5.6 Sol**
with **GPT-5.4-mini**, **Claude Sonnet 5**, and **Claude Opus 5**; **no route degradation**),
replacing a further self-directed cycle per the operator's standing instruction recorded at memory
§14.8. All nine findings were classified **same-contract completion** under P-021 — none reopened
the release unit's scope.

| # | Sev | Finding | Resolution |
|---|---|---|---|
| 1 | **P1** | **Aggregate verification missing.** The rev-7 contract derived its verified path set from `stage_head_commit`'s **single-commit** diff, so artifacts introduced by earlier commits on a multi-commit Stage branch were never read on `origin/main` — covered only by ancestry, the very property rev 7 judged insufficient. | `stage_base_commit` introduced, **captured by the Orchestrator immediately before invoking Stage** and retained as authoritative; base→head ancestry asserted; the set derived from the stable **two-tree** `git diff --name-only -z --diff-filter=d {base} {head} -- {STAGE_ARTIFACT_ROOTS}` (not the range form, empty by construction post-merge). Arm grows 5→**6** ordered checks. 3a **preserves** the base and recomputes the aggregate. Propagated through §6.1.1, §6.1.2, AC-B1.5/B1.8/B1.10/B1.11, §6.3/AC-B3.5, §7.2/AC-C2.4, §8, §5.0 map, §11 R14/R16, 018-F DoD, `018.007-T`, `018.009-T`, `018.011-T`. |
| 2 | **P1** | **`no-shipment` unreachable.** The installed Stage Step 5.5 (shipment assembly "MANDATORY") and Step 6 pre-summary gate both **halt** without a `shipment_id`, so no producer could ever emit `stage_outcome: no-shipment` and B1's entire second arm was dead prescription. | B3 gains **AC-B3.6**: Step 5.5's mandatory rule is scoped to a harvest that produced items, a reviewed **valid empty** harvest terminates as `no-shipment`, Step 6 requires `shipment_id` **only** for `stage_outcome: shipment`, the complete handback is emitted, and the run terminates without routing to Ship. **P-003 failures, harvest failures, and a missing required shipment after a non-empty harvest still HALT** and are never re-labelled — recorded as **R15**. `018.009-T` re-sized S→M. |
| 3 | **P1** | `018.007-T`'s description still carried rev-6 executable prose (root-default path set, `git show` loop) as operative-looking instructions, with rev 7 layered on as an addendum — two contradictory algorithms in one task body. | Description **rewritten** to a single current authoritative contract; all revision archaeology moved to implementation notes / review history. |
| 4 | **P1** | Plan §6.2 and `018.008-T` carried **two divergent five-criterion** B2 lists — symmetry in one, the P-016/P-001 clarification in the other, with AC-B2.3/B2.4/B2.5 bound to **different** criteria on each surface. | Reconciled to one identical **six**-criterion contract (AC-B2.1–AC-B2.6) carried verbatim on both surfaces; exact parity re-run. |
| 5 | P2 | Brittle line-ordinal locators (`workflow-policies.md` **L238**) for the Ship prohibition in §2, §5.3, §7.2/AC-C2.3, `018.011-T`, and the deliberation's §1.1 — the last still presented as **current evidence** despite §8.5 promising the switch to quoted anchors. | Replaced with the structural locator: P-010's **Ship MUST NOT** bullet "Commit or push directly to `main`". Deliberation §1.1's `L230`/`L258` anchors converted likewise; §8.5's promise is now kept rather than merely recorded. |
| 6 | P3 | Generated-baseline evidence cited "the `go test` step in job `expensive`", which names no step in this workflow. | Cites job `expensive` (`name: test`) step **`Test (race)`**, command `go test -race -mod=readonly ./...`, and states explicitly that AC-C2.7's constitutional `go test ./...` is **separate and still required**. Per-task `harness_cmd` values stay narrow and unchanged. |
| 7 | P3 | AC-A1.2 named the fixture prefixes `reject-*` / `accept-*`; the actual files are `directpush-reject-*` / `directpush-accept-*`. | Corrected on both surfaces. |
| 8 | P3 | AC-A2.1 demanded the **bare** scan exit non-zero, but A2's stub classifies every construct `accept`, so a bare scan correctly exits **0** — the criterion was unsatisfiable by a correct A2. | Non-zero scoped to **`--self-test` only**; bare-mode red is reserved for **AC-A3.3**, after a real detector exists. |
| 9 | P3 | Stale references: Constitution Check row XI cited AC-C2.5 (markdownlint) instead of AC-C2.6; §12's rev-7 row promised "(see below)" with no rev-7 block; the deliberation header claimed the planning branch had "no push, no PR". | All corrected; the rev-7 record block above is added, and the deliberation header now names its five addenda and the PR the branch actually carries. |

**Scope discipline (rev 8)**: no new task, file, fixture, manifest entry, harness function, CI step,
ledger entry, CODEOWNERS line, policy ID, or dependency edge. Two new acceptance criteria
(**AC-B2.6**, **AC-B3.6**), six strengthened (AC-B1.5, AC-B1.8, AC-B1.10, AC-B1.11, AC-B3.5,
AC-C2.4), four corrected in place (AC-A1.2, AC-A2.1, AC-C2.3, AC-C2.8), one renumbered set
(AC-B2.1–AC-B2.5 → the reconciled AC-B2.1–AC-B2.6), and two new risk rows (**R15**, **R16**).
Shipment **017-S remains one shipment** with unchanged membership (9 items) and unchanged
dependency order (A1→A2→A3→{B1,B2,B3}→C1→C2). B3 `018.009-T` is re-sized S→M for the two
additional `_stage.agent.md` edit sites; it remains inside the 2-hour rule (single file, single
domain, fully-specified replacement text). No production code, test, script, workflow, policy, or
agent file is modified by this pass — those remain Ship's execution surfaces.

**Rev 9 — PR #54 current-HEAD review, remediation cycle 7.** One **visible** Copilot finding
(thread `PRRT_kwDOTPuhps6hrbAk`, comment `3994429008`, on `018.009-T:22`), classified
**same-contract completion** under P-021. Scope limited to that finding; no other scope reopened.

| Thread | Finding | Resolution |
|---|---|---|
| `PRRT_kwDOTPuhps6hrbAk` | `018.009-T` updates Stage's **authorization** and declares a **handback**, but adds no operative Stage step to create/check out the dedicated branch or commit the artifacts. The installed `_stage.agent.md` has only planning, harvest, shipment, archive, and summary steps — no branch or commit operation anywhere — so implementing AC-B3.1–AC-B3.6 as written can still leave Stage on the default branch, making the reported `stage_head_commit` and concrete-path contract unverifiable. | **Two operative steps added to the file B3 already owns** (§6.3.1): **Step 1.9**, a pre-mutation branch gate between the learnings-retrieval and deliberation steps, which derives a `{scope-slug}`, records `stage_base_commit`, **verifies** an Orchestrator-supplied branch or **creates/checks out** `chore/stage-{scope-slug}`, refuses to write while `HEAD` is the default branch, and switches the single worktree (P-016); and **Step 5.7**, a post-mutation artifact commit between stash archival and the summary, which asserts `HEAD`, stages only the four `STAGE_ARTIFACT_ROOTS`, commits conventionally, sets `stage_head_commit` from the resulting `HEAD`, and neither pushes nor touches a PR. Both are registered in the Step Sequence Contract checklist and enforced by Step 6's pre-summary gate. **AC-B3.7** added, constraining **order** as well as presence; §6.3.2 specifies the structural (index-comparison) detection inside the existing `TestDirectPushGate_StageAgentSurfaceClean`. |

**The naming dependency the finding exposed second-order.** Rev 1–8 named the branch
`chore/stage-{shipment_id}` with `chore/stage-{chore-slug}` as a shipment-less fallback. Once the
branch must exist *before* the first artifact write, that primary form is **underivable**: Step 5.5
creates the shipment as the session's last mutation. The rev-8 wording therefore pushed an executor
toward deferring the branch past the very writes it protects — a second route back to the
`fdff9e4` shape. Rev 9 makes the form `chore/stage-{scope-slug}`, keeps a shipment ID as a
**permitted** slug rather than deleting it, and corrects §6.2's grant and route table, §6.3's Git
row, AC-B2.2, and AC-B3.2 together. `chore/stage-pipeline-policy-gap` — the branch carrying this
plan — is a conforming example.

**Producer/consumer failure class, fourth instance.** Rev 1 obliged Stage to merge a PR it could
not create; rev 6 obliged the Orchestrator to find a branch nothing announced; rev 8 keyed a
verification on an outcome nothing could emit; rev 9 finds a **reported value with no producing
action**. The discipline deliberation §12.2 wrote down — *for every value a contract consumes, name
the surface that emits it, and confirm that surface is permitted to* — was followed for
permission and emission and **not** for execution. It is extended in deliberation §13: confirm the
producing surface is not merely permitted but **instructed**, and that the instruction is **ordered**
relative to the state it describes.

**Detector deliberately not widened (rev 9).** A missing operative step is neither §5.3 construct
— it is not a push to the default branch and not a permission to commit on one — so the closed
construct set and the 14-fixture corpus stay exactly as they are, the §10.4 / §11.4 / §12.4
position unchanged. Rejection is carried by the **existing** B3 harness function, already scoped to
`_stage.agent.md` and already B3's red→green boundary. The prescribed replacement text was checked
against the detector rather than assumed safe (§6.3.1, detector-neutrality).

**Deliberately not swept (rev 9)**: the `chore/stage-{shipment_id}` form still appears in this
plan's **historical revision narratives** (the Revision 6 paragraph, §6.1.1's quotation of the 3e
text being deleted, §12's rev-6 record), in the deliberation's archaeology, and in `018.007-T`'s
rev-6 reconciliation note. Those are accurate records of what earlier revisions said and of the
text B1 removes; no acceptance criterion depends on them. Every **operative** occurrence — the
P-010 grant, the route table, the Role Boundary cell, 3a's create-arm, AC-B2.2, AC-B3.2 — is
corrected.

**One defect found by internal review before commit (rev 9).** The first draft placed Step 1.9 at
a *floating* position — "earliest point a slug is derivable, latest before the first write,
whichever comes first" — with a provisional-slug escape hatch for an early Step 1 duplicate
archival. That is **self-contradictory** against the very file it edits: `_stage.agent.md`'s Step
Sequence Contract states that a session MUST execute its steps **in order**, so a step registered
between Step 1.8 and Step 2 cannot also be specified to run during Step 1. Worse, the floating form
made the position unassertable — AC-B3.7 would have enforced the ordinal while the prose licensed an
earlier run — and it left a tracked `.backlogit/` write legal on the default branch. The fix
inverts it: the ordinal is **fixed**, and the two mutations that would otherwise precede it are
**deferred** past it by an explicit rule. Recorded rather than silently repaired, per the rev-6
precedent: the defect class — *a guard whose stated position contradicts the ordering semantics of
the contract it is inserted into* — is worth recognizing again.

**Scope discipline (rev 9)**: no new task, file, fixture, manifest entry, harness function, CI
step, ledger entry, CODEOWNERS line, policy ID, shipment, or dependency edge. One new acceptance
criterion (**AC-B3.7**), three strengthened (AC-B2.2, AC-B3.2, AC-B3.5), two re-scoped
(AC-B3.3, AC-B3.6 — their "only changes outside the Role Boundary table" clause now admits
AC-B3.7), one extended (AC-C2.4), and two new risk rows (**R17**, **R18**). AC-ID parity re-swept:
**51 plan IDs ↔ 51 task IDs**, exact bijection maintained. Shipment **017-S remains one shipment**
with unchanged membership (9 items) and unchanged dependency order
(A1→A2→A3→{B1,B2,B3}→C1→C2). B3 `018.009-T` is re-sized **M→L** for the four additional
`_stage.agent.md` edit sites; it remains inside the 2-hour rule (single file, single domain, fully
specified replacement text). No production code, test, script, workflow, policy, or agent file is
modified by this pass.

**Rev 10 — PR #54 second adversarial review.** Operator-directed second adversarial round (anchor
**GPT-5.6 Sol** with **GPT-5.4-mini**, **Claude Sonnet 5**, and **Claude Opus 5**; **no route
degradation**). Four **visible** Copilot threads plus one adversarial-only finding, all classified
**same-contract completion** under P-021. Scope limited to those five findings; no other scope
reopened.

| # | Sev | Thread / comment | Finding | Resolution |
|---|---|---|---|---|
| 1 | **P1** | `PRRT_kwDOTPuhps6hsPRZ` / `3994747640` | **The handback outcome pair failed open.** `stage_outcome` was derived *unconditionally* from whether a `shipment_id` was returned, with no `absent` branch — so a reported `shipment` outcome whose `shipment_id` did not reach the Orchestrator was silently rewritten to `no-shipment`, the shipment arm was skipped, the no-shipment arm passed on artifacts that really were on `origin/main`, the gate reported success, and Ship was never routed. A formed shipment, silently abandoned. | §6.1.2 **correction 6**: the derivation is scoped to the **absent** case and a reported value is preserved; a **fail-closed pair validation** is added to step 4 *before* arm selection (unknown outcome halts; `shipment` without a non-empty `shipment_id` halts; `no-shipment` with one halts). Placed in **verification**, not defaulting, so AC-B1.5's never-halts resolution invariant is intact. **AC-B1.12** added, AC-B1.5 strengthened; producer side carried by AC-B3.5 (explicit, consistent pair); coupling by AC-C2.4 rev-10 (a); DoD bullet added. |
| 2 | **P1** | `PRRT_kwDOTPuhps6hsPRq` / `3994747663` and `PRRT_kwDOTPuhps6hsPR5` / `3994747682` | **The pre-gate deferral rule was under-inclusive.** Rev 9 *enumerated* two deferred mutations. The installed Stage contract also schedules, before Step 1.9's ordinal, a **stash-classification** checkpoint, a **contextual-grouping / operator-selection** checkpoint, and the **write half** of the mandatory late-identifier reconciliation — all tracked, all therefore still legal on the default branch under the rev-9 wording. | §6.3.1 correction 1 rewritten: the rule is **categorical** — Step 1.9 precedes **every** tracked Stage artifact mutation — with the three classes enumerated as illustration, not as a closed list. Read-only classification and grouping still run first (they derive the slug). Gitignored disposables (index sync, `.backlogit/checkpoints/`, hook queue, `.backlogit/runtime/`) are expressly **excluded**, not overclaimed. Lifted into **AC-B3.7**, §6.3.2 assertion 3, `018.009-T`, the DoD, and AC-C2.4 rev-10 (b); R18 updated. |
| 3 | **P1** | `PRRT_kwDOTPuhps6hsPSJ` / `3994747702` | **Out-of-root detection was unimplementable.** Step 5.7 required a halt on any change outside `STAGE_ARTIFACT_ROOTS` but prescribed only a root-scoped `git add`, which **ignores** such a change rather than reporting it. The obligation had no mechanism. | Step 5.7 gains an explicit **detection sub-step before staging**: `git status --porcelain=v1 -z --untracked-files=all`, NUL-record parsing (never line splitting), both rename/copy endpoints treated as affected paths, halt with a P-010 signal on any out-of-root path. `--untracked-files=all` and `-z` are justified as load-bearing with **verified** output shape. Pre-existing unrelated dirt is **halt-and-leave-alone** — no checkout, restore, stash, revert, or delete (Constitution VII). Lifted into **AC-B3.7** and §6.3.2 assertions 3–4 (including the detection-before-staging **intra-section ordering** check); **R20** added. |
| 4 | **P2** | — (AC-A2.5) | **AC-A2.5 could not prove its own claim.** One driver-scoped `-run` invocation cannot establish that a different test is red. | Split into **two exactly-anchored invocations**: the driver test exits 0; `TestDirectPushGate_SelfTestPasses` exits non-zero **and** prints `--- FAIL:` plus the expected **detector-stub** reason, with compile errors, absent `bash`, missing fixtures, and a mis-typed selector rejected by name. The `--- FAIL:` guard is required because — **verified** — a non-matching `-run` pattern exits **0** with `[no tests to run]`. Bare-mode red stays **AC-A3.3**'s; P-004 is not weakened. §5.0 map and `018.005-T` synchronized. |
| 5 | **P1** | adversarial-only (low confidence, directly evidenced) | **The harness lifecycle deadlocked against Ship's full-suite gate.** §5.0 declared all eight per-task functions red after H0 and left them red until their task landed, while `_ship.agent.md` Step 4.3 runs `go test ./...` after **every** task. After A1 went green the seven future functions were still red, Step 4.3 failed, and no task could ever complete. | New **§5.0.1**: one selector flag (`-gatetask`, dot-free and shell-neutral — **verified**, a dotted name is split by PowerShell's tokenizer) plus one eight-row activation table. Targeted mode forces exactly one task's assertions (each `harness_cmd` carries it, so every task keeps an observed red→green transition); default mode runs surfaces that exist plus the **activation-independent A1 anchor**, which keeps H0's `go test ./...` red literal. Four mutation-proof checks (MP0 table integrity, MP1 executed predicate non-vacuity, MP2 terminal completeness inside C2, MP3 visible skips + strict selector) close the skip-forever degeneration; **R19** records the residual. Four alternatives rejected on the record, including amending Ship Step 4.3 (out of scope, generated file, weakens test-first for every future shipment). |

**Harness-ready semantics kept honest (rev 10).** `harness-ready` continues to mean exactly what
`harness-architect` verifies — compilation clean, default suite red — and §5.0 now says so
explicitly instead of letting the label imply that all eight assertions were simultaneously red.
The per-task red is a **per-task** obligation discharged by Ship at claim time from that task's
`harness_cmd`, which is the evidence `build-feature` already records. Nothing is asserted that no
actor produces, and no label is applied from prose — the rev-4 lesson, held.

**The producer/consumer failure class, fifth instance — and a new one alongside it.** Rev 10's
finding 1 is the class again, in its most dangerous form yet: the consumer did not merely fail to
find a value, it **overwrote a correct one with a derived default whose fallback was a success
terminal**. The extended discipline gains a sixth clause: *a default must never be able to
overwrite a reported value, and a derivation whose fallback is a success terminal must be
validated before it is acted on.* Findings 2, 3 and 5 are a **different** class — an obligation
written without a mechanism that can discharge it (an enumeration mistaken for a rule, a halt with
no detector, a red phase that cannot coexist with the gate that consumes it). Its discipline: *for
every MUST, name the operation that performs it and confirm that operation can observe what the
MUST is about.*

**One latent defect found by internal review before commit (rev 10).** Requirement (d) —
*the final `go test ./...` runs all eight real assertions green* — forced a check revisions 4–9
never made: is every function's assertion still true at the **end state**? Seven were. **A2's was
not.** Its map row asserted that `--self-test` reports every reject fixture failing **against the
all-accept stub** with a non-zero exit — true only while A2's stub is current, **false the moment
A3 installs the real detector**. The function would have gone red at A3 and stayed red, wedging
`go test ./...` for the rest of the shipment. Activation did not cause this; it made it visible,
because "all eight green at the end" had never been stated as a requirement before. The fix asserts
the **durable** driver contract A2 actually delivers and leaves the stub-specific observations in
**AC-A2.1**, read from the script's own output at A2's boundary — the criterion unchanged and
unweakened, only its carrier named. Recorded rather than silently repaired, per the rev-6 and rev-9
precedent: the defect class — *a regression assertion pinned to a transitional state* — is worth
recognizing again.

**Scope discipline (rev 10)**: no new task, file, fixture, manifest entry, harness function, CI
step, ledger entry, CODEOWNERS line, policy ID, shipment, or dependency edge. One new acceptance
criterion (**AC-B1.12**), four strengthened (AC-A2.5, AC-B1.5, AC-B3.5, AC-B3.7), one extended
(AC-C2.4), one new plan subsection (**§5.0.1**), one new B1 correction (**correction 6**), and two
new risk rows (**R19**, **R20**). All eight `harness_cmd` values gain the `-gatetask={ID}` selector;
no harness **function** is added or renamed. AC-ID parity re-swept: **52 plan IDs ↔ 52 task IDs**,
exact bijection maintained. Shipment **017-S remains one shipment** with unchanged membership
(9 items) and unchanged dependency order (A1→A2→A3→{B1,B2,B3}→C1→C2). No task is re-sized: rev 10
reworks clauses inside step sections rev 9 already specified rather than adding edit sites. No
production code, test, script, workflow, policy, or agent file is modified by this pass — those
remain Ship's execution surfaces.

**Deliberately not claimed (rev 10).** Stage does **not** push, open, update, comment on, or merge
PR #54, and does **not** reply to or resolve threads `PRRT_kwDOTPuhps6hsPRZ`,
`PRRT_kwDOTPuhps6hsPRq`, `PRRT_kwDOTPuhps6hsPR5`, or `PRRT_kwDOTPuhps6hsPSJ`. Those are Orchestrator
or operator actions; this record states only what the artifacts now contain.

<!-- plan-review-attempt: 2 -->

---
title: "Implementation Plan — Stage Artifact Branch/PR Policy Gap Correction"
date: 2026-09-11
revision: 7
status: reviewed
agent: Stage
governs: stash 638A410B
source: docs/decisions/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-deliberation.md
---

# Implementation Plan — Stage Artifact Branch/PR Policy Gap Correction

**Source document**:
`docs/decisions/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-deliberation.md`
(including its **§8 Revision 2 addendum**, **§9 Revision 3 addendum**, **§10 Revision 6
addendum**, and **§11 Revision 7 addendum**)
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
boundary — the §5.0 Go harness, invoked by the **generated-baseline** `go test` step rather than by
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

Ship's mirror prohibition already exists at `workflow-policies.md` **L238**
("Commit or push directly to `main`"). The fix makes Stage symmetric with it.

All three files carry `Generated by autoharness | Template: …` provenance — prose-only correction
is reversible by a future render with no signal.

## 3. Surface

| File | Change | Task |
|---|---|---|
| `tests/integration/directpush_gate_test.go` | **New.** Go regression harness driving the gate; 8 test functions, one per task. Produced by `harness-architect` at **Ship Step 2**, not by a task (§5.0). | H0 |
| `scripts/testdata/directpush/*.md` + `directpush-manifest.json` | **New.** 14 fixtures + sorted verdict manifest. | A1 |
| `scripts/check-direct-push-language.sh` | **New.** Not-implemented stub (H0) → driver (A2) → detector (A3). | H0, A2, A3 |
| `.github/agents/_orchestrator.agent.md` | Reword Step 1.5 preamble; delete 3e; fold its branch-point into 3a. **Rev 6**: require the Stage handback record in Step 1; make Step 1.5 discovery branch-specific; widen the dirtiness pathspec; add step 4's no-shipment arm. **Rev 7**: make the no-shipment arm verify concrete files derived from `stage_head_commit` (`cat-file -t` = `blob`), pin step 1's pathspec to the fixed artifact-root constant, and re-record the handback commit in 3a. | B1 |
| `.github/policies/workflow-policies.md` | P-010 Stage bullets + Amendment Log **1.25.0**. | B2 |
| `.github/agents/_stage.agent.md` | Role Boundary Git row + PR row note. **Rev 6**: emit the staging handback record in the Step 6 summary contract. **Rev 7**: emit `stage_artifact_paths` as concrete file paths derived from `stage_head_commit`. | B3 |
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
   `repoRoot(t)` helper convention. **Eight** test functions, one per task (table below).
2. `scripts/check-direct-push-language.sh` as a **structural stub**: correct shebang,
   `set -euo pipefail`, argument dispatch present, and every mode exiting non-zero after printing
   the marker `not implemented: direct-push detector`. This is the shell analogue of
   `harness-architect`'s `panic("not implemented: <reason>")` production stub — the module
   "compiles" (the script parses and runs) while every test fails for the intended reason.

**Red-phase evidence (P-004 precondition, taken literally):** after H0, `go vet ./...` exits 0 and
`go test ./...` exits non-zero with all eight functions failing on the `not implemented` marker or
on a missing fixture corpus. `Compilation: PASS`, `Red Phase: CONFIRMED`. Only then does
`harness-architect` apply `harness-ready`. **Stage does not apply that label and this plan does not
ask Ship to.** `018-F.custom_fields.harness_status` stays `pending` until H0 runs.

**Tool resolution — fail, never skip.** The harness resolves `bash` on `PATH` and **fails** the
test when it is absent; it MUST NOT `t.Skip`. A skip would make the red phase unobservable and
silently satisfy P-004 (P-012: record `TOOL_DEGRADED`, never silently skip). This diverges
deliberately from `build_script_test.go`'s `t.Skip("pwsh not available on PATH")`, because `pwsh`
is optional tooling there whereas `bash` is a hard prerequisite of every gate script in this repo.

**Per-task harness map.** Each function is the `harness_cmd` boundary `build-feature` loops on, so
every task has exactly one observable red→green transition:

| Task | Test function | Turns green when |
|---|---|---|
| A1 `018.004-T` | `TestDirectPushGate_FixtureCorpusIsComplete` | 14 fixtures + manifest exist; manifest keys sorted; bijection with the directory holds; 9 reject / 5 accept |
| A2 `018.005-T` | `TestDirectPushGate_SelfTestDriverEnumeratesFixtures` | `--self-test` names every manifest fixture and reports every **reject** fixture failing against the stub; exit stays non-zero |
| A3 `018.006-T` | `TestDirectPushGate_SelfTestPasses` | `--self-test` exits 0: actual == manifest verdict for all 14, plus both default-branch-resolution assertions |
| B1 `018.007-T` | `TestDirectPushGate_OrchestratorSurfaceClean` | scan scoped to `_orchestrator.agent.md` reports 0 findings **and** (rev 6) Step 1.5's discovery is branch-specific: Step 1 requires the handback fields, the unpushed-commit check resolves the reported branch and compares `origin/main..HEAD`, **no default-branch self-range literal remains in Step 1.5**, and step 4 carries both outcome arms. **Rev 7**: the no-shipment arm derives concrete files from `stage_head_commit` and verifies them with `cat-file -t` = `blob`; **no directory-prefix path set and no `git show`-over-`{path}` loop remains in that arm**; step 1's dirtiness pathspec is the fixed artifact-root constant; 3a re-records the handback commit |
| B2 `018.008-T` | `TestDirectPushGate_PolicySurfaceClean` | scan scoped to `workflow-policies.md` reports 0 findings **and** the Amendment Log carries row `1.25.0` |
| B3 `018.009-T` | `TestDirectPushGate_StageAgentSurfaceClean` | scan scoped to `_stage.agent.md` reports 0 findings **and** (rev 6) the Step 6 summary contract emits the staging handback record. **Rev 7**: that contract requires `stage_artifact_paths` to be concrete file paths derived from `stage_head_commit`, not directory prefixes |
| C1 `018.010-T` | `TestDirectPushGate_CIWiringIsBlocking` | job `lint` carries the task-ID-suffixed step with no `continue-on-error` and no toggle; `ci-gate` still transitively needs `lint`; every script invoked by a `ci.yml` step has a `CODEOWNERS` owner line. **Rev 5**: the step-presence check is **YAML-structural** over `jobs.lint.steps[]` and its non-vacuity is proven (AC-C1.6) |
| C2 `018.011-T` | `TestDirectPushGate_FullCorpusClean` | bare full-corpus scan exits 0 with 0 findings and the three §5.3 marker-free constructs score **accept**. **Rev 5**: this runs the detector from `go test`, so it keeps executing even if the job-`lint` step is removed by a render (AC-C2.8) |

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
  ATX heading, is the shape `workflow-policies.md` actually uses for both the Ship L238 bullet and
  the Stage bullet B2 adds; reading "heading" as ATX-only would flag both and make zero-findings
  unreachable), (b) an ATX `MUST NOT` / `Forbidden` heading, or (c) a table cell under a
  `Forbidden` / `MUST NOT` **column header**; **or**
- a prohibition marker **precedes** the construct within the same clause:
  `never`, `must not`, `do not`, `don't`, `no`, `forbidden`, `prohibited`.

`rather than` and `instead of` are **removed** from the marker set — they are directional, not
prohibitive, and would have excused "push directly to `main` rather than opening a PR".
Marker-must-precede closes the "…first; do not create a PR unless rejected" fail-open.

Structural resolution is what makes the real corpus pass: `workflow-policies.md` L238 and
`_ship.agent.md`'s Role Boundary `| Git |` row carry **no in-line marker** — their prohibition
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
asserts filename prefix and manifest verdict **agree**, and asserts **bijection** (every file on
disk is in the manifest and vice versa), closing rev 1's silently-unexercised-fixture hole.

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
- **AC-A1.2** Every manifest verdict agrees with its fixture's filename prefix (`reject-*` /
  `accept-*`), and both classes are non-empty: 9 reject, 5 accept. **Static data check only** —
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

- **AC-A2.1** Every **reject** fixture is reported **failing** against the stub, by name, and the
  bare/`--self-test` exit stays non-zero. This is the **observed red for the detector itself** —
  rev 1 never took the detector red.
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
- **AC-A2.5** `go test ./tests/integration -run TestDirectPushGate_SelfTestDriverEnumeratesFixtures`
  exits 0, and `TestDirectPushGate_SelfTestPasses` (A3's function) is still **red**.

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
>    when none was reported, create `chore/stage-{shipment_id}` (`chore/stage-{chore-slug}` when
>    no shipment was formed) **from the current commit** — and commit any uncommitted backlog and
>    planning files to it. **Then re-record the handback commit** (rev 7 — correction 5, §6.1.2):
>    set `stage_head_commit` to `git rev-parse HEAD` and discard any `stage_artifact_paths` the
>    handback reported, so step 4 derives its file set from the commit that actually carries the
>    artifacts.

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
on, once B2/B3 require Stage to create, check out, and commit to `chore/stage-{shipment_id}` /
`chore/stage-{chore-slug}` and leave the worktree there:

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
>   artifact branch Stage created and committed on), `stage_head_commit`, `stage_outcome`
>   (`shipment` or `no-shipment`), `stage_artifact_paths` (the **concrete repository-relative file
>   paths** Stage wrote and committed in `stage_head_commit` — file paths only, never directory
>   prefixes; rev 7), and — when `stage_outcome` is `shipment` — a `shipment_id` in `queued`
>   status.

> 4. Receive Stage's output: record the staging handback record, and the `shipment_id` when one
>    was returned. Apply these defaults for any field Stage did not report, recording
>    `STAGING_HANDBACK_DEGRADED: Stage did not report {fields} — defaulted` once and surfacing it
>    to the operator:
>    - `stage_branch` / `stage_head_commit`: resolve from the current checkout
>      (`git rev-parse --abbrev-ref HEAD`, `git rev-parse HEAD`). If that branch is the default
>      branch, leave `stage_branch` **UNRESOLVED** — **never** substitute the default branch for
>      the Stage artifact branch.
>    - `stage_outcome`: `shipment` when a `shipment_id` was returned, otherwise `no-shipment`.
>    - `stage_artifact_paths`: leave **empty** here and derive it in step 4 from
>      `stage_head_commit`'s own changed-file set (correction 4, check 3). It is **never**
>      defaulted to a directory list (rev 7).

**Two path notions, deliberately distinct (rev 7).** `STAGE_ARTIFACT_ROOTS` is a **fixed constant**
of the gate — `.backlogit/`, `docs/plans/`, `docs/decisions/`, `docs/memory/` — and is *not* a
handback field. It serves two roles and only two: it is step 1's dirtiness pathspec, and it is the
containment allow-list every concrete path in step 4 must lie under. `stage_artifact_paths` is a
handback field and is always a set of **concrete file paths**. Rev 6 conflated the two, using one
field for both jobs; that conflation is what let directory prefixes reach a verification loop that
had no way to reject them. Directory prefixes are correct in step 1 — it runs **before** the commit
exists, so there is no commit to derive files from — and are rejected in step 4, where a commit
does exist and the files it changed are knowable exactly.

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

> 4. **When `stage_outcome` is `shipment`:**
>    Verify the shipment manifest exists on the remote default branch:
>    `git show origin/main:.backlogit/queue/{shipment_id}.md`
>    - If the file exists: staging artifacts are confirmed on the remote. Proceed to Step 2.
>    - If the file does not exist: halt with `STAGING_GATE_FAIL: shipment manifest {shipment_id}
>      not found on origin/main`.
>
>    **When `stage_outcome` is `no-shipment`:**
>    There is no manifest to name, so verify the **commit** and the **concrete files that commit
>    changed**. Run these five checks in order; any failure halts the gate.
>
>    1. `git rev-parse --verify {stage_head_commit}^{commit}` — halt with
>       `STAGING_GATE_FAIL: stage_head_commit {stage_head_commit} is not a commit in this
>       repository`.
>    2. `git merge-base --is-ancestor {stage_head_commit} origin/main` — halt with
>       `STAGING_GATE_FAIL: stage commit {stage_head_commit} has not reached origin/main`.
>    3. Derive the commit's changed Stage artifact files:
>       `git diff-tree --no-commit-id --name-only -r -z --diff-filter=d {stage_head_commit}`,
>       keeping only paths under `STAGE_ARTIFACT_ROOTS`. Call the result `DERIVED`. If `DERIVED`
>       is **empty**, halt with `STAGING_GATE_FAIL: {stage_head_commit} changed no Stage artifact
>       files under the allowed Stage artifact roots`.
>    4. Resolve `VERIFY_SET`. When the handback reported a **non-empty** `stage_artifact_paths`,
>       every reported path MUST be a repository-relative **file** path — no trailing `/`, not a
>       bare `STAGE_ARTIFACT_ROOTS` entry, no glob or wildcard — and MUST lie under
>       `STAGE_ARTIFACT_ROOTS`; otherwise halt with `STAGING_GATE_FAIL: stage_artifact_paths
>       contains a non-file or out-of-root path: {path}`. The reported set MUST then be **equal**
>       to `DERIVED` (same members, order-insensitive); otherwise halt with
>       `STAGING_GATE_FAIL: stage_artifact_paths does not match the files changed by
>       {stage_head_commit}`. `VERIFY_SET` is that set. When no paths were reported — none were
>       supplied, or step 3a discarded them after creating the artifact commit — set
>       `VERIFY_SET := DERIVED`.
>    5. For **every** path in `VERIFY_SET`: `git cat-file -t "origin/main:{path}"` MUST exit 0
>       **and** print exactly `blob`. Halt with `STAGING_GATE_FAIL: Stage artifact {path} not
>       found on origin/main` on a non-zero exit, and with `STAGING_GATE_FAIL: Stage artifact
>       {path} is a {type}, not a file, on origin/main` on any other printed type.
>
>    When all five pass: the Stage artifacts are confirmed on the remote **as concrete files**.
>    **Stop here — there is no shipment, so there is nothing to route to Ship.**

This is the **concrete artifact verification target that requires no shipment ID**: ancestry proves
the Stage commit reached the default branch, and the per-file type check proves the artifacts
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

1. **Concreteness is structural, not stylistic.** `git diff-tree -r` recurses into subtrees and
   therefore emits **only blob paths** — a directory cannot appear in `DERIVED` by construction,
   so the derived set is concrete whether or not anyone remembers to check it. `-z` NUL-delimits
   the output so paths are never quoted or backslash-escaped, and `--diff-filter=d` excludes
   **deletions**, so a Stage artifact the commit *removed* or renamed away is not demanded to be
   readable on `origin/main` (the `HEAD` of this very branch is a rename of that shape).
2. **`git cat-file -t` replaces `git show` in this arm.** Both exit 0 on a tree, so the exit code
   alone can never discriminate a file from a directory; the discriminator is the **printed object
   type**. Requiring exactly `blob` rejects a directory even if one reached `VERIFY_SET` through
   the reported-path branch, and `cat-file -t` still exits non-zero (128) when the path is absent,
   so presence and file-ness are proven by one command.
3. **Set equality, not subset.** Equality rejects a reported path the commit never touched (an
   invented or copy-pasted path) **and** a reported set that silently drops files — which would
   otherwise let a handback shrink the verified set down to one always-present file. A subset check
   catches only the first.

**Why the set is scoped to `stage_head_commit`'s own changes.** It is the only artifact set
recoverable **after the merge lands**. A range form such as `origin/main..{stage_head_commit}` is
**empty by construction** once that commit is an ancestor of `origin/main` — precisely the state
step 4 runs in — so a range-derived set would collapse to empty and re-create the vacuous-loop
failure this correction exists to remove. A commit's changed-file list, by contrast, is a property
of the commit object and its parent: it reads identically before and after the merge. Scoping to
the tip commit loses nothing, because the two checks cover different gaps — check 2 already proves
every **ancestor** of `stage_head_commit` reached `origin/main`, so earlier Stage commits on the
branch are covered by ancestry, while the per-file check closes the gap ancestry cannot: that the
artifacts exist as readable files on the default branch.

**The handback-resolution no-halt invariant is preserved (AC-B1.5).** Every new halt lives in
**step 4's verification**, not in the handback defaulting. All fields still resolve to a
deterministic default without halting; what changed is that a degraded resolution now yields a set
that is either concretely verifiable or **provably empty**, and an empty one fails loudly instead
of passing silently. A gate that halts because it could not prove the artifacts are on
`origin/main` is the gate working, not the degraded path wedging.

**No new command shape enters the corpus.** `diff-tree`, `cat-file`, `rev-parse`, and `merge-base`
are read forms: none is a push verb, so §5.3 construct 1 cannot match them, and none grants
permission to commit anywhere, so construct 2 cannot either. The detector's closed construct set
and the fixture corpus are therefore **not** widened — the rev-6 non-expansion position is
unchanged. The shipment arm stays byte-verbatim (AC-B1.4), the resolution block's clean-tree and
single-worktree guards are untouched (AC-B1.9), and P-009, P-011, and P-016 remain textually
unchanged (AC-C2.6).

**Correction 5 (rev 7) — step 3 re-records the handback commit.** When step 1 finds a dirty tree,
**3a creates the commit that carries the artifacts** — and the handback's `stage_head_commit` was
resolved *before* that commit existed. Deriving step 4's file set from the stale value would read a
commit that changed none of the artifacts, halting the gate at check 3 for the wrong reason on the
one path Step 1.5 exists to repair. 3a therefore re-records `stage_head_commit` as `git rev-parse
HEAD` and discards any reported `stage_artifact_paths` (which described the pre-commit intent, not
this commit) so check 4 falls through to `VERIFY_SET := DERIVED`. When step 3 was entered for
unpushed commits alone and nothing was uncommitted, `HEAD` is unchanged and the re-record is a
no-op. AC-B1.3's branch-point text (`from the current commit`) is preserved verbatim.

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
(`status`, `rev-parse`, `fetch`, `log`, `show`, `merge-base`, and — rev 7 — `diff-tree`,
`cat-file`) and one `checkout`; 3b's existing push-the-branch-and-open-a-PR form is untouched, and
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
`stage_branch`, `stage_head_commit`, `stage_outcome`, and `stage_artifact_paths`. **Every** field
has a deterministic default, so the degraded path **never halts**: branch/commit resolve from
`HEAD`; `stage_outcome` derives from whether a `shipment_id` was returned; `stage_artifact_paths`
is left **empty** and derived in step 4 from `stage_head_commit`'s own changed-file set (rev 7 —
it is **never** defaulted to a directory list); `STAGING_HANDBACK_DEGRADED` is recorded once and
surfaced. When the `HEAD` resolution yields the default branch, `stage_branch` is left
**UNRESOLVED** and the run routes to step 3a's create-arm. The default branch is **never**
substituted for the Stage artifact branch.
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
`stage_outcome: no-shipment` that runs five ordered checks and halts with `STAGING_GATE_FAIL` on
any failure: (1) `git rev-parse --verify {stage_head_commit}^{commit}`; (2) `git merge-base
--is-ancestor {stage_head_commit} origin/main`; (3) derivation of `DERIVED` from
`git diff-tree --no-commit-id --name-only -r -z --diff-filter=d {stage_head_commit}` filtered to
`STAGE_ARTIFACT_ROOTS`, halting when it is **empty**; (4) resolution of `VERIFY_SET` per AC-B1.10;
(5) `git cat-file -t "origin/main:{path}"` for **every** path in `VERIFY_SET`, requiring exit 0
**and** output exactly `blob`. It terminates the gate without routing to Ship and names no
`shipment_id`. **Rev 7**: `git show origin/main:{path}` no longer appears in this arm — it exits 0
on a tree, so it could not reject the directory prefixes rev 6 defaulted to.
**AC-B1.9** **Clean-tree and single-worktree gates preserved on every path.** The branch-resolution
block sits **above** step 1's dirty/clean split, so `git rev-parse --verify` and the
HEAD-equals-`$STAGE_BRANCH` assertion execute on the dirty path too — 3a never checks out an
unverified branch name. Step 1.5 halts when HEAD is not the resolved staging branch; no new branch
or worktree is created when the handback reports one; and P-009, P-011, and P-016 remain textually
unchanged.
**AC-B1.10** **Concrete-path contract (rev 7).** `stage_artifact_paths` is defined as a **non-empty
set of concrete repository-relative file paths carried by `stage_head_commit`**. Step 4 rejects, by
halting: a path with a trailing `/`, a bare `STAGE_ARTIFACT_ROOTS` entry, or any glob/wildcard form
(not a file path); a path not under `STAGE_ARTIFACT_ROOTS` (outside the allowed Stage artifact
roots); an **empty** resolved set; any reported set that is not **equal** to the commit's derived
changed-file set (equality, not subset — a subset check would accept both an invented path and a
silently shrunken set); and any path that resolves on `origin/main` to an object type other than
`blob`. No directory prefix can satisfy completion on this arm, by construction: `diff-tree -r`
emits blob paths only, and `cat-file -t` must print exactly `blob`.
**AC-B1.11** **Handback commit re-recorded after step 3 commits (rev 7).** 3a sets
`stage_head_commit` to `git rev-parse HEAD` after committing and discards any reported
`stage_artifact_paths`, so step 4 derives its file set from the commit that actually carries the
artifacts rather than from the pre-commit `HEAD` the handback resolved. The re-record is a no-op
when step 3 was entered for unpushed commits alone. AC-B1.3's `from the current commit` branch-point
text is preserved verbatim.

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

> - Create and check out the dedicated Stage/admin artifact branch (`chore/stage-{shipment_id}`,
>   or `chore/stage-{chore-slug}` when no shipment is formed, per P-011's existing slug convention)

**Concrete no-shipment persistence route** (rev 4 — review round 3). Stating the actors end to end,
because "never to the default branch" is only actionable if the alternative route is executable:

| Step | Actor | Action |
|---|---|---|
| 1 | **Stage** | Create and check out `chore/stage-{chore-slug}` (grant above; mirrored into the `_stage.agent.md` Git row by B3, which is the cell role enforcement actually reads) |
| 2 | **Stage** | Commit the backlog/planning artifacts to that branch |
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

**AC-B2.1** No default-branch allowance remains in P-010's Stage block.
**AC-B2.2** The rule binds when no shipment is formed.
**AC-B2.3** P-010's Stage and Ship columns are symmetric on direct default-branch writes.
**AC-B2.4** No agent is left *prescribed-and-prohibited* on the same operation — every obligation
the new bullet creates names an authorized actor.
**AC-B2.5** Exactly one additive Amendment Log row, numbered 1.25.0; no prior row edited.

### 6.3 Task B3 — `_stage.agent.md` Role Boundary

Rewrite the **Git** row (deliberation §8.1 — the operative surface):

| Column | New text |
|---|---|
| Allowed | Commit backlog/planning artifacts to a **dedicated Stage/admin branch**; **create and check out that dedicated artifact branch** (`chore/stage-{shipment_id}`, or `chore/stage-{chore-slug}` when no shipment is formed); create/use an explicit, time-boxed spike/research worktree only for staging investigation |
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
requires (§6.1.2, correction 1): `stage_branch`, `stage_head_commit`, `stage_outcome`
(`shipment` | `no-shipment`), `stage_artifact_paths`, and `shipment_id` when one was formed.

**Path shape is part of that contract (rev 7).** The same line states that `stage_artifact_paths`
is a **non-empty list of concrete repository-relative file paths** — the files Stage committed in
`stage_head_commit`, never directory prefixes — and that Stage produces it with the same derivation
the Orchestrator re-runs:
`git diff-tree --no-commit-id --name-only -r -z --diff-filter=d {stage_head_commit}` filtered to
`.backlogit/`, `docs/plans/`, `docs/decisions/`, `docs/memory/`. Naming the producing command here
is what makes the consumer's **set-equality** check (AC-B1.10) satisfiable by construction rather
than by coincidence: two surfaces computing the same set from the same commit with the same command
cannot disagree. A producer left free to emit directory prefixes would either fail that check on
every run or force it to be weakened back into the always-true form rev 7 removed.

This is the **only** edit B3 makes outside the Role Boundary table, and it is not cosmetic. Without
a producer, the Orchestrator's handback requirement has no source and every pipeline run falls into
B1's `STAGING_HANDBACK_DEGRADED` path. It also completes the symmetry the paragraph above rests on:
B3 must carry the branch-creation grant because role enforcement reads *this file* at mutation
time — and by the same argument the Orchestrator consumes *this file's* declared session output, so
the branch Stage created must be declared here too. Placing the emission only in the Orchestrator's
expectations would repeat, in the opposite direction, the grant-in-one-surface-only defect AC-B3.2
exists to prevent. It transfers **no** authority: emitting a branch name is not pushing, opening, or
merging anything, so the PR-row prohibition is untouched.

**AC-B3.1** The Git row Allowed cell contains no default-branch allowance.
**AC-B3.2** The Git row Allowed cell grants creation and check-out of the dedicated artifact
branch, so the table and P-010 **agree** — no grant in one that the other omits or forbids.
**AC-B3.3** The Git row Forbidden cell is unchanged; every other Role Boundary row is unchanged
**except** the PR row, which is explicitly exempted to carry the actor note. The note attributes
an actor only — the PR row's prohibition is unweakened and no authority transfers to Stage. This
criterion governs the **Role Boundary table**; the Step 6 edit required by AC-B3.5 lies outside it.
**AC-B3.4** The file agrees with the corrected P-010 text — no surviving contradiction between
agent contract and policy, in either direction.
**AC-B3.5** **Staging handback emitted (rev 6).** Step 6's session-output contract requires
`stage_branch`, `stage_head_commit`, `stage_outcome`, `stage_artifact_paths`, and `shipment_id`
(when formed) — every field the Orchestrator's Step 1 consumes under AC-B1.5, with none required
there and absent here. **Rev 7**: the contract also fixes the **shape** of
`stage_artifact_paths` — a non-empty list of concrete repository-relative **file** paths carried by
`stage_head_commit`, never directory prefixes — and names the derivation
(`git diff-tree --no-commit-id --name-only -r -z --diff-filter=d {stage_head_commit}` filtered to
the four Stage artifact roots), so the emitted set satisfies the consumer's set-equality check
(AC-B1.10) by construction. This is the only change outside the Role Boundary table; no other
section of the file is modified.

These five criteria are carried verbatim by task `018.009-T`. The post-correction **zero-findings**
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
generated-baseline `go test` step in job `expensive`, so a render that drops this step turns CI
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
  marker-free prohibition constructs in the corrected tree — the post-change P-010 prohibition
  bullet, the Ship `workflow-policies.md` L238 bullet, and `_ship.agent.md`'s Role Boundary
  `| Git |` Forbidden cell — all score **accept**. Rev 1's AC3.3 (ci.yml step names) was vacuous
  because `.github/workflows/**` is never in the corpus. Prior art:
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
  defect class one level down, and passes a name-only comparison.
- **AC-C2.5** `markdownlint` exits **0** on every changed markdown file (P-008).
- **AC-C2.6** Branch and PR merge-commit rules intact — P-009, P-011, P-016 textually unchanged.
- **AC-C2.7** Constitution Quality Gates run **in declared order, none skipped, as separate
  commands** (never chained with `&&`, which would short-circuit and mask which gate failed):
  `gofmt -l .` empty, then `go vet ./...`, then `go test ./...`, then `go build ./...` — each
  exits 0 and each result is recorded independently. **Rev 4**: `go test ./...` is no longer a pure
  regression check — it now includes the `tests/integration/directpush_gate_test.go` harness (§5.0),
  which must be **green** in full, so all eight per-task functions pass here.
- **AC-C2.8** **Regeneration-resistant enforcement, recorded (rev 5).** The evidence records that
  the full-corpus scan of AC-C2.1 is executed by `TestDirectPushGate_FullCorpusClean` under
  `go test ./...` — i.e. from job `expensive`'s **generated-baseline** `Test (race)` step — and
  **not solely** by the job-`lint` step added in C1. Both halves of the §5.0 invariant are asserted
  green in the same run: **(a)** zero findings over the installed surfaces, and **(b)** the
  job-`lint` step still present, structurally and non-vacuously per AC-C1.6. A future render that
  drops the C1 step therefore fails (b) **loudly** while (a) continues to execute — which is what
  makes §1's "cannot silently reopen the gap" true as written rather than aspirational.

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
# Rev 7 — the no-shipment arm's primitives, verified against this repository:
git show origin/main:docs/plans/           # exits 0 on a TREE — why `show` is insufficient (AC-B1.10)
git cat-file -t origin/main:docs/plans/    # prints `tree`  → rejected
git diff-tree --no-commit-id --name-only -r -z --diff-filter=d HEAD   # blob paths only, no deletions
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

## 10. Constitution Check

Required by `.github/instructions/constitution.instructions.md` (Governance). Omitted in rev 1.

| Principle | Status | Note |
|---|---|---|
| I. Safety-First Go | **Satisfied** | Rev 4: one Go test file is added (§5.0) — no production Go surface; `go vet ./...` / `go test -race ./...` cover it |
| II. Test-First (NON-NEGOTIABLE) | **Satisfied — no deviation** | Rev 4 removes rev 3's declared deviation. See §10.1 |
| III/IV. Workspace isolation, CLI containment | **N/A** | No runtime code |
| V. Structured Observability | **Satisfied** | `::notice::` mode emission; AC-C2.1 records gate verdict as PR evidence |
| VI. Single Responsibility | **Satisfied** | One gate, one invariant; 8 tasks each single-domain |
| VII. Destructive Command Approval | **N/A** | No destructive operation |
| VIII. Explicit Safety Modes | **N/A** | — |
| IX. Git-Friendly Persistence | **Satisfied** | Additive Amendment Log row 1.25.0; §11 records the stale header `Version` field |
| X. Agent Context Efficiency | **Satisfied** | Ledger duplication reduced (§7.1 drops the magic count; residual stated once in the script header, cross-referenced elsewhere) |
| XI. Merge Commit History (NON-NEGOTIABLE) | **Satisfied** | P-009 untouched; AC-C2.5 |
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
| R5 | Regeneration wipes the job-`lint` CI step → gate silently disabled | **MITIGATED at rev 5 — no longer an accepted residual.** Enforcement also runs through `tests/integration/directpush_gate_test.go` (not generated → survives a render), invoked by the **generated-baseline** `go test` step in job `expensive` (a render *re-emits* it), and any render that rewrites `ci.yml` sets `changes.code == 'true'` so that job cannot skip. `TestDirectPushGate_FullCorpusClean` keeps executing the detector (requirement **a**); `TestDirectPushGate_CIWiringIsBlocking` fails **red** when the step is gone (requirement **b**). See §5.0, AC-C1.6, AC-C2.8 |
| R6 | Gate mistaken for a security boundary | Header states verbatim: **anti-accident control, not an anti-adversary control** — CI runs it from the PR head. CODEOWNERS (AC-C1.4) is the named mitigation |
| R7 | Legitimate line trips the detector | Narrow the **pattern** or document-and-exclude a path; never reduce corpus coverage |
| R8 | 6th hand-copied gate scaffold (no `scripts/lib` bash helper exists) | **ACCEPTED RESIDUAL**, recorded here with the existing clone-divergence precedent (stash `6C24E2E4`); extracting a shared helper is out of scope under C1 |
| R9 | A **fifth** authorization surface exists outside the §5.2 corpus | AC-A3.4 records the first full-corpus scan (file count + complete finding list) **before** B1 begins, so an unknown surface surfaces at A3 rather than at C2 |
| R10 | The §5.0 Go harness makes `go test ./...` depend on `bash` | **ACCEPTED RESIDUAL** (rev 4). CI runs on `ubuntu-latest`; every existing gate script already requires `bash` locally. The harness **fails rather than skips** when `bash` is absent (§5.0), deliberately, so the P-004 red phase can never be satisfied by a silent skip (P-012) |
| R11 | The §5.0 harness is a sixth CI-relevant surface a future render could disturb | Low: it is a plain `_test.go` file under `tests/integration/`, not autoharness-generated, and `go test -race -mod=readonly ./...` already runs in job `test`. No new CI wiring, no ledger entry, no CODEOWNERS line needed. **Rev 5** makes both of those properties **load-bearing rather than incidental** — R5's mitigation depends on the harness being non-generated *and* on that `go test` step being generated-baseline, so §5.0 now states and justifies both explicitly |
| R12 | The harness itself is deleted or neutered, disabling both halves at once | **ACCEPTED — but VISIBLE, not silent (rev 5).** Requires editing non-generated tracked files (`directpush_gate_test.go` and/or the detector) in a reviewable pull-request diff; **no render can do it**. `CODEOWNERS` is deliberately **not** claimed as the mitigation here — its own header records it is advisory-only until code-owner review is required on `main`, an operator action not taken. This is strictly narrower than rev 4's R5: the failure mode moves from *silent regeneration* to *visible human/agent edit*, which is what §1 actually promises to exclude. **If a future autoharness version begins generating `tests/**`, the §5.0 invariant must be re-verified** |
| R13 | A render reverts the rev-6 handback/discovery contract, restoring main-only discovery — a defect the **detector cannot see**, since it is neither §5.3 construct | **MITIGATED, same mechanism as R5 (rev 6).** The two per-surface harness functions carry it: `TestDirectPushGate_OrchestratorSurfaceClean` asserts the branch-specific discovery, the handback requirement, both step-4 arms, and the **absence** of the default-branch self-range literal; `TestDirectPushGate_StageAgentSurfaceClean` asserts the Step 6 emission. Both live in the non-generated `directpush_gate_test.go` and run from the **generated-baseline** `go test` step (§5.0), so a render that reverts either surface fails CI **red**. The literal-absence assertion is sound only because the corrected text states its prohibition descriptively (§6.1.2, self-match hazard; AC-B1.6) — a prohibition quoting the forbidden range would self-match and silently vacate the assertion |
| R14 | A verification loop passes over targets that are **always present**, proving nothing — the rev-7 finding | **MITIGATED (rev 7).** Structural, not procedural: `diff-tree -r` cannot emit a directory, `cat-file -t` must print `blob`, the derived set must be non-empty, and a reported set must **equal** the commit's changed-file set. `TestDirectPushGate_OrchestratorSurfaceClean` additionally asserts the **absence** of a `git show origin/main:{path}` loop and of any directory-prefix default in the no-shipment arm, so a render or edit that restores the rev-6 form fails CI **red**. **Residual, recorded not hidden**: the derived set is scoped to `stage_head_commit`'s own changes, so a Stage session that spreads artifacts over several commits has only its **tip** commit's files checked per-file; the earlier commits are covered by the ancestry check (check 2), not by a blob read. Widening to a range form is **rejected**, not overlooked — `origin/main..{stage_head_commit}` is empty by construction post-merge, which is exactly the vacuous loop this row exists to close |

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

<!-- plan-review-attempt: 2 -->

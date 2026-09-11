---
title: "Implementation Plan — Stage Artifact Branch/PR Policy Gap Correction"
date: 2026-09-11
revision: 5
status: reviewed
agent: Stage
governs: stash 638A410B
source: docs/decisions/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-deliberation.md
---

# Implementation Plan — Stage Artifact Branch/PR Policy Gap Correction

**Source document**:
`docs/decisions/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-deliberation.md`
(including its **§8 Revision 2 addendum** and **§9 Revision 3 addendum**)
**Stash entry**: `638A410B`
**Requires plan hardening**: **yes** — see §9/§10a.

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
| `.github/agents/_orchestrator.agent.md` | Reword Step 1.5 preamble; delete 3e; fold its branch-point into 3a. | B1 |
| `.github/policies/workflow-policies.md` | P-010 Stage bullets + Amendment Log **1.25.0**. | B2 |
| `.github/agents/_stage.agent.md` | Role Boundary Git row + PR row note. | B3 |
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
| B1 `018.007-T` | `TestDirectPushGate_OrchestratorSurfaceClean` | scan scoped to `_orchestrator.agent.md` reports 0 findings |
| B2 `018.008-T` | `TestDirectPushGate_PolicySurfaceClean` | scan scoped to `workflow-policies.md` reports 0 findings **and** the Amendment Log carries row `1.25.0` |
| B3 `018.009-T` | `TestDirectPushGate_StageAgentSurfaceClean` | scan scoped to `_stage.agent.md` reports 0 findings |
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
regeneration-evasion path named in §10a.1.

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

Delete sub-step **3e** in full (the "Branch protection handling" block and its three nested
bullets plus trailing rationale).

**Fold its branch-point into 3a** — required, because 3e's "Create branch
`chore/stage-{shipment_id}` **from the current commit**" is the **only** instruction covering the
*already-committed-but-unpushed* path that Step 1.5 step 2 routes into step 3. 3a currently
specifies no branch point. Revised 3a:

> a. Create `chore/stage-{shipment_id}` from the current commit and commit any uncommitted
>    backlog files to it.

**Rewrite the Step 1.5 preamble** (rev 3 — the fourth surface, §2). It currently states the gate's
postcondition as artifacts being "committed to the default branch", which remains a compliant
reading of direct commit-to-default. Revised:

> After Stage completes and before routing to Ship, verify that all staging artifacts (backlog
> items, shipment manifests) **have reached the default branch via a merged staging PR** and are
> present on the remote.

This is a **postcondition/verification** form, which §5.3 construct 2 explicitly does not treat as
a violation — so it satisfies the gate without narrowing the detector.

**Preserved unchanged**: 3b–3d; step 4's `git show origin/main:.backlogit/queue/{shipment_id}.md`
verification and `STAGING_GATE_FAIL` halt; the verified-on-origin broadcast. Step 1.5 remains
NON-NEGOTIABLE and **Orchestrator-owned**.

**No no-shipment clause here** (rev 2 correction). Step 1.5 is scoped "After Stage completes and
before routing to Ship" and step 4 requires a `{shipment_id}` manifest — a shipment-less Stage run
never enters it. Placing the rule here would put the control in the one path that cannot fire.
The no-shipment obligation lives in B2/B3, where the acting agent reads it.

**AC-B1.1** No direct-push-to-default-branch instruction remains anywhere in Step 1.5.
**AC-B1.2** The preamble states the postcondition as artifacts having **reached** the default
branch **via a merged staging PR** (rev 3's fourth surface — §2).
**AC-B1.3** 3a textually carries the branch point (`from the current commit`) for the
already-committed-but-unpushed path — textual, not structural, because the a–e sub-steps are lazy
paragraph continuations, not a real nested list (see §11 on the pre-existing list-rendering defect).
**AC-B1.4** Step 4's `git show origin/main:.backlogit/queue/{shipment_id}.md` verification and
`STAGING_GATE_FAIL` halt are byte-identical to their pre-change text.

**Rev 4 reconciliation**: this list previously carried only three criteria, worded before rev 3
added the preamble as the fourth authorization surface, while task `018.007-T` already carried the
four-criterion form and §6.3's preserved-form note already referenced **AC-B1.4**. The plan is
corrected to the task's form — the same plan↔task same-contract defect class as the C2 divergence
reconciled in §7.2, found by the rev-4 AC-ID parity sweep rather than by review.

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

**AC-B3.1** The Git row Allowed cell contains no default-branch allowance.
**AC-B3.2** The Git row Allowed cell grants creation and check-out of the dedicated artifact
branch, so the table and P-010 **agree** — no grant in one that the other omits or forbids.
**AC-B3.3** The Git row Forbidden cell is unchanged; every other Role Boundary row is unchanged
**except** the PR row, which is explicitly exempted to carry the actor note. The note attributes
an actor only — the PR row's prohibition is unweakened and no authority transfers to Stage.
**AC-B3.4** The file agrees with the corrected P-010 text — no surviving contradiction between
agent contract and policy, in either direction.

These four criteria are carried verbatim by task `018.009-T`. The post-correction **zero-findings**
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
  policy agree — no surviving contradiction across the four §2 authorization surfaces.
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

<!-- plan-review-attempt: 2 -->

---
title: "Implementation plan: mechanical enforcement of the Stage default-branch contract (020-F, re-stage)"
date: 2026-10-17
status: draft
source: docs/decisions/2026-10-17-intercom-go-020f-restage-and-one-task-rule-supersession-decision.md
supersedes: docs/plans/2026-09-12-intercom-go-stage-contract-enforcement-platform-plan.md (governing shape only; its scope discipline and provenance remain valid)
feature: 020-F
tasks: [020.001-T, 020.002-T, 020.003-T, 020.006-T, 020.004-T, 020.005-T]
---

# Implementation plan - 020-F re-stage

## Problem Frame

018-F corrected the Stage default-branch contract in two autoharness-GENERATED files: `.github/policies/workflow-policies.md` (P-010, P-016 and the role-boundary rows) and `.github/agents/_stage.agent.md` (Role Boundary Git/PR rows, Step Sequence line for Step 1.9, Step 1.9 gate). A regeneration can silently revert them. This feature makes the corrected contract mechanically enforced: a closed-set, structure-aware detector in Go, a fixture corpus proving it, blocking CI wiring, a divergence ledger, CODEOWNERS coverage and an executed mutation proof.

Code home: `tools/gatecheck/internal/stagecontract/` (the existing gate-engine location; `internal/` visibility isolates it; pathspecs anchored at `internal/**` and `cmd/**` never select it). Existing patterns: engine packages with `testdata/`, CI-wiring assertions in `tools/gatecheck/ciwiring_test.go`, CODEOWNERS at `.github/CODEOWNERS`, CI job `lint` in `.github/workflows/ci.yml`.

**Scope discipline (the original failure mode):** the construct set is CLOSED and tied to the corrected contract. No generalized policy-linting platform. A fixture that does not map to a named corrected-contract property does not belong.

**Shape decision:** one covering feature, one shipment, six sequential tasks (001, 002, 003, 006, 004, 005). The one-task-per-feature rule is superseded by operator direction (see source decision, section 2).

## Requirements Trace

| Req | Source | Unit |
|---|---|---|
| R1 corpus mapping to named properties, manifest with verdict + reason | 020.001-T | U1 |
| R2 table-driven driver over a stubbed detector; fail-loud on bad manifest | 020.002-T | U2 |
| R3 structure-aware closed-set detector | 020.003-T (parser), 020.006-T (rules + live-contract test) | U3, U3b |
| R4 blocking CI step, divergence ledger, CODEOWNERS | 020.004-T | U4 |
| R5 executed mutation proof over a workspace copy | 020.005-T | U5 |
| R6 per-task red-first harness with observed red-then-green evidence, disclosed (not authorized) DB12DA37 condition, no policy amendment | decision section 5 | U1-U5 |

## Named corrected-contract properties (the closed set, P1-P6)

* **P1** P-010 Stage commit grant is dedicated-Stage-artifact-branch-only.
* **P2** The actor-named Stage MUST NOT default-branch bullet.
* **P3** The `chore/stage-{scope-slug}` create/checkout grant.
* **P4** The P-016 spike/research worktree exception left intact.
* **P5** Role Boundary Git and PR rows (operator, never Stage, pushes and opens the PR).
* **P6** The registered fail-closed Step 1.9 gate (Step Sequence line plus section).

Verdicts: `violating` / `conforming`, with a reason string naming the property.

## Implementation Units

Common to every unit (R6, execution posture TEST-FIRST, just-in-time): the task's first commit is its own red harness (compilable, failing, expected failure marker recorded, `go vet ./...` clean). Harness for task N is written only when task N starts. The task END commit has full `go test ./...` green; the red first commit is mid-task by design and is NEVER pushed before the task is green (no branch CI run on a red commit).

**Honest characterization of the P-002/P-004 position (plan-review attempt 2, constitution P1-1/P1-2).** A shipment claim auto-activates the shipment's tasks before any harness exists (DB12DA37). That claim-ordering condition is the pre-existing, systemic, disclosed defect that affects every shipment in this workspace; this plan does NOT cure it, does NOT claim P-002 literal compliance on the claim route, and does NOT assert any operator `skip_policy` waiver. What this plan adds is a per-task sequencing so that the property P-004 protects (every generated test red before the implementation it covers, with the red observed) holds on the artifacts: Ship Step 2's "harness once, up front" is vacuous on this route and an up-front batch of red harnesses would break Step 4.3 after task 1, so harness authoring is per task. Authority for proceeding is the operator's instruction that Stage judge workability autonomously plus the Orchestrator's recorded dark-mode decision; the operator can veto by archiving the queued shipment before Ship claims it. Every task therefore carries two additional ACs: (a) a P-005 compliance annotation in the task comment and the PR readiness block stating "DB12DA37 claim-ordering condition disclosed; no waiver asserted; per-task red-then-green evidence below", and (b) the red-then-green evidence (command + failure marker, command + result). If either is absent the PR readiness block is incomplete and merge approval must not be treated as satisfied.

### U1 - 020.001-T: fixture corpus and manifest (test data; Size S, complexity low)

* Files: `tools/gatecheck/internal/stagecontract/testdata/*.md` (small contract-state fixtures, data only) and `testdata/manifest.json`; red/green harness `manifest_test.go` (manifest well-formedness only).
* Conforming fixtures are structure-faithful excerpts of the REAL corrected P1-P6 text (same headings/table/list shapes), so the detector built against them behaves on the real files.
* Corpus: at least one conforming and one violating fixture per P1-P6, plus at least one "quoted/historical/example text that mentions a construct but is not a violation" fixture (proves AC-4 of U3). Every fixture has `property`, `verdict`, `reason`.
* Red: manifest absent/ill-formed fails `TestManifestWellFormed`. Green: manifest present and every fixture file referenced exists, every property P1-P6 has both verdicts, no unmapped fixture.
* Posture: test-first. No detector logic, no CI wiring.

### U2 - 020.002-T: fixture-suite driver over a stubbed detector (Size S, complexity low)

* Files: `driver.go` (Detector interface, `ErrNotImplemented`, stub detector, `RunSuite`), `driver_test.go`.
* Behavior: read manifest, one case per fixture, assert verdict and reason via the interface; malformed/truncated/unreadable manifest FAILS (never empty/inconclusive); failure report names fixture id, expected verdict, observed verdict.
* Driver correctness is proven with injected test-double detectors (green). The corpus-bound test `TestFixtureCorpus` runs the real corpus through the stub; while the detector returns `ErrNotImplemented` it is **visibly skipped** (`t.Skip` naming 020.003-T) so Step 4.3 stays green. This skip is removed red-first in U3.
* Red: driver symbols stubbed to fail; driver tests fail with the expected marker. Green: driver tests pass, corpus test skipped with the 020.003-T reason.

### U3 - 020.003-T: structure-aware markdown block parser (Size S, complexity medium)

(Split from the original single detector task at plan-review attempt 1, finding P2-1: parser plus rules was likely over the 2-hour rule; the one-task rule is withdrawn so splitting is free.)

* Files: `blocks.go` and `blocks_test.go`. No new dependencies; a small line-oriented block parser written in-package (headings with path, table rows with cells, list items, fenced blocks, blockquotes, and sections marked historical/example), since a markdown dependency for a closed construct set is unjustified.
* Behavior: pure structure extraction, no verdicts. Exposes each block with its kind, heading path and a `quoted`/`historical` flag.
* Red: parser unit tests (table-driven, 3 scenarios: headings+lists, tables, fenced/blockquote/historical flags) fail against stubs. Green: pass. `TestFixtureCorpus` remains skipped (the detector is still the stub).
* Real-file robustness: parser unit tests include a test that parses both real contract files without error (structure only, no verdict).

### U3b - 020.006-T: closed-set P1-P6 rules and live-contract test (Size M, complexity medium)

* Files (three): `detector.go`, `detector_test.go` (includes `TestLiveContract`, so no extra file), and the `driver_test.go` edit that removes the skip (`driver.go` is unchanged: the skip lives in the test). Sizing is knowingly M: six small rule evaluators over a shared block walk; Ship STOPS AND REPORTS (no scope widening, no policy edit) if the task exceeds the 2-hour rule.
* Behavior: evaluates ONLY the closed enumerated construct set P1-P6 over U3's blocks; quoted, blockquote, fenced or historical text never yields a violating verdict. Reason strings name the property. Unknown content yields no verdict.
* `TestLiveContract` (read-only) evaluates the two REAL generated files (`.github/policies/workflow-policies.md`, `.github/agents/_stage.agent.md`) and requires conforming verdicts for P1-P6. It uses `t.Fatal`, never `t.Skip`, if a real file is missing or unreadable. It lives here, not in U4, so a detector/real-file mismatch surfaces before CI wiring. If the live verdict is violating, Ship stops, records P-005 and reports; it never edits the contract files or weakens the detector.
* Red (first commit): the `t.Skip` is removed from `TestFixtureCorpus` (stub still returns `ErrNotImplemented`, which now FAILS) and `TestLiveContract` is added failing. Green: the entire corpus passes, live files conform, no skip remains.
* ACs include: construct set enumerated in code with each member mapped to P1-P6; quoted/historical fixture conforming; no generalized lint platform (no plugin/registry/config surface).

### U4 - 020.004-T: CI wiring, divergence ledger, CODEOWNERS (config; Size S, complexity medium)

* Files: `.github/workflows/ci.yml` (one blocking step in the existing `lint` job inserted after `Set up Go`; SHA-pinned actions untouched), `.github/CODEOWNERS` (two lines + header sentence amendment), a divergence-ledger entry (new short file `docs/design-docs/stage-contract-divergence-ledger.md`; Ship creates the directory entry if absent), and the CI-wiring assertion added to `tools/gatecheck/ciwiring_test.go` (updating the existing `lint` job expectations on ordering/steps in the same commit).
* The blocking step (with `timeout-minutes`, no `if:` guard, hermetic: no network, no new `uses:`, no package installs) runs under `set -o pipefail`: `go test -json -count=1 -mod=readonly -run '^TestLiveContract$' ./tools/gatecheck/internal/stagecontract/ | tee` to a file, then asserts with `jq -e` (after `command -v jq` fail-closed check; `jq` is preinstalled on `ubuntu-latest`, never installed by the step) that a JSON event with `.Action=="pass" and .Test=="TestLiveContract"` exists. A text grep is NOT used (compound note: self-matching greps). A vacuous "no tests to run" exit 0 or a `skip` event FAILS the step. The ciwiring assertion also requires the `func TestLiveContract` symbol, the anchored `-run`, the `jq -e` check, placement in job `lint`, absence of `continue-on-error`, absence of an `if:` on the step, and that `go.mod`/`go.sum` and all `uses:` pins are unchanged by this task.
* Red: the ciwiring assertion fails before the edits. Green: it passes; U2/U3/U3b suites unchanged.
* Sizing declared: four small config/docs-class edits in one skill domain (config); Ship stops and reports if it exceeds the 2-hour rule.
* CODEOWNERS: add lines for the two generated contract files AND for `tools/gatecheck/internal/stagecontract/` (detector, fixtures, tests), and amend the header sentence "ALL CI-critical scripts and workflow surfaces, or none" to include them as governed content. Keep the advisory-only disclaimer. Do NOT enable required code-owner review (separate operator action; dark-factory PRs would deadlock).
* Ledger entry and ci.yml header note must state plainly: enforcement is TAMPER-EVIDENT, NOT TAMPER-PROOF. A `pull_request` workflow runs from the PR head, so a PR can alter the contract, detector and workflow together; the gate reliably catches ACCIDENTAL regeneration reverts only. Real enforcement needs branch protection with required code-owner review or a base-ref required workflow (operator-only, out of scope).
* Local-divergence notice for regeneration is added to the ci.yml header comment, following the existing LOCAL DIVERGENCE convention. If `lint` is skipped by the `changes` job (docs/backlog/.autoharness-only PRs) the gate is skipped; that is accepted and recorded in the ledger.

### U5 - 020.005-T: executed mutation proof (tests; Size S, complexity medium)

* Files: `mutation_proof_test.go` (one test) plus helper in the same file.
* Behavior: copy exactly `go.mod`, `go.sum`, `tools/gatecheck/**` and the two contract files into a `t.TempDir()` (never mutate the real checkout). Copy rules: regular files only via explicit walk, `Lstat` and FAIL on any symlink, reject `..` and verify each destination stays under the temp root, exclude `.git`; preserve relative depth so `TestLiveContract` in the copy resolves its files relative to the copy. Repo root resolved from the stagecontract package directory as `../../../..` (local helper; the ciwiring helper uses a different depth). Mutation: remove EVERY occurrence of the `chore/stage-{scope-slug}` create/checkout grant (P3) from the COPY of `_stage.agent.md` (or assert exactly one occurrence), with ACs that bytes changed AND the observed failure names P3. Run the wired step's `go test -json -run '^TestLiveContract$'` as a child via `exec.CommandContext` with a 5-minute timeout, `cmd.Dir` = the copy, `cmd.WaitDelay` set, killing the process tree on timeout; parse the event stream as JSON events (malformed/truncated stream fails the proof); require FAIL naming P3; restore the COPY only and require PASS with the `pass` event for `TestLiveContract`.
* Real-tree guard: snapshot `git status --porcelain=v1 -z --untracked-files=all -- <pathspecs>` scoped to the two contract files and the stagecontract package directory before and after, AND compare SHA-256 of the two real contract files before and after (status lines alone cannot see content changes to already-modified files); both byte-identical. A repo-wide snapshot is flaky when `go test ./...` runs packages in parallel. If `git` or `.git` is unavailable the git half `t.Skip`s with a named reason (the SHA-256 half still runs); it never silently passes.
* Child environment: NEVER inherit `os.Environ()` (CI tokens must not reach the child). Build an explicit allowlist: `PATH`, `HOME`, `USERPROFILE`, `LocalAppData`, `AppData`, `SystemRoot`, `TEMP`/`TMP`, plus values resolved in the parent from `go env GOROOT GOCACHE GOMODCACHE GOPATH`, and fixed `GOFLAGS=-mod=readonly`, `GOTOOLCHAIN=local`, `GOPROXY=off`, `GOWORK=off`, `STAGECONTRACT_PROOF_CHILD=1` (recursion guard: the child run skips `mutation_proof_test` itself). The child must not write to the real module cache. Must work on Linux and the Windows advisory job without `-race`/cgo dependence; the Windows job is an AC.
* Cost accepted explicitly: the nested `go test` runs wherever `go test ./...` runs (expensive, windows, lint); bounded by the 5-minute timeout and a warm cache.
* Red: helper stubs make the proof test fail; Green: real proof passes. Evidence recorded. Deadline exceeded or killed child FAILS the proof.

## Dependency Graph

`020.001-T -> 020.002-T -> 020.003-T -> 020.006-T -> 020.004-T -> 020.005-T` (strict chain; 020.006-T is new from the plan-review split; execution order is by dependency, not by ID number). 018.008-T is archived/done (corrected contract delivered); 021.001-T done. No cross-strand edges. 026-F/025-F/019-F excluded.

## Decisions and Rationale

* One shipment, six sequential tasks: reasonable size (about 4 S + 2 M), single PR, same domain.
* Skip-until-U3 design for the corpus test: keeps Step 4.3 green after U2 without a deliberately red committed test.
* Same `go test -json` stream for CI and for the mutation proof, so U5 proves the actual gate.
* Closed construct set and in-package block parser: avoids a platform and a new dependency.
* No required-code-owner-review enablement: avoids unattended-merge deadlock (existing CODEOWNERS risk R1).

## Risks and Caveats

* **DB12DA37**: shipment claim auto-activates all five tasks, so Ship Step 2 is vacuous. Mitigation: per-task first-commit red harness plus observed red-then-green evidence as acceptance criteria (R6); reviewers of the PR check the evidence. Not fixed here (026-F). Residual: if Ship ignores the ACs the bypass is possible; the PR readiness block must make the omission visible.
* Detector false verdicts on the real files if the generated files are legitimately restructured; mitigated by structure-aware walk and the ledger.
* U3 sizing: M with a block parser; if it exceeds the 2-hour rule Ship must stop and report rather than widen scope.
* Windows path/CRLF: fixtures and manifest read with newline normalization; tests run on the Windows advisory job.

## Plan Hardening Signals

* Public API, schema or contract change: PRESENT (a contract-enforcement gate; fixture manifest schema).
* Security/auth/compliance-sensitive: PRESENT, low (CODEOWNERS/CI-critical paths).
* Migration/destructive/irreversible: ABSENT.
* External integration or operator checkpoint: PRESENT (CI workflow edit; branch-protection enablement explicitly out of scope).
* High runtime, rollout or rollback risk: PRESENT, moderate (a blocking CI step can break every PR; rollback = revert the single step).

Requires plan hardening: yes

## Runtime Verification and Closure

* Runtime surfaces: the CI `lint` job (new blocking step). Verification: the step runs green on the shipment PR; U5 proves it fails when the contract is reverted.
* Closure artifact: divergence-ledger entry and the PR readiness block with red/green evidence per task. Rollback trigger: unexplained red `lint` on unrelated PRs after merge; rollback = remove the one step (ledger records this). Owner: repository owner via CODEOWNERS.

## Plan Hardening

**Hardening required:** yes (contract gate, CI blocking step, CODEOWNERS, DB12DA37 residual).

**Consulted:** `docs/compound/2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md` (a CI gate whose pattern matches its own text becomes self-defeating - applied to the detector: the live contract files and the detector's own fixtures must not be matched by a prose sweep; U3 AC-4 and the structure-aware walk), `docs/compound/2026-09-30-shipment-reconcile-skill-bypass-manual-substitution-risk.md` (do not substitute manual steps for a skill; Ship uses harness-architect and shipment-reconcile as written), `.github/CODEOWNERS` header (advisory-only disclaimer, risk R1), `.github/workflows/ci.yml` (`changes` job is a DENYLIST: only `docs/**`, `.backlogit/**`, `.backlog/**`, `.autoharness/**` are excluded, so edits to `.github/policies/**` and `.github/agents/**` DO trigger `lint`; verified by reading the filter).

### ProposedAction / ActionRisk

| Action | ActionRisk | Approval | Carried into |
|---|---|---|---|
| Add a blocking step to the `lint` job in `ci.yml` | high blast radius (a false positive blocks every code PR) | normal PR review; no operator checkpoint beyond merge approval already authorized for scope | U4 |
| Add CODEOWNERS lines for the two generated contract files | low (advisory-only until branch protection requires code-owner review) | none; do NOT enable required code-owner review | U4 |
| Execute a mutation proof that edits a COPY of contract files | medium (must never touch the real tree) | none; guarded by byte-identical `git status` snapshots | U5 |
| Add a skip-then-unskip corpus test across U2/U3 | low | none; skip is removed red-first in U3 and AC forbids any remaining skip | U2/U3 |

### Invariants to preserve

* The real working tree is never mutated by U5 (AC-4 snapshots; temp-dir copy only).
* `continue-on-error` is absent on the new `lint` step (blocking). No advisory toggle is introduced.
* The detector evaluates only the closed construct set P1-P6; unknown content yields no verdict rather than a violation (no false positives from historical/example text).
* No policy text (P-002, P-004, P-010, P-016, Ship Step 2) is amended; the 018-F corrected contract is read, never edited, by this feature.
* A task's red harness exists only after that task starts; no deliberately red test remains committed at any task boundary (Step 4.3).

### Verification depth

* Environment precheck at task start: clean worktree on the shipment feature branch; `go version` matches `go.mod`; `go vet ./...` and `go test ./...` green BEFORE the red harness is added (baseline evidence).
* U1-U3: corpus completeness test (each of P1-P6 has both verdicts; the quoted/historical fixture present); detector driven by manifest, not hard-coded fixture names.
* U4: `TestLiveContract` passes on the real files; the ciwiring assertion proves the step exists, is in job `lint`, and has no `continue-on-error`. The workflow YAML is validated by the existing actionlint/ciwiring checks in CI; the PR's own `lint` run is the runtime verification.
* U5: both outcomes (fail naming P3 on mutation; pass on restore) observed in one run from the JSON event stream; timeout bound; truncated stream fails.
* Blocked-path handling: if the detector disagrees with the real contract files (live verdict violating), Ship stops, records a P-005 and reports; it does not edit the generated contract files or weaken the detector to pass.

### Operational closure

* Monitoring signal: `lint` job result on the next three code PRs after merge.
* Rollback trigger: a red `lint` step on a PR that does not touch the contract files, or any false verdict on the real files not explained by a genuine revert.
* Rollback procedure: revert the single `ci.yml` step (detector and tests remain; ledger entry updated to note the step is removed).
* Owner: repository owner (`@softwaresalt`) via CODEOWNERS. Validation window: three subsequent PRs.

### Human checkpoints and out-of-scope

* Enabling required code-owner review in branch protection is an explicit OPERATOR action and is out of scope (deadlock risk for unattended PRs).
* DB12DA37 is not fixed here (026-F, blocked on operator-only O-1). Residual: Ship Step 2 vacuity. Mitigated only by per-task acceptance criteria and PR-readiness evidence.

### Review-gate capability risk

Plan review must emit literal `dispatch_mode:` and `decision:` markers. If a persona cannot be dispatched as an independent subagent, the fallback must be declared and the review treated as degraded; a degraded review cannot return PASS for a plan carrying the DB12DA37 mitigation.

### Unresolved operator decisions that block safe execution

None that block staging. Branch-protection enablement and 026-F/O-1 remain operator-only and non-blocking for this feature.

<!-- plan-review-attempt: 1 --> FAIL (3 P1, 5 P2), remediated.
<!-- plan-review-attempt: 2 --> remediation applied; see review record.

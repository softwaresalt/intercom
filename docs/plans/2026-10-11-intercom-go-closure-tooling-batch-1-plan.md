---
title: "Closure tooling Batch 1: one authoritative shipment-close procedure and scoped harness reconciliation: implementation plan"
date: 2026-10-11
source: "docs/decisions/2026-10-11-intercom-go-closure-tooling-batch-1-deliberation.md"
stash_ids: ["D10D3AFC", "A83C398E"]
status: "reviewed"
base: "origin/main@cc34242"
---

# Closure tooling Batch 1: implementation plan

**Source:** `docs/decisions/2026-10-11-intercom-go-closure-tooling-batch-1-deliberation.md`
(decisions D-CT-1 through D-CT-7). **Stash IDs:** D10D3AFC, A83C398E.

**Sequencing contract (operator directive).** This release unit MUST ship before any new implementation
shipment. Batch 2 (write-path stash EEE39D23, DBE25DF5, 858516B9, EDC18D59, 2613D248) and Batch 3 (stash C5E50EFD
start-script fix) are NOT staged here. When Stage later creates those shipments it MUST record
`backlogit dep add <new-shipment> <this-shipment>` (type `blocks`; the same `dependencies:` shape as
042-S -> 041-S) before handing the shipment to Ship, so `autoharness gate dag-readiness` cannot report them
ready ahead of this one. This shipment itself records no upstream dependency: 041-S and 042-S are shipped and archived.

**Harvest:** see the Harvest record at the end of this file.

**Learnings applied:** `docs/compound/2026-09-30-shipment-reconcile-skill-bypass-manual-substitution-risk.md`,
`2026-10-11-classify-only-writer-evidence-before-cascade-close.md`, `2026-10-08-bound-cascade-close-procedure.md`,
`workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md` (six-surface matrix; the
reason U1 exists), `workflow-issues/ship-step6-post-merge-branch-ordering-2026-09-19.md` (step 6.0 stays first).

## Problem Frame

Closing a shipment today depends on an agent running a prose binding (`CLASSIFICATION_BINDING`, local to this
repo since 021-S) that has been skipped or approximated in six closures, while the upstream tool
(`autoharness shipment cascade-close`, `autoharness gate closure-evidence`) performs the same revalidation in code
with durable evidence but is forbidden by the installed Ship rules. The installed harness also carries stale
routing variables and six failing targeted checks. The batch makes the tool-enforced close the single procedure,
retires the local prose control and its pinned tests together, and re-renders only the surfaces that are stale.
Evidence is in the deliberation (premise verification P1-P7).

**Operator checkpoint (approval point for D-CT-1, enforced).** Retiring the 021-S control (and its tests) and
rewriting the P-010 Ship grant wording (U6) are Stage's evidence-based decisions taken while the operator was absent.
The operator's approval point is review and merge of the staging PR for this branch. **Ship MUST NOT claim this
shipment until the operator has approved and merged that staging PR** (the established Stage-artifact-branch path,
P-010: the operator, never Stage, pushes and merges it). Rejecting or amending D-CT-1 or the grant wording there stops
this shipment with no other effect. The (a1) gate and every protected invariant stay regardless. The per-operation
approval for the U13 file deletion (Principle VII) is a separate, later checkpoint at execution time.

## Requirements Trace

| Req | Source | Units |
|---|---|---|
| R1 One authoritative close procedure; `ship` never called directly | D10D3AFC, operator condition 2, D-CT-1 | U3, U4, U5, U6, U7 |
| R2 Pre-mutation fresh revalidation is tool-enforced, not agent-approximated | D10D3AFC, D-CT-2 | U4, U5, U12 |
| R3 Adopt `cascade-close` and `gate closure-evidence` | A83C398E | U4, U5, U7, U15 |
| R4 Re-render stale routing variables; fix the six failing targeted checks (named in A83C398E) | A83C398E, D-CT-4 | U7, U8, U9, U10, U15 |
| R5 Repo-local safeguards survive (a1); pinned tests move without a red commit | D-CT-3, D-CT-5 | U1, U2, U5, U11, U13 |
| R6 Compound guidance that instructs the retired or forbidden procedure is reconciled (a consequence of D-CT-1, same contract surface) | D-CT-1 / learnings | U14 |
| R7 Routing decisions are not made silently | D-CT-6 | none (operator-decision stash entry) |

## Constitution Check

* **I Safety-First Go:** no production Go changes (test code only in U2, U11-U13). **V Structured Observability:** satisfied
  by evidence, not code: `close_evidence` records, the U15 P-005 disclosure, and the PR-recorded red/green demonstrations.
* **II Test-First:** U2 **expands** the pinned contract tests (every assertion kind, not only presence checks) to accept
  either the old or the new text (green), the text units U3-U7 then change text (green at every commit), and U11
  **contracts** the tests to the new text only (strict) with a demonstration that a reverted U5 line fails it. This is the
  standard expand/contract migration; there is no known-red commit, which Ship's per-task full-suite gate (Step 4.3)
  requires. To keep a real red observation, U2 includes a sub-check: a temporary new-only assertion is shown red against
  unchanged `main`, then removed. See "Justified deviations" for the one recorded departure from strict per-task red.
* **III / IV Containment:** `autoharness verify-workspace` and the auto-tune preview write only to the gitignored
  `.autoharness/staging/`. Non-test scratch goes under `dist/` or `logs/`. Test fixtures use `t.TempDir()` and pass
  absolute `--cwd`/`--workspace` (conflict record: `t.TempDir()` is outside the working directory; this is the repository's
  existing test convention, recorded in the 042-S plan; the simpler alternative, a fixture under `dist/`, is rejected
  because `backlogit`/`autoharness` would discover the real workspace's `.backlogit`/`.git` from a nested path).
* **VI Single Responsibility / Task Granularity:** one skill domain per task; every unit is at most 2 hours; the
  file-count heuristic is exceeded only where recorded in "Justified deviations".
* **VII Destructive Command Approval:** the mutating `cascade-close` becomes Ship's sanctioned path; every use
  still requires the existing per-closure operator approval, and **dark mode does not waive it** (P-017). U4 and U5
  carry both statements as acceptance criteria.
* **VIII Explicit Safety Modes:** Ship runs U3-U10 and U13 in **Careful** mode (`safety-modes` skill: managed-agent
  edits, the safety-control rewrite, and test deletion), scope frozen to the files each unit names.
* **IX Git-Friendly Persistence:** text and tests only. **XI Merge Commit History:** merge-commit PR, no squash/rebase.
* **P-010:** Stage writes artifacts only; every `.github`, `tests`, and manifest edit is Ship work.
* **P-021 C1:** nothing outside the two stash scopes is added; new discoveries are Stage-owned stash entries.

### Justified deviations (Constitution "Conflict resolution": principle, justification, simpler alternative)

* **U10 touches 18 files** (Task Granularity: fewer than 3 files). Justification: one identical single-line
  transformation (`name:` key) with a deterministic verifier (`frontmatter_conformity`) and effort well under
  2 hours; `A83C398E` names this failing check explicitly. Simpler alternative rejected: six 3-file tasks add ceremony
  and six identical review cycles without reducing risk.
* **U9 mixes the Ship agent routing block and `harness-manifest.yaml` variables.** Justification: both are rendered
  outputs of the same `config.model_routing` values; changing one without the other leaves `verify-workspace` reporting
  a new inconsistency between commits. U8 does not need the manifest: its checks (`context_tier` / propagation / retry
  directive text) do not depend on recorded variable values. Simpler alternative rejected: two tasks with an inconsistent
  intermediate state.
* **U2 and U11 rework a 33-row table-driven test** (Task Granularity: fewer than 4 test scenarios, fewer than 5 functions).
  Justification: it is one table in one file driven by one helper family, and the U1 row map is the single source of truth;
  splitting by row group would leave the table half-migrated between commits. Simpler alternative rejected: split by row
  group (each split commit would need its own either-form scaffolding).
* **Principle II red phase:** U2 passes on unchanged `main` by construction (expand step), so the new needles are not red
  at U2; the red evidence is the U2 temporary new-only assertion and the U11 reverted-line demonstration. Justification:
  Ship's Step 4.3 requires a green suite at every task, which a literal red-first order cannot satisfy for a text-pinned
  contract. Simpler alternative rejected: a known-red commit series (fails the per-task gate).
* **U12 has five assertions** (exit code, verdict, `mutation_possible`, evidence path, no-clobber second run) against one
  fixture and one command; they are one scenario's observations, not five scenarios.
* **No other deviation.** The text tasks touch one or two files each.

## Scope Boundaries

In scope: closure-contract text in the installed Ship agent, `shipment-reconcile` and `operational-closure` skills, the
policy file, the circuit-breaker instruction row; the pinned Go contract test, the retirement of its replica tests, and
one new smoke test; routing re-render of the orchestrator, stage, and ship agents and the manifest variables; the 18
skill `name` keys; two compound entries that instruct the forbidden or retired procedure.

Out of scope (each recorded): editing `.autoharness/config.yaml` (R1-R3 routing decisions, operator-owned); implementing
a SAFE_CLOSE executor and a tagged upstream release containing 192-F plus the CI pin bump (upstream autoharness, captured
as a stash entry, **A204A0BD**); the operator routing decisions R1-R3 (stash entry **0A215E0B**); the single-writer lock contract (`9F824B64`); closure-artifact naming (`C98B92F0`); dropping or
upstreaming the (a1) gate; a blanket tune of the other ~90 artifacts; setting `tuned_at`; Batches 2 and 3.

**Named residual risks (accepted, tracked):**

* **SAFE_CLOSE has no executor** and loses the agent-run binding. What remains on that path is still agent-executed prose
  (the skill's pre-mutation protected-set baseline gate and verify-after-each on every single-artifact archive): the same
  control class D-CT-2 rejected for CASCADE, accepted here only because the SAFE_CLOSE path has no destructive cascade
  hazard (every mutation is one non-cascading artifact). The executor is tracked as upstream stash entry **A204A0BD**
  (verified to exist before `D10D3AFC` is archived).
* **The mutating path and exit codes 5-8 are pinned by text only.** The smoke test (U12) exercises `--classify-only` only.
* **CI cannot run these commands:** `.github/constraints/autoharness-lock.txt` pins the PyPI `1.5.0` artifact, which
  predates 192-F (the installed CLI is an editable upstream snapshot that also reports `1.5.0`). The shipped text must
  halt on exit 1 ("CLI too old") and the smoke test skips with an explicit message when the `shipment` command is absent.

## Contract-surface matrix (built by U1 before the first edit; seed here)

| Surface | Retired/changed content | Owner | Pinned by |
|---|---|---|---|
| `.github/skills/shipment-reconcile/SKILL.md` | `classify-close-path` mode, binding, `classification_binding` input, unbound refusal, migration note; Cascade Sub-Procedure caller; adopt upstream flat sets and Linked-Deliberation Disposition | U3, U4 | contract test |
| `.github/agents/_ship.agent.md` Step 6.1 b and Step 7 | classify-close-path routing, "safe-close with binding"; adopt command routing by mode + exit code, `selected_close_path`, disposition-report consumption | U5 | contract test |
| `.github/policies/workflow-policies.md` P-015, P-010 | fully-covered-root exception (155-S safeguards: parity table required), Ship grant wording, version row | U6 | contract test rows 20b/20c and the `018.` guard |
| `.github/skills/operational-closure/SKILL.md`, `circuit-breaker.instructions.md` | `close_path`/`close_evidence`, gate step, Source artifact cleanup, timeout row | U7 | `closure_source_artifact_cleanup`, `TestOperationalClosurePostMergeFilenameConformance` (parses `## Output`) |
| orchestrator, stage, ship routing blocks; manifest variables | stale escalation/context_tier/propagation directives and 10 variables | U8, U9 | `stage_branch_gate_test.go`, `staging_gate_actor_contract_test.go`, verify-workspace |
| 18 skill frontmatters | add `name` | U10 | `frontmatter_conformity` |
| `ship_feature_completion_contract_test.go` | expand then contract; keep shared assert helpers | U2, U11 | itself |
| replica tests (2 files) | delete after the smoke test exists | U13 | compile |
| two compound entries | supersession notes | U14 | none |
| **Immutable history (never edited):** `docs/closure/**`, `docs/memory/**`, `docs/archive/**`, `docs/plans/**`, `docs/reviews/**`, `docs/decisions/**`, `.backlogit/**`, `.autoharness/continuous-learning/**` | retired tokens preserved | none | n/a |

## Implementation Units

Each unit is one task. `size` (effort) and `complexity` (difficulty/uncertainty) are independent axes. Every unit
is at most 2 hours of human-equivalent effort. Every unit leaves `go test ./...` green (Ship Step 4.3).

### U1: Contract-surface matrix and scoped-tune preview (docs)

* Size S, complexity medium. Posture: characterization-first (read-only except the one new doc).
* Run `autoharness verify-workspace --workspace .` and the auto-tune preview. Write
  `docs/design-docs/2026-10-11-closure-contract-surface-matrix.md` with: a keep / retire / adopt row for every surface
  that mentions a retired token or is named above, with the upstream template counterpart and pinning test; an explicit
  `RETAIN: local overlay` row for the (a1) gate, its P-010 grant, and the intake-reconciliation relocation (021-S); a
  **row-by-row map of the 33 contract-test rows to their owning text unit** (this drives U2/U11); a **P-015 safeguard
  parity table** (each 155-S safeguard -> the upstream command behavior that covers it, or a named gap); and the baseline:
  0 blockers, the stale-variable count (10 today vs the 30 in the stash, reconciled by how instances are counted), the six failing checks.
* Acceptance: (1) the doc lists every file returned by `git grep -i -E "classify-close-path|classification_binding|fully-covered-root"`
  outside the immutable-history list above; (2) every row has a disposition and an owner U2-U14; (3) any P-015 parity
  gap is a named residual, not silent; (4) baseline numbers are copied verbatim from the report.

### U2: expand the contract test (Go test)

* Size M, complexity medium. Posture: test-first. File: `tests/integration/ship_feature_completion_contract_test.go`.
* Make **every** assertion that pins retired text accept **either** the old or the new contract, for every assertion
  kind the file uses: presence, absence, `Exactly`/`AtLeast` counts, `OrderedBefore`, and the section-bounded block checks
  (for example, row 30 and the classification-section bounding must pass when the `### Classify-Close-Path Mode` heading is
  present *or* absent). Expand **`TestCascadeRoutesThroughSafeClose`** too: its "designates ... `mode: safe-close`" presence check
  and its "invoke the cascade" absence check (which must not match the substring of "invoke the cascade-close") are made
  either-form. The new-form needles are cut **verbatim** from the upstream templates (U1 records the source lines) and
  are proven against the real text by U11. Spelled out: row 30's `t.Fatalf` on a missing
  `### Classify-Close-Path Mode` heading becomes "bounded block if the heading exists, else the new-form check"; row 10's
  `Exactly` becomes "exactly once, old OR new"; the "invoke the cascade" absence uses the regexp `invoke the cascade([^-]|$)`
  so it cannot match "invoke the cascade-close". **Keep** the shared helpers `assertPresent`, `assertAbsent`,
  `assertPresentAtLeast`, `assertPresentExactly`, `assertOrderedBefore`, `countOccurrences`, `loadInstructionSurfaces`
  (other test files call them) and the name `TestShipFeatureCompletionContract33Rows`.
* Acceptance: `go test ./tests/integration -run '^(TestShipFeatureCompletionContract33Rows|TestCascadeRoutesThroughSafeClose)$' -count=1 -v`
  prints `--- PASS` for both on the unchanged `main` text; a temporary new-only assertion is shown red against `main` and
  then removed (recorded in the PR); `go vet ./tests/integration/...` is clean; no assertion was dropped; rows that
  compare surfaces accept every combination of per-surface old/new forms (the surfaces migrate in different tasks).

### U3: shipment-reconcile skill: ADD the command-routed contract (docs)

* Size M, complexity medium. Blocked by U2. File: `.github/skills/shipment-reconcile/SKILL.md`. **Additions only**: the
  retired text stays until U4, so the either-form tests see the old form present and the new form added.
* Adopt the upstream routing and sub-procedure text: Step 0(c) reads the verdict recorded by
  `autoharness shipment cascade-close --classify-only`; CASCADE executes only through the mutating command (operator
  destructive-command approval, never waived in dark mode, `--timeout` sizing, stay-attached/poll rule); the exit-code
  routing table **by mode and exit code** (classify-only exit 3 selects the SAFE_CLOSE sequence; mutating exit 3 halts;
  exit 4 never answered with SAFE_CLOSE; exits 5-8 halt to the operator with no commit, retry, or direct call; any other
  exit including 1 halts as "CLI too old"); consume `selected_close_path`, not `classifier_verdict` alone; the **flat**
  `allowed_ids`/`required_ids` postcondition sets; the **Linked-Deliberation Disposition** step (P-015 INV-12) and its
  report; the minimum-version note.
* Acceptance: (1) the Cascade Close Sub-Procedure names the command as the only caller of the cascade primitive; (2) the
  exit-code table matches `autoharness shipment --help` verbatim for codes 0 and 2-8; (3) the adopted lines match
  `templates/skills/shipment-reconcile/SKILL.md.tmpl` in the upstream checkout (read-only comparison, line refs recorded);
  (4) the file states that approval is not waivable by dark mode; (5) the full suite is green; (6) the diff contains no
  deleted line.

### U4: shipment-reconcile skill: RETIRE the binding contract (docs)

* Size M, complexity medium. Blocked by U3. File: same skill. **Removals only**, after the replacement exists.
* Remove `mode: classify-close-path` (and its enum entry), the `classification_binding` input, the canonical-binding
  serializer, the bound-revalidation and "unbound refused" steps, the binding migration note, the `RECONCILE_FAIL_CLASSIFICATION_*`
  / `_CASCADE_UNBOUND` tokens, and the SAFE_CLOSE/CASCADE branches that depend on them. Keep the SAFE_CLOSE sequence, its
  pre-mutation protected-set baseline gate, and verify-after-each. Pre-Mode's in-process classification stays advisory and is
  never a close-path source.
* Acceptance: (1) none of the retired tokens remain in the file; (2) Pre-Mode, Post-Mode and `detect-mixed-role` text is
  unchanged except for deleted references; (3) the full suite is green; (4) only this file changed; (5) the SAFE_CLOSE
  baseline gate and verify-after-each text is byte-identical to the U3 result.

### U5: Ship agent Step 6.1 and Step 7: one close procedure (docs)

* Size M, complexity medium. Blocked by U4. File: `.github/agents/_ship.agent.md`.
* Replace Step 6.1 b (classify-close-path, binding, bound safe-close) with the command-routed pointer: a pre-flight probe
  (`autoharness shipment --help` exits 0, else HALT), classify and record, mutating command for CASCADE, skill sequence for
  SAFE_CLOSE, then 6.1 c-e. Step 7 consumes the skill's linked-deliberation disposition report instead of independently
  archiving source deliberations (adopt upstream `_ship.agent.md.tmpl` Step 7 wording). The closure artifact records
  `close_path` and `close_evidence` and runs `autoharness gate closure-evidence` before commit.
  **Keep** 6.0, `a0`, and the whole `a1` gate and its grant.
* Acceptance: (1) the 6.0, `a0`, and `a1` text is byte-identical to `main` (the PR shows the diff proving it); (2) no
  retired token remains; (3) the text requires per-closure operator approval before the mutating command and states dark
  mode does not waive it; (4) the Role Boundary still reserves the destructive close to the approved command; (5) tests pass.

### U6: Policy: P-015 and the P-010 grant (docs)

* Size S, complexity medium. Blocked by U2. File: `.github/policies/workflow-policies.md`.
* Replace the "fully-covered-root exception" with the upstream flat-manifest engine-inertness containment text, consistent
  with the P-015 title and statement; P-010: Ship may invoke the mutating `cascade-close` under operator approval and never
  `backlogit shipment ship` directly; add the version-table row.
* Acceptance: (1) `fully-covered-root` does not appear in the file; (2) P-005/P-010/P-015/P-017 cross-references stay
  valid; (3) the (a1) grant sentence is unchanged; (4) the U1 parity table has no unnamed gap; (5) the full suite is green,
  including the policy pins in `stage_branch_gate_test.go` (the `| 1.25.0` version row, `Corrected P-010`, and the
  "Commit or push directly to `main`" count of at least 2) and contract rows 20b/20c (the new text must not introduce
  `021-`, `022-`, or `018.` tokens those rows forbid); (6) the grant change is the one named in the operator checkpoint.

### U7: operational-closure skill and circuit-breaker row (docs)

* Size S, complexity medium. Blocked by U4. Files: `.github/skills/operational-closure/SKILL.md`,
  `.github/instructions/circuit-breaker.instructions.md`.
* Add the `close_path` / `close_evidence` keys, the `autoharness gate closure-evidence` validation step, the
  "Source artifact cleanup" section, and the mutating `cascade-close` supervision-budget row.
* Acceptance: `closure_source_artifact_cleanup` passes; `TestOperationalClosurePostMergeFilenameConformance` passes (it
  parses the `## Output` section); the circuit-breaker row states `--timeout` + 600 s.

### U8: routing re-render: orchestrator and stage agents (docs)

* Size S, complexity medium. Blocked by U2. Files: `.github/agents/_orchestrator.agent.md`, `.github/agents/_stage.agent.md`.
* Orchestrator: restore the per-step `context_tier` / `ROUTING_DEGRADED: context_tier` directives at the Stage and Ship
  invocation sites and the substantive "Propagate to inherited skills" content. Stage: replace the stale affirmative
  post-threshold retry directive with the current `ESCALATION_DEGRADED` fallback text; **add no stage escalation model**
  (D-CT-6 R1). Both from upstream template sections only; every other line preserved (the files are `user-modified`).
* Acceptance: `orchestrator_invocation_routing_directive`, `reload_propagation_directive` and `escalation_directive_present`
  pass; the diffs touch only those sections; Stage Step 1.9 and Role Boundary are byte-identical to `main`;
  `go test ./tests/integration/... -count=1` passes (Stage text pins).

### U9: routing re-render: ship agent routing block and manifest variables (config)

* Size S, complexity medium. Blocked by U5. Files: `.github/agents/_ship.agent.md` (routing/escalation block only),
  `.autoharness/harness-manifest.yaml` (route variables only). See "Justified deviations".
* Re-render the stale `*_ESCALATION_*`, `ESCALATION_*`, and `ORCHESTRATOR_REASONING_EFFORT` values from the current
  `.autoharness/config.yaml`. **`tuned_at` stays `null`:** a scoped hand re-render is not a tune, and the matrix records it.
* Acceptance: zero `ROUTE_VARIABLE_STALE` warnings in `verify-workspace`; `.autoharness/config.yaml` has no diff; **only
  manifest values that are derived from the unchanged config change** (no new key, no value that is not the config's
  resolved value); the Step 6.1 / Step 7 text from U5 is unchanged; tests pass.

### U10: add `name` to the 18 skill frontmatters (docs)

* Size S, complexity trivial. Blocked by U7 (same operational-closure file). 18 files, one mechanical transform.
* Add the missing `name:` key (the directory name) to the skills reported by `frontmatter_conformity`.
* Acceptance: `frontmatter_conformity` passes; each file's diff is exactly one added line; tests pass.

### U11: contract the pinned contract test (Go test)

* Size M, complexity medium. Blocked by U5, U6, U7. File: `tests/integration/ship_feature_completion_contract_test.go`.
* Posture: test-first. Remove the old-text alternatives from U2 so rows pin the new text only. Each row that pinned retired
  behavior (the U1 row map is authoritative; at last count rows 4, 21-27, 30 and 31-33, **excluding** guard row 24b, which
  keeps the four mode names that U4 retains) becomes an `assertAbsent` on the retired token,
  or is dropped with the row-map entry as the record; the new-form needles from U2 are proven against the real merged text.
  Replace `TestCascadeRoutesThroughSafeClose` with **`TestCascadeRoutesThroughCommand`** (CASCADE routes only through the
  mutating command; `a0`/`a1`/6.0 retained; no retired token). Keep the shared helpers and the name
  `TestShipFeatureCompletionContract33Rows` (update its row count in the doc comment if it changes), and update the file's top
  comment, which calls the behavioral tests authoritative.
* Acceptance: `go test ./tests/integration -run '^(TestShipFeatureCompletionContract33Rows|TestCascadeRoutesThroughCommand)$' -count=1 -v`
  prints `--- PASS` for both named tests; the PR shows a deliberately reverted U5 line failing it, and a reintroduced retired
  token failing the new `assertAbsent` rows; the full package passes.

### U12: fixture-backed `cascade-close --classify-only` smoke test (Go test)

* Size M, complexity medium. Blocked by U2. File: new `tests/integration/cascade_close_classify_only_smoke_test.go`.
* Build a disposable fixture under `t.TempDir()` with the real `backlogit` (determine empirically whether `autoharness`
  also needs a git repo or `.autoharness/config.yaml`; record the minimum). Run
  `autoharness shipment cascade-close --classify-only --shipment S --feature F --sha <40 hex> --workspace <abs> --json` with
  `exec.CommandContext` (timeout) and `cmd.Dir` set; assert exit code, `classifier_verdict`, `mutation_possible: no`, that the
  evidence path (`filepath.FromSlash`, `EvalSymlinks` or `Rel`) resolves under the fixture, and that a second plain
  `--classify-only` exits 2 (documented no-clobber). Helpers use **unique names** (`cascadeSmoke...`) so they never collide
  with the old harness. Windows: set `cmd.WaitDelay` so a killed child releases the `t.TempDir()` directory; compare paths
  after `EvalSymlinks` on **both** sides (8.3 short names); distinguish a context timeout from an `*exec.ExitError` when
  asserting exit codes. Skip with an explicit message when `autoharness shipment --help` fails or `backlogit` is absent;
  never use the mutating form; never assert timestamps. Justification: it is the only live check of the real tool against
  a fixture, even though CI cannot run it (recorded residual). If the fixture cannot be made hermetic within the unit's
  budget, stop, record a follow-up, and let U13 proceed anyway with the coverage loss stated in the PR.
* Acceptance: passes locally (`go test ./tests/integration -run '^<name>$' -count=1 -v` prints `--- PASS`, not `--- SKIP`,
  in an environment that has both tools); skips cleanly where `autoharness`, its `shipment` command, or `backlogit` is
  absent; writes nothing outside `t.TempDir()`.

### U13: retire the replica tests (Go test)

* Size S, complexity low. Blocked by U12 (complete, or its follow-up recorded). Files: delete
  `tests/integration/shipment_close_path_behavior_test.go` and `shipment_close_path_harness_test.go`.
* **Principle VII:** deleting files by terminal command is destructive. Ship runs this unit in Careful mode, obtains
  per-operation operator approval before the `git rm`, and halts the unit if approval is not available. Reversibility via git
  is not an exemption.
* Accepted coverage loss, stated in the PR: these tests exercised a test-only Go **replica** of the retired contract
  (Z1, B1-B6, Z2-Z4); they verified no production behavior. Real behavior now lives in the upstream command.
* Acceptance: `go build ./... && go vet ./tests/integration/... && go test ./tests/integration/... -count=1` passes; no
  other file referenced a deleted helper (compile proves it, and the shared assert helpers are untouched).

### U14: compound refresh (docs)

* Size S, complexity low. Blocked by U5. Files: `docs/compound/2026-10-11-classify-only-writer-evidence-before-cascade-close.md`,
  `docs/compound/2026-10-08-bound-cascade-close-procedure.md`.
* Add a dated "Superseded in part" note naming the deliberation: the 2026-10-11 "never use the mutating form" guidance is
  replaced by D-CT-1, and the 2026-10-08 binding procedure is retired. Historical text stays. These two entries are the ones
  that instruct the forbidden or retired procedure (the same contract surface as R1); the other compound entries still state
  a rule that stands.
* Acceptance: each file names the superseding decision; no other line changes.

### U15: end-to-end verification and residual-state record (docs)

* Size S, complexity medium. Blocked by every other unit. File: appends to `docs/design-docs/2026-10-11-closure-contract-surface-matrix.md`.
* Re-run `autoharness verify-workspace --workspace .`; `git grep` the retired tokens (allowlist: the immutable-history
  list, the two superseded compound entries, the deliberation, and this plan); run `go build ./... && go vet ./... && go test ./... -timeout 30m`
  (the `pwsh` build-script tests can take minutes each, beyond the default 10-minute limit); confirm no CI job runs `verify-workspace` as a gate.
* Acceptance: zero `ROUTE_VARIABLE_STALE` (with the 30 vs 10 reconciliation stated); **five of the six** targeted checks pass
  and the sixth, `escalation_route_resolution` (`ESCALATION_DEGRADED`), is the ONLY failing targeted check, exactly as
  D-CT-6 R1 predicts; zero retired tokens in the **live instruction surfaces** (`.github/**`, `scripts/**`, root docs) and in
  `docs/compound/**` outside the two superseded entries, where the allowlist is the immutable-history list plus the contract
  test's negative-test needles (`tests/integration/ship_feature_completion_contract_test.go`), the surface matrix doc, the
  deliberation, this plan, and the compound entries that merely mention a token in passing (named in the matrix); the Go suite passes; the PR body carries a P-005 disclosure of
  the one expected failing check; the doc states what the next shipment's closure must observe on its first use of the
  mutating command (probe, exit codes, evidence phases `pre_close -> invoking -> post_close`).

## Dependency Graph

```text
U1 -> U2 -> U3 -> U4 -> U5 -> U9
          U2 -> U6
          U2 -> U8
          U2 -> U12 -> U13
U4 -> U7 -> U10
U5, U6, U7 -> U11
U5 -> U14
U2..U14 -> U15
```

Acyclic. Text tasks (U3-U7) never precede the test expansion (U2); the strict contraction (U11) follows all text;
the smoke test (U12) precedes the deletion (U13).

Manifest order for Ship (parent first): the covering feature, then U1, U2, U3, U4, U5, U6, U7, U8, U9, U10, U11, U12, U13, U14, U15.

## Decisions and Rationale

Recorded in the deliberation (D-CT-1 .. D-CT-7). Summary: adopt the upstream command, retire the local prose binding and its
tests together (expand/contract, no red commit), keep the (a1) overlay, scoped two-phase re-render, routing left to the operator.

## Risks and Caveats

* The first real use of the mutating command is the next shipment's closure. An exit 5-8 halts to the operator by design.
* Scoped hand-editing against upstream templates can drop a local overlay line; U1's table and the byte-identity acceptance
  on `a0`/`a1`/6.0/Step 1.9 are the controls.
* The Stage and Ship agent definitions change under a running pipeline; Batch 2 and 3 Stage sessions reload them at session start.
* `escalation_route_resolution` remains FAIL until the operator decides R1; disclosed, not hidden.
* The editable upstream CLI reports `1.5.0`; version strings cannot distinguish it from the PyPI artifact. The probe
  (`autoharness shipment --help`) is the capability test.

## Plan Hardening Signals

Yes: closure and gate tooling (destructive mutating command), rule-text changes to a safety control, edits to managed
agent files, and test deletion.

Requires plan hardening: yes

## Runtime Verification and Closure

Verification is U15 plus the full Go suite and CI. Closure of this shipment is itself the first live exercise: Ship closes
it under the procedure installed at that moment. The closure artifact must carry `close_path` and `close_evidence`, and
`autoharness gate closure-evidence` must pass.

## Plan Hardening

### Hardening record

* **Required, and why:** the plan rewrites a safety control (shipment close), names a destructive command as the sanctioned
  executor, edits `user-modified` managed agents, and deletes tests.
* **Protected invariants (provably unchanged):** (I1) Ship step 6.0; (I2) `a0` topology gate; (I3) `a1` completion gate and
  grant; (I4) Stage Step 1.9 and Role Boundary text; (I5) operator approval for any destructive close, not waivable in dark
  mode; (I6) `.autoharness/config.yaml` unchanged; (I7) `.backlogit/` untouched by any unit; (I8) the shared assert helpers in
  the contract test file.
* **Context consulted:** `constitution.instructions.md` (I-XI), `workflow-policies.md` P-005/P-010/P-015/P-017/P-021,
  `backlogit.instructions.md`, the learnings above, `autoharness shipment --help` (exit-code contract), `autoharness
  verify-workspace` (read-only baseline), upstream templates (read-only), and the Go test sources.
* **Environment prechecks (Ship, before U1):** `autoharness shipment --help` exits 0; `backlogit --version` is 1.11.x;
  `git branch --show-current` starts with the Ship feature-branch prefix; the worktree is clean apart from carried stash state.
* **Operator checkpoints:** (1) **gate before claim:** approval and merge of the staging PR (D-CT-1 retirement of the 021-S
  control, and the U6 P-010 grant wording); (2) per-operation approval for the U13 file deletion (Principle VII) at
  execution time; (3) the per-closure approval for the mutating command is unchanged; (4) R1-R3 (stash **0A215E0B**) and the
  upstream asks (stash **A204A0BD**) are tracked outside the shipment.
* **Review-gate capability risks:** reviewer subagent dispatch and the anchor route (`gpt-6.1-sol`) may be unavailable.
  Plan review MUST emit literal `dispatch_mode:` and `decision:` markers and declare any degradation (P-012) before harvest.
* **Unresolved operator decisions that gate execution:** D-CT-1 approval through the staging PR (above). R1-R3 do not
  block the shipment: they affect only the sixth targeted check and comments, and are excluded by design.
* **Monitoring / rollback trigger:** any of I1-I8 failing its byte-identity or empty-diff check is a stop-and-revert condition
  for that unit; a first mutating close exiting 5-8 is a halt to the operator.

### Failure-mode analysis

| # | Failure mode | Impact | Mitigation | Verification |
|---|---|---|---|---|
| H1 | Scoped edit deletes `a1` or a role-boundary grant | Ship loses a safeguard silently | U1 `RETAIN` rows; byte-identity in U5 | U5 diff |
| H2 | Retired token left in one surface | The original contradiction persists | six-surface matrix; U15 grep with a closed allowlist | U15 |
| H3 | Pinned tests break mid-series | Red CI, blocked Ship | expand (U2) before text (U3-U7) before contract (U11); every task runs the package | per-task acceptance |
| H4 | `autoharness` lacks 192-F in some environment | New smoke test fails; Ship misroutes | probe + explicit skip; "CLI too old" HALT; CI residual recorded | U4, U5, U12 |
| H5 | Operator-owned `config.yaml` modified | Authority breach | U9 acceptance requires an empty diff | U9 |
| H6 | Mutating command used without approval | Destructive close | approval text kept, dark-mode non-waiver added | U4, U5 |
| H7 | Re-render "fixes" `escalation_route_resolution` by inventing a model | Silent routing decision | U8 forbids it | U8, U15 |
| H8 | Retention of source deliberations bypassed by Ship Step 7 | Shared deliberation archived early | adopt upstream disposition consumption | U4, U5 |
| H9 | U13 deletes a helper another test uses | Package fails to compile | unique helper names in U12; compile gate in U13 | U13 |

### ProposedAction / ActionRisk (documentation; strict-safety is not enabled)

| ProposedAction | ActionRisk | Control |
|---|---|---|
| Ship text names the mutating `autoharness shipment cascade-close` as the sanctioned CASCADE executor | destructive at use time, not at edit time | operator approval per closure; exit 5-8 halt rules |
| Delete two replica test files | **destructive** (Principle VII: deletes files; reversibility via git is not an exemption) | operator approval per operation, Careful mode; U12 first; full suite green before and after |
| Edit three managed agents and one manifest | reversible (git) | section-scoped diffs; byte-identity checks |

### Rollback

All changes are text or test files on a feature branch; `git revert` of the merge commit restores the prior contract,
including the retired tests. No backlog or engine state is touched by any unit.

<!-- plan-review-attempt: 3 (final verdict ADVISORY, no further FAIL) -->

## Plan Review

dispatch_mode: multi-agent
decision: ADVISORY

### Attempt 1 (FAIL, superseded; reviewers dispatched on the first draft)

dispatch mode was multi-agent for attempt 1 (Constitution, Scope Boundary, Architecture [anchor route gpt-6.1-sol], Go
reviewer as subagents; the Learnings Researcher ran in Step 1.8). the decision was FAIL (0 P0, 6 P1). Resolutions applied in
this revision:

| Finding | Resolution |
|---|---|
| P1 Test-first inverted / known-red intermediate versus Ship's per-task full-suite gate (Constitution, Architecture) | Expand/contract (U2 -> text -> U11); no red commit |
| P1 Retiring the 021-S control is beyond the operator condition; no checkpoint (Scope) | Operator checkpoint added (staging-PR review is the approval point for D-CT-1) |
| P1 SAFE_CLOSE residual unmeasured (Scope) | Named residual risk with the retained baseline gate and verify-after-each; U4 acceptance keeps them |
| P1 Version equality is not capability; CI pins PyPI 1.5.0 that predates 192-F (Architecture) | Probe + "CLI too old" HALT; CI residual recorded; upstream release/pin ask captured as a stash entry |
| P1 Linked-Deliberation Disposition and flat sets not adopted; Ship Step 7 bypasses retention (Architecture) | U4 and U5 adopt them explicitly (H8) |
| P1 U13 helper collision with the old harness; `-run` can pass vacuously; shared assert helpers; temp-dir statement (Go) | Unique helper names; `-v` plus named tests; helpers kept (I8); Principle IV statement reconciled |
| P2 U10, U9, U11/U13 granularity (Constitution) | "Justified deviations" section; U11/U12 now one concern each |
| P2 Constitution Check omitted VIII, V, IX, XI; approval testability; P-015 parity; CI and P-013.6; coverage gaps | Added; acceptance criteria added to U1, U4, U5, U6, U15; residual risks named |
| P2 16 tasks too many (Scope) | Reduced to 15 (U15 old compound refresh B and the 042-S gate run dropped; orchestrator+stage merged); remaining split is required by file/width rules |
| P2 `tuned_at` ambiguity; 30 vs 10 (Scope) | `tuned_at` stays `null`; reconciliation required in U1 and U15 |
| P2 exit 3 routing by mode; allowlist of historical plans (Architecture) | U4 routing by mode+exit code; immutable-history list enumerated |
| P2 U13 feasibility, no-clobber exit 2, ctx timeout, CI skip (Go) | The no-clobber exit 2 is in `autoharness shipment --help`; timeout, absolute paths, explicit skip added |

### Attempt 2 (FAIL, superseded; reviewers re-dispatched on the revised draft)

Dispatch mode was multi-agent. Constitution: ADVISORY; Scope: ADVISORY; Architecture (anchor route): FAIL (1 P1); Go: FAIL
(1 P1). The decision was FAIL. Resolutions applied in this revision:

| Finding | Resolution |
|---|---|
| P1 U3 removes the old skill contract before U4 adds its replacement, so the either-form tests have neither (Architecture) | Reordered: U3 ADDs the command-routed contract (additions only), U4 RETIREs the old (removals only); U2 migrates structural, absence, count, order and section-bounded assertions, not only presence |
| P1 U2 omitted `TestCascadeRoutesThroughSafeClose`; "invoke the cascade" absence matches "cascade-close" (Go) | U2 expands it, fixes the substring trap, and names it in the acceptance `-run` |
| P1 Principle VII: deleting files is destructive; "not a destructive-command class" was wrong (Constitution) | U13 requires per-operation operator approval and Careful mode; ProposedAction table corrected |
| P2 One helper insufficient for Exactly/AtLeast/Ordered/section rows; retired-behavior rows in U11; verbatim upstream needles; policy pins; row-30 section bound (Go) | U2 and U11 text expanded; U6 names `stage_branch_gate_test.go` pins and rows 20b/20c; needles cut verbatim from upstream and proven in U11 |
| P2 Red phase never observed (Constitution) | U2 temporary new-only red assertion; recorded Justified deviation |
| P2 U2/U11 granularity "No other deviation" inaccurate; U9 vs U8 inconsistency (Constitution) | Both recorded as Justified deviations; U8 independence from the manifest stated |
| P2 Careful mode scope; P-010 grant self-authorization; U14 traceability; Principle IV (Constitution) | Careful extended to U3-U9 and U13; grant wording named in the checkpoint; U14 tied to D-CT-1; conflict record added |
| P2 Checkpoint not enforced; Hardening said "none" (Scope) | "Ship MUST NOT claim until the staging PR is approved and merged"; Hardening reworded |
| P2 SAFE_CLOSE residual understated; no stash ID (Scope) | Residual states it is still agent-run prose; stash A204A0BD cited and verified before D10D3AFC is archived |
| P2 U12 weak justification; U13 ambiguity if U12 fails (Scope) | Justification stated; U13 proceeds with the loss stated if U12 cannot be hermetic |
| P3 U9 routing creep; Windows `WaitDelay`/`EvalSymlinks`/timeout vs ExitError; `go test -timeout 30m`; stale top comment | Added to U9, U12, U15, U11 |

### Attempt 3 (ADVISORY; no P0, no P1)

Dispatch mode was multi-agent (Architecture on the anchor route `gpt-6.1-sol`; Constitution, Scope Boundary, Go as subagents;
Learnings Researcher ran in Step 1.8 and its prior-art findings are carried in this plan; Agent-Native Parity not triggered
(no MCP tool or agent-facing action added); Security Lens not triggered (no auth, API surface, secrets, or trust-boundary
change; the destructive-command risk is covered by Principle VII and the hardening record)). The decision was ADVISORY.
All attempt-2 findings were verified resolved by every persona. The remaining P2/P3 items were fixed in place:

| Finding | Resolution |
|---|---|
| P2 U15 zero-token grep would fail on the contract test's own `assertAbsent` needles, the matrix doc, and passing mentions (Architecture, Go) | U15 scoped to live instruction surfaces with a named allowlist |
| P2 U11 flip list omitted rows 4, 25, 30 and wrongly included guard row 24b (Go) | List corrected; the U1 row map is authoritative |
| P3 U2 row-30 and row-10 either-form and the cascade-substring regexp not spelled out; U12 must show PASS not SKIP (Go) | Spelled out in U2 and U12 |
| P3 U12 five assertions; U10 Careful mode; Principle V "N/A" (Constitution) | Recorded as one scenario; Careful extended to U10; V rephrased |
| P3 Harvest record and marker (Scope) | Added at harvest |

Escalation protocol (consecutive planning failures) was not triggered: attempts 1 and 2 were FAIL, attempt 3 is not a FAIL.

**Authorization to harvest on an ADVISORY verdict (harvest Phase 1, step 4):** every P2 and P3 item above was fixed in
place; no P0/P1 remains. The Orchestrator's invocation delegated the staging decisions to Stage ("decide with sound
judgment, halt only on genuine ambiguity of scope/safety") while the operator is absent, and no genuine scope or safety
ambiguity remains open: the one authority-bearing choice (D-CT-1) is gated by the pre-claim staging-PR approval. Stage
records this authorization here and proceeds to harvest.

## Harvest record

Harvested 2026-10-11 by Stage after plan review `decision: ADVISORY` with recorded authorization (attempt 3 section).

* **Covering feature:** 053-F (queued). **Tasks (flat, no sub-epics):** U1..U15 = 053.001-T..053.015-T (same numbering).
* **Dependency edges (`blocks`, item depends on):** 053.002<-001; 003<-002; 004<-003; 005<-004; 006<-002; 007<-004; 008<-002;
  009<-005; 010<-007; 011<-005,006,007; 012<-002; 013<-012; 014<-005; 015<-008,009,010,011,013,014.
* **Shipment:** 043-S (queued). Manifest (16): 053-F, then 053.001-T .. 053.015-T in dependency order. Scope guard: only the
  IDs this harvest emitted were added. No upstream shipment dependency. **Later shipments (Batch 2, Batch 3) must record
  `backlogit dep add <later-S> 043-S` before handoff to Ship.**
* **Sizes** (structured, `size_source: agent`, ruleset `2h-rule-v1`): S x9, M x6, none above M. **Complexity** is recorded as labelled
  prose in each task's implementation notes because this workspace's task type defines no complexity field (backlogit rejects
  `update --complexity`: "artifact type task does not define a complexity field"). Degradation flagged, not silent.
* **Consumed stash (archived):** D10D3AFC, A83C398E. **New active stash (Stage-owned captures):** A204A0BD (upstream asks),
  0A215E0B (operator routing decisions R1-R3). **Annotated, still active:** 9F824B64 (re-triage note).

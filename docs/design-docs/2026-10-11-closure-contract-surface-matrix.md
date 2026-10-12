---
title: "Closure contract surface matrix (043-S, plan unit U1)"
date: 2026-10-12
status: "draft"
source: "docs/plans/2026-10-11-intercom-go-closure-tooling-batch-1-plan.md"
backlog_task: "053.001-T"
---

This matrix is the U1 deliverable of the closure tooling Batch 1 plan. It records, for every live surface that mentions a retired close-path token, the disposition, the owning unit (U2 to U14), the upstream template counterpart, and the pinning test. It changes no surface. Units U2 to U15 use it as their source of truth.

## Scope and method

* Scope query, run exactly: `git --no-pager grep -l -i -E "classify-close-path|classification_binding|fully-covered-root"` with the eight immutable-history exclusions listed below. It returned 12 files.
* Line numbers come from the same query run with `-n`, read against the working tree on 2026-10-12.
* Upstream counterparts are read-only from the autoharness checkout (`templates/` and `src/`).
* Baseline and harness facts are in sections 5 and 6.

### Excluded immutable history

These paths keep retired tokens by design. They are excluded from the matrix, were not edited, and must not be edited by any unit:

* `docs/closure/`
* `docs/memory/`
* `docs/archive/`
* `docs/plans/`
* `docs/reviews/`
* `docs/decisions/`
* `.backlogit/`
* `.autoharness/continuous-learning/`

`.autoharness/config.yaml` is operator-owned, is not returned by the scope query, and must show no diff at any unit (U9 acceptance).

## 1. Surface matrix

### 1a. Files returned by the scope query

| Surface | Retired or changed content (grep -n lines) | Disposition | Owner | Upstream counterpart | Pinning test |
|---|---|---|---|---|---|
| `.github/agents/_ship.agent.md` | Retire L822-863: step 6.1 b close-path selection by the classify boundary (L822, L826), binding and bound safe-close routing (L828, L830, L832, L835, L843), binding carried into the closure call (L863). Changed: Step 7 item at L902 (source deliberation archival). | retire L822-863; adopt command routing by mode and exit code, `selected_close_path`, and the linked-deliberation disposition report in Step 7 | U5 (U9 also re-renders the routing block) | `templates/agents/_ship.agent.md.tmpl` (cascade-close routing at L828, L838, L879; closure-evidence contract at L932) | `ship_feature_completion_contract_test.go` rows 4, 10, 19, 25, 26, 27, 31, 32 and guards 7, 8, 9a-9d |
| `.github/policies/workflow-policies.md` | Retire L440: heading and body of the VERIFIED FULLY-COVERED-ROOT EXCEPTION, which runs to the end of P-015 before the P-016 heading at L456. Retire the token in version-table rows L773 (1.19.0) and L775 (1.21.0), which cite the exception (see N4). Keep L245 (P-010 Ship grant). | retire the exception block and adopt the flat-manifest containment text; keep L245 | U6 | `templates/policies/workflow-policies.md.tmpl` (P-015 at L497; history rows L880 and L882 still carry the token) | `ship_feature_completion_contract_test.go` rows 20a, 20b, 20c, 28 and guard `018.`; `stage_branch_gate_test.go` (reads the policy at L36; pins named in plan U6 acceptance 5) |
| `.github/skills/shipment-reconcile/SKILL.md` | Retire L3, L14, L32 (description and summary naming the fully-covered-root case); L56, L57, L60 (mode enum and `classification_binding` input); L255 (classify-close-path read-only rule); L363 to L443 (Classify-Close-Path Mode section: verdict fields, condition table, canonical binding, report path, independent invocability, incl. L377, L381, L402, L405, L432, L439); L447, L457 (fully-covered-root case selection); L595 to L597 (bound revalidation, unbound refusal, migration note); L711 (Cascade Close Sub-Procedure heading); L1111, L1170 (related-artifact and quality text). Not returned by the scope query but retired: L383 (`FULLY_COVERED_ROOT` row, see N5). | retire the classify-close-path section, the binding contract, the unbound refusal, and the fully-covered-root framing (U4); adopt cascade-close routing, `selected_close_path`, flat sets, and linked-deliberation disposition (U3 adds); keep the SAFE_CLOSE sequence, protected-set baseline gate, verify-after-each, Pre-Mode, Post-Mode, detect-mixed-role | U3 (adopt) and U4 (retire) | `templates/skills/shipment-reconcile/SKILL.md.tmpl` (cascade-close, close_path, closure-evidence present; no classify-close-path token) | `ship_feature_completion_contract_test.go` rows 21a, 21b, 21c, 22a, 22b, 22c, 23a to 23d, 24a, 30, 33; guard row 24b keeps the four mode names; `shipment_close_path_harness_test.go` computes a digest of this file (computed, not pinned to a value) |
| `docs/compound/2026-05-07-backlogit-shipment-status-constraints.md` | L54, L55, L58 (instructs classify-close-path, then safe-close with `CLASSIFICATION_BINDING`, then the fully-covered-root exception); L94 (flow). | keep for history; open residual N6 (instructs the retired flow and is not in the two-entry U14 list) | U14 | none (local compound note) | none |
| `docs/compound/2026-09-30-shipment-reconcile-skill-bypass-manual-substitution-risk.md` | L12, L17, L38, L67, L84, L85, L114, L115, L140, L154, L210, L212, L221 (incident record of the manual binding substitution in 036-S). | keep (incident narrative; mention in passing, allowlisted for U15) | U14 | none | none |
| `docs/compound/2026-10-02-backlogit-stale-binary-on-path-breaks-index.md` | L37 (binding embeds the engine line). | keep (mention in passing) | U14 | none | none |
| `docs/compound/2026-10-08-bound-cascade-close-procedure.md` | L17 (evidence path naming classify-close-path records); L31 (step 3, classify-close-path read-only). | retire in part: add a dated "Superseded in part" note naming D-CT-1; body kept | U14 | none | none |
| `docs/compound/workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md` | L123 (mention: the plan adds classify-close-path). | keep (mention in passing; the learning behind U1) | U14 | none | none |
| `docs/compound/workflow-issues/ship-step6-post-merge-branch-ordering-2026-09-19.md` | L27 (mention in the step 6 ordering list). | keep (mention in passing; step 6.0 stays first) | U14 | none | none |
| `tests/integration/ship_feature_completion_contract_test.go` | L138 (row 4), L229 (21a), L237 (21c), L241 (22a), L268 (24a), L283 (26), L287 (27), L303 (row 30 section constant), L319 (row 31). | adopt: expand (U2), then keep the shared helpers and contract the old needles (U11) | U2 (expand) and U11 (contract) | none | itself; `TestCascadeRoutesThroughSafeClose` (same file, L359) |
| `tests/integration/shipment_close_path_behavior_test.go` | L12, L16, L33, L117, L212, L224, L508, L517, L532 (behavioral replica of the retired contract). | retire, after the U12 smoke test exists | U13 | none | none (compile gate in U13) |
| `tests/integration/shipment_close_path_harness_test.go` | L24, L31, L34, L87, L88, L122, L364, L480, L485, L658 (replica classify-close-path and `computeBinding`). Not in scope-query output: L706-715 read the skill to compute a digest (no fixed value is asserted). | retire, with the behavior test | U13 | none | none (self) |

### 1b. Plan-named surfaces outside the scope query

These surfaces carry no retired token in the scope-query sense. The plan names them for units U7 to U10 and U14, so they are recorded here for ownership.

| Surface | Retired or changed content | Disposition | Owner | Upstream counterpart | Pinning test |
|---|---|---|---|---|---|
| `docs/compound/2026-10-11-classify-only-writer-evidence-before-cascade-close.md` | Matched only by the broader token set (lines 12, 19, 20 and others): writer-form evidence and "do not use the mutating `cascade-close` from Ship". | retire in part: add a "Superseded in part" note naming D-CT-1; body kept | U14 | none | none (harness U14) |
| `.github/skills/operational-closure/SKILL.md` | No retired token. Missing per `closure_source_artifact_cleanup`: "Source artifact cleanup", `source_stash_id`, `source_deliberation_id`. To add: `close_path`, `close_evidence`, closure-evidence gate. | adopt | U7 | `templates/skills/operational-closure/SKILL.md.tmpl` (cascade-close, close_path, closure-evidence, Source artifact cleanup present) | verify-workspace `closure_source_artifact_cleanup` (ok false); `tests/integration/operational_closure_post_merge_filename_test.go` |
| `.github/instructions/circuit-breaker.instructions.md` | No retired token. To add: the mutating cascade-close supervision-budget row (timeout and 600 s). | adopt | U7 | `templates/instructions/circuit-breaker.instructions.md.tmpl` (cascade-close present) | harness U7 (`logs/043s-harness.ps1`) |
| `.github/agents/_orchestrator.agent.md` and `.github/agents/_stage.agent.md` | No retired token. Failing checks: `orchestrator_invocation_routing_directive`, `reload_propagation_directive`, `escalation_directive_present`. Two P1 hardcoded-path portability findings in the orchestrator (see D2). | retire the stale directives; adopt upstream context_tier, propagation and fallback sections; keep Stage Step 1.9 and Role Boundary | U8 | `templates/agents/_orchestrator.agent.md.tmpl`; `templates/agents/_stage.agent.md.tmpl` | verify-workspace checks named above; `tests/integration/stage_branch_gate_test.go`; `tests/integration/staging_gate_actor_contract_test.go` (per plan U8) |
| `.autoharness/harness-manifest.yaml` | Thirty route-family keys; ten disagree with config (section 5). Tool-managed; U9 changes route values only. | retire the ten stale values (scoped re-render); `tuned_at` stays null | U9 | none (rendered from the operator config) | verify-workspace `route-variable-stale` (10 warnings) |
| 18 `.github/skills/*/SKILL.md` frontmatters named by `frontmatter_conformity` | No retired token. Each lacks the required `name` key. | adopt (one added line each) | U10 | `templates/skills/<name>/SKILL.md.tmpl` | verify-workspace `frontmatter_conformity` (ok false, 18 errors) |

## 2. RETAIN group

These overlays are local to this repository. The plan requires them to stay byte-identical or unchanged. Each row is a guard: the owning unit must not change the listed lines.

| Surface and anchor | Lines | Disposition | Owner | Upstream counterpart | Pinning test |
|---|---|---|---|---|---|
| (1) Ship Step 6.1 a1 "Covering-feature completion gate" (021-S) in `.github/agents/_ship.agent.md` | 777 (heading); 790, 795, 797, 799, 803 (S0, S1, S2, ALREADY_DONE, HALT) | RETAIN: local overlay | U5 (must not change) | none (no a1 text in `templates/agents/_ship.agent.md.tmpl`) | contract rows 3, 12, 13, 16a, 16b |
| (2) P-010 Ship grant wording in `.github/policies/workflow-policies.md`, with the same grant in the Ship Role Boundary | Policy L245 (inside P-010 at L211); Ship L38 (Role Boundary Backlog row) | RETAIN: local overlay | U6 (policy) and U5 (Role Boundary), must not change | none (grant absent from both templates) | contract rows 1, 2a, 2b, 2c, 2d, 20a, 28 |
| (3) Intake reconciliation relocation (021-S) in Ship Step 0.5 item 3b | 202, 205, 206, 207, 210 (Step 0.5 heading at L153); claim at L269 | RETAIN: local overlay (pre-claim placement) | U5 (must not change) | `templates/agents/_ship.agent.md.tmpl` item 6 (L276) sits after the claim (L274-275); the local pre-claim placement is a deliberate divergence | contract rows 14, 17, 18, 29 |
| Ship Step 6.0 "Post-Merge Branch Protocol" with the pre-self-close context reload | 729 (heading); 761-764 (reload clause) | RETAIN | U5 (must not change) | `templates/agents/_ship.agent.md.tmpl` L758 | contract rows 5, 6, 11 |
| a0 "TOPOLOGY_GATE: lifecycle" in Ship Step 6.1 | 771-775 | RETAIN | U5 (must not change) | `templates/agents/_ship.agent.md.tmpl` L799 | contract row 3 (ordering anchor) |

## 3. Contract-test row map (33 numbered rows)

Source: `tests/integration/ship_feature_completion_contract_test.go`. The map records each numbered row, its assertion kind, the surface and lines it checks, and the unit that changes its pinned text. "Retained" means the unit must not change the text. Rows 4, 21a to 27, 30 and 31 to 33 pin retired behavior. U11 turns them into absence checks, and guard row 24b is kept.

| Row | Assertion kind | Surface checked (lines) | Owner | Note |
|---|---|---|---|---|
| 1 | presence | Ship L38 (Role Boundary grant) | U5 guard (retained) | RETAIN (2) |
| 2a | presence | Ship L38 (conjunctive, fail-closed) | U5 guard (retained) | RETAIN (2) |
| 2b | presence | Ship L38 (no live descendant at any depth) | U5 guard (retained) | RETAIN (2) |
| 2c | presence | Ship L38 (set-equal wording; also policy L245) | U5 guard (retained) | RETAIN (2) |
| 2d | presence | Ship L38 (scoping clause; also policy L245) | U5 guard (retained) | RETAIN (2) |
| 3 | presence plus two Ordered checks | Ship L771 (a0) before L777 (a1) before L812 (a) | U5 guard (retained) | RETAIN (1) and a0 |
| 4 | presence | Ship L822 (classify-close-path delegation) | U5 | retired; becomes absence in U11 |
| 5 | presence | Ship L762 (re-read P-015 before self-close) | U5 guard (retained) | Step 6.0 |
| 6 | presence | Ship L764 (reload cannot widen authority) | U5 guard (retained) | Step 6.0 |
| 7 | absence | Ship (retired predicate "it is fully covered (every one of its children,"); absent today | U5 | guard against reintroduction |
| 8 | absence | Ship (token TASK_ONLY_FINALIZE); absent today | U5 guard | guard |
| 9a | absence | Ship (retired CS4 predicate "permitted only when, for every feature member"); absent today | U5 | guard against reintroduction |
| 9b | absence | Ship (retired "enumerates to zero children" predicate); absent today | U5 | guard |
| 9c | absence | Ship (retired "The manifest must contain nothing beyond the"); absent today | U5 | guard |
| 9d | absence | Ship (retired "qualification is never per-member"); absent today | U5 | guard |
| 10 | Exactly once plus presence | Ship L852 (`<shipment_id> --status shipped`, once); L850 (authoritative close sequence) | U5 | the once-only count must hold |
| 11 | presence | Ship L764 (reload scoped to Role Boundary changes) | U5 guard (retained) | Step 6.0 |
| 12 | presence (two) | Ship L795 (A1_NOT_APPLICABLE); L803 (HALT naming the unmet condition) | U5 guard (retained) | RETAIN (1) |
| 13 | presence | Ship L799 (A1_ALREADY_DONE) | U5 guard (retained) | RETAIN (1) |
| 14 | Exactly once plus presence | Ship L206 (once), L205 (intake delegation) | U5 guard (retained) | RETAIN (3) |
| 15 | Exactly once plus presence | Ship L817 (once), L815 (pre-archive gate, Step 6.1 a) | U5 guard (retained) | Step 6.1 a is outside the U5 replacement |
| 16a | presence | Ship L790 (S0, evaluated before n) | U5 guard (retained) | RETAIN (1) |
| 16b | presence (two) | Ship L797 (A1_MULTIPLE_FEATURE_MEMBERS; halt with no feature mutation) | U5 guard (retained) | RETAIN (1) |
| 17 | AtLeast 2 | Ship L207 and L817 ("scans for orphan items") | U5 guard (retained) | RETAIN (3) and 6.1 a |
| 18 | Exactly once | Ship L210 (scope note 139-F/139.001-T) | U5 guard (retained) | RETAIN (3) |
| 19 | absence plus two presence | Ship L855 ("proves the protected set on the safe-close path"; "no protected set by construction") | U5 | Step 6.1 b |
| 20a | presence | Policy L245 (P-010 grant) | U6 guard (retained) | RETAIN (2) |
| 20b | absence | Policy (token `018.`) | U6 | plan-assigned (see N2) |
| 20c | absence (two) | Policy (tokens `021-` and `022-`) | U6 | policy |
| 21a | presence | Skill L60 ("digest returned by `mode: classify-close-path`") | U4 (retire); U3 adds the new-form input row | retired |
| 21b | presence (three) | Skill L375-376 (CLOSE_PATH_VERDICT CASCADE, SAFE_CLOSE, BLOCK) | U4 (retire) | retired |
| 21c | AtLeast 2 | Skill, classify-close-path occurrences (e.g. L56, L57, L255, L363, L381, L432) | U4 (count reaches zero) | retired |
| 22a | presence (heading) | Skill L363 (### Classify-Close-Path Mode) | U4 | retired |
| 22b | presence | Skill L367 (no archive, no status transition, no record mutation) | U4 | retired |
| 22c | presence | Skill L392 (fail-closed BLOCK wording) | U4 | retired |
| 23a | presence (two) | Skill L595 (REFUSED; the bound verdict alone selects the branch) | U4 | retired |
| 23b | absence plus presence | Skill L602 ("bound verdict `SAFE_CLOSE` only" present); default-selection wording absent today | U4 | retired |
| 23c | presence (two) | Skill L596 (CASCADE_UNBOUND; zero archive, zero status transition) | U4 | retired |
| 23d | presence | Skill L595 (CLASSIFICATION_INVALID) | U4 | retired |
| 24a | presence (two) | Skill L595 (CLASSIFICATION_DRIFT); L255 ("strictly READ-ONLY") | U4 | retired; see N8 |
| 24b | AtLeast 1 each (four) | Skill: mode pre and post (L9), safe-close (L10), detect-mixed-role (L38) | U4 guard (retained) | guard; keeps the four mode names |
| 25 | presence | Ship L830 ("HALT on any result token the installed classification did not name") | U5 | retired |
| 26 | absence plus presence | Ship: old "in place of the safe-close sequence above" absent today; "carrying the returned CLASSIFICATION_BINDING" at L830, L832 | U5 | retired |
| 27 | presence | Ship L835 ("MUST NOT invoke `mode: safe-close` without a `classification_binding`") | U5 | retired |
| 28 | presence (two) | Policy L245 (set-equal wording; scoped wording) | U6 guard (retained) | RETAIN (2) |
| 29 | presence plus Ordered | Ship L202 (intake check) before L269 (claim) | U5 guard (retained) | RETAIN (3) |
| 30 | section-bounded presence | Skill block L363 to L443 contains "CLOSE_PATH_VERDICT: CASCADE" (L375) | U4 (section removed) | retired |
| 31 | absence plus presence (two) | Ship: old "skill with mode: safe-close, shipment_id, and the" absent today; L843 and L844 present | U5 | retired |
| 32 | presence (two) | Ship L836 (RECONCILE_FAIL_CASCADE_UNBOUND); L838 (HALT with no mutation performed) | U5 | retired |
| 33 | presence (three) | Skill L596-597 (rejected, not defaulted; refused, never served on a legacy path; no deprecation window) | U4 | retired |
| guard | absence | Ship (token `018.`) | U6 per plan (see N2) | guard |

Discrepancies with the requested mapping:

* The test is documented as "33-row", but it has 33 numbered rows and 51 leaf subtests (letter variants plus the separate guard). See N1.
* No row in this test pins `operational-closure`, `circuit-breaker`, orchestrator or stage text. The U7 and U8 checks are verify-workspace checks and other tests (see N3).

## 4. P-015 safeguard parity table

Sources: the P-015 section of `.github/policies/workflow-policies.md` (L414-455); the fully-covered-root material in `.github/skills/shipment-reconcile/SKILL.md` (Cascade Close Sub-Procedure, L711-931; Safe-Close steps at L611-705); Ship Step 6.1 b (L822-872). Upstream evidence is read-only: `autoharness shipment --help` output, and `src/autoharness/gates/shipment_closure.py`, `src/autoharness/gates/cascade_evidence.py`, `src/autoharness/shipment_close/command.py`, `postclose.py`, and `persist.py`.

| # | 155-S safeguard (source) | Upstream behavior that covers it (source) | Status |
|---|---|---|---|
| 1 | Feature members must be roots (P-015 item 1) | Any feature member with a `parent_id` returns SAFE_CLOSE (`classify_shipment_close_path`, shipment_closure.py L381) | COVERED |
| 2 | Full-depth descendant walk, not direct children (P-015 item 1, PR #407) | Descendants enumerated over the whole queue and archive index (`_enumerate_descendants`, shipment_closure.py L363); module docstring requires every reachable out-of-manifest descendant to be checked | COVERED |
| 3 | Childlessness verified, never inferred (P-015 item 2) | Scan failure or ambiguous ids return SAFE_CLOSE; symlink and reparse-point backlog paths fail closed (shipment_closure.py L109-144, L220-254, L293-316) | COVERED |
| 4 | Childless root is terminal (P-015 item 3) | A manifest member that declares the feature as parent returns SAFE_CLOSE (`declared_as_parent_by`) | COVERED |
| 5 | Manifest holds nothing beyond qualifying roots and their descendants (P-015 item 4) | Unaccounted manifest members return SAFE_CLOSE (`extras`) | COVERED |
| 6 | All-or-nothing across feature members (P-015 item 5) | The first failing member returns SAFE_CLOSE for the whole manifest | COVERED |
| 7 | Full descendant coverage by the manifest (P-015 item 1, "fully covered") | Replaced by containment (INV-6): every out-of-manifest descendant must have parsed status exactly `archived` (`_is_engine_inert`, shipment_closure.py L203). The policy text must follow this (U6) | COVERED (changed predicate) |
| 8 | Pre-archived members tolerated; declared status is the only "truly archived" authority (P-015 item 7; skill L720-748) | `_is_engine_inert` and `is_declared_archived` use parsed declared status only; location is not an input (cascade_evidence.py L770-781) | COVERED |
| 9 | Shipment record and qualifying features always required (P-015 item 7; skill step 3) | `required_ids` = {S} plus qualifying features plus manifest items not truly archived (`compute_flat_sets`, cascade_evidence.py L789-811) | COVERED |
| 10 | Two independently failing set checks, never merged (skill step 3; B57F9E24) | `compute_set_differences` returns unexpected and missing as separate tuples (cascade_evidence.py L813-824); both recorded in evidence | COVERED |
| 11 | Validated linked deliberations of qualifying features in `allowed_ids` (skill step 3 extension) | Under verified 1.11 they stay outside both sets and go through the INV-12 disposition planner and drift check (shipment_closure.py module docstring; `compute_linked_deliberation_disposition`; command.py L423). Mechanism differs: N12 | NAMED RESIDUAL |
| 12 | `returned_ids` must be empty; HALT, P-005, no retry (skill step 2) | Failure recorded at postclose.py L598-599; exit 5; exits 5 to 8 forbid re-run, direct `shipment ship`, and SAFE_CLOSE substitution (help text) | COVERED |
| 13 | `parent_id` preserved against the pre-close snapshot (skill step 4) | Re-read in command.py `_reread` (L360-385); postcondition `parent_id_preserved` (cascade_evidence.py L1251) | COVERED |
| 14 | Cascade report with verdict, sets, both differences and result (skill step 5) | Evidence record `docs/closure/evidence/{S}-{F}-close-evidence.json` with pre_close, invoking and post_close phases (cascade_evidence.py L696-715; help text). Ship records `close_path` and `close_evidence` (U5, U7) | COVERED (different artifact) |
| 15 | Gate decision requires every check to pass (skill step 6) | Exit 0 only when all postconditions pass; `postcondition_verdict` must be pass (cascade_evidence.py L1255-1257) | COVERED |
| 16 | One engine invocation; no per-item archive loop (skill sub-procedure opener) | One fixed `shipment ship` argv through `run_bounded` (command.py step 8, L451-465) | COVERED |
| 17 | No SAFE_CLOSE substitution after CASCADE (skill no-substitution rule, L791; step 2) | Exit 4 "never SAFE_CLOSE"; exits 5 to 8 forbid substitution (help text) | COVERED |
| 18 | Bound classification revalidated before any mutation (155-S Step 0 binding; retired in U4) | Whole snapshot recomputed and compared with the durable record, ignoring only `run_id` and `captured_at`, plus a binary re-hash; exit 4 with nothing invoked (command.py `_revalidate`, L308-329). The SHA-256 binding digest is not carried forward | COVERED (replaced mechanism) |
| 19 | Engine semantics checked before cascade (155-S relies on backlogit 1.8.0 source) | `assess_cascade_engine_semantics` returns VERIFIED only on minor line 1.11 with matching probe and invoke surfaces; `select_close_path` requires VERIFIED (shipment_closure.py L560-800) | COVERED (stricter); N13 |
| 20 | Operator approval for each destructive close, not waived in dark mode (Ship 6.1 b; skill) | Help text: same operator approval as a direct `shipment ship` call. No code-level approval gate | COVERED (text only; U3 and U5 wording) |
| 21 | Requeue and detach of unshipped descendants with `parent_id` cleared (Ship 6.1 b hazard) | No requeue or detach path. Any non-inert out-of-manifest descendant returns SAFE_CLOSE, so CASCADE cannot reach this hazard | COVERED (hazard unreachable) |
| 22 | Snapshot and restore of a non-member covering feature (Ship 6.1 b; skill sub-procedure text) | No restore. A postcondition failure exits 5 with operator review and forbids committing the backlog root | NAMED GAP G1 |
| 23 | Git-restore of queue and archive on cascade detection (P-015 Violation Action 1; Ship 6.1 b) | No automatic `git restore`. Exits 5 to 8 require operator review; nothing is auto-pruned | NAMED GAP G1 (restore step only) |
| 24 | Protected-set baseline gate and verify-after-each (SAFE_CLOSE only; skill L611-682) | No protected-set concept in the CLI; the SAFE_CLOSE sequence stays agent-run prose | NAMED GAP G2 (tracked upstream as A204A0BD) |
| 25 | Archive-provenance validation before close (skill L174-184 malformed-provenance; Ship 6.1 b "any provenance ambiguity") | Classifier uses declared status only; `archived_status` and `archived_from` are not classifier inputs. Post-close check requires `archived_status: shipped` on the shipment record (`is_archived_shipped`, cascade_evidence.py L844-851) | NAMED GAP G3 |
| 26 | BLOCK verdicts for ambiguous, missing, incomplete, mixed or erroring inputs (skill classify table, L381-392) | `ClosePath` has two members, SAFE_CLOSE and CASCADE (shipment_closure.py L149-153). These cases return SAFE_CLOSE with a reason | NAMED GAP G4 |
| 27 | Single-writer file lock on the shipment record (skill L239-241, L268-269) | O_EXCL pair lock at `.autoharness/gates/cascade-close/{S}-{F}.lock` (persist.py L89, L462-491); stale locks are operator-only; no shipment-record lock | NAMED RESIDUAL N9 |
| 28 | Archive ordering inside the engine (Ship 6.1 b "outside the manifest-scoped ordering") | Sets and postconditions only; order is not evidenced. 155-S also relied on the engine for order | NAMED GAP G5 |
| 29 | P-005 broadcast naming cascaded IDs (P-015 Violation Action 3) | `failures[]` in `--json` output; no broadcast. Agent-side duty, not a CLI gap | COVERED (agent-side) |

Upstream additions with no 155-S counterpart: symlink and reparse refusal on backlog reads (shipment_closure.py L109-144); `--timeout` from 30 to 3600 seconds, default 1800 (help text); plain `--classify-only` no-clobber exit 2 (help text); bounded runner with a fixed argv; a probe-based engine gate.

Named gaps:

* G1 (rows 22, 23): no automated restore of a non-member covering feature or of queue and archive after cascade detection. Recovery is operator-only through exits 5 to 8. Decision for U5: accept the operator-only recovery, or add an operator-run restore step to the Ship text.
* G2 (row 24): the SAFE_CLOSE protected-set gate stays agent-run prose. This is already a plan residual (stash A204A0BD).
* G3 (row 25): pre-close archive provenance (`malformed-provenance`) is not a classifier input upstream. Decision for U3: keep the 155-S provenance check in the skill text or drop it with a recorded reason.
* G4 (row 26): upstream collapses BLOCK-class refusals into SAFE_CLOSE. The U3 and U5 routing must state what the SAFE_CLOSE sequence does with those inputs before it runs them.
* G5 (row 28): archive ordering is not evidenced upstream. This is also true of 155-S, so it is recorded, not introduced.

## 5. Baseline

Copied verbatim from `logs/043s-baseline-verify-workspace.txt`:

```text
Strict schema blockers: 0
Blockers: 0
Warnings: 12
```

Read from `.autoharness/staging/verify-workspace-report.json` with `Get-Content -Raw | ConvertFrom-Json`. This unit did not regenerate the report.

* `strict_schema_blockers`: empty array (0). `blockers`: 0. `warnings`: 12 entries. `unresolved`: 1.
* Warnings: 10 `route-variable-stale` entries and 2 `portability-finding` entries. Both portability findings are in `.github/agents/_orchestrator.agent.md` (rules `hardcoded_user_home` and `hardcoded_ah_home`, severity P1). No warning entry is a frontmatter entry.
* Stale-variable count: 10 distinct variables and 10 warning instances. The markdown report has 10 `ROUTE_VARIABLE_STALE` lines. The ten variables are ORCHESTRATOR_REASONING_EFFORT, LEGACY_ESCALATION_FAMILY, LEGACY_ESCALATION_PROVIDER, LEGACY_ESCALATION_REASONING_EFFORT, STAGE_ESCALATION_FAMILY, STAGE_ESCALATION_PROVIDER, STAGE_ESCALATION_REASONING_EFFORT, SHIP_ESCALATION_REASONING_EFFORT, ESCALATION_FAMILY, and ESCALATION_REASONING_EFFORT.
* The stash figure of 30 is reconciled as a count-method difference. `.autoharness/harness-manifest.yaml` holds exactly 30 route-family keys: ten families of three keys (ALT_REVIEW, ALT_DOC_REVIEW, ORCHESTRATOR, STAGE, SHIP, LEGACY_ESCALATION, STAGE_ESCALATION, SHIP_ESCALATION, ESCALATION, ANCHOR_REVIEW). Ten of the thirty disagree with config and are the flagged instances. This match is an inference, because the stash does not state its derivation (N10).
* Frontmatter: `frontmatter_conformity` is a failing targeted check with 18 `FM_MISSING_REQUIRED name` errors, not a warning (N11).

Six failing targeted checks (each `ok` is false):

1. `frontmatter_conformity`: 18 `FM_MISSING_REQUIRED name` errors, one per skill.
2. `closure_source_artifact_cleanup`: `operational-closure/SKILL.md` lacks "Source artifact cleanup", `source_stash_id`, and `source_deliberation_id`.
3. `orchestrator_invocation_routing_directive`: the Stage and Ship invocation sites lack the per-step `context_tier` and `ROUTING_DEGRADED` directives in their bounded window.
4. `escalation_route_resolution`: role `stage` resolves to its own role route (`ESCALATION_DEGRADED`). Operator decision R1 (stash 0A215E0B) governs it; the plan expects this to be the one remaining failure.
5. `escalation_directive_present`: `_stage.agent.md` contains a stale affirmative post-threshold retry directive.
6. `reload_propagation_directive`: `_orchestrator.agent.md` declares "Propagate to inherited skills" without substantive content.

## 6. Auto-tune preview

Result: non-applicable: no read-only preview command found.

Help text checked: `autoharness --help`. Listed commands: home, version, verify-workspace (writes only `.autoharness/staging/`), gate check, gate size (writes the size back to the task), gate copilot-review, gate pipeline-topology, gate dag-readiness, gate closure-evidence, shipment cascade-close (mutating and `--classify-only` forms), telemetry begin, telemetry record, harness resolve, eval, setup-vscode, setup-copilot-cli, setup-claude, setup-codex, and help. No tune or preview subcommand is listed. No preview was run, and no write or apply command was run by this unit.

## 7. Named residuals and discrepancies

* N1: The test is documented as "33-row", but it has 33 numbered rows and 51 leaf subtests (including the `guard_no_018_dot_in_ship` guard). The harness check "doc has 33-row map" matches the substring "33" only (D1).
* N2: The `018.` guard reads the Ship file, but the plan assigns it to U6 (policy). The map records U6 per plan. U5 also edits the Ship file, so U5 must keep the guard passing.
* N3: No row in the 33-row map pins `operational-closure`, `circuit-breaker`, orchestrator or stage text. U7 and U8 are verified by verify-workspace checks and by the tests named in section 1b.
* N4: Policy version-table rows L773 and L775 contain the token `fully-covered-root`. U6 acceptance 1 requires the token to be absent from the file, so those rows must be rewritten or the acceptance must be amended. The upstream policy template keeps the token in history rows L880 and L882, and the upstream Ship template L855 names the token only as withdrawn text. Adopting upstream text verbatim could reintroduce the token into live text.
* N5: The scope query does not match underscore-form tokens. Skill L383 uses `FULLY_COVERED_ROOT`. The U15 zero-token scan should also cover `CLOSE_PATH_VERDICT`, `FULLY_COVERED_ROOT`, `RECONCILE_FAIL_CLASSIFICATION`, `RECONCILE_FAIL_CASCADE_UNBOUND`, and `VERDICT_REASON`.
* N6: The compound note `2026-05-07-backlogit-shipment-status-constraints.md` instructs the retired flow (L54-58, L94), but the plan names only two compound entries for U14. Decide whether to add it to U14 or allowlist it in U15.
* N7: The plan names `2026-10-11-classify-only-writer-evidence-before-cascade-close.md` for U14, but the scope query does not return it. It is recorded in section 1b.
* N8: Upstream `cascade-close --classify-only` writes a pre_close evidence record under `docs/closure/evidence/` (help text: "read-only apart from creating a new evidence record"). The retired skill text says the mode is strictly read-only (L255-262). U3 adoption must not describe classify-only as writing nothing, and U12 must treat the evidence path as the expected write.
* N9: File lock scope differs (parity row 27). The skill acquires `.backlogit/queue/{S}.md` through `file-lock` (L239-241, L268-269). Upstream uses an O_EXCL pair lock. U5 must decide whether `file-lock` still wraps the cascade-close call.
* N10: The stash figure of 30 is reconciled as a population count (30 route-family keys) against 10 stale instances. The stash does not state its method, so this is an unreconciled inference.
* N11: The request described frontmatter conformity as warnings. In the report it is a failing targeted check with 18 errors. The `warnings` array holds 12 entries: 10 stale routing variables and 2 portability findings.
* N12: Linked-deliberation handling changed upstream (parity row 11). Under the verified 1.11 line, linked deliberations sit outside both sets and go through the INV-12 disposition. The 155-S rule that places them in `allowed_ids` is not reproduced. U5 must describe the new mechanism.
* N13: The upstream engine gate verifies only minor line 1.11 (parity row 19). The CI pin `1.5.0` predates 192-F (plan residual). This unit did not re-verify the CI pin.

## Out-of-scope discoveries

* D1: The harness U1 check "doc has 33-row map" is a weak substring test. A row count check would be stronger. Not changed by this unit.
* D2: `_orchestrator.agent.md` carries two P1 hardcoded home-path portability findings in the baseline. They are outside the U8 text listed in the plan. Not changed by this unit.

## Verification (043-S, unit U15)

### Measured state (verify-workspace, read-only report)

* Blockers: 0 (baseline 0). Strict schema blockers: 0.
* ROUTE_VARIABLE_STALE warnings: 0 (baseline 10 instances, 10 variables).
* frontmatter_conformity: ok (baseline failing with 18 missing-name errors; closed by unit 053.010-T).
* closure_source_artifact_cleanup: ok (baseline FAIL; closed by 053.007-T).
* orchestrator_invocation_routing_directive: ok (baseline FAIL; closed by 053.008-T).
* escalation_directive_present: ok (baseline FAIL; closed by 053.008-T).
* escalation_route_resolution: FAIL, EXPECTED (ESCALATION_DEGRADED, stage same-route no-op). Disposition: operator decision R1 (stash 0A215E0B). Not implemented in 043-S: the plan keeps `.autoharness/config.yaml` unchanged.
* reload_propagation_directive: FAIL, NOT EXPECTED by the plan. Cause: the check requires the literal phrase 'no independent model binding' inside the 'Propagate to inherited skills' section; no upstream template contains it. Disposition: operator decision (authorize wording, or correct the checker). Recorded in task 053.008-T.
* Warnings remaining: 2, both portability-finding (hard-coded home or autoharness path in agent files; pre-existing; not in 043-S scope).

### Retired-token sweep (git grep, excluding immutable history)

* `.github/**`: only `.github/policies/workflow-policies.md`, and only in the version-history rows 1.19.0 and 1.21.0 (append-only history, kept verbatim by decision D-3).
* `docs/compound/**`: `2026-05-07-backlogit-shipment-status-constraints.md`, `2026-09-30-shipment-reconcile-skill-bypass-manual-substitution-risk.md`, `2026-10-02-backlogit-stale-binary-on-path-breaks-index.md`, `2026-10-08-bound-cascade-close-procedure.md` (superseded note), `workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md`, `workflow-issues/ship-step6-post-merge-branch-ordering-2026-09-19.md`. Passing mentions are permitted by the plan; only the two superseded entries are in scope for U14.
* `tests/integration/ship_feature_completion_contract_test.go`: negative-test needles (permitted).
* `tests/integration/shipment_close_path_behavior_test.go` and `tests/integration/shipment_close_path_harness_test.go`: replica tests for the retired contract. DEFERRED: unit 053.013-T (per-operation approval required under Principle VII; dark mode does not waive it).
* This matrix document: permitted.

### Go verification

* `go vet ./...`: clean (exit 0). `gofmt -l .`: no output.
* `go test ./tests/integration/...`: all tests pass except `TestInvokeStartScriptMainPropagatesNonZeroExitCode`, which hits the 30 s budget of a start-script test on this host (host has real engram and backlogit sidecars on PATH, which the test comment says should be absent). It passes in isolation on most runs (about 20 s). Pre-existing timing flake; Batch 3 stash C5E50EFD covers start-script hermeticity.
* `go test ./...` full: the failing-test set equals the pre-043-S baseline set (55 test names) plus the one timing flake above. The 55 baseline failures are all in `tools/gatecheck`, caused by the Windows bash wrapper path (exit 127); they are environment-specific and pre-existing.
* New tests added by 043-S: `TestCascadeCloseClassifyOnlySmoke` (passes, about 12 s; real autoharness and backlogit against a `t.TempDir` fixture) and `TestCascadeRoutesThroughCommand` (passes).

### First mutating use: what the next shipment's closure must observe

* Probe first: `autoharness shipment --help` must exit 0; otherwise HALT.
* Classify first on every close: `autoharness shipment cascade-close --classify-only --shipment <S> --feature <F> --sha <merge_sha> --json`. Exit 0 selects CASCADE; exit 3 selects SAFE_CLOSE (nothing invoked); exits 2 and 4 HALT with nothing mutated; exits 5, 6, 7, 8 HALT for operator review with no commit, retry, or direct call; any other exit (including 1) HALTs as CLI too old or unexpected.
* Mutating run (CASCADE only, after per-closure operator approval not waived in dark mode): `autoharness shipment cascade-close --shipment <S> --feature <F> --sha <merge_sha> --message <text> --author <name> --timeout <secs>`; stay attached or poll in the background; the supervision budget is `--timeout` + 600 s.
* Evidence phases the record must show, in order: pre_close, then invoking, then post_close. The closure artifact must carry close_path (cascade or safe_close) and close_evidence (path `docs/closure/evidence/<S>-<F>-close-evidence.json`) and must pass `autoharness gate closure-evidence` before it is committed.
* A direct `backlogit shipment ship` or `backlogit_ship_shipment` call is a P-005 deviation that the closure-evidence gate refuses.

### Residual register (named, not silent)

1. Operator decision R1-R3 (stash 0A215E0B): routing values in `.autoharness/config.yaml`; not implemented in 043-S by plan design.
2. reload_propagation_directive (see above): operator decision on wording source.
3. U13 deferred: delete the two replica test files only after per-operation approval (Principle VII). Until then the shipment cannot close under the a1 gate, because 053.013-T is not done.
4. Version-history collision risk: policy row 1.26.0 (043-S) may collide with an upstream 1.26.0 row on a future sync.
5. Dangling anchors adopted from upstream (for example 'SUPERSESSION NOTE', 'D6', 'pre-mode step 7', and P-015 item 7 referring to 'precondition 1', which no longer exists in the local P-015 list). Recorded for the review gate.
6. P-007 restore sentence in the Ship agent safe-close pointer ('restore .backlogit/queue/ + .backlogit/archive/' on HALT - cascade detected) is retained from HEAD; upstream adds a G1 approval requirement. Pre-existing; out of 043-S scope; recorded for the review gate.
7. Stash C5E50EFD (start-script hermeticity) and stash 9F824B64 remain as Stage-owned items.

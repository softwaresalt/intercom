---
title: "Closure tooling Batch 1: one authoritative shipment-close procedure and scoped harness reconciliation"
description: "Stage deliberation over stash D10D3AFC (invokable safe-close) and A83C398E (tune-harness adoption of the 192-F closure contract and routing re-render): premise verification, the Ship-rule contradiction, option analysis, decisions D-CT-1..D-CT-7"
topic: "Which single procedure closes a shipment, how it becomes tool-enforced, and which installed harness surfaces must change"
depth: "deep"
decision_status: "decided"
promoted_to: "plan"
linked_artifacts:
  - "docs/plans/2026-10-11-intercom-go-closure-tooling-batch-1-plan.md"
  - "docs/closure/041-S-051-F-post-merge-closure.md"
  - "docs/closure/042-S-052-F-post-merge-closure.md"
  - "docs/compound/2026-10-11-classify-only-writer-evidence-before-cascade-close.md"
  - "docs/decisions/2026-10-01-intercom-go-post-m4-stash-triage-and-go-replan-deliberation.md"
tags:
  - "closure"
  - "shipment-reconcile"
  - "cascade-close"
  - "tune-harness"
  - "p-015"
  - "p-021"
  - "stage"
---

# Deliberation: Closure tooling Batch 1

## Session context

* **Agent and mode:** Stage, invoked by the Orchestrator (sequential, non-dark). The operator is
  absent, so Stage decides on evidence and records rationale and rejected alternatives. Genuinely
  authority-bearing items (model routing) are NOT decided here (D-CT-6).
* **Operator directive (verbatim intent):** Batch 1 of three. Batch 1 = stash `D10D3AFC` and `A83C398E`
  "with the Ship-rule contradiction reconciled". Batch 2 (write-path stash) and Batch 3 (start-script
  fix) are NOT staged now. Batch 1 must ship BEFORE any new implementation shipment, so Batches 2 and 3
  must later be chained behind it with `blocks` edges.
* **Operator conditions carried from the 042-S closure (2026-10-11T21:47 -07:00):** (1) `D10D3AFC` and
  `A83C398E` are the next Stage batch, before another implementation shipment; (2) Stage reconciles
  the Ship rule "Ship never calls `backlogit shipment ship` directly" with the absence of any executable
  safe-close "by implementing safe-close (`D10D3AFC`) or amending the rule".
* **Base:** `origin/main` @ `cc34242`. Stage artifact branch: `chore/stage-closure-tooling-batch-1`
  (Step 1.9 gate passed: only the current worktree; branch created from `refs/remotes/origin/main`;
  upstream tracking removed so the branch cannot be pushed to `main` by accident).
* **Slug source:** operator-supplied scope slug (the directive named it); equals the normalized
  covering-feature title prefix.
* **Tooling state (Step 0.0):** `ALL_TOOLS_OK` for the file-backed/CLI surface. The backlogit MCP tools
  are not exposed to this session, so the registry's declared CLI fallbacks (`backlogit ...`) are used
  throughout: `TOOL_DEGRADED: backlogit MCP — CLI fallback`. `INDEX_SYNC_OK (CLI fallback)`.
  Startup recovery: 75 checkpoints enumerated, 0 anomalies, 0 `stage`-owned active (all 46 stage and 29
  ship checkpoints are `resolved`): zero-candidate normal startup. Hook poll: 21 events (seq to 1159), none of
  the routed signals (`feature_review_ready`, `blocked_stale`).

## Step 1: Triage and P-021 obligations

| Stash | Priority | Kind | Shape | Route |
|---|---|---|---|---|
| D10D3AFC | high | task | `DEFERRED SCOPE EXPANSION` marker, task-shaped | `deliberate` (P-021 C6 precedence) |
| A83C398E | medium | task | `DEFERRED SCOPE EXPANSION` marker, task-shaped | `deliberate` (P-021 C6 precedence) |

Both entries carry the marker, so the Step 1 precedence rule forces deliberation. They are grouped
(Step 1.5): one coherent release unit, because `A83C398E` is the delivery vehicle (tune-harness) for the
procedure that `D10D3AFC` asks for. A grouping of two is valid; no artificial peers were added. `9F824B64`
(single-writer lock) was named "G2" with `D10D3AFC` in an earlier session but is NOT in the operator's
Batch 1 scope; see "Dispositions".

### (A) Unconditional duplicate scan: CLEAN

Scanned all 85 active stash lines for `safe-close|cascade-close|tune-harness|ROUTE_VARIABLE|closure-evidence|shipment-reconcile`.
Related but distinct, none a duplicate: `9F824B64` (lock-script contract), `C98B92F0` (closure-artifact naming),
`675F7181` (accuracy claims in an Orchestrator memory file). No second entry describes either expansion.

### (B) Late-identifier reconciliation

* `D10D3AFC`: unchanged since the 2026-10-01 reconciliation. Task `N/A` stands (no implementing task
  existed). Review-thread `N/A` stands (the PR #86 Copilot findings are inline comment IDs, not review
  threads). No late identifier found.
* `A83C398E`: PR `N/A` -> **#119** recovered. The 041-S closure-fix PR was the post-merge branch
  `post-merge/041-s-closure-evidence-waiver` (merge `3f92d7e`), which dispositioned the
  `cascade-evidence-record-missing` condition by operator waiver. Review-thread `N/A` stands (threadless).
  Task `N/A` stands (closure-level finding).

## Step 1.8: Learnings consulted (confidence: medium)

Prior art applied: `2026-09-30-shipment-reconcile-skill-bypass-manual-substitution-risk` (recurred in
035-S, 036-S, 039-S, 033-S; this deliberation adds 041-S and 042-S), `2026-10-11-classify-only-writer-evidence-before-cascade-close`,
`2026-10-08-bound-cascade-close-procedure`, `2026-05-07-backlogit-shipment-status-constraints`,
`workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13` (six-surface matrix),
`workflow-issues/ship-step6-post-merge-branch-ordering-2026-09-19` (6.0 stays ahead of the close sequence),
`workflow-issues/stash-jsonl-is-carried-pipeline-state-2026-09-19`. The researcher flagged a **contradiction inside the
compound library**: the 2026-05-07 and 2026-10-11 entries say Ship never calls `backlogit shipment ship` and
never uses the mutating `cascade-close`; the 2026-10-08 and 2026-09-30 entries read as accepted practice for a
manual direct ship. D-CT-1 resolves it; task U14 refreshes the two entries that instruct the forbidden or retired procedure.

## Premise verification (direct reads, not the stash summaries)

### P1. The "contradiction among Ship rules" is real, and it has a precise source

* **Installed Ship rule** (`.github/agents/_ship.agent.md`, Step 6.1 b, lines ~822-865): Ship runs
  `shipment-reconcile` `mode: classify-close-path`, carries `CLASSIFICATION_BINDING` into `mode: safe-close`,
  and "NEVER calls `backlogit shipment ship` directly, for any verdict". `.github/skills/shipment-reconcile/SKILL.md`
  (lines ~595-597) refuses an unbound safe-close and revalidates the binding.
* **P-015 (installed policy)** lists a "VERIFIED FULLY-COVERED-ROOT EXCEPTION" (lines ~440-450).
* **The 192-F contract** (upstream autoharness; `autoharness shipment cascade-close` and
  `autoharness gate closure-evidence` are present in the installed editable CLI; see the P7 correction on versions): every close starts with
  `cascade-close --classify-only` (pre_close evidence); CASCADE executes **only** through the mutating
  `cascade-close`; a direct `backlogit shipment ship` is a P-005 deviation that the closure-evidence gate refuses.
* **The 2026-10-11 compound entry** says the opposite of the 192-F contract: "Do not use the mutating
  `autoharness shipment cascade-close` from Ship."

### P2. The binding contract is a repo-local construct, not part of the upstream templates (key finding)

* `git log -S"CLASSIFICATION_BINDING"` on the installed skill: first and only introduction is
  commit `0cb4a42` (shipment **021-S**, 2026-09-16): "add classify-close-path boundary and covering-feature
  completion gate". Zero occurrences of `CLASSIFICATION_BINDING`, `classify-close-path`, or
  `RECONCILE_FAIL_CASCADE_UNBOUND` exist anywhere in upstream `templates/`, `docs/`, `src/`, `schemas/` of the
  autoharness repository (read-only inspection at `C:\Source\GitHub\autoharness`).
* The same commit added the Ship Step 6.1(a1) **covering-feature completion gate** and its P-010 role-boundary
  grant. Upstream has no (a1) either.
* `.autoharness/harness-manifest.yaml` has `tuned_at: null`; `verify-workspace` reports `_ship`, `_stage`, and
  `_orchestrator` as `user-modified`. The installed harness is the 1.5.0 install plus local overlays.
* **Consequence:** the "contradiction" is a collision between a repo-local, agent-executed-prose control (021-S)
  and the later upstream, tool-enforced control (192-F). A blanket tune would silently delete the local overlay,
  including the (a1) gate and surfaces pinned by three Go contract test files.

### P3. The upstream tool already performs the literal revalidation `D10D3AFC` asks for, for the CASCADE path

`autoharness shipment cascade-close --help` (read directly): the mutating form re-runs the classifier fresh,
probes engine semantics, selects the path through `select_close_path` (fail-closed to SAFE_CLOSE when
UNVERIFIED), fingerprints the shipment, members and out-of-manifest descendants, writes owner-bound evidence
`pre_close -> invoking -> post_close`, invokes `backlogit shipment ship` itself, and evaluates postconditions.
Exit **4** is "pre-invocation revalidation drift; nothing invoked. Never answer with SAFE_CLOSE." Exits 5-8
forbid re-running, direct `ship`, or substituting SAFE_CLOSE until an operator reviews. This is `D10D3AFC`
candidate (a) in tool form, with durable evidence, for the only path that has a cascade primitive.

### P4. The residual gap is SAFE_CLOSE only, and it is smaller than the stash implies

SAFE_CLOSE has no executor. Its mutations are one-artifact-at-a-time non-cascading `backlogit move`/`archive`
steps with verify-after-each. The destructive-cascade hazard that motivated the binding does not exist on that
path. It is a real but lower-risk gap, and the implementer is the upstream autoharness package (hash-pinned,
Python), which this workspace consumes and cannot patch.

### P5. `A83C398E`'s numbers are partly stale

A fresh `autoharness verify-workspace --workspace .` (read-only; it writes only to the gitignored
`.autoharness/staging/`): 0 blockers, 12 warnings, **10** `ROUTE_VARIABLE_STALE` variables today (the stash
says 30; the count depends on how instances are counted; the plan's acceptance criterion is "zero
`ROUTE_VARIABLE_STALE` warnings", not a number), 19 migration proposals, 1 unresolved `{{VARIABLE}}` in a staged `_ship` render.
**Six targeted checks FAIL**, now identified:

| Failing check | Cause | Resolved by |
|---|---|---|
| `frontmatter_conformity` | 18 installed skills lack the required `name` frontmatter key | add the `name` key to the 18 skill frontmatters (U10) |
| `closure_source_artifact_cleanup` | operational-closure skill lacks the "Source artifact cleanup" section | U7 (same file as the close_path keys) |
| `orchestrator_invocation_routing_directive` | Stage/Ship invocation sites lack the per-step `context_tier` / `ROUTING_DEGRADED` directive | routing re-render (U8 orchestrator and stage agents) |
| `reload_propagation_directive` | orchestrator "Propagate to inherited skills" marker lacks substantive content | routing re-render (U8 orchestrator and stage agents) |
| `escalation_directive_present` | `_stage.agent.md` carries a stale affirmative post-threshold retry directive | routing re-render (U8 orchestrator and stage agents) |
| `escalation_route_resolution` | `ESCALATION_DEGRADED`: stage escalation route equals the stage route | **operator decision** (D-CT-6); NOT fixable by a re-render |

### P6. Pinned surfaces (contract-closure matrix seed)

`tests/integration/ship_feature_completion_contract_test.go` (33-row contract + `TestCascadeRoutesThroughSafeClose`),
`shipment_close_path_behavior_test.go` (533 lines), `shipment_close_path_harness_test.go` (905 lines) pin the
021-S contract. `stage_branch_gate_test.go` and `staging_gate_actor_contract_test.go` pin Stage text.
Any rule-text change that does not move these in the same release unit breaks `go test ./tests/integration/...`.

### P7. Other facts checked

* **Correction after plan review (version is not capability).** `autoharness version` prints `1.5.0` and
  `autoharness home` is the upstream source checkout: the installed CLI is an **editable snapshot** of the upstream
  repository. The CI pin `autoharness==1.5.0` is the PyPI artifact, which predates 192-F; the upstream template says
  `cascade-close` "first ships in the autoharness release containing 192-F (the first release after 1.5.0)" and that an older CLI
  exits 1 (route: HALT, "there is no hand-run fallback"). So version strings cannot distinguish the two, CI cannot run
  these commands, and the shipped text must probe capability (`autoharness shipment --help`) and halt on "CLI too old".
  A tagged release containing 192-F and the CI pin bump are an upstream ask (D-CT-7).
* `9F824B64`'s claim "acquire_lock scripts do not exist" is true only for `scripts/`: the `file-lock` skill ships
  `.github/skills/file-lock/scripts/acquire_lock.ps1`. Recorded as a re-triage note on `9F824B64`; not in scope.

## Problem frame

One question with two halves. (1) **Which procedure closes a shipment, and how is its pre-mutation
revalidation made impossible to skip?** The agent-executed binding has been skipped or approximated in six
consecutive closures (035-S, 036-S, 039-S, 033-S, 041-S, 042-S), and an operator accepted the deviation twice.
A control that fails every time it is run by hand is not a control. (2) **How do the installed harness surfaces
come to say that one procedure, without destroying repo-local safeguards that are still wanted or breaking
pinned tests?**

## Decisions

### D-CT-1: ONE authoritative close procedure = the 192-F command-routed close (amend the rule; do not build a local safe-close)

The close procedure (Ship Step 6.1, after the unchanged 6.0 post-merge-branch gate) is:

1. `a0` topology gate (`autoharness gate pipeline-topology --phase lifecycle`), unchanged.
2. `a1` covering-feature completion gate, **retained** (D-CT-3).
3. `a` `shipment-reconcile` `mode: pre` (single-writer lock; Member-Class Status Contract).
4. `b` **classify and record:** `autoharness shipment cascade-close --classify-only --shipment S --feature F --sha X --json`
   (writes the `pre_close` evidence record; the skill's own `classify` in pre-mode is not a close-path source).
5. **CASCADE (exit 0):** execute ONLY through the mutating `autoharness shipment cascade-close` (no
   `--classify-only`), under the same destructive-command operator approval a direct `ship` call needs. The tool
   revalidates fresh, mutates, and writes `post_close`. Exit codes 3-8 route per its table; 5-8 halt to operator.
6. **SAFE_CLOSE (exit 3):** the skill's existing per-artifact sequence (manifest members one at a time via
   `move --status done` -> `archive`, then the shipment record via `move --status shipped` -> verify ->
   `archive` -> verify), protected-set verify-after-each, unchanged.
7. `d` `mode: post`; Linked-Deliberation Disposition; closure artifact carries `close_path` (`cascade` |
   `safe_close`) and `close_evidence`; `autoharness gate closure-evidence` validates it before commit.

**Retired (all introduced by 021-S):** `shipment-reconcile` `mode: classify-close-path`, `CLASSIFICATION_BINDING`
and its canonical v1 serializer, the `classification_binding` input, the `RECONCILE_FAIL_CASCADE_UNBOUND` /
`_CLASSIFICATION_INVALID` / `_DRIFT` / `_REFUSED` refusal tokens, the "fully-covered-root" P-015 exception
(replaced by the upstream flat-manifest engine-inertness containment text), and Ship's "invoke safe-close
with a binding" routing.

**The rule "Ship never calls `backlogit shipment ship` directly" is KEPT in substance and re-pointed.** A direct
`backlogit shipment ship` / `backlogit_ship_shipment` call remains a P-005 deviation. What changes is the
*sanctioned executor*: from an agent-run prose skill to the tool-enforced mutating `cascade-close`. This
satisfies operator condition 2 by amending the rule, with the implementation supplied by the already-shipped
upstream tool rather than a second local one.

**Why this is not a loosening (the "stricter local rule wins" test).** The mutating command is *stricter* than the
agent-computed binding it replaces: it recomputes, re-probes, fingerprints, and refuses on drift in code (exit 4),
produces durable evidence the gate verifies, and holds a per-pair lock. The 2026-10-11 compound entry's
objection ("it bypasses the classified-close boundary of the shipment-reconcile skill") described a boundary that
only exists as local prose; the tool's own classifier-and-engine gate is that boundary, executed.

### D-CT-2: `D10D3AFC` candidate choice. Select (c) adopt the shipped upstream command; reject (a-local) and (b)

* **(a-local) a new `autoharness gate shipment-reconcile --mode safe-close` in this repo: REJECTED.** The
  `autoharness` CLI is an upstream, hash-pinned Python package; this repo cannot add a subcommand to it. The Go
  `gatecheck` here implements four unrelated gates (`write-path`, `retired-arch`, `unignore`,
  `merge-strategy-evaluate`) and is not a closure tool; hosting a parallel Go closure engine would duplicate and
  drift from upstream's classifier and engine probe.
* **(b) a Ship-side hard gate in prose: REJECTED.** It is the same agent-executed control that was approximated
  six times; one more sentence cannot make a prose step literal.
* **(c) adopt `autoharness shipment cascade-close` (already satisfies candidate (a) for CASCADE): SELECTED.**
  The residual SAFE_CLOSE-executor gap is captured as a new deferred entry for the upstream package (D-CT-7),
  not implemented here.
* **Consequence for `D10D3AFC`:** consumed by this batch (the structural remediation is adoption + retirement of the
  prose control), with the SAFE_CLOSE residual forward-referenced to its replacement entry. It is not "obsolete":
  its premise (no invokable safe-close) is true for SAFE_CLOSE and false for CASCADE.

### D-CT-3: Retain the 021-S covering-feature completion gate (a1) as a documented local overlay

The (a1) gate is independent of the binding, was exercised unchanged in 041-S and 042-S against backlogit 1.11,
and is pinned by `TestShipFeatureCompletionContract33Rows`. Dropping a shipped safety grant inside a tooling
migration would be unreviewed scope. It stays, and the matrix (U1) records it as `RETAIN: local overlay, no
upstream counterpart`, so the next full tune does not delete it silently. Revisiting it is a separate decision.

### D-CT-4: A scoped, two-phase tune, not a blanket tune

Phase A (closure contract: U1-U7, the Go test migration U2/U11-U13, compound refresh U14) and Phase B (routing re-render and frontmatter: U8-U10). Each phase is previewed
with `autoharness verify-workspace` plus the auto-tune diff before any write, applied only to the named
surfaces, and gated on the matrix. A blanket tune of all 106 rendered artifacts is REJECTED: three managed
agents are `user-modified`, and a regeneration would erase the (a1) overlay and Stage/Ship text pinned by Go tests.
`tuned_at` stays `null`: a scoped hand re-render is not a tune (plan review, Scope finding), and the matrix
records exactly which surfaces were re-rendered.

### D-CT-5: Pinned tests move in the same release unit, and the behavioral replica is retired

The behavioral tests in `shipment_close_path_behavior_test.go` / `_harness_test.go` execute a **test-only Go
replica** of the retired classify/safe-close contract against fixture workspaces. With the contract retired they
verify nothing live. They are removed (U13), and the 33-row contract test is migrated by **expand/contract** so no commit is red
(Ship's per-task gate runs the full suite): U2 makes every retired-text assertion accept old OR new text, the text units
change the text, and U11 flips the assertions to the new text only. Behavioral coverage is kept where the real behavior now
lives: a fixture-backed `autoharness shipment cascade-close --classify-only` smoke test (U12, which lands before the
removal in U13; it skips with an explicit message when the command is absent). Accepted coverage loss: the replica
tests verified a test-only Go model, not production behavior. The retired contract and its tests leave together.

### D-CT-6: Routing decisions are OPERATOR-DECISION-REQUIRED and are not made in Batch 1

Model/provider/effort selection is operator authority (`config.model_routing` is an operator-owned file whose
edits are "preserved on re-install and tuning"). Stage does not edit `.autoharness/config.yaml`. Three items, each with
a recommended default, are captured as one active stash entry (D-CT-7) and are explicitly out of the shipment:

* **R1 stage escalation route equals stage/tier3 route (`ESCALATION_DEGRADED`).** Recommended default: declare
  `model_routing.stage.escalation` as the workspace's existing anchor route (`openai` / `gpt-6.1-sol` / `high`),
  a distinct provider, which satisfies the same-route guard. Until decided, `escalation_route_resolution`
  stays FAIL by design and U15 verifies it fails ONLY for that reason.
* **R2 ship routes to `claude-haiku-5.5` / `long_context` while the Tier 2 table lists ship.** Facts: `tier2` is also
  `claude-haiku-5.5`, so the route equals the Tier 2 default and `ship_model_routing_fields` passes. Recommended
  default: no change; record as intentional. Only the operator can say a Haiku-class model is acceptable for Ship.
* **R3 `alt_doc_review` comment says "tier1 is OpenAI" but tier1 is `anthropic/claude-haiku-5.5`.** A comment-only
  accuracy fix in an operator-owned file; the intended provider mix is the operator's. Recommended default: correct
  the comment to match the values.

The 10 stale variables re-render from the config as it is today; none requires a routing decision.

### D-CT-7: Shape of the release unit and new captures

* **Operator checkpoint:** retiring the 021-S control is Stage's evidence-based decision while the operator was
  absent; review of the staging PR is the operator's approval point before Ship claims the shipment.
* **One covering feature, one shipment**, flat (no sub-epics), 15 tasks U1-U15. Ordering protects Ship's own
  closure: tests expanded first, then text, then tests contracted, verification last. Batches 2 and 3 are NOT in this
  shipment and must not be started before it ships: when they are staged, each new shipment records
  `backlogit dep add <new-S> <this-S>` (type `blocks`), which is the same `dependencies:` edge `042-S -> 041-S` used.
* **New Stage-own captures (P-021, DEFERRED SCOPE EXPANSION):** (1) upstream asks: an invokable SAFE_CLOSE
  executor, and a tagged autoharness release containing 192-F plus the CI pin bump
  (`.github/constraints/autoharness-lock.txt`); (2) the operator routing-decision entry (R1-R3), which also notes that the
  stage agent's escalation route cannot be finalized until R1 is decided.

## Dispositions (P-021: Stage alone triages)

| Entry | Disposition |
|---|---|
| D10D3AFC | **CONSUMED** (promoted to the covering feature; archived at Step 5.6), SAFE_CLOSE residual forwarded to a new entry |
| A83C398E | **CONSUMED** (promoted; archived at Step 5.6); routing decisions split out as an operator-decision entry |
| 9F824B64 | **REMAINS DEFERRED**, not selected. Re-triage note: the `file-lock` skill ships its own `acquire_lock` scripts; the premise "scripts missing" is true only under `scripts/`. Reconciling that contract is separate from the close procedure |
| C98B92F0 | REMAINS DEFERRED (closure naming; related, distinct) |
| Batch 2 / Batch 3 stash | NOT touched (operator directive) |

## Trade-off comparison

| Option | Fixes the recurring failure | Local cost | Drift risk | Verdict |
|---|---|---|---|---|
| (a-local) own safe-close command | yes, in code | high (new engine, fork of a pinned tool) | high (duplicates upstream classifier) | rejected |
| (b) Ship-side prose gate | no (same control class) | low | none | rejected |
| (c) adopt upstream cascade-close, retire local binding | yes for CASCADE, in code, with evidence | medium (text + 3 test files) | lowest (tracks upstream) | **selected** |
| Blanket tune | partly | low effort, high blast radius | erases (a1), breaks pinned tests | rejected |
| Do nothing, accept deviations | no | none | grows | rejected |

## Unresolved questions

* R1-R3 (operator).
* The SAFE_CLOSE executor (upstream).
* Whether the (a1) overlay should eventually be proposed upstream (not decided here).

## Risks and mitigations

* **Mutating cascade-close is destructive** and now Ship's sanctioned path. Mitigation: operator approval per
  closure (unchanged), exit 5-8 halt rules, evidence + `closure-evidence` gate, and the first real use is
  the next shipment's closure, which is itself this batch's verification (U15 states what to observe).
* **Pinned-test breakage** during the rewrite. Mitigation: U1's matrix, and tests move in the same shipment with
  `go test ./tests/integration/...` as a per-task acceptance gate.
* **Over-deletion** of repo-local text by a scoped tune. Mitigation: U1 keep/retire/adopt table is the checklist and
  every task's acceptance names the retained overlay lines.
* **Template-vs-installed drift on the unselected 90+ artifacts** remains; `tuned_at` stays `null`.
* **SAFE_CLOSE is left without an executor** (named residual): it keeps the skill's pre-mutation protected-set baseline
  gate and verify-after-each, and has no cascade hazard.
* **CI cannot run the commands** (PyPI pin predates 192-F): residual, tracked as an upstream ask.

## Rejected alternatives

See D-CT-2, D-CT-4. Also rejected: keeping the binding **and** adding cascade-close (two overlapping controls
with conflicting executors is the contradiction itself); implementing the safe-close CLI in `tools/gatecheck`;
editing `config.yaml` comments in this batch (R3 is operator-owned).

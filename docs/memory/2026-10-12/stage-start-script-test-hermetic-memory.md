---
title: "Stage session memory: hermetic start-script tests (Batch 3)"
date: 2026-10-12
agent: stage
session_id: stage-start-script-test-hermetic-20261012
status: complete
branch: chore/stage-start-script-test-hermetic-harness
base: "chore/stage-writepath-batch-2 @ 1068450 (-> chore/stage-closure-tooling-batch-1 @ f20b6d1 -> origin/main cc34242); operator-directed stacked base, Step 1.9 checks passed (clean worktree, single worktree, symbolic HEAD, branch differs from origin/main)"
---

# Stage session: hermetic and containable start-script integration tests (Batch 3)

## Outcome

* **Shipment (queued):** `046-S`, 4 items: `056-F`, `056.001-T`, `056.002-T`, `056.003-T`
  (parent-first order verified by read-back).
* **Edges (`blocks`):** `046-S -> 045-S` (ordering, operator directive) and `046-S -> 043-S`
  (direct Batch 1 contract edge); chain 043-S <- 044-S <- 045-S <- 046-S.
* **Required merge/ship order:** Batch 1 staging PR (`chore/stage-closure-tooling-batch-1`,
  merge commit), then Batch 2 (`chore/stage-writepath-batch-2`), then this Batch 3 staging PR.
  Ship claims 043-S, 044-S, 045-S, then 046-S. None pushed.
* **File overlap:** none with Batch 1 (`.github/**`, `.autoharness/**`, policy docs, three
  contract tests plus a new smoke test under `tests/integration`) or Batch 2 (`tools/gatecheck/**`,
  `scripts/**`). Same Go package (`package integration`) namespace only; distinct identifier
  prefixes; confirm before Ship that Batch 1's deleted replica files do not define `repoRoot`,
  `writeStubExecutable`, `filterEnvKeys`.

## Artifacts

* Decision note (no full deliberation): `docs/decisions/2026-10-12-intercom-go-start-script-test-hermetic-decision-note.md` (D-HS-1..D-HS-6, triage record for P-021 C5/C6).
* Plan: `docs/plans/2026-10-12-intercom-go-start-script-test-hermetic-plan.md`
  (`Requires plan hardening: no`; plan review `dispatch_mode: multi-agent`, `decision: ADVISORY`,
  no P0/P1; five personas as subagents plus an anchor-route (`gpt-6.1-sol`) Architecture pass on
  the revised plan; every P2 dispositioned in the revision).

## Backlog

| Item | Unit | Size / Complexity |
|---|---|---|
| `056-F` Hermetic and containable start-script integration tests | covering feature | n/a |
| `056.001-T` Add hermetic start.ps1 scratch copy and PATH helpers | U1 | S / medium |
| `056.002-T` Add WaitDelay-bounded runner with containment guard | U2 (blocked by U1) | S / medium |
| `056.003-T` Rewire two start-script tests to hermetic bounded harness | U3 (blocked by U2) | S / low |

Degradation: this workspace's `backlogit` task schema defines `size` but **not** `complexity`
(`update --complexity` rejects it), so `size: S` (+ `size_source: agent`,
`size_ruleset_version: 2h-rule-v1`) is structured and `Complexity: ...` is preserved as prose in
each task's implementation notes (same as Batch 2).

## Stash dispositions

* **Archived (consumed):** `C5E50EFD` (promoted to 056-F; no late identifier found, `N/A`
  stands; clean duplicate scan), `26EDD633` (ABSORBED: same test and symptom; its
  workspace-size hypothesis is unconfirmed by the accepted RCA; hermeticity removes the sidecar
  workload; symptom covered by 056.003-T AC 4; archived with this note as the promotion record).
* **Untouched (active):** `96B308D1` (different test and package; not absorbed) and every
  Batch 1/2 entry.
* **Created (P-021, Stage own-authority intake):** `B32FCDD1` (low): `runStartScriptSnippet`
  `WaitDelay`/hermetic parity and the `Resolve-GitHubTokens` ambient-token residual. Needs Stage
  deliberation after 046-S ships.

## Operational notes

* Tool gate: no MCP tools in this session, so the registry's `cli_command` fallbacks were used
  (`TOOL_DEGRADED: backlogit MCP, CLI fallback active`). Index sync `INDEX_SYNC_OK (CLI)` at start.
* Startup recovery: 81 checkpoints enumerated (52 stage, 29 ship, all resolved), zero anomalies,
  zero active `stage` candidates: normal startup. Hook queue: 38 `create_artifact`/
  `update_artifact` events, no actionable signals; acked through seq 1197.
* Step 1.8 learnings ran after the branch gate (ordering slip, no tracked mutation preceded it).
  Medium confidence; applied: positive PATH proof with control, case-folded env, `os.SameFile`,
  absolute interpreter path, verify `WaitDelay` against `GOROOT` source.

## Operator-decision items

1. Plan review is ADVISORY (Stage authorized harvest); override at staging-PR review if desired.
2. Task split: three green tasks (red-then-green inside each) rather than a separate red-only
   task, so Ship's green-at-every-task gate holds; Scope reviewer preferred 1 to 2 tasks.
3. G3 containment guard (helper-process) kept despite a proportionality objection, because the
   operator named containment as a requirement.

Next: operator pushes and merges Batch 1, then Batch 2, then this staging PR (merge commits);
Ship claims 043-S, 044-S, 045-S, 046-S in order.

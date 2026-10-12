---
title: "Stage session memory: Gatecheck write-path Batch 2"
date: 2026-10-11
agent: stage
session_id: stage-writepath-batch-2-20261011
status: complete
branch: chore/stage-writepath-batch-2
base: "chore/stage-closure-tooling-batch-1 @ f20b6d1 (origin/main cc34242); operator-preferred stacked base, Step 1.9 checks passed"
---

# Stage session: Gatecheck write-path Batch 2

## Outcome

* **Shipments (queued):** `044-S` (S1, 10 items: `054-F` + `054.001-T`..`054.009-T`) and `045-S`
  (S2, 8 items: `055-F` + `055.001-T`..`055.007-T`).
* **Chain:** `043-S` ← `044-S` ← `045-S` as `blocks` dependency edges (`backlogit dep add`), recorded
  in the shipment files (`dependencies`). Read back and verified.
* **Merge order (REQUIRED):** the Batch 1 staging PR (`chore/stage-closure-tooling-batch-1`) merges
  first as a **merge commit** (never squash); then this branch's staging PR. This branch carries
  the Batch 1 commit `f20b6d1`.
* **File overlap with 043-S:** none (043-S: `.github/**`, `.autoharness/**`, policy docs,
  `tests/integration/*`; Batch 2: `tools/gatecheck/**`, `scripts/**`).

## Artifacts

* Deliberation: `docs/decisions/2026-10-11-intercom-go-gatecheck-write-path-batch-2-deliberation.md`
  (D-WB2-1..D-WB2-10).
* Plan: `docs/plans/2026-10-11-intercom-go-gatecheck-write-path-batch-2-plan.md`
  (hardening section; plan review attempts 1 and 2 FAIL, attempt 3 **ADVISORY**, all six personas
  as subagents; ADVISORY authorized by Stage under the Orchestrator directive, operator may
  override at staging-PR review).
* Compound note: `docs/compound/workflow-issues/git-pathspec-glob-misses-top-level-link-named-like-the-prefix-2026-10-11.md`.

## Backlog

| Feature | Shipment | Tasks (unit) | Stash consumed |
|---|---|---|---|
| `054-F` Write-path gate completeness and CI-log output safety | `044-S` | 054.001 (U1), .002 (U2), .003 (U3), .004 (U4), .005 (U5), .006 (U5b), .007 (U6), .008 (U7), .009 (U8) | EDC18D59, EEE39D23, 858516B9 |
| `055-F` Gate git discovery and environment isolation across wrappers and engines | `045-S` | 055.001 (U9), .002 (U10), .003 (U11), .004 (U12), .005 (U13), .006 (U14), .007 (U15) | DBE25DF5, 2613D248 |

Sizes are structured (`size`, `size_source: agent`, `size_ruleset_version: 2h-rule-v1`); **complexity
is preserved as prose** (`Size: M | Complexity: medium` in each task's implementation notes)
because this workspace's `task` artifact type defines no `complexity` field (`backlogit update
--complexity` fails validation). Degradation disclosed; same as Batch 1.

## Why two shipments (D-WB2-1)

Review-size cap and gate-risk isolation. S1 is one Go package; S2 mixes shell, a shared lib and the
frozen retiredarch pin. A 14-unit plan in this repo already had to be split after three plan-review
FAILs, and B1 (5 tasks) hit the review cycle cap. S2 follows S1 because both edit `writepath.go`
and `writepath_test.go`.

## Stash dispositions

* **Archived (consumed):** `EEE39D23`, `DBE25DF5`, `858516B9`, `EDC18D59`, `2613D248`, each with a
  `STAGE DISPOSITION` block. Deferred stash entries not touched: `C5E50EFD` (Batch 3) and every
  other entry.
* **New DEFERRED SCOPE EXPANSION entries (active, P-021 C2 fields, Stage own-authority intake):**
  `36BE1F61` (low, retiredarch echo sinks and `main.go` sinks), `1B85D0F4` (medium, other wrappers'
  unscrubbed root discovery), `A8FB0682` (low, fixture hard-link refusal, the unimplemented remainder
  of 858516B9), `284270E8` (medium, retiredarch scan-set twin of EEE39D23, found by plan review
  SEC-1), `49982A4E` (low, engine-side `--root` top-level assertion).
* **P-021 C5:** duplicate scan CLEAN for all five; late identifier PR #123 (carrier of 042-S/052-F)
  recovered from `docs/closure/042-S-052-F-post-merge-closure.md`; thread and task IDs: none found,
  `N/A` stands; feature 052-F and shipment 042-S recovered for DBE25DF5 and EDC18D59.

## Decisions to review (operator checkpoints, defaults stand)

1. Two shipments, chain 043-S → 044-S → 045-S.
2. Case-variant directories out of the gate's scope (accepted residuals listed in the plan).
3. Write-path wrapper root moves from caller-cwd to script-location anchoring (with a
   same-directory assertion; a nested checkout fails closed).
4. Hard-link refusal is not implemented (`A8FB0682`).
5. Per-arm ED-7 guard rejected (plan-review AS-5).
6. Stacked base and merge order.

## Facts worth keeping

* The pathspecs `internal/**` and `cmd/**` do not list a tracked link named exactly `internal` or
  `cmd` (probe-verified); the runner's pathspecs become the literal `internal` and `cmd` (U6).
* `retiredarch/select.go gitRunnerEnv` is mirrored in `pin.go canonicalDecls`; U15 changes both in
  one commit and the MSYS rule must be inline (no new declaration, quoted
  `"GIT_CONFIG_NOSYSTEM=1"` exactly once).
* `scripts/testdata/writepath/harness/` is a legitimate real directory; U8 refuses only `.go`-named
  directories.
* Plan-review spend: three attempts; the P1s were Constitution Check, safety modes, test commits
  must compile, the top-level link pathspec, and the root same-directory assertion.

## Next

Operator: approve and merge the Batch 1 staging PR first, then this staging PR (merge commits).
Ship then claims `043-S`, later `044-S`, then `045-S`. Batch 3 (`C5E50EFD`) is a separate Stage
invocation.

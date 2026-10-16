---
title: "046-S / 056-F compacted memory (Stage Batch 3 staging and Ship session)"
date: 2026-10-16
shipment: 046-S
feature: 056-F
agent: Ship
status: compacted
originals:
  - docs/archive/memory/2026-10-12/stage-start-script-test-hermetic-memory.md
  - docs/archive/memory/2026-10-16/046-s-056-f-ship-session-memory.md
---

# 046-S / 056-F compacted memory

## Staging (2026-10-12, Stage Batch 3)

* Shipment 046-S held 056-F plus 056.001-T (hermetic start.ps1 scratch copy and PATH helpers),
  056.002-T (WaitDelay-bounded runner with containment guard) and 056.003-T (rewire two start-script
  tests). Blocks edges: 046-S -> 045-S and 046-S -> 043-S (chain 043 <- 044 <- 045 <- 046).
* Plan review was ADVISORY (no P0/P1). Three green tasks (red-then-green inside each) instead of a
  red-only task, so Ship's green-at-every-task gate holds. The G3 containment guard was kept because
  the operator named containment a requirement.
* Decision note: `docs/decisions/2026-10-12-intercom-go-start-script-test-hermetic-decision-note.md`
  (D-HS-1..D-HS-6). Plan: `docs/plans/2026-10-12-intercom-go-start-script-test-hermetic-plan.md`.
* Stash: C5E50EFD promoted to 056-F and archived; 26EDD633 absorbed and archived; B32FCDD1 (low) captured
  under P-021 for `runStartScriptSnippet` WaitDelay/hermetic parity and the `Resolve-GitHubTokens`
  ambient-token residual (still live, needs Stage deliberation); 96B308D1 untouched.
* The backlogit task schema has `size` but no `complexity`; complexity is kept as prose in task notes.
* Learnings applied: positive PATH proof with control, case-folded env, `os.SameFile`, absolute
  interpreter path, verify `WaitDelay` against `GOROOT` source.

## Shipping (2026-10-15/16, Ship)

* Feature PR #141 merged with a merge commit at 75e8755d8b8f7f98a8b952bb529352fa6590550a (reviewed
  code HEAD 1dedea6, PR head a6bb6fc, CI green, P-018 NOT_APPLICABLE). 046-S cascade-closed with
  operator approval (exit 0, postcondition pass); archive relocation committed on
  `post-merge/056-hermetic-start-script-tests`.
* Decisions: test-only change; the G1 probe uses the dot-sourced function's script file because a
  wrapper cannot see the sourced `$PSScriptRoot`; `//nolint:unparam` on `writeStubExecutable` after
  the rewire left its `stdout` parameter always "".
* Process deviation: the pre_claim topology gate was run once with the operator-only `--force` flag by
  Ship after a branch-name mismatch (audit log `.autoharness/gates/pipeline-topology-force-audit.log`);
  the branch was renamed and the unforced post_claim gate passed.
* `.autoharness/config.yaml` was carried uncommitted until it matched main (PR #140), then restored
  from main with a byte-verified backup.
* Known unrelated: 21 `tools/gatecheck` wrapper-test failures on this Windows host (WSL bash first on
  PATH); Linux CI is authoritative.
* Not claimed: 047-S, 048-S, 049-S.

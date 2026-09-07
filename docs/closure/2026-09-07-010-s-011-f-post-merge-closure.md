---
title: "Post-Merge Closure — Shipment 010-S / Feature 011-F"
date: 2026-09-07
shipment_id: 010-S
feature_id: 011-F
pr: 29
merge_commit: 797fc1ed9e4833374552d8bfaa3ddd41e46fec26
compaction_status: done
closure_status: READY
---

# Post-Merge Closure — 010-S / 011-F: Residual Hardening Round 2

## Summary

Shipment `010-S` (feature `011-F`, "Residual hardening round 2") merged to
`main` via PR [#29](https://github.com/softwaresalt/intercom/pull/29),
merge commit `797fc1ed9e4833374552d8bfaa3ddd41e46fec26`. All 20 units
(`011.001-T`–`011.020-T`) completed, reviewed, and shipped.

## Dispositions

| Task | Disposition | Commit(s) |
|---|---|---|
| 011.001-T | Done | `239b3ed` |
| 011.002-T | Done | `40683fe` |
| 011.003-T | Done | `c4f6ec8` |
| 011.004-T | Done | `272b186` |
| 011.005-T | Done | `e7d1f1e` |
| 011.006-T | Done | `272b186` |
| 011.007-T | Done | `272b186` |
| 011.008-T | Done | `8455833` |
| 011.009-T | Done | `750e8e8` |
| 011.010-T | Done (mechanical assertion only; `WINDOWS_GATE_REQUIRED` flip deferred, requires explicit operator confirmation) | `e65a6a0` |
| 011.011-T | Done | `3d9a6e2` |
| 011.012-T | Done | `a854641` |
| 011.013-T | Done | `1982c45` |
| 011.014-T | Done | `31fbecb` |
| 011.015-T | Done | `ee23297`, `64f1a32` |
| 011.016-T | Done | `e267848` |
| 011.017-T | Done | `e267848` |
| 011.018-T | Done | `e267848` |
| 011.019-T | Done | `6cf77c6` |
| 011.020-T | Done | `4fd3913` |

Plus review-remediation commits `a251191` (local multi-persona review),
`0b93ec2` (multi-model adversarial review), `19bc112` (deferred-scope
captures), `7c136c5` (critical CI YAML structural fix), `eeaac0c`
(Copilot review fix).

## Review outcomes

* **Local multi-persona review**: 9 personas (Correctness, Security, Go,
  Constitution, Maintainability, Architecture, Scope Boundary,
  Schema-CLI-Docs Coupling, Template Integrity). All findings remediated.
* **Multi-model adversarial review**: 4 models (Anchor, Tier1, Tier2,
  Tier3). Full report:
  `docs/closure/2026-09-06-010-s-011-f-residual-hardening-round2-adversarial-review.md`.
  1 CRITICAL (self-matching CI assertion), 1 CRITICAL plan-AC violation
  (missing permission-handler allowlist), 5 MAJOR, 2 MINOR — all
  remediated. 9 out-of-scope findings deferred to stash (P-021 C2).
* **Copilot review**: 2 rounds. Round 1 found 1 comment (fixture-comment
  inaccuracy), fixed, replied, thread resolved. Round 2: `SATISFIED`, zero
  unresolved threads.
* **CI**: all checks green for the merged HEAD (`test`, `test (windows,
  advisory)`, `lint`, `security`, `cross-compile` ×4, `gitignore
  append-only + un-ignore regression`, `pipeline-topology`, `ci gate`).

## Runtime verification

This shipment's 20 units touch containment-library internals
(`internal/pathsafe`, `internal/config` docs), a test-only live-SDK gate
(`internal/copilotprobe`), CI/lint configuration, verification scripts,
and `start.ps1`'s bootstrap/launch behavior. **No unit modifies
`cmd/intercom` or `cmd/intercom-ctl`'s actual runtime/CLI operational
behavior.** Runtime-verification is therefore treated as **largely
not applicable** for this shipment; a lightweight smoke check was run
as a sanity confirmation rather than a full validator-manifest pass:

* `go run ./cmd/intercom --help` — cold-starts and responds correctly
  (exit 0).
* `go run ./cmd/intercom-ctl --help` — cold-starts and responds correctly
  (exit 0).

No adapter probes, manual checkpoints, or blocked prerequisites apply
beyond the above.

## Releasability

**READY.**

* **Monitoring**: no new runtime surfaces introduced; existing monitoring
  posture (none additional required for this shipment) is unaffected.
* **Rollback**: every unit is independently revertible per the plan's own
  rollback section. The two highest-risk units (`011.009-T`,
  `011.016-T`–`011.018-T`) revert without disturbing others. `011.015-T`
  is a tail-truncation revert by construction. `011.010-T`'s gate flip was
  never applied (deferred), so there is nothing to roll back there.
* **Owner**: `softwaresalt` (repository owner).
* **Validation window**: none required beyond CI; the advisory
  `windows-latest` job (`WINDOWS_GATE_REQUIRED` unset) already produced a
  fully green run in the merged PR's own CI (run
  `34078894440`), which is the evidentiary basis an operator may cite if
  choosing to flip that variable to `'true'` later.
* **Follow-up requirements**: see Deferred Follow-Ups below.

## Deferred follow-ups (stashed, P-021 C2)

| Stash ID | Summary | Priority |
|---|---|---|
| `F47DB9A9` | `ci.yml`'s `topology-check` job references two docs files (`pipeline-topology-gate.md`, `-ci-rollout.md`) that don't exist (pre-existing) | low |
| `ECE3DAB7` | `checkSymlinkEscape`'s ancestor-existence probe may not correctly handle a dangling intermediate symlink (pre-existing, 005-S) | medium |
| `F4F4A959` | `ci.yml` lint job's retired-architecture gate has a bare, undocumented `continue-on-error: true` (pre-existing) | low |
| `5FE4A7BE` | `pathsafe` case-folding is Windows-only; darwin has no coverage (unverified) | low |
| `5158769F` | `output_path_guard_test.go`'s `createDirectorySymlink` skips too permissively on Windows (unverified) | low |
| `2787DA56` | `check-unignore-regression.sh` treats a `git show` failure as an empty baseline rather than failing closed (unverified) | low |
| `CA469B3D` | `stripUNCPrefix` may mishandle extended UNC paths (unverified, pre-existing) | low |
| `4C5BEC23` | `check-unignore-regression.sh`'s header describes a mechanism (`git worktree`) the implementation doesn't use | low |
| `805248F7` | Truncated GoDoc comments in two `pathsafe` test files | low |

Also: **`WINDOWS_GATE_REQUIRED` flip** (`011.010-T`) — the durable
pass/fail assertion mechanism is merged and proven green in this PR's own
CI run, but flipping the repository variable to `'true'` (making the
`windows` job blocking) is a Constitution VII repository-configuration
change requiring explicit operator confirmation, not something Ship
self-authorizes. Recorded here as an operator action item, not stashed
(it is a configuration decision, not a code-change backlog item).

## Source artifact cleanup

Reviewed `custom_fields.source_stash_id` / `source_deliberation_id` for
the shipped scope (`011-F`). Stash `9F9A3CB2` (the harness-bootstrap
carryover entry) is recorded as **partially resolved** per its own
framing (011.019-T) and remains active per that entry's own requirement
("remains active until Ship's PR disclosure completes" — disclosure is
now complete via this merged PR, but the entry's final closure is Stage's
own triage responsibility, not auto-archived by Ship here). No other
`source_stash_id`/`source_deliberation_id` fields were found on `011-F`
or its tasks pointing to a cleanly-archivable single source artifact
(the shipment's provenance traces to the round-2 deliberation, which
Stage retains per its own artifact-lifecycle ownership).

## Compaction status

See `compaction_status` in this file's frontmatter, finalized by the
mandatory `compact-context` invocation (P-020) at the end of this
closure session.

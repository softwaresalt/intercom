---
title: "Ship session end: 016-S / 017-F — pathsafe error-surface parity and Windows long-path precision"
date: 2026-09-11
agent: Ship
mode: DARK_MODE_ACTIVE
shipment: 016-S
feature: 017-F
status: complete
---

# Ship session end: 016-S / 017-F

## Outcome

Shipment 016-S executed end-to-end under the dark-factory contract: claim →
TDD (RED-then-GREEN) → quality gates → multi-persona local review → PR #51 →
Copilot review loop → merge-commit merge → post-merge closure → closure PR
#52 → Copilot review loop → merge-commit merge → return to `main`.

**No halt occurred.** All dark-mode stop conditions were checked at every
required gate and none tripped.

## PRs and merges

- **PR #51** (feature): merge commit `75cfe3b1b954c49d149f531464ccd97f1770dd0b`.
  Reviewed HEAD `090cbf5`. Copilot: `SATISFIED` (round 2), 2 threads
  resolved.
- **PR #52** (post-merge closure): merge commit
  `b7c8678e03d144948509fd8884e09c959a914219`. Reviewed HEAD `38a45e9`.
  Copilot: `SATISFIED` (round 3), 6 threads across 2 rounds resolved.

Both merges used the merge-commit strategy only (P-009); no admin fallback
used or required anywhere in this session.

## Shipment / backlog state

`016-S` shipped and archived via the P-015 verified fully-covered-root
cascade path (`backlogit shipment ship`). `017-F`, `017.001-T`,
`017.002-T`, `017.003-T` all archived (`status: done`, `parent_id`
preserved). Two-set gate verified (`archived_ids == allowed_ids ==
required_ids`), `returned_ids` empty, no archive-file deletions (P-007).

## Stash state (final verification)

- `BF5DE670`: active, untouched (AC-F4). Trigger (live destructive write
  primitive) still has not fired.
- `6B751D8B`: active, untouched (AC-F4). Items 1–2 harvested into
  017.002-T/017.003-T in a prior cycle; items 3–4 remain deferred.
- `7ADAC481`: absent from active stash (already harvested/archived by
  Stage prior to this session — confirmed, not touched by Ship).
- `475E76D2`: new, active. P-021 deferred-scope entry captured this
  session (threadless path) for a rejected perf micro-optimization
  (`utf16Len` byte-length fast path) that the plan had already withdrawn
  as gold-plating.

## Closure / compaction

- Runtime verification: `READY` (corrected during closure-PR review: no
  `cmd/` entrypoint currently calls `config.Load`/`Validate` at process
  startup; Go test suite is the evidence for the changed code, not CLI
  startup).
- Operational closure: releasability `READY`, no conditions.
- Compaction status (P-020): `done`. This release unit's two memory files
  (Stage planning + Ship pre-PR checkpoint) compacted into
  `docs/memory/compacted/2026-09-11-016-s-017-f-pathsafe-error-surface-parity-compacted.md`;
  verbose originals archived to `docs/archive/memory/`.
- New compound entry: `docs/compound/2026-09-11-copilot-suppressed-comments-vs-review-threads.md`
  (Copilot review-body "Suppressed comments" text can name findings with no
  corresponding GraphQL thread).
- Backlog index resynced twice (`backlogit sync`) — after shipment closure
  and again after the closure PR merged. Both `CLOSURE_INDEX_SYNC_OK`.

## Residual risks / follow-ups

None new beyond the stash entries above (`BF5DE670`, `6B751D8B` — pre-existing,
conditionally deferred; `475E76D2` — new this cycle, low priority,
performance-only). No monitoring gaps, no documentation debt, no deferred
scope beyond what is already captured in the stash.

## Branch / workspace state

Returned to `main`, fast-forwarded to `b7c8678`. **Correction (post-merge
repair, `chore/016-s-post-merge-memory-residual-repair`):** the sentence
originally written here — "working tree clean" — was inaccurate at the
moment it was written: this very checkpoint file (`docs/memory/2026-09-11-ship-016-s-session-end.md`,
untracked) was itself present in the working tree, which the original
claim failed to account for. The residual artifact was resolved via a
dedicated post-merge closure-repair PR: its unique facts were folded into
`docs/memory/compacted/2026-09-11-016-s-017-f-pathsafe-error-surface-parity-compacted.md`
(added to that summary's `compacted_from` list) and this verbose original
was archived here to `docs/archive/memory/`, per the `compact-context`
skill's archive convention. The working tree is clean as of that repair
PR's own merge.

Feature branch `feat/016-s-pathsafe-error-surface-parity-and-windows-long-path-precision`
and closure branch `post-merge/017-f-pathsafe-error-surface-parity` both
fully merged; not deleted (no explicit cleanup request/config seen this
session).

## Notable process learning this session

Copilot review remediation required **three rounds** on the closure PR
because the first fix pass corrected the primary occurrence of a factual
error (runtime-reachability overclaim) but missed two additional verbatim
recurrences of the same claim in sibling documents/sections. When a
correction spans multiple documents/sections, grep the whole affected
document set for the error pattern before considering the fix complete,
rather than fixing only the location(s) explicitly cited by the reviewer.

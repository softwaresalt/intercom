---
title: "Stage session summary — 015-S / 016-F pathsafe Windows canonicalization symmetry"
description: "Session memory for the dark-factory Stage cycle that produced shipment 015-S"
date: 2026-09-10
tags:
  - "stage"
  - "memory"
  - "015-S"
  - "016-F"
  - "dark-factory"
mode: DARK_MODE_ACTIVE
shipment: 015-S
feature: 016-F
---

# Stage session summary — 2026-09-10

**Mode**: DARK_MODE_ACTIVE, operator AFK, CLI visibility (agent-intercom not installed).
**Branch**: `main` @ `db4edf0`, single worktree, no spike worktree created (P-016 clean).
**Bounded scope**: BF5DE670, 4104AF54, E428AB46, E4C5413F. Exactly one shipment.

## Outcome

| Item | Result |
|---|---|
| Shipment | **015-S** (`queued`), 9 items, `unsized: 0` |
| Feature | 016-F |
| Tasks | 016.001-T … 016.008-T |
| Harvested entries | 4104AF54, E428AB46, E4C5413F (archived) |
| Deferred entry | BF5DE670 (remains active, evidence-backed) |
| Plan review | PASS at attempt 2 (FAIL at attempt 1 on two P1s) |

## Artifacts

- `docs/decisions/2026-09-10-intercom-go-pathsafe-windows-canonicalization-symmetry-deliberation.md`
- `docs/plans/2026-09-10-intercom-go-pathsafe-windows-canonicalization-symmetry-plan.md`
- `docs/decisions/2026-09-10-intercom-go-pathsafe-windows-canonicalization-plan-review.md`

## Decisions worth carrying forward

- **D-2**: `NewRoot` adopts `canonicalizeReparse` — provable no-op on POSIX because
  `reparse_other.go`'s `canonicalizeReparse` *is* `filepath.EvalSymlinks`.
- **D-4**: BF5DE670 deferred again. Trigger re-measured: write-path gate exits 0,
  self-test 5/5, independent grep finds zero write primitives. Third consecutive
  cycle with the same measured result.
- **D-5**: The "EvalSymlinks calls GetFinalPathNameByHandleW" claim is a
  twice-corrected falsehood; it must not re-enter any artifact.

## Review findings that changed the plan

1. **P1** — B3 (U5 long-path) promoted to a *hard* prerequisite of C2. `NewRoot`
   currently gets Go's long-path handling free via `EvalSymlinks`' `os.Lstat`
   walk; routing it through `canonicalizeReparse`'s raw `CreateFile` removes it.
   Landing C2 alone would introduce a **new** fail-closed regression.
2. **P1** — C2 changes `NewRoot`'s Windows error type from `*fs.PathError` to raw
   `syscall.Errno`. The 009-S suite would stay green while
   `errors.As(err, &*fs.PathError)` callers silently stopped matching.
3. **P2** — Added C1b, a containment-*preservation* lock. The original plan proved
   only that C2 does not over-reject, never that it does not under-reject.
4. **Retraction** — the claimed B2→C2 hard dependency was falsified (`stripUNCPrefix`
   is idempotent, both callers stay normalized). Corrected in plan **and**
   deliberation rather than left standing.

## Corrections to inherited context

- The `pwsh` silent-skip in `createDirectoryJunction` was reported by the learnings
  researcher as an open pitfall. **Already fixed** — `junction_windows_test.go:34`
  now `t.Fatalf`s. Verified directly; no scope expansion needed.
- The `DISCOVERY-STATUS: LOOKUP-UNAVAILABLE` token on three entries claimed no
  stash query/search command exists. **Falsified** at backlogit 1.10.1
  (`stash get`, `stash list`, `search` all present).

## Notes for the next cycle

- `complexity` is **not** a defined field for type `task` in this workspace, so it
  is recorded as labeled prose in `implementation-notes`. `size` IS structured in
  frontmatter `custom_fields`. Worth fixing the type config.
- Staging artifacts are **committed locally on `main` and NOT pushed**; no PR was
  opened (Stage role boundary forbids creating/pushing PRs).
- Ship should claim **015-S**, not 016-F.

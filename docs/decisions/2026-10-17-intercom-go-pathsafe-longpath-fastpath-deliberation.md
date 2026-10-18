---
title: "pathsafe addLongPathPrefix byte-length fast-path guard: deliberation"
date: 2026-10-17
stash_ids: [6B751D8B, 475E76D2]
status: accepted-pending-operator-confirmation
---

## Problem frame

The operator chartered stash group A (6B751D8B, 475E76D2) for one cycle: the
`internal/pathsafe` `addLongPathPrefix` device-namespace exclusion branch and a
`utf16Len` fast-path guard.

Verified against source at `fc60c3b` and against the stash dispositions:

| Charter item | State | Evidence |
|---|---|---|
| 6B751D8B item 1, device-namespace branch | Already landed (017.002-T) | Branch present at `reparse_windows.go`, locked by `TestAddLongPathPrefixDeviceNamespaceBranchReachability` |
| 6B751D8B item 2, UTF-16-aware threshold | Already landed (017.003-T) | `utf16Len` exists; threshold reads `utf16Len(path) < longPathThreshold` |
| 6B751D8B items 3 and 4 | Deferred, triggers unfired | Pooling and CI lint scoping; not chartered this cycle |
| 475E76D2 byte-length fast path | Not implemented; accept-as-is twice (2026-09-11, 2026-09-18) | Deferred on value, never on safety |

The only executable residual inside the charter is the 475E76D2 guard.

## Options

* A. Ship nothing: both chartered items are landed or deferred on value.
* B. Ship the 475E76D2 guard only, as one small task, treating the explicit
  operator charter as the fresh justification the stash disposition requires.
* C. Also add a panic or unreachability assertion for the device branch.
  Rejected by name: the premise is false (branch is reachable), and the
  compound record on defensive scaffolding over an unverified premise applies.

## Soundness

`len(path) >= utf16Len(path)` holds for ASCII, 2-, 3- and 4-byte UTF-8 and for
invalid UTF-8 (replacement rune counts once as 1 unit, 3 bytes). Therefore
`len(path) < longPathThreshold` implies `utf16Len(path) < longPathThreshold`:
an early return before the scan cannot produce a false negative. The guard is
behavior-preserving; it changes cost only. Re-proved by the Correctness
Reviewer in the 2026-09-11 cycle.

## Decision

D-PS-1: Option B. One task: add
`if len(path) < longPathThreshold { return path }` ahead of the `utf16Len`
call, with a table characterization test proving verdict equivalence between
guarded and unguarded behavior across ASCII, BMP, and astral inputs on both
sides of the threshold, including the byte-length-above/unit-count-below band.

D-PS-2: The device-namespace branch gets no task. It is landed and locked.
Charter item 1 is satisfied by existing work; no assertion is added (Option C).

D-PS-3: Honest framing. This is a micro-optimization on a path dominated by an
allocation and two Win32 syscalls. It is not a correctness or security fix.
The operator charter is the only justification; no profile exists. This
reverses two prior accept-as-is dispositions and is flagged for operator
confirmation.

D-PS-4: 6B751D8B stays active (items 3 and 4 remain its only residual).
475E76D2 is archived only after the shipment merges, not by Stage at harvest
time. Stage records both as consumed-by-reference in the shipment, and the
operator decides archival.

## Out of scope

Buffer pooling, CI lint scoping, any change to `canonicalizeReparse`, any
change to threshold semantics, any rendered-file edit.

## Promotion

`promote_to: plan`. Pending operator confirmation of D-PS-1 and D-PS-3.

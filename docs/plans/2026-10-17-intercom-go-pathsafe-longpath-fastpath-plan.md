---
title: "pathsafe addLongPathPrefix byte-length fast-path guard: implementation plan"
date: 2026-10-17
source: docs/decisions/2026-10-17-intercom-go-pathsafe-longpath-fastpath-deliberation.md
stash_ids: [6B751D8B, 475E76D2]
---

## Summary

Add a provably behavior-preserving byte-length early return to
`addLongPathPrefix` in `internal/pathsafe/reparse_windows.go`, locked by a
verdict-equivalence characterization test written first. No device-namespace
work (already landed, 017.002-T). No other behavior change.

Requires plan hardening: no

Rationale: the change is a single early return whose soundness is a one-line
inequality (`len(s) >= utf16Len(s)`), pre-proved and covered by an equivalence
test; it touches no trust boundary, adds no allocation, no API, no state.

## Gate

Blocked on operator confirmation of D-PS-1 and D-PS-3 in the deliberation
before Ship claims the shipment. Plan review: ADVISORY (Scope Boundary
Auditor, 2026-10-17; one P2 on oracle independence addressed, one P2 on value
evidence accepted as an explicit operator-confirmation gate, three P3 applied).

## Source of truth

* Deliberation: `docs/decisions/2026-10-17-intercom-go-pathsafe-longpath-fastpath-deliberation.md`
  (D-PS-1 to D-PS-4)
* Prior art: `2026-09-11-intercom-go-pathsafe-deferred-trigger-reassessment-deliberation.md` D-3

## Implementation units

### Unit 1: Equivalence characterization test (tests domain, test-first)

* Files: `internal/pathsafe/reparse_windows_test.go` (1 file)
* Add `TestAddLongPathPrefixByteLengthGuardEquivalence`: table-driven, asserting
  `addLongPathPrefix(in)` against an INDEPENDENT oracle,
  `len(utf16.Encode([]rune(in))) >= longPathThreshold` deciding whether the
  `\\?\` prefix is expected (not a copy of the function body). Reuse existing
  helpers; do not duplicate prefix-guard cases already covered by
  `TestAddLongPathPrefixVerdicts` and the device-namespace test.
* Cases (3 groups): (a) ASCII at 259, 260, 261 units; (b) BMP multibyte,
  U+00E9 x 200 = 400 bytes / 200 units (byte length above, unit count below);
  (c) astral, U+1F600 x 100 = 400 bytes / 200 units and x 130 = 520 bytes /
  260 units.
* Passes against current code; must keep passing after Unit 2.
* Acceptance: windows build-tag consistent with the file; every case matches the
  independent oracle; includes the byte-above/unit-below band; mutation check
  recorded in the task notes: temporarily using an unsound guard
  (`len(path) < 130`) makes the test fail, then revert.
* Posture: characterization-first. Size XS, complexity low.

### Unit 2: Add the guard (code domain)

* Files: `internal/pathsafe/reparse_windows.go` (1 file, 1 function)
* Insert `if len(path) < longPathThreshold { return path }` before the
  `utf16Len` call, with a one-to-two-line comment stating
  `len(path) >= utf16Len(path)` so the early return is verdict-neutral.
* Do not change `utf16Len`, the threshold, or later guard order.
* Acceptance: Unit 1 test passes unchanged; existing `TestUTF16LenMatchesStdlibEncoding`
  and device-namespace reachability tests pass; `GOOS=windows go vet ./internal/pathsafe/`
  succeeds; diff is limited to the guard and its short comment.
* Depends on Unit 1. Size XS, complexity trivial.

## Risks

* Reviewer perception of gold-plating: mitigated by D-PS-3 honest framing in the
  code comment (precision-neutral, cost-only).
* Windows-tagged test not run on non-Windows CI: Ship validates on the
  `GOOS=windows` lint step and the advisory Windows job; note for Ship.

## Out of scope

Buffer pooling (6B751D8B item 3), CI lint scoping (item 4), assertions on the
device branch, rendered-file edits, upstream-routed items (003/004/005-DL).

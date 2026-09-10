---
title: "Plan Review — Pathsafe Windows Canonicalization Symmetry (016-F / 015-S)"
date: 2026-09-10
status: accepted
agent: Stage
mode: DARK_MODE_ACTIVE
governs: 016-F / shipment 015-S
verdict: PASS
---

# Plan Review — Pathsafe Windows Canonicalization Symmetry

**Reviewed plan**: `docs/plans/2026-09-10-intercom-go-pathsafe-windows-canonicalization-symmetry-plan.md`
**Source deliberation**: `docs/decisions/2026-09-10-intercom-go-pathsafe-windows-canonicalization-symmetry-deliberation.md`
**Gate**: pre-harvest plan review (P-005), multi-persona adversarial
**Attempts**: 2 (attempt 1 FAIL on two P1 findings; attempt 2 PASS)

---

## 1. Reviewer panel

Four independent personas were run in parallel on **different underlying
models** to reduce correlated blind spots. Each was given the artifacts, the
source surface, and the hard scope contract, and each verified claims against
source rather than accepting the plan's assertions.

| Persona | Model | Verdict contribution |
|---|---|---|
| Correctness Reviewer | `gpt-5.6-sol` | 1 × P1, 4 × P2, 1 × P3 |
| Security Reviewer | `claude-opus-4.8` | 1 × P2 (conf 0.72), 2 × P3; **no P0/P1** |
| Scope Boundary Auditor | `gemini-3.8-flash` | 2 × P2, 1 × P3; scope PASS |
| Go Reviewer | `claude-opus-4.7` | 1 × P1, 4 × P2, 5 × P3 (several verified-precondition) |

## 2. Attempt 1 verdict: FAIL

Two P1 findings. Per the operator's dark-mode stop conditions, an unresolved
P0/P1 is a halt condition, so the plan was revised rather than harvested.

### P1-a (Correctness) — B3 was wrongly described as droppable

`NewRoot` currently obtains Go's internal long-path handling for free through
`filepath.EvalSymlinks`' `os.Lstat` walk. C2 routes it through
`canonicalizeReparse`'s raw `syscall.CreateFile`, which has none — the exact U5
defect. Landing C2 without B3 would introduce a **new** fail-closed regression
for long workspace roots, on the very control this shipment repairs.

**Resolution**: B3 promoted to a hard prerequisite of C2 (§7); original H3
superseded; landing C2 alone declared a STOP condition; joint revert required.

### P1-b (Go) — C2 silently changes `NewRoot`'s observable Windows error type

`filepath.EvalSymlinks` returns `*fs.PathError`; Windows `canonicalizeReparse`
returns a raw `syscall.Errno`. `errors.Is(…, fs.ErrNotExist)` still resolves,
so the 009-S suite — which pins `Kind` and `errors.Is`/`As`-into-`Errno` — would
stay green while any caller doing `errors.As(err, &*fs.PathError)` silently
stopped matching.

**Resolution**: C2/AC3 added, requiring either a caller audit or boundary
re-wrapping as `*fs.PathError`, with the chosen behavior pinned by a test.
H3c added.

## 3. Attempt 2 findings resolved

| # | Sev | Finding | Resolution |
|---|---|---|---|
| 1 | P1 | B3→C2 dependency | §7 hard edge; H3 superseded |
| 2 | P1 | Error-surface change | C2/AC3; H3c |
| 3 | P2 | No containment-direction test | **C1b added**; H3b |
| 4 | P2 | C1 covered only one failing transition | C1/AC4 requires both |
| 5 | P2 | Volume-GUID AC was a synthetic table test only | C2/AC6 end-to-end or explicit residual |
| 6 | P2 | B3 predicate under-specified | B3/AC2 pins it + boundary test |
| 7 | P2 | UNC long-path form unhandled | B3/AC5 explicit residual |
| 8 | P2 | B1 AC unfalsifiable; forced test seam | B1 reframed, no injection seam |
| 9 | P2 | POSIX postcondition doc asymmetry | B2/AC4 |
| 10 | P2 | `os.Stat` re-derivation single-branch | C2/AC5 both branches |
| 11 | P2 | Plan pre-declared PASS with no artifact | Removed; this artifact now exists |
| 12 | P3 | B2→C2 "hard dependency" over-claimed | **Retracted** in plan §7 and deliberation D-1 |
| 13 | P3 | Device-path / double-prefix negatives | B3/AC4 |
| 14 | P3 | U2 lock omitted | Declined with rationale, H9 |
| 15 | P3 | Caller-side strip left optional | B2/AC3 pins "retain" |

## 4. Notable verified preconditions (no action required)

Recorded because they were independently checked and should not be re-litigated:

- `syscall.LazyProc.Find()` transitively guards **both** the `kernel32.dll` load
  and the export lookup via `LazyDLL.Load()`'s internal `sync.Once`. No separate
  `modkernel32` guard and no added `sync.Once` are idiomatic.
- `golang.org/x/sys` is **not** already a dependency in `go.mod`, so D-3's
  "use `syscall`, add no module dependency" constraint is correctly grounded.
- `unsafe.Pointer(&buf[0])` in `getFinalPathNameByHandle` can never deref a
  zero-length slice; the planned edits do not touch that pointer discipline.
- The POSIX no-op claim for C2 is **verified**: `reparse_other.go`'s
  `canonicalizeReparse` *is* `filepath.EvalSymlinks`, and POSIX paths never
  carry `\\?\`, so `stripUNCPrefix` is deterministically a no-op.
- `createDirectoryJunction` already `t.Fatalf`s on missing `pwsh`
  (`junction_windows_test.go:34`) — the 013-S silent-skip finding is discharged
  and must not be re-opened.
- The F3 `ModeIrregular` strict-ancestor branch is untouched; the security
  review confirmed C2 does not reopen the 013-S bypass and is fail-closed on
  every traced path.

## 5. Scope verdict

**PASS.** The scope auditor independently re-ran
`scripts/check-write-path-precondition.sh` and its self-test, and enumerated all
17 non-test Go files under `internal/**` and `cmd/**`, confirming **zero**
filesystem write primitives. BF5DE670's deferral is therefore evidence-backed
and not a dodge. No task implements work belonging to BF5DE670 or to any of the
ten operator-excluded IDs. No unrelated refactoring is smuggled in. All eight
tasks fit the 2-hour budget heuristic.

The auditor raised one P3 noting that the deliberation *names* three excluded
IDs (37FAB8C2, 1C6C3B46, 6C24E2E4) in its duplicate-scan section.
**Assessed and rejected as a boundary violation**: P-021 C5(A) makes duplicate
detection **unconditional** and it cannot be performed without comparing the
scoped entries against the other active entries. Those entries were **read
only** — none was triaged, re-prioritized, edited, archived, harvested, or
planned. Recording the comparison result is what makes the clean scan auditable;
omitting it would make a performed scan indistinguishable from a skipped one.

## 6. Verdict

**PASS (attempt 2).** Both P1 findings are resolved, all P2 findings are either
resolved or converted into explicitly recorded residuals with triggers, and one
over-claimed dependency has been retracted in both the plan and the source
deliberation. No P0 findings were raised by any persona. The plan is cleared for
harvest into feature 016-F and shipment 015-S.

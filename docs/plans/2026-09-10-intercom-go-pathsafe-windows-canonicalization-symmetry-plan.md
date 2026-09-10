---
title: "Implementation Plan — Pathsafe Windows Canonicalization Symmetry and Windows Lint Coverage"
date: 2026-09-10
status: reviewed
feature: 016-F
---

# Implementation Plan — Pathsafe Windows Canonicalization Symmetry and Windows Lint Coverage

<!-- plan-review-attempt: 1 -->
<!-- plan-review-attempt: 2 — revised after multi-persona adversarial review; see §11 -->

**Date**: 2026-09-10
**Source deliberation**: `docs/decisions/2026-09-10-intercom-go-pathsafe-windows-canonicalization-symmetry-deliberation.md`
**Feature / Shipment**: 016-F / 015-S
**Stash entries covered**: 4104AF54, E428AB46, E4C5413F (BF5DE670 deferred, decision D-4)
**Requires plan hardening**: **yes** — NON-NEGOTIABLE containment control, a
security-relevant regression, a CI gate change, and Windows-only test evidence
**Constitution principles**: Principle IV (risk acceptances are explicit and tracked),
test-first (RED→GREEN) delivery
**Plan-review gate verdict**: see §10 — recorded only after the review artifact exists

---

## 1. Objective

Restore namespace symmetry between the two sides of `internal/pathsafe`'s
containment comparison, harden the Windows canonicalization helper introduced
by 013-S, and make the Windows-tagged production surface visible to the
repository's lint gate.

## 2. Root cause

013-S upgraded the **candidate** side of the containment check to
`canonicalizeReparse` (`GetFinalPathNameByHandleW`, junction-aware) and switched
the ancestor probe from `os.Stat` to `os.Lstat`, but left the **root** side on
`filepath.EvalSymlinks` (junction-unaware). The comparison at
`pathsafe.go:371` — `hasPathPrefix(real, root.path)` — therefore compares two
paths produced by two different canonicalization mechanisms.

Under a workspace root that is itself a directory junction, this is a
**fail-closed regression introduced by 013-S**: the configuration worked before
that shipment. Either `Lstat` on the root reports the junction's own
`ModeIrregular` and trips the unconditional ancestor rejection, or the candidate
resolves to its real target while `root.path` remains the unresolved junction
path — so every access under such a root fails, in the worse case as a
false-positive `symlinkEscapeMsg` asserting a "GENUINE, VERIFIED escape" for a
workspace that never escaped.

Separately, `reparse_windows.go` is `//go:build windows`, and Go file discovery
is GOOS-scoped, so the repository's only lint gate never analyzes it.

## 3. Surface

| File | Change |
|---|---|
| `.github/workflows/ci.yml` | Add a `GOOS=windows` golangci-lint step to the existing `lint` job |
| `internal/pathsafe/reparse_windows.go` | U5 long-path prefixing, U6 internal normalization, U7 non-panicking proc resolution |
| `internal/pathsafe/reparse_other.go` | Doc comment only — POSIX postcondition symmetry (B2/AC4) |
| `internal/pathsafe/reparse_windows_test.go` | RED tests for U5/U6/U7 |
| `internal/pathsafe/root.go` | `NewRoot` adopts `canonicalizeReparse`; risk-register reconciliation |
| `internal/pathsafe/root_windows_test.go` | Junction-rooted workspace regression lock |
| `internal/pathsafe/junction_windows_test.go` | Reuse existing `createDirectoryJunction` helper |
| `internal/pathsafe/pathsafe.go` | Doc-comment reconciliation only (no behavior change) |

**Not touched**: the F3 `ModeIrregular` strict-ancestor branch, any write path.

## 4. Key design decisions carried from deliberation

- **D-2**: `NewRoot` calls `canonicalizeReparse(abs)` then `stripUNCPrefix`.
  On non-Windows, `canonicalizeReparse` *is* `filepath.EvalSymlinks`
  (`reparse_other.go`), so this is a **provable no-op on POSIX** and the blast
  radius is confined to Windows.
- **D-3**: The Windows lint gate lands **blocking, not advisory**. Measured this
  cycle: `GOOS=windows golangci-lint run ./...` currently reports **0 issues**,
  so no remediation task is required and no advisory window is needed.
- **D-5**: No artifact or doc comment may claim `filepath.EvalSymlinks`
  internally calls `GetFinalPathNameByHandleW`. This is a twice-corrected known
  falsehood in this repository.
- **D-6**: `root_windows_test.go`'s `"volume guid"` case pins non-stripping of
  `\\?\Volume{GUID}` as **intended**; U6 must preserve it.

## 5. Verified preconditions (measured at `db4edf0`, this session)

| Claim | Evidence |
|---|---|
| Ubuntu lint cannot see the Windows file | `GOOS=linux go list -f '{{.GoFiles}}' ./internal/pathsafe` → `[lexical.go pathsafe.go reparse_other.go root.go]`; `reparse_windows.go` absent |
| `GOOS=windows` is a sufficient, minimal fix | `GOOS=windows` → `[lexical.go pathsafe.go reparse_windows.go root.go]` |
| Gate can land green | `GOOS=windows golangci-lint run ./...` → `0 issues`, exit 0 |
| Cross-GOOS analysis is sound on this toolchain | `GOOS=windows go vet ./internal/pathsafe/...` → exit 0 |
| No write path exists (justifies D-4 deferral) | `scripts/check-write-path-precondition.sh` exit 0; `--self-test` 5/5 fixtures pass; independent grep → 0 matches |
| `createDirectoryJunction` fails loud | `junction_windows_test.go:34` `t.Fatalf` on missing `pwsh` — the 013-S silent-skip finding is already discharged |

## 6. Task breakdown (test-first)

### Sub-epic A — Windows lint coverage (4104AF54)

**A1. Add `GOOS=windows` golangci-lint step to the CI lint job**
Domain: config. Files: `.github/workflows/ci.yml` (1).
Add a step to the existing `lint` job running `golangci-lint run ./...` with
`env: GOOS: windows`, reusing the already-pinned golangci-lint install. No new
runner is required — cross-GOOS analysis runs on the existing `ubuntu-latest`
runner.
- AC1: A step exists in job `lint` that invokes golangci-lint with `GOOS=windows`.
- AC2: The step is blocking (no `continue-on-error`).
- AC3: `python -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml', encoding='utf-8'))"` succeeds **and** `actionlint .github/workflows/ci.yml` passes. Both are required; the 010-S compound learning records that neither alone suffices.
- AC4: The step passes on the unmodified tree (0 issues), demonstrating it is green-on-arrival rather than requiring an advisory window.

### Sub-epic B — `canonicalizeReparse` hardening (E4C5413F)

**B1. U7 — resolve `GetFinalPathNameByHandleW` without a panic path**
Domain: code. Files: `reparse_windows.go` (1).

**Revised after review (scope audit P2 ×2).** The original AC demanded a RED
test for an unresolvable proc. That condition is unreachable on every supported
target (`GetFinalPathNameByHandleW` has shipped in `kernel32.dll` since Vista /
Server 2008; the Go 1.24 floor requires Windows 10 / Server 2016+), so the AC
was **unfalsifiable without inventing a test-injection seam in production code**
— itself a verification hazard on a NON-NEGOTIABLE control. E4C5413F's own text
classifies U7 as "theoretical only".

Reframed: land the defensive guard **without** the unfalsifiable RED test,
matching the package's existing convention for documented-unreachable defensive
branches.
- AC1: `procGetFinalPathNameByHandleW.Find()` is called and its error wrapped
  as `apperr.KindPathViolation` rather than allowed to panic through
  `LazyProc.Call`'s `mustFind`.
- AC2: The `.Find()` guard is placed at the **top of `canonicalizeReparse`,
  before `syscall.CreateFile`**, so the failure path never allocates a kernel
  handle. (Go review: `LazyProc.Find()` transitively covers both the
  `kernel32.dll` load and the export lookup via `LazyDLL.Load()`'s internal
  `sync.Once`, so no separate `modkernel32` guard and no added `sync.Once` are
  needed.)
- AC3: The branch carries a `DOCUMENTED-UNREACHABLE` annotation consistent with
  the package's existing accepted defensive branches, naming why it cannot fire
  on a supported target.
- AC4: Existing `reparse_windows_test.go` tests still pass; **no test-only
  injection seam is added to production code.**

**B2. U6 — internalize the `stripUNCPrefix` postcondition**
Domain: code. Files: `reparse_windows.go`, `reparse_other.go`, `reparse_windows_test.go` (3).
`canonicalizeReparse` applies `stripUNCPrefix` itself so the postcondition holds
for **any** caller.
- AC1: RED first — a test asserts `canonicalizeReparse`'s own return value is
  already normalized.
- AC2: `\\?\Volume{GUID}` non-stripping is preserved (D-6); the existing
  `"volume guid"` case in `root_windows_test.go` still passes unmodified.
- AC3: `checkSymlinkEscape`'s observable behavior is unchanged. **Decision
  (Go review P3): the caller-side `real = stripUNCPrefix(real)` at
  `pathsafe.go:369` is RETAINED as belt-and-suspenders, not removed** —
  `stripUNCPrefix` is idempotent (verified: unprefixed input returns unchanged),
  and retaining it avoids coupling `pathsafe.go` to a Windows-only internal
  contract.
- AC4: `reparse_other.go`'s doc comment is updated (comment-only) to state that
  the same postcondition holds trivially on POSIX, so a reader of either file
  alone sees the identical package-level contract and cannot later wire a
  POSIX canonicalizer that silently breaks the U6 invariant.

**B3. U5 — long-path (`>MAX_PATH`) prefixing before `CreateFile`**
Domain: code. Files: `reparse_windows.go`, `reparse_windows_test.go` (2).

**PROMOTED TO A HARD PREREQUISITE OF C2 by review (correctness P1).** See §7.
- AC1: RED first — a deeply-nested in-root ancestor exceeding `MAX_PATH`
  resolves successfully instead of being spuriously rejected.
- AC2: **The prefixing predicate is pinned explicitly in code and comment**
  rather than left ad hoc (Go review P2). It applies `\\?\` only when the path
  is absolute, is **not already** `\\?\`-prefixed, is **not** a `\\.\`
  device-namespace path, and is not a bare UNC path; and it documents the
  `MAX_PATH` threshold it keys on. A boundary test covers the threshold.
- AC3: Ordinary short paths are unaffected; the prefix/strip round-trip with B2
  is verified.
- AC4: **Negative coverage for the classes review flagged (security P3,
  correctness P2)**: an already-extended `\\?\Volume{GUID}\...` input is **not**
  double-prefixed, and a `\\.\` device path is **not** prefixed.
- AC5: **The UNC long-path form is an explicitly recorded residual.** Converting
  `\\server\share\...` to `\\?\UNC\server\share\...` is deliberately NOT
  implemented in this shipment; a risk-register entry records the residual and
  its trigger. (Prefixing `\\?\` also disables Win32 path normalization, so it
  must be applied only after the caller's `Abs`/`Clean` — true at both call
  sites, and stated as a precondition comment.)

### Sub-epic C — Root/candidate namespace symmetry (E428AB46)

**C1. RED — junction-rooted workspace regression lock (availability direction)**
Domain: test. Files: `root_windows_test.go` (1), reusing
`createDirectoryJunction`.
- AC1: A test constructs a workspace root that is **itself** a live directory
  junction, calls `NewRoot`, then `Resolve`.
- AC2: The test FAILS against current `main`, reproducing the fail-closed
  regression (this is the RED proof; a test that passes pre-fix is not evidence).
- AC3: Per the 011-S compound learning, the junction is the **exact reported
  shape** (root itself is the junction), not a subpath through one — a
  structurally different repro exercises an already-fixed branch and yields a
  false negative.
- AC4: **Both failing transitions are covered (correctness P2)**, because §2
  identifies two distinct ones and a single fixture can only exercise one:
  (a) an **existing** in-root descendant, which reaches `canonicalizeReparse`
  and fails the namespace comparison; and (b) a **non-existent** leaf, which
  ascends to the junction root and is rejected by the `ModeIrregular` branch.

**C1b. RED — junction-rooted workspace still REJECTS escapes (containment direction)**
Domain: test. Files: `root_windows_test.go` (1).

**Added by review (security P2, confidence 0.72).** C1 alone proves only that
C2 does not **over**-reject. On a NON-NEGOTIABLE containment control, the plan
must also prove C2 does not **under**-reject — that is the property the control
exists to guarantee.
- AC1: Under a junction-rooted workspace, an in-root symlink/junction whose
  target is **outside** the resolved root is still rejected by
  `hasPathPrefix(real, root.path)` at `pathsafe.go:~371`.
- AC2: The rejection surfaces as `apperr.KindPathViolation`.
- AC3: The test is written to pass both before and after C2 (it is a
  **containment-preservation lock**, not a RED→GREEN unit): a regression here
  at any point is a STOP condition.

**C2. GREEN — `NewRoot` adopts `canonicalizeReparse`**
Domain: code. Files: `root.go` (1).
Replace `filepath.EvalSymlinks(abs)` with `canonicalizeReparse(abs)`, retaining
`stripUNCPrefix`.
- AC1: C1's tests pass; C1b still passes.
- AC2: The 009-S `apperr.Wrapf` dual-discriminability guarantees on `NewRoot`'s
  failure branches are preserved (verified by the existing 009-S tests). **If
  any 009-S test requires modification, that is a STOP condition requiring
  re-deliberation, not a test edit.**
- AC3: **Error-surface preservation (Go review P1).** `filepath.EvalSymlinks`
  returns `*fs.PathError`; Windows `canonicalizeReparse` returns a raw
  `syscall.Errno` from `CreateFile` or the lazy-proc call. `errors.Is(…,
  fs.ErrNotExist)` still resolves (Windows `Errno.Is` maps
  `ERROR_FILE_NOT_FOUND`/`ERROR_PATH_NOT_FOUND`), but any caller doing
  `errors.As(err, &*fs.PathError)` would **silently stop matching** on Windows.
  The implementation MUST either (a) audit `NewRoot` callers and record that
  none extract `*fs.PathError`, or (b) wrap the Win32 error as a `*fs.PathError`
  (`Op`, `Path`, `Err`) at the boundary to preserve the observable surface.
  Option (b) is preferred. A test pins the chosen behavior.
- AC4: Non-directory root rejection is preserved.
- AC5: The `DOCUMENTED-UNREACHABLE` coverage-excluded justification on the
  subsequent `os.Stat` is **re-derived**, not re-pasted, and **covers both
  build branches** (Go review P2): Windows proves existence via
  `CreateFile`/`OPEN_EXISTING` (strictly stronger than an `Lstat` walk), POSIX
  via `EvalSymlinks`' own walk. It must also note that `os.Stat` now runs on a
  `stripUNCPrefix`-normalized form, which is why D-6's `\\?\Volume{GUID}`
  non-stripping is load-bearing rather than incidental.
- AC6: **End-to-end volume-GUID coverage (correctness P2).** The existing
  `"volume guid"` case is a synthetic `stripUNCPrefix` table test that never
  exercises `NewRoot`/`Resolve`/`hasPathPrefix`. Because `hasPathPrefix` is a
  string comparison, a root/candidate namespace mismatch under a GUID-volume
  root would be invisible to it. Add an end-to-end assertion (or, if a
  GUID-volume fixture is not constructible in the test environment, record the
  gap explicitly as a risk-register residual with its trigger — silence is not
  acceptable).
- AC7: Full POSIX suite passes unchanged (expected: no-op on non-Windows).

**C3. Risk-register and doc-comment reconciliation**
Domain: docs. Files: `root.go`, `pathsafe.go` (2, comments only).
- AC1: The package risk register records the asymmetry as **closed**, with the
  013-S-regression provenance.
- AC2: No text claims `EvalSymlinks` calls `GetFinalPathNameByHandleW` (D-5).
- AC3: Statements about the F3 `ModeIrregular` deferral remain accurate.

## 7. Dependency order

```
A1 (lint gate)  ──► must land first so all later Windows code is linted
      │
      ├─► B1 (U7)
      ├─► B2 (U6) ──► C2   [RECOMMENDED ordering — see correction below]
      └─► B3 (U5) ══► C2   [HARD prerequisite — see P1 below]

      C1 (RED, availability) ──┐
      C1b (containment lock) ──┼─► C2 (GREEN) ──► C3 (docs)
```

### B3 → C2 is a HARD prerequisite (correctness review, P1 — MAJOR)

`NewRoot` currently uses `filepath.EvalSymlinks`, whose `os.Lstat` walk applies
Go's own internal long-path handling. After C2, `NewRoot` routes through
`canonicalizeReparse`, whose raw `syscall.CreateFile` call has **no** long-path
handling — which is precisely the U5 defect. **Landing C2 without B3 therefore
introduces a NEW fail-closed regression for long workspace roots**, on the same
control this shipment exists to repair.

This supersedes the original H3, which wrongly described B3 as independently
droppable. B3 is now either a hard prerequisite of C2, or — if the `>MAX_PATH`
fixture proves unconstructible — C2 must not land until the resulting `NewRoot`
limitation is explicitly accepted in the risk register **and** covered by a
test. Silently dropping B3 while landing C2 is a STOP condition.

### Correction: B2 → C2 is NOT a hard dependency (correctness review, P3)

The source deliberation (D-1) asserted that C2 creates the "second caller" U6
warns about, making B2 a hard prerequisite. **Review falsified this and the
plan records the correction rather than preserving a tidier narrative.**

C2 explicitly retains `NewRoot`'s own caller-side `stripUNCPrefix`, and
`checkSymlinkEscape` already strips at `pathsafe.go:369`. `stripUNCPrefix` was
read and confirmed idempotent: unprefixed input returns unchanged, drive/UNC
forms have the prefix removed, and non-strippable `Volume{GUID}`/`GLOBALROOT`
forms return the original unchanged. **Both callers are therefore normalized
even if C2 precedes B2**, so C2 does not introduce U6's predicted defect.

B2 remains a genuine encapsulation improvement and its recommended ordering is
retained, but it is **not** load-bearing for correctness, and the grouping
rationale for E4C5413F rests on the B3→C2 dependency above plus shared file
surface — not on this retracted claim.

## 8. Evidence plan

- POSIX: `go test ./...` on the CI ubuntu runner (proves the C2 no-op claim).
- Windows: local `go test ./internal/pathsafe/... -v` on a Windows host, with
  **zero SKIP lines** confirmed. Windows CI is disclosed repo-wide as
  ADVISORY-ONLY, so the primary Windows evidence path is the local run.
- CI workflow: `yaml.safe_load` **and** `actionlint`, both mandatory (A1/AC3).
- Lint: `GOOS=windows golangci-lint run ./...` green.

## 9. Plan Hardening

**Hardening required: yes.** This plan touches a NON-NEGOTIABLE containment
control and a CI gate.

**H1 — Risk: C2 changes root construction semantics on Windows.**
`canonicalizeReparse` opens a handle via `CreateFile`/`OPEN_EXISTING`, whereas
`EvalSymlinks` performs an `Lstat` walk. Both require the root to exist, but the
failure *modes* differ (Win32 error vs. `fs.PathError`).
*Mitigation*: AC2 pins the 009-S `apperr.Wrapf` discriminability guarantees via
the existing 009-S test suite, which must pass unmodified. If any 009-S test
requires modification, that is a STOP condition requiring re-deliberation, not a
test edit.

**H2 — Risk: C2's `os.Stat` unreachability justification silently becomes
false.** The current comment is a coverage-exclusion annotation grounded in
`EvalSymlinks`' behavior.
*Mitigation*: AC4 requires re-derivation. `CreateFile`/`OPEN_EXISTING` proves
existence at least as strongly, but the reasoning must be rewritten, and if it
cannot be re-derived the coverage exclusion must be removed rather than left
mis-justified.

**H3 — SUPERSEDED by correctness review P1.** The original H3 described B3 as
independently revertible and non-blocking for C2. That was **wrong**: once C2
routes `NewRoot` through `canonicalizeReparse`, the U5 long-path defect applies
to workspace-root construction, where `EvalSymlinks` previously supplied Go's
long-path handling. B3 is now a hard prerequisite of C2 (§7).
*Residual risk*: constructing a `>MAX_PATH` fixture remains environment-sensitive.
*Mitigation*: if the fixture proves unstable, C2 is **held** and the pair is
either landed together with an explicit accepted limitation in the risk register
plus a covering test, or both are deferred. Landing C2 alone is a STOP condition.

**H3b — Risk: the containment control is proven only in the availability
direction.** A change to how `root.path` is derived on a NON-NEGOTIABLE control
must prove it does not under-reject, not merely that it does not over-reject.
*Mitigation*: task C1b adds an explicit containment-preservation lock; a
regression there is a STOP condition.

**H3c — Risk: C2 silently changes `NewRoot`'s observable Windows error type**
from `*fs.PathError` to raw `syscall.Errno`, breaking any
`errors.As(err, &*fs.PathError)` caller without failing the 009-S suite (which
pins `Kind` and `errors.Is`/`As`-into-`Errno`, not `*fs.PathError` extraction).
*Mitigation*: C2/AC3 requires either a caller audit or boundary re-wrapping,
with the chosen behavior pinned by a test.

**H4 — Risk: the Windows lint gate is green today but could surface findings
once B1–B3 and C2 land.**
*Mitigation*: A1 lands first precisely so that every subsequent task is linted
as it is written, converting a potential end-of-shipment surprise into
per-task feedback.

**H5 — Risk: a CI workflow edit passes local review but breaks real CI.**
The 010-S compound learning records a YAML-valid but Actions-schema-invalid edit
that created zero jobs and was missed by 9-persona review.
*Mitigation*: A1/AC3 makes `actionlint` mandatory and non-substitutable.
Additionally, A1 adds a step to an **existing** job rather than creating a new
job, which avoids the job-boundary defect class entirely.

**H6 — Risk: C1's test passes pre-fix and yields false assurance.**
*Mitigation*: AC2 explicitly requires the test to FAIL against current `main`,
and AC3 pins the exact-shape rule from the 011-S compound learning.

**H7 — Risk: scope creep into TOCTOU/hardlink mitigation (BF5DE670).**
*Mitigation*: Declared anti-goal (§6 of the deliberation). No write API is
introduced; `scripts/check-write-path-precondition.sh` remains the mechanical
watch and must still pass at the end of the shipment.

**H8 — Rollback**: every task is independently revertible **except the
B3+C2 pair, which must be reverted together** (§7). A1 is an additive CI step.
There is no data migration and no persisted state.

**H9 — Declined findings, recorded rather than silently dropped.**
- *U2 regression lock* (live out-of-root junction as intermediate with a
  non-existent leaf), raised by the security review at P3 and by the 013-S
  review. **Declined as out of scope**: U2 is not covered by any of the four
  chartered stash entries, and this plan preserves F3 unrelaxed as an anti-goal,
  so the shape U2 guards remains rejected. The scope audit independently
  confirmed that pulling it in would be scope creep. Recorded here so the gap
  is visible to whoever next relaxes F3.
- *UNC long-path conversion* (`\\?\UNC\server\share`), correctness review P2.
  Declined for this shipment; recorded as an explicit risk-register residual
  under B3/AC5 rather than left as an undocumented omission.

## 10. Plan-review gate

**Verdict: PASS (attempt 2).** Recorded in
`docs/decisions/2026-09-10-intercom-go-pathsafe-windows-canonicalization-plan-review.md`,
which exists and is committed alongside this plan. Attempt 1 returned FAIL on
two P1 findings (B3→C2 hard dependency; `NewRoot` error-surface change); both
are resolved in this revision, together with seven P2 findings and the
retraction of the over-claimed B2→C2 dependency.

## 11. Revision log

| Rev | Change | Source |
|---|---|---|
| 1 | Initial plan | — |
| 2 | B3 promoted to hard prerequisite of C2; H3 superseded | Correctness P1 |
| 2 | C2/AC3 error-surface preservation added; H3c added | Go review P1 |
| 2 | C1b containment-preservation lock added; H3b added | Security P2 |
| 2 | C1/AC4 both failing transitions required | Correctness P2 |
| 2 | C2/AC6 end-to-end volume-GUID coverage required | Correctness P2 |
| 2 | B3 predicate pinned; device-path and double-prefix negatives added | Go P2, Security P3 |
| 2 | B3/AC5 UNC long-path residual recorded | Correctness P2 |
| 2 | B1 reframed to remove unfalsifiable RED and injection-seam hazard | Scope audit P2 ×2 |
| 2 | B2/AC3 belt-and-suspenders decision pinned; AC4 POSIX doc symmetry | Go P3, P2 |
| 2 | C2/AC5 re-derivation must cover both build branches | Go P2 |
| 2 | B2→C2 "hard dependency" retracted as over-claimed | Correctness P3 |
| 2 | Pre-declared PASS removed; verdict recorded only after artifact exists | Correctness P2 |
| 2 | H9 added to record declined findings | Security P3 |

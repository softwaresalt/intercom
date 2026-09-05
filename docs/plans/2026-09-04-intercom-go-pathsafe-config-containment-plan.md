---
title: "Implementation Plan — pathsafe and config containment hardening (010-F)"
date: 2026-09-04
status: reviewed
phase: residual-hardening
feature: 010-F
source_deliberation: "docs/decisions/2026-09-04-intercom-go-residual-hardening-triage-deliberation.md"
resolves_stash: [4A91CA81, 3D61B5A8, 7028FBE7]
partially_resolves_stash: [7774C9CA]
depends_on: 009-F
---

# Implementation Plan — pathsafe and config containment (010-F)

## Problem Frame

`internal/pathsafe` is the workspace path-containment security control
(Constitution IV). Four deferred findings target it and its one duplicated
consumer.

1. **`4A91CA81` — `NewRoot` drops the underlying cause.** Both the
   `filepath.Abs` and `filepath.EvalSymlinks` failure branches construct via
   `apperr.New`, discarding `err`. Callers cannot use `errors.Is`/`errors.As` to
   distinguish `os.ErrNotExist` from `os.ErrPermission` for a bad workspace
   root — a materially different operator diagnosis (typo vs. permissions).
2. **`3D61B5A8` — three narrow correctness gaps.** (a) `hasPathPrefix`
   (`pathsafe.go:166`) builds `prefix + separator`, which becomes a **doubled**
   separator when the root is exactly the filesystem root (`/`), causing valid
   resolutions to be rejected. (b) `NewRoot` does not verify its canonicalised
   target is a **directory** — pointing a workspace root at a regular file
   succeeds silently. (c) The documented GO-14 behaviour (a dangling symlink at
   the final component is treated as non-existent via `os.Stat`) has **no
   regression test** pinning it.
3. **`7028FBE7` — duplicated traversal logic.** `internal/config/validate.go`
   rule 7 uses a private `containsDotDotSegment` string scanner for
   `database.path`, instead of `pathsafe`'s tested containment logic
   (drive-prefix rejection, rooted-path detection, symlink-escape checks). Two
   implementations of one security concept is exactly the composability defect
   operator rule 6 prioritises removing.
4. **`7774C9CA` (pathsafe part)** — `splitPath` (`pathsafe.go:78`) is a
   redundant one-line wrapper around `filepathSplitList` (`pathsafe.go:109`).

## Decision — how to resolve `7028FBE7` without over-widening the surface

`pathsafe`'s exported surface is currently just `Root`, `NewRoot`, `Root.Path`,
`Root.Resolve`. The stash entry correctly flags that naive reuse would widen it.

**Chosen:** export a single narrow, side-effect-free predicate for
**lexical traversal-segment detection** — the one concept `validate.go` actually
needs — and delete `containsDotDotSegment`.

Rejected alternatives:

* *Have `validate.go` call `Root.Resolve`.* Wrong semantics. `database.path` is
  validated **before** a root is established and may legitimately not exist yet;
  `Resolve` performs filesystem and symlink work that rule 7 must not require.
  It would also convert a pure validation rule into an I/O-performing one.
* *Export the whole normalisation pipeline.* Widens the security surface far
  beyond the need, contradicting the entry's own concern.
* *Move rule 7 into `pathsafe`.* Puts config-schema policy inside a
  general-purpose primitive, violating Principle VI.

The exported predicate is **lexical only** — no filesystem access — which keeps
it safe to call on a not-yet-existing path and keeps `pathsafe`'s I/O-performing
surface unchanged.

## Non-Goals

* Does not change `Root.Resolve`'s containment semantics or the symlink-escape
  algorithm hardened in 002-S.
* Does not alter which config fields are validated, or rule 7's verdict for any
  input that currently passes or fails (behaviour-preserving refactor plus the
  strictly-additional checks `pathsafe` brings).
* Does not touch `internal/apperr`'s taxonomy — 009-F owns that and lands first.

## Constitution Check

| Principle | Status | Units | Justification |
|---|---|---|---|
| I. Safety-First Go | Satisfied | U1–U6 | Error causes preserved; no new unsafe constructs. |
| II. Test-First Development (NON-NEGOTIABLE) | Satisfied | U1, U2, U3, U4, U6 | Each behaviour-affecting unit carries a test written first; U5 is a zero-behaviour deletion covered by existing tests. |
| III. Workspace Isolation and Security Boundaries | Satisfied | U2, U6 | Directory validation tightens the boundary; U6 removes a duplicate implementation without changing verdicts. |
| IV. CLI Workspace Containment (NON-NEGOTIABLE) | Satisfied | U1–U6 | This package *is* the containment control; U3 AC-4 and U6 AC-3/AC-4 exist specifically to prevent weakening it. |
| V. Structured Observability | Satisfied | U1 | Cause preservation is the observability fix. |
| VI. Single Responsibility | Satisfied | U6 | Config policy stays in `config`; path primitives stay in `pathsafe`. |
| VII. Destructive Command Approval (NON-NEGOTIABLE) | N/A | — | No destructive command is executed. |
| VIII. Explicit Safety Modes for Elevated Risk | Satisfied | U6 | The behaviour-changing variant of U6 was rejected; the remaining zero-delta move carries a blocking stop condition (AC-4) rather than proceeding on judgement. |
| IX. Git-Friendly Persistence | Satisfied | all | Source only. |
| X. Agent Context Efficiency | Satisfied | all | Small units; U5 folded into U3 to avoid a trivial standalone cycle. |
| XI. Merge Commit History Preservation (NON-NEGOTIABLE) | Satisfied | — | None. |

## Implementation Units

### U1 — Preserve the underlying cause in `NewRoot`
*Resolves `4A91CA81`. Artifact class: Go code + tests. Size S, complexity low. **Depends on 009-F U5 (`Wrapf`).***

Change both `NewRoot` failure branches to construct a `KindPathViolation` error
via **`apperr.Wrapf`** (added by 009-F U5), which sets the formatted message
**and** retains the cause.

**Adversarial review corrected the original premise here.** The first draft
claimed `apperr.Wrap` already sufficed. It does not: `Wrap(kind, cause)`
(`apperr.go:142`) takes **no message parameter** and derives `msg` from
`cause.Error()`, while `New(kind, msg)` **drops** the cause. `Error`'s fields
are unexported, so `pathsafe` cannot construct one directly. Using `Wrap` alone
would render `"path violation: <raw os error>"`, dropping the
`"workspace root invalid"` context that `TestNewRootOnMissingDirReturnsPathViolation`
(`root_test.go:38`) asserts — making AC-3 and AC-4 below mutually unsatisfiable.
`Wrapf` resolves this and is the reason 009-F must land first.

**Acceptance criteria**
1. `errors.Is(err, os.ErrNotExist)` is true for a missing workspace root, and
   `errors.Is(err, os.ErrPermission)` is true for a permission failure.
2. `errors.Is(err, apperr.ErrPathViolation)` remains true for both branches
   (dual-discriminability is the whole point — neither may regress).
3. The error message retains its existing `"workspace root invalid: …"` context.
4. `TestNewRootOnMissingDirReturnsPathViolation` passes **unchanged**.
5. If 009-F U5 is unavailable for any reason, **return this unit blocked** —
   do not substitute `Wrap` or `New`, both of which silently lose half the
   contract.

### U2 — `NewRoot` must reject a non-directory target
*Resolves `3D61B5A8` (part 1 of 3). Artifact class: Go code + tests. Size S, complexity low.*

After canonicalisation, stat the target and reject with `KindPathViolation` when
it is not a directory.

**Acceptance criteria**
1. `NewRoot` pointed at a regular file returns a `KindPathViolation` error
   naming the path and the reason.
2. `NewRoot` on a valid directory is unchanged, including the existing
   idempotence test.
3. `NewRoot` on a symlink **to** a directory still succeeds (resolution happens
   before the type check — order matters, and a test pins it).

### U3 — Fix `hasPathPrefix` doubled separator at filesystem root
*Resolves `3D61B5A8` (part 2 of 3). Artifact class: Go code + tests. Size S, complexity medium.*

Avoid appending a separator when the prefix already ends with one (the
filesystem-root case), so containment is evaluated correctly when the root is
`/` (or a drive root such as `C:\`).

**Acceptance criteria**
1. With the root at the filesystem root, a valid in-root candidate **resolves**
   rather than being rejected.
2. Escape detection is unaffected for that configuration — a traversal attempt
   from a filesystem root is still rejected (the fix must not trade a false
   negative for a false positive).
3. All existing containment tests pass unchanged.
4. **The fix must be conditional** — omit the separator only when the prefix
   already ends with one. An unconditional drop would make `/p/ws` match
   `/p/ws-evil`, a false-accept containment breach. Review confirmed the
   sibling-escape guard is preserved by the conditional form only.
5. The gated case is the POSIX filesystem root (`/`), which runs on the
   `ubuntu-latest` PR CI. A Windows drive-root (`C:\`) case is added as a
   `runtime.GOOS == "windows"` extra and is **not** shipment-blocking, since PR
   CI does not run Windows.

### U4 — Regression test for GO-14 dangling-symlink behaviour
*Resolves `3D61B5A8` (part 3 of 3). Artifact class: tests. Size S, complexity low.*

Pin the documented behaviour: a dangling/broken symlink at the **final** path
component is treated as non-existent and accepted.

**Acceptance criteria**
1. A test creates a dangling symlink as the final component and asserts the
   documented accept behaviour.
2. The test is distinct from the existing symlinked-**intermediate**-directory
   escape test, which must continue to **reject** (the two cases are easily
   conflated and must be independently pinned).
3. Skips explicitly where symlink creation is unavailable.

### U5 — Remove the redundant `splitPath` wrapper
*Partially resolves `7774C9CA`. Artifact class: Go code. Size XS, complexity trivial. **Landed in the same commit as U3.***

Delete `splitPath` (`pathsafe.go:78`) and call `filepathSplitList`
(`pathsafe.go:109`) directly. Too small to justify a standalone review cycle, so
it rides with U3, which already modifies this file.

**Acceptance criteria**
1. No redundant one-line wrapper remains.
2. Zero behaviour change; all existing `pathsafe` tests pass unmodified.
3. No new exported surface is introduced by this unit.

### U6 — De-duplicate traversal detection with ZERO verdict change
*Resolves `7028FBE7`. Artifact class: Go code + tests. Size S, complexity low. Depends on U3.*

**Adversarial review rejected the original design and it has been replaced.**
The first draft proposed exporting a `normalize`-based predicate and accepted a
"stricter" behaviour delta. Review proved that delta is **bidirectional and
includes a containment regression**:

* `containsDotDotSegment` (`validate.go`) splits on **both** `/` and `\`
  **regardless of GOOS**. A `pathsafe.normalize`-based predicate uses
  GOOS-dependent separators, so on Linux `..\..\etc` collapses to a single
  literal component that is **not** `..` — a traversal string currently
  **rejected** would newly be **accepted**. That is a weakening of a
  NON-NEGOTIABLE Constitution IV control.
* Absolute `database.path` values are **deliberately accepted** today
  (`validate.go` rule-7 comment, pinned by
  `TestValidateRule7AbsoluteDatabasePathAccepted` in `paths_test.go`). A
  `normalize`-based predicate rejects them, breaking a shipped operator-facing
  contract and contradicting AC-3.

**Revised design — a pure move, not a unification.** Relocate the *exact
existing logic* of `containsDotDotSegment` into `internal/pathsafe` as a single
exported, lexical, cross-platform helper, and have `validate.go` call it. The
logic is copied verbatim: it keeps splitting on both `/` and `\` on every
platform, keeps rejecting only literal `..` segments, and keeps accepting
absolute paths.

This still achieves the finding's actual goal — one implementation of the
concept instead of two — while guaranteeing no verdict changes. Unifying the
*semantics* of the two predicates is a genuinely separate decision with
operator-facing migration consequences; it is explicitly **not** attempted here
and is recorded as a follow-up for a future cycle.

**Acceptance criteria**
1. `internal/config/validate.go` no longer defines its own traversal detector
   and calls the `pathsafe` helper instead.
2. Exactly **one** new symbol is exported from `pathsafe`. It performs **no**
   filesystem I/O and does **not** call `normalize`, `Clean`, `Abs`, or
   `EvalSymlinks` — assertable mechanically by reading the function body.
3. The helper splits on both `/` and `\` on **all** platforms (a test asserts
   `..\..\etc` is rejected on Linux — this is the specific regression review
   caught).
4. **ZERO verdict change.** Every existing `paths_test.go` rule-7 case passes
   **unmodified**, explicitly including
   `TestValidateRule7AbsoluteDatabasePathAccepted`. No input changes verdict in
   either direction. If any existing test requires modification, **stop and
   return the unit blocked** — that is proof the move was not behaviour-neutral.
5. `pathsafe` gains no dependency on `internal/config` (dependency direction is
   one-way and must stay that way).

## Dependency Graph

```
U1                (needs 009-F U5 Wrapf)
U2  U4            (independent)
U3 + U5 ──► U6    (U3/U5 land together; helper move follows)
```
**Cross-shipment: 009-F must land first**, because U1 consumes `apperr.Wrapf`
introduced by 009-F U5. This is a hard symbol dependency, not a stylistic
preference.

## Post-Review Remediation Record (2026-09-04)

Adversarial multi-persona review (Correctness, Maintainability, Scope,
Constitution) returned **FAIL** on the pre-remediation draft. Two blockers and
several majors were fixed:

| Finding | Severity | Remediation |
|---|---|---|
| **U1 rested on a false premise.** `apperr.Wrap` takes no message parameter (`apperr.go:142`) and `New` drops the cause, so AC-3 (retain context) and AC-4 (existing test passes unchanged) were **mutually unsatisfiable**. An unattended Ship would have shipped a context-losing, test-breaking change to a security control. | **blocker** | 009-F U5 adds `Wrapf`; U1 now consumes it and carries an explicit "return blocked" instruction if it is unavailable. |
| **U6 would have caused a containment REGRESSION.** `containsDotDotSegment` splits on `/` **and** `\` on all platforms; a `normalize`-based predicate does not, so `..\..\etc` — rejected today — would newly be **accepted** on Linux. U6 also would have broken `TestValidateRule7AbsoluteDatabasePathAccepted`, since absolute `database.path` is deliberately permitted. The plan wrongly framed the delta as one-directionally "stricter". | **blocker** | U6 rewritten as a **zero-verdict-change move** of the exact existing logic. Semantic unification is explicitly deferred to a future cycle with its own deliberation. AC-3 now pins the Linux backslash case; AC-4 makes any test modification a stop condition. |
| Plan referenced "rule 10" and `validate_test.go`; ground truth is **rule 7** and `paths_test.go`. An executor grepping for "rule 10" finds nothing. | major | All references corrected throughout. |
| U6 was oversized (2 packages, 4 files) and mixed artifact classes. | major | Reduced to a mechanical move; scope no longer spans a semantic redesign. |
| U3's fix could introduce a false accept if implemented unconditionally. | major | AC-4 added requiring the conditional form and naming the `/p/ws` vs `/p/ws-evil` breach. |
| U3 AC-4 required a Windows drive-root case, but PR CI is `ubuntu-latest` only — the AC would skip forever in the only gate that runs. | major | POSIX `/` is now the gated case; `C:\` is a GOOS-conditional, non-blocking extra. |
| U5 was a trivially small standalone unit. | minor | Folded into U3's commit. |
| Constitution table used truncated titles and misclassified VII/VIII — the exact failure mode of stash `2130906D`. | major | Table rewritten with exact ratified titles and corrected classifications. |
| Stated 009-F→010-F dependency was spurious (`pathsafe` already imports `apperr`). | minor | Replaced with the real `Wrapf` symbol dependency. |

**Confirmed correct by review, unchanged:** the doubled-separator bug at the
filesystem root is **real** (`hasPathPrefix` → `pathHasPrefix(path, prefix+sep)`
yields `//`, rejecting all candidates when root is `/`); U2's resolve-then-
type-check ordering is consistent with `NewRoot`'s existing
`Abs → EvalSymlinks → stripUNC` sequence.

## Risks and Caveats

**R1 — SUPERSEDED by post-review remediation.** The original R1 accepted a validation-verdict delta. Review proved that delta contained a containment regression and a broken shipped contract, so U6 was rewritten to a zero-verdict-change move. The residual control is U6 AC-4: any required modification to an existing `paths_test.go` rule-7 test is proof of behaviour change and forces the unit blocked.

**R2 — U3 could weaken containment.** Loosening a prefix comparison is exactly
how containment bugs are introduced. AC-2 requires proving escape detection
still rejects in the same configuration; this unit must not be reviewed on the
positive case alone.

**R3 — U1 wrapping could break sentinel matching.** If `Wrap` is applied such
that `errors.Is(err, apperr.ErrPathViolation)` stops matching, callers relying
on the Kind sentinel break. AC-2 pins both directions.

**R4 — Filesystem-root testing is awkward.** U3's scenario needs a root at `/`
or a drive root, which is not directly creatable in a sandbox. Use the existing
test seams or a synthetic `Root` value; do **not** weaken the test to a lexical
unit check that never exercises `hasPathPrefix`.

**R5 — Symlink privileges.** As in 008-F, symlink tests must skip visibly, never
silently.

## Plan Hardening Signals

| Signal | Present | Note |
|---|---|---|
| Schema/contract change | No (post-remediation) | U6 rewritten to zero verdict change. |
| Security-sensitive behaviour | **Yes** | Entire package is the containment control. |
| Migration | No | — |
| External dependency | No | — |
| Breaking-change potential | No (post-remediation) | Eliminated by the U6 rewrite; AC-4 is the stop condition. |

**Requires plan hardening: yes.**

---

# Plan Hardening

**Hardening required: YES** (operator-facing schema change; security control; breaking-change potential). Hardened 2026-09-04.

## Context consulted

`internal/pathsafe/pathsafe.go` (`splitPath` L78, `isRooted` L96, `filepathSplitList` L109, `Resolve` L140, `hasPathPrefix` L166, `checkSymlinkEscape` L214), `root.go` (`NewRoot` L41), `internal/config/validate.go` (rule 7, `containsDotDotSegment`), the 002-S/003-S closure records, and `docs/plans/2026-09-03-intercom-go-apperr-pathsafe-plan.md` (whose literal `apperr.New` mandate U1 deliberately supersedes).

## Protected invariants

* **`Root.Resolve` containment semantics are unchanged.** The symlink-escape algorithm hardened in 002-S (including the intermediate-directory case) must keep rejecting.
* **Every input rejected by rule 7 today is still rejected.**
* **`errors.Is(err, apperr.ErrPathViolation)` keeps matching** for both `NewRoot` failure branches.
* **Dependency direction:** `config` may depend on `pathsafe`; `pathsafe` must never depend on `config`.
* **`pathsafe`'s exported surface grows by at most one symbol, performing no I/O.**

## Risky actions (ProposedAction / ActionRisk)

### PA-1 — SUPERSEDED by post-review remediation
The original PA-1 accepted a behaviour delta in `database.path` validation.
Adversarial review proved that delta included a **containment regression**
(Linux `..\..\etc` newly accepted) and a **broken shipped contract** (absolute
paths are deliberately permitted). U6 was rewritten as a zero-verdict-change
move, so this risk **no longer exists by construction**. The residual control is
U6 AC-4: if any existing `paths_test.go` rule-7 test requires modification, that
is proof the move was not behaviour-neutral and Ship must **return the unit
blocked**. Semantic unification of the two predicates is deferred to a future
cycle with its own deliberation and operator-facing migration decision.

### PA-1b — Introduce a new exported symbol on a security primitive (U6)
**Risk: LOW (reduced from MEDIUM).** Exporting the verbatim existing logic adds
one lexical, I/O-free function. U6 AC-2 forbids it from calling `normalize`,
`Clean`, `Abs`, or `EvalSymlinks`, which is mechanically checkable by reading
the body.

### PA-2 — Loosen a containment prefix comparison (U3)
**Risk: HIGH.** Relaxing `hasPathPrefix` to fix a false *rejection* risks introducing a false *acceptance* — a containment breach.
**Control:** AC-2 requires proving escape detection still rejects in the same filesystem-root configuration. This unit may not be reviewed on its positive case alone; the negative case is mandatory.

### PA-3 — SUPERSEDED (merged into PA-1b above).

### PA-4 — Change error construction in a security control (U1)
**Risk: MEDIUM.** Wrapping could break sentinel matching, so callers keying on `ErrPathViolation` would silently stop matching.
**Control:** AC-2 pins both discrimination directions (cause **and** sentinel).

### Not risky — explicitly classified
U5's wrapper deletion (zero behaviour change, no exported surface); U4's added regression test.

## Rollback points

U1, U2, U3, U4, U5 are independently revertible. U6 depends on U3 and U5 and is the only unit with cross-package impact; it should land last and as its own commit so it can be reverted without losing the other five fixes.

## Deepened runtime verification

1. `go test ./... -race` green; `internal/config` tests pass **unmodified** for existing rule-7 cases.
2. **Negative controls:** traversal from a filesystem root still rejected (U3); symlinked-intermediate-directory escape still rejected (regression from 002-S).
3. Dangling-symlink accept case observed running, not skipped (U4).
4. Confirm exactly one new exported symbol via package API inspection.
5. Confirm the behaviour delta list from U6 AC-4 is non-surprising and recorded in the PR body.

## Human checkpoints

* **Blocking:** if U6 requires modifying ANY existing `paths_test.go` rule-7 test, halt and return the unit blocked — that is proof the move changed behaviour (U6 AC-4).
* Otherwise none; the traversal-helper design decision is settled in the plan body.

---
title: "Implementation Plan — Pathsafe Error-Surface Parity and Windows Long-Path Precision"
date: 2026-09-10
status: reviewed
agent: Stage
mode: DARK_MODE_ACTIVE
governs: 017-F / shipment 016-S
source: docs/decisions/2026-09-10-intercom-go-pathsafe-error-surface-parity-deliberation.md
---

# Implementation Plan — Pathsafe Error-Surface Parity and Windows Long-Path Precision

**Source document**:
`docs/decisions/2026-09-10-intercom-go-pathsafe-error-surface-parity-deliberation.md`
**Top-level release unit**: feature **017-F**
**Shipment**: **016-S**
**Requires plan hardening**: **yes** — see §9 (NON-NEGOTIABLE security control
surface, Windows-only syscall behaviour, observable error-contract change).

---

## 1. Objective

Bring `internal/pathsafe`'s **error surface** to parity between `NewRoot` and
`checkSymlinkEscape`, and raise the precision of the Windows long-path
threshold — without touching the containment verdict of any path, and without
adding any filesystem write primitive.

## 2. Root cause

| Item | Root cause |
|---|---|
| `7ADAC481` | `016.007-T` (plan C2/AC3) scoped `*fs.PathError` normalization to *"NewRoot's failure branches"* only. `checkSymlinkEscape`'s structurally identical `canonicalizeReparse` failure branch was left on the pre-existing `apperr.Wrapf`-only path, so on Windows it surfaces a raw `syscall.Errno`. |
| `6B751D8B` (1) | A reachability claim about `addLongPathPrefix`'s `\\.\` guard was written from plausibility rather than measurement, and is **false** (`filepath.IsAbs` returns `true` for device paths). |
| `6B751D8B` (2) | `len(path)` counts UTF-8 bytes; Windows' MAX_PATH counts UTF-16 code units. |

## 3. Surface

| File | Change |
|---|---|
| `internal/pathsafe/pathsafe.go` | 1 error-wrap call site + comment (`checkSymlinkEscape`). |
| `internal/pathsafe/reparse_windows.go` | `addLongPathPrefix` threshold comparison + doc-comment correction; new unexported UTF-16 counter. |
| `internal/pathsafe/symlink_test.go` | New cross-platform `*fs.PathError` parity tests. |
| `internal/pathsafe/reparse_windows_test.go` | New reachability-locking + UTF-16 table tests. |

**Not touched**: `root.go` (incl. `NewRoot`, `wrapPathError`, the risk
register), `lexical.go`, `reparse_other.go`, `canonicalizeReparse`,
`scripts/check-write-path-precondition.sh`, CI workflows, `go.mod`.

## 4. Key design decisions carried from deliberation

D-3 (reuse existing `wrapPathError`, `Op: "open"`, do not relocate
normalization into `canonicalizeReparse`), D-4 (keep the `\\.\` branch; lock
reachability; correct prose), D-5 (precision not security), D-6 (`BF5DE670`
preserved as an AC, no task), D-9 (002 → 003 hard sequence).

## 5. Verified preconditions (measured at `1ff2b42`, this session)

1. `filepath.IsAbs(\\.\C:\foo)` = `true`; a 307-char `\\.\` path is also
   `IsAbs` — the branch is **reachable** (isolated `go run` probe).
2. UTF-8 bytes ≥ UTF-16 code units for all inputs (200×`é` → 400/200;
   100×😀 → 400/200) — the existing proxy is conservative, never
   under-triggering.
3. `wrapPathError` is defined in `root.go`, which carries **no build tag** →
   callable from platform-agnostic `pathsafe.go`.
4. `apperr.Wrapf` retains `cause`; `*apperr.Error` implements `Unwrap()` →
   `errors.As` traverses into an injected `*fs.PathError`.
5. `scripts/check-write-path-precondition.sh` exits 0 and `--self-test` passes
   5/5; zero write primitives across 17 non-test `.go` files.
6. `reparse_windows_test.go` already carries `//go:build windows`;
   `symlink_test.go` is cross-platform and already exercises
   `checkSymlinkEscape`'s failure branches.

## 6. Task breakdown (test-first)

### Sub-epic A — Error-surface parity (`7ADAC481`) → `017.001-T`

**Title**: Normalize `checkSymlinkEscape` reparse error via `wrapPathError`
**Size**: S · **Complexity**: medium · **Posture**: test-first

Change (single call site, `pathsafe.go` ~line 359):

```go
real, err := canonicalizeReparse(ancestor)
if err != nil {
    return "", apperr.Wrapf(apperr.KindPathViolation,
        wrapPathError("open", ancestor, err), symlinkUnverifiableMsg)
}
```

**Acceptance criteria**

- **AC1** *(test-first)* A new test in `symlink_test.go` asserts
  `errors.As(err, &pathErr)` **succeeds** for the error returned by
  `Root.Resolve` when an existing ancestor fails canonicalization. The test
  must **fail before** the change and **pass after** on Windows.
- **AC2** The same assertion holds on POSIX (parity, no regression) — the test
  is cross-platform and unguarded by a build tag.
- **AC3 (WINDOWS-ONLY — `//go:build windows`)** The recovered `*fs.PathError`
  has `Op == "open"` (matching `NewRoot`'s corrected label per 015-S
  remediation item 2) and `Path` equal to the **ancestor actually
  canonicalized**, not the originally requested path.
  **These assertions MUST NOT appear in the cross-platform AC1/AC2 test.**
  On POSIX `wrapPathError` is inert — it returns `filepath.EvalSymlinks`'
  own `*fs.PathError` unchanged, whose `Op` is `"lstat"` (or `"EvalSymlinks"`
  for `ELOOP`) and whose `Path` is a failing component/target, **not**
  `ancestor`. Asserting `Op == "open"` cross-platform would either fail on
  POSIX or force a change to the POSIX error surface, contradicting H4.
  *(Review R1 — raised independently by Go Reviewer (P1) and Correctness
  Reviewer (P2).)*
- **AC3b (WINDOWS-ONLY, privilege-independent)** Because
  `requireSymlinkOrFailClosed` legitimately **skips** on Windows when
  `SeCreateSymbolicLinkPrivilege` is not held (verified in
  `symlink_test.go`), a symlink-only test can leave the Windows
  `syscall.Errno → *fs.PathError` path **unexercised while the suite reports
  green**. Add a Windows test in `junction_windows_test.go` driving the same
  branch through a **dangling junction** fixture, which needs no symlink
  privilege. *(Review R3 — Go Reviewer P2.)*
- **AC3c** Record in a code comment why the **sibling** error branch in
  `checkSymlinkEscape` (the non-`ENOENT` `os.Lstat` probe failure) is
  deliberately left unchanged: `os.Lstat` already returns `*fs.PathError` on
  both platforms, so `errors.As` already matches there and no wrap is needed.
  Without this note a future reviewer reads the parity fix as incomplete.
  *(Review R9 — Correctness Reviewer P3.)*
- **AC3d** The `*fs.PathError.Path` now reachable via `errors.As` is an
  **absolute in-workspace ancestor**, intended for operator/diagnostic use
  only. Document that it MUST NOT be interpolated into session-facing
  MCP/agent error payloads. The agent-visible display string is unchanged:
  `apperr.Error.Error()` never appends the cause, so callers still see
  `path violation: symlink target cannot be verified`.
  *(Review R14 — Security Lens Reviewer P3.)*
- **AC4** Unchanged observable contract: `apperr.KindPathViolation` is still
  returned, the message still contains `symlinkUnverifiableMsg`, and every
  pre-existing `errors.Is(err, fs.ErrNotExist)` relationship is preserved —
  asserted explicitly, not assumed.
- **AC5** No containment verdict changes: every existing test in
  `symlink_test.go`, `contain_test.go`, `root_test.go`, `junction_windows_test.go`
  passes unmodified.
- **AC6** `go test -race -mod=readonly ./...` green; `golangci-lint run` clean
  including the `GOOS=windows` step.

### Sub-epic B — Windows long-path precision (`6B751D8B` items 1–2)

#### `017.002-T` — Lock `\\.\` device-namespace branch reachability

**Size**: S · **Complexity**: low · **Posture**: characterization-first
**Principle II deviation (documented, per Constitution Governance
"Conflict resolution")**: this task makes **no production behaviour change**,
so there is no red phase to drive. The constitution's fail-before/pass-after
letter cannot apply. The justified substitute is a Feathers-style
characterization/regression-locking test around already-verified-reachable
code; the rejected simpler alternative is "correct the comment only", rejected
because it would leave the corrected claim mechanically unenforced.
*(Review R7 — Constitution Reviewer P2.)*

**Acceptance criteria**

- **AC1** A `//go:build windows` assertion pins that `filepath.IsAbs` returns
  `true` for a `\\.\`-prefixed path at and beyond `longPathThreshold`,
  mechanically falsifying the stale "unreachable" premise. **Explicitly NOT
  labelled test-first**: this passes at pre-change HEAD by construction (it
  characterizes stdlib behaviour the corrected comment now depends on). Its
  value is that it **fails loudly if a future Go release changes
  `filepath.IsAbs`'s device-path handling**, which is exactly the premise the
  comment asserts. *(Review R12 — Scope Auditor P3.)*
- **AC2** A table test asserts `addLongPathPrefix` returns a `\\.\` device path
  **unchanged** (never `\\?\`-prefixed). **Delta over existing coverage**:
  `TestAddLongPathPrefixVerdicts` already covers the *above*-threshold device
  case — do **not** re-add it. This task adds only the **below-threshold and
  exactly-at-threshold** device cases plus `\\.\PhysicalDrive0` and
  `\\.\UNC\...` shapes. *(Review R10 — Correctness Reviewer P3, Scope
  Auditor P3.)*
- **AC3** *(falsifiable form)* The device-path assertions are written against
  `addLongPathPrefix`'s **observable output only**, with no reference to guard
  ordering, so the test stays green if the adjacent generic `\\` guard is
  reordered and fails if any `\\.\` input ever gains a `\\?\` prefix.
  Verified by temporarily reordering the guards locally and confirming the
  test still passes. *(Review R13.)*
- **AC4** The `addLongPathPrefix` doc comment is corrected to state the branch
  is **reachable but intentionally redundant** defence-in-depth relative to the
  generic `\\` guard. The prior/implied unreachability claim is removed and the
  verification method is recorded inline so the next reviewer need not
  re-litigate it.
- **AC5** *(falsifiable form)* `git diff` for this task shows **no change to
  any executable statement** in `reparse_windows.go` — comment and test lines
  only. The full pre-existing suite passes unmodified. *(Review R13.)*
- **AC6** `go test -race -mod=readonly ./...` green; `golangci-lint run`
  (incl. `GOOS=windows`) clean.

#### `017.003-T` — Make the MAX_PATH threshold UTF-16-code-unit-aware

**Size**: S · **Complexity**: medium · **Posture**: test-first
**Depends on**: `017.002-T` (**coordination**, not regression-net — see §7)

**Implementation note (R11 — Scope Auditor P3)**: use **`len(utf16.Encode(
[]rune(path)))`** or a plain `for _, r := range path` accumulator. The
deliberation's O2-c "non-allocating counter" is **withdrawn as gold-plating**:
`addLongPathPrefix`'s only caller immediately calls
`syscall.UTF16PtrFromString` (which allocates) and then two Win32 syscalls, so
saving two small allocations here buys nothing and adds bespoke infrastructure.

**Acceptance criteria**

- **AC1** *(test-first)* A `//go:build windows` table test covers ASCII,
  2-byte (`é`), 3-byte (CJK), and surrogate-pair (😀) inputs, asserting the
  prefix decision is made on **UTF-16 code units**. `len(utf16.Encode([]rune(p)))`
  is used as the independent test oracle.
- **AC2** A non-ASCII path whose UTF-8 length ≥ 260 but UTF-16 length < 260 is
  **not** extended-prefixed — the precision gain. Fails before, passes after.
- **AC3** A path whose UTF-16 length ≥ 260 **is** extended-prefixed regardless
  of its byte length — the safety property is retained.
- **AC4** A surrogate pair counts as **2** code units (explicitly asserted), so
  the implementation cannot regress to a rune count.
- **AC5** *(scoped — see D-10)* The boundary predicate remains **strict
  less-than against the same `syscall.MAX_PATH` constant**
  (`utf16len < longPathThreshold`), so for ASCII input the verdict is
  **byte-identical** to today and every existing threshold test stays green
  unmodified. This task does **not** claim verdict-neutrality for the
  precision-gain band; see **D-10**.
- **AC6** No new module dependency (`unicode/utf16` is stdlib); `go.mod`
  unchanged.
- **AC7** The doc comment states explicitly that the prior byte-length proxy
  was **conservative** (bytes ≥ code units, so it never under-triggered) and
  that this change is **precision, not a security fix**. No vulnerability
  language. It must also record **D-10** below.
- **AC8** `go test -race -mod=readonly ./...` green; `golangci-lint run`
  (incl. `GOOS=windows`) clean.

#### D-10 — Win32 normalization in the precision-gain band (ACCEPTED)

*(Review R4 — Security Lens Reviewer P2, corroborated by Correctness
Reviewer P3. This was the single most valuable review finding.)*

The original plan asserted 017.003-T was **"verdict-neutral by construction"**.
That **overclaimed**. `\\?\` does not merely bypass MAX_PATH — it also
**disables Win32 path normalization**. Dropping the prefix for the
precision-gain band (UTF-8 ≥ 260 bytes but UTF-16 < 260 units) **re-enables**,
for those paths only:

- trailing dot / trailing space stripping,
- reserved device-name interpretation (`CON`, `NUL`, `AUX`, …).

(8.3 short-name aliasing is NTFS-level and is **not** governed by the prefix.)

**Decision: accept (option (a)).** Those paths thereby behave **exactly like
every sub-MAX_PATH path in the codebase already does today** — this restores
normal Win32 semantics to a band that was anomalously bypassing them purely
because of a byte-vs-code-unit measurement artifact. It does not create a new
normalization regime.

**Consequent obligations**:

- H1 is corrected below to drop the "by construction" claim.
- AC5 is scoped to the ASCII/boundary-identity property only.
- H2's rune-count guard covers the MAX_PATH false-negative risk **only**, and
  must **not** be cited as covering normalization.
- Ship must note in the PR body that `os.Lstat` (via `os.fixLongPath`,
  `LongPathsEnabled`-dependent) and `canonicalizeReparse`'s raw
  `syscall.CreateFile` can in principle observe different names for a
  trailing-dot/space path in this band. No production guard is added; this is a
  recorded, accepted residual, consistent with the package's existing
  oracle-parity risk acceptances.

### Feature-level acceptance criteria (017-F) — `BF5DE670` preservation (D-6)

- **AC-F1** `scripts/check-write-path-precondition.sh` exits **0** and
  `--self-test` passes **5/5** at final HEAD — the deferred-mitigation gate is
  intact and still functional.
- **AC-F2** The consolidated risk register in `internal/pathsafe/root.go`
  (GO-14, Resolve→use TOCTOU window, SEC-5, U5/AC5 UNC long-path residual) is
  **unmodified and still accurate** after this feature's changes.
- **AC-F3** This feature adds **no** filesystem write primitive — verified by
  the gate above plus an independent grep.
- **AC-F4** Stash entries `BF5DE670` and `6B751D8B` remain **active
  (not archived)**; only `7ADAC481` is archived.

## 7. Dependency order

```
017.001-T  (independent — pathsafe.go)
017.002-T  ──coordination──▶  017.003-T   (both mutate addLongPathPrefix)
```

`017.001-T` touches a different file and may run in parallel or first. It is
the **highest-priority** task (P1-in-isolation, D-2) and should lead.

**Dependency rationale corrected (R8 — Correctness Reviewer P3)**: the original
plan called 002 → 003 a *hard* dependency on the grounds that 002's tests are
003's "regression net". That was **overstated**. 003 changes only the **first
(length)** guard; a `\\.\` device path of any length clears the length guard and
is caught by the unchanged device/`\\` guards regardless of byte-vs-code-unit
counting, so 003 **cannot** regress the device-path property 002 locks. 003 is
independently correct. The real constraint is **same-function, same-doc-comment
merge coordination** — sequencing avoids a textual conflict and keeps the
corrected comment coherent. Ordering is retained for that reason only.

## 8. Evidence plan

| Evidence | Method |
|---|---|
| AC1/AC2 fail-before-pass-after | Run each new test at pre-change HEAD, capture failure; re-run after. |
| Windows-only paths | Must be exercised on a Windows runner; state explicitly in the PR body if any case could not be exercised locally (compound learning: local sandbox ≠ CI ground truth). |
| No verdict change | Full `go test -race -mod=readonly ./...` before and after. |
| Gate intact | `check-write-path-precondition.sh` + `--self-test` at final HEAD. |
| Lint | `golangci-lint run` and the `GOOS=windows` step. |

## 9. Plan Hardening

### H0 — Declared safety modes (Constitution Principle VIII)

*(Review R2 — Constitution Reviewer **P1**.)* This work touches a
NON-NEGOTIABLE Constitution III control during an unattended
`DARK_MODE_ACTIVE` cycle. Two safety modes are **explicitly declared by name**:

- **`freeze-scope`** — edits are constrained to `internal/pathsafe/` and its
  tests. `root.go`, `canonicalizeReparse`, `lexical.go`, `reparse_other.go`,
  CI workflows, `go.mod`, and `scripts/` are **out of bounds**. H6 is the
  tripwire.
- **`investigate-first`** — every premise was measured before planning
  (§5): the `IsAbs` probe falsified the stash's stated premise, the UTF-8/
  UTF-16 relation was measured, and the write-path gate was re-run. No fix was
  proposed on an unverified claim.

`careful` is **not** declared: no destructive command is proposed, and
`strict_safety.enabled` is `false` in `.autoharness/config.yaml`.

**Hardening signals present**: (a) `internal/pathsafe` is a NON-NEGOTIABLE
security control (Constitution III); (b) Windows-only syscall/namespace
behaviour that cannot be fully exercised on every runner; (c) an **observable
error-contract change** visible to callers via `errors.As`.

### H1 — Blast radius

Both production edits are inside one package. `017.001-T` changes only the
*error value's shape*, never whether a path is accepted. `017.002-T` changes
**no** production logic. `017.003-T` changes only *when the `\\?\` prefix is
added*, and only in the direction of adding it **less often**.

**Corrected (R4)**: the earlier claim that 017.003-T is *"verdict-neutral by
construction"* is **withdrawn**. It is verdict-neutral with respect to
**MAX_PATH** (UTF-8 bytes ≥ UTF-16 units, so a genuinely over-length path is
still prefixed — AC3/AC4). It is **not** unconditionally verdict-neutral with
respect to **Win32 path normalization**: see **D-10**, where re-enabled
trailing-dot/space stripping and reserved device-name interpretation in the
precision-gain band are analyzed and **explicitly accepted**. Verdict-neutrality
of 017.003-T is additionally contingent on the **already-`Abs`/`Clean`
precondition** both callers satisfy (Correctness Reviewer P3); that precondition
is real and load-bearing today, and is stated here rather than left implicit.

### H2 — Highest-risk failure mode and its guard

**Risk**: `017.003-T` mis-implemented as a **rune** count would *under*-count
surrogate pairs (1 rune = 2 UTF-16 units) and could fail to prefix a genuinely
too-long path — converting today's safe false-positive into a **false
negative**. This is the only change in the feature capable of regressing safety.
**Guard**: AC4 asserts the surrogate-pair-counts-as-2 property directly, and
AC3 asserts the ≥260-units-still-prefixed property. Both fail loudly on a rune
count. O2-a was rejected by name in deliberation §5.3 to prevent re-derivation.

### H3 — Windows verification gap (CI GATING CORRECTED)

`\\.\` device-namespace and >260-char cases may not be exercisable on a
non-Windows or restricted runner.

**Correction (R6 — Constitution Reviewer P2, independently verified by Stage)**:
the earlier claim that "CI runs a `GOOS=windows` lint step" **understated the
gap**. Measured in `.github/workflows/ci.yml`:

- The blocking `test` job runs `go test -race -mod=readonly ./...` on **Linux**
  (L195).
- The `windows-latest` job's `go test -mod=readonly ./...` step (L252) carries
  `continue-on-error: ${{ vars.WINDOWS_GATE_REQUIRED != 'true' }}` (L251) — it
  is **ADVISORY-ONLY** unless that repo variable is set to `'true'`.
- Only the `golangci-lint run (GOOS=windows)` step in the `lint` job is
  unconditionally blocking, and it is **compile/lint-only, not a runtime test
  execution**.

This matters because H2's highest-risk failure mode is a **Windows-only runtime
regression**. **Required mitigations**:

1. Ship MUST determine and record whether `WINDOWS_GATE_REQUIRED == 'true'`
   for this shipment.
2. If it is not, Ship MUST paste the **actual green Windows `go test` output**
   into the PR body as manual evidence. An advisory job's status is **not**
   acceptable evidence for this feature.
3. Windows-only assertions MUST NOT depend on symlink privilege — use the
   junction fixture (AC3b), since `requireSymlinkOrFailClosed` legitimately
   **skips** on `ERROR_PRIVILEGE_NOT_HELD` and would hide a red test behind a
   green suite.
4. Long-path tests must not require the registry `LongPathsEnabled` opt-in —
   `addLongPathPrefix`'s pure-string tests are already structured this way
   (existing `TestAddLongPathPrefixVerdicts` precedent).

### H4 — Error-contract regression

Widening `checkSymlinkEscape`'s error could break a caller matching on
`apperr.Kind` or on `errors.Is(..., fs.ErrNotExist)`. **Guard**: AC4 asserts
both explicitly rather than assuming them. Note `wrapPathError` is a **no-op
when the error already contains an `*fs.PathError`** (it returns the original
unchanged), so on POSIX — where `canonicalizeReparse` is `filepath.EvalSymlinks`
— this change is provably inert.

### H5 — Rollback

Each task is an independent, single-file revert. `017.001-T` is a one-line
revert; `017.002-T` is comment+test only (revert is risk-free); `017.003-T`
reverts to `len(path)`. No migration, no persisted state, no schema, no config.
Rollback of any one task does not strand the others, except that reverting
`017.002-T` alone would remove `017.003-T`'s regression net — so revert 003
before 002 if both must go.

### H6 — Scope-expansion tripwire

`scripts/check-write-path-precondition.sh` must stay green (AC-F1). If any task
tempts the implementer toward a write call site, that is a **stop condition**
(scope expansion), not a fix.

### H7 — Known tooling hazards (from compound library)

- `backlogit stash edit --text` **replaces** the whole body; read back with
  `stash get`, concatenate, and guard with an idempotency marker.
- Task WIT exposes `size` only, **not** `complexity` — carry complexity as
  enum-validated labeled prose and flag the degradation.
- `backlogit dep add --type blocked-by` is invalid; express edges as `blocks`.
- Copilot review must be requested via the `[bot]`-suffixed REST login and
  **re-requested after every HEAD advance** (P-018).

## 9a. Constitution Check (Governance MUST-clause)

*(Review R5 — Constitution Reviewer P2. Required by the constitution's
Governance section: "Every implementation plan MUST include a 'Constitution
Check' section that maps the proposed work against these principles and
documents any justified violations.")*

| Principle | Verdict | Basis |
|---|---|---|
| **I — Safety-First Go** (NON-NEGOTIABLE) | **COMPLIANT** | Explicit errors throughout; the change *increases* `errors.As`-friendliness. No `unsafe`, no `panic`-driven control flow — option O1-b (an unreachability `panic`) was **rejected by name**. `golangci-lint` (incl. `GOOS=windows`) required green in every task AC. |
| **II — Test-First Development** (NON-NEGOTIABLE) | **COMPLIANT with one documented deviation** | `017.001-T` and `017.003-T` are strictly fail-before/pass-after. `017.002-T` has **no production change**, so no red phase exists; the deviation, its justification, and the rejected simpler alternative are documented inline at the task (per the Governance "Conflict resolution" clause). H3 additionally closes the advisory-CI evidence gap. |
| **III — Workspace Isolation / Security Boundaries** | **COMPLIANT — no weakening** | This *is* the control being touched. Both production edits are error-shape-only and threshold-precision-only; neither touches the accept/reject decision path. AC5, AC-F1–AC-F3 and anti-goals 1–5 pin this. The one genuine behavioural side effect (Win32 normalization in the precision band) is analyzed and accepted in **D-10** rather than hidden. |
| **IV — CLI Workspace Containment** (NON-NEGOTIABLE) | **COMPLIANT** | All edits confined to `internal/pathsafe/` inside the repo tree. Stage created artifacts only under `docs/`. |
| **V — Destructive Command Approval** (NON-NEGOTIABLE) | **N/A** | No destructive command proposed. Anti-goal 1 forbids adding any filesystem write primitive; H6/AC-F1 is the mechanical tripwire. |
| **VI — Safety Modes for Risky Work** | **COMPLIANT** | `freeze-scope` and `investigate-first` declared **by name** in **H0**. |
| **Observability / Persistence** | **COMPLIANT** | Deliberation, plan, and plan-review artifacts are Markdown + YAML frontmatter, revision-logged, git-committed, and cross-referenced. The §10 citation resolves to a real committed file (R-resolved). |
| **Dependency discipline** | **COMPLIANT** | `unicode/utf16` is stdlib; `go.mod` unchanged (AC6). |



See `docs/decisions/2026-09-10-intercom-go-pathsafe-error-surface-parity-plan-review.md`.

## 11. Revision log

| Rev | Change |
|---|---|
| 1 | Initial plan from accepted deliberation. |
| 2 | Added §9 Plan Hardening (H1–H7) per P-006. |
| 3 | Multi-persona adversarial review remediation (2 P1, 6 P2, 7 P3) — see plan-review artifact. R1 AC3 scoped Windows-only; R2 safety modes declared (H0); R3 junction-based Windows test (AC3b); R4 **D-10** Win32-normalization overclaim corrected; R5 §9a Constitution Check added; R6 H3 advisory-CI gap corrected; R7 Principle II deviation documented; R8 dependency restated as coordination; R9 sibling-branch note (AC3c); R10/R12/R13 `017.002-T` ACs made falsifiable and de-duplicated; R11 gold-plated UTF-16 counter withdrawn; R14 `PathError.Path` non-forwarding note (AC3d). |

---
title: "Implementation Plan — Pathsafe Reparse-Point Containment (rev 2)"
date: 2026-09-08
status: reviewed
feature: 014-F
---

# Implementation Plan — Pathsafe Reparse-Point Containment

- **Date**: 2026-09-08 · **Revision**: **2**
- **Source deliberation**: `docs/decisions/2026-09-08-intercom-go-pathsafe-reparse-containment-deliberation.md`
- **Feature**: `014-F` · **Shipment**: `013-S`
- **Stash entries covered**: `700B41CE` (critical), `F133AB7E`, `2362BBB5`, `AD0D9D1F`
- **Requires plan hardening**: **yes** — modifies `checkSymlinkEscape`, the single
  enforcement point of a control the Brief marks **NON-NEGOTIABLE** (workspace isolation),
  and deliberately **reverses a shipped, regression-locked risk acceptance** (GO-14).
- **Constitution**: Principle II (Test-First, NON-NEGOTIABLE — RED before GREEN),
  Principle VI (Single Responsibility), Principle VIII (safety mode: **`careful` + `freeze-scope`**),
  Principle XI (merge commit required)
- **Plan-review gate**: **PASS** (cycle 2 of max 3)

<!-- plan-review-attempt: 2 -->

> **Revision 2** resolves 1 P0 and 4 P1 findings raised by an independent three-model
> adversarial panel (Security / gpt-5.6-sol, Correctness / gemini-3.8-flash,
> Scope Boundary / claude-opus-4.8) against rev 1. The most important were:
> **(A)** rev 1's final-component-only fix left a **residual bypass** via an intermediate
> junction with an *existing* leaf; **(B)** relaxing the ancestor branch without making the
> post-loop containment check reparse-aware would have **introduced a new bypass**;
> **(C)** rev 1's F3 usability relaxation was unauthorized scope that *weakened* the control.
> See §10 for full dispositions.

---

## 1. Objective

Close the `700B41CE` containment bypass — a Windows directory junction whose real target is
outside the workspace root being accepted by `Root.Resolve()` — and land the coupled GO-14
replacement boundary in the same change, without weakening the dominant new-file-write path.

**Definition of done:** no input to `Root.Resolve()` returns a path whose real filesystem
target lies outside the workspace root **by way of a reparse point** (symlink, directory
junction, or mount point), at **any** path position, on Windows or POSIX.

**Explicitly NOT claimed as closed by this shipment** (see §8): the case-folding fail-open,
hardlinks (SEC-5), and the Resolve→use TOCTOU window.

---

## 2. Root cause — corrected in rev 2

`checkSymlinkEscape` walks up from `resolved` to the **nearest existing ancestor**, then
calls `filepath.EvalSymlinks(ancestor)` and compares the result to `root.path`.

`filepath.EvalSymlinks` **does not resolve `IO_REPARSE_TAG_MOUNT_POINT`** on Windows
(`os/types_windows.go` maps mount points to `ModeIrregular`; only `IO_REPARSE_TAG_SYMLINK`
yields `ModeSymlink`). Therefore **whenever the ancestor walk terminates at or below a
junction, the containment check compares a lexically in-root path and passes.**

This produces **three** distinct accepting inputs, not one:

| # | Input shape | Why the walk terminates without resolving |
|---|---|---|
| C1 | `Resolve("junction")` — junction **is** the final component | `os.Stat` follows the mount point and succeeds on iteration 1 |
| C2 | `Resolve("junction/existing.txt")` — junction is an **intermediate**, leaf **exists** | `os.Stat`/`os.Lstat` of the full path traverses the junction and succeeds on iteration 1 |
| C3 | `Resolve("junction/existing-dir/missing.txt")` | ascent stops at `existing-dir`, which lies **beyond** the junction |

**Rev 1 fixed only C1.** An `Lstat`-first probe of the *final component* cannot see C2 or C3,
because `Lstat` does not follow only the **final** component — it still traverses every
intermediate one, so it reports an ordinary file/directory and the reparse branch never fires.

**Corrected root-cause statement:** the defect is not "the final component's reparse identity
is unexamined"; it is **"the containment check is performed against a path that was never
mount-point-canonicalized."** GO-14 (dangling final symlink) is the same defect reached by the
`ENOENT` ascent path.

**Consequence for the fix:** the remedy is **mount-point-aware canonicalization of the whole
existing prefix**, replacing the naked `filepath.EvalSymlinks(ancestor)` — not a special case
bolted onto the final component.

---

## 3. Surface

| File | Change kind |
|---|---|
| `internal/pathsafe/reparse_windows.go` | **new** — Windows reparse-aware canonicalization |
| `internal/pathsafe/reparse_other.go` | **new** — non-Windows build-tagged fallback |
| `internal/pathsafe/pathsafe.go` | `checkSymlinkEscape` containment check + new message constant |
| `internal/pathsafe/root.go` | risk register, `uncPrefix` doc comment |
| `internal/pathsafe/junction_windows_test.go` | live-junction tests (reuse `createDirectoryJunction`) |
| `internal/pathsafe/symlink_test.go`, `contain_test.go`, `root_test.go` | lock inversion, message audit, F4/F11 |

**Out of surface (anti-goals):** `scripts/check-retired-architecture.sh`,
`.github/workflows/ci.yml`, `go.mod`, TOCTOU/hardlink mitigation machinery.

---

## 4. Tasks

### `014.001-T` — RED: reparse-point containment test matrix (Windows)
**Domain:** tests · **Size:** M · **Complexity:** medium · **Depends on:** —

Add to `junction_windows_test.go`, reusing `createDirectoryJunction` and its `pwsh`
`exec.LookPath` skip guard (do **not** re-implement either; F10 is already closed).

**Acceptance criteria**
1. Case **C1** — junction → outside root as the **literal final component**:
   `Resolve("junction")` **must reject**. Must use `Resolve("junction")`, **not**
   `Resolve("junction/child")`; a subpath exercises the already-fixed ancestor branch and
   yields a **false-negative "safe" result** (recorded in
   `docs/compound/2026-09-08-adversarial-review-empirical-verification-and-scope-check.md`).
2. Case **C2** — junction → outside root as **intermediate**, with an **existing** leaf
   beyond it: `Resolve("junction/existing.txt")` **must reject**. *(New in rev 2 — this
   is the residual bypass rev 1 missed.)*
3. Case **C3** — junction → outside root, ascent terminating on an existing directory
   **beyond** the junction: `Resolve("junction/existing-dir/missing.txt")` **must reject**.
4. All three **fail against current `HEAD`** (proving RED).
5. No test requires elevated privilege (junction creation needs none).

**Anti-goal:** do **not** add a case asserting that an **in-root** junction ancestor is
*accepted*. That relaxation (F3) is **deferred** — see §8.

---

### `014.002-T` — RED: invert the GO-14 regression lock
**Domain:** tests · **Size:** S · **Complexity:** medium · **Depends on:** —

**Acceptance criteria**
1. `TestResolveAllowsDanglingSymlinkAtFinalComponent` (`symlink_test.go:149`) is **renamed**
   to `TestResolveRejectsDanglingSymlinkAtFinalComponent` and inverted: `Resolve("dangling-link")`
   must return a `KindPathViolation` error. **Inverted and renamed, never deleted** — it is the
   durable evidence that the boundary moved deliberately.
2. `TestResolveAllowsNonExistentRelativePath` (`symlink_test.go:352`) semantics are pinned as an
   explicit non-regression: a **plain non-existent leaf with no `Lstat` entry** must still
   **ACCEPT**. This is the dominant new-file-write use case and the single most important
   non-regression in this shipment.
3. `TestResolveAllowsSymlinkedDirectoryWithNonExistentLeaf` (`symlink_test.go:318`) and
   `TestResolveAllowsSymlinkInsideRoot` (`symlink_test.go:88`) are confirmed to still **ACCEPT**
   (both resolve to in-root targets).
4. Assert only `apperr.KindPathViolation` here; message-constant assertions land in `014.006-T`.
5. Tests in AC 1 fail against current `HEAD` (RED).

---

### `014.003-T` — Reparse-aware canonicalization helper (platform-split)
**Domain:** code · **Size:** M · **Complexity:** **high** · **Depends on:** —

Introduce `canonicalizeReparse(path string) (string, error)` behind a build-tagged pair.
This is a **self-contained helper with no `checkSymlinkEscape` changes** — split out in rev 2
because Win32 syscalls cannot live in the cross-platform `pathsafe.go`.

**Acceptance criteria**
1. `reparse_windows.go` resolves an existing path to its **true** target using
   `GetFinalPathNameByHandleW` semantics via `syscall` (`CreateFileW` with
   `FILE_FLAG_BACKUP_SEMANTICS`, handle closed on every path including error paths).
   This resolves `IO_REPARSE_TAG_MOUNT_POINT`, which `filepath.EvalSymlinks` does not.
2. `reparse_other.go` delegates to `filepath.EvalSymlinks` (POSIX has no mount-point
   reparse construct; symlinks are already resolved correctly).
3. `go.mod` is **not** modified — no `golang.org/x/sys/windows` dependency (D-3).
4. The returned path is normalized through the existing `stripUNCPrefix` (the `\\?\`
   extended form is expected from the handle-based call).
5. Package compiles on Windows, Linux, and Darwin.
6. Unit tests cover: a plain directory, an in-root junction, an out-of-root junction, and
   a non-existent path (error).

---

### `014.004-T` — GREEN: wire canonicalization into the containment check
**Domain:** code · **Size:** S · **Complexity:** **high**
**Depends on:** `014.001-T`, `014.002-T`, `014.003-T`

Replace `filepath.EvalSymlinks(ancestor)` in `checkSymlinkEscape` with
`canonicalizeReparse(ancestor)`, and probe the final component with `os.Lstat` first so a
final-component reparse entry (live **or** dangling) is canonicalized rather than trusted.

**Acceptance criteria**
1. `014.001-T` C1/C2/C3 and `014.002-T` AC 1 all turn **GREEN**.
2. `014.002-T` AC 2/AC 3 (the accept paths) remain **GREEN** — no over-rejection.
3. A component whose target **cannot be resolved** is **rejected** (fail closed), never accepted.
4. **The strict-ancestor rejection of `ModeIrregular` entries is retained unchanged.**
   Rev 2 does **not** relax it (F3 deferred, §8). This guarantees the change is
   *monotonically fail-closed*: it converts accepts into rejects, never the reverse.
5. **GODEBUG independence:** the fix must not rely on `os.Stat`/`filepath.EvalSymlinks`
   reparse semantics, so the verdict is identical under `GODEBUG=winsymlink=0` and `=1`
   **by construction**. State this explicitly in the commit message.
6. If handle-based resolution proves infeasible within budget, fall back to deliberation
   option **O3** — reject **all** final-component reparse points unconditionally — and defer
   the usability case. **Never** fall back to leaving the bypass open. Under O3, `014.001-T`
   cases C1–C3 still pass (all are rejections), so no RED test is left stranded.
7. If `os.Readlink`-based resolution is reintroduced, the commit must state why it is no
   longer redundant (a redundant one was removed on review in `9158a05`).

---

### `014.005-T` — Flip the fail-open terminal branch
**Domain:** code · **Size:** XS · **Complexity:** low · **Depends on:** `014.004-T`

Covers `2362BBB5`(a) **and** `AD0D9D1F`(F6) — the **same finding** captured twice
(deliberation D-5). Implement **once**.

**Acceptance criteria**
1. `checkSymlinkEscape`'s `parent == ancestor` terminal branch returns
   `apperr.New(apperr.KindPathViolation, ...)` instead of `(resolved, nil)`.
2. The `011.006-T` coverage-exclusion annotation is updated in the same task.
3. Zero behavior change across the existing suite (the branch is traced-unreachable).
4. Commit/doc text must **not** claim a reproduced exploit — the branch's two hypothesized
   triggers were never reproduced.

---

### `014.006-T` — Distinct unverifiable-target error message
**Domain:** code · **Size:** S · **Complexity:** medium · **Depends on:** `014.004-T`

Covers `2362BBB5`(b). Rejections for entries that exist but are not resolvable (in-root
dangling link, ELOOP, EACCES, unresolvable reparse point) currently claim
"symlink target escapes workspace" — a correct verdict with a **false explanation** that
also discards the OS cause.

**Acceptance criteria**
1. New message constant (e.g. `symlinkUnverifiableMsg` = "symlink target cannot be verified")
   used for unverifiable-target rejections; genuine escapes keep `symlinkEscapeMsg`.
2. `apperr.KindPathViolation` retained for both (no taxonomy change).
3. OS cause wrapped and recoverable via `errors.Is` / `errors.As`.
4. **Audit all 9 existing assertion sites** that reference `symlinkEscapeMsg`
   (`symlink_test.go:78,137,192,219,252,284,319`; `contain_test.go:132`) and update every
   one whose rejection reason changes to *unverifiable*. Rev 1 missed this. Known
   affected: `TestResolveRejectsDanglingIntermediateSymlinkInsideRoot` (`symlink_test.go:210`)
   and `TestResolveRejectsChildBeneathRegularFile` (`contain_test.go:102`, POSIX ENOTDIR path).
5. Full `internal/pathsafe` suite GREEN after the audit.

---

### `014.007-T` — Close branch-coverage gaps F4 and F11
**Domain:** tests · **Size:** S · **Complexity:** low
**Depends on:** `014.006-T` *(rev 2: was `014.004-T`; F11 asserts the message/cause that `014.006-T` defines)*

**Acceptance criteria**
1. **F4** — a POSIX-executable test covers the ENOTDIR-vs-ENOENT ancestor-guard branch that
   `TestResolveRejectsChildBeneathRegularFile` is named for but does not exercise on POSIX.
2. **F11** — a dedicated test covers the `!errors.Is(err, fs.ErrNotExist)` branch (EACCES or
   ELOOP), asserting rejection, the correct message constant, **and** cause preservation.
3. Both run in the default suite on at least one CI platform.

---

### `014.008-T` — Reconcile the risk register and doc comments
**Domain:** docs · **Size:** S · **Complexity:** low
**Depends on:** `014.004-T`, `014.005-T`, `014.006-T`

**Acceptance criteria**
1. `700B41CE` register entry → **CLOSED for the reparse-point mechanism**, citing the shipping
   commit and naming all three closed shapes (C1/C2/C3). The closure must **not** overclaim:
   it explicitly does not cover case-folding, hardlinks, or TOCTOU.
2. `GO-14` entry rewritten to record the **flip**: acceptance withdrawn, replacement boundary
   named, oracle divergence justified by **D8** (the Rust repo is historical reference only,
   not a behavioral oracle).
3. **F8** — the `uncPrefix` doc comment no longer misattributes `\\?\` to
   `GetFinalPathNameByHandle`. *(Note: `014.003-T` legitimately introduces a real
   `GetFinalPathNameByHandleW` call — keep the two attributions distinct and accurate.)*
4. **F9** — register entry `5FE4A7BE`'s two contradictory STATUS/TRIGGER pairs split into two
   clearly separated entries.
5. The stale `checkSymlinkEscape` "KNOWN LIMITATION" block describing the junction bypass is
   replaced with an accurate description of the new behavior.
6. **Three new residual register entries added** (§8): case-folding fail-open, F3 in-root
   junction ancestor false-rejection, GODEBUG both-settings empirical verification.
7. `scripts/check-write-path-precondition.sh --self-test` stays green; `BF5DE670`/TOCTOU/SEC-5
   entries left intact (trigger unfired).

---

## 5. Dependency graph

```
014.001-T ─┐
014.002-T ─┼─→ 014.004-T ─┬─→ 014.005-T ─┐
014.003-T ─┘              │              ├─→ 014.008-T
                          └─→ 014.006-T ─┴─→ 014.007-T
```

RED (`001`, `002`) + helper (`003`) → GREEN (`004`) → hardening (`005`–`007`) → docs (`008`).

**Effort:** 8 tasks × 2h = **16h**.

---

## 6. Plan hardening (P-006)

| # | Risk | Sev | Mitigation |
|---|---|---|---|
| H1 | Fix closes C1 but leaves C2/C3 open — bypass survives a "successful" shipment | **critical** | Root cause restated (§2); `014.001-T` AC 2/3 make C2 and C3 mandatory RED cases |
| H2 | Relaxing the ancestor branch introduces a **new** intermediate-junction bypass | **critical** | `014.004-T` AC 4 retains the `ModeIrregular` rejection; F3 deferred (§8). Change is monotonically fail-closed |
| H3 | Fix over-rejects and breaks new-file creation | **high** | `014.002-T` AC 2/3 pin three accept paths as explicit non-regressions |
| H4 | Win32 syscalls break Linux/Darwin compilation | **high** | `014.003-T` platform-split into build-tagged files; AC 5 requires 3-OS compile |
| H5 | Verdict silently depends on `GODEBUG=winsymlink` | **high** | Handle-based resolution is GODEBUG-independent **by construction** (`014.004-T` AC 5); empirical both-settings proof deferred (§8) |
| H6 | GO-14 lock deleted rather than inverted, erasing evidence | **high** | `014.002-T` AC 1; deletion is an explicit anti-goal |
| H7 | **Windows-only tests cannot block CI** (`WINDOWS_GATE_REQUIRED` unset) | **high** | §7 — mandatory local Windows evidence + PR disclosure; flip is an operator decision |
| H8 | Message-constant change breaks existing assertions | medium | `014.006-T` AC 4 audits all 9 `symlinkEscapeMsg` sites |
| H9 | Overclaiming `700B41CE` as fully closed | medium | `014.008-T` AC 1 scopes closure to the reparse mechanism; §8 records residuals |
| H10 | Register desynchronizes from the coupled write-path gate | medium | `014.008-T` AC 7 |
| H11 | Handle leak in the Win32 path | medium | `014.003-T` AC 1 requires close on all paths incl. errors |
| H12 | Reintroduced `os.Readlink` re-triggers a prior review finding | low | `014.004-T` AC 7 |
| H13 | Fail-open flip overclaimed as an exploit fix | low | `014.005-T` AC 4 |

---

## 7. Known gate limitation — CI cannot verify the Windows fix (operator decision)

`.github/workflows/ci.yml:240` sets
`continue-on-error: ${{ vars.WINDOWS_GATE_REQUIRED != 'true' }}` on the Windows `go test`
step, and the variable is **unset**. A Windows-only test therefore is build-tagged out on
Ubuntu and has its Windows failure **masked**, with the job still reporting success to
`ci-gate`. **"Verified by CI" would be a false claim for this shipment.**

**Binding requirements on Ship for 013-S:**
1. Run `go test ./internal/pathsafe/...` **locally on Windows** and record the output as
   primary evidence in the PR body and closure artifact (precedent: 011-S, 012-S).
   Do **not** add `-race` to the Windows CI job — `windows-latest` has no cgo and
   `go test -race` fails closed with `go: -race requires cgo`.
2. Re-run the original F0 empirical reproduction and confirm it is now rejected.
3. Explicitly disclose in the PR body that Windows-only coverage is **advisory-in-CI**.
4. Do **not** flip `WINDOWS_GATE_REQUIRED` — a repository-configuration change requiring
   explicit operator confirmation (Constitution VII; precedent `011.010-T`).

**Escalated to the operator as a decision item; not a blocker for queuing 013-S.**

---

## 8. Deferred residuals — recorded, not silently dropped

Captured as risk-register entries by `014.008-T` AC 6 rather than as new stash entries,
because this dark run's scope is frozen at 22 stash IDs with no additions permitted.

| Residual | Why deferred |
|---|---|
| **Case-folding fail-open.** `pathEqual`/`pathHasPrefix` (`pathsafe.go:201-216`) use `strings.EqualFold` unconditionally on Windows. On a case-sensitivity-enabled NTFS/WSL tree an attacker can create an outside sibling differing from the root only by case; the resolved target is outside but `EqualFold` accepts it. | Already a known register item (`5FE4A7BE`). **Not** a declared member of any of the 4 selected stash entries — fixing it would be scope expansion. `014.008-T` AC 1 therefore forbids claiming `700B41CE` closed against this vector. |
| **F3 — in-root junction ancestor false-rejection.** A live junction ancestor whose target is *inside* the root is unconditionally rejected (`ModeIrregular` is neither `IsDir` nor `ModeSymlink`). | A **fail-closed usability** bug, not a containment failure. Not a declared member of any selected entry (scope-audit P2). Relaxing a control inside the shipment that closes a bypass is the wrong risk trade. |
| **GODEBUG both-settings empirical proof.** `014.004-T` AC 5 makes the fix GODEBUG-independent *by construction*; a subprocess re-exec harness proving it *empirically* is not built. | Guards a non-default toolchain configuration reachable only by deliberate override (Go 1.26.5 pins `winsymlink=1`). Scope-audit P3 flagged the dedicated harness as YAGNI. |

---

## 9. Verification strategy

1. **RED proof** — `014.001-T` C1/C2/C3 and `014.002-T` AC 1 fail at pre-change `HEAD`.
2. **GREEN proof** — full `go test ./internal/pathsafe/...` on **Windows** (local) and Ubuntu (CI).
3. **Bypass-closure proof** — re-run the F0 empirical reproduction from
   `docs/closure/2026-09-07-011-s-012-f-pathsafe-containment-adversarial-review.md`; must reject.
4. **Non-regression proof** — the three accept paths in `014.002-T` AC 2/3 stay green.
5. **Portability proof** — `go build ./...` on Windows, Linux, Darwin.
6. **Gate proof** — `scripts/check-write-path-precondition.sh --self-test` green.

---

## 10. Review findings and dispositions

### Rev 1 internal six-persona panel

| # | Persona | Sev | Finding | Disposition |
|---|---|---|---|---|
| R1 | Security | P0 | GREEN ordered before junction RED | **FIXED** — `014.004-T` hard-depends on `001`/`002` |
| R2 | Security | P0 | No fallback if handle resolution infeasible | **FIXED** — binding O3 fail-closed fallback |
| R3 | Correctness | P0 | RED test could be written as a subpath and false-negative | **FIXED** — `014.001-T` AC 1 |
| R4 | Correctness | P1 | Over-rejection breaks new-file writes | **FIXED** — `014.002-T` AC 2 |
| R5 | Constitution | P1 | Deleting the GO-14 lock erases evidence | **FIXED** — rename+invert mandated |
| R6 | Scope | P1 | Risk of pulling `BF5DE670` TOCTOU into scope | **FIXED** — explicit anti-goal |
| R7 | Architecture | P1 | Ancestor branch left with contradictory reparse policy | **SUPERSEDED** by A2 — resolved by canonicalization, not relaxation |
| R8 | Security | P1 | "Verified by CI" would be false | **FIXED** — §7 |
| R9 | Maintainability | P2 | `2362BBB5`(a) ≡ `AD0D9D1F`(F6) double-implement risk | **FIXED** — deduped to `014.005-T` |
| R10 | Constitution | P2 | Coverage exclusion not updated | **FIXED** — `014.005-T` AC 2 |
| R11 | Maintainability | P2 | Register/write-path-gate desync | **FIXED** — `014.008-T` AC 7 |
| R12 | Scope | P3 | `AD0D9D1F` F5/F10 already closed | **FIXED** — anti-goal, verified against `902c05f` |

### Rev 2 independent three-model adversarial panel

| # | Reviewer | Sev | Finding | Disposition |
|---|---|---|---|---|
| A1 | Security (gpt-5.6-sol) | **P1→P0** | Final-component-only fix leaves a **residual bypass**: intermediate junction with an *existing* leaf (C2), and existing dir beyond the junction (C3). `Lstat` traverses intermediates, so the reparse branch never fires | **FIXED** — root cause restated (§2) as *"containment checked against a non-canonicalized path"*; fix moved to whole-prefix canonicalization; `014.001-T` AC 2/3 add C2/C3 as mandatory RED |
| A2 | Correctness (gemini-3.8-flash) | **P0** | Relaxing the ancestor loop for in-root junctions while `filepath.EvalSymlinks` still cannot resolve mount points would **introduce** an intermediate-junction bypass | **FIXED** — F3 relaxation **removed from scope** (§8); `014.004-T` AC 4 retains `ModeIrregular` rejection; change is now monotonically fail-closed |
| A3 | Security (gpt-5.6-sol) | P1 | Case-folding fail-open (`EqualFold` on Windows) defeats containment on case-sensitive NTFS/WSL; plan cannot claim `700B41CE` fully closed | **ACCEPTED — deferred** (§8); `014.008-T` AC 1 forbids overclaiming |
| A4 | Correctness (gemini-3.8-flash) | P1 | Win32 syscalls cannot live in cross-platform `pathsafe.go`; needs build-tagged files, breaking the 1-file surface and 2-hour budget | **FIXED** — `014.003-T` split out as a platform-split helper task; surface updated (§3) |
| A5 | Correctness (gemini-3.8-flash) | P1 | Message-constant change breaks existing `TestResolveRejectsDanglingIntermediateSymlinkInsideRoot` and `TestResolveRejectsChildBeneathRegularFile`; unaccounted | **FIXED** — `014.006-T` AC 4 audits all 9 `symlinkEscapeMsg` assertion sites |
| A6 | Correctness (gemini-3.8-flash) | P2 | `014.007-T` (F11) must depend on the message task, not run parallel to it | **FIXED** — dependency retargeted to `014.006-T` |
| A7 | Correctness (gemini-3.8-flash) | P2 | O3 fallback contradicted `014.001-T` case (c) "must ACCEPT" | **FIXED** — case (c) removed with F3; `014.004-T` AC 6 notes C1–C3 all pass under O3 |
| A8 | Correctness (gemini-3.8-flash) | P2 | `014.003-T` oversized (syscalls + probe inversion + F3 + 2 RED suites) | **FIXED** — split into `014.003-T` (helper) and `014.004-T` (wiring) |
| A9 | Scope (claude-opus-4.8) | P2 | F3 not a declared member of any selected entry — unauthorized scope that *relaxes* the control | **FIXED** — F3 removed from scope, deferred (§8) |
| A10 | Scope (claude-opus-4.8) | P3 | Dedicated GODEBUG subprocess harness is YAGNI | **FIXED** — demoted to a by-construction AC (`014.004-T` AC 5); harness deferred (§8) |
| A11 | Scope (claude-opus-4.8) | P3 | Archiving `4989A42D` autonomously while operator is AFK is irreversible | **ACCEPTED with mitigation** — `backlogit stash archive` is non-destructive and retrievable; both contract conditions independently re-measured; surfaced in the Stage report for operator review |
| A12 | Correctness (gemini-3.8-flash) | P2 | `014.006-T` mixes production code and test updates (RED-phase purity) | **ACCEPTED** — a message-constant refactor with a same-task assertion audit is a single coherent unit; splitting would strand the suite RED mid-shipment. `complexity` raised to medium |

**No unresolved P0 or P1 findings remain.** Gate verdict: **PASS** (cycle 2 of 3).

Scope-audit confirmations carried forward: selection of Cluster A is correct under the
operator priority order; inclusion of `F133AB7E` is **legitimate in-scope work**, not scope
expansion, because `root.go`'s register contains an explicit flip procedure naming that stash
and the operator placed it in scope; all 22 scoped IDs carry explicit dispositions.

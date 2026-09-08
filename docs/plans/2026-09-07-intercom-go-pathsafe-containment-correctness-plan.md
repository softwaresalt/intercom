# Implementation Plan — pathsafe Containment Correctness (rev 2, post-review)

- **Date**: 2026-09-07
- **Source deliberation**: `docs/decisions/2026-09-07-intercom-go-pathsafe-containment-correctness-deliberation.md`
- **Stash entries covered**: ECE3DAB7, CA469B3D, 5FE4A7BE, 5158769F, 805248F7
- **Requires plan hardening**: **yes** — modifies a security containment control
  (`internal/pathsafe`) that Brief §10 marks NON-NEGOTIABLE
- **Revision 2** incorporates the resolution of one P0 and eight P1 findings from a
  four-reviewer gate (Security Lens, Correctness, Scope Boundary, and a 4-model
  Adversarial panel). See §8 for the disposition of every finding.
- **Constitution**: Principle IV (containment), Principle VIII (red-phase-first for
  containment code), Width Isolation (one skill domain per task)

---

## 1. Objective

Repair confirmed correctness defects in the workspace path-containment control and
remove two test-hygiene defects that mask regressions — without altering the
deliberately deferred TOCTOU/hardlink mitigation posture (BF5DE670) and without
touching any file owned by an excluded stash entry.

## 2. Anti-Goals (hard boundaries)

- **No** TOCTOU, O_EXCL, or hardlink mitigation. BF5DE670 stays deferred, trigger unfired.
- **No** change to the GO-14 dangling-symlink-at-**final**-component acceptance. Rev 2
  narrows the *claims* made about GO-14 but does not close it — reversing another
  shipment's deliberate, regression-locked risk acceptance is out of scope here and is
  captured to the stash instead (§8, ADV-04).
- **No** change to the `pathEqual` / `pathHasPrefix` case-folding predicate.
- **No modification to `.github/workflows/ci.yml`.** Rev 2 drops the macOS CI unit
  entirely. This file is the shared surface of excluded entries F4F4A959 and F47DB9A9
  and must not be opened by this shipment.
- **No modification to `scripts/check-unignore-regression.sh`** — the shared surface of
  excluded entries 2787DA56 and 4C5BEC23. In particular, its header comment is a known
  excluded finding (4C5BEC23) and must **not** be fixed as a drive-by.
- **No** modification to `scripts/check-retired-architecture.sh` (protects 4989A42D's
  cleanly measurable completion condition).
- **No** filesystem write primitive introduced in any **non-test** `.go` file under
  `internal/**` or `cmd/**` — `scripts/check-write-path-precondition.sh` must still pass
  and still report zero. (Test files legitimately call `os.Symlink`/`os.MkdirAll`; the
  gate scans non-test sources only.)

## 3. Units of Work

Every containment-code change is red-phase-first: failing tests land before the fix.

### U1 — RED: negative + positive test matrix for the ancestor walk (ECE3DAB7)

- **Domain**: tests. **Size**: M. **Complexity**: medium.
- Add to `internal/pathsafe/symlink_test.go`, gated through the existing
  `requireSymlinkOrFailClosed` predicate:
  - **N1** depth-1 dangling intermediate symlink → **outside**, non-existent target;
    `Resolve("danglinglink/new-file.txt")` must reject with `symlinkEscapeMsg`.
  - **N2** dangling intermediate symlink → **inside**-root, non-existent target; must
    also reject. This pins the §5 D-1 accepted over-rejection as a *decision*.
  - **N4** dangling link ≥2 levels above the leaf (`root/real/link/deep/x`) — proves the
    walk keeps climbing past the leaf's ENOENT and stops at the link itself.
  - **N5** symlink **chain** (`root/a -> root/b`, `root/b -> /outside/missing`) in
    intermediate position — transitively dangling.
  - **N6** **relative** symlink target (`os.Symlink("../../outside", ...)`) — exercises a
    different `EvalSymlinks` join path than the absolute targets all existing tests use.
  - **P1** *positive regression lock*: in-root symlinked directory with a **non-existent
    leaf** (`root/link -> root/inner`, `Resolve("link/new.txt")`) must still be
    **ACCEPTED**. No existing test covers this and it is the case most likely to regress
    from the `Stat`→`Lstat` change.
- **AC**: N1, N2, N4, N5, N6 are red against current `checkSymlinkEscape`; P1 is green
  both before and after U2. All assert `symlinkEscapeMsg` where rejection is expected.

### U2 — RED: privilege-free Windows junction coverage (ECE3DAB7) — resolves P0

- **Domain**: tests. **Size**: S. **Complexity**: medium.
- **File**: a NEW file `internal/pathsafe/junction_windows_test.go`. Go build
  constraints are **file-scoped**, not test-scoped, so this test MUST NOT be added to
  `symlink_test.go` — doing so would exclude every existing cross-platform symlink test
  from the Linux gate. The `_windows` filename suffix supplies the constraint.
- Create an intermediate **directory junction** (reparse point) to an outside target,
  with a non-existent leaf beneath it.
- **Fixture sequence (explicit — the junction must end up DANGLING)**: (1) create the
  outside target directory; (2) create the junction pointing at it; (3) **remove the
  target directory**, leaving the junction dangling; (4) call
  `Resolve("junction/new-file.txt")`. Junction-creation APIs generally require the
  target to exist at creation time, so a naive "point at a non-existent target" fixture
  may silently fail or produce a *non*-dangling junction — which would make J1 green
  before U3 and destroy its red phase.
- Junctions do **not** require `SeCreateSymbolicLinkPrivilege`, so this test must **NOT**
  be gated behind `requireSymlinkOrFailClosed` — it must run unconditionally on Windows.
  This is the point of the unit: N1/N2 skip on unprivileged Windows runners, leaving the
  platform with the widest attacker capability with zero test evidence.
- Reuse the junction-creation approach already proven in
  `tests/integration/output_path_guard_test.go`'s `createDirectoryJunction`.
- **AC**: test runs (does not skip) on a stock unprivileged Windows runner; the junction
  is verified dangling before `Resolve` is called; is red against current
  `checkSymlinkEscape`; asserts rejection.

### U3 — GREEN: depth-scoped probe + ENOENT-only ascent (ECE3DAB7) → U1, U2

- **Domain**: code. **Size**: S. **Complexity**: high.
- In `internal/pathsafe/pathsafe.go` `checkSymlinkEscape`, two changes:
  1. **Depth-scoped probe**: the **first** iteration (`ancestor == resolved`, the final
     component) keeps `os.Stat`; every **strict ancestor** uses `os.Lstat`.
  2. **ENOENT-only ascent**: ascend **only** when `errors.Is(err, fs.ErrNotExist)`. Any
     other probe error (`EACCES`, `EPERM`, `EIO`, `ELOOP`, malformed reparse point) must
     **reject**, wrapping the OS cause — not be silently treated as absence. Today a
     non-searchable intermediate directory (`root/private/link/new`, `private` unsearchable)
     lets the walk climb past a pre-planted escaping link.
- Update the function's own doc comment **in this same unit** — it currently asserts
  "os.Stat (not os.Lstat) gates each existence probe", which U3 makes false. A security
  control must not carry a comment contradicting its code, even transiently.
- **Known verdict change to assert deliberately**: `root/file.txt/sub` (a non-directory
  path component) currently yields `ENOTDIR`, which is **not** `fs.ErrNotExist`, so the
  walk will now **reject** instead of ascending-and-accepting. Re-review confirmed no
  existing test in `contain_test.go`, `root_test.go`, or `symlink_test.go` locks that
  case. The new rejection is fail-closed and correct, but Windows may classify the
  analogous error as `fs.ErrNotExist`, so add a small cross-platform verdict test pinning
  the intended behavior rather than leaving it platform-dependent by accident.
- **AC**: U1's N1/N2/N4/N5/N6 and U2's junction test pass; U1's P1 positive lock still
  passes; `TestResolveAllowsDanglingSymlinkAtFinalComponent` (GO-14) still passes;
  `TestResolveRejectsSymlinkedIntermediateDirEscapingRoot`,
  `TestResolveAllowsSymlinkInsideRoot`, `TestResolveRejectsSymlinkEscapingRoot`,
  `TestResolveAllowsNonExistentRelativePath` all still pass; the DOCUMENTED-UNREACHABLE
  filesystem-root branch remains unreachable; `go test ./...` green.

### U4 — Risk register: honest GO-14 boundary and residual (ECE3DAB7) → U3

- **Domain**: docs. **Size**: S. **Complexity**: medium.
- Update `root.go`'s consolidated risk register GO-14 entry to:
  - bound the acceptance explicitly to the **final component only**, including a
    *transitively* dangling final link (a chain ending unresolved);
  - state **plainly** that GO-14 remains an **open write-through primitive of the same
    class** that U3 fixes, reachable under the identical attacker capability with one
    fewer path component — U3 **narrows** the escape, it does not remove it;
  - record the accepted tightening: intermediate unresolvable entries are rejected
    regardless of target, because containment of a not-yet-existing target is
    unverifiable;
  - name the **class** of the new rejection ("directory entry exists but is not
    statable": dangling links, cycles, EACCES, unresolvable reparse points) so a future
    ELOOP rejection is not misread as a regression.
- **AC**: GO-14's **TRIGGER** sentence and the shared **RETIREMENT PROCEDURE** paragraph
  remain **byte-identical** (they are BF5DE670-attributed; only the acceptance *boundary*
  text changes). BF5DE670's TOCTOU and SEC-5 entries are unchanged. The word
  "fail-closed" is **not** used as an unqualified description of `checkSymlinkEscape`.

### U5 — RED: stripUNCPrefix regression test (CA469B3D)

- **Domain**: tests. **Size**: S. **Complexity**: medium.
- `stripUNCPrefix` is pure string manipulation with no GOOS-dependent behavior, so the
  **transform assertions run on all platforms** (giving coverage on the *required* ubuntu
  gate). Rev 1 build-tagged the whole test to windows, whose CI job is advisory — the fix
  would have had zero required-gate coverage.
- **Split across two files** (`filepath.IsAbs` and `filepath.VolumeName` are
  **host-dependent**: on Linux they do not interpret Windows UNC syntax at all, so
  asserting them cross-platform would fail the required gate):
  - `root_test.go` (all platforms) — **pure string transform** assertions only.
  - `root_windows_test.go` (new, Windows-only) — the `filepath.IsAbs` and
    `filepath.VolumeName(result) == \\server\share` assertions. `VolumeName` is asserted
    in addition to `IsAbs` because `IsAbs` on a bare share root is toolchain-version
    sensitive, while `VolumeName` is stable.
- Transform cases (all platforms): `\\?\UNC\server\share` → `\\server\share`;
  `\\?\UNC\server\share\ws\file` → `\\server\share\ws\file` (the real runtime shape);
  `\\?\C:\path` → `C:\path` (unchanged); lowercase `\\?\unc\...`; the `\\?\UNCfoo`
  near-miss; `\\?\UNC` with no trailing separator; `\\?\Volume{GUID}\...`;
  `\\?\GLOBALROOT\Device\...`.
- **AC**: red against current `stripUNCPrefix` for the UNC and Volume/GLOBALROOT cases;
  the all-platform file contains no host-dependent `filepath` call.

### U6 — GREEN: stripUNCPrefix fail-closed postcondition (CA469B3D) → U5

- **Domain**: code. **Size**: S. **Complexity**: medium.
- Two changes in `root.go`:
  1. Re-form a remainder beginning `UNC\` (matched **case-insensitively** and **including
     the trailing separator**) as `\\` + remainder-after-`UNC\`.
  2. **Class postcondition**: if the stripped result is not a Windows-absolute form,
     return the **original** `p` unchanged. Keeping the `\\?\` form is absolute and
     simply fails the prefix match → fail-closed. This closes `\\?\Volume{GUID}\` and
     `\\?\GLOBALROOT\` too, rather than fixing only the one reported shape.
- **The postcondition MUST use an OS-independent Windows-path classifier**, not host
  `filepath.IsAbs`. `stripUNCPrefix` is a pure string function exercised by
  cross-platform tests; using host `IsAbs` would make it retain the `\\?\` prefix on
  Linux and contradict U5's all-platform transform expectations. Implement a small
  local predicate recognising Windows-absolute forms (`\\server\...` UNC, and
  `X:\...` drive-absolute).
- **AC**: U5 passes on **both** linux and windows; all existing pathsafe tests pass
  unmodified; no non-Windows behavior change; `stripUNCPrefix` never returns a relative
  string on any host.

### U7 — Risk register: case-folding, both directions (5FE4A7BE)

- **Domain**: docs. **Size**: S. **Complexity**: medium.
- Add a register entry, carrying **its own task-ID attribution** and explicitly stating
  it is **not** a BF5DE670 item:
  - **darwin (under-folding)**: fold is Windows-only; darwin is a shipped target
    (`scripts/targets.json`) and is case-insensitive by default. STATUS: accepted,
    **fail-closed** (worst case is rejecting a legitimate path). Extending the fold was
    **rejected** because on a case-sensitive APFS volume it would convert rejection into
    acceptance — fail-**open**. The availability defect is labelled
    **plausible/unconfirmed** pending a reproduction.
  - **windows (over-folding)**: the fold that is **already enabled** is itself a
    potential fail-open. NTFS supports per-directory case sensitivity
    (`fsutil file setCaseSensitiveInfo`; WSL enables it by default on directories it
    creates), so on such a root a symlink resolving to `C:\ws\A\secret` can be folded
    into acceptance under root `C:\ws\a` — the exact construct rejected for darwin.
    TRIGGER: a workspace root on a case-sensitivity-enabled or WSL-created tree.
- No predicate change. The correction is to the register's claim, not the code.
- **AC**: both directions recorded with STATUS/TRIGGER; the Windows fold is **not**
  presented as the safe baseline; no code change in this unit.

### U8 — Fail-closed symlink helper in integration tests (5158769F)

- **Domain**: tests. **Size**: S. **Complexity**: medium.
- Add build-tagged `windows` / `!windows` privilege predicates to `tests/integration`
  mirroring `internal/pathsafe`'s `isSymlinkPrivilegeError` (ERROR_PRIVILEGE_NOT_HELD =
  1314). Rewrite `createDirectorySymlink` (`output_path_guard_test.go:78-89`).
- **Explicit per-platform behavior** (rev 1's AC was self-contradictory):
  - **Windows**: `t.Skipf` **only** on 1314; `t.Fatalf` on every other error.
  - **Non-Windows**: preserve the existing **unconditional skip**, exactly as
    `internal/pathsafe`'s `requireSymlinkOrFailClosed` does. Do **not** make non-Windows
    failures fatal — that would change behavior on the only *required* CI job.
- Duplication rationale to record: the predicate is not merely unexported, it is defined
  in `_test.go` files (`symlink_privilege_windows_test.go`,
  `symlink_privilege_other_test.go`) and is therefore unimportable even if exported. A
  shared `internal/testsupport` package was rejected as disproportionate for two call
  sites; the drift risk is noted in the helper's comment.
- **AC**: Windows non-privilege failures now fail; 1314 still skips with a reason;
  **non-Windows behavior unchanged**; existing subtests pass.

### U9 — Repair truncated test doc comments (805248F7)

- **Domain**: docs. **Size**: XS. **Complexity**: trivial.
- Restore the leading function name and opening clause on the two truncated comments.
- **Anchor by function name, not line number**: the comment above
  `TestNormalizeAcceptsCleanEquivalentForms` in `lexical_test.go` (currently at :69-70)
  and the comment above `TestResolveAllowsNonExistentRelativePath` in `symlink_test.go`
  (currently at :171). U1 appends tests to `symlink_test.go` and will shift those lines.
- **AC**: both comments name their function and read as complete sentences; no test logic
  changed.

## 4. Dependency Graph

```
U1 ─┐
    ├─→ U3 ──→ U4
U2 ─┘
U5 ──→ U6
U7   (independent; edits root.go register — sequence after U4 to avoid collision)
U8   (independent)
U9   (independent; sequence after U1 — same file)
```

**File-collision ordering (not logical dependencies, but must be respected):**
`U4`, `U6` and `U7` all edit `internal/pathsafe/root.go`, and `U4`/`U7` edit the **same**
package-doc register block — serialize them. `U1` and `U9` both edit
`internal/pathsafe/symlink_test.go` — serialize them. **U2 does NOT touch
`symlink_test.go`**: it creates a new `internal/pathsafe/junction_windows_test.go`, so it
has no collision with U1/U9. U5 adds cases to `root_test.go` and creates
`root_windows_test.go`; neither collides with `root.go`.

**Atomicity**: U1+U2+U3 MUST land in a single commit. The red observation is recorded in
the task notes, but shipping U1/U2 alone would leave the required ubuntu gate red on the
integration branch. Reverting a fix without its tests (or vice versa) is prohibited.

## 5. Verification

- `go build ./...` and `go vet ./...` clean.
- `go test ./...` green on linux (required gate); `internal/pathsafe` +
  `tests/integration` green on windows.
- U2's junction test **executes** (does not skip) on the Windows runner.
- `scripts/check-write-path-precondition.sh --self-test` passes and still reports **zero**
  write primitives (proves the BF5DE670 anti-goal held).
- `git diff --name-only` contains **no** entry for `.github/workflows/ci.yml`,
  `scripts/check-unignore-regression.sh`, or `scripts/check-retired-architecture.sh`
  (mechanically proves the §2 exclusion anti-goals held).

## 6. Risks

| Risk | Mitigation |
|---|---|
| U3 regresses GO-14 | U1 red-phase-first; GO-14 lock is an explicit U3 AC |
| U3's `Lstat` swap breaks in-root symlinked dir + new leaf | U1's **P1 positive lock** covers exactly this |
| U3's ENOENT gating causes false rejections | Only non-ENOENT errors reject; cause is wrapped and diagnosable |
| U6 breaks plain `\\?\C:\` handling | Existing tests must pass unmodified; explicit U5 case |
| U8 newly-failing tests on unprivileged Windows | 1314 remains an explicit skip; non-Windows untouched |
| Scope creep into excluded entries | §2 anti-goals + the §5 `git diff --name-only` mechanical assertion |

## 7. Plan Hardening (P-006)

### 7.1 Threat model for U3

**Asset**: the invariant that any path returned by `Root.Resolve` cannot be used to read
or write outside the workspace root.

**Attacker capability assumed**: can create symlinks/junctions inside the workspace (any
process running as the operator, including a compromised or misdirected agent), but
cannot modify `internal/pathsafe` itself.

**Attack before U3**: plant `root/link -> /outside/target` where `/outside/target` does
not yet exist. `Resolve("link/payload")` returns a blessed path. Later `/outside/target`
is created and the blessed path writes through the symlink outside the workspace. The
*dangling* state at validation time is what defeats the check, and it is fully
attacker-controlled. A second variant needs no symlink privilege on Windows at all: a
directory **junction** to a non-existent target produces the same shape.

**After U3**: the strict-ancestor `os.Lstat` probe observes the link/junction as an
existing directory entry, so the walk stops there; `filepath.EvalSymlinks` then fails and
the path is rejected. Separately, a non-ENOENT probe error no longer masquerades as
absence.

**RESIDUAL — stated plainly, not as a footnote**: `checkSymlinkEscape` is **NOT**
fail-closed as a whole after U3. Under the **identical** attacker capability, GO-14 leaves
an equally powerful arbitrary-write-outside-workspace primitive open at the **final**
component: plant `root/link -> /outside/payload` and call `Resolve("link")` rather than
`Resolve("link/payload")`. U3 costs the attacker exactly one path component. This
shipment **narrows** the escape class; it does not close it. U4 must encode this residual
in the durable register, and the deliberate reconsideration of GO-14 is captured to the
stash (§8, ADV-04) rather than silently deferred.

### 7.2 Mandatory negative/positive tests (Constitution VIII red-phase-first)

Each negative must be observed **failing** before its fix lands. A test that passes before
the fix is not evidence and must be rewritten. Every one has a named owning unit — rev 1
orphaned N2.

| ID | Owner | Must fail before | Asserts |
|---|---|---|---|
| N1 | U1 | U3 | depth-1 dangling intermediate → outside: reject |
| N2 | U1 | U3 | dangling intermediate → **inside**-root: reject (pins the accepted tightening) |
| N4 | U1 | U3 | dangling link ≥2 levels above leaf: reject |
| N5 | U1 | U3 | transitively dangling **chain**: reject |
| N6 | U1 | U3 | **relative** symlink target: reject |
| P1 | U1 | *(green before AND after)* | in-root symlinked dir + non-existent leaf: **accept** |
| J1 | U2 | U3 | Windows **junction**, privilege-free, non-skip-gated: reject |
| N3 | U5 | U6 | `\\?\UNC\...` → absolute; `Volume{GUID}`/`GLOBALROOT` never relative |

### 7.3 Regression locks that MUST stay green (verdict-change tripwires)

`TestResolveAllowsDanglingSymlinkAtFinalComponent` (GO-14),
`TestResolveRejectsSymlinkedIntermediateDirEscapingRoot`,
`TestResolveAllowsSymlinkInsideRoot`, `TestResolveRejectsSymlinkEscapingRoot`,
`TestResolveAllowsNonExistentRelativePath`, `TestResolveRejectsTraversal`, and every
`internal/pathsafe` + `tests/integration` test unmodified by this plan.

Any change in these verdicts is a **stop condition**, not a test to update. **Qualifier
(rev 2)**: the GO-14 lock encodes a *risk acceptance*, not a functional requirement. It
may not be flipped in **this** shipment, but U4 must name the condition under which a
future shipment may flip it, so the lock does not become structurally un-revisitable.

### 7.4 Blast radius

`internal/pathsafe` currently has **no production callers** performing filesystem writes
(`check-write-path-precondition.sh` reports zero across all non-test `.go` files under
`internal/**` and `cmd/**`). The practical blast radius of U3/U6 is confined to the
package and its tests — the cheapest possible moment to correct the control, and an
argument for fixing it **now**, before phases C4–C6 introduce real write paths.

### 7.5 Rollback

Each unit is an independent revert, except the atomic U1+U2+U3 triple (§4). U6 is a
single-function change with red-phase tests attached; revert the pair, never one half.

### 7.6 Options considered and rejected as in-scope work

- Closing GO-14 (wholesale `Lstat`, or reject-final-when-`Lstat`-shows-symlink) —
  reverses another shipment's regression-locked acceptance; captured to stash instead.
- Chasing symlink targets manually during the walk — reimplements symlink resolution
  inside a security control.
- Extending case-folding to darwin — fail-open on case-sensitive volumes.
- Adding an advisory macOS CI job — dropped in rev 2 (§8, ADV-03).
- `Lstat`-before-write / `O_EXCL` — BF5DE670, trigger unfired.

## 8. Review Gate Disposition (rev 1 → rev 2)

Four independent reviewers. **All P0 and P1 findings are resolved.**

| ID | Sev | Finding | Resolution |
|---|---|---|---|
| ADV-01 | **P0** | Windows junctions are a privilege-free instance of the attack with zero coverage; N1/N2 skip on unprivileged Windows | **Resolved** — new unit **U2**, non-skip-gated junction test (J1) |
| SEC-1 | P1 | §7.1 "fail-closed" overstates; GO-14 leaves the same primitive open one component lower, and U4 would bake that understatement into the durable register | **Resolved** — §7.1 RESIDUAL rewritten plainly; U4 AC forbids unqualified "fail-closed"; GO-14 reconsideration captured to stash |
| ADV-05 | P1 | Non-ENOENT probe errors (EACCES/ELOOP/EIO) treated as absence; walk climbs past a pre-planted link | **Resolved** — U3 now gates ascent on `errors.Is(err, fs.ErrNotExist)` and wraps other causes |
| ADV-02 | P1 | Group A's "single-package coherence" claim is inaccurate because U7(rev1) was YAML | **Resolved** — macOS CI unit dropped; every remaining unit is `internal/pathsafe/**` or its integration test |
| ADV-03 | P1 | macOS CI job cannot test the case-sensitivity risk it claims (runners are case-insensitive APFS); permanently advisory | **Resolved** — dropped. 5FE4A7BE's "and/or" phrasing means U7's register entry discharges it |
| SCOPE-1 | P1 | Anti-goals did not confine `ci.yml` edits — highest scope-creep path into excluded F4F4A959 | **Resolved** — `ci.yml` now a hard anti-goal (not edited at all) + mechanical `git diff --name-only` assertion |
| SEC-2 | P2 | GO-14 lock made structurally un-revisitable by §7.3 | **Resolved** — §7.3 qualifier added; U4 must name the flip condition |
| SEC-4 | P2 | `stripUNCPrefix` fixed one shape; `Volume{GUID}`, `GLOBALROOT`, lowercase `unc\` still relative | **Resolved** — U6 adds a fail-closed class postcondition + case-insensitive match |
| SEC-5 | P2 | N1/N2 covered only depth-1; no chain/relative/positive-lock coverage | **Resolved** — U1 expanded to N1,N2,N4,N5,N6 + positive lock P1 |
| SEC-6 / ADV-07 | P2 | Windows fold is *itself* fail-open on case-sensitive NTFS; register would certify it as the safe baseline | **Resolved** — U7 records both directions |
| CORR-1 / ADV-12 | P2 | U8's "non-Windows unchanged" AC contradicted its own spec; would make every non-Windows symlink failure fatal on the required gate | **Resolved** — U8 now specifies per-platform behavior explicitly |
| CORR-2 / SCOPE-7 | P2 | Mandatory test N2 had no owning unit | **Resolved** — §7.2 table assigns an owner to every test |
| CORR-3 | P2 | U4(rev1) build-tagged the UNC test windows-only → zero required-gate coverage | **Resolved** — U5 transform assertions run on all platforms |
| CORR-4 | P2 | U5(rev1) was un-threat-modeled although it widens the accept set | **Resolved** — U6's fail-closed postcondition means the accept set does not widen for non-absolute results |
| CORR-5 | P2 | U5 left `UNC\` case-sensitivity, `UNCfoo` near-miss, bare `UNC` undetermined | **Resolved** — enumerated as explicit U5 cases |
| SCOPE-6 | P2 | U3(rev1) could partially consume BF5DE670 by rewriting a BF5DE670-attributed register entry | **Resolved** — U4 AC pins TRIGGER + RETIREMENT PROCEDURE byte-identical |
| SCOPE-8 | P2 | No anti-goal protected `check-unignore-regression.sh` (excluded 2787DA56/4C5BEC23) | **Resolved** — explicit anti-goal + diff assertion |
| CORR-9 | P3 | `checkSymlinkEscape` doc comment would contradict its code between units | **Resolved** — comment update folded into U3 |
| CORR-10 | P3 | U1 landing red on main | **Resolved** — §4 atomicity clause |
| CORR-13 | P3 | U9 anchored to a line number U1 shifts | **Resolved** — U9 anchors by function name |
| SCOPE-9 | P3 | Darwin risk would inherit BF5DE670 attribution | **Resolved** — U7 carries its own attribution |
| SCOPE-10 | P3 | Write-path anti-goal read broader than the gate | **Resolved** — "non-test" qualifier added |
| SEC-7 | P3 | Walk terminal branch returns accept (fail-open default) | **Deferred, recorded** — genuinely unreachable (both reviewers concur); flipping it touches a coverage-excluded branch. Captured to stash. |
| SEC-8 / ADV-09 | P3 | `symlinkEscapeMsg` misleading for the unverifiable-target case | **Deferred, recorded** — U4 documents that the message covers both escape and unverifiable-target; a distinct message is a behavior change needing its own AC. Captured to stash. |
| ADV-04 | P1→note | ECE3DAB7 vs BF5DE670 exploitability framing "rhetorically exaggerated" | **Resolved** — deliberation §4 language corrected; both share the same unfired write-caller trigger |
| ADV-08 | P2 | F4F4A959 is not truly operator-policy-blocked (default-preserving toggle) | **Accepted, no action this cycle** — stays excluded by the one-shipment limit; noted for next cycle |
| ADV-11 | P2 | 2787DA56 severity overstated | **Accepted** — remains excluded; reclassification noted for next cycle |
| P3-typo | P3 | Deliberation scope list duplicated F4F4A959 (14 tokens for 13 entries) | **Resolved** — corrected |

### 8.1 Second-pass re-review (rev 2 → rev 2.1)

A post-remediation re-review confirmed all P0/P1 items above as resolved, and verified:
U3's ENOENT-only ascent does not break the normal new-file case; U1's positive lock P1
is correctly traced; 5FE4A7BE is honestly discharged by U7 without the dropped CI job.
It raised three **new P1 plan-precision defects**, all now fixed:

| ID | Sev | Finding | Resolution |
|---|---|---|---|
| RR-1 | P1 | U2 was directed into `symlink_test.go`, but Go build constraints are **file-scoped** — a `//go:build windows` tag there would exclude every existing cross-platform symlink test from the required Linux gate | **Resolved** — U2 now creates a separate `internal/pathsafe/junction_windows_test.go`; §4 collision set corrected |
| RR-2 | P1 | U2's dangling-junction fixture was underspecified; junction-creation APIs generally require the target to exist, so a naive fixture would produce a **non**-dangling junction, making J1 green before U3 and destroying its red phase | **Resolved** — U2 now specifies the explicit create-target → create-junction → remove-target sequence and an AC verifying the junction is dangling before `Resolve` |
| RR-3 | P1 | U5/U6 used host-dependent `filepath.IsAbs` / `filepath.VolumeName` in all-platform assertions; on Linux these do not interpret Windows UNC syntax, so the required Linux gate would fail, and U6's postcondition would retain the `\\?\` prefix on Linux | **Resolved** — U5 split into an all-platform pure-transform file plus `root_windows_test.go` for the host-dependent assertions; U6's postcondition now mandates an OS-independent Windows-path classifier |
| RR-4 | P2 | `root/file.txt/sub` (ENOTDIR, not ENOENT) changes from accept to reject under U3; no existing test locks it | **Resolved** — U3 now calls out the verdict change explicitly and requires a cross-platform test pinning it |

<!-- plan-review-attempt: 3 -->

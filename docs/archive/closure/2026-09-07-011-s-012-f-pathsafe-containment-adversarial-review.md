---
title: "Adversarial Review — 011-S / 012-F Pathsafe Containment Correctness and Cross-Platform Coverage"
date: 2026-09-07
shipment: 011-S
feature: 012-F
branch: feat/011-s-pathsafe-containment-correctness-and-cross-platform-coverage
head_commits: [f04b332, 4605076, 0340a7c, 0680667, 451fe4c, a1a7c34, 2e5f661, 9158a05, f8ac28d]
review_type: adversarial-review (multi-model consensus)
reviewers: 4
output_mode: full
verdict: BLOCKED
---

# Adversarial Review — 011-S / 012-F Pathsafe Containment Correctness and Cross-Platform Coverage

## Scope note

This session's tool environment does not expose direct shell/git execution
to the orchestrating agent; only `view`/`create`/`edit` and agent-dispatch
tools are available. The literal `git diff main..HEAD` could not be run by
the orchestrator. Adaptation: the orchestrator read the current HEAD content
of every file in the user-specified scope directly (`internal/pathsafe/*.go`,
`tests/integration/*_test.go`, `internal/apperr/*.go`), and gave each
reviewer the same file list plus the user's characterization of what changed
vs. `main`, instructing each reviewer to independently read the files itself
before answering. **One reviewer (Tier 3 / Reviewer-C) additionally had shell
access and was asked to verify a specific claim empirically; that attempt was
blocked by a sub-agent recursion-depth limit.** The orchestrator then
independently verified the resulting critical claim by reading the actual
pinned Go toolchain source (`C:\Program Files\Go\src\os\stat_windows.go`,
`types_windows.go`, `path\filepath\symlink.go`, `symlink_windows.go` — Go
1.26.5, matching `go.mod`'s pinned toolchain) rather than trusting the
reviewer's static claim at face value. That independent trace is included
below and materially changes the verdict from what raw vote-counting alone
would produce.

**Housekeeping disclosure:** a dispatched verification sub-agent created a
scratch file `internal/pathsafe/zzz_verify_junction_test.go` and could not
execute or delete it due to the same recursion-depth limit. The orchestrator
neutralized it in place (gated behind `//go:build windows && ignore`, so it
compiles into nothing under any normal `go build`/`go vet`/`go test`
invocation) and left a comment instructing deletion via
`git clean -fd internal/pathsafe`. **This file is untracked, was never part
of the reviewed diff, and must be deleted before merge — it is not a
shipment artifact.**

## Phase 1 — Reviewer pool and model routing

Anchor route (`openai` / `gpt-5.6-sol`) was dispatchable — all four models
resolved successfully. No alternate provider was configured. Reviewer count:
4 (default with anchor dispatchable).

| Slot | Route | Model | Reasoning effort |
|---|---|---|---|
| Anchor Reviewer | Anchor | `gpt-5.6-sol` | high |
| Reviewer-A | Tier 1 (fast/cheap) | `gpt-5.4-mini` | high |
| Reviewer-B | Tier 2 (standard) | `claude-sonnet-5` | high |
| Reviewer-C | Tier 3 (frontier) | `claude-opus-5` | high |

All four reviewers returned structured JSON findings before Phase 3
aggregation began (quality criterion met).

## Phase 3–4 — Aggregation and scoring

`reviewers = 4` → HIGH/Consensus requires 4/4; MEDIUM/Majority requires ≥3/4;
MEDIUM/Plurality is exactly 2/4; LOW/Unique is 1/4. **No finding reached
Consensus or Majority** — reviewers converged on the same *areas*
(Windows case-folding, junction handling, UNC edge cases) more often than on
identical file+line+rule keys, which plurality/unique buckets below reflect.

One finding (D1, the live-junction containment bypass) was raised by only
one reviewer (Reviewer-C) and would nominally be LOW confidence by raw vote
count. **The orchestrator independently verified it against the pinned Go
1.26.5 standard library source and confirms it is real, deterministic, and
requires no elevated privilege.** Per this protocol's own instruction to
apply human judgment to unique findings, and consistent with the assignment
brief's explicit instruction not to rubber-stamp prior conclusions, this
finding is escalated to **HIGH confidence / blocking** despite its 1-of-4
raw vote count. This is the single most important output of this review.

Priority = confidence_weight (HIGH=3, MEDIUM=2, LOW=1) × severity_weight
(CRITICAL=4, MAJOR=3, MINOR=2).

---

## Escalated finding (verified independently — overrides raw vote count) — BLOCKING

### F0. Live Windows directory junction at the final path component bypasses `checkSymlinkEscape` entirely

- **Severity:** CRITICAL. **Confidence:** HIGH (independently verified against Go 1.26.5 stdlib source, not merely asserted).
- **Raised by:** Reviewer-C only (1/4 by raw count) — escalated by the orchestrator after independent verification.
- **File:** `internal/pathsafe/pathsafe.go`, `checkSymlinkEscape`.
- **Verified: yes** — traced through `C:\Program Files\Go\src\os\stat_windows.go`, `types_windows.go`, and `path/filepath/symlink.go`/`symlink_windows.go` at the pinned toolchain version (`go1.26.5`, matching `go.mod`).

**Mechanism, traced step by step:**

1. `os.Stat` on Windows (`stat(funcname, name, followSurrogates=true)` in
   `stat_windows.go`) explicitly re-opens a name-surrogate reparse point
   (a junction is `IO_REPARSE_TAG_MOUNT_POINT`, which
   `isReparseTagNameSurrogate()` matches, same bit test as symlinks) and
   returns `FileInfo` describing the **resolved target**, not the reparse
   point itself. So `os.Stat(<root>/junction-out)` succeeds and reports
   `IsDir() == true` even when `junction-out` points at a directory
   entirely outside the workspace root.
2. In `checkSymlinkEscape`, the **first** probe (`finalProbe = true`, using
   `os.Stat`) therefore succeeds immediately on a live junction. The guard
   `!finalProbe && !info.IsDir() && info.Mode()&fs.ModeSymlink == 0` cannot
   fire (`finalProbe` is still `true`), so the loop breaks with `ancestor`
   still equal to the original, unresolved `resolved` path — the ancestor
   walk never climbs past the junction itself.
3. The final `filepath.EvalSymlinks(ancestor)` call does **not** resolve
   the junction either. `walkSymlinks` (`path/filepath/symlink.go`) calls
   `os.Lstat` on each component; for a junction, `fileStat.mode()`
   (`os/types_windows.go`) explicitly skips `ModeDir` for any
   name-surrogate reparse point and only sets `ModeSymlink` for the literal
   `IO_REPARSE_TAG_SYMLINK` case — a mount point (`IO_REPARSE_TAG_MOUNT_POINT`)
   falls into the `default: m |= ModeIrregular` branch. Back in
   `walkSymlinks`, `fi.Mode()&fs.ModeSymlink == 0` is therefore true for a
   junction, and because it is the **final** path component
   (`end == len(path)`), the `!fi.Mode().IsDir() && end < len(path)` guard
   also cannot fire. The loop falls through to `continue`, and
   `walkSymlinks` returns the **lexically unchanged path**.
4. `real` therefore equals `<root>/junction-out` (a string that still
   starts with `root.path`), `hasPathPrefix(real, root.path)` is `true`,
   and `checkSymlinkEscape` returns `(resolved, nil)` — **accepted**.

**Net effect:** `Root.Resolve("junction-out")`, where `<root>/junction-out`
is a *live* directory junction pointing at any directory outside the
workspace, is accepted as a valid, contained path. This is a deterministic,
100%-reproducible bypass on any Windows host, requires **no elevated
privilege or Developer Mode** (junction creation, unlike `os.Symlink`, needs
no `SeCreateSymbolicLinkPrivilege` — this is precisely why junctions are the
standard unprivileged redirection primitive on Windows), and directly
contradicts `Resolve`'s documented postcondition ("the returned path is
absolute, cleaned, and strictly contained within the root", finding GO-15).

**Why this survived the stated fix:** the shipment's own target finding
(ECE3DAB7) is "dangling intermediate symlink/junction containment bypass."
The new ancestor-walk logic correctly closes the *dangling* case (confirmed
by `TestResolveRejectsDanglingIntermediateJunctionOutsideRoot` — a dangling
junction's `os.Lstat` still reports non-dir/non-symlink reparse metadata,
which the ancestor guard catches). But the fix's own doc comment claims
generally that "a dangling intermediate symlink or junction is treated as
an existing directory entry that must be re-resolved rather than silently
walked past" — this is not true for a **live** junction at the **final**
path component, which is the shape a caller most naturally produces by
resolving a workspace-relative path whose last segment is itself a junction
(e.g. a relocated build-output or vendor directory).

**GODEBUG dependency (confirms root cause, does not excuse it):** this
behavior is specific to Go's default `winsymlink=1` semantics (Go 1.23+).
Under `GODEBUG=winsymlink=0`, `fileStat.modePreGo1_23()` maps
`IO_REPARSE_TAG_MOUNT_POINT` to `ModeSymlink`, and `walkSymlinks` would
correctly follow and resolve the junction. The containment verdict for this
NON-NEGOTIABLE control currently depends on an unpinned, undocumented
process-wide `GODEBUG` default — a second, independent defect (see F5).

**Required fix (must land before merge):** `checkSymlinkEscape` must not
rely on `os.Stat` + `filepath.EvalSymlinks` alone for the final path
component on Windows. It needs to detect a live name-surrogate reparse
point (junction/mount point) at the final component specifically and either
(a) resolve its true target explicitly (e.g. via
`GetFinalPathNameByHandle`/`os.Readlink` semantics that do resolve mount
points) and re-run the containment check against that target, or (b) pin
`GODEBUG=winsymlink=0` for the whole module (which also happens to fix F2
below) and add a regression test proving the same verdict under both
settings. **A new test creating a live (non-dangling) junction pointing
outside the root, with the junction as the literal final path component,
must be added and must fail before the fix and pass after** — this exact
shape is currently entirely untested (see F1).

---

## Plurality findings (MEDIUM confidence, 2/4)

### F1. No test coverage for a *live* (non-dangling) junction, in any position

- **Severity:** MAJOR. **Confidence:** MEDIUM (2/4 — Anchor + Reviewer-C, independently, at different depths).
- **File:** `internal/pathsafe/junction_windows_test.go`.
- **Verified: yes** — the file contains exactly one test function,
  `TestResolveRejectsDanglingIntermediateJunctionOutsideRoot`, which
  deletes the junction's target before calling `Resolve`. No test anywhere
  in `internal/pathsafe` creates a *live* junction. This is the direct
  coverage gap that let F0 (and F3 below) ship undetected, despite the
  shipment's stated goal of closing symlink/junction containment bypasses.
- **Action:** add, at minimum: (a) live junction outside root as the final
  component — must reject (currently does not, see F0); (b) live junction
  outside root as an intermediate ancestor with a missing leaf — must
  reject; (c) live junction **inside** root as an intermediate ancestor
  with a missing leaf — must accept (currently does not, see F3).

### F2. Windows case-folding in `hasPathPrefix`/`pathEqual`/`pathHasPrefix` is fail-open on case-sensitive NTFS/WSL trees

- **Severity:** MAJOR (as newly reasserted). **Confidence:** MEDIUM (2/4 — Anchor + Reviewer-A).
- **File:** `internal/pathsafe/pathsafe.go` (unchanged logic, not part of this diff) / `internal/pathsafe/root.go` (risk-register doc entry `5FE4A7BE`, which *is* part of this diff).
- **Verified: yes, but already disclosed and accepted.** `strings.EqualFold`
  folding is unconditional on `runtime.GOOS == "windows"`. On a
  case-sensitivity-enabled NTFS directory or a WSL-created tree, two
  distinct directories differing only by case can be folded into a false
  "contained" verdict. **This is not a new defect introduced by this diff**
  — `hasPathPrefix`/`pathEqual`/`pathHasPrefix` are unchanged, and the risk
  register entry `5FE4A7BE` that this diff *adds/updates* in `root.go`
  already states, verbatim: "STATUS: accepted, fail-open risk. TRIGGER: a
  workspace root on a case-sensitivity-enabled NTFS or WSL-created tree."
  Two reviewers converging on an already-disclosed, already-accepted risk
  is a useful cross-check that the register's own self-assessment is
  accurate, but it is **advisory, not blocking** for this shipment — the
  register already carries this exact risk with the correct fail-open
  direction. No action required beyond what F9 (register clarity) already
  recommends.

### F3. Any live junction used as a strict ancestor is unconditionally rejected, even when its target is inside the root

- **Severity:** MAJOR. **Confidence:** MEDIUM (2/4 — independently derived by Reviewer-C; the orchestrator additionally verified it by trace).
- **File:** `internal/pathsafe/pathsafe.go`, `checkSymlinkEscape`.
- **Verified: yes.** For any *ancestor* (non-final) probe, `os.Lstat` is
  used. `Lstat` does **not** follow the reparse point (`followSurrogates =
  false`), so it returns metadata for the junction itself:
  `isReparseTagNameSurrogate()` is true, so `ModeDir` is never set, and the
  reparse-tag switch only sets `ModeSymlink` for `IO_REPARSE_TAG_SYMLINK` —
  a mount point falls to `ModeIrregular`. The guard
  `!finalProbe && !info.IsDir() && info.Mode()&fs.ModeSymlink == 0` is
  therefore **always true** for any junction ancestor, regardless of
  whether its target is inside or outside the workspace root — `filepath.EvalSymlinks`
  is never reached and the target is never actually compared to the root.
  This contradicts the function's own doc comment, which states an
  intermediate junction "must be re-resolved rather than silently walked
  past" (it is never resolved at all) and contradicts risk-register entry
  `GO-14`'s framing that only *dangling* ancestors are rejected. This is a
  false-rejection (fails closed, not a security hole) but it means any
  Windows workspace containing a legitimate in-root junction (a common
  shape for relocated build/output directories or package-manager links)
  becomes partially unusable, and the doc comments describing the
  mechanism are inaccurate.
- **Action:** either fix per F0's remediation (which incidentally also
  fixes this), or explicitly document this as a known, intentional
  fail-closed limitation distinct from the dangling-symlink case, and
  correct the doc comment's claim that junctions are "re-resolved."

### F4. `TestResolveRejectsChildBeneathRegularFile` does not exercise the branch it is named for, on POSIX

- **Severity:** MINOR. **Confidence:** MEDIUM (2/4 — Reviewer-B and Reviewer-C, independently).
- **File:** `internal/pathsafe/contain_test.go`.
- **Verified: yes.** On POSIX, `os.Stat("<root>/file.txt/sub")` fails with
  `ENOTDIR`, not `ENOENT`. `syscall.Errno.Is` on Unix maps `fs.ErrNotExist`
  only to `ENOENT` (and `ENOTDIR` is not in that mapping), so
  `errors.Is(err, fs.ErrNotExist)` is `false` and the **first** probe
  (`finalProbe = true`) takes the generic "other error" branch
  (`apperr.Wrapf(..., symlinkEscapeMsg)`) immediately — the test's
  assertion (`strings.Contains(err.Error(), symlinkEscapeMsg)`) passes, but
  the ancestor-walk `!finalProbe && !info.IsDir() && ...` guard the test
  is named for is **never reached** on POSIX. (On Windows, the analogous
  failure is `ERROR_PATH_NOT_FOUND`, which Go *does* map to `ErrNotExist`,
  so the intended branch likely *is* exercised there — making this test's
  effective coverage platform-dependent and silently degraded on POSIX
  runners.)
- **Action:** assert on the unwrapped cause (e.g. `errors.As` to the
  underlying error type, or two separate platform-gated tests) so each
  platform's actual code path is distinctly pinned, not just the shared
  output string.

### F5. `stripUNCPrefix`'s UNC re-formation accepts a share-less remainder as "Windows absolute"

- **Severity:** MINOR. **Confidence:** MEDIUM (2/4 — Reviewer-B and Reviewer-C, independently, with Reviewer-C giving the more severe framing).
- **File:** `internal/pathsafe/root.go`, `isWindowsAbsolutePath`.
- **Verified: yes.** For input `\\?\UNC\server` (share segment entirely
  absent), the transform yields `remainder = \\server`, and
  `isWindowsAbsolutePath(\\server)` returns `true` (its UNC branch only
  checks for a leading `\\`, not a complete `\\server\share` structure).
  `\\server` alone is not a well-formed, addressable UNC path. This exact
  input is **not** covered by any existing test — `root_test.go` and
  `root_windows_test.go` test the full `\\?\UNC\server\share[...]` form and
  the "near miss"/"bare marker" non-transforming cases, but not the
  transforming-but-still-malformed middle case.
- **Real-world exploitability: low.** A legitimate Windows API
  (`os.Readlink`/`filepath.EvalSymlinks`) resolving a real network path
  would not plausibly emit a share-less `\\?\UNC\server` form, since a UNC
  path requires a share to reference any filesystem object at all — this
  is a defensive/robustness gap, not a demonstrated live bypass, and
  `NewRoot`'s subsequent `os.Stat(canonical)` call would fail on a
  malformed result in practice. Still worth tightening and testing.
- **Action:** require `filepath.VolumeName(remainder)` to name both host
  and share before accepting the re-formed UNC path; otherwise fall back to
  the original extended-path input, consistent with the function's existing
  conservative fallback for other malformed forms. Add table cases for
  `\\?\UNC\`, `\\?\UNC\server`, and `\\?\UNC\server\`.

---

## Unique findings (LOW confidence) — human judgment required

*(F0 above was originally unique but has been escalated based on independent
verification; it is listed separately, not repeated here.)*

### F6. `parent == ancestor` "documented unreachable" branch returns *accept*, not *reject*

- **Severity:** MINOR-MAJOR (framed differently by each). **Confidence:** LOW-MEDIUM (Reviewer-B and Reviewer-C both flagged the same branch from different angles — a network-share-disconnect scenario and a UNC-share-root scenario, respectively — treated here as one unique-cluster finding since neither reviewer's specific triggering scenario was independently confirmed by the orchestrator).
- **File:** `internal/pathsafe/pathsafe.go`, the `parent == ancestor` branch inside `checkSymlinkEscape`.
- **Verified: partially.** The code and its own comment are accurately
  quoted by both reviewers. The orchestrator did not independently
  reproduce either reviewer's specific triggering precondition (a volume
  becoming unreachable mid-session, or a UNC share root whose own probe
  errors classify as `fs.ErrNotExist`) — both require an unusual
  environment change between `NewRoot` construction and a later `Resolve`
  call, which is arguably already inside the documented TOCTOU acceptance,
  but the *direction* of the failure mode here (silent accept, not reject)
  is a legitimate defense-in-depth concern for a NON-NEGOTIABLE control
  regardless of how narrow the trigger window is.
- **Action (advisory, non-blocking):** consider changing this defensive,
  already coverage-excluded branch to fail closed
  (`return "", apperr.New(apperr.KindPathViolation, symlinkEscapeMsg)`)
  instead of `return resolved, nil`. This changes a "should never happen"
  branch's failure direction from silent-accept to loud-reject, which is
  strictly safer and costs nothing in the documented-unreachable case.

### F7. GODEBUG `winsymlink` dependency is unpinned and untested

- **Severity:** MINOR (as an independent finding; it is the root cause of F0/F3, which are already blocking). **Confidence:** LOW (1/4 — Reviewer-C only).
- **File:** `internal/pathsafe/pathsafe.go`.
- **Verified: yes** (see F0's mechanism trace — confirmed via
  `os/types_windows.go`'s `winsymlink` godebug switch between `mode()` and
  `modePreGo1_23()`).
- **Action:** once F0/F3 are fixed, add a test (or a documented go.mod
  `godebug` pin) proving the containment verdict for junctions is
  consistent regardless of the `winsymlink` setting, so a future Go
  toolchain upgrade or an operator's `GODEBUG` environment variable cannot
  silently flip this NON-NEGOTIABLE control's behavior again.

### F8. `uncPrefix` doc comment misattributes the prefix's origin to `GetFinalPathNameByHandle`

- **Severity:** MINOR. **Confidence:** LOW (1/4 — Reviewer-C only).
- **File:** `internal/pathsafe/root.go`, the `uncPrefix` doc comment.
- **Verified: not independently confirmed to full certainty**, but
  plausible: `path/filepath/symlink_windows.go`'s `evalSymlinks` calls only
  `walkSymlinks` + `toNorm`, neither of which calls
  `GetFinalPathNameByHandle`; the `\\?\` prefix most likely surfaces from
  `os.Readlink`'s reparse-target decoding, not handle-based canonicalization.
  The doc inaccuracy is not merely cosmetic: it implies a resolution
  mechanism that, if true, would have prevented F0 — so the comment's
  error is itself informative about how the fix's author reasoned about
  (and slightly misjudged) what `EvalSymlinks` actually does on Windows.
- **Action:** correct the attribution; low urgency, advisory only.

### F9. Risk-register entry `5FE4A7BE` embeds two contradictory `STATUS`/`TRIGGER` pairs under one identifier

- **Severity:** MINOR. **Confidence:** LOW (1/4 — Reviewer-C only).
- **File:** `internal/pathsafe/root.go`, package doc comment.
- **Verified: yes** — the entry as written contains "STATUS: accepted,
  fail-closed, plausible/unconfirmed" (for the darwin under-fold case)
  followed later, within the same bullet, by "STATUS: accepted, fail-open
  risk" (for the Windows over-fold case). A register whose entries are
  meant to be individually resolvable at a single trigger cannot cleanly
  carry two opposite-direction dispositions under one identifier — a
  future reader could easily latch onto the first (weaker, fail-closed)
  status and miss the second (fail-open, actually-gating) one.
- **Action (advisory, non-blocking; documentation clarity only):** split
  into two register entries with distinct identifiers, one per direction,
  each with exactly one `STATUS`/`TRIGGER` pair.

### F10. `junction_windows_test.go`'s `createDirectoryJunction` hard-fails instead of skipping when `pwsh` is absent

- **Severity:** MINOR. **Confidence:** LOW (1/4 — Reviewer-C only).
- **File:** `internal/pathsafe/junction_windows_test.go`.
- **Verified: yes** — the helper invokes `pwsh` unconditionally and calls
  `t.Fatalf` on any failure, with no `exec.LookPath("pwsh")` probe.
  `tests/integration/output_path_guard_test.go`'s `TestResolveOutputPathWithinRoot`
  guards the identical dependency with `exec.LookPath("pwsh")` + `t.Skip`.
  The two suites disagree about the same optional-tooling precondition; the
  011.010-T fail-closed policy is specifically about symlink *privilege*,
  not about the absence of an optional shell binary.
- **Action (advisory, non-blocking):** add the same `exec.LookPath` guard
  (skip, not fail, when `pwsh` is absent) for consistency with the
  integration suite.

### F11. Untested permission-denied / cycle branch mid-walk

- **Severity:** MINOR. **Confidence:** LOW (1/4 — Reviewer-B only).
- **File:** `internal/pathsafe/pathsafe.go`, the
  `!errors.Is(err, fs.ErrNotExist)` branch in `checkSymlinkEscape`.
- **Verified: not independently reproduced** (would require an
  EACCES-inducing or ELOOP-inducing fixture, platform- and
  privilege-dependent). Plausible and low-cost to add.
- **Action (advisory):** add a platform-gated test exercising this branch
  with a genuine OS error (permission-denied directory, or a self-referential
  symlink cycle on POSIX) and assert the wrapped cause remains inspectable
  via `errors.As`.

---

## Consensus / Majority findings

**None reached 3/4 or 4/4.** See Plurality section above for the strongest
multi-reviewer signals.

---

## Remediation plan (ordered by priority = confidence × severity)

| # | Finding | Confidence | Severity | Priority | Action class |
|---|---|---|---|---|---|
| 1 | F0 — live junction final-component containment bypass | HIGH (escalated) | CRITICAL | 12 | `manual` (requires design decision: explicit reparse-tag handling vs. GODEBUG pin) |
| 2 | F1 — no live-junction test coverage | MEDIUM | MAJOR | 6 | `gated_auto` — write tests once F0's fix lands; tests should fail before, pass after |
| 3 | F3 — live in-root junction ancestor false-rejection | MEDIUM | MAJOR | 6 | `gated_auto` — same fix as F0 likely resolves this; confirm with dedicated test |
| 4 | F2 — Windows case-fold fail-open (pre-existing, already disclosed) | MEDIUM | MAJOR | 6 | `advisory` — already accepted in register; no new action required |
| 5 | F4 — platform-dependent branch coverage gap | MEDIUM | MINOR | 4 | `gated_auto` — strengthen assertion to pin the actual branch taken |
| 6 | F5 — stripUNCPrefix share-less UNC acceptance | MEDIUM | MINOR | 4 | `gated_auto` — tighten `isWindowsAbsolutePath`'s UNC branch, add test cases |
| 7 | F6 — documented-unreachable branch fails open | LOW-MEDIUM | MAJOR (as framed) | 3-6 | `advisory` — consider flipping default to fail-closed |
| 8 | F7 — GODEBUG winsymlink dependency unpinned | LOW | MINOR | 2 | `advisory` — resolved as part of F0's fix |
| 9 | F8 — uncPrefix doc misattribution | LOW | MINOR | 2 | `advisory` |
| 10 | F9 — risk register contradictory STATUS pair | LOW | MINOR | 2 | `advisory` |
| 11 | F10 — pwsh hard-fail vs. skip inconsistency | LOW | MINOR | 2 | `advisory` |
| 12 | F11 — untested permission/cycle branch | LOW | MINOR | 2 | `advisory` |

**Housekeeping (not a finding, must still be done before merge):** delete
the neutralized scratch file
`internal/pathsafe/zzz_verify_junction_test.go` (`git clean -fd
internal/pathsafe` or a direct delete) — it is untracked, was created by
this review's verification tooling, and carries no functional content.

---

## Bug/issue queue entries (P0/P1)

```yaml
type: bug
title: "checkSymlinkEscape: live Windows directory junction at final path component bypasses containment"
description: >
  Root.Resolve accepts a candidate whose final path component is a live
  (non-dangling) directory junction pointing outside the workspace root.
  os.Stat on Windows follows name-surrogate reparse points (junctions) to
  the target and reports it as an ordinary directory; the subsequent
  filepath.EvalSymlinks call does not resolve IO_REPARSE_TAG_MOUNT_POINT
  reparse points (only IO_REPARSE_TAG_SYMLINK), so the unresolved,
  lexically-in-root path passes hasPathPrefix unchanged. Deterministic,
  reproducible on any Windows host, requires no elevated privilege (junction
  creation needs no SeCreateSymbolicLinkPrivilege).
file: "internal/pathsafe/pathsafe.go"
line: null
severity: "CRITICAL"
confidence: "HIGH"
fix: >
  Detect name-surrogate reparse points (junctions/mount points) at the final
  path component explicitly and resolve their true target before the
  containment check, or pin GODEBUG=winsymlink=0 module-wide and add a
  regression test proving consistent behavior under both GODEBUG settings.
  Add a test creating a live junction to an outside directory as the final
  candidate component and asserting rejection with symlinkEscapeMsg.
linked_review: "docs/closure/2026-09-07-011-s-012-f-pathsafe-containment-adversarial-review.md"
```

```yaml
type: bug
title: "checkSymlinkEscape: live junction as strict ancestor is unconditionally rejected regardless of target"
description: >
  Any live directory junction used as an intermediate ancestor with a
  non-existent leaf is rejected via os.Lstat's non-dir/non-symlink guard,
  even when the junction's target is inside the workspace root. The
  function's own doc comment claims such entries are "re-resolved" but
  filepath.EvalSymlinks is never reached for a junction ancestor. Fails
  closed (not a security hole) but breaks legitimate use of in-workspace
  junctions on Windows and the doc comment is inaccurate.
file: "internal/pathsafe/pathsafe.go"
line: null
severity: "MAJOR"
confidence: "MEDIUM"
fix: >
  Resolve the same fix as the final-component bypass above (explicit
  reparse-tag-aware resolution), and correct the doc comment. Add a test
  creating a live in-root junction as an intermediate ancestor with a
  missing leaf and asserting acceptance.
linked_review: "docs/closure/2026-09-07-011-s-012-f-pathsafe-containment-adversarial-review.md"
```

```yaml
type: test-gap
title: "No unit test exercises a live (non-dangling) junction in any position"
description: >
  internal/pathsafe/junction_windows_test.go contains exactly one test
  (dangling-target case). No test covers a live junction at the final
  component, as an intermediate ancestor pointing outside root, or as an
  intermediate ancestor pointing inside root — the three shapes needed to
  prove or disprove the two bugs above.
file: "internal/pathsafe/junction_windows_test.go"
line: null
severity: "MAJOR"
confidence: "MEDIUM"
fix: "Add the three test cases described in F1 above."
linked_review: "docs/closure/2026-09-07-011-s-012-f-pathsafe-containment-adversarial-review.md"
```

---

## Consensus verdict: **BLOCKED**

The five prior local reviews (Security, Correctness, Go, Scope Boundary,
Maintainability) correctly verified everything they checked, and their
conclusions about the redundant-readlink simplification, `os.IsNotExist`
normalization, and the byte-identical risk-register text are not disputed
here. But this adversarial pass found something none of them, nor three of
the four parallel reviewers in this pass, caught: **a deterministic,
unprivileged Windows containment bypass via a live directory junction at
the final path component (F0)**, independently confirmed against the pinned
Go 1.26.5 standard library source rather than taken on a single reviewer's
word. For a control whose own brief marks it NON-NEGOTIABLE (Brief §10),
and whose stated purpose is closing exactly this class of
symlink/junction containment bypass (ECE3DAB7), an unfixed live-junction
escape is disqualifying on its own merits, independent of the raw
reviewer-agreement count.

**Blocking items (must be resolved before merge):**
1. F0 — live junction final-component bypass (fix + regression test).
2. F1 — add live-junction test coverage (all three shapes).
3. F3 — live in-root junction ancestor false-rejection (very likely fixed
   by the same change as F0; confirm with a dedicated test either way).

**Non-blocking, should be acknowledged or fixed in the same or a fast-follow
change:** F4, F5, F6, F7, F9, F10 (documentation/coverage tightening,
already-low real-world exploitability).

**Non-blocking, no action required:** F2 (already disclosed and accepted in
the risk register with the correct fail-open framing), F8, F11 (advisory
only).

**Immediate housekeeping regardless of the above:** delete the neutralized
`internal/pathsafe/zzz_verify_junction_test.go` scratch file before any
further work on this branch.

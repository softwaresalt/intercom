---
title: "Adversarial Review — 013-S Pathsafe Reparse-Point Containment Hardening"
date: 2026-09-08
shipment: 013-S
feature: 014-F
mode: report-only
status: complete
---

# Adversarial Review — 013-S Pathsafe Reparse-Point Containment Hardening

**Scope**: `git diff main..HEAD -- internal/pathsafe/` on
`feat/013-s-pathsafe-reparse-point-containment-hardening`, reviewed against the
PASS-verdicted plan at
`docs/plans/2026-09-08-intercom-go-pathsafe-reparse-containment-plan.md`.

**Mode**: report-only, read-only. No files were edited as part of this review.

**Reviewer pool**: 3 independent parallel reviewers (no anchor route requested).

| Slot | Route | Model |
|---|---|---|
| Reviewer-A | Tier 1 (fast/cheap) | `gpt-5.4-mini` |
| Reviewer-B | Tier 2 (standard) | `claude-sonnet-5` |
| Reviewer-C | Tier 3 (frontier) | `claude-opus-5` |

All three reviewers were given identical scope, identical plan context, identical
instructions to *not* re-flag the plan's already-adjudicated §7/§8 disclosures
(Windows-CI-advisory, case-folding fail-open, F3 in-root-junction false-rejection,
GODEBUG empirical-proof harness), and identical instructions to return structured
JSON findings only. All three completed before aggregation began.

The assembling session independently re-traced every reviewer claim against the
current source (`pathsafe.go`, `reparse_windows.go`, `reparse_other.go`, `root.go`,
and the full test suite) before classifying it — this caught one reviewer
hallucination (below) and upgraded confidence on several unique findings that were
independently reproduced by direct code tracing.

**Baseline note**: a prior standard (single-model) review already found and fixed,
before this adversarial pass began: (1) a causal misattribution in root.go's
`700B41CE` entry, (2) a factually incorrect "not used by X" claim in the `uncPrefix`
doc comment, (3) an unchecked `syscall.CloseHandle` errcheck violation (now an
explicit discard). One of the three reviewers (Reviewer-C, F8 below) found evidence
that fix (2) may have swapped one inaccuracy for a different, unverified one — see
F8.

---

## Positive consensus (not a defect, worth stating explicitly)

All three reviewers, working independently and using different reasoning
approaches, concluded that **the core containment fix is sound**: `checkSymlinkEscape`
+ `canonicalizeReparse` correctly closes the C1/C2/C3 bypass shapes described in
plan §2, and none of the three reviewers identified a newly-introduced escape-
direction bypass in the primary threat model (a reparse point inside the workspace
redirecting outward). Every new issue found by any reviewer is either fail-closed
(over-rejection), documentation-only, or a defense-in-depth/regression-lock gap for
a *future* change — not a live containment failure in the code as shipped.

---

## Discredited finding (reviewer hallucination — verified false)

Reviewer-A (Tier 1) claimed:

> `TestResolveRejectsLiveJunctionAsIntermediateWithExistingLeaf` never creates
> `outsideRoot/existing.txt`, so it does not exercise C2.

**Verified false.** The test (`junction_windows_test.go`) does create the leaf:

```go
outsideRoot := t.TempDir()
existingLeaf := filepath.Join(outsideRoot, "existing.txt")
if err := os.WriteFile(existingLeaf, []byte("x"), 0o644); err != nil { ... }
```

This is dropped from the remediation queue entirely. It is retained here only as
evidence of why cross-model consensus and independent verification matter — a
single-model review acting on this claim would have wasted remediation effort on a
non-issue.

---

## HIGH confidence (consensus — all 3 reviewers)

None. No finding was independently raised by all three reviewers. This is a
believable outcome given the plan already passed two prior adversarial cycles
(rev 1 six-persona, rev 2 three-model) before this shipment was cut — the obvious
defects were already found and fixed upstream of this diff.

## MEDIUM confidence — Plurality (2 of 3 reviewers)

### P1. Silent test-skip defeats the shipment's own mandated evidence path
**Reviewers**: Reviewer-B (rated MINOR) + Reviewer-C (rated MAJOR) — most
conservative severity taken per Phase 3 rule 3.
**Severity**: MAJOR · **Files**: `internal/pathsafe/junction_windows_test.go`
(`createDirectoryJunction`), `internal/pathsafe/symlink_test.go`
(`requireSymlinkOrFailClosed`, for contrast)

Plan §7 makes a **local Windows** `go test ./internal/pathsafe/...` run the
**primary, mandatory** evidence for the C1/C2/C3 closure and the GO-14 flip,
because Windows CI is disclosed as advisory-only. `createDirectoryJunction`
does this:

```go
if _, err := exec.LookPath("pwsh"); err != nil {
    t.Skip("pwsh not available on PATH")
}
```

`pwsh` (PowerShell 7+) is **not** installed by default on Windows (only
`powershell.exe` 5.1 ships out of the box). Every C1/C2/C3 regression lock in
`junction_windows_test.go` and every helper test in `reparse_windows_test.go`
silently vanishes — reported as neither pass nor fail, just absent — on any
Windows evidence machine without `pwsh` on `PATH`. `go test ./internal/pathsafe/...`
would report a fully green result while never having executed the decisive
assertions this shipment's closure claim depends on, unless the operator
specifically inspects `-v`/skip output.

This directly contradicts the package's own established doctrine, encoded a few
files away in `requireSymlinkOrFailClosed`: on Windows, an unmet precondition
**must fail, not skip**, except for the single enumerated
`ERROR_PRIVILEGE_NOT_HELD` exemption. Junction creation requires no elevated
privilege (plan `014.001-T` AC 5 states this explicitly), so no enumerated
exemption applies to a missing-`pwsh` skip — by the package's own rule, this
should be a loud failure, not a silent skip.

**Fix**: Either (a) use `cmd /c mklink /J` as the junction-creation mechanism
(always present on Windows, no privilege, no temp-script/`pwsh` dependency), or
(b) keep `pwsh` as primary but fall back to `mklink /J`, or (c) at minimum mirror
`requireSymlinkOrFailClosed`'s pattern: `t.Fatalf` on Windows when `pwsh` is
absent, so the primary evidence run cannot silently omit the decisive tests.
**Action before relying on this shipment's §7 closure claim**: re-run the
Windows evidence with `-v` and manually confirm zero `SKIP` lines for the
junction/dangling-symlink tests.

### P2. Risk-register acceptance claim for in-root C2/C3 junctions is untested at the `Resolve()` level
**Reviewers**: Reviewer-A (imprecise framing, see note) + Reviewer-C (precise framing)
**Severity**: MINOR · **File**: `internal/pathsafe/root.go` (`700B41CE` entry)

`root.go`'s `700B41CE` entry claims `canonicalizeReparse` "now accepts the
previously falsely-rejected in-root intermediate/ancestor-junction shape" for
C2/C3. Verified by trace: this claim is **likely accurate** for the
existing-leaf/existing-dir-beyond-junction shapes specifically (an in-root
junction with an existing descendant does reach `canonicalizeReparse` and
would be accepted). However, **no test asserts this through `Resolve()`** —
`reparse_windows_test.go`'s `TestCanonicalizeReparseInRootJunction` only
exercises the helper function in isolation, not the containment verdict, and
every junction test in `junction_windows_test.go` uses an out-of-root target.
The claim is a documentation assertion with no regression lock.

Note: Reviewer-A's version of this finding framed it as a direct contradiction
with the F3 residual entry. That framing conflates two different shapes (C2/C3
"existing descendant beyond the junction" vs. F3's "junction itself as the
immediate ancestor of a non-existent leaf") and is not accurate as stated — the
two entries are not, in fact, contradictory. Reviewer-C's framing (untested
claim, not an inconsistent one) is the correct characterization and is what is
recorded here.

**Fix**: Add a `Resolve()`-level acceptance test for the in-root C2/C3 shapes
(e.g., `Resolve("in-root-junction/existing.txt")` and
`Resolve("in-root-junction/existing-dir/missing.txt")` must succeed), or
downgrade the risk-register wording from "verified-correct" to "traced, not
test-locked."

---

## LOW confidence — Unique findings (1 of 3 reviewers), preserved per protocol

Ordered by severity, then by independent-verification confidence.

### U1. Root/candidate canonicalization-namespace asymmetry (Reviewer-C)
**Severity**: MAJOR (fail-closed direction — availability break, not a bypass)
**Files**: `internal/pathsafe/root.go` (`NewRoot`), `internal/pathsafe/pathsafe.go`
(`checkSymlinkEscape`)
**Independently verified**: yes, by direct trace, with high confidence — this
finding should be weighted above a typical unique/LOW finding despite the 1-of-3
count.

`NewRoot` canonicalizes the workspace root via `filepath.EvalSymlinks` (which,
per this shipment's own root-cause analysis in plan §2, does **not** resolve
`IO_REPARSE_TAG_MOUNT_POINT` and does not re-express a volume mounted without a
drive letter). `checkSymlinkEscape` now canonicalizes the *candidate* side via
`canonicalizeReparse` (`GetFinalPathNameByHandleW`), which **does** resolve
mount points and **does** re-express such volumes as `\\?\Volume{GUID}\...` —
a form `stripUNCPrefix` deliberately refuses to strip (confirmed by
`root_windows_test.go`'s own `"volume guid"` test case, which pins exactly this
non-stripping behavior as intended).

Traced consequence: if the workspace root passed to `NewRoot` is **itself** a
live directory junction (a real, plausible deployment shape — redirected
project directories, cache-redirection junctions, corporate-policy-redirected
folders), `root.path` is stored as the junction's own lexical path (e.g.
`C:\workspace`), never its real target. Any subsequent `Resolve` call whose
ancestor-walk reaches the root itself now uses `os.Lstat` (changed by
`014.004-T` from `os.Stat`) — `Lstat` on the root reports the junction's own
`ModeIrregular`, which (with `finalProbe == false` at that point) triggers the
unconditional ancestor-rejection branch, or — for candidates that already exist
— `canonicalizeReparse` resolves the *candidate* to the real target while
`root.path` remains the *unresolved junction path*, so `hasPathPrefix` compares
two paths in different namespaces and **every** access under such a root fails,
either as `symlinkUnverifiableMsg` or (worse) a false-positive
`symlinkEscapeMsg` claiming a "GENUINE, VERIFIED escape" for a workspace that
never escaped.

This is a new consequence of `014.004-T`'s `Stat`→`Lstat` switch in the
ancestor loop combined with `NewRoot`'s own canonicalization step (`root.go`,
untouched by 013-S per plan §3's surface table) never having been reconciled
with the candidate-side canonicalization upgrade. Before 013-S, `os.Stat` on
the ancestor transparently followed the root's own junction and matched
`root.path` (which was equally unresolved on both sides at the time, since both
sides used `EvalSymlinks`-class resolution) — so this specific configuration
worked correctly pre-013-S and is broken (in the fail-closed direction) by this
shipment. No test in the suite constructs a workspace root that is itself a
junction or a GUID-only-addressable volume; `root_test.go`'s only symlink-root
test (`TestNewRootAcceptsSymlinkToDirectory`) uses an ordinary symlink, which
`EvalSymlinks` *does* resolve, so it does not exercise this gap.

**Fix**: Canonicalize the root through the same helper the candidate side uses
— replace `filepath.EvalSymlinks(abs)` in `NewRoot` with `canonicalizeReparse(abs)`
(then `stripUNCPrefix` as today) so both sides are guaranteed to land in one
namespace by construction. At minimum, add a root-register entry naming this
asymmetry and a regression test for a junctioned/GUID-volume workspace root
before relying on this shipment for such deployments.
**Recommended priority**: investigate before merge given severity, even though
only one reviewer flagged it — the underlying mechanism was independently
confirmed against the actual source and an existing test fixture
(`root_windows_test.go`'s `"volume guid"` case) that already anticipates half
of this exact divergence.

### U2. Missing regression lock: live out-of-root junction as intermediate with a *non-existent* leaf (Reviewer-C)
**Severity**: MAJOR (forward-looking defense-in-depth gap, not a live bug)
**File**: `internal/pathsafe/junction_windows_test.go`

`Resolve("junction-live/new-file.txt")`, where `junction-live` is a live
out-of-root junction and `new-file.txt` does not exist, is never tested.
Verified by trace: this exact shape is currently rejected — but **only**
because it hits the F3 blanket `ModeIrregular` ancestor guard (which rejects
*any* junction ancestor, in-root or out-of-root, indiscriminately), not because
`canonicalizeReparse` was called and proved the target outside root. `root.go`'s
own F3 entry explicitly names a future-shipment TRIGGER for relaxing that
guard to distinguish in-root from out-of-root junction ancestors. If a future
change relaxes F3 as documented but does not *also* make that specific ancestor
check reparse-aware before accepting, this exact shape's only remaining guard
disappears and no existing test would turn red — reintroducing a bypass of the
same shape this shipment was built to close, silently.

This also means the shape currently gets `symlinkUnverifiableMsg` even though
its target is, in principle, fully resolvable and provably outside root — the
"weaker" of the two 014.006-T messages is emitted for what would, on
verification, be a genuine escape.

**Fix**: Add `TestResolveRejectsLiveJunctionAsIntermediateWithNonExistentLeaf`
asserting rejection of `Resolve("junction-live/new-file.txt")` independent of
the F3 guard's current behavior (or add a comment/assertion that fails loudly
if this shape's rejection ever starts depending on `canonicalizeReparse`'s
verdict rather than the F3 guard, so that a future F3 relaxation is forced to
confront this test explicitly).

### U3. `uncPrefix` doc comment's specific stdlib-internals claim is unverified and may re-introduce the class of defect it was just fixed for (Reviewer-C)
**Severity**: MEDIUM (documentation-accuracy in a NON-NEGOTIABLE control's
risk register, no runtime effect)
**File**: `internal/pathsafe/root.go` (`uncPrefix` doc comment)

The current comment (edited by the prior remediation pass to fix the
"not used by X" claim) now asserts:

> "via Go's OWN standard-library `filepath.EvalSymlinks`, whose Windows-specific
> implementation internally calls `GetFinalPathNameByHandleW` to construct its
> return value."

This is a specific, confident claim about Go stdlib internals that was not
independently verified against the actual `path/filepath` source for the
toolchain in use, and it sits in tension with this shipment's own root-cause
premise (plan §2: `EvalSymlinks` does **not** resolve
`IO_REPARSE_TAG_MOUNT_POINT`) — if `EvalSymlinks` literally built its return
value from the same reparse-resolving Win32 call `canonicalizeReparse` uses,
the C1 bypass this shipment closes would not have existed in the first place.
The two facts are not necessarily contradictory (Go's implementation could use
`GetFinalPathNameByHandleW` for a final case/short-name normalization pass
without using it to drive the *symlink-walk* that decides what to follow), but
the comment does not make that distinction and the underlying claim was not
verified against source. Plan `014.008-T` AC 3 (F8) specifically required this
comment be accurate; this looks like an inaccuracy of a different shape may
have been introduced while fixing the first one.

**Fix**: Verify the specific stdlib-internals claim against the actual Go
version's `path/filepath` source before merge, or soften the comment to avoid
asserting an internal mechanism that isn't load-bearing for the doc comment's
actual purpose (distinguishing the two `\\?\`-emitting call sites) — e.g.,
"NewRoot's canonicalization may emit a `\\?\` prefix via `filepath.EvalSymlinks`
on Windows; `checkSymlinkEscape`'s canonicalization emits it via
`canonicalizeReparse`'s own, separate, direct `GetFinalPathNameByHandleW` call"
— without the "internally calls" mechanism claim.

### U4. No test combines "dangling" + "exact final component" for a junction specifically (Reviewer-B)
**Severity**: MINOR · **File**: `internal/pathsafe/junction_windows_test.go`

The suite has dangling-junction-as-intermediate
(`TestResolveRejectsDanglingIntermediateJunctionOutsideRoot`) and
live-junction-as-final-component (`TestResolveRejectsLiveJunctionAsFinalComponent`),
and dangling-symlink-as-final-component
(`TestResolveRejectsDanglingSymlinkAtFinalComponent`), but no
dangling-**junction**-as-the-exact-final-component test. `root.go`'s GO-14
entry claims the dangling-target rejection now applies "at ANY position,
including the terminal component" for reparse points generally; this specific
combination (junction, not symlink, exact final position, dangling) is never
directly exercised.

**Fix**: Add a Windows test creating a live junction as the literal final
`Resolve()` component, then removing its target to make it dangling, asserting
rejection with `symlinkUnverifiableMsg` — mirroring the existing dangling-
symlink test but for the junction/`canonicalizeReparse` code path.

### U5. Long-path (`>MAX_PATH`) handling asymmetry between `os.Lstat` and raw `syscall.CreateFile` (Reviewer-B)
**Severity**: MINOR (fail-closed direction) · **File**:
`internal/pathsafe/reparse_windows.go:44`

`canonicalizeReparse` calls `syscall.CreateFile` directly on the raw ancestor
path with no `\\?\` long-path prefixing, while `os.Lstat` (used earlier in the
same ancestor-climb loop) applies Go's own internal long-path handling. For a
legitimate, deeply nested in-root ancestor path exceeding `MAX_PATH` on a
system without process-wide long-path opt-in, `os.Lstat` could succeed while
`canonicalizeReparse` fails, spuriously rejecting a valid path. Fail-closed
(safe direction), but an undocumented robustness gap of the same character as
the already-disclosed F3 residual, just not itself adjudicated or named.

**Fix**: Either prefix non-UNC, non-already-extended paths passed to
`syscall.CreateFile` with `\\?\`, or explicitly document this as a known
long-path limitation alongside F3 in the risk register.

### U6. `canonicalizeReparse` (Windows) does not itself apply `stripUNCPrefix` (Reviewer-C)
**Severity**: MINOR (spec-compliance nitpick, no functional impact today —
the sole caller already strips) · **File**: `internal/pathsafe/reparse_windows.go`

Plan `014.003-T` AC 4 states "the returned path is normalized through the
existing `stripUNCPrefix`," which reads as a postcondition on the helper
itself. The implementation instead returns the raw `\\?\`-prefixed form and
relies on the caller (`checkSymlinkEscape`) to strip it — evidenced by
`reparse_windows_test.go` having to call `stripUNCPrefix(got)` in every
assertion. The POSIX sibling (`reparse_other.go`) never needs stripping in the
first place, so the two build-tagged implementations have divergent,
undocumented postconditions; a future second caller of `canonicalizeReparse`
could easily get an un-normalized path.

**Fix**: Apply `stripUNCPrefix` inside `canonicalizeReparse` (Windows) so both
implementations share one postcondition, and drop the now-redundant strip in
`checkSymlinkEscape`; or explicitly document "returns an extended-length
`\\?\` path; callers must strip" on the helper.

### U7. `LazyProc.Call` panics rather than returning an error if the Win32 export can't be resolved (Reviewer-C)
**Severity**: MINOR/advisory (theoretical — `GetFinalPathNameByHandleW` has
existed since Windows Vista/Server 2008, well below this package's minimum
supported Windows version) · **File**: `internal/pathsafe/reparse_windows.go`

`procGetFinalPathNameByHandleW.Call(...)` internally calls `Addr()`, which
panics if `Find()` fails to resolve the DLL/proc. A panic escapes the
`apperr.KindPathViolation` contract and is not `errors.Is`/`errors.As`
inspectable — in the single enforcement point of a NON-NEGOTIABLE control.
In practice this is unreachable on any Windows version this package targets,
consistent with several other explicitly-accepted "documented unreachable"
defensive branches already present elsewhere in this package (e.g.
`checkSymlinkEscape`'s filesystem-root branch, `NewRoot`'s post-`EvalSymlinks`
`os.Stat` branch).

**Fix (optional, low priority)**: Resolve the proc once via `Find()`/`Load()`
at package-init time and surface any resolution error as a wrapped
`symlinkUnverifiableMsg` rejection from `canonicalizeReparse`, rather than
leaving it to panic — for consistency with the package's stated "never panic
in the containment path" posture, even though the trigger condition is not
believed reachable in practice.

---

## Remediation plan (ordered by priority = confidence_weight × severity_weight)

| # | Finding | Confidence | Severity | Score | Action class |
|---|---|---|---|---|---|
| 1 | P1 — silent test-skip defeats mandated evidence path | MEDIUM (2) | MAJOR (3) | 6 | `manual` — fix harness or manually verify `-v` output before relying on §7 closure claim |
| 2 | U1 — root/candidate canonicalization namespace asymmetry | LOW (1)* | MAJOR (3) | 3* | `manual` — investigate before merge; independently verified, treat as elevated priority despite LOW confidence label |
| 3 | U2 — missing regression lock for live-junction-intermediate + missing-leaf | LOW (1) | MAJOR (3) | 3 | `manual` — add test now to pre-empt a future F3-relaxation regression |
| 4 | P2 — in-root C2/C3 acceptance claim untested at `Resolve()` level | MEDIUM (2) | MINOR (2) | 4 | `advisory` — add test or soften register wording |
| 5 | U3 — unverified `uncPrefix` stdlib-internals claim | LOW (1) | MEDIUM/MINOR (2) | 2 | `advisory` — verify against Go source or soften wording |
| 6 | U4 — no dangling-junction-as-final-component test | LOW (1) | MINOR (2) | 2 | `advisory` |
| 7 | U5 — long-path handling asymmetry (Lstat vs. raw CreateFile) | LOW (1) | MINOR (2) | 2 | `advisory` |
| 8 | U6 — `canonicalizeReparse` doesn't self-apply `stripUNCPrefix` | LOW (1) | MINOR (2) | 2 | `advisory` |
| 9 | U7 — `LazyProc.Call` panic path (theoretical) | LOW (1) | MINOR (2) | 2 | `advisory` |

\* U1 and U2 are formally LOW-confidence (unique, 1-of-3-reviewer) findings per
strict protocol scoring, but both were independently re-traced against the
actual source by the assembling session and confirmed plausible/likely-correct
with high confidence. They are listed ahead of their strict numeric score would
place them, and flagged explicitly for human attention — this is exactly the
scenario the protocol's LOW-confidence tier exists to preserve rather than
discard.

Sorted-by-score, ties broken by file path: table above already reflects this
ordering with the two elevated-attention items called out explicitly rather
than silently reordered past their computed score.

---

## Backlog work items (P0/P1-equivalent findings)

```yaml
type: bug
title: "checkSymlinkEscape: root/candidate canonicalization namespace asymmetry breaks junctioned/GUID-volume workspace roots"
description: "NewRoot canonicalizes root.path via filepath.EvalSymlinks (junction- and GUID-volume-unaware); checkSymlinkEscape now canonicalizes candidates via canonicalizeReparse (junction- and GUID-volume-aware). For a workspace root that is itself a live junction or addressable only via a \\?\\Volume{GUID}\\ path, every Resolve() call fails -- either as a false symlinkEscapeMsg claim or symlinkUnverifiableMsg -- because root.path and the canonicalized candidate land in different path namespaces. Fail-closed, but a total functional break for a plausible deployment shape, newly introduced by 014.004-T's Stat->Lstat switch in the ancestor loop."
file: "internal/pathsafe/root.go / internal/pathsafe/pathsafe.go"
line: null
severity: "MAJOR"
confidence: "LOW (unique, independently verified)"
fix: "Canonicalize the workspace root through canonicalizeReparse (not filepath.EvalSymlinks) in NewRoot so both sides of the containment comparison are guaranteed to share one namespace by construction; add a regression test for a junctioned/GUID-volume workspace root."
linked_review: "docs/closure/2026-09-08-013-s-014-f-pathsafe-reparse-containment-adversarial-review.md"
```

```yaml
type: bug
title: "junction_windows_test.go: missing pwsh silently skips the shipment's primary Windows evidence tests instead of failing"
description: "createDirectoryJunction calls t.Skip when pwsh is not on PATH. pwsh is not installed by default on Windows. Plan 013-S section 7 makes a local Windows go test run the mandatory primary evidence for the C1/C2/C3 closure claim, since Windows CI is advisory-only. A missing-pwsh evidence run reports fully green while never executing the decisive junction tests. This contradicts the package's own 011.010-T fail-closed-not-skip doctrine, which junction creation (requiring no elevated privilege) does not qualify for an exemption from."
file: "internal/pathsafe/junction_windows_test.go"
line: 20
severity: "MAJOR"
confidence: "MEDIUM (plurality, 2 of 3 reviewers)"
fix: "Use cmd /c mklink /J as the junction-creation mechanism (no pwsh dependency, no privilege required), or t.Fatalf (not t.Skip) on Windows when pwsh is unavailable, matching requireSymlinkOrFailClosed's established pattern."
linked_review: "docs/closure/2026-09-08-013-s-014-f-pathsafe-reparse-containment-adversarial-review.md"
```

```yaml
type: task
title: "Add regression lock for live out-of-root junction as intermediate with non-existent leaf, independent of the F3 ancestor guard"
description: "Resolve(\"junction-live/new-file.txt\") is currently rejected only because the F3 blanket ModeIrregular ancestor guard fires, not because canonicalizeReparse verified the target outside root. root.go's F3 entry names a future-shipment TRIGGER for relaxing that guard for in-root junctions; if that relaxation does not also make this specific ancestor check reparse-aware, this exact bypass shape could reopen silently with no existing test turning red."
file: "internal/pathsafe/junction_windows_test.go"
line: null
severity: "MAJOR"
confidence: "LOW (unique, independently verified as a real forward-looking gap)"
fix: "Add TestResolveRejectsLiveJunctionAsIntermediateWithNonExistentLeaf asserting rejection of this shape, structured so a future F3 relaxation is forced to confront it explicitly."
linked_review: "docs/closure/2026-09-08-013-s-014-f-pathsafe-reparse-containment-adversarial-review.md"
```

---

## Post-remediation re-review

```yaml
post_remediation:
  cycles_run: 0
  cap_reached: false
  residual_findings: 0
  status: "skipped"
```

Skipped: this is a report-only review per the operator's invocation mode (no
`safe_auto` fixes were applied, so Phase 7 has no modified files to re-dispatch
over). All findings above are routed to `manual`/`advisory` action classes and
require explicit operator or implementer action, not automated remediation.

---

## Overall readiness verdict: **READY_WITH_FOLLOWUPS**

**Rationale**:

* No reviewer, and no independent re-trace by the assembling session, found a
  newly-introduced containment bypass (an accept where the plan requires a
  reject) in the core fix. The C1/C2/C3 closure and the GO-14 flip appear
  correctly implemented and monotonically fail-closed, matching the plan's
  stated design intent.
* The strongest findings (U1, P1, U2) are all in the fail-closed direction —
  over-rejection or evidence-integrity gaps — not new bypasses. This matters:
  the shipment's NON-NEGOTIABLE property (no escape) is not contradicted by
  anything found here.
* However, two items should be resolved or explicitly acknowledged before this
  shipment's evidence-based closure claim (plan §7) can be trusted at face
  value:
  1. **P1** (silent test-skip) means the mandated local-Windows evidence run
     could be green without ever having executed the decisive tests — verify
     `-v` output shows zero skips, or fix the harness, before signing off on
     §7's closure requirement.
  2. **U1** (canonicalization namespace asymmetry) is a plausible, independently
     -verified functional break for junctioned/GUID-volume workspace roots that
     was not covered by the plan's own risk register or test matrix — worth a
     same-shipment or immediate-follow-up fix given its severity, even though
     only one reviewer's pass surfaced it.
* All remaining findings (P2, U3–U7) are legitimate but lower-severity
  documentation-accuracy and defense-in-depth test-coverage observations that
  can be tracked as ordinary follow-up backlog items without blocking merge.

**Recommendation**: proceed with merge on the core fix's own security merits,
but do not close 013-S's evidence claim (plan §7) until P1 is checked, and open
a same-priority follow-up for U1 given its potential to break a real (if
narrower) deployment configuration in the fail-closed direction.

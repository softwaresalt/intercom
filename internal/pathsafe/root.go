// Package pathsafe implements the workspace path-containment security
// control ported from the agent-intercom behavioral oracle
// (softwaresalt/agent-intercom @ 41df772, src/diff/path_safety.rs). Brief
// section 10 marks workspace isolation NON-NEGOTIABLE.
//
// Known limitations (both oracle-parity, not silently inherited):
//   - TOCTOU: there is a window between path validation (Resolve) and actual
//     filesystem use during which the filesystem may change.
//   - EvalSymlinks does not resolve hardlinks, so a pre-existing in-workspace
//     hardlink to an external file on the same volume passes validation
//     (finding SEC-5).
//
// Consolidated risk register (011.004-T, partially resolves BF5DE670). Each
// entry names its current mitigation status and the concrete condition
// that would force mitigation:
//
//   - GO-14 (write-through-dangling-symlink): RESOLVED / FLIPPED by 013-S
//     (014.002-T + 014.004-T), superseding the "accepted, oracle-parity"
//     status this entry previously recorded. Previously, checkSymlinkEscape
//     accepted a dangling symlink at the exact FINAL path component
//     unconditionally (012.003-T had already narrowed the same primitive by
//     rejecting a strict-ancestor dangling link or junction, but
//     Resolve("link") itself, with one fewer path component, still
//     accepted). 014.002-T inverts the regression lock
//     (TestResolveRejectsDanglingSymlinkAtFinalComponent now asserts
//     rejection where the prior test, TestResolveAllowsDanglingSymlinkAt-
//     FinalComponent, asserted acceptance), and 014.004-T's canonicalization
//     rewrite makes the final component pass through canonicalizeReparse
//     exactly like every other resolved entry -- so a dangling target at
//     ANY position, including the terminal component, is now rejected
//     (symlinkUnverifiableMsg, 014.006-T) rather than silently accepted.
//     REPLACEMENT BOUNDARY: the prior acceptance boundary (final component
//     only) is withdrawn outright, not narrowed further -- there is no
//     remaining lexical-only acceptance case for a dangling target. This is
//     a deliberate DIVERGENCE from the behavioral oracle's lexical-only
//     final-component acceptance, justified by D8 (agent-intercom's Rust
//     repository is historical reference only for this package, not a live
//     behavioral oracle this package is obligated to replicate divergence-
//     for-divergence going forward). TRIGGER: none remaining -- finding
//     closed. Originating flip procedure: stash F133AB7E.
//   - Resolve -> use TOCTOU window (see "Known limitations" above): there
//     is no atomic validate-then-open primitive in this package. STATUS:
//     accepted, oracle-parity. TRIGGER: forced the moment a real write path
//     exists (unchanged by 013-S; 013-S's anti-goal excludes this from
//     scope — see below).
//   - SEC-5 (EvalSymlinks ignores hardlinks, see "Known limitations"
//     above). STATUS: accepted, oracle-parity. TRIGGER: same as the
//     Resolve -> use TOCTOU entry above (unchanged by 013-S).
//   - 5FE4A7BE-a (012.007-T; case-folding, darwin under-fold): darwin
//     currently under-folds because pathEqual/pathHasPrefix fold on
//     Windows only, while darwin is a shipped target and is
//     case-insensitive by default on APFS/HFS+. STATUS: accepted,
//     fail-closed, plausible/unconfirmed. TRIGGER: reproduce a legitimate
//     in-root darwin path rejected solely because the volume is
//     case-insensitive and the only difference is casing. Extending the
//     fold to darwin was rejected because a case-sensitive APFS volume
//     would turn that rejection into an acceptance, i.e. fail-open (see
//     5FE4A7BE-b). NOT touched by 013-S (deliberately out of scope, plan
//     §8 residual "case-folding fail-open" — declared out of scope
//     because fixing it is not a member of any of 013-S's selected stash
//     entries and 700B41CE's closure below explicitly does not cover it).
//   - 5FE4A7BE-b (012.007-T; case-folding, Windows/WSL over-fold — split
//     out of the single 5FE4A7BE entry by 014.008-T/F9, which previously
//     paired this with 5FE4A7BE-a under one contradictory STATUS/TRIGGER
//     pair): the already-enabled Windows fold (pathEqual/pathHasPrefix
//     strings.EqualFold) is itself not a safe baseline: NTFS supports
//     per-directory case sensitivity, and WSL enables it on the
//     directories it creates, so a symlink resolving to a case-variant
//     sibling can be folded into acceptance. STATUS: accepted, fail-open
//     risk. TRIGGER: a workspace root on a case-sensitivity-enabled NTFS
//     or WSL-created tree. NOT touched by 013-S (same plan §8 residual as
//     5FE4A7BE-a; the two entries name opposite-direction risks on
//     opposite platforms and must not be conflated into one).
//   - 700B41CE: CLOSED for the reparse-point mechanism by 013-S
//     (commits 0c51669 "feat(014.004-T): GREEN wire canonicalization into
//     the containment check" and 3e1cb58 "fix(014.005-T): flip the
//     fail-open terminal branch to fail-closed"). Previously: on Windows, a
//     LIVE (non-dangling) directory junction used as the FINAL path
//     component was accepted by checkSymlinkEscape regardless of its
//     target, because os.Stat transparently followed
//     IO_REPARSE_TAG_MOUNT_POINT and filepath.EvalSymlinks never resolved
//     it as the exact terminal argument -- containment was never actually
//     checked against the junction's real target (case C1, 014.001-T).
//     canonicalizeReparse (014.003-T, GetFinalPathNameByHandleW semantics)
//     now resolves IO_REPARSE_TAG_MOUNT_POINT at any path position,
//     including the terminal component, closing C1.
//     Cases C2 (junction as an intermediate component with an existing
//     leaf) and C3 (junction ancestor with an existing directory beyond
//     it) are named here as CLOSED per this entry's scope, but with an
//     important accuracy caveat established empirically during
//     014.001-T (Go 1.26.5, GODEBUG=winsymlink=1 default on this
//     toolchain): C2 and C3 were ALREADY correctly rejected by the
//     PRE-existing implementation before 013-S touched this file --
//     filepath.EvalSymlinks, called on an ancestor path with additional
//     components following the junction, already transparently resolved
//     it during ordinary OS path traversal. Only C1 (the junction as the
//     exact terminal argument passed to EvalSymlinks/Lstat, gated on a
//     ModeSymlink check that a junction's ModeIrregular never satisfies)
//     was a live, exploitable bypass on this toolchain. C2/C3's tests
//     (junction_windows_test.go) remain in the suite as non-regression
//     locks proving canonicalizeReparse continues to handle them
//     correctly, not as evidence that 013-S fixed three separate bugs.
//     This closure explicitly does NOT cover: case-folding (5FE4A7BE-a /
//     5FE4A7BE-b, untouched), hardlinks (SEC-5, untouched), or the
//     Resolve -> use TOCTOU window (untouched) -- those remain separately
//     tracked, unresolved risks. STATUS: closed (reparse-point mechanism
//     only). TRIGGER: none remaining for this specific mechanism; the
//     three risks explicitly excluded above retain their own independent
//     triggers, listed at their own entries.
//     See docs/closure/2026-09-07-011-s-012-f-pathsafe-containment-adversarial-review.md
//     (finding F0) for the pre-013-S trace, and
//     docs/plans/2026-09-08-intercom-go-pathsafe-reparse-containment-plan.md
//     for the full 013-S remediation record.
//   - F3 (in-root junction ancestor false-rejection; new entry, 014.008-T
//     AC 6, plan §8 residual): checkSymlinkEscape's strict-ancestor
//     rejection for an entry that is neither a directory nor a symlink
//     (a live directory junction reports fs.ModeIrregular, matching
//     neither) unconditionally rejects a junction used as a strict
//     ancestor, even when its real target is INSIDE the workspace root --
//     a false rejection, not a containment failure. Deliberately NOT
//     relaxed by 013-S: rev 1 of the 013-S plan attempted exactly this
//     relaxation, and rev 2's independent adversarial review (finding A2)
//     identified that relaxing this branch while filepath.EvalSymlinks
//     still cannot resolve mount points as a non-terminal probe would
//     INTRODUCE a new intermediate-junction bypass -- the wrong risk
//     trade inside a shipment whose purpose is closing a bypass, not
//     opening one. STATUS: accepted, fail-closed usability gap
//     (deliberately not relaxed). TRIGGER: a future shipment that
//     explicitly re-derives a reparse-aware ancestor probe (not merely a
//     terminal-canonicalization step like canonicalizeReparse) capable of
//     distinguishing an in-root live junction ancestor from an
//     out-of-root one without reintroducing A2's bypass.
//   - GODEBUG-empirical (both-settings empirical proof; new entry,
//     014.008-T AC 6, plan §8 residual): 014.004-T AC 5 makes the
//     reparse-point fix GODEBUG-independent BY CONSTRUCTION --
//     canonicalizeReparse calls GetFinalPathNameByHandleW directly rather
//     than relying on os.Stat/os.Lstat/filepath.EvalSymlinks's
//     GODEBUG=winsymlink-gated reparse-resolution semantics, so the
//     containment verdict is identical under GODEBUG=winsymlink=0 and =1.
//     A subprocess re-exec harness that empirically PROVES this by running
//     the suite under both settings is deliberately not built. STATUS:
//     accepted, YAGNI (scope-audit P3, plan §8) -- guards a non-default
//     toolchain configuration reachable only by deliberate override (this
//     toolchain, Go 1.26.5, pins winsymlink=1 by default). TRIGGER: a
//     future shipment that changes canonicalizeReparse to depend, even
//     partially, on os.Stat/os.Lstat/filepath.EvalSymlinks reparse
//     semantics rather than a direct Win32 call, at which point the
//     by-construction independence claim above must be re-verified or the
//     empirical harness must be built.
//
// RETIREMENT PROCEDURE when a real write path arrives (C4-C6): each finding
// above must be re-evaluated against the concrete write call site before
// that code merges; a mitigation (or an explicit, re-justified acceptance)
// must land in the SAME change that introduces the write path. This
// register, scripts/check-write-path-precondition.sh, and
// internal/config/validate.go rule 7's Constitution Check exception all
// expire together at that one trigger.
//
// ANTI-GOAL (011.004-T): no TOCTOU/hardlink mitigation mechanism is added
// by this register's creation; it records status quo risk, it does not
// change it.
package pathsafe

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/softwaresalt/intercom-go/internal/apperr"
)

// uncPrefix is the volume-path prefix Go's filepath.EvalSymlinks emits on
// Windows. This is Go's OWN standard-library internal implementation
// detail (package path/filepath's Windows-specific EvalSymlinks calls
// GetFinalPathNameByHandleW to resolve the input path and construct its
// return value, which is where this prefix in EvalSymlinks' OUTPUT comes
// from) -- it is used exclusively by NewRoot's one-time root
// canonicalization above, NOT by checkSymlinkEscape.
//
// F8 (014.008-T): this is a DISTINCT attribution from
// canonicalizeReparse's (reparse_windows.go, 014.003-T) own, separate,
// DIRECT GetFinalPathNameByHandleW call, which this package's own code
// makes explicitly (via syscall.NewLazyDLL, not through
// filepath.EvalSymlinks) as part of checkSymlinkEscape's per-Resolve-call
// containment check. The two call sites both ultimately reach the same
// underlying Win32 API, but for different callers, at different times,
// for different purposes -- do not conflate them.
const uncPrefix = `\\?\`

// Root is a canonicalized workspace root. Construct with NewRoot; the zero
// value is not valid.
type Root struct {
	path string
}

// Path returns the canonicalized, absolute root path.
func (r Root) Path() string {
	return r.path
}

// NewRoot canonicalizes dir exactly once: filepath.Abs, then
// filepath.EvalSymlinks, then stripUNCPrefix. Canonicalizing once amortizes
// what would otherwise be an EvalSymlinks syscall walk on every Resolve
// call — P10's diff-apply may validate 10-100 paths against one root
// (finding GO-4).
func NewRoot(dir string) (Root, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return Root{}, wrapRootInvalid(err)
	}

	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return Root{}, wrapRootInvalid(err)
	}

	canonical := stripUNCPrefix(resolved)
	// DOCUMENTED-UNREACHABLE, coverage-excluded (011.006-T item (b),
	// resolves 8472E0A1 item (b)): this os.Stat call cannot observe a
	// failure in practice. filepath.EvalSymlinks above already performs an
	// os.Lstat-based syscall walk over every path component (including the
	// final one) to resolve symlinks, and it already returned successfully
	// by this point — so a subsequent os.Stat on that same, just-resolved
	// path failing would require the filesystem to change between the two
	// calls (a TOCTOU race), not a normal input-driven code path. Requiring
	// a "fails before, passes after" test for this branch is unsatisfiable
	// without an injectable stat seam, which would itself be a production
	// behavior change inside a tests-only unit (Width Isolation). The
	// error return remains defensive, not dead, code.
	info, err := os.Stat(canonical)
	if err != nil {
		return Root{}, wrapRootInvalid(err)
	}
	if !info.IsDir() {
		return Root{}, apperr.Newf(apperr.KindPathViolation, "workspace root invalid: not a directory: %s", canonical)
	}

	return Root{path: canonical}, nil
}

// wrapRootInvalid wraps err as a KindPathViolation "workspace root invalid"
// error. Collapses three previously-identical apperr.Wrapf call sites in
// NewRoot into one (011.007-T, characterization-first refactor -- resolves
// 8472E0A1 item (a)). Zero verdict change: error Kind and
// errors.Is/errors.As discriminability are identical to the pre-refactor
// call sites, and every existing pathsafe test passes unmodified.
func wrapRootInvalid(err error) error {
	return apperr.Wrapf(apperr.KindPathViolation, err, "workspace root invalid: %s", err.Error())
}

// stripUNCPrefix removes a leading \\?\ prefix, which Go's EvalSymlinks
// emits on Windows. Extended-UNC paths are re-formed as ordinary UNC paths
// (\\server\share\...), and any stripped result that is not a Windows-
// absolute path is rejected by keeping the original extended-path input.
//
// Provenance note: this mirrors src/config.rs:24 strip_unc_prefix, which the
// oracle applies to default_workspace_root at src/config.rs:546 — it is NOT
// called from src/diff/path_safety.rs, which relies on Rust's
// component-aware Path::starts_with instead. The stripping is therefore a
// Go-specific adaptation required because Go's containment comparison is
// string-based, not an oracle-faithful port of path_safety.rs (findings
// SEC-4 / ARCH-4).
func stripUNCPrefix(p string) string {
	if !strings.HasPrefix(p, uncPrefix) {
		return p
	}

	remainder := p[len(uncPrefix):]
	if len(remainder) >= len("UNC\\") && strings.EqualFold(remainder[:len("UNC\\")], "UNC\\") {
		remainder = `\\` + remainder[len("UNC\\"):]
	}
	if !isWindowsAbsolutePath(remainder) {
		return p
	}
	return remainder
}

func isWindowsAbsolutePath(p string) bool {
	if len(p) >= 2 && p[0] == '\\' && p[1] == '\\' {
		return hasUNCHostAndShare(p[2:])
	}
	if len(p) < 3 {
		return false
	}
	return ((p[0] >= 'A' && p[0] <= 'Z') || (p[0] >= 'a' && p[0] <= 'z')) && p[1] == ':' && p[2] == '\\'
}

// hasUNCHostAndShare reports whether rest (the portion following a leading
// \\) contains both a non-empty host segment and a non-empty share segment
// -- the minimum structure required for a well-formed, addressable UNC path
// (\\host\share[\...]). A host-only remainder, or one with an empty share
// segment (e.g. "server" or "server\"), is not a complete UNC path and must
// not be accepted as Windows-absolute: stripUNCPrefix's fail-closed
// postcondition requires rejecting (not merely lexically prefix-matching)
// a malformed re-formed UNC result.
func hasUNCHostAndShare(rest string) bool {
	sep := strings.IndexByte(rest, '\\')
	if sep <= 0 {
		return false
	}
	afterHost := rest[sep+1:]
	share := afterHost
	if shareEnd := strings.IndexByte(afterHost, '\\'); shareEnd >= 0 {
		share = afterHost[:shareEnd]
	}
	return share != ""
}

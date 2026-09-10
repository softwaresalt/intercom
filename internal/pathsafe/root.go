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
//
//   - Resolve -> use TOCTOU window (see "Known limitations" above): there
//     is no atomic validate-then-open primitive in this package. STATUS:
//     accepted, oracle-parity. TRIGGER: forced the moment a real write path
//     exists (unchanged by 013-S; 013-S's anti-goal excludes this from
//     scope — see below).
//
//   - SEC-5 (EvalSymlinks ignores hardlinks, see "Known limitations"
//     above). STATUS: accepted, oracle-parity. TRIGGER: same as the
//     Resolve -> use TOCTOU entry above (unchanged by 013-S).
//
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
//
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
//
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
//     toolchain) and corrected during review of 014.008-T's first draft
//     of this entry: C2 and C3 were ALREADY rejected by the PRE-existing
//     implementation before 013-S touched this file, but NOT because
//     filepath.EvalSymlinks "transparently resolved" the junction's real
//     target and compared it against root -- it did not. EvalSymlinks's
//     Windows walker Lstats each intermediate component, and a directory
//     junction reports fs.ModeIrregular (neither ModeDir nor
//     ModeSymlink); when further path components remain, the walker
//     unconditionally errors out (an ENOTDIR-class OS error) the moment
//     it hits that irregular mode, REGARDLESS of whether the junction's
//     real target is inside or outside the workspace root. This produced
//     a coincidentally-correct rejection for the malicious (target
//     outside root) C2/C3 shapes, but the exact same code path ALSO
//     falsely rejected the benign shape (an intermediate/ancestor
//     junction whose real target is INSIDE root) with an identical
//     error -- containment was never actually verified for C2/C3 either;
//     the pre-existing behavior was an unrelated path-walk error, not a
//     resolved-and-compared verdict. canonicalizeReparse's terminal-
//     handle resolution (014.003-T/014.004-T) incidentally corrects this
//     side effect for C2/C3 as well: it now accepts the previously
//     falsely-rejected in-root intermediate/ancestor-junction shape while
//     still correctly rejecting the out-of-root one -- an uncredited but
//     verified-correct side benefit of the C1 fix, not a change this
//     entry's closure claim depends on. Only C1 (the junction as the
//     exact terminal argument passed to EvalSymlinks/Lstat, gated on a
//     ModeSymlink check that a junction's ModeIrregular never satisfies)
//     was a live, exploitable CONTAINMENT bypass on this toolchain; C2/C3
//     were a false-rejection usability defect, never a bypass, both
//     before and incidentally improved after 013-S. C2/C3's tests
//     (junction_windows_test.go) remain in the suite as non-regression
//     locks proving canonicalizeReparse continues to handle them
//     correctly (now WITHOUT the prior false-rejection side effect for
//     the in-root shape), not as evidence that 013-S fixed three
//     separate containment bugs.
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
//
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
//
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
//   - Root/candidate canonicalization asymmetry (013-S regression;
//     CLOSED by 015-S/016.007-T, C2; new entry, 016.008-T AC1): 013-S
//     (014.003-T/014.004-T) routed checkSymlinkEscape's per-Resolve-call
//     candidate canonicalization through canonicalizeReparse
//     (GetFinalPathNameByHandleW semantics, resolving reparse points
//     including directory junctions at any path position), but left
//     NewRoot's one-time ROOT canonicalization on filepath.EvalSymlinks,
//     which does not resolve directory junctions at all (its os.Lstat
//     walk reports a junction as fs.ModeIrregular and either errors out
//     on a non-terminal junction or passes a terminal one through
//     unresolved). A junction-rooted workspace (NewRoot invoked directly
//     on a directory-junction path) therefore left root.path UNRESOLVED
//     while every candidate passed through canonicalizeReparse's real-
//     target resolution -- a root/candidate namespace mismatch that
//     hasPathPrefix's plain string comparison cannot detect, causing
//     checkSymlinkEscape to reject legitimate in-root candidates whose
//     resolved (real-target) form no longer had the raw junction path as
//     a string prefix (found via 016.005-T's RED regression lock,
//     TestNewRootOnJunctionRootedWorkspaceResolvesExistingDescendant /
//     ...ResolvesNonExistentLeaf). This was a FALSE-REJECTION usability
//     defect for the benign in-root case, not a containment bypass (an
//     out-of-root candidate under a junction-rooted workspace was still
//     rejected, just via a different -- also correct -- code path).
//     016.007-T (C2) closes it by routing NewRoot through the SAME
//     canonicalizeReparse function checkSymlinkEscape already used,
//     restoring root/candidate symmetry; 016.004-T's long-path prefixing
//     landed in the same change per the B3->C2 hard-prerequisite note in
//     the 015-S plan (docs/plans/2026-09-10-intercom-go-pathsafe-
//     windows-canonicalization-symmetry-plan.md §7), since routing a raw
//     syscall.CreateFile through NewRoot without prefix handling would
//     have regressed long workspace roots that EvalSymlinks previously
//     tolerated. STATUS: closed. TRIGGER: none remaining for this
//     specific asymmetry; a related, EXPLICITLY NOT closed residual is
//     recorded separately below (long-path environment-sensitivity).
//
//   - Long-path (>MAX_PATH) prefixing environment-sensitivity (016.004-T,
//     U5; new entry, 016.008-T AC1, plan §9 H3): measured empirically on
//     this development host (Windows build 10.0.26200, registry
//     LongPathsEnabled=0) that a raw syscall.CreateFile call WITHOUT any
//     `\\?\` prefix does not reproduce the classic MAX_PATH rejection
//     even for absolute paths well over 260 characters -- the failure
//     addLongPathPrefix exists to prevent could not be forced to
//     reproduce as a live RED test on this environment, an outcome the
//     015-S plan explicitly anticipated (§9 H3) and prescribed landing
//     the mitigation together with C2 plus a covering test rather than
//     deferring both. addLongPathPrefix's deterministic, OS-independent
//     predicate tests (TestAddLongPathPrefixVerdicts,
//     TestAddLongPathPrefixRoundTripsWithStripUNCPrefix) serve as that
//     covering test: they lock the prefixing DECISION (threshold,
//     already-prefixed/device-path/UNC/relative-path exclusions)
//     independent of whether any given host's CreateFile call actually
//     rejects an unprefixed long path today. STATUS: accepted, mitigation
//     landed proactively ahead of confirmed local reproduction. TRIGGER:
//     a host or Windows configuration (e.g. LongPathsEnabled=1, or a
//     future toolchain/API change) that DOES reproduce the MAX_PATH
//     rejection without addLongPathPrefix, at which point an end-to-end
//     regression test should be added on that environment; none is
//     currently constructible on this development host.
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
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/softwaresalt/intercom-go/internal/apperr"
)

// uncPrefix is the extended-length-path volume prefix Windows APIs use
// for the `\\?\`-style path convention, stripped by stripUNCPrefix
// (below) before a resolved path is compared against root or returned
// to a caller.
//
// F8 (014.008-T, corrected TWICE during review: an inaccurate first
// draft claimed uncPrefix was NOT used by checkSymlinkEscape -- it is;
// a second draft then claimed filepath.EvalSymlinks "internally calls
// GetFinalPathNameByHandleW to construct its return value" as a
// general mechanism -- verified FALSE against this toolchain's actual
// path/filepath source, corrected below; SUPERSEDED IN PART by
// 016.007-T/C2, see the trailing note below the two-site breakdown):
//
// stripUNCPrefix, the sole consumer of this constant, is called from
// BOTH of this package's two `\\?\`-prefix-emitting call sites, which is
// exactly why they share one prefix constant rather than each needing
// their own. AS OF 016.007-T (015-S, C2), both sites reach that shared
// prefix via the SAME mechanism (canonicalizeReparse); PRIOR to that
// change they were independent, sometimes-divergent code paths, verified
// by reading GOROOT/src/path/filepath/symlink.go, symlink_windows.go, and
// GOROOT/src/os/file_windows.go for this toolchain -- retained below as
// the historical record of what closed the 013-S root/candidate
// canonicalization asymmetry (see the risk register above):
//  1. NewRoot's one-time root canonicalization (this file). PRE-016.007-T,
//     via Go's OWN standard-library filepath.EvalSymlinks. Its
//     walkSymlinks (symlink.go) resolves each path component via plain
//     os.Lstat / os.Readlink -- NOT GetFinalPathNameByHandleW -- and
//     normally never touched uncPrefix at all. The ONE narrow exception:
//     os.Readlink's Windows implementation (file_windows.go,
//     normaliseLinkPath) special-cases a symlink target expressed as
//     an NT-native `\??\Volume{GUID}\...` path, where it either
//     directly string-substitutes `\\?\` for `\??\` (the
//     winreadlinkvolume GODEBUG's non-"0" fast path, avoiding any
//     Win32 call), or -- only under the legacy winreadlinkvolume="0"
//     opt-out -- opens the link and calls
//     windows.GetFinalPathNameByHandle to normalize it. Either way,
//     this was an edge case (a GUID-only-addressable volume symlink
//     target), not EvalSymlinks' general resolution mechanism. AS OF
//     016.007-T, NewRoot instead calls canonicalizeReparse directly (see
//     entry 2 below), which now applies unconditionally to the root path
//     too -- this historical EvalSymlinks path no longer executes here.
//  2. checkSymlinkEscape's per-Resolve-call containment check
//     (pathsafe.go), via canonicalizeReparse's own, separate, DIRECT,
//     UNCONDITIONAL GetFinalPathNameByHandleW call for every single
//     resolution (reparse_windows.go, 014.003-T, invoked through
//     syscall.NewLazyDLL, not through filepath.EvalSymlinks or
//     os.Readlink). 016.007-T makes NewRoot's root canonicalization
//     (entry 1) route through this exact same function.
//
// PRE-016.007-T, the two sites were independent, sometimes-divergent code
// paths that happened to converge on the same `\\?\` output convention
// only for the specific cases each of them actually emitted it -- that
// divergence was the root cause of the 013-S regression this register's
// asymmetry entry (above) records as closed. AS OF 016.007-T, both sites
// are the SAME call (canonicalizeReparse), so this divergence no longer
// exists; do not read the historical breakdown above as still describing
// current behavior.
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
// canonicalizeReparse (016.007-T, C2 -- previously filepath.EvalSymlinks;
// see the risk register above for the 013-S-regression provenance this
// change closes), then stripUNCPrefix. Canonicalizing once amortizes what
// would otherwise be a canonicalizeReparse syscall (CreateFile +
// GetFinalPathNameByHandleW on Windows) on every Resolve call — P10's
// diff-apply may validate 10-100 paths against one root (finding GO-4).
func NewRoot(dir string) (Root, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return Root{}, wrapRootInvalid(err)
	}

	resolved, err := canonicalizeReparse(abs)
	if err != nil {
		// C2/AC3 (016.007-T, Go review P1): canonicalizeReparse's Windows
		// implementation returns a raw syscall.Errno (from
		// syscall.CreateFile or the lazy-proc call), unlike
		// filepath.EvalSymlinks, which returns *fs.PathError.
		// errors.Is(err, fs.ErrNotExist) still resolves either way
		// (Windows' Errno.Is maps ERROR_FILE_NOT_FOUND/
		// ERROR_PATH_NOT_FOUND), but a caller doing
		// errors.As(err, &*fs.PathError) would silently stop matching on
		// Windows without this wrap. wrapPathError re-forms the error as
		// *fs.PathError at this boundary (a no-op on POSIX, where
		// canonicalizeReparse already returns *fs.PathError), preserving
		// the observable error surface on both platforms. Op is "open",
		// not "lstat": canonicalizeReparse's Windows implementation
		// proves/disproves existence via syscall.CreateFile, not an
		// Lstat-family call (review finding, adversarial pass 1).
		return Root{}, wrapRootInvalid(wrapPathError("open", abs, err))
	}

	// D-2 (plan Sec.4): stripUNCPrefix is called here even though
	// canonicalizeReparse already applies it internally as of 016.003-T/
	// U6 (its return value is always already normalized). Retaining this
	// call is a DELIBERATE belt-and-suspenders decision recorded in the
	// 015-S plan (C2's own description: "retaining stripUNCPrefix"), not
	// vestigial dead code: stripUNCPrefix is idempotent (a no-op on an
	// already-stripped input), so keeping it here costs nothing and
	// removes any dependency on the internal call graph staying wired
	// exactly the way it is today. checkSymlinkEscape (pathsafe.go) makes
	// the identical deliberate retention decision (plan B2/AC3) for the
	// same reason, plus avoiding coupling pathsafe.go to a Windows-only
	// internal contract.
	canonical := stripUNCPrefix(resolved)
	// DOCUMENTED-UNREACHABLE, RE-DERIVED for 016.007-T (C2, plan AC5, H2)
	// -- not re-pasted from the pre-C2 filepath.EvalSymlinks-era
	// justification below, which no longer applies to this call site.
	// This os.Stat call cannot observe a failure in practice on EITHER
	// build:
	//   - Windows: canonicalizeReparse's syscall.CreateFile with
	//     OPEN_EXISTING has ALREADY proven the resolved target exists by
	//     the time this line runs -- a stronger existence proof than a
	//     Stat/Lstat walk, since CreateFile actually opens a handle to the
	//     entry rather than merely querying its metadata.
	//   - POSIX (reparse_other.go): canonicalizeReparse IS
	//     filepath.EvalSymlinks, whose own os.Lstat-based walk over every
	//     path component already proved existence identically to the
	//     pre-C2 code path.
	// On both builds, a subsequent os.Stat on that same, just-canonicalized
	// path failing would require the filesystem to change between the two
	// calls (a TOCTOU race), not a normal input-driven code path. Requiring
	// a "fails before, passes after" test for this branch is unsatisfiable
	// without an injectable stat seam, which would itself be a production
	// behavior change inside a tests-only unit (Width Isolation). The
	// error return remains defensive, not dead, code.
	//
	// This os.Stat call now runs on a stripUNCPrefix-normalized form, which
	// is why D-6's `\\?\Volume{GUID}\...` non-stripping is LOAD-BEARING,
	// not incidental: os.Stat must still receive a form Win32 APIs accept,
	// and a Volume{GUID} path has no non-`\\?\` equivalent to strip to.
	info, err := os.Stat(canonical)
	if err != nil {
		return Root{}, wrapRootInvalid(err)
	}
	if !info.IsDir() {
		return Root{}, apperr.Newf(apperr.KindPathViolation, "workspace root invalid: not a directory: %s", canonical)
	}

	return Root{path: canonical}, nil
}

// wrapPathError ensures canonicalizeReparse's error surface always exposes
// an *fs.PathError to callers, regardless of platform (016.007-T, C2/AC3).
// On POSIX, canonicalizeReparse delegates to filepath.EvalSymlinks, which
// already returns *fs.PathError, so this is a no-op (errors.As below
// succeeds immediately and the original error is returned unchanged). On
// Windows, canonicalizeReparse's raw syscall.Errno is re-formed into an
// *fs.PathError here so `errors.As(err, &*fs.PathError)` keeps matching on
// both platforms, not just POSIX.
func wrapPathError(op, path string, err error) error {
	if err == nil {
		return nil
	}
	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		return err
	}
	return &fs.PathError{Op: op, Path: path, Err: err}
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

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
//   - GO-14 (write-through-dangling-symlink): checkSymlinkEscape's
//     documented lexical-only acceptance is bounded to a dangling symlink
//     at the FINAL path component only, including a transitively dangling
//     final link whose chain ends unresolved. 012.003-T narrows, but does
//     not remove, the same write-through-outside-workspace primitive:
//     under the identical attacker capability, a strict-ancestor dangling
//     link or junction is now rejected, but Resolve("link") still accepts
//     the final-component form with one fewer path component. Intermediate
//     directory entries that exist but are not statable as contained
//     directories are rejected regardless of target because the target is
//     not yet verifiably contained; that class includes dangling links,
//     cycles, EACCES, and unresolvable reparse points.
//     symlinkEscapeMsg deliberately covers both genuine escape and
//     unverifiable-target rejections. A future shipment may flip the GO-14
//     regression lock only if it explicitly reconsiders this final-
//     component acceptance, updates the lock, and lands the replacement
//     boundary in the same change (see stash F133AB7E). STATUS:
//     accepted, oracle-parity. TRIGGER: mitigation is forced the moment any
//     caller uses a Resolve()'d path to WRITE through a dangling symlink
//     whose target is outside the workspace — i.e. the first real
//     persistence/file-write call site (the same trigger as the
//     database.path Constitution Check exception in
//     internal/config/validate.go rule 7, and this package's own mechanical
//     CI gate, scripts/check-write-path-precondition.sh).
//   - Resolve -> use TOCTOU window (see "Known limitations" above): there
//     is no atomic validate-then-open primitive in this package. STATUS:
//     accepted, oracle-parity. TRIGGER: same as GO-14 — forced the moment a
//     real write path exists.
//   - SEC-5 (EvalSymlinks ignores hardlinks, see "Known limitations"
//     above). STATUS: accepted, oracle-parity. TRIGGER: same as GO-14.
//   - 5FE4A7BE (012.007-T; case-folding risk register, not a BF5DE670
//     item): darwin currently under-folds because pathEqual/pathHasPrefix
//     fold on Windows only, while darwin is a shipped target and is
//     case-insensitive by default on APFS/HFS+. STATUS: accepted,
//     fail-closed, plausible/unconfirmed. TRIGGER: reproduce a legitimate
//     in-root darwin path rejected solely because the volume is case-
//     insensitive and the only difference is casing. Extending the fold to
//     darwin was rejected because a case-sensitive APFS volume would turn
//     that rejection into an acceptance, i.e. fail-open. The already-
//     enabled Windows fold is itself not the safe baseline: NTFS supports
//     per-directory case sensitivity, and WSL enables it on the
//     directories it creates, so a symlink resolving to a case-variant
//     sibling can be folded into acceptance. STATUS: accepted, fail-open
//     risk. TRIGGER: a workspace root on a case-sensitivity-enabled NTFS
//     or WSL-created tree.
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
// Windows via GetFinalPathNameByHandle.
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
		return true
	}
	if len(p) < 3 {
		return false
	}
	return ((p[0] >= 'A' && p[0] <= 'Z') || (p[0] >= 'a' && p[0] <= 'z')) && p[1] == ':' && p[2] == '\\'
}

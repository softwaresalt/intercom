package pathsafe

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/softwaresalt/intercom-go/internal/apperr"
)

const (
	// absMsg is emitted when a candidate is absolute or Windows
	// drive-relative (finding SEC-1).
	absMsg = "absolute paths are not allowed; use workspace-relative paths"
	// escapeMsg is emitted when a candidate's ".." components would escape
	// the workspace root.
	escapeMsg = "path attempts to escape workspace"
	// outsideMsg is emitted when a resolved candidate fails the
	// component-aware containment check (unit B3).
	outsideMsg = "path outside workspace"
	// symlinkEscapeMsg is emitted when a resolved, existing path's
	// canonicalized form is successfully verified to escape the root (unit
	// B4). Reserved for a GENUINE, VERIFIED escape only (014.006-T): the
	// target was successfully canonicalized and containment was checked
	// against a known, real path.
	symlinkEscapeMsg = "symlink target escapes workspace"
	// symlinkUnverifiableMsg is emitted when a resolved path's existing
	// entry cannot be verified as contained -- its target could not be
	// canonicalized at all (a dangling symlink/junction, ELOOP, EACCES, or
	// a malformed reparse point), or an ancestor blocks further descent
	// without itself being resolvable (a plain file, or a live junction
	// deferred per F3) -- rather than being verified to escape (014.006-T,
	// covers 2362BBB5(b)). apperr.KindPathViolation is retained for both
	// messages (no taxonomy change); the OS cause, when present, is
	// wrapped and recoverable via errors.Is/errors.As.
	symlinkUnverifiableMsg = "symlink target cannot be verified"
)

// normalize validates and decomposes candidate into its cleaned, contained
// path components. Go has no Path::components() typed enum like Rust, so the
// mechanism is specified explicitly (finding GO-5):
//
//  1. Before any cleaning, reject if filepath.VolumeName(candidate) != "".
//     This catches Windows drive-relative forms such as "C:foo" that
//     filepath.IsAbs reports as false and that would otherwise be accepted
//     as a normal component — on NTFS "C:foo" is an alternate-data-stream
//     reference. The oracle rejects these via Component::Prefix
//     (finding SEC-1 — P1).
//  2. Reject if filepath.IsAbs(candidate). Both rejections emit
//     "absolute paths are not allowed; use workspace-relative paths".
//  3. filepath.Clean(candidate), then split on filepath.Separator.
//  4. Walk components: ".." pops the accumulated stack and errors
//     "path attempts to escape workspace" if the stack is already empty;
//     "." and "" are skipped; all else is pushed.
//
// filepath.Clean already collapses interior a/b/../c -> a/c and preserves
// only leading "..". It also collapses redundant separators (a//b -> a/b)
// and strips trailing ones (a/b/ -> a/b) — exactly the cases Rust's
// component iterator elides by yielding only Normal components. The
// post-Clean walk therefore yields the same final stack as the oracle's
// pre-Clean component walk, and for the same reason rather than by
// coincidence (finding CORR-3).
//
// Candidates MUST be raw, unexpanded strings (011.008-T, resolves
// 8472E0A1 item (g)). normalize performs NO environment-variable or
// shell-style expansion of any kind: a segment such as "$VAR" or "%VAR%"
// is treated as a literal, ordinary path component — a directory or file
// LITERALLY named "$VAR" — never substituted with an environment value.
// Expansion, if a caller wants it, is entirely the CALLER's
// responsibility to perform (or not) BEFORE calling normalize/Resolve.
//
// This is deliberately NOT tightened into a refusal of "$"/"%"-containing
// segments: a literal, unexpanded "$VAR" segment resolves INSIDE the
// root (it is just an ordinarily-named directory), so it is already
// Principle IV-compliant on its own. Refusing such segments would also
// be a false-positive hazard against real filenames — Windows ships
// "$Recycle.Bin", "$WinREAgent", and "$SysReset" as legitimate top-level
// directories, and '%' is a legal POSIX filename byte.
//
// RESIDUAL (Constitution IV, named and tracked, not silently absorbed):
// the actual expansion vector this package cannot control is a CALLER
// that expands a candidate (e.g. via os.ExpandEnv or shell interpolation)
// BEFORE passing it to normalize/Resolve. pathsafe has no visibility into
// that expansion and cannot refuse it retroactively. Callers that accept
// externally-influenced path input and perform their own expansion are
// responsible for their own containment re-validation after expanding.
func normalize(candidate string) ([]string, error) {
	if filepath.VolumeName(candidate) != "" || filepath.IsAbs(candidate) || isRooted(candidate) {
		return nil, apperr.New(apperr.KindPathViolation, absMsg)
	}

	cleaned := filepath.Clean(candidate)
	parts := filepathSplitList(cleaned)

	stack := make([]string, 0, len(parts))
	for _, part := range parts {
		switch part {
		case "", ".":
			continue
		case "..":
			if len(stack) == 0 {
				return nil, apperr.New(apperr.KindPathViolation, escapeMsg)
			}
			stack = stack[:len(stack)-1]
		default:
			stack = append(stack, part)
		}
	}

	return stack, nil
}

// isRooted reports whether candidate begins with a path separator without a
// volume name (e.g. Unix-style "/etc/passwd" on Windows). Go's
// filepath.IsAbs treats such a path as volume-relative, not absolute, on
// Windows (it is "rooted" to the current drive rather than "absolute") — so
// IsAbs alone under-rejects. Component-based oracle semantics treat any
// root/prefix component as a rejection regardless of platform, so this
// candidate must also be rejected as an absolute-style path.
//
// The leading-'/' check applies on every platform (on Unix it is redundant
// with filepath.IsAbs but harmless; on Windows it closes the gap). The
// leading-backslash check is scoped to Windows only: on Unix, '\' has no
// path-separator meaning, so a candidate that merely begins with a literal
// backslash byte is a legitimate workspace-relative filename there and must
// not be rejected as rooted.
func isRooted(candidate string) bool {
	if candidate == "" {
		return false
	}
	if candidate[0] == '/' {
		return true
	}
	return runtime.GOOS == "windows" && candidate[0] == '\\'
}

// filepathSplitList splits p on the OS path separator without relying on
// filepath.SplitList (which splits on the OS *list* separator, e.g. ';' or
// ':', not the path separator).
func filepathSplitList(p string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(p); i++ {
		if p[i] == filepath.Separator {
			parts = append(parts, p[start:i])
			start = i + 1
		}
	}
	parts = append(parts, p[start:])
	return parts
}

// Resolve validates candidate against the root's lexical and containment
// rules and returns the resolved absolute path.
//
// Postcondition: the returned path is absolute, cleaned, and strictly
// contained within the root (finding GO-15).
//
// Empty-candidate guard (required): if normalize() yields zero components,
// Resolve returns "path outside workspace" rather than falling through.
// Without this guard, filepath.Clean("") returns ".", the walk skips it,
// and filepath.Join(root, "") returns root — silently handing the caller
// the workspace root itself as a writable target (finding CORR-2,
// empirically confirmed).
//
// Deliberate omission: the oracle's absolute-path fast path
// (src/diff/path_safety.rs — an already-absolute candidate that
// Path::starts_with(root) skips the walk) is not ported, because normalize
// rejects all absolute candidates. Callers must pass workspace-relative
// paths (finding SEC-3).
func (r Root) Resolve(candidate string) (string, error) {
	components, err := normalize(candidate)
	if err != nil {
		return "", err
	}
	if len(components) == 0 {
		return "", apperr.New(apperr.KindPathViolation, outsideMsg)
	}

	joined := filepath.Join(append([]string{r.path}, components...)...)
	if !hasPathPrefix(joined, r.path) {
		return "", apperr.New(apperr.KindPathViolation, outsideMsg)
	}

	return checkSymlinkEscape(r, joined)
}

// hasPathPrefix reports whether path is exactly prefix, or is contained
// within prefix as a directory — i.e. path equals prefix, or path begins
// with prefix followed by a path separator. This is explicitly
// component-aware (finding GO-6): a raw strings.HasPrefix(path, prefix)
// check would incorrectly admit a sibling directory named "<prefix>-evil".
//
// Case is folded only on Windows (finding GO-7), since a config-supplied
// root and an EvalSymlinks-canonicalized candidate can differ only in case
// on that platform.
func hasPathPrefix(path, prefix string) bool {
	if pathEqual(path, prefix) {
		return true
	}
	if strings.HasSuffix(prefix, string(filepath.Separator)) {
		return pathHasPrefix(path, prefix)
	}
	return pathHasPrefix(path, prefix+string(filepath.Separator))
}

// pathEqual compares two paths for equality, folding case on Windows only.
func pathEqual(a, b string) bool {
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// pathHasPrefix reports whether path has the literal prefix p, folding case
// on Windows only.
func pathHasPrefix(path, p string) bool {
	if runtime.GOOS == "windows" {
		if len(path) < len(p) {
			return false
		}
		return strings.EqualFold(path[:len(p)], p)
	}
	return strings.HasPrefix(path, p)
}

// checkSymlinkEscape implements oracle steps 6-7, extended to close a gap
// beyond the literal oracle port: it walks up from resolved to the nearest
// existing ancestor (not just checking resolved itself), re-canonicalizes
// that ancestor through canonicalizeReparse, and re-asserts containment.
// This is required because the dominant real-world use case — creating a
// new file — has a non-existent leaf component; gating the check on
// resolved's own existence alone would let a symlinked or junctioned
// *intermediate directory* pointing outside the workspace silently pass
// validation whenever the leaf does not yet exist. On failure it emits
// "symlink target escapes workspace".
//
// 014.004-T (rev 2 of the 700B41CE remediation): every probe, including the
// first one against resolved itself, now uses os.Lstat rather than
// os.Stat. This is the mount-point-aware canonicalization the corrected
// root cause requires (see the package's 700B41CE risk-register entry in
// root.go): the defect was never "the final component's reparse identity is
// unexamined" in isolation, it was that containment was checked against a
// path that had never been mount-point-canonicalized. Lstat never follows
// the exact final path component it is given, so a live OR dangling
// symlink/junction sitting at that exact position is detected as a reparse
// entry (or at minimum as "something exists here") instead of being
// transparently traversed (a live target) or reported as simply absent (a
// dangling target, the pre-rev-2 GO-14 acceptance). Once an existing
// ancestor is found this way, canonicalizeReparse (platform-split,
// 014.003-T) replaces the old filepath.EvalSymlinks(ancestor) call:
// GetFinalPathNameByHandleW semantics resolve IO_REPARSE_TAG_MOUNT_POINT in
// addition to IO_REPARSE_TAG_SYMLINK, at any path position, which
// filepath.EvalSymlinks does not guarantee for a reparse point that is the
// exact terminal element of the path being evaluated (finding A1: the C1
// case from 014.001-T). A component whose target cannot be resolved this
// way (canonicalizeReparse returns an error — e.g. a dangling symlink,
// ELOOP, or EACCES) is rejected, never accepted (AC3): this is what flips
// the GO-14 dangling-final-symlink case from accept to reject (014.002-T).
//
// The strict-ancestor rejection for an entry that is neither a directory
// nor a symlink (Lstat's Mode()&fs.ModeSymlink) is retained UNCHANGED from
// the pre-rev-2 function (AC4): a live directory junction used as a
// strict ancestor still reports fs.ModeIrregular (neither ModeDir nor
// ModeSymlink) and is still unconditionally rejected without ever reaching
// canonicalizeReparse, regardless of whether its real target is inside or
// outside root. This is a deliberate, documented false-rejection
// (fail-closed usability gap, tracked as residual F3, deferred — see
// root.go), not a relaxation: relaxing it would require making this branch
// itself reparse-aware, which rev 1 attempted and rev 2's adversarial
// review (finding A2) identified as introducing a NEW intermediate-junction
// bypass. This check therefore never fires for the three 014.001-T cases:
// a junction used as the exact final probed component (C1) or reached only
// as an intermediate component of a successful Lstat on a deeper path (C2,
// C3) is never separately Lstat'd as its own standalone strict ancestor in
// those traces — the junction's redirection is instead captured by
// canonicalizeReparse operating on the fuller path that contains it.
//
// GODEBUG independence (AC5): this fix does not rely on
// os.Stat/os.Lstat/filepath.EvalSymlinks reparse-resolution semantics for
// the canonicalization step — canonicalizeReparse calls
// GetFinalPathNameByHandleW directly — so the verdict is identical under
// GODEBUG=winsymlink=0 and =1 by construction. An empirical both-settings
// proof is deferred (§8 of the shipment plan, YAGNI per scope audit).
//
// Ascent is limited to fs.ErrNotExist. Any other probe error — for example
// permission denial, a symlink cycle, or a malformed reparse point — is
// rejected with the OS cause wrapped for errors.Is/errors.As inspection.
//
// Known limitations (both oracle-parity, see the package doc comment):
// the TOCTOU window between this check and actual filesystem use, and the
// fact that canonicalizeReparse does not resolve hardlinks, so a
// pre-existing in-workspace hardlink to an external file on the same
// volume passes validation (finding SEC-5).
func checkSymlinkEscape(root Root, resolved string) (string, error) {
	ancestor := resolved
	finalProbe := true
	for {
		info, err := os.Lstat(ancestor)
		if err == nil {
			if !finalProbe && !info.IsDir() && info.Mode()&fs.ModeSymlink == 0 {
				// Ancestor blocks further descent without being resolvable
				// itself (a plain file, or a live junction -- F3 deferred):
				// unverifiable, not a proven escape (014.006-T).
				return "", apperr.New(apperr.KindPathViolation, symlinkUnverifiableMsg)
			}
			break
		} else if !errors.Is(err, fs.ErrNotExist) {
			// Any non-ENOENT probe error (ELOOP, EACCES, a malformed
			// reparse point) means the entry exists but cannot be
			// verified (014.006-T); the OS cause is wrapped for
			// errors.Is/errors.As inspection.
			return "", apperr.Wrapf(apperr.KindPathViolation, err, symlinkUnverifiableMsg)
		}

		parent := filepath.Dir(ancestor)
		if parent == ancestor {
			// DOCUMENTED-UNREACHABLE, coverage-excluded (011.006-T item
			// (e), resolves 8472E0A1 item (e); return value flipped by
			// 014.005-T, covers 2362BBB5(a) / AD0D9D1F(F6); message
			// aligned to symlinkUnverifiableMsg by 014.006-T for
			// consistency with every other unresolvable-ancestor path in
			// this function): reached the filesystem root without finding
			// an existing ancestor. resolved is already asserted (by
			// Root.Resolve's caller-side hasPathPrefix check before this
			// function is ever invoked) to be inside root, and root
			// itself is required by NewRoot to already exist and be a
			// directory -- so walking parent directories from inside an
			// existing root must find root itself (or a deeper existing
			// ancestor) before ever reaching the filesystem root.
			// Requiring a "fails before, passes after" test here is
			// unsatisfiable without an injectable filesystem seam, which
			// would be a production behavior change inside a tests-only
			// unit (Width Isolation). Guarded defensively only to avoid
			// an infinite loop, never exercised by real input -- the two
			// hypothesized triggers were never reproduced, so this is NOT
			// a reproduced-exploit fix (AC4). Previously returned
			// (resolved, nil) -- a latent fail-open terminal branch
			// (2362BBB5(a) / AD0D9D1F(F6), the same finding captured
			// twice, implemented once here). Now returns a rejection so
			// the branch is fail-closed like every other unresolvable-
			// ancestor path in this function, with zero observable
			// behavior change across the existing suite (the branch
			// remains traced-unreachable).
			return "", apperr.New(apperr.KindPathViolation, symlinkUnverifiableMsg)
		}
		ancestor = parent
		finalProbe = false
	}

	real, err := canonicalizeReparse(ancestor)
	if err != nil {
		// The entry exists (Lstat succeeded above) but its target could
		// not be canonicalized -- a dangling symlink/junction target,
		// ELOOP, EACCES, or a malformed reparse point. This is a correct
		// rejection with an accurate explanation: containment is
		// unverifiable, not a proven escape (014.006-T, covers
		// 2362BBB5(b)). The OS cause is wrapped for errors.Is/errors.As.
		return "", apperr.Wrapf(apperr.KindPathViolation, err, symlinkUnverifiableMsg)
	}
	real = stripUNCPrefix(real)

	if !hasPathPrefix(real, root.path) {
		return "", apperr.New(apperr.KindPathViolation, symlinkEscapeMsg)
	}

	return resolved, nil
}

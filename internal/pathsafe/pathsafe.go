package pathsafe

import (
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
	// canonicalized form escapes the root (unit B4).
	symlinkEscapeMsg = "symlink target escapes workspace"
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
// existing ancestor (not just checking resolved itself), re-resolves that
// ancestor's symlinks, and re-asserts containment. This is required because
// the dominant real-world use case — creating a new file — has a
// non-existent leaf component; gating the check on os.Stat(resolved) alone
// would let a symlinked *intermediate directory* pointing outside the
// workspace silently pass validation whenever the leaf does not yet exist.
// On failure it emits "symlink target escapes workspace".
//
// os.Stat (not os.Lstat) gates each existence probe, matching the oracle's
// Path::exists(), which follows symlinks — a broken symlink at the final
// component is therefore treated as non-existent and, once no existing
// ancestor remains to check beyond the root itself, takes the lexical-only
// branch (finding GO-14, documented, not silently inherited).
//
// Known limitations (both oracle-parity, see the package doc comment):
// the TOCTOU window between this check and actual filesystem use, and the
// fact that EvalSymlinks does not resolve hardlinks, so a pre-existing
// in-workspace hardlink to an external file on the same volume passes
// validation (finding SEC-5).
func checkSymlinkEscape(root Root, resolved string) (string, error) {
	ancestor := resolved
	for {
		if _, err := os.Stat(ancestor); err == nil {
			break
		}
		parent := filepath.Dir(ancestor)
		if parent == ancestor {
			// DOCUMENTED-UNREACHABLE, coverage-excluded (011.006-T item
			// (e), resolves 8472E0A1 item (e)): reached the filesystem
			// root without finding an existing ancestor. resolved is
			// already asserted (by Root.Resolve's caller-side hasPathPrefix
			// check before this function is ever invoked) to be inside
			// root, and root itself is required by NewRoot to already
			// exist and be a directory -- so walking parent directories
			// from inside an existing root must find root itself (or a
			// deeper existing ancestor) before ever reaching the
			// filesystem root. Requiring a "fails before, passes after"
			// test here is unsatisfiable without an injectable filesystem
			// seam, which would be a production behavior change inside a
			// tests-only unit (Width Isolation). Guarded defensively only
			// to avoid an infinite loop, never exercised by real input.
			return resolved, nil
		}
		ancestor = parent
	}

	real, err := filepath.EvalSymlinks(ancestor)
	if err != nil {
		return "", apperr.New(apperr.KindPathViolation, symlinkEscapeMsg)
	}
	real = stripUNCPrefix(real)

	if !hasPathPrefix(real, root.path) {
		return "", apperr.New(apperr.KindPathViolation, symlinkEscapeMsg)
	}

	return resolved, nil
}

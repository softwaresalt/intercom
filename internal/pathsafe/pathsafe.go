package pathsafe

import (
	"path/filepath"

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
func normalize(candidate string) ([]string, error) {
	if filepath.VolumeName(candidate) != "" || filepath.IsAbs(candidate) || isRooted(candidate) {
		return nil, apperr.New(apperr.KindPathViolation, absMsg)
	}

	cleaned := filepath.Clean(candidate)
	parts := splitPath(cleaned)

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

// splitPath splits a filepath.Clean-ed path on filepath.Separator.
func splitPath(cleaned string) []string {
	return filepathSplitList(cleaned)
}

// isRooted reports whether candidate begins with a path separator without a
// volume name (e.g. Unix-style "/etc/passwd" on Windows). Go's
// filepath.IsAbs treats such a path as volume-relative, not absolute, on
// Windows (it is "rooted" to the current drive rather than "absolute") — so
// IsAbs alone under-rejects. Component-based oracle semantics treat any
// root/prefix component as a rejection regardless of platform, so this
// candidate must also be rejected as an absolute-style path.
func isRooted(candidate string) bool {
	if candidate == "" {
		return false
	}
	return candidate[0] == '/' || candidate[0] == '\\'
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

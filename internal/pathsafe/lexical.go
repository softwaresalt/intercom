package pathsafe

import "strings"

// ContainsDotDotSegment reports whether path contains a literal ".."
// path-separator-delimited segment, checked against the original
// (uncleaned) string so a traversal component cannot be hidden by
// filepath.Clean's collapsing behavior. Both '/' and '\' are treated as
// separators regardless of GOOS, since a config file is portable text
// that may be authored on a different platform than the one that loads
// it.
//
// Export-surface limits (011.005-T, resolves 798002CB): this is a single,
// narrow lexical check, NOT a general containment control. It provides
// exactly none of the following three guarantees:
//   - No canonicalization: the input is never filepath.Clean'd,
//     filepath.Abs'd, or symlink-resolved before this check runs.
//   - No absolute-path rejection: an absolute path containing zero ".."
//     segments passes cleanly (this function only looks for the literal
//     segment; it does not reject absolute paths at all).
//   - No symlink resolution: a path that is lexically clean of ".." can
//     still, once resolved, point outside an intended root via a symlink;
//     this function cannot detect that.
//
// Callers that need real workspace containment (canonicalization,
// absolute-path rejection, AND symlink-escape detection) must use
// Root.Resolve instead, not this function alone.
func ContainsDotDotSegment(path string) bool {
	for _, part := range strings.FieldsFunc(path, func(r rune) bool {
		return r == '/' || r == '\\'
	}) {
		if part == ".." {
			return true
		}
	}
	return false
}

package pathsafe

import "strings"

// ContainsDotDotSegment reports whether path contains a literal ".."
// path-separator-delimited segment, checked against the original
// (uncleaned) string so a traversal component cannot be hidden by
// filepath.Clean's collapsing behavior. Both '/' and '\' are treated as
// separators regardless of GOOS, since a config file is portable text
// that may be authored on a different platform than the one that loads
// it.
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

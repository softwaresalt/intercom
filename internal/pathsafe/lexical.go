package pathsafe

import (
	"path/filepath"
	"strings"
)

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

// CLIPathKind classifies the lexical shape of a configured CLI executable
// path (copilot.cli_path). It is portable: the classification of UNC and
// device-namespace spellings does not depend on GOOS.
type CLIPathKind uint8

// CLIPathKind values.
const (
	// CLIPathEmpty is the empty string.
	CLIPathEmpty CLIPathKind = iota
	// CLIPathBare is a bare executable name resolved through PATH.
	CLIPathBare
	// CLIPathAbs is a rooted path on the current GOOS (excluding UNC/device).
	CLIPathAbs
	// CLIPathUNCOrDevice is a UNC or Windows device-namespace path.
	CLIPathUNCOrDevice
	// CLIPathDriveRelative is a Windows drive-relative path such as "C:x".
	CLIPathDriveRelative
	// CLIPathRelative is any other path containing a separator, or "." / "..".
	CLIPathRelative
)

// ClassifyCLIPath returns the CLIPathKind of path. UNC and device-namespace
// prefixes are detected before filepath.IsAbs on every GOOS.
func ClassifyCLIPath(path string) CLIPathKind {
	switch {
	case path == "":
		return CLIPathEmpty
	case hasUNCOrDevicePrefix(path):
		return CLIPathUNCOrDevice
	case filepath.IsAbs(path):
		return CLIPathAbs
	case hasWindowsDrivePrefix(path):
		return CLIPathDriveRelative
	case path == "." || path == ".." || strings.ContainsAny(path, `/\`):
		return CLIPathRelative
	default:
		return CLIPathBare
	}
}

func isSlash(b byte) bool { return b == '/' || b == '\\' }

// hasUNCOrDevicePrefix reports whether path begins with two slashes of either
// flavor (UNC, "\\.\" and "\\?\" device forms) or the NT "\??\" prefix.
func hasUNCOrDevicePrefix(path string) bool {
	if len(path) >= 2 && isSlash(path[0]) && isSlash(path[1]) {
		return true
	}
	return len(path) >= 4 && isSlash(path[0]) && path[1] == '?' && path[2] == '?' && isSlash(path[3])
}

// hasWindowsDrivePrefix reports whether path begins with an ASCII drive
// letter followed by a colon.
func hasWindowsDrivePrefix(path string) bool {
	if len(path) < 2 || path[1] != ':' {
		return false
	}
	c := path[0]
	return ('A' <= c && c <= 'Z') || ('a' <= c && c <= 'z')
}

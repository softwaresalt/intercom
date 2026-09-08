//go:build !windows

package pathsafe

import "path/filepath"

// canonicalizeReparse delegates to filepath.EvalSymlinks on non-Windows
// platforms (014.003-T AC2): POSIX has no mount-point reparse construct
// analogous to a Windows directory junction, so ordinary symlink resolution
// is already the correct, complete canonicalization.
func canonicalizeReparse(path string) (string, error) {
	return filepath.EvalSymlinks(path)
}

//go:build !windows

package pathsafe

import "path/filepath"

// canonicalizeReparse delegates to filepath.EvalSymlinks on non-Windows
// platforms (014.003-T AC2): POSIX has no mount-point reparse construct
// analogous to a Windows directory junction, so ordinary symlink resolution
// is already the correct, complete canonicalization.
//
// Postcondition (016.003-T, U6): the returned path carries no `\\?\`
// extended-path prefix. On POSIX this holds trivially -- filepath.EvalSymlinks
// never produces that Windows-specific prefix form -- so this function's
// package-level contract is identical on both platforms: the return value is
// always already normalized, and no caller on either build needs to strip
// anything itself.
func canonicalizeReparse(path string) (string, error) {
	return filepath.EvalSymlinks(path)
}

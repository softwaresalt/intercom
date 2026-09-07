//go:build !windows

package pathsafe

// isSymlinkPrivilegeError always reports false on non-Windows platforms: the
// ERROR_PRIVILEGE_NOT_HELD skip exception (011.010-T) is Windows-specific.
// Non-Windows symlink-creation failures are not gated by this exception.
func isSymlinkPrivilegeError(err error) bool {
	return false
}

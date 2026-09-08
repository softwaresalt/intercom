//go:build !windows

package integration

func isSymlinkPrivilegeError(err error) bool {
	return false
}

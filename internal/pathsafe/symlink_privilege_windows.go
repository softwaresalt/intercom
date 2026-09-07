//go:build windows

package pathsafe

import (
	"errors"
	"syscall"
)

// errPrivilegeNotHeld is Windows ERROR_PRIVILEGE_NOT_HELD (1314), returned by
// CreateSymbolicLink (surfaced through os.Symlink) when the calling process
// lacks SeCreateSymbolicLinkPrivilege and is neither elevated nor running
// with Developer Mode enabled. This is the single enumerated
// precondition-unmet reason 011.010-T permits a Windows symlink-dependent
// test to skip rather than fail.
const errPrivilegeNotHeld = syscall.Errno(1314)

// isSymlinkPrivilegeError reports whether err is the enumerated
// ERROR_PRIVILEGE_NOT_HELD skip exception.
func isSymlinkPrivilegeError(err error) bool {
	return errors.Is(err, errPrivilegeNotHeld)
}

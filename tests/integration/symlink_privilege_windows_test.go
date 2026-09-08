//go:build windows

package integration

import (
	"errors"
	"syscall"
)

// errPrivilegeNotHeld is Windows ERROR_PRIVILEGE_NOT_HELD (1314), returned by
// CreateSymbolicLink (surfaced through os.Symlink) when the calling process
// lacks SeCreateSymbolicLinkPrivilege and is neither elevated nor running
// with Developer Mode enabled. Mirrors internal/pathsafe's own definition
// (012.008-T); duplicated rather than imported because the source lives in
// an unexported _test.go symbol.
const errPrivilegeNotHeld = syscall.Errno(1314)

func isSymlinkPrivilegeError(err error) bool {
	return errors.Is(err, errPrivilegeNotHeld)
}

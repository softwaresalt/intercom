//go:build windows

package integration

import (
	"errors"
	"syscall"
)

const errPrivilegeNotHeld = syscall.Errno(1314)

func isSymlinkPrivilegeError(err error) bool {
	return errors.Is(err, errPrivilegeNotHeld)
}

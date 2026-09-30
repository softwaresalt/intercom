//go:build windows

package main

import (
	"errors"
	"os/exec"
)

// errSignalGroupUnsupported is returned by signalProcessGroup on windows,
// which has no POSIX process-group signal delivery. The runner tests skip
// their INT/TERM subtests on this platform (the Windows job is advisory
// only per the migration plan); these stubs exist solely so this package
// still builds and vets cleanly under GOOS=windows.
var errSignalGroupUnsupported = errors.New("process-group signal delivery is not supported on windows")

func setNewProcessGroup(cmd *exec.Cmd) {
	// No-op: nothing to configure. See signalTestsSupported.
	_ = cmd
}

func signalProcessGroup(cmd *exec.Cmd, signum int) error {
	_, _ = cmd, signum
	return errSignalGroupUnsupported
}

func signalTestsSupported() bool { return false }

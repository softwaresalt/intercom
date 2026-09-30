//go:build !windows

package main

import (
	"os/exec"
	"syscall"
)

// setNewProcessGroup places the wrapper's bash process (and any foreground
// child it forks, since children inherit their parent's process group by
// default) into its own new process group. This lets the INT/TERM signal
// tests deliver a signal to the WHOLE group at once — exactly how a real
// terminal Ctrl+C or CI cancellation propagates — so both bash and its
// currently-sleeping foreground child (the fake binary) receive the signal
// simultaneously. Bash defers a pending trap until the current foreground
// command completes; killing the foreground child via the same signal is
// what makes that completion happen promptly, which is what proves the
// runner's traps actually intercept the signal instead of merely
// coincidentally reporting the same 128+N-style exit code.
func setNewProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// signalProcessGroup delivers signum to the negative PID (the whole
// process group created by setNewProcessGroup).
func signalProcessGroup(cmd *exec.Cmd, signum int) error {
	return syscall.Kill(-cmd.Process.Pid, syscall.Signal(signum))
}

// signalTestsSupported reports whether this platform can exercise the
// process-group signal delivery the INT/TERM runner tests rely on.
func signalTestsSupported() bool { return true }

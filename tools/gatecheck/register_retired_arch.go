package main

import "io"

// init pre-registers the retired-arch sub-command as a stub. M2-T11 replaces
// this file's body with the real dispatch to internal/retiredarch once the
// engine lands (plan §4, M2-T11); no other register_<name>.go file is
// touched by that change.
func init() {
	registerSubcommand("retired-arch", stubRetiredArch)
}

func stubRetiredArch(args []string, root string, stdin io.Reader, stdout, stderr io.Writer) int {
	return writeExitError(stderr, &exitError{code: 1, msg: "retired-arch: not yet ported"})
}

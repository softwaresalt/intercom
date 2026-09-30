package main

import "io"

// init pre-registers the write-path sub-command as a stub. M1-T10 replaces
// this file's body with the real dispatch to internal/writepath once the
// engine lands (plan §4, M1-T10); no other register_<name>.go file is
// touched by that change.
func init() {
	registerSubcommand("write-path", stubWritePath)
}

func stubWritePath(args []string, root string, stdin io.Reader, stdout, stderr io.Writer) int {
	return writeExitError(stderr, &exitError{code: 1, msg: "write-path: not yet ported"})
}

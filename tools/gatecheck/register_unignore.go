package main

import "io"

// init pre-registers the unignore sub-command as a stub. M3-T7 replaces this
// file's body with the real dispatch to internal/unignore once the engine
// lands (plan §4, M3-T7); no other register_<name>.go file is touched by
// that change.
func init() {
	registerSubcommand("unignore", stubUnignore)
}

func stubUnignore(args []string, root string, stdin io.Reader, stdout, stderr io.Writer) int {
	return writeExitError(stderr, &exitError{code: 1, msg: "unignore: not yet ported"})
}

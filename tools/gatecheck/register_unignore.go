package main

import (
	"io"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/unignore"
)

// init pre-registers the unignore sub-command, dispatching to
// internal/unignore.Run (M3-T7; plan §6). No other register_<name>.go file
// is touched by this change.
func init() {
	registerSubcommand("unignore", runUnignore)
}

// runUnignore forwards every remaining positional argument straight to
// unignore.Run, which absorbs its own bash wrapper's argument parsing
// (--self-test, --base-ref, --head-ref, -h/--help) directly -- unlike
// retiredarch's single-flag adaptation, unignore's real CLI surface needs
// multiple flags forwarded, mirroring write-path's pattern instead.
func runUnignore(args []string, root string, stdin io.Reader, stdout, stderr io.Writer) int {
	return unignore.Run(args, root, unignore.DefaultGitRunner, stdout, stderr)
}

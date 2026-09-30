package main

import (
	"io"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/writepath"
)

// init pre-registers the write-path sub-command, dispatching to
// internal/writepath.Run (M1-T10; plan §4). No other register_<name>.go
// file is touched by this change.
func init() {
	registerSubcommand("write-path", runWritePath)
}

// runWritePath forwards the first remaining positional argument (if any) as
// writepath.Run's mode flag, mirroring the bash wrapper's own
// `"${1:-}"` dispatch: any further args are ignored, exactly as bash's case
// statement never inspected $2 onward.
func runWritePath(args []string, root string, stdin io.Reader, stdout, stderr io.Writer) int {
	flag := ""
	if len(args) > 0 {
		flag = args[0]
	}
	return writepath.Run(flag, root, writepath.DefaultGitRunner, stdout, stderr)
}

package main

import (
	"io"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/retiredarch"
)

// init pre-registers the retired-arch sub-command, dispatching to
// internal/retiredarch.Run (M2-T11; plan §5). No other register_<name>.go
// file is touched by this change.
func init() {
	registerSubcommand("retired-arch", runRetiredArch)
}

// runRetiredArch forwards the first remaining positional argument (if any)
// as retiredarch.Run's mode flag, mirroring the (pre-M2-T11) bash
// wrapper's own `"${1:-}"` dispatch and write-path's identical adaptation
// pattern (register_write_path.go): any further args are ignored, exactly
// as bash's case statement never inspected $2 onward.
func runRetiredArch(args []string, root string, stdin io.Reader, stdout, stderr io.Writer) int {
	flag := ""
	if len(args) > 0 {
		flag = args[0]
	}
	return retiredarch.Run(flag, root, retiredarch.DefaultGitRunner, stdout, stderr)
}

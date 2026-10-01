package main

import (
	"io"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/mergestrategy"
)

// init pre-registers the merge-strategy-evaluate sub-command, dispatching
// to internal/mergestrategy.Run (M3-T11; plan §6). No other
// register_<name>.go file is touched by this change.
func init() {
	registerSubcommand("merge-strategy-evaluate", runMergeStrategyEvaluate)
}

// runMergeStrategyEvaluate is a thin forward: mergestrategy.Run already
// matches the subcommandFunc signature exactly (args, root, stdin,
// stdout, stderr), so no argument adaptation is needed here.
func runMergeStrategyEvaluate(args []string, root string, stdin io.Reader, stdout, stderr io.Writer) int {
	return mergestrategy.Run(args, root, stdin, stdout, stderr)
}

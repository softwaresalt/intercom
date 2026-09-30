package main

import "io"

// init pre-registers the merge-strategy-evaluate sub-command as a stub.
// M3-T11 replaces this file's body with the real dispatch to
// internal/mergestrategy once the engine lands (plan §4, M3-T11); no other
// register_<name>.go file is touched by that change.
func init() {
	registerSubcommand("merge-strategy-evaluate", stubMergeStrategyEvaluate)
}

func stubMergeStrategyEvaluate(args []string, root string, stdin io.Reader, stdout, stderr io.Writer) int {
	return writeExitError(stderr, &exitError{code: 1, msg: "merge-strategy-evaluate: not yet ported"})
}

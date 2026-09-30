// Command gatecheck is the Go port of the repository's gate-engine scripts
// (write-path, retired-arch, unignore, merge-strategy-evaluate). It is part
// of the root module (github.com/softwaresalt/intercom-go), not a separate
// module, and lives outside cmd/ and internal/ specifically so gate
// pathspecs anchored at the repository root (internal/**, cmd/**) never
// select its own sources (C-1).
//
// Engines land under tools/gatecheck/internal/{pysem,gomask,writepath,
// retiredarch,unignore,mergestrategy}; Go's internal-visibility rule is the
// only isolation mechanism required (C-1). This file (main.go) owns the
// exit-code contract (C-2) and the common --root flag (C-3). Each
// sub-command is pre-registered from its own register_<name>.go file via
// init(), so later migration units replace different files and never touch
// adjacent lines (see plan §13 "Blast-radius minimization").
package main

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// exitError is the typed error/exit-code pair every gate engine returns
// instead of calling os.Exit directly. main is the only os.Exit call site
// (C-2).
type exitError struct {
	code int
	msg  string
}

// Error implements the error interface so exitError composes with ordinary
// Go error handling inside engine code.
func (e *exitError) Error() string { return e.msg }

// writeExitError writes e's message (plus a trailing newline) to stderr and
// returns e's code. It is the single place stub and engine code converts an
// *exitError into the (stderr-write, return-code) pair main ultimately
// surfaces as the process exit code.
func writeExitError(stderr io.Writer, e *exitError) int {
	_, _ = fmt.Fprintf(stderr, "%s\n", e.msg)
	return e.code
}

// subcommandFunc is the signature every registered sub-command implements.
// args is whatever remains after run() has consumed the sub-command name
// and the common --root flag; root is the validated, existing directory
// path.
type subcommandFunc func(args []string, root string, stdin io.Reader, stdout, stderr io.Writer) int

// subcommands is the dispatch table populated by each register_<name>.go's
// init(). It is package-level so main_test.go can register and remove a
// test-only entry (see TestRun_PanickingStubRecovers) without touching
// production registrations.
var subcommands = map[string]subcommandFunc{}

// registerSubcommand adds name to the dispatch table. Each register_<name>.go
// file calls this exactly once, from its own init().
func registerSubcommand(name string, fn subcommandFunc) {
	subcommands[name] = fn
}

// printUsage writes the tool's usage text to stderr. It is used for both the
// no-args and unknown-sub-command cases (C-2: both return 2).
func printUsage(stderr io.Writer) {
	_, _ = fmt.Fprintln(stderr, "usage: gatecheck <subcommand> --root <dir> [args...]")
	_, _ = fmt.Fprintln(stderr, "subcommands: write-path, retired-arch, unignore, merge-strategy-evaluate")
}

// parseRoot scans args for a "--root <dir>" pair, removes it, and returns the
// validated directory path plus the remaining args (in their original
// relative order). It returns an error if --root is missing, has no value,
// or does not name an existing directory.
func parseRoot(args []string) (root string, rest []string, err error) {
	rest = make([]string, 0, len(args))
	for i := 0; i < len(args); i++ {
		if args[i] == "--root" {
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("--root requires a value")
			}
			root = args[i+1]
			i++
			continue
		}
		if v, ok := strings.CutPrefix(args[i], "--root="); ok {
			root = v
			continue
		}
		rest = append(rest, args[i])
	}
	if root == "" {
		return "", nil, fmt.Errorf("--root is required")
	}
	info, statErr := os.Stat(root)
	if statErr != nil {
		return "", nil, fmt.Errorf("--root %q: %w", root, statErr)
	}
	if !info.IsDir() {
		return "", nil, fmt.Errorf("--root %q is not a directory", root)
	}
	return root, rest, nil
}

// run is the fully testable entry point: no package-level state beyond the
// subcommand table, no direct os.Exit, and every I/O stream passed in
// explicitly.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer) (code int) {
	defer func() {
		if r := recover(); r != nil {
			_, _ = fmt.Fprintf(stderr, "::error::gatecheck internal error: %v\n", r)
			code = 1
		}
	}()

	if len(args) == 0 {
		printUsage(stderr)
		return 2
	}

	name := args[0]
	fn, ok := subcommands[name]
	if !ok {
		printUsage(stderr)
		return 2
	}

	root, rest, err := parseRoot(args[1:])
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "::error::%s\n", err.Error())
		return 1
	}

	return fn(rest, root, stdin, stdout, stderr)
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

// This file (checkignore.go) ports is_ignored, is_ignored_batch and
// parse_check_ignore_verbose_line from the check-unignore-regression.sh
// Python heredoc (plan §6, M3-T4).
package unignore

import (
	"fmt"
	"strings"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem"
)

// isIgnored reproduces is_ignored: `git check-ignore --no-index -q --
// <relPath>` run in cwd. Exit 0 means ignored, exit 1 means not ignored,
// and any other exit is a real error (fail-closed: never silently treated
// as either verdict).
func isIgnored(git GitRunner, cwd, relPath string) (bool, error) {
	_, stderr, err := git(cwd, nil, "check-ignore", "--no-index", "-q", "--", relPath)
	switch exitCode(err) {
	case 0:
		return true, nil
	case 1:
		return false, nil
	default:
		return false, fmt.Errorf(
			"::error::git check-ignore errored for %s in %s: exit %d: %s",
			pysem.Repr(relPath), cwd, exitCode(err), strings.TrimSpace(string(stderr)),
		)
	}
}

// isIgnoredBatch reproduces is_ignored_batch: the batched stdin form of
// isIgnored, used for the differential check's potentially large
// candidate universe. Uses `--verbose --non-matching --stdin` rather than
// passing every path as a positional argument, since a bare batched
// `git check-ignore path1 path2 ...` exits 0 if ANY argument matches --
// exactly the masking hazard this checker exists to avoid. A line-count
// mismatch between input and output is a hard error, matching Python's
// same fail-closed check (M3-T4 AC).
func isIgnoredBatch(git GitRunner, cwd string, relPaths []string) ([]bool, error) {
	if len(relPaths) == 0 {
		return nil, nil
	}
	stdinPayload := strings.Join(relPaths, "\n") + "\n"
	stdout, stderr, err := git(cwd, []byte(stdinPayload), "check-ignore", "--no-index", "--verbose", "--non-matching", "--stdin")
	switch exitCode(err) {
	case 0, 1:
		// fall through
	default:
		return nil, fmt.Errorf(
			"::error::git check-ignore --stdin errored in %s: exit %d: %s",
			cwd, exitCode(err), strings.TrimSpace(string(stderr)),
		)
	}
	text, convErr := pysem.GitText(stdout)
	if convErr != nil {
		return nil, convErr
	}
	lines := strings.Split(text, "\n")
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) != len(relPaths) {
		return nil, fmt.Errorf(
			"::error::git check-ignore --stdin returned %d lines for %d input paths in %s; cannot safely align results",
			len(lines), len(relPaths), cwd,
		)
	}
	out := make([]bool, len(lines))
	for i, line := range lines {
		out[i] = parseCheckIgnoreVerboseLine(line)
	}
	return out, nil
}

// parseCheckIgnoreVerboseLine reproduces parse_check_ignore_verbose_line:
// `--verbose` reports the LAST matching pattern regardless of whether it
// is a positive pattern or a negation (`!pattern`) -- a naive "any match
// at all" reading would misclassify a negation match (which RE-INCLUDES
// the path, i.e. NOT ignored) as ignored. Format is
// "source:linenum:pattern<TAB>pathname", or exactly "::<TAB>pathname" for
// a non-matching path (--non-matching).
func parseCheckIgnoreVerboseLine(line string) bool {
	prefix, _, _ := strings.Cut(line, "\t")
	if prefix == "::" {
		return false
	}
	parts := strings.SplitN(prefix, ":", 3)
	pattern := ""
	if len(parts) == 3 {
		pattern = parts[2]
	}
	return !strings.HasPrefix(pattern, "!")
}

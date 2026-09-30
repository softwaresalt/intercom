// This file (select.go) ports should_scan_repo_path, engine_for_path,
// scan_path and select_repo_paths from scripts/lib/retired_arch.py.
//
// The pathspec/prefix literals inside shouldScanRepoPath and
// selectRepoPaths are pinned by pin.go (M2-T8), which parses THIS file's
// own source with go/parser and requires each literal to appear as a
// *ast.BasicLit inside the correct function body. Do not refactor these
// literals into shared constants, helper variables, or another file --
// doing so would make the pin fail closed (by design; see pin.go).
package retiredarch

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem"
)

// GitRunner runs `git ls-files -- <pathspecs...>` rooted at root and
// returns its raw stdout bytes, or an error if the process could not be
// started or exited non-zero. It is injectable so tests can simulate a
// missing/failing git without depending on the real repository tree
// (mirrors writepath.GitRunner's shape/contract exactly).
type GitRunner func(root string, pathspecs ...string) ([]byte, error)

// DefaultGitRunner is the production GitRunner: it shells out to
// `git ls-files -- <pathspecs...>` with root as the working directory.
func DefaultGitRunner(root string, pathspecs ...string) ([]byte, error) {
	args := append([]string{"ls-files", "--"}, pathspecs...)
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	var stdout, stderrBuf bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderrBuf
	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(stderrBuf.String()); msg != "" {
			return nil, fmt.Errorf("%w: %s", err, msg)
		}
		return nil, err
	}
	return stdout.Bytes(), nil
}

// shouldScanRepoPath ports should_scan_repo_path verbatim. path is a
// repo-relative, forward-slashed path exactly as `git ls-files` prints it.
func shouldScanRepoPath(path string) bool {
	if path == "config.toml.example" {
		return true
	}
	if strings.HasPrefix(path, "cmd/") {
		return strings.HasSuffix(path, ".go")
	}
	if !strings.HasPrefix(path, "internal/") {
		return false
	}
	if strings.Contains(path, "/testdata/") || strings.HasSuffix(path, "_test.go") {
		return false
	}
	return strings.HasSuffix(path, ".go")
}

// engineForPath ports engine_for_path. path is an OS-native (possibly
// absolute) filesystem path; Python's Path.name/Path.suffix are matched
// with filepath.Base/filepath.Ext (Ext includes the leading dot, matching
// Python's path.suffix).
func engineForPath(path string) string {
	if filepath.Base(path) == "config.toml.example" {
		return "toml"
	}
	switch filepath.Ext(path) {
	case ".go":
		return "go"
	case ".toml":
		return "toml"
	default:
		return ""
	}
}

// scanPath ports scan_path. An unmapped extension yields a fail-closed
// SYNTHETIC FINDING (015.010-T's V7 fix, ED-2 territory) rather than a
// panic or a silent empty return: run_repo_scan collects findings across
// every selected path in a loop, and this path must travel through the
// exact same reporting/advisory-toggle machinery as any real finding.
//
// A per-file read error from scanGo is converted into one synthetic
// finding line here rather than propagated as a Go error, matching ED-2
// ("a read error... a synthetic finding where the engine has a finding
// channel"). The "toml" case below never reaches this conversion: real
// production TOML dispatch only ever calls scanTomlPrimary (never
// scanTomlFallback, which is exercised only via --self-test), and
// scanTomlPrimary already converts its own read/parse failures into a
// fail-closed finding string internally (see tomlprimary.go) before this
// function ever sees a result -- so scanPath itself has no TOML I/O-error
// path to document beyond the one noted in the "toml" case's own comment
// below (Copilot review finding, PR #83, round 9: an earlier version of
// this comment incorrectly attributed this conversion to "scanGo's or
// scanTomlFallback's I/O error" as if both engines' errors landed here).
func scanPath(path string) []string {
	switch engineForPath(path) {
	case "go":
		findings, err := scanGo(path, true)
		if err != nil {
			return []string{fmt.Sprintf("%s: %v (fail-closed synthetic finding)", filepath.ToSlash(path), err)}
		}
		return findings
	case "toml":
		// scanTomlPrimary never returns a Go error: like Python's
		// scan_toml_with_tomllib, it converts a read/parse failure into
		// its own fail-closed finding string already (see tomlprimary.go).
		return scanTomlPrimary(path)
	default:
		return []string{fmt.Sprintf("%s: no scan engine registered for this file type (fail-closed synthetic finding)", filepath.ToSlash(path))}
	}
}

// selectRepoPaths ports select_repo_paths. The pathspec arguments
// (config.toml.example, cmd/**, internal/**) are literal string
// arguments in THIS call, exactly as pin.go requires -- never a
// variable, slice literal defined elsewhere, or constant reference.
func selectRepoPaths(root string, git GitRunner) ([]string, error) {
	out, err := git(root, "config.toml.example", "cmd/**", "internal/**")
	if err != nil {
		return nil, err
	}
	listing, err := pysem.GitText(out)
	if err != nil {
		return nil, err
	}
	var selected []string
	for _, path := range pysem.SplitLines(listing) {
		if shouldScanRepoPath(path) {
			selected = append(selected, path)
		}
	}
	sort.Strings(selected)
	return selected, nil
}

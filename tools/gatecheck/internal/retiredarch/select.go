// This file (select.go) ports should_scan_repo_path, engine_for_path,
// scan_path and select_repo_paths from the M4-deleted retired_arch module.
//
// The scan-scope surface of this file (the imports, GitRunner,
// DefaultGitRunner, gitRunnerEnv, scanArm, scanScope, shouldScanRepoPath
// and selectRepoPaths) is pinned by pin.go (M2-T8; D-030-4, D-030-6). Treat
// these declarations as frozen. Edit them only together with pin.go's
// independent expectation, in the same commit; otherwise the pin fails
// closed. scanScope is the single permitted home for the scope literals.
package retiredarch

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem"
)

// GitRunner runs `git ls-files -z -- <pathspecs...>` rooted at root and
// returns its raw stdout bytes (NUL-terminated records), or an error if the
// process could not be started or exited non-zero. It is injectable so tests
// can simulate a missing/failing git without depending on the real repository
// tree (mirrors writepath.GitRunner's shape/contract exactly).
type GitRunner func(root string, pathspecs ...string) ([]byte, error)

// DefaultGitRunner is the production GitRunner: it shells out to
// `git ls-files -z -- <pathspecs...>` with root as the working directory. The
// -z flag makes git emit NUL-terminated, unquoted paths, so quoted and
// non-ASCII tracked paths are never dropped (4537B2F6).
//
// The child runs with gitRunnerEnv's isolated environment (D7BF9F74), so
// ambient GIT_* variables and global/system git config cannot change the
// selection. A LookPath failure (cmd.Err) is reported first, and a git
// that PATH resolved to a non-absolute path is refused before it is
// launched.
func DefaultGitRunner(root string, pathspecs ...string) ([]byte, error) {
	args := append([]string{"ls-files", "-z", "--"}, pathspecs...)
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	cmd.Env = gitRunnerEnv(os.Environ())
	if cmd.Err != nil {
		return nil, cmd.Err
	}
	if !filepath.IsAbs(cmd.Path) {
		return nil, fmt.Errorf("retiredarch: refusing non-absolute git path %q", cmd.Path)
	}
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

// gitRunnerEnv returns the environment for a gate-owned git child process
// (D7BF9F74): environ with every entry whose name (the text before the
// first '=', ASCII-case-folded) starts with GIT_ removed, except
// GIT_CEILING_DIRECTORIES, followed by GIT_CONFIG_NOSYSTEM=1 and
// GIT_CONFIG_GLOBAL/GIT_CONFIG_SYSTEM pointed at os.DevNull. Windows
// "=C:"-style per-drive entries have an empty name and are kept.
func gitRunnerEnv(environ []string) []string {
	env := make([]string, 0, len(environ)+3)
	for _, kv := range environ {
		name, _, _ := strings.Cut(kv, "=")
		folded := []byte(name)
		for i, c := range folded {
			if 'a' <= c && c <= 'z' {
				folded[i] = c - ('a' - 'A')
			}
		}
		if strings.HasPrefix(string(folded), "GIT_") && string(folded) != "GIT_CEILING_DIRECTORIES" {
			continue
		}
		env = append(env, kv)
	}
	return append(env, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_CONFIG_SYSTEM="+os.DevNull)
}

// scanArm is one arm of the retired-architecture scan scope. pathspec is
// the `git ls-files` pathspec form of the arm. An exact arm (prefix == "")
// matches only the path equal to exact. A prefix arm matches every .go path
// under prefix; includeTests says whether that arm also keeps _test.go files
// and testdata/ paths.
type scanArm struct {
	pathspec     string
	prefix       string
	exact        string
	includeTests bool
}

// scanScope is the single declaration of the retired-architecture scan
// scope (D6/D6c, as broadened by 012-S): config.toml.example, every Go file
// under cmd/ INCLUDING tests and testdata (the guarded D4/AG-5 asymmetry,
// 015.011-T), and every non-test Go file under internal/. Both
// selectRepoPaths (the pathspec vector) and shouldScanRepoPath (the path
// predicate) derive from it, so a scope change is one edit here, made in the
// same commit as pin.go's independent expectation (§A-CANON).
var scanScope = []scanArm{
	{pathspec: "config.toml.example", exact: "config.toml.example"},
	{pathspec: "cmd/**", prefix: "cmd/", includeTests: true},
	{pathspec: "internal/**", prefix: "internal/", includeTests: false},
}

// shouldScanRepoPath ports should_scan_repo_path, evaluated from scanScope.
// path is a repo-relative, forward-slashed path exactly as `git ls-files`
// prints it.
func shouldScanRepoPath(path string) bool {
	for _, a := range scanScope {
		if a.prefix == "" {
			if path == a.exact {
				return true
			}
			continue
		}
		if strings.HasPrefix(path, a.prefix) {
			if !a.includeTests && (strings.Contains(path, "/testdata/") || strings.HasSuffix(path, "_test.go")) {
				return false
			}
			return strings.HasSuffix(path, ".go")
		}
	}
	return false
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

// selectRepoPaths ports select_repo_paths. The pathspec vector is built
// from scanScope, in declaration order; it restates no scope literal.
//
// The scan-scope surface of this file (the imports, GitRunner,
// DefaultGitRunner, gitRunnerEnv, scanArm, scanScope, shouldScanRepoPath
// and selectRepoPaths) is pinned by pin.go (M2-T8; D-030-4, D-030-6). Treat
// these declarations as frozen. Edit them only together with pin.go's
// independent expectation, in the same commit; otherwise the pin fails
// closed. scanScope is the single permitted home for the scope literals.
func selectRepoPaths(root string, git GitRunner) ([]string, error) {
	pathspecs := make([]string, 0, len(scanScope))
	for _, a := range scanScope {
		pathspecs = append(pathspecs, a.pathspec)
	}
	out, err := git(root, pathspecs...)
	if err != nil {
		return nil, err
	}
	if !utf8.Valid(out) {
		return nil, pysem.ErrInvalidUTF8
	}
	if len(out) == 0 {
		return nil, nil
	}
	if out[len(out)-1] != 0 {
		return nil, fmt.Errorf("output is not NUL-terminated")
	}
	var selected []string
	for _, path := range strings.Split(string(out[:len(out)-1]), "\x00") {
		if path == "" {
			return nil, fmt.Errorf("output has an empty record")
		}
		for _, r := range path {
			if r < 0x20 || r == 0x7f || r == 0x85 || r == 0x2028 || r == 0x2029 {
				return nil, fmt.Errorf("path %q contains a control or line-separator character", path)
			}
		}
		if shouldScanRepoPath(path) {
			selected = append(selected, path)
		}
	}
	sort.Strings(selected)
	return selected, nil
}

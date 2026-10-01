package unignore

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// workingTreeIsClean reports whether `git status --porcelain` in dir
// produces no output at all (no untracked, modified, or staged changes).
// Used to guard the exact self-test parity test, which enumerates the
// real working tree's untracked files and would otherwise see a
// different (still legitimate) candidate set than the one the golden was
// captured against.
func workingTreeIsClean(t *testing.T, dir string) bool {
	t.Helper()
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		t.Logf("git status --porcelain failed: %v", err)
		return false
	}
	return len(strings.TrimSpace(string(out))) == 0
}

// writeTestGitignore writes content as repoDir's root .gitignore file,
// creating repoDir if needed. This is used by tests that only need
// rootGitignoreTextAt's ref=="HEAD" disk-read path (no actual git
// repository required), as opposed to makeScratchGitignore, which always
// creates one.
func writeTestGitignore(repoDir, content string) error {
	if err := os.MkdirAll(repoDir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(repoDir, ".gitignore"), []byte(content), 0o644)
}

// newIsolatedGitRunner returns a GitRunner that isolates every git
// invocation it makes from this machine's real git configuration (plan
// §6 "Git test isolation"): GIT_CONFIG_GLOBAL/GIT_CONFIG_SYSTEM point at
// os.DevNull, GIT_CONFIG_NOSYSTEM=1, GIT_TERMINAL_PROMPT=0 are set, and
// GIT_DIR, GIT_WORK_TREE, GIT_INDEX_FILE, GIT_COMMON_DIR and
// GIT_OBJECT_DIRECTORY are removed from the child environment, so a
// hook-invoked `go test` cannot touch the real repository. Every
// invocation additionally carries `-c user.name=gatecheck -c
// user.email=gatecheck@invalid -c core.autocrlf=false -c
// init.defaultBranch=main`. Skips the calling test when git is not found
// on PATH.
func newIsolatedGitRunner(t *testing.T) GitRunner {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found on PATH")
	}

	env := filteredEnviron()
	env = append(env,
		"GIT_CONFIG_GLOBAL="+os.DevNull,
		"GIT_CONFIG_SYSTEM="+os.DevNull,
		"GIT_CONFIG_NOSYSTEM=1",
		"GIT_TERMINAL_PROMPT=0",
	)

	isolationArgs := []string{
		"-c", "user.name=gatecheck",
		"-c", "user.email=gatecheck@invalid",
		"-c", "core.autocrlf=false",
		"-c", "init.defaultBranch=main",
	}

	return func(dir string, stdin []byte, args ...string) ([]byte, []byte, error) {
		fullArgs := append(append([]string{}, isolationArgs...), args...)
		cmd := exec.Command("git", fullArgs...)
		cmd.Dir = dir
		cmd.Env = env
		if stdin != nil {
			cmd.Stdin = bytes.NewReader(stdin)
		}
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		return stdout.Bytes(), stderr.Bytes(), err
	}
}

// filteredEnviron returns the current process environment with GIT_DIR,
// GIT_WORK_TREE, GIT_INDEX_FILE, GIT_COMMON_DIR and GIT_OBJECT_DIRECTORY
// removed.
func filteredEnviron() []string {
	blocked := map[string]bool{
		"GIT_DIR":              true,
		"GIT_WORK_TREE":        true,
		"GIT_INDEX_FILE":       true,
		"GIT_COMMON_DIR":       true,
		"GIT_OBJECT_DIRECTORY": true,
	}
	src := os.Environ()
	out := make([]string, 0, len(src))
	for _, kv := range src {
		key, _, _ := strings.Cut(kv, "=")
		if blocked[key] {
			continue
		}
		out = append(out, kv)
	}
	return out
}

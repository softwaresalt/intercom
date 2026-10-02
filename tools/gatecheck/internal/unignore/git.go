// Package unignore reimplements the repository's un-ignore regression gate
// (scripts/check-unignore-regression.sh) on top of
// tools/gatecheck/internal/pysem, so the two-part check -- a fixed denylist
// asserted at head-ref, plus an old-vs-new ignored/un-ignored differential
// over currently-untracked-and-touched candidates -- can be enforced
// without a Python interpreter (plan §6, M3).
//
// git() is ported through an injectable GitRunner (plan §6, M3-T3), never
// a direct os/exec call inside the check functions themselves, so tests
// can simulate git failures and so every test that runs the real git
// binary can be driven from an isolated environment (plan §6's git test
// isolation rules) without the production code needing any test-only
// branch.
package unignore

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem"
)

// GitRunner runs `git <args...>` rooted at dir, feeding stdin (nil for
// none) on the child process's standard input, and returns its raw stdout
// and stderr bytes, or an error if the process could not be started. A
// non-zero exit is NOT itself turned into an error here -- callers that
// need to distinguish git's own exit codes (e.g. check-ignore's 0/1/other
// convention) use exitCode. This mirrors retiredarch.GitRunner and
// writepath.GitRunner's shape, widened to accept arbitrary git subcommands
// and an optional stdin payload (unignore calls check-ignore --stdin,
// show, diff, init, config, add, commit and rev-parse -- not just
// ls-files).
type GitRunner func(dir string, stdin []byte, args ...string) (stdout []byte, stderr []byte, err error)

// DefaultGitRunner is the production GitRunner: it shells out to
// `git <args...>` with dir as the working directory. It never sets any
// GIT_CONFIG_GLOBAL/GIT_CONFIG_NOSYSTEM-style isolation environment
// itself -- the pre-port Python engine never did either (plain
// subprocess.run(['git', ...])), so faithful parity keeps the production
// path identical. Hermetic isolation for git subprocesses spawned DURING
// a `go test` run is instead the test's own responsibility (t.Setenv),
// exactly as plan §6's "Git test isolation" rule describes.
func DefaultGitRunner(dir string, stdin []byte, args ...string) ([]byte, []byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if stdin != nil {
		cmd.Stdin = bytes.NewReader(stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	return stdout.Bytes(), stderr.Bytes(), err
}

// exitCode extracts the process exit code from a GitRunner error, or -1 if
// the process never started (err is non-nil but not an *exec.ExitError) or
// err is nil (exit 0).
func exitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return -1
}

// rootGitignoreTextAt reproduces root_gitignore_text_at verbatim,
// including its 035-F degradation (035-F: fixed after M4): at ref ==
// "HEAD" it reads the root .gitignore directly off disk (an absent file
// is treated as empty content, not an error); at any other ref it reads
// `git show <ref>:.gitignore` via git, and ANY non-zero exit from that
// command -- whether the ref itself is invalid, or the file is simply
// absent at that ref -- degrades to an empty baseline rather than
// surfacing an error. This degradation is faithfully preserved, not
// fixed, by this port (035-F re-plans the fix for after M4).
func rootGitignoreTextAt(repoDir, ref string, git GitRunner) (string, error) {
	if ref == "HEAD" {
		data, err := os.ReadFile(filepath.Join(repoDir, ".gitignore"))
		if err != nil {
			if os.IsNotExist(err) {
				return "", nil
			}
			return "", err
		}
		text, convErr := pysem.GitText(data)
		if convErr != nil {
			return "", convErr
		}
		return text, nil
	}
	stdout, showStderr, err := git(repoDir, nil, "show", fmt.Sprintf("%s:.gitignore", ref))
	if err != nil {
		// 035-F: an invalid ref, or a ref that simply lacks a root
		// .gitignore, both degrade to an empty baseline. Faithfully
		// ported, not fixed, here.
		return "", classifyShowFailure(repoDir, ref, git, err, showStderr)
	}
	text, convErr := pysem.GitText(stdout)
	if convErr != nil {
		return "", convErr
	}
	return text, nil
}

// classifyShowFailure decides what a failed `git show <ref>:.gitignore`
// means. git exits 128 both for an unresolvable ref and for a resolvable
// ref whose tree has no root .gitignore, and its stderr text is
// locale-dependent, so neither is inspected to tell the two apart.
// Instead git is asked two structured questions: does ref peel to a tree
// (`git rev-parse --verify --quiet <ref>^{tree}`), and does that tree's
// exact root listing (`git ls-tree -z <tree>`) contain a .gitignore entry?
//
// It returns nil ONLY when the ref resolves and the root .gitignore is
// positively absent -- the legitimate empty-baseline case. Every other
// outcome is a non-nil error: an unresolvable ref, git failing to run,
// or a .gitignore entry that exists but could not be shown.
func classifyShowFailure(repoDir, ref string, git GitRunner, showErr error, showStderr []byte) error {
	if exitCode(showErr) < 0 {
		return fmt.Errorf("::error::git show %s:.gitignore could not run: %v", ref, showErr)
	}
	// A leading "-" can never be a resolvable ref, and passing it to
	// rev-parse would risk it being parsed as an option.
	if strings.HasPrefix(ref, "-") {
		return fmt.Errorf("::error::cannot resolve ref %q for the root .gitignore baseline: not a valid ref", ref)
	}
	treeOut, revStderr, revErr := git(repoDir, nil, "rev-parse", "--verify", "--quiet", ref+"^{tree}")
	if revErr != nil {
		return fmt.Errorf(
			"::error::cannot resolve ref %q for the root .gitignore baseline: git rev-parse --verify failed: %v: %s",
			ref, revErr, strings.TrimSpace(string(revStderr)),
		)
	}
	tree := strings.TrimSpace(string(treeOut))
	if tree == "" {
		return fmt.Errorf("::error::cannot resolve ref %q for the root .gitignore baseline: git rev-parse --verify returned no object", ref)
	}
	listing, lsStderr, lsErr := git(repoDir, nil, "ls-tree", "-z", tree)
	if lsErr != nil {
		return fmt.Errorf(
			"::error::git ls-tree %s failed while resolving the root .gitignore at ref %q: %v: %s",
			tree, ref, lsErr, strings.TrimSpace(string(lsStderr)),
		)
	}
	present, parseErr := rootTreeHasGitignore(listing)
	if parseErr != nil {
		return fmt.Errorf("::error::cannot parse git ls-tree output for ref %q: %v", ref, parseErr)
	}
	if present {
		return fmt.Errorf(
			"::error::git show %s:.gitignore failed although the root .gitignore exists at that ref: %v: %s",
			ref, showErr, strings.TrimSpace(string(showStderr)),
		)
	}
	return nil
}

// rootTreeHasGitignore reports whether a `git ls-tree -z` listing of a
// single tree contains an entry named exactly ".gitignore". Each record is
// "<mode> SP <type> SP <object> TAB <name> NUL"; a record without a TAB is
// malformed and reported as an error rather than skipped.
func rootTreeHasGitignore(listing []byte) (bool, error) {
	for _, record := range bytes.Split(listing, []byte{0}) {
		if len(record) == 0 {
			continue
		}
		_, name, ok := bytes.Cut(record, []byte{'\t'})
		if !ok {
			return false, fmt.Errorf("malformed ls-tree record %q", record)
		}
		if string(name) == ".gitignore" {
			return true, nil
		}
	}
	return false, nil
}

// makeScratchGitignore reproduces make_scratch_gitignore: it isolates
// evaluation to ONLY the supplied .gitignore content (never the real
// working tree's nested .gitignore files or untracked machine-local
// noise), by creating a fresh, otherwise-empty git repository under
// scratchRoot/name and writing content as that repo's own root
// .gitignore. git check-ignore (even with --no-index) requires SOME git
// repository context to run at all, which is why the scratch directory is
// `git init`-ed rather than left as a bare directory.
func makeScratchGitignore(scratchRoot, name, content string, git GitRunner) (string, error) {
	scratchDir := filepath.Join(scratchRoot, name)
	if err := os.MkdirAll(scratchDir, 0o755); err != nil {
		return "", err
	}
	if _, stderr, err := git(scratchDir, nil, "init", "-q"); err != nil {
		return "", fmt.Errorf("git init -q in %s: %w: %s", scratchDir, err, strings.TrimSpace(string(stderr)))
	}
	if err := os.WriteFile(filepath.Join(scratchDir, ".gitignore"), []byte(content), 0o644); err != nil {
		return "", err
	}
	return scratchDir, nil
}

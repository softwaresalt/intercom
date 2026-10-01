// This file (checks.go) ports the DENYLIST constant, run_denylist_check,
// run_differential_check and run_ls_files_check from the
// check-unignore-regression.sh Python heredoc (plan §6, M3-T5).
package unignore

import (
	"fmt"
	"strings"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem"
)

// Denylist is the revision-5 fixed denylist: ONLY paths guaranteed
// ignored by HEAD (plus the two 011.014-T fixture-isolation paths, and
// the depguard/copilot probe markers) -- never a pattern introduced by
// droppable content (standing rule). Ported verbatim from DENYLIST in the
// Python heredoc.
var Denylist = []string{
	".env",
	".env.local",
	"config.toml",
	".claude/instructions.md",
	".github/copilot/settings.local.json",
	".autoharness/gates/pipeline-topology-force-audit.log",
	".backlogit/hooks_queue.jsonl",
	"internal/depguardfixture/marker",
	"internal/copilotprobe2/marker",
}

// runDenylistCheck reproduces run_denylist_check (Part 1): it evaluates
// every Denylist entry against the root .gitignore content resolved at
// ref, isolated inside a scratch git repository. Returns the evaluated
// count (always len(Denylist)) and the subset of entries that are NOT
// ignored (a regression).
func runDenylistCheck(git GitRunner, repoDir, scratchRoot, ref string) (evaluated int, failures []string, err error) {
	content, err := rootGitignoreTextAt(repoDir, ref, git)
	if err != nil {
		return 0, nil, err
	}
	scratchName := "denylist-" + strings.ReplaceAll(ref, "/", "_")
	scratchDir, err := makeScratchGitignore(scratchRoot, scratchName, content, git)
	if err != nil {
		return 0, nil, err
	}
	for _, path := range Denylist {
		ignored, ignErr := isIgnored(git, scratchDir, path)
		if ignErr != nil {
			return 0, nil, ignErr
		}
		if !ignored {
			failures = append(failures, path)
		}
	}
	return len(Denylist), failures, nil
}

// runDifferentialCheck reproduces run_differential_check (Part 2). The
// candidate universe is the UNION of (a) all untracked paths REALLY
// PRESENT in the working tree, ignored or not (`git ls-files --others`,
// WITHOUT --exclude-standard -- broader than "--ignored", since a path
// that is untracked and no longer ignored at head would never appear
// under an --ignored-filtered enumeration), and (b) every path actually
// touched between baseRef and headRef (`git diff --name-only`), so a PR
// that both removes a negation/pattern AND `git add`s the now-un-ignored
// file in the SAME change (making it TRACKED, not untracked) is still
// evaluated. Order-stable, de-duplicated union, exactly as the Python
// engine builds it.
func runDifferentialCheck(git GitRunner, repoDir, scratchRoot, baseRef, headRef string) (evaluated int, failures []string, err error) {
	untrackedOut, untrackedErr, gitErr := git(repoDir, nil, "ls-files", "--others", "-z")
	if gitErr != nil {
		return 0, nil, fmt.Errorf("::error::git ls-files --others failed: %v: %s", gitErr, strings.TrimSpace(string(untrackedErr)))
	}
	allUntracked := splitNulTerminated(untrackedOut)

	diffOut, diffErrBytes, diffErr := git(repoDir, nil, "diff", "--name-only", "-z", baseRef, headRef)
	if diffErr != nil {
		return 0, nil, fmt.Errorf(
			"::error::git diff --name-only failed resolving %s..%s: %s",
			baseRef, headRef, strings.TrimSpace(string(diffErrBytes)),
		)
	}
	diffPaths := splitNulTerminated(diffOut)

	seen := make(map[string]bool, len(allUntracked)+len(diffPaths))
	var candidates []string
	for _, p := range append(append([]string{}, allUntracked...), diffPaths...) {
		if p == "" || seen[p] {
			continue
		}
		seen[p] = true
		candidates = append(candidates, p)
	}

	if len(candidates) == 0 {
		return 0, nil, nil
	}

	baseContent, err := rootGitignoreTextAt(repoDir, baseRef, git)
	if err != nil {
		return 0, nil, err
	}
	headContent, err := rootGitignoreTextAt(repoDir, headRef, git)
	if err != nil {
		return 0, nil, err
	}
	baseScratch, err := makeScratchGitignore(scratchRoot, "diff-base-"+strings.ReplaceAll(baseRef, "/", "_"), baseContent, git)
	if err != nil {
		return 0, nil, err
	}
	headScratch, err := makeScratchGitignore(scratchRoot, "diff-head-"+strings.ReplaceAll(headRef, "/", "_"), headContent, git)
	if err != nil {
		return 0, nil, err
	}

	baseIgnored, err := isIgnoredBatch(git, baseScratch, candidates)
	if err != nil {
		return 0, nil, err
	}
	var ignoredAtBase []string
	for i, p := range candidates {
		if baseIgnored[i] {
			ignoredAtBase = append(ignoredAtBase, p)
		}
	}
	evaluated = len(ignoredAtBase)
	if evaluated == 0 {
		return 0, nil, nil
	}

	headIgnored, err := isIgnoredBatch(git, headScratch, ignoredAtBase)
	if err != nil {
		return 0, nil, err
	}
	for i, p := range ignoredAtBase {
		if !headIgnored[i] {
			failures = append(failures, p)
		}
	}
	return evaluated, failures, nil
}

// splitNulTerminated splits raw NUL-terminated git output (as produced by
// -z) into its component paths, dropping the trailing empty element a
// terminal NUL produces, and decoding with pysem.GitText's universal-
// newline rule first (git never embeds a NUL inside a single -z record's
// own bytes, so translateNewlines is safe here).
func splitNulTerminated(raw []byte) []string {
	text, err := pysem.GitText(raw)
	if err != nil {
		// Invalid UTF-8 from `git ls-files`/`git diff` output is not a
		// class the Python engine handled specially either (it decoded
		// via text=True, which would itself raise); returning no
		// candidates here is strictly more conservative than a crash, and
		// unignore.Run's own caller surfaces pysem.ErrInvalidUTF8-class
		// errors from its other call sites regardless.
		return nil
	}
	parts := strings.Split(text, "\x00")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// runLsFilesCheck reproduces run_ls_files_check: `git ls-files -i -c
// --exclude-standard` lists TRACKED files that ALSO match an ignore
// pattern -- a state that should never occur on a healthy checkout.
func runLsFilesCheck(git GitRunner, repoDir string) ([]string, error) {
	stdout, stderr, err := git(repoDir, nil, "ls-files", "-i", "-c", "--exclude-standard")
	if err != nil {
		return nil, fmt.Errorf("::error::git ls-files -i -c --exclude-standard failed: %v: %s", err, strings.TrimSpace(string(stderr)))
	}
	text, convErr := pysem.GitText(stdout)
	if convErr != nil {
		return nil, convErr
	}
	var out []string
	for _, line := range pysem.SplitLines(text) {
		if line != "" {
			out = append(out, line)
		}
	}
	return out, nil
}

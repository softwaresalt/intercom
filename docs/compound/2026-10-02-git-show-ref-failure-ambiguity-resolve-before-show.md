---
title: "git show <ref>:<path> cannot tell an invalid ref from an absent file — resolve the ref first"
date: 2026-10-02
category: security
tags: [git, fail-open, gatecheck, unignore, option-injection, 032-s]
---

# git show <ref>:<path> cannot tell an invalid ref from an absent file — resolve the ref first

## Symptom

In 032-S / 035-F, `rootGitignoreTextAt` ran `git show <ref>:.gitignore`. It treated
any failure as "no `.gitignore`" and returned an empty baseline. An invalid or
misspelled `--base-ref` therefore produced an empty baseline. No ignore rule could
then be observed as removed, so the merge-blocking unignore-regression check
passed vacuously.

## Root cause

1. `git show <ref>:<path>` exits 128 **both** for an unresolvable ref and for a
   path that is absent from a valid tree. Its stderr wording depends on the
   locale, so neither the exit code nor the stderr text can portably tell the
   two cases apart.
2. Putting the caller's ref text into the `show` argument creates a second
   fail-open path (adversarial-review anchor P1):
   * A ref that starts with `-` is parsed as an option (for example
     `--format=x`).
   * An empty ref turns `<ref>:.gitignore` into `:.gitignore`, which reads the
     **index**.

   In both cases `git show` can *succeed* and return text that does not belong
   to any baseline.

## Fix / prevention

1. **Validate the ref, then resolve it.**
   * Reject an empty ref or one starting with `-` before any git call.
   * Resolve the ref with
     `git rev-parse --verify --quiet <ref>^{tree}`. It exits non-zero
     for invalid refs and for refs that point at blobs.
   * Require a hex object ID in the output.
   * Avoid `--end-of-options`, which older git releases do not support.
2. **Use the resolved OID afterwards.** Run `git show <treeOID>:<path>`, so caller
   text never reaches the `show` argument.
3. **Classify a `show` failure with a structured query, not stderr.** Use
   `git ls-tree --full-tree -z <treeOID>`. Parse the NUL-terminated records
   exactly. If no entry is named `<path>`, the file is genuinely absent and the
   result is `("", nil)`. Any other outcome is a `::error::` that names the ref.
   * An entry that is a directory, symlink or submodule is "present but
     unreadable", which is an error, not "absent".
4. **Test both halves.**
   * A scripted runner should assert the exact git arguments, and should fail on
     any git call it was not given a reply for. That proves rejected refs never
     reach git.
   * Add a real-git test for the option-like, empty and `:` refs.
   * Get any real `*exec.ExitError` you need through the isolated test runner,
     using a builtin such as `rev-parse`. Never run an un-isolated
     `git <unknown-subcommand>`: a user alias or a `git-*` executable on `PATH`
     could intercept it (Copilot finding on PR #93).

## Related

* `tools/gatecheck/internal/unignore/git.go` — `resolveBaselineTree`, `classifyShowFailure`,
  `rootTreeHasGitignore`.
* Stash `B29A565E` — the same option-injection hardening, deferred, for
  `checks.go` `git diff` and `runCheck` `rev-parse`.
* Stash `F5958BBC` — isolating the existing `TestExitCode_ExitError`.

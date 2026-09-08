---
title: "011-S pathsafe containment correctness — compacted session memory"
date: 2026-09-08
shipment: 011-S
feature: 012-F
status: shipped
compaction_tier: 1
---

## Outcome

Shipment `011-S` (feature `012-F`, pathsafe containment correctness and
cross-platform coverage) shipped and archived via the P-015 verified
fully-covered-root cascade path. Merged to `main` via PR #34, merge commit
`54d27ba953a068cea539a1f691e5c4ad1231d2ba`. All 9 tasks (`012.001-T`–
`012.009-T`) done. Preceded by a 3-PR recovery sequence (#31 corrective
revert, #32 predecessor-closure filename fix, #33 proper staging reapply)
for a Stage direct-to-main incident on the original `f6d5879` commit.

## Key decisions

* `checkSymlinkEscape`: depth-scoped `os.Stat`/`os.Lstat` probe (final
  component vs. strict ancestors) + ENOENT-only ascent closes the
  dangling-intermediate-symlink/junction escape (`ECE3DAB7`) while
  preserving the documented `GO-14` final-component acceptance by
  construction.
* `stripUNCPrefix`: re-forms extended-UNC input to ordinary UNC, gated by
  an OS-independent `isWindowsAbsolutePath` classifier requiring both a
  non-empty host and share segment (tightened post-Copilot-review to
  reject share-less forms).
* Review-driven cleanup: removed a provably-redundant manual
  `os.Readlink`+`EvalSymlinks` re-resolution in `checkSymlinkEscape`
  (two independent reviewers converged on the same finding).

## Critical discovery, deferred (not fixed here)

Adversarial review found, and Ship empirically verified by live
reproduction, a **pre-existing** (confirmed via `git show main`, not
introduced by this shipment) critical gap: a **live** Windows junction at
the **final** path component bypasses containment entirely (`os.Stat`
follows the reparse point, `EvalSymlinks` never resolves it). Out of
011-S's chartered scope (dangling case only) per P-021 C1. Captured as
stash `700B41CE` (critical, requires deliberation) with a risk-register
entry in `root.go` and full trace in
`docs/closure/2026-09-07-011-s-012-f-pathsafe-containment-adversarial-review.md`.
7 additional advisory findings bundled into stash `AD0D9D1F` (2 of the 7
were fixed directly per matching Copilot-review comments on PR #34).

## Reusable lesson (see compound entry)

`docs/compound/2026-09-08-adversarial-review-empirical-verification-and-scope-check.md`
— empirically reproduce a review agent's load-bearing security claim
before disposing of it, and check pre-existence against `main` before
making a P-021 scope call.

## Review/CI/merge summary

5 local persona reviews (Security/Correctness/Go/Scope Boundary/
Maintainability) → all findings fixed or deferred. Adversarial 4-model
consensus → 1 critical (deferred, pre-existing), disposed per P-021.
Copilot review on PR #34: 2 comments, both fixed/replied/resolved. CI:
all green (cross-compile ×4, lint 0 issues, security, test, race).
Merged via merge commit, no admin fallback.

---
title: "032-S / 035-F progress — unignore baseline ref resolution"
description: "Ship build-phase memory for shipment 032-S (plan unit B / S-8)"
date: 2026-10-02
shipment: 032-S
feature: 035-F
branch: feat/032-s-repair-unignore-regression-checker-ref-resolution
phase: build-complete
---

# 032-S / 035-F progress — unignore baseline ref resolution

## Completed

* `035.001-T` (red phase): inverted
  `TestRootGitignoreTextAt_InvalidRef_DegradesToEmpty` into
  `TestRootGitignoreTextAt_InvalidRef_ReturnsError` and added
  `TestRootGitignoreTextAt_ValidRefAbsentGitignore_ReturnsEmpty`. The inverted
  case failed against the pre-change `git.go` (AC-B1.2 verified), and the
  absent-file case passed both before and after.
* `035.002-T`: `rootGitignoreTextAt` now routes every `git show <ref>:.gitignore`
  failure through `classifyShowFailure`. It returns `("", nil)` only when
  `git rev-parse --verify --quiet <ref>^{tree}` resolves and an exact
  `git ls-tree -z <tree>` root listing has no `.gitignore` entry. Every other
  outcome is a `::error::` diagnostic that names the ref. Scripted-runner
  tests cover each branch.
* `035.003-T`: the stale doc and inline comments were rewritten
  (comment-only).

## Decisions

* **Stop condition not triggered.** A local probe (git 2.55) showed that
  `git show` exits 128 both for an invalid ref and for an absent file, and
  that its stderr text is locale-dependent. So neither the exit status nor
  stderr was used to tell the cases apart. Instead the code asks git two
  structured, locale-independent questions (`rev-parse --verify --quiet`,
  `ls-tree -z`). Invalid-ref and absent-file are distinguished portably, and
  absent-file was NOT made into an error.
* **Leading `-` refs.** On the new `rev-parse` call, a ref that starts with
  `-` is rejected as unresolvable instead of being passed to git, which
  could parse it as an option. The happy path (`git show` succeeding) is
  unchanged, which preserves INV-1.
* The three tasks landed as separate commits: `aff681f` (red), `ccc4b2a`
  (green) and `29a365a` (comments).

## Verification

`gofmt -l .` clean, `go vet ./...`, `go build ./...`, `go test ./...` (with Git
Bash first on PATH), `go test -race` for `internal/unignore`,
`golangci-lint run ./...` (0 issues), `staticcheck ./...`, `goimports -l`
clean. The wrapper `--self-test` passed. `--base-ref origin/main` passed
(denylist 9, differential 6303).

## Next

Multi-persona adversarial review at a pinned HEAD, then the feature PR, the
Copilot loop, the merge, safe-close and closure.
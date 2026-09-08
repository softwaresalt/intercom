---
title: "Go stdlib EvalSymlinks never resolves Windows directory junctions via GetFinalPathNameByHandleW — verify internals claims against actual GOROOT source, not plausible-sounding assumptions"
description: "Two separate doc-comment drafts claimed filepath.EvalSymlinks internally uses GetFinalPathNameByHandleW as a general resolution mechanism; both were false and had to be corrected after tracing the actual pinned Go stdlib source, and a related Copilot review compile-error claim about make()/uintptr was independently falsified the same way"
date: 2026-09-08
tags:
  - "compound"
  - "pathsafe"
  - "windows"
  - "go-stdlib"
  - "copilot-review"
  - "security-review"
shipment: 013-S
feature: 014-F
---

# Verify Go-internals claims against actual `GOROOT` source, not plausibility

## Context

Shipment 013-S fixed a Windows directory-junction containment bypass by
adding `canonicalizeReparse` (a direct `GetFinalPathNameByHandleW` call),
because the existing `checkSymlinkEscape` used `filepath.EvalSymlinks`,
which does not resolve `IO_REPARSE_TAG_MOUNT_POINT` (directory junctions).
While documenting *why* the two functions differ, two independent
plausible-sounding but **false** claims about Go's internals had to be
caught and corrected — one authored by this same session, one by an
automated Copilot review.

## Failure 1: an authored doc-comment claimed EvalSymlinks uses GetFinalPathNameByHandleW generally

A first-draft `uncPrefix` comment claimed `filepath.EvalSymlinks`
"internally calls `GetFinalPathNameByHandleW` to construct its return
value" as a general mechanism — this reads as plausible (both functions
canonicalize Windows paths) but is **false**. Verified directly against
this toolchain's actual `GOROOT/src/path/filepath/symlink.go`,
`symlink_windows.go`, and `GOROOT/src/os/file_windows.go`:

- `EvalSymlinks`' general mechanism (`walkSymlinks`) uses **only**
  `os.Lstat`/`os.Readlink` — never `GetFinalPathNameByHandleW`, and never
  resolves `IO_REPARSE_TAG_MOUNT_POINT`.
- The **only** place `GetFinalPathNameByHandle` (via Go's internal
  `internal/syscall/windows` package, not `golang.org/x/sys/windows`)
  appears anywhere in this call graph is inside `os.Readlink`'s
  Windows-specific `normaliseLinkPath` helper, for the narrow edge case of
  a symlink target expressed as an NT-native `\??\Volume{GUID}\...` path —
  and even then, gated behind the `winreadlinkvolume` GODEBUG setting,
  whose apparent default branch avoids the Win32 call entirely via direct
  string substitution.

The comment was rewritten to the narrower, verified-accurate claim. A
second-round adversarial review reviewer flagged that even the corrected
version needed the "narrow edge case, GODEBUG-gated" qualifier — the first
correction attempt was itself still too broad.

## Failure 2: a Copilot review flagged a false compile error on `make(..., uintptr)`

A separate Copilot-authored review comment claimed
`getFinalPathNameByHandle`'s `n, _, callErr := procGetFinalPathNameByHandleW.Call(...)`
followed by `make([]uint16, n+1)` "will not compile" because "`make`
requires an `int` length" and `n` is `uintptr`. This is also **false**: per
the Go language spec, `make`'s size arguments "must be of integer type,
have a non-negative value, or be an untyped constant" — `uintptr` is an
integer type, so no conversion is required. Verified two ways before
replying:

1. Empirically: the code already compiled, vetted, and passed
   `go test` (45/45) with this exact form, on this branch, before the
   Copilot review ran.
2. Isolated repro: `var n uintptr = 260; buf := make([]uint16, n+1)`
   compiles and runs under `go run` on this toolchain.

Declined with the rationale (Go spec citation + both verification methods)
in the PR review-thread reply, then resolved the thread — no code change.

## Reusable pattern for future sessions

- **A confident-sounding claim about Go/Windows stdlib internals — whether
  self-authored in a doc comment or from an automated review tool — is not
  trustworthy on its own, even when it sounds mechanically plausible.**
  Trace the actual pinned `GOROOT` source (`path/filepath/symlink*.go`,
  `os/file_windows.go`, or the relevant package) before writing or
  accepting the claim as fact. This is a repeat pattern from
  `2026-09-08-adversarial-review-empirical-verification-and-scope-check.md`
  (verify security-relevant static claims empirically) applied here to
  stdlib-internals and compile-correctness claims specifically.
- **A Copilot-reported "will not compile" finding is falsifiable in
  seconds**: the code already built/tested successfully before the review
  ran, and/or a two-line isolated `go run` repro settles it. Check both
  before spending remediation effort on a fix that isn't needed — and
  before assuming the bot is right just because it sounds specific and
  technical.
- When declining a review finding, cite the language-spec rule by exact
  wording plus at least one independent verification method (empirical
  build/test evidence or an isolated repro) in the reply — a bare "this is
  wrong" reply is weaker evidence for the PR record than a reply that
  shows the check.

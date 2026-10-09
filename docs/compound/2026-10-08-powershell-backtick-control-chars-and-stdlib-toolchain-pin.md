---
title: "PowerShell backtick escapes write hidden control characters, and stdlib advisories need a toolchain-pin hotfix"
date: 2026-10-08
tags: ["powershell", "markdown", "govulncheck", "go-toolchain", "stdlib-advisory", "ci", "backlogit", "git"]
shipment: 041-S
severity: medium
---

# PowerShell backtick escapes write hidden control characters, and stdlib advisories need a toolchain-pin hotfix

Shipment 041-S (feature 051-F, PR #116) was blocked at merge by Go standard-library advisories that `govulncheck`
reported on the CI runner, not by its own change. Clearing that blocker (hotfix PR #117) and closing the shipment
surfaced two reusable lessons.

## 1. Backtick escapes inside PowerShell double-quoted strings write control characters

In a PowerShell double-quoted string, a backtick is the escape character, so `` `b `` is U+0008 (backspace),
`` `n `` is a newline, and `` `t `` is a tab. A markdown code span written inside `"..."`, such as
`` `backlogit` ``, is silently corrupted: the first backtick plus `b` becomes one hidden U+0008 character and the
word loses its first letter in rendering and in literal search. Local tests never see it. Copilot review caught it
on PR #116.

Prevention:

* Author any markdown body with a single-quoted here-string (`@' ... '@`), which is literal, or with the .NET writer
  `[IO.File]::WriteAllText(path, text, (New-Object Text.UTF8Encoding $false))`.
* Before committing a generated document, scan it for control characters (code < 32, excluding tab, LF, and CR).
  For example: `[IO.File]::ReadAllText(path).ToCharArray() | Where-Object { [int]$_ -lt 32 -and [int]$_ -notin 9,10,13 }`.

## 2. Go stdlib advisories land after a green run, and the first remedy is the toolchain pin

The advisories GO-2026-6603, -6607, -6608, -6611, -6612, -6613, and -6617 were published after the last green
`security` run on `main`. CI used `go-version: '1.26.x'`, which resolved to go1.26.8 on the runner, so
`govulncheck` reported them as "Found in go1.26.8, Fixed in go1.26.9". The repository's decision note
(`docs/decisions/2026-09-04-go-toolchain-pin-maintenance-note.md`) says the first remedy is to raise the
`toolchain` line, not the `go` language floor. Its rule 4 also requires the inline `go.mod` comment to change with it.

Verification that held up:

* Check each advisory's fixed-in version on `vuln.go.dev` and in the CI log before choosing the target patch.
* Confirm the remedy in CI, not only locally. The `security` job log must show `go: downloading go1.26.9` and
  `No vulnerabilities found.`
* A patch-level toolchain bump is not behavior-neutral. It changes standard-library behavior, which is the purpose of a
  security fix, and it needs the usual regression run. What stays tied to the `go` directive is the language-version
  semantics and the version-keyed `GODEBUG` defaults, so the `go 1.24` line does not change with this kind of bump.

## 3. Branch switching with dirty backlog files

When the working tree has an uncommitted backlog file (here `.backlogit/stash.jsonl`) that differs between the
current branch and the target, `git checkout` aborts. Do not discard operator state to get past this. Preferred order:

* Stage a reversible copy first: `git stash push -- <path>` for a tracked file, or a byte-for-byte copy inside the
  gitignored `logs/` directory (keep backups in the workspace, not outside it), with SHA-256 hashes recorded.
* Get operator approval before any `git checkout --` or reset that discards a working-tree edit.
* Restore by line-set comparison (lines in the backup that are not in HEAD), so no entry is lost or duplicated.
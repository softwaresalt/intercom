---
title: "040-S red-first evidence log (PA-4)"
description: "Per-scenario red/characterization evidence for shipment 040-S (feature 050-F) with platform and parent SHA."
date: 2026-10-07
shipment: 040-S
feature: 050-F
---

# 040-S red-first evidence log (PA-4)

Platform: windows/amd64, go1.26.5 (go.mod `go 1.24`), git 2.55.0.windows.5.
Git Bash first on PATH for package-wide runs: `where bash` ->
`C:\Program Files\Git\bin\bash.exe` (first entry).
Code base for all units before U1: `main@f50dba4` (claim commit `4ad55eb`
touches only `.backlogit/`, so production code is identical).

## U1 (050.001-T) — parent `4ad55eb` (code == `f50dba4`)

Probe (before the test was written): `walkTable` desyncs on non-contiguous
siblings, depth-3 (ancestor and descendant token), interleaved
`[a.b]/[c]/[a]`, interleaved `[[t]]/[u]/[t.sub]`. It does NOT desync on the
plain ancestor reopen `[a.b]/[a]` (peekSelf) or the array-of-tables reopen
`[[t]]/[t.sub]/[[t]]` (walkArrayOfTables); per AC-1 those two rows moved to
the desync-free characterization table (scenario 2).

### Scenario 1 — RED

Command: `go test ./tools/gatecheck/internal/retiredarch/ -run TestScanTomlPrimary_DesyncShapes_UseFallback`
Result: FAIL, 7/7 rows. Every row passed the walkTable-desync precondition,
then failed with, e.g.:

```text
scanTomlPrimary(token doc) = [".../doc.toml: TOML parse error (fail-closed): retiredarch: TOML cursor desync: table at [a] has 1 unvisited field(s) but cursor has no matching entry left"], want [".../doc.toml: retired token 'channel_id' in TOML key path 'a.y.channel_id' (segment 'channel_id') (via sequence model)"]
```
### Scenario 2 / scenario 3 — harness stub step (G-5)

`walkDecodedTOML` did not exist at the parent, so scenarios 2-3 were added
with a no-op harness stub (`return nil`) to obtain assertion reds instead of
compile reds.

- Scenario 3 (AC-3) — RED:
  `go test ./tools/gatecheck/internal/retiredarch/ -run TestWalkDecodedTOML_UnexpectedType`
  -> `walkDecodedTOML error = <nil>, want an 'unexpected decoded type' error`
  (rows `go-int`, `string-slice`).
- Scenario 2 (AC-2, characterization) — against the stub every row reported
  `multiset mismatch` (walkTable findings vs empty); after implementation all
  10 rows pass (scalars incl. all four date/time kinds, inline table, array of
  inline tables, array of arrays, array-of-tables, dotted keys, composed
  header path, ancestor reopen, array-of-tables reopen, mixed array).

### Implementation — GREEN

Test-expectation correction during GREEN: row
`non-contiguous-sibling-header-token` (`[a.channel_id]` header with leaf `k`)
correctly yields two findings (`a.channel_id` and the composed
`a.channel_id.k`, segment `k`) — the same composed-path behaviour walkTable
has for `socket.mode.x`. The row's expectation was corrected; production
code was not changed for it.

### U1 completion gates (HEAD `7b58fdc`)

`gofmt -l .` empty (LF working tree, `*.go eol=lf`); `go vet ./...` 0;
`go build ./...` 0; `go test ./...` — all packages ok except one
environmental flake: `tests/integration`
`TestInvokeStartScriptMainPropagatesNonZeroExitCode` hit its 30s start.ps1
timeout because a pre-existing engram daemon (pid 40120, started 20:38
local) held the workspace lock. Re-run with `-count=1`: ok (25.3s).
Unrelated to retiredarch.

## U2 (050.002-T) — parent `HEAD` after U1 bookkeeping (code parent `7b58fdc`)

### Seam-only step (AC-4)

`scanTomlPrimary` became a wrapper over `scanTomlPrimaryWith(path, walk,
fallback)` with pass-through behaviour (no oracle). Existing tomlprimary
tests green: `go test ./tools/gatecheck/internal/retiredarch/ -run 'TestScanTomlPrimary|TestWalkDecoded'` ok.

### Scenarios 1-3 — RED (against the seam-only step)

Command: `go test ./tools/gatecheck/internal/retiredarch/ -run TestScanTomlPrimaryWith_CompletenessOracle`

- `walker-drops-a-finding`: got the single surviving cursor finding
  (`'channel_id'`), want the completeness finding.
- `walker-invents-a-finding`: got 3 findings (two real + invented `'acp'`),
  want the completeness finding.
- `fallback-errors`: got the two cursor findings, want
  `TOML parse error (fail-closed): injected fallback failure`.

### Oracle — GREEN

All three rows pass; the whole retiredarch package (golden/ordering tests
unchanged, so the oracle agrees with the cursor walk on every golden
fixture) passes.

## U3 (050.003-T) — component-wise Lstat containment

- Platform: windows/amd64, go1.26.5. Parent SHA: 7ff578f (U2 head). The red was run with an unwired `containedRegularFile` stub that returns (true, ""), so `runRepoScan` behaves exactly as at the parent.
- AC-1 `TestRunRepoScan_FinalComponentSymlink_FailsClosed`: RED, `Code = 0, want 1 (stderr="")`. os.Symlink succeeded on this host, so the row was observed rather than skipped.
- AC-2 `TestRunRepoScan_FinalComponentJunction_FailsClosed`: RED. The stderr held the parent's directory read-error text `...internal/x.go: read ...\internal\x.go: Incorrect function. (fail-closed synthetic finding)` instead of the U3 text.
- AC-3 `TestRunRepoScan_IntermediateLink_FailsClosed` (junction made with `cmd /c mklink /J`): RED, `Code = 0, want 1 (stderr="")`.
- AC-4 `TestContainedRegularFile_RejectsMalformedComponents`: RED. Every row returned (true, ""): `a//b.go`, `./a.go`, `../x.go`, `a/`, the empty string, and on Windows `a\b.go`, `c:x.go`, `a/b:c.go`.
- AC-4 characterization `TestRunRepoScan_DeletedFile_KeepsReadErrorText`: PASSES at the parent, as expected. It pins the byte-for-byte read-error text.
- `TestContainedRegularFile_AcceptsRegularAndMissing`: PASSES at the parent. It pins the ok=true outcomes.
- AC-4b `TestContainedRegularFile_UnreadableIntermediate_FailsClosed`: SKIPPED on Windows by design (POSIX permission bits only). It runs on the Linux CI runner, so it carries HEAD-only evidence (PA-4).

### U3 gates
- Full gates at the U3 head: gofmt -l . is empty, vet=0, build=0, test=0. Git Bash is first on PATH.
- No tracked symlinks fall inside the scan scope (`git ls-files -s` shows no 120000 entries), so a repo scan of the real checkout is unaffected.

## U4+U5 (050.004-T) PA-7 inventory precheck: MATCH

- `git grep -n -E 'GIT_[A-Z_]+|DefaultGitRunner|gitShowToplevel'` over `tools/gatecheck` and `scripts/check-retired-architecture.sh` returns 78 lines at HEAD and 78 lines at Stage's read base main@372ab38. `Compare-Object` finds zero differences, including line numbers.
- Classification:
  - `check_retired_architecture_wrapper_test.go:228` sets GIT_CEILING_DIRECTORIES → kept (allow-listed).
  - `ciwiring_retire_test.go:313-335` and `unignore/testutil_test.go:44-99` build their own child envs → unaffected.
  - The retiredarch tests (`retiredarch_test.go`, `selftest_selection_test.go`, `select_test.go`) call DefaultGitRunner on the real checkout and set no GIT_* → dropped-safe.
- Lines that matched but are not setters or callers:
  - `pin_test.go:425` is a static reject-row fixture string in `TestCheckPathspecPin_PackageClosure_RejectTable`. It is parsed, never executed.
  - `unignore/git.go:42` is a comment.
  - `unignore` and `writepath` each declare their own distinct DefaultGitRunner symbols, which this change does not touch.
  - `register_retired_arch.go:26` is the production wiring, i.e. the subject of the fix. It sets no GIT_*.
  - None of these differ from Stage's read, so there is no deviation and no HALT.
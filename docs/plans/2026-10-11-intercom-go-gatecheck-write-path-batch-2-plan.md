---
title: "Gatecheck write-path Batch 2 implementation plan (two release units)"
description: "Plan for Batch 2 (stash EEE39D23, DBE25DF5, 858516B9, EDC18D59, 2613D248): RU-1 write-path gate completeness and CI-log output safety (shipment S1); RU-2 gate git discovery and environment isolation (shipment S2); sequence 043-S, then S1, then S2"
date: 2026-10-11
source: "docs/decisions/2026-10-11-intercom-go-gatecheck-write-path-batch-2-deliberation.md"
requires_plan_hardening: true
tags:
  - "gatecheck"
  - "writepath"
  - "retiredarch"
  - "fail-closed"
---

# Plan: Gatecheck write-path Batch 2

Source: `docs/decisions/2026-10-11-intercom-go-gatecheck-write-path-batch-2-deliberation.md`
(decisions **D-WB2-1..D-WB2-10**). Pattern and prerequisite: the shipped Batch B1 plan
`docs/plans/2026-10-10-intercom-go-gatecheck-batch-b1-write-path-input-hardening-plan.md`
(shipment 042-S, feature 052-F, PR #123). Consumed stash entries: `EEE39D23`, `DBE25DF5`,
`858516B9`, `EDC18D59`, `2613D248`. Newly captured and left active: `36BE1F61`, `1B85D0F4`,
`A8FB0682`.

## Problem Frame

The write-path gate (`tools/gatecheck/internal/writepath`, wrapper
`scripts/check-write-path-precondition.sh`) is a merge-blocking, fail-closed control that proves no
destructive write primitive exists under `internal/**` and `cmd/**`. After B1 it reads a
NUL-safe listing through an isolated git child. Five residuals remain, each a way the gate can
report clean without having checked what it claims to check, or leak attacker-chosen text into the
CI log:

1. **Scan set (EEE39D23):** `scannedPathsFromListing` keeps only `shouldScan` records, so a tracked
   symlink or gitlink with no `.go` suffix (verified: `git ls-files` lists `internal/fx` and
   `internal/sub`) is never checked.
2. **Root discovery (DBE25DF5):** both wrappers derive `ROOT` through unscrubbed
   `git rev-parse --show-toplevel` (verified: ambient `GIT_DIR`+`GIT_WORK_TREE` returns the decoy
   root), and ED-7 detects only an empty selection.
3. **Fixture self-test (858516B9):** a real directory named `*.go` is skipped silently;
   `scripts/testdata/writepath/harness/` is a legitimate directory that must keep working.
4. **Output (EDC18D59):** `errorLine`, the fixture PASS/FAIL prints and the unwrapped
   `*fs.PathError` from `pysem.ReadText` echo attacker-influenced text unescaped.
5. **Env twin (2613D248):** `gitRunnerEnv` (writepath and retiredarch copies) does not strip the
   MSYS runtime variables; boundary rows are missing from both twin tables.

## Release units and sequence (D-WB2-1)

| Release unit | Feature (to be created) | Shipment | Entries | Surface |
|---|---|---|---|---|
| **RU-1** | Write-path gate completeness and CI-log output safety | **S1** | EDC18D59, EEE39D23, 858516B9 | `tools/gatecheck/internal/writepath` only |
| **RU-2** | Gate git discovery and environment isolation across wrappers and engines | **S2** | DBE25DF5, 2613D248 | shell wrappers and shared lib, writepath env twin, retiredarch frozen pin |

**Chain:** `043-S` (Batch 1, queued, on the Batch 1 staging branch) → **S1** → **S2**, each with a
`blocks` edge (`backlogit dep add <later> <earlier>`). **Merge order:** the Batch 1 staging PR
merges first (as a merge commit, never squash), then this branch's staging PR.

**File overlap with 043-S: none** (043-S: `.github/**`, `.autoharness/**`, policy docs,
`tests/integration/*` contract tests; Batch 2: `tools/gatecheck/**`, `scripts/**`). Both must leave
`go test ./...` green. Within Batch 2, S1 and S2 both edit `writepath.go` and `writepath_test.go`
(S2 only `gitRunnerEnv` and `TestGitRunnerEnv`), so the chain is required for ordering, not only
priority.

## Requirements Trace

| Requirement (source) | Units |
|---|---|
| EDC18D59: central escaping at `errorLine` and fixture prints; unwrap `*fs.PathError`; `!strconv.IsPrint` for C1; cover the `not a contained regular file: <reason>` sink | U1-U4 |
| EEE39D23: every tracked record under `internal/**`, `cmd/**` passes containment or fails closed; decide case-variant scope | U5, U6 (case variant: D-WB2-3, pinned in U5) |
| 858516B9: `.go`-named directory fails closed; FIFO row without a Windows skip; hard link (decided not implemented, D-WB2-7) | U7, U8 |
| DBE25DF5: deliberate root discovery for both wrappers; wrong-tree coverage with `internal/` and `cmd/` in both repos (AS-5); per-arm guard rejected (D-WB2-5) | U9-U12 |
| 2613D248: strip `MSYS`, `CYGWIN`, `MSYS2_*`; boundary rows in both tables, identical; retiredarch pin atomicity | U13-U15 |

## Contract: repo-scan outcome matrix after RU-1 (the 042-S matrix plus one new outcome)

Evaluated in this order; the first failure wins; every echoed string passes `escapeLogText`.

| # | Condition | Result |
|---|---|---|
| 1 | git runner error | `::error::git ls-files: <err>` exit 1 (unchanged text for control-free input) |
| 2 | listing not NUL-terminated or has an empty record | `::error::git ls-files output: ...` exit 1 (unchanged) |
| 3 | a **selected** path is invalid UTF-8 or has a forbidden rune | `::error::git ls-files output: path "..." ...` exit 1 (unchanged) |
| **4 (new)** | a listed record that `shouldScan` did **not** select fails `containedRegularFile` (symlink, gitlink or other directory, symlinked ancestor, non-regular), checked in git order | `::error::<rel>: not a contained regular file: <reason>` exit 1, before ED-7 and before any selected file is read |
| 5 | no selected path (ED-7) | unchanged, exit 1 |
| 6 | per selected path: containment refusal, read error, finding | unchanged ordering and text (read error now has the `*fs.PathError` unwrapped) |
| 7 | clean | exit 0 |

A tracked non-selected record absent from the working tree keeps today's `fs.ErrNotExist` pass (a
deleted non-Go file is not built). Case-variant top-level directories (`Internal/`, `CMD/`) are
not selected and are not an error (D-WB2-3). Fixture self-test adds one outcome: a directory whose
name ends in `.go` fails closed.

## Implementation Units

Conventions: every task has both `size` (effort) and `complexity` (difficulty), set independently.
Every test unit names its red evidence (the parent-commit failure). Posture for a pair is
test-first: the test unit lands red, the production unit turns it green. Windows: filesystem
symlink rows reuse the existing `u4CreateSymlink` helper (skips only on Errno 1314 under
`GOOS=windows`, red evidence comes from the required Linux job); rows that need names Windows
cannot create live in `//go:build !windows` files, so no new `t.Skip` is added.

### RU-1: Write-path gate completeness and CI-log output safety (shipment S1)

**U1 (test; size S, complexity low): escape helper and `errorLine` rows.**
File: `tools/gatecheck/internal/writepath/writepath_test.go`. Scenarios (2): (a) a pure table for
`escapeLogText`: control characters (LF, CR, TAB, ESC), DEL, C1 including U+009B, U+0085, U+2028,
U+2029, a bidi override (U+202E), a format character (U+200B), an invalid UTF-8 byte, a lone
backslash, printable ASCII and printable non-ASCII (`é`, a CJK rune): non-printables become
`\xNN` / `\uNNNN` / `\UNNNNNNNN`, everything else is byte-identical; (b) `errorLine` rows: benign
text identical to today's golden text; context and error text with a newline followed by
`::error::x` produce exactly one output line. **Red:** `escapeLogText` is undefined (compile red)
and `errorLine` prints the raw newline. AC: both scenarios fail at the parent for those reasons.

**U2 (prod; size S, complexity low): `escapeLogText` and `errorLine`.** Blocked by U1.
File: `writepath.go`. Add `escapeLogText(string) string` (iterate with `utf8.DecodeRuneInString`;
a width-1 `RuneError` becomes `\xNN`; `!strconv.IsPrint(r)` becomes a visible escape; backslash is
**not** escaped). `errorLine` escapes `context` and `err.Error()`. AC: U1 green; every existing
golden and `TestRunRepoScan_BadListing_FailsClosed` row unchanged; `go vet` clean.

**U3 (test; size M, complexity medium): call-site sink rows.** Blocked by U2.
Files: `writepath_test.go`, new `writepath_unix_test.go` (`//go:build !windows`). Scenarios (2):
(a) repo-scan sinks as one table: a stub runner returning an error whose text contains a newline and
`::error::` stays on one line; a read error echoes **no absolute path** (the `*fs.PathError` is
unwrapped); update the pinned expectation of `TestRunRepoScan_DeletedTrackedFile_KeepsReadError`,
which today recomputes `errorLine(rel, readErr)` from the raw `*fs.PathError`, to the unwrapped
form; (b) fixture-self-test sinks (unix-only file): a fixture named with an embedded newline and
`::error::` text yields PASS/FAIL lines and a U4 `not a contained regular file` line that contain no
raw control character (`assertNoRawControl`). **Red:** PASS/FAIL and the read error currently echo
raw text and the absolute path. AC: the existing U4 linked-fixture rows (benign names) stay green.

**U4 (prod; size S, complexity low): apply the chokepoint at the sinks.** Blocked by U3.
File: `writepath.go`. `runFixtureSelfTest`: escape the name in `PASS`, `FAIL` and the joined findings;
add `unwrapPathError(error) error` and use it where `scanFile`'s error reaches `errorLine` in
`runRepoScan` and `runFixtureSelfTest`. AC: U3 green; goldens unchanged;
`TestRun_ReadError_FailsClosed` still names the relative path.

**U5 (test; size M, complexity medium): scan-set widening rows.** Blocked by U4.
File: `writepath_test.go`. Scenarios (3): (a) `runRepoScan` table over real files with a stub
runner: a non-`.go` symlink record, a `_test.go` symlink, a symlink to a directory, a symlinked
ancestor of a non-Go file, and a gitlink (real directory plus listed record) each fail with
`not a contained regular file`; controls: a regular non-Go file, a regular `_test.go`, and a record
absent from disk pass; (b) a pure row pinning the D-WB2-3 decision: `Internal/x.go` and `CMD/y.go`
listing records are neither selected nor an error; (c) one real-git end-to-end row
(`hermeticGitEnv`; `git update-index --cacheinfo 160000,...` for `internal/sub` plus a real
directory) proving the pathspec returns the gitlink record and the gate refuses it. **Red:** every
(a) refusal row and (c) pass at the parent (silent drop), controls and (b) green on arrival
(characterization, stated in the test). AC: red evidence for (a) and (c) captured on the Linux job.

**U6 (prod; size M, complexity medium): contain every non-selected record.** Blocked by U5.
File: `writepath.go`. Add `listedRecords(out []byte) ([]string, error)` (the NUL/empty-record
parse); `scannedPathsFromListing` keeps its signature and composes it; `runRepoScan` runs
`containedRegularFile` over the non-selected records in git order after the selected-path rules and
before ED-7; the in-loop check for selected records is unchanged. Update the doc comments
(`runRepoScan`, `scannedPathsFromListing`) with the outcome-4 text, the case-variant decision and
the hard-link accepted residual (`A8FB0682`). AC: U5 green; `TestScannedPathsFromListing_Table` and
`TestRunRepoScan_BadListing_FailsClosed` unchanged and green; `TestRun_RepoMode_CurrentTreeIsClean`
green on the real tree (56 records, none non-regular).

**U7 (test; size S, complexity low): fixture directory and FIFO rows.** Blocked by U6.
Files: `writepath_test.go`, `writepath_unix_test.go`. Scenarios (2): (a) a table: real directories
named `accept-d.go`, `reject-d.go` and `other.go` fail closed with the escaped name; a directory
named `harness` (no `.go` suffix) next to a valid fixture is still skipped and the self-test
passes; (b) unix-only FIFO named `reject-fifo.go` (via `syscall.Mkfifo`) is refused as not a regular
file, run in a goroutine with a timeout so a regression fails instead of hanging. **Red:** (a)
directory rows pass silently at the parent; (b) is a characterization row (green on arrival, stated).

**U8 (prod; size S, complexity low): refuse `.go`-named directories.** Blocked by U7.
File: `writepath.go`. In `runFixtureSelfTest` replace the unconditional `e.IsDir()` skip with: a
directory whose name ends in `.go` returns `errorLine(relPath, errors.New("fixture is a directory; names ending in .go must be regular files"))`
exit 1; other directories stay skipped. Comment records the hard-link residual. AC: U7 green;
`TestRun_SelfTest_MatchesGolden` and `TestRun_SelfTestIntegrity_MatchesGolden` unchanged.

### RU-2: Gate git discovery and environment isolation (shipment S2)

**U9 (test; size M, complexity medium): write-path wrapper wrong-tree rows.**
File: `tools/gatecheck/check_write_path_wrapper_test.go` (extend the go shim in this file to record
the `--root` argument it hands the generated binary; the shim is package-shared). Scenarios (3):
(a) two repositories that both contain `internal/` and `cmd/` files; ambient `GIT_DIR` and
`GIT_WORK_TREE` point at the decoy; the recorded `--root` must be the wrapper's own checkout
(`os.SameFile`, Windows 8.3 safe); (b) `GIT_DIR` alone with cwd in a subdirectory of the real tree;
(c) a no-ambient control. **Red:** (a) and (b) record the decoy or subdirectory at the parent.
Follow the existing `requireBash` / `runWritePathWrapperWithTrace` harness.

**U10 (prod; size S, complexity medium): `gatecheck_repo_root` and the write-path wrapper.** Blocked by U9.
Files: `scripts/lib/gatecheck-run.sh`, `scripts/check-write-path-precondition.sh`. Add
`gatecheck_repo_root`: run `git -C "${GATECHECK_SRC}" rev-parse --show-toplevel` with every
exported variable whose ASCII-uppercased name starts with `GIT_` unset, except
`GIT_CEILING_DIRECTORIES` (mirrors Go `gitRunnerEnv`); fail with git's message and its exit code.
The wrapper sets `ROOT="$(gatecheck_repo_root)"` and keeps a separate caller-cwd
`git rev-parse --show-toplevel >/dev/null` precondition (exit 128 outside a work tree), as the
retiredarch wrapper does. AC: U9 green; `TestGatecheckRunSh_ShellcheckCleanIfAvailable`,
`TestWritePathWrapperText_PureBashDispatch` and the runner tests green; header comment states the
anchoring change (D-WB2-4).

**U11 (test; size S, complexity medium): retiredarch wrapper wrong-tree rows.** Blocked by U10.
File: `tools/gatecheck/check_retired_architecture_wrapper_test.go`. Same three scenarios as U9
against `scripts/check-retired-architecture.sh`. **Red:** (a) the `git -C GATECHECK_SRC` form returns
the decoy under ambient variables (verified by probe).

**U12 (prod; size S, complexity low): retiredarch wrapper uses the shared function.** Blocked by U10, U11.
File: `scripts/check-retired-architecture.sh`. Replace line 198 with `ROOT="$(gatecheck_repo_root)"`;
keep the line-192 caller-cwd precondition and its comment. AC: U11 green; the
`TestRetiredArchWrapperText_*` pins and `TestCheckRetiredArchitectureWrapper_FailsOutsideAGitWorkTree`
green (read the pinned phrases first; amend the comment, never the pinned text, unless a pin names
the replaced line).

**U13 (test; size S, complexity low): twin env-table rows.** Blocked by U12 (sequence only).
Files: `writepath_test.go`, `retiredarch/select_test.go` (`TestGitRunnerEnv`, **identical rows in
both**). Scenarios (2, one per package, same table): strip rows `MSYS=1`, `msys=1`, `CYGWIN=x`,
`Msys2_Arg_Conv_Excl=*`; keep rows `MSYSTEM=MINGW64`, `MSYSX=1`, `MSYS2=1`, `CYGWINX=1`; the three
boundary rows from the entry: bare `GIT_DIR` (no `=`), exact `GIT_`, bare `GIT_CEILING_DIRECTORIES`.
**Red:** the strip rows fail in both packages; the keep rows and the three boundary rows are green
on arrival (characterization, stated).

**U14 (prod; size S, complexity low): writepath `gitRunnerEnv`.** Blocked by U13.
File: `writepath.go`. Add the MSYS rule after the `GIT_*` rule; update the doc comment (copy, keep
identical with retiredarch). AC: writepath half of U13 green.

**U15 (prod, atomic; size S, complexity medium): retiredarch `gitRunnerEnv` and its pin.** Blocked by U14.
Files: `retiredarch/select.go` and `retiredarch/pin.go` **in one commit**: the
`gitRunnerEnv` body in `select.go` and the hand-authored `canonicalDecls` copy in `pin.go`
(token-equal). `closedWorldDecls`, `pathspecFrozenDecls` and `prefixFrozenDecls` name declarations,
not text, and do not change. AC: retiredarch half of U13 green; `pin_test.go` green including
`TestCheckPathspecPin_GitRunnerIsolation_Rejected`; `git show --stat` of the commit lists exactly
those two files; the two `gitRunnerEnv` bodies are textually identical to the writepath copy.

## Dependency Graph

```text
043-S ─blocks→ S1:  U1 → U2 → U3 → U4 → U5 → U6 → U7 → U8
S1 ─blocks→ S2:     U9 → U10 → U11 → U12 → U13 → U14 → U15
```

All task-level edges are single-predecessor chains (no cycle). The strictly linear order is
intentional: RU-1 tasks all edit `writepath.go`/`writepath_test.go`; U13-U15 share the twin table.
U9-U12 (shell) and U13-U15 (env) could run in parallel across files, but Ship executes serially, so
the chain adds no cost.

## Decisions and Rationale

See the deliberation: D-WB2-1 (two shipments), D-WB2-2 (contain every non-selected record), D-WB2-3
(case variants out of scope), D-WB2-4 (script-anchored scrubbed root through one shared function),
D-WB2-5 (per-arm ED-7 guard rejected), D-WB2-6 (one local escaping chokepoint, `!strconv.IsPrint`),
D-WB2-7 (directory refusal, FIFO as a build-tag row, hard link not implemented), D-WB2-8 (MSYS
strip, twin tables, atomic refreeze), D-WB2-9 (three new deferred entries), D-WB2-10 (no overlap with
043-S).

## Risks and Caveats

| Risk | Mitigation |
|---|---|
| A new fail-closed outcome blocks every PR if a legitimate tracked symlink or gitlink exists | Verified none exists today (56 records, modes 100644/100755 only); U6 AC runs `TestRun_RepoMode_CurrentTreeIsClean`. Rollback: revert the U6 commit. |
| Escaping changes a golden | Escape only touches non-printable runes; U2 and U4 ACs require every golden byte-identical. |
| Wrapper anchoring changes behaviour for callers outside the checkout | Documented (D-WB2-4, operator checkpoint 3); caller-cwd precondition and exit 128 preserved; engine and tree now come from the same checkout. |
| Shell scrub is not portable (Git Bash, dash-less shells) | `env -u` is already used in the lib; `compgen -e` is a bash builtin (wrappers are bash). Fallback in D-WB2-4 option C. U9/U11 run under `requireBash`, including Windows Git Bash in the advisory job. |
| `pin.go` and `select.go` drift | U15 is a single atomic commit with a stat assertion; red evidence is the pin test itself. |
| Windows cannot create symlinks or control-char names | Existing Errno 1314 skip for symlinks only; unix-only rows are build-tagged, not skipped; Linux CI is the required evidence. |
| Review-cycle fatigue (B1 hit the cap) | Two shipments with disjoint surfaces (D-WB2-1); each unit is one reviewable diff. |

## Plan Hardening Signals

* **Public API, schema or contract change:** present. A new fail-closed outcome (matrix row 4), a new
  fixture refusal, a wrapper root-anchoring change.
* **Security, auth, permission or compliance-sensitive behaviour:** present. Gate integrity
  (false-clean classes) and CI workflow-command injection.
* **Migration, backfill, destructive data or irreversible step:** absent. All changes are
  source and test edits, revertable by `git revert`.
* **External integration, operator checkpoint or external dependency:** present. Merge order behind
  the Batch 1 staging PR and 043-S; six operator checkpoints (deliberation).
* **High runtime, rollout or rollback risk:** present at moderate level. A false-positive refusal
  blocks all PRs until reverted; no data is at risk.

**Requires plan hardening: yes**

## Runtime Verification and Closure

* **Runtime surfaces changed:** the `gatecheck write-path` and `gatecheck retired-arch` CLI via
  their wrappers (CI steps, pre-push gates). No service, API or UI.
* **Ship verification per unit (Stage does not run these):** `go test ./tools/gatecheck/...`
  (race on the Linux job), `go vet ./...`, the full `go test ./...`, shellcheck on the touched
  scripts, `scripts/check-write-path-precondition.sh --self-test` and
  `scripts/check-retired-architecture.sh --self-test` on the real tree (both exit 0).
* **Operational closure:** each shipment closes through the post-043-S single close procedure;
  closure record lists the accepted residuals (hard link `A8FB0682`, retiredarch echo `36BE1F61`,
  other wrappers `1B85D0F4`, case variants/module scope, TOCTOU) with stash IDs.

## Plan Hardening

**Hardening required: yes.** Triggers: contract change (new fail-closed outcome, wrapper root
anchoring), security-sensitive gate integrity and CI log injection, operator checkpoints and a
cross-shipment merge-order dependency.

### Protected invariants (must hold after every unit)

1. **Fail closed:** no path may turn a refusal into a pass; no new `continue`/skip on a record that
   can be read.
2. **Goldens are frozen:** every existing golden, `TestScannedPathsFromListing_Table` and
   `TestRunRepoScan_BadListing_FailsClosed` row is byte-identical after RU-1; only the escape-aware
   rows and the one pinned `DeletedTrackedFile` expectation (U3) change.
3. **Error precedence does not move** for selected paths (matrix rows 1-3, 5-7).
4. **Twin parity:** the two `gitRunnerEnv` bodies are textually identical; the retiredarch pin and
   `select.go` change in one commit.
5. **Isolation intent parity:** the shell scrub and Go `gitRunnerEnv` use the same rule (ASCII-fold,
   `GIT_` prefix, keep `GIT_CEILING_DIRECTORIES`).
6. **No Stage-written source:** this plan authors no source; Ship writes all code under the units.

### Learnings and instructions consulted

* `docs/compound/workflow-issues/staging-verification-commands-and-ambient-git-env-2026-10-10.md`:
  compare bytes or the passed `--root`, never exit codes; probe from a scrubbed environment first;
  record scrubbed and unscrubbed results. Applied to U9, U11 and the red-evidence capture.
* `docs/compound/2026-10-07-pin-canonical-text-maintenance-and-unbounded-ast-rule-review.md`:
  atomic refreeze, `git show --stat`, bound the threat model. Applied to U15 and D-WB2-7.
* `docs/compound/2026-10-01-textual-yaml-guard-denylist-bypass-loop.md`: shape rules over named
  spellings. Applied to U6 (every non-selected record) and U8 (any `.go`-named directory).
* `docs/compound/2026-09-30-go-msys-bash-bridging-and-unicode-case-mapping-gap.md`: compare by file
  identity, not string, under Git Bash. Applied to U9/U11 (`os.SameFile`).
* Instructions: `constitution.instructions.md` (careful mode for gate-integrity code, Principle VIII
  as cited in 043-S), `concurrency.instructions.md` (single branch, no extra worktree).

### Hardening additions found while reading the sources

* **bash 3.2 and `set -u`:** U10's function must not expand an empty array under `set -u` (macOS
  bash 3.2 errors); use the `${arr[@]+"${arr[@]}"}` form or build the `env -u` argument list in a
  way that is never empty, and keep `TestGatecheckRunSh_ShellcheckCleanIfAvailable` green.
* **`set -e` in a command substitution:** `ROOT="$(gatecheck_repo_root)"` must fail the wrapper
  with git's exit code when git fails, as today (U9 control row asserts a non-git root still
  fails, not a silent empty `--root`).
* **Windows case-insensitive env names:** the scrub uppercases names before comparing, so `git_dir`
  and `Git_Work_Tree` are removed; a U9 row sets a mixed-case ambient name.
* **U5(c) real-git row:** use `hermeticGitEnv` so ambient developer variables cannot flip the red
  evidence (B1 lesson); assert listing bytes, not only exit code.
* **U7(b) FIFO:** a regression would block forever in `os.ReadFile`; the goroutine-and-timeout
  wrapper is part of the AC, not optional.
* **Echo of non-selected record names (U6):** these names are not rune-validated by design; U2/U4
  must already be merged so the refusal line is born escaped. The chain enforces that order.
* **U13 keep-rows:** `MSYSTEM`, `MSYSX`, `MSYS2` (no underscore), `CYGWINX` pin the exact-name and
  prefix boundary so a later widening is deliberate.

### Risky actions (`ProposedAction` / `ActionRisk` / `ActionResult`)

| ID | ProposedAction | ActionRisk | Approval | Expected ActionResult and rollback |
|---|---|---|---|---|
| PA-1 | U6: add the matrix row-4 refusal to the merge-blocking gate | medium (false positive blocks all PRs) | operator review of the staging PR | Result: gate stays green on the real tree. Rollback: revert the U6 commit; U5 rows go red, so the revert must include U5 (record as a pair). |
| PA-2 | U10/U12: change wrapper root anchoring from caller-cwd to script location | medium (behaviour change) | operator checkpoint 3 | Result: same `--root` for every in-checkout invocation. Rollback: revert U10/U12 together with U9/U11. |
| PA-3 | U15: refreeze retiredarch `gitRunnerEnv` and `canonicalDecls` | medium (pin gate integrity) | none beyond review | Result: pin test green. Rollback: revert the single commit. |
| PA-4 | Archive the five consumed stash entries after shipments are assembled | low (reversible: archive, not delete) | none | Result: entries retrievable from the archive with a disposition note. |
| PA-5 | Create two features, 15 tasks, two shipments, edges `043-S ← S1 ← S2` | low | none | Rollback: `backlogit` archive of the new items before Ship claims. |

### Verification, monitoring and closure additions

* **Prechecks Ship must record per task:** `git --version` is 2.32 or later; `go version` matches
  the repo pin; bash available (`requireBash`); on the Linux job the symlink rows are not skipped.
* **Red evidence is mandatory** for U1, U3, U5(a)(c), U7(a), U9(a)(b), U11(a), U13 strip rows: the
  failing output at the parent commit is attached to the task, and characterization rows are
  labelled as green-on-arrival so nobody reads them as red evidence.
* **Monitoring signal after merge:** the CI `write-path` and `retired-arch` steps on the next three
  PRs; a new refusal on an unmodified tree is the rollback trigger.
* **Validation window:** one full CI cycle on `main` after each shipment merges; owner: Ship's
  post-merge closure for that shipment.
* **Operator checkpoints:** (1) two shipments; (2) case variants out of scope; (3) wrapper
  anchoring change; (4) hard link not implemented; (5) per-arm ED-7 guard rejected; (6) stacked
  base and merge order. Defaults are the decisions as written; an override at staging-PR review
  changes the plan before Ship claims S1.

### Review-gate capability risks

Plan review must emit literal `dispatch_mode:` and `decision:` markers. Reviewer subagent dispatch
is probed first (P-012); if cross-model routing is unavailable the same rubric runs same-model with
a declared degradation. The Security Lens persona is triggered (CI log injection, gate integrity,
trust boundary at the wrapper environment) and the Architecture Strategist is triggered (module
boundary: shared shell function, copied-not-shared Go helper). A plan that reaches the gate without
this section must FAIL.

### Unresolved operator decisions that still block safe execution

None block staging. Operator checkpoints 1-6 are defaults that stand unless overridden at PR review;
checkpoint 6 is a hard sequencing requirement (merge the Batch 1 staging PR first).

<!-- plan-review-attempt: 0 -->


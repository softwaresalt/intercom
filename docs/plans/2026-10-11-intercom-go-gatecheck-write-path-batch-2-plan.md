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
`A8FB0682`, `284270E8`, `49982A4E`.

## Problem Frame

The write-path gate (`tools/gatecheck/internal/writepath`, wrapper
`scripts/check-write-path-precondition.sh`) is a merge-blocking, fail-closed control that proves no
destructive write primitive exists under `internal/**` and `cmd/**`. After B1 it reads a
NUL-safe listing through an isolated git child. Five residuals remain, each a way the gate can
report clean without having checked what it claims to check, or leak attacker-chosen text into the
CI log:

1. **Scan set (EEE39D23):** `scannedPathsFromListing` keeps only `shouldScan` records, so a tracked
   symlink or gitlink with no `.go` suffix (verified: `git ls-files` lists `internal/fx` and
   `internal/sub`) is never checked. **Plan-review SEC-1, verified by probe:** the pathspecs
   `internal/**` and `cmd/**` do not list a symlink or gitlink named exactly `internal` or `cmd`
   (the literal pathspecs `internal` and `cmd` do), so a top-level `cmd -> ../pkg` symlink is never
   judged either. U6 therefore also changes the runner's pathspecs to the literal `internal` and
   `cmd`.
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
merges first (as a merge commit, never squash; Principle XI), then this branch's staging PR.

**Why the `043-S` edge exists although the file sets are disjoint (SB-2):** it is an
operator-directed ordering, not a file conflict. 043-S installs the single authoritative
shipment-close procedure that Ship must use to close S1 and S2, so Ship should not claim S1 before
043-S has shipped. **File overlap with 043-S: none** (043-S: `.github/**`, `.autoharness/**`,
policy docs, `tests/integration/*` contract tests; Batch 2: `tools/gatecheck/**`, `scripts/**`);
both must leave `go test ./...` green.

**Why two shipments (SB-1):** a review-size cap and gate-risk isolation, not disjoint files. S1 is
one package; S2 mixes shell, a shared lib and a frozen pin. The S1 → S2 edge is needed only
because U13/U14 edit `writepath.go` and `writepath_test.go` after S1 does; U9-U12 (shell) are
independent of S1 and are chained for serial execution only.

## Requirements Trace

| Requirement (source) | Units |
|---|---|
| EDC18D59: central escaping at `errorLine` and fixture prints; unwrap `*fs.PathError`; `!strconv.IsPrint` for C1; cover the `not a contained regular file: <reason>` sink | U1-U4 |
| EEE39D23: every tracked record under `internal/**`, `cmd/**` passes containment or fails closed (including a top-level `internal`/`cmd` link); decide case-variant scope | U5, U5b, U6 (case variant: D-WB2-3, pinned in U5) |
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
| **4 (new)** | a listed record that `shouldScan` did **not** select fails `containedRegularFile` (symlink, gitlink or other directory, symlinked ancestor, non-regular, **including a record named exactly `internal` or `cmd`**), checked in git order | `::error::<rel>: not a contained regular file: <reason>` exit 1, before ED-7 and before any selected file is read |
| 5 | no selected path (ED-7) | unchanged, exit 1 |
| 6 | per selected path: containment refusal, read error, finding | unchanged ordering and text (read error now has the `*fs.PathError` unwrapped) |
| 7 | clean | exit 0 |

A tracked non-selected record absent from the working tree keeps today's `fs.ErrNotExist` pass (a
deleted non-Go file is not built; `ls-files -z` carries no mode, so a symlink or gitlink record
whose path is absent from disk also passes, an accepted residual, G-9). Case-variant top-level
directories (`Internal/`, `CMD/`) are not selected and are not an error (D-WB2-3). Fixture
self-test adds one outcome: a directory whose name ends in `.go` fails closed (raised during
collection, before the per-fixture containment checks).

**Echo completeness (SEC-3, G-7):** "every echoed string" means: `errorLine`, the repo-scan
findings loop, the fixture `PASS`/`FAIL` lines and joined findings, and the two fixture-directory
`...not found` / `no fixtures discovered` lines (which carry `--root`). The shell
`unrecognized argument: $arg` echo carries workflow-YAML-supplied text (trusted) and stays an
accepted residual.

## Implementation Units

Conventions: every task has both `size` (effort) and `complexity` (difficulty), set independently.
Posture for a pair is test-first: the **test unit lands red and records the failing output at the
parent commit; it compiles and passes `go vet`, only its new assertions fail** (a compiling stub of
the production symbol is part of the test unit when the symbol does not exist yet, as in U1), and
the **production unit that follows restores green**. A full `go test ./...` green is required at
each pair completion (U2, U4, U6, U8, U10, U12, U15), not at the test units, whose assertions are
red by design (SB-4). Every fixture is created under `t.TempDir()` (test-runtime temporary
directories managed and cleaned by `testing`); no agent file operation targets a path outside the
cwd (Principles III and IV). Red evidence is captured by running the test unit's tests at the test
commit, before the paired production commit lands (no second worktree). Windows:
filesystem symlink rows reuse the existing `u4CreateSymlink` helper (skips only on Errno 1314 under
`GOOS=windows`; red evidence comes from the required Linux job); rows that need names or objects
Windows cannot create live in `//go:build linux || darwin` files (not `!windows`:
`syscall.Mkfifo` is absent on plan9, js and wasip1), so no new `t.Skip` is added, helpers used only
by those files live in those files, and verification adds `GOOS=linux` and `GOOS=darwin`
`go vet ./tools/gatecheck/...` (G-6, R2-4). Commit
convention: `test(...)` for test units, `fix(...)`/`feat(...)` for production units (CR-6).

**Safety modes (Principle VIII, CR-2, RC-3):** careful mode for U6, U10, U12, U14 and U15 (a
merge-blocking gate, a wrapper behaviour change and the twin env filter plus a pin refreeze:
enumerate the risks above, confirm the outcome-matrix and the `PA-*` rollback before each commit); freeze-scope mode bounded to
`tools/gatecheck/**` and `scripts/**` for every unit. The Ship agent invokes the `safety-modes`
skill at the start of each of those units.

### RU-1: Write-path gate completeness and CI-log output safety (shipment S1)

**U1 (test; size S, complexity low): escape helper and `errorLine` rows.**
Files: `tools/gatecheck/internal/writepath/writepath_test.go` and a **compiling stub**
`escapeLogText(s string) string { return s }` in `writepath.go` (harness scaffold so the test
commit builds and passes `go vet`, RC-1; U2 replaces the body). Scenarios (2): (a) a pure table for
`escapeLogText`: LF, CR, TAB, ESC, DEL, C1 including U+009B, U+0085, U+2028, U+2029, a bidi override
(U+202E), a zero-width format character (U+200B), an invalid UTF-8 byte, a lone backslash, printable
ASCII and printable non-ASCII (`é`, a CJK rune). Non-printables become `\xNN` / `\uNNNN` /
`\UNNNNNNNN`; everything else is byte-identical. Keep the table to long-stable Unicode classes
(`strconv.IsPrint` follows the toolchain's tables, G-8). (b) `errorLine` rows: benign text identical
to today's golden text; a context and error text containing a newline followed by `::error::x`
produce exactly one output line. **Red:** the stub returns its input, so the table rows fail and
`errorLine` prints the raw newline. AC: both scenarios fail at the parent for those reasons and the
package compiles.

**U2 (prod; size S, complexity low): `escapeLogText` and `errorLine`.** Blocked by U1.
File: `writepath.go`. Add `escapeLogText(string) string` (iterate with `utf8.DecodeRuneInString`; a
width-1 `RuneError` becomes `\xNN`; `!strconv.IsPrint(r)` becomes a visible escape; backslash is
**not** escaped, so an attacker-written literal `\x0a` is indistinguishable from an escaped newline,
which is cosmetic and carries no workflow-command risk: commands need a real line break). The
doc comment says "keep in sync with any retiredarch copy; the planned home is `pysem` when
`36BE1F61` lands" (AR-5). `errorLine` escapes `context` and `err.Error()`. AC: U1 green; every
existing golden and `TestRunRepoScan_BadListing_FailsClosed` row unchanged (verified: that test
already builds its expectation with `%q`); `go vet` clean.

**U3 (test; size M, complexity medium): call-site sink rows.** Blocked by U2.
Files: `writepath_test.go`, new `writepath_unix_test.go` (`//go:build linux || darwin`). Scenarios (2):
(a) repo-scan sinks as one table: a stub runner returning an error whose text contains a newline
and `::error::` stays on one line; a selected record whose name contains U+009B (passes the D-BW-1a
rune filter) prints its finding line escaped (SEC-3); a read error echoes **no absolute path** (the
`*fs.PathError` is unwrapped); a non-`PathError` error passes through unchanged (SB-8). Update the
pinned expectation of `TestRunRepoScan_DeletedTrackedFile_KeepsReadError`, which today recomputes
`errorLine(rel, readErr)` from the raw `*fs.PathError`: the test unwraps with `errors.As` itself and
must **not** call `unwrapPathError`, so the test commit compiles (G-1). (b) fixture-self-test sinks
(`linux || darwin` file): a fixture named with an embedded newline and `::error::` text yields `PASS` and
`FAIL` lines, joined findings and a U4 `not a contained regular file` line that contain no raw
control character (`assertNoRawControl`). **Red:** these currently echo raw text and the absolute
path. AC: the existing U4 linked-fixture rows (benign names) stay green.

**U4 (prod; size S, complexity low): apply the chokepoint at every sink.** Blocked by U3.
File: `writepath.go`. `runFixtureSelfTest`: escape the name in `PASS`, `FAIL` and the joined
findings, and the `fixture dir not found` / `no fixtures discovered` lines; `runRepoScan`: escape
each line of the findings loop; add `unwrapPathError(error) error` (via `errors.As`) and use it
where `scanFile`'s error reaches `errorLine` in both functions. **Escape each fully composed output
line at its print site** (SEC-11), so a component added later cannot bypass the chokepoint. AC: U3 green; goldens unchanged;
`TestRun_ReadError_FailsClosed` still names the relative path.

**U5 (test; size M, complexity medium): scan-set widening rows, stub-runner level.** Blocked by U4.
File: `writepath_test.go`. **Every row lists a real clean selected `internal/ok.go` beside the
record under test**, so the parent fails with the intended silent pass, not ED-7 (G-2); control rows
assert exit 0. Scenarios (2): (a) a `runRepoScan` table over real files with a stub runner: a
non-`.go` symlink record, a `_test.go` symlink, a symlink to a directory, a symlinked ancestor of a
non-Go file, and a gitlink (real directory plus listed record) each fail with
`not a contained regular file`; controls: a regular non-Go file, a regular `_test.go` and a record
absent from disk pass (exit 0); (b) a pure row pinning the D-WB2-3 decision: `Internal/x.go` and
`CMD/y.go` records are neither selected nor an error. **Red:** every (a) refusal row passes silently
at the parent; controls and (b) are green on arrival (characterization, stated in the test). AC: red
evidence for (a) captured on the Linux job (the symlink rows skip only on Windows Errno 1314).

**U5b (test; size S, complexity medium): scan-set widening rows, real-git level.** Blocked by U5.
File: `writepath_test.go` (and the existing `gitLsFilesCapture` baseline helper, updated to the new
pathspecs, with a separate old-pair helper for the control). Scenarios (2): (a) real-git rows
(`hermeticGitEnv`; records created with `git update-index --cacheinfo`): a gitlink `internal/sub`
beside a clean `internal/ok.go`; **one repository per row** for the top-level records because a link
named exactly `internal` cannot coexist with `internal/ok.go` in one index (D/F conflict, R2-2): a
symlink and a gitlink named `internal` paired with a clean `cmd/ok.go`, and a symlink and a gitlink
named `cmd` paired with a clean `internal/ok.go`, each with a real on-disk symlink or directory
(a record absent from disk passes, G-9). Each row asserts the **runner's listing bytes contain the
record** and that the gate refuses it; (b) a control asserting the new pathspecs return bytes
identical to the old `internal/** cmd/**` pair on the real tree, plus a new vector row in
`TestDefaultGitRunnerIgnoresGitEnv` that still changes the literal-pathspec listing without the
scrub (`GIT_ICASE_PATHSPECS=1` with a tracked `Internal/x.go`): `vector_b` (`GIT_LITERAL_PATHSPECS`)
becomes characterization-only once the pathspecs are literal (R2-1) and is labelled so. **Red:** the
(a) refusals pass silently at the parent; (b) byte-identity is green on arrival (stated). AC: red
evidence for (a) captured on the Linux job.

**U6 (prod; size M, complexity medium): contain every non-selected record.** Blocked by U5b.
File: `writepath.go`. Replace the parse loop with one single-pass iterator
`scanListing(out []byte) (selected, others []string, err error)` that keeps **today's error
precedence for multi-fault listings** (encounter order, G-3), the `len(out)==0` early return and
the not-NUL-terminated check before any record (R2-5); `scannedPathsFromListing` keeps its
signature and returns `selected`; `runRepoScan` runs `containedRegularFile` over `others` in git
order after the selected-path rules and before ED-7; the in-loop check for selected records is
unchanged. `DefaultGitRunner` lists `-- internal cmd` (literal pathspecs; identical bytes on the
real tree, plus top-level `internal`/`cmd` records; on a `core.ignorecase` host the literal
pathspec can also list `Internal/x.go`, which lands in `others` and passes as a regular file, an
accepted D-WB2-3 residual, RR-4). Update the `GitRunner`, `runRepoScan` and
`scannedPathsFromListing` doc comments with the outcome-4 text, the case-variant decision, the
gitlink-absent-from-disk residual (a mode-bearing `ls-files -s -z` would close it, but changes the
`GitRunner` contract, D-WB2-2 option B, LR-3) and the hard-link accepted residual (`A8FB0682`). AC:
U5 and U5b green; `TestScannedPathsFromListing_Table` and `TestRunRepoScan_BadListing_FailsClosed`
unchanged and green; `TestRun_RepoMode_CurrentTreeIsClean` green on the real tree (56 records, none
non-regular); `TestDefaultGitRunnerIgnoresGitEnv` green. A Windows junction as the literal final
component is refused as non-regular (`ModeIrregular`); it is not asserted on the Linux job and is
an accepted residual (LR-4).

**U7 (test; size S, complexity low): fixture directory and FIFO rows.** Blocked by U6.
Files: `writepath_test.go`, `writepath_unix_test.go`. **Each row includes at least one valid fixture**
beside the directory so the parent does not return "no fixtures discovered" (G-2). Scenarios (2):
(a) a table: real directories named `accept-d.go` and `reject-d.go` fail closed with the escaped
name; a directory named `harness` (no `.go` suffix) next to a valid fixture is still skipped and the
self-test passes (SB-7); (b) FIFO (`linux || darwin` file) named `reject-fifo.go` (via `syscall.Mkfifo`) is refused
as not a regular file, run in a goroutine with a timeout so a regression fails instead of hanging;
cleanup unblocks a stuck opener by opening the FIFO `O_RDWR|O_NONBLOCK` (G-6). **Red:** (a)
directory rows pass silently at the parent; (b) is a characterization row (green on arrival,
stated).

**U8 (prod; size S, complexity low): refuse `.go`-named directories.** Blocked by U7.
File: `writepath.go`. In `runFixtureSelfTest` replace the unconditional `e.IsDir()` skip with: a
directory whose name ends in `.go` returns `errorLine(relPath, errors.New("fixture is a directory; names ending in .go must be regular files"))`
exit 1 (raised during collection, before the per-fixture containment loop); other directories stay
skipped. The comment records the hard-link residual. AC: U7 green; `TestRun_SelfTest_MatchesGolden`
and `TestRun_SelfTestIntegrity_MatchesGolden` unchanged.

### RU-2: Gate git discovery and environment isolation (shipment S2)

**U9 (test; size M, complexity medium): write-path wrapper wrong-tree rows and a text pin.**
Files: `tools/gatecheck/check_write_path_wrapper_test.go` (extend the package-shared go shim to
record the `--root` argument it hands the generated binary) and `tools/gatecheck/wrapper_text_test.go`.
Scenarios (2): (a) one wrong-tree **table**; two repositories that **both** contain `internal/` and
`cmd/` files (AS-5); the recorded `--root` must be the wrapper's own checkout (`os.SameFile`,
Windows 8.3 safe) in these rows: ambient `GIT_DIR` and `GIT_WORK_TREE` pointing at the decoy
(including mixed-case names such as `Git_Work_Tree`); `GIT_DIR` alone with cwd in a subdirectory of
the real tree; the ambient name is **mixed-case** in one row (SEC-4); a no-ambient control; a
**symlinked checkout directory** (skips only on Windows Errno 1314); a **nested checkout** (the
script's tree lives inside another repository without its own `.git`): the wrapper must exit
nonzero and **never invoke the engine** (the shim marker stays absent), never scan the parent
(AR-1, RR-2). Every row asserts the shim marker is non-empty when the engine is expected to run, so
a shim that never ran cannot pass (LR-5), and `ROOT` equals the discovered top level, not
`GATECHECK_SRC` (RR-3); (b) a text pin written as a **shape allowlist**: exactly one non-comment
`ROOT`-binding line in the wrapper, equal to `ROOT="$(gatecheck_repo_root)"`; run against the real
committed wrappers before it is trusted and never used as red evidence on its own (LR-1; the
caller-cwd `git rev-parse --show-toplevel >/dev/null` precondition is allowed, AR-3). **Red:** the
decoy, subdirectory and nested rows record the wrong root at the parent (the behavioural rows are
the primary proof). Follow the `requireBash` / `runWritePathWrapperWithTrace` harness.

**U10 (prod; size S, complexity medium): `gatecheck_repo_root` and the write-path wrapper.** Blocked by U9.
Files: `scripts/lib/gatecheck-run.sh`, `scripts/check-write-path-precondition.sh`. Add
`gatecheck_repo_root`: in a subshell, unset every exported variable whose name matches the
**locale-free** glob `[Gg][Ii][Tt]_*` except an exact case-insensitive `GIT_CEILING_DIRECTORIES` (a
`case` pattern, not `${name^^}`: bash 3.2 lacks `^^` and Turkish locales mangle `i`, SEC-4), set
the same three config pins as the Go child (`GIT_CONFIG_NOSYSTEM=1`, `GIT_CONFIG_GLOBAL` and
`GIT_CONFIG_SYSTEM` at the null device, SEC-9), run
`git -C "${GATECHECK_SRC}" rev-parse --show-toplevel`, then **assert the result is the same
directory as `GATECHECK_SRC`** (`[ "$top" -ef "$GATECHECK_SRC" ]`, path-form safe under Git Bash)
and fail closed otherwise (AR-1/AR-6). **Every failure path returns explicitly** (`|| return $?`
after the git call, `return 2` after a failed `-ef`), prints its diagnostic to stderr only, and
never prints a partial `top` (RR-2). The refusals that are intentional false positives (vendored or
monorepo-subdirectory layout, a tarball or Docker copy with no `.git`, git's dubious-ownership
refusal) are named in the function's header comment (RR-3), which also links `1B85D0F4` for the
wrappers not yet using it (RR-5). Do not expand an empty array under `set -u` (bash 3.2). Add
`unset CDPATH` on the line **before** the existing `GATECHECK_SRC` derivation in the lib and before
the pinned `SCRIPT_DIR=` line in the wrapper (a separate added line; the pinned lines stay
byte-identical, SEC-10). The wrapper sets `ROOT="$(gatecheck_repo_root)"` and keeps a
separate caller-cwd `git rev-parse --show-toplevel >/dev/null` precondition (exit 128 outside a work
tree), as the retiredarch wrapper does. AC: U9 green; `TestGatecheckRunSh_ShellcheckCleanIfAvailable`,
`TestWritePathWrapperText_PureBashDispatch` and the runner tests green; the header comment states
the anchoring change (D-WB2-4).

**U11 (test; size S, complexity medium): retiredarch wrapper wrong-tree rows and text pin.** Blocked by U10.
Files: `tools/gatecheck/check_retired_architecture_wrapper_test.go`, `wrapper_text_test.go`. Same
two scenarios as U9 against `scripts/check-retired-architecture.sh`, **reusing the shim extension
U9 adds** (which is why this unit stays S). **Red:** the decoy and subdirectory rows (the
`git -C GATECHECK_SRC` form returns the decoy under ambient variables, verified by probe) and the
nested-checkout row; the text pin is not red evidence (LR-1).

**U12 (prod; size S, complexity low): retiredarch wrapper uses the shared function.** Blocked by U10, U11.
File: `scripts/check-retired-architecture.sh`. Replace the `ROOT="$(git -C ...)"` line with
`ROOT="$(gatecheck_repo_root)"`; keep the caller-cwd precondition line; add `unset CDPATH` before the
pinned `SCRIPT_DIR=` line; rewrite the "Root anchoring (C-3 exception)" comment block
to say `gatecheck_repo_root` is the single root-policy point (AR-3). AC: U11 green; the
`TestRetiredArchWrapperText_*` pins and `TestCheckRetiredArchitectureWrapper_FailsOutsideAGitWorkTree`
green (the pinned phrases are the four decision-record strings and the dispatch lines; none names the
replaced line).

**U13 (test; size S, complexity low): twin env-table rows.** Blocked by U12 (sequence only).
Files: `writepath_test.go`, `retiredarch/select_test.go`. Scenarios (2, one per package, the **same
table**): strip `MSYS=1`, `msys=1`, `CYGWIN=x`, `Msys2_Arg_Conv_Excl=*`; keep `MSYSTEM=MINGW64`,
`MSYSX=1`, `MSYS2=1`, `CYGWINX=1`; the three boundary rows from the entry: bare `GIT_DIR` (no `=`),
exact `GIT_`, bare `GIT_CEILING_DIRECTORIES`. Identical tables in both packages are the parity
mechanism (SB2-2: a cross-package source-parsing test was dropped as out of the entry's scope; the
U15 pin already freezes the retiredarch body, and `36BE1F61`/`5A8EC1BC` own any future shared
helper). **Red:** the strip rows fail in both packages; the keep rows and the three boundary rows
are green on arrival (characterization, stated).

**U14 (prod; size S, complexity low): writepath `gitRunnerEnv`.** Blocked by U13.
File: `writepath.go`. Add the MSYS rule **inline inside `gitRunnerEnv`**, in the same shape as the
`GIT_` rule (AR-8); update the doc comment (copy; keep bodies identical with retiredarch). AC: the
writepath table of U13 green.

**U15 (prod, atomic; size S, complexity medium): retiredarch `gitRunnerEnv` and its pin.** Blocked by U14.
Files: `retiredarch/select.go` and `retiredarch/pin.go` **in one commit**: the `gitRunnerEnv` body
in `select.go` and the hand-authored `canonicalDecls` copy in `pin.go` (token-equal). Constraints
(G-4): the MSYS rule is **inline** (no new top-level declaration or helper, because
`closedWorldDecls` and `sharedRulesOK` would clear `PathspecOK`); the literal `"GIT_CONFIG_NOSYSTEM=1"`
must appear exactly once in `select.go` (`pin_test.go` mutates its first occurrence), so do not
repeat it in new comments; no new import; every `writeMutatedCopy` anchor in `pin_test.go` still
resolves exactly once after the change (LR-2). `closedWorldDecls`, `pathspecFrozenDecls` and
`prefixFrozenDecls` name declarations, not text, and do not change. **Red evidence for U15 is the
U13 retiredarch strip rows** at the parent; the pin test is the green guard. AC: those rows green;
`pin_test.go` green including `TestCheckPathspecPin_GitRunnerIsolation_Rejected`;
`git show --stat` lists exactly those two files; the writepath and retiredarch `gitRunnerEnv`
bodies are identical (reviewer diff, recorded in the task evidence).

## Dependency Graph

```text
043-S ─blocks→ S1:  U1 → U2 → U3 → U4 → U5 → U5b → U6 → U7 → U8
S1 ─blocks→ S2:     U9 → U10 → U11 → U12 → U13 → U14 → U15
```

All task-level edges are single-predecessor chains (no cycle), except U12, which is blocked by both
U10 and U11. The linear order is intentional: RU-1 tasks all edit `writepath.go` and
`writepath_test.go`; U13-U15 share the twin table. Ship executes serially, so the chain adds no cost.
Every commit between pair completions builds and passes `go vet`; only test-unit assertions are red.


## Constitution Check

| Principle | Verdict | Note |
|---|---|---|
| I Safety-First Go | Pass | Errors wrapped and escaped at the sink; no `panic`; `strconv.IsPrint` on the exact input; `go vet` and Linux `GOOS` vet in verification. |
| II Test-First (NON-NEGOTIABLE) | Pass | Every behaviour change is a test unit then a production unit; every test commit compiles and passes `go vet` (U1 carries a compiling stub, RC-1); green-on-arrival characterization rows are labelled and never counted as red evidence; U15's red is the U13 retiredarch strip rows. |
| III Workspace Isolation | Pass | Fixtures under `t.TempDir()` (managed and cleaned by `testing`); no agent file operation targets a path outside the cwd; git child scrubbed. |
| IV CLI Workspace Containment (NON-NEGOTIABLE) | Pass | The change *strengthens* containment (every record passes `containedRegularFile`); root is anchored to the script's own checkout. |
| V Structured Observability | Pass | Stderr stays one `::error::` line per outcome; milestone broadcasts are Ship's intercom duty (degraded visibility must be warned, not bypassed); commit convention stated. |
| VI Single Responsibility | Pass | One shared shell function for root policy; `escapeLogText` stays package-local (copy, not shared, D-WB2-6); no new dependency. |
| VII Destructive Command Approval (NON-NEGOTIABLE) | Pass | No unit contains a destructive command: `git revert` appends history, `backlogit` archive moves rather than deletes; any `rm`, `git rm`, `reset`, force-push or history rewrite needs operator approval and is not planned. |
| VIII Explicit Safety Modes | Pass | Careful mode for U6, U10, U12, U14, U15; freeze-scope bounded to `tools/gatecheck/**` and `scripts/**`. |
| IX Git-Friendly Persistence | Pass | Plain-text sources and docs. |
| X Agent Context Efficiency | Pass | Units are small file sets; verification output is summarised. |
| XI Merge Commit History | Pass | Merge order and the "never squash the Batch 1 PR" rule are explicit. |
| Task Granularity (NON-NEGOTIABLE) | Pass with notes | Every unit has fewer than 3 files and a table counts as one scenario (a deliberate loosening, RC-4); U5 was split into U5 and U5b so no unit has more than 2 scenarios; U6 touches `scanListing`, the `scannedPathsFromListing` wrapper, `runRepoScan` and `DefaultGitRunner`'s argv (4, under the bound of 5); size M units (U3, U5, U6, U9) are each one reviewable diff. |

## Decisions and Rationale

See the deliberation: D-WB2-1 (two shipments), D-WB2-2 (contain every non-selected record, **as
amended by plan review: literal `internal`/`cmd` pathspecs so a top-level symlink or gitlink is
listed**), D-WB2-3 (case variants out of scope; the Windows-build exposure is an accepted residual),
D-WB2-4 (script-anchored scrubbed root through one shared function, **as amended: same-directory
assertion and a locale-free glob**), D-WB2-5 (per-arm ED-7 guard rejected), D-WB2-6 (one local
escaping chokepoint, `!strconv.IsPrint`), D-WB2-7 (directory refusal, FIFO as a build-tag row, hard
link not implemented), D-WB2-8 (MSYS strip, twin tables, atomic refreeze), D-WB2-9 (new deferred
entries `36BE1F61`, `1B85D0F4`, `A8FB0682`, plus plan-review captures `284270E8`, `49982A4E`),
D-WB2-10 (no overlap with 043-S).

## Risks and Caveats

| Risk | Mitigation |
|---|---|
| A new fail-closed outcome blocks every PR if a legitimate tracked symlink or gitlink exists | Verified none exists today (56 records, modes 100644/100755 only); U6 AC runs `TestRun_RepoMode_CurrentTreeIsClean`. Rollback: revert U6 **with** U5. |
| The literal-pathspec change in U6 lists a different set | U5b(b) asserts byte-identical listings to the old pathspecs on the real tree; `TestDefaultGitRunnerIgnoresGitEnv` stays green (its `GIT_LITERAL_PATHSPECS` vector becomes characterization-only and gains an `GIT_ICASE_PATHSPECS` vector, R2-1). |
| Escaping changes a golden | Escape only touches non-printable runes; U2 and U4 ACs require every golden byte-identical. |
| Wrapper anchoring changes behaviour for callers outside the checkout | Documented (D-WB2-4, operator checkpoint 3); caller-cwd precondition and exit 128 preserved; engine and tree now come from the same checkout; a nested checkout now fails closed. |
| Shell scrub is not portable | Locale-free glob `[Gg][Ii][Tt]_*`, `compgen -e` is a bash builtin (wrappers are bash), no empty-array expansion under `set -u`; U9/U11 run under `requireBash`, including Windows Git Bash in the advisory job. Fallback: D-WB2-4 option C. |
| `pin.go` and `select.go` drift | U15 is a single atomic commit with a stat assertion and a reviewer diff of the two `gitRunnerEnv` bodies; identical U13 tables in both packages are the behavioural parity guard. |
| Windows cannot create symlinks or control-char names | Existing Errno 1314 skip for symlinks only; `unix`-tagged rows are not skipped; Linux CI is the required evidence; `GOOS=linux go vet` covers the tagged files from any host. |
| Review-cycle fatigue (B1 hit the cap) | Two shipments (D-WB2-1); each unit is one reviewable diff. |

**Accepted residuals (carried into each closure record with stash IDs):** hard link refusal
(`A8FB0682`); retiredarch echo sinks (`36BE1F61`) so the log-injection class is closed for the
write-path gate only (SEC-7); the other wrappers' root discovery (`1B85D0F4`); the retiredarch
scan-set twin of this batch (`284270E8`); engine-side `--root` top-level assertion for direct
invocations (`49982A4E`); case-variant and module-wide scope, including the Windows-build
exposure of an `Internal/` directory (D-WB2-3, SEC-2, operator checkpoint 2); a symlink or gitlink
record whose path is absent from disk (G-9; a mode-bearing `ls-files -s -z` would close it, but
changes the `GitRunner` contract, LR-3); a Windows junction as the literal `internal`/`cmd` final
component (refused as non-regular, not asserted on the Linux job, LR-4); on a `core.ignorecase` host
the literal pathspec can list `Internal/x.go`, which passes as a regular file (RR-4); TOCTOU between
the walk and the read; the shell `unrecognized argument` echo (workflow YAML is trusted); the
`tools/gatecheck/main.go` `::error::` sinks (`parseRoot` error text and the panic-recovery `%v`),
which are outside this batch's echo chokepoint, so the log-injection class is closed for the
write-path **engine's own output only** (SEC-8; scope added to `36BE1F61`); backslash left
unescaped (cosmetic).

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
* **Ship verification (Stage does not run these):** at every pair completion (U2, U4, U6, U8, U10,
  U12, U15): `go test ./tools/gatecheck/...` (race on the Linux job), `go vet ./...`,
  `GOOS=linux go vet ./tools/gatecheck/...` (covers the `unix`-tagged files from any host), the full
  `go test ./...`, shellcheck on the touched scripts, `scripts/check-write-path-precondition.sh
  --self-test` and `scripts/check-retired-architecture.sh --self-test` on the real tree (both exit
  0). At test units record the red output only.
* **Operational closure:** each shipment closes through the post-043-S single close procedure;
  closure record lists the accepted residuals (see Risks and Caveats) with stash IDs.
* **Stage verifies after assembly:** both shipments exist as queued, the `blocks` edges
  `043-S ← S1 ← S2` read back through `backlogit dep`, manifests match the harvested IDs, and the
  feature precedes its tasks in each shipment.

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
  with git's exit code when git fails, as today (the U9 table asserts a non-git root still fails,
  not a silent empty `--root`).
* **Case-insensitive env names without locale dependence:** the scrub uses the glob
  `[Gg][Ii][Tt]_*`, never `${name^^}`, so `git_dir` and `Git_Work_Tree` are removed under any
  locale (SEC-4); a U9 row sets a mixed-case ambient name.
* **U5b real-git rows:** use `hermeticGitEnv` so ambient developer variables cannot flip the red
  evidence (B1 lesson); assert listing bytes, not only exit code.
* **U7(b) FIFO:** a regression would block forever in `os.ReadFile`; the goroutine-and-timeout
  wrapper is part of the AC, not optional.
* **Stash disposition:** the archive notes also carry the SEC-8 residual (below).
* **Echo of non-selected record names (U6):** these names are not rune-validated by design; U2/U4
  must already be merged so the refusal line is born escaped. The chain enforces that order.
* **U13 keep-rows:** `MSYSTEM`, `MSYSX`, `MSYS2` (no underscore), `CYGWINX` pin the exact-name and
  prefix boundary so a later widening is deliberate.
* **Shell/Go scrub parity (AR-4):** the shell glob and the Go `gitRunnerEnv` follow the same rule;
  the U9 mixed-case row is the parity check, and `GIT_CEILING_DIRECTORIES` is kept in both.

### Risky actions (`ProposedAction` / `ActionRisk` / `ActionResult`)

| ID | ProposedAction | ActionRisk | Approval | Expected ActionResult and rollback |
|---|---|---|---|---|
| PA-1 | U6: add the matrix row-4 refusal and the literal pathspecs to the merge-blocking gate | medium (false positive blocks all PRs) | operator review of the staging PR; careful mode | Result: gate stays green on the real tree. Rollback: revert the U6 commit together with U5 (record as a pair; `git revert` appends history, Principle VII not triggered). |
| PA-2 | U10/U12: change wrapper root anchoring from caller-cwd to script location, with a same-directory assertion | medium (behaviour change) | operator checkpoint 3; careful mode | Result: same `--root` for every in-checkout invocation; a nested checkout fails closed. Rollback: revert U10/U12 together with U9/U11. |
| PA-3 | U15: refreeze retiredarch `gitRunnerEnv` and `canonicalDecls` | medium (pin gate integrity) | none beyond review | Result: pin test green. Rollback: revert the single commit. |
| PA-4 | Archive the five consumed stash entries after shipments are assembled | low (reversible: archive, not delete) | none; the disposition note links D-WB2-3, `A8FB0682` (the unimplemented hard-link remainder of `858516B9`) and the accepted residuals so no decision is lost with the archive (SB-3, SB2-8) | Result: entries retrievable from the archive with a disposition note. |
| PA-5 | Create two features, 16 tasks, two shipments, edges `043-S ← S1 ← S2` | low | none | Rollback: `backlogit` archive of the new items before Ship claims. |

### Verification, monitoring and closure additions

* **Prechecks Ship must record per task:** `git --version` is 2.32 or later; `go version` matches
  the repo pin; bash available (`requireBash`); on the Linux job the symlink rows are not skipped.
* **Red evidence is mandatory** for U1, U3, U5(a), U5b(a), U7(a), U9(a) behavioural rows, U11(a), U13 strip rows: the
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

## Plan Review (attempt 1, superseded)

dispatch_mode: multi-agent
decision: FAIL

Five of six personas ran as subagents (Constitution Reviewer, Go Reviewer, Scope Boundary Auditor,
Architecture Strategist, Security Lens Reviewer); the Learnings Researcher did not run in this
attempt (its Step 1.8 pass is recorded in the deliberation) and runs in attempt 2. Gate: FAIL on
P1 findings. Disposition of every finding follows; the revised plan above is the re-review input.

| ID | Sev | Finding | Disposition |
|---|---|---|---|
| CR-1 | P1 | No Constitution Check section | **Fixed**: `## Constitution Check` added (all principles) |
| CR-2 | P1 | Safety mode not declared per unit | **Fixed**: careful mode U6/U10/U12/U15, freeze-scope `tools/gatecheck/**` + `scripts/**` |
| CR-3 | P1 | Principle VII and PA-4/PA-5 approvals | **Fixed**: no destructive command is planned; `git revert` appends history, `backlogit` archive moves; VII row in the check |
| AR-1 | P1 | Script anchoring trusts git discovery (nested checkout); no top-level check | **Fixed (wrapper)**: `-ef GATECHECK_SRC` assertion + nested-checkout row (U9/U10); engine part captured `49982A4E` |
| AR-2 | P1 | Direct `--root <subdir>` partial listing | **Accepted residual**: captured `49982A4E` (frozen retiredarch `DefaultGitRunner`; trusted `--root` contract) |
| SEC-1 | P1 | Top-level `internal`/`cmd` symlink or gitlink not listed | **Verified by probe and fixed**: literal pathspecs (U5c/U6); retiredarch twin captured `284270E8` |
| SEC-2 | P2 | Windows case-variant exposure | **Accepted residual**, now recorded explicitly (D-WB2-3) |
| SEC-3 | P2 | Findings loop echoes raw relPath | **Fixed**: U3/U4 cover the findings loop |
| SEC-4 | P2 | `${name^^}` locale and bash 3.2 | **Fixed**: locale-free glob (U10) |
| G-1 | P2 | U3 references `unwrapPathError` before U4 | **Fixed**: test unwraps with `errors.As` |
| G-2 | P2 | Red evidence hits ED-7 / "no fixtures" | **Fixed**: every U5/U7 row has a clean selected file or valid fixture |
| G-3 | P2 | `listedRecords` changes multi-fault precedence | **Fixed**: single-pass `scanListing` |
| G-4 | P2 | Pin closed-world constraints on `select.go` | **Fixed**: inline-only, single `GIT_CONFIG_NOSYSTEM=1` literal (U15) |
| SB-1..SB-5 | P2 | Split wording; 043-S edge; residual tracking; verification line; U9 scenario count | **Fixed** (see sequence section, PA-4, verification, U9 table) |
| AR-3..AR-7, CR-4..CR-8, G-5..G-9, SB-6..SB-10, SEC-5..SEC-7 | P2/P3 | Advisory | **Applied** where cheap (text pins U9/U11, parity test U13c, `unix` tag, FIFO cleanup, echo completeness, keep-in-sync note); SB-6 (merge U14/U15) rejected: three files breaks the 2-hour bound; SB-10 trimming deferred |

<!-- plan-review-attempt: 1 -->

## Plan Review (attempt 2, superseded)

dispatch_mode: multi-agent
decision: FAIL

All six personas ran as subagents (Constitution Reviewer, Go Reviewer, Scope Boundary Auditor,
Learnings Researcher, Architecture Strategist, Security Lens Reviewer). Gate: FAIL on one P1
(RC-1). Every attempt-1 P1 was verified closed by its owning persona; no new P0. Disposition:

| ID | Sev | Finding | Disposition |
|---|---|---|---|
| RC-1 | P1 | Test-first commits must compile; U1 was compile-red | **Fixed**: U1 carries a compiling stub; convention states every test commit compiles and passes `go vet`; Constitution Check row II updated |
| RC-2 | P2 | "no unit writes outside the cwd tree" contradicts `t.TempDir()` | **Fixed**: claim reworded (test temp dirs vs agent file operations) |
| RC-3 | P3 | U14 missing from careful mode | **Fixed** |
| RC-4 | P3 | Tables-as-one-scenario hides volume; red evidence without a second worktree | **Fixed**: loosening stated; U5 split into U5/U5b; red captured at the test commit |
| R2-1 | P2 | `GIT_LITERAL_PATHSPECS` vector becomes vacuous | **Fixed** (U5b: characterization label plus a `GIT_ICASE_PATHSPECS` vector) |
| R2-2 | P2 | `internal` link and `internal/ok.go` cannot share an index (D/F conflict) | **Fixed** (U5b: one repository per top-level row) |
| R2-3, SB2-1, SB2-2 | P3/P2 | Parity test bisect hazard and scope | **Fixed by dropping the parity test**; identical U13 tables are the parity mechanism |
| R2-4, LR-6 | P3 | `unix` tag and `syscall.Mkfifo` coverage | **Fixed**: `linux \|\| darwin` tag, vet on both |
| R2-5 | P3 | `scanListing` early returns | **Fixed** (U6 AC) |
| SB2-3 | P2 | U5 had about four scenarios | **Fixed**: U5 and U5b |
| SB2-4 | P2 | U11 sizing | **Fixed**: U11 reuses the U9 shim (stays S) |
| SB2-5, LR-1 | P3/P2 | Textual pin is a denylist | **Fixed**: shape allowlist, run against real wrappers, never sole red evidence |
| SB2-6, SB2-8 | P3 | Header IDs; PA-4 link to `A8FB0682` | **Fixed** |
| SB2-7 | P3 | Lumped attempt-1 dispositions | Accepted: attempt-1 SB-6 rejected (three files break the bound), SB-7 applied (U7 rows), SB-8 applied (U3 pass-through row), SB-9 recorded cosmetic (U2), SB-10 (trim hardening prose) deferred, no owner needed |
| LR-2 | P2 | Pin anchors uniqueness | **Fixed** (U15 AC) |
| LR-3 | P2 | Gitlink absent from disk | **Recorded** as accepted residual with the `ls-files -s -z` fix named |
| LR-4 | P2 | Windows junction | **Recorded** as accepted residual |
| LR-5 | P3 | Shim marker, Git Bash pin | **Fixed** (U9 AC) |
| RR-2 | P2 | Explicit returns in `gatecheck_repo_root` | **Fixed** (U10) |
| RR-3 | P3 | `-ef` false positives, symlinked checkout | **Fixed**: named refusals, symlinked-checkout row |
| RR-4 | P2 | `core.ignorecase` listing differs | **Recorded** in U6 and the residual list |
| RR-5, RR-6, SEC-8 | P3 | Shared-lib adoption, direct `--root`, `main.go` sinks | **Recorded**: `1B85D0F4`, `49982A4E`, scope added to `36BE1F61` |
| SEC-9, SEC-10, SEC-11 | P3 | Config pins in the shell scrub; `CDPATH`; escape composed lines | **Fixed** (U10, U12, U4) |

<!-- plan-review-attempt: 2 -->

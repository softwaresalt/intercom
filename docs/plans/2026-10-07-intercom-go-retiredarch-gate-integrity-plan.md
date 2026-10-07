---
title: "Harden retiredarch gate integrity — implementation plan"
source: "docs/decisions/2026-10-07-intercom-go-retiredarch-gate-integrity-deliberation.md"
status: "reviewed"
revision: 3
stash_ids: ["3750C37C", "9FC28DB9", "0ECC1895", "DC921AF6", "D7BF9F74", "990AFA71"]
package: "tools/gatecheck/internal/retiredarch"
base: "main@372ab38"
---

# Harden retiredarch gate integrity — implementation plan

## Revision History

| Rev | Change |
|---|---|
| 1 | Initial plan, RA-1..RA-10. Plan review attempt 1: **FAIL**. Architecture P1×3 and Security P1×2; Go Reviewer P1×2. |
| 2 | Remediation of every P1 and the actionable P2s (see "Plan Review — attempt 1" at the end). Changes: <br>• RA-5 also isolates `gitShowToplevel`. <br>• RA-4 adds an executable-location guard. <br>• New RA-7 pin-test fixture harness. <br>• RA-8 specifies cross-file helpers. <br>• New RA-9 freezes the self-test orchestration chain. <br>• RA-10 freezes the register file's import and adds a single-literal-registration rule. <br>• New RA-13 extends the init-time scan to process spawns and env writes in every gatecheck package. <br>• RA-1 fallback `default` fails closed. <br>• RA-2 seam becomes a parameter. <br>• RA-12 allowlists are keyed by (file, decl, import path). <br>• Constitution Check and quality-gate sequence added. <br>• Residuals R-A1b/R-A1c/R-A5 added. <br>Units are now RA-1..RA-13. Plan review attempt 2: **FAIL**. Architecture and Security each raised a P1 (the same one): RA-10's registration rule is name-based and bypassable through indirect references. |
| 3 | Fixes the attempt-2 P1 and the P2s (see "Plan Review — attempt 2"):<br>• RA-10 becomes a positive **reference** rule over package `main`.<br>• The pin root is resolved once (RA-5-guarded `gitShowToplevel`) and passed explicitly into `checkPathspecPin(root)`; fixtures pass their temp root (RA-7).<br>• A single `indexDecls` indexer is used.<br>• RA-4 contamination vectors are now observable (`core.quotePath`).<br>• `gitShowToplevel` exec-guard anchor is defined, with canonicalisation.<br>• RA-1 matching is stated as key-only, with explicit scalar types.<br>• RA-12 is split into RA-12 (init rules) and RA-13 (selector allowlist).<br>• Former RA-13 becomes **RA-14**, now package-level (not a function denylist) plus a `go:linkname` ban.<br>Units are now RA-1..RA-14. |

## Problem Frame

The `retired-arch` gate (`tools/gatecheck` → `retiredarch.Run`, CI `ci.yml:325` repo mode and
`:332` self-test modes) decides whether `cmd/**` and `internal/**` reintroduce retired
architecture tokens. It has five structural gaps, all verified against `372ab38`:

1. **TOML engine verdict bug (`9FC28DB9`).** `walkTable` (tomlprimary.go:148) walks
   `MetaData.Keys()` with a cursor. It returns a desync error when an ancestor table's
   self-entry appears while a deeper, incomplete descendant frame is open (for example
   `[a.b] x=1` / `[a] y=2` / `[a.b.c] k=3`), and also for non-contiguous headers
   (`[a.b]` / `[z]` / `[a.c]`). `scanTomlPrimary` (:285) converts that into a generic
   "TOML parse error" finding. The gate therefore fails closed on a valid file and never
   reports the real findings.
2. **Symlink-following scan (`990AFA71`).** `runRepoScan` (retiredarch.go:590) scans
   `filepath.Join(root, FromSlash(rel))` with no `Lstat`. A tracked symlink therefore makes the
   gate scan bytes outside the selected path.
3. **Inherited git environment (`D7BF9F74`, R-A2).** `DefaultGitRunner` (select.go:32) and the
   pin's own root discovery `gitShowToplevel` (pin.go:706) run git with the inherited
   environment. That lets `GIT_*_PATHSPECS`, `GIT_DIR`/`GIT_INDEX_FILE`/`GIT_WORK_TREE`,
   `GIT_CONFIG_COUNT`/`GIT_CONFIG_PARAMETERS` and global/system config reinterpret or
   redirect selection, or point the pin at a decoy tree.
4. **Unpinned dispatch and wiring (`DC921AF6`, R-A1).** `engineForPath` and `scanPath`
   (select.go) are closed-world names but are not token-frozen. None of the following is
   pinned: `runRepoScan`, `Run` and `runSelfTestAssertions` (retiredarch.go:590/625/670),
   `runRepoSelectionSelfTest` (selftest_selection.go:138), and the registration
   `init` + `runRetiredArch` in `tools/gatecheck/register_retired_arch.go`.
   `registerSubcommand` (main.go:59) silently overwrites a duplicate name.
5. **Open package closure (`3750C37C`, `0ECC1895`).** `packageClosureOK`/`closureFileOK`
   (pin.go:302/355) constrain names, `unsafe`/`C`, linkname, `os`/`syscall` dot imports and
   three env-mutator selectors. They do not constrain sibling-file side effects (`func init`,
   package-level initializers, process spawns) or Windows env mutation
   (`syscall.SetEnvironmentVariable`, `syscall.NewLazyDLL`, `golang.org/x/sys/windows`).
   `initstate_test.go` scans every gatecheck package for init-time process-state reads and
   filesystem ops, but not for process spawns or env writes.

The pin (`pin.go`, `SelectionPathspecPin`) is the gate's trust root. Rev 6 / 030-S / ALP-1
established the "canonical text + frozen set + reject table" pattern. It is evaluated
twice in CI:

* at runtime, via `--self-test-integrity`;
* independently of `Run`, via `go test` (`TestSelectionPathspecPin_LiveTree`).

The second path is what makes freezing the pin's own invoker chain meaningful (RA-9).

## Requirements Trace

| Req | Source | Decision | Units |
|---|---|---|---|
| R1: valid desync-shaped TOML yields its true findings, never a parse-error finding and never a silent skip | `9FC28DB9` | D-RA-1 | RA-1 |
| R2: a future cursor-walk omission fails closed | `9FC28DB9` (F-1 lesson) | D-RA-1 | RA-2 |
| R3: a selected path that is or traverses a symlink, junction or non-regular entry fails closed | `990AFA71` | D-RA-2 | RA-3 |
| R4: git selection **and pin root discovery** are immune to inherited `GIT_*` env and global/system config; the git executable is not resolved from inside the checkout | `D7BF9F74` | D-RA-3 | RA-4, RA-5 |
| R5: dispatch (`engineForPath`, `scanPath`) is token-frozen | `DC921AF6` | D-RA-4 | RA-6 |
| R6: repo-scan wiring (`runRepoScan` + containment helper) is token-frozen | `DC921AF6`, `990AFA71` | D-RA-4 | RA-7, RA-8 |
| R7: mode dispatch and the self-test orchestration chain (`Run`, `runSelfTestAssertions`, `runRepoSelectionSelfTest`) are token-frozen | `DC921AF6` | D-RA-4 | RA-9 |
| R8: CLI wiring (register import, `init`, `runRetiredArch`) is token-frozen, and `retired-arch` is registered exactly once by a literal name | `DC921AF6` | D-RA-4 | RA-10 |
| R9: package file set and per-file imports are an exact allowlist; Windows env-mutation imports are rejected | `3750C37C`, `0ECC1895` | D-RA-5 | RA-11 |
| R10: no init-time side effects in retiredarch; `os`/`exec` selectors are allowlisted per (file, decl) | `3750C37C` | D-RA-6 | RA-12, RA-13 |
| R11: no init-reachable use of process-spawn, env-write, syscall, x/sys, unsafe or plugin packages in any gatecheck package; no `go:linkname` (go-test level) | `3750C37C` (linked-package vector) | D-RA-6 | RA-14 |
| R12: the `pin.go` residual block re-states R-A1..R-A5 precisely | all | D-RA-3/4/6 | RA-5, RA-10, RA-13 |

## Invariants

* **INV-1 (verdict preservation).** On the current tree, the stdout and exit code of
  `gatecheck retired-arch` in repo mode, `--self-test` and `--self-test-integrity` are
  byte-identical before and after every unit. That covers `TestRun_*_MatchesGolden`,
  `TestRun_Repo_MatchesGolden` and `testdata/retiredarch_golden.json`, which stays unchanged.
  The only permitted behaviour changes are these enumerated deltas:
  * **ED-10** (RA-1/RA-2): a desync-shaped TOML file that previously produced the generic
    parse-error finding now produces its real findings, in lexicographic depth-first order.
    A cursor/oracle mismatch, a desync-subset violation or an unhandled value type produces a
    fail-closed finding.
  * **ED-11** (RA-3): a selected path that is or traverses a symlink, junction or non-regular
    entry produces a fail-closed synthetic finding. None exist today.
  * **ED-12** (RA-4/RA-5): `git` runs with `GIT_*` dropped (except `GIT_CEILING_DIRECTORIES`)
    and global/system config disabled. A `git` executable that resolves relative to, or
    inside, the root is refused (fail closed).
  * **ED-13** (RA-6..RA-10): edits to the newly frozen declarations, or a second or
    non-literal `retired-arch` registration, clear `PathspecOK`/`PrefixOK`.
  * **ED-14** (RA-11..RA-13): an unlisted file, import, dot import, `func init`,
    non-allowlisted initializer call or `os`/`exec` selector clears `PathspecOK`/`PrefixOK`.
* **INV-2 (fail closed).** Every new failure path yields a finding, an error exit, or a cleared
  pin flag. None may yield `nil`/skip/pass. Every type switch over decoded or AST data has a
  fail-closed `default`.
* **INV-3 (golden-stable report shape).** No new `PathspecPin` field and no new self-test
  assertion line. Every new pin rule is a conjunct of the existing `PathspecOK` **and**
  `PrefixOK` flags (as `closureOK` is today), so the `--self-test-integrity` report at
  selftest_selection.go:280-290 is unchanged.
* **INV-4 (independent authorship, H-11).** Every canonical text, manifest and allowlist in
  `pin.go` is authored as literal text. It is never derived at runtime from the file it pins,
  and never generated by tooling.
* **INV-5 (AST-shape matching).** Every new ban matches AST shape, never raw text, so `pin.go`'s
  own literals cannot self-match. The shapes are `*ast.ImportSpec`, `*ast.SelectorExpr`
  resolved through the file's import map to an import path, `*ast.FuncDecl` and
  `*ast.ValueSpec`.

## Execution Discipline

* **One implementation branch and one worktree. Units run serially** in the order below.
  Parallel worktrees or branches are prohibited (P-016), even where the DAG shows no
  dependency.
* **Posture: test-first.** Each unit's *red* test is written first and **must fail at the
  unit's parent commit** for the stated reason. Ship records the failing test name and reason
  in the run log. Rows marked *(characterization)* or *(regression guard)* are not red proof.
* **Quality gates per commit, in constitution order:**
  1. `gofmt -l .` produces no output;
  2. `go vet ./...`;
  3. `go test ./...`;
  4. `go build ./...`;
  5. `golangci-lint run` when configured (its warnings are defects);
  6. the INV-1 check: `TestRun_Repo_MatchesGolden`, both self-test goldens, and
     `bash scripts/check-retired-architecture.sh` in all three modes with unchanged exit
     codes.
* **Safety mode.**
  * Freeze scope: `tools/gatecheck/internal/retiredarch/**`, plus read-only access to
    `tools/gatecheck/register_retired_arch.go` and package `main`'s non-test `.go` files
    (RA-10).
  * `careful` mode for the ALP-2 commit (RA-4 + RA-5).
* **Scope.** Never touch scan scope (`scanScope`, `scanArm`, `shouldScanRepoPath`),
  forbidden-token tables, other engines' matching, `.github/workflows/**`, `scripts/**`, or
  any package other than retiredarch. No golden regeneration (H-INV-6).
* **HALT and return to Stage** if a unit:
  * needs a file not listed for it;
  * needs a new `PathspecPin` field;
  * changes a golden outside its ED;
  * would allowlist a write-, exec- or env-capable selector not named in this plan;
  * exceeds 2× its size estimate.
* **Visibility.** When the agent-intercom overlay is active, broadcast per-unit milestones and
  the ALP-2 landing.

## Implementation Units

### RA-1 — Desync fallback to a complete deterministic walk (`9FC28DB9`)

* **Files:** `tomlprimary.go`, `tomlprimary_test.go`.
* **Change:** The success path keeps the cursor walk. When `walkTable`/`walkArrayOfTables`
  report a desync, `scanTomlPrimary`:
  1. Re-walks the decoded `map[string]any` deterministically:
     * keys sorted bytewise per table;
     * sub-tables recursed;
     * `[]map[string]any` (arrays of tables) and `[]any` (plain arrays, including inline
       tables and nested arrays) handled element by element, using the same `keyPath`
       rendering as the cursor walk;
     * matching is **key-path only**, exactly as `walkTable` does it
       (`matchesForbiddenParts(composeTomlParts(keyPath))` on the first visit of each key per
       table instance, tomlprimary.go:~194). Values are never matched, so the RA-2 oracle
       stays consistent with INV-1;
     * explicit scalar cases for `string`, `int64`, `float64`, `bool` and `time.Time`
       (BurntSushi decodes local date/time into `time.Time`), plus a fail-closed `default`
       for any other Go type (INV-2).
  2. Requires the partial cursor findings collected before the desync to be a sub-multiset of
     the fallback findings. Otherwise it emits a fail-closed finding.

  A real `toml.Decode` failure still produces the parse-error finding.
* **Tests (scenarios: 3):**
  1. **Red.** A table-driven test over the desync shapes (ancestor reopen, non-contiguous
     headers) × forbidden-token *key* placement (a scalar key, a key in an array-of-tables
     element, a key in an inline table inside an array, a key under a nested array) yields
     exactly the expected findings. The
     first row asserts the observed `meta.Keys()` sequence, so the ordering premise is
     checked rather than assumed. It fails at parent with the parse-error finding.
  2. **Red.** The same shapes with no forbidden token yield no findings. Fails at parent with
     the parse-error finding.
  3. *(Characterization.)* Every existing primary-engine golden and dual-engine agreement test
     still matches.
* **Size: S | Complexity: medium.**

### RA-2 — Completeness oracle on the cursor-walk success path (`9FC28DB9`)

* **Files:** `tomlprimary.go`, `tomlprimary_test.go`. **Blocked by:** RA-1.
* **Change:**
  * Split the walk into `scanTomlPrimaryWith(path string, walk tomlWalker) []string`.
    `scanTomlPrimary` becomes a thin wrapper that passes the real cursor walker. This uses
    parameter injection, **not** a package-level variable. `tomlWalker` keeps the
    `findings *[]string` out-parameter, so RA-1's partial-findings subset check still works.
  * After a successful cursor walk, run the RA-1 deterministic walk and compare the two
    finding **multisets**. On mismatch, return the fallback findings plus one fail-closed
    finding, `TOML completeness check failed: cursor walk omitted or invented findings`.
  * Emission order otherwise stays the cursor order.
* **Tests (scenarios: 2):**
  1. **Red.** Pass `scanTomlPrimaryWith` a walker that drops one key; it returns the
     completeness finding. Fails at parent because no such seam or oracle exists.
  2. *(Characterization.)* Duplicate findings across array-of-tables elements do not trigger
     the oracle.
* **Size: S | Complexity: medium.**

### RA-3 — Component-wise `Lstat` containment before scanning (`990AFA71`)

* **Files:** `retiredarch.go`, `retiredarch_test.go`.
* **Change:** Add `containedRegularFile(root, rel string) error`. It walks each
  slash-separated component of `rel` from `root` with `os.Lstat`:
  * every intermediate component must be `IsDir()`, with no `os.ModeSymlink|os.ModeIrregular`;
  * the final component must be `Mode().IsRegular()`;
  * a `..`, an absolute path, an empty component or any `Lstat` error is an error.

  `runRepoScan` calls the helper before `scanPath`. On error it emits the synthetic finding
  `<rel>: not a contained regular file: <reason>`, using the same path format as existing
  synthetic findings, and does not scan.
* **Junction semantics:** Go ≥ 1.23 reports `ModeIrregular` for Windows mount points
  (`go.mod`: go 1.24, toolchain go1.26.5). Verify against `$(go env GOROOT)/src/os/types_windows.go`
  before relying on it. Banning both bits also holds under `GODEBUG=winsymlink=0`.
* **Accepted residual (TOCTOU):** the check and the engine's later `os.ReadFile` are separate
  operations. A concurrent writer swapping a component between them is out of model, because
  CI scans a fresh checkout with no concurrent writer. Recorded in the `pin.go` residual
  block by RA-13.
* **Tests (scenarios: 3).** Links and their targets live inside the same `t.TempDir()`.
  1. **Red.** A symlinked final file yields the synthetic finding and never the target's
     content findings. On Linux this must not skip. On Windows it skips only on an
     `os.Symlink` privilege error, with the reason in the skip message. Fails at parent,
     because the target's content is scanned.
  2. **Red.** A symlinked intermediate directory yields the synthetic finding. Same skip rules.
  3. **Red (Windows only).** Two subtests, an intermediate junction and a final-component
     junction, each created with `cmd /c mklink /J "<link>" "<target>"` (quoted paths). Gated
     by `runtime.GOOS == "windows"`. The windows-latest advisory job log is the required
     non-skipped evidence.
* **Size: S | Complexity: medium.**

### RA-4 — Isolate `DefaultGitRunner` (`D7BF9F74`; ALP-2 first half)

* **Files:** `select.go`, `select_test.go`.
* **Change:**
  1. Add an unexported pure function `gitRunnerEnv(environ []string) []string`:
     * drop every entry whose name (the text before the first `=` after position 0) has the
       prefix `GIT_` under ASCII case folding, except exactly `GIT_CEILING_DIRECTORIES`;
     * keep Windows `=C:`-style entries;
     * append `GIT_CONFIG_NOSYSTEM=1`, `GIT_CONFIG_GLOBAL=` + `os.DevNull` and
       `GIT_CONFIG_SYSTEM=` + `os.DevNull` last, because `exec.Cmd` de-duplicates keys and the
       last value wins;
     * preallocate `len(environ)+3`.
  2. `DefaultGitRunner` sets `cmd.Env = gitRunnerEnv(os.Environ())`. It also fails closed,
     returning an error before `Run`, in any of these cases:
     * `cmd.Err != nil` (this covers `exec.ErrDot`);
     * `cmd.Path` is not absolute (this covers `godebug execerrdot=0`);
     * `cmd.Path` lies inside `root`.

     The inside-root test uses `gitPathInside(exe, root string) (bool, error)`:
     * apply `filepath.Abs` and then `filepath.EvalSymlinks` to both paths;
     * compare with `filepath.Rel` (no `..` prefix), case-insensitively on Windows;
     * any canonicalisation error counts as inside, so the guard fails closed.
* **Contract notes:**
  * The isolation set mirrors the established test-isolation set
    (`internal/unignore/testutil_test.go:44-78`, `tools/gatecheck/ciwiring_retire_test.go:313-354`).
    Any difference is documented in a doc comment; nothing is imported across packages.
  * Running the gate *inside a git hook* now ignores the hook's `GIT_DIR`/`GIT_INDEX_FILE`.
    That is intended, and is documented in the doc comment.
* **Tests (scenarios: 3):**
  1. **Red.** `gitRunnerEnv` table test. It drops `GIT_DIR`, `git_index_file` (case),
     `GIT_LITERAL_PATHSPECS`, `GIT_CONFIG_COUNT`, `GIT_CONFIG_KEY_0`, `GIT_CONFIG_PARAMETERS`.
     It keeps `GIT_CEILING_DIRECTORIES`, `PATH` and `=C:=C:\x`, and appends the three
     isolation entries exactly once. Fails at parent because the function is absent; a
     compile failure counts as red.
  2. **Red.** One subtest per vector. Each runs `selectRepoPaths(root, DefaultGitRunner)` on a
     temp fixture git repo that tracks `internal/é.go` (a non-ASCII name) and `cmd/x/main.go`.
     Each sets one vector through `t.Setenv` and asserts the selection equals the
     uncontaminated selection, in which the default `core.quotePath=true` quotes the
     non-ASCII path:
     * (a) `GIT_INDEX_FILE` pointing at an empty temp index;
     * (b) `GIT_LITERAL_PATHSPECS=1`;
     * (c) `GIT_CONFIG_COUNT=1`, `GIT_CONFIG_KEY_0=core.quotePath`, `GIT_CONFIG_VALUE_0=false`;
     * (d) `GIT_CONFIG_PARAMETERS='core.quotepath'='false'`;
     * (e) `GIT_CONFIG_GLOBAL` pointing at a file that sets `core.quotePath = false`.

     Each fails at parent with an empty or different selection. For (c)–(e), the non-ASCII
     path comes back unquoted. Ship confirms each vector's red at parent empirically. If one
     is not observable on the installed git, HALT that vector and record it rather than
     weakening the assertion.
  3. **Red.** An executable named `git` placed in `<root>/bin` (`git.bat` on Windows), with
     `PATH` set to that absolute directory, makes `DefaultGitRunner` return an error. Fails
     at parent because the fake runs.
* **Precheck:** run `git --version` and HALT below 2.32, because `GIT_CONFIG_GLOBAL` would be
  ignored and the isolation would fail open. Confirm `git ls-files -- 'cmd/**' 'internal/**'`
  has no non-ASCII path, so INV-1 holds under default `core.quotePath`.
* **ALP-2:** RA-4 and RA-5 land in **one commit**. RA-4 alone breaks
  `TestSelectionPathspecPin_LiveTree` and `--self-test-integrity`, because select.go's import
  block and `DefaultGitRunner` are token-frozen. The per-unit quality gates for RA-4 are
  therefore evaluated on the ALP-2 commit, and RA-4's red proof is scenarios 1–3 above.
* **Size: S | Complexity: medium.**

### RA-5 — Refreeze canonical text; isolate pin root discovery (`D7BF9F74`; ALP-2 second half)

* **Files:** `pin.go`, `pin_test.go`. **Blocked by:** RA-4; lands in the same commit.
* **Change:**
  1. Update `canonicalDecls` for select.go's import block (adds `"os"`) and
     `DefaultGitRunner`. Add `gitRunnerEnv` to `canonicalDecls`, to `closedWorldDecls`
     (FUNC), and to both `pathspecFrozenDecls` and `prefixFrozenDecls`.
  2. Have `gitShowToplevel` (pin.go:706) set `cmd.Env = gitRunnerEnv(os.Environ())`. The
     `cmd.Path` guard is anchored in two steps, because the root is this command's own output:
     * **before the run:** `cmd.Err == nil` and `cmd.Path` is absolute;
     * **after the run:** `gitPathInside(cmd.Path, returnedTop)` must be false.

     A violation or any error fails closed.
  3. Re-state R-A2 in the header comment:
     * the closed part (env and global/system config);
     * **R-A2a**: PATH resolution of `git` outside the root, plus Go `godebug` settings, which
       is mitigated by the `cmd.Path` guard;
     * **R-A2b**: local repository state (`.git/config`, gitfile, index, `include.path`)
       mutated by earlier same-job or same-process code, as opposed to tracked content;
     * **R-A2c**: `pysem.GitText`/`SplitLines` integrity.
  4. Keep `envMutators` for now; RA-13 subsumes it.
* **Tests (scenarios: 3):**
  1. **Red.** `SelectionPathspecPin(startDir)`, with `GIT_DIR`/`GIT_WORK_TREE` set through
     `t.Setenv` to point at a decoy copy whose select.go is mutated, still evaluates the real
     tree and returns `OK()`. Fails at parent because the decoy is evaluated.
  2. *(Regression guard.)* Reject rows: the `cmd.Env` assignment deleted, `GIT_DIR` added to
     the keep list, `GIT_CONFIG_NOSYSTEM` dropped, the `cmd.Path` guard removed. Each clears
     both flags. These are not red at parent, because the parent rejects any changed
     select.go; the ALP-2 red proof is RA-4's.
  3. *(Characterization.)* `TestSelectionPathspecPin_LiveTree` passes with ALP-2 applied.
* **Size: S | Complexity: medium.**

### RA-6 — Token-freeze `engineForPath` and `scanPath` (`DC921AF6`)

* **Files:** `pin.go`, `pin_test.go`. **Blocked by:** RA-5.
* **Change:** Add independently authored canonical texts for `engineForPath` and `scanPath` to
  `canonicalDecls`. Add both names to the **existing** `pathspecFrozenDecls` and
  `prefixFrozenDecls`; no new set is needed (INV-3).
* **Tests (scenarios: 2):**
  1. **Red.** Reject rows, each of which fails at parent:
     * `.toml` mapped to the fallback engine;
     * a path-conditional `if strings.HasPrefix(path, "internal/x") { return nil }` in
       `scanPath`;
     * an unmapped extension that returns no finding.
  2. *(Characterization.)* The live tree passes.
* **Size: S | Complexity: low.**

### RA-7 — Explicit pin root, pin-test fixture harness, table consistency (`DC921AF6` prerequisite)

* **Files:** `pin.go`, `pin_test.go`. **Blocked by:** RA-6, RA-3.
* **Functions touched:**
  * production: `checkPathspecPin`, `packageClosureOK`, `SelectionPathspecPin`;
  * test: `clonePinFixture`, `TestPinTables_Consistent`.
* **Change:**
  1. **Explicit root (fixes attempt-2 Arch/Go P2).** Change the signature to
     `checkPathspecPin(root string) PathspecPin`. It derives
     `tools/gatecheck/internal/retiredarch/select.go` (and every later sibling or
     package-`main` path) from `root`, and reads each one only after
     `containedRegularFile(root, rel)` (RA-3) succeeds. `packageClosureOK` takes
     `(root string)` the same way.
     * `SelectionPathspecPin(startDir)` resolves `root` **once**, through the RA-5-guarded
       `gitShowToplevel`, and passes it in.
     * No evaluator code calls git or uses a `../..` hop. The live entry point keeps its
       fail-closed behaviour on a root-resolution error.
  2. **Fixture.** Add `clonePinFixture(t, mutate map[string]func(string) string) (root string)`.
     * It builds `root/tools/gatecheck/{every non-test .go of package main}` and
       `root/tools/gatecheck/internal/retiredarch/{every non-test .go}` from the live tree.
     * It then applies the given per-file mutations; a mutation may delete a file or add a
       new one.
     * It returns `root`. The fixture is not a git repository, and does not need to be one.
  3. Migrate every existing `pin_test.go` caller of `checkPathspecPin(path)` to
     `checkPathspecPin(clonePinFixture(t, ...))`. The fixture is a superset, so verdicts are
     unchanged.
  4. Add `TestPinTables_Consistent`, which is table-driven over `(canonical text, closed world,
     frozen sets)` tuples:
     * every `closedWorldDecls` name has exactly one canonical declaration;
     * every name in a frozen set exists in its canonical text;
     * every closed-world FUNC/TYPE/VAR other than the deliberately unfrozen ones is in at
       least one frozen set.

     RA-8, RA-9 and RA-10 each **add a row** for their own canonical table, with no new
     functions.
* **Tests (scenarios: 3):**
  1. *(Characterization.)* All migrated tests still pass, and each reject row first asserts
     that its *unmutated* fixture root is accepted (baseline acceptance).
  2. **Red.** On a fixture root whose `tools/gatecheck/internal/retiredarch` is a directory
     symlink (on Windows, a junction) pointing at a copy outside `root`, `checkPathspecPin(root)`
     returns all-false.
     * **Red proof:** before changing the signature, Ship logs that the parent's path-based
       call, `checkPathspecPin(root+"/tools/gatecheck/internal/retiredarch/select.go")`, returns
       OK through the link.
     * The test is skipped only on a Windows privilege reason, per the skip policy.
  3. **Red.** A mutated-table subtest of `TestPinTables_Consistent` removes one frozen name and
     expects failure. The live tables are green after RA-6.
* **Size: S | Complexity: medium.**

### RA-8 — Unified declaration indexer; cross-file freeze of `runRepoScan` + `containedRegularFile` (`DC921AF6`, `990AFA71`)

* **Files:** `pin.go`, `pin_test.go`. **Blocked by:** RA-7 (and transitively RA-3).
* **Functions touched (production):** `indexDecls` (new), `frozenAgainst` (new),
  `frozenDeclsOK` (becomes a one-line wrapper), `sharedRulesOK` (now uses `indexDecls`).
* **Change:**
  1. **One indexer for both sides.** This fixes the attempt-2 Go P2: `sharedRulesOK` skips
     names outside `closedWorldDecls`. Add
     `indexDecls(file *ast.File) (map[string]ast.Decl, error)`, keyed as follows:
     * `"import"`: the file's single import `GenDecl`. More than one import `GenDecl` is an
       error.
     * `"init"`: the single `func init`. More than one is an error.
     * `"Recv.Name"` for a method `FuncDecl`, otherwise `"Name"`.
     * Each spec name of a `type`/`var`/`const` `GenDecl`, mapped to the enclosing
       `GenDecl`.
     * A duplicate key is an error.

     `indexDecls` indexes **every** top-level declaration and has no closed-world filter.
     `sharedRulesOK` keeps its closed-world check but takes its map from `indexDecls`. This
     refactor is deliberately minimal: no other caller changes.
  2. Add `frozenAgainst(canonSrc string, fset *token.FileSet, src []byte, decls map[string]ast.Decl, names []string) bool`.
     * It parses `canonSrc` and indexes it with the **same** `indexDecls`.
     * It requires every name in `names` to be present on both sides and token-equal
       (`declTokens`).
     * Any parse error, index error or missing name returns false.

     `frozenDeclsOK(...)` becomes `frozenAgainst(canonicalDecls, ...)`.
  3. Read `root/tools/gatecheck/internal/retiredarch/retiredarch.go` through
     `containedRegularFile`, then parse it and run `indexDecls` on it. Require `runRepoScan`
     and `containedRegularFile` to be `frozenAgainst` a new literal `retiredarchCanonicalDecls`.
     A read, parse or index error fails closed.
  4. Fold the result into both flags (INV-3), and add the `retiredarchCanonicalDecls` row to
     `TestPinTables_Consistent`.
* **Tests (scenarios: 3):** all use the RA-7 fixture, and all three fail at parent.
  1. **Red.** Reject: `runRepoScan` skipping paths with a prefix.
  2. **Red.** Reject: the containment call removed, or the helper accepting `ModeSymlink`.
  3. **Red.** Reject: the fixture's retiredarch.go deleted or made unparseable, or given a
     second `runRepoScan` (a duplicate key).
* **Size: M | Complexity: medium.**

### RA-9 — Freeze `Run` and the self-test orchestration chain (`DC921AF6`)

* **Files:** `pin.go`, `pin_test.go`. **Blocked by:** RA-8.
* **Change:**
  * Add canonical texts for `Run` and `runSelfTestAssertions` (retiredarch.go:625/670) to
    `retiredarchCanonicalDecls`.
  * Add a new `selftestSelectionCanonicalDecls` for `runRepoSelectionSelfTest`
    (selftest_selection.go:138), read through containment and parsed with `indexDecls`.
  * Fold both into both flags, and add a `TestPinTables_Consistent` row.
  * **Rationale for freezing the pin's own invoker:** a sabotaged invoker could skip the
    runtime pin, but `go test` (`TestSelectionPathspecPin_LiveTree`) evaluates the pin
    independently of `Run`, so the edit is still caught in CI.
* **Tests (scenarios: 3):** all fail at parent.
  1. **Red.** Reject: `Run` routing repo mode to a no-op, or an extra mode that exits 0.
  2. **Red.** Reject: `runSelfTestAssertions` dropping the integrity assertion's failure
     propagation.
  3. **Red.** Reject: `runRepoSelectionSelfTest` no longer calling `SelectionPathspecPin`.
* **Size: M | Complexity: medium.**

### RA-10 — Freeze CLI wiring; positive reference rule for the dispatch table (`DC921AF6`)

* **Files:** `pin.go`, `pin_test.go`. **Blocked by:** RA-9.
* **Functions touched (production):** `mainWiringOK` (new) and `registryRefsOK` (new). The
  evaluator call site gains one `&&` term.
* **Change:**
  1. **Register-file freeze.** Read `root/tools/gatecheck/register_retired_arch.go` through
     containment, using the explicit `root` from RA-7. Require its `"import"`, `"init"` and
     `runRetiredArch` declarations to be `frozenAgainst` a literal `registerCanonicalDecls`.
  2. **main.go freeze.** Read `root/tools/gatecheck/main.go` through containment. Require its
     `subcommands` var declaration and its `registerSubcommand` func declaration to be
     `frozenAgainst` a literal `mainCanonicalDecls`.
  3. **Positive reference rule.** This replaces the bypassable name-based rule (attempt-2
     Arch and Security P1). Enumerate `root/tools/gatecheck` with `os.ReadDir`. Every
     non-`_test.go` `.go` entry must pass `containedRegularFile` and parse; build constraints
     are ignored, so every file is scanned. Then walk **every** `*ast.Ident` in every file,
     tracking each node's parent:
     * **Outside main.go:**
       * Any `*ast.Ident` named `subcommands`, in any position, fails closed. That covers
         index, alias, `delete`, `maps.Copy`, `range`, being passed as an argument, and
         selector-field use.
       * An `*ast.Ident` named `registerSubcommand` is allowed **only** when all of these
         hold, and fails closed otherwise:
         * it is the `Fun` of an `*ast.CallExpr`;
         * that call's first argument is an `*ast.BasicLit` of kind `STRING`;
         * the call is an `*ast.ExprStmt` directly in the body of a top-level `func init()`.

         This rejects function values (`var reg = registerSubcommand`), local shadows
         (`registerSubcommand := ...`), calls in var initialisers or nested closures, and
         `go`/`defer` forms.
     * **Inside main.go:**
       * `subcommands` may appear only as the declared name in the frozen var spec, inside
         the frozen `registerSubcommand` body, or as the operand of an `*ast.IndexExpr` that
         is the sole right-hand side of a two-value `:=` assignment (`fn, ok := subcommands[name]`,
         main.go:131).
       * `registerSubcommand` may appear only as its own declared name.
       * Any other occurrence fails closed.
     * Across all scanned files, **exactly one** allowed registration call has the literal
       `"retired-arch"`, and it is in register_retired_arch.go with second argument
       `runRetiredArch`.

     This is a syntactic, ident-name rule, and it is conservative on purpose. An unrelated
     local variable that happens to be named `subcommands` outside main.go also fails closed;
     the header documents this. No `go/types` checking is needed, because both names are
     package-scope identifiers of package `main`.
  4. **Header comment** (Scope P3) states:
     * RA-10 constrains **every** subcommand's registration, not only `retired-arch`. A new
       `register_<name>.go` that follows the existing pattern needs no `pin.go` edit; any
       other registration shape requires one.
     * R-A1, now closed except for engine-internal path-conditional sabotage in
       `scanGo`/`scanTomlPrimary`/`scanTomlFallback`, which fixtures and goldens cover.
     * R-A1c: package-`main` dispatch internals other than the frozen declarations and the
       reference rule (main.go `run`/`parseRoot` control flow), the
       `scripts/check-retired-architecture.sh` `--root` derivation, and the `ci.yml` steps.
       These are covered by `check_retired_architecture_wrapper_test.go`,
       `ciwiring_test.go:216`, `main_test.go:359`, `wrapper_text_test.go:167`, and human
       review.
  5. Package `main` files are **read only, never edited**. `_test.go` files of package main
     are out of scope; they are not linked into the binary (R-A3).
* **Tests (scenarios: 3):** all fail at parent. Each row first asserts baseline acceptance of
  the unmutated fixture.
  1. **Red.** Reject:
     * `runRetiredArch` passing a stub runner instead of `retiredarch.DefaultGitRunner`;
     * the retiredarch import aliased or rebound;
     * main.go's `registerSubcommand` body changed to ignore `"retired-arch"`.
  2. **Red.** Reject rows, each in a fixture file `zz_override.go` in package main unless
     stated:
     * (a) `func init(){ registerSubcommand("retired-arch", stub) }` (a duplicate);
     * (b) `var n = "retired-arch"; func init(){ registerSubcommand(n, stub) }`;
     * (c) `func init(){ m := subcommands; m["retired-arch"] = stub }`;
     * (d) `func init(){ maps.Copy(subcommands, map[string]subcommandFunc{"retired-arch": stub}) }`;
     * (e) `var reg = registerSubcommand; func init(){ reg("retired-arch", stub) }`;
     * (f) `func init(){ delete(subcommands, "retired-arch") }`;
     * (g) `var _ = func() int { registerSubcommand("x", stub); return 0 }()`;
     * (h) main.go mutated so that `run` assigns `subcommands["retired-arch"] = stub`.

     At parent, there is no package-main rule at all.
  3. **Red.** Reject:
     * register_retired_arch.go missing, or replaced by a symlink;
     * main.go missing;
     * a non-regular `.go` entry in `tools/gatecheck`.
* **Size: M | Complexity: medium.**

### RA-11 — Closed-world package file and import manifest (`3750C37C`, `0ECC1895`)

* **Files:** `pin.go`, `pin_test.go`. **Blocked by:** RA-10, RA-2.
* **Change:**
  * Add `packageManifest map[string][]string`, a literal mapping every non-test `.go` file in
    the package to its exact sorted import-path set, as of RA-1..RA-10.
  * `packageClosureOK(root)` requires **exact equality**: the file set equals the manifest
    keys, and each file's AST import paths equal its set. Comparison is by path, regardless of
    blank or alias form.
  * All dot imports are banned, not only `os`/`syscall`.
  * As a result, `syscall`, `golang.org/x/sys/...`, `plugin`, `reflect`, `unsafe`, `C`,
    `embed` and `os/exec` outside select.go/pin.go are rejected by construction. No `go list`
    is executed.
* **Tests (scenarios: 3):** scenarios 1 and 2 fail at parent.
  1. **Red.** Reject: a new sibling file `zz_init.go`, and a manifest file deleted.
  2. **Red.** Reject rows:
     * `import "syscall"` with `syscall.SetEnvironmentVariable`;
     * `syscall.NewLazyDLL`;
     * `golang.org/x/sys/windows`;
     * `os/exec` added to ident.go;
     * `import . "os/exec"`;
     * `import _ "embed"`.
  3. *(Characterization.)* The live package passes.
* **Size: S | Complexity: medium.**

### RA-12 — Init-time evaluation rules: no `func init`; initializer-call allowlist (`3750C37C`)

* **Files:** `pin.go`, `pin_test.go`. **Blocked by:** RA-11.
* **Functions touched (production):** `initEvalOK` (new; called from `closureFileOK`) and
  `invokedFuncLit` (new helper).
* **Change:** for every non-test file of the package:
  1. Reject any top-level `func init`.
  2. Find every package-level `var` whose initializer evaluates a call at init: a
     `*ast.CallExpr` anywhere in the initializer, including inside composite literals, index
     expressions and multi-value specs, and inside an **immediately invoked** func literal. A
     non-invoked `*ast.FuncLit` body is skipped, because it does not run at init.
     * Each such var must be in `initCallAllowlist`, keyed by `(file, var)`.
     * Its initializer must be token-equal to the allowlist's literal canonical text, using
       `declTokens`.
     * Candidates for characterization: `ident.go/vocabWords`, `scango.go/goIdentifierRe`,
       `tomlfallback.go/bareKeyRe`.
     * `retiredarch.go/fixtureSuites` is expected to pass under the non-invoked func-literal
       rule. Add it only if characterization proves it is needed.
* **Tests (scenarios: 3):** red rows sit in files whose RA-11 manifest **already** includes
  the needed import (Go P2), so RA-11 cannot be what rejects them.
  1. **Red.** Reject
     `func init() { exec.Command("git","update-index","--force-remove","x").Run() }` in
     **pin.go**, which imports `os/exec`. At parent, only select.go's closed world rejects an
     `init`.
  2. **Red.** Reject rows:
     * `var _ = exec.Command("git").Run()` in pin.go;
     * `var _, _ = os.ReadFile("x")` in retiredarch.go;
     * `var _ = func() int { _, _ = os.ReadDir("."); return 0 }()` in pin.go;
     * the allowlisted `bareKeyRe` initializer rewritten to call another function.
  3. *(Characterization.)* The live tree passes, and every allowlist entry exists.
* **Size: S | Complexity: medium.**

### RA-13 — `os`/`os/exec` selector allowlist; retire `envMutators`; residual header (`3750C37C`)

* **Files:** `pin.go`, `pin_test.go`. **Blocked by:** RA-12.
* **Functions touched (production):** `osExecSelectorsOK` (new; called from
  `closureFileOK`). The `envMutators` table and its check are removed.
* **Change:**
  1. Resolve every `*ast.SelectorExpr` through the file's import map to an import path.
     Selectors on `os` and `os/exec`, *including use as values* (for example
     `var f = exec.Command`), are allowed only per `(file, enclosing top-level decl, selector)`
     from the literal `osExecSelectorAllowlist`. The complete set present today was verified by
     AST-shape grep at `372ab38`:
     * `pin.go`: `os.ReadDir`, `os.ReadFile`, `exec.Command`;
     * `retiredarch.go`: `os.ReadFile`;
     * `select.go`: `exec.Command`.

     Planned additions:
     * RA-3: `retiredarch.go/containedRegularFile` gets `os.Lstat`, `os.ModeSymlink`,
       `os.ModeIrregular`;
     * RA-4/RA-5: `select.go/DefaultGitRunner` and `pin.go/gitShowToplevel` get `os.Environ`;
       `select.go/gitRunnerEnv` gets `os.DevNull`.

     The enclosing-decl keys are filled in from characterization. Any selector found
     elsewhere is a HALT.
  2. Remove `envMutators`; rule 1 subsumes it, and its old reject rows must still pass.
  3. Update the header comment: the closed package, R-A3 (`_test.go` exclusion) kept explicit,
     the RA-3 TOCTOU residual, and **R-A5** (linked-package and module-resolution integrity:
     `pysem`, `gomask`, `go.mod` `replace`/vendor, and `godebug`). RA-14 covers R-A5 at test
     level only.
* **Tests (scenarios: 3):**
  1. **Red.** Reject:
     * `os.WriteFile` inside pin.go's `checkPathspecPin`, which is not frozen;
     * a new unfrozen function `func cleanup() { _ = os.Remove("x") }` in retiredarch.go.

     At parent, only env mutators are checked. Rows must target **unfrozen** declarations,
     so the RA-8/RA-9 freezes cannot be what rejects them.
  2. **Red.** Reject rows:
     * `var f = exec.Command` (a value use) at package level in pin.go;
     * `exec.Command` inside a pin.go function other than `gitShowToplevel`;
     * retiredarch.go's `import "os"` rewritten to `import x "os"` together with
       `x.Chdir("/")`. The manifest compares paths, so it passes that row through.
  3. *(Characterization and regression.)* The live tree passes, every allowlist entry exists,
     and the old `Setenv`/`Unsetenv`/`Clearenv` rows still reject.
* **Size: S | Complexity: medium.**

### RA-14 — Package-level low-level-capability ban and init-time spawn/env scan across gatecheck (`3750C37C`, R-A5)

* **Files:** `initstate_test.go` only (tests domain). **Blocked by:** none.
* **Operator-overridable widening:** this unit reaches beyond retiredarch, into every
  gatecheck package (see deliberation, D-RA-6 addendum 2).
* **Change:** this addresses the attempt-2 Security P2. A function denylist is not enough,
  because `LazyProc.Call`, `SyscallN` and similar calls reach anything.
  1. **Package-level import ban.** In every non-test file of every gatecheck package
     (`readPackages(.., "../..")`), importing any of the following fails, whether the use is
     init-reachable or not:
     * `syscall`;
     * any `golang.org/x/sys/...` path;
     * `unsafe`;
     * `plugin`;
     * `C`.

     Verified at `372ab38`: no non-test gatecheck file imports any of these.
  2. **Directive ban.** Any `*ast.Comment` whose text starts with `//go:linkname` fails. This
     matches the start of the comment only, so prose mentions and string literals do not
     match (INV-5).
  3. **Init-reachable spawn and env-write watch-list.** Add `initSpawnEnvOps` and run it
     through the existing reachability engine (as `initFSOps` does):
     * `os/exec`: `Command`, `CommandContext`;
     * `os`: `StartProcess`, `Setenv`, `Unsetenv`, `Clearenv`.

  Add `TestNoInitTimeSpawnEnvOrLowLevel` and a `_Mutations` table.
* **Tests (scenarios: 3):**
  1. **Red.** Rule 1 and 2 rows: each fails at parent, because nothing scans these imports or
     directives outside retiredarch.
     * `import "golang.org/x/sys/windows"` with `windows.NewLazyDLL("k").NewProc("p").Call()`
       in package b;
     * `import "syscall"` with `syscall.LoadLibrary`/`syscall.SyscallN`;
     * `import "unsafe"`;
     * `//go:linkname` in package main.
  2. **Red.** Rule 3 rows:
     * `var _ = exec.Command("git").Run()` in package b, reachable from a;
     * `func init(){ os.Setenv("GIT_DIR","x") }`;
     * `os.StartProcess` reached through a helper.
  3. *(Characterization.)* The live gatecheck tree is clean. A comment that mentions
     `go:linkname` mid-text, as pin.go does, does not match.
* **Size: S | Complexity: low.**

## Dependency Graph

```text
RA-1 ──► RA-2 ───────────────────────────────────────────────┐
RA-3 ────────────────────────┐                               │
RA-4 ─► RA-5 ─► RA-6 ─► RA-7 ─► RA-8 ─► RA-9 ─► RA-10 ─► RA-11 ─► RA-12 ─► RA-13
       (ALP-2: RA-4 + RA-5 in one commit)
RA-14 (independent)
```

* **Edges** (`blocks`):
  * RA-1→RA-2;
  * RA-4→RA-5→RA-6→RA-7;
  * RA-3→RA-7;
  * RA-7→RA-8→RA-9→RA-10→RA-11;
  * RA-2→RA-11;
  * RA-11→RA-12→RA-13.
* `pin.go`/`pin_test.go` units (RA-5..RA-13) are serial.
* RA-11 follows every production-code unit that changes imports, so the manifest captures the
  final imports. RA-1 may add `sort`; RA-4 adds `os` to select.go. RA-12 and RA-13 add no
  imports.
* **Serial execution order:** RA-1, RA-2, RA-3, RA-4+RA-5, RA-6, RA-7, RA-8, RA-9, RA-10,
  RA-11, RA-12, RA-13, RA-14.
* There are no dependencies on prior shipments. No blocked feature (019, 020, 023–026,
  037–043) is touched. `internal/pathsafe` is not imported.

## Decisions and Rationale

See D-RA-1..D-RA-6 in the source deliberation, which revision 2 extends with D-RA-4 and D-RA-6
addenda. In summary:

| Decision | Chosen | Rejected | Why |
|---|---|---|---|
| TOML desync | Fallback walk plus oracle | State machine | Correct verdicts at lower risk |
| Symlink containment | Local `Lstat` containment | `pathsafe` | No coupling to blocked features |
| git invocation | Env isolation plus exec-location guard | Literal pathspecs; a pinned absolute git path | Globs are required; there is no portable trusted path |
| Wiring integrity | Cross-file token freeze | Behavioural canary | Path-aware sabotage defeats canaries |
| Package closure | Exact allowlist manifest | Denylist; per-file freeze | Compound learning 2026-10-01 |
| Linked-package init spawns / low-level capabilities | Go-test-level guard (RA-14) | Runtime pin | Pinning other packages from retiredarch would couple packages |

## Risks and Caveats

| Risk | Mitigation |
|---|---|
| The `Keys()` ordering premise is unverified, and no prior learning covers it | RA-1 asserts the observed sequence first; the fallback is order-insensitive |
| Oracle false positives on duplicates | Multiset comparison (RA-2 scenario 2) |
| Junction semantics vary by Go version or GODEBUG | Both bits banned; GOROOT source verified; Windows advisory job gives the evidence |
| Removing global config breaks `safe.directory` on a runner with mismatched ownership | Fails closed (INV-2); today's CI uses no container; rollback is reverting ALP-2 |
| `GIT_CONFIG_GLOBAL=NUL` behaviour on Git for Windows | RA-4 scenario 2(e) runs on the windows-latest advisory job |
| Pin churn: every future retiredarch edit needs a `pin.go` update | This is the intended review choke point; the header documents the procedure; the `9FF9EEB4` fixer is warned (see deliberation) |
| Freezing large functions (`runRepoSelectionSelfTest`, ~180 lines) | Accepted; canonical texts stay literal (INV-4); the RA-7 consistency test catches table drift |
| Self-matching bans | INV-5 |
| `pin.go` growth (~25 KB plus the canonical texts) | Accepted; splitting `pin.go` is out of scope |

## Constitution Check

| Principle (constitution.instructions.md) | Mapping |
|---|---|
| I. Safety-First Go | INV-2 (fail closed; `default` cases); explicit error returns; no `unsafe`; golangci-lint warnings are defects (quality gates) |
| II. Test-First Development (NON-NEGOTIABLE) | Every unit has a **Red** scenario that fails at parent; characterization and regression rows are labelled and are not red proof |
| III. Workspace Isolation and Security Boundaries | RA-3 containment; RA-4/RA-5 env isolation; RA-10 reads through containment; test links and targets stay inside one `t.TempDir()` |
| IV. CLI Workspace Containment (NON-NEGOTIABLE) | The gate's scan is confined to the selected tracked regular files under root (ED-11); the git exec-location guard |
| V. Structured Observability | Fail-closed findings carry a stable reason text; `::error::` lines are unchanged |
| VI. Single Responsibility | Each unit owns one concern; pin tables are checked for consistency (RA-7) |
| VII. Destructive Command Approval (NON-NEGOTIABLE) | No destructive commands; rollback is `git revert`, which adds a commit |
| VIII. Explicit Safety Modes | Freeze scope plus `careful` for ALP-2 (Execution Discipline) |
| IX. Git-Friendly Persistence | Plan, deliberation and backlog artifacts are Markdown under git |
| X. Agent Context Efficiency | Each unit fits the 2-hour rule (≤2 files, ≤3 scenarios); M units are flagged |
| XI. Merge Commit History Preservation (NON-NEGOTIABLE) | The shipment PR is merged with a merge commit; the rollback reverts that merge |
| Quality Gates | gofmt, go vet, go test, go build, and golangci-lint per commit (Execution Discipline) |
| Task Granularity (NON-NEGOTIABLE) | All units have <3 files and ≤3 scenarios; ALP-2 is a declared paired landing |
| Development Workflow (single branch, backlog traceability) | One branch, serial execution, no parallel worktrees (P-016); harvested into `.backlogit/` with one task per RA unit |

## Plan Hardening Signals

* **Public API, schema or contract change: present.** This changes a merge-blocking CI gate's
  contract: new fail-closed verdicts (ED-10..ED-14), self-test integrity pin semantics, and
  git invocation semantics.
* **Security, auth, permission or compliance-sensitive behaviour: present.** Gate
  integrity/anti-tamper, env and config isolation, symlink containment.
* **Migration, backfill, destructive or irreversible step: absent.** Every unit is revertible.
* **External integration, operator checkpoint or external dependency: present (minor).** Git
  binary version and behaviour, the GitHub runner environment, Windows junction semantics.
* **High runtime, rollout or rollback risk: present (moderate).** A false-positive gate failure
  blocks every PR to `main`.

Requires plan hardening: yes

## Runtime Verification and Closure

* **Runtime surface:** the `gatecheck retired-arch` CLI, invoked by
  `scripts/check-retired-architecture.sh` in CI (`ci.yml:325`, `:332`) and locally. There are
  no product runtime surfaces.
* **Per unit:** the Execution Discipline quality gates plus the INV-1 check.
* **Shipment level:**
  * the PR CI run shows the retired-arch steps green on ubuntu-latest;
  * the windows-latest advisory job shows RA-3 scenario 3 and RA-4 scenario 2 executed, not
    skipped.
* **Closure artifact:** the post-merge closure record states:
  * the status of R-A1..R-A5, ED-10..ED-14, the rollback trigger and the validation window;
  * a compound-learning candidate on maintaining pin canonical texts (a gap in the
    learnings library).

<!-- plan-hardening-start -->
## Plan Hardening

**Hardening required:** yes. The signals are contract, security, external dependency and
rollout risk.

**Learnings consulted:**

* 2026-10-01 denylist bypass loop: allowlists only; wiring pinned positively.
* 2026-09-28 heredoc extraction: every init-time evaluation position (RA-12 rule 2).
* 2026-09-06 self-matching grep (INV-5).
* 2026-09-08 GOROOT verification, and adversarial review on exact shapes (RA-3
  intermediate vs. final junction).
* 2026-09-30 Go/MSYS bridging.
* 2026-10-02 git show-ref ambiguity, and the backlogit stale-binary-on-PATH learning
  (R-A2a, the `cmd.Path` guard).
* workflow-issues: staging-premise (every red test fails at parent; RA-4 scenario 3 redesigned)
  and cross-artifact closure (the RA-7 consistency test).

**Instruction files:** `.github/instructions/constitution.instructions.md`, the Go
instructions in `.github/instructions/`, and `concurrency.instructions.md`.

### Protected invariants

INV-1..INV-5, plus:

* **H-INV-6:** the `--self-test-integrity` and repo goldens are never regenerated to make a unit
  pass. A diff means HALT.
* **H-INV-7:** no unit edits scan scope, forbidden-token data, `.github/workflows/**`,
  `scripts/**` or package `main`.

### Risky actions

| ProposedAction | ActionRisk | Approval | Notes |
|---|---|---|---|
| PA-1: change runner env and exec guard, refreeze `canonicalDecls`, isolate `gitShowToplevel` atomically (RA-4+RA-5, ALP-2) | **high** | Normal PR review; both halves visible in one commit | A non-atomic landing breaks `main`'s integrity self-test |
| PA-2: replace the desync parse-error path with the fallback walk (RA-1) | **medium** | Normal PR review | Loosens the verdict only for previously erroring shapes; guarded by the subset check and RA-2 |
| PA-3: cross-file canonical texts and package-main registration rule (RA-8..RA-10) | **medium** | Normal PR review | Adds layout dependencies; any break fails closed |
| PA-4: remove `envMutators` in favour of the selector allowlist (RA-13) | **medium** | Normal PR review | Same commit as the allowlist; old reject rows must still pass |
| PA-5: any golden-file edit | **high** | HALT, return to Stage | Not planned |

### Verification additions

* **Environment precheck (Ship, before RA-4):**
  * `git --version` is ≥ 2.32, else HALT.
  * `go env GOVERSION` is ≥ 1.23.
  * `git ls-files -s | Select-String '^120000'` is empty (the ED-11 premise).
  * No non-ASCII path under `cmd/**` or `internal/**`.
* **Red-at-parent log:** one entry per **Red** scenario, giving the test name and the failure
  reason at the parent commit.
* **Skip policy:** on ubuntu-latest, any skip in RA-3 scenarios 1–2 or RA-4 counts as a
  failure. On Windows, skips are allowed only with a privilege reason. RA-3 scenario 3 is
  Windows-gated, and the windows-latest log must show it executed.
* **Atomicity:** `git show --stat` of the ALP-2 commit lists select.go, select_test.go,
  pin.go and pin_test.go.

### Operational closure

* **Monitoring signal:** the CI retired-arch steps (`ci.yml:325`, `:332`) on the shipment PR
  and on the next five PRs to `main`.
* **Rollback trigger:** a retired-arch CI failure on an unrelated PR that shows any of:
  * `::error::git ls-files`;
  * a git exec-location refusal;
  * `not a contained regular file`;
  * `TOML completeness check failed`;
  * a cleared `PathspecOK`/`PrefixOK` with no pin-relevant diff.
* **Rollback procedure:** revert the shipment merge with `git revert -m 1 <merge>`. For an
  ALP-2-only incident, revert just that commit. Never revert RA-4 without RA-5.
* **Owner:** Ship for the PR; the operator for post-merge rollback.
* **Validation window:** the next five PRs to `main` or 7 days, whichever is longer.

### Review-gate capability risks

* Plan review must emit literal `dispatch_mode:` and `decision:` lines.
* Route the Architecture Strategist through `model_routing.anchor_review`.
* Trigger the Security Lens Reviewer.
* Declare any same-model fallback (P-012).

### Unresolved operator decisions

None blocking. The operator may override any of the following at staging-PR review:

* D-RA-1 option C;
* the D-RA-3 residuals;
* the RA-14 widening: a package-level ban across **all** gatecheck packages, and a test-level
  scope for R-A5;
* RA-10's constraint on **every** subcommand's registration shape.
<!-- plan-hardening-end -->

## Plan Review — attempt 1 (revision 1)

dispatch_mode: multi-agent
decision: FAIL

* **Personas.** Six ran as subagents:
  * Architecture Strategist, on the anchor route `gpt-6.1-sol`: FAIL;
  * Security Lens Reviewer: FAIL;
  * Go Reviewer: FAIL (P1×2);
  * Scope Boundary Auditor: ADVISORY;
  * Constitution Reviewer: ADVISORY;
  * Learnings Researcher: ADVISORY.
* **P1 findings, all remediated in revision 2:**

| Finding | Remediation |
|---|---|
| Register-file import rebinding is not frozen | RA-10 |
| The `Run` → `runSelfTestAssertions` → `runRepoSelectionSelfTest` chain is unpinned | RA-9 |
| `gitShowToplevel` is not isolated | RA-5 |
| The pin-test temp-dir harness makes reject rows vacuous | RA-7 |
| The freeze helpers cannot index other files' declarations | RA-8 helpers |
| A duplicate `retired-arch` registration in package main | RA-10 rule |
| Linked-package init spawn (Security P1) | RA-13, plus residual R-A5 |

* **P2 findings remediated:**

| Finding | Remediation |
|---|---|
| RA-4 scenario 3 is green at parent | Redesigned with `core.worktree` vectors |
| ALP-2 granularity | Paired verification declared |
| The RA-2 seam is a package var | Now a parameter |
| RA-10 func-literal over-breadth and an already-red row | Rule refined; rows replaced |
| Constitution Check section missing | Added |
| Quality-gate sequence missing | Added |
| Allowlist keyed by file only | Now keyed by (file, decl, selector) |
| Open characterization clause | Enumerated |
| Init evaluation positions | RA-12 rule 2 |
| Pin cross-artifact consistency | RA-7 |
| RA-3 junction shapes | Split into intermediate and final subtests |
| Fallback `default` and the desync subset check | RA-1 |
| Residual scoping (R-A2a/b, R-A3) | RA-5, RA-12 |

* **P3 findings absorbed:**
  * `dispatchFrozenDecls` dropped;
  * exact manifest equality, with no `go list`;
  * all dot imports banned;
  * skip-by-GOOS;
  * quoted `mklink` paths;
  * hook-caveat documentation;
  * TOCTOU residual;
  * the single-branch statement;
  * the P-021 extended duplicate scan, still CLEAN;
  * the late thread ID for `990AFA71`, recovered (see deliberation).

## Plan Review — attempt 2 (revision 2)

dispatch_mode: multi-agent
decision: FAIL

* **Personas.** Four ran as subagents:
  * Architecture Strategist, on the anchor route `gpt-6.1-sol`: FAIL;
  * Security Lens Reviewer: FAIL;
  * Go Reviewer: ADVISORY;
  * Scope Boundary Auditor: ADVISORY.
* **P1, raised by both Architecture and Security, remediated in revision 3.** RA-10's
  registration rule was name-based. It could be bypassed through an alias index
  (`m := subcommands; m[..]=..`), through `maps.Copy(subcommands, ..)`, or through a function
  value bound to `registerSubcommand`. **Fix:** RA-10 now applies a positive every-reference
  rule, freezes main.go's `subcommands` and `registerSubcommand`, and adds reject rows (a)–(h).
* **P2 findings, remediated:**

| Finding | Remediation |
|---|---|
| (Security) RA-13 function denylist bypassable through `LazyProc.Call`/`SyscallN` | RA-14 package-level import ban, plus a `go:linkname` ban |
| (Arch/Go) RA-10 `gitShowToplevel` lookup unreachable from a non-git fixture | Explicit `root` (RA-7) |
| (Go) Indexer asymmetry: `sharedRulesOK` filters to the closed world | `indexDecls`, with typed keys (RA-8) |
| (Go) RA-4 `core.worktree` vectors not observable | `core.quotePath` vectors on a tracked non-ASCII path |
| (Go) RA-12 rows already rejected by the RA-11 manifest, and the `init` row in select.go | Rows moved to pin.go and retiredarch.go |
| (Go) RA-1 key-only matching | Stated, with scalar cases |
| (Scope) RA-12 size | Split into RA-12 and RA-13 |

* **P3 findings absorbed:**
  * guard anchor and canonicalisation (`gitPathInside`);
  * the `tomlWalker` out-parameter;
  * the RA-6 row parameter name;
  * consistency-test coverage of the RA-8/9/10 tables;
  * the pin header note on other subcommands;
  * RA-14 marked as an operator-overridable widening;
  * a minimal RA-8 refactor.

<!-- plan-review-attempt: 2 -->

## Plan Review — attempt 3 (revision 3, FINAL; re-entry budget exhausted)

dispatch_mode: multi-agent
decision: FAIL

* **Personas.** Four ran as subagents:

| Persona | Route | Verdict |
|---|---|---|
| Architecture Strategist | anchor `gpt-6.1-sol` | FAIL (3 P1) |
| Security Lens Reviewer | default | FAIL (2 P1, 4 P2, 2 P3) |
| Go Reviewer | default | FAIL (1 P1, 4 P2, 3 P3) |
| Scope Boundary Auditor | default | ADVISORY (1 P2, 6 P3) |

* **Resolved from attempt 2:**
  * the RA-10 aliasing P1;
  * the RA-14 package-level ban (except for assembly);
  * RA-1 key-only matching and scalar cases;
  * the RA-6, RA-9, RA-10 (rows), RA-13 and RA-14 red rows (valid at their parents);
  * the RA-12 split;
  * the RA-14 widening disclosure.
* **New P1 findings:** these are not remediated. The plan-review cycle cap is reached, so the
  escalation protocol was invoked (see the Stage memory file).

| ID | Persona | Unit | Finding | Candidate fix |
|---|---|---|---|---|
| A3-P1-1 | Arch | RA-4/RA-5 | The new select.go helper `gitPathInside` is missing from the `closedWorldDecls`, canonical and frozen-set updates, so `sharedRulesOK` rejects the live tree and ALP-2 cannot pass | Add it to the ALP-2 refreeze, or host it outside select.go |
| A3-P1-2 | Arch (Security P2) | RA-4/RA-5 | `filepath.Abs` (and `EvalSymlinks`) in `gitPathInside` is reachable from the registered `runRetiredArch`. The existing `TestNoInitTimeProcessStateReads`/`initFSOps` watch lists reject it, so ALP-2 turns an existing security test red | **Design decision needed.** Either drop the inside-root exec guard (keep `cmd.Err` and absolute `cmd.Path`, and fold the rest into R-A2a), or do canonicalisation-free containment on absolute inputs, or make a scoped reachability change (must not weaken) |
| A3-P1-3 | Arch (Go P2) | RA-8 × RA-12 | Row `var _, _ = os.ReadFile("x")` is already rejected at its parent, because `indexDecls` sees a duplicate `_` key | Leave blank identifiers out of the `indexDecls` keys, and use named, unfrozen vars in red rows |
| S3-P1-1 | Security | RA-10 | The registration literal comparison is not specified. A raw-string or escaped spelling (`` `retired-arch` ``, `"retired\x2darch"`) in a later-sorted file overwrites the entry | `strconv.Unquote` every literal; require unique names across all registrations; add rows |
| S3-P1-2 | Security | RA-10/RA-14 | A `.s`/`.syso` file plus a bodyless `func` in package `main` can rewrite `·subcommands(SB)` with no Go reference | `ReadDir` extension allowlist (regular `.go` only, build tags ignored); reject bodyless top-level `FuncDecl`s in every gatecheck package; add rows |
| G3-P1-1 | Go | RA-4/RA-5 | On Windows, `filepath.Rel` errors across volumes (git on `C:`, checkout on `D:` on `windows-latest`), so treating the error as inside fails every git call there | Compare volumes first (differing volumes mean not inside); only Abs/EvalSymlinks errors fail closed; treat as inside only when `rel` is `.` or does not start with `..`+sep; add rows |

* **P2 findings (open):**
  * `sharedRulesOK` must keep its per-decl shape and kind loop (G3-P2);
  * the RA-8 "unparseable" row is a regression guard, not red;
  * RA-4: assert the exact clean selection `["cmd/x/main.go"]`, and record new residual
    **R-A2d**: git-quoted paths under `cmd/`/`internal/` are silently unscanned (a candidate
    new stash entry, out of this batch's scope);
  * `gitShowToplevel` decoy-top check: require `returnedTop` to be an ancestor of `startDir`;
  * the init-time `os.Exit`/`os.Executable` conditioning vector in sibling packages;
  * package `main` import allowlist, or an explicit R-A5 widening;
  * RA-7 is at the function limit (split out `TestPinTables_Consistent`).
* **P3 findings (open):**
  * reflection residual note;
  * `_x.go`/`.x.go` parse policy;
  * RA-4 vector details (`GIT_INDEX_FILE` pointing at a nonexistent path; the
    `GIT_CONFIG_PARAMETERS` literal);
  * RA-5 decoy must be a real repository;
  * RA-11 rows in non-indexed files;
  * RA-8 and RA-9 function counts;
  * RA-10 constraint on future main.go edits;
  * R11 wording;
  * the RA-4 halt criterion;
  * RA-7 consistency-test acceptance-criteria scope.
* **Outcome:** FAIL after three attempts (two re-entry cycles), so per Stage Step 4 the
  escalation protocol applies.
  * No harvest, shipment or stash archival happened.
  * The escalation route resolved to tier3 (`gpt-6.1-sol`/openai/xhigh), which is not the
    same route as Stage's own.
  * The payload was handed off through the engram-ingested memory document.
  * Stage halted for an operator decision.

<!-- plan-review-attempt: 3 -->

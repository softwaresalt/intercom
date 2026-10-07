---
title: "retiredarch correctness: TOML desync fallback, tracked-symlink containment, git env isolation"
date: 2026-10-07
status: reviewed
review_verdict: ADVISORY
review_attempts: 2
revision: 3
source: docs/decisions/2026-10-07-intercom-go-retiredarch-gate-integrity-deliberation.md
source_section: "D-RA-7 — Escalation disposition: Option B, split the batch"
supersedes_scope_of: docs/plans/2026-10-07-intercom-go-retiredarch-gate-integrity-plan.md
stash_ids: ["9FC28DB9", "990AFA71", "D7BF9F74"]
related_stash_ids: ["4537B2F6", "D44D8BDF"]
deferred_stash_ids: ["DC921AF6", "3750C37C", "0ECC1895"]
go_floor: "1.24"
---

# retiredarch correctness plan (Shipment 1 of the Option B split)

## Revision History

| Rev | Date | Change |
|---|---|---|
| 1 | 2026-10-07 | Initial narrowed plan (U1..U6) from D-RA-7. Plan-review attempt 1: FAIL (see "Plan Review — attempt 1"). |
| 2 | 2026-10-07 | Addresses attempt-1 findings: dependency-ordered rollback (AS-1); native launchable fake-git helper with a positive parent-red (AS-2/G-1/CR-2); follow-up stash entries created and cited (CR-1: 4537B2F6, D44D8BDF); untested fail-closed branches either removed or given red rows (CR-3/SB-1/SB-2/SB-3); depth-3 and array-of-tables shapes (CR-9); mklink argv (G-2); PATH="." (G-3); git >= 2.32 and `$GITHUB_PATH` stated in R-A2 (S1/S3/CR-8); `GIT_*` inventory precheck (L-2); harvest mapping (CR-7). Rev-1 text was replaced in place (it was never committed); the attempt-1 review record below is preserved verbatim in substance. Plan-review attempt 2: ADVISORY (no P1). |
| 3 | 2026-10-07 | Folds the attempt-2 ADVISORY findings in place. This is NOT a new review cycle; the ADVISORY verdict stands. Changes: U3 rejection rows become direct `containedRegularFile` calls plus a Windows junction-as-final row (AS2-1, S2-3, CR2-3); fake git written 0o755 with a LookPath precondition and an argv-gated `TestMain` (G2-1, S2-1); U5 staged freeze red and full-field reject assertions (CR2-2, G2-2); U6 expected-root derivation (G2-3); interleaved desync rows (G2-4); PA-2 anchor list (G2-5); pin.go comment refresh (G2-6); R-A2f safe.directory and `GIT_INDEX_FILE` ordering (G2-7); Constitution Check renumbered to the real principles (CR2-1); full quality gates (CR2-4); helper-placement rationale (CR2-5); `GIT_TERMINAL_PROMPT` removed (SB2-1); `cmd.Err` characterization row (SB2-2); test-only fallback seam note (SB2-3); R-A2c restated (S2-2); Git Bash first on PATH for package-wide runs (L2-1); symlink creation is fatal on non-Windows (L2-2); stdlib evidence citations (L2-3). |

## Problem Frame

Three defects in the retired-architecture gate engine (`tools/gatecheck/internal/retiredarch`):

1. **9FC28DB9 — TOML cursor desync (false positive).** `walkTable` (tomlprimary.go:148) walks the decoded TOML map in document order using a flat `MetaData.Keys()` cursor. Some legal TOML shapes (an ancestor table reopened after a descendant, non-contiguous sibling tables) desync the cursor. `scanTomlPrimary` (tomlprimary.go:306-310) then emits `TOML parse error (fail-closed): ... cursor desync ...`, so valid TOML fails the gate. The finding is fail-closed, so it is a **false positive**, not a false-clean (D-RA-7 correction). The defect is that it blocks valid manifests, and that a future cursor bug could also drop findings silently on the success path.
2. **990AFA71 — tracked symlink read-through.** `runRepoScan` (retiredarch.go:597-599) joins each git-selected path lexically under `root` and reads it. A tracked symlink (or a Windows junction in an intermediate directory) makes the gate scan bytes outside the selected tracked file, so a forbidden reference can be hidden behind a link whose target is clean.
3. **D7BF9F74 — git environment inheritance.** `DefaultGitRunner` (select.go:32-46) and `gitShowToplevel` (pin.go:706-723) inherit the full process environment. `GIT_INDEX_FILE`, `GIT_LITERAL_PATHSPECS`, `GIT_DIR`/`GIT_WORK_TREE`, `GIT_CONFIG_*` and global/system config can change the selection (including to empty, which is a false-clean) or redirect the pin's root. A relative `git` resolved through `.` on PATH (when `GODEBUG=execerrdot=0`) is also accepted.

### Trust-root constraints (verified in code at main@372ab38)

- `select.go` is a frozen, closed-world file. `pin.go` holds `canonicalDecls`, `closedWorldDecls` (pin.go:166), `confinedIdentDecls` (181), `pathspecFrozenDecls` (193) and `prefixFrozenDecls` (197). `sharedRulesOK` (555) fails closed on any undeclared or duplicate top-level declaration. Any new select.go declaration must be added to `closedWorldDecls` and, where frozen, refrozen in `canonicalDecls` in the SAME commit (ALP pattern).
- The `"import"` decl is in BOTH frozen sets. Adding `"os"` therefore refreezes the import text for PathspecOK and PrefixOK together.
- `envMutators` (pin.go:226) bans `os.Setenv`/`Unsetenv`/`Clearenv` package-wide. `os.Environ` and `os.DevNull` are permitted. The init-time security tests (initstate_test.go, and the pin_test.go:425 reject row) must remain green unchanged.
- The memory note "scanFile must pass ReadText output to scanSource" belongs to **writepath**, not retiredarch. Here, `scanTomlPrimary` reads via `pysem.ReadText` and its error format is part of the Python-parity contract. Do not change it.
- All test files are `package retiredarch` (internal). There is no `TestMain` in the package today.
- Local git is 2.55.0.windows.5. CI uses ubuntu-latest (git >= 2.4x).

## Requirements Trace

| Req | Stash | Requirement |
|---|---|---|
| R1 | 9FC28DB9 | A desync-shaped but valid TOML document produces the correct key-path findings (or zero), never a desync parse error. |
| R2 | 9FC28DB9 | On the cursor-walk success path, a walk that drops or invents findings fails closed with a distinct finding. |
| R3 | 990AFA71 | A selected path whose final component is a symlink or irregular file, or any of whose intermediate components is a symlink, junction or non-directory, yields a fail-closed synthetic finding and is not read. |
| R4 | D7BF9F74 | `DefaultGitRunner` runs git with `GIT_*` stripped (except `GIT_CEILING_DIRECTORIES`) and with system/global config disabled, and refuses a non-absolute resolved git path. |
| R5 | D7BF9F74 | The select.go pin is refrozen to the new texts. The pin still rejects a mutation of the new isolation code. |
| R6 | D7BF9F74 | `gitShowToplevel` gets the same isolation and the same non-absolute refusal. |

## Out of Scope

- `git ls-files` without `-z` / quoted paths silently unscanned → **stash 4537B2F6** (HIGH, separate contract surface; residual R-A2d).
- Cross-engine (writepath / unignore / mergestrategy) symlink containment → **stash D44D8BDF** (LOW; residual R-C1).
- The inside-checkout git location guard (`gitPathInside`). Dropped per D-RA-7; residual R-A2a.
- CLI registration / declaration-index hardening (old RA-6..RA-14 class) and everything from DC921AF6, 3750C37C and 0ECC1895. These return to deliberation with a bounded threat model (D-RA-7).
- The D-RA-3 addendum-2 "root-once evaluator" refactor. Not adopted.
- Any change to `pysem` or to the Python-parity error formats other than the two new finding texts below.

## Invariants

- **INV-1** No new third-party dependency. Stdlib only, Go 1.24 floor (`t.Chdir`, `t.Setenv` are available).
- **INV-2** Existing finding texts are unchanged. The only new texts are:
  - `"%s: TOML completeness check failed: cursor walk omitted or invented findings (fail-closed)"` (U2)
  - `"%s: not a contained regular file: %s (fail-closed synthetic finding)"` (U3)
- **INV-3** Frozen declarations change only through the ALP-2 atomic commit (U4+U5). init-time security tests stay green untouched.
- **INV-4** The fallback TOML walk is deterministic (keys sorted bytewise) and authoritative whenever the cursor walk errors.
- **INV-5** No `filepath.Abs` / `filepath.EvalSymlinks` is introduced on paths reachable from registered runners (initstate constraint). Containment uses `os.Lstat` per component only.

## Execution Discipline

- Test-first. Each unit's red scenarios are written and run red against the parent (the commit before the unit) BEFORE the implementation, and the red output is logged with the platform and the parent SHA. Reds that can only be produced on Linux (unprivileged Unix symlinks) may carry HEAD-only evidence from Windows, plus a note (CR-4). Ship's machine is Windows.
- Scenario ordering for compile-red avoidance (G-5): when a scenario needs a new symbol, write the scenarios that do NOT need new symbols first and run them red. Then add the symbol-dependent scenario.
- Use only `go test ./tools/gatecheck/internal/retiredarch/... -run <Name>` while a unit is in flight. Run the package-wide suite at each unit's completion. For U4 that is deferred to U5 (ALP-2).
- If a shell is needed on Windows, use Git Bash by absolute path (`C:\Program Files\Git\bin\bash.exe`) and log `command -v bash`. Never use the bare `bash` WSL shim (L-1). Before ANY package-wide or `./tools/gatecheck/...` / `./...` run on Windows, prepend `C:\Program Files\Git\bin` to PATH and log `where bash`. `check_retired_architecture_wrapper_test.go` launches bare `bash` (L2-1).
- Quality gates at U2, U3, ALP-2 (U4+U5) and U6 completion are the constitution's full set: `gofmt -l .` (empty), `go vet ./...`, `go build ./...`, `go test ./...` (CR2-4). Narrow `-run` commands are for in-flight work only.
- Stdlib evidence (L2-3), against Go toolchain 1.26.5 with `go.mod` directive `go 1.24` (which sets the `winsymlink`/`execerrdot` defaults):
  - `os/exec/lp_unix.go:22-41`: an execute bit is required.
  - `os/exec/lp_unix.go:63-69` and `lp_windows.go:150-157`: relative resolution via `.`, and `execerrdot`.
  - `os/types_windows.go`: a junction is a name surrogate, so it gets `ModeIrregular`, or `ModeSymlink` under `winsymlink=0`. U3 rejects both.
  - `os/exec/exec.go`: `PWD` is added only when `Env` is nil.
  - BurntSushi/toml v1.6.0 `decode.go`: the decoded scalar set.

  Ship re-cites file:line from its local GOROOT in the run log.

## Implementation Units

### U1 — Deterministic fallback TOML walk on desync (9FC28DB9)

Files: `tomlprimary.go`, `tomlprimary_test.go`. Size S | Complexity medium.

- Add `walkDecodedTOML(posixPath string, prefix []string, value map[string]interface{}, findings *[]string) error`. It visits every key exactly once per table instance, with keys sorted bytewise:
  - For each key, report once via `matchesForbiddenParts(composeTomlParts(keyPath))` and `reportTomlKey`. This is key-path-only matching, the same predicate as walkTable:189.
  - Recurse by explicit type switch: `map[string]interface{}`; `[]map[string]interface{}` (each element walked with the SAME prefix); `[]interface{}` (each element with the same prefix; nested tables and arrays recurse).
  - Explicit scalar cases `string`, `int64`, `float64`, `bool`, `time.Time` (BurntSushi v1.6.0's complete decoded scalar set, verified) are no-ops.
  - `default:` returns an error `retiredarch: TOML fallback walk: unexpected decoded type %T at %v` (fail-closed).
- In `scanTomlPrimary`, when `walkTable` returns an error, discard the partial cursor findings and run `walkDecodedTOML` from the root. If the fallback succeeds, return its findings. If it errors, return the existing fail-closed parse-error format with the fallback's error. (The rev-1 subset check of partial findings is removed: SB-1/CR-3. The fallback is authoritative per INV-4, and U2 catches inventions on the success path.)
- Red/characterization scenarios (3):
  1. **Desync shapes (red).** Table rows, each with a forbidden token at a key under the reopened or ancestor table, and each row's mirror without the token:
     - ancestor reopen `[a.b]` … `[a]`
     - non-contiguous siblings `[a.x]` … `[b]` … `[a.y]`
     - depth-3 `[a.b.c]` / `[a]` / `[a.b.c.d]`
     - array-of-tables reopen `[[t]]` … `[t.sub]` … `[[t]]`
     - interleaved-sibling variants (G2-4), likely the real reds, since `peekSelf` and `walkArrayOfTables` already handle the plain reopens: `[a.b]` / `[c]` / `[a]` and `[[t]]` / `[u]` / `[t.sub]`

     Each row first asserts that `walkTable` returns its desync error on that document (SB-8). A row that does not desync at the parent is moved to scenario 2 and recorded. Expected result: the token row gives exactly the key-path finding(s), and the mirror gives zero findings. The parent gives the `TOML parse error (fail-closed)` desync text, which is red.
  2. **Characterization.** On desync-free documents covering every decoded kind (all scalars, inline table, array of inline tables, array of arrays, array-of-tables, dotted keys), `walkDecodedTOML` findings equal `walkTable` findings as multisets.
  3. **Fallback default (red, direct call).** `walkDecodedTOML` given a map containing a Go `int` value, and given one containing a `[]string` value, returns the "unexpected decoded type" error.

### U2 — Completeness oracle on the cursor success path (9FC28DB9)

Files: `tomlprimary.go`, `tomlprimary_test.go`. Size S | Complexity medium. Blocked by U1.

- Refactor: `scanTomlPrimary(path)` becomes a one-line wrapper over unexported `scanTomlPrimaryWith(path string, walk, fallback func(string, []string, map[string]interface{}, *tomlCursor, *[]string) error)`. To keep the seam to a single func type, the fallback is adapted with a closure that ignores the cursor. No named seam type (SB-4). The `fallback` parameter is a test-only seam that exists solely for scenario 3. Production always passes the adapted `walkDecodedTOML` (SB2-3).
- On cursor success, also run the fallback:
  - If the fallback errors → return the existing parse-error format with that error.
  - If the cursor and fallback findings differ as multisets → return the single U2 completeness finding (INV-2).
  - Otherwise → return the cursor findings in cursor order (Python-parity ordering preserved).
- Red scenarios (3), all via `scanTomlPrimaryWith` on one fixture with two forbidden keys:
  1. A walker that drops one finding → completeness finding.
  2. A walker that appends an invented finding → completeness finding.
  3. A fallback that returns an error while the walker succeeds → fail-closed parse-error text.

  At the parent, none of these compile until the seam exists. Their red is the assertion failure obtained by first adding the seam with the pass-through behavior (seam-only step), then the oracle (G-5).

### U3 — Component-wise containment in runRepoScan (990AFA71)

Files: `retiredarch.go`, `retiredarch_test.go`. Size S | Complexity medium. Independent.

- Add `containedRegularFile(root, rel string) (ok bool, reason string)`:
  - Split `rel` on `/`. Reject an empty, `.` or `..` component as a containment violation. On Windows, also reject a component containing `\` or `:` (S2-3).
  - For each intermediate component, `os.Lstat(root/…prefix)` must report `IsDir()` with no `ModeSymlink|ModeIrregular`.
  - The final component must report `Mode().IsRegular()`.
  - **Any Lstat error returns ok=true (fall through)**, so a deleted or unreadable tracked file still hits the existing `scanPath` read-error text (SB-3, ED-11 unchanged). Only type violations produce the new finding. Stdlib only (INV-5).
- In `runRepoScan`, when `!ok`, append the U3 synthetic finding (INV-2) and skip `scanPath` for that path.
- Red scenarios (3), each building a temp git repo and using a GitRunner stub that returns the selected path list. The real index is not needed for containment.
  1. **Final-component link.**
     - Symlink subtest: `internal/x.go` is a symlink to a file outside root containing clean Go. The parent returns Code 0 (read-through, clean) → red. Expected: Code 1 with the U3 finding. Create the link with `os.Symlink`. Only on Windows may a privilege error `t.Skip`. On every other OS, a creation failure is `t.Fatal`, so Linux CI can never silently skip this red (L2-2).
     - Windows junction subtest (CR2-3): `internal/x.go` is a junction to an outside directory. At the parent it hits the existing directory read-error text. Expected: the U3 text instead. This red is observable on Ship's Windows machine.
  2. **Intermediate link.**
     - Unix subtest: `internal/d` is a symlink to an outside dir.
     - Windows subtest: `internal/d` is a junction made with `exec.Command("cmd", "/c", "mklink", "/J", link, target)` (plain argv, no embedded quotes; G-2).

     The outside dir holds clean `y.go`. The parent returns Code 0 → red. Expected: the U3 finding. Junctions report `ModeIrregular` under Go ≥1.23 (verified).
  3. **Table, direct `containedRegularFile` calls (AS2-1).** These paths never pass `shouldScanRepoPath`, so they are unit rows rather than repo-scan rows.
     - `a//b.go`, `./a.go` and `../x.go` give `ok=false`.
     - On Windows, a component containing `\` or `:` gives `ok=false` (S2-3).
     - Plus one repo-scan characterization row: a selected but deleted file keeps the existing read-error text byte-for-byte.

### U4 — Git env isolation and absolute-path guard in DefaultGitRunner (D7BF9F74)

Files: `select.go`, `select_test.go`. Size M | Complexity medium (exception noted: one table plus one repo test plus one guard test, ~1.5 h). Lands in the SAME commit as U5 (ALP-2). Package-wide tests are red between U4 and U5 by design (pin mismatch). During U4, run only `-run 'TestGitRunnerEnv|TestDefaultGitRunner'`.

- Precheck (L-2): inventory every `GIT_*` setter and every caller of `DefaultGitRunner`/`gitShowToplevel` under `tools/gatecheck/**` and `scripts/check-retired-architecture.sh`, and classify each as kept / dropped-safe / HALT. Stage's read at main@372ab38:
  - `check_retired_architecture_wrapper_test.go:228` sets `GIT_CEILING_DIRECTORIES` → **kept** (allow-listed).
  - `ciwiring_retire_test.go:313-335` and `unignore/testutil_test.go:44-99` build their own child envs → **unaffected**.
  - `retiredarch_test.go` / `selftest_selection_test.go` / `select_test.go` call `DefaultGitRunner` on the real checkout → **dropped-safe** (they set no `GIT_*`).

  Any finding not in this list → HALT and report.
- Add `"os"` to the import. Add top-level `func gitRunnerEnv(environ []string) []string`:
  - Drop every entry whose name (up to the first `=`, ASCII-case-folded) starts with `GIT_`, except `GIT_CEILING_DIRECTORIES`.
  - Keep Windows `=C:`-style entries.
  - Append `GIT_CONFIG_NOSYSTEM=1`, `GIT_CONFIG_GLOBAL=<os.DevNull>`, `GIT_CONFIG_SYSTEM=<os.DevNull>`. `GIT_TERMINAL_PROMPT=0` is deliberately omitted: no requirement needs it, and local `ls-files`/`rev-parse` never prompt (SB2-1 over S4).

  It must not contain the anchors `git(root, pathspecs...)`, `sort.Strings(selected)` or the pysem import line (keeps the pin anchors unique).
- `DefaultGitRunner`:
  - Set `cmd.Env = gitRunnerEnv(os.Environ())`.
  - Before `Run`, return `cmd.Err` if non-nil.
  - Then refuse `!filepath.IsAbs(cmd.Path)` with `fmt.Errorf("retiredarch: refusing non-absolute git path %q", cmd.Path)`.

  Note (G-6): with an explicit Env, `PWD` is not updated by os/exec. git does not depend on it.
- Test helper, shared with U6 and added to `select_test.go`: a `TestMain`.
  - **Fake mode (S2-1).** Enters only when `RETIREDARCH_FAKE_GIT=1` AND `os.Args[1]` is `ls-files` or `-C`, i.e. it was invoked git-style. It writes the file named by `RETIREDARCH_FAKE_GIT_MARKER`, prints `RETIREDARCH_FAKE_GIT_OUT` to stdout, and exits 0 before flag parsing.
  - **Misconfiguration.** If the variable is set but the args are test-style (`-test.*` or none), it exits 2 with a clear message. A stray exported variable can therefore never turn the package suite into a silent no-op.
  - Otherwise it runs `m.Run()`.
  - `copyFakeGit(t, dir)` copies `os.Executable()` to `dir/git` (`git.exe` on Windows) with mode `0o755` (G2-1). Precondition: `exec.LookPath("git")` must return a relative path with nil error under the test's PATH/GODEBUG.
  - This is a natively launchable binary on both OSes (AS-2/G-1). Neither a `.bat` nor a script is used.
  - The helper lives in U4 rather than a separate task because it has no behavior to verify on its own. Its only acceptance evidence is U4 scenario 3's positive parent-red (CR2-5).
- Scenarios (3):
  1. **`TestGitRunnerEnv` (table, pure function).** Covers:
     - mixed-case `git_dir`/`Git_Index_File` dropped
     - `GIT_CEILING_DIRECTORIES` kept
     - `=C:=C:\x` kept
     - `GITX=1` kept (not a `GIT_` prefix)
     - non-GIT vars kept in order
     - the three appended entries last

     Compile-red at the parent. Write it after scenarios 2-3 per G-5.
  2. **`TestDefaultGitRunnerIgnoresGitEnv` (real temp repo).** Tracks `cmd/x/main.go` and `internal/é.go`. The clean baseline selection via `selectRepoPaths` is exactly `["cmd/x/main.go"]`. This baseline **records the known quoted-path skip of 4537B2F6**; when 4537B2F6 lands, it must update this assertion and replace vectors (c)–(e).

     Vector table, each set with `t.Setenv`. The expected selection is unchanged for every vector.
     - (a) `GIT_INDEX_FILE=<nonexistent>`: parent → empty selection
     - (b) `GIT_LITERAL_PATHSPECS=1`: parent → empty
     - (c) `GIT_CONFIG_COUNT=1` + `GIT_CONFIG_KEY_0=core.quotePath` + `GIT_CONFIG_VALUE_0=false`: parent → adds `internal/é.go`
     - (d) `GIT_CONFIG_PARAMETERS='core.quotepath'='false'`: parent → adds it
     - (e) `GIT_CONFIG_GLOBAL=<file with quotePath=false>`: parent → adds it

     (c)–(e) are one table (CR-6). Each vector is set only AFTER the fixture repo is fully built and committed, so fixture git commands never see it (G2-7).
  3. **`TestDefaultGitRunnerRefusesRelativeGit`.**
     - Setup: `dir := t.TempDir()`, `copyFakeGit(t, dir)`, `t.Chdir(dir)`, `t.Setenv("PATH", ".")` on BOTH OSes (G-3), `t.Setenv("GODEBUG", "execerrdot=0")`, and set the fake's env vars, with the output set to an absolute path.
     - Call `DefaultGitRunner(dir, "cmd/")` with `root=dir` (G-4/S2).
     - Expected: an error containing `non-absolute`, and the marker file absent (the fake never ran).
     - Parent: returns the fake's output with nil error and writes the marker → red.
     - Characterization row (SB2-2): in an empty `t.Chdir` dir with `PATH=""`, `errors.Is(err, exec.ErrNotFound)` holds and the error does NOT contain `non-absolute`. This pins the `cmd.Err`-first ordering. It holds at the parent too, so it is characterization, not red.

### U5 — Refreeze the select.go pin (D7BF9F74)

Files: `pin.go`, `pin_test.go`. Size S | Complexity medium. Blocked by U4. Same commit as U4 (ALP-2).

- `canonicalDecls`: replace the import and `DefaultGitRunner` texts with the U4 texts verbatim, and add the `gitRunnerEnv` text.
- `closedWorldDecls`: add `"gitRunnerEnv": token.FUNC`. Add `"gitRunnerEnv"` to `pathspecFrozenDecls` only. `prefixFrozenDecls` is unchanged; it already contains `"import"`, so the refrozen import covers it.
- `confinedIdentDecls`: unchanged.
- Refresh the now-stale pin.go comments at 32-33 and 224-226 ("DefaultGitRunner inherits the process environment"). Comments are outside the frozen tokens (G2-6).
- Staged order, so the new freeze has an observable red (CR2-2):
  1. Refreeze the import and `DefaultGitRunner` texts, and add `gitRunnerEnv` to `closedWorldDecls` only. The live tree is now `OK()`.
  2. Write reject row (i) and run it. It is red: `PathspecOK` stays true because `gitRunnerEnv` is not yet frozen. Log it.
  3. Add `gitRunnerEnv` to `canonicalDecls` and `pathspecFrozenDecls`. Row (i) is now green.
- Scenarios (2; SB-5):
  1. The pin on the live tree is `OK()` (all four fields true), and the full quality gates are green, including initstate and all existing pin_test rows.
  2. Reject rows via `writeMutatedCopy` (pin_test.go:53). Each copies the whole live package with one literal mutated (L-4). Each asserts `SelectFound && GuardFound && PrefixOK && !PathspecOK`, so a parse-error all-false result cannot pass (G2-2; matches pin_test.go:268-271). Each mutation must keep the file syntactically valid.
     - (i) replace the `"GIT_CONFIG_NOSYSTEM=1"` literal in `gitRunnerEnv` with `"GIT_CONFIG_NOSYSTEM=0"`
     - (ii) replace the `cmd.Env = gitRunnerEnv(os.Environ())` statement in `DefaultGitRunner` with `_ = gitRunnerEnv(os.Environ())`. This is characterization: it is already rejected via the frozen `DefaultGitRunner` text.

### U6 — Isolation and absolute-path guard in gitShowToplevel (D7BF9F74)

Files: `pin.go`, `pin_test.go`. Size S | Complexity low. Blocked by U5. U6 is a separate task because U5+U6 together would be 5 scenarios, no atomicity with ALP-2 is required, and U6 is independently revertible (SB-6).

- `gitShowToplevel`:
  - Set `cmd.Env = gitRunnerEnv(os.Environ())` (same package; `gitShowToplevel` is not a frozen select.go decl).
  - Return `cmd.Err` if non-nil.
  - Refuse `!filepath.IsAbs(cmd.Path)` with an error containing `non-absolute`.
- Scenarios (2):
  1. **Decoy env.** A real temp repo R and a decoy repo D. With `GIT_DIR=D/.git` and `GIT_WORK_TREE=D`, `gitShowToplevel(R/sub)` must return R's root. The parent returns D → red. The expected value on EVERY OS is `gitShowToplevel(R/sub)` evaluated first WITHOUT the decoy env. This avoids 8.3 short-path, symlinked-tmp and case mismatches between `t.TempDir()` and git's output (G2-3). Assert that the with-env result equals it and differs from the same derivation for D.
  2. **Relative git.** Same setup as U4 scenario 3. `t.Chdir(dir)` is required because `gitShowToplevel` sets no `cmd.Dir`. Expected: a `non-absolute` error and no marker. The parent returns the fake's absolute path without error → red.

## Dependency Graph

```
U1 ──► U2
U3            (independent)
U4 ──► U5 ──► U6      (U4+U5 = one commit, ALP-2)
```

Serial order: U1, U2, U3, U4+U5, U6.

## Rollback (AS-1)

- U6 depends on `gitRunnerEnv` (introduced by ALP-2). An ALP-2-only revert is **prohibited while U6 is applied**.
- Rollback order: revert U6, then revert ALP-2 (U4+U5 together), or revert the whole merge commit.
- U1/U2 and U3 revert independently. Reverting U1 while U2 is applied is prohibited. Revert U2 first.

## Decisions

- D1 Fallback replaces, rather than merges with, the partial cursor findings (INV-4).
- D2 The completeness oracle uses a multiset comparison and keeps cursor ordering for parity.
- D3 Lstat errors fall through to the existing read path, so only type violations get the new text.
- D4 The env allow-list is `GIT_CEILING_DIRECTORIES` only. It is a stricter prefix-based superset of the `tools/gatecheck/internal/unignore/testutil_test.go` precedent (S4). Unlike that precedent, it does not append `GIT_TERMINAL_PROMPT=0` (SB2-1).
- D5 The in-checkout git location guard is dropped (D-RA-7). Absolute-path refusal only.
- D6 The fake git is the copied test binary plus `TestMain`. Stdlib only, launchable on both OSes.

## Risks and Accepted Residuals

| ID | Residual | Rationale |
|---|---|---|
| R-A2a | A `git` binary placed earlier on PATH resolves to an absolute path and is accepted. This covers a `git` inside the checkout, and a PATH entry appended through `$GITHUB_PATH` by PR-controlled code that ran earlier in the same job (S3). | Dropped `gitPathInside` per D-RA-7. Code that can plant a binary on PATH already executes in the job, so it is in the same trust class as R-A2b. The guard added review cost (three P1s in the old plan) without a trust-boundary gain. |
| R-A2b | PR code executing earlier in the same job can alter the environment or filesystem arbitrarily. | Same-job trust class. Env isolation defends against ambient and inherited configuration, not an in-job adversary. |
| R-A2c | Repository-local config (`.git/config`, `.git/info/*`) is not isolated. | Per-key `-c key=value` overrides are possible but deliberately not used. They would change the frozen argv (pin refreeze cost), and the `-z` fix (4537B2F6) removes the `quotePath` sensitivity. Local config is produced by actions/checkout and otherwise falls in the same-job trust class as R-A2b (S2-2). |
| R-A2f | Pointing global/system config at the null device also drops any `safe.directory` entries. A container or foreign-owner checkout then fails with "dubious ownership" (G2-7). | Fails closed and visibly (git error → `::error::git ls-files`). The standard CI checkout is owned by the runner user. |
| R-A2d | Quoted paths are silently unscanned (`ls-files` without `-z`). | Tracked as stash **4537B2F6** (HIGH). U4 scenario 2 records the known skip. |
| R-A2e | git < 2.32 ignores `GIT_CONFIG_GLOBAL`/`GIT_CONFIG_SYSTEM`; `GIT_CONFIG_NOSYSTEM` still applies (S1/CR-8). | Required floor: git ≥ 2.32. Local is 2.55. ubuntu-latest ships ≥ 2.4x. Ship records `git --version` in the run log. |
| R-T1 | TOCTOU: a path could be swapped for a link between Lstat and read. | Requires a concurrent in-job writer (R-A2b class). |
| R-L1 | OneDrive and other reparse placeholders report `ModeIrregular` locally and fail containment (G-7). | Local-only. CI checkouts are plain files. The failure is closed and visible. |
| R-C1 | Other engines remain uncontained. | Stash **D44D8BDF** (LOW). |

## Constitution Check

Mapped against `.github/instructions/constitution.instructions.md` (CR2-1).

| Principle | Assessment |
|---|---|
| I Safety-First Go | Fail-closed errors throughout. No panics, no `unsafe`. Stdlib only (INV-1). |
| II Test-First (NON-NEGOTIABLE) | Every unit's reds precede implementation, including the U5 staged-freeze red. Parent SHA and platform are logged (PA-4). |
| III Workspace Isolation | U3 confines reads to contained regular files. U4/U6 isolate git from ambient config. |
| IV CLI Workspace Containment | Tests write only under `t.TempDir()`. No agent writes outside the working directory. |
| V Structured Observability | Red evidence logging (PA-4), and `git --version` plus stdlib citations in the run log. |
| VI Single Responsibility | Each unit touches one production file plus its test. `walkDecodedTOML`, `containedRegularFile` and `gitRunnerEnv` are single-purpose. |
| VII Destructive Command Approval | No destructive commands. `git init/add/commit`, `cmd /c mklink /J` and the fake git run in temp dirs only. |
| VIII Explicit Safety Modes | Freeze-scope discipline on the pin trust root via PA-1..PA-7. |
| IX Git-Friendly Persistence | N/A. No persisted state is introduced. |
| X Agent Context Efficiency | N/A to code. The plan is scoped to three stash entries. |
| XI Merge Commit History | ALP-2 is one commit. Rollback is dependency-ordered. Ship merges with a merge commit. |
| Quality Gates | `gofmt -l .`, `go vet ./...`, `go build ./...`, `go test ./...` (CR2-4). `-race` is N/A (no concurrency introduced). |

## Plan Hardening Signals

- Requires plan hardening: yes. It touches a frozen trust-root (select.go pin), subprocess environment, and filesystem link semantics.

<!-- plan-hardening-start -->
## Plan Hardening

- **PA-1 (ALP-2 atomicity).** U4 and U5 land in one commit. Ship verifies `git show --stat` lists exactly select.go, select_test.go, pin.go and pin_test.go. A commit with only one side is a P-005 stop.
- **PA-2 (pin anchor uniqueness).** Before committing ALP-2, confirm that EVERY anchor used by `writeMutatedCopy` across pin_test.go (pin_test.go:57-60, 228-262, 323-333) still occurs exactly once in select.go (G2-5). This includes:
  - `git(root, pathspecs...)`
  - `sort.Strings(selected)`
  - the pysem import line
  - `package retiredarch\n`
  - the `args := append(...)` line
  - `\treturn selected, nil\n}\n`

  The existing pin tests failing on a non-unique anchor is the check.
- **PA-3 (init-time tests untouched).** initstate_test.go and the pin_test.go env-mutator reject rows are not edited. `gitRunnerEnv` reads `os.Environ()` at call time, never in `init` or package-level var initialisers.
- **PA-4 (red evidence).** Per scenario, log the command, platform, parent SHA and failing assertion text. HEAD-only evidence is acceptable only for Unix-symlink reds that skip on Windows. Those cannot skip on Linux (`t.Fatal` on creation failure), so CI green proves they ran (L2-2).
- **PA-5 (fake git safety).** The `TestMain` fake branch triggers only on `RETIREDARCH_FAKE_GIT=1` AND git-style argv. With test-style argv the variable causes exit 2, never a silent pass (S2-1). It never touches the real repository. The marker and output paths live under `t.TempDir()`. Production code cannot reach it (`_test.go`).
- **PA-6 (rollback).** Follow the dependency-ordered rollback section. Never revert ALP-2 alone while U6 is applied.
- **PA-7 (inventory HALT).** The U4 precheck inventory must match the Stage list or HALT before code edits.
<!-- plan-hardening-end -->

## Runtime Verification

- `go test ./tools/gatecheck/...` green on Windows (Ship) and Linux (CI).
- `gatecheck retired-arch --self-test` and `--self-test-integrity` exit 0 on the real checkout.
- The repo scan on the real checkout is unchanged: Code 0, no new findings. There are no tracked symlinks (verified: no mode 120000 entries).

## Harvest Mapping (CR-7)

- Covering feature: "Fix retiredarch TOML desync, symlink containment and git env isolation" (chore class: gate correctness hardening; priority high). It references this plan and the D-RA-7 decision.
- Tasks: one per unit, U1..U6. Each task's acceptance criteria are its unit's scenarios plus "package-wide `go test ./tools/gatecheck/...` green". For U4, the package-wide criterion moves to U5.
- Dependency edges: U2 blocked by U1; U5 blocked by U4; U6 blocked by U5.
- Stash consumption: archive only 9FC28DB9, 990AFA71 and D7BF9F74 at Step 5.6. DC921AF6, 3750C37C, 0ECC1895, 4537B2F6 and D44D8BDF stay active.
- Each task's acceptance criteria also include the full quality gates at its completion point (CR2-4). U4 is excepted; for U4 they move to U5. The deliberation's earlier Step 1(B) "archived at Step 5.6" line for all six is superseded by D-RA-7.
- **Harvest amendment (PR #109 review, before the staging merge):** U4 and U5 are harvested as ONE task, 050.004-T. They had been split into 050.004-T and 050.005-T, but Ship runs full gates, commits and marks each task done before advancing, so a U4-only task could neither pass nor be committed under ALP-2. The task numbered 050.005-T was removed. 050.006-T (U6) is now blocked by 050.004-T. The merged task carries a documented size exception (L, ~2.5 h) because ALP-2 cannot be divided. The unit numbering U1..U6 in this plan is unchanged.

## Plan Review — attempt 1 (revision 1)

- dispatch_mode: multi-agent
- decision: FAIL

| Persona | Route | Verdict | Highest |
|---|---|---|---|
| Architecture Strategist | anchor (gpt-6.1-sol) | FAIL | P1×2 |
| Security Lens Reviewer | default | PASS | P3 |
| Go Reviewer | default | FAIL | P1×1 |
| Scope Boundary Auditor | default | ADVISORY | P2 |
| Constitution Reviewer | default | FAIL | P1×1 |
| Learnings Researcher | default | ADVISORY (confidence medium) | P2 |

P1 findings and their rev-2 dispositions:

- **AS-1** The rollback for ALP-2-only is broken once U6 depends on `gitRunnerEnv`. → Fixed: Rollback section and PA-6.
- **AS-2 / G-1 / CR-2** The U6 relative-git test was green at the parent: the fake printed nothing, so the existing "returned nothing" error fired, and `.bat` is not natively launchable. → Fixed: copied-test-binary fake prints an absolute path, the test asserts `non-absolute` plus marker absence, and the parent red is positive.
- **CR-1** The follow-up stash entries did not exist. → Fixed: 4537B2F6 and D44D8BDF were created before harvest and are cited in Out of Scope and the residuals.

P2/P3 dispositions:

- **Fixed:**
  - CR-3: fallback-default red, fallback-error red, U3 table; the subset check was removed and Lstat errors fall through.
  - SB-1, SB-2 (minimal component rejection with a table row), SB-3, SB-4, SB-5, SB-6, SB-7, SB-8.
  - G-2, G-3, G-4, G-5, G-6.
  - CR-4, CR-5, CR-6, CR-7, CR-8, CR-9.
  - S1, S2, S3, S4.
  - L-1, L-2, L-4.
- **Accepted as residual:** G-7 (R-L1).
- **Fixed by the absence of shell scripts:** L-3 (marker file adopted). L-5 (GOROOT evidence) applies only if the Ship log cites toolchain behavior; adopted in PA-4 as "log failing assertion text".

<!-- plan-review-attempt: 1 -->

## Plan Review — attempt 2 (revision 2)

- dispatch_mode: multi-agent
- decision: ADVISORY (no P1; all attempt-1 P1s — AS-1, AS-2/G-1/CR-2 and CR-1 — verified resolved)
- Disposition: ADVISORY findings were folded into revision 3 in place, with no new review cycle (Step 4: ADVISORY may proceed). Operator confirmation was given by delegation: the Orchestrator runs on operator autopilot under the "go with the recommendation" disposition (D-RA-7).

| Persona | Route | Verdict | Highest |
|---|---|---|---|
| Architecture Strategist | anchor (gpt-6.1-sol) | ADVISORY | P2 (AS2-1) |
| Security Lens Reviewer | default | PASS | P3 |
| Go Reviewer | default | ADVISORY | P2 (G2-1..G2-3) |
| Scope Boundary Auditor | default | ADVISORY | P3 |
| Constitution Reviewer | default | ADVISORY | P2 (CR2-1, CR2-2) |
| Learnings Researcher | default | ADVISORY (confidence medium-high) | P3 |

All findings were fixed in rev 3. See the revision-history row for each fix:

- **Architecture:** AS2-1 (P2).
- **Go:** G2-1 (P2), G2-2 (P2), G2-3 (P2), G2-4 (P3), G2-5 (P3), G2-6 (P3), G2-7 (P3).
- **Constitution:** CR2-1 (P2), CR2-2 (P2), CR2-3 (P3), CR2-4 (P3), CR2-5 (P3).
- **Security:** S2-1 (P3), S2-2 (P3), S2-3 (P3).
- **Scope:** SB2-1 (P3, `GIT_TERMINAL_PROMPT` removed), SB2-2 (P3, characterization row), SB2-3 (P3, note).
- **Learnings:** L2-1 (P3), L2-2 (P3), L2-3 (P3).

<!-- plan-review-attempt: 2 -->
---
title: "Gatecheck Batch A correctness: implementation plan"
date: 2026-10-08
source: "docs/decisions/2026-10-08-intercom-go-gatecheck-batch-a-correctness-deliberation.md"
stash_ids: ["4537B2F6", "0D643BE8", "D44D8BDF"]
status: "harvested"
base: "main@135ca59"
---

# Gatecheck Batch A correctness: implementation plan

**Source:** `docs/decisions/2026-10-08-intercom-go-gatecheck-batch-a-correctness-deliberation.md`
(decisions D-BA-1 to D-BA-5).

**Stash IDs:** 4537B2F6, 0D643BE8, D44D8BDF.

**Harvest:** feature 051-F; tasks U1=051.001-T, U2=051.002-T, U3=051.003-T, U4=051.004-T,
U5=051.005-T, U6=051.006-T; shipment 041-S (queued). Deferred stash entries: B83F53BB (writepath
runner `-z`/env isolation, high) and 05E12A6F (writepath fixture self-test containment, low).

**Learnings applied:**

* `docs/compound/2026-10-07-pin-canonical-text-maintenance-and-unbounded-ast-rule-review.md`
* `docs/compound/workflow-issues/staging-premise-verification-before-planning-2026-09-18.md`

## Problem Frame

### 1. The repo-mode retired-architecture gate can pass while silently dropping tracked paths (4537B2F6)

* `retiredarch.DefaultGitRunner` (`tools/gatecheck/internal/retiredarch/select.go:39-60`)
  runs `git ls-files -- <pathspecs>` without `-z`.
* `selectRepoPaths` (`select.go:196-217`) decodes the result with `pysem.GitText` and
  splits it with `pysem.SplitLines`.
* **Quoting.** Git C-quotes any path with bytes >= 0x80, `"`, `\` or control characters,
  e.g. `"internal/\303\251.go"`. The quoted form begins with `"`, so `shouldScanRepoPath`
  drops it.
* **Splitting.** `SplitLines` also splits on U+0085, U+2028, U+2029 and U+001C..U+001E.
* **Result.** A tracked Go file whose name needs quoting is never scanned, and the gate
  passes. `select_test.go:404-455` (`TestDefaultGitRunnerIgnoresGitEnv`) pins this
  behaviour on main today: its baseline excludes `internal/\u00e9.go`.

### 2. An empty repo-mode selection exits 0 (0D643BE8)

`runRepoScan` (`retiredarch.go:592-615`) loops over zero paths and returns `Result{Code: 0}`.
The sibling writepath gate already fails closed on the same condition
(`writepath.go:791-796`, ED-7).

### 3. Two other engines read through tracked symlinks (D44D8BDF, narrowed)

* **writepath.** `runRepoScan` (`writepath.go:798-805`) calls `scanFile` →
  `pysem.ReadText(filepath.Join(root, rel))` (`writepath.go:459`). It has no `Lstat`.
* **unignore.** `rootGitignoreTextAt` (`unignore/git.go:91-92`) reads
  `os.ReadFile(repoDir/.gitignore)` at `HEAD`, which follows a symlinked `.gitignore`.
* **Already satisfied or refuted:**
  * retiredarch is satisfied by PR #112 (`containedRegularFile`, `retiredarch.go:716`);
  * mergestrategy is refuted (`mergestrategy/evaluate.go:9-17`, a documented exception).

## Requirements Trace

| Req | Source | Requirement | Units |
|---|---|---|---|
| R1 | 4537B2F6, D-BA-1 | Every tracked in-scope path is selected byte-exact, whatever characters it contains (non-ASCII, quote, backslash, Unicode line boundaries) | U1, U2 |
| R2 | 4537B2F6, D-BA-1 | A listing that is not NUL-delimited, has empty records, or is invalid UTF-8 fails closed, never false-clean | U2 |
| R3 | 4537B2F6, constitution ALP | `select.go` frozen declarations and `pin.go` `canonicalDecls` change in one atomic commit; the pin stays green, and a new mutation row rejects dropping `-z` | U2 |
| R4 | 4537B2F6, D-BA-1 | The independent self-test oracle `expectedInternalRepoPaths` keeps working under `-z` with its own parse | U2 |
| R5 | 4537B2F6, D7BF9F74 | `TestDefaultGitRunnerIgnoresGitEnv` keeps proving env isolation with vectors that distinguish under `-z` (baseline flip in U2; vector replacement in U6) | U2, U6 |
| R5a | plan-review SEC-2, D-BA-1a | With `-z`, raw path bytes can reach CI logs. A selected path containing a control character (< 0x20, 0x7f) or U+0085, U+2028 or U+2029 fails the selection closed with a `%q`-escaped message, so it can never inject workflow commands | U2 |
| R6 | 0D643BE8, D-BA-2 | Repo mode (`""`) with an empty selection exits 1 with a non-empty `::error::` line | U3 |
| R7 | D44D8BDF, D-BA-3 | The writepath repo scan never reads a selected path through a symlink, junction or non-regular component; it fails closed instead | U4 |
| R8 | D44D8BDF, D-BA-3 | The unignore `HEAD` baseline never reads `.gitignore` through a link or a non-regular file; it fails closed instead, and an absent file stays `""` | U5 |
| R9 | constitution II | Every behaviour change has a test that is red at the parent and green after the change | U2-U5 |

## Constitution Check

The constitution is at `.github/instructions/constitution.instructions.md`.

| Principle | Status | Note |
|---|---|---|
| I. Safety-first Go | compliant | Bounds-checked parse, no panics, error identity kept (INV-D), wrapped errors (U5) |
| II. Test-first (NON-NEGOTIABLE) | compliant, with recorded exceptions below | Every behaviour unit has an observed red. Characterization rows are labelled. U4(a)/U5(a) reds must be OBSERVED before implementation (see U4/U5) |
| III. Workspace isolation | compliant | U4/U5 make the gates stricter; INV-B holds; TOCTOU accepted with the same rationale as retiredarch U3 |
| IV. CLI containment | compliant | All fixtures live inside `t.TempDir()`, including symlink targets |
| V. Observability | compliant | Every new failure is an `::error::` line or a `%q`-escaped error |
| VI. Single responsibility | compliant | stdlib only (`unicode/utf8`); the writepath containment copy is justified in D-BA-3 |
| VII. Destructive approval | N/A | No destructive operation; rollback is `git revert` |
| VIII. Safety modes | compliant | U2 runs under **careful** mode with a **freeze-scope** of `tools/gatecheck/internal/retiredarch/` (PA-1) |
| IX. Git-friendly persistence | compliant | Markdown artifacts |

**Recorded exceptions (Governance conflict-resolution rule):**

* **EX-1, U2 exceeds the 2-hour file and scenario bounds.**
  * Principle bent: the 2-hour rule (fewer than 3 files and fewer than 4 test scenarios).
  * Why: ALP. The frozen `select.go` declarations, the `pin.go` `canonicalDecls` and the
    pin-test anchor must land in one commit, along with the shared-runner second consumer
    (`expectedInternalRepoPaths`) and the in-package fake format.
  * Simpler alternative rejected: splitting U2 leaves either a pin-red commit or a
    `--self-test-integrity`-red commit on the branch.
  * Mitigation: test scenarios are consolidated into 3 table-driven tests plus one
    real-git test plus the baseline flip. The env-vector replacement is split out to U6.
* **EX-2, harness-ready labelling for characterization work.** U1 is characterization and
  test-only, so under P-004 it has no failing test. Ship labels it harness-ready on the
  basis of "green before and after, same test count". It does not fabricate a red. The
  characterization rows in U2, U4, U5 and U6 are exempt from the red requirement and are
  labelled as such in the evidence.

## Scope Boundaries

**Out of scope.** Each item is captured or accounted for:

* **writepath `DefaultGitRunner` gaps:** no `-z`, and no D7BF9F74 environment isolation.
  Captured as a new DEFERRED SCOPE EXPANSION stash entry.
* **writepath fixture self-test** (`runFixtureSelfTest`, `writepath.go:833-880`), which
  reads `os.ReadDir`-enumerated fixtures through `scanFile` with no `Lstat`. It is not a
  git-selected path, so it falls outside D44D8BDF's premise. Captured as a new low-priority
  DEFERRED SCOPE EXPANSION stash entry.
* **mergestrategy containment:** refuted (documented exception).
* **GitRunner consolidation** (5A8EC1BC) and a **shared containment package:** rejected
  in D-BA-3.
* **Pin trust-root entries** DC921AF6, 3750C37C and 0ECC1895 (D-RA-7, operator-excluded).

**Invariants that must not regress:**

* **INV-A, writepath mask flow.** `scanFile` stays exactly ReadText → `scanSource`, with a
  single mask and a single `pysem.SplitLines` (pinned by `retiredarch/writepath_mask_test.go`).
  U4 must not edit `scanFile` or `scanSource`.
* **INV-B.** No `filepath.Abs` or `filepath.EvalSymlinks` reachable from registered runners
  (`retiredarch/initstate_test.go`).
* **INV-C, pin closed world.** `closedWorldDecls`, `confinedIdentDecls`,
  `pathspecFrozenDecls` and `prefixFrozenDecls` membership is unchanged: no new top-level
  declaration goes into `select.go`.
* **INV-D, error identity.** Invalid UTF-8 in the listing still surfaces as
  `pysem.ErrInvalidUTF8`.
* **INV-E, `--self-test` and `--self-test-integrity` output.** Golden outputs
  (`TestRun_SelfTest*_MatchesGolden`) are unchanged.

## Implementation Units

Every unit is Go code in a single gatecheck package, together with that package's tests
(TDD, constitution II). Width is isolated per package.

### U1: Characterization test-prep, a shared listing helper (test-only)

* **Goal.** Centralise the fake `git ls-files` output format so that U2 flips it in one place.
* **Files:**
  * `tools/gatecheck/internal/retiredarch/select_test.go`
  * `tools/gatecheck/internal/retiredarch/retiredarch_test.go`
* **Change:**
  * Add a test-only helper `lsFilesListing(paths ...string) []byte` to `select_test.go`.
    For now it returns `strings.Join(paths, "\n") + "\n"`, which is today's runner format.
  * `TestSelectRepoPaths_FiltersAndSorts` (`select_test.go:120-140`) returns
    `lsFilesListing("internal/z/a.go", "internal/z/a_test.go", "cmd/x/y_test.go", "config.toml.example", "internal/a/b.go")`.
  * `u3StubGit` (`retiredarch_test.go:584-588`) returns `lsFilesListing(paths...)`.
* **Posture.** Characterization-first. **Green on arrival BY DESIGN**: there is no
  behaviour change, so U1 has no red.
* **Tests.** The full `go test ./tools/gatecheck/...` stays green, with the same test count.
* **AC:**
  1. Both fakes use the helper, and no other inline newline listing remains:
     `rg -n 'Join\(paths, "\\n"\)|\\ninternal/z/a_test.go' tools/gatecheck/internal/retiredarch/*_test.go`
     finds only the helper.
  2. The tests stay green.
  3. The diff touches only the two test files.
* **Size / complexity:** XS / trivial.

### U2: Atomic NUL-safe selection (`-z`), with the pin refreeze (ONE commit)

* **Goal.** R1-R5, R5a.
* **Files** (5; an atomic exception to the "fewer than 3 files" rule):
  * `retiredarch/select.go`
  * `retiredarch/pin.go`
  * `retiredarch/pin_test.go`
  * `retiredarch/selftest_selection.go`
  * `retiredarch/select_test.go`
* **Why 5 files.** ALP: frozen `select.go` declarations, `pin.go` `canonicalDecls`, the
  `pin_test.go` mutation anchor, and the shared runner's second consumer must all change in
  the same commit. Any split leaves either a pin-red commit or an oracle-red
  `--self-test-integrity` commit. Precedent: 050.004-T / compound doc §L. Ship commits per
  task, so this must stay one task.
* **Change (`select.go`, frozen surface):**
  * the import block adds `"unicode/utf8"` (stdlib group, sorted after `"strings"`);
  * the `DefaultGitRunner` argv becomes `append([]string{"ls-files", "-z", "--"}, pathspecs...)`;
  * `selectRepoPaths` replaces the `GitText`/`SplitLines` decode with the inline NUL parse in
    §CANON below;
  * update the `GitRunner` and `DefaultGitRunner` doc comments to say `-z` and
    NUL-terminated (comments are not pinned).
* **Change (`pin.go`):**
  * `canonicalDecls` gets the same three texts, authored **independently** (H-11). Do not
    copy-paste `select.go`; retype from §CANON;
  * every frozen-set map and slice stays unchanged (INV-C).
* **Change (`pin_test.go`):** update one anchor and add reject rows, as specified under
  "`pin_test.go` rows" below.
* **Change (`selftest_selection.go`, `expectedInternalRepoPaths`, `:79-102`):**
  * replace `GitText`/`SplitLines` with an independently written NUL parse:
    * `utf8.Valid`, else return `pysem.ErrInvalidUTF8`;
    * empty output → empty set;
    * `bytes.HasSuffix(out, []byte{0})`, else error;
    * `bytes.Split(out[:len(out)-1], []byte{0})`, rejecting empty records;
  * use a different code shape from `selectRepoPaths` (bytes vs strings); it must not call
    `selectRepoPaths` or a shared helper;
  * keep the filter rules verbatim.
* **Change (`select_test.go`):**
  * flip `lsFilesListing` to `strings.Join(paths, "\x00") + "\x00"`, or return `nil` for no
    paths;
  * add the tests below;
  * flip only the `TestDefaultGitRunnerIgnoresGitEnv` baseline. Vector replacement is U6.
* **Posture.** Test-first, with a staged freeze (compound doc: to get a real red, freeze
  only after the reds are observed):
  1. Write the new tests and the helper flip only, and run them at the parent's `select.go`
     to observe the reds below.
  2. Apply the `select.go` + `selftest_selection.go` change.
  3. Apply the `pin.go` refreeze and the anchor update.
  4. Run all green.
  5. Commit once.

  Record the red evidence in the task's implementation-notes and the PR body.

**Tests.** These are consolidated into table-driven tests to respect the scenario budget
(EX-1). The column "Red kind" separates *bug-evidence* reds, *new-contract* reds and
*characterization*:

* **Bug-evidence red:** the parent code mishandles real git output.
* **New-contract red:** red only because the runner output format changed. This is
  legitimate, but it is NOT evidence of the 4537B2F6 bug.
* **Characterization:** green on arrival.

| # | Test | Red kind | Parent behaviour | After |
|---|---|---|---|---|
| d | `TestRunRepoScan_NonASCIIPath_IsScanned_RealGit` | **bug-evidence (required)** | Uses a real `git init` temp repo, with `fixtureGit` as in `TestDefaultGitRunnerIgnoresGitEnv`. It tracks a clean `cmd/x/main.go` (so the selection is non-empty under either U2/U3 order) and `internal/\u00e9.go`, whose body holds a retired-architecture token copied from an existing reject fixture. `runRepoScan(root, DefaultGitRunner)` returns **Code 0** at the parent (false-clean) | Code 1; stderr names `internal/\u00e9.go` (compare via `filepath.ToSlash`) |
| base | `TestDefaultGitRunnerIgnoresGitEnv` baseline only | **bug-evidence (required)** | The baseline `want` becomes `{cmd/x/main.go, internal/\u00e9.go}`. The parent selects only `cmd/x/main.go`, so this is red. Delete the 4537B2F6 TODO comment. Vectors (a)-(e) stay as they are in U2; under `-z` the quotePath vectors (c)-(e) become green but non-distinguishing, and U6 replaces them | Baseline green |
| a | `TestNULListing_PreservesSpecialNames`: a table over **both** `selectRepoPaths` and `expectedInternalRepoPaths`, with `lsFilesListing("internal/\u00e9.go", "internal/q\"uote.go", "internal/back\\slash.go", "cmd/x/main.go")` | new-contract | The parent `SplitLines` reads the NUL-joined listing as one garbage record | All in-scope paths are selected byte-exact and sorted, by both functions |
| b | `TestSelectRepoPaths_BadListing_FailsClosed`: a table with these rows: newline listing `"internal/a.go\n"`; empty record `"internal/a.go\x00\x00"`; lone `"\x00"`; control char `lsFilesListing("internal/a\nb.go")`; separator `lsFilesListing("internal/a\u2028b.go")`; ESC `lsFilesListing("internal/a\x1bb.go")` | new-contract / bug-evidence (newline row) | Parent: the newline row returns `{internal/a.go}` with a nil error (red). The other rows return nil errors or garbage (red) | Each row returns a non-nil error. The control and separator rows' error text contains the `%q` form and **no raw control byte** |
| e | `TestSelectRepoPaths_InvalidUTF8_ErrInvalidUTF8`: the fake returns `[]byte{0xff, 0}` | characterization | green (via `GitText`) | `errors.Is(err, pysem.ErrInvalidUTF8)` (INV-D). Ship may drop it if an existing test already asserts this |
| g | `pin_test` reject rows (below) | characterization of the new pin | Not runnable at the parent: the anchor is absent there | Pin fails on each mutated copy; the original passes |

**`pin_test.go` rows.** Write every anchor as a **raw (backtick) Go string**, so that `\x00`
matches the source text byte for byte and not a real NUL byte.

* Update `default_git_runner_drops_pathspecs` (`:244-246`) to the anchor
  `` `args := append([]string{"ls-files", "-z", "--"}, pathspecs...)` ``.
* **Add** `default_git_runner_drops_nul_flag`, which mutates that anchor to the parent argv.
  It guards against reverting `-z`.
* **Add** `select_repo_paths_newline_split`, which mutates
  `` `strings.Split(string(out[:len(out)-1]), "\x00")` `` to the `"\n"` form. It guards
  the newly pinned parse, because `-z` alone does not guarantee a NUL split. D-BA-1a
  records this.
* **Optional:** `select_repo_paths_empty_record_skipped`, which mutates the empty-record
  error return to a `continue`.

Before relying on any anchor, Ship confirms with `rg` that it is unique in the new
`select.go`.

**§CANON: exact frozen texts.** These are the same tokens in `select.go` and in
`pin.go` `canonicalDecls`; comments are excluded.

```go
import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem"
)

func DefaultGitRunner(root string, pathspecs ...string) ([]byte, error) {
	args := append([]string{"ls-files", "-z", "--"}, pathspecs...)
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	cmd.Env = gitRunnerEnv(os.Environ())
	if cmd.Err != nil {
		return nil, cmd.Err
	}
	if !filepath.IsAbs(cmd.Path) {
		return nil, fmt.Errorf("retiredarch: refusing non-absolute git path %q", cmd.Path)
	}
	var stdout, stderrBuf bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderrBuf
	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(stderrBuf.String()); msg != "" {
			return nil, fmt.Errorf("%w: %s", err, msg)
		}
		return nil, err
	}
	return stdout.Bytes(), nil
}

func selectRepoPaths(root string, git GitRunner) ([]string, error) {
	pathspecs := make([]string, 0, len(scanScope))
	for _, a := range scanScope {
		pathspecs = append(pathspecs, a.pathspec)
	}
	out, err := git(root, pathspecs...)
	if err != nil {
		return nil, err
	}
	if !utf8.Valid(out) {
		return nil, pysem.ErrInvalidUTF8
	}
	if len(out) == 0 {
		return nil, nil
	}
	if out[len(out)-1] != 0 {
		return nil, fmt.Errorf("output is not NUL-terminated")
	}
	var selected []string
	for _, path := range strings.Split(string(out[:len(out)-1]), "\x00") {
		if path == "" {
			return nil, fmt.Errorf("output has an empty record")
		}
		for _, r := range path {
			if r < 0x20 || r == 0x7f || r == 0x85 || r == 0x2028 || r == 0x2029 {
				return nil, fmt.Errorf("path %q contains a control or line-separator character", path)
			}
		}
		if shouldScanRepoPath(path) {
			selected = append(selected, path)
		}
	}
	sort.Strings(selected)
	return selected, nil
}
```

Notes for the §CANON texts:

* **Unchanged declarations.** `GitRunner`, `gitRunnerEnv`, `scanArm`, `scanScope` and
  `shouldScanRepoPath` are unchanged.
* **Universe coverage.** The texts introduce no new universe identifier, so
  `universeDeclNames` coverage is unchanged. `TestUniverseDeclNames_CoverCanonicalTexts`
  must stay green.
* **Error-message prefix.** The messages carry no `retiredarch:` or `git ls-files` prefix,
  because `runRepoScan` already prints `::error::git ls-files: %v`.
* **Imports.** `pysem` remains imported, which keeps the `aliased_import` row's anchor valid.
* **Independent oracle.** `expectedInternalRepoPaths` does NOT need to replicate the
  control-character rejection. If such a path exists, `selectRepoPaths` errors and the
  self-test fails closed anyway.
* **Pin violations.** If Ship finds a pin rule that §CANON violates, it **stops and reports
  back**, and does not weaken the pin.

**PR review aid (SEC-4, H-11).** The PR body pastes the `pin.go` `canonicalDecls` diff next
to this §CANON block. Reviewers compare the pin against **§CANON**, not against `select.go`.

**AC:**

1. Bug-evidence reds (d) and the baseline are observed red at the parent, and the evidence
   is recorded.
2. New-contract reds (a) and (b) are observed red at the parent with the flipped helper.
3. Everything is green after the change.
4. `go test ./tools/gatecheck/...` is green.
5. `go run ./tools/gatecheck retired-arch --self-test-integrity` exits 0, with the golden
   unchanged (INV-E).
6. `go run ./tools/gatecheck retired-arch` exits 0 on the live tree.
7. `git show --stat HEAD` lists exactly the 5 files, in one commit.
8. Each `pin_test` reject row fails the pin on its mutated copy.
9. `rg -n 'ls-files", "--"' tools/` finds no other text check that still expects the old
   argv; if it does, the plan is stopped and reported.

**Size / complexity:** M / medium.

### U3: Fail-closed empty-selection guard in `runRepoScan`

* **Goal.** R6.
* **Files:**
  * `retiredarch/retiredarch.go` (`runRepoScan` only; not frozen)
  * `retiredarch/retiredarch_test.go`
* **Change.** After the `selectRepoPaths` error check, insert:

  ```go
  if len(relPaths) == 0 {
      return Result{Stderr: "::error::retired-arch repo scan selected no files under config.toml.example, cmd/** or internal/** (fail-closed)\n", Code: 1}
  }
  ```

  * Ship may refine the wording. It must start with `::error::` and end with a newline.
  * Update the `runRepoScan` doc comment.
* **Posture.** Test-first.
* **Red tests (Code 0 at the parent):**
  * (a) `TestRunRepoScan_EmptySelection_FailsClosed`: the fake returns
    `lsFilesListing()` (empty) → expect `Code 1` and an `::error::` stderr.
  * (b) `TestRunRepoScan_OnlyOutOfScopePaths_FailsClosed`: the fake returns
    `lsFilesListing("internal/a_test.go", "internal/x/testdata/y.go", "docs/readme.md")`,
    where every path is filtered out → `Code 1`.
  * `Run`'s exit-code pass-through is already covered by `TestRun_RepoScan_GitError_FailsClosed`
    (`retiredarch_test.go:171`), so no third test is needed (plan-review SCOPE-6).
* **Characterization (must stay green).** `TestRun_Repo_MatchesGolden` on the live tree, and
  the self-test goldens. The self-test path never reaches `runRepoScan` with an empty
  selection, because its own "selection non-empty" assertion fires first. Confirm this by
  running.
* **AC:**
  1. (a) and (b) are red at the parent and green after the change.
  2. The repo-mode and self-test goldens are unchanged.
  3. `go test ./tools/gatecheck/...` is green.
  4. `select.go` and `pin.go` are untouched.
* **Format independence.** U3 depends on U1 (the helper), not on U2. The fakes go through
  `lsFilesListing`, so U3's tests hold in either format, whichever of U2 or U3 lands first.
* **Size / complexity:** S / low.

### U4: writepath repo-scan containment (D44D8BDF, writepath arm)

* **Goal.** R7.
* **Files:**
  * `tools/gatecheck/internal/writepath/writepath.go`
  * `tools/gatecheck/internal/writepath/writepath_test.go`
* **Change:**
  * Add an unexported `containedRegularFile(root, rel string) (bool, string)` to
    `writepath.go`. It is a semantic copy of `retiredarch.containedRegularFile`
    (`retiredarch.go:700-750`):
    * a component-wise `os.Lstat` walk;
    * reject empty, `.` or `..` components, and on Windows reject `\` and `:`;
    * intermediates must be real directories (no `ModeSymlink` or `ModeIrregular`);
    * the final component must be regular;
    * `fs.ErrNotExist` → ok (so the existing read-error text is preserved);
    * any other Lstat error → fail.

    The doc comment cites retiredarch U3 and records why the code is copied, not shared
    (D-BA-3 / 5A8EC1BC).
  * In `runRepoScan`'s loop (`writepath.go:799-805`), **before** `scanFile`:

    ```go
    if ok, reason := containedRegularFile(root, rel); !ok {
        return Result{Stderr: errorLine(rel, fmt.Errorf("not a contained regular file: %s", reason)), Code: 1}
    }
    ```

    This is abort-on-first semantics, consistent with writepath's existing ED-2 read-error
    abort.
  * Rename the loop-local `fs, err := scanFile(...)` to `fileFindings, err := ...`. The
    loop-local `fs` would shadow the `io/fs` import that the copied helper needs. The
    rename is in `runRepoScan`, not in `scanFile`, so INV-A still holds. Add the `errors`,
    `io/fs` and `runtime` imports as needed.
  * The copy strips `*fs.PathError` exactly as retiredarch does, so CI logs show only
    repo-relative prefixes (SEC-4).
  * **Coverage note (SEC-1).** U4 protects only paths that writepath's own runner actually
    selects. writepath's runner still lacks `-z`, so a quoted path is dropped before
    containment ever sees it. That gap is the separately captured DEFERRED SCOPE EXPANSION
    entry, which has high priority.
  * Do **NOT** modify `scanFile` or `scanSource` (INV-A).
  * Use no `filepath.Abs` or `EvalSymlinks` (INV-B).
* **Posture.** Test-first. Mirror the retiredarch U3 test helpers locally in `writepath_test.go`:
  a stub GitRunner returning a newline listing (writepath's runner is NOT `-z`), and a
  Windows junction helper.
* **Red tests:**
  * (a) `TestRunRepoScan_FinalComponentSymlink_FailsClosed`: `internal/x.go` is an
    `os.Symlink` to a clean Go file elsewhere in the same `t.TempDir()`, outside `root`.
    Skip **only** when `runtime.GOOS == "windows"` and the error is a privilege error. On
    any other OS a symlink failure is `t.Fatalf` (SEC-3). Red: parent `Code 0`.
  * (b) `TestRunRepoScan_IntermediateJunctionOrSymlink_FailsClosed`: `internal/d` is a
    junction (Windows, `mklink /J`) or a directory symlink (Unix) to a dir holding a clean
    `y.go`, and the stub lists `internal/d/y.go`. Red: parent `Code 0`.
  * (c) `TestRunRepoScan_DeletedTrackedFile_KeepsReadError` (characterization): the stub
    lists a non-existent `internal/gone.go` → `Code 1`, with the parent's exact error text.
    Green at the parent.
* **AC:**
  1. (b) is OBSERVED red at the parent before any `writepath.go` change: junction on
     Windows, directory symlink elsewhere. (a) is observed red wherever symlinks can be
     created. Both are green after the change.
  2. (c) is unchanged.
  3. The PR body records that `go test -v` on the Linux CI job reports (a) as PASS, not
     SKIP.
  4. `retiredarch` `TestWritePath*` mask-flow tests are green (INV-A).
  5. `retiredarch/initstate_test.go` is green (INV-B).
  6. `go run ./tools/gatecheck write-path` exits 0 on the live tree.
  7. `go test ./tools/gatecheck/...` is green.
* **Size / complexity:** S / medium.

### U5: unignore `HEAD` `.gitignore` regular-file check (D44D8BDF, unignore arm)

* **Goal.** R8.
* **Files:**
  * `tools/gatecheck/internal/unignore/git.go`
  * `tools/gatecheck/internal/unignore/git_test.go`
* **Change.** In `rootGitignoreTextAt`'s `ref == "HEAD"` branch (`git.go:91-104`), first run
  `os.Lstat(filepath.Join(repoDir, ".gitignore"))`:
  * `errors.Is(err, fs.ErrNotExist)` → `("", nil)` (unchanged behaviour);
  * any other error → `fmt.Errorf("unignore: lstat root .gitignore: %w", err)`;
  * `!info.Mode().IsRegular()` → `fmt.Errorf("unignore: root .gitignore is not a regular file (mode %v)", mode)`.
    Ship matches the prefix convention the callers use when printing (check how
    `resolveBaselineTree` and `classifyShowFailure` errors surface).
  * otherwise `os.ReadFile`, as today.

  Update the function's doc comment. The callers (`checks.go:43`, `:124`, `:128`) already
  propagate errors fail-closed.
* **Note.** A TOCTOU gap between `Lstat` and `ReadFile` is accepted, with the same rationale
  as retiredarch U3 (local, non-adversarial CI checkout).
* **Posture.** Test-first.
* **Red tests:**
  * (a) `TestRootGitignoreTextAt_HEAD_SymlinkGitignore_Errors` is **the only red**.
    * Setup: `.gitignore` in a temp repo is an `os.Symlink` to a file in the same
      `t.TempDir()` containing `secret/\n`.
    * Skip rule: skip only on Windows with a privilege error; anywhere else, `t.Fatalf`.
    * Red: the parent returns the target text with a nil error.
  * (b) `TestRootGitignoreTextAt_HEAD_DirectoryGitignore_Errors` is **characterization**.
    `os.ReadFile` on a directory already errors at the parent on every OS. It pins the new
    error path's message.
  * (c) The existing absent-file test stays green (`("", nil)`).
* **AC:**
  1. (a) is OBSERVED red at the parent BEFORE any `git.go` change, in a
     symlink-capable environment. Options:
     * Linux/WSL;
     * Windows with Developer Mode or elevated rights;
     * a CI run on a pushed test-only commit of the Ship branch.

     If none of these is available, Ship halts under P-005 and does not implement without
     a red (constitution II; plan-review CONST-F2). (a) is green after the change, and
     the PR body shows it as PASS, not SKIP, on Linux CI.
  2. The existing `git_test.go` and `checks_test.go` tests are green.
  3. `go run ./tools/gatecheck unignore` exits 0 on the live tree.
  4. `go test ./tools/gatecheck/...` is green.
* **Size / complexity:** XS / low.

### U6: Replace the env-isolation vectors that `-z` neutralised (test-only)

* **Goal.** R5 (D7BF9F74 stays live after 4537B2F6).
* **Files:** `tools/gatecheck/internal/retiredarch/select_test.go` (only
  `TestDefaultGitRunnerIgnoresGitEnv`).
* **Change.** Replace the quotePath vectors (c)-(e). Each new vector makes an
  **unisolated** git fail:
  * (c) `GIT_CONFIG_COUNT` = `"1"` with no `GIT_CONFIG_KEY_0` (git 2.31 or later);
  * (d) `GIT_CONFIG_PARAMETERS` = Go string `"not-a-valid-config"`;
  * (e) `GIT_CONFIG_GLOBAL` = a temp file containing `"[core\n"` (git 2.32 or later).
* **Liveness control, per vector.** A test-local runner,
  `exec.Command("git", "ls-files", "-z", "--", "cmd/**", "internal/**")`, with:
  * `cmd.Env` set to `os.Environ()` with every `GIT_*` entry removed (the same filter as
    `fixtureGit`);
  * `GIT_CONFIG_NOSYSTEM=1` and `GIT_CONFIG_GLOBAL=os.DevNull` (except where the vector
    sets it);
  * **only** the vector added on top.

  This control must exit **non-zero**. That proves the vector is live, and that it is
  live because of the vector and not because of ambient config.
* **Isolation assertion, per vector.** `DefaultGitRunner` under `t.Setenv(vector)` must
  still equal the baseline.
* **Version floor.** Skip a vector with a message when `git version` is below its floor.
  A control that does not fail on a supported git version is `t.Fatalf`, never a pass.
  Ship may substitute another isolation-only channel (`GIT_CONFIG_SYSTEM`,
  `HOME`/`XDG_CONFIG_HOME`) and records the substitution in the PR body.
* **Red kind: characterization plus liveness.** `gitRunnerEnv` isolation already exists at
  the parent, so the isolation assertions are green on arrival. The control runner is what
  shows the vectors are not vacuous. This is a labelled exception to the red requirement
  (EX-2).
* **AC:**
  1. Vectors (a) and (b) are unchanged.
  2. (c)-(e) are replaced, and each has a passing liveness control.
  3. `go test ./tools/gatecheck/...` is green.
  4. The diff touches only `select_test.go`.
* **Size / complexity:** S / low.

## Dependency Graph

```text
U1 (test helper) --blocks--> U2 (-z atomic refreeze)
U1 (test helper) --blocks--> U3 (empty-selection guard)
U2 (-z refreeze)  --blocks--> U6 (env-isolation vector replacement)
U4 (writepath containment)   independent
U5 (unignore .gitignore)     independent
```

* **Suggested execution order:** U1 → U2 → U6 → U3 → U4 → U5.
* **U2 and U3 are commutative.** Both touch `retiredarch_test.go` or `select_test.go` only
  through the helper, and they touch different production functions.
* **No cycles.**

## Decisions and Rationale

See the deliberation for the full options tables.

* **D-BA-1: `-z` with an inline parse in the frozen `selectRepoPaths`.** This is the only
  option that closes both the quoting and the Unicode-boundary hazards, and the parse stays
  inside the pinned surface.
* **D-BA-2: the guard goes in `runRepoScan`.** That function is unfrozen and already decides
  the exit code, and the guard mirrors writepath ED-7.
* **D-BA-3: per-engine containment.** Mergestrategy is excluded as refuted. A shared package
  was rejected because it pre-empts 5A8EC1BC and re-touches 040-S code.
* **D-BA-4: the U1 helper keeps U2 to a format flip,** and U2 stays one atomic task.
* **D-BA-1a (plan-review amendment).**
  * Inside the same frozen parse, any selected path containing a control character or
    U+0085, U+2028 or U+2029 fails the selection closed with a `%q` message.
  * Why: `-z` passes raw bytes through, and printing them unescaped could inject GitHub
    Actions workflow commands (SEC-2). It also closes the U+2028 split hazard by refusing
    such paths rather than scanning them.
  * The new `select_repo_paths_newline_split` pin row guards the NUL split itself.
* **D-BA-5: one feature, six tasks, no sub-epics** (050-F precedent). U6 was split out of
  U2 at plan-review (SCOPE-2) because the old vectors stay green, but no longer
  distinguish anything, under `-z`. They are not red, so they do not need to share U2's
  atomic commit.

## Risks and Caveats

| ID | Risk | Likelihood / impact | Mitigation |
|---|---|---|---|
| RK-1 | A hidden pin rule rejects §CANON (e.g. selector, confinement or universe coverage) | low / medium | U2 AC 3-4; stop-and-report rule; never weaken the pin |
| RK-2 | A test fake still emits newline listings after U2 and silently passes | low / high | NUL-terminator check makes it an error; U1 centralises the fakes; U2 greps for leftovers |
| RK-3 | Replacement env vectors don't distinguish on the CI git version | medium / medium | Per-vector liveness control with `t.Fatalf`; documented substitution path |
| RK-4 | Symlink tests skip on unprivileged Windows hosts, so the red is never observed locally | high / low | Linux CI runs them; the junction variant (U4 b) runs on Windows; record where the red was observed |
| RK-5 | The U4 check placement breaks the writepath mask-flow pin | low / high | It goes in `runRepoScan`, not `scanFile`; INV-A tests are in the AC |
| RK-6 | The exit-code change (empty selection 0 → 1) breaks a consumer | very low / low | No consumer found; wrappers pass the exit code through |
| RK-8 | Path bytes reach CI logs unescaped once `-z` lands (workflow-command injection) | low / medium | D-BA-1a: control/separator paths fail selection closed with a `%q` message; test (b) asserts there are no raw control bytes |
| RK-7 | Invalid-UTF-8 path names: git `-z` emits raw bytes, so a non-UTF-8 tracked filename fails the whole gate with `ErrInvalidUTF8` | low / low | Intended fail-closed behaviour (INV-D), matching today's `GitText` behaviour |

## Plan Hardening Signals

* **Public API, schema or contract change: PRESENT.**
  * The `GitRunner` output contract (newline → NUL) changes, along with the frozen pin
    texts.
  * The repo-mode exit code on an empty selection changes.
* **Security, auth, permission or compliance-sensitive behaviour: PRESENT.** These are
  merge-blocking integrity gates, and the fixes are fail-open → fail-closed.
* **Migration, backfill, destructive data/config action or irreversible step: ABSENT.**
  Code-only, and fully revertible.
* **External integration, operator checkpoint or external dependency: PRESENT (minor).** The
  real `git` binary's `-z` and config-channel behaviour (RK-3).
* **High runtime, rollout or rollback risk: ABSENT.** CI-only tooling, and each task can be
  reverted with a single commit.

Requires plan hardening: yes

## Runtime Verification and Closure

| Unit | Runtime surface | Verification before absorption | Closure artifact |
|---|---|---|---|
| U1 | none (tests) | `go test ./tools/gatecheck/...` | task notes |
| U2 | `gatecheck retired-arch` CLI (repo, `--self-test`, `--self-test-integrity`); CI job `check-retired-architecture` | Repo mode exits 0 on the live tree; `--self-test-integrity` golden unchanged; real-git non-ASCII red→green (d) | PR body red/green evidence plus `git show --stat` |
| U6 | none (tests) | `go test ./tools/gatecheck/...`; each vector's liveness control fails | task notes, plus any vector substitution recorded in the PR body |
| U3 | `gatecheck retired-arch` repo mode | Live-tree exit 0; empty-fake exit 1 | PR body evidence |
| U4 | `gatecheck write-path` CLI; CI write-path job | Live-tree exit 0; symlink/junction red→green | PR body evidence (OS where observed) |
| U5 | `gatecheck unignore` CLI; CI unignore job | Live-tree exit 0; symlink red→green on Linux CI | PR body evidence |

* **Rollback trigger.** Any of the three gates failing on main after the merge with a
  non-finding error. Revert the offending task's commit. For U2, revert the whole commit and
  never the pin alone.
* **Validation window.** First CI run on main after the merge.
* **Owner.** Ship (shipment), then Orchestrator closure.

## Plan Hardening

*Appended by plan-harden, Stage session 2026-10-08.*

**Hardening need confirmed:** three signals are present (contract change, security-sensitive
gate, external git dependency).

**Proposed actions and risk:**

| ProposedAction | ActionRisk | Approval needed | Notes |
|---|---|---|---|
| PA-1: edit frozen `select.go` declarations + `pin.go` `canonicalDecls` (U2) | high: the pin trust surface | No extra approval beyond plan-review and PR review; ALP atomic commit is mandatory | The pin is the integrity anchor for the scan scope. A non-atomic or copy-pasted refreeze defeats H-11 independence. Ship must retype `canonicalDecls` from §CANON and diff the two token streams through the existing pin test, not by eye |
| PA-2: change the `GitRunner` output format to NUL (U2) | medium | No | Every in-package fake must change in the same commit; RK-2 tripwire |
| PA-3: change the repo-mode exit-code semantics (U3) | low | No | Fail-closed direction only |
| PA-4: add a containment refusal to writepath (U4) and unignore (U5) | low | No | Fail-closed direction only; no read-path restructuring |

**Fail-closed invariants Ship must preserve (checked in every PR):**

* Every new branch in U2-U5 converts an unknown state into Code 1 or an error, never into
  Code 0 or a skip.
* No `continue` without a finding in U2's parse: an empty record is an error, not a skip.
* No new `t.Skip` except symlink-privilege skips, and each of those names the CI job that
  runs the case.

**Rollback:**

* U2 is a single commit; revert that commit as a whole. Reverting `pin.go` alone leaves the
  pin red on main by design.
* U1, U3, U4, U5 and U6 are independent single-commit reverts. U6 must be reverted before U2 if U2 is reverted.
* There is no data, config or schema state to restore.

**Verification commands** (Ship runs them per task, and in full before the PR):

```text
go build ./...
go vet ./tools/gatecheck/...
go test ./tools/gatecheck/...
go run ./tools/gatecheck retired-arch
go run ./tools/gatecheck retired-arch --self-test-integrity
go run ./tools/gatecheck write-path
go run ./tools/gatecheck unignore
```

* Ship confirms each subcommand's exact name from the `register_*.go` files before running it.
* After U2, also run `git show --stat HEAD` and check that exactly 5 files changed.

**Residual risks carried forward:**

* RK-3 and RK-4 are accepted, with documented fallbacks.
* **Closure doc hygiene (CONST-F8).** Two contract changes need recording: the newline→NUL
  runner contract and the change from exit 0 to exit 1 on an empty selection.
  * During compound capture, Ship refreshes
    `docs/compound/2026-10-07-pin-canonical-text-maintenance-and-unbounded-ast-rule-review.md`
    with the `-z` refreeze outcome.
  * The PR body records both contract changes.
  * No separate docs task is created, because no operator-facing doc states the old
    contract (verified by the deliberation grep).
The two DEFERRED SCOPE EXPANSION entries (the writepath runner, and writepath fixture
containment) are explicitly not mitigated by this shipment.

## Plan Review

<!-- plan-review-attempt: 1 -->

dispatch_mode: parallel-subagents
reviewers: Go Reviewer, Scope Boundary Auditor, Security Lens Reviewer, Constitution Reviewer
(the Step 1.8 learnings were consulted directly; their findings are applied in §Learnings
applied)

Per-reviewer verdicts (attempt 1): every reviewer returned **ADVISORY**, and none returned
FAIL. Every finding below is resolved in this revision of the plan:

| ID | Sev | Finding | Resolution |
|---|---|---|---|
| GO-1 | medium | The env-vector control inherits ambient git env/config | U6 control env is scrubbed, with NOSYSTEM and GLOBAL set to DevNull, plus a version floor |
| GO-2 | low | The `\x00` anchor must be a raw string | "`pin_test.go` rows": every anchor is a raw string, uniqueness checked with `rg` |
| GO-3 | low | Real-git red (d) needs a non-empty selection that is independent of U3 | (d) also tracks a clean `cmd/x/main.go`, and compares with `ToSlash` |
| GO-4 | low | The writepath loop-local `fs` shadows `io/fs` | U4 renames it to `fileFindings` (in `runRepoScan`, not `scanFile`) |
| GO-5 | low | U5 error-prefix and `IsNotExist` style | U5 uses `errors.Is`, a wrapped `lstat` error, and the callers' prefix convention |
| GO-6 | adv | Doubled `git ls-files` prefix in §CANON messages | §CANON messages shortened |
| SCOPE-1 / CONST-F4 | high | The two DEFERRED SCOPE EXPANSION entries were not yet written, and the fixture entry was missing from the deliberation's Dispositions table | Written to the stash in this Stage commit; Dispositions table updated |
| SCOPE-2 | medium | U2 test-scenario count | Tests consolidated into tables; vector replacement split out to U6; EX-1 recorded |
| SCOPE-3 / CONST-F2 | medium / P1 | U5 (b) mislabelled; U5's red could be unobservable before implementation | (b) relabelled as characterization. (a) must be OBSERVED red pre-change (local symlink-capable, or CI on a test-only commit), else halt under P-005. The same rule applies to U4 |
| SCOPE-4 | low-medium | Format-change reds are not bug evidence | Red-kind column added; (d) and the baseline are the required bug evidence |
| SCOPE-5 | low | The `newline_split` row is unjustified | Justified in D-BA-1a (guards the newly pinned parse) |
| SCOPE-6 | low | U3 (c) duplicates existing coverage | Dropped; cites `TestRun_RepoScan_GitError_FailsClosed` |
| SEC-1 | P2 | The writepath runner `-z` gap is under-prioritised | Deferred entry captured at **high** priority; U4 coverage note added |
| SEC-2 | P2 | Raw control bytes from `-z` paths can inject workflow commands into CI logs | D-BA-1a: fail selection closed on control/separator characters with `%q`; RK-8; test (b) rows |
| SEC-3 | P3 | A loose symlink skip hides the only red | Skip only on Windows with a privilege error, otherwise `t.Fatalf`; the PR shows PASS, not SKIP, on Linux |
| SEC-4 | P3 | H-11 independence is procedural only | PR review aid: compare the `canonicalDecls` diff against §CANON; optional empty-record reject row; PathError stripping in U4 |
| CONST-F1 | P1 | Constitution Check section missing | Added, with exceptions EX-1 and EX-2 |
| CONST-F3 | P2 | Harness-ready labelling for characterization work | EX-2 |
| CONST-F5 | P3 | No named safety mode for PA-1 | U2 runs in careful mode, with freeze-scope `retiredarch/` |
| CONST-F6 | P3 | U5 error handling style | Same as GO-5 |
| CONST-F7 | P3 | Vector labelling; symlink target inside `TempDir` | U6 labelled characterization plus liveness; U4/U5 targets are inside `t.TempDir()` |
| CONST-F8 | P3 | Contract changes undocumented | Closure doc hygiene under Plan Hardening |

**Gate decision.** All reviewer verdicts were ADVISORY, and every advisory finding is
resolved in the plan text. No finding remains open that would warrant FAIL. Under the
Stage rule, an ADVISORY result proceeds when the operator confirms. Confirmation here is
relayed by the Orchestrator's `stage next on Batch A` instruction, and the operator keeps
the ability to override at staging-PR review.

decision: PASS

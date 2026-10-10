---
title: "Gatecheck Batch B1 write-path input hardening: implementation plan"
date: 2026-10-10
source: "docs/decisions/2026-10-10-intercom-go-gatecheck-batch-b1-write-path-input-hardening-deliberation.md"
stash_ids: ["B83F53BB", "05E12A6F"]
status: "reviewed"
base: "origin/main@1b130a2"
---

# Gatecheck Batch B1 write-path input hardening: implementation plan

**Source:** `docs/decisions/2026-10-10-intercom-go-gatecheck-batch-b1-write-path-input-hardening-deliberation.md`
(decisions D-BW-1, D-BW-1a, D-BW-2, D-BW-3, D-BW-4, D-BW-5).

**Stash IDs:** B83F53BB, 05E12A6F.

**Harvest:** feature 052-F; tasks U1=052.001-T, U2=052.002-T, U3=052.003-T, U4=052.004-T,
U5=052.005-T (no sub-epics); shipment 042-S (queued; manifest 052-F first, then the five tasks in
dependency order). Dependency edges: 052.002-T blocked by 052.001-T; 052.003-T and 052.005-T blocked
by 052.002-T; 052.004-T independent. Consumed stash entries B83F53BB and 05E12A6F (archived). New
residual stash entries (active): 4372BAD4, DBE25DF5, EDC18D59. Evidence note appended to 5A8EC1BC.
`size` is a structured field on every task; `complexity` is recorded as labelled prose because the
workspace task type defines no complexity field (degradation flagged; see the harvest notes).

**Learnings applied:**

* `docs/compound/2026-10-07-pin-canonical-text-maintenance-and-unbounded-ast-rule-review.md`
* `docs/compound/workflow-issues/staging-premise-verification-before-planning-2026-09-18.md`
* `docs/compound/workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md`
* `docs/compound/2026-09-08-adversarial-review-empirical-verification-and-scope-check.md`
* `docs/compound/2026-09-08-verify-go-stdlib-internals-claims-against-goroot-source.md`
* `docs/compound/2026-10-02-git-show-ref-failure-ambiguity-resolve-before-show.md`
* `docs/compound/2026-09-04-backlogit-sizing-is-wit-gated.md`
* `docs/compound/2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md`

## Problem Frame

### 1. The write-path gate can pass while silently not scanning tracked files (B83F53BB, part 1)

* `DefaultGitRunner` (`tools/gatecheck/internal/writepath/writepath.go:147-160`) runs
  `git ls-files -- internal/** cmd/**` with no `-z`.
* `runRepoScan` (`writepath.go:830-848`) decodes the result with `pysem.GitText` (`:835`) and splits
  it with `pysem.SplitLines`, then filters with `shouldScan` on the raw string.
* **Quoting.** Git C-quotes a path with bytes >= 0x80, `"`, `\` or a control character, for
  example `"internal/\303\251.go"` (confirmed with git 2.55 during staging). The quoted form begins
  with `"`, so `shouldScan` drops it. The gate passes without scanning the file.
* **Splitting.** `GitText` rewrites `\r`; `SplitLines` splits on U+0085, U+2028, U+2029 and
  U+001C..U+001E.
* **Exposure today: none.** 56 tracked files under `internal/**` and `cmd/**`, none needs quoting.

### 2. The same runner inherits the process's git environment (B83F53BB, part 2)

* `DefaultGitRunner` sets only `cmd.Dir`; there is no `cmd.Env`.
* Staging probe (git 2.55): `GIT_INDEX_FILE` pointing at an alternate index that lists only
  `cmd/x/main.go` makes the listing omit tracked `internal/bad.go`. The listing is non-empty,
  so ED-7 (`writepath.go:850-855`) does not fire: **false-clean**. `GIT_LITERAL_PATHSPECS=1`, a
  missing `GIT_INDEX_FILE`, `GIT_DIR`, `GIT_CONFIG_PARAMETERS` and `GIT_CONFIG_GLOBAL` (malformed)
  fail closed but make the gate flaky. The agent host itself exports `GIT_CONFIG_COUNT=3` and
  `KEY_0..2`.
* **Boundary of this plan.** It isolates the `git ls-files` child for a **given root**. The CI
  wrapper's own root discovery (`scripts/check-write-path-precondition.sh:150`,
  `ROOT="$(git rev-parse --show-toplevel)"`, unscrubbed) is a separate surface, captured as a new
  stash entry (see Scope Boundaries).

### 3. The fixture self-test reads through links (05E12A6F)

* `runFixtureSelfTest` (`writepath.go:895-959`) uses `os.Stat(fixtureDir)` (`:899`, follows links),
  `os.ReadDir` (`:904`), skips only `e.IsDir()` (`:910`), then calls
  `scanFile(root, "scripts/testdata/writepath/"+name)` (`:927`) with no `Lstat`.
* A tracked symlink or junction named `*.go`, or a linked `scripts/`, `scripts/testdata/` or
  fixture directory, is followed. There are 28 tracked direct `.go` fixtures and no tracked
  symlink today.
* `containedRegularFile` (`writepath.go:786-823`, from 051.004-T) already exists in-package and
  walks every component with `os.Lstat`.

## Requirements Trace

| Req | Source | Requirement | Units |
|---|---|---|---|
| R1 | B83F53BB, D-BW-1 | Every in-scope tracked `.go` path is selected byte-exact, whatever characters it contains (non-ASCII, quote, backslash, space) | U1, U2 |
| R2 | B83F53BB, D-BW-1 | A listing that is not NUL-terminated or has an empty record fails closed, and a scanned record that is not valid UTF-8 fails closed with `pysem.ErrInvalidUTF8`; none can be false-clean | U2 |
| R3 | D-BW-1a | A path that will be scanned and contains a rune `< 0x20`, `0x7f`, `0x85`, `0x2028` or `0x2029` fails the selection closed with a `%q` message and no raw control byte in stderr; an unscanned path with such a rune is tolerated | U2 |
| R4 | B83F53BB, D-BW-2 | The `git ls-files` listing that `DefaultGitRunner` returns for a given root does not depend on ambient `GIT_*` variables or global/system git config. (The wrapper's root *discovery* is out of scope; captured.) | U3 |
| R5 | D-BW-2 | Each env vector is observed red at the parent, and the same test's no-vector baseline succeeds; assertions compare the listing bytes, never the exit code alone | U3 |
| R6 | 05E12A6F, D-BW-3 | The fixture self-test refuses a linked fixture (a symlink, or a junction on Windows), and a fixture reached through a linked ancestor (`scripts`, `scripts/testdata`, or the fixture directory). A non-link non-regular fixture cannot come from git, so it has no row | U4 |
| R7 | D-BW-4 | The frozen oracle is adapted with exactly one authorized edit (listing decode) | U1 |
| R8 | D-BW-4 | The stale `retiredarch/select.go` comment about writepath's listing is corrected | U5 |
| R9 | constitution II | Every behaviour change has a test that is red at its parent and green after; characterization is labelled | U1-U5 |

## Constitution Check

The constitution is at `.github/instructions/constitution.instructions.md`.

| Principle | Status | Note |
|---|---|---|
| I. Safety-first Go | compliant | Strict bounds-checked parse, explicit wrapped errors, `errors.Is`-friendly `pysem.ErrInvalidUTF8`; no panics |
| II. Test-first (NON-NEGOTIABLE) | compliant, with recorded characterization | U2-U4 have observed, behavioural reds (compiling stubs, never a compile-red). U1 and U5 are characterization or comment-only and are labelled as such |
| III. Workspace isolation | compliant | U4 makes the self-test stricter; U3 narrows what git reads |
| IV. CLI containment | compliant, by interpretation | This constrains the agent's CLI operations, not Go test code. Test fixtures (real git repos, links, the alternate index) live in `t.TempDir()` under the OS temp directory, the repository's convention. Decision: do **not** redirect `TMPDIR`/`TEMP` into the checkout (git discovery would find the outer repository); evidence captures that Ship wants on disk go under the git-ignored `logs/`. Record the interpretation in the Ship PR body, as 041-S did |
| V. Observability | compliant | Every new refusal is an `::error::` line or a `%q`-escaped error |
| VI. Single responsibility | compliant | stdlib only; the `gitRunnerEnv` copy is justified in D-BW-2 / 5A8EC1BC |
| VII. Destructive approval | N/A | No destructive operation; rollback is `git revert` |
| VIII. Safety modes | compliant | **Careful** mode for U1 (frozen-oracle edit), U2 (runner contract flip) and U3 (subprocess environment). Pause points: before U1 touches `writepath_oracle_test.go`, Ship confirms the staging PR is approved (the PA-2 authorization) and halts if it cannot; before U2's real `writepath.go` change, Ship confirms the reds were observed. All units run with a **freeze-scope** of `tools/gatecheck/internal/writepath/`; the single comment-only edit to `retiredarch/select.go` (U5) is a declared freeze-scope deviation |
| IX. Git-friendly persistence | compliant | Markdown artifacts |
| X. Context efficiency | compliant | Targeted reads and queries |
| XI. Merge commit history | compliant | The staging PR and the Ship PR use merge commits; Ship verifies the merge strategy (P-009) before merging |
| Capability overlays | noted | backlogit applies (harvest, traceability, shipment). agent-engram and graphtor-docs are not used for implementation. agent-intercom is not installed, so there are no broadcasts |

**Counting convention (2-hour rule).** A *scenario* is a distinct test function, or a table row that
exercises a distinct behaviour; sub-cases that differ only by an input string or a path count once.
A *function* is any function added or whose body changes (test helpers included; one-line swaps
included). The exceptions below state the counts under this convention.

**Recorded exceptions (Governance conflict-resolution rule; each in the four-part form: principle
bent, why, simpler alternative rejected, mitigation):**

* **EX-1, U1 touches more than 4 functions.**
  * *Principle bent:* the 2-hour rule (fewer than 5 functions): 6 (two new helpers, four existing).
  * *Why:* U1 adds two helpers, swaps a literal for the helper in two test functions
    (`u4StubGit`, and the inline fakes in `TestRun_ReadError_FailsClosed` and
    `TestRun_InvalidUTF8_FailsClosed`, which also gain a path assertion), and replaces the decode
    block in `TestOracle_TrackedProductionTree_Parity`. Centralising the fakes is the point of the
    task.
  * *Alternative rejected:* folding the helpers into U2. That makes U2 larger and removes the
    characterization baseline that proves the U2 flip changes only the format.
  * *Mitigation:* every edit is mechanical and the task adds no behaviour.
* **EX-2, harness-ready labelling for characterization work.**
  * *Principle bent:* P-004 / constitution II ("harness-ready means a red").
  * *Why:* U1 and U5 have no failing test by design.
  * *Alternative rejected:* inventing a red.
  * *Mitigation:* Ship labels them harness-ready on "green before and after, same test count", and
    labels the tightened assertions and regression guards as characterization in the evidence.
  * *Precedent (so Ship's harness step does not halt on a green-on-arrival U1 or U5):* Batch A
    recorded this same exception as its own EX-2
    (`docs/plans/2026-10-08-intercom-go-gatecheck-batch-a-correctness-plan.md`; the Batch A
    deliberation states the same green-on-arrival labelling rule), and its characterization tasks
    051.001-T and 051.006-T (shipment 041-S) were labelled `harness-ready` and shipped as
    green-on-arrival characterization tasks, with no red claimed. A green-on-arrival harness result
    is the **expected** outcome for U1 and U5, never a halt.
* **EX-3, U2 exceeds the 2-hour bounds.**
  * *Principle bent:* the 2-hour rule: about 10 functions (production: `DefaultGitRunner`,
    `runRepoScan`, the new helper; tests: two helper flips, `fixtureGit`, `hermeticGitEnv`, three
    tests) against fewer than 5; about 3 scenarios by the counting convention (a real-git proof,
    helper-level rows, runner-level rows) with roughly 20 rows; test infrastructure with production
    code (Width Isolation).
  * *Why:* the runner-format flip, the strict parse, the selected-path rules, and the real-git proof
    (`fixtureGit`, which travels with its first consumer) are one atomic behaviour. A split leaves a
    commit with stale fakes or a known raw-byte echo.
  * *Alternative rejected:* splitting the control-character rule into its own task. The interim
    commit would echo raw tracked-path bytes (the workflow-command injection that D-BW-1a closes).
  * *Mitigation:* three table-driven tests, two files, one commit, and a compiling stub so every red
    is behavioural.
* **EX-4, U3 exceeds the 2-hour bounds.**
  * *Principle bent:* the 2-hour rule: about 7 functions (`gitRunnerEnv`, `DefaultGitRunner`,
    `fixtureGitEnv`, the `fixtureGit` wrapper, the baseline-capture helper, two tests) against fewer
    than 5; 2 scenarios with about 8 rows; test infrastructure with production code.
  * *Why:* the env filter and its real-git proof are one behaviour. The vectors are cut to three
    (one bug-evidence, two config/pathspec channels) from five.
  * *Alternative rejected:* splitting the pure `TestGitRunnerEnv` table from the real-git vectors.
    That leaves a production function whose use by the runner is unproven, or the reverse.
  * *Mitigation:* table-driven; mirrors the shipped retiredarch tests.
* **EX-5, U4 exceeds the 2-hour scenario bound.**
  * *Principle bent:* the 2-hour rule: about 5 table sub-cases (a leaf file symlink, a leaf
    directory link, and three linked ancestors) against fewer than 4 scenarios; test infrastructure
    (`linkDir`) with production code.
  * *Why:* the sub-cases share one fixture builder and one assertion and differ only by which path
    is linked. The production change is one call in one loop.
  * *Alternative rejected:* splitting by link kind. That duplicates the builder across tasks.
  * *Mitigation:* one table-driven test; sub-cases differ only by the linked path.

## Scope Boundaries

**Out of scope.** Each item is captured or accounted for:

* **Non-absolute `git` path refusal** (retiredarch's `cmd.Err` + `filepath.IsAbs` guard). Not in
  B83F53BB's stated scope; the guarded actor is the in-job attacker (residual R-A2b). **Captured**
  as a new DEFERRED SCOPE EXPANSION stash entry `4372BAD4` (low).
* **The wrapper's root discovery and incomplete-listing detection.** `ROOT="$(git rev-parse
  --show-toplevel)"` runs unscrubbed in `scripts/check-write-path-precondition.sh`, and ED-7
  detects only an *empty* selection, not a partial one (for example a per-arm check that the
  selection holds both an `internal/` and a `cmd/` file). **Captured** as a new DEFERRED SCOPE
  EXPANSION stash entry `DBE25DF5` (medium, kind bug).
* **Escaping of other CI-log sinks.** Fixture names (`PASS %s`, `FAIL %s`), `errorLine`'s `%v`
  tail (parser errors that quote multi-line literals, raw git stderr) and the absolute path inside
  the unwrapped `ReadText` error are pre-existing echo sinks outside the listing arm. **Captured**
  as a new DEFERRED SCOPE EXPANSION stash entry `EDC18D59` (low). D-BW-1a closes only the *new*
  raw-byte path that `-z` opens in the listing arm.
* **retiredarch changes** other than the one comment in U5; `31F33EFE` (control-character ordering
  twin), `C8827920` (retiredarch env-test table), `4995F8C3` (unit-tag labels) stay active.
* **GitRunner consolidation** (5A8EC1BC) and a shared git package: rejected in D-BW-2. The
  operator allowed an evidence note on the entry ("note any new evidence only"); one is appended.
* **unignore git isolation** (documented parity decision), `B29A565E`, `F5958BBC`.
* **The other 33 write-path entries** (Root.Resolve precision, harness hygiene, write-primitive
  coverage), and every retiredarch/unignore/CI/docs entry.

**Accepted residuals (recorded, not mitigated):**

* TOCTOU between the `Lstat` walk and the read (local, non-adversarial CI checkout; the same
  rationale as retiredarch U3 and 051.004-T).
* A `git` shim on `PATH`, an in-job attacker, and a self-modifying PR (R-A2b trust class).
* Git older than 2.32 does not honour `GIT_CONFIG_GLOBAL`/`GIT_CONFIG_SYSTEM`; it still reads
  `$HOME` config. No global key is known to change default `ls-files` membership.

**Invariants that must not regress:**

* **INV-A, mask flow.** `scanFile` stays exactly ReadText then `scanSource`, with a single mask and
  a single `pysem.SplitLines` inside `scanSource` (pinned by
  `retiredarch/writepath_mask_test.go`). No unit edits `scanFile` or `scanSource`.
* **INV-B.** No `filepath.Abs` or `filepath.EvalSymlinks` reachable from registered runners
  (`retiredarch/initstate_test.go`).
* **INV-C, no local masker.** No writepath non-test file declares a masker or references
  `scanSource` outside the authorized path.
* **INV-D, error identity.** A scanned record that is not valid UTF-8 still surfaces as
  `pysem.ErrInvalidUTF8` (`errors.Is`).
* **INV-E, goldens.** `TestRun_SelfTest*_MatchesGolden` and `TestRun_RepoMode_CurrentTreeIsClean`
  output is unchanged.
* **INV-F, finding order.** Selected paths keep git's order (no re-sort), so finding order and the
  goldens are unchanged.
* **INV-G, oracle.** The legacy side and every expectation in `writepath_oracle_test.go` are
  unchanged; only the single authorized decode edit lands (U1).

## Contract-surface matrix (build before the first edit; re-run at U2)

| Surface | Evidence | Check (run at the U2 base and after U2) |
|---|---|---|
| Producer: `DefaultGitRunner` | `writepath.go:147-160` | read the argv: no `-z` at the U2 base, `-z` after U2 |
| Runner doc contract | `GitRunner` and `DefaultGitRunner` doc comments (`writepath.go:138-146`, "for pysem.GitText decoding") | updated by U2 (U2 AC5) |
| Consumer: `runRepoScan` | `writepath.go:830-848` | read the function; no `GitText`/`SplitLines` left in it after U2 |
| Consumer: oracle decode | `writepath_oracle_test.go:194-203` | after U1 it calls only `decodeLsFilesListing`; the oracle file is unchanged in U2 |
| Consumer: wiring | `register_write_path.go:25` | signature unchanged (`GitRunner`) |
| Fakes | `u4StubGit`, inline fakes in `TestRun_ReadError_FailsClosed`, `TestRun_InvalidUTF8_FailsClosed` | every **well-formed-listing** fake returns through `lsFilesListing`; the only raw literals are the malformed-listing rows of test (b), including the named `legacyNewlineListing` row, added by U2 |
| Wrappers and wiring tests | `scripts/check-write-path-precondition.sh`; `tools/gatecheck/check_write_path_wrapper_test.go`; `tools/gatecheck/main_test.go` write-path tests | no new failures versus the parent baseline (see the environment prechecks) |
| Goldens | `TestRun_SelfTest*_MatchesGolden`, `TestRun_RepoMode_CurrentTreeIsClean` | unchanged output |
| Pins | `retiredarch/writepath_mask_test.go`, `initstate_test.go` | green; `scanFile`/`scanSource` diff empty |
| Cross-package comment | `retiredarch/select.go:25-31` | fixed by U5 |
| Operator-facing docs | **Named-surface read, not a grep count.** Read the wrapper usage block in `scripts/check-write-path-precondition.sh`, the top-level `docs/*.md`, and the writepath doc comments. The wrapper header does not state the listing format (staging read) | re-read at U2. A supporting `rg -n 'newline-separated|GitText|one path per line' scripts docs tools --glob '!docs/archive/**' --glob '!docs/plans/**' --glob '!docs/decisions/**' --glob '!docs/memory/**' --glob '!docs/closure/**'` is expected to hit only `retiredarch/select.go` (until U5) and the writepath doc comments U2 rewrites; it is a locator, **not closure evidence** |

## Implementation Units

Every unit is Go code in the writepath package (plus one comment in retiredarch for U5), with its
tests (TDD, constitution II). Width is isolated per package. Use only standard-library APIs that
exist at the `go.mod` floor (Go 1.24); `go vet`'s `stdversion` check flags newer ones. Every new
test helper calls `t.Helper()`; no new test uses `t.Parallel`. Write `é` as `\u00e9` in test
source and keep CLI text ASCII.

### U1: Characterization test-prep, shared listing helpers and the oracle decode adaptation (test-only)

* **Goal.** R7, and set up R1/R2 so U2 flips only the format.
* **Files:**
  * `tools/gatecheck/internal/writepath/writepath_test.go`
  * `tools/gatecheck/internal/writepath/writepath_oracle_test.go`
* **Change:**
  * Add `lsFilesListing(paths ...string) []byte` to `writepath_test.go`. For now it returns `nil`
    for no paths and `[]byte(strings.Join(paths, "\n") + "\n")` otherwise, which is the runner's
    **current** format.
  * Add `decodeLsFilesListing(t *testing.T, out []byte) []string`. For now it applies
    `pysem.GitText` and `pysem.SplitLines`, drops empty strings, and calls `t.Fatalf` on a decode
    error. Its doc comment states: an independent decode of real runner output; it must not call
    production selection code; it frames records only (it applies no UTF-8 policy beyond what its
    format needs), because the oracle applies `shouldScan` afterwards; **changing it changes what
    the frozen oracle examines, so it is treated as an oracle edit (D-BW-4, PA-2)**.
  * `u4StubGit` returns `lsFilesListing(paths...)`.
  * `TestRun_ReadError_FailsClosed` and `TestRun_InvalidUTF8_FailsClosed` return
    `lsFilesListing(...)` instead of an inline literal, and **also assert that stderr names the
    intended path** (`internal/does-not-exist-anywhere.go`, `internal/bad.go`). Both messages
    already contain it, so this is green before and after. It exists so that, after the U2 flip, a
    stale newline fake cannot satisfy these tests with the wrong error.
    `TestRun_EmptySelection_FailsClosed_ED7` already returns empty bytes (format-agnostic) and is
    **not** edited.
  * `TestOracle_TrackedProductionTree_Parity` (`writepath_oracle_test.go:188-212`): replace the
    decode block `:194-203` (the `GitText` call, its error check, `var rels []string` and the
    `SplitLines` loop) with:

    ```go
    var rels []string
    for _, p := range decodeLsFilesListing(t, out) {
    	if shouldScan(p) {
    		rels = append(rels, p)
    	}
    }
    ```

    Keep the `len(rels) == 0` fatal and the `t.Run` loop.
  * **Reconciled header wording (written out here so the file does not contradict itself, and true
    whether or not U2 has landed).** Replace the LIFECYCLE paragraph
    (`writepath_oracle_test.go:9-12`) with:

    ```text
    // LIFECYCLE: the legacy side and its expectations are frozen. Two edits are
    // authorised. First, the single AC-E2.6 adaptation in Unit E (production
    // side, input list, and moving unparseable-input expectations to stricter
    // fail-closed assertions). Second, the B83F53BB listing-format adaptation
    // (D-BW-4, staging PR for Gatecheck Batch B1): the tracked-tree test decodes
    // the runner's listing through decodeLsFilesListing, the runner-format
    // adapter in writepath_test.go, instead of pysem.GitText and
    // pysem.SplitLines. Any other edit is an H-3 stop.
    ```
* **Posture.** Characterization-first. **Green on arrival BY DESIGN**: there is no behaviour change,
  so U1 has no red (EX-2). Careful mode: Ship confirms the PA-2 authorization (the staging PR is
  approved) before touching the oracle file, records PA-2 as `approved` with the staging-PR
  reference, and halts if it cannot.
* **Tests.** The full `go test ./tools/gatecheck/...` shows no new failures versus the parent
  baseline, with the same test count (record `go test -list '.*'
  ./tools/gatecheck/internal/writepath` before and after).
* **AC:**
  1. Structural: every well-formed-listing `GitRunner` fake in the writepath tests returns through
     `lsFilesListing` (or empty bytes). The U1 diff contains no listing literal other than the
     helper's own. Verify by reading the diff, not by a negative `rg` (a "finds nothing" grep would
     match the malformed-listing rows U2 adds and is not a valid gate).
  2. The tests stay green and the test count is unchanged (characterization).
  3. The diff touches only the two test files. The oracle diff is limited to the decode block and
     the header paragraph above; **if anything else in the oracle file would have to change, stop
     and report (H-3)**.
  4. The two tightened assertions are labelled characterization in the task notes.
  5. `gofmt -l`, `go vet ./tools/gatecheck/...` clean; `goimports -l` and `golangci-lint run` clean
     when installed (CI runs both).
* **Size / complexity:** XS / trivial.

### U2: NUL-safe selection (`-z`) with the selected-path rules (ONE commit)

* **Goal.** R1, R2, R3 (and R9).
* **Files:**
  * `tools/gatecheck/internal/writepath/writepath.go`
  * `tools/gatecheck/internal/writepath/writepath_test.go`
* **Why one task.** The runner-format flip, the parse, and every fake and decode that depends on it
  must land together: any split leaves a commit whose fakes are stale, or one that echoes raw
  bytes. There is no pin refreeze (D-BW-4). Recorded as EX-3.
* **Change (`writepath.go`):**
  * `DefaultGitRunner` argv becomes `exec.Command("git", "ls-files", "-z", "--", "internal/**", "cmd/**")`.
    Update the `GitRunner` and `DefaultGitRunner` doc comments: NUL-terminated records; no
    `pysem.GitText` decoding.
  * Add the unexported helper and two unexported sentinels (so tests assert `errors.Is`, not message
    text). Its name avoids a near-collision with retiredarch's frozen `selectRepoPaths`:

    ```go
    var (
    	errListingNotTerminated = errors.New("output is not NUL-terminated")
    	errListingEmptyRecord   = errors.New("output has an empty record")
    )

    // scannedPathsFromListing parses the NUL-terminated `git ls-files -z` listing
    // and returns the paths shouldScan selects, in git's order. It fails closed on
    // anything that is not exactly a NUL-terminated list of non-empty records. A
    // record that shouldScan rejects is never read or echoed, so it is judged on
    // nothing else. A selected path must be valid UTF-8 and must not contain a
    // control or line-separator character, because selected paths are echoed to
    // CI logs (D-BW-1a).
    func scannedPathsFromListing(out []byte) ([]string, error) {
    	if len(out) == 0 {
    		return nil, nil
    	}
    	if out[len(out)-1] != 0 {
    		return nil, errListingNotTerminated
    	}
    	var selected []string
    	for _, p := range strings.Split(string(out[:len(out)-1]), "\x00") {
    		if p == "" {
    			return nil, errListingEmptyRecord
    		}
    		if !shouldScan(p) {
    			continue // by design: never read or echoed
    		}
    		if !utf8.ValidString(p) {
    			return nil, fmt.Errorf("path %q: %w", p, pysem.ErrInvalidUTF8)
    		}
    		for _, r := range p {
    			if r < 0x20 || r == 0x7f || r == 0x85 || r == 0x2028 || r == 0x2029 {
    				return nil, fmt.Errorf("path %q contains a control or line-separator character", p)
    			}
    		}
    		selected = append(selected, p)
    	}
    	return selected, nil
    }
    ```

    The rune set is identical to retiredarch's (`< 0x20`, `0x7f`, `0x85`, `0x2028`, `0x2029`); the
    `SplitLines` boundaries U+001C..U+001E are already inside `< 0x20`. Any extension (C1 controls,
    bidi overrides) routes to stash `31F33EFE` and `EDC18D59`, not to this plan.

    The invalid-UTF-8 error is **wrapped with the `%q` path** so the log names the offending record
    without echoing raw bytes; `errors.Is(err, pysem.ErrInvalidUTF8)` and every planned row below
    still pass (INV-D). **Ship adds a code comment on `scannedPathsFromListing`** noting that
    writepath judges UTF-8 validity and control/separator runes only for the records it will scan,
    whereas retiredarch's `selectRepoPaths` judges every listed record (the asymmetry tracked by
    stash `31F33EFE`); cite the stash ID, not a unit tag (`4995F8C3`).
  * `runRepoScan`: replace the `pysem.GitText` + `SplitLines` block (`:835-848`) with
    `relPaths, err := scannedPathsFromListing(out)`; on error return
    `Result{Stderr: errorLine("git ls-files output", err), Code: 1}`. The ED-7 guard, the
    `containedRegularFile` loop and the findings block are unchanged. `errors`, `fmt`, `strings`
    and `unicode/utf8` are already imported, and `pysem` stays used.
  * Do **not** edit `scanFile` or `scanSource` (INV-A).
* **Change (`writepath_test.go`):**
  * flip `lsFilesListing` to `strings.Join(paths, "\x00") + "\x00"` (still `nil` for none);
  * flip `decodeLsFilesListing` to an **independent** bytes-based NUL framing (empty means no paths;
    require a trailing NUL; `bytes.Split`; reject empty records; **no UTF-8 policy**, so it tolerates
    the unscanned `internal/\xff.txt` that the production helper also tolerates; the oracle applies
    `shouldScan` and the file read afterwards). It must not call `scannedPathsFromListing`. Keep the
    oracle file itself unchanged in U2: the flip is the second half of the PA-2 edit, recorded as
    such;
  * add `fixtureGit(t, dir, args...)`, a copy of `retiredarch/select_test.go`'s scrubbed helper
    (every `GIT_*` entry removed, case-folded, skipping empty names, `GIT_CONFIG_NOSYSTEM=1`,
    `GIT_CONFIG_GLOBAL=os.DevNull`, and `-c user.name`, `-c user.email`, `-c commit.gpgsign=false`).
    It builds fixture repos; it does not use the code under test. Update the import block (`bytes`,
    `unicode/utf8` and others as needed) in `goimports` order;
  * add `hermeticGitEnv(t)`: removes every ambient `GIT_*` variable from the **test process**
    (case-folded; skip empty names; `t.Setenv(k, "")` to register restoration, then
    `os.Unsetenv(k)` with its error checked) and pins `GIT_CONFIG_NOSYSTEM=1` and
    `GIT_CONFIG_GLOBAL=os.DevNull`. Tests (d) here, and e1 in U3, call it first;
  * write `é` as `\u00e9` in test source, as the retiredarch precedent does, so an editor cannot
    normalise it.
* **Posture.** Test-first, staged with a **compiling stub** so every red is behavioural, never an
  `undefined:` compile failure:
  1. At the parent, add a stub `scannedPathsFromListing` that returns
     `nil, errors.New("unimplemented")` (it is not committed) **and that also declares the two
     sentinels `errListingNotTerminated` and `errListingEmptyRecord`** (test (a) references both by
     name; without them the red is an `undefined:` compile failure, which this plan forbids), plus
     the tests and helper flips, and run them. Observe the reds below.
  2. Apply the real `writepath.go` change.
  3. Run all green; commit once.

  Record the red evidence, `git --version`, and `git config --show-origin --get core.quotepath` in
  the task implementation-notes and the PR body.

**Tests** (3 test functions, table-driven, to keep the scenario count manageable; EX-3). The "Red
kind" column separates *bug-evidence*, *new-contract* and *characterization*:

| # | Test | Red kind | Parent behaviour | After |
|---|---|---|---|---|
| d | `TestRunRepoScan_NonASCIIPath_IsScanned_RealGit`: a real `git init` temp repo (full Go files with a `package` clause) tracking a clean `cmd/x/main.go` and `internal/\u00e9.go` whose body holds a write primitive (for example `os.Remove`). **Call `hermeticGitEnv(t)` first.** Otherwise a host `core.quotePath=false` or a hook-exported `GIT_INDEX_FILE` makes the red host-dependent | **bug-evidence (required)** | Real git C-quotes `internal/\u00e9.go`, the parent drops it: `runRepoScan(root, DefaultGitRunner)` returns **Code 0** (false-clean) | Code 1; stderr names `internal/\u00e9.go` (compared via `filepath.ToSlash`) **and contains the finding text** `write primitive 'os.Remove' found`, so a parse error cannot satisfy it |
| a | `TestScannedPathsFromListing_Table`: helper-level success rows over `scannedPathsFromListing(lsFilesListing(...))`: `internal/\u00e9.go`, `internal/q"uote.go`, `internal/back\slash.go`, `internal/sp ace.go`, `cmd/x/main.go` are selected byte-exact in listing order; unscanned `internal/README.md`, `internal/x_test.go`, `internal/tab\there.txt` (a control rune in an unscanned name) and `internal/\xff.txt` (invalid UTF-8, unscanned) are tolerated with no error. Helper-level error rows: `errors.Is(err, errListingNotTerminated)` for a listing without the trailing NUL; `errors.Is(err, errListingEmptyRecord)` for an empty record; `errors.Is(err, pysem.ErrInvalidUTF8)` for the scanned record `internal/\xff.go`. Every error row asserts a specific error, never bare `err != nil`. A helper-level test is required because names such as `q"uote.go` and `back\slash.go` cannot be created on a Windows filesystem; the control-rune rows live in (b), where the `%q` message is observable | new-contract (red through the stub's `unimplemented` error) | Stub returns an error for every row | The exact scanned paths, byte-exact, in order; each error row returns the expected error |
| b | `TestRunRepoScan_BadListing_FailsClosed`: `runRepoScan` over a temp root and a fake runner. Rows: the named constant `legacyNewlineListing = "internal/a.go\n"` (a clean `internal/a.go` exists); empty record `"internal/a.go\x00\x00"`; lone `"\x00"`; a scanned path containing `\n`, U+2028, U+0085, ESC, DEL, U+2029 and `\x1f`; invalid UTF-8 in a scanned record `"internal/\xff.go\x00"`; an all-unscanned listing `lsFilesListing("internal/README.md")` expecting ED-7. These raw literals are the only malformed listings in the package; every well-formed fake goes through `lsFilesListing` | new-contract; the `legacyNewlineListing` row is also bug-evidence; the invalid-UTF-8 row is **characterization** | Parent: the newline row returns Code 0 (red); the other rows return ED-7 or other text, not the asserted message (red); the invalid-UTF-8 row (`ErrInvalidUTF8`) is already Code 1 (green); the all-unscanned row is already ED-7 (green) | Each row Code 1 with its specific message (`not NUL-terminated`, `empty record`, `control or line-separator character`, `pysem.ErrInvalidUTF8.Error()`, `ED-7`); the control rows' stderr contains the `%q` escape and **no raw control byte** once the trailing newline is trimmed |

* The existing `TestRun_*` fakes now emit NUL-terminated listings via the helper; they are the
  new-contract reds at the parent for the flipped fakes.
* `TestOracle_TrackedProductionTree_Parity` now decodes through the flipped helper and must stay
  green on the live tree. **The oracle file diff in U2 is empty.**

* **AC:**
  1. (d) is OBSERVED red at the parent (Code 0) before any real `writepath.go` change, and green
     after. Evidence is recorded with the git version and the `core.quotepath` origin (run under the
     same hermetic environment).
  2. (a) and (b) new-contract reds are observed against the stub and the parent; the invalid-UTF-8
     and all-unscanned rows are labelled characterization. All green after.
  3. `go test -count=1 ./tools/gatecheck/...` shows no new failures versus the parent baseline. On
     the live tree, `go run ./tools/gatecheck write-path --root .` and
     `go run ./tools/gatecheck write-path --root . --self-test` exit 0, and
     `go run ./tools/gatecheck write-path --root . --self-test-integrity` exits 0 with the goldens
     unchanged. **Regression guards, green at the parent; not fix verification.**
  4. `git diff` shows `scanFile` and `scanSource` unchanged, and `writepath_oracle_test.go`
     unchanged in U2; `retiredarch` `TestWritePath*` and `initstate_test.go` are green (INV-A,
     INV-B). Regression guard.
  5. The contract-surface matrix holds: every well-formed-listing fake returns through
     `lsFilesListing` and the malformed-listing rows of (b) are the only raw literals; the
     `GitRunner` and `DefaultGitRunner` doc comments no longer say "for pysem.GitText decoding".
  6. Re-measure at the U2 base `1b130a2` (or the Ship branch base) and record the counts and
     `git --version` in the evidence: tracked `internal/**` + `cmd/**` files (56 at staging), paths
     needing quoting (0), tracked symlinks (0), tracked direct `.go` fixtures (28).
  7. `gofmt -l`, `go vet ./tools/gatecheck/...` and `go build ./...` are clean; `goimports -l` and
     `golangci-lint run` are clean when installed (CI runs both).
* **Size / complexity:** M / medium.

### U3: git environment isolation for `DefaultGitRunner` (D7BF9F74 shape)

* **Goal.** R4, R5.
* **Files:**
  * `tools/gatecheck/internal/writepath/writepath.go`
  * `tools/gatecheck/internal/writepath/writepath_test.go`
* **Change (`writepath.go`):**
  * Add `gitRunnerEnv(environ []string) []string`, a semantic copy of
    `retiredarch.gitRunnerEnv`: drop every entry whose name (the text before the first `=`,
    **ASCII-case-folded**) starts with `GIT_`, except `GIT_CEILING_DIRECTORIES`; keep Windows
    per-drive `=C:`-style entries; then append `GIT_CONFIG_NOSYSTEM=1`,
    `GIT_CONFIG_GLOBAL=`+`os.DevNull`, `GIT_CONFIG_SYSTEM=`+`os.DevNull`. The doc comment cites
    D7BF9F74, says "copied, not shared (D-BW-2, 5A8EC1BC)" and "keep in sync by hand with
    `retiredarch.gitRunnerEnv`; the pure table test must be kept identical".
  * In `DefaultGitRunner`, after `cmd.Dir = root`, set `cmd.Env = gitRunnerEnv(os.Environ())`.
    Update the doc comment with a short threat-model paragraph (mirroring the retired-architecture
    D6b stance): the adversary is *ambient* environment (agent hosts, git hooks, wrapper scripts);
    git 2.32 or later is assumed for the config pins; a `PATH` shim, an in-job attacker and a
    self-modifying PR are out of scope (R-A2b); the wrapper's root discovery is a separate surface.
  * **No** `cmd.Err` or non-absolute-path refusal (D-BW-2 Option B is out of scope).
  * Do not touch `scanFile`/`scanSource`; no `filepath.Abs` or `EvalSymlinks` (INV-A, INV-B).
* **Change (`writepath_test.go`):** `fixtureGit` and `hermeticGitEnv` (from U2) are reused. Add a
  small `fixtureGitEnv(t, dir, extraEnv []string, args ...string)` that applies the extra entries
  **after** the scrub and always sets `cmd.Dir`; `fixtureGit` becomes a thin wrapper. Add a
  `git ls-files -z` capture helper that uses the scrubbed environment and `cmd.Output()` (stdout
  only: git's stderr hints would corrupt the bytes) with the runner's pathspec order.
* **Posture.** Test-first, staged with a **compiling identity stub** so reds are behavioural:
  1. At the parent, add a stub `gitRunnerEnv` that returns its input unchanged (not committed), plus
     both tests, and run them. Observe the reds below.
  2. Replace the stub with the real filter and set `cmd.Env`.
  3. Run all green; commit once.

**Tests** (2 test functions; EX-4):

| # | Test | Red kind | Detail |
|---|---|---|---|
| e1 | `TestDefaultGitRunnerIgnoresGitEnv` | **bug-evidence for vector (a)**; new-contract for (b), (c) | A real-git fixture repo (full Go files with a `package` clause) tracks a clean `cmd/x/main.go`, a clean `internal/ok.go`, a violating `internal/bad.go` (for example `os.Remove`) and `internal/\u00e9.go`. **Baseline** = the scrubbed `git ls-files -z -- internal/** cmd/**` stdout bytes (the runner's pathspec order); the test first asserts the baseline names all four files (a vacuous baseline cannot pass). **Each vector is its own `t.Run` subtest** that calls `hermeticGitEnv(t)` first and then sets only that vector with `t.Setenv`, **after** the repo and the alternate index are built (vectors are never cumulative). A **no-vector subtest** asserts `DefaultGitRunner(root)` returns bytes equal to the baseline (green at the parent and after: the causal control). For every vector the oracle is: `DefaultGitRunner(root)` returns `err == nil` and bytes **equal to the baseline**. Never compare the exit `Code` alone. **(a)** `GIT_INDEX_FILE` = an alternate index listing only `cmd/x/main.go`, created with `fixtureGitEnv`. At the parent the bytes differ, and `runRepoScan(root, DefaultGitRunner)` returns **Code 0** although tracked `internal/bad.go` holds a write primitive; after, the baseline, and Code 1 with a stderr finding `write primitive 'os.Remove' found` for `internal/bad.go`. **(b)** `GIT_LITERAL_PATHSPECS=1`: the parent returns an empty listing. **(c)** `GIT_CONFIG_PARAMETERS=not-a-valid-config`: the parent runner errors |
| e2 | `TestGitRunnerEnv` | new-contract (red through the identity stub) | A pure table, **identical to retiredarch's** so drift fails a test: lower/mixed-case `git_dir` and `Git_Index_File` dropped; `GIT_CEILING_DIRECTORIES`, `Git_Ceiling_Directories` and `git_ceiling_directories` kept; `=C:=C:\x` kept; `GITX=1` kept; `GIT_CONFIG_NOSYSTEM=0` and `GIT_CONFIG_GLOBAL=...` replaced; the three appended entries last; `nil` input yields just the three entries |

* **Why three vectors, not five.** The parent-red proves each vector is live when written. The
  `GIT_CONFIG_GLOBAL` pin is covered by the pure table's replaced-entry row. The
  `GIT_CONFIG_COUNT`/`KEY_n`/`VALUE_n` channel is covered by the generic `GIT_` prefix rule, **not
  independently observed**: no table row and no vector exercises it, and (c) exercises the same
  prefix rule through `GIT_CONFIG_PARAMETERS`. Do not add a COUNT/KEY row to the table: it would
  break the identical-to-retiredarch invariant. The mandatory-liveness-control design in stash
  `C8827920` is **not** adopted here; it stays with that entry for the retiredarch table.
* **Test hygiene.** A missing `git`, or an unparseable `git --version`, is `t.Fatalf`, never a skip.
  Failure messages and the PR evidence never print `os.Environ()` or `GIT_CONFIG_VALUE_*` (agent
  hosts use them for credential headers). Extra environment entries are applied after the scrub and
  `cmd.Dir` is always set. No `t.Parallel` (process-global env).
* **AC:**
  1. Vector (a)'s false-clean (Code 0) is OBSERVED at the parent before any production change; (b)
     and (c) are observed red. Evidence names the git version and does not print the environment.
  2. All green after. `go test -count=1 ./tools/gatecheck/...` shows no new failures versus the
     parent baseline. Regression guards, green at the parent: live-tree `write-path` exits 0 and
     goldens unchanged.
  3. `retiredarch` `TestWritePath*` and `initstate_test.go` green (INV-A, INV-B). Regression guard.
  4. `gofmt`, `go vet`, `go build ./...` clean; `goimports` and `golangci-lint` clean when installed.
* **Size / complexity:** S / medium.

### U4: fixture self-test containment (05E12A6F)

* **Goal.** R6.
* **Files:**
  * `tools/gatecheck/internal/writepath/writepath.go`
  * `tools/gatecheck/internal/writepath/writepath_test.go`
* **Change (`writepath.go`).** In `runFixtureSelfTest`'s loop, **before** `scanFile`:

  ```go
  if ok, reason := containedRegularFile(root, relPath); !ok {
      return Result{Stderr: errorLine(relPath, fmt.Errorf("not a contained regular file: %s", reason)), Code: 1}
  }
  ```

  This is abort-on-first semantics, consistent with the repo scan and with writepath's ED-2 read
  abort. The `fixture dir not found` and `no fixtures discovered` messages are untouched. A deleted
  fixture falls through to `scanFile` (`fs.ErrNotExist` returns ok), keeping the existing read-error
  text. Do **not** edit `scanFile` (INV-A). Update the `runFixtureSelfTest` doc comment, and state
  the accepted TOCTOU residual. The new message echoes the raw fixture name, the same sink class as
  stash `EDC18D59`; do **not** switch it to `%q` here (the row assertions pin the raw text), and
  record the site on that entry in the PR body.
* **Posture.** Test-first. Add one `linkDir(t, link, target)` test helper that creates a junction on
  Windows (`u4Junction`) and a directory symlink elsewhere (`u4CreateSymlink`). The two existing
  helpers take swapped, same-typed arguments, so directory links (rows 2 and 3) use only the
  wrapper. Row 1 is a *file* symlink and calls `u4CreateSymlink` directly, so its Windows-privilege
  skip is preserved. Fixture bodies are full Go files: `reject-*.go` holds a write primitive,
  `accept-*.go` holds `u4CleanGo`.
* **Test** (1 table-driven function, `TestRunFixtureSelfTest_LinkedFixture_FailsClosed`; EX-5).
  Every link is a LIVE link **at the entry itself** (not dangling, not through a subpath).
  Everything is inside one `t.TempDir()`:

  | Row | Setup | Red kind | Parent | After |
  |---|---|---|---|---|
  | 1 leaf file symlink | `scripts/testdata/writepath/reject-x.go` is a symlink (`u4CreateSymlink`) to a file **outside root** holding a write primitive | **bug-evidence** | Code 0, `PASS reject-x.go: rejected as expected` (the link is read) | Code 1, stderr contains `::error::scripts/testdata/writepath/reject-x.go: not a contained regular file: `, and no `PASS reject-x.go` line |
  | 2 leaf directory link | `reject-dirlink.go` is a directory link (`linkDir`: a junction on Windows, which needs no privilege; a symlink elsewhere) | new-contract (a message-only red; **not** bug evidence) | Not asserted at plan time (see below); the red is observed, not assumed | Code 1 with the containment text, and no `PASS reject-dirlink.go` line |
  | 3 linked ancestors | three sub-cases, each linking one ancestor to a real directory holding a fixture: `scripts`, `scripts/testdata`, `scripts/testdata/writepath` (`linkDir`) | **bug-evidence** | Code 0 (`os.Stat` and `ReadDir` follow the link) | Code 1; the reason names the linked ancestor (`... is not a real directory`) |

  The regular-fixture control (a regular `harness/` subdirectory plus regular fixtures) is already
  covered by the live-tree goldens (`TestRun_SelfTest*_MatchesGolden`), so it has no row of its own;
  it is a regression guard in the AC.

  Symlink creation skips **only** on the Windows privilege error (`u4CreateSymlink` already does
  this), so row 1 skips on an unprivileged Windows host; rows 2 and 3 use a junction there and never
  skip. The Linux CI job runs all three (the skip is guarded by `runtime.GOOS == "windows"`, so
  Linux cannot skip row 1). Row 3 relies on the same ancestor detection that
  051.004-T's `TestRunRepoScan_IntermediateJunctionOrSymlink_FailsClosed` already exercises, so its
  red is observable on an unprivileged Windows dev host. Record where each red was observed. On such
  a host, rows 2 and 3 are the P-004 red evidence for U4's one production call, and row 1's red
  comes from a symlink-capable host (WSL, Developer Mode) when one exists (AC1).

  **Windows classification (optional trace, no AC).** A reviewer read the local toolchain's
  `$GOROOT/src/os/types_windows.go` and found a mount point is `ModeIrregular` only (not `IsDir()`,
  not `IsRegular()`), so a leaf junction is kept by `ReadDir` and refused with the containment text
  (consistent with 051.004-T's `TestRunRepoScan_FinalDirectoryLink_FailsClosed`). Ship re-confirms
  on the pinned go1.26.9 toolchain and records the result in the task notes. The same trace should
  cover a Windows directory symlink, or mark that behaviour as unverified; no assertion depends on
  it.
* **AC:**
  1. **Red evidence (P-004) for U4's one production call.** Row 3 (a linked ancestor; bug-evidence;
     a junction on an unprivileged Windows host, a symlink elsewhere) and row 2 (a leaf directory
     link; red wherever a directory link can be created) are OBSERVED red at the parent, before the
     production change, and green after. Row 1 (the leaf file symlink; bug-evidence) is observed
     red from a symlink-capable local host **in the same checkout** (WSL, Developer Mode; no second
     clone or worktree) when one exists. When it cannot be (an unprivileged Windows host skips it
     with Errno 1314), record that in the task evidence and the PR body: rows 2 and 3 stand as the
     red evidence for the same one-line production change, and row 1 must pass un-skipped on the
     Linux CI test step. **Do not push a standalone `test:` commit to CI for U4:** `ci.yml` sets
     `cancel-in-progress: true` per workflow and ref, so a fix pushed while that run is in flight
     cancels the red run and leaves no evidence, and a second commit would break U4's
     single-commit revert. If neither row 2 nor row 3 can be observed red (no directory link can be
     created on any local host), halt under P-005.
  2. `TestRun_SelfTest*_MatchesGolden` unchanged (this covers regular fixtures and the `harness/`
     subdirectory); `scanFile`/`scanSource` unchanged; `retiredarch` mask-flow and
     `initstate_test.go` green. Regression guards.
  3. On the live tree, `go run ./tools/gatecheck write-path --root . --self-test-integrity` and
     `go run ./tools/gatecheck write-path --root . --self-test` exit 0. Regression guard.
  4. The PR body includes a **local** `go test -v -run TestRunFixtureSelfTest_LinkedFixture
     ./tools/gatecheck/internal/writepath` run showing rows 2 and 3 as PASS, never SKIP (they use a
     junction on Windows). Row 1 also shows PASS, **except** on an unprivileged Windows host, where
     it may show SKIP (Errno 1314) provided the skip line is recorded verbatim in the task
     evidence together with where row 1 passed (a WSL or Developer-Mode run, or the Linux CI test
     step, which cannot skip it). The CI test step has no `-v`, so CI cannot show subtest verdicts;
     adding one is out of scope.
  5. `gofmt`, `go vet`, `go build ./...`, `go test -count=1 ./tools/gatecheck/...` clean;
     `goimports` and `golangci-lint` clean when installed.
* **Size / complexity:** S / low.

### U5: correct the stale cross-package comment (comment only)

* **Goal.** R8.
* **File:** `tools/gatecheck/internal/retiredarch/select.go` (the `GitRunner` doc comment,
  `:25-31`).
* **Change.** Replace "Its signature matches writepath.GitRunner, but writepath's listing is
  newline-separated and not -z (deferred entry B83F53BB), so the two outputs differ." with an
  accurate statement: retiredarch's runner takes the pathspecs as arguments, writepath's takes only
  the root and fixes its own pathspecs, and **both** now emit `-z` NUL-terminated records
  (B83F53BB). Comments are not pinned (`pin.go` tokenises declarations without comments); change
  no declaration and no directive-shaped (`//go:`, `// +build`) comment. The file header's
  "same commit as pin.go" rule governs token changes, so a comment-only edit is exempt.
* **Posture.** Documentation only, **green before and after (EX-2)**. No red, and none is faked.
* **AC:**
  1. `git diff -U0` shows only `//` comment lines in one file.
  2. The new comment text contains no `pin_test.go` mutation anchor verbatim. Run `rg -n -F
     '<anchor>' tools/gatecheck/internal/retiredarch/select.go` for each anchor string in
     `pin_test.go`: each must still resolve **exactly once** (its code line) after the edit, and no
     added comment line may contain one.
  3. `go test ./tools/gatecheck/internal/retiredarch/...` is green (the pin is unaffected) and
     `gofmt -l` is clean.
  4. The phrase "newline-separated" no longer appears in the comment.
* **Size / complexity:** XS / trivial.

## Dependency Graph

```text
U1 (test prep, oracle decode) --blocks--> U2 (-z selection)
U2 (-z selection)             --blocks--> U3 (env isolation)   same function, same test fixtures
U2 (-z selection)             --blocks--> U5 (comment fix)     the comment must be true when it lands
U4 (fixture containment)         independent
```

* **Suggested execution order:** U1, U2, U3, U4, U5.
* **No cycles.** U4 is independent of U1-U3 (different function, uses the existing link helpers),
  but is sequenced after U3 so that the serial commits touch `writepath.go` in one direction.

## Decisions and Rationale

See the deliberation for the full options tables.

* **D-BW-1: `-z` with a strict, unexported NUL-record helper.** The only option that closes the
  quoting, `\r` and line-boundary hazards at once. The parse is not inline because `writepath.go` is
  unfrozen and a table test through a function is clearer than one through a `Result`.
* **D-BW-1a: UTF-8 validity and control/separator runes are judged only for paths that will be
  scanned.** Only scanned paths are read and echoed in the listing arm, so the rule is minimal and
  sufficient, and it avoids the false positive recorded in `31F33EFE`. It diverges from retiredarch
  until `31F33EFE` is decided; an operator who wants parity can switch it in one line and one table
  row. It closes the *new* raw-byte path that `-z` opens; pre-existing echo sinks are captured
  separately.
* **D-BW-2: an in-package copy of `gitRunnerEnv`, no PATH guard, three vectors.** Matches
  B83F53BB's stated scope, closes the one fail-open vector (`GIT_INDEX_FILE`), and keeps 5A8EC1BC
  untouched. The adversary is ambient environment, not an in-job attacker.
* **D-BW-3: reuse the existing `containedRegularFile` for fixtures.** No new code or copy; covers
  the leaf and every ancestor.
* **D-BW-4: a prep task, a single authorized oracle edit, and no pin refreeze.** Nothing here edits a
  pinned declaration.
* **D-BW-5: one feature, five tasks, no sub-epics** (the 050-F/051-F precedent).

## Risks and Caveats

| ID | Risk | Likelihood / impact | Mitigation |
|---|---|---|---|
| RK-1 | A test fake still emits newline listings after U2 and passes for the wrong reason | low / high | The strict parser rejects a non-NUL-terminated listing; U1 centralises fakes and tightens two assertions to name the path; the matrix check is repeated at U2 |
| RK-2 | The frozen-oracle edit is treated as an H-3 violation | low / medium | Recorded as a single authorized adaptation (D-BW-4, INV-G) with the exact replacement header written out; stop and report if the diff exceeds the decode block and the header paragraph |
| RK-3 | An env vector is non-live on the CI git version | low / medium | Each vector is observed red at the parent; the baseline is asserted non-vacuous; bytes are compared, not exit codes |
| RK-4 | Symlink tests skip on unprivileged Windows, so a red is never observed locally | high / low | Junction rows run unprivileged on Windows and are U4's red evidence (rows 2 and 3); row 1's red comes from WSL or Developer Mode, or is disclosed as unobserved; Linux CI runs the symlink rows; record where each red was observed |
| RK-5 | A U2-U4 change breaks the mask-flow pin | low / high | No edit to `scanFile`/`scanSource` (INV-A); the pin tests are in every AC |
| RK-6 | Isolation breaks the gate on a runner or host with an injected global `safe.directory` | low / medium | The retired-architecture gate already runs in the same CI job and checkout with the same isolation (PRs #112, #116) and passes; the first CI run on the Ship PR is the validation window. If live-tree tests fail on the Ship host, inspect the injected keys with `git config --show-origin`, treat it as an environment issue, and do not loosen the scrub |
| RK-7 | Path bytes reach CI logs unescaped (workflow-command injection) | low / medium | D-BW-1a: scanned paths with control/separator runes fail closed with `%q`; test (b) asserts no raw control byte. Other sinks are captured separately |
| RK-8 | A tracked `.go` path that is not valid UTF-8 now fails the whole gate with `ErrInvalidUTF8` | low / low | Intended fail-closed behaviour (INV-D). An unscanned non-UTF-8 name is tolerated (D-BW-1a) |
| RK-9 | Two hand-synced copies (`gitRunnerEnv`, `containedRegularFile`) drift from retiredarch | medium / low | The pure env table is identical to retiredarch's; doc comments cite each other; the 5A8EC1BC evidence note proposes extending its reconsideration trigger to repeated cross-engine fixes. A paired-review expectation applies whenever either copy changes |

## Plan Hardening Signals

* **Public API, schema or contract change: PRESENT.** The `GitRunner` output contract (newline to
  NUL) changes for the writepath gate, and four new fail-closed outcomes appear (malformed
  listing, non-UTF-8 or control-character scanned path, linked fixture).
* **Security, auth, permission or compliance-sensitive behaviour: PRESENT.** These are
  merge-blocking integrity gates; the changes convert fail-open states to fail-closed.
* **Migration, backfill, destructive data/config action or irreversible step: ABSENT.** Code-only,
  and fully revertible.
* **External integration, operator checkpoint or external dependency: PRESENT (minor).** The real
  `git` binary's `-z` output and its config-channel behaviour, and a written authorization for the
  frozen-oracle edit (given by the staging-PR approval).
* **High runtime, rollout or rollback risk: ABSENT.** CI-only tooling; each task is a
  single-commit revert (a U2 or U3 that used the blocked-path CI fallback is two commits, reverted
  together).

Requires plan hardening: yes

## Runtime Verification and Closure

| Unit | Runtime surface | Verification before absorption | Closure artifact |
|---|---|---|---|
| U1 | none (tests) | `go test -count=1 ./tools/gatecheck/...`, same test count | task notes |
| U2 | `gatecheck write-path` CLI (repo, `--self-test`, `--self-test-integrity`); the CI write-path steps; the wrapper `scripts/check-write-path-precondition.sh` | Live tree exits 0; `--self-test-integrity` golden unchanged; real-git non-ASCII red then green (d) | PR body red/green evidence |
| U3 | same CLI, wrapper and CI steps | Live tree exits 0; each vector red then green; (a) false-clean red then green | PR body evidence including the git version (no environment dump) |
| U4 | `gatecheck write-path --self-test-integrity` and CI | Live tree exits 0; rows 2 and 3 red then green locally (row 1's red from a symlink-capable host when one exists); a local `go test -v` run shows the rows as PASS (row 1 SKIP is permitted on an unprivileged Windows host if recorded, U4 AC4) | PR body evidence (OS where each red was observed) |
| U5 | none (comment) | `go test ./tools/gatecheck/internal/retiredarch/...` green | task notes |

* **Rollback trigger.** Any of the write-path CI steps failing on `main` after the merge with a
  non-finding error. Revert the offending task's commit in reverse dependency order: U5 and U3
  before U2; U4 at any time. U1 (characterization) may stay, because reverting U2 restores the
  newline-format helpers it flipped. A unit that used the blocked-path CI fallback (a standalone
  `test:` commit, then the fix) is two commits: revert **both**, the fix first and then its `test:`
  commit, because reverting only the fix leaves a red test on `main`.
* **Validation window.** The first CI run on the Ship PR and the first run on `main` after the
  merge. Record the `WRITE_PATH_GATE_ADVISORY` repository-variable value in the PR evidence (unset
  or any non-`true` value means the verdict step is blocking).
* **Owner.** Ship (shipment), then Orchestrator closure.
* **Closure items owned by Ship:** verify the merge strategy is a merge commit (P-009, constitution
  XI) before merging; run P-018 Copilot engagement before merge (see
  `docs/compound/2026-09-10-p018-copilot-engagement-must-precede-merge.md` and
  `docs/compound/2026-09-06-requesting-copilot-review-non-collaborator-repo.md`); run the P-020
  compact-context at post-merge closure; carry `.backlogit/stash.jsonl` (the three new stash
  entries and the 5A8EC1BC evidence note) forward on the next authorized branch (see
  `docs/compound/workflow-issues/stash-jsonl-is-carried-pipeline-state-2026-09-19.md`).
* **Closure doc hygiene.** Record in the PR body and the closure record: (1) the contract change,
  writepath `GitRunner` newline to NUL, plus the four new fail-closed outcomes; (2) that the
  isolation covers the `git ls-files` child **for a given root** and **does not** claim end-to-end
  environment independence; (3) the residual stash IDs `DBE25DF5`, `EDC18D59`, `4372BAD4` and
  `31F33EFE`. During compound capture, Ship refreshes
  `docs/compound/2026-10-07-pin-canonical-text-maintenance-and-unbounded-ast-rule-review.md` only if
  the work disproves it. No separate docs task: the contract-surface matrix above is a named-surface
  read and shows no operator-facing document states the old contract.
* **Harvest notes.** Confirm `backlogit --version` and PATH order before harvest. Probe `size` and
  `complexity` per artifact type and read every mutation back: the workspace task type defines no
  `complexity` field (see `docs/compound/2026-09-04-backlogit-sizing-is-wit-gated.md`), so
  complexity is recorded as labelled prose and the degradation is flagged. Keep CLI text ASCII
  (write `U+00E9`, not the character) and author markdown with single-quoted here-strings or the
  editor tools; scan generated text for control characters below 0x20.

## Plan Hardening

*Appended by plan-harden, Stage session 2026-10-10.*

**Hardening need confirmed:** three signals are present (contract change, security-sensitive merge
gate, external `git` dependency). The plan needs no new units; hardening tightens verification,
rollback and the reviewable action record.

**Protected invariants:** INV-A to INV-G (Scope Boundaries). The most load-bearing is INV-A, because
`retiredarch/writepath_mask_test.go` fails the whole retiredarch package if `scanFile` or
`scanSource` changes shape.

**Learnings and instruction files consulted:** the compound documents listed at the top of this
plan, `.github/instructions/strict-safety.instructions.md`,
`.github/instructions/constitution.instructions.md`, `.github/instructions/technology-go.instructions.md`.

**Proposed actions and risk (strict-safety vocabulary):**

| ProposedAction | Targets | change_kind | ActionRisk | Approval needed | ActionResult | Rollback / notes |
|---|---|---|---|---|---|---|
| PA-1: flip the writepath `GitRunner` output to NUL, the strict parse, and every fake and decode that depends on it (U1 prep, U2) | `writepath.go`, `writepath_test.go` | local edit; contract change to a merge-blocking gate | **high** (security-sensitive contract change) | The staging-PR approval pre-approves PA-1 as planned (strict-safety rule 3); Ship PR review covers the diff | planned | `git revert` of U2 (reverse dependency order: U5, U3 first). Never leave a fake in newline format |
| PA-2: the authorized edits to the header-frozen oracle chain: the decode block and header paragraph of `writepath_oracle_test.go` (U1), and the flip of the `decodeLsFilesListing` helper it now calls (U2). **The oracle file itself is unchanged in U2** | `writepath_oracle_test.go` (decode block `:194-203`, header `:9-12`); `writepath_test.go` `decodeLsFilesListing` | local edit to a frozen test and its helper | **high** (frozen trust surface) | **Yes: the operator's staging-PR approval is the written authorization.** Ship records PA-2 as `approved` with the staging-PR reference **before** U1 edits the oracle, and halts if it cannot | planned | `git revert` U1 (and U2 first). Stop and report (H-3) if the oracle diff exceeds the decode block and the header paragraph |
| PA-3: scrub the `git ls-files` child environment (U3) | `writepath.go` `DefaultGitRunner`, `gitRunnerEnv` | local edit; subprocess environment change | **high** (security-sensitive, external binary) | The staging-PR approval pre-approves PA-3 as planned; Ship PR review covers the diff | planned | `git revert` U3; no state to restore |
| PA-4: add a containment refusal to the fixture self-test (U4) | `writepath.go` `runFixtureSelfTest` | local edit | moderate. Justification: it only *reduces* what the self-test reads, reuses an existing helper, and changes no contract or signature | No | planned | `git revert` U4 |
| PA-5: comment-only edit to the frozen-surface file `retiredarch/select.go` (U5) | `retiredarch/select.go` doc comment | comment edit | low | No; confirm the retiredarch tests stay green | planned | `git revert` U5. Change no declaration, no directive-shaped comment |

None is `destructive`. None needs the `agent-intercom` approval path. The `ActionResult` vocabulary
in use is `planned`, `approved`, `blocked`, `applied`, `rolled-back` and `abandoned`. **Ship records
the final `ActionResult` for each PA in the Ship PR body and the closure record.**

**Fail-closed invariants Ship must preserve (checked in every PR):**

* Every new branch in U2-U4 converts an unknown state into Code 1 or an error, never into Code 0 or
  a skip. The one deliberate `continue` in `scannedPathsFromListing` discards an *unscanned* path
  that is never read or echoed and carries a comment saying so; an empty record is an error.
* No new `t.Skip` except symlink-privilege skips (each names the CI job that runs the case). A
  missing `git` or an unparseable `git --version` is `t.Fatalf`.

**Verification commands.** Per task, Ship runs the scoped set with `-count=1` (a cached `ok` can mask
a missing red). Before the PR, Ship runs the repo-wide set in the constitution's order. `--root` is
mandatory for every direct `gatecheck` invocation (`tools/gatecheck/main.go` rejects a missing
`--root`); the wrapper derives it itself and also exercises root discovery.

```text
# per task (scoped)
gofmt -l tools/gatecheck
go vet ./tools/gatecheck/...
go test -count=1 ./tools/gatecheck/...
go build ./...
go run ./tools/gatecheck write-path --root .
go run ./tools/gatecheck write-path --root . --self-test
go run ./tools/gatecheck write-path --root . --self-test-integrity
go run ./tools/gatecheck retired-arch --root .
bash scripts/check-write-path-precondition.sh
# before the PR (constitution order, repo-wide)
gofmt -l .
go vet ./...
go test -count=1 ./...
go build ./...
goimports -l .
golangci-lint run ./...
# platform-dependent extras (see the notes below)
go test -race -count=1 ./tools/gatecheck/...
GOOS=windows golangci-lint run ./tools/gatecheck/...
GOOS=linux go vet ./tools/gatecheck/...
```

Confirm each subcommand's exact name from `register_*.go` first. Notes: `-race` needs cgo, so on a
Windows dev host without a C toolchain the Linux CI job is the evidence. On a Windows host
`GOOS=windows ...` is a no-op for vet, so run `GOOS=linux go vet` as the cross-platform check
(PowerShell: set `$env:GOOS = 'linux'` for the command, then restore it). CI additionally runs
`goimports`, `golangci-lint run ./...`, a `GOOS=windows` lint pass, and `staticcheck`/`govulncheck`
in the security job; those two are CI-only.

**Environment prechecks:**

* `git --version` is recorded in the evidence (do not print the environment).
* The agent host exports ambient `GIT_CONFIG_COUNT`/`KEY_n`/`VALUE_n`, `GIT_ASKPASS`,
  `GIT_OPTIONAL_LOCKS` and `GIT_TERMINAL_PROMPT`. Every fixture builder and `hermeticGitEnv(t)`
  must strip `GIT_*` first (case-folded). Do not rely on a clean ambient environment.
* **Windows host baseline.** 041-S's closure records `windows-host-env-test-failures` (the wrapper
  tests and one integration test fail identically at baseline: bash path mapping, exit 127 with the
  WSL shim). Prepend `C:\Program Files\Git\bin` to PATH, record the failing set **at the parent**
  before U1, and require "no new failures versus the parent baseline" instead of a literal green.
  The Linux CI `test` and race jobs are authoritative.
* On an unprivileged Windows host the file-symlink row (U4 row 1) skips; the junction rows do not.
  Record the SKIP line in the task evidence (U4 AC4) and take row 1's pass from a WSL run or the
  Linux CI test step.

**Blocked-path handling:**

* If a U2-U4 red cannot be observed on the dev host (for example, symlinks unavailable), use a
  symlink-capable local run (WSL, Developer Mode) **in the same checkout**; never a second clone or
  worktree. **U4 never uses a CI fallback:** its rows 2 and 3 are red on any host that can create a
  directory link (U4 AC1). For U2 and U3, whose reds need only a real `git` and an environment
  variable and so are observable on any dev host, a CI fallback is the last resort: push the
  `test:` commit alone, **wait for that CI run to finish red** (`ci.yml` sets `cancel-in-progress:
  true` per workflow and ref, so pushing the fix while it is in flight cancels the red run and
  leaves no evidence), record the failing run URL as the red evidence, then push the fix; this
  consumes fix-ci budget and is disclosed in the PR body. That unit is then **two commits**, and
  its rollback reverts both (see Rollback). Otherwise halt under P-005. A recorded green-only run is
  not acceptable for a required bug-evidence red.
* If a pin or retiredarch test fails after a writepath change, stop and report: never weaken the pin,
  and never edit `scanFile`/`scanSource` to satisfy it.

**Monitoring signals and validation window:** the CI `Run write-path-precondition gate` and
`... self-test` steps on the Ship PR and on `main` after the merge (the same job that runs the
retired-architecture gate). A red from any of them after merge triggers the rollback above.

**Rollback:**

* Each unit is a single-commit revert. Order: U5 and U3 before U2; U4 at any time; U1 may stay.
  The exception is a unit that used the blocked-path CI fallback (U2 or U3 only): it is two
  commits, so revert **both**, the fix first and then its `test:` commit.
* There is no data, config or schema state to restore.

**Operator checkpoints (override at staging-PR review, before Ship claims the shipment):**

1. **PA-2, the oracle chain.** The staging-PR approval is the written authorization for the edit to
   `writepath_oracle_test.go` (U1: the decode block `:194-203` and the header paragraph `:9-12`,
   nothing else) and for the later flip of the `decodeLsFilesListing` helper it calls (U2; the oracle
   file is unchanged in U2). The current header reads: "LIFECYCLE: the legacy side and its
   expectations are frozen. The only authorised edit is the single AC-E2.6 adaptation in Unit E ...
   Any other edit is an H-3 stop." The exact replacement is in U1.
2. **Stage-recommended options:** D-BW-1a (UTF-8 and control characters judged only for scanned
   paths; diverges from retiredarch's frozen behaviour) and D-BW-2 (no relative-`git` refusal).
3. **Recorded exceptions EX-1 to EX-5** (2-hour rule and P-004 labelling).
4. **The 5A8EC1BC evidence note** (an excluded entry; allowed by the operator: "note any new
   evidence only"). The note records evidence only; the trigger-extension suggestion stays in this
   plan (RK-9) for the operator to act on.
5. The Ship PR needs the operator's normal merge-commit approval (constitution XI).
6. **The deferred wrapper-root surface (stash `DBE25DF5`, medium, kind bug; the deferral stands
   under P-021 C1).** R4 isolates the `git ls-files` child **for a given root**. The wrapper's root
   discovery (`scripts/check-write-path-precondition.sh:150`, `ROOT="$(git rev-parse
   --show-toplevel)"`, unscrubbed) and ED-7, which detects only an *empty* selection
   (`writepath.go:850-855`), are not changed by this shipment, so ambient `GIT_DIR` or
   `GIT_WORK_TREE` can still point the gate at a wrong or partial tree and pass.
   **Override option:** if the operator wants this closed, direct Stage to stage it as a
   **separate, future task** (not a change to 042-S, whose six members stay as they are) that
   anchors `ROOT` to `GATECHECK_SRC` (already derived by the sourced
   `scripts/lib/gatecheck-run.sh:32`) or scrubs the `rev-parse` environment. It is a cross-engine
   change and must be coordinated with retiredarch's wrapper, which has the same unscrubbed
   discovery (`scripts/check-retired-architecture.sh:198`, `git -C "${GATECHECK_SRC}" rev-parse
   --show-toplevel`). A per-arm `internal/` plus `cmd/` non-empty guard is only a non-vacuity
   heuristic, not proof of root identity (plan-review AS-5), so it is deliberated there, not here.

**Review-gate capability risks:** plan review must emit literal `dispatch_mode:` and `decision:`
markers. If parallel subagent dispatch is unavailable, the fallback must be declared in the plan
before harvest (P-012); a silent single-reviewer fallback is not acceptable.

**Residual risks carried forward:**

* RK-3, RK-4 and RK-9 are accepted, with documented fallbacks.
* The relative-`git`-path refusal (`4372BAD4`), the wrapper root discovery and incomplete-listing
  detection (`DBE25DF5`), the other CI-log echo sinks (`EDC18D59`), and the retiredarch/writepath
  divergence on control-character scoping (`31F33EFE`) are explicitly not mitigated here; each is
  tracked as a stash entry. A terminal-escape rune such as C1 `U+009B` is outside the bounded set.
* B83F53BB and 05E12A6F are fully addressed by this shipment as scoped; the new residuals are the
  three captured entries.

## Plan Review

<!-- plan-review-attempt: 2 -->

dispatch_mode: multi-agent
decision: ADVISORY

Two review attempts ran (attempt 1: FAIL; attempt 2, the first re-entry cycle: ADVISORY). The
literal markers above are the **latest** verdict. Reviewers in both attempts were dispatched as
parallel subagents (`TOOL_OK: reviewer-subagent-dispatch`); the anchor route `gpt-6.1-sol` / openai
/ high was dispatchable, so no degradation applies.

**Persona coverage (attempt 2)**

| Persona | Mode | Result |
|---|---|---|
| Constitution Reviewer | subagent | no P0/P1; 3 P2, 5 P3 |
| Go Reviewer | subagent | no P0/P1; 1 P2, 7 P3 |
| Scope Boundary Auditor | subagent | no P0/P1; 1 P2, 9 P3 |
| Architecture Strategist | subagent, **anchor reviewer pass** (`gpt-6.1-sol`, openai, high) | no P0/P1; 1 P2, 1 P3 |
| Security Lens Reviewer | subagent | no P0/P1; 1 accepted-residual P2 (tracked as `DBE25DF5`), 4 P3 |
| Learnings Researcher | subagent | no P0/P1; 4 P2, 4 P3 |

**Gate decision (attempt 2): ADVISORY.** No P0 or P1 remains. Every P2 is resolved in the text of
this plan (table below); the P3 items are resolved or recorded as accepted. Under the Stage rule an
ADVISORY result proceeds when the operator confirms. **Authorization recorded:** the operator's
`stage next` instruction, relayed by the Orchestrator (batch B1 operator-confirmed; the expected
output is a queued shipment; Stage-recommended options are submitted for override at staging-PR
review), is the explicit authorization to proceed to harvest. The operator keeps the ability to
override at staging-PR review, before Ship claims the shipment. The in-place revisions below are
the reviewers' own recommendations and add no new design; no third attempt is needed for an
ADVISORY verdict.

Plan hardening was required (three signals) and is satisfied: the `## Plan Hardening` section
carries the strict-safety action record, verification, rollback and operator checkpoints.

### Attempt 2 resolutions

| ID | Sev | Finding | Resolution |
|---|---|---|---|
| CON-F1 / SCOPE-N1 | P2 | U4 exceeds the scenario bound with no exception; EX-3/EX-4 understate counts | Counting convention stated; EX-1/EX-3/EX-4 counts corrected with the function bound and Width Isolation; **EX-5** recorded for U4 in four-part form; U4 trimmed (control row and separate Windows sub-row dropped) |
| CON-F2 | P2 | The CI red-observation fallback would not observe a red; same-checkout rule; TMPDIR wording | Fallback defined: push the `test:` commit alone, record the failing run URL, then push the fix (consumes fix-ci budget); local routes must use the same checkout; Principle IV now a decision (no TMPDIR redirect into the checkout; `logs/` for captures). **Refined by the 2026-10-10 adversarial-review remediation (M-1):** U4 no longer uses a CI fallback; for U2/U3 the red run must finish before the fix is pushed and the unit's rollback reverts both commits |
| CON-F3 | P2 | Principle VIII mapping inconsistent with the grading | U1 now in careful mode with named pause points (PA-2 authorization, first real runner change); the U5 edit declared a freeze-scope deviation |
| GO-N1 / CON-F7 / SEC-N4 | P2 / P3 / P3 | e1 not hermetic; no no-vector control through the runner | Shared `hermeticGitEnv(t)` called first in (d) and in every e1 subtest; each vector is its own `t.Run`; a no-vector subtest asserts the runner bytes equal the baseline |
| ARCH-AS4 | P2 | The planned oracle decoder validates the whole listing, conflicting with D-BW-1a's tolerance of unscanned invalid-UTF-8 names | `decodeLsFilesListing` frames records only (no UTF-8 policy); the oracle applies `shouldScan` afterwards; a tolerance row is in test (a) |
| LRN-L2-1 | P2 | The docs "needle" is one local zero-count and self-hits | The matrix row is now a named-surface read; the supporting grep is a locator with stated expected hits, not closure evidence; a doc-contract row was added |
| LRN-L2-2 | P2 | "every fake returns through `lsFilesListing`" is defeated by test (b)'s raw literals | Reworded to "every well-formed-listing fake"; the malformed-listing rows of (b) are the only named raw literals |
| LRN-L2-3 | P2 | Literal green is not satisfiable on the Windows dev host; `-race` needs cgo | Environment prechecks: record the parent failing set, "no new failures versus the parent baseline", Git\bin PATH, Linux CI authoritative; `-race` and Windows lint marked platform-dependent |
| LRN-L2-4 | P2 | The oracle-edit procedure is not closed over U2 | PA-2 now covers the `decodeLsFilesListing` flip; the oracle file diff in U2 is empty (U2 AC4); the helper stays independent |
| SEC-N0 | P2 (accepted residual) | The wrapper can still scan the wrong tree or a partial listing | Tracked as `DBE25DF5`; closure wording pinned so no end-to-end claim is made (SEC-N1); AS-5 caution added to the entry |
| ARCH-AS5 | P3 | `DBE25DF5` overstates the per-arm guard | The entry now says it is a limited non-vacuity heuristic, not proof of root identity |
| GO-N2 | P3 | After-assertions in (d) and e1(a) are weak; fixtures must be valid Go | Both assert the finding text `write primitive 'os.Remove' found`; fixture bodies are full Go files; U4 uses `reject-*.go` with a primitive and `accept-*.go` with `u4CleanGo` |
| GO-N3 / SCOPE-N6 | P3 | The oracle snippet omits `var rels []string`; the header text says "now emits -z" before U2 lands | The snippet keeps `var rels []string`; the header wording is format-neutral and true at every commit |
| GO-N4 / SCOPE-N4 / LRN-L2-7 | P3 | `linkDir` and row 2 ambiguity; Windows classification unverified | Row 1 calls `u4CreateSymlink` directly; rows 2 and 3 use `linkDir`; the Windows `types_windows.go` trace is an optional note (a reviewer confirmed a mount point is `ModeIrregular`) with no assertion depending on it |
| GO-N5 | P3 | e1 mechanics left implicit | Each vector a `t.Run`; baseline captured with `cmd.Output()` (stdout only); the runner's pathspec order |
| GO-N6 | P3 | Bare `errors.New` strings in the helper | Two unexported sentinels (`errListingNotTerminated`, `errListingEmptyRecord`); test (a) asserts `errors.Is` |
| GO-N7 / SCOPE-N10 | P3 | Verification command portability | `-race` and Windows lint marked platform-dependent; `GOOS=linux go vet` added for Windows hosts; PowerShell note |
| GO-N8 / SEC-N3 | P3 | U4's refusal echoes the raw fixture name | Noted as another instance on `EDC18D59` (added to the entry); not switched to `%q` because the row assertions pin the raw text |
| SCOPE-N2 / N3 | P3 | U4 residual scope; AC4 relies on CI `-v` | Row 4 and the separate sub-row dropped (goldens cover regular fixtures); R6 narrowed to "linked"; AC4 requires a local `-v` run |
| SCOPE-N5 | P3 | Matrix check quality; doc-contract row missing | Matrix rewritten (producer check reads the argv; doc-contract row added; U2 AC5 covers it) |
| SCOPE-N7 / LRN-L2-5 | P3 | Wording drift: "plus the SplitLines boundaries"; control rows location; "one-line swap" | Rune set stated as identical; control rows live in (b); EX-1 wording corrected; U5 AC2 states "exactly once" |
| SCOPE-N8 | P3 | The 5A8EC1BC note may carry recommendations | The note is evidence only; the trigger-extension suggestion stays in RK-9 |
| SCOPE-N9 | P3 | New stash entries lack a duplicate-scan line | Each of `4372BAD4`, `DBE25DF5`, `EDC18D59` now carries a DISCOVERY / DUPLICATE SCAN: CLEAN line |
| CON-F4 / F5 / F6 / F8 | P3 | Function counts; strict-safety residuals; matrix self-grep; overlays | Counts corrected; PA-1/PA-3 pre-approved by the staging PR, PA-2 recorded `approved` before U1, PA-4 `moderate` justified; matrix reworded; overlays row added |
| SEC-N1 | P3 | The Ship closure wording is not pinned | Closure hygiene now states the isolation boundary and the residual stash IDs |
| SEC-N2 | P3 | The rune set blocks workflow-command injection but not terminal escapes (C1 `U+009B`) | Recorded as a residual; any extension routes to `31F33EFE` and `EDC18D59` |
| SEC-N5 | P3 | Nothing compares production selection with an independent decode on the live tree | Declined: the oracle and goldens cover scanning, test (d) covers the under-selection class with real git, and a fourth test would add a scenario to an over-budget unit |
| LRN-L2-6 | P3 | Premises rest on one-time staging measurements | U2 AC6: re-measure the counts and `git --version` at the Ship base |
| LRN-L2-8 | P3 | Harvest and closure hygiene | Harvest notes (`backlogit --version`, PATH), `logs/` for captures, `stash.jsonl` carry-forward and the extra Copilot compound cite added |

### Adversarial review remediation (2026-10-10)

A multi-model adversarial review of the staged package returned READY_WITH_FOLLOWUPS (no P0/P1).
Because Ship runs unattended, the follow-ups were folded in as text-level clarifications only: no
design change, no change to shipment 042-S membership, task IDs, titles or dependency edges.

| ID | Sev | Disposition | Where |
|---|---|---|---|
| M-1 | P2 | Applied. U4's red evidence is rows 2 and 3 (junction-based, unprivileged Windows); row 1's red from WSL/Developer Mode or disclosed; row 1 SKIP allowed on unprivileged Windows if recorded; no standalone `test:` push for U4; remaining U2/U3 fallback waits for the red run and reverts both commits | U4 AC1, AC4, posture paragraph; Blocked-path handling; both Rollback sections; Runtime Verification U4 row; RK-4; task 052.004-T; Rollback lines of 052.002-T, 052.003-T |
| M-2 | P2 | Applied. `DBE25DF5` added as operator checkpoint 6 with the separate-future-task override; stays deferred | Operator checkpoints |
| O-1 | P2 | Applied to stash entries `4372BAD4`, `DBE25DF5`, `EDC18D59` (explicit task/feature/shipment ID `N/A` in source refs) | `.backlogit/stash.jsonl` |
| P-1 | P3 | Applied. COUNT/KEY_n coverage is the generic `GIT_` prefix rule, not independently observed; no table row added | U3 "Why three vectors"; deliberation D-BW-2 |
| L-1 | P3 | Applied. Batch A precedent (its EX-2; 051.001-T, 051.006-T) cited | EX-2; tasks 052.001-T, 052.005-T |
| L-2 | info | Not applied: an empty-linked-ancestor row is green at the parent (the loop never runs, `writepath.go:919-921`), so it is characterization only and would add a sixth sub-case to a unit already past its bound (EX-5) | none |
| L-3 | P3 | Applied. The compile-stub also declares both sentinels | U2 Posture step 1; task 052.002-T AC2 |
| L-4 | P3 | Applied. Invalid-UTF-8 error wrapped with the `%q` path | U2 helper; deliberation D-BW-1; task 052.002-T |
| L-6 | P3 | Applied. EX-3 count 10; "same set" wording; e1 pathspec order | EX-3; U3 e1; deliberation D-BW-1a |
| L-7 | P3 | Applied. Ship adds a cross-referencing comment on `scannedPathsFromListing` (`31F33EFE`) | U2 helper notes; task 052.002-T |
| L-5, L-8 | P3 | No change (L-5 stays as tracked by `EDC18D59`; L-8 was a false positive) | none |

### Attempt 1 (superseded by attempt 2)

<!-- plan-review-attempt: 1 -->

*Attempt 1 (2026-10-10).* Attempt-1 gate: **FAIL** (not the latest verdict). Three P1 findings
(CON-F1, ARCH-AS1, LRN-L1), none of P0. The findings and their resolutions in the first revision:

| ID | Sev | Finding | Resolution in this revision |
|---|---|---|---|
| CON-F1 / SCOPE-F1 | P1 / P2 | U2 and U3 exceed the 2-hour bounds with no recorded exception; EX-1 sat on the wrong unit | EX-1 corrected; EX-3 (U2) and EX-4 (U3) recorded in four-part form; U3 cut from five vectors to three; the Width Isolation point is covered in EX-3 |
| ARCH-AS1 / SEC-1 | P1 / P2 | The wrapper's root discovery (`git rev-parse --show-toplevel`) is unscrubbed, so ambient `GIT_DIR` can redirect the scan root | R4 narrowed to a fixed root; the boundary stated in Problem Frame 2; captured as a new DEFERRED SCOPE EXPANSION stash entry together with per-arm incomplete-listing detection (SEC-4) |
| LRN-L1 | P1 | The "finds nothing" `rg` AC matches the plan's own legacy-newline row and is unsatisfiable | U1 AC1 and U2 AC5 are now structural (read the diff; the named `legacyNewlineListing` row is the only newline literal); a contract-surface matrix replaces the grep |
| ARCH-AS2 | P2 | Verification commands omit the mandatory `--root` | Every direct `gatecheck` invocation carries `--root .`; the wrapper is added as an end-to-end check |
| CON-F2 / LRN-L6 | P2 / P3 | (a) and `TestGitRunnerEnv` are compile-reds only | Both unit postures stage a compiling stub, so every red is behavioural |
| CON-F3 | P2 | strict-safety grading under-calibrated; `targets`/`change_kind` missing; no ActionResult record | PA-1, PA-2, PA-3 re-graded high; columns added; Ship records the final ActionResult in the PR body and closure record |
| CON-F4 / GO-G9 | P2 / P3 | Quality-gate scope narrower than the constitution's | Repo-wide pre-PR set in constitution order, `-count=1`, `-race`, `GOOS=windows` vet, goimports and golangci-lint added |
| GO-G1 / LRN-L7 | P2 / P3 | Vectors (b)-(e) can be fake-green if compared by exit code | Oracle is the listing bytes equal to a scrubbed baseline; the baseline is asserted non-vacuous; (a) also asserts the stderr finding |
| GO-G2 | P2 | INV-D error identity not pinned at helper level | Test (a) gains helper-level `errors.Is(err, pysem.ErrInvalidUTF8)` rows |
| SCOPE-F2 | P2 | R5 and the per-vector liveness apparatus port excluded entry C8827920 | Apparatus dropped; vectors cut to three; R5 reworded; C8827920 stays with the retiredarch table |
| SEC-2 / SEC-3 | P2 / P2 | Other CI-log sinks (fixture names, `errorLine` tail, absolute path in the `ReadText` error) are unescaped | D-BW-1a claim re-scoped to the listing arm; captured as a new DEFERRED SCOPE EXPANSION stash entry |
| SEC-4 | P2 | A non-empty but incomplete listing is undetected | Folded into the captured wrapper-root entry (per-arm non-empty guard) |
| LRN-L2 | P2 | No contract-surface matrix | Added (see "Contract-surface matrix"), with the commands to re-run at U2 |
| LRN-L3 | P2 | Reconciled oracle header wording not written out | Exact replacement paragraph is in U1, with an AC bounding the diff |
| LRN-L4 | P2 | Harvest and authoring hazards not carried | "Harvest notes" added (sizing prose, ASCII, read back, control-character scan) |
| SCOPE-F3 | P3 | RK-8 wording off; whole-listing UTF-8 check inconsistent with D-BW-1a | UTF-8 validity is judged per scanned record; RK-8 reworded; the deliberation's D-BW-1 failure modes updated |
| SCOPE-F5, SCOPE-F6, CON-F11 | P3 | Evidence note on an excluded entry; U1 rationale; EmptySelection swap; count | The operator allowed the note (checkpoint 4); the no-op edit dropped; counts corrected to four existing functions |
| SCOPE-F8, CON-F7 | P3 | Rows 2 and 4 weak; regression guards presented as verification | Row 2 labelled message-only; regression guards labelled in every AC |
| SCOPE-F10, CON-F12 | P3 | Citation drift; premature status | Citations corrected (`containedRegularFile` `:786-823`; parse `:835-848`; ED-7 `:850-855`; oracle `:188-212`; comment `:25-31`); `status` held at `hardened` |
| CON-F5 | P3 | PA-2 approval deemed, not explicit; exceptions not shown | Operator checkpoints now quote the oracle header and list EX-1..EX-4 |
| CON-F6 / GO-G4 | P3 | Test (d) not hermetic at U2's commit | (d) strips ambient `GIT_*` and pins config; the `core.quotepath` origin is recorded |
| CON-F8 | P3 | Principle IV rests on a convention | Interpretation recorded in the Constitution Check |
| CON-F9 | P3 | A red test-only commit in CI history | Prefer a symlink-capable local run; if CI, a `test:` commit immediately followed by the fix, disclosed |
| CON-F10, LRN-L10 | P3 | Closure omits P-009, P-020, P-018 and the gate toggle | Added to Runtime Verification and Closure |
| GO-G3 | P3 | Ambient-key removal detail | Moot: vector (e) dropped; the same idiom is used for the hermetic setup in (d) and checks the `os.Unsetenv` error |
| GO-G5 | P3 | Swapped link-helper arguments; linked `scripts` and `scripts/testdata` untested | `linkDir` wrapper; row 3 iterates three linked ancestors |
| GO-G6 | P3 | Rune-rule boundaries untested | DEL, U+2029 and `\x1f` rows and an all-unscanned ED-7 row added; `\u00e9` written escaped |
| GO-G7, GO-G8 | P3 | Import and Go 1.24 floor details | Noted in the unit preamble and U2/U3 |
| GO-G10 | P3 | Injected `safe.directory` on the host | Added to RK-6 |
| GO-G11, ARCH-AS3 | P3 | Hand-synced copies unpinned | Pure env table identical to retiredarch's; RK-9 and the 5A8EC1BC note propose extending its trigger |
| GO-G12 | P3 | `selectedRepoPaths` vs `selectRepoPaths` near-collision | Renamed `scannedPathsFromListing` |
| SEC-5 | P3 | Oracle-edit authority is self-referential and advisory | The helper's doc comment treats changes to it as oracle edits; the header cites D-BW-4 and the staging PR; the enforcement is process-only (R-A2b class) |
| SEC-6 | P3 | TOCTOU not stated; leaf-junction not asserted | TOCTOU recorded as accepted; a Windows sub-row asserts "never read through" |
| SEC-7 | P3 | Env-isolation boundary under-documented | Threat-model paragraph added to the `DefaultGitRunner` doc comment (U3) |
| SEC-8, LRN-L8 | P3 | Rune set covers injection, not spoofing; denylist bounding | Bounded in U2 and D-BW-1a: any extension routes to `31F33EFE` and the output-sink entry |
| SEC-9 | P3 | Test-hygiene risks | U3 hygiene bullets (fatal on missing git, no environment dump, extra env applied last, no `t.Parallel`) |
| LRN-L5 | P3 | Junction behaviour determinable from GOROOT | U4 requires Ship to trace `types_windows.go` and cite 051.004-T's leaf-junction test |
| LRN-L9 | P3 | U5 must not contain a pin anchor verbatim | U5 AC2 added |

Attempt 2 re-reviewed this revision; its verdict is at the top of this section.

---
title: "Gatecheck Batch A correctness: NUL-safe retiredarch selection, empty-selection guard, cross-engine symlink containment"
description: "Stage deliberation over Batch A (stash 4537B2F6, 0D643BE8, D44D8BDF): premise verification against main@135ca59, option analysis, and decisions D-BA-1..D-BA-5"
topic: "Batch A: retiredarch git ls-files quoting false-clean (4537B2F6), repo-mode empty-selection exit 0 (0D643BE8), cross-engine symlink containment remainder (D44D8BDF)"
depth: "standard"
decision_status: "decided"
promoted_to: "plan"
linked_artifacts:
  - "docs/plans/2026-10-08-intercom-go-gatecheck-batch-a-correctness-plan.md"
  - "docs/decisions/2026-10-07-intercom-go-retiredarch-gate-integrity-deliberation.md"
  - "docs/compound/2026-10-07-pin-canonical-text-maintenance-and-unbounded-ast-rule-review.md"
  - "docs/compound/workflow-issues/staging-premise-verification-before-planning-2026-09-18.md"
tags:
  - "gatecheck"
  - "retiredarch"
  - "writepath"
  - "unignore"
  - "p-021"
  - "stage"
---

# Deliberation: Gatecheck Batch A correctness

## Session context

* **Agent and mode:** Stage, invoked by the Orchestrator (`stage next`). Sequential, non-dark.
  The route is `claude-opus-5.5` / anthropic / high (P-013.5).
* **Operator instructions,** relayed by the Orchestrator at 2026-10-08T14:12-07:00:
  * "stage next on Batch A". Batch A is exactly 4537B2F6, 0D643BE8 and D44D8BDF.
  * "Carry forward changes to stash.jsonl and .gitignore; include in next Stage commit."
* **Operator approval scope.** Stage applies the relayed instruction to two things:
  * the batch selection (the operator selected the grouping);
  * the Stage-recommended option in each decision below.

  Any override can be made at staging-PR review, before Ship claims the shipment.
* **Base:** `main` @ `135ca59`.
  * Stage artifact branch: `chore/stage-gatecheck-batch-a-correctness`.
  * Slug source: this covering-feature title, "Gatecheck Batch A correctness".
* **Explicitly excluded** (operator):
  * DC921AF6, 3750C37C and 0ECC1895. These are held by D-RA-7 for a bounded-threat-model
    deliberation.
  * Every writepath-harness entry, every Root.Resolve detector entry, and every other stash
    entry.

## Step 1: Triage and P-021 obligations

| Stash ID | Kind / priority | Shape | DEFERRED SCOPE EXPANSION marker | Route |
|---|---|---|---|---|
| `4537B2F6` | bug / high | task | no (Stage-captured) | deliberate (entry says "light") |
| `0D643BE8` | task / medium | task | **yes** | deliberate, forced by P-021 C6 |
| `D44D8BDF` | task / low | task | no (Stage-captured) | deliberate (entry says "yes") |

### (A) Unconditional duplicate scan (0D643BE8, extended to all three)

We scanned the full active stash for these terms: empty selection, zero paths, non-empty,
symlink, junction, Lstat, quotePath, ls-files, gitRunnerEnv, D7BF9F74, env isolation.

* **Matches:** 5A8EC1BC (consolidating three GitRunner shapes; deferred until a fourth engine
  exists) and C0D28448 (write-primitive denylist that names `os.Symlink`).
* **Verdict:** neither describes the same expansion as any Batch A entry. 5A8EC1BC is a
  related but different concern (abstraction consolidation). **Duplicate scan: CLEAN** for
  all three entries.

### (B) Late-identifier reconciliation (0D643BE8; its PR and review-thread fields were `N/A`)

* **Searched:** the Ship-owned residual-risk records that cite `0D643BE8`.
* **Recovered PR #112.** Records:
  * `docs/closure/040-S-050-F-post-merge-closure.md`, condition `deferred-scope-0d643be8`:
    "PR #112 Local Review Readiness block (follow-ups)";
  * `docs/memory/compacted/2026-10-08-040-s-050-f-compacted.md`, section "Review";
  * `docs/archive/memory/2026-10-07/040-S-red-evidence.md`, line 197.

  The finding came from the pre-PR local adversarial review of shipment 040-S. PR #112 is
  the PR that carries it as a residual. The capture commit is `3f2e71c`.
* **Review-thread ID: no late identifier found.** The finding came from a local adversarial
  review, not a GitHub thread, and no record cites a thread. The `N/A` stands as a truthful
  terminal record.
* **Other refs carried forward:** task 050.003-T, feature 050-F, shipment 040-S.
* **Write half:** the in-place stash annotation was deferred until after the Step 1.9 gate.
  It is applied as this entry's archival disposition (see "Dispositions").

4537B2F6 and D44D8BDF are Stage-captured, not Ship C2 captures. Their source refs are already
complete: D-RA-7, PR #83, review thread `PRRT_kwDOTPuhps6nd8rH`. No reconciliation is needed.

## Step 1.8: Learnings consulted

* `docs/compound/2026-10-07-pin-canonical-text-maintenance-and-unbounded-ast-rule-review.md`
  (high relevance):
  * refreeze `select.go` and `pin.go` atomically in one commit (ALP);
  * stage the freeze to get a real red;
  * avoid fake reds;
  * `writeMutatedCopy` anchors must stay unique;
  * `"import"` is in both frozen sets.
* `docs/compound/workflow-issues/staging-premise-verification-before-planning-2026-09-18.md`
  (high relevance):
  * re-read each premise in its enclosing function;
  * label green-on-arrival criteria as characterization;
  * reconcile freeze-vs-change in the plan.

## Premise verification (against `main` @ `135ca59`, direct reads)

### 4537B2F6: the premise HOLDS

* `select.go:40` (`DefaultGitRunner`):
  `args := append([]string{"ls-files", "--"}, pathspecs...)`. There is no `-z`.
* `select.go:196-217` (`selectRepoPaths`; the parse is at `:205-214`): `pysem.GitText(out)` and then
  `pysem.SplitLines(listing)`. Neither does C-style unquoting.
* `shouldScanRepoPath` matches on the raw `"`-prefixed quoted form, so the file is dropped.
* **Executable evidence on main.** `select_test.go:404-455`, `TestDefaultGitRunnerIgnoresGitEnv`:
  * the baseline asserts that the selection is **only** `cmd/x/main.go`;
  * its comment says `internal/\u00e9.go` "is printed quoted by git and is therefore not
    selected. When 4537B2F6 lands it must update this baseline and replace vectors (c)-(e)".

  So the false-clean is pinned by a passing test today.
* **A second hazard the entry does not name.** `pysem.SplitLines` splits on Python line
  boundaries, including U+0085, U+2028, U+2029 and U+001C..U+001E. A path containing U+2028
  is printed raw under `core.quotePath=false`. Under the default quoting it is printed as
  octal-quoted multibyte. Either way it can never round-trip into a scannable path. Only a
  NUL-delimited listing closes every variant.
* **Coupled surfaces:**
  * `selftest_selection.go:79-102`: `expectedInternalRepoPaths` makes its own call through
    the same runner and parses with `SplitLines`.
  * `pin_test.go:245`: a `writeMutatedCopy` anchor on the exact `DefaultGitRunner` argv text.
  * `pin.go:81-190`: `canonicalDecls` holds both `DefaultGitRunner` and `selectRepoPaths`.
  * Two test fakes emit newline-separated listings: `select_test.go:121`
    (`TestSelectRepoPaths_FiltersAndSorts`) and `retiredarch_test.go:585` (`u3StubGit`).

### 0D643BE8: the premise HOLDS

* `retiredarch.go`, `runRepoScan`:
  * `selectRepoPaths` errors return Code 1;
  * otherwise the function loops over `relPaths`, and with no findings it returns
    `Result{Code: 0}`;
  * an empty `relPaths` therefore returns Code 0.
* The only guard is the `selection non-empty` assertion in `selftest_selection.go:214-219`.
  It runs in `--self-test` and `--self-test-integrity`, never in repo mode.
* **Precedent in a sibling engine.** `writepath.go:791-796` (ED-7) already fails closed on an
  empty selection: `"::error::write-path repo scan matched no files under internal/** or cmd/** (ED-7)"`,
  Code 1.
* **Contract check:**
  * No doc or wrapper asserts that an empty selection passes.
  * `scripts/check-retired-architecture.sh` passes the exit code through unchanged.
  * The repo-mode tests in `tools/gatecheck/main_test.go:331` and
    `retiredarch_test.go:158` run against the real, non-empty repository.

### D44D8BDF: the premise HOLDS ONLY PARTIALLY. It is narrowed, with evidence

| Engine | Premise "joins git-selected paths lexically and reads through a tracked link" | Evidence |
|---|---|---|
| retiredarch | **Already satisfied.** PR #112 (040-S U3) added `containedRegularFile` | `retiredarch.go:604` (call), `:716` (Lstat walk) |
| writepath | **Holds** | `writepath.go:798-805` (`runRepoScan` loop) calls `scanFile(root, rel)`, then `pysem.ReadText(filepath.Join(root, filepath.FromSlash(relPath)))` at `:459`. There is no Lstat. |
| unignore | **Holds narrowly.** It does not read git-selected paths. It reads one fixed path, the working-tree root `.gitignore`, when `ref == "HEAD"` | `unignore/git.go:91-92` uses `os.ReadFile(filepath.Join(repoDir, ".gitignore"))`. A tracked symlinked `.gitignore` is followed, but git itself (2.32 and later) does not follow an in-tree `.gitignore` symlink, so the gate's baseline text would diverge from git's effective rules. |
| mergestrategy | **Refuted.** It never reads git-selected repo paths | `mergestrategy/evaluate.go:9-17` documents this as an accepted, scoped Principle III exception. Run discards root and reads a caller-supplied payload path, which must work for temp payloads outside the repo. |

There are still no tracked symlinks: `git ls-files -s` shows no mode-120000 entries, so
urgency stays low.

### Newly discovered out-of-scope expansion (captured, not planned)

`writepath.DefaultGitRunner` (`writepath.go:144-158`) runs
`git ls-files -- internal/** cmd/**`. It has two gaps:

* **No `-z`.** It parses with `GitText` + `SplitLines`, which is the same quoted-path
  false-clean class as 4537B2F6, in the merge-blocking write-path gate.
* **No D7BF9F74 environment isolation.** There is no `cmd.Env` scrubbing, so ambient GIT_*
  variables and global/system config apply.

None of this is in Batch A, and the operator forbade expansion. It is captured as a new
**DEFERRED SCOPE EXPANSION** stash entry (see "Dispositions").

## Problem frame

* **The core problem.** The retired-architecture gate can pass ("clean") without having
  scanned files it is meant to scan. This happens in two ways:
  * a tracked path that git prints quoted, or that `SplitLines` splits, is silently dropped
    (4537B2F6);
  * a degenerate empty selection exits 0 in repo mode (0D643BE8).
* **The secondary problem.** The writepath and unignore gates read through a tracked
  symlink, so the bytes they read can differ from the tracked blob (D44D8BDF remainder).
* **Success criteria:**
  * every scoped tracked path, whatever its bytes, is either scanned or fails the gate closed;
  * an empty repo-mode selection exits non-zero;
  * writepath and unignore refuse non-regular or linked inputs, with the same semantics as
    retiredarch U3;
  * every new test is red at its parent;
  * the `select.go` + `pin.go` freeze stays atomic.
* **Out of scope:**
  * the writepath `-z` and env isolation gap (captured as a new entry);
  * GitRunner consolidation (5A8EC1BC);
  * the pin trust-root entries (DC921AF6, 3750C37C, 0ECC1895);
  * mergestrategy, whose premise is refuted.

## Decisions

### D-BA-1, `4537B2F6`: a NUL-delimited listing (`-z`), parsed inline in the frozen `selectRepoPaths`

| Option | Description | Pros | Cons | Effort |
|---|---|---|---|---|
| **A: `-z` + inline NUL split** (recommended) | `DefaultGitRunner` passes `ls-files -z --`. `selectRepoPaths` validates UTF-8, requires a NUL terminator, rejects empty records, splits on NUL, and does no newline translation. `expectedInternalRepoPaths` gets its own independent NUL parse. | Closes every quoting and splitting variant, including U+2028 and control characters. The parse stays inside the frozen, pinned surface. | Changes the GitRunner output contract, so test fakes must follow. It needs an atomic `select.go` + `pin.go` refreeze. | medium |
| B: `-c core.quotePath=false` + fail closed on any `"`-prefixed line | Non-ASCII paths print raw. Remaining quoted forms fail the gate. | No change to the fake contract. | The U+2028/U+0085 `SplitLines` hazard remains, and files whose names need quoting are refused instead of scanned. It is still a refreeze. | low-medium |
| C: fail closed on any quoted line only | The scope is minimal. | It is very small. | Every non-ASCII path fails the gate permanently. The `SplitLines` hazard remains. | low |
| D: C-unquote quoted lines | Implement git's C-style unquoting. | No change to the runner contract. | It adds a new parser with its own parity risk. The `SplitLines` hazard remains. | medium-high |

**Decision: Option A.** Details:

* **Parse location.** The parse stays inline in `selectRepoPaths`. A helper outside
  `select.go` would move selection semantics out of the pinned surface (R-A1).
* **Failure modes:**
  * `pysem.ErrInvalidUTF8` is kept for invalid UTF-8, so the existing error identity survives;
  * a non-empty listing that does not end in NUL is an error. This catches a runner that
    ignores `-z`, or a stale newline fake that would otherwise false-clean (`"cmd/x.go\n"`
    fails `HasSuffix(".go")`);
  * an empty record is an error;
  * an empty listing returns `nil, nil`. D-BA-2 then fails it closed at the scan level.
* **Independence.** `expectedInternalRepoPaths` keeps its *independent* re-derivation (H-11
  spirit) with its own NUL parse. It does not call or share the `selectRepoPaths` parser.
* **TestDefaultGitRunnerIgnoresGitEnv.**
  * The baseline becomes `{cmd/x/main.go, internal/\u00e9.go}`.
  * The quotePath vectors (c)-(e) stop distinguishing once `-z` lands. They are replaced by
    configuration-channel vectors that make an *unisolated* git fail:
    * `GIT_CONFIG_COUNT=1` with no `GIT_CONFIG_KEY_0`;
    * a malformed `GIT_CONFIG_PARAMETERS`;
    * `GIT_CONFIG_GLOBAL` pointing at a syntactically invalid file.
  * Each vector carries an in-test **control**: an unisolated test-local runner must
    observably differ under the same vector. This proves the vector is live and not
    green-on-arrival.

### D-BA-1a, plan-review amendment (security lens SEC-2): fail closed on control and line-separator characters in selected paths

* **Problem.** Under `-z`, git emits path bytes raw. A tracked path that contains `\n`,
  `\r`, ESC, U+0085, U+2028 or U+2029 would reach CI stderr unescaped through
  `scanPath` findings, which could inject GitHub Actions workflow commands.
* **Decision.** Inside the same frozen `selectRepoPaths` parse (no new top-level
  declaration), any path containing a rune below 0x20, or 0x7f, 0x85, 0x2028 or 0x2029,
  fails the whole selection closed. The error message prints the path with `%q`.
  * Such paths are therefore **refused (gate red)**, not scanned. This is fail-closed, and
    no legitimate repository path needs these characters.
  * It replaces the earlier claim that U+2028 paths would be "selected intact".
* **New pin row.** A `select_repo_paths_newline_split` reject row is added to guard the
  newly pinned NUL split itself.

### D-BA-2, `0D643BE8`: a fail-closed empty-selection guard in `runRepoScan`, not in the frozen `selectRepoPaths`

| Option | Pros | Cons |
|---|---|---|
| **A: guard in `runRepoScan`** (recommended): `len(relPaths) == 0` returns Code 1 with an `::error::` line in writepath ED-7 style | `runRepoScan` already decides the exit code, and it is unfrozen, so there is no pin churn. It covers the `""` and `--self-test` modes. It mirrors writepath ED-7. | Repo mode and `--self-test` now both report emptiness. That is harmless: self-test fails earlier on its own assertion. |
| B: return an error from `selectRepoPaths` on an empty selection | One place. | It is frozen, so it needs a refreeze. It changes self-test semantics: `runRepoSelectionSelfTest` would error before reporting its named assertions. |
| C: status quo (rely on `--self-test-integrity` in CI) | No change. | Repo mode stays fail-open, which is the finding. |

**Decision: Option A.**

* **Message:**
  `::error::retired-arch repo scan selected no files under config.toml.example, cmd/** or internal/** (fail-closed)`.
  Ship may adjust the wording, but it must contain `::error::` and must not be empty.
* **Exit-code contract change:** an empty selection moves from 0 to 1. This is the
  hardening the entry asked for. No consumer relies on the old 0.

### D-BA-3, `D44D8BDF`: narrowed per-engine containment for writepath and unignore. Mergestrategy is excluded, as refuted

| Option | Pros | Cons |
|---|---|---|
| **A: per-engine** (recommended) | For writepath, a private `containedRegularFile` copy that mirrors retiredarch U3, called in `runRepoScan` before `scanFile`. For unignore, a single `os.Lstat` of the root `.gitignore` in `rootGitignoreTextAt(HEAD)`. The blast radius is minimal, and the writepath `scanFile` mask-flow shape is untouched. It is consistent with 5A8EC1BC's deferral of consolidation. | Duplicates the ~30-line Lstat walk between retiredarch and writepath. |
| B: a shared `internal/contain` package, used by retiredarch, writepath and unignore | One implementation. | It re-touches freshly shipped 040-S code and changes retiredarch's imports. It pre-empts the 5A8EC1BC consolidation decision. It is a larger review surface. |
| C: keep D44D8BDF active and defer it | No work. | The operator selected it. The premise holds for writepath, and the fix is cheap. |

**Decision: Option A.** Constraints:

* **writepath:**
  * the containment check goes in the `runRepoScan` loop and must NOT go inside `scanFile`.
    `retiredarch/writepath_mask_test.go` pins the body of `scanFile`
    (ReadText, then scanSource, then a single mask);
  * on failure it aborts with `errorLine(rel, …)` and Code 1, which matches writepath's
    existing ED-2 abort-on-read-error semantics;
  * a non-existent final component falls through to `scanFile`, so the existing read-error
    text is unchanged;
  * no `filepath.Abs` or `filepath.EvalSymlinks` anywhere reachable from registered runners
    (`retiredarch/initstate_test.go`).
* **unignore:**
  * a missing `.gitignore` keeps returning `("", nil)`;
  * a `.gitignore` that exists and is not a regular file returns an error, and the caller
    already fails closed.
* **mergestrategy:** no change. Its non-containment is a documented, accepted exception.

### D-BA-4: sequencing and atomicity

* **First, a test-only prep task (characterization).** It introduces one shared test helper
  that renders a listing in the runner's current output format. Both newline fakes switch to
  that helper. This means:
  * the `-z` task only flips the helper;
  * the empty-selection task's "only out-of-scope paths" fake is format-agnostic.
* **The `-z` task is ONE executable task and ONE commit** (ALP). It covers `select.go`,
  `pin.go`, `pin_test.go` (anchor plus a new "drops -z" reject row), `selftest_selection.go`
  and `select_test.go`. Ship commits per task, and any split would leave a red or un-pinned
  intermediate commit.
* **Independence.** The empty-selection guard, writepath containment and unignore `.gitignore`
  tasks are independent of the `-z` task, apart from their shared dependence on the prep
  helper.

### D-BA-5: the shape of the release unit

* **One covering feature:** "Gatecheck Batch A correctness".
* **Six tasks, directly under the feature.** This follows the 050-F precedent; there are no
  sub-epics. Plan-review split out U6, the env-vector replacement, as a test-only task
  after U2.
* **One queued shipment.** Harvest outcome: feature 051-F, tasks 051.001-T to 051.006-T,
  shipment 041-S (queued).

## Dispositions (P-021: Stage alone triages)

| Stash ID | Disposition |
|---|---|
| `4537B2F6` | **Consumed.** Harvested into the `-z` task (051.002-T), plus the prep task (051.001-T) and the env-vector task (051.006-T). Archived after shipment assembly. |
| `0D643BE8` | **Consumed.** Harvested into the empty-selection guard task (051.003-T). Archived after shipment assembly. Reconciliation outcome: PR #112 recovered; review thread N/A stands; duplicate scan clean. |
| `D44D8BDF` | **Consumed, narrowed.** Harvested into the writepath containment (051.004-T) and unignore `.gitignore` (051.005-T) tasks. Mergestrategy is dropped as refuted, and retiredarch is recorded as already satisfied by PR #112. Archived after shipment assembly. |
| `B83F53BB` *(new, high)* | **Captured.** A DEFERRED SCOPE EXPANSION for the writepath `DefaultGitRunner` gaps: no `-z` (the same quoted-path false-clean as 4537B2F6) and no D7BF9F74 environment isolation. Left active for a future Stage session. |
| `05E12A6F` *(new, low)* | **Captured.** A DEFERRED SCOPE EXPANSION for the writepath fixture self-test (`runFixtureSelfTest`, `writepath.go:833-880`). It reads `os.ReadDir`-enumerated fixtures through `scanFile` with no `Lstat` containment. This is not a git-selected path, so it falls outside D44D8BDF's premise. Left active. |

## Trade-off comparison (summary)

| Criterion | D-BA-1 A | D-BA-1 B | D-BA-2 A | D-BA-2 B | D-BA-3 A | D-BA-3 B |
|---|---|---|---|---|---|---|
| Closes the finding completely | yes | partial | yes | yes | yes (narrowed) | yes |
| Pin churn | 1 atomic refreeze | 1 refreeze | none | refreeze | none | none (retiredarch.go touched) |
| Blast radius | retiredarch only | retiredarch only | `runRepoScan` | selection + self-test | 2 call sites | 3 engines + new package |

## Rejected alternatives

* **D-BA-1 B, C and D.** Each leaves the `SplitLines` Unicode-boundary hazard, and C
  permanently refuses legitimate non-ASCII paths.
* **D-BA-2 B.** It causes pin churn and reorders the self-test failure reporting.
* **D-BA-3 B.** It re-touches 040-S code and pre-empts 5A8EC1BC.
* **D-BA-3 C.** The operator selected D44D8BDF, and the premise holds for writepath.

## Unresolved questions

* **Git behaviour for the D-BA-1 replacement vectors.** We expect the unisolated git to exit
  non-zero under each vector, from git 2.31 onward (`GIT_CONFIG_COUNT`). Ship must confirm
  this with the in-test control before relying on a vector. If a vector does not distinguish,
  Ship drops it, keeps the others, and records the substitution.
* **Symlink creation on Windows hosts.** The writepath and unignore tests need it. As with
  040-S U3, a Windows host without the privilege skips the file-symlink case. Linux CI
  (`ubuntu-latest`) runs it. For writepath, a junction variant on an intermediate directory
  runs on Windows.

## Risks and mitigations

* **Risk: a stale newline fake silently false-cleans after `-z`.**
  Mitigation: the NUL-terminator requirement turns it into an error, and the prep helper
  centralises the format.
* **Risk: a broken pin mid-change.**
  Mitigation: one atomic commit, verified with
  `git show --stat` = {`select.go`, `pin.go`, `pin_test.go`, `selftest_selection.go`, `select_test.go`}.
* **Risk: the containment check in writepath breaks the mask-flow pin.**
  Mitigation: place it in `runRepoScan`, not `scanFile`, and run the
  `TestWritePath*` mask tests in the retiredarch package.
* **Risk: changing the exit-code contract breaks a consumer.**
  Mitigation: the grep in "Premise verification" found no consumer, and every repo-mode test
  runs on the real, non-empty tree.

## Operator checkpoint

These decisions apply the operator's relayed Batch A instruction. Before Ship claims the
shipment, the operator can override any decision during staging-PR review. Two decisions in
particular may deserve a look:

* D-BA-3, the narrowed inclusion of D44D8BDF;
* D-BA-2, the exit-code change.

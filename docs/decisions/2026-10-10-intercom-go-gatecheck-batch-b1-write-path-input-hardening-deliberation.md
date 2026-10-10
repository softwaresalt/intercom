---
title: "Gatecheck Batch B1: write-path input hardening (NUL-safe selection, git environment isolation, fixture containment)"
description: "Stage deliberation over Batch B1 (stash B83F53BB, 05E12A6F): premise verification against origin/main@1b130a2, option analysis, and decisions D-BW-1..D-BW-5"
topic: "Batch B1: writepath DefaultGitRunner has no -z and no git environment isolation (B83F53BB); writepath fixture self-test reads through links (05E12A6F)"
depth: "standard"
decision_status: "decided"
promoted_to: "plan"
linked_artifacts:
  - "docs/plans/2026-10-10-intercom-go-gatecheck-batch-b1-write-path-input-hardening-plan.md"
  - ".backlogit/queue/052-F.md"
  - ".backlogit/queue/042-S.md"
  - "docs/decisions/2026-10-08-intercom-go-gatecheck-batch-a-correctness-deliberation.md"
  - "docs/compound/2026-10-07-pin-canonical-text-maintenance-and-unbounded-ast-rule-review.md"
  - "docs/compound/workflow-issues/staging-premise-verification-before-planning-2026-09-18.md"
tags:
  - "gatecheck"
  - "writepath"
  - "fail-closed"
  - "p-021"
  - "stage"
---

# Deliberation: Gatecheck Batch B1, write-path input hardening

## Session context

* **Agent and mode:** Stage, invoked by the Orchestrator (`stage next`). Sequential, non-dark.
  The route is `claude-sonnet-5.5` / anthropic / xhigh (P-013.5, `config.model_routing.stage`).
* **Operator selection.** "stage next", immediately after the Orchestrator's pipeline assessment
  recommended Batch B1. The batch is exactly two active stash entries, **B83F53BB** (high, bug) and
  **05E12A6F** (low, task). This is the write-path twin of shipped Batch A (feature 051-F,
  shipment 041-S). Both were "left active for a future Stage session" by the Batch A
  deliberation.
* **Operator approval scope.** Stage applies the relayed instruction to two things: the batch
  selection, and the Stage-recommended option in each decision below. Any override can be made
  at staging-PR review, before Ship claims the shipment.
* **Base:** `origin/main` @ `1b130a2`. Stage artifact branch:
  `chore/stage-gatecheck-batch-b1-write-path-input-hardening`.
  * Slug source: the Step 1.5 grouping proposal title, "Gatecheck Batch B1 write-path input
    hardening". The slug is fixed at Step 1.9, before deliberation, as the Step Sequence
    Contract requires. Deliberation confirmed the title, so no rename is needed.
  * `1b130a2` differs from the Orchestrator's `3f92d7e` only by the merged memory-compaction PR
    #120 (docs only). No gatecheck code changed.
* **Explicitly excluded** (operator), none of which was triaged into this session:
  * the 33 other write-path entries from 039-S / 049-F (Root.Resolve detector precision,
    harness hygiene, write-primitive coverage);
  * retiredarch entries: DC921AF6, 3750C37C and 0ECC1895 (held by D-RA-7); 4A3851F0, A85D25F7,
    D17D5C5A, 31F33EFE, 4995F8C3, C8827920;
  * 5A8EC1BC (GitRunner-shape consolidation; trigger unfired), and every unignore, CI/script
    and docs entry.

## Step 1: Triage and P-021 obligations

| Stash ID | Kind / priority | Shape | DEFERRED SCOPE EXPANSION marker | Route |
|---|---|---|---|---|
| `B83F53BB` | bug / high | task | **yes** | deliberate, forced by P-021 C6 |
| `05E12A6F` | task / low | task | **yes** | deliberate, forced by P-021 C6 |

The grouping is the operator's selection: a single proposal, so no alternative groupings were
generated (the two-option rule applies from three eligible entries). The coherence rationale is
the same package (`tools/gatecheck/internal/writepath`), the same threat class (a gate input
that is silently wrong or read through a link), and the same deferral note.

### (A) Unconditional duplicate scan (both entries)

Ran over all 74 active stash entries (`logs/stash-snapshot.json`, refreshed at session start)
with the terms: `DefaultGitRunner`, `ls-files`, `-z`, `NUL`, `quotePath`, `gitRunnerEnv`,
`D7BF9F74`, `isolat`, `runFixtureSelfTest`, `fixture`, `symlink`, `Lstat`,
`containedRegularFile`, `junction`, `ReadDir`, `GIT_`, `SplitLines`, `GitText`, `GitRunner`,
`runRepoScan`, `shouldScan`, `ED-7`, `LookPath`, `contain`.

* **Nearest matches, none the same expansion:**
  * `5A8EC1BC`: consolidating three GitRunner shapes (an abstraction concern; trigger is a
    fourth engine). Related, handled under "Exclusions and evidence notes".
  * `31F33EFE`: the retiredarch twin of decision D-BW-1a below.
  * `C8827920`: retiredarch env-test table hardening (see D-BW-2).
  * `F5958BBC`, `B29A565E`: unignore test isolation and ref-argument hardening.
  * `6EF084E1`: oracle fixture-existence assertion. `C0D28448`: write-primitive denylist.
  * `4A3851F0`, `A85D25F7`, `D17D5C5A`, `4995F8C3`, `F9D52027`: retiredarch or unignore review
    follow-ups that only cite B83F53BB/05E12A6F as related.
* **B83F53BB vs 05E12A6F:** different surfaces (git-selected listing versus `os.ReadDir`
  fixture enumeration), so neither duplicates the other.
* **Verdict: Duplicate scan CLEAN** for both entries. No merge or archival of duplicates.

### (B) Late-identifier reconciliation (both entries: PR and review-thread were `N/A`)

* **Searched:** the Ship-owned residual-risk records that cite the entry IDs.
* **Recovered PR #116 (both entries).**
  * `docs/closure/041-S-051-F-post-merge-closure.md`, line 175: the Stage deferred list
    (`B83F53BB`, `05E12A6F`, ...), "None are fixed here (P-021 C1)".
  * `docs/archive/memory/2026-10-08/ship-041s-pr116-halt.md`, line 46: "Reused: `B83F53BB`
    (writepath `-z` and env isolation, pre-existing P1), `05E12A6F`, ...".
  * PR #116 (Batch A, shipment 041-S, merge `fcc61f3`) is the carrier PR that cites them as
    residual. It is **not** their capture origin: both were Stage own-authority intake during
    Batch A, before any PR existed.
* **Review-thread ID: no late identifier found.** PR #116 has three review threads (stash
  hygiene, readiness evidence, a stray control character); none concerns these findings. The
  findings came from local review. The `N/A` stands as a truthful terminal record.
* **Other refs carried forward:** feature 051-F; shipment 041-S; tasks 051.002-T (the retiredarch
  `-z` precedent) and 051.004-T (writepath containment); related stash D7BF9F74 (PR #95 carrier,
  see `docs/decisions/2026-10-07-intercom-go-retiredarch-gate-integrity-deliberation.md`).
* **Write half:** applied in place on both stash entries after the Step 1.9 gate (tracked
  mutation), as a `STAGE TRIAGE 2026-10-10` block. Read back and verified.

## Step 1.8: Learnings consulted

The learnings-researcher returned `confidence: medium` (no entry covers writepath `-z`, git
config isolation, or the fixture self-test directly). Applied:

* `docs/compound/2026-10-07-pin-canonical-text-maintenance-and-unbounded-ast-rule-review.md`:
  refreeze atomically; stage the freeze for a real red; fake `git` markers; bound the threat
  model first. **Outcome here:** no pin refreeze is needed (see D-BW-4), but the "no fake red"
  and "bounded threat model" rules govern D-BW-2.
* `docs/compound/workflow-issues/staging-premise-verification-before-planning-2026-09-18.md`:
  re-read each premise in its enclosing function; every AC must fail at the parent commit.
* `docs/compound/2026-09-08-adversarial-review-empirical-verification-and-scope-check.md`:
  containment tests need a LIVE link at the entry itself (not dangling, not a subpath); a
  junction needs `Lstat`, not `Stat`.
* `docs/compound/2026-10-02-git-show-ref-failure-ambiguity-resolve-before-show.md`: exact
  NUL-record parsing; never map a git failure to "nothing there"; scripted-runner tests.
* `docs/compound/workflow-issues/stash-jsonl-is-carried-pipeline-state-2026-09-19.md`,
  `docs/compound/2026-09-04-backlogit-sizing-is-wit-gated.md`: stash and sizing procedure.
* `docs/compound/2026-10-01-textual-yaml-guard-denylist-bypass-loop.md`: prefer small positive
  allowlists over growing denylists (informs D-BW-2).

## Premise verification (against `origin/main` @ `1b130a2`, direct reads)

### B83F53BB: the premise HOLDS (two independent defects)

**(1) No `-z`. HOLDS.**

* `writepath.go:147-160` (`DefaultGitRunner`): argv is
  `exec.Command("git", "ls-files", "--", "internal/**", "cmd/**")`. There is no `-z`.
* `writepath.go:830-848` (`runRepoScan`): `pysem.GitText(out)` then `pysem.SplitLines(listing)`,
  `p == ""` skipped, then `shouldScan(p)` (`:163-173`) keyed on the raw string.
* **Empirical confirmation (staging probe, git 2.55.0.windows.5, scratch repo under `logs/`,
  since deleted):** a tracked `internal/é.go` is printed as `"internal/\303\251.go"`, which
  begins with `"` and so fails `HasPrefix("internal/")`. The file is never scanned and the gate
  would pass. With `-z` the same listing is the exact bytes
  `cmd/x/main.go\0internal/bad.go\0internal/\xc3\xa9.go\0`.
* **A second hazard the entry does not name.** `GitText` also translates `\r` and `\r\n`, and
  `pysem.SplitLines` splits on Python line boundaries (U+0085, U+2028, U+2029, U+001C..U+001E).
  Only a NUL-delimited listing closes every variant.
* **No live exposure today.** Of 56 tracked `internal/**` + `cmd/**` files, none needs quoting
  and none contains a control character. `git ls-files -s` shows no mode-120000 entries.
* **Batch A did not change this.** `git diff 135ca59 origin/main` on `writepath.go` (+66/-4)
  adds `containedRegularFile` and the loop check only. `DefaultGitRunner` and
  `runFixtureSelfTest` are untouched.

**(2) No environment isolation. HOLDS.** There is no `cmd.Env`. Empirical per-vector results
from the staging probe (git 2.55, `git ls-files -z -- cmd/** internal/**`):

| Ambient vector | Observed effect on the unisolated runner | Gate outcome |
|---|---|---|
| `GIT_INDEX_FILE` = an alternate index that lists only `cmd/x/main.go` | listing silently omits `internal/bad.go` | **false-clean (fail-open)**: non-empty, so ED-7 does not fire |
| `GIT_INDEX_FILE` = a missing index | empty listing | fail-closed via ED-7 |
| `GIT_LITERAL_PATHSPECS=1` | `internal/**` matches literally: empty listing | fail-closed via ED-7 |
| `GIT_DIR` = a missing dir | `fatal: not a git repository` (rc 128) | fail-closed |
| `GIT_CONFIG_PARAMETERS` malformed | `fatal: unable to parse command-line config` (rc 128) | fail-closed |
| `GIT_CONFIG_GLOBAL` = a malformed file | `fatal: bad config line` (rc 128) | fail-closed |
| `GIT_CONFIG_COUNT=1` with no `KEY_0` | **not live on this host**: the agent host itself exports `GIT_CONFIG_COUNT=3`, `KEY_0..2` (`safe.bareRepository`, `credential.interactive`, `core.fsmonitor`), so `KEY_0` exists | needs a scrubbed control (see D-BW-2) |

* **Severity, stated precisely.** The env gap is **fail-open only** through a partial-index
  variable (`GIT_INDEX_FILE`, the one git sets for hooks and `git commit <paths>`); the other
  vectors fail closed but still make the gate flaky or blind to why it failed. The `-z` gap is
  fail-open for any tracked path that git C-quotes. This session's own host is a live example of
  ambient `GIT_*` leakage.
* **Adversary model (bounded, per the 2026-10-07 learning).** The adversary is *ambient or
  accidental* environment: agent hosts, git hooks, wrapper scripts. It is **not** an in-job
  attacker who can set arbitrary env or `PATH`: that actor already executes code in the job
  and is the same trust class as accepted residual R-A2b.

### 05E12A6F: the premise HOLDS

* `writepath.go:895-959` (`runFixtureSelfTest`): `os.Stat(fixtureDir)` at `:899` (follows
  links), `os.ReadDir` at `:904`, skips only `e.IsDir()` at `:910`, keeps names ending `.go`,
  then `scanFile(root, "scripts/testdata/writepath/"+name)` at `:927` with no `Lstat`.
* **Why a link slips through.** `DirEntry.IsDir()` reports the entry's own type, so a symlink (or
  a Windows junction, `ModeIrregular`) named `x.go` is not skipped and `pysem.ReadText` follows
  it. A linked ancestor (`scripts/`, `scripts/testdata/`, or the fixture dir itself) is followed
  by `os.Stat`.
* **There are 28 tracked direct `.go` fixtures** (37 tracked paths under the directory, the rest
  in `harness/` subdirectories that `ReadDir` does not enumerate). No tracked symlink exists.
* **New fact that shapes the fix.** `containedRegularFile` (051.004-T, `writepath.go:786-823`)
  already exists in this package. It does a component-wise `Lstat` walk from `root`, so reusing
  it covers the leaf **and** every ancestor. The entry did not know this.
* **Impact is limited, hence low priority.** Reading outside the tracked blob lets
  `--self-test`/`--self-test-integrity` verify bytes that are not the tracked fixture, so a
  "clean" or "rejected" fixture verdict could be made by a link target. Findings text contains
  only selector names and positions, so disclosure is minimal.

### Contract-surface matrix (what a change to the runner format must touch)

| Surface | Evidence (enclosing function) | Needs change for `-z`? |
|---|---|---|
| Runner argv | `DefaultGitRunner` `:147-160` | yes |
| Runner doc contract | `GitRunner` `:138-143` ("for pysem.GitText decoding") | yes (comment) |
| Parse + selection | `runRepoScan` `:835-848` | yes |
| Fake runner format | `u4StubGit` (`writepath_test.go:41-45`) | yes (via helper) |
| Inline newline fakes | `TestRun_ReadError_FailsClosed` (`:521`), `TestRun_InvalidUTF8_FailsClosed` (`:543`); `TestRun_EmptySelection_FailsClosed_ED7` is empty bytes (format-agnostic) | yes (via helper) |
| **Frozen oracle decode** | `TestOracle_TrackedProductionTree_Parity` (`writepath_oracle_test.go:188-212`) decodes the real runner output with `GitText` + `SplitLines`. Under `-z` that yields zero paths and the test `t.Fatal`s | **yes, see D-BW-4** |
| Wiring | `register_write_path.go:25` passes `DefaultGitRunner` | no (signature unchanged) |
| Real-repo goldens | `TestRun_SelfTest*_MatchesGolden`, `TestRun_RepoMode_CurrentTreeIsClean` run the real tree; output unchanged | no (INV-E) |
| Retiredarch pins | `retiredarch/writepath_mask_test.go` pins `scanFile`/`scanSource` mask flow and "no local masker" across writepath non-test sources only | **no refreeze** (INV-A: not touched) |
| Cross-package comment | `retiredarch/select.go:29-31`: "Its signature matches writepath.GitRunner, but writepath's listing is newline-separated and not -z (deferred entry B83F53BB)". The signature half is already wrong (retiredarch takes `pathspecs ...string`) and the listing half goes stale | yes (comment only) |
| Operator-facing docs / scripts | `rg` of `ls-files` / `newline-separated` outside tests and archives: only the comment above and `scripts/check-retired-architecture.sh:144` (an unrelated forward-looking note) | no |

## Problem frame

* **The core problem.** The merge-blocking write-path gate can pass without having examined every
  tracked `internal/**`/`cmd/**` Go file, for two reasons: git's default quoting drops some
  paths before the gate sees them (`-z` gap), and ambient git environment can redirect which
  files git lists (env gap). Separately, the gate's own self-test can read its fixtures through a
  link (05E12A6F).
* **Success criteria:**
  * every in-scope tracked path, whatever its bytes, is either scanned or fails the gate closed;
  * the listing the gate acts on does not depend on ambient `GIT_*` variables or global/system git
    config;
  * the fixture self-test refuses a linked or non-regular fixture, and a linked fixture directory;
  * every new behaviour test is red at its parent commit, and any characterization is labelled;
  * `scanFile`/`scanSource` stay byte-for-byte as pinned (INV-A).
* **Out of scope:** retiredarch changes beyond one stale doc comment; GitRunner consolidation
  (5A8EC1BC); unignore isolation; the non-absolute-`git`-path refusal (see D-BW-2); the other 33
  write-path entries.

## Decisions

### D-BW-1, `B83F53BB` (1): a NUL-delimited listing (`-z`) with a strict, unexported parse helper

| Option | Description | Pros | Cons | Effort |
|---|---|---|---|---|
| **A: `-z` + a strict NUL-record helper** (recommended) | `DefaultGitRunner` passes `ls-files -z --`. A new unexported helper in `writepath.go` requires a NUL terminator, rejects empty records, splits on NUL (no newline translation), applies `shouldScan`, validates UTF-8 (`pysem.ErrInvalidUTF8`) and control runes for the selected paths, and returns them in git's order. `runRepoScan` calls it. | Closes every quoting, CR and line-boundary variant. `writepath.go` is unfrozen, so the parse can be its own directly testable function. | Changes the `GitRunner` output contract, so every fake and the frozen oracle decode must follow (D-BW-4). | medium |
| B: `-c core.quotePath=false` + fail closed on any `"`-prefixed record | Non-ASCII paths print raw. | No fake-contract change. | `SplitLines` splits on U+0085/U+2028 and `GitText` rewrites `\r`; a `"` or `\` in a name is still quoted. Files needing quoting are refused, not scanned. | low-medium |
| C: C-unquote quoted records | Implement git's C-style unquoting. | No runner contract change. | A new parser with its own parity risk; line-split hazards remain. | medium-high |
| D: `-z`, parsed inline in `runRepoScan` (retiredarch's shape) | Same as A without a helper. | Mirrors retiredarch. | Retiredarch parsed inline only because `selectRepoPaths` is pin-frozen. Here it only makes a table test go through a `Result`. | medium |

**Decision: Option A.** Details:

* **Helper name and shape.** `scannedPathsFromListing(out []byte) ([]string, error)` (final name is
  Ship's; it must stay clearly distinct from retiredarch's frozen `selectRepoPaths`; plan-review
  flagged a one-letter difference as a review hazard). It returns selected paths in the **order
  git printed them**. It does not re-sort, so finding order and the goldens stay unchanged.
* **Failure modes** (each returns an error, never a skip):
  * a scanned record that is not valid UTF-8 returns `pysem.ErrInvalidUTF8` (error identity
    preserved; the `runRepoScan` context string `git ls-files output` is unchanged). A record that
    `shouldScan` rejects is never read or echoed, so it is not judged (plan-review SCOPE-F3: a
    whole-listing check would hard-fail the gate on an unscanned non-UTF-8 name, the same false
    positive D-BW-1a avoids);
  * a non-empty listing that does not end in NUL is an error. This catches a stale newline fake or
    a runner that ignored `-z`, which would otherwise be dropped silently;
  * an empty record is an error (no `continue` without a finding);
  * an empty listing returns `nil, nil`, and the existing ED-7 guard (`:850-855`) fails it closed.
* **`runRepoScan` otherwise unchanged:** ED-7, the `containedRegularFile` loop (051.004-T), and the
  findings block stay as they are.

### D-BW-1a, `B83F53BB` (1): fail closed on control and line-separator characters, **in paths that will be scanned**

* **Problem.** Under `-z`, git emits path bytes raw. A tracked `.go` path containing `\n`, `\r`, ESC,
  U+0085, U+2028 or U+2029 would be echoed verbatim by `errorLine(rel, ...)` and by the finding prefix
  `rel:line:` (both reach CI stderr), which could inject GitHub Actions workflow commands
  (`::error::`, `::set-output::`). Retiredarch fixed this in Batch A as D-BA-1a.

| Option | Description | Pros | Cons |
|---|---|---|---|
| A: reject any listed record containing such a rune (retiredarch parity) | The check runs before `shouldScan`. | Identical to retiredarch; simplest to state. | The listing includes **every** file under `internal/**`/`cmd/**`. A tracked non-Go file with a tab in its name would hard-fail the whole gate although it is never printed. This false positive is exactly what stash `31F33EFE` (medium) records for retiredarch. |
| **B: apply `shouldScan` first, then reject control/separator runes in the selected paths only** (recommended) | Only paths that are scanned are echoed, so only they need protecting. | Minimal sufficient rule; no false positive on unscanned names; every echoed path is safe. | Writepath and retiredarch differ until `31F33EFE` is decided. |

**Decision: Option B**, with the error text using `%q` and no raw control byte, for example
`::error::git ls-files output: path "internal/a\nb.go" contains a control or line-separator
character`. The rune set is `< 0x20`, `0x7f`, `0x85`, `0x2028`, `0x2029`, the same set as retiredarch.

* **Why this is safe, stated precisely.** In the **repo-scan listing arm**, the only places a listed
  path reaches output are `errorLine(rel, ...)` and the finding prefix, and both run only for paths
  that passed `shouldScan`. Paths that failed `shouldScan` are discarded silently by design (non-Go,
  `_test.go`), not echoed. This decision closes the **new** raw-byte path that `-z` opens; it does
  **not** claim that every CI-log sink is escaped. Pre-existing sinks outside the listing arm
  (fixture names in the self-test, `errorLine`'s `%v` tail with parser errors or raw git stderr,
  the absolute path inside the unwrapped `ReadText` error) are captured as a separate stash
  entry (plan-review SEC-2, SEC-3).
* **Bounded rule set (2026-10-01 learning).** The rune set equals retiredarch's set plus the
  `SplitLines` boundaries. A denylist over an open grammar does not converge, so any extension (C1
  controls, bidi overrides) routes to `31F33EFE` and the output-sink entry, not to this decision.
* **Parity note.** `31F33EFE` is retained and un-pulled: its retiredarch decision remains its own.
  This decision supplies the writepath precedent, and the operator may override to Option A at
  staging-PR review (a one-line change in the helper and one table row).

### D-BW-2, `B83F53BB` (2): git environment isolation by a semantic in-package copy of `gitRunnerEnv` (no PATH guard)

| Option | Description | Pros | Cons | Effort |
|---|---|---|---|---|
| **A: copy `gitRunnerEnv` into `writepath.go` and set `cmd.Env`** (recommended) | Strip every `GIT_*` entry (case-folded) except `GIT_CEILING_DIRECTORIES`, then append `GIT_CONFIG_NOSYSTEM=1`, `GIT_CONFIG_GLOBAL=<os.DevNull>`, `GIT_CONFIG_SYSTEM=<os.DevNull>`. Doc comment cites D7BF9F74 and says "copied, not shared (D-BA-3 / 5A8EC1BC); keep in sync by hand". | Matches the entry's stated scope. Proven shape (retiredarch ships it, and its CI job is the same job that runs this gate). Closes the fail-open partial-index vector. | A second copy to keep in sync. | low-medium |
| B: A, plus retiredarch's `cmd.Err` + non-absolute `git` path refusal | Also refuse a `git` that `PATH` resolved to a relative path. | Full runner parity. | Outside the entry's scope (the entry names env/config isolation only). The refusal needs a fake-`git` `TestMain` scaffold in a package that has none. The guarded actor (controls `PATH` and `GODEBUG`) is the in-job attacker, the same trust class as R-A2b. | medium |
| C: a shared `internal/gitrun` package for both engines | One implementation. | No duplication. | Re-touches the pin-frozen `retiredarch/select.go` and `pin.go` (atomic refreeze), pre-empts 5A8EC1BC, and its trigger (a fourth engine) has not fired. | high |

**Decision: Option A.** Constraints:

* **Scope.** `cmd.Env = gitRunnerEnv(os.Environ())` is set in `DefaultGitRunner` after `-z`. No
  `cmd.Err` guard and no relative-path refusal; that residual is **captured as a new stash entry**
  (P-021 C1), not silently dropped (see "Dispositions").
* **No `filepath.Abs` / `filepath.EvalSymlinks`** (INV-B, `retiredarch/initstate_test.go`).
* **Positive wiring, not a denylist** (2026-10-01 learning): the filter keeps what is not `GIT_*`
  and pins the three config entries explicitly.
* **Adversary and boundary.** The adversary is *ambient or accidental* environment; the in-job
  attacker is out of model (R-A2b). The isolation covers the `git ls-files` child **for a given
  root**. The CI wrapper's own root discovery (`ROOT="$(git rev-parse --show-toplevel)"`,
  `scripts/check-write-path-precondition.sh:150`) is unscrubbed, so ambient `GIT_DIR` and
  `GIT_WORK_TREE` can still select a different tree. That, and the fact that ED-7 detects only an
  *empty* selection (not a partial one), are captured as a separate stash entry (plan-review
  ARCH-AS1, SEC-1, SEC-4); this decision does not claim end-to-end environment independence.
* **Test vectors (writepath-specific; each red at the parent; three, not five).** A real-git
  fixture repo (scrubbed `fixtureGit`, as in `retiredarch/select_test.go`) is built first. The
  baseline is the scrubbed `git ls-files -z` bytes, asserted to name all fixture files. Each vector
  is set with `t.Setenv` **after** the repo exists, and the oracle is that `DefaultGitRunner(root)`
  returns no error and bytes **equal to the baseline** (never the exit code alone):
  * (a) `GIT_INDEX_FILE` → an alternate index that lists only the clean file. **Bug evidence:** at the
    parent `runRepoScan(root, DefaultGitRunner)` returns Code 0 although tracked `internal/bad.go`
    holds a write primitive; after the change it returns Code 1 naming `internal/bad.go`;
  * (b) `GIT_LITERAL_PATHSPECS=1` → at the parent an empty listing; after, the baseline listing;
  * (c) `GIT_CONFIG_PARAMETERS` malformed → at the parent git exits non-zero, so the runner errors;
    after, the baseline listing.
* **Why not the mandatory-liveness design.** An earlier draft added a per-vector liveness control
  (stash `C8827920`'s idea) and two more vectors (`GIT_CONFIG_GLOBAL`, `GIT_CONFIG_COUNT`).
  Plan-review (SCOPE-F2) showed that this ports an operator-excluded entry and that the observed
  red at the parent already proves each vector is live when written. The `GLOBAL`/`COUNT`
  channels are covered by the pure filter table (the pins are in the slice) and by (c), which
  exercises the same `GIT_*` prefix filter. `C8827920` stays active for the retiredarch table.
* **Pure filter test.** A table test of `gitRunnerEnv`, identical to retiredarch's so that drift
  fails a test (case-insensitive prefix, `GIT_CEILING_DIRECTORIES` kept in any case, Windows `=C:`
  per-drive entries kept, a `GITX` lookalike kept, the three appended entries, `nil` input).
* **Doc comment.** `DefaultGitRunner` states the bounded threat model (ambient env only; git 2.32+
  assumed for the config pins; `PATH` shims, an in-job attacker and a self-modifying PR out of
  scope).

### D-BW-3, `05E12A6F`: reuse the existing `containedRegularFile` per enumerated fixture

| Option | Description | Pros | Cons |
|---|---|---|---|
| **A: call `containedRegularFile(root, relPath)` for each enumerated `.go` name, before `scanFile`** (recommended) | Fail closed with `errorLine(relPath, fmt.Errorf("not a contained regular file: %s", reason))`, Code 1, the same message shape as the repo scan. | No new code or copy (the function is in-package). Covers the leaf link **and** a linked `scripts/`, `scripts/testdata/` or fixture directory. Same semantics as 051.004-T. | None significant. |
| B: test only `e.Type()` (symlink or irregular) on the `DirEntry` | A leaf-only check. | Tiny. | Misses a linked fixture directory (`os.Stat(fixtureDir)` follows it), so it is a partial fix of the same class. |
| C: `Lstat` the fixture dir plus the leaf type | Two bespoke checks. | Covers both. | Re-implements what `containedRegularFile` already does. |

**Decision: Option A.** Constraints:

* The check goes in `runFixtureSelfTest`'s loop, **not** inside `scanFile` (INV-A: the retiredarch
  pin requires `scanFile` to be exactly ReadText then `scanSource`).
* A non-existent final component falls through (`fs.ErrNotExist` returns ok), so a fixture deleted
  after `ReadDir` keeps the existing read-error text. The `fixture dir not found` and
  `no fixtures discovered` messages are untouched.
* Real fixtures are regular files, so the self-test goldens are unchanged (INV-E).
* **Test rows** (a table; a live link at the entry itself, per the empirical-verification
  learning): (1) a leaf symlink `reject-x.go` to a file outside root; (2) a leaf **directory
  symlink** named `reject-dirlink.go` (a message-only red, not bug evidence); (3) a linked
  **ancestor**, three sub-cases (`scripts`, `scripts/testdata`, `scripts/testdata/writepath`), each a
  junction (Windows, unprivileged) or a symlink to a real directory holding a fixture, which makes
  the red observable on the Windows dev host; (4) a characterization control: regular fixtures and a
  regular subdirectory still pass. A Windows leaf junction is not pinned to one outcome: Ship traces
  the pinned toolchain's `os` source (`types_windows.go`) to record how `ReadDir` and `Lstat`
  classify it, and a Windows-only sub-row asserts the property that matters, "never read through"
  (no `PASS <name>` line; either skipped or failed with the containment text). The TOCTOU between
  the `Lstat` walk and the read is an accepted residual, as in retiredarch U3 and 051.004-T.
  Symlink creation skips **only** on the Windows privilege error (`u4CreateSymlink` already does
  this); the Linux CI job runs rows (1) and (2).

### D-BW-4: sequencing, atomicity, and the frozen oracle

* **No pin refreeze is needed.** Unlike Batch A's U2, nothing here edits `scanFile`, `scanSource`,
  or any retiredarch pinned declaration. The retiredarch writepath pin
  (`writepath_mask_test.go`) inspects only those two functions and local masker declarations
  across writepath non-test files. The "refreeze atomically" ALP rule therefore does not
  apply; its spirit does: a runner-format flip and every fake that depends on it land in one
  commit.
* **First, a test-only prep task (characterization).** It introduces `lsFilesListing` (renders a
  listing in the runner's *current* format) and `decodeLsFilesListing` (decodes a real runner
  output, *current* format) in `writepath_test.go`. Every fake and the oracle decode switch to
  them. The `-z` task then only flips the two helpers. `TestRun_ReadError_FailsClosed` and
  `TestRun_InvalidUTF8_FailsClosed` are also tightened to assert that stderr **names the intended
  path** (green before and after, labelled characterization). Reason: after the flip, a stale newline
  fake would be rejected as "not NUL-terminated" and those two tests would still pass for the wrong
  reason.
* **The frozen oracle.**
  * `writepath_oracle_test.go` is header-frozen: "The only authorised edit is the single AC-E2.6
    adaptation... Any other edit is an H-3 stop."
  * `TestOracle_TrackedProductionTree_Parity` reads the real runner output directly, so a runner
    format change cannot avoid touching it.
  * **Decision:** the prep task makes exactly one edit there: it replaces the inline
    `GitText` + `SplitLines` decode (`:194-203`) with `decodeLsFilesListing` and replaces the
    LIFECYCLE paragraph (`:9-12`) with the reconciled wording written out in the plan (U1), so the
    file does not contradict itself. The legacy side and every expectation are untouched. The
    operator's approval of the staging PR is the written authorization for this single edit. Ship
    stops and reports (H-3) if more than the decode block and the header paragraph would have to
    change.
  * **Limits of the authorization (plan-review SEC-5).** The authority is process-only (the same
    trust class as R-A2b), and the decode now lives in a helper in the unfrozen `writepath_test.go`.
    The helper's doc comment therefore says that changing it is an oracle edit, and the header cites
    D-BW-4 and the staging PR.
  * Alternatives rejected: leaving a newline shim in production (keeps the `SplitLines` hazard) or
    leaving the oracle red.
* **Independence of the decode.** The flipped `decodeLsFilesListing` is its own bytes-based NUL parse.
  It must not call the production helper (independent re-derivation, H-11 spirit).
* **Task order:** prep, then `-z`, then env isolation (same function), then fixture containment
  (independent), then the stale-comment fix.

### D-BW-5: the shape of the release unit

* **One covering feature:** "Gatecheck Batch B1 write-path input hardening".
* **Five tasks, directly under the feature** (the 050-F/051-F precedent; no sub-epics):
  1. test-only prep (shared listing helper, oracle decode adaptation);
  2. `-z` selection and the control-character rule;
  3. git environment isolation;
  4. fixture-self-test containment;
  5. refresh the stale `retiredarch/select.go` comment (comment only).
* **One queued shipment.**

## Exclusions and evidence notes (retained, not triaged)

* **5A8EC1BC (GitRunner consolidation): stays deferred, trigger unfired.** B1 adds a second copy of
  `gitRunnerEnv` semantics, a second NUL parse and a second `containedRegularFile` across engines,
  and `unignore` still has no env isolation by documented parity. This is new evidence for the
  entry's cost argument but is not its trigger (a *fourth* engine). The operator allowed an
  evidence note ("note any new evidence only"); one is appended. It also records the architecture
  reviewer's suggestion to extend the reconsideration trigger to repeated cross-engine fixes or
  unintended divergence, and which contracts must stay equivalent (environment filtering,
  containment) versus which are engine policy (control-character scoping, sorting).
* **31F33EFE:** the retiredarch twin of D-BW-1a; retained. **C8827920:** retained for the retiredarch
  table; its mandatory-liveness idea is **not** adopted in B1 (see D-BW-2).
  **4995F8C3:** new comments cite stash IDs, not unit tags; retained.

## Dispositions (P-021: Stage alone triages)

| Stash ID | Disposition |
|---|---|
| `B83F53BB` | **Consumed.** Harvested into the prep task, the `-z` task, the env-isolation task, and the stale-comment task. Duplicate scan clean; PR #116 recovered; thread `N/A` stands. Archived after shipment assembly. |
| `05E12A6F` | **Consumed.** Harvested into the fixture-containment task. Same reconciliation outcome. Archived after shipment assembly. |
| *new, low* `4372BAD4` | **Captured.** DEFERRED SCOPE EXPANSION: writepath `DefaultGitRunner` does not refuse a relative `git` path (retiredarch parity: `cmd.Err` + non-absolute refusal; residual R-A2b). Left active. |
| *new, medium* `DBE25DF5` | **Captured** (plan-review ARCH-AS1, SEC-1, SEC-4). DEFERRED SCOPE EXPANSION (kind bug): the write-path wrapper derives `ROOT` with an unscrubbed `git rev-parse --show-toplevel`, and ED-7 detects only an empty selection, so a wrong-tree or partial listing can pass (a per-arm `internal/` and `cmd/` non-empty guard is one option). Left active. |
| *new, low* `EDC18D59` | **Captured** (plan-review SEC-2, SEC-3). DEFERRED SCOPE EXPANSION: pre-existing CI-log echo sinks outside the listing arm are unescaped: fixture names in the self-test, `errorLine`'s `%v` tail (parser errors, raw git stderr), and the absolute path in the unwrapped `ReadText` error. Left active. |
| `5A8EC1BC` | **Retained**, deferred, with an evidence note. |

## Trade-off comparison (summary)

| Criterion | D-BW-1 A | D-BW-1 B | D-BW-2 A | D-BW-2 B | D-BW-3 A | D-BW-3 B |
|---|---|---|---|---|---|---|
| Closes the finding | yes | partial | yes (entry scope) | yes (wider) | yes | partial |
| Pin churn | none | none | none | none | none | none |
| Blast radius | `writepath.go` + tests | same | `writepath.go` + tests | + `TestMain` scaffold | 1 loop | 1 loop |
| Reuses precedent | retiredarch D-BA-1 | no | retiredarch D7BF9F74 | retiredarch U4 | 051.004-T | none |

## Unresolved questions

* **Control-character scoping (D-BW-1a)** is a recommendation, not a certainty: if the operator
  wants strict engine parity, switch to Option A and decide `31F33EFE` the same way.
* **The relative-`git`-path refusal** is deliberately out of scope (D-BW-2 B); the captured entry
  records it for a later decision.
* **End-to-end environment independence** is not claimed: the wrapper's root discovery and
  partial-listing detection are captured separately.
* **Whether a given env vector stays live on a future git version** is not guarded by a control;
  each vector is observed red at the parent when written, and the baseline is asserted non-vacuous.

## Risks and mitigations

| Risk | Mitigation |
|---|---|
| A fake or test still emits newline listings after the flip and passes for the wrong reason (RK-2 in Batch A) | The prep task centralises fakes and tightens two assertions to name the path; the strict parser rejects a non-NUL-terminated listing |
| The frozen oracle edit is read as an H-3 violation | Recorded here and in the plan as a single authorized adaptation, with the exact replacement header written out; stop-and-report if wider |
| A replacement env vector is non-live on the CI git version | Each vector is observed red at the parent; bytes are compared with a non-vacuous scrubbed baseline, never the exit code alone |
| Isolation breaks the gate on a runner with a global `safe.directory` | The retired-architecture gate already runs in the same CI job and checkout with the same isolation (040-S/041-S, PRs #112 and #116) and passes; the first CI run on the Ship PR is the validation window |
| Symlink tests skip on unprivileged Windows | Junction rows run unprivileged on Windows; the Linux CI job runs the symlink rows; record where the red was observed |
| Path bytes reach CI logs (workflow-command injection) | D-BW-1a: control/separator characters in scanned paths fail closed with a `%q` message; other sinks are captured separately |

## Rejected alternatives

* **D-BW-1 B, C and D.** B and C leave the `SplitLines`/`\r` hazards and refuse legitimate paths;
  D adds nothing for an unfrozen file.
* **D-BW-2 B and C.** B exceeds the entry's scope and adds heavy test scaffolding for the
  in-job-attacker class; C pre-empts 5A8EC1BC and forces a pin refreeze.
* **D-BW-3 B and C.** Partial or duplicative of an existing helper.

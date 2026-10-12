---
title: "Gatecheck write-path Batch 2: scan-set completeness, root discovery, fixture self-test, CI-log escaping, env-twin hardening"
description: "Stage deliberation over Batch 2 (stash EEE39D23, DBE25DF5, 858516B9, EDC18D59, 2613D248): premise verification against the 042-S/PR #123 merge state, option analysis, decisions D-WB2-1..D-WB2-10, shipment split"
topic: "Batch 2: write-path scan-set misses non-.go tracked records (symlink, gitlink) (EEE39D23); wrapper root discovery is unscrubbed and ED-7 detects only an empty listing (DBE25DF5); fixture self-test silently skips a .go-named directory (858516B9); unescaped CI-log echo sinks (EDC18D59); MSYS/CYGWIN env and twin-table boundary rows (2613D248)"
depth: "standard"
decision_status: "decided"
promoted_to: "plan"
linked_artifacts:
  - "docs/plans/2026-10-11-intercom-go-gatecheck-write-path-batch-2-plan.md"
  - "docs/decisions/2026-10-10-intercom-go-gatecheck-batch-b1-write-path-input-hardening-deliberation.md"
  - "docs/plans/2026-10-10-intercom-go-gatecheck-batch-b1-write-path-input-hardening-plan.md"
  - "docs/plans/2026-10-11-intercom-go-closure-tooling-batch-1-plan.md"
  - "docs/closure/042-S-052-F-post-merge-closure.md"
  - "docs/compound/workflow-issues/staging-verification-commands-and-ambient-git-env-2026-10-10.md"
  - "docs/compound/2026-10-07-pin-canonical-text-maintenance-and-unbounded-ast-rule-review.md"
tags:
  - "gatecheck"
  - "writepath"
  - "retiredarch"
  - "fail-closed"
  - "p-021"
  - "stage"
---

# Deliberation: Gatecheck write-path Batch 2

## Session context

* **Agent and mode:** Stage, invoked by the Orchestrator (normal sequential mode; operator may be
  absent). The Orchestrator relayed an ordered three-batch directive. Batch 1 (shipment 043-S,
  closure tooling) is staged. **This session is Batch 2 only.** Batch 3 (stash `C5E50EFD`) is not
  staged here.
* **Operator selection:** exactly five active stash entries: `EEE39D23` (medium, bug), `DBE25DF5`
  (medium, bug), `858516B9` (low, bug), `EDC18D59` (low, task), `2613D248` (low, task). The
  selection is the operator's directive, so no alternative grouping menu was generated.
  Deliberation validates the grouping and may split it (D-WB2-1).
* **Base and branch:** Stage artifact branch `chore/stage-writepath-batch-2`, created from the
  Batch 1 branch tip `f20b6d1` (`chore/stage-closure-tooling-batch-1`), **not** from `origin/main`
  (`cc34242`). This is the operator's stated preference, taken so the new shipments can carry a
  `blocks` edge on 043-S, which exists only on that branch. The Step 1.9 checks all passed on that
  base: one worktree (the current one), the branch name did not exist, `git merge-base` with
  `origin/main` is `cc34242`, and the base branch's whole diff against `origin/main` lies under
  Stage artifact roots (`.backlogit/`, `docs/decisions`, `docs/memory/2026-10-11`, `docs/plans`).
  * Slug source: the proposed covering-feature grouping title "Writepath batch 2". Deliberation
    refined the feature titles (below); the slug is fixed at Step 1.9 and is not renamed.
* **Merge order (REQUIRED):** the Batch 1 staging PR (`chore/stage-closure-tooling-batch-1`) must
  merge first, as a merge commit. This branch carries the Batch 1 commit `f20b6d1`, so after that
  merge this PR's diff is only Batch 2 artifacts. A squash merge of the Batch 1 PR would
  re-introduce its content here as a conflicting duplicate; do not squash it.
* **Excluded by the operator and by scope:** Batch 3 (`C5E50EFD`); every other write-path,
  retiredarch, unignore and docs stash entry. Near neighbours that were read and deliberately NOT
  consumed: `4372BAD4`, `31F33EFE`, `C8827920`, `9EC5C536`, `65AE88CA`, `96B2E8CC`, `5A8EC1BC`,
  `C0D28448`, `FE2F02FF`, `CDBA41E3`, `F9D52027`.

## Step 1: Triage and P-021 obligations

| Stash ID | Kind / priority | DEFERRED SCOPE EXPANSION marker | Shape | Route |
|---|---|---|---|---|
| `EEE39D23` | bug / medium | **yes** | task-shaped | deliberate (P-021 C6) |
| `DBE25DF5` | bug / medium | **yes** | task-shaped | deliberate (P-021 C6) |
| `858516B9` | bug / low | **yes** | task-shaped | deliberate (P-021 C6) |
| `EDC18D59` | task / low | **yes** | task-shaped | deliberate (P-021 C6) |
| `2613D248` | task / low | **yes** | task-shaped | deliberate (P-021 C6) |

All five carry the marker, so the precedence rule forces the deliberate route; none may reach
planning without this artifact.

### (A) Unconditional duplicate scan

Scanned all active stash entries for: `show-toplevel`, `rev-parse`, `shouldScan`,
`containedRegularFile`, symlink, gitlink, `gitRunnerEnv`, `MSYS`, `CYGWIN`, `errorLine`, fixture
self-test, FIFO, hard link, `IsDir`, workflow command, `::error::`, case-variant, wrong tree, ED-7,
escap*.

* **Verdict: CLEAN for all five entries.** No entry describes the same expansion as another, and
  none of the five duplicates an existing entry.
* Nearest neighbours, none the same expansion: `4372BAD4` (writepath relative-`git` refusal),
  `31F33EFE` (retiredarch control-character ordering), `C8827920` (retiredarch env-test liveness
  control), `9EC5C536` (writepath test oracles for B1), `65AE88CA` (writepath test-helper
  consolidation), `96B2E8CC` (GitRunner context), `5A8EC1BC` (GitRunner consolidation), `C0D28448` /
  `FE2F02FF` / `CDBA41E3` (write-primitive coverage), `F9D52027` (unignore `::error::` doubling).
* The five entries are also pairwise distinct: they cover the scan-set (EEE39D23), wrapper root
  discovery (DBE25DF5), fixture directory enumeration (858516B9), output escaping (EDC18D59) and
  the git environment filter (2613D248).

### (B) Late-identifier reconciliation (every entry recorded PR and thread as `N/A`)

* **Retrieval source:** Ship-owned closure record `docs/closure/042-S-052-F-post-merge-closure.md`
  (lines 46, 123, 124, 133, 171) and `docs/archive/memory/2026-10-11/ship-042-s-session-memory.md`.
* **PR recovered: #123** (carrier PR of shipment 042-S / feature 052-F, merged; its post-merge
  closure record was merged as PR #124). The closure record cites all five entries as
  accepted-interim residuals (operator disposition 2026-10-11T21:47 -07:00, "item 2": the
  write-path batch). It is the carrier, **not** the capture origin: `EEE39D23`, `858516B9` and
  `2613D248` were Ship C2 captures made pre-PR; `DBE25DF5` and `EDC18D59` were Stage own-authority
  intake during Batch B1 plan-review.
* **Feature and shipment IDs recovered for `DBE25DF5` and `EDC18D59`** (recorded `N/A at capture`):
  feature `052-F`, shipment `042-S`.
* **Review-thread ID: no late identifier found.** PR #123's Copilot review was COMMENTED with zero
  review threads; the findings came from local and plan-review passes. The recorded `N/A` stands as
  a truthful terminal record. **Task ID: no late identifier found** (the findings are not tied to
  a single task); the `N/A` stands.
* **Write half (the stash `text` edit) is applied after the Step 1.9 gate.** The reconciled
  identifiers are carried here and into the stash entries as a `STAGE TRIAGE 2026-10-11` block.

## Step 1.5: Grouping

The grouping is the operator's selection (five entries, one proposal). Coherence rationale: every
entry is a residual of the **write-path gate** (`tools/gatecheck/internal/writepath` and its
wrapper) or of its retiredarch twin, all accepted as interim by the operator at the 042-S closure.
Deliberation concluded that the five **do not form one safely reviewable release unit** and splits
them into two sequenced shipments (D-WB2-1).

## Step 1.8: Learnings consulted (confidence: medium)

* `docs/compound/workflow-issues/staging-verification-commands-and-ambient-git-env-2026-10-10.md`:
  covers `DBE25DF5` directly. Isolating the child does not isolate how the script chose its inputs;
  compare listing **bytes**, never exit codes; probe from a scrubbed environment first.
* `docs/compound/2026-10-07-pin-canonical-text-maintenance-and-unbounded-ast-rule-review.md`:
  `pin.go` `canonicalDecls`, `closedWorldDecls` and the frozen sets must be refrozen **atomically**
  with a `select.go` change; split a batch rather than loop in review.
* `docs/compound/2026-10-01-textual-yaml-guard-denylist-bypass-loop.md`: prefer shape allowlists
  that fail closed over named-spelling denylists (informs D-WB2-2 and D-WB2-6).
* `docs/compound/2026-09-30-go-msys-bash-bridging-and-unicode-case-mapping-gap.md`: MSYS path-form
  and short-name traps (informs D-WB2-4 and the Windows rows).
* `docs/compound/2026-09-08-adversarial-review-empirical-verification-and-scope-check.md`:
  reproduce the exact reported shape before scoping (used for every premise below).
* `docs/memory/compacted/2026-10-11-042-s-052-f-compacted.md`: B1 pattern; Windows symlink fixtures
  skip on Errno 1314 and take their red evidence from Linux CI; the plan failed review once.

## Premise verification (direct reads and probes, this session)

### EEE39D23: the premise HOLDS

* `writepath.go` `scannedPathsFromListing` returns only records for which `shouldScan` is true
  (`.go` suffix, not `_test.go`, under `internal/` or `cmd/`). Every other listed record is
  dropped silently (`continue // by design: never read or echoed`).
* **Probe (scratch repo, git 2.55.0.windows.5, `git update-index --cacheinfo`):** a tracked symlink
  record `internal/fx` (mode 120000) and a gitlink record `internal/sub` (mode 160000) **are**
  returned by `git ls-files -z -- internal/** cmd/**`. Neither ends in `.go`, so `shouldScan` drops
  them, `containedRegularFile` is never called, and nothing is read. A symlink `internal/fx`
  pointing at a directory outside the scan scope that holds a write primitive, imported by a `cmd`
  package, would let the gate exit 0. This is the exposure the Ship review reported.
* A `.go`-suffixed symlink record is already refused (B1/D44D8BDF `containedRegularFile`), so the
  gap is exactly the non-`.go` and `_test.go` records.
* The repo today has 56 tracked records under `internal/**` and `cmd/**` and none is a symlink or
  gitlink (`git ls-files -s` shows only modes 100644 and 100755), so widening the check creates no
  new failure on the current tree.
* A case-variant record (`Internal/x.go`) is **not** returned for the pathspec `internal/**` in
  the probe (on a `core.ignorecase=true` Windows repo), and on a case-sensitive filesystem it is a
  different directory.

### DBE25DF5: the premise HOLDS (both halves)

* `scripts/check-write-path-precondition.sh:150`: `ROOT="$(git rev-parse --show-toplevel)"`, no
  environment scrub; `gatecheck_invoke` (`scripts/lib/gatecheck-run.sh`) passes it as `--root`.
* `scripts/check-retired-architecture.sh:198`: `ROOT="$(git -C "${GATECHECK_SRC}" rev-parse
  --show-toplevel)"`, also unscrubbed (and `:192` is a separate caller-cwd precondition).
* **Probe (scratch repo, ambient variables set in the same shell):**
  * `GIT_DIR=<decoy>/.git GIT_WORK_TREE=<decoy>`: `git rev-parse --show-toplevel` from the real
    checkout printed the **decoy** root.
  * The retiredarch form `git -C <real>/scripts rev-parse --show-toplevel` under the same ambient
    variables printed the **decoy** root.
  * `GIT_DIR` alone printed the cwd (the real root only because the probe ran at the root); run
    from a subdirectory it would select that subdirectory, i.e. a partial tree.
* `runRepoScan` (`writepath.go`) fails closed only when `len(relPaths) == 0` (ED-7). A decoy that
  contains `internal/` and `cmd/` files yields a non-empty listing and a clean verdict for the wrong
  tree.
* `retiredarch.gitShowToplevel` (`pin.go:753`) already scrubs through `gitRunnerEnv` (D7BF9F74), so
  the Go-side pin root is **not** affected. Only the two shell wrappers are.
* The wrapper root discovery is the same pattern in `check-unignore-regression.sh:106`,
  `check-merge-strategy.sh:80` and `check-depguard-fixtures.sh:41` (cwd-based). Those are out of
  scope here and captured as a new entry (D-WB2-9).

### 858516B9: the premise HOLDS

* `runFixtureSelfTest` skips any `os.ReadDir` entry for which `e.IsDir()` is true, so a real
  directory named `accept-x.go` or `reject-x.go` is dropped silently and the self-test can pass
  with a fixture it never ran.
* **Constraint found:** `scripts/testdata/writepath/` legitimately contains one real directory,
  `harness/`. A rule that refuses every directory would break the live self-test.
* A symlink entry is not `IsDir()` (DirEntry type comes from `lstat`), so it already reaches
  `containedRegularFile` and is refused (B1 U4). A FIFO named `*.go` also reaches it and is refused
  as "not a regular file", but no test covers it.

### EDC18D59: the premise HOLDS (three sinks)

* `errorLine` formats `fmt.Sprintf("::error::%s: %v\n", context, err)` with no escaping.
* `runFixtureSelfTest` prints `PASS %s`, `FAIL %s` (name and joined findings) and the U4
  `not a contained regular file: <reason>` line, all with the raw directory-entry name.
* `scanFile` returns `pysem.ReadText`'s `*fs.PathError` unwrapped; the absolute runner path reaches
  stderr. `containedRegularFile` already unwraps its own `Lstat` error.
* `pysem.Repr` already escapes selectors in finding text, and repo-scan relPaths are rune-validated
  (D-BW-1a), so the repo-scan finding lines are not a sink. The retiredarch engine has the same
  `::error::%v` pattern (`retiredarch.go:384,389,536,596,695`); see D-WB2-9.

### 2613D248: the premise HOLDS as a boundary gap, no exploit demonstrated

* `writepath.gitRunnerEnv` and `retiredarch.gitRunnerEnv` (`select.go:73`) are identical copies
  that strip names with a `GIT_` prefix except `GIT_CEILING_DIRECTORIES`. Neither strips `MSYS`,
  `CYGWIN` or `MSYS2_*`, which the MSYS2 runtime in Git for Windows reads (path conversion and
  runtime behaviour). Ship's probe found no demonstrated exposure; this is defence in depth for a
  Windows-only surface, and the required gate is Linux. Priority stays low.
* **Pin coupling (verified):** `retiredarch/select.go gitRunnerEnv` is a frozen declaration. Its
  text is mirrored in `pin.go` `canonicalDecls` (line ~121), listed in `closedWorldDecls`
  (line 214) and `pathspecFrozenDecls` (line 241), and exercised by
  `pin_test.go TestCheckPathspecPin_GitRunnerIsolation_Rejected`. A change to `select.go` alone
  turns the pin gate red.

## Decisions

### D-WB2-1: Split into TWO sequenced shipments

| Option | Verdict |
|---|---|
| **A. One shipment of ~15 tasks** across three surfaces (writepath Go, shell wrappers, frozen retiredarch pin) | Rejected. Review blast radius is too wide for one PR; the prior 14-unit plan in this repo had to be split after three plan-review FAILs; B1 (5 tasks) already needed two review attempts and hit the review-cycle cap. One review finding in the shell surface would hold the engine fixes hostage. |
| **B. Two shipments, split by surface (chosen)** | **S1 "Write-path gate completeness and CI-log output safety"** (EEE39D23, EDC18D59, 858516B9): one package, `writepath.go` and its tests only. **S2 "Gate git discovery and environment isolation"** (DBE25DF5, 2613D248): two wrappers plus the shared shell lib, the twin `gitRunnerEnv` tables, and the atomic retiredarch pin refreeze. |
| C. Split by priority (mediums first) | Rejected. EEE39D23 and DBE25DF5 are both medium, but DBE25DF5 needs ambient `GIT_DIR`/`GIT_WORK_TREE`, while EEE39D23 needs only a committed symlink (a hostile-PR path), so the engine fix is the more likely exposure. It also would put new raw-echo refusal paths into the codebase before the escaping chokepoint exists. |

**Sequence:** `043-S` (Batch 1) → S1 → S2, each with a `blocks` dependency edge on its predecessor.
**Why S2 follows S1 rather than running in parallel:** both edit `writepath.go` and
`writepath_test.go` (S2 touches only `gitRunnerEnv` and `TestGitRunnerEnv`); stacking avoids a
conflicting rebase. **Why S1 precedes S2 on content:** the new refusal messages of EEE39D23 and
858516B9 should be born escaped, so the EDC18D59 chokepoint lands first inside S1.

### D-WB2-2: EEE39D23, widen the scan set by applying `containedRegularFile` to every non-selected record

| Option | Verdict |
|---|---|
| **A. After the listing parse, run `containedRegularFile` on every listed record that `shouldScan` does not select; keep the existing in-loop check for selected records (chosen)** | Smallest change that closes the gap with the existing, twice-reviewed control. A symlink record fails at the final component (`ModeSymlink` is not regular); a gitlink fails because a checked-out submodule is a directory (not regular); a symlinked ancestor fails at the intermediate component. A tracked record absent from the working tree keeps the existing `fs.ErrNotExist` pass (a deleted non-Go file is not built). Ordering is preserved for selected records, so existing error precedence does not move. |
| B. Switch the runner to `git ls-files -z -s` and refuse modes 120000 and 160000 from the index | Rejected. Changes the injectable `GitRunner` output contract pinned by the 042-S wiring (`register_write_path.go:25`) and every scripted-runner test. It also checks the index, not the working tree that `go build` compiles (core.symlinks=false checkouts). |
| C. Fold every record into the existing in-loop check | Rejected. Moves a containment failure on record N ahead of a read error on record 1, a visible error-precedence change for no added safety. |
| D. Per-arm non-empty guard (`internal/` and `cmd/` both present) | Out of this decision; see D-WB2-5. |

* **New outcome:** one new fail-closed refusal, reusing the existing message shape
  `::error::<rel>: not a contained regular file: <reason>` (same sink class as the U4 line, so a
  single escaping rule covers it). The plan carries a full outcome matrix.
* **Non-selected record names are not rune-validated** (unlike selected paths). Hostile names are
  safe only because every echo goes through the D-WB2-6 chokepoint; this is why the escaping
  tasks precede the widening tasks. The control-character false-positive asymmetry
  (`31F33EFE`) is deliberately not widened.
* **Tradeoff accepted:** a future legitimately tracked symlink under `internal/**` or `cmd/**`
  fails the gate. That is the intended fail-closed shape; the operator removes the link or amends
  the contract.

### D-WB2-3: EEE39D23 case-variant directory scope: OUT OF SCOPE, documented and pinned

The entry asked this to be decided. **Decision:** a tracked path whose first component is a case
variant of `internal` or `cmd` (`Internal/`, `CMD/`) is **not** brought into the gate's scope.

* On a case-sensitive filesystem (the required Linux gate) it is a different directory, exactly
  like `pkg/`, `tools/` or a root-level Go file. The gate's scope has always been the two product
  directories, and `tools/gatecheck` itself legitimately contains write primitives. Treating one
  spelling of an unscanned directory as special while `pkg/` stays unscanned is inconsistent.
* A `:(icase)` pathspec or a case-fold refusal would change the pathspec contract pinned by the
  wiring and by retiredarch-style freezes, and add a refusal for a layout the repo does not have.
* **Residual (accepted, operator checkpoint 2):** module Go code outside `internal/**` and `cmd/**`
  is outside the gate by design. A test row pins that a case-variant record is neither selected nor
  an error, so any future widening is a deliberate change. If the operator wants module-wide scope
  that is a separate deliberation, not a defect of this gate.

### D-WB2-4: DBE25DF5, anchor the wrapper root to the script location through one scrubbed shared function

| Option | Verdict |
|---|---|
| **A. Add `gatecheck_repo_root` to `scripts/lib/gatecheck-run.sh`: run `git -C "${GATECHECK_SRC}" rev-parse --show-toplevel` with every `GIT_*` variable unset except `GIT_CEILING_DIRECTORIES`; both wrappers use it (chosen)** | Removes both ambient redirection (scrubbed) and the cwd dependency (anchored), in one chokepoint instead of two copies. Keeps the path form (`C:/...` under Git Bash) identical to today's, which a git-free `ROOT=${GATECHECK_SRC}` would change to MSYS `/c/...`. Mirrors the Go `gitRunnerEnv` rule so the two surfaces cannot drift in intent. |
| B. Scrub only `GIT_DIR` and `GIT_WORK_TREE` (`env -u`) | Rejected. A denylist of two names misses `GIT_COMMON_DIR`, `GIT_INDEX_FILE` and friends; the Go side already moved from a denylist to prefix-stripping for that reason (D7BF9F74). |
| C. `ROOT=${GATECHECK_SRC}` with no git call | Rejected as primary. Windows path-form change (`/c/...` vs `C:/...`) rides on MSYS argument conversion; the compound note on MSYS bridging shows that is a false-pass source. Kept as a fallback if A's shell scrub proves unportable. |
| D. Go-side identity assertion (`git rev-parse --show-prefix` empty, scrubbed) | Rejected as YAGNI. After A, the root is the script's own checkout and the child git is already scrubbed, so a second check adds a failure mode without closing a demonstrated path. |

* **Behaviour change (operator checkpoint 3):** the write-path wrapper currently anchors on the
  caller's cwd; after A it anchors on the script's own checkout, which is what the gate already
  builds (`gatecheck_build` uses `GATECHECK_SRC`), so the engine and the scanned tree now always
  come from the same checkout. retiredarch has anchored this way since M2-T11. Each wrapper keeps
  its own caller-cwd `git rev-parse --show-toplevel >/dev/null` precondition (exit 128 outside a
  work tree) as a separate statement, as the retiredarch wrapper already does.
* **Wrong-tree coverage (AS-5):** tests use two repositories that **both** contain `internal/` and
  `cmd/` files. The oracle is the `--root` value the wrapper hands the built binary (shim-recorded),
  compared by file identity, under ambient `GIT_DIR`+`GIT_WORK_TREE`, under `GIT_DIR` alone from a
  subdirectory cwd, and with no ambient variables (control). The red evidence is the parent
  commit's wrapper printing the decoy root.

### D-WB2-5: DBE25DF5, the per-arm ED-7 guard is REJECTED as the fix

The Architecture Strategist's AS-5 caution is adopted: a "selection must contain at least one
`internal/` file and one `cmd/` file" guard is a non-vacuity heuristic, not proof of root identity
or listing completeness. A different clean repository that has both arms, or a partial listing
that keeps one file from each, still passes. It also adds a contract change (a repo with no `cmd/`
Go file would start failing). After D-WB2-4 the selected root is the script's own checkout and the
child git is isolated, so the partial-listing vectors that remain are not environmental (a sparse
checkout is local repository state and already fails closed on the missing file read). The ED-7
empty-selection guard is left exactly as it is.

### D-WB2-6: EDC18D59, one local escaping chokepoint, applied at the sinks, not per call

| Option | Verdict |
|---|---|
| **A. A package-local `escapeLogText(string) string` in `writepath.go`; `errorLine` and the fixture-name prints call it; unwrap `*fs.PathError` at the two `scanFile` call sites (chosen)** | One rule. Every non-printable rune (`!strconv.IsPrint`: C0 and C1 controls, DEL, U+0085, U+2028/2029, bidi and other format characters) becomes a visible `\xNN`, `\uNNNN` or `\UNNNNNNNN` escape; an invalid UTF-8 byte becomes `\xNN`. Backslash and printable text, including non-ASCII, are left unchanged, so goldens and every control-free message are byte-identical. |
| B. Per-sink `%q` | Rejected. `%q` also quotes and escapes printable text, changing goldens, and it leaves the next new sink unprotected. |
| C. A shared helper in `pysem` or a new package | Rejected for now. The repo's pattern is "copied, not shared, each gate package keeps its own surface" (D-BA-3, 5A8EC1BC). Extraction is the right move when a second engine adopts it, which is captured (D-WB2-9). |
| D. Bounded rune set (C0, DEL, U+0085, U+2028/2029) | Rejected in favour of `!strconv.IsPrint`: it covers C1 controls such as U+009B (CSI) and bidi overrides that the bounded set misses, with no larger code. This is the entry's own optional note. |

* **Scope of the chokepoint:** `errorLine` (context and error text, which also covers go/parser
  multi-line literal tokens and the raw multi-line git stderr that `DefaultGitRunner` appends), and
  the three fixture-name prints in `runFixtureSelfTest` (PASS, FAIL, and the joined findings). The
  U4 refusal line is covered because it goes through `errorLine`.
* **Test seam:** hostile fixture names cannot be created on Windows, so filesystem rows live in a
  `//go:build !windows` test file (no `t.Skip`); the helper and `errorLine` rows are pure and run
  everywhere. The U4 rows that pin the raw refusal text are updated in the same task that makes
  the escape live.
* **The pre-existing prefix** `::error::` produced by the tool itself is intentional and is not
  escaped.

### D-WB2-7: 858516B9, refuse `.go`-named directories and add a FIFO row; hard-link refusal NOT implemented

* **Directory rule:** in `runFixtureSelfTest`, a real directory whose name ends in `.go` fails
  closed (`escape`d name, one new message) instead of being skipped. Directories without a `.go`
  suffix, notably `harness/`, stay skipped. This is a shape rule (a `.go` name is a fixture-class
  name, so it must be a regular file), not a named-spelling denylist.
* **FIFO row:** added in a `//go:build !windows` file, so there is **no new `t.Skip` and no Windows
  skip**: Windows simply does not have the row, and the required Linux job runs it. The row runs
  the scan in a goroutine with a timeout so a regression fails fast instead of hanging on a FIFO
  read. It is a characterization row (the refusal already exists), so it is green on arrival and
  recorded as such.
* **Hard link: decided NOT to implement** (operator checkpoint 4, override possible):
  1. There is no stdlib portable link-count API; it needs a build-tag pair
     (`syscall.Stat_t` on unix, a Win32 handle call on Windows), which widens the surface.
  2. A link count above 1 does not prove the other link is outside the root, and overlay or
     cache-restored CI workspaces can report counts above 1 for ordinary files, so a blocking gate
     would flake.
  3. Creating a hard link requires write access to the checkout, the in-job attacker class the
     gate documents as out of scope (R-A2b, an anti-accident control). Git cannot track a hard link.
  The remainder is recorded in a new low entry (D-WB2-9) so it is not lost, and the code comment
  names the accepted residual.

### D-WB2-8: 2613D248, strip `MSYS`, `CYGWIN` and `MSYS2_*` identically in both packages, with an atomic retiredarch refreeze

* **Rule:** in addition to the `GIT_*` rule, drop an entry whose ASCII-case-folded name is exactly
  `MSYS` or `CYGWIN`, or has the prefix `MSYS2_`. `MSYSTEM` and similarly spelled names
  (`MSYSX`, `CYGWINX`) are **kept** (not in the entry; rows pin the exact boundary).
* **Rows (identical in both `TestGitRunnerEnv` tables):** MSYS / CYGWIN / `MSYS2_ARG_CONV_EXCL`
  stripped in mixed case; `MSYSTEM`, `MSYSX`, `MSYS2` (no underscore) kept; the entry's three
  boundary rows: bare `GIT_DIR` with no `=`, exact `GIT_`, bare `GIT_CEILING_DIRECTORIES`.
  The three boundary rows are green on arrival (characterization); the MSYS rows are the red.
* **Atomicity (memory fact, verified):** the retiredarch change is **one task, one commit**:
  `select.go` `gitRunnerEnv` and `pin.go` `canonicalDecls` (a hand-authored copy, never derived).
  `closedWorldDecls` and the two frozen name sets name declarations, not their text, so they do not
  change; the plan states this and requires `git show --stat` to confirm both files in one commit.
  The `GIT_CONFIG_*` literals that `pin_test.go` mutates are untouched.
* **Rejected:** a shared helper (5A8EC1BC trigger is a fourth engine; not fired) and stripping
  `MSYSTEM` (not named by the entry, and Git Bash itself reads it).

### D-WB2-9: New out-of-scope discoveries become DEFERRED SCOPE EXPANSION entries (Stage own-authority intake, P-021)

1. **`36BE1F61` Retiredarch CI-log echo sinks (low, task):** the twin of EDC18D59 in `retiredarch.go`
   (`::error::%v` and `::error::%s: %v` at ~384, 389, 536, 596, 695).
2. **`1B85D0F4` Other wrappers' unscrubbed cwd root discovery (medium, bug):** `check-unignore-regression.sh:106`,
   `check-merge-strategy.sh:80`, `check-depguard-fixtures.sh:41`, the same false-clean class as
   `DBE25DF5`, now fixable by calling the shared `gatecheck_repo_root` that S2 adds.
3. **`A8FB0682` Fixture hard-link refusal (low, bug):** the part of `858516B9` decided against in D-WB2-7,
   kept as an open, unscheduled item with the rationale.

### D-WB2-10: Interaction with shipment 043-S (Batch 1)

No file overlap. 043-S touches `.github/` instructions and skills, `.autoharness/` config and
manifest, policy docs, and Go contract tests under `tests/integration/`
(`ship_feature_completion_contract_test.go`, `stage_branch_gate_test.go`,
`shipment_close_path_behavior_test.go`, `shipment_close_path_harness_test.go`, one new smoke
test). Batch 2 touches `tools/gatecheck/**` and `scripts/**` only (verified against the 043-S plan:
its only `scripts/**` mention is a zero-retired-token sweep of live surfaces, a read, not an edit). The only shared surface is the global gate: Batch 2's tasks must leave
`go test ./...` green on top of 043-S, and 043-S must be merged and closed first because its
closure procedure is what Ship uses to close S1 and S2.

## Operator checkpoints (decisions Stage made on judgment; each can be overridden at staging-PR review)

1. **Two shipments, not one** (D-WB2-1), chain `043-S → S1 → S2`.
2. **Case-variant directories are out of the gate's scope** (D-WB2-3); no `:(icase)` pathspec.
3. **Write-path wrapper root moves from caller-cwd to script-location anchoring** (D-WB2-4).
4. **Hard-link refusal is not implemented** (D-WB2-7); recorded as a deferred entry.
5. **The per-arm ED-7 guard is rejected** (D-WB2-5).
6. **Stacked base:** this branch carries the Batch 1 tip; the Batch 1 staging PR must merge first as
   a merge commit.

## Newly discovered scope

See D-WB2-9; entries `36BE1F61`, `1B85D0F4` and `A8FB0682` were created after the Step 1.9 gate with
full P-021 C2 fields and are left active (not consumed by this batch).

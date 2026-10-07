---
title: "Harden retiredarch gate integrity — C5 retiredarch security/correctness batch"
description: "Stage deliberation over six DEFERRED SCOPE EXPANSION stash entries against tools/gatecheck/internal/retiredarch: package closure side effects, Windows env mutation, git runner environment isolation, dispatch and wiring pins, symlink containment, and the walkTable ancestor-reopening desync"
topic: "retiredarch security/correctness batch (stash 3750C37C, 9FC28DB9, 0ECC1895, DC921AF6, D7BF9F74, 990AFA71)"
depth: "deep"
decision_status: "decided"
promoted_to: "plan"
linked_artifacts:
  - "docs/plans/2026-10-07-intercom-go-retiredarch-gate-integrity-plan.md"  # superseded (D-RA-7)
  - "docs/plans/2026-10-07-intercom-go-retiredarch-correctness-plan.md"     # Shipment 1 (D-RA-7)
  - "docs/compound/2026-10-07-pin-canonical-text-maintenance-and-unbounded-ast-rule-review.md"
tags:
  - "retiredarch"
  - "gatecheck"
  - "security"
  - "p-021"
  - "stage"
---

# Deliberation — Harden retiredarch gate integrity

## Session context

* **Agent / mode:** Stage, invoked by the Orchestrator (`stage next`), sequential, non-dark.
* **Operator approval:** relayed by the Orchestrator at 2026-10-07T13:55-07:00, "let's go with
  the recommendation". Stage applies it to the batch selection (the grouping below) and to the
  Stage-recommended option in each decision. Any override can be made at staging-PR review,
  before Ship claims the shipment (see "Operator checkpoint").
* **Standing operator policy:** reliability and security before features.
* **Base:** `main` @ `372ab38`. Stage artifact branch `chore/stage-harden-retiredarch-gate-integrity`.
* **Learnings retrieval (Step 1.8):** confidence medium. Applied learnings:
  `docs/compound/2026-10-01-textual-yaml-guard-denylist-bypass-loop.md` (allowlist, not
  denylist), `docs/compound/2026-09-28-extracting-bash-heredoc-python-engines-into-importable-modules.md`
  (init-time side effects include package-level initializers),
  `docs/compound/2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md`
  (self-matching bans), `docs/compound/2026-09-08-verify-go-stdlib-internals-claims-against-goroot-source.md`
  and `docs/compound/2026-09-08-adversarial-review-empirical-verification-and-scope-check.md`
  (Windows junctions vs. lexical containment), `docs/compound/2026-09-30-go-msys-bash-bridging-and-unicode-case-mapping-gap.md`
  (Windows test hygiene), and `docs/compound/workflow-issues/staging-premise-verification-before-planning-2026-09-18.md`
  (every AC must fail at its parent commit).

## Step 1 — Triage and P-021 obligations

All six entries carry the literal `DEFERRED SCOPE EXPANSION` marker, so the Step 1 precedence
rule forced the `deliberate` route for each, regardless of shape (P-021 C6). All six are
task-shaped (`task`/`bug`), and all target `tools/gatecheck/internal/retiredarch`.

| Stash | Pri | Kind | Residual | One-line summary | Verified against code at `372ab38` |
|---|---|---|---|---|---|
| `3750C37C` | high | task | closure gap | Package closure does not constrain sibling-file side effects (an `init` that runs `git update-index --force-remove`) | Yes: `closureFileOK` (pin.go:355) checks names, imports of `unsafe`/`C`, linkname and `envMutators` only; no `init`, `os/exec` or write rule |
| `0ECC1895` | med | task | closure gap | SEC4-3 env rule misses Windows `syscall.SetEnvironmentVariable` / `NewLazyDLL` | Yes: `envMutators` (pin.go:226) is `Setenv/Unsetenv/Clearenv` only; no file imports `syscall` or `golang.org/x/sys` today |
| `D7BF9F74` | med | task | R-A2 | `DefaultGitRunner` inherits process git env/config (pathspec reinterpretation, `GIT_DIR`/`GIT_INDEX_FILE`/`GIT_CONFIG_*` redirection, PATH, `pysem` integrity) | Yes: select.go:32 sets only `cmd.Dir`; no `cmd.Env` |
| `DC921AF6` | med | task | R-A1 | Runner wiring (`runRepoScan`, `Run`, `runRetiredArch`) and dispatch (`scanPath`, `engineForPath`) are unpinned | Yes: neither is in `pathspecFrozenDecls`/`prefixFrozenDecls`; `scanPath`/`engineForPath` are closed-world names but not token-frozen |
| `990AFA71` | med | task | containment | Tracked symlink under `cmd/**`/`internal/**` is followed after a lexical join | Yes: `runRepoScan` (retiredarch.go:590) does `filepath.Join(root, FromSlash(rel))` with no `Lstat`. No tracked symlink exists today (`git ls-files -s` mode `120000`: none) |
| `9FC28DB9` | high | bug | port divergence | `walkTable` desyncs (fails closed) on ancestor-table reopening into a deeper incomplete descendant | Yes (by code reading): `walkTable` (tomlprimary.go:148) only tolerates an exact self-entry for its own prefix; an ancestor's self-entry while a descendant frame is open returns the desync error |

### (A) Unconditional duplicate scan

Ran over all 62 active stash entries with the patterns `retiredarch`, `retired.arch`,
`walkTable`, `tomlprimary`, `scanPath`, `engineForPath`, `selectRepoPaths`, `DefaultGitRunner`,
`GitRunner`, `closure`, `envMutator`, `Setenv`, `SetEnvironmentVariable`, `NewLazyDLL`,
`symlink`, `GIT_INDEX_FILE`, `GIT_DIR`, `pysem`, `init()`, `update-index`, `pin.go`, `pathspec`.

**Result: CLEAN for all six.** No entry describes the same expansion. After plan review
(Constitution Reviewer P3), the scan was re-run with these additional terms:
`runRepoScan`, `runRetiredArch`, `register_retired_arch`, `Lstat`, `junction`, `desync`,
`ancestor`, `GIT_CONFIG`, `safe.directory`, `R-A1`, `R-A2`, `gitShowToplevel`,
`registerSubcommand`, `initstate`, `execerrdot`. The only hits were the in-batch entries
(`9FC28DB9`, `DC921AF6`, `D7BF9F74`, `0ECC1895`), so the result is still **CLEAN**. Adjacent
entries were checked and are different expansions:

* `5A8EC1BC`: consolidate three GitRunner shapes (maintainability). Soft coupling with
  `D7BF9F74` only, see "Out-of-scope dependency check".
* `FE2F02FF`: write-path dynamic proc invocation (writepath package, not retiredarch).
* `C312BD4C`: gomask struct-tag regex/comment (gomask package).
* `9FF9EEB4`: fixture-manifest non-string parity (retiredarch self-test only, low).
* `47F54477`: plan-doc prose. `295C9148`: writepath/retiredarch coupling comment.

None was merged or archived as a duplicate.

### (B) Late-identifier reconciliation

| Stash | N/A fields | Residual-risk record searched | Outcome |
|---|---|---|---|
| `3750C37C` | PR, review-thread | `docs/closure/030-S-033-F-post-merge-closure.md` (lines 93-110, 248) | **Recovered:** PR **#95** (030-S / 033-F, merge `e1d61a7`); the PR body recorded it as a follow-up. Review-thread `N/A` **stands** (pre-PR local adversarial review, no GitHub thread) |
| `0ECC1895` | PR, review-thread | same record (lines 98-110, 250) | **Recovered:** PR **#95**. Review-thread `N/A` stands (local review) |
| `DC921AF6` | PR, review-thread | same record (line 254: "already captured before this shipment") | **Recovered (carrier only):** PR **#95** carried the residual. Origin stays Stage plan-review round 4 (D-030-6); review-thread `N/A` stands |
| `D7BF9F74` | PR, review-thread | same record (line 254) | **Recovered (carrier only):** PR **#95**, as above |
| `990AFA71` | none literally `N/A`; review-thread recorded as "captured-separately-see-graphql" | `docs/closure/035-S-045-F-post-merge-closure.md` (lines 87-89, 297); then GitHub GraphQL `reviewThreads` of PR #83 (authoritative, read-only), on plan-review advice | **Recovered:** review-thread **`PRRT_kwDOTPuhps6nd8rH`** (contains comment 4143002496). No Ship residual-risk record carried it; it came from the GitHub API |
| `9FC28DB9` | none | n/a | Not triggered. Refs complete: PR #83, comment 4146023143, thread `PRRT_kwDOTPuhps6nlaid` |

The stash-edit operation is MCP-only in this workspace (`stash_edit` has no `cli_command`), so
the recovered identifiers are recorded here, alongside the originally captured refs, rather
than written back into the entries. All six are consumed by this session and archived at Step
5.6 with a forward reference to the feature.

### Out-of-scope dependency check

The Orchestrator named four nearby entries that are out of scope unless a hard dependency is
proved. **No hard dependency exists**, so scope was not widened and the session did not halt.

* `C312BD4C` (gomask): `scanGo` consumes gomask, but no decision here changes masking. None.
* `9FF9EEB4` (fixture-manifest parity): touches `loadFixtureManifest` in retiredarch.go, which
  no unit here freezes. After this batch lands, any future fix that adds an import or an `os`
  call to retiredarch.go must also update the new package manifest (D-RA-6). That is a
  coupling for the future fixer, not a dependency of this batch.
* `47F54477` (plan-doc prose): none.
* `5A8EC1BC` (shared GitRunner): `D7BF9F74` hardens only `retiredarch.DefaultGitRunner`, in
  place. Consolidation is not needed to do that. A later consolidation must preserve the
  isolation added here. Recorded as a soft coupling for the 5A8EC1BC re-triage.

## Problem frame

The retired-architecture gate (`scripts/check-retired-architecture.sh` → `tools/gatecheck`
`retired-arch` → `retiredarch.Run`) is a merge-blocking CI gate. Its integrity rests on a
source-text pin (`pin.go`, `SelectionPathspecPin`) that token-freezes the scan-scope surface of
`select.go` and applies package-closure rules. Rev 6 (030-S / 033-F) disclosed residuals R-A1
to R-A4. Two later local reviews added two closure gaps. The 035-S port left one disclosed
TOML divergence and one parity-matched containment gap.

**Threat model (unchanged from rev 6).** The pin defends against an in-repo PR that narrows the
effective scan scope without touching the frozen declarations: accidental, or a
low-sophistication adversarial edit inside the retiredarch package. R-A4 stays out of model:
cross-package linkname, reflect/unsafe elsewhere, binary patching, and a coordinated `pin.go` +
target edit. Human review of any `pin.go` diff covers R-A4.

**Success criteria.**

1. Every in-package mechanism named by the six entries fails closed in the self-test, or is
   neutralised at runtime.
2. The current tree's gate verdict is unchanged (INV-1). Every new rejection is an explicitly
   enumerated delta (ED-10..ED-14 in the plan).
3. Every new rule is an **allowlist** with its own reject-table test, and every acceptance
   criterion fails at its parent commit.
4. Residuals that remain are re-stated precisely in `pin.go`'s residual block.

**Out of scope.** Symlink containment for the writepath, unignore and mergestrategy engines
(captured as a new low-priority stash follow-up). GitRunner consolidation (`5A8EC1BC`). Pinning
the `pysem` package. Pinning an absolute git binary path. Any change to scan scope, forbidden
tokens, or engine matching.

## Decisions

### D-RA-1 — `9FC28DB9`: desync fallback to a complete deterministic walk, plus a completeness oracle

**Options.**

* **A. Stack-based suspend/resume cursor state machine** (the entry's suggestion). It keeps
  exact Python document order in every shape. Effort high, `complexity: high`, and the entry
  shows that a rushed cursor change regressed into a silent skip.
* **B. Ordered trie rebuilt from `MetaData.Keys()`**, with array-of-tables element
  segmentation. Effort medium-high. It still depends on unverified `Keys()` shape claims for
  arrays of tables.
* **C. Keep the cursor walk; on desync, fall back to a complete, deterministic, order-insensitive
  walk** of the decoded tree (keys sorted per table, arrays in index order). Separately, on the
  **success** path, compare the cursor walk's finding *multiset* with the order-insensitive
  walk's and fail closed on any mismatch. Effort low-medium.

**Decision: C.** The verdict becomes correct in every shape. Findings are complete by
construction, because the fallback walks the decoded tree, not the cursor. The oracle turns any
future cursor bug, including the F-1 class, from a silent skip into a fail-closed error.
Existing goldens are unchanged, because the fallback triggers only where today's output is the
generic desync parse-error finding.

**Delta.** Finding *order* in a desync shape is lexicographic depth-first, not Python document
order. Python is retired (M4), so only the Go goldens bind. Enumerated as ED-10.

**Rejected.** A: disproportionate complexity and a demonstrated regression risk. B: still
order-coupled, without the oracle's safety property.

### D-RA-2 — `990AFA71`: retiredarch-local component-wise `Lstat` containment

**Options.**

* **A. Local check in retiredarch.** Before scanning each selected path, `os.Lstat` every
  component from `root` down to the file. Require each intermediate component to be a plain
  directory (`ModeDir`, no `ModeSymlink`/`ModeIrregular`) and the final component to be a
  regular file. Any violation, or any `Lstat` error, becomes a fail-closed synthetic finding.
* **B. Reuse `internal/pathsafe`** (`Root.Resolve`). That couples a CI gate to a product package
  whose verification features 038-F and 039-F are **blocked**, and must not be touched.
* **C. Defer everything to a cross-engine hardening unit.**

**Decision: A.** It is small and dependency-free, and it fails closed. On Go ≥ 1.23, Windows
mount points (junctions) report `ModeIrregular`, not `ModeSymlink`, so the
`ModeSymlink|ModeIrregular` mask covers junctions. The implementer must verify this against the
pinned GOROOT source, per the learnings. There are no tracked symlinks today, so the current
verdict is unchanged (ED-11 covers the new rejection). The cross-engine remainder (writepath,
unignore, mergestrategy) is captured as a new low-priority stash follow-up instead of widening
this batch.

### D-RA-3 — `D7BF9F74`: isolate `DefaultGitRunner`'s environment; accept PATH, repo config and `pysem` as residuals

**Options.**

* **A. Environment isolation.** Build `cmd.Env` from `os.Environ()`:
  * drop every variable whose name starts with `GIT_` (case-insensitive on Windows), except
    an allowlist of exactly `GIT_CEILING_DIRECTORIES`, which only restricts repository
    discovery and is relied on by `check_retired_architecture_wrapper_test.go`;
  * then append `GIT_CONFIG_NOSYSTEM=1`, `GIT_CONFIG_GLOBAL=<os.DevNull>` and
    `GIT_CONFIG_SYSTEM=<os.DevNull>`.

  This neutralises the `*_PATHSPECS` reinterpretation, the
  `GIT_DIR`/`GIT_WORK_TREE`/`GIT_INDEX_FILE`/`GIT_COMMON_DIR`/`GIT_OBJECT_DIRECTORY` redirection,
  and `GIT_CONFIG_COUNT`/`GIT_CONFIG_PARAMETERS` injection. It also removes global/system
  `core.fsmonitor`-style hooks. It mirrors the established test-isolation set
  (`internal/unignore/testutil_test.go:44-78`, `ciwiring_retire_test.go:313-354`).
* **B. A + `--literal-pathspecs`.** Rejected: the frozen pathspecs `cmd/**`/`internal/**`
  *require* glob semantics, so literal mode would select nothing.
* **C. A + absolute git path.** Rejected: there is no trusted absolute location across the
  Ubuntu and Windows runners, and Go's `exec.LookPath` already refuses relative-PATH (`ErrDot`)
  resolution. Recorded as residual R-A2a.

**Decision: A.** `DefaultGitRunner` and select.go's import block are token-frozen in
`canonicalDecls`, so the change must land atomically with the `pin.go` canonical text (atomic
landing pair ALP-2, as ALP-1 did in 030-S).

**Accepted residuals (re-stated in `pin.go`):**

* R-A2a: PATH resolution of `git`.
* R-A2b: repository-local `.git/config` (part of the checkout under test).
* R-A2c: `pysem.GitText`/`SplitLines` integrity. That is a shared package with its own tests;
  pinning it from retiredarch would create cross-package coupling.

**Fail-closed consequence.** If a runner needs `safe.directory` from global config (container
ownership mismatch), git exits non-zero and the gate fails **closed** with `::error::git
ls-files`. CI runs on `ubuntu-latest` without a container (`ci.yml:294-332`), so this does not
apply today. The rollback trigger covers it.

### D-RA-4 — `DC921AF6`: token-freeze the dispatch and wiring chain

**Options.**

* **A. Extend the frozen-declaration pin across the whole chain:** `runRetiredArch`
  (tools/gatecheck/register_retired_arch.go) → `Run` → `runRepoScan` + containment helper
  (retiredarch.go) → `scanPath`/`engineForPath` (select.go).
* **B. Behavioural canary self-test.** A path-aware sabotage can detect and spare the canary,
  so this is weaker.
* **C. Accept R-A1.**

**Decision: A, in three increments.** Freeze `engineForPath` + `scanPath` (select.go). Freeze
`runRepoScan` + the D-RA-2 helper (retiredarch.go). Freeze `Run` + `runRetiredArch`.

Each increment adds independently authored canonical text in `pin.go` (H-11 independence) and a
mutation reject row per frozen declaration. Engine internals (`scanGo`, `scanTomlPrimary`) stay
covered by the fixture self-test and goldens. Path-conditional sabotage inside an engine
remains an R-A4-class residual.

### D-RA-5 — `3750C37C` + `0ECC1895`: closed-world package manifest (file + import allowlist)

**Options (the entry's own fork).**

* **A. Allowlist.** `pin.go` carries an independently authored manifest of the exact non-test
  `.go` file names in the package, and, per file, the exact import-path set. An unknown file, a
  missing file, or any import outside the file's set fails closed.
* **B. Token-freeze every non-test file.** Rejected: every legitimate edit becomes a `pin.go`
  edit, which is brittle with no added protection over A + D-RA-6.
* **C. Human review only.** Rejected: that is the status quo the entries flag.

**Decision: A.** `syscall`, `golang.org/x/sys/...`, `unsafe`, `C`, `plugin`, `reflect` and
`os/exec` (outside `select.go` and `pin.go`) are rejected by construction, because no file's
set contains them. That closes `0ECC1895` without a selector denylist. A new sibling file
(`3750C37C`'s example) is rejected by name.

### D-RA-6 — `3750C37C`: init-time side-effect and `os`/`exec` use allowlist

**Decision.** In every non-test package file:

1. Reject any `func init`.
2. Reject any package-level `var` initializer containing a call expression unless its
   `(file, var)` pair is in an authored allowlist. The candidate set, to be confirmed by
   characterization, is `vocabWords` (ident.go), `goIdentifierRe` (scango.go) and `bareKeyRe`
   (tomlfallback.go).
3. Allow `os.*` and `exec.*` selectors only from a per-file selector allowlist. Today that is
   `pin.go {os.ReadDir, os.ReadFile, exec.Command}`, `retiredarch.go {os.ReadFile}` and
   `select.go {exec.Command}`. It is extended by exactly the selectors that D-RA-2 (`os.Lstat`)
   and D-RA-3 (`os.Environ`, `os.DevNull`) add.

Process spawning, index writes and env mutation through `os`/`exec` then fail closed in any
file. The ban lists must match by AST shape (import spec, `SelectorExpr`, `FuncDecl`), never by
raw text, so the pin does not match its own declarations.

## Grouping (Step 1.5)

Six task-shaped entries, all in one package, sharing one trust root (`pin.go`) and one gate.
Splitting them would serialise several `pin.go`/golden histories across PRs.

**Selected grouping:** all six, under the covering feature "Harden retiredarch gate integrity".
**Estimated scope:** 14 tasks × 2 h, after plan-review revisions 2 and 3. **Risk:** moderate. The
changes are internal CI tooling and only tighten the verdict; there are no product surfaces.

## Addenda after plan review, attempt 1 (Stage, same session)

The FAIL verdict from plan-review attempt 1 raised P1 findings that extend three decisions.
Each extension stays inside the six entries' stated scope and inside
`tools/gatecheck/internal/retiredarch`. Package `main` is read, never edited.

* **D-RA-3 addendum.** The pin's own root discovery (`gitShowToplevel`, pin.go:706) gets the
  same env isolation as `DefaultGitRunner`. Both runners also fail closed when `cmd.Path` is
  not absolute or lies inside the root. That neutralises an in-repo `git` reached through
  `godebug execerrdot=0` or a PATH entry inside the checkout. R-A2b is re-scoped to "local
  repository state mutated by earlier same-job or same-process code".
* **D-RA-4 addendum.** The freeze extends through the self-test orchestration chain
  (`runSelfTestAssertions`, `runRepoSelectionSelfTest`). That chain is meaningful to freeze
  because `go test`'s live-tree pin test evaluates the pin independently of `Run`. Package
  `main` also gets a single-literal-registration rule: exactly one `registerSubcommand` call
  registers `"retired-arch"`, and every call uses a string literal.
  Residual **R-A1c** covers `main.go` dispatch internals, the wrapper script and `ci.yml`
  steps. Existing wrapper and CI wiring tests and human review cover it.
* **D-RA-6 addendum.** The 3750C37C vector also exists one directory over: an init-time spawn
  in any package linked into gatecheck. Pinning other packages from retiredarch's runtime pin
  would couple packages. Instead, the existing go-test-level init scan
  (`initstate_test.go`, which already covers every gatecheck package) is extended to process
  spawns and env writes (RA-13 in revision 2, RA-14 in revision 3). Residual **R-A5** covers linked-package and
  module-resolution integrity: `pysem`, `gomask`, `go.mod` `replace`/vendor, and `godebug`.

## Addenda after plan review, attempt 2 (Stage, same session)

* **D-RA-4 addendum 2.** The name-based single-registration rule was bypassable through an
  alias (`m := subcommands`), through `maps.Copy`, or through a function value bound to
  `registerSubcommand` (a P1 from Architecture and Security). It is replaced by a **positive
  every-reference rule** over package `main`'s non-test files:
  * Outside main.go, `subcommands` must not be referenced at all.
  * `registerSubcommand` may appear only as the callee of a direct call with a string-literal
    first argument, in a top-level `func init` body.
  * main.go's `subcommands` and `registerSubcommand` declarations are token-frozen, and its
    only other permitted use is the comma-ok lookup.
  * The rule constrains **every** subcommand's registration shape, not only `retired-arch`'s.
    The operator may override this at staging-PR review.
* **D-RA-3 addendum 2.** The pin root is resolved once, through the guarded
  `gitShowToplevel`, and passed explicitly to the evaluator, so the evaluator never calls git
  and test fixtures need no git repository. The `cmd.Path` exec-location guard uses
  `filepath.Abs` + `EvalSymlinks` canonicalisation, with a case-insensitive compare on
  Windows. For `gitShowToplevel`, the guard is anchored *after* the run against the returned
  top level. Any canonicalisation error fails closed.
* **D-RA-6 addendum 2 — operator-overridable widening.** A denylist of syscall and x/sys
  functions is bypassable through `LazyProc.Call` and `SyscallN`. RA-14 therefore bans the
  `syscall`, `golang.org/x/sys/...`, `unsafe`, `plugin` and `C` imports, plus `//go:linkname`
  directives, **at package level in every gatecheck package**. This is a test-level rule in
  `initstate_test.go`, and no current file violates it (verified at `372ab38`).
  * It widens the entry's scope beyond the retiredarch package's runtime surface. The
    justification is that `3750C37C`'s vector, an init-time spawn, reaches through any
    package linked into the gatecheck binary.
  * The operator may narrow RA-14 to the init-reachable watch-list only, or drop it, at
    staging-PR review. Doing so leaves R-A5 broader.
  * The former single RA-12 is split into RA-12 (init rules) and RA-13 (selector allowlist)
    to satisfy the 2-hour rule.

## Risks and mitigations

* **The pin is self-matching** (bans matching their own declarations). Use AST-shape matching
  and run the real tree before and after each change (INV-1 check).
* **Allowlist churn.** Future legitimate edits to retiredarch need a `pin.go` manifest update.
  That is the intended human-review choke point, documented in `pin.go`'s header.
* **Junction semantics claim** (D-RA-2). Characterization-first; verify against GOROOT; skip
  Windows symlink tests honestly when privileges are absent.
* **Global-config removal** (D-RA-3). Fails closed, and the rollback is a one-commit revert of
  ALP-2.
* **Oracle false positives** (D-RA-1). The multiset comparison must use identical finding text,
  and duplicate findings across array elements must be counted, not set-collapsed.

## Unresolved questions

None blocking. The `(file, var)` initializer allowlist and per-file selector sets are confirmed
by characterization at implementation time. A divergence from the candidate sets above is a
HALT-and-return-to-Stage condition if it would add a write-, exec- or env-capable selector.

## D-RA-7 — Escalation disposition: Option B, split the batch (operator/Orchestrator, 2026-10-07 ~14:30-07:00)

**Provenance.** Plan review of the 14-unit plan (rev 3, RA-1..RA-14) failed three times, and
Stage escalated and halted (see the Stage memory file). The Orchestrator, acting under
operator delegation in autopilot, relayed the operator's instruction "let's go with the
recommendation". Stage's recommended option was **B (split the batch)**, so B is the
disposition. This is the **operator/Orchestrator disposition of the escalation**. It is
**not** a fourth review cycle on the old plan. The rev-3 plan is marked `superseded` and kept
unchanged below its banner, for history.

**Shipment 1, planned and harvested now.** Stash `9FC28DB9` (TOML desync), `990AFA71`
(symlink containment) and `D7BF9F74` (git environment isolation). The new plan is
`docs/plans/2026-10-07-intercom-go-retiredarch-correctness-plan.md`, a new review unit with a
fresh 3-attempt cap. It carries D-RA-1, D-RA-2 and D-RA-3 with these changes:

* **D-RA-3 narrowed.** The inside-checkout exec-location guard (`gitPathInside`:
  `filepath.Abs`/`EvalSymlinks` canonicalisation) is **dropped**. It caused all three
  attempt-3 P1s on this surface:
  * A3-P1-1: missing from the closed-world refreeze;
  * A3-P1-2: `filepath.Abs` is reachable from the registered `runRetiredArch` and turns
    `TestNoInitTimeProcessStateReads` red;
  * G3-P1-1: a cross-volume `filepath.Rel` error on `windows-latest`.

  What remains is env isolation, plus `cmd.Err != nil` and `!filepath.IsAbs(cmd.Path)`
  fail-closed checks. Those use no process-state reads.
* **Accepted residual: an in-checkout `git` binary (folded into R-A2a).** A `git` executable
  inside the checkout, reached through an *absolute* PATH entry that names a checkout
  directory, is not refused. Rationale:
  * The PATH of the CI job is set by the workflow, not by the checkout. A PR that can put a
    checkout directory on PATH can already edit `.github/workflows/**` or
    `scripts/check-retired-architecture.sh`, so that vector is a self-modifying-PR vector.
    That is the CODEOWNERS/branch-protection gap below, not a runner-level gap.
  * Go's `exec.LookPath` already refuses the relative/current-directory resolution (`ErrDot`).
    The `IsAbs` check covers `GODEBUG=execerrdot=0`.
  * A correct inside-root test needs path canonicalisation, which conflicts with the
    init-time process-state test, and cross-volume handling on Windows. That is complexity
    that three review cycles could not make sound.
* **9FC28DB9 correction.** The entry is *not* a false-clean bug. Per its own text, today's
  behaviour is a **false-positive fail-closed** generic "TOML parse error". The false-clean
  risk belongs to a naive fix, and the completeness oracle (D-RA-1) guards against it.
* **Excluded.** The CLI-registration and declaration-index hardening (old RA-6..RA-14), and
  everything else from `DC921AF6`, `3750C37C` and `0ECC1895`.

**Shipment 2 is not planned now and returns to deliberation.** It covers `DC921AF6` (R-A1),
`3750C37C` and `0ECC1895` (package closure and Windows env mutation). The entries stay
**active** in the stash. Re-deliberation must start from a **bounded threat model**: an
explicit attacker capability set, and which of those capabilities the gate is responsible
for. It must not start from an open-ended list of AST rules. Three cycles showed that each
new rule invites a new bypass (aliasing, `maps.Copy`, raw-string literals, `.s`/`.syso`
files, bodyless funcs, blank-identifier index keys). That matches the 2026-10-01 denylist
bypass-loop learning.

* **Candidate alternative to record.** Required review through **CODEOWNERS on
  `tools/gatecheck/**`**, instead of ever-stricter code rules.
  * Today `.github/CODEOWNERS` is advisory metadata only. It covers workflows, the
    constraints directory and `scripts/check-retired-architecture.sh`, but **not**
    `tools/gatecheck/**`.
  * Making it a control needs an operator action: enable "Require review from Code Owners"
    on `main`.
  * It also needs the CODEOWNERS risk R1 to be resolved first, which means provisioning an
    approval path for unattended dark-factory PRs. Otherwise those PRs deadlock.
  * Re-deliberation should weigh that control, plus a small pinned core, against further AST
    rules.

**New stash entries** (captured this session):

* **HIGH bug.** `selectRepoPaths` runs `git ls-files` without `-z` (select.go:33), so paths
  that git quotes (non-ASCII and others) under `cmd/`/`internal/` are silently never scanned.
  This was R-A2d in attempt 3.
  * **Kept out of Shipment 1.** It changes `selectRepoPaths`/`DefaultGitRunner` argv and
    `pysem` splitting, which is the frozen selection contract. It is not the TOML engine
    surface that `9FC28DB9` fixes, so it fails the "trivially within the same contract
    surface" test.
  * **Coupling.** Once `-z` lands, Shipment 1's `core.quotePath` contamination vectors no
    longer distinguish contaminated from clean. That fix must replace them.
* **LOW.** Cross-engine symlink containment for writepath, unignore and mergestrategy. This
  is the D-RA-2 remainder.
* The compound learning on maintaining pin canonical texts went to `docs/compound/` rather
  than the stash.

**Consumption.** Only `9FC28DB9`, `990AFA71` and `D7BF9F74` are consumed and archived at Step
5.6. This supersedes the "All six are consumed" sentence in (B) above. `3750C37C`,
`0ECC1895` and `DC921AF6` stay active. Their reconciliation outcomes in (B) remain valid.

**Outcome (recorded at the end of the session, 2026-10-07).**

* **Stash IDs:** the HIGH `-z` entry is **4537B2F6** and the LOW cross-engine entry is
  **D44D8BDF**. The duplicate scan was clean for both.
* **Annotations:** `stash edit` annotated DC921AF6, 3750C37C and 0ECC1895 with this
  disposition.
* **Plan-review cycle for the narrowed plan:** it is a fresh review unit.
  * Attempt 1 (rev 1): FAIL.
  * Attempt 2 (rev 2): ADVISORY with no P1. The advisories were folded into rev 3 in place.
* **Harvest:** covering feature **050-F**, tasks **050.001-T..050.006-T** and queued shipment
  **040-S**.
* **Archived:** 9FC28DB9, 990AFA71 and D7BF9F74.
* **Compound learning:**
  `docs/compound/2026-10-07-pin-canonical-text-maintenance-and-unbounded-ast-rule-review.md`.

## Operator checkpoint

**Superseded in part by D-RA-7.** Only D-RA-1..D-RA-3, as narrowed by D-RA-7, proceed to a plan.
D-RA-4..D-RA-6 are not adopted, and Shipment 2 returns to deliberation.

The recommended options D-RA-1..D-RA-6 were applied under the relayed approval. The operator
can override any of them on the staging PR before Ship claims the shipment. The highest-impact
choices to review are D-RA-1 (C, not A) and the D-RA-3 accepted residuals R-A2a..R-A2c.

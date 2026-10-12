---
title: "Hermetic and containable start-script integration tests implementation plan (Batch 3)"
description: "Plan for Batch 3 (stash C5E50EFD, absorbing 26EDD633): scratch-copy start.ps1 under a sidecar-free PATH and cmd.WaitDelay for the two launcher integration tests; one release unit, test-only, three tasks; ships after 043-S, 044-S, 045-S"
date: 2026-10-12
source: "docs/decisions/2026-10-12-intercom-go-start-script-test-hermetic-decision-note.md"
requires_plan_hardening: false
tags:
  - "tests"
  - "start-script"
  - "hermetic"
  - "integration"
---

# Plan: Hermetic and containable start-script integration tests

Source: `docs/decisions/2026-10-12-intercom-go-start-script-test-hermetic-decision-note.md`
(decisions **D-HS-1..D-HS-6**). Consumed stash: `C5E50EFD` (accepted Orchestrator RCA,
operator disposition 2026-10-11T21:47 -07:00), `26EDD633` (absorbed, D-HS-6). Not absorbed:
`96B308D1`. Origin: shipment 042-S / feature 052-F closure (PR #123, PR #124).

## Problem Frame

`tests/integration/start_script_test.go` drives the real `start.ps1` for two tests:
`TestInvokeStartScriptMainPropagatesNonZeroExitCode` (`pwsh -File <repo>/start.ps1` directly) and
`TestLocationRestoredEvenOnNonZeroCopilotExit` (a helper script that runs
`pwsh -File <repo>/start.ps1`; its outer command sets no `cmd.Env`/`cmd.Dir`). Both inherit the
developer's environment and run the script **in place**. Root causes (accepted RCA; each anchor
below was re-read in `start.ps1` by Stage):

1. `Get-EnabledSidecars` (start.ps1, line ~255) is a render-time constant (`engram`,
   `graphtor-docs`, `backlogit`), called unconditionally by `Invoke-StartScriptMain`; the script
   uses `$PSScriptRoot` (the real repository) for `Push-Location`, `.env.local`, `.copilot`,
   `.engram`, and the `.graphtor\bin\graphtor-docs.exe` fallback. On a dev box with sidecars on PATH
   (`C:\Tools\backlogit.exe`, `engram.exe`, `graphtor-docs.exe`) the test runs a real backlogit
   sync (about 8 s), an Engram pre-warm that exhausts its 15 s budget, and a graphtor sync: 25 to
   35 s against a 30 s limit (about 2 s on Linux CI, which has no sidecars).
2. The 30 s `context.WithTimeout` does not contain the process tree. `start.ps1`
   (`Invoke-EngramCommandWithProgress`, `Kill()` at line ~95) terminates only the launched engram
   process; the index worker survives holding the inherited stdout pipe, and `CombinedOutput` has
   no `WaitDelay`, so the test blocks until the worker exits (observed 140 to 206 s; one 10-minute
   package panic on 2026-10-08).
3. The tests write real workspace state.

Verified fix shape (Orchestrator RCA): a scratch copy run with no sidecars on PATH measured 2.1 s,
exit 7, same assertions. `26EDD633` (same test, same symptom, workspace-size hypothesis not
confirmed) is absorbed: hermeticity removes the workload it asked to budget. `start.ps1` has no
dot-sourcing or `Import-Module`, and its only repo-relative reads are the five `$PSScriptRoot`
uses above, so a byte copy in a scratch directory loses nothing the tests need.

## Release unit and sequence

One release unit (feature "Hermetic and containable start-script integration tests"), one
shipment (S3). **Required merge/ship order:** Batch 1 staging PR (shipment 043-S), then Batch 2
staging PR (044-S, 045-S), then this Batch 3 staging PR; Ship claims 043-S, 044-S, 045-S, then S3.
S3 records **two** `blocks` edges, per the Batch 1 contract ("`backlogit dep add <later-S> 043-S`")
and the operator directive: `backlogit dep add <S3> 045-S` (ordering edge) and
`backlogit dep add <S3> 043-S` (direct Batch 1 contract edge). Neither edge reflects a file
dependency; they encode "before any new implementation shipment" ordering. All staging PRs are
merge commits (Principle XI).

**File overlap check (Step 1.9 base is the Batch 2 tip):**

| Batch | Files | Overlap with this batch |
|---|---|---|
| 1 (043-S) | `.github/**`, `.autoharness/**`, policy docs, skill frontmatter; `tests/integration/ship_feature_completion_contract_test.go`, `shipment_close_path_behavior_test.go`, `shipment_close_path_harness_test.go`, new `cascade_close_classify_only_smoke_test.go` | **No file overlap.** Same Go package (`package integration`): shared namespace only. New identifiers use `hermeticStart*`, `startScript*`, and `startScriptHermetic*` prefixes. Batch 1 deletes two replica test files; before Ship, confirm they do not define `repoRoot`, `writeStubExecutable`, or `filterEnvKeys`, which stay where they are. Each task's AC runs `go vet ./tests/integration/...` on a tree that includes Batch 1; a collision after the base merges is fixed by renaming on a merge of the base branch (no rebase). |
| 2 (044-S, 045-S) | `tools/gatecheck/**`, `scripts/**` | None |
| 3 (this) | new `tests/integration/start_script_hermetic_test.go`, new `tests/integration/start_script_bounded_test.go`, edit `tests/integration/start_script_test.go` | n/a |

`start.ps1` is **not** modified (test-only change). File ownership is one-directional: new
helper files consume the existing shared helpers `filterEnvKeys`, `startScriptTimeout`,
`repoRoot`, and `writeStubExecutable` (which stay in place); `start_script_test.go` consumes the
new helpers.

## Requirements Trace

| Requirement (stash C5E50EFD / operator) | Unit | Decision |
|---|---|---|
| Run a scratch (hermetic) copy of start.ps1 | U1 (guard + helper), U3 (wiring) | D-HS-1 |
| Under a PATH containing no sidecars | U1, U3 | D-HS-2 |
| `cmd.WaitDelay` on both named tests | U2 (guard + helper), U3 (wiring) | D-HS-3 |
| Same assertions, exit 7, about 2 s | U3 | D-HS-1..3 |
| Reconcile 26EDD633 (no budget guard needed; symptom covered) | U3 (AC 4) | D-HS-6 |
| Test-first (Principle II) | red-then-green inside every unit | |
| Linux CI + Windows dev host | U1..U3 | D-HS-4 |

## Implementation Units

All units are test code in `tests/integration` (single skill domain: tests). Posture: test-first
**inside each unit**: write the guard, observe it red against a deliberately naive or absent
implementation (compile error on the missing helper counts as red, recorded), implement, observe
green. Every unit therefore ends green and is a closable, buildable Ship task (no red commit is
ever produced). No `start.ps1` or production change.

### U1: Hermetic scratch copy and sidecar-free PATH, with guards

* **Task:** "Add hermetic start.ps1 scratch copy and sidecar-free PATH helpers with guards".
  Size S, complexity medium.
* **File:** new `tests/integration/start_script_hermetic_test.go` (1 file).
* **Helpers (2):**
  * `hermeticStartScript(t *testing.T, root string) string`: read the real `<root>/start.ps1`,
    write the bytes to `t.TempDir()/start.ps1`, return that path. (`t.TempDir()` is already
    private; no explicit mode games.)
  * `hermeticStartEnv(t *testing.T, base []string, overrides map[string]string) (env []string, pwshExe string)`:
    resolve `pwshExe` as an absolute path with `exec.LookPath("pwsh")` against the **parent** PATH
    (so callers invoke it by absolute path, since `exec.Command("pwsh")` resolves with the parent
    PATH, not `cmd.Env`); filter the env with `filterEnvKeys` (case-insensitive `Path`/`PATH`);
    rebuild PATH with `filepath.SplitList`/`os.PathListSeparator`, dropping every directory that
    contains an executable named `backlogit`, `engram`, or `graphtor-docs` (candidate names: the bare
    name, `.ps1` on every OS, and every extension from `PATHEXT` on Windows); keep all other
    directories; append PATH under one spelling plus `overrides`. If `pwsh`'s own directory holds
    a sidecar, `t.Skipf` with an explicit reason (one guarded line, honest skip). The scrub is
    intentionally coarse (a directory such as `C:\Tools` is dropped whole). Environment surface
    decision (D-HS-2): PATH plus the caller's overrides only; `start.ps1` does not read `GIT_*`,
    and `COPILOT_EXE` is shadowed by the test's `COPILOT_EXE_PATH`; `Resolve-GitHubTokens`'s
    read-only `gh auth token` call and ambient `GITHUB_*` tokens are a **noted, accepted
    residual** (not a sidecar, no workspace writes).
* **Guards (2 scenarios):**
  * **G1 scratch isolation:** the returned path differs from the real script path; a real `pwsh`
    (absolute path, `-NoProfile`) that dot-sources it prints `$PSScriptRoot`; `os.SameFile` shows
    it is the scratch directory (the parent of the returned path) and **not** `root`.
  * **G2 sidecar-free PATH (positive proof with a control):** write fake sidecar stubs
    (`backlogit`, `engram`, `graphtor-docs` with the platform extension; mode `0o755` on POSIX) into
    a temp dir and prepend it to a base env PATH. **Control:** under the unscrubbed env,
    `Get-Command backlogit,engram,graphtor-docs -ErrorAction SilentlyContinue` finds them. After
    `hermeticStartEnv`, the temp dir is absent from PATH and the same real `pwsh -NoProfile`
    command outputs nothing, while `pwsh` still resolves. The injected stub dir makes the guard
    non-vacuous on sidecar-free CI.
* **Red evidence (recorded in the task):** the guards are written before the helpers; the compile
  failure on the missing helpers, then (with a naive pass-through helper) G1 and G2 failing on
  assertions, are captured before the real helpers go in.
* **AC:** G1 and G2 green on the Windows dev host; `go vet ./tests/integration/...` and
  `$env:GOOS='linux'; go vet ./tests/integration/...; Remove-Item Env:GOOS` clean;
  `gofmt -l tests/integration` empty; `git diff --stat` lists exactly the one new file.

### U2: WaitDelay bounded runner with containment guard

* **Task:** "Add WaitDelay-bounded command runner with pipe-holder containment guard". Size S,
  complexity medium. Blocked by U1 (ordering only; separate files).
* **File:** new `tests/integration/start_script_bounded_test.go` (1 file).
* **Helper (1):** `startScriptCombinedOutputBounded(cmd *exec.Cmd) ([]byte, error)` sets
  `cmd.WaitDelay = startScriptWaitDelay` (`const startScriptWaitDelay = 5 * time.Second`) then
  returns `cmd.CombinedOutput()`. Callers keep the
  `errors.Is(ctx.Err(), context.DeadlineExceeded)` timeout branch first: when the context kills
  the process, `Wait` returns the process's `*exec.ExitError`; `exec.ErrWaitDelay` surfaces only
  when the process exited 0 while a descendant still held the pipe (verified against
  `GOROOT/src/os/exec/exec.go` for the pinned go1.26.9 toolchain, with the `file:line` citations
  for `WaitDelay` handling and kill-on-context recorded in the task evidence/comment).
* **Guard G3 (helper-process pattern, no `pwsh`, runs on Linux CI):**
  `TestStartScriptHermeticHelperProcess` returns immediately unless an anchored marker env var is
  set (`-test.run=^TestStartScriptHermeticHelperProcess$`). In helper mode it spawns a
  grandchild (same test binary, sleep mode, about 20 s) with **`Stdout`/`Stderr` set to
  `os.Stdout`/`os.Stderr`** so the grandchild genuinely holds the inherited pipe, writes the
  grandchild PID to a file atomically (temp file then rename), and sleeps. The guard runs the
  helper through `startScriptCombinedOutputBounded` under `context.WithCancel`; it polls for the
  PID file (outer deadline 10 s; **fail** if it never appears, so the guard cannot pass
  vacuously), cancels, and asserts return within 10 s **measured from the cancel**. The kill
  `t.Cleanup` is registered after `t.TempDir()`, so it runs first, kills the recorded PID, and
  ignores "process already finished". The package has no `TestMain` (Stage-verified).
* **Red evidence:** with the helper's `WaitDelay` assignment absent (naive runner), G3 blocks about
  20 s and fails the 10 s bound; recorded before the assignment goes in.
* **AC:** G3 green; `go vet` (Windows and linux) clean; `gofmt -l tests/integration` empty;
  `git diff --stat` lists exactly the one new file; no orphan `test binary` process remains after
  the run (checked and recorded).

### U3: Rewire the two launcher tests and verify containment

* **Task:** "Rewire the two start-script tests to the hermetic bounded harness". Size S,
  complexity low. Blocked by U2.
* **Files:** `tests/integration/start_script_test.go` (edit the two tests) and
  `tests/integration/start_script_hermetic_test.go` (add one tiny `writeCwdEchoStub` helper
  that writes a copilot stub printing `stub copilot ran` and `STUB_CWD=<Get-Location>`; the
  existing `writeStubExecutable` is untouched).
* **Change:**
  * `TestInvokeStartScriptMainPropagatesNonZeroExitCode`: use `hermeticStartScript`, the env from
    `hermeticStartEnv` (existing `COPILOT_EXE_PATH`/`COPILOT_HOME`/`ENGRAM_DATA_DIR` overrides),
    the cwd-echo stub (exit 7), and `startScriptCombinedOutputBounded`. Keep every existing
    assertion (exit code 7, `stub copilot ran`) and add one **wiring assertion**: the
    `STUB_CWD` printed by the stub (`start.ps1` does `Push-Location $PSScriptRoot`) is
    `os.SameFile` as the scratch script's directory and not `root`. This makes the rewiring
    discriminating even on sidecar-free CI, where the old assertions would pass either way.
  * `TestLocationRestoredEvenOnNonZeroCopilotExit`: this test currently sets **no** `cmd.Env` and
    **no** `cmd.Dir`; U3 adds the `hermeticStartEnv` env to the outer `pwsh` command, passes the
    scratch script path and the absolute `pwshExe` into the helper script (replacing bare
    `pwsh`), and calls `startScriptCombinedOutputBounded`. `Set-Location` to `callerDir` stays, so
    `LOCATION_UNCHANGED` semantics are unchanged. Its scratch isolation is covered by G1
    (characterization; the helper discards the inner output).
  * Correct the stale "sidecars ... absence is non-fatal" comment so it states the real mechanism.
* **AC:**
  1. G1, G2, G3 and both rewired tests green.
  2. On the Windows dev host with `C:\Tools` sidecars on the ambient PATH,
     `go test ./tests/integration -run '^(TestInvokeStartScriptMainPropagatesNonZeroExitCode|TestLocationRestoredEvenOnNonZeroCopilotExit)$' -count=3 -v`
     passes, each run **at most 10 s** (expect about 2 s; durations recorded).
  3. Containment evidence: `git status --porcelain` and a listing of `.backlogit`, `.engram`,
     `.copilot` **with timestamps** are identical before and after the run (no real-workspace
     writes), and no stray `engram`/`backlogit`/`graphtor-docs` process is left running by the
     run.
  4. `26EDD633` coverage: one full-package run `go test ./tests/integration/... -count=1` passes
     with both tests at most 10 s each and no `start.ps1 exceeded` message.
  5. Gates: `go build ./...`, `go vet ./tests/integration/...`, the linux cross-vet,
     `gofmt -l tests/integration` empty, and `go test ./...` with **no new failure versus the
     baseline** (the pre-existing `96B308D1` `TestGatecheckCleanup_*` Windows failures are known
     and not in scope).
  6. `git diff --stat` lists only the two (U3) test files; `start.ps1` unchanged.

## Dependency Graph

```text
U1 (hermetic copy + PATH + G1/G2) -> U2 (WaitDelay runner + G3) -> U3 (rewire + verification)
```

No cycles. External ordering: shipment S3 `blocks` on 045-S and on 043-S (see sequence section).
U1 and U2 touch separate files; the chain U1 -> U2 is ordering only (single-active Ship).

## Constitution Check

* **II Test-First:** each unit writes its guard first and records red before green; no unit ends
  red, so Ship's green-at-every-task gate holds.
* **III Workspace Isolation / IV Containment:** removes real-workspace writes from the two tests;
  scratch files live under `t.TempDir()`.
* **I Safety-First Go:** the helpers return errors via `t.Fatalf`; no `unwrap`-style shortcuts.
* **VII Destructive Command Approval / role boundary:** no destructive command; Stage writes no
  test code (this plan only specifies it); the G3 cleanup kills only a PID the guard itself
  spawned.
* **XI Merge Commit History:** staging PR and shipment PR are merge commits; collision fixes use a
  merge of the base, not a rebase.
* **Task Granularity:** every unit is at most 2 files, fewer than 5 functions (U1: 2 helpers + 2
  guards = 4; U2: 1 helper + helper-process test + G3 = 3; U3: 2 tests + 1 stub writer = 3),
  at most 2 scenarios, one skill domain (tests), one atomic milestone.
* **VIII Safety modes:** standard mode is sufficient; test-only, no security or contract surface.

## Decisions and Rationale

See the decision note D-HS-1..D-HS-6. Summary: byte-copy rather than symlink (D-HS-1); drop
sidecar-bearing PATH directories with a positive `Get-Command` proof, a control, and an injected
fake sidecar so the guard is red even on sidecar-free CI (D-HS-2); one shared `WaitDelay` runner
proven empirically by a helper-process guard (D-HS-3); no new build tags or Windows-only skips
(D-HS-4); scope held to two tests (D-HS-5); `26EDD633` absorbed, `96B308D1` not (D-HS-6).
**Review-driven adjustments (plan-review, Stage disposition):** merged the former red-only guard
task into the implementing tasks so every task ends green; dropped the tautological byte-equality
guard; kept G3 (operator named containment as a requirement and the compound learning requires
empirical proof of stdlib claims) but hardened it against vacuous pass; added a wiring assertion
(stub cwd) and a G2 control.

## Risks and Caveats

| Risk | Mitigation |
|---|---|
| `pwsh` shares a directory with a sidecar (scoop/choco/`/usr/local/bin` layouts), so the scrub removes it | absolute `pwshExe` pinned from the parent PATH; honest one-line `t.Skipf` with reason |
| PATH key spelled `Path` on Windows, so a duplicate PATH wins nondeterministically | `filterEnvKeys` case-insensitive removal, then a single re-append; G2 runs through the real child |
| Denylist PATH scrub misses a spelling | G2 is a positive `Get-Command` proof with a control and an injected stub; candidate names include bare, `.ps1`, and `PATHEXT` |
| `WaitDelay` semantics misstated | read `GOROOT/src/os/exec/exec.go` (`go env GOROOT`, pinned go1.26.9) and cite lines; G3 proves behavior |
| G3 passes vacuously (pipe never held) | PID-file gate plus explicit `Stdout`/`Stderr` inheritance; fail if the PID file never appears |
| `WaitDelay` does not kill orphans | containment only; orphan creation is removed at source by U1; G3 cleanup kills its grandchild |
| Same-package identifier collision with Batch 1 | distinct prefixes; `go vet` on the merged tree; fix by renaming on a merge of the base |
| Timing flakiness on slow hosts | G3 bound is 10 s against a 20 s grandchild; the launcher tests keep the 30 s budget against a measured 2 s (AC bound 10 s) |
| Residual non-hermetic surface (`gh auth token`, ambient `GITHUB_*`) | read-only, not a sidecar; noted and accepted (D-HS-2) |

## Plan Hardening Signals

* public API, schema, or contract change: **absent** (test-only; no production or contract change).
* security, auth, permission, or compliance-sensitive behavior: **absent** (no trust-boundary
  change; env scrubbing narrows ambient inheritance in tests only).
* migration, backfill, destructive data/config action, or irreversible step: **absent**.
* external integration, operator checkpoint, or external dependency: **absent** (sidecars are
  removed from the tests, not integrated).
* high runtime, rollout, or rollback risk: **absent** (revert is a revert of test-file commits).

Requires plan hardening: no

## Runtime Verification and Closure

No runtime surface changes (test harness only). Verification is the U3 AC set: measured durations
for both tests with sidecars present on the ambient PATH, before/after workspace-state equality,
no stray process, one full-package run (covers the `26EDD633` full-suite symptom), and Linux vet.
Closure artifact: the Ship verification evidence records those durations. No monitoring or rollback
window is needed beyond reverting the test commits. Stage closure steps (not Ship tasks): archive
`C5E50EFD` and `26EDD633` with promotion notes, and capture the snippet-runner `WaitDelay`
residual (D-HS-5) as a P-021 stash entry.

## Plan Review

dispatch_mode: multi-agent
decision: ADVISORY

**Gate decision: ADVISORY.** No P0 or P1 finding from any persona (round 1: five personas on the
first draft; round 2: anchor Architecture pass on the revised plan, no findings). Findings were P2
and P3 only, so the gate is ADVISORY; Stage **authorizes harvest** under the Orchestrator
directive for this batch (the operator may override at staging-PR review). All P2 findings were
dispositioned in this revision (below) rather than deferred. Plan hardening: not required
(`Requires plan hardening: no`, all five signals absent); no `## Plan Hardening` section is
needed and none of the five hardening-signal FAIL conditions apply.

Capability probe: `TOOL_OK: reviewer-subagent-dispatch`. Anchor route `model_routing.anchor_review`
(`openai` / `gpt-6.1-sol` / `high`) was dispatched for the Architecture Strategist pass on the
revised plan (`TOOL_OK: anchor-review-route`). Round 1 personas ran at harness-default models.

| Persona | Mode | Result |
|---|---|---|
| Constitution Reviewer | subagent | 0 P0/P1, 2 P2, 4 P3 |
| Go Reviewer | subagent | 0 P0/P1, 2 P2, several P3 |
| Scope Boundary Auditor | subagent | 0 P0/P1, 4 P2, 3 P3 |
| Architecture Strategist | subagent (round 1), anchor reviewer pass (round 2) | 0 P0/P1, 3 P2, 3 P3; round 2: none |
| Learnings Researcher | subagent | 0 P0/P1, 2 P2, 3 P3 |
| Agent-Native Parity Reviewer | not triggered (no MCP/agent-facing action) | n/a |
| Security Lens Reviewer | not triggered (no auth, API, secrets, or trust-boundary change) | n/a |

### P2 findings and Stage disposition (all resolved in this revision)

* **U1 ended red / not closable; conflicts with Ship's green-at-every-task gate**
  (Constitution, Scope, Architecture): resolved. The red-only task is gone; red-then-green now
  lives inside each of U1..U3, so every task ends green. Scope's alternative (collapse to 1-2
  tasks) was rejected to keep tasks under the constitution's fewer-than-5-functions heuristic.
* **G3 can pass vacuously; stdout not inherited; PID reuse; error type** (Go): resolved: PID-file
  gate that fails if the pipe is never held, explicit `Stdout`/`Stderr` inheritance, atomic PID
  write, kill-by-recorded-PID in a `t.Cleanup` that ignores finished processes, timing measured
  from cancel, and the `*exec.ExitError` vs `ErrWaitDelay` semantics recorded.
* **G1 temp-dir comparison unreliable** (Go): resolved: `os.SameFile` on the printed
  `$PSScriptRoot` against the parent of the scratch path; the tautological byte-equality guard was
  dropped (Scope).
* **G3 and helper machinery disproportionate** (Scope): partly accepted. The `combinedOutputBounded`
  wrapper was kept as a single shared, renamed helper (`startScriptCombinedOutputBounded`) because
  the operator named containment as a requirement and the compound learning demands empirical
  proof of stdlib claims. G3 was hardened, not removed.
* **Unreachable `t.Skipf`** (Scope) vs plausible on scoop/choco layouts (Go): kept as a single
  honest guarded line; reachability is no longer claimed.
* **Verification gaps** (Scope): resolved: orphan-process check, `-count=3` on the two tests, one
  full-package run, a single 10 s bound, mtime-aware workspace-state comparison, and explicit
  Stage closure ownership for archiving and the snippet-runner stash entry.
* **Quality gates** (Constitution): resolved: `gofmt -l` empty and `go test ./...` with no new
  failure versus baseline (known `96B308D1` failures noted).
* **Plan accuracy: second test has no `cmd.Env`/`cmd.Dir`** (Constitution): resolved: U3 states it
  adds that wiring.
* **Dependency edges differ from the Batch 1 contract** (Architecture): resolved: S3 records both
  `blocks` edges (045-S and 043-S).
* **Cross-file helper ownership cycle** (Architecture): resolved: one-directional ownership stated.
* **Wiring side-effect assertion; env-surface decision** (Learnings): resolved: stub-cwd
  `os.SameFile` assertion in test 1, G2 control, explicit env-surface decision (PATH plus
  overrides; `gh`/`GITHUB_*` accepted residual).

### P3 acknowledgments

Kill-by-PID reuse risk (accepted, handled by ignore-finished), rebase wording (now merge of base),
`0o700` mode (dropped), `GOOS=linux` PowerShell syntax (fixed), `PATHEXT` and `.ps1` candidates
(adopted), citation of `exec.go` and `start.ps1` lines (adopted), `pwshExe` return value (adopted),
Principle I and IV acknowledgments (added), `Resolve-GitHubTokens` residual (recorded), Batch 1 U13
replica-file check (added to overlap table).

### Runtime verification and closure

Test-only; no runtime surface. Verification and closure expectations are explicit (U3 AC 2 to 6
and "Runtime Verification and Closure"); no gaps outstanding.

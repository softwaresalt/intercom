---
title: "Validation method repair: gatecheck temp containment diagnostic"
description: "Evidence, corrected validation recipe, temp contract and gate provenance for the open policy-snapshot containment breaker"
date: 2026-10-30
status: method-repaired-suite-not-rerun
tags: [memory, validation, diagnostics, circuit-breaker]
---

## Scope and authority

* Operation identity: `validation-method-temp-contract-20261030`
* Approval: operator reply "ok, apply the recommended fixes now" after the
  rubber-duck critique. It authorizes ONE standalone diagnostic, same-process
  environment setup, evidence capture and this documentation.
* It does not authorize a fourth run of the blocked full-suite or
  `tools/gatecheck` operation, any assertion, skip, timeout or containment
  change, original shipment 073-S, a PR, a push or a merge.
* The breaker in `circuit-break-policy-snapshot-temp-containment.md` stays
  OPEN at `attempts: 3`. This diagnostic is a separate operation and does not
  reset, extend or close that counter.
* Branch `fix/dark-staging-publication` at
  `4dcf82efeb8549a1c1aea3bb466fccb659a81ab8`. Stage branch
  `chore/stage-gatecheck-wrapper-test-ci-reliability` (`3a6f5e8`) untouched.
* Engram was not probed (two earlier readiness failures stand). Reads used
  known file paths only: degraded lookup, no external retrieval.

## Verified call order (read-only source)

Every claim below is from source read in this session.

* `workspaceTempDir` (`tools/gatecheck/wrapper_harness_test.go` lines 22-55)
  takes its workspace from `os.Getwd`, which in a `go test` binary is the
  package directory `tools/gatecheck`, not the repository root (lines 25-31).
* The helper sets TMPDIR, TMP and TEMP to that package directory (lines
  33-35), then calls `t.TempDir()` (line 37), checks containment with
  `filepath.Rel` against the package directory and fails with
  `temporary directory is outside workspace` (lines 42-49) and finally
  repoints all three variables at the new directory (lines 51-53).
* The helper's `t.TempDir()` therefore overrides the initial TEMP, TMP and
  TMPDIR before use. Their initial values cannot change the helper's verdict
  unless a `t.TempDir()` parent was created earlier in the same top-level test.
* `installGoShim` calls the helper first (line 61 of the same file). Callers:
  * `TestCheckRetiredArchitectureWrapper_WrongTreeRootAnchoring`:
    `requireBash` then `installGoShim`
    (`check_retired_architecture_wrapper_test.go` lines 277-278)
  * `TestCheckWritePathWrapper_WrongTreeRootAnchoring`: same order
    (`check_write_path_wrapper_test.go` lines 216-217)
  * `TestCheckWritePathWrapper_LegitimateModesReachBuild`: `requireBash`,
    `writePathWrapperPath` (path math only, lines 200-207), `installGoShim`
    (lines 46-48)
  * `TestInstallGoShim_UnexpectedInvocationFailsClearly`: `requireBash` then
    `installGoShim` (lines 116-117)
  * `TestSameBashFile_ComparesByIdentity`: `requireBash` then the helper
    directly (`wrapper_harness_test.go` lines 508-509)
* `requireBash` and `requireWindowsCompatibleBash`
  (`runner_test.go` lines 26-49) only run `exec.LookPath` and `uname -o`.
  They create no temp directory. None of the five tests changes directory.
* So in all five failing tests the helper's `t.TempDir()` is the FIRST
  `t.TempDir()` of the top-level test. The cached-parent hypothesis does not
  apply to them. The same five tests are all of the helper's users.
* Go 1.27.0 `testing.(*common).makeTempDir`
  (`C:\Program Files\Go\src\testing\testing.go` lines 1581-1636) creates the
  per-test parent with `os.MkdirTemp(os.Getenv("GOTMPDIR"), pattern)`
  (line 1613). A non-empty GOTMPDIR wins over TMP and TEMP. The parent is
  cached for the life of that test (lines 1587-1590).
* `os.TempDir` on Windows calls `GetTempPath2` when available
  (`os/file_windows.go` lines 284-304).

## Diagnostic run

One standalone file, no repository package imported, no tracked source change:
`logs/diagnostics/validation-method-temp-contract-20261030/tempprobe_windows_test.go`.
It replicates the helper statement by statement, records the verdict instead
of calling `t.Fatalf`, and logs only these environment keys: TEMP, TMP,
TMPDIR, GATECHECK_TMPDIR, GOTMPDIR. Raw environment was never read.

* Pre-run checks: `logs/` is a real directory, no reparse point; the scope
  directory is a real directory below it; `git check-ignore` reports
  `.gitignore:59:logs/`.
* Single-file `gofmt -w` and `go vet` ran on the probe only, before the run.
* Command, run from the scope directory in one PowerShell process:

```text
go test -count=1 -v -timeout 4m .\tempprobe_windows_test.go
```

* The launcher wrote stdout and stderr to separate files and read the native
  exit code from the process object, not from the shell.
* Start 2026-10-30T04:53:15Z, end 04:53:29Z. Native exit code: 0.
  Stdout 32 lines, 20138 bytes. Stderr 0 bytes. First launch attempt was
  rejected by the shell tool before execution (no process started), so the
  probe ran exactly once.
* Parent environment at launch: TEMP and TMP
  `C:\Users\DEWILL~1\AppData\Local\Temp` (8.3 form), TMPDIR, GATECHECK_TMPDIR
  and GOTMPDIR unset. Go: `C:\Program Files\Go\bin\go.exe`, go1.27.0.
* The five subtests report `PASS` because the probe itself ran. The
  containment verdicts are in the JSON evidence, not the PASS marker.

| Case | Setup before the helper | Helper verdict |
|---|---|---|
| A | none (inherited TEMP outside) | inside |
| B | one earlier `t.TempDir()` under outside TEMP | OUTSIDE |
| C | GOTMPDIR set to the outside TEMP | OUTSIDE |
| D | TEMP, TMP, TMPDIR already set to the package directory | inside |
| E | `LookPath("bash")` and `uname -o` first (TestSameBashFile order) | inside |

Observed facts:

* Child CWD and its canonical path were identical, no junction or short-name
  difference for the workspace itself.
* After the helper's `t.Setenv`, `os.TempDir()`, `GetTempPath` and
  `GetTempPath2` all returned the package directory, so Windows honored the
  in-process change.
* Case A equals case D: initial TEMP is irrelevant when no earlier
  `t.TempDir()` exists.
* Cases B and C reproduce the exact failure text condition
  (`temporary directory is outside workspace`) with an unchanged helper.
* The outside directory canonicalizes to a long-name path
  (`C:\Users\dewilliams\...`) while the inherited env uses `DEWILL~1`. Lexical
  `filepath.Rel` is already outside in both spellings, so this is not the cause.
* On this machine's inherited PATH, `bash` resolves to
  `C:\Windows\system32\bash.exe` and `uname -o` returns `GNU/Linux`. The
  helper's gate (`requireWindowsCompatibleBash`) would SKIP the five tests
  here, not fail them. That call took about 10 seconds.
* The validator failures were assertion failures, not skips, so the validator
  process had a different bash on PATH. Its PATH was not captured.

Evidence (ignored local files, SHA-256 prefix, repository-relative):

* `logs/diagnostics/validation-method-temp-contract-20261030/evidence.jsonl`
  (18894 bytes, `88F588B1018281CC`)
* `logs/diagnostics/validation-method-temp-contract-20261030/go-test.stdout.log`
  (`A35C9894FAF48DF1`)
* `logs/diagnostics/validation-method-temp-contract-20261030/go-test.stderr.log`
  (empty)
* `logs/diagnostics/validation-method-temp-contract-20261030/tempprobe_windows_test.go`
  (`5C6687DB3040F595`)

## Root cause status

Proven:

* The helper is correct when it is the first `t.TempDir()` of the test and
  GOTMPDIR is unset or inside the package directory.
* Initial TEMP, TMP and TMPDIR changes could never fix the five failures. The
  three validator attempts varied only those variables.
* A prior cached `t.TempDir()` is ruled out for the five tests by call order.

Not proven:

* What the validator process actually contained. The only mechanism found that
  produces the failure for a first-use `TestSameBashFile` is a non-empty
  GOTMPDIR that points outside `tools/gatecheck` (case C). That GOTMPDIR was
  never observed in the validator. Treat it as the leading hypothesis only.
* Which bash the validator used.

No defect in the existing helper is proven. No test-infrastructure fix is
proposed.

## Corrected validation method and temp contract

Premises that were wrong:

* A global TEMP remapping, or an immutable clone, reproduces normal test
  semantics. It does not. The gatecheck helper confines itself to its own
  package directory, and the integration tests need the process temp root to be
  outside the repository (see the 03:41 blocker in
  `dark-staging-correction-memory.md`).
* Environment and CWD persist across tool calls. They do not. Each PowerShell
  call is a new process.

Contract (stated, not yet validated by a suite run):

* Gatecheck package tests
  * inside the package directory: the helper's `t.TempDir()` parent and the
    shim directory
  * outside requirement: none, but GOTMPDIR must be unset, or point inside the
    package directory
  * inherited TEMP, TMP and TMPDIR: leave at their defaults
  * bash: Git for Windows bash first on PATH, or the tests skip
* Integration tests (`tests/integration/build_script_test.go` output-path
  guards)
  * the process temp root must be OUTSIDE the repository under test
  * a workspace-local TEMP inside the repository breaks this premise
* Therefore one global TEMP value cannot serve both groups. Run them as
  separate native invocations with different scratch settings. This repairs
  the method only; it is not a result.
* "Repository under test" must be named per run. A clone under
  `logs/validation/...` is inside the authorized workspace but outside the
  checked-out repository, which is why sibling-TEMP and in-clone TEMP both
  differ from a normal local run.

Same-process recipe (preflight, then one native command). It is a recipe, not
a result, and it has not been run against the suite.

```powershell
# All setup and the native command in ONE PowerShell process.
$pkg = Join-Path (Get-Location).Path 'tools\gatecheck'
$env:PATH = 'C:\Program Files\Git\usr\bin;' + $env:PATH
Remove-Item Env:GOTMPDIR -ErrorAction SilentlyContinue
# Child-observed preflight: print only allowlisted keys and resolved tools.
foreach ($k in 'TEMP','TMP','TMPDIR','GATECHECK_TMPDIR','GOTMPDIR') {
  '{0}={1}' -f $k, $(if ($null -eq [Environment]::GetEnvironmentVariable($k)) { '<unset>' } else { [Environment]::GetEnvironmentVariable($k) })
}
(Get-Command bash).Source; (Get-Command go).Source
# Then the single authorized native command, with the exit code read from the
# process object, bounded output (1 MiB or 10000 lines), timeout of 5 minutes.
```

Required preflight observed by the Go child, not only by the shell: a
standalone probe like the one above (or the `snap` records it logs) must show
`os.TempDir`, GOTMPDIR, `Getwd` and the bash `uname -o` result (`Msys`) before
any suite run. That preflight is itself a counted, separately authorized step.

## Gate provenance

Sources read in this session:

* Constitution Principle II
  (`.github/instructions/constitution.instructions.md` lines 18-23): all tests
  MUST pass via `go test ./...` before any code is merged.
  * Constitution Quality Gates (same file, lines 197-205): run in order, "Do not
  skip any gate"; `go test ./...` is one of four. `AGENTS.md` lines 29, 208
  and 227 repeat it.
* `github-pr-automation.instructions.md` lines 560-600 (readiness Checks 2-5):
  the pre-PR gate is a local readiness verdict plus a full local BUILD for
  code-changing PRs, then the P-018 Copilot review gate. Line 138-139 asks
  for `go test ./...` when fixing review comments. Lines 640-662 require
  "required CI/checks green or explicitly marked non-applicable", P-009
  merge-commit-only and P-016 topology.
* So P-009, P-014 and P-018 as read here govern merge strategy, operator
  approval and Copilot review. They do not themselves require the full suite.
  The full-suite requirement comes from the constitution and AGENTS.md.
  The policy-registry text for P-014 itself was not re-read in this task.
* Conclusion: the local full suite cannot be silently waived. Nothing read
  allows a partial suite or a focused probe to stand in for it. The 03:41
  failures stand as unresolved.

Can remote CI supply canonical full-suite evidence before merge?

* `.github/workflows/ci.yml` triggers on `pull_request` to `main` and `push`
  to `main` (lines 102-108). A push to the corrective branch alone does not
  run it. A PR would be required, and none exists.
* Job `test` (id `expensive`, lines 195-241) runs `go build`, `go vet` and
  `go test -race -mod=readonly ./...` on ubuntu with Go `1.26.x` when the
  fail-closed `changes` filter reports code changed. The corrective PR edits
  Go tests, so it qualifies (docs and `.backlogit` paths only are excluded).
* The Windows job (lines 272-298) is advisory unless the repository variable
  `WINDOWS_GATE_REQUIRED` is `true`. Its value was not read.
* Whether the `test` check is a required branch-protection check was not
  verified. It is not assumed.
* A Linux CI pass is a different environment (Linux, Go 1.26, real bash,
  default temp root). It would be real full-suite evidence on that platform,
  not a repeat of the local Windows run. Whether that satisfies the
  constitution's "before any code is merged" without a local green run is an
  operator disposition, not something this task decides.
* No PR, push or merge was done here.

## Result and remaining authorization

* A method fix is justified for the temp contract and the same-process recipe
  above. No source, test, policy, timeout or containment change was made.
* The full suite, the five tests and original 073-S were not run. The open
  breaker is unchanged at 3 attempts.
* Needed from the operator before any suite run: (1) whether a single
  preflight probe of the real validator environment (GOTMPDIR, PATH bash) is
  authorized; (2) the disposition of the integration-versus-gatecheck temp
  premise conflict, and whether separate invocations are approved; (3) whether
  Linux PR CI may count toward the full-suite gate if a local green run stays
  unreachable.
* The immutable clone `logs/validation/policy-4dcf82e` and its ignored
  fixtures were not touched. Cleanup remains pending approval.

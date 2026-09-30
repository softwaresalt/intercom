# Session Memory — 034-S Gate-engine Go migration M1

Date: 2026-09-29 (UTC)
Shipment: 034-S ("Gate-engine Go migration M1: gatecheck scaffold, pysem, masker, write-path")
Feature: 044-F (12 tasks: 044.001-T .. 044.012-T)
Branch: `feat/034-s-gate-engine-go-migration-m1-gatecheck-scaffold-pysem-masker-write-path`
Mode: P-017 dark factory mode. merge_approval_pre_authorized=false, admin_fallback_pre_authorized=false.

## Task state (as of this checkpoint)

| Task | Status | Commit |
|---|---|---|
| 044.001-T write-path struct-tag fixtures | done | (prior session) |
| 044.002-T masker/pysem goldens from Python | done | (prior session) |
| 044.003-T write-path engine goldens | done | (prior session) |
| 044.004-T gatecheck main, exit contract, stubs | done | (prior session) |
| 044.005-T pysem character classes | done | (prior session) |
| 044.006-T pysem case mapping incl. Final_Sigma | done | cfe0ec9 |
| 044.007-T pysem string helpers | done | 2d9562f |
| 044.008-T pysem I/O and lookaround | done | 6e130f7 |
| 044.009-T gomask rune-level masker | done | 5a44cc6 |
| 044.010-T write-path engine port | done | f0d2915 |
| 044.011-T shared bash runner (M1-T9) | done | 5558dd6 |
| 044.012-T switch write-path wrapper to Go | **done** | **b352754** |

## 044.010-T summary
Fixed a compile error in `writepath_test.go` (undefined `maskForTest` helper
→ direct `gomask.MaskGoNonCode` call), fixed a test-fixture bug (missing
non-ASCII adjacency in a lookaround test string), quality-gated, committed
`f0d2915`.

## 044.011-T summary (this session, the bulk of the work)
Wrote `scripts/lib/gatecheck-run.sh` (C-3 shared runner: `gatecheck_build`,
`gatecheck_invoke`, `gatecheck_cleanup`) plus `tools/gatecheck/runner_test.go`
and OS-specific signal helpers (`runner_signal_unix_test.go` /
`runner_signal_windows_test.go`, needed because `syscall.SysProcAttr.Setpgid`
does not exist on Windows).

**Root cause of an extended debugging detour**: `toBashPath()` (converts a
Windows drive-letter path to MSYS POSIX form for a PATH-shim fake-bin
directory) had an off-by-slice bug — `p[len(m[0])-2:]` instead of
`"/" + p[len(m[0]):]` — which silently produced `/c:/Users/...` instead of
`/c/Users/...`. This caused the fake `go` PATH shim to never resolve, so
tests were silently invoking the REAL system `go` instead — and one test
(`TestGatecheckBuild_Failure_WrapperExits2WithErrorLine`) coincidentally
"passed" anyway because a real `go` failing on a missing `go.mod` produces
the same exit-2 signature the test was asserting for a deliberately-failing
fake `go`. Fixed by correcting the slice math. Confirmed via a from-scratch
minimal repro (`dbgtool.go`/`dbgtool2.go`, deleted after use) before touching
the real test file, to isolate Go's `exec.Command` env-passing semantics
from the MSYS PATH-auto-translation-at-bash-startup behavior established in
the prior session.

Also fixed `TestGatecheckBuild_DifferentWorkingDirectory_StillAnchorsAndPassesRoot`:
bash's own `cd ... && pwd` (used to derive `GATECHECK_SRC`) returns an MSYS
POSIX path that, for paths under the Windows TEMP directory, uses Git-Bash's
special `/tmp` mount alias rather than a plain drive-letter conversion — so a
regex-based reverse conversion cannot reproduce it. Fixed by round-tripping
through bash's own `pwd -W` (Windows-only branch, `runtime.GOOS == "windows"`)
and comparing by filesystem identity (`os.SameFile`) rather than string
equality, since the current user's profile directory also exhibited a
short-name (8.3, `DEWILL~1`) vs long-name (`dewilliams`) alias mismatch on
this box. POSIX platforms compare `os.SameFile` directly (no `pwd -W`).

**Manual out-of-band verification of the INT/TERM signal-trap tests**: since
Windows CI/dev-loop cannot exercise `runner_signal_unix_test.go` (it's
`!windows`-tagged and Windows uses the stub that reports
`signalTestsSupported() == false`), the underlying shell-script trap/cleanup
logic in `gatecheck-run.sh` was independently hand-verified end-to-end against
a real POSIX bash via WSL2 (`wsl.exe -e bash -c '...'`, using `set -m` +
`kill -INT/-TERM -- -$PGID` to replicate the Go test's process-group signal
delivery): confirmed exit 130 (INT) / 143 (TERM), no resumption past the trap
(`resumed.flag` absent), and `$GATECHECK_TMP` cleaned up in both cases. This
gives strong confidence the real Linux CI run of `runner_signal_unix_test.go`
will pass; it is otherwise entirely unverified on this Windows dev box (tests
report SKIP there, not PASS).

**Quality gates for 044.011-T**: `gofmt` (via the established LF-normalized-copy
technique) found and fixed a real struct-tag alignment issue in
`wrapperOpts` (not a CRLF false positive this time). `go vet ./tools/gatecheck/...`
clean both native and `GOOS=windows`. `go build ./...` and `go test ./...`
(full repo suite, ~170s) both clean. Committed as `5558dd6`. Marked
`044.011-T` done; commit tracked via `backlogit update 044.011-T --commit 5558dd6`.

## 044.012-T summary (this session, M1-T10 — final task, all 12/12 done)
Replaced `tools/gatecheck/register_write_path.go`'s stub with dispatch to
`internal/writepath.Run` (forwards the first remaining positional arg as
the mode flag). Reduced `scripts/check-write-path-precondition.sh` to
argument parsing plus `gatecheck_build` / `gatecheck_invoke write-path
"$@"` with the C-3 three-trap set, sourcing `scripts/lib/gatecheck-run.sh`
(M1-T9). No Python interpreter dependency remains in the wrapper.

**Local environment gotcha discovered and worked around (not a code fix)**:
`bash` on `PATH` resolves to Windows' built-in WSL shim
(`C:\Windows\system32\bash.exe`, GNU bash 5.3.9,
`x86_64-pc-linux-gnu`) ahead of Git for Windows' own bash
(`C:\Program Files\Git\bin\bash.exe`). The WSL shim does not perform the
Windows-path argument translation `gatecheck-run.sh` depends on, causing
every `runner_test.go` wrapper-invocation test to fail with exit 127 and a
mangled (backslash-stripped) path. Fixed for this session by prepending
`C:\Program Files\Git\bin` to `PATH` before every `go test`/`bash`
invocation; not a code change, has zero effect on CI (`ubuntu-latest` has
only one `bash`). Recorded in `m1.md`'s M1-T10 environment-precheck note.

Added two CLI-level dispatch tests to `tools/gatecheck/main_test.go`
(`TestRun_WritePath_DispatchesToEngine`, `TestRun_WritePath_BogusFlag`)
proving `run()`'s `--root` parsing and positional-arg forwarding wire
correctly into the engine (the engine's own per-mode/per-fixture behavior
is already exhaustively covered by `internal/writepath`'s own test suite).
Narrowed `TestRun_EachStub` to the three sub-commands still stubbed
(`write-path` is no longer one).

**Evidence captured in `m1.md` (M1-T10 section)**: ordered stdout/stderr
bytes + exit codes identical parent (`5558dd6`) vs. head across repo mode,
`--self-test`, `--self-test-integrity` and a bogus flag (zero ED-3/ED-5
delta — M1-T1 fixtures were already present in the parent commit); a
deliberately broken `tools/gatecheck` (temporary syntax error in
`main.go`, restored immediately after) exits 2, never 0;
`git grep -n -E '\bpython3?\b' -- scripts/check-write-path-precondition.sh`
is empty; `ci.yml` confirmed unchanged; live-corpus masked-output SHA-256
parity is 100% across all 83 tracked `cmd/**`/`internal/**`/
`scripts/testdata/**` `.go` files (compared via temporary generator
scripts on both the Python and Go side, both deleted after the comparison
was recorded — same convention as the M1-T2/T3 goldens). Environment
precheck recorded: `go1.26.5 windows/amd64`, `git 2.55.0.windows.5`, Git
for Windows bash, Python 3.14.3.

Committed as `b352754`. Marked `044.012-T` done; commit tracked via
`backlogit update 044.012-T --commit b3527549eef811597012b3cda3abf9e5b33c4acc`.

**All 12/12 tasks in feature 044-F are now done.** Feature 044-F and
shipment 034-S remain `active` — closing them is Step 6 (post-merge
closure), out of reach this run since merge is not pre-authorized
(P-017 dark mode).

## Shipment closeout summary (final, this session)

1. **Quality gates**: full sequence green (gofmt via `git archive`-extracted
   tree, `go vet ./...`, `go build ./...`, `go test ./...`,
   `go test -race ./tools/gatecheck/...`). Two genuine pre-PR/CI defects
   were found and fixed (both P-021 C1 same-contract-surface completions of
   this shipment's own work, not scope expansion — no deferred-scope-
   expansion stash entries were needed):
   - `9df7dfb`: renamed 3 deliberately-malformed `filebased/*.go` reference
     fixtures (M1-T3) to `*.go.txt` — they broke the repo-wide `gofmt -l .`
     CI step (parse errors / BOM). Content unchanged; nothing reads them by
     path at runtime. See evidence doc §8.
   - `de005ff` (fix-ci cycle 1 of 5): M1-T10's Go-dispatch rewrite of
     `check-write-path-precondition.sh` broke a pre-existing Python test
     (`test_shared_masker_is_imported`, from an earlier shipment's
     032.005-T) that asserted a now-removed `from gomask import ...` line.
     Replaced with two successor assertions preserving the original intent
     (no duplicated masker logic in the wrapper). Confirmed in-scope via the
     migration plan's R-13/M4-T6 scoping (`scripts/lib/tests/` stays a live
     blocking gate through M1-M3; full retirement is M4-T6's scope, not
     M1's). See evidence doc §9.
   - `715063d`: docs-only commit recording the fix-ci evidence (§9 above).
2. **Local review gate**: verdict **READY**, 0 P0/P1 findings, across the
   full 18-commit branch diff. D-4/`C312BD4C` masker multiline-struct-tag
   golden pin confirmed intact (pinned, not "fixed"). All 4 named
   out-of-scope stash entries (8E9F8E55, B72E9715, 8E18CCF5, 56B16321)
   confirmed NOT implemented. No P-021 deferred-scope-expansion captures
   were required — both fixes above were in-scope completions.
3. **PR #81** created: base `main`, head
   `feat/034-s-gate-engine-go-migration-m1-gatecheck-scaffold-pysem-masker-write-path`.
   Body includes the `## Local Review Readiness` block per §1.9, updated
   twice as HEAD advanced through the fix-ci cycle. Final reviewed HEAD:
   `715063d533140422d1bce5da785e4e6c20e826af`.
4. **CI**: RED on first push (`lint` job — the Python test regression
   above), fixed in fix-ci cycle 1, GREEN on the re-run (`715063d`) across
   all required checks: `lint`, `test`, `test (windows, advisory)`,
   `security`, all 4 `cross-compile` targets, plus the repo's structural
   gates (`detect code changes`, `gitignore append-only + un-ignore
   regression (I6)`, `merge-strategy structural verification`,
   `pipeline-topology (ambient)`, `load cross-compile targets`, `ci gate`).
5. **P-018 copilot-review gate**: `autoharness gate copilot-review 81
   --repo softwaresalt/intercom --enforcement auto --max-wait 0` returned
   `NOT_APPLICABLE: PASS` (exit 0) both immediately after CI went green and
   again on the final HEAD after the docs-only commit — Copilot review is
   not engaged on this PR and `copilot_review.enforcement` is unset
   (defaults to `auto`), so the gate does not hold merge.
6. **No reviews, no review threads, no PR comments** exist on PR #81 as of
   the final HEAD — nothing to resolve.
7. **STOP at merge readiness** — dark mode, `merge_approval_pre_authorized:
   false`, `admin_fallback_pre_authorized: false`. No merge attempted, no
   `--admin` used. Shipment 034-S and feature 044-F remain `active` in
   backlogit (all 12/12 tasks `done`) — post-merge closure (P-020
   compaction, closure doc, shipment archive) is explicitly out of reach
   this run and will only proceed after an operator-approved merge in a
   future session.

## Circuit breaker status
No circuit breakers tripped this session. No consecutive task failures.
Fix-ci cycles used: 1 of 5 (resolved the only CI failure encountered).
Review-fix cycles: 0 (local review gate returned READY on first pass, no
review-driven fix cycle was needed). No P-021 deferred-scope-expansion
stash entries were created — both in-scope fixes passed the C1
same-contract-surface test.

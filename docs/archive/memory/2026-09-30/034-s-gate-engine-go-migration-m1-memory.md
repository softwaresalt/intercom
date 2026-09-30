---
title: "034-S session memory — gate-engine Go migration M1"
date: 2026-09-30
shipment: 034-S
feature: 044-F
mode: dark-mode (P-017), merge_approval_pre_authorized=false
---

# 034-S session memory (checkpoint 1)

## Branch / topology state

- Branch: `feat/034-s-gate-engine-go-migration-m1-gatecheck-scaffold-pysem-masker-write-path`
- Base: `main` @ `5efa25ecabdf72e106e0112c96dfbfd0fb9da343`
- Shipment `034-S`: claimed, `active`. Sole active shipment (post-claim topology verified).
- Commits so far:
  - `1f11d00` — `test(ci): add write-path struct-tag selector characterization fixtures` (044.001-T)
  - `83d99c9` — `test(ci): capture masker, pysem, and write-path goldens from Python engine` (044.002-T, 044.003-T)

## Task status (backlogit ground truth as of this checkpoint)

| Task | Status |
|---|---|
| 044.001-T | done (archived) |
| 044.002-T | done (archived) |
| 044.003-T | done (archived) |
| 044.004-T through 044.012-T | `active` in backlogit (see anomaly note below) — none of their actual implementation work has started |

**Anomaly note (non-blocking)**: after `backlogit move 044.00{2,3}-T --status done`, all 9
remaining sibling tasks under 044-F flipped from `queued` to `active` in backlogit, including
tasks whose declared dependencies are NOT yet done (e.g. `044.010-T` depends on `044.008-T`,
`044.003-T`, `044.009-T` — only `044.003-T` is done). This does not correspond to
dependency-based unblocking; it appears to be a parent-feature-active → all-children-visible
side effect in backlogit's status model, not a genuine multi-claim. Dependency enforcement is
presumably still gated at `move --status done` time via the pre-completion gate broker
(`autoharness gate check`, exit 6/7/8 per `backlogit move --help`). Ship continues to execute
tasks strictly one at a time, in the plan's dependency order (§8), regardless of this cosmetic
status field. Flagged for the operator/Orchestrator in the final report; not treated as a stop
condition since it does not affect correctness or concurrency safety (Ship remains the sole
executor, single-threaded, one task at a time).

## Environment discoveries (local dev machine only, recorded for reproducibility)

1. Bare `bash` on PATH resolves to WSL2 (`C:\Windows\System32\bash.exe`), which has no Go
   toolchain and a separate path namespace. All local script/tool invocations that need `go` use
   the explicit Git for Windows path `C:\Program Files\Git\bin\bash.exe` instead.
2. Git for Windows bash's `command -v python3` resolves to the broken Windows Store
   app-execution-alias stub before ever reaching the real `python.exe` (this machine has no
   `python3.exe`). Workaround: exclude the `WindowsApps` alias directory from the subprocess PATH
   used for golden capture (not a repo change).
3. Windows-native `python.exe` applies `\n`→`\r\n` universal-newline translation on stdout/stderr
   even under git-bash; this is an artifact of the interpreter being a Windows build, unrelated to
   which bash flavor invokes it. Golden captures normalize `\r\n`→`\n` so the golden reflects
   portable (Linux CI) behavior, not a local quirk. Both discoveries are documented in
   `docs/plans/evidence/2026-09-28-gate-engine-go-migration/m1.md` §2.
4. **Implication for 044.011-T (shared bash runner)**: local red/green test verification for
   `scripts/lib/gatecheck-run.sh` (once written) must be run via
   `C:\Program Files\Git\bin\bash.exe` explicitly, not bare `bash`, for `go build` to succeed on
   this machine. CI (GitHub Actions, presumably ubuntu-latest) is unaffected — native Linux bash
   has `go` on PATH directly.

## Completed work (044.001-T, 044.002-T, 044.003-T)

- 044.001-T: two new write-path fixtures (`reject-struct-tag-selector.go`,
  `accept-non-tag-raw-string-selector.go`) verified against the current Python engine via
  `--self-test`; both gofmt-clean on LF-normalized content.
- 044.002-T: `internal/gomask/testdata/masker_golden.json` (42 entries) and
  `internal/pysem/testdata/pysem_golden.json` (17 categories), both verified byte-reproducible
  across two generator runs.
- 044.003-T: `internal/writepath/testdata/writepath_golden.json` +
  `internal/writepath/testdata/filebased/*.go`, verified byte-reproducible, stream captures
  inspected line-by-line for correctness (self-test PASS lines, unknown-mode usage message on
  stderr with exit 2, subdirectory-invoked repo scan clean at exit 0 confirming C-3 root
  anchoring, BOM passthrough behavior, invalid-UTF-8 decode-error capture for ED-2).
- Evidence doc `docs/plans/evidence/2026-09-28-gate-engine-go-migration/m1.md` records the
  environment precheck, both transient generator scripts verbatim, exact commands, and
  reproducibility proof. Both transient generator scripts deleted from
  `C:\Users\Public\` after evidence capture; confirmed no `.py` file added to the repo.

## Next steps

1. Claim 044.004-T: scaffold `tools/gatecheck/main.go` (`run()` dispatch, exit contract, 2/1
   split — bash wrappers own exit 2, `main` owns exit 1 via panic recover), sub-command stubs
   (`register_<name>.go` per plan-named sub-command), red-then-green table tests in
   `main_test.go`.
2. 044.005-T through 044.008-T: `internal/pysem/{classes.go, digit_table.go, case.go, text.go,
   io.go}` verified 100% against `pysem_golden.json` (needs its own small transient digit-table
   generator, not committed, per the same C-5 discipline).
3. 044.009-T: `internal/gomask/gomask.go` port, verified against `masker_golden.json`.
4. 044.010-T: `internal/writepath/writepath.go` port, verified against `writepath_golden.json`.
5. 044.011-T: `scripts/lib/gatecheck-run.sh` shared bash runner + `tools/gatecheck/runner_test.go`
   (red tests first).
6. 044.012-T: switch `scripts/check-write-path-precondition.sh` to Go, parity evidence.
7. Quality gates (gofmt on LF-normalized content, `go vet ./...`, `go test ./... -race` where
   applicable, `go build ./...`), local review gate, PR creation, CI, P-018 copilot-review gate,
   P-014 §1.9 readiness gate, then STOP for explicit operator merge approval (dark mode:
   `merge_approval_pre_authorized=false`).

## Circuit breaker counters

- Tasks attempted this session: 3 of 12 complete (044.001-T, 044.002-T, 044.003-T). Well under
  the 20-task session limit.
- Consecutive task failures: 0.
- Review-fix cycles: none yet (no PR/review gate reached).

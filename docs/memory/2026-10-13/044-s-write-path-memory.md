# 044-S session memory — write-path gate completeness and CI-log output safety (054-F)

- **Date:** 2026-10-13 (session start 07:24 PDT; commit timestamps are UTC)
- **Shipment:** 044-S (queued -> active via `backlogit shipment claim`; post-claim topology gate exit 0)
- **Covering feature:** 054-F (activated by the shipment claim cascade)
- **Branch:** `feat/044-s-write-path-gate-completeness-and-ci-log-output-safety` from `main@bf141ae`
- **Operator mode:** DARK_MODE_ACTIVE (P-017), scope 044-S only. Merge pre-authorized for this scope if §1.9 and P-018 pass. `--admin` NOT authorized.
- **Visibility:** agent-intercom not exposed in this session (no broadcast tool); status recorded here and in the final report.
- **Backlogit access:** CLI fallback only (backlogit MCP tools not exposed in this session). `backlogit sync` at intake: `Indexed 549 artifacts`.
- **Engram:** bound to `C:\Source\GitHub\intercom-go`, branch main at intake, scan complete, `stale_files: false`.

## Pipeline state

| Step | Result |
|---|---|
| Step 0.0 tool probe | backlogit CLI OK (MCP not exposed; CLI fallback). engram OK. gh OK (softwaresalt). autoharness OK. |
| Step 0.1 index sync | `INDEX_SYNC_OK` (CLI) |
| Step 0.5 item 1a early warning | all manifest tasks `queued`, record `queued`: no `SHIPMENT_STATE_INCONSISTENT` |
| Step 0.5 item 3a topology pre_claim | exit 0 (`BRANCH_CREATE_ELIGIBLE`, `WORKTREE_TOPOLOGY_OK`, `WORKTREE_CLEAN`, predecessor 043-S) |
| Step 0.5 item 3b shipment-reconcile pre (expected queued) | PROCEED. Report `.backlogit/reconcile/044-S-pre-2026-10-13T14-30-00Z.md` |
| Predecessor 043-S | archived, `archived_status: shipped` |
| P-001 | no other active release units; no open PRs at intake |
| Claim | `backlogit shipment claim 044-S` exit 0; post_claim topology gate exit 0 (`CLAIM_VERIFY_OK`) |

## Commits so far (feature branch)

| SHA | Subject | Unit |
|---|---|---|
| 56d977e | chore(docs): claim shipment 044-S for feature 054-F | backlog |
| 0c87ebd | test(ci): add failing escape-helper rows with compiling stub | 054.001-T (U1 red) |
| dd20501 | fix(ci): escape attacker-influenced text at the write-path log sink | 054.002-T (U2 green) |
| 6ba583d | test(ci): add failing call-site escaping and PathError rows | 054.003-T (U3 red) |
| 8c519b2 | chore(docs): complete 054.001-T..054.003-T and label 054.004-T harness-ready | backlog |
| 1850d2f | test(ci): align U+009B finding expectation with the xNN escape rule | U3 correction |
| 0e913df | fix(ci): escape every gate output sink and unwrap PathError | 054.004-T (U4 green) |
| c01e890 | chore(docs): complete 054.004-T and record 054.005-T harness pending | backlog |
| fab3dfb | test(ci): add failing non-selected record containment rows | 054.005-T (U5 red) |
| 110a033 | chore(docs): complete 054.005-T and record 054.006-T harness pending | backlog |

## Decisions and interpretations (for review)

1. **Harness step runs per pair, not wholly up front.** The plan's red units are sequenced after the green units they characterize (U3 blocked by U2, U5 by U4, U5b by U5, U7 by U6), so each red unit is scaffolded immediately before its pair. The operator directive (TDD, commit per task, one at a time) supports this. Red tasks go through the harness-architect step; green tasks go through build-feature after the paired red task carries `harness-ready`.
2. **Red-phase marker.** The harness-architect default is `panic("not implemented")`. This plan specifies compiling stubs that return input, plus behavioural red rows, so each red row fails for the reason stated in its task AC. Each test comment documents this.
3. **Escape format.** Runes at or below 0xFF escape as `\xNN`; at or below 0xFFFF as `\uNNNN`; otherwise `\UNNNNNNNN`. Invalid UTF-8 bytes escape as `\xNN`. Backslash is never escaped. This matches the U1 table and the U2 implementation. Go's `strconv.Quote` would render C1 runes as `\uNNNN`; the plan does not pin that choice, so the committed rule stands. The choice is cosmetic (no workflow-command risk).
4. **U3 expectation defect corrected** (commit 1850d2f). The U+009B finding row expected `\u009b` while the U1 rule gives `\x9b`. Fixed as a one-line test-unit correction. The row stays red at the test commit on the raw-control check.
5. **Linux-only red rows.** Windows cannot run `linux || darwin` rows. They were executed by cross-compiling the test binary (`go test -c` with GOOS=linux) and running it under WSL Ubuntu from the gitignored scratch directory. CI's Linux job remains the authoritative evidence.
6. **Backlog completion timing.** Each task is marked `done` at its own commit (Step 4.5), which archives its queue record. Backlog state goes in separate `chore(docs)` commits.

## Gate evidence (local)

- Baseline on Windows (main, before changes): the `tools/gatecheck` root package has 19 pre-existing failing top-level tests (bash temp-path translation on this host; CI Linux is authoritative). All `internal/*` packages pass. The failing set is identical after U2 (diffed programmatically).
- After U2: `go vet ./...`, `gofmt` on tracked files, `go build ./...`, `go test ./tools/gatecheck/internal/writepath/`, and `GOOS=linux`/`GOOS=darwin go vet ./tools/gatecheck/...` all pass. Full module `go test ./...` shows only the 19 pre-existing Windows wrapper failures.
- After U4: writepath package PASS on Windows and under Linux (WSL). Repo-scan wrapper (`scripts/check-write-path-precondition.sh`, blocking CI step) exit 0. `--self-test` exit 0. `--self-test-integrity` exit 0. `check-retired-architecture.sh --self-test` exit 0. Run via Git Bash (WSL bash cannot see the Windows `go`).
- After U5 (red): Windows fails only `TestRunRepoScan_NonSelectedRecordContainment` a1-a5. Linux fails the same set (`Code = 0, want 1`). Controls and the case-variant pin pass.

## Containment and hygiene incidents (self-reported)

- The U1 subagent wrote a throwaway `isprint-check` program under `.copilot/session-state/.../files/` (gitignored, inside the repo tree). It is left in place. It breaks an unfiltered local `gofmt -l .`, so local gofmt runs over `git ls-files '*.go'`. Not deleted (operator may remove).
- I copied the Linux test binary to WSL `/tmp/wp.test` and removed it. Later subagents were told not to use `/tmp`; the U5 subagent redirected test temp directories into the scratch folder.
- I wrote a temporary comparison file under `%TEMP%` and deleted it.
- I wrote a self-test log to MSYS `/tmp_wp_self.txt` (outside the repo) and deleted it.
- The U3 subagent's Linux recipe copied the binary to `/tmp`; removed.
- Two premature harness labels/statuses (054.005-T `harness-ready`, 054.006-T `scaffolded`) were reverted to `pending` before any harness existed. No commit contained the `harness-ready` label for 054.005-T before its red commit.

## Units completed (all nine tasks done; 054-F active until post-merge closure)

| Commit | Subject | Unit |
|---|---|---|
| 2d24d94 | test(ci): pin outcome order, hostile-name escapes and refusal labels | review fix batch 1 |
| f488297 | docs(ci): record the Lstat-to-read TOCTOU residual in the gate docs | review fix batch 1 (docs) |
| 767b484 | test(ci): pin matrix rows 1-2 order and label green-on-arrival rows | review fix batch 2 |
| cc0f96a | docs(ci): name the TOCTOU residual without a decision ID | review fix batch 2 (docs) |
| bfd0f1d | chore(docs): capture deferred scope expansions from the 054-F review | P-021 captures |
| c70ea4a | chore(docs): capture accepted review follow-ups for the 054-F gate | P-021 captures (3, circuit breaker) |

Earlier units: 56d977e claim; 0c87ebd U1 red; dd20501 U2 green; 6ba583d U3 red; 1850d2f U3 expectation fix; 0e913df U4 green; fab3dfb U5 red; e693e43 U5b red; 2f2a22c U6 green (containment, literal pathspecs); 8323b52 U7 red; 7bb4780 U8 green; backlog moves in 8c519b2, c01e890, 110a033, 396a57d.

## Review gates (P-014 / operator requirement)

- Adversarial Review (3 reviewers, report-only) on 396a57d: READY_WITH_FOLLOWUPS. No P0; one LOW P1 (single source, test-only) fixed in batch 1. Escalated items recorded.
- Standard review skill (report-only, personas) on 396a57d: READY_WITH_FOLLOWUPS; P2/P3 items: test batch and docs batch applied; code batch items declined as out of scope and captured.
- Post-remediation re-review on f488297 (report-only): READY_WITH_FOLLOWUPS; F1-F4 applied in batch 2.
- Final review on bfd0f1d (two reviewers: Adversarial Review pool and review skill, report-only): READY_WITH_FOLLOWUPS, no P0, no P1 confirmed. Remaining P2/P3 accepted as follow-ups (circuit breaker: three review-fix cycles reached) and captured.
- Delta bfd0f1d..HEAD: `.backlogit/stash.jsonl` only (no code change).

## Stash captures (P-021 C5, capture-only; no later edits)

85D5A4D5 (static literal lines), 054EAAB0 (fixture early-return stdout), A6535A74 (test hygiene, scannedPathsFromListing placement), 8E8B65CB (percent sign in escapeLogText, deliberation required), 88036DC5 (ReadDir absolute --root, deliberation required), AF0C83BA (matrix wording and retiredarch comment), 585FCD6D (residual env and reparse verifications), B7032DCB (dead and duplicate code, decision record amendment), A1327CF4 (documentation accuracy follow-ups), 5B1F37E7 (pre-existing residuals; stash metadata triage note for Stage).

Stage triage notes (Ship did not edit captured entries): the structured `priority` of 85D5A4D5, 054EAAB0, A6535A74, 8E8B65CB, 88036DC5, AF0C83BA, 585FCD6D was set by the CLI default (medium) while each payload states provisional priority low (8E8B65CB: medium). Stage should reconcile. Existing entries already cover the escapeLogText retiredarch copy, main.go echoes (SEC-8), nested-tree ED-7 (49982A4E), and promoting task-gated writepath harness tests into the always-on suite; these were not duplicated.

## Known test-environment limits (recorded, not defects)

- Windows: 19 pre-existing tools/gatecheck root wrapper tests fail on this host (bash cannot resolve Windows temp paths). Identical set before and after this change. CI Linux is authoritative.
- WSL Ubuntu on /mnt/c cannot create FIFOs: TestRunFixtureSelfTest_FIFOFixtureRefused fails there with mkfifo operation not supported. Covered by the CI Linux job; not executed locally.
- WSL has no Go toolchain; Linux test binaries were cross-compiled on Windows and run under WSL from the gitignored scratch directory.

## Next

- Final full-module suite on c70ea4a (running at time of writing); compare the failing set with baseline (19 Windows-only).
- Push branch; create PR with `## Local Review Readiness` block (reviewed content HEAD bfd0f1d; PR head c70ea4a; delta backlog-only).
- Copilot review request, back-off polling, thread-by-thread fix/reply/resolve loop, P-018 gate, §1.9 gate, merge commit only on PASS and operator-authorized scope (DARK_MODE merge pre-authorized; admin fallback NOT authorized).
- Post-merge closure: shipment-reconcile pre (expected done), classify-only cascade-close, safe-close via shipment-reconcile (no direct `backlogit shipment ship`), closure artifact, closure PR through the same loop, P-020 compact-context.

# 044-S / 054-F closure session: compacted (P-020)

- **Release unit:** 044-S (shipment), covering feature 054-F (write-path gate completeness and CI-log output safety)
- **Period:** 2026-10-13 to 2026-10-14 UTC (session 1: intake through the PR 133 merge; session 2: resumed closure)
- **Outcome:** PR 133 merged as merge commit `19ad28f0d79bdfe0b0bd211456d0f74edaf9a7b6` (P-009). Shipment 044-S closed on the CASCADE path (`autoharness shipment cascade-close`, exit 0, postconditions pass). The closure artifact is `docs/closure/044-S-054-F-post-merge-closure.md` (READY_WITH_CONDITIONS).
- **Verbose originals (traceable, moved, not deleted):** `docs/archive/memory/044-s-write-path-memory.md`, `docs/archive/memory/044-s-closure-halt-memory.md`, `docs/archive/memory/044-s-closure-resume-engram-halt-memory.md`, `docs/archive/memory/044-s-closure-resumed-session-memory.md`

## Session 1: intake to merge (054-F build)

**Decisions**
* Harness scaffolding ran per red/green pair, not wholly up front. Red units were sequenced after the green units they characterize. Each red task went through harness-architect, then each green task went through build-feature after its paired red task carried `harness-ready`.
* Red-phase markers were compiling stubs that return their input plus behavioural red rows. The harness-architect default `panic("not implemented")` was not used, so each red row fails for the reason stated in its acceptance criteria.
* Escape format fixed: runes at or below 0xFF as `\xNN`; at or below 0xFFFF as `\uNNNN`; otherwise `\UNNNNNNNN`. Invalid UTF-8 bytes as `\xNN`. Backslash is never escaped.
* Each task was marked done at its own commit (Step 4.5), which archived its queue record. Backlog state went in separate `chore(docs)` commits.

**Corrections and outcomes**
* U3 expectation defect fixed in `1850d2f`: the U+009B row expected `\u009b` while the rule gives `\x9b`.
* Linux-only red rows could not run on Windows. They were executed by cross-compiling the test binary and running it under WSL from the gitignored scratch folder. CI Linux stays authoritative.
* Premature `harness-ready` and `scaffolded` labels on 054.005-T and 054.006-T were reverted to `pending` before any harness existed. No commit carried the label before its red commit.
* Nine tasks reached `done`. Commits of note: `2f2a22c` (containment with literal pathspecs), `0e913df` (escape every gate output sink), `7bb4780` (refuse `.go`-named directories in the fixture self-test).

**Review and gate record**
* Adversarial review on `396a57d`: READY_WITH_FOLLOWUPS, no P0, one LOW P1 (test-only) fixed. Standard review on `396a57d`: READY_WITH_FOLLOWUPS. Post-remediation re-review on `f488297`: READY_WITH_FOLLOWUPS. Final review on `bfd0f1d` (two reviewers): READY_WITH_FOLLOWUPS, no P0 and no P1.
* Circuit breaker reached three review-fix cycles. Remaining P2/P3 items were accepted and captured under P-021.
* PR 133 head `76db3af`: Copilot gate SATISFIED on the head with zero threads and not forced. All 14 CI check runs SUCCESS (workflow run 38433778306).

**Captures (P-021 C5, capture-only, no later edits)**: `85D5A4D5`, `054EAAB0`, `A6535A74`, `8E8B65CB` (deliberation required), `88036DC5` (deliberation required), `AF0C83BA`, `585FCD6D`, `B7032DCB`, `A1327CF4`, `5B1F37E7`. Stage owns priority reconciliation: the CLI default `medium` conflicts with payload text `low` on several entries.

**Known environment limits (not defects)**
* Windows: 19 pre-existing `tools/gatecheck` root wrapper tests fail on this host because bash cannot resolve Windows temp paths. The failing set was identical before and after the change.
* WSL Ubuntu on `/mnt/c` cannot create FIFOs, so `TestRunFixtureSelfTest_FIFOFixtureRefused` fails there. Covered by the CI Linux job.

**Containment incidents (self-reported, handled)**
* A subagent left a throwaway `isprint-check` program in the gitignored `.copilot` scratch tree. It breaks an unfiltered local `gofmt -l .`, so local gofmt runs over `git ls-files '*.go'`.
* Temporary files written to `/tmp` (WSL) and `%TEMP%` were removed. The U5 subagent redirected test temp directories into the scratch folder.

## Session 2: resumed closure (cascade close, closure artifact)

**Blocker and resolution**
* Engram daemon failure in session 1 (orphaned workspace daemons; `daemon-status` did not reach Ready within 30 s). Operator cleared the processes. A fresh bounded probe returned overall `green`. `workspace-status` and `query-memory` returned exit 0. No index, sync or stats call was made.
* Git launcher defect (environment): `C:\Program Files\Git\cmd\git.exe` needs `mingw64\bin\git.exe`, which is missing. This session used Git bundled with GitHub Copilot (`2.53.0.windows.4`) by prefixing its `cmd` directory to `PATH` for each command. The system install was not modified.

**Checkpoint 165942 resume**
* Selected by filename by the operator. Restore: `valid: true`, `conforming: true`, `agent: ship`. Unfiltered enumeration: 87 checkpoints, zero anomalies.
* Prune-on-restore: the restored checkpoint is already a bounded structured summary with no superseded action-observation history. Nothing was dropped. The active cursor, the checkpoint pointer and the recorded gate verdicts were preserved.
* Checkpoint `150837` (ship, build phase, stale resume hint for 054.006-T) was NOT selected, NOT touched and NOT resolved. Operator decision pending.
* Resolve of `165942` is deferred until after the closure PR merges, so the recovery pointer stays available through the destructive run and the closure PR.

**Cascade close**
* Classify-only record from session 1 (`51cba89d7ac2440880f41ae4bafadd76`, `CASCADE`, `VERIFIED`, `selected_close_path: cascade`) was reused without replacement (no-clobber). Sha256 of ten manifest members and of the shipment record matched the snapshot before the run.
* Attempt 1 with the `--message=<value>` / `--author=<value>` form: exit 2, `Unknown shipment cascade-close argument`, `mutation_possible: no`, nothing invoked. This matches `A744E584` (captured by 043-S). Not re-captured.
* Attempt 2 with the documented space-separated form and identical values: exit 0, `postcondition_verdict: pass`, `phase_written: post_close`, `returned_ids: []`, 11 of 11 items archived, `parent_id` preserved, disposition complete (zero dispositions).
* Post-mode reconcile: PROCEED. Zero archive deletions (P-007). Shipment lock acquired for pre-mode and released after post-mode.
* Commit `d13a621` archived the backlog and the evidence record (15 paths). Commit `d0df3a7` added the closure artifact and the three closure memories.

**Learnings (candidates for compound capture, where they are not already captured)**
* A skill's prescribed argument form can disagree with the CLI's own usage text. Check the CLI's `--help` before running a destructive command. Here the equals form was refused before any mutation, so it was safe to retry once with the documented form.
* A runaway or orphaned engram daemon makes a healthy-looking supervisor fail-closed. The fix is to probe with a bounded timeout and kill only PIDs the session itself spawned.
* A reconciling Git launcher can break silently when the bundled `mingw64` tree is missing. Probe `git --version` before the first repo command.

## Compaction disposition

* Candidates in this pass: the four 044-S memory files above (the just-closed release unit, per the skill's Tier-1 floor).
* Other memory folders (`2026-09-18` through `2026-10-12`) were out of this pass's scope and were not touched.
* Originals are moved, not deleted, to `docs/archive/memory/` using `git mv`, so the path back to the originals stays traceable.

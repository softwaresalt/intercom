---
title: 045-S merge-gate halt (dark factory run)
date: 2026-10-14
shipment: 045-S
feature: 055-F
pr: 135
reviewed_head: df7d8fd6baf46af3b698e5d50da49311b6c66a09
status: DARK_MODE_HALTED
---

# 045-S merge-gate halt

## Outcome

Shipment 045-S built, reviewed, and opened as PR #135. The run halted before merge.
The halt is a conditional-authority stop, not a gate failure. Every technical gate passed.

* Tasks 055.001-T through 055.007-T: all `done`. Commits f038032..df7d8fd (12 commits).
* PR #135 open, `headRefOid` df7d8fd, `MERGEABLE` and `CLEAN`, CI green (ci gate, cross-compile
  on four targets, lint, test, security, merge-strategy, topology, Windows advisory test).
* P-009: the repository allows merge commits only (`allow_merge_commit` true, squash and rebase
  false). Merge strategy is compliant.
* P-016 topology: `pipeline-topology` pre_claim, post_claim, and lifecycle all passed.
* P-018 copilot-review gate: `NOT_APPLICABLE` (exit 0). Copilot never engaged on PR #135, and
  the request route the protocol names (the Copilot review MCP tool) is not available in this
  session. No repository-approved fallback is configured.

## Why the run halted

The dark-mode merge authority for 045-S is conditional on the P-018 gate returning PASS for the
current HEAD. The operator also required that Copilot review be requested and worked to a clean
latest review, with every thread resolved. `NOT_APPLICABLE` means Copilot is not in play, so the
gate does not hold, but it does not satisfy the operator's condition either. No Copilot review
exists to resolve. Merging on `NOT_APPLICABLE` would bypass an explicit operator requirement, and
the stop list names ambiguous approval authority. The run stopped instead of guessing.

Admin fallback was not used. `--admin` is not pre-authorized.

## Local review record

Three rounds, seven reviewers in round one (Adversarial Review with anchor gpt-6.1-sol, Security,
Correctness, Go, Maintainability, Constitution, Scope, Architecture), then a delta round and a
final adversarial check.

* Round 1 verdict: adversarial BLOCKED on one P1 (no row proved the `-C` anchor). Other personas
  READY_WITH_FOLLOWUPS. Fixed by the decoy-cwd row (841d1b4); removing `-C` now turns it red.
* Round 2 (delta) verdict: adversarial BLOCKED on one P1 (readonly refusal untested) and one P2
  (diagnostic matched trace lines). Fixed by df7d8fd. Each new row was mutation-checked by hand.
* Round 3 (final): adversarial READY_WITH_FOLLOWUPS, no P0 or P1.
* Final outcome: `READY_WITH_FOLLOWUPS`, P0=0, P1=0.

Reviewer note on verification: the reviewers in this session had no shell, so their findings are
static reads. The mutation checks and every test run were executed by the Ship agent.

## Gates and evidence at df7d8fd

* `gofmt -l .` empty.
* `go vet ./...` exit 0 on host, `GOOS=linux`, and `GOOS=darwin`.
* `go build ./...` exit 0.
* `go test -p 1 ./...` all packages `ok`. The serialized run is needed on this host, because
  Git Bash shares the `/tmp` mount across concurrent packages. See the host note below.
* shellcheck clean on `scripts/lib/gatecheck-run.sh`. The two wrappers keep the same three
  pre-existing findings as main (SC1091, SC2034, SC2329) and gain no new ones.

## Mutation evidence (executed)

| Mutation | Rows that went red |
|---|---|
| Remove `-C` from `gatecheck_repo_root` | decoy-cwd no-repository row (exit 2, want 128) |
| Remove the readonly refusal arm | readonly `GIT_DIR` row |
| Remove the write-path caller precondition | caller-outside-work-tree row |
| Move the nested diagnostic off stderr | nested checkout rows, both wrappers |

Every mutation was restored byte-for-byte. The lib diff and the wrapper diff were checked against
the committed content afterwards.

## Red evidence at the red commits

* 9e42d16: decoy, mixed-case, subdirectory, and nested rows fail; ROOT pin fails. Control,
  symlink, and no-repo rows are green on arrival (characterization).
* fc2c86a: retired-arch decoy, mixed-case, and nested rows fail; ROOT pin fails. The subdirectory
  row is green on arrival. The plan expected it red. `git -C "$GATECHECK_SRC"` anchors the
  worktree, so `GIT_DIR` alone cannot mislead it. The row stays as a regression guard.
* 2cc3bbb: strip rows (`MSYS`, `msys`, `CYGWIN`, `Msys2_*`) fail in both packages. Keep and boundary
  rows are green on arrival.

## Copilot comments

None. No Copilot review exists on PR #135, so the comment counts are zero: fixed 0, declined 0,
captured 0.

## Stash captures (P-021 C2, capture-only)

New `DEFERRED SCOPE EXPANSION` entries:

* 83C1369E: Linux-discriminating scrub and config-pin recorder.
* 815C2B28: export-pin refusal row for an unexported readonly `GIT_CONFIG_GLOBAL`.
* C37F75D5: shell-side MSYS scope decision.
* BEFE6DF2: capture `compgen -e` before iterating (fail closed on a failed listing).
* EF11062A: move the shared wrong-tree harness to a harness file.
* 63BCA7C6: write-path header paragraph and PS4/LC_ALL hygiene.

Existing entries cited, not duplicated: 5A8EC1BC, 585FCD6D, 1B85D0F4, 4372BAD4, 49982A4E.

## Disclosed deviations

1. ded479b carries a test-support harness change (install the shim once per table). No assertion
   changed. Red state re-verified against the pre-change production files with the restructured
   harness.
2. Commit footers end with the Co-authored-by trailer requested by the operator. The literal
   `- Generated by Copilot` line appears before it.
3. 055.003-T AC1 expected the retired-arch subdirectory row red. It is green on arrival (see
   above). Acceptance criteria are planning fields, so Ship does not edit them. The deviation is
   recorded here and in the PR body.
4. Ship did not formally invoke the safety-modes skill for U10, U12, U14, and U15 (careful mode).
   The blast radius was enumerated by hand (scripts, gitRunnerEnv bodies, canonicalDecls), and each
   pair was gated before commit. Recorded as a process gap.

## Host note

On this Windows host `bash` on PATH is WSL. Tests ran with `C:\Program Files\Git\bin` first on
PATH. The Git Bash `/tmp` mount is shared, so shim-based wrapper tests can flake when packages run
in parallel. The gate ran serialized. CI runs on Linux, where `/tmp` is a real directory.

## Operator actions to resume

1. Request Copilot review on PR #135 through the approved route (the Copilot review MCP tool, or
   the GitHub UI). Then either let the P-018 gate reach PASS for `df7d8fd` (or a later HEAD after
   any fix), or explicitly authorize the merge on `NOT_APPLICABLE` in writing.
2. Resume at Step 5: re-run the P-018 gate, re-check head SHA and §1.9, then merge head-pinned
   with `gh pr merge 135 --merge --match-head-commit df7d8fd6baf46af3b698e5d50da49311b6c66a09`.
   No `--admin`.
3. Step 6 closure: post-merge branch and closure PR, `shipment-reconcile` pre and post gates,
   `autoharness shipment cascade-close --classify-only` first, and explicit per-closure approval
   before any mutating run. Dark mode does not satisfy that approval. Then P-020 compact-context.

## State left behind

* Branch `feat/045-s-gate-git-discovery-and-environment-isolation-across-wrappers-and-engines`
  checked out, pushed, PR #135 open. Ship stays on the feature branch while awaiting merge.
* Shipment 045-S and feature 055-F remain `active` (P-001 holds the next top-level unit). Tasks
  `done`. Task queue files for done tasks moved to `.backlogit/archive/` (carry-forward backlog
  state, not committed on the feature branch).
* Operator `config.yaml` routing edit: restored from stash `84aabe4` (popped). The stash entry is
  gone. Two unrelated stashes remain (038-S and 026-S).
* Stale checkpoint `checkpoint-20261013-150837.json` (ship, 044-S build): untouched. Needs an
  explicit operator decision.
* New checkpoints: `checkpoint-20261015-055107.json` (build milestone) and the merge-gate halt
  checkpoint created with this note.
* Working tree: backlog queue, archive, and stash.jsonl carry-forward changes; this memory note is
  untracked. None of them is staged.

## Compaction (P-020)

Closure has not started. The P-020 compaction status applies to the post-merge closure artifact,
which does not exist yet.

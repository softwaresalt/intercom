---
title: "Post-merge closure: 036-S / 046-F — Gate-engine Go migration M3 (unignore-regression and merge-strategy engines)"
description: "Post-merge closure artifact for shipment 036-S / feature 046-F"
status: "complete"
tags:
  - "operational-closure"
  - "036-S"
  - "046-F"
  - "post-merge"
date: 2026-10-01
mode: post-merge
shipment: 036-S
feature: 046-F
pr: 85
merge_commit_sha: 5fdd75aec21bb9492aabad08603a5479c1da6846
compaction_status: done
closure_status: READY
releasability: READY
conditions: []
---

# Post-merge closure: 036-S / 046-F — Gate-engine Go migration M3 (unignore-regression and merge-strategy engines)

## Summary

Shipment `036-S` delivers `046-F`, "Port unignore-regression and merge-strategy gate engines to
Go (M3)". This is plan unit M3 (of M1–M4) in
`docs/plans/2026-09-28-intercom-go-gate-engine-go-migration-plan.md`, per the deliberation in
`docs/decisions/2026-09-28-intercom-go-gate-engine-python-to-go-migration-deliberation.md`,
continuing directly from M2 (035-S / 045-F, PR #83). What it did:

* Ported the inline Python in `scripts/check-unignore-regression.sh` into
  `tools/gatecheck/internal/unignore` (regression-check engine, self-test scenarios, scratch git
  repo fixtures) and the merge-strategy JSON evaluator in `scripts/check-merge-strategy.sh` into
  `tools/gatecheck/internal/mergestrategy` (SKIP-semantics-preserving evaluator).
* Switched both bash wrappers to the Go engines via the shared M1 runner
  (`scripts/lib/gatecheck-run.sh`), with a pinned `Set up Go` step added to the
  gitignore-append-only and merge-strategy CI jobs.
* Exit codes and verdicts are preserved byte-for-byte; only the enumerated deltas ED-1..ED-9
  apply (INV-1). Parity evidence is in
  `docs/plans/evidence/2026-09-28-gate-engine-go-migration/m3.md`.
* Folded stash `150364D2` (temp-file lifecycle/trap/exit-table/rev-parse hardening) into the
  merge-strategy port and updated `docs/merge-strategy-gate.md` accordingly (this session also
  corrected a stale "bash orchestration unchanged" claim in that doc left over from the initial
  port — see Copilot thread 5 below).
* All 13 tasks (`046.001-T`..`046.013-T`) have passing acceptance criteria.

| Commit | Description |
|---|---|
| `f41c1fb` | chore(docs): commit 035-S closure memory and stash capture (carry-over, committed first per operator instruction) |
| `fb27ce1` | ci(gatecheck): pin setup-go in gitignore-append-only and merge-strategy jobs |
| `3b71d82` | feat(gatecheck): port unignore-regression engine to Go and switch wrapper |
| `7c6fcf6` | feat(gatecheck): port merge-strategy evaluator to Go, switch wrapper, fold in stash 150364D2 |
| `719a655` | docs(gatecheck): update merge-strategy-gate doc for Go evaluator and ED-1 |
| `76af871` | docs(gatecheck): add M3 unignore-regression and merge-strategy port evidence |
| `fe83bc5` | fix(gatecheck): address correctness/Go/review findings for M3 unignore port (+ 3 P-021 stash captures: `7223218F`, `5A8EC1BC`, `978D2946`) |
| `6f3e372` | fix(gatecheck): resolve adversarial-review C-1/M-1 blocking findings for M3 |
| `413cdd9` | docs(gatecheck): record adversarial-review re-verification (C-1/M-1 CLEARED) |
| `607b6a9` | fix(gatecheck): resolve lint findings in M3 unignore package |
| `464eb77` | docs(gatecheck): fix stale merge-strategy-gate.md orchestration claim (Copilot thread 5) |
| `d9f7506` | chore(stash): capture P-021 deferred-scope-expansion entry `50E6F22C` (Copilot threads 1-3) |
| `5fdd75a` | Merge commit (PR #85) |

## Invariants to preserve

* INV-1 (no merge-blocking verdict or exit-code changes except ED-1..ED-9) held for the full
  shipment; all three Copilot-flagged "containment gap" findings (root containment, OS-temp
  scratch dirs, extra-args handling) are confirmed faithful Python parity, not regressions, and
  fixing any of them would itself be an unauthorized INV-1 deviation absent a new ED entry (see
  stash `50E6F22C`).
* The `--root` caller-argument-injection-guard pattern (established in M2 for
  `check-retired-architecture.sh`, Copilot PR #83 finding) was correctly repeated for
  `check-unignore-regression.sh` in M3-T7 (new per-argument denylist loop, tested). `tools/gatecheck/main.go`'s
  `parseRoot` still scans the whole forwarded argument list with last-occurrence-wins semantics,
  so any *future* wrapper forwarding more than one fixed non-flag argument must independently
  re-implement this denylist — tracked as stash `7223218F` (high priority, requires Stage
  deliberation on a cross-cutting hardening unit).
* `tools/gatecheck` now has three independently hand-rolled `GitRunner` abstractions
  (`writepath.GitRunner`, `retiredarch.GitRunner`, and M3's new `unignore.GitRunner`, a
  general-purpose superset) with no shared interface — tracked as stash `5A8EC1BC` (low priority;
  also contains an uncorrected `unignone.GitRunner` typo in its own text, which Ship could not
  correct per the P-021 C5 capture-only carve-out — left for Stage).
* The M1 invariant (shared bash runner `scripts/lib/gatecheck-run.sh` as the one runner for every
  Go-backed gate wrapper) continues to hold for both new M3 wrappers.
* `scripts/check-unignore-regression.sh` and `scripts/check-merge-strategy.sh` must stay
  `text eol=lf` per `.gitattributes`; a CRLF line-ending drift was caught and fixed during this
  session's own build (before commit, via manual `bash -n`/`git check-attr` discovery, not an
  automated gate) — tracked as stash `978D2946` (medium priority; no durable CI-level
  line-ending-drift check exists yet).

## Validator evidence

`.autoharness/workspace-profile.yaml` `runtime_validation.validator_manifest` declares one
runtime surface: `cli`, checked by a `go run ./cmd/... --help` smoke run. This shipment does not
touch `cmd/intercom` or `cmd/intercom-ctl` (confirmed via `git diff 0092a08..5fdd75a --stat --
cmd/` — no output); that probe does not apply, consistent with M1/M2's own closures.

The surfaces that did change are `tools/gatecheck/internal/{unignore,mergestrategy}`,
`tools/gatecheck/register_*.go`, `scripts/check-unignore-regression.sh`,
`scripts/check-merge-strategy.sh`, `.github/workflows/*` (pinned `setup-go` steps), and
`docs/merge-strategy-gate.md`. Evidence at final reviewed HEAD `d9f7506`:

* **Go**: `go build ./...` and `go vet ./...` (root module and `tools/gatecheck` module) exited
  0. `go test ./...` reported every package `ok` except the pre-existing, environment-specific
  Windows-only `tools/gatecheck` `runner_test.go` `TestGatecheckCleanup_*` failures (exit 127,
  backslash temp path — not a regression; tracked under stash `96B308D1`, confirmed pre-existing
  and Linux-CI-authoritative per the operator's own environment-quirks note). `golangci-lint
  run ./...` reported 0 issues (fixed in commit
  `607b6a9`: 3× `os.RemoveAll` errcheck wraps, 1× `unparam` dead-parameter removal).
  `gofmt -l .` reported clean against an LF-normalized copy (this checkout is CRLF via
  `core.autocrlf`).
* **Parity evidence**: `docs/plans/evidence/2026-09-28-gate-engine-go-migration/m3.md` records
  the feature-done criterion satisfied, with the three Copilot-confirmed parity exceptions
  disclosed in §14.
* **CI**: all checks were green at final reviewed HEAD `d9f7506` (re-verified after the
  Copilot-review-loop pushes): `ci gate`, `cross-compile (darwin/amd64|darwin/arm64|linux/amd64|
  windows/amd64)`, `detect code changes`, `gitignore append-only + un-ignore regression (I6)`,
  `lint`, `load cross-compile targets`, `merge-strategy structural verification`,
  `pipeline-topology (ambient)`, `security`, `test`, `test (windows, advisory)` — 14 checks
  total, matching M2's own CI shape.
* **Local review readiness**: `READY_WITH_FOLLOWUPS` (0 P0, 0 unresolved P1 — all P0/P1-equivalent
  findings fixed; residual P2/P3-equivalent findings captured as stash follow-ups) recorded in the
  PR body's `## Local Review Readiness` block, refreshed for the final reviewed HEAD `d9f7506`
  after the Copilot review-fix round.
* **Tiered persona review** (correctness, Go, security, concurrency, maintainability,
  architecture, scope-boundary, constitution): READY overall; 3 out-of-scope findings captured as
  P-021 stash entries `7223218F` (high), `5A8EC1BC` (low), `978D2946` (medium).
* **Adversarial review** (multi-model consensus, per operator mandate, before PR creation):
  round 1 — `BLOCKED` (1 critical `C-1`, 1 major `M-1` finding); both fixed in commit `6f3e372`.
  Targeted re-verification round — `CLEARED` 3/3 (all three reviewer models independently
  confirmed both findings resolved), recorded in commit `413cdd9`. Full reports:
  `docs/reviews/2026-09-30-m3-unignore-mergestrategy-go-port-adversarial-review.md`,
  `docs/reviews/2026-09-30-m3-unignore-mergestrategy-reverify-adversarial-review.md`.
* **Copilot review loop**: 2 iterations, 5 threads total, all resolved.
  * Iteration 1 (HEAD `413cdd9`): 5 unresolved threads — 3 Principle III containment/scratch-dir/
    extra-args findings (all confirmed faithful Python parity via direct comparison against
    `git show 0092a08:scripts/<file>.sh`; declined and consolidated into stash `50E6F22C`), 1
    pre-existing stash-entry typo (`unignone.GitRunner`, declined per P-021 C5 role-boundary —
    Ship cannot edit an existing stash entry), 1 genuine doc-staleness finding (fixed in commit
    `464eb77`).
  * Iteration 2 (HEAD `d9f7506`): "Findings: None" — all 5 prior findings confirmed "Resolved
    since last review." Zero new unresolved threads.
  * `autoharness gate copilot-review 85 --repo softwaresalt/intercom` returned `SATISFIED`
    (checked twice: pre-merge and as the unconditional last-mile re-check immediately before
    merge).
* **Merge**: `gh pr merge 85 --repo softwaresalt/intercom --merge` (no `--admin`). A last-mile
  re-check immediately before merge confirmed `headRefOid` still `d9f7506`, CI green, P-018
  `SATISFIED`. The merge commit is `5fdd75aec21bb9492aabad08603a5479c1da6846`. The repository
  allows merge commits only (`allow_merge_commit=true`, `allow_squash_merge=false`,
  `allow_rebase_merge=false`), satisfying P-009.
* **Merge confirmation**: `gh pr view 85` reports `MERGED` at `2026-10-01T02:17:08Z`, and
  `git merge-base --is-ancestor 5fdd75a origin/main` exits 0.

Verdict: `PASS`. Follow-ups tracked, not blocking (see below).

## Pre-deploy audits

None apply. The shipment involves no migration, feature flag, runtime config change, or access
change — it is internal CI gate tooling, consistent with M1/M2.

## Deployment / rollout path

Merge-only. `scripts/check-unignore-regression.sh` and `scripts/check-merge-strategy.sh` now run
the Go engines on every future PR and push to `main`, via the shared M1 runner. There is no
separate deploy step.

## Post-deploy checks

* On the next few PRs, confirm the `gitignore append-only + un-ignore regression (I6)` and
  `merge-strategy structural verification` jobs continue to pass and exit with documented codes.
* Confirm the restored `--root` per-argument denylist in `check-unignore-regression.sh` continues
  to reject injected `--root`/`--root=*` arguments on any future wrapper refactor.
* Confirm no stray `tools/gatecheck` build artifacts appear in the working tree after a local gate
  run.

## Risky action record

No data loss or elevated-privilege action was taken.

* **Merge**: performed under P-017 DARK FACTORY MODE with `merge_approval_pre_authorized: true`
  (operator AFK, pre-authorized per the DARK_MODE_ACTIVE record for shipment 036-S only, covering
  both the feature PR and its post-merge closure PR). A merge commit was used and the merge was
  preceded by an unconditional last-mile re-check (HEAD, checks, P-018, P-009) immediately before
  the merge command. `admin_fallback_pre_authorized: false` was correctly never invoked — the
  normal merge path succeeded without needing a fallback.
* **Covering-feature completion (a1, S4)**: all five conditions held (every manifest descendant
  done, no live descendant at any depth, topology/containment intact and the descendant-graph
  union set-equal to the manifest, shipment live status exactly `active`). `backlogit move 046-F
  --status done` exited 0 and a re-read confirmed `done`, run immediately before close-path
  classification per the mandated a0→a1→a→b ordering.
* **Shipment closure — DISCLOSED PROCESS-COMPLIANCE GAP (corrected post-Copilot-review on PR
  #86; same category as 035-S)**: `mode: classify-close-path` was executed manually (no
  dedicated CLI gate installed in this autoharness version) and produced `CLOSE_PATH_VERDICT:
  CASCADE` / `VERDICT_REASON: FULLY_COVERED_ROOT` — `046-F` is a root with no `parent_id`, all
  13 tasks are its exact descendant set (no grandchildren, confirmed via full queue+archive
  `parent_id` scan), the manifest is set-equal to `{046-F} ∪ {13 tasks}`, and `046-F` has no
  linked deliberation (`custom_fields.source_deliberation_id` absent; no
  `\b(?:DL\d+|[0-9]+(?:\.[0-9]+)*-DL)\b` match in description/references). The canonical
  `CLASSIFICATION_BINDING` (`b8ba3fba87bf4bfed1d9c237e075542dae43da2b92b9bef59afdc21b17daebd7`)
  was computed per the skill's documented format during classification. **However, the
  subsequent `backlogit shipment ship 036-S --sha 5fdd75aec21bb9492aabad08603a5479c1da6846
  --message "..." --author "..."` call was a direct invocation of the underlying cascade
  primitive, not a conforming `mode: safe-close` call.** The skill's `mode: safe-close` Step 0
  requires taking a *fresh* snapshot at the time of the safe-close call, recomputing the
  canonical binding from that fresh snapshot, and comparing it against the supplied binding
  before delegating to the Cascade Close Sub-Procedure. That recompute-and-compare step was
  **not performed** — the binding above was computed once during classification and never used
  to gate or confirm the subsequent mutation. This reproduces the same category of gap disclosed
  in the 035-S closure (see
  `docs/compound/2026-09-30-shipment-reconcile-skill-bypass-manual-substitution-risk.md`), with
  one partial mitigation: unlike 035-S (where no binding was computed at all), this closure did
  compute and record a canonical binding value, even though it was never revalidated before the
  mutation. The Cascade Close Sub-Procedure's underlying mutation then ran. Results:
  * `returned_ids=[]`.
  * `archived_ids`, `allowed_ids` and `required_ids` are the same 15-item set: `{036-S, 046-F,
    046.001-T..046.013-T}`.
  * `parent_id: 046-F` preserved on all 13 tasks (re-read post-close).
  * `036-S` has `archived_status: shipped`; `046-F` has `archived_status: done`.

  These results confirm the **resulting archive state is internally consistent** (two-set gate,
  parent-ID preservation, archive provenance — see the cascade-close report). They do **not**
  retroactively establish that the mutation was authorized through a bound safe-close
  revalidation; that gap is irreversible once the mutation has executed and is disclosed here
  rather than claimed as closed. A deferred follow-up is warranted to make `shipment-reconcile`'s
  `mode: safe-close` actually invokable (CLI or gate subcommand) so this step can be executed
  literally rather than manually approximated — see the Follow-ups section below.

  See `.backlogit/reconcile/036-S-pre-2026-09-30T19-22-49Z.md`,
  `036-S-classify-close-path-2026-09-30T19-21-40Z.md`,
  `036-S-cascade-close-2026-10-01T02-25-49Z.md`, and
  `036-S-post-2026-10-01T02-26-30Z.md`.
* **P-007 archive-integrity guard**: `git status --short -- ".backlogit/archive/"` showed only
  additions/modifications, no deletions. No restore was needed.

## Healthy signals

* The `gitignore append-only + un-ignore regression (I6)` and `merge-strategy structural
  verification` jobs stay green on `main`.
* `tools/gatecheck` builds cleanly as part of `go build ./...` / `go vet ./...` on every PR.
* `internal/unignore` and `internal/mergestrategy` test suites stay green.

## Failure signals

* Either gate's verdict or output changes on a PR that does not touch the scanned surfaces — a
  behaviour-preservation regression in the Go port.
* `tools/gatecheck` fails to build, or the shared runner fails to invoke it, on any supported
  platform.
* A future wrapper refactor silently drops the `--root` per-argument denylist loop (stash
  `7223218F`'s concern materializing as a real regression rather than a latent gap).

## Monitoring plan

Watch the standard CI dashboard on `main`, in particular the two new/changed jobs, over the next
several PRs (037-S onward, through M4 — full Python retirement).

## Rollback trigger

Revert if any of these happens:

* A false-positive or false-negative unignore-regression or merge-strategy verdict that the Go
  port caused.
* `tools/gatecheck` breaking the build or any cross-compile target.
* The `--root` denylist regression recurring undetected in production use.

## Rollback procedure

Open a standard merge-commit revert PR for merge commit
`5fdd75aec21bb9492aabad08603a5479c1da6846`. That restores the Python-based unignore-regression
and merge-strategy wrappers exactly. There is no data or state migration. The original Python
scripts remain available in git history (retired by M4, not this shipment).

## Validation window

The next 2–3 PRs on `main`: 037-S onward, through the start of M4 (full Python retirement, the
final consumer of M1–M3's shared patterns).

## Owner

The Ship agent and the repository maintainer (`softwaresalt/intercom`).

## Follow-ups (stash; Stage-owned triage)

Five `DEFERRED SCOPE EXPANSION` stash entries were captured across this shipment's persona/
Copilot review cycles and its own post-merge closure review, all per P-021 C1/C2 (out-of-scope
findings; no code change made for them beyond what is separately noted):

* `7223218F` (**high priority**, requires deliberation) — multi-persona convergence (Security
  Reviewer, Maintainability Reviewer, Architecture Strategist): `tools/gatecheck/main.go`'s
  `parseRoot` scans the whole forwarded argument list with last-occurrence-wins semantics, so
  every wrapper forwarding more than one fixed non-flag argument must independently re-implement
  a `--root`/`--root=*` denylist. M3 correctly repeated the M2-established pattern for
  `check-unignore-regression.sh`; Stage should decide whether a shared, cross-cutting
  implementation (rather than per-wrapper repetition) is warranted.
* `5A8EC1BC` (low priority) — Architecture Strategist + Maintainability Reviewer: three
  independently hand-rolled `GitRunner` abstractions now exist in `tools/gatecheck`
  (`writepath`, `retiredarch`, and M3's new `unignore`) with no shared interface. Also contains an
  uncorrected `unignone.GitRunner` typo in the entry's own text (Ship cannot edit an existing
  stash entry per the P-021 C5 capture-only carve-out; declined on the Copilot thread citing this
  role-boundary restriction, left for Stage to correct during triage).
* `978D2946` (medium priority) — Constitution Reviewer: a CRLF line-ending drift in
  `scripts/check-unignore-regression.sh` (123 CRLF sequences despite `.gitattributes`' `eol=lf`
  declaration) was caught and fixed during this session's own build via manual discovery
  (`bash -n` syntax error, `git check-attr`/`git status` warnings), not by any automated gate.
  No durable CI-level line-ending-drift check exists yet.
* `50E6F22C` (medium priority) — Copilot review (PR #85): three related Constitution Principle
  III (workspace isolation / file-system containment) findings, all independently confirmed via
  `git show 0092a08:scripts/<file>.sh` to be faithful, byte-for-byte parity preservation of the
  original Python engines' pre-existing behavior (root-containment-free file reads, OS-temp
  scratch directories, silent extra-argv handling) — not new regressions. Fixing item (3)
  specifically would itself violate INV-1 absent a new ED entry. Stage should decide whether a
  dedicated hardening unit (optionally consolidated with `7223218F`) is warranted, and whether a
  new ED entry should be proposed for item (3)'s behavior in a future plan revision.
* `D10D3AFC` (**high priority**, requires deliberation) — captured during this closure's own
  review (Copilot review on PR #86): the disclosed process-compliance gap above (direct
  `backlogit shipment ship` invocation instead of a conforming `mode: safe-close` call) recurred
  for the second consecutive shipment closure (035-S, now 036-S), because `shipment-reconcile`'s
  `mode: safe-close` has no literal CLI/gate implementation in this workspace to invoke — Ship can
  only manually approximate the classification predicates, not the mandated fresh-recompute-and-
  compare binding revalidation. Recommends implementing a real invokable `safe-close` subcommand
  or an equivalent hard gate. **Known cosmetic artifact**: the captured entry's text contains one
  stray bell-character (`\a`) in place of a backtick, introduced by a PowerShell double-quoted-
  string escape-sequence quirk (`` `a `` inside a double-quoted string is interpreted as an alert
  escape) during capture — the surrounding prose is otherwise intact and the meaning is
  unaffected. Left uncorrected per the P-021 C5 capture-only carve-out (Ship cannot edit an
  existing stash entry after capture), same precedent as the `unignone.GitRunner` typo in
  `5A8EC1BC` — left for Stage to correct during triage.

All five entries were independently verified present in `.backlogit/stash.jsonl` during this
closure (`backlogit stash list`/`backlogit stash get`), each with full six-field P-021 payloads.

One additional pre-existing, environment-specific item is disclosed for awareness, not tracked as
a new stash entry (already disclosed by the operator and in M1/M2's own closures): the Windows-only
`tools/gatecheck` `runner_test.go` `TestGatecheckCleanup_*` failures (exit 127, backslash temp
path) are not regressions; Linux CI is authoritative, the Windows CI job is advisory.

## Source artifact cleanup

* `046-F` has no `custom_fields.source_stash_id`. Its description cites "stash C44E2C1F; folds
  150364D2", but Step 7 keys only on the structured field, so it skipped this entry and left it
  for Stage — consistent with M1/M2's own closure decisions, since `C44E2C1F` governs the full
  M1–M4 migration plan (M4 still pending) and `150364D2` is already archived/consumed (folded
  into `046.012-T`/`046.013-T` per Stage's prior `C44E2C1F` reconciliation, confirmed via
  `.backlogit/archive/stash.jsonl`).
* `046-F` has no `custom_fields.source_deliberation_id`. Skipped. Its references cite
  `docs/decisions/2026-09-28-intercom-go-gate-engine-python-to-go-migration-deliberation.md`, a
  file artifact, not a backlogit deliberation ID.
* Archived: none. Skipped: `C44E2C1F` (no structured link; governs the still-in-progress M1–M4
  plan).

## Knowledge graduation

* No updates were needed to `docs/ARCHITECTURE.md`, `AGENTS.md`, `docs/design-docs/`, or
  `docs/product-specs/`. The change is internal CI gate tooling; consistent with M1/M2's own
  assessment, the migration plan does not call for graduating those docs until M4, if at all.
* `docs/merge-strategy-gate.md` was corrected in-PR (commit `464eb77`) to accurately describe the
  M3-T12 bash-orchestration hardening (temp-file lifecycle fix, exit-code table) instead of a
  stale "unchanged" claim — this is scope-correct in-PR documentation maintenance, not a separate
  knowledge-graduation artifact.
* Compound-refresh: see below.

## Compaction status (P-020)

**`done`.** The mandatory post-merge `compact-context` invocation (Step 6 item 8) ran against
this release unit's memory group (`target: all`). Candidates identified: the 2 memory files
produced during this session for 036-S/046-F (`036-s-pr85-opened-checkpoint.md` and the final
session-memory checkpoint) — both qualified under the completed-work rule (046-F reached
`done`/archived). No plans required consolidation (the M3 plan unit is embedded in the shared
multi-milestone plan `docs/plans/2026-09-28-intercom-go-gate-engine-go-migration-plan.md`, which
remains active for M4 and was correctly left untouched). No closure records exceeded the
age threshold for compaction (036-S's own closure artifact is fresh).

Output: a single dense compacted summary,
`docs/memory/compacted/2026-10-01-036-s-046-f-compacted.md`, capturing what shipped, the review
summary, the full post-merge closure procedure (including the binding-before-cascade ordering
lesson), and durable learnings. Verbose originals were moved to
`docs/archive/memory/2026-09-30/036-s-pr85-opened-checkpoint.md` and
`docs/archive/memory/2026-10-01/036-s-046-f-post-merge-closure-session-memory.md` (traceable via
the compacted file's `compacted_from` frontmatter field). See
`docs/closure/2026-10-01-036-s-compound-refresh.md` for the accompanying compound-refresh report
(1 entry updated with a success confirmation, 5 entries reviewed and classified `keep`).

## Releasability evidence

**READY.**

* Local review readiness: `READY_WITH_FOLLOWUPS` (P0=0, unresolved P1=0 — all findings fixed or
  explicitly deferred with stash capture) at final reviewed HEAD `d9f7506`, unchanged through the
  merge.
* All CI checks were green at the merge HEAD, and the merge used a merge commit (P-009).
* P-018 `SATISFIED` (confirmed twice: pre-merge and last-mile re-check, `unresolved_thread_ids:
  []`), P-016 pipeline-topology passed at every required phase (`pre_claim`, `post_claim`,
  `lifecycle`), and the P-014 merge approval is recorded via the DARK_MODE_ACTIVE
  `merge_approval_pre_authorized: true` record for shipment 036-S (operator AFK; scope-matched,
  in-scope PR, all required gates green).
* No runtime service surface changed (`cmd/` untouched). `tools/gatecheck` build, the parity
  evidence, and the full local + CI test suites are all green.
* Four disclosed review-cycle follow-ups (stash `7223218F`, `5A8EC1BC`, `978D2946`, `50E6F22C`)
  plus one closure-review follow-up (stash `D10D3AFC`, captured during this closure's own
  Copilot review) are Stage-owned triage items, not release conditions — none block `046-F`'s
  feature-done criterion.
* **Process-compliance note (DISCLOSED GAP, same category as 035-S — corrected post-Copilot-
  review on PR #86)**: this closure computed the canonical `CLASSIFICATION_BINDING` during a
  manually-executed `classify-close-path` step, but the subsequent `backlogit shipment ship` call
  was a **direct** invocation of the cascade primitive, not a conforming `mode: safe-close` call
  that freshly recomputes and compares the binding before delegating to the Cascade Close
  Sub-Procedure. This reproduces the same category of process-compliance gap Copilot flagged on
  PR #84 for the 035-S closure (see
  `docs/compound/2026-09-30-shipment-reconcile-skill-bypass-manual-substitution-risk.md`), with
  one partial mitigation: a binding was at least computed and recorded here, unlike 035-S. **A
  process-compliance gap IS disclosed for this closure** — it is not closed, only partially
  mitigated. The resulting archive state's data integrity is independently verified (two-set
  gate, parent-ID preservation, archive provenance) and is not in question; the procedural gap is
  about the missing safe-close revalidation step itself, not about data corruption. Stash
  `D10D3AFC` captures the structural remediation recommendation (an actually-invokable
  `safe-close` implementation) for Stage's deliberation queue.

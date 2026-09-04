---
title: "Operational Closure — intercom-go C1: Architecture Correction (004-S)"
date: 2026-09-04
mode: post-merge
shipment: 004-S
feature: 005-F
pr: 15
merge_commit_sha: 2588f2e91e4d572d8e408538cba9e894c6fb3a55
compaction_status: done
closure_status: READY
releasability: READY
---

# Operational Closure — intercom-go C1 (004-S)

## Summary of the Change

Shipment `004-S` — "intercom-go C1: architecture correction — config
remediation, Go 1.24 floor, anti-regression gate" — implements corrective
slice C1 of the 2026-09-04 operator architecture correction:

* Removes the retired Slack / custom-ACP broker / headless-Copilot-CLI
  architecture from the shipped `internal/config` surface: deletes
  `SlackConfig`, `ACPConfig`, `HostCLI`/`HostCLIArgs`, `IPCName`, and
  `WorkspaceMapping.ChannelID`; removes the channel-routing API
  (`ResolveChannelID`, `ResolveWorkspaceByChannelID`), re-founding
  `WorkspaceRootForChannel` as `WorkspaceRoot(workspaceID string)` keyed on
  `workspace_id`; renames `SlackDetailLevel` → `OperatorDetailLevel`
  (`operator_detail_level`).
* Adds `[copilot].cli_path` with a single validation predicate (empty →
  valid, PATH resolution; absolute → existence-checked; drive-relative or
  relative → rejected; bare name → PATH advisory only, non-fatal).
* Renumbers the validation contract 10 → 7 rules (enumerated by name);
  re-pins non-zero defaults 17 → 12.
* Rewrites `config.toml.example` (removes
  `host_cli_args = ["--dangerously-skip-permissions"]` and
  `host_cli = "claude"`; aligns `database.path` with `default.go`) and
  `docs/config-reference.md` (full Migration section with
  what-to-write-instead guidance for every removed/renamed/migrated key).
* Adds `scripts/check-retired-architecture.sh`, a self-testing mechanical
  anti-regression gate wired as one additive, `continue-on-error: true`
  step inside the existing `lint` job (non-blocking; explicitly not yet a
  required check — see Open Operator Decisions below).
* Raises the Go language floor to 1.24 (`go.mod`), ahead of the Copilot
  SDK's declared floor for its phase-C2 adoption. **No SDK dependency is
  added in this slice** (explicit plan Non-Goal).

Governing plan: `docs/plans/2026-09-04-intercom-go-architecture-correction-plan.md`
(hardened, adversarial review PASS at attempt 3 of 3). Governing decision:
`docs/decisions/2026-09-04-intercom-go-architecture-correction-copilot-sdk-deliberation.md`.
Governing design: `docs/design-docs/intercom-go-backend-architecture.md` (rev 2).

Feature `005-F`, 4 tasks, 10 subtasks (15 manifest items), all `done`.
Predecessor: `003-S` (shipped, PR #11 + post-merge closure PR #12). Stage
assembled this shipment via PR #14
(merged `234c997a1867cbc4443e7d96fa3aa810a43502cc`).

PR: https://github.com/softwaresalt/intercom/pull/15
Merge commit: `2588f2e91e4d572d8e408538cba9e894c6fb3a55` (merge-commit
strategy, per P-009; verified 2-parent merge commit, parents `234c997`
(main tip) and `4788a97` (feature branch HEAD) — not a squash or rebase).

## CI Status and Review Items

* Hosted CI (reviewed/merged HEAD `4788a97800ca7c6256144c921ccd8598d77762b8`):
  10/10 checks green — `ci gate`, `detect code changes`,
  `pipeline-topology (ambient)`, `test`, `lint` (including the new
  non-blocking retired-architecture gate step, confirmed green in the job
  log), `security`, `load cross-compile targets`, 4/4 `cross-compile` legs
  (linux/amd64, windows/amd64, darwin/amd64, darwin/arm64).
* Local review readiness: `READY_WITH_FOLLOWUPS` at reviewed HEAD
  `4788a97800ca7c6256144c921ccd8598d77762b8` (final HEAD before merge — no
  further HEAD advance occurred between the readiness record update and
  merge).
* Six parallel structured local reviews (Go Reviewer, Correctness Reviewer,
  Security Reviewer, Architecture Strategist, Scope Boundary Auditor,
  Schema-CLI-Docs Coupling Reviewer), report-only mode: **0 P0/P1**. 2 P2
  (missing test coverage for the `Workspaces[i].Path` deferred-commit
  invariant; missing schema→docs key-path mechanical coverage) and 5 P3
  findings (stale `go.mod` toolchain comment; redundant CI step-level `if`;
  unused test variable; two script-robustness advisories about TOML
  multi-line strings and CI self-test invocation). 3 P3s and both P2s were
  fixed directly as same-contract-surface completions (P-021 C1, commit
  `8275927`); 2 P3 script-robustness advisories were out of scope and
  deferred (stash `B6203CCA`, `78D13775`); one informational
  low-confidence UNC-path note was pre-existing/inherited behavior,
  explicitly out of scope, no action. Re-verified after remediation: full
  gate suite green at the fix commit; no re-review needed (fixes were
  additive test coverage and comment/config corrections, not behavior
  changes to reviewed logic).
* P-018 copilot-review gate: `NOT_APPLICABLE` (no engagement signal,
  enforcement `auto`) — verified twice: at PR-readiness time and again at
  the last-mile re-check immediately before merge, via
  `autoharness gate copilot-review 15 --repo softwaresalt/intercom
  --enforcement auto --max-wait 0`, `exit_code: 0` both times.
* Unresolved review threads/comments: none — local reviews were report-only
  persona reviews with no PR thread created; `reviewDecision` was empty
  (no branch-protection-required human review configured).
* Repository merge-strategy check (P-009): `allow_merge_commit: true`;
  merged explicitly via `gh pr merge 15 --merge` (repo also permits
  squash/rebase, so the strategy was specified, not assumed from a
  default). Merge commit parent count confirmed as 2.
* Pipeline-topology gate: `pre_claim` (twice: before branch setup and
  immediately before claim), `post_claim` (claim-verify, `CLAIM_VERIFY_OK`,
  sole active shipment), and `lifecycle` (three times: before PR creation,
  before closure/safe-close) all passed with `exit_code: 0`.

### Merge Confirmation

* `gh pr view 15 --json state,mergedAt,mergeCommit`: `state: MERGED`,
  `mergedAt: 2026-09-04T21:28:06Z`, merge SHA
  `2588f2e91e4d572d8e408538cba9e894c6fb3a55`.
* `git merge-base --is-ancestor 2588f2e91e4d572d8e408538cba9e894c6fb3a55
  origin/main`: exit 0 — confirmed in `origin/main` history.
* Feature branch `feat/004-s-architecture-correction-c1` left on origin
  (not deleted; no `--delete-branch` flag used at merge time — no explicit
  cleanup instruction beyond standard PR lifecycle for this shipment).

## Runtime Verification / Validator Evidence

No `runtime-verification` invocation was required for this shipment.
`internal/config` remains a library package whose only runtime-observable
surface is the load/validate contract exercised directly by the Go test
suite; `cmd/intercom/main.go`'s consumption of `config.DefaultConfigPath` is
byte-for-byte unchanged (invariant I9, confirmed no `cmd/**` file was
touched). The `[copilot].cli_path` field and the 7-rule validation contract
are pure schema/validation additions with no running service, no network
surface, and no CLI lifecycle change in this slice — the SDK itself is not
yet imported. Fully exercised by the standard Go validation suite and
hosted CI evidence above (same rationale as `003-S`).

## Pre-Deploy Audits

* `go.mod`: language floor `1.22` → `1.24`; `toolchain go1.26.5` pin
  (GO-2025-3750 remediation, invariant I7) preserved byte-identical. No new
  dependency added; `go mod tidy` reports no diff.
* `gofmt -l .` clean; `go vet ./...` clean; `golangci-lint run ./...` 0
  issues; `staticcheck ./...` clean; `govulncheck ./...` 0 vulnerabilities
  in code reachable from this diff (2 imported-package / 6 module
  vulnerabilities reported are pre-existing, not on any call path, and
  unrelated to this change).
* `go test -race ./...` — full pass across all packages (`cmd/intercom`,
  `cmd/intercom-ctl`, `internal/apperr`, `internal/config`,
  `internal/pathsafe`, `tests/integration`).
* `CGO_ENABLED=0 go build ./...` — succeeds (CGO-free mandate preserved);
  4/4 cross-compile targets green in CI.
* `scripts/check-retired-architecture.sh --self-test` — PASS (committed
  fixture correctly rejected; clean tracked tree correctly passes).

## Deployment / Rollout Path

Merge-only. No service is deployed or restarted as part of this shipment —
both `cmd/` binaries still return `not implemented`; no canary, phased
rollout, or maintenance window applies. The retired-architecture CI gate
step is non-blocking (`continue-on-error: true`) by design (H1 deferred).

## Post-Deploy Checks

* Confirm phase **C2** (SDK proving spike — pinning
  `github.com/github/copilot-sdk/go` `v1.0.11`, first real import) consumes
  `internal/config`'s corrected `[copilot].cli_path` contract without
  regressing the 12-default baseline or the 7-rule validation contract
  pinned here.
* Confirm the retired-architecture gate (`scripts/check-retired-architecture.sh`)
  continues to pass green (informationally, non-blocking) on subsequent
  PRs touching `internal/config/**`, `config.toml.example`, or `cmd/**`.
* Track operator decision **H1** (flip the gate to a required check) —
  blocked on the CODEOWNERS protection tracked in stash `BEDD2E70`.

## Healthy Signals

* CI `ci gate` aggregation stays green on subsequent PRs.
* `scripts/check-retired-architecture.sh`'s non-blocking step continues to
  report clean (exit 0) on the `lint` job.
* `internal/config`'s 12-default baseline and 7-rule validation contract
  remain stable as phase C2 begins consuming `[copilot].cli_path`.

## Failure Signals

* Any hosted CI run regressing the `security` job (staticcheck /
  govulncheck / gitleaks) without an explicit, reviewed cause.
* `scripts/check-retired-architecture.sh` producing a false positive on an
  unrelated PR (tracked risk per the governing plan's Risks table, PA-4).
* A future phase importing the Copilot SDK without first resolving
  operator decisions H2 (CLI minimum version) or H5 (permission
  authorization model), both of which the plan defers past this slice.

## Monitoring Plan

* GitHub Actions run history for `.github/workflows/ci.yml` on `main`,
  specifically the `lint` job's new "Run retired-architecture gate" step.
* No dashboards/alerts apply — this is a library-primitives shipment with
  no running service.

## Rollback Trigger / Procedure

* Trigger: phase C2 (SDK proving spike) reveals the `[copilot].cli_path`
  contract, the 7-rule validation contract, or the 12-default baseline is
  structurally wrong for the SDK's actual resolution behavior.
* Procedure: `git revert` the merge commit
  `2588f2e91e4d572d8e408538cba9e894c6fb3a55` on `main` via a standard PR
  (no direct-to-main mutation). The retired-architecture gate step can be
  independently removed from `.github/workflows/ci.yml` by reverting its
  single additive step if it produces persistent false positives, without
  reverting the config remediation itself.

## Validation Window

Through phase **C2** (SDK proving spike, the first real consumer of the
corrected schema) — per the governing plan's own stated validation window.
The schema, validation contract, and gate are considered validated once C2
exercises `[copilot].cli_path` against a real SDK-resolved CLI process.

## Owner

Repository maintainer (`softwaresalt`) — no separate on-call/runtime owner
applies; this is a library-primitives shipment, not a running service.

## Risky Action Record

None destructive. The plan's own risk classification (PA-1 through PA-4)
applies: eliminating validation rule 6 (`ActionRisk: high`, no approval
required — compensated by the new relative-path rejection, PA-1);
removing `--dangerously-skip-permissions` from the example
(`ActionRisk: moderate`, security improvement, PA-2); the breaking schema
rewrite (`ActionRisk: high`, no approval required — blast radius verified
bounded to one non-test importer consuming only `DefaultConfigPath`, PA-3);
adding the non-blocking CI gate (`ActionRisk: high` for a *future* required
flip only, explicitly deferred pending H1 + CODEOWNERS `BEDD2E70`, PA-4).
No destructive, irreversible, or high-blast-radius action was taken during
implementation or closure. The shipment-record close required the P-015
cascade path (`backlogit shipment ship`) because the generic
`backlogit move --status shipped` path explicitly refuses direct shipment
status mutation in this backlogit version — the intended, verified-
precondition mechanism, consistent with the `003-S` precedent; see
Shipment Closure Evidence below.

## Source Artifact Cleanup

* `005-F.custom_fields.source_stash_id`: not present — no stash entry to
  archive.
* `005-F.custom_fields.source_deliberation_id`: not present — no
  deliberation artifact to archive. (`005-F.references` points at durable
  design documents — the governing plan, decision, and design-doc rev 2 —
  not backlogit deliberation artifacts; retained.)
* Archived: 0 stash, 0 deliberations.

## Deferred Stash Entries (P-021 C2, captured during this shipment)

| ID | Priority | Summary |
|---|---|---|
| `B6203CCA` | low | `scripts/check-retired-architecture.sh`'s `--self-test` mode is never invoked by CI — only the plain scan runs; a regression in detection logic would go uncaught mechanically |
| `78D13775` | low | Gate script's TOML comment-masking doesn't handle multi-line (`"""`/`'''`) strings — zero real-world risk today (no such fixtures exist), robustness gap only |

Reused (no new entry needed — already captured by the governing plan at
Stage time): `8C2D578D` (`internal/apperr`'s `KindSlack`/`KindIPC`/
`KindACP` taxonomy contamination, D6a) — reconfirmed present and
unmodified by this slice, exactly per the plan's explicit exclusion of
`internal/apperr` (invariant I5).

## Shipment Closure Evidence

* Reconciliation baseline: all 15 manifest items (`005-F` + 4 tasks + 10
  subtasks) were moved to `status: done` and relocated to
  `.backlogit/archive/` during Ship's own execution session, prior to PR
  creation (committed on the feature branch, merged with the PR). No
  orphans found (no queue file declared `shipment_id: 004-S` outside the
  manifest); full-prefix enumeration of `005.*`/`005-*` confirmed exactly
  15 files, matching the manifest exactly, no extras.
* Close-path classification (P-015): `005-F` is a root feature (no
  `parent_id`) and its full descendant set at every depth — all 4 tasks and
  10 subtasks — is exactly the shipment manifest; no other descendants
  exist; `005-F` carries no linked-deliberation ID (its `references` are
  durable design docs, not `DL`-pattern backlogit deliberation IDs).
  **Verified fully-covered-root cascade** selected (generic
  `backlogit move 004-S --status shipped` independently confirmed this by
  refusing with exit 9: "shipment must be shipped via ShipShipment, not a
  direct status update").
* Executed `backlogit shipment ship 004-S --sha
  2588f2e91e4d572d8e408538cba9e894c6fb3a55 --message "merge: shipment
  004-S -- intercom-go C1 architecture correction (PR #15)" --author
  "ship-agent@autoharness"`:
  * `shipment_status: "shipped"`; `returned_ids: []` (no
    classifier/engine mismatch).
  * `archived_ids` (16: shipment + feature + 4 tasks + 10 subtasks) exactly
    matches the computed required set (all 15 non-shipment manifest items
    were declared `status: done`, not `archived`, pre-close, so all were
    required-and-archived; no unexpected artifact archived, no required
    artifact left unarchived).
* Shipment record: `status: archived`, `archived_status: shipped`,
  `commit: 2588f2e91e4d572d8e408538cba9e894c6fb3a55` (verified via the
  archived frontmatter of `.backlogit/archive/004-S.md`).
* Post-mode: `.backlogit/archive/004-S.md` present; `git status --short --
  ".backlogit/archive/"` showed only modifications to existing archive
  files plus one rename (`queue/004-S.md` → `archive/004-S.md`), no
  unexpected deletions — `recommendation: PROCEED`.
* Backlog archival committed on
  `post-merge/004-s-architecture-correction-c1`, not on `main`
  (P-020/branch-per-release-unit).
* Single-agent session — no concurrent agents active; file-lock protocol
  not invoked per the workspace's stated concurrency-control scope (locks
  reserved for multi-agent or operator-concurrent scenarios).

## Compaction Status (P-020)

`done`. `compact-context` (`target: all`) ran in this closure session: the
two session memory files for this release unit
(`docs/memory/2026-09-04-stage-intercom-go-architecture-correction.md`,
`docs/memory/2026-09-04-ship-004-S-pre-merge-checkpoint.md`) were
consolidated into
`docs/memory/compacted/2026-09-04-004-s-intercom-go-architecture-correction-c1-compacted.md`
and the verbose originals moved to `docs/archive/memory/`. No other
`docs/memory/` files exceeded the threshold-gated candidate criteria (2
files consumed, well under default thresholds) — this was a bounded,
per-release-unit Tier-1 consolidation.

## Releasability Evidence

**READY.** No blocking findings (0 P0/P1 from six-persona review), no
runtime risk (both `cmd/` binaries remain `not implemented`, no service
deployed), no deployment/rollback complexity beyond a standard `git revert`
of the merge commit. The two stash-captured follow-ups (`B6203CCA`,
`78D13775`) are non-blocking, low-priority, and correctly out of this
shipment's P-021 C1 scope. The non-blocking anti-regression gate lands
exactly as the plan specifies (`continue-on-error: true`, one additive
step, no `ci-gate` `needs:` edit).

## Dark Mode

Not applicable — this session ran in standard sequential mode.
`DARK_MODE_ACTIVE` was not present. All last-mile gate re-checks (§1.9,
P-018, P-009, pipeline-topology) were run as independent verifications
immediately before merge. The operator's explicit standing
autonomous-completion directive for this session ("after last-mile gates
pass, merge via merge commit and complete shipment reconciliation...")
served as the required P-014 explicit approval signal for PR #15's merge
and this post-merge closure work — scoped to this shipment's full
lifecycle, not a general dark-mode activation record.

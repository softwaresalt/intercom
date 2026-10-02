---
title: "Post-merge closure: 038-S / 048-F — Harden gatecheck CLI argument containment"
description: "Post-merge closure artifact for shipment 038-S / feature 048-F"
status: "complete"
tags:
  - "operational-closure"
  - "038-S"
  - "048-F"
  - "post-merge"
date: 2026-10-02
mode: post-merge
shipment: 038-S
feature: 048-F
pr: 91
merge_commit_sha: 0fc58cede5e6e5f5f189df281cc8533b2c67c3f7
compaction_status: done
closure_status: READY
releasability: READY
conditions: []
---

# Post-merge closure: 038-S / 048-F — Harden gatecheck CLI argument containment

## Summary

Shipment `038-S` delivered feature `048-F`, "Harden gatecheck CLI argument
containment" (post-M4 Unit C). PR #91 merged to `main` at
`0fc58cede5e6e5f5f189df281cc8533b2c67c3f7` on 2026-10-02. The change makes
the trusted `--root` authoritative for the gatecheck engines, adds a
pre-build `--root` denylist to the write-path wrapper, and records the
remaining CLI containment exceptions as explicit, scoped residuals.

The merge was documentation/developer-tooling-only with respect to deployment:
it changes gatecheck sources, wrappers, tests, and their documentation, not
Intercom's shipped CLI, WebSocket/API surfaces, or deployment configuration.
No production deployment or data migration is required.

## Invariants to preserve

* The trusted `--root` supplied by the wrapper remains authoritative; later
  `--root`/`--root=` arguments are rejected rather than overriding it.
* The write-path wrapper rejects root-injection arguments before invoking
  `gatecheck_build`; legitimate modes continue to work.
* The documented merge-strategy/unignore exceptions remain accepted,
  scoped Principle III exceptions, not silently inherited behavior.
* `.gitignore`'s Python-cache rules remain present and append-only; the
  historical comment does not authorize deleting or reordering the rules.

## Validator evidence

| Gate | Result |
|---|---|
| `gofmt -l .` | Clean |
| `go vet ./...` | Passed |
| `go test ./...` | Passed with Git for Windows Bash before WSL Bash on `PATH` |
| `go build ./...` | Passed |
| `golangci-lint run ./...` | Passed at the same reviewed HEAD |
| `staticcheck ./...` | Passed at the same reviewed HEAD |
| `goimports -l .` | Passed at the same reviewed HEAD |
| Four gate-wrapper self-tests | Passed at the same reviewed HEAD |
| PR #91 CI | All required checks passed; run 37028114843 |
| First post-merge `main` CI | Passed; run 37029204358 at the merge SHA |

The first local `go test ./...` attempt used the WSL `bash.exe` found first on
the default Windows `PATH`; wrapper tests then failed to resolve Windows paths
and `cygpath`. This was an environment-selection failure, not a code failure.
Rerunning with `C:\Program Files\Git\bin` and `C:\Program Files\Git\usr\bin`
first on `PATH` passed the full test suite. No source change was made to mask
the environment issue.

### Runtime validation applicability

Runtime verification was not run. The shipment changes internal gatecheck
tooling and CI wrappers; it does not change Intercom's product CLI, TUI,
WebSocket/API, or deployment surfaces. The workspace profile's product
runtime probes and manual checkpoints are therefore not applicable to this
release unit. The GitHub CI run and the gate-wrapper tests are the relevant
validator evidence.

## Review, CI, and merge evidence

* The Orchestrator's three-reviewer adversarial review covered exact HEAD
  `888bb963e40ffbbd403a75c3cdd90449b9d52f3b`: consensus `READY`, zero findings,
  P0=0, P1=0. No implementation changes followed that review.
* Copilot was requested once on PR #91. Its `COMMENTED` review completed for
  that same HEAD with no findings and no review threads. The P-018 gate returned
  `SATISFIED`; unresolved Copilot threads: none.
* The PR body carried the current-HEAD `READY` record, zero blocking findings,
  successful full-build evidence, and explicit handling of `D10E82EC`.
* All required CI checks were green. The PR-head and first post-merge `main`
  workflows both concluded successfully.
* P-016 topology passed. P-009 settings permitted merge commits and disabled
  squash/rebase. P-014 approval came from the scoped `DARK_MODE_ACTIVE`
  activation (`merge_approval_pre_authorized: true`). The normal merge-commit
  path succeeded; admin fallback was not authorized and was not used.

## Shipment reconciliation and backlog closure

The feature-completion gate (a1) passed all five conjunctive conditions:
all six manifest tasks were complete; no live descendant remained; the full
descendant set union the feature equaled the manifest; every descendant was
contained by `048-F`; and the feature was live `active`. Ship moved `048-F`
from `active` to `done` and verified the reread.

`shipment-reconcile` pre-mode (`expected_status: done`) returned `PROCEED`;
the shipment record was `active` and record-consistent, and the orphan scan
was empty. The bound close-path classification returned
`CASCADE` / `FULLY_COVERED_ROOT` with
`CLASSIFICATION_BINDING:
ec978763b1e5101fc66b2b51c67a84342868d2feb013fd41cd1462d4f873d42b`.
Safe-close revalidated the binding before mutation and returned `CLOSED`:

* `returned_ids` was empty.
* `archived_ids`, `allowed_ids`, and `required_ids` each contained exactly the
  six tasks, feature `048-F`, and shipment `038-S`.
* Both set differences were empty; every task retained `parent_id: 048-F`.
* Shipment `038-S` is archived with `archived_status: shipped` and merge SHA
  `0fc58cede5e6e5f5f189df281cc8533b2c67c3f7`. Feature `048-F` is archived
  with `archived_status: done`.
* Post-mode returned `PROCEED`; all eight archive files were present and the
  P-007 check found no archive deletions.

The repository's `scripts/acquire_lock.ps1` and `scripts/release_lock.ps1` were
absent. The documented file-lock advisory-lock contract was followed with an
atomic create-new lock at `.backlogit/queue/.038-S.md.lock`; it was released
after post-mode. No lock remained. Reconciliation reports are:

* `.backlogit/reconcile/038-S-pre-2026-10-02T15-54-25Z.md`
* `.backlogit/reconcile/038-S-safe-close-2026-10-02T15-58-29Z.md`
* `.backlogit/reconcile/038-S-post-2026-10-02T15-59-37Z.md`

Backlog closure was committed on the post-merge branch as `f73e784`
(`chore: archive 038-S backlog artifacts`), never directly to `main`.

## Source artifact cleanup

`048-F` has neither `custom_fields.source_stash_id` nor
`custom_fields.source_deliberation_id`. Per the manifest-derived cleanup rule,
no source stash or deliberation was archived; none was inferred from the
feature description.

The pre-existing Stage-owned deferred entry `D10E82EC` remains open for Stage
triage. It was not the source artifact for this feature, was not edited or
archived, and was not duplicated. It records an out-of-scope rationale-comment
expansion and retains its original pre-PR source references. The entry was
captured during the 048-F build but was not committed in PR #91. It is first
committed in backlog commit `f73e784` on this closure branch, and it reaches
`main` when this closure PR merges. Its content is unchanged.

## Pre-deploy audits and rollout path

No data migration, feature flag, access change, or production deployment is
required. The rollout path is merge-only: the code takes effect in CI and
developer gate runs on `main`.

## Post-merge checks and monitoring

* First `main` CI for merge SHA
  `0fc58cede5e6e5f5f189df281cc8533b2c67c3f7` passed (run 37029204358).
* Continue monitoring CI's test, lint, security, topology, merge-strategy,
  and gitignore/unignore jobs on the next two to three PRs that touch gate
  tooling or `.github/workflows/ci.yml`.

### Healthy signals

* Gatecheck builds and tests pass, including the wrapper guard tests.
* Main-branch CI and the gatecheck-related structural jobs remain green.
* Root-override attempts fail closed before wrapper build execution.

### Failure signals

* A caller-supplied root is accepted after the trusted root.
* A wrapper reaches `gatecheck_build` before rejecting a root override.
* Any CI gate's verdict changes without a corresponding authorized contract
  change, or a required CI job fails.

## Risky action record

* **Shipment closure:** the bound P-015 `CASCADE` path was selected only after
  full-root/full-descendant coverage verification. Its two-set gate, empty
  `returned_ids`, preserved parent links, archived provenance, and post-mode
  result are recorded above.
* **Merge:** normal `gh pr merge 91 --merge --delete-branch` succeeded under
  scoped P-017 authorization. No `--admin` attempt occurred.
* **Locking:** the repository lock scripts were missing; the documented atomic
  lock-file fallback was acquired and released around reconciliation.

## Rollback trigger and procedure

Trigger rollback if a confirmed regression allows a caller-supplied
`--root` to override the trusted root, or if required CI containment guards
are bypassed. Open a merge-commit revert PR for PR #91 and rerun the affected
gate tests; do not revert directly on `main`.

## Validation window and owner

Observe the next two to three PRs that exercise gatecheck or edit CI workflow
gates. Owner: Ship and the repository maintainer (`softwaresalt/intercom`).

## Follow-ups (Stage-owned)

* `D10E82EC` — pre-existing deferred-scope entry; Stage owns deliberation and
  any subsequent prioritization.
* No new follow-up stash entries were identified by runtime verification,
  local review, or this closure. The deferred entry above already exists.

## Knowledge graduation

No architecture or product-spec change is required. The shipment itself
records the accepted gatecheck containment exceptions in the relevant
gatecheck package documentation. Compound refresh found the two relevant
existing learnings still accurate and distinct; see
`docs/closure/2026-10-02-038-s-compound-refresh.md`.

## Compaction status (P-020)

**`done`.** Ship invoked `compact-context` with `target: all`; the invocation
succeeded without degradation. The four completed 038-S/048-F memory records
were summarized in
`docs/memory/compacted/2026-10-02-038-s-048-f-compacted.md` and archived,
not deleted, under `docs/archive/memory/2026-10-02/`. The compacted summary
preserves the checkpoint recovery result, reviewed HEAD and merge SHA,
reconciliation evidence, `D10E82EC`, and the remaining closure-PR work. No
active checkpoint was compacted; all four October 2 Ship checkpoints were
resolved. The completed-plan candidate check skipped the multi-shipment
re-plan document because the entire plan is not complete.

## Closure PR gate status

The closure PR is **#92**, from branch
`post-merge/048-f-harden-gatecheck-cli-argument-containment`.

**Earlier halt.** Before the closure PR was opened, Ship ran
`autoharness gate pipeline-topology --mode agent --shipment 038-S --phase lifecycle --json`.
It returned exit 1, `blocked: true`, with token `LIFECYCLE_NO_ACTIVE_SHIPMENT`
(`expected exactly one active shipment`; `active_shipment_ids: []`). Ship
halted rather than bypass the gate. No force, override, re-claim or direct
shipment mutation was attempted.

**Orchestrator ruling.** The Orchestrator ruled that the gate had been
misapplied and was not a real blocker:

* The Ship contract defines the `--phase lifecycle` topology gate at only
  three points. Two are in Step 5 (`1a` and `5a`), before the build and before
  the feature PR, while the shipment is still active. The third is Step 6
  `a0`, before safe-close. For 038-S, `a0` passed before safe-close ran.
* Step 6.0 item 4 (push the closure branch and open the closure PR) defines no
  lifecycle gate.
* Once safe-close archives the shipment, no shipment is active, so the
  lifecycle gate cannot pass. Earlier closure PRs were also opened after
  archival. For example, the 037-S / 047-F closure PR #88 was opened after
  037-S was archived.

Shipment `038-S` stays archived as `shipped`. It was not re-claimed, reopened
or mutated. This record reports the ruling for this closure only. It does not
change the Ship contract.

**Closure-PR topology evidence.** On the closure branch, Ship re-ran
`autoharness gate pipeline-topology --mode manual --phase ambient --json`. It
returned exit 0, `topology gate pass`, with these check results:

* `active_shipment_invariant`: passed, with `active_shipment_ids: []`.
* `worktree_topology`: passed, with token `WORKTREE_TOPOLOGY_OK` and one
  implementation worktree.
* `branch_ownership` and `shipment_readiness`: skipped, because an ambient
  check has no target shipment.

The Orchestrator had run the same command earlier and got the same result.

**Closure PR gates.** Before the PR opened, a local multi-persona review ran on
the closure branch. The P-018 copilot-review gate, required CI and the §1.9
readiness gate all apply to the closure PR, which merges with a normal merge
commit under the scoped P-017 authorization. Admin fallback is not authorized.
The closure PR body records the readiness result.

The backlog index resync succeeded after archival and knowledge maintenance
(`backlogit sync`, 398 artifacts indexed). Ship created checkpoint
`.backlogit/checkpoints/checkpoint-20261002-161423.json` at the halt. Resuming
from the ambient gate did not resolve it. The checkpoint stays active until
the closure PR merges, and Ship resolves it then.

## Releasability evidence

**READY.** The change is internal CI/developer tooling with no product runtime
or deployment surface change. Exact-HEAD local review was `READY` (P0=0,
P1=0); Copilot review completed cleanly, P-018 was `SATISFIED`, all PR and
post-merge CI checks passed, P-009/P-016 passed, and the merge used the
authorized merge-commit path. No runtime evidence or deployment condition
applies to this scope.

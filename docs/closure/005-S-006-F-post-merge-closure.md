---
title: "Post-merge closure: 005-S -- intercom-go C2 design reconciliation + Copilot SDK proving spike (PR #17)"
description: "Operational closure artifact for shipment 005-S / feature 006-F"
status: "complete"
tags:
  - "closure"
  - "005-S"
  - "006-F"
  - "post-merge"
---

# Post-merge closure: 005-S / 006-F

**Mode:** post-merge
**PR:** #17, merged `2026-09-05T03:19:31Z`, merge commit `3fdb871b5367f7cde54af9406840f15669e4f1ca`
**Shipment:** 005-S (shipped, archived)
**Feature:** 006-F (done, archived)

## Summary of the change

Sub-epic A: reconciliation governance for the operator's untracked candidate
design document (byte-for-byte preserved, provenance header added) and
amendments to design rev 2. Sub-epic B: phase-C2 GitHub Copilot SDK proving
spike -- a disposable `internal/copilotprobe` package that added the pinned
`github.com/github/copilot-sdk/go@v1.0.11` dependency and proved S1-S3 plus
SQ-a..SQ-d **live** against a real, authenticated Copilot backend. **No
production SDK wiring** -- `cmd/intercom` / `cmd/intercom-ctl` unmodified.

## CI status and review

* CI: all checks green on merge commit's source HEAD (`ci gate`, `lint`,
  `test`, `security`, cross-compile x4, `pipeline-topology (ambient)`,
  `detect code changes`).
* Standard review (6 personas) -- P1/P2 found and remediated pre-PR.
* Adversarial review (5-reviewer panel, operator-mandated) -- initial
  BLOCKED verdict on 9 findings, all remediated; re-verification review:
  READY.
* Copilot review (PR #17): 2 comments. One fixed in scope (`atomic.Int64`
  alignment). One matched an already-captured P-021 deferred entry
  (`A0A2D049`) -- replied citing it, no new entry. Both threads resolved;
  P-018 gate `SATISFIED` for final merged HEAD `7ad59a4`.
* No unresolved review items at merge time.

## Runtime verification

**Not applicable.** This shipment touches no runtime surface: no changes to
`cmd/intercom`, `cmd/intercom-ctl`, the WebSocket API, the TUI, or any
config surface consumed by production code. `internal/copilotprobe` is a
disposable, non-imported spike package. The `runtime_validation` block in
`.autoharness/workspace-profile.yaml` (CLI/API surfaces, WebSocket upgrade,
TUI smoke) does not apply to this change set.

## Invariants to preserve

* `cmd/intercom` and `cmd/intercom-ctl` continue to return their
  `not implemented` sentinel (unaffected by this change).
* The operator's IMPL-DESIGN candidate document remains preserved
  byte-for-byte in content (verified at merge: A1's two-commit hash
  comparison; content unaffected by anything merged after A1).
* Design rev 2 (`docs/design-docs/intercom-go-backend-architecture.md`)
  remains the sole governing design artifact; IMPL-DESIGN remains
  non-governing (`status: candidate-vision`, `governing: false`).

## Pre-deploy audits

Not applicable -- no deployment, no migration, no config/flag change, no
access change. This is a docs + disposable-test-package merge to `main`.

## Deployment or rollout path

Merge-only. No deploy, no release artifact, no rollout.

## Post-deploy checks

Not applicable (no runtime surface changed). The one meaningful "post-merge
check" already performed: `go build ./...`, `go vet ./...`, `go test ./...`
(including `-race` on `internal/copilotprobe`) all pass on `main` post-merge
(re-verified via `origin/main` fast-forward to `3fdb871`).

## Risky action record

| Action | Risk | Outcome |
|---|---|---|
| Adding a new direct dependency (`github.com/github/copilot-sdk/go`) to a previously 2-dependency module | Medium (transitive surface growth) | Isolated to disposable `internal/copilotprobe`; no production import path; G2 tidy-clean; G7 go-directive-unchanged |
| Live, credential-consuming calls against a real Copilot backend during proving tests | Medium (real API/credit consumption, real shell-tool execution via the live agent) | Bounded per-test deadlines throughout (`probeDeadline` = 90s, sub-deadlines on Abort/Stop/ForceStop); working directory rooted at `t.TempDir()` per test, never the repo checkout; flagged via P-021 capture (`A0A2D049`) for a future explicit opt-in gate |
| Modifying an untracked operator file (IMPL-DESIGN) | Medium | Two-commit procedure with `git hash-object` verification made preservation mechanically verifiable; confirmed additions-only |

## Healthy signals

* `go build ./...`, `go vet ./...`, `gofmt -l .`, `go test ./...` all green
  on `main` post-merge.
* `internal/copilotprobe` remains unimported by any production package.

## Failure signals

* Any future PR importing `internal/copilotprobe` from a production package
  (would violate its disposability and design section 7.1's ACL-confinement
  intent -- flagged as follow-up `309FBF5A`).
* `go mod tidy` producing a diff against `go.mod`/`go.sum` (would indicate
  drift from the pinned SDK version).

## Monitoring plan

Not applicable -- no runtime surface, no deployed service, no dashboards or
alerts affected by this change.

## Rollback trigger / rollback procedure

Not applicable for a docs + disposable-spike-package merge with no runtime
surface. If ever needed: `git revert 3fdb871` (single merge commit revert;
the pinned SDK dependency and `internal/copilotprobe` package were added in
one shipment and have no other consumers, so a revert is clean).

## Validation window

N/A (no deployed/runtime surface to watch).

## Owner

Ship agent (this session) / repository maintainer for any follow-up
disposition of the P-021 deferred entries below.

## Compaction status (P-020)

`done`. `compact-context` invoked with `target: memory`; the just-closed
release unit's session memory
(`docs/memory/2026-09-04-ship-005-s-session.md`) was the intended candidate
under the completed-work rule -- compacted to
`docs/memory/compacted/2026-09-04-005-s-intercom-go-c2-reconciliation-spike-compacted.md`,
verbose original archived to
`docs/archive/memory/2026-09-04-ship-005-s-session.md`. Plan consolidation
for `docs/plans/2026-09-04-intercom-go-c2-sdk-spike-plan.md` was evaluated
and deferred (the plan's own appended review-remediation history remains
useful context for anyone auditing the B1-B7 acceptance-criteria trail
against the live findings, and is not yet a compaction candidate under the
skill's own criteria -- no threshold exceeded).

## Releasability evidence

**READY.** No runtime surface changed; `runtime_validation.releasability`
requirements in the workspace profile (CLI/TUI smoke, WebSocket
connect/hydrate/tool-approval evidence) are not applicable to this shipment.
All CI, review, and quality gates passed for the merged HEAD.

## Follow-up items (P-021 deferred stash entries)

* `A0A2D049` (medium, requires deliberation) -- explicit opt-in gate for
  live-SDK-credential-consuming test execution. Also independently raised by
  Copilot review on PR #17 (thread resolved citing this entry).
* `6E701953` (low) -- `commit-message.instructions.md` still lists the
  retired "acp" scope token (D12 terminology drift, unrelated file).
* `309FBF5A` (medium) -- no C3 planning acceptance criterion for
  deleting/merging `internal/copilotprobe` into the real ACL adapter; no
  depguard/CI lint enforcing design section 7.1's adapter-confinement rule.

## Source artifact cleanup

* `006-F.custom_fields.source_stash_id`: not present on this feature record
  (006-F was harvested directly from the plan/deliberation via Stage's
  standard planning flow, not from a discrete stash entry) -- no stash
  archival action applicable.
* `006-F.custom_fields.source_deliberation_id`: not present as a single
  literal field; the feature references two deliberations
  (`docs/decisions/2026-09-04-intercom-go-architecture-correction-copilot-sdk-deliberation.md`
  and
  `docs/decisions/2026-09-04-intercom-go-implementation-design-reconciliation-deliberation.md`)
  as plain markdown `references`, not as `backlogit` artifacts with their
  own `artifact_type: deliberation` records -- no `backlogit_archive_item`
  action applicable via this mechanism.

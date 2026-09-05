---
title: "Ship session memory: 005-S intercom-go C2 design reconciliation + Copilot SDK proving spike"
description: "Full Ship pipeline execution record for shipment 005-S through PR creation"
status: "in-progress"
tags:
  - "ship"
  - "005-S"
  - "copilot-sdk"
  - "phase-c2"
---

# Ship session memory: 005-S

**Mode:** DARK_MODE_ACTIVE, careful mode, merge_approval_pre_authorized=true,
admin_fallback_pre_authorized=false. visibility_mode=local structured output
(agent-intercom unavailable in this session).

## Shipment scope

005-S: feature 006-F ("intercom-go C2: design reconciliation + Copilot SDK
proving spike") + 11 task/subtask descendants. Predecessor 004-S satisfied.
Pre-claim topology gate passed before this session began.

## Branch

`feat/005-s-intercom-go-c2-design-reconciliation-copilot-sdk-proving-spike`
(created from `main` at `7ee36c4`).

## Items completed (all 12 manifest items, all `done`)

* 006-F (feature) -- done
* 006.001-T (Sub-epic A: reconciliation governance) -- done
  * 006.001.001-ST (A1: IMPL-DESIGN provenance header, two-commit
    byte-for-byte preservation, hash-verified) -- done
  * 006.001.002-ST (A2: design rev 2 Non-Goals/terminology/deferred
    register/section 8 amendment) -- done
* 006.002-T (Sub-epic B: C2 Copilot SDK proving spike) -- done
  * 006.002.001-ST (B1: pinned SDK dependency + internal/copilotprobe
    package + findings skeleton) -- done
  * 006.002.002-ST (B2: S1 permission round-trip, live-proven, all 4
    handler-constructible variants) -- done
  * 006.002.003-ST (B3: S2 event-union, SQ-d, R5, live-proven) -- done
  * 006.002.004-ST (B4: S3 shutdown, SQ-a, live-proven) -- done
  * 006.002.005-ST (B5: SQ-b/SQ-c, live-proven) -- done
  * 006.002.006-ST (B6: findings artifact completed) -- done
  * 006.002.007-ST (B7: design amendment, C2 closure) -- done

## Key decision: real live execution, not zero-execution fallback

A live, authenticated Copilot SDK connection was available in this
environment (`client.GetAuthStatus().IsAuthenticated == true`). All S1-S3 /
SQ-a..SQ-d criteria were proven empirically against a real backend, not
via the plan's zero-execution UNPROVEN fallback path. See
`docs/decisions/2026-09-04-intercom-go-c2-sdk-spike-findings.md` for the
full evidence.

## Empirical answers (for quick reference)

* S1 PROVEN, S2 PROVEN, S3 PROVEN
* SQ-a: NO (Abort does not unblock an already-blocked handler)
* SQ-b: SERIALISED (max concurrent callback = 1)
* SQ-c: YES (dispatch independent, no head-of-line blocking)
* SQ-d: YES (SessionEvent.ID is a stable UUID v4 identity) -- refutes design's
  prior prediction
* R5: CONFIRMED (SendAndWait doc/signature drift)
* Go/no-go for C3: GO (with B7's design amendment satisfying C2 closure)

## Review cycle

1. Standard review (6 personas: Constitution, Go, Correctness, Maintainability,
   Concurrency, Architecture Strategist) -- found real P1/P2 issues, all
   remediated in commit `b5d2f24` (Go code) + `eab0d45` (design doc) +
   `4d2e84c` (findings doc footnote).
2. Adversarial review (5-reviewer consensus panel, mandatory per operator
   request regardless of standard-review finding count) -- initial verdict
   **BLOCKED** on 9 findings (F1 critical evidence-figure inconsistency,
   F2/F3 shutdown-ladder leak/escalation, F4 SQ-d fallback deletion, F6
   disputed DriveTurn error-path copy, F7/F8 unasserted test claims, F9
   blockCh leak). All remediated in commit `671b08b`.
3. Targeted re-verification review confirmed all 9 findings RESOLVED, no new
   defects introduced. Verdict: **READY**.

## P-021 deferred scope-expansion captures (out-of-scope findings, threadless/pre-PR path)

* `A0A2D049` (medium) -- explicit opt-in gate for live-SDK-credential-consuming
  test execution (Constitution Reviewer P1).
* `6E701953` (low) -- commit-message.instructions.md still lists retired
  "acp" scope token (Constitution Reviewer P3, D12 terminology drift in an
  unrelated file).
* `309FBF5A` (medium) -- no C3 acceptance criterion for deleting/merging
  internal/copilotprobe, no depguard enforcing section 7.1's adapter
  confinement rule (Architecture Strategist P2).

## Preserved operator work (never touched)

`.gitignore`, `start.ps1`, `.backlogit/stash.jsonl` (modified),
`.backlogit/hooks_queue.jsonl`, `.claude/`, `.github/copilot/` (untracked) --
all confirmed untouched throughout via `git status --short` at multiple
checkpoints.

## Next steps

1. Push branch, create PR with Local Review Readiness block.
2. Request/wait for Copilot review (P-018 hard gate); resolve all comments
   per the required protocol (classify -> fix if in scope -> verify -> commit
   -> push -> reply with commit SHA -> resolve via gh api graphql).
3. Merge with merge commit only (no squash/rebase, no --admin).
4. Post-merge closure: runtime verification, operational closure, shipment
   reconciliation (shipment-reconcile safe-close), compound refresh,
   mandatory compact-context.

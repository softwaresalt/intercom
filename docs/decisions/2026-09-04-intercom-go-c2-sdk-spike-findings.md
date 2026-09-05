---
title: "Phase-C2 Copilot SDK proving spike: findings"
description: "S1-S3 plus SQ-a..SQ-d empirical findings for the pinned github.com/github/copilot-sdk/go@v1.0.11 dependency (shipment 005-S, sub-epic B)"
status: "complete"
source_document: "docs/plans/2026-09-04-intercom-go-c2-sdk-spike-plan.md"
linked_artifacts:
  - "docs/decisions/2026-09-04-intercom-go-implementation-design-reconciliation-deliberation.md"
  - "docs/decisions/2026-09-04-intercom-go-architecture-correction-copilot-sdk-deliberation.md"
  - "docs/design-docs/intercom-go-backend-architecture.md"
  - "internal/copilotprobe"
tags:
  - "copilot-sdk"
  - "spike"
  - "phase-c2"
  - "findings"
---

# Findings: phase-C2 Copilot SDK proving spike

**Governing plan:**
`docs/plans/2026-09-04-intercom-go-c2-sdk-spike-plan.md` (sub-epic B).
**Governing deliberation:**
`docs/decisions/2026-09-04-intercom-go-implementation-design-reconciliation-deliberation.md`.
**Pinned dependency:** `github.com/github/copilot-sdk/go@v1.0.11`
(tag `go/v1.0.11`, commit `a550258d5c37bd662197536992a23d633bfe5804`).
**Probe package:** `internal/copilotprobe` (disposable, no production wiring;
tests B2-B5 in `permission_test.go`, `events_test.go`, `shutdown_test.go`,
`concurrency_test.go`; shared harness/fixture in `permission.go`,
`fixture.go`, `client.go`).

All eight criteria below were exercised **live** against a real,
authenticated Copilot runtime connection (`client.GetAuthStatus` reported
`IsAuthenticated: true`) -- this is not the zero-execution fallback.

## Criteria

| Criterion | Status | Evidence |
|---|---|---|
| S1 -- permission round-trip | **PROVEN** | All 4 handler-constructible `rpc.PermissionDecision` variants (`ApproveOnce`, `Reject`, `UserNotAvailable`, `NoResult`) exercised via a real `PermissionRequestShell`. `ApproveOnce`/`Reject`/`UserNotAvailable` each completed their turn (`turn_ended=true`); `NoResult` correctly left the turn unresolved for the full 90s deadline (sole responder, no other client answered -- consistent with the SDK's own doc comment). See `internal/copilotprobe/permission_test.go`. |
| S2 -- event-union handling | **PROVEN** | 33 distinct `SessionEvent.Data` variants observed in one probe session (`session.*`, `model.*`, `assistant.*`, `tool.*`, `permission.*`, `sandbox.decision`, `pending_messages.modified`, `user.message`). A `switch` explicitly handling only 2 of the 33 (`*rpc.AssistantMessageData`, `*rpc.AssistantTurnStartData`) reached `default:` 74 times across 31 distinct omitted types. See `internal/copilotprobe/events_test.go`. |
| S3 -- cancellation/shutdown | **PROVEN** | `Session.Abort` -> `Client.Stop` -> `Client.ForceStop` escalating ladder exercised under a deadline-bearing `context.Context`; all three returned promptly (`Abort` returned_promptly=true err=nil; `Stop` err=nil; `ForceStop` returned). No stage hung. See `internal/copilotprobe/shutdown_test.go`. |
| SQ-a -- Abort vs. blocked handler | **ANSWERED: NO** | `Session.Abort` returned promptly (`err=nil`, the genuine empirical signal) while a deliberately-blocked `PermissionHandlerFunc` remained parked. The "handler remains blocked" half of this conclusion follows from `PermissionHandlerFunc`'s signature carrying no `context.Context` (verified via `go doc`), not from an independently observable live unblock signal (this harness's blocking channel can only be closed by the test itself, so it cannot by construction serve as evidence of an SDK-driven unblock -- see the caveat recorded directly in `internal/copilotprobe/shutdown_test.go`). **Confirms** design section 3.1's conservative assumption ("until answered, assume it does not"). |
| SQ-b -- callback re-entrancy | **ANSWERED: SERIALISED** | Across 62 sampled `Session.On` callback invocations (each held ~15ms to make overlap observable) during one token-streaming turn, the observed maximum concurrent in-flight count was **1**. No overlapping invocation was observed in this sample. |
| SQ-c -- dispatch independence | **ANSWERED: YES** | With a `PermissionHandlerFunc` genuinely blocked (same technique as SQ-a), 42 further session events arrived strictly after the block signal (62 total vs. 21 observed before blocking began, from two independent counters -- see footnote[^counters]). Permission dispatch is independent of event dispatch; **no head-of-line blocking observed**. |
| SQ-d -- stable event identity | **ANSWERED: YES** | `SessionEvent.ID` (`rpc.SessionEvent.ID string`, doc: "Unique event identifier (UUID v4), generated when the event is emitted") is non-empty and unique across every observed event (76/76 unique in the S2 sample; 0 duplicates). `ParentID *string` additionally provides a linked-chain ordering signal. **This REFUTES design section 5.1's prediction that no stable identity exists.** |
| R5 -- `SendAndWait` doc drift | **CONFIRMED** | The pinned signature is `func (s *Session) SendAndWait(ctx context.Context, options MessageOptions) (*SessionEvent, error)` -- exactly 2 parameters (excl. receiver), confirmed via `reflect.TypeOf(session).MethodByName("SendAndWait")`. The method's own doc comment nonetheless documents a `timeout` parameter ("How long to wait for completion. Defaults to 60 seconds if zero.") that does not exist in the signature. Documentation drift confirmed exactly as F5/R5 predicted. |

## Additional finding: `rpc.PermissionDecision` variant-set surface (informs B2 scope)

The `rpc.PermissionDecision` interface has **16 concrete implementers** at the
v1.0.11 pin (excluding the internal `RawPermissionDecisionData` wire
fallback): `Approved`, `ApprovedForLocation`, `ApprovedForSession`,
`ApproveForLocation`, `ApproveForSession`, `ApproveOnce`,
`ApprovePermanently`, `Cancelled`, `DeniedByContentExclusionPolicy`,
`DeniedByPermissionRequestHook`, `DeniedByRules`,
`DeniedInteractivelyByUser`,
`DeniedNoApprovalRuleAndCouldNotRequestFromUser`, `NoResult`, `Reject`,
`UserNotAvailable`. Only **4** of these are documented by
`PermissionHandlerFunc`'s own doc comment as constructible handler-return
values: `ApproveOnce`, `Reject`, `UserNotAvailable`, `NoResult`. The
remaining 12 are server-reported outcome/result variants (e.g. after a
rules-based or host-policy auto-approval) observable via events/RPC
responses, not values a handler constructs. Because the constructible set is
exactly 4 -- the plan's own non-split threshold -- B2 was **not** split.

## Copilot CLI / runtime version validated against

`1.0.84-1` (SDK protocol version `3`, matching the `SDKProtocolVersion`
constant). Obtained via `client.GetStatus(ctx)` against the SDK's own
embedded/managed runtime connection (not necessarily identical to whatever
`copilot` binary may be on `PATH`, per `ClientOptions.BaseDirectory`'s doc
comment -- the SDK extracts and manages its own embedded CLI binary
independently of `PATH` resolution). **Q1 (D10) is resolved provisionally**:
`1.0.84-1` / SDK protocol `3` is recorded as the provisional CLI/runtime
floor validated against this pin.

## Executing environment

Live probe run against a real, authenticated Copilot backend connection
(`GetAuthStatus.IsAuthenticated == true`, logged-in-user auth via the
environment's `gh`/stored OAuth credentials -- `ClientOptions.UseLoggedInUser`
default). Windows amd64, Go `go1.26.5` toolchain, `go 1.24` language floor.
Each probe ran with a working directory rooted at a disposable `t.TempDir()`,
never the repository checkout. All four proving tests plus the S1 table
completed in well under their individual deadlines (longest: S1's `NoResult`
subtest, which correctly consumed its full 90s deadline by design).

## Go/no-go recommendation for C3

**GO**, with one explicit condition satisfied by B7: design section 4.2
requires the design to be amended with the chosen SQ-b re-entrancy fallback
before C2 may close. B7 performs that amendment using this artifact's SQ-b
(SERIALISED) and SQ-d (YES, stable identity) answers. With B7 complete, C2 is
closed and C3 may proceed. No criterion in this artifact was left UNPROVEN;
the zero-execution fallback was not invoked.

[^counters]: SQ-c's "62 total vs. 21 before" and "42 strictly after the
    block signal" figures come from two independent counters in
    `internal/copilotprobe/concurrency_test.go` (a before/after snapshot of
    total callback invocations, versus a separate atomic counter gated on a
    `blockedSignaled` flag set from a side goroutine) and are not expected to
    subtract to an identical value (62-21=41, not 42) -- this is a
    measurement-mechanism difference, not an arithmetic or test-logic defect
    (the test's own pass/fail condition, `postWindowEvents > preBlockEvents
    || afterBlock > 0`, does not depend on the two agreeing). Raised by
    standard review (Correctness Reviewer, P3).

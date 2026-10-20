---
title: "Backlogit CLI operating pitfalls and permission-outcome design lessons (059-S)"
date: 2026-10-20
tags: ["backlogit", "go", "hub", "concurrency", "shutdown", "ci-flake"]
shipment: 059-S
severity: medium
---

# Backlogit CLI pitfalls and permission-outcome lessons (059-S)

## 1. Never kill a running `backlogit` mutation, and never run two in parallel

Every `backlogit` command takes about two minutes to open the workspace and a mutating command can linger after its
work appears complete. Stopping `backlogit shipment claim` mid-run left a shipment operation record under
`.backlogit/ops/`; the next command ran lifecycle recovery and rolled the claim back to `queued`, which was only noticed
later. Parallel `backlogit` invocations contend on `.backlogit/.locks/shipment-lifecycle-global` and fail with
"gate in progress"; a silent partial success is possible. Run mutating commands one at a time, attached, and wait for
exit; verify the resulting status in the queue file afterwards.

## 2. Report permission delivery explicitly; never infer it from channel state

Inferring "the handler received the decision" from an empty capacity-1 reply buffer left a window where the hub published
`resolved(allow)` while the handler downgraded to UserNotAvailable. A one-shot outcome claimed atomically by exactly one
of the handler (`Deliver`) or the hub (`Revoke`) closes the window with no waiting. Settlement is then purely a function of
the recorded outcome.

## 3. A shutdown stage that waits on the child must outlive parent cancellation

Running `Agent.Stop` under the shutdown context lets a second Ctrl-C cancel the stage and exit while the CLI child is
still alive. Run the stage under `context.WithoutCancel` with a hard budget, hurry the adapter through its own derived
context, and surface an unconfirmed ForceStop as an explicit orphan-risk error (also from the Start-failure path).

## 4. Idle accounting: a Send error is delivery-uncertain

A prompt whose `Send` errored after being invoked may still have started a turn. Retracting its pending count on the
error double-counts against a later prompt (Copilot review thread on PR 179). Retract only for prompts provably never
delivered; keep uncertain ones busy and rely on the bounded idle wait.

## 5. Unrelated CI flake

`tools/gatecheck/internal/retiredarch` `TestMaskerDefinedExactlyOnce` fails intermittently under `-race` when parallel
tests remove a shared temp directory it walks. It passed on rerun; not caused by hub changes.

---
title: "Compacted memory - Engram daemon readiness circuit breakers"
date: 2026-09-28
status: compacted
source_count: 2
compacted_from:
  - docs/archive/memory/2026-09-17/circuit-break-engram-daemon-readiness.md
  - docs/archive/memory/2026-09-18/circuit-break-engram-daemon-readiness.md
---

# Compacted: Engram daemon readiness circuit breakers

## Outcome

Two circuit-breaker records documented the same class of failure: the Engram daemon or workspace
readiness path could not be trusted during Stage restore/prune or indexed-search handoff. Both were
preserved as diagnostics and are now compacted because later sessions proceeded by recording the
handoff degradation and preserving local checkpoint state.

## Failure chain summary

* The daemon readiness operation failed repeatedly enough to trip the universal circuit breaker.
* The affected workflow phase was Stage checkpoint recovery / prune-on-restore and later Stage
  escalation handoff.
* The correct disposition was fail-closed for the dependent workflow: do not prune or resume from a
  checkpoint when the installed Engram substrate is required but unreachable; preserve the active
  checkpoint and hand off to the operator.

## Decisions and rationale

* The daemon process was not killed because it was a foreign process and outside the acting Stage
  boundary.
* File-based degraded pruning was not used; the installed `agent-engram` protocol requires a
  fail-closed operator handoff when Engram is installed but unreachable at restore time.
* Diagnostic detail stays in the archived originals. The durable memory is the workflow rule: Engram
  unavailability blocks prune-on-restore and asynchronous escalation handoff, but does not authorize
  a speculative resume.

## Follow-up status

No open repository change is carried by these circuit breakers. They remain historical evidence for
future daemon-health troubleshooting and for validating the fail-closed recovery path.

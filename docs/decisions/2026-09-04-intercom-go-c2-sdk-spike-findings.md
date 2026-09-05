---
title: "Phase-C2 Copilot SDK proving spike: findings"
description: "S1-S3 plus SQ-a..SQ-d empirical findings for the pinned github.com/github/copilot-sdk/go@v1.0.11 dependency (shipment 005-S, sub-epic B)"
status: "in-progress"
source_document: "docs/plans/2026-09-04-intercom-go-c2-sdk-spike-plan.md"
linked_artifacts:
  - "docs/decisions/2026-09-04-intercom-go-implementation-design-reconciliation-deliberation.md"
  - "docs/decisions/2026-09-04-intercom-go-architecture-correction-copilot-sdk-deliberation.md"
  - "docs/design-docs/intercom-go-backend-architecture.md"
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

> This artifact is completed by B6. Each row below is a placeholder until
> then; a placeholder criterion is never a PASS.

## Criteria

| Criterion | Status | Evidence |
|---|---|---|
| S1 — permission round-trip | PENDING | (B2) |
| S2 — event-union handling | PENDING | (B3) |
| S3 — cancellation/shutdown | PENDING | (B4) |
| SQ-a — Abort vs. blocked handler | PENDING | (B4) |
| SQ-b — callback re-entrancy | PENDING | (B5) |
| SQ-c — dispatch independence | PENDING | (B5) |
| SQ-d — stable event identity | PENDING | (B3) |
| R5 — `SendAndWait` doc drift | PENDING | (B3) |

## Copilot CLI / runtime version validated against

PENDING (B6).

## Executing environment

PENDING (B6).

## Go/no-go recommendation for C3

PENDING (B6).

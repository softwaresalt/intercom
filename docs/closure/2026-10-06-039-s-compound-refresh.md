---
title: "039-S compound refresh"
description: "Evidence-backed maintenance of the backlogit shipment lifecycle learning after shipment 039-S."
date: 2026-10-06
mode: apply
shipment: 039-S
feature: 049-F
pr: 105
merge_commit_sha: ecf07ed66f992785b988993d0aa17c1a85b045d1
tags:
  - compound-refresh
  - backlogit
  - shipment
  - 039-S
---

# Compound refresh — shipment 039-S

## Entries reviewed

### `docs/compound/2026-05-07-backlogit-shipment-status-constraints.md`

- **Classification:** update.
- **Evidence:** The entry correctly describes why generic
  `backlogit move <shipment_id> --status shipped` is rejected and documents
  the distinction between archive-file location and declared status. Its
  resolution nevertheless instructed Ship to call
  `backlogit shipment ship` directly for cascade closure. The current
  `.github/agents/_ship.agent.md` Step 6 and
  `.github/skills/shipment-reconcile/SKILL.md` require Ship to invoke
  `mode: classify-close-path` and then `mode: safe-close` with the returned
  `CLASSIFICATION_BINDING`; only the skill's Cascade Close Sub-Procedure may
  dispatch the cascade CLI.
- **Change:** Updated the resolution and compounding-value sections to state
  the binding-carrying skill boundary and distinguish the engine-level CLI
  from the Ship-level invocation. Retained the generic-move failure evidence,
  shipment status constraints, and terminal-relocation explanation.
- **Traceability:** This correction is made explicit after the P-005 process
  deviation during 039-S closure. ORCH-D9 accepted the already archived
  state but did not reclassify the earlier direct CLI invocation as
  skill-mediated.

No other `docs/compound/` entry was found in scope. No consolidation,
replacement, deletion, or new standalone learning was needed.

## Follow-up

None created by this refresh. The P-005 deviation and existing deferred
scope captures are disclosed in the post-merge closure record; no additional
backlog item was inferred or created.

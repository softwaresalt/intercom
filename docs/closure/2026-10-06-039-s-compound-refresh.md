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

### `docs/compound/2026-09-30-shipment-reconcile-skill-bypass-manual-substitution-risk.md`

- **Classification:** update.
- **Evidence:** During 039-S closure, Ship directly invoked
  `backlogit shipment ship 039-S`, bypassing the required
  `classify-close-path` → binding-carrying `safe-close` sequence. ORCH-D9
  accepted the archived state and prohibited reversal. The authoritative
  post-mode report returned `PROCEED`; the separate read-only P-015
  classifier returned `cascade`, qualifying `049-F` with no
  out-of-manifest descendants. Neither post-hoc result supplies a
  retroactive classification binding.
- **Change:** Added the 039-S recurrence, the accepted no-reversal
  disposition, the post-mode and classifier evidence, and the boundary on
  what that evidence establishes. The entry now distinguishes terminal-state
  verification from conformant pre-mutation skill execution.
- **Traceability:** `.backlogit/reconcile/039-S-post-2026-10-06T07-24-42Z.md`
  and `docs/closure/039-S-049-F-post-merge-closure.md`. ORCH-D9 accepted the
  state; it did not erase the P-005 deviation.

No consolidation, replacement, deletion, or new standalone learning was
needed.

## Follow-up

None created by this refresh. The P-005 deviation and existing deferred
scope captures are disclosed in the post-merge closure record; no additional
backlog item was inferred or created.

---
title: "stash.jsonl carry-forward learning"
description: "Captured the operator-mandated invariant that stash.jsonl changes travel with Stage and Ship commits and never block pipeline gates."
date: 2026-09-19
tags:
  - "compound"
  - "backlogit"
  - "pipeline"
mode: "orchestrator"
shipment: "025-S"
---

## Outcome

Created
`docs/compound/workflow-issues/stash-jsonl-is-carried-pipeline-state-2026-09-19.md`.

The durable invariant is:

* `.backlogit/stash.jsonl` changes are carried pipeline state
* Stage and Ship preserve them onto their working branch
* Their branch commits include the carried changes
* Staging, cleanliness, topology, and branch-creation gates never block because
  `.backlogit/stash.jsonl` is modified
* Unrelated workspace changes remain subject to normal safety rules

## Files Modified

* `docs/compound/workflow-issues/stash-jsonl-is-carried-pipeline-state-2026-09-19.md`
* `docs/memory/2026-09-19/stash-jsonl-carry-forward-learning-memory.md`

## Decision

The previous 025-S dark-factory halt incorrectly treated
`.backlogit/stash.jsonl` as blocking dirt. Future Stage, Ship, and Orchestrator
flows must classify it as state to carry and commit.

This is confirmed by
`docs/memory/2026-09-04-stage-residual-hardening-triage-session.md`, which
records that mixed operator and Ship stash additions were preserved verbatim
and included in the staging commit.

## Next Step

Apply this invariant when 025-S pipeline execution resumes. Do not require
separate stash cleanup or disposition.

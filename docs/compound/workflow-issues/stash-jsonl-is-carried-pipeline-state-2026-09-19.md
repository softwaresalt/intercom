---
title: "Treat stash.jsonl changes as carried pipeline state"
description: "Stage and Ship must carry stash.jsonl changes onto their branch and commit them instead of blocking on a dirty-worktree gate."
source: "docs/compound/workflow-issues/stash-jsonl-is-carried-pipeline-state-2026-09-19.md"
doc_type: "learning"
problem_type: "pipeline state misclassification"
category: "workflow-issues"
component: "Stage, Ship, and Orchestrator backlog gates"
root_cause: "Gates treated modified stash.jsonl as unrelated workspace dirt instead of state that travels with pipeline operations."
resolution_type: "design_change"
severity: "high"
message: "STAGING_GATE_UNCOMMITTED: .backlogit/stash.jsonl is modified"
file_path: ".backlogit/stash.jsonl"
citations:
  - ".backlogit/stash.jsonl"
  - ".github/agents/_orchestrator.agent.md"
  - "docs/memory/2026-09-04-stage-residual-hardening-triage-session.md"
tags:
  - "backlogit"
  - "stash"
  - "stage"
  - "ship"
  - "pipeline-gates"
---

## Problem

Pipeline startup or handoff can observe `.backlogit/stash.jsonl` as modified
and classify that state as an unclean workspace or uncommitted Stage artifact.
That classification can halt the staging, branch-creation, topology, or Ship
routing gate even when the selected shipment is otherwise ready.

The modified state of `.backlogit/stash.jsonl` is not a valid gate blocker.
It is pipeline state that must travel with the work.

## Root Cause

The gate applied a broad rule that any modification under `.backlogit/`
represented unresolved staging work. That rule ignored the repository's
carry-forward contract for the stash ledger.

Stage and Ship operations can begin with or produce legitimate
`.backlogit/stash.jsonl` changes. Leaving those changes behind, reverting them,
or requiring a separate cleanup cycle breaks continuity between intake,
planning, implementation, and closure.

The established Stage precedent is explicit: mixed operator and Ship additions
were preserved verbatim and included in the staging commit because they shared
the stash ledger with Stage's archival work. The later gate failure repeated a
previously solved workflow mistake by treating that same carry-forward state as
blocking dirt.

## Resolution

Stage and Ship must always carry `.backlogit/stash.jsonl` changes forward onto
the branch used for their operation. The branch commits must include those
changes with the other release-unit artifacts.

All pipeline gates must treat a modified `.backlogit/stash.jsonl` as
carry-forward state:

* Do not fail staging or clean-worktree checks because this file is modified
* Do not revert, discard, hide, or leave the file behind on the prior branch
* Preserve the current file content when creating or switching to the Stage or
  Ship branch
* Include the file in the applicable branch commit
* Verify before handoff that the committed branch contains the carried version

This rule applies to Stage, Ship, and Orchestrator preflight logic. A gate may
report that the file will be carried, but it must not block because the file is
modified.

## Prevention

Implement clean-worktree and staging gates with an explicit
`.backlogit/stash.jsonl` carry-forward classification before evaluating
blocking dirt. Ensure Stage and Ship branch and commit checks include the file
when it differs from the branch base.

When diagnosing a dirty workspace, separate this ledger from unrelated files.
Handle unrelated changes under the normal workspace-safety rules without
turning `.backlogit/stash.jsonl` into a blocker.

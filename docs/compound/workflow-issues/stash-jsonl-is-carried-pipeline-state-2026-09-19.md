---
title: "Carry incomplete Stage and Ship state into the next operation"
description: "The next Stage or Ship branch must preserve and commit stash.jsonl plus every attributable uncommitted or unpushed change from the prior pipeline operation."
source: "docs/compound/workflow-issues/stash-jsonl-is-carried-pipeline-state-2026-09-19.md"
doc_type: "learning"
problem_type: "pipeline state misclassification"
category: "workflow-issues"
component: "Stage, Ship, and Orchestrator backlog gates"
root_cause: "Pipeline handoff treated attributable uncommitted or unpushed state as unrelated workspace dirt instead of carrying it into the next authorized branch and commit."
resolution_type: "design_change"
severity: "high"
message: "A prior Stage or Ship operation left changes uncommitted or unpushed, causing the next clean-worktree gate to block itself."
file_path: ".backlogit/stash.jsonl"
citations:
  - ".backlogit/stash.jsonl"
  - ".github/agents/_orchestrator.agent.md"
  - ".github/agents/_stage.agent.md"
  - ".github/policies/workflow-policies.md"
  - "docs/memory/2026-09-04-stage-residual-hardening-triage-session.md"
tags:
  - "backlogit"
  - "stash"
  - "stage"
  - "ship"
  - "pipeline-gates"
  - "branch-handoff"
  - "carry-forward"
---

## Problem

Pipeline startup or handoff can observe `.backlogit/stash.jsonl` or another
artifact from the preceding Stage or Ship operation as uncommitted or
committed only on a local branch. Treating that state as unrelated workspace
dirt can halt the staging, branch-creation, topology, or Ship routing gate even
when the selected shipment is otherwise ready.

This creates a self-failing sequence. The prior operation leaves required state
behind, then the next operation rejects that same state under its clean-worktree
or remote-presence gate. Discarding the state avoids the gate only by losing
pipeline history and is not a valid resolution.

## Root Cause

The gate applied a broad rule that any modification under `.backlogit/`
represented unresolved staging work. More generally, the handoff assumed the
prior operation had committed and pushed every change without first proving
that premise.

Stage and Ship operations can begin with or produce legitimate
`.backlogit/stash.jsonl` changes along with plans, backlog artifacts, source,
tests, review fixes, and closure evidence. Leaving any attributable prior-
operation change behind, reverting it, or starting the next branch from a base
that does not contain it breaks continuity between intake, planning,
implementation, and closure.

The established Stage precedent is explicit: mixed operator and Ship additions
were preserved verbatim and included in the staging commit because they shared
the stash ledger with Stage's archival work. The later gate failure repeated a
previously solved workflow mistake by treating that same carry-forward state as
blocking dirt.

## Resolution

`.backlogit/stash.jsonl` always travels forward onto the next Stage or Ship
branch and is included in the next applicable commit.

At every Stage-to-Ship, Ship-to-Stage, or same-role handoff, inventory all
changes produced by the preceding pipeline operation and verify that they are
both committed and present on the expected remote branch. If either condition
is false, the next operation must preserve those changes on its authorized
branch and include them in its next commit and push path. Where policy reserves
the push to the operator, the operation must preserve the commit and request
that operator-owned push rather than pretending the remote is current.

Use this disposition matrix:

| Observed state | Required disposition |
|---|---|
| `.backlogit/stash.jsonl` differs from the branch base | Always preserve it on the next authorized Stage or Ship branch and include it in the next applicable commit |
| A file is an attributable, uncommitted output of the prior Stage or Ship operation | Carry it onto the next authorized branch and include it in the next commit |
| A prior-operation commit exists locally but is not pushed | Preserve the commit in the next operation's ancestry or otherwise carry its exact changes forward, then complete the policy-authorized push path |
| The change is already committed and present on the expected remote | No carry-forward repair is needed |
| A dirty file cannot be attributed to the prior pipeline operation | Do not absorb it silently; halt for ownership clarification under normal workspace-safety rules |

The carry-forward rule does not grant Stage, Ship, or Orchestrator new Git
authority. It changes the handoff treatment of known pipeline state: preserve
first, commit on the next authorized branch, and use the existing actor-bound
push and merge path.

## Prevention

Before ending each Stage or Ship operation:

* Inventory every changed file and local-only commit produced by the operation
* Commit all attributable changes on the authorized branch
* Complete or explicitly hand off the policy-authorized push
* Verify the remote contains the handoff commit before claiming the next
  operation is starting from a clean baseline

At the next operation's preflight, classify dirty or local-only state by
provenance before applying the clean-worktree gate. Carry known prior-operation
state forward. Treat only unrelated or unattributed changes as blocking dirt.

This ordering prevents the pipeline from creating its own blocker: a prior
operation cannot leave state behind and then allow the next operation to fail
because that state still exists.

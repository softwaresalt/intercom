---
title: "Ship Step 6 covering-feature completion + cascade close must run on the post-merge branch, not main"
description: "Executing the Step 6 item 1 (covering-feature transition, reconciliation, cascade close) sequence before creating the post-merge/{feature_slug} branch commits backlog archival directly to main"
source: "docs/compound/workflow-issues/ship-step6-post-merge-branch-ordering-2026-09-19.md"
doc_type: "learning"
problem_type: "process-ordering"
category: "workflow-issues"
component: "ship-agent-step-6-post-merge-closure"
root_cause: "Step 6's numbered items (1: close the shipment, 2: operational-closure, ...) read as a flat sequence, but item 1's own sub-steps (covering-feature completion gate a1, pre-archive reconciliation a, close-path selection b, backlog commit e) must not execute until Step 6.0's post-merge/{feature_slug} branch already exists and is checked out. Reading Step 6 top-to-bottom without first locating and executing 6.0 lets the backlog-archival commit land on main."
resolution_type: "workaround"
severity: "medium"
message: "chore: archive {shipment_id} backlog artifacts committed on main instead of post-merge/{feature_slug}"
file_path: ".github/agents/_ship.agent.md"
citations:
  - "docs/closure/024-S-027-F-post-merge-closure.md"
  - "PR #64 (shipment 024-S / feature 027-F)"
tags:
  - "ship-agent"
  - "post-merge-closure"
  - "branch-protocol"
  - "P-010"
---

## Problem

During post-merge closure for shipment 024-S, the Ship Step 6 item 1 sequence (covering-feature
completion gate `a1`, pre-archive reconciliation, `classify-close-path`, cascade
`backlogit shipment ship`, post-mode reconciliation, and the `.backlogit/` commit) was executed
immediately after confirming the merge, while still on `main` (freshly fast-forwarded from
`origin/main`). This produced a commit (`chore: archive 024-S backlog artifacts ...`) directly
on local `main`, in violation of Step 6.0's non-negotiable rule that all Step 6 closure commits
must land on a `post-merge/{feature_slug}` branch, never on `main` directly.

## Root Cause

The agent instructions present Step 6.0 ("Post-Merge Branch Protocol") as its own subsection,
immediately followed by a numbered list ("1. Close the shipment ... 2. Invoke
`operational-closure` ... 3. Evaluate documentation ..."). It is easy to read the numbered list
as the literal execution order and start at item 1 without first re-confirming that 6.0's branch
creation is a **prerequisite gate**, not merely an earlier list entry. Because item 1's sub-steps
(especially the covering-feature `active -> done` transition, which is an authorized mutation in
its own right) look like they can run standalone right after the Merge Confirmation Gate, the
branch-creation step gets silently skipped.

## Resolution

Caught before any push (nothing had left the local session): created
`post-merge/{feature_slug}` from the tip of the errant `main` commit, ran
`git reset --hard origin/main` on `main` to restore it to exactly the merge commit, and
re-verified `main` was no longer ahead of `origin/main` before continuing all subsequent Step 6
work (item 1's already-completed mutations, plus items 2–12) on the new branch. The errant
commit's content was preserved (it became the first commit on the closure branch), so no
backlog-state work was lost or redone.

## Prevention

Before executing **any** part of Step 6 item 1 (even read-only reconciliation checks that
precede a mutation), explicitly perform this checklist:
1. Has Step 6.0 already run in this session? (`git branch --show-current` should report
   `post-merge/{feature_slug}`, not `main`.)
2. If not on a `post-merge/*` branch yet, stop, run Step 6.0 exactly as written (checkout
   `main`, pull, `git checkout -b post-merge/{feature_slug}`), confirm the checkout, and only
   then resume at item 1.
3. Treat "Merge Confirmation Gate passed" as a trigger for Step 6.0, not for item 1 directly —
   the numbered list under Step 6 only begins **after** 6.0 has produced and checked out the
   closure branch.

A cheap guard for future sessions: run `git branch --show-current` immediately before the first
mutating call in item 1 (the covering-feature `backlogit move ... --status done`, or the
pre-archive reconciliation gate) and halt if it does not start with `post-merge/`.

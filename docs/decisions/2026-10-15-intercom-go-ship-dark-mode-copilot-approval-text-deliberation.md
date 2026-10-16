---
title: "Ship dark-mode text: P-018 Copilot review block and per-closure destructive approval (8E1E2D80)"
description: "Stage deliberation over stash 8E1E2D80 (P-021 C6), with 724FA2FB read as a related, separate expansion: merge-approval contract wording in the Ship agent (rendered)"
topic: "Ship Step 7b and Step 14 dark-mode text: P-018 Copilot review block applies once Copilot is engaged; Step 14 dark-mode approval list must include the P-018 PASS requirement; orchestrator autonomy text must carve out per-closure destructive approval"
depth: "standard"
decision_status: "decided-pending-operator-confirmation"
promoted_to: "none (upstream request recorded; no local harvest)"
linked_artifacts:
  - "docs/closure/043-S-053-F-post-merge-closure.md"
  - ".github/agents/_ship.agent.md"
  - ".github/policies/workflow-policies.md"
tags:
  - "ship"
  - "dark-mode"
  - "p-018"
  - "p-014"
  - "p-017"
  - "upstream-render"
  - "stage"
---

# Deliberation: Ship dark-mode approval text (8E1E2D80)

## Session context

* Stage, Orchestrator-invoked, sequential. `ROUTING_DEGRADED` declared (see the 3750C37C artifact).
* P-021 C6 applies; the entry itself says the merge approval contract change "needs operator
  deliberation". Priority high, kind defect.
* **Duplicate scan (A):** `724FA2FB` (P-014 merge approval cannot satisfy per-closure destructive
  approval; step 6.1.b.3) is a **related but different** expansion. It is not merged; both are
  carried to the same upstream request so one template change covers both.
* **Late-identifier reconciliation (B):** PR and thread `N/A`; the 043-S closure record cites the
  entry; nothing further found. The `N/A` stands.

## Premise verification (read-only)

* P-018 (policies lines ~592-614) already says the gate is conditional on Copilot review being
  enabled and that, in dark factory mode (P-017), it is "preserved in full": a `DARK_MODE_ACTIVE`
  record does not satisfy it and admin fallback may never bypass a `COPILOT_REVIEW_BLOCK`.
* `_ship.agent.md` line ~582 treats the shadow-review loop as optional/advisory. The defect is
  that the dark-mode text of Step 7b and Step 14 can be read as keeping the timeout advisory even
  when Copilot is engaged. The policy is authoritative; the agent text under-states it.
* Both are weakenings by omission, not by contradiction, so the fix is a wording change that
  restates existing policy. It does **not** change the approval contract (which is why it can be
  routed rather than escalated as a policy amendment).

## Options

| Option | Description | Verdict |
|---|---|---|
| A | Hand-edit `_ship.agent.md` | Rejected (rendered file; overwritten, and a hand edit to approval text in one workspace is the worst place to drift). |
| B | Upstream template change restating P-018/P-014/P-017 in Steps 7b and 14 | Chosen. |
| C | Amend P-014/P-017 to add a per-closure approval rule | Rejected as unnecessary: INV/closure text already requires a per-closure approval; wording, not policy, is missing. Reopen only if the operator wants the carve-out to be a new rule. |

## Decision (D-8E1E-1)

Upstream request with three wording items plus 724FA2FB:

1. Steps 7b and 14: the P-018 Copilot review block applies once Copilot is engaged; the
   shadow-review timeout is advisory only when Copilot is not engaged.
2. Step 14 dark-mode approval list includes the P-018 PASS requirement.
3. Orchestrator autonomy text: per-closure destructive approval is never a routine coordination
   decision (also covers 724FA2FB: a P-014 "merge approved" message cannot satisfy it).

No local edit. Priority stays high (security-adjacent merge gate wording).

## Open question for the operator

Confirm the upstream route. Operator caution until the re-render: treat dark-mode merge approval
as insufficient for any cascade-close; require an explicit per-closure approval.

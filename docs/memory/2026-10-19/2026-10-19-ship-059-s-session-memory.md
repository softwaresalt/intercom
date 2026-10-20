# Ship session memory - 059-S (MVP S4: permission arbitration, operator actions, shutdown)

- Shipment: 059-S, covering feature 067-F, tasks 067.001-T..067.008-T (all done).
- Branch: feat/059-s-mvp-s4-permission-arbitration-operator-actions-shutdown.
- Mode: DARK_MODE_ACTIVE (P-017); operator AFK; admin fallback not authorised.

## Items completed

All eight tasks implemented test-first in `internal/hub` (registry/CAS, empty broadcast set, hub-owned timeout and reply hand-off, hydrate carrying pending permissions, ADR MVP-3 authorization, operator prompt/abort routing, shutdown executor, real-adapter Stop/ForceStop shutdown) with supporting changes in `internal/agent` and `internal/agent/copilot`.

## Decisions with rationale

- Delivery of a permission decision is reported by the handler through an atomic `PermissionOutcome` (Deliver/Revoke), not inferred from the reply buffer, so resolved/aborted always matches what the SDK received.
- Shutdown stage for the agent runs under a hard budget on a context the parent cannot cancel; an unconfirmed ForceStop is surfaced as `ErrOrphanRisk` so a live CLI child is never reported as a clean exit.
- Shutdown step order amended in the design (abort, positive idle, quiesce, stop) because latching first discards the idle event.
- Unknown permission kinds are destructive per ADR MVP-3: denied for remote approvers unless `remote_approve_destructive`; captured as a deliberation stash rather than changed here (P-021).
- Hub-added reason codes `agent_unavailable`, `shutting_down`, `client_closed` recorded in the ADR addendum.

## Review

Three review-fix cycles. Cycle 1 (8 persona subagents plus adversarial consensus) found one P1 (Agent.Stop abandoned before ForceStop) and several P2; all fixed. Cycles 2 and 3 were delta reviews; no HIGH-confidence P0/P1 remained.

## Follow-up stash (DEFERRED SCOPE EXPANSION)

026CF750, 550437DB, 4AFE4DA4, 65A52C5C, 8769C9A0, 70315264. Pre-existing stash E02CA4AA, 6738A8E5, 2C74C5FC, DA760FBB, E7B78125 were not absorbed.

## Operational notes

- `backlogit` commands take about two minutes to open the workspace and may linger after finishing; never kill a shipment claim mid-run (a killed claim was rolled back by lifecycle recovery and had to be repeated). Do not run backlogit commands in parallel (lifecycle lock).

## Next steps

Open the PR, run CI and Copilot review, merge with a merge commit, then post-merge closure.

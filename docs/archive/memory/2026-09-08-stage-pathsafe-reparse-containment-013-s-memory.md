---
title: "Stage session memory — 013-S / 014-F pathsafe reparse containment"
date: 2026-09-08
agent: Stage
mode: DARK_MODE_ACTIVE
shipment: 013-S
---

# Stage Session Memory — 2026-09-08 — Shipment 013-S / Feature 014-F

## Session parameters

- Mode: `DARK_MODE_ACTIVE`, operator AFK, visibility CLI-only (agent-intercom unavailable)
- Scope: exactly 22 active stash IDs, no additions permitted
- `shipment_limit`: 1 · `merge_approval_pre_authorized`: true · `admin_fallback_pre_authorized`: false
- Queue items at start: 0 · Active recovery checkpoints at start: 0 (one resolved `ship` checkpoint present, not actionable)

## Tooling status

- `TOOL_OK: backlogit` v1.10.1 (CLI; `features.sizing`, `shipments`, `stash`, `dependencies`, `checkpoints` all true)
- `INDEX_SYNC_OK` at session start (168 artifacts) and at session end
- **`TOOL_DEGRADED: complexity`** — registry advertises `features.sizing: true`, but this workspace's
  `task` artifact template defines **no structured `complexity` field**
  (`backlogit update <id> --complexity` fails validation). `size` written structurally;
  `complexity` preserved as enum-validated labeled prose at the head of each task description.
- `agent-intercom` not installed → no broadcasts; all visibility via this report and committed artifacts.

## Triage outcome

22 entries classified into 4 clusters. **Cluster A (`internal/pathsafe` containment security)
selected** — it holds the only `critical` entry in scope (`700B41CE`), an empirically
reproduced, unprivileged, live containment bypass in a control the Brief marks NON-NEGOTIABLE.

| Cluster | Entries | Outcome |
|---|---|---|
| A — pathsafe containment | `700B41CE`, `F133AB7E`, `2362BBB5`, `AD0D9D1F` | **SELECTED → 014-F / 013-S**, archived |
| B — retired-arch gate | `24B63533`, `8988120D`, `6285779C`, `6C24E2E4`, `8D603E29`, `9A14D3B7`, `E403E3C5`, `09CD6ACF`, `F4F4A959` | Deferred — strongest candidate for next cycle |
| C — docs/script hygiene | `F47DB9A9`, `2787DA56`, `4C5BEC23`, `1C6C3B46`, `9D45E62E` | Deferred — docs rank last |
| D — trackers / blocked | `A92E3FA0` (retain), `BF5DE670` (retain), `EF9352FB` (blocked), `4989A42D` (**archived**) | See below |

## Key decisions

- **D-1** Cluster A selected; a hygiene gate cannot outrank a live security bypass.
- **D-2** `F133AB7E` (GO-14 flip) is **in-scope, not scope expansion** — `root.go`'s register
  contains an explicit flip procedure naming that stash ID, and the operator placed it in scope.
- **D-3** Fix via reparse-aware resolution (option a); module-wide `GODEBUG` pin rejected.
- **D-4 (revised)** F3 ancestor relaxation **removed from scope** after review found it would
  have introduced a new bypass. Shipment is now monotonically fail-closed.
- **D-4a** Corrected root cause: containment is checked against a **non-canonicalized path**;
  the bypass is not final-component-only.
- **D-9** `4989A42D` archived — completion condition MET for the first time.
- **D-10** `EF9352FB` isolated into no shipment; operator-only credential rotation.
- **D-11** Staging artifacts reach `main` via **branch → PR → merge**, never direct commit.

## Adversarial review

Three independent models in parallel (Security `gpt-5.6-sol`, Correctness `gemini-3.8-flash`,
Scope `claude-opus-4.8`) reviewed plan rev 1 and found **1 P0 + 4 P1**, all resolved in rev 2:

1. **A1 (P0-equivalent)** Final-component-only fix left a residual bypass — intermediate
   junction with an existing leaf (C2) and existing dir beyond the junction (C3).
2. **A2 (P0)** Relaxing the ancestor branch would have **introduced** a bypass.
3. **A3 (P1)** Case-folding fail-open — deferred, closure explicitly de-scoped.
4. **A4 (P1)** Win32 syscalls need build-tagged files → task split.
5. **A5 (P1)** Message-constant change breaks 2 existing tests → 9-site audit AC added.

## Artifacts

- `docs/decisions/2026-09-08-intercom-go-pathsafe-reparse-containment-deliberation.md`
- `docs/plans/2026-09-08-intercom-go-pathsafe-reparse-containment-plan.md` (rev 2, PASS cycle 2)

## Outputs

- Feature **014-F**, tasks **014.001-T … 014.008-T**, shipment **013-S** (`queued`, 9 items, 0 unsized)
- 9 dependency edges recorded
- 5 stash entries archived with forward-reference dispositions; 17 remain active, all annotated

## Open items for the operator

1. **`EF9352FB`** — rotate the Tavily API key (external console; time-sensitive).
2. **`WINDOWS_GATE_REQUIRED`** — unset, so Windows-only tests cannot block CI. Flip is a
   repository-configuration decision; Stage/Ship must not self-authorize it.
3. **`4989A42D` archival** — autonomous, contract-prescribed, non-destructive and retrievable.

## Next step

Ship claims shipment **013-S**. Execution order: `014.001-T`, `014.002-T`, `014.003-T` →
`014.004-T` → `014.005-T`, `014.006-T` → `014.007-T` → `014.008-T`.

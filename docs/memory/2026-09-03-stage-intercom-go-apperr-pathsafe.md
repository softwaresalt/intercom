---
title: "Stage session — intercom-go P1 (apperr + pathsafe), shipment 002-S"
date: 2026-09-03
agent: stage
session: stage-002-s-intercom-go-apperr-pathsafe
phase: complete
shipment: 002-S
feature: 002-F
stash_processed: ["4989A42D", "037B1552"]
stash_created: ["8ACF7110"]
merge_base: "87b800d5d3a6457d86fecd12fb7a299f170af718"
behavioral_oracle: "softwaresalt/agent-intercom @ 41df772"
---

# Stage Session Memory — intercom-go P1

## Entry state

* Branch `main` @ `87b800d` (PR #5 merged). Backlog, queue, shipment list all empty.
* No checkpoints (`total: 0`, `needs_quarantine: 0`) — zero-candidate normal startup.
* One hook event (`seq 1`, `ship_shipment 001-S active→shipped`), consumed and acked.
* Tool status: **DEGRADED_MODE** — backlogit MCP tools not present in this session's
  tool surface; used the registry-declared CLI fallback (`backlogit` v1.10.1) throughout.
  Not a silent fallback: declared before any pipeline work (P-012).

## Decisions

1. **Oracle correction (P0).** `references/herdr` is **not** the behavioral oracle — it is
   `ogulcancelik/herdr`, a Rust terminal multiplexer (0 hits for `acp`, `AgentDriver`,
   `clearance`, `socketmode`, `ServerMode`, `stall_alert`). The real oracle is
   `softwaresalt/agent-intercom` at `C:\Source\GitHub\intercom` @ `41df772`, confirmed by
   remote URL, brief §4 module map, 21,762 `src/` LOC, and exact 56/22/46/6 corpus counts.
   Both stash entries' "ORACLE MOUNT RESOLVED … references\herdr" annotations were false
   and have been corrected. `references/herdr` left untouched.

2. **MCP-retirement blocker CLOSED (D1).** Split into three surfaces:
   retire MCP **as a server mode** (no `internal/mode`, no `McpDriver`, no
   `--mode`/`--transport`, no stdio transport); **retain the MCP HTTP tool endpoint** as
   phase P11 (`internal/agenttools`) because `src/main.rs:365-372` shows `start_sse` has
   no mode guard and ACP subprocesses reach `check_clearance`/`transmit`/`auto_check`
   through it (HITL-003/FR-032); retain the `protocol_mode` column (write `'acp'`,
   read-tolerant) and the 9 wire tool names. Config omits any mode field (mode was
   CLI-only in the oracle); `host_cli` becomes unconditionally required.

3. **Group handled as one planning unit, one bounded harvest.** `4989A42D` + `037B1552`
   deliberated together as a single roadmap; only the dependency root (P1) harvested.

## Artifacts

| Artifact | Path |
|---|---|
| Deliberation | `docs/decisions/2026-09-03-intercom-go-port-continuation-deliberation.md` |
| Plan (hardened, reviewed) | `docs/plans/2026-09-03-intercom-go-apperr-pathsafe-plan.md` |
| This memory | `docs/memory/2026-09-03-stage-intercom-go-apperr-pathsafe.md` |

## Review outcome

Plan review **PASS at attempt 3**. Attempt 1: 1×P0 + 12×P1 across Security Lens, Go,
Architecture, and Scope reviewers. Attempt 2 (Correctness) re-verified every attempt-1
closure against the plan body and found 2 new P1s introduced by the resize. All closed.
Three load-bearing Go path claims were confirmed empirically rather than assumed
(`Clean("")`→`"."`, `Join(root,"")`→`root`, `VolumeName("C:foo")`=`"C:"` while
`IsAbs`=`false`).

## Backlog created

`002-F` → `002.001-T` (A1–A3), `002.002-T` (B1–B4), `002.003-T` (C1) = 12 items,
9 dependency edges, all sequential. Shipment **`002-S`** (`queued`, 12 items, parent-first).

**Sizing degradation flagged:** the registry declares `features.sizing: true`, but this
backend's templates define `size` on `task` only and `complexity` on **no** type. Size and
complexity were set structurally on the 3 tasks and recorded as labeled prose on the
feature and all 8 subtasks.

## Deferred

* `037B1552` → rewritten as the P2 (config) / P3 (credentials) successor; blocker resolved.
* `4989A42D` → rewritten to cover P4–P15 with explicit successor ordering.
* `8ACF7110` → **new**: machine-checkable oracle-drift gate (residual P2 finding ARCH-5).
* Untouched by operator instruction: `A92E3FA0`, `EFFAA358`, `90350C9A`, `F7C6420D`,
  `BEDD2E70`, `35D76D5E`, `90EE7758`, `EF9352FB`.

## Dirty-state fingerprint (must be preserved)

* `.gitignore` — modified, **unstaged**. SHA-256
  `E4D7071C607C953D9BF330336DD189CD76C94B922D6FCCE17F0105E6D031E291`.
* `.claude/` — untracked, never staged.
* `.backlogit/hooks_queue.jsonl` — untracked machine-local runtime state, never committed.
* `references/herdr` — ignored via `.git/info/exclude`, never staged; oracle at
  `C:\Source\GitHub\intercom` used strictly read-only (its 27 pre-existing dirty files
  untouched).
* `.backlogit/stash.jsonl` + `.backlogit/queue/` — backlog source state, **committed** as
  the legitimate output of this Stage cycle.

## Next step

Ship claims shipment `002-S`. Do **not** hand Ship a feature ID.

---
title: "Stage session memory — first staging round (design-doc reconciliation + C2)"
date: 2026-09-04
agent: stage
session: "intercom-go first staging round"
shipment: "005-S"
feature: "006-F"
tags: [stage, staging, reconciliation, copilot-sdk, phase-c2]
---

# Stage session memory — 2026-09-04 first staging round

## Outcome

Queued shipment **`005-S`** — *intercom-go C2: design reconciliation + Copilot
SDK proving spike*. 12 items: feature `006-F`, tasks `006.001-T` / `006.002-T`,
9 subtasks. 9 dependency edges. `backlogit doctor`: no issues.

## Entry state

Clean slate: 0 active/queued shipments, 0 checkpoints, all 001-F–005-F /
001-S–004-S archived. 19 active stash entries. Working tree dirty with
unrelated operator work.

## Tooling posture

* **`TOOL_DEGRADED`** — no backlogit MCP tools in this runtime; operated via
  CLI fallback at `C:\Tools\backlogit.exe` (v1.10.1) throughout.
* **`INDEX_SYNC_OK`** at session start and end (81 → 94 artifacts).
* No agent-intercom surface; broadcasts degraded to local structured output.

## Key decisions

* **D10** — Q1 (Copilot CLI version floor) is **spike-resolved, not
  spike-blocking**. Stash `A92E3FA0` said "Q1 blocks C2"; the governing
  deliberation says the spike must *record* the validated version and that
  becomes the floor. Q1 is an output of C2. Confirm-by-exception.
* **D11** — `docs/design-docs/intercom-architecture-implementation-design.md`
  (IMPL-DESIGN) is a **candidate vision document, not governing**. Design rev 2
  + the correction deliberation remain governing.
* **D12** — "ACP" retired as a product term in **both** expansions (Agent
  Client Protocol *and* Agent Control Plane).
* **D13** — First round = Option C: reconciliation governance gating the C2
  spike, one shipment.
* **D14** — Deferred-item register (Q7, R8, adoptable event-bus/TUI material)
  written into design rev 2 so it survives the untracked file.

## Reconciliation result (7 conflicts)

REJECT: RC-1 "ACP" term, RC-2 sidecar supervision/Named Pipes, RC-3
React/iOS/VAPID. AMEND: RC-4 "MITM/raw token streams", RC-7 `.intercom/session.json`
over-scope. DEFER: RC-5 clean-room delegation engine (≥ C9, gated on Q6), RC-6
`git reset` circuit breaker (risk **R8**). ADOPT: native SDK hosting, projection
semantics, the event-bus design, the Bubble Tea TUI layout.

## Plan-review history

* **Attempt 1 — FAIL.** Decisive: design rev 2 defines **four required C2 spike
  questions** (§3.1 SQ-a, §4.2 SQ-b, §5.1 SQ-d, §5.4 SQ-c) that the first
  decomposition missed; §4.2 carries an explicit *"C2 may not close until the
  design has been amended"* condition. Added tasks B5 + B7. Also: the
  `go mod tidy` build-tag claim was factually wrong; R5 was mislabelled.
* **Attempt 2 — FAIL.** Two remediations under-delivered: **G6 was inert**
  (markdown table `\|` escaping read as an ERE escaped pipe → matched a literal
  never-occurring string, so the gate could never fail), and the UNPROVEN
  fallback was applied to B6 only. Plus a new MD041 defect in A1.
* **Remediated → PASS** on the reviewer's explicit conditional. All
  pipe-bearing gate commands moved into a fenced `bash` block; **G6 empirically
  verified** (exits 0 clean, exits 1 on `hostCLI`/`channelID`).

## Harvest degradation (P-012 honest record)

`features.sizing: true` in the registry is **not sufficient** — sizing is
WIT-gated per artifact type. Verified on this workspace:

| type | `size` | `complexity` |
|---|---|---|
| feature | ✗ | ✗ |
| task | ✓ | ✗ |
| subtask | ✗ | ✗ |

So `size` was set structurally on the two tasks only (`S`, `L`;
`size_source: agent`, `size_ruleset_version: v1`); every other size and **all**
complexity values were preserved as enum-validated prose (`Size: … |
Complexity: …`). Probed with stderr visible — `--complexity low` returns exit 1
on backlogit 1.10.1. Matches `docs/compound/2026-09-04-backlogit-sizing-is-wit-gated.md`.

## Stash disposition

**Zero entries archived — correct, not an omission.** No stash entry was
consumed: `A92E3FA0` is a standing C3–C11 roadmap tracker (updated in place, not
archived), `4989A42D` still tracks live contamination, and the 17 task entries
were all deferred (15 carry the `DEFERRED SCOPE EXPANSION` marker and are
forced onto the `deliberate` route by P-021 C6 before they can be planned).

P-021 C5/C6 work completed on `B6203CCA` and `78D13775` (the only two
unreconciled entries): **PR #15** recovered from
`docs/closure/004-S-005-F-post-merge-closure.md`; review-thread ID stands `N/A`
(terminal — threadless pre-PR local review); **duplicate scan CLEAN** for both.

## Next session

1. Ship claims `005-S`. HC-1 (D10) is confirm-by-exception, so B may proceed.
2. If no Copilot CLI is obtainable, take the **A-only closure path** — do not
   run B and report it as unproven.
3. After C2: the deferred-expansion backlog (15 entries) needs its own
   deliberation round; `8C2D578D` (apperr `KindSlack`/`KindIPC`/`KindACP`) and
   `7774C9CA` should be deliberated **together** — same taxonomy.
4. `8ACF7110` (oracle-drift gate) is likely **void** — D8 demoted the Rust repo
   from behavioural oracle to historical reference, so the pin it would guard no
   longer governs. Confirm before spending on it.
5. **`EF9352FB` is a live security exposure** — an unrotated Tavily API key in
   plaintext `.env.local` (verified present). Never committed, so not a repo
   leak, but rotation is an operator action outside both agents' authority.

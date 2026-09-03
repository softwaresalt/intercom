---
title: "Stage session — intercom-go foundation staging"
date: 2026-09-03
agent: stage
session_outcome: "shipment 001-S queued"
shipment_id: "001-S"
feature_id: "001-F"
tags:
  - "stage"
  - "go"
  - "intercom-go"
---

# Stage Session Memory — 2026-09-03

## Inputs

Two operator-selected source documents, both untracked at session start:

* `docs/decisions/2026-07-06-go-port-reference-brief.md` — self-declared authoritative spec
  for a feature-complete Go port of `agent-intercom` (Rust).
* `docs/design-docs/intercom-go-backend-architecture.md` — "Copilot Remote Multiplexer (Backend)"
  architecture design.

Operator preference: treat both as one coherent Go-backend-port planning source unless
grouping analysis found a concrete reason to split.

## Triage and grouping decision

Stashed both for traceability: `4989A42D` (port brief), `A92E3FA0` (architecture doc).
Both classified **feature-shaped**. Neither carried a `DEFERRED SCOPE EXPANSION` marker,
so no P-021 reconciliation or duplicate-scan obligation applied.

**Grouping analysis found a concrete reason to SPLIT.** The documents describe
architecturally incompatible systems. Three load-bearing conflicts (confirmed independently
by both the Scope Boundary Auditor and the anchor Architecture Strategist):

1. **Durability model** — the brief's `recover_state`/ADR-0011, crash respawn+resume,
   at-least-once steering with `origin_session_id` rebind, and retention pruning are
   *semantics* unimplementable on the architecture doc's in-memory `SessionState`;
   conversely late-joiner hydration is definitionally in-memory.
2. **Operator surface is embedded in the persistence schema** — `slack_ts` appears in 5 of
   7 tables, `thread_ts`/`channel_id` are session columns. Slack is not a pluggable
   presentation layer in the brief.
3. **Lifecycle ownership** — one server routing many `[[workspace]]` channels under
   `max_concurrent_sessions` versus one process per workspace with ports allocated by a
   Python parent.

Group A (port brief) staged; Group B (architecture doc) preserved active in stash,
annotated with the conflicts and a reviewer correction: `AgentDriver` is **agent-facing**,
so composing the multiplexer as a driver is a category error — it would need a separate
`OperatorGateway` seam.

## Artifacts produced

| Artifact | Path / ID |
|---|---|
| Deliberation | `docs/decisions/2026-09-03-intercom-go-architecture-reconciliation-deliberation.md` |
| Implementation plan (rev 2) | `docs/plans/2026-09-03-intercom-go-foundation-plan.md` |
| Covering feature | `001-F` |
| Sub-epic tasks | `001.001-T`, `001.002-T` |
| Atomic subtasks | `001.001.001-ST` … `001.001.004-ST`, `001.002.001-ST` … `001.002.004-ST` |
| Shipment (queued) | `001-S` — 11 items |

## Plan review — two attempts

**Attempt 1: FAIL.** Six personas dispatched (4 always-on same-model, Architecture
Strategist on the anchor route `gpt-5.6-sol`/high, Security Lens cross-model).
2 P0 + 12 P1.

The two P0s came from the Learnings Researcher and were **independently verified** against
the real instruction files before acceptance:

* `.github/instructions/ci-security.instructions.md` MUST-requires full-SHA action pinning
  and explicitly forbids version tags; rev 1 pinned `actions/setup-go` by version.
* `.github/instructions/technology-go.instructions.md` requires "Target Go 1.22 or later";
  rev 1 set `go 1.21`.

Root cause: rev 1 anchored on the port brief without consulting `.github/instructions/`.
**Lesson worth compounding:** an external spec does not override workspace instruction
files; the stricter workspace standard governs.

The Go Reviewer also refuted rev 1's single highest-rated risk. `go test -race` with
`CGO_ENABLED=0` does **not** silently disable the detector — `cmd/go` pre-flight-checks and
aborts (`go: -race requires cgo`, exit 2). The failure mode is fail-**closed**. Rev 1's
sentinel-package guard was over-engineering on a false premise and was removed.

**Attempt 2: PASS.** Slice narrowed to the genuinely fork-independent portion; all P0
closed; 5 P1 closed in-plan; 7 P1 deferred *with the work* to stash `037B1552`, each
carrying its diagnosis and recommended resolution.

## Decisions with rationale

* **Narrowed the slice.** Rev 1 staged config + error taxonomy as "fork-independent."
  Review established they are Group-A-specific (`slack_detail_level`, `SLACK_*`,
  `[[workspace]].channel_id`, `[database].path`, `Slack`/`Db`/`Mcp` error variants) and
  should be validated against the Rust oracle before being frozen as operator contracts.
* **No `internal/` packages created.** Empty placeholder packages prematurely ratify
  boundaries the unresolved fork may invalidate.
* **CI sequenced second**, not sixth, so every unit after the module skeleton is gated at
  authoring time.
* **Did not commit.** The working tree carries pre-existing torn submodule state
  (`.gitmodules` + pointer staged, `references` deleted). Any commit risks sweeping it in.

## Degradations recorded (P-012)

* **Backlog sizing partially unsupported.** This workspace's WIT metadata defines `size`
  only on `task`, and defines `complexity` on **no** type. Structured `size` was set on
  `001.001-T`/`001.002-T`; for the feature and all 8 subtasks, size and complexity were
  preserved as enum-validated labeled prose (`Size: X | Complexity: Y`) at the head of each
  description. No item was left unsized.
* **MCP tool surface unavailable to this agent**; all backlog operations ran through the
  registry-declared `cli_command` fallbacks. Not an ad hoc filesystem fallback.

## Open items for next session

1. **BLOCKING before `001.001.001-ST`:** confirm module path
   `github.com/softwaresalt/intercom-go` and the Go 1.22 floor.
2. **Architecture fork (blocks `A92E3FA0`):** which architecture governs `intercom-go`?
3. **Torn `references` submodule:** blocks the Rust oracle mount, which blocks `037B1552`
   and all parity work. Requires operator action; out of bounds for Stage and Ship.
4. **MCP retirement confirmation:** would drop the `Mcp` error variant and `internal/mode`.
5. `docs/compound/` is empty. This session's P0 lesson is the first strong candidate.

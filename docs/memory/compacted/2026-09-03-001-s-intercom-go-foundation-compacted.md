---
title: "Compacted memory — intercom-go Foundation (001-S)"
date: 2026-09-03
shipment: 001-S
feature: 001-F
status: shipped
pr: 4
merge_commit_sha: d85eaeaf4c9ee36e2fbc9e3b3a407ffc8ea2052a
compacted_from:
  - docs/archive/memory/2026-09-03-stage-intercom-go-foundation.md
  - docs/archive/memory/2026-09-03-ship-intercom-go-foundation.md
---

# Compacted Memory — intercom-go Foundation (001-S)

## Outcome

Shipment `001-S` shipped. PR #4 merged (merge commit
`d85eaeaf4c9ee36e2fbc9e3b3a407ffc8ea2052a`). Covering feature `001-F`, 2 tasks,
8 subtasks — all `done`, archived via the P-015 verified fully-covered-root
cascade close. Closure artifact:
`docs/closure/2026-09-03-intercom-go-foundation-closure.md`.

## Key decisions (with rationale)

* **Slice narrowed to fork-independent scope.** Config schema, credential
  resolution, and the `AppError` taxonomy are Group-A-specific and deferred to
  stash `037B1552` pending Rust-oracle validation, rather than frozen as
  operator-facing contracts prematurely.
* **No `internal/` packages created** — packages are created by the unit that
  gives them behavior, not speculatively (rejected a review suggestion to
  extract duplicated `cmd/intercom`/`cmd/intercom-ctl` logic for this reason).
* **CI sequenced second, not last** — every unit after the module skeleton is
  gated at authoring time.
* **Workspace instruction floor overrides external spec.** Plan rev 1 anchored
  on the external Rust port brief (`go 1.21`, version-tag action pins) without
  consulting `.github/instructions/`; rev 2 corrected to `go 1.22` floor +
  full-SHA action pinning per `technology-go.instructions.md` /
  `ci-security.instructions.md`. **Compounding candidate**: an external spec
  never overrides a stricter workspace instruction standard.
* **`go test -race` + `CGO_ENABLED=0` fails closed** (`cmd/go` pre-flight
  aborts, exit 2) — not a silent detector disable. A rev-1 sentinel-package
  guard built on the opposite (false) premise was removed as over-engineering.

## Failed approaches

* Plan rev 1: 2 P0 (workspace-instruction floor violations) + 12 P1 — full
  rewrite to rev 2, which passed with 0 P0, 5 P1 fixed in-plan, 7 P1 deferred
  with the work to stash `037B1552`.

## Pre-merge remediation (this closure session)

A stray commit (`0e82e2b`, "Configs") landed on the PR branch after the
recorded review HEAD, introducing 4 machine-local files
(`.mcp.json`, `intercom-go.code-workspace`, `.backlogit/hooks_queue.jsonl`,
an unreviewed `.gitignore` edit) never in scope for `001-S`. Reverted
(`08d0c4a`); verified tree-identical to the reviewed HEAD before merging. See
closure artifact for full detail.

## Follow-ups (stash-captured, non-blocking)

`EFFAA358`, `90350C9A`, `F7C6420D`, `BEDD2E70`, `35D76D5E`, `EF9352FB`,
`90EE7758` — deferred P-021 C2 findings (CI/hardening hygiene, toolchain-pin
maintenance documentation). Plus pre-existing deferred stash `037B1552`
(config schema / credential resolution / `AppError` taxonomy, pending Rust
oracle validation).

## Open items surfaced during staging (still relevant post-ship)

* **Architecture fork unresolved**: Group B (multiplexer architecture doc,
  stash `A92E3FA0`) remains active/unstaged pending an operator decision on
  which architecture governs `intercom-go` beyond this fork-independent slice.
* **Torn `references` submodule**: blocks the Rust oracle mount, which blocks
  stash `037B1552` and all parity work. Requires operator action.
* **MCP retirement confirmation**: would drop the `Mcp` error variant and
  `internal/mode` in a future slice.

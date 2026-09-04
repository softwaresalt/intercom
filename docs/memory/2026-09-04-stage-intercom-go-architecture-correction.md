# Stage Session Memory — intercom-go Architecture Correction

**Date:** 2026-09-04
**Branch:** `stage/architecture-correction-copilot-sdk`
**Mode:** standard sequential (not dark mode); no `agent-intercom` pack
**Route:** Stage `claude-opus-5` (high); escalation `gpt-5.6-sol`/openai/high (nested per-role; legacy flat empty → no P-013.6 ambiguity)

## Trigger

Corrective Stage lifecycle after an explicit operator architecture correction.
PR #13 (P3 credentials) was closed unmerged as obsolete; the correction
resolves the prior Stage halt and supersedes the Slack / custom-ACP assumptions
that plan was built on.

## Tool gate

`ALL_TOOLS_OK` · `INDEX_SYNC_OK` · checkpoint scan: **0 candidates, 0
quarantined** → normal startup (not a failure). Queue empty, no shipments, no
open PR, main current at `2a54791`.

## Corrected architecture (governing)

* Agent boundary = official `github.com/github/copilot-sdk/go`, pinned
  **v1.0.11** = `a550258d5c37bd662197536992a23d633bfe5804`.
* **No** ACP broker, ACP wire protocol, or headless-CLI lifecycle in
  intercom-go. The SDK's JSON-RPC-to-CLI transport is an SDK implementation
  detail (codec under `internal/`, unimportable; `RuntimeConnection` sealed).
* Local UX Bubble Tea; remote UX **SPA** (not iOS); Dev Tunnels transport; one
  process per workspace, autoharness-orchestrated.
* Slack fully retired. Rust repo demoted from oracle to historical reference.

## Artifacts

| Artifact | Path |
|---|---|
| Governing decision | `docs/decisions/2026-09-04-intercom-go-architecture-correction-copilot-sdk-deliberation.md` |
| Governing design (rev 2) | `docs/design-docs/intercom-go-backend-architecture.md` |
| Plan (hardened, review PASS) | `docs/plans/2026-09-04-intercom-go-architecture-correction-plan.md` |

Superseded and marked non-governing (frontmatter + banner + cross-link):
`2026-07-06-go-port-reference-brief.md`,
`2026-09-03-...-architecture-reconciliation-deliberation.md`,
`2026-09-03-...-port-continuation-deliberation.md`,
`2026-09-03-...-p2-config-deliberation-addendum.md`.

## Backlog

* Feature **005-F**; sub-epics `005.001-T` … `005.004-T`; 10 subtasks
  `005.001.001-ST` … `005.004.002-ST`.
* **Shipment `004-S`** (queued, 15 items) — the handoff token to Ship.
* 11 dependency edges (default `blocks` type; `blocked-by` is not a valid type
  in this backlogit build).
* Orphan `004-F` (created by a malformed first `add` call) archived.

### Degradation recorded

`features.sizing: true` in the registry, but this workspace's WIT metadata
defines **no `complexity` field for task or subtask, and no `size` field for
subtask**. Structured emission therefore failed. Size and complexity were
enum-validated and recorded as labeled prose on every task and subtask
(`Size: X | Complexity: Y`), per the harness contract's non-structured
fallback. Task-level `size`/`size_source`/`size_ruleset_version` did persist
structurally.

## Stash reconciliation

* **13 deferred-scope-expansion entries** reconciled under P-021 C6:
  unconditional duplicate scan (all CLEAN, recorded), and late PR identifiers
  recovered from Ship-owned closure records (PR #4 / #8 / #11). Review-thread
  IDs remain a truthful terminal `N/A` — those closure records state 0 PR
  review comments and 0 formal reviews.
* `037B1552` (Slack P3 credentials) — rewritten with a RETIRED / DO-NOT-IMPLEMENT
  banner, then **archived** so it cannot be picked up in triage.
* `A92E3FA0` — rewritten as the **governing product direction**, corrected from
  iOS/coder-acp to SPA/Copilot SDK; remains active (covers C1–C11).
* `4989A42D` — rewritten from "feature-complete Rust port" into a
  **migration-remediation tracker**; Rust-oracle authority removed.
* New `8C2D578D` — deferred `internal/apperr` taxonomy contamination (D6a).
* 14 unrelated quality follow-ups preserved untouched.

## Review

Six personas, three attempts. Attempt 1 FAIL (4 personas), attempt 2 FAIL (3),
attempt 3 all P0s closed. 51 remediations recorded in the plan's Post-Review
Remediation Record. Highest-value catches: `workspace.go`'s channel-routing API
owned by no unit; the `10 → 7` rule arithmetic; a gate that could never pass
(twice, at two different addresses); the D4 durability bootstrap hole; the
`seq` generation anchor; the permission dual-timeout-owner race.

## Open operator decisions carried forward

| # | Decision | Blocks |
|---|---|---|
| H1 | Flip gate to a required check (needs CODEOWNERS per `BEDD2E70`) | — |
| H2 / Q1 | Copilot CLI minimum version / protocol-3 floor | C2 |
| H3 / Q3 | Dev Tunnel authentication model | C10 |
| H4 / Q6 | Sessions per workspace process | C9 |
| H5 | Permission authorization model | C5 |
| H6 | D4a session-pointer store semantics | C4 |

## Next step

Ship claims **`004-S`**. Stage did not implement product code and did not
invoke Ship.

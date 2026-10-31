---
title: "Stage Session Memory — Remote Permission Safety, Transport Wire Integrity, and TUI Anti-Spoofing (Capability Set 1)"
description: "Session memory for Stage execution under DARK_MODE_ACTIVE (P-017) on Capability Set 1 (ACD578AE, D46C1B38, 7BDC38FF, 026CF750, 7EF4DA98, FE0D5ECD, 86BE6345, 21F9D21A, 6617141A)."
date: 2026-10-30
agent: stage
branch: "chore/stage-remote-permission-safety-transport-wire-integrit"
shipments:
  - "076-S"
  - "077-S"
features:
  - "083-F"
  - "084-F"
consumed_stash_ids:
  - "ACD578AE"
  - "D46C1B38"
  - "7BDC38FF"
  - "026CF750"
  - "7EF4DA98"
  - "FE0D5ECD"
  - "86BE6345"
  - "21F9D21A"
  - "6617141A"
---

## Session Summary

* **Mode**: `DARK_MODE_ACTIVE` (P-017); Stage Artifact Branch Gate (Step 1.9) passed on `chore/stage-remote-permission-safety-transport-wire-integrit` branched from `refs/remotes/origin/main` (`d5b209e80ee39baa520a85175ebda1f00a6e3f2f`, code baseline `075-S` `a2898942a12c70fe23be050d698ae4622153d97f`).
* **P-021 C5/C6 Triage Audit**:
  * **Unconditional Duplicate Scan (Trigger A)**: CLEAN across all active (`.backlogit/stash.jsonl`) and archived (`.backlogit/archive/stash.jsonl`) stash entries and `.backlogit/queue/`.
  * **Late-Identifier Reconciliation (Trigger B)**: Reconciled `ACD578AE` and `D46C1B38` (`PR #223`, `074-S`, merge commit `966f6a977ef6590e1d7a9155c3f64d0ee8861061`), `7BDC38FF` (`PR #212`, `071-S`, merge commit `67d3f5c03e893dde348ecdec08c6fd7b4c69a686`), `026CF750` (`PR #179`, `059-S`, merge commit `8c9f0f299fb2784ec0cbe03b843c84402927595d`), `7EF4DA98` and `FE0D5ECD` (`PR #181`, `060-S`, merge commit `24e381e0d6a7a81909707ec73a5954360fc805f1`), and `86BE6345`, `21F9D21A` (review thread `PRRT_kwDOTPuhps6wl2fF`), and `6617141A` (`PR #195`, `065-S`, merge commit `427bfbd5a3fcb22e117914c869a7faf2ab3d06c5`).
* **Artifacts Produced**:
  * Deliberation: `docs/decisions/2026-10-30-intercom-go-remote-permission-safety-transport-wire-and-tui-anti-spoofing-deliberation.md`
  * Implementation Plan (hardened + reviewed): `docs/plans/2026-10-30-intercom-go-remote-permission-safety-transport-wire-and-tui-anti-spoofing-plan.md` (`dispatch_mode: multi-agent`, `decision: PASS` across all 6 personas including Anchor `gpt-6.1-sol`)
* **Structured Sizing / Complexity Note**: Every task (`083.001-T`..`083.004-T`, `084.001-T`..`084.008-T`) carries structured `size: S` or `XS` (`size_source: agent`, `size_ruleset_version: 2h-rule-v1`) in frontmatter and prose-embedded `[Size: <S|XS> | Complexity: <low|medium>]` in the task description (as `.backlogit` task schema defines structured `size` without a separate `complexity` frontmatter field).

## Harvested Backlog Hierarchy & Shipments

1. **Shipment `076-S`** (Covering Feature `083-F`: *"Remote clipped-write approval refusal, shell/write summary invisible-rune and homoglyph anti-spoofing, and shared TUI modal header continuation helper (ACD578AE, D46C1B38, 7BDC38FF)"*, declared DAG root with `labels: ["dag-root"]` / `predecessor_source: "declared_root"` under explicit operator P-001 policy exception over `062-S`, code baseline `075-S` `a289894`):
   * `083.001-T` (`size: S`, `complexity: low`, `code`): Refuse remote approval of `PreviewClipped` write permissions (`authzInput.previewClipped = true` -> `ReasonDestructiveRemoteRefused` even when `remote_approve_destructive = true`) and clear/bound `PermissionDecision.Feedback` (`MaxDecisionFeedbackBytes = 1024`) in `internal/hub` (`ACD578AE`, `026CF750`).
   * `083.002-T` (`size: S`, `complexity: medium`, `code`, blocked by `083.001-T`): Harden `sanitizeBound`, `sanitizeIdent`, `sanitizeShellURL`, and `indentURLsFraming` against visually blank non-space non-`Cf` filler runes (`U+2800`, `U+3164`, `U+115F`, `U+1160`, `U+FFA0`, `U+034F`, `U+17B4`, `U+17B5`, `Co`, `Variation_Selector`) and `urls:` homoglyph/fullwidth lookalikes in `internal/agent/copilot` (`D46C1B38`).
   * `083.003-T` (`size: S`, `complexity: low`, `code`, blocked by `083.002-T`): Introduce shared `modalHeader` / `(Model).renderModalHeaders` continuation-indentation helper and column-0 header invariant in `internal/tui` (`7BDC38FF`).
   * `083.004-T` (`size: S`, `complexity: low`, `docs`, blocked by `083.003-T`): Update ADR MVP-3, backend architecture design doc (§5.2, §6.1), and `docs/config-reference.md` (`ACD578AE`, `D46C1B38`).
2. **Shipment `077-S`** (Covering Feature `084-F`: *"Transport wire/authz hardening, atomic hub-step credential revalidation, rejected-upgrade accounting, and pumpRejects rate-limiting (026CF750, 7EF4DA98, FE0D5ECD, 86BE6345, 21F9D21A, 6617141A)"*, blocked by `076-S`):
   * `084.001-T` (`size: S`, `complexity: medium`, `code`): Add `ConnectOptions.Validate` atomic hub-step credential revalidation at `admission`/`completeConnect` and `processAction` in `internal/hub` (`7EF4DA98`, `026CF750`).
   * `084.002-T` (`size: S`, `complexity: medium`, `code`, blocked by `084.001-T`): Enforce `Feedback` clearing/bounding and `local_only_action` refusals via `decodedAction` in `internal/transport`, and close connection on hub-step credential failure (`026CF750`, `7EF4DA98`).
   * `084.003-T` (`size: S`, `complexity: low`, `code`, blocked by `084.002-T`): Add non-blocking off-accept-loop over-cap `503` refusal queue (`refuseQ`) and `ServerConfig.OnReject` hook in `internal/transport` (`FE0D5ECD`, `86BE6345`).
   * `084.004-T` (`size: S`, `complexity: medium`, `code`, blocked by `084.003-T`): Wire `ConnectOptions.Validate` and `ServerConfig.OnReject`, validate pre-connect WebSocket handshake headers (`Sec-WebSocket-Version` / `Sec-WebSocket-Key`), count wrong-path `404` and library `websocket.Accept` refusals (`acceptResponseRecorder`), and exclude benign draining `503`s in `internal/transport` (`7EF4DA98`, `FE0D5ECD`, `86BE6345`, `21F9D21A`).
   * `084.005-T` (`size: XS`, `complexity: low`, `config`, blocked by `084.004-T`): Add `transport-layering-boundary`, `wsclient-standalone-boundary`, and `lower-layers-no-transport` depguard rules in `.golangci.yml` (`7EF4DA98`).
   * `084.006-T` (`size: S`, `complexity: low`, `code`, blocked by `084.005-T`): Reject URL fragments and opaque URLs in `wsclient.Dial` and add `config<->transport`, wire-constant drift, and layering conformance tests in `internal/transport/conformance_test.go` (`7EF4DA98`, `FE0D5ECD`).
   * `084.007-T` (`size: S`, `complexity: low`, `code`, blocked by `084.006-T`): Rate-limit `pumpRejects` forwarding of rejected-upgrade notices (`defaultRejectNoticeInterval = 250ms`) in `internal/app` (`6617141A`).
   * `084.008-T` (`size: S`, `complexity: low`, `docs`, blocked by `084.007-T`): Ratify `Host` default-port non-equivalence in ADR MVP-2 and update `docs/runbooks/intercom-operator-runbook.md` for rejected-upgrade accounting and rate-limiting (`FE0D5ECD`, `86BE6345`, `21F9D21A`, `6617141A`, `ACD578AE`).

## Consumed Stash Archival & Working-Tree Note

* Archived all 9 consumed stash entries via `backlogit stash archive`: `ACD578AE`, `D46C1B38`, `7BDC38FF`, `026CF750`, `7EF4DA98`, `FE0D5ECD`, `86BE6345`, `21F9D21A`, `6617141A`.
* The operator's uncommitted `.autoharness/config.yaml` model-routing edit remains preserved in `git stash` under message `"preserve-operator-config"` (`stash@{0}`) so Orchestrator Step 1.5 can publish and merge the clean Stage PR before popping it.

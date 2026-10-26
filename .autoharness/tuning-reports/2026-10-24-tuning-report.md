---
title: "Harness Tuning Report — 2026-10-24"
description: "Autoharness 1.6.0 tuning report for intercom-go: schema contract upgrades, model routing alignment, overlay refresh, and local divergence preservation"
date: 2026-10-24T19:45:00Z
workspace: "C:\\Source\\GitHub\\intercom-go"
autoharness_version: "1.6.0"
autoharness_home: "C:\\Python\\Python315\\Lib\\site-packages\\autoharness\\data"
scope: "all"
status: "applied-local-uncommitted"
---

## Drift Summary

* Breaking changes: 0
* Degrading changes: 33 (resolved: 31; advisory/deferred: 2)
* Growth opportunities: 6 (applied: 5; advisory: 1)
* Cosmetic adjustments: 1 (applied: 1)

## Composition

* Installed primary stack pack: `api-service`
* Current primary stack pack: `api-service`
* Installed stack packs: `["api-service", "cli-tool"]`
* Current stack packs: `["api-service", "cli-tool"]`
* Installed layers: `["foundation", "instructions", "workflow", "review", "runtime", "backlog", "knowledge", "overlays"]`
* Recommended layers: `["foundation", "instructions", "workflow", "review", "runtime", "backlog", "knowledge", "overlays"]`
* Capability packs: `["agent-engram", "graphtor-docs", "backlogit", "browser-verification", "continuous-learning", "strict-safety", "release-observability", "adversarial-review"]`

## Checksum Scan

* Pre-tune manifest-tracked artifacts: 106 (`unchanged`: 78, `user-modified`: 28, `missing`: 0, `ignored`: 0)
* Post-tune manifest-tracked artifacts: 106 (`unchanged`: 106, `user-modified`: 0, `missing`: 0, `ignored`: 0)
* Backups written before modification: `.autoharness/backups/2026-10-24/` (58 artifacts backed up)

## Learning Signals

* Compound patterns detected: 6 (`cross_cutting_tags`: `backlogit` [9], `go` [9], `gatecheck` [6], `git` [4], `shipment-reconcile` [3]; `category_concentration`: `build-errors` [7/28 entries])
* Promotion-ready instincts: 1 (`closure-frontmatter-seed`, already promoted in `mode: propose` at `.github/instructions/learned-closure-frontmatter-seed.instructions.md` and reinforced by 1.6.0 `operational-closure/SKILL.md` required frontmatter keys and `autoharness gate closure-evidence`)
* Recurring closure findings: 0 unresolved (legacy date-prefixed closure naming and missing frontmatter keys addressed by `028-F` and 1.6.0 `operational-closure/SKILL.md`)
* Policy-gap candidates detected: 0 (existing `P-015`, `P-016`, `P-018`, `P-020`, `P-021`, and in-repo `gatecheck` gates cover the compound patterns)
* Learning-driven proposals generated: 1 (`TUNE-009`)

## Proposed and Applied Changes (Ordered by Priority)

### P1 — Degrading

1. **[TUNE-001] Upgrade `schema_contracts` across `.autoharness/` (`source: schema-contract`, Status: Applied)**
   * `upgrade-config-contract-1.1.0-to-1.2.0` / `1.3.0`: Upgraded `.autoharness/config.yaml` `schema_version` from `1.1.0` to `1.3.0`.
   * `upgrade-profile-contract-1.0.0-to-1.1.0` & `remove-retired-capability-pack`: Upgraded `.autoharness/workspace-profile.yaml` `schema_version` from `1.0.0` to `1.1.0` and removed the retired `agent_intercom` block.
   * `upgrade-manifest-contract-1.0.0-to-1.1.0`: Upgraded `.autoharness/harness-manifest.yaml` `schema_version` from `1.0.0` to `1.1.0` and `autoharness_version` from `1.5.0` to `1.6.0`.

2. **[TUNE-002] Restore required P-013.4/P-013.5 model-routing frontmatter on pipeline agents (`source: schema-contract`, Status: Applied)**
   * Resolved 9 `frontmatter-conformity` (`FM_MISSING_REQUIRED`) migration proposals and failing `orchestrator_model_routing_fields`, `stage_model_routing_fields`, and `ship_model_routing_fields` checks on `.github/agents/_orchestrator.agent.md`, `.github/agents/_stage.agent.md`, and `.github/agents/_ship.agent.md` using the authoritative routes from `.autoharness/config.yaml`.
   * Preserved all in-repo `LOCAL DIVERGENCE` sections in `_orchestrator.agent.md` (Step 1.5 actor contract), `_stage.agent.md` (Role Boundary and Step 1.9 branch gate), and `_ship.agent.md` (Step 2/3/4.1 P-002 per-task gate and Step 6 closure contract).

3. **[TUNE-003] Refresh stale model-routing frontmatter and 1.6.0 content across all 16 subagents (`source: schema-contract`, Status: Applied)**
   * Resolved 16 `model-routing` (`stale-route`) migration proposals across `.github/agents/subagents/*.agent.md`:
     * Tier 1 personas (`architecture-strategist`, `concurrency-reviewer`, `constitution-reviewer`, `correctness-reviewer`, `go-reviewer`, `learnings-researcher`, `maintainability-reviewer`, `prompt-builder`, `schema-cli-docs-coupling-reviewer`, `scope-boundary-auditor`, `template-integrity-reviewer`) -> `claude-haiku-5.5` / `anthropic` / `xhigh` / `long_context`.
     * Tier 2 subagents (`go-engineer`, `security-lens-reviewer`, `security-reviewer`) -> `claude-sonnet-5.5` / `anthropic` / `high` / `long_context`.
     * Tier 3 subagents (`adversarial-review`, `security-sentinel`) -> `gemini-4-argon` / `google` / `high` / `long_context` (plus `anchor_review_family: "gpt-6.1-sol"` on `adversarial-review`).

4. **[TUNE-004] Resolve `ESCALATION_DEGRADED` same-route no-op for Stage (`source: structural-drift`, Status: Applied)**
   * Added nested `model_routing.stage.escalation` (`gpt-6.1-sol` / `openai` / `high` / `long_context`) in `.autoharness/config.yaml` and updated `.github/agents/_stage.agent.md` and `.github/agents/_ship.agent.md` Escalation Protocol sections so `escalation_route_resolution` passes (`ok: true`).

5. **[TUNE-005] Update `.autoharness/backlog-registry.yaml` CLI hook surface (`source: structural-drift`, Status: Applied)**
   * Synced `.autoharness/backlog-registry.yaml` with the autoharness 1.6.0 `backlogit.registry.yaml` template, adding `cli_command` entries for `poll_hook_events`, `ack_hook_events`, `append_comment`, and `save_memory` (resolving `backlog_registry_cli_surface`).

6. **[TUNE-006] Add P-017 Dark Factory Visibility Protocol table to `.github/policies/workflow-policies.md` (`source: structural-drift`, Status: Applied)**
   * Replaced the legacy `agent-intercom` dark-mode visibility paragraph in P-017 with the 1.6.0 `Dark Factory Visibility Protocol` table and `Degraded visibility` rule (resolving `dark_factory_visibility_contract`) while preserving all in-repo P-010, P-015, and P-016 contract clauses enforced by `tools/gatecheck/internal/stagecontract/` and `tests/integration/`.

7. **[TUNE-007] Resolve portability findings in `.github/agents/_orchestrator.agent.md` (`source: structural-drift`, Status: Applied)**
   * Replaced the tilde path `~/.autoharness/` in the Session-Start Dynamic Reload section with `$HOME/.autoharness/` and aligned the Model Routing section with `.autoharness/config.yaml`, resolving both P1 portability findings (`hardcoded_user_home`, `hardcoded_ah_home`).

8. **[TUNE-008]Startup-script contract `refresh-startup-script-ps1-contract` (`source: schema-contract`, Status: Deferred / Preserved Local Divergence)**
   * `verify-workspace` classifies `start.ps1` as `known-legacy` because it defines `function Invoke-EngramCommandWithProgress`. In `intercom-go`, `start.ps1` carries an intentional, test-enforced `LOCAL DIVERGENCE` (`011.018-T`, resolves `9F9A3CB2`: bounded shared wall-clock budget for Engram pre-warm and `Push-Location`/`Pop-Location` location restoration), verified by `tests/integration/start_script_bounded_test.go`, `start_script_hermetic_test.go`, and `start_script_test.go`. Overwriting `start.ps1` with the generic template would break those integration tests, so `start.ps1` is preserved and its manifest checksum refreshed.

9. **[TUNE-010] Legacy `.backlogit/` storage root (`source: structural-drift`, Status: Advisory / Operator Review)**
   * The workspace uses the legacy `.backlogit/` directory (`.backlog/` is absent). Per `tune-harness` Step 1.5 (detect-and-report only), review `backlogit migrate --workspace-dir --dry-run` when ready to migrate to `.backlog/`. Tune did not modify or rename `.backlogit/`.

### P2 — Growth

1. **[TUNE-009] Refresh foundation, instruction, skill, and file-lock script artifacts to autoharness 1.6.0 (`source: structural-drift` & `learning-driven`, Status: Applied)**
   * Re-rendered `AGENTS.md`, `.github/copilot-instructions.md`, 12 `.github/instructions/*.instructions.md` files, 15 `.github/skills/*/SKILL.md` files, `.github/skills/file-lock/scripts/*`, `.github/prompts/feature-flow-dark.prompt.md`, `start.sh`, and `scripts/deploy-harness.{ps1,sh}` from autoharness 1.6.0 templates:
     * Updated Go version and architecture metadata (`Go 1.25.0`, toolchain `go1.26.9`, `cmd/{intercom,intercom-ctl}`, `internal/{agent,app,apperr,config,hub,pathsafe,transport,tui}`, `tools/gatecheck/`), removing stale `Go 1.21+` and retired `acp` references.
     * Removed retired `agent-intercom` capability pack sections across foundation, instruction, and skill files.
     * Upgraded `concurrency.instructions.md`, `file-lock/SKILL.md`, and `acquire_lock`/`release_lock` scripts with capability-token ownership verification (`LOCK_TOKEN` / `--token` / `--force` / `--workspace-root`).
     * Upgraded `github-pr-automation.instructions.md` with §1.6.1 Threadless Review-Body Findings (`UNDISPOSITIONED_BODY_FINDINGS`) and `circuit-breaker.instructions.md` with `--detach`/`--poll` exit-9 supervision rules.
     * Updated `operational-closure/SKILL.md` with required frontmatter keys (`closure_status`, `compaction_status`) while preserving the `post-merge` Output line enforced by `TestOperationalClosurePostMergeFilenameConformance`.
     * Preserved `.github/skills/harness-architect/SKILL.md`, `.github/skills/shipment-reconcile/SKILL.md`, and `.github/workflows/ci.yml` local divergences enforced by repository conformance tests.
   * Note on `copilot-code-review.instructions.md`: Confirmed intact (`applyTo: '**'`, `excludeAgent: 'cloud-agent'`, FOCUS framing). Reminder: GitHub Copilot code review reads instruction files from the pull request's **base branch** (`main`), so any instruction changes take effect for Copilot code review only after the tuning PR is merged to `main` (and Settings -> Copilot -> Code review custom instructions is enabled as an advisory prerequisite).

## Verification Results

* `autoharness verify-workspace`: **PASS** (exit code 0; `strict_schema_blockers: 0`, `blockers: 0`, `warnings: 0`, `unresolved: 0`, all 52 `targeted_checks` `ok: true`, 106/106 checksums `unchanged`, 1 intentionally deferred `migration_proposal` for `start.ps1` documented in `TUNE-008`).
* Repository conformance & gatecheck suite (`go test ./tests/integration/... ./tools/gatecheck/...`): **PASS** (exit code 0).

## Adversarial Verification

### Reviewer Dispatch

* `TOOL_OK: anchor-review-model` (`openai` / `gpt-6.1-sol`, `reasoning_effort: high`)
* Anchor Reviewer (`model_routing.anchor_review` — `gpt-6.1-sol`): 1 finding (P2/MINOR advisory)
* Reviewer A (Template Fidelity — `gemini-4-argon`, `high`, `long_context`): 0 findings
* Reviewer B (Overlay Coherence — `claude-sonnet-5.5`, `high`, `long_context`): 5 findings (5 MINOR)
* Reviewer C (Cross-Reference Integrity — `claude-haiku-5.5`, `xhigh`, `long_context`): 9 findings (3 MAJOR downgraded to MEDIUM/advisory after template verification, 1 MINOR promoted to HIGH and auto-fixed, 5 MINOR advisory)

### Consensus Findings (Ordered by Confidence x Severity)

#### HIGH Confidence

| # | Severity | Domain | File | Issue | Status |
|---|---|---|---|---|---|
| 1 | MINOR | cross-reference | `.github/agents/subagents/go-engineer.agent.md`, `.github/instructions/github-pr-automation.instructions.md` | `FORMAT_CHECK_COMMAND` was empty (`''`) in `variables_used`, rendering an empty backtick span (````) in quality-gate lists | AUTO-FIXED (`FORMAT_CHECK_COMMAND: 'gofmt -l .'`) |
| 2 | MINOR | overlay-coherence | `.autoharness/config.yaml` | Comment above `model_routing.alt_doc_review` contained stale provider assumptions (`tier1 is OpenAI`) | AUTO-FIXED |
| 3 | MINOR | overlay-coherence | `.autoharness/workspace-profile.yaml` | `ci.pipeline_steps` and `ci.quality_gates` were empty lists (`[]`) despite `.github/workflows/ci.yml` defining `changes`, `lint`, `test`, `build`, `topology-check`, and `ci-gate` | AUTO-FIXED |

#### MEDIUM Confidence (Template-Level / Advisory)

| # | Severity | Domain | File | Issue | Status |
|---|---|---|---|---|---|
| 4 |MINOR | cross-reference | `.github/prompts/feature-flow-dark.prompt.md`, `.github/policies/workflow-policies.md` | `feature-flow-dark.prompt.md` and `P-017` match `autoharness 1.6.0` canonical templates verbatim; `P-018` section and `github-pr-automation.instructions.md` §1.9.4/§1.9.6 explicitly enforce `COPILOT_REVIEW_BLOCK` in dark mode | NEEDS REVIEW (upstream template suggestion) |
| 5 | MINOR | cross-reference | `AGENTS.md`, `.github/instructions/constitution.instructions.md` | Stop Conditions quick-reference table summarizes `Review-fix cycles per task` as `Accept remaining as backlog, commit, move on` while pointing to `circuit-breaker.instructions.md` (which states P-021 C1/C3 rules in full) | NEEDS REVIEW (upstream template suggestion) |
| 6 | MINOR | overlay-coherence | `.github/instructions/concurrency.instructions.md` | Lock script path is written as `scripts/acquire_lock.{ps1,sh}` in the upstream template while manifest installs lock scripts under `.github/skills/file-lock/scripts/` | NEEDS REVIEW (upstream template suggestion) |

#### LOW Confidence (Advisory)

* `.github/instructions/constitution.instructions.md`: Capability Overlays section in upstream template lists `agent-engram`, `backlogit`, and `graphtor-docs` while `AGENTS.md` and `copilot-instructions.md` list all 8 enabled packs.
* `.github/instructions/adversarial-review.instructions.md` & `.github/agents/subagents/adversarial-review.agent.md`: Upstream template renders empty inline backticks when `alt_review` is unset.

### Auto-Remediation Summary

* Findings auto-fixed: 3
* Findings requiring manual review (upstream template suggestions): 3
* Low-confidence / template-verbatim advisories: 2

### Verification Result: PASS WITH WARNINGS

## Branch Safety & Next Steps

* Because the workspace was on the default branch (`main`) with pre-existing uncommitted work in `internal/agent/copilot/`, all tuning updates were applied as local uncommitted changes with full backups under `.autoharness/backups/2026-10-24/`.
* **Recommended Git workflow**: Never commit or push tune output directly to `main`. Stash or commit the separate `internal/agent/copilot/` work first, switch to a dedicated feature branch (for example `git switch -c chore/autoharness-tune-2026-10-24`), review the diff, commit the tuned harness artifacts, and open a pull request.

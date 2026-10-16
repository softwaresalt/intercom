---
title: "Residual batch triage: Go gatecheck hardening, scripts/CI/Windows fixes, closure doc accuracy, upstream routing"
description: "Stage deliberation (P-021 C6) over 28 stash entries in three release units, plus routing of eight upstream-rendered entries and one archived refuted entry"
topic: "Residual stash cleanup: clusters A+B (Go gatecheck), D (scripts/CI/Windows), C-local (closure docs and tests), C-upstream (rendered harness text)"
depth: "standard"
decision_status: "decided"
promoted_to: "plan"
linked_artifacts:
  - "docs/plans/2026-10-15-intercom-go-residual-batch-plan.md"
  - "docs/decisions/2026-10-15-intercom-go-retiredarch-sibling-sideeffect-threat-model-deliberation.md"
  - "docs/decisions/2026-10-15-intercom-go-closure-reconcile-text-alignment-deliberation.md"
  - "docs/decisions/2026-10-15-intercom-go-ship-dark-mode-copilot-approval-text-deliberation.md"
  - "docs/decisions/2026-10-11-intercom-go-gatecheck-write-path-batch-2-deliberation.md"
  - "docs/decisions/2026-10-10-intercom-go-gatecheck-batch-b1-write-path-input-hardening-deliberation.md"
tags:
  - "gatecheck"
  - "p-021"
  - "upstream-render"
  - "stage"
---

# Deliberation: residual batch triage

## Session context

* **Agent and mode:** Stage, Orchestrator-invoked, sequential, not dark. Routing directive
  P-013.5: `config.model_routing.stage` = `gemini-4-argon` / google / high / long_context. The
  runtime cannot select that model: **`ROUTING_DEGRADED`** (ran on the invoking model).
* **Operator selection:** the Orchestrator relayed the grouping (clusters A+B as one Go-gatecheck
  shipment; cluster D as a second; cluster C triaged). Per precedent (batch-b1, batch-2) the
  relayed instruction stands in for operator confirmation, and any override can be made at
  staging-PR review before Ship claims a shipment. Flagged in the handback.
* **Base and branch:** `chore/stage-gatecheck-go-hardening-and-test-hygiene` from
  `origin/main` @ `ddbe0d0`. Step 1.9: single worktree, branch absent, clean tree at creation
  (the operator's pre-existing edits to `.autoharness/config.yaml` and `.backlogit/stash.jsonl`
  were preserved with `git stash push` and restored after the session; nothing discarded).
* **Planning overlap:** shipment 046-S (feature 056-F) is queued and untouched. These shipments
  are chained behind it.

## Step 1: Triage and P-021 obligations

**Correction to the Orchestrator's premise.** The relayed eligibility said "requires
deliberation: no" for clusters A, B and most of D. P-021 C6 does not key on that field: every
entry carrying the `DEFERRED SCOPE EXPANSION` marker is forced to the deliberate route regardless
of shape, size or apparent triviality. Almost every staged entry carries the marker, so this
artifact is the C6 deliberation for the whole batch. (The two entries 8E18CCF5 and 56B16321 also
carry `requires_deliberation=true` themselves; D-RB-6 settles them.) Entries without the marker:
none of the staged ones except 96B308D1 (bug) and 5CC5532E (follow-up), which are routed here as
well because they join marked entries.

### (A) Unconditional duplicate scan

Method: token-overlap (Jaccard) scan of every staged entry against all 130 active entries, plus
keyword checks for the three deliberated entries.

* **Verdict: CLEAN for all staged entries.** Highest overlaps were pairs that are the same
  surface but different expansions: 4A3851F0 vs A85D25F7 (0.63; symlink-skip vs scanPath
  output), 815C2B28 vs BEFE6DF2 (0.45), 3750C37C vs 0ECC1895 (0.48; both held by D-RA-7),
  8E1E2D80 vs 724FA2FB (0.32; related approval wording, carried together upstream).
* No merge, no duplicate archival required.

### (B) Late-identifier reconciliation

* **Retrieval source:** Ship-owned closure records in `docs/closure/` that cite each entry ID.
* **Recovered carrier shipments (not capture origins):** 032-S/035-F (F5958BBC, 48EA04C1);
  041-S/051-F (F5958BBC, 48EA04C1, 4A3851F0, A85D25F7, 4995F8C3, C8827920, F9D52027, 06CE25E5);
  045-S/055-F (815C2B28, BEFE6DF2, EF11062A, 63BCA7C6); 042-S/052-F (65AE88CA, 96B308D1);
  036-S/046-F (96B308D1); 034-S/044-F (8E18CCF5, 56B16321); 025-S/028-F (B4D38D63);
  043-S/053-F (5CC5532E, 86381564, 762A4FCA, BAB25951, 9C6C2078, F2759510, 8E1E2D80).
* **PR and review-thread identifiers:** no late identifier found. Every staged entry's
  `N/A` for PR and review thread was a pre-PR local-review finding; the `N/A` stands as a truthful
  terminal record. The recorded task, feature and shipment IDs in each entry remain accurate.
* The write half (a stash text edit) is not applied: no identifier was recovered, so there is
  nothing to write.

## Step 1.5: Grouping

* **Unit 1 (A+B, 18 entries -> 24 tasks):** all Go code or Go tests in `tools/gatecheck` and its wrapper
  script. Coherent: shared package family and test harness; low blast radius. Includes an
  explicit sub-epic split (unignore, retiredarch, wrapper, writepath).
* **Unit 2 (D, 6 entries -> 9 tasks):** small test, script, CI-doc and Windows fixes; separate
  artifact class from unit 1 as the Orchestrator directed.
* **Unit 3 (C-local, 4 entries -> 5 tasks):** design-doc, local rendered-deviation revert (A744E584), and integration-test accuracy for the 043-S
  closure contract. Kept separate from unit 2 (different domain: closure contract surface).

## Decisions

* **D-RB-1 (three release units, chained).** Shipments in order: 046-S (existing) -> unit 1 (047-S) ->
  unit 2 (048-S) -> unit 3 (049-S), by `blocks` edges. Each successor stays queued. Rationale: Ship's one-shipment-
  at-a-time model; unit 3 touches `tests/integration` that 046-S (056-F) also edits.
* **D-RB-2 (per-entry disposition).** All 28 staged entries are consumed into tasks; see the plan
  units U1..U18b, V1..V5b, W0..W3. Authorized scope for these units is the plan's acceptance
  criteria; the P-021 C1 deferral reason no longer applies once the plan is reviewed.
* **D-RB-3 (ordering from the stash).** F9D52027 after 4A3851F0 and C8827920 (U4 after U5, U3).
  The wrapper units are sequenced U8 -> U9 -> U10 -> U10b -> U11 because they edit the wrapper test files and U11 also
  edits the script alongside U9's shared library helper.
* **D-RB-4 (4995F8C3 split).** The retiredarch half is split across U7 (`retiredarch.go`, `retiredarch_test.go`), U7b (`select_test.go`, `tomlprimary.go`), and U7c (`tomlprimary_test.go`, `pin_test.go`) to keep each task at <=2 files; the writepath half is folded into U13b (`057.013-T`)
  so one task edits `writepath_test.go` and `writepath.go` naming once. Comment-only AC prevents code drift.
* **D-RB-5 (96B308D1 + 5CC5532E item 1).** One root cause (Windows temp path loses backslashes
  into `/bin/bash`), split into V1 (`runner_test.go`, covering all five `TestGatecheckCleanup_*` tests at lines 612-719) and V1b (root-package wrapper tests). 5CC5532E item 2 (start-script timing) is already covered by
  046-S/056-F and is excluded. The "commit untracked docs/memory/2026-09-30/..." clause of
  96B308D1 is dropped: that file does not exist in this workspace.
* **D-RB-6 (the two requires_deliberation entries).** 8E18CCF5: add the `::notice::` emission for
  parity, split into V3 (`writepath.go`, `writepath_test.go`) and V3b (`writepath_golden.json`, `tools/gatecheck/main_test.go`). Plan review corrected the premise: the write-path wrapper has no mode dispatch and
  the retired-architecture notice already moved into the Go engine, so V3/V3b is a Go change in
  `internal/writepath` and root CLI dispatch test, not a script edit. 56B16321: **document `--self-test` as manual-only, no new CI step** (V4), because
  a new CI job is a config change with ongoing runtime cost for a mode whose sub-parts
  (`--self-test-integrity`, repo scan) CI already exercises (confirmed via `ci.yml` lint steps and `go test` `TestRun_SelfTest_MatchesGolden`). V4's AC makes the implementer verify
  that claim and halt if a sub-mode is uncovered. Operator may override toward a smoke step.
* **D-RB-7 (upstream-rendered entries vs local deviation A744E584).** AGENTS.md, the Ship agent,
  and the skills are "Generated by autoharness". However, git history (commits `2749d5d` and `b5abe5f`) proves that **A744E584 is NOT an upstream defect**: the upstream template used the space-separated form (`--message <value>` / `--author <value>`), and Ship hand-edited the `=` form (`--message=<value>` and `--author=<value>`) locally into `.github/agents/_ship.agent.md` (~line 828) and `.github/skills/shipment-reconcile/SKILL.md` (~line 636). Because `autoharness shipment cascade-close` rejects the `=` form with exit 2 (proven in `docs/closure/044-S-054-F-post-merge-closure.md`), leaving it breaks every future mutating closure. Therefore A744E584 is staged locally in 049-S as W0 (`059.004-T`, reverting those two files to the upstream space-separated form; a future `autoharness install`/`tune` re-render would also restore it) and W0b (`059.005-T`, adding a contract-test guard in `ship_feature_completion_contract_test.go`). The remaining rendered-file entries — 9B7E94E9, D5D103A1, 0606E72F,
  92708A06 and 8134B7B7 (and the deliberated F2759510, 8E1E2D80, related 724FA2FB) — describe genuine upstream template defects. Route: one consolidated upstream-issue stash entry (40BCCD69, amended to remove A744E584) listing each entry, the file, and the requested template change; the originals stay
  active (they remain true of this workspace until re-render), and nothing is hand edited for them.
  Note 92708A06 (AGENTS.md dead links) and 8134B7B7 items 1-6 need the autoharness template or
  `autoharness install`/`tune`, not local edits.
* **D-RB-8 (9C6C2078 archived as refuted premise).** It asked to change the design-doc example to the `--message=value`
  form. Verification confirmed that `docs/design-docs/2026-10-11-closure-contract-surface-matrix.md` line 283 already shows the **space form**
  (`--message <text> --author <name>`), which is the exact form `autoharness shipment cascade-close` accepts (while the `=value` form was a local Ship deviation that the CLI rejects with exit 2). Applying 9C6C2078 would make the design doc wrong. Archived via `backlogit stash archive 9C6C2078` as a refuted premise. Entries
  86381564 and A744E584 were noted as "after 9C6C2078": both now proceed in 049-S (W0/W0b and W1).
* **D-RB-9 (housekeeping).** Archived as resolved/completed via `stash archive` (non-destructive):
  0B6CCE5A (go1.26.9 toolchain bump, resolved by PR #117 / 041-S closure PR #116), C13CF9BA (the
  Stage archival follow-up for it), 3D161831 and 51117F2D (045-S compaction follow-ups completed
  by PR #138, merge ddbe0d0).
* **Not staged, by Orchestrator instruction:** A92E3FA0, 0A215E0B, B32FCDD1, DB12DA37, 8ED0A706,
  475E76D2, 6B751D8B, 585FCD6D, FC6FE1D5, the blocked features, and the 83 deliberate-first
  entries other than the three deliberated this session.

## Open questions for the operator

1. Confirm D-RB-6 (manual-only for `--self-test`) or ask for a CI smoke step.
2. File the upstream autoharness issue (D-RB-7, stash 40BCCD69); Stage cannot open issues.

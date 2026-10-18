---
title: "Stage batch B/C/D deliberation: test/script hygiene, closure docs/tests, closure-record compaction"
date: 2026-10-18
stash_ids: [B32FCDD1, 80E3F89C, FB759F5D, 17326572, 1D5A783D, FC6FE1D5, 8ED0A706]
status: proposed (operator ratification of the scope splits D-BCD-2, D-BCD-3, D-BCD-5 requested)
---

## Frame

Orchestrator handoff (main at c778db8; 052-S shipped). Bounded scope: stash group B
(B32FCDD1, 80E3F89C, FB759F5D, 17326572), group C (1D5A783D, FC6FE1D5) and group D
(8ED0A706). Every group-B/C entry except FC6FE1D5 and 8ED0A706 carries the
`DEFERRED SCOPE EXPANSION` marker, so P-021 C6 forces this deliberation before planning.
Source refs carried: B32FCDD1 (feature 056-F, shipment 046-S), 80E3F89C (057-F, 047-S;
tasks 057.009/013/019-T), FB759F5D and 17326572 (058-F, 048-S; tasks 058.003-T, 058.006-T),
1D5A783D (059-F, 049-S; tasks 059.001/004-T), FC6FE1D5 (053-F, 043-S; task 053.013-T),
8ED0A706 (052-F, 042-S). All PR numbers and review-thread IDs were recorded N/A.

## Intake obligations (P-021 C5/C6)

* Duplicate scan (unconditional): CLEAN. No other active stash entry describes the same
  expansion for any of the seven entries (keyword scan of the active stash).
* Late-identifier reconciliation: the `N/A` PR/thread fields triggered a search of the
  Ship-owned closure and compacted-memory records that cite each entry
  (046-S, 047-S, 048-S, 049-S). Outcome: no late identifier found; the recorded `N/A`
  stands for all five deferred entries. No identifier was overwritten.

## Reconciliation of the FC6FE1D5 premise (verified against main c778db8)

* FC6FE1D5 item (1) (retired replica tests `shipment_close_path_behavior_test.go` and
  `shipment_close_path_harness_test.go` await deletion): STALE. Both files are absent from
  `tests/integration/`; they were deleted by commit 775e21f (053.013-T). The design-doc matrix
  rows 49 and 50 already say "deleted". 1D5A783D's flag is correct. Item (1) is closed by
  evidence; no work.
* FC6FE1D5 item (2) (real-binary smoke `TestCascadeCloseClassifyOnlySmoke` skips when
  autoharness is absent; CI `test` job may not install it): CONFIRMED LIVE. In
  `.github/workflows/ci.yml` the only autoharness install step is in job `topology-check`
  (line 661); job `test` (`expensive`) has none. So the smoke is not CI evidence today.
  `ci.yml` is autoharness-generated, so installing autoharness in `test` is an upstream/
  re-render matter and is out of scope here (do not rewrite rendered files).

## Decisions

* D-BCD-1 (group split, domains): two shipments, as directed. Shipment B holds Go-test and
  shell-script hygiene (tests and one script file, no production code). Shipment C holds
  closure documentation and one closure test. Within each, one skill domain per task.
* D-BCD-2 (partial harvest of multi-part entries, ratification requested):
  * 80E3F89C: harvest only (a) shim self-check path spelling and (c) junction target
    normalization (both test-only). Items (b) CI log injection escape, (d) symlink
    `.gitignore` baseline read and (e) `gitRunnerEnv` scrub dedup change production gatecheck
    behavior or gate verdicts and need their own deliberation; they stay on the active
    entry, which is annotated, not archived.
  * B32FCDD1: harvest the `WaitDelay` containment and hermetic env/scratch for
    `runStartScriptSnippet`. The optional `gh auth token` / ambient `GITHUB_*` scrub of
    `Resolve-GitHubTokens` is deferred (low risk: the snippet runner never runs
    `Invoke-StartScriptMain`); it stays on the annotated entry.
  * 17326572: harvest the shellcheck items (confirm SC1091 directive already present at the
    current line, SC2034, SC2329) and make the WSL-bash skip visible. The CRLF working-copy
    inconsistency of `tools/gatecheck/internal/writepath/testdata/writepath_golden.json`
    (index LF, working copy CRLF, no attribute) needs a `.gitattributes` rule; that file
    carries an autoharness-managed header, so the rule is an operator decision (D-BCD-6) and
    stays on the annotated entry. The toolchain-pin decision doc is already LF (`docs/decisions/**
    text eol=lf`), so that sub-item is closed by evidence.
  * 1D5A783D: harvest the stale line-number refresh in the design-doc matrix only. The
    059.004-T rationale parenthetical lives in rendered agent-facing files
    (`_ship.agent.md`, `shipment-reconcile/SKILL.md`) and must not be hand-edited; it is
    added to the upstream/re-render family of 40BCCD69 (operator to confirm; no upstream
    filing performed by Stage). Entry stays active, annotated.
* D-BCD-3 (8ED0A706 is a shipment task, not an in-Stage operational run): the compaction
  moves 37 Ship-owned `docs/closure/` records to `docs/archive/closure/`. `docs/closure/` is
  not among the Stage artifact roots enforced by the Step 1.9 ownership discriminator, the
  closure records are Ship-owned, and a bulk move needs PR-reviewed evidence. It is also a
  single docs domain and mechanical. Candidate set is frozen to the 37 `*.md` records whose
  last commit date is on or before 2026-09-20 (matches the 37 records 042-S counted;
  `.gitkeep`, `evidence/` and everything newer are excluded). Compaction is archive-not-
  delete (compact-context behavioral constraint). It becomes a third shipment (D) chained
  after C, so one PR carries one concern.
* D-BCD-4 (sizing): every task is `size` XS or S and `complexity` trivial, low or medium;
  nothing exceeds 2 hours.
* D-BCD-5 (chain): B blocks-depends on 052-S (already shipped), C on B, D on C.
  Rationale: each shipment touches different files, but the pipeline-topology gate requires
  a linear chain; D last because it moves the closure records that C's doc may cite.
* D-BCD-6 (operator decisions needed): see the plan "Open decisions" section.

## Bookkeeping decided in this session (no harvest)

* 003-DL / 3750C37C: operator decision DEFER CODEOWNERS review on `tools/gatecheck/**`;
  annotated, comment appended to 003-DL.
* 40BCCD69 and dependents (F2759510, 8E1E2D80, 724FA2FB, 9B7E94E9, D5D103A1, 0606E72F,
  92708A06, 8134B7B7): annotated filed-upstream, awaiting re-render.
* 060.001-T: acceptance-criterion mutation example corrected (`len(path) < 130` is
  verdict-neutral and does not fail the test; the failing unsound mutations are
  `len(path) < 2*longPathThreshold` and `len([]rune(path)) < longPathThreshold`). Done by a
  minimal edit of the archived record with an inline CORRECTION marker plus an append-only
  `backlogit comment` carrying the original wording, and the matching line of the 060 plan.

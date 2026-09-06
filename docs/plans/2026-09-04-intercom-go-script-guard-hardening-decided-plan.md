---
title: "Decided Plan — Build and scan script hardening (008-F)"
date: 2026-09-04
shipped: 2026-09-06
feature: 008-F
shipment: 007-S
pr: 22
source_deliberation: "docs/decisions/2026-09-04-intercom-go-residual-hardening-triage-deliberation.md"
resolves_stash: [35D76D5E, 78D13775, B6203CCA]
archived_full_plan: docs/archive/plans/2026-09-04-intercom-go-script-guard-hardening-plan.md
---

# Decided Plan — Build and scan script hardening (008-F)

## Final scope

Two deferred findings, script logic only (no Go source changes):

1. `35D76D5E` -- `scripts/build.ps1`'s I5 repo-root-escape guard is lexical
   only (blind to symlinks/junctions).
2. `78D13775` -- `check-retired-architecture.sh`'s TOML comment masking is
   line-oriented (blind to multi-line strings).
3. `B6203CCA` -- wire `--self-test` into CI (relocated from 007-F; must land
   after 1/2 so the test-first fixtures can be committed failing first).

## Units that survived review (dependency order)

* **U2 → U1**: symlink/junction escape tests authored first, then
  `build.ps1`'s guard extracted into `scripts/lib/OutputPathGuard.ps1` and
  made symlink/junction-aware. **Ordinal (case-sensitive) comparison
  preserved unchanged** -- an earlier draft's AC-3 required a Windows
  case-fold; review correctly identified that as loosening a containment
  control and rejected it. Windows junctions covered alongside symlinks
  (cheaper, unprivileged bypass). TOCTOU between validation and write
  recorded as documented residual risk, not claimed closed.
* **U4 → U3**: multi-line TOML fixtures + a verdict manifest authored
  first, then the masking rewritten to a real `tomllib` parse (ground-truth
  corrected: the masking code is Python in a heredoc, not bash, contra an
  earlier draft), failing closed on parse errors. A hand-written lexer
  fallback (for pre-3.11 Python) is retained defensively, capped to the
  same acceptance criteria. `--self-test` discovers fixtures by
  glob + manifest so U4 adds files only.
* **U5**: `--self-test` wired into CI as a second, blocking step, landing
  last (after U3/U4 are green).

## Key constraints (protected invariants)

* I5 may only get **stricter**, never looser.
* The retired-architecture scan must never report clean on genuinely
  contaminated config (R2 -- the highest-severity risk in this plan).
* Existing single-line masking and non-symlink negative test behavior must
  keep passing.

## Rejected alternatives / corrected drafts

* A Windows case-fold for the I5 comparison (loosens containment -- rejected).
* A hand-rolled bash state machine for TOML masking (ground truth: it's
  Python, not bash; a hand-written lexer would also need to handle escaped
  closes, quote runs, and nested delimiters -- not reliably a small unit).
* A single positive-control fixture (insufficient to prove fail-closed
  behavior -- expanded to include an escaped-delimiter close and a
  malformed-file case).
* U2 ordered after U1 (would have driven the full ~5-minute build matrix
  and violated test-first -- reversed to U2 → U1 with U1 extracting a
  directly testable function).

## Post-ship additions (found during PR review, not in the original plan)

* Fallback TOML lexer's own EOF fail-closed logic was untested dead code in
  CI (tomllib always wins on Python >=3.11); `--self-test` now runs both
  engines against every fixture explicitly.
* Guard's ancestor walk used `Test-Path`, which can miss a dangling
  reparse point; switched to `Get-Item -Force`, narrowed to only treat
  `ItemNotFoundException` as "does not exist."
* Embedded Python needed `from __future__ import annotations` for PEP 585
  generic-subscript annotations to remain compatible with the fallback's
  own pre-3.9 target.
* Test helper `createDirectoryJunction` needed the same context-timeout
  bound already used elsewhere in the file.
* One pre-existing edge case (`repoRootWithSep` double-separator when the
  repo root is itself a drive/UNC root) found and deliberately deferred
  (P-021 C1 -- out of scope, fails safe, cannot manifest for this repo).

## Full history

`docs/archive/plans/2026-09-04-intercom-go-script-guard-hardening-plan.md`
(includes the full problem frame, constitution check, per-unit acceptance
criteria, the pre-remediation review findings table, risks R1-R4, and the
original plan-hardening record).

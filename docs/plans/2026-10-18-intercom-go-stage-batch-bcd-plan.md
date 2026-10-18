---
title: "Stage batch B/C/D: test/script hygiene, closure docs/tests, closure-record compaction: implementation plan"
date: 2026-10-18
source: docs/decisions/2026-10-18-intercom-go-stage-batch-bcd-deliberation.md
stash_ids: [B32FCDD1, 80E3F89C, FB759F5D, 17326572, 1D5A783D, FC6FE1D5, 8ED0A706]
---

## Summary

Three chained shipments. B: test/script hygiene (Go tests plus one shell script; no
production code). C: closure documentation and one closure test. D: P-020 closure-record
compaction (docs only, archive not delete). Chain: B blocks-depends on 052-S, C on B, D on C.
No rendered harness file is touched in any unit. Ship executes; Stage does not.

Requires plan hardening: no

Plan Hardening Signals (all absent): no public API/schema/contract change (test and docs
changes only; D moves records, archive not delete); no security/auth behavior; no migration
or destructive action (D uses `git mv`, nothing is deleted, fully reversible by moving back);
no external integration or operator checkpoint inside the units; low rollout/rollback risk
(each unit is a single revertable commit). Residual risk tracked in the Risks section.

## Source of truth

Deliberation `docs/decisions/2026-10-18-intercom-go-stage-batch-bcd-deliberation.md`
(D-BCD-1 to D-BCD-6). Verification of claims against main c778db8 is recorded there.

## Shipment B: Test and script hygiene follow-ups (feature "Test and script hygiene follow-ups")

### B1: Bound runStartScriptSnippet like the launcher tests (tests domain; size S, complexity low)
* File: `tests/integration/start_script_test.go` (1 file, 1 function). Route the snippet
  runner's `cmd` through the existing `startScriptCombinedOutputBounded` helper
  (`start_script_bounded_test.go`) instead of `cmd.CombinedOutput()`, keeping the existing
  context timeout and `t.Fatalf` behavior; keep the already-isolated `cmd.Dir`/env treatment
  and keep the per-call `cmd.Dir = t.TempDir()` scratch directory and the filtered hermetic env already present (no new isolation is added).
* Out of scope: scrubbing `gh auth token`/ambient `GITHUB_*` for `Resolve-GitHubTokens`.
* Acceptance: the snippet runner uses the shared bounded helper (no direct
  `CombinedOutput`); existing snippet tests still pass; `start_script_bounded_test.go`
  helper tests still pass. Posture: characterization-first.

### B2: Tighten closure-template `slug` token class (tests domain; size S, complexity low)
* File: `tests/integration/operational_closure_post_merge_filename_test.go`
  (`closureOutputPlaceholderClass`). Replace `strings.Contains(token, "slug")` by an exact
  `slug` or `_slug` suffix match, mirroring the id-token rule. Add a table case proving
  `slugger` and `width`-style tokens are rejected as unrecognized while `slug` and
  `post_merge_slug` still map.
* Acceptance: unrecognized look-alike tokens return the "unrecognized placeholder" error; the
  real template `{YYYY-MM-DD}-{slug}` still compiles; the conformance test still passes.
  Posture: test-first.

### B3: Compare shim path spellings by identity, not string (tests domain; size S, complexity medium)
* File: `tools/gatecheck/wrapper_harness_test.go` (the `command -v go` shim self-check around
  lines 118-125). Compare the located shim with the expected one via normalization that
  treats `/tmp/x` and `/c/...` as the same directory (shell `-ef` or `cygpath -u` applied to
  both sides), instead of `EqualFold` on strings.
* Acceptance: the self-check passes when Git Bash spells the same directory two ways, and
  still fails when `go` resolves to a different file. Evidence: wrapper tests pass; Ship
  records one full `go test ./tools/gatecheck/...` run on this host (flake was observed once).
  Posture: characterization-first.

### B4: Normalize UNC and `\\?\` junction targets in test helper (tests domain; size S, complexity medium)
* File: `tools/gatecheck/internal/writepath/writepath_junction_windows_test.go`
  (`nativeJunction`/`setMountPoint`, windows build tag). Normalize UNC (`\\server\share`)
  and `\\?\`-prefixed targets before building the `\??\` substitute name and print name.
* Acceptance: a junction to a `\\?\C:\...` target and (when constructible) a UNC target
  resolves to the same substitute-name form; existing junction cases unchanged.
  Posture: characterization-first.

### B5: Resolve shellcheck warnings in check-write-path-precondition.sh (config/script domain; size XS, complexity low)
* File: `scripts/check-write-path-precondition.sh`. Confirm the `# shellcheck source=`
  directive for `lib/gatecheck-run.sh` is present (it is on main; SC1091 is then already
  clean or needs `-x`), and resolve SC2034 (`ROOT` assigned for its side effect) and SC2329
  (`cleanup` is invoked only via `trap`) with justified `# shellcheck disable=` directives (if `shellcheck` is unavailable on the host, Ship records a manual review of the three codes instead) or
  a code-neutral restructure. No behavior change.
* Acceptance: `shellcheck -x scripts/check-write-path-precondition.sh` reports none of
  SC1091/SC2034/SC2329 (Ship runs it if available; otherwise records the manual review);
  wrapper tests unchanged. Posture: characterization-first.

### B6: Make WSL-bash wrapper-test skips visible (tests domain; size S, complexity medium)
* Files: `tools/gatecheck/check_write_path_wrapper_test.go` and the shared bash-resolution
  helper in `tools/gatecheck/wrapper_harness_test.go` (2 files). Where tests skip because
  the resolved bash is the WSL launcher (`C:\Windows\System32\bash.exe`), emit an explicit
  `t.Skipf` message naming the cause and the fix (use Git Bash on PATH) instead of a
  silent skip.
* Acceptance: running under a WSL-first PATH reports a SKIP line with the reason; with Git
  Bash first the tests run unchanged. Posture: test-first.

Dependency edges inside B: B3 blocks B6 (same file). Others independent.

## Shipment C: Closure docs and test follow-ups (feature "Closure docs and test follow-ups")

### C1: Refresh stale line references in the closure-contract matrix (docs domain; size S, complexity low)
* File: `docs/design-docs/2026-10-11-closure-contract-surface-matrix.md` (1 file). Refresh the
  `tests/integration/ship_feature_completion_contract_test.go` line refs (L138, L229, L237,
  L241, L268, L283, L287, L303, L319, L359 and any others) to the current file, or replace
  them with stable row/test-name anchors. Add a short dated note that rows 49 and 50
  (retired replica tests) are verified deleted (commit 775e21f) and that FC6FE1D5 item (1) is
  closed by that evidence.
* Acceptance: every cited line number points at the described content in the current test
  file (spot-verifiable by grep), or is replaced by a name anchor; no rendered file edited.
  Posture: doc-only.

### C2: Make smoke-test skip state explicit as non-CI-evidence (tests domain; size XS, complexity low)
* File: `tests/integration/cascade_close_classify_only_smoke_test.go`
  (`cascadeSmokeRequireTools`). Make the autoharness-absent skip message state that the CI
  `test` job does not install autoharness (it is installed only in `topology-check`), so a
  skip must not be read as CI evidence. Do not edit `ci.yml` (rendered).
* Acceptance: skip message names the cause and the caveat; test still skips cleanly when the
  binary is absent and runs unchanged when present. Posture: test-first.

Dependency edges inside C: none. Out of scope (kept on annotated 1D5A783D): the
059.004-T parenthetical in rendered agent text, and installing autoharness in the CI `test`
job; both are upstream/re-render items.

## Shipment D: P-020 closure-record compaction (feature "Compact old closure records")

Candidate set (frozen): the 37 `docs/closure/*.md` records whose last commit date is on or
before 2026-09-20; `.gitkeep`, `evidence/` and newer records are excluded. Each unit runs the
`compact-context` closure-compaction phase for its band: write one compacted summary
`docs/closure/{date}-closure-compacted-{band}-closure-summary.md` listing every release unit
in the band (what was verified, signals, follow-ups, plus the original paths), and `git mv`
the originals to `docs/archive/closure/`. Never delete. These are move-only units; the
"fewer than 3 files" figure is waived for the mechanical moves and each unit authors exactly
one new file.

* D1: band 2026-09-03..2026-09-07 (14 records) (docs domain; size S, complexity low)
* D2: band 2026-09-08..2026-09-10 (13 records) (docs domain; size S, complexity low)
* D3: band 2026-09-11..2026-09-20 (10 records) (docs domain; size S, complexity low)
* Candidate cutoff uses the LAST COMMIT date (`git log -1 --format=%cs`), not the filename date; e.g. `2026-09-11-016-s-017-...` (committed 2026-09-10) is in, `2026-09-21-027-s-030-...` (committed 2026-09-27) is out. The authoritative per-unit path lists are in the appendix; Ship must not re-derive them.
* 22 of the 37 are `*-post-merge-closure.md` records. Archiving them shrinks the universe of `TestOperationalClosurePostMergeFilenameConformance` (reduced coverage is accepted: those records are retired, and the new summary files end in `-closure-summary.md` so they fall outside the glob).
* The summary per unit is an index plus one-line per-record findings (time-box: 2 hours); it does not re-narrate each record.
* Acceptance (each): originals present under `docs/archive/closure/` byte-identical; the
  summary lists every moved path; `docs/closure/` retains no moved record; no non-candidate
  file moved. D3 additionally: total archived closure records equals 37 and
  `TestOperationalClosurePostMergeFilenameConformance` still passes (it only globs what
  remains).
* Dependency edges: D2 blocks-depends on D1, D3 on D2 (shared summary naming and count).

## Dependency graph

052-S (shipped) <- Shipment B <- Shipment C <- Shipment D (blocks edges on the shipments).
Within B: B3 before B6. Within D: D1, D2, D3 sequential.

## Decisions and Rationale

See the deliberation: partial harvests (D-BCD-2), 8ED0A706 as a shipment (D-BCD-3), chain
(D-BCD-5). Size/complexity are separate axes and set per task.

## Risks and Caveats

* B3/B4/B6 touch Windows/Git-Bash behavior; Ship must verify on this host. If a unit proves
  unreliable, split or de-risk rather than force.
* D moves records that other docs may cite by path; Ship greps for inbound references to the
  37 paths and records them in the compacted summary instead of rewriting history.
* Compaction state flags (`compaction_status: degraded` in 042-S closure) stay as historical
  record; Ship may add a closure note when D lands.

## Open decisions for the operator

1. Ratify the partial harvests of B32FCDD1, 80E3F89C, 17326572, 1D5A783D (residuals stay on
   annotated active entries).
2. `.gitattributes` LF rule for `tools/gatecheck/internal/writepath/testdata/writepath_golden.json`
   (autoharness-managed header on the file): approve a local one-line rule or route upstream.
3. Stale example comment `len(path) < 130` at `internal/pathsafe/reparse_windows_test.go:215`
   (source, outside this scope): approve a one-line comment fix task.
4. Add the 059.004-T agent-text parenthetical to the upstream family of 40BCCD69.
5. Confirm D is acceptable as a third shipment rather than an operational run.

## Runtime Verification and Closure

No runtime surfaces change. Verification is by the units' acceptance criteria plus Ship's
standard gates; B3/B4/B6 need a local Windows full-package `go test` evidence line.

## Appendix: frozen closure-record candidate lists (shipment D)

Paths are under `docs/closure/`. Source: last-commit date <= 2026-09-20, `*.md` only (37 records).

### D1
* `001-S-001-F-post-merge-closure.md` (2026-09-03)
* `002-S-002-F-post-merge-closure.md` (2026-09-03)
* `003-S-003-F-post-merge-closure.md` (2026-09-03)
* `004-S-005-F-post-merge-closure.md` (2026-09-04)
* `005-S-006-F-post-merge-closure.md` (2026-09-04)
* `2026-09-04-007-f-ci-supply-chain-hardening-pass2-adversarial-review.md` (2026-09-04)
* `006-S-007-F-post-merge-closure.md` (2026-09-05)
* `007-S-008-F-post-merge-closure.md` (2026-09-05)
* `008-S-009-F-post-merge-closure.md` (2026-09-06)
* `009-S-010-F-post-merge-closure.md` (2026-09-06)
* `2026-09-06-010-s-011-f-residual-hardening-round2-adversarial-review.md` (2026-09-06)
* `010-S-011-F-post-merge-closure.md` (2026-09-07)
* `011-S-012-F-post-merge-closure.md` (2026-09-07)
* `2026-09-07-011-s-012-f-pathsafe-containment-adversarial-review.md` (2026-09-07)
### D2
* `012-S-013-F-post-merge-closure.md` (2026-09-08)
* `013-S-014-F-post-merge-closure.md` (2026-09-08)
* `014-S-015-F-post-merge-closure.md` (2026-09-08)
* `2026-09-08-012-s-retire-d6a-gate-narrowing-adversarial-review.md` (2026-09-08)
* `2026-09-08-013-s-014-f-pathsafe-reparse-containment-adversarial-review.md` (2026-09-08)
* `2026-09-08-013-s-014-f-pathsafe-reparse-containment-runtime-verification.md` (2026-09-08)
* `2026-09-08-014-s-015-f-retired-arch-gate-hardening-runtime-verification.md` (2026-09-08)
* `2026-09-08-retired-arch-gate-hardening-adversarial-review.md` (2026-09-08)
* `015-S-016-F-post-merge-closure.md` (2026-09-10)
* `016-S-017-F-post-merge-closure.md` (2026-09-10)
* `2026-09-10-015-s-016-f-pathsafe-windows-canonicalization-symmetry-adversarial-review.md` (2026-09-10)
* `2026-09-10-015-s-016-f-pathsafe-windows-canonicalization-symmetry-runtime-verification.md` (2026-09-10)
* `2026-09-11-016-s-017-f-pathsafe-error-surface-parity-runtime-verification.md` (2026-09-10)
### D3
* `021-S-022-F-post-merge-closure.md` (2026-09-16)
* `2026-09-16-021-s-022-f-ship-covering-feature-completion-runtime-verification.md` (2026-09-16)
* `017-S-018-F-post-merge-closure.md` (2026-09-17)
* `024-S-027-F-post-merge-closure.md` (2026-09-18)
* `2026-09-19-024-s-027-f-staging-artifact-handoff-actor-runtime-verification.md` (2026-09-18)
* `2026-09-19-repair-closure-reconcile-skill-contract-plan-unit-1-adversarial-review.md` (2026-09-19)
* `025-S-028-F-post-merge-closure.md` (2026-09-20)
* `026-S-029-F-post-merge-closure.md` (2026-09-20)
* `027-S-030-F-post-merge-closure.md` (2026-09-20)
* `2026-09-20-026-s-029-f-repair-ship-agent-execution-contract-runtime-verification.md` (2026-09-20)

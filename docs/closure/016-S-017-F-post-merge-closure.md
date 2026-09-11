---
title: "Post-merge closure: 016-S / 017-F — pathsafe error-surface parity and Windows long-path precision"
description: "Operational closure artifact for shipment 016-S / feature 017-F"
status: "complete"
tags:
  - "closure"
  - "016-S"
  - "017-F"
  - "post-merge"
date: 2026-09-11
mode: post-merge
shipment: 016-S
feature: 017-F
pr: 51
merge_commit_sha: 75cfe3b1b954c49d149f531464ccd97f1770dd0b
compaction_status: done
closure_status: READY
releasability: READY
conditions: []
---

# Post-merge closure: 016-S / 017-F — pathsafe error-surface parity and Windows long-path precision

- Mode: `post-merge`
- PR: #51 (`fix(pathsafe): error-surface parity and Windows long-path precision (016-S)`)
- Merge commit: `75cfe3b1b954c49d149f531464ccd97f1770dd0b` (merge-commit strategy, P-009)
- Date: 2026-09-11
- Compaction status (P-020): `done` — `compact-context` invoked (target:
  all) during this closure session; bounded Tier-1 consolidation of this
  release unit's own memory
  (`docs/archive/memory/2026-09-10-stage-016s-pathsafe-error-surface-parity.md`
  + `docs/archive/memory/2026-09-10-ship-016-s-pre-pr-checkpoint.md`) into
  `docs/memory/compacted/2026-09-11-016-s-017-f-pathsafe-error-surface-parity-compacted.md`,
  verbose originals moved to `docs/archive/memory/`. Other top-level
  `docs/memory/` files are all dated 2026-09-04–2026-09-08 (3–7 days old),
  below the 14-day compaction threshold, so they were out of scope for this
  bounded invocation.

## Summary of the change

Brought `internal/pathsafe`'s error surface to parity between `NewRoot` and
`checkSymlinkEscape`, and raised the precision of the Windows long-path
(MAX_PATH) threshold, without changing any containment verdict and without
adding any filesystem write primitive. Three tasks, one feature:

- `017.001-T`: `checkSymlinkEscape`'s `canonicalizeReparse` error branch now
  routed through the existing `wrapPathError` helper, so
  `errors.As(err, &*fs.PathError)` matches on Windows exactly as it already
  does for `NewRoot`'s own failure branch. No-op on POSIX (`wrapPathError`
  is a passthrough there). New privilege-independent (directory-junction
  based, not symlink-based) Windows test locks the `*fs.PathError` surface.
- `017.002-T`: characterization test + doc-comment correction for
  `addLongPathPrefix`'s `\.\` device-namespace exclusion branch. Corrects a
  stale "unreachable" premise carried in stash entry `6B751D8B` — measured
  this cycle: `filepath.IsAbs` reports `true` for every `\.\` form tested,
  so the branch **is** reachable, currently redundant with the
  immediately-following generic `\\` guard, and is kept as documented
  defense-in-depth (D-4) rather than deleted or replaced with a
  panic/unreachability assertion (that option was explicitly considered and
  rejected — the "unreachable" premise it would encode is false).
- `017.003-T` (depends on `017.002-T`): `addLongPathPrefix`'s MAX_PATH
  threshold comparison is now UTF-16 code-unit aware (`utf16Len`) instead of
  using UTF-8 byte length as a proxy. Precision improvement only — the old
  proxy could only ever over-prefix (false positive), never under-prefix
  (false negative), so this is explicitly not a correctness or security fix
  (framing it as one is a documented anti-goal, D-5).

`BF5DE670` (deferred write-path mitigation tracker) was preserved untouched
per the operator's binding scope contract and Stage's D-6 decision —
carried forward as feature-level acceptance criteria AC-F1–AC-F5 rather than
silently broadened into speculative engineering against a write API that
does not exist. `6B751D8B` was partially harvested (items 1–2 above); items
3–4 (buffer pooling, CI lint scoping) remain deferred and the entry stays
active.

## CI status and unresolved review items

- All CI checks green at the PR HEAD before merge (`090cbf56fcc916309913430f6a50f0bea57ae0d2`):
  13 checks pass (cross-compile ×4, lint, security, test, test (windows,
  advisory), pipeline-topology, ci gate, gitignore regression,
  detect-changes, load-cross-compile-targets). The merge commit itself is
  `75cfe3b1b954c49d149f531464ccd97f1770dd0b` — CI ran against the PR HEAD
  prior to the (unmodified-tree) merge commit, per standard GitHub merge
  semantics.
- **Local review (pre-PR)**: three independent persona passes (general
  code-review, security-review, Go-idiom review) — all **PASS,
  merge-ready**, zero P0/P1/P2. Two P3s remediated directly (invalid-UTF-8
  test case; clarifying comment on the Lstat non-ENOENT branch); one P3
  (byte-length fast-path micro-optimization for `utf16Len`, re-proposing a
  plan-rejected gold-plating option) captured as P-021 deferred-scope stash
  entry `475E76D2` rather than fixed.
- **Copilot review (P-018 gate): `SATISFIED`.** Requested via the
  documented `[bot]`-suffixed REST fallback
  (`docs/compound/2026-09-06-requesting-copilot-review-non-collaborator-repo.md`)
  immediately after PR creation, per the lesson captured in
  `docs/compound/2026-09-10-p018-copilot-engagement-must-precede-merge.md`
  (this shipment correctly followed that lesson — no repeat of the 015-S
  process gap). Copilot completed 2 review rounds:
  - Round 1 (commit `24ed6d3`): posted 2 real review threads plus a
    "Suppressed comments" section in the review body naming 2 *different*
    findings that had no corresponding thread (see new compound entry
    below). Both real threads were addressed in commit `090cbf5`: reframed
    `TestAddLongPathPrefixDeviceNamespaceBranchReachability`'s doc comment
    to remove an "enforces retention" overclaim and extended every test
    case past `longPathThreshold` so each one genuinely exercises the guard
    chain instead of short-circuiting at the threshold check; added YAML
    frontmatter to the new memory checkpoint doc matching sibling
    convention. Both suppressed (threadless) findings were also fixed
    proactively as valid engineering concerns (clarified the PR body's
    AC-F5 scope wording; the test-threshold gap was the same fix already
    applied above). Replied to both real threads citing the fixing commit
    `090cbf5`, then resolved both via `gh api graphql`
    (`resolveReviewThread`).
  - Round 2 (commit `090cbf5`): Copilot re-reviewed, opened zero new
    threads. `autoharness gate copilot-review 51 --enforcement required`
    returned `SATISFIED` with `unresolved_thread_ids: []`.
  - Last-mile re-check immediately before merge: re-ran unconditionally,
    still `SATISFIED`, HEAD unchanged since the round-2 pass.
- No unresolved P0/P1 findings at merge. One P-021 deferred-scope entry
  (`475E76D2`) and the two long-standing advisory/deferred stash entries
  (`BF5DE670`, `6B751D8B`) remain active per AC-F4/D-6.
- Mechanical scope check: production/test edits confined to
  `internal/pathsafe/pathsafe.go`, `internal/pathsafe/reparse_windows.go`,
  `internal/pathsafe/reparse_windows_test.go`, and the new
  `internal/pathsafe/symlink_windows_test.go` (AC-F5). `root.go` untouched
  (AC-F2, zero diff). Independent grep for filesystem write primitives
  across all non-test `.go` files returns zero matches (AC-F3).

## Runtime verification report

See
`docs/closure/2026-09-11-016-s-017-f-pathsafe-error-surface-parity-runtime-verification.md`.

**Verdict: `READY`** — `internal/pathsafe` has no runtime entrypoint of its
own. **Correction (Copilot review, PR #52):** neither `cmd/intercom` nor
`cmd/intercom-ctl` currently calls `config.Load`/`Validate` at process
startup (`cmd/intercom`'s `RunE` returns `errNotImplemented` immediately;
verified by repo-wide search for non-test callers of `config.Load` /
`(*Config).Validate` — none exist outside `internal/config` itself). The
substantive evidence is the Go test suite (`internal/pathsafe` and
`internal/config`, which directly exercise the changed
`checkSymlinkEscape` / `addLongPathPrefix` code). CLI smoke (`--help`)
confirms basic process-startup health only and is not evidence for the
changed paths; see the runtime-verification report for full detail.

## Invariants to preserve

- `checkSymlinkEscape`'s `canonicalizeReparse` error branch must continue
  routing through `wrapPathError` — reverting to a bare/unwrapped error
  return reopens the exact error-surface asymmetry with `NewRoot` this
  shipment closed.
- The `addLongPathPrefix` `\.\` device-namespace guard must not be deleted
  or replaced with a panic/unreachability assertion — it is reachable (not
  unreachable) and currently redundant with the following bare `\\` guard;
  deletion is a documented-safe no-op today but removes self-documenting
  intent and couples correctness to the adjacent guard's continued
  existence (D-4).
- The MAX_PATH threshold comparison must remain UTF-16-code-unit-aware
  (`utf16Len`), not a UTF-8 byte-length proxy — reverting reintroduces
  (harmless but imprecise) over-prefixing for non-ASCII paths near the
  threshold.
- `BF5DE670`'s deferred write-path mitigation remains gated on
  `scripts/check-write-path-precondition.sh` firing — do not implement
  speculatively against a write API that does not yet exist.
- `TestAddLongPathPrefixDeviceNamespaceBranchReachability`'s scope is
  deliberately limited to the corrected observable premise
  (`filepath.IsAbs` verdict + unchanged output) — it does **not** enforce
  retention of the `\.\` branch as a mutation-tested invariant; do not
  reintroduce that overclaim in future doc-comment edits (Copilot review
  finding, round 1).

## Pre-deploy audits

Not applicable — internal library correctness/precision-hardening change
with no runtime service surface, migration, flag, or config-schema change
of its own. As corrected above, no `cmd/` entrypoint currently calls
`config.Load`/`Validate` at startup, so the changed code is not exercised
on any current process-startup path; the `internal/pathsafe` and
`internal/config` test suites are the evidence, already covered by CI.

## Deployment / rollout path

**Merge-only, immediately active.** The change lands in `main`. It has no
currently-wired runtime entrypoint of its own — it takes effect only once
some future caller wires `cmd/intercom` or `cmd/intercom-ctl` to
`internal/config.Load`/`Validate` (not yet implemented) — so no phased
rollout, canary, or separate release event applies to this internal
library fix today.

## Post-deploy checks

No additional manual post-deploy verification step is required beyond the
standard CI green-check already gating merges repository-wide, and the
existing `internal/pathsafe` / `internal/config` test suites which directly
exercise the changed code paths on every future CI run.

## Risky action record

No destructive or irreversible action was taken during implementation or
review. No process deviation occurred this cycle: Copilot review was
actively requested before merge (correctly following the lesson from
`docs/compound/2026-09-10-p018-copilot-engagement-must-precede-merge.md`),
awaited to completion for both HEAD advances, and both real review threads
were replied-to-and-resolved before merge.

The shipment safe-close used the **P-015 verified fully-covered-root
cascade path** (`backlogit shipment ship`), independently classified before
invocation: `017-F` is a root feature (no `parent_id`) whose only
descendants — confirmed via a direct `parent_id` scan across
`.backlogit/queue/` + `.backlogit/archive/` — are exactly its 3 manifest
tasks (`017.001-T`, `017.002-T`, `017.003-T`), with no grandchildren; the
manifest contains nothing beyond the feature and its descendants. Fully
verified post-invocation per `shipment-reconcile`'s Cascade Close
Sub-Procedure:
- `returned_ids`: `[]` (empty — no classifier/engine mismatch).
- `archived_ids` == `allowed_ids` == `required_ids` == `{016-S, 017-F,
  017.001-T, 017.002-T, 017.003-T}` (two-set gate satisfied: no unexpected
  artifact archived, nothing required left unarchived).
- `parent_id: 017-F` preserved on all 3 archived tasks (unchanged from the
  pre-close snapshot).
- Post-mode: all 5 expected archive files present;
  `git status -- .backlogit/archive/` showed no deletions (P-007 guard
  clean).
- AC-F4 independently re-verified post-cascade: stash entries `BF5DE670`
  and `6B751D8B` remain active/untouched (neither is part of the shipment
  manifest; the cascade op never touches stash artifacts).

## Healthy signals

- `go build ./...`, `go vet ./...`, `gofmt -l .`, and `go test ./...`
  remain green on the post-merge closure branch (built on merge commit
  `75cfe3b1`) — re-verified during this closure session.
- `go test ./internal/pathsafe/... -race` was clean pre-merge (no code
  changed between that run and the merge commit, so not independently
  re-run post-merge).
- CLI smoke probes (`go run ./cmd/intercom --help`, `go run
  ./cmd/intercom-ctl --help`) both exit 0 with usage printed, no panic.
- CI's full 13-check suite passed cleanly on PR #51's final CI run.
- Copilot review reached `SATISFIED` with zero unresolved threads before
  merge — no gap this cycle.

## Failure signals

- A future regression that removes `wrapPathError` routing from
  `checkSymlinkEscape`'s error branch, reopening the error-surface
  asymmetry with `NewRoot`.
- A future doc-comment edit to
  `TestAddLongPathPrefixDeviceNamespaceBranchReachability` that
  reintroduces the "enforces branch retention" overclaim Copilot correctly
  flagged this cycle.
- A future change reverting the MAX_PATH threshold check from
  `utf16Len(path)` back to `len(path)`.

## Monitoring plan

The durable monitoring for this shipment is the existing CI pipeline and
test suite: `internal/pathsafe`'s test suite directly exercises the changed
error-wrapping and threshold-comparison code paths on every future PR.

## Rollback trigger / procedure

Not applicable in the operational sense (no live service to roll back). If
a future regression is found, the remedy is a standard follow-up PR fixing
or reverting the specific change, gated through the same review/CI process.

## Validation window

Not applicable — no live deployment window; the change is exercised on
every future config-load/CI run starting immediately after this merge.

## Owner

Repository maintainer (`softwaresalt`) for the merged change and the
shipment closure.

## Releasability evidence

**READY.** No conditions. The review chain (3-persona local review, zero
P0/P1/P2; Copilot review `SATISFIED` with both real threads resolved) and
runtime verification (`READY`) are both clean with no open caveats. One
P-021 deferred-scope stash entry (`475E76D2`) and two long-standing
deferred/advisory entries (`BF5DE670`, `6B751D8B`) remain active by design,
not as blocking conditions.

## Post-merge re-verification (on the post-merge closure branch, built on `main` after merge)

Re-ran the full quality-gate sequence on the post-merge closure branch
(`post-merge/017-f-pathsafe-error-surface-parity`, built on `main` @
`75cfe3b1`): `go build ./...`, `go vet ./...`, `gofmt -l .` — all clean;
`go test ./...` — all green; CLI smoke (`--help` on both entrypoints) —
exit 0, no panic. Confirms the merge itself introduced no regression.

## Stash follow-up items

None new. `475E76D2` (P-021 deferred-scope, captured pre-PR) is the only
new stash entry from this shipment's execution; it already carries the
full six-field payload and requires no additional stashing action here.

## Source artifact cleanup

Checked `custom_fields.source_stash_id` and
`custom_fields.source_deliberation_id` on the shipped top-level item
(`017-F`) before archival — neither field is present (`017-F`'s only
`custom_fields` entry is `harness_status: pending`), consistent with the
pattern observed on the two immediately preceding shipments (014-S, 015-S).
`7ADAC481` (cited in `017-F`'s own description as "harvested, archived")
was independently checked via `backlogit stash get 7ADAC481` — confirmed
already absent (`not found`), consistent with Stage's harvest workflow
consuming/removing fully-harvested stash entries rather than retaining them
as separately archivable artifacts. `BF5DE670` and `6B751D8B` were
independently confirmed still present, active, and untouched in
`.backlogit/stash.jsonl` (full text re-read during this closure session) —
no stash or deliberation archival action was applicable or taken beyond
what the shipment cascade-close itself already performed on manifest items.

## Feed Back Into the Harness

- `docs/compound/2026-09-06-requesting-copilot-review-non-collaborator-repo.md`
  reviewed for staleness and found fully accurate — the `[bot]`-suffixed
  REST fallback worked exactly as documented, twice (rounds 1 and 2), on
  this shipment's PR. Classification: **keep**.
- `docs/compound/2026-09-10-p018-copilot-engagement-must-precede-merge.md`
  reviewed for staleness and found fully accurate — this shipment
  correctly followed its prescribed sequencing (request Copilot review
  immediately after PR creation, before any merge attempt), and the gate
  reached `SATISFIED` cleanly. Classification: **keep**. No repeat of the
  015-S process gap.
- `docs/compound/2026-05-07-backlogit-shipment-status-constraints.md`
  reviewed for staleness and found fully accurate — this session's cascade
  close (`backlogit shipment ship`) transitioned the shipment record
  through `active -> shipped -> archived` exactly as documented; no direct
  `backlogit move --status shipped` was attempted. Classification: **keep**.
- New compound entry created for this shipment's own discovery:
  `docs/compound/2026-09-11-copilot-suppressed-comments-vs-review-threads.md`
  ("Copilot review body 'Suppressed comments' can differ entirely from the
  actual posted review threads — always fetch reviewThreads via GraphQL").
  This is a genuinely new, previously-undocumented lesson: the two
  "suppressed comments" named in Copilot's round-1 review body
  (`.backlogit/queue/016-S.md:12` scope wording;
  `reparse_windows_test.go:120` short-case-threshold gap) were entirely
  different findings from the two real, resolvable review threads
  (`reparse_windows_test.go:103` branch-retention overclaim; missing memory
  checkpoint frontmatter) — both sets were addressed, but only the real
  threads had a reply/resolve obligation.
- No other `docs/compound/` entries reference this shipment's surface;
  none required update, consolidation, or replacement.

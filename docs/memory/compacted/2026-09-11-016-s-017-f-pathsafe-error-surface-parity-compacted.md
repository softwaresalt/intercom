---
title: "Compacted session memory: 016-S / 017-F — pathsafe error-surface parity and Windows long-path precision"
date: 2026-09-11
shipment: 016-S
feature: 017-F
status: shipped
pr: 51
merge_commit_sha: 75cfe3b1b954c49d149f531464ccd97f1770dd0b
compacted_from:
  - docs/archive/memory/2026-09-10-stage-016s-pathsafe-error-surface-parity.md
  - docs/archive/memory/2026-09-10-ship-016-s-pre-pr-checkpoint.md
  - docs/archive/memory/2026-09-11-ship-016-s-session-end.md
---

# Compacted session memory: 016-S / 017-F

Consolidates the Stage planning session and the Ship execution session for
shipment 016-S / feature 017-F into a single durable summary. Verbose
originals archived to `docs/archive/memory/`.

## Outcome

Shipped and closed. PR #51, merge commit `75cfe3b1b954c49d149f531464ccd97f1770dd0b`.
Shipment archived via the P-015 verified fully-covered-root cascade path.
Full detail: `docs/closure/016-S-017-F-post-merge-closure.md`.

## Scope

Bounded stash scope: `BF5DE670`, `6B751D8B`, `7ADAC481`. Three tasks under
one feature (`017-F`): `017.001-T` (wrapPathError parity, high),
`017.002-T` (device-namespace branch reachability, medium), `017.003-T`
(UTF-16-aware MAX_PATH threshold, medium, depends on `017.002-T`).

## Key decisions and corrections (from Stage planning)

- `7ADAC481` re-prioritized low → high (015-S closure showed it was P1 in
  isolation, deferred only on P-021 C1 scope grounds).
- `BF5DE670` contributes no task this cycle — trigger re-measured, has not
  fired (write-path gate exits 0, zero write primitives). Preserved as
  feature-level acceptance criteria AC-F1–AC-F5 rather than archived or
  broadened into speculative engineering.
- `6B751D8B` items (1)–(2) harvested into 017.002-T/017.003-T; items (3)–(4)
  (buffer pooling, CI lint scoping) deferred as pure perf/CI, outside
  operator's named scope — entry stays active.
- **Material premise correction**: `6B751D8B` claimed
  `addLongPathPrefix`'s `\.\` device-namespace branch is "unreachable" —
  FALSE. Measured: `filepath.IsAbs` returns true for every `\.\` form
  tested; the branch is reachable but currently redundant with the
  following bare `\\` guard. A panic/unreachability assertion was
  considered and rejected (the premise it would encode is false).
- UTF-16 threshold change is a precision improvement, not a security fix —
  UTF-8 byte count is always ≥ true UTF-16 code-unit count, so the old
  proxy could only over-prefix, never under-prefix. Framing as a
  vulnerability fix was an explicit anti-goal (D-5).
- Plan review: 5 personas/models, PASS on attempt 1, 2 P1 + 6 P2 + 7 P3, zero
  P0, all remediated pre-harvest.

## Execution summary (Ship)

TDD cycle for all three tasks (RED test first, then GREEN implementation):

- `017.001-T`: routed `checkSymlinkEscape`'s `canonicalizeReparse` error
  branch through the existing `wrapPathError` helper. New privilege-independent
  test uses a dangling directory **junction** (not a symlink) so it runs
  without `SeCreateSymbolicLinkPrivilege`.
- `017.002-T`: added a black-box, order-independent characterization test
  for the `\.\` branch's reachability; doc-comment correction (D-4) — kept
  the branch, did not delete or assert-unreachable.
- `017.003-T`: added `utf16Len` helper (manual rune-by-rune accumulator, no
  allocation) and switched `addLongPathPrefix`'s threshold comparison from
  `len(path)` to `utf16Len(path)`.

Full local review (3 independent personas: general/security/Go-idiom) —
all PASS, zero P0/P1/P2. One P3 (byte-length fast-path micro-opt,
re-proposing a plan-rejected gold-plating option) captured as P-021
deferred-scope stash entry `475E76D2` rather than fixed.

## PR lifecycle and Copilot review

PR #51. Copilot review actively requested via the documented `[bot]`-suffixed
REST fallback immediately after PR creation (correctly following the
`docs/compound/2026-09-10-p018-copilot-engagement-must-precede-merge.md`
lesson — no repeat of the 015-S process gap). Two review rounds:

- Round 1 (`24ed6d3`): 2 real review threads (branch-retention-enforceability
  overclaim in a test doc comment; missing YAML frontmatter on the new
  memory checkpoint) — both fixed in `090cbf5`, replied-to citing the
  commit, resolved via GraphQL `resolveReviewThread`. The review body also
  named 2 different "suppressed comments" with no corresponding thread
  (test-threshold coverage gap; PR-description scope wording) — both were
  also fixed proactively as valid findings even though no thread existed
  to reply to/resolve. New compound lesson captured:
  `docs/compound/2026-09-11-copilot-suppressed-comments-vs-review-threads.md`.
- Round 2 (`090cbf5`): zero new threads. Gate `SATISFIED`.

All CI checks green throughout (13 checks). Merged via merge-commit
strategy (P-009) at `75cfe3b1b954c49d149f531464ccd97f1770dd0b`.

## Post-merge closure

Cascade-close (P-015 verified fully-covered-root exception) — `017-F` is a
root feature whose only descendants are its 3 manifest tasks, all present,
no grandchildren. Fully verified: empty `returned_ids`, `archived_ids` ==
`allowed_ids` == `required_ids`, `parent_id` preserved on all tasks. AC-F4
re-confirmed: `BF5DE670`/`6B751D8B` remain active and untouched.

Runtime verification: `READY` — `internal/pathsafe` has no runtime
entrypoint of its own, and, as corrected during closure-PR review, no
`cmd/` entrypoint currently calls `config.Load`/`Validate` at process
startup, so the changed code is not live on any current runtime path. The
`internal/pathsafe`/`internal/config` test suites are the evidence.
Releasability: `READY`, no conditions.

## Post-merge closure PR (#52)

The post-merge closure itself was delivered via a dedicated closure PR,
per the branch/PR-per-release-unit rule (never committed directly to
`main`):

- Branch `post-merge/017-f-pathsafe-error-surface-parity`, PR #52
  (`chore: post-merge closure for 017-F — pathsafe error-surface parity
  and Windows long-path precision`).
- Copilot review required **three rounds** (unusual — every other gate on
  this shipment cleared in one or two): round 1 found a factual
  runtime-reachability overclaim; the round-2 fix (commit `0749dc7`)
  corrected only the primary cited occurrence and missed two additional
  verbatim recurrences of the same claim in sibling documents/sections;
  round 3 (commit `38a45e9`) corrected the remaining recurrences. Gate
  reached `SATISFIED` with 6 threads resolved across the 2 fix rounds.
- Reviewed HEAD at merge: `38a45e9`. Merged via merge-commit strategy
  (P-009) at `b7c8678e03d144948509fd8884e09c959a914219`. All applicable CI
  green.
- Backlog index resynced twice (`backlogit sync`) — after shipment closure
  and again after this closure PR merged. Both `CLOSURE_INDEX_SYNC_OK`.
- **Process learning**: when a correction spans multiple
  documents/sections, grep the *whole* affected document set for the error
  pattern before considering the fix complete, rather than fixing only the
  location(s) explicitly cited by the reviewer — fixing only the cited
  location left two sibling recurrences of the same factual error
  unaddressed into a third review round.

## Compounding value carried forward

- `docs/compound/2026-09-06-requesting-copilot-review-non-collaborator-repo.md` — reconfirmed accurate (worked twice this cycle).
- `docs/compound/2026-09-10-p018-copilot-engagement-must-precede-merge.md` — reconfirmed accurate; lesson correctly applied.
- `docs/compound/2026-05-07-backlogit-shipment-status-constraints.md` — reconfirmed accurate (cascade transitioned `active -> shipped -> archived` as documented).
- `docs/compound/2026-09-11-copilot-suppressed-comments-vs-review-threads.md` — new entry from this cycle.

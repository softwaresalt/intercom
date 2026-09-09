# Compacted memory — 014-S / 015-F retired-architecture gate detection-quality hardening

- Release unit: shipment 014-S, feature 015-F, tasks 015.001-T–015.013-T
- Dates: 2026-09-08 / 2026-09-09
- Outcome: **SHIPPED** — PR #46 merged (merge commit
  `d7be7e883f1d309a22037aae97ec769cb647b020`), shipment cascade-closed,
  post-merge closure complete (`closure_status: READY`, `releasability:
  READY`), compact-context invoked.
- Original verbose checkpoints archived to `docs/archive/memory/`:
  `2026-09-09-stage-014-s-session.md`, `2026-09-09-ship-014-s-session.md`.

## Stage phase (queue → review → harvest)

- Queued shipment 014-S covering feature 015-F + 13 tasks, consolidating 8
  selected stash entries (`24B63533`, `8988120D`, `6285779C`, `8D603E29`,
  `9A14D3B7`, `E403E3C5`, `09CD6ACF`, `F4F4A959`) — all "Cluster B" members
  flagged by the prior 013-S cycle as the strongest next-cycle candidate.
- P-021 C5 duplicate scan clean; C6 late-identifier reconciliation resolved
  against PRs #36 and #29.
- Plan hardening required and applied. Plan review: rev1–rev3 FAIL →
  P-013.6 escalation (route `gpt-5.6-sol`/`openai`/`high`, not
  same-route-degraded) → adjudicator reproduced a blocking `TOMLDecodeError`
  and specified corrections → rev4 PASS, zero P0/P1.
- Scope guard: manifest == harvest_ids exactly (14 items).
- Known workspace degradation carried forward (not specific to this
  shipment): task schema has no `complexity` field despite CLI/registry
  advertising it; `size` set structurally, `complexity` preserved as
  enum-validated prose in `implementation-notes`.
- Deferred for future cycles: 14 stash entries remain active (`6C24E2E4`,
  `EF9352FB` [severity lowered high→low per operator rule 6], `A92E3FA0`,
  `BF5DE670`, `F47DB9A9`, `2787DA56`, `4C5BEC23`, `9D45E62E`, `1C6C3B46`,
  `4A01C53E`, `4104AF54`, `E428AB46`, `E4C5413F`, `37FAB8C2`). Next natural
  single-shipment candidates: `A92E3FA0` or `BF5DE670`. Operator-only
  residual: Tavily key rotation for `EF9352FB` still outstanding if the key
  ever left the repo.

## Ship phase (implement → review → merge → close)

- All 13 tasks implemented test-first (harness-red → implement → green),
  hardening `scripts/check-retired-architecture.sh`'s TOML/Go
  tokenization, adding a differential dual-engine self-test suite +
  `--self-test-integrity` mode, and flipping the CI advisory toggle to
  fail-closed by default.
- Adversarial multi-model review (3-persona) found/fixed M-1 (P0, tokenizer
  infinite loop), M-2 (P1, fused-plural evasion), U-1 (P1, doc inaccuracy),
  U-2 (P2, doc overclaim). Verdict `READY`.
- PR #46 opened with full Local Review Readiness block.
- Copilot review: 2 rounds. Round 1 (HEAD `7c00203`) — 3 comments
  (struct-tag regex grammar gap, missing closure-doc frontmatter,
  misleading fixture comment), all P-021 C1 in-scope, fixed directly
  (commit `eb6a816`), replied + resolved via `gh api graphql`. Round 2
  (HEAD `eb6a816`, after explicit reviewer re-request — Copilot requests
  are consumed per review and don't auto-re-trigger on push in this repo)
  — 0 new comments, gate `SATISFIED`.
- Merged via `gh pr merge --merge` (normal merge, no admin fallback). Merge
  Confirmation Gate passed.
- Post-merge closure: P-015 verified fully-covered-root cascade
  (`backlogit shipment ship 014-S`) — all verification gates passed (empty
  `returned_ids`, exact two-set `allowed_ids`/`required_ids` match,
  `parent_id` preserved). Self-caught process near-miss: closure commit
  briefly landed on `main` before push, corrected to
  `post-merge/retired-architecture-gate-detection-quality-hardening`
  without data loss or history rewrite (no actual violation — caught
  pre-push).
- Runtime-verification verdict `READY` (the hardened script is itself the
  live CI entrypoint, already exercised by PR #46's own CI run) — contrast
  with the 013-S/pathsafe `BLOCKED` precedent (no wired live caller).
- Source artifact cleanup: no `source_stash_id`/`source_deliberation_id`
  fields present on 015-F/014-S; the 8 stash IDs exist only as prose
  provenance citations in task descriptions (consistent with 013-S
  precedent — two data points now confirm this is the workspace norm).
- compound-refresh: reviewed
  `docs/compound/2026-09-06-requesting-copilot-review-non-collaborator-repo.md`,
  confirmed accurate and validated by this session's own experience —
  kept unchanged. No other compound entries affected.
- No P-021 deferred-scope entries created; no open follow-ups; no
  outstanding P0/P1 findings.

## Key learnings (durable)

- backlogit `shipment` records cannot transition to `shipped` via a direct
  `move` in this backlogit version — cascade (`backlogit shipment ship`)
  is required, contradicting the shipment-reconcile skill's generic
  "non-cascading move" description of Safe-Close Mode step 8.
- Copilot PR-review requests are consumed after each submitted review and
  must be explicitly re-requested (`POST requested_reviewers` with the
  `[bot]`-suffixed login) for each new HEAD — they do not auto-re-trigger
  on push in this repo's configuration.
- When a change's own surface is the live CI gate itself (not a library
  awaiting a caller), runtime-verification can reach `READY` directly from
  the shipment's own CI run — no separate runtime-verification activation
  event is needed.

## Full detail (if needed)

Archived verbose originals: `docs/archive/memory/2026-09-09-stage-014-s-session.md`,
`docs/archive/memory/2026-09-09-ship-014-s-session.md`. Closure artifacts
(not compacted — under 14-day threshold): `docs/closure/014-S-015-F-post-merge-closure.md`,
`docs/closure/2026-09-08-014-s-015-f-retired-arch-gate-hardening-runtime-verification.md`,
`docs/closure/2026-09-08-retired-arch-gate-hardening-adversarial-review.md`,
`.backlogit/reconcile/014-S-safe-close-2026-09-08.md`.

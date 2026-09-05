---
title: "Compacted memory: 006-S -- CI supply-chain hardening (PR #20)"
date: 2026-09-05
shipment: 006-S
feature: 007-F
pr: 20
merge_commit_sha: c627a9d3e28354b10323bf1d6605529fc679c267
---

# Compacted memory — 006-S / 007-F

## Outcome

Shipped and closed. 6 tasks + covering feature 007-F, cascade-closed via
the P-015 verified fully-covered-root exception. PR #20 merged
(`c627a9d3e28354b10323bf1d6605529fc679c267`). No Go source changes.

## Key decisions

* Used the CASCADE close path (not manual safe-close): 007-F is a root
  feature fully covered by exactly its 6 manifest children, no
  grandchildren, no linked deliberation — all P-015 preconditions
  independently verified before invoking `backlogit shipment ship`.
* No source-artifact stash archival: `custom_fields.source_stash_id`
  absent on 007-F; the 4 originally-scoped stash IDs were already
  consumed by Stage's prior harvest.
* Preserved pre-existing unrelated operator working-tree state
  throughout (never committed `.gitignore`/`start.ps1` local edits or
  the untracked `.autoharness/gates/`, `.backlogit/hooks_queue.jsonl`,
  `.claude/`, `.github/copilot/`).
* Two out-of-scope findings captured as P-021 deferred stash entries
  rather than fixed: `11ECB954` (gitignore-negation semantic gap),
  `AFE0E95C` (pin the installer itself, not just autoharness's closure).

## Files modified (feature branch)

`.github/constraints/autoharness-lock.txt` (new), `.github/workflows/ci.yml`,
`.github/workflows/secret-scan-history.yml` (new), `.github/CODEOWNERS`
(new), `scripts/check-gitignore-append-only.sh` (new),
`scripts/testdata/gitignore/{insertion,deletion,reorder}/*` (new).

## Key learnings

* A locally configured pip index/proxy's "latest" can silently diverge
  from what a real, unrestricted-network CI runner actually resolves —
  full write-up: `docs/compound/2026-09-05-pip-index-proxy-staleness-vs-real-ci.md`.
* Cross-platform `pip download --python-version X` resolution can omit a
  real, marker-conditional transitive dependency that a native-interpreter
  resolve correctly includes.
* Naively toggling `set +e`/`set -e` inside a bash helper function is
  unsafe (errexit is global, not function-scoped) — use
  `if var="$(...)"; then ... else status=$?; fi` instead.
* Copilot code review does not auto-re-review every push in this repo's
  configuration; re-request explicitly via the GraphQL `requestReviews`
  mutation with Copilot's `botIds` (found via an existing review's
  `author { ... on Bot { id } }`), since `gh pr edit --add-reviewer` and
  the REST `requested_reviewers` endpoint both reject Copilot's bot
  identity.

## Failed approaches (caught and corrected before/without operator impact)

* Initial `autoharness` pin to `1.4.11` (wrong — real CI resolved
  `1.5.0`); caught by real CI, not local verification.
* Initial markdown-table pipe-escaping fix (`sed` on the whole rendered
  row) broke the common no-embedded-pipe case; caught by a dedicated
  post-remediation re-review, not self-caught.
* Initial `git cat-file -e`-based first-ever-file check couldn't
  distinguish genuine absence from other failure modes; replaced with
  `git ls-tree`-based check per re-review feedback.

## Follow-ups

`11ECB954`, `AFE0E95C` (see above) — both require Stage deliberation
before scheduling. Full detail: `docs/closure/006-S-007-F-post-merge-closure.md`.

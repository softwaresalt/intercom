---
title: "Ship session: 006-S -- CI supply-chain hardening (PR #20)"
date: 2026-09-05
agent: ship
shipment: 006-S
feature: 007-F
pr: 20
status: complete
---

# Ship session memory — 006-S / 007-F

## Items completed

All 6 tasks (007.001-T .. 007.006-T) + covering feature 007-F, shipped and
archived via the P-015 verified fully-covered-root cascade close path.

* 007.001-T (U1): hash-pinned `autoharness` install (`--require-hashes
  --only-binary=:all:`), corrected mid-flight from a wrong version
  (1.4.11, resolved against a stale internal proxy) to the real
  CI-resolved 1.5.0 after CI caught the break.
* 007.002-T (U2): scheduled full-history secret scan, redacted + deduped
  GitHub-issue handoff.
* 007.003-T (U3): advisory-only `CODEOWNERS`, AC-0 pre-flight verified
  no live branch-protection/ruleset code-owner requirement.
* 007.004-T/007.005-T (U4a/U4b): test-first `.gitignore` append-only
  fixtures + self-test stub, then the real subsequence-predicate logic.
* 007.006-T (U5): wired the checker into `ci.yml`, added to `ci-gate`.

## Blocked conditions

None outstanding. Two items determined out of authorized scope were
captured as P-021 deferred stash entries rather than blocking:
`11ECB954` (gitignore-negation semantic gap), `AFE0E95C` (pin the
installer itself, not just autoharness's closure).

## Branch state

* Feature branch `feat/006-s-ci-supply-chain-hardening`: merged via PR
  #20, merge commit `c627a9d3e28354b10323bf1d6605529fc679c267`.
* Post-merge closure branch `post-merge/007-f-ci-supply-chain-hardening`:
  active, holds backlog-archive + closure-doc + compound-learning commits;
  closure PR to be opened next.

## Decisions with rationale

* **Preserved pre-existing operator working-tree state throughout**
  (`.gitignore`, `start.ps1` modifications; `.autoharness/gates/`,
  `.backlogit/hooks_queue.jsonl`, `.claude/`, `.github/copilot/`
  untracked) — never touched, never committed, per explicit operator
  instruction. One accidental `git add .backlogit/` staged
  `hooks_queue.jsonl`; caught and unstaged before commit.
* **Proceeded with branch creation despite a dirty pre-existing working
  tree** — an explicit, pre-briefed exception to the standard "halt on
  dirty worktree" rule, since the operator had already identified and
  authorized this exact condition.
* **Used the CASCADE close path, not safe-close**, for shipment closure:
  007-F is a root feature fully covered by exactly its 6 manifest
  children with no grandchildren and no linked deliberation — the P-015
  verified-fully-covered-root exception's preconditions were all
  independently verified before invoking `backlogit shipment ship`.
* **No source-artifact stash archival performed**: `007-F.custom_fields`
  has no `source_stash_id`; the 4 originally-scoped stash IDs (EFFAA358,
  90350C9A, F7C6420D, BEDD2E70) were already absent from the active
  stash (consumed by Stage's prior harvest) — consistent with Ship's
  role boundary (discretionary stash archival is Stage-only).
* **No `file-lock` acquisition** during shipment-reconcile: single-agent,
  single-branch session, no concurrent editing indicated (concurrency
  protocol exemption); the lock scripts are not installed in this
  workspace regardless.

## Errors encountered and resolution

1. **Wrong `autoharness` version pin** (1.4.11 vs the real 1.5.0) — see
   `docs/compound/2026-09-05-pip-index-proxy-staleness-vs-real-ci.md` for
   the full root-cause writeup. Caught by real CI, not local verification.
2. **Copilot review found 3 real issues** (job-level `continue-on-error`
   masking install failures; `grep` exit-1 mishandling; process-
   substitution error-swallowing) — all fixed, but the FIRST fix attempt
   introduced a NEW bug (naive `set +e`/`set -e` toggling clobbered an
   outer caller's errexit state, silently aborting the self-test) — caught
   by re-testing before pushing, fixed with the `if var=$(...); then ...
   else status=$?; fi` pattern.
3. **My own markdown-table-escaping fix (in the pass-2 adversarial-review
   remediation) was itself broken** (`sed` on the fully-rendered row
   instead of per-field) — caught by a dedicated post-remediation
   re-review pass, not by me.
4. Copilot does not auto-re-review every push in this repo config; had to
   explicitly re-request via the GraphQL `requestReviews` mutation
   targeting Copilot's bot node ID (`botIds` input field) after
   discovering `gh pr edit --add-reviewer` and the REST
   `requested_reviewers` endpoint both fail for Copilot's bot identity.

## Next steps

1. Open the post-merge closure PR from `post-merge/007-f-ci-supply-chain-hardening`.
2. Run local review + §1.9 readiness gate on the closure PR.
3. Present for operator/dark-mode merge approval.
4. After the closure PR merges: return to `main`, pull.
5. `compact-context` invocation (P-020) — see this same session's
   compaction note for outcome.

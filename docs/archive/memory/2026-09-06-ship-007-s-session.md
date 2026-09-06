---
title: "Ship session memory: 007-S -- Build and scan script hardening"
date: 2026-09-06
shipment: 007-S
feature: 008-F
pr: 22
status: shipped
---

# Ship session: 007-S / 008-F

## Items completed

* 008.001-T (U2, symlink/junction escape tests) -- done
* 008.002-T (U1, build.ps1 guard symlink/junction awareness) -- done
* 008.003-T (U4, multi-line TOML fixtures + manifest) -- done
* 008.004-T (U3, TOML masking rewritten to tomllib parse) -- done
* 008.005-T (U5, `--self-test` wired into CI) -- done
* 008-F (covering feature) -- done, archived
* 007-S (shipment) -- shipped, archived (P-015 fully-covered-root cascade)

## Items blocked

None.

## Branch state

* Feature branch `feat/build-and-scan-script-hardening-008-f` -- merged via
  PR #22 (merge commit `c521b2cf07555493ad35a1d20844acbba3f12045`).
* Post-merge closure branch `post-merge/008-f-build-scan-script-hardening`
  -- created from `main` post-merge; carries backlog closure + this memory
  + the closure artifact + compaction commits; PR to be opened next.

## PR status

PR #22 merged (merge commit, per P-009; no squash/rebase/admin used).
3 rounds of Copilot review, all threads resolved via GraphQL. Standard +
adversarial multi-model review both ran pre-PR with full remediation and
re-review; zero residual blocking findings at merge.

## Decisions with rationale

* **Branch created despite a dirty `main`** (operator's own uncommitted
  `.gitignore`/`start.ps1` edits plus several untracked tooling
  directories): operator explicitly pre-disclosed and named these exact
  files/directories with an instruction to preserve them, never
  reset/clean/stash/discard/sweep. Verified the diffs were unrelated
  tooling changes before proceeding; never staged or committed them at any
  point; double-checked after every `git add` that no preserved path was
  swept in (caught and corrected one accidental inclusion of
  `.backlogit/hooks_queue.jsonl` in a backlog-closure commit via
  `git reset --soft HEAD~1` + selective re-stage, recommitted clean).
* **P-015 cascade close selected over safe-close** for 007-S: 008-F is a
  root feature, fully covered by its 5 manifest tasks with no other
  descendants, and the manifest contains nothing beyond the feature +
  its tasks -- verified live against `.backlogit/queue/`+`archive/`
  before invoking `backlogit shipment ship`. This was also forced by the
  installed backlogit version (1.10.1): the generic
  `move --status shipped` path the `shipment-reconcile` skill's
  safe-close step 8 describes no longer exists (hard error, exit 9) --
  confirmed this repo's own compound entry
  (`2026-05-07-backlogit-shipment-status-constraints.md`) already
  documents this correctly, so no compound-refresh was needed there.
* **Copilot review had to be explicitly requested via REST with the
  `[bot]`-suffixed login** (`copilot-pull-request-reviewer[bot]`) -- the
  unsuffixed login and `gh pr edit --add-reviewer` both failed with a
  "not a collaborator" error. New compound entry captured:
  `2026-09-06-requesting-copilot-review-non-collaborator-repo.md`.
* **One Copilot finding classified out of scope per P-021 C1** (the
  `repoRootWithSep` double-separator edge case) -- pre-existing logic
  carried over unchanged from the original `build.ps1`, fails safe, not a
  bypass, cannot manifest for this repo. Captured as deferred entry
  `707FE72B`, thread replied to and resolved without a code change (the
  threadless-vs-thread-present P-021 procedure's thread-present path).
* **One standard-review P2 finding classified out of scope** (Windows-
  junction test has no Windows CI runner) -- captured as deferred entry
  `DA945722` before any PR existed (threadless path).
* **Two review findings WERE fixed in-scope** despite initial temptation
  to defer: the fallback TOML lexer's fail-open EOF bug (same contract
  surface as U3's own AC-6), and the dangling-reparse-point bypass in the
  guard's ancestor walk (same function as U1, non-negotiable I5 stakes).
  A third round found the fix for the latter was itself too broad
  (suppressed all Get-Item errors, not just not-found) -- narrowed again.

## Errors encountered and how they were resolved

* `git commit` for backlog closure state initially swept in
  `.backlogit/hooks_queue.jsonl` (an operator-preserved untracked file)
  via an overly broad `git add .backlogit/`. Caught immediately by
  reviewing `git status --short` output post-commit. Fixed via
  `git reset --soft HEAD~1`, selectively re-staged excluding the file,
  recommitted. No data loss; file remains untracked as required.
* `backlogit move 007-S --status shipped` (the generic path the
  `shipment-reconcile` skill's safe-close step 8 names) hard-failed with
  exit code 9 on this backlogit version. Resolved by verifying the P-015
  fully-covered-root exception applied and using
  `backlogit shipment ship 007-S` instead (the cascade path), which is
  explicitly permitted for exactly this shape of shipment.

## Next steps

* Push `post-merge/008-f-build-scan-script-hardening`, open the closure PR,
  await operator approval (does not inherit PR #22's approval).
* Stage to triage deferred entries `DA945722` and `707FE72B` in a future
  triage cycle.

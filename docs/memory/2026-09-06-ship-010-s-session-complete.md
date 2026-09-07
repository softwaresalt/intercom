---
title: "Ship session complete: 010-S residual hardening round 2 merged"
date: 2026-09-06
shipment_id: 010-S
feature_id: 011-F
pr: 29
merge_commit: 797fc1ed9e4833374552d8bfaa3ddd41e46fec26
status: complete
---

# Ship Session Summary — 010-S / 011-F

## Outcome

**Shipped.** All 20 units (`011.001-T`–`011.020-T`) completed, reviewed,
merged via PR #29 (merge commit `797fc1ed9e4833374552d8bfaa3ddd41e46fec26`).
Shipment 010-S closed via the P-015 verified fully-covered-root cascade
path (`backlogit shipment ship`), status `shipped` → `archived`.

## Items completed

All 20 tasks — see `docs/closure/2026-09-07-010-s-011-f-post-merge-closure.md`
for the full disposition table with commit SHAs.

## Items blocked / deferred

* `WINDOWS_GATE_REQUIRED` flip (`011.010-T`): mechanical assertion merged
  and proven green in CI; the actual flip requires explicit operator
  confirmation (Constitution VII), not self-authorized.
* 9 out-of-scope findings deferred to stash (P-021 C2): `F47DB9A9`,
  `ECE3DAB7`, `F4F4A959`, `5FE4A7BE`, `5158769F`, `2787DA56`, `CA469B3D`,
  `4C5BEC23`, `805248F7`.

## Branch state

* Feature branch `feat/010-s-residual-hardening-round-2-011-f-review-finding-backlog-from-005-s-009-s-harness-bootstrap-carryover`
  merged and closed.
* Post-merge closure branch `post-merge/011-f-residual-hardening-round2`
  in progress (this session), carrying: backlog archival commit, this
  closure artifact, a compound learning entry, and the mandatory
  compact-context pass. Not yet PR'd as of this checkpoint.

## Decisions with rationale

* **Process correction recorded**: mid-closure, a backlog-archival commit
  was accidentally made directly on local `main` (never pushed) before
  the `post-merge/{slug}` branch was created. Caught immediately via
  `git status` (showed "ahead of origin/main by 1 commit"), corrected by
  branching off that commit and hard-resetting local `main` back to
  `origin/main` before any push occurred. No corruption of remote `main`.
* **P-015 cascade path used** (not safe-close) for the 010-S backlog
  close: 011-F is a root feature fully covered by exactly its 20 manifest
  tasks (verified directly against `.backlogit/queue/` +
  `.backlogit/archive/` — no extra descendants at any depth) before
  invoking `backlogit shipment ship`. `returned_ids: []` confirmed zero
  unintended side effects. This was also the only available path: this
  backlogit version rejects the generic `backlogit move <shipment> --status
  shipped` step 8 of safe-close ("shipment must be shipped via
  ShipShipment, not a direct status update"), making cascade the sole
  viable close mechanism once the fully-covered-root precondition was
  confirmed.
* **Two critical CI-authoring bugs found and fixed during this shipment's
  own review/CI cycle** — see the compound learning at
  `docs/compound/2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md`
  for full detail: (1) a self-matching bare-substring grep in a
  CI assertion step, (2) a YAML step-boundary line accidentally deleted
  during an earlier edit, producing a workflow GitHub Actions could not
  parse at all. Both were invisible to local YAML-syntax validation and
  a full 9-persona local review; the first was caught by multi-model
  adversarial review, the second only by the actual first CI run.
  `actionlint` was added to this session's own verification discipline
  as a result and is recommended for all future `ci.yml` edits in this
  repository.

## Errors encountered and how they were resolved

* PowerShell/WSL bash boundary quirks (line-ending normalization needed
  for every new `.sh`/`.go` fixture file before it would run; ambient
  `COPILOT_EXE_PATH`/`PATH` ("C:\Tools") pre-set on the dev machine
  required explicit test-level overrides to avoid false test passes).
* `git check-ignore --verbose` reports the LAST matching pattern even
  when it is a negation (`!pattern`), which required negation-aware
  parsing in `check-unignore-regression.sh`'s batched differential check
  (a naive "any match = ignored" reading silently inverted several
  real-repo results during development).
* `git ls-files --others` (no `-z`) C-quotes unusual filenames, which
  broke round-tripping through `git check-ignore --stdin`; fixed with
  `-z`/NUL-delimited enumeration throughout.

## Next steps

1. Push `post-merge/011-f-residual-hardening-round2`, open the closure
   PR, run local review + CI + Copilot review to the same standard as
   the feature PR, obtain operator approval, merge (merge-commit only).
2. Backlog index resync after closure merges.
3. Operator follow-up: review the green `windows` job CI evidence and
   decide whether to authorize the `WINDOWS_GATE_REQUIRED` flip.

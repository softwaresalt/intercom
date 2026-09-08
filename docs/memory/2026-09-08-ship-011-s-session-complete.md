---
title: "Ship session complete: 011-S pathsafe containment correctness (dark-factory)"
date: 2026-09-08
shipment: 011-S
feature: 012-F
mode: dark-factory (DARK_MODE_ACTIVE)
---

# Ship Session Complete — 011-S

## Incident recovery (pre-shipment)

Stage had committed shipment 011-S's staging artifacts directly to `main`
as `f6d5879` (P-001/P-016 violation). Recovered via 3 reviewed PRs, all
merge-commit, all with local + Copilot review:

1. **PR #31** (`chore/revert-stage-011-s-direct-main`) — corrective revert
   of `f6d5879`. Mechanically verified exact inverse
   (`git diff --cached --exit-code f6d5879^` clean before and after
   commit). Merge SHA `68d0f1659d4a8d327251aca74d105d4b86f8aa47`.
2. **PR #32** (`chore/010-s-closure-filename-fix`) — fixed a pre-existing,
   upstream-owned (`softwaresalt/autoharness`) closure-artifact filename
   contract-drift defect blocking 011-S's `PREDECESSOR_CLOSURE_INCOMPLETE`
   pre-claim check for 010-S. 4th instance of the same narrow rename
   mitigation (prior: 001-S, 005-S, 008-S). Merge SHA
   `64fdee7af0361fd046dd7baa7532467d3d7fe398`.
3. **PR #33** (`chore/stage-011-s`) — proper reapplication of the preserved
   staging artifacts (cherry-picked from `f6d5879`, preserved at
   `origin/chore/stage-011-s-preserve`) via branch → PR → review → CI →
   Copilot → merge commit. 2 Copilot comments fixed/replied/resolved.
   Merge SHA `42b9530c6f19a09350199c2102a1a98f307ac331`.

Stale detached worktree registration `/tmp/tmpyf_7dfsx` was pruned after
dry-run + remote-preservation confirmation. `git worktree list` clean
throughout (single implementation worktree).

## Implementation (PR #34)

Claimed shipment `011-S` after both pre_claim and post_claim topology-gate
verification passed. Delegated 9-task harness-first/TDD implementation to
a general-purpose engineering agent; independently re-verified every
quality gate myself (build/vet/lint/test/race) rather than trusting the
agent's own report. Ran 5 independent local persona reviews plus a 4-model
adversarial consensus review before PR creation. Fixed all P2/P3 findings
that were same-surface/low-risk; deferred the 1 critical
(pre-existing, out-of-chartered-scope) finding and remaining advisory
items via P-021 C2 stash capture (`700B41CE`, `AD0D9D1F`).

PR #34 merged via merge commit, no admin fallback, Copilot gate SATISFIED
for final HEAD `902c05f`. Merge SHA
`54d27ba953a068cea539a1f691e5c4ad1231d2ba`.

## Post-merge closure (this branch, `post-merge/011-s-pathsafe-containment-correctness`)

* Merge confirmation gate: PASS (`merge-base --is-ancestor` verified).
* Predecessor-closure filename fix (PR #32) also unblocks any future
  shipment depending on 010-S's closure evidence.
* Shipment closed via P-015 verified fully-covered-root cascade
  (`backlogit shipment ship`): `returned_ids: []`, `archived_ids` exactly
  matched the computed two-set gate (11 items). Archive integrity (P-007)
  confirmed clean, no restore needed.
* Operational closure artifact:
  `docs/closure/011-S-012-F-post-merge-closure.md`. Releasability: READY.
* Runtime verification: no runtime-surface adapters touched; CLI cold-start
  smoke check (`intercom --help`, `intercom-ctl --help`) both exit 0.
* P-020 compact-context: **done** (Tier-1, this shipment's own memory
  compacted; see `docs/memory/compacted/2026-09-08-011-s-pathsafe-containment-compacted.md`).
* Compound learning captured:
  `docs/compound/2026-09-08-adversarial-review-empirical-verification-and-scope-check.md`.
* Source artifact cleanup: no `source_stash_id`/`source_deliberation_id`
  fields populated on `012-F` (predates convention for this artifact);
  no action taken, consistent with 010-S precedent.
* Backlog index resynced (`backlogit sync`) after every mutating step.

## Excluded stash items (untouched, per dark-mode scope)

`A92E3FA0`, `4989A42D`, `EF9352FB`, `BF5DE670`, `F47DB9A9`, `F4F4A959`,
`2787DA56`, `4C5BEC23` — none read, harvested, or modified this session.

## Next steps (for Stage)

* `700B41CE` (critical, requires deliberation): live Windows junction
  final-path-component containment bypass. Pre-existing, not introduced by
  011-S. Needs its own red-phase-first design.
* `AD0D9D1F` (medium, 5 of 7 items still open, requires deliberation for 3):
  remaining adversarial-review advisory findings.
* Pre-existing `F133AB7E` (GO-14 reconsideration) and `2362BBB5` (P3
  pathsafe residuals) remain open from this shipment's own plan review,
  unchanged.

## Closure PR

This memory file, the operational-closure artifact, and the compacted
memory are being submitted via a `post-merge/011-s-pathsafe-containment-correctness`
closure PR, following the same local-review → CI → Copilot → merge-commit
protocol as the implementation PR. Awaiting that PR's own review cycle.

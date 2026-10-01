---
title: "037-S / 047-F post-merge closure session memory"
description: "Dark-mode Ship session: 037-S M4 (retire gate-engine Python), PR #87 merge and post-merge closure"
date: 2026-10-01
tags: [ship, dark-mode, post-merge-closure, gate-engine, go-migration]
mode: dark-factory
shipment: 037-S
feature: 047-F
---

# 037-S / 047-F post-merge closure session memory

## Scope

This was a dark-factory run (P-017), scoped strictly to shipment 037-S: "Gate-engine Go migration M4:
retire gate-engine Python". It covers feature 047-F and tasks 047.001-T to 047.009-T. The operator was
AFK. Merge approval was pre-authorized and admin fallback was not. Admin fallback was never used.

## Completed

* Tasks 047.001-T to 047.009-T were implemented on the feature branch:
  * M4 parity evidence: the comparison of `f221023` against `b51d1f0` matched exit codes 15/15 and
    streams 30/30.
  * Operator-approval record for removal (§5).
  * `scripts/lib/*.py` engines, Python CI steps and Python docs removed.
  * Retire guards in `tools/gatecheck`.
* Pre-PR multi-persona adversarial review: 4 reviewers, P0=0 and P1=0. The remediation commits were
  `d9eefab` and `99b8d00`. One deferral was captured to stash `44F8CC48` (P-021 C2).
* PR #87 went through 23 Copilot reviews:
  * Rounds 1 to 22 raised findings. The 22 fix commits run from `75f107e` to `5ff8b19`.
  * Round 23 was clean.
  * All 18 threads were replied to and resolved through GraphQL.
  * `autoharness gate copilot-review` returned SATISFIED.
* PR #87 merged by merge commit as `ec6d1d9938559292363508134ddbf462a62e564d` at 2026-10-01T18:15:58Z.
  The merged head was `5ff8b19`.
* Post-merge closure ran on branch `post-merge/037-s-closure`:
  * The a0 topology gate and the a1 gate (047-F → done) both passed.
  * classify-close-path returned CASCADE / FULLY_COVERED_ROOT, with binding `0f9be8a3…92d9`.
  * A fresh-snapshot recompute matched that binding before the cascade ran.
  * `backlogit shipment ship 037-S` archived 11 IDs. P-007 was clean.
  * Four reconcile reports were written.
  * Wrote the closure document, the compound refresh, the new compound entry and the compaction.

## Files touched in closure

* `.backlogit/reconcile/037-S-*.md` (4 reports)
* `.backlogit/archive/` (037-S, 047-F and 047.001-T to 047.009-T)
* `docs/closure/037-S-047-F-post-merge-closure.md`
* `docs/closure/2026-10-01-037-s-compound-refresh.md`
* `docs/compound/2026-09-30-shipment-reconcile-skill-bypass-manual-substitution-risk.md` (updated)
* `docs/compound/2026-10-01-textual-yaml-guard-denylist-bypass-loop.md` (new)
* `docs/memory/compacted/2026-10-01-037-s-047-f-compacted.md`

## Decisions and disclosures

* **Tasks left `active` (disclosed lapse).** The tasks stayed `active` through the PR and were moved to
  `done` only at closure, before a1. The closure document discloses this.
* **Revalidation was agent-executed.** No invokable `safe-close` exists, so a session-local helper
  performed the safe-close revalidation. Stash `D10D3AFC` stays relevant.
* **Why the Copilot loop was long.** The CI-YAML guards were textual denylists, and each Copilot round
  found another equivalent YAML spelling. The lesson is captured as a new compound entry.

## Next steps (Stage-owned)

1. Triage the follow-up stash entries: `44F8CC48`, `D10D3AFC`, `C312BD4C` (the masker multiline-tag
   stash), `7223218F`, `5A8EC1BC`, `978D2946` and `50E6F22C`.
2. No action is needed on C44E2C1F. Stage archived it on 2026-09-28, and its staging memory is
   compacted here.
3. Re-plan held shipments 030-S, 031-S and 032-S (features 033-F, 034-F and 035-F) onto Go and clear
   their STAGE HOLD. The suggested order is 032-S, then 030-S, then 031-S.

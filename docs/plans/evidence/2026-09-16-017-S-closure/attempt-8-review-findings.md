---
title: "Plan review attempt 8 — findings and remediation (017-S closure plan)"
date: 2026-09-16
status: complete
agent: Stage
shipment: 017-S
attempt: 8
reviewed_revision: 11
reviewed_sha: 714fc2f
remediated_revision: 12
dispatch_mode: multi-agent
decision: ADVISORY
---

# Plan review attempt 8 — 017-S closure plan

Authorized by the operator token
`AUTHORIZE PLAN REVIEW ATTEMPT 8 FOR 017-S — full multi-persona review of revision 11 with the
open P-013 circuit acknowledged`.

Full-document multi-persona review of **both** governing artifacts at `714fc2f`:

* `docs/plans/2026-09-16-intercom-go-017-S-closure-plan.md` (revision 11)
* `docs/plans/2026-09-16-017-S-closure-canonical-contract.md`

## Verdicts

| Persona | Verdict | P0 | P1 | P2 | P3 |
|---|---|---|---|---|---|
| Correctness | ADVISORY | 0 | 0 | 2 | 4 |
| Constitution | ADVISORY | 0 | 0 | 0 | 6 |
| Scope boundary | ADVISORY | 0 | 0 | 0 | 2 |

**First non-FAIL outcome in this lineage.** Attempts 1–7 all returned at least one FAIL. No
persona found a defect that strands the shipment, makes a gate unsatisfiable, drives Ship to an
unsafe action, breaches the 13-member manifest, weakens the `PA-017-CASCADE` escrow, or falsely
certifies policy compliance.

Independent confirmations recorded by the reviewers:

* Every post-claim step (`S-3`…`S-22`, `R-1`…`R-8`) has an explicit, safe disposition; the
  Regime A/B split is correct and every post-claim halt routes to the **operator**, never to
  Stage (which P-010 forbids from disposing of shipments).
* No variable is consumed before its binding step.
* The manifest holds at exactly 13 across the claim auto-transition, both archive-routed
  `→ done` relocations, and the cascade close.
* `018.008-T` stays locked to its recorded 3 files; `_ship.agent.md` remains excluded.
* All three P-021 deferrals (`DB12DA37`, `11B75632`, `3F546E63`) remain out of scope and
  unsmuggled.
* P-020 non-blocking-failure semantics are reproduced, not inverted.

## Findings and disposition

Every finding was independently verified by direct measurement before remediation. All are
same-contract-surface reference-integrity or disclosure-accuracy defects — none widens `017-S`,
so no P-021 capture was required.

The dominant defect class held: four of the findings are stale citations left by revision 11's
`D-C*` → `M-*` identifier collapse, where the **contract** retained the pre-collapse reference
while the plan already carried the correct one.

| # | Severity | Location | Finding | Remediation |
|---|---|---|---|---|
| 1 | P2 | contract `D-H5` | Cited `M-23` (mode existence) for the "writes no report" fact, which is `M-19` | Citation corrected to `M-19`, matching plan §6.1 D5a |
| 2 | P2 | plan §3 consumer block | `M-23` was orphaned — its only occurrence was its own definition row, falsifying the block's "no row is inert" invariant | Bound `M-23` → `S-17` (the step that invokes the mode) |
| 3 | P3 | plan §3 consumer block | `M-13` was bound to a §8 clause that did not exist — §8.1 cited only `M-11`/`M-12` | §8.1's forbidden-mechanism list extended to name `move --status archived` as a silent no-op (`M-13`), making the binding true and closing a real hazard |
| 4 | P3 | contract `D-H1(e)` | Cited `M-19` for the archive **routing table**, which is `M-17` | Citation corrected to `M-17`, matching plan §6.1 |
| 5 | P3 | contract `D-H3` P2 | Cited `M-20` (cascade emission set) for manifest membership | Corrected to `M-1`/`M-3` manifest set-equality |
| 6 | P3 | contract §F | `{revert_merge_sha}` bound at "R-5 step (4)"; `D-K7` binds it at step (5) | Corrected to step (5); plan §7 already correct |
| 7 | P3 | plan §7 | `{pre_cascade_sha}` first consumer recorded as `R-8`; contract §F says `R-6` | Plan aligned to `R-6`, the true earliest functional consumer |
| 8 | P3 | plan §4.2 | Claimed the S-7 pass criterion adopts P-004's precondition "verbatim" while substituting a scoped `-run` invocation | Wording corrected; both `go test ./...` and the scoped run are now required, with the equivalence stated |
| 9 | P3 | plan §13 | Mapped only the derived `P-*` policies, never the constitutional Principles I–XI that Governance requires | Rows added for Principles III/IV, VII and IX |
| 10 | P3 | plan §13 P-013 row | Embedded review-loop narrative (attempt count, escalation route) in the executable procedure | Reduced to the required disclosure; attempt history moved here |
| 11 | P3 | plan Plan Review | "Current status" prose still described revision 8/9 while the document was revision 11 | Rewritten to the current, non-narrative state |

Findings 10 and 11 were flagged independently by all three personas and also discharge the
operator's standing directive that the plan is not a log.

## Verification performed after remediation

* **Mechanism-reference closure** over every changed token (`M-13`, `M-17`, `M-19`, `M-20`,
  `M-23`, `{pre_cascade_sha}`, `{revert_merge_sha}`): every occurrence in both artifacts
  enumerated by literal search and dispositioned; zero stale residue.
* **Contract↔plan parity** re-verified on both corrected variable-table rows.
* Markdown lint, `backlogit doctor`, `backlogit sync`, and ID-contiguity checks across
  `AC-*`, `I-*`, `PF-*`, `M-*`, `R-*` and every `D-*` family.

## Gate state

Attempt 8 recorded `decision: ADVISORY`, **not** `PASS`.

`PF-2` requires a literal `decision: PASS` row in the plan's verdict ledger before `017-S` is
claimable. No such row exists. Stage does not self-promote an ADVISORY verdict to PASS; that
disposition is the operator's, per the plan-review gate's ADVISORY rule ("present findings to the
operator; proceed if the operator confirms").

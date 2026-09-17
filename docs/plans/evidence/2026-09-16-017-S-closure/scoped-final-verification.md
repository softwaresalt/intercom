---
title: "017-S scoped final verification and OPERATOR_ACCEPTED_RESIDUALS disposition"
description: "Bounded four-class verification of the attempt-10 P1 remediations on revision 14, and the operator residual-risk acceptance recorded for eligibility ground 3."
status: complete
date: 2026-09-17
shipment: 017-S
feature: 018-F
plan: docs/plans/2026-09-16-intercom-go-017-S-closure-plan.md
contract: docs/plans/2026-09-16-017-S-closure-canonical-contract.md
pr: 57
---

# Scoped final verification — 017-S closure plan

## Authorization

The operator issued exactly:

```text
AUTHORIZE SCOPED FINAL VERIFICATION FOR 017-S — verify only the four attempt-10 P1
remediations on revision 14; if they pass, record OPERATOR_ACCEPTED_RESIDUALS as a
PASS-equivalent disposition for eligibility ground 3. Amend PF-2 and the canonical
contract in place to recognize this disposition, without adding review-history prose.
Do not run another full multi-persona review. Stop at PR #57 merge authorization.
```

This authorized a bounded verification, not plan-review attempt 11. No multi-persona
review was dispatched.

## Scope

Revision 14 at `ca32335`. Four remediation classes only.

## Results

| Class | Subject | Result | Residual |
|---|---|---|---|
| V1 | `{run_log}` and plan-generated evidence excluded consistently from changed-path and cleanliness bindings, without weakening production or manifest protection | PASS after 1 fix | Contract "Binding discipline" paragraph spelled the pre-D-H6 two-term pathspec while citing `D-H6` |
| V2 | P-009 merge strategy and topology verified before merge and confirmed after merge | PASS after 1 fix | Plan §13 **Principle XI** row named S-11/S-22 only, omitting the new S-10 pre-merge gate |
| V3 | Every post-claim S5/C3/failure disposition routes to the operator, never Stage | PASS | none |
| V4 | Cross-references, step IDs, row refs, P-004 literal count, PF-6 anchor, §13 mappings, AC labels, P-008 lint governance | PASS | none |

### V1 detail

Every `git status` invocation across plan and contract was enumerated and classified
as carrying the canonical exclusion, citing `D-H6` by reference, or carrying neither.
`S-12` and `R-5(1)` correctly cite `D-H6`. Plan L575 and L628 mentions are narrative,
not executable. The one executable spelling that contradicted its own cited clause was
corrected in place.

No weakening: the `D-H6` exclusion and the `D1(e)` carve-out affect only the binding of
`{harness_paths}` / `{impl_paths}` and the cleanliness gates, scoped to a single
directory — every other `docs/plans/**` path stays denied. Production-code protection is
`P5`, which asserts `git diff --name-only {merge_base}...HEAD` **repo-scoped and
independent of the evaluation set**, so no exclusion can conceal a production change.
Manifest protection (`P2`, exactly 13 members) and backlog-scope protection (`P3`) are
untouched.

### V2 detail

Verified the S-10 pre-merge gate (`allow_merge_commit` check plus `gh pr merge --merge`),
S-11 demoted to empirical post-merge confirmation, S-22 pre-merge, contract `D-I1`/`D-I2`,
and the §13 P-009 row. Only the §13 Principle XI row lagged; corrected to match.

### V3 detail

All remaining Stage-routed halts are pre-claim (PF-1…PF-12, S-0, §6.1 D5 "Before S-3",
§8.1 Regime A, AC-26). S-14 `S5` quotes the installed wording and routes to the operator
under Regime B / `D-J11`. S-14 `S3` carries an explicit success disposition.

### V4 detail

Defined-ID sets: S=24, AC=35, R=8, PF=19, M=29, contract D-=82 (all contiguous). Zero
dangling AC-, PF-, M-, D- references. `S-018` is part of a filename; `S-20.3` is S-20's
numbered postcheck 3 — both false positives, not defects.

## Residual remediation

Both residuals are same-contract-surface under P-021 C1: clerical propagation gaps of
the very remediations under verification (`D-H6`, `D-I1`), not new scope. Applied
contract-first, then remastered into the plan. No correction narrative was added to
either executable artifact.

## Disposition

`OPERATOR_ACCEPTED_RESIDUALS` is recorded as the winning verdict-ledger row for
revision 15 at `4b92d52`.

This is an **operator residual-risk acceptance**, not a reviewer PASS. No persona issued
a PASS on any revision. It rests on: attempts 8, 9 and 10 each returning **zero P0**; the
attempt-10 P1 findings being remediated and scoped-verified above; and the operator
electing to accept the remaining unverified residuals rather than fund an eleventh
attempt.

Contract **D-L7** governs the disposition and bounds it: it satisfies PF-2 and `D-L6`
exactly as `PASS` does, and carries no further authority. It does not authorize the claim
(`D-L2`), does not widen `PA-017-CASCADE` scope (`D-L3`), and does not exempt any of the
nine escrow conditions from re-measurement at invocation (`D-L4`).

### Residuals accepted

Not covered by the four verified classes, and accepted by this disposition:

- Constitution persona F4 — harness-architect stub emission versus `D1(a)` / `D2` / `P5`.
- Constitution persona F6 — P-001 pre-flight placement versus post-claim `S-6`.
- Constitution persona F8 — `S-10` records no P-014 readiness artifact.
- Scope persona — the `S-3` MCP escape hatch.
- Contract `M-29` internal tension.
- Residual correction-voice phrasing at contract C216/C234/C243 and plan L811/L1054–1057.

### Out-of-scope deferrals (unchanged)

P-021 captures `DB12DA37`, `11B75632` and `3F546E63` remain deferred and outside 017-S.

## Eligibility ground 3

PF-2 has two conditions:

- **(a) winning `decision` cell** — **discharged** by the disposition row above.
- **(b) plan present on `origin/main`** — **not discharged**; requires PR #57 to merge.

`017-S` is not claimable until (b) lands.

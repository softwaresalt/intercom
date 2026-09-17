---
title: "Plan review attempt 10 — 017-S closure plan revision 13"
description: "Multi-persona current-HEAD confirmation review of revision 13; consensus verdict and remediation map into revision 14."
date: 2026-09-17
attempt: 10
revision_reviewed: 13
reviewing_sha: e63ecc2
dispatch_mode: multi-agent
personas: [correctness, constitution, scope-boundary]
decision: FAIL
---

# Plan review attempt 10 — revision 13 @ `e63ecc2`

Operator authorization: `AUTHORIZE PLAN REVIEW ATTEMPT 10 FOR 017-S — current-HEAD
confirmation review of revision 13`.

Scope: full multi-persona current-HEAD confirmation review of the canonical contract, the
remastered plan, the accumulated review evidence, the `017-S` record and the PR #57
description/readiness block.

## 1. Persona verdicts

| Persona | Verdict | P0 | P1 | P2 | P3 |
|---|---|---|---|---|---|
| Correctness | **FAIL** | 0 | 4 | 5 | 3 |
| Constitution | **FAIL** | 0 | 3 | 6 | 7 |
| Scope Boundary | **ADVISORY** | 0 | 0 | 8 | 7 |

**Consensus decision: `FAIL`.** No `decision: PASS` is recorded for `e63ecc2`.

Consensus-weighted P1 count after independent measurement and de-duplication across personas:
**4**. Zero P0 findings in any persona — revision 13's structural remediations held.

## 2. Confirmation objectives — outcome

Attempt 10's mandate was to confirm the two attempt-9 P1 fixes and the five restored attempt-8
remediations. Independently measured result:

| Item | Status |
|---|---|
| Attempt-9 P1 #1 (post-claim halts routed to operator, not Stage) | **HOLDS** at all four sites |
| Attempt-9 P1 #2 | **HOLDS** |
| Attempt-8 restored remediations #2–#5 | **HOLD** outright |
| Attempt-8 restored remediation #1 | **PARTIAL** — the false claim was removed but the replacement §13 citations were wrong |

## 3. Confirmed findings and remediation into revision 14

Every finding below was independently re-measured against the working tree before it was
accepted. Findings that measurement did not substantiate were rejected and are not listed.

### P1-1 — `{run_log}` collides with the deny-list, the path bindings and the cleanliness gates

Revision 13 introduced `{run_log}` at `docs/plans/evidence/2026-09-16-017-S-closure/run-log.md`
and a `D-G6` duty to commit it at S-21.5. That artifact was self-defeating in three independent
ways, each of which strands `017-S` **after the irreversible claim**:

1. the path matches the `D1(e)` `docs/plans/**` deny glob, so the S-22 closure-PR diff halts;
2. both changed-path bindings excluded only `.backlogit/`, so the continuously appended
   uncommitted run log was ingested into `{harness_paths}`/`{impl_paths}`, making the `D2`
   closed-set membership rule and the `≤ 3` cardinality bound unsatisfiable by construction;
3. `git status --short` at S-1, S-12 and R-5(1) reports it — S-1 fails cheaply under Regime A,
   but S-12 fails post-claim and post-merge with no `unclaim` available.

**Remediation (contract first).** `D-H6` was rewritten as the single canonical exclusion
pathspec `-- . ':(exclude).backlogit/' ':(exclude)docs/plans/evidence/2026-09-16-017-S-closure/'`
and extended to govern **every** cleanliness assertion, not just the two path bindings.
`D-H1(e)` gained an explicit carve-out for the evidence directory; `D-H5` added `{run_log}` to
the S-22 expected-artifact set; `D-G6` gained the cross-references. The plan was regenerated:
four path bindings and three cleanliness gates now cite `contract D-H6`, `D1(e)` carries the
carve-out, `D5a` carries the expected-set membership, and **AC-35** was added so the invariant
is verifiable rather than incidental.

### P1-2 (Constitution F2) — no pre-merge P-009 gate

`.github/policies/workflow-policies.md` L185–207 fixes the P-009 Gate Point literally at
*"Ship Step 5 pre-merge (before any merge action)"*, and `constitution.instructions.md` L186
(Principle XI, NON-NEGOTIABLE) requires verifying the merge strategy *"before executing any
merge"*. Revision 13 only **asserted** merge-commit shape at S-10 and **confirmed** it
post-hoc at S-11. Post-hoc detection is unrecoverable: a squash or rebase merge already on
`main` cannot be undone without history rewriting, which Principles VII and XI forbid.

**Remediation.** `D-I1` was rewritten as a genuine pre-merge gate requiring the
`gh api repos/{owner}/{repo} --jq` merge-configuration check and the explicit
`gh pr merge --merge` method, quoting both sources. `D-I2` was reworded to "confirms … after
it lands (the pre-merge gate is D-I1)". The plan's S-10 now carries the full pre-merge gate
with the policy quotations; S-11 is explicitly a confirmation gate; S-22, R-5(5) and AC-17
were extended; the §13 P-009 row now names the gate.

### P1-3 — residual post-claim halt routed to Stage (S-14 `S5`)

Plan L418 `S-14` selector branch `S5` routed a post-claim halt "return to Stage", contradicting
`D-J11` and P-010. Measurement confirmed this was the **only** residual post-claim
Stage-routed halt — every other `HALT to Stage` site is pre-claim and therefore correct.

Measurement also confirmed the mitigating claim: "return to Stage" is **verbatim installed
wording** at `.github/agents/_ship.agent.md` L809. The fix is therefore **disclosure, not
amendment**. `D-J11` gained an installed-wording precedence clause (quote the installed text,
but route the disposition to the operator); S-14 `S5` now quotes the installed wording and
halts to **OPERATOR (Regime B)**. S-14 `S3`'s missing success disposition was added in the
same edit.

### P1-4 — dangling and incorrect cross-references

Measured: **15** `§6.1 D-XX` citations pointing at a section that defines only
`D1, D2, D3, D4, D5, D5a, P1–P6`; a dangling `D6` at plan L327 and L623; a dangling
"§8.1 S-20 row" in §13's P-007 row (§8.1 covers S-0/PF-1 through the S-18 pre-mutation
refusal and contains no S-20 row); wrong `D1(a)`/`P1` citations in the §13 Principle III/IV
row; a "for R-7" mis-anchor in PF-6; and "five-literal red phase" at two plan sites plus
contract `D-D8` where the red phase is **six** literals.

**Remediation.** All 15 `§6.1 D-XX` citations were mechanically rewritten to `contract D-XX`
(verified 0 remaining, 0 stray `D6`); the §13 P-007 row now cites the §8.2 rollback trigger;
the §13 Principle III/IV row cites `D-H6`/`D1(e)`/`P2`/`P3`; PF-6 states the `D-G4` anchor
explicitly (S-15, never PF-6); and the red-phase literal count was corrected to six at all
three sites. AC-1's mislabel of `S-13` was corrected to the `a0` topology gate.

### P2 — P-008 markdown lint had no canonical decision

`markdownlint` appeared at three plan sites but S-21.5 carried no lint clause and no `D-`
row governed it. **Remediation.** New contract row **`D-G7`** fixes two gate points —
pre-S-8-commit and pre-S-21.5-commit — both pre-commit, therefore always recoverable.
S-21.5 gained the lint bullet and the `{run_log}` commit bullet, which also repaired the
AC-34 anchor.

## 4. Dispositioned without change

| Finding | Disposition |
|---|---|
| Correctness P3 "`{impl_paths}` unconsumed" | **Rejected** — measured consumed at plan L322 |
| Scope `S-3` MCP escape hatch | Advisory; behaviour is within the authorized contract surface |
| P-002 structurally unsatisfiable under claim auto-transition | Already the governing **deferred blocker** `DB12DA37`; not remediated inside 017-S |
| Claim-last restructuring | Already deferred as `3F546E63` |
| `return-blocked` single-item semantics | Already deferred as `11B75632` |

No finding required scope expansion beyond the same-contract surface, so **no new P-021
capture was created by attempt 10**. The three pre-existing deferrals remain out of scope.

## 5. Method note

Personas operate read-only with no shell. Standing Stage discipline is to independently
re-measure every finding against the installed workspace before accepting it; four separate
persona claims were rejected on measurement in this attempt. Contract-first remediation was
applied throughout: the canonical decision is changed once at its authoritative location and
the executable plan is regenerated from it, with three-direction mechanism-reference closure
(plan→contract, contract→plan, sibling sections) run after every changed mechanism.

## 6. Outcome

Revision 13 is **rejected**. Revision 14 carries the remediation. Attempt 10 consumed
**review-fix cycle 2 of a maximum 3**.

A further review attempt against revision 14 requires **separate explicit operator
authorization** — the P-013 circuit is open and acknowledged, and Stage may not self-authorize.

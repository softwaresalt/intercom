# Plan review attempt 7 — findings and remediation

**Artifact reviewed:** `docs/plans/2026-09-16-intercom-go-017-S-closure-plan.md` revision 8
**Canonical contract:** `docs/plans/2026-09-16-017-S-closure-canonical-contract.md`
**Reviewed SHA:** `394a45f`
**Dispatch mode:** multi-agent (3 personas, parallel, full-plan + full-contract)
**Decision: FAIL**
**Remediated into:** revision 9

---

## Persona verdicts

| Persona | Verdict | P0 | P1 | P2/P3 |
|---|---|---|---|---|
| Correctness Reviewer | **FAIL** | 2 (1 refuted) | 4 | 10 |
| Constitution Reviewer | **FAIL** | 4 | 7 | 12 |
| Scope Boundary Auditor | **ADVISORY** | **0** | 3 | 10 |

Scope-boundary returned **zero P0 for the second consecutive attempt**, again confirming the
13-member manifest holds on every traced path and that revision 8 introduced no deliverables or
obligations beyond the authorized closure scope. It specifically cleared the deny-list redesign —
the change most at risk of over-granting — and confirmed `DB12DA37` and `11B75632` remain properly
deferred.

---

## One reported P0 was REFUTED by direct measurement

**Correctness F-1** asserted that S-11's check (iii) could never pass, because `git show --stat` on
a two-parent merge emits the dense combined (`--cc`) diff, which lists only files differing from
*all* parents — empty for a conflict-free merge. It further asserted this invalidated M-21's cited
evidence (raised separately as F-9).

Measured against the actual merge commit:

```text
git show --stat f2d4cf9        -> 17 files changed, 1307 insertions(+), 6 deletions(-)
git diff --stat f2d4cf9^1 f2d4cf9 -> 17 files changed, 1307 insertions(+), 6 deletions(-)
```

`git show --stat` enumerates all 17 files, **identical** to the first-parent pairwise diff. The
premise is false on this repository and this commit. **F-1 and F-9 are both withdrawn**, and M-21's
evidence citation stands as recorded.

This is worth stating plainly: the finding was well-argued and internally coherent, but rested on
an unverified claim about tool behaviour. Reviewer assertions about tool semantics are measured
before they are adopted — the same discipline that produced the `M-` matrix in the first place.

---

## Genuine findings, verified and remediated

### F-2 (P0) — `{impl_paths}` had no binding command, making S-8 unsatisfiable

S-8 said *"Bind `{impl_paths}` — the exact changed-path set"* and named no command. The parallel
`{harness_paths}` binding (S-7) specifies `git status --porcelain=v1 -- . ':(exclude).backlogit/'`
and states the exclusion is **required**, because the S-3 claim writes `.backlogit/` records that no
step has yet committed and ingesting them *"would make the check unsatisfiable by construction."*

That rationale applies identically to `{impl_paths}` — the same uncommitted claim writes plus the
S-2 reconcile report are in the working tree at S-8 — but was never propagated. The natural reading
ingests them, none is in D2's closed three-path set, so S-8's membership test fails ⇒ **guaranteed
post-claim halt**, the exact stranding mode §6.1 exists to eliminate.

**Fix:** D-H6 now governs *every* changed-path binding, not just `{harness_paths}`. Both bindings
share one rule, and `.backlogit/**` is stated to lie outside D2's domain entirely (tested only by
P3), so a `.backlogit/` path can never be a D2 membership failure.

### C-01 (P0) — post-claim halts routed to Stage

Four sites routed halts to Stage that occur **after** the S-3 one-way door: S-7 twice (L285, L295),
S-10's P-021 C3 disposition (L336), and the §11 S-4 risk row (L733). Regime A ("HALT to Stage") is
defined only for `017-S` *not* `active`. Routing a post-claim halt to Stage hands a stranded
`active` shipment to the one agent P-010 forbids from disposing of shipments — producing either
indefinite P-001 slot occupation or a P-010 breach.

**Fix:** all four now route to the operator under Regime B, governed by new **D-J11**: halt routing
follows the *regime*, never the authoring agent. Stage re-measurement is advisory input, never the
halt target.

### C-02 (P0) — the P-002 mitigation's load-bearing producer was validated post-claim

The entire §4.1 P-002 deviation rests on the `harness-architect` skill applying the label. That
producer was first exercised at **S-7 — after the claim** — and the plan's own text predicts
refusal: *"`018.008-T` is already `active` at this point, and the skill's Step 1 excludes non-ready
work items."* M-29 establishes only that the skill is *installed*, not that it admits an `active`
task. Installed P-012 requires probing before relying on a capability.

**Fix:** new **PF-12** validates producer admission **pre-claim**, where failure costs nothing
(Regime A, zero mutation). S-7's residual failure path now halts to the operator.

### C-03 (P0) — P-005 carried one of three required actions in the general clause

Installed P-005 enumerates three unconditional required actions. §13's map asserted all three, but
the *general* clause governing every non-S-0 policy halt, and **AC-22**, carried only the event
record. AC-22 therefore certified a one-of-three shortfall as acceptance-passing — the same
false-compliance shape as attempt 6's C-02.

**Fix:** the telemetry clause, AC-22 and **D-J9** now enumerate the full triple, with action 2
recorded `N/A (no PR exists at this gate)` when the halt precedes PR creation.

### C-05 (P1, high value) — P-020 failure is NON-BLOCKING, and the plan would have halted

Installed P-020 states verbatim: *"A compact-context run that **FAILS** is **NON-BLOCKING**: record
`compaction: degraded` in the closure artifact, log a warning, and continue closure."* S-21.5 had no
failure disposition, so §8's fail-closed default would have halted — **inverting the installed
policy after the merge had already landed**.

**Fix:** **D-J12** explicitly excludes S-21.5 from the fail-closed default and from every R-trigger.
Only *skipping* the invocation is the violation.

### F-6, F-8 (P2)

* S-21.5's insertion made the closure branch carry **three** commits; R-5's justification clause
  still said two. The revert target itself was already correct.
* D-C11 claimed a report is persisted *"for every mode it runs"* — contradicted by the installed
  skill and by the contract's own D-C16. Restated to the three modes that actually persist.

---

## C-04 — P-013 circuit state: disclosed, with the framing corrected

The constitution persona is **right** that the circuit is open and was unrecorded. Seven consecutive
FAILs is far past P-013.3's three-attempt threshold, and neither artifact carried a P-013 row, an
escalation-route record, or the P-005 telemetry.

It is **not** right that each attempt was an unauthorized re-execution. The P-013.6 escalation was
compiled and dispatched once, to `gpt-5.6-sol`/`openai`/`high` — genuinely distinct from Stage's own
Tier-3 route, so not `ESCALATION_DEGRADED` — and its findings were consumed. Every attempt from 4
onward proceeded on an exact, explicitly recorded operator authorization token. P-013.6's
authority-preservation invariant bars the **agent** from self-authorizing a re-attempt; the terminal
state it mandates is *"halt + handoff for operator review,"* and an operator then directing another
attempt **is** that review resolving, not a bypass of it.

**Fix:** recorded rather than disputed — a **P-013** row in §13 and **D-J13** in the contract, both
stating the open circuit, the dispatched route, and the authorization condition on any further
attempt.

---

## Mechanism-reference closure — results, and its limit

Applied to the revision-9 mechanisms, the sweep again caught what the targeted edits missed:

* **§7's binding-discipline paragraph** still carried the old unqualified `{impl_paths}` wording — an
  exact F-2 relapse, in the very edit meant to fix F-2.
* **PF-12 landed out of order** in the preflight table (between PF-10 and PF-11).

Both would have been attempt-8 findings.

**But its limit should be stated honestly.** The sweep only covers tokens chosen in advance. Attempt
7's F-2 and C-01 were *themselves* the dominant defect class, in mechanisms nobody had enumerated —
binding commands and halt-regime routing were not on the six-token list the operator named for
attempt 7. The countermeasure narrows the failure mode; it does not close it.

---

## Mechanical validation after remediation

* AC-1 … AC-31 strictly sequential, no duplicates or dangling references
* I-1 … I-13 parity between plan and contract
* Contract D-J1 … D-J13 contiguous
* PF-1 … PF-12 correctly ordered
* No orphan fragments or truncated table rows
* `backlogit doctor` — No issues found; `backlogit sync` — 244 artifacts
* markdownlint — 0 issues

---

## Status

`017-S` remains **NOT claimable**. Eligibility ground 3 (a review-PASSed governing closure plan
present on `origin/main`) is still **OPEN** and remains the sole blocker.

Revision 9 is remediated but **unreviewed**. Attempt 8 requires explicit operator authorization.

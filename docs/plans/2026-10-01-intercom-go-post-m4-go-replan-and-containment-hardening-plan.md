---
title: "Plan — Post-M4 Go re-plan (030-S, 031-S, 032-S), gatecheck CLI containment hardening, and the write-path go/ast follow-up"
description: "Retargets held plan units 6, 7 and 8 from the retired Python engines onto tools/gatecheck, plans a new containment-hardening unit from stash entries 7223218F, 50E6F22C and 44F8CC48, and splits unit 7 (031-S) into a masked-text hardening increment plus a deferred go/ast follow-up (049-F / 039-S)"
source_document: docs/decisions/2026-10-01-intercom-go-post-m4-stash-triage-and-go-replan-deliberation.md
agent: Stage
date: 2026-10-01
stage_branch: chore/stage-post-m4-followups-and-go-replan
bound_snapshot: 9b299c8 (origin/main); Unit E rev 7 re-plan bound to 961b652 (origin/main)
revision: 7
requires_plan_hardening: yes
---

# Plan — Post-M4 Go re-plan and gatecheck CLI containment hardening

**Source document:**
`docs/decisions/2026-10-01-intercom-go-post-m4-stash-triage-and-go-replan-deliberation.md`

**Scope.** Five units. Units A and B retarget held shipments 030-S and 032-S
onto the Go gate engines. Unit C is a new release unit harvested from stash
group G1 plus G6. Unit D re-plans held shipment 031-S as the **Option A
masked-text hardening increment** (decision D-031-2), folding in stash entries
B72E9715 and 8E9F8E55. Unit E is the **new, deferred `go/ast` follow-up**
(feature 049-F, shipment 039-S) that receives the parts of the original unit 7
that are hard as text scanning and that a `go/ast` engine would replace.

**Out of scope.** Stash groups G2, G3, G5 — deferred with recorded rationale
(deliberation §1.5). Stash entry **DD0BB60F** was out of scope through
revision 4. **Revision 5 harvests it into 030-S as A-T3a/A-T3b/A-T3c
(033.004-T, 033.005-T, 033.006-T)** by operator direction (decision D-030-4).

**Requires plan hardening: yes.** See `## Plan Hardening`.

> **Revision 2** incorporates plan-review round 1 (FAIL, 2×P1 + 6×P2 + 5×P3).
> Material corrections: the Unit A "silent pin" hazard narrative was **factually
> inverted** and is rewritten (A-T2); the A-T1/A-T2 split was not independently
> landable under the `unused` linter and is restructured; bare gate invocations
> that always exit 1 are corrected; Unit C is split from 4 to 6 tasks to respect
> the 2-hour rule and width isolation.

> **Revision 3** incorporates plan-review round 2 (FAIL, 1×P1 + 1×P2 + 5×P3) and
> round 3 (ADVISORY, 9×P3). Material corrections: A-T1 was still not
> independently landable — it necessarily reddens the pathspec pin — so **INV-5**
> and the **ALP-1** atomic-landing-pair declaration were added with a bounded
> INV-1 carve-out and a pair-boundary gate (AC-A2.6); the refactor-prohibition
> comments A-T1 supersedes are now owned (AC-A1.7); **C-T2 and C-T3 were
> inverted** so the wrapper test supplies the red before the guard lands.
>
> **Gate status: ADVISORY — proceed with operator awareness.** No P1 or P2
> findings remain. The round-3 P3 residuals are applied in this revision.

> **Revision 4** records operator decision **D-031-2** (deliberation §4.5): 031-S
> is **split**. Units A, B and C are **unchanged** by this revision. Unit D
> (031-S, revised) is the Option A increment — a frozen detection oracle,
> call-extent extraction (D-1′), the access-mode allowance predicate (D-2′), six
> new selectors (four planned plus B72E9715's `io.CopyN`/`io.CopyBuffer`),
> residual documentation, and three fixture tasks (the last harvested from
> 8E9F8E55). D-1′/D-2′ are the Go-era amendments of D-1/D-2 recorded as
> deliberation **D-031-3** after rev-4 review round 1. Unit E (049-F / 039-S,
> new) holds the deferred `go/ast` work — a Stage-executed spike, the engine
> migration, the allowance-predicate port, named import alias resolution
> (034.006-T, re-parented) and the `Root.Resolve`/`NewRoot` tripwire (034.005-T,
> re-parented). Unit E is **dependency-gated on 031-S** and carries its own
> STAGE HOLD pending the spike. Hardening entries for both units are added to
> H-1..H-5 and a new **H-6 Rollback** section. Revision 4 is gated separately
> under `## Plan Review — Revision 4` (rounds 1–2 FAIL, remediated; **round 3
> ADVISORY, no P0/P1 → 031-S STAGE HOLD LIFTED**); the revision-3 verdict above
> covers Units A–C only.

> **Revision 5 — changelog (2026-10-01, operator decisions on the rev-4 open
> items; deliberation §4.7).** Units B and C are **unchanged**. Unit D's
> execution content is unchanged except as noted below.
>
> 1. **D-031-3 acknowledged** by the operator (D-031-4). D-1′/D-2′ are now
>    operator-ratified, not pending.
> 2. **H-5 re-measure is POLICY** (D-031-5), replacing a CI tripwire. It is no
>    longer a recommendation. This session's run is recorded in H-5 and in the
>    session memory.
> 3. **`os.Chtimes` confirmed** in 034.009-T (D-031-6). D-T6 is no longer
>    flagged.
> 4. **No-op closure authorised** for 034.001-T (031-S) and 033.003-T (030-S)
>    (D-000-2). Ship closes each inside its own shipment, `queued → active →
>    done`, with pre-task-completion gate evidence. See D-T0 and "Dropped from
>    Unit A". `.backlogit/hooks.yaml` is not changed.
> 5. **Spike scheduled** (D-049-1). E-T1's Stage-executed `go/ast` spike runs
>    **after 031-S merges and before 039-S is claimed**. Stash entry
>    **1EEBECA5** is the trigger for the next Stage cycle. See E-T1.
> 6. **Residual item 6 captured** as stash entry **458F9385** (D-S-5).
>    Stage recommends folding it into 039-S / 049-F. See D-T4 item 6 and H-5.
> 7. **DD0BB60F harvested** into 030-S as **A-T3** (D-030-4), a deliberate
>    design change to D-6 / INV-6 of the gate-engine migration plan: the pin
>    keeps source-text anchoring and replaces presence with **call binding**.
>    Plan review split the work for the 2-hour rule (rounds 1 and 2) into
>    **A-T3a (033.004-T)** for the empty-input contract, **A-T3b (033.005-T)**
>    for the pathspec half, and **A-T3c (033.006-T)** for the prefix half.
>    Binding is enforced as a **use-whitelist** over one canonical syntactic
>    path per half (round 2). A-T1 gains **AC-A1.8**, which requires that
>    shape, and A-T2's sibling collector definition is pinned. All three tasks
>    are sequenced after ALP-1 and are **not** part of it; INV-5 still declares
>    exactly one pair. Unit A is re-gated under
>    `## Plan Review — Revision 5 (Unit A)`.
>
> **Revision 5 gate outcome: FAIL → 030-S back on STAGE HOLD.** Round 3, the
> final re-entry allowed, returned **FAIL**. Five personas returned ADVISORY or
> PASS-level verdicts. The Security Lens returned **five P1 findings**
> (SEC-1..SEC-5): fail-open narrowings that the canonical binding contract still
> accepts. The operator made lifting the hold conditional on ADVISORY or better
> with no P0/P1, so **033-F stays `blocked` and 030-S is on hold**. The A-T3
> tasks were still harvested by operator direction (decision 7). They are
> marked STAGE HOLD, and their acceptance criteria are expected to change at
> remediation. The Escalation Protocol resolved to `ESCALATION_DEGRADED`, so the
> next step is operator review. See the review section and the Harvest Record —
> Revision 5.

> **Revision 6 — changelog (2026-10-01, operator decisions on the rev-5 open
> items; deliberation §4.8).** Units B–E are **unchanged in execution
> content**. Only Unit A's A-T3 contract is redesigned.
>
> 1. **OPTION A chosen (D-030-6).** Stage remediates every round-3 P1
>    (SEC-1..SEC-5) and P2 (GO-1..GO-4, SC-1..SC-4, CN-1..CN-3, SEC-6, SEC-7)
>    finding. The operator authorises **one additional plan-review round
>    (round 4)** on Unit A. That round is an explicit, operator-authorised
>    exception to the 3-cycle cap, and there is **no** fifth round.
> 2. **The A-T3 contract is redesigned** from a use-whitelist into a
>    **closed-world, frozen-declaration pin**. It has four parts:
>    * the §A-CANON canonical texts, compared by token equality;
>    * a closed world for `select.go`;
>    * `scopeDataOK` set equality with magic, duplicate and shadowing
>      rejection;
>    * package closure.
>
>    The finding-to-fix map is in Unit A, "Pin contract".
> 3. **New tasks.** **A-T3d (033.007-T)** adds package closure. **A-T4
>    (033.008-T)** adds the SEC-2 `cmd/x/main.go` self-test probe. Both are
>    test-first and within the 2-hour rule. The new `blocks` edges are
>    `033.007-T → 033.006-T`, `033.008-T → 033.002-T` and
>    `033.007-T → 033.008-T`. ALP-1 is **not** split, and INV-5 is unchanged.
> 4. **AC-A1.8 is rewritten** to require token equality with §A-CANON. New
>    **AC-A1.9** requires a conformance checklist in the ALP-1 commit body
>    (SC-4/CN-6).
> 5. **IVL-1** is a bounded INV-2 interval with a 030-S merge precondition
>    (CN-3).
> 6. **Residuals off the contract surface** are captured as P-021 DEFERRED
>    SCOPE EXPANSION stash entries: **`DC921AF6`** for runner wiring outside
>    `select.go` (R-A1), and **`D7BF9F74`** for the git environment and config
>    and the `pysem` integrity (R-A2).
> 7. **P-002 skip authorised (D-000-3)** for the verified no-op closures of
>    **034.001-T** and **033.003-T only**. The CN-2 hard precondition is
>    discharged, and the authorisation is not a general waiver.
> 8. **035-F moved `blocked → queued` (D-032-4)**, so 032-S is claimable. See
>    the Harvest Record — Revision 6 for the lifecycle path used.
>
> Unit A is re-gated under `## Plan Review — Revision 6 (Unit A)`.
>
> **Revision 6 gate outcome: ADVISORY → 030-S STAGE HOLD LIFTED (D-030-7).**
> Round 4, the operator-authorised exception, returned **ADVISORY** with no
> P0/P1:
> * five personas returned ADVISORY;
> * the anchor (Architecture Strategist) returned PASS.
>
> All P2 findings, and most P3s, were applied as text amendments inside the
> reviewed design. 033-F is `queued`, and Unit A is claimable in dependency
> order. The revision-5 FAIL paragraph above is historical.

> **Revision 7 — changelog (2026-10-02, the post-spike Unit E re-plan,
> operator decision D-049-1; deliberation §4.9).** Units A–D are
> **unchanged**: 030-S, 031-S and 032-S have shipped and 038-S is closed.
> Only Unit E changes. The re-plan is bound to `961b652` (`origin/main`, after
> the 031-S merge).
>
> 1. **E-T1 spike complete.** Stage ran the `go/ast` spike under the P-016
>    spike-worktree exception (`logs/spike-049-wt`, since removed). The
>    findings are in
>    `docs/decisions/2026-10-02-intercom-go-writepath-go-ast-spike.md`.
>    * G-1 and G-2 are FEASIBLE.
>    * G-3 reaches **exact verdict parity**: 0 divergences across the 19
>      fixtures, the 3 `filebased` inputs, the 17 tracked in-scope files and
>      all 155 `.go` files in the scratch worktree (154 tracked + the harness).
>    * G-4 is a representation change only, and G-5 is additive.
>    * G-6 is syntactic intra-function tracking; `go/types` is rejected.
>    * G-7 puts every task within the 2-hour rule.
> 2. **E-T2..E-T6 lose "(provisional)".** Each now carries a spike citation,
>    an explicit file list, a scenario count and a red/green phase
>    (AC-E1.2). E-T2 is **de-risked from complexity `high` to `medium`**,
>    with the spike as the recorded de-risking step.
> 3. **New task E-T7 (049.007-T), the selector-set widening.** It folds in
>    stash entry **458F9385** (D-T4 residual item 6, review finding R4-P2-g),
>    as the rev-5 recommendation proposed. The widening is **import-path
>    keyed** and appends to `Selectors`. Golden changes are additive only
>    (AC-E7.2), and the D-T1a oracle stays bound to the frozen 20 (H-2 E).
>    E-T7 **narrows** item 6 to the 50 enumerated primitives; it does not
>    close it. The uncovered same-family remainder stays a disclosed residual
>    (stash **C0D28448**).
> 4. **Edges are linearised for golden serialisation and one shared import
>    table:** E-T2 → E-T3 → E-T4 → E-T5 → E-T7 → E-T6. The new `blocks`
>    edges are `049.004-T → 049.003-T`, `049.007-T → 049.005-T` and
>    `049.006-T → 049.007-T`.
> 5. **E-T6 is re-scoped.** It also narrows item 6 and syncs the selector
>    count and list in both header locations (AC-D4.2). Items 2 and 6 are
>    narrowed, not removed: item 2 to the G-6 cross-function residual, item 6
>    to the un-enumerated same-family remainder. It also discloses the
>    new **residual item 8**, dynamic proc invocation (`syscall.NewLazyDLL`
>    and similar). That item is live in production and stays out of scope;
>    it is captured as low-priority stash entry **FE2F02FF**.
> 6. **H-1 E, H-2 E, H-5, H-6, the Verification golden paragraph and
>    Constitution rows IV and X** are updated to match.
>
> 7. **Plan-review cycle 1 amendments** (see `## Plan Review — Revision 7
>    (Unit E)`):
>    * item 6 is **narrowed**, not closed (PR7-1, SB-1), and is added to the
>      H-2 E protected list;
>    * physical file counts are stated next to the D-T6 artifact counting
>      convention (PR7-2, SB-6);
>    * AC-E7.3 records an exact-name scan over all 50 selectors (PR7-3);
>    * the E-T1 ACs mirror 049.001-T (PR7-4, SB-3);
>    * residual item 8 and the item 6 remainder have stash IDs, and H-5 has
>      full-set re-measure commands (SB-2, SB-9);
>    * E-T7 adds a dot-import fixture row (SB-7), and H-6 orders the E-T6 and
>      E-T7 reverts (SB-8).
>
> 8. **Plan-review cycle 2 amendments:**
>    * E-T3 grows to 3 files and size **M**. It ports the AC-D1.1–D1.3 extent
>      tests in `writepath_extent_test.go` to named AST tests (AC-E3.3), and
>      deletes `scanText` and `advanceCursor` with the extractor. E-T2 states
>      that they survive unchanged until E-T3, so every task boundary compiles
>      (SB2-2);
>    * H-5 records the inclusive (`_test.go`-including) item-6 command that
>      AC-E7.3 cites (PR7C2-1);
>    * H-6 states that AC-D4.2 is a manual review check, so the E-T7 → E-T6
>      header lag is an accepted drift window (SB2-4);
>    * AC-E7.1 says canonical `unix.*`, matching E-T4's `os.*` (GM2-2);
>    * the backlog bodies of 049.006-T and 049.007-T now mirror the plan,
>      item 6 is narrowed rather than removed, and the stale description text
>      is marked superseded (PR7-1, SB2-1, SB2-6, GM2-1). The review section
>      below is added (SB2-3).
>
> 9. **Plan-review cycle 3 (FINAL) — FAIL; the hold is kept.** All three
>    responding reviewers returned FAIL with a new P1 each: gpt-5.6-sol,
>    claude-opus-5.5, and grok-4.7 standing in for an empty gemini-3.8-flash
>    response. Stage verified every P1 against the code. The fixes are
>    **applied but NOT re-gated**:
>    * E-T3's complete deletion list now names `lineReportsSelector` and
>      `findSelector`, and the AC-D2.3 same-line test is ported (SB3-1,
>      GPT3-1);
>    * non-parseable extent inputs become G-2 fail-closed parse errors
>      (GK3-1);
>    * `TestFindSelector_LookaroundTable` is re-pointed (GK3-2);
>    * the AC-E3.3 grep is extended and scoped (SB3-2..SB3-4);
>    * the E-T2 note is corrected (GK3-3);
>    * 049.005-T gains a supersession sentence (GK3-4).
>
>    The re-entry limit is exhausted, so **049-F stays blocked and 039-S
>    stays on hold** pending an operator-authorized re-gate (deliberation
>    D-049-7).
>
> Unit E is re-gated under `## Plan Review — Revision 7 (Unit E)`. The 049-F
> STAGE HOLD lifts only on ADVISORY or better with no P0/P1. **Outcome: FAIL
> at cycle 3, FAIL again at the operator-authorised round 4 (D-049-8), and
> FAIL again at the operator-authorised round 5 (D-049-9), and FAIL again at the
> operator-authorised round 6 (D-049-10, where the conditional hold-lift rule
> did not apply). The operator then lifted the hold by override, without
> another review round (D-049-11; see the revision 7 gate outcome after
> item 13).**
>
> 10. **Staging PR #99 review amendments (NOT re-gated).** Copilot's shadow
>     review added `token.ASSIGN` and multi-name `var` binding coverage to E-T5, an import-path
>     scan to AC-E7.3, and a verbatim H-5 evidence file. It also hardened the
>     spike harness. See "Staging PR #99 review amendments" under the rev 7
>     review. The hold is unchanged.
>
> 11. **Round 4 (operator-authorised exception, D-049-8) — FAIL; the hold is
>     kept.** The operator approved exactly one extra round. It returned:
>     * gpt-5.6-sol (anchor): FAIL, GPT4-1 (P1);
>     * claude-opus-5.5: FAIL, OP4-1..OP4-4 (P1);
>     * gemini-3.8-flash: PASS.
>
>     Stage verified all four distinct P1s against the code:
>     * a non-parseable "parseable" extent row (GPT4-1/OP4-2);
>     * an `unparam` lint break at the E-T2 boundary (OP4-1);
>     * non-parseable predicate rows `:467–468` (OP4-3);
>     * a non-discriminating AC-D1.3 port (OP4-4).
>
>     Amendments **R4-1..R4-8** are applied as **un-gated amendments**. No
>     round 5 is authorised, so **049-F stays `blocked` and 039-S stays on
>     hold**, pending another operator disposition. See "Round 4" under
>     `## Plan Review — Revision 7 (Unit E)`.
>
> 12. **Round 5 (operator-authorised exception, D-049-9) — FAIL; the hold is
>     kept.** The operator later approved exactly one more round (round 5),
>     which supersedes item 11's "no round 5" for that single round only. It
>     returned:
>     * gpt-5.6-sol (anchor): FAIL, GPT5-1 and GPT5-2 (P1);
>     * claude-opus-5.5: ADVISORY, OP5-1 (P2), OP5-2..OP5-4 (P3);
>     * gemini-3.8-flash: PASS.
>
>     Stage verified GPT5-1 (the predicate-port rows were not specified as
>     unmasked source; masked `:464`–`:466` are also parse errors) and
>     accepted it. Stage rejected GPT5-2 as a false positive: with the real
>     package and the E-T2 caller set, `golangci-lint` reports the `unparam`
>     finding, and R4-3 clears it. Amendments **R5-1..R5-4** are applied as
>     **un-gated amendments**. No round 6 is authorised, so **049-F stays
>     `blocked` and 039-S stays on hold**. See "Round 5" under
>     `## Plan Review — Revision 7 (Unit E)`.
>
> 13. **Round 6 (operator-authorised exception, D-049-10) — FAIL; the hold is
>     kept.** The operator approved exactly one more round (round 6), which
>     supersedes item 12's "no round 6" for that single round only. The
>     approval came with a conditional rule: lift the hold if every verified
>     P0/P1 is classification-only **and** mechanically caught by a
>     test-first build. Round 6 returned:
>     * gpt-5.6-sol (anchor): FAIL, GPT6-1 (P1);
>     * claude-opus-5.5: ADVISORY, OP6-1 (P2), OP6-2..OP6-6 (P3);
>     * gemini-3.8-flash: ADVISORY, GM6-1 (P2).
>
>     GPT6-1 is PRE5-1 re-raised: `Args[3]` and `Args[6]` have no rejection
>     rows. Stage verified it with a throwaway parser probe: a predicate that
>     skips those positions passes every planned row, with 0 divergences. A
>     weak predicate therefore passes the specified tests, so condition (b)
>     fails. Condition (a) fails too: this is a coverage gap, not the
>     classification of an existing input, and doubt counts as not
>     qualifying. The result is **FAIL**. Amendments **R6-1..R6-6** are applied
>     as **un-gated amendments**. No round 7 is authorised, so **049-F stays
>     `blocked` and 039-S stays on hold**. See "Round 6" under
>     `## Plan Review — Revision 7 (Unit E)`.
>
> **Revision 7 gate outcome: OPERATOR-LIFT → 039-S STAGE HOLD LIFTED
> (D-049-11).** At 2026-10-03T22:25-07:00 the operator accepted R6-1 as
> written and lifted the hold **without another plan-review round**. This
> overrides the D-049-7 lift rule for revision 7 only:
> * no round 7 ran, and the plan-review verdict of record stays round 6 FAIL;
> * R4-1..R4-8, R5-1..R5-4 and R6-1..R6-6 are accepted as written. No
>   plan-review round has gated them, so they are **implementation-time
>   verification obligations** that Ship's TDD harness and pre-PR adversarial
>   code review must close;
> * 049-F moves `blocked → queued`, and 039-S is claimable by Ship in a new
>   session.
>
> The "hold is kept" statements in items 9–13 above are historical. See
> "Operator hold-lift decision (D-049-11)" under
> `## Plan Review — Revision 7 (Unit E)`.

---

## Invariants (all units)

* **INV-1 — No gate verdict changes.** No unit may change the PASS/FAIL verdict
  of any gate on the tracked tree at `9b299c8`, except where an acceptance
  criterion names the change explicitly. Every unit verifies against its own
  parent commit, not a fixed snapshot.
* **INV-2 — Fail-closed direction only.** Where behaviour changes, it may only
  make the gate scan more, reject more, or error louder. No unit may widen an
  allowance or convert an error into a silent pass.
  * **Exception (rev 6, CN4-5):** the bounded interval **IVL-1** (Unit A),
    which exists on the 030-S branch only and is closed by a merge
    precondition (AC-A3d.6). `main` never observes it.
* **INV-3 — No write primitives introduced.** No destructive filesystem write
  primitive may be introduced into `internal/**` or `cmd/**`; doing so is a named
  stop condition (it would trip the very gate being hardened).
* **INV-4 — Lint-clean on every task boundary.** `ci.yml` runs blocking
  `golangci-lint run ./...` (job `lint`, `.golangci.yml` `default: standard`,
  which includes `unused`) and `staticcheck ./...` (job `security`), and the lint
  job formats the whole tree **including `testdata`**. Therefore **every task
  must leave the tree lint-clean and gofmt/goimports-clean on its own** — a task
  may not land an unused declaration awaiting a later consumer.
* **INV-5 — Atomic landing pairs are declared, never implied.** Where two tasks
  are genuinely not independently landable — one necessarily reddens a gate that
  **was passing on the tracked tree** and that only its partner can re-green —
  they must be declared an **atomic landing pair** here, carry an explicit
  bounded INV-1 carve-out naming which gate is red and for how long, and move
  their full-suite gate to the **pair boundary**. Ship lands such a pair as a
  **single commit**. Exactly one pair exists in this plan: **ALP-1 = {A-T1,
  A-T2}**. No other task in this plan may be sequenced this way without amending
  this invariant.
  > **Test-first red phases are explicitly excluded.** A task that adds a *new*
  > failing test which its successor turns green (B-T1 → B-T2, C-T2 → C-T3) is
  > the constitutionally required harness-before-code posture, not an atomic
  > landing pair: the red is a newly authored assertion, not a previously-passing
  > gate changing verdict. Those pairs land as separate commits. The distinction
  > is **"was this assertion green on the tracked tree at `9b299c8`?"** — for
  > ALP-1 it was; for B-T1 and C-T2 it did not exist. (B-T1 inverts an existing
  > assertion at `git_test.go:103-121`, which is an authored change to a test
  > this plan explicitly owns, not a collateral gate break.) **A-T3a, A-T3b,
  > A-T3c (rev 5), and A-T3d and A-T4 (rev 6)** fall under the same exclusion.
  > Their red tests are new assertions authored inside each task, and they land
  > in the same commit as the code that turns them green. None of them is a
  > pair or a carve-out. The INV-2 interval that ALP-1 opens is declared
  > separately as **IVL-1** (Unit A). It is an INV-2 interval, not an INV-5
  > pair.

---

## Unit A — 030-S / 033-F: Unify retired-architecture scan scope declaration

**Retarget.** `scripts/lib/retired_arch.py` → `tools/gatecheck/internal/retiredarch/`.

**Premise correction carried forward (still true in Go):** there is **no live
scan-scope drift**. `selectRepoPaths` *does* enumerate `cmd/**`
(`select.go:126`). ALP-1 removes a maintenance hazard; it does **not** fix a
live defect. Any report claiming that ALP-1 fixed a defect is wrong. **A-T3
(rev 5)** closes a *latent fail-open in the guard itself* (DD0BB60F), not a
defect in the selected path set.

**The guarded decision (D4/AG-5) that survived the port:** `shouldScanRepoPath`
(`select.go:50-63`) is **not** a flat prefix disjunction. `cmd/` intentionally
**includes** `_test.go` and `testdata/` (`:54`); `internal/` **excludes** both
(`:57`, `:60`); `config.toml.example` (`:51`) is a third arm with no prefix at
all. A naive one-prefix-list unification would silently narrow `cmd/` coverage
while looking correct. This is the unit's central hazard.

### ALP-1 — A-T1 and A-T2 are an atomic landing pair (INV-5)

**A-T1 is not independently landable, by construction.** `pin.go` asserts against
`select.go`'s *source text*, so the instant AC-A1.2 moves the scope literals out
of `selectRepoPaths`/`shouldScanRepoPath`, `collectStringLits` finds nothing,
`containsAll` returns false (`pin.go:76`, `:80`), and the following all go red:

* `TestCheckPathspecPin_UnmodifiedCopy_Accepted` and `TestSelectionPathspecPin_LiveTree`;
* every `writeMutatedCopy` subtest — its anchor
  `git(root, "config.toml.example", "cmd/**", "internal/**")` must appear exactly
  once or the helper calls `t.Fatalf`;
* `TestCheckPathspecPin_PrefixMutation_Rejected` (`pin_test.go:133-141`), whose
  separate anchor `strings.HasPrefix(path, "internal/")` (`select.go:57`) AC-A1.2
  also removes;
* `TestRunRepoSelectionSelfTest_AssertionNamesAndOrder_MatchGolden`
  (`selftest_selection_test.go:11-43`), which fails on `out.failed`;
* the `"selection pathspec pin (AC-6/AG-1)"` assertion in `selftest_selection.go`
  (`:278-280`) — so `scripts/check-retired-architecture.sh --self-test` and
  `--self-test-integrity` flip PASS → FAIL.

This list is **exhaustive**: no other test or gate may be red at the ALP-1
interior boundary. This is `select.go:5-9`'s documented fail-closed design
working correctly, not a defect to engineer around.

Merging A-T1 and A-T2 into one task is **rejected**: it would span `select.go`,
`pin.go` and `pin_test.go` (3 files, ~5 functions) and breach the 2-hour rule.

**Therefore, per INV-5:**
* **Bounded INV-1 carve-out.** Between A-T1 and A-T2 — and *only* in the interior
  of ALP-1 — the retired-architecture pin assertion and the `pin_test.go` cases
  named above are **expected RED**. No other gate may be red at that boundary.
  The carve-out closes at A-T2.
* **Ship lands ALP-1 as a single commit.** A-T1 must not be pushed, reported
  complete, or CI-gated on its own.
* **The full-suite gate moves to the pair boundary** (AC-A2.6). A-T1's own
  acceptance is compile-, lint- and non-pin-test-scoped (AC-A1.6).

### A-T1 — Introduce `scanScope` and derive both consumers from it

Size **M**, complexity **medium**. Single file: `select.go`. **Half of ALP-1 —
not independently landable (INV-5).**

Introduce one `scanScope` declaration stating, per arm, both its `git ls-files`
pathspec form **and** its test/testdata policy, and rewire **both** consumers in
the same change: `selectRepoPaths` (`select.go:126`) builds its argument vector
from `scanScope`, and `shouldScanRepoPath` (`select.go:50-63`) evaluates from it.
Neither may restate a scope literal.

> **Why declaration and rewiring are one task (INV-4).** An unexported
> `scanScope` with no consumer is an `unused`/U1000 finding and would redden both
> blocking CI gates. A declaration-only task is therefore not independently
> landable. Scope stays within the 2-hour rule because it is one file and two
> functions.

**Acceptance criteria**
* AC-A1.1 — Exactly one `scanScope` declaration exists in `select.go`, carrying
  per-arm pathspec form and per-arm test policy **as data**, not as code paths.
* AC-A1.2 — Neither the pathspec construction nor the path predicate restates a
  scope literal; both read from `scanScope`.
* AC-A1.3 — The governing decision (D6/D6c as broadened by 012-S) is cited in the
  declaration. A future scope change is made in **one authoritative production
  declaration**, with the pin expectation (§A-CANON in `pin.go`) updated in the
  same reviewed commit, so it cannot half-apply silently (AR4-2).
* AC-A1.4 — The selected path set is **identical to the set computed at this
  change's own parent commit** (not at a fixed snapshot).
* AC-A1.5 — The `cmd/`-includes-tests vs `internal/`-excludes-tests asymmetry and
  the `config.toml.example` arm are each preserved and each independently
  asserted. The existing assertions at `selftest_selection.go:185-197` and
  `:256-276` must still pass **unmodified in intent**.
* AC-A1.6 — Scoped to this task's own boundary (per ALP-1's carve-out):
  `go build ./...`, `golangci-lint run ./...` and `staticcheck ./...` are clean
  (INV-4), and `go test ./tools/gatecheck/... -count=1` passes **except for** the
  tests exhaustively enumerated in the ALP-1 declaration, which are expected red
  until A-T2. Those are gated at the pair boundary by AC-A2.6 — they
  are **not** waived. Any *other* red test fails this task.
* AC-A1.7 — The refactor prohibitions that A-T1 deliberately supersedes are
  rewritten in the same change: the whole `select.go:4-9` paragraph (both the
  "requires each literal to appear as a `*ast.BasicLit` inside the correct
  function body" premise at `:4-7` and the "Do not refactor these literals into
  shared constants, helper variables, or another file…" prohibition at `:7-9`)
  and `select.go:121-124` ("literal string arguments in THIS call … never a
  variable, slice literal defined elsewhere, or constant reference"). All must
  state the post-`scanScope` contract and name `scanScope` as the single
  permitted home for the literals. Leaving any of them in place is a false
  instruction to the next reader.
* AC-A1.8 — **Token-equal to §A-CANON (rev 6, for A-T3).** In `select.go`,
  the import declaration, `GitRunner`, `DefaultGitRunner`, `scanArm`,
  `scanScope`, `shouldScanRepoPath` and `selectRepoPaths` are each
  **token-for-token equal** to the canonical text in Unit A, §A-CANON.
  Comments and whitespace are free. In addition:
  * `select.go` contains **only** the closed-world declarations (the A-T3 pin
    contract). That means no `init`, no `_`, no `const`, no extra `var` or
    `type`, and no method;
  * `select.go` carries no build constraint;
  * `scanScope` and `scanArm` are referenced only inside the frozen
    declarations.

  The rewritten comments (AC-A1.7) must stay true once the freeze lands. If
  `select.go` cannot be made token-equal to §A-CANON while staying lint-clean
  and verdict-identical, HALT to Stage (H-2 A).
* AC-A1.9 — **Conformance checklist in the ALP-1 commit body (rev 6;
  SC-4/CN-6).** The single ALP-1 commit body records a §A-CANON checklist with
  these items:
  * one line per frozen declaration, each marked `token-equal: yes`;
  * the closed-world inventory of `select.go`'s top-level declarations;
  * `build constraint: none`;
  * `scanScope/scanArm references outside frozen decls: none`.

  A missing or false checklist item fails ALP-1.

### A-T2 — Re-anchor the pathspec pin and add a negative control

Size **S**, complexity **medium**. `pin.go` plus its test. **Half of ALP-1 —
closes the carve-out and carries the pair-boundary gate (INV-5).**

**Corrected hazard (plan-review round 1, finding 3).** The round-1 narrative
claimed the pin would "silently stop pinning while still passing". That is
**wrong and inverted**. `checkPathspecPin` sets
`PathspecOK = containsAll(lits, pathspecPinLiterals)` (`pin.go:76`) and `PrefixOK`
likewise (`pin.go:80`). If the literals move into `scanScope`, `collectStringLits`
finds none, `containsAll` returns **false**, and the "selection pathspec pin
(AC-6/AG-1)" assertion in `selftest_selection.go` **FAILS loudly**. `select.go:5-9`
documents this as deliberate fail-closed behaviour.

**The real hazard is the opposite one, and it is unguarded today:**
`containsAll(_, [])` returns **`true`** — `containsAll` spans `pin.go:117-128`
and its vacuous `return true` is reached at `:127` when `wanted` is empty. So the
tempting way to make A-T1 land green — emptying `pathspecPinLiterals`
(`pin.go:33`) or `prefixPinLiterals` (`pin.go:37`) during this lockstep edit —
yields a **vacuously green pin that asserts nothing**. That, not a moved literal,
is what this task must prevent.

Re-anchor `checkPathspecPin` onto whichever declaration now holds the literals.
Note `findFuncBody` resolves only `*ast.FuncDecl`, so anchoring on a
package-level `scanScope` (`*ast.GenDecl`) requires a sibling collector.
**Collector definition (rev 5, pinned for A-T3's red phase):** the sibling
collector gathers **every string `*ast.BasicLit` within the `scanScope`
declaration**, whatever field it sits in. A-T3c deletes it once binding
replaces presence.

> **Not a self-comparison.** `pin.go:9-15` records that `pathspecPinLiterals` and
> `prefixPinLiterals` are **this file's own independent expectation, never derived
> from `select.go`'s text**. Anchoring the pin on `scanScope` therefore remains
> genuinely falsifiable. (Round 1's AC-A2.5 wrongly called this a
> self-comparison, creating a contradiction with AC-A2.1; that claim is
> withdrawn.)

**Acceptance criteria**
* AC-A2.1 — The pin remains **source-text-anchored**: it parses the on-disk
  `select.go` via `go/parser` and compares against independently-authored literal
  lists that are **never derived from `select.go`'s own text**.
* AC-A2.2 — **Non-vacuity guard:** a committed test asserts that
  `pathspecPinLiterals` and `prefixPinLiterals` are each **non-empty**, so the
  `containsAll(_, []) == true` vacuous-pass mode (`pin.go:127`) cannot be
  reached by emptying them.
* AC-A2.3 — **Negative control, as a committed test artifact** (not a one-time
  manual demonstration): a deliberately narrowed scope — e.g. a fixture omitting
  the `cmd/**` arm — makes the pin **FAIL**. A pin that passes against a narrowed
  scope fails this task.
* AC-A2.4 — All pre-existing `--self-test` and `--self-test-integrity` cases pass.
* AC-A2.5 — `golangci-lint run ./...` and `staticcheck ./...` clean (INV-4).
* AC-A2.6 — **ALP-1 pair-boundary gate (INV-5).** At this task's completion the
  carve-out is closed: `go test ./tools/gatecheck/... -count=1` passes in full
  — including every test enumerated in the ALP-1 declaration and deferred by
  AC-A1.6 — and `scripts/check-retired-architecture.sh`, `--self-test` and
  `--self-test-integrity` all return to PASS. ALP-1 is reported complete only
  here, as one commit.
* AC-A2.7 — The pin's own fixtures and prose are re-anchored with it: the
  `writeMutatedCopy` anchors in `pin_test.go` (including
  `TestCheckPathspecPin_PrefixMutation_Rejected`'s separate
  `strings.HasPrefix(path, "internal/")` anchor at `pin_test.go:133-141`) target
  `scanScope`, and `pin.go:1-7` — which states the pin requires the literals
  "INSIDE the `selectRepoPaths`/`shouldScanRepoPath` function bodies" — is
  rewritten to describe the `scanScope` anchor. (`pin.go:9-15`, the H-11
  self-comparison note, stays true and is **not** modified.)

### A-T3 — Freeze the scan-scope contract surface (rev 6; A-T3a / A-T3b / A-T3c / A-T3d, plus A-T4)

Harvested from stash **DD0BB60F** under decision **D-030-4**, and re-designed in
rev 6 under decision **D-030-6** to close the round-3 Security Lens findings
SEC-1..SEC-5. Files: `pin.go` and `pin_test.go` only. `select.go` is
**read-only** for every A-T3 task. A-T4 (SEC-2) owns the self-test files
instead.

> **What changed from rev 5, and why.** The rev-5 contract was a
> **use-whitelist**: it bound only the *arguments* of the `git` call and the
> `HasPrefix` call, and then tried to enumerate forbidden mutations around them.
> Round 3 showed that this cannot be made sound by adding more blacklist rows.
> The surrounding control flow (SEC-3, SEC-5), the runner body (SEC-4) and the
> data semantics (SEC-1) were all open. Rev 6 replaces it with a
> **closed-world, frozen-declaration pin**:
> * every declaration in `select.go` that decides *which paths are selected*
>   is frozen **token-for-token** against an independently authored canonical
>   text. Downstream dispatch through `scanPath`/`engineForPath` is not frozen
>   and is recorded as residual R-A1 (round 4, SEC4-2);
> * `select.go` may contain **only** the declarations named in the closed world;
> * the frozen `scanScope` data must satisfy semantic rules (`scopeDataOK`) that
>   reject magic pathspecs, duplicates, shadowing prefixes and set drift;
> * the rest of the package may not reach into the frozen surface
>   (package closure).
>
> This is a **stronger** pin than rev 5's, and is simpler to reason about:
> "the selection surface is byte-equivalent to the reviewed text" replaces
> "no forbidden shape appears". The rev-5 use-whitelist, the `S` name rule and
> the canonical-loop shape rules are **withdrawn** (they are subsumed). This
> applies the allowlist lesson in
> `docs/compound/2026-10-01-textual-yaml-guard-denylist-bypass-loop.md`
> (round 4, LR4-8).

#### §A-CANON — canonical declarations (normative)

A-T1 writes these declarations into `select.go` **token-for-token** (AC-A1.8).
`pin.go` holds the same text as its **own independently authored raw-string
expectation**, never derived from `select.go`'s text, so H-11 still holds.
Comments and whitespace are not part of the contract. Doc comments and the
AC-A1.7 rewrites are free text.

```go
import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/softwaresalt/intercom-go/tools/gatecheck/internal/pysem"
)

type GitRunner func(root string, pathspecs ...string) ([]byte, error)

func DefaultGitRunner(root string, pathspecs ...string) ([]byte, error) {
	args := append([]string{"ls-files", "--"}, pathspecs...)
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	var stdout, stderrBuf bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderrBuf
	if err := cmd.Run(); err != nil {
		if msg := strings.TrimSpace(stderrBuf.String()); msg != "" {
			return nil, fmt.Errorf("%w: %s", err, msg)
		}
		return nil, err
	}
	return stdout.Bytes(), nil
}

type scanArm struct {
	pathspec     string
	prefix       string
	exact        string
	includeTests bool
}

var scanScope = []scanArm{
	{pathspec: "config.toml.example", exact: "config.toml.example"},
	{pathspec: "cmd/**", prefix: "cmd/", includeTests: true},
	{pathspec: "internal/**", prefix: "internal/", includeTests: false},
}

func shouldScanRepoPath(path string) bool {
	for _, a := range scanScope {
		if a.prefix == "" {
			if path == a.exact {
				return true
			}
			continue
		}
		if strings.HasPrefix(path, a.prefix) {
			if !a.includeTests && (strings.Contains(path, "/testdata/") || strings.HasSuffix(path, "_test.go")) {
				return false
			}
			return strings.HasSuffix(path, ".go")
		}
	}
	return false
}

func selectRepoPaths(root string, git GitRunner) ([]string, error) {
	pathspecs := make([]string, 0, len(scanScope))
	for _, a := range scanScope {
		pathspecs = append(pathspecs, a.pathspec)
	}
	out, err := git(root, pathspecs...)
	if err != nil {
		return nil, err
	}
	listing, err := pysem.GitText(out)
	if err != nil {
		return nil, err
	}
	var selected []string
	for _, path := range pysem.SplitLines(listing) {
		if shouldScanRepoPath(path) {
			selected = append(selected, path)
		}
	}
	sort.Strings(selected)
	return selected, nil
}
```

* **Equivalence (INV-1).** The import declaration, `GitRunner` and
  `DefaultGitRunner` are identical to `select.go` at `9b299c8`. The
  `shouldScanRepoPath` loop is semantically equivalent to `select.go:50-63`.
  The `config.toml.example` arm matches exactly. `cmd/` includes tests and
  `testdata` (the D4/AG-5 asymmetry). `internal/` excludes both. Every other
  path is `false`. No path starts with two prefixes, so arm order does not
  change any verdict. The `selectRepoPaths` pathspec vector is
  `config.toml.example, cmd/**, internal/**`, in today's order. AC-A1.4 is the
  executable check.
* **Spellings decided (GO-1, GO-2).** `make([]string, 0, len(scanScope))` is
  the canonical spelling. There is no `S` name rule any more: the local is
  called `pathspecs`, and only token equality constrains it.
* **Scope changes after rev 6.** Every frozen declaration is now an edit
  surface shared by `select.go` and `pin.go`. A future scope change must edit
  both in one commit. That is a lockstep pair which the future plan must
  declare under its own INV-5. This is the intended cost: the pin turns a
  scope change into a reviewed two-file edit and never lets it pass silently.
  The import declaration is frozen too. A new import needed by the unfrozen
  `engineForPath` or `scanPath`, or any new declaration in `select.go`, also
  needs a paired `pin.go` edit (round 4, SC4-4/AR4-2).
* **Pinned `select.go` comment wording (AC-A1.7; round 4, LR4-1).** The
  rewritten `select.go:4-9` paragraph and the `selectRepoPaths` doc comment
  must carry this text. It is chosen to stay true in every IVL-1 state:
  > The scan-scope surface of this file (the imports, `GitRunner`,
  > `DefaultGitRunner`, `scanArm`, `scanScope`, `shouldScanRepoPath` and
  > `selectRepoPaths`) is pinned by `pin.go` (M2-T8; D-030-4, D-030-6).
  > Treat these declarations as frozen. Edit them only together with
  > `pin.go`'s independent expectation, in the same commit; otherwise the pin
  > fails closed. `scanScope` is the single permitted home for the scope
  > literals.

#### Pin contract (normative)

All checks are **syntactic** and use only `go/parser`, `go/ast`, `go/token`
and `go/scanner`. There is no `go/types` and no execution, and the deprecated
`ast.Object`/`Ident.Obj` (SA1019) is never used.

* **Token equality.** One helper, `declTokens`, tokenises a declaration's
  source span with `go/scanner` (comments dropped). The same helper is used on
  both sides:
  * parse `select.go` and index its top-level declarations by name;
  * parse the canonical text with `package retiredarch` prepended, and index
    its declarations by name;
  * compare the `declTokens` of each frozen name on the two sides
    (round 4, GO4-3).

  Byte offsets come from `fset.File(p).Offset(p)` and never from `Position`,
  so `//line` directives inside the free comments cannot move a span. The
  comparison is on `(token, literal)` pairs. Every `SEMICOLON` literal is
  normalised, so an automatically inserted `"\n"` equals an explicit `";"`.
  A `FuncDecl`/`GenDecl` `Pos()` excludes its `Doc` comment, so doc text is
  free. Any difference fails the frozen declaration. Ship records the stdlib
  facts this depends on (`Pos()` and `Doc`, and the auto-`SEMICOLON` literal
  at the end of a span) as GOROOT citations or a `go run` probe in A-T3b's
  evidence (round 4, LR4-4).
* **Comments and directives (round 4, SEC4-4/GO4-4).** Every pin parse uses
  `parser.ParseComments`. A directive counts only when an `*ast.Comment`'s
  `Text` starts with `//go:build`, `// +build` or `//go:linkname`, so prose
  or string literals that mention them (including `pin.go`'s own doc and
  canonical text) never count. Imports are detected through
  `ImportSpec.Path`.
* **Closed world (`select.go`).** The file contains exactly:
  * one import `GenDecl`;
  * single-spec `GenDecl`s for `GitRunner`, `scanArm` and `scanScope`;
  * receiver-less `FuncDecl`s for `DefaultGitRunner`, `shouldScanRepoPath`,
    `engineForPath`, `scanPath` and `selectRepoPaths`.

  Anything else fails closed. That includes `init`, `_`, any `const`, a second
  `var` or `type`, any method, and a duplicate of a named declaration. The file
  carries **no** `//go:build` or `// +build` constraint. The identifiers
  `scanScope` and `scanArm` occur **only** inside the **§A-CANON declaration
  names** (import, `GitRunner`, `DefaultGitRunner`, `scanArm`, `scanScope`,
  `shouldScanRepoPath`, `selectRepoPaths`). That confinement set is fixed from
  A-T3b onward, whether or not a declaration is token-compared yet (round 4,
  GO4-8/CN4-8). A reference from `engineForPath` or `scanPath`, which are not
  frozen, fails closed.
* **Flag ownership (GO-3, SC-1, CN-4).**
  * `PathspecOK` requires the frozen set {import, `GitRunner`,
    `DefaultGitRunner`, `scanArm`, `scanScope`, `selectRepoPaths`}.
  * `PrefixOK` requires the frozen set {import, `scanArm`, `scanScope`,
    `shouldScanRepoPath`}.
  * A **shared-rule** violation clears **both** flags. The shared rules are the
    closed world, the build-constraint rule, the identifier-confinement rule,
    `scopeDataOK` and package closure.
  * `SelectFound` and `GuardFound` keep their meaning: the named function
    exists. A parse error still clears everything.
* **`scopeDataOK` (A-T3c; SEC-1).** This check runs over per-arm values
  extracted from `select.go`'s `scanScope` composite literal into a
  **`pin.go`-local type**, `armValues{pathspec, prefix, exact string;
  includeTests bool}`. `includeTests` is read syntactically from the
  `true`/`false` identifier, and an absent key means `false`. `pin.go` never
  references `scanArm` or `scanScope` **as identifiers**: they appear only
  inside its raw-string canonical text, so package closure still accepts
  `pin.go` (round 4, GO4-1). The check fails closed unless every rule holds:
  * the pathspec set **equals** `pathspecPinLiterals` and the non-empty prefix
    set **equals** `prefixPinLiterals`. Equality is `containsAll` in both
    directions, reusing A-T3a's non-vacuous `containsAll`. Containment-only
    (expected ⊆ resolved) is **withdrawn**;
  * there are no duplicate pathspecs and no duplicate prefixes;
  * no pathspec begins with `:`. This rejects git magic pathspecs:
    `:(exclude)`, `:!`, `:^`, `:(glob)`, `:(icase)` and the rest;
  * no prefix is a prefix of another prefix, and no `exact` value starts with
    any prefix (first-match shadowing);
  * every prefix arm's pathspec equals its prefix followed by `**`, and every
    exact arm's pathspec equals its `exact` value;
  * **test policy (D4/AG-5; round 4, GO4-2):** the set of prefixes with
    `includeTests == true` **equals** an independent `pin.go` expectation,
    `includeTestsPinPrefixes = {"cmd/"}`. Every other arm has
    `includeTests == false`.

  **What these rules guarantee, and what they do not** (round 4,
  SEC4-5/AR4-1/SC4-5). The canonical text is a **trusted, reviewed
  specification**. `scopeDataOK` checks that specification against the
  enumerated data invariants above, using expectations authored separately
  (H-11). This catches an **accidental** or partial edit: for example, an
  `:(exclude)` arm or a flipped `includeTests` written into both `select.go`
  and the canonical text is still rejected. It **cannot** prevent a
  deliberate, coordinated edit of `pin.go` and `select.go` together, because
  every expectation lives in `pin.go`. A coordinated edit of the
  `shouldScanRepoPath` body in both files is caught only by the behavioural
  self-test, whose expectations live in `selftest_selection.go`. That is why
  A-T4's independent `cmd/x/main.go` probe is **needed**, not optional. The
  remaining control is human review of any `pin.go` diff.
* **Package closure (A-T3d).** The rule set covers every non-`_test.go` `.go`
  file in `select.go`'s directory, including `select.go` itself. Every rule
  fails closed and clears both flags. A read or parse error also fails closed.
  * Only `select.go` may contain the identifier `scanScope` or `scanArm`.
  * Only `select.go` may declare a closed-world name. This is belt and
    braces: Go already rejects the redeclaration (round 4, SC4-2).
  * No file may declare a Go **universe** name at package scope (for example
    `append`, `len`, `make`, `string`, `bool`, `true`, `false`, `nil`). The
    list is authored in `pin.go`. A **toolchain-stable** test checks that every
    identifier token in the §A-CANON texts that `types.Universe` resolves is
    in the authored list (round 4, SC4-1, GO4-6). A toolchain that adds a new
    builtin does not redden it, because the frozen code cannot use a builtin
    it does not mention.
  * There is no `import "unsafe"`, no `import "C"` and no `//go:linkname`
    directive.
  * **No in-package environment mutation (round 4, SEC4-3).** No non-test
    file may call `os.Setenv`, `os.Unsetenv`, `os.Clearenv`, `syscall.Setenv`,
    `syscall.Unsetenv` or `syscall.Clearenv`. `DefaultGitRunner` inherits the
    process environment, so a package `init` could otherwise redirect git
    through `GIT_DIR`, `GIT_INDEX_FILE`, `GIT_WORK_TREE`, `GIT_CONFIG_*` or
    `PATH`. The tracked package has no such call at `9b299c8`. Environment
    set **outside** the package is R-A2.
  * **No non-Go sources (round 4, SEC4-1).**
    `go/build.Context{UseAllFiles: true}.ImportDir` on the directory must
    report **empty** `CgoFiles`, `CFiles`, `CXXFiles`, `MFiles`, `HFiles`,
    `FFiles`, `SFiles`, `SwigFiles`, `SwigCXXFiles` and `SysoFiles`. This
    allowlist replaces an extension list, so `.swig`/`.swigcxx` code
    injection and assembly or object files are all rejected.

  A temp-dir fixture holding only `select.go` passes closure trivially. The
  existing `writeMutatedCopy` fixtures are therefore unaffected.
* **Non-vacuity.** An empty expected set or an empty extracted set is
  rejected (A-T3a).

**Finding-to-fix map (round 3 → rev 6).**

| Finding | Fix | Task |
|---|---|---|
| SEC-1 `:(exclude)` / first-match shadowing pass expected ⊆ resolved | `scopeDataOK`: set **equality**; no `:`-magic pathspec; no duplicate; no prefix-of-prefix; no exact under a prefix; pathspec = prefix + `**`. Token-frozen `scanScope` and `shouldScanRepoPath` also fix the arm order and the first-match loop | A-T3b (freeze), A-T3c (`scopeDataOK`) |
| SEC-2 no `cmd/` non-test probe | New A-T4: a `cmd/x/main.go` probe in the "selection cmd/ coverage (AG-5/D4)" condition, test-first through an injectable helper. The residual text is corrected | A-T4 |
| SEC-3 rest of `selectRepoPaths` unpinned | All of `selectRepoPaths` is token-frozen. Dead code, a reassigned `root`, discarded output, a second runner and extra filtering all change tokens | A-T3b |
| SEC-4 `DefaultGitRunner` body unpinned | `DefaultGitRunner` and `GitRunner` are token-frozen. The git environment and config residual is captured as P-021 stash `D7BF9F74` (R-A2) | A-T3b |
| SEC-5 prefix-function control flow outside the loop | All of `shouldScanRepoPath` is token-frozen. Early return, `goto`, a second loop, `defer`, a falsifying `&&` operand and an `if` returning `false` all change tokens | A-T3c |
| SEC-6, SEC-7 reject-table hardening | Subsumed. The reject tables are now single-predicate token-equality rows (below) | A-T3b, A-T3c |
| GO-1 / SC-2 `S` collides with locals | The `S` rule is withdrawn. Only token equality constrains locals | — |
| GO-2 `make(…, len(scanScope))` | Decided: canonical | §A-CANON |
| GO-3 / SC-1 / CN-4 rule and flag ownership | Explicit flag ownership (above) | A-T3b, A-T3c |
| GO-4 discarded `git` call | Row in the A-T3b reject table, subsumed by the frozen `selectRepoPaths` | A-T3b |
| SC-3 row count vs one-scenario convention | Each table is one predicate (token equality or the closure verdict) over data rows, so it counts as **one** scenario. The convention is unchanged | A-T3b..A-T3d |
| SC-4 / CN-6 AC-A1.8 checklist in commit body | New **AC-A1.9** | A-T1 |
| CN-1 Constitution rows and P-002 deviation format | Rows added under "Constitution Check — Unit A (rev 6)" | plan |
| CN-2 P-002 hard precondition | Discharged by operator decision **D-000-3**: the confirmation is given, narrowly | D-T0, "Dropped from Unit A" |
| CN-3 ALP-1 interval | **IVL-1**, a bounded INV-2 interval with a 030-S merge precondition (below) | plan |

**IVL-1 — bounded INV-2 interval (CN-3).** ALP-1 moves the literals into
`scanScope`, where A-T2's presence collector counts them. From the ALP-1
commit until A-T3b, a `select.go` that keeps `scanScope` intact while narrowing
a consumer is accepted, although today's pin rejects some of those shapes.
That is a temporary INV-2 regression *on the 030-S branch only*.
* **Opens** at the ALP-1 commit.
* **Closes for `select.go`** at A-T3c, when every frozen declaration and
  `scopeDataOK` are live.
* **Closes package-wide** at A-T3d.
* **Merge precondition for the 030-S PR** (carried as AC-A3d.6 and in the
  030-S shipment description; round 4, SC4-7). 033.004-T, 033.005-T,
  033.006-T, 033.007-T and 033.008-T are all `done` before the PR merges. If any of them
  halts to Stage, Ship must not merge 030-S until Stage has re-captured the
  open part of IVL-1 as a P-021 DEFERRED SCOPE EXPANSION entry, and the
  operator has accepted the residual. With the precondition met, `main` never
  observes IVL-1.

**Residuals (recorded; not on this contract surface).** R-A1 and R-A2 exist
identically on `main` at `9b299c8`, and rev 6 does not make either worse
(round 4, LR4-5).
* **R-A1 — runner wiring and downstream dispatch** are not pinned.
  * `Run` and `runRepoScan` (`retiredarch.go`) and
    `register_retired_arch.go` pass the runner and consume `selectRepoPaths`'
    result. They could wrap the runner or filter the result.
  * Inside `select.go`, the unfrozen `scanPath` and `engineForPath` dispatch
    each selected path to an engine. A `scanPath` that returns `nil` for
    `cmd/` paths would silently skip selected files (round 4, SEC4-2/GO4-7).

  Both are captured as **P-021 DEFERRED SCOPE EXPANSION** stash entry
  **`DC921AF6`**.
* **R-A2 — the git process environment and configuration** set **outside**
  this package can do two things:
  * reinterpret pathspecs (`GIT_LITERAL_PATHSPECS`, `GIT_GLOB_PATHSPECS`,
    `GIT_NOGLOB_PATHSPECS`, `GIT_ICASE_PATHSPECS`);
  * redirect `DefaultGitRunner` to a different index, repository or binary
    (`GIT_INDEX_FILE`, `GIT_DIR`, `GIT_WORK_TREE`, `GIT_CONFIG_COUNT`/`KEY`/
    `VALUE`, `PATH`).

  The integrity of `pysem.GitText` and `pysem.SplitLines` is also not pinned.
  In-package environment mutation is closed by package closure (round 4,
  SEC4-3). The rest is captured as **P-021 DEFERRED SCOPE EXPANSION** stash
  entry **`D7BF9F74`**.
* **R-A3 — `_test.go` files** are outside package closure. A test-only `init`
  can mutate `scanScope` in test binaries only, and never in the shipped gate.
  Recorded; no stash entry.
* **R-A4 — out-of-model attacks.** Cross-package `linkname` push, `reflect`
  over an exported pointer, `unsafe` in another package and debugger or
  binary patching are out of the source-text threat model. Recorded; no
  stash entry.
* **The rev-5 "per-arm control flow" and "cross-file `scanScope` mutation"
  residuals are closed.** The frozen `shouldScanRepoPath` closes the first.
  Package closure closes the second. The behavioural self-test is no longer
  the compensating control for those two residuals. It **remains the only
  control** against a coordinated edit of `pin.go` and `select.go` (see
  `scopeDataOK`), which is why A-T4's independent probe is required.

**Function budget convention** (unchanged from rev 5). Budgets count production
functions in `pin.go` that are authored or materially edited, and a
call-site-only rewire of `checkPathspecPin` counts. A canonical-text constant or
a name list is data, not a function. Test helpers count against the
test-scenario budget. Under the plan's fixture-counting convention (D-T6), a
table-driven test counts as **one** scenario.

**Red-phase rule (rev 6).** Each reject table states which prior pin it is red
against.
* **A-T3b rows** keep `scanScope`'s literals present. The presence pin
  (ALP-1 + A-T3a) therefore **accepts** every one of them, so each row is red
  against the prior pin.
* **A-T3c rows** change only `shouldScanRepoPath`, which A-T3b does not
  freeze. They are red against A-T3b's pin.
* **A-T3d rows** live in a second package file. A-T3c's pin reads only
  `select.go`, so they are red against it.
* **`scopeDataOK`'s pure-predicate rows** call a function that does not exist
  until A-T3c. They are **compile-red**, and are declared as such. So is
  AC-A3d.2's universe-coverage test, which needs the authored list (round 4,
  CN4-1).
* **Positive controls are green on arrival, declared (round 4, CN4-1).** They
  are regression guards, not red phases:
  * AC-A3b.2 (comment and whitespace insensitivity, and the live tree);
  * AC-A3c.3's live-tree acceptance;
  * AC-A3d.1's `select.go`-only fixture.
* Each red test is **authored and observed red before** the code that turns
  it green. The two land in the **same commit**, and the red output is
  recorded (P-004) in the harness manifest or the commit body (round 4,
  CN4-4).

**Sequencing.** All edges are `blocks`:

| Edge | Note |
|---|---|
| A-T3a → A-T2 | existing |
| A-T3b → A-T3a | existing |
| A-T3c → A-T3b | existing |
| A-T3d → A-T3c | new |
| A-T4 → A-T2 | new |
| A-T3d → A-T4 | new. A-T3d is the single last task before merge, and its AC-A3d.4 docs cite A-T4's probe (round 4, SC4-3) |

Each task lands as **its own commit after ALP-1**. None is part of ALP-1, and
ALP-1 is not split. It is staged as **one commit**: A-T1 is never committed
alone, and no squash, fixup or rebase is used (CN4-6). INV-5 still declares
exactly one pair. Each red phase is a
new assertion authored inside its task, so INV-5's test-first exclusion
applies.

#### A-T3a — 033.004-T: Make `containsAll` non-vacuous on empty input

Size **XS**, complexity **low**. Posture **test-first**. Depends on A-T2.
Budget: 1 production function (`containsAll`) and 1 scenario. **Unchanged in
substance from rev 5.** Rev 6 relies on it more: `scopeDataOK`'s set equality
calls `containsAll` in both directions.

**Acceptance criteria**
* AC-A3a.1 — **Test-first pure-predicate test, committed.** It asserts
  `containsAll` returns **false** when `wanted` is empty, and also when
  `haystack` is empty. It is red against today's `containsAll`, and its red
  output is recorded (P-004). No package-level variable is mutated.
* AC-A3a.2 — `containsAll`'s doc comment states the empty-input contract, and
  that `scopeDataOK` (A-T3c) relies on it for set equality. AC-A2.2's
  non-emptiness test stays in place and is **not weakened**.
* AC-A3a.3 — Regression gate:
  * `go test ./tools/gatecheck/... -count=1` passes in full;
  * the three `check-retired-architecture.sh` entry points PASS;
  * `golangci-lint run ./...` and `staticcheck ./...` are clean (INV-4);
  * the live tree's verdict is unchanged (INV-1).

#### A-T3b — 033.005-T: Freeze the pathspec surface (closed world + token equality)

Size **M**, complexity **medium**. Posture **test-first**. Depends on A-T3a.
Budget: at most 4 production functions:
* `declTokens`, which tokenises a declaration span or a canonical text;
* the closed-world, build-constraint and identifier-confinement check;
* `frozenDeclsOK`, which compares a named frozen set against its canonical
  texts;
* the `checkPathspecPin` rewire.

The canonical-text constants are data. Test budget: at most 2 scenarios.

**Acceptance criteria**
* AC-A3b.1 — **Test-first reject table (one scenario), authored and observed
  red before the freeze code, and landed in the same commit.** Every row keeps `scanScope`'s literals present, so the
  presence pin accepts it (red-phase rule), and the red output is recorded
  (P-004). Rows:
  * the `git` call in dead code (`if false { … }`);
  * a discarded `git` call (`_, _ = git(…)`) with a second call producing
    `out` (GO-4);
  * `root` reassigned before the call;
  * a second runner or a direct `exec.Command` producing `out`;
  * an extra filter beyond `shouldScanRepoPath`;
  * `pathspecs` resliced (`pathspecs[1:]...`);
  * `DefaultGitRunner` dropping or rewriting its `pathspecs` (SEC-4);
  * an extra `:(exclude)cmd/**` arm in `scanScope` (SEC-1);
  * a duplicated `cmd/` arm placed first (SEC-1);
  * an `init()` in `select.go`;
  * a `//go:build` constraint on `select.go`;
  * an aliased import;
  * `engineForPath` referencing `scanScope`.

  Each row asserts `SelectFound == true` **and** `PathspecOK == false` (and
  `PrefixOK == false` for shared-rule rows), not merely `!OK()`, so a parse
  error cannot satisfy it.
* AC-A3b.2 — **Positive controls (the second scenario; green on arrival,
  declared as regression guards).** One table holds both:
  * comment and whitespace insensitivity: a `select.go` with rewritten doc
    comments and re-indented bodies is **accepted**;
  * the live post-ALP-1 `select.go` is **accepted** by `SelectionPathspecPin`.
* AC-A3b.3 — **Contract.**
  * `PathspecOK` = presence (retained through IVL-1) **AND** the frozen
    pathspec set **AND** the shared rules.
  * A shared-rule violation also clears `PrefixOK`.
  * The canonical texts in `pin.go` are token-equal to §A-CANON and are
    authored independently of `select.go` (H-11).
  * `pin.go:9-15` is **not modified**.
  * `SelectionPathspecPin` keeps its signature.
  * No `go/types`, and no SA1019.
  * Every parse uses `parser.ParseComments`, and directives are detected as
    described in the Pin contract.
  * Evidence records the stdlib facts behind token equality (LR4-4).
  * Evidence also records that no text-scanning gate self-matches the
    canonical text in `pin.go`. `tools/` is outside the `internal/**` and
    `cmd/**` scan scope (round 4, LR4-9).
* AC-A3b.4 — **Regression gate.** AC-A3a.3 holds, and every pre-existing pin
  test passes, including AC-A2.3 and AC-A2.7. Lint stays clean, and no helper
  is left without a caller.

#### A-T3c — 033.006-T: Freeze the prefix surface and add `scopeDataOK`

Size **S**, complexity **medium**. Posture **test-first**. Depends on A-T3b.
Budget: at most 4 production functions:
* `scopeFieldValues`, which extracts `[]armValues` (`pathspec`, `prefix`,
  `exact`, `includeTests`) from the `scanScope` composite;
* `scopeDataOK`;
* the `checkPathspecPin` rewire;
* deletion of the presence helpers left without a caller: A-T2's sibling
  collector, and `collectStringLits` if unused (INV-4).

Adding `shouldScanRepoPath` to the frozen set is data. Test budget: at most
2 scenarios.

**Acceptance criteria**
* AC-A3c.1 — **Test-first prefix reject table (one scenario), committed.**
  Every row changes only `shouldScanRepoPath`, so A-T3b's pin accepts it
  (red-phase rule). Rows (SEC-5):
  * an early `return` before the loop;
  * a `goto`;
  * a second loop;
  * a `defer`;
  * a `&& false` operand;
  * an `if` that returns `false` for `cmd/`;
  * a negated `HasPrefix`;
  * swapped `HasPrefix` arguments;
  * a `continue` keyed on `a.prefix == "cmd/"`;
  * `path` reassigned.

  Each row asserts `GuardFound == true` **and** `PrefixOK == false`.
* AC-A3c.2 — **`scopeDataOK` pure-predicate table (the second scenario;
  compile-red, declared).** It calls `scopeDataOK` directly on `[]armValues`, and
  asserts **false** for each of:
  * an `:(exclude)cmd/**` pathspec;
  * `:!cmd/**`;
  * `:^cmd/**`;
  * a duplicate pathspec;
  * a duplicate prefix;
  * an overlapping prefix (`c` with `cmd/`);
  * an exact value under a prefix (`cmd/x.go`);
  * a pathspec that is not prefix + `**`;
  * a missing expected pathspec;
  * an extra pathspec;
  * `cmd/` with `includeTests: false` (GO4-2);
  * `internal/` with `includeTests: true`;
  * empty sets.

  It asserts **true** for the §A-CANON values.
* AC-A3c.3 — **Contract.**
  * `PrefixOK` = the frozen prefix set **AND** `scopeDataOK` **AND** the
    shared rules.
  * `PathspecOK` additionally requires `scopeDataOK`.
  * **No presence check remains**, and every presence helper left without a
    caller is deleted.
  * Positive control: the live `select.go` is accepted. This reuses
    AC-A3b.2's positive-control table with one added row, so it is not a third
    scenario (SC4-7).
* AC-A3c.4 — AC-A3b.3 and AC-A3b.4 hold again at this task's boundary.

#### A-T3d — 033.007-T (NEW, rev 6): Package closure for the frozen surface

Size **S**, complexity **medium**. Posture **test-first**. Depends on A-T3c and
A-T4. Budget: at most 3 production functions:
* the package-file enumeration and parse;
* the per-file closure rule check;
* the `checkPathspecPin` rewire.

The universe-name list and `includeTestsPinPrefixes` are data. The
`go/build` `ImportDir` call sits inside the enumeration function. Test budget:
at most 2 scenarios.

**Acceptance criteria**
* AC-A3d.1 — **Test-first closure reject table (one scenario), committed.**
  Each row writes the live `select.go` plus **one** extra non-test `.go` file
  into a temp dir. A-T3c's pin reads only `select.go`, so it accepts every row
  (red-phase rule), and the red output is recorded. Rows:
  * a second file with `func init() { scanScope = scanScope[1:] }`;
  * a second file that mentions `scanArm`;
  * a package-level `func append(…)`;
  * a package-level `var len = …`;
  * `type string = …`;
  * a package-level `var nil` or `true`;
  * `import "unsafe"`;
  * `import "C"`;
  * a `//go:linkname` directive;
  * a second file calling `os.Setenv("GIT_INDEX_FILE", …)` from `init`
    (SEC4-3);
  * a `.s` file;
  * a `.syso` file;
  * a `.swig` file (SEC4-1);
  * a second file that declares a closed-world name (belt and braces);
  * an unparseable second file;
  * an unreadable entry, made portably as a **directory** named `x.go`
    (GO4-5/SC4-6).

  Each row asserts `PathspecOK == false` **and** `PrefixOK == false`. Every
  row except the unparseable and unreadable ones **also** asserts
  `SelectFound && GuardFound`, and that the extra file parses without error.
  A row therefore cannot pass for the wrong reason (LR4-2). A temp-dir
  fixture holding only `select.go` is accepted (green on arrival, declared).
* AC-A3d.2 — **Universe coverage (the second scenario; compile-red,
  declared).** A test tokenises the §A-CANON texts. It asserts that every
  identifier token that `types.Universe.Lookup` resolves is in `pin.go`'s
  authored universe-name list. The check is toolchain-stable and is not an
  equality with `types.Universe.Names()` (SC4-1, GO4-6). `go/types` is used in
  the **test only**.
* AC-A3d.3 — **Contract.** Package closure scans every non-`_test.go` `.go`
  file in `select.go`'s directory. Every rule fails closed and clears both
  flags. The live package is accepted.
* AC-A3d.4 — **Docs.** The `pin.go` header doc (`:1-7`, as A-T2 rewrites it)
  must:
  * cite **D-030-4** and **D-030-6**;
  * name the frozen declaration sets and their flag ownership;
  * describe the closed world, `scopeDataOK` and package closure;
  * state residuals R-A1..R-A4 and cite stash `DC921AF6` and `D7BF9F74`;
  * record the round-4 refinements: the `ImportDir` non-Go-sources
    allowlist, the environment-mutation rule, the `includeTests` rule and the
    toolchain-stable universe check;
  * cite A-T4's independent `cmd/x/main.go` probe as the control for a
    coordinated `pin.go` + `select.go` edit.

  (This criterion moved here from rev-5 AC-A3c.4.)
* AC-A3d.5 — The AC-A3b.4 regression gate holds.
* AC-A3d.6 — **IVL-1 merge precondition.** This is the last Unit A task
  before merge. At its completion, Ship confirms 033.004-T..033.008-T are all
  `done` before 030-S's PR may merge. If any one halted, Ship does not merge
  030-S until Stage has re-captured the open IVL-1 part as a P-021 entry and
  the operator has accepted it.

#### A-T4 — 033.008-T (NEW, rev 6): Probe non-test `cmd/` Go files in the selection self-test (SEC-2)

Size **XS**, complexity **low**. Posture **test-first**. Depends on A-T2.
Files: `selftest_selection.go` and `selftest_selection_test.go`. Budget:
1 production function (an injectable helper, for example
`cmdProbesHold(pred func(string) bool) bool`, plus its call-site in the
existing condition) and 1 scenario.

**Acceptance criteria**
* AC-A4.1 — **Test-first, committed.** A pure-predicate test calls the
  helper with a predicate that **drops non-test `cmd/` Go files**, for example
  `func(p string) bool { return strings.HasSuffix(p, "_test.go") ||
  strings.Contains(p, "/testdata/") }`, and expects **false**.
  * The test also asserts **true** for `shouldScanRepoPath` itself.
  * Its red phase is recorded against a helper that carries only the two
    pre-existing probes (`cmd/x/y_test.go`, `cmd/x/testdata/z.go`). That helper
    returns **true** for the dropping predicate, so the red is genuine and not
    merely a compile failure (P-004).
* AC-A4.2 — The "selection cmd/ coverage (AG-5/D4)" condition adds the probe
  `cmd/x/main.go`, whose expected result (included) is a literal in the
  self-test, independent of `scanScope`. The condition's **name, order and
  success text are unchanged**, so `TestRunRepoSelectionSelfTest_AssertionNamesAndOrder_MatchGolden`
  passes unmodified. The failure text may add the probe value.
* AC-A4.3 — The AC-A3a.3 regression gate holds.

**Stop conditions (H-2 A, rev 6).**
* If A-T1 cannot write `select.go` token-equal to §A-CANON while staying lint
  clean (INV-4) and verdict-identical (INV-1), **HALT and return to Stage**.
  Do **not** adjust the canonical text in `pin.go` to match a different shape.
  That is a re-plan, and it belongs to Stage.
* If any A-T3 task or A-T4 would exceed its declared function or test budget,
  **HALT and return to Stage** for a re-split.
* Do **not** fall back to presence, do **not** edit `select.go` in any A-T3
  task, and do **not** weaken AC-A2.2 or AC-A2.3.

### Dropped from Unit A

* **033.003-T — DROPPED, closed as a verified no-op (D-000-2, D-000-3).** The
  Python `inspect.getsource()` pin became a stronger `go/parser` AST pin during
  the port (`pin.go:56`, `:156-162`). The asymmetry assertions already exist
  (`selftest_selection.go:185-197`, `:256-276`). Re-implementing it would be
  green on arrival. Its surviving intent is carried by AC-A2.1 through AC-A2.3,
  and A-T3a..A-T3d and A-T4 strengthen it.

  > **Operator authorisation (D-000-2, confirmed by D-000-3).** Ship may close
  > 033.003-T **inside 030-S** as a **verified no-op**. The path is
  > `queued → active → done`, and Ship reads back the status after each
  > transition. The pre-task-completion gate evidence is that the work is
  > already satisfied: `pin.go` supersedes it. Cite it **by symbol**
  > (`checkPathspecPin`, `SelectionPathspecPin`) **and by the commit SHA at
  > closure**, not by line number.
  > * **Timing (rev 6).** Close it after A-T3d (033.007-T) lands, so the
  >   evidence cites the strongest pin. If any A-T3 task halts to Stage, close
  >   it citing the strongest pin that did land. Either is valid evidence, so
  >   no dependency edge is added and the item can never be stranded.
  > * **Disclosure.** The closure artifact must state that the task was
  >   **satisfied by prior work**, with no new code.
  > * **P-002 skip — AUTHORISED (operator decision D-000-3, 2026-10-01).** A
  >   no-op adds no code and has no red phase. The operator explicitly
  >   authorises Ship to record a **P-002 test-first skip** for this item as
  >   part of its verified no-op closure (`queued → active → done`, with gate
  >   evidence), citing D-000-3.
  >   * The authorisation is **item-scoped**: it covers 033.003-T here and
  >     034.001-T in 031-S, and nothing else.
  >   * It is **not** a general P-002 waiver.
  >   * It does **not** extend to any task that adds or changes code,
  >     including every other 030-S task.
  >   * The CN-2 hard precondition is **discharged**.
  > * `.backlogit/hooks.yaml` is **not** changed.

**Known residual — closed by A-T3a..A-T3d and A-T4 (rev 6).** Stash
`DD0BB60F` recorded that `checkPathspecPin` only verified that the pin literals
occur *somewhere* in the target body (`pin.go:76`), and not that they reach the
`git` call. Rev 5 harvested it as A-T3a/b/c, and DD0BB60F is archived. Rev 6
closes the residual by freezing the whole selection surface. The rev-5 per-arm
control-flow and cross-file residuals are closed. R-A1..R-A4 remain, and two of
them are captured as P-021 stash entries.

---

## Unit B — 032-S / 035-F: Repair unignore-regression checker ref resolution

**Retarget.** Python `root_gitignore_text_at` → `rootGitignoreTextAt`
(`tools/gatecheck/internal/unignore/git.go:84`).

**Measured correction, restated for Go.** The checker does **not** currently fail
open in practice: `runDifferentialCheck` fails closed on an unresolvable ref
(`checks.go:90-96`) **before** the use sites at `checks.go:124` and `:128`. The
defect is **latent defence-in-depth**, not a live escape. It is worth fixing
because the guard and the use site are separated and a future refactor could
reorder them.

**What the port made worse.** `rootGitignoreTextAt` returns `"", nil` on **any**
non-zero `git show` exit (`git.go:100-104`), discarding stderr
(`stdout, _, err := …`) and collapsing *invalid ref* and *legitimately absent
`.gitignore`* into one indistinguishable result.

### B-T1 — Invert and extend the direct baseline-resolution tests

Size **S**, complexity **medium**. Posture: **test-first — this task supplies the
red and is authored first.** Single file: `git_test.go`.

Tests must drive `rootGitignoreTextAt` **directly**, not through the scenario
harness — the end-to-end path is short-circuited by the upstream guard and cannot
reach the defect. A scenario-level test would be green-on-arrival and certify
nothing.

**This task is partly a test *correction*.** `git_test.go:103-121`
(`TestRootGitignoreTextAt_InvalidRef_DegradesToEmpty`) currently asserts
`err == nil` and `text == ""` for an invalid ref — i.e. it **pins the defect as
correct** — and its comment at `:103-107` names 035-F. It must be inverted and
renamed, and that comment is **owned by this task**, not by B-T3.

Required cases: (a) a **valid non-HEAD ref** whose `.gitignore` is **absent** →
`("", nil)`; (b) an **invalid/nonexistent ref** → distinct non-nil error.
Case (a) is **not optional**: without it, B-T2 could regress the legitimate
absent-file path into an error undetected.

**Acceptance criteria**
* AC-B1.1 — Both cases drive `rootGitignoreTextAt` directly, not via the scenario
  harness.
* AC-B1.2 — The inverted `:103-121` assertion and case (b) both **FAIL against the
  pre-change function**. Leaving the original degrade-to-empty assertion in place
  fails this task.
* AC-B1.3 — The stale characterisation comment at `git_test.go:103-107` is
  corrected as part of this task.
* AC-B1.4 — Pre-existing coverage at `git_test.go:35-43` (HEAD/absent → `""`) and
  `:58-101` (valid ref/present) still passes unmodified.

### B-T2 — Fail closed on an unresolvable base ref at the point of use

Size **S**, complexity **medium**. Depends on B-T1 (turns its red green). Single
file: `git.go`.

Make `rootGitignoreTextAt` distinguish its two failure modes: a valid ref whose
`.gitignore` is absent returns `("", nil)`; an unresolvable ref returns a
**distinct non-nil error**. This requires inspecting `git show`'s stderr/exit
shape — currently discarded at `git.go:99` — not merely its exit status, which is
why this is S/medium rather than XS.

**Acceptance criteria**
* AC-B2.1 — An unresolvable ref produces a distinct non-nil error at the point of
  use, independent of the upstream `checks.go:90-96` guard.
* AC-B2.2 — A valid ref whose `.gitignore` is absent still returns `("", nil)`.
  The legitimate absent-file path is **not** converted into an error.
* AC-B2.3 — The HEAD path (`git.go:85-98`) is unchanged: absent `.gitignore` on
  disk still yields `""`.
* AC-B2.4 — Every B-T1 test passes; existing merge-blocking behaviour is unchanged
  for resolvable refs (INV-1).

> **Declared interim inconsistency.** B-T2 changes the failure semantics but
> leaves the doc comments at `git.go:75-83` and `:101-103` still asserting the old
> "both degrade to an empty baseline" contract. They are knowingly false for
> exactly one task boundary and are corrected by B-T3, which owns them. This is a
> comment-only lag, not an INV-5 atomic-landing pair: no gate is red between B-T2
> and B-T3.

### B-T3 — Correct the stale baseline-semantics comments in `git.go`

Size **XS**, complexity **low**. Documentation-only. Single file: `git.go`.

Correct `git.go:75-83` and `git.go:101-103`, which both still state that an
invalid ref and an absent `.gitignore` "both degrade to an empty baseline".

**Scope boundary:** `git_test.go:103-107` is **not** in scope here — B-T1 owns it.
This task touches `git.go` comments only.

**Acceptance criteria**
* AC-B3.1 — Each corrected comment names the actual mechanism (`git show` of the
  ref-qualified path) and the corrected failure mode.
* AC-B3.2 — No executable line in `git.go` is modified by this task.

---

## Unit C — NEW: gatecheck CLI containment hardening

Harvested from stash **7223218F** (high), **50E6F22C** (medium), **44F8CC48** (low).

**The live gap.** `parseRoot` (`main.go:74-102`) scans the **whole** forwarded
argument list for `--root`/`--root=` with **last-occurrence-wins** semantics
(`root = args[i+1]`, no `break`). `gatecheck-run.sh:66` appends the trusted
`--root ${ROOT}` **ahead of** whatever a wrapper forwards. Therefore any wrapper
forwarding more than a single fixed non-flag argument must carry its own
`--root` denylist.

Two wrappers do, each hand-rolled differently:
`check-retired-architecture.sh:231-249` (allowlists the first arg only) and
`check-unignore-regression.sh:94-101` (per-argument denylist loop).
**`check-write-path-precondition.sh:84` forwards raw `"$@"` with no guard at
all** — a real override gap today.

Verified safe to change: all four `register_*.go` subcommands forward args and
none consumes `--root`; no wrapper passes two `--root`s; `main_test.go` contains
no test pinning last-wins semantics.

### C-T1 — Make the trusted `--root` authoritative in `parseRoot`

Size **S**, complexity **medium**. Posture: test-first. `main.go` + `main_test.go`.

Change `parseRoot` so the **first** `--root` occurrence wins and any **subsequent**
`--root`/`--root=` is a hard error (exit 1, named diagnostic) rather than a silent
override. Because `gatecheck-run.sh:66` always prepends the trusted root,
first-wins makes it authoritative for **all four** engines at once, removing the
need for every wrapper to carry its own copy of the guard. This is the central
cross-cutting fix (option (a) of 7223218F); **C-T3**'s wrapper denylist is
defence-in-depth, not the primary control.

**Acceptance criteria**
* AC-C1.1 — A caller-supplied `--root` or `--root=` appearing **after** the
  trusted one causes a non-zero exit with a named diagnostic; it never overrides.
* AC-C1.2 — The trusted root prepended by `gatecheck_invoke` is the one used, for
  all four subcommands.
* AC-C1.3 — Existing single-`--root` invocations are unaffected (INV-1),
  demonstrated by the **actual** self-test entry points:
  `scripts/check-retired-architecture.sh --self-test`,
  `scripts/check-write-path-precondition.sh --self-test`,
  `scripts/check-unignore-regression.sh --self-test`, and
  `scripts/check-merge-strategy.sh --self-test`. (The merge-strategy self-test is
  implemented in the **wrapper** `run_self_test` at
  `check-merge-strategy.sh:171`; `mergestrategy.Run` itself has no self-test mode.)
* AC-C1.4 — Table-driven tests in `main_test.go` cover: trusted-only, trusted +
  trailing `--root`, trusted + trailing `--root=`, and `--root` with no value.

### C-T2 — Add the write-path wrapper guard test

Size **S**, complexity **low**. Go test only. Posture: test-first. Depends on C-T1.

Add `tools/gatecheck/check_write_path_wrapper_test.go` mirroring
`check_unignore_wrapper_test.go`, including that file's reliance on the guard
firing before `gatecheck_build` (`check_unignore_wrapper_test.go:26-27`).

> **Ordering (plan-review round 2, finding P3-1).** This test task is sequenced
> **before** the shell guard it covers, so Unit C keeps the same test-first
> posture C-T1 adopts: C-T2 supplies the red, C-T3 turns it green. The earlier
> guard-then-test ordering left the new guard unverified for one boundary.
>
> Split from C-T3 because mirroring the unignore test means two test functions
> plus a path helper and a trace helper. The unignore mirror has **10 scenarios**
> (5 rejection + 5 acceptance); write-path's legitimate CLI surface is narrower
> (`no args`, `--self-test`, `--self-test-integrity` — unknown flags already exit
> 2 via `writepath.Run`), so the realistic mirror is ~5 rejection + 3 acceptance.
> Either way, bundled with the shell edit this would exceed the 2-hour rule and
> mix the shell and Go-test skill domains, so the split stands.

**Acceptance criteria**
* AC-C2.1 — Rejection cases cover `--root`, `--root=…`, and a `--root` following
  a legitimate flag; acceptance cases cover `--self-test` and
  `--self-test-integrity`.
* AC-C2.2 — Each rejection subtest **asserts** that `gatecheck_build` was not
  reached, using the same trace mechanism as
  `check_unignore_wrapper_test.go:103` (`bash -x` plus a trace scan). The
  assertion is authored here and **goes green at C-T3** — pre-C-T3 the wrapper
  has no guard, so every rejection case necessarily reaches
  `gatecheck_build` (`check-write-path-precondition.sh:78`). This AC owns the
  assertion's *existence and shape*; AC-C3.2 owns its *satisfaction*.
* AC-C2.3 — The rejection cases **fail** against the pre-C-T3 wrapper. A suite
  that is green on arrival fails this task.
* AC-C2.4 — The file is gofmt and goimports clean (INV-4).

> **Red phase is bounded and declared.** C-T2's rejection cases are red until
> C-T3. Per INV-5's test-first exclusion this is **not** an atomic landing pair:
> the red assertions are newly authored and did not exist on the tracked tree at
> `9b299c8`, and `scripts/check-write-path-precondition.sh` itself stays PASS
> throughout. C-T2 and C-T3 may land as separate commits.

### C-T3 — Add the per-argument `--root` denylist to the write-path wrapper

Size **XS**, complexity **trivial**. Shell only. Depends on C-T2.

Add to `scripts/check-write-path-precondition.sh` the per-argument denylist loop
used at `check-unignore-regression.sh:94-101`.

**Placement is load-bearing:** the guard must run **before** `gatecheck_build`
(currently `check-write-path-precondition.sh:78`), mirroring the unignore
analogue which guards at `:94` before sourcing and building. A post-build
placement makes every rejection test pay a full `go build`.

**Acceptance criteria**
* AC-C3.1 — Every forwarded argument is checked against a literal
  `--root`/`--root=*` denylist before pass-through; a match exits non-zero with
  the same `::error::unrecognized argument:` diagnostic shape the unignore wrapper
  uses.
* AC-C3.2 — The guard executes **before** `gatecheck_build`.
* AC-C3.3 — Legitimate invocations (no args, `--self-test`,
  `--self-test-integrity`) are unaffected.
* AC-C3.4 — Every C-T2 scenario now passes.

### C-T4 — Correct the wrapper rationale comments superseded by C-T1

Size **XS**, complexity **trivial**. Documentation-only. Depends on C-T1.

C-T1 makes five live rationale comments factually wrong; each currently
documents the superseded last-wins security model:
`check-unignore-regression.sh:70-72`, `check-retired-architecture.sh:210-211`
(the "LAST occurrence wins" sentence) and `:224-227` (the "`$2` onward ignored"
rationale that depends on it), `check_unignore_wrapper_test.go:14-20`, and
`check_retired_architecture_wrapper_test.go:14-20`.

**Acceptance criteria**
* AC-C4.1 — All five comments describe the post-C-T1 model: the trusted root is
  authoritative because the **first** occurrence wins, and the wrapper denylists
  are retained as defence-in-depth.
* AC-C4.2 — No executable line is modified by this task.

### C-T5 — Record gatecheck's CLI containment posture as explicit residuals

Size **S**, complexity **low**. Documentation-only. Independent.

Document — do not silently inherit — the three Principle III items from 50E6F22C,
**all three confirmed faithful parity with the retired Python original, not
regressions introduced by the port**:

1. `mergestrategy.Run` accepts `root` and discards it (`evaluate.go:192`
   `_ = root`), reading the caller-supplied path via `os.ReadFile` (`:59`) with no
   containment check. Required by the live-transport design: the wrapper's
   temp-file payload is created in the OS temp directory, outside the repo root,
   by design. Matches the original `evaluate_json`'s `open(sys.argv[1])`.
2. `unignore`'s self-test creates scenario/scratch repos via `os.MkdirTemp("", …)`
   outside root. Matches the original `tempfile.TemporaryDirectory()`.
3. `mergestrategy` silently evaluates only `args[0]` (`:197`) and ignores extra
   positional arguments. Matches the original's unconditional `sys.argv[1]`.

**Acceptance criteria**
* AC-C5.1 — All three are recorded in the `mergestrategy`/`unignore` package docs
  with their parity justification and the evidence that each matches the retired
  Python.
* AC-C5.2 — The record states explicitly that these are **accepted, scoped
  Principle III exceptions**, not undiscovered holes.
* AC-C5.3 — **No behaviour is changed by this task.** In particular item (3) is
  **not** fixed here. The distinguishing factor is **INV-1's explicit-AC
  carve-out**, not exit-code or trigger novelty: C-T1 also adds a new trigger to
  a currently-passing invocation shape, but it does so under an acceptance
  criterion (AC-C1.1) that names the change. Item (3) has no such authorisation,
  so changing it would be an unauthorised INV-1 deviation. (For completeness:
  `mergestrategy.Run` already returns 2 for usage at `evaluate.go:193-195`, so
  the fix would reuse an existing exit code — the exit code was never the
  obstacle.) Recorded as a candidate ED entry for a future revision.

### C-T6 — Correct the stale `.gitignore` gate-engine comment

Harvested from 44F8CC48. Size **XS**, complexity **trivial**. Independent.

`.gitignore:126-127` still reads "Python bytecode caches written when the
extracted gate engines under `scripts/lib/` are imported or executed (032-F,
shipment 029-S)". Those engines were retired in M4 (037-S). Append a clarifying
line noting the rule is now defensive/historical.

`.gitignore` is an **append-only gated file** and a different contract surface
from the Go code — hence a comment **addition**, never a rewrite or deletion.

**Acceptance criteria**
* AC-C6.1 — A comment line is **appended** near the existing block noting the
  `scripts/lib/` engines were retired in M4 (037-S) and the rule is now
  defensive/historical.
* AC-C6.2 — The existing `__pycache__/` and `*.pyc` rules are **not** removed or
  reordered — removing an ignore rule is exactly what
  `check-unignore-regression.sh` blocks.
* AC-C6.3 — Both gates pass **using their real CI entry points** (neither script
  may be invoked bare — both fail closed with exit 1 when given no ref
  arguments): `scripts/check-gitignore-append-only.sh --self-test` plus the
  `--base-ref`/`--head-ref` form used at `ci.yml:464`, and
  `scripts/check-unignore-regression.sh --self-test` plus the `--base-ref`/
  `--head-ref` form used at `ci.yml:486`.

---

## Unit D — 031-S / 034-F (revised): Harden the write-path gate — masked-text increment

**Decision.** D-031-2 (operator, 2026-10-01): **split 031-S.** This unit keeps the
Option A work that **carries over** to a later `go/ast` engine — the D-1/D-2
mechanism, the selector set, the residual record and, above all, the fixture
corpus, which becomes Unit E's **verdict-parity regression corpus**. The parts
that are hard as text scanning and that `go/ast` would replace (named import
aliases, `NewRoot` receiver tracking) move to Unit E.

**Rev-4 plan-review amendments (deliberation D-031-3).** Round 1 of the rev-4
review returned **FAIL** (1 P0, 6 P1). Stage amended the Python-era mechanism
decisions for the Go engine:

* **D-1′** keeps the `SplitLines` line loop plus a byte cursor, so line parity
  and the `retiredarch` scan-loop pin both hold. It tracks `()[]{}` and
  enumerates every occurrence.
* **D-2′** adds exact creation-disposition and flags tokens, a bare-operand rule
  and per-occurrence evaluation.
* **D-T1** and **D-T5** are split under the 2-hour rule, into new tasks 034.010-T
  and 034.011-T.

Neither D-1′ nor D-2′ widens what the gate allows. Each only narrows the
allowance or preserves parity.

**Retarget.** `scripts/check-write-path-precondition.sh` (Python body, retired) →
`tools/gatecheck/internal/writepath/writepath.go`. Fixtures stay at
`scripts/testdata/writepath/*.go` (`runFixtureSelfTest`, `writepath.go:220-221`).

**Facts re-measured at `9b299c8` that shape every task:**

* `scanText` (`writepath.go:112-124`) is line-based over `pysem.SplitLines`;
  `findSelector` (`:89-102`) is `strings.Index` plus `pysem` boundary checks.
  Findings are emitted **per (line, selector) pair in `Selectors` order** and a
  selector hit **at most once per line** — both are pinned by the golden.
* `pysem.SplitLines` uses **Python `splitlines` boundaries** (`\n`, `\r`, `\r\n`
  **and** `\v`, `\f`, `\x1c`–`\x1e`, `\x85`, `\u2028`, `\u2029`). D-1's
  "count `\n` before the offset" is therefore **not** automatically identical to
  the current line numbering. D-1′ (D-T1) closes this hazard by construction:
  it keeps the `SplitLines` loop and adds a byte cursor.
* `retiredarch/writepath_mask_test.go` **pins writepath's scan-loop shape**
  (`:15-29`, `:265-302`):
  * `scanText` must range over `pysem.SplitLines` of its second parameter in a
    top-level statement;
  * `scanFile` must pass a direct `gomask.MaskGoNonCode` call;
  * no other non-test writepath source may reference `scanText`.

  The original D-1 ("`scanText` stops iterating lines") would break this pin.
  This is the reason for D-1′.
* `writepath_golden.json` pins **three** things that this unit changes:
  the `selectors` array (`TestSelectors_MatchesGolden`), per-fixture
  `fixture_findings` rows, and the `self_test` / `self_test_integrity` stream
  captures, whose stdout carries one `PASS` line per fixture. **Every task that
  adds a selector or a fixture must update the golden in the same commit** or
  `go test ./tools/gatecheck/...` goes red (INV-4).
* The target call `internal/pathsafe/reparse_windows.go:164-172` still has
  **exactly 7 arguments with arg index 1 the literal `0`** (`:166`, followed by
  a comment the masker blanks) and a trailing comma before `)`. The four
  comment-only `syscall.CreateFile` mentions are at `:24`, `:28`, `:33`, `:145`.
* None of `syscall.CreateFile` (as a code reference other than `:164`),
  `syscall.Write`, `os.OpenRoot`, `os.Root`, `io.CopyN`, `io.CopyBuffer` occurs in
  any of the **17** non-test `.go` files under `internal/**`/`cmd/**`.
  `io.CopyN` is **not** caught by the existing `io.Copy` selector
  (`FollowedByWord` rejects the trailing `N`), confirming B72E9715.
* The only named import alias in scope is `copilot` →
  `github.com/github/copilot-sdk/go` (`internal/copilotprobe/{client,fixture,permission}.go`),
  which is not write-capable. There are **zero** production `.Resolve(` callers.
* No atomic landing pair is needed in this unit (INV-5 unchanged): the predicate
  (D-T2) lands **before** the selector that needs it (D-T3), so the repo scan
  never sees an unallowed `syscall.CreateFile`. Every intermediate task state
  stays inside the single 031-S shipment PR, so the scan-level positive control
  arriving one task after D-T3 (in D-T5a) never reaches `main` separately.

### D-T0 — 034.001-T: SATISFIED by the port (no executable scope)

Handled exactly as 033.003-T was (Unit A, "Dropped from Unit A"). The spike at
`docs/decisions/2026-09-19-intercom-go-call-extent-extraction-and-allowance-predicate-spike.md`
exists and records D-1..D-4; this revision re-anchors D-1/D-2 onto the Go engine.
The task is retitled `SATISFIED (post-M4 Go re-plan) …`, keeps `queued` status and
its 031-S membership, and is **removed from the live dependency chain** (edge
`034.002-T → 034.001-T` deleted) so a no-op item cannot gate real work.

> **Lifecycle constraint (same as 033.003-T).** `.backlogit/hooks.yaml` allows
> only `queued → active|blocked`; there is no `queued → archived|rejected`
> transition, so Stage cannot give the item a terminal state. Removal from the
> manifest **is** possible (`shipment return-blocked`, see Unit E) but would leave
> a permanently queued orphan; retention keeps it visible in the manifest Ship
> already reconciles. **Named prerequisite for 031-S closure** (shared with
> 033.003-T in 030-S): before Ship's closure classification, the operator either
> authorises Ship to close it as an explicit no-op item (counted in the
> two-set gate as "satisfied-by-prior-work", not as delivered scope), or gives
> both items a terminal state. If the first option is taken, the closure
> artifact records it as an **explicit, disclosed deviation** from the
> two-set gate computation (compound lesson
> `2026-09-30-shipment-reconcile-skill-bypass-manual-substitution-risk.md`).
> A narrative compliance claim does not satisfy this.

> **Rev 5 — operator decision D-000-2 (authorised).** The operator took the
> first option. **Ship may close 034.001-T inside 031-S as a verified no-op**
> by the normal path `queued → active → done`. No `hooks.yaml` change is
> needed or made. The pre-task-completion gate (`evidence_required: true`)
> evidence is that the work is already satisfied:
> * the spike document
>   `docs/decisions/2026-09-19-intercom-go-call-extent-extraction-and-allowance-predicate-spike.md`
>   (D-1..D-4);
> * D-031-3, the operator-acknowledged (D-031-4) adaptation of D-1/D-2 to the
>   Go engine as D-1′/D-2′.
>
> The task adds no code. The closure artifact counts it as
> **satisfied-by-prior-work**, not as delivered scope, and discloses it as such.
> A no-op has no red phase, so Ship records the `harness-ready` precondition
> (P-002/P-004) at claim as a disclosed `skip_policy: P-002`, scoped to this
> item and citing D-000-2.
> 033.003-T in 030-S is authorised on the same terms (Unit A, "Dropped from
> Unit A").

> **Rev 6 — operator decision D-000-3 (P-002 skip authorised, narrowly).** The
> operator **confirms** that Ship may record a **P-002 test-first skip** for
> 034.001-T, citing D-000-3. The skip is recorded as part of the verified no-op
> closure (`queued → active → done`, with pre-task-completion gate evidence).
> * Scope is **034.001-T (031-S) and 033.003-T (030-S) only**.
> * It is **not** a general P-002 waiver.
> * It does **not** cover any task in 031-S, 030-S or elsewhere that adds or
>   changes code.
> * The CN-2 hard precondition (AC-D0.4) is **discharged**.

### D-T1a — 034.010-T (NEW): Commit a frozen differential oracle for detection parity

Size **S**, complexity **low**. Posture: characterization-first (declared
**green on arrival** — it pins today's behaviour). One new file,
`tools/gatecheck/internal/writepath/writepath_oracle_test.go`. No dependency.

Split out of D-T1 at plan-review rev-4 (2-hour rule; P1 oracle-drift finding).
Copy the pre-change `scanText` **and** `findSelector` (`writepath.go:89-124`)
verbatim into the test file under test-only names (never the identifier
`scanText` — `retiredarch`'s `TestWritePathScanFileConsumesCanonicalMask` only
inspects non-test files, but the names must not collide), together with a
**frozen copy of the original 20 selectors**. The oracle compares the frozen
legacy scan against production `scanText` **restricted to findings that name
one of the frozen 20 selectors**.

The restriction is an **exact set lookup** on the selector parsed from the
finding text, compared against `pysem.Repr(sel)` for each frozen selector. The
token is parsed **from the right** (`TrimSuffix(" found")`, then the text after
the `LastIndex` of `"write primitive "`), because the path is free text. It is never a substring
test, because `'io.CopyN'` and `'io.CopyBuffer'` contain `io.Copy`.

D-T3 only **appends** selectors, and the D-2′ allowance only ever applies to
`syscall.CreateFile`, which is not in the frozen 20. The restriction therefore
keeps the oracle valid, **unedited**, through all of Unit D.

**Oracle lifecycle (rev-4 round-2 P1 fix).** The oracle is a **Unit D**
masked-text harness. Its **legacy side and expectations are frozen for its
whole lifetime**. Exactly three things may be adapted, **exactly once**, in
E-T2, by AC-E2.6:
* the production side;
* the input list;
* the expectations for **unparseable** inputs, which may only move to
  fail-closed. That is the only authorised edit, and it
can only make the oracle stricter. No other task may edit the oracle (H-3).

**Acceptance criteria**
* AC-D1a.1 — The oracle runs over:
  * every `scripts/testdata/writepath/*.go` fixture;
  * every `filebased` golden input;
  * every tracked non-test `internal/**`/`cmd/**` file.

  Each input is masked exactly as `scanFile` masks it. The restricted
  production findings must equal the oracle's findings in text, line and order.
  An input that `pysem.ReadText` rejects (invalid UTF-8) is not a parity input:
  the oracle asserts that **both** sides error.
* AC-D1a.2 — Hand-built inputs pin line numbering to the `pysem.SplitLines`
  boundary set:
  * one input contains `\f`, `\v` and `\u2028` in **code position**, with a
    selector hit on each following line;
  * one input contains the same boundary runes **inside a comment and inside an
    interpreted string**, followed by a selector-hit line.

  The masker turns in-comment and in-string boundaries into spaces, so the
  legacy line number of that hit is the **masked** line number. This is a parity
  hazard Unit E's G-3 must reproduce. Every hand-built input except the
  code-position one is a **complete, parseable Go file**, with a `package`
  clause.
* AC-D1a.3 — Green on arrival at its own parent commit; no production file is
  modified; lint clean.

### D-T1 — 034.002-T: Call-extent extraction over masked text (D-1′)

Size **S**, complexity **medium** (already de-risked by the 2026-09-19 spike's
D-1 feasibility proof). Posture: characterization-first.
`writepath.go` + `writepath_test.go`. Depends on D-T1a.

Implement **D-1′** (D-1 as amended by deliberation D-031-3; the extent
algorithm is unchanged, the scan loop is not):

* **Keep the line loop.** `scanText` keeps its signature and its single
  top-level `range pysem.SplitLines(maskedText)`. It additionally maintains a
  **byte cursor** into the whole masked text: each line is exactly
  `maskedText[cur:cur+len(line)]`, and the cursor then advances past exactly
  one boundary (`\r\n` as two bytes, otherwise the boundary rune's UTF-8 width).
  This is why D-1's "count `\n` before the offset" is dropped.

  The cursor **fails closed**:
  * it asserts `maskedText[cur:cur+len(line)] == line`;
  * it bounds-checks the boundary decode on the last line.

  On any mismatch, `scanText` **disables the allowance for the rest of that
  text**, so every occurrence is reported. This fails closed as a false
  positive, with no signature change. `scanText` returns `[]string` and has no
  error channel, and the oracle and `TestFilebased_*` call it single-valued.

  The mismatch is unreachable through `scanFile`, because:
  * `ReadText` rejects invalid UTF-8 and reports it as ED-2 `::error::` via
    `errorLine`;
  * `MaskGoNonCode` emits only input runes, `' '` and `'\n'`.

  A direct-call test with invalid UTF-8 pins the fallback. Line numbers stay **identical by construction**, and the `retiredarch`
  architectural pin
  (`retiredarch/writepath_mask_test.go:15-29,265-302`: `scanText` must range
  over `pysem.SplitLines` of its second parameter at top level, `scanFile` must
  pass `gomask.MaskGoNonCode` directly, and nothing else may reference
  `scanText`) passes **unmodified**.
* **Enumerate every occurrence.** For each `(line, selector)` it enumerates
  **every** non-word-bounded occurrence of the selector on that line, not only the
  first. The legacy `findSelector` stops at the first match.
* **Extent extraction** for each occurrence, at its absolute offset, over the
  whole masked text:
  * skip whitespace (newlines included);
  * require `(`;
  * walk to the balanced `)` while tracking a **bracket stack over `()`, `[]` and `{}`**. A mismatched closer is **undecidable**, so is reaching EOF;
  * split the interior at commas where the stack depth is exactly 1, i.e. inside the call's own parentheses only.

  A selector followed by anything but `(` is a **non-call reference**.
* **No verdict change.** Per D-4, the extent result **cannot change any verdict**
  in this task (no predicate yet). A line still yields **at most one finding
  per selector**, in `(line, Selectors index)` order.
* **Consumed, not discarded (INV-4).** The per-occurrence evaluator consumes
  the extent classification, which feeds the allowance hook. In D-T1 that hook
  returns "not allowed" for every classification, and D-T2 replaces it. No
  computed-and-unused value is introduced (staticcheck SA4006).
* **`findSelector` stays a production function.** It is re-expressed as the
  first enumerated occurrence, so `TestFindSelector_LookaroundTable`
  (`writepath_test.go:166-189`) passes **unmodified**. Its named production
  caller is the per-line fast-path guard (`if !findSelector(line, sel) {
  continue }`) ahead of full enumeration. It must not become a test-only
  declaration.

**Acceptance criteria**
* AC-D1.1 — For the masked text of `internal/pathsafe/reparse_windows.go`, the
  extractor recovers the `syscall.CreateFile` call (located **by content**, not by
  a hard-coded line number) as one balanced extent of **8** raw depth-1 segments
  (7 arguments plus the trailing-comma artifact). Segment 1 is `0` after
  `strings.TrimSpace`, and segment 4 is `syscall.OPEN_EXISTING`. The four
  comment-only mentions produce **no** occurrence.
* AC-D1.2 — Unit tests classify **non-call**, **unbalanced (EOF)**, and
  **mismatched-bracket** extents as such. Each still yields the unchanged
  finding text `write primitive '<sel>' found`, so no new classification text
  appears.
* AC-D1.3 — A multi-line extent whose lines are separated by `\f` or `\u2028`
  maps every occurrence back to the correct `SplitLines` line (cursor test).
* AC-D1.4 — D-T1a's oracle, every golden test, and `retiredarch`'s
  `TestWritePathScanFileConsumesCanonicalMask` / `TestWritePathMaskFlow_*` pass
  **unmodified**. All three `check-write-path-precondition.sh` entry points
  pass. `golangci-lint`/`staticcheck` are clean (INV-1, INV-4). The extractor is
  reached from the production scan path, with no test-only declaration (INV-4).
* AC-D1.5 — If D-1′ cannot be followed as specified, or the `retiredarch` pin
  cannot stay unmodified, **HALT and return to Stage** (P-010). Do not
  re-decide the mechanism, and do not edit `retiredarch`.

### D-T2 — 034.003-T: Access-mode allowance predicate (D-2′)

Size **S**, complexity **medium**. Posture: test-first.
`writepath.go` + `writepath_test.go`. Depends on D-T1.

Implement **D-2′** (D-2 as amended by D-031-3). A single occurrence is
**allowed** only when **all** of the following hold:

1. the selector is exactly `syscall.CreateFile`;
2. the occurrence is a real call whose extent balanced under D-1′;
3. the extent interior contains **none** of `(` `)` `[` `]` `{` `}`, a
   backquote, or a double quote. This rejects call arguments, composite
   literals, index expressions, generic instantiation and `T{…}.Args()`-style
   multi-value smuggling.

   The quote check can only catch **visible tag-shaped raw strings**. `gomask`
   blanks interpreted strings and rune literals, delimiters included, so those
   are caught by rule 3b or by type checking.
   * **3b.** After discarding **one** trailing empty segment, **every** segment
     is non-empty after `strings.TrimSpace`. A blanked string or rune argument
     leaves an empty segment and is rejected.
4. after discarding **one** trailing empty segment there are **exactly 7**
   arguments;
5. argument 1 (`dwDesiredAccess`), trimmed, is **exactly the token `0`**;
6. argument 4 (`dwCreationDisposition`), trimmed, is **exactly the token
   `syscall.OPEN_EXISTING`**. Access `0` with `CREATE_NEW`, `CREATE_ALWAYS`,
   `OPEN_ALWAYS` or `TRUNCATE_EXISTING` can still create or modify a file;
7. argument 5 (`dwFlagsAndAttributes`), trimmed, is **exactly the token
   `syscall.FILE_FLAG_BACKUP_SEMANTICS`** and nothing else. No `|`
   combination, no other flag, and no operator is allowed. This excludes
   `FILE_FLAG_DELETE_ON_CLOSE` and every other flag.

   Round 2 narrowed this rule to the live call's exact flag. An earlier draft
   also allowlisted `FILE_FLAG_OPEN_REPARSE_POINT`, which no caller uses.

Any other spelling (`0x0`, `00`, `(0)`, `uint32(0)`, a named constant, or an
equivalent numeric disposition) is **rejected**. The predicate can only fail
as a false positive (INV-2).

The predicate runs on **each occurrence independently**. A line emits a finding
for a selector when **any** occurrence of it on that line is not allowed, and
the at-most-one-finding de-duplication is applied **after** the predicate. This
way an allowed call never hides a writing call on the same line.

The predicate is wired into the production scan path keyed on the selector
string, so it stays **dormant until D-T3** adds `syscall.CreateFile` to
`Selectors`. Its scan-level positive control therefore lands with the D-T5a
fixtures, inside the same shipment PR.

**Acceptance criteria**
* AC-D2.1 (AC-4.2, function level) — The real `reparse_windows.go` extent
  (located by content) is **allowed**. The allowance is keyed on the call's
  arguments, never on the file path.
* AC-D2.2 — Table tests **reject** each of the following:
  * a non-call reference, an unbalanced extent, and a mismatched-bracket extent;
  * 6 and 8 arguments;
  * each non-`0` access spelling listed above;
  * `CREATE_NEW`, `CREATE_ALWAYS`, `OPEN_ALWAYS` and `TRUNCATE_EXISTING`;
  * `FILE_FLAG_DELETE_ON_CLOSE`, alone and OR-ed with the allowed flag;
  * a brace composite (`T{p, 0, a, b, c, d, e}.Args()`);
  * a call-valued argument;
  * a visible tag-shaped raw-string argument (rule 3);
  * an interpreted-string or rune argument in an unconstrained slot (rule 3b,
    empty segment);
  * `syscall.FILE_FLAG_BACKUP_SEMANTICS|syscall.FILE_FLAG_OPEN_REPARSE_POINT`
    (rule 7 exactness).
* AC-D2.3 — A per-occurrence test proves that an allowed call and a writing call
  on **one line**, in both orders, produce the finding. Until D-T3 adds
  `syscall.CreateFile` to `Selectors`, this test calls the per-line evaluator
  with the selector **passed as a parameter**, not through `scanText`.

*Scenario counting (2-hour rule):* a table-driven test counts as **one** test
scenario. D-T2 therefore has three scenarios: AC-D2.1 allow, the AC-D2.2
rejection table, and AC-D2.3 per-occurrence.
* AC-D2.4 — These tests are authored first and **fail** before the predicate
  exists (red phase; newly authored assertions, so not an INV-5 pair).
* AC-D2.5 — D-T1a's oracle, the golden tests and the `retiredarch` pin are
  unmodified and green; `Selectors` is unchanged; lint is clean.
* AC-D2.6 — If D-2′ cannot be followed as specified, **HALT and return to Stage**.

### D-T3 — 034.004-T: Add six decidable selectors (folds in B72E9715)

Size **S**, complexity **low**. `writepath.go` +
`testdata/writepath_golden.json`. Depends on D-T2.

Append `syscall.CreateFile`, `syscall.Write`, `os.OpenRoot`, `os.Root`,
`io.CopyN` and `io.CopyBuffer` to `Selectors` (20 → **26**). Also correct the
`Selectors` doc comment (`writepath.go:28-31`, "ordered list of 20 … ported
verbatim"), which this change makes false.

Scope corrections carried from the original task:
* `os.Root` **method** calls (`root.Create`) are **not** decidable by
  qualified-selector matching and are **not** attempted (residual, D-T4). Only
  `os.Root` type-use and `os.OpenRoot` construction are added.
* `syscall.Write` does **not** match `syscall.WriteFile`, because
  `FollowedByWord` rejects the trailing `File` (residual item 6).
* The unlisted DB-driver selectors item stays removed (no database dependency
  in `go.mod`).

**Acceptance criteria**
* AC-D3.1 — The six selectors are present and appended **after** the existing
  20, so existing finding order is unchanged. The golden `selectors` array is
  updated in the **same commit**: red first (golden edited, test fails), then
  `Selectors`.
* AC-D3.2 (AC-4.2, gate level) — `scripts/check-write-path-precondition.sh`
  (bare) still **PASSES** on the tracked tree: D-T2 allows the
  `reparse_windows.go` call, and the comment-only mentions are masked.
  `TestRun_RepoMode_CurrentTreeIsClean` passes unmodified.
* AC-D3.3 (AC-4.8) — Zero false positives across the 17 non-test files.
* AC-D3.4 — `--self-test` and `--self-test-integrity` pass. Their golden stream
  captures are unchanged, since no fixture is added in this task. D-T1a's oracle
  stays green **unmodified**: it is restricted to the frozen 20.
* AC-D3.5 — Stash B72E9715 is cited in the commit body.

### D-T4 — 034.007-T: Record the residual evasion surface honestly

Size **S**, complexity **low**. Documentation only.
`scripts/check-write-path-precondition.sh` (header comment) + `writepath.go`
(package doc comment). Depends on D-T3.

Update the wrapper's "Detector scope" list to the 26 selectors and record —
never silently leave unhandled — the residual surface (AC-4.7):

1. **Named import aliases — KNOWN OPEN, pending Unit E (049-F).** An aliased
   import of a write-capable package (`import o "os"` → `o.WriteFile(…)`, or
   an alias of `database/sql` / `go.etcd.io/bbolt`) is **not detected**.
2. **`pathsafe.NewRoot` receiver / `Root.Resolve` first-caller tracking — KNOWN
   OPEN, pending Unit E.** No tripwire exists. The pathsafe risk-register
   triggers ("forced the moment a real write path exists", `root.go`) and
   feature 038-F's "once a live caller exists" trigger stay **awaited, not
   monitored**.
3. Dot-imports, blank imports, and local identifiers shadowing a package name
   (including a local `syscall` identifier, which the D-2′ predicate trusts by
   spelling).
4. `os.Root` **method** calls and `(*os.File).Write*` (the latter already
   recorded).
5. **Undecidable or non-simple call shape → rejected.** A `syscall.CreateFile`
   in any of these shapes is **rejected, not exempted** (D-2′ rules 2–7):
   * reached via a wrapper or function value;
   * with an extent that does not balance;
   * with any argument that is not a bare operand (call, composite literal,
     index, string or raw string);
   * with any access, disposition or flags spelling outside D-2′'s exact
     tokens.

   This is a **false-positive** surface. A future legitimate metadata-only call
   in such a shape trips the gate and needs an explicit, reviewed widening. It
   is never a silent hole.
6. **Write primitives outside the selector set.** These are not detected:
   * `ioutil.WriteFile`, `ioutil.TempFile` and `ioutil.TempDir`;
   * `syscall` write and namespace calls other than `syscall.CreateFile` and
     `syscall.Write`: `syscall.WriteFile`, `Open`, `Unlink`, `Rename`, `Mkdir`,
     `CreateHardLink` and `DeleteFile`;
   * `golang.org/x/sys/windows` and `golang.org/x/sys/unix` equivalents.

   None occurs in `internal/**` or `cmd/**` at `9b299c8`. Widening the selector
   set is **out of D-031-2's scope** and is tracked as stash entry
   **458F9385** (rev 5, D-S-5; kind task, priority low). Stage recommends
   folding it into 039-S / 049-F, because `go/ast` import-path resolution makes
   detection exact.
7. **Selector split by a newline or comment — KNOWN OPEN, closed by Unit E's
   AST engine.** Go inserts no semicolon after `.`, so both of these are valid
   Go that gofmt preserves:
   * `syscall.` + newline + `CreateFile(p, GENERIC_WRITE, …)`;
   * `os./**/WriteFile(…)`.

   After masking, neither contains the contiguous selector text, so **every**
   selector, not only the new ones, is evaded. This predates the plan and is
   recorded here because the risk statement must not understate the fail-open
   set. It is pinned as a reject fixture in E-T4.

**Residual-risk statement (required text, verbatim intent).** "Until feature
049-F ships, this gate is a **qualified-selector tripwire, not a complete
mechanical proof**. Items 1, 2, 6 and 7 are **known open fail-open surfaces**.
At `9b299c8` there are zero aliased write-capable imports, zero production
`Root.Resolve` callers and zero item-6 primitives in `internal/**`/`cmd/**`, so
items 1, 2 and 6 are not exploited today. None of the four is mechanically
guarded. The compensating control is human and agent PR review against this
list."

**Acceptance criteria**
* AC-D4.1 (AC-4.7) — All seven items and the residual-risk statement are present
  in **both** locations. Items 1 and 2 name feature **049-F** / shipment
  **039-S** as the closing unit, and the gate is **not** presented as complete.
* AC-D4.2 — The wrapper's detector-scope list matches `Selectors` exactly.
* AC-D4.3 — No executable line is modified; all gate entry points still pass.

*Ship guidance (not a task AC; it applies during PR review, after D-T4 is
done):* **review-loop terminator** (compound lesson
`2026-10-01-textual-yaml-guard-denylist-bypass-loop.md`). During 031-S PR
review, a bypass already listed here is answered by citing its item and 049-F.
It does not reopen the masked-text engine.

### D-T5a — 034.008-T: Verdict-boundary fixtures for the access-mode allowance

Size **S**, complexity **medium**. `scripts/testdata/writepath/*.go` (4 new
fixtures) + `testdata/writepath_golden.json`. Depends on D-T3.

Fixtures:

* `accept-syscall-createfile-metadata.go`: the metadata-only call alone.
* `reject-syscall-createfile-write.go`: the **positive control**. It contains
  the metadata-only call, a `GENERIC_WRITE` call on another line, and one line
  holding an allowed call **and** a writing call.
* `reject-syscall-createfile-evasion.go`: one line per D-2′ rejection shape:
  * access `0` + `CREATE_NEW`;
  * access `0` + `FILE_FLAG_DELETE_ON_CLOSE`;
  * access `0` + an OR-ed flag combination;
  * a brace-composite multi-value argument;
  * a tag-shaped raw-string argument;
  * a function-value reference.
* `reject-tag-shaped-raw-string-expr.go`: a tag-shaped raw string **outside a
  struct field** (`` q := `x:"os.Remove"` ``). It pins the masker behaviour that
  a `Field.Tag`-only `go/ast` walk would silently flip, so it is a parity
  hazard for Unit E's G-3.

**Acceptance criteria**
* AC-D5.1 (AC-4.1, AC-4.3) — Each fixture is self-tested with the expected
  verdict and has a `fixture_findings` golden row pinning the **exact** finding
  text (selector name and line). The positive-control row pins findings
  **only** on the writing lines, and none on the metadata-only line. That is
  the committed proof the allowance is per-occurrence, not file-level.
* AC-D5.2 — Falsifiability is **committed, not demonstrated once**. Each golden
  row names its selector, so removing that selector, disabling the predicate
  (accept fixture), or widening it (evasion fixture) fails
  `TestScanFile_FixtureFindings_MatchGolden`. Declared honestly (H-4): the
  fixtures are **green on arrival** against D-T3; their red phase lives in
  D-T2/D-T3.
* AC-D5.3 (AC-4.9) — All fixtures are gofmt- and goimports-clean (the lint job
  formats `testdata`).
* AC-D5.4 — `self_test` and `self_test_integrity` stream captures are updated in
  the same commit. D-T1a's oracle stays green unmodified, and all gate entry
  points pass.
* AC-D5.5 (AC-4.5) — No TOCTOU/hard-link mitigation and **no write primitive**
  is introduced outside `testdata` (INV-3, a named stop condition).
* AC-D5.6 — Each fixture's header comment states it is part of the
  **verdict-parity corpus for the 049-F `go/ast` migration**.

### D-T5b — 034.011-T (NEW): Presence fixtures for the new selectors

Size **S**, complexity **low**. `scripts/testdata/writepath/*.go` (4 new
fixtures) + `testdata/writepath_golden.json`. Depends on D-T5a (serialised only
because both edit the golden).

Split out of D-T5 at plan-review rev-4 (2-hour rule). Fixtures:

* `reject-syscall-write.go`;
* `reject-os-openroot.go`;
* `reject-os-root-type.go`: an `os.Root`-typed declaration or parameter **with
  no `os.OpenRoot` call**;
* `reject-io-copyn-copybuffer.go`: one B72E9715 selector per line, two golden
  findings.

**Acceptance criteria**
* AC-D5b.1 — Each fixture is rejected and has a golden row pinning its exact
  finding(s). Stream captures are updated in the same commit.
* AC-D5b.2 — Declared **green on arrival**; committed falsifiability via the
  selector-naming golden rows (as AC-D5.2).
* AC-D5b.3 — gofmt/goimports clean; each header comment marks it as 049-F parity
  corpus; stash B72E9715 is cited in the commit body.

### D-T6 — 034.009-T (NEW): Regression fixtures for pre-existing selectors (8E9F8E55)

Size **S**, complexity **trivial**. `scripts/testdata/writepath/*.go` (4 new
fixtures) + `testdata/writepath_golden.json`. Depends on D-T5b (serialised only
because both edit the golden).

Add `reject-os-mkdirtemp.go`, `reject-os-chown.go`, `reject-os-lchown.go` and
`reject-os-chtimes.go`. 8E9F8E55 names all four. The operator's summary named
only three, but directed that the stash entry be folded in. `os.Chtimes` is
included because the entry's own text lists it and it shares the gap.

It was flagged for operator confirmation in rev 4. **Rev 5: CONFIRMED by the
operator (D-031-6).** All four fixtures stay in scope.

*Fixture counting (2-hour rule; applies to D-T5a, D-T5b and D-T6):* a fixture
data file plus its golden row and stream-capture lines count as **one**
artifact and **one** scenario of a single table-driven golden test. These
tasks are therefore two artifacts: the fixture set and the golden.

**Acceptance criteria**
* AC-D6.1 — Each fixture is rejected and has a golden row pinning its exact
  finding. Stream captures are updated in the same commit.
* AC-D6.2 — Declared **green on arrival** (the selectors predate this unit).
  These are regression/parity fixtures whose committed falsifiability is the
  selector-naming golden row (as AC-D5.2).
* AC-D6.3 — gofmt/goimports clean; each header comment marks it as 049-F parity
  corpus; stash 8E9F8E55 is cited in the commit body.

### Moved out of Unit D

| Task | Disposition |
|---|---|
| 034.005-T `Root.Resolve` tripwire | Removed from 031-S via `shipment return-blocked`, re-parented to 049-F via `backlogit adopt`, rewritten as E-T5 |
| 034.006-T named import aliases | Same path, rewritten as E-T4 |

---

## Unit E — 049-F / 039-S (NEW): Migrate the write-path scanner to `go/ast` and close the alias and `Resolve` residuals

> **Rev 7 title note.** Rev 7 *narrows* the `Resolve` residual (item 2) to
> the G-6 cross-function cases; it does not close it (see E-T6). The heading
> and the 049-F title are kept for traceability.

**Gating.** (1) `blocks` edge **039-S → 031-S**: Unit E cannot be claimed before
Unit D ships, because Unit D's fixtures are its parity corpus. 031-S shipped at
`961b652`, so this edge is satisfied. (2) **STAGE HOLD** on 049-F: E-T2..E-T6
were **provisional** until the spike.
* **Rev 7.** They are re-planned against the spike findings, and E-T7 is added.
* The hold lifts only when the `## Plan Review — Revision 7 (Unit E)` gate
  returns ADVISORY or better with no P0/P1. Its outcome is recorded there.

**Why a hold and not just an edge.** Option B **invalidates** the original
spike's mechanism decisions (deliberation §4.2), so the spike that replaces them
must exist **before** implementation tasks are final. Spikes and their findings
are Stage artifacts (P-010: Ship may not create or modify spike or plan
artifacts), so the investigation is **Stage-executed** under the P-016
spike-worktree exception; the in-shipment task E-T1 is the read-only gate that
verifies it, mirroring the original 034.001-T.

**Invariant carried by every Unit E task — gomask is retained.**
`tools/gatecheck/internal/gomask` **must not be deleted or narrowed**:
`retiredarch` depends on it independently of `writepath`. Unit E may remove
`writepath`'s own import of it, nothing more.

### E-T1 — 049.001-T: Go-era `go/ast` spike (Stage-executed) and its verification gate

Size **XS** (Ship's verification half), complexity **trivial**. Read-only.
Depends on 034.009-T (the last task of the Unit D corpus chain; the full
corpus must exist on `main`).

> **Spike status — COMPLETE (rev 7, 2026-10-02).** Stage ran the spike at
> `961b652`, after 031-S merged and before any 039-S claim. It used the P-016
> spike-worktree exception: the worktree `logs/spike-049-wt` was created, used
> for throwaway test code only, and removed. Nothing was committed from it.
>
> **Findings:** `docs/decisions/2026-10-02-intercom-go-writepath-go-ast-spike.md`.
>
> | Question | Verdict |
> |---|---|
> | G-1 | FEASIBLE. The boundary is `scanSource(relPath, src string) ([]string, error)`. `gomask` is still used, for raw-string interiors and the line map. |
> | G-2 | FEASIBLE. Any parser error is ED-2 exit 1, and positions use `PositionFor(pos, false)`. |
> | G-3 | EXACT PARITY, with 0 divergences in each corpus: 19/19 fixtures, 3/3 `filebased`, 17/17 tracked, 155/155 `.go` files in the scratch worktree (154 tracked + the harness). AC-E2.6 items 1–3 suffice. |
> | G-4 | FEASIBLE. A representation change only, made strictly stricter by import-path `syscall` resolution. |
> | G-5 | FEASIBLE and **additive**. The binding additivity rule is restated in E-T4. |
> | G-6 | Syntactic intra-function tracking. `go/types` is rejected. |
> | G-7 | Every task within the 2-hour rule, and E-T2 de-risked to `medium`. |
>
> The H-5 re-measure at `961b652` found no new trigger. The commands are the H-5
> commands of this plan; the results are summarized under H-5 ("Rev 7 run").
> Ship's verification of AC-E1.1–AC-E1.3 is
> the only remaining E-T1 work.

> **Spike timing — SCHEDULED (rev 5, operator decision D-049-1).** The
> Stage-executed spike runs **after 031-S merges to `main` and before Ship
> claims 039-S**. G-3 verdict parity needs the whole Unit D corpus, up to
> and including 034.009-T, on `main`.
> * **Execution.** Stage runs it under the P-016 explicit, time-boxed
>   spike/research worktree exception. Stage records the spike context, and
>   cleans up the worktree or hands off findings before Ship claims 039-S.
> * **Trigger mechanism.** Stash entry **1EEBECA5** (kind spike, priority
>   high, "STAGE-SCHEDULED SPIKE") brings it up in the first Stage triage after
>   031-S closes. If it is triaged before then, Stage defers it with the
>   reason "trigger not fired", and archives it only once the spike document
>   exists.
> * **Backstop.** 049-F's `blocked` status feeds the `blocked_stale` hook
>   signal (7 days), to which Stage is subscribed.
> * **After the spike.** Stage re-plans E-T2..E-T6, folds in stash 458F9385
>   (residual item 6), re-runs plan-review, and only then lifts the 049-F
>   STAGE HOLD. *(Done in rev 7. The re-plan covers E-T2..E-T7, and 458F9385
>   became E-T7. The re-gate is under "Plan Review — Revision 7 (Unit E)".)*
> * **Why `go/ast` at all (operator rationale, accepted Q&A, D-049-1).** The
>   write-path gate is **CI tooling that guards intercom-go's own source**,
>   not runtime code. `go/ast` gives:
>   * exact call boundaries;
>   * import-alias resolution;
>   * `NewRoot` / `Root.Resolve` receiver tracking;
>   * immunity to split selectors.
>
>   None of those gaps is exploited today (H-5 re-measure). Their value grows
>   as the Copilot SDK client gains real workspace write paths. That is why
>   039-S stays **behind the spike** and **sequenced after 031-S**, rather
>   than being cancelled or pulled forward.

Stage produces `docs/decisions/{date}-intercom-go-writepath-go-ast-spike.md`
answering, with evidence against the live corpus:

* **G-1 Feasibility and boundary** — `go/parser` + `go/ast` reproduce every
  finding's text, line and order across the full corpus.
  * Define a **source-level detection boundary**: it takes the decoded,
    **unmasked** `pysem.ReadText` output and returns findings **plus an
    error**. The masked-text `scanText` signature can carry neither valid Go
    syntax nor parse errors.
  * Decide how `retiredarch/writepath_mask_test.go`'s writepath half is
    re-anchored without weakening `retiredarch`'s canonical-mask and
    single-definition protections.
* **G-2 Fail-closed on unparseable input** — **any** `go/parser` error, even
  when a partial AST is also returned, makes that file an `::error::` ED-2
  failure with exit 1. The partial AST is **never scanned and never skipped**,
  and no `AllErrors` partial-scan mode is used.
  * The existing `filebased` inputs (BOM, CRLF, lone-CR, invalid UTF-8) keep
    their verdicts.
  * Positions use `fset.PositionFor(pos, false)`, so `//line` directives cannot
    move findings; a `//line` fixture pins this.
* **G-3 Verdict parity** — every existing fixture, every Unit D fixture and the
  live tree keep their verdicts. **Named hazards:**
  * **Struct tags.** `reject-struct-tag-selector.go` and
    `reject-tag-shaped-raw-string-expr.go` are rejected today **only** because
    the masker leaves tag-shaped raw strings visible. Under `go/ast` these are
    `*ast.BasicLit`s, so a naive `SelectorExpr` walk (or a `Field.Tag`-only
    walk) would flip them to **accept**, a fail-open (INV-2).
  * **Line numbering.** `go/token` counts lines on `\n` only, while the golden
    uses `pysem.SplitLines` boundaries (`\v`, `\f`, `\x1c`–`\x1e`, `\x85`,
    `\u2028`, `\u2029`). Offsets must be mapped through `SplitLines` boundaries
    **of the masked text**.
    * Boundary runes inside comments or strings are erased by the masker, so
      the legacy line number of a later hit is the **masked** line number
      (AC-D1a.2). Parity must reproduce this.
    * Code-position `\f`, `\v` and `\u2028` are **not Go whitespace**, so they
      are `go/parser` errors and fall under G-2 fail-closed.
    * Verify `go/scanner`'s whitespace and illegal-character handling, and
      `go/token`'s line counting, against the pinned GOROOT source, per compound
      lesson `2026-09-08-verify-go-stdlib-internals-claims-against-goroot-source.md`.
  * **Oracle adaptation.** Confirm that AC-E2.6 items 1–3 suffice, or HALT.
    Any adaptation beyond them returns to Stage and plan-review. The spike may
    not redesign the oracle after seeing where it disagrees (H-3).
* **G-4** — D-2′ mapped onto `*ast.CallExpr` as a **representation change only,
  never a widening**. Every D-2′ rejection is preserved:
  * bare-operand arguments only;
  * exactly 7 `Args`;
  * `Args[1]` an INT `*ast.BasicLit` with value `0`;
  * `Args[4]` exactly `syscall.OPEN_EXISTING`;
  * `Args[5]` exactly `syscall.FILE_FLAG_BACKUP_SEMANTICS`;
  * the `X` of the callee, `Args[4]` and `Args[5]` each resolving to the import
    path `"syscall"`;
  * `call.Ellipsis == token.NoPos`.

  Any allowance the masked engine rejects needs a separately approved policy
  exception with committed regression cases.
* **G-5** — named-alias resolution via `*ast.ImportSpec`. A **dot import** of
  any write-capable package fails closed (a finding), and a blank import is
  inert.
* **G-6** — `NewRoot` receiver binding: syntactic intra-function tracking vs
  `go/types`; cost and false-negative profile of each.
* **G-7** — sizing of E-T2..E-T6 against the 2-hour rule.

**Acceptance criteria (Ship, read-only)** — rev 7 mirrors these verbatim in
049.001-T's `acceptance-criteria` section.
* AC-E1.1 — The spike document
  (`docs/decisions/2026-10-02-intercom-go-writepath-go-ast-spike.md`) exists.
  It was produced by Stage, not Ship, under the P-016 spike-worktree exception
  **after** 031-S merged (034.009-T on `main`), and it records G-1..G-7 with a
  decision each.
* AC-E1.2 — The post-spike re-plan gives E-T2..E-T7 (049.002-T..049.007-T;
  E-T7 was added by rev 7) each:
  * a citation of the spike;
  * an explicit file list and test-scenario count within the 2-hour rule
    (or a recorded, justified deviation);
  * at least one verifiable AC;
  * a red/green phase declaration.

  None is left at complexity `high` without a recorded de-risking step. The
  re-plan was re-gated through plan-review (ADVISORY or better, no P0/P1),
  and the 458F9385 fold-in decision is recorded. *(D-049-11: this re-gate
  clause is discharged by the operator hold-lift override. No round 7 ran,
  and the verdict of record stays round 6 FAIL. Ship checks instead that
  D-049-11 is recorded; see "Operator hold-lift decision (D-049-11)".)*
* AC-E1.3 — The spike worktree was cleaned up, or its findings handed off,
  before 039-S was claimed. Stash 1EEBECA5 is archived with a pointer to the
  spike document.
* AC-E1.4 — Ship decides, re-sizes and narrows nothing here (P-010). If any of
  AC-E1.1..AC-E1.3 is unmet, **HALT and return to Stage**.

### E-T2 — 049.002-T: Replace masked-text detection with `go/parser` + `go/ast`

Size **M**, complexity **medium**. This was `high` before rev 7; the E-T1 spike
is the recorded de-risking step (G-3 exact parity, G-7). Depends on E-T1.

* **Spike citation:** G-1, G-2, G-3 and G-7 in
  `docs/decisions/2026-10-02-intercom-go-writepath-go-ast-spike.md`.
* **Files (5; the fifth added by R4-3):** `writepath.go`, `writepath_test.go` and
  `writepath_oracle_test.go`, all in `tools/gatecheck/internal/writepath/`, and
  `tools/gatecheck/internal/retiredarch/writepath_mask_test.go`. Round-4
  amendment R4-3 adds a fifth file, `writepath_extent_test.go`, for a single
  literal edit (see "`scanText` after E-T2").
  * The production change is confined to `writepath.go`. The other four are
    test files that must change atomically with it (AC-E2.6, G-1, R4-3).
  * This is the upper edge of the 2-hour heuristic, and it is accepted because
    splitting the task would leave `main` red between commits.
* **Test scenarios (3):**
  1. Unparseable input, including a partial AST, fails closed with exit 1.
  2. A `//line` directive does not move positions.
  3. The golden byte-identity is re-routed through `scanSource`.
* **Phase:** **red → green.** The scenario 1 and 2 tests are committed red
  against the masked engine, then made green by the swap in the same task.
  The golden and oracle are characterization (green before and after).

**Engine design, as proven by the spike (G-1 and the spike's algorithm section).**
* The production boundary is `scanSource(relPath, src string) ([]string, error)`.
  It takes the decoded, unmasked `pysem.ReadText` output.
* It parses with `parser.ParseFile(fset, relPath, src,
  parser.ParseComments|parser.SkipObjectResolution)`.
* On **any** error, it returns the error and nothing else. The caller emits an
  ED-2 `::error::` and exits 1.
* It computes `gomask.MaskGoNonCode(src)` once, as the single canonical masker.
  Line numbers come from `pysem.SplitLines` boundaries of that masked text,
  through a rune-index map.
* **Selector hits.** It matches each `*ast.SelectorExpr` whose `X` is an
  `*ast.Ident` against `Selectors`, using the line of `X`.
* **Raw-string hits.** For each raw-string `*ast.BasicLit`, it compares the
  interior (`Pos+1..End-1`) with the masked text. A **visible** interior is
  scanned with `selectorOccurrences`, which preserves the tag-shaped
  raw-string rule (G-3).
* **Output.** Hits are de-duplicated by (line, selector index), sorted by line
  and then by `Selectors` order, and formatted exactly as `scanText` formats
  them.

Swap the detection engine in **one commit**. Presence selectors become
`*ast.SelectorExpr` matches on the package identifier. Findings keep the exact
text, line and order. Parse failure fails closed (G-2), the struct-tag rule is
preserved per G-3, and production detection goes through the G-1 source-level
boundary. The masked-text extractor and predicate stay in place for
`syscall.CreateFile` until E-T3: `scanSource` calls the existing
`occurrenceAllowed(sel, extractExtent(masked, offset))` for a
`syscall.CreateFile` hit.
* **`scanText` after E-T2 (rev 7 cycle 2, SB2-2).** E-T2 removes `scanText`'s
  two non-extent callers: the golden row loop (`writepath_test.go:148`, scenario
  3) and the oracle's production side (`writepath_oracle_test.go:242`, AC-E2.6).
  The production entry point (`writepath.go:371`) is also re-pointed to
  `scanSource`. `scanText`, `advanceCursor`, `extractExtent` and the text-based
  `occurrenceAllowed` stay **unexported and unchanged** in E-T2, together with
  `lineReportsSelector` and `findSelector`. Their remaining test callers are in
  `writepath_extent_test.go` and in the `writepath_test.go` predicate,
  same-line and lookaround tests (`:161–182`, `:384–512`). E-T2 does not edit
  any of those tests, so the package compiles and every Unit D extent test
  stays green at the E-T2 boundary. E-T3 retires them; see E-T3 (post-cycle-3
  amendment GK3-3).
* **Round-4 amendment R4-3 (OP4-1; applied after round 4, NOT re-gated).**
  INV-4 makes `golangci-lint run ./...` blocking, and `.golangci.yml` enables
  `unparam`.
  * After the re-pointing above, `scanText`'s only callers are the four extent
    tests (`writepath_extent_test.go:109`, `:118`, `:185`, `:210`). All four
    pass the literal `"x.go"`. `unparam` would then report that `relPath`
    always receives `"x.go"`, so E-T2's own boundary would be lint-red.
  * E-T2 therefore makes **one test-only, verdict-preserving literal edit**.
    In `TestScanText_OneFindingPerSelectorPerLine`
    (`writepath_extent_test.go:118`), it changes `"x.go"` to `"y.go"` in the
    call and in both expected finding strings. No other line of that file
    changes in E-T2.
  * That makes `writepath_extent_test.go` a fifth file of E-T2 (a single
    literal). This is an accepted deviation from the 2-hour heuristic, for the
    same atomicity reason as the other three test files.
  * Two alternatives are rejected. Moving the `scanText` deletion forward into
    E-T2 would merge E-T2 and E-T3 beyond the 2-hour rule. A `//nolint`
    suppression would weaken an H-3 control.
* **Round-4 amendment R4-5 (OP4-5; NOT re-gated).** `gomask` preserves the
  **rune** count, not the byte count. The interim `extractExtent(masked,
  offset)` call therefore cannot take a `go/token` source byte offset
  directly.
  * `offset` is derived from the selector's end position in three explicit
    steps:
    1. `SelectorExpr.End()` is a `token.Pos`, not a byte offset. Convert it
       first to a source byte offset with the owning file's
       `token.File.Offset`: `tf := fset.File(sel.Pos())`, then
       `tf.Offset(sel.End())`. This follows the `retiredarch/pin.go:643-648`
       pattern, including its fail-closed handling of a nil file or an
       invalid position.
    2. Map that source byte offset to a rune index.
    3. Map the rune index to the masked text's byte offset, through the same
       rune-index map that the line table uses.
  * Neither a raw `token.Pos` nor a source byte offset may be used to index
    the masked text. A raw `token.Pos` still includes the file base, so it
    would shift the extent and could flip the `syscall.CreateFile` allowance.

E-T2 owns the G-1 re-anchoring of
`retiredarch`'s writepath scan-loop pin: the `writePathMaskFlow` pin at
`writepath_mask_test.go:35` is re-anchored on `scanSource`. Its
canonical-mask and single-definition assertions are not weakened.
* **Round-6 amendment R6-6 (OP6-6 P3; applied after round 6, NOT re-gated).**
  Today the pin positively requires `scanText`. It returns false when
  `scanText` is absent (`writepath_mask_test.go:51`). It also fails with "no
  writepath source declares both scanFile and scanText" (`:300`). E-T3
  deletes `scanText` but cannot edit `retiredarch`, so the pin that E-T2
  re-anchors must not depend on `scanText` at all. It anchors on `scanSource`
  calling `gomask.MaskGoNonCode` directly on its source parameter, and on that
  masked result being consumed through `pysem.SplitLines`. `scanSource` masks
  into a local variable rather than receiving masked text as a parameter, so
  the `scanTextRangesParam` "second parameter" shape is not reused verbatim.
* **Round-6 amendment R6-2, E-T2 half (OP6-2 P3; NOT re-gated).** When E-T2
  re-points the golden row loop (`writepath_test.go:148`) to `scanSource`, it
  also renames the `t.Fatalf` message prefixes `"scanText(%s) …"` at `:150`
  and `:154` to `"scanSource(%s) …"`. Otherwise the AC-E3.3 `git grep`
  (`\bscanText\(`) matches those string literals after E-T3 and cannot be
  empty.

**Acceptance criteria**
* AC-E2.1 — The golden JSON is **byte-identical** (`selectors`,
  `fixture_findings`, stream captures, `filebased`). The test harness may be
  re-routed through the G-1 boundary so that the golden tests exercise the
  production engine and not a retained legacy path. The Unit D corpus is the
  parity contract.
* AC-E2.2 — `reject-struct-tag-selector.go` and
  `reject-tag-shaped-raw-string-expr.go` stay **rejected**.
* AC-E2.3 — A committed unparseable-input test, including a partial-AST case,
  fails closed with exit 1, and a `//line` test pins positions.
* AC-E2.4 — `gomask` is not deleted or modified, and `retiredarch` keeps its
  canonical-mask protection.
* AC-E2.5 — Closure evidence includes a before/after byte diff of all three
  gate entry points' output (compound lesson
  `2026-09-28-extracting-bash-heredoc-python-engines-into-importable-modules.md`).
* AC-E2.6 — **The single authorised oracle adaptation** (D-T1a lifecycle). In
  the same commit as the engine swap, D-T1a's oracle is adapted as follows:
  1. **Production side.** It is re-pointed to the G-1 source-level boundary and
     fed the original unmasked source. The legacy side stays fed the masked
     text, so the oracle tests the production engine and never a retained
     legacy path.
  2. **Fixture inputs.** The fixture input set is frozen to an **explicit name
     list** equal to the corpus on `main` at 031-S merge. E-T4 and E-T5
     fixtures (alias, dot-import, split-selector, `Resolve`) are therefore never
     oracle inputs.
  3. **Unparseable inputs.** AC-D1a.2's code-position `\f`/`\v`/`\u2028` input
     is moved from a parity assertion to a G-2 **fail-closed assertion**
     (error, exit 1). This is a stricter verdict, never a weaker one.
     Invalid UTF-8 already fails at `ReadText` (AC-D1a.1), so it is unchanged.

  The legacy implementation, the frozen 20 and every other expectation stay
  **unedited**. Any further oracle change HALTs (H-3).
* AC-E2.7 (round 6, D-049-10; NOT re-gated) — The re-anchored `retiredarch`
  pin has no positive dependency on `scanText`. It anchors on `scanSource`
  masking its source parameter with `gomask.MaskGoNonCode` and consuming the
  result through `pysem.SplitLines` (R6-6). The `writepath_test.go:150`/`:154`
  message prefixes name `scanSource` (R6-2).

### E-T3 — 049.003-T: Port the allowance predicate onto `*ast.CallExpr`

Size **M**, complexity **medium** (rev 7 cycle 2: was S; the third file and
the AC-D1.1–D1.3 test port below raise the size). Depends on E-T2.

* **Spike citation:** G-4, plus the G-3 stricter-divergence list.
* **Files (3):** `writepath.go`, `writepath_test.go` and
  `writepath_extent_test.go`, all in `tools/gatecheck/internal/writepath/`.
  `writepath_extent_test.go` is required because it calls `extractExtent`
  (`:52`, `:93`, `:180`), `scanText` (`:109`, `:118`, `:185`, `:210`) and
  `advanceCursor` (`:22`, `:158`, `:207`, `:226`, `:236`; `:236` added by
  round-6 amendment R6-4, OP6-4). The package would not
  compile if E-T3 deleted those functions without it (rev 7 cycle 2, SB2-2).
  `writepath_test.go:384–512` is ported in the same commit (post-cycle-3
  amendment SB3-1/GPT3-1). That range covers the helpers `createFileCall`,
  `allowedCreateFileArgs` and `firstExtent`, the `occurrenceAllowed` predicate
  tests, and the AC-D2.3 same-line test
  `TestLineReportsSelector_AllowedCallCannotHideWritingCall`.
  `writepath_test.go:161–182` (`TestFindSelector_LookaroundTable`) is also
  ported. Three files is at the edge of
  the 2-hour heuristic. It is accepted for the same reason as E-T2: the
  deletion and the test port must land atomically or the package does not
  compile.
* **Test scenarios (3):**
  1. An 8-argument call with a string argument is rejected.
  2. A variadic (`...`) call is rejected.
  3. The `syscall` qualifier resolves to import path `"syscall"`, and an
     alias of another package named `syscall` loses the allowance.
* **Phase:** **red → green.** Scenario 3 is red against the name-based
  predicate. Scenarios 1–2 are characterization (rejected by both engines),
  written to pin the port.

Re-express D-2′ over the AST (G-4) as a representation change only, and delete
the masked-text extent extractor (`extractExtent` and the text-based
`occurrenceAllowed`), together with `scanText`, `advanceCursor` and the
`extent` type, which have no remaining caller once the extractor is gone.
**Complete deletion list (post-cycle-3 amendment SB3-1/SB3-2/GK3-2):**

* `scanText`, `lineReportsSelector`, `findSelector`, `advanceCursor`,
  `extractExtent` and the text-based `occurrenceAllowed`;
* the `extent` type, `extentKind` and its constants, and `bracketOpener`;
* the test helpers `firstExtent` and `absoluteOccurrences`. `createFileCall`
  and `allowedCreateFileArgs` are either re-expressed as AST test inputs or
  deleted.

`lineReportsSelector` is deleted because its only production caller is
`scanText`. `findSelector` is deleted because its only production caller is
`lineReportsSelector`, and Unit D forbids it from becoming test-only.
`nextOccurrence` is **retained** as the raw-string boundary helper behind
`selectorOccurrences`.
`selectorOccurrences` is **retained**: `scanSource` uses it for visible
raw-string interiors (G-3), and `TestSelectorOccurrences_EnumeratesEvery`
stays unchanged.
**Round-5 amendment R5-3 (OP5-4 P3; NOT re-gated).** Imports left unused by
the deletions are removed in the same commit. For example, `unicode/utf8`
(`writepath.go:72`) is used today only by `advanceCursor` (`:192–193`), unless
the `scanSource` rune-index map uses it. `scanSource` is added to `writepath.go`
in E-T2 (the production change is confined to that file), so the
`gomask` import of `writepath.go` stays used at both boundaries.
**Round-6 amendment R6-5 (OP6-5 P3; NOT re-gated).** That sentence covers
`writepath.go` only. In `writepath_test.go`, E-T2 re-points `:148` to
`scanSource`, and E-T3 drops the per-row and `createFileCall`
`gomask.MaskGoNonCode` calls (R5-1). E-T3 also ports `:411` and `:499`/`:505`.
After that, the `gomask` import of `writepath_test.go` (`:11`) is unused, and
E-T3 removes it in the same commit. Leaving it in place is a compile error.
`writepath_extent_test.go` keeps its `gomask` import, because the R4-4
port computes the expected lines from `SplitLines` of the masked text.
**Round-6 amendment R6-3 (OP6-3 P3; NOT re-gated).** The new AST allowance
predicate must not reuse any name on the complete deletion list, because
AC-E3.3's `git grep` would then match correct code. The name is, for example, `callAllowed`; any name that is not on the list
is acceptable.
* **Extent-test port (rev 7 cycle 2, SB2-2).** The Unit D extent contracts
  move to named AST tests in `writepath_extent_test.go`, in the same commit:

  | Unit D test (contract) | Ported to |
  |---|---|
  | `TestExtractExtent_ReparseWindowsCreateFile` (AC-D1.1) | `TestCallExpr_ReparseWindowsCreateFile`: the live `syscall.CreateFile` in `internal/pathsafe/reparse_windows.go`, located by content and parsed with `go/parser`, is one `*ast.CallExpr` with 7 arguments that the predicate allows. Its comment-only mentions yield no `scanSource` finding. |
  | `TestExtractExtent_Classifications` (AC-D1.2, extractor level), **parseable cases** (`os.Remove`, `os.Remove;`, call, nested depth-1 split, trailing comma, empty call) | `TestCallExpr_Classifications`: each case wrapped as a parseable file (for example `var f = os.Remove`), classified over the AST (non-call selector vs `*ast.CallExpr` with its argument count), with the same verdicts. `len(CallExpr.Args)` replaces the legacy segment count (round-4 amendment R4-7, OP4-7). It equals the legacy count except in two rows: trailing comma (`os.Remove(a,\n)`, 2 legacy segments, 1 argument) and empty call (`os.Remove()`, 1 empty legacy segment, 0 arguments). |
  | `TestExtractExtent_Classifications`, **non-parseable cases** (unbalanced EOF, unbalanced with no closer, mismatched closer, mismatched brace, form feed in code position, and — round-4 amendment R4-1, GPT4-1/OP4-2 — call after whitespace and newline, `os.Remove \t\n(a, b)` at `writepath_extent_test.go:82`: automatic semicolon insertion after `Remove` makes `(a, b)` a syntax error) — post-cycle-3 amendment GK3-1 | `TestScanSource_SyntaxErrorFailsClosed`: each case, wrapped in a function body, is a **parse error** under G-2. It returns the error with exit 1, yields no finding, and the partial AST is not scanned. That is a stricter fail-closed verdict, never a weaker one. The distinct `extentUnbalanced` and `extentMismatched` kinds are **not** preserved; they were masked-text artefacts. |
  | `TestScanText_ClassificationsKeepFindingText` (AC-D1.2, scan level), `TestScanText_OneFindingPerSelectorPerLine` | Re-routed through `scanSource`, with the same expected finding text and the same one-finding-per-(line, selector) rule, for the **parseable** inputs (`f := os.Remove` and `os.Remove(a)` inside a function body). The non-parseable inputs `os.Remove(a, (b` and `os.Remove(a, [b)` move to `TestScanSource_SyntaxErrorFailsClosed` (GK3-1). |
  | `TestLineReportsSelector_AllowedCallCannotHideWritingCall` (AC-D2.3, R4-P1-3; post-cycle-3 amendment SB3-1/GPT3-1) | `TestScanSource_AllowedCallCannotHideWritingCall`: the same three single-line cases, each wrapped as a parseable function body and run through `scanSource`. An allowed `syscall.CreateFile` alone gives no finding; allowed-then-writing and writing-then-allowed each give exactly one `syscall.CreateFile` finding on that line. The "allowance disabled" sub-case is deleted with `advanceCursor`, its only subject. |
  | `TestFindSelector_LookaroundTable` (`writepath_test.go:161–182`; post-cycle-3 amendment GK3-2) | Same table, re-pointed at `selectorOccurrences` (want `len(...) > 0`). The lookaround boundary contract stays pinned on the retained `nextOccurrence`. The `t.Fatalf` prefix `"findSelector(%q, %q) …"` at `:180` is renamed to `"selectorOccurrences(%q, %q) …"` (round-6 amendment R6-2, E-T3 half, OP6-2). |
  | `TestCursor_MultiLineExtentAcrossSeparators` (AC-D1.3) | `TestScanSource_MultiLineCallAcrossSeparators`: a multi-line call whose lines are separated by `\f` or `\u2028` in **comment or string positions** maps every finding to the correct `pysem.SplitLines` line through the rune-index line table, including a hit on the final line with and without a trailing boundary. The same separators in **code position** are a parse error, so they fail closed with exit 1 (G-2, AC-E2.6 item 3). That is a stricter verdict, never a weaker one. **Round-4 amendment R4-4 (OP4-4):** separators in comments and interpreted strings are erased by the masker, so they cannot tell the `SplitLines` table apart from `go/token`'s `\n`-only lines. The test therefore adds one **discriminating case**: a visible raw-string struct tag containing `\u2028` (or `\f`), followed by a multi-line writing call. The expected finding lines come from `SplitLines` of the masked text. The test first asserts, as a precondition, that `fset.PositionFor(pos, false).Line` **differs** from each expected line, so the case cannot pass vacuously. |
  | `TestCursor_InvalidUTF8FailsClosed`, `TestCursor_LastLineBoundsCheck` | Deleted with `advanceCursor`, their only subject. Invalid UTF-8 still fails closed at `ReadText` (AC-D1a.1). Last-line mapping is pinned by the AC-D1.3 port above. |

  The `writepath_test.go` predicate tests (`firstExtent` and its callers) are
  re-expressed over the `*ast.CallExpr` with unchanged cases and verdicts,
  except the "unbalanced extent" and "mismatched extent" rows (`:443–444`).
  Those are non-parseable and move to `TestScanSource_SyntaxErrorFailsClosed`
  under the GK3-1 rule above.
  * **Round-4 amendment R4-2 (OP4-3; NOT re-gated).** Every non-parseable row
    of `TestOccurrenceAllowed_RejectionTable` moves to
    `TestScanSource_SyntaxErrorFailsClosed`, not only `:443–444`. At
    `961b652`, those rows are:
    * `:443` unbalanced extent;
    * `:444` mismatched extent;
    * `:467` leading empty segment, `syscall.CreateFile(, …)`, which gives
      "expected operand";
    * `:468` two trailing empty segments, `…, 0,,)`, which gives "expected
      operand".

    Every other row parses. The executor confirms this list with a
    `go/parser` probe of each wrapped row before porting.
  * **Round-5 amendment R5-1 (GPT5-1 P1 / OP5-1 P2; applied after round 5,
    NOT re-gated). The ported rows are built from unmasked source.**
    * The AST engine parses the decoded, **unmasked** source, so every ported
      row is reconstructed as the unmasked call text inside the R4-6 wrapper.
      `createFileCall` is re-expressed **without** `gomask.MaskGoNonCode`, and
      the per-row `gomask.MaskGoNonCode(...)` calls are dropped. The existing
      `tc.masked` values are **never** parsed or probed: masking blanks string
      and rune literals and the raw-string delimiters, so at `961b652` the
      masked values of `:464` (tag-shaped raw string, "missing ','"), `:465`
      (interpreted string, "expected operand") and `:466` (rune argument,
      "expected operand") are also parse errors, while their unmasked source
      parses (Stage probe, round 5).
    * `:464`, `:465` and `:466` therefore stay **parseable predicate
      rejections** in the AST port. They pin, at the third argument
      (`Args[2]`), the G-4 rule that STRING and CHAR `*ast.BasicLit`
      arguments are not bare operands and are rejected (the other generic
      positions are covered by round-6 amendment R6-1 below; spike G-4: a bare operand is an `*ast.Ident`, an
      INT/FLOAT/IMAG `*ast.BasicLit`, or a `*ast.SelectorExpr`,
      `*ast.UnaryExpr`, `*ast.StarExpr` or `*ast.BinaryExpr` built only from
      bare operands). They must **not** move to
      `TestScanSource_SyntaxErrorFailsClosed`.
    * The executor's `go/parser` probe over the unmasked wrapped rows must
      reproduce **exactly** `{:443, :444, :467, :468}` as the parse errors.
      Any other result (an extra parse error, or one of the four parsing) is
      an **H-2 HALT**, not a licence to move a row.
  * **Round-6 amendment R6-1 (GPT6-1 P1 / OP6-1 P2 / GM6-1 P2, which closes
    PRE5-1; applied after round 6, NOT re-gated). Positional coverage of the
    bare-operand rule.**
    * The ported table pins non-bare-operand rejection at `Args[0]` (call and
      index, `:462`–`:463`) and at `Args[2]` (string, rune and raw string,
      `:464`–`:466`) only. `Args[3]` and `Args[6]` are always `nil` and `0`.
      A predicate that never checks `Args[3]` or `Args[6]` therefore passes
      every specified row, and that would be a widening of D-2′ (INV-2).
      Stage reproduced this in round 6. A throwaway `go/parser` program
      compared a full G-4 predicate with one that checks only `Args[0]` and
      `Args[2]`. The two gave 0 divergences over `:442`–`:469` and the
      allowed control.
    * E-T3 adds **nine** rows to the ported rejection table. Each is built
      from unmasked source inside the R4-6 wrapper. Each changes exactly one
      generic argument of `allowedCreateFileArgs`, and each must be
      **rejected** (a `syscall.CreateFile` finding):

      | Row | Argument |
      |---|---|
      | string at `Args[0]` | `"p"` |
      | rune at `Args[0]` | `'p'` |
      | call-valued at `Args[2]` | `name()` |
      | string at `Args[3]` | `"sa"` |
      | rune at `Args[3]` | `'s'` |
      | call-valued at `Args[3]` | `name()` |
      | string at `Args[6]` | `"t"` |
      | rune at `Args[6]` | `'t'` |
      | call-valued at `Args[6]` | `name()` |

    * All nine rows **parse** (Stage probe, round 6). Each row fails against
      a predicate that skips the bare-operand check at its position. These
      rows are new, so the R5-1 parse-error
      set over the existing rows stays exactly `{:443, :444, :467, :468}`. If
      any new row fails to parse, that is an H-2 HALT, not a licence to drop
      the row.
  * **Round-5 amendment R5-2 (OP5-2 P3; NOT re-gated).** AC-D2.1
    (`TestOccurrenceAllowed_ReparseWindowsCreateFile`, `writepath_test.go:402–425`)
    is carried as follows. The live-call sub-case is carried by
    `TestCallExpr_ReparseWindowsCreateFile`, and the single-line allowed shape
    by the "allowed alone" case of `TestScanSource_AllowedCallCannotHideWritingCall`.
    The selector-keying sub-case becomes a `scanSource` assertion: a wrapped
    file calling `os.Remove(p, 0, 0, nil, syscall.OPEN_EXISTING,
    syscall.FILE_FLAG_BACKUP_SEMANTICS, 0)` (importing `"os"` and
    `"syscall"`) yields exactly one `os.Remove` finding on that line.
  * **Non-call row (Stage, round 4).** The "non-call reference" row
    (`f := syscall.CreateFile`, `:442`) has no `*ast.CallExpr`. It is ported
    as a `scanSource` assertion that the finding is reported.
  * **Wrapper scaffolding (round-4 amendment R4-6, OP4-6).** Every wrapper in
    the predicate port and in `TestScanSource_AllowedCallCannotHideWritingCall`
    is a whole file: `package p`, then an `import` of each package it
    references (`"syscall"`, `"os"`), then `func f() { … }`.
    * AC-E3.2 grants the allowance only when `X` resolves to import path
      `"syscall"`, so a wrapper without the import would turn "allowed alone"
      into a finding.
    * The import also keeps the wrappers valid under E-T4's import-path
      keying.
    * Expected finding lines are counted on the whole wrapped file.
* **Per-file import table (rev 7).** E-T3 introduces a table from the
  identifier name to the import path, built from `f.Imports`. An unnamed
  import binds `path.Base(path)`. E-T3 uses the table only for the `"syscall"`
  check. E-T4 and E-T7 extend the same table. They never introduce a second
  one, which is why the edges are linear.

**Acceptance criteria**
* AC-E3.1 — All Unit D predicate tests and fixtures keep their verdicts. The
  golden JSON stays **byte-identical**, and the positive-control row still pins
  only the writing lines.
* AC-E3.2 — Every D-2′ rejection is preserved (INV-2). In particular, only an
  INT `*ast.BasicLit` `0` access, an exact `syscall.OPEN_EXISTING` disposition
  and the exact `syscall.FILE_FLAG_BACKUP_SEMANTICS` flag are allowed, on a non-variadic call whose `X` resolves to
  import path `"syscall"`. Any widening HALTs (H-2 E). Every other argument
  must be a G-4 bare operand; STRING and CHAR `*ast.BasicLit` arguments are
  not bare operands and are rejected. The parseable `:464`–`:466` rows pin
  this rejection at the third argument (`Args[2]`). The nine round-6 R6-1
  rows pin STRING and CHAR rejection at `Args[0]`, `Args[3]` and `Args[6]`,
  and call-valued rejection at `Args[2]`, `Args[3]` and `Args[6]`. Together
  with `:462`–`:463` at `Args[0]`, every generic position has a
  non-bare-operand rejection row (closes PRE5-1).
* AC-E3.3 (rev 7 cycle 2, SB2-2) — The AC-D1.1, AC-D1.2 and AC-D1.3 contracts
  are carried by the named AST tests in the port table, with unchanged cases
  and verdicts, or a stricter fail-closed verdict for code-position separators
  and the non-parseable extent cases (GK3-1, R4-1) and predicate rows (R4-2).
  The predicate rows are probed and ported from **unmasked** source, and the
  probe reproduces exactly `{:443, :444, :467, :468}` (R5-1). The nine R6-1
  positional rows all parse and are rejected.
  AC-D1.3 includes the R4-4 discriminating raw-string case. The AC-D2.3 same-line contract is
  carried by `TestScanSource_AllowedCallCannotHideWritingCall` (SB3-1). After
  E-T3, every symbol in the complete deletion list has no definition and no
  caller. The check is
  `git grep -nE '\b(extractExtent|scanText|advanceCursor|lineReportsSelector|findSelector|occurrenceAllowed|firstExtent|absoluteOccurrences)\(|\b(extentKind|extentNonCall|extentBalanced|extentUnbalanced|extentMismatched|bracketOpener)\b|\btype extent\b'`
  over `tools/gatecheck/internal/writepath/`, which must be empty. It is scoped
  to that package because the `retiredarch` mutation table embeds `scanText(`
  inside Go source strings (SB3-4). `go build ./...`, `go vet ./...` and
  `go test ./tools/gatecheck/...` pass.
* AC-E3.4 (round 6, D-049-10; NOT re-gated) — The new AST allowance
  predicate does not reuse any name on the deletion list (R6-3). The
  `writepath_test.go` `gomask` import, which is unused after the R5-1 port, is
  removed (R6-5). The `:180` message prefix names `selectorOccurrences`
  (R6-2). With these, the AC-E3.3 grep is empty over correct code.

### E-T4 — 049.004-T (re-parented 034.006-T): Resolve named import aliases

Size **M**, complexity **medium** (re-sized after E-T1). Depends on E-T3. Rev 7
changed this from E-T2, to keep one shared import table and serialise the
`writepath.go` edits.

* **Spike citation:** G-5 (including its **additivity rule**), plus residual
  items 1, 3 and 7.
* **Files (4 artifacts; 7 physical files)**, counted under the D-T6 fixture
  counting convention: a fixture data file plus its golden row and
  stream-capture lines is one artifact. The artifacts are:
  * `writepath.go`, the only production file;
  * `writepath_test.go`;
  * the four new fixtures under `scripts/testdata/writepath/`, one artifact;
  * the golden JSON.

  The deviation from the fewer-than-3-files heuristic is recorded in the
  Constitution Check: the extra files are test data, and the fixtures and
  golden must land atomically (AC-E4.2). **Functions changed:** 2, the
  import-table builder and the selector matcher.
* **Test scenarios (4):** the four AC-E4.1 fixtures, as rows of the one
  table-driven golden self-test (`runFixtureSelfTest`). Under the D-T6
  convention that is one table-driven test with four rows; the plan
  still lists four scenarios, which is the upper edge of the heuristic.
* **Phase:** **red → green.** `reject-alias-os-writefile.go` and
  `reject-dot-import-os.go` are red against the E-T3 engine.
  `reject-split-selector.go` is green on arrival, because E-T2 closed item 7;
  it is declared as a regression pin. `accept-alias-copilot.go` is a negative
  control.

Map **named** aliases to canonical import paths via `*ast.ImportSpec` (extending
E-T3's table) and match selectors on the canonical path. Dot-imports and blank
imports follow the G-5 disposition; shadowing stays a residual.
* **Additivity rule (G-5, binding).** Alias resolution **adds** matches and
  never removes one.
  * A `SelectorExpr` whose `X` is spelled exactly as a selector's qualifier
    stays a finding even when that name is not bound to the canonical path.
  * Presence matching keyed only on import paths is forbidden, because it
    would fail open (H-2 E).
* **Dot imports.** A dot import of a write-capable package is reported at the
  import spec's line, using the canonical selector prefix (for example
  `os.*`). The import is the finding, so `Selectors` gains **no** bare-name
  entries for dot imports. The exact finding text is pinned by the
  `reject-dot-import-os.go` golden row.

**Acceptance criteria**
* AC-E4.1 — New fixtures have golden rows:
  * `reject-alias-os-writefile.go` (aliased `os`);
  * `reject-dot-import-os.go` (G-5);
  * `accept-alias-copilot.go` (the live `copilot` alias shape);
  * `reject-split-selector.go`: a selector split by a newline after `.` and by
    an inline comment. It pins D-T4 residual item 7, which the E-T2 AST engine
    closes. It lives here because E-T2/E-T3's golden must stay byte-identical.
* AC-E4.2 — The golden changes are **additive only**: new rows and the
  matching new `PASS` lines in the stream captures, updated atomically with the
  fixtures. Every pre-existing `selectors` and `fixture_findings` row stays
  unchanged.

### E-T5 — 049.005-T (re-parented 034.005-T): `Root.Resolve` first-caller tripwire

Size **M**, complexity **medium**. Depends on E-T4 (the tripwire must resolve
`internal/pathsafe` through aliases).

* **Spike citation:** G-6.
* **Files (4 artifacts; 5 physical files)**, counted under the D-T6
  convention:
  * `writepath.go`;
  * `writepath_test.go`;
  * two new fixtures under `scripts/testdata/writepath/`:
    `reject-resolve-first-caller.go` (constructs a root with
    `pathsafe.NewRoot` and calls `.Resolve(` on it in one function) and
    `accept-unrelated-resolve.go` (a `.Resolve(` call on an unrelated
    receiver, in a file that imports `pathsafe`);
  * the golden JSON.

  **Functions changed:** 1–2, the binding tracker and its call from the scan
  loop.
* **Test scenarios (3):**
  1. The tripwire fires on a fixture caller.
  2. It does not fire on an unrelated `.Resolve(` receiver.
  3. The tracked tree stays clean.
* **Phase:** **red → green.** Scenario 1 is red against the E-T4 engine.
  Scenarios 2–3 are negative controls.

Fire only when a file imports `internal/pathsafe` **and** calls `.Resolve(` on a
receiver bound from `pathsafe.NewRoot`. A bare `.Resolve(` match is
insufficient.
* **Binding strategy (G-6).** Syntactic tracking within one `*ast.FuncDecl` or
  `*ast.FuncLit`. It records identifiers bound by `x := pathsafe.NewRoot(...)`,
  `x, err := pathsafe.NewRoot(...)` or the multi-name declaration
  `var x, err = pathsafe.NewRoot(...)` (a `*ast.ValueSpec` with two names;
  `NewRoot` returns `(Root, error)`, so a single-name `var x = ...` does not
  compile, PR #99 review amendment), with the qualifier resolved through E-T4's table to
  `github.com/softwaresalt/intercom-go/internal/pathsafe`. It also records
  plain assignments (`token.ASSIGN`): `x = pathsafe.NewRoot(...)` and
  `x, err = pathsafe.NewRoot(...)`, for example after `var root pathsafe.Root`
  (PR #99 review amendment, **not re-gated**). The AC-E5 shape scan at
  `961b652` finds no plain-assignment or `var` binding in the tree today (recorded in
  `docs/decisions/assets/2026-10-02-writepath-go-ast-spike/h5-rev7-remeasure.txt`).
* **Round-4 amendment R4-8 (GPT4-2; NOT re-gated).** `NewRoot` returns
  `(Root, error)` (`internal/pathsafe/root.go:340`). The single-name forms
  `x := pathsafe.NewRoot(...)` and `x = pathsafe.NewRoot(...)` therefore do
  not type-check, just like the single-name `var x = ...`.
  * The tracker may still record them. Recording them is syntactic and only
    makes the tripwire fire more, which is the fail-closed direction.
  * The fixture and the AC-E5.1 callers use **only** the compiling two-name
    forms: `root, err := ...`, then `var root pathsafe.Root; var err error;
    root, err = ...`, then `var r, err = ...`.
* **No `go/types`.** Cross-function flows, struct fields, package-level
  variables and method values remain a residual, carried by E-T6.

**Acceptance criteria**
* AC-E5.1 (AC-4.4) — Fires on a fixture introducing a production
  `Root.Resolve` caller; does **not** fire on the tracked tree (zero production
  callers, re-measured at the parent commit). Scenario 1's fixture holds three
  callers, and each fires: one receiver bound by `:=`, one bound by plain
  assignment (`var root pathsafe.Root` and `var err error`, then
  `root, err = pathsafe.NewRoot(dir)`, then `root.Resolve(...)`; the `err`
  declaration is round-4 amendment R4-8), and one bound by
  `var r, err = pathsafe.NewRoot(dir)` then `r.Resolve(...)`. This stays within the three scenarios (PR #99
  review amendment, **not re-gated**).
* AC-E5.2 — Does not fire on an unrelated `.Resolve(` receiver.
* AC-E5.3 — Golden changes are additive only (as AC-E4.2).

### E-T6 — 049.006-T: Retire closed residuals from the gate documentation

Size **XS**, complexity **trivial**. Documentation only. Depends on E-T3, E-T5
and E-T7 (rev 7 added E-T7; the linear chain makes E-T7 the effective
predecessor).

* **Spike citation:** the residual-item disposition table.
* **Files (2):** the `writepath.go` package-doc header and the
  `scripts/check-write-path-precondition.sh` header. These are the two
  residual-record locations; no code changes.
* **Test scenarios (1):** the AC-D4.2 detector-scope check.
* **Phase:** green (documentation). The gate's existing tests must stay green.

**Rev 7 disposition**, in both locations:

| Item | Disposition |
|---|---|
| 1 | Remove; closed by E-T4 / `reject-alias-os-writefile.go`. |
| 2 | **Narrow, do not remove.** Same-function `NewRoot` → `Resolve` binding is closed by E-T5. The G-6 residual (cross-function flows, struct fields, package-level variables, method values) is retained. |
| 3 | Remove the dot-import clause, closed by E-T4 / `reject-dot-import-os.go`. State that blank imports are inert. Shadowing remains. |
| 6 | **Narrow, do not remove.** The 50 enumerated `ioutil`/`syscall`/`windows`/`unix` primitives are closed by E-T7 and its fixtures. The retained residual names the uncovered same-family remainder: metadata and attribute writes (`syscall.Chmod`/`Fchmod`/`Chown`/`Utimes`/`SetFileAttributes`, `unix.Fchmod`/`Chown`/`Fchown`/`Lchown`/`Utimes`/`Setxattr`, `windows.SetFileAttributes`), positional and vector writes (`syscall.Pwrite`, `unix.Writev`/`Pwritev`), `syscall.Ftruncate`, `syscall.Link`/`Symlink`, `unix.Mknod`/`Mknodat`, and any other un-enumerated write-capable symbol in those packages. Deferred as stash **C0D28448**. |
| 7 | Remove; closed by E-T2, pinned by E-T4's `reject-split-selector.go`. |
| 4, 5 | Remain verbatim. |
| 8 (new) | Add as a disclosed residual (deliberation §4.9): dynamic invocation via `syscall.NewLazyDLL` / `LazyProc.Call`, `syscall.Syscall*` and the `golang.org/x/sys` equivalents. These can reach any OS write API without a write-named selector. `syscall.NewLazyDLL` is live in production, so it cannot become a finding without an allowance design. Out of Unit E's scope; deferred as stash **FE2F02FF**. |

Also:
* Replace the residual-risk statement with the post-migration posture.
* Sync the detector-scope selector count and list in both headers with the
  widened `Selectors`, so AC-D4.2 holds again.

**Acceptance criteria**
* AC-E6.1 — Each removed or narrowed residual item cites the task and fixture
  that closed it, or closed part of it. Every retained item, and the retained
  part of each narrowed item (2 and 6), is still present verbatim. AC-D4.2's
  detector-scope check holds against the widened `Selectors`.
* AC-E6.2 — Item 8 is recorded in both locations, citing stash FE2F02FF.

### E-T7 — 049.007-T (rev 7, stash 458F9385): Widen the selector set to import-path-keyed write primitives

Size **M**, complexity **medium**. Depends on E-T5 for golden serialisation,
and on E-T4 for the alias table, which is transitive through E-T5. This
**narrows** D-T4 residual item 6 (review finding R4-P2-g) to the un-enumerated
same-family remainder; it does not close it (see E-T6 row 6 and stash
**C0D28448**).

* **Spike citation:** G-5, G-7 and the residual-item disposition table.
* **Deliberation:** §4.9.
* **Files (4 artifacts; 6 physical files)**, counted under the D-T6
  convention:
  * `writepath.go` (`Selectors` and its doc comment, plus the import table);
  * `writepath_test.go`;
  * three new fixtures under `scripts/testdata/writepath/`;
  * the golden JSON.

  **Functions changed:** 1, the import-table entries. `Selectors` is data.
* **Test scenarios (3):** the three AC-E7.1 fixtures, as rows of the one
  table-driven golden self-test. The existing tracked-tree scan is the
  clean-tree control.
* **Phase:** **red → green.** Each fixture is committed red (no finding under
  the E-T5 engine), then made green by the widening.

**Change.**
* **Selectors.** Append the following to the **end** of `Selectors`, in this
  order:
  * `ioutil`: `ioutil.WriteFile`, `ioutil.TempFile`, `ioutil.TempDir`.
  * `syscall`, the write and namespace calls other than the existing
    `syscall.CreateFile` and `syscall.Write`: `syscall.WriteFile`,
    `syscall.Open`, `syscall.Unlink`, `syscall.Rename`, `syscall.Mkdir`,
    `syscall.Rmdir`, `syscall.CreateHardLink`, `syscall.DeleteFile`,
    `syscall.MoveFile`, `syscall.RemoveDirectory`, `syscall.CreateDirectory`,
    `syscall.CreateSymbolicLink`, `syscall.Truncate`, `syscall.Creat`.
  * `windows` (`golang.org/x/sys/windows`): `windows.WriteFile`,
    `windows.CreateFile`, `windows.DeleteFile`, `windows.MoveFile`,
    `windows.MoveFileEx`, `windows.CreateDirectory`,
    `windows.RemoveDirectory`, `windows.CreateHardLink`,
    `windows.CreateSymbolicLink`, `windows.SetEndOfFile`,
    `windows.SetFileInformationByHandle`.
  * `unix` (`golang.org/x/sys/unix`): `unix.Open`, `unix.Openat`,
    `unix.Openat2`, `unix.Creat`, `unix.Write`, `unix.Pwrite`, `unix.Unlink`,
    `unix.Unlinkat`, `unix.Rename`, `unix.Renameat`, `unix.Renameat2`,
    `unix.Mkdir`, `unix.Mkdirat`, `unix.Rmdir`, `unix.Link`, `unix.Linkat`,
    `unix.Symlink`, `unix.Symlinkat`, `unix.Truncate`, `unix.Ftruncate`,
    `unix.Chmod`, `unix.Fchmodat`.

  That is 50 additions, for 76 in total.
* **Import table.** Extend E-T4's qualifier → canonical import-path table:

  | Qualifier | Import path |
  |---|---|
  | `ioutil` | `io/ioutil` |
  | `syscall` | `syscall` |
  | `windows` | `golang.org/x/sys/windows` |
  | `unix` | `golang.org/x/sys/unix` |

  The G-5 additivity rule applies uniformly: spelled-qualifier matching **or**
  path-resolved alias matching. Dot imports of these paths are findings as
  in E-T4. After E-T7 the table covers all **eight** `Selectors` qualifiers:
  `os`, `io`, `syscall`, `sql` (`database/sql`) and `bbolt`
  (`go.etcd.io/bbolt`) from Unit D, plus the new `ioutil`, `windows` and
  `unix`.
* **No allowance.** `windows.CreateFile` is always reported. The D-2′
  allowance stays `"syscall"`-only (G-4), so no predicate widening is implied.
* **Build tags.** `go/parser` ignores build constraints, so x/sys files are
  scanned on every `GOOS`. The repo has no x/sys dependency; the gate parses
  source and does not type-check it.

**Acceptance criteria**
* AC-E7.1 — New fixtures have golden rows:
  * `reject-ioutil-write-primitives.go`;
  * `reject-syscall-namespace-primitives.go`;
  * `reject-xsys-write-primitives.go`. Its `golang.org/x/sys/windows` import
    is aliased, so it proves path keying, and it imports
    `golang.org/x/sys/unix` unaliased.

  Each reports the expected selectors at the expected lines.
  `reject-ioutil-write-primitives.go` also carries a dot import,
  `. "golang.org/x/sys/unix"`. It is reported at the import spec's line with
  the canonical `unix.*` prefix (matching E-T4's `os.*`), which proves that the E-T4 dot-import rule
  extends to the new paths (G-5).
* AC-E7.2 — **The golden changes are additive only.**
  * The `selectors` array gains exactly the 50 entries, appended in order.
  * The new fixture rows and the matching new stream `PASS` lines are added.
  * Every pre-existing `selectors`, `fixture_findings`, stream and
    `filebased` byte stays unchanged.
  * The `len(Selectors) == len(golden.Selectors)` pin holds.
  * The D-T1a oracle stays bound to the frozen 20 and to AC-E2.6's frozen
    fixture list.
* AC-E7.3 — The tracked-tree scan of `internal/**` and `cmd/**` stays clean at
  the parent commit. Rev 7 recorded an exact-name scan of **all 50** appended
  selectors at `961b652`:
  * the command was the H-5 item-6 **inclusive variant** (the item-6 pattern,
    an anchored alternation of every appended `qualifier.Name` plus
    `golang.org/x/sys`, with the `':!*_test.go'` exclusion removed; recorded
    in H-5);
  * it ran over `internal/**` and `cmd/**` **including** `_test.go` files;
  * it found **zero** code-position hits. The only match is the comment at
    `internal/pathsafe/reparse_windows.go:15`, which mentions
    `golang.org/x/sys/windows`. The gate skips comments, so that is not a
    finding.

  Ship re-runs the inclusive variant at the parent commit. Any code-position
  hit is a HALT (H-2 E).

  The selector scan only sees the canonical qualifiers (`ioutil.`,
  `syscall.`, `windows.`, `unix.`). An aliased or dot import
  (`import u "io/ioutil"` then `u.WriteFile`) would evade it. So Ship also
  runs an **import-path scan** at the parent commit (PR #99 review
  amendment, **not re-gated**):
  `git grep -nE '"(io/ioutil|syscall|golang\.org/x/sys/(windows|unix))"' <parent> -- 'internal/**' 'cmd/**'`.
  Every hit must be an unaliased import spec. Any aliased or dot import spec
  of these paths is a HALT pending review (H-2 E). At `961b652` the scan lists
  four unaliased `"syscall"` imports (`internal/pathsafe/reparse_windows.go:8`
  and three `_test.go` files) and no aliased or dot import. The output is
  recorded in
  `docs/decisions/assets/2026-10-02-writepath-go-ast-spike/h5-rev7-remeasure.txt`.

---

## Constitution Check (Units D and E)

This maps Units D and E against `.github/instructions/constitution.instructions.md`.
Units A–C were checked in revisions 1–3.

| Principle | How Units D/E comply | Justified deviation |
|---|---|---|
| I. Safety-First Go | Gatecheck tool code only. INV-4 (lint clean, no dead code). The D-T1 cursor fails closed with no panic. E-T2 parse errors are exit 1. | None |
| II. Test-First (NON-NEGOTIABLE) | Red/green phases are declared per task (H-4). D-T2 has a genuine red phase (AC-D2.4). D-T1a, D-T5a, D-T5b and D-T6 are **declared** green on arrival as characterization or regression fixtures, with committed falsifiability via selector-naming golden rows. | Green-on-arrival fixtures are disclosed, not claimed as red phases |
| III. Workspace Isolation | No write primitive outside `testdata` (INV-3, AC-D5.5). The gate *is* the workspace-write tripwire. Units D/E only narrow what it allows (INV-2). | None |
| IV. CLI Workspace Containment (NON-NEGOTIABLE) | The residual-risk statement (D-T4) names every known open fail-open surface: items 1, 2, 6, 7. Deferral of alias and `Resolve` coverage is operator-approved (D-031-2) and bounded by 039-S, with the H-5 re-measure. **Rev 7:** Unit E closes items 1 and 7 and the item-3 dot clause, and narrows items 2 and 6. It discloses the new item 8 (dynamic proc invocation) rather than leaving it implicit. | Interval fail-open residuals are accepted by the operator, visible, and not hidden. Retained items 2 (narrowed), 3 (shadowing), 4, 5, 6 (narrowed; stash C0D28448) and 8 (stash FE2F02FF) are disclosed in E-T6. |
| V. Structured Observability | Finding text is unchanged (`write primitive '<sel>' found`). ED-2 `::error::` lines are retained. | None |
| VI. Single Responsibility | Each task covers one skill domain: code, docs or fixtures. Scenario counting is stated (table-driven = 1). **Rev 7:** file counts use the D-T6 artifact convention, so a fixture data file plus its golden row and stream lines is one artifact. The physical counts are stated alongside: E-T2 5 (R4-3; corrected in round 5, R5-4), E-T4 7, E-T5 5, E-T7 6. | E-T4/E-T5/E-T7 exceed the fewer-than-3-files heuristic in physical files only. The extra files are fixture data that must land atomically with the golden (AC-E4.2, AC-E7.2), and production code stays one file. |
| VII. Destructive Command Approval (NON-NEGOTIABLE) | No backlog item deleted. `return-blocked`/`adopt` were operator-directed (D-031-2) and recorded with an ID map (H-6). `WRITE_PATH_GATE_ADVISORY` stays operator-only (H-3). | None |
| VIII. Explicit Safety Modes | `WRITE_PATH_GATE_ADVISORY` is the explicit elevated-risk lever. Ship must not use it (H-3). | None |
| IX. Git-Friendly Persistence | Golden JSON and fixtures are committed; golden changes are atomic with fixtures. | None |
| X. Agent Context Efficiency | Tasks are S/XS within the 2-hour rule. Unit E stayed provisional until the spike, so there is no speculative detail. **Rev 7:** Unit E is now concrete, with M/S/XS sizing measured by G-7. E-T2's five files are an atomic engine-plus-test-pin set (G-1; the fifth file, a single literal edit, comes from round-4 amendment R4-3). | None |
| XI. Merge Commit History Preservation (NON-NEGOTIABLE) | One commit per task; rollback by revert in reverse dependency order (H-6). No squash or rebase is implied. | None |

## Constitution Check — Unit A (rev 6) and the P-002 deviation (CN-1)

| Principle | How Unit A (rev 6) complies | Justified deviation |
|---|---|---|
| I. Safety-First Go | The pin fails closed on every parse, read or rule failure. It has no panic path, no `go/types` in production, and no SA1019. | None |
| II. Test-First (NON-NEGOTIABLE) | A-T3a..A-T3d and A-T4 each declare a red phase against the prior pin, recorded per P-004. `scopeDataOK`'s pure-predicate rows are declared **compile-red**. | **P-002 skip for 034.001-T and 033.003-T only.** Item-scoped and operator-authorised (D-000-3). These are verified no-op closures that add no code, so they have no red phase. The skip is recorded at claim and disclosed in the closure artifact. It is not a general waiver. |
| III. Workspace Isolation | `pin.go` reads source files only. Test fixtures write only into `t.TempDir()`, which is Go-managed temp space. This follows the existing `writeMutatedCopy` fixture pattern (CN4-7). | None |
| IV. CLI Workspace Containment (NON-NEGOTIABLE) | The scan-scope guard becomes strictly stronger (INV-2). IVL-1 is bounded to the 030-S branch, with a merge precondition. Off-surface residuals are captured as P-021 entries `DC921AF6` and `D7BF9F74`. | IVL-1 is declared and bounded, not hidden. |
| V. Structured Observability | Self-test assertion names, order and success text are unchanged (AC-A4.2). Only the failure text gains the probe value. | None |
| VI. Single Responsibility | One skill domain per task. Each reject table is a single predicate over data rows, so it counts as one scenario (SC-3). | None |
| VII. Destructive Command Approval (NON-NEGOTIABLE) | No backlog item is deleted. The P-021 entries are added, not removed. | None |
| VIII. Explicit Safety Modes | No advisory lever is used or introduced. | None |
| IX. Git-Friendly Persistence | The canonical texts are committed in `pin.go`. The AC-A1.9 checklist is committed in the ALP-1 commit body. | None |
| X. Agent Context Efficiency | Every task is XS/S/M within the 2-hour rule, with explicit function and scenario budgets. | None |
| XI. Merge Commit History Preservation (NON-NEGOTIABLE) | One commit per task, and ALP-1 is one commit. Rollback is by revert in reverse order (H-6). | None |

**P-002 deviation entry (deviation format).**
* **Policy:** P-002 (test-first / `harness-ready` precondition).
* **Principle:** II, Test-First (NON-NEGOTIABLE).
* **Items:** 034.001-T (031-S) and 033.003-T (030-S).
* **Authority:** the operator, decision D-000-3 (2026-10-01).
* **Reason:** each is a verified no-op closure that adds no code, so no red
  phase exists.
* **Bound:** only these two items. Not transferable.
* **Evidence:** pre-task-completion gate evidence of prior satisfaction (see
  D-T0 and "Dropped from Unit A").
* **Disclosure:** the Ship closure artifact counts each as
  satisfied-by-prior-work.
* **Rejected alternative:** archive each item, or move it straight to a
  terminal state, without `active`. `.backlogit/hooks.yaml` does not allow
  `queued → done` or `queued → archived`, and a silent archive would hide
  the closure from shipment evidence.
* **Compliance:** Ship broadcasts the skip (P-005 visibility) and states it in
  the 030-S and 031-S PR bodies. The `harness-ready` precondition is **not
  applied** to these two items, and P-004 red capture is **not invoked**,
  because no test is written.
* **Closure of 030-S (LR4-6):** Ship closes the shipment only through
  classify-close-path → safe-close, after 033.003-T is `done`.

## Dependency graph and execution order

```
Unit A (030-S):  A-T1 → A-T2 → A-T3a → A-T3b → A-T3c → A-T3d
                          └──→ A-T4 ─────────────────────┘
                 (ALP-1 = {A-T1, A-T2}: one atomic commit;
                  A-T3a..A-T3d and A-T4: separate later commits — rev 6)
                 (033.001 → 033.002 → 033.004 → 033.005 → 033.006 → 033.007;
                  033.002 → 033.008 → 033.007;
                  033.003 verified no-op, D-000-2/D-000-3, no edge)

Unit B (032-S):  B-T1 → B-T2 → B-T3        (B-T1 is the test-first red)

Unit C (new):    C-T1 → C-T2 → C-T3        (C-T2 is the test-first red)
                 C-T1 → C-T4
                 C-T5  (independent)
                 C-T6  (independent)

Unit D (031-S):  D-T1a → D-T1 → D-T2 → D-T3 ─┬→ D-T4                (docs)
                                             └→ D-T5a → D-T5b → D-T6 (golden-serialised)
                 (034.010 → 034.002 → 034.003 → 034.004 → {034.007, 034.008};
                  034.008 → 034.011 → 034.009)
                 D-T0 (034.001-T) SATISFIED — detached from the chain; verified
                 no-op closure authorised (D-000-2); P-002 skip authorised (D-000-3)

Unit E (039-S):  [039-S blocks-on 031-S]   [049-F depends-on 034-F]
                 [spike ran after 031-S merged (961b652) — D-049-1, stash 1EEBECA5]
                 rev 7 (linear; one shared import table, serialised golden):
                 D-T6 → E-T1 → E-T2 → E-T3 → E-T4 → E-T5 → E-T7 → E-T6
                 (034.009 → 049.001 → 049.002 → 049.003 → 049.004 → 049.005
                  → 049.007 → 049.006; the rev 4 edges 049.004→049.002 and
                  049.006→049.003/049.005 are retained and are now redundant)
```

**Recommended shipment order: 038-S (C) → 032-S (B) → 030-S (A) → 031-S (D) →
039-S (E).**

* **Unit C first** — the only unit closing a *live* security-adjacent gap
  (`check-write-path-precondition.sh` has no `--root` guard today), and C-T1
  touches shared CLI dispatch that both other units' engines run through. Landing
  it first means A and B develop against the final dispatch contract.
* **Unit B second** — self-contained in one package, no design fork, lowest risk.
* **Unit A last** of the re-planned trio — the only unit that must mutate `pin.go` in lockstep with the
  code the pin asserts against (formalised as **ALP-1**, INV-5); benefits from a
  quiet tree, since its interior boundary is the plan's only sanctioned red gate.
* **Unit D (031-S) fourth** — independent of A–C at the file level (only
  `writepath` and its fixtures), but it benefits from C-T1's final dispatch
  contract and C-T2/C-T3's write-path wrapper guard, which share
  `check-write-path-precondition.sh`. D-T4 edits that wrapper's **header
  comment** only; C-T3 edits its **body** — sequencing C first avoids a textual
  conflict.
* **Unit E (039-S) last, and gated** — hard `blocks` edge on 031-S plus a STAGE
  HOLD pending the Stage-executed spike. The alias and `Resolve` residuals stay
  open for that interval (H-5); keep 039-S directly behind 031-S.

This supersedes the earlier post-M4 suggestion (032-S → 030-S → 031-S), which
predated both Unit C's discovery and the 031-S option fork.

---

## Plan Hardening

### H-1 — Blast radius

| Unit | Surface | Worst case |
|---|---|---|
| A | `retiredarch` scope declaration + its own pin | **Vacuously green pin.** Emptying `pathspecPinLiterals`/`prefixPinLiterals` to make the refactor land yields `containsAll(_, []) == true` (`pin.go:127`) — the pin passes while asserting nothing. Guarded by AC-A2.2. Secondary: ALP-1 split across two commits, leaving the pin gate red on `main`. Guarded by INV-5 + AC-A2.6. **Rev 5 (A-T3a/b/c):** *fail-open* if binding accepts a narrowed call because a dead literal survives, accepts a discarded, negated or mis-argued `HasPrefix`, misses a shadowed name or a mutation of `scanScope`/`S`, or resolves to an empty set and passes vacuously. Guarded by AC-A3a.1, AC-A3b.1 and AC-A3c.1. **Rev 6:** the use-whitelist is replaced by the frozen-declaration closed world. The new fail-open risks are a canonical text in `pin.go` edited in lockstep with `select.go` to admit a narrowing, and a mutation from another package file. An accidental or partial lockstep edit is caught by `scopeDataOK`'s separately authored data invariants (AC-A3c.2). A deliberate, coordinated `pin.go` + `select.go` edit is caught only by the behavioural self-test, including A-T4's independent `cmd/x/main.go` probe (AC-A4.2), and by human review of any `pin.go` diff (round 4, SEC4-5/AR4-1). Cross-file mutation is guarded by package closure (AC-A3d.1). *Fail-closed:* the pin rejects ALP-1's legitimate `select.go`, or any benign edit to a frozen declaration. That is loud by design. Guarded by AC-A1.8/AC-A1.9, the AC-A3b.2 positive controls and the H-2 A stop. |
| B | One function + its tests | Over-strict error converts a legitimately absent `.gitignore` into a merge-blocking failure. Bounded to one gate; guarded by AC-B2.2. |
| C | `parseRoot`, shared by **all four** engines | A regression breaks every gate at once. Guarded by AC-C1.3's four real self-test entry points. |
| D | `writepath` scanner — the **only** mechanical enforcement of the write-path precondition (merge-blocking at `ci.yml:333-347` unless the operator sets `WRITE_PATH_GATE_ADVISORY`) plus its fixtures/golden. No product code. | **Fail-open:** D-T1's per-occurrence enumeration drops or mis-lines a hit, or D-T2's predicate allows a writing or file-creating `syscall.CreateFile` (non-zero access, non-`OPEN_EXISTING` disposition, `DELETE_ON_CLOSE`, smuggled composite/raw-string arguments, or a same-line pairing) — a silent hole in a security gate. Guarded by AC-D1a.1 (frozen differential oracle over corpus and live tree), AC-D2.2/AC-D2.3 (rejection table and same-line test), AC-D5.1 (fixture-level positive control and evasion corpus), and the golden. **Fail-closed:** line/order drift reddens golden or the repo scan — loud, blocks merges until fixed. **Collateral:** `retiredarch`'s writepath scan-loop pin — kept unmodified by D-1′ (AC-D1.4). |
| E | Same scanner, **engine swap** | **Fail-open:** the tag-shaped-raw-string rule, a partial-AST-on-parse-error scan, `//line`/`\n`-only line drift, or a predicate *widening* during the port flips a reject to accept. Guarded by G-2/G-3/G-4, AC-E2.1–E2.3, AC-E3.2 and the Unit D parity corpus. Collateral: deleting `gomask` would break `retiredarch` — forbidden by the Unit E invariant; the `retiredarch` writepath pin is re-anchored by E-T2 (G-1). **Rev 7:** *fail-open* if E-T4's alias resolution or E-T7's path keying **replaces** spelled-qualifier matching instead of adding to it (G-5 additivity rule), or if E-T5's binding tracker silently drops the bare tripwire. *Fail-closed:* E-T7's widening flags a live tracked-tree identifier. The rev 7 measurement at `961b652` found none, so this is guarded by AC-E7.3 and is loud by design. *Collateral:* E-T7 changes the `Selectors` length. The `len(Selectors)==len(golden.Selectors)` pin forces an atomic golden append, and the oracle stays bound to the frozen 20. |

### H-2 — Named stop conditions

Halt and return to Stage rather than deciding:
* **A** — if `scanScope` cannot express all three arms (including the
  prefix-less `config.toml.example`) without a special case; or if AC-A2.3's
  negative control cannot be made to fail; or if re-anchoring the pin would
  require deriving the expected literals **from `select.go`'s own text** (which
  *would* make it a true self-comparison). Do **not** empty the pin literal lists
  and do **not** weaken the pin to make the refactor land. **A-T3a..A-T3d and
  A-T4 (rev 6):** HALT if ALP-1's landed `select.go` is not token-equal to
  §A-CANON (AC-A1.8). HALT if any of them would exceed its declared function or
  test budget. Do **not** fall back to presence, do **not** edit `select.go`,
  and do **not** edit the canonical text to match a different shape.
* **B** — if `git show`'s exit/stderr shape cannot distinguish invalid-ref from
  absent-file portably. Do **not** satisfy AC-B2.1 by making absent-file an error.
* **C** — if first-wins `parseRoot` changes any existing engine's verdict or exit
  code. Do **not** proceed by relaxing AC-C1.3.
* **D** — if D-1′ or D-2′ cannot be implemented exactly as specified (AC-D1.5,
  AC-D2.6); if D-T1a's frozen oracle (restricted to the frozen 20 selectors)
  shows **any** divergence; if `retiredarch`'s writepath pin would need editing;
  or if adding the six selectors makes the bare repo scan fail. Do **not** edit
  the oracle or golden rows for pre-existing fixtures, do **not** widen any D-2′
  token or allowlist, and do **not** attempt alias or receiver tracking in this
  unit (that is Unit E's scope).
* **E** — if the spike cannot establish fail-closed parse handling (including
  partial ASTs) or full verdict parity (including both tag-shaped fixtures and
  `SplitLines` line numbering); if E-T2/E-T3 would change **any** golden byte, or
  E-T4/E-T5 would change any **pre-existing** golden row; if the predicate port
  would allow anything D-2′ rejects; or if the migration would require deleting
  or modifying `gomask`. **Rev 7 additions:**
  * HALT if E-T4 or E-T7 would key presence matching on import paths **only**
    (G-5 additivity).
  * HALT if E-T7 would remove or reorder any existing `Selectors` entry,
    change any pre-existing golden byte (AC-E7.2), or extend the D-2′
    allowance beyond `"syscall"`.
  * HALT if E-T7's widening makes the tracked-tree scan fail (AC-E7.3). Do
    **not** drop the colliding selector to make it pass; return it to Stage.
  * HALT if E-T6 would delete a residual item that rev 7 marks as narrowed
    or retained (items 2, 3-shadowing, 4, 5, 6 and 8). Items 2 and 6 are
    narrowed, not removed.
* **All** — if any task would introduce a write primitive into `internal/**` or
  `cmd/**` (INV-3), or would leave the tree lint-dirty (INV-4), or would require
  an atomic landing pair not declared in INV-5.

### H-3 — Controls that must not be weakened to make work land

* **AC-A2.2** (pin literals non-empty) and **AC-A2.3** (committed negative
  control) exist specifically because the refactor's easiest failure mode is a
  green pin that pins nothing. Neither may be relaxed. **AC-A3a.1, AC-A3b.1,
  AC-A3c.1/AC-A3c.2 and AC-A3d.1** (rev 6) extend this to the frozen surface.
  The A-T3 tasks may not satisfy their reject tables by weakening AC-A2.2 or
  AC-A2.3, by deriving the canonical text or the expected sets from
  `select.go`, or by asserting only `!OK()`.
* **AC-B1.2** (inverted assertion must fail pre-change) exists because the unit
  could otherwise be reported complete while `git_test.go:103-121` still pins
  fail-open behaviour.
* **AC-C5.3** (C-T5 changes no behaviour) exists because "fixing" item (3) would
  change the verdict of a currently-passing invocation shape without an
  authorising acceptance criterion (INV-1's explicit-AC carve-out).
* **INV-5 / AC-A2.6** (ALP-1 lands as one commit) exists because the cheapest way
  to make A-T1 "pass" on its own is to weaken or empty the pin. The pair boundary
  is where the pin must be proven, and it may not be moved earlier.
* **AC-D1a.1** (frozen differential oracle) exists because "line numbers
  unchanged" is otherwise asserted only for the handful of golden fixtures, and
  the `SplitLines`-vs-`\n` hazard would pass them while mis-lining a live
  finding. It is restricted to the frozen 20 selectors (exact set lookup) so
  that it is never edited in Unit D; its **only** permitted adaptation is
  AC-E2.6 (production side re-pointed to the G-1 boundary, input list frozen by
  name, unparseable inputs moved to stricter fail-closed assertions) — an oracle
  that is edited whenever it disagrees is no oracle.
* **AC-D2.2 / AC-D2.3 / AC-D5.1** (rejection table, same-line test, positive
  control pinning only the writing lines) exist because a file-level or
  first-occurrence exemption passes every "metadata call is allowed" test.
* **AC-D1.4** (`retiredarch` pin unmodified) exists because editing a
  cross-package architectural guard to make a refactor land is exactly the
  weakening H-3 forbids.
* **AC-D4.1** (residual-risk statement names 049-F) exists because deferring
  alias and receiver coverage is only acceptable if it is **visible** to every
  reviewer of the gate in the interval.
* **AC-E2.1 / AC-E3.1** (golden byte-identical across E-T2/E-T3) and
  **AC-E4.2 / AC-E5.3** (additive-only thereafter) are the parity contract;
  editing a pre-existing golden row to make the migration land defeats the Unit D
  corpus.
* **`WRITE_PATH_GATE_ADVISORY`** — Ship must **not** set it, request it, or rely
  on it to land any Unit D or Unit E work; a D/E PR that is green only because
  the repo-scan step was advisory is not green. It is an operator-only
  emergency lever (H-6). *Recommendation (requires operator approval before it
  becomes policy):* the operator logs each use with a restore condition and
  re-scans any commit merged while it was set with the bare gate before
  clearing it.

### H-4 — Green-on-arrival declared honestly

* **A-T1's derivation-agreement property is green-on-arrival** — the two
  derivations already agree today. The genuinely new assertions are AC-A1.5
  (asymmetry preservation) and A-T2's AC-A2.3 (negative control), which are red
  against a naive flat-prefix unification.
* **A-T2's re-anchoring half is red** (the pin breaks the moment A-T1 moves the
  literals, by design — `select.go:5-9`); its non-vacuity guard (AC-A2.2) is
  green-on-arrival today and exists to stay green.
* **A-T3a..A-T3d and A-T4 (rev 6) each have a declared red phase.**
  * A-T3a's pure-predicate test is red against today's `containsAll`, which
    returns `true` on an empty `wanted`.
  * The reject tables follow the rev-6 red-phase rule (Unit A):
    * A-T3b's rows are red against the presence pin;
    * A-T3c's rows are red against A-T3b's pin;
    * A-T3d's rows are red against A-T3c's pin.
  * `scopeDataOK`'s pure-predicate rows are **compile-red**, declared.
  * A-T4's test is red against a helper that holds only the two pre-existing
    probes.
  * AC-A3d.2's universe-coverage test is **compile-red**, declared.
  * **Positive controls are green on arrival** and are declared as regression
    guards, not red phases (round 4, CN4-1): AC-A3b.2, AC-A3c.3's live-tree
    acceptance, and AC-A3d.1's `select.go`-only fixture.
  * The red output is recorded (P-004). The code that turns it green lands in
    the same commit.
* **B-T1 is the red phase** for Unit B; B-T2 turns it green.
* **B-T3, C-T4, C-T5 and C-T6 are documentation-only** and claim no red phase.
* **C-T1 has a genuine red phase** (test-first on `main_test.go`).
* **C-T2 is the red phase** for the wrapper guard; **C-T3 is the green phase.**
  Neither is green-on-arrival: AC-C2.3 requires C-T2's rejection cases to fail
  against the pre-C-T3 wrapper, and AC-C3.4 requires C-T3 to turn them green.
  C-T3 must therefore be verified, not assumed.
* **D-T1a is green on arrival by design** — a characterization oracle that
  pins today's behaviour. **D-T1's** new extent and cursor assertions (AC-D1.1–
  AC-D1.3) are red before the extractor exists. **D-T2 has a genuine red phase**
  (AC-D2.4). **D-T3's red** is the golden `selectors` edit (AC-D3.1).
  **D-T4 is documentation-only.** **D-T5a, D-T5b and D-T6 fixtures are green on
  arrival** — declared, with committed falsifiability via selector-naming golden
  rows (AC-D5.2, AC-D5b.2, AC-D6.2).
* **Unit E** red/green phases are declared by the post-spike re-plan; E-T1 is
  read-only.

### H-5 — Residuals carried, not closed

* ~~`DD0BB60F`~~ — **harvested in rev 5 as A-T3a (033.004-T), A-T3b
  (033.005-T) and A-T3c (033.006-T)** (D-030-4), and **re-designed in rev 6**
  (D-030-6) with A-T3d (033.007-T) and A-T4 (033.008-T). The stash entry is
  archived.
  * **IVL-1:** from ALP-1 until A-T3b, the pathspec check is presence only.
    Until A-T3c, the prefix half is presence only. Until A-T3d, other package
    files are unchecked. IVL-1 is bounded by the edges and by the 030-S merge
    precondition, so `main` never observes it.
  * After A-T3d, the rev-5 per-arm control-flow and cross-file mutation
    residuals are **closed**. The remaining residuals R-A1..R-A4 (Unit A) are
    off the contract surface. **R-A1 → stash `DC921AF6`** and **R-A2 → stash
    `D7BF9F74`** are P-021 DEFERRED SCOPE EXPANSION entries, which require
    deliberation before any planning (P-021 C6).
* `C312BD4C` — the `gomask` multiline-tag contradiction (`gomask.go:138-141`
  comment vs `gomask_test.go:57` pinned behaviour) stays open; deliberation
  D-030-3 re-scopes it to a comment fix.
* `50E6F22C` item (3) — extra-positional-argument tolerance stays open pending a
  new ED entry (AC-C5.3).
* `5A8EC1BC` — three `GitRunner` shapes and exported mutable
  `Denylist`/`Selectors` stay open; trigger not fired.
* **Deferred alias coverage (D-031-2) — KNOWN OPEN fail-open surface from 031-S
  ship until 039-S ship.** An aliased or dot import of any package whose
  qualifier appears in `Selectors` (`os`, `io`, `syscall`, `database/sql`,
  `go.etcd.io/bbolt`) evades the gate. *Likelihood:* low — zero instances at `9b299c8`; the sole alias in scope
  is `copilot` (not write-capable), and Go convention rarely aliases those
  packages. *Impact:* high — a write path could land without tripping the
  precondition, so the pathsafe TOCTOU/hard-link register triggers would fire
  unnoticed. *Compensating controls:* AC-D4.1's residual record in the gate's own
  header, PR review, and E-T4.
* **Deferred `NewRoot`/`Resolve` tracking (D-031-2) — KNOWN OPEN over the same
  interval.** The first production `Root.Resolve` caller lands unmonitored, so
  feature 038-F's trigger and the register's "real write path" triggers stay
  *awaited*, not *monitored*. *Likelihood:* medium — product work (043-F) will
  eventually add callers. *Impact:* medium — 038-F's black-box verification and
  the TOCTOU mitigation decision are skipped until someone notices.
  *Compensating controls:* the residual record, PR review, and E-T5. 038-F
  now carries a `blocks` edge on 049-F (Harvest Record), so its trigger cannot
  be mistaken as satisfied by 031-S.
* **Split selector (D-T4 item 7) — KNOWN OPEN until E-T2** (pinned by E-T4's
  `reject-split-selector.go`). Pre-existing; affects all selectors. *Likelihood:*
  very low (gofmt-hostile, never idiomatic). *Impact:* high. *Compensating
  controls:* the residual record and PR review.
* **Selector coverage gap (D-T4 item 6) — KNOWN OPEN until 039-S ships;
  narrowed, not closed, by E-T7 (049.007-T) in rev 7.**
  `ioutil.*` write helpers, `syscall` write/namespace calls beyond the two
  selected, and `golang.org/x/sys` equivalents are undetected. *Likelihood:*
  low (none at `9b299c8` or `961b652`; `ioutil` is deprecated). *Impact:*
  high. Widening the selector set was outside D-031-2's operator-approved
  scope. It was captured as stash entry **458F9385** (rev 5, D-S-5) and folded
  into 039-S / 049-F at the rev 7 post-spike re-plan (deliberation §4.9). The
  stash entry is archived.
  E-T7 covers the 50 enumerated primitives. The uncovered same-family
  remainder (see E-T6 row 6) **stays KNOWN OPEN after 039-S ships** and is
  deferred as stash **C0D28448**.
* **Dynamic proc invocation (new D-T4 residual item 8, rev 7) — KNOWN OPEN,
  not scheduled.** This covers `syscall.NewLazyDLL`/`LazyProc.Call`,
  `syscall.Syscall*` and the `golang.org/x/sys` equivalents. They reach any OS
  write API with no write-named selector.
  * *Likelihood:* low. The only live use is `internal/pathsafe`'s
    metadata-only `GetFinalPathNameByHandleW` probe (`reparse_windows.go:19`).
  * *Impact:* high.
  * A selector cannot cover it without an allowance design for that live use,
    so it is disclosed by E-T6 and captured as low-priority stash entry
    **FE2F02FF** (deliberation §4.9).
  * *Compensating controls:* the residual record, PR review, the
    `reparse_windows.go` file being under `internal/pathsafe` review, and the
    item-8 re-measure below.
* **Interval bound and re-measurement — POLICY (rev 5, operator decision
  D-031-5).** This replaces a CI tripwire. Neither deferred residual has a time
  bound if 039-S stalls. **Every Stage session re-measures until 039-S ships**,
  and records the command and verbatim output in that session's memory:
  `git grep -nE '^[[:space:]]*(import[[:space:]]+)?([A-Za-z_][A-Za-z0-9_]*|\.)[[:space:]]+"(os|io|syscall|database/sql|go\.etcd\.io/bbolt)"' -- 'internal/**' 'cmd/**' ':!*_test.go' | grep -v 'return "'`
  (package set derived from the `Selectors` qualifiers; trailing comments
  tolerated; the `grep -v` drops `return "io"`-style string literals such as
  `internal/apperr/apperr.go:67`) and
  `git grep -nE '\.Resolve\(|pathsafe\.NewRoot\(' -- 'internal/**' 'cmd/**' ':!*_test.go' ':!internal/pathsafe/**'`
  and, for item 6 (rev 7: the full 50-name E-T7 set; it supersedes the rev-5
  seven-name subset),
  `git grep -nE '(^|[^A-Za-z0-9_.])(ioutil\.(WriteFile|TempFile|TempDir)|syscall\.(WriteFile|Open|Unlink|Rename|Mkdir|Rmdir|CreateHardLink|DeleteFile|MoveFile|RemoveDirectory|CreateDirectory|CreateSymbolicLink|Truncate|Creat)|windows\.(WriteFile|CreateFile|DeleteFile|MoveFile|MoveFileEx|CreateDirectory|RemoveDirectory|CreateHardLink|CreateSymbolicLink|SetEndOfFile|SetFileInformationByHandle)|unix\.(Open|Openat|Openat2|Creat|Write|Pwrite|Unlink|Unlinkat|Rename|Renameat|Renameat2|Mkdir|Mkdirat|Rmdir|Link|Linkat|Symlink|Symlinkat|Truncate|Ftruncate|Chmod|Fchmodat))([^A-Za-z0-9_]|$)|golang\.org/x/sys' -- 'internal/**' 'cmd/**' ':!*_test.go'`
  and, for item 8 (rev 7),
  `git grep -nE 'NewLazyDLL|NewProc|Syscall[0-9]*\(' -- 'internal/**' 'cmd/**' ':!*_test.go'`
  (POSIX `[[:space:]]`, not GNU `\s`, per compound lesson
  `2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md`; the
  item-6 command's only current hit is the comment at
  `internal/pathsafe/reparse_windows.go:15`, re-verified at `961b652`. The
  item-8 command's expected hits are `reparse_windows.go:15/:19/:20` and
  `internal/pathsafe/root.go:308`. Lines 15 and `root.go:308` are comments,
  `:19` and `:20` are the live probe. Any other hit is a trigger).
  At `9b299c8` the first is empty and the second lists only `internal/config/validate.go`
  (`pathsafe.NewRoot` calls at `:42`/`:67` and a comment at `:32`); any new hit is a
  trigger to promote 039-S to the head of the queue. A blocking CI tripwire was
  considered and **not** adopted: it would re-implement the deferred 034.006-T /
  034.005-T scope in text form, contrary to D-031-2. With D-031-5 the operator
  makes the per-session re-measure the compensating control, which closes
  R4b-P2-5. **Rev-5 run (2026-10-01, HEAD `52f15d4`):** the first command is
  empty; the second lists only `internal/config/validate.go:32/:42/:67`; the
  third lists only the comment at `internal/pathsafe/reparse_windows.go:15`.
  There is no new trigger. The verbatim output is in the session memory.
  **Rev 7 run (2026-10-02, HEAD `961b652`):** the first command is empty
  (positive control: the regex matches the live `copilot` alias when the
  package set includes it). The second lists only
  `internal/config/validate.go:32/:42/:67`, with zero `.Resolve(` callers. The
  third (rev 7 full 50-name item-6 command) lists only the comment at
  `internal/pathsafe/reparse_windows.go:15`. The **inclusive variant** (the
  same command with the `':!*_test.go'` pathspec removed, run as
  `git grep -nE '<same item-6 pattern>' origin/main -- 'internal/**' 'cmd/**'`
  at `961b652`) gives the same single hit, so `_test.go` files add no
  code-position match. The fourth (item 8) lists exactly the four
  expected hits. There is no new trigger. The commands and their verbatim
  output are recorded in
  `docs/decisions/assets/2026-10-02-writepath-go-ast-spike/h5-rev7-remeasure.txt`
  (re-run at `961b652` for the PR #99 review amendment, which corrected an
  earlier pointer to the session memory doc).

### H-6 — Rollback

* **Units A–C** — revert the task commit (ALP-1 as its single commit). No data
  or schema state is involved. **Unit A (rev 6):** revert in reverse order:
  A-T3d, A-T4, A-T3c, A-T3b, A-T3a, then ALP-1. Reverting ALP-1 while the A-T3
  tasks remain would leave a frozen-declaration pin whose canonical text no
  longer matches `select.go`. That fails closed (loud, and merge-blocking).
* **Unit D** — every task is its own commit; **revert in reverse dependency
  order** (D-T6 → D-T5b → D-T5a → D-T4 → D-T3 → D-T2 → D-T1 → D-T1a). The one
  unsafe partial revert is **reverting D-T2 while D-T3 remains**:
  `syscall.CreateFile` would then be a selector with no allowance, and the bare
  repo scan fails on `reparse_windows.go:164` — fail-closed, but merge-blocking.
  Revert D-T3 with or before D-T2. Emergency lever (operator-only, not a Ship
  action): the `WRITE_PATH_GATE_ADVISORY` repository variable makes the repo-scan
  step advisory while the integrity self-test stays blocking (`ci.yml:333-347`).
* **Unit E** — roll back by **landed dependency closure**: revert E-T6, E-T7,
  E-T5, E-T4, E-T3, then E-T2 (with their fixtures, golden additions and the
  `retiredarch` re-anchoring). E-T7's selectors and golden rows are appended,
  but it is **not** independently revertible once E-T6 has landed: E-T6's two
  headers list the widened 76-selector set, and AC-D4.2 would then fail.
  Revert E-T6 first, or revert E-T6 and E-T7 as a pair. Between the E-T7 and
  E-T6 merges the header count (26) lags `Selectors` (76). That gap is
  shipment-internal and is closed by E-T6 before 039-S completes. AC-D4.2 is a
  **manual review check**, not an automated test: rev 4 declined a
  header-vs-`Selectors` consistency test as new scope. The lag therefore fails
  no gate, and the drift window is accepted (rev 7 cycle 2, SB2-4). Only at the
  immediate E-T2 boundary — before E-T3
  deletes the masked-text extractor — is a single revert sufficient. `gomask` is
  retained so the masked engine can be restored cleanly; the Unit D corpus is
  the parity contract in **both** directions. `WRITE_PATH_GATE_ADVISORY` does
  not bypass the blocking integrity self-test, so it is not a substitute for
  rollback.
* **Backlog** — `shipment return-blocked` removal is reversible with
  `shipment add 031-S <id>` while the item is in no other shipment.
  `backlogit adopt` **renumbers** IDs (034.006-T → 049.004-T, 034.005-T →
  049.005-T); re-adopting back to 034-F would renumber again, so the ID map in
  the deliberation (§4.5) is the rollback record. **Harvest read-back
  obligation** (compound lessons `2026-09-04-backlogit-sizing-is-wit-gated.md`,
  `2026-05-07-backlogit-shipment-status-constraints.md`): every backlog mutation
  is run with stderr visible and its exit code checked, then read back
  (`shipment get`, `dep list`, item files); the resulting ID map and edge
  directions are recorded in the session memory.

---

## Verification

Gate scripts must be invoked at their **real entry points**. Neither
`check-unignore-regression.sh` nor `check-gitignore-append-only.sh` may be run
bare — both fail closed with exit 1 when no ref arguments are supplied.

```bash
go build ./...
go test ./tools/gatecheck/... -count=1
golangci-lint run ./...
staticcheck ./...

scripts/check-retired-architecture.sh
scripts/check-retired-architecture.sh --self-test
scripts/check-retired-architecture.sh --self-test-integrity

scripts/check-write-path-precondition.sh
scripts/check-write-path-precondition.sh --self-test
scripts/check-write-path-precondition.sh --self-test-integrity

scripts/check-merge-strategy.sh --self-test
scripts/check-merge-strategy.sh

scripts/check-unignore-regression.sh --self-test
scripts/check-unignore-regression.sh --base-ref <base> --head-ref <head>

scripts/check-gitignore-append-only.sh --self-test
scripts/check-gitignore-append-only.sh --base-ref <base> --head-ref <head>
```

Each unit additionally verifies its selected-path / verdict set against **its own
parent commit**, per INV-1.

**ALP-1 exception (INV-5).** The full suite above is expected to be red at the
A-T1 → A-T2 interior boundary, bounded to the tests exhaustively enumerated in
the ALP-1 declaration. Run the full suite at the **pair boundary** (AC-A2.6),
not at A-T1.

**Test-first red phases (not INV-5 pairs).** `go test ./tools/gatecheck/...` is
also expected red at the B-T1 → B-T2 and C-T2 → C-T3 boundaries, bounded to the
assertions those tasks author (AC-B1.2, AC-C2.3). No *gate script* changes
verdict at either boundary, and both land as separate commits.

**Golden/fixture coupling (Units D and E).** `writepath_golden.json` pins
`selectors`, every `fixture_findings` row, and the `--self-test` /
`--self-test-integrity` stream captures (one `PASS` line per fixture). Any
selector or fixture addition (D-T3, D-T5a, D-T5b, D-T6) **must** update the golden in
the same commit, or `go test ./tools/gatecheck/...` goes red. E-T2 and E-T3 must
leave the golden **byte-identical** (AC-E2.1, AC-E3.1); E-T4 and E-T5 may only
**add** rows and their stream-capture `PASS` lines (AC-E4.2, AC-E5.3). E-T7 may
only **append** its 50 selectors to the `selectors` array and add its fixture
rows and `PASS` lines (AC-E7.2). D-T2's red phase is a `go test` red
only; no gate script changes verdict at that boundary because the predicate is
dormant until D-T3 adds `syscall.CreateFile`.

---

<!-- plan-review-attempt: 3 -->
<!-- revision: 3 — round 2 (FAIL) remediations: P1-1 ALP-1/INV-5, P2-1 AC-A1.7,
     P3-1 C-T2/C-T3 inversion, P3-2 B-T2 interim note, P3-3 AC-C5.3 rationale,
     P3-4 citation drift, P3-5 scenario count.
     Round 3 (ADVISORY, 9xP3) residuals all applied: INV-5 test-first exclusion,
     exhaustive ALP-1 red enumeration + AC-A2.7, AC-C2.2 reword, C-T4 fifth
     comment, AC-A1.7 widened to select.go:4-9, H-4 green-phase correction,
     C-T1 cross-reference, frontmatter revision bump, C-T2 sizing note. -->
<!-- GATE: plan-review round 3 = ADVISORY. No P1/P2 outstanding. Proceeds to
     harvest under the ADVISORY disposition with operator awareness. -->

## Plan Review — Revision 4 (Units D and E only)

<!-- plan-review-attempt-rev4: 1 -->

### Round 1 — FAIL

dispatch_mode: multi-agent
decision: FAIL

Six personas were dispatched as subagents. Five ran on the caller's model:
Constitution Reviewer, Go Reviewer, Scope Boundary Auditor, Learnings
Researcher and Security Lens Reviewer. Security Lens was triggered because the
write-path gate is a security control. The sixth, Architecture Strategist, ran
through the `model_routing.anchor_review` route (`gpt-6.1-sol`, xhigh). Plan
hardening was required and present: H-1..H-6.

Merged counts after deduplication: **1 P0, 6 P1**, about 12 P2, and P3s.

| ID | Sev | Source | Finding | Remediation (rev 4, round-1 fixes) |
|---|---|---|---|---|
| R4-P0-1 | P0 | Security | Paren-only depth lets brace composites or tag-shaped raw strings fake D-2's 7-segment shape | D-1′ `()[]{}` stack; D-2′ bare-operand rule; evasion fixture (D-T5a) |
| R4-P1-1 | P1 | Constitution, Go, Security | Access `0` + `CREATE_*`/`OPEN_ALWAYS` creates files; `DELETE_ON_CLOSE` unchecked | D-2′ rules 6–7; table and fixture rejects; G-4 carries them |
| R4-P1-2 | P1 | Constitution, Go | Differential oracle breaks at D-T3/D-T5 (legacy copy has no allowance) | Oracle restricted to a **frozen 20-selector** set; split into 034.010-T |
| R4-P1-3 | P1 | Security | First-occurrence-only `findSelector` lets an allowed call hide a writing call on the same line | Per-occurrence predicate, de-duplication afterwards (AC-D2.3, D-T5a fixture) |
| R4-P1-4 | P1 | Security | G-2 too weak: partial AST scanned on parse error; `//line` shifts positions | G-2 any-error = exit 1, no partial scan, `PositionFor(pos,false)`, `//line` fixture |
| R4-P1-5 | P1 | Architecture | `retiredarch/writepath_mask_test.go` pins `scanText`'s top-level `SplitLines` loop; D-1's whole-text loop breaks it | D-1′ keeps the loop + byte cursor; pin must pass unmodified (AC-D1.4); E-T2 owns re-anchoring (G-1) |
| R4-P1-6 | P1 | Learnings | `go/token` counts `\n` only — line-number hazard for Unit E | Named G-3 hazard; map through `SplitLines`; oracle inputs retained |
| R4-P2-a | P2 | Constitution, Scope, Architecture | "Unit E golden byte-identical" contradicts E-T4/E-T5 fixture additions | Byte-identical for E-T2/E-T3; additive-only for E-T4/E-T5 |
| R4-P2-b | P2 | Constitution, Scope, Go | D-T1 and D-T5 exceed the 2-hour rule | D-T1 → D-T1a + D-T1; D-T5 → D-T5a + D-T5b |
| R4-P2-c | P2 | Go | No fixture pins tag-shaped raw strings outside struct fields | `reject-tag-shaped-raw-string-expr.go` (D-T5a) |
| R4-P2-d | P2 | Architecture, Go | `scanTextWith` is not an engine-swap boundary; masked text cannot parse | Seam dropped from Unit D; G-1 defines a source-level boundary |
| R4-P2-e | P2 | Architecture | E-T3 AST port could widen the allowance | G-4 / AC-E3.2: representation-only; widening HALTs |
| R4-P2-f | P2 | Architecture, Constitution | Unit E rollback by single revert is wrong once E-T3+ land | H-6 rollback by landed dependency closure |
| R4-P2-g | P2 | Security, Constitution | Uncovered write primitives (`ioutil.*`, other `syscall`, `x/sys`) unrecorded | D-T4 residual item 6; H-5 entry; stash candidate (rev 5: captured as stash **458F9385**, D-S-5) |
| R4-P2-h | P2 | Security | `WRITE_PATH_GATE_ADVISORY` could be used to land D/E work | H-3 control: unset while D/E in flight; logged; re-scan |
| R4-P2-i | P2 | Learnings, Security | Re-measure grep too narrow, unrecorded | Broadened grep; recorded output; marked as a recommendation pending operator approval |
| R4-P2-j | P2 | Learnings | D-T0 queued no-op affects closure | Named prerequisite for 031-S closure |
| R4-P2-k | P2 | Learnings | Bypass review loop risk on the 031-S PR | AC-D4.4 review-loop terminator |
| R4-P2-l | P2 | Security | No automated tripwire for the deferred residuals | **Declined, with rationale.** It would re-implement deferred 034.005-T/034.006-T in text form, contrary to D-031-2. The re-measure recommendation stands. |
| R4-P2-m | P2 | Learnings | Backlog mechanics (return-blocked/adopt/edge direction) need read-back | Applied at harvest; tool output recorded in the session memory |

P3s applied:
* findSelector copy;
* the fixed finding text;
* content-located tests;
* `strings.TrimSpace`;
* reverse-order rollback wording;
* the "reverting D-T2 while D-T3 remains" wording;
* the redundant edge D-T5 → E-T1 removed;
* D-T5's dependency on D-T4 replaced with D-T3;
* E-T1 re-plan obligations (AC-E1.2);
* the `Ellipsis` and import-path checks (G-4);
* the dot-import fail-closed rule (G-5).

P3s acknowledged and not applied:
* a header-vs-`Selectors` consistency test (AC-D4.2 already requires the
  match; a test would be new scope);
* a CI `::notice::` for residuals.

<!-- plan-review-attempt: 2 -->

### Round 2 (re-entry 1 of 2) — FAIL

dispatch_mode: multi-agent
decision: FAIL

The same six personas were re-dispatched against the round-1 remediation. The
Architecture Strategist again ran on the `anchor_review` route (`gpt-6.1-sol`,
xhigh).

**Round-1 results:**
* Every round-1 P0/P1 was confirmed **RESOLVED** by the persona that raised it:
  R4-P0-1, R4-P1-1 to R4-P1-6.
* Security Lens found no valid Go that D-2′ exempts while writing.
* The Go and Architecture reviewers confirmed the D-1′ cursor is byte-correct,
  and that the `retiredarch` pin passes **unmodified**.
* The Architecture reviewer confirmed the task DAG is correct and minimal.

**Persona verdicts:**

| Persona | Model | Outstanding P0/P1/P2/P3 | Verdict |
|---|---|---|---|
| Security Lens | caller | 0 / 0 / 3 / 2 | ADVISORY |
| Go Reviewer | caller | 0 / 0 / 3 / 4 | ADVISORY |
| Constitution | caller | 0 / 0 / 1 / 5 | ADVISORY |
| Architecture Strategist | anchor (`gpt-6.1-sol`) | 0 / **1** / 0 / 0 | FAIL |
| Scope Boundary Auditor | caller | 0 / 0 / 1 / 7 | ADVISORY |
| Learnings Researcher | caller | 0 / 0 / 1 / 3 | ADVISORY |

**Round-2 findings and dispositions:**

| ID | Sev | Source | Finding | Remediation |
|---|---|---|---|---|
| R4b-P1-1 | P1 | Architecture (also Go N3 P2, Constitution N-1 P2, Learnings P2) | D-T1a's oracle is bound to masked `scanText`, but G-1 replaces that boundary. Alias fixtures and code-position `\f`/`\v`/`\u2028` inputs (parser errors under G-2) break "stays green unedited". | Oracle lifecycle: legacy side frozen; **single authorised adaptation AC-E2.6** (production side → G-1 on unmasked source; fixture list frozen by name at 031-S merge; unparseable inputs → stricter G-2 fail-closed assertions); G-3 names the hazards; H-3 updated |
| R4b-P2-1 | P2 | Security, Go | Rule 3's quote check cannot see interpreted strings or runes (gomask blanks the delimiters); the AC-D2.2 "string argument" row was unmeetable | New **rule 3b** (non-empty segments); rule 3 reworded; AC-D2.2 rows split |
| R4b-P2-2 | P2 | Go | A substring frozen-20 filter would count `io.CopyN` as `io.Copy` | Exact set lookup on the parsed selector token (D-T1a) |
| R4b-P2-3 | P2 | Security | A selector split by a newline or comment evades every selector; the risk statement understated the fail-open set | D-T4 **residual item 7**; risk statement names items 1, 2, 6, 7; H-5 entry; E-T4 `reject-split-selector.go` |
| R4b-P2-4 | P2 | Scope | D-T2's scenario count exceeds the 2-hour rule | Table-driven = one scenario (stated); D-T2 = 3 scenarios. **Not split**: the rules share one evaluator, and a split would leave a half-predicate commit |
| R4b-P2-5 | P2 | Security | Declining the CI tripwire relies on an unapproved re-measure | **Accepted as open:** the re-measure recommendation is surfaced as an operator decision at hold-lift |
| R4b-P3 | P3 | various | Rule 7 under-specified; G-4 constants' `X`; AC-D2.3 before D-T3; `findSelector` disposition; in-comment boundary case; cursor fail-closed; D-T1 dead code; item-6 grep; deliberation D-T0 mismatch; D-031-3 acknowledgment field; D-T5/D-T6 file counting; `os.Chtimes` confirmation; `FILE_FLAG_OPEN_REPARSE_POINT` speculative; H-3 operator duties; AC-D4.4 not task-verifiable; `[[:space:]]`; D-T0 disclosed deviation; read-back obligation | All applied. Rule 7 is narrowed to exact `FILE_FLAG_BACKUP_SEMANTICS`. H-3's operator logging is now a recommendation. AC-D4.4 is moved to Ship guidance. `os.Chtimes` is flagged for operator confirmation. G-5 dot-import fail-closed is confirmed at re-plan |

<!-- plan-review-attempt: 3 -->

### Round 3 (re-entry 2 of 2, final) — ADVISORY

dispatch_mode: multi-agent
decision: ADVISORY

The personas that owned round-2 findings were re-dispatched: Architecture
Strategist on the anchor route (`gpt-6.1-sol`, xhigh), Security Lens, Go
Reviewer, and Constitution Reviewer. Scope Boundary and Learnings had no
round-2 P0/P1, and their round-2 P2/P3 items were all remediated.

**Round-2 P1 and P2 status:**
* **R4b-P1-1 is RESOLVED**, confirmed by the Architecture reviewer (anchor
  route). The tracked-tree inputs stay live as extra regression coverage, and
  any divergence HALTs.
* Every round-2 P2/P3 is resolved, or accepted as an open operator decision
  (R4b-P2-5).
* The live `reparse_windows.go:164-172` call still passes every D-2′ rule,
  including rule 3b and the narrowed rule 7 (Security Lens).

| Persona | Model | P0 / P1 / P2 / P3 | Verdict |
|---|---|---|---|
| Architecture Strategist | anchor (`gpt-6.1-sol`) | 0 / 0 / 0 / 1 | ADVISORY |
| Security Lens | caller | 0 / 0 / 1 (accepted-open R4b-P2-5) / 2 | ADVISORY |
| Go Reviewer | caller | 0 / 0 / 1 / 3 | ADVISORY |
| Constitution | caller | 0 / 0 / 1 / 4 | ADVISORY |

**Round-3 advisory findings, all applied after the verdict (no re-review
needed at ADVISORY):**
* **Go P2.** The cursor-mismatch rule said "ED-2 `::error::`", which the
  `[]string` signature cannot carry. The rule now disables the allowance for
  the rest of the text, which fails closed as a false positive. The case is
  unreachable through `scanFile`, and a direct-call invalid-UTF-8 test pins it.
* **Constitution P2.** The plan had no Constitution Check section. Added:
  **Constitution Check (Units D and E)**.
* **P3s:**
  * E-T6 retires residual item 7;
  * `findSelector` has a named production caller (fast-path guard);
  * hand-built oracle inputs are complete Go files;
  * invalid UTF-8 handling is moved to `ReadText` both-sides-error;
  * the selector token is parsed from the right;
  * the lifecycle paragraph lists the unparseable-expectations adaptation;
  * the G-3 oracle bullet is now "confirm or HALT";
  * the deliberation's D-031-3 row is updated.

**Gate outcome.** The verdict is ADVISORY with **no P0/P1 outstanding**, which
meets the operator's D-031-2 hold-lift condition. The **031-S STAGE HOLD is
LIFTED**.

The remaining open items are operator decisions, not plan defects:
* acknowledging D-031-3;
* approving the H-5 re-measure (on which declining the CI tripwire rests,
  R4b-P2-5);
* `os.Chtimes` in D-T6;
* the terminal state for 034.001-T and 033.003-T;
* the selector-set widening (item 6).

> **Rev 5 — all five resolved by the operator (2026-10-01):** D-031-4
> (D-031-3 acknowledged), D-031-5 (H-5 re-measure is policy; closes
> R4b-P2-5), D-031-6 (`os.Chtimes` confirmed), D-000-2 (verified no-op
> closure authorised for both items), and D-S-5 (item 6 captured as stash
> 458F9385).

The 049-F STAGE HOLD stays in place until the Go-era spike re-plan, as designed.

---

## Harvest Record — Revision 4 (2026-10-02)

Every mutation used `C:\Tools\backlogit.exe` v1.11.0 with stderr visible and
exit codes checked, followed by a read-back (`shipment get`, `dep list` and the
item files). This satisfies H-6. All expected IDs were assigned exactly as
planned.

| Plan unit | Backlog ID | Disposition |
|---|---|---|
| D (feature) | 034-F | Retitled "Harden write-path precondition gate (masked-text increment)". The `blocked` status moved to `queued` (**hold lifted**). Edges to 030-F/032-F retained (D-000-1). |
| D-T0 | 034.001-T | Retitled `SATISFIED (post-M4 Go re-plan) …`. Kept `queued` in 031-S with **no live edges** (edge 034.002→034.001 removed). Lifecycle: no queued→terminal transition in `hooks.yaml`, so closure is a named operator prerequisite (shared with 033.003-T). |
| D-T1a | 034.010-T | NEW, S |
| D-T1 | 034.002-T | Rewritten (D-1′), S |
| D-T2 | 034.003-T | Rewritten (D-2′), S |
| D-T3 | 034.004-T | Rewritten (six selectors; folds B72E9715), S |
| D-T4 | 034.007-T | Rewritten (seven residuals + residual-risk statement), S |
| D-T5a | 034.008-T | Rewritten, M→S |
| D-T5b | 034.011-T | NEW, S |
| D-T6 | 034.009-T | NEW, S (folds 8E9F8E55) |
| E (feature) | 049-F | NEW. Status `blocked` = the **STAGE HOLD** pending the spike re-plan. |
| E-T1 | 049.001-T | NEW, XS |
| E-T2 | 049.002-T | NEW, M |
| E-T3 | 049.003-T | NEW, S (rev 7 cycle 2: re-sized M; see "Plan Review — Revision 7 (Unit E)", cycle 2, SB2-2) |
| E-T4 | 049.004-T | **was 034.006-T**: `shipment return-blocked` from 031-S → `move --status queued` → `adopt --parent 049-F`. Body rewritten, M. |
| E-T5 | 049.005-T | **was 034.005-T**: same path as E-T4. Body rewritten, M. |
| E-T6 | 049.006-T | NEW, XS |
| E (shipment) | 039-S | NEW, "Migrate write-path scanner to go/ast (post-M4 unit E)". |

**Final manifests (read back).**
* 031-S, retitled "Harden write-path gate, masked-text increment (post-M4 unit
  D)": `034-F, 034.001-T, 034.002-T, 034.003-T, 034.004-T, 034.007-T,
  034.008-T, 034.010-T, 034.011-T, 034.009-T`. Manifest order is not execution
  order; Ship follows the edges, and 034.010-T is the chain root.
* 039-S: `049-F, 049.001-T … 049.006-T`.

**Edges (read back; `A → B` means A depends on B, type `blocks`).**
* Removed:
  * 034.002→034.001, 034.007→034.006, 034.008→034.005 and 034.008→034.007;
  * the adopt-rewritten 049.004→034.004 and 049.005→034.003.
* Added:
  * 034.002→034.010, 034.007→034.004, 034.008→034.004, 034.011→034.008 and
    034.009→034.011;
  * 049.001→034.009, 049.002→049.001, 049.003→049.002, 049.004→049.002,
    049.005→049.004, 049.006→049.003 and 049.006→049.005;
  * 049-F→034-F and **039-S→031-S**;
  * **038-F→049-F** (see below).

**Consequence found at harvest — 038-F re-pointed.** 038-F ("Black-box verify
`Root.Resolve` once a live caller exists") names "the `Root.Resolve`
first-caller tripwire delivered by 034.005-T" as its trigger, and its only edge
pointed at 034-F. The split moved that tripwire to 049.005-T. Stage therefore:
* added edge 038-F→049-F, keeping 038-F→034-F as historical record;
* appended a marked description section recording the re-identification.

The legacy text and the `blocked` status are unchanged. Without this, 038-F
would read as unblocked once 031-S ships, even though its trigger mechanism
does not exist yet.

**Stash.** 8E9F8E55 → 034.009-T and B72E9715 → 034.004-T (with the
`reject-io-copyn-copybuffer.go` presence fixture in 034.011-T). Each entry's
text was edited to carry the reconciled `PR=#71` and its harvest target. Both
were then **archived** (`backlogit stash archive`, non-destructive).

**Degraded tooling.** `update --complexity` fails ("task does not define a
complexity field"). Complexity is recorded as prose
(`Size: X | Complexity: y`) in every task description. Size is set as a
structured field (`size_source: agent`, `size_ruleset_version: 2h-rule-v1`).

---

## Plan Review — Revision 5 (Unit A)

<!-- plan-review-attempt-rev5: 3 -->

* `dispatch_mode: multi-agent`
* `scope: Unit A only` (A-T1, A-T2, A-T3a, A-T3b, A-T3c, AC-A1.8, "Dropped from
  Unit A"). Units B–E are not re-gated by this round.
* `decision: FAIL`

**Rounds 1 and 2 (reconstructed).** An earlier Stage pass on this same request
was interrupted before it committed anything or wrote memory. Its review
transcripts were lost. Both rounds are reconstructed from the rev-5 plan text
and counted as **FAIL**, which is the conservative reading:
* **Round 1.** One A-T3 task breached the 2-hour rule (Go P1-3, Scope F1).
  This led to the first split.
* **Round 2.** The pathspec task was still over budget once the shared
  whitelist checker was counted (Go N5, Scope N3). This led to the 3-way split,
  the use-whitelist and AC-A1.8.

Round 3 is therefore re-entry 2 of 2, the **final** allowed cycle.

**Round 3 personas (2026-10-01).**

| Persona | Route | Verdict | P1 | P2 | P3 |
|---|---|---|---|---|---|
| Go Reviewer | caller | ADVISORY | 0 | 4 (GO-1..GO-4) | 7 |
| Scope Boundary Auditor | caller | ADVISORY | 0 | 4 (SC-1..SC-4) | 4 |
| Constitution Reviewer | caller | ADVISORY | 0 | 3 (CN-1..CN-3) | 5 |
| Learnings Researcher | caller | ADVISORY | 0 | 0 | 6 (LR-1..LR-6) |
| Architecture Strategist | anchor `gpt-6.1-sol` / openai / xhigh | PASS-level | 0 | 0 | 2 (AR-1, AR-2) |
| Security Lens | caller | **FAIL** | **5 (SEC-1..SEC-5)** | 2 (SEC-6, SEC-7) | 1 |

**P1 findings (all from the Security Lens).** Each is a fail-open narrowing of
the scan scope that the rev-5 contract would still **accept**.
* **SEC-1 — the containment rule is unsound.** The rule is expected ⊆
  resolved. Two cases defeat it:
  * a git exclude pathspec (`:(exclude)cmd/**`, `:!`, `:^`) **subtracts**
    from the selection while adding an element;
  * `shouldScanRepoPath` takes the first matching prefix, so an earlier,
    broader or duplicate arm can shadow a later one.

  Fix: reject magic pathspecs and duplicate or overlapping prefixes, or
  require exact set equality.
* **SEC-2 — the compensating control has a gap.** The behavioural self-test
  (`selftest_selection.go`) does not probe non-test `cmd/**/*.go`, yet the
  "per-arm control flow" residual relies on it. Fix: add a
  `shouldScanRepoPath("cmd/x/main.go")` probe whose expectation is independent
  of `scanScope`, or correct the residual text.
* **SEC-3 — the `selectRepoPaths` shape is unpinned.** The contract binds the
  arguments of `git(root, S...)` but not the rest of the function. It still
  accepts:
  * the call placed in dead code;
  * `root` reassigned;
  * the output discarded;
  * a second, alternate runner or `exec` call that produces the output;
  * extra filtering beyond `shouldScanRepoPath`.

  Fix: pin `out, err := git(root, S...)` at the top level, with `root` never
  assigned, `out` consumed only via `pysem.GitText`, and `shouldScanRepoPath`
  as the only filter.
* **SEC-4 — the `DefaultGitRunner` body is unpinned** (`select.go:31-46`).
  It can drop or rewrite its pathspec arguments without tripping the pin.
* **SEC-5 — prefix-function control flow outside the loop is unanalysed.**
  None of the following trips the pin:
  * an early `return` before the loop;
  * `goto`;
  * a second loop;
  * `defer`;
  * a falsifying operand allowed by the `&&` chain;
  * an `if` body that returns `false`.

**P2 findings.**
* **GO-1 / SC-2.** The file-wide `S` name rule collides with the `args` and
  `pathspecs` locals in `DefaultGitRunner`. Scope it to the `selectRepoPaths`
  body.
* **GO-2.** `make(..., len(scanScope))` is a natural spelling, but the
  whitelist forbids it. Decide explicitly.
* **GO-3 / SC-1 / CN-4.** Ownership of the use-whitelist rules between A-T3b
  and A-T3c is ambiguous, and so is which `OK` flag a violation clears.
* **GO-4.** A discarded `git` call (`_, _ = git(...)`) has no reject row. This
  overlaps SEC-3.
* **SC-3.** The reject-table row count strains the one-scenario convention.
  Trim it or justify it.
* **SC-4 / CN-6.** The AC-A1.8 checklist should be recorded in the ALP-1
  commit body.
* **CN-1.** The Constitution Check has no rev-5 Unit A rows, and the
  item-scoped P-002 skip has no deviation-format entry. The persona also
  proposed archiving 033.003-T as an alternative.
* **CN-2.** The P-002 skip reading for the no-op closures needs a hard
  precondition and a fallback. *Applied* to 034.001-T and 033.003-T as
  fail-closed ACs (AC-D0.4, AC-X3.4).
* **CN-3.** ALP-1 opens an interval in which the `scanScope` presence pin
  accepts mutations that today's pin rejects. That interval needs:
  * a bounded INV-2 declaration;
  * a 030-S merge precondition: A-T3b/A-T3c are done, or the residual is
    re-captured (DD0BB60F is now archived).
* **SEC-6, SEC-7.** Hardening of the reject tables. These are subsumed by the
  SEC-1/SEC-3 remediations.

**P3 findings** (GO-5..GO-11, SC-5..SC-8, CN-4..CN-8, LR-1..LR-6, AR-1, AR-2,
SEC-8) are editorial or advisory. **LR-6** proposes landing ALP-1 + A-T3a
alone and deferring A-T3b/A-T3c. They are recorded for the remediation pass and
are not applied in this revision, except as noted.

**Gate outcome — FAIL, final re-entry, 030-S back on hold.** The merged
verdict is the most severe persona verdict: **FAIL**.
* The operator's decision-7 condition is ADVISORY or better with no P0/P1.
  That condition is **not met**.
* **033-F stays `blocked`.** That was also its actual frontmatter status
  before this session: the rev-4 lift was recorded in prose only.
* 033-F and 030-S carry a **STAGE HOLD RE-IMPOSED** description section.
* No fourth review round was run. The plan-review cycle limit is reached.

**Escalation Protocol (consecutive planning failures).**
* **Payload:**
  * threshold: rev-5 plan-review attempt 3, consecutive FAIL;
  * summary: SEC-1..SEC-5;
  * artifacts: this plan, deliberation D-030-4, 033.004-T..033.006-T;
  * resumption: the session memory, "Follow-up session".
* **Route.** No Stage-specific escalation override is declared, so the route
  falls back per field to `claude-opus-5`.
* **Degraded mode.** No engram handoff surface is configured for this session.
  The protocol therefore resolves to **`ESCALATION_DEGRADED`** and falls back
  to operator halt.

**Remediation options for the operator** (Stage recommends **(a)**):
* **(a)** Stage remediates SEC-1..SEC-5 and the P2s in A-T3's contract
  (rev 6). The operator then authorises **one further** plan-review round.
* **(b)** Narrow the scope per LR-6. Land ALP-1 + A-T3a + the 033.003-T no-op
  now, and hold A-T3b/A-T3c behind (a). This still requires a re-gate of the
  narrowed Unit A and a CN-3 interval declaration.
* **(c)** Accept SEC-1..SEC-5 as recorded residuals (an operator risk
  acceptance), then re-gate.

---

## Harvest Record — Revision 5 (2026-10-01)

Every mutation used `C:\Tools\backlogit.exe` v1.11.0 with exit codes checked,
followed by a read-back (`shipment get 030-S`, `dep list`, the item files).
Body edits used `--section name=value` only. Afterwards, `backlogit sync`
indexed 393 artifacts and `backlogit doctor` reported "No issues found".

**Harvest under a FAIL verdict.** Stage does not normally harvest a plan that
failed review. Decision 7 explicitly directs the harvest, and it makes only the
**hold** conditional on the verdict. Stage therefore created the tasks **under
the held feature**: 033-F is `blocked`, and each task carries a STAGE HOLD
paragraph and a HOLD acceptance criterion. Ship cannot claim them, and their
acceptance criteria are marked "expected to change" at remediation.

| Plan unit | Backlog ID | Disposition |
|---|---|---|
| A (feature) | 033-F | Status stays `blocked`, re-confirmed as the STAGE HOLD. A `description` section is prepended with **STAGE HOLD RE-IMPOSED**, and the rev-4 "HOLD LIFTED" text is marked historical. |
| A (shipment) | 030-S | A `description` section records the STAGE HOLD. Manifest read back: `033-F, 033.001-T, 033.002-T, 033.003-T, 033.004-T, 033.005-T, 033.006-T`. |
| A-T3a | 033.004-T | NEW, XS, low, test-first. Includes the `containsAll(_, [])` vacuous-pass hazard at `pin.go:127`. |
| A-T3b | 033.005-T | NEW, M, medium, test-first. Reject row (a) is "narrowed call plus retained dead literal". |
| A-T3c | 033.006-T | NEW, S, medium, test-first. Reject row (b) is the prefix analogue. |
| Dropped | 033.003-T | The D-000-2 no-op closure authorisation is added to `description`, and `acceptance-criteria` AC-X3.1..AC-X3.5 are set. |
| D-T0 | 034.001-T | The D-000-2 authorisation is added to `description`, and `acceptance-criteria` AC-D0.1..AC-D0.5 are set (fail-closed on the P-002 confirmation). |
| E (feature) | 049-F | A `description` section records the D-049-1 spike timing, the 1EEBECA5 trigger, the blocked_stale backstop, the 458F9385 fold-in recommendation and the operator's `go/ast` rationale. |
| E-T1 | 049.001-T | The same timing and rationale are added, with `acceptance-criteria` AC-E1.1..AC-E1.4. The plan reference is updated from rev 4 to rev 5. |

**Edges added** (`A → B` means A depends on B, type `blocks`):
* 033.004-T → 033.002-T
* 033.005-T → 033.004-T
* 033.006-T → 033.005-T

None was removed. None is added for 033.003-T, by design.

**Stash.**
* **DD0BB60F**: its text was edited to carry the harvest targets
  (033.004-T/005-T/006-T, D-030-4, the 030-S hold). It was then **archived**
  (non-destructive).
* **458F9385** (task, low; residual item 6): NEW and active.
* **1EEBECA5** (spike, high; the post-031-S spike trigger): NEW and active.

**Degraded tooling.** `--complexity` is still unsupported, so complexity is
recorded as prose.

---

## Plan Review — Revision 6 (Unit A)

<!-- plan-review-attempt-rev6: 4 -->

* **dispatch_mode:** multi-agent (six personas in parallel).
* **decision:** **ADVISORY**. No P0 and no P1.
* **Authority.** This is the **operator-authorised exception** to the
  3-cycle cap (D-030-6). It is the fourth and final round for Unit A. **No
  round 5** will be run, whatever the outcome.
* **Scope.** Unit A (A-T1..A-T4, the rev-6 Pin contract and §A-CANON), the D-T0
  and "Dropped from Unit A" P-002 text, and the Constitution Check for Unit A
  (rev 6). Units B..E were not re-reviewed; they are unchanged since rev 5.

| Persona | Route | Verdict | P0 | P1 | P2 | P3 |
|---|---|---|---|---|---|---|
| Security Lens | caller (claude-opus-5.5 / anthropic / high) | ADVISORY | 0 | 0 | 3 (SEC4-1..3) | 2 (SEC4-4, SEC4-5) |
| Go Reviewer | caller | ADVISORY | 0 | 0 | 2 (GO4-1, GO4-2) | 6 (GO4-3..8) |
| Scope Boundary Auditor | caller | ADVISORY | 0 | 0 | 1 (SC4-1) | 6 (SC4-2..7) |
| Constitution Reviewer | caller | ADVISORY | 0 | 0 | 2 (CN4-1, CN4-2) | 6 (CN4-3..8) |
| Learnings Researcher | caller | ADVISORY (confidence medium) | 0 | 0 | 3 (LR4-1..3) | 6 (LR4-4..9) |
| Architecture Strategist | **anchor** (gpt-6.1-sol / openai / xhigh) | **PASS** | 0 | 0 | 0 | 2 (AR4-1, AR4-2) |

**Round-3 findings closed.** Each persona re-checked its own round-3
findings against rev 6.
* **SEC-1..SEC-7 — CLOSED** (Security Lens).
  * SEC-1: the shared rules and token equality reject extra, `:(exclude)`
    and duplicated arms. `scopeDataOK` rejects magic, duplicate and
    shadowing values.
  * SEC-2: A-T4's `cmd/x/main.go` probe.
  * SEC-3: `selectRepoPaths` is frozen.
  * SEC-4: `DefaultGitRunner` is frozen. The runtime environment is residual
    R-A2 (`D7BF9F74`).
  * SEC-5: `shouldScanRepoPath` is frozen.
* **GO-1..GO-4 — CLOSED** (Go Reviewer).
* **SC-1..SC-4 — CLOSED** (Scope Auditor).
* **CN-1..CN-3 — CLOSED** (Constitution).

**P2 findings — all applied in this revision** (text amendments inside the
reviewed design; no new mechanism):
* **SEC4-1** — non-Go sources other than `.s`/`.syso` (for example `.swig`)
  → the `go/build` `ImportDir` allowlist and a `.swig` row in AC-A3d.1.
* **SEC4-2** — the "what changed" box over-claimed "which paths reach the
  scan" → reworded to "which paths are *selected*". `scanPath`/`engineForPath`
  dispatch is added to R-A1 and to `DC921AF6`.
* **SEC4-3** — environment redirection (`GIT_DIR`, `GIT_INDEX_FILE`,
  `GIT_WORK_TREE`, `GIT_CONFIG_*`, `PATH`) → the in-package
  `os.Setenv/Unsetenv/Clearenv` rule and AC-A3d.1 row (package closure). The
  out-of-package part is added to R-A2 and `D7BF9F74`.
* **GO4-1** — token-equality mechanics → parse and index by name, slice via
  `fset.File.Offset`, compare `go/scanner` streams (Pin contract).
* **GO4-2** — `includeTests` was not in `scopeDataOK` → `armValues` carries
  it, with the `includeTestsPinPrefixes = {"cmd/"}` rule and two AC-A3c.2 rows.
* **SC4-1** — the universe cross-check was toolchain-fragile → a
  toolchain-stable coverage check (AC-A3d.2), and the related stop condition
  is removed.
* **CN4-1** — the positive controls were implied as red phases → declared
  green-on-arrival regression guards (red-phase rule, H-4). AC-A3d.2 is
  declared compile-red.
* **CN4-2** — the IVL-1 precondition was not an acceptance criterion →
  AC-A3d.6, plus the 030-S shipment description.
* **LR4-1** — pin the rewritten `select.go` comment wording (§A-CANON).
* **LR4-2** — reject rows could pass for the wrong reason → AC-A3d.1 also
  asserts `SelectFound && GuardFound` and a clean parse of the extra file.
* **LR4-3** — the `scopeDataOK` guarantee needed honest limits → the
  "What these rules guarantee, and what they do not" paragraph.

**P3 findings.**
* **Applied:** SEC4-5/AR4-1 (H-1 A row and the guarantee paragraph; A-T4 is
  *needed*); AR4-2 (AC-A1.3); SC4-2 (belt-and-braces label); SC4-3 (the reason
  for the A-T3d → A-T4 edge); SC4-5 ("needed" wording); SC4-6/GO4-5 (the
  unreadable row is a directory named `x.go`); SC4-7 (AC-A3c.3 reuses the
  AC-A3b.2 table; the IVL-1 AC); GO4-6 (stable universe check); GO4-7 (R-A1);
  CN4-3 (deviation entry: Principle II, rejected alternative, compliance);
  CN4-4 (AC-A3b.1 "authored and observed red"); CN4-5 (INV-2 IVL-1 exception);
  CN4-6 (ALP-1 staged as one commit); CN4-7 (Principle III row); LR4-4
  (stdlib evidence note); LR4-5 (R-A1/R-A2 exist identically at `9b299c8`);
  LR4-6 (030-S closes via classify-close-path → safe-close); LR4-8 (compound
  citation); LR4-9 (no self-match evidence).
* **Recorded, not applied** (editorial or Ship-time judgement): SEC4-4,
  GO4-3, GO4-4, GO4-8, SC4-4, CN4-8.
* **LR4-7 — dispositions of rev-5 LR-1..LR-6.** LR-1 (the P-002 hard
  precondition) is discharged by D-000-3. LR-2..LR-5 (P3, advisory, not
  individually recorded in rev 5) are subsumed by the rev-6 contract rewrite
  and by the round-4 learnings pass (LR4-1..LR4-9), which re-surveyed the
  compound library against rev 6. LR-6 (narrow to ALP-1 + A-T3a) is
  **declined**: the operator chose option (a) (D-030-6).

**Gate outcome — ADVISORY; the 030-S hold is LIFTED (D-030-7).**
* The operator's condition is met: ADVISORY or better with no P0/P1.
* 033-F moves `blocked → queued`, and its STAGE HOLD section is replaced with
  HOLD LIFTED. The 030-S description section is replaced the same way.
* The do-not-claim notes on 033.004-T, 033.005-T and 033.006-T are
  superseded by HOLD LIFTED description sections. The legacy free-text
  preambles cannot be edited through `--section`. Each new section states
  that the preamble hold text is historical and void.
* The P2 amendments were applied after the round without a re-review. They
  are text inside the reviewed design. A fifth round is not authorised.
* **Escalation:** not triggered (no FAIL).

---

## Harvest Record — Revision 6 (2026-10-01)

Every mutation used `C:\Tools\backlogit.exe` v1.11.0 with exit codes checked,
followed by a read-back (`shipment get 030-S`, `dep list`, item frontmatter).
Body edits used `--section name=value` only.

| Plan unit | Backlog ID | Disposition |
|---|---|---|
| A (feature) | 033-F | **`blocked → queued`** by a direct `backlogit move 033-F --status queued` (exit 0). The `description` section is replaced with **HOLD LIFTED** (D-030-7). |
| A (shipment) | 030-S | The `description` section is replaced with HOLD LIFTED, the item order, the IVL-1 merge precondition (AC-A3d.6) and the safe-close rule. Manifest read back: `033-F, 033.001-T, 033.002-T, 033.003-T, 033.004-T, 033.005-T, 033.006-T, 033.008-T, 033.007-T`. |
| A-T1 | 033.001-T | An `acceptance-criteria` section is added with AC-A1.1..AC-A1.9 (rev 6: AC-A1.3 reworded, AC-A1.7 pinned wording, AC-A1.8 token equality, AC-A1.9 checklist). |
| A-T3a | 033.004-T | `description` (rev 6, HOLD LIFTED) and `acceptance-criteria` AC-A3a.1..3. |
| A-T3b | 033.005-T | Retitled "Freeze retiredarch pathspec selection declarations". Rev-6 `description` and AC-A3b.1..4. M, medium. |
| A-T3c | 033.006-T | Retitled "Freeze retiredarch prefix surface and add scopeDataOK". Rev-6 `description` and AC-A3c.1..4. S, medium. |
| A-T3d | **033.007-T** | **NEW.** "Add package closure to the scan-scope pin". S, medium, test-first. AC-A3d.1..6. |
| A-T4 | **033.008-T** | **NEW.** "Probe non-test cmd/ Go files in selection self-test". XS, low, test-first. AC-A4.1..3. |
| Dropped | 033.003-T | `description` and `acceptance-criteria` record D-000-3. AC-X3.2 timing: after 033.007-T. AC-X3.4: the precondition is discharged. |
| D-T0 | 034.001-T | `description` and `acceptance-criteria` record D-000-3. AC-D0.4: the precondition is discharged. Legacy criterion (3) is read as "revision 5 or later". |
| B (feature) | 035-F | **`blocked → queued`** by a direct `backlogit move 035-F --status queued` (exit 0), per D-032-4. 032-S is **claimable**: it is `queued`, its only dependency 037-S is archived, and 035-F and 035.001-T..035.003-T are all `queued`. |

**Lifecycle note.** `.backlogit/hooks.yaml` lists `blocked → active` as the
only transition out of `blocked`. The CLI nevertheless accepted a direct
`blocked → queued` move for both features. No `--force-gates` was used.

**Edges added** (`A → B` means A depends on B, type `blocks`):
* 033.007-T → 033.006-T
* 033.008-T → 033.002-T
* 033.007-T → 033.008-T

None was removed.

**Stash.** P-021 DEFERRED SCOPE EXPANSION entries `DC921AF6` (R-A1, with
`scanPath`/`engineForPath` added after round 4) and `D7BF9F74` (R-A2, with the
environment redirect variables added after round 4). Both are `task`,
`medium`, and require deliberation (P-021 C6).

**Sizing.** `--size` with `--size-source agent --size-ruleset-version
2h-rule-v1` succeeded. **`--complexity` is still unsupported**: "artifact type
task does not define a complexity field". Complexity is recorded as prose.

**Index.** `backlogit sync` indexed 397 artifacts, and `backlogit doctor`
reported "No issues found".

---

## Plan Review — Revision 7 (Unit E)

This section records the plan-review gate for the revision 7 Unit E re-plan
(D-049-1). Scope: plan rev 7, the spike findings
(`docs/decisions/2026-10-02-intercom-go-writepath-go-ast-spike.md`),
deliberation §4.9, and the 049-F / 049.001-T..049.007-T / 039-S backlog
bodies. The gate rule is the 049-F STAGE HOLD rule: the hold lifts only on
ADVISORY or better with no open P0/P1. There are at most 3 review-fix cycles
(circuit-breaker Review-Fix Cycle Definition).

dispatch_mode: multi-agent (three reviewers on three model families, each
reading the repository read-only from a brief under `logs/`)

### Cycle 1 — FAIL

| Reviewer | Verdict | Findings |
|---|---|---|
| gpt-5.6-sol | FAIL | PR7-1 (P1), PR7-2 (P1), PR7-3 (P2), PR7-4 (P2) |
| claude-opus-5.5 | ADVISORY | SB-1..SB-9 (P2/P3) |
| gemini-3.8-flash | ADVISORY | 049.007-T ACs; full qualifier set; dot-import format (P3) |

| Finding | Disposition |
|---|---|
| PR7-1 / SB-1 — item 6 claimed closed while same-family primitives stay uncovered | Applied. Item 6 is **narrowed** everywhere; the remainder is stash C0D28448 (P-021 C2). |
| PR7-2 / SB-6 — file counts understated | Applied. Each task states artifacts and physical files (D-T6 convention). |
| PR7-3 — AC-E7.3 sampled selectors | Applied. Exact-name scan of all 50 at `961b652`. |
| PR7-4 / SB-3 — E-T1 ACs differ from 049.001-T | Applied. The backlog mirrors AC-E1.1..1.4. |
| SB-2 / SB-9 — stash IDs, full-set H-5 commands | Applied (FE2F02FF, C0D28448; H-5 item-6 and item-8 commands). |
| SB-7 — dot-import fixture for new paths | Applied (AC-E7.1). |
| SB-8 — E-T6 / E-T7 revert ordering | Applied (H-6). |
| gemini P3s | Applied (049.007-T sections; 8 qualifiers; canonical prefix). |

<!-- plan-review-attempt: 1 -->

### Cycle 2 (re-entry 1 of 2) — FAIL

| Reviewer | Verdict | Findings |
|---|---|---|
| gpt-5.6-sol | FAIL | PR7-1 unresolved in `049.006-T` (P1); PR7C2-1 (P2) |
| claude-opus-5.5 | FAIL | SB2-1 (P1), SB2-2 (P1), SB2-3 (P2), SB2-4..SB2-6 (P3) |
| gemini-3.8-flash | ADVISORY | GM2-1 (P2), GM2-2 (P3) |

| Finding | Disposition |
|---|---|
| PR7-1 / SB2-1 / GM2-1 — `049.006-T` replan still said "item 6 remove"; its ACs lacked the narrowed parts of items 2/6 and item 8 | Applied. A cycle-2 amendment in `049.006-T` supersedes "item 6 remove", and its acceptance criteria mirror AC-E6.1/E6.2. |
| SB2-2 — E-T3 deletes `extractExtent` but omits `writepath_extent_test.go`, which calls it, so the package would not compile | Applied. E-T2 keeps `scanText`/`advanceCursor`/`extractExtent` unchanged; E-T3 has 3 files, size M, the extent-test port table and AC-E3.3. |
| PR7C2-1 — AC-E7.3 claimed a `_test.go`-inclusive scan, but the recorded command excludes `_test.go` | Applied. H-5 records the inclusive variant (re-run at `961b652`: the single comment hit); AC-E7.3 cites it. |
| SB2-3 — this section was referenced but missing | Applied (this section). |
| SB2-4 — header-count drift window between E-T7 and E-T6 | Applied. H-6 states AC-D4.2 is a manual review check; the window is accepted. |
| SB2-5 — spike doc file-count wording and column name | Applied in the spike doc. |
| SB2-6 — stale description paragraphs in 049.006-T / 049.007-T | Applied. Each replan section marks the description text as superseded. |
| GM2-2 — `unix.` vs `unix.*` | Applied (AC-E7.1 and 049.007-T). |

<!-- plan-review-attempt: 2 -->

### Cycle 3 (re-entry 2 of 2, FINAL) — FAIL

| Reviewer | Verdict | Findings |
|---|---|---|
| gpt-5.6-sol | FAIL | GPT3-1 (P1) |
| claude-opus-5.5 | FAIL | SB3-1 (P1), SB3-2..SB3-5 (P3) |
| gemini-3.8-flash | (empty response) | — replaced by grok-4.7 |
| grok-4.7 | FAIL | GK3-1 (P1), GK3-2 (P2), GK3-3 (P3), GK3-4 (P3) |

Stage verified each P1 against the code at `961b652`. All three are real:

* **SB3-1 / GPT3-1 (P1):** E-T3 neither deleted nor ported
  `lineReportsSelector` (`writepath.go:315`), which calls `extractExtent` and
  `occurrenceAllowed`. As planned, E-T3 either would not compile or would drop
  the AC-D2.3 same-line security test
  (`writepath_test.go:484–506`).
* **GK3-1 (P1):** the `writepath_extent_test.go` classification inputs
  `os.Remove(a, (b`, `os.Remove(`, `os.Remove(a, [b)` and `os.Remove(a}` are
  not valid Go. Under G-2 they are parse errors with exit 1 and no finding, so
  the port table's "same finding text" claim contradicted G-2.

| Finding | Disposition (applied after cycle 3; **NOT re-gated**) |
|---|---|
| SB3-1 / GPT3-1 | E-T3 deletion list adds `lineReportsSelector`; port table adds `TestScanSource_AllowedCallCannotHideWritingCall`; AC-E3.3 grep extended; cited range corrected to `:384–512`. |
| GK3-1 | Classification port split into parseable (same verdicts) and non-parseable (`TestScanSource_SyntaxErrorFailsClosed`, G-2 stricter fail-closed) rows; `extentUnbalanced`/`extentMismatched` not preserved. |
| GK3-2 | `findSelector` deleted; `nextOccurrence` retained; `TestFindSelector_LookaroundTable` re-pointed at `selectorOccurrences`. |
| SB3-2 / SB3-3 | Complete deletion list names `extentKind`, `bracketOpener`, `firstExtent`, `absoluteOccurrences`, and the `createFileCall`/`allowedCreateFileArgs` disposition. |
| SB3-4 | AC-E3.3 grep scoped to `tools/gatecheck/internal/writepath/`. |
| SB3-5 | Recorded; it is subsumed by the SB3-2 deletion list. |
| GK3-3 | E-T2 note corrected: the predicate, same-line and lookaround tests stay unchanged until E-T3. |
| GK3-4 | `049.005-T` carries a supersession sentence (item 2 is narrowed, not closed). |

<!-- plan-review-attempt: 3 -->

### Gate outcome — FAIL (cycle limit reached)

Each of the three review cycles found at least one new P1. The plan-review
re-entry limit (2 re-entries, 3 attempts) is now exhausted. Under the Stage
agent's Step 4 rule and P-021 C3, in-scope findings that remain unverified at
the cycle limit **halt and escalate to the operator**. They are not accepted as
residual risk.

The cycle-3 fixes above are applied, but **no reviewer has verified them**.
Consequences:

* `049-F` stays **blocked** (the STAGE HOLD is **not** lifted).
* `039-S` stays **on hold** and is not claimable. The Orchestrator reports
  `DARK_MODE_HALTED` for 039-S as unsafe without operator input.
* The P-013.6 escalation route is `ESCALATION_DEGRADED`, because engram is
  unavailable for the handoff. Stage falls back to the operator halt.
* The next step needs the operator: an explicit, operator-authorized re-gate
  (a fresh plan-review run over revision 7 with these amendments), or another
  disposition. No harvest record for revision 7 is written until a gate passes.

The spike result itself (G-1..G-7, exact verdict parity) is **not** in question.
The failures are all in the decomposition detail of E-T3, the test-port
contract.

### Staging PR #99 review amendments (applied after the gate; **NOT re-gated**)

Copilot's shadow review of staging PR #99 raised these. Stage classified each
as the same contract surface (P-021 C1), so it applied them. They do **not**
change the gate outcome: the hold stays in place, and the operator-authorized
re-gate covers them too.

| Finding | Disposition |
|---|---|
| Spike harness only logged divergences (iteration 1) | Parity, golden, read, walk and filebased decode errors are now `t.Errorf`/`t.Fatal`. Re-run at `961b652`: the same tallies, PASS. |
| 039-S hold text overstated its mechanism (iteration 1) | The `hold` section names the mechanical guard (Ship Step 0.5 `shipment-reconcile` pre-mode). |
| E-T5 binding misses `token.ASSIGN` (iteration 2) | Binding strategy and AC-E5.1 cover `x = pathsafe.NewRoot(...)` and `x, err = ...`. |
| AC-E7.3 misses aliased/dot `io/ioutil` imports (iteration 2) | AC-E7.3 adds an import-path scan; an aliased or dot import is a HALT. |
| H-5 pointed at memory for verbatim output (iteration 2) | Commands and output recorded in `docs/decisions/assets/2026-10-02-writepath-go-ast-spike/h5-rev7-remeasure.txt`. |
| E-T5 binding omits `var x, err = pathsafe.NewRoot(...)` (iteration 3) | The single-name `var x = ...` (which does not compile) is replaced by the multi-name `ValueSpec` form. The AC-E5.1 fixture adds a third, `var`-bound caller. A `var` shape scan at `961b652` finds none. |
| ALLGO counted the untracked harness (iteration 3) | The ALLGO figure is described as 155 `.go` files in the scratch worktree: the 154 tracked at `961b652` plus the untracked harness. |

### Round 4 (operator-authorised exception, D-049-8) — FAIL

<!-- plan-review-attempt: 4 -->

* **Authority.** At 2026-10-02T22:36-07:00 the operator explicitly approved
  **one** fourth plan-review round over revision 7 Unit E (**D-049-8**). It
  is an exception to the 3-cycle cap, with the same precedent as revision 6
  round 4 (D-030-6 / D-030-7). The approval covers **exactly one** round.
  **No round 5 is authorised**, whatever the outcome.
* **dispatch_mode:** multi-agent. Three reviewers on three model families
  read the repository read-only from a brief under `logs/`. The brief scoped
  them to:
  * rev 7 Unit E, including every post-cycle-3 disposition and the PR #99
    amendments;
  * the spike document and its H-5 evidence file;
  * the 049-F, 049.001-T..049.007-T and 039-S bodies.

  Each reviewer verified claims against `tools/gatecheck/internal/writepath`,
  `gomask` and `retiredarch` on `main`.
* **Reviewed HEAD:** every reviewer reported
  `e3250d37e295906b1733fcf5bd0a29ebe4bdf2d6`. `git diff 961b652 HEAD --
  tools/gatecheck` is empty, so the plan's `961b652` line citations still hold.

| Reviewer | Route | Verdict | P0 | P1 | P2 | P3 |
|---|---|---|---|---|---|---|
| gpt-5.6-sol | **anchor** (openai / high) | **FAIL** | 0 | 1 (GPT4-1) | 2 (GPT4-2, GPT4-3) | 0 |
| claude-opus-5.5 | anthropic | **FAIL** | 0 | 4 (OP4-1..OP4-4) | 2 (OP4-5, OP4-6) | 1 (OP4-7) |
| gemini-3.8-flash | google | PASS | 0 | 0 | 0 | 1 (GM4-1) |

gemini-3.8-flash returned a non-empty answer, so no grok-4.7 substitution was
needed. Its PASS asserted that "the 18 parseable inputs preserve findings".
Stage's verification of GPT4-1/OP4-2 below shows that claim is wrong, so its
PASS does not outweigh the verified P1s.

**P0/P1 verification (Stage, against code at `e3250d37`).** Every P1 is
**accepted**. None is rejected as a false positive.

| Finding | Stage verification | Disposition |
|---|---|---|
| **GPT4-1 = OP4-2 (P1)** — the "call after whitespace and newline" row is listed as parseable with an unchanged verdict | `writepath_extent_test.go:82` is `"os.Remove \t\n(a, b)"`. Go inserts a semicolon after an identifier at a newline, so `(a, b)` begins a new statement and is a syntax error ("expected ')', found ','"). Under G-2 the verdict cannot be "balanced, 2 segments". **Confirmed.** This is the GK3-1 defect class, which the cycle-3 fix missed. | **R4-1** (applied, NOT re-gated): the row moves to the non-parseable row and to `TestScanSource_SyntaxErrorFailsClosed`. |
| **OP4-1 (P1)** — the E-T2 boundary is lint-red | `.golangci.yml` enables `unparam`, and INV-4 makes `golangci-lint` blocking. After E-T2 re-points `writepath.go:371`, `writepath_test.go:148` and `writepath_oracle_test.go:242`, `git grep 'scanText('` leaves only `writepath_extent_test.go:109/:118/:185/:210`. All four pass `"x.go"`, so `relPath` always receives `"x.go"`. The reviewer reproduced the `unparam` report in a scratch module. Stage did not run the linter (Stage role boundary); acceptance rests on the config and the caller set. **Confirmed.** | **R4-3** (applied, NOT re-gated): E-T2 changes one literal at `:118` to `"y.go"`, test-only and verdict-preserving. It is a fifth file, recorded as a deviation. |
| **OP4-3 (P1)** — predicate rows `:467–468` cannot be parsed | `writepath_test.go:467` builds `syscall.CreateFile(, 0, …)` and `:468` builds `…, 0,,)`. Both are "expected operand" syntax errors. The plan exempted only `:443–444`, so "unchanged cases and verdicts" could not be met. Stage re-checked every other row (`:442–469`): each parses as a function-body statement. **Confirmed.** | **R4-2** (applied, NOT re-gated): `:443`, `:444`, `:467` and `:468` move to `SyntaxErrorFailsClosed`. The executor re-probes the rows, and the `:442` non-call row is asserted through `scanSource`. |
| **OP4-4 (P1)** — the AC-D1.3 port cannot detect a `\n`-only line table | The Unit D test (`writepath_extent_test.go:150–195`) pins `SplitLines` numbering with separators in **code** position. The plan moves those to fail-closed. It keeps comment and string positions, where gomask erases the boundary runes, so masked `SplitLines` lines equal `go/token` lines. The lone-CR `filebased` golden does not discriminate either: `ReadText` translates CR to `\n` (`translated_text` in the golden). **Confirmed.** Stage notes that the impact is the reported line number, not the verdict. The severity is not contested, because the gate fails on the other P1s regardless. | **R4-4** (applied, NOT re-gated): one discriminating case, a visible raw-string tag containing `\u2028` or `\f`, with a non-vacuity precondition that the `go/token` line differs. |

**P2/P3 dispositions** (all are on the same contract surface under P-021 C1;
no stash capture was needed).

| Finding | Disposition |
|---|---|
| GPT4-2 (P2) — the single-name `:=`/`=` `NewRoot` forms do not type-check, and the AC-E5.1 `err` is undeclared | **R4-8** applied to E-T5, AC-E5.1, the spike's G-6 note, and 049.005-T. |
| GPT4-3 (P2) — 049-F's description still says the `Resolve` residual is closed | Applied: 049-F gains a `rev7-scope` section saying item 2 is **narrowed**, not closed. |
| OP4-5 (P2) — the interim E-T2 offset is not mapped from a source byte offset to a masked byte offset | **R4-5** applied to E-T2 and 049.002-T. The PR #101 Copilot review then added the explicit conversion `token.Pos` → `token.File.Offset` → rune index → masked byte offset. |
| OP4-6 (P2) — the wrapper scaffolding is unspecified, and the `"syscall"` import is needed for the allowance | **R4-6** applied to E-T3 and 049.003-T. |
| OP4-7 (P3) — the argument count and the segment count differ | **R4-7** applied to the E-T3 port table. |
| GM4-1 (P3) — the 039-S manifest lists 049.006-T before 049.007-T | **Recorded, not applied.** `backlogit shipment add` only appends, and execution order is governed by the `049.006-T → 049.007-T` `blocks` edge (D-049-4), not by manifest position. |

**Gate outcome — FAIL; the 049-F STAGE HOLD is KEPT (D-049-8).**
* The gate rule is not met. Two of three reviewers returned FAIL, and Stage
  verified four distinct P1s, all still open at review time.
* The R4-1..R4-8 amendments above are applied as **un-gated amendments**.
  **No reviewer has verified them.** Round 5 is not authorised, so they
  cannot be re-gated in this session.
* `049-F` stays `blocked`. `039-S` stays on hold and **must not be claimed**.
  The Ship Step 0.5 `shipment-reconcile` pre-mode guard still classifies the
  blocked 049-F as a status mismatch.
* **Escalation (P-013.6).** The resolved escalation route is gpt-5.6-sol /
  openai / high. The engram handoff surface is not available in this session
  (no engram MCP tool and no running daemon), so this is
  `ESCALATION_DEGRADED`, and Stage falls back to the operator halt.
* **Pattern.** Cycle 1's P1s concerned residual-item scope and file counts,
  and cycle 2 carried one of those over (PR7-1). From cycle 2 onward, every
  round has also found new P1s in the E-T2/E-T3 test-port and boundary detail:
  * cycle 2, SB2-2: the omitted `writepath_extent_test.go` boundary;
  * cycles 3 and 4: row-level parseability, plus the `unparam` boundary in
    round 4.

  No round found a P1 in the spike result (D-049-2) or the engine design.
  Ship-time compile, lint and test evidence would surface this defect class
  mechanically, but plan review cannot.

  **The hold contract still applies.** Under the current lift rule (D-049-7,
  carried forward by D-049-8), only a plan-review gate of ADVISORY or better
  with no P0/P1 moves 049-F to `queued`. Until then, Ship's pre-claim
  reconciliation stops any 039-S claim. Every alternative path therefore needs
  a **new, recorded operator decision** before it can be selected, and that
  decision must explicitly amend or replace the lift rule. Two examples of
  what such a decision could authorise:
  * a Ship-executed E-T2/E-T3, gated by an explicit rule that port-table rows
    are verified by a `go/parser` probe at execution;
  * merging E-T2 and E-T3 under a recorded 2-hour deviation.

  Alternatively, the operator could authorise a further plan-review round.
  Choosing among these options, and recording the transition, are operator
  decisions. Stage does not make them.

### Round 5 (operator-authorised exception, D-049-9) — FAIL

<!-- plan-review-attempt: 5 -->

dispatch_mode: multi-agent
decision: FAIL

* **Authority.** At 2026-10-03T00:13-07:00 the operator explicitly approved
  **one** fifth plan-review round over revision 7 Unit E, as amended by
  R4-1..R4-8 (**D-049-9**). It is the second operator-authorised exception to
  the 3-cycle cap for the Unit E rev 7 gate, after D-049-8 (the third overall,
  after D-030-6 for Unit A). It supersedes round 4's "no round 5" for
  this single round only. **It does not authorise a round 6**, whatever the
  outcome.
* **dispatch_mode:** multi-agent. Three reviewers on three model families
  read the repository read-only from a brief under `logs/`. The brief scoped
  them to:
  * rev 7 Unit E as amended, including R4-1..R4-8 and every prior
    disposition;
  * the 049-F `rev7-scope` section;
  * the spike document;
  * the 049-F, 049.001-T..049.007-T and 039-S bodies.

  The brief required each reviewer to run `go/parser` over every
  classification input, to check lint against `.golangci.yml` (including
  `unparam`), and to check every line citation and discrimination claim.
* **Reviewed HEAD:** every reviewer reported
  `121e972dbabf20d9b6bbf27b9cff6694cf04ebc6`. `git diff 961b652 121e972 --
  tools/gatecheck` is empty, so the plan's `961b652` citations still hold.

| Reviewer | Route | Verdict | P0 | P1 | P2 | P3 |
|---|---|---|---|---|---|---|
| gpt-5.6-sol | **anchor** (openai / high) | **FAIL** | 0 | 2 (GPT5-1, GPT5-2) | 0 | 0 |
| claude-opus-5.5 | anthropic | ADVISORY | 0 | 0 | 1 (OP5-1) | 3 (OP5-2..OP5-4) |
| gemini-3.8-flash | google | PASS | 0 | 0 | 0 | 0 |

gemini-3.8-flash returned a non-empty answer, so no grok-4.7 substitution was
needed. Its unmasked-row parse table agrees with Stage's probe. However, its
amendment audit attaches the R4-1..R4-8 labels to the wrong amendments, and it
did not examine the masked row values. Its PASS therefore does not outweigh
the verified P1 below.

**Stage mechanical verification (at `121e972`).**
* **Parse probe.** Stage ran a temporary in-package test file under
  `tools/gatecheck/internal/writepath/`. It was deleted after the run and never
  committed. The test calls `parser.ParseFile` (`ParseComments|SkipObjectResolution`)
  on each row, wrapped as `package p` + import + `func f() { … }`. The R4-1
  `var` form was also checked.
  * **Extent rows `writepath_extent_test.go:78–89`.** These parse: `os.Remove`,
    `os.Remove;`, the call, nested depth-1, trailing comma and the empty call.
    These are errors: code-position `\f`, `os.Remove \t\n(a, b)` ("expected
    ')', found ','"), `(a, (b`, `os.Remove(`, `(a, [b)` and `(a}`. This
    matches GK3-1 and R4-1.
  * **Scan-level rows.** `f := os.Remove` and `os.Remove(a)` parse. `(a, (b`
    and `(a, [b)` are errors.
  * **Predicate rows `writepath_test.go:442–469`, unmasked source.** Only
    `:443`, `:444`, `:467` and `:468` are errors. This matches R4-2.
  * **Predicate rows, the masked `tc.masked` values as they exist.** `:464`,
    `:465` and `:466` are **also** errors ("missing ','" and "expected
    operand"). gomask blanks the string/rune literal and the raw-string
    backticks.
  * **R4-4 discrimination.** A visible raw struct tag containing `\u2028` or
    `\f`, followed by `os.Remove(`, parses. The tag stays visible after
    masking. `go/token` gives line 10 and masked `SplitLines` gives line 11,
    so the precondition is non-vacuous.
* **Lint experiment.** Stage made a temporary local edit, then reverted it
  with `git checkout -- tools/gatecheck`. golangci-lint 2.13.2 ran with the
  repo `.golangci.yml`.
  * **Baseline** at `121e972`: `0 issues`.
  * **E-T2 caller set.** Stage re-pointed `writepath.go:371`,
    `writepath_test.go:148` and `writepath_oracle_test.go:242` to a stub that
    does not call `scanText`. Lint reported `writepath.go:341:15: scanText -
    relPath always receives "x.go" (unparam)`, 1 issue.
  * **R4-3 applied** (`:118` and both expected strings changed to `"y.go"`):
    `0 issues`. The narrowed `go test` run passes.

**P0/P1 dispositions.**

| Finding | Stage verification | Disposition |
|---|---|---|
| **GPT5-1 (P1)** — R4-2 says only `:443`, `:444`, `:467` and `:468` are non-parseable and "every other row parses", but it does not say the rows are probed and ported from unmasked source. The existing row values are masked. | **Confirmed** by the parse probe. Probing the masked values gives three more parse errors, at `:464`–`:466`. Read literally, R4-2 would then move the string, rune and tag-shaped raw-string rejections into `TestScanSource_SyntaxErrorFailsClosed`. That would remove the only unit pins for the G-4 STRING/CHAR literal rejection (INV-2). No fixture covers an interpreted-string or rune `CreateFile` argument. OP5-1 independently reports the same defect at P2. **Accepted as P1.** | **R5-1** (applied, NOT re-gated): ported rows are built from unmasked source, and `createFileCall` drops `MaskGoNonCode`. `tc.masked` is never probed. `:464`–`:466` stay parseable rejections. The probe must reproduce exactly `{:443, :444, :467, :468}`, or it is an H-2 HALT. AC-E3.2 states the G-4 bare-operand set positively, and AC-E3.3 and 049.003-T mirror the change. |
| **GPT5-2 (P1)** — R4-3 is unnecessary because `unparam` does not report a constant parameter whose callers are all in `_test.go` | **Rejected as a false positive.** The lint experiment above uses the real package and the exact E-T2 caller set, and `unparam` **does** report `relPath always receives "x.go"`. R4-3 clears it. The reviewer's `0 issues` came from a synthetic scratch package, not the real one. claude-opus-5.5 ran a full-package simulation and reproduced the finding before R4-3, with `0 issues` after it, which matches Stage. | R4-3 is **retained**. GPT5-2's secondary observation is real: the Constitution Check row VI said "E-T2 4", but row X says five. It is accepted as P3 → **R5-4** (row VI now says "E-T2 5"). |

**P2/P3 dispositions** (all are on the same contract surface under P-021 C1;
no stash capture was needed).

| Finding | Disposition |
|---|---|
| OP5-1 (P2) — the masked/unmasked ambiguity in the predicate port, and the missing positive bare-operand set | Merged into **R5-1**. |
| OP5-2 (P3) — the AC-D2.1 test (`writepath_test.go:402–425`) has no port-table row, and its selector-keying sub-case has no AST form | **R5-2** applied: the live and single-line sub-cases are mapped to their carriers, and the `os.Remove` 7-argument keying case becomes a `scanSource` finding assertion. |
| OP5-3 (P3) — D-049-9 is not recorded in a tracked file | Applied: this section, plan changelog item 12, and the deliberation §4.9 / §5 D-049-9 row. |
| OP5-4 (P3) — import housekeeping at the E-T2/E-T3 boundaries | **R5-3** applied: imports left unused (for example `unicode/utf8`) are removed in E-T3. `scanSource` lives in `writepath.go`, so `gomask` stays used. |

**Gate outcome — FAIL; the 049-F STAGE HOLD is KEPT (D-049-9).**
* The gate rule is not met. The anchor returned FAIL, and Stage verified one
  P1 (GPT5-1) that was open at review time.
* R5-1..R5-4 are applied as **un-gated amendments**. **No reviewer has
  verified them.** No round 6 is authorised, so they cannot be re-gated in
  this session.
* `049-F` stays `blocked`. `039-S` stays on hold and **must not be claimed**.
  The Ship Step 0.5 `shipment-reconcile` pre-mode guard still classifies the
  blocked 049-F as a status mismatch.
* **Escalation (P-013.6).** The resolved escalation route is gpt-5.6-sol /
  openai / high. The engram handoff surface is not available in this session,
  so this is `ESCALATION_DEGRADED`, and Stage falls back to the operator halt.
* **Pattern.** Round 5 again found its P1 in the E-T3 test-port row detail,
  the same defect class as cycles 3–4: whether a port row parses. Round 5 is
  the first round with only one verified P1, and no reviewer found a P1 in the
  spike result, the engine design, or E-T4..E-T7. The round-4 note about
  alternative paths still applies unchanged. Every path other than a further
  plan-review round needs a **new, recorded operator decision** that amends
  the D-049-7 lift rule. A further round would also need an explicit
  operator authorisation, because D-049-9 does not cover round 6.

**Pre-PR review of this record** (gpt-5.6-sol ADVISORY, claude-opus-5.5
ADVISORY; both reviewed the same pinned staged diff at
`4b0491d51d8bdb63253fdfc4b5fecae183c4cd10`). Neither reviewer found a P0 or
P1. Both independently reproduced the R5-1 parse claims and confirmed the line
citations against HEAD.

| Finding | Disposition |
|---|---|
| **PRE5-1 (P2, gpt-5.6-sol)** — AC-E3.2 and 049.003-T claimed STRING/CHAR literals are rejected "at every position", pinned by `:464`–`:466`. Those rows only place the literal at `Args[2]`, so a predicate that validates only that position could pass the specified tests. | **Accepted.** The claim is narrowed in R5-1, AC-E3.2 and 049.003-T so that it no longer overstates what `:464`–`:466` pin. Adding table-driven literal cases at the other generic positions (`Args[0]`, `Args[3]`, `Args[6]`) would be a new test requirement in E-T3. It is **left open** for the next gate rather than applied here as another un-gated amendment. *(Round 6: re-raised as GPT6-1 (P1) and the reason for the round-6 FAIL. Closed in the text by R6-1, un-gated; see Round 6 below.)* |
| **PRE5-2 (P3, claude-opus-5.5)** — D-049-9 was called "the second exception to the 3-cycle cap", but D-030-6 (Unit A) was an earlier one | **Applied.** The wording now says "the second for the Unit E rev 7 gate, after D-049-8 (the third overall, after D-030-6)" here, in the deliberation D-049-9 row, and in the session memory. |

### Round 6 (operator-authorised exception, D-049-10) — FAIL

<!-- plan-review-attempt: 6 -->

dispatch_mode: multi-agent
decision: FAIL

* **Authority (recorded verbatim).** At 2026-10-03T21:10-07:00 the operator
  wrote: "I approve a sixth round of review; however, if the findings again
  identify are all about how test inputs are classified and would be caught
  anyway through test-first build, then lift the hold and let the code review
  be the gate." This is **D-049-10**. It authorises **exactly one** round,
  round 6, over rev 7 Unit E as amended by R4-1..R4-8 and R5-1..R5-4, plus a
  conditional hold-lift rule that applies to round 6 only. **It does not
  authorise a round 7.**
* **Round-6 gate rule (D-049-10), applied in order:**
  1. **PASS.** Every reviewer returns ADVISORY or better and no verified
     P0/P1 is open. The hold lifts under the D-049-7 rule.
  2. **CONDITIONAL LIFT.** Verified P0/P1s exist, but every one is (a)
     classification-only and (b) mechanically caught by a test-first build.
     (a) means it concerns only how test inputs or fixture rows are
     classified. (b) means a mis-classification makes a test fail to compile
     or fail at run time, so it cannot slip into a passing build.
  3. **FAIL.** Any verified P0/P1 falls outside (a)+(b). That includes a
     test that could pass while asserting the wrong thing, an AC that a
     passing test would not expose, or a false-negative risk in write-path
     detection. Any doubt counts as not qualifying.
* **dispatch_mode:** multi-agent. Three reviewers on three model families
  read the repository read-only from a brief under `logs/`. The brief scoped
  them to:
  * rev 7 Unit E as amended, including R4-*, R5-* and every prior
    disposition, with the open PRE5-1;
  * the spike document and its evidence;
  * the 049-F, 049.001-T..049.007-T and 039-S bodies.

  The brief required each reviewer to:
  * parse-probe every classification input with `go/parser` from unmasked
    source;
  * check the R4-4 line discrimination against the real `gomask`;
  * check lint against `.golangci.yml` at the E-T2 and E-T3 boundaries;
  * tag every P0/P1 `classification-only` and `caught-by-test-first`.
* **Reviewed HEAD:** every reviewer reported
  `4c369448de039049e70158c9714bbeb0a2f9bb76`. `git diff 961b652 4c36944 --
  tools/gatecheck` is empty (Stage re-ran it), so the `961b652` citations
  still hold.

| Reviewer | Route | Verdict | P0 | P1 | P2 | P3 |
|---|---|---|---|---|---|---|
| gpt-5.6-sol | **anchor** (openai / high) | **FAIL** | 0 | 1 (GPT6-1) | 0 | 0 |
| claude-opus-5.5 | anthropic | ADVISORY | 0 | 0 | 1 (OP6-1) | 5 (OP6-2..OP6-6) |
| gemini-3.8-flash | google | ADVISORY | 0 | 0 | 1 (GM6-1) | 0 |

* gemini-3.8-flash returned a non-empty answer, so no grok-4.7 substitution
  was needed. Its parse verdicts agree with Stage's, but some of its row labels
  do not match the file. It puts the form-feed row at extent `:81` (it is
  `:80`). It also labels predicate rows `:445`–`:463` "wrong package", "too
  few args" and so on, whereas the file has six arguments, eight arguments,
  the access variants and the rest. Its table is therefore weak evidence.
* Every reviewer independently confirmed the rest of the R5-1 and earlier
  claims:
  * the unmasked parse-error set is exactly `{:443, :444, :467, :468}`;
  * masked `:464`–`:466` do not parse;
  * the extent and scan-level classifications are correct;
  * the R4-4 `go/token` and `SplitLines` lines differ (10 vs 11, or 7 vs 8,
    depending on the wrapper);
  * the R4-3 `unparam` report appears before the edit and is gone after it.

  The anchor and claude-opus-5.5 ran the lint check on scratch copies.

**Stage mechanical verification of the P1 (at `4c36944`).** Stage wrote a
throwaway `go/parser` program under `logs/round6-stage/` and deleted it after
the session; it was never tracked. The program:
* rebuilds every predicate row `writepath_test.go:442–469` from **unmasked**
  source inside the R4-6 wrapper;
* applies two allowance predicates to each parsed `*ast.CallExpr`:
  * the full G-4 predicate: a bare-operand check at every generic position
    `Args[0]`, `Args[2]`, `Args[3]` and `Args[6]`, plus the exact
    access/disposition/flags checks, non-variadic, and 7 arguments;
  * a **fail-open variant** that runs the bare-operand check only at
    `Args[0]` and `Args[2]`.

Results:
* The parse errors are exactly `{:443, :444, :467, :468}`.
* Over every specified row plus the allowed control, the two predicates give
  **0 divergences**. A predicate that never examines `Args[3]` or `Args[6]`
  therefore passes the whole planned table.
* The candidate rows (STRING, CHAR or call-valued at `Args[3]` and `Args[6]`)
  all parse, and each one separates the two predicates.
* In the first run, a candidate call row failed to parse because of the
  probe's own doubled `)`. That illustrates why every row must be probed.

**P0/P1 disposition and the D-049-10 conditional-rule evaluation.**

| Finding | Stage verification | (a) classification-only | (b) caught by test-first | Disposition |
|---|---|---|---|---|
| **GPT6-1 (P1, anchor) = OP6-1 (P2) = GM6-1 (P2)**: PRE5-1 is still open. AC-E3.2 requires every generic argument to be a G-4 bare operand, but the STRING/CHAR rows only exercise `Args[2]`. | **Confirmed** by the probe above. `Args[3]` and `Args[6]` have **no** non-bare-operand rejection row of any kind: they are always `nil` and `0`. Two reviewers rated it P2. Stage keeps the anchor's P1 for two reasons. First, the defect is a path to widening the D-2′ allowance (INV-2, H-2 E) that no specified test would expose. Second, the anchor's verdict is FAIL. Even at P2, rule 1 would not be met, because rule 1 requires every reviewer to return ADVISORY or better. | **No.** Doubt counts as not qualifying. The fix adds rows that expect a finding, and the reviewer tagged it yes. But the defect is **missing coverage**: no `Args[3]`/`Args[6]` input exists to be classified. It is not the classification of an existing input. (Corrected from "arguably yes" after Copilot review on PR #103. The outcome is unchanged.) | **No.** The defect is **missing** rows. A position-specific, fail-open predicate compiles and passes every planned test (0 divergences), so nothing in the red/green phase fails. The reviewer tagged it no. | **Does not qualify**, because both (a) and (b) fail. It is also on the operator's FAIL list: an AC (AC-E3.2) whose violation a passing test would not expose, and a false-negative risk in write-path detection (an allowance widening suppresses a `syscall.CreateFile` finding). **Rule 3 → FAIL.** The same-surface fix is applied un-gated as **R6-1**. |

**P2/P3 dispositions.** All are on the same contract surface (P-021 C1), so
no stash capture was needed. All are applied as **un-gated amendments**.

| Finding | Disposition |
|---|---|
| OP6-1 (P2), GM6-1 (P2) | Same defect as GPT6-1, so merged into **R6-1**. E-T3 adds nine parseable rejection rows: STRING and CHAR at `Args[0]`, `Args[3]` and `Args[6]`, and call-valued at `Args[2]`, `Args[3]` and `Args[6]`. AC-E3.2 and AC-E3.3 mirror them, and 049.003-T carries them in its `round6` section and ACs. PRE5-1 is closed in the text. |
| OP6-2 (P3): AC-E3.3's grep matches the `t.Fatalf` strings `"scanText(%s) …"` (`writepath_test.go:150`, `:154`) and `"findSelector(%q, %q) …"` (`:180`) | **R6-2.** E-T2 renames `:150`/`:154` when it re-points `:148`, and E-T3 renames `:180` when it re-points the lookaround table. Added as AC-E2.7 and AC-E3.4. |
| OP6-3 (P3): the new predicate is unnamed, so reusing `occurrenceAllowed` would trip the grep | **R6-3.** The new predicate must not reuse a deletion-list name such as `occurrenceAllowed`. A name that is not on the list, such as `callAllowed`, is acceptable. Added as AC-E3.4. |
| OP6-4 (P3): the `advanceCursor` call at `writepath_extent_test.go:236` is missing from the call-site list | **R6-4.** `:236` is added. Stage confirmed it with `git grep`. |
| OP6-5 (P3): the `gomask` import in `writepath_test.go` becomes unused after the R5-1 port, which is a compile error | **R6-5.** R5-3's "stays used" is scoped to `writepath.go`. E-T3 removes the `writepath_test.go:11` import, and `writepath_extent_test.go` keeps its own. Stage confirmed that every `gomask` use in `writepath_test.go` (`:148`, `:385`, `:411`, `:442`–`:444`, `:464`, `:499`, `:505`) is re-pointed or retired. Added as AC-E3.4. |
| OP6-6 (P3): the `retiredarch` pin positively requires `scanText` (`writepath_mask_test.go:51`, `:300`), but E-T3 deletes `scanText` and cannot edit `retiredarch` | **R6-6.** The pin that E-T2 re-anchors anchors on `scanSource` masking its source parameter and consuming the result through `SplitLines`, with no positive `scanText` dependency. Added as AC-E2.7, and in 049.002-T. |

**Gate outcome — FAIL; the 049-F STAGE HOLD is KEPT (D-049-10).**
* The anchor returned FAIL, and Stage verified GPT6-1 as an open P1. GPT6-1
  fails both conditions of the operator's conditional rule: (a), because it is
  a coverage gap and doubt counts as not qualifying; and (b), because a weak
  predicate passes. So neither PASS nor CONDITIONAL LIFT applies, and the
  result is **FAIL**.
* R6-1..R6-6 are applied as **un-gated amendments**. **No plan-review round
  has re-gated them.** The pre-PR record reviews below checked their text and
  citations, but they are not a gate. D-049-10 covers round 6 only, so **no
  round 7 is authorised**.
* `049-F` stays `blocked` (it was not moved, and `--force-gates` was not
  used). `039-S` stays on hold and **must not be claimed**. The Ship Step 0.5
  `shipment-reconcile` pre-mode guard still classifies the blocked 049-F as a
  status mismatch.
* **Escalation (P-013.6).** The resolved escalation route is gpt-5.6-sol /
  openai / high. The engram handoff surface is not available in this session,
  so this is `ESCALATION_DEGRADED`, and Stage falls back to the operator halt.
* **Pattern.** This is the second round in a row with exactly one verified P1,
  and every round-6 finding is in the E-T2/E-T3 test-port detail. For the
  first time in rev 7, every classification claim (parseability, masked vs
  unmasked, line discrimination, the lint boundary) was independently
  confirmed by all three reviewers. The remaining P1 is a **coverage** gap
  (missing rows), not a classification error. That is exactly the class the
  operator's conditional rule excludes. Further progress needs a **new,
  recorded operator decision**, for example:
  * accept R6-1 as written and lift the hold;
  * authorise a round 7;
  * amend the D-049-7 lift rule in some other way.

  Stage does not choose.

**Pre-PR review of this record** (staged diff pinned to `logs/round6-staged.patch`):

| Reviewer | Verdict | Findings | Disposition |
|---|---|---|---|
| gpt-5.6-sol | PASS | none | — |
| claude-opus-5.5 | ADVISORY | PRE6-1 (P3): the R6-3 example `callAllowed` read as a deletion-list name in 049.003-T and in the OP6-3 row | **Applied.** Both now say `occurrenceAllowed` is the forbidden example and `callAllowed` is an acceptable name. The reviewer confirmed every code citation, the R6-1 coverage claim, the section markers and the record integrity. |

**Copilot review on PR #103 and re-review.**
* Copilot cycle 1 flagged four threads on the same point: the record called
  GPT6-1 "classification-only". That was corrected to (a) **no** in
  `d6cc0ad`. The outcome is unchanged.
* Copilot cycle 2 asked for a local re-review of the corrected HEAD.

| Reviewer (at `d6cc0ad`) | Verdict | Findings | Disposition |
|---|---|---|---|
| gpt-5.6-sol | PASS | none | — |
| claude-opus-5.5 | ADVISORY | PRE6-2 (P3): changelog item 13, the gate-outcome bullet and the 039-S hold still named only (b) | **Applied.** All three now record that both (a) and (b) fail. |

* Both models then reviewed the PRE6-2 delta at `c4d8ae6` and returned PASS.
* Copilot cycle 3, the last allowed, flagged "No reviewer has verified
  them". That was reworded to "No plan-review round has re-gated them".
  Pre-PR record reviews are not a gate.

### Operator hold-lift decision (D-049-11)

decision: OPERATOR-LIFT

* **Authority (recorded verbatim).** At 2026-10-03T22:25-07:00 the operator
  wrote: "Accept R6-1 as written and lift the hold as recommended, then stop
  for manual new session initiation." This is **D-049-11**.
* **The recommendation it accepts.** The Orchestrator recommended to the
  operator:
  * accept R6-1 as written and lift the hold. R6-1 adds nine parseable
    `CreateFile` rejection rows that vary `Args[3]` and `Args[6]` (among
    other positions), which closes GPT6-1 and PRE5-1;
  * R6-1 is a narrow test-row addition, and pre-PR code review plus the TDD
    harness gate its correctness at implementation time.
* **Override semantics.**
  * D-049-11 is an operator **override of the D-049-7 hold-lift rule for
    revision 7**. Under that rule only a plan-review gate of ADVISORY or
    better with no P0/P1 lifts the hold. The override lifts it **without
    another plan-review round**.
  * **No review round ran.** The plan-review verdict of record for revision
    7 stays **round 6 FAIL** (D-049-10). This section does not change it, and
    no `plan-review-attempt` marker is added.
  * R6-1..R6-6 are accepted **as written**. So are R4-1..R4-8 and
    R5-1..R5-4, which rounds 5 and 6 reviewed without any gate passing.
    All of them stay **un-gated by plan-review**.
  * D-049-11 authorises **no change to plan content** beyond recording the
    decision and the lift. Unit E's execution content is the revision 7 text
    as it stood at `bba0534`.
  * It does not authorise Stage to claim or start 039-S. Ship starts 039-S
    in a new session that the operator opens manually.
  * **Effect on AC-E1.2 (049.001-T).** AC-E1.2 requires that "the re-plan
    was re-gated through plan-review (ADVISORY or better, no P0/P1)". D-049-11
    **discharges that clause** by operator override, so AC-E1.4 does not
    HALT on it. Ship verifies instead that D-049-11 is recorded here and in
    the deliberation §5 table. Every other clause of AC-E1.1..AC-E1.3 still
    applies. 049.001-T carries the same note in its acceptance criteria.
  * **Hold wording elsewhere.** Earlier "stays blocked", STAGE HOLD,
    PROVISIONAL and do-not-claim wording in 049-F, 039-S and the 049.00x-T
    bodies is historical. The "NOT re-gated" wording stays true, and it is
    the reason the obligations below exist.
* **Implementation-time verification obligations.** Each un-gated amendment
  below becomes an obligation that Ship's TDD harness and the pre-PR
  adversarial code review **must** close against the implemented diff. An
  obligation found unmet is an in-scope finding (P-021 C1), fixed before
  merge rather than deferred. The obligations are listed in the
  `implementation-verification` section of each owning task. Each task also
  gains a verification acceptance criterion (AC-E2.V, AC-E3.V and AC-E5.V).

  | Amendment | Source finding | Owning task | Obligation |
  |---|---|---|---|
  | R4-1 | GPT4-1 / OP4-2 | 049.003-T | `writepath_extent_test.go:82` is a `TestScanSource_SyntaxErrorFailsClosed` row |
  | R4-2 (as superseded by R5-1) | OP4-3 | 049.003-T | Exactly `{:443, :444, :467, :468}` fail to parse; `:442` is asserted through `scanSource` |
  | R4-3 | OP4-1 | 049.002-T | The one `"x.go"` → `"y.go"` literal at `writepath_extent_test.go:118`; `golangci-lint` reports 0 issues at the E-T2 boundary |
  | R4-4 | OP4-4 | 049.003-T | The discriminating raw-string tag case, with its non-vacuity precondition |
  | R4-5 | OP4-5 | 049.002-T | The interim offset conversion `token.Pos` → `tf.Offset` → rune index → masked byte offset, failing closed |
  | R4-6 | OP4-6 | 049.003-T | Whole-file wrappers that import every referenced package |
  | R4-7 | OP4-7 | 049.003-T | `len(CallExpr.Args)` replaces the segment count |
  | R4-8 | GPT4-2 | 049.005-T | Only compiling two-name `NewRoot` binding forms in the fixture and AC-E5.1 callers |
  | R5-1 | GPT5-1 / OP5-1 | 049.003-T | Rows are built from unmasked source; `:464`–`:466` stay parseable rejections |
  | R5-2 | OP5-2 | 049.003-T | The AC-D2.1 carriers, including the 7-argument `os.Remove` keying assertion |
  | R5-3 | OP5-4 | 049.003-T | Imports left unused by the deletions are removed |
  | R5-4 | GPT5-2 (secondary) | 049.002-T | The E-T2 five-file deviation (Constitution Check rows VI and X) |
  | R6-1 | GPT6-1 / OP6-1 / GM6-1 | 049.003-T | The nine positional rejection rows. **Mutation check:** a fail-open predicate that ignores `Args[3]`/`Args[6]` must fail at least one test |
  | R6-2 | OP6-2 | 049.002-T (`:150`/`:154`), 049.003-T (`:180`) | The renamed `t.Fatalf` prefixes |
  | R6-3 | OP6-3 | 049.003-T | The new predicate does not reuse a deletion-list name |
  | R6-4 | OP6-4 | 049.003-T | The `advanceCursor` call at `writepath_extent_test.go:236` is removed |
  | R6-5 | OP6-5 | 049.003-T | The `writepath_test.go` `gomask` import is removed |
  | R6-6 | OP6-6 | 049.002-T | The re-anchored `retiredarch` pin has no positive `scanText` dependency |

* **Lift.**
  * `049-F` moves `blocked → queued`.
  * The `hold` sections of 049-F and 039-S say **HOLD LIFTED (D-049-11)**.
  * 039-S is **eligible** and is the next shipment. Ship Step 0.5
    `shipment-reconcile` no longer finds a status mismatch.
  * See the Harvest Record — Revision 7 below.
* **Escalation:** not triggered. The operator made the disposition directly.

---

## Harvest Record — Revision 7 (2026-10-03)

The revision 7 harvest (049.007-T, the edges and the rev 7 bodies) was
created in the rev 7 re-plan sessions. Under the hold, no harvest record was
written until a gate passed (see "Gate outcome — FAIL (cycle limit
reached)"). This record is written at the D-049-11 lift. Every mutation used
`C:\Tools\backlogit.exe` v1.11.0 with exit codes checked, followed by a
read-back. Body edits used `--section name=value` only.

| Plan unit | Backlog ID | Disposition |
|---|---|---|
| E (feature) | 049-F | **`blocked → queued`** by a direct `backlogit move 049-F --status queued` (exit 0). No `--force-gates` was used, and the move was not gate-rejected. The `hold` section is replaced with **HOLD LIFTED** (D-049-11). It states that every earlier hold, PROVISIONAL, do-not-claim, stays-blocked and re-gated-before-execution text in the item, and in any section of its tasks (description, acceptance-criteria, replan and round sections), is historical as to hold status. It also records the AC-E1.2 discharge. |
| E (shipment) | 039-S | `queued`. The `hold` section is replaced with HOLD LIFTED. It records eligibility, the new-session start by Ship, the implementation-time obligations and the execution order. Manifest read back: `049-F, 049.001-T, 049.002-T, 049.003-T, 049.004-T, 049.005-T, 049.006-T, 049.007-T`, which covers every rev 7 task. It depends on 031-S, which has shipped. |
| E-T1 | 049.001-T | Status unchanged (`queued`, Ship's read-only verification half, D-049-6). The acceptance criteria gain a D-049-11 note: AC-E1.2's re-gate clause is discharged by the operator override, and Ship verifies that D-049-11 is recorded instead (pre-PR review GPTH-001/OPH-1). |
| E-T2 | 049.002-T | An `implementation-verification` section is added (R4-3, R4-5, R5-4, R6-2 E-T2 half, R6-6), and the acceptance criteria gain **AC-E2.V**. AC-E2.1..AC-E2.7 already match the plan. |
| E-T3 | 049.003-T | An `implementation-verification` section is added (R4-1, R4-2/R5-1, R4-4, R4-6, R4-7, R5-1..R5-3, R6-1..R6-5 E-T3 parts), with the R6-1 mutation check. The acceptance criteria gain **AC-E3.V**. AC-E3.1..AC-E3.4 already match the plan, including the nine R6-1 rows. |
| E-T4 | 049.004-T | Unchanged. AC-E4.1/AC-E4.2 and the `replan` section match the plan. No R4-*/R5-*/R6-* amendment binds it. |
| E-T5 | 049.005-T | An `implementation-verification` section is added (R4-8), and the acceptance criteria gain **AC-E5.V**. AC-E5.1..AC-E5.3 already match the plan. |
| E-T6 | 049.006-T | Unchanged. AC-E6.1/AC-E6.2 match the plan. |
| E-T7 | 049.007-T | Unchanged. AC-E7.1..AC-E7.3 match the plan. |

**Edges** (unchanged; `A → B` means A depends on B, type `blocks`). The
execution order is E-T2, E-T3, E-T4, E-T5, E-T7, E-T6:
* 049.002-T → 049.001-T
* 049.003-T → 049.002-T
* 049.004-T → 049.002-T and 049.003-T
* 049.005-T → 049.004-T
* 049.007-T → 049.005-T
* 049.006-T → 049.003-T, 049.005-T and 049.007-T

**Stash.** No entries were created. Every change is on the decision-record
and backlog surface of 049-F (P-021 C1).

*Generated by Copilot*

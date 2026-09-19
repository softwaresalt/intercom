---
title: "Implementation Plan — Checker correctness, test coverage, docs and local hygiene (S-7…S-10)"
date: 2026-09-18
status: plan-review attempt 1 FAIL (1× P0) → revised; attempt 2 FAIL (0 P0, 5× P1) → reviewer disposition ADOPTED in revision 4; no attempt 3 submitted
agent: Stage
source_document: docs/decisions/2026-09-18-intercom-go-ship-contract-and-gate-reliability-deliberation.md
governs: two shippable units (S-7 checker correctness, S-8 topology-gate docs + diagnostics retention) and one HELD unit (staging-gate test coverage)
bound_snapshot: 14d44e3c1f28b321e8db24ab8cafcba346996133
stage_branch: chore/stage-ship-pipeline-contract-repair
requires_plan_hardening: yes
revision: 4
---

<!-- plan-review-attempt: 2 -->
<!-- attempt 2 verdict: FAIL (5x P1, 0 P0). Reviewer's recommended disposition adopted verbatim:
     "ship S-7 and S-9, hold S-8/S-10 pending these bounded corrections, rather than re-planning
     the whole cycle." No attempt 3 was submitted. Rationale in the banner below. -->

> ## Disposition (revision 4) — why there is no attempt 3
>
> Attempt 2 returned **FAIL with zero P0s** and an explicit reviewer disposition. Two independent
> constraints make adopting that disposition the correct action rather than spending a third gate:
>
> 1. **Plan-review budget.** Stage Step 4 permits a maximum of 2 re-entry cycles; a third
>    consecutive FAIL is a P-013.6 escalation, not another attempt
>    (`docs/memory/2026-09-18/021-001-T-harness-execution-model-decision-blocked.md`).
> 2. **The compound STOP RULE has fired.** `docs/compound/workflow-issues/`
>    `cross-artifact-contract-closure-requires-every-surface-2026-09-13.md` mandates a
>    compound-learning pass — not another gate — at the second consecutive same-class failure.
>    Attempt 2's Plan-1 review used the words "the same defect class as attempt-1's P0-2".
>
> **Adopted disposition:**
>
> | Unit | Source entries | Outcome |
> |---|---|---|
> | Unit 1 — unignore-checker correctness | `2787DA56`, `4C5BEC23` | **SHIPS** as S-7, after the shrinking corrections below |
> | Unit 2 — staging-gate test coverage | `3289EB69` | **HELD / blocked.** Its gap list was measured against the wrong file and the wrong AC numbering. Re-measurement is a prerequisite, not a revision. |
> | Unit 3 — topology-gate documentation | `F47DB9A9` | **SHIPS** as S-8 |
> | Unit 4 — diagnostics hygiene | `775E4A35`, `29DC2014` | **FOLDED into Unit 3** (S-8). Not a standalone release unit. |
>
> Every correction below **shrinks** this plan. Nothing was added to argue past a finding.

> **Revision 3 — corrections from plan-review attempt 1 (FAIL, 1× P0, 6× P1).** The P0 was the
> frontmatter: `requires_plan_hardening: no` was wrong, because
> `scripts/check-unignore-regression.sh` runs at `ci.yml:420` and `:426–431` **without**
> `continue-on-error`, inside job `gitignore-append-only`, which is in the required `ci-gate` job's
> `needs:` array (`ci.yml:610`; `ci-gate` is the branch-protection required check, `:596–606`).
> Unit 1 therefore modifies a **merge-blocking control** and is hardened in §6. Full resolution
> table in §7.

**Requires plan hardening: yes** — see above and §6.

## 1. Objective

Repair one latent correctness defect in a merge-blocking checker, close measured coverage gaps in
the staging-gate contract test, resolve two dangling CI documentation references, and discharge two
already-satisfied local-hygiene entries with a retention rule.

## 2. Unit 1 — Unignore-regression checker correctness (S-7)

**Problem, corrected (rev 3, was P1).** Revision 2 described a *live fail-open*. It is not live.
`root_gitignore_text_at(ref)` returns `""` when `git show {ref}:.gitignore` fails, so an invalid
base ref would be read as an empty baseline under which every deletion looks like an addition — but
`run_differential_check` first executes `git diff --name-only -z {base}{head}` and `raise SystemExit`
on failure (~`:293–301`) **before** ever calling `root_gitignore_text_at(base_ref)` (~`:316`), and
`run_denylist_check` only ever passes `ref="HEAD"`. So an invalid base ref aborts the run before the
empty-baseline path is reachable.

The defect is therefore **latent defence-in-depth**, not an exploitable fail-open: the function
cannot distinguish "`.gitignore` genuinely absent at this ref" from "this ref does not exist", and
any future caller reaching it without the preceding `git diff` guard would silently pass. Combined
with 4C5BEC23 (a header comment claiming `git show` "fails cleanly", which contradicts the
`except`-to-empty-string implementation), this is a real and worth-fixing maintenance hazard — just
not an urgent one. **Ranked accordingly: below the S-1…S-6 chain.**

*Citation correction (rev 4).* The stale-claim site is **not** `:27`. Lines `:25–28` describe an
abandoned `git worktree add --detach` design (matching M-9 of the deliberation). The comment U1-T1
most directly invalidates is the inline one at `:154–157`. U1-T4 and AC-1.5 are scoped to quoted
comment text rather than line indices for exactly this reason.

| ID | Task | Scope | Size | Cx |
|---|---|---|---|---|
| U1-T1 | Distinguish "ref invalid" from "`.gitignore` absent at ref" | `scripts/check-unignore-regression.sh`. Verify ref existence (e.g. `git rev-parse --verify`) before reading; raise a distinct error for an invalid ref; return `""` only for a genuinely absent file at a valid ref. | S | medium |
| U1-T2 | Add `--self-test` cases at the **function** level | *(rev 3, was P1)* An end-to-end scenario cannot reach the defect, because `run_differential_check` SystemExits first. The new cases must drive `root_gitignore_text_at` **directly** with (a) a valid ref whose `.gitignore` is absent → `""`, (b) an invalid/nonexistent ref → distinct error. | S | medium |
| U1-T3 | *(deleted in revision 4)* | `run_scenario` (`~:382–410`) is specialised to differential scenarios — it builds a fixture repo via `make_scenario_repo` and calls `run_differential_check`. U1-T2's cases are at the **function** level and never pass through it, so a new expectation kind would have **zero users**: dead configuration added to a merge-blocking script. If a harness change proves necessary, scope it to a direct-call assertion inside `do_self_test_scenarios` and leave `run_scenario` untouched. | — | — |
| U1-T4 | Correct the stale header comments **and the inline comment on the changed branch** | *(broadened in rev 3, extended in rev 4)* Not only `:27`. The header block carries stale claims at `:25–28` and `:28–32`/`:52–55`, **and** the inline comment at `:154–157` — "Absent at that ref is a legitimate (if unlikely) state -- treated as an empty ignore file, not an error" — sits on the very branch U1-T1 changes and is the comment the repair most directly invalidates. | S | low |

**Red phase (genuine, at the function level).** U1-T2 case (b) fails pre-change (invalid ref
silently yields `""`) and passes post-change.

### Acceptance criteria

* **AC-1.1** An invalid or nonexistent ref produces a distinct, non-empty error and never an empty
  baseline.
* **AC-1.2** A valid ref with no `.gitignore` still returns `""` — the legitimate case is preserved.
* **AC-1.3** Both new self-test cases exercise `root_gitignore_text_at` directly; case (b) fails
  pre-change.
* **AC-1.4** *(deleted in revision 4 with U1-T3.)* `run_scenario` is **untouched**; all pre-existing
  scenarios still pass unmodified.
* **AC-1.5** Every comment describing baseline-reading behaviour matches the implementation —
  expressed against **quoted comment text**, not line indices, since U1-T1 shifts them: the
  `TWO-PART CHECK` header block and the inline "Absent at that ref is a legitimate (if unlikely)
  state" comment on the changed branch.
* **AC-1.6** The checker's verdicts on the existing corpus are unchanged: this repairs a latent
  path, it does not change live behaviour.
* **AC-1.7** The `gitignore-append-only` job still passes and is still in `ci-gate`'s `needs:`.

## 3. Unit 2 — Staging-gate actor-contract test coverage — **HELD / NOT SHIPPED THIS CYCLE**

> **Status: blocked pending gap re-measurement.** This unit is **not** harvested into a queued
> shipment. Attempt 2 established that its entire gap list was measured against the wrong file and
> the wrong acceptance-criterion numbering. Re-measurement is a prerequisite, not a plan revision —
> writing a revision 5 on top of unreliable measurements is exactly the failure this cycle's
> compound STOP RULE forbids.

**What was refuted (attempt 2, independently verified):**

| Revision-3 claim | Measured reality |
|---|---|
| "AC-6's exactly-once property is asserted only as `assertPresentAtLeast(..., 1)`" | `staging_gate_actor_contract_test.go` contains **no `assertPresentAtLeast` call at all** — every row uses `assertPresent`. The cited call and the ordering assertions live in a **different harness**, `stage_branch_gate_test.go` (018.008-T). |
| "Add an exactly-once assertion helper" | `assertPresentExactly` **already exists** in package `integration` (`ship_feature_completion_contract_test.go:75`) and is already used by the sibling harness. `assertOrderedBefore` (`:92`) likewise. The task would have produced two duplicate helpers in one package. |
| "AC-7's ordering property is not asserted" | **No governing document defines an ordering property for AC-7.** Under the authoring plan AC-7 is the `STAGING_GATE_NO_ROUTE` tokenization; under `.backlogit/archive/027.001-T.md` it is "item 4's fence needles survive unchanged". The ordering ACs belong to 018.008-T. |
| "AC-8 has a positive needle `git fetch origin`" | That needle is **already satisfied** inside the asserted slice by Step 1.5 item 2 (`git fetch origin main`), so the assertion could not fail for the intended reason — a vacuous row. |
| "Relabel G1–G4 from `(AC-8)` to AC-6/AC-7" | **Two divergent numberings exist** for the same task, and the test file's own header cites the *plan* as needle authority. Executing the relabel as written swaps one misleading label for another. |

**Prerequisite for re-entry (the unblocking condition):**

1. Name the **governing AC numbering** — `.backlogit/archive/027.001-T.md` is the likely authority —
   and reconcile the test file's header citation to it in the same change.
2. Re-derive the gap list from `staging_gate_actor_contract_test.go` **plus** that governing record,
   not from the sibling `stage_branch_gate_test.go` harness.
3. Restate the exactly-once requirement against its real subject: **027.001-T AC-5's five halt
   tokens** (`UNCOMMITTED`, `LOCAL_MAIN_AHEAD`, `AWAITING_OPERATOR`, `OPERATOR_TIMEOUT`, `NO_ROUTE`),
   three of which the revision-3 scoping left uncovered.
4. Reuse `assertPresentExactly` / `assertOrderedBefore`; introduce no new helper.
5. Bound the AC-8 positive needle to the item-3(e) region and enumerate AC-8's four **absence**
   needles (`git checkout`, `git pull`, `git merge`, `git push`).

Source entry `3289EB69` is therefore carried as a **blocked backlog feature** with this list as its
acceptance-ready condition. It is deliberately *not* made falsely claimable.
## 4. Unit 3 — Pipeline-topology gate documentation (S-9)

**Problem.** `ci.yml:512` points at a rollout doc — specifically its "Threat Model & CODEOWNERS
Hardening" section — and `:531` names both that doc and an operational-runbook doc. **Neither file
exists.** Every CI comment the topology gate emits therefore directs a reader to a dead path.

**Blocked by S-1** (Plan 1 Unit 1): these docs describe the closure-artifact convention S-1 repairs.
Writing them first would document the wrong convention. See
`docs/plans/2026-09-18-intercom-go-ship-pipeline-contract-repair-plan.md`.

| ID | Task | Scope | Size | Cx |
|---|---|---|---|---|
| U3-T1 | Author the operational runbook | The doc named at `ci.yml:531`. Cover: what the gate checks, how to read a failure, how to remediate, and the advisory/required toggle. | M | low |
| U3-T2 | Author the rollout doc with the referenced section | The doc named at `:512`/`:531`. **Must** contain a "Threat Model & CODEOWNERS Hardening" section, since `:512` cites it by name. | M | low |
| U3-T3 | *(deleted in rev 3)* | Revision 2's repo-wide dangling-reference check was gold-plating with known false positives (`templates/ci/README.md`, prose fragments) and is removed. | — | — |
| U3-T4 | *(folded in from Unit 4, rev 4)* Verify and record the already-satisfied diagnostics state | Confirm at the bound snapshot that `scripts/diagnostic.ps1` exists git-tracked with no root copy (29DC2014) and that the crash report under `.autoharness/staging/` is ignored via `.autoharness/.gitignore:3` (775E4A35). Record as closure evidence; no file is moved or deleted. | XS | trivial |
| U3-T5 | *(folded in from Unit 4, rev 4)* Add the diagnostics-retention rule to the **U3-T1 runbook** | Attempt 2 established that **no qualifying standalone conventions document exists**: there is no `CONTRIBUTING.md`; `README.md` has only a "Configuration" section; and `AGENTS.md` (`:7` "Generated by autoharness") and `.github/copilot-instructions.md` are **generated**, so a rule appended there is silently reverted on regeneration — the exact residual already recorded for `_orchestrator.agent.md` in 027.001-T. The rule therefore lands as one short subsection of the runbook U3-T1 already creates: local diagnostics belong under an ignored directory; tooling scripts belong in `scripts/`; the root is not a scratch space. Cross-reference the existing `.autoharness/.gitignore` comment rather than restating it. | S | low |

### Acceptance criteria

* **AC-3.1** Both referenced paths exist at exactly the paths `ci.yml:512` and `:531` name.
* **AC-3.2** The rollout doc contains a section titled "Threat Model & CODEOWNERS Hardening".
* **AC-3.3** The runbook covers all four required subjects in U3-T1.
* **AC-3.4** Both docs describe the closure-artifact convention **as repaired by S-1**, not the
  pre-repair form.
* **AC-3.5** No `ci.yml` change is required by this unit; the paths are made true rather than
  rewritten.
* **AC-3.6** Both docs carry standard YAML frontmatter, consistent with the 1C6C3B46 convention this
  cycle applies.

## 5. Unit 4 — *(folded into Unit 3 in revision 4)*

**Both source entries are already physically satisfied** — `scripts/diagnostic.ps1` exists
git-tracked with no root copy (29DC2014), and the crash report under `.autoharness/staging/` is
ignored via `.autoharness/.gitignore:3` (775E4A35). What remained was a written retention rule.

Attempt 2 found this disproportionate as a standalone release unit: `U4-T1` verified an
already-true state (zero change surface) and `U4-T2` appended one prose subsection, yet shipping it
separately would cost a full claim / PR / review / closure cycle for a paragraph. It also found
that **AC-4.2's "existing conventions document" resolved to no real file** — the only candidates
are generated and would silently revert.

Both tasks are therefore **folded into Unit 3 as U3-T4 and U3-T5** (same documentation domain,
same reviewer, no oversizing risk under operator grouping rule 5). `775E4A35` and `29DC2014`
retarget from S-10 to **S-8**. There is no S-10 in this cycle.

## 6. Plan Hardening *(new in rev 3 — required by the P0 correction)*

**H-1 — Unit 1 modifies a merge-blocking control.** `check-unignore-regression.sh` runs at
`ci.yml:420`/`:426–431` with no `continue-on-error`, in `gitignore-append-only`, which is in
`ci-gate.needs` (`:610`); `ci-gate` is the required branch-protection check (`:596–606`). A
regression here blocks every PR. Mitigation: AC-1.6 requires verdict-invariance on the existing
corpus — the repair is confined to a path the live callers cannot reach — and AC-1.7 requires the
job still pass and stay wired.

**H-2 — Harness change ordering.** *(deleted in revision 4 with U1-T3.)* The constraint was
premised on `run_scenario` needing a new expectation kind, which attempt 2 showed U1-T2 never
passes through. `run_scenario` is untouched.

**H-3 — Characterization tests presented as defect repair.** Unit 2's gaps are in the test, not the
contract. *(rev 4)* This hazard proved worse than stated: attempt 2 showed the gap list itself was
measured against the wrong file. Mitigation: Unit 2 is **HELD**, not revised, with re-measurement
as its explicit unblocking condition (§3).

**H-4 — Documentation encoding a convention under concurrent repair.** Mitigation: Unit 3 is blocked
by S-1 and AC-3.4 binds its content to the repaired convention.

**H-5 — Hygiene work expanding into policy sprawl.** Mitigation: AC-4.2 forbids a new standalone
policy file; AC-4.3 forbids restating ignore rules.

**H-6 — Latent-defect over-ranking.** Revision 2 ranked Unit 1 as a live fail-open. Mitigation: §2
states the correct severity and §8 ranks this plan's units below the S-1…S-6 chain.

**H-7 — Blocking self-test step.** *(new in rev 4)* `ci.yml:420` runs `--self-test` and is **equally
merge-blocking**, so U1-T2's harness edits carry the same blast radius as the function change
itself. **Rollback criterion:** if AC-1.6 verdict-invariance cannot be demonstrated, revert U1-T1
and re-run `--self-test` before proceeding; do not land a partial repair.

**H-8 — Line-index acceptance criteria.** *(new in rev 4)* Criteria pinned to bare line numbers
become unverifiable the moment the change shifts them. Mitigation: AC-1.5 is expressed against
quoted comment text and named blocks.

## 7. Review findings applied (attempt 2 → revision 4)

Attempt 2 returned **FAIL with 0 P0 and 5 P1**. Every resolution below either **deletes** work or
**holds** it; nothing was added to argue past a finding.

| Finding | Sev | Resolution |
|---|---|---|
| U2-T1 mandates a new exactly-once helper, but `assertPresentExactly` (`ship_feature_completion_contract_test.go:75`) and `assertOrderedBefore` (`:92`) already exist | P1 | Unit 2 **HELD**; helper reuse is prerequisite item 4 |
| The "`assertPresentAtLeast(...,1)`" gap was measured against `stage_branch_gate_test.go`, a different harness; the file under repair has no such call | P1 | Unit 2 **HELD**; re-measurement is prerequisite items 1–3 |
| AC-2.3 required an "AC-7 ordering property" that no governing document defines | P1 | Deleted with the held unit; prerequisite item 1 forces the governing numbering to be named |
| U2-T4's relabel swaps one misleading label for another while two AC numberings are in conflict | P1 | Deleted with the held unit; prerequisite item 1 |
| U1-T3's new expectation kind would have **zero users** — dead config in a merge-blocking script | P1 | **U1-T3, AC-1.4 and H-2 deleted** |
| AC-4.2's "existing conventions document" resolves to no real non-generated file | P1 | Unit 4 folded into Unit 3; the rule lands in the U3-T1 runbook (U3-T5) |
| AC-2.1's `"git fetch origin"` needle already satisfied by Step 1.5 item 2 — vacuous | P2 | Deleted with the held unit; prerequisite item 5 bounds it to the 3(e) region and adds the four absence needles |
| AC-1.5 omitted the inline comment at `:154–157` on the changed branch; §2 mis-cited `:27` | P2 | U1-T4 and AC-1.5 broadened; §2 citation corrected |
| Frontmatter still pre-declared `PASS` | P2 | Status records the actual attempt-2 verdict and the adopted disposition |
| Unit 4 disproportionate as a standalone release unit | P2 | Folded into Unit 3; no S-10 in this cycle |
| H-1 omitted that the `--self-test` step is equally blocking; no rollback criterion | P3 | **H-7** added |
| Line-index ACs shift under their own edits | P3 | **H-8**; AC-1.5 restated against quoted text |
| U3-T1/U3-T2 at the top of the 2-hour envelope | P3 | Accepted; split the threat-model section into a third task if it exceeds a page |

## 8. Review findings applied (attempt 1 → revision 3)

| Finding | Severity | Resolution |
|---|---|---|
| `requires_plan_hardening: no` wrong — checker is merge-blocking via `ci-gate` | P0 | Flipped to `yes`; §6 added; AC-1.6, AC-1.7 |
| "Live fail-open" premise wrong — `run_differential_check` SystemExits first | P1 | §2 re-characterized as latent defence-in-depth; re-ranked (H-6) |
| Integration-level red phase unreachable | P1 | U1-T2 moved to the function level |
| `run_scenario` has no "expect error" expectation kind | P1 | U1-T3 added as a prerequisite; AC-1.4 |
| Stale header claims extend beyond `:27` | P1 | U1-T4 broadened to `:25–28`, `:28–32`, `:52–55`; AC-1.5 |
| U2 presented as defect repair; no genuine red phase | P1 | Declared characterization-only; AC-2.4 helper-level negative is the one red assertion |
| AC-8 scope too narrow (item (e) only; positive needle missing) | P1 | U2-T3 covers the full slice incl. `"git fetch origin"`; AC-2.1 |
| U3-T3 dangling-reference check is gold-plating with known FPs | P1 | Task deleted |
| Unit 4 risks a third restatement of ignore rules | P1 | AC-4.2/AC-4.3 constrain it to one subsection in an existing doc |
| G1–G4 mislabeled `(AC-8)` | P2 | U2-T4; AC-2.5 |
| U3-T2 had no content AC for the cited section | P2 | AC-3.2 names the section |
| S-1 dependency asserted without a path | P2 | §4 cites the S-1 plan path explicitly |
| Frontmatter pre-declared `PASS` | P2 | Status records the actual gate outcome |
| Unit-4 target document unnamed | P2 | U4-T2 names an existing conventions doc |

## 9. Ranking within the cycle

Both shipping units rank **below** the S-1…S-6 chain under the operator's priority policy: Unit 1 is
a latent (not live) correctness repair; Unit 3 is documentation plus folded hygiene. Execution
order within this plan: **S-7 → S-8**, with S-8 additionally blocked by S-1. Unit 2 ships nothing
this cycle and is carried as a blocked backlog feature.

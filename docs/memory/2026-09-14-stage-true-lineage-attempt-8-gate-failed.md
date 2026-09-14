# Stage session — TRUE-LINEAGE ATTEMPT 8 ran and FAILED

**Date:** 2026-09-14
**Agent:** Stage
**Branch:** `chore/stage-pipeline-policy-gap`
**Resumed from:** `checkpoint-20260914-052703.json` (stage-owned, active, operator-selected)

## Operator authorization (exact scope)

> `Authorize TRUE-LINEAGE ATTEMPT 8 and ADOPT D-1.`

Two grants, both now consumed:

1. **ADOPT `D-1`** — expand Option A so `.github/skills/shipment-reconcile/SKILL.md`'s
   pre-existing public `mode: safe-close` refuses/fails closed when invoked without a valid
   `classification_binding`, rather than falling through to mutation.
2. **Authorize exactly one** formal plan-review gate invocation (attempt 8). **Attempt 9 was
   explicitly not authorized.**

## What was done

1. **Resumed** `checkpoint-20260914-052703.json` under the owner-exclusive, fail-closed recovery
   protocol. `backlogit checkpoint list` was called with **no `status`/`agent` filter** (the
   fail-closed scan rule); all 22 checkpoints enumerated; no quarantine flags, no validation
   errors, no malformed `agent`/`status`. Ownership validated as `agent: stage`. The two older
   stage-owned active checkpoints were left untouched.
2. **Reauthored revision 12 → 13** (clean reauthoring, no append-only diary), adopting `D-1`
   across the compound invariant's full surface: producer/API (`§3.4.3 E`, the total 6-row
   mapping), consumers (`CS6`, and the newly discovered **`CS16`**), authority (`§2`, `§3.4.3 J`,
   `α18`, closure check 6), refusal/fallback (`R1`–`R5`), compatibility/migration (`§3.4.3 F`
   and the new **`(K)`**), tests (rows 23a–23d, 31, 32, 33 — table grew 30 → **33**), and
   Scope-T probes (**`T-13`, `T-14`, `T-15`**).
3. **Re-ran all pre-review mechanical probes**, which passed: Scope-A zero-guard 28/28 = 0;
   provenance 50 needles / 0 failures *(under whitespace normalization — see the lesson below)*;
   matrix 21 α-rows / 45 edges / 0 unresolved; H0 needle counts matched all 7 declared guard rows
   with no unexpected anomaly; fence balance even; single worktree; `backlogit doctor` exit 0.
4. **Invoked the gate exactly once**: 7/7 personas, `dispatch_mode: multi-agent`, anchor
   Architecture Strategist on `gpt-5.6-sol` / `high`.
5. **Recorded the FAIL** and stopped. No remediation, no attempt 9.

## Outcome

**FAIL. P0 = 0. P1 = 26 raw / 14 deduplicated. 7/7 personas, all FAIL.**
Verdict: `docs/reviews/2026-09-13-true-lineage-attempt-8-verdict.md`.

### Two discoveries that improved the plan (both products of the compound invariant working)

* **`CS16`** — probe `T-13` measured `_ship.agent.md` L782–783 (Step 6.1(b)): an **unbound**
  `mode: safe-close` invocation that revision 12 had never treated as a clause site. Adopting
  `D-1` without migrating it would have made the narrowing refuse Ship's *own primary close
  call* at runtime. Three reviewers independently re-measured and confirmed it.
* **`R5` branch-binding, not cascade-only** — revision 12 said a matching binding whose verdict
  "is not `CASCADE`" halts. Under an adopted `D-1` that makes the bound `SAFE_CLOSE` path
  unreachable, but P-015 (`workflow-policies.md` L439/L442/L443, measured at `T-15`) mandates
  the single-artifact safe-close procedure as the **default** close path. Cascade-only refusal
  would have made the policy's own default unreachable.

### Why it failed

**Not because of `D-1`.** Every persona that addressed the adoption found it operator-authorized
rather than self-granted, found that it *removes* destructive reachability (Constitution VII
strengthened), and found `CS16` necessary rather than scope creep. `P0 = 0`.

**Because the defect class relocated instead of closing.** The Learnings Researcher named it:
*"the contradiction has relocated to the measurement-probe and verification surfaces rather than
being closed, which is the same defect class one surface along, not a new one."* Revision 13
carried `D-1` through §3.4.3, §4.0 and §6.1 — and left §4.2 `CS11`/`CS3`, `AC-35`, `AC-39`,
§13.4's bounded table and `T-13`'s measurement scope on the prior contract.

Attempt 7: 27 raw / 15 deduped. Attempt 8: 26 raw / **14** deduped. The curve is flat.

## The 14 open themes (see the verdict for detail)

`N-1` R5 contradicted at §4.2 CS11/CS3 + AC-35 · `N-2` AC-39 certifies §4.0 against five wrong
checks, dropping the D-1 authority check · `N-3` §13.4's bounded table left on revision 12,
contradicting the zero-guard table on the same date · `N-4` T-13's three-file probe backs a
workspace-scoped claim, and `_ship.agent.md` L822 is a **direct `backlogit_ship_shipment` engine
call** a skill-side refusal cannot intercept · `N-5` independent falsifiability still incomplete
(rows 21c, 24a, 25, 28, 29, 30) · `N-6` §4.2 soft-wraps ~16 needles mid-span · `N-7` fact 2b has
no execution site · `N-8` S1a/S1b session boundary declared but unimplemented · `N-9` S2 performs
the authorized transition but never claims the shipment · `N-10` binding-integrity gaps (engine
identity, lock scope, serialization) · `N-11` ADD-1/CS10 equivalence not test-enforced · `N-12`
§10.1 merges drift with bound-refusal, bypassing the Stage handoff · `N-13` MCP/CLI parity gaps ·
`N-14` §15 still describes a two-file surface.

## Lessons for whoever picks this up

1. **My own provenance probe passed only under whitespace normalization**, which the plan's
   NON-NEGOTIABLE verbatim rule does not authorize (`N-6`, Go Reviewer `G8-1`). A green
   mechanical probe is only as strong as the rule it actually implements. Before attempt 9,
   either re-render §4.2 so every pinned needle occupies one unbroken line, or add an explicit
   normalization clause — and re-run.
2. **I updated §13.4's zero-guard table and missed the bounded table twelve lines below it**
   (`N-3`). Two reviewers caught it. When a section holds two tables that measure the same
   regexes with opposite expectations, they must be edited as one unit.
3. **`N-4` is the most consequential**: the claim "the sink is closed for every caller" is false
   in kind, not merely in scope. A direct `backlogit_ship_shipment` call never enters the skill,
   so `D-1` **narrows** the residual rather than eliminating it. The honest restatement is
   "closed for every caller that enters via `mode: safe-close`", with the direct-engine path
   disclosed as an inherited P-015 follow-up. **Do not widen scope to guard the engine command** —
   that is outside the operator's grant.
4. **§13.3 is the remediation diary the operator forbade** (`S8-4`), surviving as ~150 lines of
   per-attempt disposition tables while the header claims "no inline review narrative". It is
   already drifting. Reduce it to a one-line-per-theme index or move it into the verdict
   artifacts.

## State at session end

* `021-S` — `queued`, **NOT CLAIMABLE**, blocked. Authorization SPENT.
* `017-S` — `queued`, blocked on all three independent grounds; dependency `[021-S]` unchanged.
* `022-F` / `022.001-T` — `queued`, **NOT READY**.
* PR #54 — **OPEN and untouched**; remote head unchanged at `8999867ae42dcfa48ec9894a93393c352385ad0b`.
  **Nothing pushed.** The operator gated push/PR-update on attempt 8 passing; it did not.
* Local branch ahead and unpushed, as before.
* Bug report — left in the autoharness workspace; not touched, not recreated.
* Untracked `diagnostic.ps1`, `report.*.json`, `temp_commands.sh` — pre-existing, unowned, untouched.

## Next owner

**The operator.** The only open decision is whether to authorize a revision 14 and an attempt 9.
Stage did not open one and will not without an explicit grant.

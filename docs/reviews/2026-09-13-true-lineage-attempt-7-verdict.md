# Plan-review gate — TRUE-LINEAGE ATTEMPT 7

| Field | Value |
|---|---|
| Plan | `docs/plans/2026-09-13-intercom-go-ship-feature-completion-decided-plan.md` |
| Revision reviewed | **11** |
| True-lineage attempt | **7** |
| Route | OPTION A (operator-selected 2026-09-13) — read-only pre-mutation classification boundary |
| Personas | **7 / 7** (4 always-on + 3 cross-model, all triggered) |
| Anchor | Architecture Strategist — `openai` / `gpt-5.6-sol` / `high` |
| Degradation | none declared; full dispatch, no inline fallback |
| **Decision** | **FAIL** |
| Counts | **P0 = 0** · **P1 = 27 raw / 15 deduplicated** · P2 = 26 · P3 = 22 |
| Gate rule | PASS requires `P0 = 0` **and** `P1 = 0` (§13.1). P0 is met; P1 is not. |
| Authorization | Attempt-7 authorization is **SPENT**. No attempt 8 was opened and no post-review patch was applied. |

## Per-persona

| Persona | Model | P0 | P1 | P2 | P3 | Verdict |
|---|---|---|---|---|---|---|
| **Architecture Strategist** *(anchor)* | `gpt-5.6-sol` high | 0 | 5 | 1 | 0 | FAIL |
| Constitution Reviewer | caller | 0 | 3 | 5 | 8 | FAIL |
| Go Reviewer | caller | 0 | 2 | 7 | 5 | FAIL |
| Scope Boundary Auditor | caller | 0 | 3 | 8 | 6 | FAIL |
| Learnings Researcher | caller | 0 | 4 | 3 | 2 | FAIL |
| Security Lens Reviewer | `gpt-5.6-sol` | 0 | 3 | 1 | 1 | FAIL |
| Agent-Native Parity Reviewer | `gpt-5.6-sol` | 0 | 7 | 1 | 0 | FAIL |
| **Total** | | **0** | **27** | **26** | **22** | **FAIL** |

## What revision 11 achieved

**The attempt-6 P0 is closed.** `K-1` was `P0 = 1` at attempt 6; at attempt 7 **no persona raised a
P0**, and every persona that addressed the fourth file agreed the read-only boundary is the correct
remedy and that `CS11`–`CS14` are minimal and additive. The Scope Boundary Auditor recorded the
fourth file as *"not scope creep"*; the Architecture Strategist recorded the expansion as
*"cohesive."* The P0 → 0 transition is the first in this plan's seven-attempt lineage.

## `K-1`…`K-7` disposition at attempt 7

| ID | Attempt-6 severity | Attempt-7 outcome | Evidence |
|---|---|---|---|
| **K-1** | **P0** | **P0 CLOSED on the skill side; RECURS as P1 on the Ship side** | `M-1` — §4.1 pins replacement wording for CS7–CS10 and CS11–CS14 but **not for CS3–CS6**, which still authorize the cascade on *"the verdict reported by `mode: safe-close` Step 0(c)"* — the mutation-coupled boundary K-1 rejected |
| **K-2** | P1 | **RECURS** | `M-3` — CS5 borrows row 10 (assigned to CS2), CS6 shares row 4 with CS4, CS11's row 21a matches anywhere in the file. Whole-file substring assertions are not anchor-bounded |
| **K-3** | P1 | **RECURS** | `M-15` — rows 11, 12, 16 (and 3, 4, 5, 9d, 10) carry PRESENT needles that conflict with, or are absent from, the plan's own §4.1 pinned wording |
| **K-4** | P1 | **CLOSED** | No persona raised the `PA-021-CASCADE` anchor. `step [1]6` measures **0** across Scope A; §11 now reads "§9 steps 13a–15" |
| **K-5** | P1 | **CLOSED** | Security Lens: quarantine-commit + `git revert` *"is not, by itself, a third unapproved action… bounded to the two backlog trees, preserves history."* Constitution: *"materially improves K-5."* Residual P2/P3 only (terminology; missing `-m 1`) |
| **K-6** | P1 | **PARTIALLY CLOSED** | Plan-side two-site inventory (§9.3) verified correct. But `M-12` — the installed `_ship.agent.md` still orders intake **after** the claim, and CS7 is a pointer-text replacement that does not move it |
| **K-7** | P1 | **IMPROVED, RECURS DEEPER** | Scope T exists and caught two non-reproducing commands. But `M-8` — no probe of `backlogit doctor` (the sole containment mechanism), none of a feature `active → done` move (the grant's only mutation), none of the new classifier's behaviour |

## Deduplicated P1 themes (15)

| ID | Theme | Raised by |
|---|---|---|
| **M-1** | CS3–CS6 unpinned; Ship side still cites the Step 0(c) mutation-coupled verdict | Arch `A7-3`, Constitution, Parity `A7-04` |
| **M-2** | `CLASSIFICATION_BINDING` canonical serialization undefined; omits `shipment_id`, verdict, classifier/engine identity; residual window after compare | Arch `A7-2`, Security `SEC-11-04`, Parity `A7-01`, Scope `SB-5`, Learnings `L-8` |
| **M-3** | Independent falsifiability incomplete — shared/borrowed rows, unanchored assertions (`K-2` recurrence) | Arch `A7-4`, Go `G-1`, Scope `SB-2`/`SB-3` |
| **M-4** | Rows 23b/24b do not assert backward compatibility; suite stays green if a public mode is rewritten | Arch `A7-5`, Parity `A7-07` |
| **M-5** | Boundary is **optional at the destructive sink** — unbound `safe-close` can still reach CASCADE | Security `SEC-11-03` |
| **M-6** | CS10's pinned P-010 wording lands a dangling *"five conditions below"*; shipment-active is not among the conditions | Security `SEC-11-01`, Parity `A7-05`, Go `G-13` |
| **M-7** | §14 sizing: the 106-min S1 stretch excludes the 10-min H0/H1 gate work §9 requires **pre-merge** | Scope `SB-1`, Arch `A7-6` |
| **M-8** | Scope T omits the three load-bearing probes: `doctor`, feature `active → done`, classifier behaviour | Learnings `L-2`/`L-3`/`L-4` |
| **M-9** | §13.5 still miscites its prior art — attributes a *"declare your count scope"* reading that does not exist in the source | Learnings `L-1` |
| **M-10** | Clause (B) *"verbatim"* conflicts with clause (H) `BLOCK`; no exhaustive condition → verdict mapping | Arch `A7-1`, Parity `A7-02` |
| **M-11** | `PA-021-CASCADE` does not cover validated **linked deliberations** the cascade may archive | Security `SEC-11-02` |
| **M-12** | Pre-mode-before-claim is plan prose only; installed agent ordering unchanged (`K-6` residue) | Parity `A7-06` |
| **M-13** | Principle XI merge-strategy verification not named at both merges; `git revert` of a merge prescribed without `-m 1` | Constitution |
| **M-14** | §3.4.2's *"fails fact 1 or fact 2"* claim is false as written — SKILL.md is one of the four files, so any change inside it satisfies fact 2 | Constitution |
| **M-15** | PRESENT needles conflict with the plan's own normative wording (`K-3` recurrence) | Go `G-2` |

## Structural reading

The failure is **not** a repeat of attempt 6. `K-1` moved from `P0` to closed-on-the-skill-side, and
`K-4`/`K-5` closed outright. What attempt 7 exposes is that **Option A was applied to one half of the
contract**: the skill now offers a read-only boundary, but the Ship-side clauses that must *consume*
it (`CS3`–`CS6`), the policy bullet that must *authorize* the outcome (`CS10`), and the sink that must
*refuse* an unbound cascade (`M-5`) were not brought along. Seven of the fifteen themes (`M-1`, `M-3`,
`M-4`, `M-5`, `M-6`, `M-12`, `M-15`) are that one shape.

The second cluster (`M-2`, `M-8`, `M-10`) is the `K-7` measurement lesson applying to itself once more:
the plan asserts tool and digest behaviour it has not probed. The Learnings Researcher's `L-9` records
that **six failed gates have produced zero harvested compound learnings**, which is why this class
keeps recurring.

## Disposition

* **No attempt 8 was opened.** Attempt-7 authorization is spent.
* **No post-review patch was applied** to the plan, the four records, or any instruction surface.
* Revision 11, the four resynced records and this verdict are committed as the terminal state.
* `021-S` remains **queued / unclaimed / not claimable**. `017-S` remains **queued and
  dependency-ineligible**. No shipment was assembled and no harvest was performed — the §13.1
  harvest-ready gate (`P0 = 0` **and** `P1 = 0`) is not met.
* **Next authorization required from the operator.** Stage cannot self-authorize attempt 8.

## Reproduction

All seven persona passes were dispatched as subagents with the full plan, the attempt-6 verdict, the
four backlog records, the three instruction surfaces and the constitution. Persona transcripts were
captured this session; the deduplicated themes above carry each persona's own finding IDs so any
theme can be traced to its source pass.

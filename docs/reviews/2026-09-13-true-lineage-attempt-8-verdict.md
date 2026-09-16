# True-lineage attempt 8 — plan-review verdict

| | |
|---|---|
| Plan | `docs/plans/2026-09-13-intercom-go-ship-feature-completion-decided-plan.md` |
| Revision under gate | **13** |
| Authorization | GRANTED 2026-09-14 (operator; `D-1` ADOPTED) — **SPENT** |
| Gate ran | **yes**, 2026-09-14 |
| `dispatch_mode` | `multi-agent` |
| `decision` | **FAIL** |
| P0 | **0** |
| P1 | **26 raw / 14 deduplicated** |
| P2 | 18 raw |
| P3 | 9 raw |
| Personas | **7 / 7** (4 always-on + 3 cross-model, all triggered) |
| Anchor | Architecture Strategist — `openai` / `gpt-5.6-sol` / `high` |
| Attempt 9 | **NOT authorized.** Do not open one. |

## Per-persona

| Persona | Model | P0 | P1 | P2 | P3 | Verdict |
|---|---|---|---|---|---|---|
| **Architecture Strategist** *(anchor)* | `gpt-5.6-sol` high | 0 | 5 | 1 | 0 | FAIL |
| Constitution Reviewer | caller | 0 | 3 | 5 | 3 | FAIL |
| Go Reviewer | caller | 0 | 2 | 5 | 2 | FAIL |
| Scope Boundary Auditor | caller | 0 | 2 | 2 | 3 | FAIL |
| Learnings Researcher | caller | 0 | 4 | 2 | 1 | FAIL |
| Security Lens Reviewer | `gpt-5.6-sol` high | 0 | 7 | 1 | 0 | FAIL |
| Agent-Native Parity Reviewer | `gpt-5.6-sol` high | 0 | 3 | 2 | 0 | FAIL |

**Persona coverage:** all seven dispatched as subagents with the full plan, the attempt-7 verdict,
the compound learning, the four backlog records, the three instruction surfaces and the
constitution. `TOOL_DEGRADED: agent-native-parity-reviewer identity file absent at
`.github/agents/subagents/` — declared fallback: the Persona Rubric Adapter focus applied via
general-purpose dispatch on the cross-model route.` No persona was skipped.

## The headline

**The `D-1` adoption itself is sound and is not the reason this attempt failed.** Every persona
that addressed it agreed: the narrowing is operator-authorized rather than self-granted, it
*removes* destructive reachability rather than adding it (Constitution VII is strengthened), the
`R1`–`R5` refusal contract fails closed on missing, invalid, drifted and bound-`BLOCK` input, and
**`CS16` is a necessary consequence of the narrowing, not scope creep** — three personas
independently re-measured `_ship.agent.md` L783 and confirmed that without `CS16` the adopted
refusal would reject Ship's own primary close call.

**The attempt failed because the defect class relocated instead of closing.** Attempt 7 returned
27 raw / 15 deduplicated P1s, all of one class: a cross-artifact contract closed on some surfaces
while others stayed on the prior contract. Attempt 8 returns 26 raw / **14** deduplicated, and the
Learnings Researcher named the pattern precisely — *"the contradiction has relocated to the
measurement-probe and verification surfaces rather than being closed, which is the same defect
class one surface along, not a new one."* Revision 13 propagated `D-1` correctly through
§3.4.3 (E)/(J)/(K), §4.0 and §6.1, and **failed to propagate it** through §4.2 `CS11`/`CS3`,
`AC-35`, §13.4's bounded table and `T-13`'s scope.

This is the reappearance the compound learning's stop rule exists to catch. It is recorded as
such, and no attempt 9 is opened.

## Deduplicated P1 themes (14)

| Ref | Theme | Raised by |
|---|---|---|
| **N-1** | **`R5` branch-binding is contradicted on three surfaces.** §3.4.3 (E)/(J) route a matching `SAFE_CLOSE` into steps 1–10, but §4.2 `CS11` still says safe-close halts on *any* bound non-`CASCADE` verdict, §4.2 `CS3` halts on `SAFE_CLOSE`, and `AC-35` repeats the non-`CASCADE` halt. P-015's mandated default close path is left unreachable or ambiguous — the exact failure `T-15` was measured to prevent | Arch `A8-1`, SecLens `SL8-05`, Parity `AP8-01` |
| **N-2** | **`AC-39` certifies §4.0 against the wrong check set.** It asserts "all **five**" checks and enumerates a weaker five; §4.0 defines **six**. It drops check 5 (probe + verbatim output) and **check 6** — the authority check that makes the `D-1` narrowing authorized rather than self-granted. Under `AC-39` as written, several other findings here pass undetected | Arch `A8-4`, Const `C8-2`, Scope `S8-3`, Learn `L8-4` |
| **N-3** | **§13.4's "Bounded, not zero" table was left on revision 12** and contradicts the zero-guard table directly above it on the same date: `30 [r]ows` and `15 [s]ites` are required ×0 above and ≥6 below, and `attempt [8]` appears twice with incompatible cells, the second demanding every occurrence be qualified `NOT RUN` / `candidate` — which condemns §13.1's own authorized wording | Scope `S8-1`, Learn `L8-1` |
| **N-4** | **`T-13`'s caller inventory is three-file but its claim is workspace-scoped.** §3.4.3 (F)/(K), `AC-46`/`AC-47` and §15 VII generalise to "closed for **every** caller" / "exactly two invocations **in this workspace**", while the probe never enumerates `.github/**` (30 skills, three agent files). Worse, **`_ship.agent.md` L822 is a direct `backlogit_ship_shipment` engine cascade call, not a `mode: safe-close` invocation** — a skill-side refusal cannot intercept it, so the residual is *narrowed, not eliminated* | Const `C8-5`, Learn `L8-2`, Scope `S8-2`, Parity `AP8-05` |
| **N-5** | **Independent falsifiability still incomplete (`M-3`/`K-2` recurrence).** Row 21c counts a token that three other sites' pinned wording also emits, so the `mode` enum cell at `SKILL.md:56` can go unedited with all 33 rows green; row 24a carries a `CS13` literal under `CS14` provenance inside the very table that certifies nothing is borrowed; rows 25/28/29/30 are whole-file or lack an end delimiter where block-bounding is claimed | Arch `A8-3`, Go `G8-2`, `G8-4`, Learn `L8-3` |
| **N-6** | **§4.2's pinned wording is soft-wrapped mid-needle at ~16 rows.** The NON-NEGOTIABLE rule says *verbatim, character for character*; applying §4.2 as literally written therefore leaves those rows RED at H1. §4.0's recorded "0 provenance failures" holds only under a whitespace normalization the plan never authorizes | Go `G8-1` |
| **N-7** | **§3.4.2 fact 2b has no execution site.** It is described as the only control detecting unreviewed in-merge edits inside `SKILL.md`, but §9.2 check 7 and `AC-29` perform only contract/file-set/working-tree identity checks. The load-bearing review-diff comparison is never run | Arch `A8-5`, SecLens `SL8-03` |
| **N-8** | **The S1a/S1b session boundary is declared but not implemented.** §7 and §14 promote it to a checkpointed *session* boundary; §9 has no checkpoint-write step for it, §9.1 scopes crash-resumption to S2 only, and `AC-12` still asserts "the **two** sessions are distinct" against `AC-44`'s three | Const `C8-3`, Scope `S8-6` |
| **N-9** | **The authorized transition happens in a session that never claimed the shipment.** ADD-1/`CS10` scope the grant to "the shipment **this session has claimed**", but a1 executes in fresh session S2, which *resumes* an S1-claimed shipment. The privilege predicate is unsatisfied in the only session that exercises it | SecLens `SL8-01` |
| **N-10** | **Binding integrity gaps.** `engine=backlogit --version` is a version string, not engine identity — a replaced binary passes revalidation; the single-writer lock covers `queue/{shipment_id}.md` while the verdict consumes and the cascade mutates feature, descendant, deliberation and archive records, so the TOCTOU bound does not cover the classified state; and the canonical serialization leaves dependency-tuple encoding and version-output normalization undefined | SecLens `SL8-02`, `SL8-04`, `SL8-08`, Arch `A8-6` |
| **N-11** | **ADD-1 / `CS10` equivalence is asserted but not test-enforced.** Rows 2a–2d, 20a and 28 check the grant phrase, scope and set-equality conjunct only; P-010 can omit or alter the completeness, no-live-descendant, containment and feature-status conditions with every row still green | SecLens `SL8-06` |
| **N-12** | **§10.1's recovery row merges "binding drift / bound refusal"** and sends both back to reclassification, contradicting the rule that only drift may reclassify. This creates a path that resumes after a `BLOCK` or unknown bound verdict without the required Stage handoff | SecLens `SL8-07` |
| **N-13** | **MCP/CLI parity is incomplete on two executable surfaces.** The a1 outcome contract is defined only as CLI exit codes 0/6/7/8 with no MCP mapping, so an MCP-first caller cannot deterministically emit `A1_MOVE_BLOCKED` / `_CONFIG_ERROR` / `_BUSY`; and the full descendant graph is specified as repeated `backlogit_get_item` over `parent_id`, which is an ID-addressed point lookup that cannot enumerate unknown children | Parity `AP8-02`, `AP8-03` |
| **N-14** | **§15's Constitution cells still describe a two-file surface.** Principle I asserts "no I/O beyond reading two files" against a three-file suite, and Principle II's NON-NEGOTIABLE test-first declaration omits `shipment-reconcile/SKILL.md` — the file applied *first* in S1a and the only one carrying the non-additive narrowing | Const `C8-1` |

## Notable P2s

* **`S8-4` / the operator's own constraint:** §13.3 (≈150 lines of `M-`/`K-`/`J-` disposition
  tables) is an inline findings diary, while the header banner and §13.2 both claim the plan
  "carries no inline review history". It is already drifting — a `J-12` cell still reads "§6.1
  pins all **30** rows". This is the append-only remediation diary the operator forbade,
  surviving as a disposition table.
* **`G8-5`:** the `CS11` anchor as pinned resolves ×0 — the installed cell escapes its intra-cell
  pipes (`` `pre` \| `post` ``) and the pinned anchor does not.
* **`G8-7`:** `CS15`'s "renumbered without any other edit" leaves the fallback-path
  cross-reference to "primary-path step 6" dangling, with no guard row.
* **`C8-6`:** `PA-017-CASCADE`'s escrow release (§11.1) omits the condition-7 linked-deliberation
  equivalent that `PA-021-CASCADE` carries.
* **`G8-8`:** §6 names `testing`/`strings` as the stdlib set; `os` and `path/filepath` are
  unavoidable.

## What did close at revision 13

Recorded so the next revision does not re-litigate settled ground:

* `D-1` is adopted with a real authority trail — §2, §3.4.3 (J), §13.1, `AC-42`, §16 and `021-S`
  all carry the 2026-09-14 grant, and §13.4's α18 probe verifies the withholding wording is gone.
  **No persona found the narrowing self-authorized.**
* `CS16` is confirmed necessary and correctly anchored by three independent re-measurements.
* `R1`–`R4` fail closed on missing, malformed, stale and bound-`BLOCK` bindings; `classification_binding`
  is a real required input, not prose; the four refusal tokens are discoverable.
* `PA-021-CASCADE`'s conditions are conjunctive and fail-closed; the §10.3.2 unwind contains no
  destructive primitive; `PA-017-CASCADE` remains non-exercisable under this plan.
* Scope: no persona found quiet widening beyond the operator's grant, and the Scope Boundary
  Auditor explicitly declined to raise the approval's "no scope expansion" condition, judging it
  to bound the cascade action's runtime target rather than the plan's file surface.
* The attempt-6 P0 stays closed; **attempt 8 returned P0 = 0**.

## Disposition

`021-S` and `022-F` / `022.001-T` remain **NOT CLAIMABLE**. `017-S` remains blocked on all three
of its independent grounds. Harvest-readiness is unchanged: it requires a gate that has actually
RUN returning **P0 = 0 and P1 = 0**; this gate ran and returned P1 = 14 deduplicated.

**No remediation was attempted in this session and no attempt 9 was opened**, per the operator's
authorization, which covered exactly one gate invocation. The next owner is the **operator**, who
must decide whether to authorize a revision 14 and an attempt 9.

Per-persona transcripts were captured this session; every deduplicated theme above carries the
originating persona finding IDs so any claim can be traced to its source pass.

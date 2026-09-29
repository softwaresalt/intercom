# Stage session — true-lineage attempt 6 reauthoring and gate FAIL

**Date**: 2026-09-13 · **Agent**: Stage · **Branch**: `chore/stage-pipeline-policy-gap`
**Session**: `stage-20260913-true-lineage-attempt-6-reauthor`
**Resumed from**: `.backlogit/checkpoints/checkpoint-20260913-231952.json`

## Operator direction

Stop append-only patching. Reauthor the governing plan and its coupled backlog records into a
clean corrected present-state version, relying on Git/PR history for lineage. Run **exactly
one** further gate as true-lineage attempt 6. No attempt 7.

## What was done

**Recovery.** Resumed `checkpoint-20260913-231952.json` under Stage owner protocol — explicit
single-checkpoint operator selection, ownership validated (`agent: stage`), explicit
confirmation. `checkpoint-20260913-194412.json` and `checkpoint-20260913-084125.json` left
active and untouched, as instructed.

**Tooling.** CLI fallback / degraded mode (no MCP surface in this harness); `backlogit` v1.10.1.
`backlogit sync` → 240 artifacts, `parse_failures=0`. `backlogit doctor` → no issues.
`markdownlint-cli2` → 0 issues.

**Reauthored (not patched):**

* `docs/plans/…-decided-plan.md` revision 9 → **10**, rewritten from scratch. ~180 lines of
  inline review narrative removed; lineage moved to §13.2 audit pointers. Surface expanded
  2 → 3 files; clause inventory converted from ordinal ranges to **anchor addressing**, 8 → 10
  sites; test 18 → 20 rows with **all** rows pinned; §9.1 recovery algorithm withdrawn in favour
  of pointer-only delegation; §10.3.1 added defining both path inventories; §11.1 added binding
  `PA-017-CASCADE` to escrow.
* All four backlog records (`021-S`, `017-S`, `022-F`, `022.001-T`) rewritten from stacked
  append-only blocks into single clean current-state bodies. Frontmatter preserved exactly:
  statuses stay `queued`, `017-S.dependencies = [021-S]` intact, `022.001-T` sizing fields intact.

**Verification performed before the gate** (all mechanical, all passing):

* All 11 clause anchors resolve **exactly once** in the file the plan names.
* All §4.2 ordinals exact against the live tree.
* All §13.4 Scope-A measurement claims verified true with self-match-safe regexes.
* All §6.1 H0 counts verified against `_ship.agent.md` and `workflow-policies.md`.

## Gate result — FAIL

7/7 personas returned, anchor `openai`/`gpt-5.6-sol`/`high`. **P0 = 1, P1 = 7 deduplicated.**
Full record: `docs/reviews/2026-09-13-true-lineage-attempt-6-verdict.md`.

Per operator direction the plan was **not patched** in response. Only the gate-state marker,
§13.1 readiness row, the lineage table row, and the four records' readiness lines were updated
to state the FAIL truthfully.

## The lesson (the important part)

Fourteen of the eighteen attempt-5 findings were genuinely resolved, and every anchor, ordinal
and count in the reauthored plan verified clean. The gate still failed, on **one** finding —
`J-5` — whose remediation moved the defect rather than removing it.

`J-5` said: *the plan delegates to a classify API the skill does not expose.* Revision 10
retargeted the delegation to `mode: safe-close` Step 0(c) and read the verdict back from the
produced report. That is where the verdict genuinely lives. But nobody measured **when** it
becomes readable. It is produced at cascade sub-procedure step 5 — *after* the destructive call
— and on the safe-close path `SKILL.md` L510 continues straight into mutation and never emits a
verdict token at all. So the plan's central guarantee, "HALT before any close mutation," is
unreachable, and §10.1's claim that nothing was mutated is false.

**Root cause across six gates.** The measurement scope was always drawn around the surface the
author controls. `H-3`, `H-6`, `J-5`, `J-9` and now `K-1` all failed on the **installed tooling**
surface, which §13.4 never placed in any declared scope. The plan states the rule — *name the
scope, give the regex, give the number* — and then never applies it to the one scope that keeps
breaking it. Revision 10 even cited the right compound document for this and paraphrased it into
a weaker lesson ("declare your count scope") than the one it actually teaches ("independently
reproduce a claim about installed behaviour").

A corollary worth keeping: **reauthoring is not remediation.** Compaction made the artifact
honest and readable, and it genuinely closed the sweep-scope defect class by removing the
surface that kept going stale. It could not close a defect whose evidence lives outside the
artifact entirely.

## State at session end

| Item | State |
|---|---|
| `021-S` | `queued`, unclaimed, **not claimable** |
| `017-S` | `queued`, ineligible on three independent grounds |
| `022-F` / `022.001-T` | `queued`, not harvest-ready |
| Shipment assembly | **not performed** — gate did not pass, so there is nothing to hand to Ship |
| PR #54 | OPEN, remote head unchanged at `8999867`; **not pushed** — gates do not permit |
| Working tree | plan + 4 records modified, 1 review artifact + this memory added |
| Untracked, preserved | `diagnostic.ps1`, `report.20260913.140024.22564.0.001.json`, `temp_commands.sh` |

## Next owner

**Operator.** Stage's authorization is spent: the reauthoring is done and the single authorized
gate has been run and recorded. Stage will not open attempt 7 without new explicit direction.

The decision in front of the operator is `K-1`, and it is a scope question, not a wording
question: either `shipment-reconcile` gains a read-only pre-mutation classification boundary —
a **fourth file**, and Ship-runtime work outside Stage's role boundary — or the plan's safety
model is rewritten from *prevent* to *detect and restore*. The remaining six blockers are
tractable plan-local edits once that is settled.

---
title: "Stage — compound root-cause remediation; attempt-8 candidate prepared, NOT authorized"
date: 2026-09-13
agent: Stage
phase: compound-root-remediation-complete-attempt-8-awaiting-authorization
plan: docs/plans/2026-09-13-intercom-go-ship-feature-completion-decided-plan.md
plan_revision: 12
shipment: 021-S
feature: 022-F
task: 022.001-T
commit: 91d427b
---

# Session record — compound learning + root remediation to plan revision 12

## Authorization boundary for this session

The operator authorized a **compound-learning / root-cause pass before attempt 8**, the
durable learning, root remediation of the plan and coupled backlog records, and any surgical
planning-surface expansion needed to describe Option A coherently.

The operator **explicitly did not authorize** spending true-lineage attempt 8, and instructed
that the formal plan-review gate must not be invoked, the counter must not be reset, and no
fake PASS/PENDING marker may be created. **All three constraints were honoured.**

## Was the loop circular? Yes — proven, not inferred

Each cycle's remediation **created** the next cycle's finding by relocating the contradiction
to an adjacent surface, and each cycle's closure evidence was a measurement structurally
incapable of seeing that surface. Recurrence chains named by the verdicts themselves:

| Chain | Surface it kept moving to |
|---|---|
| `J-5` → `K-1` → `M-1` | delegation target |
| `J-12` → `K-2` → `M-3` | needle coverage |
| `J-2`/`J-7` → `K-4` → `M-12` | scope of the sweep |
| `H-3`/`J-2`/`J-9` → `K-7` → `M-8` | measurement scope |
| `K-3` → `M-15` | needle satisfiability |

**Mechanical evidence gathered this session (the decisive proof):**

1. §4.1 (old) pinned replacement wording for **9 of 14** clause sites. `CS1`–`CS6` had none —
   while the producer sites `CS11`–`CS14` were pinned **in the same revision that adopted
   Option A**.
2. **8 of 26 transition-row PRESENT needles** (rows 3, 4, 5, 9d, 10, 11, 12, 16) were **not
   verbatim substrings of any pinned wording**. Row 11 pinned `scoped to Role Boundary changes
   only` while the only governing text reads `scoped to **Role Boundary** changes only` —
   emphasis markers inside the span. These would be **red at `H1`**.
3. §6.1's own *"independent falsifiability"* table assigned **CS3** a needle in the **CS4**
   block (row 9d) and **CS5** a needle in the **CS2** block (row 10) — borrowed needles in the
   very table certifying no borrowing.
4. §6.1 asserted *"re-measured at revision 11 … every PRESENT needle ×0"* — a **true but
   worthless `H0`-absence count** that cannot detect any of (1)–(3).

**Root cause.** Closure of a contract spanning producer, consumers, authority grant, refusal
path, tests and tooling probes was asserted from a measurement scoped only to the producer or
only to the reporting artifact. Such a measurement is structurally incapable of detecting a
stale consumer, authority, refusal, test-needle or probe surface, so each gate re-reported the
same defect at whichever surface was still stale, under a new label.

## Deliverable 1 — the compound learning

`docs/compound/workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md`

Complete YAML frontmatter per the `compound` skill contract (`title`, `description`, `source`,
`doc_type: learning`, `problem_type`, `category: workflow-issues`, `component`, `root_cause`,
`resolution_type: design_change`, `severity: high`, `message`, `file_path`, 13 `citations`,
8 `tags`), plus `## Problem`, `## Root Cause`, `## Resolution`, `## Prevention`.

**The invariant:**

> A cross-artifact contract change is complete only when the **producer**, **every consumer**,
> the **authority grant**, **every refusal/fallback path**, the **tests**, and the
> **measurement probes** are all updated and verified against the **same bound contract
> snapshot**. A local zero-occurrence result, or any check scoped only to the producer or only
> to the artifact where the finding was reported, **cannot prove closure**.

**Anti-circularity techniques recorded:** the contract-surface matrix/graph; before-and-after
executable probes (a probe that cannot go from red to green is not evidence); and the **stop
rule** — when a finding reappears under a new label, stop patching and build the matrix.

## Deliverable 2 — plan revision 12 (reauthored, not appended)

| Change | Closes |
|---|---|
| **NEW §4.0 contract-surface matrix** — 3 contracts, 27 rows, each with current evidence / intended change / executable verification / dependency, plus 5 mechanical closure checks | `M-15` root |
| §4.2 now pins **every** site (was 9 of 14); §6.1 gains a **Provenance** column | `M-1`, `M-15` |
| `CS3`–`CS6` retargeted to `mode: classify-close-path` | `M-1` |
| **`CS15` added** — relocates Step 0.5 intake **before** the claim | `M-12` |
| `CS10` renders all five conditions **inline** + shipment-active scoping clause | `M-6` |
| §3.4.3 (D) full canonical binding serialization + disclosed residual window | `M-2` |
| §3.4.3 (B1) exhaustive condition→verdict mapping | `M-10` |
| §3.4.2 fact 2b; false "fails fact 1 or fact 2" claim withdrawn | `M-14` |
| §5 `backlogit move` exit-code contract (0/6/7/8) | `M-8`/`M-16` |
| §6/§6.1 **24 → 30 rows**; CS3/CS5/CS6 get own-block needles (25/26/27) | `M-3`, `M-4` |
| §10.4 `git revert -m 1`; §15 XI verifies merge shape at **both** merges | `M-13` |
| §11 condition 7 — linked-deliberation set measured empty | `M-11` |
| §13.5 miscitation withdrawn | `M-9` |
| §13.6 **T-9…T-12** added (first actual Scope-T probes) | `M-8` |
| §14 sizing → **S1a / S1b / S2** execution split, no manifest change | `M-7` |
| §12 AC-3/AC-5/AC-27 corrected; **AC-39…AC-44** added | all |

**14 of 15 themes root-fixed. `M-5` is partially fixed with a named blocker.**

## Residual blocker `D-1` (operator decision required)

Closing the unbound destructive sink **inside the skill** would narrow a pre-existing public
mode, contradicting §2's operator-approved *"surgical and additive"* Option-A scope. Stage
adopted the **in-scope half** (`CS6` forbids Ship from invoking `safe-close` without a
binding) and recorded the skill-side change as a **named blocker** rather than inventing
permission. Recorded at §3.4.3 (J), §13.1, `AC-42`, and on `021-S`.

## Decisions that avoided voiding an operator approval

* **`M-7` was NOT fixed by splitting the task.** A second task becomes a depth-1 live
  descendant of `022-F`, forcing a third manifest member, live-mismatching `PA-021-CASCADE`
  condition 3 and voiding the approval under §11's "no re-scoped version" clause. Fixed by a
  **session split** instead, which changes no record.
* **`M-11` was closed by measurement, not re-scope.** `backlogit link list` → `[]` for all
  three manifest members; the engine's linked-deliberation regex → 0 matches; no
  `source_deliberation_id`; index holds zero deliberation artifacts.

## Validation evidence (pre-review only — the gate was NOT run)

| Check | Result |
|---|---|
| `backlogit sync` | `Indexed 240 artifacts` — **INDEX_SYNC_OK** |
| `backlogit doctor` | `No issues found`, exit **0** |
| Scope-A zero-guard sweep (19 regexes) | **all 0** |
| Provenance probe — transition PRESENT needles ⊂ §4.2 | **40 checked, 0 failures**; 6 anchor/guard needles correctly excluded by Provenance cell |
| Matrix edge resolution (closure check 4) | **56 edges, 0 broken** |
| `H0` ABSENT needles resolve in live tree | **all ×1** (guards ×0) as declared |
| `H0` PRESENT needles red | **41 at ×0**; anchors/guards present as declared |
| §6.1 distinct numbered rows | **30** |
| Markdown fence balance | even in both files |
| Worktree topology | **single worktree** (P-016) |
| Working tree | only planning artifacts + backlog records modified — **no `.github/` file touched** (P-010) |

**The provenance probe and the matrix-edge probe both failed on first run and were fixed.**
That is the invariant working: revision 11 shipped because no such probe existed.

## State at session end

* **Commit `91d427b`** on `chore/stage-pipeline-policy-gap`. **Not pushed. PR #54 untouched**
  (still OPEN at head `8999867`).
* Gate state: last gate that **RAN** = attempt 7, **FAIL**. Revision 12 = **candidate for
  attempt 8, NOT RUN, NOT AUTHORIZED**. Counter not reset. No PASS/PENDING marker written.
* `021-S` queued / not claimable. `017-S` queued / blocked on three independent grounds.
* `PA-021-CASCADE` and `PA-017-CASCADE` both intact, unexercised, manifests unchanged.

## Next step

**Await explicit operator authorization to spend true-lineage attempt 8.** Do not invoke
plan-review without it.

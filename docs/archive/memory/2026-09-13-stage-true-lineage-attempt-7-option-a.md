# Stage session — TRUE-LINEAGE ATTEMPT 7 (Option A)

**Date:** 2026-09-14
**Agent:** Stage
**Branch:** `chore/stage-pipeline-policy-gap`
**Restored from:** `.backlogit/checkpoints/checkpoint-20260914-014612.json`
**Outcome:** attempt-7 gate **FAIL** (P0 = 0, P1 = 27). Authorization spent. Stopped without a patch cycle.

## Entry state

Operator selected **Option A** (prevention-first: add a read-only pre-mutation classification
boundary to the installed `shipment-reconcile` skill) over Option B (rewrite safety as
detect-and-restore), and authorized the fourth-file scope expansion plus **exactly one** further
gate as true-lineage attempt 7.

Attempt 6 had returned FAIL with `P0 = 1` (`K-1`) and `P1 = 7` (`K-2`…`K-7`) over plan revision 10.

## Recovery protocol executed

* `backlogit checkpoint list` run **unfiltered** — 19 checkpoints, `needs_quarantine: 0`,
  `quarantined: 0`. No validation or quarantine anomaly, so the fail-closed gate passed.
* Partitioned to `agent: stage` + `status: active` → 3 candidates. Operator had named a single
  filename, satisfying unique selection.
* `checkpoint get` → `valid: true`, `agent: stage`. Owner validation passed.
* `agent-engram` installed and reachable (`engram stats` → 297 symbols / 86 code files) →
  **ENGRAM_OK**, so bounded prune-on-restore applied rather than a fail-closed halt. The three
  never-prune classes were preserved: the active cursor (`021-S` / `022-F` / `022.001-T`), the
  unresolved-checkpoint pointer, and the recorded gate verdicts for attempts 1–6.
* Tool gate: `backlogit` 1.10.1 `TOOL_OK`; index sync `INDEX_SYNC_OK` (240 artifacts).

## Work performed

**Plan reauthored revision 10 → 11** (surgical edits, no append-only narrative):
§2 four-file surface · §3.4.1 retargeted to the read-only mode · §3.4.2 blob pin withdrawn and
replaced by three post-merge facts · **§3.4.3 new** — the Option-A contract, clauses (A)–(I) ·
§4/§4.1/§4.2 extended to `CS11`–`CS14` · §6/§6.1 rebuilt at 24 rows with measured H0 counts and a
per-site falsifiability table · §7 ordering · §9 **new step 14a** classification boundary, step 1
intake moved before the claim · **§9.3 new** (`K-6`) · §10.3.2 step 2 rewritten to
quarantine-then-`git revert` (`K-5`) · §11 execution site corrected to steps 13a–15 (`K-4`) ·
§12 `AC-33`…`AC-38` added · §13.3 `K` disposition table · §13.4 three scopes · **§13.6 new**
Scope-T measurement table (`K-7`) · §14 sizing · §15 constitution · §16 out-of-scope · §17 history.

**Four backlog records resynced** to revision 11 / attempt 7 / four-file surface / 14 clause sites /
24 rows, with `021-S` kept **queued and unclaimed** and `017-S` kept **queued and
dependency-ineligible**.

**Measurement discipline actually applied.** Running the Scope-T commands rather than asserting
them caught two that did not reproduce: `backlogit query-sql` does not exist (the command is
`backlogit query`, and it rejects non-SELECT), and `Select-String '^### .*Mode'` returns five
headings, not four. Both were corrected before the gate. All 14 clause anchors were re-verified ×1
and the Scope-A regex table was re-measured after the records changed.

## Gate result

7/7 personas dispatched, anchor = Architecture Strategist on `gpt-5.6-sol` / `high`, no degradation.
**All seven returned FAIL. P0 = 0 across every persona** — the attempt-6 P0 is closed — but
P1 = 27 raw / 15 deduplicated.

Full verdict: `docs/reviews/2026-09-13-true-lineage-attempt-7-verdict.md`.

## The lesson this session produced

Option A was **applied to one half of the contract**. The skill now offers a read-only boundary, but
the Ship-side clauses that must consume it (`CS3`–`CS6`), the policy bullet that must authorize the
outcome (`CS10`), and the sink that must refuse an unbound cascade were not brought along. Seven of
the fifteen P1 themes are that single shape. When a contract is split across a producer and a
consumer, fixing the producer is not fixing the contract.

The second cluster is `K-7` recurring against itself: Scope T was added, and it still omits probes of
`backlogit doctor` (the sole containment mechanism), of the feature `active → done` move (the grant's
only mutation), and of the new classifier's behaviour. The Learnings Researcher recorded that **six
failed gates have produced zero harvested compound learnings**, which is the mechanism by which this
class keeps recurring. That is the strongest candidate for a compound entry before any attempt 8.

## Disposition

* No attempt 8 opened; no post-review patch applied.
* No harvest, no shipment assembly — the `P0 = 0` **and** `P1 = 0` rule is not met.
* Committed locally. **Not pushed** — the operator gated push/PR update on a PASS.
* `checkpoint-20260914-014612.json` resolved (resume confirmed successful).
  `checkpoint-20260913-194412.json` and `checkpoint-20260913-084125.json` left **active**.
* Next action requires **fresh operator authorization**. Stage cannot self-authorize attempt 8.

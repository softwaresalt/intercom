# 017-S Ship Execution Attempt — Halted at S-1 (Regime A, zero mutation)

**Session:** Ship, 2026-09-17T18:49–18:57 PDT
**Shipment:** 017-S ("Stage artifact branch/PR policy gap correction")
**Governing plan:** `docs/plans/2026-09-16-intercom-go-017-S-closure-plan.md` (revision 15, on
`origin/main`), canonical contract `docs/plans/2026-09-16-017-S-closure-canonical-contract.md`.

## Outcome

**HALTED before the claim.** No shipment claim, no branch, no backlog mutation, no source edit.
`017-S` remains `queued`, unclaimed, unchanged.

## What ran (all read-only, pre-claim)

1. Session-start checkpoint scan: `backlogit checkpoint list` — zero `agent: ship` checkpoints
   exist (zero-candidate normal startup per the Crash-Resumption Protocol; nothing to
   restore/resume). Confirmed Stage's own checkpoints are as the operator described:
   `checkpoint-20260917-235351.json` and `checkpoint-20260917-234624.json` both `status:
   resolved`; five other Stage checkpoints remain `active` and were not touched.
2. S-0 safety-mode declaration (`careful` + `freeze-scope`) and the operator's verbatim P-002
   deviation authorization token, recorded in
   `docs/plans/evidence/2026-09-16-017-S-closure/run-log.md`.
3. Preflight gates **PF-1 through PF-12** — all **PASS**, independently measured this session
   (engine identity digest, plan-on-origin/main byte-identity + PR #57 merge-commit parent-shape
   verification, `backlogit doctor`, manifest/topology, §3.3 allowlist, linked-deliberation set,
   both cascade probes, topology gate, 021-S shipped ground, installed-file Step 6.1(a1) ground,
   close-path authority, harness-architect admission). Full evidence in the run log.
4. **S-1 pre-claim branch check FAILED**: the canonical D-H6 exclusion pathspec
   (`git status --short -- . ':(exclude).backlogit/' ':(exclude)docs/plans/evidence/2026-09-16-017-S-closure/'`)
   still reports 6 pre-existing untracked entries with no plan-authorized carve-out:
   `diagnostic.ps1`, `report.20260913.140024.22564.0.001.json`, `temp_commands.sh`, and three
   `docs/memory/**` files. Neither the closure plan nor the canonical contract names these paths
   or extends D-H6 to cover them. Per the operator's explicit instruction, these were preserved
   as-is — not absorbed, deleted, or stashed.
5. Per plan §8.1, `S-1` failure is an explicitly named **Regime A** halt case identical to a
   `PF-1…PF-12` failure: "no claim performed... nothing to unwind... the only cheap-exit window."
   Halted here accordingly.

## Next step (for Stage or operator)

All PF-1…PF-12 evidence remains valid for a future attempt. To unblock S-1, either:
- amend the closure plan's D-H6 exclusion pathspec to explicitly cover the 6 listed paths
  (or the directories that contain them), with review, **or**
- have the operator/Stage relocate, `.gitignore`, or otherwise dispose of those specific
  untracked files through a reviewed, plan-authorized mechanism (Ship must not invent one).

No branch was created; no time was spent inside the one-way-door claim; `018-F` / `018.008-T`
and the 11 archived members are untouched.

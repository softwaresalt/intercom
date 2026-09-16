# 021-S — GREEN / pre-commit checkpoint

**Session**: Ship execution of shipment `021-S` (covering feature `022-F`, task `022.001-T`).
**Plan**: `docs/plans/2026-09-13-intercom-go-ship-feature-completion-decided-plan.md`, revision 13 (frozen).
**Timestamp**: 2026-09-16 ~13:32 local (post-GREEN, pre-commit).

## State at this checkpoint

- Branch: `feat/021-s-ship-covering-feature-completion-foundation`, no commits yet (about to make the first).
- Shipment `021-S`: `active` (claimed this session).
- Feature `022-F`: `active` — activated as a side effect of the backlogit engine's
  `shipment claimed` cascade (audit log: `.backlogit/logs/022-F.jsonl`, event
  `status_changed`, `reason: "shipment claimed"`, `from: queued`, `to: active`,
  timestamp `2026-09-16T12:34:22-07:00`). This happened automatically when `021-S` was
  claimed; it is **not** a Ship-initiated feature-completion action, and the new a1
  covering-feature completion gate has **not** been exercised this session (self-host
  constraint honored: feature stays `active`, not `done`).
- Task `022.001-T`: moved `active -> done` this session (archived: `.backlogit/archive/022.001-T.md`).
- Shipment `017-S`: untouched, still `queued` (verified via `backlogit get 017-S`).

## Work completed (RED → GREEN, full plan revision 13 scope)

1. **Behavioral test harness** (`tests/integration/shipment_close_path_harness_test.go`):
   disposable `backlogit`-backed fixture workspace; Go reimplementation of
   `classifyClosePath` (read-only classifier + 6 BLOCK-reason mapping) and `safeClose`
   (binding revalidation R1–R5 + dispatch); whole-tree SHA-256 snapshot/equality helper
   with no exemption list.
2. **Behavioral tests** (`tests/integration/shipment_close_path_behavior_test.go`): Z1
   (whole-tree equality across CASCADE/SAFE_CLOSE/all 6 BLOCK reasons), B4 (no mutation
   until verdict), B1 (4 refusal tokens, zero mutation), B5 (workspace drift → refusal),
   B2 (bound CASCADE archives exactly manifest subtree), B3 (bound SAFE_CLOSE = single-
   artifact delta, reachable path preserved), B6 (direct engine path guarded), Z2/Z3/Z4
   (return-value-only fields, no persistence). All GREEN, including 2 genuine RED→GREEN
   bug fixes discovered honestly during development (fixture misclassification; binding
   not computed on BLOCK paths).
3. **33-row static contract test** (`tests/integration/ship_feature_completion_contract_test.go`):
   confirmed RED against unedited instruction files (26/44 assertions failing, matching
   the plan's stated H0 baseline exactly), then confirmed GREEN after all instruction-
   surface edits below.
4. **Instruction-surface edits** (CS1–CS16, ADD-1, ADD-2 from plan §4.2, using the exact
   pinned wording as substring needles):
   - `.github/skills/shipment-reconcile/SKILL.md`: CS11 (mode enum +
     `classification_binding` input row), CS12 (new `### Classify-Close-Path Mode`
     read-only section), CS13 (Step 0 dispatch: bound revalidation clause, D-1 unbound
     refusal clause, migration clause, replaced unconditional default with bound-only
     `SAFE_CLOSE` branch), CS14 (new read-only Behavioral Constraint bullet).
   - `.github/agents/_ship.agent.md`: ADD-1 (Role Boundary covering-feature completion
     grant, 5 conditions, scoping clause), ADD-2 / new Step 6.1 item `a1` (S0–S5 selector,
     move exit-code table), CS1 (context reload: P-015 + merge-commit + no-authority-
     widening + Role-Boundary-only scope), CS2 (direct status prescription replaced by
     pointer to skill's authoritative close sequence), CS3 (`Do NOT call ship` gated on
     `CLOSE_PATH_VERDICT: CASCADE`; HALT on unknown/SAFE_CLOSE/BLOCK/ambiguous), CS4
     (delegated classification — Ship performs no root/coverage/per-member/fallback
     predicate of its own; old re-derived P-015 predicate paragraph removed), CS5
     (carry `CLASSIFICATION_BINDING` into the close call), CS6 (Ship never invokes
     unbound `safe-close`; treats all 4 refusal tokens as a HALT with no mutation), CS7
     (Step 0.5 intake check: pointer-only delegation to `mode: pre`, no per-item
     predicate, `record-consistent` + orphan scan preserved), CS8 (same pointer-level
     delegation for the pre-archive gate at `expected_status: done`), CS9 (protected-set
     claim scoped to the safe-close path only — cascade has none by construction), CS15
     (intake reconciliation relocated to run **before** the claim, as new item `3b`,
     preserving the `Scope note (139-F/139.001-T)` and orphan scan verbatim), CS16 (Step
     6.1(b) invocation carries the binding; old unbound invocation literal removed).
   - `.github/policies/workflow-policies.md`: CS10 (P-010 `Ship MAY` bullet extended with
     the same covering-feature completion grant + 5 conditions, verbatim-matching ADD-1;
     no new policy clause, no release-instance ID added).
5. **Quality gates**: `gofmt -l .` empty; `go vet ./...` clean; `go build ./...` clean;
   `go test ./...` full suite green (all packages, ~126s); `backlogit doctor` — "No issues
   found"; `backlogit sync` — clean reindex.
6. **Self-review pass** (in place of a separate review-skill invocation in this
   environment): read the full diffs of all three instruction files end-to-end; no P0/P1
   issues found. Two pre-existing/expected observations recorded, not remediated as
   in-scope defects:
   - `022-F` transitioning `queued -> active` was the backlogit engine's own
     `shipment claimed` cascade (audit-logged), not a Ship classification action — expected
     and consistent with the plan's a1-gate S4 design (which expects `active` at that gate).
   - The installed file already contained a duplicate `3a.` step label (pre-existing, not
     introduced by this change); the new intake-reconciliation item was inserted as `3b.`
     to avoid colliding with either existing `3a.`, since renumbering the whole Step 0.5
     list was assessed as higher-risk than an additional lettered sub-step (documented
     equivalent per the plan's "unless repository mechanics require a documented
     equivalent" allowance).

## Explicitly NOT done this session (by design — self-host / merge-gate constraints)

- The new a1 covering-feature completion gate has **not** been exercised: `022-F` remains
  `active`, not `done`. That transition is reserved for a **fresh session** after this PR
  merges, per the user's explicit self-host constraint (this session must not use the new
  authority it is introducing).
- No shipment closure (`021-S` safe-close/cascade) has been attempted — that also waits for
  a fresh post-merge session.
- No merge has occurred and none will be attempted without a fresh, explicit operator
  approval for this specific PR (PR #54's approval does not carry forward).

## Next steps

1. Commit the coherent change set (see file list below), excluding pre-existing untracked
   hygiene files and PR #54's prior memory files.
2. Push the branch, create the implementation PR with a `## Local Review Readiness` block.
3. Manage CI/review feedback within bounded protocols; do not merge without new explicit
   operator approval for this PR.
4. Stop at the merge-approval gate and report.

## Files in this commit

- `.backlogit/queue/021-S.md` (status `queued -> active`, this session's claim)
- `.backlogit/queue/022-F.md` (status `queued -> active`, engine claim cascade)
- `.backlogit/archive/022.001-T.md` (new; task moved `active -> done` and archived)
- `.github/agents/_ship.agent.md`
- `.github/policies/workflow-policies.md`
- `.github/skills/shipment-reconcile/SKILL.md`
- `tests/integration/shipment_close_path_harness_test.go` (new)
- `tests/integration/shipment_close_path_behavior_test.go` (new)
- `tests/integration/ship_feature_completion_contract_test.go` (new)
- `docs/memory/2026-09-16-125455-021-s-red-confirmed-checkpoint.md` (new, prior checkpoint)
- `docs/memory/2026-09-16-133200-021-s-green-precommit-checkpoint.md` (this file)

## Files explicitly excluded (pre-existing untracked hygiene, preserved untouched)

- `diagnostic.ps1`
- `report.20260913.140024.22564.0.001.json`
- `temp_commands.sh`
- `docs/memory/2026-09-14-ship-pr54-copilot-review-block-halt.md`
- `docs/memory/2026-09-15-ship-pr54-copilot-review-cycle3-merged.md`
- `docs/memory/2026-09-15-ship-pr54-copilot-review-remediation-cycle2-halt.md`

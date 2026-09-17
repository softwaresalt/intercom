---
title: "Canonical decision contract — 017-S closure"
date: 2026-09-16
status: canonical
agent: Stage
shipment: 017-S
covering_feature: 018-F
governs: docs/plans/2026-09-16-intercom-go-017-S-closure-plan.md
---

# Canonical decision contract — `017-S` closure

**Purpose.** This is the single, internally consistent list of decisions and actions chosen for
`017-S` closure. The governing closure plan is a remaster of *this* contract. Where the plan and
this contract disagree, **this contract is authoritative and the plan is defective**.

**Scope note.** This document records *decisions*, not their history. Prior revisions,
corrections and review findings live in Git history, PR #57, and the review evidence — not here
and not in the operational procedure.

---

## A. Identity and frozen scope

| ID | Decision |
|---|---|
| **D-A1** | The release unit is shipment **`017-S`**; its covering feature is **`018-F`**. |
| **D-A2** | The manifest is **frozen at exactly 13 members** and is never added to, removed from, or substituted: `018-F`, `018.008-T`, `018.001-T`, `018.001.001-ST`, `018.001.002-ST`, `018.001.003-ST`, `018.002-T`, `018.002.001-ST`, `018.002.002-ST`, `018.002.003-ST`, `018.003-T`, `018.003.001-ST`, `018.003.002-ST`. |
| **D-A3** | Exactly **two members are live**: `018-F` (`queued`) and `018.008-T` (`queued`). The other **11 are pre-archived** and exempt under prerequisite §3.3. |
| **D-A4** | The governing deliberation is `docs/decisions/2026-09-13-intercom-go-a-only-simplification-decision.md` (verdict `SIMPLIFICATION_VALID`, option O-2). The prerequisite plan is `docs/plans/2026-09-13-intercom-go-ship-feature-completion-decided-plan.md`, **revision 13, FROZEN**; its attempt-8 FAIL stands and is not re-opened. |
| **D-A5** | This plan **inherits** prerequisite §3.1 (declared status is lifecycle, location is integrity), §3.2 (five conjunctive grant conditions) and §3.3 (pre-archived exemption + the 11-ID allowlist). It amends none of them and creates **no new verdict, token, classifier or policy surface**. |
| **D-A6** | Authority to close is **Ship's**, exercised under this plan. Stage authors, gates and measures; Stage does **not** claim, implement, merge, or invoke the cascade. |

## B. Eligibility grounds

| ID | Decision |
|---|---|
| **D-B1** | Ground 1 (dependency on `021-S`) is **discharged**: `021-S` is `archived` with `archived_status: shipped`, commit `74330d3`, merge `f2d4cf9`, PR #56. Re-verified mechanically at preflight. |
| **D-B2** | Ground 2 (parent-completion capability) is **discharged**: `0cb4a42` landed Ship Step 6.1(a1) and the §3.2 grant. Re-verified against the **installed** files at preflight. |
| **D-B3** | Ground 3 (governing closure plan) is discharged **only** when the remastered plan passes its review gate and is present on `origin/main`. Until then `017-S` is **not claimable**. |
| **D-B4** | Ground 4 (decision §7 item 4 probe) is **discharged** by Probe 25 at `e36d853`. No new probe is authored by this plan. |
| **D-B5** | No ground is ever waived by assertion; each is re-measured by Ship at invocation. A Stage-time measurement is evidence about a different moment and is **never inherited**. |

## C. Measured tooling ground truth

| ID | Decision |
|---|---|
| **D-C1** | **The claim is a ONE-WAY DOOR.** `backlogit shipment` exposes exactly `add`, `claim`, `create`, `get`, `list`, `return-blocked`, `ship`. There is **no** unclaim/release/return-to-queued path. The only exits from `active` are `shipped` and `abandoned`. |
| **D-C2** | `shipment return-blocked` returns a **single item** and therefore **breaks the 13-member manifest**. It is forbidden on this route. |
| **D-C3** | `abandoned` is terminal, destructive, **not** covered by `PA-017-CASCADE`, and has **no operator approval**. It is forbidden on this route. Any abandon requires a new explicit approval and is outside this plan. |
| **D-C4** | The generic `backlogit move` route **refuses shipment status writes with exit 9**. It is never the close path. |
| **D-C5** | `backlogit move <id> --status archived` is **not** an archive operation — it silently no-ops. The archive operation is `backlogit archive <id>`. No step archives via `move`. |
| **D-C6** | The cascade close command is `backlogit shipment ship <id> --sha <s> --message <m> --author <a>`. Its flags are exactly `--sha`, `--message`, `--author`. |
| **D-C7** | **No binding parameter exists on either close surface** — CLI flags as in D-C6; registry `ship_shipment.params` = `shipment_id`, `sha`, `message`, `author`. The only surface accepting `classification_binding` is `shipment-reconcile` `mode: safe-close`. |
| **D-C8** | `backlogit link list` takes a **positional** ID; the `--id` flag form no longer resolves on 1.10.1. |

## D. Close-path authority

| ID | Decision |
|---|---|
| **D-D1** | Installed Ship requires the returned `CLASSIFICATION_BINDING` to be carried into the close call. By D-C7 the direct cascade op accepts no binding. Therefore **`mode: safe-close` carrying a `CASCADE` binding is the only executable realization of installed Ship's own mandate**, not an alternative to it. The same `backlogit shipment ship` operation runs either way. |
| **D-D2** | A **direct, unbound** `backlogit shipment ship` on this route is a **HALT** — it would discard the mandated binding and bypass the Cascade Close Sub-Procedure's gates. |
| **D-D3** | **No prerequisite Ship release unit is required.** Harmonizing the installed Ship text is a clarity improvement, captured out-of-scope as stash entry **`11B75632`**, not a correctness precondition. |
| **D-D4** | This reconciliation is made **falsifiable at preflight** (PF-11). If the cascade op ever gains a binding parameter, the reconciliation is **void** and the route halts to Stage for re-measurement. Drift in the SKILL's bound-`CASCADE` routing also halts. Drift in the *installed Ship prose* toward the guarded path is **recorded, not halted** — a correction of the known-defective clause must not fail the route. |
| **D-D5** | `classify-close-path` is **read-only** and must not write a reconcile artifact or mutate any state before returning its verdict. Its verdict must be read **before any close call**. |

## E. Happy path (canonical sequence)

**Preflight gates `PF-1…PF-11`** run in order on the pre-claim branch, before the claim. Any
failure halts with **zero mutation**.

| ID | Gate |
|---|---|
| **PF-1** | Engine identity: version + binary SHA-256 match the recorded digest. Mismatch invalidates every measurement ⇒ halt to Stage. |
| **PF-2** | This plan resolves on `origin/main`; body above `## Plan Review` is byte-identical to the reviewed revision; the **most recent** review attempt records `decision: PASS` with named persona coverage and is not superseded. |
| **PF-3** | `backlogit doctor` ⇒ `No issues found.`, exit 0. |
| **PF-4** | Manifest = 13; descendant graph ∪ `{018-F}` set-equal to manifest; every `parent_id` resolves transitively. |
| **PF-5** | The archived-declaring descendant set is **exactly** the 11-ID allowlist, each satisfying §3.3 conditions 1–4. Any off-list archived descendant ⇒ halt. |
| **PF-6** | Linked-deliberation set for all 13 members verified **EMPTY** by three independent methods. Records `019.007-T`'s pre-state **for provenance only — informational, never an assertion baseline**. |
| **PF-7** | Read-only re-verification of the two committed probe artifacts and the Probe 26 script, plus engine agreement. Authors nothing, mutates nothing, adds no backlog item. |
| **PF-8** | Topology gate `--phase pre_claim` ⇒ exit 0. |
| **PF-9** | Ground 1: `021-S` `archived` **and** `archived_status: shipped` with commit provenance. `archived` alone is insufficient — it is also the terminal state of an abandoned shipment. |
| **PF-10** | Ground 2: the **installed** `_ship.agent.md` carries Step 6.1(a1) with the S0–S5 selector and `workflow-policies.md` carries the §3.2 grant. |
| **PF-11** | Close-path authority agreement per D-D4. |

**Steps `S-1…S-22`.** Monotonic, each ID unique.

| ID | Step | Binds |
|---|---|---|
| **S-1** | Pre-claim branch created; P-011 branch discipline and P-016 worktree classification verified; clean worktree. | — |
| **S-2** | Pre-claim reconcile `mode: pre`, `expected_status: queued`. Continue only on `PROCEED`. | — |
| **S-3** | **Claim `017-S`** via the CLI surface `backlogit shipment claim 017-S`. **One-way door.** | — |
| **S-4** | Re-read `018-F`; **must be exactly `active`**. The claim performs this transition; no agent writes it. | — |
| **S-5** | Archived-member non-drift check over all 11: archive-located, `status: archived`, `archived_status` unchanged, `parent_id` unchanged. | — |
| **S-6** | Topology gate `--phase post_claim` ⇒ exit 0. | — |
| **S-7** | Harness first: generate the failing harness for `018.008-T`, observe and record **red before any implementation line**. | `{harness_paths}` |
| **S-8** | Implement `018.008-T` to green, scope-locked to its recorded scope. | `{impl_paths}` |
| **S-9** | `018.008-T → done`. Discharges §3.2 conditions 1 and 2. | — |
| **S-10** | Implementation PR: open, review, P-018 engagement, operator-approved **merge commit**. Changed-path allowlist asserted. | — |
| **S-11** | Merge confirmation gate: bind and verify the merge. | `{merge_sha}`, `{msg}`, `{author}` |
| **S-12** | Closure branch created **before any S2 mutation**; P-011/P-016 re-verified; no S2 mutation on `main`. | — |
| **S-13** | `a0` topology gate `--phase lifecycle` ⇒ exit 0. | — |
| **S-14** | `a1` covering-feature completion gate: S0→S5 selector in order; `018-F` `active → done` at S4. | — |
| **S-15** | Pre-cascade baseline commit on the **existing** closure branch, with verified non-empty staged diff. | `{pre_cascade_sha}`, `{pre_paths}` |
| **S-16** | Pre-archive reconcile `mode: pre`, `expected_status: done`. Continue only on `PROCEED`. | — |
| **S-17** | `classify-close-path` (read-only). Required verdict **`CASCADE`**. | `CLASSIFICATION_BINDING` |
| **S-18** | **Bound cascade close** via `mode: safe-close` carrying the binding. | — |
| **S-19** | **Unconditional `{post_paths}` capture**, immediately on return from S-18. | `{post_paths}` |
| **S-20** | Postchecks — all must pass before anything is committed. | — |
| **S-21** | Commit the cascade result. | `{cascade_commit_sha}` |
| **S-22** | Closure PR: P-014 local review gate, operator-approved **merge commit**, verified on `origin/main`. Changed-path allowlist asserted. | `{closure_merge_sha}` |

## F. Variable bindings (defined before first use)

| Variable | Bound at | First consumed at |
|---|---|---|
| `{harness_paths}` | **S-7** | S-10 |
| `{impl_paths}` | **S-8** | S-10 |
| `{merge_sha}` | **S-11** | S-18 |
| `{msg}`, `{author}` | **S-11** | S-18 |
| `{pre_cascade_sha}` | **S-15** | R-6 |
| `{pre_paths}` | **S-15** | S-20 |
| `{post_paths}` | **S-19** | S-20 |
| `{cascade_commit_sha}` | **S-21** | R-4 |
| `{closure_merge_sha}` | **S-22** | R-5 |
| `{quarantine_sha}` | **R-3** | R-3 |

## G. Evidence capture timing and anchors

| ID | Decision |
|---|---|
| **D-G1** | `{pre_paths}` is captured at **S-15** against `{pre_cascade_sha}`: full sorted path+blob-hash inventory of `.backlogit/queue/` + `.backlogit/archive/`, plus raw `019.007-T` frontmatter. |
| **D-G2** | `{post_paths}` is captured at **S-19**, **immediately on return from the cascade call** — before P-007 verification, before `mode: post`, before any commit, cleanup or postcheck-driven restore. |
| **D-G3** | **The S-19 capture is UNCONDITIONAL** — taken on every return, success or failure. Rationale: installed Step 6.1(c)/(d) restores sit between the close and post-mode and mutate exactly the inventoried trees, so a later capture would compute the scope assertion over an already-restored tree. |
| **D-G4** | The `019.007-T` untouched-assertion is anchored to the **S-15** baseline at both S-20 and R-7. **Never to PF-6** — the S-9 dependency carve-out permits `019.007-T` to change legitimately between PF-6 and S-15, so a PF-6 anchor guarantees a false-positive unwind. |
| **D-G5** | Record equivalence is read from **raw frontmatter in the artifact files** — never a list, index or query view. |

## H. Path allowlists

| ID | Decision |
|---|---|
| **D-H1** | **S-10 (implementation PR), part 1 — source/test/config:** changed paths MUST equal `{harness_paths}` ∪ `{impl_paths}`. |
| **D-H2** | **S-10, part 2 — backlog records:** confined to `.backlogit/queue/017-S.md`, `018-F.md`, `018.008-T.md`, **plus `019.007-T.md` when and only when the engine's own dependency-unblock produced it**, evidenced in the run log by a `018.008-T → done` transition immediately preceding the delta, with the delta confined to dependency/status fields. |
| **D-H3** | The `019.007-T` carve-out exists because `019.007-T` (`status: blocked`, `dependencies: [018.008-T]`, parent `019-F`, **outside the manifest**) may legitimately unblock at S-9, which runs on the pre-claim branch and therefore lands inside the S-10 PR diff. Without the carve-out the gate is unsatisfiable by construction and would halt **after** the one-way door. |
| **D-H4** | **S-22 (closure PR):** changed paths confined to `.backlogit/queue/` and `.backlogit/archive/` records of the 13 manifest members plus the `017-S` record, plus any evidenced `019.007-T` dependency delta. Anything else is scope drift ⇒ halt. |
| **D-H5** | **S-20 cascade-scope assertion:** `{post_paths} △ {pre_paths}` is confined to the 13 manifest members and the `017-S` record. Any out-of-manifest artifact archived or deleted ⇒ `HALT — cascade detected, revert required`. |
| **D-H6** | Anything outside an applicable allowlist is scope drift under PA-017 condition 9 ⇒ halt. |

## I. Merge shape

| ID | Decision |
|---|---|
| **D-I1** | **P-009 is Merge-Commit-Only.** Both the S-10 and S-22 merges MUST be true merge commits. Squash-merge and rebase-merge are **forbidden**. |
| **D-I2** | S-11 verifies the S-10 merge empirically: **exactly two parents**, parent 1 = mainline, parent 2 = the reviewed PR head. Any other shape is a **P-009 violation ⇒ halt**. |
| **D-I3** | `{merge_sha}` is bound from the **PR-reported merge commit** (`gh pr view <n> --json mergeCommit`), never from `git rev-parse origin/main` — the remote tip equals the merge only if nothing else lands in between, and an ancestry test can never detect that error. |
| **D-I4** | **P-009 governs PR merges only.** The single-parent `{cascade_commit_sha}` produced at S-21 is an intra-branch commit and is **not** a P-009 signal. |

## J. Failure semantics

| ID | Decision |
|---|---|
| **D-J1** | Every gate and step has an explicit failure disposition. There is no implicit continue. |
| **D-J2** | **Regime A — `017-S` is NOT `active`** (pre-claim, or a failed claim that left it `queued`): mutate nothing, record the failure token and step ID, **halt to Stage**. No unwind needed or performed. |
| **D-J3** | **Regime B — `017-S` IS `active`**: stop; record the failure token verbatim, the step ID, and live status of `017-S`, `018-F`, `018.008-T` and the 11 allowlisted members; **leave `017-S` `active` with zero backlog mutation**; **halt to the operator**. Disposition of the claim is the operator's decision, never Ship's improvisation. |
| **D-J4** | In Regime B, **forbidden**: `return-blocked` (D-C2), `move` on the shipment (D-C4), `abandoned` (D-C3), and R-1…R-8 (they presuppose cascade effects and a `{pre_cascade_sha}` anchor). |
| **D-J5** | **Disclosed cost:** Regime B leaves `017-S` occupying the P-001 single-active slot until an operator acts. This is stated, not engineered around — every available workaround is manifest-breaking or an unapproved destructive act. |
| **D-J6** | The regime-deciding question is exactly one: **is `017-S` live-status `active`?** At S-3 specifically, re-read the live status to decide, and run S-5's non-drift check when `active` to characterize a partial claim. |
| **D-J7** | **S-18 pre-mutation refusals** (`RECONCILE_FAIL_CASCADE_UNBOUND`, `_CLASSIFICATION_INVALID`, `_CLASSIFICATION_DRIFT`, `_CLASSIFICATION_REFUSED`) mutate nothing, so **R-1…R-8 are not invoked**; disposition is Regime B. `PA-017-CASCADE` remains unexercised. |
| **D-J8** | R-1…R-8 trigger **only** on a post-invocation S-18 failure, an S-20 failure, out-of-manifest archival, or a postcheck failure. Every halt at PF-1…S-17 is governed by the regimes instead. |
| **D-J9** | **Violation telemetry:** every halt corresponding to a named policy emits a P-005 event recording the policy ID, failure token, step ID and affected artifact IDs; the closure record carries it. |
| **D-J10** | **P-021 disposition:** findings raised during S-8 implementation or the S-10 review cycle are classified C1 (in-scope, fix now), C2 (out-of-scope ⇒ capture as a deferred stash entry with source refs), or C3 (defect in this plan ⇒ halt to Stage). Scope is never widened in place. |

## K. Rollback (post-cascade only)

| ID | Decision |
|---|---|
| **D-K1** | Restoration is achieved **only by adding commits** — never by working-tree overwrite of backlog records and never by file deletion. This ban is scoped to **R-1…R-8 only**. |
| **D-K2** | The installed `_ship.agent.md` Step 6.1(b)/(c)/(d) P-007 remedy (`git restore .backlogit/archive/`) **remains in force, unmodified**, and takes precedence where it applies. This plan has no authority to override an installed policy. |
| **D-K3** | **R-1 Quarantine:** stop, mutate nothing further, do not commit the cascade result, record the failure token verbatim. `{post_paths}` is already captured at S-19. |
| **D-K4** | **R-2 Classify:** (a) uncommitted ⇒ R-3; (b) committed, closure branch unmerged ⇒ R-4; (c) committed and closure merge landed ⇒ R-5; (d) committed with postchecks failed or unrun ⇒ **out of contract: halt to operator after R-1, no automatic revert**. |
| **D-K5** | **R-3 Uncommitted:** use the prerequisite §10.3.2 mechanism **on the closure branch**: `git add -A -- .backlogit/queue/ .backlogit/archive/`, commit as `{quarantine_sha}`, then `git revert --no-edit {quarantine_sha}` on that same branch. A `quarantine/017-S-<utc>` name may be a **ref** pointing at `{quarantine_sha}` for evidence — never a separate commit target. |
| **D-K6** | **R-4 Committed, unmerged:** target `{cascade_commit_sha}`; plain `git revert --no-edit`. A single parent is **expected** here and is not a P-009 signal. |
| **D-K7** | **R-5 Committed, merge landed:** target `{closure_merge_sha}`; verify two parents with parent 1 = mainline, then `git revert -m 1`. Any other shape ⇒ P-009 violation, halt, no automatic revert. |
| **D-K8** | **`{merge_sha}` is NEVER a rollback target.** It is the S-10 implementation merge; reverting it would unwind `018.008-T`'s implementation rather than the cascade. |
| **D-K9** | **R-6 Tree equivalence:** recompute the inventory and require it set-equal, path-for-path and blob-hash-for-blob-hash, to `{pre_paths}`. |
| **D-K10** | **R-7 Record equivalence:** tree equality is not sufficient. For all 13 members plus `017-S` plus `019.007-T`, re-read raw frontmatter and require `status`, `archived_status`, `archived_from`, `parent_id`, `artifact_type`, `commit` to match their S-15 values. Then `backlogit sync` and require `doctor` clean. |
| **D-K11** | **R-8 Verify-or-escalate:** a restore is complete only when **both** R-6 and R-7 pass. Otherwise halt to the operator with the quarantine ref, `{pre_cascade_sha}` and both inventories. **No second automated unwind.** |
| **D-K12** | Post-restore the shipment returns to Stage; `PA-017-CASCADE` is not re-exercised in the same session. |

## L. `PA-017-CASCADE` safeguards

| ID | Decision |
|---|---|
| **D-L1** | `PA-017-CASCADE` is a recorded operator approval held in escrow (2026-09-13T11:35:43-07:00; `change_kind` DESTRUCTIVE; approved). It does not expire and is **not re-litigated**. |
| **D-L2** | **It is NOT a claim authorization.** Claiming `017-S` is authorized by the shipment's own eligibility. |
| **D-L3** | Its scope is the **cascade close only**. It does not cover abandonment, manifest mutation, or any widening. |
| **D-L4** | **All nine conditions are re-measured by Ship at invocation; none is inherited.** Any mismatch **invalidates the approval** and halts. |
| **D-L5** | The nine conditions: (1) released by a review-passed Stage closure plan present on `origin/main` supplying execution site, gates, baseline and rollback — **verified at PF-2**; (2) `classify-close-path` returns `CASCADE` before any close call — S-17; (3) `safe-close` revalidates the binding before any mutation — S-18; (4) `021-S` shipped — PF-9; (5) parent-completion capability live — PF-10; (6) manifest unchanged at 13 — PF-4; (7) pre-cascade baseline and **pre-invocation** inventory exist and are verified — S-15 (`{pre_cascade_sha}` + `{pre_paths}` only; `{post_paths}` cannot be a precondition of the call that produces it); (8) linked-deliberation set for this manifest verified EMPTY by this plan's own measurement — PF-6; (9) no scope expansion — S-8 scope lock, PF-7 manifest invariant, out-of-scope list. |
| **D-L6** | Escrow release is **conditional, not self-certified**: the release condition is satisfied only when this plan is review-PASSed and present on `origin/main`, as verified at PF-2. |

## M. Out of scope

| ID | Decision |
|---|---|
| **D-M1** | Claiming `017-S` (Stage does not claim), implementing `018.008-T` (Ship work at S-8), and merging either PR without explicit operator authorization. |
| **D-M2** | Widening or re-litigating `PA-017-CASCADE`. |
| **D-M3** | Amending P-015, P-010, `_ship.agent.md` or `shipment-reconcile/SKILL.md`. |
| **D-M4** | Adding any backlog item to `018-F` or `017-S` — forbidden by the manifest invariant. |
| **D-M5** | Harmonizing the installed Ship CASCADE-branch prose — captured as stash **`11B75632`** (D-D3). |
| **D-M6** | Decision §7 follow-ups 1, 2, 3 and 5 — recorded, off-route, explicitly not current blockers. Item 4 alone is pulled in, as PF-7. Item 5 is referenced at S-2 as a **disclosure only**, with no derived halt condition. |
| **D-M7** | Re-opening prerequisite attempt 9. Revision 13 stays FROZEN. |

## N. Protected invariants

| ID | Invariant |
|---|---|
| **I-1** | Manifest is exactly 13 members from PF-4 through S-18. |
| **I-2** | The §3.3 allowlist is exactly 11 IDs; no archived descendant is off-list. |
| **I-3** | `018-F` is `queued` at S-2, `active` at S-14 entry, `done` only via S-14/S4. No agent performs a `queued → active` write on it. |
| **I-4** | No mutation precedes `CLOSE_PATH_VERDICT` being read. |
| **I-5** | `{pre_cascade_sha}` exists with a verified non-empty staged diff before any cascade. |
| **I-6** | Nothing outside the manifest is archived or deleted. |
| **I-7** | Prerequisite revision 13 stays FROZEN; attempt 8 FAIL stands. |
| **I-8** | `017-S` is never un-claimed, returned, moved, or abandoned. |

---

**Contract status:** canonical and internally consistent as of 2026-09-16. The governing plan is
remastered from this contract. Any divergence between the two is a plan defect, not a contract
defect.

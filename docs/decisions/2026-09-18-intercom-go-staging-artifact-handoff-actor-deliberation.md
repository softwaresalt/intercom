# Staging artifact handoff actor and push authority (B2) — Deliberation

- **Date**: 2026-09-18
- **Agent**: Stage
- **Scope**: `027-F` — Staging artifact handoff actor and push authority (B2)
- **Session branch**: `chore/stage-staging-artifact-handoff-actor-and-push-authorit`
- **Base**: `origin/main` @ `47b7bd327c20d83e784565ba00f100cb1cf61449` (PR #62)
- **Source refs**: `docs/decisions/2026-09-18-intercom-go-shipment-scoped-harness-contract-amendment-deliberation.md` (F-5), `docs/decisions/2026-09-18-intercom-go-harness-execution-model-decision.md`
- **Status**: DECIDED — Option 1 adopted

---

## 1. Problem frame

The installed Orchestrator Staging Artifact Merge Gate
(`.github/agents/_orchestrator.agent.md` Step 1.5, lines 270–297) carries two measured
defects, both localized to **item 3**:

**(a) Item 3 names no actor.** Items 3(a)–(d) direct *someone* to commit backlog files to a
staging branch, push it, open a PR to `main`, wait for merge, and pull. Every verb is
subjectless. Because the text lives in the Orchestrator's own agent file, the default reading
binds the Orchestrator — but P-010 states the Orchestrator "must not perform Stage or Ship work
directly — it routes to them as subagents." Committing backlog artifacts is enumerated Stage
work (`Stage MAY: Commit backlog and planning artifacts...`); pushing and opening a PR is
enumerated Ship work (`Ship MAY: Create and checkout feature/chore branches, commit, push` /
`Create, update, and merge pull requests`). So item 3 as written instructs the Orchestrator to
perform work that P-010 forbids it, and no other role is bound to do it instead. The gate has no
executor.

**(b) Item 3(e) directs a direct push to the default branch.** Line 287: *"Attempt a direct push
to `main` first. If the push is rejected... fall back to creating a staging PR."* This conflicts
with:

- **P-009 (Merge-Commit-Only)** — a direct push produces *no merge commit at all* and bypasses
  the PR merge path entirely, so the merge-commit guarantee becomes unenforceable rather than
  merely violated.
- **P-010** — no role in the boundary table holds direct-push-to-`main` authority. Ship is
  explicitly forbidden (`Ship MUST NOT: Commit or push directly to main`); Stage is forbidden
  from pushing at all; the Orchestrator may perform neither agent's work.

Item 3(e) is additionally **redundant**: its fallback body restates 3(a)–(d) verbatim. The only
thing it adds over the already-specified PR path is the forbidden direct-push attempt.

### 1.1 The measured surface (verbatim, lines 282–291)

```
3. When staging artifacts are uncommitted or unpushed:
   a. Commit any uncommitted backlog files to a staging branch: `chore/stage-{shipment_id}`
   b. Push the staging branch and create a PR to `main`
   c. Wait for the staging PR to merge (operator approval required)
   d. After merge, pull `main` and proceed to step 4
   e. **Branch protection handling**: Attempt a direct push to `main` first. If the push is
      rejected (exit code non-zero, typically due to branch protection rules), fall back to
      creating a staging PR: ...
```

---

## 2. Decisive finding: 027-F breaks a real dependency cycle (F-5 confirmed, and extended)

Prior deliberation recorded finding F-5: B2 is *separable and not universally circular*, because
the Step 1.5 gate already works in practice. **This session confirms F-5 by execution and
extends it with a materially stronger finding.**

### 2.1 F-5 confirmed by execution (non-circularity proof)

The claim "Stage artifacts cannot reach `origin/main`" is false. Measured this session:

| Check | Command | Result |
|---|---|---|
| 027-F's own record on remote default branch | `git cat-file -e origin/main:.backlogit/queue/027-F.md` | exit `0` — **present** |
| Backlog queue artifacts on `origin/main` | `git ls-tree --name-only origin/main .backlogit/queue/` | 19 entries |
| Route used | PR #62 merged at `47b7bd3` from `chore/stage-shipment-scoped-harness-contract-amendment` | merge commit |

The `chore/stage-*` → operator push → staging PR → merge-commit route **demonstrably lands Stage
artifacts on `origin/main`**, and has done so for PRs #59, #60, #61 and #62. A shipment manifest
created by this session reaches `origin/main` by the identical route. **027-F is deliverable
through the existing working gate. No circularity.**

### 2.2 NEW FINDING (F-6): the scope of 027-F is currently trapped inside a dependency cycle

`019.004-T` ("Correct Orchestrator Step 1.5 authorization and branch discovery") already owns
027-F's exact two defects as acceptance criteria:

- **AC-2**: "Step 1.5 names the actor authorized to commit, push and open the staging PR, cites
  the authority it already holds, and records the operator as the always-available substitute...
  No new Orchestrator authority is created by this task." → **defect (a)**
- **AC-3**: "The step-3 contradiction is resolved in favour of the branch-and-PR path; no
  instruction to attempt a direct push to the default branch remains." → **defect (b)**

Same file, same section, same lines. But `019.004-T` is **blocked**, and tracing its edges yields
a genuine cycle:

```
019.004-T  --depends on-->  025.001-T        (recorded edge)
025.001-T  --blocked by-->  B2               (recorded on 025-F / 025.001-T)
025-F      --depends on-->  027-F  (= B2)    (recorded edge)
B2's fix   --lives in----->  019.004-T AC-2/AC-3
```

So: **the fix for B2 is currently held inside a task that is blocked on B2.** That is the
circularity — not the artifact-delivery route (which works), but the *scope allocation*.
Leaving AC-2/AC-3 inside `019.004-T` makes B2 permanently unreachable.

**Carving AC-2/AC-3 out of `019.004-T` into `027-F` is what breaks the cycle.** This is not a
convenience split; it is the necessary and sufficient structural correction.

### 2.3 Corollary: leaving the scope duplicated would inject the C3 defect into 019.004-T

If `027-F` delivered AC-2/AC-3 while `019.004-T` retained them, then when `019.004-T` later ran,
its AC-2/AC-3 assertions would be **green-on-arrival** — precisely the
"an 'unchanged' assertion is green-on-arrival and can never satisfy P-004 red-phase" defect (C3)
that currently blocks `026-F`. Amending `019.004-T` is therefore **required for correctness**,
not optional hygiene.

---

## 3. Acceptance surface — admissible without answering O-1

O-1 (blocking `026-F`) asks whether a structural contract test is an admissible **IV-5**
acceptance surface *when no executable engine exists*. **O-1 does not gate 027-F**, on three
independent grounds:

1. **An engine exists and the precedent is already shipped.** `tests/integration/stage_branch_gate_test.go`
   is an installed, merged (018.008-T, shipment 017-S, PR #58) **characterization harness** that
   reads `.github/agents/_stage.agent.md` and `.github/policies/workflow-policies.md` as files and
   asserts AC-1..AC-12 over their text. The engine is `go test`. 027-F's harness is the *same
   shape over a sibling file*.

2. **The deliverable IS text, so the assertion is a direct measurement, not supplemental
   evidence.** `ship_feature_completion_contract_test.go` records that a static needle test is
   "SUPPLEMENTAL evidence only... cannot by itself satisfy IV-1..IV-5" — because there the
   deliverable was *behaviour* and the needle only proved the prose said the right thing. 027-F's
   deliverable is the contract text itself. There is no behaviour behind the text to under-measure;
   the text is the artifact. The gap that makes a needle test supplemental does not exist here.

3. **The assertions are genuinely red-on-arrival, so P-004 is satisfiable.** Unlike `026-F`'s
   green-on-arrival problem (C3), the property "no direct-push instruction remains in Step 1.5" is
   **false today** — the literal string *"Attempt a direct push to `main` first"* is present at
   line 287. `assertAbsent` fails now and passes after the edit. Genuine red → green transition.

**Fence**: 027-F asserts only over `.github/agents/_orchestrator.agent.md`. It does not claim, and
must not be read to claim, any resolution of O-1 for `026-F`'s behavioural IV-5 surface.

---

## 4. Options considered

### Option 1 — Carve AC-2/AC-3 into 027-F; rewrite item 3 authority-neutrally; REMOVE 3(e) — **ADOPTED**

Assign explicit actors, reduce the Orchestrator to observation/routing, delete the direct-push
arm, and amend `019.004-T` to drop AC-2/AC-3 and depend on `027-F`.

- Breaks the F-6 cycle (§2.2). Avoids the C3 injection (§2.3).
- Grants **no new authority to any role** — it only *names* authority each role already holds.
- Deliverable now, through the proven gate (§2.1).
- Cost: requires amending a blocked task in a blocked root (`019.004-T`), which is within Stage's
  backlog authority and is explicitly contemplated by 019-F's own reconstitution procedure.

### Option 2 — Close 027-F as a duplicate of 019.004-T; no shipment

- **Rejected.** This is the cycle in §2.2 restated as a decision: B2's fix would wait on
  `025.001-T`, which waits on B2. The live P-009/P-010 defect would remain installed indefinitely.

### Option 3 — Deletion-only: remove 3(e), leave actor naming to 019.004-T

- **Rejected as incomplete.** The two defects are coupled. Items 3(a)–(d) are *themselves* the
  P-010 hole — 019-F's own text records that item 3's "commit arm contradicts the Orchestrator's
  own installed orchestration-only Role statement." Deleting only (e) leaves a subjectless commit
  arm instructing the Orchestrator to do Stage work. It also leaves half the cycle intact.

### Option 4 — Fence 3(e) behind a P-009 conditional rather than removing it

- **Rejected.** Leaves the direct-push instruction textually present, so the structural assertion
  "no direct-push instruction remains" cannot pass, and a partial read or regeneration can
  re-activate it. 019.004-T's AC-3 wording ("no instruction... remains") already adjudicated
  removal over fencing.

---

## 5. Decision — the authority-neutral handoff

**Actors, each exercising only authority they already hold under P-010:**

| Sub-step | Actor | Authority basis |
|---|---|---|
| Commit backlog/planning artifacts to the Stage artifact branch | **Stage**, in-session, before handback | `Stage MAY: Commit backlog and planning artifacts` on its dedicated branch |
| Push the Stage artifact branch; open the staging PR; approve/merge it | **Operator** | Installed Stage Role Boundary (**PR row**): "The operator — never Stage — pushes the Stage artifact branch and opens/approves the resulting merge-commit staging PR; no Orchestrator authority to do so is granted or assumed by this release unit" |
| Detect state, emit the operator action request, wait, verify on `origin/main` | **Orchestrator** | Routing + verification only — already its declared role; no mutation |

**Explicitly preserved / explicitly granted-nothing:**

- **No Orchestrator authority is created.** The Orchestrator's role shrinks to detect → request →
  wait → verify. It never commits, pushes, opens or merges anything.
- **Stage/Ship boundaries unchanged.** Stage still never pushes or opens PRs. Ship is untouched by
  this change and still branches from `origin/main`.
- **Single implementation branch/worktree** (P-016) — unchanged; this adds no branch or worktree.
- **Merge-commit-only** (P-009) — strengthened: the PR path becomes the *only* route.
- **Stage artifacts reach `origin/main` before Ship** — preserved and made executable: the gate's
  item 4 verification on `origin/main` is retained unchanged as the terminal.

**Item 3's commit arm is removed, not reassigned.** Because Stage commits in-session, artifacts
are never uncommitted when Step 1.5 runs. If they nonetheless are, the Orchestrator **HALTS** with
a named token and routes back to Stage/operator — it must not commit them itself. This preserves
the fail-closed property.

**Item 3(e) is deleted outright**, replaced by an explicit statement that a direct push to the
default branch is forbidden for all pipeline agents under P-009/P-010, with the PR path as the
sole route.

---

## 6. Scope fences (NON-NEGOTIABLE)

To avoid duplicating adjacent blocked roots, `027-F` **MUST NOT**:

1. **Define the Stage handback record** — its fields, paths, or derivation. That is `025.001-T`'s
   exclusive scope. 027-F refers to the Stage artifact branch abstractly and never names
   `stage_branch`, `stage_base_commit`, `stage_head_commit`, `stage_artifact_paths` or
   `stage_outcome` as a consumed contract.
2. **Specify branch discovery mechanics** — discovering the branch *by name from the handback
   record* is `019.004-T` AC-1/AC-4. 027-F removes the commit arm (which is why the stale
   `chore/stage-{shipment_id}` literal disappears from it) without specifying the replacement
   discovery mechanism.
3. **Specify post-merge `stage_artifact_paths` verification** — `019.004-T` AC-5. 027-F leaves
   Step 1.5 item 4 (manifest existence + `STAGING_GATE_FAIL`) **byte-for-byte unchanged**.
4. **Build a generalized direct-push detector, fixture corpus, or CI wiring** — that is `020-F`
   (Strand B). 027-F's harness is a single-file characterization assertion in the already-shipped
   `stage_branch_gate_test.go` shape.
5. **Answer O-1, or touch `026-F` / `025-F` / any other blocked root's scope.**

---

## 7. Disclosed residual risks

- **R-1 — bounded overlap with 020-F.** 027-F's structural assertion covers text that 020-F's
  future generalized detector would also cover. Disclosed, not resolved. This is the *same*
  relationship the already-shipped `stage_branch_gate_test.go` has with 020-F, so the precedent
  covers it; 020-F's detector supersedes the narrow assertion when it lands.
- **R-2 — DB12DA37 remains live.** A shipment claim auto-transitions descendant tasks
  `queued → active` before Ship Step 2 runs, and the Stage-shipment substitution at
  `_ship.agent.md:362–367` drops the `harness-ready` predicate. Ship Step 3's derived executable
  set does include `active` members, so the ready queue is **not** vacuous for a Stage-prepared
  shipment — but a task could still be executed without a harness. **Mitigation (operator
  action, disclosed not authorized-around): harness-architect MUST run over `027.001-T` and the
  `harness-ready` label MUST be applied BEFORE the shipment is claimed.** B1 is refuted
  (`_ship.agent.md` provides a legal in-PR slot), so the harness has somewhere to stand.
- **R-3 — 019-F prose staleness.** 019-F's body describes 019.004-T as "where any Orchestrator
  authority would be DELIBERATED, SCOPED AND REVIEWED". After the carve, the *authority* question
  is settled here in 027-F; 019.004-T retains the *branch-discovery* consumption. A pointer note
  is added to 019.004-T. Full 019-F prose reconciliation is deferred (not required for correctness).
- **R-4 — regeneration risk.** `_orchestrator.agent.md` is autoharness-generated; a regeneration
  can revert the fix. This is the accepted interim residual that 020-F exists to remove.

---

## 8. Open questions

None blocking. O-1 remains open but is **not** a gate on 027-F (§3).

---

## 9. Outcome

**Option 1 adopted.** Proceed to impl-plan → plan-harden (REQUIRED — policy-surface change with
P-009/P-010 blast radius) → plan-review. Form a shipment only on an acceptable verdict.

---

## 10. AMENDMENT 1 — post-review correction (2026-09-18, plan-review attempt 1 = FAIL)

Two independent adversarial reviewers (policy persona, scope persona) returned **FAIL**. The scope
persona refuted this deliberation's central structural claim. **The refutation is correct and is
accepted. Finding F-6 is WITHDRAWN.** Re-measured, edge by edge:

| Claimed hop (§2.2) | Measured (`backlogit dep list`) | Verdict |
|---|---|---|
| `019.004-T → 025.001-T` | `019.004-T → 025.001-T (blocks)` | **REAL edge** |
| `025.001-T → B2` | `025.001-T → 018.008-T (blocks)` — the entire list | **NOT an edge** (prose only) |
| `025-F → 027-F` | `025-F → 026-F`, `025-F → 027-F` | **REAL**, but §2.2 omitted `026-F` |
| `B2's fix → 019.004-T` | `019.004-T` reverse-edge set is **EMPTY** | **NOT an edge** (semantic assertion) |

`027-F` has **zero** forward dependencies and never routed through `025.001-T`. **The recorded
graph is an acyclic DAG. There was no cycle, and the carve-out breaks none.** §2.2's promotion of
a scope-allocation problem into a graph cycle was an overstatement, and using it as the *sole*
grounds to reject Option 2 was an error.

### 10.1 The actual, corrected rationale for the split (replaces F-6)

The split is a **sequencing/independence** decision, not a cycle repair:

1. `019.004-T` is blocked by a **real recorded edge** on `025.001-T`.
2. `025.001-T` carries five recorded blockers **B1–B5** and failed adversarial plan-review on
   **two consecutive attempts**. B1 and B2 are now refuted — but **B3, B4 and B5 remain live**:
   B3 (fallback authorization scoped to the wrong policy), B4 (closure procedure not executable —
   unbound cascade, missing `returned_ids == []` assertion, undefined `allowed_ids`), B5
   (unsatisfiable edit placement). These are independent execution-surface defects.
3. `027-F`'s two defects require **none** of that machinery — no handback record, no closure
   procedure, no cascade classifier.

So the split delivers a live P-010/P-009 defect in installed text **now**, instead of coupling it
to a twice-failed unit that remains blocked on three unrelated defects. That is a real and
sufficient rationale; it is simply not the one originally given.

### 10.2 Rebuttal — scope finding 5 (fold back into `019.004-T`) is REJECTED on evidence

The scope persona proposed that correcting "stale blocker prose" on `025-F`/`025.001-T` would
unblock `025.001-T` → `019.004-T`, permitting one coherent edit. **This is rejected**: the
proposal addresses only B1 and B2 and is silent on **B3, B4 and B5**, which are recorded on
`025-F` as P0/P1 execution-surface defects and which drove two consecutive plan-review FAILs.
Refuting B1/B2 does not unblock `025.001-T`. The fold-back is therefore not the cheaper path it
appears to be. (Accepted from the same finding: the split's *stated* rationale was wrong — fixed
in §10.1.)

### 10.3 `027-F` does NOT unblock `025-F` (scope finding 2 — ACCEPTED)

`025-F.dependencies = [026-F, 027-F]`. `026-F` is blocked on operator determination **O-1**, which
is out of scope for this session. **Completing `027-F` therefore does not unblock `025-F`**; the
chain `027-F → 025-F → 025.001-T → 019.004-T` remains severed at `026-F` regardless. §2.2's and
§7's framing overstated the downstream value. Corrected: `027-F`'s value is **self-contained** —
it removes a live policy-violating instruction from installed text. No downstream unblock is
claimed.

### 10.4 Carve-out extended to AC-4 (scope finding 3 — ACCEPTED, was blocking)

The literal `chore/stage-{shipment_id}` occurs **only** at lines 283 and 288, **both inside item
3**, and the replacement deletes both. `019.004-T` AC-4 ("the branch name used is the discovered
`stage_branch`, **not** `chore/stage-{shipment_id}`") would therefore go partially
green-on-arrival — the exact C3 defect §2.3 calls fatal. The carve-out is extended to **AC-4 and
body scope item (4)**. Correspondingly, §2.1's and §6's assignment of AC-4 to items 1–2 was a
factual error (scope finding 4) and is corrected: AC-4 lives in item 3.

### 10.5 Authority corrections (policy persona — ACCEPTED)

- **The operator-authority quote is from the `_stage.agent.md` PR row, not the Git row.** Fixed
  in §5.
- **The direct-push prohibition is grounded in P-010, not P-009.** P-010 states without
  qualification: `Ship MUST NOT: … Commit or push directly to main`. P-009 governs *pull request
  merges* by Ship at Ship Step 5 and is silent on direct pushes; it supplies history-integrity
  *rationale*, not the prohibition. §1(b) and the replacement text are reordered accordingly.
- **The clause "Ship's push/PR authority does not extend to a Stage artifact branch it did not
  create and is not claiming" is DELETED.** `chore/stage-*` is a chore branch and P-010 grants
  Ship `Create and checkout feature/chore branches, commit, push` without qualification. Asserting
  the narrowing in `_orchestrator.agent.md` while `workflow-policies.md` says otherwise would put
  two installed contract surfaces in direct disagreement. Invariant **I1** ("no role gains
  authority") is only true once this clause is removed. 3(e)'s blanket prohibition already does
  the necessary work.

### 10.6 Net effect on the decision

**Option 1 remains adopted**, on the corrected §10.1 rationale. No re-deliberation required — the
chosen direction, the actor assignment, and the non-circularity proof (§2.1, verified by
execution) all survive. What changed is the *justification* for splitting, the *scope* of the
carve-out (now including AC-4), the *authority grounding* (P-010 over P-009), and the *claimed
downstream value* (none). The plan is revised accordingly and re-submitted for review.

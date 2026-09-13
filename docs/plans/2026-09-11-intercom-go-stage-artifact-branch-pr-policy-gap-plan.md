---
title: "Implementation Plan — Stage Artifact Branch/PR Policy Gap Correction"
date: 2026-09-11
revision: 19
status: awaiting-review
agent: Stage
governs: stash 638A410B
source: docs/decisions/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-deliberation.md
re_plan_decision: docs/decisions/2026-09-12-intercom-go-stage-policy-gap-re-plan-decision.md
split_decision: docs/decisions/2026-09-12-intercom-go-deferred-split-and-readiness-lock-decision.md
deferred_plan_strand_a: docs/plans/2026-09-12-intercom-go-stage-persistence-enforcement-plan.md
deferred_plan_strand_b: docs/plans/2026-09-12-intercom-go-stage-contract-enforcement-platform-plan.md
---

# Implementation Plan — Stage Artifact Branch/PR Policy Gap Correction

> **REVISION 19 (2026-09-12) — §R16 BELOW IS THE WHOLE GOVERNING CONTRACT.**
> Everything from §1 onward is **HISTORICAL AUDIT APPENDIX** describing the retired
> eight-task architecture (revisions 1–15). It is **preserved for audit and deliberately
> not deleted**, but it **no longer governs execution**. Where §R16 and any later section
> disagree, **§R16 wins without exception**. Revisions 17, 18 and 19 rewrote §R16 in place
> rather than appending further revision sections; the rev-16 text is in git at `14d64e4`,
> the rev-17 text at `9e44bfe` and the rev-18 text at `76c372c`.
>
> **Rev 19 changes exactly one thing**: it adds §R16.3.1a, which proves the withdrawn
> 13-entry `CASCADE` shape is **unreachable by construction** rather than merely unreached.
> The governing close path, the manifest, the executable set and every other rev-18
> conclusion are **unchanged**. No backlog artifact was mutated in rev 19.
>
> **Requires plan hardening**: **yes — applied in rev 18, extended in rev 19** (see §R16.10).

## §R16 — Current Contract (authoritative)

### R16.1 Root cause of the reduction

**P-004** requires *every* generated test function to be RED before implementation begins.
**Ship Step 4.3** requires the *full* `go test ./...` to be GREEN after *every* task.
**Ship Step 2 item 1** harnesses every `queued` task of the **covering feature** in one batch.

For any covering feature with more than one queued test-bearing task these are **jointly
unsatisfiable**: task 1 going green leaves the rest red, Step 4.3 fails, `build-feature` burns
its attempts on unrelated future work, and **no task can complete**. Revisions 10–15 tried to
bridge this with surface predicates, a harness-state manifest, a terminal witness, a `-gatetask`
selector and an MP0–MP6 mutation-proof series, and failed adversarial review every time.

**Revision 16 removed the cause instead of bridging it**: exactly one queued task ⇒ one red
function ⇒ one red→green transition ⇒ full suite green immediately. No activation manifest,
terminal witness, selector flag, surface predicate or MP-series check is required or permitted
in this release unit.

### R16.2 The reduced release unit

> **REV 19 UPDATE (2026-09-13) — the manifest row below is CORRECTED.** Governing
> decision: `docs/decisions/2026-09-13-intercom-go-a-only-simplification-decision.md`
> (`SIMPLIFICATION_VALID`). `017-S` is **no longer task-only**. Its root `018-F` has been
> restored to the manifest, giving a **13-entry fully-covered root**, and it depends on
> **`021-S` only**. It closes by the **existing** P-015 `CASCADE` exception once shipment
> A lands Ship Step 6.1(a1). The `SAFE_CLOSE` step-8 exit-9 conflict is **off this route**.

| Item | Value |
|---|---|
| Feature | `018-F` (root; no `parent_id`; **IS** a shipment member — rev 19) |
| Shipment | `017-S` (**13-entry fully-covered root**: `018-F` + all 12 descendants; depends on `021-S`) |
| Executable tasks | **1** — `018.008-T` (size **M**, complexity **medium**) |
| Harness | ONE function `TestStageBranchGate_ContractCorrected`, ONE expected marker |

**Superseded row, retained for audit:** rev 18 recorded *"`018-F` … **not** a shipment
member"* and *"`017-S` … **task-only** 12-entry manifest"*. Both are **withdrawn**. The
membership change does **not** widen the executable set — Ship Step 3 filters the manifest
by `artifact_type`, so `018-F` is excluded from execution before any status is read, and
the derived executable set remains exactly `[018.008-T]`.

`018.008-T` carries exactly three surfaces: **(1)** P-010 in
`.github/policies/workflow-policies.md`; **(2)** the `.github/agents/_stage.agent.md` Role
Boundary Git and PR rows; **(3)** the Step 1.9 gate (§R16.4).

### R16.3 Shipment `017-S` manifest and P-015 closure (re-repaired rev 18)

> **The rev-17 13-entry `CASCADE` shape is WITHDRAWN.** It was closable in principle but
> **unreachable in practice**, because Ship Step 6 would halt *before* closure ever began.

The manifest is the **closure membership record**, never the executable set. It now contains
**12** entries and is **task-only**: `018.008-T` plus the 11 pre-archived legacy descendants
`018.001-T`, `018.002-T`, `018.003-T` and their eight subtasks. **`018-F` is deliberately not a
member.** `018-F` was removed with the registry-declared supported operation
`backlogit shipment return-blocked` (registry op `return_blocked`, MCP
`backlogit_return_blocked`), then restored to `queued` with `backlogit move 018-F --status queued`,
which also cleared the `blocked_reason` the removal stamped. No hand-edit of
`custom_fields.items` was required or performed.

#### R16.3.1 Why the rev-17 shape failed

Ship Step 6 item 1.a invokes `shipment-reconcile` with `mode: pre` and **`expected_status: done`**.
Pre-mode step 3 checks `.backlogit/queue/` **first** and, when the record is found there, compares
its declared `status` to `expected_status` with **no artifact-type exemption** (the task-artifact
filter applies only to step 5's record-status classification).

At Step 6, `018-F` is live in `.backlogit/queue/` with `status: active` — `backlogit shipment claim`
moves the covering feature `queued → active` — and **no Ship step ever moves a covering feature to
`done`**: Step 4.5 moves *tasks* only. `018-F` would therefore classify **`status-mismatch`**,
pre-mode would return `HALT — operator reconcile required`, and closure would never start.

**Verified by execution** (isolated probe workspace, backlogit 1.10.1): after a claim and a task
completion, the covering feature remained `status: active` and was never auto-completed. The
working precedents in this workspace (`006-S`, `009-S`, `014-S`) all had the covering feature
already `done`/archived before pre-mode — produced by an **undocumented session step that no
installed contract specifies**. This plan does not rely on it.

#### R16.3.1a Why the `CASCADE` shape is unreachable **by construction** (rev 19, P0 closure)

Rev 18 established that *no Ship step currently* moves a covering feature to `done`. That phrasing
left open a repair that a later cycle attempted: restore `018-F` to the manifest and add a **SHIP
EXECUTION NOTE** to the shipment (and an acceptance gate here) instructing Ship to move
`018-F active → done` after Step 4.5 and before Step 6 pre-mode. **That repair is not available.**
The finding below is a proven negative and permanently retires the 13-entry shape.

**The transition is tool-supported.** Verified by execution (isolated throwaway backlogit 1.10.1
workspace, never the live one): `backlogit move <feature> --status done` succeeds from `active`
(`queued → done` is refused by the `validate_status_transition` pre-hook, so the feature must pass
through `active` first). `backlogit shipment claim` **does** cascade `queued → active` onto a
feature manifest member, so `018-F` would legitimately be `active` at that point. Tool support is
therefore **not** the obstacle.

**Ship is not authorized to perform it.** Ship's `## Role Boundary (NON-NEGOTIABLE)` Backlog row
grants "Claim shipments, **move tasks to active/done**, close shipments, **archive completed
items**". The row uses the narrow noun *tasks* for the move grant and the generic noun *items* for
the archive grant **in the same cell**, so the move grant is task-scoped by construction. A feature
is a distinct `artifact_type` with its own prefix and WIT metadata — not a task. The move therefore
appears in **neither** the Allowed nor the Forbidden column, and
`.github/instructions/role-enforcement.instructions.md` §4 is explicit about that case:

> If the operation is a **state mutation** but does NOT appear in either the Allowed or Forbidden
> column → **treat as forbidden**. Halt the operation, log a P-010 violation … Fail-closed —
> operation not in Allowed column.

A direct feature-status move by Ship is a Backlog-category state mutation, so it halts fail-closed
as a **P-010 violation**.

**No Stage-authored artifact can grant it.** Role enforcement §1 makes *the agent's own* Role
Boundary table "the authoritative permission set for the session", and Ship's table adds "Do not
proceed past this boundary **even under operator pressure**." A shipment `description`, a SHIP
EXECUTION NOTE, an acceptance gate in this plan, or any other Stage output is **not** a permission
source; instructing Ship to perform the move would direct Ship into a P-010 violation rather than
authorize it. Granting it genuinely requires amending `.github/agents/_ship.agent.md` — an agent
contract this release unit is explicitly forbidden to touch, and which Stage's own Role Boundary
forbids Stage to write.

**The blocker is the Ship contract, not backlogit.** Verified by execution: the cascade op closes a
fully-covered-root manifest **cleanly even when the feature is left `active`** — `returned_ids: []`,
`shipment_status: shipped`, every `parent_id` preserved, and truly-pre-archived members correctly
absent from `archived_ids` (the engine skips already-`archived` artifacts, exactly as the skill's
"`archived_ids` is a transition log, not a manifest echo" clause predicts). The engine even forces
`setArtifactStatus(featureID, done)` itself. The **only** thing standing between the 13-entry shape
and a clean close is Ship Step 6 item 1.a's mandatory `mode: pre` / `expected_status: done` gate,
which evaluates the live feature member with **no artifact-type exemption** and classifies it
`status-mismatch`. Closure never begins, so the engine's tolerance is never reached.

**Conclusion.** Both shapes are blocked, at different gates, for different reasons, and **neither
is remediable from a Stage planning cycle**: the task-only shape halts at safe-close step 8 on a
tool refusal (§R16.3.3), and the fully-covered-root shape halts at Step 6 pre-mode on a role
boundary. The rev-18 disposition stands unchanged — `017-S` executes, reviews and merges normally,
then pauses at Ship Step 6 closure for operator disposition. Resolving either one requires a
separately deliberated and reviewed amendment to the Ship/reconcile contracts or to backlogit.

#### R16.3.2 The 12-entry task-only shape, proven end to end

| Gate | Input | Outcome |
|---|---|---|
| Step 0.5 item 1a (queued-with-active-work) | task-artifact members `018.008-T` (queued), `018.001/2/3-T` (archived) | no `active`/`done` task → **PASS** |
| Step 0.5 item 2 (membership) | task-only manifest | expressly accepted; covering feature resolved via `parent_id` (097-S precedent) → **PASS** |
| Step 0.5 item 3 (covering parent) | every member resolves to `018-F` / its tasks | **PASS** |
| Step 0.5 item 6 (intake pre-mode, `expected_status: queued`) | `018.008-T` `matched`; 11 legacy `pre-archived`; 0 orphans; record `queued` with no active/done task → `record-consistent` | **PROCEED** |
| Step 2 (harness) | queued tasks of `018-F` = `[018.008-T]` | exactly **1** generated function → P-004 satisfiable |
| Step 3 item 1 (executable set) | `artifact_type: task` filter, then positive status rule | KEEP `[018.008-T]`; `pre_archived_skipped` `[018.001-T, 018.002-T, 018.003-T]`; `already_done` `[]`; fail-closed halts **none**; the 8 subtasks are excluded at the `artifact_type` filter **before** any status read |
| Step 6 pre-mode (`expected_status: done`) | `018.008-T` relocated to `archive/` by Step 4.5's `move --status done` → `pre-archived`; 11 legacy → `pre-archived`; 0 orphans; record `active` → `record-consistent` | **PROCEED** — and `018-F` is **never evaluated** |

**Classification.** Safe-close Step 0(c) returns **`SAFE_CLOSE`**, deterministically. The P-015
exception is quantified over *every feature member of the manifest*; with **zero** feature members
there are zero qualifying roots, so precondition 4 ("the manifest contains nothing beyond the
qualifying root feature member(s) and their descendants … a task whose ancestry does not lead back
to one of the qualifying root features disqualifies the exception for the whole manifest") is
failed by every member. The skill's default rule reaches the same verdict independently.

**Protected set = `{018-F}` exactly.** Safe-close step 2 derives the covering feature from the
manifest hierarchy, finds it absent, and classifies `017-S` a partial-feature shipment. Every other
descendant of `018-F` at every depth — the three legacy tasks and the eight subtasks — **is** a
manifest member and contributes nothing. The eight subtasks are included precisely so that a
conservative reading of "every task sharing the covering feature's hierarchy prefix" cannot pull
them into the protected set.

**Baseline integrity gate (step 3) PASSES**: the sole protected-set member `018-F` is present in
`.backlogit/queue/`. This is exactly the gate that **halted** under the earlier two-entry manifest,
where the 11 archived non-members were protected and the protected set has **no** pre-archived
exemption.

**Archive loop (step 4) mutates nothing**: all 12 members are already in `.backlogit/archive/`, so
each classifies `pre-archived` and is skipped. The verify-after-each invariant (step 5) and the
final re-check (step 7) both confirm `{018-F}` intact. Verified by execution: neither
`backlogit shipment claim` nor a single-artifact archive mutates an archived manifest member
(byte-identical file hashes before and after), and neither touches a covering feature that is not
an explicit manifest member.

**Post-close state**: `017-S` archived; `018.008-T` + 11 legacy members archived; `018-F` **still
live in `.backlogit/queue/`**, protected, `queued`/`active` — which is the stated goal.

#### R16.3.3 Residual closure blocker — P0, pre-existing, shape-independent, **not** remediable here

Safe-close **step 8** closes the shipment record by
`backlogit move <shipment_id> --status shipped`. **backlogit 1.10.1 refuses that command for
shipment artifacts with exit code 9**: `shipment must be shipped via ShipShipment, not a direct
status update`. Verified by execution in this cycle from **both** `queued` and `active` record
states; already recorded in this workspace by the `009-S` safe-close report ("the generic
safe-close move path described in this skill's steps 1–10 is not available in `backlogit` 1.10.1
for shipment records") and by `docs/compound/2026-05-07-backlogit-shipment-status-constraints.md`.

A `SAFE_CLOSE`-classified shipment therefore halts fail-closed at step 8 with
`RECONCILE_FAIL_SHIPMENT_RECORD_LIVE_STATUS`. The contract's own negative scenarios forbid both
available substitutes: archiving an `active` record yields `archived_status: active` →
`RECONCILE_FAIL_SHIPMENT_RECORD_PROVENANCE`, and calling the cascade op after a `SAFE_CLOSE`
classification is a **P-005 process deviation**.

**No manifest shape avoids this.** The only shape that reaches the *permitted* cascade path is one
containing `018-F` — and that shape fails pre-mode first (§R16.3.1), **unreachably so** (§R16.3.1a:
the corrective feature transition is tool-supported but role-forbidden for Ship, and no Stage-authored
artifact can authorize it). The two failures are at different gates and are mutually exclusive; there
is no third shape.

**Disposition.** This is **out of scope for `018.008-T`** and is **not a defect introduced by this
release unit**. Resolving it requires either a backlogit change or a separately deliberated and
reviewed amendment to `shipment-reconcile` step 8 / P-015 — and this cycle is explicitly forbidden
from amending Ship/reconcile contracts. Recorded as a stash entry for Stage intake. Until it is
resolved, `017-S` **executes, reviews, and merges normally**, and then **pauses at Ship Step 6
closure for operator disposition**.

**Ship's executable set is exactly `[018.008-T]`.** The authoritative filter is
`artifact_type: task`, **never an ID suffix** — subtask IDs end `-ST`, whose final character is also
`T`, so a suffix test would admit all eight subtasks. With the `artifact_type` filter applied first,
the three archived tasks are `pre_archived_skipped` and the subtasks never enter the derivation.

### R16.4 Step 1.9 is an operative mandatory fail-closed gate

Step 1.9 is **not** prose that merely sits before Step 2. It has two mandatory parts:

**(a) Registration — exact anchors, never line ordinals.** One line is inserted into the single
fenced `text` block that follows the heading `## Step Sequence Contract (NON-NEGOTIABLE)`. The
insertion point is located by **two byte-exact anchor lines**, quoted here as they exist in the
installed file (both separators are **EM DASH U+2014**, bytes `E2 80 94`; note the **three** ASCII
spaces after `Step 2`):

```text
[ ] Step 1.8 — Learnings retrieval
[ ] Step 2   — Deliberation
```

The new line, inserted strictly between them (one ASCII space after `1.9`, EM DASH U+2014, one
ASCII space either side):

```text
[ ] Step 1.9 — Stage artifact branch gate (fail-closed)
```

*Normalized fallback locator.* If an exact byte match for either anchor is absent, locate — inside
that same fenced block — the **unique** line whose text, after (1) stripping a leading `[ ] ` or
`[x] ` marker, (2) replacing every run of EM DASH (U+2014), EN DASH (U+2013) or ASCII hyphen with a
single ASCII `-`, and (3) collapsing internal whitespace runs to one ASCII space, equals exactly
`Step 1.8 - Learnings retrieval` (anchor A) or `Step 2 - Deliberation` (anchor B).

*Fail on zero or multiple.* If **either** locator yields **zero** matches or **more than one**
match, halt with `STAGE_BRANCH_GATE_ANCHOR_FAIL` and record P-005. Never insert by line ordinal,
never guess a position, never insert into a second fenced block. Without this registration the gate
is not covered by the existing rule that blocks the summary until every applicable step completes,
so skipping it would not be a P-005 violation.

**(b) Section.** `Step 1.9: Stage Artifact Branch Gate (NON-NEGOTIABLE)`, strictly after Step 1.8
and strictly before Step 2, specifying (i)–(vii) **in this order**:

1. **Worktree topology precheck (P-016) — runs first, before any branch create/select/resume.**
   Run `git worktree list --porcelain` and classify **every** attached worktree into exactly one of:
   (1) the current worktree; (2) an **explicit, time-boxed Stage spike/research worktree** whose
   spike context is recorded for this session, per the P-016 paragraph beginning
   `**Allowed exception (Stage spike/research only)**`; (3) prohibited/ambiguous. Class (2)
   **passes** — the precheck must not narrow, qualify or supersede the P-016 exception. Any class-(3)
   worktree, and any worktree that cannot be positively classified into (1) or (2), halts with
   `STAGE_BRANCH_GATE_FAIL` plus a P-016/P-005 record, **before** any branch operation and **before**
   any tracked mutation.
2. **Deterministic slug derivation — exact source fields.** Feature-shaped intake: the `title` field
   of the Step 1 classified feature/epic/chore artifact, or the verbatim stash-entry title text used
   as the Step 1 classification subject when no artifact exists yet. Task-shaped intake: the
   **`Proposed covering feature title`** string of the Step 1.5 grouping the operator **selected**.
   Shipment-shaped (**optional, never required**): the `title` field of an already-known shipment
   record — usable on resumption, never required, because Step 5.5 creates the shipment as the
   session's **last** mutation.
   *Normalization, in this exact order*: (1) NFC-normalize; (2) lowercase by **ASCII case folding
   only**; (3) replace every maximal run of characters outside `[a-z0-9]` with a single `-`;
   (4) trim leading/trailing `-`; (5) **truncate to at most 48 bytes**, then trim any trailing `-`
   the truncation produced. The 48-byte bound keeps `chore/stage-` (12 bytes) + slug within the
   workspace's configured `max_slug_length: 60`.
   *Empty fallback*: if the result is the empty string, use the literal `stage-session`. The gate
   must **not** halt on an empty slug and must not invent a slug from any other source.
3. **Base-ref resolution — never assumed.** `git symbolic-ref --short refs/remotes/origin/HEAD`,
   stripping the leading `origin/`; on failure `git rev-parse --abbrev-ref origin/HEAD`, same strip.
   If **both** fail, halt with `STAGE_BRANCH_GATE_FAIL`. The gate must not assume the literal name
   `main`.
4. **Create-or-select/resume with an ownership / base-identity / collision discriminator.**
   (a) Branch absent (`git rev-parse --verify --quiet refs/heads/{expected}` non-zero): require a
   clean worktree (`git status --porcelain=v1` empty), then create and check out from
   `refs/remotes/origin/{base_ref}` — the ownership-establishing case, no discriminator.
   (b) Branch already checked out: select, **after** the discriminator.
   (c) Branch exists but not checked out: require a clean worktree, check out, resume, **after** the
   discriminator.
   *Discriminator (cases b and c only), both conditions required*: (1) **base identity** —
   `git merge-base refs/heads/{expected} refs/remotes/origin/{base_ref}` succeeds and returns a
   non-empty SHA; (2) **content ownership** —
   `git diff --name-only refs/remotes/origin/{base_ref}...refs/heads/{expected}` is empty or lists
   **only** paths under `.backlogit/`, `docs/plans/`, `docs/decisions/`, `docs/memory/`. A single
   path outside those roots proves the branch is not a Stage artifact branch. If either condition
   fails, the name match is a **collision, not a resumption**: halt with `STAGE_BRANCH_GATE_FAIL`,
   record P-005, perform no tracked mutation, and do not create, rename, delete, force-update or
   reuse the colliding branch. **Branch resume is permitted only when identity matches; otherwise
   the gate halts.**
5. **Symbolic HEAD equality** — `git symbolic-ref --short HEAD` must equal the expected branch; a
   detached HEAD (non-zero exit) fails.
6. **Default-branch inequality** — the expected branch must differ from the resolved `base_ref`.
   Retained even though the `chore/stage-` prefix makes equality structurally unreachable today, so
   a future prefix change cannot silently defeat the gate.
7. **Named fail-closed halt and categorical deferral rule** — `STAGE_BRANCH_GATE_FAIL` (or
   `STAGE_BRANCH_GATE_ANCHOR_FAIL` for the (a) locator), a P-005 record, and **no** tracked mutation.
   The gate precedes every tracked Stage artifact mutation of the session without exception; anything
   otherwise scheduled earlier is deferred until it passes. The enumeration is **categorical**; the
   two named write halves (late-identifier reconciliation, duplicate-scan archival) are
   **illustrative, not a closed list**.

**P-016 preservation (non-negotiable).** The P-016 paragraph beginning
`**Allowed exception (Stage spike/research only)**` must remain **byte-for-byte identical**. The
word "only" in the new commit grant constrains **which branch** Stage may commit on; it must not be
drafted so as to narrow, qualify, supersede or delete that exception. Item 1 above explicitly
**passes** the spike/research class for the same reason.

### R16.5 Interim persistence — operator-only after Stage handback

This unit delivers the **branch gate only**. It delivers **no** Stage commit step (`019.009-T`),
**no** handback record (`019.007-T`), **no** no-shipment terminal (`019.008-T`) and **no**
Orchestrator branch discovery (`019.004-T`). The honest consequence:

* **Stage branch gate.** Step 1.9 prevents any tracked Stage artifact mutation on the default
  branch. All bounded Stage mutations happen on `chore/stage-{scope-slug}`.
* **Stage terminus.** Stage ends its session at a **named fail-closed handback**,
  `STAGE_ARTIFACTS_UNCOMMITTED`, with its bounded artifacts **uncommitted** on that branch under the
  four Stage artifact roots (`.backlogit/`, `docs/plans/`, `docs/decisions/`, `docs/memory/`). They
  are uncommitted **because commit and handback automation is deferred** (`019.009-T` / `019.007-T`),
  not because leaving them uncommitted is desirable.
* **The Orchestrator MUST HALT here.** The Orchestrator **MUST NOT** execute Step 1.5 item 3 (its
  uncommitted/unpushed arm) for this interim route. That arm commits Stage artifacts, and its
  sub-item **3(e) instructs a direct push to `main` to be attempted first**. Both contradict the
  Orchestrator's own installed Role statement — *"You do NOT triage stash entries yourself. You do
  NOT write code or create PRs yourself."* — and a direct-`main` attempt is precisely the behaviour
  this release unit exists to prohibit.
* **Operator-only completion sequence.** **Only the operator** may: (1) verify the branch, worktree
  topology and the uncommitted artifact set; (2) commit **only** the four allowed Stage artifact
  roots on that branch; (3) push the side branch; (4) open and approve a **merge-commit** staging PR
  (P-009: no squash, no rebase); (5) update the default branch; (6) re-verify that the shipment
  manifest is present on `origin/main`. No agent performs any of these steps under this contract.
* **The current dark run pauses at that operator checkpoint.** Because the operator is AFK, current
  automation **may not proceed past it** until the staging-automation work in `019-F` ships. **This
  means the dark pipeline cannot finish autonomously after this release unit is implemented.** That
  is stated plainly rather than papered over.

**No Orchestrator authority is created, and none is claimed to pre-exist.** The rev-17 assertions
that Orchestrator Step 1.5 is "naturally fail-closed" for this route and that item 3 supplies
pre-existing authority to commit Stage artifacts are **factually wrong and are withdrawn**: item 3's
dirty-artifact arm contains a direct-`main` attempt and a commit action, both outside the
Orchestrator's orchestration-only boundary. This release unit grants the Orchestrator **nothing**,
amends **nothing** in P-010's Orchestrator statement, and routes the entire interim sequence to the
operator.

**Withdrawn overclaim.** The rev-16 assertion that this is "exactly the interim behaviour already
in use" is **false and withdrawn**: the behaviour in use before this unit permitted Stage to commit
on the default branch — the very defect being closed.

### R16.6 Harness footprint — P-004 compliant unconditionally (rev 18)

The production contract surfaces are **two Markdown files**:
`.github/policies/workflow-policies.md` and `.github/agents/_stage.agent.md`. **No Go production
code implements this contract**, so the harness is a characterization harness that reads those two
installed files directly. **No companion Go production stub is created**, and no new package, file
or exported function is added outside the single test file: a file-reading test compiles with no
stub, so harness-architect Step 4 item 3's "matching production stub" instruction — which exists so
the *module compiles* while a test calls not-yet-existing production code — has no object here.

**The expected marker is mandatory, not conditional.** P-004's precondition is
`go test ./...` exiting non-zero **with expected failure markers in the output for every test
function**, and harness-architect Step 5.2 requires **all** harness tests to fail with the expected
marker `panic("not implemented: <reason>")`. Both are unconditional. **The rev-17 "*fallback, only
if harness-architect insists*" framing was P-004-noncompliant and is WITHDRAWN.**

The harness therefore carries, at H0, **exactly one expected marker**: an **unexported helper
declared inside** `tests/integration/stage_branch_gate_test.go` with body
`panic("not implemented: stage branch gate contract")`, invoked **unconditionally** by
`TestStageBranchGate_ContractCorrected`. Never a new file, package, or anything under `internal/` or
`cmd/`. Its final state after the task is an ordinary test helper carrying the contract assertions,
with no panic.

**H0 evidence — explicit and captured, before any implementation:**

| # | Check | Required result |
|---|---|---|
| 1 | `go vet ./...` | exit 0 |
| 2 | `go test ./tests/integration/ -run TestStageBranchGate_ContractCorrected` | exit **non-zero** |
| 3 | captured output | contains the literal `not implemented: stage branch gate contract`, attributed to `TestStageBranchGate_ContractCorrected` |
| 4 | H0 red count over this unit's generated functions | literally **1 of 1** |

After the **sole** task completes: the marker is gone, the function passes, `go vet ./...` exits 0,
and the **full** `go test ./...` exits 0 — Ship Step 4.3 is green immediately because `018-F` has
exactly one queued task.

| Metric | Value |
|---|---|
| Files | **3** — 2 edited contract files + 1 generated test file |
| New production files | **0** |
| Functions | 1 generated test function + at most 2 unexported helpers in the same file, one carrying the H0 marker |
| Test scenarios | 1 scenario group |
| Skill domains | 1 (policy/contract text) |

**Within the 2-hour rule. `018.008-T` stays atomic and is not split.** Every edit is a named,
quoted, pre-located textual insertion or replacement in two files; no design work, no discovery, no
build-system or schema change.

**Locator precision.** The Ship mirror bullet is located by its exact Markdown text
``- Commit or push directly to `main` `` (with `main` in backticks), never by line ordinal. If an
exact match is absent, the documented normalized locator applies — the unique Markdown list item
under the `**Ship MUST NOT**:` heading whose text, after stripping backticks and collapsing internal
whitespace, equals `Commit or push directly to main` — and the executor **halts** if that normalized
match is also absent or non-unique.

**Scoped zero-skip claim.** `TestStageBranchGate_ContractCorrected` and every assertion inside it
execute unconditionally — no `t.Skip`/`t.Skipf`, no build tag, no env-var or platform gate, no
activation manifest, no selector flag. **This claim is scoped to this release unit's generated
harness assertions only.** Pre-existing environment-conditional skips elsewhere in the repository
(`pwsh`-availability guards in `tests/integration/build_script_test.go` and
`start_script_test.go`; symlink/junction privilege guards in
`tests/integration/output_path_guard_test.go` and `internal/pathsafe/*`) are out of scope and are
neither introduced nor modified. The rev-16 claim of "no skipped, pending, or
conditionally-activated assertions **anywhere in the suite**" was factually unsatisfiable —
14 existing test files already call `t.Skip` — and is **withdrawn**.

### R16.7 Deferred work — two feature strands, no live shipment (rev 18)

The former single deferred feature mixed **persistence semantics** with **reusable enforcement
infrastructure**. Revision 17 split it; revision 18 removed the invalid shipment representation.

| Strand | Feature | Shipment | Plan |
|---|---|---|---|
| A — Stage handback / staging-PR persistence / no-shipment semantics | `019-F` (blocked) | **none** | `docs/plans/2026-09-12-intercom-go-stage-persistence-enforcement-plan.md` |
| B — detector / fixture corpus / event parser / CI / regeneration proof | `020-F` (blocked) | **none** | `docs/plans/2026-09-12-intercom-go-stage-contract-enforcement-platform-plan.md` |
| Readiness decision (not a delivery strand) | `021-F` (blocked) | **none** | `021.001-T` decision artifact |

Deferred: handback emission · `stage_artifact_paths` derivation · no-shipment terminal
reachability · the Stage artifact commit step · out-of-root path safety · Orchestrator Step 1.5
branch-discovery and post-merge verification · the structure-aware direct-push detector · the
fixture corpus · `go test -json` event parsing · regeneration-resistant CI wiring · CODEOWNERS
and divergence-ledger updates · the MP-series mutation proofs · regeneration proof.

**A reviewer MUST NOT fail the current unit for the absence of any of the above.**

#### R16.7.1 Why the rev-17 readiness lock was invalid and was retired

The rev-17 lock made `018-S` and `019-S` `status: blocked` and gave each a `blocks` edge to a
lock-token shipment `020-S`. **`blocked` is not a valid backlogit shipment status.** The shipment
status enum is exactly `queued | active | shipped | abandoned` (`.backlogit/header-def.yaml`;
`docs/compound/2026-05-07-backlogit-shipment-status-constraints.md`). Verified by execution against
backlogit 1.10.1: `backlogit move <shipment> --status blocked` exits **0** — the generic mover
validates against the *generic* status enum, which is the known CLI defect that wrote those values
— while `backlogit move <shipment> --status shipped` exits **9**, and
`backlogit move <shipment> --status abandoned` from `queued` is rejected by
`validate_status_transition`. A lock token that can exist **only** in an invalid state is not
enforcement.

**There is no safe valid queued-shipment readiness lock in current backlogit semantics** unless a
genuine predecessor shipment is actually intended to ship: a `queued` shipment is by definition
claimable, and `blocked → queued` as an "unlock transition" is unsupported for shipments. Leaving
implementation-unready future shipments claimable is therefore unacceptable.

**This supersedes the earlier request for a queued future shipment.** The adversarial gate proved no
valid safe queued representation exists; **reliability and safety take precedence** over having a
shipment record present.

#### R16.7.2 What replaces it

`018-S`, `019-S` and `020-S` were **archived** with the official non-destructive
`backlogit archive` operation, after their `blocks` edges were removed. **Nothing was deleted**; the
archived records preserve their manifests, IDs and titles, and carry `archived_status: blocked` —
an honest provenance record of the defect-written value, and a value the `shipment-reconcile`
contract already enumerates and handles (`archived_status: active|queued|blocked|abandoned`).

Both deferred features and **all** their tasks remain `status: blocked` and belong to **no
shipment**. Orchestrator Step 2 and Ship Step 0.5 both operate on **shipments**: with no shipment
there is nothing to list, select, claim or mis-route. **It is acceptable — and safer — to retain
blocked future features and tasks without any live shipment until implementation readiness.**

Verified by execution that archival costs no future capability: members of an archived shipment are
freely re-assemblable into a new shipment, and `backlogit archive <shipment>` does not cascade to
its members.

**Stage creates fresh shipment(s) only after P-006 plan review** selects one of `021.001-T`'s three
admissible options: (1) single-task features/shipments; (2) proven one-queued-child serialization;
(3) a reviewed shipment-scoped harness contract amendment. **No deferred shipment is eligible or
claimable now, because none exists.**

### R16.8 Single-shipment boundary and dependency direction

`017-S` is the **only** shipment in the workspace and the **only** P-017 in-scope item; future work
is outside that scope. Dependency direction is **future → current** and is verified to contain **no
`current → future` edge and no cycle**: `019.007-T`, `019.009-T`, `019.004-T`, `020.001-T`,
`020.004-T` each depend on `018.008-T`; `018-F`, `018.008-T` and `017-S` depend on nothing. The two
readiness-lock edges (`018-S → 020-S`, `019-S → 020-S`) were removed in rev 18 before archival, so
no dependency now references an archived shipment. After the split, Strand A and Strand B are
independent of each other.

`018.008-T` carries one `related_to` **semantic link** to `019.007-T`. That is traceability, not a
dependency edge: it is not a `blocks` relation, it does not appear in `backlogit dep list`, and it
does not create a `current → future` dependency.

### R16.9 Accepted residual risk (interim)

| Risk | Disposition |
|---|---|
| No mechanical CI enforcement until Strand B lands | **Accepted.** The corrected contract is regeneration-vulnerable in the interim; the prohibition still binds the agent contract, which is what role enforcement reads at mutation time. Tracked by `020-F`. |
| No automated persistence route until Strand A lands | **Accepted, operator-only.** Stage halts at `STAGE_ARTIFACTS_UNCOMMITTED`; the Orchestrator must halt and must not run Step 1.5 item 3; only the operator may commit, push, open/approve the merge-commit staging PR and re-verify. **The dark run cannot finish autonomously past that checkpoint while the operator is AFK.** Tracked by `019-F`. |
| Deferred work cannot be scheduled until the harness execution model is decided | **Accepted.** `021.001-T` is the decision; both deferred features are `blocked` with no shipment, so nothing is claimable. |
| `SAFE_CLOSE` shipment-record close is tool-blocked (backlogit exit 9) | **P0, pre-existing, out of scope for this unit** — see §R16.3.3. `017-S` executes, reviews and merges normally, then pauses at Ship Step 6 closure for operator disposition. Recorded as a stash entry. |

Neither capability risk is a regression: both describe capability **not yet added**, not protection
removed. The closure blocker is **not** a regression either — it predates this unit and affects
every partial-feature shipment in the workspace.

### R16.10 Plan hardening status

Plan hardening was **applied** in rev 18 and **extended** in rev 19, not merely asserted. The
hardening actions of rev 18 are: (1) replacing the unreachable `CASCADE` manifest with the proven
task-only shape and tracing **every** pre/post gate (§R16.3.2); (2) surfacing the step-8 closure
conflict as a named P0 with executed evidence instead of an unexamined assumption (§R16.3.3);
(3) removing the invalid shipment-status lock and replacing it with a representation that has no
claimable surface at all (§R16.7); (4) making the P-004 red phase unconditional with explicit H0
evidence (§R16.6); and (5) replacing an asserted Orchestrator authority with an operator-only halt
contract (§R16.5). Rev 19 adds (6): converting "the `CASCADE` shape was not reached" into a proven
"it is unreachable by construction", with the role-boundary citation and the executed engine
evidence that isolates the blocker to the Ship pre-mode contract (§R16.3.1a) — closing the repair
route a later cycle attempted, so it is not re-attempted a fourth time.

The current reduced unit remains **text-only, single-task, single-domain, with no destructive or
irreversible operation**. **Each deferred strand is intrinsically more complex and requires its own
`impl-plan`, `plan-harden` and `plan-review` before any shipment is created for it** — see each
strand's plan and `021.001-T`.

---

## Historical audit appendix (revisions 1–15 — non-governing)

**Source document**:
`docs/decisions/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-deliberation.md`
(including its **§8 Revision 2 addendum**, **§9 Revision 3 addendum**, **§10 Revision 6
addendum**, **§11 Revision 7 addendum**, **§12 Revision 8 addendum**, **§13 Revision 9
addendum**, **§14 Revision 10 addendum**, **§15 Revision 11 addendum**, **§16 Revision 12
addendum**, **§17 Revision 13 addendum**, **§18 Revision 14 addendum**, and **§19 Revision 15
addendum**)
**Stash entry**: `638A410B`
**Requires plan hardening**: **yes** — see §9.

**Revision 2** incorporated the seven-persona plan review. Three rev-1 claims were falsified and
are corrected here: a **third** authorization surface exists (`_stage.agent.md` L42), the rev-1
P-010 wording was **self-deadlocking**, and deleting 3e **does** lose a behaviour. See
deliberation §8.

**Revision 3** incorporates the round-2 review. A **fourth** authorization surface was found — the
Step 1.5 **preamble**, which is commit-shaped and marker-free — and the detector specification is
tightened per-construct (block-level verdicts masked mixed blocks), structurally (bold lead-in, not
ATX heading; cell-scoped, not row-scoped), and defensively (guarded default-branch resolution,
`pull`/`checkout` exclusions, permissive-vs-verificational discrimination). Sub-epic A is split
A1/A2/A3 for granularity. See deliberation §9.

**Revision 4** incorporates the PR #54 current-HEAD review (two P-021 C1 same-contract blockers).
Rev 3's §10.1 declared a **P-002/P-004 deviation** and asked Ship to apply `harness-ready` from a
prose label — which Ship's Step 2 cannot honour, because it invokes `harness-architect` for every
queued task lacking that label and that skill only produces Go `_test.go` harnesses. Rev 4 removes
the deviation entirely: the release unit is restructured so the **existing** harness path is
genuinely satisfied by a real Go regression harness that drives the bash gate (§5.0), matching the
established `tests/integration/build_script_test.go` precedent of a Go test whose whole subject is
a non-Go script. No policy is amended, no test-first requirement is weakened, and no fake
scaffolding is added. Rev 4 also reconciles §7.2's C2 contract with task `018.011-T` on a single
**fixture-backed, no-dirty-tree-edit** mutation proof.

**Revision 5** incorporates the PR #54 current-HEAD review, cycle 4 (one finding, thread
`PRRT_kwDOTPuhps6hnPkD`). Rev 4's **R5 accepted the residual** that a future render could wipe the
job-`lint` gate step, leaving CI green with the detector never invoked — which directly contradicted
§1's "cannot **silently** reopen the gap" and the feature DoD. Rev 5 **does not narrow the
objective**. It instead identifies the tracked enforcement path that already survives the render
boundary — the §5.0 Go harness, invoked by the **generated-baseline** `Test (race)` step
(`go test -race -mod=readonly ./...`, job `expensive`) rather than by
a LOCAL DIVERGENCE step — and makes that path explicit, load-bearing, and mechanically
self-verifying (§5.0 "Regeneration-resistant enforcement", AC-C1.6, AC-C2.8). R5 is reclassified
from accepted residual to mitigated; the narrower, genuinely-visible residual is recorded as R12.
**Rev 12 corrects the provenance claim in the sentence above**: `go test -race -mod=readonly ./...`
is the **observed installed** command, not the generated baseline. The template emits
`{{TEST_COMMAND}}` and this workspace's recorded render input is `go test ./...`, so what a render
re-emits is a **full-suite `./...` invocation in job `expensive`** — which is all R5 needs — and
**not** those exact flags. See §5.0 "Invocation provenance" and AC-C1.7.

**Revision 6** incorporates the PR #54 current-HEAD review, cycle 5 (one finding, thread
`PRRT_kwDOTPuhps6hqWDe`, comment 3993998429, plan L517). Revisions 1–5 corrected *what* Step 1.5
**authorizes** but never touched *how* Step 1.5 **discovers** the work it must gate. Its two inputs
are `git status --short -- .backlogit/` and `git log origin/main..main` — a `.backlogit/`-only
pathspec and a **default-branch-only** commit range. Once B2/B3 require Stage to commit on
`chore/stage-{shipment_id}` / `chore/stage-{chore-slug}` and leave the worktree there, **both
inputs read "synchronized"**: the commit is not on the default branch and (for a no-shipment run)
the written paths are not under `.backlogit/`. Step 1.5 then falls through to its `origin/main`
manifest check and fails — so the no-shipment route B2 documents end to end was **unexecutable as
written**. Rev 6 makes the Stage artifact branch an **explicit handback** (`stage_branch`,
`stage_head_commit`, `stage_outcome`, `stage_artifact_paths`), makes discovery **branch-specific**
(`origin/main..HEAD`, proven equal to `origin/main..$STAGE_BRANCH` by a HEAD assertion that doubles
as the P-016 single-worktree gate), widens the dirtiness pathspec to the Stage artifact paths, and
gives step 4 a **second verification arm** keyed on `no-shipment` that needs no `shipment_id`. No
task, file, fixture, harness function, or dependency edge is added; 017-S stays one shipment.

**Revision 7** incorporates the PR #54 current-HEAD review, cycle 6 (two findings, threads
`PRRT_kwDOTPuhps6hrC0c` / comment 3994280334 and `PRRT_kwDOTPuhps6hrC0o` / comment 3994280351).
Rev 6's no-shipment arm iterated `git show origin/main:{path}` over a path set whose documented
default was the **directory list** `.backlogit/ docs/plans/ docs/decisions/ docs/memory/`. `git
show` on a tree **exits 0**, and those four directories already exist on `origin/main` in every
repository the gate will run in — so the loop could pass having read **no file this Stage run
wrote**, and paired with the degraded `stage_head_commit` default (`git rev-parse HEAD`, routinely
already an ancestor of `origin/main`) the arm reported a **false pass**. Rev 6 removed the
zero-iteration loop and left the always-satisfied one. Rev 7 makes `stage_artifact_paths` a
**non-empty set of concrete repository-relative file paths tied to `stage_head_commit`**, derived
deterministically from that commit's own changed-file set when the handback does not supply it,
validated for file-ness / root-containment / set-equality when it does, and verified with `git
cat-file -t` = `blob` rather than `git show` (§6.1.2 correction 4). Step 3 re-records the handback
commit after it creates one (correction 5), so the derived set is taken from the commit that
actually carries the artifacts. Step 1's dirtiness pathspec is pinned to the **fixed** allowed-root
constant rather than the handback field, closing the same class of hole one step earlier. Rev 7
also refreshes the stale plan-revision and section cross-references in the session memory record.
No task, file, fixture, harness function, or dependency edge is added; 017-S stays one shipment.

**Revision 8** incorporates the PR #54 **adversarial review** (anchor GPT-5.6 Sol, with
GPT-5.4-mini, Claude Sonnet 5, and Claude Opus 5; no route degradation), which classified every
finding below as **same-contract completion** under P-021. Four P1 blockers and five P2/P3
precision defects are remediated. **Blocker 1 — aggregate verification.** Rev 7's no-shipment arm
derived its verified file set from `stage_head_commit`'s **single-commit** diff, so artifacts
introduced by *earlier* commits on a multi-commit Stage branch were never read on `origin/main`: a
session that writes the deliberation in commit 1 and the plan in commit 2 had commit 1 covered by
ancestry alone. Rev 8 introduces `stage_base_commit` — captured by the **Orchestrator immediately
before it invokes Stage** and retained as the authoritative pre-invocation value — and derives the
artifact set from a stable **two-tree** diff over `{stage_base_commit} {stage_head_commit}`, which
reads identically before and after the merge (unlike the brittle `origin/main..{commit}` range
form, empty by construction post-merge). **Blocker 2 — reachable `no-shipment`.** Rev 6/7 built a
`no-shipment` verification arm that no producer could ever trigger: the installed Stage Step 5.5
and Step 6 pre-summary gate both **halt** without a `shipment_id`, so `stage_outcome: no-shipment`
was unreachable and the arm was dead prescription. B3 gains the narrowly-scoped Stage-side change
that lets a **reviewed, valid** empty-harvest terminal outcome emit the full handback and stop —
with P-003 failures preserved as failures, never re-labelled as success. **Blocker 3** rewrites
`018.007-T`'s description to a single authoritative contract (its rev-6 prose was still
operative-looking). **Blocker 4** reconciles §6.2 and `018.008-T` onto one identical **six**-criterion
B2 contract. **P2/P3**: line-ordinal locators for the Ship prohibition are replaced with structural
ones; generated-baseline evidence cites the exact `Test (race)` command (**superseded at rev 12** —
that command is the *observed installed* one, not the generated baseline; see §5.0 "Invocation
provenance"); fixture-prefix wording is
corrected to the real `directpush-` names; A2's red obligation is scoped to `--self-test`; and
stale Constitution Check and deliberation-header references are corrected. No task, file, fixture,
harness function, or dependency edge is added; 017-S stays one shipment of nine items.

**Revision 9** incorporates the PR #54 current-HEAD review, cycle 7 (one visible Copilot finding,
thread `PRRT_kwDOTPuhps6hrbAk`, comment 3994429008, on `018.009-T:22`), classified **same-contract
completion** under P-021. Revisions 1–8 gave Stage the *permission* to create and commit on a
dedicated artifact branch (B2's P-010 grant, B3's Role Boundary cell) and the *obligation* to
report what it did (AC-B3.5's handback), but **never an operative workflow step that performs
either action**. The installed `_stage.agent.md` runs triage → grouping → deliberation → planning →
review → harvest → shipment → archive → summary and contains **no branch operation and no commit
operation anywhere**. An executor satisfying AC-B3.1–AC-B3.6 exactly as written could therefore
finish the session still on the default branch, having written every artifact there — the precise
`fdff9e4` shape this release unit exists to close — while emitting a handback whose
`stage_head_commit` and `stage_artifact_paths` describe commits that were never made on a branch
that was never created. Rev 9 closes it with **two operative steps** in the file B3 already owns:
a **pre-mutation branch gate** (Step 1.9) that runs once a stable scope slug is derivable and
before the session's first backlog/doc write, and a **post-mutation artifact commit** (Step 5.7)
that runs after every Stage mutation and before the Step 6 handback. Rev 9 also removes the
**shipment-ID naming dependency** that made the branch name underivable at gate time: the branch is
`chore/stage-{scope-slug}`, a shipment ID may *be* the slug when one is already known, and it is
never *required* — Step 5.5 creates the shipment long after the branch must exist. One acceptance
criterion is added (**AC-B3.7**) covering both orderings. No task, file, fixture, harness function,
shipment, or dependency edge is added; 017-S stays one shipment of nine items.

**Revision 10** incorporates the PR #54 **second adversarial review** (anchor GPT-5.6 Sol, with
GPT-5.4-mini, Claude Sonnet 5, and Claude Opus 5; no route degradation), which classified every
finding as **same-contract completion** under P-021. Four visible Copilot threads and one
adversarial-only finding are remediated. **Blocker 1 — the handback outcome pair could fail open.**
Rev 6–9 derived `stage_outcome` *unconditionally* from whether a `shipment_id` was returned, so a
Stage run that reported `stage_outcome: shipment` but whose `shipment_id` failed to reach the
Orchestrator was **silently reclassified** as a successful `no-shipment` — the loudest failure the
pipeline has, converted into a passing terminal. Rev 10 makes the field **preserved when reported**
and derived only when **absent**, and adds a fail-closed **pair validation** before arm selection
(`shipment` requires a non-empty `shipment_id`; `no-shipment` requires none; an unknown value or an
inconsistent pair halts `STAGING_GATE_FAIL`). **Blocker 2 — the pre-gate deferral rule was
under-inclusive.** Rev 9's Step 1.9 deferral named exactly two mutations, while the installed Stage
contract mandates a stash-classification checkpoint, a contextual-grouping/operator-selection
checkpoint, and the **write half** of late-identifier reconciliation — all of which precede
Step 1.9 and are all tracked. Rev 10 makes the rule **categorical and exhaustive** (Step 1.9
precedes *every* tracked Stage artifact mutation), enumerates the three classes explicitly, and
keeps gitignored disposable index/checkpoint/hook-queue operations **distinct** rather than
overclaiming them. **Blocker 3 — out-of-root detection was unimplementable.** Step 5.7 told Stage
to halt on a change outside `STAGE_ARTIFACT_ROOTS` while prescribing only a root-scoped `git add`,
which *ignores* such a change rather than detecting it. Rev 10 requires an explicit NUL-safe
full-tree status inspection **before** staging, with a concrete parse algorithm covering tracked
modifications, deletions, both rename endpoints, and recursed untracked files, and a fail-closed
halt that leaves pre-existing unrelated dirt **untouched**. **Blocker 4 — AC-A2.5 could not prove
what it claimed.** A single driver-scoped `-run` invocation cannot establish that another test is
red; rev 10 splits it into two exactly-anchored invocations with a non-vacuity guard. **Adversarial
finding — the harness lifecycle deadlocked against Ship's full-suite gate.** Rev 9's H0 left all
eight per-task assertions simultaneously red while Ship runs `go test ./...` after **every**
task, so no task could ever complete. Rev 10 adds one minimal **per-task activation** mechanism
(§5.0.1) that keeps the red phase honest and observable per task without wedging the default suite.
No task, file, fixture, shipment, or dependency edge is added; 017-S stays one shipment of nine
items.

**Revision 11** incorporates the PR #54 **third adversarial review** (anchor **GPT-5.6 Sol**, with
**GPT-5.4-mini**, **Claude Sonnet 5**, and **Claude Opus 5**; 4/4 usable, **no route degradation**),
which **unanimously** confirmed two P0 direct-completion blockers in the rev-10 activation mechanism
plus one stale cross-reference. All three are **same-contract completion** under P-021.
**Blocker 1 — C1/C2 activation was self-disarming.** Rev 10 keyed C1's default-mode activation on
the *presence of the job-`lint` step* — the very surface `TestDirectPushGate_CIWiringIsBlocking`
exists to verify — and C2 **reused C1's predicate**. A regeneration that dropped the step therefore
made **both** functions `t.Skip` instead of fail, and MP2 (which lives inside C2) never executed. The
one control R5 rests on disarmed itself precisely when it was needed, and `go test ./...` stayed
green. **Blocker 2 — H0 violated the installed P-002/P-004/`harness-architect` contract.** That
contract requires `go test ./...` **red with every generated test failing its expected
not-implemented marker** before *any* task may be labelled `harness-ready`; rev 10 offered **one**
failing function (the A1 anchor) and **seven skips**, which is not that postcondition. But simply
reverting to rev 9's all-eight-red re-created the Step 4.3 deadlock rev 10 existed to remove.
Rev 11 resolves both with **one** added artifact: a tracked, non-generated, fail-closed
**harness-state manifest** at `tests/integration/testdata/directpush-gate/harness-state.json`,
initialized by H0 to `{"phase":"red","completed":[],"terminal":false}`. Activation is derived
**entirely from that manifest** and never from any surface the harness asserts on, which kills the
self-disarming class at the root: in **red** phase all eight functions activate and fail with their
own task-specific marker (the literal 8/8 the installed contract asks for); in **build** phase the
default suite runs the **completed set** and skips only not-yet-started tasks (so Step 4.3 is
green after every task); in **terminal** phase all eight substantive assertions run with no
activation skip anywhere. Targeted `-gatetask={ID}` always executes the real assertion regardless of
phase, so every task keeps its observed red→green boundary. **Blocker 3 — a stale deleted-`3e`
reference.** `018-F`'s DoD still named `_orchestrator.agent.md` Step 1.5 **sub-step 3e** as a live
authorization surface although B1 deletes it; it is replaced by the live **preamble** + **3a** pair
plus an explicit **3e-absence** clause. **No foundational contract is touched**: this revision plans
**no** modification to `.github/skills/harness-architect/SKILL.md`, to P-002/P-004, or to
`_ship.agent.md` Step 2 or Step 4.3 — see §5.0.1 "Why the release-specific manifest satisfies the
installed contract literally". No task, harness function, fixture, shipment, or dependency edge is
added; 017-S stays one shipment of nine items.

**Revision 12** incorporates the PR #54 **post-remediation adversarial re-review** (anchor
**GPT-5.6 Sol** with **GPT-5.4-mini**, **Claude Sonnet 5**, and **Claude Opus 5**; **4/4 usable**,
**no route degradation**), which classified every finding as **same-contract completion** under
P-021. Three P1 residuals in the rev-11 mechanism plus a set of direct consistency defects are
remediated. **P1-1 — MP5 deadlocked a legitimate state.** Rev 11 retired the rev-10 surface
predicates from *activation* but **retained them as failure probes** inside MP5, which asserted
that any task outside `completed` must have its surface absent. C2's "surface" was C1's old
predicate — the job-`lint` step — which **C1 itself creates**. So in the entirely legitimate
post-C1 / pre-C2 window, C2's probe read `true`, MP5 fired inside A1, and Ship's default full
suite went red: the exact Step 4.3 deadlock revisions 10 and 11 exist to remove, reintroduced by
the very check meant to protect the mechanism. **MP5 is RETIRED** (removed, not renumbered — §5.0.1).
The principle that replaces it is stated once and applied everywhere: **activation and failure
alike must never be derived from a surface the harness asserts on.** Rollback detection for a
**completed** task is already supplied by that task's own substantive assertion, which is active in
default mode from its completion onward; an **incomplete future** task may legitimately have its
prerequisites and even parts of its surface present without that being a defect, and must not fail
the default suite. **P1-2 — a consistent terminal-manifest rollback could skip C2 before MP2.**
Because MP2 lives *inside* C2 and C2's default-mode activation is read from the same manifest MP2
validates, a coordinated edit that rewound `completed`, `phase`, and `terminal` **together** left a
self-consistent `build` state in which C2 never ran and MP2 never fired. Rev 12 adds one small,
tracked, non-generated **terminal completion witness** at
`tests/integration/testdata/directpush-gate/terminal-witness.json`, independent of the main
manifest, plus **MP6** — a witness/manifest cross-agreement check that runs inside `gate()` **before
any activation decision**, so it cannot be skipped past. Rolling back only the manifest, or only the
witness, now **fails loudly**. The residual that survives is recorded honestly rather than papered
over: deleting **both** in one diff reproduces a state indistinguishable at runtime from legitimate
pre-C2 build, and is caught by **diff review**, not by runtime history (**R22**). No claim is made
that a stateless validator detects every coordinated rollback. **P1-3 — CI regeneration provenance
was overstated.** Revisions 5–11 cited `go test -race -mod=readonly ./...` as the *generated
baseline*. It is not: the template emits `{{TEST_COMMAND}}`, and this workspace's recorded render
input (`.autoharness/harness-manifest.yaml`, `.autoharness/workspace-profile.yaml`) is
`go test ./...`. The `-race -mod=readonly` flags are the **observed installed** command, not a
guarantee a render reproduces. The R5 mitigation is restated at the strength it actually has — it
needs **a full-suite invocation including `./...` that runs the integration harness**, which the
baseline does provide — and a **post-render verification** obligation (**AC-C1.7**) replaces the
implicit promise that the exact flags survive. Rev 12 also fixes the direct consistency defects the
re-review enumerated: the memory and deliberation artifact indexes, the "five checks" miscount, the
residual `completed`-prefix wording, the unqualified zero-skip claims, the `gateDeps` direction
gloss, and AC-C2.4's reference to the historical §2 enumeration that still counted the deleted `3e`.
**No foundational contract is touched**: as at rev 11, this revision plans **no** modification to
`.github/skills/harness-architect/SKILL.md`, to P-002/P-004, or to `_ship.agent.md` Step 2 or
Step 4.3. No task, harness function, fixture, shipment, or dependency edge is added; 017-S stays one
shipment of nine items.

**Revision 13** incorporates the PR #54 **final capped post-remediation adversarial re-review**
(anchor **GPT-5.6 Sol** with **GPT-5.4-mini**, **Claude Sonnet 5**, and **Claude Opus 5**; **4/4
usable**, **no route degradation**), run under an **explicit operator authorization of one
additional Stage remediation plus re-review cycle** after the rev-12 cycle budget was exhausted and
the session was recorded BLOCKED. All six findings are **same-contract completion** under P-021;
none is P0. **P1-1 (MEDIUM) — targeted selector precedence vs. terminal witness activation.** The
rev-12 activation table placed the witness rows **before** the `-gatetask` rows and stated that a
valid witness "forces `C2` active; all eight activate" without qualifying the mode. Read
sequentially — which is how a normative table is read — a **targeted** invocation in terminal state
activated all eight bodies, contradicting MP3, AC-C2.7, AC-C2.9, and AC-C2.11, each of which
requires exactly the selected test to run and the seven unrelated functions to skip. Rev 13 splits
`gate()` into two explicitly ordered stages — a **mode-independent integrity gate** (MP1 + MP6
validation) and a **mode-dependent activation resolution** in which a valid `-gatetask` selector has
**activation precedence** — and scopes every "forces `C2` active / all eight activate" clause to
**default mode**. Witness integrity still runs **first** and can fail a targeted invocation; what it
may no longer do is activate unrelated test bodies (§5.0.1). **P1-2 (LOW) — H0 harness eligibility
with dependency-blocked tasks.** Resolved **against the installed contract without changing it**:
`_ship.agent.md` **Step 2** selects the shipment's **full `queued` task set** for harness
generation and **Step 3** separately builds the dependency-ordered execution queue, so
implementation dependencies gate **claim and execution**, never upfront harness generation (§5.0
"H0 eligibility"). **P1-3 (LOW) — control-state input exemption.** The independence invariant is
narrowed precisely: no activation and no gate-level failure may derive from a **substantive
surface** the same test asserts on; the **dedicated harness control-state inputs** — the validated
manifest and the terminal witness — are expressly **allowed and required** inputs to MP1/MP6
(§5.0.1 "Control-state inputs"). **P1-4 (LOW) — plan/task AC-C1 parity.** §7.1 and `018.010-T` now
carry **byte-identical** AC-C1.1–AC-C1.7 text, verified by content parity rather than ID parity.
**P1-5 (LOW) — vacuous terminal lint-removal evidence.** §5.0.2 point 7 used targeted
`-gatetask=C1`, which activates C1 **by construction** and therefore cannot prove terminal
**default-mode** activation, and never exercised C2 at all. It is replaced by a **default-mode** run
against a **copied terminal-state repository fixture** under `t.TempDir()` with the lint step
removed, asserting C1 **fails for the missing step** *and* C2 **executes** — both by test-event
name. The targeted command is retained **separately**, as selector evidence only. **P1-6 (P2, LOW)
— residual provenance overstatement.** §6.3.2's "Regeneration exposure" paragraph still called the
installed `Test (race)` command the generated baseline; it now carries the same exact distinction
used everywhere else. No task, harness function, fixture, shipment, or dependency edge is added;
017-S stays one shipment of nine items and **no task is re-sized**.

**Revision 14** incorporates the PR #54 **rev-13 post-remediation adversarial re-review**, run under
a second **explicit operator authorization of one remediation plus one re-review cycle**. The
verdict was **NOT READY** with **0 P0**, **one LOW-confidence P1**, and **four LOW-confidence P2**
findings, all `post_remediation_residual` and all **same-contract completion** under P-021.
**P1 — the rev-13 copied-fixture evidence named no executable mechanism.** §5.0.2 point 7 required a
**default-mode** `go test` against a copied terminal-state repository under `t.TempDir()` but
specified only bare commands. The harness's own `repoRoot(t)` helper resolves the **real checkout**
by walking up to `go.mod` from the process working directory, so an in-process run could never see
the copy; and because the observation lives **inside** `TestDirectPushGate_FullCorpusClean` (C2), an
unguarded re-run would have spawned itself recursively. Rev 14 replaces the bare commands with one
concrete bounded mechanism — a deterministic **tracked-file copy helper**, a **child `go test`
process** whose `Dir` is the copied module root, a **non-recursive sentinel environment variable**,
and **`go test -json` event parsing** — specified in §5.0.2 point 7, AC-C1.6, and AC-C2.11.
**P2-1 — H0 eligibility citation overstated.** Rev 13 asserted that "**no step** of `_ship.agent.md`
Step 2 consults the dependency graph". Step 2's closing paragraph **does** direct Ship to *verify the
dependency graph before proceeding* when dependency operations are supported. The citation is
corrected to the faithful distinction — Step 2 may **verify graph validity**, while dependency
**readiness** gates execution **order** in Step 3 and **claim** in Step 4 — and the plan now requires
Ship to pass the **exact eight task IDs** through `harness-architect`'s `${input:tasks}` parameter
(§5.0 "H0 eligibility"). **P2-2 — targeted C2 red under-specified.** §5.0.2 point 4's
"`-gatetask=C2` exits non-zero" accepted any non-zero exit; it is now anchored to a named
`--- FAIL: TestDirectPushGate_FullCorpusClean` with the **MP2 non-terminal-state** reason, with five
impostor exits explicitly rejected. The paired "green by assertion, never by bookkeeping" rule is
**qualified** rather than contradicted: C2 is the one task whose completion bookkeeping is itself
part of its substantive contract, and it still cannot pass on bookkeeping alone. **P2-3 / P2-4 — the
memory current-artifact index and the deliberation addenda index were stale**; both are advanced to
this revision without erasing history. No task, harness **function**, fixture, shipment, or
dependency edge is added; 017-S stays one shipment of nine items and **no task is re-sized**.

The planning artifacts themselves are committed on the dedicated branch
`chore/stage-pipeline-policy-gap` and reach
`main` only via PR — the exact path this shipment installs. This shipment does not reproduce the
`fdff9e4` pattern it exists to close.

---

## 1. Objective

Make a dedicated Stage/admin branch plus a pull request the **only** path by which a Stage
artifact reaches the default branch — including when no shipment is formed — and enforce it
mechanically so regeneration of the autoharness-generated contract files cannot silently reopen
the gap.

## 2. Root cause — four authorization surfaces

| Surface | Authorizing text | Shape |
|---|---|---|
| `.github/agents/_stage.agent.md` Role Boundary table, **Git** row, Allowed cell | "Commit backlog/planning artifacts **on default or admin branch**" | **commit**, table cell |
| `.github/policies/workflow-policies.md` P-010 **Stage MAY** bullet | "Commit backlog and planning artifacts (**on the default branch** or a dedicated chore/admin branch)" | **commit**, list bullet |
| `.github/agents/_orchestrator.agent.md` Step 1.5 sub-step **3e** | "**Attempt a direct push to `main` first** … fall back to creating a staging PR" | **push**, instruction |
| `.github/agents/_orchestrator.agent.md` Step 1.5 **preamble** (rev 3) | "verify that all staging artifacts … **are committed to the default branch** and present on the remote" | **commit**, postcondition |

**Rev 3 — fourth surface.** The Step 1.5 preamble states the gate's postcondition as artifacts
being "committed to the default branch", with no PR in the stated end-state. Left as-is it both
(a) remains a compliant reading of direct commit-to-default, defeating §1's objective, and
(b) sits in the scanned corpus as a commit-shaped, marker-free construct — so it would either
trip the gate (making zero-findings unreachable) or force the detector to be written narrowly
enough to miss the `_stage.agent.md` L42 shape it exists to catch. It is corrected in B1.

`.github/instructions/role-enforcement.instructions.md` makes the **agent's own Role Boundary
table** the authoritative permission set at mutation time — so `_stage.agent.md` L42, not P-010,
is what the agent actually read before producing `fdff9e4`. Correcting only the other two would
produce a **false green**.

Ship's mirror prohibition already exists in the same policy — P-010's **Ship MUST NOT** bullet
"Commit or push directly to `main`", anchored by quoted text rather than by line ordinal, per the
ci.yml ledger's own "by description, not line number" rule. The fix makes Stage symmetric with it.

All three files carry `Generated by autoharness | Template: …` provenance — prose-only correction
is reversible by a future render with no signal.

## 3. Surface

| File | Change | Task |
|---|---|---|
| `tests/integration/directpush_gate_test.go` | **New.** Go regression harness driving the gate; 8 test functions, one per task, plus the **rev-10 per-task activation** helper (§5.0.1). Produced by `harness-architect` at **Ship Step 2**, not by a task (§5.0). | H0 |
| `scripts/testdata/directpush/*.md` + `directpush-manifest.json` | **New.** 14 fixtures + sorted verdict manifest. | A1 |
| `scripts/check-direct-push-language.sh` | **New.** Not-implemented stub (H0) → driver (A2) → detector (A3). | H0, A2, A3 |
| `.github/agents/_orchestrator.agent.md` | Reword Step 1.5 preamble; delete 3e; fold its branch-point into 3a. **Rev 6**: require the Stage handback record in Step 1; make Step 1.5 discovery branch-specific; widen the dirtiness pathspec; add step 4's no-shipment arm. **Rev 7**: make the no-shipment arm verify concrete files (`cat-file -t` = `blob`), pin step 1's pathspec to the fixed artifact-root constant, and re-record the handback commit in 3a. **Rev 8**: capture `stage_base_commit` before invoking Stage and derive the artifact set from the **aggregate** `{stage_base_commit} {stage_head_commit}` two-tree diff. **Rev 10**: preserve a **reported** `stage_outcome` and derive it only when absent, and validate the outcome/`shipment_id` **pair** fail-closed before selecting an arm. | B1 |
| `.github/policies/workflow-policies.md` | P-010 Stage bullets + Amendment Log **1.25.0**. | B2 |
| `.github/agents/_stage.agent.md` | Role Boundary Git row + PR row note. **Rev 6**: emit the staging handback record in the Step 6 summary contract. **Rev 7**: emit `stage_artifact_paths` as concrete file paths. **Rev 8**: emit `stage_base_commit` and the aggregate derivation, and make the `no-shipment` terminal outcome **reachable** in Step 5.5 / Step 6 without weakening P-003 failure handling. **Rev 9**: add the two **operative** steps the permission always presupposed — a pre-mutation **branch gate** (Step 1.9) and a post-mutation **artifact commit** (Step 5.7) — and register both in the Step Sequence Contract. **Rev 10**: make the Step 1.9 deferral rule **categorical and exhaustive** over every tracked pre-gate mutation, and give Step 5.7 an implementable out-of-root **detection** step before it stages anything. | B3 |
| `.github/workflows/ci.yml` | One step in job `lint`; LOCAL DIVERGENCE ledger. | C1 |
| `.github/CODEOWNERS` | Own the new gate + the three unowned gate scripts. | C1 |

**Name change (rev 2)**: `check-direct-push-language.sh`, not `…-policy.sh`. The gate detects
**contract language**, not a push event. The rev-1 name implied behavioural enforcement it cannot
deliver.

**Not touched**: `.copilot/installed-plugins/**` and `.autoharness/staging/**` — gitignored
(`.gitignore:60`, `.autoharness/.gitignore:3`), therefore uncommittable, unreviewable, and
overwritten on reinstall (deliberation D2).

## 4. Ordering (test-first; `main` stays green)

```text
H0  Ship Step 2: harness-architect                → Go harness + stub script; go vet 0, go test RED
A1  fixture corpus + manifest                    → test data only
A2  driver over the fixture corpus               → reject fixtures RED (detector still stubbed)
A3  detector implementation                      → fixtures GREEN, real-tree scan RED
B1..B3  contract correction (4 surfaces)         → real-tree scan GREEN
C1  CI wiring + ledger + CODEOWNERS              → gate blocking from first wiring
C2  coupling verification                        → zero-findings + fixture-backed mutation proof
```

**Rev 10 — the default suite stays green between tasks (mechanism corrected at rev 11).** Ship's
Step 4.3 runs `go test ./...` after **every** task, so the harness must not leave later tasks'
assertions failing in the default suite once an earlier task completes. Per-task red is established
by **targeted activation** (§5.0.1). **Rev 11**: the default suite runs the assertions named in the
harness-state manifest's `completed` set — never "the assertions whose surfaces exist", which was
the self-disarming rev-10 rule — and H0's red phase is the **literal 8/8**, all eight functions
active and failing with their own markers.

The gate enters CI only after the violations are removed, so `main` is never wedged and **no
advisory toggle is introduced** — strictly safer than a bounded advisory window, and simpler.
`blocks` edges: A1→A2→A3→{B1,B2,B3}→C1→C2. **H0 is not a task and carries no `blocks` edge** — it
is Ship's Step 2, which runs once up front for the whole queue and therefore precedes A1 by
construction (§5.0).

---

## 5. Sub-epic A — Failing regression harness

### 5.0 Ship Step 2 harness contract (P-002 / P-004) — rev 4

Rev 3 declared a P-002/P-004 **deviation** and asked Ship to label the tasks `harness-ready` from a
substitute bash harness. That is unexecutable. `_ship.agent.md` Step 2 partitions the queue on the
`harness-ready` label, invokes **`harness-architect`** for every task lacking it, and then halts
unless *every* queued task carries it. `harness-architect` only emits Go `_test.go` harnesses and
only applies the label after `go vet ./...` exits 0 and `go test ./...` is red. A Stage-applied
prose label would therefore be a **forged P-004 postcondition**, and withholding the label would
drive `harness-architect` to invent Go scaffolding for markdown and bash tasks. Rev 4 removes the
deviation instead of widening the policy.

**The harness is a real Go test whose subject is the bash gate.** This is not new scaffolding
invented for the label: `tests/integration/build_script_test.go`, `start_script_test.go`, and
`output_path_guard_test.go` are existing tests in this repository whose entire subject is a
non-Go script, and `internal/apperr/taxonomy_drift_test.go` is an existing Go **drift guard** over
a non-runtime invariant. This gate is the same shape.

**H0 deliverable** — produced by `harness-architect` at Ship Step 2, before any task is claimed:

1. `tests/integration/directpush_gate_test.go`, package `integration`, reusing the existing
   `repoRoot(t)` helper convention. **Eight** test functions, one per task (table below), plus the
   **per-task activation helper and manifest state machine** of §5.0.1 (rev 10, corrected rev 11),
   plus — **rev 14** — the two unexported **fixture helpers** of §5.0.2 point 7,
   `copyTrackedRepoFixture` and `runFixtureChildGoTest`, and the parent-only spawn block inside
   `TestDirectPushGate_FullCorpusClean` that calls them. Those helpers are **not** gate functions:
   the function count stays **eight**, they hold no `gateOrder` entry, and `gate()` never activates
   or skips them. They are listed here because they are authored at **H0** by `harness-architect`,
   like everything else in this file — **no task authors them**, and no task is re-sized by them.
2. `scripts/check-direct-push-language.sh` as a **structural stub**: correct shebang,
   `set -euo pipefail`, argument dispatch present, and every mode exiting non-zero after printing
   the marker `not implemented: direct-push detector`. This is the shell analogue of
   `harness-architect`'s `panic("not implemented: <reason>")` production stub — the module
   "compiles" (the script parses and runs) while every test fails for the intended reason.
3. **Rev 11** — `tests/integration/testdata/directpush-gate/harness-state.json`, the **harness-state
   manifest**, written exactly once at H0 with the literal content
   `{"phase":"red","completed":[],"terminal":false}`. It is **tracked**, **hand-authored (not
   generated)**, and sits under `tests/**`, which no autoharness template covers — the same
   render-boundary argument the table below already makes for the harness file itself. It records
   **implementation/harness state**, never backlog status (§5.0.1, rejected alternative 4).
4. **Rev 12** — `tests/integration/testdata/directpush-gate/terminal-witness.json`, the **terminal
   completion witness**, is **NOT an H0 deliverable**. Its **absence** is part of the H0
   postcondition and of every legitimate pre-C2 build state; only C2's completion creates it
   (§5.0.1, MP6). H0 MUST NOT write it, and a witness present at H0 is a fail-closed error.

**H0 eligibility — why all eight tasks are harnessed up front, even though seven have unmet
implementation dependencies (rev 13; citation corrected rev 14).** A reviewer asked whether
`harness-architect` may legitimately batch tasks that cannot yet be worked. It may, and the
**installed contract already says so** — this plan resolves the question by citing that contract
accurately, not by amending it:

* **`_ship.agent.md` Step 2 (Harness Generation)** — "This step runs once, up front — **not in a
  loop**." Its item 1 lists "all tasks for the target feature or chore that are in **`queued`**
  status"; item 2 partitions that list **solely** on the `harness-ready` label; item 3 invokes
  `harness-architect` **for the batch**; and item 4 halts unless **every queued task** carries the
  label afterwards. **Rev 14 — stated faithfully.** Step 2's closing paragraph does direct Ship,
  "when dependency operations are supported", to **verify the dependency graph before proceeding
  rather than assuming the backlog ordering is already valid". The rev-13 wording — "*no step of
  this sequence consults the dependency graph*" — was therefore **wrong as written and is
  withdrawn**. What Step 2 verifies is the graph's **validity**; what it does **not** do is filter
  the batch by dependency **readiness**. Harness generation remains an **upfront batch over the
  explicitly selected full shipment task set**, and a graph-validity check over that set is not a
  membership filter on it.
* **`_ship.agent.md` Step 3 (Build Ready Queue)** — runs **after** Step 2 ("Now that all tasks are
  harnessed") and is where dependency **readiness** first changes behaviour: item 3 "Sort the queue
  by dependency order (tasks with no unfinished dependencies first)". Claiming happens later still,
  at Step 4.1. **Readiness gates execution ORDER and CLAIM; graph VALIDITY is what Step 2 checks.**
* **`harness-architect` Step 1 — BOTH clauses, reconciled without overstating certainty (rev 15).**
  Step 1 contains **three** items that bear on selection, and rev 14 cited only the third. Quoted
  exactly: item 1 "Load the feature or chore and its **ready descendants** through backlog query or
  queue operations"; item 2 "If `${input:tasks}` is present, **restrict the scope to that explicit
  task set**"; item 3 "Exclude **blocked, done, or otherwise non-ready** work items". The
  **plan's reading** is that `${input:tasks}` supplies the scope and that the exclusion in item 3
  turns on the **lifecycle status values** it names — `blocked`, `done` — so that **all eight tasks
  of `017-S`, every one of them `queued`** (verified in `.backlogit/queue/`), are in scope. **An
  unmet task dependency is NOT `status: blocked`**: `blocked` is a value the backlog stores, and
  `018.005-T`–`018.011-T` do not carry it; they carry `queued` with `dependencies` edges, which is
  a different thing entirely.
  **What this plan does NOT claim.** It does **not** claim the phrases "ready descendants" and
  "otherwise non-ready" are *unambiguous*. They are **not**. A reader may reasonably interpret
  "otherwise non-ready" as covering **dependency-unready** work, in which case item 1's "ready
  descendants" and item 2's explicit-set restriction pull in opposite directions and the skill could
  legitimately refuse seven of the eight. **This plan does not resolve that ambiguity by assertion,
  does not invent an override, and does not amend the installed skill** (§11 keeps foundational
  contracts out of scope). It handles the ambiguity **operationally**, fail-closed:
  1. Ship passes the **exact eight task IDs** as `${input:tasks}` (below) and **verifies all eight
     are `status: queued`** immediately before invocation.
  2. If `harness-architect` **accepts** the batch, H0 proceeds and the literal 8/8 red is recorded.
  3. If it **refuses, drops, or silently omits** any of the eight — i.e. the returned selected set is
     **not** equal to the eight — **Ship MUST HALT BEFORE ANY MUTATION** and **route to the
     operator**, reporting the contradiction between that behaviour and Step 2 item 4's
     all-queued-tasks halt. **Silently proceeding with a partial batch is prohibited**, and so is
     relaxing the 8/8 red requirement to accommodate it. The set-equality check of the Mechanical
     check below is what makes a partial batch **detectable** rather than discovered late.
  **Does this ambiguity require the decomposition or lifecycle to change? No — and the alternatives
  are worse (rev 15).** The ambiguity lives in the **installed skill's wording**, not in this
  feature's shape. Removing the dependency edges to make all eight trivially "ready" would
  **falsify** a real and verified ordering constraint (A1→A2→A3→{B1,B2,B3}→C1→C2) purely to satisfy
  a tool's parser — the backlog would then misstate the work. Harnessing only the dependency-ready
  task (A1) would **halt the shipment at Step 2 item 4** anyway, trading a detectable halt for a
  guaranteed one. Splitting or merging tasks changes nothing, because the ambiguity is about
  **dependency state**, which any decomposition into ordered work necessarily has. The fail-closed
  halt above is therefore the **proportionate** response, and no task, dependency edge, or shipment
  membership changes on account of this finding.
* **Rev 14 — Ship MUST supply the explicit eight-ID input.** `${input:tasks}` is declared
  **Optional**, and when it is omitted the skill falls back to "all ready tasks under the feature" —
  a selection this plan must not delegate. Ship therefore passes the **exact eight `017-S` task
  IDs** — `018.004-T`, `018.005-T`, `018.006-T`, `018.007-T`, `018.008-T`, `018.009-T`,
  `018.010-T`, `018.011-T` — as the `${input:tasks}` value, which is what Step 2's "for the batch"
  invocation supplies. **Set equality with the shipment manifest is checked BEFORE invocation**
  (manifest items minus the covering feature `018-F`, which is resolved through `parent_id` and
  never executed), and Step 2 item 4's **all-queued postcondition halts on omission** afterwards —
  so an under-selected batch is caught on both sides of the call.

The ordering is therefore unambiguous: **dependency READINESS gates CLAIM and EXECUTION ORDER
(Steps 3–4); Step 2 may verify graph VALIDITY, but harness generation is an upfront batch over the
explicitly selected full shipment task set.** H0 batches **all eight** explicitly selected queued
shipment tasks and verifies a **literal 8/8** red before labelling any of them `harness-ready`.
This is not merely permitted but **required**: Step 2 item 4 halts unless *every* queued task
carries the label, so harnessing only the dependency-ready task (A1) would **halt the shipment** at
Step 2 rather than advance it.

**Mechanical check (recorded with AC-C2.11 evidence point 1, the H0 observation).** The H0 record
MUST assert set equality, not a count: the `${input:tasks}` set passed to `harness-architect`, and
the selected task set it reports, each equal **exactly** the eight task IDs of shipment `017-S` —
`018.004-T`, `018.005-T`, `018.006-T`, `018.007-T`, `018.008-T`, `018.009-T`, `018.010-T`,
`018.011-T` — with the covering feature `018-F` resolved through `parent_id` and **never executed**.
No task may be omitted on account of dependency state, and no task outside the manifest may be
added. A selected set that is a proper subset of those eight is an **H0 failure**, not a partial
success.

**Rev 13 note — if the installed contract is ever read to say otherwise (citation corrected rev
14).** Nothing here invents permission, and nothing here claims a behaviour the installed skill does
not have. Should a future revision of `_ship.agent.md` Step 2 or `harness-architect` Step 1
genuinely exclude dependency-unready tasks from batch harness generation, the correct response is to
**report the contradiction** between that contract and Step 2 item 4's all-queued-tasks halt — not
to relax this plan's 8/8 red-phase requirement, and not to amend the contract from inside this
release unit (§11 keeps every foundational contract out of scope).

**Red-phase evidence (P-004 precondition, taken literally; rev 11):** after H0, `go vet ./...`
exits 0 and `go test ./...` exits **non-zero** with **all eight** functions failing, each carrying
its **own task-specific** `not implemented: <ID> <surface>` marker (§5.0.1 marker table). The
manifest's `phase` is `red` and `completed` is empty, **the terminal witness is absent (rev 12)**,
and in red phase the activation helper
activates **every** function — **no function skips**. `Compilation: PASS`, `Red Phase: CONFIRMED`.
That is the installed postcondition `harness-architect` checks, met **literally** rather than by a
single anchor: **only after the literal 8/8 expected-red evidence is recorded** may the eight tasks
be labelled `harness-ready`. **Each task additionally has its own observed red**, taken by Ship at
task-claim time from that task's `harness_cmd`, which forces activation for exactly that task
(§5.0.1) — that per-task red is retained, not replaced by the 8/8 default-suite red. **Stage does
not apply the `harness-ready` label and this plan does not ask Ship to.**
`018-F.custom_fields.harness_status` stays `pending` until H0 runs.

**Why this no longer deadlocks Ship's Step 4.3.** The 8/8 red is the **red phase only**. A1's
completion performs the one-way `red`→`build` transition, after which the default suite runs the
**completed set** and skips only not-yet-started tasks, so `go test ./...` is green after every
task (§5.0.1). Rev 10 tried to buy that green by weakening the red phase to one anchor plus seven
skips; rev 11 keeps the red phase literal **and** the build phase green, because the two are now
distinct manifest phases rather than one static predicate set.

**Tool resolution — fail, never skip.** The harness resolves `bash` on `PATH` and **fails** the
test when it is absent; it MUST NOT `t.Skip`. A skip would make the red phase unobservable and
silently satisfy P-004 (P-012: record `TOOL_DEGRADED`, never silently skip). This diverges
deliberately from `build_script_test.go`'s `t.Skip("pwsh not available on PATH")`, because `pwsh`
is optional tooling there whereas `bash` is a hard prerequisite of every gate script in this repo.

**This rule is about tool absence, not about activation (rev 10, narrowed rev 11).** §5.0.1's
per-task activation does use `t.Skip`, and the two are not in conflict: a **missing tool** means the
assertion could not be evaluated and its result is unknown — which must fail — whereas an
**inactive task** means the assertion is not yet applicable, because the task that builds its
surface has not started. The discriminator is stated so an implementer cannot read one rule as
licence to weaken the other: a `bash` skip hides a failure that exists, an activation skip defers an
assertion whose subject has not been written yet. **Rev 11 narrows the licence further**: an
activation skip is legal **only** in `build` phase, **only** for a task whose ID is absent from the
manifest's `completed` set, and **never** for any reason derived from the state of a surface the
harness asserts on. Skips are impossible in `red` and `terminal` phase. Every other unexpected
condition — missing or malformed manifest, unknown or duplicate or out-of-order ID, skipped
prerequisite, inconsistent `phase`/`terminal` pair, unknown `-gatetask` value — **fails**, never
skips.

**Rev 12 extends the same discriminator to the failure direction.** The rule "activation must never
read a surface the harness asserts on" is not only about *skipping*; it is equally about *failing*.
Rev 11's MP5 obeyed the rule for activation and broke it for failure — it read those same surfaces
to raise a rollback error — and that is what deadlocked the legitimate post-C1 / pre-C2 state
(§5.0.1, "MP5 — RETIRED at rev 12"). The invariant is therefore stated in its general form: **no
assertion's activation and no gate-level failure may be derived from a SUBSTANTIVE SURFACE that the
same test asserts on** — agent, policy, workflow, or script behaviour. Rollback of a **completed**
task's surface is caught by that task's own substantive assertion, which default mode runs from its
completion onward; an **incomplete** task having its prerequisites or even parts of its surface
present is a legal mid-shipment state and MUST NOT fail the default suite.

**Control-state inputs are expressly exempt, and the exemption is narrow (rev 13).** The invariant
above is scoped to **substantive surfaces** — the product and governance artifacts this harness
exists to verify. It is **not** a prohibition on reading the harness's own **dedicated control-state
inputs**: the validated **harness-state manifest** and the **terminal completion witness** are
**allowed and required** inputs to MP1 and MP6, and both activation and fail-closed behaviour are
*supposed* to derive from them. The distinction is not a loophole and does not generalize:

| | Substantive surface | Dedicated control-state input |
|---|---|---|
| Examples | `_orchestrator.agent.md` Step 1.5, `workflow-policies.md` P-010, `_stage.agent.md` Role Boundary, `jobs.lint.steps[]`, `scripts/check-direct-push-language.sh` | `harness-state.json`, `terminal-witness.json` |
| What the harness does with it | **Asserts on it** — it is the subject of a test | **Reads it** — it is the control state of the lifecycle |
| May it drive activation or gate-level failure? | **No** — that is the self-disarming class (rev 10) and the MP5 class (rev 11) | **Yes** — that is precisely MP1's and MP6's job |
| Is it itself verified? | Yes, by its own substantive assertion | Yes, by **separate** mutation/integrity assertions (MP0/MP1/MP4/MP6, AC-C2.2's ten cases) |

The property that makes the exemption safe is that a control-state input is **never a surface whose
absence can self-disable its own check**: a missing, malformed, or mismatched manifest or witness
**fails closed** and fails **all eight** functions, whereas a missing job-`lint` step under the
rev-10 design **silently skipped** the very function that existed to notice it. The failure
directions are opposite, and that is the whole test for whether something belongs in the right-hand
column. Nothing else may be added to it: a file is a dedicated control-state input only if it is (a)
authored by the harness for the harness, (b) never a subject of any substantive assertion, and
(c) covered by its own integrity proof.

**Per-task harness map.** Each function is the `harness_cmd` boundary `build-feature` loops on, so
every task has exactly one observable red→green transition. **Rev 10**: every `harness_cmd` carries
the `-gatetask={ID}` activation selector (§5.0.1), which is what makes the transition observable
for a task whose surface does not yet exist.

| Task | Test function | Turns green when |
|---|---|---|
| A1 `018.004-T` | `TestDirectPushGate_FixtureCorpusIsComplete` | 14 fixtures + manifest exist; manifest keys sorted; bijection with the directory holds; 9 reject / 5 accept. **Rev 11**: A1 **initializes nothing** — H0 writes the harness-state manifest — but A1 **performs the one-way `red`→`build` transition** on completion and carries the activation machinery's static integrity check (§5.0.1, MP0). **Rev 12**: A1 no longer carries MP5 — **RETIRED** — nor any "probe non-vacuity" obligation, because the surface probes those referred to no longer exist (§5.0.1). It is the **dependency root**: `A1` is a transitive prerequisite of every other task, so it is a member of every non-empty dependency-closed `completed` set and therefore runs in every default-mode run from `build` onward; in `red` phase it runs because *everything* runs |
| A2 `018.005-T` | `TestDirectPushGate_SelfTestDriverEnumeratesFixtures` | `--self-test` **enumerates every manifest fixture by name and emits a per-fixture verdict**, and runs fixtures before the real-tree scan — the **driver** contract, which is what this task delivers. **Rev 10 correction**: the function asserts the driver contract and **not** the transient all-accept-stub behaviour, because a permanent assertion that the detector is still a stub would turn **red the moment A3 replaces it** and would wedge the suite from A3 onward. The stub-specific observations (every reject fixture reported failing; `--self-test` exit non-zero) stay **AC-A2.1**, taken from the script's own output at A2's boundary where they are true. The task's completion evidence is **two** exactly-anchored invocations, not one (AC-A2.5) |
| A3 `018.006-T` | `TestDirectPushGate_SelfTestPasses` | `--self-test` exits 0: actual == manifest verdict for all 14, plus both default-branch-resolution assertions |
| B1 `018.007-T` | `TestDirectPushGate_OrchestratorSurfaceClean` | scan scoped to `_orchestrator.agent.md` reports 0 findings **and** (rev 6) Step 1.5's discovery is branch-specific: Step 1 requires the handback fields, the unpushed-commit check resolves the reported branch and compares `origin/main..HEAD`, **no default-branch self-range literal remains in Step 1.5**, and step 4 carries both outcome arms. **Rev 7**: the no-shipment arm verifies concrete files with `cat-file -t` = `blob`; **no directory-prefix path set and no `git show`-over-`{path}` loop remains in that arm**; step 1's dirtiness pathspec is the fixed artifact-root constant; 3a re-records the handback commit. **Rev 8**: Step 1 captures `stage_base_commit` before invoking Stage, the arm asserts base→head ancestry and derives its set from the **aggregate two-tree** diff, and **no single-commit `diff-tree` derivation and no `origin/main..{commit}` range form remains** in that arm. **Rev 10**: a **reported** `stage_outcome` is preserved and derivation from `shipment_id` applies **only when it is absent**, and a fail-closed **pair validation** precedes arm selection |
| B2 `018.008-T` | `TestDirectPushGate_PolicySurfaceClean` | scan scoped to `workflow-policies.md` reports 0 findings **and** the Amendment Log carries row `1.25.0` |
| B3 `018.009-T` | `TestDirectPushGate_StageAgentSurfaceClean` | scan scoped to `_stage.agent.md` reports 0 findings **and** (rev 6) the Step 6 summary contract emits the staging handback record. **Rev 7**: that contract requires `stage_artifact_paths` to be concrete file paths, not directory prefixes. **Rev 8**: it also emits `stage_base_commit` and the aggregate derivation, and Step 5.5 / Step 6 admit a **reachable** `no-shipment` terminal outcome that requires `shipment_id` only for `stage_outcome: shipment` while still halting on P-003 failure. **Rev 9**: the **operative** pre-mutation branch gate and post-mutation artifact commit are both present and **ordered** — the Step Sequence Contract checklist carries a branch-gate entry ordinally **before** the deliberation step and a commit entry ordinally **after** stash archival and **before** the summary step, each backed by a step section carrying its mandatory clauses (§6.3.1). **Rev 10**: the branch-gate section's deferral clause is **categorical** over every tracked pre-gate mutation and enumerates its three classes, and the commit section carries the NUL-safe full-tree **out-of-root detection** clause positioned ordinally **before** its staging clause |
| C1 `018.010-T` | `TestDirectPushGate_CIWiringIsBlocking` | job `lint` carries the task-ID-suffixed step with no `continue-on-error` and no toggle; `ci-gate` still transitively needs `lint`; every script invoked by a `ci.yml` step has a `CODEOWNERS` owner line. **Rev 5**: the step-presence check is **YAML-structural** over `jobs.lint.steps[]` and its non-vacuity is proven (AC-C1.6). **Rev 11**: once `C1` is in the manifest's `completed` set this function **stays active from manifest state alone** — its activation is **never** derived from the presence of the step it asserts on, which is the rev-10 self-disarming defect (AC-C1.6) |
| C2 `018.011-T` | `TestDirectPushGate_FullCorpusClean` | bare full-corpus scan exits 0 with 0 findings and the three §5.3 marker-free constructs score **accept**. **Rev 5**: this runs the detector from `go test`, so it keeps executing even if the job-`lint` step is removed by a render (AC-C2.8). **Rev 11**: it activates from **manifest state**, never from C1's surface (rev 10 reused C1's predicate, so both disarmed together), and it carries the terminal-completeness check (§5.0.1, MP2) — evaluated from the manifest **before** any surface-dependent assertion, and requiring `completed` to be the exact eight-element ordered set with `terminal: true` (AC-C2.9). **Rev 12**: C2's completion also writes the **terminal completion witness** (AC-C2.10), and `gate()`'s **MP6** cross-check makes the witness and the manifest inseparable — so a manifest-only or witness-only rollback fails loudly, and C2 can no longer be skipped past MP2 by a self-consistent manifest rewind. **Rev 14**: this same function is the **parent** of the §5.0.2 point-7 fixture observation — it calls the unexported helpers `copyTrackedRepoFixture` and `runFixtureChildGoTest` and asserts on the child's parsed `go test -json` events. That block is **parent-only** and is disabled in the child by the `DIRECTPUSH_GATE_FIXTURE_CHILD=1` sentinel, so the recursion is bounded at one level. **The function count stays EIGHT** — helpers are not gate functions, have no `gateOrder` entry, and are never activated or skipped by `gate()` |

The per-surface scoping of B1/B2/B3 is what gives each of those three tasks an independent
transition; a whole-corpus assertion would only go green after all three landed and would trip
`build-feature`'s 5-attempt circuit breaker on the first two.

**Regeneration-resistant enforcement (rev 5 — R5 closure).** `.github/workflows/ci.yml` is
autoharness-generated, so the job-`lint` gate step added by C1 is a LOCAL DIVERGENCE sitting
**inside** the render replacement boundary, exactly like the five divergences already enumerated in
that file's own ledger. A render can drop it. The control is nevertheless **not** silently
disableable by regeneration, because enforcement also runs through this harness, and every link in
that path is either untouched by a render or **re-emitted** by one:

| Link | Artifact | Render behaviour |
|---|---|---|
| Detector | `scripts/check-direct-push-language.sh` | **Not generated** — hand-authored, like `check-retired-architecture.sh`, `check-write-path-precondition.sh`, `check-unignore-regression.sh`, `check-depguard-fixtures.sh` (only 5 of the 10 `scripts/*.sh` carry autoharness provenance). Survives |
| Assertion | `tests/integration/directpush_gate_test.go` | **Not generated** — no autoharness template covers `tests/**`. Survives |
| Activation state (rev 11) | `tests/integration/testdata/directpush-gate/harness-state.json` | **Not generated** — same `tests/**` boundary, and a testdata file besides. Survives. This is what makes C1/C2 activation render-independent |
| Terminal witness (rev 12) | `tests/integration/testdata/directpush-gate/terminal-witness.json` | **Not generated** — same `tests/**` boundary. Survives. Written only by C2's completion; absent at H0 and throughout `build` |
| Invocation | job `expensive` (`name: test`) runs the **full suite over `./...`** — the rendered `{{TEST_COMMAND}}` step. **Observed installed command**: step `Test (race)` → `go test -race -mod=readonly ./...` | **Generated template baseline — as a full-suite `./...` invocation, not as an exact command string (rev 12)**. The core four jobs of this file's header include the expensive test job; it carries **no** task-ID comment and is absent from the LOCAL DIVERGENCE list, so a render *re-emits* it. What the render guarantees is the **`{{TEST_COMMAND}}` substitution**, whose recorded input in this workspace is `go test ./...` — so `-race -mod=readonly` may or may not survive. R5 needs only that a full-suite `./...` invocation exists (see "Invocation provenance" below and AC-C1.7) |
| Trigger | job `expensive` runs when `changes.outputs.code == 'true'` | The `code` filter is a **denylist** — `'**'` minus `docs/**`, `.backlogit/**`, `.backlog/**`, `.autoharness/**`. `.github/workflows/**` is not excluded |

The fourth row is what makes this airtight rather than lucky: the *same* edit that removes the
job-`lint` step is itself a `.github/**` change, so it necessarily sets `code == 'true'` and
**cannot skip the job that runs the harness**.

**Invocation provenance — observed installed command vs. generated baseline (rev 12).** Revisions
5–11 asserted that `go test -race -mod=readonly ./...` *is* the generated baseline. **It is not**,
and the distinction changes what may honestly be promised:

| Kind | Value | Evidence |
|---|---|---|
| **Observed installed command** | `go test -race -mod=readonly ./...`, step `Test (race)`, job `expensive` (`name: test`) | `.github/workflows/ci.yml` as installed at this HEAD |
| **Generation template** | emits `{{TEST_COMMAND}}` for that step | `ci/ci.yml.tmpl` (gitignored source — D2; the substitution token is what the render writes) |
| **Recorded render input** | `TEST_COMMAND: go test ./...` | `.autoharness/harness-manifest.yaml`; corroborated by `.autoharness/workspace-profile.yaml` (`test.command: "go test ./..."`) |

So the flags `-race` and `-mod=readonly` are **local to the installed artifact**, not a property a
render reproduces. **What the regeneration-resistant guarantee actually requires** is narrower and
is satisfied by the baseline as recorded: *a full-suite invocation whose package pattern is `./...`,
running in a job the render re-emits and the `code` filter cannot skip* — because `./...` is what
pulls `tests/integration` (and therefore this harness) into the run. Anything stronger — "the race
detector will still be on", "`-mod=readonly` will still be set" — is **not claimed** and would be
false. Two consequences are binding:

1. Every citation of `go test -race -mod=readonly ./...` elsewhere in this plan and in the task
   contracts is an **observed-installed-command** citation, recorded so the evidence names a real
   step rather than "the `go test` step". It is **never** a claim about what a render emits.
2. A render is therefore **not** self-certifying for this control. **AC-C1.7** adds the missing
   obligation: after any `autoharness install`/`tune` regeneration of `ci.yml`, the rendered
   expensive-job test command MUST be verified to still be a full-suite `./...` invocation, and the
   LOCAL DIVERGENCE ledger must carry that obligation in writing. If a future render narrows the
   pattern (for example to `./cmd/...`), the harness stops running and **R5 reverts to an open
   residual** — which is why the check is post-render and mandatory rather than assumed. The
   narrower residual is **R23**.

Two harness functions carry the invariant, and the pair is precisely what R5 now rests on:

* `TestDirectPushGate_FullCorpusClean` (C2) re-runs the **detector** over the installed surfaces.
  Requirement **(a)** — forbidden direct-push language is absent — therefore still holds with the
  job-`lint` step gone.
* `TestDirectPushGate_CIWiringIsBlocking` (C1) asserts the job-`lint` step **still exists**.
  Requirement **(b)** — the enforcement invocation itself survives — therefore fails **red** on
  render removal instead of passing silently.

**Both functions must remain ACTIVE for that pair to mean anything (rev 11 — the P0 correction).**
Rev 10 derived C1's default-mode activation from *the presence of the job-`lint` step*, and gave C2
**the same predicate**. A render that removed the step therefore removed both functions' activation
along with it: C1 skipped instead of failing (b), C2 skipped instead of running (a), MP2 lives
inside C2 so it never executed, and `go test ./...` stayed **green** — the control disarmed itself
at exactly the moment it was supposed to fire, which is the inverse of what this whole subsection
promises. Rev 11 derives activation **only** from the harness-state manifest, an artifact in the
"Survives" column above that no assertion in this harness reads as a subject. Once `C1` and `C2` are
in `completed`, they are active **permanently**, so a render that drops the step produces a loud
`--- FAIL: TestDirectPushGate_CIWiringIsBlocking` rather than a silent skip. This is the difference
between R5 being mitigated and R5 merely appearing mitigated.

**The step-presence assertion MUST be YAML-structural, never a text search.** C1 also adds a ledger
entry to `ci.yml` that *names* `check-direct-push-language.sh` in prose, so a file-wide
`grep 'check-direct-push-language.sh' .github/workflows/ci.yml` would match that comment and keep
passing after the step itself was deleted. That is the exact self-matching failure already recorded
in `docs/compound/2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md` and warned
about at length in `ci.yml`'s own `011.002-T` step comment. The assertion parses the workflow and
walks `jobs.lint.steps[]`; its non-vacuity is proven mechanically (AC-C1.6).

**What this does NOT claim.** `CODEOWNERS` is **advisory metadata only** in this repository — its
own header records that it has no enforcement effect until code-owner review is required on `main`,
a deliberate operator action explicitly not taken here. It is therefore **not** part of this
guarantee (R6 continues to name it only as an ownership prerequisite). The guarantee is exactly and
only §1's: *a regeneration cannot **silently** reopen the gap.* Disabling the control still requires
editing a non-generated tracked file in a reviewable pull-request diff — a visible act, not a render
artifact. That narrower residual is recorded as **R12**, not hidden.

**Why not a separate workflow.** `secret-scan-history.yml` is the repo's one hand-authored,
non-generated workflow and would also survive a render — but it is `schedule`/`workflow_dispatch`
only and its header declares it **NOT A PR-REQUIRED CHECK** by design, so it cannot carry a blocking
merge gate without inverting its stated contract. `ci-topology-check.sh` is itself generated
(`ci/ci-topology-check.sh.tmpl`), so it sits inside the same boundary. The harness route adds **no
new file, no new workflow, no new CI step, no new ledger entry, and no new CODEOWNERS line** — it
only strengthens assertions in a file H0 already produces. It is the smallest sufficient mechanism.

**Why not the alternative.** Defining a first-class non-Go harness route would require amending
P-002 and P-004 in `workflow-policies.md`, Step 2 in `_ship.agent.md`, and the `harness-architect`
skill. Those last two are autoharness-generated with gitignored `.tmpl` sources (D2), so the change
would be silently reverted by the next render — the exact failure mode R4/R5 already document — and
amending P-002/P-004 is a **workspace-wide weakening of test-first** affecting every future
shipment, far outside this release unit's branch-policy contract and outside the governing
deliberation. Rejected on both counts.

#### 5.0.1 Per-task activation (rev 10, corrected rev 11, MP5 retired + witness added rev 12) — the harness lifecycle vs Ship's full-suite gate

**The defect.** Rev 4–9's §5.0 required that after H0 **all eight** functions fail, and left them
failing until their task landed. `_ship.agent.md` Step 4.3 runs the **full test suite**
(`go test ./...`) as a quality gate after *every* task's `build-feature` success. So the moment A1
went green, the seven future-task functions were still intentionally red, Step 4.3 failed, and the
task could never be marked complete — the shipment deadlocks on its second gate. The harness was
correct per-task and unexecutable per-shipment.

**The rev-10 mechanism, and why it failed (rev 11).** Rev 10 fixed the deadlock with a selector flag
plus an eight-row table of **surface predicates**: in default mode a function activated iff *its own
deliverable was observably present*. That was wrong in two independent, unanimously-confirmed ways.

1. **It was self-disarming for C1 and C2.** C1's predicate was "`jobs.lint.steps[]` contains a step
   invoking `check-direct-push-language.sh`" — which is **the exact proposition
   `TestDirectPushGate_CIWiringIsBlocking` exists to assert**. A regeneration that dropped the step
   therefore made C1 **skip** rather than fail. C2's predicate was *literally C1's predicate*, so C2
   skipped in the same breath, taking MP2 (terminal completeness, which lives inside C2) with it. A
   control whose activation is conditioned on the surface it verifies cannot detect that surface's
   removal, by construction. **C1 and C2 must remain active independently of any surface they
   assert on.**
2. **It broke the installed P-002/P-004/`harness-architect` contract.** That contract requires
   `go test ./...` **red with all generated tests failing their expected not-implemented markers**
   before the tasks become `harness-ready`. Rev 10 delivered **one** failing function plus **seven
   skips**. Reverting to rev 9's simultaneous all-eight-red satisfies P-004 but restores the Step
   4.3 deadlock. The two requirements are only irreconcilable if activation is a **static**
   predicate set; they reconcile immediately once it is a **phase**.

**The rev-11 mechanism — one selector flag and one fail-closed state manifest.** The harness
declares the same package-level flag, plus a fixed order and a tracked state file:

```go
var gateTask = flag.String("gatetask", "", "activate assertions for exactly one task ID")

// The declared dependency order. Exactly these eight IDs, exactly this order.
var gateOrder = []string{"A1", "A2", "A3", "B1", "B2", "B3", "C1", "C2"}

// The declared dependency graph. Each key maps to that task's PREREQUISITES.
//
// DIRECTION (rev 12 — stated because the stored edge TYPE reads the other way):
// backlogit records these with `dep add <item-id> <depends-on> --type blocks`, so the
// stored edge is `item -> depends-on`, i.e. DEPENDENT -> PREREQUISITE, even though its
// type label is the word "blocks". `018.005-T -> 018.004-T (blocks)` therefore means
// "018.005-T DEPENDS ON 018.004-T", NOT "018.004-T blocks-list contains 018.005-T".
// gateDeps mirrors that stored direction exactly: gateDeps["A2"] = {"A1"} means A2's
// prerequisite is A1. Reading the label as prerequisite -> dependent inverts the graph
// and would make the dependency-closure check accept exactly the sets it must reject.
//
// B1/B2/B3 are PARALLEL (each depends only on A3), so `completed` is a
// dependency-closed, order-preserving SUBSEQUENCE of gateOrder -- not a prefix.
var gateDeps = map[string][]string{
    "A1": nil, "A2": {"A1"}, "A3": {"A2"},
    "B1": {"A3"}, "B2": {"A3"}, "B3": {"A3"},
    "C1": {"B1", "B2", "B3"}, "C2": {"C1"},
}

// tests/integration/testdata/directpush-gate/harness-state.json
type harnessState struct {
    Phase     string   `json:"phase"`     // "red" | "build" | "terminal"
    Completed []string `json:"completed"` // dependency-closed subsequence of gateOrder
    Terminal  bool     `json:"terminal"`
}

// tests/integration/testdata/directpush-gate/terminal-witness.json (rev 12).
// ABSENT at H0 and throughout `build`; written ONLY by C2's completion.
// Fixed schema, fixed version, contract identity, exact terminal set, and a digest
// over that set. NO timestamps and NO environment-specific data, so the file is
// byte-reproducible from the contract alone and carries no run-dependent state.
type terminalWitness struct {
    Schema    string   `json:"schema"`    // exactly "directpush-gate/terminal-witness"
    Version   int      `json:"version"`   // exactly 1
    Contract  string   `json:"contract"`  // exactly "017-S/018-F"
    Completed []string `json:"completed"` // exactly gateOrder
    Digest    string   `json:"digest"`    // lowercase hex SHA-256 of strings.Join(Completed, "\n")
}
```

`gate(t, "B3")` is the **first statement** of every test function. It runs in **two explicitly
ordered stages (rev 13)**: a **mode-independent integrity gate**, then a **mode-dependent activation
resolution**. The stages are separate because they answer different questions, and collapsing them
is what produced the rev-12 defect: *is the control state coherent?* must be answered for every
invocation, while *should this test body run?* depends on how the harness was invoked.

**Stage 1 — integrity gate (mode-independent; runs in EVERY mode, before any activation decision).**
Every row below **fails** and never skips. Nothing in this stage activates anything:

| Condition | Outcome |
|---|---|
| Manifest missing, unreadable, or malformed JSON | **fail** (never skip) — MP1 |
| Manifest invalid per MP1 (unknown/duplicate/out-of-order ID, non-subsequence set, unmet prerequisite, inconsistent `phase`/`completed`/`terminal` triple) | **fail** (never skip) — MP1 |
| Witness present but invalid per MP6 (schema/version/contract mismatch, `completed` ≠ `gateOrder`, digest mismatch, malformed JSON) | **fail** (never skip) — MP6 |
| Witness present **and** manifest **not** terminal | **fail** (never skip) — MP6 |
| Manifest terminal **and** witness absent | **fail** (never skip) — MP6 |
| `-gatetask` names an ID absent from `gateOrder`, or the function's own ID is absent from `gateOrder` | **fail** (never skip) — MP0/MP3 |

**Stage 2 — activation resolution (mode-dependent). A valid `-gatetask` selector has activation
precedence over every default-mode rule, including the witness rule.** The **Mode** column is the
first discriminator and it is exhaustive: an invocation is either **Targeted** or **Default**, never
both, and **no Default row may be applied to a Targeted invocation**. Within a mode, the first
matching row governs — except the final witness row, which is **additive reinforcement** of the
`phase: terminal` row rather than an alternative to it:

| Mode | Condition | Outcome |
|---|---|---|
| **Targeted** (valid `-gatetask` supplied) | `-gatetask` names **this** function's task ID | **activate** — the real assertion runs, in **every** phase, regardless of manifest phase, witness presence, or surface state |
| **Targeted** (valid `-gatetask` supplied) | `-gatetask` names a **different** valid task ID | `t.Skip` — a **selector** skip. Applies in **every** phase **including `terminal`**, and **regardless of witness presence** |
| **Default** (`-gatetask` empty) | `phase: red` | **activate** — *all eight* activate; no function skips |
| **Default** (`-gatetask` empty) | `phase: build` | **activate** iff this task's ID is in `completed`; otherwise `t.Skip` naming the ID, the phase, and the current `completed` set |
| **Default** (`-gatetask` empty) | `phase: terminal` | **activate** — *all eight* activate; no activation skip remains |
| **Default** (`-gatetask` empty) | Witness present and valid (⇒ manifest terminal, by Stage 1) | `C2` is **forced active** — belt-and-braces over the `terminal` row above, so a defect in the phase path cannot un-force it |

**Why selector precedence does not reopen the MP6 hole (rev 13).** MP6 has two halves, and only one
of them is an activation rule. Its **integrity** half — *witness present ⇒ manifest terminal; manifest
terminal ⇒ valid witness present* — lives in **Stage 1** and is **mode-independent**, so a targeted
invocation against a rewound or mismatched state **still fails**, loudly, before Stage 2 is reached.
Its **forcing** half is an activation rule and is therefore **default-mode only**. The hole MP6 was
added to close — a self-consistent manifest rewind letting C2 skip past MP2 — is closed by the
**integrity** half, which fires in both modes; the forcing half is redundant reinforcement of the
`terminal` row, not the mechanism. Scoping it to default mode consequently subtracts **no**
detection, while removing the contradiction with MP3, AC-C2.7, AC-C2.9, and AC-C2.11.

**What rev 12 got wrong, stated so it cannot recur.** The rev-12 table was a single flat list whose
witness row — "`C2` is forced active; **all eight** activate" — appeared **before** the `-gatetask`
rows. Read sequentially, as a normative table must be, a **targeted** run in terminal state
activated all eight bodies: `-gatetask=C1` would have executed C2, B3, A2 and the rest, contradicting
MP3's "selector skips of the seven non-selected functions are expected in targeted mode in **every**
phase", AC-C2.7's rev-12 qualification, AC-C2.9's targeted clause, and AC-C2.11's evidence rules.
**Witness integrity may fail a targeted invocation; witness presence may never activate a test body
the operator did not select.**

There is **no substantive-surface predicate anywhere in either table**. That is the rev-11
correction, preserved intact: activation is a function of **manifest state, witness state, and the
selector only**, so no assertion's activation can be switched off by mutating the thing it asserts
on. The manifest and witness are **dedicated control-state inputs**, expressly permitted here per
§5.0 "Control-state inputs are expressly exempt"; they are not substantive surfaces. **Rev 12**
added the symmetric rule for the failure direction: no row of either table, and no gate-level check,
may *fail* on the basis of a **substantive surface** the harness asserts on either (that was MP5's
defect — see "MP5 — RETIRED at rev 12").

**Valid states, defined exhaustively (rev 12).** Exactly three combinations are legal. Anything not
in this table is invalid and **fails closed** in `gate()` on every call:

| Phase | `completed` | `terminal` | Witness | Meaning |
|---|---|---|---|---|
| `red` | **empty** | `false` | **absent** | H0. All eight active, all eight failing their own marker |
| `build` | **non-empty**, order-preserving, dependency-closed, **proper subset** of `gateOrder` | `false` | **absent** | Mid-shipment. Completed set active; not-yet-started tasks skip |
| `terminal` | **exactly** `["A1","A2","A3","B1","B2","B3","C1","C2"]` | `true` | **present and valid** | Closed. All eight active, no activation skip anywhere |

Every other combination is **rejected**, including each of these named cases, which are the ones a
partial write, a rollback, or a hand-edit actually produces:

* `red` with `terminal: true` — **rejected** (the red/terminal contradiction).
* `red` with a non-empty `completed` — **rejected**.
* `build` with an empty `completed` — **rejected** (that is `red`'s shape, not `build`'s).
* `build` with all eight IDs — **rejected** (a complete set is `terminal`, never `build`).
* `build` with `terminal: true` — **rejected**.
* `terminal` with a proper subset of `gateOrder` — **rejected**.
* `terminal` with `terminal: false` — **rejected**.
* `terminal` with the witness absent — **rejected** (MP6).
* any phase other than `red`/`build`/`terminal` — **rejected**.
* a witness present alongside `red` or `build` — **rejected** (MP6).

The three-row table is the normative definition; the bullet list enumerates the rejections that
matter operationally and is not a second, weaker rule.

**Zero skips — what it does and does not claim (rev 12; precedence stated rev 13).** "No skip
remains anywhere" is a statement about **default mode** in `red` and `terminal` phase: with no
`-gatetask` given, all eight gate functions activate and none skips. It is **not** a claim about
targeted mode. `-gatetask={ID}` is a selector with **activation precedence over every default-mode
rule, including the witness-forcing rule**: by design it runs exactly the selected function's real
assertion and **skips the other seven**, in every phase, **including `terminal`, and regardless of
whether a valid terminal witness is present**. Those are **selector** skips, and they are correct.
What targeted mode may never do is skip **the test it selected** — if `-gatetask=C1` causes
`TestDirectPushGate_CIWiringIsBlocking` itself to skip, that is a defect and a failure of AC-C1.6.
Witness and manifest **integrity** are unaffected by the selector and still fail a targeted
invocation (Stage 1). Evidence that cites "zero `--- SKIP:` lines" therefore MUST be taken from a
**default-mode** run; a targeted run showing seven skips is expected and proves nothing either way.

**Why a Go test flag rather than an environment variable.** A flag is invoked identically from
`bash`, `pwsh`, and `cmd`; `VAR=value go test …` is a POSIX-shell-only construct and
`$env:VAR=…` is PowerShell-only, so an env-var selector would make `harness_cmd` shell-specific.
**Verified on this platform**: the flag name must be **dot-free** — `go test . -run '^T$'
-gatetask=A1` parses and reaches the test binary unquoted, whereas a dotted name such as
`-directpush.task=A1` is split by PowerShell's tokenizer and rejected by the binary as
`flag provided but not defined: -directpush`. `-gatetask` is therefore the portable form, and the
plan names it explicitly rather than leaving it to the implementer.

**The fixture-child sentinel (rev 14) — an environment variable, and why that does not contradict
the paragraph above.** The §5.0.2 point-7 observation requires a **child `go test` process** run
against a copied repository fixture (the mechanism is specified there). The parent marks that child
with a single environment variable, `DIRECTPUSH_GATE_FIXTURE_CHILD=1`, and the rationale above is
untouched because the two are different kinds of input:

| Property | `-gatetask` selector | `DIRECTPUSH_GATE_FIXTURE_CHILD` sentinel |
|---|---|---|
| Who supplies it | A **human or `harness_cmd`**, typed into a shell | The **parent test process**, through `exec.Cmd.Env` |
| Shell portability | Load-bearing — must parse under `bash`, `pwsh`, `cmd` | **Not applicable** — no shell is involved; nothing in `harness_cmd`, §8, or any documented command sets it |
| Effect on activation | **Selects** which function runs; skips the other seven | **None.** It is **not** a selector and has **no** row in either `gate()` table |
| Effect on integrity | None (Stage 1 is mode-independent) | **None** — MP1 and MP6 run in the child exactly as in the parent |

The sentinel's **only** effect is to disable the **parent-only fixture-spawn block** inside
`TestDirectPushGate_FullCorpusClean`, which is what makes the mechanism **non-recursive**: the child
cannot spawn a grandchild. **Rev 15 — this is a PRECONDITION, not a detector.** Recursion is
**prevented structurally** on the parent side (the block is entered only when the sentinel is
**unset**, so depth is bounded at exactly one level by construction) and **bounded** by the child's
explicit `context.WithTimeout`. It is **not** established by any child-side check that the sentinel
"is already set when the block is reached" — that condition is **unsatisfiable** given this very
guard, and §5.0.2 item 5 withdraws it as evidence. Three prohibitions are normative, because each of
them would convert the guard into the very defect class this plan exists to exclude:

* The sentinel **MUST NOT skip C2**. A sentinel-set run executes C2's **full substantive
  assertions** — the default-mode full-corpus scan, MP2 terminal completeness, and the regeneration
  assertions — exactly as an unset run does. A `t.Skip` keyed on the sentinel would make point 7's
  "C2 executed" evidence unobtainable **by construction**, which is the rev-13 vacuity defect
  wearing a different hat.
* The sentinel **MUST NOT bypass MP1 or MP6**, nor any other `gate()` Stage 1 check. Integrity is
  mode-independent and is **environment-independent** too; a fixture whose manifest and witness
  disagree must fail the child run rather than be waved through.
* The sentinel **MUST NOT be readable as a substantive-surface predicate**. It is a
  **process-topology** input describing *which process this is*, not a proposition any assertion
  makes about the repository, so it neither activates nor deactivates a test body and adds nothing
  to the §5.0 control-state exemption, whose closed membership (manifest + witness) is unchanged.

An unset sentinel is the **normal** case and is never an error: every ordinary local or CI run of
the suite has it unset, spawns exactly one child from the parent block **when the real tree is in
`terminal` phase**, and is bounded by that block's own timeout. In `red` phase C2 fails on its own
marker before reaching the block; in `build` phase C2 is not active in default mode at all — so the
child is spawned only in the one state in which the observation means anything (§5.0.2 point 7,
item 4).

**The terminal completion witness (rev 12) — what it closes, and what it does not.** MP2 (terminal
completeness) lives **inside C2**, and C2's default-mode activation is read from the **same
manifest** MP2 validates. That circularity is the defect: an edit that rewound `phase`,
`completed`, and `terminal` **together** produced a perfectly self-consistent `build` state in which
MP1 passed, C2 **skipped as a not-yet-started task**, and MP2 therefore never ran. The terminal
claim could be withdrawn without any check firing.

The fix is one small, tracked, non-generated file **independent of the main manifest**, at
`tests/integration/testdata/directpush-gate/terminal-witness.json`:

```json
{
  "schema": "directpush-gate/terminal-witness",
  "version": 1,
  "contract": "017-S/018-F",
  "completed": ["A1","A2","A3","B1","B2","B3","C1","C2"],
  "digest": "<lowercase hex SHA-256 of the eight IDs joined with \n, no trailing newline>"
}
```

It is deliberately **minimal**: a fixed schema string, a fixed integer version, the contract
identifier, the exact terminal task set, and a digest over that set. **No timestamps, no paths, no
host, user, branch, commit, or run identifiers** — nothing environment-specific — so the file is
byte-reproducible from the contract alone, produces no diff churn, and cannot drift between a local
run and CI. The digest is computed with stdlib `crypto/sha256` over
`strings.Join(completed, "\n")`; it is redundant with the array by construction, which is the point:
it turns a one-character edit of the array into a detectable inconsistency rather than a silent one.

**Write order at C2's completion is normative**: write the **witness first**, then **atomically
replace** the manifest with the exact terminal state. Tests are green only when the two **agree**.
The order matters because it decides what a crash between the two writes leaves behind: witness-first
leaves *witness present, manifest not terminal*, which MP6 rejects **loudly**. It also guarantees the
manifest — the file `gate()` reads for activation on every call — is never terminal without its
witness.

**MP6's rules run in `gate()`, before any activation decision**, which is what stops C2 being skipped
past its own check:

* Witness **present** ⇒ the manifest MUST be terminal (exact eight-element set, `terminal: true`,
  `phase: terminal`), **and in default mode `C2` is forced active** regardless of what the manifest
  says. Forcing is stated explicitly rather than left to follow from terminal-phase activation, so a
  defect in the phase path cannot silently un-force it. **Rev 13**: the forcing clause is an
  **activation** rule and is therefore **default-mode only**; the *integrity* requirement in the same
  sentence is **mode-independent** and fails a targeted invocation too. A targeted run selecting some
  other task still skips `C2` — a selector skip — while any witness/manifest disagreement has already
  failed the invocation in Stage 1.
* Manifest **terminal** ⇒ a **valid** witness MUST be present.
* Witness **absent** and manifest **not** terminal ⇒ normal H0/`build` operation. **H0 absence is
  valid**, and so is absence throughout `build`; neither is an error.
* A witness that is **missing where terminal evidence exists**, **malformed**, or **mismatched**
  (schema, version, contract, set, or digest) ⇒ **fail closed**.

Consequently: **rolling back only the manifest fails** (witness still present), and **rolling back
only the witness fails** (manifest still terminal). Both are loud.

**The residual, stated honestly and not overclaimed.** Deleting or rolling back **both** the witness
**and** the manifest in one change produces a state that is, at runtime, **indistinguishable from a
legitimate pre-C2 `build` state** — because that is exactly what it is, as far as a stateless
validator can see. No check here detects it, and **this plan does not claim otherwise**: a validator
that reads only current files cannot detect every coordinated rollback, because it has no history to
compare against. What catches it is that both files are **tracked and non-generated**, so the
rollback appears as a reviewable diff in a pull request — caught by **diff and review**, not by
runtime history. That is the same visible-edit boundary as R12/R21, and it is recorded as **R22**
rather than mitigated on paper.

**Why JSON, and why these paths.** `encoding/json` is stdlib, so the harness parses its own state
with no new dependency (a `.yaml` state file would need one, and a bespoke line format would need a
hand-written parser to fail closed on). The path
`tests/integration/testdata/directpush-gate/harness-state.json` is **release-specific** — it is
namespaced to this gate, so it cannot collide with, or be mistaken for, shared harness
infrastructure — and `testdata/` is the Go-idiomatic location the toolchain already excludes from
builds. Critically, it is **outside every workflow render boundary**: no autoharness template covers
`tests/**` (§5.0 render table), so unlike `.github/workflows/ci.yml` it cannot be rewritten by a
regeneration. The rev-12 terminal witness sits in the **same release-specific directory** for
exactly the same four reasons, and is a **separate file** rather than a field of the manifest
precisely so that a single-file rewrite cannot move both halves of the terminal claim at once.

**Red-phase markers.** In `red` phase every function activates and fails; each failure carries its
**own** task-specific marker so the 8/8 evidence is literal and greppable rather than aggregate:

| Task | Expected red-phase marker |
|---|---|
| A1 | `not implemented: A1 direct-push fixture corpus` |
| A2 | `not implemented: A2 self-test fixture-suite driver` |
| A3 | `not implemented: A3 direct-push detector implementation` |
| B1 | `not implemented: B1 orchestrator Step 1.5 surface` |
| B2 | `not implemented: B2 workflow-policies P-010 surface` |
| B3 | `not implemented: B3 stage agent Role Boundary surface` |
| C1 | `not implemented: C1 CI wiring and ownership` |
| C2 | `not implemented: C2 full-corpus coupling evidence` |

These are the **test** markers. They are deliberately distinct from the **script** stub's single
`not implemented: direct-push detector` marker (§5.0 deliverable 2), which is the shell analogue of
`harness-architect`'s production stub and is unchanged. H0's recorded evidence is the literal count:
eight `--- FAIL:` lines, each carrying its own marker from the column above.

**The per-task lifecycle.** Each task's boundary is unchanged in shape — a targeted red, then an
implementation, then a targeted green — with one added completion obligation:

1. **Targeted red, before implementation.** Ship runs the task's `harness_cmd`, which carries
   `-gatetask={ID}`. Targeted mode activates unconditionally, so the function runs its **real**
   assertion and fails for that task's **expected** reason. This is the P-004 per-task red, and it
   is observable in every phase.
2. **Implementation.** The task builds its surface until the substantive assertion passes.
3. **Completion — atomic, ordered, and in the same task.** As part of the *same* task's completion,
   the task inserts its ID into `completed` at its **`gateOrder` position**, so the set stays an
   order-preserving subsequence. The insertion and the surface work land together; the manifest is
   never advanced ahead of the assertion it unlocks. **A1 additionally performs the one-way `phase`
   transition `red` → `build`.** There is **no reverse transition**: `build` → `red` and
   `terminal` → `build` are impossible states and fail closed (MP1).
4. **Green by assertion, never by bookkeeping.** Writing an ID into `completed` does **not** make a
   test pass. Targeted mode ignores the manifest entirely for activation and always evaluates the
   real assertion; default mode evaluates the real assertion for every ID in `completed`. A task that
   appended its ID without doing the work fails on its own assertion, immediately and loudly.
   **Rev 14 — the rule stated precisely, with its one deliberate exception.** For **A1 through C1**
   the rule is absolute: the completion state is **bookkeeping only**, it is never itself a subject
   of the task's substantive assertion, and it can therefore **never on its own** turn that
   assertion green. **C2 is the deliberate exception, and it is an exception of SUBJECT, not of
   rigour.** Finalizing the terminal state — writing a valid `terminal-witness.json` and atomically
   replacing the manifest with the exact eight-element terminal triple — *is itself part of C2's
   substantive verification contract* (MP2 + AC-C2.9 + AC-C2.10), because the property C2 exists to
   verify includes *the lifecycle is closed and the closure is carried by two agreeing tracked
   files*. So for C2 alone, a control-state write is legitimately load-bearing on its own verdict.
   What does **not** follow, and is expressly forbidden: C2 **cannot pass from bookkeeping alone**.
   It must still execute and pass its **full substantive assertions** — the bare full-corpus scan
   at zero findings (AC-C2.1), the `--self-test` fixture and manifest/witness mutation proofs
   (AC-C2.2), the marker-free-construct proof (AC-C2.3), and the regeneration-resistance evidence
   (AC-C2.8) — and MP2 is evaluated from manifest state **before** any surface-dependent assertion
   so a surface failure can never be masked by the completeness verdict. A C2 that finalized the
   witness and manifest but left the corpus dirty or the regeneration evidence unmet **fails**.
5. **C2 closes the lifecycle.** C2's completion requires `completed` to become the **exact
   eight-element ordered set** and `terminal` to become `true` **atomically**, with `phase` moving
   to `terminal`. **Rev 12**: it also writes the **terminal completion witness** — **witness first,
   then the atomic manifest replacement** (AC-C2.10). From that point the default suite activates all
   eight substantive assertions and **no activation skip exists anywhere in the harness**.

Between step 3 and the next task, `go test ./...` runs the completed set and skips only
not-yet-started tasks — which is exactly what keeps Ship's Step 4.3 green after every task.

**Scope of the per-task manifest edit (rev 11) — stated once, authoritatively.** Every one of the
eight tasks edits **one additional file**, `tests/integration/testdata/directpush-gate/harness-state.json`,
by **one line**: inserting its own ID into `completed` at its `gateOrder` position (and, for A1 and
C2 only, flipping `phase` and — for C2 — `terminal`). This is deliberately **not** a second skill
domain and **not** a new deliverable. It is a shared **completion record** in a file H0 authored,
written as part of the same task completion that makes the task's substantive assertion pass. It
adds no design decision, no test, and no reviewable logic; it is a state transition whose legal
values are fully enumerated above and mechanically validated by MP1. **No task's `size` or
`complexity` changes**, and every task remains inside the 2-hour rule: the per-task file count rises
by one testdata file while the per-task *skill domain* count stays at one. Each task's backlog entry
states this explicitly so an executor does not read it as scope creep or as licence to restructure
the harness.

**Ordering rule — subsequence, not prefix, because B1/B2/B3 are parallel.** The declared dependency
graph is `A1→A2→A3→{B1,B2,B3}→C1→C2`: B1, B2, and B3 each depend only on A3 and have **no edge among
themselves** (verified against the `blocks` edges recorded in `.backlogit/`). Ship may therefore
complete them in any order. `completed` is accordingly validated as an **order-preserving
subsequence of `gateOrder` that is closed under `gateDeps`** — not as a prefix:

* **Order-preserving** — entries appear in `gateOrder` sequence, so a permuted list (`["A2","A1"]`)
  **fails**. This is the **reordering** detector.
* **Dependency-closed** — every prerequisite of every member is also a member, so `["A1","B2"]`
  **fails** (B2's prerequisite A3 is absent). This is the **skipped prerequisite** detector, and it
  is also what catches a task that did its work but forgot its insertion: the *next* task's
  completion produces a non-closed set and fails immediately, rather than leaving a silent gap until
  C2.
* `["A1","A2","A3","B2"]` is **valid** — B2 landing before B1 is legal, because the graph permits it.

A **strict-prefix** rule was **rejected**: it would have forced an ordering the dependency graph does
not declare, and would have deadlocked a legal execution order (B2 first) in which B2's surface
exists while its ID is not yet in `completed`. **Rev 12 note**: the rev-11 text finished that
sentence with "— a state MP5 would then correctly flag as a rollback", which was itself the MP5
defect in miniature. A legal mid-shipment state in which an incomplete task's surface is present is
**not** a rollback and must not fail anything; MP5 is **RETIRED** (below) and the clause is
withdrawn. The subsequence-plus-closure rule admits exactly the orders the graph admits, and no
others. Note also that `completed` is never described as a "prefix" anywhere in the normative
algorithm — it is a **dependency-closed, order-preserving subsequence**, and in `build` phase
specifically a **proper subset** of `gateOrder`.

The eight insertions, shown in the canonical `gateOrder` sequence (the B-block rows may land in any
order among themselves):

| After task | `phase` | `completed` |
|---|---|---|
| *(H0, before any task)* | `red` | `[]` |
| A1 `018.004-T` | `build` *(one-way transition)* | `["A1"]` |
| A2 `018.005-T` | `build` | `["A1","A2"]` |
| A3 `018.006-T` | `build` | `["A1","A2","A3"]` |
| B1 `018.007-T` | `build` | adds `"B1"` — any of the three may be first |
| B2 `018.008-T` | `build` | adds `"B2"` — any of the three may be first |
| B3 `018.009-T` | `build` | adds `"B3"` — any of the three may be first |
| C1 `018.010-T` | `build` | `["A1","A2","A3","B1","B2","B3","C1"]` — C1 requires all three |
| C2 `018.011-T` | `terminal` *(with `terminal: true`, atomically)* | all eight |

**Mutation-proof checks.** The risk this mechanism introduces is a manifest that lies. **Six active
checks** close it — **MP0, MP1, MP2, MP3, MP4, and MP6** — and they are the whole of the added
machinery. **MP5 is RETIRED at rev 12** (see below); the numbering is **not** compacted, because
renumbering five checks across the plan, eight task contracts, the feature DoD, the deliberation,
and the memory record would be pure churn and would silently rewrite the identifiers under which
the retired check was reviewed. The rev-11 text called this list "five checks" while enumerating
six; that miscount is corrected here.

* **MP0 — order and graph integrity.** `gateOrder` has exactly **eight** IDs, exactly the eight task
  IDs of the §5.0 map, in the declared sequence; `gateDeps` has a row for each, and each row lists
  that task's **prerequisites** — the `depends-on` targets of the edges recorded in `.backlogit/`
  for this shipment (stored as `item → depends-on` under the type label `blocks`; see the direction
  note in the `gateDeps` declaration above). Every function's own ID must
  resolve to a position in `gateOrder`; a function whose ID is missing **fails**. Carried by A1,
  which runs in every default-mode run at every phase (it is a prerequisite of every other task, so
  it is in every non-empty dependency-closed set, and in `red` everything runs).
* **MP1 — manifest validity, fail-closed, enforced on every call.** `gate()` itself rejects: a
  missing file; unreadable or malformed JSON; a `phase` outside `{red, build, terminal}`; an ID in
  `completed` that is not in `gateOrder`; a **duplicate** ID; a set that is not an **order-preserving
  subsequence** of `gateOrder` — the **reordering** detector; a set that is not **dependency-closed**
  under `gateDeps` — the **skipped prerequisite** detector; and any triple outside the three legal
  states enumerated exhaustively above (`red` ⟺ `completed` empty ∧ `terminal: false`; `build` ⟺
  `completed` a non-empty dependency-closed **proper subset** ∧ `terminal: false`; `terminal` ⟺
  `completed` the exact eight ∧ `terminal: true`). `red` with `terminal: true`, `build` with all
  eight, `terminal` with a proper subset, and every other combination are **rejected**. Every one of
  these **fails**; none skips. Because it runs inside `gate()`, a corrupt manifest fails **all
  eight** functions, not one.
* **MP2 — terminal completeness, evaluated from manifest state FIRST.** Inside
  `TestDirectPushGate_FullCorpusClean` (C2), and **before any surface-dependent assertion in that
  function**, the manifest must show `completed` == the exact eight-element ordered set,
  `terminal: true`, and `phase: terminal`; then the real full-corpus assertions run. Ordering the
  manifest check first is deliberate: a surface-dependent assertion placed ahead of it could fail or
  error for an unrelated reason and mask the completeness verdict. Because C2 is itself in
  `completed` at the end state, C2 **always activates** in default mode from its own completion
  onward — so a post-landing regression in *any* surface is a **failure**, never a skip. **Rev 12**:
  MP2 alone was not sufficient, because it lives *inside* C2 and C2's activation was read from the
  same manifest MP2 validates — a self-consistent rewind skipped C2 and MP2 with it. **MP6 closes
  that**, and MP2 keeps its role as the in-function completeness assertion.
* **MP3 — visible skips, strict selector.** Every skip is a `t.Skip` naming the task ID, the phase,
  and the current `completed` set, so the state is greppable in `go test -v` rather than silent.
  An unknown `-gatetask` value **fails** rather than skipping, so a typo in a `harness_cmd` cannot
  turn the suite green. **Activation** skips are structurally impossible in `red` and `terminal`
  phase; **selector** skips of the seven non-selected functions are expected in targeted mode in
  every phase, and the selected function itself may never skip (see the zero-skip qualification
  above). **Rev 13**: this is now guaranteed by construction rather than by convention — the
  selector holds **activation precedence** in `gate()`'s Stage 2, so a valid terminal witness can no
  longer activate the seven unselected bodies in a targeted run.
* **MP4 — the state files are themselves part of the mutation proof (AC-C2.2, AC-C2.9, AC-C2.10).**
  Against a **scratch copy** under `t.TempDir()` — never the committed tree, same no-dirty-tree-edit
  rule as AC-A2.4/AC-C1.6 — the proof asserts that `gate()` **fails** for each of: the manifest
  **deleted**; the manifest **malformed** (invalid JSON, and an unknown `phase` value); a
  **rollback** (`completed` truncated, or `terminal`/`phase` left inconsistent with it); a
  **reordering** (`completed` permuted out of `gateOrder` sequence); a **skipped prerequisite** (a
  member whose `gateDeps` prerequisite is absent); a **duplicate** ID; and an **unknown** ID.
  **Rev 12 adds the three witness cases**: **witness-only** (witness present, manifest not
  terminal); **terminal-manifest-only** (manifest terminal, witness absent); and a **malformed or
  mismatched witness** (invalid JSON, wrong schema/version/contract, `completed` ≠ `gateOrder`, or a
  digest that does not match its own array). Each case asserts an
  observed **failure**, never a skip — the same executed-non-vacuity discipline AC-A2.4 and AC-C1.6
  already use.
* **MP5 — RETIRED at rev 12. REMOVED, not renumbered. MUST NOT be implemented.** The rev-11 check
  read every *incomplete* task's surface probe and **failed** when one was present, on the theory
  that a present surface for an incomplete task signals rollback or tamper. It does not. C2's
  probe was C1's old predicate — **the job-`lint` step, which C1 itself creates** — so in the
  entirely legitimate window after C1 completes and before C2 starts, the probe read `true`, MP5
  fired inside A1, and the **default full suite went red**: precisely the Ship Step 4.3 deadlock
  §5.0.1 exists to remove. The root error is that MP5 derived a **failure** from a surface the
  harness asserts on, which is the same prohibited edge rev 11 removed from *activation*, merely
  pointed the other way. **What replaces it, and why nothing is lost**: rollback of a **completed**
  task's surface is already caught by that task's **own substantive assertion**, which default mode
  runs from its completion onward (that is exactly what makes MP2's "a post-landing regression in
  any surface is a failure" true); rollback of the **manifest or the witness** is caught by MP1 and
  MP6; and an **incomplete** task having prerequisites or surface fragments present is a **legal**
  mid-shipment state that must not fail anything. The "probe non-vacuity" obligation attached to
  MP5 is withdrawn with it — there are no probes left to prove non-vacuous. **No normative
  algorithm in this plan calls MP5 active**, and any reference to it elsewhere is historical.
* **MP6 — witness/manifest cross-agreement, in `gate()`, before activation (rev 12).** The terminal
  claim is carried by **two independent tracked files** that must agree. `gate()` enforces, on every
  call and **before any activation decision is taken**: a present witness must be **valid** (schema
  `directpush-gate/terminal-witness`, version `1`, contract `017-S/018-F`, `completed` exactly
  `gateOrder`, digest matching its own array) — otherwise **fail**; a present **valid** witness
  requires a **terminal** manifest — otherwise **fail**; a **terminal** manifest requires a present
  valid witness — otherwise **fail**; and a witness **absent** with a non-terminal manifest is
  **valid**, which is the H0 and `build` state. In **default mode** a present valid witness
  additionally **forces `C2` active**. **Rev 13 separates MP6's two halves explicitly**: the
  **integrity** half above is **mode-independent** and lives in `gate()`'s Stage 1, so it fails a
  targeted invocation exactly as it fails a default one; the **forcing** half is an **activation**
  rule and is **default-mode only**, because a targeted run must execute exactly the selected test
  (MP3). Because the integrity half precedes activation in every mode, **C2 can no longer be skipped
  past MP2** by a self-consistent manifest rewind: the surviving witness fails the whole suite on
  the mismatch, and in default mode it also forces C2 active. Scoping the forcing half to default
  mode therefore subtracts no detection. Treatment is **fail-closed** for missing, malformed, or
  mismatched state **wherever terminal or witness evidence exists**; **H0 absence is valid** and is
  not an error. The coordinated both-files rollback remains undetectable at runtime and is recorded
  as **R22**, not claimed as closed.

**What this preserves, point by point.**

| Requirement | How it holds |
|---|---|
| (a) harness-architect produces the **complete** file before implementation | Unchanged — H0 still emits all eight functions, the helper, `gateOrder`, and the state manifest. No task creates a test |
| (b) the installed P-002/P-004 red phase is satisfied **literally** | In `red` phase all eight activate and all eight fail with their own expected `not implemented:` marker. `go vet ./...` = 0, `go test ./...` ≠ 0, 8/8. Only then may the tasks be labelled `harness-ready` |
| (c) every task has an observed targeted **red→green** transition before its implementation | Each `harness_cmd` carries `-gatetask={ID}`, which activates unconditionally in every phase, so the function is red at claim time and green after implementation. Ship's `build-feature` boundary is unchanged |
| (d) future-task expected-red cases do not fail the **default** suite after A1 | A1's completion moves `phase` to `build`; default mode then runs the completed set and skips only not-yet-started tasks, so Step 4.3's `go test ./...` is green between tasks |
| (e) the final `go test ./...` runs all eight **real** assertions green | At `terminal` every ID is in `completed`, so all eight activate and no skip remains; MP2 asserts exactly that, from manifest state, before any surface check |
| (f) C1/C2 cannot be disarmed by the surface they verify | Activation is manifest-only. A render removing the job-`lint` step leaves both **active in default mode**, so C1 fails loudly on (b) while C2 keeps executing the detector for (a) — the §5.0.2 point-7 observation, taken **unselected** |
| (g) a legitimate mid-shipment state never fails the default suite (rev 12) | No check reads an **incomplete** task's surface. MP5, which did, is **RETIRED**; the post-C1/pre-C2 window — in which C2's old probe (the lint step) is present because **C1 created it** — is now simply a valid `build` state |
| (h) the terminal claim cannot be withdrawn silently (rev 12) | The claim is carried by **two** tracked files that MP6 requires to agree, checked in `gate()` **before** activation, so a manifest-only rewind no longer skips C2 past MP2. Rolling back **both** is the recorded residual **R22**, caught by diff review — not claimed as runtime-detectable |

**Why the release-specific manifest satisfies the installed contract literally — and changes none of
it.** This revision plans **no** edit to `.github/skills/harness-architect/SKILL.md`, to P-002 or
P-004 in `workflow-policies.md`, or to `_ship.agent.md` Step 2 or Step 4.3. Each foundational
contract is satisfied **as written**:

* **`harness-architect`'s deliverable scope.** The skill emits a Go `_test.go` harness plus the
  production stubs its tests need. H0 already emits a **non-Go** deliverable — the structural
  `check-direct-push-language.sh` stub (§5.0 deliverable 2) — as the shell analogue of a
  `panic("not implemented")` stub. A `testdata/` JSON file initialized to a literal constant is
  strictly **smaller** than that: it is harness-internal state, written once, never read by
  production code, and located inside the harness's own `tests/integration/` tree. It adds no route,
  no policy, and no label semantics.
* **P-004's red-phase precondition.** Met **literally**, which rev 10 did not do: `go vet ./...`
  exits 0, `go test ./...` exits non-zero, and **all eight** generated tests fail with expected
  not-implemented markers. The `harness-ready` label is applied by `harness-architect` from that
  evidence, by the normal mechanism. Stage applies no label and asks for no exemption.
* **P-002 / Ship Step 2.** Unchanged. Step 2 partitions the queue on `harness-ready`, invokes
  `harness-architect` for tasks lacking it, and halts unless all carry it. That flow runs exactly as
  installed.
* **Ship Step 4.3.** Unchanged and **not weakened**. It still runs the full `go test ./...` after
  every task. What changed is the harness's own phase, not the gate's scope — rev 10's rejected
  alternative of narrowing Step 4.3 to `go test ./tests/integration -run …` stays rejected (§11).
* **Backlog independence.** The manifest is **implementation/harness state**, not backlog status. It
  is never derived from, compared against, or synchronized with `.backlogit/` item statuses, which
  are mutable by agents and operators outside the release unit. A status edit therefore cannot
  activate or deactivate any assertion.

**Harness-ready semantics stay honest.** `harness-ready` continues to mean exactly what
`harness-architect` verifies: the file compiles (`go vet ./...` = 0) and the default suite is
**red** (`go test ./...` ≠ 0) before any implementation. **Rev 11 makes the claim stronger, not
weaker**: the red is now the literal 8/8 the installed contract describes, so the label no longer
implies more than the evidence shows — rev 10's caveat that "it does not claim all eight were
simultaneously red" is withdrawn, because at rev 11 they *are*. The per-task red remains a
**per-task** obligation too, discharged by Ship at claim time from that task's `harness_cmd`, and it
is exactly the evidence `build-feature` already records.

**A latent defect this mechanism exposed, recorded rather than silently repaired.** Requirement (d)
forced a check that revisions 4–9 never made: *is every function's assertion still true at the end
state?* Seven are. **A2's was not.** Its map row asserted that `--self-test` reports every reject
fixture **failing against the all-accept stub** with a non-zero exit — a property that is true only
while A2's stub is the current implementation and becomes **false the moment A3 installs the real
detector**. The function would have gone red at A3 and stayed red, wedging `go test ./...` for the
rest of the shipment and failing (d) outright. This had nothing to do with activation; activation
merely made it visible, because "all eight green at the end" had never been stated as a requirement
before. The fix is to assert the **durable** property A2 actually delivers — the driver enumerates
every manifest fixture by name, emits a per-fixture verdict, and runs fixtures before the real-tree
scan — and to leave the stub-specific observations where they belong, as **AC-A2.1**, read from the
script's own output at A2's boundary. **AC-A2.1 is not weakened**: it still demands the observed red
for the detector, still binds `--self-test` only, and still reserves bare-mode red for AC-A3.3. What
changes is only that it is not *also* encoded as a permanent Go assertion that contradicts A3. The
defect class — *a regression assertion pinned to a transitional state* — is worth recognizing again.

**What was rejected.** (1) *Weakening Ship's Step 4.3* to a scoped `go test ./tests/integration
-run …` — that amends `_ship.agent.md`, which is autoharness-generated with a gitignored template
(D2), and weakens the full-suite quality gate for every future shipment; explicitly out of scope
(§11). (2) *Skipping every function until its task lands, with no anchor* — H0's `go test ./...`
would then exit **0** and P-004's red phase would be unobservable, which is the silent-skip failure
P-012 and §5.0's own `bash`-resolution rule exist to forbid. (3) *A build-tag or generated
per-phase file* — that is a new runtime framework, needs a file per phase, and moves state out of
the single artifact where MP0–MP4 and MP6 can check it. (4) *Deriving activation from backlog item status* —
couples the test suite to `.backlogit/` mutable state, and a status edit would silently disable
assertions; the rev-11 manifest is deliberately **harness state**, not backlog status, and is never
reconciled against `.backlogit/`. (5) *Mode-scoped assertions* — letting targeted mode assert
strictly more than default mode would have preserved A2's stub-specific check at its boundary, but
it adds a second behavioural axis to the helper for one task's transitional property. The
durable-assertion fix above achieves the same outcome with no extra mechanism.

**Rejected at rev 11, additionally.** (6) *Keeping surface predicates and merely exempting C1/C2* —
that treats the two functions the reviewers happened to catch rather than the class. Any future
assertion whose predicate names its own subject would reintroduce the defect, and the exemption
itself would be an untested invariant. Removing the predicate→activation edge entirely is smaller
and closes the class. (7) *Deriving the phase from the surfaces present, rather than from a tracked
file* — that is the rev-10 design under a different name, with the same self-disarming property.
(8) *Amending P-002/P-004 or `harness-architect` so that "one failing test plus seven skips" counts
as a red phase* — a workspace-wide weakening of test-first for every future shipment, to accommodate
one release unit's harness; it is already out of scope (§11) and the manifest makes it unnecessary.
(9) *A per-task state file instead of one manifest* — eight files cannot be validated for ordering
or subsequence/closure consistency as a set, so the "skipped prerequisite" and "reordering" cases
would have no
detector. (10) *Putting the manifest under `.backlogit/` or `.autoharness/`* — both are inside
render or tooling boundaries and both invite exactly the backlog-coupling rejected in (4);
`tests/integration/testdata/` is outside every render boundary and is Go-idiomatic for harness
fixtures.

#### 5.0.2 Lifecycle evidence — the observation points, and what each must show (rev 12; point 7 rewritten rev 13)

The mechanism above is only as good as the evidence that it behaved as specified, so the **eight
numbered observation points** below are **required evidence**, recorded by AC-C2.11 and collected as
each occurs. **Rev 13** adds a paired selector observation **7a** alongside the rewritten point 7;
it is a companion to that point, not a ninth point. Each names the state, the command, and the
expected observation. "Zero skips" always means **default mode** (§5.0.1's qualification), and a
**targeted** run is never evidence for or against a default-mode activation claim.

| # | Point | Required observation |
|---|---|---|
| 1 | **H0** (post-`harness-architect`, pre-A1) | `go test ./...` red with **exactly 8** expected failures, each carrying its own `not implemented: <ID> <surface>` marker; **zero default-mode activation skips**; **terminal witness absent**; manifest `{"phase":"red","completed":[],"terminal":false}`. **Rev 13**: the record also asserts **set equality** — the harnessed task set is exactly the eight `017-S` task IDs, none omitted for dependency state (§5.0 "H0 eligibility") |
| 2 | **Post-A2** | A2 targeted **green** (`-gatetask=A2` exits 0); A3 targeted **expected-red** (`-gatetask=A3` exits non-zero with `--- FAIL:` and the detector-stub reason — AC-A2.5); **default suite green** |
| 3 | **Legal B-block orders** | Each of the B1/B2/B3 completion orders `B1,B2,B3` / `B2,B1,B3` / `B3,B1,B2` (and any other permutation) is **accepted** by the validator as a dependency-closed order-preserving subsequence; no legal order is rejected and no prefix rule is applied |
| 4 | **Post-C1** | **Default suite green**; C2 targeted **expected-red** — **rev 14: anchored, not merely non-zero.** The run MUST print `--- FAIL: TestDirectPushGate_FullCorpusClean` **by that exact name** and cite the **MP2 non-terminal-state** reason (`completed` is the seven-element post-C1 set, `phase: build`, `terminal: false`, witness absent — the terminal triple is not yet established). **Witness absent.** This is the state rev-11's MP5 deadlocked, and it must now be quiet |
| 5 | **Terminal** | Manifest carries the **exact eight-task set** with `terminal: true` / `phase: terminal`, **and** a valid **witness**; all **8** substantive assertions **green**; **zero default-mode skips** |
| 6 | **Terminal tree, manifest rolled back to seven tasks, witness retained** | **Failure** (MP6 integrity: witness present ⇒ manifest must be terminal). The integrity half is **mode-independent**, so this fails in default *and* targeted mode; in default mode `C2` is additionally forced active |
| 7 | **Terminal-state fixture, job-`lint` step removed — DEFAULT MODE (rewritten rev 13; mechanism specified rev 14)** | A **copied terminal-state repository fixture** under `t.TempDir()` — real tree never mutated — with a **valid manifest and witness** and the job-`lint` step removed from its `ci.yml`. A **child `go test -json` process** rooted at the copy is run **unselected** (no `-gatetask`); the mechanism is normative and is specified below. Evidence MUST show, **from parsed test events**: an `"Action":"fail"` event for `TestDirectPushGate_CIWiringIsBlocking` citing the **missing lint step**, and a **terminal event** (`pass` **or** `fail`, never `skip` and never absent) for `TestDirectPushGate_FullCorpusClean` proving it executed. The R5 pair, both halves live |
| 7a | **Same fixture, targeted — SELECTOR EVIDENCE ONLY (rev 13)** | `-gatetask=C1` against the same fixture: `TestDirectPushGate_CIWiringIsBlocking` **executes and fails**, and the other seven **skip**. This is evidence about the **selector**, and is explicitly **NOT** evidence of terminal default-mode activation |
| 8 | **Witness-only / terminal-manifest-only / malformed witness** | **Failure** in each of the three cases (MP6; proven by execution under `t.TempDir()` per MP4/AC-C2.2) |

**Fixture construction constraints for point 7 (normative; rev 13, made executable rev 14).** The
rev-13 form named a bare `go test` command and a copied fixture but **no mechanism**, and two facts
made that unexecutable as written. First, the existing `repoRoot(t)` helper — which this harness
reuses by §5.0's H0 contract — resolves the repository root by **walking up from the process working
directory to the nearest `go.mod`**, so an **in-process** run always resolves the **real checkout**,
never the copy. Second, the observation lives **inside** `TestDirectPushGate_FullCorpusClean`, so
re-running the suite from inside itself without a guard **recurses**. Rev 14 therefore specifies one
concrete bounded mechanism, in the **existing** `tests/integration/directpush_gate_test.go`. **No
new harness function and no new test is added** — the parts below are unexported **helpers** called
from the existing C2 function:

1. **Deterministic copy helper — `copyTrackedRepoFixture(t)`.** Called by the **parent**
   `TestDirectPushGate_FullCorpusClean`. It resolves the real root via `repoRoot(t)` and builds a
   **self-contained module-and-repository** fixture under `t.TempDir()` in the six ordered
   sub-steps below. Tracked-file enumeration is chosen over a hand-listed file set because it is
   **deterministic**, needs no maintenance as the corpus grows, and is **complete by construction**
   for everything C1 and C2 read — `go.mod`/`go.sum`, `tests/**` (including this harness and the
   `testdata/directpush-gate/` manifest, witness, and fixture corpus), `scripts/**`,
   `.github/**` (the §5.2 corpus and `ci.yml`), and `AGENTS.md`. It excludes the real `.git/`,
   build output, and every ignored path, so the copy is a **module root** and nothing else.

   **(a) Enumerate the index WITH MODES, and fail closed on anything unsupported (rev 15).** The
   enumeration is `git ls-files -s -z` executed in the real root — **`-s` is load-bearing**,
   because the plain `git ls-files -z` form named at rev 14 emits paths only and therefore cannot
   distinguish a regular file from a symlink or a gitlink, which is precisely the distinction a
   faithful copy has to make. Each record is parsed NUL-wise (never line-wise: a path may contain a
   newline) into `mode`, object, stage, and path. Handling is a **closed set**:

   | Index mode | Handling |
   |---|---|
   | `100644` regular file | **Copy** with regular permissions |
   | `100755` executable file | **Copy**, preserving the **executable bit** on platforms that carry one |
   | `120000` symlink | **FAIL CLOSED** — the harness does not reproduce link semantics |
   | `160000` gitlink / submodule | **FAIL CLOSED** — no submodule recursion, no network |
   | any other mode, or any **stage ≠ 0** (unmerged) | **FAIL CLOSED** |

   **Measured against this repository, rev 15 (`git ls-files -s`): all 629 index entries are
   `100644`. There are ZERO `100755`, ZERO `120000`, and ZERO `160000` entries, and there is no
   `.gitmodules` file.** Two consequences are stated so neither is assumed. First, **no
   skip-list carve-out is written for `references/`, because `references/` is not in the index at
   all** — it holds `references/herdr`, a **nested independent clone** excluded by
   `.git/info/exclude`, so `git ls-files` never enumerates it and there is nothing for the copy to
   skip. A plan clause that "skips the `160000 references` gitlink" would describe an index entry
   this repository does not have. Second, the fail-closed rows are therefore **not dead text but a
   forward guard**: the table is evaluated on every run, so the day a submodule or symlink is added
   the fixture **fails loudly** instead of silently producing an incomplete copy. The `100755` row
   is accepted rather than rejected for the same forward reason, and costs nothing today.

   **(b) Copy the enumerated entries.** Each accepted entry is copied to the same relative path
   under the destination, creating parent directories as needed. **Only** entries the table accepts
   are copied; nothing is inferred from the working tree, so ignored and untracked content — the
   `references/herdr` clone included — is structurally incapable of entering the fixture.

   **(c) Copy the generated terminal witness EXPLICITLY (rev 15).** `tests/integration/testdata/`
   `directpush-gate/terminal-witness.json` is created by **C2's own completion** (§5.0.1, AC-C2.10)
   and at the moment this helper runs it may not yet be **in the index** — so sub-step (a) can
   legitimately fail to enumerate it. Rev 14's index-only copy therefore had a real hole: a fixture
   missing the witness is a **witness-absent + terminal-manifest** state, which MP6 fails as an
   **integrity** error, making point 7 vacuous for a reason unrelated to the lint step. The helper
   therefore copies the **real checkout's current** `terminal-witness.json` **from the working
   tree**, after sub-step (b) and before sub-step (d), and: **requires it to exist** (absence
   **fails** the observation — it is never treated as "nothing to copy"); **parses** it; and
   **validates it against the copied manifest** to the full MP6 standard — exact eight-element set,
   `terminal: true`, `phase: terminal`, and **matching digest**. **Copied manifest and witness must
   be valid and in agreement BEFORE sub-step (e) removes the copied lint step.** This ordering is
   what keeps the child's failure **attributable to the lint removal**: the fixture is proven sound
   first, and only then is the single intended defect introduced. Copying it here rather than
   special-casing it in (a) also keeps the index-mode table closed and honest.

   **(d) Make the copy a SELF-CONTAINED Git repository (rev 15).** Sub-steps (a)–(c) produce a
   module root with **no** repository: the real `.git/` is never copied (that would import the real
   branch, index, hooks, and `info/exclude`, and is forbidden). But the copied `scripts/**` and the
   harness itself use tracked-corpus enumeration and `HEAD`, and a Git command run inside a
   directory that is not a repository **walks upward** — which under `t.TempDir()` could resolve
   some unrelated ancestor repository, or none. The helper therefore **initializes Git in the
   copy**, without network and without touching global state: `git init` in the copy root;
   `git -C {copy} config --local user.email` / `user.name` set to a **non-secret, deterministic,
   repo-local** identity (local scope only — never `--global`, and no credentials of any kind);
   `git -C {copy} add -A`; and a **deterministic local commit**. A commit is **preferred over a
   bare index** so that scripts using the tracked corpus **and** `HEAD` work normally.
   **Assertion before the child is launched:** `git -C {copy} rev-parse --show-toplevel`, after
   canonicalization (symlink and `8.3`/case normalization, so the comparison is meaningful on
   Windows and on macOS `/var`→`/private/var`), **MUST equal the copy root**. A value that resolves
   to the real checkout, to any ancestor of the temp directory, or to any other repository **fails
   the observation** — this is the assertion that makes "no nested or sibling repository can be
   selected" checkable rather than asserted. No submodule recursion and no remote is configured, so
   nothing in this step can reach the network.

   **(e) Remove only the copied job-`lint` step** — item 2 below.

   **(f) No-mutation evidence — BYTE-IDENTICAL STATUS SNAPSHOTS, not an empty tree (rev 15).**
   Rev 14 asserted `git status --porcelain` over the real checkout was **empty before and after**.
   That form is **unsound in this repository for two independent, verified reasons**, and it
   deadlocks rather than protects. **First**, the observation runs **inside C2**, whose own
   completion legitimately writes `terminal-witness.json` and the terminal `harness-state.json`
   (AC-C2.9/AC-C2.10) — and the spawn block is entered **only** in `terminal` phase (item 4), so at
   the one moment point 7 is meaningful the real tree **necessarily carries** that control-state
   dirt. An empty-tree precondition is therefore **unsatisfiable exactly when the evidence is
   required**. **Second**, `references/herdr` is excluded only by **`.git/info/exclude`**, which is
   **local and not part of the repository** — it does not survive a clone — so on CI or any fresh
   checkout `git status --porcelain=v1 -z --untracked-files=all` reports that tree as untracked and
   the "empty" assertion fails for a reason that has nothing to do with this harness.

   The rule is therefore **equality of two snapshots**, which is strictly the **right** predicate:
   what must be proven is that **the fixture mechanism changed nothing**, not that the checkout was
   pristine. The helper captures a **full-tree** status snapshot **immediately before** and
   **immediately after** the copied-fixture child run, using the **same NUL-safe, untracked-
   inclusive** form this plan already mandates for the Stage commit gate (§5.7):

   ```text
   git status --porcelain=v1 -z --untracked-files=all
   ```

   Both snapshots are taken over the **real checkout**, parsed **NUL-wise** (never split on
   newlines), and compared as **byte-identical** record multisets. `--untracked-files=all` is
   load-bearing for the same reason §5.7 gives: the default `normal` mode collapses an untracked
   subtree into a single `?? dir/` entry, which no per-file comparison can evaluate. **The parent
   fails if and only if the snapshots differ** — i.e. if and only if the fixture mechanism itself
   changed the real checkout. **Pre-existing dirt is permitted only by being unchanged**: C2's own
   manifest and witness records, and any `references/herdr` records, appear **identically** in both
   snapshots and so cancel. **This does not weaken the no-mutation evidence — it strengthens it.**
   The empty form proved "the tree was clean, and (by inference) we did not dirty it"; the snapshot
   form proves the load-bearing claim **directly and unconditionally**, and it detects mutations the
   empty form could never see, including a **modification to an already-dirty file**, a
   **deletion**, and a **new untracked artifact written beside existing dirt**. A snapshot that
   differs is a failure **even if the difference is a file returning to a clean state**. This
   snapshot-equality form is the **canonical** no-mutation evidence for this plan; where AC-A2.4,
   AC-C1.6, and AC-C2.2 take the same evidence they take it in **this** form, and where the harness
   already owns it the existing assertion is reused rather than duplicated.
2. **Fixture preparation — ordered, so the injected defect stays attributable.** In the **copy
   only**: the copied `harness-state.json` and `terminal-witness.json` MUST be **valid and in
   agreement** (exact eight-element set, `terminal: true`, `phase: terminal`, matching digest) —
   otherwise `gate()` Stage 1 fails the child for an **integrity** reason and the observation
   reverts to vacuous for a second reason. **Rev 15 makes the ordering normative**: that agreement
   is established and checked in item 1 sub-step (c), and the fixture is committed in sub-step (d),
   **before** the lint step is removed here. The removal is therefore the **only** divergence
   between the fixture's `HEAD` and its working tree, which is exactly what makes the child's
   failure attributable to **it** rather than to a malformed or incomplete copy. **Only the
   job-`lint` gate step is removed**, by parsing the copied `.github/workflows/ci.yml` and deleting
   that one entry from `jobs.lint.steps[]` — the same structural walk AC-C1.6 mandates, never a
   text edit. The removal is left as an **uncommitted worktree edit in the copy**; the copied
   `ci.yml` stays tracked, so the copied scripts' tracked-corpus enumeration is unaffected.
3. **Child process — `runFixtureChildGoTest(t, fixtureRoot)`.** The parent launches a **child `go
   test` process**, because that is what makes `repoRoot(t)` resolve the **copy**: the child's
   `exec.Cmd.Dir` is set to the **copied module root**, so the child test binary's working directory
   is the copy's `tests/integration` and its upward `go.mod` walk terminates inside the fixture. The
   command is Go's own invocation with Go's **built-in test selection**, and no shell:
   `go test -json ./tests/integration -run '^TestDirectPushGate_(CIWiringIsBlocking|FullCorpusClean)$' -count=1`.
   It carries **no `-gatetask`**, so the child runs in **default activation mode** — the only mode in
   which manifest-derived activation is under test — while `-run` bounds the child's **process
   scope** to C1 and C2. `-run` is a **process-scope** bound and is **not** an activation input:
   it never appears in either `gate()` table and has no bearing on the default-mode claim, which is
   about which of the *selected* functions activate versus skip. `-count=1` defeats the test cache.
   The child inherits the parent environment (so the module cache and toolchain resolve normally)
   **plus** the sentinel of item 4. Invocation is via `exec.CommandContext` with the Go binary —
   never `bash -c`, `pwsh -Command`, or any shell — so the mechanism is identical on Linux, macOS,
   and Windows.
4. **Non-recursive sentinel — `DIRECTPUSH_GATE_FIXTURE_CHILD=1`.** The parent sets exactly this one
   variable on the child. `TestDirectPushGate_FullCorpusClean` reads it and, when set, executes
   **only** its substantive default-mode assertions and **does not enter the parent-only spawn
   block** — so the child never spawns a grandchild and the recursion is bounded at exactly **one**
   level. Per §5.0.1 "The fixture-child sentinel", the sentinel **MUST NOT** skip C2, **MUST NOT**
   bypass MP1/MP6 or any other Stage 1 integrity check, and has **no** activation effect; its sole
   function is disabling the parent-only spawn block.
   **Entry condition for the spawn block, stated so the mechanism is bounded in every phase.** The
   block is entered only when **both**: the sentinel is **unset** (this is the parent), **and** the
   real tree's manifest `phase` is **`terminal`** — which is the only state in which point 7's claim
   is meaningful, since the fixture is a copy of the real tree. In `red` the function fails on its
   own not-implemented marker long before the block; in `build` the function is not active at all in
   default mode. Reading `phase` here is a **control-state** read (§5.0 "Control-state inputs are
   expressly exempt"), **not** a substantive-surface predicate, and it is **not** an activation
   decision: it selects a step **inside** an already-active body and can never cause C2 to skip. In
   a `terminal`-state **targeted** `-gatetask=C2` run the block may also be reached; that is
   harmless and non-recursive, and its child is still a default-mode run — but per §5.0.2 it is the
   **default-mode** parent invocation that is recorded as point-7 evidence.
5. **Evidence by parsed `go test -json` events, never by scraping text.** The parent decodes the
   child's stdout as a **stream of JSON test events** (`encoding/json` over the `-json` stream) and
   evaluates, against the **exact** test names:
   * `TestDirectPushGate_CIWiringIsBlocking` MUST have a terminal `"Action":"fail"` event, and its
     accumulated `output` events MUST cite the **missing job-`lint` gate step** — not some other
     failure.
   * `TestDirectPushGate_FullCorpusClean` MUST have a **terminal event** whose action is `pass`
     **or** `fail`. A `skip` action is a **failure of this observation**, and so is the **absence**
     of any terminal event for that name. This is the half that proves C2 **executed**; its verdict
     is deliberately not constrained, because what point 7 proves is **activation**, not C2's
     substantive result against a deliberately broken fixture.
   * **Expected child exit status is non-zero**, because C1 detects the removed step. A **zero**
     child exit is a failure of this observation. Exit status alone is never sufficient evidence —
     it is checked **in addition to**, never instead of, the two event assertions above.
   * **No recursion and no timeout.** **Rev 15 — stated as a precondition plus a bound, because the
     rev-14 phrasing named an impossible condition.** Rev 14 required "no second-level child
     (detected by the sentinel already being set when the parent block is reached)". That detector
     can **never fire**: item 4's guard means a sentinel-set process **does not reach** the parent
     block at all, so "sentinel set **when the block is reached**" is unsatisfiable by construction
     and proves nothing. Recursion is prevented — not detected — on the **parent side**: entering
     the spawn block **requires** the sentinel to be **unset** (item 4), so the child, which always
     has it set, cannot spawn a grandchild, and the depth is bounded at exactly **one** level
     structurally rather than by observation. What the parent **does** check, and what remains
     genuine evidence, is the **bound**: the child must terminate **within its explicit
     `context.WithTimeout`**, and a **context-deadline / killed-child termination fails** the
     observation rather than being read as inconclusive. A defensive assertion that the sentinel
     **is** set in the child environment the parent constructed is permitted as a cheap
     self-check on the parent's own `exec.Cmd.Env`, but it is **not** the recursion control and
     must not be recorded as one.
   * A **malformed event stream** — undecodable JSON, a truncated stream, or a stream containing no
     terminal event for either name — **fails** the observation. It is never treated as an
     inconclusive pass.
6. **Bounded execution and bounded output.** The child runs under an explicit
   `context.WithTimeout` (the existing file-level timeout-constant convention, e.g. the
   `buildScriptTimeout` shape already used in `tests/integration/`), so a hung child is killed and
   reported rather than blocking CI. Child **stdout and stderr are captured and surfaced only in
   bounded failure output** — a truncated tail on failure, nothing on success — so a passing run
   stays quiet and a failing one stays readable.
7. **The parent passes only when the child evidence matches.** Every condition in item 5 must hold;
   any one of them failing fails `TestDirectPushGate_FullCorpusClean` in the parent. The **real
   checkout is unchanged** throughout — only `t.TempDir()` is written — and that claim is
   **discharged by the item 1(f) snapshot equality**, not by an empty-tree assertion: the
   before/after `git status --porcelain=v1 -z --untracked-files=all` snapshots must be
   **byte-identical**, so C2's own legitimate manifest and witness dirt is tolerated **only** by
   being unchanged, and any mutation the mechanism causes fails the parent.

Point **7a** is kept **separate and unchanged in kind**: the targeted `-gatetask=C1` invocation
against the same fixture is **selector evidence only**. It activates C1 **by construction** and is
explicitly **not** evidence of terminal default-mode activation, so it is never merged into point 7
or offered as a substitute for it.

**Why point 7 was rewritten (rev 13).** The rev-12 form ran `-gatetask=C1`. Targeted mode activates
the selected function **by construction**, so the observation was **vacuous** as proof of terminal
default-mode activation: it would have passed identically under the rev-10 self-disarming design
that R5 exists to exclude. It also never exercised **C2**, so the "both halves live" claim rested on
nothing. The replacement is **unselected** — which is the only mode in which manifest-derived
activation is actually under test — and asserts **both** functions by name. Point **7a** keeps the
targeted command, correctly labelled as selector evidence, so no observation is lost. **Rev 14 adds
what rev 13 left unsaid**: a correct observation that no one can execute is not evidence either, so
the copy helper, the child-process root, the sentinel, and the event-parsing contract above are
**normative**, not illustrative.

Points 1, 2, and 4 are taken at the boundaries where they occur; points 3, 5, 6, 7, 7a, and 8 are
executed or replayed at C2. None of them is satisfied by narration: each is a recorded command and
its observed output.

### 5.1 House-style contract

Matches `check-retired-architecture.sh` (closest precedent) on `::notice::` mode emission and the
fixture/manifest harness; matches retired-arch / write-path / unignore on the `python3`→`python`
probe, `ROOT="$(git rev-parse --show-toplevel)"`, and exit codes **0** pass / **1** violation /
**2** invalid invocation. Deliberately does **not** adopt `check-gitignore-append-only.sh`'s
`SCRIPT_DIR` + `--base-ref` shape, because this gate asserts a whole-tree invariant, not a diff.
`set -euo pipefail`. **No waiver mechanism.** Header states INVARIANT, detector scope, THREAT
MODEL, and ACCEPTED RESIDUALS.

**Two modes only** (rev 2): bare (repo scan) and `--self-test` (fixtures + selection assertions,
**then** real-tree scan). The rev-1 third mode is dropped — `--self-test-integrity` exists in
retired-arch solely because its verdict step is *toggled* by `RETIRED_ARCH_GATE_ADVISORY`; §7.1
has no toggle, so the split would be semantically identical to one `--self-test` call. Running
fixtures *before* the real scan preserves the `gitignore-append-only` rationale: a broken checker
reports "self-test failed", never a false violation.

### 5.2 Corpus (widened in rev 2)

`git ls-files` pathspec, markdown only:

`.github/agents/**`, `.github/policies/**`, `.github/skills/**`, `.github/instructions/**`,
`.github/prompts/**`, `.github/copilot-instructions.md`, `AGENTS.md`

`scripts/**` and `docs/**` are outside the pathspec **by construction** — there is no exclusion
predicate, so rev-1's "assert the exclusion" selection checks are dropped as vacuous. `docs/**`
being out of corpus is what lets the deliberation and this plan quote 3e freely.

**ACCEPTED RESIDUAL** (script header): `.github/workflows/**` and `.github/constraints/**` are not
scanned (the latter holds no markdown). R4 is therefore scoped to "a regeneration of the scanned
surfaces".

### 5.3 Detector (rev 3 — per-construct, structure-aware, closed construct set)

Rev 1's line-scoped in-line-marker heuristic was **unsound in both directions** and is replaced.

**Verdict granularity — per construct, never per block (rev 3).** Every matched construct is
scored and reported **individually**. A block is never accepted as a whole. This matters because
after normalization, Step 1.5's item 3 becomes one block containing *both* 3b (`create a PR to
\`main\`` — accept) and 3e (`Attempt a direct push to \`main\` first` — reject); a block-level
accept verdict would mask the reintroduced violation.

**Block normalization (rev 3).** Continuation lines are joined into logical blocks before
matching, covering **both** well-formed nested list items **and lazy paragraph continuations** —
the latter is the shape Step 1.5's a–e sub-steps actually use (see §11), i.e. precisely the file
the gate exists to protect. A re-render that wraps a construct across two lines cannot evade it.

**Violation constructs** (closed set — two shapes):

1. **Push** — an imperative push **verb-anchored on a push** whose *target* binds to a
   default-branch token. Excludes PR-base forms (`PR to`, `pull request to`, `--base`) and
   read/checkout verbs (`fetch`, `log`, `show`, `diff`, **`pull`**, **`checkout`** — rev 3; the
   corpus legitimately contains 3d's "pull `main`", P-011's `git checkout main; git pull`, and
   Ship Step 0.5's "Switch to the default branch"). This is the discrimination rev 1 never
   specified: 3b `Push the staging branch and create a PR to \`main\`` is **accept**
   (target = branch); 3e `Attempt a direct push to \`main\` first` is **reject** (target = `main`).
2. **Commit authorization** — text **permitting** artifacts to be committed/persisted *on* a
   default-branch token. **Verificational / postcondition phrasing is not a violation** (rev 3):
   "verify/confirm that X **has reached** the default branch" asserts a state; "artifacts may be
   committed **on** the default branch" grants a permission. Only the permissive form is a
   finding. This shape catches `_stage.agent.md` L42 and the P-010 bullet — which rev 1 missed
   entirely — without flagging B1's reworded preamble.

**Table handling is CELL-scoped, not row-scoped (rev 3).** A Role Boundary row is one line whose
Allowed cell holds verbs ("commit, push") and whose Forbidden cell holds the default-branch token.
Row-scoped matching would pair them and fire on every correct row. Constructs are matched **within
a single cell**, and the governing context is that cell's **column header**.

**Default-branch token** = alternation over: the configured default branch, the literal `main`,
the literal `{{DEFAULT_BRANCH}}`, and the prose "default branch". The placeholder form closes the
regeneration-evasion path named in §5.0 and §9.

**Default-branch resolution must not abort the gate (rev 3).** `git symbolic-ref
refs/remotes/origin/HEAD` is **not** created by `actions/checkout`, so under the mandated
`set -euo pipefail` an unguarded command substitution would exit non-zero in CI and be
indistinguishable from a real violation — wedging job `lint` the moment C1 wires it. Resolution
MUST be explicitly guarded (`|| true`, empty-result check) with a documented fallback to `main`,
and `--self-test` MUST assert both the resolution-failure fallback and a non-`main` default-branch
name.

**Prohibition resolution — structural, not in-line-only.** A construct is **accepted** when either:

- its **governing context** is prohibitive — (a) a **bold lead-in line** of the form
  `**<actor> MUST NOT**:` or `**Forbidden**` immediately preceding the list (rev 3 — this, not an
  ATX heading, is the shape `workflow-policies.md` actually uses for both the Ship
  "Commit or push directly to `main`" bullet and the Stage bullet B2 adds; reading "heading" as
  ATX-only would flag both and make zero-findings unreachable), (b) an ATX `MUST NOT` / `Forbidden`
  heading, or (c) a table cell under a `Forbidden` / `MUST NOT` **column header**; **or**
- a prohibition marker **precedes** the construct within the same clause:
  `never`, `must not`, `do not`, `don't`, `no`, `forbidden`, `prohibited`.

`rather than` and `instead of` are **removed** from the marker set — they are directional, not
prohibitive, and would have excused "push directly to `main` rather than opening a PR".
Marker-must-precede closes the "…first; do not create a PR unless rejected" fail-open.

Structural resolution is what makes the real corpus pass: P-010's **Ship MUST NOT** bullet
"Commit or push directly to `main`" in `workflow-policies.md` and `_ship.agent.md`'s Role
Boundary `| Git |` row carry **no in-line marker** — their prohibition
lives in the bold lead-in and the `Forbidden` column header respectively. Rev 1 would have flagged
both, making zero-findings unreachable.

### 5.4 Task A1 — fixture corpus + manifest (test data only)

**Files**: `scripts/testdata/directpush/` (14 fixtures), `scripts/testdata/directpush-manifest.json`,
and — **rev 11** — one line of `tests/integration/testdata/directpush-gate/harness-state.json` (the
shared **harness-state completion record**, §5.0.1). That file is **not a second skill domain**: the
edit is a single-key state transition (`phase` `red`→`build`, append `"A1"` to `completed`) in a file
H0 created, performed as part of this task's completion. Every task in this release unit makes the
same one-line append; the scope is stated explicitly here and in each task so it is not mistaken for
scope creep. A1's size band is unchanged.

| Fixture | Verdict | Locks |
|---|---|---|
| `directpush-reject-attempt-first.md` | reject | literal 3e construct |
| `directpush-reject-push-origin-main.md` | reject | `git push origin main` form |
| `directpush-reject-default-branch-prose.md` | reject | "default branch" phrasing |
| `directpush-reject-branch-placeholder.md` | reject | literal `{{DEFAULT_BRANCH}}` form |
| `directpush-reject-commit-on-default.md` | reject | **commit-shaped** permission (the L42 shape) |
| `directpush-reject-marker-bearing.md` | reject | violation with a **trailing** `do not` clause |
| `directpush-reject-rather-than.md` | reject | "push directly to `main` **rather than** opening a PR" — proves the pruned markers do not fail open |
| `directpush-reject-wrapped-construct.md` | reject | construct split across a line break **in lazy-paragraph-continuation shape** (the real Step 1.5 shape) |
| `directpush-reject-mixed-block.md` | reject | one normalized block holding an accepted **and** a violating construct — proves per-construct verdicts |
| `directpush-accept-prohibition-boldleadin.md` | accept | `**Stage MUST NOT**:` **bold lead-in** bullet, marker-free (the real P-010 shape) |
| `directpush-accept-prohibition-tablecell.md` | accept | `Forbidden` **column-header** table cell, marker-free (the real `_ship.agent.md` Git row shape, incl. an Allowed cell with push verbs to prove cell-scoping) |
| `directpush-accept-marker-precedes.md` | accept | "**never** commit directly to the default branch" — the in-line marker-precedes branch |
| `directpush-accept-read-and-pr-forms.md` | accept | `git fetch origin main`, `git log origin/main..main`, `git show origin/main:…`, `git checkout main`, `pull \`main\``, and 3b's `create a PR to \`main\`` |
| `directpush-accept-verification-form.md` | accept | "verify that artifacts **have reached** the default branch" — the postcondition shape B1 installs |

**Manifest is authoritative** for verdicts, with **sorted keys** (Principle IX). `--self-test`
asserts filename prefix (`directpush-reject-` / `directpush-accept-`) and manifest verdict
**agree**, and asserts **bijection** (every file on disk is in the manifest and vice versa),
closing rev 1's silently-unexercised-fixture hole.

**The corpus is deliberately NOT widened for the rev-6 defect.** `directpush-accept-read-and-pr-forms.md`
already carries `git log origin/main..main` as an **accept** case, and it stays exactly as it is:
its job is to prove the detector does not flag *read* verbs, and that property is unrelated to
whether the Orchestrator happens to use that range. The rev-6 defect — **main-only discovery** —
is a different defect class from the two §5.3 violation constructs (it is neither a push to the
default branch nor a permission to commit on it), so folding it into the detector would mean
opening the **closed construct set**, adding fixtures, and re-deriving A1's 14/9/5 counts through
A1→A2→A3→C2. That is a materially larger change than the finding requires. The rev-6 regression is
instead rejected mechanically by the **B1 harness function**, which already exists, is already
scoped to `_orchestrator.agent.md`, and is already B1's red→green boundary (§5.0). No fixture, no
manifest entry, and no construct-set change.

Each fixture carries a single leading H1 so **P-008 markdownlint** (`MD041`/`MD025`) passes.

**Acceptance criteria**

- **AC-A1.1** 14 fixtures exist; manifest keys are sorted and in bijection with the directory.
- **AC-A1.2** Every manifest verdict agrees with its fixture's filename prefix
  (`directpush-reject-*` / `directpush-accept-*` — the actual on-disk names, rev 8), and both
  classes are non-empty: 9 reject, 5 accept. **Static data check only** —
  A1 authors fixture data and owns no executable logic, so the *executed* non-vacuity proof of
  these assertions is **AC-A2.4 on task A2**, which owns the fixture-suite driver (rev 4 — review round 3).
- **AC-A1.3** `markdownlint "scripts/testdata/directpush/**"` exits 0.
- **AC-A1.4** **Harness-state lifecycle opened (rev 11).** H0's
  `tests/integration/testdata/directpush-gate/harness-state.json` is present and was initialized to
  exactly `{"phase":"red","completed":[],"terminal":false}`, and the recorded H0 evidence is the
  **literal 8/8** red — eight `--- FAIL:` lines, each carrying its own task-specific
  `not implemented:` marker from the §5.0.1 marker table, and **zero default-mode activation skips**
  with the **terminal witness absent** (rev 12). On this task's completion, and
  **atomically with** the fixture work above, the manifest performs the **one-way** `red`→`build`
  transition and `completed` becomes exactly `["A1"]`. The transition is one-way: `build`→`red` is
  an impossible state and fails closed (MP1). The append does **not** by itself make this test pass
  — `-gatetask=A1` evaluates AC-A1.1–AC-A1.3 regardless of manifest state, so a manifest advanced
  ahead of the fixtures fails on the real assertion. A1 additionally carries **MP0** (order and
  graph integrity: `gateOrder` is exactly the eight IDs in declared sequence, and each `gateDeps`
  row lists that task's **prerequisites** — the `depends-on` targets of the recorded edges).
  **Rev 12**: A1 carries **no** MP5 obligation (MP5 is **RETIRED** — it deadlocked the legitimate
  post-C1/pre-C2 state) and **no** "MP1 probe non-vacuity" obligation (MP1 is the manifest
  validator and has no probes; the phrase was a stale attribution of MP5's probe check). The
  **terminal completion witness** is **absent** at H0 and stays absent throughout `build` — a
  witness present at H0 is a fail-closed error (MP6).

### 5.5 Task A2 — fixture-suite driver over the stubbed detector

**Files**: `scripts/check-direct-push-language.sh` (replace the H0 stub's mode dispatch with the
real driver: mode dispatch, fixture/manifest harness, selection assertions, `::notice::` emission),
plus the shared one-line append to `tests/integration/testdata/directpush-gate/harness-state.json`
(`completed` becomes `["A1","A2"]`) — the harness-state completion record of §5.0.1, not a second
skill domain; size band unchanged.
The **detector remains stubbed** — A2 owns no detection logic, so the stub is narrowed from
"everything fails with the not-implemented marker" to "every construct classified `accept`", and
A3 replaces it.

**Scope change (rev 4)**: A2 no longer creates the harness. The Go harness and the not-implemented
stub script are H0 deliverables produced by `harness-architect` at Ship Step 2 (§5.0), because
Step 2 runs once up front for the whole queue — a *task* whose content is "create the harness" is
structurally unreachable in Ship's flow. A2 keeps the driver work, which is genuine implementation.

- **AC-A2.1** Every **reject** fixture is reported **failing** against the all-accept stub, by
  name, and **`--self-test` exits non-zero**. This is the **observed red for the detector itself**
  — rev 1 never took the detector red. **Rev 8 scope correction**: the non-zero requirement binds
  **`--self-test` only**. Bare-scan exit is deliberately **not** constrained here, because A2's
  stub classifies every construct `accept`, so a bare scan correctly reports zero findings and
  exits **0** — demanding red there would make AC-A2.1 unsatisfiable by a correct A2
  implementation, or satisfiable only by a stub that lies about the real tree. Bare-mode red is
  A3's obligation once a real detector exists (**AC-A3.3**), and the A2→A3 dependency edge
  guarantees it is evaluated after this one. **Rev 10 clarification**: this criterion is verified
  from the **script's own `--self-test` output** at A2's boundary, and is deliberately **not**
  encoded as a permanent assertion in `TestDirectPushGate_SelfTestDriverEnumeratesFixtures` — a Go
  assertion that the detector is still a stub would go **red the moment A3 replaces it** and would
  wedge `go test ./...` for the rest of the shipment (§5.0.1). The criterion itself is unchanged
  and unweakened; only its carrier is named.
- **AC-A2.2** `--self-test` runs fixtures **before** the real-tree scan, so a broken checker
  reports "self-test failed", never a false violation.
- **AC-A2.3** `git diff --exit-code -- .github/workflows/ci.yml` is clean (no CI change yet).
- **AC-A2.4** Bijection and prefix/manifest-agreement assertions each **fail** when deliberately
  violated — non-vacuity proven by **execution**, not asserted. The proof is performed against a
  **temporary fixture copy under `t.TempDir()`**, never by editing the committed fixture tree
  (same no-dirty-tree-edit rule as AC-C2.2). Relocated here from A1 (rev 4 — review round 3): A1
  owns fixture data only and cannot execute an assertion, while this task owns the driver that
  runs it. The existing A1 → A2 dependency edge already guarantees the fixtures exist before this
  criterion is evaluated, so execution order is unaffected.
- **AC-A2.5** **Two separate, exactly-anchored invocations (rev 10).** A single driver-scoped
  `-run` command cannot establish anything about a test it does not run, so the evidence is split:
  1. `go test ./tests/integration -run '^TestDirectPushGate_SelfTestDriverEnumeratesFixtures$'
     -gatetask=A2` exits **0**.
  2. `go test ./tests/integration -run '^TestDirectPushGate_SelfTestPasses$' -gatetask=A3` exits
     **non-zero**, **and** its output carries `--- FAIL: TestDirectPushGate_SelfTestPasses` together
     with the expected **detector-stub** reason — A2's stub classifies every construct `accept`, so
     the reject fixtures' actual verdicts disagree with the manifest. An arbitrary failure does
     **not** satisfy this: a compile error, an absent `bash`, a missing fixture corpus, or a
     mis-typed `-gatetask` value is a harness defect, not the observed red, and each is rejected by
     name.

  Both regexes are **anchored** (`^…$`). The `--- FAIL:` requirement is the non-vacuity guard and is
  load-bearing: **verified on this platform**, `go test -run '^TestNoSuchTest$'` prints
  `ok … [no tests to run]` and exits **0**, so "exits non-zero" alone could be satisfied — or
  silently *unsatisfiable* — by a pattern that selects nothing. **Bare-mode red stays A3's
  obligation** (AC-A3.3) and is not asserted here; P-004 is not weakened in either direction — this
  criterion adds an observation, it removes none.

### 5.6 Task A3 — detector implementation

**Files**: `scripts/check-direct-push-language.sh` (replace stub with the §5.3 detector), plus the
shared one-line append to `tests/integration/testdata/directpush-gate/harness-state.json`
(`completed` becomes `["A1","A2","A3"]`) — the harness-state completion record of §5.0.1, not a
second skill domain; size band unchanged.

- **AC-A3.1** `--self-test` reports every fixture by name with actual == manifest verdict.
- **AC-A3.2** Guarded default-branch resolution is asserted for **both** the resolution-failure
  fallback to `main` and a non-`main` configured default branch.
- **AC-A3.3** Bare invocation exits **1** and names all four §2 surfaces — identified by **file and
  matched construct text**, not line ordinal.
- **AC-A3.4** The **first full-corpus scan output** — resolved corpus file count and the complete
  enumerated finding list — is recorded before B1 begins, so any additional surface is discovered
  now rather than at C2.
- **AC-A3.5** A `::notice::` names the resolved mode on every invocation.

---

## 6. Sub-epic B — Contract correction

### 6.1 Task B1 — Orchestrator Step 1.5

B1 corrects `.github/agents/_orchestrator.agent.md` on **two axes**. Revisions 1–5 corrected what
Step 1.5 **authorizes** (§6.1.1). Rev 6 corrects how Step 1.5 **discovers and verifies** the work
it gates (§6.1.2) — without which the authorization correction is unexecutable.

#### 6.1.1 Authorization surfaces (rev 1–5, unchanged)

Delete sub-step **3e** in full (the "Branch protection handling" block and its three nested
bullets plus trailing rationale).

**Fold its branch-point into 3a** — required, because 3e's "Create branch
`chore/stage-{shipment_id}` **from the current commit**" is the **only** instruction covering the
*already-committed-but-unpushed* path that Step 1.5 step 2 routes into step 3. 3a currently
specifies no branch point. Revised 3a (rev 6 extends this to reuse the reported branch — see
§6.1.2):

> a. Check out the Stage artifact branch `$STAGE_BRANCH` reported by the handback record — or,
>    when none was reported, create `chore/stage-{scope-slug}` **from the current commit** — and
>    commit any uncommitted backlog and
>    planning files to it. **Then re-record the handback commit** (rev 7 — correction 5, §6.1.2):
>    set `stage_head_commit` to `git rev-parse HEAD` and discard any `stage_artifact_paths` the
>    handback reported. **Preserve `stage_base_commit` unchanged** (rev 8), or — when it is still
>    UNRESOLVED — set it to the `HEAD` captured **before** this commit, so step 4 derives its file
>    set from the whole `{stage_base_commit}`→`{stage_head_commit}` range rather than from the tip
>    commit alone.

**Rewrite the Step 1.5 preamble** (rev 3 — the fourth surface, §2). It currently states the gate's
postcondition as artifacts being "committed to the default branch", which remains a compliant
reading of direct commit-to-default. Revised (rev 6 widens the scope clause so a shipment-less run
is inside the gate rather than outside it — see §6.1.2):

> After Stage completes — **whether or not a shipment was formed** — and before any routing to
> Ship, verify that all staging artifacts (backlog items, shipment manifests, and the plan,
> deliberation, and memory records Stage wrote) **have reached the default branch via a merged
> staging PR** and are present on the remote.

This is a **postcondition/verification** form, which §5.3 construct 2 explicitly does not treat as
a violation — so it satisfies the gate without narrowing the detector.

**Preserved unchanged**: 3b–3d; step 4's `git show origin/main:.backlogit/queue/{shipment_id}.md`
verification and `STAGING_GATE_FAIL` halt; the verified-on-origin broadcast. Step 1.5 remains
NON-NEGOTIABLE and **Orchestrator-owned**.

#### 6.1.2 Branch discovery and two-outcome verification (rev 6)

**The defect.** Step 1.5's two discovery inputs are both blind to the branch Stage actually commits
on, once B2/B3 require Stage to create, check out, and commit to `chore/stage-{scope-slug}` and
leave the worktree there:

| Step 1.5 input | Existing form | Why it misses the Stage branch |
|---|---|---|
| step 1 — dirtiness | `git status --short -- .backlogit/` | A no-shipment Stage run writes only `docs/plans/`, `docs/decisions/`, `docs/memory/` — the exact `fdff9e4` shape. The pathspec reports **clean**. |
| step 2 — unpushed commits | `git log origin/main..main` | Compares the **default branch** against its own remote-tracking ref. Stage's commit is on `chore/stage-…`, which is not the default branch, so the range is **empty** by construction. |

Both inputs report "synchronized", Step 1.5 skips step 3 entirely, and step 4 then fails its
`origin/main` manifest check — or, for a shipment-less run, has no manifest to name at all. The
**no-shipment persistence route documented end to end in §6.2 therefore cannot complete**. The
authorization fix and the discovery fix are not separable: correcting only the former relocates
Stage's commit to a branch the gate cannot see.

**Correction 1 — the Stage artifact branch becomes an explicit handback.** Step 1 (Route to Stage)
currently declares Stage's output as "a `shipment_id` in `queued` status" only. Revised bullet
under step 3, and revised step 4:

> * Stage's expected output — the **staging handback record**: `stage_branch` (the dedicated Stage
>   artifact branch Stage created and committed on), `stage_base_commit` (rev 8 — the commit the
>   Orchestrator captured **immediately before invoking Stage**; see below), `stage_head_commit`,
>   `stage_outcome` (`shipment` or `no-shipment`), `stage_artifact_paths` (the **concrete
>   repository-relative file paths** Stage wrote and committed between `stage_base_commit` and
>   `stage_head_commit` — file paths only, never directory prefixes; rev 7/8), and — when
>   `stage_outcome` is `shipment` — a `shipment_id` in `queued` status.

> **Before invoking Stage** (rev 8), capture and **retain** the pre-invocation tip:
> `stage_base_commit := git rev-parse HEAD`. This value is the Orchestrator's own, taken before
> Stage can mutate anything, and it is **authoritative**. Retaining it is what makes the
> verification in step 4 *aggregate*: every commit Stage creates during the invocation is a
> descendant of it, so one two-tree diff covers a Stage branch of **any** commit count.

> 4. Receive Stage's output: record the staging handback record, and the `shipment_id` when one
>    was returned. Apply these defaults for any field Stage did not report, recording
>    `STAGING_HANDBACK_DEGRADED: Stage did not report {fields} — defaulted` once and surfacing it
>    to the operator:
>    - `stage_branch` / `stage_head_commit`: resolve from the current checkout
>      (`git rev-parse --abbrev-ref HEAD`, `git rev-parse HEAD`). If that branch is the default
>      branch, leave `stage_branch` **UNRESOLVED** — **never** substitute the default branch for
>      the Stage artifact branch.
>    - `stage_base_commit` (rev 8): use the **retained pre-invocation value** — it wins over any
>      reported value. Use the reported value only when no pre-invocation value was retained (a
>      direct Stage invocation the Orchestrator did not bracket). When neither exists, leave it
>      **UNRESOLVED**; resolution still does not halt, and step 4's verification rejects the
>      UNRESOLVED case loudly.
>    - `stage_outcome`: **preserved when Stage reported one** (rev 10). Derive it from whether a
>      `shipment_id` was returned **only when the field is absent from the handback**. A reported
>      outcome is never overwritten by the derivation, and a reported `shipment` outcome is
>      **never** silently rewritten to `no-shipment` because no `shipment_id` reached this step —
>      that case is an inconsistent pair and is rejected by the validation below, not defaulted
>      away.
>    - `stage_artifact_paths`: leave **empty** here and derive it in step 4 from the
>      `stage_base_commit`→`stage_head_commit` aggregate (correction 4, check 4). It is **never**
>      defaulted to a directory list (rev 7).

**Two path notions, deliberately distinct (rev 7).** `STAGE_ARTIFACT_ROOTS` is a **fixed constant**
of the gate — `.backlogit/`, `docs/plans/`, `docs/decisions/`, `docs/memory/` — and is *not* a
handback field. It serves two roles and only two: it is step 1's dirtiness pathspec, and it is the
containment allow-list every concrete path in step 4 must lie under. `stage_artifact_paths` is a
handback field and is always a set of **concrete file paths**. Rev 6 conflated the two, using one
field for both jobs; that conflation is what let directory prefixes reach a verification loop that
had no way to reject them. Directory prefixes are correct in step 1 — it runs **before** the commit
exists, so there is no commit to derive files from — and are rejected in step 4, where the commits
do exist and the files they changed are knowable exactly.

**Two commit notions, also deliberately distinct (rev 8).** `stage_base_commit` is the
Orchestrator's **pre-invocation** tip; `stage_head_commit` is the Stage branch tip **after** the
invocation. Rev 7 carried only the latter and derived its file set from that one commit's own diff
against its parent, which silently assumed **a Stage session commits exactly once**. That
assumption is false in the general case and false in this very repository: the branch carrying this
plan has accumulated many commits. Under rev 7 an artifact written in an earlier commit and never
re-touched by the tip was covered by the **ancestry** check alone — which proves the *commit*
reached `origin/main` but never reads the *file* there, exactly the gap rev 7 said the per-file
check existed to close. The pair of commits closes it: the aggregate set is everything the Stage
session changed, not merely what its last commit changed.

**The degraded path defaults; it does not halt** (the rev-2 lesson, deliberation §8.2). Every field
has a deterministic fallback, so a Stage invocation predating B3 — which reports nothing and leaves
the worktree on the default branch — still completes: `stage_branch` stays **UNRESOLVED**, and
Step 1.5's resolution block routes that case straight to step 3, whose 3a create-arm builds the
artifact branch from the current commit. That is exactly the pre-change behaviour, so nothing
regresses, 3a's create-arm stays **live rather than dead prescription**, and the one substitution
that caused this defect is still refused. B3 supplies the producer, so the degraded path is the
exception rather than the steady state.

**Correction 2 — resolve the staging branch BEFORE the dirty/clean split, and widen the dirtiness
pathspec.** The resolution must sit **above step 1**, not inside step 2: step 1 routes a dirty tree
straight to step 3, so anything placed in step 2 is skipped on the very path that mutates the
repository. Add this block immediately after the preamble, and revise step 1:

> **Resolve the staging branch first.** Set `STAGE_BRANCH` from the Step 1 handback record.
> When it is resolved (not UNRESOLVED):
> `git rev-parse --verify "$STAGE_BRANCH"` — if this fails, halt with
> `STAGING_GATE_FAIL: staging branch {stage_branch} does not exist`
> `git rev-parse --abbrev-ref HEAD` — MUST equal `$STAGE_BRANCH`; if it does not, halt with
> `STAGING_GATE_FAIL: worktree is on {head} but Stage reported {stage_branch}` (this is also the
> P-016 single-worktree gate: exactly one worktree, on the branch Stage reported).
> When `STAGE_BRANCH` is UNRESOLVED, record that and apply **no** branch assertion; steps 1 and 2
> still run unchanged, and any work they find routes to step 3a's create-arm, which builds the
> artifact branch from the current commit.

> 1. Check `git status --short -- {STAGE_ARTIFACT_ROOTS}` — the **fixed** constant
>    `.backlogit/ docs/plans/ docs/decisions/ docs/memory/`, never a handback-supplied path list
>    (rev 7) — for uncommitted staging artifacts:
>    - If dirty: staging artifacts need to be committed first (proceed to step 3).
>    - If clean: proceed to step 2.

Both guards now execute on **every** path, including the dirty one, so 3a never checks out an
unverified branch name and the P-016 assertion is not bypassed in the one scenario where the
Orchestrator mutates the repo. Leaving the UNRESOLVED case inside the normal step-1/step-2 flow
(rather than short-circuiting it to step 3) is what preserves **pre-change semantics exactly** for a
legacy Stage run: with `HEAD` on the default branch, step 2's range is the same comparison Step 1.5
made before this change, so an already-committed-but-unpushed default-branch commit — the `fdff9e4`
shape — is still discovered and still routed to step 3 rather than silently skipped.

**Correction 3 — discovery becomes branch-specific.** Revised step 2:

> 2. Check for unpushed commits **on the checked-out staging branch** (`$STAGE_BRANCH`, resolved
>    and verified above):
>    `git fetch origin main`
>    `git log origin/main..HEAD --oneline`
>    - If output is empty: the staging branch carries nothing beyond the remote default branch.
>      Proceed to step 4.
>    - If output is non-empty: staging commits exist that are not on the remote. Proceed to step 3.
>
>    The comparison **must** be taken against the checked-out staging branch, never against the
>    default branch as a fixed endpoint — Stage commits on `$STAGE_BRANCH`, so a default-branch
>    self-range is empty by construction and reports a false "synchronized".

`origin/main..HEAD` and `origin/main..$STAGE_BRANCH` are the **same range** whenever `$STAGE_BRANCH`
is resolved, and that equality is *proven*, not assumed: the `git rev-parse --abbrev-ref HEAD`
assertion in the resolution block above is what makes `HEAD` a legitimate stand-in for the reported
branch. Writing the range against `HEAD` rather than against the branch name is what lets the one
command serve both populations — it is branch-specific discovery for a resolved Stage branch, and
it degrades to exactly the pre-change comparison for a legacy run whose `HEAD` is still the default
branch, so neither case is skipped. Either literal satisfies AC-B1.6; the `HEAD` form is written
because it composes with the clean-tree and single-worktree gates asserted on the same branch.

**Self-match hazard — the corrected text must NOT quote the forbidden range (rev 6).** The obvious
way to write correction 3 is a prohibition naming the literal `origin/main..main`. That would
reintroduce, in the very file this gate protects, exactly the failure documented in
`docs/compound/2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md` and already
guarded against at §5.0 and AC-C2.3: the B1 harness asserts the literal is **absent** from Step 1.5,
and a prohibition quoting it would match itself and make that assertion unsatisfiable — or force it
to be weakened into something that no longer rejects the regression. The prohibition is therefore
stated **descriptively** ("the default branch against its own remote-tracking ref"), and the
absence assertion stays a sound literal check. This constraint is load-bearing and is carried by
AC-B1.6.

**Correction 4 — step 4 gains a second, shipment-free verification arm.** Revised step 4:

> 4. **Validate the outcome/shipment pair before selecting an arm** (rev 10). Resolve
>    `stage_outcome` per step 1's rule — reported value preserved, derived only when absent — then,
>    **before** either arm runs:
>    - `stage_outcome` MUST be exactly `shipment` or `no-shipment`. Any other value, including an
>      empty one that no derivation could resolve, halts with `STAGING_GATE_FAIL: unknown
>      stage_outcome {stage_outcome}`.
>    - When `stage_outcome` is `shipment`, a **non-empty** `shipment_id` MUST be present. Otherwise
>      halt with `STAGING_GATE_FAIL: stage_outcome is shipment but no shipment_id was reported`.
>    - When `stage_outcome` is `no-shipment`, **no** `shipment_id` may be present. Otherwise halt
>      with `STAGING_GATE_FAIL: stage_outcome is no-shipment but shipment_id {shipment_id} was
>      reported`.
>
>    Only a validated pair selects an arm. A reported `shipment` outcome is **never** reclassified
>    as a successful `no-shipment` run.
>
>    **When `stage_outcome` is `shipment`:**
>    Verify the shipment manifest exists on the remote default branch:
>    `git show origin/main:.backlogit/queue/{shipment_id}.md`
>    - If the file exists: staging artifacts are confirmed on the remote. Proceed to Step 2.
>    - If the file does not exist: halt with `STAGING_GATE_FAIL: shipment manifest {shipment_id}
>      not found on origin/main`.
>
>    **When `stage_outcome` is `no-shipment`:**
>    There is no manifest to name, so verify the **commit range** and **every concrete file the
>    Stage session changed across it**. Run these six checks in order; any failure halts the gate.
>
>    1. `git rev-parse --verify {stage_head_commit}^{commit}` — halt with
>       `STAGING_GATE_FAIL: stage_head_commit {stage_head_commit} is not a commit in this
>       repository`.
>    2. `stage_base_commit` MUST be resolved and MUST be a commit:
>       `git rev-parse --verify {stage_base_commit}^{commit}` — halt with
>       `STAGING_GATE_FAIL: stage_base_commit is unresolved or is not a commit in this repository`.
>    3. `git merge-base --is-ancestor {stage_base_commit} {stage_head_commit}` — halt with
>       `STAGING_GATE_FAIL: stage_base_commit {stage_base_commit} is not an ancestor of
>       stage_head_commit {stage_head_commit}`. This is what makes the pair a **range** rather
>       than two unrelated commits.
>    4. `git merge-base --is-ancestor {stage_head_commit} origin/main` — halt with
>       `STAGING_GATE_FAIL: stage commit {stage_head_commit} has not reached origin/main`.
>    5. Derive the **aggregate** set of Stage artifact files the session changed:
>       `git diff --name-only -z --diff-filter=d {stage_base_commit} {stage_head_commit} --
>       {STAGE_ARTIFACT_ROOTS}`. Call the result `AGGREGATE`. If `AGGREGATE` is **empty**, halt
>       with `STAGING_GATE_FAIL: no Stage artifact files changed between {stage_base_commit} and
>       {stage_head_commit} under the allowed Stage artifact roots`.
>    6. Resolve `VERIFY_SET`, then read every member on the default branch:
>       - When the handback reported a **non-empty** `stage_artifact_paths`, every reported path
>         MUST be a repository-relative **file** path — no trailing `/`, not a bare
>         `STAGE_ARTIFACT_ROOTS` entry, no glob or wildcard — and MUST lie under
>         `STAGE_ARTIFACT_ROOTS`; otherwise halt with `STAGING_GATE_FAIL: stage_artifact_paths
>         contains a non-file or out-of-root path: {path}`. The reported set MUST then be
>         **equal** to `AGGREGATE` (same members, order-insensitive); otherwise halt with
>         `STAGING_GATE_FAIL: stage_artifact_paths does not match the files changed between
>         {stage_base_commit} and {stage_head_commit}`. `VERIFY_SET` is that set. When no paths
>         were reported — none were supplied, or step 3a discarded them after creating the
>         artifact commit — set `VERIFY_SET := AGGREGATE`.
>       - For **every** path in `VERIFY_SET`: `git cat-file -t "origin/main:{path}"` MUST exit 0
>         **and** print exactly `blob`. Halt with `STAGING_GATE_FAIL: Stage artifact {path} not
>         found on origin/main` on a non-zero exit, and with `STAGING_GATE_FAIL: Stage artifact
>         {path} is a {type}, not a file, on origin/main` on any other printed type.
>
>    When all six pass: every file the Stage session wrote is confirmed on the remote **as a
>    concrete file**. **Stop here — there is no shipment, so there is nothing to route to Ship.**

This is the **concrete artifact verification target that requires no shipment ID**: ancestry proves
the Stage commits reached the default branch, and the per-file type check proves the artifacts
themselves are readable **files** there rather than merely that some merge occurred.

**Why the path set must be concrete files, not directory prefixes (rev 7).** Rev 6 wrote this arm
as `git show origin/main:{path}` over `stage_artifact_paths`, whose documented default was the
directory list `.backlogit/ docs/plans/ docs/decisions/ docs/memory/`. `git show` on a **tree**
exits **0** and prints a directory listing — verified against this repository:
`git show origin/main:docs/plans/` prints `tree origin/main:docs/plans/` and returns status 0.
Those four directories already exist on `origin/main` in every repository this gate will ever run
in, so the loop passed **without reading a single file the Stage run wrote**. Paired with the
degraded `stage_head_commit` default — `git rev-parse HEAD`, which on a post-merge checkout is
routinely already an ancestor of `origin/main` — check 2 is trivially true as well, and the arm
reports a **false pass**. This is rev 6's own defect one layer down: rev 6 fixed *"the loop can run
zero times"* and left *"the loop can run only over things that are always there."* Three properties
close it, and all three are load-bearing:

1. **Concreteness is structural, not stylistic.** `git diff --name-only` between two trees emits
   **only blob paths** — a directory cannot appear in `AGGREGATE` by construction, so the derived
   set is concrete whether or not anyone remembers to check it. `-z` NUL-delimits the output so
   paths are never quoted or backslash-escaped, and `--diff-filter=d` excludes **deletions**, so a
   Stage artifact the session *removed* or renamed away is not demanded to be readable on
   `origin/main` (the `HEAD` of this very branch is a rename of that shape). The trailing
   `-- {STAGE_ARTIFACT_ROOTS}` pathspec applies the containment allow-list at the source rather
   than by post-filtering, so an out-of-root path cannot enter the set at all.
2. **`git cat-file -t` replaces `git show` in this arm.** Both exit 0 on a tree, so the exit code
   alone can never discriminate a file from a directory; the discriminator is the **printed object
   type**. Requiring exactly `blob` rejects a directory even if one reached `VERIFY_SET` through
   the reported-path branch, and `cat-file -t` still exits non-zero (128) when the path is absent,
   so presence and file-ness are proven by one command.
3. **Set equality, not subset.** Equality rejects a reported path the session never touched (an
   invented or copy-pasted path) **and** a reported set that silently drops files — which would
   otherwise let a handback shrink the verified set down to one always-present file. A subset check
   catches only the first.

**Why the set is the aggregate base→head diff, and why the range form is still refused (rev 8).**
A **two-tree** diff over `{stage_base_commit} {stage_head_commit}` is a property of the two commit
objects' trees: it reads **identically before and after the merge**, so it is recoverable when
step 4 runs. A **range** form such as `origin/main..{stage_head_commit}` is **empty by
construction** once that commit is an ancestor of `origin/main` — precisely the state step 4 runs
in — so a range-derived set would collapse to empty and re-create the vacuous-loop failure rev 7
exists to remove. The two-tree form has the post-merge stability of rev 7's single-commit
`diff-tree` **and** the session-wide coverage rev 7 lacked; that is the whole reason rev 8 changes
the derivation rather than merely widening it.

**Rev 7's "scoping to the tip loses nothing" argument is WITHDRAWN.** Rev 7 held that the per-file
check could safely read only `stage_head_commit`'s own changes, because check 2 already proved
every **ancestor** reached `origin/main`. That is a **category error**: commit ancestry proves the
*commit* landed, not that the *file* is readable on the default branch — which is exactly the gap
rev 7 introduced the per-file check to close. Applying the per-file check to the tip alone leaves
every artifact written in an earlier commit of the same session covered **only** by the property
rev 7 itself judged insufficient. A Stage session that writes its deliberation in commit 1 and its
plan in commit 2 therefore had the deliberation verified by nothing stronger than ancestry. The
aggregate set removes the asymmetry: every file the session changed is read on `origin/main`,
whichever commit changed it.

**Known widening, recorded not hidden (rev 8).** If a Stage session merges `origin/main` into its
artifact branch mid-flight, the two-tree diff also reports files that arrived from the default
branch rather than from Stage. The `-- {STAGE_ARTIFACT_ROOTS}` pathspec bounds that to the Stage
artifact roots, and every such file is by definition already on `origin/main`, so the per-file
check passes for it. The effect is therefore a few **extra** verified paths, never a missed one —
a strictly conservative failure direction. It is recorded as **R14** rather than engineered away,
because a first-parent or merge-aware derivation would trade this harmless over-coverage for the
under-coverage rev 8 exists to remove.

**The handback-resolution no-halt invariant is preserved (AC-B1.5).** Every new halt lives in
**step 4's verification**, not in the handback defaulting. All fields still resolve to a
deterministic default without halting; what changed is that a degraded resolution now yields a set
that is either concretely verifiable or **provably empty**, and an empty one fails loudly instead
of passing silently. A gate that halts because it could not prove the artifacts are on
`origin/main` is the gate working, not the degraded path wedging.

**No new command shape enters the corpus.** `diff`, `cat-file`, `rev-parse`, and `merge-base`
are read forms: none is a push verb, so §5.3 construct 1 cannot match them, and none grants
permission to commit anywhere, so construct 2 cannot either. The detector's closed construct set
and the fixture corpus are therefore **not** widened — the rev-6 non-expansion position is
unchanged. The shipment arm stays byte-verbatim (AC-B1.4), the resolution block's clean-tree and
single-worktree guards are untouched (AC-B1.9), and P-009, P-011, and P-016 remain textually
unchanged (AC-C2.6).

**Correction 5 (rev 7, extended rev 8) — step 3 re-records the head and PRESERVES the base.** When
step 1 finds a dirty tree, **3a creates the commit that carries the artifacts** — and the
handback's `stage_head_commit` was resolved *before* that commit existed. Deriving step 4's file
set from the stale value would read a commit that changed none of the artifacts, halting the gate
for the wrong reason on the one path Step 1.5 exists to repair. 3a therefore:

- re-records `stage_head_commit` as `git rev-parse HEAD` **after** committing;
- **preserves `stage_base_commit` unchanged** (rev 8). The base predates 3a's commit and is
  therefore still the correct lower bound; re-recording it to the post-commit `HEAD` would collapse
  the range to nothing, and re-recording it to the *pre-commit* `HEAD` would silently discard any
  already-committed unpushed Stage commits step 2 routed into step 3 — reproducing the tip-only
  blind spot rev 8 removes, on the one path most likely to carry several commits. When
  `stage_base_commit` is still **UNRESOLVED** at 3a — a direct invocation the Orchestrator never
  bracketed — 3a sets it to the `HEAD` captured **before** its own commit, which is the tightest
  sound lower bound available at that point;
- discards any reported `stage_artifact_paths` (they described the pre-commit intent, not the
  committed range) so check 6 falls through to `VERIFY_SET := AGGREGATE`, **recomputed from the
  preserved base to the new head** rather than from the tip commit alone.

When step 3 was entered for unpushed commits alone and nothing was uncommitted, `HEAD` is unchanged
and the re-record is a no-op. AC-B1.3's branch-point text (`from the current commit`) is preserved
verbatim.

**Correction 6 (rev 10) — the outcome field must not fail open.** Revisions 6–9 wrote the
`stage_outcome` default as an *unconditional derivation*: `shipment` when a `shipment_id` was
returned, `no-shipment` otherwise. That reads as a harmless fallback and is not one. It has no
`absent` branch, so it overwrites a **reported** value with a derived one, and the derivation's
`otherwise` clause is a **success** terminal. The failure is therefore silent and inverted:

| Reported by Stage | `shipment_id` reached the Orchestrator | Rev 6–9 resolution | Consequence |
|---|---|---|---|
| `shipment` | yes | `shipment` | correct |
| `shipment` | **no** (dropped, truncated, mis-parsed handback) | **`no-shipment`** | the shipment arm is **skipped**, the no-shipment arm runs and **passes** (the artifacts really are on `origin/main`), the gate reports success, and **Ship is never routed** — a formed shipment silently abandoned |
| `no-shipment` | n/a | `no-shipment` | correct |
| *(absent)* | either | derived | the intended use of the rule |

Row 2 is the whole finding. The no-shipment arm is *designed* to pass for a run whose artifacts
reached the default branch, so it cannot be relied on to notice that a shipment was expected; it
verifies files, not intent. The derivation converts the loudest failure the pipeline has — a
shipment that was formed but not handed off — into a clean terminal, and does it on the one input
class most likely to be degraded, since a handback that lost its `shipment_id` is by definition a
handback that was not fully received.

Two changes close it, and they are deliberately at **different** points in the flow:

1. **Preservation, in the defaulting phase.** The derivation is scoped to the **absent** case. A
   reported value wins. This is a strictly weaker rule than the one it replaces, so the
   never-halts invariant is untouched — resolution still cannot halt, and a handback that reports
   nothing still resolves exactly as before.
2. **Validation, in the verification phase.** The pair check above sits in **step 4**, alongside
   the other `STAGING_GATE_FAIL` halts, and runs **before** arm selection. Placing it here rather
   than in the defaulting block is what preserves AC-B1.5's "all halts live in verification, never
   in field defaulting" invariant — the same placement discipline rev 6 arrived at after its first
   draft hard-halted the degraded population it claimed not to wedge (§12, rev-6 internal defect 1).

**Why validation and not a third arm.** An inconsistent pair is not a third outcome needing its own
verification; it is evidence that the handback is **untrustworthy**, and every downstream check
reads that same handback. Continuing on either arm would verify a claim whose provenance has
already been contradicted. Halting is the fail-closed response, and it is loud: the operator sees
which half disagreed, and a re-invocation with a complete handback proceeds normally.

**No new command shape, no widened detector.** The pair validation is a comparison between two
already-declared fields. It adds no git command, no push verb, and no permission to commit
anywhere, so §5.3's closed construct set and the 14-fixture corpus are unchanged — the rev-6/7/8/9
non-expansion position, unchanged again.

**The rev-2 "no no-shipment clause here" position is WITHDRAWN.** Rev 2 argued that Step 1.5 is
scoped "After Stage completes and before routing to Ship" and that step 4 requires a
`{shipment_id}`, so a shipment-less run never enters the gate and the obligation belongs in B2/B3
instead. That reasoning was internally inconsistent with §6.2's own rev-4 route table, which names
the **Orchestrator (pipeline invocation)** as the actor for pushing, opening, merging, and
post-merge-verifying the no-shipment branch — and Step 1.5 is the **only** Orchestrator gate that
does any of those things. Rev 2 thus assigned the Orchestrator a duty and simultaneously declared
the one place it could discharge it out of scope. Corrections 1–5 give that duty a home; B2/B3 keep
the Stage-side obligations they already carry. Nothing is duplicated: B2/B3 say *where Stage may
write*, §6.1.2 says *how the Orchestrator finds and verifies it*.

**No direct default-branch push is introduced.** Corrections 1–5 add only read forms
(`status`, `rev-parse`, `fetch`, `log`, `show`, `merge-base`, `cat-file`, and — rev 8 — `diff`)
and one `checkout`; 3b's existing push-the-branch-and-open-a-PR form is untouched, and
3e stays deleted. P-009, P-011, and P-016 are textually unchanged (AC-C2.6).

**AC-B1.1** No direct-push-to-default-branch instruction remains anywhere in Step 1.5.
**AC-B1.2** The preamble states the postcondition as artifacts having **reached** the default
branch **via a merged staging PR** (rev 3's fourth surface — §2), and its scope clause binds
**whether or not a shipment was formed** (rev 6).
**AC-B1.3** 3a textually carries the branch point (`from the current commit`) for the
already-committed-but-unpushed path — textual, not structural, because the a–e sub-steps are lazy
paragraph continuations, not a real nested list (see §11 on the pre-existing list-rendering defect).
**AC-B1.4** Step 4's **shipment arm** preserves, verbatim, the pre-change lead sentence
(`Verify the shipment manifest exists on the remote default branch:`, capitalization included), the
`git show origin/main:.backlogit/queue/{shipment_id}.md` command, both outcome bullets, and the
`STAGING_GATE_FAIL: shipment manifest {shipment_id} not found on origin/main` halt string, with
pass/fail semantics unchanged. **Rev 6 scope refinement**: the criterion previously required the
whole step-4 block to be byte-identical. That form is no longer satisfiable, because step 4 must
gain the `no-shipment` arm (AC-B1.8) — and an unmodified block is precisely what leaves that path
unverified. The criterion is narrowed to what it was always protecting: the authoritative
post-merge gate is preserved and unrelaxed. The step may gain **outcome-arm label lines** and the
second arm, **and nothing else** — no outer lead sentence replaces or rewords the preserved one.
**AC-B1.5** **Staging handback recorded, with total defaults.** Step 1 declares and step 4 records
`stage_branch`, `stage_base_commit` (rev 8), `stage_head_commit`, `stage_outcome`, and
`stage_artifact_paths`. **Every** field has a deterministic default, so the degraded path **never
halts**: branch/head resolve from `HEAD`; `stage_base_commit` resolves to the value the
Orchestrator **captured and retained immediately before invoking Stage** — authoritative over any
reported value — falling back to the reported value only when nothing was retained, and otherwise
left UNRESOLVED without halting; `stage_outcome` is **preserved when reported** and derived from
whether a `shipment_id` was returned **only when it is absent** (rev 10 — the derivation never
overwrites a reported value);
`stage_artifact_paths` is left **empty** and derived in step 4 from the
`stage_base_commit`→`stage_head_commit` aggregate (rev 7/8 — it is **never** defaulted to a
directory list); `STAGING_HANDBACK_DEGRADED` is recorded once and surfaced. When the `HEAD`
resolution yields the default branch, `stage_branch` is left **UNRESOLVED** and the run routes to
step 3a's create-arm. The default branch is **never** substituted for the Stage artifact branch.
**AC-B1.6** **Branch-specific discovery.** Step 1.5's unpushed-commit check compares
`origin/main..HEAD` — equivalently `origin/main..$STAGE_BRANCH` whenever a branch is resolved,
proven equal by the `git rev-parse --abbrev-ref HEAD` assertion in the resolution block, and
degrading to exactly the pre-change comparison when it is UNRESOLVED, so neither population is
skipped. The literal default-branch self-range is **absent from the whole of Step 1.5**, including
from the prohibition sentence, which is phrased descriptively so the absence assertion cannot
self-match.
**AC-B1.7** **Dirtiness pathspec covers Stage artifacts.** Step 1's `git status` pathspec is the
**fixed** `STAGE_ARTIFACT_ROOTS` constant — `.backlogit/ docs/plans/ docs/decisions/
docs/memory/`, not `.backlogit/` alone — so a no-shipment run that writes only `docs/**` is not
reported clean. **Rev 7**: the pathspec is that constant and **not** the handback's
`stage_artifact_paths`, so a handback reporting one narrow path cannot shrink the dirtiness scan
and hide uncommitted artifacts. Directory prefixes are correct *here* — step 1 runs before the
artifact commit exists, so there is no commit to derive files from — and are rejected in step 4,
where a commit does exist.
**AC-B1.8** **No-shipment verification arm.** Step 4 carries a second arm keyed on
`stage_outcome: no-shipment` that runs **six** ordered checks and halts with `STAGING_GATE_FAIL`
on any failure: (1) `git rev-parse --verify {stage_head_commit}^{commit}`; (2) `git rev-parse
--verify {stage_base_commit}^{commit}`, rejecting an UNRESOLVED base; (3) `git merge-base
--is-ancestor {stage_base_commit} {stage_head_commit}`; (4) `git merge-base --is-ancestor
{stage_head_commit} origin/main`; (5) derivation of `AGGREGATE` from `git diff --name-only -z
--diff-filter=d {stage_base_commit} {stage_head_commit} -- {STAGE_ARTIFACT_ROOTS}`, halting when
it is **empty**; (6) resolution of `VERIFY_SET` per AC-B1.10 followed by `git cat-file -t
"origin/main:{path}"` for **every** path in it, requiring exit 0 **and** output exactly `blob`. It
terminates the gate without routing to Ship and names no `shipment_id`. **Rev 7**: `git show
origin/main:{path}` no longer appears in this arm — it exits 0 on a tree, so it could not reject
directory prefixes. **Rev 8**: the single-commit `git diff-tree` derivation no longer appears
either — it covered only the tip commit — and the `origin/main..{commit}` range form remains
refused, being empty by construction at the moment this arm runs.
**AC-B1.9** **Clean-tree and single-worktree gates preserved on every path.** The branch-resolution
block sits **above** step 1's dirty/clean split, so `git rev-parse --verify` and the
HEAD-equals-`$STAGE_BRANCH` assertion execute on the dirty path too — 3a never checks out an
unverified branch name. Step 1.5 halts when HEAD is not the resolved staging branch; no new branch
or worktree is created when the handback reports one; and P-009, P-011, and P-016 remain textually
unchanged.
**AC-B1.10** **Concrete-path contract (rev 7, aggregate at rev 8).** `stage_artifact_paths` is
defined as a **non-empty set of concrete repository-relative file paths changed between
`stage_base_commit` and `stage_head_commit`**. Step 4 rejects, by halting: a path with a trailing
`/`, a bare `STAGE_ARTIFACT_ROOTS` entry, or any glob/wildcard form (not a file path); a path not
under `STAGE_ARTIFACT_ROOTS` (outside the allowed Stage artifact roots); an **empty** resolved set;
any reported set that is not **equal** to `AGGREGATE`, the aggregate base→head changed-file set
(equality, not subset — a subset check would accept both an invented path and a silently shrunken
set); and any path that resolves on `origin/main` to an object type other than `blob`. No directory
prefix can satisfy completion on this arm, by construction: a two-tree `git diff --name-only` emits
blob paths only, the `-- {STAGE_ARTIFACT_ROOTS}` pathspec bounds containment at the source, and
`cat-file -t` must print exactly `blob`.
**AC-B1.11** **Handback commits re-recorded and preserved across step 3 (rev 7, extended rev 8).**
3a sets `stage_head_commit` to `git rev-parse HEAD` after committing and discards any reported
`stage_artifact_paths`, so step 4 derives its file set from the commits that actually carry the
artifacts rather than from the pre-commit `HEAD` the handback resolved. **`stage_base_commit` is
preserved unchanged** across 3a — or set to the `HEAD` captured *before* 3a's commit when it was
UNRESOLVED — so the recomputed set is the **aggregate over the preserved base**, never the tip
commit alone. The re-record is a no-op when step 3 was entered for unpushed commits alone.
AC-B1.3's `from the current commit` branch-point text is preserved verbatim.
**AC-B1.12** **Fail-closed outcome/shipment pair (rev 10).** `stage_outcome` is **preserved** when
the handback reports it and derived from `shipment_id` presence **only when it is absent**. Before
step 4 selects an arm, the resolved pair is validated and any failure halts with
`STAGING_GATE_FAIL`: an outcome that is neither `shipment` nor `no-shipment` (including an
unresolvable empty value) halts as an unknown outcome; `shipment` with an empty or missing
`shipment_id` halts; `no-shipment` with a `shipment_id` present halts. A reported `shipment`
outcome is **never** reclassified as a successful `no-shipment` run. The validation lives in
step 4's **verification**, not in the handback defaulting, so AC-B1.5's never-halts resolution
invariant is preserved intact.

**Rev 4 reconciliation**: this list previously carried only three criteria, worded before rev 3
added the preamble as the fourth authorization surface, while task `018.007-T` already carried the
four-criterion form and §6.3's preserved-form note already referenced **AC-B1.4**. The plan is
corrected to the task's form — the same plan↔task same-contract defect class as the C2 divergence
reconciled in §7.2, found by the rev-4 AC-ID parity sweep rather than by review.

**Rev 6 reconciliation**: AC-B1.5–AC-B1.9 are new and are mirrored into `018.007-T` in the same
change, so the plan↔task same-contract invariant holds. AC-B1.4's narrowing is recorded above
rather than silently applied, because a criterion that is weakened without a stated reason is
indistinguishable from one that was abandoned.

**Rev 7 reconciliation**: AC-B1.10 and AC-B1.11 are new, and AC-B1.5, AC-B1.7 and AC-B1.8 are
**strengthened** (never relaxed) — they now reject the directory-prefix path set rev 6 permitted.
All five are mirrored into `018.007-T` in the same change, preserving the plan↔task same-contract
invariant.

**Rev 8 reconciliation**: no criterion is added or removed — the B1 contract stays at
AC-B1.1–AC-B1.11. AC-B1.5, AC-B1.8, AC-B1.10 and AC-B1.11 are **strengthened** to the aggregate
base→head form; AC-B1.1–AC-B1.4, AC-B1.6, AC-B1.7 and AC-B1.9 are untouched. All four amended
criteria are mirrored into `018.007-T`, and that task's **description is
rewritten** so it states this contract once, in its current form, rather than layering a rev-7
addendum over operative-looking rev-6 prose (adversarial-review blocker 3).

**Rev 10 reconciliation**: one criterion is added — **AC-B1.12**, the fail-closed outcome/shipment
pair — and **AC-B1.5** is strengthened to preserve a reported `stage_outcome`. AC-B1.1–AC-B1.4 and
AC-B1.6–AC-B1.11 are untouched. Both are mirrored into `018.007-T` in the same change, so the
plan↔task same-contract invariant holds. The B1 contract is now AC-B1.1–AC-B1.12.

### 6.2 Task B2 — P-010 + Amendment Log

**Stage MAY** — replace the default-branch bullet (actor named, per deliberation §8.2):

> - Commit backlog and planning artifacts to a **dedicated Stage/admin branch** — never to the
>   default branch — **including when no shipment is formed**. Push, PR creation, and merge for
>   that branch are performed by the **Orchestrator under Step 1.5**, or by the **operator** in a
>   direct Stage invocation; **never by Stage**.

**Stage MUST NOT** — add, mirroring the existing Ship bullet:

> - Commit or push Stage artifacts directly to `main`

Stage's existing "Create, push, or merge pull requests" prohibition is **left intact**. The actor
clause is what makes the new rule satisfiable without weakening it.

Add a clarifying sentence that Step 1.5 staging orchestration is an **Orchestrator-owned gate**,
not delegated Ship work — so it does not collide with P-010's "Orchestrator must not perform Stage
or Ship work directly" — **and** that a Stage artifact branch/PR, carrying no source, test, or
config change, is **not an implementation branch** for P-016 or P-001 purposes (rev 3; see §10.2).

**Stage MAY** also gains an explicit branch-creation grant (rev 3), since no existing text names
who creates the branch in the no-shipment case:

> - Create and check out the dedicated Stage/admin artifact branch `chore/stage-{scope-slug}`,
>   where `{scope-slug}` is derived from the scope under staging per P-011's existing slug
>   convention. A shipment ID **may** be used as the slug when one is already known; it is **not
>   required**, and Stage never waits for one.

**Naming must not depend on a shipment ID (rev 9).** Revisions 1–8 wrote this grant as
`chore/stage-{shipment_id}`, with `chore/stage-{chore-slug}` as a shipment-less fallback. That form
is **underivable at the moment the branch is needed**: the branch must exist before Stage's first
artifact write (§6.3.1), and the shipment is not created until Step 5.5 — the *last* mutation of the
session. A shipment-forming run following the rev-8 wording literally therefore had no name to use
at gate time, and the wording pushed an executor toward deferring the branch until after the very
mutations it exists to protect. `{scope-slug}` is derivable from the selected grouping's covering
feature title (task-shaped intake) or the feature/epic/chore title (feature-shaped intake) as soon
as Step 1/1.5 completes, so the name is stable for the whole session and identical on both
invocation paths. The shipment-ID form is preserved as a **permitted** slug value rather than
deleted, so an Orchestrator that already holds a shipment ID may still name the branch after it
(`chore/stage-017-S`). The branch carrying this plan, `chore/stage-pipeline-policy-gap`, is a
conforming example of the `{scope-slug}` form.

**Concrete no-shipment persistence route** (rev 4 — review round 3). Stating the actors end to end,
because "never to the default branch" is only actionable if the alternative route is executable:

| Step | Actor | Action |
|---|---|---|
| 1 | **Stage** | Create and check out `chore/stage-{scope-slug}` **before its first artifact write** (grant above; mirrored into the `_stage.agent.md` Git row by B3, which is the cell role enforcement actually reads, and made operative by B3's Step 1.9 branch gate — §6.3.1) |
| 2 | **Stage** | Commit the backlog/planning artifacts to that branch (operative as B3's Step 5.7, after every Stage mutation and before the Step 6 handback — §6.3.1) |
| 3 | **Orchestrator** (pipeline invocation) or **operator** (direct Stage invocation) | Push the branch, open the PR, merge it |
| 4 | **Orchestrator** or **operator** | Post-merge verification |

Stage's role ends at step 2. Steps 3–4 are **never** Stage actions — this is what keeps the new
obligation satisfiable without weakening Stage's standing PR prohibition.

**How step 3's actor learns the branch (rev 6).** This table names the actors but not the handoff,
and a route whose step-3 actor cannot *find* the branch produced by step 1 is not executable. The
handoff is specified in B1 (the Orchestrator's Step 1 **staging handback record**) and produced in
B3 (Stage's Step 6 summary contract emits it). P-010's text is unchanged by rev 6 — the handback is
a mechanism, not a permission — but the route above is only complete when read together with
§6.1.2 and §6.3. No P-010 acceptance criterion changes.

**Verification evidence, pre- vs post-merge** — the two `git show` forms are **not**
interchangeable and neither replaces the other:

- **Pre-merge**, on the staging branch: `git show {branch}:{path}` is **local evidence only**. It
  proves the artifact was committed; it proves nothing about the default branch.
- **Post-merge**, the authoritative gate: `git show origin/main:{path}` **remains unchanged and
  remains required**. Step 1.5 step 4 keeps this form byte-identical (task B1, AC-B1.4).

Correcting *where Stage may write* does not relax *where the Orchestrator must verify*. The
post-merge `origin/main` gate is the reason the branch route is safe, not a casualty of it.

**Amendment Log**: append exactly one row, version **1.25.0** (the log stands at 1.24.0 — rev 1
said only "bump minor"), Change cell "Corrected P-010", body ending "Corrects, and does not delete
or edit, the 1.5.0 row above".

**Acceptance criteria (rev 8 — reconciled to a single six-criterion contract).** Revisions 1–7
carried **two** divergent five-criterion lists: this section asserted Stage/Ship symmetry but
omitted the P-016/P-001 clarification, `018.008-T` asserted the clarification but omitted symmetry,
and the shared IDs AC-B2.3/AC-B2.4/AC-B2.5 were bound to **different** criteria on each surface —
so an executor satisfying one list would have failed the other while both claimed to be the same
contract. The union is six criteria; both surfaces now carry them verbatim, in this order and with
these IDs.

**AC-B2.1** No P-010 text authorizes Stage to commit or push on the default branch — no
default-branch allowance remains anywhere in P-010's Stage block.
**AC-B2.2** The rule holds explicitly **even when no shipment is formed**, and the branch name is
derivable **without** one: the grant names `chore/stage-{scope-slug}` and treats a shipment ID as a
**permitted but not required** slug value (rev 9).
**AC-B2.3** P-010's **Stage and Ship columns are symmetric** on direct default-branch writes: the
Stage block carries a prohibition mirroring the existing **Ship MUST NOT** bullet
"Commit or push directly to `main`".
**AC-B2.4** **No actor is both prescribed and prohibited** on the same operation (no
self-deadlock): every obligation the new bullet creates names an authorized actor, with pushing the
branch and opening/merging the PR attributed to the **Orchestrator** (pipeline invocation) or the
**operator** (direct Stage invocation) — never to Stage, whose standing PR prohibition is
unweakened.
**AC-B2.5** The **P-016/P-001 non-implementation clarification** is present in the **policy text**,
not only in this plan: a Stage artifact branch/PR carries no source, test, or config change and is
therefore neither an implementation branch under P-016 nor a release unit under P-001.
**AC-B2.6** The Amendment Log gains **exactly one additive row, version 1.25.0**, following the
existing format; 1.24.0 and every earlier row are unmodified.

**Rev 9 reconciliation**: no criterion is added or removed — the B2 contract stays at
AC-B2.1–AC-B2.6. **AC-B2.2 is corrected in place** to the shipment-ID-independent naming form, and
the grant bullet and route table above are corrected with it. Both are mirrored into `018.008-T` in
the same change, preserving the plan↔task same-contract invariant. The correction **strengthens**
the criterion: the rev-8 form was satisfiable by a policy whose branch name could not be computed
at the moment the branch is required.

### 6.3 Task B3 — `_stage.agent.md` Role Boundary

Rewrite the **Git** row (deliberation §8.1 — the operative surface):

| Column | New text |
|---|---|
| Allowed | Commit backlog/planning artifacts to a **dedicated Stage/admin branch**; **create and check out that dedicated artifact branch** (`chore/stage-{scope-slug}`; a shipment ID may be used as the slug when one is already known, but is never required); create/use an explicit, time-boxed spike/research worktree only for staging investigation |
| Forbidden | **Commit or push directly to the default branch**; create or checkout feature/chore branches for code execution; create/use parallel implementation branches or worktrees |

The **branch creation/check-out grant is load-bearing, not decorative** (rev 4 — review round 3).
`role-enforcement.instructions.md` makes this table — not P-010 — the permission set consulted at
mutation time (§2). B2 grants Stage `Create and check out the dedicated Stage/admin artifact
branch`; if that grant appears only in P-010 and not in this cell, role enforcement **fails
closed** on the very first action the corrected policy requires, and the no-shipment persistence
route becomes unexecutable. The two surfaces must carry the grant in both places.

Add a parenthetical to the **PR** row recording that branch **push**, PR creation, and merge are
performed by the **Orchestrator** (pipeline invocation, Step 1.5) or the **operator** (direct Stage
invocation) on Stage's behalf — keeping the prohibition intact while removing the apparent
deadlock. The PR row is the **only** row other than Git that this task may touch, and it may gain
**only** that actor note.

**Staging handback emission (rev 6).** Add to Stage's **Step 6 (Summary)** session-output contract
one required line emitting the **staging handback record** that the Orchestrator's Step 1 now
requires (§6.1.2, correction 1): `stage_branch`, `stage_base_commit` (rev 8), `stage_head_commit`,
`stage_outcome` (`shipment` | `no-shipment`), `stage_artifact_paths`, and `shipment_id` when one
was formed.

**Outcome and shipment ID are emitted as a consistent pair (rev 10).** The same line states that
`stage_outcome` is **always emitted explicitly** — never omitted for the consumer to infer — and
that the pair is internally consistent: `shipment` is emitted together with a non-empty
`shipment_id`, and `no-shipment` is emitted with none. This is the producing half of AC-B1.12. The
consumer's validation is a fail-closed **agreement** check, and an agreement check needs two
parties: if the producer may omit the outcome and let the consumer derive it, the derivation is
back — and with it the silent reclassification of a reported `shipment` run into a successful
`no-shipment` one (§6.1.2, correction 6). Producer and consumer move together here for the same
reason they did at rev 7 and rev 8.

**Path shape is part of that contract (rev 7, aggregate at rev 8).** The same line states that
`stage_artifact_paths` is a **non-empty list of concrete repository-relative file paths** — the
files Stage committed between `stage_base_commit` and `stage_head_commit`, never directory prefixes
— and that Stage produces it with the same derivation the Orchestrator re-runs:
`git diff --name-only -z --diff-filter=d {stage_base_commit} {stage_head_commit} --
.backlogit/ docs/plans/ docs/decisions/ docs/memory/`. Naming the producing command here
is what makes the consumer's **set-equality** check (AC-B1.10) satisfiable by construction rather
than by coincidence: two surfaces computing the same set from the same commit pair with the same
command cannot disagree. A producer left free to emit directory prefixes, or free to report only
its tip commit's files, would either fail that check on every run or force it to be weakened back
into a form that verifies less than the session wrote. Stage echoes `stage_base_commit` rather than
inventing one: the Orchestrator's retained pre-invocation value is authoritative (AC-B1.5), and the
echo exists so a **direct** Stage invocation — which no Orchestrator bracketed — still hands the
operator a usable base.

This was the **first** edit B3 makes outside the Role Boundary table — rev 8 adds the Step 5.5 /
Step 6 edits below and rev 9 the two operative steps of §6.3.1; the enumerated set is fixed by
AC-B3.5, AC-B3.6, and AC-B3.7 and by nothing else. It is not cosmetic. Without
a producer, the Orchestrator's handback requirement has no source and every pipeline run falls into
B1's `STAGING_HANDBACK_DEGRADED` path. It also completes the symmetry the paragraph above rests on:
B3 must carry the branch-creation grant because role enforcement reads *this file* at mutation
time — and by the same argument the Orchestrator consumes *this file's* declared session output, so
the branch Stage created must be declared here too. Placing the emission only in the Orchestrator's
expectations would repeat, in the opposite direction, the grant-in-one-surface-only defect AC-B3.2
exists to prevent. It transfers **no** authority: emitting a branch name is not pushing, opening, or
merging anything, so the PR-row prohibition is untouched.

**Making `no-shipment` reachable (rev 8) — the second Step 5.5/Step 6 edit.** Everything above, and
the whole no-shipment arm of §6.1.2, is **dead prescription** against the installed
`_stage.agent.md` as it stands. Two clauses in that file make `stage_outcome: no-shipment`
unreachable:

| Clause | Installed behaviour | Consequence |
|---|---|---|
| **Step 5.5** | Shipment assembly is "MANDATORY — not optional" whenever the registry advertises `features.shipments: true`; ending the session without a `shipment_id` is declared a **P-005 violation** | A run that legitimately produces nothing to ship cannot terminate compliantly |
| **Step 6** pre-summary gate | "If no `shipment_id` exists, **HALT** and go back to Step 5.5" | The summary — which is where AC-B3.5's handback line lives — is never reached, so no handback is ever emitted for a shipment-less run |

So B1 waits on an outcome no producer can emit. B3 therefore also makes the terminal outcome
executable, with the **narrowest** change that achieves it:

1. **Step 5.5** is scoped to a harvest that produced items. Its existing guardrail — *do not
   assemble a shipment if the harvest step produced no items or produced items with unresolved
   P-003 violations; halt and report before creating an empty shipment* — is preserved **verbatim**
   and is precisely what identifies the two cases: a **valid empty harvest** (nothing to decompose)
   terminates as `no-shipment`, while **unresolved P-003 violations** still **halt**.
2. **Step 6**'s pre-summary verification gate requires a `shipment_id` **only when `stage_outcome`
   is `shipment`**. When it is `no-shipment`, the gate instead requires the **complete** handback
   record and an explicit terminal statement that there is nothing to route to Ship.
3. The summary for a `no-shipment` run **must not** direct the operator to Ship. This preserves,
   rather than contradicts, the existing prohibition on handing Ship a feature ID instead of a
   shipment ID: with no shipment, the correct handoff is *none*, and the Orchestrator's Step 1.5
   no-shipment arm is the terminal gate.

**Failures stay failures — this is the load-bearing constraint.** `no-shipment` is a **success**
terminal for a run that had nothing to ship. It is **never** a re-label for failure. A P-003
lineage violation, a harvest that errored, or a harvest that produced items but could not form the
required shipment all continue to **halt** exactly as they do today, and MUST NOT be recorded as
`no-shipment`. Without that sentence the edit would convert the pipeline's loudest failure mode
into a silent success — a strictly worse defect than the unreachable arm it fixes.

#### 6.3.1 Operative branch gate and artifact commit (rev 9)

**The defect.** Revisions 1–8 corrected *what Stage may do* and *what Stage must report*. Neither
is an instruction to **do** it. The installed `_stage.agent.md` contains no branch operation and no
commit operation at any step: its Required Steps run triage → grouping → learnings → deliberation →
planning → review → harvest → shipment assembly → stash archival → summary, and its Step Sequence
Contract checklist — the file's own ordered, machine-readable statement of what a session must
execute — names none either.

| Rev 1–8 surface | What it establishes | What it does not |
|---|---|---|
| Role Boundary Git row (AC-B3.1/B3.2) | Stage **may** create, check out, and commit on the artifact branch | that any step ever does |
| P-010 grant (AC-B2.2) | the same permission, in policy | the same gap |
| Step 6 handback (AC-B3.5) | Stage **must report** `stage_branch`, `stage_base_commit`, `stage_head_commit`, `stage_artifact_paths` | where those values come from |

An executor satisfying AC-B3.1–AC-B3.6 to the letter can therefore complete a session **on the
default branch**, having written every artifact there — the exact `fdff9e4` shape — and then emit a
handback naming a branch that was never created and commits that were never made. The Orchestrator's
step-4 arm would reject that handback (the fields resolve to the default branch, so `stage_branch`
is left UNRESOLVED and the aggregate diff is empty), which is the gate working; but the artifacts
are already on the default branch by then, and rejection after the fact is not prevention. This is
the **producer/consumer failure class a fourth time**, in its purest form: rev 1 obliged Stage to
merge a PR it could not create; rev 6 obliged the Orchestrator to find a branch nothing announced;
rev 8 keyed a verification on an outcome nothing could emit; rev 9 finds a **reported value with no
producing action**.

**Correction 1 — Step 1.9, the pre-mutation branch gate.** Add a new step section between the
existing Step 1.8 (Learnings Retrieval) and Step 2 (Deliberation):

> ### Step 1.9: Stage Artifact Branch Gate (NON-NEGOTIABLE)
>
> Stage MUST NOT write a backlog, plan, decision, or memory artifact while `HEAD` is the default
> branch. This gate occupies a **fixed position** in the Step Sequence Contract — after Step 1.8
> and before Step 2 — which is the first point at which a stable scope slug is derivable for both
> intake shapes, Step 1 classification having settled the feature-shaped case and Step 1.5
> grouping selection the task-shaped one.
>
> **Deferral rule — NOTHING TRACKED WRITES BEFORE THIS GATE (rev 10).** This gate precedes
> **every** tracked Stage artifact mutation of the session, without exception. Any such mutation
> that would otherwise occur at an earlier step is **deferred until after** it. The rule is
> categorical; the enumeration below names the classes the installed contract actually schedules
> earlier, and is illustrative of the rule rather than a closed list that a new earlier mutation
> could escape:
>
> 1. **The write half of the mandatory late-identifier reconciliation** (Step 1). The
>    *analysis* — searching Ship's residual-risk records for a late `review-thread ID` or PR
>    number — is **read-only** and completes in Step 1. The **in-place update** of the deferred
>    stash entry that records the recovered identifiers is a tracked write and moves here.
> 2. **The archival of a duplicate stash entry** produced by the unconditional duplicate-detection
>    scan (Step 1). Detection is read-only and completes in Step 1; the `stash archive` call moves.
> 3. **Every tracked `docs/memory/` checkpoint whose triggering milestone precedes this gate.**
>    That is specifically the **stash-classification** checkpoint (Step 1) and the
>    **contextual-grouping / operator-selection** checkpoint (Step 1.5), and any other checkpoint
>    a session would write before reaching this step. All are written **after** the gate succeeds,
>    against the same session state they would have recorded earlier.
>
> **Read-only work still runs first, and must.** Step 1 classification and Step 1.5 grouping
> analysis and operator selection are *reads plus a decision*; they are what make the scope slug
> derivable, so they necessarily precede the gate. Deferring them would make the gate's own
> position underivable. What is deferred is the **write**, never the analysis or the decision.
>
> **Disposable, gitignored operations are NOT tracked artifacts and are NOT deferred.** Index
> synchronization (`backlogit sync` / `backlogit_sync_index` → `.backlogit/backlogit.db*`),
> structured backlogit checkpoints (`.backlogit/checkpoints/`), hook-event polling and
> acknowledgement (`.backlogit/hooks_queue.jsonl`), and `.backlogit/runtime/` state are all
> **gitignored in this repository** — they never appear in `git status`, cannot be committed, and
> therefore cannot reach the default branch. They are outside this rule, and this plan does **not**
> claim they are covered by it. Conflating them with tracked artifacts would overstate the gate's
> reach and would wrongly forbid the Step 0.1 index sync that every later query depends on.
>
> Deferring the tracked writes is sound because this gate is at most two steps later, nothing
> downstream consumes any of them before Step 2, and each is a Stage artifact write of exactly the
> kind this gate exists to place on the correct branch. An artifact written before this step is a
> **P-005** step-order violation as well as a P-010 one.
>
> 1. **Derive the scope slug.** `{scope-slug}` is the lower-kebab-case reduction of the selected
>    grouping's covering feature title (task-shaped intake) or of the feature, epic, or chore
>    title (feature-shaped intake), restricted to `[a-z0-9-]` and truncated at 48 characters.
>    Record it once and do not rename it later in the session. When an authoritative shipment ID
>    for this scope is **already** known, it may be used as the slug instead. Never wait for a
>    shipment ID: Step 5.5 creates the shipment after every mutation this gate exists to protect.
> 2. **Capture `stage_base_commit`.** In a **pipeline invocation** the Orchestrator captured and
>    retains it immediately before invoking Stage; Stage echoes the value it was given and does
>    not invent one. In a **direct invocation** Stage captures `git rev-parse HEAD` **before**
>    sub-step 3 creates or checks out the branch. Either way it is recorded once and never
>    re-captured later in the session.
> 3. **Resolve the branch — verify, or create.**
>    - *Pipeline invocation with a branch supplied by the Orchestrator*:
>      `git rev-parse --verify "$STAGE_BRANCH"`, then `git checkout "$STAGE_BRANCH"` when
>      `git rev-parse --abbrev-ref HEAD` does not already name it. Stage **verifies and uses** the
>      supplied branch and does not create a second one.
>    - *Direct invocation, or no branch supplied*: `git checkout -b chore/stage-{scope-slug}` from
>      the current commit, or `git checkout chore/stage-{scope-slug}` when it already exists.
> 4. **Assert, then proceed.** `git rev-parse --abbrev-ref HEAD` MUST equal the resolved Stage
>    artifact branch and MUST NOT be the default branch. On failure, halt with
>    `STAGE_BRANCH_GATE_FAIL: refusing to write staging artifacts while HEAD is {head}` and record
>    P-010. No Stage mutation proceeds past a failed assertion.
> 5. **Single worktree (P-016).** The branch is entered by switching the **one existing** worktree.
>    Stage does not run `git worktree add` and does not open a second checkout for it. The
>    time-boxed spike/research worktree exception is unrelated and unchanged.
> 6. **Idempotent on re-entry.** A resumed session already on the resolved branch verifies and
>    proceeds: no second branch, no re-derived slug, no re-captured base.

**Correction 2 — Step 5.7, the post-mutation artifact commit.** Add a new step section between the
existing Step 5.6 (Archive Consumed Stash Entries) and Step 6 (Summary):

> ### Step 5.7: Stage Artifact Commit (NON-NEGOTIABLE)
>
> Runs after **every** Stage mutation of the session — backlog items, the shipment manifest, the
> plan, deliberation, and memory records, and consumed-stash archival — and **before** Step 6
> emits the handback.
>
> 1. **Assert HEAD.** `git rev-parse --abbrev-ref HEAD` MUST equal the branch resolved at Step 1.9
>    and MUST NOT be the default branch. Otherwise halt with
>    `STAGE_COMMIT_GATE_FAIL: HEAD is {head}, expected the Stage artifact branch {stage_branch}`.
>    Do not commit the session's artifacts anyway.
> 2. **Detect out-of-root changes BEFORE staging anything (rev 10).** A root-scoped `git add`
>    cannot discharge this obligation: it *ignores* paths outside its pathspec rather than
>    reporting them, so a change outside the allowed roots would be silently left behind, not
>    detected. Run an explicit full-tree inspection first:
>
>    ```bash
>    git status --porcelain=v1 -z --untracked-files=all
>    ```
>
>    `--untracked-files=all` is **load-bearing**: the default `normal` mode collapses an untracked
>    subtree to a **directory** entry (`?? sub/`), which no per-file containment test can evaluate;
>    `all` recurses to `?? sub/deep.md`. `-z` is likewise load-bearing: without it, paths
>    containing spaces are quoted and paths containing unusual bytes are backslash-escaped, and
>    line-oriented splitting is unsafe on any of them.
>
>    **Parse the NUL-delimited records, never lines.** Split the output on NUL and drop the
>    trailing empty element. Each record is `XY` + one space + `PATH`, where `XY` is the two-column
>    status. When either status column is `R` (rename) or `C` (copy), the **next** record is the
>    ORIG_PATH and carries no `XY` prefix — consume it as part of the same entry. Both endpoints of
>    a rename or copy are affected paths and both are checked. Verified output shape:
>
>    ```text
>     D keep.txt\0R  new name.txt\0old name.txt\0?? sub/deep.md\0?? untracked.md\0
>    ```
>
>    This one command covers tracked modifications, deletions, renames, copies, and recursed
>    untracked files — the full set of ways a change can be present outside the roots.
>
>    **Halt fail-closed on any out-of-root path.** If **any** affected path — including a rename's
>    origin — does not lie under the fixed `STAGE_ARTIFACT_ROOTS` constant, halt with
>    `STAGE_COMMIT_GATE_FAIL: change outside Stage artifact roots: {path}`, record a **P-010**
>    signal, and surface the full offending list to the operator. Stage MUST NOT widen the commit
>    to include it, and MUST NOT `checkout`, `restore`, `stash`, revert, or delete it — the change
>    may be a human operator's unrelated in-flight work, and discarding it would be both a
>    destructive act without approval (Constitution VII) and a data-loss risk. **Pre-existing
>    unrelated dirt is therefore a halt, not a cleanup**: Stage stops and hands the decision to the
>    operator, leaving the working tree exactly as it found it. Gitignored paths never appear in
>    this output and are out of scope by construction.
> 3. **Stage only the allowed roots.** Only after the inspection above passes:
>    `git add -- .backlogit/ docs/plans/ docs/decisions/ docs/memory/` — the same fixed
>    `STAGE_ARTIFACT_ROOTS` constant the Orchestrator's Step 1.5 uses. The pathspec is a
>    **bound**, not a detector; detection already happened in sub-step 2.
> 4. **Commit, conventionally.** One commit per Stage session, in Conventional Commits form
>    (`chore(stage): …` or `docs(plan): …`), whose subject names the staged scope. When nothing is
>    staged the commit is a **no-op** and the step proceeds — a session that already committed
>    incrementally on this branch is compliant.
> 5. **Record the head.** `stage_head_commit := git rev-parse HEAD`, taken **after** sub-step 4, so
>    it names the commit that actually carries the artifacts.
> 6. **Derive the paths.** `stage_artifact_paths := git diff --name-only -z --diff-filter=d
>    {stage_base_commit} {stage_head_commit} -- .backlogit/ docs/plans/ docs/decisions/
>    docs/memory/` — the aggregate two-tree derivation the Orchestrator re-runs, over the base
>    **preserved** from Step 1.9.
> 7. **Stop at the commit.** Stage MUST NOT push, and MUST NOT create, update, or merge a pull
>    request. The Orchestrator (pipeline invocation) or the operator (direct invocation) owns every
>    remote operation on this branch.

**Correction 3 — register both steps and extend the Step 6 gate.** Two further edits, both
required, because a step section the Step Sequence Contract does not list is advisory in a file
whose own preamble says the checklist is what a session MUST execute:

- The **Step Sequence Contract** checklist gains `[ ] Step 1.9 — Stage artifact branch gate
  (pre-mutation)` between its Step 1.8 and Step 2 entries, and `[ ] Step 5.7 — Stage artifact
  commit (pre-handback)` between its Step 5.6 and Step 6 entries.
- Step 6's **Pre-Summary Verification Gate** gains two checks alongside the ones it already runs:
  Step 1.9 completed and `HEAD` is still the Stage artifact branch; Step 5.7 completed and
  `stage_head_commit` was recorded from its result. A summary presented without both is the same
  **P-005** class the gate already enforces for shipment assembly.

**This is an ordering contract, not a presence contract.** Both failure modes the finding describes
survive mere presence. A branch gate placed *after* harvest leaves every artifact written on the
default branch; a commit step placed *after* the summary emits a `stage_head_commit` that does not
yet exist. The ordinal positions above are therefore load-bearing and are asserted as such
(§6.3.2, AC-B3.7) — the checklist is an **ordered** list, so "between Step 1.8 and Step 2" and
"between Step 5.6 and Step 6" are mechanically checkable facts, not editorial preferences.

**Detector-neutrality (§5.3).** None of the prescribed text is a violation construct, and this was
checked rather than assumed. `checkout` is an excluded verb in construct 1, and `add`, `commit`,
`rev-parse`, `status`, `diff`, and — rev 10 — `restore` and `stash` are not push verbs at all; the
only default-branch tokens introduced appear in **prohibitive** clauses (`MUST NOT be the default
branch`, `MUST NOT write … while HEAD is the default branch`) where the marker **precedes** the
construct, which construct 2 accepts by the marker-precedes rule. Rev 10's out-of-root clause adds
`MUST NOT checkout, restore, stash, revert, or delete it`, which is likewise marker-preceded and
names no default-branch token at all. No absence assertion is introduced that the text could
self-match, because AC-B3.7 asserts **presence and order**, not absence. The closed construct set
and the fixture corpus are therefore **not** widened — the rev-6/7/8/9 non-expansion position,
unchanged.

**Why this belongs to B3 and not to a new task.** The two steps are edits to
`.github/agents/_stage.agent.md`, the single file `018.009-T` already owns, in the single domain it
already works in (contract prose), with replacement text fully specified above. Splitting them out
would add a ninth task to a shipment whose harness map is already one function per task (§5.0), and
would require a ninth harness function for assertions that belong to the surface
`TestDirectPushGate_StageAgentSurfaceClean` already scans. `018.009-T` is re-sized **M→L** on volume
alone and stays inside the 2-hour rule.

#### 6.3.2 Mechanical detection of the two operative steps (rev 9)

The assertions live in the **existing** `TestDirectPushGate_StageAgentSurfaceClean` function and use
the **existing** parse-the-installed-surface technique, adding no function, fixture, manifest entry,
or task. Three structural facts are asserted over `_stage.agent.md`:

1. **Checklist ordinals.** Parse the fenced Step Sequence Contract block, which is an ordered list
   of `[ ] Step N — …` entries. Assert a branch-gate entry exists whose index is **greater than**
   the learnings-retrieval entry's and **less than** the deliberation entry's, and a commit entry
   whose index is **greater than** the stash-archival entry's and **less than** the summary
   entry's. Index comparison — not string search — is what makes this an ordering assertion.
2. **Step sections exist and are correspondingly ordered.** Assert an ATX step heading for each of
   the two steps, and that their heading positions in the document obey the same two inequalities
   against the Step 1.8, Step 2, Step 5.6, and Step 6 headings. A checklist entry with no section,
   or a section the checklist omits, fails.
3. **Mandatory clauses per step.** Within the branch-gate section: the default-branch refusal, the
   **deferral rule** placing every earlier-step **tracked** mutation after the gate — including its
   three enumerated classes, the **reconciliation write half**, the **duplicate archival**, and
   **every pre-gate `docs/memory/` checkpoint** (naming at least the stash-classification and
   contextual-grouping/operator-selection checkpoints) — the verify-supplied-branch arm, the
   create-or-checkout arm, the `stage_base_commit` capture, and the single-worktree sentence.
   Within the commit section: the HEAD assertion, the **full-tree out-of-root detection** clause
   (the NUL-safe `--porcelain=v1 -z --untracked-files=all` inspection, the rename-endpoint rule,
   the out-of-root halt, and the do-not-revert sentence), the `STAGE_ARTIFACT_ROOTS`-bounded
   staging, the conventional-commit requirement, the `stage_head_commit` assignment taken after
   the commit, and the no-push / no-PR sentence.
4. **Intra-section ordering of the commit step (rev 10).** The detection clause's position in the
   commit section MUST be **before** the staging clause's. This is the same index-comparison
   technique as assertion 1, applied within one section rather than across the checklist, and it
   is required for the same reason: a detection clause placed after `git add` documents an
   inspection that runs too late to prevent the commit it exists to gate. Presence alone would
   satisfy a section that stages first and looks afterwards.

Assertion 1 is deliberately **not** a whole-file text search. The same self-matching hazard recorded
at §5.0 and AC-C1.6 applies here: `_stage.agent.md` will mention its own step names in the Step 6
gate and in the Role Boundary discussion, so a bare `grep` for a step name would pass with the step
section deleted. Parsing the checklist block and comparing indices is the structural analogue of
AC-C1.6's `jobs.lint.steps[]` walk, and it is chosen for the same reason.

**Regeneration exposure.** `_stage.agent.md` is autoharness-generated, so a render can revert both
steps. That is the R13 shape, and it is carried by the same mechanism: this function lives in the
non-generated `tests/integration/directpush_gate_test.go` and runs from job `expensive`
(`name: test`)'s **full-suite `./...`** test step, so a render that drops the steps turns CI **red**.
**Rev 13 — provenance precision**, matching §5.0 "Invocation provenance" exactly: the **observed
installed** command is `go test -race -mod=readonly ./...` (step `Test (race)`); the **template**
emits `{{TEST_COMMAND}}`; this workspace's **recorded render input** is `go test ./...`. What a
render re-emits is therefore **the step and its substitution — a full-suite `./...` invocation that
includes the integration harness — not those flags**, and AC-C1.7's post-render verification is what
keeps that package pattern true across a regeneration. The rev-9 wording of this paragraph called
the installed command the "generated-baseline" step; that was the same overstatement corrected
elsewhere at rev 12 and is withdrawn here. Recorded as **R17**.

**AC-B3.1** The Git row Allowed cell contains no default-branch allowance.
**AC-B3.2** The Git row Allowed cell grants creation and check-out of the dedicated artifact
branch, so the table and P-010 **agree** — no grant in one that the other omits or forbids. **Rev
9**: the branch is named `chore/stage-{scope-slug}`, with a shipment ID permitted but **not
required** as the slug, so the cell does not authorize a branch whose name is underivable at the
moment Step 1.9 needs it.
**AC-B3.3** The Git row Forbidden cell is unchanged; every other Role Boundary row is unchanged
**except** the PR row, which is explicitly exempted to carry the actor note. The note attributes
an actor only — the PR row's prohibition is unweakened and no authority transfers to Stage. This
criterion governs the **Role Boundary table**; the Step 6 edit required by AC-B3.5, the
Step 5.5 / Step 6 edits required by AC-B3.6, and the Step 1.9 / Step 5.7 / Step Sequence Contract /
Step 6-gate edits required by AC-B3.7 lie outside it.
**AC-B3.4** The file agrees with the corrected P-010 text — no surviving contradiction between
agent contract and policy, in either direction.
**AC-B3.5** **Staging handback emitted (rev 6).** Step 6's session-output contract requires
`stage_branch`, `stage_base_commit` (rev 8), `stage_head_commit`, `stage_outcome`,
`stage_artifact_paths`, and `shipment_id` (when formed) — every field the Orchestrator's Step 1
consumes under AC-B1.5, with none required there and absent here. **Rev 7/8**: the contract also
fixes the **shape** of `stage_artifact_paths` — a non-empty list of concrete repository-relative
**file** paths changed between `stage_base_commit` and `stage_head_commit`, never directory
prefixes and never only the tip commit's files — and names the derivation (`git diff --name-only
-z --diff-filter=d {stage_base_commit} {stage_head_commit} --` over the four Stage artifact roots),
so the emitted set satisfies the consumer's set-equality check (AC-B1.10) by construction.
**Rev 9**: the emitted values are **produced**, not asserted — `stage_branch` and
`stage_base_commit` come from the Step 1.9 branch gate and `stage_head_commit` and
`stage_artifact_paths` from the Step 5.7 artifact commit (AC-B3.7). **Rev 10**: the emitted
`stage_outcome` and `shipment_id` form a **consistent pair** — `shipment` is emitted with a
non-empty `shipment_id`, `no-shipment` is emitted with none, and the field is always emitted
explicitly rather than left for the consumer to infer, which is what makes the consumer's
fail-closed pair validation (AC-B1.12) a check on agreement rather than a substitute for a
missing producer.
**AC-B3.6** **The `no-shipment` terminal outcome is reachable (rev 8).** Step 5.5's mandatory
shipment assembly is scoped to a harvest that **produced items**; a reviewed, valid **empty**
harvest terminates with `stage_outcome: no-shipment` instead of halting, and Step 5.5's existing
empty-harvest / unresolved-P-003 guardrail is preserved **verbatim**. Step 6's pre-summary
verification gate requires `shipment_id` **only when `stage_outcome` is `shipment`**; for
`no-shipment` it requires the complete handback record (AC-B3.5) plus an explicit terminal
statement, and the summary does **not** route the operator to Ship. **P-003 failures, harvest
failures, and a missing required shipment after a non-empty harvest continue to HALT and are never
recorded as `no-shipment`** — the outcome is a success terminal, never a failure re-label.
**AC-B3.7** **Branch-before-mutation and commit-before-handback are operative and ordered
(rev 9).** `_stage.agent.md` carries **two operative steps**, each registered in the Step Sequence
Contract checklist and each backed by its own step section (§6.3.1):
(a) a **pre-mutation branch gate** whose checklist entry and section both sit **after** the
learnings-retrieval step and **before** the deliberation step — a **fixed** position, carrying a
**categorical** deferral rule that places **every** tracked Stage artifact mutation of the session
after the gate (rev 10), enumerating at least its three classes: the **write half** of the
mandatory late-identifier reconciliation, the **unconditional duplicate-entry archival**, and
**every tracked `docs/memory/` checkpoint whose triggering milestone precedes the gate**, naming
the stash-classification and contextual-grouping/operator-selection checkpoints explicitly. The
read-only classification and grouping analysis that derive the scope slug still run first, and
gitignored disposable operations (index sync, backlogit checkpoints, hook queue) are expressly
**outside** the rule rather than claimed by it — which derives a `{scope-slug}`
without requiring a `shipment_id`, records `stage_base_commit` (echoed from the Orchestrator in a
pipeline invocation, captured from `HEAD` **before** branch creation in a direct invocation),
**verifies and uses** an Orchestrator-supplied branch or **creates and checks out**
`chore/stage-{scope-slug}` when none was supplied, **refuses to write any artifact while `HEAD` is
the default branch**, and switches the single existing worktree rather than adding one (P-016);
and (b) a **post-mutation artifact commit** whose checklist entry and section both sit **after**
the consumed-stash-archival step and **before** the summary step, which asserts `HEAD` is the
Stage artifact branch and not the default branch, **detects out-of-root changes before staging
anything** (rev 10) via an explicit NUL-safe full-tree inspection
(`git status --porcelain=v1 -z --untracked-files=all`) whose records are parsed NUL-wise rather
than by line, whose rename/copy entries contribute **both** endpoints, and which halts fail-closed
with a P-010 signal on any path outside `STAGE_ARTIFACT_ROOTS` while leaving that pre-existing
change **untouched** (no checkout, restore, stash, revert, or delete), then stages only the four
`STAGE_ARTIFACT_ROOTS`,
commits in Conventional Commits form, sets `stage_head_commit` from the **resulting** `HEAD`, and
states that Stage neither pushes nor creates, updates, or merges a pull request. Step 6's
pre-summary verification gate requires both steps complete before any summary is presented.
**Ordering is asserted structurally**, by checklist-index and heading-position comparison rather
than by text search, and — rev 10 — the commit section's detection clause is asserted to precede
its staging clause by the same index comparison applied within the section (§6.3.2). AC-B3.5,
AC-B3.6 and AC-B3.7 are the **only** changes outside the Role Boundary table; no other section of
the file is modified.

These seven criteria are carried verbatim by task `018.009-T`. The post-correction **zero-findings**
gate run is deliberately **not** duplicated here: it is owned by **AC-C2.1** on `018.011-T`, which
scans the tree once after *all four* surfaces are corrected. Asserting it at B3 — when only three
of the four surfaces have landed — would be unsatisfiable.

---

## 7. Sub-epic C — CI coupling and verification

### 7.1 Task C1 — wiring, ledger, CODEOWNERS

**One step** in job `lint`, immediately after the write-path-precondition self-test step
(referenced by description, not line number):

```yaml
      - name: Run direct-push-language gate (0NN.00X-T)
        run: bash scripts/check-direct-push-language.sh --self-test
```

Unconditionally blocking — no `continue-on-error`, no toggle variable. The step name carries a
**task-ID suffix** so it satisfies the ledger's own membership rule; rev 1 omitted this.

**Ledger** (ci.yml header): add the new gate; **generalize** the membership sentence from the
literal `011.0xx-T` to "any `NNN.NNN-T` task-ID comment or task-ID-suffixed step name" (already
stale against the `015.0xx-T` group); **drop the "all five" magic count** in favour of the
structural membership rule; record that this gate deliberately has **no** toggle variable, so a
future reader does not add one for false symmetry; record that removal of this step is **caught,
not accepted** — `tests/integration/directpush_gate_test.go`
(`TestDirectPushGate_CIWiringIsBlocking`) is not autoharness-generated and runs from job
`expensive`'s **full-suite `./...`** test step, so
a render that drops this step turns CI
**red** rather than green and the re-apply obligation is discoverable at the point of divergence
(§5.0); **record the post-render verification obligation of AC-C1.7** (rev 12); and record the D2
residual — `.tmpl` sources
are gitignored, so template↔artifact equality is not CI-enforceable; the gate asserts the
*invariant* over installed artifacts instead.

**Provenance precision for the ledger text (rev 12).** The ledger MUST NOT state that
`go test -race -mod=readonly ./...` is the generated baseline. That command is the **observed
installed** one; the template emits `{{TEST_COMMAND}}` and this workspace's recorded render input is
`go test ./...` (§5.0 "Invocation provenance"). The ledger's claim is therefore scoped to what is
true: the harness runs because job `expensive`'s test step invokes the **full suite over `./...`**,
and a render re-emits *that step*, not those *flags*.

**CODEOWNERS**: add `/scripts/check-direct-push-language.sh @softwaresalt`. R6 rests on CODEOWNERS
review, and the file's own header promises it "Covers ALL CI-critical scripts … or none" — shipping
the gate unowned would widen an already-false claim. Also add the three already-missing gate
scripts (`check-write-path-precondition.sh`, `check-unignore-regression.sh`,
`check-depguard-fixtures.sh`) so the all-or-none contract becomes true.

No `uses:` is added, so `ci-security.instructions.md`'s 40-hex-SHA pinning MUST is satisfied
vacuously; job `lint` already declares `permissions: contents: read` and
`persist-credentials: false`, so the least-privilege MUSTs need no new work.

- **AC-C1.1** The step is present in job `lint`, **BLOCKING from its FIRST wiring** (no
  `continue-on-error`, no toggle variable, no advisory window), and task-ID-suffixed.
- **AC-C1.2** `.github/workflows/ci.yml` remains valid YAML
  (`python -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))"` passes) **and**
  `ci-gate` still transitively depends on job `lint`, so the new step is genuinely merge-blocking.
- **AC-C1.3** The LOCAL DIVERGENCE ledger names the gate, uses the generalized membership rule
  (wildcard/range form rather than the literal `011.0xx-T`), carries no magic count, records the
  no-toggle decision and the D2 residual, and records that removal of this step is **CAUGHT, NOT
  ACCEPTED** — `tests/integration/directpush_gate_test.go`
  (`TestDirectPushGate_CIWiringIsBlocking`) is not autoharness-generated and runs from job
  `expensive` (`name: test`)'s **full-suite `./...`** test step, so a render that drops this step
  turns CI **red** rather than green (§5.0). **Rev 8**: the CI step is cited by its exact name and
  command, because "the `go test` step in job `expensive`" names no step in this workflow.
  **Rev 12 — provenance correction**: the ledger MUST NOT assert that the observed installed command
  `go test -race -mod=readonly ./...` is the **generated baseline**. The template emits
  `{{TEST_COMMAND}}` and this workspace's recorded render input is `go test ./...` (§5.0 "Invocation
  provenance"), so the flags are local to the installed artifact. The installed command may still be
  cited **as the observed command** — that is what makes the evidence name a real step — but the
  ledger's survival claim is scoped to the **full-suite `./...` package pattern**, which is what a
  render actually re-emits.
- **AC-C1.4** CODEOWNERS coverage is asserted **structurally, not by count** (rev 3): every script
  invoked by a step in `.github/workflows/ci.yml` has an owner line in `.github/CODEOWNERS`; the
  header's fixture-data carve-out is generalized from `scripts/testdata/gitignore/**` to
  `scripts/testdata/**` so the new fixture tree is covered by the same stated rationale.
- **AC-C1.5** The script header states **VERBATIM** that this is an **ANTI-ACCIDENT control, not an
  anti-adversary control** (CI runs it from the PR's own head), and discloses that the job-`lint`
  wiring step sits **INSIDE** the autoharness render boundary (deliberation D6) **and** that its
  removal is **DETECTED — NOT ACCEPTED — by** `TestDirectPushGate_CIWiringIsBlocking` (§5.0), so the
  disclosure points a future reader at the surviving enforcement path rather than at a silent
  failure. Added at rev 4 — review round 3 — to close a plan↔task drift: `018.010-T` already carried
  this criterion while the plan did not state it. **Rev 5** replaced the "accepted residual" half of
  the disclosure per the R5 reclassification.
- **AC-C1.6** **Render-survival assertion — structural and non-vacuous (rev 5).**
  `TestDirectPushGate_CIWiringIsBlocking` determines step presence by **PARSING**
  `.github/workflows/ci.yml` and walking `jobs.lint.steps[]` for a step whose `run` invokes
  `scripts/check-direct-push-language.sh` — **NEVER** by a file-wide text search, which would
  self-match the LOCAL DIVERGENCE ledger comment naming the same script and pass vacuously after the
  step itself was deleted (§5.0; prior art
  `docs/compound/2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md`). Non-vacuity
  is proven by **EXECUTION**: the test writes a copy of the current `ci.yml` with that step removed
  into `t.TempDir()` and asserts the same check **FAILS** against the copy. **No edit is made to the
  real tree** — proven by **byte-identical** `git status --porcelain=v1 -z --untracked-files=all`
  snapshots taken immediately before and after (rev 15: snapshot **equality**, not an empty tree,
  because C2's own manifest and witness writes and any locally-excluded path make "empty"
  unsatisfiable exactly when the evidence is needed) — the same no-dirty-tree-edit rule as AC-A2.4
  and AC-C2.2. **Rev 11 — activation independence (the P0 correction).** This function's
  default-mode activation is derived **solely** from the harness-state manifest (`C1` present in
  `completed`), and **never** from the presence of the job-`lint` step it asserts on. Rev 10 keyed
  activation on that step, so a regeneration that dropped it made this function `t.Skip` instead of
  fail — the assertion disarmed itself at precisely the moment it was meant to fire. The criterion
  is therefore satisfied only if, with `C1` in `completed` and the job-`lint` step **absent**, the
  function **runs and fails**; a skip is an explicit failure of this criterion. C1's completion also
  inserts `"C1"` into `completed` at its `gateOrder` position (§5.0.1) — one line in
  `tests/integration/testdata/directpush-gate/harness-state.json`, the shared harness-state
  completion record, not a second skill domain; C1's size band is unchanged. **Rev 12**: "runs and
  fails" is a **default-mode** requirement; a targeted `-gatetask=C1` run must likewise execute the
  real assertion and never skip. C1 carries **no** MP5 obligation — MP5 is **RETIRED** (§5.0.1), and
  its retirement does not weaken this criterion, because C1's activation was already manifest-only.
  **Rev 13 — the evidence must be taken UNSELECTED.** A targeted `-gatetask=C1` run activates this
  function **by construction**, so it cannot demonstrate that manifest-derived activation survives
  removal of the lint step, and it would have passed identically under the rev-10 self-disarming
  design. The criterion's evidence is therefore the **default-mode** §5.0.2 point-7 observation
  against a **copied terminal-state fixture** under `t.TempDir()` whose manifest and witness remain
  valid and in agreement: it must show, **by test-event name**, `--- FAIL:
  TestDirectPushGate_CIWiringIsBlocking` citing the **missing lint step**, and
  `TestDirectPushGate_FullCorpusClean` **executing** rather than skipping. The targeted invocation is
  retained **separately** as §5.0.2 point 7a and is **selector evidence only**.
  **Rev 14 — the mechanism that makes that evidence obtainable is NAMED, not left to the
  implementer.** Rev 13 specified the observation but no way to run it: `repoRoot(t)` resolves the
  **real checkout** by walking up to `go.mod` from the process working directory, so an in-process
  run can never see the copy, and the observation lives inside C2, so an unguarded re-run recurses.
  The criterion is therefore satisfied only by the §5.0.2 point-7 **process contract**, implemented
  as unexported helpers in the existing `tests/integration/directpush_gate_test.go` (**no new
  harness function, no new test**): (i) `copyTrackedRepoFixture(t)` builds the fixture from
  `git ls-files -s -z` **tracked files only** into `t.TempDir()` — **rev 15**: `-s` so index
  **modes** are read, `100644`/`100755` copied (executable bit preserved) and **every** other mode,
  including `120000` symlinks and `160000` gitlinks, and every unmerged stage, **failing closed**;
  the generated `terminal-witness.json` copied **explicitly** from the working tree and validated
  against the copied manifest; the copy made a **self-contained Git repository** (`git init`,
  local non-secret identity, `git add`, deterministic local commit, **no** real `.git` copy, **no**
  submodule recursion, **no** network) with
  `git -C {copy} rev-parse --show-toplevel` asserted **equal to the copy root**; and the real tree
  left unwritten, proven by **byte-identical** `git status --porcelain=v1 -z --untracked-files=all`
  snapshots before and after; (ii) only the job-`lint` step is removed from the
  **copied** `ci.yml`, by the same `jobs.lint.steps[]` structural walk this criterion already
  mandates, **after** the copied manifest and witness are proven **valid and in agreement** (rev 15
  ordering, so the child's failure stays attributable to the removal); (iii)
  `runFixtureChildGoTest(t, fixtureRoot)` launches a **child `go test` process** whose
  `exec.Cmd.Dir` is the **copied module root** — which is precisely what makes `repoRoot(t)` resolve
  the copy — invoking
  `go test -json ./tests/integration -run '^TestDirectPushGate_(CIWiringIsBlocking|FullCorpusClean)$' -count=1`
  with **no `-gatetask`**, so activation is **default mode** while `-run` bounds only the child's
  **process scope**; (iv) the parent sets the non-recursive sentinel
  `DIRECTPUSH_GATE_FIXTURE_CHILD=1` on the child, which disables the **parent-only spawn block** and
  nothing else — it MUST NOT skip C2 and MUST NOT bypass MP1/MP6 (§5.0.1 "The fixture-child
  sentinel"); and (v) the parent decodes the child's `-json` **event stream** and requires a
  terminal `"Action":"fail"` event for `TestDirectPushGate_CIWiringIsBlocking` citing the missing
  step, a terminal `pass`-or-`fail` event for `TestDirectPushGate_FullCorpusClean` (a `skip` or an
  **absent** terminal event fails this criterion), a **non-zero** child exit, **no** recursion or
  timeout marker, and it **fails on a malformed event stream** rather than reading it as
  inconclusive. The child runs under an explicit `context.WithTimeout`, and its stdout/stderr are
  surfaced **only** in bounded failure output. Invocation is `exec.CommandContext` on the Go binary —
  **no shell**, so the mechanism is identical on every supported platform.
- **AC-C1.7** **Post-render verification of the rendered full-suite command (rev 12).** The
  regeneration-resistant guarantee (§5.0) depends on job `expensive`'s test step invoking the **full
  suite over `./...`**, because `./...` is what pulls `tests/integration` — and therefore this
  harness — into the run. That step is rendered from `{{TEST_COMMAND}}`, so the **package pattern is
  a render output, not a fixed constant**, and **no claim is made** that the observed `-race` /
  `-mod=readonly` flags survive a render. C1 therefore records, in the LOCAL DIVERGENCE ledger, a
  standing obligation: **after any `autoharness install` / `tune` regeneration of
  `.github/workflows/ci.yml`, verify — by reading the rendered step, not by assuming it — that the
  rendered expensive-job test command still invokes the full suite over `./...`**, and re-apply the
  C1 divergence step. The verification is satisfied by observing a rendered command whose package
  pattern is `./...`; it is **not** satisfied by the flags matching. If a render narrows the pattern,
  the harness stops running and **R5 reverts to an open residual** — which is why this obligation is
  written into the ledger rather than left to memory. Residual: **R23**. Size/complexity
  **UNCHANGED**: one additional sentence in a ledger this task already authors.

### 7.2 Task C2 — coupling verification

This numbered list is the **single verification contract** for C2. Task `018.011-T` carries it
verbatim; the two MUST NOT diverge (rev 4 reconciliation — the rev-3 task text asserted a
dirty-tree mutation proof that contradicted this section's fixture-backed one, and the two used
different numbering for the same criteria).

- **AC-C2.1** Full-corpus bare scan at HEAD exits **0** with **zero findings**; the resolved corpus
  file count and the findings count are recorded in the PR description as structured evidence.
- **AC-C2.2** **Mutation proof — fixture-backed, NO dirty-tree edit.** Reintroduction-detection for
  both violation shapes is proven by `--self-test` against the committed fixture corpus:
  `directpush-reject-attempt-first.md` locks the literal 3e **push** construct and
  `directpush-reject-commit-on-default.md` locks the **default-branch commit authorization**
  construct, and the run reports every fixture **by name** with actual == manifest verdict. **No
  temporary edit is made to the real tree at any point**; this is proven by **byte-identical**
  `git status --porcelain=v1 -z --untracked-files=all` snapshots taken immediately before and after
  the proof (**rev 15**: snapshot **equality**, not an empty tree — C2's own legitimate manifest and
  witness writes, and any path excluded only by a local non-portable `.git/info/exclude`, make an
  empty-tree precondition unsatisfiable precisely when this evidence is required; equality is also
  the **stronger** predicate, catching modification of already-dirty files, deletions, and new
  artifacts written beside existing dirt), and the evidence recorded is the `--self-test` output,
  not a narrated local edit. A reintroduction proof performed by mutating the committed tree is
  **explicitly rejected**: it is unrecorded, unreproducible in CI, and leaves a window in which the
  repository contains the very construct this gate exists to forbid. **Rev 11 extends this criterion
  to the harness-state manifest itself** (§5.0.1 MP4): against a **scratch copy** under
  `t.TempDir()` — same no-dirty-tree-edit rule, never the committed file — `gate()` is proven to
  **fail**, not skip, for each of **seven** mutations of
  `tests/integration/testdata/directpush-gate/harness-state.json`: (i) **deleted**; (ii) **malformed**
  (invalid JSON); (iii) **unknown `phase` value**; (iv) **rollback** (`completed` truncated, or
  `terminal`/`phase` left inconsistent with it); (v) **reordering** (`completed` permuted out of
  `gateOrder` sequence); (vi) **skipped prerequisite** (a member whose `gateDeps` prerequisite is
  absent); (vii) **duplicate or unknown ID**. **Rev 12 adds the three witness cases** to the same
  scratch-copy proof: (viii) **witness-only** — a valid witness present while the manifest is
  `red` or `build`; (ix) **terminal-manifest-only** — a terminal manifest with the witness
  **absent**; and (x) **malformed or mismatched witness** — invalid JSON, wrong `schema`,
  `version`, or `contract`, `completed` ≠ `gateOrder`, or a `digest` that does not match its own
  array. Each of the **ten** cases records an observed failure. The state files are
  part of the mutation proof because they are now part of the control: an unguarded state file would be
  a silent off-switch for every assertion in the harness.
- **AC-C2.3** **Self-match negative proof, executed.** In the same bare scan, the three real
  marker-free prohibition constructs in the corrected tree — the post-change P-010 **Stage**
  prohibition bullet, P-010's **Ship MUST NOT** bullet "Commit or push directly to `main`" in
  `workflow-policies.md`, and `_ship.agent.md`'s Role Boundary `| Git |` Forbidden cell — all score
  **accept**. Each is located by **file plus quoted construct text**, never by line ordinal
  (rev 8), so the criterion survives any edit that shifts line numbers in those files. Rev 1's
  AC3.3 (ci.yml step names) was vacuous because `.github/workflows/**` is never in the corpus.
  Prior art:
  `docs/compound/2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md`.
- **AC-C2.4** **Authorization-surface agreement, as the surfaces exist at the end state (corrected
  rev 12).** The installed `.github/**` artifacts and the corrected policy agree — no surviving
  contradiction across the **live post-change** authorization surfaces. **Rev 12 correction**: this
  criterion previously read "the four **§2** authorization surfaces". §2 is the **historical
  root-cause enumeration**, and one of its four entries is `_orchestrator.agent.md` Step 1.5
  **sub-step 3e**, which **B1 deletes** — so the criterion asserted agreement across a surface that
  does not exist at the end state, the same defect AC-C2.4's own rev-11 DoD fix corrected one file
  over. The live set it now compares is: (1) `_stage.agent.md` **Role Boundary Git row** (the
  Stage surface); (2) `workflow-policies.md` **P-010 Stage** bullets (the P-010 surface);
  (3) `_orchestrator.agent.md` Step 1.5 **preamble**, in its reworded verification/postcondition
  form; and (4) `_orchestrator.agent.md` Step 1.5 **sub-step 3a**, which carries the branch-point
  instruction folded in from the deleted 3e. Plus an explicit **3e-absence** clause: sub-step 3e
  MUST NOT be present at the end state (AC-B1.1), so it is not a surface that can agree or
  disagree. §2's narrative quotation of the 3e construct is a **historical record** outside the
  scanned corpus (§5.2) and is deliberately retained. **Rev 6**: the **staging handback contract** also agrees across its two surfaces — every field
  `_orchestrator.agent.md` Step 1 requires (AC-B1.5) is emitted by `_stage.agent.md` Step 6
  (AC-B3.5), with no field required in one and absent from the other. **Rev 7**: that agreement
  extends to the **shape** of `stage_artifact_paths` — the producer (`_stage.agent.md` Step 6,
  AC-B3.5) and the consumer (`_orchestrator.agent.md` Step 1.5 step 4, AC-B1.10) declare the same
  concrete-file contract and the same derivation command, and **neither** surface retains a
  directory-prefix default. A field agreeing in name but disagreeing in shape is the same
  defect class one level down, and passes a name-only comparison. **Rev 8**, two further
  agreements: (a) both surfaces carry `stage_base_commit` and the **same aggregate** two-tree
  derivation, so neither retains a single-commit form; and (b) the `no-shipment` **outcome value
  itself** is agreed — the consumer's arm is keyed on an outcome the producer can actually emit
  (AC-B3.6). A verification arm keyed on an outcome no producer can reach is not a passing check
  but an **unreachable** one, and it passes every agreement test that compares only field names.
  **Rev 9**, the producing-action agreement: every handback field the consumer reads has a
  **producing step** in the producer surface, and those steps are **ordered** correctly —
  `_stage.agent.md` carries an operative branch gate before its first artifact write and an
  operative artifact commit before its summary (AC-B3.7), so `stage_branch`, `stage_base_commit`,
  `stage_head_commit`, and `stage_artifact_paths` name a branch that was created and commits that
  were made. Both surfaces also agree on the **branch-naming form**: neither requires a
  `shipment_id` to name the branch (AC-B2.2, AC-B3.2). A field that is declared, shaped, and
  emitted but never *produced* is the same defect class a third level down, and it passes both a
  name-only and a shape-only comparison. **Rev 10**, two final agreements: (a) the
  outcome/`shipment_id` **pair** is agreed — the producer emits `stage_outcome` explicitly and
  consistently with `shipment_id` (AC-B3.5), and the consumer **preserves** a reported outcome,
  derives only when absent, and **halts fail-closed** on an unknown value or an inconsistent pair
  before selecting an arm (AC-B1.12); neither surface retains the unconditional
  derive-from-`shipment_id` rule that silently reclassified a reported `shipment` run as a
  successful `no-shipment` one. (b) The producer's **pre-gate deferral** and **out-of-root
  detection** clauses are present and correctly ordered (AC-B3.7) — the deferral rule is
  categorical over every tracked mutation and names its three classes while expressly excluding
  gitignored disposables, and the commit step's full-tree inspection is positioned **before** its
  staging clause, so the handback the consumer reads describes a commit that could not have
  silently omitted or silently widened its contents. A field whose value is agreed but whose
  *production order* admits a write to the wrong branch is the same defect class a fourth level
  down, and it passes name, shape, and producing-step comparisons alike.
- **AC-C2.5** `markdownlint` exits **0** on every changed markdown file (P-008).
- **AC-C2.6** Branch and PR merge-commit rules intact — P-009, P-011, P-016 textually unchanged.
- **AC-C2.7** Constitution Quality Gates run **in declared order, none skipped, as separate
  commands** (never chained with `&&`, which would short-circuit and mask which gate failed):
  `gofmt -l .` empty, then `go vet ./...`, then `go test ./...`, then `go build ./...` — each
  exits 0 and each result is recorded independently. **Rev 4**: `go test ./...` is no longer a pure
  regression check — it now includes the `tests/integration/directpush_gate_test.go` harness (§5.0),
  which must be **green** in full, so all eight per-task functions pass here. **Rev 11**: at C2's
  boundary the manifest is in `terminal` phase, so "green in full" means **all eight functions
  activate and pass their real assertions** — `go test -v ./tests/integration -run TestDirectPushGate`
  shows **zero `--- SKIP:`** lines for the eight gate functions. A skip at this point is a failure of
  this criterion, not a pass. **Rev 12 qualification**: that zero-skip claim is scoped to a
  **default-mode** run (no `-gatetask`). A **targeted** run legitimately skips the seven
  non-selected functions in any phase — those are selector skips, not activation skips — and the
  only targeted-mode requirement is that the **selected** test actually executes. Evidence for this
  criterion MUST therefore be taken from a default-mode invocation. **Rev 13**: that is now
  structural rather than conventional — the selector holds **activation precedence** in `gate()`'s
  Stage 2 (§5.0.1), so a valid terminal witness does **not** activate the seven unselected bodies in
  a targeted run. Witness and manifest **integrity** still fail a targeted invocation, because
  Stage 1 is mode-independent.
- **AC-C2.8** **Regeneration-resistant enforcement, recorded (rev 5; provenance corrected rev 12).**
  The evidence records that the full-corpus scan of AC-C2.1 is executed by
  `TestDirectPushGate_FullCorpusClean` from job `expensive` (`name: test`)'s test step — whose
  **observed installed** command is `go test -race -mod=readonly ./...` (rev 8 — cited precisely,
  because "the `go test` step" does not identify a step in this workflow) — and **not solely** by
  the job-`lint` step added in C1. **Rev 12 corrects what may be claimed about that command's
  provenance**: it is the **observed installed** command, **not** the generated baseline. The
  template emits `{{TEST_COMMAND}}` and this workspace's recorded render input is `go test ./...`
  (§5.0 "Invocation provenance"), so the property the evidence may assert is that a **full-suite
  `./...` invocation** exists in a job the render re-emits and the `code` filter cannot skip —
  which is exactly what makes `tests/integration` run. The evidence MUST NOT assert that the
  `-race` or `-mod=readonly` flags are render-guaranteed, and MUST cite AC-C1.7's post-render
  verification as the mechanism that keeps the `./...` property true across a regeneration.
  This is **separate from**, and does not replace, AC-C2.7's constitutional `go test ./...`
  requirement: that gate is run locally in declared order, this one names the CI step that survives
  a render. Both halves of the §5.0 invariant are asserted green in the same run: **(a)** zero
  findings over the installed surfaces, and **(b)** the job-`lint` step still present, structurally
  and non-vacuously per AC-C1.6. A future render that drops the C1 step therefore fails (b)
  **loudly** while (a) continues to execute — which is what makes §1's "cannot silently reopen the
  gap" true as written rather than aspirational. **Rev 11 supplies the property this criterion
  silently assumed**: both halves only execute if both functions are **active**, and at rev 10 both
  were activated by the very step whose removal they were meant to detect, so a render dropped the
  step *and* the two detectors together. Activation now comes from the harness-state manifest alone
  (AC-C2.9), so the evidence must record that C1 and C2 are in `completed` and that neither
  function's activation reads any surface it asserts on.
- **AC-C2.9** **Terminal completeness and activation independence, from manifest state (rev 11).**
  C2's completion sets `tests/integration/testdata/directpush-gate/harness-state.json` to the
  **exact eight-element ordered set** `["A1","A2","A3","B1","B2","B3","C1","C2"]` with
  `terminal: true` and `phase: terminal`, **atomically** — the three fields move together, and a
  partial write is an inconsistent triple that MP1 rejects. Inside
  `TestDirectPushGate_FullCorpusClean`, that completeness check is evaluated **from manifest state
  before any surface-dependent assertion in the function**, so no surface failure can mask the
  completeness verdict; the real full-corpus assertions then run. In `terminal` phase the default
  suite activates **all eight** substantive assertions and **no activation skip remains anywhere in
  the harness** — evidenced by zero `--- SKIP:` lines for the eight gate functions **in a
  default-mode run** (rev 12: a targeted run skips the seven non-selected functions by design and is
  not evidence for or against this clause). Neither C1 nor
  C2 derives activation from the job-`lint` step or from any other surface either function asserts
  on: once their IDs are in `completed` they are active **permanently**, so a regeneration that
  removes the lint step leaves both **active and failing loudly** rather than skipping. Targeted
  `-gatetask=C1` and `-gatetask=C2` execute their real assertions **regardless** of current surface
  presence, phase, or manifest contents (manifest **validity** is still enforced — MP1 fails closed
  in every mode). **Rev 13**: the converse is equally binding — a targeted run **skips the seven
  non-selected functions**, in every phase **including `terminal`** and **regardless of witness
  presence**, because the selector has **activation precedence** over the default-mode rules
  including MP6's forcing clause (§5.0.1 Stage 2). Evidence that this clause holds MUST therefore be
  taken **unselected**; the §5.0.2 point-7 lint-removal observation is run in **default mode** for
  exactly that reason, with the targeted form retained separately as point 7a selector evidence.
  C2's completion also performs the one-line insertion into `completed` described in
  §5.0.1 — the shared harness-state completion record, not a second skill domain; C2's size band is
  unchanged.
- **AC-C2.10** **Terminal completion witness — independent, minimal, fail-closed (rev 12).** C2's
  completion creates `tests/integration/testdata/directpush-gate/terminal-witness.json`, a tracked,
  non-generated file **independent of the harness-state manifest**, with **exactly** these fields
  and no others: `schema` = `"directpush-gate/terminal-witness"`, `version` = `1`, `contract` =
  `"017-S/018-F"`, `completed` = the exact eight-element ordered set, and `digest` = the lowercase
  hex SHA-256 of those eight IDs joined with `\n` (no trailing newline), computed with stdlib
  `crypto/sha256`. It carries **no timestamps and no environment-specific data** — no host, user,
  path, branch, commit, or run identifier — so it is byte-reproducible from the contract alone.
  **Write order is normative**: the **witness is written first**, then the manifest is **atomically
  replaced** with the exact terminal state; the tests are green only when the two **agree**.
  `gate()` enforces MP6 **before any activation decision**: a present witness must be valid and
  requires a terminal manifest; a terminal manifest requires a present valid witness; a **missing,
  malformed, or mismatched** witness **wherever terminal or witness evidence exists** is
  **fail-closed**; and **absence at H0 and throughout `build` is valid**, not an error. In
  **default mode** a present valid witness additionally **forces `C2` active**. **Rev 13 separates
  the two halves**: the **integrity** requirements just listed are **mode-independent** (`gate()`
  Stage 1) and fail a targeted invocation exactly as they fail a default one, while the **forcing**
  clause is an **activation** rule and is **default-mode only**, so a targeted run selecting another
  task still skips `C2` (MP3). Scoping it that way subtracts no detection: the rewind MP6 exists to
  catch is caught by the integrity half, which runs in both modes. Consequently a rollback of
  **only** the manifest and a rollback of **only** the witness each **fail loudly** (proven by
  execution — AC-C2.2 cases viii–x). **Recorded residual, not
  claimed as closed**: rolling back or deleting **both** files in one change yields a state that a
  stateless validator cannot distinguish from legitimate pre-C2 `build`, because that is what it
  looks like; it is caught by **diff and review** of two tracked non-generated files, **not** by
  runtime history (**R22**). No claim is made here that any stateless validator detects every
  coordinated rollback.
- **AC-C2.11** **Lifecycle evidence ledger (rev 12; point 7 rewritten rev 13; point 7's mechanism,
  point 1's input set, and point 4's anchor specified rev 14).** The PR evidence
  records all **eight** §5.0.2 observation points, each as a run command plus its observed output,
  never as narration: (1) **H0** — 8 expected failures with their own markers, **zero default-mode
  activation skips**, **witness absent**, **and set equality** between the harnessed task set and
  the eight `017-S` task IDs, none omitted for dependency state (rev 13; §5.0 "H0 eligibility");
  **rev 14** — that set equality is recorded on **both sides of the call**: the **exact eight task
  IDs passed to `harness-architect` as `${input:tasks}`** (an explicit input, because that parameter
  is Optional and its omission falls back to "all ready tasks under the feature"), checked equal to
  the shipment manifest minus the covering feature **before** invocation, and the selected set
  `harness-architect` reports **after** it; the record also states the accurate contract reading —
  Ship Step 2 may **verify dependency-graph validity**, while dependency **readiness** gates
  execution order at Step 3 and claim at Step 4, and `blocked` is a **lifecycle status value** none
  of the eight `queued` tasks carries;
  (2) **post-A2** — A2 targeted green, A3 targeted expected-red, **default suite
  green**; (3) **legal B1/B2/B3 orders** — each dependency-closed order-preserving subsequence is
  **accepted**, no legal order rejected; (4) **post-C1** — **default suite green**, C2 targeted
  expected-red, **witness absent** (this is the state rev-11's MP5 deadlocked); **rev 14 — that
  expected-red is ANCHORED, not merely a non-zero exit**: the recorded output must carry
  `--- FAIL: TestDirectPushGate_FullCorpusClean` **by that exact name**, citing the **MP2
  non-terminal-state** reason (seven-element `completed`, `phase: build`, `terminal: false`, witness
  absent). Five impostor exits are **explicitly rejected** as evidence for this point — a
  **compile/build error**, a **malformed-state** (MP1/MP6 Stage 1 integrity) failure, a **missing
  `bash`** failure, an **invalid-selector** failure, and any **unrelated non-zero** exit — because
  each of them exits non-zero without demonstrating the property the point exists to show;
  (5) **terminal** — exact eight-task set **plus** valid witness, all 8 substantive assertions
  green, **zero default-mode skips**; (6) **terminal tree with the manifest rolled back to seven
  tasks while the witness remains** — **failure** (MP6 integrity, which is mode-independent);
  (7) **terminal-state fixture with the job-`lint` step removed, run in DEFAULT MODE** — against a
  **copied** terminal-state repository fixture under `t.TempDir()` whose manifest and witness remain
  **valid and in agreement**, with **no mutation of the real tree**, the evidence shows **by
  test-event name** that `TestDirectPushGate_CIWiringIsBlocking` **fails for the missing lint step**
  and that `TestDirectPushGate_FullCorpusClean` **executes** (a terminal `pass`/`fail` event, never
  `skip` and never absent). **Rev 14 — the executable mechanism is part of this criterion**, because
  rev 13's bare command could not produce the observation at all: `repoRoot(t)` resolves the **real
  checkout**, and the observation lives inside C2 and would recurse. The evidence is produced by the
  §5.0.2 point-7 process contract, implemented as unexported helpers in the existing
  `tests/integration/directpush_gate_test.go` — `copyTrackedRepoFixture(t)` (tracked-file copy via
  `git ls-files -s -z` into `t.TempDir()`; **rev 15** — index **modes** read, `100644`/`100755`
  copied with the executable bit preserved and **all** other modes, `120000` and `160000` included,
  plus any unmerged stage, **failing closed**; the generated `terminal-witness.json` copied
  **explicitly** from the working tree, **required to exist**, parsed, and **validated against the
  copied manifest** before the lint step is removed; the copy made a **self-contained Git
  repository** via `git init` + local non-secret identity + `git add` + a deterministic local
  commit, with **no** real `.git` copy, **no** submodule recursion and **no** network, and
  `git -C {copy} rev-parse --show-toplevel` asserted **equal to the canonical copy root** so no
  nested or sibling repository can be selected; real tree never written, proven by
  **byte-identical** `git status --porcelain=v1 -z --untracked-files=all` snapshots before and
  after rather than by an empty-tree assertion)
  and `runFixtureChildGoTest(t, fixtureRoot)` (a **child `go test` process** whose
  `exec.Cmd.Dir` is the **copied module root**, running
  `go test -json ./tests/integration -run '^TestDirectPushGate_(CIWiringIsBlocking|FullCorpusClean)$' -count=1`
  with **no `-gatetask`** so activation is default mode and `-run` bounds only process scope) —
  guarded by the non-recursive sentinel `DIRECTPUSH_GATE_FIXTURE_CHILD=1`, which disables the
  **parent-only spawn block** and **must not** skip C2 or bypass MP1/MP6. The recorded evidence is
  **parsed `go test -json` events**, never scraped text, and must show the C1 `fail` event with the
  missing-step reason, the C2 terminal `pass`/`fail` event, a **non-zero child exit**, **no
  context-deadline / killed-child termination**, and a **well-formed** stream — a malformed or
  truncated stream
  **fails** rather than passing as inconclusive. **Rev 15 — recursion is PREVENTED, not detected**:
  the spawn block is entered only when the sentinel is **unset**, which bounds depth at exactly one
  level by construction, so the rev-14 "recursion marker" (a child observing the sentinel already
  set **when the block is reached**) is **unsatisfiable given that guard** and is **withdrawn** as
  evidence; the genuine bound is the timeout condition above. The child is bounded by an explicit
  `context.WithTimeout`, invoked through `exec.CommandContext` on the Go binary with **no shell**,
  and its stdout/stderr appear **only** in bounded failure output;
  (7a) the **targeted** `-gatetask=C1` form of the same fixture is recorded **separately and
  labelled selector evidence only** — it activates C1 by construction and is explicitly **not**
  evidence of terminal default-mode activation; (8) **witness-only, terminal-manifest-only, and
  malformed witness** — **failure** in each. Points 1, 2 and 4 are captured at the boundaries where
  they occur; points 3, 5, 6, 7, 7a and 8 are executed or replayed at C2 under `t.TempDir()` per
  AC-C2.2's no-dirty-tree-edit rule. Every "zero skips" figure is taken from a **default-mode** run,
  and **no targeted run may be offered as evidence of a default-mode activation claim** (rev 13).

## 8. Verification commands

```bash
bash scripts/check-direct-push-language.sh --self-test   # fixtures + selection + real-tree
bash scripts/check-direct-push-language.sh               # verdict; expect 0 findings
go test ./tests/integration -run TestDirectPushGate      # the P-002/P-004 harness (§5.0)
# R5 invariant (§5.0): (b) invocation survives the render boundary + (a) detector still runs
go test ./tests/integration -run 'TestDirectPushGate_(CIWiringIsBlocking|FullCorpusClean)'
markdownlint "**/*.md"
python -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml'))"
actionlint .github/workflows/ci.yml   # LOCAL, operator-run — no CI counterpart exists in this repo
git status --porcelain=v1 -z --untracked-files=all   # rev 15 — snapshot taken BEFORE and AFTER;
# the two must be BYTE-IDENTICAL around AC-C2.2 and around the point-7 fixture-child run. NOT an
# empty-tree check: C2's own witness/manifest writes, and paths excluded only by the local
# non-portable .git/info/exclude, make "empty" unsatisfiable exactly when the evidence is needed.
# Rev 7/8 — the no-shipment arm's primitives, verified against this repository:
git show origin/main:docs/plans/           # exits 0 on a TREE — why `show` is insufficient (AC-B1.10)
git cat-file -t origin/main:docs/plans/    # prints `tree`  → rejected
# Rev 8 aggregate derivation (two-tree, stable across the merge). NOT the range form
# `origin/main..{head}`, which is empty by construction once the head is an ancestor of origin/main:
git diff --name-only -z --diff-filter=d {stage_base_commit} {stage_head_commit} -- \
  .backlogit/ docs/plans/ docs/decisions/ docs/memory/
git merge-base --is-ancestor {stage_base_commit} {stage_head_commit}   # base→head ancestry (check 3)
# Rev 9 — the producing side, exercised by the B3 harness function rather than by a shell command:
go test ./tests/integration -run TestDirectPushGate_StageAgentSurfaceClean   # branch gate + commit step, present AND ordered
# Rev 10/11 — per-task activation (§5.0.1). Activation comes from the harness-state manifest,
# NEVER from a substantive surface the harness asserts on (the rev-10 self-disarming defect).
# Default mode: red = all eight active; build = the completed set; terminal = all eight active.
# Targeted mode: force exactly one task's assertions in any phase — how each task's red is observed.
# Rev 13 — the selector has ACTIVATION PRECEDENCE: a targeted run skips the other seven in EVERY
# phase, including terminal and regardless of witness presence. Witness/manifest INTEGRITY is
# mode-independent and still fails a targeted invocation (gate() Stage 1).
go test ./tests/integration -run '^TestDirectPushGate_StageAgentSurfaceClean$' -gatetask=B3
# Rev 11 — inspect the harness-state manifest (JSON: stdlib-parseable, outside every render boundary):
cat tests/integration/testdata/directpush-gate/harness-state.json
# Rev 12 — the terminal completion witness. ABSENT at H0 and throughout `build`; present ONLY at terminal.
# Written BEFORE the manifest's atomic terminal replacement; MP6 requires the two to agree.
test ! -e tests/integration/testdata/directpush-gate/terminal-witness.json   # H0/build: MUST NOT exist
cat tests/integration/testdata/directpush-gate/terminal-witness.json         # terminal: MUST exist and be valid
# Rev 12 — the witness digest, recomputed independently of the harness (must equal the `digest` field):
printf 'A1\nA2\nA3\nB1\nB2\nB3\nC1\nC2' | sha256sum
# Rev 11 — H0 red-phase evidence must be LITERAL 8/8, each with its own task-specific marker:
go test ./tests/integration -run TestDirectPushGate 2>&1 | Select-String -Pattern '--- FAIL:' # expect 8
go test ./tests/integration -run TestDirectPushGate 2>&1 | Select-String -Pattern 'not implemented: '
# Rev 11 — terminal-state proof: all eight active, ZERO skips among the gate functions (AC-C2.9).
# Rev 12 — this is a DEFAULT-MODE claim. A targeted run skips the seven non-selected functions BY DESIGN;
# the only targeted-mode requirement is that the SELECTED test actually executes (never skips).
go test -v ./tests/integration -run TestDirectPushGate 2>&1 | Select-String -Pattern '--- SKIP:' # expect none
# Rev 11, REWRITTEN rev 13, MECHANISM SPECIFIED rev 14 — C1/C2 stay active from manifest state with
# the job-`lint` step removed (AC-C1.6, AC-C2.8, §5.0.2 point 7). The AUTHORITATIVE evidence is
# DEFAULT MODE (no -gatetask) against a COPIED terminal-state fixture — real tree never mutated, and
# the fixture's manifest + witness must stay VALID AND IN AGREEMENT or Stage 1 fails the run instead.
# This observation is NOT a hand-typed command: the rev-13 form named bare `go test` invocations that
# CANNOT produce it, because repoRoot(t) walks up to go.mod from the process working directory and so
# resolves the REAL CHECKOUT, and because the observation lives INSIDE C2 and would recurse.
# It is produced by the PARENT TestDirectPushGate_FullCorpusClean, which:
#   1. copyTrackedRepoFixture(t)   — `git ls-files -s -z` tracked-file copy into t.TempDir().
#      REV 15: -s reads INDEX MODES. 100644/100755 copied (exec bit preserved); 120000 symlinks,
#      160000 gitlinks, any other mode and any unmerged stage FAIL CLOSED. Verified on this repo:
#      all 629 entries are 100644 — zero 100755/120000/160000, no .gitmodules — so NO skip-list is
#      written: `references/` is NOT in the index at all (it holds the locally-excluded nested clone
#      references/herdr), and the fail-closed rows are a forward guard, not dead text.
#   1b. REV 15 — copies the GENERATED terminal-witness.json explicitly from the WORKING TREE (it may
#      not be indexed yet, since C2's own completion writes it), REQUIRES it to exist, parses it and
#      validates it against the copied manifest, then makes the copy a SELF-CONTAINED GIT REPO:
#      git init + local non-secret identity + git add + deterministic local commit; NO real .git
#      copy, NO submodule recursion, NO network. Asserts, before the child runs:
#         git -C {copy} rev-parse --show-toplevel  ==  canonical copy root
#      so no nested/sibling repository can be selected.
#   2. removes ONLY the job-`lint` step from the COPIED ci.yml (structural jobs.lint.steps[] walk),
#      AFTER manifest+witness are proven valid and in agreement — so child failure stays
#      ATTRIBUTABLE to the removal rather than to a malformed or incomplete copy.
#   3. runFixtureChildGoTest(t, fixtureRoot) — CHILD process, exec.Cmd.Dir = COPIED MODULE ROOT
#      (which is what makes repoRoot(t) resolve the copy), env + DIRECTPUSH_GATE_FIXTURE_CHILD=1,
#      exec.CommandContext with an explicit timeout, NO SHELL. The child command is exactly:
#         go test -json ./tests/integration \
#           -run '^TestDirectPushGate_(CIWiringIsBlocking|FullCorpusClean)$' -count=1
#      NO -gatetask  => DEFAULT ACTIVATION MODE. -run bounds PROCESS SCOPE only, never activation.
#   4. parses the -json EVENT STREAM (never scraped text) and requires ALL of:
#         - terminal "Action":"fail" for TestDirectPushGate_CIWiringIsBlocking, citing the MISSING STEP
#         - terminal "Action":"pass" OR "fail" for TestDirectPushGate_FullCorpusClean
#           (an "Action":"skip", or NO terminal event for that name, FAILS the observation)
#         - NON-ZERO child exit (C1 detects the removed step); a zero exit FAILS
#         - NO context-deadline / killed-child termination. REV 15: the rev-14 "recursion marker
#           (sentinel already set)" is WITHDRAWN — the guard means a sentinel-set process never
#           REACHES the block, so that condition is unsatisfiable and proved nothing. Recursion is
#           PREVENTED parent-side (block entered only when sentinel is UNSET => depth bounded at 1)
#           and BOUNDED by the child timeout.
#         - a WELL-FORMED stream; malformed/truncated FAILS rather than reading as inconclusive
#   5. REV 15 — real-tree no-mutation evidence is SNAPSHOT EQUALITY, not an empty tree:
#         git status --porcelain=v1 -z --untracked-files=all   (NUL-parsed, before AND after)
#      the two snapshots MUST be BYTE-IDENTICAL. Legitimate C2 manifest/witness dirt is tolerated
#      only by being UNCHANGED; the parent fails iff the mechanism changed the real checkout.
# The sentinel disables ONLY the parent-only spawn block: it MUST NOT skip C2 and MUST NOT bypass
# MP1/MP6 (§5.0.1 "The fixture-child sentinel"). Child stdout/stderr appear ONLY in bounded failure
# output. The whole observation is therefore driven by the default suite:
go test -v ./tests/integration -run '^TestDirectPushGate_FullCorpusClean$'   # parent drives the child
# Rev 13 — SELECTOR EVIDENCE ONLY (§5.0.2 point 7a), KEPT SEPARATE from the default-mode evidence
# above. This activates C1 BY CONSTRUCTION and is therefore NOT evidence of terminal default-mode
# activation; it proves only that the selector works.
go test ./tests/integration -run '^TestDirectPushGate_CIWiringIsBlocking$' -gatetask=C1  # expect non-zero, NOT a skip
# Rev 14 — §5.0.2 point 4's post-C1 targeted C2 red is ANCHORED, not merely non-zero. It MUST print
# `--- FAIL: TestDirectPushGate_FullCorpusClean` and cite the MP2 NON-TERMINAL-STATE reason.
# REJECTED as evidence for this point: a compile/build error, an MP1/MP6 malformed-state failure,
# a missing-`bash` failure, an invalid-selector failure, or any unrelated non-zero exit.
go test -v ./tests/integration -run '^TestDirectPushGate_FullCorpusClean$' -gatetask=C2  # expect the ANCHORED fail
# Rev 10 — AC-A2.5's two exactly-anchored invocations (the second MUST print `--- FAIL:`;
# a non-matching -run pattern exits 0 with `[no tests to run]`, so exit code alone proves nothing):
go test ./tests/integration -run '^TestDirectPushGate_SelfTestDriverEnumeratesFixtures$' -gatetask=A2
go test ./tests/integration -run '^TestDirectPushGate_SelfTestPasses$' -gatetask=A3   # expect non-zero
# Rev 10 — Step 5.7's out-of-root detection primitive. `--untracked-files=all` is load-bearing:
# the default `normal` mode emits a DIRECTORY entry (`?? sub/`) that no per-file check can evaluate.
git status --porcelain=v1 -z --untracked-files=all
gofmt -l .          # must print nothing
go vet ./...
go test ./...
go build ./...
# Constitution gate order: run as SEPARATE commands, in this order, none skipped.
# Do NOT chain with && — a chain masks which gate failed and short-circuits the rest,
# which is exactly what "in declared order, none skipped" forbids (AC-C2.7).
```

`actionlint` is **not** installed or invoked by any workflow here; AC-C1.2 (`yaml.safe_load`) is
the CI-enforced half. Rev 1 implied CI enforcement that does not exist. If a verification tool is
absent locally, record `TOOL_DEGRADED` rather than silently skipping (P-012).

## 9. Plan hardening signals

Edits a NON-NEGOTIABLE merge-gate contract (Step 1.5), the authoritative policy registry (P-010),
and the agent's operative Role Boundary table; adds a **blocking** CI gate; carries a documented
in-repo self-matching failure mode; touches autoharness-generated files with untracked templates.
**Rev 9** adds two **NON-NEGOTIABLE operative steps** to the Stage agent's own Step Sequence
Contract, which changes what every future Stage session must execute — a wider behavioural blast
radius than the prose corrections in revisions 1–8, and the reason AC-B3.7 constrains ordering
rather than presence alone. **Rev 10** adds a **fail-closed halt on a new input class** (the
outcome/shipment pair), makes a pre-mutation deferral rule **categorical** rather than enumerated,
and changes the **harness activation lifecycle** that every task's red→green boundary depends on —
the last of which interacts directly with Ship's Step 4.3 full-suite gate and is therefore the
rev-10 element most in need of the mutation-proof checks §5.0.1 specifies.
**Rev 11** replaces that activation lifecycle outright. The rev-10 form was **self-disarming** for
the two functions R5 depends on and did **not** satisfy the installed P-002/P-004 red-phase
postcondition; the replacement introduces a **new tracked control-state artifact**
(`tests/integration/testdata/directpush-gate/harness-state.json`) on which **every** assertion in
the harness now depends, and a **three-phase** lifecycle that every task must transition correctly.
That is the widest hardening signal in this plan: a single state file that, if unguarded, would
silently disable all eight assertions at once. It is the reason MP1 runs inside `gate()` on every
call, MP4 proves ten mutation cases by execution, and — at rev 12 — MP6 cross-checks the terminal
witness before any activation decision (MP5's live no-rollback anchor is **RETIRED**: it read
incomplete tasks' surfaces and deadlocked a legal state) —
and the reason R19 is rewritten and R21 added rather than the mechanism being assumed safe.
**Rev 12** narrows rather than widens that blast radius, but two of its three changes are still
hardening signals in their own right. Retiring **MP5** removes a check that could fail the default
suite from a **legitimate** state, which is a correctness fix, but it also means the plan now
relies on completed tasks' own substantive assertions for rollback detection — a claim that must be
true task-by-task, not assumed. Adding the **terminal completion witness** introduces a **second**
tracked control-state artifact and a cross-agreement rule (**MP6**) that runs before every
activation decision, so a defect in MP6 fails all eight functions exactly as an MP1 defect would.
The third change, the **CI provenance correction**, is a *reduction* in claimed guarantee: it makes
the regeneration-resistance argument depend on a property the recorded render input actually
supports (a full-suite `./...` invocation) and adds the post-render verification (**AC-C1.7**) that
the previous, overstated claim made look unnecessary.

## 10. Constitution Check

Required by `.github/instructions/constitution.instructions.md` (Governance). Omitted in rev 1.

| Principle | Status | Note |
|---|---|---|
| I. Safety-First Go | **Satisfied** | Rev 4: one Go test file is added (§5.0) — no production Go surface; `go vet ./...` and job `expensive`'s full-suite `./...` test step (observed installed command: `go test -race -mod=readonly ./...`) cover it. **Rev 11**: one `testdata/` JSON state file is added alongside it, parsed with stdlib `encoding/json` and fail-closed on every error path — no new dependency, no production surface. **Rev 12**: one further `testdata/` JSON file (the terminal witness), parsed the same way, with its digest computed via stdlib `crypto/sha256` — still no new dependency and no production surface |
| II. Test-First (NON-NEGOTIABLE) | **Satisfied — no deviation** | Rev 4 removes rev 3's declared deviation. See §10.1 |
| III/IV. Workspace isolation, CLI containment | **N/A** | No runtime code |
| V. Structured Observability | **Satisfied** | `::notice::` mode emission; AC-C2.1 records gate verdict as PR evidence |
| VI. Single Responsibility | **Satisfied** | One gate, one invariant; 8 tasks each single-domain |
| VII. Destructive Command Approval | **N/A** | No destructive operation |
| VIII. Explicit Safety Modes | **N/A** | — |
| IX. Git-Friendly Persistence | **Satisfied** | Additive Amendment Log row 1.25.0; §11 records the stale header `Version` field |
| X. Agent Context Efficiency | **Satisfied** | Ledger duplication reduced (§7.1 drops the magic count; residual stated once in the script header, cross-referenced elsewhere) |
| XI. Merge Commit History (NON-NEGOTIABLE) | **Satisfied** | P-009 untouched; **AC-C2.6** (rev 8 — this row previously cited AC-C2.5, which is the markdownlint criterion; AC-C2.6 is the one asserting P-009/P-011/P-016 textually unchanged) |
| Task Granularity (NON-NEGOTIABLE) | **Satisfied** | Rev 3 splits sub-epic A into A1 (fixture data) / A2 (driver + stub) / A3 (detector) and B into B1/B2/B3 — 8 tasks, each single-domain and within 2 h. C1 combines ci.yml + CODEOWNERS + ledger as one **CI-configuration** domain (no code/doc mixing). **Rev 11**: each task additionally makes a one-line insertion into the shared harness-state completion record (§5.0.1) — a state transition in a file H0 authored, not a second skill domain; no task's `size` or `complexity` changes and every task stays inside the 2-hour rule |
| Development Workflow #2 (backlog-driven) | **Satisfied** | These tasks are not a static markdown list: they are harvested into `.backlogit/` under the covering chore with the full P-003 lineage chain before any execution |

### 10.1 P-002 / P-004 compliance — no deviation (rev 4)

**Rev 3 declared a deviation here and was wrong to.** It asked Ship to apply `harness-ready` from a
substitute bash harness by prose. `_ship.agent.md` Step 2 cannot honour that: it invokes
`harness-architect` for every queued task lacking the label and halts unless every one carries it,
and `harness-architect` applies the label only after a Go `_test.go` harness compiles under
`go vet ./...` and fails under `go test ./...`. A Stage-applied label would forge P-004's
postcondition; withholding it would drive `harness-architect` to invent Go scaffolding for markdown
and bash tasks. Both outcomes are P-002/P-004 failures.

**Rev 4 satisfies the existing harness path literally, with no policy change.**

- **Principle deviated**: **none**. P-004's precondition is met as written — `go vet ./...` exits 0
  and `go test ./...` exits non-zero with the expected `not implemented` marker across all eight
  harness functions after H0. **Rev 11 makes this literal rather than approximate**: all eight
  functions activate in `red` phase and all eight fail, each with its **own** task-specific
  `not implemented: <ID> <surface>` marker (§5.0.1 marker table). Rev 10's one-anchor-plus-seven-skips
  form did **not** meet this postcondition and is withdrawn.
- **Harness**: `tests/integration/directpush_gate_test.go` (§5.0), a real Go regression harness
  whose subject is the bash gate. Not invented scaffolding: `tests/integration/build_script_test.go`,
  `start_script_test.go`, and `output_path_guard_test.go` are existing tests in this repository
  whose entire subject is a non-Go script, and `internal/apperr/taxonomy_drift_test.go` is an
  existing Go drift guard over a non-runtime invariant.
- **Red phase**: H0's structural stub (`scripts/check-direct-push-language.sh` printing
  `not implemented: direct-push detector` and exiting non-zero) is the shell analogue of
  `harness-architect`'s `panic("not implemented: <reason>")` stub. The detector is observed failing
  before the implementation that makes it pass — Principle II's requirement, met by execution.
  **Rev 10**: the *default-suite* red at H0 comes from the activation-independent A1 anchor
  (§5.0.1), and each remaining task's red is observed **per task** at claim time through that
  task's `-gatetask`-selected `harness_cmd`. Every task still has an executed red before its
  implementation; what changes is that the eight reds are no longer required to be **simultaneous**,
  which is what made the suite incompatible with Ship's Step 4.3 full-suite gate. No red is
  removed, none is asserted from prose, and Principle II is not relaxed.
  **Rev 11 supersedes the rev-10 sentence above.** Making the eight reds non-simultaneous was the
  wrong trade: the installed contract asks for `go test ./...` red with **all** generated tests
  failing their expected markers, and one failing anchor plus seven skips is not that. Rev 11 restores
  the simultaneous 8/8 red **in the `red` phase** and buys Step 4.3 compatibility from the **phase
  transition** instead — A1's completion moves the manifest to `build`, after which the default suite
  runs the completed set and skips only not-yet-started tasks. So the eight reds *are* simultaneous
  at H0, *and* the suite is green between tasks; the two are no longer in tension because they occur
  in different phases. The per-task targeted red is retained on top of the 8/8, not in place of it.
- **Harness-state manifest (rev 11)**: `tests/integration/testdata/directpush-gate/harness-state.json`,
  initialized by H0 to `{"phase":"red","completed":[],"terminal":false}`. It requires **no** change to
  `.github/skills/harness-architect/SKILL.md`: H0 already emits a non-Go deliverable (the structural
  `check-direct-push-language.sh` stub), and a write-once `testdata/` constant is strictly smaller
  than that. It is **harness state, never backlog status** — never derived from or reconciled against
  `.backlogit/`, so no item-status edit can activate or deactivate an assertion. See §5.0.1 "Why the
  release-specific manifest satisfies the installed contract literally".
- **Ship's P-002 precondition**: satisfied by `harness-architect` at Step 2 in the normal way.
  **Stage applies no `harness-ready` label and this plan asks Ship for no prose exemption.**
  `018-F.custom_fields.harness_status` stays `pending` until H0 records
  `Compilation: PASS` / `Red Phase: CONFIRMED`.
- **Rejected alternative 1 — amend the governing workflow.** A first-class non-Go harness route
  needs edits to P-002/P-004 in `workflow-policies.md`, Step 2 in `_ship.agent.md`, and the
  `harness-architect` skill. The latter two are autoharness-generated from gitignored `.tmpl`
  sources (D2), so the change would be reverted by the next render (R4/R5); and amending P-002/P-004
  is a workspace-wide weakening of test-first binding every future shipment, outside both this
  release unit's branch-policy contract and the governing deliberation.
- **Rejected alternative 2 — ship the detector and fixtures together** (rev 1's Task 1). It never
  observes the detector failing, so a detector that is right for the wrong reason ships undetected.
- **Rejected alternative 3 — rewrite the gate in Go** to avoid the shell subject entirely. It would
  discard the five-gate bash precedent, the shared `--self-test` house style (§5.1), and the CI
  step/CODEOWNERS conventions (§7.1), and put governance-document scanning inside the product
  module. The Go *harness* over a bash *gate* keeps both contracts.

### 10.2 P-016 / P-014 / P-018 disposition for the staging PR

P-016's Statement is scoped to **implementation** branches: "there must be exactly one agent-owned
**implementation branch/worktree** in use." A Stage artifact branch carries no source, test, or
config change, so it is not an implementation branch and does not consume the single-active
implementation slot during planning overlap; it is likewise **non-release-unit under P-001**.

Rev 3 does not leave this as an assertion: B2 adds the same clarification **to P-010's text**, so
the disposition is contract-visible at the point an agent reads it rather than inferred from a
plan. No P-016 amendment is needed — the clarification records an existing scope boundary rather
than creating an exception.

P-014 readiness and P-018 thread resolution apply to the staging PR as to any PR; the
**Orchestrator** produces the readiness record under Step 1.5, since it owns push/PR/merge.

## 11. Risks, residuals, out of scope

| # | Risk | Mitigation |
|---|---|---|
| R1 | Detector self-matches the corpus's own prohibition text | §5.3 structural prohibition resolution; AC-C2.3 **executes** the proof on three real marker-free constructs |
| R2 | Gate wired while violations present → `main` red | §4 ordering; CI wiring is C1, after B1–B3 |
| R3 | Detector too narrow → fails open | 9 reject fixtures across push/commit/prose/placeholder/wrapped-lazy/marker-bearing/`rather than`/mixed-block shapes; bijection + non-vacuity assertions |
| R4 | Regeneration restores a violation | Gate asserts the invariant over installed artifacts (D2) → CI red, scoped to the §5.2 corpus |
| R5 | Regeneration wipes the job-`lint` CI step → gate silently disabled | **MITIGATED at rev 5 — no longer an accepted residual.** Enforcement also runs through `tests/integration/directpush_gate_test.go` (not generated → survives a render), invoked by job `expensive`'s **full-suite `./...`** test step — a step a render *re-emits* — and any render that rewrites `ci.yml` sets `changes.code == 'true'` so that job cannot skip. `TestDirectPushGate_FullCorpusClean` keeps executing the detector (requirement **a**); `TestDirectPushGate_CIWiringIsBlocking` fails **red** when the step is gone (requirement **b**). **Rev 11 repairs the hole rev 10 opened in this very mitigation**: rev 10 activated both of those functions from the presence of the job-`lint` step, so a render that removed the step also removed both detectors and R5 was mitigated only on paper. Activation now derives solely from the harness-state manifest (§5.0.1, AC-C1.6, AC-C2.9), which no render can reach. **Rev 12 corrects the provenance this row rested on**: the observed installed command `go test -race -mod=readonly ./...` is **not** the generated baseline — the template emits `{{TEST_COMMAND}}` and the recorded render input is `go test ./...`. The mitigation therefore rests on the **full-suite `./...` package pattern** surviving, not on the flags, and **AC-C1.7** adds the post-render verification that keeps that true. See §5.0 "Invocation provenance", AC-C1.6, AC-C2.8, AC-C2.9, and **R23** |
| R6 | Gate mistaken for a security boundary | Header states verbatim: **anti-accident control, not an anti-adversary control** — CI runs it from the PR head. CODEOWNERS (AC-C1.4) is the named mitigation |
| R7 | Legitimate line trips the detector | Narrow the **pattern** or document-and-exclude a path; never reduce corpus coverage |
| R8 | 6th hand-copied gate scaffold (no `scripts/lib` bash helper exists) | **ACCEPTED RESIDUAL**, recorded here with the existing clone-divergence precedent (stash `6C24E2E4`); extracting a shared helper is out of scope under C1 |
| R9 | A **fifth** authorization surface exists outside the §5.2 corpus | AC-A3.4 records the first full-corpus scan (file count + complete finding list) **before** B1 begins, so an unknown surface surfaces at A3 rather than at C2 |
| R10 | The §5.0 Go harness makes `go test ./...` depend on `bash` | **ACCEPTED RESIDUAL** (rev 4). CI runs on `ubuntu-latest`; every existing gate script already requires `bash` locally. The harness **fails rather than skips** when `bash` is absent (§5.0), deliberately, so the P-004 red phase can never be satisfied by a silent skip (P-012) |
| R11 | The §5.0 harness is a sixth CI-relevant surface a future render could disturb | Low: it is a plain `_test.go` file under `tests/integration/`, not autoharness-generated, and job `expensive`'s full-suite `./...` test step already runs it. No new CI wiring, no ledger entry, no CODEOWNERS line needed. **Rev 5** makes both of those properties **load-bearing rather than incidental** — R5's mitigation depends on the harness being non-generated *and* on that test step being generated-baseline. **Rev 12 sharpens the second half**: what is generated-baseline is the **step and its `{{TEST_COMMAND}}` substitution**, not the observed `-race -mod=readonly` flags, so the load-bearing property is the `./...` package pattern (AC-C1.7, R23) |
| R12 | The harness itself is deleted or neutered, disabling both halves at once | **ACCEPTED — but VISIBLE, not silent (rev 5).** Requires editing non-generated tracked files (`directpush_gate_test.go` and/or the detector) in a reviewable pull-request diff; **no render can do it**. `CODEOWNERS` is deliberately **not** claimed as the mitigation here — its own header records it is advisory-only until code-owner review is required on `main`, an operator action not taken. This is strictly narrower than rev 4's R5: the failure mode moves from *silent regeneration* to *visible human/agent edit*, which is what §1 actually promises to exclude. **If a future autoharness version begins generating `tests/**`, the §5.0 invariant must be re-verified** |
| R13 | A render reverts the rev-6 handback/discovery contract, restoring main-only discovery — a defect the **detector cannot see**, since it is neither §5.3 construct | **MITIGATED, same mechanism as R5 (rev 6).** The two per-surface harness functions carry it: `TestDirectPushGate_OrchestratorSurfaceClean` asserts the branch-specific discovery, the handback requirement, both step-4 arms, and the **absence** of the default-branch self-range literal; `TestDirectPushGate_StageAgentSurfaceClean` asserts the Step 6 emission. Both live in the non-generated `directpush_gate_test.go` and run from job `expensive`'s **full-suite `./...`** test step (§5.0 "Invocation provenance"; observed installed command `go test -race -mod=readonly ./...`) — so a render that reverts either surface fails CI **red**. The literal-absence assertion is sound only because the corrected text states its prohibition descriptively (§6.1.2, self-match hazard; AC-B1.6) — a prohibition quoting the forbidden range would self-match and silently vacate the assertion |
| R14 | A verification loop passes over targets that are **always present**, proving nothing — the rev-7 finding | **MITIGATED (rev 7, widened rev 8).** Structural, not procedural: a two-tree `git diff --name-only` cannot emit a directory, `cat-file -t` must print `blob`, the derived set must be non-empty, and a reported set must **equal** the aggregate changed-file set. **Rev 8** removes the residual rev 7 accepted rather than closed: the set is now the **aggregate** `{stage_base_commit}`→`{stage_head_commit}` diff, so a Stage session spread over several commits has **every** artifact read as a file on `origin/main`, not only its tip commit's. `TestDirectPushGate_OrchestratorSurfaceClean` additionally asserts the **absence** of a `git show origin/main:{path}` loop, of any directory-prefix default, and (rev 8) of any single-commit `diff-tree` derivation in the no-shipment arm, so a render or edit that restores either earlier form fails CI **red**. **Residual, recorded not hidden**: if a Stage session merges `origin/main` into its artifact branch mid-flight, the two-tree diff also reports files that arrived from the default branch. They are bounded to `STAGE_ARTIFACT_ROOTS` by the pathspec and are already on `origin/main`, so the per-file check passes — the effect is a few **extra** verified paths, never a missed one. Widening to a range form is **rejected**, not overlooked: `origin/main..{stage_head_commit}` is empty by construction post-merge, which is exactly the vacuous loop this row exists to close |
| R15 | The `no-shipment` terminal outcome becomes a dumping ground for genuine failures | **MITIGATED by construction (rev 8).** AC-B3.6 makes `no-shipment` a **success** terminal available only to a reviewed, valid **empty** harvest, and states explicitly that P-003 lineage violations, harvest failures, and a missing required shipment after a **non-empty** harvest continue to **halt**. Step 5.5's existing empty-harvest / unresolved-P-003 guardrail is preserved **verbatim** rather than rewritten, so the discriminator between "nothing to ship" and "failed to ship" is the one already in force. `TestDirectPushGate_StageAgentSurfaceClean` asserts the halt clauses are still present alongside the new terminal. **Residual**: an agent could still mis-classify a failed harvest as empty — a judgement error the contract narrows but cannot eliminate; the Orchestrator's step-4 no-shipment arm is the backstop, since a mis-classified run with no artifacts on `origin/main` fails check 5 or check 6 |
| R16 | `stage_base_commit` is captured wrongly — too late (missing early commits) or too early (over-wide set) | **BOUNDED (rev 8).** Capture is a single `git rev-parse HEAD` by the **Orchestrator immediately before it invokes Stage**, at a point where Stage has not yet run and cannot have committed — so "too late" requires the Orchestrator to reorder its own Step 1, which `TestDirectPushGate_OrchestratorSurfaceClean` asserts against. "Too early" is bounded by the `STAGE_ARTIFACT_ROOTS` pathspec and fails **conservatively** (extra paths verified, never fewer). The retained value is authoritative over any Stage-reported one, so a wrong report cannot shrink the set; and base→head ancestry (check 3) rejects an unrelated or rewritten base outright rather than silently producing a nonsense diff. **Rev 9** supplies the direct-invocation half the row previously left implicit: Step 1.9 captures the base from `HEAD` **before** it creates or checks out the branch, and never re-captures it, so the same "too late" argument holds on the unbracketed path |
| R17 | A render reverts the rev-9 operative steps, restoring a Stage agent that is *permitted* to branch and commit but never *instructed* to — a defect the detector cannot see, since it is neither §5.3 construct | **MITIGATED, same mechanism as R13 (rev 9).** `TestDirectPushGate_StageAgentSurfaceClean` asserts both steps present **and ordered**, by parsing the Step Sequence Contract checklist and comparing entry indices and heading positions rather than by text search (§6.3.2). It lives in the non-generated `directpush_gate_test.go` and runs from job `expensive`'s **full-suite `./...`** test step (observed installed command `go test -race -mod=readonly ./...`; §5.0 "Invocation provenance" — the *step and its `{{TEST_COMMAND}}` substitution* are render-baseline, the flags are not), so a render that drops either step fails CI **red**. The structural form is load-bearing: `_stage.agent.md` names its own steps in the Step 6 gate, so a bare name search would pass with the step sections deleted — the AC-C1.6 hazard, one file over |
| R18 | An artifact is written before the branch gate runs, landing on the default branch | **CLOSED BY ORDERING, with a narrow residual (rev 9).** The gate occupies a **fixed** checklist position (after learnings retrieval, before deliberation), and every mutation that would otherwise precede it — the Step 1 deferred-expansion duplicate archival, the session's first memory checkpoint — is **deferred until after** it by an explicit deferral rule rather than by racing the gate earlier. **Rev 10 closes the under-inclusion**: the rule was an *enumeration* of two mutations, while the installed Stage contract also schedules a stash-classification checkpoint, a contextual-grouping/operator-selection checkpoint, and the **write half** of the mandatory late-identifier reconciliation before that ordinal — each of them tracked, and each therefore still legal on the default branch under the rev-9 wording. The rule is now **categorical** ("every tracked Stage artifact mutation"), with those classes enumerated as illustration rather than as the closed set, and with gitignored disposables expressly excluded rather than silently claimed. The rev-9 draft's alternative, letting the gate fire early against a provisional slug, was **rejected**: it contradicts the Step Sequence Contract's own "execute in order" semantics, leaves the real position unassertable, and re-admits the default-branch write it was meant to prevent. `TestDirectPushGate_StageAgentSurfaceClean` asserts both the ordinal and the deferral clause. **Residual, recorded not hidden**: an agent that writes before reaching Step 1.9 violates the contract — now a **P-005** step-order violation as well as P-010 — without tripping the harness, which reads the document, not the run. The Orchestrator's step-4 arm is the backstop: artifacts written on the default branch leave `stage_branch` UNRESOLVED and the aggregate diff empty, which halts the gate rather than passing it |
| R19 | The activation mechanism degenerates — assertions skip forever and the suite is green while proving nothing | **REWRITTEN AT REV 11 — the rev-10 form had already degenerated.** Rev 10 keyed default-mode activation on **surface predicates**, so C1's activation was the very proposition C1 asserts and C2 reused it verbatim: a render dropping the job-`lint` step disarmed **both**, MP2 (inside C2) never ran, and `go test ./...` stayed green. Rev 11 removes the predicate→activation edge entirely — activation derives **only** from the tracked harness-state manifest and the `-gatetask` selector, so no assertion's activation can be switched off by mutating its own subject. Six executed checks bound the replacement: **MP0** fixes `gateOrder`/`gateDeps` at eight IDs matching the recorded `blocks` edges; **MP1** runs inside `gate()` on **every** call and fails closed on a missing, malformed, unknown-phase, duplicate, unknown-ID, non-subsequence, dependency-open, or phase/terminal-inconsistent manifest — so a corrupt state file fails all eight functions rather than silently skipping them; **MP2** requires the exact eight-element ordered set with `terminal: true`, evaluated from manifest state **before** any surface assertion in C2; **MP3** makes every skip a named `t.Skip` and an unknown selector a **failure**, with skips structurally impossible in `red` and `terminal`; **MP4** proves ten manifest- and witness-mutation cases fail by execution against a `t.TempDir()` copy (AC-C2.2); **MP6** (rev 12) cross-checks the terminal witness against the manifest inside `gate()` **before** any activation decision, so a self-consistent manifest rewind can no longer skip C2 past MP2. **MP5 is RETIRED at rev 12**: its live no-rollback anchor read *incomplete* tasks' surface probes and failed when one was present, but C2's probe was the job-`lint` step **that C1 creates**, so the entirely legitimate post-C1/pre-C2 window failed the default suite — the Step 4.3 deadlock, reintroduced by the check meant to guard against degeneration. Its function is covered without it: a **completed** task's rollback is caught by that task's own substantive assertion (active in default mode from its completion onward), and manifest/witness rollback is caught by MP1 and MP6. **Residual, recorded not hidden**: a rollback in which the manifest **and** the witness **and** the corresponding surface are all reverted together is not distinguishable at runtime from a legitimate mid-shipment state, so it is caught by review of a tracked-file diff rather than by a live assertion — the R22/R21/R12 class, a visible edit, never a render artifact. **Rev 13 closes a narrower degeneracy the rev-12 table admitted**: because the witness row preceded the `-gatetask` rows and was unqualified by mode, a *targeted* run in terminal state would have activated all eight bodies, contradicting MP3 and making every targeted observation ambiguous. `gate()` is now split into a **mode-independent integrity gate** and a **mode-dependent activation resolution** in which the selector takes **precedence**; MP6's integrity half still fails a targeted invocation, only its forcing half is default-mode. Separately, the §5.0.2 point-7 lint-removal evidence — previously taken with `-gatetask=C1`, which activates C1 **by construction** and would have passed under the rev-10 self-disarming design — is now taken **unselected** against a copied terminal-state fixture and asserts **both** C1's failure and C2's execution by test-event name. **Rev 14 closes the last gap in that evidence: a specified-but-unexecutable observation is a degeneracy of the same family.** Rev 13's point 7 named a bare default-mode `go test` against a copy, but `repoRoot(t)` resolves the **real checkout** by walking up to `go.mod` from the process working directory, and the observation lives **inside** C2, so as written it could neither see the fixture nor run without recursing — the evidence would have been reported from the wrong tree or not at all. The mechanism is now concrete and bounded (tracked-file copy helper, child `go test -json` process rooted at the copy, non-recursive sentinel, parsed-event assertions, explicit timeout), and its own failure modes are carried as **R24** |
| R20 | Step 5.7's out-of-root halt fires on a human operator's unrelated in-flight edit, blocking the Stage commit | **ACCEPTED AND DELIBERATE (rev 10).** Fail-closed is the correct direction: the alternative is a Stage commit that silently omits a change the operator believed was being saved, or one that silently widens beyond `STAGE_ARTIFACT_ROOTS`. The halt names every offending path, and Stage **leaves the change exactly as found** — no checkout, restore, stash, revert, or delete — so nothing is lost and the operator decides. Discarding it would be a destructive act without approval (Constitution VII). Gitignored paths never reach this check, so routine index/checkpoint/hook-queue churn cannot trip it |
| R21 | The harness-state manifest becomes a **single point of failure**: one tracked file whose corruption, deletion, or rollback could disable all eight assertions at once | **BOUNDED, AND THE TRADE IS DELIBERATE (rev 11).** Centralizing activation state is what makes it *checkable* — the rev-10 alternative distributed the same power across eight surface predicates where it was invisible and, in C1/C2's case, self-referential. The file is bounded four ways. **(1) Fail-closed by default**: MP1 runs inside `gate()`, so a missing, unreadable, malformed, or invalid manifest fails **all eight** functions loudly; the degenerate direction is red, not green. **(2) Not skippable**: no manifest state produces a skip in `red` or `terminal` phase, and targeted mode ignores the manifest for **activation** entirely, so `-gatetask={ID}` still executes the real assertion even against a manifest whose `phase`/`completed` state has been rewound. **Rev 13 precision**: "ignores" is scoped to activation only — a manifest that is *invalid* (malformed, unknown phase, duplicate/unknown ID, non-subsequence, dependency-open, or an illegal triple) still **fails closed in every mode**, because MP1 lives in `gate()`'s mode-independent Stage 1. Tampering therefore either fails the run or fails to suppress the assertion; it cannot do neither. **(3) Outside every render boundary**: it lives under `tests/**`, which no autoharness template covers (§5.0 render table), so — unlike the job-`lint` step — **no regeneration can touch it**. **(4) Executed mutation proof**: MP4/AC-C2.2 prove ten corruption and rollback cases fail rather than skip. **(5) Rev 12 — the terminal claim is no longer single-file**: `terminal-witness.json` is an independent tracked file, and MP6 requires the two to agree **before** activation, so neither a manifest-only nor a witness-only rollback can pass. **Accepted residual, same class as R12**: deleting or rolling back the state files *together with* the surfaces they gate remains possible through a reviewable pull-request diff to non-generated tracked files. That is a **visible human or agent edit, not a silent render**, which is exactly the boundary §1 promises. **If a future autoharness version begins generating `tests/**`, this row and R11/R12/R22 must be re-verified together** |
| R22 | **Coordinated rollback of BOTH state files in one change** (rev 12) — the witness deleted and the manifest rewound together, reproducing a state that looks exactly like legitimate pre-C2 `build` | **ACCEPTED AND STATED HONESTLY — NOT claimed as runtime-detectable.** MP6 closes the *asymmetric* cases: witness-only and manifest-only rollbacks each fail loudly, which is what a partial write or a single-file edit produces. It does **not** close the symmetric one, and the plan does not pretend otherwise: a validator that reads only the current tree has **no history to compare against**, so a state that is byte-identical to a legal earlier state is, to it, that state. Claiming a stateless validator detects every coordinated rollback would be false, and inventing a runtime history mechanism (a signed log, an append-only ledger, a CI-side cache) is a materially larger control than this release unit's branch-policy contract warrants. What does catch it: **both files are tracked and non-generated**, so the rollback appears as a reviewable diff in a pull request — deleting the witness is a visible file deletion and rewinding the manifest is a visible content change, both under `tests/integration/testdata/directpush-gate/`. Caught by **diff and review**, not by runtime history. Same visible-edit boundary as R12 and R21 |
| R23 | **A future render narrows the expensive-job test command's package pattern**, so `tests/integration` — and therefore the whole harness — stops running, silently reopening R5 | **BOUNDED BY AN EXPLICIT POST-RENDER CHECK (rev 12), and the previous overstatement withdrawn.** Revisions 5–11 asserted `go test -race -mod=readonly ./...` *was* the generated baseline, which made this risk invisible: if the exact command were render-guaranteed there would be nothing to verify. It is not — the template emits `{{TEST_COMMAND}}` and the recorded render input here is `go test ./...` (§5.0 "Invocation provenance"), so the flags are local and only the **step plus its substitution** is re-emitted. The guarantee is therefore scoped to what the baseline supports — **a full-suite invocation whose package pattern is `./...`** — and **AC-C1.7** makes verifying that pattern a written, standing obligation in the LOCAL DIVERGENCE ledger after every regeneration of `ci.yml`. **Residual**: the check is **procedural, not mechanical** — nothing in CI asserts the rendered pattern, because a check on the rendered file would itself sit inside the render boundary. If the obligation is skipped and a render narrows the pattern, R5 reverts to open. Recorded, not hidden |
| R24 | **The §5.0.2 point-7 child-process mechanism misbehaves** (rev 14) — it recurses, hangs, floods the log, or passes on an unreadable event stream, turning the R5 evidence into either a CI outage or a false green | **BOUNDED BY FOUR EXPLICIT CONSTRUCTION RULES, each stated normatively in §5.0.2 point 7 rather than left to the implementer.** **(1) Recursion is bounded at exactly one level** by the `DIRECTPUSH_GATE_FIXTURE_CHILD=1` sentinel: the child never enters the parent-only spawn block. **Rev 15 corrects how that bound is EVIDENCED.** Rev 14 additionally claimed the parent "treats an already-set sentinel at that block as a **recursion marker and a failure**" — an **unsatisfiable** condition, because the guard means a sentinel-set process never **reaches** the block, so that detector could never fire and proved nothing. Recursion is therefore **PREVENTED structurally** (entry requires the sentinel to be **unset**), not detected; the **timeout** in rule (2) is the real runtime bound, and a defensive check that the parent's own `exec.Cmd.Env` carries the sentinel is permitted as a self-check but is **not** recorded as recursion control. **(2) Hangs are bounded** by an explicit `context.WithTimeout` on `exec.CommandContext`; a deadline termination is a **failure of the observation**, never an inconclusive pass. **(3) Log volume is bounded** — child stdout/stderr are captured and surfaced **only** as a truncated tail **on failure**, so a passing run adds nothing to CI output. **(4) A malformed, truncated, or terminal-event-less `-json` stream FAILS**; the parent never reads an unparsable stream as evidence in either direction, and a **zero** child exit fails too, because C1 must detect the removed step. **Cost, accepted and stated**: the child compiles and runs `tests/integration` a second time against a ~5 MB tracked-file copy, bounded to two test functions by `-run`. That is the price of observing **default-mode** activation at all — an in-process observation is impossible here, because `repoRoot(t)` resolves the real checkout by construction, and the rev-13 attempt to specify the observation without a mechanism is exactly what this row replaces. **Residual, recorded not hidden**: the mechanism asserts activation against a **copy**, so it proves the manifest-derived activation property, not the state of the real `ci.yml` — which is what AC-C1.6's ordinary structural assertion against the real tree covers. The two are complementary and neither is claimed to subsume the other |
| R25 | **The point-7 fixture is not a faithful, self-contained subject** (rev 15) — the copy silently omits a file, is not a Git repository, resolves a **different** repository, or its no-mutation evidence deadlocks — so point 7 fails, or passes, for reasons unrelated to the lint step | **FOUR CONSTRUCTION DEFECTS CLOSED EXPLICITLY, each with a fail-closed check rather than an assumption.** **(1) Silent omission by index mode.** Enumeration is `git ls-files -s -z`, and mode handling is a **closed set**: `100644`/`100755` copied (executable bit preserved), while `120000`, `160000`, any other mode, and any **unmerged stage** **fail closed**. Measured on this repository, **all 629 index entries are `100644`** (zero `100755`/`120000`/`160000`, no `.gitmodules`), so **no skip-list is written** — notably **none for `references/`, which is not in the index at all**; it holds the nested independent clone `references/herdr`, excluded by `.git/info/exclude`. The fail-closed rows are a **forward guard** that turns a future symlink or submodule into a loud failure instead of a quietly incomplete fixture. **(2) The untracked terminal witness.** `terminal-witness.json` is written by **C2's own completion** and may not be indexed when the helper runs, so an index-only copy yields a **witness-absent + terminal-manifest** fixture that MP6 fails as an **integrity** error — vacuity for a reason unrelated to the lint step. It is therefore copied **explicitly from the working tree**, **required to exist**, parsed, and **validated against the copied manifest**, **before** the lint step is removed. **(3) Not a repository, or the wrong one.** The real `.git/` is never copied, so the copy would otherwise have no repository and Git commands run inside it would **walk upward** to an unrelated ancestor. The helper runs `git init` + a **local, non-secret** identity + `git add` + a deterministic **local** commit (no network, no submodule recursion, no `--global` state), and asserts `git -C {copy} rev-parse --show-toplevel` **equals the canonical copy root** before the child runs — which is what makes "no nested or sibling repository can be selected" **checkable**. A commit rather than a bare index is preferred so tracked-corpus **and** `HEAD` consumers both work. **(4) The no-mutation evidence deadlocked.** The rev-14 empty-`git status` precondition was **unsatisfiable exactly when needed** (C2's own witness/manifest dirt, plus any path excluded only by the local non-portable `.git/info/exclude`); it is replaced by **byte-identical** `git status --porcelain=v1 -z --untracked-files=all` snapshots before and after. **Residual, recorded not hidden**: the fixture's Git history is **synthetic** — one local commit, not the real branch history — so point 7 proves nothing about real commit ancestry, and any future assertion needing genuine history must say so and build it explicitly |

**Known pre-existing defects, recorded and out of scope**: `workflow-policies.md` header
`**Version**: 1.0.0` is stale against Amendment Log 1.24.0; Step 1.5's a–e sub-steps are lazy
paragraph continuations rather than a real nested list (so AC-B1.3 is **textual**, not structural);
`_orchestrator.agent.md`'s provenance footer names `orchestrator.agent.md.tmpl` while the real
template is `_orchestrator.agent.md.tmpl`; `ci-topology-check.sh` cites two non-existent
`docs/pipeline-topology-gate*.md` files; job `lint` runs only when `changes.outputs.code == 'true'`
(benign — the `code` filter excludes only `docs/**`, `.backlogit/**`, `.backlog/**`,
`.autoharness/**`, so any `.github/**` edit triggers it).

**Out of scope**: reverting/rewriting `fdff9e4`; editing untracked `.copilot/` or
`.autoharness/staging/` templates; GitHub branch-protection configuration; a new policy ID;
**any amendment to P-002, P-004, `_ship.agent.md` Step 2, or the `harness-architect` skill**
(rev 4 — §10.1 rejected alternative 1); **any amendment to `_ship.agent.md` Step 4.3's full-suite
scope** (rev 10 rejected alternative 1, reaffirmed rev 11 — the harness-state phase makes narrowing
it unnecessary as well as unwanted); a shared `scripts/lib/gate-common.sh`; the missing compound
entry on generated-artifact divergence; the upstream template defect report.

---

## 12. Plan review record

**Gate**: `plan-review`, 7 personas (Architecture Strategist, Correctness, Scope Boundary Auditor,
Constitution, Maintainability, Template Integrity, Schema-CLI-Docs Coupling).

| Attempt | Revision reviewed | Verdict | Findings |
|---|---|---|---|
| 1 | rev 1 | **FAIL** | 3 x P0, 11 x P1 |
| 2 | rev 2 | **FAIL** | all round-1 P0s confirmed resolved; 1 new P0 (fourth surface), 5 P1 detector defects |
| — | rev 3 | **ADVISORY — accepted** | all P0/P1 remediated in rev 3; residual P2/P3 recorded below |
| — | rev 4 | **PR #54 current-HEAD review, cycle 2** | 2 x P-021 C1 same-contract blockers, both resolved in rev 4 (see below) |
| — | rev 5 | **PR #54 current-HEAD review, cycle 4** | 1 finding — R5 contradicted §1 and the feature DoD; resolved in rev 5 (see below) |
| — | rev 6 | **PR #54 current-HEAD review, cycle 5** | 1 finding — Step 1.5's discovery could not see the Stage branch; resolved in rev 6 (see below) |
| — | rev 7 | **PR #54 current-HEAD review, cycle 6** | 2 findings — the no-shipment arm verified directory prefixes that always exist (false pass); the session memory cited an obsolete plan revision and a nonexistent section. Both resolved in rev 7 (see below) |
| — | rev 8 | **PR #54 adversarial review** (anchor GPT-5.6 Sol + GPT-5.4-mini + Claude Sonnet 5 + Claude Opus 5; no route degradation) | 4 P1 blockers + 5 P2/P3 precision findings, all classified **same-contract completion** under P-021. All resolved in rev 8 (see below) |
| — | rev 9 | **PR #54 current-HEAD review, cycle 7** | 1 visible Copilot finding — B3 granted the branch permission and declared the handback but added no operative step that creates the branch or commits the artifacts. Resolved in rev 9 (see below) |
| — | rev 10 | **PR #54 second adversarial review** (anchor GPT-5.6 Sol + GPT-5.4-mini + Claude Sonnet 5 + Claude Opus 5; no route degradation) | 4 visible Copilot threads (3 P1, 1 P2) + 1 adversarial-only P1 harness-lifecycle finding, all classified **same-contract completion** under P-021. All resolved in rev 10 (see below) |
| — | rev 11 | **PR #54 third adversarial review** (anchor GPT-5.6 Sol + GPT-5.4-mini + Claude Sonnet 5 + Claude Opus 5; 4/4 usable, no route degradation) | 3 visible Copilot threads from review `5185014458` — **2 unanimously-confirmed P0 direct-completion blockers** (self-disarming C1/C2 activation; H0 failing the installed P-002/P-004 red-phase contract) + 1 P2 (stale deleted-`3e` DoD reference), all classified **same-contract completion** under P-021. All resolved in rev 11 (see below) |
| — | rev 12 | **PR #54 post-remediation adversarial re-review** (anchor GPT-5.6 Sol + GPT-5.4-mini + Claude Sonnet 5 + Claude Opus 5; 4/4 usable, no route degradation) | **3 P1 residuals** in the rev-11 mechanism (MP5 deadlocked a legitimate post-C1/pre-C2 state; a consistent terminal-manifest rollback could skip C2 before MP2; CI regeneration provenance overstated) + 8 direct consistency defects, all classified **same-contract completion** under P-021. All resolved in rev 12 (see below) |
| — | rev 13 | **PR #54 final capped post-remediation adversarial re-review** (anchor GPT-5.6 Sol + GPT-5.4-mini + Claude Sonnet 5 + Claude Opus 5; 4/4 usable, no route degradation), under the **first** explicit operator authorization of one extra remediation + re-review cycle | **6 `post_remediation_residual` findings** (1 MEDIUM P1 + 4 LOW P1 + 1 LOW P2), none P0, all **same-contract completion** under P-021. All resolved in rev 13 (§12.13) |
| — | rev 14 | **PR #54 rev-13 post-remediation adversarial re-review** (four-model, no route degradation), under the **second** explicit operator authorization of one remediation + re-review cycle | **NOT READY** — **0 P0**, **1 LOW-confidence P1** (the copied-fixture evidence named no executable fixture-root/child-process mechanism and no re-entry guard) + **4 LOW-confidence P2** (H0 eligibility citation; targeted C2 red anchor + bookkeeping qualification; stale memory index; stale deliberation addenda index), all `post_remediation_residual` and **same-contract completion** under P-021. All resolved in rev 14 (§12.14) |

**Round-1 P0s (resolved in rev 2)**: missing third authorization surface `_stage.agent.md` L42;
self-deadlocking P-010 wording (mandated "merged via PR" against Stage's "must not create, push,
or merge pull requests"); false claim that deleting 3e loses no behaviour.

**Round-2 P0 (resolved in rev 3)**: Step 1.5 **preamble** is a fourth authorization surface —
commit-shaped, marker-free, in-corpus. Addressed by §2 (four surfaces) and B1's preamble rewrite
to a verification/postcondition form.

**Round-2 P1s (resolved in rev 3)**: block-level verdicts masked mixed blocks (→ per-construct);
ATX-heading-only prohibition context would have flagged the real bold lead-in shapes (→ bold
lead-in + column header); row-scoped table matching paired Allowed verbs with Forbidden tokens
(→ cell-scoped); unguarded `git symbolic-ref` aborts under `set -euo pipefail` (→ guarded with
`main` fallback); `pull`/`checkout` not excluded (→ added); Task A1 spanned 11 files (→ A1/A2/A3).

**Residual non-blocking findings (P2/P3), accepted**: the stale `**Version**: 1.0.0` header and
other pre-existing defects in §11; the `(0NN.00X-T)` ci.yml step-name suffix is a **placeholder**
to be resolved by Ship at C1 against the then-current ledger numbering; the LOCAL DIVERGENCE
membership rule's literal `011.0xx-T` wording may need a wildcard/range form; no compound entry
yet exists for generated-artifact divergence (out of scope, §11).

**Accepted at attempt 2 under the 2-cycle limit.** No P0 or P1 remains open.

**Rev 4 — PR #54 current-HEAD review, remediation cycle 2.** Two unresolved threads, both P-021 C1
same-contract blockers, both fixed rather than deferred:

| Thread | Finding | Resolution |
|---|---|---|
| `PRRT_kwDOTPuhps6hl8My` (plan L547) | Rev 3's §10.1 substitute harness does not satisfy Ship's actual execution gate; Step 2 invokes `harness-architect` for every task lacking `harness-ready`, and that skill requires Go `_test.go` harnesses plus `go vet`/`go test` red-phase evidence. Prose labels leave 017-S unable to pass P-002/P-004 without out-of-scope Go scaffolding. | **§10.1 deviation deleted.** New §5.0 defines the H0 harness contract: a real Go regression harness driving the bash gate, produced by `harness-architect` at Ship Step 2, with a not-implemented shell stub as the structural stub and eight per-task functions as `build-feature` boundaries. No policy amended, no test-first requirement weakened, no fake scaffolding. §5.5 (A2) retargeted to driver-only, since a task whose content is "create the harness" is unreachable in Ship's flow. Governing-workflow amendment explicitly rejected and recorded in §11 Out of scope. |
| `PRRT_kwDOTPuhps6hl8Nd` (`018.011-T:26`) | The task's AC contradicted §7.2 C2.2's fixture-backed mutation proof by mandating a dirty-tree reintroduction edit; C2 numbering also diverged, falsifying the "all 8 task contracts match" claim. | **§7.2 rewritten as the single C2 verification contract**, fixture-backed with an explicit no-dirty-tree-edit rule and a `git status --porcelain` empty check, renumbered AC-C2.1…AC-C2.7. `018.011-T` now carries that contract **verbatim**. The dirty-tree form is explicitly rejected in the criterion text so it cannot be reintroduced by paraphrase. AC-A2.4's non-vacuity proof was aligned to the same rule (temporary fixture copy under `t.TempDir()`). |

**Residual accepted at rev 4**: R10 (`go test` now depends on `bash`) and R11 (harness as an
additional surface), both recorded in §11.

**Found by the rev-4 verification sweep, not by review** — a third instance of the same
plan↔task same-contract defect class: §6.1's B1 acceptance criteria still carried the three-criterion
rev-2 form while `018.007-T` carried the four-criterion rev-3 form covering the preamble surface,
and §6.3 already referenced the then-nonexistent **AC-B1.4**. §6.1 is corrected to the task's form.
An AC-ID parity sweep across all eight tasks now shows an exact bijection (38 plan IDs ↔ 38 task
IDs, no orphan on either side), so the PR's "all 8 task contracts match" claim is true as stated.

**Rev 5 — PR #54 current-HEAD review, remediation cycle 4.** Operator-authorized fourth cycle
beyond the three-cycle cap, scope limited to thread `PRRT_kwDOTPuhps6hnPkD` (comment 3992775719,
plan L720). No other scope reopened.

| Thread | Finding | Resolution |
|---|---|---|
| `PRRT_kwDOTPuhps6hnPkD` (plan L720) | R5 contradicted §1's guarantee and the feature DoD. Because `ci.yml` is itself generated, a future render could both restore the scanned direct-push wording **and** remove the job-`lint` gate step, leaving CI green with the detector never invoked. The plan had to either add a tracked enforcement path surviving regeneration, or narrow the objective and admit silent disablement. | **Enforcement path added; objective NOT narrowed.** The tracked, regeneration-resistant path already existed in the release unit but was never claimed: the §5.0 Go harness is non-generated, and it is invoked by the **generated-baseline** `go test` step in job `expensive` — which a render *re-emits* rather than removes. New §5.0 subsection "Regeneration-resistant enforcement" proves the four-link chain (detector, assertion, invocation, trigger) and shows the trigger link is airtight: the same edit that removes the job-`lint` step is a `.github/**` change, so `changes.code == 'true'` and the harness job cannot skip. **AC-C1.6** makes the step-presence assertion YAML-structural (a text grep would self-match C1's own ledger comment) and proves its non-vacuity by execution against a `t.TempDir()` copy. **AC-C2.8** records that the detector still runs from `go test` with the C1 step gone. R5 reclassified accepted→mitigated; **AC-C1.5** disclosure reworded from "accepted residual" to "detected, not accepted"; ledger wording in §7.1 updated to match. |

**Honest residual, deliberately not fabricated (rev 5).** The guarantee established is exactly
§1's: *regeneration cannot **silently** reopen the gap*. It is **not** a claim that the control
cannot be removed at all — deleting the harness in a reviewable PR diff still disables it, recorded
as **R12**. `CODEOWNERS` is explicitly **not** claimed as mitigation, because this repo's
`CODEOWNERS` header states it has no enforcement effect until code-owner review is required on
`main`, which is deliberately not enabled. Two rejected alternatives are recorded in §5.0:
`secret-scan-history.yml` (non-generated and would survive, but declares itself NOT a PR-required
check by design) and `ci-topology-check.sh` (itself generated, same boundary).

**Scope discipline**: no new file, workflow, CI step, ledger entry, or CODEOWNERS line was added —
only two acceptance criteria (AC-C1.6, AC-C2.8), one reworded (AC-C1.5), and strengthened assertions
in a harness H0 already produces. AC-ID parity re-swept: **40 plan IDs ↔ 40 task IDs**, exact
bijection maintained. Shipment 017-S remains one shipment with unchanged membership and unchanged
dependency order (A1→A2→A3→{B1,B2,B3}→C1→C2).

**Rev 6 — PR #54 current-HEAD review, remediation cycle 5.** Operator-authorized fifth cycle,
scope limited to thread `PRRT_kwDOTPuhps6hqWDe` (comment 3993998429, plan L517). No other scope
reopened; the separate thread `PRRT_kwDOTPuhps6hqWD8` is an Orchestrator-owned incident-record
correction and is explicitly **not** touched here.

| Thread | Finding | Resolution |
|---|---|---|
| `PRRT_kwDOTPuhps6hqWDe` (plan L517) | Revisions 1–5 corrected what Step 1.5 **authorizes** but never its **discovery** logic. Step 1.5 inspects only `.backlogit/` dirtiness and `git log origin/main..main`, not the checked-out branch/HEAD. Once B2/B3 put Stage's commit on `chore/stage-{chore-slug}` and leave the worktree there, both inputs read "synchronized"; the gate skips its commit/push/PR step and then fails the `origin/main` manifest check — and for a shipment-less run there is no manifest to check at all. The **no-shipment route documented in §6.2 was unexecutable as written**. | **B1 expanded, no task added.** §6.1 restructured into §6.1.1 (authorization, rev 1–5) and §6.1.2 (discovery/verification, rev 6) with four corrections: (1) Step 1 requires an explicit **staging handback record** (`stage_branch`, `stage_head_commit`, `stage_outcome`, `stage_artifact_paths`) with a **total** set of deterministic defaults, so the degraded path never halts and never substitutes the default branch; (2) a **branch-resolution block above step 1's dirty/clean split** carries `rev-parse --verify` and the HEAD-equality assertion (P-016), and step 1's pathspec widens to the Stage artifact paths; (3) step 2 compares `origin/main..HEAD` against the already-resolved branch; (4) step 4 gains a `no-shipment` arm verifying `git merge-base --is-ancestor` plus per-path `git show origin/main:{path}` over a **non-empty** path set — a concrete target naming no `shipment_id`. Rev 2's "no no-shipment clause here" position is **explicitly withdrawn** with its inconsistency against §6.2's own route table recorded. B3 gains the producing half (Step 6 emission, AC-B3.5). Five new B1 criteria (AC-B1.5–B1.9) and one new B3 criterion. |

**Four defects found by the rev-6 internal review, before commit.** The first draft of these
corrections was reviewed against the finding and four high-confidence problems were fixed rather
than shipped:

1. **The degraded handback hard-halted on the population it claimed not to wedge.** Halting when
   the `HEAD` resolution yielded the default branch would have blocked every pre-B3 Stage
   invocation — the exact deadlock the paragraph three lines below it said it had avoided — and it
   made 3a's create-arm dead prescription, since the halt sat upstream of it. Now the branch is
   left **UNRESOLVED** and the run routes to 3a, restoring pre-change behaviour.
2. **`stage_outcome` and `stage_artifact_paths` had no degraded default**, leaving step 4's arm
   selection undefined; and an empty path set made the `no-shipment` arm **vacuous** — a zero-run
   `show` loop plus a trivially-true ancestry check would have passed the gate having verified
   nothing. Defaults are now total, and an empty path set is an explicit `STAGING_GATE_FAIL`.
3. **The new guards were unreachable on the dirty path.** Step 1 routes a dirty tree straight to
   step 3, so guards placed in step 2 were skipped on the one path that mutates the repository,
   falsifying AC-B1.9 as written. The resolution block was hoisted **above** step 1.
4. **Step 4's prescribed text could not satisfy its own AC-B1.4.** An outer lead sentence replaced
   the pre-change one and the original reappeared lowercased. The arm label is now a separate line
   above the byte-verbatim original, and AC-B1.4 states that no outer lead may replace or reword it.

Recorded because a review-fix cycle that silently repairs its own first draft teaches nothing; the
defect classes here (a halt contradicting its own rationale, a vacuous verification loop, a guard
placed on the unreached branch of a fork) are the ones worth recognizing again.

**Self-match hazard caught during rev 6, not by review.** The natural phrasing of correction 3 is a
prohibition quoting the literal `origin/main..main`. In `_orchestrator.agent.md` — which **is** in
the scanned corpus — that literal would match the very absence assertion that rejects the
regression, reproducing the failure documented in
`docs/compound/2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md` and already
guarded at §5.0/AC-C2.3. The prohibition is therefore phrased **descriptively**, and AC-B1.6 makes
that constraint contractual rather than incidental.

**AC-B1.4 narrowed, deliberately and on the record.** The criterion previously demanded that step
4's whole block be byte-identical. That is unsatisfiable once step 4 must carry the `no-shipment`
arm — and an unmodified block is exactly what leaves the no-shipment path unverified. AC-B1.4 is
narrowed to what it always protected: the shipment arm's lead sentence, `git show` command, both
bullets, and `STAGING_GATE_FAIL` string are preserved verbatim with unchanged semantics; the step
may gain arm labels and the second arm, and nothing else. The narrowing is recorded here rather
than applied silently, because a weakened criterion with no stated reason is indistinguishable from
an abandoned one.

**Detector deliberately not widened (rev 6).** Main-only discovery is neither §5.3 violation
construct, so catching it in the bash detector would mean opening the **closed construct set**,
adding fixtures, and re-deriving A1's 14/9/5 counts through A1→A2→A3→C2 — materially more than the
finding requires. It is rejected instead by the **existing** B1 harness function, already scoped to
`_orchestrator.agent.md` and already B1's red→green boundary. `directpush-accept-read-and-pr-forms.md`
keeps `git log origin/main..main` as an **accept** case unchanged: it proves the detector does not
flag read verbs, a property independent of what the Orchestrator uses. Recorded at §5.4 and §11 R13.

**Scope discipline (rev 6)**: no new task, file, fixture, manifest entry, harness function, CI step,
ledger entry, CODEOWNERS line, policy ID, or dependency edge. Six new acceptance criteria
(AC-B1.5–B1.9, AC-B3.5), one narrowed with rationale (AC-B1.4), two extended (AC-B1.2, AC-C2.4),
and one new risk row (R13). AC-ID parity re-swept: **46 plan IDs ↔ 46 task IDs**, exact bijection
maintained. Shipment **017-S remains one shipment** with unchanged membership (9 items) and
unchanged dependency order (A1→A2→A3→{B1,B2,B3}→C1→C2). B1 `018.007-T` is re-sized S→M and B3
`018.009-T` XS→S; both remain inside the 2-hour rule (single file, single domain, fully-specified
replacement text).

**Rev 7 — PR #54 current-HEAD review, remediation cycle 6.** Operator-authorized sixth cycle, scope
limited to threads `PRRT_kwDOTPuhps6hrC0c` (comment 3994280334, plan §6.1.2) and
`PRRT_kwDOTPuhps6hrC0o` (comment 3994280351, session memory §3).

| Thread | Finding | Resolution |
|---|---|---|
| `PRRT_kwDOTPuhps6hrC0c` | Rev 6's no-shipment arm iterated `git show origin/main:{path}` over a path set whose documented default was the **directory list** `.backlogit/ docs/plans/ docs/decisions/ docs/memory/`. `git show` exits **0** on a tree and those directories exist on `origin/main` in every repository, so the loop passed having read **no file the Stage run wrote**; paired with the degraded `stage_head_commit` default the ancestry check was trivially true too — a **false pass**. | `stage_artifact_paths` redefined as a **non-empty set of concrete repository-relative file paths**, derived from the commit when not supplied, validated for file-ness / root-containment / set-equality when supplied, and verified with `git cat-file -t` = `blob` rather than `git show`. `STAGE_ARTIFACT_ROOTS` split out as a fixed constant so step 1's pathspec cannot be shrunk by a handback. 3a re-records `stage_head_commit` after committing. AC-B1.10 and AC-B1.11 added; AC-B1.5/B1.7/B1.8 strengthened. |
| `PRRT_kwDOTPuhps6hrC0o` | The session-memory artifacts table cited plan **revision 6** and a hardening section **§10a** that no longer exists. | Corrected, and the stale-reference sweep widened file-wide: memory §3 plan and deliberation rows, plan front matter, and plan §5.3's `§10a.1` citation. Recorded in memory §14.5. |

**Rev 8 — PR #54 adversarial review.** Operator-directed adversarial round (anchor **GPT-5.6 Sol**
with **GPT-5.4-mini**, **Claude Sonnet 5**, and **Claude Opus 5**; **no route degradation**),
replacing a further self-directed cycle per the operator's standing instruction recorded at memory
§14.8. All nine findings were classified **same-contract completion** under P-021 — none reopened
the release unit's scope.

| # | Sev | Finding | Resolution |
|---|---|---|---|
| 1 | **P1** | **Aggregate verification missing.** The rev-7 contract derived its verified path set from `stage_head_commit`'s **single-commit** diff, so artifacts introduced by earlier commits on a multi-commit Stage branch were never read on `origin/main` — covered only by ancestry, the very property rev 7 judged insufficient. | `stage_base_commit` introduced, **captured by the Orchestrator immediately before invoking Stage** and retained as authoritative; base→head ancestry asserted; the set derived from the stable **two-tree** `git diff --name-only -z --diff-filter=d {base} {head} -- {STAGE_ARTIFACT_ROOTS}` (not the range form, empty by construction post-merge). Arm grows 5→**6** ordered checks. 3a **preserves** the base and recomputes the aggregate. Propagated through §6.1.1, §6.1.2, AC-B1.5/B1.8/B1.10/B1.11, §6.3/AC-B3.5, §7.2/AC-C2.4, §8, §5.0 map, §11 R14/R16, 018-F DoD, `018.007-T`, `018.009-T`, `018.011-T`. |
| 2 | **P1** | **`no-shipment` unreachable.** The installed Stage Step 5.5 (shipment assembly "MANDATORY") and Step 6 pre-summary gate both **halt** without a `shipment_id`, so no producer could ever emit `stage_outcome: no-shipment` and B1's entire second arm was dead prescription. | B3 gains **AC-B3.6**: Step 5.5's mandatory rule is scoped to a harvest that produced items, a reviewed **valid empty** harvest terminates as `no-shipment`, Step 6 requires `shipment_id` **only** for `stage_outcome: shipment`, the complete handback is emitted, and the run terminates without routing to Ship. **P-003 failures, harvest failures, and a missing required shipment after a non-empty harvest still HALT** and are never re-labelled — recorded as **R15**. `018.009-T` re-sized S→M. |
| 3 | **P1** | `018.007-T`'s description still carried rev-6 executable prose (root-default path set, `git show` loop) as operative-looking instructions, with rev 7 layered on as an addendum — two contradictory algorithms in one task body. | Description **rewritten** to a single current authoritative contract; all revision archaeology moved to implementation notes / review history. |
| 4 | **P1** | Plan §6.2 and `018.008-T` carried **two divergent five-criterion** B2 lists — symmetry in one, the P-016/P-001 clarification in the other, with AC-B2.3/B2.4/B2.5 bound to **different** criteria on each surface. | Reconciled to one identical **six**-criterion contract (AC-B2.1–AC-B2.6) carried verbatim on both surfaces; exact parity re-run. |
| 5 | P2 | Brittle line-ordinal locators (`workflow-policies.md` **L238**) for the Ship prohibition in §2, §5.3, §7.2/AC-C2.3, `018.011-T`, and the deliberation's §1.1 — the last still presented as **current evidence** despite §8.5 promising the switch to quoted anchors. | Replaced with the structural locator: P-010's **Ship MUST NOT** bullet "Commit or push directly to `main`". Deliberation §1.1's `L230`/`L258` anchors converted likewise; §8.5's promise is now kept rather than merely recorded. |
| 6 | P3 | Generated-baseline evidence cited "the `go test` step in job `expensive`", which names no step in this workflow. | Cites job `expensive` (`name: test`) step **`Test (race)`**, command `go test -race -mod=readonly ./...`, and states explicitly that AC-C2.7's constitutional `go test ./...` is **separate and still required**. **PROVENANCE SUPERSEDED AT REV 12** — naming the step precisely was correct and stands, but calling that command the *generated baseline* was not: the template emits `{{TEST_COMMAND}}` and the recorded render input is `go test ./...`, so it is the **observed installed** command and the guarantee is scoped to the full-suite `./...` pattern (§5.0 "Invocation provenance", AC-C1.7, R23). Per-task `harness_cmd` values stay narrow and unchanged. |
| 7 | P3 | AC-A1.2 named the fixture prefixes `reject-*` / `accept-*`; the actual files are `directpush-reject-*` / `directpush-accept-*`. | Corrected on both surfaces. |
| 8 | P3 | AC-A2.1 demanded the **bare** scan exit non-zero, but A2's stub classifies every construct `accept`, so a bare scan correctly exits **0** — the criterion was unsatisfiable by a correct A2. | Non-zero scoped to **`--self-test` only**; bare-mode red is reserved for **AC-A3.3**, after a real detector exists. |
| 9 | P3 | Stale references: Constitution Check row XI cited AC-C2.5 (markdownlint) instead of AC-C2.6; §12's rev-7 row promised "(see below)" with no rev-7 block; the deliberation header claimed the planning branch had "no push, no PR". | All corrected; the rev-7 record block above is added, and the deliberation header now names its five addenda and the PR the branch actually carries. |

**Scope discipline (rev 8)**: no new task, file, fixture, manifest entry, harness function, CI step,
ledger entry, CODEOWNERS line, policy ID, or dependency edge. Two new acceptance criteria
(**AC-B2.6**, **AC-B3.6**), six strengthened (AC-B1.5, AC-B1.8, AC-B1.10, AC-B1.11, AC-B3.5,
AC-C2.4), four corrected in place (AC-A1.2, AC-A2.1, AC-C2.3, AC-C2.8), one renumbered set
(AC-B2.1–AC-B2.5 → the reconciled AC-B2.1–AC-B2.6), and two new risk rows (**R15**, **R16**).
Shipment **017-S remains one shipment** with unchanged membership (9 items) and unchanged
dependency order (A1→A2→A3→{B1,B2,B3}→C1→C2). B3 `018.009-T` is re-sized S→M for the two
additional `_stage.agent.md` edit sites; it remains inside the 2-hour rule (single file, single
domain, fully-specified replacement text). No production code, test, script, workflow, policy, or
agent file is modified by this pass — those remain Ship's execution surfaces.

**Rev 9 — PR #54 current-HEAD review, remediation cycle 7.** One **visible** Copilot finding
(thread `PRRT_kwDOTPuhps6hrbAk`, comment `3994429008`, on `018.009-T:22`), classified
**same-contract completion** under P-021. Scope limited to that finding; no other scope reopened.

| Thread | Finding | Resolution |
|---|---|---|
| `PRRT_kwDOTPuhps6hrbAk` | `018.009-T` updates Stage's **authorization** and declares a **handback**, but adds no operative Stage step to create/check out the dedicated branch or commit the artifacts. The installed `_stage.agent.md` has only planning, harvest, shipment, archive, and summary steps — no branch or commit operation anywhere — so implementing AC-B3.1–AC-B3.6 as written can still leave Stage on the default branch, making the reported `stage_head_commit` and concrete-path contract unverifiable. | **Two operative steps added to the file B3 already owns** (§6.3.1): **Step 1.9**, a pre-mutation branch gate between the learnings-retrieval and deliberation steps, which derives a `{scope-slug}`, records `stage_base_commit`, **verifies** an Orchestrator-supplied branch or **creates/checks out** `chore/stage-{scope-slug}`, refuses to write while `HEAD` is the default branch, and switches the single worktree (P-016); and **Step 5.7**, a post-mutation artifact commit between stash archival and the summary, which asserts `HEAD`, stages only the four `STAGE_ARTIFACT_ROOTS`, commits conventionally, sets `stage_head_commit` from the resulting `HEAD`, and neither pushes nor touches a PR. Both are registered in the Step Sequence Contract checklist and enforced by Step 6's pre-summary gate. **AC-B3.7** added, constraining **order** as well as presence; §6.3.2 specifies the structural (index-comparison) detection inside the existing `TestDirectPushGate_StageAgentSurfaceClean`. |

**The naming dependency the finding exposed second-order.** Rev 1–8 named the branch
`chore/stage-{shipment_id}` with `chore/stage-{chore-slug}` as a shipment-less fallback. Once the
branch must exist *before* the first artifact write, that primary form is **underivable**: Step 5.5
creates the shipment as the session's last mutation. The rev-8 wording therefore pushed an executor
toward deferring the branch past the very writes it protects — a second route back to the
`fdff9e4` shape. Rev 9 makes the form `chore/stage-{scope-slug}`, keeps a shipment ID as a
**permitted** slug rather than deleting it, and corrects §6.2's grant and route table, §6.3's Git
row, AC-B2.2, and AC-B3.2 together. `chore/stage-pipeline-policy-gap` — the branch carrying this
plan — is a conforming example.

**Producer/consumer failure class, fourth instance.** Rev 1 obliged Stage to merge a PR it could
not create; rev 6 obliged the Orchestrator to find a branch nothing announced; rev 8 keyed a
verification on an outcome nothing could emit; rev 9 finds a **reported value with no producing
action**. The discipline deliberation §12.2 wrote down — *for every value a contract consumes, name
the surface that emits it, and confirm that surface is permitted to* — was followed for
permission and emission and **not** for execution. It is extended in deliberation §13: confirm the
producing surface is not merely permitted but **instructed**, and that the instruction is **ordered**
relative to the state it describes.

**Detector deliberately not widened (rev 9).** A missing operative step is neither §5.3 construct
— it is not a push to the default branch and not a permission to commit on one — so the closed
construct set and the 14-fixture corpus stay exactly as they are, the §10.4 / §11.4 / §12.4
position unchanged. Rejection is carried by the **existing** B3 harness function, already scoped to
`_stage.agent.md` and already B3's red→green boundary. The prescribed replacement text was checked
against the detector rather than assumed safe (§6.3.1, detector-neutrality).

**Deliberately not swept (rev 9)**: the `chore/stage-{shipment_id}` form still appears in this
plan's **historical revision narratives** (the Revision 6 paragraph, §6.1.1's quotation of the 3e
text being deleted, §12's rev-6 record), in the deliberation's archaeology, and in `018.007-T`'s
rev-6 reconciliation note. Those are accurate records of what earlier revisions said and of the
text B1 removes; no acceptance criterion depends on them. Every **operative** occurrence — the
P-010 grant, the route table, the Role Boundary cell, 3a's create-arm, AC-B2.2, AC-B3.2 — is
corrected.

**One defect found by internal review before commit (rev 9).** The first draft placed Step 1.9 at
a *floating* position — "earliest point a slug is derivable, latest before the first write,
whichever comes first" — with a provisional-slug escape hatch for an early Step 1 duplicate
archival. That is **self-contradictory** against the very file it edits: `_stage.agent.md`'s Step
Sequence Contract states that a session MUST execute its steps **in order**, so a step registered
between Step 1.8 and Step 2 cannot also be specified to run during Step 1. Worse, the floating form
made the position unassertable — AC-B3.7 would have enforced the ordinal while the prose licensed an
earlier run — and it left a tracked `.backlogit/` write legal on the default branch. The fix
inverts it: the ordinal is **fixed**, and the two mutations that would otherwise precede it are
**deferred** past it by an explicit rule. Recorded rather than silently repaired, per the rev-6
precedent: the defect class — *a guard whose stated position contradicts the ordering semantics of
the contract it is inserted into* — is worth recognizing again.

**Scope discipline (rev 9)**: no new task, file, fixture, manifest entry, harness function, CI
step, ledger entry, CODEOWNERS line, policy ID, shipment, or dependency edge. One new acceptance
criterion (**AC-B3.7**), three strengthened (AC-B2.2, AC-B3.2, AC-B3.5), two re-scoped
(AC-B3.3, AC-B3.6 — their "only changes outside the Role Boundary table" clause now admits
AC-B3.7), one extended (AC-C2.4), and two new risk rows (**R17**, **R18**). AC-ID parity re-swept:
**51 plan IDs ↔ 51 task IDs**, exact bijection maintained. Shipment **017-S remains one shipment**
with unchanged membership (9 items) and unchanged dependency order
(A1→A2→A3→{B1,B2,B3}→C1→C2). B3 `018.009-T` is re-sized **M→L** for the four additional
`_stage.agent.md` edit sites; it remains inside the 2-hour rule (single file, single domain, fully
specified replacement text). No production code, test, script, workflow, policy, or agent file is
modified by this pass.

**Rev 10 — PR #54 second adversarial review.** Operator-directed second adversarial round (anchor
**GPT-5.6 Sol** with **GPT-5.4-mini**, **Claude Sonnet 5**, and **Claude Opus 5**; **no route
degradation**). Four **visible** Copilot threads plus one adversarial-only finding, all classified
**same-contract completion** under P-021. Scope limited to those five findings; no other scope
reopened.

| # | Sev | Thread / comment | Finding | Resolution |
|---|---|---|---|---|
| 1 | **P1** | `PRRT_kwDOTPuhps6hsPRZ` / `3994747640` | **The handback outcome pair failed open.** `stage_outcome` was derived *unconditionally* from whether a `shipment_id` was returned, with no `absent` branch — so a reported `shipment` outcome whose `shipment_id` did not reach the Orchestrator was silently rewritten to `no-shipment`, the shipment arm was skipped, the no-shipment arm passed on artifacts that really were on `origin/main`, the gate reported success, and Ship was never routed. A formed shipment, silently abandoned. | §6.1.2 **correction 6**: the derivation is scoped to the **absent** case and a reported value is preserved; a **fail-closed pair validation** is added to step 4 *before* arm selection (unknown outcome halts; `shipment` without a non-empty `shipment_id` halts; `no-shipment` with one halts). Placed in **verification**, not defaulting, so AC-B1.5's never-halts resolution invariant is intact. **AC-B1.12** added, AC-B1.5 strengthened; producer side carried by AC-B3.5 (explicit, consistent pair); coupling by AC-C2.4 rev-10 (a); DoD bullet added. |
| 2 | **P1** | `PRRT_kwDOTPuhps6hsPRq` / `3994747663` and `PRRT_kwDOTPuhps6hsPR5` / `3994747682` | **The pre-gate deferral rule was under-inclusive.** Rev 9 *enumerated* two deferred mutations. The installed Stage contract also schedules, before Step 1.9's ordinal, a **stash-classification** checkpoint, a **contextual-grouping / operator-selection** checkpoint, and the **write half** of the mandatory late-identifier reconciliation — all tracked, all therefore still legal on the default branch under the rev-9 wording. | §6.3.1 correction 1 rewritten: the rule is **categorical** — Step 1.9 precedes **every** tracked Stage artifact mutation — with the three classes enumerated as illustration, not as a closed list. Read-only classification and grouping still run first (they derive the slug). Gitignored disposables (index sync, `.backlogit/checkpoints/`, hook queue, `.backlogit/runtime/`) are expressly **excluded**, not overclaimed. Lifted into **AC-B3.7**, §6.3.2 assertion 3, `018.009-T`, the DoD, and AC-C2.4 rev-10 (b); R18 updated. |
| 3 | **P1** | `PRRT_kwDOTPuhps6hsPSJ` / `3994747702` | **Out-of-root detection was unimplementable.** Step 5.7 required a halt on any change outside `STAGE_ARTIFACT_ROOTS` but prescribed only a root-scoped `git add`, which **ignores** such a change rather than reporting it. The obligation had no mechanism. | Step 5.7 gains an explicit **detection sub-step before staging**: `git status --porcelain=v1 -z --untracked-files=all`, NUL-record parsing (never line splitting), both rename/copy endpoints treated as affected paths, halt with a P-010 signal on any out-of-root path. `--untracked-files=all` and `-z` are justified as load-bearing with **verified** output shape. Pre-existing unrelated dirt is **halt-and-leave-alone** — no checkout, restore, stash, revert, or delete (Constitution VII). Lifted into **AC-B3.7** and §6.3.2 assertions 3–4 (including the detection-before-staging **intra-section ordering** check); **R20** added. |
| 4 | **P2** | — (AC-A2.5) | **AC-A2.5 could not prove its own claim.** One driver-scoped `-run` invocation cannot establish that a different test is red. | Split into **two exactly-anchored invocations**: the driver test exits 0; `TestDirectPushGate_SelfTestPasses` exits non-zero **and** prints `--- FAIL:` plus the expected **detector-stub** reason, with compile errors, absent `bash`, missing fixtures, and a mis-typed selector rejected by name. The `--- FAIL:` guard is required because — **verified** — a non-matching `-run` pattern exits **0** with `[no tests to run]`. Bare-mode red stays **AC-A3.3**'s; P-004 is not weakened. §5.0 map and `018.005-T` synchronized. |
| 5 | **P1** | adversarial-only (low confidence, directly evidenced) | **The harness lifecycle deadlocked against Ship's full-suite gate.** §5.0 declared all eight per-task functions red after H0 and left them red until their task landed, while `_ship.agent.md` Step 4.3 runs `go test ./...` after **every** task. After A1 went green the seven future functions were still red, Step 4.3 failed, and no task could ever complete. | New **§5.0.1**: one selector flag (`-gatetask`, dot-free and shell-neutral — **verified**, a dotted name is split by PowerShell's tokenizer) plus one eight-row activation table. Targeted mode forces exactly one task's assertions (each `harness_cmd` carries it, so every task keeps an observed red→green transition); default mode runs surfaces that exist plus the **activation-independent A1 anchor**, which keeps H0's `go test ./...` red literal. Four mutation-proof checks (MP0 table integrity, MP1 executed predicate non-vacuity, MP2 terminal completeness inside C2, MP3 visible skips + strict selector) close the skip-forever degeneration; **R19** records the residual. Four alternatives rejected on the record, including amending Ship Step 4.3 (out of scope, generated file, weakens test-first for every future shipment). **SUPERSEDED AT REV 11** — this resolution's surface-predicate table was itself defective on two counts (self-disarming C1/C2; a red phase that did not meet the installed P-004 postcondition) and is replaced in full by the harness-state manifest. Retained here as the historical record of what rev 10 decided, **not** as an operative contract; §5.0.1 is authoritative. |

**Harness-ready semantics kept honest (rev 10).** `harness-ready` continues to mean exactly what
`harness-architect` verifies — compilation clean, default suite red — and §5.0 now says so
explicitly instead of letting the label imply that all eight assertions were simultaneously red.
The per-task red is a **per-task** obligation discharged by Ship at claim time from that task's
`harness_cmd`, which is the evidence `build-feature` already records. Nothing is asserted that no
actor produces, and no label is applied from prose — the rev-4 lesson, held.
**Corrected at rev 11**: the caveat in the middle sentence was the tell. `harness-architect`'s
installed postcondition *is* "all generated tests failing expected not-implemented markers", so
declining to claim simultaneity was not a scrupulous clarification — it was an admission that the
postcondition was unmet, phrased as a virtue. At rev 11 the eight reds **are** simultaneous in the
`red` phase, so the label claims exactly what the evidence shows and the caveat is withdrawn.

**The producer/consumer failure class, fifth instance — and a new one alongside it.** Rev 10's
finding 1 is the class again, in its most dangerous form yet: the consumer did not merely fail to
find a value, it **overwrote a correct one with a derived default whose fallback was a success
terminal**. The extended discipline gains a sixth clause: *a default must never be able to
overwrite a reported value, and a derivation whose fallback is a success terminal must be
validated before it is acted on.* Findings 2, 3 and 5 are a **different** class — an obligation
written without a mechanism that can discharge it (an enumeration mistaken for a rule, a halt with
no detector, a red phase that cannot coexist with the gate that consumes it). Its discipline: *for
every MUST, name the operation that performs it and confirm that operation can observe what the
MUST is about.*

**One latent defect found by internal review before commit (rev 10).** Requirement (d) —
*the final `go test ./...` runs all eight real assertions green* — forced a check revisions 4–9
never made: is every function's assertion still true at the **end state**? Seven were. **A2's was
not.** Its map row asserted that `--self-test` reports every reject fixture failing **against the
all-accept stub** with a non-zero exit — true only while A2's stub is current, **false the moment
A3 installs the real detector**. The function would have gone red at A3 and stayed red, wedging
`go test ./...` for the rest of the shipment. Activation did not cause this; it made it visible,
because "all eight green at the end" had never been stated as a requirement before. The fix asserts
the **durable** driver contract A2 actually delivers and leaves the stub-specific observations in
**AC-A2.1**, read from the script's own output at A2's boundary — the criterion unchanged and
unweakened, only its carrier named. Recorded rather than silently repaired, per the rev-6 and rev-9
precedent: the defect class — *a regression assertion pinned to a transitional state* — is worth
recognizing again.

**Scope discipline (rev 10)**: no new task, file, fixture, manifest entry, harness function, CI
step, ledger entry, CODEOWNERS line, policy ID, shipment, or dependency edge. One new acceptance
criterion (**AC-B1.12**), four strengthened (AC-A2.5, AC-B1.5, AC-B3.5, AC-B3.7), one extended
(AC-C2.4), one new plan subsection (**§5.0.1**), one new B1 correction (**correction 6**), and two
new risk rows (**R19**, **R20**). All eight `harness_cmd` values gain the `-gatetask={ID}` selector;
no harness **function** is added or renamed. AC-ID parity re-swept: **52 plan IDs ↔ 52 task IDs**,
exact bijection maintained. Shipment **017-S remains one shipment** with unchanged membership
(9 items) and unchanged dependency order (A1→A2→A3→{B1,B2,B3}→C1→C2). No task is re-sized: rev 10
reworks clauses inside step sections rev 9 already specified rather than adding edit sites. No
production code, test, script, workflow, policy, or agent file is modified by this pass — those
remain Ship's execution surfaces.

**Deliberately not claimed (rev 10).** Stage does **not** push, open, update, comment on, or merge
PR #54, and does **not** reply to or resolve threads `PRRT_kwDOTPuhps6hsPRZ`,
`PRRT_kwDOTPuhps6hsPRq`, `PRRT_kwDOTPuhps6hsPR5`, or `PRRT_kwDOTPuhps6hsPSJ`. Those are Orchestrator
or operator actions; this record states only what the artifacts now contain.

**Rev 11 — PR #54 third adversarial review.** Operator-directed third adversarial round (anchor
**GPT-5.6 Sol** with **GPT-5.4-mini**, **Claude Sonnet 5**, and **Claude Opus 5**; **4/4 usable**,
**no route degradation**). The panel **unanimously** confirmed two **P0 direct-completion blockers**
from Copilot review `5185014458` plus one stale cross-reference. All three are **same-contract
completion** under P-021 — each corrects a mechanism this plan already owns, and none reopens scope.

| # | Sev | Thread / comment | Finding | Resolution |
|---|---|---|---|---|
| 1 | **P0** | `PRRT_kwDOTPuhps6hstWa` / `3994928139` | **Self-disarming C1/C2 activation.** Rev 10's §5.0.1 activated C1 in default mode iff `jobs.lint.steps[]` contained the gate step — *the exact proposition `TestDirectPushGate_CIWiringIsBlocking` exists to assert* — and gave C2 **literally C1's predicate**. A regeneration that dropped the step therefore made both functions `t.Skip` rather than fail, MP2 (which lives inside C2) never executed, and `go test ./...` stayed green. The one control R5 rests on disarmed itself precisely when it was needed. | §5.0.1 **rewritten**: the predicate→activation edge is removed entirely. Activation derives **only** from a tracked harness-state manifest and the `-gatetask` selector, so no assertion's activation can be switched off by mutating its own subject. Rejected alternative (6) — exempting only C1/C2 — treats the instance, not the class. **AC-C1.6** extended (a skip with the step absent is an explicit failure of the criterion); **AC-C2.9** added; **AC-C2.8** gains the activation-independence clause it had silently assumed; §5.0 map rows for C1/C2 and the R5 row corrected. |
| 2 | **P0** | `PRRT_kwDOTPuhps6hstWg` / `3994928153` | **H0 violated the installed P-002/P-004/`harness-architect` contract.** That contract requires `go test ./...` red with **all** generated tests failing expected not-implemented markers before any task becomes `harness-ready`. Rev 10 delivered **one** failing function plus **seven skips**. Reverting to rev 9's simultaneous all-eight-red would satisfy P-004 but restore the Step 4.3 deadlock rev 10 existed to remove. | The two requirements are only irreconcilable while activation is a **static predicate set**; they reconcile once it is a **phase**. H0 writes `tests/integration/testdata/directpush-gate/harness-state.json` = `{"phase":"red","completed":[],"terminal":false}`. In `red`, **all eight** activate and fail, each with its own task-specific marker — the literal 8/8, and **only then** may the tasks be labelled `harness-ready`. A1's completion performs the one-way `red`→`build` transition, after which the default suite runs the completed set and skips only not-yet-started tasks, so **Step 4.3 is green after every task**. At `terminal` all eight run again with **no skip anywhere**. §5.0, §5.0.1, §10.1 and **AC-A1.4** carry it. |
| 3 | **P2** | `PRRT_kwDOTPuhps6hstWn` / `3994928163` | **Stale deleted-`3e` reference.** `018-F`'s DoD still named `_orchestrator.agent.md` Step 1.5 **sub-step 3e** as one of the four live authorization surfaces, although B1 **deletes** 3e. The DoD asserted agreement across a surface that will not exist at the end state. | The bullet now names the **live** pair — the Step 1.5 **preamble** (reworded to the verification/postcondition form) and **3a** (which absorbs 3e's branch-point instruction) — and adds an explicit **3e-absence** clause, so the DoD asserts the corrected contract rather than a deleted one. The plan's §2/§3/§6.1.1 narrative quotations of 3e are **left intact**: they are historical records of the defect and sit outside the scanned corpus (§5.2). |

**The self-referential-guard class, first instance — and the discipline it adds.** Finding 1 is a
class this plan had not yet met: *a control whose activation condition is the proposition it
verifies*. It passes every test that asks "does the assertion exist?", "is it structural, not a text
search?", and "is it non-vacuous when the surface is present?" — AC-C1.6 asked all three and all
three held. What it fails is the only question that matters for a regression guard: *does it still
run when the thing it guards is gone?* The discipline: **for every guard, state what activates it,
and confirm that the activator is disjoint from the subject.** Note that this is the mirror image of
the producer/consumer class that findings 1–5 of rev 10 kept producing — there the consumer trusted
a value the producer might not send; here the detector trusted a condition the defect itself
removes.

**The second class, recurring: an accommodation that quietly redefines the contract.** Finding 2 is
rev 10's own remedy overshooting. Faced with a real deadlock, rev 10 weakened the **red phase** —
the thing P-004 actually specifies — rather than the **activation lifecycle**, which nothing
specifies. It even recorded the weakening honestly ("it does **not** claim that all eight assertions
were simultaneously red"), which is how it survived a review round: the caveat made the divergence
look like a documented decision instead of a contract violation. The discipline: **when a
constraint and a gate conflict, first check whether the conflict is between them or between the gate
and an implementation choice you are free to change.** Here it was the latter — introducing a phase
cost one state file and preserved both contracts exactly.

**Foundational contracts held (rev 11).** This revision plans **no** modification to
`.github/skills/harness-architect/SKILL.md`, to P-002 or P-004 in `workflow-policies.md`, or to
`_ship.agent.md` Step 2 or Step 4.3 — all four remain in §11's out-of-scope list, unamended. §5.0.1
carries the clause-by-clause argument that the release-specific manifest satisfies each of them
**literally**: `harness-architect` already emits a non-Go H0 deliverable so a write-once `testdata/`
constant is strictly smaller; P-004's red phase is met as the literal 8/8; Step 2's label partition
is untouched; and Step 4.3 keeps its **full** `go test ./...` scope, with rev 10's rejected proposal
to narrow it still rejected.

**Scope discipline (rev 11)**: no new task, sub-epic, fixture, harness **function**, CI step, ledger
entry, CODEOWNERS line, policy ID, shipment, or dependency edge. **One** new file enters the plan —
the harness-state manifest, an H0 deliverable, not a task deliverable. Two new acceptance criteria
(**AC-A1.4**, **AC-C2.9**), three extended (AC-C1.6, AC-C2.2, AC-C2.8), one clarified (AC-C2.7), one
plan subsection rewritten (**§5.0.1**), one risk row rewritten (**R19**) and one added (**R21**), and
R5/R12 cross-references corrected. AC-ID parity re-swept: **54 plan IDs ↔ 54 task IDs**, exact
bijection maintained. Shipment **017-S remains one shipment** with unchanged membership (9 items)
and unchanged dependency order (A1→A2→A3→{B1,B2,B3}→C1→C2 — the `blocks` edges were read back from
`.backlogit/` and confirmed to make **B1/B2/B3 parallel**, which is why `completed` is validated as a
dependency-closed **subsequence** and not a prefix). **No task is re-sized**: each task gains a
one-line insertion into a shared state file in a domain it already touches, which changes no size
band and breaches no 2-hour boundary. No production code, test, script, workflow, policy, or agent
file is modified by this pass — those remain Ship's execution surfaces.

**Deliberately not claimed (rev 11).** Stage does **not** push, open, update, comment on, or merge
PR #54, and does **not** reply to or resolve threads `PRRT_kwDOTPuhps6hstWa`,
`PRRT_kwDOTPuhps6hstWg`, or `PRRT_kwDOTPuhps6hstWn` (comments `3994928139`, `3994928153`,
`3994928163`). No comment on review `5185014458` has been replied to or resolved by this pass. Those
are Orchestrator or operator actions; this record states only what the artifacts now contain.

**Rev 12 — PR #54 post-remediation adversarial re-review.** Operator-directed re-review of the rev-11
remediation (anchor **GPT-5.6 Sol** with **GPT-5.4-mini**, **Claude Sonnet 5**, and **Claude Opus 5**;
**4/4 usable**, **no route degradation**). Three **P1 residuals** in the rev-11 mechanism plus a set
of direct consistency defects, all **same-contract completion** under P-021 — each corrects a
mechanism this plan already owns, and none reopens scope.

| # | Sev | Finding | Resolution |
|---|---|---|---|
| 1 | **P1** | **MP5 deadlocked a legitimate post-C1/pre-C2 state.** Rev 11 removed surface predicates from *activation* but **retained them as failure probes** in MP5, which failed whenever a task outside `completed` had its surface present. C2's probe was C1's old predicate — the job-`lint` step — which **C1 itself creates**. So after C1 completed and before C2 started, C2's probe read `true`, MP5 fired inside A1, and Ship's **default full suite went red**: the exact Step 4.3 deadlock revisions 10 and 11 exist to remove, reintroduced by the guard meant to protect the mechanism. | **MP5 RETIRED** — removed, not renumbered (renumbering five checks across six artifacts is churn and rewrites reviewed identifiers). The prohibited edge is restated in its general form (§5.0): **neither activation nor gate-level failure may read a surface the harness asserts on**. Nothing is lost: a **completed** task's rollback is caught by its **own substantive assertion**, active in default mode from its completion onward; an **incomplete** task's prerequisites or surface fragments being present is a **legal** `build` state that must not fail anything. The stale "MP1 probe non-vacuity" attribution is withdrawn with it — MP1 is the manifest validator and has no probes. §5.0 map row, §5.4 AC-A1.4, `018.004-T`, R19, the DoD and the memory record are all corrected. |
| 2 | **P1** | **A consistent terminal-manifest rollback could skip C2 before MP2.** MP2 lives *inside* C2, and C2's default-mode activation was read from the **same manifest** MP2 validates. A coordinated rewind of `phase`, `completed`, and `terminal` **together** left a self-consistent `build` state in which MP1 passed, C2 **skipped as not-yet-started**, and MP2 never fired — the terminal claim withdrawn with no check firing. | One small, tracked, non-generated **terminal completion witness** at `tests/integration/testdata/directpush-gate/terminal-witness.json`, **independent of the main manifest**, with a fixed schema/version, the contract identifier, the exact terminal set, and a digest over it — **no timestamps, no environment-specific data**. H0 requires it **absent**; legitimate pre-C2 build requires it **absent**. C2's completion writes **the witness first**, then **atomically replaces** the manifest with the exact terminal state; green requires the two to **agree**. New **MP6** runs in `gate()` **before any activation decision**: a present witness forces **C2 active** and requires a terminal manifest; a terminal manifest requires a valid witness. Manifest-only and witness-only rollbacks therefore **fail loudly**. **AC-C2.10** added; AC-C2.2 extended to ten mutation cases; **R22** records the honest residual — rolling back **both** in one reviewed diff resembles pre-C2 state and is caught by **diff/review, not runtime history**. No claim is made that a stateless validator detects every coordinated rollback. |
| 3 | **P1** | **CI regeneration provenance overstated.** Revisions 5–11 cited `go test -race -mod=readonly ./...` as the **generated baseline**. The installed `.github/workflows/ci.yml` does run that command in `expensive` / `Test (race)`, but the generation template emits `{{TEST_COMMAND}}` and this workspace's recorded profile input is `go test ./...` — so the flags are **local to the installed artifact**, not render-guaranteed. | The **observed installed command** and the **generated baseline inputs** are now distinguished explicitly (§5.0 "Invocation provenance", with the three-row evidence table). The regeneration-resistant guarantee is restated at the strength it has: it requires **a full-suite invocation including `./...`** that runs the integration harness, **not** a permanent promise that the exact race/`-mod` flags survive a render. **AC-C1.7** adds mandatory **post-render verification of the rendered full-suite command**, written into the LOCAL DIVERGENCE ledger; **R23** records the procedural residual. Plan, `018.010-T`, `018.011-T`, the feature DoD, R5/R11/R13, §10 and the evidence obligations in AC-C1.3/AC-C2.8 are all corrected. The installed race command is **no longer claimed as template baseline** anywhere. |

**Direct consistency fixes (rev 12).** Eight, each a case where the artifacts disagreed with
themselves rather than with reality:

1. **Memory current-artifact index** named plan **revision 10** and deliberation addenda through
   §14 — corrected to **revision 12** and through **§16**.
2. **Decision live addenda index** stopped at §14 — now carries **§15 (rev 11)** and
   **§16 (rev 12)**.
3. **MP5 probe language** removed everywhere it appeared, and the **stale A1 "MP1 probe
   non-vacuity" attribution** withdrawn (MP1 has no probes; the phrase described MP5's check).
4. **Valid manifest states defined exhaustively** — `red` ⇒ `completed` empty ∧ `terminal: false`;
   `build` ⇒ non-empty dependency-closed **proper subset** ∧ `terminal: false`; `terminal` ⇒ the
   exact eight-task set ∧ `terminal: true` ∧ a **valid witness**. `red` + `terminal`, `build` with
   all eight, `terminal` with a subset, and every other combination are **rejected** by name.
5. **"Five checks" miscount** corrected — the list enumerated six. It is now **six active checks**
   (**MP0, MP1, MP2, MP3, MP4, MP6**) with **MP5 RETIRED**, and the residual
   **`completed`-prefix** wording is replaced with the accurate **dependency-closed
   order-preserving subsequence** (a **proper subset** in `build`).
6. **Zero-skips qualified** as **default-mode** `red`/`terminal` behaviour. Targeted mode skips the
   seven non-selected functions **by design**; its only obligation is that the **selected** test
   executes.
7. **Dependency direction clarified** in the gate dependency map: backlogit stores
   `dep add <item> <depends-on> --type blocks`, so the recorded edge reads **dependent →
   prerequisite** despite the `blocks` label, and `gateDeps[X]` is **X's prerequisites**. Reading
   the label the other way inverts the graph and would make the closure check accept exactly the
   sets it must reject.
8. **AC-C2.4** no longer points at the historical **§2** four-surface enumeration, one of whose
   entries (**3e**) B1 **deletes**. It now names the **live post-change** set — the Step 1.5
   **preamble** and **3a**, plus an explicit **3e-absence** clause — together with the
   **Stage/P-010** surfaces.

**Scope discipline (rev 12)**: no new task, sub-epic, fixture, harness **function**, CI step, ledger
entry, CODEOWNERS line, policy ID, shipment, or dependency edge. **One** new file enters the plan —
the terminal completion witness, written by C2's completion, not a task deliverable of its own.
Three new acceptance criteria (**AC-C1.7**, **AC-C2.10**, **AC-C2.11**), five corrected or extended
(AC-A1.4, AC-C1.3, AC-C2.2, AC-C2.4, AC-C2.8), two qualified (AC-C2.7, AC-C2.9), one check
**retired** (**MP5**) and one added (**MP6**), one new plan subsection (**§5.0.2**), one risk row
rewritten (**R19**) and two added (**R22**, **R23**). AC-ID parity re-swept: **57 plan IDs ↔ 57 task
IDs**, exact bijection maintained. Shipment **017-S remains one shipment** with unchanged membership
(9 items) and unchanged dependency order (A1→A2→A3→{B1,B2,B3}→C1→C2). **No task is re-sized**: the
witness is one additional write inside C2's existing completion transaction, in a directory and a
domain C2 already touches. No production code, test, script, workflow, policy, or agent file is
modified by this pass — those remain Ship's execution surfaces.

**Deliberately not claimed (rev 12).** Stage does **not** push, open, update, comment on, or merge
PR #54, and does **not** reply to or resolve any thread or comment on it — specifically **not**
comments `3994928139`, `3994928153`, or `3994928163`, which remain **unreplied and unresolved** by
this pass. Reply text for them is prepared for the operator or Orchestrator to post; preparing it is
not posting it. Those are Orchestrator or operator actions; this record states only what the
artifacts now contain.

### 12.13 Revision 13 — final capped post-remediation remediation (operator-authorized)

**Authorization (explicit, and recorded because the budget was already exhausted).** The rev-12
session terminated **BLOCKED** with the adversarial cycle cap reached at `cycles_run: 2` and six
uncapped residuals outstanding (memory §20). Per that record's §20.5, the operator chose option 1
and **explicitly authorized ONE additional Stage remediation plus adversarial re-review cycle**.
This revision is that single authorized pass. The authorization is **consumed** by it; no further
remediation cycle is implied, and the re-review it enables has **not** been performed by Stage.

**The six capped residuals, and their disposition.** All six were classified
`post_remediation_residual`; none is P0; all are **same-contract completion** under P-021.

| # | Conf. | Sev. | Residual | Disposition |
|---|---|---|---|---|
| 1 | MEDIUM | P1 | **Targeted selector precedence vs. terminal witness activation.** The rev-12 activation table placed the witness rows before the `-gatetask` rows and said a valid witness "forces `C2` active; all eight activate" with no mode qualifier. Read sequentially, a **targeted** run in terminal state activated all eight bodies, contradicting MP3, AC-C2.7, AC-C2.9, and AC-C2.11. | **FIXED.** `gate()` is split into **Stage 1 — mode-independent integrity** (MP1 + MP6 validation, fails in every mode) and **Stage 2 — mode-dependent activation**, in which a valid `-gatetask` has **activation precedence** over every default-mode rule including MP6's forcing clause. Witness integrity may still **fail** a targeted invocation; it may no longer **activate** unrelated bodies. Default mode is unchanged: red/terminal activate all eight; build activates the completed set; a valid witness forces `C2` active. Synchronized across §5.0.1 (both tables, MP3, MP6, zero-skip paragraph, preservation row (f)), §5.0.2 (points 6, 7, 7a), §7.2 (AC-C2.7, AC-C2.9, AC-C2.10, AC-C2.11), §8, §11 R19, `018-F`, all eight task files, the deliberation, and the memory record |
| 2 | LOW | P1 | **H0 harness eligibility with dependency-blocked tasks** — could `harness-architect` legitimately batch seven tasks with unmet dependencies? | **RESOLVED AGAINST THE INSTALLED CONTRACT, WHICH IS UNCHANGED.** §5.0 "H0 eligibility" cites `_ship.agent.md` **Step 2** (runs "once, up front — not in a loop"; lists **`queued`** tasks; partitions **only** on the `harness-ready` label; halts unless **every** queued task carries it) and **Step 3** (runs after, and is where "sort the queue by dependency order" first appears), plus `harness-architect` Step 1's exclusion of **`blocked`** as a *status value* — and all eight `017-S` tasks are `queued`. Dependencies gate **claim/execution**, not harness generation. A mechanical **set-equality** check is added to AC-C2.11 evidence point 1. **No permission was invented**: §5.0 states that if the installed contract is ever read to exclude such tasks, the correct response is to **report the contradiction** with Step 2 item 4, not to relax the 8/8 red requirement |
| 3 | LOW | P1 | **Control-state input exemption wording** — the independence invariant, read literally, forbade MP1/MP6 from reading the manifest and witness. | **FIXED, NARROWLY.** §5.0 "Control-state inputs are expressly exempt" scopes the invariant to **substantive surfaces** (agent/policy/workflow/script behaviour) and expressly **allows and requires** the validated manifest and terminal witness as MP1/MP6 inputs. A four-row table gives the discriminator, and a three-part test (authored by the harness for the harness; never a subject of a substantive assertion; covered by its own integrity proof) bounds the exemption so it cannot generalize into a loophole |
| 4 | LOW | P1 | **Plan/task AC-C1 parity** — §7.1 and `018.010-T` carried materially different AC-C1 text (ID parity only). | **FIXED.** One canonical AC-C1.1–AC-C1.7 block is now written **byte-identically** into both, matching the §7.2 ↔ `018.011-T` convention. Verified by **content** parity, not ID parity: both blocks are **7,197 bytes**, and §7.2 ↔ `018.011-T` is re-verified at **19,273 bytes** |
| 5 | LOW | P1 | **Vacuous terminal lint-removal evidence** — §5.0.2 point 7 used `-gatetask=C1`, which activates C1 **by construction**, so it could not prove terminal default-mode activation and would have passed under the rev-10 self-disarming design; it also never exercised C2. | **FIXED.** Point 7 is rewritten as a **default-mode** (unselected) run against a **copied terminal-state repository fixture** under `t.TempDir()` — real tree never mutated, `git status --porcelain` empty, fixture manifest and witness **valid and in agreement** — asserting **by test-event name** that `TestDirectPushGate_CIWiringIsBlocking` **fails for the missing lint step** and `TestDirectPushGate_FullCorpusClean` **executes** (never `--- SKIP:`). The targeted command is retained **separately** as point **7a**, labelled selector evidence only. AC-C1.6 and AC-C2.11 carry the same requirement; §8 carries both commands, distinctly labelled |
| 6 | LOW | P2 | **Residual CI provenance overstatement** — §6.3.2's "Regeneration exposure" still called the installed `Test (race)` command the generated baseline. | **FIXED.** That paragraph now carries the same exact distinction used everywhere else: **observed installed command** `go test -race -mod=readonly ./...`; **template** emits `{{TEST_COMMAND}}`; **recorded render input** `go test ./...`; **invariant** is a rendered **full-suite `./...`** invocation that includes the integration harness, with **AC-C1.7** post-render verification. Historical quotations elsewhere are retained and remain clearly labelled non-normative |

**Invariants explicitly preserved (re-verified, not assumed).** Valid manifest states stay
**exhaustive at exactly three**; **MP5 stays RETIRED** and un-renumbered; **witness-first then
atomic manifest transition** is unchanged; the **coordinated two-file rollback residual (R22)**
remains stated honestly and is **not** claimed as runtime-detectable; **fail-closed** behaviour on
malformed state is unchanged; the **literal 8/8 H0 red phase** is unchanged; **default full-suite
green task boundaries** are unchanged; **dependency-closed B-task ordering** (subsequence, not
prefix) is unchanged; and the **live 3a / 3e-absence** references in AC-C2.4 are untouched.

**Scope discipline (rev 13)**: no new task, sub-epic, fixture, harness **function**, file, CI step,
ledger entry, CODEOWNERS line, policy ID, shipment, or dependency edge. **No new acceptance
criterion** — AC-C1.3, AC-C1.6, AC-C2.7, AC-C2.9, AC-C2.10, and AC-C2.11 are **clarified in place**.
One `gate()` subsection restructured (§5.0.1), one evidence point rewritten with a paired companion
(§5.0.2 point 7 / 7a), one risk row extended (**R19**). AC-ID parity re-swept: **57 plan IDs ↔ 57
task IDs**, exact bijection maintained, and **content** parity now verified for both C-block pairs.
Shipment **017-S remains one shipment** with unchanged membership (9 items) and unchanged dependency
order (A1→A2→A3→{B1,B2,B3}→C1→C2). **No task is re-sized**: every change is a clarification of an
assertion's mode or evidence form, not of the work any task performs. No production code, test,
script, workflow, policy, agent, or skill file is modified by this pass — those remain Ship's
execution surfaces.

**Deliberately not claimed (rev 13).** Stage did **not** perform the authorized adversarial
re-review — that is the next action and it is **pending**. Stage did **not** push, open, update,
comment on, or merge PR #54, did **not** reply to or resolve any thread (`PRRT_kwDOTPuhps6hstWa`,
`PRRT_kwDOTPuhps6hstWg`, `PRRT_kwDOTPuhps6hstWn` remain **unresolved**), and did **not** claim,
modify, or close shipment `017-S`. **No success is claimed for the re-review that has not yet run.**

<!-- plan-review-attempt: 2 -->

### 12.14 Revision 14 — second operator-authorized capped remediation

**Authorization (explicit, and recorded because the budget was already exhausted twice).** The
rev-13 session ended **NOT READY**: the single authorized adversarial re-review ran (four models, no
route degradation) and returned **0 P0**, **one LOW-confidence P1**, and **four LOW-confidence P2**
`post_remediation_residual` findings (memory §22). The operator **explicitly authorized ONE more
Stage remediation plus adversarial re-review cycle** for PR #54. This revision is that single
authorized remediation pass. The authorization is **consumed** by it; no further remediation cycle
is implied, and **the re-review it enables has NOT been performed by Stage**.

**The five findings, and their disposition.** All five are `post_remediation_residual`; none is P0;
all are **same-contract completion** under P-021.

| # | Conf. | Sev. | Finding | Disposition |
|---|---|---|---|---|
| 1 | LOW | **P1** | **Executable copied-fixture default-mode mechanism absent.** §5.0.2 point 7 required a default-mode `go test` against a copied terminal repository under `t.TempDir()` but named only bare commands. The harness's `repoRoot(t)` helper resolves the **real checkout** (it walks up to `go.mod` from the process working directory), and the observation lives **inside** C2, so an unguarded re-run would recurse. As written the point could neither see the fixture nor produce the required C1-failure-plus-C2-execution evidence | **FIXED — one concrete bounded mechanism, specified normatively.** §5.0.2 point 7 now specifies: (i) `copyTrackedRepoFixture(t)`, a deterministic `git ls-files -z` **tracked-file** copy into `t.TempDir()` (real tree never written; `git status --porcelain` empty before and after); (ii) removal of **only** the copied job-`lint` step via the structural `jobs.lint.steps[]` walk, with the copied manifest and witness left **valid and in agreement**; (iii) `runFixtureChildGoTest(t, fixtureRoot)`, a **child `go test` process** whose `exec.Cmd.Dir` is the **copied module root** — which is exactly what makes `repoRoot(t)` resolve the copy — invoking `go test -json ./tests/integration -run '^TestDirectPushGate_(CIWiringIsBlocking\|FullCorpusClean)$' -count=1` with **no `-gatetask`** (default activation mode; `-run` bounds **process scope** only); (iv) the non-recursive sentinel `DIRECTPUSH_GATE_FIXTURE_CHILD=1`, which disables the **parent-only spawn block** and nothing else — it **MUST NOT** skip C2 and **MUST NOT** bypass MP1/MP6 (§5.0.1 "The fixture-child sentinel"); (v) **parsed `go test -json` events**, not scraped text — a terminal `fail` event for `TestDirectPushGate_CIWiringIsBlocking` citing the missing step, a terminal `pass`/`fail` event for `TestDirectPushGate_FullCorpusClean` (`skip` or absent **fails**), a **non-zero** child exit, **no** recursion/timeout marker, and a **malformed stream fails**; (vi) an explicit `context.WithTimeout` and **bounded** failure-only stdout/stderr, invoked through `exec.CommandContext` on the Go binary with **no shell**. Point **7a** stays separate as selector evidence only. Synchronized into AC-C1.6, AC-C2.11, §8, §11 **R19**/new **R24**, `018-F`, `018.010-T`, `018.011-T`, the deliberation, and the memory record |
| 2 | LOW | P2 | **H0 eligibility citation overstated.** Rev 13 claimed "**no step** of `_ship.agent.md` Step 2 consults the dependency graph". Step 2's closing paragraph does direct Ship to *verify the dependency graph before proceeding* when dependency operations are supported | **FIXED — corrected to the faithful reading, inventing nothing.** §5.0 "H0 eligibility" now states that Step 2 may **verify dependency-graph VALIDITY**, while harness generation remains an **upfront batch over the explicitly selected full shipment task set**, and dependency **READINESS** gates execution **order** (Step 3) and **claim** (Step 4). Because `${input:tasks}` is **Optional** and its omission falls back to "all ready tasks under the feature", the plan now **requires Ship to pass the exact eight task IDs** through that parameter, with **set equality against shipment `017-S`** (manifest minus the covering feature `018-F`, resolved via `parent_id`, never executed) checked **before** invocation and Step 2 item 4's **all-queued postcondition** halting on omission **after** it. All eight statuses are **`queued`** (verified in `.backlogit/queue/`), and `blocked` is a **lifecycle status value**, not merely an unmet dependency. Recorded on both sides of the call in AC-C2.11 evidence point 1 |
| 3 | LOW | P2 | **Targeted C2 red under-specified, and the bookkeeping rule unqualified.** §5.0.2 point 4 accepted any non-zero exit for `-gatetask=C2`; and "green by assertion, never bookkeeping" read as absolute while C2's completion legitimately finalizes control state | **FIXED — both halves.** Point 4 is **anchored**: the run must print `--- FAIL: TestDirectPushGate_FullCorpusClean` **by that exact name** and cite the **MP2 non-terminal-state** reason (seven-element `completed`, `phase: build`, `terminal: false`, witness absent), and five impostor exits are **explicitly rejected** as evidence — compile error, malformed state (MP1/MP6), missing `bash`, invalid selector, any unrelated non-zero. The bookkeeping rule is **qualified, not weakened** (§5.0.1 lifecycle item 4): for **A1–C1** completion state alone **never** makes the substantive test pass; **C2 is the deliberate exception of SUBJECT**, because terminal manifest + witness finalization *is itself* part of C2's substantive verification contract (MP2/AC-C2.9/AC-C2.10) — but C2 **still cannot pass from bookkeeping alone**: it must execute and pass the full corpus, self-test, mutation-proof, marker-free-construct, and regeneration assertions, with MP2 evaluated from manifest state **before** any surface-dependent assertion |
| 4 | LOW | P2 | **Memory current-artifact index stale** — §3's table still named plan **revision 12** and deliberation addenda through **§16** | **FIXED, HISTORY INTACT.** The memory record's §3 table and its cross-reference currency note now name plan **revision 14** and the deliberation's **§18** addendum. Nothing is erased: the note retains the rev-7/rev-12 recurrence history and adds the rev-14 instance, and the new §23 records this pass |
| 5 | LOW | P2 | **Deliberation addenda index stale** — the live header list stopped at **§16 (rev 12)**, omitting §17 | **FIXED.** The header's `Addenda` line now runs through **§17 (rev 13)** and **§18 (rev 14)**, and §18 is the new addendum |

**Invariants explicitly preserved (re-verified, not assumed).** Selector **activation precedence**
is unchanged; **MP5 stays RETIRED** and un-renumbered; **MP6 integrity stays mode-independent**;
valid manifest states stay **exhaustive at exactly three**; the **literal 8/8 H0** red phase and its
new **set-equality** check are unchanged in strength; the **default full-suite green** task
boundaries are unchanged; the **terminal witness** contract (witness-first, then atomic manifest
replacement) is unchanged; the **coordinated two-file rollback residual (R22)** remains stated
honestly and is **not** claimed as runtime-detectable; the **aggregate base→head** Stage
verification, the **no-shipment fail-closed handback pair**, the
**branch-before-mutation / commit-before-handback** ordering, and the **exact CI provenance
distinction** (observed installed command vs. `{{TEST_COMMAND}}` vs. recorded render input vs. the
guaranteed full-suite `./...` invariant) are all untouched.

**Scope discipline (rev 14)**: no new task, sub-epic, fixture, harness **function**, test, file, CI
step, ledger entry, CODEOWNERS line, policy ID, shipment, or dependency edge. **No new acceptance
criterion** — AC-C1.6 and AC-C2.11 are **clarified in place**. The point-7 mechanism is implemented
as **unexported helpers inside the existing `tests/integration/directpush_gate_test.go`**, which is
why it adds no harness function. One risk row extended (**R19**) and one added (**R24**, the
child-process mechanism's own failure modes). AC-ID parity re-swept: **57 plan IDs ↔ 57 task IDs**,
exact bijection maintained, and **content** parity re-verified byte-for-byte for both C-block pairs.
Shipment **017-S remains one shipment** with unchanged membership (9 items) and unchanged dependency
order (A1→A2→A3→{B1,B2,B3}→C1→C2). **No task is re-sized**: every change is a clarification of an
assertion's mechanism, mode, or evidence form, not of the work any task performs — the point-7
helpers are authored at **H0** by `harness-architect`, not by C1 or C2. No production code, test,
script, workflow, policy, agent, or skill file is modified by this pass — those remain Ship's
execution surfaces.

**Deliberately not claimed (rev 14).** Stage did **not** perform the authorized adversarial
re-review — that is the next action and it is **pending**. Stage did **not** push, open, update,
comment on, or merge PR #54, did **not** reply to or resolve any thread (`PRRT_kwDOTPuhps6hstWa`,
`PRRT_kwDOTPuhps6hstWg`, `PRRT_kwDOTPuhps6hstWn` remain **unresolved**), and did **not** claim,
modify, or close shipment `017-S`. **No success is claimed for the re-review that has not yet run.**

### 12.15 Revision 15 — third operator-authorized capped remediation, with a decomposition audit

**Authorization (explicit, recorded, and consumed by this pass).** The rev-14 session ended with the
authorized adversarial re-review **still pending**. The operator **explicitly authorized ONE narrow
remediation plus adversarial re-review cycle** for PR #54, and additionally **required a
decomposition audit** of `018-F`, shipment `017-S`, and tasks `018.004-T`–`018.011-T` — with special
attention to C2 `018.011-T` — **before** any editing. This revision is that single authorized
remediation pass. The authorization is **consumed** by it; **the re-review it enables has NOT been
performed by Stage**, and no outcome for it is claimed.

#### Phase A — decomposition audit (performed BEFORE any edit)

**VERDICT: `SUFFICIENTLY_DECOMPOSED`. No task is split, added, merged, re-parented, or re-sized, and
`017-S` membership is unchanged at 9 items.** The verdict is evidence-based, and the three questions
the operator separated are answered separately, because they have different answers' shapes.

**(1) Feature decomposition — eight tasks across three sub-epics.** The dependency graph is explicit,
acyclic, and recorded in `.backlogit/queue/`: `A1 018.004-T → A2 018.005-T → A3 018.006-T →
{B1 018.007-T, B2 018.008-T, B3 018.009-T} → C1 018.010-T → C2 018.011-T`. Sizes and complexity are
assigned on **both** axes and none is derived from the other: `S/low`, `S/medium`, `M/high`,
`M/medium`, `S/medium`, `L/medium`, `S/medium`, `S/medium`. Every task maps to **exactly one** harness
function in the §5.0 per-task map, so every task has **exactly one observable red→green transition** —
the Atomic Milestone criterion, met structurally rather than by assertion. The B-block is deliberately
**per-surface scoped** precisely so that B1/B2/B3 have independent transitions; §5.0 records that a
whole-corpus assertion would only go green after all three landed and would trip `build-feature`'s
5-attempt circuit breaker. That is the signature of decomposition that has **already been tested
against the execution loop**, not merely partitioned on paper.

**(2) Task C2's implementation width — the operator's specific concern, and the decisive evidence.**
The concern is that C2 has accumulated a test harness, fixture helpers, and evidence capture. **It has
not, and §5.0 already settles this in terms that predate this audit.** The H0 deliverable list, item 1,
states that the fixture helpers `copyTrackedRepoFixture` and `runFixtureChildGoTest` **and** the
parent-only spawn block are authored **at H0 by `harness-architect`**, that they are **not** gate
functions, that the function count **stays eight**, that they hold no `gateOrder` entry — and,
verbatim, that **"no task authors them, and no task is re-sized by them."** The helpers are therefore
**harness infrastructure produced before any task is claimed**, on the Ship Step 2 boundary, not C2
implementation work. What C2 itself performs is: **run** the scans, gates, and suite; **capture**
evidence; and **write two control-state files** (`terminal-witness.json`, then the atomic terminal
`harness-state.json` replacement). That is **one skill domain — verification/evidence** — and **two
written files**, inside the fewer-than-three-files heuristic. **This revision's own edits land
entirely inside those H0-authored helpers and the prose describing them, so they cannot widen C2
either** — which is why the audit was run **before** the edits rather than after.

**Why AC count is not width, and why the fixture mechanism is ONE atomic verification concern.** C2
carries 11 acceptance criteria — the second-highest count, behind `018.007-T`'s 12 at size `M`. Criterion
count measures **how many propositions one milestone asserts**, not how much work it takes; the
2-Hour Rule governs effort. The fixture mechanism is the clearest case: its multiple helper operations
— enumerate the index, copy, copy and validate the witness, `git init`/`add`/commit, assert
`--show-toplevel`, remove the lint step, spawn the bounded child, parse the event stream, compare the
two status snapshots — are **internal steps of a single assertion**, evaluated in **one** parent
invocation of `TestDirectPushGate_FullCorpusClean` and yielding **one** verdict (§5.0.2 point 7 item 7:
"any one of them failing fails `TestDirectPushGate_FullCorpusClean` in the parent"). They produce no
independent red→green transitions and no separately shippable outcome, which is exactly the test for
whether work should be a separate task.

**Why splitting C2 would ACTIVELY BREAK the design — the decisive structural argument.** A split is not
merely unnecessary here; it is **incoherent** with the lifecycle. (a) Every task requires its own harness
function, and §5.0 fixes the function count at **eight**; a ninth task would need a ninth function,
which the plan forbids. (b) C2's completion is defined by an **atomic two-file terminal transition** —
witness written **first**, then the manifest **atomically** replaced (§5.0.1) — and MP6 **fails closed**
on any witness/manifest disagreement. Splitting C2 across two tasks would necessarily create an
**intermediate state in which one file has flipped and the other has not**, which is **precisely the
fail-closed state MP6 exists to reject**. The split would therefore manufacture the defect class the
mechanism was built to detect. (c) `017-S` must remain **one shipment**; adding tasks to satisfy a
numeric preference is explicitly declined, per the operator's own instruction not to.

**(3) Shared harness-state edits.** Each task inserts **one line** into `completed` in the shared
`tests/integration/testdata/directpush-gate/harness-state.json` at its `gateOrder` position. §5.0.2
already records this as the **completion record, not a second skill domain**, and confirms no task's
size band changes on account of it. Re-verified at rev 15 and unchanged. This is shared-state
**coordination**, which the dependency graph serializes, not shared-state **contention**.

**Did the P2 H0 ambiguity itself require restructuring? Explicitly evaluated — NO.** The operator
required this be decided rather than assumed. The ambiguity lives in the **installed skill's wording**
(`harness-architect` Step 1 item 1's "ready descendants" and item 3's "otherwise non-ready" versus item
2's explicit `${input:tasks}` restriction), **not** in this feature's shape. Every restructuring that
could dissolve it is **worse**: deleting the dependency edges would **falsify a real, verified ordering
constraint** to satisfy a parser; harnessing only the dependency-ready A1 would **halt at Step 2 item 4**
anyway; and splitting or merging tasks changes nothing, because any decomposition into ordered work has
dependency state. §5.0 therefore handles it **operationally and fail-closed** (below) and **no backlog
artifact changes**.

#### Phase B — the findings and their disposition

| # | Sev. | Finding | Disposition |
|---|---|---|---|
| 1 | **P1 MEDIUM** | **Real-tree cleanliness deadlock.** §5.0.2 item 1 asserted `git status --porcelain` over the real checkout was **empty before and after** the fixture run. That precondition is **unsatisfiable exactly when the evidence is required** | **FIXED — replaced by byte-identical full status snapshots, which is the STRONGER predicate.** §5.0.2 item 1**(f)** now takes `git status --porcelain=v1 -z --untracked-files=all` **before and after** the copied-fixture child run — the same NUL-safe, untracked-inclusive form §5.7 already mandates for the Stage commit gate — parses records **NUL-wise**, and requires the two snapshots to be **byte-identical**. **The parent fails if and only if the fixture mechanism changed the real checkout.** Legitimate C2 manifest/witness dirt is tolerated **only by being unchanged**. **Two independent, repository-verified reasons** the old form was unsound are recorded: (a) the spawn block is entered **only** in `terminal` phase, so C2's own witness and manifest writes are **necessarily present**; (b) `references/herdr` is excluded **only** by `.git/info/exclude`, which is **local and does not survive a clone**, so "empty" fails on CI for an unrelated reason. **No weakening**: equality additionally catches modification of an already-dirty file, deletion, and a new artifact written beside existing dirt — none of which an empty-tree check can see. Synchronized into AC-C1.6, AC-C2.2, AC-C2.11 point 7, §8, and **R25** |
| 2 | **P1 LOW** | **Fixture was not a self-contained Git repository.** The copy excluded the real `.git/`, so Git commands inside it would **walk upward** to an unrelated ancestor repository, or none | **FIXED — deterministic, executable construction.** §5.0.2 item 1**(d)**: `git init` in the copy; **local, non-secret** `user.email`/`user.name` (never `--global`, no credentials); `git add`; a **deterministic local commit** — preferred over a bare index so tracked-corpus **and** `HEAD` consumers both work. **No `.git` copy from the real repo, no submodule recursion, no network.** The binding check is `git -C {copy} rev-parse --show-toplevel`, **canonicalized** (symlink, Windows `8.3`/case, macOS `/var`→`/private/var`) and asserted **equal to the copy root before the child is launched**; resolution to the real checkout, to any ancestor of the temp directory, or to any other repository **fails the observation**. This is what makes "no nested or sibling repo can be selected" **checkable rather than asserted** |
| 3 | **P1 LOW** | **Untracked terminal witness omitted.** `git ls-files` enumerates the **index**; `terminal-witness.json` is written by **C2's own completion** and may not be indexed when the helper runs, so the fixture could be **witness-absent with a terminal manifest** — an MP6 **integrity** failure, making point 7 vacuous for a reason unrelated to the lint step | **FIXED — copied explicitly, validated, staged, and ORDERED.** §5.0.2 item 1**(c)**: after the indexed copy and **before** the fixture commit, the helper copies the real checkout's **current** `terminal-witness.json` **from the working tree**, **requires it to exist** (absence **fails**; never "nothing to copy"), **parses** it, and **validates it against the copied manifest** to the full MP6 standard (exact eight-element set, `terminal: true`, `phase: terminal`, **matching digest**). It is staged by item 1(d)'s `git add`. Item 2 makes the ordering normative: **agreement is proven BEFORE the copied lint step is removed**, so the child's failure stays **attributable to the lint removal** |
| 4 | **P2** | **H0 readiness ambiguity.** Rev 14 cited only `harness-architect` Step 1 item 3 and asserted `blocked` is a lifecycle value. Step 1 item 1 **also** speaks of "**ready descendants**", and item 3 of "**otherwise non-ready**" work | **RECONCILED WITHOUT OVERSTATING CERTAINTY, AND WITHOUT INVENTING AN OVERRIDE.** §5.0 now quotes **all three** Step 1 items exactly, states the plan's reading (scope comes from `${input:tasks}`; the exclusion turns on the **status values** named), and records that **an unmet dependency is NOT `status: blocked`** — the seven downstream tasks carry `queued` plus `dependencies` edges, a different thing. It then states plainly what the plan **does not** claim: the phrases **are ambiguous**, and a reader may reasonably read "otherwise non-ready" as covering dependency-unready work. The ambiguity is handled **operationally and fail-closed**: Ship passes the **exact eight task IDs** and **verifies all eight are `queued`** immediately before invocation; if the returned selected set is **not equal** to the eight, **Ship MUST HALT BEFORE ANY MUTATION and route to the operator**, reporting the contradiction with Step 2 item 4's all-queued halt. **Silently proceeding with a partial batch is prohibited**, and relaxing the 8/8 red to accommodate one is prohibited. **No hidden override is invented and the installed skill is not amended** (§11). Decomposition impact **explicitly evaluated and declined** — see Phase A |
| 5 | **P2/P3** | **Gitlink handling and sentinel wording.** Index-mode handling was unspecified (`git ls-files -z` emits **paths only**); and the "recursion marker (sentinel already set **when the block is reached**)" is **unsatisfiable** given the guard that a sentinel-set process never reaches the block | **BOTH FIXED, AND ONE PREMISE CORRECTED ON EVIDENCE.** **Index modes**: enumeration becomes `git ls-files -s -z` (`-s` is load-bearing), with a **closed table** — `100644`/`100755` **copied** (executable bit preserved); `120000`, `160000`, **any** other mode, and **any unmerged stage** **FAIL CLOSED**. **The premise that a `160000 references` gitlink must be skipped is FALSE for this repository and is NOT written into the plan**: `git ls-files -s` shows **all 629 entries are `100644`** — zero `100755`, zero `120000`, zero `160000` — and **`references/` is not in the index at all**. It holds `references/herdr`, a **nested independent clone** excluded by `.git/info/exclude`, so enumeration never sees it and there is nothing to skip. The fail-closed rows are retained as a **forward guard**, not as dead text. **Sentinel**: the impossible child-side detector is **withdrawn** in §5.0.1, §5.0.2 item 5, §8, and **R24**; recursion is **PREVENTED parent-side** (entry requires the sentinel **unset** ⇒ depth bounded at exactly one **by construction**) and **BOUNDED** by the child's explicit `context.WithTimeout`, whose deadline termination **fails** the observation. A parent-side self-check that its own `exec.Cmd.Env` carries the sentinel is permitted but is **not** recursion control |
| 6 | **PROCESS** | **Rev 14 commit footer/trailer mismatch** | **RECORDED AS PROCESS HYGIENE; rev 14 is NOT amended** (history stays intact). The rev-14 commit `da38be5` diverged from `commit-message.instructions.md` in four ways: scope `plan` is **not** in the allowed set (`tui, server, hub, tunnel, session, config, ci, docs`); the body far exceeded the **<300 byte** limit; the footer carried **no emoji** and did **not** end with `- Generated by Copilot`; and the trailer used `Copilot <copilot@github.com>` instead of the required `Copilot <223556219+Copilot@users.noreply.github.com>` (which rev 13 had used correctly). **This revision's commit conforms**: allowed type/scope, description **<100 bytes**, body **<300 bytes**, footer after a blank line with an emoji ending exactly `- Generated by Copilot`, and the exact trailer on a later trailer line |

**Invariants explicitly preserved (re-verified at rev 15, not assumed).** The bounded child keeps its
`exec.Cmd.Dir`, sentinel, and **JSON event** contract; **selector activation precedence** is unchanged;
**MP5 stays RETIRED** and un-renumbered; **MP6 integrity stays mode-independent**; valid manifest states
stay **exhaustive at exactly three**; the **literal 8/8 H0** red phase and its set-equality check are
unchanged in strength; the **default full-suite green** task boundaries are unchanged; the
**witness-first, then atomic manifest replacement** transition is unchanged; the **aggregate base→head**
Stage verification, the **no-shipment fail-closed handback pair**, the **branch-before-mutation /
commit-before-handback** ordering, and the **exact CI provenance distinction** are all untouched. The
**coordinated two-file rollback residual (R22)** remains stated honestly and is still **not** claimed as
runtime-detectable.

**Scope discipline (rev 15)**: no new task, sub-epic, fixture, harness **function**, test, file, CI step,
ledger entry, CODEOWNERS line, policy ID, shipment, or dependency edge. **No new acceptance criterion** —
AC-C1.6, AC-C2.2, and AC-C2.11 are **clarified in place**. One risk row corrected (**R24**) and one added
(**R25**). **No task is re-sized**, and no backlog item is split, merged, or re-parented: the point-7
helpers are authored at **H0** by `harness-architect`, never by C1 or C2. Shipment **`017-S` remains one
shipment** with unchanged membership (**9 items**) and unchanged dependency order. No production code,
test, script, workflow, policy, agent, or skill file is modified — those remain Ship's execution surfaces.

**Deliberately not claimed (rev 15).** Stage did **not** perform the authorized adversarial re-review —
that is the next action and it is **pending**. Stage did **not** push, open, update, comment on, or merge
PR #54, did **not** reply to or resolve any thread, and did **not** claim, modify, or close shipment
`017-S`. **No success is claimed for the re-review that has not yet run.**

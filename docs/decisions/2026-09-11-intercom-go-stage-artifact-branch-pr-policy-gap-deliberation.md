---
title: "Deliberation — Stage Artifact Branch/PR Policy Gap Correction"
date: 2026-09-11
status: accepted
agent: Stage
governs: stash 638A410B
---

# Deliberation — Stage Artifact Branch/PR Policy Gap Correction

- **Date**: 2026-09-11
- **Agent**: Stage (operator-authorized, interactive)
- **Stash entry**: `638A410B` (high, feature-shaped)
- **Planning branch**: `chore/stage-pipeline-policy-gap` — Stage artifacts only. **Stage commits on
  this branch and does nothing else with it**; push, PR **#54**, and merge are performed by the
  Orchestrator or the operator, which is the exact route this deliberation exists to install.
  (Rev 8 corrects the header's original "no push, no PR" claim, which was true when written and
  became false once PR #54 was opened — a session record that describes a state the repository has
  left is read as current by the next agent.)
- **Addenda**: §8 (rev 2), §9 (rev 3), §10 (rev 6), §11 (rev 7), §12 (rev 8), §13 (rev 9)
- **Evidence commit**: `fdff9e4` — a Stage no-shipment decision artifact pushed directly to
  `main`. Treated as **historical evidence only**; not reverted, not rewritten.

---

## 1. Framing — what problem are we actually solving?

A Stage planning artifact reached the default branch without a pull request. That was not an
agent misbehaviour: **the contract authorized it.** Two coupled surfaces made the direct push
the compliant reading.

### 1.1 Surface A — the policy root cause

`.github/policies/workflow-policies.md`, **P-010 Agent Role Boundary**, the **Stage MAY** block:

> **Stage MAY** (within its legitimate scope):
> - Commit backlog and planning artifacts *(on the default branch or a dedicated chore/admin branch)*

The same policy's **Ship MUST NOT** block already reads:

> **Ship MUST NOT**: … Commit or push directly to `main`

So P-010 is **asymmetric**: Ship is forbidden from touching `main` directly, Stage is expressly
permitted. This clause is the authorization that makes the Orchestrator instruction below
internally consistent. Fixing the Orchestrator alone would leave the contradiction alive in the
authoritative policy registry.

**Anchoring note (rev 8).** Both bullets above were originally cited by line ordinal (`L230`,
`L258`). They are now anchored by **block lead-in plus quoted text**, which is what §8.5 said it
was switching to and did not. The line numbers were not merely brittle — B2 appends an Amendment
Log row to this very file, so every ordinal below the insertion point shifts as a direct result of
the change this deliberation governs.

### 1.2 Surface B — the Orchestrator instruction

`.github/agents/_orchestrator.agent.md`, **Step 1.5 Staging Artifact Merge Gate**, sub-steps
3a–3d already mandate the correct behaviour unconditionally:

```text
a. Commit any uncommitted backlog files to a staging branch: `chore/stage-{shipment_id}`
b. Push the staging branch and create a PR to `main`
c. Wait for the staging PR to merge (operator approval required)
d. After merge, pull `main` and proceed to step 4
```

Sub-step **3e (L287)** then reverses it — instructing that a direct push to `main` be attempted
first, with the staging PR used only as a rejection fallback, justified as "deterministic
regardless of when branch protection was enabled or changed."

Three independent defects in 3e:

1. **It contradicts 3a–3d.** Those steps are unconditional; 3e re-enters them conditionally.
2. **Its determinism claim is inverted.** 3a–3d are *strictly more* deterministic: they never
   depend on a server-side rejection whose presence is unverified (see §1.4). 3e's behaviour is
   a function of remote branch-protection configuration; 3a–3d's is not.
3. **Its fallback body is a verbatim restatement of 3a–3c.** Deleting 3e therefore loses *no*
   behaviour — the fallback path it describes is the path 3a–3d already require.

### 1.3 The in-file counter-precedent

`_orchestrator.agent.md`'s own *Elective Agent Behavioral Notes* already state that both agents
"never commit directly to the default branch." 3e violated a norm asserted elsewhere in the very
same file. This is a self-inconsistent generated artifact, not a considered design.

### 1.4 Verified blast radius

A scan of the tracked harness contract surface
(`.github/agents/`, `.github/policies/`, `.github/skills/`, `.github/instructions/`) for
direct-push-to-default-branch language returns **exactly one hit**:
`_orchestrator.agent.md:287`. `_stage.agent.md`, the subagents, the skills, and the instructions
are clean. The correction is therefore genuinely small, and a
post-fix zero-findings acceptance is achievable (see D5).

**Recorded unknown**: no decision record, closure record, or docs file establishes whether GitHub
branch protection is actually enabled on `main` for this repository. 3e was written precisely to
paper over that uncertainty. We do **not** assume protection exists; we remove the dependence on
it. This is stated rather than assumed.

---

## 2. Constraints carried from the operator

| # | Constraint | Consequence for this deliberation |
|---|---|---|
| C1 | Smallest complete contract change | No generalized branch framework, no new policy ID, no waiver mechanism |
| C2 | Mechanical enforcement, not prose-only | A deterministic CI gate is mandatory, not optional |
| C3 | Source templates and installed artifacts stay coupled | Must be resolved against the untracked-template finding (D2) |
| C4 | Preserve role isolation | Orchestrator owns staging branch/PR orchestration; Stage owns planning artifacts; Ship owns implementation delivery |
| C5 | Do not revert `fdff9e4` | Historical evidence preserved |
| C6 | Preserve branch and PR merge-commit rules | P-009 merge-commit-only, P-011, P-016 untouched |

---

## 3. Options considered

### Option 1 — Prose-only correction

Delete 3e, reword the P-010 bullet. Nothing else.

- **Pro**: minimal diff; satisfies C1 maximally.
- **Con**: violates **C2 outright**. Both files carry
  `Generated by autoharness | Template: …` provenance; a future `install`/`tune` re-render
  restores 3e silently and no signal fires. The gap reopens undetected. The evidence commit
  `fdff9e4` proves prose alone did not hold the line even *before* regeneration.
- **Verdict**: **Rejected** — fails the reliability requirement that motivated the request.

### Option 2 — Prose correction + deterministic regression gate (CHOSEN)

Delete 3e; make the P-010 Stage bullet symmetric with the existing Ship prohibition; add
`scripts/check-direct-push-language.sh` asserting the invariant over the **installed** contract
surface, wired into the existing `lint` job in `.github/workflows/ci.yml`; record the divergence
in the ci.yml LOCAL DIVERGENCE ledger.

- **Pro**: satisfies C2 with a mechanism the repository already uses five times over
  (`check-retired-architecture.sh`, `check-write-path-precondition.sh`,
  `check-gitignore-append-only.sh`, `check-unignore-regression.sh`, `check-depguard-fixtures.sh`).
  Regeneration that restores 3e turns CI red — which is the *only* enforceable coupling available
  given D2. Reuses house conventions rather than inventing a shape, satisfying C1.
- **Con**: adds one script + fixtures + **one blocking CI step**. Accepted as the irreducible cost
  of C2.
- **Verdict**: **Chosen.**

### Option 3 — Option 2 + a new dedicated policy ID (e.g. P-022 "Stage Branch-Before-Persist")

- **Pro**: maximally explicit; a named policy is greppable.
- **Con**: **violates C1.** P-010 already *is* the role-boundary policy and already contains the
  exact parallel prohibition for Ship. The correct change is to make one existing asymmetric
  bullet symmetric — not to add a second policy that partially duplicates P-010 and creates a new
  precedence question against it. Rejected as invented framework.
- **Verdict**: **Rejected.**

### Option 4 — Enforce via GitHub branch protection / ruleset configuration instead

- **Pro**: a true server-side control; unbypassable by a local agent.
- **Con**: out of repository scope (not a tracked artifact), unverifiable from the repo, and
  **orthogonal** — it would stop the push but leave the contract still *instructing* the push,
  so the agent would keep attempting a forbidden action and the contradiction would persist.
  Also cannot be asserted by CI. Does not address the root cause.
- **Verdict**: **Rejected as the primary mechanism.** Noted as a complementary operator-side
  hardening, explicitly out of scope here.

---

## 4. Key decisions

### D1 — Correct both surfaces, not just the Orchestrator

The P-010 bullet is the *authorization*; Step 1.5 3e is the *instruction*. Removing only the
instruction leaves the authorization standing, and the next regeneration or the next agent
reading P-010 reopens the gap. Both move together, in one shipment.

The P-010 Stage bullet becomes symmetric with the existing Ship prohibition:

> - Commit backlog and planning artifacts **on a dedicated Stage/admin branch merged via PR —
>   never directly to the default branch**

and the clause is extended so it binds **even when no shipment is formed** — the exact case that
produced `fdff9e4`, which was a no-shipment decision artifact and therefore fell outside any
shipment-keyed staging flow.

An **Amendment Log** row is appended per the registry's own additive convention (new minor
version; "Corrects, and does not delete or edit, the N.N.N row above"), matching the worked
precedent of the P-015 1.21.0 supersession note.

### D2 — Template coupling: assert the invariant over installed artifacts

**Investigation finding that changes the shape of scope item 1.** The `.tmpl` sources
(`.copilot/installed-plugins/autoharness/autoharness/templates/agents/_orchestrator.agent.md.tmpl`
L287, and the `.autoharness/staging/` mirror) **are not tracked by git**: `.gitignore:60` excludes
`.copilot/` and `.autoharness/.gitignore:3` excludes `staging/`. Confirmed via
`git check-ignore -v` and `git ls-files --error-unmatch`.

Consequences, recorded rather than assumed:

- An edit to those templates **cannot be committed, diffed, or reviewed**, and is **overwritten
  by the next `autoharness install`**, because they are vendored upstream plugin content.
- Therefore **no CI check can enforce template↔artifact equality** in this repository.

Decision: **do not edit the untracked templates.** Instead, satisfy C3 with the only coupling
that is actually enforceable here —

1. the gate asserts the **invariant** (absence of the direct-push instruction shape) over the
   **installed** `.github/**` artifacts, so a regeneration that restores 3e from the upstream
   template turns CI red; and
2. the ci.yml **LOCAL DIVERGENCE ledger** is extended to name this change, per the established
   in-repo convention for generated-file divergence.

This is a deliberate, disclosed narrowing of scope item 1, and it is *stronger* than editing an
untracked file would be: an untracked edit is invisible to review and erased on reinstall, whereas
the gate detects the reintroduction regardless of which path reintroduces it. The upstream
template defect is recorded for an upstream report; fixing upstream is out of scope.

### D3 — Detector shape: structural, prohibition-aware, self-exclusion by pathspec

The single highest-risk element. Prior art —
`docs/compound/2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md` — records a
CI assertion that **permanently self-failed** because a bare-substring match caught the step name,
comments, and error text that had to mention the forbidden string to be comprehensible. This gate
is squarely in that class: it must reject direct-push instruction language while its own header,
fixtures, error strings, CI step name, **and the new P-010 sentence forbidding the practice** all
necessarily discuss it.

Mitigations, all mandatory and all asserted in `--self-test`:

1. **Match a violation *shape*, not a mention** — an *imperative instruction* to push/attempt a
   push to the default branch. A descriptive or prohibitive sentence is not a violation.
2. **Prohibition-aware**: a candidate line carrying a negation/prohibition marker
   (`never`, `must not`, `MUST NOT`, `do not`, `forbidden`, `prohibited`) is **not** a finding.
   This is what lets P-010's new prohibition sentence live inside the scanned corpus.
3. **Corpus scoped to the contract surface** via `git ls-files` pathspec:
   `.github/agents/**`, `.github/policies/**`, `.github/skills/**`, `.github/instructions/**`.
   `docs/**` is deliberately **outside** the corpus, so decision/plan/closure/memory records
   (including this file, which quotes 3e) never trip the gate. This pathspec is the
   **authoritative and only** self-exclusion mechanism: because `scripts/**` is not a member of
   it, the checker itself and `scripts/testdata/**` are excluded **by construction**. Rev 2's
   separate path-self-exclusion assertion is therefore **withdrawn as vacuous** — asserting that
   a never-selected path is absent from the selection can never fail, which is precisely the
   rev-1 vacuity failure mode (§10.1). Non-vacuity of the *selection* is covered by item 4.
4. **Non-vacuity**: `--self-test` fails if the fixture corpus or the selected file set is empty,
   and reports the resolved corpus file count.

### D4 — Task ordering keeps `main` green while remaining test-first

Constitution Principle II requires test-first. Prior art
(`docs/decisions/2026-09-08-…-retired-arch-gate-detection-quality-deliberation.md`, D2′) records
that shipping a deliberately-red gate previously **net-downgraded a live merge-blocking control**.

Resolution — ordering alone, with **no advisory-window toggle needed**:

| Order | Task | State of `main` |
|---|---|---|
| 1 | Gate script + fixtures + `--self-test`, **not yet wired to CI**. Real-tree scan **fails** (3e still present) — the intentional red. | green (gate not in CI) |
| 2 | Delete 3e; make P-010 symmetric + Amendment Log. Gate now passes. | green |
| 3 | Wire the single blocking CI step + LOCAL DIVERGENCE ledger; verify installed-surface coupling. | green |

Because the gate enters CI only *after* the violation is removed, `main` is never wedged and no
fail-open toggle is introduced. This is strictly safer than the bounded-advisory-window pattern
and simpler (C1). No toggle variable is added — the gate is **blocking from the moment it is
wired**, consistent with the post-`015.013-T` fail-closed posture.

### D5 — Zero-findings acceptance at HEAD

Carried from the retired-arch D5 constraint: every detector must be proven against the real tree,
not merely against fixtures. After task 2, a full corpus scan at HEAD **must report zero
findings**, and this is a hard acceptance criterion on task 3. If a legitimate line trips the
detector, the **pattern is narrowed or a path is documented-and-excluded** — coverage is never
reduced by weakening the corpus (retired-arch D4).

### D6 — Threat-model honesty

The gate is an **anti-accident control, not an anti-adversary control**, and its header must say
so verbatim in the house style: GitHub's `pull_request` trigger runs the workflow and this script
from the PR's own head, so a PR that violates the invariant can also edit the gate. It is
effective against regeneration, drift, and careless reintroduction. The companion residual —
"the CI workflow is also autoharness-generated and may overwrite the wiring step on a future
render" — is disclosed in the same header. CODEOWNERS review of the script + workflow is the
named mitigation; `pull_request_target` re-architecture is explicitly out of scope.

### D7 — Role isolation preserved (C4)

No role boundary moves. Step 1.5 remains the **Orchestrator's** gate — it keeps owning staging
branch creation, PR, and merge-wait; only its self-contradicting escape hatch is removed. P-010
continues to describe **Stage's** artifact authorship. Ship's implementation-delivery boundary,
P-009 merge-commit-only, P-011, and P-016 are untouched. This shipment changes *where Stage
artifacts may land*, never *who does what*.

---

## 5. Scope boundary

**In scope** (four authorization surfaces across three installed files, plus the gate):
`_orchestrator.agent.md` Step 1.5 **sub-step 3e** *and* the Step 1.5 **preamble** (§9.1);
`_stage.agent.md` Role Boundary **Git row** — the operative surface role-enforcement actually
reads (§8.1); `workflow-policies.md` P-010 Stage bullet + Amendment Log;
`scripts/check-direct-push-language.sh` + `scripts/testdata/directpush/`;
`.github/workflows/ci.yml` **one blocking `lint` step** + ledger.

**Explicitly out of scope**: reverting/rewriting `fdff9e4` (C5); editing untracked `.copilot/` or
`.autoharness/staging/` templates (D2); GitHub branch-protection configuration (Option 4); a new
policy ID (Option 3); any unrelated cleanup; the three dangling doc references cited by
`ci-topology-check.sh`; the missing compound entry on generated-artifact divergence (candidate
follow-up, not required here).

**P-021 note**: the corpus scan found no second instance of the offending language, so no
same-contract-surface completion work is pending. Should task 2 surface one, it is **in scope**
as a same-contract-surface completion; anything else requires a captured deferred-scope-expansion
stash entry.

---

## 6. Open questions

None blocking. The one recorded unknown — whether branch protection is enabled on `main` — is
deliberately rendered **irrelevant** by this design: 3a–3d never consult it.

---

## 7. Outcome

Option 2 accepted. Proceed to implementation planning under a single covering chore, ordered
test-first per D4. **Superseded in scope by §8 below** — the chosen direction stands; the surface
count and detector design are corrected.

---

## 8. Revision 2 addendum (2026-09-11, post plan-review)

The seven-persona plan review **falsified three claims made above**. They are corrected here
rather than silently rewritten, per the registry's additive convention.

### 8.1 §1.4 "exactly one hit" is WITHDRAWN

The scan behind §1.4 used a **push-shaped** predicate. It therefore missed a **commit-shaped**
authorization on a third surface:

`.github/agents/_stage.agent.md` **L42**, Role Boundary (NON-NEGOTIABLE) table, Git row, Allowed
column:

> Commit backlog/planning artifacts **on default or admin branch**; create/use an explicit,
> time-boxed spike/research worktree only for staging investigation

`.github/instructions/role-enforcement.instructions.md` makes **that table**, not P-010, the
authoritative permission set consulted at mutation time. So this — not P-010 — is the surface the
Stage agent actually read before producing `fdff9e4`. Correcting only P-010 and Step 1.5 would
have left the operative authorization intact and produced a **false green**: a gate reporting zero
findings while the gap stayed open.

Under **P-021 C1** this is a *same-contract-surface completion* and is therefore **in scope for
this shipment**, not a deferrable expansion.

**Consequence for D3**: the detector must match **commit-to-default-branch** authorization
constructs as well as imperative push constructs, and must scan **table cells**, not just prose.

### 8.2 The D1 P-010 rewrite as drafted was self-deadlocking — CORRECTED

P-010's **Stage MUST NOT** list already forbids Stage to "Create, push, or merge pull requests"
(mirrored at `_stage.agent.md` L44). A bullet requiring Stage artifacts to be "merged via PR"
without naming an actor therefore prescribed a path the same policy forbids the actor to walk —
an unsatisfiable instruction whose Violation Action is *halt*, worst in exactly the no-shipment
case that produced `fdff9e4`.

**Correction**: the rule must **separate authorship from delivery**. Stage *commits to* the
dedicated branch; **push, PR creation, and merge are performed by the Orchestrator under
Step 1.5, or by the operator in a direct Stage invocation — never by Stage.** Stage's PR
prohibition stays intact and unweakened.

### 8.3 "Deleting 3e loses no behaviour" is WITHDRAWN

3e sub-bullet L288 reads "Create branch `chore/stage-{shipment_id}` **from the current commit**".
Sub-step 3a (L283) says "Commit any uncommitted backlog files to a staging branch" and specifies
**no branch point**. Step 1.5 step 2 (L277–281) routes the *already-committed-but-unpushed* case
into step 3, and 3e's clause was the **only** text telling the agent how to move that work off
local `main`. Deleting 3e outright would leave that path uninstructed, dead-ending at step 4's
`STAGING_GATE_FAIL` — fail-closed, but stuck.

**Correction**: fold the branch-point into 3a as part of the deletion.

### 8.4 Corpus scope widened

`AGENTS.md`, `.github/copilot-instructions.md`, and `.github/prompts/**` are tracked,
autoharness-generated contract surfaces of equal authority and were outside the §D3 corpus. A
regeneration emitting the instruction into any of them would have passed silently — fail-open on
the exact threat the gate exists for. The corpus is widened accordingly (never narrowed, per the
retired-arch D4 rule).

### 8.5 Corrected citation

§1.1 originally cited the Ship prohibition at "L258". The correct line was **L238**; L258 falls
inside the P-011 header table. Anchors are switched to quoted text, per the ci.yml ledger's own
"by description, not line number" rule. **Rev 8 note**: the switch was *recorded* here at rev 2 but
not *applied* to §1.1, which continued to carry `L230`/`L258` as current evidence for six
revisions. It is applied now, in §1.1 and in every plan and task surface that cited the Ship
bullet by ordinal (plan §2, §5.3, §7.2/AC-C2.3; `018.011-T`). A correction that is announced but
never performed is indistinguishable from one that was not made.

### 8.6 Net effect

The chosen direction (Option 2) is **unchanged and still correct**. What changed is its *scope*:
**three** authorization surfaces, not two; a **structure-aware** detector covering commit-shaped
and table-cell constructs; and an explicit **actor attribution** so the new rule is satisfiable.


---

## 9. Revision 3 addendum — the fourth authorization surface

Round 2 of the plan-review gate found one further blocking defect in the framing. This addendum is
**additive**: §1-§7 and §8 stand as written, corrected here rather than rewritten.

### 9.1 Withdrawn: "three authorization surfaces"

§8.1 widened the count from one to three. **Three was still wrong.** A fourth surface exists: the
**preamble** of Orchestrator Step 1.5, which instructs:

> verify that all staging artifacts (backlog items, shipment manifests) are committed to the
> default branch and present on the remote.

This is **commit-shaped and marker-free**, so it survives both the original push-shaped grep and
the commit-shaped sweep that found `_stage.agent.md` L42 — it reads as a verification instruction,
not an authorization, yet it states the gate's **postcondition** as artifacts being committed *to*
the default branch. Leaving it intact would keep a compliant reading of direct commit-to-default
alive in the very step that is supposed to enforce the opposite, and it sits **inside the corpus**
the new gate scans, so it would also block zero-findings at C2.

**Correction**: D1's scope becomes four surfaces. The preamble is reworded to a
verification/postcondition form ("**have reached** the default branch **via a merged staging PR**")
under plan task B1. This form is explicitly outside the detector's construct 2, so no detector
narrowing is required to accommodate it.

### 9.2 Reinforced: D4's "enforce mechanically, not by prose"

That a fourth surface survived two rounds of deliberate manual search — by an agent that had
already been corrected once for undercounting — is direct evidence for D4. Manual enumeration of
contract surfaces is not reliable at this scale. The plan now records the **first full-corpus scan
output** (AC-A3.4) before any correction begins, so the detector, not a reader, establishes the
surface count. Risk R9 tracks the possibility of a fifth.

### 9.3 Grounding for the P-016 disposition

§5's claim that a Stage artifact branch does not consume the single-active implementation slot is
now grounded in P-016's own Statement text, which scopes to "exactly one agent-owned
**implementation** branch/worktree". A Stage artifact branch carries no source, test, or config
change. The plan additionally makes this contract-visible by adding the clarification to P-010's
text (task B2), rather than leaving it as a plan-only assertion.

### 9.4 Net effect

The chosen path (Option 2) is unchanged. D1 covers four surfaces instead of three; D4 is
strengthened with a mechanical surface-enumeration step; D5's P-016 disposition is grounded in
quoted policy text and promoted into the contract. No decision is reversed.

---

## 10. Revision 6 addendum — authorization without discovery is not a route

**Source**: PR #54 current-HEAD review, cycle 5, thread `PRRT_kwDOTPuhps6hqWDe`
(comment 3993998429). Operator-authorized, scope limited to this thread. Numbered to match the
plan revision it grounds (§8 → plan rev 2, §9 → plan rev 3, §10 → plan rev 6).

### 10.1 Withdrawn: "correcting the authorization surfaces completes the route"

D1 framed the problem as a set of **authorization** surfaces — text that permits a Stage artifact
to reach the default branch without a PR — and §8.2/§9.1 grew that set from two to four. Every
revision through plan rev 5 worked inside that frame. The frame was **incomplete**.

Correcting authorization *relocates* Stage's commit onto `chore/stage-{shipment_id}` /
`chore/stage-{chore-slug}`. It does not teach the gate that has to verify that commit where to
look. Orchestrator Step 1.5 discovers work through exactly two inputs — a `.backlogit/`-only
dirtiness pathspec and a **default-branch-only** commit range — and after the relocation both are
blind:

- a no-shipment Stage run writes only `docs/plans/`, `docs/decisions/`, `docs/memory/`, so the
  dirtiness pathspec reports **clean** — precisely the `fdff9e4` shape this deliberation exists to
  close;
- Stage's commit is not on the default branch, so a default-branch-against-its-own-remote range is
  **empty by construction**.

Both read "synchronized", the gate skips its commit/push/PR step, and its `origin/main`
verification then fails — or, for a shipment-less run, has no manifest to name at all. The route
§8.2 made *permissible* was never made *executable*.

### 10.2 The rev-2 self-deadlock lesson generalizes

§8.2 established that a rule is only sound when the **actor** who must satisfy it is authorized to
act. This finding is the same defect one layer down: a rule is only sound when the actor who must
**verify** it can **observe** the thing it verifies. §8.2's route table (carried into plan §6.2)
names the Orchestrator as the actor for push, PR, merge, and post-merge verification of the
no-shipment branch — while plan rev 2 simultaneously declared Step 1.5, the only Orchestrator gate
that performs any of those, out of scope for shipment-less runs. That position is **withdrawn**.

### 10.3 Decision — explicit handback, branch-specific discovery, two verification arms

D1 is extended, not reversed. The correction stays inside the existing four surfaces and adds no
new policy, task, or enforcement mechanism:

1. **Explicit handback.** Stage returns `stage_branch`, `stage_head_commit`, `stage_outcome`, and
   `stage_artifact_paths`; the Orchestrator requires and records them. Inference is a degraded,
   loudly-recorded fallback (`STAGING_HANDBACK_DEGRADED`) covering **every** field, so it never
   halts: branch and commit resolve from `HEAD`, outcome derives from whether a shipment was
   formed, and paths default to the artifact path set. When the `HEAD` resolution yields the
   default branch the branch is left **unresolved** and the run routes to the branch-creation
   sub-step — the pre-change behaviour — rather than substituting the default branch.
   Defaults-with-a-recorded-degradation rather than hard halt, because a hard halt would
   reintroduce a deadlock against any Stage invocation predating the fix.
2. **Branch-specific discovery.** The unpushed-commit range is taken against the resolved staging
   branch, with a `HEAD`-equals-branch assertion that doubles as the P-016 single-worktree gate.
   That resolution and its guards sit **above** the dirty/clean split, so they also cover the
   dirty path — the one on which the gate actually mutates the repository.
3. **Two verification arms.** The shipment arm keeps the authoritative `origin/main` manifest gate
   unrelaxed; a second arm keyed on `no-shipment` verifies commit **ancestry** in `origin/main`
   plus per-path readability there, over a **non-empty** path set — a concrete target requiring no
   shipment ID. The non-empty guard matters: an empty path loop combined with a trivially-true
   ancestry check would report success having verified nothing.

D7 (role isolation) is **unchanged and reinforced**: Stage creates, checks out, and commits its
artifact branch and reports it; the Orchestrator pushes, opens, merges, and verifies. Neither
gains an action the other's Role Boundary forbids.

### 10.4 D3 is NOT extended to cover this defect

Main-only discovery is neither of D3's two violation shapes — it is not a push to the default
branch and not a permission to commit on one. Folding it into the detector would mean opening the
closed construct set and re-deriving the fixture corpus through the whole A→C chain, which is
materially larger than the finding warrants. The regression is rejected instead by the
**already-existing** per-surface harness assertions. One consequence is recorded because it is
easy to get wrong: the corrected instruction must state its prohibition **descriptively** and must
not quote the forbidden range, or it self-matches the absence assertion and silently vacates it.

### 10.5 Net effect

Option 2 is unchanged. D1 is extended from "authorization surfaces" to "authorization **and
discovery**"; plan rev 2's no-no-shipment-clause position is withdrawn; D3, D5, and D7 are
unchanged. No decision is reversed and no new option is opened.

## 11. Revision 7 addendum — a verification that cannot fail is not a verification

Source: PR #54 current-HEAD review, cycle 6 — thread `PRRT_kwDOTPuhps6hrC0c` (comment
3994280334) on plan L626, and thread `PRRT_kwDOTPuhps6hrC0o` (comment 3994280351) on the session
memory record.

### 11.1 Withdrawn: "ancestry plus a per-path read proves the artifacts are on origin/main"

§10.3 decided the no-shipment arm as *commit ancestry* + *per-path `git show origin/main:{path}`*
over a path set defaulting to `.backlogit/ docs/plans/ docs/decisions/ docs/memory/`. That
conjunction was asserted to be a concrete verification target. It is **not**, and the claim is
withdrawn:

* `git show` on a **tree** exits **0** and prints a directory listing. Verified against this
  repository: `git show origin/main:docs/plans/` prints `tree origin/main:docs/plans/`, status 0.
* All four defaulted entries are **directories that already exist on `origin/main`** in any
  repository this gate runs in. The loop therefore succeeded over targets that are unconditionally
  present, reading **no file the Stage run wrote**.
* The paired ancestry check is *also* routinely satisfied in the degraded case, because
  `stage_head_commit` defaults to `git rev-parse HEAD` and a post-merge checkout's `HEAD` is
  commonly already an ancestor of `origin/main`.

Both conjuncts could hold with nothing from the Stage session present, so the arm could report a
**false pass** — the precise failure §10.3 set out to remove.

### 11.2 The generalized lesson, third instance

Rev 6 caught *"the loop can run zero times"* and added a non-empty guard. The surviving defect is
the sibling: *"the loop runs only over things that are always there."* Both are instances of one
shape already twice recorded in this deliberation — a check whose **pass** is not evidence of the
property it names. §8.2's self-deadlock, §10.1's empty-by-construction commit range, and now the
always-present path set are the same error in three costumes: **the artifact under test was never
bound to the run under test.**

The corrective principle, stated once so it generalizes: *a verification must name targets derived
from the specific run it verifies, and its target set must be capable of being empty or wrong.* A
target list that is identical for every possible run carries no information about any of them.

### 11.3 Decision — bind the verified set to the commit, structurally

`stage_artifact_paths` becomes a **non-empty set of concrete repository-relative file paths tied to
`stage_head_commit`**. When the handback supplies it, it is validated (file-ness, root containment,
set-equality with the commit's changed files); when it does not, it is **derived** from the commit
itself. The binding is structural rather than procedural, which is the whole point:
`git diff-tree -r` cannot emit a directory, and `git cat-file -t` must print exactly `blob` — so
the defect cannot recur through an author forgetting to check.

Three design choices carry the decision and each rejects a plausible alternative:

* **`cat-file -t` over `show`.** Both exit 0 on a tree; only `cat-file -t` reports the object
  **type**, which is the actual discriminator. Keeping `show` and "just checking the output isn't a
  listing" is a parse of human-readable text — fragile in exactly the way the gate is not allowed
  to be.
* **Set equality over subset.** A subset check rejects an invented path but accepts a handback
  that silently drops files down to one always-present entry — which reintroduces §11.1's defect
  through the reported-path branch.
* **The commit's own changed-file set over a branch range.** A range (`origin/main..
  {stage_head_commit}`) is **empty by construction** after the merge — the state the gate actually
  runs in — so it would collapse to the vacuous loop being removed. A commit's changed-file list is
  a property of the commit object and reads identically before and after the merge. The narrowing
  to the tip commit is sound because ancestry already covers earlier commits on the branch; the
  per-file check closes a different gap (files readable on the default branch), not the same one.
  **[Withdrawn at rev 8 — see §12.1. The last sentence is a category error: ancestry proves the
  *commit* landed, which is precisely the property §11.1 judged insufficient when it introduced the
  per-file check. The range-form rejection stands; the tip-only narrowing does not.]**

### 11.4 D3 is again NOT extended

An always-passing verification loop is neither D3 violation shape — it is not a push to the default
branch and not a permission to commit on one. The closed construct set and the fixture corpus stay
as they are (the §10.4 position, unchanged). Rejection is carried by the existing per-surface
harness assertion, which additionally asserts the **absence** of the `git show origin/main:{path}`
loop and of any directory-prefix default in the no-shipment arm.

### 11.5 Net effect

Option 2 is unchanged. D1's scope is unchanged from rev 6 (authorization **and** discovery); the
rev-6 *shape* of the discovery verification is corrected, not its existence. D3, D5, and D7 are
unchanged. No decision is reversed, no new option is opened, and no task, fixture, harness
function, or dependency edge is added.

---

## 12. Revision 8 addendum — a verification nobody can trigger, and a set that stops at the tip

**Source**: PR #54 **adversarial review** — anchor **GPT-5.6 Sol** with **GPT-5.4-mini**,
**Claude Sonnet 5**, and **Claude Opus 5**; **no route degradation**. Convened per the operator's
standing instruction (recorded at session memory §14.8) that a non-converging cycle escalates to an
adversarial round rather than a seventh self-directed one. All findings were classified
**same-contract completion** under P-021.

### 12.1 Withdrawn: "scoping the verified set to the tip commit loses nothing"

§11.3 argued that deriving the per-file check from `stage_head_commit`'s own diff was sufficient,
because commit **ancestry** already covered earlier commits on the branch. That argument is a
**category error**, and §11.1 had already refuted it two sections earlier:

| Property | What it proves | What it does not |
|---|---|---|
| `merge-base --is-ancestor {head} origin/main` | the **commit** reached the default branch | that any **file** is readable there |
| `cat-file -t origin/main:{path}` = `blob` | the **file** is readable there, as a file | anything about commits that did not change it |

§11.1 introduced the per-file check **precisely because ancestry is insufficient**. Applying that
check to the tip alone then left every artifact written in an earlier commit of the same session
resting on the insufficient property — the exact gap the check exists to close, relocated rather
than removed. A Stage session that writes its deliberation in commit 1 and its plan in commit 2 —
the ordinary shape; this very branch has many commits — had the deliberation verified by nothing
stronger than ancestry.

**The generalized lesson, fourth instance.** Rev 6 fixed *"the loop can run zero times"*. Rev 7
fixed *"the loop can run only over things that are always there"*. Rev 8 fixes *"the loop runs over
the right kind of thing, but not over all of them."* Each revision corrected the **shape** of a
verification while leaving its **extent** unexamined. The recurring failure is not carelessness
about any one predicate; it is checking that a verification *can fail* without checking that it
*covers everything it claims to cover*.

### 12.2 Withdrawn: "the no-shipment route is now executable end to end"

Rev 6 declared the no-shipment route executable once the Orchestrator could discover and verify the
branch. It is not, and rev 6 through rev 7 never tested the claim against the **producer**. The
installed `_stage.agent.md` makes `stage_outcome: no-shipment` **unreachable**:

* **Step 5.5** declares shipment assembly "MANDATORY — not optional" whenever the registry
  advertises `features.shipments: true`, and calls ending a session without a `shipment_id` a
  **P-005 violation**.
* **Step 6**'s pre-summary verification gate says that if no `shipment_id` exists, **HALT** and
  return to Step 5.5 — so the summary, which is where the handback line lives, is never reached.

A Stage run with nothing to ship therefore cannot terminate compliantly, cannot emit a handback,
and cannot set the outcome the Orchestrator's second arm is keyed on. Everything rev 6, rev 7 and
rev 8 built on the consumer side was gated on an outcome **no producer could emit**.

This is the **rev-2 self-deadlock lesson (§10.2) in its third instance**: rev 1 obliged Stage to
merge a PR it was forbidden to create; rev 6 obliged the Orchestrator to find a branch nothing told
it about; rev 8 finds a verification arm keyed on an outcome nothing can produce. In each case a
*consumer* obligation was specified without checking that a *producer* existed. The general
discipline this deliberation should have carried from §10.2 onward: **for every value a contract
consumes, name the surface that emits it, and confirm that surface is permitted to.**

### 12.3 Decision — bind the set to a commit *pair*, and make the terminal outcome reachable

Two decisions, both narrow, neither reopening Option 2 or D1–D7.

**(a) `stage_base_commit`.** The Orchestrator captures `git rev-parse HEAD` **immediately before it
invokes Stage** and **retains** it; the retained value is authoritative over any Stage-reported one,
which exists only so a *direct* Stage invocation still yields a usable base. The verified set
becomes the **two-tree** diff `{stage_base_commit} {stage_head_commit}`, bounded by the
`STAGE_ARTIFACT_ROOTS` pathspec. This keeps §11.3's post-merge-stability property — a two-tree diff
reads identically before and after the merge, so the range-form rejection **stands** — while
covering every commit the session made. Base→head ancestry is asserted, so the pair is a range
rather than two unrelated commits. Step 3a **preserves** the base across its commit; re-recording it
would collapse the range, and re-recording it to the pre-commit `HEAD` would discard the
already-committed unpushed commits that path exists to handle.

**(b) A reachable `no-shipment` terminal.** B3 scopes Step 5.5's mandatory rule to a harvest that
**produced items**, requires `shipment_id` in Step 6's gate **only** for
`stage_outcome: shipment`, and lets a reviewed, valid **empty** harvest emit the complete handback
and stop without routing to Ship.

**The constraint that makes (b) safe.** Step 5.5's existing guardrail — *do not assemble a shipment
if the harvest produced no items or produced items with unresolved P-003 violations; halt and
report* — is preserved **verbatim** and is exactly the discriminator between the two cases. P-003
lineage violations, harvest failures, and a missing required shipment after a **non-empty** harvest
continue to **halt** and are **never** recorded as `no-shipment`. Without that, this decision would
trade an unreachable verification for a **silent failure channel** — strictly worse than the defect
it fixes, and the reason the change is scoped to two clauses rather than to Step 5.5 as a whole.

### 12.4 D3 is again NOT extended

Neither an under-scoped verified set nor an unreachable outcome is a D3 violation shape: neither is
a push to the default branch, and neither is a permission to commit on one. The closed construct
set and the fixture corpus stay as they are (the §10.4 / §11.4 position, unchanged). Rejection is
carried by the existing per-surface harness assertions, which additionally assert the **absence** of
a single-commit derivation in the Orchestrator's no-shipment arm and the **presence** of the
reachable terminal in the Stage surface.

### 12.5 Net effect

Option 2 is unchanged. D1's scope is unchanged (authorization **and** discovery); rev 8 corrects
the **extent** of the discovery verification and supplies the **producer** the route always assumed.
D2, D3, D4, D5, D6 and D7 are unchanged. No decision is reversed, no new option is opened, and no
task, file, fixture, harness function, or dependency edge is added. Shipment `017-S` remains one
shipment of nine items. The only scope growth is two additional edit sites inside
`_stage.agent.md`, a file `018.009-T` already owns, which re-sizes that task S→M on volume alone.

---

## 13. Revision 9 addendum — a permission is not an instruction

**Source**: PR #54 current-HEAD review, cycle 7 — one **visible** Copilot finding (thread
`PRRT_kwDOTPuhps6hrbAk`, comment `3994429008`, on `018.009-T:22`), classified **same-contract
completion** under P-021.

### 13.1 Withdrawn: "correcting the Role Boundary and emitting the handback completes the Stage side"

Revisions 1–8 treated the Stage side of the route as finished once three things were true: P-010
**permitted** the branch (rev 3), the Role Boundary cell **granted** it at mutation time (rev 2–3),
and Step 6 **reported** what happened (rev 6–8). None of the three is an instruction to act. The
installed `_stage.agent.md` runs triage → grouping → learnings → deliberation → planning → review →
harvest → shipment assembly → stash archival → summary, and contains **no branch operation and no
commit operation at any step**; its Step Sequence Contract checklist — the file's own statement of
what a session MUST execute — names none either.

| What rev 1–8 established | What it does not establish |
|---|---|
| Stage **may** create, check out, and commit on the artifact branch | that any step does |
| Stage **must report** `stage_branch`, `stage_base_commit`, `stage_head_commit`, `stage_artifact_paths` | where those values come from |

So an agent satisfying AC-B3.1–AC-B3.6 exactly can finish a session **on the default branch**, with
every artifact written there — the `fdff9e4` shape this deliberation exists to eliminate — and then
emit a handback naming a branch that was never created and commits that were never made. The
Orchestrator's step-4 arm rejects that handback, which is the gate working; but the artifacts are
already on the default branch by then. **Detection after the fact is not prevention**, and §1's
objective is prevention.

### 13.2 The producer/consumer lesson, fourth instance — and its missing clause

§12.2 recorded the discipline this deliberation should carry: *for every value a contract consumes,
name the surface that emits it, and confirm that surface is **permitted** to.* Rev 9 shows the
clause is incomplete. `stage_branch` had a named emitting surface (`_stage.agent.md` Step 6) that
was unambiguously **permitted** to emit it — and the value still had no origin, because permission
and emission are both satisfied by a surface that never **acts**.

| Instance | Consumer obligation | What was missing |
|---|---|---|
| rev 1 | Stage merges the staging PR | Stage was **forbidden** to create one |
| rev 6 | Orchestrator finds the Stage branch | nothing **announced** it |
| rev 8 | Orchestrator verifies `no-shipment` | nothing could **emit** it |
| **rev 9** | Orchestrator consumes `stage_head_commit` | nothing **produced** it |

The extended discipline: **for every value a contract consumes, name the surface that emits it,
confirm that surface is permitted to, confirm it is *instructed* to, and confirm the instruction is
*ordered* relative to the state the value describes.** The last clause is not decoration — a branch
gate placed after harvest and a commit step placed after the summary both satisfy "instructed" and
both still produce values describing a state that does not exist.

### 13.3 Decision — two operative steps, ordered, in the file B3 already owns

Narrow, and reopening neither Option 2 nor D1–D7.

**(a) A pre-mutation branch gate.** A new Stage step between learnings retrieval and deliberation —
a **fixed** position, which is also the first point at which a stable scope slug is derivable for
both intake shapes. Every Stage mutation that would otherwise occur earlier (the Step 1
deferred-expansion duplicate archival, the session's first memory checkpoint) is **deferred until
after** it, so nothing writes before the gate. The alternative considered and **rejected** was to
let the gate fire early against a provisional slug: that contradicts the Step Sequence Contract's
own execute-in-order semantics, leaves the real position unassertable, and re-admits the
default-branch write it exists to prevent. The step derives the slug, records `stage_base_commit`
(echoed from
the Orchestrator in a pipeline invocation, captured from `HEAD` **before** branch creation in a
direct invocation), **verifies and uses** an Orchestrator-supplied branch or **creates and checks
out** its own, refuses to write anything while `HEAD` is the default branch, and switches the single
existing worktree rather than adding one (P-016 unchanged).

**(b) A post-mutation artifact commit.** A new Stage step between consumed-stash archival and the
summary. It asserts `HEAD`, stages only the four `STAGE_ARTIFACT_ROOTS`, commits conventionally,
sets `stage_head_commit` from the **resulting** `HEAD`, derives the aggregate path set over the
preserved base, and **stops** — Stage neither pushes nor touches a pull request, so D7's role
isolation and P-010's standing PR prohibition are untouched. Steps 3–4 of §6.2's route table remain
the Orchestrator's or the operator's, exactly as before.

**(c) The naming dependency this exposed.** `chore/stage-{shipment_id}` is **underivable** once the
branch must exist before the first artifact write: the shipment is created by Stage's *last*
mutation. Following the rev-8 wording literally therefore pushed an executor toward deferring the
branch past the very writes it protects — a second route back to `fdff9e4`. The form becomes
`chore/stage-{scope-slug}`; a shipment ID stays a **permitted** slug rather than being deleted, so
an Orchestrator already holding one may still use it.

**What makes (a) and (b) safe rather than authority-expanding.** Neither step grants Stage anything
P-010 and the Role Boundary did not already grant at rev 3. They convert a standing permission into
a sequenced obligation. No push, no PR, no merge, no source/test/config write, no second worktree,
no new policy ID.

### 13.4 D3 is again NOT extended

A missing operative step is neither D3 violation shape — it is not a push to the default branch and
not a permission to commit on one — so the closed construct set and the fixture corpus stay as they
are (the §10.4 / §11.4 / §12.4 position, unchanged). Rejection is carried by the existing
per-surface harness assertion over `_stage.agent.md`, extended to compare **checklist indices and
heading positions** rather than to search for text. That structural form is required, not stylistic:
the file names its own steps in its Step 6 gate, so a name search would pass with the step sections
deleted — the self-matching failure already recorded in
`docs/compound/2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md`.

### 13.5 Net effect

Option 2 is unchanged. D1's scope is unchanged (authorization **and** discovery); rev 9 supplies the
**execution** the authorization always presupposed. D2, D3, D4, D5, D6 and D7 are unchanged — D7 in
particular is *reinforced*, since the commit step ends precisely where Stage's role ends. No
decision is reversed, no new option is opened, and no task, file, fixture, harness function,
shipment, or dependency edge is added. Shipment `017-S` remains one shipment of nine items. The only
scope growth is four further edit sites inside `_stage.agent.md`, a file `018.009-T` already owns,
which re-sizes that task M→L on volume alone.
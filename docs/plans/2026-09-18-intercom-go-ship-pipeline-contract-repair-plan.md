---
title: "Implementation Plan — Ship pipeline contract repair (S-1, S-2)"
date: 2026-09-18
status: plan-review attempt 1 FAIL (2× P0) → revised; attempt 2 ADVISORY (0 P0, 2 P1) → P1s applied in revision 4
agent: Stage
source_document: docs/decisions/2026-09-18-intercom-go-ship-contract-and-gate-reliability-deliberation.md
secondary_source: docs/decisions/2026-09-18-intercom-go-shipment-claim-reversibility-deliberation.md
governs: two release units — closure/reconcile skill contract repair, and Ship agent execution-contract repair
bound_snapshot: 14d44e3c1f28b321e8db24ab8cafcba346996133
stage_branch: chore/stage-ship-pipeline-contract-repair
requires_plan_hardening: yes
revision: 4
---

<!-- plan-review-attempt: 2 -->
<!-- attempt 2 verdict: ADVISORY. Findings applied in revision 4; see §8. No attempt 3 submitted. -->

> **Revision 3 — corrections from plan-review attempt 1 (FAIL, 2× P0, 5× P1).** Every finding is
> applied below and summarized in §7. The two P0s were: (1) Problem A misread `:442` as an
> unspecified *evidence format* when `:440` is the *command slot of quality gate #2*; (2) AC-2.3 and
> AC-2.4 were mutually unsatisfiable because `:830` conditionally authorizes the direct cascade call.

**Requires plan hardening: yes** — this plan modifies the *installed execution boundary*
(`.github/agents/_ship.agent.md` and two skill contracts). In this workspace the installed markdown
**is** the contract, so an error here mis-executes every future shipment. See §5.

## 1. Objective

Repair four measured defects in the controls that gate every shipment, so that subsequent units in
this cycle close against a trustworthy verification substrate.

## 2. Unit 1 — Closure and reconcile skill contract repair (S-1)

**Problem A.** `operational-closure/SKILL.md:23` documents the artifact form
`docs/closure/{YYYY-MM-DD}-{slug}-closure.md`. All 19 `*-post-merge-closure.md` artifacts in
`docs/closure/` use `{shipment_id}-{feature_id}-post-merge-closure.md`. The documented form matches
**zero** of them. This already blocked 013-S with `PREDECESSOR_CLOSURE_INCOMPLETE`.

*Causal-claim scope (rev 4).* The 013-S claim is inherited from stash `4A01C53E` and is **not
re-measured here**. This unit changes documentation and adds a test; it does **not** alter external
gate behaviour and does not by itself remove that failure mode (see H-7).

*Scoping correction (rev 3, was P2).* `:23` is the generic **Output** line covering all three
declared skill modes (`pre-merge`, `post-merge`, `post-deploy`). Only the **post-merge** form is
wrong. `docs/closure/` legitimately also holds ~14 `{YYYY-MM-DD}-{slug}-…` artifacts
(adversarial-review, runtime-verification) that are **not** post-merge closures and must not be
disturbed.

*Premise correction (rev 3, was P1).* Revision 2 claimed a test could assert parity with "the
topology gate's glob". The gate is an **external `autoharness` binary**; the glob appears nowhere in
this repository. No in-repo test can observe it. The acceptance surface is therefore re-scoped to
**document↔artifact conformance**, which is genuinely falsifiable, and the external glob is recorded
as a stated assumption rather than a tested property.

**Problem B.** `shipment-reconcile/SKILL.md` cites `src/autoharness/gates/shipment_closure.py`
(`:503`, `:1168`) and `src/autoharness/cli.py` (`:1043`). No `src/autoharness/` tree exists here.
`:503` is *already* conditionally worded; `:1043` and `:1168` are not.

### Tasks

| ID | Task | Scope | Size | Cx |
|---|---|---|---|---|
| U1-T1 | Correct the documented **post-merge** closure filename convention | `operational-closure/SKILL.md`. Give the post-merge mode its own explicit Output form `docs/closure/{shipment_id}-{feature_id}-post-merge-closure.md`. Leave the pre-merge and post-deploy Output forms intact and still documented. | S | low |
| U1-T2 | Add a document↔artifact conformance test | One new test that (i) **parses** the post-merge filename form out of `operational-closure/SKILL.md` — it must not hardcode the pattern — and (ii) asserts every `docs/closure/*-post-merge-closure.md` artifact matches the parsed form. **Parse rule:** take the post-merge-specific Output form if one exists; otherwise fall back to the generic Output line in the mode-independent `## Output` section (pre-repair there is only the generic `:23`; post-repair both exist, so the rule must be stated to be reproducible). Artifacts not matching `*-post-merge-closure.md` are out of the test's universe. | S | medium |
| U1-T3 | Make the two unconditional `src/autoharness/**` citations non-authoritative | `shipment-reconcile/SKILL.md`. Required post-state, per citation: `:1043` and `:1168` each either removed, or rewritten to state that the installed skill markdown is authoritative and the path is an upstream reference that may be absent. `:503` is already conditional and is left unchanged. | S | low |

**Red phase (genuine).** U1-T2 parses the form *from the file under repair*. Pre-change it parses
`{YYYY-MM-DD}-{slug}-closure.md`, which matches none of the 19 artifacts → **red**. Post-change it
parses the corrected form, which matches all 19 → **green**. The tautological variant (hardcoding
the pattern) is explicitly forbidden by U1-T2's scope.

### Acceptance criteria

* **AC-1.1** The **post-merge** mode of `operational-closure/SKILL.md` documents exactly one output
  form, and it is `docs/closure/{shipment_id}-{feature_id}-post-merge-closure.md`.
* **AC-1.2** The pre-merge and post-deploy output forms remain documented and semantically unchanged.
* **AC-1.3** U1-T2 fails against the pre-repair `SKILL.md` and passes after, with the form read from
  the file rather than hardcoded.
* **AC-1.4** All 19 existing `*-post-merge-closure.md` artifacts satisfy the documented form. **No
  artifact is renamed by this unit**; a non-conformer is reported, not fixed.
* **AC-1.5** Non-post-merge artifacts in `docs/closure/` are untouched and outside the test universe.
* **AC-1.6** `shipment-reconcile/SKILL.md:1043` and `:1168` no longer assert an unconditional
  dependency on a path outside this repository. `:503` is byte-unchanged.
* **AC-1.7** No change to `autoharness` gate code, which is external and out of scope.

## 3. Unit 2 — Ship agent execution-contract repair (S-2)

### 3.1 Problem A — the empty Format gate command *(re-derived in rev 3)*

Measured verbatim at `_ship.agent.md:435–443`:

```text
435 : #### Step 4.3: Quality Gates
439 : 1. **Lint**: `go vet ./...`
440 : 2. **Format**: ``
441 : 3. **Full Test Suite**: `go test ./...`
```

`:440` is the **command slot of quality gate #2**, sibling to the lint and full-suite commands — not
an evidence-format declaration. The correct repair is to supply the repository's canonical
formatting command. Revision 2's "derive an evidence format from run logs" framing was wrong, and
its sanctioned fallback ("delete the `**Format**:` line") would have **deleted a quality gate** — a
behavioural regression. Both are withdrawn.

**026-F collision resolved, not merely asserted.** Blocked feature 026-F declares *"Step 4.3
full-suite text unmodified (anti-regression pin)"* as a test scenario and non-goal. That pin is on
gate **#3** at `:441`. This unit edits gate **#2** at `:440` only. AC-2.2 makes `:441` byte-equality
an acceptance criterion, so this unit **satisfies** 026-F's pin rather than colliding with it.

### 3.2 Problem B — the unexecutable CASCADE route

`:832–833` mandates carrying `CLASSIFICATION_BINDING` into the close call. `:861–869` routes the
CASCADE close through a direct `backlogit_ship_shipment` call, which accepts only
`shipment_id`/`sha`/`message`/`author` on both surfaces. The binding is executable only via
`shipment-reconcile mode:safe-close` (adopted option O1-a).

**Contradiction resolved (rev 3, was P0).** `:830` reads *"**Do NOT call `backlogit shipment ship` /
`backlogit_ship_shipment`** unless `mode: classify-close-path` has already returned
`CLOSE_PATH_VERDICT: CASCADE`"* — a **conditional authorization** of the direct call. Revision 2
simultaneously froze `:827–832` and forbade any such direction file-wide; those cannot both hold.
Resolution: `:830` is **explicitly in scope** for edit; the binding mandate at `:832–833` and the
"only the named close path" rule at `:834–835` are frozen.

**Edit surface corrected (rev 3, was P2).** The CASCADE branch runs to `:869`, not `:864`; `:865–869`
carry the requeue/detach rationale. The scope below names the true range.

### Tasks

| ID | Task | Scope | Size | Cx |
|---|---|---|---|---|
| U2-T1 | Fill the Step 4.3 **Format** gate command slot | `_ship.agent.md:440` only. Supply the **non-mutating failing check** `ci.yml`'s lint job already gates on: **capture the output of `gofmt -l .` and fail (exit non-zero) when that output is non-empty**, matching `ci.yml:303–310`. A **bare `gofmt -l .`** does **not** satisfy this task: it only prints unformatted paths and still exits 0, leaving the Format quality gate unable to fail. A **mutating** variant (`go fmt ./...`, `gofmt -w`) does **not** satisfy this task either: installing a writing command into a verification gate is a defect. Gates #1 (`:439`) and #3 (`:441`) byte-unchanged. Deleting the `**Format**:` line is **not** an option. | XS | trivial |
| U2-T2 | Re-point the CASCADE branch at `shipment-reconcile mode:safe-close` | `_ship.agent.md:830` and `:861–869`. Rewrite `:830`'s conditional authorization so the CASCADE verdict routes to `mode:safe-close`, **explicitly naming safe-close as the close path a `CLOSE_PATH_VERDICT: CASCADE` designates** and carrying `CLASSIFICATION_BINDING` into it; replace the direct-cascade branch at `:861–869` with the safe-close route, preserving the requeue/detach semantics at `:865–869`. The replacement wording must be **pinned in the plan before the edit** (see §3.3). | M | medium |
| U2-T3 | Contract test asserting no binding-incapable CASCADE route remains | One new contract test whose PRESENT/ABSENT needles are cut **verbatim** from §3.3's pinned replacement wording — never authored from prose or intent (compound rule, `cross-artifact-contract-closure-requires-every-surface-2026-09-13`). **Single red→green assertion only**; revision 2's green-on-arrival second assertion is deleted. | S | medium |

**Red phase (genuine).** U2-T3 fails against the pre-repair file, where `:830` authorizes and
`:861–869` directs the direct cascade call. It passes only after U2-T2. U2-T1 is a one-line
mechanical fill with no test and **is declared characterization-free**: AC-2.1 and AC-2.2 are
inspection criteria, not assertions, and this plan does not claim a red phase for it.

### Acceptance criteria

* **AC-2.1** `_ship.agent.md:440` carries a **non-mutating** `gofmt` check whose failure condition is
  **non-empty `gofmt -l .` output** (behavioural parity with the `ci.yml:303–310` gate: unformatted
  files produce a non-zero exit). A bare `gofmt -l .`, which exits 0 even when it lists unformatted
  files, does **not** satisfy this criterion. No empty command slot remains in Step 4.3, and no
  mutating formatter is installed into a verification gate.
* **AC-2.2** `_ship.agent.md:439` and `:441` are **content-identical** to the bound snapshot. This
  is the criterion that preserves blocked 026-F's Step 4.3 full-suite anti-regression pin.
* **AC-2.3** No direction to invoke `backlogit shipment ship` / `backlogit_ship_shipment` for a
  CASCADE verdict remains anywhere in `_ship.agent.md`, **including `:830`**.
* **AC-2.4** The `CLASSIFICATION_BINDING` mandate at `:832–833` and the "only the named close path"
  rule at `:834–836` — the full clause, which ends at `:836` ("…violation and HALTs") — are
  content-identical to the bound snapshot.
* **AC-2.5** The requeue/detach semantics currently at `:865–869` survive the rewrite in substance.
* **AC-2.6** U2-T3 fails on the pre-repair file and passes on the post-repair file.
* **AC-2.7** `_ship.agent.md:265`, `:322`, `:338–358` and `:362–367` — the clause sites blocked
  feature 026-F's plan claims — are **content-identical** to the bound snapshot.
* **AC-2.8** *(rev 4, closes attempt-2 P1)* The rewritten `:830` **explicitly names**
  `shipment-reconcile mode: safe-close` (its Cascade Close Sub-Procedure) as the close path a
  `CLOSE_PATH_VERDICT: CASCADE` designates, and states that `CLASSIFICATION_BINDING` is carried into
  it. Without this, the frozen `:832–836` rule ("Ship invokes **only** the close path the returned
  verdict names, and **never a path absent from that verdict**") becomes false of the post-repair
  file — the same defect class as attempt-1's P0-2, merely under-constrained instead of
  unsatisfiable.
* **AC-2.9** *(rev 4)* `go test ./tests/integration/` — which loads both `_ship.agent.md` and
  `shipment-reconcile/SKILL.md` and asserts 33 verbatim contract rows — is green **before and
  after** U2-T2 and U1-T3. The rows falling inside the rewritten `:861–869` span are enumerated in
  the change record.

> **Line-index note (rev 4).** AC-2.2, AC-2.4 and AC-2.7 are stated as **content**-equality of the
> quoted clause text, not line-index equality: the `:440` fill and the `:861–869` rewrite both shift
> downstream indices. §6 already verifies them in this form via `git diff`.

### 3.3 Pinned replacement wording (authored before the edit)

Per the binding compound rule, U2-T3's needles are derived mechanically from this block; backticks,
emphasis markers and punctuation are part of the literal. The executor pins the exact replacement
text for `:830` and `:861–869` **here, in this plan**, before touching the file, and the reviewer
checks needle↔wording correspondence rather than intent. A needle that does not appear verbatim in
this block is inadmissible.

*Contract-surface matrix (required before the first edit, per compound entry
`cross-artifact-contract-closure-requires-every-surface-2026-09-13`, which is 026-F unblock step
O-3):* enumerate for **every** surface — verdict producer (`classify-close-path`), consumer
(`_ship.agent.md` CASCADE clauses), authority grant (P-010), refusal sink (BLOCK / safe-close),
pinned needles (`tests/integration/`), and tooling probes (`backlogit shipment ship` contract) — the
columns *Current evidence / Intended change / Executable verification / Depends on*. Any empty cell
is an open edge and blocks the edit. A producer-scoped or single-file measurement may **not** be
cited as contract closure.

## 4. Dependencies

* **S-1 blocks S-2.** The closure-evidence naming convention must be correct before the CASCADE
  route that produces closure evidence is re-pointed; otherwise S-2 lands a corrected route that
  still emits gate-invisible artifacts.
* **S-1 blocks S-9** (`docs/plans/2026-09-18-intercom-go-checker-test-docs-hygiene-plan.md`, Unit 3)
  — those docs describe this convention.
* **Related, not blocking**: blocked feature 026-F. AC-2.2 and AC-2.7 make the non-collision
  mechanically checkable rather than asserted.

## 5. Plan Hardening

**H-1 — Blast radius.** Both units edit files that *are* the execution contract. Mitigation: each
task names an explicit line range; AC-2.2 and AC-2.7 pin every clause site this plan must not touch,
as byte-equality checks.

**H-2 — Green-on-arrival acceptance surfaces.** The dominant failure mode for agent-prose repairs
(026-F blockers C3/C4). Mitigation: U1-T2 must parse the form from the file, not hardcode it;
U2-T3's green-on-arrival second assertion is deleted; U2-T1 is declared test-free rather than given
a decorative assertion.

**H-3 — Wrong-repair risk.** Revision 2 prescribed the wrong repair for `:440` from a misread line
number. Mitigation: §3.1 quotes the measured lines verbatim so the repair target is unambiguous, and
the delete-the-gate fallback is withdrawn.

**H-4 — Frozen-vs-forbidden contradiction.** Mitigation: §3.2 states exactly which lines are in
scope for edit (`:830`, `:861–869`) and which are frozen (`:832–835`), replacing revision 2's
overlapping ranges.

**H-5 — Authority boundary.** The claim-reversibility question (`3F546E63`) is excluded and carried
as operator determination O-5.

**H-6 — Renaming existing artifacts.** AC-1.4 verifies conformance and explicitly forbids renaming.

**H-7 — Unverifiable external gate.** The topology gate is an external binary. Mitigation: the
acceptance surface asserts document↔artifact conformance only; the external glob is a documented
assumption, and U1-T3 exists precisely to stop this repo asserting hard dependencies on absent
upstream paths.

## 6. Verification

Unit 1: run U1-T2 before and after U1-T1; confirm red→green. Confirm the parsed-not-hardcoded
property by inspection. Grep `shipment-reconcile/SKILL.md` for `src/autoharness`; confirm `:503`
unchanged and `:1043`/`:1168` conditional or removed.

Unit 2: `git diff` the bound snapshot and confirm `:439`, `:441`, `:832–836`, `:265`, `:322`,
`:338–358`, `:362–367` are untouched in content. Run U2-T3 before and after U2-T2; confirm
red→green. **Run `go test ./tests/integration/` before and after U2-T2 and U1-T3** — the 33-row
contract suite reads both edited files and must be green on both sides (AC-2.9).

## 7. Review findings applied (attempt 2 → revision 4)

| Finding | Sev | Resolution |
|---|---|---|
| `:832–835` freeze vs. CASCADE→safe-close reroute leaves the frozen "only the named close path" rule false — same defect class as attempt-1's P0-2, under-constrained | P1 | **AC-2.8** added; U2-T2 scope now mandates naming safe-close as the designated CASCADE close path |
| Frontmatter still pre-stamped its own verdict | P1 | Status records the actual attempt-2 verdict (ADVISORY) and that revision 4 applies its findings |
| "canonical formatting command" not single-valued; `go fmt ./...` **mutates** | P2 | U2-T1 and AC-2.1 pin a captured-output `gofmt -l .` check that fails on non-empty output; bare (non-failing) and mutating variants explicitly excluded |
| U2-T3's oracle not mechanically decidable over markdown | P2 | **§3.3** pins the replacement wording before the edit; needles derived verbatim from it |
| §6 omitted the existing 33-row contract suite that reads both edited files | P2 | AC-2.9 + §6 |
| U1-T2's parse anchor unspecified pre/post repair | P3 | Parse rule stated in U1-T2 |
| Line-index byte-equality vs. index-shifting edits; `:834–835` bisects a clause ending at `:836` | P3 | ACs restated as content-equality; AC-2.4 extended to `:832–836` |
| §2 overstated that Unit 1 removes the 013-S failure mode | P3 | Scope qualifier added to §2 Problem A |
| D-A's "singular binding route" assertion silently dropped | P3 | Recorded below as an intentional deviation |

*Intentional deviation from the reversibility deliberation:* **D-A's "a contract test asserts the
binding route is singular" is deliberately not implemented** — it was green-on-arrival at
`14d44e3c` and would be a decorative test. No work change; recorded so the plan and its source
deliberation do not silently disagree.

## 8. Review findings applied (attempt 1 → revision 3)

| Finding | Severity | Resolution |
|---|---|---|
| `:442` misread; `:440` is a gate **command** slot, not an evidence format | P0 | §3.1 re-derived from verbatim source; U2-T1 rewritten; delete-the-line fallback withdrawn |
| AC-2.3 ⊥ AC-2.4 — `:830` conditionally authorizes the direct call | P0 | `:830` moved into scope; freeze narrowed to `:832–835` (AC-2.4) |
| 026-F pins Step 4.3 full-suite text; fence drawn in the wrong place | P1 | AC-2.2 pins `:441` byte-equality; AC-2.7 pins all four 026-F clause sites |
| U2-T3 assertion (b) green-on-arrival | P1 | Assertion deleted; single red→green assertion retained |
| Topology-gate glob unverifiable in-repo | P1 | Re-scoped to document↔artifact conformance; external glob recorded as assumption (H-7) |
| U2-T1 had no red phase behind a disjunctive AC | P1 | Declared test-free; AC-2.1 made single-valued (fill, never delete) |
| Frontmatter pre-declared `PASS` | P1 | Status now records the actual gate outcome |
| CASCADE branch extends to `:869`, not `:864` | P2 | Scope corrected; AC-2.5 preserves `:865–869` semantics |
| AC-1.4 partially green-on-arrival; `:503` already conditional | P2 | Per-citation post-state enumerated (AC-1.6); `:503` frozen |
| `:23` is the generic Output for three modes | P2 | AC-1.1 scoped to post-merge; AC-1.2 protects the other two |
| `:827–832` mis-delimited | P3 | Ranges corrected throughout §3.2 |
| U1-T2 could hardcode the pattern | P3 | Forbidden in scope; AC-1.3 requires parsed-from-file |

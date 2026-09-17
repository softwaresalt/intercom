---
title: "Implementation Plan — 017-S closure (Stage artifact branch/PR policy gap correction)"
date: 2026-09-16
revision: 9
status: draft
agent: Stage
shipment: 017-S
covering_feature: 018-F
source_decision: docs/decisions/2026-09-13-intercom-go-a-only-simplification-decision.md
prerequisite_plan: docs/plans/2026-09-13-intercom-go-ship-feature-completion-decided-plan.md
canonical_contract: docs/plans/2026-09-16-017-S-closure-canonical-contract.md
prerequisite_shipment: 021-S
requires_plan_hardening: yes
---

# Implementation Plan — `017-S` closure

This plan supplies the execution site, preflight gates, baseline and rollback that prerequisite
plan §11.1 requires before `PA-017-CASCADE` may be exercised. It is a remaster of
`docs/plans/2026-09-16-017-S-closure-canonical-contract.md`; where the two disagree, **the
contract governs and this plan is defective**.

**Executor:** Ship. **Author and gatekeeper:** Stage. Stage does not claim, implement, merge, or
invoke the cascade.

---

## 1. Scope

| Item | Value |
|---|---|
| Release unit | shipment `017-S` |
| Covering feature | `018-F` |
| Manifest | **frozen at 13 members**: `018-F`, `018.008-T`, `018.001-T`, `018.001.001-ST`, `018.001.002-ST`, `018.001.003-ST`, `018.002-T`, `018.002.001-ST`, `018.002.002-ST`, `018.002.003-ST`, `018.003-T`, `018.003.001-ST`, `018.003.002-ST` |
| Live members | `018-F` (`queued`), `018.008-T` (`queued`) |
| Pre-archived members | the other **11**, exempt under prerequisite §3.3 |
| Source decision | `docs/decisions/2026-09-13-intercom-go-a-only-simplification-decision.md` (`SIMPLIFICATION_VALID`, option O-2) |
| Prerequisite plan | `docs/plans/2026-09-13-intercom-go-ship-feature-completion-decided-plan.md` — revision 13, **FROZEN** |

This plan inherits prerequisite §3.1, §3.2 and §3.3 unchanged. It creates no new verdict, token,
classifier or policy surface.

## 2. Eligibility grounds

| Ground | State | Re-verified at |
|---|---|---|
| 1 — `021-S` shipped | discharged (`archived` + `archived_status: shipped`, `74330d3`, merge `f2d4cf9`, PR #56) | **PF-9** |
| 2 — parent-completion capability live | discharged (`0cb4a42`) | **PF-10** |
| 3 — governing closure plan exists | discharged by this document **on its review PASS and presence on `origin/main`** | **PF-2** |
| 4 — root-included cascade probe committed | discharged (Probe 25 @ `e36d853`) | **PF-7** |

No ground is waived by assertion. Every Stage-time measurement below is **re-measured by Ship at
invocation** and is never inherited.

## 3. Measured ground truth

| ID | Measurement |
|---|---|
| **M-1** | Manifest = **13** members (enumerated in §1). |
| **M-2** | `017-S` dependencies = `[021-S]` (`type: blocks`), satisfied. |
| **M-3** | Live descendant graph of `018-F` ∪ `{018-F}` is **set-equal** to the manifest. |
| **M-4** | Exactly 2 members live; the other 11 declare `status: archived`, are archive-located, carry well-formed `archived_from` and present `archived_status` (`queued`), and are **exactly** the §3.3 allowlist. |
| **M-5** | Validated linked-deliberation set for all 13 = **EMPTY** (three independent methods). `COUNT(*)=2` deliberations exist (`001-DL`, `002-DL`) — they belong to the `021-S` lineage, link to no member here, and are unreachable by the cascade. |
| **M-6** | `018.008-T` carries one outgoing link `related_to → 019.007-T`. **Not a deliberation link.** `019.007-T` is `status: blocked`, `parent_id: 019-F`, **outside the manifest**, outside the cascade's archival scope. |
| **M-7** | `backlogit doctor` ⇒ `No issues found.`, exit 0. |
| **M-8** | Topology gate `--phase pre_claim` ⇒ exit 0. |
| **M-9** | The claim over this exact shape (13 members, root included, 11 genuinely archived) transitions the root feature `queued → active` and the live task `queued → active`, and leaves **all 11 archived members unchanged**. Evidence: Probe 26. |
| **M-10** | **The claim is a one-way door.** `backlogit shipment` = `add|claim|create|get|list|return-blocked|ship`. No unclaim/release/return-to-queued. Exits from `active` are only `shipped` and `abandoned`. |
| **M-11** | `return-blocked` returns a **single item** ⇒ breaks the 13-member manifest. Forbidden here. |
| **M-12** | Generic `backlogit move` **refuses shipment status writes, exit 9**. Never the close path. |
| **M-13** | `backlogit move <id> --status archived` is **not** an archive operation (silent no-op). The archive operation is `backlogit archive <id>`. No step here archives via `move`. |
| **M-14** | Cascade close command: `backlogit shipment ship <id> --sha <s> --message <m> --author <a>`. Flags exactly `--sha`, `--message`, `--author`. |
| **M-15** | **No binding parameter on either close surface** — CLI as M-14; registry `ship_shipment.params` = `shipment_id`, `sha`, `message`, `author`. Only `shipment-reconcile` `mode: safe-close` accepts `classification_binding`. |
| **M-16** | `backlogit link list` takes a **positional** ID; the `--id` form no longer resolves on 1.10.1. |

### 3.1 External-interface measurements (attempt-5 remediation)

Measured against the installed workspace. Where the `021-S` closure of 2026-09-16 supplies
empirical evidence, it is cited: cascade commit `17f891f`, closure merge `f2d4cf9` (PR #56).

| ID | Measurement | Evidence |
|---|---|---|
| **M-17** | **`status: done` is ARCHIVE-ROUTED.** `.backlogit/registry.yaml` routes `done`/`accepted`/`rejected`/`archived` → `archive`, and `queued`/`active`/`blocked`/`review` → `queue`. A `→ done` transition **relocates the record file**, appearing in a diff as a **rename** `.backlogit/{queue => archive}/<id>.md`. So `018.008-T → done` at S-9 moves the record out of `queue/`. | `registry.yaml`; `f2d4cf9` shows `{queue => archive}/021-S.md` and `{queue => archive}/022-F.md` |
| **M-18** | **`.backlogit/reconcile/` is version-controlled, NOT gitignored.** `.gitignore` excludes only `hooks_queue.jsonl`, `backlogit.db*`, `checkpoints/`, `runtime/`, `.telemetry-checkpoint.json`. Committed reports already exist for `006-S`, `008-S`, `009-S`, `012-S`, `021-S`. | `git check-ignore .backlogit/reconcile/` → exit 1; `git ls-files` |
| **M-19** | `shipment-reconcile` persists `.backlogit/reconcile/{shipment_id}-{mode}-{timestamp}.md` for the modes that persist. **`classify-close-path` writes NO report under ANY invocation** — the installed skill states the simplest conforming implementation omits optional persistence entirely, "which this skill does". Its report is therefore never present and must never be required (§6.1 D5a). The measured filenames are `017-S-pre-*.md`, `017-S-cascade-close-*.md`, `017-S-post-*.md`; note the cascade file is named `cascade-close`, not the `safe-close` mode token. | `SKILL.md` L65, L321, L432; `ls .backlogit/reconcile/` |
| **M-20** | **Measured cascade-commit emission set:** backlog record changes plus reconcile reports, and nothing else. `17f891f` contained exactly 6 files — 2 queue→archive renames, 1 archive-side modify, 3 reconcile reports. | `git show --stat 17f891f` |
| **M-21** | **Measured closure-PR emission set:** the cascade commit plus `docs/closure/{shipment}-{feature}-post-merge-closure.md`, `docs/memory/*.md`, `docs/compound/*.md`. `f2d4cf9` totalled 17 files. `docs/closure/` is an established convention holding records from `001-S` onward. | `git show --stat f2d4cf9`; `ls docs/closure/` |
| **M-22** | **P-005 has THREE required actions, not two:** (1) broadcast the violation with policy ID and one-line summary; (2) **include violation details in PR descriptions as compliance annotations**; (3) record the violation in memory checkpoints. All three are mandatory. No standalone telemetry *file* is required; `.backlogit/.telemetry-checkpoint.json` is machine-local and gitignored. The broadcast surface must be probed (P-012) with a written fallback (P-017) so the record cannot silently no-op. | `workflow-policies.md` §P-005, §P-012, §P-017 |
| **M-23** | `mode: classify-close-path` **exists** in the installed mode enum (`pre \| post \| safe-close \| classify-close-path \| detect-mixed-role`). | `SKILL.md` L56 |
| **M-24** | **P-009 is literal** and empirically observed: `f2d4cf9` has **exactly two parents** — parent1 `74330d3` (mainline), parent2 `f88d413` (reviewed PR head). | §P-009; `git rev-list --parents -n 1 f2d4cf9` |
| **M-25** | **The cascade commit is single-parent** (`17f891f` parent count 1) ⇒ a plain `git revert` target; `-m 1` must not be used on it. | `git rev-list --parents -n 1 17f891f` |
| **M-26** | **P-010 forbids Ship committing or pushing directly to `main`.** Any revert of a landed merge must be branch-based and land via its own reviewed merge-commit PR. | §P-010 |
| **M-27** | Merge gates in force: **P-014** (readiness, verified in order), **P-018** (Copilot review completion when enabled), **P-009** (merge commit). | §§P-014, P-018, P-009 |
| **M-28** | **Worktree state is single** (`git worktree list` → one entry), satisfying P-016. P-011 requires branch-before-mutation. | `git worktree list` |
| **M-29** | **P-002's ordering is structurally unsatisfiable here.** P-002 gates task claiming on a `harness-ready` label, but by M-9 the *shipment* claim auto-activates `018.008-T`, so the task claim precedes any harness. The `harness-architect` skill **is** installed — the producer exists, the ordering does not. Disposition in §4.1. | §P-002; Probe 26 `C2`; `.github/skills/harness-architect/` |

## 4. Close-path authority

Installed Ship requires the returned `CLASSIFICATION_BINDING` to be carried into the close call.
By **M-15** the cascade operation accepts no binding on either surface. The only surface that
accepts one is `mode: safe-close`, which routes a bound `CASCADE` verdict into the Cascade Close
Sub-Procedure — and that Sub-Procedure invokes the same `backlogit shipment ship` operation under
its `returned_ids`, two-set `allowed_ids`/`required_ids`, and `parent_id` gates.

**Therefore `mode: safe-close` carrying a `CASCADE` binding is the only executable realization of
installed Ship's own mandate — not an alternative to it.** A direct, unbound
`backlogit shipment ship` on this route is a **HALT**: it discards the mandated binding and
bypasses the Sub-Procedure's gates.

**No prerequisite Ship release unit is required.** Harmonizing the installed Ship prose is a
clarity improvement, captured out-of-scope as stash entry **`11B75632`** (§10).

**PF-11 makes this falsifiable** rather than asserted.

### 4.1 P-002 disclosed deviation (requires explicit operator authorization)

By **M-29** the shipment claim at S-3 auto-activates `018.008-T`, so P-002's precondition — "the
task carries the `harness-ready` label" before the task claim — **cannot be satisfied on any
shipment-claim route**. This plan does not pretend otherwise.

**P-002's Violation Action is literally "Halt and suggest running the harness-architect."** No
disposition in the installed registry authorizes proceeding past a Halt-mandated gate on
*notification*; every override this registry defines is an **explicit recorded authorization** —
P-001's `skip_policy`, P-014's merge approval, P-018's audited force, P-012's declared degradation,
`PA-017-CASCADE`'s own escrow. Disclosure alone is therefore not a sufficient basis to cross an
irreversible one-way door.

**This plan therefore:**

1. **Discloses the deviation explicitly** rather than omitting the policy.
2. **Gates on explicit recorded operator authorization at S-0, before S-3.** The pass criterion is
   an authorization captured verbatim in the run log, reusing the existing `skip_policy: P-002`
   acknowledgement shape rather than minting a new token. **Absence, ambiguity or silence ⇒ HALT
   to Stage, zero mutation (Regime A).**
3. **Emits the full three-action P-005 record** (**M-22**): broadcast with policy ID and one-line
   summary; **compliance annotation in the S-10 and S-22 PR descriptions**; memory-checkpoint
   record — carrying `violation_policy: P-002`, `gate: S-3`, `skip_policy: P-002`. The broadcast
   surface is probed per **P-012** (`TOOL_OK`/`TOOL_DEGRADED`) with a mandatory written fallback to
   the session log and PR description, so the safeguard can never silently no-op.
4. **Preserves P-002's protective intent through P-004's real producer, not a self-applied label**:
   S-7 invokes the installed `harness-architect` skill, which applies `harness-ready` itself after
   its own red-phase checks (**§4.2**). Red is observed and recorded **before the first
   implementation line** of S-8.
5. **Does not amend P-002** and **does not assert conformance.**

### 4.2 P-004 is satisfied in full — the label's producer is never bypassed

Only P-002's *ordering* is deviated from. **P-004 is not deviated from at all.**

P-004's Violation Action is literally *"Do NOT apply `harness-ready` label"*, and P-002 names the
**harness-architect skill as the producer** — Ship is only the consumer. Ship self-applying the
label would emit a **false compliance signal on the exact label P-002 consumes**, which is worse
than omitting it and would void §4.1(4) entirely.

S-7 therefore **invokes the installed `harness-architect` skill** (present at
`.github/skills/harness-architect/`, **M-29**); the skill applies the label at its own Step 6 after
its Step 5.1/5.2 checks. The pass criterion adopts P-004's precondition **and** `018.008-T`'s
recorded **AC-13** literals verbatim:

* `go vet ./...` exits **0**;
* `go test ./tests/integration/ -run TestStageBranchGate_ContractCorrected` exits **non-zero**;
* the captured output contains the literal marker `not implemented: stage branch gate contract`,
  attributed to `TestStageBranchGate_ContractCorrected`;
* the H0 red count over this release unit's generated functions is **literally 1 of 1**;
* the harness manifest records `Compilation: PASS` and `Red Phase: CONFIRMED`.

Any shortfall ⇒ **do not apply the label** and **HALT** (Regime B).

The general conflict affects every shipment-claim route in this workspace. It is captured as stash
**`DB12DA37`** and is out of scope for `017-S` (§10) — resolving it would require amending P-002 or
changing claim semantics, either of which exceeds this shipment's frozen 13-member scope.

## 5. Preflight gates

Run **in order** on the pre-claim branch, **before** the claim. **Any failure halts with zero
mutation.**

| Gate | Check | Pass criterion | On failure |
|---|---|---|---|
| **PF-1** | Engine identity | `backlogit version` ⇒ `1.10.1-…`; SHA-256 of the resolved binary = `1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98`. Resolve via `Get-Command backlogit`, never a hardcoded path | **HALT to Stage** — a mismatch invalidates every measurement in §3 and PF-7 |
| **PF-2** | This plan on `origin/main` | `git show origin/main:docs/plans/2026-09-16-intercom-go-017-S-closure-plan.md` resolves; everything **above the `## Plan Review` heading** is byte-identical to `git show <reviewing_sha>:<this plan path>`, where `reviewing_sha` is read from the **`reviewing_sha` column** of the winning attempt row; that row's `decision` cell is `PASS` and its `Reviewers` cell names the persona coverage; **no later attempt supersedes it**. **No literal in-row marker tokens are required** — the ledger is a table and renders values as cells, so demanding literal `dispatch_mode:` / `decision: PASS` strings made this gate unsatisfiable against the plan's own format | **HALT to Stage** |
| **PF-3** | Workspace integrity | `backlogit doctor` ⇒ `No issues found.`, exit 0 | **HALT to Stage** |
| **PF-4** | Manifest + topology | manifest = 13; descendant graph ∪ `{018-F}` set-equal to manifest; every `parent_id` resolves to `018-F` transitively; enumeration error-free | **HALT to Stage** |
| **PF-5** | §3.3 allowlist | the archived-declaring descendant set is **exactly** the 11-ID allowlist; each satisfies §3.3 conditions 1–4 | **HALT to Stage** — any off-list archived descendant |
| **PF-6** | Linked-deliberation set | re-run §3 M-5's three methods (**positional** `link list`) ⇒ set **EMPTY**. Record `019.007-T`'s pre-state **for provenance only — informational, never an assertion baseline** — and its `dependencies: [018.008-T]` edge for R-7 | **HALT to Stage** |
| **PF-7** | Probe re-verification (**read-only**) | §5.1 | **HALT to Stage** |
| **PF-8** | Topology gate | `--phase pre_claim` ⇒ exit 0 | **HALT to Stage** |
| **PF-9** | Ground 1 | `021-S` ⇒ `status: archived` **and** its own frontmatter `archived_status` is exactly `shipped`, with `commit` provenance. `archived` alone is insufficient — it is also the terminal state of an *abandoned* shipment. Cross-check `74330d3` / `f2d4cf9` / PR #56 | **HALT to Stage** |
| **PF-10** | Ground 2 | the **installed** `_ship.agent.md` contains Step 6.1(a1) with the S0–S5 selector, and `.github/policies/workflow-policies.md` carries the §3.2 grant. Verified against the installed files, not against this plan's description of them | **HALT to Stage** |
| **PF-11** | Close-path authority | §5.2 | §5.2 |
| **PF-12** | Producer admission (**P-012**, read-only, **pre-claim**) | `.github/skills/harness-architect/SKILL.md` resolves **and** its Step 1 intake admits a work item whose status is `active` — or a declared P-012 fallback exists. The §4.1 P-002 mitigation rests entirely on this producer; validating it at S-7 would place the check **after** the S-3 one-way door, so a refusal there would strand `017-S`. Log `TOOL_OK` / `TOOL_DEGRADED`. | **HALT to Stage** |

### 5.1 PF-7 — probe re-verification (READ-ONLY)

This gate **authors nothing, mutates nothing, and adds no backlog item.** It re-verifies two
already-committed artifacts and the script that produced one of them.

| Sub-gate | Artifact | Required values |
|---|---|---|
| **PF-7.1** | `docs/plans/evidence/2026-09-12-task-only-shipment-finalization/probe25-root-included-cascade-fixture.txt` (committed `e36d853`) | `DIGEST_GATE=PASS`; `TOPOLOGY_MEMBERS=13`; `ROOT_INCLUDED=True`; `PRE_ARCHIVED=11`; `ARCHIVED_STATUS_FLAVOUR=queued`; C1–C8 all `PASS`; `FAILED_CRITERIA=0`; `PROBE25_RESULT=PASS` |
| **PF-7.2** | `docs/plans/evidence/2026-09-16-017-S-closure/probe26-claim-cascade-root-included.txt` **and** its generating script `docs/plans/evidence/2026-09-16-017-S-closure/probe26.ps1` — **both MUST resolve via `git show origin/main:<path>`**. A working-tree-only copy does **not** satisfy this gate | `DIGEST_GATE=PASS`; `FIXTURE_VALIDITY_GATE=PASS`; `TOPOLOGY_MEMBERS=13`; `ROOT_INCLUDED=True`; `PRE_ARCHIVED=11`; `ARCHIVED_STATUS_FLAVOUR=queued`; `CLAIM_EXIT_CODE=0`; C1–C6 all `PASS`; `FAILED_CRITERIA=0`; `PROBE26_RESULT=PASS` |
| **PF-7.3** | Engine agreement | both artifacts' `ENGINE_SHA256` equal PF-1's measured digest |

**Why two probes.** They close different gaps and neither subsumes the other. Each is credited
with exactly what its artifact records and nothing more.

* **Probe 25** covers the **structural/topology** half over the root-included, archived-sibling
  shape, via criteria C1–C8 as literally named in its file.
* **Probe 26** covers the **claim** half — what `shipment claim` does to a manifest holding 11
  genuinely archived members. Its `C3` is the measured basis for S-5; its `C1` for S-4; its `C2`
  for S-7's pre-emption disclosure.

**What neither probe measures:** the cascade close itself over this shape, the `safe-close`
binding revalidation path, and the behaviour of the installed Step 6.1(c)/(d) remedies. Those are
guarded at runtime by S-17, S-18, S-20 and §8 — not by probe evidence.

### 5.2 PF-11 — close-path authority agreement

| Sub-check | Requirement | On divergence |
|---|---|---|
| **PF-11.1** | `backlogit shipment ship --help` exposes **exactly** `--sha`, `--message`, `--author` and **no** binding parameter | If a binding parameter has appeared, the direct-cascade route is now executable and §4's reconciliation is **VOID** ⇒ **HALT to Stage for re-measurement** |
| **PF-11.2** | `.autoharness/backlog-registry.yaml` `ship_shipment.params` is **exactly** `shipment_id`, `sha`, `message`, `author` | as PF-11.1 |
| **PF-11.3** | `shipment-reconcile/SKILL.md` still accepts `classification_binding` for `safe-close` **and** still routes a bound `CASCADE` verdict into the Cascade Close Sub-Procedure | **HALT to Stage** — the route's executable realization has changed |
| **PF-11.4** | The installed `_ship.agent.md` close-path delegation is read and its current shape **recorded in the run log** | **Record, do not halt.** A correction of the known-defective binding-carrying clause must not fail the route; only PF-11.1–11.3 gate it |

## 6. Execution sequence

### Phase 1 — preflight and claim

* **S-0** — **P-002 deviation gate (pre-claim, before S-1).**
  * **Probe the broadcast surface first (P-012):** log `TOOL_OK` or `TOOL_DEGRADED`. When
    `agent-intercom` is unavailable, the **mandatory written fallback** applies — record the
    disclosure in the session log and the PR description (P-017). The safeguard preceding the
    one-way door must never silently no-op.
  * Surface the §4.1 disclosed deviation to the **operator** and emit the full **three-action**
    P-005 record (**M-22**): broadcast, PR-description compliance annotation, memory checkpoint —
    carrying `violation_policy: P-002`, `gate: S-3`, `skip_policy: P-002`.
  * **Pass criterion: an EXPLICIT RECORDED OPERATOR AUTHORIZATION to proceed under the declared
    P-002 deviation, captured verbatim in the run log.** P-002's Violation Action is *Halt*;
    notification is not a disposition the registry recognizes. **Absence, ambiguity or silence ⇒
    HALT to Stage, zero mutation (Regime A).**
  * **The irreversible S-3 claim is never taken under an undisclosed or unauthorized policy
    violation.** This step mutates nothing.
* **S-1** — **Pre-claim branch (P-011, P-016).** Create the Ship working branch
  `feat/017-s-stage-artifact-branch-pr-policy-gap-correction` (or a `feat/`/`chore/` variant
  matching installed branch-name policy). Verify: clean `git status --short`;
  `git worktree list --porcelain` classified per P-016 required checks 1–4, **failing closed on
  any prohibited or ambiguous extra worktree**. This branch is **distinct** from the closure
  branch of S-12, which does not exist yet.
* **S-2** — **Pre-claim reconciliation.** `mode: pre`, `expected_status: queued`, while every
  manifest member holds its measured pre-claim status — **2 members `queued`, 11 classifying as
  `pre-archived` per M-4**, not a uniform status. **PROCEED requires every item to classify as
  `matched` or `pre-archived`; continue only on `PROCEED`.**
  *Disclosure (decision §7 item 5, reference only — no halt condition derives from it): the
  installed Step 6.1 preamble describes `mode: pre` as queue-only; that summary is stale and the
  `shipment-reconcile` SKILL is the runtime authority. Trust the skill's actual output.*
* **S-3** — **Claim `017-S`** via the CLI surface: `backlogit shipment claim 017-S`.
  * **This is a one-way door (M-10). Every PF gate must have passed before this line executes.**
  * The CLI surface is pinned because that is the surface Probe 26 measured
    (`AUTHORIZATION_SURFACE=CLI-only`). If operational constraints force the MCP surface, that is
    a **known-unmeasured** path: S-4 and S-5 become the only guard, and the substitution MUST be
    **recorded in the run log**, never made silently.
  * Record resulting status.
* **S-4** — **Condition-5 verification (fail-closed).** Re-read `018-F`'s live status.
  **Pass criterion: exactly `active`.** The claim performs this transition (measured, Probe 26
  `C1`); **Ship performs no `queued → active` write on a feature**, and no agent repairs a wrong
  value by writing the status directly.
* **S-5** — **Archived-member non-drift verification (fail-closed).** Re-read all **11**
  allowlisted members. Per member: still archive-located, `status` exactly `archived`,
  `archived_status` present and unchanged, `parent_id` unchanged. Probe 26 `C3` predicts no
  drift; S-5 makes that falsifiable **at the cheapest unwind point**.
* **S-6** — Topology gate `--phase post_claim` ⇒ exit 0 (`017-S` is the sole active shipment).

### Phase 2 — discharge §3.2 conditions 1 and 2

* **S-7** — **Harness first (P-002/P-004).**
  * The claim has **already transitioned `018.008-T` `queued → active`** (Probe 26 `C2`). Ship
    performs no separate transition; if it re-reads as anything other than `active` ⇒ **HALT**.
  * **Invoke the installed `harness-architect` skill** for `018.008-T` ("Gate Stage artifact
    mutation behind a dedicated branch") to generate the failing harness. **Ship does NOT apply
    the `harness-ready` label — the skill is P-002's named producer and applies it at its own
    Step 6** (§4.2). P-004's Violation Action is literally *"Do NOT apply `harness-ready` label"*;
    a Ship-applied label would be a false compliance signal on the exact label P-002 consumes.
  * **Probe the producer before relying on it (P-012).** `018.008-T` is already `active` at this
    point, and the skill's Step 1 excludes non-ready work items. Confirm the skill accepts an
    `active` task; if it refuses, log `TOOL_DEGRADED` and **HALT to the operator (Regime B)** — S-7 is post-claim, so Regime A is unavailable (§6.1 D-J11). PF-12 pre-validates this admission before the one-way door precisely so this halt is rare. The §4.1 deviation's
    entire mitigation is unexecutable without the producer, and that must surface before S-8, not
    be discovered as a silent omission.
  * **Observe and record the red phase before writing any implementation line** (**AC-6**).
  * **Pass criterion — P-004's precondition plus `018.008-T`'s recorded AC-13 literals, verbatim:**
    `go vet ./...` exits **0**; `go test ./tests/integration/ -run TestStageBranchGate_ContractCorrected`
    exits **non-zero**; the captured output contains the literal marker
    `not implemented: stage branch gate contract` attributed to
    `TestStageBranchGate_ContractCorrected`; the H0 red count is **literally 1 of 1**; and the
    harness manifest records `Compilation: PASS` and `Red Phase: CONFIRMED`. Record all five in the
    run log. Any shortfall ⇒ **the label is NOT applied** and **HALT to the operator (Regime B)** — S-7 is post-claim (§6.1 D-J11): a harness that
    passes on first run, or fails for an unrelated reason (compile error, missing fixture,
    unrelated breakage), does not establish red.
  * **Bind `{harness_paths}`** — the exact set of harness/test/config paths created or modified,
    from **`git status --porcelain=v1 -- . ':(exclude).backlogit/'`** at S-7 exit (§6.1 D4),
    normalized to repo-relative paths. The exclusion is required: the S-3 claim wrote
    `.backlogit/` records that no step has committed yet, and ingesting them would make the S-10
    check unsatisfiable by construction.
* **S-8** — **Implement `018.008-T` to green.** Scope is **exactly** `018.008-T`'s recorded
  scope; no scope expansion. **Bind `{impl_paths}`** from **`git status --porcelain=v1 -- . ':(exclude).backlogit/'`**
  at S-8 exit (§6.1 D4/D6), normalized to repo-relative paths — **the same exclusion `{harness_paths}`
  uses, and required for the same reason**: the S-3 claim wrote `.backlogit/` records and S-2 wrote a
  `.backlogit/reconcile/` report, none of which any step has committed yet, so an unfiltered binding
  would make this membership test unsatisfiable by construction. **Then assert MEMBERSHIP:
  `{impl_paths}` ⊆ the §6.1 D2 authorized surface.** Any path outside it ⇒ **HALT (Regime B — S-8 is
  post-claim)**. The membership test is the check; binding alone proves nothing, because a set compared
  against itself can never detect an unauthorized surface. `.backlogit/**` is **outside D2's domain
  entirely** and is tested only by P3 — a `.backlogit/` path is never a D2 membership failure.
* **S-9** — **`018.008-T → done`** (Step 4.5). This discharges §3.2 conditions 1 and 2: every
  descendant is complete-or-exempt and no live descendant remains.
  * **EXPECTED RENAME (M-17).** `done` is archive-routed, so this step **relocates** the record:
    the diff shows `.backlogit/{queue => archive}/018.008-T.md`, not an in-place edit. This
    rename is an **expected artifact under §6.1 P4** and is **never** drift. The former
    queue-only allowlist would have halted here — **post-claim** — because the archive path was
    outside it.
  * **`019.007-T` carve-out.** `019.007-T` declares `dependencies: [018.008-T]` and is outside
    the manifest. Completing `018.008-T` may legitimately unblock it, so **a status change on
    `019.007-T` after S-9 is EXPECTED and is not drift.** All untouched-assertions for it are
    anchored to the **S-15** baseline, never to PF-6.
* **S-10** — **Implementation PR.** Open, review, P-018 engagement, operator-approved merge.
  * **The merge MUST be a true merge commit (P-009). Squash-merge and rebase-merge are
    FORBIDDEN.**
  * **Changed-path control — deny-list plus positive invariants (§6.1). The exhaustive
    allowlist is withdrawn; see §6.1 for why and for the post-claim failure semantics.**
    * **D1 deny match** ⇒ HALT (§9 condition 9, per §8).
    * **P1–P6 positive invariants** (§6.1). A **D1 deny match or a P2/P3/P5/P6 breach ⇒ HALT**; a
    **P4 shortfall is RECORDED and does NOT halt** (§6.1 D5). The two dispositions differ and must
    not be collapsed — collapsing them reintroduces the post-claim stranding this design removes.
    * An **unexpected-but-not-denied** path is **recorded in the run log and does NOT halt**.
  * **P-021 disposition for review findings:** C1 in-scope ⇒ fix now; C2 out-of-scope ⇒ capture a
    deferred stash entry carrying PR number, review-thread ID and the task/feature/shipment IDs,
    and do **not** widen this shipment; C3 defect in this plan ⇒ **HALT to the operator (Regime B)** — S-10 is post-claim (D-J11); Stage re-measurement of the defective clause is advisory input to the operator, never the halt target.
* **S-11** — **Merge confirmation gate.** After the S-10 merge lands, bind and verify:
  * **`{merge_sha}`** := the **PR-reported merge commit**, via
    `gh pr view <n> --json mergeCommit -q .mergeCommit.oid`. **Never `git rev-parse origin/main`**
    — the remote tip equals the S-10 merge only if nothing else lands in between, and an ancestry
    test could never detect that error because any later tip also contains the S-10 changes as an
    ancestor. A wrong `{merge_sha}` propagates irreversibly into `shipment ship --sha` and into
    the archived `commit:` provenance checked at S-20.
  * **Verify, all three required:** (i) `git merge-base --is-ancestor {merge_sha} origin/main`
    succeeds; (ii) the S-10 branch head is an ancestor of `{merge_sha}`; (iii) every path in
    `{impl_paths}` appears in `git show --stat {merge_sha}`.
  * **Verify the parent shape (P-009):** `git cat-file -p {merge_sha}` MUST report **exactly two
    parents**, parent 1 = the mainline (`origin/main`) commit, parent 2 = the reviewed PR head —
    both confirmed by `git merge-base --is-ancestor` against the pre-merge tip and the branch head
    respectively. **Any other shape — one parent (squash/rebase), three or more, or wrong parent
    order — is a P-009 VIOLATION ⇒ HALT.**
  * Bind **`{msg}`** := the merge commit subject; **`{author}`** := the merge commit author.
  * **Pass criterion:** (i)–(iii) hold, the parent shape is exactly as specified, and all three
    values are recorded.

### Phase 3 — post-merge closure

* **S-12** — **Closure branch (P-011, P-016).** Create the closure branch **before any S2
  mutation**; **no S2 mutation on `main`**. Re-verify clean worktree and P-016 worktree
  classification, failing closed on any prohibited or ambiguous extra worktree.
* **S-13** — **`a0`** topology gate `--phase lifecycle` ⇒ exit 0.
* **S-14** — **`a1` covering-feature completion gate.** Execute the **S0 → S5** selector in
  order, stopping at the first that applies:
  * **S0** (anomaly gate, evaluated **before** `n`): no member with missing/unresolvable
    `artifact_type`; none present in both queue and archive; none in neither; no `status:
    archived` member with malformed provenance; **no `status: archived` member off the §3.3
    allowlist**; enumeration complete and error-free. Any anomaly ⇒ **HALT, no feature mutation.**
  * `n` = count of manifest members whose **`artifact_type` is `feature`** — never by ID suffix.
    Here `n = 1` (`018-F`).
  * **S1** (`n == 0`) — not applicable. **S2** (`n > 1`) — not applicable.
  * **S3** — not applicable unless `018-F` already declares `done`.
  * **S4** — all five §3.2 conditions hold ⇒ transition **`018-F` `active → done`** via
    `backlogit_move_item` (CLI fallback `backlogit move 018-F --status done`).
    **Exit-code contract (fail-closed):** `0` = success; **`6` blocked, `7` configuration,
    `8` retryable are each a HALT, never a retry.** A zero exit whose re-read is not exactly
    `done` is **also a HALT**. `--force-gates` / `--force-reason` are **forbidden**.
  * **S5** — any other live status ⇒ **HALT**, record observed status, return to Stage.
* **S-15** — **Pre-cascade baseline.** On the **existing** closure branch — no second branch, no
  second worktree:
  1. Stage **explicit paths only**: `.backlogit/queue/`, `.backlogit/archive/`.
  2. **Verify the stage is non-empty:** `git diff --cached --name-status` MUST list S-14's
     `018-F` mutation. *(`git add` aborts the whole invocation on a single nonexistent pathspec,
     staging nothing; an unverified baseline commit silently destroys the rollback anchor.)*
  3. Commit. Bind **`{pre_cascade_sha}`**.
  4. Verify clean `.backlogit/queue/` + `.backlogit/archive/` trees.
  5. Bind **`{pre_paths}`** — the full sorted path+blob-hash inventory of both directories
     against `{pre_cascade_sha}`, plus the raw frontmatter of `019.007-T`.

  **§8's restore uses exactly this baseline.**
* **S-16** — **Pre-archive reconciliation.** `mode: pre`, `expected_status: done`. Expected:
  `018-F` `done`, `018.008-T` `done`, the 11 pre-archived. **Continue only on `PROCEED`.**
* **S-17** — **`classify-close-path` (read-only, pre-mutation boundary).** Invoke
  `shipment-reconcile` `mode: classify-close-path` with `shipment_id: 017-S`. Read
  `CLOSE_PATH_VERDICT`, `VERDICT_REASON`, `CLASSIFICATION_BINDING`.
  * The verdict MUST be read **before any close call is made**.
  * This mode **MUST NOT** write a reconcile artifact or mutate any workspace/backlog state
    before returning its verdict; the fields are returned as machine-consumable output.
  * Validate the token against the fixed allowlist `CASCADE` / `SAFE_CLOSE` / `BLOCK`.
    **HALT on `SAFE_CLOSE`, on `BLOCK`, and on any absent, empty, unparseable or unnamed token.**
    **Required verdict here: `CASCADE`.**
* **S-18** — **Bound cascade close.** Invoke `mode: safe-close`, passing the
  `CLASSIFICATION_BINDING` returned by S-17 (see §4).
  * Safe-close refuses unbound invocations (`RECONCILE_FAIL_CASCADE_UNBOUND`) and
    drift/invalid/ambiguous bindings.
  * On a revalidated `CASCADE`, the Cascade Close Sub-Procedure performs the close via
    **`backlogit shipment ship 017-S --sha {merge_sha} --message {msg} --author {author}`**.
  * The generic shipment-status-move route is **not** the close path and is not used (M-12).
  * Cascade step 2 HALTs on non-empty `returned_ids`; step 3's two-set gate HALTs on mismatch.
* **S-19** — **UNCONDITIONAL `{post_paths}` capture.** **Immediately on return from S-18** —
  **before** P-007 archive-integrity verification, **before** `mode: post`, **before any commit,
  cleanup, postcheck-driven restore or further mutation** — bind **`{post_paths}`**: the full
  sorted path+blob-hash inventory of `.backlogit/queue/` + `.backlogit/archive/`, plus the raw
  frontmatter of `019.007-T`.

  **This capture is UNCONDITIONAL — taken on EVERY return from the cascade, success or failure.**
  The installed Step 6.1(c) P-007 remedy and 6.1(d) restore both sit between the close and
  post-mode and mutate exactly these two trees; a later capture would compute S-20.3 over an
  already-restored tree. **S-20.3 consumes this capture.** (R-6 recomputes a live inventory and
  compares it to `{pre_paths}`, so R-6 is **not** a consumer of `{post_paths}`.)
* **S-20** — **Postchecks.** All must pass **before anything is committed**:
  1. `mode: post` reconciliation ⇒ every manifest item present in archive.
  2. An optional post-remedy inventory may be taken for diagnostics; it is **not** the assertion
     input. **S-20.3 consumes the S-19 capture.**
  3. **Cascade-scope assertion:** `{post_paths} △ {pre_paths}` is confined to the 13 manifest
     members and the `017-S` record. **`019.007-T` MUST be raw-byte identical to its `{pre_paths}`
     S-15 baseline** — never to its PF-6 snapshot. Any out-of-manifest artifact archived or
     deleted ⇒ `HALT — cascade detected, revert required` ⇒ §8.
  4. **Closure triple:** `backlogit doctor` ⇒ `No issues found.` exit 0; `017-S` ⇒
     `status: archived`; raw archive frontmatter ⇒ `archived_status: shipped` +
     `commit: {merge_sha}`.
* **S-21** — **Commit the cascade result**, only after every S-20 postcheck passes. Bind
  **`{cascade_commit_sha}`** — an ordinary single-parent commit on the closure branch, and **R-4's
  revert target**.
* **S-21.5** — **Closure knowledge artifacts and P-020 discharge.**
  * Author `docs/closure/017-S-018-F-post-merge-closure.md`.
  * **Invoke the compact-context skill with `target: all`.** **P-020** makes this MANDATORY at
    every post-merge closure; **SKIPPING** it is the violation — recorded through the full P-005
    triple with `violation_policy: P-020`, `gate: Ship Step 6 closure`, `action: closure-incomplete`
    — which leaves the closure incomplete and **holds the next shipment under P-001**. The mechanism
    is the skill, never an environment-specific in-conversation compaction command.
  * **A compact-context run that FAILS is NON-BLOCKING** (installed P-020 Violation Action,
    verbatim): record `compaction: degraded` in the closure artifact, log a warning, and
    **continue to S-22**. S-21.5 is therefore **explicitly excluded from §8's fail-closed default
    and from every §8.2 R-trigger** — the merge has already landed and the Tier-1 skill is
    non-destructive (it archives, never deletes). Halting here would invert the installed policy.
  * Record `compaction: {status}` in the closure artifact — `ok` or `degraded`, never unset.
  * These are the `docs/closure/`, `docs/memory/` and `docs/compound/` paths **M-21** measured in
    the `021-S` precedent and that **§6.1 D5a** expects in the S-22 PR. Without this step P4 would
    require artifacts no step authors.
* **S-22** — **Closure PR.** P-014 local review gate, operator-approved merge, verified on
  `origin/main`. Bind **`{closure_merge_sha}`** — consumed at **R-5 for its two-parent evidence
  check only. It is NEVER a revert target** (I-12): it contains the S-15 baseline as well as the
  cascade commit, so reverting it would destroy the rollback anchor.
  * **The merge MUST be a true merge commit (P-009)**, verified exactly as at S-11.
  * **Changed-path control per §6.1**, evaluated **before the merge**. A violation here is a
    **PRE-MERGE HALT to the operator, never an R-trigger** — the closure merge is operator-gated,
    so path control is asserted before it lands, and a documentation-shaped violation must never
    unwind a correct, verified, operator-approved close.
  * **Expected set (M-21), used to satisfy P4 — not as a deny-list:** the cascade commit plus
    `docs/closure/017-S-018-F-post-merge-closure.md`, `docs/memory/*.md`, `docs/compound/*.md`.

### 6.1 Changed-path control (authoritative)

**Why the allowlist was withdrawn.** The exhaustive form (`anything not enumerated ⇒ HALT`)
required perfectly predicting every path a multi-tool chain touches, and its failure mode was
catastrophic: a halt **after** the one-way door, stranding `017-S` `active` with no unclaim.
Three of attempt 5's seven P0s were instances of that single design choice — the `018.008-T`
queue→archive rename (**M-17**), the reconcile reports (**M-18**–**M-20**), and the closure
knowledge artifacts (**M-21**) were all real, expected emissions that the allowlist would have
treated as fatal drift.

**D1 — FORBIDDEN SURFACES (deny-list).** A changed path is a violation **if and only if** it
matches one of:

* **(a)** production implementation code — `internal/**`, `cmd/**`, any `*.go` that is **not** a
  `*_test.go` file, `go.mod`, `go.sum`. The `*_test.go` carve-out applies uniformly, including
  under `internal/` and `cmd/`, though `018.008-T`'s recorded scope independently forbids any
  harness there;
* **(b)** `.github/policies/**`, `.github/agents/**`, `.github/skills/**`,
  `.github/instructions/**` or `.github/workflows/**` **other than** the two surfaces authorized
  in D2;
* **(c)** any `.backlogit/queue/**` or `.backlogit/archive/**` record not belonging to the 13
  manifest members, the `017-S` record, or the evidenced `019.007-T` delta (D3);
* **(d)** any **deletion** of a `.backlogit/reconcile/**` report;
* **(e)** **governance and control-plane surfaces** — `docs/plans/**` and `docs/decisions/**`
  (including this plan, the canonical contract, and the **FROZEN** prerequisite
  `docs/plans/2026-09-13-intercom-go-ship-feature-completion-decided-plan.md`), `.autoharness/**`,
  `.backlogit/registry.yaml`, `.backlogit/header-def.yaml`, `.backlogit/config.yaml`,
  `scripts/**`, `.gitignore`, `go.work*`.

Nothing else is fatal by enumeration.

**Why (e) is required.** A deny-list is fail-open by construction: anything unnamed is merely
recorded (D5). Every surface whose immutability this plan asserts elsewhere — §10's out-of-scope
list, **I-7**'s FROZEN prerequisite, the routing table **M-17** depends on — must be named here,
or that assertion has **no path-level enforcement at all**. `docs/**` writes are normal on this
route (the closure PR legitimately carries `docs/closure/`, `docs/memory/`, `docs/compound/`), so
an errant `docs/plans/` write is plausible and would otherwise pass every gate.

**Worktree/parallel-branch paths are deliberately NOT a D1 clause.** A path set derived from
`git status` or `git show --stat` in the current worktree can never name a path in another one, so
as a deny predicate it would be vacuous. That prohibition is enforced where it is decidable — the
**P-016 fail-closed worktree classification at S-1, S-12 and before R-5's branch creation**.

**D2 — AUTHORIZED IMPLEMENTATION SURFACE (a CLOSED set of exactly three paths).**
`018.008-T`'s recorded scope authorizes exactly:

1. `.github/policies/workflow-policies.md`
2. `.github/agents/_stage.agent.md`
3. `tests/integration/stage_branch_gate_test.go` — the harness file named literally by the task's
   recorded `harness_cmd` and AC-13

The task records `FILE COUNT: 3 files total … 0 new production files`, so
`|{impl_paths} ∪ {harness_paths}| ≤ 3` and **`{impl_paths}` MUST be a SUBSET of that closed set** —
a membership test. This closes both the tautology (the former rule compared `{impl_paths}` against
a set derived from `{impl_paths}` itself) and the undecidability of an unenumerated "the test
harness file".

**D3 — `019.007-T` carve-out.** `019.007-T` (`status: blocked`, `dependencies: [018.008-T]`,
parent `019-F`, outside the manifest) may legitimately unblock when `018.008-T` reaches `done`.
Its delta is permitted when evidenced in the run log by a `018.008-T → done` transition
immediately preceding it and confined to `status`, `dependencies` and **record metadata
timestamps (`updated_at`)**. The timestamp is included because **any engine write updates it** —
confining the carve-out to status/dependency fields alone would make a *legitimate* unblock match
**D1(c)** and halt **post-claim**, reintroducing the exact stranding failure mode this section
exists to remove. The run-log evidencing is the operative condition.

**P1–P6 — POSITIVE INVARIANTS.** All must hold; each is independently falsifiable.

**Evaluation set (defined once, used by P1 and P3):** `git diff --name-status {merge_base}...HEAD`
for the PR under assertion, where `{merge_base}` is `git merge-base origin/main HEAD`. P5 is
deliberately **not** evaluated over this set (see below).

| ID | Invariant |
|---|---|
| **P1** | No changed path in the evaluation set matches the D1 deny-list. |
| **P2** | The manifest is exactly 13 members before and after. |
| **P3** | Every backlog record change belongs to a manifest member, the `017-S` record, or the evidenced `019.007-T` delta. |
| **P4** | The **expected** artifacts are PRESENT, scoped per D5a — the queue→archive renames for every member reaching a terminal status (**M-17**), and the reconcile reports for the modes that **persist** one. |
| **P5** | **No production implementation code changed** — asserted **independently of the evaluation set** by a repo-scoped command: `git diff --name-only {merge_base}...HEAD -- internal/ cmd/ go.mod go.sum` returns empty, and no non-test `*.go` appears in the full diff. P5 is independent of P1 by construction, so it can fail even where the path-set enumeration itself is wrong. |
| **P6** | **SUB-FILE CONTENT LOCK.** D1/D2 test *paths*, but `018.008-T`'s recorded scope binds *regions within* those files. Assert by diff: changed hunks in `.github/policies/workflow-policies.md` fall **only** within the P-010 section and the Amendment Log; changed hunks in `.github/agents/_stage.agent.md` fall **only** within the Role Boundary Git/PR rows, the Step 1.9 section and the Step Sequence Contract checklist line; and the P-016 paragraph beginning `**Allowed exception (Stage spike/research only)**`, P-010's Orchestrator statement, and the **P-015** section are **byte-for-byte identical** pre/post. Without P6, an edit to P-015, P-009 or P-016 inside an authorized file passes every other gate while §10 forbids it and escrow condition 9 leans on it. |

**D5a — EXPECTED-ARTIFACT SCOPE for P4.** P4 is satisfied over the modes that **persist** a
report — `pre`, `safe-close`/`cascade-close`, and `post`. Per **M-19** the installed skill writes
**no** report for `classify-close-path` under any invocation, so requiring one would make P4
unsatisfiable by construction, and S-17 independently forbids that write. The measured filenames
are `017-S-pre-*.md`, `017-S-cascade-close-*.md` and `017-S-post-*.md`; a literal `{mode}`
substitution would mis-predict the cascade filename. The closure knowledge artifacts
(`docs/closure/`, `docs/memory/`, `docs/compound/`) are authored at **S-21.5**.

**D4 — `{harness_paths}` binding excludes backlog state.** Bind from
`git status --porcelain=v1 -- . ':(exclude).backlogit/'`. The S-3 claim writes `.backlogit/`
records that no step has yet committed; an unfiltered `git status` would ingest them and render
the gate unsatisfiable by construction. `.backlogit/` paths are governed solely by **P3**.

**D5 — FAILURE SEMANTICS (safe after the one-way door).**

| When | Condition | Disposition |
|---|---|---|
| Before S-3 (pre-claim) | any violation | HALT to Stage, zero mutation — cheap |
| After S-3, before S-18 | **D1 deny match**, or a **P2/P3/P5/P6** breach | HALT, `017-S` left `active` per §8.1 Regime B |
| After S-3, before S-18 | **unexpected-but-not-denied path (P4 shortfall only)** | **RECORD in the run log; DO NOT HALT.** This is the specific change that prevents an unforeseen tool emission from stranding the shipment |
| After S-18 (post-mutation) | any | §8.2 governs; a **P4 shortfall alone never triggers an unwind** of a verified close |

## 7. Variable bindings

| Variable | Bound at | First consumed at |
|---|---|---|
| `{harness_paths}` | S-7 | S-10 |
| `{impl_paths}` | S-8 | **S-8** (membership assertion), then S-10 |
| `{merge_sha}`, `{msg}`, `{author}` | S-11 | S-18 |
| `{pre_cascade_sha}` | S-15 | R-8 |
| `{pre_paths}` | S-15 | S-20 |
| `{post_paths}` | S-19 | S-20 |
| `{cascade_commit_sha}` | S-21 | R-4, R-5 |
| `{closure_merge_sha}` | S-22 | R-5 (two-parent evidence check only — NOT the revert target) |
| `{quarantine_sha}` | R-3 | R-3 |
| `{revert_merge_sha}` | R-5 step (5) | R-6, R-7 (evaluation site on `origin/main`) |

No variable is consumed before the step that binds it.

**Binding discipline.** **Both** `{harness_paths}` (S-7) and `{impl_paths}` (S-8) are bound from
`git status --porcelain=v1 -- . ':(exclude).backlogit/'` (§6.1 D4/D6), so the claim's uncommitted
backlog writes and the S-2 reconcile report are never ingested into either binding. The exclusion
is **not** a `{harness_paths}`-only convenience — it is required identically for `{impl_paths}`,
because the same uncommitted `.backlogit/` writes are present in the working tree at S-8 and an
unfiltered binding would make the S-8 membership test unsatisfiable by construction. `{impl_paths}`
is **then tested for membership** in the §6.1 D2 authorized surface — the binding alone is not the
check. `.backlogit/**` lies outside D2's domain entirely and is tested only by P3.

## 8. Failure semantics and rollback

### 8.1 Halt regimes (all halts from S-0 and PF-1 through the S-18 pre-mutation refusal)

**Measured ground truth (M-10): a claimed shipment CANNOT be un-claimed.** The only transitions
out of `active` are `shipped` and `abandoned`.

> **`abandoned` IS FORBIDDEN ON THIS ROUTE.** It is terminal and destructive, it is **not**
> covered by `PA-017-CASCADE` (whose scope is the cascade close), and **no operator approval for
> abandoning `017-S` exists.** Improvising an abandon as a "rollback" would be an unapproved
> destructive act — strictly worse than the halt it purports to remedy.

**The regime is decided by exactly one question: is `017-S` live-status `active`?**

**Regime A — `017-S` is NOT `active`** (pre-claim, or a failed claim that left it `queued`):
1. Mutate nothing further. Record the failure token verbatim and the step ID.
2. **HALT to Stage.** No unwind is required and none is performed.

**Regime B — `017-S` IS `active`:**
1. **Stop.** Mutate no further backlog record; do not proceed to the next step.
2. Record the failure token verbatim, the step ID, and the live status of `017-S`, `018-F`,
   `018.008-T` and the 11 allowlisted members.
3. **Leave `017-S` `active`. Perform NO backlog mutation of any kind.** Specifically
   **FORBIDDEN**: `shipment return-blocked` (M-11), any `backlogit move` on the shipment (M-12),
   `abandoned`, and **R-1…R-8** (they presuppose cascade effects and a `{pre_cascade_sha}` anchor
   that may not exist).
4. **HALT to the operator** with the record from (2). Disposition of the claim is the
   **operator's** decision, informed by Stage re-measurement — never Ship's improvisation.
5. `PA-017-CASCADE` is **not** invoked and remains unexercised and intact.

**Disclosed cost:** in Regime B, `017-S` occupies the **P-001 single-active-release-unit slot**
until an operator acts. This is stated rather than engineered around, because every available
workaround is either manifest-breaking or an unapproved destructive act.

**Violation telemetry (P-005) — all THREE required actions, not one.** Every halt corresponding to
a named policy performs the **full installed P-005 triple** (§6.1 D-J9): (1) **broadcast** the
violation with policy ID and one-line summary; (2) **include the violation details in the PR
description as a compliance annotation** — recorded `N/A (no PR exists at this gate)` when the halt
precedes PR creation, per P-021 C3's per-field availability precedent; (3) **record it in the
memory checkpoint** for the affected artifacts. Each event records the policy ID, the failure
token, the step ID and the affected artifact IDs, and the closure record carries it. Performing
action (1) or (3) alone is a P-005 shortfall and **must not be certified as acceptance-passing**.

| Halt at | State | Regime + row-specific addition |
|---|---|---|
| **S-0** (no explicit recorded operator authorization for the P-002 deviation, or the authorization is ambiguous) | no claim performed; nothing mutated | **Regime A.** P-002's Violation Action is *Halt*; S-0 is fail-closed and never proceeds on notification alone (§4.1) |
| **PF-1 … PF-12, S-1** | no claim performed | **Regime A.** Nothing to unwind — the only cheap-exit window, which is why §5's gates are exhaustive |
| **S-2** (reconcile fail, pre-claim) | no claim performed | **Regime A.** S-2 runs strictly before S-3 |
| **S-3** (the claim itself fails) | claim may be **partial** | **Re-read `017-S` live status — the regime-deciding question.** Still `queued` ⇒ **Regime A**. `active` ⇒ **Regime B**, and additionally run S-5's non-drift check to characterize what the partial claim touched |
| **S-4** (`018-F` not `active`) | claim active; no code, no record write | **Regime B.** Contradicts Probe 26 `C1` ⇒ **Stage re-measurement trigger** |
| **S-5** (archived-member drift) | claim mutated archived members | **Regime B**, **plus**: quarantine the drifted records to a `quarantine/017-S-<utc>` ref by additive commit before halting. **This is S-5's own disposition — not R-1…R-8**, which presuppose cascade effects that have not occurred. Contradicts Probe 26 `C3` ⇒ the evidentiary basis is void; **Stage must re-measure before any re-route** |
| **S-6** (post-claim topology gate) | claim active | **Regime B** |
| **S-7** (red not established per §4.2's AC-13 literals, or the `harness-architect` producer refuses the `active` task) | claim active; no implementation written; **`harness-ready` NOT applied** | Discard the harness working tree **on the pre-claim branch only** — ordinary pre-commit iteration, not a backlog-record unwind, and outside §8.2's additive-only ban. Then **Regime B** |
| **S-8 / S-9** (implementation or `done` transition fails) | claim active; code on the pre-claim branch, **unmerged** | Leave the branch intact as evidence. **Do not merge.** Nothing is on `main`, so no revert is required. **Regime B** |
| **S-10** (PR not merged, P-009 shape violation, or a §6.1 **D1 deny match / P2/P3/P5/P6 breach**) | claim active; PR open or improperly merged | Close or park the PR. A squash/rebase merge is a **P-009 violation** ⇒ escalate to the operator. **Regime B** |
| **S-11** (`{merge_sha}` unbindable or parent shape wrong) | **merged** | **Do not proceed to S-12** — every later step consumes `{merge_sha}`. The merge itself **stands** (legitimate, reviewed work). **Regime B** |
| **S-12 / S-13** (branch or `a0` gate fails) | merged impl on `main`; `018-F` **not yet mutated** | Park the closure branch. **Regime B** |
| **S-14** (`a1` fails, possibly mid-mutation) | merged impl on `main`; `018-F` **possibly mutated, UNCOMMITTED** | The baseline commit is S-15, so the S-14 mutation is **uncommitted** — there is no commit to revert, and a `done → active` write is outside the §3.2 grant and outside Ship's Role Boundary. The closure branch is **unmerged**, so **nothing needs reverting**: park the branch, leave the record as-is. **Regime B.** Where the discrepancy is one the **installed** Step 6.1(c) remedy addresses, Ship executes that remedy **as installed** |
| **S-15** (baseline commit or staged-diff verification fails) | `018-F` `done`; **baseline anchor NOT established** | **The rollback anchor does not exist** — R-1…R-8 are unavailable from here. Do **not** proceed to S-16. **Regime B**, escalated: this is the HIGH-rated anchor-destroying case in §11 |
| **S-16** (`mode: pre` fail at `expected_status: done`) | claim active; `018-F` `done`; baseline committed | Baseline exists but **no cascade has run**, so R-1…R-8 do not apply. **Regime B** |
| **S-17** (verdict `SAFE_CLOSE`, `BLOCK`, absent, or unparseable) | claim active; `018-F` `done`; baseline committed; **zero mutation** | `classify-close-path` is read-only, so nothing has been mutated. A realistic fail-closed outcome, **not** an error. **Regime B.** `PA-017-CASCADE` is **not** invoked |
| **S-18 PRE-MUTATION REFUSAL** (`RECONCILE_FAIL_CASCADE_UNBOUND`, `_CLASSIFICATION_INVALID`, `_CLASSIFICATION_DRIFT`, `_CLASSIFICATION_REFUSED`) | claim active; `018-F` `done`; baseline committed; **zero archive, zero status transition, zero record mutation** | The skill refuses **before any mutation**, so there is nothing to unwind. **R-1…R-8 are NOT invoked.** **Regime B.** `PA-017-CASCADE` is **not** exercised — the refusal is the escrow's fail-closed guard working as designed |

**Invariant across this table:** `PA-017-CASCADE` has **not** been invoked at any point above,
**no destructive action is unwound**, and the escrow remains unexercised and intact.

### 8.2 Rollback R-1 … R-8 (post-cascade only)

**Trigger:** any failure **from S-18 onward once the cascade op has actually been invoked** —
specifically a post-invocation **S-18** failure, an **S-19** capture failure, any **S-20**
postcheck failure, an **S-21** commit failure, or any out-of-manifest archival discovered at any
point after the close. Every halt at PF-1 … S-17, and any **S-18 pre-mutation refusal**, is
governed by §8.1 instead.

**An S-22 closure-PR path-control violation is NOT an R-trigger.** It is a **pre-merge HALT to the
operator** (§6.1 D5): the closure merge is operator-gated, so path control is asserted **before**
it lands, and a documentation-shaped violation must never unwind a correct, verified,
operator-approved close. A **P4 shortfall alone never triggers an unwind** at any point.

**Prohibited unwind mechanisms, scoped to R-1…R-8 only:** no working-tree overwrite of backlog
records from a prior commit, and no file deletion. Restoration here is achieved **only by adding
commits**. *(Needle scope: this prohibition names the mechanisms in order to forbid them; a
conformance check for their absence MUST exclude this plan file and MUST anchor to the mechanism's
invocation shape, not to a bare mention, or it will self-match this paragraph.)*

**The installed `_ship.agent.md` Step 6.1(b)/(c)/(d) P-007 remedy remains in force, unmodified**,
and takes precedence where it applies. This plan has no authority to override an installed policy;
its additive-only discipline governs only the post-cascade unwind it defines.

| Step | Action |
|---|---|
| **R-1 QUARANTINE** | **Stop. Mutate nothing further.** Do not commit the cascade result. Record the observed failure token verbatim. `{post_paths}` is captured at S-19 **when that capture succeeded**, and is retained as evidence. **R-6's evidentiary anchor is `{pre_paths}`, not `{post_paths}`** (R-6 recomputes a live inventory and compares it to `{pre_paths}`). If the S-19 capture itself failed — one of this section's own triggers — record that fact and proceed to R-2(a); `{post_paths}` is simply unbound on that path |
| **R-2 CLASSIFY** | **Evaluated in this order; first match wins.** **(d)** committed with any S-20 postcheck **failed or unrun** ⇒ **out of contract: HALT to the operator after R-1, no automatic revert** — an automatic unwind over state whose provenance is unknown is itself unsafe. **(a)** uncommitted, including an S-19 capture failure or an S-21 commit failure ⇒ **R-3**. **(b)** committed with **every S-20 postcheck passed** and the closure branch unmerged ⇒ **R-4**. **(c)** committed, all postchecks passed, and the S-22 closure merge landed ⇒ **R-5**. Evaluating (d) first resolves the former overlap in which committed-unmerged-postchecks-failed matched both (b) and (d) with contradictory dispositions |
| **R-3 UNCOMMITTED** | Use the prerequisite §10.3.2 operator-approved mechanism, **on the closure branch**: (1) `git add -A -- .backlogit/queue/ .backlogit/archive/`; (2) commit as **`{quarantine_sha}`** on that same branch; (3) `git revert --no-edit {quarantine_sha}` on that same branch. A `quarantine/017-S-<utc>` name may be a **ref pointing at `{quarantine_sha}`** for evidence retention — never a separate commit target, which would leave the closure branch with no commit to revert and a dirty tree `git revert` refuses to run against. Proceed to R-6 |
| **R-4 COMMITTED, CLOSURE BRANCH UNMERGED** | Target = **`{cascade_commit_sha}`**, an ordinary single-parent commit. Revert with plain **`git revert --no-edit {cascade_commit_sha}`**. **A single parent is EXPECTED here and is NOT a P-009 signal — P-009 governs PR merges, not intra-branch commits.** Proceed to R-6 |
| **R-5 COMMITTED, CLOSURE MERGE LANDED** | **Revert target = `{cascade_commit_sha}`, NEVER `{closure_merge_sha}`.** The closure branch carries **three** commits: the S-15 baseline `{pre_cascade_sha}` (containing S-14's `018-F → done`), the S-21 cascade `{cascade_commit_sha}`, and the S-21.5 closure-documentation commit. **Only `{cascade_commit_sha}` is ever reverted; the other two must survive.** `{closure_merge_sha}` contains **both**, so `git revert -m 1 {closure_merge_sha}` would unwind the baseline as well — returning `018-F` to `active`/queue-located and making **R-6/R-7 unsatisfiable by construction**, since `{pre_paths}` was captured *after* the baseline commit. **(1) P-011/P-016 prechecks FIRST, before any branch creation:** `git status --short` clean and `git worktree list --porcelain` classified fail-closed; this limb runs in a post-failure state where tree cleanliness is not guaranteed. **(2)** Verify `{closure_merge_sha}` empirically as **evidence only**: exactly two parents, parent 1 = mainline, parent 2 = the reviewed closure-PR head by **SHA equality** (`gh pr view <n> --json baseRefOid,headRefOid`); any other shape ⇒ **P-009 violation, HALT to the operator, no automatic revert**. **(3) Execution site — NEVER on `main` (M-26/P-010):** `git checkout main && git pull && git checkout -b chore/017-s-cascade-revert-<utc>`. **(4)** `git revert --no-edit {cascade_commit_sha}` **on that branch** — `{cascade_commit_sha}` is single-parent (**M-25**), so **plain revert, never `-m 1`**. **(5)** Land it through its **own revert PR** under the full gate set — P-014, P-018, explicit operator approval, **P-009 merge commit** — binding **`{revert_merge_sha}`**. **(6)** R-6/R-7 equivalence is evaluated **against `origin/main` at `{revert_merge_sha}`**, never against an unmerged branch. **(7)** Any failure at (5) ⇒ **HALT to the operator**; Ship never lands the revert by any other path. **This limb lands a merge on `main` under a new explicit operator approval — a moderate-to-high risk action, not a low additive one (§14.2).** Proceed to R-6 |
| **R-6 TREE EQUIVALENCE** | Recompute the inventory over `.backlogit/queue/` + `.backlogit/archive/` and require it **set-equal, path-for-path and blob-hash-for-blob-hash**, to `{pre_paths}` captured at S-15. **`{pre_paths}` is the post-baseline state (`018-F` already `done`), so every revert target across R-3/R-4/R-5 must unwind the cascade ONLY and leave `{pre_cascade_sha}` intact.** Evaluation site: the working tree for R-3/R-4; `origin/main` at `{revert_merge_sha}` for R-5 |
| **R-7 RECORD EQUIVALENCE** | Tree equality is **not sufficient**. For all 13 members plus `017-S` plus `019.007-T`, re-read **raw frontmatter from the artifact files** — never a list, index or query view — and require `status`, `archived_status`, `archived_from`, `parent_id`, `artifact_type` and `commit` to match their **S-15** baseline values exactly. Then `backlogit sync` and require `backlogit doctor` ⇒ `No issues found.`, exit 0 |
| **R-8 VERIFY-OR-ESCALATE** | **A restore is complete only when R-6 AND R-7 both pass.** If either fails, **HALT to the operator** with the quarantine ref, `{pre_cascade_sha}`, and both inventories. **Do not attempt a second automated unwind** |

**`{merge_sha}` is NEVER a rollback target.** It is the S-10 implementation merge; reverting it
would unwind `018.008-T`'s implementation rather than the cascade.

**Post-restore disposition:** the shipment returns to Stage. `PA-017-CASCADE` is **not**
re-exercised in the same session.

## 9. `PA-017-CASCADE` — escrow and invocation conditions

`PA-017-CASCADE` is a **recorded operator approval held in escrow** (authorized
2026-09-13T11:35:43-07:00; `change_kind` **DESTRUCTIVE**; `ActionRisk` destructive; `ActionResult`
approved). **It does not expire and is not re-litigated here.** This plan supplies the execution
site prerequisite §11.1 required, and nothing about the approval is widened.

**It is NOT a claim authorization.** Claiming `017-S` is authorized by the shipment's own
eligibility, never by this approval. Its scope is the **cascade close only** — not abandonment,
not manifest mutation.

**The escrow's release condition is satisfied only when this plan is review-PASSed and present on
`origin/main`, as verified at PF-2.** This plan does not self-certify that condition.

**ALL conditions must hold at invocation, re-measured by Ship — never inherited:**

| # | Condition | Gate |
|---|---|---|
| 1 | Escrow released by an existing, review-passed Stage closure plan on `origin/main` supplying execution site, preflight gates, baseline and rollback | **PF-2** |
| 2 | `classify-close-path` returns `CASCADE` **before any close call** | **S-17** |
| 3 | `safe-close` **revalidates the returned binding before any mutation** | **S-18** |
| 4 | `021-S` has shipped | **PF-9** |
| 5 | parent-completion capability merged and live | **PF-10** |
| 6 | manifest **unchanged at 13 members** | **PF-4** |
| 7 | pre-cascade baseline **and the pre-invocation path inventory** exist and are verified — `{pre_cascade_sha}` + `{pre_paths}` only; `{post_paths}` cannot be a precondition of the call that produces it | **S-15** |
| 8 | validated linked-deliberation set for **this** manifest verified EMPTY by this plan's own measurement | **PF-6** |
| 9 | **no scope expansion** | **PF-4** live manifest invariant; **§6.1 D1/D2** deny-list and closed-set membership; **§6.1 P6** sub-file content lock; **S-10/S-22** changed-path control; **S-20.3** cascade-scope assertion; §10. *(Not PF-7: it is read-only probe re-verification over two committed artifacts and performs no live manifest check.)* |

**Any mismatch INVALIDATES the approval and HALTs.**

## 10. Out of scope

* **Claiming `017-S`** — Stage does not claim; this plan authorizes and sequences only.
* **Implementing `018.008-T`** — Ship work at S-8 under this plan's scope lock.
* **Merging either PR without explicit operator authorization.**
* **Widening or re-litigating `PA-017-CASCADE`.**
* **Amending P-015, `_ship.agent.md` or `shipment-reconcile/SKILL.md`, or any policy/agent surface
  other than the two authorized by `018.008-T`'s recorded scope (§6.1 D2).** **Amending P-010
  within that recorded scope is the authorized work of S-8 and is explicitly IN scope** — the
  prior blanket prohibition contradicted the very step this plan mandates.
* **Resolving the general P-002 / shipment-claim ordering conflict** — captured as stash
  **`DB12DA37`** (§4.1). It affects every shipment-claim route in this workspace and would require
  amending P-002 or changing claim semantics, either of which exceeds this frozen scope.
* **Adding any backlog item to `018-F` or `017-S`** — forbidden by the manifest invariant.
* **Harmonizing the installed Ship CASCADE-branch prose** — captured as stash entry
  **`11B75632`**; a clarity improvement, not a correctness precondition (§4).
* **Decision §7 follow-ups 1, 2, 3 and 5** — recorded, off-route, explicitly not current
  blockers. Item **4 alone** is pulled in, as **PF-7**. Item **5** is referenced at S-2 as a
  **disclosure only**, with no derived halt condition.
* **Re-opening prerequisite attempt 9.** Revision 13 stays **FROZEN**; attempt 8 **FAIL STANDS**.

## 11. Risk register

| Risk | Severity | Mitigation |
|---|---|---|
| Claim does not set `018-F` to `active` | **HIGH** — blocks condition 5 with no in-band repair | **S-4** fail-closed verification; **halt to the operator (Regime B)** — S-4 is post-claim (D-J11). Deliberately not repaired by a direct write, which would violate the §3.2 grant shape |
| Claim mutates pre-archived members | **HIGH** — silently breaks §3.2 condition 2 and the §3.3 exemption | **S-5** non-drift check at the cheapest unwind point |
| Baseline commit silently stages nothing | **HIGH** — destroys the rollback anchor | **S-15 step 2** mandatory non-empty staged-diff verification |
| Wrong artifact reverted during unwind | **HIGH** — would undo `018.008-T`'s implementation, or destroy the S-15 rollback anchor | **R-4/R-5 both target `{cascade_commit_sha}`**; `{merge_sha}` and `{closure_merge_sha}` are never revert targets (**I-12**) |
| Quarantine commit written where nothing can be reverted | **HIGH** — destructive action with a non-executing rollback | **R-3** commits the quarantine **on the closure branch**, then reverts it there; `quarantine/*` is a ref only |
| Post-capture taken after a restore | **HIGH** — scope assertion computed over already-restored state | **S-19** unconditional capture immediately on return, before P-007 and post-mode |
| Changed-path control unsatisfiable after the one-way door | **MEDIUM** (reduced from HIGH by the §6.1 redesign) | The exhaustive allowlist is **withdrawn**. §6.1 D5 record-don't-halt semantics mean an unexpected-but-not-denied path is logged, not fatal; only a **D1 deny match or a P2/P3/P5/P6 breach** halts. D3's `updated_at` allowance keeps a legitimate `019.007-T` unblock from matching D1(c). §8.1 Regime B disclosed |
| `git revert -m 1` applied to a non-merge commit | **LOW** (eliminated by design) | **No limb uses `-m 1`.** R-3/R-4/R-5 all revert single-parent commits with a plain `git revert`; parent counts are still verified empirically as evidence |
| Tree restored but records semantically drifted | **MEDIUM** | **R-7** record equivalence from raw frontmatter |
| Off-allowlist archived descendant appears later | **MEDIUM** | **PF-5** + **S-14 S0** both halt |
| Cascade route rendered inexecutable by tooling drift | **MEDIUM** | **PF-11** falsifies §4's reconciliation at preflight |
| Stale `--id` flag form in inherited probes | **LOW** | **M-16**; PF-6 pins the positional form |

## 12. Acceptance criteria

1. **AC-1** — **PF-1 … PF-12** all pass before the claim; any failure halts with **no mutation**.
2. **AC-2** — **PF-7 is read-only**: it re-verifies two already-committed probe artifacts plus the
   Probe 26 script and their engine agreement. This plan authors no probe, no test and no backlog
   item, and PF-7 adds no manifest member.
3. **AC-3** — **PF-2 gates escrow release**: the plan resolves on `origin/main`, its pre-review
   body is byte-identical to the reviewed revision, and the most recent attempt records
   `decision: PASS` unsuperseded.
4. **AC-4** — `018-F` is `queued` at S-2, **exactly `active` at S-14 entry**, and **exactly `done`
   on S-14 success**; **no agent performs a `queued → active` write on it**.
5. **AC-5** — S-5 verifies all 11 archived members undrifted after the claim; any drift halts.
6. **AC-6** — S-7 establishes **red before any implementation line**, with a recorded failing test
   name and assertion message attributable to the absent `018.008-T` behavior.
7. **AC-7** — S-9 transitions `018.008-T` to `done`, discharging §3.2 conditions 1 and 2.
8. **AC-8** — S-14 executes S0→S5 **in order**; `n` is computed by `artifact_type`, never by ID
   suffix; `move` exit codes 6/7/8 each halt **without retry**; `--force-gates`/`--force-reason`
   are never used.
9. **AC-9** — `{pre_cascade_sha}` is committed on the **existing** closure branch with a
   **verified non-empty** staged diff containing S-14's mutation.
10. **AC-10** — S-16 reconciles at `expected_status: done` and continues only on `PROCEED`.
11. **AC-11** — `CLOSE_PATH_VERDICT` is read **before** any close call; `safe-close` revalidates
    the binding before any mutation; the close is performed by the **M-14** cascade command and
    the generic shipment-status-move route is never invoked on `017-S`.
12. **AC-12** — **`{post_paths}` is captured at S-19**, immediately on return from the cascade
    call, before P-007 verification, post-mode, any commit, cleanup or restore —
    **unconditionally, on success and on every failure return**. **S-20.3 consumes that capture**
    (R-6 recomputes a live inventory and compares it to `{pre_paths}`, so it is not a consumer of
    `{post_paths}`).
13. **AC-13** — The cascade result is **not committed until every S-20 postcheck passes**.
14. **AC-14** — `019.007-T` is **raw-byte identical between its S-15 `{pre_paths}` capture and the
    S-20 postcheck** (and again at R-7 if an unwind runs). The baseline is **S-15, never PF-6**.
15. **AC-15** — The manifest is **13 members at every point**; set equality holds at **PF-4** and
    at the **S-14 §3.2 condition 3** evaluation.
16. **AC-16 (ONE-WAY DOOR)** — No step attempts to un-claim, release or return `017-S`, and no
    step invokes `shipment return-blocked`, a `backlogit move` on the shipment, or `abandoned`. On
    **every halt at S-3 … S-17, and on any S-18 pre-mutation refusal**, the shipment is left
    **`active`**, **no further backlog mutation is performed after the failure is detected**
    (mutations already legitimately made by earlier steps stand and are recorded), and the session
    halts **to the operator** (Regime B). A halt before the claim, or after a failed claim that
    left `017-S` `queued`, halts **to Stage** (Regime A). **Post-invocation S-18 failures are
    governed by §8.2**, where R-1…R-8 legitimately add revert commits and run `backlogit sync`.
17. **AC-17 (P-009)** — Both the S-10 and S-22 merges are **true merge commits**; squash and
    rebase merges are forbidden. `{merge_sha}` has exactly two parents, parent 1 = mainline,
    parent 2 = the reviewed PR head, verified at S-11. **P-009 governs these PR merges only** —
    the single-parent `{cascade_commit_sha}` reverted at R-4 is not a P-009 signal.
18. **AC-18 (CHANGED-PATH CONTROL)** — S-10 and S-22 each satisfy **§6.1**: no changed path
    matches the **D1** deny-list; `{impl_paths}` is a **subset of the D2 closed three-path set**;
    **P2, P3, P5 and P6** hold; and a **P4 shortfall or an unexpected-but-not-denied path is
    RECORDED, not fatal** (**D5**). S-20.3 separately confines `{post_paths} △ {pre_paths}` to the
    manifest plus the `017-S` record. **The exhaustive allowlist is withdrawn and is not an
    acceptance criterion** — re-imposing it here would mandate a post-claim halt exactly where D5
    mandates record-and-continue.
19. **AC-19 (ROLLBACK)** — Every rollback limb targets **`{cascade_commit_sha}`** (or
    `{quarantine_sha}` at R-3) and **never `{closure_merge_sha}` or `{merge_sha}`**, so the S-15
    baseline `{pre_cascade_sha}` survives every unwind and R-6/R-7 remain **reachable**. R-5
    verifies `{closure_merge_sha}`'s two-parent shape as **evidence only**, runs P-011/P-016
    prechecks before creating its branch, never executes on `main`, uses a **plain revert (no
    `-m`)** against the single-parent cascade commit, and lands through its own P-014/P-018/P-009
    operator-approved merge-commit PR binding `{revert_merge_sha}`. R-3 quarantines **on the
    closure branch**. A restore completes only when **both** R-6 and R-7 pass.
20. **AC-20 (ESCROW)** — All nine §9 conditions are re-measured at invocation; none is inherited
    from this plan's Stage-time measurements.
21. **AC-21 (BRANCH DISCIPLINE)** — S-1 and S-12 each verify P-011 branch naming, a clean
    worktree, and P-016 worktree classification, failing closed on any prohibited or ambiguous
    extra worktree. No S2 mutation occurs on `main`.
22. **AC-22 (TELEMETRY)** — Every halt corresponding to a named policy performs **all three**
    installed P-005 required actions: broadcast, PR-description compliance annotation (recorded
    `N/A` when no PR exists at that gate), and memory-checkpoint record — each carrying the policy
    ID, failure token, step ID and affected artifact IDs. A one-of-three or two-of-three record
    **fails** this criterion.
23. **AC-23 (P-021)** — Findings raised during S-8 or the S-10 review cycle are dispositioned C1
    (fix now), C2 (capture a deferred stash entry; do not widen this shipment), or C3 (halt to
    Stage). Scope is never widened in place. **A C2 capture carries the full six-field payload**:
    the literal greppable token `DEFERRED SCOPE EXPANSION`; a one-sentence statement of the
    expansion; the C1 citation; the `requires deliberation` flag; **per-field source refs**
    (`task=`, `feature=`, `shipment=`, `PR=`, `thread=` — each present or an explicit `N/A`, never
    fused); and kind plus provisional priority.
24. **AC-24 (PATH CONTROL — DENY-LIST)** — No changed path matches the §6.1 **D1** deny-list at
    S-10 or S-22, and positive invariants **P1–P6** all hold. An unexpected-but-not-denied path is
    **recorded and does not halt** post-claim; a **D1** match or a **P2/P3/P5/P6** breach halts.
    `{impl_paths}` ⊆ the **D2 closed three-path set**, asserted by **membership**, never by
    comparing the set against itself. **D1(e)** covers the governance and control-plane surfaces
    (`docs/plans/**`, `docs/decisions/**`, `.autoharness/**`, `.backlogit/*.yaml`, `scripts/**`,
    `.gitignore`, `go.work*`), so the FROZEN prerequisite plan and the routing table **M-17**
    depends on have actual path-level enforcement rather than an unenforced assertion.
25. **AC-25 (EXPECTED ARTIFACTS)** — The **P4** expected set is present: the queue→archive rename
    for `018.008-T` at S-9 (**M-17**) and for every member reaching a terminal status, plus a
    reconcile report for **each mode that persists one** — `017-S-pre-*.md`,
    `017-S-cascade-close-*.md`, `017-S-post-*.md`. **`classify-close-path` persists no report
    under any invocation (M-19) and none is required or expected of it.** The closure PR
    additionally carries the `docs/closure/`, `docs/memory/` and `docs/compound/` artifacts
    **authored at S-21.5** (**M-21**). None of these is treated as drift, and a P4 shortfall is
    recorded rather than fatal.
26. **AC-26 (P-002 AUTHORIZATION)** — The §4.1 deviation is surfaced to the operator at **S-0,
    before the S-3 claim**, and **S-0 proceeds only on an explicit recorded operator authorization
    captured verbatim; absence, ambiguity or silence HALTS to Stage with zero mutation.** The
    broadcast surface is probed (**P-012**) with a mandatory written fallback. The **full
    three-action P-005 record** is emitted (**M-22**): broadcast, **compliance annotation in the
    S-10 and S-22 PR descriptions**, and memory checkpoint. The plan **does not claim P-002
    conformance** and does not amend P-002. The general conflict is captured as stash
    **`DB12DA37`** and is out of scope.
27. **AC-27 (R-5 EXECUTION SITE)** — Any revert of a landed closure merge is created on a
    dedicated branch (**never on `main`**, per **M-26**/P-010) **after P-011/P-016 prechecks**,
    targets the **single-parent `{cascade_commit_sha}`** with a plain revert, lands via its **own
    reviewed merge-commit PR** under P-014/P-018/P-009 with explicit operator approval, and binds
    `{revert_merge_sha}`. R-6/R-7 equivalence is evaluated against `origin/main` at
    `{revert_merge_sha}`, never against an unmerged branch.
28. **AC-28 (S-22 IS NOT AN R-TRIGGER)** — An S-22 closure-PR path-control violation halts
    **pre-merge to the operator** and never enters R-1…R-8. A **P4** shortfall alone never
    triggers an unwind of a verified close.
29. **AC-29 (P-004 PRODUCER, NOT DEVIATED)** — `harness-ready` is applied **only** by the
    installed `harness-architect` skill (§4.2); Ship never self-applies it. The producer is probed
    before reliance (**P-012**) because the task is already `active` at S-7, and any shortfall
    against P-004's precondition or `018.008-T`'s AC-13 literals means **the label is not applied**
    and the run halts.
30. **AC-30 (SUB-FILE CONTENT LOCK)** — **P6** holds: the `workflow-policies.md` diff is confined
    to the P-010 section and the Amendment Log; the `_stage.agent.md` diff is confined to the
    Role Boundary Git/PR rows, the Step 1.9 section and the Step Sequence Contract checklist line;
    and the P-016 `**Allowed exception (Stage spike/research only)**` paragraph, P-010's
    Orchestrator statement, and the **P-015** section are **byte-for-byte identical** pre/post.
31. **AC-31 (P-020 CLOSURE COMPACTION)** — **S-21.5** invokes the compact-context skill with
    `target: all` and records `compaction: {status}` in
    `docs/closure/017-S-018-F-post-merge-closure.md`. Skipping it is a P-020/P-005 violation that
    leaves the closure incomplete and holds the next shipment under P-001.

## 13. Constitution check

| Principle / policy | Engagement |
|---|---|
| **P-001** single active release unit | `017-S` is the sole active shipment from S-3; S-6 verifies. Regime B's disclosed cost is explicit occupation of this slot pending operator action |
| **P-002** harness-ready precondition | **DISCLOSED DEVIATION — NOT conformance.** Structurally unsatisfiable on any shipment-claim route (**M-29**). Disclosed at **S-0** before the irreversible claim, gated on **explicit recorded operator authorization** (fail-closed), with the full P-005 record. §4.1; **AC-26**; §8.1 S-0 halt row. P-002 is not amended; the general conflict is stash `DB12DA37` |
| **P-004** red phase before implementation | **CONFORMANT — the producer is never bypassed.** `harness-ready` is applied solely by the installed `harness-architect` skill at **S-7**, against P-004's precondition plus `018.008-T`'s AC-13 literals and a `Compilation: PASS` / `Red Phase: CONFIRMED` manifest. §4.2; **AC-6**, **AC-29** |
| **P-005** violation telemetry | All **three** required actions: broadcast, **PR-description compliance annotation**, memory checkpoint. **S-0** (P-002 deviation), §8.1 telemetry clause; **AC-22**, **AC-26**; **M-22** |
| **P-007** archive integrity | Installed Step 6.1(b)/(c)/(d) remedy preserved unmodified and takes precedence (§8.2). Verified as an explicit **S-20** postcheck over `.backlogit/archive/`, before the S-21 commit |
| **P-008** markdown lint | Run before the S-8 policy/agent edits are committed and before the S-21.5 closure artifacts are committed |
| **P-009** merge-commit-only | **S-11**, **S-22**, and **R-5's own revert PR**; **AC-17**, **AC-19**, **AC-27**; scoped away from intra-branch commits at R-4 and from the single-parent cascade commit |
| **P-010** role boundary | Stage authors and gates only; every mutating step is Ship's. **S-8 amends P-010 within `018.008-T`'s recorded scope — authorized work, bounded by §6.1 D2 and P6.** **R-5 never executes on `main`** (**M-26**). §10; **AC-27**, **AC-30** |
| **P-011** branch discipline | **S-1**, **S-12**, and **R-5 step (1)** before its branch creation; **AC-21**, **AC-27** |
| **P-012** tool availability | Broadcast surface probed at **S-0**; `harness-architect` producer probed at **S-7**; `TOOL_OK`/`TOOL_DEGRADED` logged with declared fallbacks. **AC-26**, **AC-29** |
| **P-013** consecutive-failure circuit breaker | **DISCLOSED — OPEN CIRCUIT, operator-authorized re-attempts.** This plan's review lineage has failed **7 consecutive** multi-persona attempts, far past P-013.3's 3-attempt threshold. The P-013.6 escalation was compiled and dispatched once (route `gpt-5.6-sol`/`openai`/`high` — distinct from Stage's own Tier-3 route, so **not** `ESCALATION_DEGRADED`), its findings were consumed, and the circuit is **open**. Every attempt from 4 onward proceeded **only** on an exact, explicitly recorded operator authorization token; Stage never self-authorized a re-attempt, which preserves P-013.6's authority-preservation invariant ("the handoff is for asynchronous or **operator** review"). The P-013 event is recorded through the full P-005 triple (D-J13). **Ship inherits no obligation from this row** — it governs Stage's planning loop, not the closure execution. |
| **P-014 / P-018** review and merge approval | **S-10**, **S-22**, and **R-5's revert PR** — operator-approved merges only |
| **P-015** single-artifact shipment closure (no cascade ship) | **This route's authorizing exception.** The cascade close is permitted only under P-015's verified fully-covered-root exception, classified at **S-17** (`CLOSE_PATH_VERDICT`) and revalidated at **S-18**; **PF-4**/**PF-5** establish the covered-root facts. **P6** keeps the P-015 section itself byte-identical — this plan never amends it |
| **P-016** worktree classification | **S-1**, **S-12**, and **R-5 step (1)**, fail-closed; **AC-21**, **AC-27**. Not a changed-path predicate (§6.1 D1 rationale) |
| **P-020** post-merge context compaction | **S-21.5** invokes compact-context `target: all` and records `compaction: {status}` in the closure artifact. Skipping it is a P-020/P-005 violation that holds the next shipment under P-001. **AC-31** |
| **P-021** finding disposition | **S-10**; **AC-23** with the full six-field C2 payload; stash `11B75632` and `DB12DA37` are the C2 precedents |

## 14. Plan hardening

**Requires plan hardening: YES.** Triggers: (a) an irreversible/destructive action
(`PA-017-CASCADE` archives a 13-member subtree); (b) an operator-approval boundary held in escrow;
(c) high blast radius over backlog state no build step can reconstruct; (d) a rollback path with
zero prior compound learning; (e) a claim-time state transition with no in-band repair; (f) a
one-way-door claim with no unclaim operation.

### 14.1 Protected invariants

| # | Invariant | Violation consequence |
|---|---|---|
| **I-1** | Manifest is exactly **13 members** from PF-4 through S-18 | Invalidates §9 condition 6 and §3.2 condition 3 |
| **I-2** | The §3.3 allowlist is exactly **11 IDs**; no archived descendant is off-list | S-14 S0 anomaly ⇒ halt |
| **I-3** | `018-F` is `queued` at S-2, `active` at S-14, `done` only via S-14/S4 | Pre-claim `RECONCILE_FAIL`, or an ungranted feature write |
| **I-4** | No mutation precedes `CLOSE_PATH_VERDICT` being read | Re-opens the defect class where a verdict is recorded only after the destructive call |
| **I-5** | `{pre_cascade_sha}` exists with a verified non-empty staged diff | Rollback anchor destroyed |
| **I-6** | Nothing outside the manifest is archived or deleted | `HALT — cascade detected, revert required` |
| **I-7** | Prerequisite revision 13 stays **FROZEN**; attempt 8 **FAIL STANDS** | Unauthorized attempt 9 |
| **I-8** | `017-S` is never un-claimed, returned, moved or abandoned | Manifest break or unapproved destructive act |
| **I-9** | No production implementation code changes. `{impl_paths}` stays a **subset** of the §6.1 D2 closed three-path set, verified by membership, never by comparing the set against itself | Unauthorized surface reaches `main` undetected |
| **I-10** | After the S-3 one-way door, an **unexpected-but-not-denied** changed path is recorded and **never halts** (§6.1 D5). Only a D1 deny match or a P2/P3/P5/P6 breach halts | Reintroduces the post-claim stranding the §6.1 redesign removes |
| **I-11** | `harness-ready` is applied **only** by the `harness-architect` skill; Ship never self-applies it | False compliance signal on the exact label P-002 consumes (P-004 violation) |
| **I-12** | Every rollback limb targets `{cascade_commit_sha}`/`{quarantine_sha}`, never `{closure_merge_sha}` or `{merge_sha}`, so `{pre_cascade_sha}` survives and R-6/R-7 stay reachable | Rollback with no success path — guaranteed R-8 escalation on every execution |
| **I-13** | S-0 proceeds only on an **explicit recorded operator authorization**; notification is not a disposition | Crossing the one-way door past a Halt-mandated P-002 gate |

### 14.2 Risky actions

| ProposedAction | ActionRisk | ActionResult | Approval |
|---|---|---|---|
| Cascade close of `017-S` (archives the 13-member subtree) via the M-14 command at **S-18** | **destructive** | **approved** — `PA-017-CASCADE`, operator-authorized 2026-09-13T11:35:43-07:00 | Held in escrow; release condition verified at **PF-2**. All nine §9 conditions re-measured at invocation; any mismatch ⇒ approval **INVALIDATED**, halt |
| `018-F` `active → done` at **S-14/S4** | **moderate** — single reversible status write under a five-condition conjunctive gate | pending | Covered by the §3.2 Role Boundary grant; no separate approval |
| Quarantine/revert commits at **R-3/R-4** | **low** — purely additive, same-branch; no overwrite, no deletion | pending | **In-approval.** Prerequisite §10.3.2 sanctions this mechanism precisely because it uses **only** the `git revert` primitive — `git add -A` on the two record trees, commit, then `git revert --no-edit` on the **same** branch: exactly those actions and nothing else. R-3 reproduces that shape; R-4 adds no new primitive |
| **Revert of a landed closure merge at R-5** | **moderate–high** — creates a branch off `main`, opens a PR, and **lands a merge commit on `main`** | pending | **NOT covered by prerequisite §10.3.2's same-branch shape — this is a new primitive.** Requires its own **explicit operator approval** at R-5 step (5) under **P-014 / P-018 / P-009**, on a branch that is never `main` (**P-010**/M-26), after P-011/P-016 prechecks. **R-8 escalates to the operator on any verification failure** |
| Implementation of `018.008-T` at **S-8** | **low** — scope-locked by §6.1 D2 and P6 | pending | Ordinary task execution |

> `strict_safety.enabled` is `false` in `.autoharness/config.yaml`, so no automated approval broker
> gates these actions. The classification is recorded explicitly anyway; `PA-017-CASCADE`'s
> approval is a **recorded operator decision**, not a broker artifact, and is unaffected by that
> flag.

### 14.3 Discipline applied

Every measurement in §3 names the surface it covers and is credited with **exactly** what its
artifact records — never extended to a different surface. §5.1 states explicitly what neither
probe measures. Where a Stage-time measurement exists, Ship re-measures at invocation rather than
inheriting it.

---

## Plan Review

Verdict ledger. One row per attempt; detail lives in Git history, PR #57 and the review evidence.

| Attempt | Revision | dispatch_mode | Reviewers | decision | reviewing_sha |
|---|---|---|---|---|---|
| 1 | 1 | multi-agent | correctness | FAIL | — |
| 2 | 2 | multi-agent | correctness, scope-boundary | FAIL | — |
| 3 | 4 | multi-agent | correctness, constitution | FAIL | — |
| 4 | 5.1 | multi-agent | correctness, constitution, scope-boundary | FAIL | 79103d9 |
| 5 | 6 | multi-agent | correctness, constitution, scope-boundary | FAIL | 5004c84 |
| 6 | 7 | multi-agent | correctness, constitution, scope-boundary | FAIL | 2f72187 |
| 7 | 8 | multi-agent | correctness, constitution, scope-boundary | FAIL | 394a45f |

**Current status:** revision 8 FAILED attempt 7. Correctness FAIL (2 P0), constitution FAIL (4 P0),
scope-boundary **ADVISORY (0 P0)** — the second consecutive clean scope verdict, again confirming the
13-member manifest holds on every traced path. Revision 9 remediates the confirmed findings.

**One reported P0 was REFUTED by direct measurement.** Correctness F-1 asserted that `git show --stat`
on a two-parent merge emits a combined (`--cc`) diff listing no files, making the S-11 check
unsatisfiable. Measured against `f2d4cf9`: `git show --stat` enumerates all **17** files, identical to
`git diff --stat f2d4cf9^1 f2d4cf9`. The finding and its dependent M-21 evidence challenge are both
withdrawn. Reviewer assertions about tool semantics are verified, not adopted.

**Genuine findings remediated in revision 9:**

* **F-2 (P0)** — S-8 bound `{impl_paths}` with no command named, while the parallel `{harness_paths}`
  binding carried a `:(exclude).backlogit/` filter whose stated rationale applies identically. An
  unfiltered binding ingests the S-3 claim writes and the S-2 reconcile report, so the S-8 membership
  test was unsatisfiable by construction — a guaranteed post-claim halt. Both bindings now share one
  rule (D-H6), and `.backlogit/**` is stated to lie outside D2's domain.
* **C-01 (P0)** — S-7 (×2), S-10 and the §11 S-4 risk row routed **post-claim** halts to Stage, the
  one agent P-010 forbids from disposing of shipments. All now route to the operator under Regime B
  (**D-J11**); Stage re-measurement is advisory input, never the halt target.
* **C-02 (P0)** — the harness-architect producer, on which the entire §4.1 P-002 mitigation rests,
  was first validated at S-7 — **after** the one-way door, with refusal predicted by the plan's own
  text. New **PF-12** validates producer admission pre-claim, where failure costs nothing.
* **C-03 (P0)** — the general telemetry clause and AC-22 carried **one** of P-005's three required
  actions while §13 asserted all three, certifying a shortfall as acceptance-passing. Both now
  enumerate the full triple (**D-J9**), with action 2 recorded `N/A` when no PR exists at that gate.
* **C-05** — installed P-020 states verbatim that a **failed** compaction is NON-BLOCKING. The plan's
  fail-closed default would have inverted it after the merge landed. **D-J12** excludes S-21.5 from
  the fail-closed default and from every R-trigger; only *skipping* the invocation is the violation.
* **F-6** — S-21.5's insertion made the closure branch carry **three** commits, not two; R-5's
  justification clause is corrected. The revert target itself was already right.
* **F-8** — D-C11 claimed a report is persisted "for every mode", contradicted by the installed skill
  and by D-C16. Restated to the three modes that actually persist.

**C-04 — P-013 circuit state, disclosed rather than disputed.** The constitution persona is correct
that the circuit is open and was unrecorded: seven consecutive FAILs, far past P-013.3's threshold.
It is **not** correct that each attempt was an unauthorized re-execution — the P-013.6 escalation was
dispatched once to a genuinely distinct route, and every attempt from 4 onward ran on an exact,
explicitly recorded operator authorization. P-013.6 bars the *agent* from self-authorizing a
re-attempt; it does not bar the operator from directing one. The state is now recorded in §13 and
**D-J13** instead of being left implicit.

**Mechanism-reference closure continues to earn its place.** Applied to the revision-9 mechanisms it
again caught a residue the targeted edits missed — §7's binding-discipline paragraph still carried
the old unqualified `{impl_paths}` wording, an exact F-2 relapse — plus a PF-table ordering error.
Both would have been attempt-8 findings. But note its limit, honestly: the sweep only covers tokens
chosen in advance, and attempt 7's F-2 and C-01 were *themselves* this defect class in mechanisms
nobody had enumerated. The countermeasure narrows the failure mode; it does not close it.

`017-S` is **NOT claimable** until an attempt records `decision: PASS` here and this plan is
present on `origin/main` (PF-2).

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

## C. Measured external-interface matrix

Every row is measured against the **installed** workspace, not assumed. Where a prior closure
supplies empirical evidence, the measurement cites it: shipment `021-S` closed through this exact
route on 2026-09-16, producing cascade commit `17f891f` and closure merge `f2d4cf9` (PR #56).

### C.1 Backlog lifecycle and path routing

| ID | Measured fact | Evidence |
|---|---|---|
| **D-C1** | **The claim is a ONE-WAY DOOR.** `backlogit shipment` exposes exactly `add`, `claim`, `create`, `get`, `list`, `return-blocked`, `ship`. There is **no** unclaim/release/return-to-queued path. The only exits from `active` are `shipped` and `abandoned`. | `backlogit shipment --help`; `unclaim` absent from the whole surface |
| **D-C2** | `shipment return-blocked` returns a **single item** and therefore **breaks the 13-member manifest**. Forbidden on this route. | `backlogit shipment return-blocked --help` |
| **D-C3** | `abandoned` is terminal, destructive, **not** covered by `PA-017-CASCADE`, and has **no operator approval**. Forbidden. Any abandon requires new explicit approval and is outside this plan. | `.backlogit/queue/017-S.md` escrow text |
| **D-C4** | The generic `backlogit move` route **refuses shipment status writes with exit 9**. Never the close path. | measured exit code |
| **D-C5** | `backlogit move <id> --status archived` is **not** an archive operation — it silently no-ops. The archive operation is `backlogit archive <id>`. | Probe 26 |
| **D-C6** | The cascade close command is `backlogit shipment ship <id> --sha <s> --message <m> --author <a>`. Flags are exactly `--sha`, `--message`, `--author`. | `backlogit shipment ship --help` |
| **D-C7** | **No binding parameter exists on either close surface** — CLI flags as in D-C6; registry `ship_shipment.params` = `shipment_id`, `sha`, `message`, `author`. The only surface accepting `classification_binding` is `shipment-reconcile` `mode: safe-close`. | CLI help + `.autoharness/backlog-registry.yaml` |
| **D-C8** | `backlogit link list` takes a **positional** ID; the `--id` flag form no longer resolves on 1.10.1. | `backlogit link list --help` |
| **D-C9** | **`status: done` is ARCHIVE-ROUTED.** `.backlogit/registry.yaml` routes `done`, `accepted`, `rejected`, `archived` → `path: archive`, and `queued`, `active`, `blocked`, `review` → `path: queue`. A `→ done` transition therefore **relocates the record file** from `.backlogit/queue/` to `.backlogit/archive/`, appearing in a diff as a **rename**, not an in-place edit. | `.backlogit/registry.yaml` directories block; empirically `f2d4cf9` shows `.backlogit/{queue => archive}/021-S.md` and `{queue => archive}/022-F.md` |
| **D-C10** | The shipment claim **auto-transitions descendant tasks `queued → active`** as a side effect. The task claim is therefore not a separately orderable event on this route. | Probe 26 case `C2`; recorded as M-9 |

### C.2 Tool-emitted artifacts (the set the plan must expect, not forbid)

| ID | Measured fact | Evidence |
|---|---|---|
| **D-C11** | `shipment-reconcile` persists a report at **`.backlogit/reconcile/{shipment_id}-{mode}-{timestamp}.md`** for every mode it runs. | `SKILL.md` L65, L321 |
| **D-C12** | **`.backlogit/reconcile/` is NOT gitignored** and is version-controlled. `.gitignore` excludes only `.backlogit/hooks_queue.jsonl`, `backlogit.db*`, `checkpoints/`, `runtime/`, `.telemetry-checkpoint.json`. The directory already carries committed reports for `006-S`, `008-S`, `009-S`, `012-S`, `021-S`. | `git check-ignore .backlogit/reconcile/` → exit 1; `git ls-files .backlogit/reconcile/` |
| **D-C13** | **The cascade commit emits exactly two artifact classes:** backlog record changes (queue→archive renames plus archive-side modifies) and the reconcile reports. Measured on `021-S`: cascade commit `17f891f` contained 6 files — `{queue => archive}/021-S.md`, `{queue => archive}/022-F.md`, `archive/022.001-T.md`, and three `reconcile/021-S-{pre,cascade-close,post}-*.md` reports. **Nothing else.** | `git show --stat 17f891f` |
| **D-C14** | **The closure PR carries the cascade commit plus the Step 6.x knowledge work**: `docs/closure/{shipment}-{feature}-post-merge-closure.md`, `docs/memory/*.md`, and `docs/compound/*.md`. Measured on `021-S`: merge `f2d4cf9` totalled 17 files. `docs/closure/` is an established convention holding records from `001-S` onward. | `git show --stat f2d4cf9`; `ls docs/closure/` |
| **D-C15** | **P-005 telemetry has THREE required actions, not two.** P-005 mandates: (1) **broadcast** the violation with policy ID and one-line summary; (2) **include violation details in PR descriptions as compliance annotations**; (3) **record the violation in memory checkpoints** for affected tasks. All three are mandatory. Action (2) means the P-002 deviation MUST appear as a compliance annotation in the S-10 and S-22 PR descriptions. No standalone telemetry *file* is required or emitted; `.backlogit/.telemetry-checkpoint.json` is machine-local and gitignored. **The broadcast surface must be probed (P-012): `TOOL_OK`/`TOOL_DEGRADED` logged, with a mandatory written fallback to the session log and PR description when `agent-intercom` is unavailable**, so the safeguard preceding the one-way door can never silently no-op. | `workflow-policies.md` §P-005, §P-012, §P-017 |
| **D-C16** | `mode: classify-close-path` **exists** in the installed skill's mode enum (`pre \| post \| safe-close \| classify-close-path \| detect-mixed-role`). **It writes NO report file under any invocation** — the installed skill states the simplest conforming implementation omits optional persistence entirely, "which this skill does". Its artifact is therefore never present and must never be required (see D-H5). | `SKILL.md` L56, L432 |

### C.3 Git / PR / merge topology

| ID | Measured fact | Evidence |
|---|---|---|
| **D-C17** | **P-009 is literal**: "All pull request merges MUST use merge commits. Squash merge and rebase merge" are excluded. Empirically the `021-S` closure merge `f2d4cf9` has **exactly two parents** — parent1 `74330d3` (mainline), parent2 `f88d413` (reviewed PR head). | `workflow-policies.md` §P-009; `git rev-list --parents -n 1 f2d4cf9` |
| **D-C18** | **The cascade commit is single-parent.** `17f891f` has parent count 1, so it is a plain `git revert` target — `-m 1` must **not** be used on it. | `git rev-list --parents -n 1 17f891f` |
| **D-C19** | **P-010 forbids Ship committing or pushing directly to `main`.** Any revert of a landed merge must therefore be created on a branch and land through its own reviewed, merge-commit PR. | `workflow-policies.md` §P-010 |
| **D-C20** | Merge gates in force for any PR on this route: **P-014** (local review readiness, verified in order), **P-018** (Copilot review completion and thread resolution when enabled), **P-009** (merge commit). | `workflow-policies.md` §§P-014, P-018, P-009 |
| **D-C21** | **Worktree state is single.** `git worktree list` returns one entry, satisfying P-016's no-parallel-execution requirement. P-011 requires branch-before-mutation. | `git worktree list` |

### C.4 Harness ordering (structurally unsatisfiable — disclosed, not pretended)

| ID | Measured fact | Evidence |
|---|---|---|
| **D-C22** | **P-002's ordering cannot be satisfied on this route.** P-002 states Ship "may only claim and implement a task after the harness-architect has confirmed… red phase", with precondition "the task carries the `harness-ready` label" and gate point "task claiming (Step 3)". But by **D-C10** the *shipment* claim at S-3 auto-activates `018.008-T`, so the task claim necessarily precedes any harness. No ordering of this plan's steps can satisfy P-002. | `workflow-policies.md` §P-002; Probe 26 `C2` |
| **D-C23** | The `harness-architect` **skill is installed** (`.github/skills/harness-architect/`). The producer exists; the ordering does not. | filesystem |
| **D-C24** | **Disposition: this is a DISCLOSED DEVIATION, not compliance.** The general conflict is captured as stash **`DB12DA37`** per P-021 and is out of scope for `017-S`. This plan does not amend P-002 and does not claim conformance. See D-D6. | stash `DB12DA37` |

## D. Close-path authority

| ID | Decision |
|---|---|
| **D-D1** | Installed Ship requires the returned `CLASSIFICATION_BINDING` to be carried into the close call. By D-C7 the direct cascade op accepts no binding. Therefore **`mode: safe-close` carrying a `CASCADE` binding is the only executable realization of installed Ship's own mandate**, not an alternative to it. The same `backlogit shipment ship` operation runs either way. |
| **D-D2** | A **direct, unbound** `backlogit shipment ship` on this route is a **HALT** — it would discard the mandated binding and bypass the Cascade Close Sub-Procedure's gates. |
| **D-D3** | **No prerequisite Ship release unit is required.** Harmonizing the installed Ship text is a clarity improvement, captured out-of-scope as stash entry **`11B75632`**, not a correctness precondition. |
| **D-D4** | This reconciliation is made **falsifiable at preflight** (PF-11). If the cascade op ever gains a binding parameter, the reconciliation is **void** and the route halts to Stage for re-measurement. Drift in the SKILL's bound-`CASCADE` routing also halts. Drift in the *installed Ship prose* toward the guarded path is **recorded, not halted** — a correction of the known-defective clause must not fail the route. |
| **D-D5** | `classify-close-path` is **read-only** and must not write a reconcile artifact or mutate any state before returning its verdict. Its verdict must be read **before any close call**. |
| **D-D6** | **P-002 is DISCLOSED AS UNSATISFIABLE and requires an EXPLICIT RECORDED OPERATOR AUTHORIZATION to proceed — disclosure alone is NOT sufficient.** Per D-C22, the shipment claim at S-3 auto-activates `018.008-T`, so P-002's "task claim after `harness-ready`" ordering cannot hold on any shipment-claim route. **P-002's Violation Action is literally "Halt and suggest running the harness-architect."** No disposition in the installed registry authorizes proceeding past a Halt-mandated gate on notification; every override this registry defines is an explicit recorded authorization (P-001 `skip_policy`, P-014 merge approval, P-018 audited force, P-012 declared degradation, `PA-017-CASCADE` escrow). This plan therefore: (1) states the deviation explicitly rather than omitting the policy; (2) surfaces it to the **operator before S-3** and **requires an explicit recorded operator authorization to proceed**, captured verbatim in the run log — **absence, ambiguity, or silence ⇒ HALT to Stage, zero mutation (Regime A)**; the authorization reuses the existing `skip_policy: P-002`-shaped acknowledgement rather than minting a new token; (3) emits a P-005 record per D-C15 with `violation_policy: P-002`, `gate: S-3`, `skip_policy: P-002`; (4) preserves P-002's **protective intent** via D-D7, not via a self-applied label; (5) does **not** amend P-002 and does **not** assert conformance. The general conflict is captured as stash **`DB12DA37`** and is out of scope (D-M8). |
| **D-D7** | **P-004's PRODUCER CONTRACT IS NOT DEVIATED FROM. The `harness-ready` label is applied ONLY by the installed `harness-architect` skill.** P-004's Violation Action is literally "Do NOT apply `harness-ready` label", and P-002 names the harness-architect as the **producer** (Ship is only the consumer). Ship self-applying the label would emit a **false compliance signal on the exact label P-002 consumes** — worse than omitting it. S-7 therefore **invokes the installed `harness-architect` skill** for `018.008-T`; the skill applies the label at its own Step 6 after its Step 5.1/5.2 checks. The pass criterion adopts P-004's precondition **and** `018.008-T`'s recorded AC-13 literals verbatim: `go vet ./...` exits 0; `go test ./tests/integration/ -run TestStageBranchGate_ContractCorrected` exits non-zero; the captured output contains the literal marker `not implemented: stage branch gate contract` attributed to that function; the H0 red count is literally 1 of 1; and the harness manifest records `Compilation: PASS` and `Red Phase: CONFIRMED`. Any shortfall ⇒ **do not apply the label** and HALT (Regime B). Only P-002's *ordering* is deviated from (D-D6); P-004 is satisfied in full. |

## E. Happy path (canonical sequence)

**Preflight gates `PF-1…PF-11`** run in order on the pre-claim branch, before the claim. Any
failure halts with **zero mutation**.

| ID | Gate |
|---|---|
| **PF-1** | Engine identity: version + binary SHA-256 match the recorded digest. Mismatch invalidates every measurement ⇒ halt to Stage. |
| **PF-2** | This plan resolves on `origin/main`. The body above `## Plan Review` is **byte-identical** to the reviewed revision, compared as `git show <reviewing_sha>:<plan path>` where `reviewing_sha` is read from the **`reviewing_sha` column** of the winning attempt row. That row's `decision` cell is `PASS`, its `Reviewers` cell names the persona coverage, and no later row supersedes it. **No literal in-row marker tokens are required** — the ledger is a table and renders values as cells, so demanding literal `dispatch_mode:` / `decision: PASS` strings made this gate unsatisfiable by construction against the plan's own format. |
| **PF-3** | `backlogit doctor` ⇒ `No issues found.`, exit 0. |
| **PF-4** | Manifest = 13; descendant graph ∪ `{018-F}` set-equal to manifest; every `parent_id` resolves transitively. |
| **PF-5** | The archived-declaring descendant set is **exactly** the 11-ID allowlist, each satisfying §3.3 conditions 1–4. Any off-list archived descendant ⇒ halt. |
| **PF-6** | Linked-deliberation set for all 13 members verified **EMPTY** by three independent methods. Records `019.007-T`'s pre-state **for provenance only — informational, never an assertion baseline**. |
| **PF-7** | Read-only re-verification of the two committed probe artifacts and the Probe 26 script, plus engine agreement. Authors nothing, mutates nothing, adds no backlog item. |
| **PF-8** | Topology gate `--phase pre_claim` ⇒ exit 0. |
| **PF-9** | Ground 1: `021-S` `archived` **and** `archived_status: shipped` with commit provenance. `archived` alone is insufficient — it is also the terminal state of an abandoned shipment. |
| **PF-10** | Ground 2: the **installed** `_ship.agent.md` carries Step 6.1(a1) with the S0–S5 selector and `workflow-policies.md` carries the §3.2 grant. |
| **PF-11** | Close-path authority agreement per D-D4. |

**Steps `S-0…S-22`,** plus `S-21.5`. Monotonic, each ID unique. `S-0` is the pre-claim P-002 deviation gate (D-D6); `S-21.5` authors the closure knowledge artifacts and discharges P-020.

| ID | Step | Binds |
|---|---|---|
| **S-0** | **P-002 deviation gate (pre-claim, before S-1).** Probe the broadcast surface (P-012); disclose the D-D6 deviation to the operator; emit the three-action P-005 record (D-C15). **Pass criterion: an explicit recorded operator authorization to proceed, captured verbatim. Absence, ambiguity or silence ⇒ HALT to Stage, zero mutation (Regime A).** Mutates nothing. | — |
| **S-1** | Pre-claim branch created; P-011 branch discipline and P-016 worktree classification verified; clean worktree. | — |
| **S-2** | Pre-claim reconcile `mode: pre`, `expected_status: queued`. Continue only on `PROCEED`. | — |
| **S-3** | **Claim `017-S`** via the CLI surface `backlogit shipment claim 017-S`. **One-way door.** | — |
| **S-4** | Re-read `018-F`; **must be exactly `active`**. The claim performs this transition; no agent writes it. | — |
| **S-5** | Archived-member non-drift check over all 11: archive-located, `status: archived`, `archived_status` unchanged, `parent_id` unchanged. | — |
| **S-6** | Topology gate `--phase post_claim` ⇒ exit 0. | — |
| **S-7** | Harness first: generate the failing harness for `018.008-T`, observe and record **red before any implementation line**. | `{harness_paths}` |
| **S-8** | Implement `018.008-T` to green, scope-locked to its recorded scope. | `{impl_paths}` |
| **S-9** | `018.008-T → done`. Discharges §3.2 conditions 1 and 2. | — |
| **S-10** | Implementation PR: open, review, P-018 engagement, operator-approved **merge commit**. §H changed-path control asserted (D-H1 deny-list, D-H2 membership, P1-P6). | — |
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
| **S-21.5** | **Author the closure knowledge artifacts and discharge P-020.** Write `docs/closure/017-S-018-F-post-merge-closure.md`; invoke the **compact-context** skill with `target: all` (P-020 makes this MANDATORY at every post-merge closure — skipping it is a P-005 violation that leaves the closure incomplete and holds the next shipment under P-001); record `compaction: {status}` in the closure artifact. These are the `docs/closure/`, `docs/memory/`, `docs/compound/` paths D-H5 expects in the S-22 PR; without this step P4 would require artifacts no step authors. | — |
| **S-22** | Closure PR: P-014 local review gate, operator-approved **merge commit**, verified on `origin/main`. §H changed-path control asserted (D-H1 deny-list, D-H2 membership, P1-P6). | `{closure_merge_sha}` |

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
| `{closure_merge_sha}` | **S-22** | R-5 (two-parent evidence check only — NEVER the revert target, I-12) |
| `{quarantine_sha}` | **R-3** | R-3 |
| `{revert_merge_sha}` | **R-5** step (4) | R-6, R-7 |

**Binding discipline.** `{harness_paths}` is bound from `git status --porcelain=v1 -- . ':(exclude).backlogit/'` (D-H6) so the claim's uncommitted backlog writes are never ingested. `{impl_paths}` is bound as the exact changed-path set of S-8 **and then tested for membership** in the D-H2 authorized surface — the binding alone is not the check (I-9).

## G. Evidence capture timing and anchors

| ID | Decision |
|---|---|
| **D-G1** | `{pre_paths}` is captured at **S-15** against `{pre_cascade_sha}`: full sorted path+blob-hash inventory of `.backlogit/queue/` + `.backlogit/archive/`, plus raw `019.007-T` frontmatter. |
| **D-G2** | `{post_paths}` is captured at **S-19**, **immediately on return from the cascade call** — before P-007 verification, before `mode: post`, before any commit, cleanup or postcheck-driven restore. |
| **D-G3** | **The S-19 capture is UNCONDITIONAL** — taken on every return, success or failure. Rationale: installed Step 6.1(c)/(d) restores sit between the close and post-mode and mutate exactly the inventoried trees, so a later capture would compute the scope assertion over an already-restored tree. |
| **D-G4** | The `019.007-T` untouched-assertion is anchored to the **S-15** baseline at both S-20 and R-7. **Never to PF-6** — the S-9 dependency carve-out permits `019.007-T` to change legitimately between PF-6 and S-15, so a PF-6 anchor guarantees a false-positive unwind. |
| **D-G5** | Record equivalence is read from **raw frontmatter in the artifact files** — never a list, index or query view. |

## H. Changed-path control (deny-list + positive invariants)

**Design decision.** The exhaustive-allowlist formulation (`anything not enumerated ⇒ HALT`) is
withdrawn. It required perfectly predicting every path a multi-tool chain touches, and its failure
mode was catastrophic: a halt **after** the one-way door, stranding `017-S` `active` with no
unclaim. Three of attempt 5's seven P0s were instances of that single choice. It is replaced by a
**deny-list plus positive invariants**, which is robust to unanticipated tool emissions because an
unforeseen artifact is no longer automatically fatal.

| ID | Decision |
|---|---|
| **D-H1** | **FORBIDDEN SURFACES (deny-list).** A changed path is a violation if and only if it matches one of: (a) production implementation code — `internal/**`, `cmd/**`, any `*.go` that is not a `*_test.go` file, `go.mod`, `go.sum` (the `*_test.go` carve-out applies uniformly, including under `internal/` and `cmd/`, though `018.008-T`'s recorded scope independently forbids any harness there); (b) `.github/policies/**`, `.github/agents/**`, `.github/skills/**`, `.github/instructions/**` or `.github/workflows/**` **other than** the two surfaces `018.008-T`'s recorded scope authorizes (see D-H2); (c) any `.backlogit/queue/**` or `.backlogit/archive/**` record **not** belonging to the 13 manifest members, the `017-S` record, or the evidenced `019.007-T` delta (D-H4); (d) any deletion of a `.backlogit/reconcile/**` report; (e) **governance and control-plane surfaces** — `docs/plans/**` and `docs/decisions/**` (including this contract, the governing closure plan, and the FROZEN prerequisite `docs/plans/2026-09-13-intercom-go-ship-feature-completion-decided-plan.md`), `.autoharness/**`, `.backlogit/registry.yaml`, `.backlogit/header-def.yaml`, `.backlogit/config.yaml`, `scripts/**`, `.gitignore`, `go.work*`. Nothing else is fatal by enumeration. **Rationale for (e):** a deny-list is fail-open by design, so every surface whose immutability this plan asserts elsewhere (§M out-of-scope list, I-7's FROZEN prerequisite, the routing table D-C11 depends on) must be named here or that assertion has no path-level enforcement at all. |
| **D-H2** | **AUTHORIZED IMPLEMENTATION SURFACE — a CLOSED set of exactly three paths.** `018.008-T`'s recorded scope authorizes exactly: (1) `.github/policies/workflow-policies.md`, (2) `.github/agents/_stage.agent.md`, (3) `tests/integration/stage_branch_gate_test.go` — the harness file named literally by the task's recorded `harness_cmd` and AC-13. The task records `FILE COUNT: 3 files total … 0 new production files`, so `|{impl_paths} ∪ {harness_paths}| ≤ 3` and `{impl_paths}` MUST be a **subset** of that closed set — a membership test, not a self-comparison. This closes both the tautology (`{impl_paths}` compared against itself) and the undecidability of an unenumerated "the test harness file". |
| **D-H3** | **POSITIVE INVARIANTS (all must hold; each is independently falsifiable).** **P1** no path matches the D-H1 deny-list. **P2** the manifest is exactly 13 members before and after (D-C13 scope check). **P3** every backlog record change belongs to a member, `017-S`, or the evidenced `019.007-T` delta. **P4** the **expected** artifacts are PRESENT, scoped per D-H5. **P5** no production implementation code changed — asserted **independently of the changed-path enumeration** by a repo-scoped command (`git diff --name-only {merge_base}...HEAD -- internal/ cmd/ go.mod go.sum` returns empty, and no non-test `*.go` appears in the full diff), so P5 can fail even where the path-set enumeration itself is wrong. **P6 (SUB-FILE CONTENT LOCK)** — D-H1/D-H2 test *paths*, but `018.008-T`'s recorded scope binds *regions within* those files. P6 asserts by diff: changed hunks in `.github/policies/workflow-policies.md` fall **only** within the P-010 section and the Amendment Log; changed hunks in `.github/agents/_stage.agent.md` fall **only** within the Role Boundary Git/PR rows, the Step 1.9 section and the Step Sequence Contract checklist line; and the P-016 paragraph beginning `**Allowed exception (Stage spike/research only)**`, P-010's Orchestrator statement, and the P-015 section are **byte-for-byte identical** pre/post. Without P6 an edit to P-015, P-009 or P-016 inside an authorized file passes every other gate, while §M forbids it and escrow condition 9 leans on it. |
| **D-H4** | **`019.007-T` carve-out.** `019.007-T` (`status: blocked`, `dependencies: [018.008-T]`, parent `019-F`, **outside the manifest**) may legitimately unblock when `018.008-T` reaches `done`. Its delta is permitted when evidenced in the run log by a `018.008-T → done` transition immediately preceding it, confined to `status`, `dependencies` and **record metadata timestamps (`updated_at`)**. The timestamp is included because any engine write updates it; confining the carve-out to status/dependency fields alone would make a *legitimate* unblock match D-H1(c) and **halt post-claim**, reintroducing the exact stranding failure mode §H exists to remove. The run-log evidencing is the operative condition. |
| **D-H5** | **EXPECTED-ARTIFACT SETS, measured (D-C13/D-C14), used to satisfy P4 — never as a deny-list.** **Cascade commit (S-21):** backlog record changes (queue→archive renames, archive-side modifies) **and** the reconcile reports for the modes that persist one. **Closure PR (S-22):** the cascade commit plus `docs/closure/017-S-018-F-post-merge-closure.md`, `docs/memory/*.md`, `docs/compound/*.md`, authored at S-21.5. **P4 is scoped to modes that PERSIST a report** — `pre`, `safe-close`/`cascade-close`, and `post`. Per D-C16 the installed skill writes **no** report for `classify-close-path` under any invocation, so requiring one would make P4 unsatisfiable by construction. The measured report filenames are `017-S-pre-*.md`, `017-S-cascade-close-*.md` and `017-S-post-*.md`; a literal `{mode}` substitution would mis-predict the cascade filename. A path in these sets is **expected**; a path outside them is **reported and recorded** but is only fatal if it matches D-H1. |
| **D-H6** | **`{harness_paths}` binding excludes backlog state.** Bind from `git status --porcelain=v1 -- . ':(exclude).backlogit/'`. The S-3 claim writes `.backlogit/` records that no step has committed yet; an unfiltered `git status` would ingest them and make the gate unsatisfiable by construction. `.backlogit/` paths are governed solely by P3. |
| **D-H7** | **S-20 cascade-scope assertion.** `{post_paths} △ {pre_paths}` is confined to the 13 manifest members and the `017-S` record. Any **out-of-manifest artifact archived or deleted** ⇒ `HALT — cascade detected, revert required`. This is a true deny condition and is retained: it detects the destructive over-reach the escrow exists to bound. |
| **D-H8** | **POST-CLAIM FAILURE SEMANTICS OF THIS CONTROL.** Before S-3 (pre-claim): any violation halts cheaply to Stage. After S-3 and before S-18 (post-claim, pre-mutation): a **D-H1 deny match or a P2/P3/P5/P6 breach** halts, leaving `017-S` `active` per §J Regime B; an **unexpected-but-not-denied path (P4 shortfall only)** is **recorded in the run log and does NOT halt** — this is the specific change that prevents an unforeseen tool emission from stranding the shipment. After S-18 (post-mutation): §K governs; a P4 shortfall alone never triggers an unwind of a verified close. **Worktree/parallel-branch prohibition is NOT a changed-path predicate** and is therefore not a D-H1 clause: a path set derived from `git status`/`git show --stat` in the current worktree can never name a path in another one, so as a deny clause it would be vacuous. It is enforced where it is actually decidable — the P-016 fail-closed worktree classification at S-1 and S-12, re-verified before any branch creation including R-5's. |

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
| **D-J8** | R-1…R-8 trigger **only** on a failure from S-18 onward **once the cascade op has actually been invoked**: a post-invocation S-18 failure, an S-19 capture failure, an S-20 postcheck failure, an S-21 commit failure, or out-of-manifest archival discovered at any later point. Every halt at PF-1…S-17, and any **S-18 pre-mutation refusal**, is governed by the regimes instead. **An S-22 closure-PR path-control violation is a PRE-MERGE halt to the operator, never an R-trigger** — the closure merge is operator-gated, so path control is asserted before the merge, and a documentation-shaped violation must never unwind a correct, verified, operator-approved close. |
| **D-J9** | **Violation telemetry:** every halt corresponding to a named policy emits a P-005 event recording the policy ID, failure token, step ID and affected artifact IDs; the closure record carries it. |
| **D-J10** | **P-021 disposition:** findings raised during S-8 implementation or the S-10 review cycle are classified C1 (in-scope, fix now), C2 (out-of-scope ⇒ capture as a deferred stash entry with source refs), or C3 (defect in this plan ⇒ halt to Stage). Scope is never widened in place. |

## K. Rollback (post-cascade only)

| ID | Decision |
|---|---|
| **D-K1** | Restoration is achieved **only by adding commits** — never by working-tree overwrite of backlog records and never by file deletion. This ban is scoped to **R-1…R-8 only**. |
| **D-K2** | The installed `_ship.agent.md` Step 6.1(b)/(c)/(d) P-007 remedy (`git restore .backlogit/archive/`) **remains in force, unmodified**, and takes precedence where it applies. This plan has no authority to override an installed policy. |
| **D-K3** | **R-1 Quarantine:** stop, mutate nothing further, do not commit the cascade result, record the failure token verbatim. `{post_paths}` is already captured at S-19. |
| **D-K4** | **R-2 Classify — evaluated in this order, first match wins.** (d) committed with any S-20 postcheck **failed or unrun** ⇒ **out of contract: HALT to operator after R-1, no automatic revert** (an automatic unwind over state of unknown provenance is itself unsafe); (a) uncommitted — including an S-19 capture failure or an S-21 commit failure ⇒ R-3; (b) committed with **every S-20 postcheck passed** and the closure branch unmerged ⇒ R-4; (c) committed, all postchecks passed, and the closure merge landed ⇒ R-5. Evaluating (d) first resolves the former overlap in which a committed-unmerged-postchecks-failed state matched both (b) and (d) with contradictory dispositions. |
| **D-K5** | **R-3 Uncommitted:** use the prerequisite §10.3.2 mechanism **on the closure branch**: `git add -A -- .backlogit/queue/ .backlogit/archive/`, commit as `{quarantine_sha}`, then `git revert --no-edit {quarantine_sha}` on that same branch. A `quarantine/017-S-<utc>` name may be a **ref** pointing at `{quarantine_sha}` for evidence — never a separate commit target. |
| **D-K6** | **R-4 Committed, unmerged:** target `{cascade_commit_sha}`; plain `git revert --no-edit`. A single parent is **expected** here and is not a P-009 signal. |
| **D-K7** | **R-5 Committed, closure merge landed. The revert target is `{cascade_commit_sha}`, NEVER `{closure_merge_sha}`.** The closure branch carries **two** commits: the S-15 pre-cascade baseline `{pre_cascade_sha}` (which contains S-14's `018-F → done`) and the S-21 cascade commit `{cascade_commit_sha}`. `{closure_merge_sha}` contains **both**, so `git revert -m 1 {closure_merge_sha}` would unwind the baseline as well as the cascade — returning `018-F` to `active`/queue-located and making D-K9/D-K10 **unsatisfiable by construction**, since `{pre_paths}` was captured *after* the baseline commit. R-5 therefore: (1) **P-011/P-016 prechecks first** — `git status --short` clean and `git worktree list --porcelain` classified fail-closed, **before any branch creation**; (2) verify `{closure_merge_sha}` empirically as **evidence only**: exactly two parents, parent 1 = mainline, parent 2 = the reviewed closure-PR head by **SHA equality**; any other shape ⇒ P-009 violation, HALT to operator, no automatic revert; (3) **execution site (P-010): never on `main`** — `git checkout main && git pull && git checkout -b chore/017-s-cascade-revert-<utc>`; (4) `git revert --no-edit {cascade_commit_sha}` **on that branch** — `{cascade_commit_sha}` is single-parent (D-C18), so **plain revert, never `-m 1`**; (5) land it through its own revert PR under the full gate set — P-014, P-018, explicit operator approval, **P-009 merge commit** — binding `{revert_merge_sha}`; (6) D-K9/D-K10 are evaluated **against `origin/main` at `{revert_merge_sha}`**, never against an unmerged branch; (7) any failure at (5) ⇒ HALT to operator; Ship never lands the revert by any other path. **This limb lands a merge on `main` under a new explicit operator approval — it is a moderate-to-high risk action, not a low additive one.** |
| **D-K8** | **`{merge_sha}` is NEVER a rollback target.** It is the S-10 implementation merge; reverting it would unwind `018.008-T`'s implementation rather than the cascade. |
| **D-K9** | **R-6 Tree equivalence:** recompute the inventory and require it set-equal, path-for-path and blob-hash-for-blob-hash, to `{pre_paths}`. **`{pre_paths}` is the post-baseline state (`018-F` already `done`), so every revert target across R-3/R-4/R-5 must unwind the cascade ONLY and must leave `{pre_cascade_sha}` intact.** Evaluation site: the working tree for R-3/R-4; `origin/main` at `{revert_merge_sha}` for R-5. |
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
| **D-L5** | The nine conditions: (1) released by a review-passed Stage closure plan present on `origin/main` supplying execution site, gates, baseline and rollback — **verified at PF-2**; (2) `classify-close-path` returns `CASCADE` before any close call — S-17; (3) `safe-close` revalidates the binding before any mutation — S-18; (4) `021-S` shipped — PF-9; (5) parent-completion capability live — PF-10; (6) manifest unchanged at 13 — PF-4; (7) pre-cascade baseline and **pre-invocation** inventory exist and are verified — S-15 (`{pre_cascade_sha}` + `{pre_paths}` only; `{post_paths}` cannot be a precondition of the call that produces it); (8) linked-deliberation set for this manifest verified EMPTY by this plan's own measurement — PF-6; (9) no scope expansion — **PF-4** live manifest invariant, **D-H1/D-H2** deny-list and closed-set membership, **P6** sub-file content lock, S-10/S-22 changed-path control, S-20.3 cascade-scope assertion, and the §M out-of-scope list. *(PF-7 is read-only probe re-verification over two committed artifacts and performs no live manifest check; anchoring condition 9 to it was a mis-anchor. The live manifest invariant is PF-4, which condition 6 also binds.)* |
| **D-L6** | Escrow release is **conditional, not self-certified**: the release condition is satisfied only when this plan is review-PASSed and present on `origin/main`, as verified at PF-2. |

## M. Out of scope

| ID | Decision |
|---|---|
| **D-M1** | Claiming `017-S` (Stage does not claim), implementing `018.008-T` (Ship work at S-8), and merging either PR without explicit operator authorization. |
| **D-M2** | Widening or re-litigating `PA-017-CASCADE`. |
| **D-M3** | Amending P-015, `_ship.agent.md`, `shipment-reconcile/SKILL.md`, or **any policy/agent surface other than the two that `018.008-T`'s recorded scope authorizes** (D-H2): `.github/policies/workflow-policies.md` (P-010 Stage bullets and Amendment Log only) and `.github/agents/_stage.agent.md` (Role Boundary Git/PR rows, Step 1.9, Step Sequence Contract line). **Amending P-010 within that recorded scope is the authorized work of S-8 and is explicitly IN scope** — the prior blanket prohibition contradicted the step this plan mandates. |
| **D-M4** | Adding any backlog item to `018-F` or `017-S` — forbidden by the manifest invariant. |
| **D-M5** | Harmonizing the installed Ship CASCADE-branch prose — captured as stash **`11B75632`** (D-D3). |
| **D-M6** | Decision §7 follow-ups 1, 2, 3 and 5 — recorded, off-route, explicitly not current blockers. Item 4 alone is pulled in, as PF-7. Item 5 is referenced at S-2 as a **disclosure only**, with no derived halt condition. |
| **D-M7** | Re-opening prerequisite attempt 9. Revision 13 stays FROZEN. |
| **D-M8** | **Resolving the general P-002 / shipment-claim ordering conflict** — captured as stash **`DB12DA37`** (D-D6). It affects every shipment-claim route in this workspace and would require amending P-002 or changing claim semantics; either exceeds this shipment's frozen scope. `017-S` proceeds under the disclosed deviation. |

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
| **I-9** | No production implementation code changes. `{impl_paths}` stays a **subset** of the D-H2 **closed three-path set**, verified by membership, never by self-comparison. P5 is asserted repo-scoped, independently of the changed-path enumeration. |
| **I-10** | After the S-3 one-way door, an **unexpected-but-not-denied** changed path is recorded and never halts (D-H8). Only a deny-list match or a P2/P3/P5/P6 breach halts. |
| **I-11** | `harness-ready` is applied **only** by the installed `harness-architect` skill (D-D7). Ship never self-applies it — that would be a false compliance signal on the exact label P-002 consumes. |
| **I-12** | Every rollback limb targets `{cascade_commit_sha}` (or `{quarantine_sha}` at R-3), **never** `{closure_merge_sha}` or `{merge_sha}`, so the S-15 baseline `{pre_cascade_sha}` survives every unwind and D-K9/D-K10 remain reachable. |
| **I-13** | S-0 proceeds only on an **explicit recorded operator authorization** (D-D6). Notification is not a disposition; P-002's Violation Action is *Halt*. |

---

**Contract status:** canonical and internally consistent as of 2026-09-16. The governing plan is
remastered from this contract. Any divergence between the two is a plan defect, not a contract
defect.

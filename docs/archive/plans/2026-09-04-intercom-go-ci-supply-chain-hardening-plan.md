---
title: "Implementation Plan — CI supply-chain and gate integrity hardening (007-F)"
date: 2026-09-04
status: reviewed
phase: residual-hardening
feature: 007-F
source_deliberation: "docs/decisions/2026-09-04-intercom-go-residual-hardening-triage-deliberation.md"
resolves_stash: [EFFAA358, 90350C9A, F7C6420D]
partially_resolves_stash: [BEDD2E70]
moved_to_008F: [B6203CCA]
---

# Implementation Plan — CI supply-chain and gate integrity (007-F)

## Problem Frame

Five deferred findings all target the **CI trust root**: the pipeline that
validates every later shipment. Under dark-factory (unattended) operation this
pipeline is the only thing standing between a bad change and `main`, so its own
integrity is load-bearing.

Verified current state (post-005-S):

* `.github/workflows/ci.yml:325` — `pip install autoharness`, **floating latest**.
* `.github/workflows/ci.yml:210` — `gitleaks detect --source . --no-git -v`,
  **working tree only**; committed history is never scanned.
* `.github/workflows/ci.yml:145` — `bash scripts/check-retired-architecture.sh`,
  the **bare scan only**; the script's `--self-test` mode (implemented at
  `check-retired-architecture.sh:273`) is never invoked by CI.
* `.github/CODEOWNERS` — **absent**; the topology-check job's own comments
  document a self-modifying-PR gap with no ownership control.
* `.gitignore` invariant **I6 (append-only)** is enforced by reviewer diligence
  only — no mechanical check.

## Non-Goals (scope fence)

* No Go source changes. This shipment touches CI configuration and one new
  checker script only.
* **Does NOT flip `PIPELINE_TOPOLOGY_GATE_REQUIRED` to required**, and does NOT
  modify branch-protection settings (see R1 — this is the plan's single most
  dangerous adjacent action).
* Does not alter the scanner logic inside `check-retired-architecture.sh` —
  that is 008-F's scope. This shipment only *invokes* it.

## Constitution Check

Uses the ratified numbering (see the source deliberation's correction section).

| Principle | Status | Units | Justification |
|---|---|---|---|
| I. Safety-First Go | N/A | — | No Go code in this shipment. |
| II. Test-First Development (NON-NEGOTIABLE) | Satisfied | U4a, U4b | The `.gitignore` checker's fixtures and `--self-test` (U4a) are authored before the detection logic (U4b). |
| III. Workspace Isolation and Security Boundaries | Satisfied | U1, U2 | Version pinning and history secret-scanning both tighten the supply-chain boundary. |
| IV. CLI Workspace Containment (NON-NEGOTIABLE) | Satisfied | U4b | Checker operates on repo-relative paths only; no absolute-path or out-of-tree access. |
| V. Structured Observability | Satisfied | U4b, U5 | Every new gate emits an explicit pass/fail line naming the invariant it enforces and the resolved base/HEAD SHAs. |
| VI. Single Responsibility | Satisfied | all | One gate per unit; fixtures (U4a) separated from logic (U4b) and from CI wiring (U5). |
| VII. Destructive Command Approval (NON-NEGOTIABLE) | N/A | — | No destructive command is executed. (The earlier draft misclassified CODEOWNERS here; ownership metadata is not a destructive operation.) |
| VIII. Explicit Safety Modes for Elevated Risk | Satisfied | U3 | U3 AC-0 requires probing the *effective* ruleset before creating CODEOWNERS and aborting if code-owner review is already enforced — an explicit pre-flight safety check on an elevated-risk change, not merely refraining from an action. |
| IX. Git-Friendly Persistence | Satisfied | all | Text config and one shell script; line-oriented, diff-friendly. |
| X. Agent Context Efficiency | Satisfied | all | Each unit is independently reviewable. |
| XI. Merge Commit History Preservation (NON-NEGOTIABLE) | Satisfied | U2 | U2 **reads** history only (`fetch-depth: 0` + scan); no rewriting, no force-push, no history mutation anywhere in this plan. |

## Implementation Units

### U1 — Pin the `autoharness` CI install to an exact version
*Resolves `EFFAA358`. Artifact class: CI config. Size S, complexity low.*

Replace the floating `pip install autoharness` at `ci.yml:325` with a
**hash-pinned constraints install**, not merely an `==` version pin.

Adversarial review correctly noted that `pip install autoharness==X.Y.Z` is
*not* hermetic: it still fetches an unhashed distribution and resolves
transitive dependencies at install time, so a compromised dependency or a
mutable index can still execute arbitrary code in the runner. A version pin
alone narrows the window but does not close the execution path.

Add a committed, reviewed `.github/constraints/autoharness-lock.txt` containing
exact versions **and SHA-256 hashes** for `autoharness` and every transitive
dependency, installed with `--require-hashes`.

**Acceptance criteria**
1. No floating (`==`-less) install of `autoharness` remains in any workflow.
2. The install uses `--require-hashes` against a committed lock file pinning
   exact versions and SHA-256 hashes for the full transitive closure.
3. Ship resolves the currently-working version at implementation time (e.g. via
   `pip index versions autoharness`) and records it in the commit message; the
   pin must reproduce the version CI resolves today so the job's behaviour is
   unchanged.
4. An adjacent comment states the maintenance expectation (bump deliberately,
   regenerate hashes, never float).
5. The topology-check job still passes on a clean tree.
6. If the full hashed closure cannot be generated in-session, land AC-1's exact
   `==` pin as an interim improvement and **record the residual
   non-hermeticity explicitly in the PR body** — do not silently claim
   supply-chain closure that was not achieved.

### U2 — Add a scheduled full-history secret scan
*Resolves `90350C9A`. Artifact class: CI config. Size S, complexity low.*

Add `.github/workflows/secret-scan-history.yml`: a `schedule`-triggered
(weekly) plus `workflow_dispatch` job running gitleaks **with** git history
(i.e. without `--no-git`), using `fetch-depth: 0`. Reuse the existing pinned,
checksum-verified gitleaks install (version `8.30.1`, sha256
`551f6fc8…`) exactly as `ci.yml:197-208` does.

**Acceptance criteria**
1. New workflow exists with both `schedule` (`cron: '0 6 * * 1'`, weekly Monday)
   and `workflow_dispatch` triggers.
2. Checkout uses `fetch-depth: 0`; the scan command omits `--no-git`.
3. gitleaks is installed via the same pinned version (`8.30.1`) and the same
   **full** sha256
   (`551f6fc83ea457d62a0d98237cbad105af8d557003051f41f3e7ca7b3f2470eb`) as
   `ci.yml:197-208` — written literally, no floating action, no unpinned
   download.
4. The per-PR `--no-git` scan at `ci.yml:210` is left **unchanged** (this
   complements it; it does not replace it).
5. **No secret material may be disclosed.** The scan runs with `--redact=100`;
   no raw gitleaks report is uploaded as an artifact; and any stash entry
   created from a finding records only rule ID, fingerprint, commit SHA, and
   file path — **never the matched value**. This is an acceptance criterion, not
   guidance: copying `ci.yml:210` verbatim does not redact.
6. This workflow is **not** added to the PR-required check set. A first-ever
   history scan can surface pre-existing findings, and blocking all delivery on
   historical debt would halt the pipeline; it alerts without blocking.
7. Because the operator is AFK, a finding is routed to the operator handoff as a
   **high-priority blocked stash entry** requiring an explicit rotation
   decision — not left as a silently red scheduled job.

### U3 — Add CODEOWNERS ownership metadata for CI-critical paths
*Partially resolves `BEDD2E70`. Artifact class: config. Size XS, complexity trivial.*

Create `.github/CODEOWNERS` assigning ownership of all CI-critical paths to the
repository owner (`@softwaresalt`).

**Scope claim corrected after adversarial review.** Two independent reviewers
established that this unit does **not** close the self-modifying-PR gap. The
`pull_request` workflow executes from the PR head, so a PR can modify `ci.yml`,
the topology check, or the new gates and make its own required status succeed;
advisory CODEOWNERS has no enforcement effect against this. This unit therefore
ships **ownership metadata** — a prerequisite for a future control, not the
control itself. The residual gap is carried to the operator handoff as an
explicit open decision rather than reported as resolved.

**Acceptance criteria**
0. **PRE-FLIGHT (blocking).** Before creating the file, probe the *effective*
   protection on `main` — repository **and** organisation rulesets and branch
   protection (e.g. `gh api repos/:owner/:repo/rulesets`,
   `.../branches/main/protection`). If "Require review from Code Owners" is
   already enabled anywhere in the effective set, **abort and return this unit
   blocked.** With no CODEOWNERS file such a rule matches no owner; creating one
   would suddenly require `@softwaresalt` approval on every CI-touching PR and
   **deadlock unattended merges** — harness-level `merge_approval_pre_authorized`
   does not satisfy a GitHub review requirement.
1. `.github/CODEOWNERS` covers `/.github/workflows/`, `/.github/CODEOWNERS`,
   `/scripts/ci-topology-check.sh`, `/scripts/check-retired-architecture.sh`,
   and the new `/scripts/check-gitignore-append-only.sh` — all CI-critical
   scripts, or none (partial ownership is a misleading control).
2. Every owner reference resolves to a real account with repo access, verified
   mechanically (`gh api users/softwaresalt`). A dangling owner silently
   disables the rule, producing the appearance of a control with none of the
   effect.
3. A header comment states that this file is **advisory metadata only** until
   branch protection enables code-owner review, that flipping it is an operator
   action, and that doing so first requires provisioning an approval path for
   unattended PRs.
4. **No branch-protection or ruleset setting is modified by this unit.**

### U4a — `.gitignore` append-only fixtures and self-test harness
*Resolves `F7C6420D` (part 1 of 3). Artifact class: tests/fixtures. Size S, complexity low.*

Add committed fixture pairs under `scripts/testdata/gitignore/` plus the
`--self-test` driver, authored **before** the detection logic (Principle II).

**Precise invariant definition** (review found "reordered" undefined):

> I6 holds iff the old file's non-comment, non-blank lines appear as an ordered
> **subsequence** of the new file's non-comment, non-blank lines.

Insertions anywhere pass; deletions and reorders fail. One mechanical
predicate, not four ad-hoc cases.

**Acceptance criteria**
1. Three fixture pairs exist: pure insertion → pass; deletion → fail; reorder →
   fail. (No-change is subsumed by the subsequence rule.)
2. `--self-test` runs every fixture pair and reports each by name.
3. Fixtures fail against an absent/stub implementation, proving they test
   something.

### U4b — `.gitignore` append-only checker logic
*Resolves `F7C6420D` (part 2 of 3). Artifact class: script/code. Size S, complexity medium. Depends on U4a.*

Add `scripts/check-gitignore-append-only.sh` implementing the subsequence
predicate above.

**Acceptance criteria**
1. Exits non-zero on deletion and on reorder; zero on insertion and no-change.
2. All U4a fixtures pass.
3. The comparison base is **explicit**, not implied: on `pull_request` it is
   `${{ github.event.pull_request.base.sha }}`; the script prints the resolved
   base and HEAD SHA so the comparison is auditable (Principle V).
4. **Fail-closed in CI.** If the base cannot be resolved, the script **fails**.
   A skip exists **only** behind an explicit `--allow-no-merge-base` flag that
   CI never passes. (The earlier draft exited zero on an unresolvable base —
   review correctly identified that as a fail-open hole in the very invariant
   the gate exists to enforce.)

### U5 — Wire the `.gitignore` checker into CI
*Resolves `F7C6420D` (part 3 of 3). Artifact class: CI config. Size XS, complexity trivial. Depends on U4b.*

Add a CI step invoking `scripts/check-gitignore-append-only.sh`.

**Acceptance criteria**
1. The step runs on `pull_request` and fails the job on a deletion/reorder.
2. The step explicitly fetches the PR base SHA (or sets `fetch-depth: 0`) on
   **that job only**, so the base resolves without slowing every other job.
3. The target job is named explicitly in the diff — placement is not left to the
   executor's judgement.
4. A PR that only appends to `.gitignore` passes.

### U6 — MOVED to 008-F

`B6203CCA` (wiring `check-retired-architecture.sh --self-test` into CI) was
originally this plan's U6. **Adversarial review found the placement created an
unmergeable ordering.** Wiring `--self-test` as a job-failing gate here, before
008-F changes the scanner's masking logic, means 008-F's test-first fixtures —
which must fail against the current implementation — could never be committed
without reddening the gate 007-F had just installed. Test-first and
"install the harness first" cannot both hold across a shipment boundary.

The unit is therefore relocated to **008-F as its final unit**, landing after
the logic and fixtures it gates. `B6203CCA` is dispositioned to 008-F.
## Dependency Graph

```
U1  U2  U3                (independent)
U4a ──► U4b ──► U5        (fixtures, then logic, then CI wiring)
```
No cross-shipment dependency remains. 007-F may land entirely on its own.

## Post-Review Remediation Record (2026-09-04)

Adversarial multi-persona review (Security ×2, Scope, Constitution,
Maintainability) returned **FAIL** on the pre-remediation draft. Changes:

| Finding | Severity | Remediation |
|---|---|---|
| **U6 placement made 008-F's test-first fixtures unmergeable** — wiring `--self-test` as a blocking gate before 008-F's logic change means fixtures that must fail against current masking could never be committed. | **blocker** | U6 **moved to 008-F** as its final unit. `B6203CCA` re-dispositioned to 008-F. 007-F now has no cross-shipment dependency. |
| **U3 was claimed to close the self-modifying-PR gap; it does not.** `pull_request` runs from the PR head, so a PR can still alter the workflow and satisfy its own required check. Advisory CODEOWNERS has no enforcement effect. | **blocker** | Disposition narrowed to "ownership metadata"; residual gap escalated to the operator handoff as an open decision. |
| **U3 could deadlock unattended merges** if any repo/org ruleset *already* enables code-owner review — creating the file would newly match an owner. | **blocker** | AC-0 added: probe the effective ruleset first and **abort blocked** if enforcement is live. |
| **U1's `==` pin is not hermetic** — unhashed distribution plus transitive resolution still permits arbitrary code execution in the runner. | **blocker** | Rewritten to `--require-hashes` against a committed lock file; AC-6 requires explicitly disclosing residual non-hermeticity if the full closure cannot be produced. |
| **U4 AC-3 was fail-open** — exiting zero on an unresolvable base defeats the invariant the gate enforces. | major | Now fail-closed in CI; skip only behind `--allow-no-merge-base`, which CI never passes. Base SHA made explicit and logged. |
| U2 lacked disclosure controls; "alert the operator" is meaningless with an AFK operator. | major | `--redact=100`, no artifact upload, metadata-only stash records, explicit cron, and routing to a high-priority **blocked** handoff item. |
| "Reordered" was undefined; U4 exceeded the 2-hour rule. | major | Defined as an ordered-subsequence predicate; split into U4a (fixtures) / U4b (logic). |
| U5 left job placement and `fetch-depth` unspecified. | major | Target job must be named; base fetch scoped to that job only. |
| Constitution table: truncated titles; VII/VIII misclassified. | major | Rewritten with exact ratified titles; VII → N/A, VIII → justified by AC-0. |

## Risks and Caveats

**R1 — CODEOWNERS could deadlock unattended operation (HIGHEST RISK IN THIS PLAN).**
If branch protection is ever set to "Require review from Code Owners", every
dark-factory PR touching `.github/workflows/**` would block awaiting a human
review that, by definition, is not coming. `merge_approval_pre_authorized=true`
does **not** override a GitHub branch-protection rule. Mitigation: U3 ships the
file **advisory-only** and explicitly forbids modifying branch protection. The
enforcement flip is recorded as a deliberate operator decision, to be taken only
alongside a decision about how unattended PRs obtain code-owner approval.

**R2 — A dangling CODEOWNERS owner fails open, silently.** GitHub ignores rules
whose owner cannot review, producing the *appearance* of a control with none of
the effect. U3 AC-2 requires verifying the handle resolves.

**R3 — The full-history scan may surface pre-existing historical findings.**
U2 scans history for the first time and could fail on something committed long
ago. This is signal, not noise — but it is a *new* red build on a schedule.
Mitigation: the job is `schedule`/`dispatch` only and is **not** wired into the
PR-blocking path, so a historical finding raises an alert without blocking
delivery. Any finding it surfaces becomes a new stash entry for operator triage,
never an in-flight code change.

**R4 — Merge-base unavailability.** U4 AC-3 makes a missing merge-base an
explicit skip rather than a failure, so shallow-clone environments cannot cause
spurious red builds.

**R5 — Version pin staleness.** U1 trades supply-chain risk for maintenance
burden; the pin will drift. Accepted deliberately: an unreviewed floating
install into CI is the larger exposure, and U1 AC-2 records the bump
expectation.

## Plan Hardening Signals

| Signal | Present | Note |
|---|---|---|
| Schema/contract change | No | — |
| Security-sensitive behaviour | **Yes** | Supply-chain pinning, secret scanning, ownership controls. |
| Migration | No | — |
| External dependency | **Yes** | `autoharness` package version; gitleaks binary. |
| Could block all future PRs | **Yes** | R1 — new CI gates plus CODEOWNERS. |

**Requires plan hardening: yes.**

---

# Plan Hardening

**Hardening required: YES** (security-sensitive; external dependency; could block all future PRs). Hardened 2026-09-04.

## Context consulted

`.github/workflows/ci.yml` (lines 76, 145, 197-210, 324-325), `scripts/check-retired-architecture.sh` (`--self-test` at L273-282), `.gitignore` (I6 invariant), absence of `.github/CODEOWNERS`, and the 001-S/004-S closure records that deferred these five findings.

## Protected invariants

* **I6 — `.gitignore` is append-only.** U4/U5 mechanise it; they must not themselves reorder the file.
* **Existing CI topology must keep passing.** No unit may change the topology-check job's contract.
* **`PIPELINE_TOPOLOGY_GATE_REQUIRED` stays at its current (non-required) setting.**
* **Per-PR `--no-git` gitleaks scan at `ci.yml:210` remains untouched.** U2 complements, never replaces.
* **Unattended merge capability must survive this shipment.** Nothing here may introduce a human-review dependency into the PR path.

## Risky actions (ProposedAction / ActionRisk)

### PA-1 — Introduce `.github/CODEOWNERS` (U3)
**Risk: HIGH (deadlock potential).** If branch protection later requires code-owner review, unattended PRs block forever awaiting a human. `merge_approval_pre_authorized=true` is a *harness* pre-authorisation and has no effect on a GitHub branch-protection rule.
**Control:** file ships advisory-only; branch protection is explicitly out of scope; header comment records the enforcement flip as an operator decision requiring a companion decision about unattended approval. **Ship must not enable code-owner enforcement under any circumstance in this shipment.**

### PA-2 — Add two new PR-blocking CI gates (U5, U6)
**Risk: MEDIUM.** A false positive in either gate blocks all delivery.
**Control:** U4 AC-3 makes an unresolvable merge-base a *skip*, not a failure. U6 wraps an already-proven self-test path. Both are validated against a clean tree before landing.

### PA-3 — Scan full git history for secrets (U2)
**Risk: MEDIUM (disclosure-adjacent).** A finding may surface a historical credential. Any such finding must be reported as a **stash entry for operator triage**, never remediated in-flight, and **no secret value may appear in logs, artifacts, PR text, or backlog entries**.
**Control:** job is schedule/dispatch-only and never PR-blocking, so a historical finding raises an alert without halting delivery.

### PA-4 — Pin an external package version (U1)
**Risk: LOW.** A wrong pin breaks the topology-check job.
**Control:** pin to the version already resolving today; verify the job passes before landing.

### Not risky — explicitly classified
U1's comment addition; U3's file creation in isolation; U4's script when unwired.

## Rollback points

Each unit is independently revertible. U5 and U6 are single CI steps (revert = delete the step). U3 is a single file (revert = delete). U2 is a standalone workflow (revert = delete). No unit mutates shared state that another unit depends on except U4→U5.

## Deepened runtime verification

1. Clean-tree run: all six gates green.
2. Negative control per gate: `.gitignore` line removal fails U5; a tampered retired-architecture fixture fails U6. **A gate that has never been observed failing is unverified.**
3. Confirm the CODEOWNERS handle resolves to an account with repo access.
4. Confirm the scheduled scan does **not** appear in the PR-required check set.

## Human checkpoints

* **Operator decision required before** enabling code-owner enforcement or flipping `PIPELINE_TOPOLOGY_GATE_REQUIRED` — both out of scope here.
* **Operator triage required for** any historical secret surfaced by U2.

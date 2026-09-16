# Stage-gate disposition — operator-accepted residuals (2026-09-14)

| | |
|---|---|
| Owner | Stage |
| Plan | `docs/plans/2026-09-13-intercom-go-ship-feature-completion-decided-plan.md` |
| Plan revision | **13 — FROZEN** (no revision 14) |
| Shipment | `021-S` — `queued`, **NOT claimed** |
| Feature / task | `022-F` / `022.001-T` |
| Last gate that RAN | **true-lineage attempt 8**, 2026-09-14 |
| Gate decision | **FAIL** — P0 = 0, P1 = **26 raw / 14 deduplicated**, **7/7 personas FAIL** |
| Gate verdict | `docs/reviews/2026-09-13-true-lineage-attempt-8-verdict.md` (unedited) |
| Attempt 9 | **NOT authorized, not requested, not run** |
| Stage state | **`stage-ready-under-operator-residual-acceptance`** |
| Review gate | **FAIL — NOT a PASS, NOT relabelled** |
| Operator confirmation | `ok, make it so` |

## The disposition

The operator reviewed the attempt-8 failures and directed: **stop iterating the plan; freeze the
current engineering invariants; explicitly accept the documentation / gating-evidence findings as
non-blocking residuals; move toward Ship with a failing test harness covering the implementation
invariants; record this as an operator-approved Stage-gate disposition rather than pretending
attempt 8 passed.**

This record is that disposition. It is a **procedural gate override with acknowledged residuals**,
not a passing gate.

## Attempt-8 truth that remains visible

* Attempt 8 **ran once** and **FAILED**: P0 = 0, P1 = 26 raw / 14 deduplicated, 7/7 personas FAIL,
  anchor Architecture Strategist (`gpt-5.6-sol`, high).
* `D-1` was **incorporated** across the producer, refusal, consumer, authority, compatibility,
  test and tooling surfaces — and **several semantic / evidence surfaces remained stale**. The
  defect class **relocated** to the verification and measurement surfaces rather than closing.
* The key technical observation is the **direct `backlogit_ship_shipment` engine call** at
  `_ship.agent.md` L822. It is an **implementation consumer that must be guarded or replaced**.
  It is **not** deemed an architectural feasibility blocker.
* `N-1`…`N-14` are **open**. **None is asserted fixed.** The attempt is **not** relabelled PASS.
* **No attempt 9 is authorized** and none was run.

## Frozen implementation invariants (§12.1 `IV-1`…`IV-5`) — BINDING on Ship

1. **`IV-1`** — Classify the close path before any mutation.
2. **`IV-2`** — Bind the classification to the manifest and dependency snapshot, and reject drift.
3. **`IV-3`** — Require the binding on every mutating close path, including the direct `backlogit_ship_shipment` engine invocation, or replace that invocation with the guarded path.
4. **`IV-4`** — Fail closed on missing, invalid, ambiguous, mixed, stale or non-matching classifications.
5. **`IV-5`** — The TDD harness proves that refusal paths perform zero mutation and that valid `CASCADE` / `SAFE_CLOSE` paths preserve intended behaviour.

Each invariant is a **single unwrapped line** so it is quotable character-for-character
(anti-`N-6` rendering rule, plan §12.1).

**Blocking rule.** Residual acceptance covers **plan-text, documentation and gating-evidence
findings only**. **Any failing implementation invariant is BLOCKING for Ship regardless of any
accepted plan-evidence residual**, and is not dischargeable by editing the plan.

**Ordering.** Test-first is **NON-NEGOTIABLE**: the harness encoding `IV-1`…`IV-5` lands **RED**
before any instruction-surface file is edited, and is **GREEN** before closure. A green-on-arrival
harness is itself an `IV-5` failure.

## Residual classification (full adjudication in plan §13.7)

| Category | Count | Themes |
|---|---|---|
| **(a)** implementation-guard acceptance criterion — **blocking** | **8** | `N-1`, `N-4`(a), `N-5`, `N-9`, `N-10`, `N-11`, `N-12`, `N-13` |
| **(b)** non-blocking documentation / evidence mismatch — **accepted** | **7** | `N-2`, `N-3`, `N-4`(b), `N-6`, `N-7`, `N-8`, `N-14` |
| **(c)** truly unresolved implementation blocker | **0** | — |

`N-4` splits across (a) and (b); 14 themes yield 15 entries.

**Category (c) is empty by adjudication, not by arrangement.** The nearest candidate, `N-4`(a),
is a real measured open defect that is nonetheless implementable — hence `IV-3`. No finding was
rewritten, downgraded or hidden. A `(c)` discovered at build time blocks Ship without further
operator consultation.

### Theme → invariant map (category (a))

| Theme | Frozen as |
|---|---|
| `N-1` R5 branch-binding contradiction | `IV-4` + `IV-5` |
| `N-4`(a) direct engine cascade call, L822 | **`IV-3`** |
| `N-5` non-independent falsifiability | `IV-5` |
| `N-9` claim-predicate unsatisfiable on resume | `IV-4` |
| `N-10` binding integrity / serialization / lock scope | `IV-2` + `IV-4` |
| `N-11` ADD-1 ↔ CS10 equivalence untested | `IV-5` |
| `N-12` drift and bound-refusal merged in recovery | `IV-4` + `IV-5` |
| `N-13` descendant enumeration via point lookup; MCP parity | `IV-1` + `IV-4` |

## Residual risk accepted

| Risk | Severity | Acceptance rationale |
|---|---|---|
| Plan text (§4.2 `CS11`/`CS3`, `AC-35`, `AC-39`, §13.4, §15, `T-13` scope) remains internally inconsistent | moderate | `IV-1`…`IV-5` **govern over** any conflicting plan statement (§12.1). The harness, not the prose, is authoritative at build time |
| §4.0 self-certification (`AC-39`) is weaker than §4.0 itself | low | Self-audit apparatus only; superseded as the acceptance surface by `IV-1`…`IV-5` |
| Verbatim-wording provenance (`N-6`) unreliable | low | Invariants are semantic; the harness must not be derived from §4.2 wrapped literals |
| In-merge edit detection (`N-7`) has no execution site | low | Covered by ordinary PR review and CI gates on PR #54 and the implementation PR |
| Caller-inventory claims are three-file but stated workspace-wide (`N-4`(b)) | moderate | Restated as a **guard obligation** (`IV-3`), not a completed enumeration |

## Boundary honoured this session

* **No runtime, source, skill, agent or policy file modified.** `.github/agents/_ship.agent.md`,
  `.github/policies/workflow-policies.md` and `.github/skills/shipment-reconcile/SKILL.md` are
  **untouched** — they remain the implementation surface for Ship.
* **No shipment claimed. No PR merged. No code implemented. No attempt 9 run.**
* `021-S` status `queued`, manifest `[022-F, 022.001-T]`, dependencies unchanged.
* `017-S` remains blocked on all of its own independent grounds.
* Single worktree; no parallel branch; commit without amend or history rewrite.

## Next owner

**Operator** — review and merge PR #54 to `main`.

Ship **MUST NOT** be invoked until PR #54 is merged to `main` **and** the staging artifacts are
verified present on `origin/main`. Ship's first action thereafter is the **RED** `IV-1`…`IV-5`
harness, before any instruction-surface edit.

# Deliberation — Verified Task-Only Shipment Finalization (A10EF3D0)

- **Date:** 2026-09-12
- **Stage route:** `claude-opus-5` (tier 3), resolved from fresh `.autoharness/config.yaml` (`schema_version: 1.1.0`)
- **Chartered scope:** EXACTLY one stash entry — `A10EF3D0` (`kind: deliberation`, `priority: critical`). No other stash entry triaged, edited, harvested or archived.
- **Measured at:** clean worktree, branch `chore/stage-task-only-shipment-finalization`, base `fdff9e4` (fresh `origin/main`)
- **Blocks:** closure of shipment `017-S` (and every future SAFE_CLOSE-classified shipment)
- **Upstream refs:** `docs/decisions/2026-09-12-intercom-go-deferred-split-and-readiness-lock-decision.md` §1.6 / OQ-2; `docs/plans/2026-09-11-intercom-go-stage-artifact-branch-pr-policy-gap-plan.md` §R16.3.3

---

## 1. Problem frame

`shipment-reconcile` Safe-Close Mode step 8 instructs the closer to move the
shipment record to `shipped` with the generic, non-cascading path:

```text
backlogit move <shipment_id> --status shipped
```

On backlogit 1.10.1 that call is **refused by the engine**. Consequently every
SAFE_CLOSE-classified shipment halts fail-closed at step 8 with
`RECONCILE_FAIL_SHIPMENT_RECORD_LIVE_STATUS` and can never reach
`archived_status: shipped`.

The contract forbids both obvious substitutes:

- archiving a still-`active` record yields `archived_status: active` →
  `RECONCILE_FAIL_SHIPMENT_RECORD_PROVENANCE`;
- calling the cascade op after a SAFE_CLOSE classification is a **P-005**
  process deviation.

This is **workspace-wide and shape-independent** under the current contract: the
only CASCADE-qualified shape must include the covering feature (P-015
precondition 1), which then fails Ship Step 6 pre-mode (`expected_status: done`)
because no Ship step ever moves a covering **feature** to `done`. So the
CASCADE escape hatch is unreachable in practice, and SAFE_CLOSE is a dead end.

### 1.1 Executed evidence (this session, disposable sandbox)

All probes ran in a throwaway backlogit workspace under `%TEMP%`, **never** in
this repository, and the sandbox was deleted afterwards. The repository tree was
verified clean before and after.

**Probe 1 — the blocker is real.**

```text
backlogit move 001-S --status shipped
→ exit 9
→ "shipment must be shipped via ShipShipment, not a direct status update"
```

**Probe 2 — SAFE topology (task-only manifest, live parent feature outside it).**

Manifest `[001.001-T]`; parent `001-F` live and **not** a member.

```text
backlogit shipment ship 001-S  → exit 0
archived_ids = ["001.001-T", "001-S"]
returned_ids = []
```

Post-state:

| Property | Observed |
|---|---|
| `returned_ids` | `[]` |
| `archived_ids` | exactly the task + the shipment record — nothing else |
| parent `001-F` | still in `queue/`, `status: active` |
| parent `001-F` `updated_at` | **byte-identical to pre-close** — the file was never rewritten |
| shipment `001-S` | in `archive/`, `status: archived`, **`archived_status: shipped`** |

**Probe 3 — UNSAFE topology (the hazard the guards must exclude).**

Manifest `[002.001-T]`, but `002.001-T` owns a subtask `002.001.001-ST`
(`status: queued`) that is **omitted** from the manifest.

```text
backlogit shipment ship 002-S  → exit 0
archived_ids = ["002.001.001-ST", "002.001-T", "002-S"]
```

The omitted, still-`queued` subtask was **silently archived** and force-stamped
`archived_status: done`. This is a genuine unintended cascade and is exactly the
P-015 hazard. It **proves the omitted-descendant pre-call guard is mandatory**,
not defensive boilerplate.

**Probe 4 — omitted SIBLING topology (run to resolve a plan-review P1).**

Manifest `[001.001-T]`; an omitted sibling `001.002-T` and its subtask, all under
the same parent `001-F`.

```text
backlogit shipment ship 001-S  → exit 0
archived_ids = ["001.001-T", "001-S"]
returned_ids = ["001.002-T", "001.002.001-ST"]
```

The sibling **files survived** in `queue/` — no collateral **archiving** — but
`returned_ids` was **non-empty**. Note the probe did **not** verify whether the
returned records' `parent_id` was preserved, so this topology is
**uncharacterized**, not established as safe. This establishes that the two
hazards are structurally distinct:

- descendants **of a manifest item** are **ARCHIVED** (Probe 3) — a correctness hazard;
- other live descendants **of the parent feature** are **RETURNED** (Probe 4) —
  uncharacterized, and it makes `returned_ids` non-empty.

Consequence: `returned_ids == []` is **only** a sound post-guard when the safe
topology is tightened so the parent feature has **no live out-of-manifest
descendants**. Under a looser topology the guard would fire on a manifest the
looser guards would themselves have authorized. The plan's guard G5 is defined
accordingly, comparing against the **non-archived manifest subset** so that
pre-archived members remain supported.

**Incidental finding (load-bearing).** `backlogit move <task> --status done`
relocates the record into `archive/` while its declared `status` stays `done`.
Location is therefore **not** evidence of "truly archived" — confirming the
existing skill rule that the declared-status snapshot must be read from
frontmatter, never inferred from directory.

---

## 2. Options considered

| # | Option | Verdict |
|---|---|---|
| O-1 | Wait for backlogit to expose a non-cascading shipment-record close | **Rejected** — unbounded external dependency; `017-S` and every future SAFE_CLOSE shipment stay unclosable meanwhile. No upstream commitment exists. |
| O-2 | Relax P-015 to allow the cascade generally | **Rejected** — Probe 3 shows the cascade silently archives omitted descendants. A general grant re-opens precisely the corruption P-015 exists to prevent. |
| O-3 | Drop the `archived_status: shipped` provenance requirement | **Rejected** — destroys the audit property that distinguishes a shipped shipment from an abandoned one; weakens post-mode verification for all shipments. |
| O-4 | Add a generic `move shipment → shipped` capability | **Rejected** — out of scope (engine change), and explicitly excluded by charter. |
| **O-5** | **Authorize a narrow, machine-verified TASK-ONLY finalization path, gated by pre/post-call guards, fail-closed to the existing SAFE_CLOSE/HALT behavior** | **CHOSEN** |

## 3. Chosen direction (O-5)

Introduce a **new, narrowly classified** close path — `TASK_ONLY_FINALIZE` —
into `shipment-reconcile`, authorized **only** after machine-verifying a
precisely-defined safe topology. It sits alongside, and never replaces,
`SAFE_CLOSE` and the existing P-015 fully-covered-root `CASCADE` exception.

**Non-negotiable boundaries** (carried verbatim into the plan):

1. Does **not** expand Ship's Role Boundary.
2. Does **not** add a generic `move shipment → shipped`.
3. Does **not** bypass Ship Step 6 pre-mode.
4. Does **not** authorize an unrestricted cascade.
5. Does **not** weaken the existing fully-covered-root P-015 exception —
   that exception's text is preserved byte-for-byte.

### 3.1 Safety contract (summary; normative form lives in the plan)

**Pre-call guards** — ALL must hold, else fail closed:

- every manifest member is `artifact_type: task` (no feature, no subtask, no shipment);
- every manifest task shares **one** `parent_id`, and that parent is a feature;
- that parent feature is **outside** the manifest and is live in `queue/`;
- every manifest task is terminal-executable (`done`, or truly `archived`);
- **complete parent-rooted live coverage**: walking the **parent feature's** full
  `parent_id` graph at every depth yields no live (non-truly-archived) descendant
  outside the manifest — i.e. `live_descendants(parent) == manifest`. This single
  guard subsumes both Probe 3 (omitted descendants of manifest items) and Probe 4
  (omitted siblings), and is what makes `returned_ids == []` sound;
- no out-of-manifest record declares this `shipment_id` (orphan scan);
- a clean `.backlogit/` baseline is captured so rollback is bounded;
- pre-close declared-status + `parent_id` + full parent frontmatter snapshot
  captured from frontmatter.

**Post-call guards** — ALL must hold, else HALT and revert:

- `returned_ids == []` (sound only under the tightened G5 topology above);
- `archived_ids ⊆ allowed_ids` (manifest tasks + shipment record) — nothing else;
- `required_ids ⊆ archived_ids`, where the shipment record is a `required_ids`
  member **unconditionally**;
- **independent filesystem diff** against the baseline — the engine's
  `archived_ids` is never trusted as the sole detector;
- parent feature still live in `queue/`, full frontmatter **byte-identical**;
- no torn records (nothing in both `queue/` and `archive/`, nothing missing from both);
- shipment record `archived_status: shipped`;
- pre-archived members handled explicitly (no re-archive, no false "missing").

**Fail-closed:** any deviation in topology or evidence → fall back to the
existing SAFE_CLOSE/manual path, or HALT, with a baseline-derived rollback that
also removes newly created untracked `archive/` files. Never "best effort".

### 3.2 Authorization surface

Authorizing this path requires amending **P-015 itself** in
`.github/policies/workflow-policies.md`, not only the skill: P-015 states the
prohibition and names its sole exception, so a skill-only edit would leave a
compliant Ship agent obligated to refuse the call. The amendment adds a
**second, strictly narrower** exception and leaves the existing fully-covered-root
exception byte-for-byte unchanged.

### 3.3 Self-hosting closure order (full ordered sequence in plan §8)

The prerequisite must close **itself** using the very path it introduces:

1. Prerequisite implementation PR **merges** to `main` (Ship).
2. Ship verifies the merge SHA, checks out **fresh `main`**, and cuts the closure branch.
3. Ship **reloads** the merged P-015 + `shipment-reconcile` skill and **positively
   verifies** the new tokens and merge commit are present.
4. Pre-mode → `TASK_ONLY_FINALIZE` close → P-007 handling → post-mode.
5. Closure PR merged, operational closure, then **P-020** compaction.
6. Only then does `017-S` become eligible.

Step 3 is sound because the prerequisite's own shipment is deliberately built in
the exact safe topology: the covering feature has **exactly one descendant at any
depth** (the sole task), the feature is root and excluded from the manifest, the
task is `done` at closure time, the shipment record is `active`, and no
out-of-manifest record declares the shipment. Hence
`live_descendants(parent) == manifest` holds trivially and `returned_ids == []`
follows. The full thirteen-step sequence, with actors, is normative in plan §8;
the no-fallback contingency is plan §8.1.

---

## 4. Sequencing — resolved (rev 2)

The rev-1 blocker ("dependency and ID allocation are unsatisfiable on this
branch") was an artifact of staging on an `origin/main`-based branch. It is
**dissolved**, not deferred, by harvesting on the policy branch instead.

**Root fact:** backlogit ID allocation is **branch-local**, because `.backlogit/`
is tracked, branch-scoped state. `017-S`, `018-F` and stash `A10EF3D0` exist on
`chore/stage-pipeline-policy-gap`. Harvesting there makes the dependency edge and
the stash archival writable at once, with no collision and no renumbering.

**The rev-1 recommendation is WITHDRAWN.** It proposed two sequential Stage PRs:
land the policy-gap records first, then re-stage the prerequisite from fresh
`main`. Independent adversarial review returned `MUST_REPLAN` against it for a
decisive reason: that sequence lands `017-S` on `main` **before** any blocking
dependency exists, opening a window in which Ship could legitimately claim `017-S`
with no prerequisite in place. The two-PR split is both unsafe and unnecessary.

**Adopted:** a **single artifacts-only Stage PR (#54)** on the policy branch
carrying the policy-gap records **and** the prerequisite feature/task/shipment
**and** the `017-S depends_on <prerequisite shipment>` edge. The edge lands
atomically in the same merge as `017-S`, so no eligibility window exists at any
reachable state. Gates and the Ship PR-lifecycle-only actor boundary are
normative in plan §14 and §11.

## 5. Open questions

- **OQ-1:** **RESOLVED.** Sequencing is the single artifacts-only Stage PR in §4;
  harvest proceeds on the policy branch.
- **OQ-2:** Should `TASK_ONLY_FINALIZE` also permit a manifest whose tasks span
  **multiple** protected parent features? Deliberately answered **no** for now —
  single-parent keeps the protected set trivially computable. Revisit only on a
  fired trigger.
- **OQ-3 (rev 2):** Should multi-**live**-task manifests be supported? Answered
  **no** for this first authorization: no executed probe establishes their
  safety, so plan G10 excludes them.

## 6. Stash lineage

`A10EF3D0` is harvested in this session on the policy branch and archived only
after lineage verification against the created feature, task and shipment IDs.

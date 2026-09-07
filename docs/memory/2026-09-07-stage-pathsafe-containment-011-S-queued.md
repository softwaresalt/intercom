# Stage Session — 011-S / 012-F pathsafe Containment Correctness

- **Date**: 2026-09-07
- **Agent**: Stage (P-017 dark-factory, DARK_MODE_ACTIVE, operator AFK)
- **Outcome**: **1 shipment queued** — `011-S`
- **Constraint honored**: exactly ONE shipment from the highest-value coherent grouping

## Result

| Field | Value |
|---|---|
| Shipment | **`011-S`** (status `queued`) |
| Covering feature | `012-F` |
| Tasks | `012.001-T` … `012.009-T` (9) |
| Manifest | 10 items, parent-first, dependency-ordered, `unsized=0` |
| Size composition | XS=1, S=7, M=1 |
| Estimated effort | ~9 × 2h = **18h** human-equivalent |

## Stash disposition

**Consumed and archived (5)** → all became `012-F`:

| Stash | → | Tasks |
|---|---|---|
| ECE3DAB7 | → | `012.001-T`, `012.002-T`, `012.003-T`, `012.004-T` |
| CA469B3D | → | `012.005-T`, `012.006-T` |
| 5FE4A7BE | → | `012.007-T` |
| 5158769F | → | `012.008-T` |
| 805248F7 | → | `012.009-T` |

**Excluded, preserved unchanged (8)**: `A92E3FA0`, `4989A42D`, `EF9352FB`, `BF5DE670`
(non-implementable trackers / operator-only); `F47DB9A9`, `F4F4A959`, `2787DA56`,
`4C5BEC23` (Group B — CI/script hygiene, outranked, one-shipment limit).

**Newly captured from the review gate (2)**: `F133AB7E` (medium — GO-14 final-component
write-through reconsideration), `2362BBB5` (low — terminal-branch fail-open default +
`symlinkEscapeMsg` attribution).

## Grouping decision

Three groups formed by code similarity. **Group A (pathsafe) selected** because it holds
the only confirmed containment defect in scope, is the tightest single-package cluster,
is 100% bugs (top of the operator's type order), and every item has one defensible
technical answer — no operator input required. Group B was outranked: zero runtime
security impact, looser cohesion, and two of four items need operator *policy* decisions
that would have triggered a stop condition under AFK. Group C is not implementable.

## Gates

- **Deliberation**: `docs/decisions/2026-09-07-intercom-go-pathsafe-containment-correctness-deliberation.md`
- **Plan (rev 2.1)**: `docs/plans/2026-09-07-intercom-go-pathsafe-containment-correctness-plan.md`
- **P-006 hardening**: applied (§7 — threat model, mandatory negative tests, regression
  locks, blast radius, rollback)
- **Plan review**: 4 independent reviewers (Security Lens, Correctness, Scope Boundary,
  4-model Adversarial panel) → 1 P0 + 8 P1 raised, **all resolved** (plan §8)
- **Post-remediation re-review**: FAIL with 3 new P1 plan-precision defects → all fixed
  (plan §8.1) → **final gate confirmation PASS**

### Material changes forced by review

1. **P0** — added `012.002-T`: privilege-free Windows **junction** coverage. Junctions
   need no symlink privilege, so the original tests would have skipped on unprivileged
   Windows runners, leaving the widest attacker capability untested.
2. **ENOENT-only ascent** added to `012.003-T` — non-ENOENT probe errors (EACCES/ELOOP)
   were being treated as absence, letting the walk climb past a pre-planted link.
3. **Dropped the advisory macOS CI job** — it could not exercise the case-sensitivity
   hazard it claimed to cover (GitHub macOS runners are case-insensitive APFS), and it
   was the sole reason this shipment would open `.github/workflows/ci.yml`, the shared
   surface of excluded entries F4F4A959/F47DB9A9.
4. **Honest GO-14 framing** — the shipment *narrows* the escape class rather than closing
   it; "fail-closed" is no longer claimed unqualified. Reconsideration captured to stash.
5. **Rejected the stash's own suggested darwin fix** — extending case-folding to darwin
   would be fail-**open** on case-sensitive APFS. Recorded both directions instead,
   including the already-enabled Windows fold's own fail-open on case-sensitive NTFS.

## Known tooling degradation

`complexity` could **not** be stored as a structured field: this workspace's `task` WIT
type does not define one (`backlogit update --complexity` errors for
`artifact_type=task`). `size` IS structured (`custom_fields.size`, `size_source=agent`,
`size_ruleset_version=stage-2h-rule-v1`). Per the documented fallback, complexity is
recorded as labeled prose in each task's `implementation-notes` section
(`Size: X | Complexity: Y`) with the degradation noted inline. Both axes were assigned
independently; neither was derived from the other.

## Anti-goals carried into the shipment

No TOCTOU/hardlink mitigation (BF5DE670 stays deferred, trigger unfired); GO-14 not
closed; no modification to `.github/workflows/ci.yml`,
`scripts/check-unignore-regression.sh`, or `scripts/check-retired-architecture.sh`.
Plan §5 includes a mechanical `git diff --name-only` assertion proving these held.

## Next step

Hand `011-S` to the **Ship** agent. Execution order:
`012.001-T`, `012.002-T`, `012.005-T`, `012.008-T` (no upstream deps) →
`012.003-T`, `012.006-T`, `012.009-T` → `012.004-T` → `012.007-T`.
`012.001-T` + `012.002-T` + `012.003-T` **must land in one commit** (red-phase atomicity).

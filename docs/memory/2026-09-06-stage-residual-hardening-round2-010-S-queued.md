# Stage session — dark-factory residual hardening round 2 (post-009-S)

**Date:** 2026-09-06
**Agent:** Stage
**Mode:** DARK_FACTORY (operator AFK, full autonomous pipeline authorized)
**Outcome:** Queued shipment `010-S` ready for Ship.

## Scope

Bounded, operator-supplied: exactly 16 stash entries.
`4989A42D, A92E3FA0, EF9352FB, A0A2D049, 6E701953, 309FBF5A, 11ECB954,
AFE0E95C, DA945722, 707FE72B, 9F9A3CB2, BF5DE670, E89E5D00, 798002CB,
8D953C4B, 8472E0A1`

Hard constraint: **exactly ONE** queued shipment. Satisfied.

## Starting state (verified)

* No active or queued shipments; backlog queue empty; `001-F`…`010-F` all archived.
* One checkpoint, `ship`-owned and `resolved`, 0 quarantined → **zero-candidate
  normal startup** for Stage (not a recovery event).
* MCP surface not exposed → all backlogit operations via registry-declared
  **CLI fallback** (`TOOL_DEGRADED`, P-012-compliant). `INDEX_SYNC_OK`.

## Output

* Feature **`011-F`**, 20 tasks `011.001-T`…`011.020-T`, 7 sub-epics.
* Shipment **`010-S`** — 21 items, feature-first, dependency-ordered,
  read-back verified as an exact match to the harvest scope.
* 9 explicit dependency edges recorded (`--type blocks`).
* `backlogit doctor`: **No issues found.**

## Artifacts

* `docs/decisions/2026-09-06-intercom-go-residual-hardening-round2-deliberation.md`
  (+ Addendum: D-I revised, D-E re-scoped, D-J new)
* `docs/plans/2026-09-06-intercom-go-residual-hardening-round2-plan.md`
  (revision 5, reviewed)

## Plan-review gate

Five revisions. Six personas (Architecture, Correctness, Security,
Constitution, Scope Boundary, Maintainability).

| Attempt | Verdict | Blocking findings |
|---|---|---|
| 1 | FAIL | 4×P0, ~20×P1 |
| 2 | FAIL | 0 P0, 2×P1 |
| 3 | FAIL | 1×P1 (introduced by rev-2 remediation) |
| 4 | FAIL | 2×P1 (introduced by rev-3 remediation) |
| 5 | PASS | 0 P0/P1 |

Severity strictly monotonic decreasing; no cycling on the same defect.

### Highest-value catches

1. **depguard denied the wrong package** (4 personas). Everything imports the
   SDK **root** package; denying only `/rpc` matched **zero files** — a gate
   that would have shipped as false assurance for a C3 exit contract.
2. **Un-ignore rule banned every functional negation** — as written it was
   self-defeating; replaced with an old-vs-new behavioural differential using
   git's own matcher.
3. **`011.002-T` branch (a) was infeasible** — `pathsafe.normalize` rejects
   all absolute candidates and `NewRoot` requires an existing directory.
   Resolved by a new branch (c): a bounded, **expiring** Constitution III
   exception whose trigger is the mechanical write-path gate.
4. **Containment loosening disguised as a refactor** — GOOS-gating
   `hasWindowsDrivePrefix` would have made `cli_path = "C:evil"` *accepted*
   on Linux. Dropped; invariant J11 + anti-goal added.
5. **`011.011-T` sibling-prefix bypass** — the separator suffix is the sole
   defence against `<root>-evil`; the existing suite has no such case, so the
   natural fix would have converted a fail-safe over-rejection into a real
   containment bypass.
6. **My own rev-3/rev-4 defects**, caught and then **empirically verified by
   me** before accepting: the denylist referenced paths that are *not* ignored
   at `HEAD` and are ignored only by **droppable stowaway lines**. Verified by
   rebuilding `HEAD`'s `.gitignore` in a scratch repo and re-running
   `git check-ignore` per path. Standing rule added: *the denylist may never
   reference a pattern introduced by droppable content.*

## Key decisions

* **D-B** — `BF5DE670` rescoped *mitigate → consolidate + mechanically gate*.
  Measured: **zero** write primitives under non-test `internal/**` or `cmd/**`,
  so the entry's own trigger has not fired and TOCTOU machinery would be
  speculative.
* **D-C** — `A0A2D049` promoted to first: `testhelpers_test.go:77` skips only
  on `Start()` failure, so ambient credentials silently authorize live SDK
  calls *and real shell execution*. Active safety hole.
* **D-E** — measured `git show HEAD:.gitignore | grep -c '^!'` = **0**; the
  stowaway introduces the first-ever negations, so the checker must land first.
* **D-J** — the four preserve-only ignore lines are pure additions,
  **non-droppable**, and sequenced **first** rather than last.

## Dispositions

**Archived (11)** — annotated with forward references to their tasks:
`A0A2D049, 8D953C4B, 798002CB, 8472E0A1, E89E5D00, DA945722, 707FE72B,
11ECB954, AFE0E95C, 309FBF5A, 6E701953`

**Retained active (5)**, each annotated:

| Entry | Disposition |
|---|---|
| `9F9A3CB2` | Partially resolved — active until Ship discloses all 8 include-list paths |
| `BF5DE670` | Partially resolved — register delivered, mitigation deferred to the gate trigger |
| `A92E3FA0` | Standing roadmap tracker + sole index of 5 open operator decisions |
| `4989A42D` | Completion condition (b) **measured UNMET** (`check-retired-architecture.sh:75`) |
| `EF9352FB` | **BLOCKED — operator-only**, time-sensitive |

## Hard blocker for the operator

`EF9352FB` — a live Tavily API key sits in plaintext in untracked
`.env.local`. Repo-side containment is **closed** (never committed, never in
history, ignored). The residual action is **external credential rotation**,
which requires provider-console authority Stage does not hold and must not
simulate. **No secret material was read, revealed, copied, or committed at any
point this session.** Exposure persists until the operator rotates the key.

## Degradations flagged

* **MCP unavailable** → CLI fallback throughout (declared, not silent).
* **`complexity` is WIT-gated**: `features.sizing: true` but artifact type
  `task` defines no `complexity` field, so `--complexity` fails validation.
  `size` persisted structurally (`custom_fields.size`); `complexity` recorded
  as clearly labeled prose on all 20 tasks per the harness contract. This
  matches the prior learning `2026-09-04-backlogit-sizing-is-wit-gated.md`.

## Not done (deliberately)

* No C3 work — reliability/security precedes feature expansion, and C3 builds
  on all three packages this shipment corrects.
* U-E1a gate not broadened — unrequested; would be scope expansion.
* Working tree stowaways (`.gitignore`, `start.ps1`) untouched — they ride
  with Ship's feature PR under the operator-authorized stowaway protocol.

## Next step

Hand `010-S` to Ship. Execution order is encoded in the manifest and the
dependency edges; sub-epics A–D are independently valuable and land first.

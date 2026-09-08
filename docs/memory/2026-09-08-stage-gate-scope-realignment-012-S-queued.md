# Stage Session — U-E1a Gate Scope Realignment (012-S queued)

- **Date**: 2026-09-08
- **Agent**: Stage (P-017 DARK_MODE_ACTIVE, operator AFK)
- **Mode**: strictly bounded scope, exactly ONE shipment, Copilot CLI only (agent-intercom NOT installed)
- **Outcome**: `012-S` queued — ready for Ship

---

## Scope contract honoured

- **In scope (2)**: `A92E3FA0`, `4989A42D` — both **RETAINED**, dispositions updated
- **Excluded (12)**: `700B41CE`, `EF9352FB`, `BF5DE670`, `F133AB7E`, `AD0D9D1F`, `F47DB9A9`,
  `F4F4A959`, `2787DA56`, `4C5BEC23`, `2362BBB5`, `9D45E62E`, `1C6C3B46`
- **Verified untouched**: `git diff .backlogit/stash.jsonl` = 2 insertions / 2 deletions;
  a per-line comparison against `HEAD` confirmed all 12 excluded entries **byte-identical**.

## Decision

Broaden the U-E1a retired-architecture gate from `internal/config/**` to `internal/**`,
retiring the **vestigial D6a narrowing**. This is `4989A42D` completion condition (b) and
the single artifact where both in-scope trackers intersect — the literal alignment point
the run theme asked for.

**Decisive fact**: the narrowing's *sole* recorded cause (architecture-correction plan
L860 — `internal/apperr` declared `KindSlack`/`KindIPC`/`KindACP`, so an `internal/**` gate
"could never pass") was removed by `009.002-T`. Verified: zero matches in
`internal/apperr/*.go`. Empirically confirmed by running the real scan engine patched to
`internal/**` against the real tree → **exit 0, zero findings**.

**C3 deferred again**, but with a *stronger* reason than prior cycles' "moving base"
(now weak — apperr/pathsafe/config are settled): C3 lands new `internal/**` code, and until
the gate covers `internal/**` that code is never scanned. 012-S is a genuine **prerequisite**
for C3, not a deferral of it. No new stash entry for C3 (`A92E3FA0` already tracks it).

## Corrections of record produced this cycle

1. **"Born narrowed"** — both trackers say the gate must be broadened "**back** to"
   `internal/**`, "**reversing**" D6a. Git disproves this:
   `git log -S"internal/config/"` → one commit (`6dec85a`); `git log -S"internal/**"` → none.
   It is a **first-time broadening**. Condition (b)'s intent is unaffected.
2. **P0 — dead TOML engine** — `Path('config.toml.example').suffix` is `.example`, so
   `scan_path()` returns `[]` and `scan_toml()` is **never invoked on the real tree**.
   Included as same-surface correctness (`013.005-T`): a "pass" from an engine that never
   ran is not the signal condition (b) asks for.
3. **Principle IV deviation (self-reported)** — the §2.3 probe was written to an
   out-of-tree temp path. Disclosed in the deliberation, artifact deleted, future probes
   constrained to the workspace tree.

## Artifacts

- `docs/decisions/2026-09-08-intercom-go-gate-scope-realignment-deliberation.md`
- `docs/plans/2026-09-08-intercom-go-gate-scope-realignment-plan.md` (rev 2.1)

## Backlog created

| ID | Type | Size | Complexity | Unit |
|---|---|---|---|---|
| `013-F` | feature (chore-shaped) | — | — | covering item |
| `013.001-T` | task | S | medium | U1 RED structural scope assertions |
| `013.002-T` | task | S | medium | U2 Go fixture harness + engine dispatch |
| `013.003-T` | task | S | low | U3 Go fixture corpus (1 reject + 1 accept) |
| `013.004-T` | task | XS | low | U4 broaden scope to `internal/**` |
| `013.005-T` | task | XS | low | U5 P0 `config.toml.example` dispatch |
| `013.006-T` | task | S | low | U6 docs + D6c amendment |

**Shipment `012-S`** — 7 items, 0 unsized, status `queued`.

Dependency edges: `002←001`, `003←002`, `004←001`, `004←003`, `005←001`, `006←004`, `006←005`.

## Review

Seven-persona panel (Architecture, Correctness, Security Lens, Scope Boundary,
Constitution, Schema-CLI-Docs Coupling, Learnings). Rev 1 → **5 P0 + 6 P1**; all resolved
in rev 2. Re-review (cycle 2): all prior P0/P1 verified resolved,
`SCOPE VERDICT: COMPLIANT`, **zero open P0/P1**. Rev 2.1 applied the cycle-2 P2/P3
advisories. Gate verdict: **PASS**.

## Degradations recorded

- **`DEGRADED_MODE`** — no MCP surface this session; all backlog ops used the
  registry-declared **CLI fallback**. Not silent: declared in
  `.autoharness/backlog-registry.yaml`.
- **`chore` artifact type unsupported** — covering item recorded as `feature` with
  `CHORE-SHAPED` noted in its description.
- **`complexity` field undefined on the task template** despite the registry advertising
  `features.sizing: true`. `size` landed structurally in `custom_fields`; complexity
  recorded as labeled prose (`Size: X | Complexity: Y`) per the prescribed fallback.

## Deferred captures created (P-021, do NOT expand 012-S)

`24B63533` detection-vocabulary false negatives (splitter plurals + struct-tag masking);
`6285779C` `mask_go_non_code` characterization fixture suite;
`6C24E2E4` structural refactor (single `SCAN_SCOPE` source + extract Python engine);
`8D603E29` `cmd/` branch filter asymmetry.

**Deliberately NOT captured** (would duplicate an excluded entry's surface and read as
de-facto triage) — recorded in plan §8 only: the `mask_go_non_code` clone in
`check-write-path-precondition.sh` (`BF5DE670`), and `ci.yml`'s redundant
`continue-on-error: true` gate step (`F4F4A959`/`F47DB9A9`).

## Handoff to Ship

**`012-S`**. All six tasks MUST ship in **one PR** with a **merge commit** (Principle XI):
`--self-test` is deliberately RED from `013.001-T` until **both** `013.004-T` and
`013.005-T` land. A squash/rebase destroys the red→green evidence → halt with P-009.

Open advisories carried to Ship: `EF9352FB`, `AD0D9D1F`, `9D45E62E` have no declared file
surface, so U6's non-intersection with them is asserted rather than verified; and the
"+8 files" delta in §9.4 must be **re-derived at execution** (`internal/copilotprobe` is a
self-documented disposable spike).

## Still blocked on the operator (unchanged, all five)

Q3/H3 Dev Tunnel auth (C10); Q6/H4 sessions-per-process (C9); H5 permission authorization
model (C5); H6 session-pointer store semantics (C4); Q7 workspace tooling reach.
Dark mode cannot resolve any of these — each is a product-direction or security-model
answer Stage must not self-authorize.

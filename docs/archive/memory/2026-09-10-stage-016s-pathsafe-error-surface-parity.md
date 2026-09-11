---
title: "Stage Session Memory — 017-F / 016-S Pathsafe Error-Surface Parity"
date: 2026-09-10
agent: Stage
mode: DARK_MODE_ACTIVE
session_id: stage-dark-016s-pathsafe-error-surface-parity-20260910T185034
status: complete
---

# Stage Session Memory — 017-F / shipment 016-S

**Session**: dark-factory cycle, operator AFK, visibility mode CLI
(agent-intercom not installed).
**Outcome**: **shipment `016-S` queued** and ready for Ship.

## Scope

| | |
|---|---|
| Bounded stash scope | `BF5DE670`, `6B751D8B`, `7ADAC481` |
| Excluded (untouched) | `A92E3FA0`, `37FAB8C2`, `4A01C53E`, `1C6C3B46`, `2787DA56`, `4C5BEC23`, `6C24E2E4`, `9D45E62E`, `EF9352FB`, `F47DB9A9` |
| Shipment limit | exactly one — satisfied (`016-S`, the only shipment in the workspace) |

## Pipeline executed

Step 0.0 tool gate (`TOOL_OK: backlogit 1.10.1`) → 0.1 index sync
(`INDEX_SYNC_OK`, 194 artifacts) → checkpoint recovery (zero active
candidates; both existing checkpoints `resolved` — normal startup, not a
failure) → Step 1 triage → Step 1.8 learnings (high confidence) → Step 2
deliberation → Step 3 plan + P-006 hardening → Step 4 multi-persona
adversarial review (**PASS**, attempt 1) → Step 5 harvest → Step 5.5 shipment
→ Step 5.6 stash disposition.

## Artifacts

- `docs/decisions/2026-09-10-intercom-go-pathsafe-error-surface-parity-deliberation.md`
- `docs/plans/2026-09-10-intercom-go-pathsafe-error-surface-parity-plan.md` (Rev 3)
- `docs/decisions/2026-09-10-intercom-go-pathsafe-error-surface-parity-plan-review.md` (PASS)

## Backlog created

```
016-S (queued, high)
└── 017-F  Pathsafe error-surface parity and Windows long-path precision
    ├── 017.001-T  Normalize checkSymlinkEscape reparse error via wrapPathError   [S/medium, high]
    ├── 017.002-T  Lock and correct device-namespace branch reachability          [S/low,    medium]
    └── 017.003-T  Make MAX_PATH threshold UTF-16 code-unit aware                 [S/medium, medium]
                   └── depends on 017.002-T (blocks edge recorded)
```

## Key decisions (full rationale in the deliberation)

- **D-2** `7ADAC481` re-prioritized **low → high**. The 015-S closure record
  shows the Architecture Strategist rated it **P1 in isolation**; it was
  deferred on P-021 C1 *scope* grounds, not severity. It leads the shipment.
- **D-6** `BF5DE670` contributes **no task**. Its trigger was re-measured at
  `1ff2b42` and **has not fired**: write-path gate exits 0, `--self-test` 5/5,
  zero write primitives across 17 non-test `.go` files. Preserved as feature
  ACs **AC-F1…AC-F5**. **Not archived.**
- **D-7/D-8** `6B751D8B` items (3) buffer pooling and (4) CI lint scoping
  deferred (pure perf/CI, outside operator's named scope). **Not archived.**
- **D-10** *(from review)* Accepted that the UTF-16 precision gain re-enables
  ordinary Win32 normalization in a narrow band. The prior "verdict-neutral by
  construction" claim was **withdrawn**.
- **D-11** *(from review)* Safety modes **`freeze-scope`** and
  **`investigate-first`** declared by name.

## Material corrections made this session

1. **`6B751D8B`'s stated premise was FALSE.** It claimed the `\\.\`
   device-namespace branch is "unreachable given `filepath.IsAbs`". Measured:
   `filepath.IsAbs` returns **true** for `\\.\` paths including a 307-char one.
   The branch is **reachable but behaviourally redundant**. A
   panic/unreachability assertion was rejected by name.
2. **UTF-16 change is precision, not security.** UTF-8 bytes ≥ UTF-16 units
   always, so the old proxy could only over-trigger. Vulnerability framing is
   an explicit anti-goal.
3. **Windows CI test job is advisory** (`ci.yml` L251,
   `WINDOWS_GATE_REQUIRED != 'true'`). Ship must paste real Windows test output
   as PR evidence.
4. **`requireSymlinkOrFailClosed` skips** on `ERROR_PRIVILEGE_NOT_HELD` — a
   symlink-only Windows test could hide a red path behind a green suite. Hence
   AC3b's junction fixture.

## Review

5 personas, 5 different models, parallel. **2 P1 + 6 P2 + 7 P3, zero P0 — all
remediated in-cycle.** Two findings were independently corroborated by two
reviewers each (POSIX `Op`/`Path` defect; Win32-normalization overclaim).
Plan-review attempt counter: **1** (no re-entry cycles).

## P-021 compliance

- **Duplicate detection (A)** — unconditional, all 13 active entries pairwise.
  **CLEAN** for all three scoped entries.
- **Reconciliation (B)** — `7ADAC481` PR **`N/A` → #48** (merge `fbd8d8d0`),
  closure #49, remediation #50, recovered from the Ship-owned residual-risk
  record joined on 015-S. `review-thread` remains `N/A` (truthful terminal
  record, threadless path). `BF5DE670` — idempotent no-op.
- **C6** — both marker-carrying entries deliberated before any planning.

## Next step for Ship

Claim **`016-S`**. Lead with `017.001-T`. Merge approval for the resulting PR
is pre-authorized; **admin fallback is NOT authorized**. Re-request Copilot
review via the `[bot]`-suffixed REST login after every HEAD advance (P-018).

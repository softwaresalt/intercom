---
title: "Stage session memory — residual hardening triage (dark-factory, post-005-S)"
date: 2026-09-04
agent: Stage
mode: DARK_FACTORY
status: complete
---

# Stage session memory — residual hardening triage cycle

## Scope

Exactly 19 stash entries carried forward after `005-S` closed. Three entries
created during 005-S (`A0A2D049`, `6E701953`, `309FBF5A`) were **excluded by
operator directive** and were not triaged, planned, harvested, or archived.
Verified untouched at session end.

## Tooling posture

* MCP surface unavailable in this invocation → **DEGRADED_MODE**, all backlog
  operations performed via the declared `backlogit` CLI fallback (P-012). No ad
  hoc filesystem substitution was used.
* `INDEX_SYNC_OK` at session start; checkpoint enumeration returned
  **zero total, zero quarantined** → ZERO-CANDIDATE NORMAL STARTUP, no recovery.

## Dispositions (19/19, exactly one each)

| Disposition | Count | IDs |
|---|---|---|
| Harvested | 13 | `EFFAA358` `90350C9A` `F7C6420D` `BEDD2E70` `B6203CCA` `35D76D5E` `78D13775` `8C2D578D` `9A4C8749` `7774C9CA` `4A91CA81` `3D61B5A8` `7028FBE7` |
| Resolved by Stage in-cycle | 2 | `2130906D` `90EE7758` |
| Archived obsolete | 1 | `8ACF7110` |
| Retained standing tracker | 2 | `4989A42D` `A92E3FA0` |
| Blocked / operator action | 1 | `EF9352FB` |

16 entries archived; 3 remain active (2 trackers + 1 blocked).

## Shipments produced

`006-S → 007-S → 008-S → 009-S`, chained with `blocks` edges, all `queued`.

| Shipment | Feature | Tasks | Theme |
|---|---|---|---|
| `006-S` | `007-F` | 6 | CI supply-chain hardening |
| `007-S` | `008-F` | 5 | Build/scan script hardening |
| `008-S` | `009-F` | 6 | apperr taxonomy correction |
| `009-S` | `010-F` | 5 | pathsafe/config containment |

**Next eligible shipment: `006-S`.** All 22 tasks carry both `size` and
`complexity` (`unsized: 0` on every manifest).

## Key decisions

* **`8ACF7110` obsolete** — oracle demoted by D8; protected P2–P14 ordering is
  void; `docs/oracle-pin.md` has no live consumer (independently re-verified).
  File left in place as historical record.
* **`EF9352FB` blocked** — repo-side containment verified closed (never
  committed, no `.env*` ever added, ignored at `.gitignore:36`); only external
  rotation remains, which is operator-only. Isolated into no shipment.
* **C3 deliberately not harvested** — operator rule 6: the primitives C3 builds
  on are corrected first, so C3's blast radius does not grow.
* **Order `008-S → 009-S` is a hard symbol dependency** via `apperr.Wrapf`
  (009-F U5 → 010-F U1), not a stylistic preference.

## Adversarial review

Five reviewers across four model families; **all returned FAIL** on the first
draft. Six convergent blockers remediated in-cycle (`Wrap` cannot carry a
message; U6 containment regression; golangci v2 schema; CODEOWNERS
deadlock/overclaim; cross-shipment test-first ordering; ground-truth errors —
rule 7 not 10, Python not bash, Ordinal comparison). Each plan carries a
Post-Review Remediation Record. Post-remediation verdict: **PASS with recorded
residual risk**.

Two findings partially accepted with reasoning recorded rather than adopted:
archive-`4989A42D` (retention kept, terminating condition added) and reorder
`010-S` before `009-S` (rejected — `Wrapf` dependency makes original order
technically required).

## Operator work preserved

Untouched, never staged: `.gitignore`, `start.ps1`, `.claude/`,
`.github/copilot/`, `.backlogit/hooks_queue.jsonl`.

`.backlogit/stash.jsonl` is **mixed**: it already carried three uncommitted
operator/Ship additions (the excluded follow-ups) before this session. Those
were preserved verbatim and are included in the staging commit because they
share a file with Stage's archival work; no operator content was altered or
discarded.

## Open items for the operator

1. **`EF9352FB`** — rotate the Tavily key (external console; time-sensitive,
   plaintext on disk).
2. **CODEOWNERS enforcement** — enabling code-owner review requires first
   provisioning an unattended-PR approval path, else dark-factory merges
   deadlock.
3. **Semantic unification of the two traversal predicates** — deferred; needs
   its own deliberation and an operator-facing migration decision.
4. Five C3–C11 decisions still tracked by `A92E3FA0` (Q3/H3, Q6/H4, H5, H6, Q7).

# Memory — Stage re-plan cycle after adversarial MUST_REPLAN at `7bc0318`

> **HISTORICAL SESSION SNAPSHOT — SUPERSEDED (2026-09-13).**
> Governing authority: `docs/decisions/2026-09-13-intercom-go-a-only-simplification-decision.md`
> (verdict `SIMPLIFICATION_VALID`).
>
> **Do NOT route from this file.** The current route is **`021-S` (A) → `017-S`**, two
> shipments only. `017-S` is a **13-member fully-covered root** (root `018-F` + all 12
> descendants) depending on **`021-S` only**, and closes by the **existing** P-015
> `CASCADE` exception. The three-shipment `A → B → C` chain is retired: `023-F`,
> `023.001-T`, `024-F`, `024.001-T` are `blocked` and shipment records `022-S` / `023-S`
> are archived. `TASK_ONLY_FINALIZE` is deferred platform work with **no current
> consumer**, and the `SAFE_CLOSE` exit-9 conflict is **off `017-S`'s route**.
>
> Everything below is retained as a record of what was true during that session.

- **Date**: 2026-09-12
- **Agent**: Stage
- **Branch**: `chore/stage-pipeline-policy-gap`
- **Entry HEAD**: `7bc0318` → **Exit HEAD**: `ce42fd2`
- **Scope**: docs / backlog / evidence only. No implementation, push, PR, claim, or merge.

## What this session was

One bounded re-plan/remediation cycle over the three-shipment bootstrap
(A `021-S`, B `022-S`, C `023-S`) that installs `TASK_ONLY_FINALIZE` so the 12-member
task-only shipment `017-S` becomes closable.

## The decisive measurement

The primary blocker was whether C's covering feature `024-F` becomes `active`. It does.

| Probe | Measured | Result |
|---|---|---|
| **22** | feature status across claim and task-only close, both manifest shapes | `queued` → **`active`** → **`active`**; also `active` at claim on the fully-covered-root shape |
| **23** | `dep add <successor-shipment> <feature> --type blocks` | **REFUSED** — *"both endpoints must be shipments"*. Successor already eligible (`ready_set=[002-S]`) while the feature is `active` |
| **24** | cascade over `[feature, task]`, in **both** manifest orders + control | `archived_ids` includes the feature; 0 other active top-level units; `ORDER_INDEPENDENCE_MEASURED=True`; control (task-only) leaves it `active` |

**Conclusion.** `024-F` becomes `active` and stays `active`, so P-001's `Active` condition
is tripped. **No mechanically enforced barrier exists** — backlogit accepts no
feature-endpoint dependency edge and defines no shipment `blocked` status. Rather than
return `MUST_REPLAN` into a measured deadlock, the hazard was removed at its source: the
feature was moved **inside** C's manifest so the engine's own cascade archives it.

## Decisions

1. **`023-S` re-shaped** to `[024.001-T, 024-F]`. C now closes by the pre-existing
   `CASCADE`. Shipment IDs, dependency chain and the 6+6+16 inventory split all preserved.
2. **C is a two-session close.** Its merge activates the token, so S1 checkpoints and ends;
   a fresh S2 runs the full owner-selected, operator-confirmed recovery, re-verifies
   `main`/merge SHA/merged tokens, then classifies. The rule lives in C's `B18` as
   *procedure*, leaving A's Role-Boundary-scoped carve-out and its Go test row 11 untouched.
3. **Cost accepted and recorded**: C no longer self-proves the new path. Relocated to Probe
   21's post-half against the exact `017-S` 12-member fixture; residual **R-8** states that
   no *live* task-only closure precedes `017-S`.
4. **Overflow valve removed** from C entirely — an overrun is HALT to Stage, never a
   partial `Done`.
5. **`017-S`'s own residue** (`018-F` left `active`) is recorded, **not** gated: `017-S` is
   terminal, so nothing downstream can be blocked (**R-10**).

## Internal review

Ran a Correctness review; it returned **FAIL** with 9 MAJOR + 9 MINOR findings, all on this
session's own surface. Every one was fixed, including two that required re-running probes:

* Probe 24 had never tested the **live** manifest order → added **ARM QR**, which
  reproduces the live construction (`shipment add`) exactly and measures the same result.
* A PowerShell helper that both wrote and returned had swallowed intermediate snapshots and
  malformed two verdict fields → helpers now write-only or use `Write-Host`.
* Plan A's PO-1b read a queue path for a `done` feature, making its own idempotent-resume
  branch unreachable → now reads status via `backlogit get`.
* Probe 21 was mandatory but unbudgeted, and sequenced before a closure `sync` with no
  scratch cleanup → budgeted (~30 min, Ship closure) and given mandatory scratch discipline.

## Validation at exit

`backlogit doctor`: no issues · index 238 artifacts, **zero** duplicate-source-id warnings ·
`ready_set: [021-S]` only · chain `021-S → 022-S → 023-S → 017-S`, no cycles · all statuses
`queued` · one task per bootstrap feature · markdownlint clean · one worktree · clean tree.

Removed stale probe scratch (`p18`, `p19-*`, `p20`, `seedtest`) that had been polluting the
live index and surfacing a phantom `002-S` in `shipment list`.

## Commits (local only — no push, no amend)

| SHA | Subject |
|---|---|
| `d8ef931` | measure covering-feature status at claim and the barrier options |
| `f1ca7cd` | reshape C to a covered root and split its close into two sessions |
| `340abfc` | align A and B with the installed lifecycle and measured claim |
| `ce42fd2` | correct the governing artifacts to the rev-2 topology |

## Next step

Independent adversarial re-review of the rev-2 surface. Nothing is routed to Ship;
`021-S` remains the only eligible shipment and is unclaimed.

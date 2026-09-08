---
title: "013-S / 014-F pathsafe reparse-point containment hardening — compacted release summary"
date: 2026-09-08
shipment: 013-S
feature: 014-F
status: shipped
compacted_from:
  - docs/memory/2026-09-08-stage-pathsafe-reparse-containment-013-s-memory.md
  - docs/memory/2026-09-08-1601-013-s-pre-pr-checkpoint.md
---

# 013-S / 014-F — compacted release summary (Stage + Ship, full lifecycle)

## Outcome

**Shipped.** PR #40 merged via merge commit `17daffb5614f2f8ed797d7647aeadcc1dc29b1d2`.
Shipment `013-S` and feature `014-F` closed via the P-015 verified
fully-covered-root cascade path (`backlogit shipment ship`), fully verified
(empty `returned_ids`, exact two-set `allowed_ids`/`required_ids` match,
`parent_id` preserved on all 8 tasks). Post-merge closure complete except
for the compaction step this file itself represents.

## Origin (Stage)

Triaged from 22 active stash entries; Cluster A (`internal/pathsafe`
containment security: `700B41CE`, `F133AB7E`, `2362BBB5`, `AD0D9D1F`)
selected over a retired-architecture gate cluster and a docs-hygiene
cluster because it held the only `critical` entry — an empirically
reproduced, unprivileged, live Windows directory-junction containment
bypass in a NON-NEGOTIABLE security control. Plan went through 2
deliberation cycles (rev 2 PASS): a 3-model adversarial review of rev 1
found 1 P0-equivalent (final-component-only fix left C2/C3 residual
bypasses) and 4 P1s, all resolved before execution — most notably, an
attempt to relax the F3 ancestor-rejection branch was explicitly *removed*
from scope after review showed it would introduce a *new* bypass; the
shipment stayed monotonically fail-closed.

Produced: feature `014-F`, tasks `014.001-T`…`014.008-T`, shipment `013-S`
(9 manifest items, dependency-ordered). 5 stash entries archived with
forward-reference dispositions during triage; 17 remained active
(deferred clusters + trackers), annotated for the next Stage cycle.

## Execution (Ship)

Strict test-first (RED→GREEN) across all 8 tasks:

1. `014.001-T` RED containment matrix (C1/C2/C3) — only C1 was a live bug.
2. `014.002-T` RED inverted GO-14 dangling-symlink test.
3. `014.003-T` `canonicalizeReparse` helper (direct `GetFinalPathNameByHandleW`
   via `syscall.NewLazyDLL`, no new module dependency).
4. `014.004-T` GREEN: wired into `checkSymlinkEscape`.
5. `014.005-T` flipped one fail-open terminal branch to fail-closed.
6. `014.006-T` split escape message into `symlinkEscapeMsg` (verified escape)
   vs `symlinkUnverifiableMsg` (unverifiable).
7. `014.007-T` closed 2 POSIX branch-coverage gaps (F4, F11).
8. `014.008-T` reconciled risk register + doc comments.

**Post-implementation review** (before PR, per operator mandate):
standard report-only review (3 fixes: causal misattribution, `uncPrefix`
misattribution, an errcheck violation invisible to CI's ubuntu-only lint
job) + 3-reviewer adversarial review (5 fixes: pwsh silent-skip→fail-loud,
3 new `Resolve()`-level regression-lock tests; 3 findings deferred via
P-021). Both READY_WITH_FOLLOWUPS. Full detail:
`docs/closure/2026-09-08-013-s-014-f-pathsafe-reparse-containment-adversarial-review.md`.

**Notable hard-won corrections** (see
`docs/compound/2026-09-08-verify-go-stdlib-internals-claims-against-goroot-source.md`):
two independent false claims about Go/Windows internals — one self-authored
(EvalSymlinks does NOT call GetFinalPathNameByHandleW as a general
mechanism), one from an automated Copilot review (a false "won't compile"
claim about `make([]uint16, n+1)` with a `uintptr` argument) — were both
caught and corrected only by tracing actual `GOROOT` source / running an
isolated repro, not by trusting plausible-sounding claims.

## PR / merge / gates

- PR #40, HEAD `3b929ad1325dc01018e58f4d039af67b9587b96a` (final), merged
  with merge-commit strategy (P-009).
- All CI green (cross-compile ×4, lint, security, test, test-windows-advisory,
  pipeline-topology, ci gate) at both the initial and post-docs-commit HEAD.
- Copilot review (P-018): `SATISFIED`, 1 round, 1 thread (the false-positive
  above), replied + resolved.
- Dark-mode merge authorization: all 6 §1.9.6 conditions met;
  `merge_approval_pre_authorized: true` satisfied the P-014 approval signal.

## Runtime verification

**`BLOCKED`** — `pathsafe.Root.Resolve()` (the exact function hardened) has
no wired live entrypoint anywhere in `cmd/**` yet (pre-existing
project-maturity gap, not introduced by this shipment). Compensating
function-level evidence: 45/45 pathsafe tests, full `go test ./...`,
cross-platform build. Follow-up stashed: `37FAB8C2`.

## Follow-up stash entries (all P-021, none P0/P1)

- `4104AF54` — CI lint job is ubuntu-only, can't lint Windows-tagged files.
- `E428AB46` — NewRoot vs. canonicalizeReparse mechanism asymmetry if a
  workspace root is itself ever a junction/GUID volume.
- `E4C5413F` — bundled: long-path CreateFile prefixing asymmetry, caller-side
  stripUNCPrefix reliance, theoretical LazyProc.Call panic path.
- `37FAB8C2` — add black-box runtime verification for Root.Resolve() once a
  live caller exists.

## Open items for the operator (from Stage, still relevant)

1. `EF9352FB` — rotate the Tavily API key (external console; time-sensitive,
   isolated into no shipment; operator-only).
2. `WINDOWS_GATE_REQUIRED` — still unset by design; not flipped by this
   shipment (repository-configuration decision reserved for the operator).

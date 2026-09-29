---
title: "Compacted memory - early Stage/Ship closure addenda (008-S through 013-S)"
date: 2026-09-28
status: compacted
release_units:
  - 008-S
  - 009-S
  - 010-S
  - 011-S
  - 013-S
source_count: 12
compacted_from:
  - docs/archive/memory/2026-09-04-stage-first-round-c2-reconciliation.md
  - docs/archive/memory/2026-09-04-stage-residual-hardening-triage-session.md
  - docs/archive/memory/2026-09-05-195408-ship-008-s-session-complete.md
  - docs/archive/memory/2026-09-06-004136-ship-pr26-closure-repair-merged.md
  - docs/archive/memory/2026-09-06-004800-stage-closure-contract-bug-capture.md
  - docs/archive/memory/2026-09-06-214700-ship-009-s-post-merge-closure-complete.md
  - docs/archive/memory/2026-09-06-ship-010-s-session-complete.md
  - docs/archive/memory/2026-09-06-stage-residual-hardening-round2-010-S-queued.md
  - docs/archive/memory/2026-09-07-stage-pathsafe-containment-011-S-queued.md
  - docs/archive/memory/2026-09-08-013-s-session-end-summary.md
  - docs/archive/memory/2026-09-08-ship-011-s-session-complete.md
  - docs/archive/memory/2026-09-11-stage-pathsafe-deferred-trigger-reassessment-no-shipment.md
---

# Compacted: early Stage/Ship closure addenda (008-S through 013-S)

## Scope

This summary preserves the useful details from early September memory checkpoints whose release-unit
summaries already exist elsewhere in `docs/memory/compacted/`. It keeps the deltas that were not
captured in the earlier compacted records: Stage triage decisions, closure repair incidents, and
handoff constraints that still explain later pipeline behavior.

## Outcomes by release unit

* `008-S` / `009-F` shipped the `apperr` taxonomy correction through PR #24, then required a
  closure-evidence repair PR #26 because the post-merge close recognizer did not initially identify
  the completed 008-S closure evidence. The repair confirmed the root cause was closure-recognition
  metadata, not implementation drift. Stage separately captured the closure-contract bug for later
  pipeline hardening rather than changing runtime code.
* `009-S` shipped the pathsafe/config containment hardening follow-on and completed post-merge
  closure. The useful residual was process, not code: closure PR approval stays separate from the
  feature PR, and post-merge closure records must be explicit about which checks were runtime
  applicable versus not applicable.
* `010-S` / `011-F` completed Ship execution with no new runtime blockers. Stage's preceding
  residual-hardening round assembled and reviewed the next work, recorded no P0/P1 residuals, and
  kept follow-up stash items outside the active release unit.
* `011-S` / `012-F` completed pathsafe containment correctness after the earlier Stage direct-to-main
  incident had been unwound by a revert and proper staging reapply. The key invariant preserved here
  is that operator-visible staging and closure PRs are mandatory even for corrective evidence-only
  work.
* `013-S` / `014-F` completed the Windows reparse-point containment shipment end to end under dark
  factory mode. PR #40 merged via merge commit `17daffb5614f2f8ed797d7647aeadcc1dc29b1d2`; closure
  finished with runtime verification blocked only because `Root.Resolve()` still had no live CLI or
  server entrypoint.

## Decisions and rationale

* Stage grouped security-critical pathsafe containment items ahead of docs hygiene and retired-arch
  gate items because the cluster included the only critical stash entry.
* Residual hardening work remained backlog/stash driven. Items outside the authorized release surface
  were captured or left active rather than folded into the shipment.
* Closure repair work confirmed that post-merge closure correctness is a separate contract from code
  correctness. A release can be technically shipped while its closure evidence still needs a
  dedicated approval/repair PR.
* Runtime verification was kept honest: pure internal-package changes were marked not applicable or
  blocked when no live runtime entrypoint existed, instead of fabricating browser or CLI evidence.

## Files and artifacts modified in the original sessions

* `internal/apperr`, `internal/pathsafe`, and `internal/config` production and test files across
  the shipped release units
* `docs/closure/*` closure artifacts for 008-S, 009-S, 011-S, and 013-S
* Backlog records for the shipped features and their post-merge closure status transitions
* Stage planning artifacts and stash dispositions for the residual-hardening and pathsafe clusters

## Failed approaches and gotchas

* A closure recognizer that relies on incomplete metadata can miss an otherwise completed shipment;
  evidence repair must fix the closure contract, not re-open shipped implementation work.
* Windows containment review findings must be empirically reproduced against `main` before deciding
  whether they are in scope. This discipline avoided mixing pre-existing Windows junction risk into
  a narrower dangling-junction shipment.
* `backlogit` field support is WIT-gated: visible CLI flags do not prove a field is valid for the
  artifact type.

## Still-relevant follow-ups

* Operator-only external-secret work and Windows CI policy choices remained outside these release
  units.
* Black-box runtime verification for `Root.Resolve()` remained blocked until a real live caller is
  introduced.

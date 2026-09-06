---
title: "Compacted memory: 007-S -- Build and scan script hardening"
date: 2026-09-06
shipment: 007-S
feature: 008-F
pr: 22
status: shipped
compacted_from: docs/archive/memory/2026-09-06-ship-007-s-session.md
---

# 007-S / 008-F -- compacted session summary

**Outcome:** Shipped. PR #22 merged (merge commit
`c521b2cf07555493ad35a1d20844acbba3f12045`). Shipment closed via P-015
fully-covered-root cascade (`backlogit shipment ship`).

**Scope:** 5 units, test-first, in dependency order U2→U1, U4→U3, U3+U4→U5.
No Go source (non-test) changes -- `scripts/build.ps1`'s I5 guard extracted
into `scripts/lib/OutputPathGuard.ps1` and made symlink/junction-aware;
`scripts/check-retired-architecture.sh`'s TOML masking rewritten to a real
`tomllib` parse with a tested fallback lexer; `--self-test` wired into CI.

**Key decisions:**
* P-021 same-contract-surface completions fixed in-scope (fallback lexer
  fail-open EOF bug; dangling-reparse-point guard bypass; its own
  error-suppression over-broadening; PEP 585 annotation crash on Python <3.9;
  missing test timeout) vs. two genuinely pre-existing/unrelated findings
  deferred (`DA945722` Windows CI runner gap; `707FE72B` drive-root
  double-separator edge case) -- both captured per the C1/C2 defer-capture
  procedure with thread reply + GraphQL resolution where a thread existed.
* Cascade close (not safe-close) selected for 007-S after live-verifying the
  P-015 fully-covered-root preconditions; also the only path this backlogit
  version (1.10.1) permits (generic `move --status shipped` hard-errors).
* Branch created from a dirty `main` after the operator explicitly
  pre-disclosed and named the exact preserved files; never staged/committed
  any of them (one accidental inclusion of an untracked hooks-queue file was
  caught and corrected via `git reset --soft HEAD~1` before push).

**Failed/corrected approaches:** an initial `-ErrorAction SilentlyContinue`
fix for the dangling-reparse-point bypass was itself found (by post-fix
re-review) to over-suppress unrelated Get-Item errors; narrowed to catch
only `ItemNotFoundException`.

**New compounding learning:** requesting Copilot review on a repo where it
is not a listed collaborator requires the REST endpoint with the
`[bot]`-suffixed login (`copilot-pull-request-reviewer[bot]`); the
unsuffixed login and `gh pr edit --add-reviewer` both fail with a
"not a collaborator" 422. See
`docs/compound/2026-09-06-requesting-copilot-review-non-collaborator-repo.md`.

**Full detail:** `docs/archive/memory/2026-09-06-ship-007-s-session.md`.

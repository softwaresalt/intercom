---
title: "Adversarial-review findings must be empirically verified and scope-checked against main before disposition"
description: "A single-reviewer adversarial finding on Windows junction containment was escalated to blocking by independent stdlib trace; Ship additionally reproduced it live and confirmed pre-existence via git show main before deciding P-021 scope, avoiding both silent-ship and scope-creep failure modes"
date: 2026-09-08
tags:
  - "compound"
  - "adversarial-review"
  - "p-021"
  - "pathsafe"
  - "windows"
  - "security-review"
shipment: 011-S
feature: 012-F
---

# Adversarial-Review Findings: Verify Empirically, Scope Against `main`, Then Decide

## Context

Shipment 011-S's chartered scope was closing a **dangling** intermediate
symlink/junction containment bypass (`ECE3DAB7`) in
`internal/pathsafe.checkSymlinkEscape`. A 4-model adversarial consensus
review (report-only) escalated a single reviewer's finding — a **live**
(non-dangling) Windows junction at the final path component bypasses
containment entirely — from LOW confidence (1-of-4 raw vote) to HIGH
confidence / blocking, based on the orchestrating review agent's own trace
of the pinned Go stdlib source (`os/stat_windows.go`,
`path/filepath/symlink_windows.go`).

## The two failure modes this pattern avoids

1. **Silent-ship**: trusting the review's own "verified: yes" language at
   face value and either (a) ignoring a genuinely critical, exploitable
   finding, or (b) unilaterally expanding the shipment's scope to fix a
   NON-NEGOTIABLE security control's design without the deliberation/
   plan-hardening cycle that kind of change requires (P-021 violation in
   the other direction — same-surface completions are mandatory, but a
   genuinely new vulnerability class discovered mid-execution is not a
   same-surface completion).
2. **Rubber-stamp trust**: accepting a review agent's static-trace
   conclusion as ground truth without independent confirmation, when the
   agent doing the tracing had already disclosed a `sub-agent
   recursion-depth limit` prevented it from executing its own proposed
   empirical verification.

## What Ship did before deciding disposition

1. **Empirically reproduced the claim directly**, not just re-read the
   trace: wrote a throwaway (never committed) Go test creating a *live*
   junction pointing outside a test root, called `Resolve()` on the
   junction path itself as the literal final component (not a subpath
   through it — the first attempt used a subpath and produced a
   misleading "safe" result, because that exercises a *different*,
   already-fixed ancestor-rejection branch), and confirmed by actually
   reading a planted secret file through the accepted path.
2. **Checked pre-existence against `main`** via `git show
   main:internal/pathsafe/pathsafe.go` and manual trace, before deciding
   P-021 scope: `main`'s original `checkSymlinkEscape` used `os.Stat`
   unconditionally for every ancestor including the final component, so
   the identical Stat-follows-junction / EvalSymlinks-does-not-resolve-
   junction / hasPathPrefix-passes-lexically mechanism already existed.
   This confirmed the finding was **pre-existing, not introduced or
   worsened** by the shipment's own dangling-symlink fix — the decisive
   fact for P-021 C1 scope classification (out of chartered scope, defer-
   capture required, not a same-surface completion).
3. **Cleaned up the review agent's own scratch artifacts** (an
   uncommitted, gated-off verification test file it could not delete due
   to the same tool-recursion limit) before proceeding — a review agent's
   incomplete self-cleanup is not itself a merge blocker, but it must not
   be silently left in the tree either.
4. **Disposed via the existing P-021 defer-capture mechanism**: documented
   in the security control's own risk register (with an explicit
   cross-reference to the full adversarial-review artifact), captured as
   a `critical`-priority stash entry with `requires_deliberation: true`,
   and disclosed prominently in the PR body's own "known residual risk"
   section — not silently merged, not unilaterally fixed.

## Reusable pattern for future sessions

* A review agent's own "empirically verified" or "traced against source"
  language is not sufficient on its own when the agent has *also*
  disclosed a tooling limitation that prevented it from actually running
  the verification it proposes. Independently reproduce load-bearing
  claims when the tooling to do so is available (it usually is, for the
  orchestrating agent, even when a sub-agent lacked it).
* When reproducing a containment-escape claim, exercise the **exact**
  reported shape (here: the reparse point *as the literal final path
  component*, not as an intermediate segment of a longer candidate) —
  a plausible-looking but structurally different reproduction attempt can
  produce a false-negative that looks like a successful rebuttal.
* Before treating a newly-discovered security finding as in-scope or
  out-of-scope for the shipment that surfaced it, check whether the exact
  mechanism already existed in `main` (`git show main:<path>` plus a
  short manual trace is usually enough). Pre-existing-and-unworsened vs.
  newly-introduced-or-newly-exacerbated is the load-bearing fact for
  P-021 C1 scope classification and for whether "ship now, defer the fix"
  is a defensible disposition.

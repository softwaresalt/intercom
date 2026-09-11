---
title: "P-018 Copilot-review engagement must be actively attempted before merge, not just gate-checked"
date: 2026-09-10
category: build-errors
tags: [github, copilot, pr-review, p-018, dark-mode, process]
---

# P-018 Copilot-review engagement must be actively attempted before merge, not just gate-checked

## Symptom

`autoharness gate copilot-review <pr> --enforcement auto` returns
`NOT_APPLICABLE: PASS` ("no engagement signal") before merge, so the Ship
agent proceeds to merge. Only afterward does it become clear that Copilot
review COULD have been actively engaged on this repository via the
documented `[bot]`-suffixed REST fallback
(`docs/compound/2026-09-06-requesting-copilot-review-non-collaborator-repo.md`),
which was not attempted before merge — only a GraphQL `suggestedActors`
probe (empty) and `gh pr edit --add-reviewer copilot-pull-request-reviewer`
(unsuffixed, fails "not found") were tried. Once the PR is merged/closed, a
subsequent `POST .../requested_reviewers` with the correct `[bot]`-suffixed
login returns HTTP 200 with the PR object, but registers no actual review
request (`reviewRequests` stays empty) — the window to engage Copilot on
that PR is permanently closed once it is merged.

## Root cause

`autoharness gate copilot-review`'s `NOT_APPLICABLE` verdict answers "is
Copilot currently engaged," not "was engagement actively attempted using
every available method." A GraphQL `suggestedActors` probe returning no
Copilot entry, and an unsuffixed `gh pr edit --add-reviewer` failing, both
look like "Copilot review is unavailable on this repository" — but this
repository has a previously-verified, working, `[bot]`-suffixed REST
fallback for requesting Copilot review even when it doesn't appear as a
normal suggested/collaborator reviewer (see the referenced compound entry,
and PR #46 in shipment 014-S which received 2 genuine rounds of Copilot
review this way). Treating the gate's `NOT_APPLICABLE` as sufficient
evidence that "Copilot is not in play" without first trying the documented
working fallback conflates "not currently engaged" with "cannot be
engaged."

## Resolution / correct sequencing

Before relying on `autoharness gate copilot-review`'s `NOT_APPLICABLE`
verdict to justify skipping the P-018 hard-gate wait:

1. Check `docs/compound/` for any repository-specific Copilot-engagement
   fallback (e.g., the `[bot]`-suffixed REST POST documented in
   `2026-09-06-requesting-copilot-review-non-collaborator-repo.md`) and
   attempt it explicitly.
2. Poll per the §1.2 back-off cadence (Copilot shadow review typically
   completes within 2–5 minutes; timeout at 15 minutes) BEFORE treating
   the PR as ready to merge.
3. Only after the documented fallback has been tried and either (a)
   genuinely produces no reviewer-registration signal after a real attempt,
   or (b) the workspace has independently confirmed Copilot review is not
   licensed/available at the organization level, is it safe to treat
   `NOT_APPLICABLE` as the terminal verdict and proceed without waiting.
4. This sequencing must happen **before** merge — once a PR is
   merged/closed, GitHub silently no-ops any further reviewer-request POST
   against it (returns 200 with the PR object, but `reviewRequests` stays
   empty), so there is no way to retroactively engage Copilot review on an
   already-merged PR.

## Compounding value

A "gate returns NOT_APPLICABLE" check is necessary but not sufficient
evidence that "Copilot could not have been engaged" on a repository that
has a known-working non-default engagement path. Always cross-check
`docs/compound/` for a documented fallback and attempt it before treating
non-engagement as final, especially under an operator directive to "be
patient" with Copilot review (P-018). If a genuine miss occurs and the PR
is already merged, disclose it transparently in the closure report and run
a compensating independent review pass (fresh reviewer persona against the
merged diff) rather than attempting any unsafe post-hoc remediation
(re-opening, reverting, or force-re-engaging review on a closed PR).

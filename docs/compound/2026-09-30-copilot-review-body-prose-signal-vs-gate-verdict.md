---
title: "Copilot review: read full body prose for narrative-only findings, but trust the programmatic gate verdict over body-text 'quality'"
date: 2026-09-30
category: build-errors
tags: [github, copilot, pr-review, p-018, p-021, gate-verification]
---

# Copilot review body prose: two independent signals, neither substitutable for the other

## Symptom

Across 14 Copilot review rounds on PR #83 (shipment 035-S), two distinct
surprises recurred:

1. **Real, actionable findings sometimes exist only as narrative prose in
   the review's top-level `body` text, with zero inline comments and no
   review thread at all** (rounds 6 and 13 both produced this). Round 6's
   finding (a manifest non-string-expectation `TypeError`-surface
   divergence) and round 13's finding (a lost caller-cwd git precondition
   guard) were each discoverable ONLY by reading the full review body text
   end-to-end — neither had a `PullRequestReviewThread` node, a
   `databaseId`, or an inline comment of any kind. This is a **different**
   failure mode from the already-documented "Suppressed comments" section
   pattern (see
   `2026-09-11-copilot-suppressed-comments-vs-review-threads.md`): these
   findings were not under a labeled "Suppressed comments" heading at all —
   they were just... prose, in the middle of the review summary.
2. **Round 14's review body was near-empty boilerplate**
   ("This security-sensitive 53-file gate rewrite still has a current-HEAD
   Copilot review in progress." / "Findings: None", 228 characters total)
   — this initially looked like a stub, placeholder, or truncated/incomplete
   review, prompting an instinct to wait longer or re-request. It was, in
   fact, the genuine first fully-clean round: `autoharness gate
   copilot-review 83 --repo softwaresalt/intercom --enforcement auto --json`
   returned `"verdict": "SATISFIED"`, `"unresolved_thread_ids": []`,
   `"exit_code": 0`.

## Root cause

Copilot's review body is free-form prose generated per-round; its length,
tone, and structure carry **no reliability signal** about review
completeness or outstanding findings in either direction:

* A short, generic-sounding body does not mean the review is incomplete,
  stubbed, or still running — it can be the correct, terse output of a
  round that found genuinely nothing new.
* A long, detailed-sounding body is not the only place findings appear as
  structured, actionable items — some genuine findings are only ever
  narrated in free text with no corresponding thread, and are just as real
  and in-scope-checkable as inline-commented findings.

These are two independent axes (body prose richness vs. actual review
state) and conflating either with the other produces two opposite wrong
instincts: over-trusting a verbose body as "more reviewed" than a terse
one, or under-trusting a terse body as "not really done yet."

## Resolution / correct sequencing

1. **Always read the full review body text on every round**, not just the
   count/list of inline comments or GraphQL review threads. A narrative-only
   finding carries the same P-021 classification obligation as an
   inline-commented one, even though it has no thread to reply to or
   resolve (same threadless-path handling as the P-021 C2 defer-capture
   procedure's threadless branch).
2. **Never treat review body prose quality (length, genericness,
   boilerplate phrasing) as a completeness or convergence signal in either
   direction.** The only authoritative convergence signal is the
   programmatic gate: `autoharness gate copilot-review <pr> --repo <repo>
   --enforcement auto --json`, specifically its `verdict` and
   `unresolved_thread_ids` fields. A `SATISFIED` verdict with an empty
   `unresolved_thread_ids` list is sufficient to proceed even when the
   accompanying body text looks like a stub.
3. This complements (does not replace) the existing guidance in
   `2026-09-11-copilot-suppressed-comments-vs-review-threads.md`: that
   entry covers a *labeled* "Suppressed comments" section disagreeing with
   real threads; this entry covers *unlabeled* narrative findings embedded
   directly in the summary prose, plus the separate lesson about not
   over- or under-reading body-text tone as a completeness signal.

## Compounding value

When running a multi-round Copilot review loop (P-018), budget time to read
every round's full body text regardless of inline-comment count or body
length, and gate every merge/proceed decision exclusively on the
programmatic `autoharness gate copilot-review` verdict — never on a
subjective read of how "thorough" or "complete" the review body prose
sounds.

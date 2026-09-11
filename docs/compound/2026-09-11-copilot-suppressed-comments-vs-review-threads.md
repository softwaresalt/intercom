---
title: "Copilot review body 'Suppressed comments' can differ entirely from the actual posted review threads — always fetch reviewThreads via GraphQL"
date: 2026-09-11
category: build-errors
tags: [github, copilot, pr-review, graphql, p-018, p-021]
---

# Copilot review body "Suppressed comments" can differ entirely from the actual posted review threads

## Symptom

`gh api repos/<owner>/<repo>/pulls/<pr>/reviews --jq '...'` returns a
Copilot-authored review whose `body` includes a `<details><summary>Suppressed
comments (N)</summary>` section listing N findings inline in the markdown
text. It is tempting to treat this list as "the Copilot comments to
classify/reply/resolve" per P-021/P-018 obligations. On shipment 016-S
(PR #51), the two "suppressed comments" named in the review body
(`.backlogit/queue/016-S.md:12` scope-wording concern, and a
`reparse_windows_test.go:120` short-case-threshold concern) were **entirely
different findings** from the two real, actionable review threads returned
by a GraphQL `reviewThreads` query (a branch-retention-enforceability
overclaim at `reparse_windows_test.go:103`, and a missing-YAML-frontmatter
finding on the new memory checkpoint doc). `autoharness gate
copilot-review`'s `unresolved_thread_ids` also only ever reflected the
GraphQL-visible threads, never the review-body "suppressed" list.

## Root cause

GitHub's Copilot reviewer generates more candidate comments internally than
it always posts as inline review threads; comments it decides not to post
as threads (lower confidence, deduplicated, or otherwise suppressed) are
instead summarized in the review's own top-level `body` text under
"Suppressed comments," with **no corresponding `PullRequestReviewThread`
node, no `databaseId`, and no `in_reply_to` target** — they cannot be
replied to or resolved via the normal review-comment reply/resolve
mechanism because no comment record exists for them. They are informational
only. The set of "suppressed" comments and the set of actually-posted
review threads are independent; one is not a superset or subset of the
other in general.

## Resolution / correct sequencing

1. Never rely on parsing the Copilot review `body` text for "Suppressed
   comments" as the authoritative list of findings requiring P-021
   classification, reply, or thread resolution — those are display-only and
   have no thread to resolve.
2. Always fetch the actual review threads via GraphQL:
   ```
   query {
     repository(owner: "...", name: "...") {
       pullRequest(number: N) {
         reviewThreads(first: 20) {
           nodes { id isResolved path line comments(first: 5) { nodes { id databaseId author { login } body } } }
         }
       }
     }
   }
   ```
   The `unresolved_thread_ids` field of `autoharness gate copilot-review`'s
   JSON output is sourced from this same GraphQL surface and is the
   authoritative list of what actually blocks the P-018 gate.
3. If a "suppressed comment" in the review body still represents a genuine,
   valid, in-scope concern (as both of 016-S's suppressed comments turned
   out to be — a real test-coverage gap and a real PR-description clarity
   issue), it is reasonable to fix it proactively as good engineering
   practice even though no thread exists to reply to or resolve — there is
   simply no reply/resolve obligation for it since no thread was ever
   created. Do not fabricate a thread reply against a comment ID that does
   not exist.
4. Re-requesting Copilot review after a fix does not guarantee the same
   "suppressed" finding reappears in the body text, is promoted to a real
   thread, or is dropped entirely — its absence on the next review pass is
   not evidence of resolution and its presence is not evidence of an
   unresolved thread; only the GraphQL `reviewThreads`/`unresolved_thread_ids`
   surface tracks genuine blocking state.

## Compounding value

When processing "every Copilot comment" per an operator directive or P-021,
scope that obligation to actual `PullRequestReviewThread` nodes fetched via
GraphQL (or equivalently, the `unresolved_thread_ids` reported by
`autoharness gate copilot-review`), not to every finding mentioned anywhere
in the review's rendered body text. The two surfaces can disagree
completely, as they did on PR #51.

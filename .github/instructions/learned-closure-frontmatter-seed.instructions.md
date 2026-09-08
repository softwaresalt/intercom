---
status: "proposed"
applyTo: "**"
description: "PROPOSAL (not yet applied) — Ship should seed closure-doc frontmatter from a prior example before drafting, to avoid a recurring Copilot review finding."
source_instinct: ".autoharness/continuous-learning/instincts/2026-09-08-013-s-closure-instincts.md#instinct-1-closure-artifacts-missing-frontmatter-reliably-draw-a-copilot-finding-promoted"
source_observations:
  - ".autoharness/continuous-learning/observations/2026-09-05.jsonl"
  - ".autoharness/continuous-learning/observations/2026-09-08.jsonl"
evidence_count: 3
promotion_threshold: 3
generated: 2026-09-08
mode: propose
---

# PROPOSAL: Seed post-merge closure artifact frontmatter from a prior example

**This is a `learn`/`evolve` proposal awaiting operator review — it is not yet
applied as binding guidance.** Do not treat this file as authoritative until an
operator (or a later `evolve --mode apply` invocation) promotes it.

## Rationale

Across 3 independent shipments (008-S, 012-S, 013-S), Ship authored a
`docs/closure/*.md` post-merge closure artifact without the YAML frontmatter
block already established by every earlier closure doc in this repository
(`title`, `description`, `status`, `tags`, `date`, `mode`, `shipment`,
`feature`, `pr`, `merge_commit_sha`, `compaction_status`, `closure_status`,
`releasability`). In all 3 cases, Copilot's automated PR review caught the
omission on the closure PR, costing a fix → reply → resolve cycle each time
that could have been avoided entirely.

## Proposed addition

Add to the Ship agent template's Step 6 ("Post-Merge Closure"), immediately
before "Invoke `operational-closure` in `mode=post-merge`":

> Before drafting a new `docs/closure/*.md` artifact, copy the YAML
> frontmatter block from the most recently merged prior closure doc in
> `docs/closure/` (matching field names: `title`, `description`, `status`,
> `tags`, `date`, `mode`, `shipment`, `feature`, `pr`, `merge_commit_sha`,
> `compaction_status`, `closure_status`, `releasability`) and populate it with
> this shipment's own values before writing the prose body. Do not defer the
> frontmatter to a later editing pass or rely on review to catch its absence.

## Evidence

1. Branch `chore/008-s-closure-evidence-frontmatter-fix` — retrofitted
   frontmatter onto the 008-S closure doc after review.
2. Branch `post-merge/012-s-retire-d6a-gate-narrowing`, commit `f2afdd9` —
   "fix closure frontmatter schema, compound learning section ref, and
   compacted-memory frontmatter/paths per Copilot review (PR #37)".
3. This session — PR #41 (013-S post-merge closure), review thread
   `PRRT_kwDOTPuhps6gcjSY`, fixed in commit `3653ff0`.

## Recommendation

Promote to an **instruction** (stable rule/reminder), not a skill — this is a
single-step authoring reminder, not a multi-step workflow. Suggested location
if applied: `.github/instructions/learned-closure-frontmatter-seed.instructions.md`,
or fold directly into the Ship agent template's Step 6 preamble (preferred,
since it is Ship-specific and narrowly scoped).

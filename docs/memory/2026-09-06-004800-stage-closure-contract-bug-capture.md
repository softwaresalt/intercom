---
title: "Stage session — 008-S closure gate root cause captured as bug record"
date: 2026-09-06
agent: stage
session_type: bug-capture
branch: chore/008-s-closure-evidence-frontmatter-fix
outcome: documentation-only
bug_doc: docs/bugs/2026-09-06-closure-evidence-producer-consumer-contract-mismatch.md
stash_created: 1C1E5467
stash_preserved: 9F9A3CB2
---

## Scope

Operator requested root-cause identification for the 008-S closure gate block and a durable bug
record under `docs/bugs` for tool-integration follow-up. Explicitly no harvest, no plan, no
shipment, no commit, no branch change.

## Root cause established

Producer/consumer contract drift for post-merge closure evidence. `operational-closure` (SKILL.md
line 23) specifies `docs/closure/{YYYY-MM-DD}-{slug}-closure.md`; the topology reader
`closure_complete` (topology.py 654-666) globs `docs/closure/{shipment_id}-*-post-merge-closure.md`
and requires frontmatter passing `_closure_artifact_complete` (294-329). No filename satisfies both
patterns, so conforming producer output is undiscoverable. Discovery precedes parsing, so a
non-matching artifact is never opened and returns silent `None`, collapsed into
`PREDECESSOR_CLOSURE_INCOMPLETE` at 1611-1618.

## Key corroboration

Byte-identical content at the producer-spec filename returns `None`; at the consumer-spec filename
returns `True`. Filename divergence alone is sufficient. Malformed frontmatter under a matching
name raises `BacklogUnavailableError` (loud), so the likelier failure is the quieter one.

Recurrence: `6ed75e8` (001-S rename), `1baffd4` (005-S frontmatter), `385afd8` (008-S both). Three
per-file repairs, never a contract fix.

## Decisions

* Framed the durable fix as contract unification plus enforcement, not another rename.
* Required an integration test over the real producer and real consumer; a hand-written fixture
  would have passed throughout the defect's history and is explicitly insufficient.
* Flagged that the consumer ships in the installed autoharness package, not this repo, so the fix
  likely needs an upstream or template change. Scope during deliberation.

## Next step

Deliberate `1C1E5467` in a future Stage round, resolving the upstream-versus-local question before
planning. Keep `9F9A3CB2` (`.gitignore` / `start.ps1` carryover) as a separate release unit.

## Preservation confirmed

`.gitignore` and `start.ps1` untouched and still dirty; `docs/closure/` clean; 009-S still
`queued` and unmodified; no commits, pushes, branches, worktrees, or PRs created.

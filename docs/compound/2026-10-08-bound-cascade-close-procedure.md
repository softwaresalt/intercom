---
title: "Bound cascade close: classify, recompute, compare, then ship (and closure gotchas)"
date: 2026-10-08
category: process
tags: [shipment-reconcile, cascade-close, classification-binding, p-015, p-005, backlogit, git, powershell, 040-s]
---

# Bound cascade close: classify, recompute, compare, then ship

## Context

No executable `shipment-reconcile` `safe-close` CLI or gate exists here (stash D10D3AFC, active). Every closure therefore applies the skill prose by hand and invokes `backlogit shipment ship` directly.
- 035-S, 036-S, 039-S and 033-S did so without a fresh-snapshot recompute-and-compare.
- 037-S added a manual recompute-and-compare before the call.
- 040-S (PR #112, merge `49ff6b9`) is another manual application of the 037-S approach: the binding was recomputed from a fresh snapshot, matched, and the primitive was then called directly. It is not tool-enforced.

The evidence is in `.backlogit/reconcile/040-S-classify-close-path-2026-10-08T06-49Z.md` and `040-S-safe-close-2026-10-08T06-54-43Z.md`.

It was not fully conforming:
- It skipped the mandated `file-lock` single-writer lock (procedural deviation, tracked by 9F824B64).
- It wrote the helper script and the backup outside the workspace (Principle IV).

The gotchas below show how to avoid both.

## Practice

1. **a1 first.** Move the covering feature `active -> done`.
   - backlogit 1.11 auto-archives it to `.backlogit/archive/` while it still declares `status: done`.
   - That makes it "pre-archived by location", which is tolerated on the cascade path. It is not truly archived.
2. **Pre-mode at `expected_status: done`.** Every member is `pre-archived` and the record is `active` (`record-consistent`), so the result is `PROCEED`.
3. **classify-close-path (read-only).**
   - Take the snapshot from frontmatter and walk `parent_id` to a fixed point over queue plus archive.
   - Check the linked deliberations: `source_deliberation_id` plus the `\b(?:DL\d+|[0-9]+(?:\.[0-9]+)*-DL)\b` scan. A docs-path reference is not a match.
   - Compute the v1 binding, and **validate the serializer by reproducing a known precedent digest** (036-S `b8ba3fba…`) before trusting it.
4. **safe-close Step 0.** Retake the snapshot fresh, recompute, and compare against the supplied binding.
   - Only on MATCH, follow the Cascade Close Sub-Procedure steps and run `backlogit shipment ship <id> --sha <merge> --message <subject> --author <author>`. Because there is no `safe-close` tool (D10D3AFC), this is a direct, agent-executed call: disclose it as not tool-enforced.
   - Then verify that `returned_ids == []`, that both two-set differences are empty, and that every `parent_id` is preserved.
5. Back up `.backlogit/queue` and `.backlogit/archive` before the cascade so a revert is possible. The cascade took about 4.5 minutes for 7 records.

## Gotchas

- **The lock scripts live under the skill, not the repo root.**
  - The `file-lock` scripts are at `.github/skills/file-lock/scripts/acquire_lock.{ps1,sh}`. The Ship agent text cites `scripts/acquire_lock.ps1`, which does not exist here.
  - Check the skill directory before declaring the lock unavailable.
  - The skill grants no single-agent exemption for Ship Step 6. Skipping the lock is a deviation, not a degradation.
- **Keep helpers and backups inside the workspace.**
  - Write the binding helper script and the pre-cascade backup under a gitignored workspace path such as `dist/`, not `%TEMP%`.
  - CLI workspace containment (Principle IV) forbids writes outside the working tree.
- **PowerShell comma precedence.**
  - Inside an array literal, `"manifest=" + ($ids -join ',')` binds as `(prev, "manifest=") + …`. This silently splits the line and changes the digest.
  - Build canonical lines one statement at a time (`List[string].Add`).
- **`git add` is atomic on a bad pathspec.**
  - `backlogit shipment ship` stages the shipment's `queue -> archive` rename itself.
  - Naming the now-missing `.backlogit/queue/<id>.md` in `git add` fails the WHOLE add (`fatal: pathspec … did not match`). A following `git commit` then records only the pre-staged rename.
  - Use `git add -A -- <removed path>` for deletions and check `git show --stat` before pushing.
  - The staged rename also captured the pre-close content, so `.backlogit/archive/<id>.md` has to be re-added after the engine rewrites it.
- **Engine line.**
  - `backlogit --version` appends a network update-check suffix.
  - Precedent bindings include the full first line, so do not pass `--no-update-check` to the version call used for `engine=`.
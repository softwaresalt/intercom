## Step Sequence Contract (NON-NEGOTIABLE)

Every Stage session MUST execute the following steps in order.

```text
[ ] Step 1   — Stash triage and entry classification
[ ] Step 1.8 — Learnings retrieval
[ ] Step 2   — Deliberation
```

### Step 1.9: Stage Artifact Branch Gate (NON-NEGOTIABLE)

This gate is fail-closed and runs strictly after Step 1.8 and strictly before Step 2. It
establishes (or resumes) the dedicated Stage artifact branch `chore/stage-{scope-slug}`
before any tracked Stage artifact mutation this session.

1. **Worktree topology precheck (P-016) — runs first, before any branch create, select, or
   resume.** Run `git worktree list --porcelain` and classify every attached worktree.

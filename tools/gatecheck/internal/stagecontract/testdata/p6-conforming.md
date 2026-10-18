## Step Sequence Contract (NON-NEGOTIABLE)

Every Stage session MUST execute the following steps in order.

```text
[ ] Step 1   — Stash triage and entry classification
[ ] Step 1.8 — Learnings retrieval
[ ] Step 1.9 — Stage artifact branch gate (fail-closed)
[ ] Step 2   — Deliberation
```

### Step 1.9: Stage Artifact Branch Gate (NON-NEGOTIABLE)

This gate is fail-closed and runs strictly after Step 1.8 and strictly before Step 2. It
establishes (or resumes) the dedicated Stage artifact branch `chore/stage-{scope-slug}`
before any tracked Stage artifact mutation this session.

1. **Worktree topology precheck (P-016) — runs first, before any branch create, select, or
   resume.** Run `git worktree list --porcelain` and classify every attached worktree.

7. **Named fail-closed halt and categorical deferral rule.** On any failure of steps 1–6
   the gate halts with the named token `STAGE_BRANCH_GATE_FAIL`, records a P-005
   policy-violation event, and performs no tracked mutation.

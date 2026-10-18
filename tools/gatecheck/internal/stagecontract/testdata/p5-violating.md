## Role Boundary (NON-NEGOTIABLE)

| Category | Allowed | Forbidden |
|---|---|---|
| Git | Create and check out a dedicated Stage artifact branch (`chore/stage-{scope-slug}`); commit backlog/planning artifacts on that branch only; create/use an explicit, time-boxed spike/research worktree only for staging investigation | Create or checkout feature/chore branches for code execution; create/use parallel implementation branches or worktrees |
| Build | — | Run build systems, test suites, or linters |
| PR | Stage pushes the Stage artifact branch and opens the resulting staging PR | Merge pull requests |

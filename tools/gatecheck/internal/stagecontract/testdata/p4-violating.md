# P-016: No Parallel Branch/Worktree Execution

**Statement**: Agents MUST NOT work across parallel implementation branches or worktrees. At any point where an agent will mutate the workspace for implementation, backlog execution, PR preparation, or closure, there must be exactly one agent-owned implementation branch/worktree in use.

**Precondition**: Before any agent claims a shipment, creates a branch, creates/uses a worktree, or mutates the workspace, the agent has checked current branch/worktree state.

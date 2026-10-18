# P-016: No Parallel Branch/Worktree Execution

**Statement**: Agents MUST NOT work across parallel implementation branches or worktrees. At any point where an agent will mutate the workspace for implementation, backlog execution, PR preparation, or closure, there must be exactly one agent-owned implementation branch/worktree in use.

**Allowed exception (Stage spike/research only)**: Stage MAY create or use a separate worktree only for an explicit, time-boxed spike or research investigation during staging. That worktree MUST NOT be used for implementation, template/source/config mutation, shipment claim, PR preparation, or Ship execution. Stage MUST record the spike/research context and clean up the worktree or hand off findings before Ship execution begins.

**Precondition**: Before any agent claims a shipment, creates a branch, creates/uses a worktree, or mutates the workspace, the agent has checked current branch/worktree state.

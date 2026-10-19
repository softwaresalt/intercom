# 057-S (065-F) compacted memory

- Scope: MVP S2 agent adapter anti-corruption layer. Tasks 065.001-T..065.007-T done. PR #175 merged as a merge commit e43ad660aa2f0bd98c9eff58833cbfbe0e4963a2 (reviewed HEAD 35cdacc).
- Delivered: `internal/agent` (domain types, Agent interface, sentinels) and `internal/agent/copilot` (single SDK adapter: normalization table, permission bridge that fails closed and denies unreviewable requests, one-way callback latch, version/protocol assertion at Start, bounded cancellation-aware Stop, single-flight ForceStop); `[copilot].startup_timeout_seconds`; `internal/copilotprobe` deleted; depguard allowlist is exactly `internal/agent/copilot/*.go`.
- Review: three local cycles (single-pass adversarial, six-persona, three-persona) plus one focused delta review; Copilot raised four threads (bounded ForceStop, cancellation-aware Stop, unignore denylist marker), all fixed, replied, resolved.
- Lessons: `.gitignore` is append-only (I6), so move fixtures by appending, never renaming; the unignore denylist golden counts entries. The system Go install was replaced mid-session; the cached go1.26.9 toolchain (same as the go.mod pin) was used with `GOTOOLCHAIN=local`.
- Closure: a1 moved 065-F active to done; classify-only CASCADE (engine VERIFIED); operator-approved mutating cascade-close exited 0, postcondition pass.
- Deferred scope (stash): 56CE7077, E7B78125, 6738A8E5, DA760FBB, FD093EFC; follow-up P3 bundle FA1C38F7.
- Originals: docs/archive/memory/2026-10-19/057-s-ship-session.md.

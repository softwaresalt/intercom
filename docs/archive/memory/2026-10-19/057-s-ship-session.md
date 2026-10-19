# Ship session memory: 057-S (MVP S2 agent adapter anti-corruption layer)

Date: 2026-10-19. Branch: feat/057-s-mvp-s2-agent-adapter-anti-corruption-layer. Dark mode (P-017) active.

## Completed

* 065.001-T..065.007-T done (TDD: tests observed failing first, then implemented).
* New packages: `internal/agent` (domain types, Agent interface, sentinels) and `internal/agent/copilot`
  (single SDK adapter: normalization table, permission bridge, callback latch, Start/Send/Abort/Stop).
* `[copilot].startup_timeout_seconds` (default 30, 1..3600) folded into validation rule 5.
* `internal/copilotprobe` deleted; depguard re-pointed at `internal/agent/copilot/*.go` only;
  depguard fixture script, gitignore marker and gatecheck unignore denylist moved to `internal/agent/copilot2`.

## Review (local, pre-PR)

* Cycle 1 (single-pass adversarial review): P1 Start/Stop race fixed, bounded Stop, bridge released by stop,
  version-floor semantics (`-N` is prerelease), summary bound, latch zero-timeout, non-blocking-first send.
* Cycle 2 (six parallel personas): OnEvent registered at creation, truncated requests denied, sanitization,
  PermissionKind enum, per-kind summaries, sentinels, Stop cancels in-flight Send/Abort, Resolved on asks.
* Cycle 3 (concurrency, security, correctness): no P0-P2; P3 follow-ups stashed.

## Decisions

* Standard SDK mode (no ModeEmpty) per ADR MVP-1.
* Version floor `1.0.84-1` treats `-N` as prerelease (final 1.0.84 satisfies it).
* Deferred scope (stash): 56CE7077 UNC cli_path policy, E7B78125 warn rate limit, 6738A8E5 hub wiring/ordering,
  DA760FBB event bounds/enrichment, FD093EFC cli_path hardening. Follow-up P3 bundle: FA1C38F7.

## Next

Done: PR 175 merged as merge commit e43ad660aa2f0bd98c9eff58833cbfbe0e4963a2 (reviewed HEAD 35cdacc). Copilot review: 4 threads fixed, replied and resolved over multiple rounds; P-018 gate SATISFIED. Post-merge: 065-F done via the a1 gate; classify-only CASCADE (engine VERIFIED); operator pre-approved mutating cascade-close exited 0 (postcondition pass).

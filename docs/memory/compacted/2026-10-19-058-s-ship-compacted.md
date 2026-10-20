# 058-S (066-F) compacted memory

- Scope: MVP S3 hub core, envelope, hydration, remote config. Tasks 066.001-T..066.009-T done test-first (red phase observed per task). PR #177 merged as a merge commit ec447f7e7fe6140bc99da28802e96b4317860aa3; cascade close done; dark mode (P-017) active.
- Delivered: `internal/hub` (envelope v1, UUIDv7 `gen`, bounded fold, sole-writer hub, hydrate and replay ring, capacity-1 degraded channel, `Hub.Send`, caps on frames, ring bytes, remote clients, connect intake); `[remote]` config group (`internal/config/remote.go`) with secure defaults, fail-closed validation and credential-key screen; `config.toml.example`, `docs/config-reference.md`, design section 8; integration test `internal/agent/copilot/hub_ingress_test.go` (test-only).
- Review: three pre-PR cycles. Cycle 1 (six personas) fixed two P1 doc-vs-code findings (host ports, default-port origins). Cycle 2 (three personas plus three-reviewer adversarial consensus) fixed P1 fold string pinning and unbounded connect intake. Cycle 3 (Concurrency, Security) on c3ecf1f: no P0/P1/P2, READY_WITH_FOLLOWUPS.
- Decisions: UUIDv7 from crypto/rand plus clock (no new dependency); TUI overflow marks stale, drops oldest, forces re-hydrate, never drops the TUI member; degraded and unknown-event detail never reaches the wire or logs (fixed messages plus byte lengths); `allowed_hosts` accepts an optional port and bracketed IPv6, origins reject explicit default ports (ADR MVP-2).
- Lessons: system Go 1.27 breaks the installed golangci-lint; use `GOTOOLCHAIN=go1.26.9` for lint and tests.
- Degradations: telemetry begin/record and per-task harness-architect invocation not run separately.
- Deferred scope (stash): 2C74C5FC, E02CA4AA, FC76C5D8.
- Next: closure PR through CI and Copilot loop, then confirm 059-S pre_claim eligibility on main.
- Original: docs/memory/2026-10-19/2026-10-19-ship-058-s-session-memory.md (left in place; session still open for closure PR).

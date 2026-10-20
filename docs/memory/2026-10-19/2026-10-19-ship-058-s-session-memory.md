# Ship session memory: 058-S (MVP S3 hub core, envelope, hydration, remote config)

Date: 2026-10-19. Branch: feat/058-s-mvp-s3-hub-core-envelope-hydration-remote-config.
Mode: DARK_MODE_ACTIVE (P-017), operator AFK, scope 058-S only in this session.

## Completed

- 066.001-T..066.009-T implemented test-first (red phase observed per task) and moved to done.
- New package `internal/hub` (envelope v1, UUIDv7 `gen`, bounded fold, sole-writer hub, hydrate and replay ring,
  capacity-1 degraded channel, `Hub.Send`, caps on frames, ring bytes, remote clients, connect intake).
- `[remote]` config group (`internal/config/remote.go`), secure defaults, fail-closed validation, credential-key screen,
  `config.toml.example`, `docs/config-reference.md`, design section 8 text.
- 066.008-T integration test lives in `internal/agent/copilot/hub_ingress_test.go` (test-only, no adapter code changes).

## Review (pre-PR, three cycles)

- Cycle 1: Concurrency, Correctness, Security, Go, Maintainability, Schema-CLI-Docs personas. Two P1 doc-vs-code contract
  findings (host ports, default-port origins) fixed with the P2 set.
- Cycle 2: Concurrency, Security, Correctness personas plus the Adversarial Review consensus (3 reviewers). P1 items
  (fold string pinning, unbounded connect intake) fixed.
- Cycle 3: Concurrency and Security personas on c3ecf1f: no P0/P1/P2, READY_WITH_FOLLOWUPS.
- Out-of-scope residue captured capture-first as DEFERRED SCOPE EXPANSION stash entries: 2C74C5FC, E02CA4AA, FC76C5D8.

## Decisions

- UUIDv7 built from crypto/rand plus clock (no new dependency).
- TUI overflow: mark stale, drop-oldest, forced re-hydrate, never drop the TUI member.
- Degraded and unknown-event detail never goes on the wire or to logs; fixed messages plus byte lengths only.
- allowed_hosts accepts an optional port and bracketed IPv6; origins reject explicit default ports (ADR MVP-2 canonical form).
- The system Go 1.27 toolchain breaks the installed golangci-lint; use `GOTOOLCHAIN=go1.26.9` for lint and tests.

## Degradations

- Telemetry begin/record and per-task harness-architect invocation were not run separately; the tasks carried the
  harness-ready label and each was built test-first with red evidence reported.

## Next

Open PR, wait for CI and Copilot review, readiness gate, merge, then post-merge closure.

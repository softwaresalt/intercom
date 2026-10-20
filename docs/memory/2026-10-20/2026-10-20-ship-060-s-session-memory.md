# Ship session memory - 060-S (MVP S5: authenticated WebSocket transport)

- Shipment: 060-S, covering feature 068-F, tasks 068.001-T..068.008-T (all done).
- Branch: feat/060-s-mvp-s5-authenticated-websocket-transport.
- Mode: DARK_MODE_ACTIVE (P-017); operator AFK; admin fallback not authorised.

## Items completed

All eight tasks implemented test-first in the new `internal/transport` package and the `internal/transport/wsclient` reference client: loopback-only listener with bind assertion, per-role 256-bit credentials (digest-only storage, constant-time verify, redacting token type), pre-upgrade authentication (Authorization header or handler-selected subprotocol, never URL), Origin allowlist (default empty) and Host validation, pre-auth connection cap counting unauthenticated sockets plus handshake timeout, single-writer write pump with drop protocol, read pump with connect/replay/hydrate, per-action re-validation with rate limit, reference client and integration tests.

## Decisions with rationale

- `github.com/coder/websocket` v1.8.15 (ADR default) promoted from indirect to direct; no other dependency added; govulncheck clean.
- Origin is checked by our own exact-match pre-check; the library check is not loosened (`InsecureSkipVerify` false).
- Connection cap enforced in a listener wrapper at accept time so unauthenticated sockets consume slots; keep-alives disabled so rejected sockets free slots at once.
- The library itself writes the 1009 close and pong/close-echo frames from Read; recorded in design 5.5 as an accepted, library-serialised exception. All application writes go through the pump.
- Shutdown-revoked credentials are retained so a presented credential gets 503 (retry) rather than 401; only a holder of that credential can observe the difference.
- Resume cursor travels as `?gen=&seq=` (not secrets).

## Review

Cycle 1: seven persona subagents (Security, Concurrency, Correctness, Go, Maintainability, Architecture, Schema-CLI-Docs): no P0/P1; 19 fix items applied. Delta reviews + Adversarial Review (3 reviewers): no HIGH-confidence findings. Cycle 2: 8 fix items; Security + Concurrency delta reviews: no P0/P1.

## Follow-up stash (DEFERRED SCOPE EXPANSION)

7EF4DA98 (composition/Gate/atomic revalidation), FE0D5ECD (pre-auth DoS hardening), 26FD9AB8 (maintainability). Stash 026CF750 was only implemented where a 060-S task required it.

## Operational notes

- Completing a task with `backlogit move --status done` archives the task file automatically.
- `backlogit` commands are run one at a time.

## Next steps

Open PR, CI + Copilot review loop, merge with a merge commit, post-merge closure.

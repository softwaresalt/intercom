---
title: "Ship session checkpoint: shipment 077-S (feature 084-F)"
date: 2026-10-31
shipment: 077-S
feature: 084-F
branch: feat/077-s-transport-wire-authz-hardening-atomic-hub-step-credential-revalidation-rejected-upgrade-accounting-and-pumprejects-rate-limiting
status: pr-preparation
---

## Items completed

* 084.001-T hub `ConnectOptions.Validate` atomic hub-step revalidation (5107b42)
* 084.002-T wire feedback bounds, local-only refusals, close on hub-step failure (24755a8)
* 084.003-T off-accept-loop over-cap 503 refusal worker and listener `OnReject` (a849ab7)
* 084.004-T `RecordReject` accounting, handshake validation, `grantValidator` (c799270)
* 084.005-T depguard layering rules (84dcf99)
* 084.006-T wsclient URL fragment/opaque/host checks and conformance tests (e6c3f53)
* 084.007-T rate-limited `pumpRejects` (82a11d2)
* 084.008-T ADR MVP-2 and runbook updates (616f07a)
* Review fixes: validator leak on hydrate failure, refusal queue drain race (dc065d1)

## Decisions and rationale

* Followed the plan's real file names (`session.go`, `limiter.go`, `listen.go`,
  `docs/runbooks/...`) where the dispatch brief named approximate files.
* `dispatchWire` closes the socket after a hub-step credential failure only when
  `revalidate` reports `closeConn`, so shutdown revocation keeps the drain open.
* Config conformance test asserts containment (config accepts implies transport
  accepts); the transport canonicalizes and is intentionally more lenient.
* Telemetry context was `disabled`; skipped per Step 4.1a.

## Review

* Local review (security, concurrency, architecture, correctness personas):
  no P0/P1. P3 fixes applied. P2 (wsclient still imports `internal/hub`)
  deferred as stash entry 7919A9A1 (P-021 C2).

## Observations

* `internal/tui` failed once in a full `go test -race ./...` run under load and
  passed on three reruns. The package is untouched by this shipment.

## Next steps

* Push, open PR with Local Review Readiness block, poll CI, run the copilot-review
  gate, merge (merge commit), then post-merge closure up to the cascade-close
  classify-only step and stop for per-closure operator approval.

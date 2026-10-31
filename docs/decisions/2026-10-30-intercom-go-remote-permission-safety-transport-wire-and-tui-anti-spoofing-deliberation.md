---
title: "Remote Permission Safety, Transport Wire Integrity, and TUI Anti-Spoofing Deliberation"
description: "Deliberation on Capability Set 1: resolving deferred scope expansions ACD578AE, D46C1B38, 7BDC38FF, 026CF750, 7EF4DA98, FE0D5ECD, 86BE6345, 21F9D21A, and 6617141A across internal/agent/copilot, internal/hub, internal/tui, internal/transport, internal/transport/wsclient, internal/app, .golangci.yml, and docs."
topic: "Remote Permission Safety, Transport Wire Integrity, and TUI Anti-Spoofing"
depth: "deep"
decision_status: "decided"
promoted_to: "plan"
linked_artifacts:
  - "docs/plans/2026-10-30-intercom-go-remote-permission-safety-transport-wire-and-tui-anti-spoofing-plan.md"
  - ".backlogit/queue/083-F.md"
  - ".backlogit/queue/084-F.md"
  - ".backlogit/queue/076-S.md"
  - ".backlogit/queue/077-S.md"
tags:
  - "permissions"
  - "authz"
  - "transport"
  - "wsclient"
  - "tui"
  - "security"
---

## Problem Frame

Following the merge and post-merge operational closure of the `074-S..075-S` shipment chain (`075-S` merged at `a2898942a12c70fe23be050d698ae4622153d97f`, closure commit `d5b209e80ee39baa520a85175ebda1f00a6e3f2f`), the operator selected **Option 1: Capability Set 1 — Remote Permission Safety, Transport Wire Integrity & TUI Anti-Spoofing** for execution in dark-factory mode (`DARK_MODE_ACTIVE`, P-017). Capability Set 1 comprises nine active stash entries captured during `059-S`, `060-S`, `065-S`, `071-S`, and `074-S`:

1. **`ACD578AE`** (`medium` / `task`): Decide whether the hub should refuse remote approval of `PreviewClipped` write permissions (`authzInput` has no clipped flag, so a remote approver with `remote_approve_destructive=true` can approve a clipped write; only the local TUI gets the warning and `APPROVE CLIPPED WRITE` confirmation).
2. **`D46C1B38`** (`low` / `task`): Harden `shellSummary` `urls:` anti-spoofing and write/shell summary sanitization against visually blank non-space non-`Cf` runes (`U+2800`, `U+3164`, `U+115F`, `U+FFA0`) and homoglyph or fullwidth lookalikes of `urls:` that bypass `indentURLsFraming` and `sanitizeBound`; remote WebSocket clients render the raw `Summary`.
3. **`7BDC38FF`** (`low` / `task`): Introduce a single shared helper/invariant for indenting all-but-first lines across every TUI modal header (beyond `kind:` and `tool call:`).
4. **`026CF750`** (`high` / `feature`): S5 transport-owned authz/wire items surfaced by `059-S`: per-message credential re-validation (`credential_expired`, `generation_mismatch`), `action.refused` wire message shape including action identifier, `local_only_action` refusals for credential reveal/quit, `prompt`/`Feedback` size bounds (cap `Feedback` and clear unless deny), `ConnectOptions` `Kind`/`Role` must never derive from client input.
5. **`7EF4DA98`** (`medium` / `task`): `060-S` transport composition for `062-S`: composed transport constructor wiring `config.RemoteConfig->Config/ServerConfig/CredentialStore` (`OnBind`/`Drain` pairing), `hub.Gate` adapter (`RevokeCredentials+CloseListener` using `Server.CloseListener`), `depguard` layering rule for `internal/transport`, `config<->transport` conformance test, wire-constant drift test (`wsclient` vs server), atomic hub-step revalidation (carry grant into hub action; close TOCTOU between revalidate and `Client.Do` and `Connect`).
6. **`FE0D5ECD`** (`low` / `task`): `060-S` pre-auth/DoS hardening beyond the connection cap and handshake timeout: per-IP or per-credential connection and upgrade rate limits (tunnel peers look loopback), reserved authenticated slots, rejection-log sampling, `Accept`-before-`Connect` or `Sec-WebSocket-Version`/`Key` validation before hub registration, `503` refusal write off the accept goroutine, `Host` default-port equivalence ratification (ADR MVP-2 vs matcher), reject URL fragment/opaque and IP-literal-only `ws://` in `wsclient`, TLS enforcement between tunnel and client.
7. **`86BE6345`** (`low` / `task`): Count over-cap `503` refusals from `internal/transport/limiter.go` (`limitListener.Accept`) and wrong-path `404`s as rejected upgrades, without blocking the accept loop, so `max_connections` slot exhaustion is visible in the TUI status line; also exclude or label the benign draining `503`.
8. **`21F9D21A`** (`low` / `task`): Count library-generated `websocket.Accept` handshake refusals (e.g. an authenticated `/ws` request with an unsupported non-empty `Sec-WebSocket-Version`, `internal/transport/handler.go` near the `Accept` call) as rejected upgrades without writing a second response, with an unsupported-version regression test.
9. **`6617141A`** (`low` / `task`): Rate-limit `pumpRejects` forwarding of rejected-upgrade notices to at most one `prog.Send` per interval (e.g. 250 ms ticker) so a remote flood of bad upgrades cannot drive proportional TUI `Update` traffic on the permission-modal channel.

Under P-021 C6, every entry carrying the `DEFERRED SCOPE EXPANSION` marker must pass through deliberation before planning and harvest. In addition, Ship Step 6.1(a1) `S2` enforces `n == 1` covering feature per shipment (`A1_MULTIPLE_FEATURE_MEMBERS` halts if `n > 1`), so this deliberation defines the covering-feature and shipment boundary architecture (`076-S` covering `083-F`, chained into `077-S` covering `084-F`).

## P-021 C5/C6 Triage Audit: Unconditional Duplicate Scan and Late-Identifier Reconciliation

### 1. Unconditional Duplicate Scan (Trigger A)

An unconditional pairwise scan was executed across all 143 active stash entries in `.backlogit/stash.jsonl`, all 165 archived stash entries in `.backlogit/archive/stash.jsonl`, and all items in `.backlogit/queue/`:

* **`ACD578AE`**: **CLEAN DUPLICATE SCAN**. Archived entry `773B2CB2` shipped `PreviewClipped` propagation and local-TUI warnings in `074-S` (`081-F`) and explicitly spawned `ACD578AE` for the remote-approval authorization policy decision. No other entry duplicates `ACD578AE`.
* **`D46C1B38`**: **CLEAN DUPLICATE SCAN**. `081.001-T` (`074-S`) shipped `unicode.IsSpace`/`Cf` sanitization and ASCII `urls:` indentation; `D46C1B38` is the sole entry covering visually blank non-space non-`Cf` filler runes (`U+2800`, `U+3164`, `U+115F`, `U+1160`, `U+FFA0`, `U+034F`, `U+17B4`, `U+17B5`) and `urls:` homoglyph/fullwidth lookalikes in `internal/agent/copilot/permission.go`.
* **`7BDC38FF`**: **CLEAN DUPLICATE SCAN**. `078.002-T` (`071-S`) added `wrapHeader` for `kind:` and `tool call:`; `7BDC38FF` is the sole entry covering a shared multi-header helper and column-0 header invariant across every TUI modal header.
* **`026CF750`**: **CLEAN DUPLICATE SCAN**. Partially implemented in `060-S` (`068-F`); no other entry duplicates its remaining open items (`local_only_action` refusals for credential reveal/re-issue and quit/shutdown, and `Feedback` clearing on non-deny decisions plus rune-safe `Feedback` size capping across transport and hub).
* **`7EF4DA98`**: **CLEAN DUPLICATE SCAN**. The composed `NewRemote` constructor and `hub.Gate` adapter (`RevokeCredentials` + `CloseListener`) were absorbed in `062-S` (`070.001-T`); no other entry duplicates the remaining four items (`depguard` layering rule for `internal/transport`, `config<->transport` conformance test, `wsclient` vs server wire-constant drift test, and atomic hub-step credential revalidation closing the TOCTOU window between `revalidate`/`Verify` and `Client.Do`/`Hub.Connect`).
* **`FE0D5ECD`**: **CLEAN DUPLICATE SCAN**. No other entry duplicates pre-registration WebSocket handshake header validation, off-accept-goroutine `503` refusal writes, `Host` default-port non-equivalence ratification in ADR MVP-2, and `wsclient` URL fragment/opaque rejection.
* **`86BE6345`**: **CLEAN DUPLICATE SCAN**. No other entry duplicates counting over-cap `503` refusals from `limitListener.Accept` and wrong-path `404`s in `RejectedUpgrades` / `OnReject` and excluding the benign draining `503`.
* **`21F9D21A`**: **CLEAN DUPLICATE SCAN**. No other entry duplicates counting library-generated `websocket.Accept` handshake refusals without writing a second response.
* **`6617141A`**: **CLEAN DUPLICATE SCAN**. Active entries `7BFA4182` and `26E15024` cover structural file/method relocation of `watchUI`/`pumpRejects` into `localUI`/`ui.go`; `6617141A` is the sole entry covering interval rate-limiting of `pumpRejects` `prog.Send` calls.

### 2. Late-Identifier Reconciliation (Trigger B)

Using the shipment closure records in `docs/closure/`, all missing PR numbers and merge commit SHAs were recovered and reconciled in place via `backlogit stash edit` after passing the Step 1.9 branch gate:

| Entry | Originally Captured Source Refs | Reconciled Source Refs & Retrieval Source | Disposition |
|---|---|---|---|
| `ACD578AE` | `task 081.003-T/081.004-T, feature 081-F, shipment 074-S, PR: N/A (pre-PR local review), review-thread: N/A` | `PR #223 (merge commit 966f6a977ef6590e1d7a9155c3f64d0ee8861061, pre-PR local review), review-thread: N/A` recovered from `docs/closure/074-S-081-F-post-merge-closure.md`. | Reconciled in place via `backlogit stash edit ACD578AE` |
| `D46C1B38` | `task 081.001-T, feature 081-F, shipment 074-S, PR: N/A (pre-PR local review), review-thread: N/A` | `PR #223 (merge commit 966f6a977ef6590e1d7a9155c3f64d0ee8861061, pre-PR local review), review-thread: N/A` recovered from `docs/closure/074-S-081-F-post-merge-closure.md`. | Reconciled in place via `backlogit stash edit D46C1B38` |
| `7BDC38FF` | `task 078.002-T, feature 078-F, shipment 071-S, PR N/A, review-thread N/A` | `PR #212 (merge commit 67d3f5c03e893dde348ecdec08c6fd7b4c69a686, pre-PR local review), review-thread N/A` recovered from `docs/closure/071-S-078-F-post-merge-closure.md`. | Reconciled in place via `backlogit stash edit 7BDC38FF` |
| `026CF750` | `task 067.004-T/067.005-T, feature 067-F, shipment 059-S, PR N/A, thread N/A` | `PR #179 (merge commit 8c9f0f299fb2784ec0cbe03b843c84402927595d, pre-PR local review), thread N/A` recovered from `docs/closure/059-S-067-F-post-merge-closure.md`. | Reconciled in place via `backlogit stash edit 026CF750` |
| `7EF4DA98` | `task 068.00N-T, feature 068-F, shipment 060-S, PR N/A (pre-PR local review), thread N/A` | `PR #181 (merge commit 24e381e0d6a7a81909707ec73a5954360fc805f1, pre-PR local review), thread N/A` recovered from `docs/closure/060-S-068-F-post-merge-closure.md`. | Reconciled in place via `backlogit stash edit 7EF4DA98` |
| `FE0D5ECD` | `task 068.00N-T, feature 068-F, shipment 060-S, PR N/A (pre-PR local review), thread N/A` | `PR #181 (merge commit 24e381e0d6a7a81909707ec73a5954360fc805f1, pre-PR local review), thread N/A` recovered from `docs/closure/060-S-068-F-post-merge-closure.md`. | Reconciled in place via `backlogit stash edit FE0D5ECD` |
| `86BE6345` | `feature 072-F; shipment 065-S; PR N/A (pre-PR local review); review thread N/A` | `PR #195 (merge commit 427bfbd5a3fcb22e117914c869a7faf2ab3d06c5, pre-PR local review); review thread N/A` recovered from `docs/closure/065-S-072-F-post-merge-closure.md`. | Reconciled in place via `backlogit stash edit 86BE6345` |
| `21F9D21A` | `feature 072-F; shipment 065-S; PR 195; review thread PRRT_kwDOTPuhps6wl2fF` | `PR #195 (merge commit 427bfbd5a3fcb22e117914c869a7faf2ab3d06c5); review thread PRRT_kwDOTPuhps6wl2fF` recovered from `docs/closure/065-S-072-F-post-merge-closure.md`. | Reconciled in place via `backlogit stash edit 21F9D21A` |
| `6617141A` | `feature 072-F; shipment 065-S; PR N/A (pre-PR local review); review thread N/A` | `PR #195 (merge commit 427bfbd5a3fcb22e117914c869a7faf2ab3d06c5, pre-PR local review); review thread N/A` recovered from `docs/closure/065-S-072-F-post-merge-closure.md`. | Reconciled in place via `backlogit stash edit 6617141A` |

## Research Findings

### Compound Learnings Consulted

* `docs/compound/2026-10-20-tui-permission-modal-safety-and-linux-test-clock.md`: Documents sharing a single predicate and line-count calculation between `permissionBody` and `m.approvable()`, keeping display sanitization off the user-typed prompt input path, and pinning both `Clock` and `Now` seams in tests.
* `docs/compound/2026-10-20-websocket-close-semantics-and-linux-ci-pitfalls.md`: Documents `coder/websocket` close semantics, `ConnContext` slot accounting, avoiding RFC 6455 sample `Sec-WebSocket-Key` literals in tests, and polling for asynchronous prompt handoff.
* `docs/compound/2026-10-20-backlogit-cli-pitfalls-and-permission-outcome-lessons.md`: Documents eliminating TOCTOU windows between transport/handler and hub by performing state validation atomically inside the authoritative hub step.
* `docs/compound/2026-09-30-go-msys-bash-bridging-and-unicode-case-mapping-gap.md`: Documents that Go stdlib `unicode` category predicates (`unicode.IsControl`, `unicode.IsSpace`, `unicode.Cf`) do not cover visually blank Letter/Symbol code points (`U+2800`, `U+3164`, `U+115F`, `U+1160`, `U+FFA0`, `U+034F`, `U+17B4`, `U+17B5`), requiring explicit rune enumeration and table-driven tests.
* `docs/compound/workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md`: Documents pairing wire-constant and config/transport contract changes with bidirectional conformance tests.

### Codebase Inspection Findings (Baseline `075-S` / `d5b209e`)

1. **`ACD578AE` (`internal/hub/authz.go`, `internal/hub/permission.go`)**:
   * In `internal/hub/permission.go:82`, `permRegistry.register` already records `previewClipped: ask.Request.PreviewClipped && ask.Request.Kind == agent.PermissionKindWrite` on `pendingPerm`.
   * However, `authzInput` (`internal/hub/authz.go:13-20`) has no `previewClipped bool` field, so `resolvePermission` (`internal/hub/permission.go:467-472`) allows a remote approver to approve a `PreviewClipped` write request whenever `remote_approve_destructive = true`. Because the hub cannot enforce that a remote client renders `preview_clipped` or requires a clipped-write confirmation, and because the truncated tail of the file diff/contents is not transmitted over the wire, approving a clipped write is safe only on the local TUI (which renders the non-spoofable column-0 `warning: write preview clipped — full content not shown` header and `APPROVE CLIPPED WRITE?` confirmation on the machine holding the workspace).

2. **`D46C1B38` (`internal/agent/copilot/permission.go`)**:
   * `internal/tui/sanitize.go` (`isHiddenRune`) already flags visually blank non-space non-`Cf` filler runes (`\u2800`, `\u3164`, `\u115f`, `\u1160`, `\uffa0`, `\u034f`, `\u17b4`, `\u17b5`, `unicode.Co`, `unicode.Variation_Selector`), but `internal/agent/copilot/permission.go` (`sanitizeBound`, `sanitizeIdent`, `sanitizeShellURL`) only checks `unicode.IsControl(r) || unicode.Is(unicode.Cf, r) || r == '\u2028' || r == '\u2029'`.
   * Consequently, remote WebSocket clients (which receive `PermissionRequestedPayload.Summary` without `tui.Sanitize`) receive raw invisible filler runes, and `indentURLsFraming` (`strings.TrimLeftFunc(l, unicode.IsSpace)`) can be bypassed by leading invisible filler runes, combining marks (`Mn`/`Me`), fullwidth ASCII (`ＵＲＬＳ：`), colon lookalikes (`：`, `︰`, `︓`, `∶`, `꞉`, `ː`), or Latin/Greek/Cyrillic/small-capital homoglyphs of `urls:`.

3. **`7BDC38FF` (`internal/tui/permission.go`)**:
   * `permissionBody` (`internal/tui/permission.go:180-191`) manually concatenates `m.wrapHeader("kind: ", p.Kind, inner)`, `m.wrapHeader("tool call: ", p.ToolCallID, inner)`, and conditionally `m.wrapHeader("warning: ", clippedWriteWarning, inner)`. Introducing a shared `modalHeader` / `renderModalHeaders` helper that sanitizes each header value as a single line (`SanitizeLine`) and wraps/indents all continuation lines (`i >= 1`) by two spaces (`"  "`) establishes a single structural invariant across every modal header.

4. **`026CF750` & `7EF4DA98` (`internal/hub`, `internal/transport`, `internal/transport/wsclient`, `.golangci.yml`)**:
   * **Already implemented in `060-S` / `062-S`**:
     * `NewRemote` composed constructor (`internal/transport/remote.go`) wiring `config.RemoteConfig` into `CredentialStore`, `Handler`, and `Server` (`OnBind: handler.SetBoundAddr`, `Drain: handler.Drain`).
     * `Remote.RevokeCredentials` and `Remote.CloseListener` implementing `hub.Gate` via `Server.CloseListener`.
     * Per-action `Handler.revalidate` (`credential_expired`, `generation_mismatch`, `credential_revoked`), `action.refused` wire shape (`v`, `type`, `op`, `reason`, `winner`, `request_id`), `MaxMessageBytes` (`16 KiB`), and `ConnectOptions` `Kind`/`Role` sourced strictly from `hub.ClientRemote` and `grant.Role()`.
   * **Verified remaining gaps to implement**:
     * **Atomic hub-step credential revalidation (`7EF4DA98`, `026CF750`)**: Currently `Handler.ServeHTTP` calls `Creds.Verify` before `Hub.Connect`, and `session.dispatch` calls `Handler.revalidate` on the read-pump goroutine before `Client.Do`. A credential revoked or expiring in the window between `Verify` and `completeConnect` (which may be deferred across hub iterations when hydration budget is full) or between `revalidate` and `processAction` on the hub goroutine is not re-checked inside the hub step. Adding `Validate func(gen string) Reason` on `hub.ConnectOptions` (stored in `Hub.validators[c.id]` and invoked inside `admission`/`completeConnect` and inside `processAction`) closes both TOCTOU windows atomically on the hub goroutine without coupling `internal/hub` to `internal/transport`.
     * **`local_only_action` refusals for credential reveal/re-issue and quit/shutdown (`026CF750`)**: ADR MVP-3 specifies `local_only_action` for remote attempts to reveal/re-issue credentials or quit/shut down the process. Recognizing `reveal_credentials`, `reissue_credentials`, `quit`, and `shutdown` in `decodeAction` / `session.dispatch` and returning `action.refused` with `reason: "local_only_action"` (while keeping other unknown `op` strings on `action_unknown`) closes this ADR MVP-3 matrix row.
     * **`Feedback` clearing on non-deny decisions and size cap (`026CF750`)**: `decodeAction` (`internal/transport/actions.go:105`) and `Hub.resolvePermission` (`internal/hub/permission.go:454`) currently retain `Decision.Feedback` even when `Decision.Kind == DecisionAllowOnce` and do not cap `Feedback` below the 16 KiB frame limit. Clearing `Feedback` whenever `Decision.Kind != DecisionDeny` and capping `Feedback` at `MaxDecisionFeedbackBytes = 1024` (rune-safe UTF-8 truncation) in both `decodeAction` and `Hub.resolvePermission` enforces the contract at both boundaries.
     * **`depguard` layering rule for `internal/transport` (`7EF4DA98`)**: `.golangci.yml` currently defines only `copilot-sdk-boundary`. Adding `transport-layering-boundary` and `wsclient-standalone-boundary` rules in `.golangci.yml` (plus a `go list`/AST layering test in `internal/transport/conformance_test.go`) enforces that `internal/transport` never imports `internal/app`, `internal/tui`, or `internal/agent/copilot`, `internal/transport/wsclient` never imports `internal/transport`, and lower/sibling packages (`internal/hub`, `internal/agent`, `internal/tui`, `internal/config`, `internal/pathsafe`) never import `internal/transport`.
     * **`config<->transport` conformance test & `wsclient` vs server wire-constant drift test (`7EF4DA98`)**: Dedicated tests in `internal/transport/conformance_test.go` locking `config.Validate()` and `transport.NewHandler`/`NewRemote` agreement over `RemoteConfig` and locking wire constants between `wsclient`, `transport`, and `hub`.

5. **`FE0D5ECD`, `86BE6345`, `21F9D21A`, & `6617141A` (`internal/transport`, `internal/transport/wsclient`, `internal/app`, `docs/`)**:
   * **Pre-registration WebSocket header validation & counting `websocket.Accept` refusals (`FE0D5ECD`, `21F9D21A`)**: In `Handler.upgrade` (`internal/transport/handler.go`), wrapping `w` in an `acceptResponseRecorder` (preserving `http.Hijacker` and `http.Flusher`) and checking `validWebSocketHandshake(r)` (`Sec-WebSocket-Version == "13"` and valid 16-byte base64 `Sec-WebSocket-Key`) before `Hub.Connect` ensures that malformed or unsupported-version handshakes skip `Hub.Connect` (zero hub registration or hydration churn) and go straight to `websocket.Accept(rec, ar, &opts)`, which writes the single RFC 6455 HTTP error response (`426` or `400`), after which `h.recordReject(rec.status, rejectHandshake)` (`"invalid_handshake"`) increments `RejectedUpgrades()` and invokes `OnReject` without writing a second response.
   * **Counting over-cap `503`s and wrong-path `404`s, off-accept-loop `503` writes, and excluding benign draining `503`s (`FE0D5ECD`, `86BE6345`)**:
     * In `limitListener.Accept` (`internal/transport/limiter.go`), when `l.active.Add(1) > l.max`, immediately decrement `l.active.Add(-1)`, invoke `l.onReject()` synchronously on the accept loop (wired via `ServerConfig.OnReject: handler.RecordReject` to `Handler.RecordReject(http.StatusServiceUnavailable, rejectOverCap)` where `rejectOverCap = "connection_cap_exceeded"`), and enqueue `c` non-blockingly onto a bounded 64-slot `refuseQ` channel drained in FIFO order off the accept goroutine by a background refusal worker that writes `refusal503` (with `refuseWriteTimeout = 50ms`) and closes `c` (closing `c` immediately on the accept loop if `refuseQ` is ever full, so the accept loop never performs a socket write or blocks).
     * In `Handler.ServeHTTP`, route `r.URL.Path != WSPath` through `h.reject(w, http.StatusNotFound, rejectNotFound)` (`"wrong_path"`).
     * In `Handler.RecordReject`, exclude `reason == rejectDraining` (`"draining"`) from incrementing `h.rejected` and calling `OnReject` (logging at `Debug` instead of `Warn`), so normal shutdown draining does not trigger false-positive security notices on the TUI status line.
   * **`wsclient` URL fragment and opaque rejection (`FE0D5ECD`)**: Harden `checkURL` in `internal/transport/wsclient/client.go` to reject `u.Opaque != ""`, `u.Fragment != ""`, `u.RawFragment != ""`, or `#` in the raw URL string with `ErrInvalidURL`, while retaining `localhost` alongside loopback IP literals in `checkTransport` as ratified by `wsclient` docs and `TestDialRefusesPlainWSUnlessLoopback`.
   * **ADR MVP-2 `Host` default-port non-equivalence ratification & deferred items (`FE0D5ECD`)**: Amend ADR MVP-2 (`docs/decisions/2026-10-19-intercom-go-mvp-adr-2-credentials-and-origin.md`) §MVP-D11 in place to match the 060-S addendum, `docs/config-reference.md`, and `hostKey` (`Host` carries no scheme, so `a` and `a:443` / `a:80` are distinct keys). Keep per-IP/per-credential upgrade rate limiting and reserved authenticated slots deferred per ADR MVP-2 §MVP-D12 (all tunnel traffic arrives from loopback over one relay).
   * **`pumpRejects` interval rate-limiting (`6617141A`)**: In `internal/app/assemble.go`, rate-limit `pumpRejects` to at most one `ui.prog.Send` per `defaultRejectNoticeInterval = 250 * time.Millisecond` (delivering the first notice immediately, then disabling `a.rejectCh` reads until a 250ms timer fires while `a.onReject` coalesces the highest-`Total` notice in `a.rejectCh`).

## Options Evaluated

### Option A: Two Chained Single-Feature Shipments (`076-S -> 077-S`) with Fail-Closed Clipped-Write Remote Refusal and Atomic Hub-Step Revalidation (Recommended)

* **Description**:
  * **Shipment `076-S` (Covering Feature `083-F`)**: Refuse remote approval of `PreviewClipped` write permissions (`ACD578AE`) and enforce `Feedback` clearing/bounding in `internal/hub`; harden `sanitizeBound`, `sanitizeIdent`, `sanitizeShellURL`, and `indentURLsFraming` against visually blank non-space non-`Cf` runes and `urls:` homoglyph/fullwidth lookalikes (`D46C1B38`); introduce shared `modalHeader` / `renderModalHeaders` continuation indentation and column-0 header invariant in `internal/tui` (`7BDC38FF`); and update ADR MVP-3, design doc §5.2/§6.1, and `docs/config-reference.md`.
  * **Shipment `077-S` (Covering Feature `084-F`, blocked by `076-S`)**: Add `ConnectOptions.Validate` atomic hub-step credential revalidation at `completeConnect` and `processAction` (`7EF4DA98`, `026CF750`); enforce `Feedback` clearing/bounding and `local_only_action` refusals in `internal/transport` (`026CF750`); add off-accept-loop `503` writes and count over-cap `503`s (`connection_cap_exceeded`), wrong-path `404`s (`wrong_path`), and library `websocket.Accept` refusals (`invalid_handshake`, skipping `Hub.Connect` on invalid handshake headers) while excluding benign draining `503`s (`FE0D5ECD`, `86BE6345`, `21F9D21A`); reject URL fragments and opaque URLs in `wsclient` and add `config<->transport`, `wsclient` wire-constant drift, and package-layering conformance tests (`7EF4DA98`, `FE0D5ECD`); add `transport-layering-boundary` and `wsclient-standalone-boundary` depguard rules in `.golangci.yml` (`7EF4DA98`); rate-limit `pumpRejects` to at most one `prog.Send` per 250ms interval in `internal/app` (`6617141A`); and update ADR MVP-2 and the operator runbook.
* **Pros**:
  * Satisfies Ship Step 6.1(a1) `S2` (`n == 1` covering feature per shipment) and keeps every task within `<= 3` files and a single skill domain.
  * Fails closed on remote clipped-write approvals and closes both connect-time and action-time credential TOCTOU windows on the hub goroutine without violating package layering.
* **Cons**:
  * Requires two sequential PRs (`076-S` then `077-S`).
* **Effort estimate**: Medium (4 tasks in `076-S`, 8 tasks in `077-S`).
* **Fit**: Best alignment with security invariants, ADR MVP-2/MVP-3, and `S2`.

### Option B: Keep Remote Clipped-Write Approval Permitted under `remote_approve_destructive = true` and Rely on Pre-Enqueue Revalidation Only

* **Description**: Leave `authzInput` unchanged for `PreviewClipped` writes (documenting that remote clients must inspect `preview_clipped`), and keep credential revalidation on the transport read-pump goroutine before `Client.Do` rather than inside the hub step.
* **Pros**: Smaller diff in `internal/hub`.
* **Cons**: Leaves remote approvers able to approve clipped writes whose truncated tail never crossed the wire, and leaves a TOCTOU window between `Creds.Verify`/`Handler.revalidate` and the hub goroutine's `completeConnect`/`processAction` steps.
* **Effort estimate**: Low.
* **Fit**: Poor — fails to close the security and TOCTOU gaps identified by `ACD578AE` and `7EF4DA98`.

## Trade-off Comparison

| Criterion | Option A (Two Single-Feature Shipments, Fail-Closed Clipped Write & Atomic Hub Revalidation) | Option B (Status Quo Remote Clipped-Write Approval & Pre-Enqueue Revalidation Only) |
|---|---|---|
| Remote clipped-write safety (`ACD578AE`) | Enforced fail-closed in `hub.authorize` (`destructive_remote_refused` remotely even with opt-in) | Relies on unverified remote client UI warnings |
| Connect & action credential TOCTOU (`7EF4DA98`) | Closed atomically inside `Hub.completeConnect` and `Hub.processAction` | Leaves race window across channel queues |
| Pre-auth / DoS & status-line visibility (`FE0D5ECD`, `86BE6345`, `21F9D21A`, `6617141A`) | Full accounting (`over-cap 503`, `404`, `Accept`), benign `draining` excluded, 250ms TUI rate-limit | Partial or unthrottled |
| Ship Step 6.1(a1) `S2` compliance | `n == 1` covering feature per shipment (`083-F` in `076-S`, `084-F` in `077-S`) | N/A |

## Decision

Adopt **Option A**. Proceed immediately to implementation planning (`impl-plan`), plan hardening (`plan-harden`), multi-agent plan review (`plan-review`), harvest (`083-F` and `084-F`), shipment assembly (`076-S -> 077-S`), and stash archival under `DARK_MODE_ACTIVE` (P-017).

## Rejected Alternatives

* **Option B** was rejected because remote clients never receive the clipped tail of a truncated write preview and cannot be verified by the hub to enforce a clipped-write warning, and because pre-enqueue credential checks on the read-pump goroutine leave a TOCTOU window before the hub goroutine executes `completeConnect` or `processAction`.
* **Changing `wsclient` `checkTransport` to reject `ws://localhost`** (`FE0D5ECD` sub-item) was rejected because `wsclient` package docs and `TestDialRefusesPlainWSUnlessLoopback` (`internal/transport/wsclient/dial_test.go:25-26`) explicitly ratify name-based `localhost` acceptance alongside loopback IP literals (`127.0.0.0/8`, `::1`).
* **Per-IP or per-credential pre-auth upgrade rate limits and reserved authenticated slots** (`FE0D5ECD` sub-item) remain deferred per ADR MVP-2 §MVP-D12 because all Dev Tunnel connections arrive from `127.0.0.1` over a single local relay and pre-auth sockets have not yet presented a credential.

## Unresolved Questions

None. All nine stash entries have deterministic, test-verifiable designs grounded in the current codebase.

## Risks and Mitigations

* **Risk — `TestAuthorizeInvariantsOverTheWholeInputSpace` and `TestPolicyNeverReadsClientSuppliedFields` in `internal/hub/authz_test.go`**:
  * *Mitigation*: Keep `OperatorAction` struct fields unchanged (`Type,RequestID,Decision,Prompt`), add `previewClipped` to `authzInput` and update `TestAuthorizeInvariantsOverTheWholeInputSpace` to iterate over `previewClipped` (`false`, `true`), and handle `local_only_action` wire op names (`reveal_credentials`, `reissue_credentials`, `quit`, `shutdown`) via a transport-owned `decodedAction` struct (`internal/transport/actions.go` and `session.go`) without overloading `OperatorAction.Prompt`.
* **Risk — `TestOverCapRefusalDoesNotStallAcceptLoop` FIFO flood assertion in `internal/transport/review1_test.go`**:
  * *Mitigation*: Keep `refuseWriteTimeout = 50 * time.Millisecond`, perform `l.active.Add(-1)` and `l.onReject()` synchronously on the accept loop in `limitListener.Accept`, and enqueue refused sockets non-blockingly onto a bounded 64-slot `refuseQ` channel drained in FIFO order off the accept goroutine (closing immediately on the accept loop only if `refuseQ` is saturated), ensuring zero accept-loop socket writes and preserving FIFO `503` delivery across the 41-socket flood test.
* **Risk — Existing TUI golden snapshots (`internal/tui/testdata/modal.golden`)**:
  * *Mitigation*: `renderModalHeaders` produces byte-for-byte identical output to the existing `wrapHeader` calls for single-line and wrapped `kind:`, `tool call:`, and `warning:` headers.

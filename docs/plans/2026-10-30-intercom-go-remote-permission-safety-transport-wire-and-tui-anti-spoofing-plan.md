---
title: "Remote Permission Safety, Transport Wire Integrity, and TUI Anti-Spoofing Implementation Plan"
description: "Implementation plan for Capability Set 1 across two chained single-feature shipments (076-S..077-S): remote clipped-write approval refusal, shell/write summary invisible-rune and homoglyph anti-spoofing, and shared TUI modal header continuation helper (083-F); and transport wire/authz hardening, atomic hub-step credential revalidation, rejected-upgrade accounting, and pumpRejects rate-limiting (084-F)."
source: "docs/decisions/2026-10-30-intercom-go-remote-permission-safety-transport-wire-and-tui-anti-spoofing-deliberation.md"
status: "ready"
date: 2026-10-30
tags:
  - "permissions"
  - "authz"
  - "transport"
  - "wsclient"
  - "tui"
  - "security"
---

## Problem Frame

With `074-S..075-S` merged and closed on `main` (`075-S` merge commit `a2898942a12c70fe23be050d698ae4622153d97f`, closure commit `d5b209e80ee39baa520a85175ebda1f00a6e3f2f`), Capability Set 1 (**Remote Permission Safety, Transport Wire Integrity & TUI Anti-Spoofing**) resolves nine active stash entries (`ACD578AE`, `D46C1B38`, `7BDC38FF`, `026CF750`, `7EF4DA98`, `FE0D5ECD`, `86BE6345`, `21F9D21A`, and `6617141A`) spanning `internal/hub`, `internal/agent/copilot`, `internal/tui`, `internal/transport`, `internal/transport/wsclient`, `internal/app`, `.golangci.yml`, and `docs/`.

To satisfy Ship Step 6.1(a1) `S2` (`n == 1` covering feature per shipment) and maintain strict architectural layering, the work is organized into two sequential single-feature shipments (`076-S -> 077-S`). Shipment `076-S` (`083-F`) is declared a DAG root (`labels: ["dag-root"]`, `predecessor_source: "declared_root"`) under the operator's explicit P-001 policy exception over `062-S`'s open operator-present live runtime conditions, with `075-S` (`a2898942a12c70fe23be050d698ae4622153d97f` / `d5b209e80ee39baa520a85175ebda1f00a6e3f2f`) documented as the code baseline. Shipment `077-S` (`084-F`) is blocked by `076-S` (`backlogit dep add 077-S 076-S --type blocks`).

## Requirements Trace

| Source Requirement | Stash Origin | Target Shipment & Feature | Implementation Unit(s) |
|---|---|---|---|
| Refuse remote approval of `PreviewClipped` write permissions (`authzInput.previewClipped = true` -> `ReasonDestructiveRemoteRefused` for remote approvers even when `remote_approve_destructive = true`, while keeping local-TUI approval and remote `deny` allowed); clear `Decision.Feedback` unless `Decision.Kind == DecisionDeny` and cap `Decision.Feedback` at `MaxDecisionFeedbackBytes = 1024` (rune-safe UTF-8) in `Hub.resolvePermission` | `ACD578AE`, `026CF750` | `076-S` / `083-F` | Unit 1.1 (`083.001-T`) |
| Replace visually blank non-space non-`Cf` filler runes (`U+2800`, `U+3164`, `U+115F`, `U+1160`, `U+FFA0`, `U+034F`, `U+17B4`, `U+17B5`, `unicode.Co`, `unicode.Variation_Selector`) with `U+FFFD` in `sanitizeBound`, `sanitizeIdent`, and `sanitizeShellURL`; harden `indentURLsFraming` against leading invisible/blank fillers, combining marks (`Mn`/`Me`), fullwidth ASCII (`ＵＲＬＳ：`), colon lookalikes, and Latin/Greek/Cyrillic/small-capital homoglyphs of `urls:` | `D46C1B38` | `076-S` / `083-F` | Unit 1.2 (`083.002-T`) |
| Introduce shared `modalHeader` and `(Model).renderModalHeaders` continuation-indentation helper across every TUI modal header (`kind:`, `tool call:`, and conditional `warning:`) in `internal/tui/permission.go`, enforcing `SanitizeLine` and two-space (`"  "`) indentation on all continuation lines (`i >= 1`) so only registered modal header prefixes can start at column 0 | `7BDC38FF` | `076-S` / `083-F` | Unit 1.3 (`083.003-T`) |
| Update ADR MVP-3 (`docs/decisions/2026-10-19-intercom-go-mvp-adr-3-authorization-roles.md`), `docs/design-docs/intercom-go-backend-architecture.md` (§5.2, §6.1), and `docs/config-reference.md` to document local-TUI-only approval of clipped write previews (`preview_clipped: true`) and invisible-rune/homoglyph `urls:` anti-spoofing | `ACD578AE`, `D46C1B38` | `076-S` / `083-F` | Unit 1.4 (`083.004-T`) |
| Add `ReasonCredentialExpired`, `ReasonGenerationMismatch`, `ReasonLocalOnlyAction`, `ErrCredentialRevoked`, and `ConnectOptions.Validate func(gen string) Reason` in `internal/hub`, and invoke `Validate(h.gen)` atomically inside `Hub.admission`/`completeConnect` and `Hub.processAction` (closing the connect-time and action-time credential TOCTOU windows) | `7EF4DA98`, `026CF750` | `077-S` / `084-F` | Unit 2.1 (`084.001-T`) |
| Enforce `Feedback` clearing on non-deny decisions and rune-safe capping at `MaxFeedbackBytes = 1024` in `decodeAction`; return `action.refused` with `reason: "local_only_action"` for remote `reveal_credentials`, `reissue_credentials`, `quit`, and `shutdown` operations via a transport-owned `decodedAction` wrapper in `decodeWireAction`/`session.dispatch` (without overloading `OperatorAction.Prompt` and after `v.reason != ""` shutdown-revocation checks); and close the connection after flushing the refusal when `Client.Do` returns a hub-step credential revalidation failure (`credential_revoked`, `credential_expired`, `generation_mismatch`, or role mismatch) | `026CF750`, `7EF4DA98` | `077-S` / `084-F` | Unit 2.2 (`084.002-T`) |
| Enqueue over-cap sockets non-blockingly onto a bounded 64-slot `refuseQ` channel drained in FIFO order off the accept goroutine in `limitListener.Accept` (`internal/transport/limiter.go`) and invoke an `OnReject` callback wired through `ServerConfig` / `NewServer` (`internal/transport/listen.go`) without ever writing or blocking on the accept loop | `FE0D5ECD`, `86BE6345` | `077-S` / `084-F` | Unit 2.3 (`084.003-T`) |
| Pass `Validate: h.grantValidator(grant)` in `Handler.upgrade`; validate single-header RFC 6455 `Sec-WebSocket-Version == "13"` and 16-byte base64 `Sec-WebSocket-Key` before `Hub.Connect`; wire `ServerConfig.OnReject: handler.RecordReject` in `NewRemote`; count over-cap `503` (`connection_cap_exceeded`), wrong-path `404` (`wrong_path`), and library-generated `websocket.Accept` handshake refusals (`invalid_handshake`, via `acceptResponseRecorder` without writing a second response) in `RejectedUpgrades`/`OnReject`; and exclude benign shutdown `draining` `503`s from `RejectedUpgrades`/`OnReject` | `7EF4DA98`, `FE0D5ECD`, `86BE6345`, `21F9D21A` | `077-S` / `084-F` | Unit 2.4 (`084.004-T`) |
| Add `transport-layering-boundary`, `wsclient-standalone-boundary`, and `lower-layers-no-transport` depguard rules in `.golangci.yml` | `7EF4DA98` | `077-S` / `084-F` | Unit 2.5 (`084.005-T`) |
| Reject URLs carrying a `#fragment` or `Opaque` part in `wsclient.Dial` (`checkURL` returning `ErrInvalidURL`), and add `config<->transport` conformance, `wsclient` vs server wire-constant drift, and `internal/transport` package-layering + `.golangci.yml` depguard tests in `internal/transport/conformance_test.go` | `7EF4DA98`, `FE0D5ECD` | `077-S` / `084-F` | Unit 2.6 (`084.006-T`) |
| Rate-limit `pumpRejects` forwarding of rejected-upgrade notices in `internal/app/assemble.go` to at most one `prog.Send` per interval (`defaultRejectNoticeInterval = 250 * time.Millisecond`, immediate on first notice after idle, coalescing subsequent notices in `a.rejectCh` during cooldown) | `6617141A` | `077-S` / `084-F` | Unit 2.7 (`084.007-T`) |
| Ratify `Host` default-port non-equivalence in ADR MVP-2 (`docs/decisions/2026-10-19-intercom-go-mvp-adr-2-credentials-and-origin.md` §MVP-D11) and update `docs/runbooks/intercom-operator-runbook.md` for rejected-upgrade accounting (`connection_cap_exceeded`, `wrong_path`, `invalid_handshake`, `draining` exclusion, and 250ms rate-limiting) and remote clipped-write refusal | `FE0D5ECD`, `86BE6345`, `21F9D21A`, `6617141A`, `ACD578AE` | `077-S` / `084-F` | Unit 2.8 (`084.008-T`) |

## Constitution Check

| Principle | Compliance | Notes |
|---|---|---|
| I. Safety-First Go | PASS | All production code uses Go 1.25.0 (`GOTOOLCHAIN=go1.26.10`), explicit `apperr` / sentinel error wrapping, bounded waits, and zero `panic` calls in library packages. |
| II. Test-First Development (NON-NEGOTIABLE) | PASS | Every `code` implementation unit specifies red-phase failing test assertions in its colocated `_test.go` file before production code changes and verifies green via `go test -race ./...`. |
| III. Workspace Isolation and Security Boundaries | PASS | Zero new filesystem write calls or `pathsafe.Root.Resolve` call sites (`go run ./tools/gatecheck write-path --self-test --root .` stays 100% green). Remote clipped-write approvals fail closed (`destructive_remote_refused`), and all permission summaries, identifiers, and feedback strings are sanitized and rune-bounded. |
| IV. CLI Workspace Containment (NON-NEGOTIABLE) | PASS | All modifications are confined to the repository working tree. |
| V. Structured Observability | PASS | Pre-upgrade rejections log fixed reason codes (`connection_cap_exceeded`, `wrong_path`, `invalid_handshake`, `draining` at `Debug`) with zero secret or header leakage. |
| VI. Single Responsibility | PASS | Zero new external dependencies in `go.mod`; `depguard` rules in `.golangci.yml` enforce `internal/transport` and `internal/transport/wsclient` package boundaries. |
| VII. Destructive Command Approval (NON-NEGOTIABLE) | PASS | No destructive commands are used; consumed stash entries are archived via non-destructive `backlogit stash archive`. |
| VIII. Explicit Safety Modes | PASS | Freeze-scope mode active on the 9 chartered stash entries of Capability Set 1. |
| IX. Git-Friendly Persistence | PASS | Deliberation, plan, session memory, and backlog items use Markdown with YAML frontmatter. |
| X. Agent Context Efficiency | PASS | Targeted code and backlog queries used throughout staging. |
| XI. Merge Commit History Preservation | PASS | All shipments merge via `--merge` commits (P-009). |

## Implementation Units

### Sub-Epic 1: Shipment `076-S` — Covering Feature `083-F` (Remote clipped-write approval refusal, shell/write summary invisible-rune and homoglyph anti-spoofing, and shared TUI modal header continuation helper)

#### Unit 1.1 (`083.001-T`): Refuse remote approval of `PreviewClipped` write permissions and enforce `Feedback` clearing/bounding in `internal/hub` (`ACD578AE`, `026CF750`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/hub/authz.go`
  * `internal/hub/permission.go`
  * `internal/hub/authz_test.go`
* **Changes needed**:
  * In `internal/hub/authz.go`:
    * Add `previewClipped bool` to `authzInput` (documenting that a write permission whose preview was clipped is local-TUI-only for approval, never unlocked by `remoteApproveDestructive`).
    * In `authorize(in authzInput)`: inside `if in.op == OpResolvePermission && in.decision != agent.DecisionDeny`, check `if unknownPermissionKind(in.kind) || in.previewClipped { return false, ReasonDestructiveRemoteRefused }` before `if knownDestructive(in.kind) && !in.remoteApproveDestructive`. Local TUI approval (`in.p.surface == surfaceTUI`) and remote denial (`in.decision == agent.DecisionDeny`) remain allowed.
  * In `internal/hub/permission.go`:
    * Export `const MaxDecisionFeedbackBytes = 1024` (the maximum byte length of `PermissionDecision.Feedback` on a `DecisionDeny`, truncated on a UTF-8 rune boundary).
    * In `(*Hub).resolvePermission(r *actionReq)`:
      * Pass `previewClipped: e.previewClipped` in the `authzInput` struct passed to `authorize(...)`.
      * Before `h.perm.cas(r.act.RequestID, r.from.label(), d, "")`, normalize `d := r.act.Decision`: if `d.Kind != agent.DecisionDeny`, set `d.Feedback = ""`; else if `len(d.Feedback) > MaxDecisionFeedbackBytes`, set `d.Feedback = truncUTF8(d.Feedback, MaxDecisionFeedbackBytes)`.
  * In `internal/hub/authz_test.go`:
    * Add matrix rows in `TestAuthorizeMatrixRows` for `previewClipped: true` with `remoteApproveDestructive: false` and `true` (remote approver allow -> `false, ReasonDestructiveRemoteRefused`; remote approver deny -> `true, ""`; local TUI allow -> `true, ""`; remote observer allow/deny -> `false, ReasonRoleNotPermitted`).
    * Update `TestAuthorizeInvariantsOverTheWholeInputSpace` to also iterate over `clipped := []bool{false, true}` and assert that `op == OpResolvePermission && dec != agent.DecisionDeny && clipped` on `surfaceRemote` is always refused with `ReasonDestructiveRemoteRefused`.
    * Add `TestRemoteClippedWriteApprovalRefusedEvenWhenOptedInAndFeedbackBounded`: verify through a running `Hub` with `RemoteApproveDestructive: true` that a `PermissionKindWrite` ask with `PreviewClipped: true` is refused to a remote approver (`ReasonDestructiveRemoteRefused`, leaving the ask pending), can still be denied by a remote approver (with an oversized multi-byte UTF-8 `Feedback` capped at `<= MaxDecisionFeedbackBytes` valid UTF-8 bytes) or approved by the local TUI (with any non-deny `Feedback` cleared to `""`).
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/hub -run "TestAuthorize|TestRemoteClippedWrite|TestHubPackageHasNoBareChannelSends"`

#### Unit 1.2 (`083.002-T`): Harden `sanitizeBound`, `sanitizeIdent`, `sanitizeShellURL`, and `indentURLsFraming` against visually blank non-space non-`Cf` runes and `urls:` homoglyph/fullwidth lookalikes (`D46C1B38`)

* **Domain**: `code` (`size: S`, `complexity: medium`)
* **Posture**: `test-first`
* **Files affected** (2 files):
  * `internal/agent/copilot/permission.go`
  * `internal/agent/copilot/permission_cycle2_test.go`
* **Changes needed**:
  * In `internal/agent/copilot/permission.go`:
    * Add `func isInvisibleOrBlankRune(r rune) bool` returning true for `unicode.IsControl(r) || r == '\u2028' || r == '\u2029' || unicode.Is(unicode.Cf, r) || unicode.Is(unicode.Co, r) || unicode.Is(unicode.Variation_Selector, r) || r == '\u2800' || r == '\u3164' || r == '\u115f' || r == '\u1160' || r == '\uffa0' || r == '\u034f' || r == '\u17b4' || r == '\u17b5'` (aligning adapter-level sanitization with `internal/tui/sanitize.go` `isHiddenRune` so remote WebSocket clients that render raw `Summary` never receive visually blank filler runes).
    * Update `sanitizeBound`, `sanitizeIdent`, and `sanitizeShellURL` to replace `isInvisibleOrBlankRune(r)` with `utf8.RuneError` (`U+FFFD`) (while `sanitizeBound` continues to preserve `\n` and `\t`).
    * Add `func foldURLsLookalikeRune(r rune) byte` and `func hasURLsLookalikePrefix(line string) bool`:
      * Strip leading runes where `unicode.IsSpace(r) || isInvisibleOrBlankRune(r) || r == utf8.RuneError || unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r)`.
      * Scan the remaining runes, skipping combining marks (`unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r)`), folding fullwidth ASCII (`r >= 0xff01 && r <= 0xff5e` -> `r - 0xfee0`), ASCII letters (lowercased), colon lookalikes (`:`, `\uff1a` `：`, `\ufe55` `︰`, `\ufe13` `︓`, `\u2236` `∶`, `\ua789` `꞉`, `\u02d0` `ː`, `\u0589` `։`), and Latin/Greek/Cyrillic/small-capital confusable homoglyphs of `u` (`\u03c5` `υ`, `\u0443` `у`, `\u0423` `У`, `\u1d1c` `ᴜ`, `\u057d` `ս`), `r` (`\u0433` `г`, `\u0280` `ʀ`, `\u0440` `р`, `\u03c1` `ρ`), `l` (`1`, `|`, `I`, `\u0131` `ı`, `\u04cf` `ӏ`, `\u04c0` `Ӏ`, `\u0456` `і`, `\u0406` `І`, `\u03b9` `ι`, `\u029f` `ʟ`, `\u2113` `ℓ`), and `s` (`5`, `\u0455` `ѕ`, `\u0405` `Ѕ`, `\ua731` `ꜱ`, `\u03c2` `ς`, `\u03c3` `σ`) into a 5-byte prefix buffer and checking whether the first 5 folded characters equal `"urls:"`.
    * Update `indentURLsFraming(text string)` to prefix `"  "` whenever `hasURLsLookalikePrefix(l)` is true.
  * In `internal/agent/copilot/permission_cycle2_test.go`:
    * Add `TestSanitizeBoundAndShellSummaryRejectInvisibleFillersAndURLsHomoglyphs`:
      1. Verify `sanitizeBound`, `sanitizeIdent`, `sanitizeShellURL`, `writeSummary`, and `shellSummary` replace visually blank non-space non-`Cf` runes (`U+2800`, `U+3164`, `U+115F`, `U+1160`, `U+FFA0`, `U+034F`, `U+17B4`, `U+17B5`, variation selector `U+FE0F`, private-use `U+E000`) with `U+FFFD`.
      2. Verify `shellSummary` / `indentURLsFraming` indents command lines starting with visually blank fillers + `urls:`, fullwidth `ＵＲＬＳ：` / `ｕｒｌｓ：`, Cyrillic/Greek/small-capital homoglyphs (`υгӏѕ:`, `уrls∶`, `ᴜʀʟꜱ：`), and combining-mark-decorated `u`+`\u0301`+`rls:` by two spaces (`"  "`), while leaving the adapter's own trailing `\nurls: ...` line unindented at column 0.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/agent/copilot -run "TestSanitizeBoundAndShellSummaryRejectInvisibleFillersAndURLsHomoglyphs|TestSummary|TestShellSummary"`

#### Unit 1.3 (`083.003-T`): Introduce shared `modalHeader` / `renderModalHeaders` continuation-indentation helper and column-0 header invariant in `internal/tui` (`7BDC38FF`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (2 files):
  * `internal/tui/permission.go`
  * `internal/tui/permission_test.go`
* **Changes needed**:
  * In `internal/tui/permission.go`:
    * Define `type modalHeader struct { prefix, value string }` and `func (m Model) renderModalHeaders(inner int, headers ...modalHeader) string`: for each `h` in `headers`, calls `m.wrapHeader(h.prefix, SanitizeLine(h.value), inner)` and joins the rendered headers with `"\n"`.
    * In `wrapHeader(prefix, val string, inner int)`: sanitize `val = SanitizeLine(val)` as defense in depth before wrapping `prefix + val` to `max(inner-2, 1)` columns and indenting every continuation line (`i >= 1`) with `"  "`.
    * Refactor `permissionBody(p permission)` to build `headers := []modalHeader{{"kind: ", p.Kind}, {"tool call: ", p.ToolCallID}}` (appending `modalHeader{"warning: ", clippedWriteWarning}` when `p.PreviewClipped` is true) and call `m.renderModalHeaders(inner, headers...)`.
  * In `internal/tui/permission_test.go`:
    * Add `TestModalHeadersSharedHelperAndColumnZeroInvariant`: test `renderModalHeaders` and `permissionBody` across narrow widths (`inner = 10..24`), oversized `Kind` and `ToolCallID` values containing embedded newlines and header-like tokens (`"\nkind: read\nwarning: fake"`), and `PreviewClipped: true`, asserting that EVERY non-empty line in `permissionBody` that does not begin with `"  "` starts with one of the registered modal header prefixes (`"kind: "`, `"tool call: "`, `"warning: "`).
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/tui -run "TestModal|TestPackageHasNoDirectAgentOrSDKReach"`

#### Unit 1.4 (`083.004-T`): Update ADR MVP-3, backend architecture design doc, and config reference for remote clipped-write refusal and summary anti-spoofing (`ACD578AE`, `D46C1B38`)

* **Domain**: `docs` (`size: S`, `complexity: low`)
* **Posture**: `migration-first`
* **Files affected** (3 files):
  * `docs/decisions/2026-10-19-intercom-go-mvp-adr-3-authorization-roles.md`
  * `docs/design-docs/intercom-go-backend-architecture.md`
  * `docs/config-reference.md`
* **Changes needed**:
  * In `docs/decisions/2026-10-19-intercom-go-mvp-adr-3-authorization-roles.md`:
    * Update the Action-by-Role Matrix row for `Approve a modeled destructive permission (shell, write, url)` and add a dedicated row for `Approve a clipped write permission (preview_clipped: true)`: Local TUI = `Y`, Remote approver = `N (never, even with remote_approve_destructive)`, Remote observer = `N`, Refusal reason = `destructive_remote_refused` (approver) / `role_not_permitted` (observer).
    * Replace the `074-S` accepted-exposure bullet on remote approval of clipped writes with a `076-S` amendment recording that clipped write previews (`preview_clipped: true`) are local-TUI-only for approval (`destructive_remote_refused` remotely even when `remote_approve_destructive = true`; remote `deny` remains allowed) and that shell/write summaries sanitize visually blank filler runes and indent homoglyph/fullwidth `urls:` lookalikes.
  * In `docs/design-docs/intercom-go-backend-architecture.md` (§5.2 and §6.1):
    * Document that `preview_clipped: true` write requests are local-TUI-only for approval (refused to remote approvers with `destructive_remote_refused` even when `remote_approve_destructive = true`), and that adapter summary sanitization replaces visually blank non-space non-`Cf` filler runes (`U+2800`, `U+3164`, `U+115F`, `U+1160`, `U+FFA0`, `U+034F`, `U+17B4`, `U+17B5`, `Co`, variation selectors) with `U+FFFD` and indents fullwidth/homoglyph lookalikes of `urls:`.
  * In `docs/config-reference.md`:
    * Update the `remote.remote_approve_destructive` description to state that clipped write previews (`preview_clipped = true`) and unknown/unmodeled kinds always require the local TUI for approval even when `remote_approve_destructive = true`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/config -run "TestConfigReference"`

---

### Sub-Epic 2: Shipment `077-S` — Covering Feature `084-F` (Transport wire/authz hardening, atomic hub-step credential revalidation, rejected-upgrade accounting, and `pumpRejects` rate-limiting)

#### Unit 2.1 (`084.001-T`): Add `ConnectOptions.Validate` atomic hub-step credential revalidation at `completeConnect` and `processAction` in `internal/hub` (`7EF4DA98`, `026CF750`)

* **Domain**: `code` (`size: S`, `complexity: medium`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/hub/hub.go`
  * `internal/hub/permission.go`
  * `internal/hub/action_closed_test.go`
* **Changes needed**:
  * In `internal/hub/hub.go`:
    * Export sentinel `var ErrCredentialRevoked = errors.New("hub: credential revoked")` and stable refusal reasons `const ( ReasonCredentialExpired Reason = "credential_expired"; ReasonGenerationMismatch Reason = "generation_mismatch"; ReasonLocalOnlyAction Reason = "local_only_action" )`.
    * Add `Validate func(gen string) Reason` to `ConnectOptions`: when non-nil, the hub calls `Validate(h.gen)` inside the hub step during `admission` / `completeConnect` and before every `processAction` for that client, returning `""` when valid or the refusal `Reason` (`ReasonCredentialRevoked`, `ReasonCredentialExpired`, `ReasonGenerationMismatch`, `ReasonRoleNotPermitted`, or `ReasonShuttingDown` when revoked by shutdown). Document the strict execution contract on `ConnectOptions.Validate`: it runs synchronously on the sole-writer hub goroutine, so it MUST be an O(1) in-memory check with no I/O, no blocking waits, and no re-entrant calls into `Hub` methods or channels (strict lock hierarchy: hub goroutine -> `CredentialStore.mu`, never `CredentialStore.mu` -> hub).
    * Add `validators map[uint64]func(gen string) Reason` on `Hub` (initialized in `New`, populated in `completeConnect` when `req.opts.Validate != nil`, and deleted via `delete(h.validators, id)` inside `(*Hub).unregister` so abandoned connects and shutdown clean up every validator closure).
    * In `(*Hub).admission(req *connectReq)`: when `req.opts.Validate != nil`, call `if reason := req.opts.Validate(h.gen); reason != ""` — if `reason == ReasonShuttingDown`, return `apperr.Wrap(apperr.KindPolicy, ErrShuttingDown)`; otherwise return `apperr.Wrap(apperr.KindUnauthorized, ErrCredentialRevoked)`.
  * In `internal/hub/permission.go`:
    * In `(*Hub).processAction(r *actionReq)`: after `if r.client != nil && r.client.isClosed()`, check `if r.client != nil { if v := h.validators[r.client.id]; v != nil { if reason := v(h.gen); reason != "" { if reason == ReasonShuttingDown { return ActionResult{Reason: ReasonCredentialRevoked} }; return ActionResult{Reason: reason} } } }`.
  * In `internal/hub/action_closed_test.go`:
    * Add `TestConnectAndActionRevalidateCredentialAtomicallyInHubStep`:
      1. Verify that when `ConnectOptions.Validate` transitions to returning `ReasonCredentialRevoked` or `ReasonShuttingDown` before `completeConnect` executes on the hub goroutine, `Hub.Connect` fails with `ErrCredentialRevoked` (`apperr.KindUnauthorized`) or `ErrShuttingDown` without registering a client or consuming hydration budget.
      2. Verify that when an action is queued via `Client.Do` and the client's `Validate` callback begins returning `ReasonCredentialExpired`, `ReasonGenerationMismatch`, `ReasonCredentialRevoked`, or `ReasonShuttingDown` before the hub goroutine runs `processAction`, the hub step refuses the action with the corresponding `Reason` (`ReasonCredentialRevoked` for `ReasonShuttingDown`) and never forwards it to the agent or resolves a pending permission.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/hub -run "TestActionQueued|TestConnectAndActionRevalidateCredentialAtomicallyInHubStep|TestHubPackageHasNoBareChannelSends"`

#### Unit 2.2 (`084.002-T`): Enforce `Feedback` clearing/bounding and `local_only_action` refusals via `decodedAction` in `internal/transport`, and close connection on hub-step credential failure (`026CF750`, `7EF4DA98`)

* **Domain**: `code` (`size: S`, `complexity: medium`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/transport/actions.go`
  * `internal/transport/session.go`
  * `internal/transport/revalidate_test.go`
* **Changes needed**:
  * In `internal/transport/actions.go`:
    * Define `const MaxFeedbackBytes = hub.MaxDecisionFeedbackBytes` (`1024`), `const ReasonLocalOnlyAction = string(hub.ReasonLocalOnlyAction)` (`"local_only_action"`), and wire op names `const ( opRevealCredentials = "reveal_credentials"; opReissueCredentials = "reissue_credentials"; opQuit = "quit"; opShutdown = "shutdown" )`.
    * Define a transport-internal `type decodedAction struct { act hub.OperatorAction; localOnlyOp string }` and `func decodeWireAction(data []byte) (decodedAction, error)` (with `decodeAction(data []byte) (hub.OperatorAction, error)` calling `decodeWireAction` and returning `da.act, err` so `OperatorAction.Prompt` is never overloaded with wire op names and existing `decodeAction` callers/tests stay unchanged), plus `func truncUTF8Bytes(s string, maxBytes int) string` truncating on a UTF-8 rune boundary.
    * In `decodeWireAction`:
      * In `case opResolvePermission:`: set `act.Decision.Feedback = truncUTF8Bytes(fields["feedback"], MaxFeedbackBytes)` ONLY when `act.Decision.Kind == agent.DecisionDeny` (leaving `act.Decision.Feedback == ""` when `act.Decision.Kind == agent.DecisionAllowOnce`).
      * Add `case opRevealCredentials, opReissueCredentials, opQuit, opShutdown:`: return `decodedAction{act: hub.OperatorAction{Type: hub.OpUnknown, RequestID: fields["request_id"]}, localOnlyOp: op}`, nil.
    * Add `refusalForWire(da decodedAction, reason string) []byte` which uses `da.localOnlyOp` as `Op` when non-empty and `opName(da.act.Type)` otherwise.
  * In `internal/transport/session.go`:
    * In `(*session).readPump`: call `da, err := decodeWireAction(data)` and pass `da` to `s.dispatchWire(ctx, da)` (keeping `s.dispatch(ctx, act)` as a wrapper calling `s.dispatchWire(ctx, decodedAction{act: act})`).
    * In `(*session).dispatchWire(ctx context.Context, da decodedAction) bool`:
      * Run `v := s.h.revalidate(s.grant, s.client.Role())`, `s.bucket.allow()`, and `if v.reason != ""` first (using `refusalForWire(da, ...)` so shutdown-revoked or rate-limited credentials still receive `credential_revoked` / `rate_limited` before any policy/op check, preserving ADR MVP-3 check order).
      * After `s.violations = 0`, if `da.localOnlyOp != ""`, call `s.queueReply(refusalForWire(da, ReasonLocalOnlyAction))` and return `true` (refusing the local-only action while keeping the connection open).
      * After `res, err := s.client.Do(actx, da.act)`: if `!res.Accepted` and `res.Reason == hub.ReasonCredentialRevoked || res.Reason == hub.ReasonCredentialExpired || res.Reason == hub.ReasonGenerationMismatch || res.Reason == hub.ReasonRoleNotPermitted`, check `if v := s.h.revalidate(s.grant, s.client.Role()); v.closeConn`: when true, call `s.queueReply(replyFor(da.act, res))`, `s.requestClose(websocket.StatusPolicyViolation, closeReasonEnded)`, wait `<-ctx.Done()`, and return `false`.
  * In `internal/transport/revalidate_test.go`:
    * Add `TestLocalOnlyActionsRefusedAndFeedbackBoundedAndClearedOnAllow`:
      1. Verify `reveal_credentials`, `reissue_credentials`, `quit`, and `shutdown` return `action.refused` with `reason == "local_only_action"`, echo `op` and `request_id`, keep the connection open for active approver and observer sessions, and return `credential_revoked` when sent during shutdown revocation.
      2. Verify `decodeAction` clears `Feedback` to `""` on `allow_once` even when `"feedback"` is supplied in JSON, and truncates `Feedback` on `deny` to `<= MaxFeedbackBytes` (`1024`) valid UTF-8 bytes.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/transport -run "TestRevalidate|TestLocalOnlyActionsRefusedAndFeedbackBoundedAndClearedOnAllow|TestDecodeAction"`

#### Unit 2.3 (`084.003-T`): Non-blocking off-accept-loop over-cap `503` refusal queue and `OnReject` hook in `limitListener` and `Server` (`FE0D5ECD`, `86BE6345`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/transport/limiter.go`
  * `internal/transport/listen.go`
  * `internal/transport/review1_test.go`
* **Changes needed**:
  * In `internal/transport/limiter.go`:
    * Define `const refuseQueueCap = 64` and add `onReject func()`, `refuseQ chan net.Conn`, and `workerOnce sync.Once` to `limitListener` (`newLimitListener(l net.Listener, limit int, onReject func()) *limitListener`).
    * In `(*limitListener).Accept()`: when `l.active.Add(1) > l.max`, immediately call `l.active.Add(-1)` and `if l.onReject != nil { l.onReject() }` on the accept loop, then enqueue `c` non-blockingly onto `l.refuseQ` (`l.workerOnce.Do(l.startRefuseWorker); select { case l.refuseQ <- c: default: _ = c.Close() }`), where `startRefuseWorker` runs a background goroutine draining `l.refuseQ` in FIFO order and writing `refusal503` (with `refuseWriteTimeout = 50 * time.Millisecond`) before closing each socket. The accept loop itself never writes to `c` or blocks.
  * In `internal/transport/listen.go`:
    * Add `OnReject func(status int, reason string)` to `ServerConfig`.
    * In `NewServer(cfg ServerConfig)`: resolve `onReject := cfg.OnReject`; if `onReject == nil`, check `if rh, ok := cfg.Handler.(interface{ RecordReject(status int, reason string) }); ok { onReject = rh.RecordReject }`. Pass `var onOverCap func(); if onReject != nil { onOverCap = func() { onReject(http.StatusServiceUnavailable, rejectOverCap) } }` into `newLimitListener(raw, cfg.MaxConnections, onOverCap)`.
  * In `internal/transport/review1_test.go`:
    * Add `TestOverCapRefusalCountsRejectionOffAcceptLoop`: verify that over-cap `503` refusals invoke `OnReject` with `(http.StatusServiceUnavailable, "connection_cap_exceeded")`, complete off the accept loop without stalling admission of a freed slot, and degrade cleanly by closing excess sockets if `refuseQ` is saturated while still counting every refusal.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/transport -run "TestOverCapRefusal|TestLimitListener"`

#### Unit 2.4 (`084.004-T`): Wire `ConnectOptions.Validate` and `ServerConfig.OnReject`, count wrong-path `404` and library `websocket.Accept` refusals (without a second response or hub registration churn), and exclude benign draining `503`s (`7EF4DA98`, `FE0D5ECD`, `86BE6345`, `21F9D21A`)

* **Domain**: `code` (`size: S`, `complexity: medium`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/transport/handler.go`
  * `internal/transport/remote.go`
  * `internal/transport/remote_test.go`
* **Changes needed**:
  * In `internal/transport/handler.go`:
    * Define rejection reason constants `rejectOverCap = "connection_cap_exceeded"`, `rejectNotFound = "wrong_path"`, and `rejectHandshake = "invalid_handshake"`.
    * Export `func (h *Handler) RecordReject(status int, reason string)`: when `reason == rejectDraining`, log at `Debug` (`h.log.Debug("transport: upgrade refused while draining", "status", status, "reason", reason)`) and return without incrementing `h.rejected` or calling `h.cfg.OnReject`; otherwise log `h.log.Warn("transport: upgrade rejected", "status", status, "reason", reason)`, increment `total := h.rejected.Add(1)`, and invoke `if h.cfg.OnReject != nil { h.cfg.OnReject(status, reason, total) }`. Have `(*Handler).reject` call `h.RecordReject(status, reason)` before writing headers and status.
    * In `(*Handler).ServeHTTP`: for `r.URL.Path != WSPath`, call `h.reject(w, http.StatusNotFound, rejectNotFound)` (setting `Connection: close`, writing `404` with empty body, and counting `wrong_path`).
    * Add `func validWebSocketHandshake(r *http.Request) bool` (requiring single header values `len(r.Header.Values("Sec-WebSocket-Version")) == 1 && len(r.Header.Values("Sec-WebSocket-Key")) == 1`, `strings.TrimSpace(r.Header.Get("Sec-WebSocket-Version")) == "13"`, and that `strings.TrimSpace(r.Header.Get("Sec-WebSocket-Key"))` base64-decodes via `base64.StdEncoding.DecodeString` to `16` bytes) and `type acceptResponseRecorder struct { http.ResponseWriter; status int; wroteHeader bool }` (setting `Connection: close` on `WriteHeader` when `code >= 400`, and forwarding `Hijack` and `Flush` to `w`).
    * Add `func (h *Handler) grantValidator(g *Grant) func(gen string) hub.Reason` which re-checks `g` against `h.cfg.Creds.Validate(g, gen)` (returning `hub.ReasonShuttingDown` on `ErrShuttingDown`, `hub.ReasonCredentialExpired` on `ErrCredentialExpired`, `hub.ReasonGenerationMismatch` on `ErrGenerationMismatch`, `hub.ReasonCredentialRevoked` on any other error).
    * In `(*Handler).upgrade`:
      * Prepare `ar := r.Clone(r.Context()); ar.Header.Del("Origin")`, `opts := acceptOptions()`, and `rec := &acceptResponseRecorder{ResponseWriter: w, status: http.StatusBadRequest}`.
      * If `!validWebSocketHandshake(r)`: invoke `if c, _ := websocket.Accept(rec, ar, &opts); c != nil { _ = c.CloseNow() }` (so the library writes its single RFC 6455 error response without touching `Hub.Connect`), call `h.RecordReject(rec.status, rejectHandshake)` (without writing a second response), and return.
      * Pass `Validate: h.grantValidator(grant)` in `hub.ConnectOptions`: if `Hub.Connect` returns an error, map `errors.Is(err, hub.ErrCredentialRevoked)` to `h.reject(w, http.StatusUnauthorized, rejectCredential)`, `errors.Is(err, hub.ErrShuttingDown)` to `h.reject(w, http.StatusServiceUnavailable, rejectDraining)`, and any other error to `h.reject(w, http.StatusServiceUnavailable, rejectHubRegister)`.
      * Call `conn, err := websocket.Accept(rec, ar, &opts)`: if `err != nil`, call `h.RecordReject(rec.status, rejectHandshake)` (without writing a second response) and return.
    * Update GoDoc on `Handler.RejectedUpgrades` and `Remote.RejectedUpgrades` in `internal/transport/handler.go` and `internal/transport/remote.go`, and wire `OnReject: handler.RecordReject` explicitly in `NewRemote`'s `ServerConfig` (`internal/transport/remote.go`).
  * In `internal/transport/remote_test.go`:
    * Add `TestRejectedUpgradesCountsOverCap404AndAcceptRefusalsAndExcludesDraining`:
      1. Verify wrong-path `404` (`wrong_path`), over-cap `503` (`connection_cap_exceeded`), and an authenticated `/ws` request with an unsupported non-empty `Sec-WebSocket-Version` (`"99"`, `invalid_handshake`) each increment `RejectedUpgrades()` and invoke `OnReject` without writing a second response or consuming a hub hydration slot.
      2. Verify that after `RevokeCredentials` / `Drain`, a `503` draining refusal does NOT increment `RejectedUpgrades()` or invoke `OnReject`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/transport -run "TestRemote|TestRejectedUpgrades|TestDrain|TestVerifyShutdown"`

#### Unit 2.5 (`084.005-T`): Add `transport-layering-boundary` and `wsclient-standalone-boundary` depguard rules in `.golangci.yml` (`7EF4DA98`)

* **Domain**: `config` (`size: XS`, `complexity: low`)
* **Posture**: `migration-first`
* **Files affected** (1 file):
  * `.golangci.yml`
* **Changes needed**:
  * In `.golangci.yml` under `linters.settings.depguard.rules`:
    * Add `transport-layering-boundary` (`files: ["**/internal/transport/**/*.go"]`, denying `github.com/softwaresalt/intercom-go/internal/app`, `github.com/softwaresalt/intercom-go/internal/tui`, and `github.com/softwaresalt/intercom-go/internal/agent/copilot`).
    * Add `wsclient-standalone-boundary` (`files: ["**/internal/transport/wsclient/*.go"]`, denying `github.com/softwaresalt/intercom-go/internal/transport`).
    * Add `lower-layers-no-transport` (`files: ["**/internal/hub/**/*.go", "**/internal/agent/**/*.go", "**/internal/tui/**/*.go", "**/internal/config/**/*.go", "**/internal/pathsafe/**/*.go"]`, denying `github.com/softwaresalt/intercom-go/internal/transport`).
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race ./internal/transport/...`

#### Unit 2.6 (`084.006-T`): Reject URL fragments and opaque URLs in `wsclient.Dial` and add `config<->transport`, wire-constant drift, and layering conformance tests (`7EF4DA98`, `FE0D5ECD`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/transport/wsclient/client.go`
  * `internal/transport/wsclient/review3_test.go`
  * `internal/transport/conformance_test.go`
* **Changes needed**:
  * In `internal/transport/wsclient/client.go`:
    * Update `checkURL(rawURL string, u *url.URL) error` (called from `Dial` with `o.URL, u`) to return `ErrInvalidURL` when `u.Opaque != "" || u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.RawFragment != "" || strings.Contains(rawURL, "#")`.
  * In `internal/transport/wsclient/review3_test.go`:
    * Extend `TestDialRefusesUserinfoAndQuery` with fragment (`"ws://" + host + "/ws#frag"`, `"ws://" + host + "/ws#"`) and opaque (`"ws:opaque/ws"`, `"wss:example.com/ws"`) URL test cases asserting `errors.Is(err, ErrInvalidURL)` and zero server hits.
  * Create `internal/transport/conformance_test.go`:
    1. `TestConfigAndTransportRemoteConformance`: verify `config.DefaultRemoteMaxConnections == DefaultMaxConnections` and `== hub.DefaultMaxRemote`, and verify over a table of valid and invalid `allowed_origins` and `allowed_hosts` that `config.Validate()` and `NewHandler` / `NewRemote` agree on canonical origin and host acceptance/rejection.
    2. `TestWSClientAndServerWireConstantDrift`: verify `wsclient.Path == WSPath`, `wsclient.Subprotocol == SubprotocolV1`, bearer prefix parity, action op names (`prompt`, `abort_turn`, `resolve_permission`), decision names (`allow_once`, `deny`), reply types (`action.accepted`, `action.refused`), refusal reason constants, envelope kind constants, and `wsclient.ReadLimit` vs hub hydrate bounds.
    3. `TestTransportPackageLayeringAndDepguardConfig`: verify via `go list` imports and `.golangci.yml` inspection that `internal/transport` does not import `internal/app`, `internal/tui`, or `internal/agent/copilot`, `internal/transport/wsclient` does not import `internal/transport`, and lower/sibling packages (`internal/hub`, `internal/agent`, `internal/tui`, `internal/config`, `internal/pathsafe`) do not import `internal/transport`, and that `.golangci.yml` defines `transport-layering-boundary`, `wsclient-standalone-boundary`, and `lower-layers-no-transport`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/transport/... -run "TestDialRefusesUserinfoAndQuery|TestConfigAndTransportRemoteConformance|TestWSClientAndServerWireConstantDrift|TestTransportPackageLayering"`

#### Unit 2.7 (`084.007-T`): Rate-limit `pumpRejects` forwarding of rejected-upgrade notices in `internal/app` (`6617141A`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (2 files):
  * `internal/app/assemble.go`
  * `internal/app/reject_notice_test.go`
* **Changes needed**:
  * In `internal/app/assemble.go`:
    * Define `const defaultRejectNoticeInterval = 250 * time.Millisecond` and add `rejectInterval time.Duration` on `App` (zero selects `defaultRejectNoticeInterval`).
    * Update `(*App).pumpRejects(ui *localUI)` (delegating to `(*App).pumpRejectsTo(done <-chan struct{}, send func(tui.SecurityNoticeMsg))`) so the first notice after an idle period is delivered immediately via `send(msg)`, which starts a cooldown timer for `interval` (`250ms` by default) and suspends reading `a.rejectCh` until the timer fires (while `a.onReject` coalesces the highest-`Total` notice in `a.rejectCh` during the cooldown window). When `done` closes, stop any active timer and return immediately.
  * In `internal/app/reject_notice_test.go`:
    * Add `TestPumpRejectsRateLimitsFloodToOneSendPerInterval`: drive a burst of 200 rejections across two cooldown intervals into `pumpRejectsTo` and assert that `send` is invoked at most once per interval while delivering the final highest `Total` (`200`), and that closing `done` mid-cooldown returns immediately without waiting for the timer.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/app -run "TestOnReject|TestPumpRejects|TestRejectedUpgradeIsShownOnLocalTUI"`

#### Unit 2.8 (`084.008-T`): Ratify `Host` default-port non-equivalence in ADR MVP-2 and update operator runbook for rejected-upgrade accounting and rate-limiting (`FE0D5ECD`, `86BE6345`, `21F9D21A`, `6617141A`, `ACD578AE`)

* **Domain**: `docs` (`size: S`, `complexity: low`)
* **Posture**: `migration-first`
* **Files affected** (2 files):
  * `docs/decisions/2026-10-19-intercom-go-mvp-adr-2-credentials-and-origin.md`
  * `docs/runbooks/intercom-operator-runbook.md`
* **Changes needed**:
  * In `docs/decisions/2026-10-19-intercom-go-mvp-adr-2-credentials-and-origin.md` (§MVP-D11 `Decision: Host validation`):
    * Amend the `Host` port canonicalization text in place so it matches the 060-S implementation addendum, `docs/config-reference.md`, and `internal/transport/hostorigin.go` (`hostKey`): unlike `Origin` (which carries an `http`/`https` scheme and drops `:80`/`:443`), `Host` carries no scheme, so an `allowed_hosts` entry without a port matches only a `Host` header without a port, and explicit `:80` / `:443` ports are distinct keys.
  * In `docs/runbooks/intercom-operator-runbook.md` (sections 3, 7, and 9):
    * Update section 3 (`remote_approve_destructive`) and section 9 (`Approve is refused remotely`) to state that clipped write previews (`preview_clipped: true`) and unknown/unmodeled kinds are always local-TUI-only for approval.
    * Update section 7 (`Rejected upgrades on the status line`) and section 9 (`Remote client gets 503` / `Status line shows rejected upgrades`) to document that the counter covers handler-level upgrade refusals (`400`, `401`, `403`, `503 hub_refused`), wrong-path `404`s (`wrong_path`), over-cap `503`s from the connection limiter (`connection_cap_exceeded`), and library-generated `websocket.Accept` handshake refusals (`invalid_handshake`), excludes benign shutdown `draining` `503`s, and is coalesced and rate-limited to at most one TUI update per 250 ms.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go run ./tools/gatecheck write-path --self-test --root .`

## Dependency Graph

### Shipment-Level DAG

* `076-S` (`labels: ["dag-root"]`, `predecessor_source: "declared_root"`, under explicit operator P-001 policy exception over `062-S`, code baseline `075-S` `a289894` / `d5b209e`; Covering Feature `083-F`: `083.001-T -> 083.002-T -> 083.003-T -> 083.004-T`)
  * `<-blocks-- 077-S` (Covering Feature `084-F`: `084.001-T -> 084.002-T -> 084.003-T -> 084.004-T -> 084.005-T -> 084.006-T -> 084.007-T -> 084.008-T`)

### Intra-Shipment Task Dependencies

* **Within `076-S` (`083-F`)**: `083.001-T` -> `083.002-T` -> `083.003-T` -> `083.004-T`.
* **Within `077-S` (`084-F`)**: `084.001-T` -> `084.002-T` -> `084.003-T` -> `084.004-T` -> `084.005-T` -> `084.006-T` -> `084.007-T` -> `084.008-T` (serialized because `084.002-T` and `084.004-T` consume `hub.ConnectOptions.Validate`, `hub.ErrCredentialRevoked`, and `hub.ReasonLocalOnlyAction` from `084.001-T`, `084.004-T` consumes `ServerConfig.OnReject` / `limitListener` over-cap hooks from `084.003-T`, and `084.005-T` updates `.golangci.yml` before `084.006-T` runs `TestTransportPackageLayeringAndDepguardConfig`).

## Decisions and Rationale

1. **Two single-feature shipments (`076-S..077-S`)**: Satisfies Ship Step 6.1(a1) `S2` (`n == 1` covering feature per shipment) and cleanly separates permission-approval/summary/modal anti-spoofing (`076-S` / `083-F`) from WebSocket transport wire/authz, pre-auth rejection accounting, and `pumpRejects` rate-limiting (`077-S` / `084-F`).
2. **Fail-closed local-TUI-only approval of `PreviewClipped` write permissions (`ACD578AE`)**: Because the truncated tail of a clipped write diff/content preview is not sent over the wire and the hub cannot verify that a remote client displays a warning or requires a clipped-write confirmation, `authorize` refuses remote approvals of `previewClipped: true` write permissions with `ReasonDestructiveRemoteRefused` even when `remote_approve_destructive = true`, while preserving remote `deny` and local-TUI approval.
3. **Callback-based `ConnectOptions.Validate` for atomic hub-step credential revalidation (`7EF4DA98`, `026CF750`)**: Passing `Validate func(gen string) Reason` on `hub.ConnectOptions` lets the hub goroutine re-verify credential validity atomically inside `admission`/`completeConnect` and `processAction` without `internal/hub` importing `internal/transport`.
4. **Excluding benign `draining` `503`s while counting over-cap `503`s, wrong-path `404`s, and `websocket.Accept` refusals (`86BE6345`, `21F9D21A`, `FE0D5ECD`)**: Makes `max_connections` slot exhaustion, path probes, and bad WebSocket handshakes visible on the TUI status line without false-positive alerts during normal process shutdown.

## Risks and Caveats

* **`TestHubPackageHasNoBareChannelSends` AST guard (`internal/hub/degraded_test.go`)**: Every channel send in `internal/hub/*.go` production files must remain a `select` communication clause (or `trySend`).
* **`TestPolicyNeverReadsClientSuppliedFields` reflection guard (`internal/hub/authz_test.go`)**: `hub.OperatorAction` struct fields must remain `Type,RequestID,Decision,Prompt` in exact order.
* **Existing golden tests (`internal/tui/testdata/modal.golden`)**: `renderModalHeaders` preserves the exact output of `wrapHeader` for `kind:`, `tool call:`, and `warning:`.

## Plan Hardening Signals

* **public API, schema, or contract change**: **PRESENT** — Adds `hub.MaxDecisionFeedbackBytes`, `hub.ErrCredentialRevoked`, `hub.ReasonCredentialExpired`, `hub.ReasonGenerationMismatch`, `hub.ReasonLocalOnlyAction`, `hub.ConnectOptions.Validate`, `transport.MaxFeedbackBytes`, `transport.ReasonLocalOnlyAction`, `transport.ServerConfig.OnReject`, and amends `hub.authorize` for `PreviewClipped` write permissions and `Handler.RejectedUpgrades` / `OnReject` coverage.
* **security, auth, permission, or compliance-sensitive behavior**: **PRESENT** — Modifies remote permission authorization (`ACD578AE`), permission summary invisible-rune and homoglyph sanitization (`D46C1B38`), TUI modal header continuation indentation (`7BDC38FF`), and atomic hub-step credential revalidation (`7EF4DA98`, `026CF750`).
* **migration, backfill, destructive data/config action, or irreversible step**: **ABSENT** — No data migration or destructive action.
* **external integration, operator checkpoint, or external dependency**: **ABSENT** — Zero new external dependencies in `go.mod`.
* **high runtime, rollout, or rollback risk**: **PRESENT** — Modifies pre-auth listener accept loop refusal handling (`limiter.go`), WebSocket upgrade handshake validation (`handler.go`), and TUI rejection notice pumping (`assemble.go`).

Requires plan hardening: yes

## Runtime Verification and Closure

* **Shipment `076-S` (`083-F`)**: Remote clipped-write approval refusal, shell/write summary invisible-rune and homoglyph anti-spoofing, and shared TUI modal header continuation helper. Verified by `GOTOOLCHAIN=go1.26.10 go test -race ./internal/hub/... ./internal/agent/... ./internal/tui/... ./internal/config/...`, `GOTOOLCHAIN=go1.26.10 go vet ./...`, `gofmt -l .`, and `GOTOOLCHAIN=go1.26.10 go run ./tools/gatecheck write-path --self-test --root .`.
* **Shipment `077-S` (`084-F`)**: Transport wire/authz hardening, atomic hub-step credential revalidation, rejected-upgrade accounting (`over-cap 503`, `404`, `websocket.Accept`, `draining` exclusion), `wsclient` URL fragment/opaque rejection, `depguard` transport layering rules, and `pumpRejects` 250ms rate-limiting. Verified by `GOTOOLCHAIN=go1.26.10 go test -race ./...`, `GOTOOLCHAIN=go1.26.10 go vet ./...`, `gofmt -l .`, `scripts/check-depguard-fixtures.sh`, and `GOTOOLCHAIN=go1.26.10 go run ./tools/gatecheck write-path --self-test --root .`. Rollback for each shipment is a clean `git revert` of its merge commit.

## Plan Hardening

### 1. Why Hardening Was Required and Context Consulted

Hardening was triggered by three signals: (1) authorization and wire-contract additions across `internal/hub`, `internal/transport`, and `internal/transport/wsclient`, (2) permission-approval and summary/modal anti-spoofing controls (`ACD578AE`, `D46C1B38`, `7BDC38FF`), and (3) pre-auth listener accept-loop and WebSocket upgrade refusal handling (`FE0D5ECD`, `86BE6345`, `21F9D21A`, `6617141A`).

**Reinforced instructions and compound learnings consulted**:

* `.github/instructions/constitution.instructions.md` (Principles I, II, III, V, VI, VIII, XI)
* `.github/instructions/strict-safety.instructions.md` (`ProposedAction` / `ActionRisk` / `ActionResult` classification)
* `docs/compound/2026-10-20-tui-permission-modal-safety-and-linux-test-clock.md`
* `docs/compound/2026-10-20-websocket-close-semantics-and-linux-ci-pitfalls.md`
* `docs/compound/2026-10-20-backlogit-cli-pitfalls-and-permission-outcome-lessons.md`
* `docs/compound/2026-09-30-go-msys-bash-bridging-and-unicode-case-mapping-gap.md`
* `docs/compound/workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md`

### 2. Protected Invariants

1. **No bare channel sends in `internal/hub` (`TestHubPackageHasNoBareChannelSends`)**: Every channel send in `internal/hub/*.go` production files must live inside a `select` statement (or use `trySend`), never a bare `ch <- v`.
2. **Zero new production `pathsafe.Root.Resolve` call sites and zero filesystem write primitives**: `go run ./tools/gatecheck write-path --self-test --root .` must exit 0 after every task.
3. **No direct SDK or adapter imports in `internal/tui` (`TestPackageHasNoDirectAgentOrSDKReach`) and strict `internal/transport` layering**: `internal/tui` non-test files must not import `copilot-sdk` or `internal/agent/copilot`; `internal/transport` must not import `internal/app`, `internal/tui`, or `internal/agent/copilot`; and `internal/transport/wsclient` must not import `internal/transport`.
4. **Closed `OperatorAction` struct layout (`TestPolicyNeverReadsClientSuppliedFields`)**: `hub.OperatorAction` fields remain `Type,RequestID,Decision,Prompt`.
5. **Non-blocking accept loop and single-response WebSocket handshake rejections (`TestOverCapRefusalDoesNotStallAcceptLoop`)**: `limitListener.Accept` keeps `refuseWriteTimeout = 50 * time.Millisecond`, decrements `l.active` and records `onReject` synchronously in FIFO order on the accept loop, and enqueues refused sockets non-blockingly onto a bounded 64-slot `refuseQ` channel drained in FIFO order off the accept goroutine (closing immediately on the accept loop only when `refuseQ` is saturated); `websocket.Accept` refusals are recorded via `acceptResponseRecorder` without writing a second HTTP response.

### 3. Strict-Safety Action Classifications (`ProposedAction` / `ActionRisk` / `ActionResult`)

| Action ID | `ProposedAction` Summary & Targets | `change_kind` | `ActionRisk` | `approval_required` | Rollback | `ActionResult` |
|---|---|---|---|---|---|---|
| `PA-01` | Refuse remote approval of `PreviewClipped` write permissions (`ReasonDestructiveRemoteRefused` even when `remote_approve_destructive = true`), clear non-deny `Feedback` and cap deny `Feedback` at `1024` bytes in `internal/hub`, harden `sanitizeBound`/`sanitizeIdent`/`sanitizeShellURL` and `indentURLsFraming` against visually blank non-space non-`Cf` runes and `urls:` homoglyphs/fullwidth lookalikes in `internal/agent/copilot`, introduce shared `renderModalHeaders` in `internal/tui`, and update ADR MVP-3, design doc §5.2/§6.1, and `docs/config-reference.md` (`076-S` / `083-F`). | `local edit` (hub authz, adapter summary sanitization, TUI modal headers & docs) | `high` | `false` (pre-approved under `DARK_MODE_ACTIVE` scope) | `git revert` of `076-S` merge commit | `planned` |
| `PA-02` | Add `ConnectOptions.Validate` atomic hub-step credential revalidation in `internal/hub`, enforce `Feedback` bounds and `local_only_action` refusals via `decodedAction` in `internal/transport`, dispatch over-cap `503` writes off the accept loop via `refuseQ` and count over-cap `503`s (`connection_cap_exceeded`), wrong-path `404`s (`wrong_path`), and library `websocket.Accept` refusals (`invalid_handshake`, skipping `Hub.Connect` on invalid handshake headers) while excluding benign draining `503`s, add `.golangci.yml` depguard rules, reject URL fragments/opaque URLs in `wsclient` and add `config<->transport` / wire-constant / layering conformance tests, rate-limit `pumpRejects` to 250ms in `internal/app`, and update ADR MVP-2 and the operator runbook (`077-S` / `084-F`). | `local edit` (hub, transport, wsclient, app, `.golangci.yml` & docs) | `high` | `false` (pre-approved under `DARK_MODE_ACTIVE` scope) | `git revert` of `077-S` merge commit | `planned` |

### 4. Deepened Verification, Rollback, and Review-Gate Requirements

* **AST and structural guards**: Run `GOTOOLCHAIN=go1.26.10 go test -race ./internal/hub -run TestHubPackageHasNoBareChannelSends` after every `internal/hub` task, `GOTOOLCHAIN=go1.26.10 go test -race ./internal/tui -run TestPackageHasNoDirectAgentOrSDKReach` after every `internal/tui` task, and `GOTOOLCHAIN=go1.26.10 go run ./tools/gatecheck write-path --self-test --root .` after every shipment.
* **Review-gate capability requirement (P-012)**: Plan review (Step 4) MUST spawn the always-on and triggered cross-model persona subagents (`Constitution Reviewer`, `Go Reviewer`, `Scope Boundary Auditor`, `Learnings Researcher`, `Architecture Strategist`, `Security Lens Reviewer`), using `model_routing.anchor_review` (`gpt-6.1-sol` / `high`) for `Architecture Strategist`, and append a `## Plan Review` section with literal `dispatch_mode:` and `decision:` markers before harvest.
* **Unresolved operator decisions**: None.

<!-- plan-review-attempt: 1 -->

## Plan Review

```text
dispatch_mode: multi-agent
decision: PASS
```

### Gate Decision and Rationale

* **Verdict**: **PASS** (after incorporating all P1, P2, and P3 findings from the 6-persona multi-agent review pass directly into the plan and deliberation artifacts).
* **Plan Hardening Requirement**: `Requires plan hardening: yes` — **SATISFIED** (`## Plan Hardening` section present with `ProposedAction` / `ActionRisk` / `ActionResult` classifications `PA-01`..`PA-02` and protected invariants).

### Persona Coverage Table

| Persona | Mode | Model Route | Initial Verdict | Final Verdict (Post-Refinement) |
|---|---|---|---|---|
| **Constitution Reviewer** | Subagent | Default (`gemini-4-argon`) | ADVISORY (0 P0, 0 P1, 1 P2, 0 P3) | **PASS** |
| **Go Reviewer** | Subagent | Default (`gemini-4-argon`) | PASS (0 P0, 0 P1, 0 P2, 3 P3) | **PASS** |
| **Scope Boundary Auditor** | Subagent | Default (`gemini-4-argon`) | PASS (0 P0, 0 P1, 0 P2, 0 P3) | **PASS** |
| **Learnings Researcher** | Subagent | Default (`claude-haiku-5.5`) | PASS (6 high-confidence solutions verified) | **PASS** |
| **Architecture Strategist** | Anchor Subagent | Anchor (`gpt-6.1-sol` / `high`) | FAIL -> Remediated (1 P1, 4 P2) | **PASS** |
| **Security Lens Reviewer** | Subagent | Default (`gemini-4-argon`) | PASS (0 P0, 0 P1, 0 P2, 2 P3) | **PASS** |

### Findings and Applied Remediations

1. **[P1 — Remediated] Unit 2.3 (`084.003-T`) Non-Blocking Off-Accept-Loop `503` Refusal Queue (`Architecture Strategist`)**:
   * *Finding*: Falling back to a synchronous 50ms socket write on the accept loop when an 8-slot semaphore is saturated would restore accept-loop backpressure under a refusal flood.
   * *Remediation*: Updated Unit 2.3 (`084.003-T`) and the deliberation to enqueue over-cap sockets non-blockingly (`select`/`default`) onto a bounded 64-slot `refuseQ` channel drained in FIFO order off the accept goroutine by a background refusal worker (`writeRefusal503` with `refuseWriteTimeout = 50ms` + `Close`), closing `c` immediately on the accept loop with zero wait if `refuseQ` is ever full while still counting every refusal via `l.onReject()`.
2. **[P2 — Remediated] Task Ordering Between `.golangci.yml` Depguard Rules and `TestTransportPackageLayeringAndDepguardConfig` (`Constitution Reviewer` & `Architecture Strategist`)**:
   * *Finding*: Running `TestTransportPackageLayeringAndDepguardConfig` (which inspects `.golangci.yml`) before the `.golangci.yml` task would fail at the first task commit.
   * *Remediation*: Ordered Unit 2.5 (`084.005-T`, `.golangci.yml` `depguard` rules) ahead of Unit 2.6 (`084.006-T`, `wsclient` URL check and `internal/transport/conformance_test.go`).
3. **[P2 — Remediated] Unit 2.2 (`084.002-T`) Transport-Owned `decodedAction` Wrapper (`Architecture Strategist`)**:
   * *Finding*: Overloading `hub.OperatorAction.Prompt` with local-only wire operation names when `Type == hub.OpUnknown` conflates domain prompt text with wire metadata.
   * *Remediation*: Updated Unit 2.2 (`084.002-T`) to use a transport-internal `decodedAction` struct (`decodeWireAction` / `dispatchWire` / `refusalForWire`), keeping `hub.OperatorAction.Prompt` strictly for `OpPrompt` text and preserving `decodeAction` unchanged for existing tests.
4. **[P2 — Remediated] Units 2.1, 2.3, and 2.4 (`084.001-T`, `084.003-T`, `084.004-T`) `ConnectOptions.Validate` Execution/Lock Contract and Explicit `ServerConfig.OnReject` Wiring (`Architecture Strategist`)**:
   * *Remediation*: Specified the non-blocking O(1) execution and lock-ordering contract on `ConnectOptions.Validate` (`hub goroutine -> CredentialStore.mu`, never reverse) and exported `(*Handler).RecordReject`, wiring `OnReject: handler.RecordReject` explicitly in `NewRemote`'s `ServerConfig` while retaining the interface fallback in `NewServer`.
5. **[P3 — Remediated] Value Receiver Consistency, `unregister` Map Cleanup, `dispatchWire` Shutdown-Revocation Check Order, and `validWebSocketHandshake` Single-Header / Defensive-Close Guards (`Go Reviewer` & `Security Lens Reviewer`)**:
   * *Remediation*: Aligned `(Model).renderModalHeaders` value receiver in the Requirements Trace, ensured `delete(h.validators, id)` runs inside `(*Hub).unregister`, placed the `local_only_action` check in `dispatchWire` after `if v.reason != ""`, required single header values in `validWebSocketHandshake(r)`, set `Connection: close` on `acceptResponseRecorder.WriteHeader` for `code >= 400`, and defensively closed any non-nil `*websocket.Conn` in the `!validWebSocketHandshake(r)` branch.


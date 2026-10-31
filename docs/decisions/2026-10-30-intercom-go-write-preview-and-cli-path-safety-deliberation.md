---
title: "Write-permission preview safety, shell URL anti-spoofing, unknown-kind authorization, and Copilot CLI path and [remote] config hardening (773B2CB2, 70315264, 56CE7077, FD093EFC, FC76C5D8)"
description: "Deliberation for Capability Set 1 (773B2CB2, 70315264, 56CE7077, FD093EFC, FC76C5D8): add structured PreviewClipped propagation and non-spoofable TUI warning banner, harden shellSummary URL framing, enforce local-TUI-only approval for unknown permission kinds, unify cli_path classification with UNC/device rejection and regular-file/LookPath resolution, and reject unknown [remote] keys and hex IPv4 hosts."
date: 2026-10-30
topic: "Write-permission preview safety, CLI path trust, and [remote] authorization/config hardening (773B2CB2, 70315264, 56CE7077, FD093EFC, FC76C5D8)"
depth: "deep"
decision_status: "decided"
promoted_to: "plan"
stash_ids:
  - "773B2CB2"
  - "70315264"
  - "56CE7077"
  - "FD093EFC"
  - "FC76C5D8"
linked_artifacts:
  - "docs/plans/2026-10-30-intercom-go-write-preview-and-cli-path-safety-plan.md"
tags:
  - "security"
  - "permissions"
  - "config"
  - "copilot-adapter"
  - "tui"
  - "hub"
---

## Problem Frame

Under `DARK_MODE_ACTIVE` (`Option 1: Run the pipeline in dark factory mode.`, activated at `2026-10-30T23:04:57Z`, `merge_approval_pre_authorized=true`, `admin_fallback_pre_authorized=false`), this Stage session processes Capability Set 1 across five `DEFERRED SCOPE EXPANSION` stash entries spanning write-permission preview safety, shell URL framing, unknown-kind permission authorization, Copilot `cli_path` trust and resolution, and `[remote]` configuration validation:

* **`773B2CB2`** (`priority: high`, `kind: bug`): In `internal/agent/copilot/permission.go`, `writeSummary` clips a large write diff or new-file body to fit within `maxSummaryRunes` (`2048` runes), appends the plain-text notice `"\n[write preview truncated]"`, and returns `truncated = false`. Because the notice is plain text inside `Summary`, file content can imitate it, and neither `agent.PermissionRequest`, `hub.PermissionRequestedPayload`, nor `internal/tui` carries a structured flag distinguishing a clipped write preview from an unclipped one. Additionally, `shellSummary` appends `"\nurls: " + strings.Join(urls, ", ")` without sanitizing embedded newlines, whitespace, or commas inside `PossibleURLs` entries and without preventing a multi-line `FullCommandText` from starting a spoofed `urls:` line at column 0 of the summary.
* **`70315264`** (`priority: medium`, `kind: task`): (a) ADR MVP-3 and `internal/hub/authz.go` (`destructive(k)`) treat `PermissionKindUnknown`, `PermissionKindOther`, and any unrecognized `PermissionKind` identically to `shell` and `write`, so enabling `[remote].remote_approve_destructive = true` allows a remote approver to approve unmodeled/unknown SDK permission kinds whose arguments the adapter cannot even extract (`Summary` carries only `RawKind`); (b) `docs/config-reference.md` should state the three-tier precedence for `[timeouts].approval_seconds` (`omitted -> 3600s config default`, `1..86400 -> N seconds`, `0 -> hub.DefaultPermissionTimeout 2m`).
* **`56CE7077`** (`priority: medium`, `kind: task`): Design §6.2 and `validateCLIPath` in `internal/config/validate.go` and `internal/agent/copilot/adapter.go` treat Windows UNC paths (`\\host\share\...`, `//host/share/...`) and Windows device-namespace paths (`\\?\...`, `\\.\...`, `\??\...`) as valid via `filepath.IsAbs`, permitting `copilot.cli_path` to trigger outbound SMB authentication or address device namespaces.
* **`FD093EFC`** (`priority: medium`, `kind: task`): `internal/config` and `internal/agent/copilot` maintain duplicate copies of the `cli_path` lexical classifier that can drift; `validateCLIPath` checks `os.Stat` existence for absolute paths without verifying `info.Mode().IsRegular()` (accepting directories or device nodes); and bare executable names are not resolved once via `exec.LookPath` to an absolute regular-file path before reaching `copilotsdk.StdioConnection`.
* **`FC76C5D8`** (`priority: low`, `kind: task`): Four `[remote]` follow-ups from `058-S` review: (1) reject (not only report) any unknown key at or below `[remote]` instead of relying solely on the credential-substring denylist (`S4`); (2) reject hexadecimal IPv4 host forms such as `0x7f.0x1` (`inet_aton` / WHATWG URL style) in `allowed_hosts` and `allowed_origins`; (3) lock `hub.DefaultMaxRemote` (`8`) and `config.DefaultRemoteMaxConnections` (`8`) against drift (`F11`); and (4) verify wiring of `remote.max_connections` into `hub.Config.MaxRemote` and `remote.token_lifetime_seconds` into credential minting (`S5`, already shipped in `060-S` and `062-S`).

### Safety Mode and Strict-Safety Classification

Per Constitutional Principle VIII and `.github/instructions/strict-safety.instructions.md`, this session operates under **`careful`**, **`investigate-first`**, and **`freeze-scope`** modes (constraining planning, harvest, and archival strictly to `773B2CB2`, `70315264`, `56CE7077`, `FD093EFC`, and `FC76C5D8` on branch `chore/stage-write-preview-and-cli-path-safety`).

* **`ProposedAction`**: Scoped Stage deliberation, implementation planning, plan hardening, multi-agent plan review, backlog harvest (`081-F` and `082-F`), shipment assembly (`074-S` and `075-S`), and consumed-stash archival for `773B2CB2`, `70315264`, `56CE7077`, `FD093EFC`, and `FC76C5D8`
* **`ActionRisk`**: `high` (downstream implementation modifies security-sensitive permission arbitration, TUI approval modal rendering, Copilot CLI binary path validation, and `[remote]` config validation)
* **`approval_required`**: `false` for Stage artifact creation and dark-mode pre-authorized PR merge (`merge_approval_pre_authorized=true`); `true` for any `autoharness shipment cascade-close` mutating execution during Ship closure (`admin_fallback_pre_authorized=false`)
* **`ActionResult`**: `approved` (under `DARK_MODE_ACTIVE` operator trigger)

### P-021 C5 and C6 Triage and Provenance Audit

All five entries carry the literal `DEFERRED SCOPE EXPANSION` token, forcing the Step 2 `deliberate` route before planning (P-021 C6) and triggering both Step 1 triage obligations: (A) Unconditional Duplicate Detection across active stash, archived stash, and queue items, and (B) Late-Identifier Reconciliation against `docs/closure/` records (with `backlogit stash edit` writes deferred until after the Step 1.9 branch gate passed):

| Stash ID | Originating Source Refs | Trigger A: Unconditional Duplicate Scan | Trigger B: Late-Identifier Reconciliation |
|---|---|---|---|
| `773B2CB2` | `task: 079.001-T`, `feature: 079-F`, `shipment: 072-S`, `PR: #214` (reconciled from `PR: N/A`; merge commit `d0d863fae5ec4bdc1d197b35800223a786ed87c5`), `review-thread: N/A` | **CLEAN DUPLICATE SCAN**: Scanned all active entries in `.backlogit/stash.jsonl`, `.backlogit/archive/`, and `.backlogit/queue/`; no duplicate entry describes `PreviewClipped`, clipped write-permission previews, or `shellSummary` `urls:` line spoofing. | **Reconciled `PR: #214`; no late review-thread identifier found**: Recovered `pr: 214` and `merge_commit_sha: d0d863fae5ec4bdc1d197b35800223a786ed87c5` from `docs/closure/072-S-079-F-post-merge-closure.md` and updated `773B2CB2` in place via `backlogit stash edit` after Step 1.9. `review-thread: N/A` stands as the truthful terminal record (pre-PR local Adversarial/Security/Correctness review finding). |
| `56CE7077` | `task: 065.004-T`, `feature: 065-F`, `shipment: 057-S`, `PR: #175` (reconciled from `PR: N/A`; merge commit `e43ad660aa2f0bd98c9eff58833cbfbe0e4963a2`), `review-thread: N/A` | **CLEAN DUPLICATE SCAN**: Scanned `.backlogit/stash.jsonl`, `.backlogit/archive/`, and `.backlogit/queue/`; no duplicate entry covers UNC or device-namespace `cli_path` policy (`FD093EFC` is its distinct sibling covering classifier sharing, `LookPath`, and regular-file checks). | **Reconciled `PR: #175`; no late review-thread identifier found**: Recovered `pr: 175` and `merge_commit_sha: e43ad660aa2f0bd98c9eff58833cbfbe0e4963a2` from `docs/closure/057-S-065-F-post-merge-closure.md` and updated `56CE7077` via `backlogit stash edit` after Step 1.9. `review-thread: N/A` stands as the truthful terminal record (pre-PR local adversarial review finding F5). |
| `FD093EFC` | `task: 065.004-T`, `feature: 065-F`, `shipment: 057-S`, `PR: #175` (reconciled from `PR: N/A`; merge commit `e43ad660aa2f0bd98c9eff58833cbfbe0e4963a2`), `review-thread: N/A` | **CLEAN DUPLICATE SCAN**: Scanned `.backlogit/stash.jsonl`, `.backlogit/archive/`, and `.backlogit/queue/`; no duplicate entry describes sharing the `cli_path` classifier, `exec.LookPath` resolution to an absolute path, or regular-file checks. | **Reconciled `PR: #175`; no late review-thread identifier found**: Recovered `pr: 175` and `merge_commit_sha: e43ad660aa2f0bd98c9eff58833cbfbe0e4963a2` from `docs/closure/057-S-065-F-post-merge-closure.md` and updated `FD093EFC` via `backlogit stash edit` after Step 1.9. `review-thread: N/A` stands as the truthful terminal record (pre-PR local review). |
| `70315264` | `tasks: 067.004-T/067.002-T`, `feature: 067-F`, `shipment: 059-S`, `PR: #179` (reconciled from `PR: N/A`; merge commit `8c9f0f299fb2784ec0cbe03b843c84402927595d`), `review-thread: N/A` | **CLEAN DUPLICATE SCAN**: Scanned `.backlogit/stash.jsonl`, `.backlogit/archive/`, and `.backlogit/queue/`; no duplicate entry covers unknown-kind remote approval policy under `remote_approve_destructive` or `timeouts.approval_seconds` precedence documentation. | **Reconciled `PR: #179`; no late review-thread identifier found**: Recovered `pr: 179` and `merge_commit_sha: 8c9f0f299fb2784ec0cbe03b843c84402927595d` from `docs/closure/059-S-067-F-post-merge-closure.md` and updated `70315264` via `backlogit stash edit` after Step 1.9. `review-thread: N/A` stands as the truthful terminal record. |
| `FC76C5D8` | `task: 066.006-T`, `feature: 066-F`, `shipment: 058-S`, `PR: #177` (reconciled from `PR: N/A`; merge commit `ec447f7e7fe6140bc99da28802e96b4317860aa3`), `review-thread: N/A` | **CLEAN DUPLICATE SCAN**: Scanned `.backlogit/stash.jsonl`, `.backlogit/archive/`, and `.backlogit/queue/`; no duplicate entry covers unknown-key rejection under `[remote]`, hex/octal IPv4 host rejection, or `DefaultMaxRemote` parity. | **Reconciled `PR: #177`; no late review-thread identifier found**: Recovered `pr: 177` and `merge_commit_sha: ec447f7e7fe6140bc99da28802e96b4317860aa3` from `docs/closure/058-S-066-F-post-merge-closure.md` and updated `FC76C5D8` via `backlogit stash edit` after Step 1.9. `review-thread: N/A` stands as the truthful terminal record. |

## Research Findings

### Prior Learnings Consulted (`docs/compound/`)

The `Learnings Researcher` pass (`confidence: high`, 37 files searched across `docs/compound/`) surfaced eight directly applicable institutional learnings:

* `docs/compound/2026-10-20-tui-permission-modal-safety-and-linux-test-clock.md`: View and `approvable()` in `internal/tui` must share one viewability predicate, call `syncModal` on every pending-set transition, and sanitize at the display boundary while keeping column-0 headers non-spoofable by summary lines.
* `docs/compound/2026-10-20-process-wiring-credential-panel-and-approval-disarm.md`: Any change to TUI permission modal rendering or hints must preserve the single top-of-`onKey` disarm choke point (`if k.Type != tea.KeyEnter { m.armed = "" }`).
* `docs/compound/2026-10-20-backlogit-cli-pitfalls-and-permission-outcome-lessons.md`: Hub permission changes must preserve the atomic `Deliver`/`Revoke` `PermissionOutcome` state machine; never run two `backlogit` CLI mutations in parallel.
* `docs/compound/2026-10-19-hub-test-oracles-and-toolchain-pitfalls.md`: Avoid the forbidden retired-architecture substring in any identifier, comment, or test name; keep zero bare channel sends in `internal/hub/*.go` (`TestHubPackageHasNoBareChannelSends`).
* `docs/compound/2026-09-08-verify-go-stdlib-internals-claims-against-goroot-source.md`: Verify `filepath.IsAbs`, `exec.LookPath`, `os.Stat` `Mode().IsRegular()`, and `netip.ParseAddr` / `url.Parse` behavior against Go stdlib semantics and empirical table-driven tests.
* `docs/compound/2026-09-30-go-msys-bash-bridging-and-unicode-case-mapping-gap.md`: In `exec.LookPath` tests on Windows, ensure fake executables in `PATH` are valid regular files with `.exe` suffix (`PATHEXT`) so `LookPath` resolves the intended fixture.
* `docs/compound/workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md`: Cross-package contract changes (`PreviewClipped`, unknown-kind policy, `cli_path` classifier, `[remote]` validation) must update producer, payload/transport, policy, consumer UI, tests, and documentation together.

### Codebase Inspection Findings (`investigate-first`)

1. **`773B2CB2` — Clipped write-permission previews and `shellSummary` URL framing**:
   * In `internal/agent/copilot/permission.go` (lines 138-167), `writeSummary(r *rpc.PermissionRequestWrite)` checks `r.FileName` first (`fnSan, fnTrunc, fnLen := sanitizeBound(r.FileName, maxSummaryRunes)`). When `fnTrunc` is `false` and `body` (`r.Diff` or `*r.NewFileContents`) exceeds `avail := maxSummaryRunes - fnLen - 1`, it clips `body` to `avail - suffixLen - 1`, appends `writePreviewTruncatedSuffix` (`"\n[write preview truncated]"`), and returns `truncated = false, fullLen = fnLen`.
   * If a clipped write preview instead returned `truncated = true`, `permissionBridge.handle` (line 277) and `hub.handleAsk` (line 301) would unconditionally auto-deny every file write whose diff or new content exceeds ~2,000 runes (~40 lines of code), breaking normal agent file creation and editing. Conversely, relying solely on the in-band string `"\n[write preview truncated]"` inside `Summary` allows file content to spoof the notice and gives `hub.PermissionRequestedPayload` and `internal/tui` no structured signal that the write preview is incomplete.
   * In `internal/tui/permission.go` (lines 147-166), `permissionBody(p)` renders `m.wrapHeader("kind: ", p.Kind, inner)` and `m.wrapHeader("tool call: ", p.ToolCallID, inner)` at column 0 and indents every line of `p.Summary` by two spaces (`"  " + l`). Therefore, adding a structured `PreviewClipped bool` field to `agent.PermissionRequest`, `hub.PermissionRequestedPayload` (`json:"preview_clipped,omitempty"`), and `tui.permission` allows `permissionBody` to render a non-spoofable column-0 header `m.wrapHeader("warning: ", "write preview clipped — full content not shown", inner)` and `modalView` to render explicit clipped-write confirmation hints (`"APPROVE CLIPPED WRITE? press Enter to confirm, any other key cancels"` when armed, and `"y approve clipped write (then Enter)   n deny"` when approvable) whenever `p.PreviewClipped` is `true`.
   * In `shellSummary(r *rpc.PermissionRequestShell)` (lines 125-136), `u.URL` values are joined with `", "` before `sanitizeBound` runs (which preserves `\n` and `\t`), so a URL entry containing `\n`, `\t`, spaces, or `,` can inject forged lines or split URL entries; and a multi-line `r.FullCommandText` containing a line whose trimmed prefix starts with `"urls:"` can mimic the adapter's `\nurls: ` trailer. Sanitizing each `u.URL` so any whitespace (`unicode.IsSpace`), comma `','`, control, format, or line/paragraph separator (`\u2028`, `\u2029`) rune becomes `utf8.RuneError` (`\uFFFD`), and indenting any line in `r.FullCommandText` whose `strings.TrimLeftFunc(line, unicode.IsSpace)` starts with `"urls:"` (case-insensitively) by prefixing `"  "`, prevents both injection vectors (including non-ASCII Unicode space prefixes such as `U+00A0`) while keeping single-line and normal commands (`"curl x\nurls: https://a.example, https://b.example"`) byte-for-byte identical.

2. **`70315264` — Unknown permission kinds with `remote_approve_destructive` and `timeouts.approval_seconds` documentation**:
   * In `internal/hub/authz.go` (lines 19-40), `destructive(k)` is `return k != agent.PermissionKindRead`, and `authorize` permits a remote `RoleApprover` to approve any kind when `in.remoteApproveDestructive` is `true`.
   * However, for `PermissionKindOther` (unmodeled SDK permission kinds such as `PermissionRequestMemory` or future SDK tools) and `PermissionKindUnknown` (zero value / unclassified), `toDomainRequest` cannot extract target paths, commands, or diffs (`Summary` contains only `rawKind`). Allowing a remote approver to approve an unmodeled or unclassified permission kind violates least privilege.
   * Splitting non-`read` kinds into **modeled destructive kinds** (`PermissionKindShell`, `PermissionKindWrite`, `PermissionKindURL` — unlocked for remote approvers when `remote_approve_destructive = true`) and **unknown/unmodeled kinds** (`PermissionKindUnknown`, `PermissionKindOther`, and any undefined `PermissionKind` value — **always refused for remote approval** with `ReasonDestructiveRemoteRefused` even when `remote_approve_destructive = true`, while remote `DecisionDeny` and local TUI approval remain allowed) enforces fail-closed authorization for unrecognized tool kinds.
   * For `[timeouts].approval_seconds`, `internal/config/default.go` sets `Timeouts.ApprovalSeconds = 3600` (1 hour), `internal/app/assemble.go` maps `0` to `time.Duration(0)`, and `internal/hub/hub.go` maps `PermissionTimeout == 0` to `DefaultPermissionTimeout` (`2 * time.Minute`). Documenting this three-tier precedence explicitly in `docs/config-reference.md` closes item `70315264(b)` with zero config-schema change.

3. **`56CE7077` & `FD093EFC` — Copilot `cli_path` UNC/device policy, shared classifier, regular-file check, and `LookPath` resolution**:
   * Both `internal/config` and `internal/agent/copilot` can share a single exported classifier in `internal/pathsafe/lexical.go`: `pathsafe.ClassifyCLIPath(path string) CLIPathKind` returning one of `CLIPathEmpty`, `CLIPathUNCOrDevice`, `CLIPathAbs`, `CLIPathDriveRelative`, `CLIPathRelative`, or `CLIPathBare`. Because `internal/config` already imports `internal/pathsafe` and `internal/pathsafe` has no dependency on `internal/config` or the SDK, placing `ClassifyCLIPath` in `internal/pathsafe` avoids coupling `internal/agent/copilot` to the TOML config schema package.
   * Checking any two leading path separators (`len(path) >= 2 && isSlash(path[0]) && isSlash(path[1])`, where `isSlash` matches `'\\'` or `'/'`) or any NT device-namespace prefix (`len(path) >= 4 && isSlash(path[0]) && path[1] == '?' && path[2] == '?' && isSlash(path[3])`) **before** `filepath.IsAbs(path)` in `ClassifyCLIPath` rejects Windows UNC paths (`\\server\share\copilot.exe`, `//server/share/copilot.exe`, `\/server\share\copilot.exe`, `/\server\share\copilot.exe`), extended-length/UNC paths (`\\?\...`, `\\?\UNC\...`, `\/?\...`), and Win32/NT device namespaces (`\\.\...`, `\??\...`, `/??/...`) uniformly on every `GOOS` (portable config text rule), and classifying `"."` and `".."` as `CLIPathRelative` prevents bare dot segments from falling through to `CLIPathBare`.
   * For `CLIPathAbs`, both `internal/config.validateCLIPath` and `internal/agent/copilot.resolveCLIPath` check `info, err := os.Stat(cliPath)` and require `info.Mode().IsRegular()` (rejecting directories, FIFOs, sockets, and device nodes while still permitting symlinks whose target is a regular file).
   * For `CLIPathBare`:
     * In `internal/config/validate.go`, `validateCLIPath` resolves `resolved, err := exec.LookPath(cliPath)`. When `err != nil`, it appends the existing non-fatal advisory `"copilot.cli_path '%s' was not resolvable via PATH"` to `report.Warnings` and leaves `c.Copilot.CLIPath` unchanged (preserving `TestValidateRule5BareCLIPathUnresolvableProducesWarningOnly`). When `err == nil`, it converts `resolved` via `filepath.Abs`, verifies `pathsafe.ClassifyCLIPath(abs) == pathsafe.CLIPathAbs` and `info.Mode().IsRegular()`, and commits `c.Copilot.CLIPath = abs` in `Validate()`'s deferred-commit block (preserving Protected Invariant I6: failed validation never mutates the receiver).
     * In `internal/agent/copilot/adapter.go`, `resolveCLIPath` (called by `validateConfig` / `New`) switches on `pathsafe.ClassifyCLIPath(cfg.CLIPath)`. When `cfg.CLIPath` was already resolved to `CLIPathAbs` by `config.Validate()`, `LookPath` is not called a second time (`resolveCLIPath` only verifies `os.Stat` + `info.Mode().IsRegular()`). When `copilot.New` is called directly with a bare name (`CLIPathBare`), `resolveCLIPath` resolves it once via `exec.LookPath` + `filepath.Abs` + regular-file verification and passes the resulting absolute path to `copilotsdk.StdioConnection{Path: resolved}` (and records it on `a.cfg.CLIPath`).

4. **`FC76C5D8` — `[remote]` unknown-key rejection, hex/octal IPv4 host rejection, and `DefaultMaxRemote` parity**:
   * **Already-shipped sub-items verified on `main`**: `internal/app/assemble.go:191` already wires `MaxRemote: cfg.Remote.MaxConnections` into `hub.New`, and `internal/transport/remote.go:52` already wires `Lifetime: LifetimeFromSeconds(o.Remote.TokenLifetimeSeconds)` into `NewCredentialStore`.
   * **Sub-item 1 (`S4` unknown keys under `[remote]`)**: In `internal/config/remote.go`, `rejectRemoteCredentialKeys(keys []toml.Key)` currently scans `[remote]` keys only for `isCredentialShapedKey(seg)`. Extending it with a second pass that rejects any remaining key at or below `[remote]` (`len(k) >= 2 && strings.EqualFold(k[0], "remote")` that is not `len(k) == 2 && isKnownRemoteKey(k[1])`) with `apperr.Newf(apperr.KindConfig, "%s is an unknown key under [remote]; only known [remote] schema keys are permitted", sanitizeKeyPath(k.String()))` makes `[remote]` a closed schema section that never echoes values, while preserving credential-key error priority (`"is a credential-shaped key..."`) and `report.UnknownKeys` report-on-error population.
   * **Sub-item 2 (hex/octal numeric host forms in `allowed_hosts` and `allowed_origins`)**: In `internal/config/remote.go`, `numericHostProblem(h string)` currently checks `strings.Trim(last, "0123456789") == ""`, which misses hexadecimal IPv4 labels (`0x7f.0x1`, `127.0.0.0x1`, `0x7f000001`, `0x0`) because `'x'` and `'a'..'f'` are not decimal digits. Updating the numeric-label detector `isIPv4NumberLabel(last)` to match either all decimal digits OR `"0x"` followed by hex digits `[0-9a-f]` (matching the WHATWG URL host parser's IPv4 number rule) causes `numericHostProblem` to reject all hex and octal `inet_aton` host forms in both `allowed_hosts` and `allowed_origins`.
   * **Sub-item 3 (`F11` `DefaultMaxRemote` parity)**: Defining `const DefaultRemoteMaxConnections = 8` in `internal/config/remote.go` (used by `config.Default().Remote.MaxConnections` in `internal/config/default.go`) and locking `hub.DefaultMaxRemote == config.DefaultRemoteMaxConnections` and `hub.MaxRemoteHardLimit == config.MaxRemoteConnections` in `internal/config/remote_test.go` (test-scope import of `internal/hub`) prevents constant drift without adding a production import edge between `internal/hub` and `internal/config`.

## Options Evaluated

### Option 1: Structured `PreviewClipped` + Non-Spoofable TUI Warning, Local-TUI-Only Unknown Kinds, Shared `pathsafe.ClassifyCLIPath` with UNC/Regular-File/Single-`LookPath` Resolution, and Closed `[remote]` Validation Across Two Chained Shipments (`074-S` and `075-S`) (Recommended)

Partition the five stash entries into two cohesive, single-feature shipments (`n == 1` covering feature per shipment per Ship Step 6.1(a1) `S2`):

* **Shipment 1 (`074-S`, covering feature `081-F`, blocked by `073-S`)**: Permission Preview Safety, Shell URL Anti-Spoofing, and Unknown-Kind Local-TUI Authorization (`773B2CB2`, `70315264`) — 5 atomic tasks (`081.001-T`..`081.005-T`).
* **Shipment 2 (`075-S`, covering feature `082-F`, blocked by `074-S`)**: Copilot `cli_path` Trust/Resolution Hardening and `[remote]` Config Validation (`56CE7077`, `FD093EFC`, `FC76C5D8`) — 5 atomic tasks (`082.001-T`..`082.005-T`).

* **Pros**:
  * Preserves usability for legitimate multi-KB file writes while giving `agent.PermissionRequest`, `hub.PermissionRequestedPayload` (live and `session.hydrate`), and `internal/tui` a structured, non-spoofable `PreviewClipped` signal and explicit TUI warning banner/confirmation prompt.
  * Closes `shellSummary` URL line-framing and delimiter spoofing without breaking existing `shellSummary` format expectations.
  * Enforces least privilege by making unknown/unmodeled permission kinds (`PermissionKindUnknown`, `PermissionKindOther`, undefined values) local-TUI-only even when `remote_approve_destructive = true`.
  * Eliminates `cli_path` classifier drift via `pathsafe.ClassifyCLIPath`, blocks UNC/device paths on all platforms, requires regular files for `cli_path`, and resolves bare names once via `exec.LookPath`.
  * Closes `[remote]` against unknown keys and hex IPv4 bypasses (`0x7f.0x1`).
* **Cons**: Spans two shipments (`074-S` and `075-S`, 10 tasks total).
* **Effort estimate**: Medium (`10 tasks × S`).
* **Fit**: Satisfies all security, usability, architectural boundary, and 2-hour granularity requirements.

### Option 2: Auto-Deny (`Truncated = true`) Any Write Permission Whose Diff or New Content Exceeds 2,048 Runes

Set `out.Truncated = true` in `writeSummary` whenever `r.Diff` or `r.NewFileContents` exceeds `maxSummaryRunes - fnLen - 1`.

* **Pros**: Slightly smaller change in `internal/hub` and `internal/tui`.
* **Cons**: Causes `permissionBridge.handle` and `hub.handleAsk` to auto-deny every file creation or edit larger than ~2 KB (~40 lines of code) without ever asking the operator, crippling normal coding workflows and reversing `072-S` Decision 4.
* **Effort estimate**: Low.
* **Fit**: Poor — unacceptable usability regression for normal agent file writes.

### Option 3: Single Monolithic Shipment Combining All 5 Stash Entries Under One Feature

Bundle all 10 tasks across `internal/agent`, `internal/hub`, `internal/tui`, `internal/pathsafe`, `internal/config`, `internal/app`, and `docs/` into a single 10-task shipment.

* **Pros**: Only one PR lifecycle.
* **Cons**: Mixes two distinct architectural domains (runtime permission arbitration/TUI modal vs startup `cli_path`/`[remote]` config validation) into one large PR, increasing review surface and blast radius.
* **Effort estimate**: Medium.
* **Fit**: Inferior to the clean two-shipment split (`074-S` -> `075-S`).

## Trade-off Comparison

| Criterion | Option 1 (Structured `PreviewClipped` + 2 Shipments `074-S`/`075-S`) | Option 2 (Auto-Deny Writes > 2,048 Runes) | Option 3 (Single 10-Task Monolithic Shipment) |
|---|---|---|---|
| Prevents spoofed/unnoticed clipped write approvals | Yes — structured `PreviewClipped` flag + column-0 `warning:` header + `APPROVE CLIPPED WRITE?` prompt | Yes, by denying all writes > 2,048 runes | Yes |
| Preserves usability for >2 KB file writes | Yes — operator can review and approve with explicit clipped-write warning | No — auto-denies every write > 2,048 runes | Yes |
| Enforces unknown-kind local-TUI-only authz & `cli_path` / `[remote]` hardening | Yes | Partial | Yes |
| PR blast radius & domain cohesion | High cohesion — Shipment 1 (`074-S`) is permission/TUI; Shipment 2 (`075-S`) is `cli_path`/`[remote]` config | Medium | Low cohesion — 10 tasks across 7 packages in one PR |

## Decision

Adopt **Option 1** and promote immediately to implementation planning (`docs/plans/2026-10-30-intercom-go-write-preview-and-cli-path-safety-plan.md`):

1. **Shipment 1 (`074-S`, covering feature `081-F`, `blocks`-dependent on `073-S`)**: Write-Permission Preview Safety, Shell URL Anti-Spoofing, and Unknown-Kind Local-TUI Authorization (`773B2CB2`, `70315264`).
2. **Shipment 2 (`075-S`, covering feature `082-F`, `blocks`-dependent on `074-S`)**: Copilot `cli_path` Classification/Resolution Hardening and `[remote]` Config Validation (`56CE7077`, `FD093EFC`, `FC76C5D8`).

## Rejected Alternatives

* **Option 2 (Auto-Deny Writes > 2,048 Runes)** was rejected because auto-denying every file write whose diff or content exceeds 2,048 runes prevents the agent from creating or editing ordinary source files.
* **Option 3 (Single Monolithic Shipment)** was rejected to keep each shipment's PR cohesive and small (`5` tasks per shipment, separated by permission/TUI runtime vs `cli_path`/`[remote]` configuration hardening).

## Unresolved Questions

None. All design contracts, code touchpoints, and already-shipped sub-items (`FC76C5D8` S5 wiring in `internal/app/assemble.go` and `internal/transport/remote.go`) are verified against `main`.

## Risks and Mitigations

* **Risk — Existing TUI golden snapshots (`internal/tui/testdata/*.golden`)**: Modifying `permissionBody` or `modalView` unconditionally would break `TestModalShowsRequestAndApproveGoesThroughIngress` (`testdata/modal.golden`).
  * **Mitigation**: Render the column-0 `warning: write preview clipped — full content not shown` line in `permissionBody` and the clipped-write hints in `modalView` **only** when `p.PreviewClipped` is `true`, leaving unclipped permission modals byte-for-byte identical to `testdata/modal.golden`.
* **Risk — Cross-platform `exec.LookPath` behavior in unit tests (`internal/config` and `internal/agent/copilot`)**: On Windows, `exec.LookPath` requires an executable extension in `PATHEXT` (such as `.exe`), whereas on POSIX it requires execute permission bits (`0o755`).
  * **Mitigation**: Create test fixtures in `t.TempDir()` with `0o755` permissions and platform-appropriate executable naming (or create both `copilot` and `copilot.exe` with `0o755` inside the test `PATH` directory) so `exec.LookPath("copilot")` resolves deterministically on both Windows and Linux.
* **Risk — Schema-to-docs coupling tests in `internal/config/docs_test.go` and `internal/config/remote_test.go`**: `TestConfigReferenceCoversValidationMessages` and `TestConfigReferenceDocumentsRemote` assert exact substrings in `docs/config-reference.md`.
  * **Mitigation**: Preserve all existing tested validation message strings in `internal/config/validate.go`, `internal/config/remote.go`, and `docs/config-reference.md` while adding the new UNC/device, regular-file, and `[remote]` unknown-key messages.

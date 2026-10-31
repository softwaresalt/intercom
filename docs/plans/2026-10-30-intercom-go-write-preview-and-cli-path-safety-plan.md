---
title: "Write-permission preview safety, shell URL anti-spoofing, unknown-kind authorization, and Copilot CLI path and [remote] config hardening: implementation plan"
description: "Implementation plan for Capability Set 1 across two chained single-feature shipments: 074-S / 081-F (773B2CB2, 70315264) and 075-S / 082-F (56CE7077, FD093EFC, FC76C5D8)."
date: 2026-10-30
source: "docs/decisions/2026-10-30-intercom-go-write-preview-and-cli-path-safety-deliberation.md"
stash_ids:
  - "773B2CB2"
  - "70315264"
  - "56CE7077"
  - "FD093EFC"
  - "FC76C5D8"
---

## Problem Frame

Capability Set 1 (`DARK_MODE_ACTIVE` Option 1) resolves five `DEFERRED SCOPE EXPANSION` entries across two cohesive single-feature shipments (`074-S` covering `081-F`, and `075-S` covering `082-F`):

1. **Shipment 1 (`074-S` / `081-F`, `773B2CB2` and `70315264`)**:
   * **`773B2CB2` (`priority: high`, `kind: bug`, from `079.001-T` / `079-F` / `072-S` / `PR #214`)**:
     * In `internal/agent/copilot/permission.go`, `writeSummary` clips a write diff or new-file content body exceeding the remaining `maxSummaryRunes` (`2048` runes) budget and appends `"\n[write preview truncated]"`, returning `truncated = false`. Because no structured `PreviewClipped` flag exists on `agent.PermissionRequest`, `hub.PermissionRequestedPayload`, or `tui.permission`, file content can imitate the plain-text notice, and the local TUI renders no out-of-band column-0 warning header or clipped-write confirmation prompt.
     * In `internal/agent/copilot/permission.go`, `shellSummary` joins `r.PossibleURLs` with `", "` after `r.FullCommandText + "\nurls: "` without sanitizing whitespace, commas, or newlines inside `u.URL` and without preventing a multi-line `r.FullCommandText` from starting a line with `"urls:"` at column 0 of the summary.
   * **`70315264` (`priority: medium`, `kind: task`, from `067.004-T`/`067.002-T` / `067-F` / `059-S` / `PR #179`)**:
     * In `internal/hub/authz.go`, `destructive(k)` treats `PermissionKindUnknown`, `PermissionKindOther`, and undefined `PermissionKind` values identically to modeled destructive kinds (`shell`, `write`, `url`), allowing a remote approver to approve unmodeled/unknown SDK permission kinds when `[remote].remote_approve_destructive = true` even though their summaries carry only `RawKind`.
     * `docs/config-reference.md` needs explicit documentation of the three-tier `[timeouts].approval_seconds` precedence (`omitted -> 3600s config default`, `1..86400 -> N seconds`, `0 -> hub.DefaultPermissionTimeout 2m`) and the updated `remote.remote_approve_destructive` policy.
2. **Shipment 2 (`075-S` / `082-F`, `56CE7077`, `FD093EFC`, and `FC76C5D8`)**:
   * **`56CE7077` & `FD093EFC` (`priority: medium`, `kind: task`, from `065.004-T` / `065-F` / `057-S` / `PR #175`)**:
     * `internal/config/validate.go` and `internal/agent/copilot/adapter.go` duplicate the §6.2 `cli_path` lexical classifier; both treat Windows UNC (`\\server\share`, `//server/share`) and device-namespace (`\\?\`, `\\.\`, `\??\`) paths as valid when `filepath.IsAbs` returns `true`; `validateCLIPath` accepts directories and non-regular files for absolute paths; and bare executable names are not resolved once via `exec.LookPath` to an absolute regular-file path before reaching the SDK.
   * **`FC76C5D8` (`priority: low`, `kind: task`, from `066.006-T` / `066-F` / `058-S` / `PR #177`)**:
     * In `internal/config/remote.go`, unknown keys under `[remote]` that do not match the credential-substring denylist are reported in `Report.UnknownKeys` rather than rejected (`S4`); `numericHostProblem` only checks decimal-digit final labels and misses hexadecimal IPv4 forms (`0x7f.0x1`, `127.0.0.0x1`, `0x7f000001`, `0x0`) in `allowed_hosts` and `allowed_origins`; and `DefaultRemoteMaxConnections` (`8`) should be exported and locked by test against `hub.DefaultMaxRemote` (`8`) (`F11`). (The `S5` wiring of `remote.max_connections` and `remote.token_lifetime_seconds` is already shipped in `internal/app/assemble.go:191` and `internal/transport/remote.go:52`.)

## Requirements Trace

| Requirement ID | Source Stash / Decision Ref | Requirement Summary | Target Implementation Unit |
|---|---|---|---|
| `R-1` | `773B2CB2` (Option 1) | Add `PreviewClipped bool` to `agent.PermissionRequest`; set `PreviewClipped = true` and `FullLen = fnLen + 1 + bodyLen` in `writeSummary` when the write diff/content preview is clipped while keeping `Truncated = false` and `Summary` `<= maxSummaryRunes`. | Unit 1.1 (`081.001-T`) |
| `R-2` | `773B2CB2` (Option 1) | Harden `shellSummary` in `internal/agent/copilot/permission.go`: sanitize each `PossibleURLs` entry by replacing whitespace (`unicode.IsSpace`), commas `','`, control, format, and line/paragraph separator (`\u2028`, `\u2029`) runes with `utf8.RuneError` (`\uFFFD`), and indent any line in `r.FullCommandText` whose trimmed prefix (`strings.TrimLeftFunc(line, unicode.IsSpace)`) begins with `"urls:"` (case-insensitively) by prefixing `"  "`. | Unit 1.1 (`081.001-T`) |
| `R-3` | `70315264(a)` (Option 1) | Enforce local-TUI-only approval for unknown/unmodeled permission kinds (`PermissionKindUnknown`, `PermissionKindOther`, and any undefined `PermissionKind` value) in `internal/hub/authz.go`: a remote approver attempting `DecisionAllowOnce` on an unknown kind is always refused with `ReasonDestructiveRemoteRefused` even when `remote_approve_destructive = true`, while `PermissionKindShell`, `PermissionKindWrite`, and `PermissionKindURL` remain remotely approvable when `remote_approve_destructive = true`. | Unit 1.2 (`081.002-T`) |
| `R-4` | `773B2CB2` (Option 1) | Add `PreviewClipped bool` (`json:"preview_clipped,omitempty"`) to `hub.PermissionRequestedPayload` (`internal/hub/envelope.go`) and propagate `ask.Request.PreviewClipped` through `pendingPerm`, `handleAsk` live `permission.requested` broadcasts, and `hydratePayload` `pending_permissions` snapshots. | Unit 1.2 (`081.002-T`), Unit 1.3 (`081.003-T`) |
| `R-5` | `773B2CB2` (Option 1) | Propagate `PreviewClipped` into `tui.permission` (`internal/tui/permission.go`); when `p.PreviewClipped` is `true`, render a non-spoofable column-0 header line `m.wrapHeader("warning: ", "write preview clipped — full content not shown", inner)` in `permissionBody(p)` immediately after `tool call: `, and render explicit clipped-write hints in `modalView` (`"APPROVE CLIPPED WRITE? press Enter to confirm, any other key cancels"` when armed; `"y approve clipped write (then Enter)   n deny"` when approvable). | Unit 1.4 (`081.004-T`) |
| `R-6` | `773B2CB2`, `70315264(a,b)` (Option 1) | Amend `docs/decisions/2026-10-19-intercom-go-mvp-adr-3-authorization-roles.md`, `docs/design-docs/intercom-go-backend-architecture.md` (§5.2, §6.1), and `docs/config-reference.md` (`remote.remote_approve_destructive` and `[timeouts].approval_seconds` three-tier precedence). | Unit 1.5 (`081.005-T`) |
| `R-7` | `56CE7077`, `FD093EFC` (Option 1) | Export `CLIPathKind` (`CLIPathEmpty`, `CLIPathBare`, `CLIPathAbs`, `CLIPathUNCOrDevice`, `CLIPathDriveRelative`, `CLIPathRelative`) and `ClassifyCLIPath(path string) CLIPathKind` in `internal/pathsafe/lexical.go`, classifying any path with two leading slashes (`\\`, `//`, `\/`, `/\`) or NT device prefix (`\??\`, `/??/`, etc.) as `CLIPathUNCOrDevice` ahead of `filepath.IsAbs` on every `GOOS`, and classifying `"."` and `".."` as `CLIPathRelative`. | Unit 2.1 (`082.001-T`) |
| `R-8` | `56CE7077`, `FD093EFC` (Option 1) | Update `validateCLIPath` in `internal/config/validate.go` to use `pathsafe.ClassifyCLIPath`, reject `CLIPathUNCOrDevice`, require `info.Mode().IsRegular()` for `CLIPathAbs`, and resolve `CLIPathBare` via `exec.LookPath` + `filepath.Abs` (verifying `CLIPathAbs` and `info.Mode().IsRegular()` and committing the resolved absolute path to `c.Copilot.CLIPath` in `Validate()`'s deferred commit block when `LookPath` succeeds, while retaining the non-fatal `Report.Warnings` advisory when `LookPath` fails). | Unit 2.2 (`082.002-T`) |
| `R-9` | `56CE7077`, `FD093EFC` (Option 1) | Replace `validateCLIPath` in `internal/agent/copilot/adapter.go` with `resolveCLIPath(p string) (string, error)` using `pathsafe.ClassifyCLIPath`: reject `CLIPathUNCOrDevice`, `CLIPathDriveRelative`, and `CLIPathRelative`; require `os.Stat` + `info.Mode().IsRegular()` for `CLIPathAbs`; resolve `CLIPathBare` once via `exec.LookPath` + `filepath.Abs` + regular-file verification; and store and pass the resolved absolute path to `copilotsdk.StdioConnection{Path: cfg.CLIPath}`. | Unit 2.3 (`082.003-T`) |
| `R-10` | `FC76C5D8` (Option 1) | In `internal/config/remote.go`: (a) reject any unknown key at or below `[remote]` after the credential-shaped check without echoing values; (b) reject hexadecimal IPv4 host forms (`0x7f.0x1`, `127.0.0.0x1`, `0x7f000001`, `0x0`) in `numericHostProblem` for both `allowed_hosts` and `allowed_origins`; and (c) export `DefaultRemoteMaxConnections = 8` and lock parity against `hub.DefaultMaxRemote`. | Unit 2.4 (`082.004-T`) |
| `R-11` | `56CE7077`, `FD093EFC`, `FC76C5D8` (Option 1) | Amend `docs/design-docs/intercom-go-backend-architecture.md` (§6.2, §8) and `docs/config-reference.md` for `copilot.cli_path` (`pathsafe.ClassifyCLIPath`, UNC/device rejection, regular-file check, bare-name `LookPath` resolution) and `[remote]` (unknown-key rejection and hex IPv4 host rejection). | Unit 2.5 (`082.005-T`) |

## Constitution Check

| Principle | Compliance | Notes |
|---|---|---|
| I. Safety-First Go | Compliant | All code uses standard-library types and explicit `apperr` / `fmt.Errorf` error wrapping with zero `panic` calls and zero `golangci-lint` / `go vet` / `gofmt` findings. |
| II. Test-First Development (NON-NEGOTIABLE) | Compliant | Every `code`-domain task (`081.001-T`..`081.004-T`, `082.001-T`..`082.004-T`) specifies an explicit red-to-green verification sequence in inline `_test.go` files. |
| III. Workspace Isolation and Security Boundaries | Compliant | Rejects UNC/device `cli_path` values, requires regular files, blocks unknown keys and hex IPv4 hosts under `[remote]`, and denies remote approval of unknown permission kinds. Zero new production `pathsafe.Root.Resolve` calls and zero forbidden filesystem write selectors in production code (`cmd/**`, `internal/**`). |
| IV. CLI Workspace Containment (NON-NEGOTIABLE) | Compliant | All file operations and `t.TempDir()` fixtures stay strictly inside the workspace and test temp directories. |
| V. Structured Observability | Compliant | `PreviewClipped` is surfaced as a structured wire field (`preview_clipped`) and TUI modal warning header/hint; config validation errors name exact key paths without echoing secret values. |
| VI. Single Responsibility | Compliant | Zero new external dependencies in `go.mod`; `pathsafe.ClassifyCLIPath` is placed in `internal/pathsafe/lexical.go` so `internal/agent/copilot` does not depend on `internal/config`. |
| VII. Destructive Command Approval (NON-NEGOTIABLE) | Compliant | Strengthens destructive approval controls by making unknown permission kinds local-TUI-only even when `remote_approve_destructive = true` and surfacing an explicit `APPROVE CLIPPED WRITE?` confirmation prompt in the local TUI. |
| VIII. Explicit Safety Modes for Elevated Risk | Compliant | Executed under `careful`, `investigate-first`, and `freeze-scope` modes with `ProposedAction` / `ActionRisk` / `ActionResult` entries in `## Plan Hardening`. |
| IX. Git-Friendly Persistence | Compliant | All Stage artifacts are UTF-8 Markdown / JSONL files committed to `chore/stage-write-preview-and-cli-path-safety`. |
| X. Agent Context Efficiency | Compliant | Grounded via `Learnings Researcher` and targeted file reads. |
| XI. Merge Commit History Preservation (NON-NEGOTIABLE) | Compliant | All staging and feature PRs merge via merge commits only (P-009). |

## Implementation Units

### Shipment 1 (`074-S`) — Covering Feature `081-F`: Write-Permission Preview Safety, Shell URL Anti-Spoofing, and Unknown-Kind Local-TUI Authorization (`773B2CB2`, `70315264`)

#### Unit 1.1: Emit `PreviewClipped` on clipped write summaries and harden `shellSummary` URL framing (`081.001-T`)

* **Source Stash ID**: `773B2CB2` (originating refs: `079.001-T`, `079-F`, `072-S`, `PR #214`, `review-thread: N/A`)
* **Priority**: `high`
* **Size**: `S` (`size_source: agent`, `size_ruleset_version: 2h-rule-v1`)
* **Complexity**: `medium`
* **Skill Domain**: `code`
* **Execution Posture**: `test-first`
* **Files Affected (3)**:
  * `internal/agent/agent.go`
  * `internal/agent/copilot/permission.go`
  * `internal/agent/copilot/permission_cycle2_test.go`
* **Functions Changed / Added (4)**:
  * `writeSummary` in `internal/agent/copilot/permission.go` (modified)
  * `toDomainRequest` in `internal/agent/copilot/permission.go` (modified)
  * `shellSummary` in `internal/agent/copilot/permission.go` (modified)
  * `sanitizeShellURL` in `internal/agent/copilot/permission.go` (added helper)
* **Test Scenarios (3)**:
  1. `writeSummary` / `toDomainRequest` sets `PreviewClipped = false` on unclipped write diffs/contents and `PreviewClipped = true` (with `Truncated = false`, `FullLen` equal to the unclipped rune count `fnLen + 1 + bodyLen`, and `Summary` ending with `"\n[write preview truncated]"` within `maxSummaryRunes`) when the diff or new-file content is clipped.
  2. `writeSummary` / `toDomainRequest` still sets `Truncated = true` and `PreviewClipped = false` when `FileName` alone exceeds `maxSummaryRunes` or leaves fewer than `len([]rune(writePreviewTruncatedSuffix)) + 1` runes for the preview.
  3. `shellSummary` / `toDomainRequest` preserves `"curl x\nurls: https://a.example, https://b.example"` for clean URLs, replaces whitespace (`\n`, `\r`, `\t`, `' '`, non-ASCII Unicode spaces), commas `','`, control, format, and line/paragraph separator (`\u2028`, `\u2029`) runes inside `PossibleURLs` entries with `U+FFFD` (`\uFFFD`), and prefixes `"  "` to any line in `r.FullCommandText` whose `strings.TrimLeftFunc(line, unicode.IsSpace)` starts with `"urls:"` (case-insensitively) so command text (including lines prefixed with `U+00A0` non-breaking space) cannot spoof the adapter's `\nurls: ` line.
* **Detailed Changes**:
  1. In `internal/agent/agent.go`:
     * Add `PreviewClipped bool` to `PermissionRequest` with a GoDoc comment explaining that `PreviewClipped` reports that a `write` request's diff or new-file content preview was clipped to fit `Summary` while the target file name is complete (`Truncated` is `false`, so the request is still presented to the operator with an explicit clipped-preview warning), and update the `Truncated` GoDoc comment to note this explicit `PreviewClipped` exception for write previews.
  2. In `internal/agent/copilot/permission.go`:
     * Add `sanitizeShellURL(s string) string` that replaces any rune `r` where `unicode.IsSpace(r) || r == ',' || r == '\u2028' || r == '\u2029' || unicode.IsControl(r) || unicode.Is(unicode.Cf, r)` with `utf8.RuneError`.
     * Update `shellSummary(r *rpc.PermissionRequestShell) string`:
       * Split `r.FullCommandText` on `"\n"` and for any line where `strings.HasPrefix(strings.ToLower(strings.TrimLeftFunc(line, unicode.IsSpace)), "urls:")`, prefix the line with `"  "` so no command line can start at column 0 of the summary with `"urls:"`.
       * For each non-empty `u.URL` in `r.PossibleURLs`, append `sanitizeShellURL(u.URL)` to `urls`, and if `len(urls) > 0` append `"\nurls: " + strings.Join(urls, ", ")`.
     * Update `writeSummary(r *rpc.PermissionRequestWrite) (summary string, truncated, previewClipped bool, fullLen int)`:
       * When `fnTrunc` is `true` or `avail < suffixLen+1`, return `fnSan, true, false, fnLen`.
       * When `body == ""`, return `fnSan, false, false, fnLen`.
       * Compute `bodySan, bodyTrunc, bodyLen := sanitizeBound(body, avail)` and `totalLen := fnLen + 1 + bodyLen`.
       * When `!bodyTrunc`, return `fnSan + "\n" + bodySan, false, false, totalLen`.
       * When `bodyTrunc`, compute `preview, _, _ := sanitizeBound(body, avail-suffixLen-1)` and return `fnSan + "\n" + preview + writePreviewTruncatedSuffix, false, true, totalLen`.
     * In `toDomainRequest`, assign `out.Summary, out.Truncated, out.PreviewClipped, out.FullLen = writeSummary(r)`.
  3. In `internal/agent/copilot/permission_cycle2_test.go`:
     * Extend `TestToDomainRequestSurfacesShellPossibleURLsAndWriteDiffOrContent` (and add `TestShellSummaryPreventsURLFramingSpoofing`) to assert `PreviewClipped == false` on unclipped writes, `PreviewClipped == true` and `FullLen > maxSummaryRunes` on clipped writes, `PreviewClipped == false && Truncated == true` on oversized filenames, and `shellSummary` sanitization of `PossibleURLs` (containing `\n`, `,`, spaces, control/bidi runes) and indentation of `urls:` lines (including `U+00A0` leading whitespace) inside `FullCommandText`.
* **Red-to-Green Verification Sequence (P-002 / P-004)**:
  * **Red phase**: Add `PreviewClipped bool` to `agent.PermissionRequest` and the new assertions in `internal/agent/copilot/permission_cycle2_test.go`; confirm `GOTOOLCHAIN=go1.26.10 go test ./internal/agent/copilot -run 'TestToDomainRequestSurfacesShellPossibleURLsAndWriteDiffOrContent|TestShellSummaryPreventsURLFramingSpoofing'` fails before `permission.go` is updated.
  * **Green phase**: Update `internal/agent/copilot/permission.go` and confirm `GOTOOLCHAIN=go1.26.10 go test -race ./internal/agent/...` and `GOTOOLCHAIN=go1.26.10 go test -race ./...` pass.
* **Acceptance Criteria**:
  1. `agent.PermissionRequest` declares `PreviewClipped bool`; `writeSummary` sets `PreviewClipped = true` (with `Truncated = false` and `FullLen = fnLen + 1 + bodyLen`) when and only when the write diff/content preview is clipped.
  2. `shellSummary` sanitizes whitespace, commas, control, format, and separator runes in `PossibleURLs` entries to `U+FFFD` and indents any line in `FullCommandText` whose `strings.TrimLeftFunc(line, unicode.IsSpace)` prefix begins with `"urls:"` (case-insensitively).

#### Unit 1.2: Enforce local-TUI-only approval for unknown permission kinds and declare `PreviewClipped` on `PermissionRequestedPayload` (`081.002-T`)

* **Source Stash IDs**: `70315264`, `773B2CB2`
* **Priority**: `high`
* **Size**: `S` (`size_source: agent`, `size_ruleset_version: 2h-rule-v1`)
* **Complexity**: `low`
* **Skill Domain**: `code`
* **Execution Posture**: `test-first`
* **Files Affected (3)**:
  * `internal/hub/authz.go`
  * `internal/hub/envelope.go`
  * `internal/hub/authz_test.go`
* **Functions Changed / Added (3)**:
  * `knownDestructive` in `internal/hub/authz.go` (replaces `destructive`)
  * `unknownPermissionKind` in `internal/hub/authz.go` (added)
  * `authorize` in `internal/hub/authz.go` (modified)
* **Test Scenarios (3)**:
  1. A remote `RoleApprover` attempting `DecisionAllowOnce` on `PermissionKindUnknown`, `PermissionKindOther`, or an undefined `PermissionKind(200)` is refused with `ReasonDestructiveRemoteRefused` both when `RemoteApproveDestructive` is `false` AND when `RemoteApproveDestructive` is `true`.
  2. A remote `RoleApprover` attempting `DecisionAllowOnce` on `PermissionKindShell`, `PermissionKindWrite`, or `PermissionKindURL` is refused with `ReasonDestructiveRemoteRefused` when `RemoteApproveDestructive` is `false` and allowed when `RemoteApproveDestructive` is `true`; `PermissionKindRead` is allowed regardless of `RemoteApproveDestructive`.
  3. A remote `RoleApprover` may still `DecisionDeny` any permission kind (including `PermissionKindUnknown` and `PermissionKindOther`), and the local TUI (`surfaceTUI`) may approve or deny any permission kind regardless of `RemoteApproveDestructive`.
* **Detailed Changes**:
  1. In `internal/hub/authz.go`:
     * Replace `destructive(k agent.PermissionKind) bool` with:
       * `knownDestructive(k agent.PermissionKind) bool`: returns `true` for `agent.PermissionKindShell`, `agent.PermissionKindWrite`, and `agent.PermissionKindURL`.
       * `unknownPermissionKind(k agent.PermissionKind) bool`: returns `k != agent.PermissionKindRead && !knownDestructive(k)` (`PermissionKindUnknown`, `PermissionKindOther`, and any undefined `PermissionKind` value).
     * In `authorize(in authzInput)`: when `in.op == OpResolvePermission && in.decision != agent.DecisionDeny`:
       * If `unknownPermissionKind(in.kind)`, return `false, ReasonDestructiveRemoteRefused` (local-TUI-only; never unlocked by `in.remoteApproveDestructive`).
       * If `knownDestructive(in.kind) && !in.remoteApproveDestructive`, return `false, ReasonDestructiveRemoteRefused`.
  2. In `internal/hub/envelope.go`:
     * Add `PreviewClipped bool `json:"preview_clipped,omitempty"`` to `PermissionRequestedPayload` with a GoDoc comment explaining that it is set when a write permission's diff/content preview was clipped to fit `Summary`.
  3. In `internal/hub/authz_test.go`:
     * Update `TestAuthorizeMatrixRows`, `TestAuthorizeInvariantsOverTheWholeInputSpace`, and `TestUnknownPermissionKindsAreDeniedToRemoteApprovers` so `PermissionKindUnknown`, `PermissionKindOther`, and `PermissionKind(200)` expect `(false, ReasonDestructiveRemoteRefused)` for a remote approver even when `remoteApproveDestructive: true`, while `PermissionKindShell`, `PermissionKindWrite`, and `PermissionKindURL` remain `(true, "")` when `remoteApproveDestructive: true`.
* **Red-to-Green Verification Sequence (P-002 / P-004)**:
  * **Red phase**: Update `internal/hub/authz_test.go` to assert that remote approval of `PermissionKindUnknown`, `PermissionKindOther`, and `PermissionKind(200)` is refused when `RemoteApproveDestructive: true`; confirm `GOTOOLCHAIN=go1.26.10 go test ./internal/hub -run 'TestAuthorize|TestUnknownPermissionKindsAreDeniedToRemoteApprovers'` fails before `authz.go` is updated.
  * **Green phase**: Update `internal/hub/authz.go` and `internal/hub/envelope.go`, and confirm `GOTOOLCHAIN=go1.26.10 go test -race ./internal/hub/...` and `GOTOOLCHAIN=go1.26.10 go test -race ./...` pass (including `TestHubPackageHasNoBareChannelSends`).
* **Acceptance Criteria**:
  1. `authorize` in `internal/hub/authz.go` refuses remote `DecisionAllowOnce` for `PermissionKindUnknown`, `PermissionKindOther`, and any undefined `PermissionKind` with `ReasonDestructiveRemoteRefused` even when `remoteApproveDestructive` is `true`, while keeping `shell`, `write`, and `url` governed by `remoteApproveDestructive` and keeping remote `DecisionDeny` and local TUI approval permitted for all kinds.
  2. `hub.PermissionRequestedPayload` in `internal/hub/envelope.go` declares `PreviewClipped bool `json:"preview_clipped,omitempty"``.

#### Unit 1.3: Propagate `PreviewClipped` through hub registry, `permission.requested`, and `session.hydrate` (`081.003-T`)

* **Source Stash ID**: `773B2CB2`
* **Depends On**: `081.001-T`, `081.002-T`
* **Priority**: `high`
* **Size**: `S` (`size_source: agent`, `size_ruleset_version: 2h-rule-v1`)
* **Complexity**: `low`
* **Skill Domain**: `code`
* **Execution Posture**: `test-first`
* **Files Affected (3)**:
  * `internal/hub/permission.go`
  * `internal/hub/hub.go`
  * `internal/hub/hydrate_permission_test.go`
* **Functions Changed / Added (3)**:
  * `(*permRegistry).register` in `internal/hub/permission.go` (modified)
  * `(*Hub).handleAsk` in `internal/hub/permission.go` (modified)
  * `(*Hub).hydratePayload` in `internal/hub/hub.go` (modified)
* **Test Scenarios (2)**:
  1. Posting an `agent.PermissionAsk` with `Request.PreviewClipped = true` broadcasts `KindPermissionRequested` with `PreviewClipped: true` (`"preview_clipped":true` in JSON) and includes `PreviewClipped: true` in `session.hydrate`'s `pending_permissions` slice (including across forced TUI re-hydration).
  2. Posting an `agent.PermissionAsk` with `Request.PreviewClipped = false` omits `"preview_clipped"` from the JSON encoding of both `permission.requested` and `session.hydrate` `pending_permissions` (`omitempty`).
* **Detailed Changes**:
  1. In `internal/hub/permission.go`:
     * Add `previewClipped bool` to `pendingPerm`.
     * In `(*permRegistry).register`, copy `previewClipped: ask.Request.PreviewClipped` into the registered `*pendingPerm`.
     * In `(*Hub).handleAsk`, set `PreviewClipped: e.previewClipped` on the published `PermissionRequestedPayload`.
  2. In `internal/hub/hub.go`:
     * In `(*Hub).hydratePayload`, set `PreviewClipped: e.previewClipped` on each `PermissionRequestedPayload` in `views`, and update the GoDoc comment on `Config.RemoteApproveDestructive` to state that when `false` (the default) a remote approver may not approve modeled destructive permissions (`shell`, `write`, `url`), while unknown/unmodeled kinds are always local-TUI-only even when `true`.
  3. In `internal/hub/hydrate_permission_test.go`:
     * Add `TestPermissionPreviewClippedPropagatesToBroadcastAndHydrate` verifying both `PreviewClipped: true` propagation and `PreviewClipped: false` JSON `omitempty` behavior across live `permission.requested` and `session.hydrate`.
* **Red-to-Green Verification Sequence (P-002 / P-004)**:
  * **Red phase**: Add `TestPermissionPreviewClippedPropagatesToBroadcastAndHydrate` in `internal/hub/hydrate_permission_test.go`; confirm `GOTOOLCHAIN=go1.26.10 go test ./internal/hub -run TestPermissionPreviewClippedPropagatesToBroadcastAndHydrate` fails before `permission.go` and `hub.go` are updated.
  * **Green phase**: Update `internal/hub/permission.go` and `internal/hub/hub.go`, and confirm `GOTOOLCHAIN=go1.26.10 go test -race ./internal/hub/...` and `GOTOOLCHAIN=go1.26.10 go test -race ./...` pass.
* **Acceptance Criteria**:
  1. `pendingPerm` records `ask.Request.PreviewClipped`, `(*Hub).handleAsk` sets `PreviewClipped: e.previewClipped` on `permission.requested`, and `(*Hub).hydratePayload` sets `PreviewClipped: e.previewClipped` on every unresolved entry in `pending_permissions`.
  2. `TestPermissionPreviewClippedPropagatesToBroadcastAndHydrate` passes and `TestHubPackageHasNoBareChannelSends` remains green.

#### Unit 1.4: Render non-spoofable clipped-write warning header and confirmation hints in `internal/tui` (`081.004-T`)

* **Source Stash ID**: `773B2CB2`
* **Depends On**: `081.002-T`, `081.003-T`
* **Priority**: `high`
* **Size**: `S` (`size_source: agent`, `size_ruleset_version: 2h-rule-v1`)
* **Complexity**: `low`
* **Skill Domain**: `code`
* **Execution Posture**: `test-first`
* **Files Affected (3)**:
  * `internal/tui/permission.go`
  * `internal/tui/permission_test.go`
  * `internal/hub/doc.go`
* **Functions Changed / Added (3)**:
  * `newPermission` in `internal/tui/permission.go` (modified)
  * `permissionBody` in `internal/tui/permission.go` (modified)
  * `modalView` in `internal/tui/permission.go` (modified)
* **Test Scenarios (3)**:
  1. When `PermissionRequestedPayload.PreviewClipped` is `true` (via live `permission.requested` or `session.hydrate`), `permissionBody` renders the column-0 header `warning: write preview clipped — full content not shown` after `tool call: `, and `modalView` renders `"y approve clipped write (then Enter)   n deny"` when approvable and `"APPROVE CLIPPED WRITE? press Enter to confirm, any other key cancels"` when armed with `y`.
  2. When `PermissionRequestedPayload.PreviewClipped` is `false`, `permissionBody` and `modalView` render identically to existing behavior (`testdata/modal.golden` unchanged), and a summary containing `"warning: write preview clipped"` or `"[write preview truncated]"` in file content is indented by 2 spaces and cannot spoof the column-0 `warning:` header or clipped-write hint.
  3. A long clipped write request still requires scrolling to the end (`m.modalSeen >= m.modal.TotalLineCount()`) and the two-key `y` then `Enter` confirmation (with top-of-`onKey` disarm on any other key) before approval is sent.
* **Detailed Changes**:
  1. In `internal/tui/permission.go`:
     * Add `PreviewClipped bool` to `permission` (updating `permission`'s doc comment to note the explicit `PreviewClipped` write-preview exception) and populate `PreviewClipped: p.PreviewClipped` in `newPermission`.
     * Define `const clippedWriteWarning = "write preview clipped — full content not shown"`.
     * In `permissionBody(p permission)`: build `header := m.wrapHeader("kind: ", p.Kind, inner) + "\n" + m.wrapHeader("tool call: ", p.ToolCallID, inner)`; if `p.PreviewClipped`, append `"\n" + m.wrapHeader("warning: ", clippedWriteWarning, inner)`; return `header + "\n\n" + strings.Join(lines, "\n")`.
     * In `modalView(height int)`:
       * In the `case m.armed == id:` branch: if `p.PreviewClipped`, set `hint = "APPROVE CLIPPED WRITE? press Enter to confirm, any other key cancels"`, else keep `"APPROVE? press Enter to confirm, any other key cancels"`.
       * In the `case m.approvable():` branch: if `p.PreviewClipped`, set `hint = "y approve clipped write (then Enter)   n deny"`, else keep `"y approve (then Enter to confirm)   n deny"`.
  2. In `internal/tui/permission_test.go`:
     * Add `TestClippedWritePreviewShowsColumnZeroWarningAndArmedBanner` and `TestUnclippedSummaryCannotSpoofWarningHeader` covering live delivery, `session.hydrate` delivery, scroll-to-end requirement, two-key confirmation, and anti-spoofing of the column-0 `warning:` header.
  3. In `internal/hub/doc.go`:
     * Update the package GoDoc comment under `# Permission arbitration` so it states that modeled destructive kinds (`shell`, `write`, `url`) need the remote destructive-approval opt-in, unknown kinds are local-TUI-only for approvals, and denial is always allowed.
* **Red-to-Green Verification Sequence (P-002 / P-004)**:
  * **Red phase**: Add `TestClippedWritePreviewShowsColumnZeroWarningAndArmedBanner` and `TestUnclippedSummaryCannotSpoofWarningHeader` in `internal/tui/permission_test.go`; confirm `GOTOOLCHAIN=go1.26.10 go test ./internal/tui -run 'TestClippedWritePreviewShowsColumnZeroWarningAndArmedBanner|TestUnclippedSummaryCannotSpoofWarningHeader'` fails before `internal/tui/permission.go` is updated.
  * **Green phase**: Update `internal/tui/permission.go` and `internal/hub/doc.go` and confirm `GOTOOLCHAIN=go1.26.10 go test -race ./internal/tui/...` and `GOTOOLCHAIN=go1.26.10 go test -race ./...` pass (including golden tests and `TestPackageHasNoDirectAgentOrSDKReach`).
* **Acceptance Criteria**:
  1. `tui.permission` records `PreviewClipped`; when `true`, `permissionBody` includes the non-spoofable column-0 `warning: write preview clipped — full content not shown` header and `modalView` displays the clipped-write approvable and armed confirmation hints.
  2. Unclipped modals (`PreviewClipped: false`) keep their exact existing rendering (`testdata/modal.golden` passes unmodified), and summary text cannot spoof a column-0 `warning:` header.

#### Unit 1.5: Update ADR MVP-3, backend architecture design doc, and config reference for permission preview and authorization rules (`081.005-T`)

* **Source Stash IDs**: `773B2CB2`, `70315264`
* **Depends On**: `081.002-T`, `081.003-T`, `081.004-T`
* **Priority**: `medium`
* **Size**: `S` (`size_source: agent`, `size_ruleset_version: 2h-rule-v1`)
* **Complexity**: `low`
* **Skill Domain**: `docs`
* **Execution Posture**: `characterization-first`
* **Files Affected (3)**:
  * `docs/decisions/2026-10-19-intercom-go-mvp-adr-3-authorization-roles.md`
  * `docs/design-docs/intercom-go-backend-architecture.md`
  * `docs/config-reference.md`
* **Functions Changed / Added (0)**: Documentation-only unit.
* **Test Scenarios (2)**:
  1. `GOTOOLCHAIN=go1.26.10 go test ./internal/config -run 'TestConfigReference'` passes with the updated `docs/config-reference.md`.
  2. `GOTOOLCHAIN=go1.26.10 go run ./tools/gatecheck retired-arch --root .` and `GOTOOLCHAIN=go1.26.10 go run ./tools/gatecheck write-path --self-test --root .` pass with zero findings.
* **Detailed Changes**:
  1. In `docs/decisions/2026-10-19-intercom-go-mvp-adr-3-authorization-roles.md`:
     * Amend the *Decision: permission kind classes*, *Decision: action-by-role matrix*, and *Addendum (059-S)* sections (and add a dated `074-S` amendment note) to distinguish **modeled destructive kinds** (`shell`, `write`, `url`, which a remote `approver` may approve only when `[remote].remote_approve_destructive = true`) from **unknown/unmodeled kinds** (`other`, `unknown`, and any unrecognized kind, which are **local-TUI-only** for approvals and always refused to remote surfaces with `destructive_remote_refused` even when `remote_approve_destructive = true`, while remote `deny` remains allowed).
  2. In `docs/design-docs/intercom-go-backend-architecture.md`:
     * In §5.2 (`permission.requested` bullet), document the optional boolean `preview_clipped` field set when a write diff/content preview is clipped, the local TUI's non-spoofable column-0 `warning:` header and `APPROVE CLIPPED WRITE?` confirmation prompt, and `shellSummary`'s URL sanitization and `urls:` line indentation.
     * In §6.1 (`Authorization model` bullet), note that `[remote].remote_approve_destructive` unlocks remote approval of modeled destructive kinds (`shell`, `write`, `url`) while unknown/unmodeled kinds (`other`, `unknown`) always require the local TUI.
  3. In `docs/config-reference.md`:
     * In the `[remote]` table row for `remote.remote_approve_destructive`, update the note to: `When true, remote approvers may approve modeled destructive kinds (shell, write, url); unknown/unmodeled kinds always require the local TUI.`
     * In the `[timeouts]` table row and `### [timeouts] rule` section for `timeouts.approval_seconds`, explicitly state the three-tier precedence: (1) omitting `timeouts.approval_seconds` in `config.toml` uses `config.Default().Timeouts.ApprovalSeconds = 3600` (1 hour); (2) setting `1..86400` uses that exact duration in seconds; (3) explicitly setting `0` passes `0` to `hub.Config.PermissionTimeout`, which selects `hub.DefaultPermissionTimeout` (`2m`).
* **Acceptance Criteria**:
  1. `docs/decisions/2026-10-19-intercom-go-mvp-adr-3-authorization-roles.md`, `docs/design-docs/intercom-go-backend-architecture.md` (§5.2, §6.1), and `docs/config-reference.md` accurately document `preview_clipped`, `shellSummary` URL framing, local-TUI-only unknown-kind approval, and the three-tier `timeouts.approval_seconds` precedence.
  2. All `internal/config` doc-coupling tests (`TestConfigReferenceCoversDefaultFieldNames`, `TestConfigReferenceCoversValidationMessages`, `TestConfigReferenceCoversEverySchemaKeyPath`, `TestConfigReferenceDocumentsRemote`) pass.

---

### Shipment 2 (`075-S`, blocked by `074-S`) — Covering Feature `082-F`: Copilot `cli_path` Classification/Resolution Hardening and `[remote]` Config Validation (`56CE7077`, `FD093EFC`, `FC76C5D8`)

#### Unit 2.1: Add shared `pathsafe.ClassifyCLIPath` classifier with UNC and device-namespace rejection (`082.001-T`)

* **Source Stash IDs**: `56CE7077`, `FD093EFC` (originating refs: `065.004-T`, `065-F`, `057-S`, `PR #175`, `review-thread: N/A`)
* **Priority**: `medium`
* **Size**: `S` (`size_source: agent`, `size_ruleset_version: 2h-rule-v1`)
* **Complexity**: `low`
* **Skill Domain**: `code`
* **Execution Posture**: `test-first`
* **Files Affected (2)**:
  * `internal/pathsafe/lexical.go`
  * `internal/pathsafe/lexical_test.go`
* **Functions Changed / Added (3)**:
  * `ClassifyCLIPath` in `internal/pathsafe/lexical.go` (added)
  * `hasUNCOrDevicePrefix` in `internal/pathsafe/lexical.go` (added helper)
  * `hasWindowsDrivePrefix` in `internal/pathsafe/lexical.go` (added helper)
* **Test Scenarios (3)**:
  1. `ClassifyCLIPath("")` returns `CLIPathEmpty`; `ClassifyCLIPath("copilot")` and `ClassifyCLIPath("copilot.exe")` return `CLIPathBare`.
  2. `ClassifyCLIPath` returns `CLIPathUNCOrDevice` on every `GOOS` for UNC and device-namespace paths (including mixed-slash forms: `\\server\share\copilot.exe`, `//server/share/copilot`, `\/server\share\copilot.exe`, `/\server\share\copilot.exe`, `\\?\C:\Tools\copilot.exe`, `\\?\UNC\server\share\copilot.exe`, `\\.\pipe\copilot`, `\/.\pipe\copilot`, `\??\C:\Tools\copilot.exe`, `/??/C:/Tools/copilot.exe`) before `filepath.IsAbs` is consulted.
  3. `ClassifyCLIPath` returns `CLIPathDriveRelative` for `"C:copilot"` and `"d:bin/copilot"`, `CLIPathRelative` for `"."`, `".."`, `"./bin/copilot"`, `".\bin\copilot"`, `"..\x\copilot"`, and `"bin/copilot"`, and `CLIPathAbs` for `filepath.Join(t.TempDir(), "copilot")`.
* **Detailed Changes**:
  1. In `internal/pathsafe/lexical.go`:
     * Define exported `CLIPathKind uint8` and constants:
       ```go
       type CLIPathKind uint8

       const (
           CLIPathEmpty CLIPathKind = iota
           CLIPathBare
           CLIPathAbs
           CLIPathUNCOrDevice
           CLIPathDriveRelative
           CLIPathRelative
       )
       ```
     * Implement `ClassifyCLIPath(path string) CLIPathKind`:
       1. If `path == ""` -> `return CLIPathEmpty`
       2. If `hasUNCOrDevicePrefix(path)` (`(len(path) >= 2 && isSlash(path[0]) && isSlash(path[1])) || (len(path) >= 4 && isSlash(path[0]) && path[1] == '?' && path[2] == '?' && isSlash(path[3]))`, where `isSlash(b)` is `b == '/' || b == '\\'`) -> `return CLIPathUNCOrDevice`
       3. If `filepath.IsAbs(path)` -> `return CLIPathAbs`
       4. If `hasWindowsDrivePrefix(path)` (`len(path) >= 2 && path[1] == ':' && (('A' <= path[0] && path[0] <= 'Z') || ('a' <= path[0] && path[0] <= 'z'))`) -> `return CLIPathDriveRelative`
       5. If `path == "." || path == ".." || strings.ContainsAny(path, `/\`)` -> `return CLIPathRelative`
       6. Otherwise -> `return CLIPathBare`
  2. In `internal/pathsafe/lexical_test.go`:
     * Add table-driven `TestClassifyCLIPath` covering all six `CLIPathKind` classifications across portable Windows, POSIX, and mixed-slash path spellings (including `"."` and `".."`).
* **Red-to-Green Verification Sequence (P-002 / P-004)**:
  * **Red phase**: Add `TestClassifyCLIPath` in `internal/pathsafe/lexical_test.go` against a stub `ClassifyCLIPath` returning `CLIPathEmpty`; confirm `GOTOOLCHAIN=go1.26.10 go test ./internal/pathsafe -run TestClassifyCLIPath` fails.
  * **Green phase**: Implement `ClassifyCLIPath` in `internal/pathsafe/lexical.go` and confirm `GOTOOLCHAIN=go1.26.10 go test -race ./internal/pathsafe/...` and `GOTOOLCHAIN=go1.26.10 go test -race ./...` pass.
* **Acceptance Criteria**:
  1. `pathsafe.ClassifyCLIPath` is exported from `internal/pathsafe/lexical.go` and deterministically classifies `CLIPathEmpty`, `CLIPathUNCOrDevice`, `CLIPathAbs`, `CLIPathDriveRelative`, `CLIPathRelative` (including `"."` and `".."`), and `CLIPathBare`.
  2. `CLIPathUNCOrDevice` precedes `filepath.IsAbs` and checks `isSlash` for both `'\\'` and `'/'` so `\\server\share`, `//server/share`, `\/server\share`, `/\server\share`, `\\?\`, `\\.\`, `\??\`, and `/??/` are classified as `CLIPathUNCOrDevice` on all platforms.

#### Unit 2.2: Harden `internal/config` `validateCLIPath` with `pathsafe.ClassifyCLIPath`, regular-file check, and single `LookPath` resolution (`082.002-T`)

* **Source Stash IDs**: `56CE7077`, `FD093EFC`
* **Depends On**: `082.001-T`
* **Priority**: `medium`
* **Size**: `S` (`size_source: agent`, `size_ruleset_version: 2h-rule-v1`)
* **Complexity**: `medium`
* **Skill Domain**: `code`
* **Execution Posture**: `test-first`
* **Files Affected (2)**:
  * `internal/config/validate.go`
  * `internal/config/copilotcli_test.go`
* **Functions Changed / Added (2)**:
  * `(*Config).Validate` in `internal/config/validate.go` (modified)
  * `validateCLIPath` in `internal/config/validate.go` (modified)
* **Test Scenarios (3)**:
  1. `Validate()` rejects UNC and device-namespace `copilot.cli_path` values (`\\server\share\copilot.exe`, `//server/share/copilot`, `\\?\C:\copilot.exe`, `\\.\pipe\copilot`, `\??\C:\copilot.exe`) with `"config: copilot.cli_path '<path>' must not be a UNC or device-namespace path; use an absolute path, a bare name, or empty"`.
  2. `Validate()` rejects an existing directory passed as an absolute `copilot.cli_path` (`t.TempDir()`) with `"config: copilot.cli_path '<dir>' must be a regular file"`, and accepts an existing regular file.
  3. For a bare `copilot.cli_path` resolvable on `PATH`, `Validate()` resolves it once via `exec.LookPath` + `filepath.Abs`, verifies the resolved target is a regular file and not a UNC/device path, and commits the resolved absolute path to `c.Copilot.CLIPath` only after all validation rules pass (leaving `c.Copilot.CLIPath` unmutated if a subsequent rule such as rule 7 or `[remote]` fails); an unresolvable bare name still appends the single warning `"copilot.cli_path '<name>' was not resolvable via PATH"` and leaves `c.Copilot.CLIPath` unchanged.
* **Detailed Changes**:
  1. In `internal/config/validate.go`:
     * Change `validateCLIPath(cliPath string, report *Report) (string, error)` to switch on `pathsafe.ClassifyCLIPath(cliPath)`:
       * `pathsafe.CLIPathEmpty`: return `"", nil`.
       * `pathsafe.CLIPathUNCOrDevice`: return `"", apperr.Newf(apperr.KindConfig, "copilot.cli_path '%s' must not be a UNC or device-namespace path; use an absolute path, a bare name, or empty", cliPath)`.
       * `pathsafe.CLIPathDriveRelative`: return `"", apperr.Newf(apperr.KindConfig, "copilot.cli_path '%s' must not be drive-relative; use an absolute path, a bare name, or empty", cliPath)`.
       * `pathsafe.CLIPathRelative`: return `"", apperr.Newf(apperr.KindConfig, "copilot.cli_path '%s' must not be relative; use an absolute path, a bare name, or empty", cliPath)`.
       * `pathsafe.CLIPathAbs`:
         * `info, err := os.Stat(cliPath)`; if `err != nil`, return `"", apperr.Newf(apperr.KindConfig, "copilot.cli_path '%s' does not exist", cliPath)`.
         * If `!info.Mode().IsRegular()`, return `"", apperr.Newf(apperr.KindConfig, "copilot.cli_path '%s' must be a regular file", cliPath)`.
         * Return `cliPath, nil`.
       * `pathsafe.CLIPathBare`:
         * `resolved, err := exec.LookPath(cliPath)`; if `err != nil`, append `fmt.Sprintf("copilot.cli_path '%s' was not resolvable via PATH", cliPath)` to `report.Warnings` and return `cliPath, nil`.
         * Convert `abs, err := filepath.Abs(resolved)`; if `err != nil`, return `"", apperr.Newf(apperr.KindConfig, "copilot.cli_path '%s' could not be resolved to an absolute path", cliPath)`.
         * If `pathsafe.ClassifyCLIPath(abs) != pathsafe.CLIPathAbs`, return `"", apperr.Newf(apperr.KindConfig, "copilot.cli_path '%s' must not resolve to a UNC or device-namespace path", cliPath)`.
         * `info, err := os.Stat(abs)`; if `err != nil || !info.Mode().IsRegular()`, return `"", apperr.Newf(apperr.KindConfig, "copilot.cli_path '%s' must be a regular file", cliPath)`.
         * Return `abs, nil`.
     * Remove the now-redundant local `hasWindowsDrivePrefix` from `internal/config/validate.go`.
     * In `(*Config).Validate()`, store `resolvedCLIPath, err := validateCLIPath(c.Copilot.CLIPath, &report)` in a local variable and assign `c.Copilot.CLIPath = resolvedCLIPath` in the deferred commit block at the end of `Validate()`.
  2. In `internal/config/copilotcli_test.go`:
     * Add `TestValidateRule5RejectsUNCAndDeviceCLIPaths`, `TestValidateRule5RejectsDirectoryCLIPathAndAcceptsRegularFile`, and `TestValidateRule5BareCLIPathResolvesToAbsWithDeferredCommit` (creating a temporary `PATH` directory with executable regular-file fixtures `copilot` and `copilot.exe` so `exec.LookPath("copilot")` succeeds on both Windows and Linux).
* **Red-to-Green Verification Sequence (P-002 / P-004)**:
  * **Red phase**: Add the three new tests in `internal/config/copilotcli_test.go`; confirm `GOTOOLCHAIN=go1.26.10 go test ./internal/config -run 'TestValidateRule5RejectsUNCAndDeviceCLIPaths|TestValidateRule5RejectsDirectoryCLIPathAndAcceptsRegularFile|TestValidateRule5BareCLIPathResolvesToAbsWithDeferredCommit'` fails before `validate.go` is updated.
  * **Green phase**: Update `internal/config/validate.go` and confirm `GOTOOLCHAIN=go1.26.10 go test -race ./internal/config/...` and `GOTOOLCHAIN=go1.26.10 go test -race ./...` pass.
* **Acceptance Criteria**:
  1. `internal/config/validate.go` delegates lexical classification to `pathsafe.ClassifyCLIPath`, rejects `CLIPathUNCOrDevice`, requires `info.Mode().IsRegular()` for `CLIPathAbs`, and resolves `CLIPathBare` via `exec.LookPath` + `filepath.Abs` with deferred commit to `c.Copilot.CLIPath`.
  2. Failed validation of any rule after rule 5 leaves `c.Copilot.CLIPath` unmutated (Protected Invariant I6).

#### Unit 2.3: Harden `internal/agent/copilot` `CLIPath` validation and `LookPath` resolution using `pathsafe.ClassifyCLIPath` (`082.003-T`)

* **Source Stash IDs**: `56CE7077`, `FD093EFC`
* **Depends On**: `082.001-T`
* **Priority**: `medium`
* **Size**: `S` (`size_source: agent`, `size_ruleset_version: 2h-rule-v1`)
* **Complexity**: `medium`
* **Skill Domain**: `code`
* **Execution Posture**: `test-first`
* **Files Affected (2)**:
  * `internal/agent/copilot/adapter.go`
  * `internal/agent/copilot/adapter_test.go`
* **Functions Changed / Added (4)**:
  * `New` in `internal/agent/copilot/adapter.go` (modified)
  * `newWithClient` in `internal/agent/copilot/adapter.go` (modified)
  * `validateConfig` in `internal/agent/copilot/adapter.go` (modified)
  * `resolveCLIPath` in `internal/agent/copilot/adapter.go` (replaces `validateCLIPath` and `isASCIILetter`)
* **Test Scenarios (3)**:
  1. `copilot.New` rejects `CLIPathUNCOrDevice` (`\\server\share\copilot.exe`, `//server/share/copilot`, `\/server\share\copilot.exe`, `\\?\C:\copilot.exe`, `\\.\pipe\copilot`, `\??\C:\copilot.exe`), `CLIPathDriveRelative` (`C:copilot`), and `CLIPathRelative` (`.`, `..`, `./bin/copilot`, `..\x\copilot`, `bin/copilot`).
  2. `copilot.New` rejects an absolute `CLIPath` that does not exist or is a directory (`t.TempDir()`), and accepts an existing regular file.
  3. `copilot.New` resolves a bare `CLIPath` (`"copilot"`) once via `exec.LookPath` + `filepath.Abs` (verifying `CLIPathAbs` and `info.Mode().IsRegular()`), stores the resolved absolute path on `ad.cfg.CLIPath` (and passes it to `copilotsdk.StdioConnection{Path: cfg.CLIPath}`), and rejects an unresolvable bare name.
* **Detailed Changes**:
  1. In `internal/agent/copilot/adapter.go`:
     * Import `"os"`, `"os/exec"`, and `"github.com/softwaresalt/intercom-go/internal/pathsafe"`.
     * Replace `validateCLIPath(p string) error` and `isASCIILetter` with `resolveCLIPath(p string) (string, error)` switching on `pathsafe.ClassifyCLIPath(p)`:
       * `pathsafe.CLIPathEmpty`: return `"", nil`.
       * `pathsafe.CLIPathUNCOrDevice`: return `"", fmt.Errorf("agent: cli_path %q must not be a UNC or device-namespace path", p)`.
       * `pathsafe.CLIPathDriveRelative`: return `"", fmt.Errorf("agent: cli_path %q must not be drive-relative", p)`.
       * `pathsafe.CLIPathRelative`: return `"", fmt.Errorf("agent: cli_path %q must not be relative", p)`.
       * `pathsafe.CLIPathAbs`:
         * `info, err := os.Stat(p)`; if `err != nil`, return `"", fmt.Errorf("agent: cli_path %q does not exist: %w", p, err)`.
         * If `!info.Mode().IsRegular()`, return `"", fmt.Errorf("agent: cli_path %q must be a regular file", p)`.
         * Return `p, nil`.
       * `pathsafe.CLIPathBare`:
         * `resolved, err := exec.LookPath(p)`; if `err != nil`, return `"", fmt.Errorf("agent: cli_path %q not found in PATH: %w", p, err)`.
         * `abs, err := filepath.Abs(resolved)`; if `err != nil`, return `"", fmt.Errorf("agent: resolve cli_path %q: %w", p, err)`.
         * If `pathsafe.ClassifyCLIPath(abs) != pathsafe.CLIPathAbs`, return `"", fmt.Errorf("agent: cli_path %q must not resolve to a UNC or device-namespace path", p)`.
         * `info, err := os.Stat(abs)`; if `err != nil`, return `"", fmt.Errorf("agent: stat resolved cli_path %q: %w", p, err)`.
         * If `!info.Mode().IsRegular()`, return `"", fmt.Errorf("agent: cli_path %q must resolve to a regular file", p)`.
         * Return `abs, nil`.
     * Change `validateConfig(cfg *Config) error` to take `*Config`, call `resolved, err := resolveCLIPath(cfg.CLIPath)`, and assign `cfg.CLIPath = resolved` on success; update `New(cfg Config)` and `newWithClient(cfg Config, c sdkClient)` to pass `&cfg` to `validateConfig` so both `opts.Connection = copilotsdk.StdioConnection{Path: cfg.CLIPath}` and `newAdapter(cfg, ...)` receive the resolved absolute path.
  2. In `internal/agent/copilot/adapter_test.go`:
     * Update `absTestPath(t *testing.T) string` to create a regular executable file inside `t.TempDir()` (`0o755`), and in `TestNewValidatesConfig` configure a temporary `PATH` directory containing regular executable files `copilot` and `copilot.exe` (`0o755`) via `t.Setenv("PATH", ...)` for the `"bare name"` case (asserting `filepath.IsAbs(ad.cfg.CLIPath)`), and add cases for UNC/device paths, non-existent absolute path, directory path, and unresolvable bare name.
* **Red-to-Green Verification Sequence (P-002 / P-004)**:
  * **Red phase**: Update `TestNewValidatesConfig` in `internal/agent/copilot/adapter_test.go` with the UNC/device, directory, non-existent absolute, and bare-name-to-absolute assertions; confirm `GOTOOLCHAIN=go1.26.10 go test ./internal/agent/copilot -run TestNewValidatesConfig` fails before `adapter.go` is updated.
  * **Green phase**: Update `internal/agent/copilot/adapter.go` and confirm `GOTOOLCHAIN=go1.26.10 go test -race ./internal/agent/...` and `GOTOOLCHAIN=go1.26.10 go test -race ./...` pass.
* **Acceptance Criteria**:
  1. `internal/agent/copilot/adapter.go` uses `pathsafe.ClassifyCLIPath` (eliminating duplicate classifier logic), rejects UNC and device-namespace paths, requires `os.Stat` + `info.Mode().IsRegular()` for absolute paths, and resolves bare names once via `exec.LookPath` + `filepath.Abs` into `cfg.CLIPath` and `StdioConnection.Path`.
  2. When `config.Validate()` has already resolved a bare `cli_path` to an absolute path, `copilot.New` sees `CLIPathAbs` and does not invoke `exec.LookPath` a second time.

#### Unit 2.4: Reject unknown keys under `[remote]`, reject hex IPv4 host forms, and lock `DefaultRemoteMaxConnections` (`082.004-T`)

* **Source Stash ID**: `FC76C5D8` (originating refs: `066.006-T`, `066-F`, `058-S`, `PR #177`, `review-thread: N/A`)
* **Priority**: `low`
* **Size**: `S` (`size_source: agent`, `size_ruleset_version: 2h-rule-v1`)
* **Complexity**: `medium`
* **Skill Domain**: `code`
* **Execution Posture**: `test-first`
* **Files Affected (3)**:
  * `internal/config/remote.go`
  * `internal/config/default.go`
  * `internal/config/remote_test.go`
* **Functions Changed / Added (4)**:
  * `rejectRemoteCredentialKeys` in `internal/config/remote.go` (modified to also reject non-credential unknown keys under `[remote]` in a second pass)
  * `numericHostProblem` in `internal/config/remote.go` (modified)
  * `isIPv4NumberLabel` in `internal/config/remote.go` (added helper)
  * `Default` in `internal/config/default.go` (modified to use `DefaultRemoteMaxConnections`)
* **Test Scenarios (3)**:
  1. Any unknown key at or below `[remote]` (such as `[remote]\nfuture_option = 1` or `[remote.sub]\nsetting = "v"`) is rejected by `Decode` with `apperr.KindConfig` containing `"remote.future_option is an unknown key under [remote]; only known [remote] schema keys are permitted"`, never echoing the value, while still populating `report.UnknownKeys` (`["remote.future_option"]`) and keeping credential-shaped key errors higher priority when both appear in the same document.
  2. Hexadecimal and mixed hex/octal/decimal IPv4 host forms (`"0x7f.0x1"`, `"0x7f.0.0.0x1"`, `"127.0.0.0x1"`, `"0x7f000001"`, `"0x0"`) are rejected in both `Remote.AllowedHosts` (with and without `:port`) and `Remote.AllowedOrigins` with `"non-canonical IPv4 address"`, while valid DNS labels ending in non-hex-prefixed alphanumeric labels (`"1password.example"`, `"example.123abc"`, `"a1.example.com"`) remain accepted.
  3. `config.DefaultRemoteMaxConnections` (`8`, used in `config.Default().Remote.MaxConnections`) equals `hub.DefaultMaxRemote` (`8`), and `config.MaxRemoteConnections` (`1024`) equals `hub.MaxRemoteHardLimit` (`1024`), verified via a test-scope import of `internal/hub` in `internal/config/remote_test.go`.
* **Detailed Changes**:
  1. In `internal/config/remote.go`:
     * Export `const DefaultRemoteMaxConnections = 8` (and document that it matches `Default().Remote.MaxConnections` and `hub.DefaultMaxRemote`).
     * In `rejectRemoteCredentialKeys(keys []toml.Key) error`:
       * Pass 1 (unchanged): scan every key `k` at or below `[remote]` (`len(k) >= 2 && strings.EqualFold(k[0], "remote")`, skipping `len(k) == 2 && isKnownRemoteKey(k[1])`) and if any segment in `k[1:]` satisfies `isCredentialShapedKey(seg)`, return `apperr.Newf(apperr.KindConfig, "%s is a credential-shaped key; credentials must never be stored in config.toml (remove it from [remote])", sanitizeKeyPath(k.String()))`.
       * Pass 2: scan every key `k` at or below `[remote]` (`len(k) >= 2 && strings.EqualFold(k[0], "remote")`) and if `!(len(k) == 2 && isKnownRemoteKey(k[1]))`, return `apperr.Newf(apperr.KindConfig, "%s is an unknown key under [remote]; only known [remote] schema keys are permitted", sanitizeKeyPath(k.String()))`.
     * Add `isIPv4NumberLabel(label string) bool`:
       * Returns `false` if `label == ""`.
       * Returns `true` if `strings.Trim(label, "0123456789") == ""` (all decimal/octal digits).
       * Returns `true` if `strings.HasPrefix(label, "0x") && strings.Trim(label[2:], "0123456789abcdef") == ""` (WHATWG / `inet_aton` hex label).
       * Otherwise returns `false`.
     * In `numericHostProblem(h string) string`:
       * Set `last := h[strings.LastIndexByte(h, '.')+1:]` and `numeric := isIPv4NumberLabel(last)`.
       * Keep the rest of `numericHostProblem` (`if !numeric && strings.Trim(h, "0123456789.") != "" { return "" }`, then canonical `netip.ParseAddr(h)` check).
  2. In `internal/config/default.go`:
     * Replace `MaxConnections: 8` with `MaxConnections: DefaultRemoteMaxConnections`.
  3. In `internal/config/remote_test.go`:
     * Import `"github.com/softwaresalt/intercom-go/internal/hub"`.
     * Update `TestDecodeRemoteUnknownKeyStillReported` to assert that `Decode` returns an `apperr.KindConfig` error containing `"remote.future_option is an unknown key under [remote]"` (never echoing the value), while `report.UnknownKeys` still equals `[]string{"remote.future_option"}`, and verify that when both `future_option = 1` and `token = "secret"` are present under `[remote]`, the credential-shaped error wins.
     * In `TestDefaultRemoteIsSecure`, assert `r.MaxConnections == DefaultRemoteMaxConnections`, `DefaultRemoteMaxConnections == hub.DefaultMaxRemote`, and `MaxRemoteConnections == hub.MaxRemoteHardLimit`.
     * Add `TestValidateRemoteRejectsHexIPv4HostsAndOrigins` testing `"0x7f.0x1"`, `"0x7f.0.0.0x1"`, `"127.0.0.0x1"`, `"0x7f000001"`, and `"0x0"` across `AllowedHosts` (bare and `:3000`) and `AllowedOrigins`.
* **Red-to-Green Verification Sequence (P-002 / P-004)**:
  * **Red phase**: Update `internal/config/remote_test.go` with the unknown-`[remote]`-key, hex-IPv4, and `DefaultRemoteMaxConnections == hub.DefaultMaxRemote` test cases; confirm `GOTOOLCHAIN=go1.26.10 go test ./internal/config -run 'TestDecodeRemoteUnknownKeyStillReported|TestValidateRemoteRejectsHexIPv4HostsAndOrigins|TestDefaultRemoteIsSecure'` fails before `remote.go` and `default.go` are updated.
  * **Green phase**: Update `internal/config/remote.go` and `internal/config/default.go` and confirm `GOTOOLCHAIN=go1.26.10 go test -race ./internal/config/...` and `GOTOOLCHAIN=go1.26.10 go test -race ./...` pass.
* **Acceptance Criteria**:
  1. `Decode` rejects any unknown key at or below `[remote]` with `"<key> is an unknown key under [remote]; only known [remote] schema keys are permitted"`, without echoing values and while preserving credential-shaped error precedence and `report.UnknownKeys` population.
  2. `numericHostProblem` rejects hex IPv4 host forms (`0x7f.0x1`, `0x7f.0.0.0x1`, `127.0.0.0x1`, `0x7f000001`, `0x0`) in both `allowed_hosts` and `allowed_origins`.
  3. `DefaultRemoteMaxConnections` (`8`) is exported in `internal/config`, used in `config.Default().Remote.MaxConnections`, and locked by test against `hub.DefaultMaxRemote` (`8`) (alongside `MaxRemoteConnections == hub.MaxRemoteHardLimit`).

#### Unit 2.5: Update backend architecture design doc and config reference for `cli_path` and `[remote]` validation (`082.005-T`)

* **Source Stash IDs**: `56CE7077`, `FD093EFC`, `FC76C5D8`
* **Depends On**: `082.002-T`, `082.003-T`, `082.004-T`
* **Priority**: `low`
* **Size**: `S` (`size_source: agent`, `size_ruleset_version: 2h-rule-v1`)
* **Complexity**: `low`
* **Skill Domain**: `docs`
* **Execution Posture**: `characterization-first`
* **Files Affected (2)**:
  * `docs/design-docs/intercom-go-backend-architecture.md`
  * `docs/config-reference.md`
* **Functions Changed / Added (0)**: Documentation-only unit.
* **Test Scenarios (2)**:
  1. `GOTOOLCHAIN=go1.26.10 go test ./internal/config -run 'TestConfigReference'` passes with the updated `docs/config-reference.md`.
  2. `GOTOOLCHAIN=go1.26.10 go run ./tools/gatecheck retired-arch --root .` and `GOTOOLCHAIN=go1.26.10 go run ./tools/gatecheck write-path --self-test --root .` pass with zero findings.
* **Detailed Changes**:
  1. In `docs/design-docs/intercom-go-backend-architecture.md`:
     * Amend §6.2 (`Copilot CLI resolution is a security-relevant path`) to document the shared `pathsafe.ClassifyCLIPath` predicate:
       * `cli_path == ""` -> valid (SDK resolves via `PATH` / `COPILOT_CLI_PATH` / embedded CLI)
       * starts with `\\`, `//`, or `\??\` -> ERROR (UNC or Windows device-namespace path rejected on every `GOOS`)
       * `filepath.IsAbs(cli_path)` -> valid iff the path exists and is a regular file (`info.Mode().IsRegular()`), else ERROR
       * matches `^[A-Za-z]:` and not absolute -> ERROR (Windows drive-relative)
       * contains `'/'` or `'\\'` and not absolute -> ERROR (relative path)
       * otherwise (bare name) -> resolved once via `exec.LookPath` to an absolute regular-file path (with a non-fatal `Report.Warnings` advisory at `config.Validate()` time if not currently on `PATH`, and fail-closed rejection at `copilot.New` if still unresolvable at adapter construction).
     * Amend §8 (`New [remote] validation rule group`) to note that `[remote]` is a closed section (any unknown key under `[remote]` is rejected in `Decode`) and `allowed_hosts`/`allowed_origins` reject hex (`0x7f.0x1`) as well as octal/abbreviated IPv4 forms.
  2. In `docs/config-reference.md`:
     * Update the `[copilot]` table, rule-5 validation messages, and rule-5 classifier block to include the UNC/device rejection message (`copilot.cli_path '{path}' must not be a UNC or device-namespace path; use an absolute path, a bare name, or empty`), the regular-file message (`copilot.cli_path '{path}' must be a regular file`), and single `LookPath` resolution of bare names (while retaining all existing rule-5 message substrings required by `TestConfigReferenceCoversValidationMessages`).
     * Update the `[remote]` section and rule table to document unknown-key rejection under `[remote]` (`remote.{key} is an unknown key under [remote]; only known [remote] schema keys are permitted`) and hex IPv4 host rejection (`0x7f.0x1`).
* **Acceptance Criteria**:
  1. `docs/design-docs/intercom-go-backend-architecture.md` (§6.2, §8) and `docs/config-reference.md` accurately reflect the shared `pathsafe.ClassifyCLIPath` contract, UNC/device and non-regular-file rejection, single `LookPath` bare-name resolution, unknown-key rejection under `[remote]`, and hex IPv4 host rejection.
  2. `GOTOOLCHAIN=go1.26.10 go test -race ./internal/config/...` passes.

## Dependency Graph

### Shipment-Level DAG

* `073-S` (`shipped`)
  * `<-blocks-- 074-S` (Covering Feature `081-F`: Write-permission preview safety, shell URL anti-spoofing, and unknown-kind local-TUI authorization [`773B2CB2`, `70315264`])
    * `<-blocks-- 075-S` (Covering Feature `082-F`: Copilot `cli_path` classification/resolution hardening and `[remote]` config validation [`56CE7077`, `FD093EFC`, `FC76C5D8`])

### Intra-Shipment Task Dependencies

* **Within `074-S` (`081-F`)**:
  * `081.001-T` (`internal/agent`, `internal/agent/copilot`) — root task
  * `081.002-T` (`internal/hub/authz.go`, `internal/hub/envelope.go`, `internal/hub/authz_test.go`) — root task
  * `081.003-T` (`internal/hub/permission.go`, `internal/hub/hub.go`, `internal/hub/hydrate_permission_test.go`) — `blocks`-depends on `081.001-T` and `081.002-T`
  * `081.004-T` (`internal/tui/permission.go`, `internal/tui/permission_test.go`, `internal/hub/doc.go`) — `blocks`-depends on `081.002-T` and `081.003-T`
  * `081.005-T` (`docs/`) — `blocks`-depends on `081.002-T`, `081.003-T`, and `081.004-T`
* **Within `075-S` (`082-F`)**:
  * `082.001-T` (`internal/pathsafe/lexical.go`, `internal/pathsafe/lexical_test.go`) — root task
  * `082.002-T` (`internal/config/validate.go`, `internal/config/copilotcli_test.go`) — `blocks`-depends on `082.001-T`
  * `082.003-T` (`internal/agent/copilot/adapter.go`, `internal/agent/copilot/adapter_test.go`) — `blocks`-depends on `082.001-T`
  * `082.004-T` (`internal/config/remote.go`, `internal/config/default.go`, `internal/config/remote_test.go`) — root task (independent of `082.001-T`..`082.003-T`)
  * `082.005-T` (`docs/`) — `blocks`-depends on `082.002-T`, `082.003-T`, and `082.004-T`

## Decisions and Rationale

1. **Two single-feature shipments (`074-S` and `075-S`)**: Satisfies Ship Step 6.1(a1) `S2` (`n == 1` covering feature per shipment) and cleanly separates runtime permission/TUI modal changes (`074-S`) from `cli_path` and `[remote]` configuration hardening (`075-S`).
2. **Structured `PreviewClipped` + column-0 TUI warning + `APPROVE CLIPPED WRITE?` confirmation (`773B2CB2`)**: Preserves the ability to review and approve legitimate multi-KB file writes while eliminating reliance on the spoofable in-band `"[write preview truncated]"` text alone. Because all `p.Summary` lines in `permissionBody` are indented by 2 spaces, the column-0 `warning: write preview clipped — full content not shown` header cannot be forged by file content.
3. **Unknown permission kinds are local-TUI-only even with `remote_approve_destructive = true` (`70315264(a)`)**: Unmodeled SDK permission kinds (`PermissionKindOther`, `PermissionKindUnknown`, undefined values) carry only `RawKind` in `Summary` without tool arguments. Refusing remote `DecisionAllowOnce` on unknown kinds with `ReasonDestructiveRemoteRefused` regardless of `remote_approve_destructive` enforces fail-closed least privilege while still allowing remote `DecisionDeny` and local TUI approval.
4. **Shared `pathsafe.ClassifyCLIPath` in `internal/pathsafe/lexical.go` (`56CE7077`, `FD093EFC`)**: Placing the lexical classifier in `internal/pathsafe` allows both `internal/config` and `internal/agent/copilot` to share one classifier without making `internal/agent/copilot` import `internal/config`, and checking `\\`, `//`, and `\??\` ahead of `filepath.IsAbs` rejects UNC and Windows device-namespace paths on every `GOOS`.

## Risks and Caveats

* **`writepath` gate (`tools/gatecheck`)**: Production files in `cmd/**` and `internal/**` must not call forbidden filesystem write selectors (`os.WriteFile`, `os.Create`, `os.OpenFile`, `os.Mkdir`, `os.MkdirAll`, `os.Remove`, `os.RemoveAll`, `os.Rename`, `os.Symlink`, `os.Link`, `os.Chmod`, `os.Chown`, `os.Chtimes`, `os.Truncate`, `os.CreateTemp`, `os.MkdirTemp`) or add unreviewed `pathsafe.Root.Resolve` call sites. None of the production changes in `074-S` or `075-S` use any filesystem write selector (`os.Stat`, `exec.LookPath`, and `filepath.Abs` are read-only), and test fixtures in `_test.go` are exempt from the production-only `writepath` gate.
* **AST and structural guards**:
  * `TestHubPackageHasNoBareChannelSends` (`internal/hub/degraded_test.go`): Every channel send in `internal/hub/*.go` must remain inside a `select` statement or `trySend`.
  * `TestPackageHasNoDirectAgentOrSDKReach` (`internal/tui/model_test.go`): `internal/tui` non-test files must not import `copilot-sdk` or `internal/agent/copilot`.
* **Toolchain pin**: Use `GOTOOLCHAIN=go1.26.10` (or the workspace's `go.mod` toolchain) when running `go test`, `go vet`, and `golangci-lint`.

## Plan Hardening Signals

* **public API, schema, or contract change**: **PRESENT** — Adds `agent.PermissionRequest.PreviewClipped`, `hub.PermissionRequestedPayload.PreviewClipped` (`json:"preview_clipped,omitempty"`), `pathsafe.CLIPathKind` and `pathsafe.ClassifyCLIPath`, `config.DefaultRemoteMaxConnections`, amends `ADR MVP-3` so unknown permission kinds are local-TUI-only even when `remote_approve_destructive = true`, amends design §6.2 to reject UNC/device `cli_path` values and require regular files, and closes `[remote]` against unknown TOML keys and hex IPv4 hosts.
* **security, auth, permission, or compliance-sensitive behavior**: **PRESENT** — Modifies hub permission authorization (`internal/hub/authz.go`), TUI permission modal warning/confirmation rendering (`internal/tui/permission.go`), shell URL anti-spoofing (`internal/agent/copilot/permission.go`), Copilot CLI binary path validation (`internal/pathsafe`, `internal/config`, `internal/agent/copilot`), and `[remote]` config validation (`internal/config/remote.go`).
* **migration, backfill, destructive data/config action, or irreversible step**: **ABSENT** — No data migration or destructive action.
* **external integration, operator checkpoint, or external dependency**: **ABSENT** — No new external dependencies in `go.mod`.
* **high runtime, rollout, or rollback risk**: **PRESENT** — Touches runtime permission arbitration, TUI modal rendering, and adapter startup `cli_path` resolution.

Requires plan hardening: yes

## Runtime Verification and Closure

* **Shipment `074-S` (`081-F`)**: Modifies `internal/agent`, `internal/agent/copilot`, `internal/hub`, and `internal/tui` (runtime surfaces: `cli`, `api`).
  * **Automated Verification**: `GOTOOLCHAIN=go1.26.10 go test -race ./internal/agent/... ./internal/hub/... ./internal/tui/... ./internal/app/... ./internal/config/...`, `GOTOOLCHAIN=go1.26.10 go vet ./...`, `gofmt -l .`, `GOTOOLCHAIN=go1.26.10 go run ./tools/gatecheck retired-arch --root .`, and `GOTOOLCHAIN=go1.26.10 go run ./tools/gatecheck write-path --self-test --root .`.
  * **Operational Closure**: Post-merge closure artifact (`docs/closure/074-S-081-F-post-merge-closure.md`) with ConPTY/loopback live verification of the release-required manual checkpoints (`tui-render`, `devtunnel-connect`, `tool-approval-routing`) and mandatory P-020 `compact-context` invocation. Rollback is a clean `git revert` of the `074-S` merge commit.
* **Shipment `075-S` (`082-F`)**: Modifies `internal/pathsafe`, `internal/config`, and `internal/agent/copilot` (runtime surface: `cli` startup config and adapter initialization).
  * **Automated Verification**: `GOTOOLCHAIN=go1.26.10 go test -race ./internal/pathsafe/... ./internal/config/... ./internal/agent/... ./internal/app/...`, `GOTOOLCHAIN=go1.26.10 go vet ./...`, `gofmt -l .`, `GOTOOLCHAIN=go1.26.10 go run ./tools/gatecheck retired-arch --root .`, and `GOTOOLCHAIN=go1.26.10 go run ./tools/gatecheck write-path --self-test --root .`.
  * **Operational Closure**: Post-merge closure artifact (`docs/closure/075-S-082-F-post-merge-closure.md`) with live startup/checkpoint verification and mandatory P-020 `compact-context` invocation. Rollback is a clean `git revert` of the `075-S` merge commit.

## Plan Hardening

### 1. Why Hardening Was Required and Context Consulted

Hardening was required (`Requires plan hardening: yes`) because both shipments modify security-critical permission authorization, TUI approval modal rendering, Copilot CLI binary path trust, and `[remote]` configuration validation.

**Reinforced instructions and compound learnings consulted**:

* `.github/instructions/constitution.instructions.md` (Principles I, II, III, IV, V, VI, VII, VIII, XI)
* `.github/instructions/strict-safety.instructions.md` (`ProposedAction` / `ActionRisk` / `ActionResult` classification)
* `docs/compound/2026-10-20-tui-permission-modal-safety-and-linux-test-clock.md`
* `docs/compound/2026-10-20-process-wiring-credential-panel-and-approval-disarm.md`
* `docs/compound/2026-10-20-backlogit-cli-pitfalls-and-permission-outcome-lessons.md`
* `docs/compound/2026-10-19-hub-test-oracles-and-toolchain-pitfalls.md`
* `docs/compound/2026-09-08-verify-go-stdlib-internals-claims-against-goroot-source.md`
* `docs/compound/2026-09-30-go-msys-bash-bridging-and-unicode-case-mapping-gap.md`
* `docs/compound/workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md`

### 2. Protected Invariants

1. **Unclipped TUI permission modal golden stability (`internal/tui/testdata/modal.golden`)**: When `PreviewClipped` is `false`, `permissionBody` and `modalView` produce the exact existing output, and `modalBodyRows()` (`m.areaHeight() - 6`) and the top-of-`onKey` armed-approval disarm choke point remain untouched.
2. **Column-0 header non-spoofability in `internal/tui/permission.go`**: Every line of `p.Summary` is indented by 2 spaces (`"  " + l`), while `kind: `, `tool call: `, and (when `p.PreviewClipped` is `true`) `warning: write preview clipped — full content not shown` are rendered at column 0 via `wrapHeader` (whose continuation lines are also indented by 2 spaces), so summary content can never forge a column-0 header.
3. **Unknown permission kinds are local-TUI-only (`internal/hub/authz.go`)**: `PermissionKindUnknown`, `PermissionKindOther`, and any undefined `PermissionKind` value are always refused for remote `DecisionAllowOnce` with `ReasonDestructiveRemoteRefused` even when `RemoteApproveDestructive` is `true`, while remote `DecisionDeny` and local TUI approvals remain permitted.
4. **No bare channel sends in `internal/hub` (`TestHubPackageHasNoBareChannelSends`) and no SDK imports outside `internal/agent/copilot` (`copilot-sdk-boundary` depguard & `TestPackageHasNoDirectAgentOrSDKReach`)**.
5. **Zero new production filesystem write selectors and zero new `pathsafe.Root.Resolve` call sites (`tools/gatecheck write-path`)**: Production files use only read-only `os.Stat`, `exec.LookPath`, and `filepath.Abs`.
6. **Config validation receiver non-mutation on failure (Decision R10, Protected Invariant I6)**: `c.Copilot.CLIPath = resolvedCLIPath` is committed only in the deferred commit block at the end of `(*Config).Validate()` after all rules pass.
7. **Secret-free config error messages**: Unknown-key and credential-shaped-key errors under `[remote]` report only `sanitizeKeyPath(k.String())` and never echo TOML values.

### 3. Strict-Safety Action Classifications (`ProposedAction` / `ActionRisk` / `ActionResult`)

| Action ID | `ProposedAction` Summary & Targets | `change_kind` | `ActionRisk` | `approval_required` | Rollback | `ActionResult` |
|---|---|---|---|---|---|---|
| `PA-01` | Propagate `PreviewClipped` across `agent.PermissionRequest`, `hub.PermissionRequestedPayload` (live + `session.hydrate`), and `tui.permission` with a non-spoofable column-0 `warning:` header and `APPROVE CLIPPED WRITE?` prompt; sanitize `shellSummary` URLs and indent `urls:` lines in `FullCommandText`; enforce local-TUI-only approval for unknown permission kinds in `internal/hub/authz.go`; and amend `ADR MVP-3`, `intercom-go-backend-architecture.md`, and `docs/config-reference.md` (`074-S` / `081-F`). | `local edit` (agent, hub, TUI, and docs permission safety hardening) | `high` | `false` (pre-approved under `DARK_MODE_ACTIVE` Option 1 scope) | `git revert` of `074-S` merge commit | `planned` |
| `PA-02` | Add `pathsafe.ClassifyCLIPath` in `internal/pathsafe/lexical.go`; harden `internal/config` and `internal/agent/copilot` `cli_path` handling to reject UNC/device paths, require regular files (`info.Mode().IsRegular()`), and resolve bare names once via `exec.LookPath` + `filepath.Abs`; reject unknown keys under `[remote]` and hex IPv4 host forms (`0x7f.0x1`) in `internal/config/remote.go`; and update `intercom-go-backend-architecture.md` and `docs/config-reference.md` (`075-S` / `082-F`). | `local edit` (pathsafe, config, copilot adapter, and docs hardening) | `high` | `false` (pre-approved under `DARK_MODE_ACTIVE` Option 1 scope) | `git revert` of `075-S` merge commit | `planned` |

### 4. Review-Gate Requirements (P-012)

* Plan review (Step 4) MUST spawn the always-on and triggered cross-model persona subagents (`Constitution Reviewer`, `Go Reviewer`, `Scope Boundary Auditor`, `Learnings Researcher`, `Architecture Strategist`, and `Security Lens Reviewer`), routing `Architecture Strategist` through `model_routing.anchor_review` (`gpt-6.1-sol` / `high` when available), and append a `## Plan Review` section with literal `dispatch_mode:` and `decision:` markers before harvest.
* Unresolved operator decisions: None.

<!-- plan-review-attempt: 1 -->

## Plan Review

```text
dispatch_mode: multi-agent
decision: PASS
```

### Gate Decision and Rationale

* **Verdict**: **PASS** (after incorporating all P2 and P3 refinements from the 6-persona multi-agent review pass directly into the plan and deliberation artifacts).
* **Plan Hardening Requirement**: `Requires plan hardening: yes` — **SATISFIED** (`## Plan Hardening` section present with `ProposedAction` / `ActionRisk` / `ActionResult` classifications `PA-01` and `PA-02` and protected invariants).

### Persona Coverage Table

| Persona | Mode | Model Route | Initial Verdict | Final Verdict (Post-Refinement) |
|---|---|---|---|---|
| **Constitution Reviewer** | Subagent | Default (`gemini-4-argon`) | PASS (0 P0, 0 P1, 0 P2, 2 P3) | **PASS** |
| **Go Reviewer** | Subagent | Default (`gemini-4-argon`) | ADVISORY (0 P0, 0 P1, 2 P2, 1 P3) | **PASS** |
| **Scope Boundary Auditor** | Subagent | Default (`gemini-4-argon`) | ADVISORY (0 P0, 0 P1, 1 P2, 1 P3) | **PASS** |
| **Learnings Researcher** | Subagent | Default (`claude-haiku-5.5`) | ADVISORY (0 P0, 0 P1, 2 P2, 1 P3) | **PASS** |
| **Architecture Strategist** | Anchor Subagent | Anchor (`gpt-6.1-sol` / `high`) | ADVISORY (0 P0, 0 P1, 3 P2) | **PASS** |
| **Security Lens Reviewer** | Subagent | Default (`gemini-4-argon`) | ADVISORY (0 P0, 0 P1, 1 P2, 2 P3) | **PASS** |

### Findings and Applied Remediations

1. **[P2 — Remediated] Unit 2.1 (`082.001-T`) Mixed-Slash UNC/NT-Device Prefixes (`\/server\share`, `/\server\share`, `\??/`, `/??/`) and Bare `.` / `..` Relative Paths**:
   * *Finding* (`Go Reviewer`, `Security Lens Reviewer`): On Windows, `filepath.IsAbs` treats `/` and `\` as interchangeable (`isSlash`), so checking only `strings.HasPrefix(path, `\\`) || strings.HasPrefix(path, "//") || strings.HasPrefix(path, `\??\`)` allowed mixed-slash UNC/device paths (`\/server\share\copilot.exe`, `/??/C:/copilot.exe`) to reach `filepath.IsAbs` (`true`) and `os.Stat`; additionally, bare `"."` and `".."` contain no separator and fell through to `CLIPathBare`.
   * *Remediation applied*: Updated `hasUNCOrDevicePrefix` in Unit 2.1 (`082.001-T`) to check `isSlash(b)` (`b == '/' || b == '\\'`) for any two leading slashes and for the 4-byte `?` `?` NT device prefix, and updated step 5 of `ClassifyCLIPath` to classify `path == "." || path == ".."` as `CLIPathRelative`.
2. **[P2 — Remediated] Unit 2.4 (`082.004-T`) Cross-Package `DefaultRemoteMaxConnections == hub.DefaultMaxRemote` Parity Lock and `default.go` Wiring**:
   * *Finding* (`Constitution Reviewer`, `Go Reviewer`, `Scope Boundary Auditor`, `Learnings Researcher`, `Architecture Strategist`, `Security Lens Reviewer`): Unit 2.4 originally omitted `internal/config/default.go` and only compared `DefaultRemoteMaxConnections` to `Default().Remote.MaxConnections` within `internal/config` without asserting parity against `hub.DefaultMaxRemote` (`8`) and `hub.MaxRemoteHardLimit` (`1024`).
   * *Remediation applied*: Updated Unit 2.4 (`082.004-T`) to modify `internal/config/remote.go`, `internal/config/default.go` (`MaxConnections: DefaultRemoteMaxConnections`), and `internal/config/remote_test.go` (3 files total), importing `internal/hub` in `remote_test.go` to lock `DefaultRemoteMaxConnections == hub.DefaultMaxRemote` and `MaxRemoteConnections == hub.MaxRemoteHardLimit` alongside the unknown-`[remote]`-key and hex-IPv4 host tests.
3. **[P2 — Remediated] Units 1.1, 1.3, 1.4 (`081.001-T`, `081.003-T`, `081.004-T`) Package/Type GoDoc Contract Alignment and Repository-Wide `go test -race ./...` Verification**:
   * *Finding* (`Architecture Strategist`, `Learnings Researcher`): Package/type comments in `internal/agent/agent.go` (`PermissionRequest.Truncated`), `internal/hub/hub.go` (`Config.RemoteApproveDestructive`), `internal/hub/doc.go` (`# Permission arbitration`), and `internal/tui/permission.go` (`permission`) needed alignment with `PreviewClipped` and local-TUI-only unknown-kind approval, and every unit's Green phase should verify `go test -race ./...`.
   * *Remediation applied*: Added the GoDoc comment updates to `081.001-T`, `081.003-T`, and `081.004-T` (including `internal/hub/doc.go` as the 3rd `.go` file in `081.004-T`) and added `GOTOOLCHAIN=go1.26.10 go test -race ./...` to every `code` unit's Green phase.
4. **[P3 — Remediated] Units 1.1 & 2.3 (`081.001-T`, `082.003-T`) Unicode Space Trimming in `shellSummary` and `os.Stat` Error Wrapping in `resolveCLIPath`**:
   * *Finding* (`Security Lens Reviewer`, `Go Reviewer`, `Constitution Reviewer`): Used `strings.TrimLeftFunc(line, unicode.IsSpace)` in `shellSummary` (`081.001-T`) so non-ASCII Unicode whitespace (`U+00A0`) before `"urls:"` is indented; separated the `os.Stat(abs)` error branch (`%w`) from the `!info.Mode().IsRegular()` branch in `resolveCLIPath` (`082.003-T`) and updated its function count metadata to `4`.


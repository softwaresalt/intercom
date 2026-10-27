---
title: "MVP Post-062-S Follow-ups and Root.Resolve Black-Box Verification Implementation Plan"
description: "Implementation plan for Capability Set 1 across four chained single-feature shipments (063-S..066-S): 038-F black-box Root.Resolve verification, BDC0ADDF contract follow-ups, FAE4134F process-lifecycle hardening, and AD8472DD maintainability refactors."
source: "docs/decisions/2026-10-26-intercom-go-mvp-post-062-s-follow-ups-and-root-resolve-deliberation.md"
status: "ready"
date: 2026-10-26
tags:
  - "pathsafe"
  - "internal-app"
  - "process-lifecycle"
  - "tui"
  - "config"
---

## Problem Frame

Following the merge and closure of `062-S` / `070-F` (PR #185, commit `6290151f08fcb300b2bd787c15ce28382242b233`), the first production caller of `pathsafe.Root.Resolve` is live in `internal/app/workspace.go` (`resolveWorkspace`), unblocking `038-F` (sourced from archived stash `37FAB8C2`). In addition, three pre-PR deferred scope expansions from `062-S` (`BDC0ADDF`, `FAE4134F`, and `AD8472DD`) require targeted verification, config-contract, process-lifecycle, credential-hygiene, TUI-observability, and move-only maintainability improvements across `internal/app`, `internal/config`, `internal/hub`, `internal/transport`, `internal/tui`, and operator/decision documentation.

To satisfy Ship Step 6.1(a1) `S2` (`n == 1` covering feature per shipment) and keep behavioral changes cleanly separated from move-only file splits, the work is organized into four sequential single-feature shipments (`063-S -> 064-S -> 065-S -> 066-S`). Shipments `063-S`..`066-S` functionally depend on code merged in `062-S` (`6290151`, `internal/app/workspace.go`), whose post-merge closure (`docs/closure/062-S-070-F-post-merge-closure.md`) has `compaction_status: done` and `closure_status: READY_WITH_CONDITIONS` with two open operator-owned live-environment conditions (`live-sdk-test-pass` and `tool-approval-routing-live`). Under `.github/policies/workflow-policies.md` P-001 ("unless the operator explicitly overrides P-001"), the operator explicitly invoked dark factory mode on `2026-10-26` strictly scoped to Option 1 (`BDC0ADDF`, `FAE4134F`, `AD8472DD`, `038-F` — the post-`062-S` remediation chain) after reviewing the state assessment that highlighted `062-S`'s open `READY_WITH_CONDITIONS` status; accordingly, `063-S` records this explicit operator P-001 policy exception and uses `labels: ["dag-root"]` (`predecessor_source: "declared_root"`) while preserving `062-S` (`6290151`) as the documented code baseline.

## Requirements Trace

| Source Requirement | Stash / Queue Origin | Target Shipment & Feature | Implementation Unit(s) |
|---|---|---|---|
| Black-box verify `pathsafe.Root.Resolve` through its live production caller (`internal/app`) | `038-F` (`37FAB8C2`) | `063-S` / `038-F` | Unit 1.1 (`038.001-T`) |
| Annotate dated decision records that stated point-in-time zero production callers for `pathsafe.Root.Resolve` | `BDC0ADDF(g)` | `063-S` / `038-F` | Unit 1.2 (`038.002-T`) |
| Clarify `already_resolved` (during `permDecided` hand-off) vs `request_unknown` (after retirement) in `wsclient` and `hub` docs | `BDC0ADDF(a)` | `064-S` / `071-F` | Unit 2.1 (`071.001-T`) |
| Drive a deterministic late TUI approve action while holding a pending request in `TestE2EFirstResponder` | `BDC0ADDF(b)` | `064-S` / `071-F` | Unit 2.2 (`071.002-T`) |
| Enforce upper bound (`MaxApprovalTimeoutSeconds = 86400`) on `[timeouts].approval_seconds` in `internal/config` | `BDC0ADDF(c)` | `064-S` / `071-F` | Unit 2.3 (`071.003-T`), Unit 2.6 (`071.006-T`) |
| Replace PowerShell `Win32_Process` snapshot in `liveproc_test.go` with native Windows `CreateToolhelp32Snapshot` API | `BDC0ADDF(d)` | `064-S` / `071-F` | Unit 2.4 (`071.004-T`) |
| Harden `readInsideWorkspace` using `pathsafe.NewRoot`/`Resolve` and cross-reference `062-S` live verification evidence | `BDC0ADDF(e, f)` | `064-S` / `071-F` | Unit 2.5 (`071.005-T`), Unit 2.6 (`071.006-T`) |
| Honour `startCtx` in `startUI` ready-handshake, bound post-`Kill` wait in `localUI.stop`, and synchronize `watchUI` errors before `takeErrs` | `FAE4134F(c, f, h)` | `065-S` / `072-F` | Unit 3.1 (`072.001-T`) |
| Cancel `startCtx` unconditionally on first signal in `watchSignals`, call `hub.BeginShutdown` in failed-start `teardown`, and warn at startup for accepted-but-unconsumed config keys | `FAE4134F(d, f, i)` | `065-S` / `072-F` | Unit 3.2 (`072.002-T`) |
| Zero plaintext credentials in `App.creds` after terminal TUI delivery, avoid reopening dismissed panel on role-0 `tuiReissue` failure, and refuse re-issue after revoke/shutdown | `FAE4134F(e, g)` | `065-S` / `072-F` | Unit 3.3 (`072.003-T`) |
| Surface rejected WebSocket upgrade counts and last status/reason on the TUI status line | `FAE4134F(b)` | `065-S` / `072-F` | Unit 3.4 (`072.004-T`), Unit 3.5 (`072.005-T`) |
| Document ADR MVP-2 TUI-only / programmatic-headless credential decision, TUI security notices, and unconsumed-config warnings in operator runbook | `FAE4134F(a, b, i)` | `065-S` / `072-F` | Unit 3.6 (`072.006-T`) |
| Split `internal/app/assemble.go` into `credentials.go` and `late_agent.go`, adding `lateAgent.get()` and `App.watchUI` | `AD8472DD` | `066-S` / `073-F` | Unit 4.1 (`073.001-T`) |
| Group `App` shutdown fields into `shutdownState` and extract shared `reap()` helper for `teardown` and `supervise` | `AD8472DD` | `066-S` / `073-F` | Unit 4.2 (`073.002-T`) |
| Deduplicate `factory`/`fakeAgent` knobs via embedded `agentKnobs` in `helpers_test.go` and decouple `partialRig` from `credBytes` in `review1_test.go` | `AD8472DD` | `066-S` / `073-F` | Unit 4.3 (`073.003-T`) |

## Constitution Check

| Principle | Compliance | Notes |
|---|---|---|
| I. Safety-First Go | PASS | Explicit error wrapping (`apperr`), bounded waits, no `panic` in library code, `golangci-lint` and `go vet` clean under `GOTOOLCHAIN=go1.26.10`. |
| II. Test-First Development (NON-NEGOTIABLE) | PASS | Every code/test unit defines concrete red-phase failing assertions before green implementation; pure documentation units verify via `go test ./...` and markdown/doc-coupling tests. |
| III. Workspace Isolation and Security Boundaries | PASS | No new production `pathsafe.Root.Resolve` call sites and zero filesystem write primitives in `cmd/**` or `internal/**` (`writepath` gate remains 100% green). Plaintext credentials are zeroed in `App.creds` after TUI delivery (`!o.TUI.Disabled && !o.TUI.RetainPlaintext`) and on revocation. |
| IV. CLI Workspace Containment (NON-NEGOTIABLE) | PASS | All edits are strictly inside the repository working tree. |
| V. Structured Observability | PASS | Startup warnings for unconsumed config keys use `slog.Logger.Warn`; rejected upgrades surface on the TUI status line when stderr logs are discarded. |
| VI. Single Responsibility | PASS | Zero new external module dependencies (uses standard library `syscall` on Windows for process snapshots). |
| VII. Destructive Command Approval (NON-NEGOTIABLE) | PASS | No destructive commands; stash archival uses non-destructive `backlogit stash archive`. |
| VIII. Explicit Safety Modes | PASS | Freeze-scope mode active on the 4 chartered items (`038-F`, `BDC0ADDF`, `FAE4134F`, `AD8472DD`). |
| IX. Git-Friendly Persistence | PASS | All deliberation, plan, memory, and backlog artifacts use Markdown with YAML frontmatter. |
| X. Agent Context Efficiency | PASS | Targeted queries and file reads used throughout staging. |
| XI. Merge Commit History Preservation | PASS | All shipments will merge via `--merge` commits (P-009). |

## Implementation Units

### Sub-Epic 1: Shipment `063-S` — Covering Feature `038-F` (Black-box verify `Root.Resolve` once a live caller exists)

#### Unit 1.1 (`038.001-T`): Black-box integration verification of `pathsafe.Root.Resolve` through `app.Start`

* **Domain**: `tests` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (1 file):
  * `internal/app/workspace_blackbox_test.go` (new file in external test package `package app_test`)
* **Changes needed**:
  * Add `TestStartBlackBoxWorkspaceResolveContainment` and `TestStartBlackBoxWorkspaceMappedRootAndRefusals` in `package app_test`, driving the exported `app.Start(ctx, app.Options{...})` entrypoint with an injected `AgentFactory` that records `AgentDeps.WorkingDirectory`.
  * Verify that passing `Options.Workspace` as a nested subdirectory under `default_workspace_root` or under a `[[workspace]].path` mapping invokes `resolveWorkspace`'s `base.Resolve(rel)` path and delivers the canonicalized subdirectory to both `a.Workspace()` and `AgentDeps.WorkingDirectory`.
  * Verify that passing an in-root symlink pointing outside every authorized root, a `..` traversal escaping the root, a prefix-sibling directory (`root + "-evil"`), a nonexistent child path, or a regular file path fails closed at `app.Start` with `apperr.KindPathViolation` before `NewAgent` is invoked (`invoked == 0`).
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/app -run TestStartBlackBoxWorkspace`
  * `./gatecheck.exe write-path --self-test` exits 0 (confirming `reviewedResolveCallers` count in `internal/app/workspace.go` remains 1).

#### Unit 1.2 (`038.002-T`): Annotate dated decision records with point-in-time `Root.Resolve` caller note (`BDC0ADDF(g)`)

* **Domain**: `docs` (`size: XS`, `complexity: low`)
* **Posture**: `migration-first` (documentation gardening)
* **Files affected** (3 files):
  * `docs/decisions/2026-09-11-intercom-go-pathsafe-deferred-trigger-reassessment-deliberation.md`
  * `docs/decisions/2026-09-18-intercom-go-ship-contract-and-gate-reliability-deliberation.md`
  * `docs/decisions/2026-10-01-intercom-go-post-m4-stash-triage-and-go-replan-deliberation.md`
* **Changes needed**:
  * Add a concise historical point-in-time callout (`> [!NOTE]`) to each of the three deliberation records where "zero production `pathsafe.Root.Resolve` callers" is stated, recording that the zero-caller measurement was accurate at that date prior to `062-S` (`070.001-T`, PR #185, `internal/app/workspace.go`), when the first read-only production caller landed, fired the `049.005-T` tripwire, and was reviewed in `internal/pathsafe/root.go` and `tools/gatecheck/internal/writepath/writepath.go` (`reviewedResolveCallers`).
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test ./...` passes.

---

### Sub-Epic 2: Shipment `064-S` — Covering Feature `071-F` (Verification-surface and contract follow-ups (`BDC0ADDF`))

#### Unit 2.1 (`071.001-T`): Clarify `already_resolved` vs `request_unknown` lifecycle boundary in `wsclient` and `hub` package docs (`BDC0ADDF(a)`)

* **Domain**: `docs` (`size: XS`, `complexity: trivial`)
* **Posture**: `characterization-first`
* **Files affected** (2 files):
  * `internal/transport/wsclient/client.go`
  * `internal/hub/doc.go`
* **Changes needed**:
  * Update the package GoDoc in `internal/transport/wsclient/client.go` (under `# Frames the server sends` -> `Refusal reasons`) and `internal/hub/doc.go` to state explicitly that `already_resolved` (with `winner` naming the winning actor) is returned when a losing responder's `resolve_permission` arrives while the winning decision is in hand-off awaiting handler acknowledgement (`permDecided`, before `permission.resolved` is published), whereas `request_unknown` is returned once `permission.resolved` or `permission.aborted` has been published and retired the registry entry (or when the `request_id` was never registered).
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test ./internal/hub ./internal/transport/...` passes.

#### Unit 2.2 (`071.002-T`): Drive deterministic late TUI approve action while holding `perm.gate` in `TestE2EFirstResponder` (`BDC0ADDF(b)`)

* **Domain**: `tests` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (1 file):
  * `internal/app/e2e_test.go`
* **Changes needed**:
  * In `TestE2EFirstResponder` (`"remote wins, TUI loses"` subtest in `internal/app/e2e_test.go`), while `perm` is held via `r.askHeld(...)` (after `first.resolve(id, true)` wins the CAS and `perm.awaitDecision(t)` returns, before `perm.release()` is called), drive the TUI to arm (`r.press("y")`, `waitOutput(t, r.out, "APPROVE? press Enter to confirm")`) and confirm (`r.press("\r")`) its own late `OpResolvePermission` action against the real hub.
  * Assert via `waitOutput(t, r.out, "request already answered by remote-approver")` that the TUI receives the hub's `ReasonAlreadyResolved` refusal with `Winner: "remote-approver"` and drops the pending modal deterministically before `perm.release()` publishes `permission.resolved`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/app -run TestE2EFirstResponder` passes deterministically.

#### Unit 2.3 (`071.003-T`): Enforce `MaxApprovalTimeoutSeconds = 86400` upper bound on `[timeouts].approval_seconds` in `internal/config` (`BDC0ADDF(c)`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/config/config.go`
  * `internal/config/validate.go`
  * `internal/config/validate_test.go`
* **Changes needed**:
  * Define `const MaxApprovalTimeoutSeconds = 86400` (24 hours, matching `MaxRemoteTokenLifetimeSeconds`; `0` remains valid and selects the hub's 2-minute default) on `[timeouts].approval_seconds` in `internal/config/config.go`.
  * In `(*Config).Validate()` (`internal/config/validate.go`), after `validateRemote(c.Remote)` (preserving the fixed order of rules 1–7 and the `[remote]` group), reject `c.Timeouts.ApprovalSeconds > MaxApprovalTimeoutSeconds` with `apperr.Newf(apperr.KindConfig, "timeouts.approval_seconds must not exceed %d", MaxApprovalTimeoutSeconds)`.
  * Add unit tests in `internal/config/validate_test.go` verifying `0` and `86400` pass while `86401` fails with the exact `apperr.KindConfig` message and leaves the receiver unmutated.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race ./internal/config` passes.

#### Unit 2.4 (`071.004-T`): Replace PowerShell `Win32_Process` snapshot with native Windows `syscall.CreateToolhelp32Snapshot` + command-line lookup in `internal/app` (`BDC0ADDF(d)`)

* **Domain**: `tests` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/app/liveproc_test.go`
  * `internal/app/liveproc_windows_test.go` (new `//go:build windows` file)
  * `internal/app/liveproc_other_test.go` (new `//go:build !windows` file)
* **Changes needed**:
  * In `internal/app/liveproc_windows_test.go`, implement `listProcesses(ctx context.Context) ([]procEntry, error)` using standard library `syscall` (`if err := ctx.Err(); err != nil { return nil, err }`, `syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)`, `defer func() { _ = syscall.CloseHandle(snap) }()`, `var entry syscall.ProcessEntry32; entry.Size = uint32(unsafe.Sizeof(entry))`, `syscall.Process32First(snap, &entry)`, `syscall.Process32Next(snap, &entry)` terminating cleanly on `errors.Is(err, syscall.ERROR_NO_MORE_FILES)`, extracting `int(entry.ProcessID)`, `int(entry.ParentProcessID)`, and `syscall.UTF16ToString(entry.ExeFile[:])`), and query each candidate process's full command line via native Win32 APIs (`syscall.OpenProcess(PROCESS_QUERY_LIMITED_INFORMATION|PROCESS_VM_READ, false, entry.ProcessID)` + `ntdll` `NtQueryInformationProcess` (`ProcessBasicInformation` = 0) reading `PEB.ProcessParameters.CommandLine` via `kernel32` `ReadProcessMemory`, falling back to `syscall.UTF16ToString(entry.ExeFile[:])` only when the process has already exited or access is denied), eliminating the flaky `powershell.exe Get-CimInstance Win32_Process` subprocess on Windows with zero `go.mod` changes while preserving `newChildren`'s command-line needle matching (`strings.Contains(strings.ToLower(p.cmd), strings.ToLower(needle))`) after an intermediate launcher exits (e.g. a reparented `node.exe ...\copilot\index.js` child).
  * In `internal/app/liveproc_other_test.go`, retain the POSIX `ps -axo pid=,ppid=,args=` implementation of `listProcesses(ctx context.Context) ([]procEntry, error)` and move `parseProcessTable(s string) []procEntry` (without the obsolete `pipeSeparated bool` parameter) into `liveproc_other_test.go` so `golangci-lint` (`unused` and `unparam`) passes on both Windows and Linux.
  * Keep shared `procEntry`, `descendants`, `pidSet`, `newChildren`, `executableName`, and `TestProcessTreeSeesAndForgetsAChild` in `internal/app/liveproc_test.go`, and add an explicit regression test in `internal/app/liveproc_test.go` that spawns a short-lived intermediate launcher process which spawns a grandchild whose executable name is not the needle (`os.Args[0]`) but whose command-line arguments contain a sentinel token (e.g. `-intercom-sentinel=copilot-orphan-check`), waits for the intermediate launcher to exit so the grandchild is orphaned (`ppid` no longer in the live tree under `os.Getpid()`), and asserts `listProcesses` + `newChildren(table, os.Getpid(), baseline, "copilot-orphan-check")` still detects the surviving orphaned grandchild via its command line before reaping it.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -count=10 ./internal/app -run TestProcessTree` passes rapidly with zero subprocess flakiness.

#### Unit 2.5 (`071.005-T`): Harden `readInsideWorkspace` and snapshot pre-shutdown credentials in `internal/app/live_test.go` (`BDC0ADDF(e)`)

* **Domain**: `tests` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (1 file):
  * `internal/app/live_test.go`
* **Changes needed**:
  * Update `readInsideWorkspace(ws, summary string) bool` in `internal/app/live_test.go` without using `filepath.EvalSymlinks` and **without mutating `summary` with `strings.TrimSpace` or stripping quotes or `file://` URL prefixes** (in `internal/agent/copilot/permission.go`, `Summary` for `KindRead` is populated directly from raw `r.Path` via `sanitizeBound` which preserves ordinary spaces and replaces control/format runes with `U+FFFD`, so trimming whitespace, stripping quotes or `file://`, cleaning `..` before symlink resolution, or accepting lossy `U+FFFD`/truncated summaries creates a mismatch between the path `readInsideWorkspace` validates and the literal path the SDK runtime opens — e.g., on POSIX `" notes.txt "`, `'notes.txt'`, `notes\x1b.txt`, or `link/../notes.txt` where `link -> <outside>/sub` can open an outside file while `notes.txt` or `notes\uFFFD.txt` is an inside file): reject empty strings (`summary == ""`), **reject (`return false`) any path that changes when trimmed (`strings.TrimSpace(summary) != summary`, i.e., any leading or trailing whitespace as well as whitespace-only strings)**, reject NUL bytes (`\x00`), **reject (`return false`) any `summary` containing invalid UTF-8 (`!utf8.ValidString(summary)`), the Unicode replacement character `'\uFFFD'` (`strings.ContainsRune(summary, '\uFFFD')`), any Unicode control (`unicode.IsControl(r)`) or format (`unicode.Is(unicode.Cf, r)`) rune, or exceeding the adapter's `maxSummaryRunes = 2048` truncation bound or carrying its ellipsis marker (`utf8.RuneCountInString(summary) > 2048 || strings.HasSuffix(summary, "…")` — matching `sanitizeBound(text, 2048)` in `internal/agent/copilot/permission.go:66,125`, which copies up to `2048` runes and appends `"…"` (`U+2026`, yielding `2049` runes) only when `fullLen > 2048`, while allowing valid untruncated contained paths of 512–2048 runes)**, **reject (`return false`) any path containing a literal `".."` path component before any `filepath.Clean`, `filepath.Dir`, `filepath.Rel`, or `root.Resolve` call** (splitting the raw path on both `'/'` and `'\\'` before any cleaning and returning `false` if any segment equals `".."`), **fail closed (`return false`) on any `file:` URL scheme prefix (`strings.HasPrefix(strings.ToLower(p), "file:")`)**, and preserve literal quote characters (`"..."` or `'...'`) untouched so `root.Resolve` evaluates the exact literal filesystem entry (e.g., `'notes.txt'`), never an unquoted or trimmed lookalike (`notes.txt`); perform only Windows extended-length prefix normalization before the `filepath.IsAbs` / `filepath.Rel` split (following `stripUNCPrefix` in `internal/pathsafe/root.go`): check `strings.HasPrefix(p, `\\?\`) && len(p) >= 8 && strings.EqualFold(p[4:8], `UNC\`)` FIRST and rewrite it to `\\` + `p[8:]`, reject `\\.\` device-namespace prefixes outright (`return false`), and only then strip `\\?\` for drive-letter paths (`\\?\C:\...` -> `C:\...`) before `filepath.IsAbs` / `filepath.Rel`; construct `root, err := pathsafe.NewRoot(ws)`; for an absolute candidate `p`, canonicalize its parent directory via `dirRoot, err := pathsafe.NewRoot(filepath.Dir(p))` (which uses `GetFinalPathNameByHandleW` on Windows), compute `rel, err := filepath.Rel(root.Path(), filepath.Join(dirRoot.Path(), filepath.Base(p)))`, and validate via `resolved, err := root.Resolve(rel)` plus `info, err := os.Stat(resolved)` (`err == nil && !info.IsDir()`); for a relative candidate `p`, validate directly via `resolved, err := root.Resolve(p)` plus `os.Stat(resolved)`.
  * In `TestLiveCopilotEndToEnd` (`internal/app/live_test.go`), snapshot `issued := a.Credentials()` before `sig <- os.Interrupt` and iterate over `issued` in the post-shutdown log-leak assertion so clearing `a.creds` on shutdown in `065-S` cannot turn that loop into a vacuous 0-iteration check.
  * Expand `TestLiveReadApprovalIsConfinedToWorkspace` in `internal/app/live_test.go` to cover relative `./notes.txt`, `\\?\`-prefixed Windows drive-letter paths, an explicit untruncated long-path regression case (a contained nested path inside `<workspace>` whose summary length is between 512 and 2048 runes, e.g., ~600 runes of nested subdirectories/file, asserting `readInsideWorkspace` returns `true`), symlink/traversal escapes, and six explicit lookalike-collision / truncation regression cases: (1) external-UNC/local-collision where `<workspace>/UNC/outside/share/secret.txt` exists locally inside the workspace and the request supplies `\\?\UNC\outside\share\secret.txt`, asserting `readInsideWorkspace` returns `false`; (2) `file:`-URL/local-collision where `<workspace>/outside/share/secret.txt` exists inside the workspace and the request supplies `file://outside/share/secret.txt` (or `file:///...`), asserting `readInsideWorkspace` returns `false`; (3) quoted-path/local-collision where unquoted `<workspace>/notes.txt` exists as a regular file inside the workspace while literal `'notes.txt'` (or `"notes.txt"`) is either absent or (on POSIX/symlink-capable hosts) a symlink pointing outside the workspace, asserting `readInsideWorkspace` returns `false` (never resolving the unquoted `notes.txt` lookalike); (4) whitespace-collision where unquoted/unpadded `<workspace>/notes.txt` exists as a regular file inside the workspace while the request supplies a leading/trailing-space path `" notes.txt "` (and `"notes.txt "`, or where literal `" notes.txt "` on POSIX/symlink-capable hosts is an outside-pointing symlink), asserting `readInsideWorkspace` returns `false` (never resolving the trimmed `<workspace>/notes.txt` lookalike); (5) dot-dot symlink collision for both relative (`link/../notes.txt`) and absolute (`filepath.Join(ws, "link") + "/../notes.txt"`) forms where `<ws>/notes.txt` exists inside the workspace and `<ws>/link` points outside (e.g., `<outside>/sub`), asserting `readInsideWorkspace` returns `false` for both; and (6) adapter-to-approver `sanitizeBound` `U+FFFD` and truncation negative cases where `<ws>/notes\uFFFD.txt` exists as a regular file inside the workspace and the request summary is `notes\uFFFD.txt` (the `sanitizeBound` output of an outside-pointing `notes\x1b.txt`), asserting `readInsideWorkspace` returns `false`, and where a summary exceeds `maxSummaryRunes = 2048` (`> 2048` runes / `2048` runes + `"…"` from `sanitizeBound`), asserting `readInsideWorkspace` returns `false`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/app -run TestLiveReadApprovalIsConfinedToWorkspace` passes.

#### Unit 2.6 (`071.006-T`): Update `docs/config-reference.md` and `docs/runbooks/live-sdk-tests.md` for `approval_seconds` bound and `062-S` live evidence (`BDC0ADDF(c, e, f)`)

* **Domain**: `docs` (`size: XS`, `complexity: low`)
* **Posture**: `migration-first`
* **Files affected** (2 files):
  * `docs/config-reference.md`
  * `docs/runbooks/live-sdk-tests.md`
* **Changes needed**:
  * Update `docs/config-reference.md` to document that `timeouts.approval_seconds` must be between `0` and `86400` (`timeouts.approval_seconds must not exceed 86400`).
  * Update `docs/runbooks/live-sdk-tests.md` to reflect that Windows process-table snapshots use `syscall.CreateToolhelp32Snapshot` (no PowerShell subprocess), document the confirmed `rpc.PermissionRequestRead.Path` -> `Summary` mapping verified in the 2026-10-20 live run, and cross-reference `docs/closure/evidence/062-S-live-verification-evidence.md` for the completed Dev Tunnel `--allow-anonymous false` and `Authorization` header pass-through verification (`BDC0ADDF(f)`).
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test ./internal/config/...` passes.

---

### Sub-Epic 3: Shipment `065-S` — Covering Feature `072-F` (Process-lifecycle and TUI observability hardening (`FAE4134F`))

#### Unit 3.1 (`072.001-T`): Honour `startCtx` in `startUI` ready-handshake, bound post-`Kill` wait in `localUI.stop`, and synchronize `watchUI` errors (`FAE4134F(c, f, h)`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files, plus `internal/app/shutdown.go` for `supervise` error propagation):
  * `internal/app/ui.go`
  * `internal/app/assemble.go`
  * `internal/app/tui_test.go`
* **Changes needed**:
  * In `internal/app/ui.go`:
    * Add `watchDone chan struct{}` and `stopErr error` to `localUI`, leaving `ui.watchDone == nil` inside `startUI` until `assemble` / `watchUI` synchronously assigns `ui.watchDone = make(chan struct{})` right before spawning the UI watcher goroutine (so any `ui.stop()` cleanup invoked inside `startUI` before the watcher is registered sees `u.watchDone == nil` and never deadlocks).
    * Guard `(*localUI).failure()` with a non-blocking check on `<-u.done` (`select { case <-u.done: default: return nil }`) so `u.err` is never read unless `prog.Run()` has returned and closed `u.done` (preventing any data race if `stop()` times out after `Kill()`).
    * In `startUI(startCtx context.Context, ...)`: check `if err := startCtx.Err(); err != nil { return nil, fmt.Errorf("terminal startup: %w", err) }` before launching `go prog.Run()`, include `case <-startCtx.Done(): _ = ui.stop(); return nil, fmt.Errorf("terminal startup: %w", startCtx.Err())` in the `ready` handshake `select` (and re-check `startCtx.Err()` after `<-ready`), and pass `context.WithoutCancel(startCtx)` to `ui.feed.Start` (calling `_ = ui.stop()` on error).
    * In `(*localUI).stop() error`: when `<-u.done` closes in either the `Quit()` wait or the bounded `killTimer := time.NewTimer(uiKillWaitBudget)` (`uiStopBudget`) wait after `u.prog.Kill()`, wait on `<-u.watchDone` **only when `u.watchDone != nil`** before returning `nil` so startup cleanup in `startUI` before watcher registration returns immediately without blocking, while once `watchUI` is registered the UI watcher goroutine in `assemble.go` is guaranteed to have finished calling `a.addErr` / `a.addUIErr` before `supervise()` reads `a.takeErrs()` / `a.takeUIErr()`. In the `<-killTimer.C` timeout arm (where `u.done` did not close within `uiKillWaitBudget` after `u.prog.Kill()`), record a non-nil timeout error in `u.stopErr` (e.g. `errors.New("tui stop: timed out waiting for program exit after kill")`), proceed directly to `u.feed.Close()` without waiting on `u.watchDone` or reading `u.err`, and return `u.stopErr`.
  * In `internal/app/assemble.go` and `internal/app/shutdown.go`:
    * Pass `ctx` (`startCtx`) directly to `startUI`, synchronously assign `ui.watchDone = make(chan struct{})` immediately before launching the UI watcher goroutine, and in the UI watcher goroutine add `defer close(ui.watchDone)` so `ui.watchDone` closes immediately after `a.addErr` / `a.addUIErr` and `a.Shutdown`.
    * Propagate any non-nil error returned by `a.ui.stop()` in both `teardown()` (joined into `errs`) and `supervise()` (joined into `a.err` before `close(a.done)`, via `a.addUIErr(a.ui.stop())` / `a.addErr(a.ui.stop())` before `a.takeUIErr()` / `a.takeErrs()`) so a wedged TUI that times out on `uiKillWaitBudget` never reports a clean nil `App.Err()`.
  * Add deterministic unit tests in `internal/app/tui_test.go` verifying: (1) `startUI` startup cleanup returns promptly without deadlocking on a pre-cancelled `startCtx`, on cancellation during the ready handshake, and on `ui.feed.Start` failure (before `watchUI` is registered, when `u.watchDone == nil`), (2) a TUI error occurring concurrently with `Shutdown` (after `watchUI` is registered) is always present once in `a.Err()`, and (3) when `u.done` stays open after `Kill()`, `(*localUI).stop()` completes within `2*uiStopBudget` (`uiStopBudget + uiKillWaitBudget`) AND returns/propagates a non-nil shutdown error (`stop() != nil` and `App.Err() != nil`).
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/app -run "TestTUI|TestCleanTUI"` passes.

#### Unit 3.2 (`072.002-T`): Unconditional `cancelStart` on startup signals, bounded `hub.BeginShutdown` in `teardown`, and startup warning for unconsumed config keys (`FAE4134F(d, f, i)`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/app/shutdown.go`
  * `internal/app/assemble.go`
  * `internal/app/shutdown_test.go`
* **Changes needed**:
  * In `internal/app/shutdown.go` (`watchSignals`): call `cancelStart()` unconditionally on every signal received (even if `a.shuttingDown()` is already true), so a signal during startup always cancels `startCtx`.
  * In `internal/app/assemble.go` (`teardown`): call `a.hub.BeginShutdown(pctx)` bounded by `permBudget := a.timeouts.Permissions` (defaulting to `5 * time.Second` when `<= 0`) before `a.remote.Shutdown(ctx)` and `a.agent.Stop(ctx)`, so any permission ask parked in the hub during a failed start is immediately resolved to `UserNotAvailable` without a wedged handler exhausting the 15s `teardownBudget`.
  * In `internal/app/assemble.go` (`Start`): add `warnUnconsumedConfig(log, cfg)` comparing `cfg` against `config.Default()` in a fixed deterministic order for the accepted-but-unconsumed MVP keys (`max_concurrent_sessions`, `retention_days`, `operator_detail_level`, `database.path`, `timeouts.prompt_seconds`, `timeouts.wait_seconds`, `stall.*`, and `commands`) and logging `log.Warn("config has accepted-but-unconsumed keys", "keys", keys)` when any non-default value is present.
  * Add deterministic tests in `internal/app/shutdown_test.go` covering failed-start `teardown` with a parked permission ask (completing in `< 1s`), signal cancellation when `Shutdown` was already requested during startup, and startup warning emission for unconsumed config keys.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/app -run "TestShutdown|TestTeardown|TestUnconsumed|TestSignal"` passes.

#### Unit 3.3 (`072.003-T`): Zero plaintext credentials after TUI delivery and harden `tuiReissue` failure/revoke semantics (`FAE4134F(e, g)`)

* **Domain**: `code` (`size: S`, `complexity: medium`)
* **Posture**: `test-first`
* **Files affected**:
  * `internal/app/app.go`
  * `internal/app/assemble.go`
  * `internal/app/shutdown.go`
  * `internal/app/review1_test.go`
  * `internal/app/credentials_test.go` (1-line `RetainPlaintext: true` fixture opt-in in `newUIRig`)
  * `internal/app/e2e_test.go` (1-line `RetainPlaintext: true` fixture opt-in in `newE2ERig`)
* **Changes needed**:
  * In `internal/app/app.go` (`TUIOptions`): add `RetainPlaintext bool` (`// RetainPlaintext keeps plaintext tokens in App.Credentials() after TUI delivery (tests only; never set in production).`), independent of `o.TUI.Output`. Update both `newUIRig` (`internal/app/credentials_test.go`) and `newE2ERig` (`internal/app/e2e_test.go`) in this unit to set `RetainPlaintext: true` on `o.TUI` so existing TUI-enabled tests that read `App.Credentials()` continue authenticating in `065-S`. When `!o.TUI.Disabled && !o.TUI.RetainPlaintext`, immediately after `credentialsMsg` constructs the `tui.CredentialsMsg` for `startUI` or `tuiReissue`, zero the `Token` field on each stored entry in `a.creds` (`a.zeroPlaintextCredentials()` synchronized with `a.mintMu` and `a.mu` / `a.credMu`), retaining only `Role`, `Fingerprint`, and `ExpiresAt` in `a.creds`.
  * In `internal/app/assemble.go` and `internal/app/shutdown.go` (`mint`, `Reissue`, `tuiReissue`, `credentialsMsg`, `clearCreds`, `teardown`, `supervise`):
    * Add a synchronized credential-clearing helper `a.clearCreds()` that sets a sticky terminal flag `a.credsCleared = true` and zeroes/clears `a.creds` (`a.creds = nil` / `Credentials{}`) under the publication mutex `a.mu` (`a.credMu`) without holding `a.mintMu` while calling any helper that also acquires `a.mintMu` (avoiding recursive locking), and invoke `a.clearCreds()` unconditionally during normal shutdown in `supervise()` before `close(a.done)` (regardless of `TUI.Disabled` or `RetainPlaintext`), as well as in failed-start `teardown()` and on shutdown/revocation inside `mint()`.
    * In `mint()` (under `a.mintMu`): if `a.shuttingDown()` is true at entry, call `a.clearCreds()` and return `nil, apperr.New(apperr.KindUnavailable, "app: credentials revoked; shutting down")` (or `transport.ErrCredentialRevoked` joined with `transport.ErrShuttingDown`). In the role loop, if `a.remote.Mint(role)` returns an error: when `i == 0` OR `a.shuttingDown()` OR `errors.Is(err, transport.ErrShuttingDown)`, call `a.clearCreds()` if shutting down/revoked and return `nil, fmt.Errorf("mint %s credential: %w", roleName(role), err)` so `tuiReissue` returns an empty `tui.CredentialsMsg{}` and a dismissed TUI panel does not reopen on a total failure or mid-mint shutdown. When publishing the newly minted credential set into `a.creds` under the `a.mu` (`a.credMu`) critical section (including after the final successful `a.remote.Mint(role)`), check `if a.credsCleared || a.shuttingDown()` under that same `a.mu` (`a.credMu`) lock: if true, refuse to write `a.creds`, keep `a.creds` cleared/zeroed (`nil` / `Credentials{}`), and return `nil, transport.ErrCredentialRevoked` (joined with `transport.ErrShuttingDown`) so a `Reissue` that paused after its final successful `remote.Mint()` can never republish revoked plaintext into `a.creds` after shutdown clears it and closes `Done`.
    * In `a.credentialsMsg(issued)`: return `tui.CredentialsMsg{}` immediately when `len(issued) == 0`; skip any entry whose `iss.Token.Reveal() == ""` (an unreplaced role whose plaintext was already zeroed after initial delivery), returning `tui.CredentialsMsg{}` if no entry has a secret.
  * Add unit tests in `internal/app/review1_test.go` verifying: (1) `tuiReissue` when role 0 fails returns an empty `CredentialsMsg{}` (leaving a dismissed TUI panel closed), (2) `Reissue` / `tuiReissue` after shutdown/revoke returns no credentials, (3) with TUI enabled and default `RetainPlaintext: false`, `App.Credentials()` returns zeroed `Token` fields (`Token.Reveal() == ""`) after initial TUI delivery while fingerprints remain intact and authentication via the `CredentialStore` (`dialToken` using the token rendered to the TUI output) still works, (4) starting an `App` with `RetainPlaintext: true` (or `TUI.Disabled: true`), triggering normal shutdown without calling `Reissue`, and awaiting `<-app.Done()` clears `app.Credentials()` so any returned entries have zeroed `Token` fields (`""`) / empty credentials, and (5) a deterministic concurrent `Reissue`/shutdown regression test (using a test seam or hook after `remote.Mint` before publication under `a.mu` / `a.credMu`, or coordinated concurrent `Reissue` + shutdown revocation) proving that a `Reissue` pausing after its final successful `remote.Mint()` and resuming after shutdown revocation + `clearCreds()` / `<-app.Done()` returns `transport.ErrCredentialRevoked` and never leaves non-empty plaintext tokens in `App.Credentials()`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/app -run "TestPartial|TestTUIReissue|TestZeroPlaintext|TestReissue"` passes, and `GOTOOLCHAIN=go1.26.10 go test -race ./internal/app/...` passes.

#### Unit 3.4 (`072.004-T`): Add rejected-upgrade counter and `OnReject` callback to `internal/transport` (`FAE4134F(b)` part 1)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/transport/handler.go`
  * `internal/transport/remote.go`
  * `internal/transport/remote_test.go`
* **Changes needed**:
  * In `internal/transport/handler.go`: add `OnReject func(status int, reason string, total uint64)` to `Config` and an atomic counter `rejected atomic.Uint64` on `Handler` (exposed via `(*Handler).RejectedUpgrades() uint64`). In `(*Handler).reject(w, status, reason)` (which handles `/ws` upgrade rejections `400`, `401`, `403`, `503`), increment `total := h.rejected.Add(1)` and, if `h.cfg.OnReject != nil`, invoke `h.cfg.OnReject(status, reason, total)`.
  * In `internal/transport/remote.go`: add `OnReject func(status int, reason string, total uint64)` to `RemoteOptions`, forward it to `Handler`'s `Config`, and expose `(*Remote).RejectedUpgrades() uint64`.
  * Add a deterministic test in `internal/transport/remote_test.go` verifying that rejected upgrades (`401`, `403`) increment `RejectedUpgrades()` and invoke `OnReject` with the status, sanitized reason, and monotonic total.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/transport/...` passes.

#### Unit 3.5 (`072.005-T`): Surface `SecurityNoticeMsg` rejected-upgrade counter on `internal/tui` status line and wire in `internal/app` (`FAE4134F(b)` part 2)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/tui/model.go`
  * `internal/tui/model_test.go`
  * `internal/app/assemble.go`
* **Changes needed**:
  * In `internal/tui/model.go`: define `type SecurityNoticeMsg struct { Status int; Reason string; Total uint64 }`. Handle `SecurityNoticeMsg` in `Model.Update` with an out-of-order guard (`if msg.Total >= m.rejectedUpgrades { m.rejectedUpgrades = msg.Total; m.lastReject = fmt.Sprintf("%d %s", msg.Status, SanitizeLine(msg.Reason)) }`), and include `fmt.Sprintf("rejected upgrades: %d (last: %s)", m.rejectedUpgrades, m.lastReject)` in `statusLine()` whenever `m.rejectedUpgrades > 0`.
  * In `internal/app/assemble.go`: pre-allocate a capacity-1 coalescing channel `rejectCh chan tui.SecurityNoticeMsg` on `App` (before `transport.NewRemote`), wire `RemoteOptions.OnReject` to push latest-only into `a.rejectCh` under `a.mu` (keeping the highest `Total` without reading `a.ui` or spawning per-request goroutines), and drain `a.rejectCh` into `ui.prog.Send` from a single forwarder goroutine started after `startUI` and bounded by `<-ui.done` (before `close(ui.watchDone)` so `assertNoLeaks` sees zero in-flight goroutines after shutdown).
  * Add unit tests in `internal/tui/model_test.go` verifying `SecurityNoticeMsg` rendering on the status line, out-of-order `Total` guard, and sanitization of `Reason`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race ./internal/tui/... ./internal/app/...` passes.

#### Unit 3.6 (`072.006-T`): Document headless credential decision, TUI security notices, and unconsumed-config warnings in operator runbook (`FAE4134F(a, b, i)`)

* **Domain**: `docs` (`size: XS`, `complexity: trivial`)
* **Posture**: `migration-first`
* **Files affected** (1 file):
  * `docs/runbooks/intercom-operator-runbook.md`
* **Changes needed**:
  * Update `docs/runbooks/intercom-operator-runbook.md` (sections 2, 4, 7, and 9) to document:
    1. The explicit headless credential-delivery decision (`FAE4134F(a)`): the CLI binary requires an interactive terminal for the local TUI because ADR MVP-2 item 3 forbids writing bearer credentials to stdout pipes, logs, or disk files (avoid writing an `intercom --...` command snippet with non-existent flags so `cmd/intercom/runbook_test.go` `TestRunbookFlagsExistOnRootCommand` remains green); programmatic headless operation is supported in-process via `internal/app.Options{TUI: app.TUIOptions{Disabled: true}}`, and non-TTY service mode awaits post-MVP identity-provider authentication (ADR MVP-2 option b).
    2. The on-screen TUI status-line counter for rejected `/ws` WebSocket upgrades (`rejected upgrades: N (last: STATUS REASON)`) when stderr logs are discarded (`FAE4134F(b)`).
    3. The startup warning logged when accepted-but-unconsumed config keys are set to non-default values (`FAE4134F(i)`).
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test ./cmd/intercom/...` passes (including `runbook_test.go`).

---

### Sub-Epic 4: Shipment `066-S` — Covering Feature `073-F` (Maintainability refactors of `internal/app` (`AD8472DD`))

#### Unit 4.1 (`073.001-T`): Split `internal/app/assemble.go` into `credentials.go` and `late_agent.go`, adding `lateAgent.get()` and `App.watchUI`

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `characterization-first`
* **Files affected** (3 files):
  * `internal/app/assemble.go`
  * `internal/app/credentials.go` (new file)
  * `internal/app/late_agent.go` (new file)
* **Changes needed**:
  * Extract `lateAgent` from `internal/app/assemble.go` into `internal/app/late_agent.go` and add `func (l *lateAgent) get() agent.Agent` used by both `Send` and `Abort`.
  * Extract `credentialRoles`, `mintAll`, `mint`, `Reissue`, `tuiReissue`, `credentialsMsg`, `zeroPlaintextCredentials`, `clearCreds` (preserving the sticky `a.credsCleared` terminal revocation check under `a.mu` / `a.credMu` without recursive locking on `a.mintMu`), `roleName`, and `Credentials` from `internal/app/assemble.go` into `internal/app/credentials.go`.
  * Extract the inline TUI watcher goroutine in `assemble` into a named method `func (a *App) watchUI(ui *localUI)` (which synchronously assigns `ui.watchDone = make(chan struct{})` before spawning the goroutine).
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race ./internal/app/...` passes with zero behavior changes.

#### Unit 4.2 (`073.002-T`): Group `App` shutdown fields into `shutdownState` and extract shared `reap()` helper

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `characterization-first`
* **Files affected** (3 files):
  * `internal/app/assemble.go`
  * `internal/app/shutdown.go`
  * `internal/app/review1_test.go`
* **Changes needed**:
  * Group `App`'s shutdown-specific fields (`req chan struct{}`, `once sync.Once`, `reason string`, `ctx context.Context`, `hurry context.CancelFunc`, `timeouts hub.ShutdownTimeouts`, `err error`) into a `shutdownState` struct held on `App` (`shut shutdownState` in `internal/app/shutdown.go` / `assemble.go`), and update `internal/app/review1_test.go:240` (`case <-a.shutReq:` -> `if a.shuttingDown()`).
  * Extract a shared `func (a *App) reap(rctx context.Context) error` helper in `internal/app/shutdown.go` with explicit nil-guards (`a.remote != nil`, `a.hub != nil` calling idempotent `a.hub.Stop()`, `a.hubCancel()`, and `<-a.hubRunDone`, `a.serveDone != nil` waiting on `<-a.serveDone`, `a.ui != nil` calling `a.ui.stop()` and propagating any returned timeout error via `a.addUIErr` / `errs`, and calling `a.clearCreds()` unconditionally under `a.mu` / `a.credMu` to set sticky `a.credsCleared = true` and clear `a.creds` without recursive locking on `a.mintMu`), preserving `teardown`'s exact failed-start ordering (`ui.stop` -> `hub.BeginShutdown` -> `remote.Shutdown` -> `agent.Stop` -> `reap`) and `supervise`'s normal-shutdown ordering (`sd` -> `reap`).
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race ./internal/app/...` passes with zero behavior changes.

#### Unit 4.3 (`073.003-T`): Deduplicate `factory`/`fakeAgent` knobs in `helpers_test.go` and decouple `partialRig` from `credBytes` in `review1_test.go`

* **Domain**: `tests` (`size: S`, `complexity: low`)
* **Posture**: `characterization-first`
* **Files affected** (2 files):
  * `internal/app/helpers_test.go`
  * `internal/app/review1_test.go`
* **Changes needed**:
  * In `internal/app/helpers_test.go`, extract `type agentKnobs struct { startErr error; startBlock, stopBlock bool; abortHook func(); stopHook func(context.Context) }` and embed `agentKnobs` in both `fakeAgent` and `factory` so `factory.build` copies `agentKnobs: f.agentKnobs` in one field without duplicating the five fields across both structs.
  * In `internal/app/review1_test.go`, replace the hardcoded `const credBytes = 32` byte-count coupling in `partialRig` / `budgetRand` with a call-budgeted reader (`calls int`, where each `Read` call satisfies one `Mint` entropy read until `calls == 0`), decoupling the test from `transport.CredentialStore`'s internal byte length.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race ./internal/app/...` passes.

## Dependency Graph

### Shipment-Level DAG

* `063-S` (`labels: ["dag-root"]`, `predecessor_source: "declared_root"`, under explicit `2026-10-26` operator P-001 policy exception over `062-S` code baseline `6290151`; Covering Feature `038-F`: `038.001-T -> 038.002-T`)
  * `<-blocks-- 064-S` (Covering Feature `071-F`: `071.001-T -> 071.002-T -> 071.003-T -> 071.004-T -> 071.005-T -> 071.006-T`)
    * `<-blocks-- 065-S` (Covering Feature `072-F`: `072.001-T -> 072.002-T -> 072.003-T -> 072.004-T -> 072.005-T -> 072.006-T`)
      * `<-blocks-- 066-S` (Covering Feature `073-F`: `073.001-T -> 073.002-T -> 073.003-T`)

### Intra-Shipment Task Dependencies

* **Within `063-S` (`038-F`)**: `038.001-T` -> `038.002-T`
* **Within `064-S` (`071-F`)**: `071.001-T` -> `071.002-T` -> `071.003-T` -> `071.004-T` -> `071.005-T` -> `071.006-T` (note: `071.006-T` depends on `071.003-T` and `071.005-T` because it updates `docs/config-reference.md` and `internal/config/docs_test.go` for the new `86400` validation message).
* **Within `065-S` (`072-F`)**: `072.001-T` -> `072.002-T` -> `072.003-T` -> `072.004-T` -> `072.005-T` -> `072.006-T` (`072.005-T` depends on `072.004-T`'s `OnReject` hook).
* **Within `066-S` (`073-F`)**: `073.001-T` -> `073.002-T` -> `073.003-T` (serialized because `073.001-T` and `073.002-T` both touch `internal/app/assemble.go`).

## Decisions and Rationale

1. **Four single-feature shipments (`063-S..066-S`)**: Required by Ship Step 6.1(a1) `S2` (`n == 1` covering feature per shipment) and isolates behavioral changes (`064-S`, `065-S`) ahead of move-only file splits (`066-S`).
2. **Code baseline `062-S` (`6290151`) and explicit `2026-10-26` operator P-001 policy exception on `063-S`**: `063-S`..`066-S` functionally depend on code merged in `062-S` (`6290151`, `internal/app/workspace.go`), whose post-merge closure (`docs/closure/062-S-070-F-post-merge-closure.md`) has `compaction_status: done` and `closure_status: READY_WITH_CONDITIONS` with two open operator-owned live-environment conditions (`live-sdk-test-pass` and `tool-approval-routing-live`). Under `.github/policies/workflow-policies.md` P-001 ("unless the operator explicitly overrides P-001"), the operator explicitly invoked dark factory mode on `2026-10-26` strictly scoped to Option 1 (`BDC0ADDF`, `FAE4134F`, `AD8472DD`, `038-F` — the post-`062-S` remediation chain) after reviewing the state assessment that highlighted `062-S`'s open `READY_WITH_CONDITIONS` status. Therefore, `063-S` records this explicit operator P-001 policy exception and uses `labels: ["dag-root"]` (`predecessor_source: "declared_root"`) rather than a synthetic `061-S` edge, while preserving `062-S` (`6290151`) as the documented code baseline.
3. **Keeping `cmd/intercom` TUI-only under ADR MVP-2 (`FAE4134F(a)`)**: Any CLI `--no-tui` flag that outputs bearer credentials to a pipe, log, or file violates ADR MVP-2 item 3 and the `writepath` gate; programmatic headless operation uses `app.Options{TUI: app.TUIOptions{Disabled: true}}`.
4. **Native Windows `syscall.CreateToolhelp32Snapshot` + command-line lookup in `liveproc_windows_test.go` (`BDC0ADDF(d)`)**: Uses the standard library `syscall` package on Windows (`CreateToolhelp32Snapshot` plus `OpenProcess` + `NtQueryInformationProcess` + `ReadProcessMemory` for `PEB.ProcessParameters.CommandLine`) to take an in-process snapshot in `< 1ms` while preserving `newChildren` command-line needle detection after an intermediate launcher exits, eliminating PowerShell/WMI startup flakiness under load with zero `go.mod` changes.

## Risks and Caveats

* **`writepath` gate enforcement**: `tools/gatecheck/internal/writepath` rejects any new filesystem write call in non-test `.go` files under `cmd/**` and `internal/**` and pins `reviewedResolveCallers` (`internal/app/workspace.go`: 1 call). No unit in this plan adds a production `Root.Resolve` call or any forbidden write primitive.
* **Cross-platform compilation**: `liveproc_windows_test.go` (`//go:build windows`) and `liveproc_other_test.go` (`//go:build !windows`) must both compile cleanly under `GOOS=windows` and `GOOS=linux`.

## Plan Hardening Signals

* **public API, schema, or contract change**: **PRESENT** — Unit 2.3 (`071.003-T`) adds an upper bound (`MaxApprovalTimeoutSeconds = 86400`) on `[timeouts].approval_seconds` in `internal/config.Validate()`, and Units 3.4–3.5 (`072.004-T`, `072.005-T`) add `OnReject` / `RejectedUpgrades()` to `internal/transport` and `SecurityNoticeMsg` to `internal/tui`.
* **security, auth, permission, or compliance-sensitive behavior**: **PRESENT** — Unit 1.1 (`038.001-T`) black-box verifies `pathsafe.Root.Resolve` containment via `app.Start`; Unit 3.3 (`072.003-T`) zeroes plaintext credentials in `App.creds` after terminal TUI delivery and refuses re-issue after credential revocation.
* **migration, backfill, destructive data/config action, or irreversible step**: **ABSENT** — No data migration or destructive action.
* **external integration, operator checkpoint, or external dependency**: **ABSENT** — No new external dependencies or live external calls required in CI.
* **high runtime, rollout, or rollback risk**: **PRESENT** — Units 3.1–3.2 (`072.001-T`, `072.002-T`) modify startup cancellation, TUI ready-handshake, failed-start teardown, and `localUI.stop` bounded termination.

Requires plan hardening: yes

## Runtime Verification and Closure

* **Shipment `063-S` (`038-F`)**: Test and documentation only; no production runtime surface changed. Verified by `go test -race ./internal/app -run TestStartBlackBoxWorkspace` and `./gatecheck.exe write-path --self-test`.
* **Shipment `064-S` (`071-F`)**: Config validation (`internal/config`) and test/doc surfaces. Verified headlessly via `go test -race ./internal/config/... ./internal/app/... ./internal/hub/... ./internal/transport/...`.
* **Shipment `065-S` (`072-F`)**: Runtime process lifecycle (`internal/app`), transport rejection hook (`internal/transport`), and TUI status line (`internal/tui`). Verified headlessly in CI via injected pipe/terminal seams (`go test -race ./internal/app/... ./internal/tui/... ./internal/transport/... ./cmd/intercom/...`). Rollback is a clean revert of the shipment merge commit.
* **Shipment `066-S` (`073-F`)**: Move-only refactoring of `internal/app`. Verified by full package and e2e test suite (`go test -race ./...`) and `./gatecheck.exe write-path --self-test`.

## Plan Hardening

### 1. Why Hardening Was Required and Context Consulted

Hardening was triggered by three signals: (1) config validation and package contract additions (`[timeouts].approval_seconds` upper bound in `internal/config`, `OnReject` in `internal/transport`, `SecurityNoticeMsg` in `internal/tui`), (2) security/credential-hygiene changes (`pathsafe.Root.Resolve` black-box containment verification and `App.creds` plaintext zeroing / post-revoke re-issue refusal), and (3) process-lifecycle concurrency changes in `internal/app` (`startUI` ready-handshake cancellation, `localUI.stop` post-`Kill` bound, `teardown` `BeginShutdown`, and `watchUI` error synchronization).

**Reinforced instructions and compound learnings consulted**:

* `.github/instructions/constitution.instructions.md` (Principles I, II, III, V, VI, VIII, XI)
* `.github/instructions/strict-safety.instructions.md` (`ProposedAction` / `ActionRisk` / `ActionResult` classification)
* `docs/compound/2026-10-20-process-wiring-credential-panel-and-approval-disarm.md`
* `docs/compound/2026-10-20-tui-permission-modal-safety-and-linux-test-clock.md`
* `docs/compound/2026-10-20-backlogit-cli-pitfalls-and-permission-outcome-lessons.md`
* `docs/compound/2026-09-08-verify-go-stdlib-internals-claims-against-goroot-source.md`

### 2. Protected Invariants

1. **Zero new production `pathsafe.Root.Resolve` call sites and zero filesystem write primitives**: `./gatecheck.exe write-path --self-test` and `./gatecheck.exe retired-arch --self-test` must exit 0 after every task. `internal/app/workspace.go` remains the sole entry in `reviewedResolveCallers` with `calls: 1`.
2. **Fixed config validation rule ordering (R8, I6)**: In `internal/config.Validate()`, rules 1–7 and `validateRemote(c.Remote)` keep their exact existing order; the new `timeouts.approval_seconds <= 86400` check runs immediately after `validateRemote`, and receiver mutation (`c.DefaultWorkspaceRoot`, `c.Workspaces[i].Path`) still happens only after every validation check passes.
3. **In-process test credential seam preservation**: `App.Credentials()` continues to return plaintext tokens when `o.TUI.Disabled` is true or when `o.TUI.RetainPlaintext` is true (an explicit test-only option independent of `o.TUI.Output`), so all existing `e2e_test.go` and `credentials_test.go` tests pass without regression. In production (`!o.TUI.Disabled && !o.TUI.RetainPlaintext`), plaintext tokens in `a.creds` are zeroed immediately after constructing `tui.CredentialsMsg`.
4. **Non-blocking WebSocket upgrade rejection**: `transport.Handler.reject` must never block on the Bubble Tea event loop; the `OnReject` hook in `internal/app` dispatches `tui.SecurityNoticeMsg` non-blockingly via a pre-allocated capacity-1 coalescing channel drained by a single goroutine bounded by `<-ui.done`.

### 3. Strict-Safety Action Classifications (`ProposedAction` / `ActionRisk` / `ActionResult`)

| Action ID | `ProposedAction` Summary & Targets | `change_kind` | `ActionRisk` | `approval_required` | Rollback | `ActionResult` |
|---|---|---|---|---|---|---|
| `PA-01` | Add black-box `app.Start` containment verification for `pathsafe.Root.Resolve` in `internal/app/workspace_blackbox_test.go` and annotate 3 dated decision records (`063-S` / `038-F`). | `local edit` (tests + docs) | `low` | `false` | `git revert` of `063-S` merge commit | `planned` |
| `PA-02` | Enforce `MaxApprovalTimeoutSeconds = 86400` in `internal/config/validate.go`, replace Windows `Win32_Process` PowerShell helper with `syscall.CreateToolhelp32Snapshot` in `internal/app/liveproc_windows_test.go`, harden `readInsideWorkspace`, and update `wsclient`/`hub`/config docs (`064-S` / `071-F`). | `config change` + `local edit` | `moderate` | `false` (pre-approved dark-factory scope) | `git revert` of `064-S` merge commit | `planned` |
| `PA-03` | Harden `internal/app` startup/teardown lifecycle (`startUI` ctx, post-`Kill` bound, `teardown` `BeginShutdown`, `watchUI` error join), zero plaintext in `App.creds` after TUI delivery (`!o.TUI.RetainPlaintext`), refuse post-revoke re-issue, and surface rejected upgrades on TUI status line (`065-S` / `072-F`). | `local edit` (security & lifecycle hardening) | `high` | `false` (pre-approved under `DARK_MODE_ACTIVE` scope) | `git revert` of `065-S` merge commit | `planned` |
| `PA-04` | Move-only refactor of `internal/app` (`assemble.go` -> `credentials.go` + `late_agent.go`, `shutdownState`, `reap()`, test helper deduplication in `helpers_test.go` and `review1_test.go`) (`066-S` / `073-F`). | `local edit` (move-only refactor) | `moderate` | `false` | `git revert` of `066-S` merge commit | `planned` |

### 4. Deepened Verification, Rollback, and Review-Gate Requirements

* **Cross-platform build check for `071.004-T`**: Verify both `GOTOOLCHAIN=go1.26.10 go test -race ./internal/app -run TestProcessTreeSeesAndForgetsAChild` (Windows native) and `GOOS=linux GOTOOLCHAIN=go1.26.10 go test -c ./internal/app` (Linux cross-compile of `liveproc_other_test.go`) so neither build tag breaks CI.
* **Review-gate capability requirement (P-012)**: Plan review (Step 4) MUST spawn the always-on and triggered cross-model persona subagents (`Constitution Reviewer`, `Go Reviewer`, `Scope Boundary Auditor`, `Learnings Researcher`, `Architecture Strategist`, `Security Lens Reviewer`) plus `Adversarial Review`, and append a `## Plan Review` section with literal `dispatch_mode:` and `decision:` markers before harvest.
* **Unresolved operator decisions**: None.

<!-- plan-review-attempt: 1 -->

## Plan Review

```text
dispatch_mode: multi-agent
decision: PASS
```

### Gate Decision and Rationale

* **Verdict**: **PASS** (after incorporating Cycle 1 multi-persona and adversarial review refinements directly into the plan and deliberation artifacts).
* **Plan Hardening Requirement**: `Requires plan hardening: yes` — **SATISFIED** (`## Plan Hardening` section present with `ProposedAction` / `ActionRisk` / `ActionResult` classifications `PA-01`..`PA-04` and protected invariants).

### Persona Coverage Table

| Persona | Mode | Model Route | Initial Verdict | Final Verdict (Post-Refinement) |
|---|---|---|---|---|
| **Constitution Reviewer** | Subagent | Default (`gemini-4-argon`) | PASS (1 P3) | **PASS** |
| **Go Reviewer** | Subagent | Default (`gemini-4-argon`) | ADVISORY (4 P1, 3 P2, 1 P3) | **PASS** |
| **Scope Boundary Auditor** | Subagent | Default (`gemini-4-argon`) | PASS (0 findings) | **PASS** |
| **Learnings Researcher** | Subagent | Default (`claude-haiku-5.5`) | PASS (11 high-confidence solutions integrated) | **PASS** |
| **Architecture Strategist** | Anchor Subagent | Anchor (`gpt-6.1-sol` / `high`) | FAIL -> Remediated (1 P1, 3 P2, 1 P3) | **PASS** |
| **Security Lens Reviewer** | Subagent | Default (`gemini-4-argon`) | ADVISORY (2 P2, 1 P3) | **PASS** |
| **Adversarial Review** | Subagent | Multi-model consensus | PASS WITH REFINEMENTS (3 P1, 3 P2, 2 P3) | **PASS** |

### Findings and Applied Remediations

1. **[P1 — Remediated] Unit 2.4 (`071.004-T`) Windows process-table helper dependency, `unused` linter, and orphan command-line detection**:
   * *Finding*: `golang.org/x/sys` is `// indirect` in `go.mod:62`, keeping `parseProcessTable` in `liveproc_test.go` when only `!windows` calls it fails `golangci-lint` (`unused` / `unparam`), and relying solely on `ProcessEntry32.ExeFile` would miss an orphaned `node.exe ...\copilot\index.js` child after its launcher exits.
   * *Remediation applied*: Specified standard library `syscall` (`CreateToolhelp32Snapshot`, `ProcessEntry32` with `entry.Size = uint32(unsafe.Sizeof(entry))`, `Process32First`, `Process32Next`, `CloseHandle`, `UTF16ToString`) plus native Win32 command-line lookup (`OpenProcess` + `NtQueryInformationProcess` PEB `ProcessParameters.CommandLine` via `ReadProcessMemory`, falling back to `ExeFile` only on exited/inaccessible processes) in `liveproc_windows_test.go` (zero `go.mod` edits), moved `parseProcessTable(s string)` into `liveproc_other_test.go`, and added an explicit orphaned-grandchild regression test in `liveproc_test.go`.
2. **[P1 — Remediated] Toolchain pin (`GOTOOLCHAIN=go1.26.10`)**:
   * *Finding*: `go.mod:25` declares `toolchain go1.26.10`.
   * *Remediation applied*: Updated all verification commands and Constitution Check references to `GOTOOLCHAIN=go1.26.10`.
3. **[P1 — Remediated] Unit 3.1 (`072.001-T`) `localUI.failure()` race, `watchUI` synchronization (`watchDone` nil-guard before registration), and post-`Kill` timeout error propagation**:
   * *Finding*: Bounding the post-`Kill()` wait in `(*localUI).stop()` means `stop()` can return while `prog.Run()` is still running; reading `u.err` without checking `<-u.done` causes a data race, initializing `watchDone` inside `startUI` would deadlock startup cleanup if `startUI` calls `ui.stop()` on ready-handshake cancellation or `ui.feed.Start` failure before `assemble` registers the watcher goroutine, and returning from `stop()` after the post-`Kill` timeout without recording/returning an error would allow `supervise()` to finalize a nil `App.Err()` on a wedged TUI.
   * *Remediation applied*: Guarded `(*localUI).failure()` with a non-blocking `select { case <-u.done: default: return nil }`, left `ui.watchDone == nil` in `startUI` until `assemble`/`watchUI` synchronously assigns `ui.watchDone = make(chan struct{})` right before spawning the watcher goroutine, waited on `<-u.watchDone` in `(*localUI).stop()` only when `u.watchDone != nil`, changed `(*localUI).stop() error` to record and return a non-nil timeout error (`u.stopErr`) when `uiKillWaitBudget` (`uiStopBudget`) expires after `u.prog.Kill()` and propagated that error through both `teardown()` and `supervise()` (joined into `a.err` before `close(a.done)` via `a.addUIErr(u.stop())` / `a.addErr(u.stop())` before `a.takeUIErr()` / `a.takeErrs()`), checked `startCtx.Err()` before launching `prog.Run()` in `startUI`, and required unit test coverage in `tui_test.go` for ready-handshake cancellation and `feed.Start` failure cleanup without deadlock as well as concurrent TUI failure join and a wedged-TUI test asserting both bounded post-`Kill` completion and a non-nil shutdown error (`stop() != nil` and `App.Err() != nil`).
4. **[P1 — Remediated] Unit 3.5 (`072.005-T`) `OnReject` race on `a.ui`, goroutine flood, and out-of-order `Total`**:
   * *Finding*: `remote.Serve()` starts before `a.ui = ui` is assigned in `assemble.go`, so reading `a.ui` in `OnReject` races under `-race`, per-request goroutines can flood/leak, and concurrent `ServeHTTP` calls can deliver `Total` out of order.
   * *Remediation applied*: Pre-allocated a capacity-1 coalescing channel `rejectCh` on `App` before `NewRemote` (written under `a.mu` keeping the highest `Total` without reading `a.ui`), drained by a single forwarder goroutine started after `startUI` and bounded by `<-ui.done` before `close(ui.watchDone)`, and guarded `Model.Update` with `if msg.Total >= m.rejectedUpgrades`.
5. **[P2 — Remediated] Unit 2.5 (`071.005-T`), Unit 3.2 (`072.002-T`), Unit 3.3 (`072.003-T`), and Unit 4.2 (`073.002-T`)**:
   * *Remediation applied*: Replaced `filepath.EvalSymlinks` in `readInsideWorkspace` without mutating `summary` with `strings.TrimSpace` and without stripping quotes or `file://` URL prefixes (rejecting empty strings `summary == ""` and any path that changes when trimmed `strings.TrimSpace(summary) != summary` so leading/trailing whitespace or whitespace-only strings fail closed with `return false`, rejecting NUL strings, rejecting invalid UTF-8 `!utf8.ValidString(summary)` or `'\uFFFD'` `strings.ContainsRune(summary, '\uFFFD')` or any Unicode control/format rune `unicode.IsControl(r) || unicode.Is(unicode.Cf, r)` or `sanitizeBound`-truncated summaries `utf8.RuneCountInString(summary) > 2048 || strings.HasSuffix(summary, "…")` matching `maxSummaryRunes = 2048` in `internal/agent/copilot/permission.go:66,125` while allowing valid untruncated contained paths of 512–2048 runes, rejecting any literal `".."` path component before any `filepath.Clean` / `filepath.Dir` / `filepath.Rel` / `root.Resolve` call by splitting on `'/'` and `'\\'`, failing closed with `return false` on any `file:` URL scheme prefix via `strings.HasPrefix(strings.ToLower(p), "file:")`, and preserving literal quote characters untouched so `root.Resolve` evaluates the exact literal filesystem entry rather than an unquoted or trimmed lookalike), performing only Windows extended-length prefix normalization before the `filepath.IsAbs` / `filepath.Rel` split (checking `strings.HasPrefix(p, `\\?\`) && len(p) >= 8 && strings.EqualFold(p[4:8], `UNC\`)` FIRST -> `\\` + `p[8:]`, rejecting `\\.\` device-namespace prefixes outright, and only then stripping `\\?\` for drive-letter paths, with an explicit untruncated long-path regression for a contained nested path of 512–2048 runes inside `<workspace>` returning `true` plus six explicit lookalike-collision / truncation regression cases: `<workspace>/UNC/outside/share/secret.txt` vs `\\?\UNC\outside\share\secret.txt`, `<workspace>/outside/share/secret.txt` vs `file://outside/share/secret.txt`, unquoted `<workspace>/notes.txt` vs quoted `'notes.txt'` / `"notes.txt"` absent or outside-pointing symlink, unpadded `<workspace>/notes.txt` vs leading/trailing-space `" notes.txt "` / `"notes.txt "`, relative `link/../notes.txt` and absolute `filepath.Join(ws, "link") + "/../notes.txt"` where `<ws>/notes.txt` exists inside and `<ws>/link` points outside, and `<ws>/notes\uFFFD.txt` vs `notes\uFFFD.txt` from `sanitizeBound` on `notes\x1b.txt` alongside a truncated-summary negative case `> 2048` runes / `2048` runes + `"…"` from `sanitizeBound`) plus `pathsafe.NewRoot(filepath.Dir(p))` + `root.Resolve(rel)` and snapshotted `issued := a.Credentials()` before shutdown in `TestLiveCopilotEndToEnd`; bounded `a.hub.BeginShutdown` in `teardown` by `a.timeouts.Permissions` (default 5s); replaced implicit `o.TUI.Output == nil` detection with explicit `TUIOptions.RetainPlaintext` policy, moved both `newUIRig` (`internal/app/credentials_test.go`) and `newE2ERig` (`internal/app/e2e_test.go`) `RetainPlaintext: true` fixture updates into Unit 3.3 (`072.003-T`) alongside dedicated tests in `review1_test.go` for default `RetainPlaintext: false` zeroing while `CredentialStore` authentication still works, for normal-shutdown clearing (`a.clearCreds()` setting sticky `a.credsCleared = true` and clearing `a.creds` under `a.mu` / `a.credMu` without recursive locking on `a.mintMu`, called unconditionally in `supervise()` before `close(a.done)` so `App.Credentials()` returns zeroed `Token` fields after `<-app.Done()` without calling `Reissue`), and for a deterministic concurrent `Reissue`/shutdown regression proving that a `Reissue` pausing after its final `remote.Mint()` and resuming after shutdown revocation + `clearCreds()` checks `a.credsCleared || a.shuttingDown()` under the `a.mu` (`a.credMu`) publication lock, refuses to write `a.creds`, and returns `transport.ErrCredentialRevoked`; included `internal/app/review1_test.go` in Unit 4.2 (`073.002-T`) and specified nil-guards + `a.hub.Stop()` + `a.ui.stop()` error propagation + `a.clearCreds()` in `reap(rctx)` while preserving failed-start vs normal-shutdown step ordering.
6. **[P3 — Remediated] Unit 2.6 (`071.006-T`) Width Isolation & Unit 3.6 (`072.006-T`) `runbook_test.go` guard**:
   * *Remediation applied*: Kept Unit 2.6 strictly markdown-only (`docs/config-reference.md` and `docs/runbooks/live-sdk-tests.md`), clarified `Handler.reject` `/ws` status scope (`400`, `401`, `403`, `503`), and noted the `cmd/intercom/runbook_test.go` invocation-line regex constraint in Unit 3.6.


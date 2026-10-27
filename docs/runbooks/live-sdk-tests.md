# Live Copilot SDK tests (opt-in)

These tests talk to a **real** Copilot runtime. They are skipped by default
(CI never runs them) and are enabled with `INTERCOM_LIVE_SDK_TESTS`, the
workspace-wide default-deny gate described in design section 7.6
(`strconv.ParseBool`; empty or unparsable means skipped).

## Commands

Bash / zsh:

```sh
INTERCOM_LIVE_SDK_TESTS=1 go test ./internal/app -run Live -v
INTERCOM_LIVE_SDK_TESTS=1 go test ./internal/agent/copilot -run Live -v
```

PowerShell:

```powershell
$env:INTERCOM_LIVE_SDK_TESTS = "1"; go test ./internal/app -run Live -v
$env:INTERCOM_LIVE_SDK_TESTS = "1"; go test ./internal/agent/copilot -run Live -v
Remove-Item Env:INTERCOM_LIVE_SDK_TESTS   # switch it off again
```

## Prerequisites

* The GitHub Copilot CLI installed and on `PATH` (or point at it with
  `INTERCOM_LIVE_CLI_PATH`, an absolute path), at or above the version the
  adapter validates (`copilot.ValidatedCLIVersion`).
* The CLI authenticated: either an interactive login already done for the
  current user, or a GitHub token in the environment the CLI reads
  (for example `GH_TOKEN` or `GITHUB_TOKEN`). Never put a token in
  `config.toml`; the config loader rejects credential-shaped keys.
* Network access to GitHub, and a Copilot entitlement for the account.

## What runs

| Test | Package | Covers |
|---|---|---|
| `TestLiveCopilotEndToEnd` | `internal/app` | Scenario steps 2, 5 and 7 with the real runtime through `app.CopilotAgent`, the real hub, WebSocket server and reference client: `Start` is the readiness gate; a prompt that reads `notes.txt` in a temp workspace raises a permission request that the reference client answers (allowed only for a `read`; anything else is denied); SIGINT-style shutdown completes; **no CLI child process is left**. |
| `TestLiveSQbSQcReverification` | `internal/agent/copilot` | Re-verifies SQ-b (session event callbacks are serialised: max in-flight is 1) and SQ-c (events keep being dispatched while a permission handler is parked) at the pinned SDK version, by instrumenting the callbacks the adapter registers. Denies the request; never approves. |
| `TestLiveStartAndStop` | `internal/agent/copilot` | Start, session id, clean Stop. |
| `TestLiveOptInSemantics`, `TestLiveReadApprovalIsConfinedToWorkspace`, `TestProcessTreeSeesAndForgetsAChild`, `TestProcessTreeFindsOrphanedGrandchildByCommandLine` | `internal/app` | Not live, and not gated: they run in the normal `go test` suite. They cover the gate's default-deny parse, the read-approval path confinement, and checks that the process-table method below works on the current OS (including finding a reparented child by command line). |

SQ-b and SQ-c live in the adapter package because they need the SDK callback
seam, and the `copilot-sdk-boundary` depguard rule allows the SDK import only
under `internal/agent/copilot`. Re-run both whenever `copilot-sdk/go` is
upgraded (design 4.2 and 4.4 require it) and if either fails, revert to the
conservative postures those sections name.

## How "no stranded CLI child" is checked

The SDK starts the CLI with `exec.Command`, so the CLI is a descendant of the
test process. The test snapshots the OS process table before `Start`
(Windows: `syscall.CreateToolhelp32Snapshot`, in process; Linux and macOS:
`ps -axo pid=,ppid=,args=`) and again after shutdown, and fails if any process
that was not in the first snapshot is still a descendant of the test process,
or (to catch a shim whose parent already exited) has a command line containing
`copilot` (override with `INTERCOM_LIVE_CLI_PROCESS`). It first requires that
the CLI **was** visible after `Start`, so a method that cannot see the child
fails loudly instead of passing vacuously. Only pids and executable names are
logged, never command lines.

On Windows the snapshot spawns no child process. A PowerShell
`Get-CimInstance Win32_Process` version would need a separate interpreter
process for every snapshot, which adds startup latency and another failure
mode to a check that runs twice per test, so the native `syscall` approach
replaces it. `internal/app/liveproc_windows_test.go` uses only the
standard library `syscall` package:

* `CreateToolhelp32Snapshot`, `Process32First` and `Process32Next` yield the
  pid, parent pid and executable name of every process
* the full command line of each process is read from its own memory with
  `NtQueryInformationProcess` (the PEB `ProcessParameters.CommandLine`) and
  `ReadProcessMemory`, so a reparented child such as `node.exe ...\copilot\index.js`
  still matches the `copilot` needle after its launcher exits
* a process that already exited, denies access, or cannot be read falls back
  to its executable name

The POSIX implementation (`ps`) is in `internal/app/liveproc_other_test.go`.

## Optional environment

| Variable | Effect |
|---|---|
| `INTERCOM_LIVE_PROMPT` | Replaces the default read-a-file prompt (use it if your runtime auto-approves reads and never asks). |
| `INTERCOM_LIVE_CLI_PATH` | Absolute path of the CLI, written to `[copilot].cli_path` for the test. |
| `INTERCOM_LIVE_CLI_PROCESS` | Process-name needle for the stranded-child check (default `copilot`). |

## Results of the first live run (2026-10-20)

Environment: Copilot CLI 1.0.100, `GH_TOKEN` authentication, Windows, go1.26.9.
These are results from an operator-present session, not from CI. The run
predates the native `CreateToolhelp32Snapshot` helper and the hardened
`readInsideWorkspace`, so live re-confirmation of the native helper is pending.
Until then it is covered by the offline tests
`TestProcessTreeSeesAndForgetsAChild` and
`TestProcessTreeFindsOrphanedGrandchildByCommandLine`.

| Test | Result |
|---|---|
| `TestLiveCopilotEndToEnd` | Pass. The real CLI started through `app.CopilotAgent` with the real hub, WebSocket server and reference client. The prompt raised a `read` permission request that the reference client answered. Shutdown was clean. The CLI process tree was seen after `Start` and no CLI child remained after shutdown. |
| `TestLiveReadApprovalIsConfinedToWorkspace` | Passed alongside the live tests in the same `internal/app` run. It is an offline test in the normal suite, not one of the live tests. |
| `TestLiveStartAndStop` | Pass. |
| `TestLiveSQbSQcReverification` | Failed as inconclusive. SQ-b is **verified**: the maximum number of in-flight callbacks was 1 across 106 events, with one permission handler entry. SQ-c is **not re-verified**: no session event was dispatched in the 30 seconds that the permission handler was parked, although 60 events had been dispatched before the request. The runtime was probably waiting for the answer, but that cannot be told apart from head-of-line blocking. |

Design section 4.4 SQ-c therefore remains an unverified assumption, and the
conservative posture stays in force. Re-verifying it needs a prompt or event
source that emits events while the permission request is pending. Until then,
treat a failure of `TestLiveSQbSQcReverification` for SQ-c as "unknown", not as
"blocking confirmed".

Only the `read` permission kind has been exercised live. Shell and write kinds
have not.

## Read permission path to summary mapping

For a `read` permission the adapter (`internal/agent/copilot/permission.go`)
takes the SDK's `rpc.PermissionRequestRead.Path` and publishes it as the
request `Summary`. The 2026-10-20 live run confirmed the end of that chain: the
real runtime raised a `read` request through `app.CopilotAgent` and the
reference client received and answered it (see
[the 062-S evidence](../closure/evidence/062-S-live-verification-evidence.md),
section A). The evidence does not record the literal `Summary` value, so the
properties below come from the adapter code, not from the run.

* `Summary` is the raw `Path`, with no trimming, quote handling or `file:`
  stripping
* `sanitizeBound` bounds it to 2048 runes (`maxSummaryRunes`) and appends `…`
  an ellipsis marker (`U+2026`) when the path was longer
* control and bidi/format runes are replaced by `U+FFFD`, so the summary can be
  a lossy copy of the path the runtime opens

`readInsideWorkspace` in `internal/app/live_test.go` therefore fails closed. It
treats any padded, lossy (`U+FFFD`, invalid UTF-8, control or format rune) or
truncated (over 2048 runes, or ending in the ellipsis marker) summary as not inside the
workspace, and it rejects `file:` URLs, device (`\\.\` and `//./`) paths,
literal `..` segments, and extended (`\\?\`) paths off Windows (on Windows
only the backslash `\\?\` spelling is accepted) before resolving anything. A
summary that passes those checks is resolved with `internal/pathsafe` and must
name an existing non-directory inside the temp workspace. The live test
approves a request only when its kind is `read` and `readInsideWorkspace`
returns true.

## Dev Tunnel verification

The Dev Tunnel check for `devtunnel create ... --allow-anonymous false` and for
the intercom `Authorization` header reaching the backend unmodified through the
tunnel was completed in the same session. Results, setup and the list of
untested items are in
[the 062-S live verification evidence](../closure/evidence/062-S-live-verification-evidence.md),
section B. It also records that a tunnel client needs a separate Dev Tunnels
access token in `X-Tunnel-Authorization`.

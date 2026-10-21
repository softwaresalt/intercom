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
| `TestLiveOptInSemantics`, `TestProcessTreeSeesAndForgetsAChild` | `internal/app` | Not live: the gate's default-deny parse, and a check that the process-table method below works on the current OS. They run in the normal suite. |

SQ-b and SQ-c live in the adapter package because they need the SDK callback
seam, and the `copilot-sdk-boundary` depguard rule allows the SDK import only
under `internal/agent/copilot`. Re-run both whenever `copilot-sdk/go` is
upgraded (design 4.2 and 4.4 require it) and if either fails, revert to the
conservative postures those sections name.

## How "no stranded CLI child" is checked

The SDK starts the CLI with `exec.Command`, so the CLI is a descendant of the
test process. The test snapshots the OS process table before `Start`
(Windows: `Win32_Process` through PowerShell; Linux and macOS:
`ps -axo pid=,ppid=,args=`) and again after shutdown, and fails if any process
that was not in the first snapshot is still a descendant of the test process,
or (to catch a shim whose parent already exited) has a command line containing
`copilot` (override with `INTERCOM_LIVE_CLI_PROCESS`). It first requires that
the CLI **was** visible after `Start`, so a method that cannot see the child
fails loudly instead of passing vacuously. Only pids and executable names are
logged, never command lines.

## Optional environment

| Variable | Effect |
|---|---|
| `INTERCOM_LIVE_PROMPT` | Replaces the default read-a-file prompt (use it if your runtime auto-approves reads and never asks). |
| `INTERCOM_LIVE_CLI_PATH` | Absolute path of the CLI, written to `[copilot].cli_path` for the test. |
| `INTERCOM_LIVE_CLI_PROCESS` | Process-name needle for the stranded-child check (default `copilot`). |

## Results of the first live run (2026-10-20)

Environment: Copilot CLI 1.0.100, `GH_TOKEN` authentication, Windows, go1.26.9.
These are results from an operator-present session, not from CI.

| Test | Result |
|---|---|
| `TestLiveCopilotEndToEnd` | Pass. The real CLI started through `app.CopilotAgent` with the real hub, WebSocket server and reference client. The prompt raised a `read` permission request that the reference client answered. Shutdown was clean. The CLI process tree was seen after `Start` and no CLI child remained after shutdown. |
| `TestLiveReadApprovalIsConfinedToWorkspace` (and the other `Live` tests in `internal/app`) | Pass. |
| `TestLiveStartAndStop` | Pass. |
| `TestLiveSQbSQcReverification` | Failed as inconclusive. SQ-b is **verified**: the maximum number of in-flight callbacks was 1 across 106 events, with one permission handler entry. SQ-c is **not re-verified**: no session event was dispatched in the 30 seconds that the permission handler was parked, although 60 events had been dispatched before the request. The runtime was probably waiting for the answer, but that cannot be told apart from head-of-line blocking. |

Design section 4.4 SQ-c therefore remains an unverified assumption, and the
conservative posture stays in force. Re-verifying it needs a prompt or event
source that emits events while the permission request is pending. Until then,
treat a failure of `TestLiveSQbSQcReverification` for SQ-c as "unknown", not as
"blocking confirmed".

Only the `read` permission kind has been exercised live. Shell and write kinds
have not.

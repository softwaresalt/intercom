---
type: circuit-breaker
timestamp: 2026-09-19T05:52:00Z
agent: "Orchestrator"
skill: "safety-modes investigate-first"
breaker_type: universal
operation: "engram daemon readiness for intercom-go"
attempts: 3
identity: "engram-readiness-intercom-go-exit-2-readiness-timeout"
---

# Circuit Breaker - Engram daemon readiness for intercom-go

## Failure Chain

### Attempt 1

* Exit/timeout: native exit code 2
* Operation evidence: `engram health`; cwd
  `C:\Source\GitHub\intercom-go`; recovery phase; daemon PID `25080`;
  `.engram\run\engram.pid`
* Normalized message: daemon failed to reach Ready within 30 seconds
* Diagnostic artifact: command result in session shell `204`

### Attempt 2

* Exit/timeout: native exit code 2
* Operation evidence: `engram --workspace
  C:\Source\GitHub\intercom-go health`; cwd
  `C:\Source\GitHub\intercom-go`; recovery phase; replacement daemon PID
  `25036`; `.engram\run\engram.pid`
* Normalized message: replacement daemon failed to reach Ready within 30
  seconds
* Diagnostic artifact: command result in session shell `211`

### Attempt 3

* Exit/timeout: native exit code 2 on bounded readiness probes
* Operation evidence: foreground `engram daemon --workspace
  C:\Source\GitHub\intercom-go --format text`; cwd
  `C:\Source\GitHub\intercom-go`; recovery phase; diagnostic daemon PID
  `38764`; readiness probes `daemon-status`, `workspace-status`, and `health`
* Normalized message: daemon acquired the workspace lock but did not reach
  Ready; CPU exceeded 327 seconds, working set exceeded 1.2 GiB, and captured
  output grew to 278.9 KiB before termination
* Diagnostic artifact: bounded session shells `engram-diag`, `220`, `221`,
  and `222`; raw temporary transport output was not copied into the workspace

## Context

* Files involved: `.engram\run\engram.pid`,
  `.engram\run\engram.lock`, `.engram\config.toml`,
  `.engram\diagnostics\shim-startup-failures.jsonl`
* Process classification: only three daemon processes existed across observed
  workspaces; the other Engram processes were client shims owned by live
  Copilot or VS Code processes
* Protected daemons: backlogit PID `17592` and engram PID `18600` were not
  touched
* Approved cleanup: the operator authorized stale Engram cleanup; only the
  confirmed-unready intercom-go daemon PIDs `25080` and `25036` were stopped
* Provisional-to-concrete identity link: all attempts resolved to the same
  workspace, exit code, readiness-timeout message, PID-file target, and
  recovery phase
* Logging controls: output was bounded by the tool transport; no raw external
  temporary output, environment values, credentials, or payloads were
  persisted
* ActionResult: applied after operator-completed offline indexing
* Resolution: the operator completed `engram index --direct`, creating a
  materially new startup condition; the next authorized daemon reached Ready
  in 8 seconds
* Verified daemon: PID `4388`, Engram `0.3.0-rc.1`, 45 MiB initial memory
* Verified workspace: correct workspace identity and branch, 91 code files,
  342 functions, and 367 edges; all daemon health checks green

## Recovery Outcome

The daemon failure was not caused by ten duplicate daemons. Only three Engram
daemon processes existed across the observed workspaces; the remaining
processes were client shims owned by live Copilot or VS Code sessions.

The operator-completed offline index removed the expensive startup condition.
The supported CLI lifecycle then spawned exactly one healthy daemon for
`intercom-go`. No other workspace daemon was stopped or modified.

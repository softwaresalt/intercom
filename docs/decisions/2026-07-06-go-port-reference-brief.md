---
title: "intercom-go — Port Reference Brief"
type: reference
date: 2026-07-06
source_repo: "softwaresalt/agent-intercom (Rust)"
target_repo: "intercom-go (new, Go)"
related: "docs/decisions/2026-07-06-rust-vs-go-intercom-spike.md"
tags:
  - "go"
  - "port"
  - "architecture"
superseded_by: "docs/decisions/2026-09-04-intercom-go-architecture-correction-copilot-sdk-deliberation.md"
governing: false
status: superseded
---

> [!IMPORTANT]
> **SUPERSEDED / NON-GOVERNING — 2026-09-04.**
> This artifact is retained as **historical evidence only**. It is no longer
> authoritative for intercom-go's architecture, operator surface, transport,
> configuration schema, or credentials.
>
> It was written under assumptions that an explicit operator architecture
> correction has since retired: Slack as the operator surface, an
> application-owned ACP broker / `coder/acp-go-sdk` / headless Copilot CLI
> lifecycle, an iOS remote client, and the Rust `softwaresalt/agent-intercom`
> repository as a **behavioral oracle**.
>
> The governing architecture is
> [`docs/decisions/2026-09-04-intercom-go-architecture-correction-copilot-sdk-deliberation.md`](2026-09-04-intercom-go-architecture-correction-copilot-sdk-deliberation.md)
> and the governing design is
> [`docs/design-docs/intercom-go-backend-architecture.md`](../design-docs/intercom-go-backend-architecture.md)
> (revision 2).
>
> **Do not plan, implement, or review against this document.** The Rust
> repository is now *historical reference only*, not a behavioral oracle.


# intercom-go — Port Reference Brief

This is the authoritative spec for building a **feature-complete Go port** of
`agent-intercom` (Rust). It is written for the agent/operator working in the new
`intercom-go` workspace. The Rust repo is the **behavioral oracle**: when this
brief and the Rust source/tests disagree, the Rust source + its test corpus win.

## 1. How to use this brief

1. Stand up `intercom-go` as its own repo with its own CI.
2. Give the Go workspace **read-only access to the Rust repo** (pick one):
   * add the Rust repo as a git **submodule/subtree** under `reference/intercom-rust/`
     (keeps reads inside the Go workspace's own tree), or
   * launch the Go session with `../intercom` provided as explicit read context.
3. Port **contract-first**: the Rust `tests/contract/` and `tests/integration/`
   suites define the exact behaviors. Re-express them as Go tests and make them pass.
4. **Front-load the three risk slices** (Section 11) before breadth.
5. Keep the Rust app as the **live fallback** until the Go port reaches parity.

## 2. What intercom is (one paragraph)

A remote agent-supervision server. It spawns/attaches an AI coding agent, mediates
that agent's requests (code-change approvals, continuation prompts, standby waits,
terminal-command gating) to a **human operator over Slack**, enforces workspace
policy (auto-approve rules) and path isolation, detects stalls and nudges/escalates,
persists all state to SQLite, and exposes a companion CLI (`agent-intercom-ctl`) over
IPC. It runs in two protocol modes: **ACP** (Agent Communication Protocol — spawn a
CLI agent, talk JSON-RPC over stdio; the strategic direction) and **MCP** (legacy;
being retired). Port ACP-first.

## 3. Recommended Go stack (CGO-free)

The central win of Go here is a fully CGO-free static binary. Preserve that.

| Concern | Rust today | Go target | Notes |
|---|---|---|---|
| Async runtime | `tokio` (full) | goroutines + `context.Context` | Use `context` for cancellation/shutdown everywhere. |
| HTTP / SSE | `axum` 0.8 | `net/http` (+ `go-chi/chi`) | Health endpoint + (ACP) MCP-tool HTTP surface. |
| Slack Socket Mode | `slack-morphism` 2.17 | `slack-go/slack` (+ `socketmode`) | **Risk slice #1** — verify block kit, modals, interactions, threading. |
| SQLite | `sqlx` + `libsqlite3-sys` (C) | **`modernc.org/sqlite`** (pure Go) + `database/sql` | Keeps build CGO-free. **Risk slice #3** — verify perf/queries. |
| TLS | rustls (`ring`/`aws-lc-sys`, C) | stdlib `crypto/tls` (pure Go) | No extra dep; this removes the C crypto surface. |
| Subprocess/stdio (ACP) | `tokio::process` + custom codec | stdlib `os/exec` + `bufio` + `encoding/json` | **Risk slice #2**. |
| IPC (named pipes / UDS) | `interprocess` 2.0 | UDS via `net`; Windows pipes via `microsoft/go-winio` | Companion `ctl` transport. |
| FS watch (config hot-reload) | `notify` 6.1 | `fsnotify/fsnotify` | Watch the config dir, not the file (atomic rename). |
| OS keychain | `keyring` 3 | `zalando/go-keyring` | Bot/app tokens; env-var fallback. |
| Unified diff | `diffy` 0.4 | `sourcegraph/go-diff` or `sergi/go-diff` | Parse + apply patches; atomic writes. |
| CLI args | `clap` 4.5 | `spf13/cobra` or `urfave/cli` | Two binaries: server + `ctl`. |
| Config (TOML) | `toml` 0.8 | `BurntSushi/toml` or `pelletier/go-toml` | Match `config.toml` schema (Section 9). |
| Logging/tracing | `tracing` + `tracing-subscriber` | stdlib `log/slog` (JSON handler) | Structured logs; keep field names stable for observability. |
| Unique IDs | `uuid` v4 | `google/uuid` | Prefixed IDs (`session:uuid`, etc.). |
| Time | `chrono` | stdlib `time` (RFC3339) | DB stores timestamps as TEXT (RFC3339). |
| MCP SDK (legacy) | `rmcp` 0.13 | `mark3labs/mcp-go` (only if MCP retained) | MCP is being retired — deprioritize. |

## 4. Architecture & module map

Layered, protocol-agnostic core with pluggable drivers. Suggested Go package layout
(`internal/` for app packages):

| Rust module | Responsibility | Go package |
|---|---|---|
| `config.rs`, `config_watcher.rs` | Global config, `[[workspace]]` mappings, hot-reload | `internal/config` |
| `errors.rs` | `AppError` taxonomy (Section 10) | `internal/apperr` |
| `mode.rs` | `ServerMode` {mcp, acp} | `internal/mode` |
| `driver/` | **`AgentDriver` trait + `AgentEvent`** (Section 5) | `internal/driver` |
| `driver/acp_driver.rs` | ACP driver impl | `internal/driver/acp` |
| `driver/mcp_driver.rs` | MCP driver impl (legacy) | `internal/driver/mcp` |
| `acp/` | ACP subprocess: `spawner`, `handshake`, `reader`, `writer`, `codec` | `internal/acp` |
| `models/` | Domain types (Section 6) | `internal/models` |
| `persistence/` | SQLite repos + schema + retention | `internal/store` |
| `policy/` | Workspace auto-approve rules | `internal/policy` |
| `diff/` | Unified-diff parse + atomic apply within workspace root | `internal/diff` |
| `slack/` | Socket Mode client, block kit, slash commands, interaction handlers | `internal/slack` |
| `mcp/` | MCP tools + HTTP/SSE transport (legacy/ACP-tool bridge) | `internal/mcp` |
| `orchestrator/` | Session lifecycle, spawn, stall consumer | `internal/orchestrator` |
| `ipc/` | Named-pipe/UDS server for `ctl` | `internal/ipc` |
| `audit/` | Append-only audit logger | `internal/audit` |
| `state/` | `AppState` (shared deps handle) | `internal/state` |
| `main.rs` | Server bootstrap (tokio → goroutines) | `cmd/intercom` |
| `ctl/main.rs` | Companion CLI | `cmd/intercom-ctl` |

## 5. Core abstraction to preserve — `AgentDriver`

This is the seam that made ACP/MCP interchangeable; keep it as a Go interface.

```go
// AgentEvent variants (from src/driver/mod.rs):
//   ClearanceRequested{ RequestID, SessionID, Title, Description, Diff *string, FilePath, RiskLevel }
//   PermissionRequested{ RequestID, RequestIDRaw json.RawMessage, SessionID, Title, Description, FilePath, RiskLevel, Options []PermissionOption }  // standard ACP session/request_permission (ADR-0016)
//   StatusUpdated{ SessionID, Message }
//   PromptForwarded{ SessionID, PromptID, PromptText, PromptType }
//   HeartbeatReceived{ SessionID, Progress *[]ProgressItem }
//   SessionTerminated{ SessionID, ExitCode *int, Reason }
//   StreamActivity{ SessionID }   // resets the stall inactivity timer (S063)

type PermissionOption struct { OptionID, Name, Kind string } // kind: allow_once|allow_always|reject_once|reject_always

type AgentDriver interface {
    ResolveClearance(ctx context.Context, requestID string, approved bool, reason *string) error
    SendPrompt(ctx context.Context, sessionID, prompt string) error
    Interrupt(ctx context.Context, sessionID string) error            // idempotent
    ResolvePrompt(ctx context.Context, promptID, decision string, instruction *string) error
    ResolveWait(ctx context.Context, sessionID string, instruction *string) error
}
```

Drivers emit `AgentEvent`s onto a shared channel; a single consumer goroutine
dispatches them (create/persist approval, post to Slack, register for reply routing).
This mirrors the Rust `run_acp_event_consumer`.

## 6. Domain models & DB schema

Match the schema **exactly** (7 tables). Timestamps are TEXT (RFC3339). Booleans are
INTEGER 0/1. Reproduce this DDL (from `src/persistence/schema.rs`), including the
`CHECK` constraints and indexes, and keep the idempotent additive-migration pattern
(`CREATE TABLE IF NOT EXISTS` + `PRAGMA table_info` column adds):

```sql
session(id PK, owner_user_id, workspace_root,
        status CHECK IN ('created','active','paused','terminated','interrupted'),
        prompt, mode CHECK IN ('remote','local','hybrid'),
        created_at, updated_at, terminated_at, last_tool,
        nudge_count INT DEFAULT 0, stall_paused INT DEFAULT 0, progress_snapshot,
        -- additive migrations:
        protocol_mode TEXT NOT NULL DEFAULT 'mcp', channel_id, thread_ts,
        connectivity_status TEXT NOT NULL DEFAULT 'online', last_activity_at,
        restart_of, agent_session_id, title)
approval_request(id PK, session_id, title, description, diff_content, file_path,
        risk_level CHECK IN ('low','high','critical'),
        status CHECK IN ('pending','approved','rejected','expired','consumed','interrupted'),
        original_hash, slack_ts, created_at, consumed_at)
checkpoint(id PK, session_id, label, session_state, file_hashes, workspace_root,
        progress_snapshot, created_at)
continuation_prompt(id PK, session_id, prompt_text,
        prompt_type CHECK IN ('continuation','clarification','error_recovery','resource_warning'),
        elapsed_seconds, actions_taken, decision, instruction, slack_ts, created_at)
stall_alert(id PK, session_id, last_tool, last_activity_at, idle_seconds,
        nudge_count INT DEFAULT 0,
        status CHECK IN ('pending','nudged','self_recovered','escalated','dismissed'),
        nudge_message, progress_snapshot, slack_ts, created_at)
steering_message(id PK, session_id, channel_id, message,
        source CHECK IN ('slack','ipc'), created_at, consumed INT DEFAULT 0, origin_session_id)
task_inbox(id PK, channel_id, message, source CHECK IN ('slack','ipc'), created_at, consumed INT DEFAULT 0)
```

Indexes: `idx_session_channel(channel_id,status)`, `idx_session_channel_thread(channel_id,thread_ts)`,
`idx_approval_session`, `idx_checkpoint_session`, `idx_prompt_session`, `idx_stall_session`,
`idx_steering_session_consumed(session_id,consumed)`, `idx_inbox_channel_consumed(channel_id,consumed)`.

Repos to port (`src/persistence/`): `session_repo`, `approval_repo`, `checkpoint_repo`,
`prompt_repo`, `stall_repo`, `steering_repo`, `inbox_repo`, `intercom_queue_repo`,
plus `db` (connect; use in-memory SQLite for tests) and `retention` (prune by `retention_days`).
Domain model files (`src/models/`): approval, checkpoint, inbox, intercom_queue, policy,
progress, prompt, session, stall, steering.

## 7. Operator-facing capability surface (feature inventory)

Every item below must have a Go equivalent. Grouped by subsystem.

**ACP session lifecycle** (`/arc` slash prefix)
* `session-start <prompt>` — spawn agent, ACP handshake, create session row, post
  "session started" as the Slack thread root; enforce `max_concurrent_sessions`.
* `session-stop [id]`, `session-restart [id]` (sets `restart_of`), `session-cleanup`, `sessions` (list).
* Per-session **threading** — all session messages threaded under the root (`thread_ts`).
* **Workspace routing** — resolve incoming channel → `[[workspace]]` entry → agent cwd (`path`);
  channel_id must be unique per workspace (enforced at startup **and** on the hot-reload snapshot).

**Approval workflow**
* `ask_approval` → create `approval_request`, post Slack card (Accept/Reject + diff block).
* `accept_diff` / clearance resolution → apply the unified diff **atomically within workspace root**;
  reject on `original_hash` divergence (`PatchConflict`).
* **Terminal-command gate** — destructive commands require operator approval.
* `check_auto_approve` — evaluate workspace policy auto-approve rules; auto-approve suggestion flow.
* Risk levels: `low` / `high` / `critical`.

**Continuation prompts, steering, standby (US17)**
* `forward_prompt` (transmit) — continuation prompt; operator decides Continue / Refine / Stop.
* `wait_for_instruction` (standby) — operator resumes with instructions.
* **US17 text-only thread @-mention replies:** when a session thread exists, `transmit`/`standby`/
  approval post **plain text** and await `@agent-intercom <keyword> …` replies:
  * transmit/forward: `continue` | `refine <text>` | `stop` (empty → `continue`)
  * standby/wait: `resume <text>` | `stop`
  * approval/clearance: `approve` | `reject <reason>`
  * first word = decision keyword (case-insensitive); rest = instruction/reason; only the session
    owner's mention is honored. Main-channel (non-thread) messages keep block-kit buttons.
* **Modal instruction capture** — Refine/Resume open a Slack modal for free-text
  (with the documented thread-`trigger_id` suppression caveat).
* **Steering queue** — operator messages persisted (`steering_message`), delivered on heartbeat;
  at-least-once delivery (mark consumed only after successful send). Scoped per owning session;
  survives crash/restart via `origin_session_id` rebind.

**Task inbox** — channel-scoped operator tasks (`task_inbox`), consumed flag.

**Stall detection / nudge / escalation**
* Inactivity threshold → `stall_alert`; auto-nudge (`send_prompt`) after escalation interval;
  `max_retries` then escalate to operator; `StreamActivity` resets the timer; pause/resume.

**Observability & lifecycle**
* `heartbeat` (ACP: mirrors MCP heartbeat) — liveness + progress snapshot + steering delivery.
* `remote_log` / `broadcast` — status messages to Slack (info/warning/success levels).
* `set_operational_mode` / `switch_freq` — remote vs local mode toggle.
* `recover_state` — reconnect/repost pending messages after disconnect (ADR-0011); crash respawn+resume.
* **Audit logging** — append-only trace of operator/agent actions.
* **Slack detail levels** — minimal / standard / verbose (`slack_detail_level`).
* `ping` — connectivity check (returns `pending_steering`).

**Companion CLI (`agent-intercom-ctl`)** — talks to the server over IPC (named pipe/UDS)
to inject steering/tasks and query state locally.

## 8. Protocol contracts

**ACP (port first).** Spawn `host_cli` with `host_cli_args`; JSON-RPC over stdio. See
`src/acp/{spawner,handshake,reader,writer,codec}.rs` and `src/driver/acp_driver.rs` for exact
wire shapes. Key facts to preserve:
* Envelope `id` is string-or-number; **preserve the raw id type** for JSON-RPC `result` replies
  (correlation), while using a rendered key internally.
* Outbound writes are **serialized per session** (one writer goroutine per session consuming a
  channel of frames) so concurrent sends interleave whole frames, never corrupt the stream.
* Two permission reply shapes: bespoke `clearance/response` and standard ACP
  `session/request_permission` result with nested `result.outcome.outcome` = selected/optionId or
  cancelled (ADR-0016).
* Methods in play include `session/prompt`, `session/interrupt`, `session/request_permission`,
  `prompt/*`, heartbeat/status/update events. Confirm names against `reader.rs`/`writer.rs`.

**Slack.** Block Kit cards (approval, prompt, session-started), modals (`views.open`), interactive
button `action_id`s, and Socket Mode envelopes. See `src/slack/{blocks,client,commands,events,push_events}.rs`
and `src/slack/handlers/*`. Handlers to port: `approval`, `command_approve`, `modal`, `nudge`,
`prompt`, `steer`, `task`, `thread_reply`, `wait`.

**MCP (legacy).** 10 tools in `src/mcp/tools/`: `accept_diff`, `ask_approval`, `check_auto_approve`,
`forward_prompt`, `heartbeat`, `recover_state`, `remote_log`, `set_operational_mode`,
`wait_for_instruction` (+ `util`). Only port if MCP-mode is retained; the strategic direction is ACP-only.

## 9. Config schema (`config.toml`)

Top-level: `default_workspace_root`, `http_port`, `ipc_name`, `max_concurrent_sessions`,
`retention_days`, `host_cli`, `host_cli_args[]`, `slack_detail_level` (minimal|standard|verbose).
Sections: `[database] path`; `[slack]` (empty — creds from env/keychain); `[timeouts]`
(`approval_seconds`, `prompt_seconds`, `wait_seconds`); `[stall]` (`enabled`,
`inactivity_threshold_seconds`, `escalation_threshold_seconds`, `max_retries`, `default_nudge_message`);
`[commands]` (e.g. `status`); `[acp]` (`max_sessions`, `startup_timeout_seconds`, `http_port`);
and repeated `[[workspace]]` (`workspace_id`, `channel_id`, `label`, `path`).
Credentials from env (`SLACK_BOT_TOKEN`, `SLACK_APP_TOKEN`, `SLACK_TEAM_ID`, `SLACK_MEMBER_IDS`;
ACP tries `*_ACP`-suffixed first) or OS keychain (service `agent-intercom` / `agent-intercom-acp`).
ACP mode validates `host_cli` and each `[[workspace]].channel_id` uniqueness.

## 10. Cross-cutting rules

* **Error taxonomy** (`AppError`, 14 variants) → a Go error set (sentinel errors + `errors.Is/As`
  or a typed error): Config, Db, Slack, Mcp, Diff, Policy, Ipc, PathViolation, PatchConflict,
  NotFound, Unauthorized, AlreadyConsumed, Io, Acp. Messages lowercase, prefixed (`config: …`).
* **Workspace isolation (non-negotiable):** every FS op must resolve inside the configured
  workspace root; reject `..`/symlink/absolute escapes (`PathViolation`). Port `diff/` path checks faithfully.
* **Safety discipline:** Rust enforced `#![forbid(unsafe_code)]` + no `unwrap`/`expect`. Go analog:
  no ignored errors, `go vet` + `staticcheck` clean, **always run tests with `-race`** (Go's only
  data-race defense — this replaces Rust's compile-time `Send`/`Sync` guarantee), and use
  `context.Context` for all cancellation.
* **Atomic, git-friendly persistence:** SQLite is the source of truth; keep the idempotent
  startup-migration pattern so the DB converges on every boot.

## 11. Parity oracle — the test corpus

The Rust tests are the acceptance criteria. Re-express them in Go and make them pass.

| Tier | Rust location | Count | Port intent |
|---|---|---|---|
| Unit | `tests/unit/` | 56 files | Isolated logic (config, models, policy, stall, routing, commands). |
| Contract | `tests/contract/` | 22 files | **Highest value** — exact tool/response/driver contracts. Start here. |
| Integration | `tests/integration/` | 46 files | End-to-end flows (sessions, approvals, steering, threading, workspace routing, recovery) with in-memory SQLite. |
| Live Slack | `tests/live/` | 6 files | Real Slack Web API (feature-gated) — port as build-tagged tests. |
| Visual (Playwright) | `tests/visual/scenarios/` | 13 specs | Slack web rendering — **reusable as-is** against the Go server (they drive Slack, not Rust). |

Key contract files to mirror first: `driver_contract_tests`, `acp_driver_contract_tests`,
`acp_event_contract`, `acp_capacity_contract`, `ask_approval_contract_tests`,
`auto_check_contract_tests`, `prompt_contract_tests`, `wait_contract_tests`,
`recover_state_tests`, `schema_tests`, `tool_names_tests`, `ping_contract_tests`.

**Front-load-the-risk slices (do these first):**
1. **Slack Socket Mode** via `slack-go/slack` — post an approval card + open a modal + handle an
   interaction + thread a reply + survive a reconnect. Gate: parity with `slack/handlers/*`.
2. **ACP subprocess** — spawn `host_cli --acp`, handshake, per-session writer goroutine,
   reader → `AgentEvent` channel, resolve a permission both reply-shapes. Gate: `acp_*` contracts.
3. **SQLite via `modernc.org/sqlite`** — bootstrap the schema, run the repo query set, confirm
   latency is acceptable. Gate: `schema_tests` + repo round-trips + CGO-free `go build`.

## 12. Phased port plan

1. **Skeleton + risk slices** — repo, CI (build/vet/staticcheck/test `-race`/govulncheck), config,
   store (schema + repos), and the three Section-11 risk slices proven.
2. **Driver + ACP core** — `AgentDriver` interface, ACP driver, spawner/reader/writer, event consumer.
3. **Slack surface** — Socket Mode client, block kit, slash commands, all interaction handlers, US17.
4. **Approval + policy + diff** — approval workflow, terminal-command gate, auto-approve policy,
   atomic in-root diff apply, audit.
5. **Resilience** — stall detection/nudge/escalation, steering queue (at-least-once), reconnect/repost,
   crash respawn+resume, retention.
6. **IPC + ctl** — companion CLI over UDS/named pipe.
7. **Parity sweep** — port remaining unit/integration/live tests; run the shared Playwright suite
   against the Go server; compare behavior to Rust for each contract.
8. **Cutover decision** — only when parity is demonstrated; Rust remains fallback until then.

## 13. Go-specific watch-outs

* **Data races:** Go won't catch them at compile time. Serialize per-session stream writes via a
  channel + single writer goroutine (as Rust does); guard shared maps with mutexes; test with `-race`.
* **Nil safety:** replace Rust `Option<T>` with pointers/`ok` returns deliberately; avoid nil-deref.
* **Error handling:** don't drop errors; wrap with `%w`; map to the Section-10 taxonomy.
* **Cancellation/shutdown:** thread `context.Context` from `main` through spawn/stream/HTTP for
  graceful shutdown (Rust used `CancellationToken`).
* **Keep it CGO-free:** verify `CGO_ENABLED=0 go build` succeeds and cross-compiles to all 4 targets
  (linux-amd64, windows-amd64, darwin-amd64, darwin-arm64) from one runner — that is the whole point.

## References

* Rust source: `src/` (~21.9k LOC), `ctl/`, `tests/` (~30.5k LOC)
* Spike decision: `docs/decisions/2026-07-06-rust-vs-go-intercom-spike.md`
* ADRs: `docs/adrs/` (esp. driver-trait, ACP correctness, reconnect-repost, stdio-resume, ADR-0016 permission)
* Architecture: `docs/ARCHITECTURE.md`

---
title: "intercom-go Port Continuation — Oracle Correction, MCP Retirement, and the Next Bounded Slice"
description: "Deliberate the dependency-linked group 4989A42D + 037B1552, resolve the MCP-retirement blocker against the real Rust oracle, and select the next bounded implementation slice"
topic: "What is the next shippable slice of the agent-intercom Go port, and is MCP retired for this port?"
depth: "deep"
decision_status: "decided"
promoted_to: "plan"
date: 2026-09-03
linked_artifacts:
  - "docs/plans/2026-09-03-intercom-go-apperr-pathsafe-plan.md"
  - "docs/decisions/2026-07-06-go-port-reference-brief.md"
  - "docs/decisions/2026-09-03-intercom-go-architecture-reconciliation-deliberation.md"
  - "docs/closure/2026-09-03-intercom-go-foundation-closure.md"
source_documents:
  - path: "docs/decisions/2026-07-06-go-port-reference-brief.md"
    stash_id: "4989A42D"
    role: "authoritative port spec (Group A)"
  - path: "docs/decisions/2026-07-06-go-port-reference-brief.md"
    stash_id: "037B1552"
    role: "deferred config/credential/error taxonomy slice (Group A)"
behavioral_oracle:
  repo: "softwaresalt/agent-intercom"
  path: "C:/Source/GitHub/intercom"
  head: "41df772"
  access: "read-only, out-of-tree, never committed"
tags:
  - "go"
  - "port"
  - "acp"
  - "mcp-retirement"
  - "staging"
---

# intercom-go Port Continuation — Oracle Correction, MCP Retirement, and the Next Bounded Slice

## Problem Frame

Shipment `001-S` (foundation: module skeleton + hardened CI) is shipped, closed, and
merged to `main` at `87b800d5d3a6457d86fecd12fb7a299f170af718`. The backlog, queue,
and shipment list are empty. This session must select and stage the next slice.

The operator nominated two stash entries as one dependency-linked group:

* **`4989A42D`** — the remaining feature-complete CGO-free Go port of `agent-intercom`.
* **`037B1552`** — config package, credential resolution, and the `AppError` taxonomy,
  deferred out of the foundation slice by plan review.

`037B1552` carried one open blocker: **"MCP-retirement confirmation."** This session
must resolve it with an explicit, non-vague decision.

Three items are explicitly excluded from this shipment by operator instruction:
`A92E3FA0` (multiplexer backend — architecturally incompatible, see the prior
reconciliation deliberation), the six foundation hardening follow-ups
(`EFFAA358`, `90350C9A`, `F7C6420D`, `BEDD2E70`, `35D76D5E`, `90EE7758`), and
`EF9352FB` (operator credential rotation — an operator action, not a code change).

### Constraints

* CGO-free static binary is non-negotiable ("that is the whole point" — brief §13).
* The Rust repository is the behavioral oracle; where brief and Rust source disagree,
  **the Rust source and its test corpus win** (brief §1.3).
* Every task ≤ 2 hours of human-equivalent effort.
* Stage may not write product code.
* The decomposition must be safe for a later **dark-factory** (unattended) execution
  trigger: small reviewable shipments, explicit dependency edges, deterministic gates,
  no parallel implementation worktrees, each phase independently verifiable.

### Success Criteria

* The MCP blocker is closed with a recorded, source-grounded decision.
* The next slice is bounded (not an oversized shipment) and dependency-rooted.
* Remaining port phases are preserved as successor stash entries with explicit ordering.
* Provenance links to brief, prior deliberation, oracle commit, and foundation closure.

---

## Research Findings

### Prior learnings

`docs/compound/` contains two durable learnings from the `001-S` cycle:

* `external-spec-yields-to-workspace-instructions-2026-09-03.md`
* `go-race-requires-cgo-fails-closed-2026-09-03.md`

The second is directly load-bearing: `go test -race` requires CGO, which collides with
the CGO-free mandate. The foundation shipment resolved this by running `-race` as a
dedicated CI job with CGO enabled while keeping release builds `CGO_ENABLED=0`. Every
slice planned below inherits that gate topology unchanged. The first is a process
learning (workspace instructions outrank an external spec on conflict) and is applied
directly in the MCP decision below.

### F1 — [P0] The mounted "Rust oracle" is the wrong repository

The session brief and both stash entries assert:

> "ORACLE MOUNT RESOLVED 2026-09-03: the read-only Rust behavioral oracle required by
> brief section 1.2 is now available locally at `references\herdr` (clean at `b07ba9ce`)."

**This is false.** Verified directly:

| Probe | Result |
|---|---|
| `git remote -v` in `references/herdr` | `https://github.com/ogulcancelik/herdr.git` |
| `Cargo.toml` `name` / `description` | `herdr` — "terminal workspace manager for AI coding agents" |
| `keywords` | `terminal, tui, ai, agents, multiplexer` |
| Occurrences of `agent-intercom`, `AgentDriver`, `clearance`, `socketmode`, `ServerMode`, `stall_alert`, `steering_message`, `ask_approval` in `src/` | **0 each** |
| Occurrences of `acp`, `ACP`, `jsonrpc`, `session/prompt`, `request_permission` in `src/` | **0 each** |
| `src/` layout | `pty/`, `pane/`, `client/shell/`, `api/`, `protocol/`, `terminal/` — no `slack/`, `acp/`, `driver/`, `persistence/`, `mcp/` |
| Test corpus | 25 `.rs` files in `tests/{cli,fixtures,support}` |

`herdr` is a Rust terminal multiplexer/TUI. It is **not** `softwaresalt/agent-intercom`
and cannot serve as the behavioral oracle for this port. Notably, `herdr` *is*
thematically adjacent to stash `A92E3FA0` (the Copilot Remote Multiplexer TUI backend),
which is the most plausible explanation for why it was mounted — but `A92E3FA0` is
explicitly out of scope for this shipment.

### F2 — [P0] The real oracle exists and is already reachable

Brief §1.2 offers two mount options, the second being: *"launch the Go session with
`../intercom` provided as explicit read context."* That path resolves, and it is correct:

| Probe | Result |
|---|---|
| `git remote -v` in `C:\Source\GitHub\intercom` | `https://github.com/softwaresalt/agent-intercom.git` |
| `HEAD` | `41df772` — *"feat(config): enforce unique channel_id in ACP mode validation (#36)"* |
| `Cargo.toml` `description` | "MCP remote agent server — review and approve AI agent code changes via Slack" |
| `src/` top level | `acp/ audit/ diff/ driver/ ipc/ mcp/ models/ orchestrator/ persistence/ policy/ slack/ state/ config.rs config_watcher.rs errors.rs mode.rs` |
| `src/` LOC | **21,762** (brief §References says "~21.9k LOC") |
| Test corpus | **56 unit / 22 contract / 46 integration / 6 live** |

The module map matches brief §4 entry-for-entry and the test corpus matches brief §11
**exactly** (56/22/46/6). This is conclusively the authoritative oracle.

All research below is grounded in `C:\Source\GitHub\intercom` @ `41df772`, used
strictly read-only and entirely outside this repository's working tree. Its working
tree carries 27 pre-existing uncommitted operator changes which were **not** touched.

### F3 — [P0] MCP is *not* retired in the oracle — but the port still should not carry MCP mode

This is the blocker resolution. The evidence splits cleanly into three distinct surfaces
that the word "MCP" has been conflating.

**Surface 1 — MCP as a server mode (`--mode mcp`, stdio transport, `McpDriver`).**
Still present and still the *default*:

```rust
// src/mode.rs:13-20
#[derive(Debug, Copy, Clone, Default, Eq, PartialEq, ValueEnum, Serialize, Deserialize)]
#[serde(rename_all = "snake_case")]
pub enum ServerMode {
    #[default]
    Mcp,
    Acp,
}
```

`--mode` defaults to `ServerMode::Mcp` (`src/main.rs`), and `McpDriver` is constructed
on every default run. `src/mcp/` is 4,384 LOC across 18 files and is live, not dead code.

**Surface 2 — the MCP *tool endpoint over HTTP*.** This is the decisive discovery.
It runs in **both** modes, and in ACP mode it is how the spawned agent subprocess
performs human-in-the-loop calls:

```rust
// src/main.rs:365-372
// The HTTP transport starts in BOTH MCP and ACP modes. In ACP mode,
// the endpoint lets agent subprocesses call MCP tools (check_clearance,
// transmit, auto_check, etc.) via HTTP. Without it, tools are unreachable
// from spawned ACP sessions (HITL-003 / FR-032).
let start_stdio = args.mode == ServerMode::Mcp
    && matches!(args.transport, Transport::Stdio | Transport::Both);
let start_sse = matches!(args.transport, Transport::Sse | Transport::Both);
```

`start_sse` has **no mode guard**. The registered wire tool names
(`src/mcp/handler.rs:224-401`) are `check_clearance`, `check_diff`, `auto_check`,
`transmit`, `broadcast`, `reboot`, `switch_freq`, `standby`, `ping` — note these are
the *wire* names and differ from the `src/mcp/tools/*.rs` filenames the brief §8 lists
(`accept_diff`, `ask_approval`, `check_auto_approve`, …). Brief §8 names the modules,
not the protocol surface. `tests/contract/tool_names_tests.rs` pins the wire names.

**Surface 3 — `protocol_mode` as persisted session state.** Load-bearing and pervasive:

```sql
-- src/persistence/schema.rs:48-49
ALTER TABLE session ADD COLUMN protocol_mode TEXT NOT NULL DEFAULT 'mcp'
```

`ProtocolMode::Mcp` is the model default (`src/models/session.rs:165`); ACP capacity
counting filters on `protocol_mode = 'acp'` (`src/persistence/session_repo.rs:473-488`);
and `stall_consumer.rs:232`, `slack/blocks.rs:325,335`, `slack/commands.rs:652,751,1128,1289`,
`slack/handlers/steer.rs:97,159` all branch on it.

**Retirement is planned but not executed.** The oracle's own backlog epic `013-F`
states as *future* acceptance criteria: *"ACP is the default mode; `src/mcp/*` and
`rmcp` removed"*. `docs/adrs/0018-keep-agentdriver-trait-after-mcp-removal.md` is an
**Accepted** ADR that pre-decides the post-removal architecture and assigns the deletion
to task `013.005.003-T`. That deletion has not landed at `41df772`.

### F4 — The config contract is materially richer and partly different from brief §9

Brief §9 is a summary, and the oracle contradicts it in ways that matter. Selected deltas:

| Brief §9 says | Oracle at `41df772` | Impact |
|---|---|---|
| `[slack]` is "(empty — creds from env/keychain)" | `SlackConfig` has TOML keys `channel_id` and `markdown_upload_extensions`; only `app_token`/`bot_token`/`team_id` are `#[serde(skip)]` | Confirms stash finding **P1-i**; a strict decoder would reject a real `config.toml` |
| "Credentials from env … or OS keychain" | Order is **keychain first, then env**, within a mode-scoped chain | Stash `037B1552`'s phrasing ("env with `*_ACP` precedence + OS keychain fallback") is **backwards** |
| `[acp]` = `max_sessions`, `startup_timeout_seconds`, `http_port` | also `max_msg_rate` (default `10`) | Missing key |
| — | **No `deny_unknown_fields` anywhere**; unknown keys are silently ignored | Go must **not** use strict decoding |
| "ACP mode validates `host_cli`" | `host_cli` is required by TOML shape in *all* modes; the *non-empty* + existence checks run only in `validate_for_acp_mode()` | Refines stash finding **P1-g** |

The exact credential precedence (`src/config.rs::load_credential`), ACP mode:

1. keychain service `agent-intercom-acp`, key `{keyring_key}`
2. env `{ENV_KEY}_ACP`
3. keychain service `agent-intercom`, key `{keyring_key}`
4. env `{ENV_KEY}`

MCP mode uses steps 1–2 only with an empty suffix (i.e. service `agent-intercom`, env
`{ENV_KEY}`). Empty values are treated as absent at every step. Secrets never enter
error messages — errors name the *source* (service/env-var name), never the value —
and `SlackConfig` has a hand-written `Debug` impl printing `[REDACTED]` for
`app_token`/`bot_token` (but **not** `team_id`).

**Defaults that are not Go zero values** (a naive Go decode silently corrupts these):
`stall.enabled=true`, `timeouts.approval_seconds=3600`, `timeouts.prompt_seconds=1800`,
`max_concurrent_sessions=3`, `retention_days=30`, `ipc_name="agent-intercom"`,
`http_port=3000`, `acp.http_port=3001`, `acp.max_sessions=5`,
`acp.startup_timeout_seconds=30`, `acp.max_msg_rate=10`,
`stall.inactivity_threshold_seconds=300`, `stall.escalation_threshold_seconds=120`,
`stall.max_retries=3`, `slack_detail_level="standard"`, `database.path="data/agent-rc.db"`.
This is exactly stash finding **P1-h** and it is confirmed as real.

### F5 — The `AppError` taxonomy is confirmed at 14 variants, with two corrections

The brief's variant list is **correct and complete**: `Config`, `Db`, `Slack`, `Mcp`,
`Diff`, `Policy`, `Ipc`, `PathViolation`, `PatchConflict`, `NotFound`, `Unauthorized`,
`AlreadyConsumed`, `Io`, `Acp`. Every variant carries a single `String` and renders as
`"{lowercase prefix}: {msg}"` with no trailing period. It is **hand-rolled**, not
`thiserror`-derived; the three `From` impls (`toml::de::Error`, `sqlx::Error`,
`std::io::Error`) stringify the source immediately, discarding the error chain.

Two corrections to naive expectations:

* **`PatchConflict` is defined but never constructed** anywhere in `src/`. The conflict
  path returns an MCP-level `error_result("patch_conflict", …)` instead. It is vestigial.
* **There is no central boundary mapping.** MCP tools build `rmcp::ErrorData` ad hoc,
  IPC uses its own `IpcResponse{ok,error}`, and `src/mcp/sse.rs` sets `axum::StatusCode`
  directly. There is no `IntoResponse` for `AppError`. This validates the stash's own
  suggestion to prefer package-local errors translated at boundaries.

`PathViolation` is produced by `src/diff/path_safety.rs`, **not** by `errors.rs`. Its
algorithm — the security control this port must reproduce faithfully — is:

1. `canonicalize()` the workspace root; failure → `"workspace root invalid: {err}"`.
2. Fast path: candidate already absolute **and** `starts_with(root)` → if it exists,
   `canonicalize()` and re-check containment, else accept as-is.
3. Component walk: `..` pops the stack (empty stack → `"path attempts to escape workspace"`);
   `.` skipped; `RootDir`/`Prefix` → `"absolute paths are not allowed; use workspace-relative paths"`;
   `Normal` pushed.
4. `root.join(normalized)`.
5. Containment check → else `"path outside workspace"`.
6. If the path exists, `canonicalize()` and re-check → else `"symlink target escapes workspace"`.
7. Non-existent paths get lexical normalization + prefix check **only** — symlinks are
   resolved for existing paths only. This is an inherent TOCTOU gap in the oracle.

Pinned by `tests/unit/path_validation_tests.rs` (9 tests) and
`tests/unit/error_tests.rs` (7 tests).

### F6 — Subsystem sizing and dependency structure for the roadmap

| Subsystem | LOC | Internal deps | Tier |
|---|---|---|---|
| `errors.rs`, `mode.rs`, `lib.rs` | 328 | — | 0 (leaf) |
| `models/` | 865 | self only | 0 (leaf) |
| `diff/` (incl. `path_safety.rs`) | 258 | errors | 0 (leaf) |
| `audit/` | 226 | errors | 0 (leaf) |
| `config.rs` | 873 | errors | 0 (leaf) |
| `policy/` | 408 | models | 1 |
| `persistence/` | 2,396 | models, errors | 1 |
| `driver/` | 914 | models, state | 2 (cycle) |
| `state/` | 161 | audit, config, driver, mode, policy, slack, orchestrator | 2 (hub, cycle) |
| `acp/` | 1,709 | driver, models, persistence, slack/client | 3 |
| `orchestrator/` | 1,455 | config, models, persistence, slack, state, driver | 3 |
| `ipc/` | 373 | models, persistence, slack/handlers, state | 3 |
| `slack/` | 5,759 | ~10 subsystems | 4 (hub) |
| `mcp/` | 4,384 | ~10 subsystems | 4 (hub) |
| `ctl/` | 180 | none (IPC client only) | independent |

**Coupling hazards** that constrain phase boundaries:

* `state ↔ slack` is a genuine import cycle (`state/mod.rs` imports `slack::client` and
  re-exports `slack::handlers::thread_reply`; `slack/events.rs` and `slack/commands.rs`
  import `state::AppState`). Go forbids this; it must be broken with interfaces. Note
  `driver/` is **not** in the cycle — only the retired `mcp_driver.rs:18` imports `state`.
* `slack_ts` columns appear in the `approval_request`, `continuation_prompt`, and
  `stall_alert` tables; `channel_id`/`thread_ts` are session columns. Persistence cannot
  be shipped Slack-agnostic.
* `slack/commands.rs` is a 1,651-LOC mega-module importing `acp`, `diff`, `driver`,
  `orchestrator`, `persistence`, `policy` — it cannot ship before nearly everything else.
* `acp/reader.rs:11` imports `slack::client`, coupling the protocol layer to notification.

**Concurrency hotspots** (Go race risk): ~18 `tokio::spawn` sites, ~12 `Arc<Mutex<HashMap>>`
maps (6 in `AcpDriver` alone), 4 `mpsc` channels, 3 `oneshot` types. The per-session
ACP reader/writer goroutine pair and `AcpDriver`'s six concurrent maps are the highest risk.

---

## Options Evaluated

### Decision 1 — How to treat the group `4989A42D` + `037B1552`

#### Option 1A: Plan `4989A42D` and `037B1552` together as one covering feature

* **Pros:** Matches the operator's "one dependency-linked group" framing literally.
* **Cons:** `4989A42D` is the *entire remaining port* — brief §12 phases 1–8, roughly
  19,000 LOC of Rust plus a 130-file test corpus. Harvesting it produces an oversized
  shipment, which the operator explicitly forbade.
* **Fit:** Poor as a *harvest* unit.

#### Option 1B: Treat the group as one **planning** unit, harvest only its dependency root

Deliberate and plan both entries together (they are genuinely dependency-linked —
`037B1552` is the extracted first phase of `4989A42D`), but harvest only the bounded
root slice, and rewrite both stash entries with explicit successor ordering.

* **Pros:** Honours the dependency link where it matters (a single coherent roadmap)
  while keeping the shipment small and reviewable. Directly satisfies the operator's
  "harvest only the next bounded implementation slice" instruction.
* **Cons:** Requires disciplined stash rewriting so nothing is lost.
* **Fit:** Strong.

**Selected: 1B.**

### Decision 2 — MCP retirement for this port (the blocker)

#### Option 2A: Port MCP fully (mode switch, `McpDriver`, stdio transport, all tools)

* **Pros:** Maximum oracle fidelity; `--mode mcp` deployments keep working.
* **Cons:** ~4,384 LOC of a surface the oracle's own Accepted ADR-0018 and epic `013-F`
  commit to deleting. Porting code that is scheduled for deletion upstream is waste.
* **Fit:** Poor.

#### Option 2B: Drop everything named "MCP" — no `internal/mode`, no `internal/driver/mcp`, no `internal/mcp`

This is what deliberation question 4 and stash `037B1552` implicitly proposed.

* **Pros:** Smallest surface; matches the brief's ACP-only strategic direction.
* **Cons:** **Breaks ACP mode.** `src/main.rs:365-372` proves the MCP HTTP tool endpoint
  is how spawned ACP subprocesses reach `check_clearance`/`transmit`/`auto_check`. Drop
  it and the human-in-the-loop path — the product's entire reason for existing — is
  unreachable from an ACP session. Also silently discards the `protocol_mode` column,
  which ACP capacity counting depends on.
* **Fit:** Poor, and dangerous. This option looks correct from the brief alone and is
  only falsified by the oracle. This is precisely the risk the `037B1552` deferral existed
  to catch.

#### Option 2C: Retire MCP *as a server mode*; retain the MCP *tool endpoint* as the ACP agent-facing HITL surface

Split the three surfaces identified in F3 and decide each on its own merits.

* **Pros:** Preserves the load-bearing HITL path, drops genuinely dead weight, and aligns
  with where the oracle is *going* (ADR-0018 / epic `013-F`) rather than where it *is*.
  Keeps the port strictly smaller than the oracle without breaking it.
* **Cons:** Requires carrying a `protocol_mode` column whose value is effectively constant,
  and requires naming discipline so "MCP" is not overloaded in the Go tree.
* **Fit:** Strong.

#### Option 2D: Defer the decision again

* **Pros:** None beyond avoiding commitment.
* **Cons:** The blocker is already stale, the evidence is now decisive, and deferring
  blocks `037B1552` indefinitely. The operator explicitly asked for a non-vague decision.
* **Fit:** Poor.

**Selected: 2C.**

### Decision 3 — What is the next bounded slice?

Candidates, all drawn from `037B1552`'s scope:

| Option | Content | Est. tasks | Deps | Risk |
|---|---|---|---|---|
| 3A | All of `037B1552` (errors + path safety + config + credentials) | 12–14 | — | Oversized shipment; mixes a security-sensitive credential surface with routine decode work |
| 3B | `internal/apperr` + `internal/pathsafe` only | 5 | none | Dependency root; both Tier-0 leaves; 16 oracle tests; no secrets |
| 3C | `internal/config` schema + validation | 5–6 | needs `apperr`, `pathsafe` | Blocked until 3B lands |
| 3D | Credential resolution + `Secret` type | 4 | needs `config` | Security-sensitive; needs its own hardening pass |

`apperr` is the true dependency root: `AppError` has ~470 construction sites across the
oracle's `src/`, and every config validation rule is *expressed as* an `AppError::Config`.
`pathsafe` belongs with it because stash finding **P1-e** explicitly scopes
"canonicalize + EvalSymlinks workspace roots and emit `PathViolation`" into this slice,
and because `config.validate()` needs root canonicalization before it can be written.

**Selected: 3B.** 3C and 3D become ordered successors.

---

## Decision

**Plan the group as one roadmap (1B); retire MCP-as-a-mode while retaining the MCP tool
endpoint (2C); harvest `internal/apperr` + `internal/pathsafe` as the next shipment (3B).**

### D1 — MCP retirement: the binding resolution

The blocker "MCP-retirement confirmation" is **CLOSED**. The decision is per-surface:

| Surface | Oracle state @ `41df772` | Decision for intercom-go | Rationale |
|---|---|---|---|
| **`ServerMode` enum / `--mode` flag** | present, `Mcp` is `#[default]` | **RETIRE.** No `internal/mode`. The Go server is ACP-only; no `--mode` flag. | Brief §2/§8 direct ACP-only; oracle ADR-0018 (Accepted) and epic `013-F` commit to removal. A greenfield port should not reproduce a surface upstream is deleting. |
| **`McpDriver`** | present, constructed on default runs | **RETIRE.** No `internal/driver/mcp`. `AgentDriver` keeps exactly one implementation (ACP). | ADR-0018 explicitly keeps the `AgentDriver` trait *after* `McpDriver` deletion, so the seam survives without the second impl. |
| **stdio MCP transport** | `--transport stdio\|both`, MCP-mode only | **RETIRE.** No stdio JSON-RPC server transport. | Guarded by `args.mode == ServerMode::Mcp`; unreachable once mode is retired. |
| **MCP *tool endpoint over HTTP*** | runs in **both** modes; ACP subprocesses' only HITL path | **RETAIN — load-bearing.** Port as `internal/agenttools` (HTTP, served on `acp.http_port`, default `3001`). | `src/main.rs:365-372`: without it, tools are unreachable from spawned ACP sessions (HITL-003 / FR-032). Dropping it breaks the product. |
| **Wire tool names** | `check_clearance`, `check_diff`, `auto_check`, `transmit`, `broadcast`, `reboot`, `switch_freq`, `standby`, `ping` | **RETAIN verbatim.** | Agent-facing contract pinned by `tests/contract/tool_names_tests.rs`. Renaming breaks every configured agent. |
| **`protocol_mode` session column** | `TEXT NOT NULL DEFAULT 'mcp'`; ACP capacity counts on `= 'acp'` | **RETAIN the column** (schema fidelity, brief §6 "match exactly"); **write `'acp'` only**. Keep the reader tolerant of `'mcp'` rows. | Preserves the 7-table schema contract and forward-compatibility with an existing DB, while the port itself never creates MCP sessions. |
| **`AppError::Mcp` variant** | 4 construction sites | **RETAIN** in the 14-variant taxonomy. | The tool endpoint still needs it; and brief §10 pins the variant list. Dropping it would desynchronise the taxonomy from the oracle for no gain. |

**Compatibility surface that remains after retirement:** the nine HTTP wire tools above,
the `acp.http_port` config key, the `protocol_mode` column (read-tolerant / write-`acp`),
and `AppError::Mcp`. **Surface deliberately lost:** `--mode`, `--transport`, stdio
transport, `McpDriver`, and the ability to serve a legacy `--mode mcp` deployment.

**Effect on the config contract:** none additive. `protocol_mode` is *not* a `config.toml`
key in the oracle (it is CLI-only), so the Go config schema omits a mode field entirely —
this is a *simplification*, not a break. `acp.http_port` becomes unconditionally
meaningful rather than mode-dependent. Stash finding **P1-g** ("make `host_cli`
unconditionally required per ACP-only direction") is **upheld and strengthened**: with no
MCP mode, `validate_for_acp_mode()`'s checks become the *only* validation path, so
`host_cli` non-empty + absolute-path-existence are unconditional in the Go port.

**Effect on the CLI contract:** `--mode` and `--transport` are dropped. `--config` and
`--log-level` (already shipped in `cmd/intercom`) are retained. This is a deliberate,
recorded divergence from the oracle's CLI, justified by the ACP-only direction; it is
recorded here so a future parity sweep does not flag it as a regression.

**Confidence: high.** Grounded in direct source reads at `41df772`
(`src/mode.rs:13-20`, `src/main.rs:365-372`, `src/mcp/handler.rs:224-401`,
`src/persistence/schema.rs:48-49`, `src/persistence/session_repo.rs:473-488`) plus the
oracle's own Accepted ADR-0018.

### D2 — Oracle correction

`references/herdr` is **rejected** as the behavioral oracle (F1). The authoritative
oracle is `C:\Source\GitHub\intercom` @ `41df772`, consumed read-only and out-of-tree.
Both stash entries' "ORACLE MOUNT RESOLVED … `references\herdr`" annotations are
factually wrong and are corrected in this session's stash rewrite.

`references/herdr` is left **exactly as found** — untouched, unstaged, uncommitted, and
still excluded via `.git/info/exclude`. It is retained because it is plausibly relevant
to `A92E3FA0` (multiplexer TUI), which is out of scope here.

Because the oracle is out-of-tree, this port has **no in-repo pin** on the oracle commit.
That is a durable traceability gap: a future session cannot verify which oracle revision
a contract was frozen against. Mitigated by recording `41df772` in this deliberation, in
the plan, and in every harvested backlog item's provenance — and by task `002.001.001-ST`
emitting a machine-readable `docs/oracle-pin.md`.

### D3 — Next slice

Harvest `internal/apperr` + `internal/pathsafe`. Defer `internal/config` (successor 1)
and credential resolution (successor 2).

---

## Phased Roadmap

Sequential only. No parallel implementation branches or worktrees (P-016). Each phase is
one shipment, independently verifiable, with a deterministic gate.

| Phase | Scope | Depends on | Gate | Est. |
|---|---|---|---|---|
| **P0** ✅ | Module skeleton, entrypoints, CGO-free build, hardened CI | — | shipped `001-S` | done |
| **P1** ◀ *this session* | `internal/apperr` (14 variants) + `internal/pathsafe` (containment) | P0 | 16 ported oracle tests; `-race`; `CGO_ENABLED=0` build | 5 tasks |
| **P2** | `internal/config` — schema, non-zero defaults, 9 validation rules, workspace routing | P1 | ~53 config + 9 workspace-mapping tests | 5–6 tasks |
| **P3** | Credential resolution — 4-step precedence, keychain, `Secret` type, redaction | P2 | 8 credential tests; no-secret-in-logs proof | 4 tasks |
| **P4** | `internal/models` (10 domain types) + risk slice: SQLite via `modernc.org/sqlite` — 7-table schema + `session_repo` | P2 | `schema_tests`, `session_repo` round-trips, CGO-free build | 6–8 tasks |
| **P5** | Remaining 7 repos + retention | P4 | repo unit tests, `retention_tests` | 5–7 tasks |
| **P6** | Risk slice: ACP subprocess — spawner, handshake, codec, reader, writer, **plus the `Notifier` interface** (see below) | P5 | `acp_codec`, `acp_handshake`, `acp_capacity`, `acp_event` contracts | 8–10 tasks |
| **P7** | `AgentDriver` interface + ACP driver + event consumer | P6 | `driver_contract`, `acp_driver_contract` | 6–8 tasks |
| **P8** | Risk slice: Slack Socket Mode — client, reconnect, block kit; implements `Notifier` | P7 | `slack_client`, `slack_interaction`, `slack_fallback` | 8–10 tasks |
| **P9** | Slack surface — slash commands, modals, all 9 handlers, US17 thread @-mentions | P8 | `blocks_*`, `command_*`, `thread_reply_fallback` | 10–14 tasks |
| **P10** | `internal/diff` apply + approval workflow + terminal gate + `internal/policy` + `internal/audit` | P9 | `accept_diff`, `ask_approval`, `auto_check`, `diff_apply` | 8–10 tasks |
| **P11** | `internal/agenttools` — the retained MCP HTTP tool endpoint (9 wire tools) | P10 | `tool_names_tests`, `ping_contract`, `heartbeat` | 6–8 tasks |
| **P12** | Resilience — stall detect/nudge/escalate, steering at-least-once, reconnect/repost, crash respawn+resume | P11 | `stall_*`, `crash_recovery`, `recover_state` | 8–10 tasks |
| **P13** | `internal/ipc` + `intercom-ctl` companion CLI | P12 | `ipc_server_tests`, `cli_tests` | 4–5 tasks |
| **P14** | Parity sweep — remaining unit/integration/live tests + Playwright against the Go server | P13 | full corpus green | 10–15 tasks |
| **P15** | Cutover decision — Rust remains fallback until parity demonstrated | P14 | operator decision | — |

Ordering derives from the measured import graph (F6), not from the brief's narrative
order. Deliberate departures from brief §12:

* **Persistence (P4/P5) moves ahead of ACP (P6)**, because `acp/` imports
  `persistence::{session_repo, steering_repo, db}` (`src/acp/reader.rs:45-47`). The brief
  lists SQLite as risk slice #3 but the dependency graph makes it a prerequisite.
* **`internal/models` moves into P4, not P5.** `src/persistence/session_repo.rs:8-11`
  imports `models::progress::ProgressItem` and `models::session::{…}`, so P4's
  `session_repo` round-trip gate cannot pass without the session/progress models. Keeping
  models in P5 would have created a circular phase dependency. *(Corrected during plan
  review; see Plan Review finding ARCH-2.)*
* **The MCP tool endpoint becomes P11** (not dropped, not phase 1), per D1.

#### The `acp → slack` edge and the `Notifier` seam

`src/acp/reader.rs:48` contains `use crate::slack::client::{SlackMessage, SlackService};`
— the ACP reader posts Slack notifications directly on reconnect/flush. Taken literally
this would make P6 (ACP) depend on P8 (Slack), inverting the roadmap order.

**Resolution:** P6 defines a minimal consumer-side `Notifier` interface (a single
`Notify(ctx, sessionID, msg) error`-shaped seam) in a Tier-0/1 package and depends on
*that*, not on Slack. P6 ships with a no-op/recording test implementation. P8 supplies
the real Slack implementation. This is the idiomatic Go inversion (define the interface
at the consumer) and it removes the edge entirely rather than reordering phases.
*(Added during plan review; see Plan Review finding ARCH-1.)*

#### Import-cycle characterization (corrected)

The prior characterization "`state ↔ driver ↔ slack`" overstated the cycle. Verified:

* `src/driver/mod.rs` and `src/driver/acp_driver.rs` do **not** import `state` or `slack`.
* Only `src/driver/mcp_driver.rs:18` imports `crate::state::{…}`.
* The real cycle is **`state ↔ slack`**: `state/mod.rs` imports `slack::client::SlackService`
  and re-exports `slack::handlers::thread_reply::PendingThreadReplies`, while
  `slack/events.rs` and `slack/commands.rs` import `state::AppState`.

Because decision **D1 retires `McpDriver`**, the only `driver → state` edge disappears in
the Go port by construction. The Go port therefore has to break exactly one cycle
(`state ↔ slack`), and the `Notifier` seam introduced at P6 is the same mechanism that
breaks it. P7 remains the phase that finalizes the `AgentDriver`/`AppState` boundary.
*(Corrected during plan review; see Plan Review finding ARCH-3.)*

### Dark-factory readiness

The decomposition is designed so a later exact dark-mode trigger can execute safely.
Dark mode is **NOT** activated in this session.

* **Small shipments** — P1 is 5 tasks; no phase exceeds ~14.
* **Explicit dependency graph** — recorded as backlogit dependency edges, not prose.
* **Deterministic gates** — every acceptance criterion is a command with a binary
  outcome (`go test -race ./internal/apperr`, `CGO_ENABLED=0 go build ./...`,
  `golangci-lint run`). No "reviewer judges quality" criteria.
* **Rollback boundary** — every phase is additive net-new packages under `internal/`.
  Through P3 nothing is imported by `cmd/`, so any phase reverts by deleting its package
  directory with zero downstream breakage. P4 onward, revert = revert the merge commit.
* **Compatibility boundary** — the frozen operator-facing contracts are the `config.toml`
  schema (P2), the credential source names (P3), the 7-table SQL schema (P4), and the
  nine wire tool names (P11). Each is pinned by ported oracle tests before anything
  depends on it.
* **Independent verification** — each phase's gate runs without any later phase present.
* **No parallel worktrees** — sequential single-active execution (P-016).

### Why P1 is the right next slice

1. **It is the dependency root.** Every later phase constructs `AppError`; P2's
   validation rules are literally `AppError::Config` values.
2. **It is fully oracle-pinned.** 16 tests (`error_tests.rs` ×7,
   `path_validation_tests.rs` ×9) transfer almost mechanically.
3. **It carries the security control.** `pathsafe` is workspace isolation — brief §10
   calls it "non-negotiable". Landing it early means every later FS-touching phase builds
   on a tested primitive rather than reimplementing containment ad hoc.
4. **It contains no secrets.** Credential handling is deliberately *not* here; it gets
   its own hardened slice (P3), so a security-sensitive surface is never mixed into a
   routine slice.
5. **It is genuinely small.** ~320 LOC of Rust behavior; 5 tasks at ≤2h each.

---

## Rejected Alternatives

* **Harvest all of `037B1552` (3A)** — rejected: 12–14 tasks is an oversized shipment
  and mixes credential handling into a routine slice. The operator explicitly required a
  bounded slice with successors preserved.
* **Drop all MCP surfaces (2B)** — rejected on direct source evidence: it breaks the ACP
  HITL path (`src/main.rs:365-372`). This was the intuitively obvious reading of the brief
  and it is wrong; only the oracle falsifies it.
* **Port MCP fully (2A)** — rejected: reproduces ~4,384 LOC that the oracle's own Accepted
  ADR-0018 and epic `013-F` schedule for deletion.
* **Accept `references/herdr` as the oracle** — rejected on conclusive evidence (F1).
  Proceeding on it would have produced fabricated "oracle-grounded" contracts.
* **Defer the MCP decision again (2D)** — rejected: evidence is now decisive and the
  operator required a recorded decision.

---

## Unresolved Questions

1. **[Non-blocking, P4]** Does `modernc.org/sqlite` meet the latency bar for the repo
   query set? Must be proven in P4 before P5 is planned. Unchanged from prior deliberation.
2. **[Non-blocking, P6]** `coder/acp-go-sdk` vs a hand-ported codec. The brief requires
   preserving raw JSON-RPC `id` types for correlation; the SDK's behavior must be verified
   before adoption. Settle during P6, not now.
3. **[Non-blocking, P2]** **Oracle hot-reload defect.** Brief §7 states `channel_id`
   uniqueness is enforced "at startup **and** on the hot-reload snapshot". The oracle's
   `validate_unique_channel_ids()` doc-comment claims the same, but the reload path
   (`config_watcher.rs` → `MappingsOnlyConfig` → `parse_workspace_mappings()`) **does not
   call it**. Brief §1.3 says the Rust source wins on conflict — but this is a documented
   intent/implementation mismatch, i.e. an oracle bug, not a deliberate behavior. Proposed
   resolution: implement the *documented* behavior (validate on reload) and record the
   divergence. Decide in P2.
4. **[Non-blocking, P10]** `PatchConflict` is vestigial in the oracle (never constructed).
   Implement it properly in the Go port's diff-apply path, or preserve the oracle's
   `error_result("patch_conflict", …)` shape? Decide in P10.
5. **[BLOCKING for `A92E3FA0` only]** The multiplexer architecture fork remains
   unresolved. Out of scope here; entry stays active and unplanned.
6. **[Operator action]** `EF9352FB` — the Tavily key in the untracked `.env.local` still
   warrants rotation. Not a code change; not in this shipment.

---

## Risks and Mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| Contracts frozen against the *wrong* oracle | **Critical** | Detected and corrected this session (F1/F2). Oracle identity re-verified by remote URL, module map, LOC, and exact test-corpus counts. Pin `41df772` in every artifact and emit `docs/oracle-pin.md` in `002.001.001-ST`. |
| Oracle is out-of-tree, so the pin is documentary only | Medium | `docs/oracle-pin.md` records repo + commit + verification method; every backlog item carries the commit. A future session re-verifies before trusting it. |
| Dropping MCP-as-a-mode silently breaks ACP HITL | **Critical** | Averted by D1: the HTTP tool endpoint is explicitly retained as P11 with the nine wire names pinned. |
| Naive Go TOML decode overwrites explicit `false`/`0` with defaults | High | P2 decodes into a pre-populated `Default()` struct (stash **P1-h**); 16 non-zero defaults enumerated in F4 as test criteria. |
| Strict unknown-key rejection breaks real `config.toml` | High | F4 confirms the oracle has **no** `deny_unknown_fields`. P2 must use non-strict decoding; pinned as an explicit test. |
| Credential precedence implemented backwards (env-first) | High | F4 documents the true keychain-first, mode-scoped 4-step chain. P3 pins all 8 credential tests. |
| Secrets leak into logs or error strings | High | P3 introduces `type Secret string` with redacting `String`/`GoString`/`MarshalJSON`/`LogValue` (stash **P1-l**), so the property holds by type. Oracle confirms errors name sources, never values. Fixtures MUST be synthetic. |
| `pathsafe` TOCTOU gap inherited from the oracle | Medium | Documented in the plan as a known, accepted limitation matching oracle behavior; not silently inherited. Revisit if the threat model changes. |
| `state↔slack` import cycle is legal in Rust, illegal in Go | High | Surfaced now (F6) and narrowed during plan review: `driver/` is not in the cycle, and D1's retirement of `McpDriver` removes the only `driver→state` edge. Broken by the `Notifier` seam introduced at P6 and finalized at P7. Not discovered late. |
| `acp/reader.rs:48` imports `slack::client`, inverting P6/P8 order | High | Resolved during plan review: P6 defines a consumer-side `Notifier` interface and ships a test implementation; P8 supplies the Slack implementation. No phase reordering needed. |
| Oracle pin is documentary, so a dark-factory run could freeze contracts against a drifted oracle | Medium | `docs/oracle-pin.md` records commit + verification method, and every backlog item carries `41df772`. An automated oracle-drift check is deferred as a successor stash entry rather than expanding this slice. |
| Roadmap phases grow past 2h/task during later planning | Medium | Each phase re-enters the full Stage pipeline (deliberate → plan → harden → review → harvest) before shipping. The table's estimates are for sequencing, not commitments. |
| Stash successors lost or silently planned out of order | Medium | Both entries rewritten with explicit `SUCCESSOR ORDERING` and backlogit dependency edges; nothing archived that is not fully consumed. |

---

## Provenance

| Artifact | Reference |
|---|---|
| Authoritative port spec | `docs/decisions/2026-07-06-go-port-reference-brief.md` (stash `4989A42D`) |
| Prior architecture deliberation | `docs/decisions/2026-09-03-intercom-go-architecture-reconciliation-deliberation.md` (Group A) |
| Deferred config/error slice | stash `037B1552` |
| Behavioral oracle | `softwaresalt/agent-intercom` @ `41df772` (read-only, `C:/Source/GitHub/intercom`) |
| Rejected oracle mount | `ogulcancelik/herdr` @ `b07ba9ce` at `references/herdr` — not agent-intercom |
| Prior shipment closure | `docs/closure/2026-09-03-intercom-go-foundation-closure.md` (`001-S`, `releasability: READY`) |
| Foundation merge base | `87b800d5d3a6457d86fecd12fb7a299f170af718` (PR #5) |
| Prior learnings applied | `docs/compound/build-errors/go-race-requires-cgo-fails-closed-2026-09-03.md`, `docs/compound/best-practices/external-spec-yields-to-workspace-instructions-2026-09-03.md` |
| This decision → plan | `docs/plans/2026-09-03-intercom-go-apperr-pathsafe-plan.md` |

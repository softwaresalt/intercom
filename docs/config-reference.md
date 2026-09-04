---
title: "Configuration Reference — intercom-go"
date: 2026-09-03
status: reference
---

# Configuration Reference — intercom-go

This document describes every key in `config.toml`, its default value, and
the validation rules `internal/config` enforces at load time. It is ported
from the `agent-intercom` behavioral oracle
(`softwaresalt/agent-intercom` @ `41df772`, `src/config.rs`), with nine
recorded, deliberate divergences (V1–V9) documented below.

See `config.toml.example` for a complete, loadable synthetic example
covering every key in this reference.

## Top-level keys

| Key | Type | Default | Notes |
|---|---|---|---|
| `default_workspace_root` | string | *(required)* | Canonicalized through `pathsafe.NewRoot`; must exist. |
| `max_concurrent_sessions` | integer | `3` | Must be greater than zero. |
| `host_cli` | string | *(required)* | The host CLI binary to spawn in a later phase. A non-absolute value is checked as a non-fatal, load-time advisory (`exec.LookPath`) — resolution against `PATH` at actual spawn time happens only in that later phase, not in `internal/config`. |
| `host_cli_args` | array of strings | `[]` | Extra arguments passed to `host_cli`. |
| `commands` | table of string to string | `{}` | Named command shortcuts. |
| `http_port` | integer (0–65535) | `3000` | |
| `ipc_name` | string | `"agent-intercom"` | |
| `retention_days` | integer | `30` | |
| `slack_detail_level` | string enum | `"standard"` | One of `minimal`, `standard`, `verbose`. |

## `[slack]`

| Key | Type | Default | Notes |
|---|---|---|---|
| `slack.channel_id` | string | `""` | |
| `slack.markdown_upload_extensions` | table of string to string | `{}` | |

Secret-bearing fields (`app_token`, `bot_token`, `team_id`) are **never**
read from `config.toml` — see the migration section below.

## `[timeouts]`

| Key | Type | Default | Notes |
|---|---|---|---|
| `timeouts.approval_seconds` | integer | `3600` | |
| `timeouts.prompt_seconds` | integer | `1800` | |
| `timeouts.wait_seconds` | integer | `0` | `0` means "no timeout" (oracle parity; the only zero default). |

## `[stall]`

| Key | Type | Default | Notes |
|---|---|---|---|
| `stall.enabled` | bool | `true` | |
| `stall.inactivity_threshold_seconds` | integer | `300` | |
| `stall.escalation_threshold_seconds` | integer | `120` | |
| `stall.max_retries` | integer | `3` | |
| `stall.default_nudge_message` | string | `"Continue working on the current task. Pick up where you left off."` | |

## `[database]`

| Key | Type | Default | Notes |
|---|---|---|---|
| `database.path` | string | `"data/agent-rc.db"` | May not contain a `..` segment; absolute paths are permitted. |

## `[acp]`

| Key | Type | Default | Notes |
|---|---|---|---|
| `acp.max_sessions` | integer | `5` | |
| `acp.startup_timeout_seconds` | integer | `30` | |
| `acp.max_msg_rate` | integer | `10` | |
| `acp.http_port` | integer (0–65535) | `3001` | |

## `[[workspace]]`

Each entry maps a workspace to a Slack channel.

| Key | Type | Default | Notes |
|---|---|---|---|
| `workspace.workspace_id` | string | *(required per entry)* | Must be unique across entries. |
| `workspace.channel_id` | string | *(required per entry)* | Must be unique across entries. |
| `workspace.label` | string | `""` | Free-text label, no validation. |
| `workspace.path` | string | `""` | If non-empty, canonicalized through `pathsafe.NewRoot` and used as the workspace root for this channel instead of `default_workspace_root`. |

## The 17 non-zero defaults

`internal/config.Default()` reproduces exactly these 17 non-zero values
from the oracle (`config.rs:167-341`):

| # | Field | Value |
|---|---|---|
| 1 | `Timeouts.ApprovalSeconds` | `3600` |
| 2 | `Timeouts.PromptSeconds` | `1800` |
| 3 | `Stall.Enabled` | `true` |
| 4 | `Stall.InactivityThresholdSeconds` | `300` |
| 5 | `Stall.EscalationThresholdSeconds` | `120` |
| 6 | `Stall.MaxRetries` | `3` |
| 7 | `Stall.DefaultNudgeMessage` | `"Continue working on the current task. Pick up where you left off."` |
| 8 | `RetentionDays` | `30` |
| 9 | `MaxConcurrentSessions` | `3` |
| 10 | `ACP.MaxSessions` | `5` |
| 11 | `ACP.StartupTimeoutSeconds` | `30` |
| 12 | `ACP.MaxMsgRate` | `10` |
| 13 | `ACP.HTTPPort` | `3001` |
| 14 | `HTTPPort` | `3000` |
| 15 | `IPCName` | `"agent-intercom"` |
| 16 | `Database.Path` | `"data/agent-rc.db"` |
| 17 | `SlackDetailLevel` | `DetailStandard` |

`Timeouts.WaitSeconds` is the only field with a *zero* default (`0`,
meaning "no timeout") and is therefore excluded from this list of 17.

## Validation rules

`Validate()` performs one unconditional, fail-fast pass, returning on the
first violated rule. The order is fixed, so a given invalid config always
yields the same error.

| Order | Rule | Message |
|---|---|---|
| 1 | `max_concurrent_sessions` must be non-zero | `max_concurrent_sessions must be greater than zero` |
| 2 | `default_workspace_root` must be set and must canonicalize | `default_workspace_root must be set` (absent or empty); `default_workspace_root invalid: {err}` (canonicalization failure) |
| 3 | Per `[[workspace]]` entry, `workspace_id` must be non-empty | `workspace_id cannot be empty in [[workspace]] entry` |
| 4 | Per `[[workspace]]` entry, `channel_id` must be non-empty | `channel_id cannot be empty in [[workspace]] entry` |
| 5 | `workspace_id` must be unique across entries | `duplicate workspace_id '{id}' in [[workspace]] entries` |
| 6 | `host_cli` must be set | `host_cli must be set` |
| 7 | An absolute `host_cli` must exist | `host_cli '{path}' does not exist` |
| 8 | `channel_id` must be unique across entries | `duplicate channel_id '{id}' in [[workspace]] entries; ACP routing requires each channel_id to map to exactly one workspace` |
| 9 | Per `[[workspace]]` entry, a non-empty `path` must canonicalize | `workspace path invalid for workspace_id '{id}': {err}` |
| 10 | `database.path` must not contain a `..` segment | `database.path must not contain '..' segments` |

A relative or bare (non-absolute) `host_cli` that `exec.LookPath` cannot
resolve produces a single non-fatal advisory instead of a validation
failure.

**Order-of-checks caveat.** The rule table above describes `Validate()`'s
own internal order, which is fixed and unconditional whenever `Validate()`
is called directly. `Load`/`Decode` (the package's primary entry points)
additionally perform an earlier, separate check for
`default_workspace_root` being entirely absent from the document, before
`Validate()` is ever invoked. A document that both omits
`default_workspace_root` and violates a `Validate()` rule ordered earlier
than rule 2 (e.g. rule 1, `max_concurrent_sessions == 0`) therefore still
reports `default_workspace_root must be set` from `Load`/`Decode`, not the
rule-1 message — `Validate()`'s fixed-order guarantee holds fully only
when `Validate()` is called directly. This is a deliberate design
predating this document (Decision R8) and is locked by a regression test.

## Recorded divergences from the oracle

| # | Divergence | Direction | Notes |
|---|---|---|---|
| V1 | Missing `[slack]`, `[timeouts]`, `[stall]` sections default instead of erroring | More permissive | The oracle requires these sections due to a missing `#[serde(default)]`, not a designed contract. |
| V2 | Case-fold-colliding keys (e.g. `host_cli` and `Host_CLI`) are **rejected** | Stricter | Prevents nondeterministic field selection from the decoder's `EqualFold` fallback. |
| V2b | A lone non-canonical spelling with no colliding variant still decodes | More permissive | Harmless without a collision. |
| V3 | Unknown keys are reported to the caller | Additive | The oracle ignores them silently. |
| V4 | `channel_id` uniqueness is enforced eagerly at load | Stricter | The oracle enforces this at ACP session start, which does not exist in this port. |
| V5 | The `host_cli` empty-value message drops "to use ACP mode" | Cosmetic | ACP/transport mode selection is retired in this port. |
| V6 | Secret-bearing fields (`app_token`, `bot_token`, `team_id`, `authorized_user_ids`) are omitted from the schema entirely | Narrower | Never read from `config.toml`; owned by a later credential-resolution phase. |
| V7 | A non-empty `[[workspace]].path` must canonicalize | Stricter | The oracle passes it unvalidated into a subprocess working directory. |
| V8 | `database.path` may not contain a `..` segment | Stricter | The oracle leaves it unrestricted. Absolute paths remain permitted as an explicit operator privilege. |
| V9 | The config file is bounded at 1 MiB | Stricter | The oracle reads the file unbounded. |

Divergences V2, V4, V7, V8, and V9 are the only ones that can reject a
configuration the oracle would accept.

## Operator migration from agent-intercom (Rust)

If you are migrating a `config.toml` from the Rust `agent-intercom`
implementation to `intercom-go`, note the following behavioral changes:

* **Duplicate `channel_id` is now a startup failure (V4).** The Rust
  implementation only rejected a duplicate `channel_id` when an ACP session
  started; `intercom-go` rejects it at load time. If your existing config
  has two `[[workspace]]` entries sharing a `channel_id`, loading it will
  now fail immediately rather than at first use.
* **`--mode` and `protocol_mode` are gone.** Multiplexer/transport mode
  selection is retired in this port; any `mode` or `protocol_mode` key in
  an existing `config.toml` is not part of the schema and will be reported
  as an unknown key (V3) rather than applied.
* **Tokens in `config.toml` are ignored.** `app_token`, `bot_token`,
  `team_id`, and any `authorized_user_ids` entry are not read from the
  config file at all (V6) — they will be reported as unknown keys if
  present. Credential values are resolved outside `config.toml` by a
  separate credential-resolution phase; do not rely on tokens present in a
  migrated config file being honored.

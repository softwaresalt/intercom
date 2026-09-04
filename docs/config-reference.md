---
title: "Configuration Reference — intercom-go"
date: 2026-09-04
status: reference
---

# Configuration Reference — intercom-go

This document describes every key in `config.toml`, its default value, and
the validation rules `internal/config` enforces at load time for the
corrected C1 architecture.

See `config.toml.example` for a complete, loadable synthetic example
covering every key in this reference.

## Top-level keys

| Key | Type | Default | Notes |
|---|---|---|---|
| `default_workspace_root` | string | *(required)* | Canonicalized through `pathsafe.NewRoot`; must exist. |
| `max_concurrent_sessions` | integer | `3` | Must be greater than zero. |
| `http_port` | integer (0–65535) | `3000` | Backend HTTP/WebSocket listener port. |
| `retention_days` | integer | `30` | Retained as a deferred persistence-facing setting. |
| `operator_detail_level` | string enum | `"standard"` | One of `minimal`, `standard`, `verbose`. |

## `[copilot]`

| Key | Type | Default | Notes |
|---|---|---|---|
| `copilot.cli_path` | string | `""` | Empty means resolve the Copilot CLI from `PATH` / `COPILOT_CLI_PATH`. A bare name is allowed with a non-fatal advisory if it is not currently resolvable on `PATH`. An absolute path must exist. A relative path is rejected. |

## `[commands]`

| Key | Type | Default | Notes |
|---|---|---|---|
| `commands.<name>` | string | `{}` | Named command shortcuts. Values are operator-authored prompt content. |

## `[timeouts]`

| Key | Type | Default | Notes |
|---|---|---|---|
| `timeouts.approval_seconds` | integer | `3600` | |
| `timeouts.prompt_seconds` | integer | `1800` | |
| `timeouts.wait_seconds` | integer | `0` | `0` means "no timeout". |

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
| `database.path` | string | `"data/agent-rc.db"` | May not contain a `..` segment. Absolute paths are permitted. |

## `[[workspace]]`

| Key | Type | Default | Notes |
|---|---|---|---|
| `workspace.workspace_id` | string | *(required per entry)* | Must be unique across entries. |
| `workspace.label` | string | `""` | Free-text label, no validation. |
| `workspace.path` | string | `""` | If non-empty, canonicalized through `pathsafe.NewRoot` and used as the workspace root for that entry instead of `default_workspace_root`. |

## The non-zero defaults

`internal/config.Default()` pre-populates these non-zero values:

| Field | Value |
|---|---|
| `Timeouts.ApprovalSeconds` | `3600` |
| `Timeouts.PromptSeconds` | `1800` |
| `Stall.Enabled` | `true` |
| `Stall.InactivityThresholdSeconds` | `300` |
| `Stall.EscalationThresholdSeconds` | `120` |
| `Stall.MaxRetries` | `3` |
| `Stall.DefaultNudgeMessage` | `"Continue working on the current task. Pick up where you left off."` |
| `RetentionDays` | `30` |
| `MaxConcurrentSessions` | `3` |
| `HTTPPort` | `3000` |
| `Database.Path` | `"data/agent-rc.db"` |
| `OperatorDetailLevel` | `DetailStandard` |

`Timeouts.WaitSeconds` is the only field with a zero default (`0`, meaning
"no timeout") and is therefore excluded from this table.

## Validation rules

`Validate()` performs one unconditional, fail-fast pass, returning on the
first violated rule. The order is fixed, so a given invalid config always
yields the same error.

| Order | Rule | Message |
|---|---|---|
| 1 | `max_concurrent_sessions` must be non-zero | `max_concurrent_sessions must be greater than zero` |
| 2 | `default_workspace_root` must be set and must canonicalize | `default_workspace_root must be set` (absent or empty); `default_workspace_root invalid: {err}` (canonicalization failure) |
| 3 | Per `[[workspace]]` entry, `workspace_id` must be non-empty | `workspace_id cannot be empty in [[workspace]] entry` |
| 4 | `workspace_id` must be unique across entries | `duplicate workspace_id '{id}' in [[workspace]] entries` |
| 5 | `copilot.cli_path` must be well-formed | `copilot.cli_path '{path}' does not exist`; `copilot.cli_path '{path}' must not be drive-relative; use an absolute path, a bare name, or empty`; `copilot.cli_path '{path}' must not be relative; use an absolute path, a bare name, or empty` |
| 6 | Per `[[workspace]]` entry, a non-empty `path` must canonicalize | `workspace path invalid for workspace_id '{id}': {err}` |
| 7 | `database.path` must not contain a `..` segment | `database.path must not contain '..' segments` |

Rule 5 has one additional non-fatal advisory: a bare name that is not
currently resolvable on `PATH` produces `copilot.cli_path '{name}' was not
resolvable via PATH` in `Report.Warnings`, but loading still succeeds.

The rule-5 classifier is:

```text
cli_path == ""                              -> valid, return early with no warning
filepath.IsAbs(cli_path)                    -> valid iff the file exists
matches ^[A-Za-z]: and not absolute         -> error (Windows drive-relative)
contains '/' or '\' and not absolute       -> error (relative path)
otherwise                                   -> bare name, PATH advisory only
```

**Order-of-checks caveat.** The table above describes `Validate()`'s own
internal order, which is fixed whenever `Validate()` is called directly.
`Load`/`Decode` additionally perform an earlier, separate check for
`default_workspace_root` being entirely absent from the document, before
`Validate()` is ever invoked. A document that both omits
`default_workspace_root` and violates an earlier `Validate()` rule (for
example rule 1) therefore still reports `default_workspace_root must be set`
from `Load`/`Decode`, not the rule-1 message.

## Migration from pre-correction configs

The C1 correction removes, renames, or re-founds several previously shipped
keys. Tolerant decode keeps most legacy files loadable by reporting retired
keys in `Report.UnknownKeys` rather than failing immediately. One documented
exception is listed below.

### Removed keys and sections

| Previously written | What to write instead |
|---|---|
| `[slack]` | Delete the section. There is no replacement in C1. |
| `[slack].channel_id` | Delete the key. There is no replacement. |
| `[slack].markdown_upload_extensions` | Delete the table. There is no replacement. |
| `[acp]` | Delete the section. There is no replacement in C1. |
| `[acp].max_sessions` | Delete the key. Do not migrate it. |
| `[acp].startup_timeout_seconds` | Delete the key in C1. Do not add a replacement yet; the follow-on `[copilot]` timeout work is deferred to C2. |
| `[acp].max_msg_rate` | Delete the key. There is no replacement. |
| `[acp].http_port` | Delete the key. There is no replacement. |
| `host_cli_args` | Delete the key. Do not pass `--dangerously-skip-permissions`; that bypasses operator approval prompts and is intentionally removed. |
| `ipc_name` | Delete the key. There is no replacement in C1. |
| `[[workspace]].channel_id` | Delete the key from every workspace entry. Workspace lookup is now keyed by `workspace_id`. |

### Renamed key

| Previously written | What to write instead |
|---|---|
| `slack_detail_level` | `operator_detail_level` |

### Migrated key

| Previously written | What to write instead |
|---|---|
| `host_cli` | Write `[copilot]` with `cli_path = ""`, a bare executable name, or an absolute path to the CLI binary. Empty means SDK default resolution from `PATH` / `COPILOT_CLI_PATH`. Bare names are allowed with a warning if unresolved. Absolute paths must exist. Relative paths and Windows drive-relative paths are rejected. |

### Validation behavior changes

* The old per-workspace presence rule for `channel_id` is gone because the
  key itself is gone.
* The old uniqueness rule over `channel_id` is gone for the same reason.
* The old required-value rule for `host_cli` is gone. An empty
  `[copilot].cli_path` is now valid and means SDK default resolution.
* The old absolute-path existence check now applies to
  `[copilot].cli_path`, with the added rejection of relative and
  drive-relative paths.

### Documented tolerant-decode exception

Most retired keys now surface in `Report.UnknownKeys` and remain non-fatal.
One exception is a legacy `[slack.markdown_upload_extensions]` table whose
keys differ only by case, such as `".md"` and `".MD"`. That shape now fails
with a case-fold collision error rather than loading with warnings, because
the legacy nested map field no longer exists in the schema.

### Security note

The example and migration target deliberately remove
`host_cli_args = ["--dangerously-skip-permissions"]`. The corrected
architecture treats permission approval as a real operator control; a
migration should not preserve a config line that disables it.

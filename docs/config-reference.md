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
| `default_workspace_root` | string | *(required)* | Canonicalized through `pathsafe.NewRoot`; must exist. The process workspace when `--workspace` is not given; `--workspace` overrides it (it may name a subdirectory of this root or of a `[[workspace]].path`; see [runbook section 2](runbooks/intercom-operator-runbook.md#2-start)). |
| `max_concurrent_sessions` | integer | `3` | Must be greater than zero. Accepted, not yet consumed by the MVP process (it runs exactly one session). |
| `http_port` | integer (0–65535) | `3000` | Backend HTTP/WebSocket listener port. `--port` overrides it, including `--port 0` for a free port (see [runbook section 2](runbooks/intercom-operator-runbook.md#2-start)). |
| `retention_days` | integer | `30` | Retained as a deferred persistence-facing setting. Accepted, not yet consumed by the MVP process (no persistence). |
| `operator_detail_level` | string enum | `"standard"` | One of `minimal`, `standard`, `verbose`. Accepted, not yet consumed by the MVP process. |

## `[copilot]`

| Key | Type | Default | Notes |
|---|---|---|---|
| `copilot.cli_path` | string | `""` | Empty means resolve the Copilot CLI from `PATH` / `COPILOT_CLI_PATH`. A bare name is allowed with a non-fatal advisory if it is not currently resolvable on `PATH`. An absolute path must exist. A relative path is rejected. |
| `copilot.startup_timeout_seconds` | integer | `30` | Deadline for the agent adapter's `Client.Start` (also covers the version/protocol assertion and session creation). Must be between `1` and `3600`. |

## `[commands]`

| Key | Type | Default | Notes |
|---|---|---|---|
| `commands.<name>` | string | `{}` | Named command shortcuts. Values are operator-authored prompt content. Accepted, not yet consumed by the MVP process. |

## `[timeouts]`

| Key | Type | Default | Notes |
|---|---|---|---|
| `timeouts.approval_seconds` | integer | `3600` | How long a permission request waits for an operator answer before the hub resolves it `UserNotAvailable`. Consumed by the process wiring (`internal/app`); `0` selects the hub's built-in default (2 minutes). Must be between `0` and `86400` (24 hours, `MaxApprovalTimeoutSeconds`) inclusive; a larger value fails config load with `timeouts.approval_seconds must not exceed 86400`. |
| `timeouts.prompt_seconds` | integer | `1800` | Accepted, not yet consumed by the MVP process. |
| `timeouts.wait_seconds` | integer | `0` | `0` means "no timeout". Accepted, not yet consumed by the MVP process. |

## `[stall]`

Every `stall.*` key is accepted, not yet consumed by the MVP process (no stall
detection runs).

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
| `database.path` | string | `"data/agent-rc.db"` | May not contain a `..` segment. Absolute paths are permitted. Validated only: accepted, not yet consumed by the MVP process (nothing is persisted). |

## `[remote]`

The MVP remote-surface group (ADR MVP-2 / MVP-3). Every default is secure:
no origins or hosts are allowed and remote approval of destructive kinds is
off. This section is secret-free by design.

| Key | Type | Default | Notes |
|---|---|---|---|
| `remote.token_lifetime_seconds` | integer | `14400` | Must be between `60` and `86400` inclusive. |
| `remote.allowed_origins` | array of string | `[]` | Exact lowercase `http(s)://host[:port]` origins in canonical form; bracketed IPv6 literals (`http://[::1]:3000`) are allowed. An explicit default port (`https://x:443`, `http://x:80`) is rejected: the default port must be omitted. Wildcards, paths (including a trailing `/`), queries (including a bare `?`), fragments, userinfo, `%` escapes, leading-zero or out-of-range ports, uppercase, other schemes, trailing dots, empty and duplicate entries are rejected. |
| `remote.allowed_hosts` | array of string | `[]` | A lowercase host with an optional `:port` (decimal 1-65535, no leading zeros), e.g. `app.example.com`, `localhost:3000`, `[::1]`, `[::1]:3000`. An entry without a port matches the default port only (the request carries no port); an entry with an explicit port matches only that port. **Default-port policy:** explicit `:80` and `:443` are accepted as ordinary explicit ports and are NOT treated as equivalent to the bare host by config validation; `example.com` and `example.com:443` are distinct entries (both may be listed), and the matcher (`internal/transport` `hostorigin.go`) does not equate them. Numeric-looking hosts must be canonical dotted-quad IPv4 (`127.0.0.1`); forms such as `127.1`, `2130706433`, `01.2.3.4` or `999.999.999.999` are rejected as a non-canonical IPv4 address (as are bracketed non-IPv6 origins such as `http://[127.0.0.1]`). Host names are at most 253 characters with no empty label, no trailing dot, labels of at most 63 characters that do not start or end with a hyphen. Wildcards, schemes, paths, userinfo, whitespace/control or non-ASCII characters, empty and duplicate entries are rejected. |
| `remote.max_connections` | integer | `8` | Must be between `1` and `1024` (`MaxRemoteConnections`) inclusive. |
| `remote.remote_approve_destructive` | bool | `false` | When `true`, remote surfaces may approve `shell`, `write`, and unknown kinds. |

**Credential-shaped keys are rejected.** `Decode`/`Load` fail closed on any
key at or below `[remote]` whose name looks like a credential (it contains
`token`, `secret`, `pass`, `pwd`, `credential`, `bearer`, `key`, `auth`,
`jwt`, `cookie`, `signature`, `session`, `pat`, `psk`, `otp`, `pin`, `private`,
`cert`, `hmac`, `salt`, `dsn`, `connstr`, or `connection`, ignoring case and
punctuation), other than the five known keys above. The screen is
intentionally broader than ADR MVP-2's list; a new `[remote]` key must be
added to `remoteKnownKeys` in `internal/config/remote.go`.

An absent `[remote]` section = secure defaults; no migration is needed for
existing config files. The error names the key path and never
echoes the value. Credentials never come from `config.toml`.

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
| `Copilot.StartupTimeoutSeconds` | `30` |
| `Database.Path` | `"data/agent-rc.db"` |
| `OperatorDetailLevel` | `DetailStandard` |
| `Remote.TokenLifetimeSeconds` | `14400` |
| `Remote.MaxConnections` | `8` |
| `Remote.AllowedOrigins` | `[]` (empty) |
| `Remote.AllowedHosts` | `[]` (empty) |
| `Remote.RemoteApproveDestructive` | `false` |

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
| 5 | `[copilot]` must be well-formed | `copilot.cli_path '{path}' does not exist`; `copilot.cli_path '{path}' must not be drive-relative; use an absolute path, a bare name, or empty`; `copilot.cli_path '{path}' must not be relative; use an absolute path, a bare name, or empty`; `copilot.startup_timeout_seconds must be greater than zero`; `copilot.startup_timeout_seconds must not exceed 3600` |
| 6 | Per `[[workspace]]` entry, a non-empty `path` must canonicalize | `workspace path invalid for workspace_id '{id}': {err}` |
| 7 | `database.path` must not contain a `..` segment | `database.path must not contain '..' segments` |

### `[remote]` rule group

A separate group, run inside `Validate()` after the seven rules above (their
order is unchanged). Within the group the order is: token lifetime, origins,
hosts, max connections. Messages never echo entry values.

| Rule | Message |
|---|---|
| `token_lifetime_seconds` within 60..86400 | `remote.token_lifetime_seconds must be between 60 and 86400` |
| each `allowed_origins` entry is an exact origin, unique | `remote.allowed_origins[{i}] {reason}` |
| each `allowed_hosts` entry is a canonical host[:port], unique | `remote.allowed_hosts[{i}] {reason}` |
| `max_connections` at least 1 | `remote.max_connections must be at least 1`; `remote.max_connections must not exceed 1024` |
| no credential-shaped key under `[remote]` (checked by `Decode`, before `Validate`, because it needs the raw key set) | `remote.{key} is a credential-shaped key; credentials must never be stored in config.toml (remove it from [remote])` |

### `[timeouts]` rule

A final rule runs inside `Validate()` after the `[remote]` group (the earlier
orders are unchanged), so a failure never mutates the receiver.

| Rule | Message |
|---|---|
| `approval_seconds` at most 86400 (`0` stays valid and selects the hub's 2-minute default) | `timeouts.approval_seconds must not exceed 86400` |

A config that sets `approval_seconds` above 86400 and loaded before this rule
existed now fails validation at load time; lower the value to 86400 or less.

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
from `Load`/`Decode`, not the rule-1 message. Likewise the `[remote]`
credential-key screen runs in `Decode` after the case-fold check and before the
required-key check, so a credential-shaped `[remote]` key is reported ahead of
a missing `default_workspace_root` and ahead of every `Validate()` rule.

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
| `[acp].startup_timeout_seconds` | Deleted in C1. Re-introduced in C3 as `[copilot].startup_timeout_seconds` with its first consumer, the agent adapter's `Client.Start` deadline. |
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

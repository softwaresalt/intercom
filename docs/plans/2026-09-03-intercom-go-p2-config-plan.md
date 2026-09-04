---
title: "Implementation Plan — intercom-go P2: Configuration Package"
description: "Port the agent-intercom config.toml schema, 17 defaults, tolerant TOML decode, and unified validation into internal/config"
date: 2026-09-03
status: reviewed
phase: P2
source_deliberation: "docs/decisions/2026-09-03-intercom-go-p2-config-deliberation-addendum.md"
parent_deliberation: "docs/decisions/2026-09-03-intercom-go-port-continuation-deliberation.md"
source_brief: "docs/decisions/2026-07-06-go-port-reference-brief.md"
behavioral_oracle:
  repo: "softwaresalt/agent-intercom"
  commit: "41df772"
  access: "read-only, out-of-tree"
stash_entries: ["037B1552"]
predecessor: "002-F / 002-S (P1) — merge 9efb156"
successor: "P3 credential resolution (stash 037B1552)"
merge_base: "5635e77"
prior_closure: "docs/closure/002-S-002-F-post-merge-closure.md"
requires_plan_hardening: yes
plan_review_verdict: PASS
tags: ["go", "port", "config", "toml", "validation"]
---

# Implementation Plan — intercom-go P2: Configuration Package

## Problem Frame

Shipment `002-S` delivered `internal/apperr` (the 14-variant error taxonomy) and
`internal/pathsafe` (workspace path containment). Both are currently **zero-importer
primitives** — nothing in the tree consumes them yet. P2 is the first phase that does.

P2 delivers `internal/config`: the `config.toml` schema, the 17 non-zero defaults, tolerant TOML
decoding that preserves both defaults and explicit `false`/`0`, one unified validation pass, and
workspace→channel routing. It is the last phase before persistence, and every later phase reads
its output.

All contracts are fixed by the deliberation addendum, which re-read the oracle specifically for
P2 and corrected three quantitative claims carried in the stash entry.

## Scope Boundaries

**In scope**

- `internal/config`: schema types, `Default()`, `Load`/`Decode`, `Validate`, workspace resolvers.
- One new direct dependency: `github.com/BurntSushi/toml` v1.6.0.
- `config.toml.example`, `docs/config-reference.md`, a README configuration section, and
  alignment of the existing `--config` flag default.
- Unit and contract tests, synthetic fixtures only.

**Explicitly out of scope**

| Excluded | Owner |
|---|---|
| Credential resolution, keychain, env precedence, `Secret` type | **P3** — next successor under stash `037B1552` |
| Secret-bearing fields (`app_token`, `bot_token`, `team_id`, `authorized_user_ids`) | **P3** — see Decision R6 |
| Config hot-reload / file watcher | P-later |
| `protocol_mode` (a DB column, not a config key) | P4/P5 persistence |
| Reading config in `cmd/intercom`'s `RunE` / any server behavior | P4+ |
| Oracle-drift CI gate | stash `8ACF7110` — separate artifact class |
| Multiplexer architecture | stash `A92E3FA0` — blocked |
| Fixing `pathsafe`/`apperr` cause-dropping | stash `4A91CA81` — see Constraint K1 |

**Scope guard — authorized paths.** `internal/config/**`, `go.mod`, `go.sum`,
`config.toml.example`, `docs/**`, `README.md`, and `cmd/intercom/main.go` (one-line flag default
only). Nothing else. Explicitly **not** authorized: `internal/apperr/**`, `internal/pathsafe/**`,
`.golangci.yml`, `.github/workflows/**`, `scripts/**`.

## Constraint K1 — the `apperr` construction limit (governs every error in this slice)

`internal/apperr` exposes exactly three constructors, and its `Error` fields are unexported:

- `New(kind, msg)` → custom message, **`cause == nil`**.
- `Newf(kind, format, args...)` → custom message, **`cause == nil`**.
- `Wrap(kind, cause)` → cause retained, but **`msg` is forced to `cause.Error()`**.

There is therefore **no way to construct an error that carries both a custom message and a
retained cause**. Additionally, `pathsafe.NewRoot` builds its error with `apperr.New` and
stringifies the underlying failure, so the original `os.ErrNotExist` is already gone before P2
sees it (the defect recorded as stash `4A91CA81`).

**Consequence, and why it is acceptable:** the oracle has the same property. Its
`AppError::Config(String)` variant carries a formatted string and no source —
`AppError::Config(format!("default_workspace_root invalid: {err}"))` at `config.rs:545` and
`AppError::Config(format!("cannot read config file '{path}': {err} — …"))` at `main.rs:110` both
stringify. Preserving a machine-inspectable cause would be **beyond-oracle behavior**, not
parity.

**Binding rule for this slice:** every validation and load error is built with
`apperr.New`/`apperr.Newf` carrying the oracle's message text. No unit may assert
`errors.Is(err, os.ErrNotExist)` or `errors.As` cause retrieval, and **no unit may modify
`internal/apperr` to make that possible**. Callers discriminate on `Kind` and on message text.

Adding a cause-preserving constructor (`apperr.Wrapf`) is the correct long-term fix; it is the
same root cause as stash `4A91CA81`, which already carries `Requires deliberation: yes`. P2
records the broadened evidence and does **not** act on it.

## Requirements Trace

| ID | Requirement | Oracle source | Unit |
|---|---|---|---|
| R1 | Schema matches the oracle's TOML key names exactly | `config.rs:344-419, 52-72, 104-127, 158-191, 243-278, 294-316` | A1 |
| R2 | All 17 non-zero defaults reproduced | `config.rs:167-341` | A2 |
| R3 | Decode is tolerant of unknown keys | no `deny_unknown_fields` in the tree | B1 |
| R4 | Unknown keys surfaced to the caller, bounded and sanitized | Go addition (P2-D3) | B1 |
| R5 | Defaults survive absent keys; explicit `false`/`0` overrides them | `config.rs:413-425`, P2-D2 | B3 |
| R6 | `default_workspace_root` required; `host_cli` required | `config.rs:346, 379` | B2, C3 |
| R7 | Missing / malformed / oversized config produces a `KindConfig` error | `config.rs:413-420`, `main.rs:105-113` | B2 |
| R8 | 10 hard validation rules, fixed order, fail-fast | `config.rs:539-550, 560-582, 661-677` | C1–C4 |
| R9 | `default_workspace_root` canonicalized + UNC-stripped, must exist | `config.rs:543-547` | C1 |
| R10 | `host_cli` PATH advisory never fails the load | `main.rs:139` (`.ok()`) | C3 |
| R11 | Workspace/channel resolvers with default-root fallback | `config.rs:600-640` | D1 |
| R12 | `channel_id` uniqueness enforced eagerly at load | `config.rs:86-100`, P2-D6 | C2 |
| R13 | Operator-facing example, reference doc, and migration notes | Go addition | E1, E2 |
| R14 | Config path discovery constant aligned with the oracle default | `main.rs:60` | E3 |

## Constitution Check

Verified against `.github/instructions/constitution.instructions.md` as ratified in this
workspace. (The P1 plan's table used stale principle titles — recorded as stash `2130906D`; this
table is regenerated against the current numbering.)

| Principle | Compliance |
|---|---|
| I. Safety-First Go | No `unsafe`, no CGO, no panics on operator input; every failure returns a typed `*apperr.Error`. Unsigned integer types make negative counts a decode error. Config reads are size-bounded. |
| II. Test-First Development (NON-NEGOTIABLE) | Every unit names the test that must exist and fail before implementation. |
| III. Workspace Isolation and Security Boundaries | `default_workspace_root` **and** every `[[workspace]].path` are canonicalized through `pathsafe.NewRoot`; `database.path` traversal is rejected. Config cannot widen the containment boundary. |
| IV. CLI Workspace Containment (NON-NEGOTIABLE) | P2 adds no CLI behavior. E3 changes one flag *default* string; `RunE` still returns not-implemented. |
| V. Structured Observability | Unknown keys and `host_cli` advisories are returned as bounded, sanitized data for the caller to log; library code never logs. |
| VI. Single Responsibility | Four files with disjoint roles: schema/defaults, decode, validate, resolve. |
| VII. Destructive Command Approval (NON-NEGOTIABLE) | No destructive operations. |
| VIII. Explicit Safety Modes for Elevated Risk | Not triggered. |
| IX. Git-Friendly Persistence | No persistence in this slice. |
| X. Agent Context Efficiency | 13 units, each ≤2h with self-contained, machine-verifiable criteria. |
| XI. Merge Commit History Preservation (NON-NEGOTIABLE) | Ship merges with `--no-ff`; unchanged by this plan. |

## Type Contract

Package `internal/config`. Exact declarations the implementation must produce:

```go
// DefaultConfigPath matches the oracle's --config default (main.rs:60).
const DefaultConfigPath = "config.toml"

// MaxConfigBytes bounds a config file read (security hardening, no oracle counterpart).
const MaxConfigBytes = 1 << 20 // 1 MiB

type Config struct {
    DefaultWorkspaceRoot  string             `toml:"default_workspace_root"`
    Slack                 SlackConfig        `toml:"slack"`
    MaxConcurrentSessions uint32             `toml:"max_concurrent_sessions"`
    HostCLI               string             `toml:"host_cli"`
    HostCLIArgs           []string           `toml:"host_cli_args"`
    Commands              map[string]string  `toml:"commands"`
    HTTPPort              uint16             `toml:"http_port"`
    IPCName               string             `toml:"ipc_name"`
    Timeouts              TimeoutConfig      `toml:"timeouts"`
    Stall                 StallConfig        `toml:"stall"`
    RetentionDays         uint32             `toml:"retention_days"`
    Database              DatabaseConfig     `toml:"database"`
    SlackDetailLevel      SlackDetailLevel   `toml:"slack_detail_level"`
    ACP                   ACPConfig          `toml:"acp"`
    Workspaces            []WorkspaceMapping `toml:"workspace"`
}

type SlackConfig struct {
    ChannelID                string            `toml:"channel_id"`
    MarkdownUploadExtensions map[string]string `toml:"markdown_upload_extensions"`
}

type TimeoutConfig struct {
    ApprovalSeconds uint64 `toml:"approval_seconds"`
    PromptSeconds   uint64 `toml:"prompt_seconds"`
    WaitSeconds     uint64 `toml:"wait_seconds"`
}

type StallConfig struct {
    Enabled                    bool   `toml:"enabled"`
    InactivityThresholdSeconds uint64 `toml:"inactivity_threshold_seconds"`
    EscalationThresholdSeconds uint64 `toml:"escalation_threshold_seconds"`
    MaxRetries                 uint32 `toml:"max_retries"`
    DefaultNudgeMessage        string `toml:"default_nudge_message"`
}

type ACPConfig struct {
    MaxSessions           uint32 `toml:"max_sessions"`
    StartupTimeoutSeconds uint64 `toml:"startup_timeout_seconds"`
    MaxMsgRate            uint32 `toml:"max_msg_rate"`
    HTTPPort              uint16 `toml:"http_port"`
}

type DatabaseConfig struct {
    Path string `toml:"path"`
}

type WorkspaceMapping struct {
    WorkspaceID string `toml:"workspace_id"`
    ChannelID   string `toml:"channel_id"`
    Label       string `toml:"label"`
    Path        string `toml:"path"`
}

type SlackDetailLevel uint8

const (
    DetailMinimal SlackDetailLevel = iota
    DetailStandard
    DetailVerbose
)

func (d *SlackDetailLevel) UnmarshalText(text []byte) error

// Report carries bounded, non-fatal load observations for the caller to log.
type Report struct {
    UnknownKeys []string // sorted, sanitized, capped dotted TOML paths
    Warnings    []string // host_cli PATH advisory; never fatal
}

func Default() *Config
func Decode(data string) (*Config, Report, error)
func Load(path string) (*Config, Report, error)
func (c *Config) Validate() (Report, error)
func (c *Config) ResolveChannelID(workspaceID string) (string, bool)
func (c *Config) ResolveWorkspaceByChannelID(channelID string) (WorkspaceMapping, bool)
func (c *Config) WorkspaceRootForChannel(channelID string) string
```

**Type mapping rationale.** Rust `u16`/`u32`/`u64` map to Go `uint16`/`uint32`/`uint64`;
`ACPConfig.max_sessions` is Rust `usize` and maps to `uint32`. Unsigned types make a negative or
over-range TOML integer a *decode* error (`unifyInt` rejects `num < 0` and `OverflowUint`) rather
than a validation concern. `PathBuf` maps to `string`. `Option<String>`/`Option<PathBuf>` on
`WorkspaceMapping` map to plain `string`, where empty means absent — the oracle only ever tests
these for emptiness or falls back, never distinguishing `None` from `Some("")`.

`ResolveWorkspaceByChannelID` returns a **value, not a pointer**, so callers cannot mutate a
validated mapping in place or retain a stale alias into a slice that a future hot-reload
replaces.

**Secret-field omission (Decision R6).** The oracle's `SlackConfig` also declares `app_token`,
`bot_token`, `team_id`, and `GlobalConfig` declares `authorized_user_ids` — all four are
`#[serde(skip)]`, i.e. **never read from `config.toml`**. P2 omits them entirely rather than
declaring unpopulated fields, making "no secrets in config fixtures" structurally true rather
than a review convention.

**P3 boundary (Decision R7).** `internal/config` remains a **secret-free file-schema package
permanently**, not just in P2. P3 must return a separate `ResolvedCredentials` value composed
with `*config.Config` in a higher-level bootstrap package. P3 must **not** add secret fields to
`config.Config`, which would invert the dependency (config → credential layer) while credential
resolution already consumes config.

## Default Contract

`Default()` returns a `*Config` with exactly these 17 non-zero values.

| # | Field | Value | Oracle |
|---|---|---|---|
| 1 | `Timeouts.ApprovalSeconds` | `3600` | `config.rs:167` |
| 2 | `Timeouts.PromptSeconds` | `1800` | `config.rs:169` |
| 3 | `Stall.Enabled` | `true` | `default_true` |
| 4 | `Stall.InactivityThresholdSeconds` | `300` | `default_inactivity_threshold` |
| 5 | `Stall.EscalationThresholdSeconds` | `120` | `default_escalation_threshold` |
| 6 | `Stall.MaxRetries` | `3` | `default_max_retries` |
| 7 | `Stall.DefaultNudgeMessage` | `"Continue working on the current task. Pick up where you left off."` | `default_nudge_message` |
| 8 | `RetentionDays` | `30` | `default_retention_days` |
| 9 | `MaxConcurrentSessions` | `3` | `default_max_concurrent_sessions` |
| 10 | `ACP.MaxSessions` | `5` | `default_acp_max_sessions` |
| 11 | `ACP.StartupTimeoutSeconds` | `30` | `default_acp_startup_timeout_seconds` |
| 12 | `ACP.MaxMsgRate` | `10` | `default_acp_max_msg_rate` |
| 13 | `ACP.HTTPPort` | `3001` | `default_acp_http_port` |
| 14 | `HTTPPort` | `3000` | `default_http_port` |
| 15 | `IPCName` | `"agent-intercom"` | `default_ipc_name` |
| 16 | `Database.Path` | `"data/agent-rc.db"` | `default_db_path` |
| 17 | `SlackDetailLevel` | `DetailStandard` | `default_slack_detail_level` |

`Timeouts.WaitSeconds` defaults to `0` (oracle parity — `0` means "no timeout").
`Slack.ChannelID` defaults to `""`. `Commands`, `HostCLIArgs`,
`Slack.MarkdownUploadExtensions`, and `Workspaces` default to empty; resolvers must be nil-safe.

**`DetailStandard` must not be the iota zero value.** `DetailMinimal` is `0`, so a `Config` built
by struct literal rather than `Default()` would silently be `minimal`. A2 asserts this directly.

## Decode Contract

`Decode(data)` performs, in order:

1. `cfg := Default()` — pre-populated baseline.
2. `md, err := toml.Decode(data, cfg)` — syntax errors and type mismatches become
   `apperr.KindConfig` via `apperr.Wrap` (message is the decoder's, which is the desired text).
3. **Case-fold collision rejection.** Scan `md.Keys()`; if two distinct key paths are equal under
   `strings.EqualFold`, return `apperr.KindConfig`. See divergence V2.
4. **Required-key check** — `md.IsDefined("default_workspace_root")` only. `host_cli` is *not*
   checked here; validation rule 6 covers both the absent and the explicitly-empty case with one
   message (see Decision R8).
5. `md.Undecoded()` → `Report.UnknownKeys`: sorted, control characters escaped, each path
   truncated to 128 bytes, list capped at 64 entries with a final `"… N more"` marker.
6. `cfg.Validate()` → its `Report` merged into the returned `Report`.

`Load(path)` stats the file, rejects anything larger than `MaxConfigBytes`, reads it, and
delegates to `Decode`.

**Report-on-error semantics.** `Decode` and `Load` return the `Report` populated with everything
gathered *so far* even when they return a non-nil error, so a caller can log unknown keys
alongside the failure. A returned `Config` is only valid when `err == nil`.

**Why decode-over-`Default()` is correct.** `BurntSushi/toml`'s `unifyStruct` iterates
`for key, datum := range tmap` — only keys present in the document — and `decode.go` contains no
`reflect.Zero` or `SetZero`. Absent keys leave the pre-populated value untouched, while an
explicitly written `enabled = false` *is* present in `tmap` and *does* overwrite the `true`
default. Both halves of R5 hold structurally. Verified against library source, not assumed.

**Relative-path base (Decision R9).** `Database.Path` and `WorkspaceMapping.Path` may be
relative. Their base is the **process working directory**, matching the oracle, which performs no
resolution at all. `Load` deliberately does *not* retain the config file's directory, and later
phases **must not** invent config-file-relative semantics. This is stated here so P4 (which
creates the database parent directory) and P6 (which uses `workspace.path` as a subprocess
working directory) cannot choose independently.

### Recorded divergences from the oracle

| # | Divergence | Direction | Justification |
|---|---|---|---|
| V1 | Missing `[slack]`, `[timeouts]`, `[stall]` sections default instead of erroring | More permissive | Oracle's required-ness is an artifact of a missing `#[serde(default)]`, not a designed contract (P2-D4). |
| V2 | Case-fold-colliding keys are **rejected** | Stricter | `BurntSushi/toml` falls back to `strings.EqualFold`, so `host_cli` and `Host_CLI` both target one field and Go map iteration picks a winner **nondeterministically**. Rather than accept that, P2 rejects the collision outright. Non-colliding case variants still decode (V2b). |
| V2b | A lone non-canonical spelling (e.g. only `Host_CLI`) still decodes | More permissive | Inherent to the decoder; harmless without a collision. Locked by a test. |
| V3 | Unknown keys reported to the caller | Additive | Oracle ignores silently; Go loads successfully but makes typos observable (P2-D3). |
| V4 | `channel_id` uniqueness enforced at load | Stricter | Oracle enforces at ACP session start (`commands.rs:731`), which does not exist until P6/P9 (P2-D6). |
| V5 | `host_cli` empty message drops "to use ACP mode" | Cosmetic | Mode is retired (parent D1). |
| V6 | Secret fields omitted from the struct | Narrower | `#[serde(skip)]` in the oracle; owned by P3 (R6). |
| V7 | Non-empty `[[workspace]].path` must canonicalize | Stricter | Oracle passes it through unvalidated straight into a subprocess `current_dir()`. Treating each mapping path as an explicitly authorized workspace root closes a containment gap P6 would otherwise inherit. |
| V8 | `database.path` may not contain a `..` segment | Stricter | Oracle leaves it unrestricted while P4 will create parent directories there — a traversal primitive. Absolute paths remain permitted as an explicit, visible operator privilege. |
| V9 | Config file size bounded at 1 MiB | Stricter | Oracle reads unbounded. No legitimate config approaches this. |

Divergences V2, V4, V7, V8, V9 are the only ones that can reject a config the oracle accepts.
Each is a deliberate security or determinism control, documented for operators in E2.

## Validation Contract

`Validate()` is one unconditional, fail-fast pass returning on the first error. Order is fixed so
a given invalid config always yields the same error.

| Order | Rule | Error message | Oracle |
|---|---|---|---|
| 1 | `MaxConcurrentSessions == 0` | `max_concurrent_sessions must be greater than zero` | `config.rs:540` |
| 2a | `DefaultWorkspaceRoot == ""` | `default_workspace_root must be set` | — (see below) |
| 2b | canonicalization fails | `default_workspace_root invalid: {err}` | `config.rs:543-547` |
| 3 | per entry: `WorkspaceID == ""` | `workspace_id cannot be empty in [[workspace]] entry` | `config.rs:565` |
| 4 | per entry: `ChannelID == ""` | `channel_id cannot be empty in [[workspace]] entry` | `config.rs:570` |
| 5 | per entry: duplicate `WorkspaceID` | `duplicate workspace_id '{id}' in [[workspace]] entries` | `config.rs:576` |
| 6 | `HostCLI == ""` | `host_cli must be set` | `config.rs:663` (V5) |
| 7 | `HostCLI` absolute and does not exist | `host_cli '{path}' does not exist` | `config.rs:668` |
| 8 | duplicate `ChannelID` | `duplicate channel_id '{id}' in [[workspace]] entries; ACP routing requires each channel_id to map to exactly one workspace` | `config.rs:86-100` |
| 9 | per entry: non-empty `Path` fails canonicalization | `workspace path invalid for workspace_id '{id}': {err}` | — (V7) |
| 10 | `Database.Path` contains a `..` segment | `database.path must not contain '..' segments` | — (V8) |

Rules 3–5 and 9 iterate `Workspaces` in declaration order; the first offending entry wins. All
errors use `apperr.New`/`apperr.Newf` with `apperr.KindConfig`, per Constraint K1.

**Rule 2a is required.** Without it, an absent or explicitly empty `default_workspace_root`
reaches `pathsafe.NewRoot("")`, where `filepath.Abs("")` resolves to the **process working
directory** and `EvalSymlinks` succeeds — silently binding the workspace root to the CWD. Rust's
`canonicalize("")` returns `ENOENT`, so this is a Go-specific hazard with no oracle analogue.
`md.IsDefined` (decode step 4) catches the absent case; rule 2a catches the explicitly-empty case.

**Ordering is enforced by structure, not convention.** Rules 3–5 and rule 8 both operate on
`Workspaces` but sit on opposite sides of the `host_cli` rules. They are therefore implemented as
two **separate unexported stages** — `validateMappingFields` (3–5) and `validateChannelUniqueness`
(8) — invoked by `Validate()` at their respective positions. There is deliberately **no**
combined exported `ValidateWorkspaceMappings`: a single helper covering 3–5 *and* 8 cannot be
called at one point in the sequence without violating the contract.

**Canonicalization is committed only on success (Decision R10).** Rules 2b and 9 compute
canonical paths into locals. `Validate()` assigns them to `c.DefaultWorkspaceRoot` and
`c.Workspaces[i].Path` **only after every rule has passed**, so a failed validation never mutates
its receiver. This preserves the oracle's observable outcome (a successfully loaded config carries
canonical paths) without the oracle's in-place-mutation-on-partial-failure hazard.

**Advisory warning** (never fatal, appended to `Report.Warnings`): `HostCLI` is a bare name (no
separator) that `exec.LookPath` cannot resolve. This is the *only* advisory. The oracle's
additional "not in a standard install location" warning is **deliberately not ported**: the set of
standard locations is undefined, unverifiable, and the oracle discards the result anyway
(`main.rs:139`).

**`host_cli` execution contract for P6.** P2 stores `HostCLI` exactly as written (oracle parity).
Because a bare name is resolved against `PATH` at spawn time, P6 **must** resolve it once via
`exec.LookPath`, execute via direct argv (`exec.Command(resolved, HostCLIArgs...)`), and **must
never** pass it through a shell. Recorded here so the control is specified at the boundary that
defines the field, and inherited rather than rediscovered.

## Implementation Units

Thirteen units across five sub-epics. Each touches at most 3 files and at most 5 test scenarios,
and produces an independently verifiable outcome. Every acceptance criterion below is a test that
passes or fails — no prose judgement.

### Sub-epic A — Schema and defaults

**A1 — Declare the config type hierarchy**
Files: `internal/config/config.go`, `internal/config/detail_test.go`.
Declare all seven structs, `DefaultConfigPath`, `MaxConfigBytes`, `SlackDetailLevel` with its
three constants and a pointer-receiver `UnmarshalText` (accepting `minimal`/`standard`/`verbose`,
rejecting others with `apperr.KindConfig`), and `Report`. No decode, no validation, no defaults.
*Acceptance:* `go build ./...` and `go vet ./...` clean; a test asserts `UnmarshalText` accepts
each of the three valid values and rejects `"loud"` with a `KindConfig` error.
*Size:* S. *Complexity:* low.

**A2 — Implement `Default()` with all 17 defaults**
Files: `internal/config/config.go`, `internal/config/default_test.go`.
*Acceptance:* a table-driven test asserts all 17 values from the Default Contract; a test asserts
`Default().SlackDetailLevel == DetailStandard` **and** `DetailStandard != SlackDetailLevel(0)`;
a test asserts `Default().Timeouts.WaitSeconds == 0`.
*Size:* S. *Complexity:* low.

### Sub-epic B — Decode

**B1 — Add the TOML dependency; tolerant `Decode` with bounded unknown-key reporting**
Files: `go.mod`, `go.sum`, `internal/config/load.go`.
Add `github.com/BurntSushi/toml v1.6.0` **in this unit** — a unit adding the dependency without
importing it would be reverted by CI's `go mod tidy` dirty-check, so dependency and first import
must land together. Implement Decode steps 1, 2 and 5.
*Acceptance:* (i) a config with `wobble = 1` and `[nonsense]\nx = 2` loads without a decode error
and reports both dotted paths sorted; (ii) malformed TOML yields `apperr.KindConfig`;
(iii) `slack_detail_level = "verbose"` decoded **through `toml.Decode`** yields `DetailVerbose`,
proving the decoder honours `UnmarshalText` on a value field; (iv) a config with 200 unknown keys
reports 64 entries plus a `"… more"` marker, and a key containing `\n` is escaped;
(v) `go mod tidy` leaves `go.mod`/`go.sum` unchanged.
*Size:* M. *Complexity:* medium.

**B2 — Collision rejection, required key, bounded `Load`**
Files: `internal/config/load.go`, `internal/config/load_test.go`.
Implement Decode steps 3 and 4, and `Load` with the `MaxConfigBytes` guard.
*Acceptance:* (i) a config defining both `host_cli` and `Host_CLI` yields `KindConfig`;
(ii) a config omitting `default_workspace_root` yields `KindConfig`; (iii) `Load` on a
non-existent path yields `KindConfig` whose message contains the path and the operator guidance
text; (iv) `Load` on a file of `MaxConfigBytes+1` yields `KindConfig` without decoding.
*Size:* M. *Complexity:* medium.

**B3 — Lock the defaulting-and-override semantics**
Files: `internal/config/decode_semantics_test.go`.
Pure test unit — no production code. This is the regression barrier for R5.
*Acceptance:* (i) a minimal valid config yields all 17 defaults intact; (ii) `[stall]\nenabled =
false` yields `Stall.Enabled == false`; (iii) `[timeouts]\napproval_seconds = 0` yields `0`;
(iv) a lone `Host_CLI` key (no collision) decodes into `HostCLI`, documenting V2b;
(v) `http_port = -1` and `http_port = 70000` each yield a decode error.
*Size:* S. *Complexity:* low.

### Sub-epic C — Validation

**C1 — `Validate()` skeleton, rules 1, 2a, 2b, deferred commit**
Files: `internal/config/validate.go`, `internal/config/validate_test.go`.
*Acceptance:* (i) `MaxConcurrentSessions = 0` yields exactly `config: max_concurrent_sessions must
be greater than zero`; (ii) `DefaultWorkspaceRoot = ""` yields `config: default_workspace_root
must be set` and does **not** resolve to the CWD; (iii) a non-existent root yields a message
starting `config: default_workspace_root invalid:`; (iv) a valid `t.TempDir()` root is rewritten
to its canonicalized absolute form with no `\\?\` prefix; (v) a config failing rule 6 leaves
`DefaultWorkspaceRoot` **unmodified**, proving deferred commit.
*Size:* M. *Complexity:* medium.

**C2 — Mapping stages: rules 3–5 and rule 8, correctly ordered**
Files: `internal/config/validate.go`, `internal/config/workspace_test.go`.
Implement unexported `validateMappingFields` (3–5) and `validateChannelUniqueness` (8); wire both
into `Validate()` at their contract positions.
*Acceptance:* (i)–(iv) each of the four faults produces its exact message; (v) a config violating
**rule 6 and rule 8 simultaneously** returns the **rule 6** error, proving `host_cli` validation
runs before channel-uniqueness.
*Size:* M. *Complexity:* medium.

**C3 — `host_cli` rules 6–7, PATH advisory, order determinism**
Files: `internal/config/validate.go`, `internal/config/hostcli_test.go`.
*Acceptance:* (i) empty `host_cli` yields `config: host_cli must be set`; (ii) an absolute
non-existent `host_cli` yields `config: host_cli '{path}' does not exist`; (iii) a bare name not
resolvable by `exec.LookPath` produces **no error** and exactly one `Report.Warnings` entry;
(iv) a config violating rules 1, 6 and 8 simultaneously returns the **rule 1** error.
*Size:* M. *Complexity:* medium.

**C4 — Path containment rules 9 and 10**
Files: `internal/config/validate.go`, `internal/config/paths_test.go`.
*Acceptance:* (i) a `[[workspace]]` with a non-existent `path` yields `config: workspace path
invalid for workspace_id '{id}': …`; (ii) a `[[workspace]]` whose `path` is a valid `t.TempDir()`
is rewritten to its canonical form on success; (iii) `database.path = "../../etc/db"` yields
`config: database.path must not contain '..' segments`; (iv) an absolute `database.path` is
**accepted** (explicit operator privilege) and left unmodified.
*Size:* M. *Complexity:* medium.

### Sub-epic D — Workspace routing

**D1 — Resolvers with default-root fallback**
Files: `internal/config/workspace.go`, `internal/config/resolve_test.go`.
Implement `ResolveChannelID`, `ResolveWorkspaceByChannelID` (by value), and
`WorkspaceRootForChannel` (first match wins; falls back to `DefaultWorkspaceRoot`).
*Acceptance:* (i) a known workspace resolves to its channel; (ii) an unknown workspace returns
`false`; (iii) a channel whose mapping sets `path` returns that path, and one without returns
`DefaultWorkspaceRoot`; (iv) all three resolvers are safe on a `Default()` config with no
mappings.
*Size:* S. *Complexity:* low.

### Sub-epic E — Operator surface and coupling

**E1 — `config.toml.example` with a bidirectional drift test**
Files: `config.toml.example`, `internal/config/example_test.go`.
The example must be a **valid, loadable** config using synthetic values only, with
`default_workspace_root = "."` so it validates against the repository root without templating.
*Acceptance:* (i) a test loads `config.toml.example` and asserts `err == nil` and
`len(Report.UnknownKeys) == 0` — catching example keys absent from the struct; (ii) a
**reflection-based** test walks every `toml`-tagged field in the `Config` hierarchy and asserts
each key path appears in the example, catching struct fields absent from the example. Both
directions are covered.
*Size:* M. *Complexity:* medium.

**E2 — `docs/config-reference.md` with a completeness test**
Files: `docs/config-reference.md`, `internal/config/docs_test.go`.
Document every key, its default, and its validation rule; the nine divergences V1–V9; and an
**"Operator migration from agent-intercom (Rust)"** section covering: duplicate `channel_id` is
now a startup failure (V4); `--mode`/`protocol_mode` are gone and appear as unknown keys (V3);
tokens in `config.toml` are ignored and reported as unknown keys, with credentials resolved
outside the file (V6, forward reference to P3).
*Acceptance:* a test asserts every one of the 17 default field names and all 10 validation
message strings appear verbatim in `docs/config-reference.md`, so doc completeness is enforced
rather than assumed; markdownlint passes.
*Size:* M. *Complexity:* low.

**E3 — Align the `--config` default and the README**
Files: `cmd/intercom/main.go`, `README.md`.
`cmd/intercom/main.go` currently registers `--config` with default `""`, diverging from the
oracle's `config.toml` (`main.rs:60`) and from the guidance text in `Load`'s error. Change **only
the default string** to `config.DefaultConfigPath`. `RunE` still returns not-implemented; no
config is read. Add a README configuration section linking `config.toml.example` and
`docs/config-reference.md`.
*Acceptance:* (i) the existing `cmd/intercom` test suite still passes; (ii) a test asserts the
`--config` flag's `DefValue` equals `"config.toml"`; (iii) markdownlint passes on README.
*Size:* S. *Complexity:* low.

## Dependency Graph

```text
002-F (P1, shipped) ──► C1, C4 (pathsafe.NewRoot)

A1 ──► A2 ──► B1 ──► B2 ──► B3 ──► C1 ──► C2 ──► C3 ──► C4 ──► D1 ──► E1 ──► E2 ──► E3
```

Strictly sequential; no unit starts before its predecessor merges. Ship executes them in the
listed order on a single branch in a single worktree (P-016).

**Rollback boundary.** Every unit leaves the tree building and green, and no unit depends on a
later one. The shipment can be abandoned cleanly after any unit; the partial result is a
smaller-but-coherent `internal/config` with no importers, since runtime wiring is P4+. The only
externally visible artifacts land in E1–E3, at the very end.

## Plan Hardening Signals

`Requires plan hardening: yes` — three signals: (1) a **new third-party dependency** enters the
supply chain; (2) the slice touches the **workspace path-containment security boundary**; (3) it
defines the **operator-facing configuration contract**, whose defaults are load-bearing for every
later phase.

## Plan Hardening

### Protected invariants

| ID | Invariant | Enforcement |
|---|---|---|
| I1 | `internal/apperr` and `internal/pathsafe` are **not modified** | Scope guard; Constraint K1 removes the only pressure to change them. |
| I2 | No secret-bearing field exists in the P2 config struct | Decision R6; `gitleaks` in CI; fixtures synthetic by construction. |
| I3 | Decode never fails on unknown keys | B1 acceptance (i). |
| I4 | Explicit `false`/`0` always beats a non-zero default | B3 acceptance (ii)/(iii). |
| I5 | Validation order is fixed and structurally enforced | C2 acceptance (v) and C3 acceptance (iv); two separate mapping stages. |
| I6 | Failed validation never mutates its receiver | C1 acceptance (v). |
| I7 | `go.mod` gains exactly one direct dependency, pinned, no transitive additions | B1 acceptance (v); CI `go mod verify` + tidy dirty-check. |
| I8 | P2 changes no runtime behavior in `cmd/**` | E3 changes one flag default only; `RunE` still returns not-implemented. |
| I9 | Config never widens the path-containment boundary | Rules 2, 9, 10 (V7/V8); C1 and C4 acceptance. |

### Risky actions

- **Adding `BurntSushi/toml`.** Pin v1.6.0 exactly; `go mod verify` and `govulncheck` are already
  CI-gated; confirm zero transitive dependencies. If `govulncheck` flags it, **halt and escalate**
  rather than switching to an unvetted alternative.
- **Rejecting case-fold collisions (V2).** A stricter-than-oracle rule. Must reject only true
  collisions, never a lone non-canonical spelling — B3 acceptance (iv) is the guard against
  over-rejection.
- **Canonicalizing `[[workspace]].path` (V7).** Makes previously-loadable configs fail when a
  mapping path does not exist. Deliberate, documented in E2's migration section.
- **Deferred canonicalization commit.** The natural implementation mutates as it goes; the
  contract requires locals plus a commit phase. C1 acceptance (v) is the guard.

### Forbidden commands (require explicit operator approval)

- Any write, commit, checkout, or clean inside `C:\Source\GitHub\intercom` (the oracle) or
  `references/herdr`.
- Editing `internal/apperr/**`, `internal/pathsafe/**`, `.golangci.yml`,
  `.github/workflows/**`, `scripts/**`.
- Any change to `cmd/intercom/main.go` beyond the single flag-default string in E3.
- `go get -u` or any unpinned dependency upgrade.
- `git push --force`, history rewrite, or squash merge (Principle XI).

### Deepened verification

The shipment is not complete until:

1. `go build ./...` with `CGO_ENABLED=0` succeeds for all four targets in `scripts/targets.json`.
2. `go test -race ./...` passes.
3. `go vet`, `gofmt -l`, `goimports -l`, `golangci-lint`, `staticcheck`, `govulncheck`, and
   `gitleaks detect` all pass.
4. `go mod tidy` produces no diff.
5. Every one of the 17 defaults and all 10 validation rules is covered by an asserting test.
6. `config.toml.example` loads with zero unknown keys, and the bidirectional drift test passes.

### Deepened closure

Closure must record: the oracle commit (`41df772`); the resolved divergences V1–V9; the final
dependency set; confirmation that P3 remains unstarted in stash `037B1552`; and an appended P2
parity-traceability table in `docs/oracle-pin.md` mapping the 17 defaults and 10 rules to their
oracle lines.

### Operator checkpoints

- **Before B1** — dependency introduction. If the operator prefers zero new dependencies, the
  fallback is a hand-rolled TOML subset parser, a materially larger and riskier slice. Flag rather
  than assume.
- **Before C4** — divergences V7/V8 reject configs the oracle accepts. Confirm the operator wants
  the stricter containment posture rather than oracle parity.

### Review-gate capability risks

Cross-platform path behavior (rules 2, 9, UNC stripping) cannot execute on the Linux-only PR CI.
It is covered by `pathsafe`'s existing Windows-aware tests from P1 and by local verification; the
cross-compile matrix proves it *builds* everywhere, not that it *behaves* everywhere. Accepted
limitation, unchanged from P1.

### Unresolved operator decisions

None blocking. The two checkpoints above are confirmations, not open questions.

## Risks and Caveats

| Risk | Impact | Mitigation |
|---|---|---|
| Ported default count wrong | High | Corrected to 17 with per-line citations (addendum F1); A2 asserts each. |
| `DetailStandard` collides with iota zero | Medium | A2 asserts `DetailStandard != 0`. |
| Empty `default_workspace_root` silently binds to CWD | **High** | Rule 2a plus decode step 4; C1 acceptance (ii). Go-specific hazard with no oracle analogue. |
| Rule ordering violated by a combined mapping helper | High | Two separate unexported stages; C2 acceptance (v). |
| Acceptance criteria satisfiable in unintended ways | Medium | Every criterion names a concrete assertion; the two prose criteria (docs completeness, example drift) were converted to reflection/string-presence tests in E1/E2. |
| Dependency added in a unit that does not import it | Medium | B1 bundles dependency + first import. |
| Config tests depend on ambient filesystem | Medium | `t.TempDir()` exclusively; the sole exception is E1's example, which uses `"."`. |
| Scope creep into P3 credentials | High | Decision R6 removes the fields; R7 fixes the package boundary permanently. |
| Divergences forgotten by later phases | Medium | Recorded in the plan, the addendum, and `docs/config-reference.md` (E2). |
| Hostile config exhausts memory or forges log lines | Medium | `MaxConfigBytes` (V9); unknown-key sanitization and caps (B1 acceptance iv). |

## Plan Review

### Attempt history

| Attempt | Verdict | Notes |
|---|---|---|
| 1 | **FAIL** | 1 P0 and 8 P1 findings across five personas. |
| 2 | **PASS** | P0 and all P1s remediated; 7 P2/P3 findings adopted, 4 accepted as recorded residuals. |

<!-- plan-review-attempt: 2 -->

### Persona coverage

Architecture Strategist, Go Reviewer (correctness/idiom), Security Reviewer, Scope Boundary
Auditor, and Schema-CLI-Docs Coupling Reviewer ran independently and in parallel on attempt 1,
each on a different model. Concurrency Reviewer was **not** engaged: the slice introduces no
goroutines, channels, or shared mutable state — the hot-reload watcher that would require it is
explicitly deferred.

### Attempt-1 P0 — closed

**P0-a (Go Reviewer) — the plan was unimplementable as written.**
Rule 2 and unit C1 required an error that carried both the oracle's message text *and* a retained
cause for `errors.As`. `internal/apperr` cannot express that: `New`/`Newf` leave `cause == nil`,
`Wrap` forces `msg = cause.Error()`, and the struct fields are unexported. Satisfying the
criterion required adding a constructor to `apperr`, which invariant I1 forbids — an internal
contradiction that would have stalled Ship mid-shipment.
The Architecture reviewer reached the same conclusion independently, and further noted that
`pathsafe.NewRoot` already stringifies its cause, so `errors.Is(err, os.ErrNotExist)` was
unreachable regardless of how P2 wrapped.
*Remediation:* verified directly against `internal/apperr/apperr.go` — confirmed. Added
**Constraint K1** as a first-class plan section, and checked the oracle: `AppError::Config` is a
`String` variant that also stringifies, so cause-retention was **beyond-oracle invention, not
parity**. All error construction is now specified as `apperr.New`/`Newf`, and no unit asserts
cause retrieval. The broadened evidence is recorded against stash `4A91CA81` (which already
carries `Requires deliberation: yes`) without acting on it.

### Attempt-1 P1 findings — all closed

**P1-b (Architecture + Coupling, found independently) — validation-order contradiction.**
The exported `ValidateWorkspaceMappings` was specified to cover rules 3–5 *and* rule 8, but the
contract places rules 6–7 between them. Calling it at any single point violated the order. The
original C3 test used faults 1/6/8, where rule 1 masks the defect entirely.
*Remediation:* split into two unexported stages invoked at their contract positions; added C2
acceptance (v), a rule-6-versus-rule-8 conflict that fails if the order is wrong.

**P1-c (Coupling) — empty `default_workspace_root` binds to the CWD.**
`filepath.Abs("")` returns the process working directory and `EvalSymlinks` succeeds, so
`pathsafe.NewRoot("")` silently canonicalizes to the CWD — a workspace-isolation break with no
oracle analogue (Rust's `canonicalize("")` returns `ENOENT`).
*Remediation:* added rule 2a and C1 acceptance (ii).

**P1-d (Coupling) — the example config could not pass its own test.**
`Decode` runs `Validate`, whose rule 2 requires an existing directory; a synthetic placeholder
path would fail. The original criterion said the example validates "against a `t.TempDir()` root"
without explaining how a static file obtains a dynamic path.
*Remediation:* the example uses `default_workspace_root = "."`, which exists wherever the test
runs; no templating needed.

**P1-e (Coupling) — one-directional drift test.**
Asserting zero unknown keys catches example keys missing from the struct, but not struct fields
missing from the example — the more likely direction as the schema grows.
*Remediation:* E1 acceptance (ii) adds a reflection-based walk over every `toml`-tagged field.

**P1-f (Security) — case-fold collisions select a field nondeterministically.**
`BurntSushi/toml`'s `EqualFold` fallback means `host_cli` and `Host_CLI` both target `HostCLI`;
because the decoder iterates a Go map, the winner is nondeterministic. A config could shadow
`host_cli` (the binary later executed) or `default_workspace_root` (the containment root).
Attempt 1 had accepted this as a benign permissive divergence (SEC-6).
*Remediation:* V2 inverted from "accepted" to **rejected**: collisions now fail the load. B2
acceptance (i) tests rejection; B3 acceptance (iv) guards against over-rejecting a lone variant.

**P1-g (Security) — `[[workspace]].path` was an unvalidated subprocess-CWD escape.**
Passed through untouched to P6's `current_dir()`, an absolute or `../../` path bypasses the
canonical `DefaultWorkspaceRoot` entirely, contradicting the plan's own claim that config never
widens the containment boundary. "The oracle does it too" is not adequate justification for a
security-control gap that a later phase inherits.
*Remediation:* rule 9 (V7) canonicalizes every non-empty mapping path via `pathsafe.NewRoot`,
making each an explicitly authorized root. Unit C4 added.

**P1-h (Security) — `database.path` was an arbitrary-write primitive.**
Unrestricted, while P4 will create parent directories there.
*Remediation:* rule 10 (V8) rejects `..` segments. Absolute paths remain permitted as an
explicit, visible operator privilege — the reviewer's own stated acceptable alternative — and are
documented as such.

**P1-i (Architecture) — no unit owned the `Decode` → `Validate` integration.**
C1 authorized only `validate.go`, so its criteria were satisfiable by calling `Validate`
directly, leaving `Load` able to return an unvalidated config.
*Remediation:* the Decode Contract now names step 6 explicitly, C1's file list is scoped to
`validate.go` while B-series `Load`/`Decode` tests exercise the integrated path, and report-merge
and report-on-error semantics are specified.

### Attempt-2 adopted improvements (P2/P3 findings)

- **Architecture:** `Validate()` no longer mutates on partial failure — canonical paths are
  committed only after every rule passes (Decision R10, I6, C1 acceptance v).
- **Architecture:** `ResolveWorkspaceByChannelID` returns a value rather than a pointer into the
  mutable `Workspaces` slice.
- **Architecture:** the permanent P3 package boundary is fixed (Decision R7) so credentials cannot
  invert the dependency by attaching to `config.Config`.
- **Architecture:** the relative-path base is decided now (Decision R9 — process working
  directory) so P4 and P6 cannot choose incompatible semantics.
- **Scope:** the undefined "standard install location" advisory was **dropped** rather than
  implemented — it was unverifiable, had no consumer, and the oracle discards it.
- **Scope:** `ValidateWorkspaceMappings` is no longer exported. Exporting for a deferred watcher
  was YAGNI, and the ordering fix already required unexported stages. This adjusts deliberation
  P2-D6, which is annotated accordingly.
- **Scope:** E1 was split into E1/E2 (example vs. reference doc); the original single S-sized unit
  understated the doc work. E2 gained a machine-checked completeness assertion.
- **Go:** B1 acceptance (iii) added — decoding `slack_detail_level` **through** `toml.Decode`,
  since testing `UnmarshalText` directly would not prove the decoder invokes it on a value field.
- **Go:** the redundant `host_cli` `IsDefined` check was removed; rule 6 now covers the absent and
  the explicitly-empty cases with a single message, eliminating two-messages-for-one-error.
- **Security:** `MaxConfigBytes` (V9) bounds the read; unknown-key output is sanitized, truncated,
  and capped against log forging and amplification.
- **Coupling:** E3 added — the `--config` flag default was `""` against the oracle's
  `config.toml`, while `Load`'s error text tells operators to pass `--config`.
- **Coupling:** E2 gained an operator migration section; `docs/oracle-pin.md` gains a P2 parity
  table at closure.

### Residual findings — accepted, recorded

| ID | Finding | Disposition |
|---|---|---|
| GO-15 | `WorkspaceMapping.Label`/`Path` collapse `Option<String>` to `""` | Accepted — the oracle never distinguishes `None` from `Some("")`. |
| GO-16 | Double-prefixed message (`config: … : path violation: …`) when rule 2b stringifies a `pathsafe` error | Accepted; truthful and oracle-shaped. C1 acceptance (iii) pins the prefix so it cannot drift unnoticed. |
| SEC-7 | `host_cli` bare names remain PATH-resolved at spawn time, so a PATH-hijack window persists | Accepted for P2 — resolving at load would not close the TOCTOU window anyway. The mitigation is specified as a binding P6 contract (`exec.LookPath`, direct argv, never a shell). |
| ARCH-7 | `Report.Warnings` has no consumer until P4 wires config into the runtime | Accepted — reduced to a single well-defined producer after the undefined advisory was dropped. The structure is needed by `Decode`'s signature regardless. |

### Gate rationale

**PASS.** The P0 is closed by a constraint that removes the contradiction at its root rather than
papering over it, and it was verified directly against source rather than taken on report. All
eight P1s are closed with structural remediations — new validation rules, split stages,
reflection-based drift detection, and bounded input — not with review notes. Two independent
reviewers converged on the ordering defect and two on the empty-root defect, which raises
confidence that the remaining surface is well explored. Scope is bounded by an exclusion table, an
authorized-path guard, and nine invariants; the Scope auditor found no P0/P1 boundary violations.
Every acceptance criterion is a named, machine-verifiable assertion, and the two prose criteria
flagged in attempt 1 were converted into tests.

## Provenance

- Deliberation: `docs/decisions/2026-09-03-intercom-go-p2-config-deliberation-addendum.md`
- Parent deliberation: `docs/decisions/2026-09-03-intercom-go-port-continuation-deliberation.md`
- Brief: `docs/decisions/2026-07-06-go-port-reference-brief.md`
- Oracle: `softwaresalt/agent-intercom` @ `41df772`, read-only, out-of-tree, never modified.
- Predecessor: `002-F` / `002-S`, merge `9efb156`, closure
  `docs/closure/002-S-002-F-post-merge-closure.md`.
- Successor: P3 credential resolution, preserved in stash `037B1552`.

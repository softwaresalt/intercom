---
title: "intercom-go P3 — Credential Resolution, Provider Seam, and the Redacting Secret Type"
description: "Deliberate the P3 credential slice of stash 037B1552: oracle-grounded precedence, keychain dependency vs provider seam, platform/headless behavior, and secret redaction design"
topic: "How should intercom-go resolve Slack credentials, and what dependency, seam, and redaction design should carry it?"
depth: "deep"
decision_status: "decided"
promoted_to: "plan"
linked_artifacts:
  - "docs/plans/2026-09-03-intercom-go-p3-credentials-plan.md"
tags:
  - "intercom-go"
  - "credentials"
  - "keychain"
  - "secret-redaction"
  - "go-port"
  - "p3"
---

# intercom-go P3 — Credential Resolution, Provider Seam, and the Redacting Secret Type

**Stash:** `037B1552` (P3 residual scope) · **Roadmap parent:** `4989A42D` (P4+ preserved)
**Predecessor:** `003-F` / shipment `003-S` (P2 config, SHIPPED, merge `40c7d289`, closure `2a547913`)
**Oracle:** `softwaresalt/agent-intercom` @ `41df772` — read-only sibling at `C:\Source\GitHub\intercom`, never modified.

## Problem Frame

P0–P2 delivered the foundation, `internal/apperr` (14-variant taxonomy), `internal/pathsafe`, and
`internal/config` (P2). P2 established a **binding, permanent boundary** (plan Decision R7):
`internal/config` is a secret-free file-schema package. `config.toml` carries no credential values;
`app_token`, `bot_token`, `team_id`, and `authorized_user_ids` are all `#[serde(skip)]` in the
oracle and are never read from the config file.

P3 is the residual scope of the original config+errors slice: **credential resolution**. It must
deliver the oracle's precedence chain, a credential-source abstraction spanning OS keychain and
environment variables, and a redacting `Secret` type that the oracle does not have.

**Who cares and why.** Operators migrating from the Rust `agent-intercom` need their existing
keychain entries and environment variables to keep working unchanged. Reviewers and the eventual
Slack slice (P8) need a credential value that cannot leak through logs, errors, or JSON. The
dark-factory execution model needs machine-verifiable acceptance criteria and deterministic,
network-free, keychain-free tests.

**Constraints.**

* **CGO-free shipped binaries.** `scripts/build.ps1:85` and `.github/workflows/ci.yml:115,262` set
  `CGO_ENABLED=0`; the cross-compile matrix (`scripts/targets.json`) covers `linux/amd64`,
  `windows/amd64`, `darwin/amd64`, `darwin/arm64`. Only the `-race` test job is cgo-capable.
  Any credential dependency must cross-compile CGO-free on all four targets.
* **Go 1.22 language floor**, toolchain pinned `go1.26.5` (`go.mod`).
* **`log/slog` is the mandated structured-logging surface** (`technology-go.instructions.md:75`).
* **Workspace instruction floors outrank the port brief** — compound learning
  `external-spec-yields-to-workspace-instructions-2026-09-03.md`.
* CI gates: `golangci-lint` v2.13.2, `govulncheck` v1.7.0, `go test -race`.
* No secret values may enter fixtures, docs, logs, errors, or the repository. `.env.local` exists
  at both repo roots and was never opened.

**Success criteria.**

1. Precedence is byte-for-byte oracle-equivalent for every credential.
2. A secret value provably cannot reach `fmt`, `log/slog`, `encoding/json`, or any error string.
3. Every test is deterministic, hermetic, and never touches a real OS keychain.
4. The P2 secret-free config boundary is structurally preserved, not merely observed.
5. Rollback is a single revert with no state or schema migration.

**Explicitly out of scope.** P4+ (models/persistence/Slack/ACP), oracle-drift gate `8ACF7110`,
CI-hardening follow-ups, P1/P2 review follow-ups, multiplexer `A92E3FA0`, operator rotation
`EF9352FB`, any keychain **write**/provisioning path, and `AppState::ipc_auth_token` (a distinct
per-process secret belonging to the IPC slice, P13).

## Research Findings

Primary research was a full read of the oracle's credential path (`src/config.rs`), plus targeted
dependency verification. Every finding below is source-grounded; the six labelled **corrections**
supersede claims recorded in stash `037B1552`.

### F1 — Precedence is keychain-first, per-field, four tiers (stash CONFIRMED)

`src/config.rs:857-928` (`load_credential`) is a flat sequence of early-return guards — no
combinator chain:

1. keychain service `mode_keychain_service(mode)`, account `{keyring_key}`
2. env `{ENV_KEY}{mode_suffix}`
3. keychain service `KEYCHAIN_SERVICE` (`agent-intercom`)
4. env `{ENV_KEY}`

Keychain-first is confirmed; it is **not** env-with-keychain-fallback. Empty string is treated as
absent at every tier (`src/config.rs:837-840`, `:878`, `:902`). Resolution is **per-field**
(`src/config.rs:452-460`), so different fields may legitimately resolve from different tiers.

### F2 — CORRECTION: `authorized_user_ids` does not use the keychain chain at all

Stash `037B1552` groups all four credentials under one chain. It is wrong.
`load_authorized_users` (`src/config.rs:484-508`) is a **separate, synchronous, env-only,
two-tier** path that never consults the keychain:

* `SLACK_MEMBER_IDS_ACP` → `SLACK_MEMBER_IDS` (shared fallback **ungated** by mode, unlike
  `load_credential`)
* parses CSV: `split(',')` → `trim()` → drop empties (`src/config.rs:494-498`)
* has its own distinct error message, not the `credential ... not found` form
* there is no `slack_member_ids` keychain account anywhere

A port that routes all four fields through one generic resolver would be incorrect.

### F3 — CORRECTION: tiers 3–4 are hard-gated, and the surviving mode decides the chain

`src/config.rs:890` gates tiers 3–4 behind `if mode != ServerMode::Mcp`. In MCP mode the function
runs **two** tiers; in ACP mode, four. Mode comes from exactly one place — the `--mode` clap flag
(`src/main.rs:83-89`, passed at `:132`) — never from env, never from `config.toml`
(`GlobalConfig` has no `mode` field).

This makes "which mode survives retirement" the load-bearing question for P3. The answer is
already settled in-repo: the port brief says *"Port ACP-first"* (`2026-07-06-go-port-reference-brief.md:43`),
continuation decision D1 retires MCP **as a server mode**, and P2-D5 concluded *"there is exactly
one mode"*. **ACP is the surviving runtime**, so the ACP chain — all four tiers — is the behavior
to port. The chain becomes a static constant with no `mode` parameter, which is strictly simpler
than the oracle: no mode plumbing, no `if mode != Mcp` gate, no dual error-message form, and no
redundant lookups (tiers 3–4 are genuinely distinct from 1–2 under ACP).

### F4 — Namespaces and env names (stash CONFIRMED)

`KEYCHAIN_SERVICE = "agent-intercom"` (`src/config.rs:804`); ACP variant is the hardcoded literal
`"agent-intercom-acp"` (`src/config.rs:821-827`) — a two-arm match, not string interpolation.
`keyring::Entry::new(service, key)` passes `keyring_key` as the **account** (`src/config.rs:833`).

| Field | Keychain account | Env (ACP) | Env (shared) |
|---|---|---|---|
| `app_token` | `slack_app_token` | `SLACK_APP_TOKEN_ACP` | `SLACK_APP_TOKEN` |
| `bot_token` | `slack_bot_token` | `SLACK_BOT_TOKEN_ACP` | `SLACK_BOT_TOKEN` |
| `team_id` | `slack_team_id` | `SLACK_TEAM_ID_ACP` | `SLACK_TEAM_ID` |
| `authorized_user_ids` | *(none — env only)* | `SLACK_MEMBER_IDS_ACP` | `SLACK_MEMBER_IDS` |

`SLACK_TEAM_ID_ACP` is real but emergent — synthesized by `format!` at `src/config.rs:864`, never
written literally, absent from `README.md:201` and `config.toml.example:172`, and untested.

### F5 — Required/optional semantics (stash CONFIRMED)

`app_token` and `bot_token` are **required** (`?` at `src/config.rs:454-455`). `team_id` is
**optional and never fails**, defaulting to `""` via `load_optional_credential`
(`src/config.rs:928-941`). `authorized_user_ids` is **required non-empty**
(`src/config.rs:499-504`). Error type is `AppError::Config(String)`, rendered `"config: {msg}"`.
Error messages name the **source** (service and env-var names) and never the value.

An oracle wart: in MCP mode the authorized-users message renders `set SLACK_MEMBER_IDS or
SLACK_MEMBER_IDS ...` — a duplicated name. Under the ACP-only chain this cannot occur.

### F6 — CORRECTION: keychain error classes are collapsed and undiagnosable

`try_keyring` returns `std::result::Result<String, ()>` (`src/config.rs:829-841`), discarding the
`keyring::Error` entirely. "No backend available", "keychain locked", "access denied", "entry not
found", `JoinError`, and "entry present but empty" all produce byte-identical behavior and
byte-identical operator-facing text. A locked macOS keychain is indistinguishable from an unset
credential.

### F7 — CORRECTION: the oracle's keychain tiers may be inert in production

`Cargo.toml` declares `keyring = "3"` with **no feature list**; `Cargo.lock` resolves 3.6.3.
keyring 3.x makes platform stores opt-in by feature. With no features, the crate selects its
mock/no-op keystore, which would make tiers 1 and 3 fail unconditionally and reduce the effective
chain to env-only on every platform. This is consistent with — and would fully explain — the test
suite's stated assumption that the keychain is "almost certainly absent"
(`tests/unit/credential_loading_tests.rs:59-60`) and the total absence of any keychain-hit test.
Marked INFERENCE; not verified by running the oracle binary. It does not change what the Go port
should implement (the *documented, intended* contract is keychain-first), but it means the Go port
may be the first working keychain implementation, so its keychain path must be tested on its own
merits rather than by trusting oracle parity.

### F8 — CORRECTION: the oracle never tests its own precedence

`tests/unit/credential_loading_tests.rs:3` claims to validate "keychain precedence". No test in
the oracle sets a keychain entry or asserts keychain-beats-env. Every test **depends on the
keychain being empty**. There is no seam: `try_keyring` calls `keyring::Entry::new` directly with
no trait, no injection, no `#[cfg(test)]` override, and `Select-String` for keychain symbols
across `tests/` returns zero hits.

Consequence: all eight tests silently invert and **fail on any machine that actually has
`agent-intercom` credentials provisioned**, because tier 1 would beat the asserted tier-2 env
value. This is the single strongest argument for a provider seam in the Go port.

### F9 — CORRECTION: the credential test corpus is 14, not 8

Eight is exact for `tests/unit/credential_loading_tests.rs`. Six more live in
`tests/unit/config_tests.rs` — including the **only** redaction test
(`slack_config_debug_redacts_tokens:696`), the **only** `load_authorized_users` failure test
(`missing_authorized_user_ids_env_var_fails:242`), the **only** `#[serde(skip)]` proof
(`credential_env_fallback:265`), and the keychain-service-name regression guard
(`keychain_service_constant_is_agent_intercom:289`). Porting only the eight would drop redaction
and secret-free-config coverage entirely.

### F10 — Redaction: no `Secret<T>`, and the redaction that exists is leaky (stash CONFIRMED, extended)

No `Secret<T>`, no `secrecy`, no `zeroize` (`Cargo.toml`). `SlackConfig` has a hand-written
`Debug` (`src/config.rs:133-147`) printing `[REDACTED]` for `app_token`/`bot_token` while leaving
`team_id` **and** `channel_id` visible. Gaps:

* `GlobalConfig` **derives** `Debug` (`src/config.rs:343`), so `{:?}` prints the entire
  `authorized_user_ids` operator ACL in full — the careful redaction one level down is undermined
  one level up.
* No `Display` redaction anywhere, no tracing field filters, no serialization redaction.
* Credential-load logging is safe only by construction: `tracing::info!` logs `key`, `source`,
  `service`, `var` — never `value` (`src/config.rs:867-908`).

The Go `Secret` type is therefore an **improvement, not a port**, and its API is unconstrained by
the oracle.

### F11 — Secret-bearing field inventory (one addition to the stash)

Exhaustive search for `signing_secret|webhook|api_key|password|secret|DATABASE_URL|token` across
`src/` and `ctl/` found **no** Slack signing secret (consistent with Socket Mode), no webhooks, no
third-party API keys, and no DB credentials (SQLite, filesystem path only). One secret the stash
omits: `AppState::ipc_auth_token` (`src/state/mod.rs:133-134`), a random UUIDv4 generated per
process (`src/main.rs:223-224`), validated at `src/ipc/server.rs:195-200`, and **not redacted**.
It is outside the credential-resolution flow (no keychain, no env) and belongs to P13.

Related hardening worth carrying forward later: the ACP spawner does `env_clear()` plus a strict
allowlist so tokens never reach child agents (`src/acp/spawner.rs:109-118`).

### F12 — Integration point and ordering

`config.load_credentials(args.mode).await?` is called **exactly once, from exactly one place**
(`src/main.rs:132`) — after TOML parse+validate and CLI overrides, before mode validation, DB
init, `AppState`, and all network I/O, and before the config is frozen into `Arc`
(`src/main.rs:161`). Credentials are **never hot-reloaded**: `ConfigWatcher` parses only workspace
mappings (`src/config_watcher.rs:40-60`).

In intercom-go, `cmd/intercom/main.go` does **not** yet load config at all — `RunE` logs one line
and returns `errNotImplemented`. `TestRootCmd_LogLevelDebugEmitsJSON` parses the whole of stderr
as a **single** JSON object, so any additional startup log line breaks it. There is no startup
pipeline to integrate into yet.

### F13 — Dependency verification (measured, not assumed)

Per the compound learning `go-race-requires-cgo-fails-closed-2026-09-03.md`, the candidate was
verified empirically rather than trusted:

| Check | Result |
|---|---|
| `github.com/zalando/go-keyring` version | v0.2.8 |
| Transitive deps | `danieljoos/wincred` v1.2.3, `godbus/dbus/v5` v5.2.2, `golang.org/x/sys` v0.27.0 — all pure Go |
| `CGO_ENABLED=0` build, `linux/amd64` | **OK** |
| `CGO_ENABLED=0` build, `windows/amd64` | **OK** |
| `CGO_ENABLED=0` build, `darwin/amd64` | **OK** |
| `CGO_ENABLED=0` build, `darwin/arm64` | **OK** |
| `govulncheck` v1.7.0 | **No vulnerabilities found** (0 affecting) |
| Licenses | MIT (go-keyring), MIT (wincred), BSD-2-Clause (godbus) — all permissive |

Platform internals, audited:

* **darwin** — shells out to the **absolute** path `/usr/bin/security`
  (`keyring_darwin.go:29`), so there is no `PATH`-hijack surface.
* **windows** — `danieljoos/wincred`, pure-Go syscalls against the Windows Credential Manager.
* **linux/unix** — `godbus` Secret Service over D-Bus.
* **unsupported platforms** — `fallbackServiceProvider` returns a distinct
  `ErrUnsupportedPlatform` (`keyring_fallback.go`).

Decisively for F6: the library **distinguishes `ErrNotFound` from `ErrUnsupportedPlatform`** and
from backend errors, so the Go port can fix the oracle's collapsed error classes at zero cost.
The library also exposes `Set`/`Delete`/`DeleteAll` — a write surface the oracle does not have and
the port does not want.

### F14 — Existing Go conventions to match

`internal/apperr` pins: unexported fields so errors cannot be re-categorized after construction
(GO-8), pointer receivers with `*Error` constructors (GO-9), immutable value-type kind sentinels
(GO-1, P0), `Is` by kind and `Unwrap` for cause as deliberately separate mechanisms, and a frozen
display contract `"{prefix}: {msg}"`. `internal/config` exposes a `Report` type for advisory
warnings alongside hard errors — the natural precedent for source attribution.

## Options Evaluated

### Decision P3-D1 — Which precedence chain survives mode retirement

| Option | Description | Assessment |
|---|---|---|
| (a) **Static four-tier ACP chain** | `agent-intercom-acp` keychain → `*_ACP` env → `agent-intercom` keychain → shared env, as a fixed constant with no mode parameter | ACP is the surviving runtime (brief §43, D1, P2-D5), so ACP's behavior *is* the behavior. Strict superset: accepts both ACP-namespaced and shared credentials, so no migrating operator breaks. Removing the mode parameter deletes the oracle's gate, dual error form, and redundant lookups. **Chosen.** |
| (b) Two-tier shared-only chain | `agent-intercom` keychain → shared env; drop all `_ACP` surface | Simplest, but silently breaks any operator whose credentials live only in the ACP namespace, with a "not found" error that never mentions the namespace it ignored. Discards real oracle behavior on the strength of retiring the *other* mode. **Rejected.** |
| (c) Configurable chain | Expose namespace/suffix as config or flags | Re-introduces the mode concept P2-D5 retired, under a new name. Adds operator surface with no demonstrated demand, and makes precedence non-deterministic across deployments. **Rejected (YAGNI).** |

### Decision P3-D2 — Keychain dependency vs interface-and-adapters

| Option | Description | Assessment |
|---|---|---|
| (a) Direct library call, no seam | Call `go-keyring` inline in the resolver, as the oracle does | Reproduces F8 exactly: tests become machine-dependent and invert on any developer box with real credentials provisioned. **Rejected.** |
| (b) **Seam + real adapter + fake** | Narrow read-only `Source` interface; `envSource` and `keychainSource` (go-keyring) implementations; `fakeSource` for tests | Precedence becomes testable in-process and hermetically. The interface is narrowed to **lookup-only**, so the library's `Set`/`Delete`/`DeleteAll` write surface is unreachable from our code by construction. Real adapter ships now, so the chain has no permanently-dead tier. **Chosen.** |
| (c) Seam + env only; defer keychain | Ship the interface with a null keychain that always misses | Leaves tier 1 and 3 permanently dead while `README`/`config.toml.example` document them as working — a false operator contract, and precisely the "success-shaped" ambiguity to avoid. Also defers the only genuinely risky integration past its natural slice. **Rejected.** |
| (d) `99designs/keyring` or `keybase/go-keychain` | Alternative libraries | `keybase/go-keychain` requires cgo on darwin, violating the CGO_ENABLED=0 cross-compile matrix outright. `99designs/keyring` pulls a substantially larger dependency tree (file/pass/vault backends) for surface we do not use. **Rejected.** |

### Decision P3-D3 — Bespoke platform adapters in this slice?

| Option | Description | Assessment |
|---|---|---|
| (a) **No per-platform code in our tree** | One `keychainSource` adapter; let go-keyring dispatch per platform | go-keyring already implements all four target platforms plus a distinct unsupported-platform error (F13). Writing our own `_windows.go`/`_darwin.go`/`_linux.go` would duplicate an audited dependency and triple the untestable surface. **Chosen.** |
| (b) Own platform adapters | Hand-roll wincred / Security.framework / D-Bus | Large, high-risk, platform-specific, and largely untestable in CI. No stated requirement. **Rejected (YAGNI).** |

### Decision P3-D4 — Behavior when the keychain is unavailable

| Option | Description | Assessment |
|---|---|---|
| (a) Oracle-faithful collapse | Swallow every keychain error identically, fall through silently | Preserves the F6 defect: a locked keychain is indistinguishable from an unset credential, and operators get no signal. **Rejected.** |
| (b) Hard-fail on unavailable | Abort startup when no keychain backend exists | Breaks headless CI and containers outright, where env-only operation is the correct and intended path. **Rejected.** |
| (c) **Classify, continue, attribute** | Distinguish `Miss` (absent/empty) from `Unavailable` (no backend / unsupported platform / backend error); both continue to the next tier, but `Unavailable` is recorded in the attribution trail, logged once at WARN, and named in the terminal not-found error | Resolution *outcomes* stay byte-for-byte oracle-equivalent — an unavailable keychain never yields a value and never blocks env fallback — while the diagnostic gap closes. Headless CI stays green. **Chosen.** |

`Unavailable` must never produce a value or a synthesized default; it only ever annotates. That is
the concrete meaning of "no success-shaped fallback" here.

### Decision P3-D5 — Redaction API for the `Secret` type

| Option | Description | Assessment |
|---|---|---|
| (a) `String`/`GoString`/`MarshalJSON`/`LogValue` only | The four surfaces named in the request | Covers `%s`, `%v`, `%q`, `%#v`, `encoding/json`, and `slog`. But `fmt` only consults `Stringer` for verbs `v s q x X`; a stray `%d` on a `Secret` renders `{%!d(string=xoxb-…)}` — **a real leak**. **Rejected as insufficient.** |
| (b) **Total `fmt.Formatter` + the four surfaces** | Implement `Format(fmt.State, rune)` so *every* verb renders `[REDACTED]`, plus `String`, `GoString`, `MarshalJSON`, `MarshalText`, `LogValue` | Closes the `%d`/`%x` hole by construction rather than by reviewer vigilance. Value reachable only through an explicit, greppable `Expose()`. **Chosen.** |
| (c) Wrap in an opaque struct with no accessors | Never expose the value | Unusable — P8 must hand the token to the Slack client. **Rejected.** |

Supporting sub-decisions:

* **Uniform rendering.** A zero/empty `Secret` renders `[REDACTED]` exactly like a populated one.
  Rendering `""` for empty would leak one bit of state on every log line. Emptiness is available
  programmatically via `IsEmpty()`.
* **No `UnmarshalJSON`/`UnmarshalText`.** Deliberately absent, so a `Secret` cannot be populated by
  decoding a config file. This makes the P2 secret-free boundary structural rather than advisory.
* **No `zeroize` equivalent.** Go strings are immutable and garbage-collected; a best-effort wipe
  would be security theatre. Recorded explicitly so a reviewer does not read the omission as an
  oversight.

### Decision P3-D6 — Which fields are `Secret`

| Field | Type | Rationale |
|---|---|---|
| `AppToken` | `Secret` | Slack app-level token; redacted by the oracle too. |
| `BotToken` | `Secret` | Slack bot token; redacted by the oracle too. |
| `TeamID` | `string` | A workspace identifier (`T0123…`), not a credential. The oracle deliberately leaves it visible (F10). Resolved via the same chain for parity, but not secret-typed. |
| `AuthorizedUserIDs` | `[]string` | An ACL of Slack user IDs — sensitive, not secret. Not `Secret`-typed, but redacted **to a count** in every rendered form, closing the oracle's `GlobalConfig` ACL leak (F10). |

### Decision P3-D7 — Where resolved credentials live, and how they compose

| Option | Description | Assessment |
|---|---|---|
| (a) Add fields to `config.Config` | Mirror the oracle's mutate-in-place `load_credentials` | Directly violates the binding P2 R7 boundary and inverts the dependency: credential resolution already consumes config. **Rejected.** |
| (b) New `internal/bootstrap` package | A composition package holding config + credentials | Anticipated by the stash, but there is exactly one consumer today and no startup pipeline (F12). A package whose only job is to hold two values for a single call site is premature abstraction. **Rejected for P3.** |
| (c) **Standalone `internal/credentials`, composed by the caller** | `credentials.Resolve` returns a `*Resolved` value; callers hold `*config.Config` and `*credentials.Resolved` side by side | Preserves the R7 boundary structurally, keeps the dependency direction correct, and adds no speculative package. Revisit a bootstrap package at P8, when a second consumer actually exists. **Chosen.** |

### Decision P3-D8 — How far to take "startup validation integration"

| Option | Description | Assessment |
|---|---|---|
| (a) Wire into `cmd/intercom` `RunE` | Load config, validate, resolve credentials on every invocation | The server does not exist; `RunE` returns `errNotImplemented`. This would make `intercom` fail at startup for every developer without Slack credentials, and it breaks `TestRootCmd_LogLevelDebugEmitsJSON`, which parses stderr as a single JSON object (F12). Inventing a startup pipeline to host one call is scope creep. **Rejected.** |
| (b) **Specify, order, and test the composition without wiring** | An executable integration test in `tests/integration/` composing `config.Load → Validate → credentials.Resolve` in oracle order with a fake keychain and `t.Setenv`, plus a package doc contract | `internal/config` already landed without being wired into `RunE` — the precedent is established. The ordering contract is captured executably rather than prosaically, and the wiring lands with the slice that introduces the real startup pipeline. **Chosen.** |
| (c) Nothing | Ship the library alone | Leaves the ordering contract unrecorded and unverified, which is the part most likely to be got wrong later. **Rejected.** |

### Decision P3-D9 — Test strategy

Chosen: **fake-provider-first, table-driven, hermetic.**

* Precedence is proven with a `fakeSource` that can be programmed to return hit / miss /
  unavailable per (namespace, account) — covering the four-tier matrix the oracle never tested
  (F8), including keychain-beats-env, which is *only* checkable with a seam.
* Environment isolation uses `t.Setenv`, which restores automatically and forbids `t.Parallel()` —
  strictly better than the oracle's `serial_test` + manual, non-RAII cleanup that leaks env vars on
  a panicking assertion (F9).
* The real `keychainSource` gets a thin contract test asserting error **classification** only
  (`ErrNotFound` → `Miss`, `ErrUnsupportedPlatform` → `Unavailable`), never a live keychain
  round-trip. No test performs keychain I/O.
* All fixtures are synthetic canaries (`xoxb-CANARY-...`). `.env.local` is never read.

## Trade-off Comparison

| Criterion | Chain (D1a) | Seam+adapter (D2b) | Classify (D4c) | Total Format (D5b) | Composed, unwired (D7c/D8b) |
|---|---|---|---|---|---|
| Oracle behavioral parity | Exact (ACP) | Exact | Exact outcomes | N/A (improvement) | Exact ordering |
| Implementation complexity | Low (constant) | Low–medium | Low | Low | Low |
| Blast radius on existing code | None | None | None | None | None |
| Testability | High | **Enables precedence tests at all** | High | High | High |
| Security posture | Neutral | Removes write surface | Improves diagnostics | Closes `%d` leak | Preserves R7 structurally |
| Rollback | Single revert | Single revert + dep | Single revert | Single revert | Single revert |
| YAGNI risk | None | None | None | None | Avoids premature bootstrap pkg |

## Decision

Adopt **P3-D1(a), P3-D2(b), P3-D3(a), P3-D4(c), P3-D5(b), P3-D6, P3-D7(c), P3-D8(b), P3-D9**.

Build a new `internal/credentials` package that resolves Slack credentials through a fixed
four-tier, keychain-first, ACP-namespaced chain, behind a narrow read-only provider seam with an
environment source, a `go-keyring`-backed keychain source, and a deterministic fake. Credentials
are returned as a standalone `*Resolved` value composed with `*config.Config` by the caller, never
merged into it. Secret values are carried in a `Secret` type that renders `[REDACTED]` through
**every** `fmt` verb, `encoding/json`, and `log/slog`, and yields its value only via an explicit
`Expose()`. Keychain failures are classified into `Miss` and `Unavailable`, both of which continue
the chain but only the latter of which is surfaced diagnostically. Every test is hermetic and
fake-driven; no test touches a real keychain, and no fixture contains a real secret.

**Rationale.** The three decisions that carry the slice are the seam, the total `Format`, and the
error classification. The seam is what makes precedence testable at all — the oracle's own
precedence is untested and its tests invert on a provisioned machine (F8), which is a defect this
port should not inherit. The total `Format` closes a leak that the four named surfaces alone do
not (`%d` on a `Secret` would print the token), converting redaction from a review rule into a
type guarantee. The classification fixes the oracle's collapsed error classes (F6) at zero
behavioral cost, using a distinction the chosen dependency already provides for free (F13).
Everything else follows from the binding P2 R7 boundary and from declining to build surface —
bootstrap package, platform adapters, CLI subcommands, configurable namespaces — that no consumer
yet requires.

## Rejected Alternatives

* **Two-tier shared-only chain (D1b)** — silently strands ACP-namespaced operators.
* **Configurable namespaces (D1c)** — resurrects retired mode-switching under a new name.
* **Inline library call, no seam (D2a)** — reproduces the oracle's machine-dependent tests.
* **Deferring the keychain source (D2c)** — ships a documented-but-dead tier; a false operator
  contract.
* **`keybase/go-keychain` (D2d)** — requires cgo on darwin; violates the CGO-free matrix.
* **`99designs/keyring` (D2d)** — large unused backend surface for no gain.
* **Own platform adapters (D3b)** — duplicates an audited dependency with untestable code.
* **Oracle-faithful error collapse (D4a)** — preserves a known diagnostic defect.
* **Hard-fail when keychain is unavailable (D4b)** — breaks headless CI and containers.
* **Four redaction surfaces only (D5a)** — leaves a `%d`/`%x` leak.
* **Secrets on `config.Config` (D7a)** — violates the binding R7 boundary and inverts dependencies.
* **`internal/bootstrap` package now (D7b)** — premature; one consumer, no pipeline.
* **Wiring credential resolution into `RunE` (D8a)** — invents a startup pipeline and breaks an
  existing single-log-line test contract.
* **A `Secret` write path / keychain provisioning** — the oracle has none (`set_password` appears
  nowhere); the narrowed interface makes it unreachable.

## Unresolved Questions

1. **Is the oracle's keychain path live in production?** (F7) `keyring = "3"` with no features may
   select the mock keystore, making the oracle env-only in practice. Does not block P3 — the
   documented contract is what we port — but it means Go is likely the first *working*
   implementation, so its keychain tier must be validated on its own merits. Carry to the
   oracle-drift gate (`8ACF7110`), out of scope here.
2. **Keep `SLACK_TEAM_ID_ACP`?** (F4) It is mechanically reachable but undocumented and untested in
   the oracle. P3 keeps it for chain uniformity (every credential gets the same four tiers) and
   documents it — the first time it is written down anywhere.
3. **Cross-tier mixing.** Per-field resolution lets `app_token` come from the ACP keychain while
   `bot_token` comes from shared env, silently pairing two half-configured Slack apps. P3
   **preserves** the oracle behavior (changing it is a behavioral divergence needing its own
   deliberation) but makes it **visible**: per-field source attribution is recorded and logged, so
   a mixed resolution is diagnosable rather than invisible. A consistency warning is a candidate
   follow-up, deliberately not taken here.
4. **`ipc_auth_token` redaction** (F11) — a real unredacted secret, but P13 scope.
5. **Env-clear allowlist for spawned agents** (`src/acp/spawner.rs:109-118`) — credential-adjacent
   hardening belonging to P6.

## Risks and Mitigations

| Risk | Severity | Mitigation |
|---|---|---|
| New dependency breaks the CGO-free cross-compile matrix | High | **Already measured**: all four targets build at `CGO_ENABLED=0` (F13). A CI cross-compile job re-proves it every run. |
| A secret leaks through an unconsidered surface | High | Total `fmt.Formatter` + `MarshalJSON` + `MarshalText` + `LogValue`; a negative-control test sweeps a verb matrix (`%v %s %q %#v %+v %d %x %X`), both slog handlers, and `encoding/json`, asserting a canary never appears. |
| A secret reaches an error string | High | Errors carry only source descriptors (service/account/env-var names). A negative-control test asserts the canary is absent from every constructed error, including the terminal not-found message. |
| Tests accidentally touch a real keychain | Medium | Resolver depends only on the `Source` interface; the real adapter is constructed solely at the composition root. A test asserts the default test resolver holds no keychain source. |
| P2 secret-free boundary erodes | Medium | `Secret` has no unmarshal methods; a test asserts no `config.Config` field is a `Secret` or a token-named string. |
| Headless CI has no keychain backend | Medium | `Unavailable` is a first-class classified outcome that continues the chain; env-only operation is a supported, tested path. |
| Chain choice (ACP) proves wrong | Low | Four-tier ACP is a strict superset of the two-tier shared chain — shared credentials still resolve, at tiers 3–4. |
| `godbus`/`wincred` transitive vulnerabilities | Low | `govulncheck` v1.7.0 clean today and gated in CI on every run. |
| Scope creep into P4+ | Low | Explicit out-of-scope list above; no models, persistence, Slack client, ACP, or CLI subcommands. |

## Scope Boundary

**In:** `internal/credentials` (Secret, typed errors, source seam, env source, keychain source,
fake source, four-tier resolver, authorized-users resolver, `Resolved` aggregate, source
attribution), the `go-keyring` dependency, negative-control leak tests, a composition integration
test, and documentation updates to `README.md` and `config.toml.example`.

**Out:** P4+ phases; `8ACF7110` oracle-drift gate; CI-hardening follow-ups; P1/P2 review
follow-ups; `A92E3FA0`; `EF9352FB`; keychain writes/provisioning; `ipc_auth_token`; ACP spawner
env allowlist; new CLI subcommands; any change to `internal/config` behavior; any modification to
either reference repository.

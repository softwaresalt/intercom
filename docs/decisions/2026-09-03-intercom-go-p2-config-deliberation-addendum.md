---
title: "P2 Config — Deliberation Addendum: Oracle Re-verification and Decode/Validation Contract"
description: "Resolves the P2 open questions left by the continuation deliberation and pins the exact decode, defaulting, validation, and path-integration contract for internal/config"
date: 2026-09-03
status: accepted
phase: P2
parent_deliberation: "docs/decisions/2026-09-03-intercom-go-port-continuation-deliberation.md"
source_brief: "docs/decisions/2026-07-06-go-port-reference-brief.md"
behavioral_oracle:
  repo: "softwaresalt/agent-intercom"
  commit: "41df772"
  access: "read-only, out-of-tree"
  reverified: "2026-09-03 (HEAD and remote URL re-checked this session)"
stash_entries: ["037B1552"]
predecessor: "002-F / 002-S (P1, shipped, merge 9efb156)"
successor: "P3 credential resolution (remains in stash 037B1552)"
tags: ["go", "port", "config", "toml", "validation", "deliberation"]
---

# P2 Config — Deliberation Addendum: Oracle Re-verification and Decode/Validation Contract

## Problem Frame

The continuation deliberation selected P1 (`internal/apperr` + `internal/pathsafe`) as the
bounded slice and deferred P2 (config) and P3 (credentials) to stash `037B1552`. P1 has now
shipped (`002-S`, merge `9efb156`; closure `docs/closure/002-S-002-F-post-merge-closure.md`,
`closure_status: READY`). P2 is the next dependency-ready slice.

This addendum exists because a focused oracle re-read for P2 **contradicted three quantitative
claims** carried in stash `037B1552` and in the parent deliberation's Risks table, and because
the parent deliberation explicitly deferred one question ("Unresolved Questions" item 3) to P2.
Planning P2 against the uncorrected numbers would have frozen wrong contracts into ~53 ported
tests.

Scope of this addendum is **P2 only**: schema, defaults, decode, validation, workspace routing,
and path integration. P3 credential resolution is deliberately untouched and remains the next
successor under `037B1552`.

### Constraints

- The oracle is authoritative on conflict (brief §1.3), but documented-intent/implementation
  mismatches are called out rather than silently cloned.
- CGO-free, pure-Go, cross-platform (the 4-target matrix in `scripts/targets.json`).
- P1 primitives (`internal/apperr`, `internal/pathsafe`) are reused, not re-implemented.
- MCP/ACP is already resolved by parent decision D1: no `mode`, no `--transport`, no
  `McpDriver`, no stdio. Config carries no mode field.

### Success Criteria

Every field, default, decode rule, and validation rule in P2 is traceable to an oracle
`file:line`, or is an explicitly recorded, justified divergence.

## Oracle Re-verification

`git -C C:\Source\GitHub\intercom log --oneline -1` → `41df772 feat(config): enforce unique
channel_id in ACP mode validation (#36)`. Remote → `https://github.com/softwaresalt/agent-intercom.git`.
Pin holds; the repository was read-only throughout and is uncommitted/unmodified.

Config surface: `src/config.rs` (941 lines) and `src/config_watcher.rs` (229 lines). There is no
`src/config/` directory. Discovery lives in `src/main.rs`; the channel-uniqueness enforcement
call lives in `src/slack/commands.rs`.

## Research Findings

### F1 — [Correction] There are 17 non-zero defaults, not 16

Stash `037B1552` finding (e) and the parent deliberation's Risks table both say "16 defaults".
The oracle defines **17** named default functions (`config.rs:167-341`), all producing non-zero
or non-empty values. The seventeenth is `default_slack_detail_level() -> SlackDetailLevel::Standard`,
which is easy to miss because it is an enum rather than a scalar literal.

`TimeoutConfig::wait_seconds` is the only field with a *zero* default (bare `#[serde(default)]`,
`0` meaning "no timeout") and is therefore excluded from the 17.

This matters directly: the count is used as a test-criteria checklist. Porting 16 would have
left `slack_detail_level` silently defaulting to Go's zero value.

### F2 — [Correction] "9 validation rules" is a total across three functions; only 5 run on the normal load path

The oracle splits validation three ways:

| Function | Rules | Called from | Failure mode |
|---|---|---|---|
| `validate()` `config.rs:539` | 5 | `from_toml_str()` — every load | hard error |
| `validate_for_acp_mode()` `config.rs:661` | 3 | `main.rs:136`, `commands.rs:703` | hard error |
| `validate_host_cli_path()` `config.rs:695` | 1 | `main.rs:139` via `.ok()` | **discarded** |

The stash's "9 validation rules" is arithmetically right but operationally misleading: a plain
`GlobalConfig::from_toml_str()` enforces only 5. The other 4 were gated on ACP mode — a concept
this port has retired.

`validate_host_cli_path()`'s single error branch (absolute `host_cli` that does not exist) is a
strict **subset** of `validate_for_acp_mode()` rule 5; its unique contribution is warning-only
(non-standard install location, relative name not on `PATH`). Since `main.rs` discards its
`Result` entirely, it has no error semantics in practice.

All validation is **fail-fast** — first error returns; nothing accumulates. Error type is
`AppError::Config(String)` throughout.

### F3 — [Confirmed] No `deny_unknown_fields` anywhere

Grep across the whole `src` tree finds the token only inside an explanatory *comment* at
`config_watcher.rs:66-67`. Unknown TOML keys are silently ignored on both the full-config path
and the hot-reload path. Stash finding (b) is confirmed; Go must not use strict decoding.

### F4 — [Confirmed, with mechanism correction] Defaulting is per-field serde, not a pre-populated struct

Stash finding (h) prescribes "a pre-populated `Default()` struct". The oracle actually achieves
the same observable result via per-field `#[serde(default = "...")]`; there is no
`..Default::default()` spread, and `GlobalConfig` has **no** `impl Default` at all — it can only
be constructed by deserialization.

The prescription is still the correct *Go* strategy (see P2-D2), but it is an adaptation, not a
port. Recording this prevents a future reader from "correcting" the Go code toward a
non-existent oracle `Default` impl.

### F5 — [New] Three config sections are structurally required

`slack`, `timeouts`, and `stall` carry **no** `#[serde(default)]` on the `GlobalConfig` field
(`config.rs:349, 386, 389`), so serde errors if the section is absent — even though nearly every
field *inside* them has a default. Together with the two required scalars
(`default_workspace_root`, `host_cli`) that is five required TOML entries.

This is a decode-behavior subtlety with no Go analogue: decoding into a pre-populated struct
makes everything optional by construction. Addressed by P2-D4.

### F6 — [Refines parent Q3] `channel_id` uniqueness *is* enforced — just not where the docs say

The parent deliberation recorded this as a plain oracle bug ("the reload path does not call it").
The re-read shows a more precise picture:

- `validate_unique_channel_ids()` is defined at `config.rs:86-100` (`pub(crate)`).
- It is called from `validate_for_acp_mode()` at `config.rs:676`.
- It is **also** called at `slack/commands.rs:731`, inside `handle_acp_session_start()`,
  operating on the **hot-reloaded live snapshot** (`state.workspace_mappings.read()`).
- **Confirmed:** `config_watcher.rs` calls no validation whatsoever. `parse_workspace_mappings()`
  (`config_watcher.rs:59-73`) and the notify callback (`config_watcher.rs:137-165`) simply
  re-parse `[[workspace]]` and atomically swap an `Arc<RwLock<Vec<WorkspaceMapping>>>`.

So uniqueness is not unenforced on reloaded data — it is enforced **late**, at ACP session start,
rather than **eagerly**, at reload. The doc-comment at `config.rs:77-83` is accurate about *that*
the check applies, and misleading about *when*. This reframes the question and is resolved by
P2-D6.

Note also that the normal load path checks `workspace_id` uniqueness but **not** `channel_id`
uniqueness — the latter was ACP-gated.

### F7 — [Confirmed] `protocol_mode` is a database column, not a config key

`models/session.rs:90` (`protocol_mode: ProtocolMode`), `persistence/schema.rs:48-49`
(`ALTER TABLE session ADD COLUMN protocol_mode TEXT NOT NULL DEFAULT 'mcp'`), and
`session_repo.rs:162-166` (`parse_protocol_mode` accepting `"mcp"` | `"acp"`). The server-wide
selector was the CLI flag `--mode`, retired by parent decision D1.

P2 therefore defines **no** `protocol_mode` field and **no** `mode` field. The read-tolerant /
write-`acp` behavior the operator specified is a P4/P5 persistence concern, not a config concern.

### F8 — [Confirmed] `host_cli` is required, and redaction is narrower than assumed

`host_cli: String` at `config.rs:379` — required, no default, not `Option`. Under mode retirement
it becomes unconditionally required, strengthening stash finding P1-g exactly as predicted.

There is **no `Secret<T>` type in the oracle**. `SlackConfig` has a hand-written `Debug` impl
(`config.rs:129-152`) that redacts `app_token` and `bot_token`, and leaves `team_id` **and**
`channel_id` visible. Stash finding (f) is confirmed and extends to `channel_id`. The redacting
`Secret` type described for P3 is a Go *improvement*, not a port.

Critically for P2: `app_token`, `bot_token`, `team_id` and `authorized_user_ids` are all
`#[serde(skip)]` — **they are never read from `config.toml` at all**. P2's config file surface
contains no secret-bearing keys, which makes the "no secrets in fixtures" requirement structurally
satisfiable rather than merely a review rule.

### F9 — [New] Library verification: `BurntSushi/toml` satisfies the decode contract

The decode strategy is load-bearing, so the candidate library was verified against source rather
than assumed (`BurntSushi/toml` v1.6.0, latest as of this session):

- `func (md *MetaData) IsDefined(key ...string) bool` — exists (`meta.go`).
- `func (md *MetaData) Undecoded() []Key` — exists (`meta.go`).
- `unifyStruct` (`decode.go:298`) iterates `for key, datum := range tmap`, i.e. **only keys
  present in the TOML document**. There is no `reflect.Zero` or `SetZero` anywhere in `decode.go`.

That last point is the decisive one. It structurally guarantees both required properties at once:
absent keys leave a pre-populated destination untouched (the 17 defaults survive), and an
explicitly written `false` or `0` *is* present in `tmap` and therefore *does* overwrite a
non-zero default. The "naive decode clobbers explicit false/zero" risk from the parent
deliberation is eliminated by construction, not by discipline.

One divergence surfaced: `unifyStruct` falls back to `strings.EqualFold` for key matching, so Go
accepts `Host_CLI` where Rust's `rename_all = "snake_case"` serde is case-sensitive. See P2-D8.

`pelletier/go-toml/v2` (v2.4.3) also decodes into an existing struct, but exposes no
`Undecoded()` equivalent, which P2-D3 needs.

## Options Evaluated

### Decision P2-D5 — How to handle validation now that mode is retired

| Option | Description | Assessment |
|---|---|---|
| (a) Faithful split | Keep `Validate()` and `ValidateForACP()` as separate exported methods | Preserves oracle shape, but the second is now unconditional — two methods that must always both be called is a trap. A caller that forgets one silently loses `host_cli` validation. **Rejected.** |
| (b) Unified unconditional `Validate()` | Merge all 8 hard rules into one deterministic sequence; expose advisory warnings separately | Matches the retired-mode reality: there is exactly one mode, so a mode-gated rule is just a rule. Single call site, impossible to under-validate. **Chosen.** |
| (c) Unified, host_cli existence downgraded to warning | As (b) but never hard-error on a missing `host_cli` binary | Diverges from oracle rule 5, which *does* hard-error. Would let a misconfigured host CLI fail late, at spawn time, in a subprocess. **Rejected.** |

### Decision P2-D4 — Required-key semantics

| Option | Description | Assessment |
|---|---|---|
| (a) Everything optional | Accept the natural consequence of pre-populated defaulting | Silently accepts a `config.toml` with no `default_workspace_root` and no `host_cli`, then fails deep in validation with a confusing message — or worse, canonicalizes `""`. **Rejected.** |
| (b) Full parity via `IsDefined` | Require all five oracle-required entries | Restores exact oracle behavior, but requiring the `[timeouts]` and `[stall]` *sections* is an artifact of a missing serde attribute, not a designed contract — every field inside them has a default, so an empty `[stall]` and an absent `[stall]` are semantically identical. Propagating that accident makes the Go config strictly more annoying with no safety gain. **Rejected.** |
| (c) Parity for required scalars, defaulted sections | Hard-require `default_workspace_root` and `host_cli` via `MetaData.IsDefined`; let `[slack]`/`[timeouts]`/`[stall]` default when absent, and record the divergence | Keeps the two checks that carry real meaning (a workspace root cannot be guessed; `host_cli` is now unconditionally required per F8) and drops the two that carry none. Strictly more permissive than the oracle, so no previously valid config becomes invalid. **Chosen.** |

### Decision P2-D6 — Where `channel_id` uniqueness is enforced (resolves parent Q3)

| Option | Description | Assessment |
|---|---|---|
| (a) Clone oracle exactly | Enforce only at "session start" | P2 has no session-start path — that is P6/P9. Choosing this defers the check to a phase that does not exist yet, leaving P2 with a documented-but-absent guarantee. **Rejected.** |
| (b) Implement the *documented* behavior | Enforce eagerly in `Validate()` at load; export a reusable mapping validator for the future watcher; record the divergence | The parent deliberation's proposed resolution, and F6 strengthens it: the oracle's own doc-comment asserts this behavior, and the late check at `commands.rs:731` proves the invariant is genuinely required, not optional. Eager enforcement is strictly safer — it converts a per-session runtime failure into a startup failure. **Chosen.** |
| (c) Both eager and late | Enforce at load *and* re-check at session start | The late re-check is correct but belongs to the phase that owns session start. Adding it now would require inventing a call site in P2. Deferred to P6/P9 as a note, not built here. |

## Decision

### P2-D1 — TOML library: `github.com/BurntSushi/toml` v1.6.0

Pure Go, CGO-free, single dependency, no transitive additions. Selected specifically for
`MetaData.IsDefined` (P2-D4) and `MetaData.Undecoded` (P2-D3), which the alternative does not
provide. This is the module's second direct dependency after `spf13/cobra`.

### P2-D2 — Decode into a pre-populated `Default()` struct

`Default()` returns a fully populated `Config` carrying all 17 non-zero defaults. `Load` decodes
the TOML document *over* that value. Per F9 this preserves defaults for absent keys and honours
explicit `false`/`0` for present keys. `Default()` is exported so tests and future phases can
obtain the baseline without touching the filesystem — closing the gap that the oracle's missing
`impl Default` leaves.

### P2-D3 — Tolerant decode, observable unknown keys

No strict decoding (F3, oracle parity). Unknown keys never fail the load. But `Load` calls
`MetaData.Undecoded()` and returns the unknown key paths to the caller so they can be logged as
a warning. Tolerant *and* observable: the oracle's permissiveness is preserved, while a typo'd
`max_concurent_sessions` stops being invisible. This is a pure addition — no config that loads in
Rust fails in Go.

### P2-D4 — Required scalars enforced, required sections relaxed

Per option (c) above. `default_workspace_root` and `host_cli` are hard-required via
`MetaData.IsDefined`. Absent `[slack]`, `[timeouts]`, `[stall]` sections take their defaults.
Recorded divergence: **Go accepts three configs the oracle would reject.** Strictly more
permissive; no valid config is broken.

### P2-D5 — One unconditional `Validate()`, 8 hard rules, deterministic order

Per option (b) above. Fail-fast on first error (oracle parity), `apperr.KindConfig` for every
error, and a fixed rule order so the *same* invalid config always yields the *same* error —
required for deterministic tests and for dark-factory reproducibility. Advisory `host_cli`
warnings (non-standard location, not on `PATH`) are returned separately and never fail the load,
matching `main.rs:139`'s discarded `Result`.

### P2-D6 — Eager `channel_id` uniqueness at load

Per option (b) above. Resolves parent deliberation Unresolved Question 3. Divergence recorded:
the oracle enforces this at ACP session start (`commands.rs:731`), not at load or reload; Go
enforces it at load, implementing the behavior the oracle's own doc-comment (`config.rs:77-83`)
and brief §7 both describe.

**Amended during plan review (attempt 1 → 2).** This decision originally also exported the mapping
validator "so the P-later config watcher can reuse it". Plan review rejected that on two
independent grounds: the Scope auditor flagged exporting API surface for a consumer that does not
exist as YAGNI, and the Architecture and Coupling reviewers independently showed that a *single*
validator covering both the field rules and the uniqueness rule **cannot** be called at one point
in the sequence without violating the fixed rule order, because the `host_cli` rules sit between
them. The mapping rules are therefore implemented as two **unexported** stages invoked at their
respective contract positions. Exporting a composite for the future watcher is a one-line change
in the phase that actually needs it. See the plan's Validation Contract and finding P1-b.

### P2-D7 — `default_workspace_root` canonicalized through `internal/pathsafe`

The oracle canonicalizes and UNC-strips `default_workspace_root` during `validate()`
(`config.rs:543-547`), mutating the config in place, and requires the directory to exist.
`pathsafe.NewRoot` — shipped in P1 — performs exactly `filepath.Abs` → `filepath.EvalSymlinks` →
`stripUNCPrefix`, returns `apperr.KindPathViolation` on failure, and already fails on a
non-existent directory because `EvalSymlinks` does. This is a direct reuse of the P1 primitive
and the concrete dependency edge from `002-F` to P2.

Consequence accepted: config validation touches the filesystem, so fixtures need real
directories. Tests use `t.TempDir()`, which is deterministic and cross-platform. `database.path`
and `workspace.path` are **not** canonicalized (oracle parity — `config.rs` does neither).

### P2-D8 — Key-matching case sensitivity: collisions rejected

`BurntSushi/toml`'s `EqualFold` fallback (F9) makes Go's key matching case-insensitive where the
oracle's is case-sensitive.

**Amended during plan review (attempt 1 → 2).** This was originally *accepted* as a harmless
permissive divergence. The Security reviewer showed it is not harmless when two case variants are
present **simultaneously**: `host_cli` and `Host_CLI` both target the same Go field, and because
the decoder iterates a Go map, which value wins is **nondeterministic**. A config could therefore
shadow `host_cli` — the binary a later phase executes — or `default_workspace_root` — the
containment root — with a value a reviewer reading the file would not expect to take effect.

Revised decision: a **case-fold collision between two distinct key paths is rejected** at load
with `apperr.KindConfig`. A lone non-canonical spelling still decodes (inherent to the decoder,
harmless without a collision) and is locked by a test so the boundary between the two cases cannot
drift. This is recorded as divergence V2/V2b in the plan and is stricter than the oracle.

### P2-D9 — No `mode`, no `protocol_mode`, no transport keys in config

Confirmed by F7 with source evidence. Restated here so the absence is visibly *decided* rather
than merely forgotten.

## Corrections to Propagate

These supersede the corresponding claims in stash `037B1552` and the parent deliberation:

| Claim | Status | Correct value |
|---|---|---|
| "16 non-zero defaults" | **Wrong** | 17 (F1) |
| "9 validation rules" | **Misleading** | 9 total across 3 functions; 5 on the normal load path; 8 hard rules after unification (F2, P2-D5) |
| "config_watcher does not call the uniqueness check → oracle bug" | **Refined** | Correct that the watcher does not; but the check *is* enforced at ACP session start (F6) |
| "pre-populated `Default()` struct per oracle" | **Refined** | Correct strategy for Go, but an adaptation — the oracle uses per-field serde defaults and has no `impl Default` (F4) |
| "SlackConfig redacts app_token/bot_token but not team_id" | **Confirmed, extended** | `channel_id` is also unredacted; and no `Secret` type exists in the oracle (F8) |
| "no `deny_unknown_fields`" | **Confirmed** | (F3) |

## Rejected Alternatives

- **Porting `config_watcher.rs` hot-reload in P2.** Deferred. It needs a file-watch dependency
  and a concurrency model (`Arc<RwLock<...>>` → Go equivalent), which would roughly double the
  slice and pull in concurrency review. P2 ships the reusable mapping validator that the future
  watcher will need; the watcher itself is a later phase.
- **Porting credential loading alongside config.** That is P3, explicitly preserved as the next
  successor. Merging them would re-create the oversized original slice that was already split
  once.
- **Vendoring oracle config fixtures.** Attractive for oracle-independence (and noted in stash
  `8ACF7110`), but it is a separate artifact class with its own CI surface. Out of scope here.

## Unresolved Questions

1. **[Deferred to P6/P9]** Should `channel_id` uniqueness *also* be re-checked at session start,
   mirroring `commands.rs:731`, once a session-start path exists? P2-D6 makes this belt-and-braces
   rather than load-bearing.
2. **[Deferred to P-later]** When the config watcher lands, does it validate the whole config on
   reload or only the mappings? P2 exports the mapping validator either way.
3. **[Non-blocking, P3]** The `Secret` type is a Go improvement with no oracle counterpart, so its
   exact API is unconstrained by the oracle. Decide in P3.

## Risks and Mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| Wrong default count freezes into ~53 ported tests | High | Corrected to 17 (F1); every default enumerated with its oracle line in the plan's default table. |
| Decode clobbers explicit `false`/`0` | High | Eliminated structurally by F9 (`unifyStruct` iterates present keys only, no zeroing); locked by dedicated tests. |
| Strict decoding breaks real configs | High | P2-D3 mandates tolerant decode; a test asserts an unknown key loads successfully. |
| Silent typo in a config key | Medium | P2-D3 surfaces `Undecoded()` keys to the caller for warning-level logging. |
| Validation order non-deterministic → flaky tests | Medium | P2-D5 fixes an explicit rule order; a test asserts which error a multi-fault config yields. |
| Filesystem dependency makes config tests flaky | Medium | `t.TempDir()` only; no reliance on ambient paths, `HOME`, or CWD. |
| Secrets in fixtures | High | Structurally impossible for P2: every secret-bearing field is `#[serde(skip)]` in the oracle (F8), so the Go config struct has no secret field to populate. Fixtures are synthetic regardless. |
| Cross-platform path divergence (UNC, separators) | Medium | Canonicalization delegated to the already-shipped, already-tested `pathsafe.NewRoot` (P2-D7). |
| New dependency expands supply-chain surface | Low | One pure-Go module, no transitive deps; `go mod verify`, `govulncheck`, and the `go mod tidy` dirty-check already gate it in CI. |

## Provenance

- Oracle: `softwaresalt/agent-intercom` @ `41df772`, read-only, out-of-tree at
  `C:\Source\GitHub\intercom`. Re-verified this session by HEAD and remote URL. Never modified.
- Parent deliberation: `docs/decisions/2026-09-03-intercom-go-port-continuation-deliberation.md`
  (decisions D1–D3; this addendum resolves its Unresolved Question 3).
- Brief: `docs/decisions/2026-07-06-go-port-reference-brief.md`.
- Predecessor: `002-F` / `002-S` (P1), shipped, merge `9efb156`, closure
  `docs/closure/002-S-002-F-post-merge-closure.md`.
- Stash: `037B1552` (P2 consumed by this addendum and its plan; P3 preserved).
- `references/herdr` is unrelated to this port and was not read.

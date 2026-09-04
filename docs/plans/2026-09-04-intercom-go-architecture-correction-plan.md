---
title: "intercom-go Architecture Correction — Config Remediation, Toolchain Floor, Anti-Regression Gate"
description: "Corrective implementation plan removing the retired Slack/ACP/headless-host architecture from the shipped internal/config surface, raising the Go language floor to the Copilot SDK's requirement, and installing a mechanical gate against regression"
date: 2026-09-04
source: "docs/decisions/2026-09-04-intercom-go-architecture-correction-copilot-sdk-deliberation.md"
governing_design: "docs/design-docs/intercom-go-backend-architecture.md"
phase: "C1 (corrective slice 1)"
status: "planned"
---

# Implementation Plan — Corrective Slice C1

## Problem Frame

The governing architecture correction
(`docs/decisions/2026-09-04-intercom-go-architecture-correction-copilot-sdk-deliberation.md`)
retires Slack, the application-owned ACP broker, and the headless-Copilot-CLI
lifecycle, and adopts `github.com/github/copilot-sdk/go` as the agent
boundary.

The shipped `internal/config` package (merged to `main` at `2a54791` via PR
#11) encodes the retired architecture **as enforced validation**, not merely as
unused fields:

* `validate.go` **rule 4** requires a non-empty `channel_id` on every
  `[[workspace]]` entry, and **rule 8** enforces `channel_id` uniqueness.
* `validate.go` **rule 6** requires `host_cli` to be set.

Consequence, verified by reading the shipped rules: **a `config.toml` written
for the corrected architecture cannot load today.** It is rejected for a
missing `channel_id` and a missing `host_cli`. This is finding **P2-COMPAT-1**
in the governing decision, and it is why a documentation-only correction was
rejected.

Separately, `go.mod` declares `go 1.22` while the Copilot SDK declares
`go 1.24`. This is a hard compile-time blocker on SDK adoption
(**F6**).

### Blast radius (verified, not assumed)

`internal/config` has exactly **one** non-test importer:
`cmd/intercom/main.go`, which consumes **only** `config.DefaultConfigPath` as
a Cobra flag default. Both `cmd` binaries still return a `not implemented`
sentinel from `RunE`. **No `cmd/` file is touched by any unit** —
`DefaultConfigPath` is unchanged. The remediation touches `internal/config`
(5 non-test files + 12 test files = 17, each enumerated in the unit
inventory below), two operator-facing artifacts, one new script, and one CI
step.

## Requirements Trace

| Source requirement | Implementation units |
|---|---|
| D0.7 — Slack completely retired | U-B0, U-B1, U-B2, U-D2, U-D3, U-E1a |
| D0.1/D0.2 — no ACP broker, no headless CLI lifecycle; SDK owns transport | U-B1, U-E1a |
| D1 — SDK pin `v1.0.11` (Go 1.24 floor) | U-A1 |
| D6 / D6a / D6b — retirement enforced mechanically, with scoped limits | U-E1a, U-E1b |
| D7 — persistence deferred, not decided | *(explicit non-goal; `default.go`'s `database` value byte-identical per I8)* |
| Audit §A — schema dispositions | U-B0, U-B1 |
| Audit §B — validation contract 10 → 7 rules | U-B1 (rules), U-B2 (contract test) |
| Audit §C — operator-facing artifacts | U-D1, U-D2, U-D3 |
| Audit §D — P2-SEC-1 (`--dangerously-skip-permissions`) | U-B1, U-D2 |
| Audit §D — P2-SEC-2 (`host_cli = "claude"`) | U-D2 |
| Audit §D — P2-COMPAT-1 (config cannot load) | U-B1 |
| Audit §D — P2-NAME-1 (`agent-rc.db` vs `agent-intercom.db`) | U-D2 (example side only) |
| Design §6.2 — reject relative `cli_path` | U-B1, U-B2 |

## Non-Goals (explicit scope fence)

* **No SDK dependency is added in this slice.** `go mod tidy` prunes modules
  with no importer and CI runs a hard `go mod tidy` dirty-check, so adding
  `copilot-sdk/go` here — where the Non-Goals forbid using it — would be
  mechanically unstable. The dependency lands in phase **C2** with its first
  real import. This narrowing of the governing decision's Option C is recorded
  as an explicit amendment (**D9a**), not left to inference.
* No Bubble Tea, WebSocket, hub, event-envelope, or Dev Tunnel code.
* No persistence decision. `retention_days` and `default.go`'s `database`
  value are untouched (I8). *(D4a's minimal session pointer is a later-phase
  concern and is not implemented here.)*
* No changes to `internal/apperr` or `internal/pathsafe` (I5). In particular:
  * stash `7028FBE7` (rule 10 duplicating pathsafe traversal logic) stays
    deferred — fixing it needs a new `pathsafe` export this slice does not
    authorize;
  * `apperr`'s `KindSlack`/`KindIPC`/`KindACP` taxonomy contamination is
    **acknowledged and tracked** (decision D6a), explicitly excluded from the
    U-E1a gate's scope, and deferred to its own slice.
* No credential resolution (the retired P3 scope).
* No SPA framework selection.
* No `.ps1` counterpart for the gate script.

## Implementation Units

**Unit IDs are namespaced `U-*` to avoid collision with roadmap phase IDs
`C1…C11`.** (Review found the previous `C1/C2/C3` unit IDs ambiguous against
phase `C1/C2/C3` in the same document.)

Execution posture is **test-first** for every code unit unless noted.

### Ground-truth file inventory (verified, drives the decomposition)

The first draft of this plan under-scoped the surface. Verified references to
fields being deleted:

| File | Contaminated by | Owning unit |
|---|---|---|
| `internal/config/config.go` | Slack, ACP, HostCLI, IPCName, ChannelID, SlackDetailLevel | U-B1 |
| `internal/config/workspace.go` | **11 × `ChannelID`** — three exported resolvers | **U-B0** |
| `internal/config/validate.go` | HostCLI, ChannelID (rules 4/6/7/8) | U-B1 |
| `internal/config/default.go` | ACP, IPCName, SlackDetailLevel | U-B1 |
| `internal/config/load.go` | doc comments: "10 rules", rule 6 required, `Slack.MarkdownUploadExtensions` | U-B1 |
| `internal/config/resolve_test.go` | 23 × `ChannelID` — entirely channel routing | **U-B0** |
| `internal/config/workspace_test.go` | 10 × `ChannelID`; `newValidBaseConfig` sets `HostCLI`; rule-3/4/5 tests | U-B2 |
| `internal/config/hostcli_test.go` | 25 matches; rule-6/7 tests; **the multi-violation determinism fixtures** | U-B2 |
| `internal/config/paths_test.go` | ChannelID fixtures + shared helper | U-B2 |
| `internal/config/decode_semantics_test.go` | asserts `cfg.ACP`, `cfg.IPCName`, `cfg.SlackDetailLevel` | U-B2 |
| `internal/config/load_test.go` | `host_cli` fixtures, case-fold collision | U-B2 |
| `internal/config/decode_test.go` | unknown-key assertions | U-B2 |
| `internal/config/detail_test.go` | SlackDetailLevel | U-B2 |
| `internal/config/validate_test.go` | rule assertions | U-B2 |
| `internal/config/default_test.go` | `want 17`; the I3 iota-zero property test | U-D1 |
| `internal/config/docs_test.go` | **hardcoded `!= 17` and `!= 11`**, two count-bearing test names | U-D1 |
| `internal/config/example_test.go` | reflection over every toml tag | U-D2 |
| `config.toml.example` | Slack, ACP, host_cli, ipc_name, channel_id | U-D2 |
| `docs/config-reference.md` | 26 matches | U-D3 |

`cmd/intercom/main.go` consumes only `config.DefaultConfigPath`, which **no
unit changes**. **No `cmd/` file is touched.** `README.md` contains no
Go-version prerequisite line — verified, so A1 touches `go.mod` only.

### Sub-epic A — Toolchain floor

**U-A1 — Raise the Go language floor to 1.24** *(optional early-signal unit)*
*Files:* `go.mod`.
*Change:* `go 1.22` → `go 1.24`. The `toolchain go1.26.5` pin (GO-2025-3750
remediation, invariant I7) is preserved byte-identical.
*Tests:* full suite unchanged; CI `lint`, `security`, 4/4 `cross-compile`.
*Verifiable exit:* `go build ./... && go test ./...` pass; **`go mod tidy`
leaves the tree clean** (CI runs a hard dirty-check); `go vet` clean; 4/4
cross-compile legs.
*Posture:* migration-first. *Rollback:* single-line revert.
*Note:* nothing in this slice imports the SDK, so U-A1 is **deferrable to
phase C2 without coherence loss**. It is retained as cheap early signal (R2),
not because the slice requires it.

### Sub-epic B — Package remediation (ONE atomic Go-code unit set)

> **Sequencing correction (P0).** The first draft split schema edits (B) from
> validation edits (C) and claimed the build was green in between. It is not:
> `validate.go` calls `validateHostCLI(c.HostCLI, …)` and reads `m.ChannelID`,
> and `workspace.go` is built entirely on `ChannelID`. Deleting a field in one
> unit and its consumers in a later unit leaves a **non-compiling tree**.
> Schema and validation edits therefore land **together**.

**U-B0 — Remove the channel-routing API**
*Files:* `internal/config/workspace.go`, `internal/config/resolve_test.go`.
*Change:* delete `ResolveChannelID` and `ResolveWorkspaceByChannelID`
outright — both are channel-keyed lookups that have no meaning without Slack.
Re-found `WorkspaceRootForChannel` as `WorkspaceRoot(workspaceID string)`,
keyed on `workspace_id`, preserving the existing fallback to
`DefaultWorkspaceRoot` and the documented containment precondition.
*Tests:* `resolve_test.go` rewritten against `workspace_id` lookup. The containment precondition is **re-founded** on `WorkspaceRoot` (the shipped comment phrases it "as with ResolveWorkspaceByChannelID", which no longer exists). `TestResolveWorkspaceByChannelIDReturnsValueNotPointer` is **deleted, not re-founded** — `WorkspaceRoot` returns a `string`, so there is no by-value mapping return left to assert. Record explicitly that **no exported by-ID mapping accessor survives**; this is an intentional capability removal with zero consumers today.
*Verifiable exit:* no `ChannelID` reference in `workspace.go` or
`resolve_test.go`; package still compiles in isolation (this unit alone does
not delete the struct field).
*Sequenced FIRST* so U-B1's field deletion has no surviving consumer.

**U-B1 — Remove the retired schema surface and its validation** *(split into
two commits: U-B1a then U-B1b)*

> **Split rationale (attempt-3 scope finding).** Compile-integrity requires a
> deleted field and its consumers to die together — but it does **not** require
> the *additive* `[copilot]` work to ride along. Deleting first and adding
> second leaves a compiling tree at both boundaries and keeps each commit
> inside the 2-hour rule. The atomicity that matters is
> delete-field-with-its-consumers, which U-B1a preserves.

**U-B1a — Delete the retired schema and its validation**
*Files:* `internal/config/config.go`, `internal/config/validate.go`,
`internal/config/default.go`, `internal/config/load.go`.
*Change (one commit — the tree cannot compile otherwise):*

* delete `SlackConfig`, the `Slack` field, and `WorkspaceMapping.ChannelID`;
* delete `ACPConfig`/`ACP` **and all four of its defaults**;
* delete `HostCLI` and `HostCLIArgs`; delete `IPCName` and its default;
* rename `SlackDetailLevel` → `OperatorDetailLevel` and the TOML key to
  `operator_detail_level`, preserving the iota-zero property (I3) and updating
  the `UnmarshalText` error message;
* **validation:** delete rule 4 (`channel_id` presence), rule 8 (`channel_id`
  uniqueness, including `validateChannelUniqueness`), and **eliminate rule 6
  entirely** — not relax it. `validateHostCLI` is removed wholesale here; its
  surviving absolute-path existence semantics are re-established on
  `cli_path` in U-B1b.

*Verifiable exit:* `go build ./...` green; no `Slack`/`ACP`/`HostCLI`/
`IPCName`/`ChannelID` identifier in non-test package code; no stale rule-count,
rule-6, **defaults-count**, channel-routing, or Slack-secret claim in **any**
doc comment. Specifically: `config.go`'s package doc loses "the 17 non-zero
defaults", "workspace-to-channel routing", and the
`app_token`/`bot_token`/`team_id` enumeration; `default.go`'s doc comment loses
"the oracle's 17 non-zero defaults"; `Report`'s comment loses "the `host_cli`
PATH advisory"; `load.go` loses "its own 10 rules", the rule-6-required claim,
and the `Slack.MarkdownUploadExtensions` reference.

**U-B1b — Add `[copilot].cli_path` and the rule-5 contract**
*Files:* `internal/config/config.go`, `internal/config/validate.go`.
*Change:*

* add `CopilotConfig{ CLIPath string }` bound to `cli_path` on a `[copilot]`
  section. **`startup_timeout_seconds` is deliberately NOT added in C1** — it
  would be an unconsumed key (no SDK in this slice, per R1), it would add a
  sixth defaults pin site to the plan's highest-risk coupling, and because an
  explicitly written `0` overrides a default it would ship a
  zero-`Client.Start`-deadline hazard with no rule guarding it. It lands in
  **C2** with its first real consumer and its own validation;
* implement the **single rule-5 `cli_path` contract** exactly per design
  §6.2's predicate:

```text
cli_path == ""                              -> valid, return early, NO PATH advisory
                                               (must NOT fall through to exec.LookPath(""))
filepath.IsAbs(cli_path)                    -> valid iff the file exists, else ERROR
matches ^[A-Za-z]: and not absolute         -> ERROR (Windows drive-relative, e.g. "C:copilot")
contains '/' or '\' (and not absolute)      -> ERROR (relative path)
otherwise (no separator, no drive prefix)   -> bare name: PATH advisory only, never fatal
```

  Both separators are tested regardless of `GOOS`, following the existing
  `containsDotDotSegment` precedent, because `config.toml` is portable text;
* renumber the contract to the **7** rules enumerated by name in the governing
  decision, preserving the fixed-order property (I2).

*Verifiable exit:* `go build ./...` green; the contract enumerates exactly the
7 named rules in fixed order; the relative-path and drive-relative controls
live **inside** rule 5 and do not appear as an eighth rule.

**U-B2 — Re-found the package test surface** *(tests domain)*
*Files:* `internal/config/{workspace,hostcli→copilotcli,paths,decode_semantics,load,decode,detail,validate}_test.go`.
*Change:* rename `hostcli_test.go` → `copilotcli_test.go` and rewrite against
`copilot.cli_path`. Update the shared `newValidBaseConfig` helper (drops
`HostCLI`). **Delete `TestValidateRule6WinsOverRule8`** — that precedence no
longer exists, both rules are gone; it is not renumbered. **Re-pin
`TestValidateRule1WinsOverRules6And8Simultaneously`** as a multi-violation
determinism fixture over the surviving 7-rule contract (I2).
*Tests (new/changed scenarios):*
* V1 empty `cli_path` → loads clean, **zero** warnings;
* V3 absolute non-existent `cli_path` → error;
* V4 bare name not on PATH → advisory only;
* **relative `cli_path` → error** (new control, design §6.2);
* V2 shipped-shape config → all retired keys in `Report.UnknownKeys`,
  non-fatal. **Two verified edge cases must be pinned:** (a) with
  `Slack.MarkdownUploadExtensions` gone from `mapFieldPaths`, a legacy
  `[slack.markdown_upload_extensions]` table whose keys differ only by case
  (`.MD`/`.md`) becomes a **fatal** case-fold collision. **Decision (owned
  here, not left as a fork):** accept this as a *documented, tested behaviour
  change* rather than restoring the exemption. Restoring it would require
  hard-coding the literal string `slack.markdown_upload_extensions` into
  `load.go` — a production-code change in another unit's file, and a string
  the U-E1a gate forbids in non-test package code. The exception is therefore
  asserted by test here and called out in U-D3's Migration section as the one
  case where a legacy file fails rather than warns. This is a **knowing,
  recorded I4 exception with an owner**, not an unnoticed violation;
  (b) two `[[workspace]]` entries each carrying `channel_id` produce the
  **duplicate** unknown-key path `workspace.channel_id` twice (the sanitizer
  sorts and caps but does not dedupe), so assertions must expect the duplicate.
* **Numeric couplings in `decode_test.go` must be re-derived, not assumed**
  (this is the plan's top declared risk class): the cap marker's "136 more"
  becomes 137 once `host_cli` joins the unknown set; the escapes test's
  `len(UnknownKeys) != 1` becomes 2; `TestDecodeReportsUnknownKeysTolerantly`'s
  want-slice gains `host_cli`; and `TestDecodeRejectsCaseFoldCollision`'s
  `host_cli`/`Host_CLI` fixture stops colliding once `host_cli` is unknown.
* **Re-found the deferred-commit coverage.**
  `TestValidateRule6FailureLeavesDefaultWorkspaceRootUnmodified` is the **only**
  test proving that a validation failure occurring **after** rule 2b's
  canonicalization does not mutate the receiver (Decision R10);
  `validate_test.go`'s equivalent fails at rule 1, *before* canonicalization.
  Rule 6's elimination would silently delete that coverage, so it must be
  re-founded on a surviving post-2b rule (rule 5's absolute-path branch, 6, or
  7) — renamed accordingly.
*Verifiable exit:* `go test ./internal/config/...` green.

### Sub-epic D — Defaults, docs, and operator artifacts

**U-D1 — Re-pin the defaults and the count-pinned test tables** *(tests domain)*
*Files:* `internal/config/default_test.go`, `internal/config/docs_test.go`.
*Ownership note:* the two **doc-comment** pin sites (`config.go`'s package doc
and `default.go`'s doc comment) are owned by **U-B1**, whose exit criterion
covers them explicitly. U-D1 owns the two **test** pin sites; U-D3 owns the
reference doc. Five sites, three owners — no gap and no double-ownership.
*Change:* recompute the non-zero default set **from the code**, never by
arithmetic in a document. *(Cross-check only, not the source of truth: the
shipped set is 17; removing the 4 ACP defaults and `IPCName`, and adding
nothing — `startup_timeout_seconds` is deferred to C2 — gives **12**.)*
Update the two test pin sites: `default_test.go` (`want 17`) and
`docs_test.go` (`len(fieldNames) != 17` **and** `len(messages) != 11` → the
7-rule contract's message templates). **Rename all THREE count-bearing test
names** so no count is re-encoded in an identifier:
`TestDefaultHasAll17NonZeroValues`,
`TestConfigReferenceCoversAll17DefaultFieldNames`, and
`TestConfigReferenceCoversAll10ValidationMessages`.
*Verifiable exit:* the defaults test enumerates field **names**, not a count;
no `17`/`11` literal remains as a pinned count; no count appears in a test
name. *(The `go test -run TestConfigReference` green assertion belongs to
**U-D3**, which lands the reference doc those tests read — asserting it here
would repeat the "gate asserted where it cannot hold" defect.)*

**U-D2 — Rewrite `config.toml.example`** *(config domain)*
*Files:* `config.toml.example`.
*Change:* remove `[slack]`, `[slack.markdown_upload_extensions]`, `[acp]`,
`host_cli`, `host_cli_args`, `ipc_name`, workspace `channel_id`. Rename
`slack_detail_level`. Add `[copilot]` with its single key written explicitly — `example_test.go`
reflects over every toml tag and requires each to appear:
```toml
[copilot]
cli_path = ""   # empty = resolve the Copilot CLI from PATH / COPILOT_CLI_PATH
```
`cli_path` **must** be the empty string: an absolute placeholder would fail the
retained existence check, a path-separator form would now be rejected as
relative, and a bare name is exactly the P2-SEC-2 hazard being removed.
`startup_timeout_seconds` is **not** present — it arrives in C2 with its
consumer. Keep `default_workspace_root = "."` so the example still validates
from the package test working directory.
**Delete `host_cli_args = ["--dangerously-skip-permissions"]`** (P2-SEC-1) and
`host_cli = "claude"` (P2-SEC-2).
**P2-NAME-1:** change the example's `database.path` to `data/agent-rc.db` to
match `default.go`. This edits the **example side only**, so invariant I8
(`default.go`'s `database` value byte-identical) is preserved and the D7
persistence fence is not breached.
**Close the drift-test blind spot** that let P2-NAME-1 survive on `main`:
`example_test.go` is *key-coverage only* — `TestExampleConfigCoversEveryStructField`
compares toml key **paths** via `md.Keys()` and never compares a **value**
against `Default()`. Add a value-agreement assertion with an **enumerated** exemption list for keys
that are legitimately illustrative rather than default-equal. The list must be
declared, not left to the implementer, or U-D3 -- the sole green boundary --
is red on landing. At minimum it must exempt `default_workspace_root`
(`"."` in the example vs `""` in `Default()`), `[commands]`, and the
`[[workspace]]` entries. Every other key must match `Default()` exactly. Without this, U-D2 fixes one instance of P2-NAME-1 but not its cause.
*Verifiable exit:* `go test ./internal/config -run TestExampleConfig` green;
example loads with **0** unknown keys; the new value-agreement assertion fails
if `database.path` is reverted to `agent-intercom.db`.

**U-D3 — Rewrite `docs/config-reference.md`** *(docs domain)*
*Files:* `docs/config-reference.md`.
*Change:* rewrite for the corrected schema and the 7-rule contract. Update
every **rule-number** cross-reference (renumbering is a classic stale-reference
source). Add a **Migration** section that, for every removed/renamed/migrated
key, states *what to write instead* — a list of "what changed" without
replacements does not satisfy this. Must cover: the seven removed keys, the one
rename, the two migrations, the changed validation behaviour (rules 4/6/8), and
an explicit security note that `--dangerously-skip-permissions` was removed and
why.
*Verifiable exit:* the (renamed) completeness test passes against U-D1's
recomputed set; no `slack`/`acp`/`host_cli` reference outside the Migration
section.

### Sub-epic E — Anti-regression gate

**U-E1a — Retired-architecture gate script** *(script domain)*
*Files:* `scripts/check-retired-architecture.sh` (new), plus a committed
self-test fixture stored **outside** the gate's scan scope (e.g.
`scripts/testdata/`).
*Change:* fail when `slack`, `socketmode`, `channel_id`, `team_id`, `acp`,
`host_cli`, or `ipc_name` appears as a Go identifier or TOML key in
**`internal/config/**` (excluding `*_test.go` and any `testdata/` directory)**,
plus `config.toml.example`, plus `cmd/**`.
> **Scope corrections (two P0s, attempts 2 and 3).** (i) The first draft scoped
> this to `internal/**`, which **cannot pass**: `internal/apperr` legitimately
> declares `KindSlack`, `KindIPC`, `KindACP` and invariant I5 freezes it.
> (ii) Narrowing to `internal/config/**` *still* could not pass, because U-B2's
> **mandatory** V2 migration fixtures must contain `[slack]`, `[acp]`,
> `host_cli`, `ipc_name`, and `channel_id` as TOML string literals inside
> `internal/config/*_test.go` — they are precisely what proves tolerant decode
> (I4). **`*_test.go` and `testdata/` are therefore excluded**, and the
> self-test fixture lives outside the scan scope so the gate does not redden on
> its own control. `cmd/**` is *added* (the retired `intercom-ctl` lives there;
> verified clean today). `docs/**`, Migration sections, and Go **comments** are
> exempt — stated explicitly so each exclusion is deliberate, not accidental.
> This scope is mirrored normatively in governing decision **D6**, so the plan
> does not silently broaden its governing artifact.
*Change (robustness):* `set -euo pipefail`; scan **tracked paths only**, so an
untracked local artifact cannot alter the verdict; `--self-test` mode driven by
the committed fixture, so the positive control is a **repeatable artifact**
rather than a one-shot manual reintroduce-then-revert procedure. The script
header must reproduce the threat-model limits from D6b.
*Verifiable exit:* `--self-test` exits non-zero on the fixture and zero on the
clean tree.
*Note:* `.ps1` counterpart is explicitly **out of scope** for this slice.

**U-E1b — Wire the gate into CI, non-blocking** *(CI domain)*
*Files:* `.github/workflows/ci.yml`.
*Change:* add the script as **one additive step inside the existing `lint`
job**, carrying **`continue-on-error: true`** (a step, not a new job — a new
job would additionally require editing `ci-gate`'s `needs:` array, breaking
invariant I6; and without `continue-on-error` a failing step fails the `lint`
job and therefore `ci gate`, which is exactly PA-4's risk).
> **Durability caveat (P1).** `ci.yml` carries
> `# Generated by autoharness | Template: ci/ci.yml.tmpl`, so a hand edit can
> be reverted by the next harness render. This must be recorded in the closure
> artifact, and the durable fix (upstream template change) noted as follow-up.
> Note also that `lint` is gated on `needs.changes.outputs.code == 'true'`, so
> the gate is **skipped on docs-only PRs** — acceptable, but stated.
*Verifiable exit:* the step is present with `continue-on-error: true` and the
`lint` job stays green (checkable by diff + one CI run).
*Blocked on:* **H1 operator approval** before any flip to a *required* check —
and per D6b that flip additionally needs the CODEOWNERS protection tracked in
stash `BEDD2E70`, without which "required" conveys assurance it does not have.

## Dependency Graph

```text
U-A1  (go 1.24 floor)            [independent; deferrable]

U-B0  (remove channel routing)
   └─► U-B1a (delete retired schema + validation)
          └─► U-B1b (add [copilot].cli_path + rule 5)
                 └─► U-B2 (test surface re-founded)
                        └─► U-D1 (defaults + count-pinned tables)
                               ├─► U-D2 (example)
                               └─► U-D3 (reference doc)
                                      └─► U-E1a (gate script)
                                             └─► U-E1b (CI wiring, non-blocking)
```

Strictly linear apart from U-A1 and the U-D2/U-D3 pair. No cycles.
The previous draft's claim that "B1–B5 are independent" was wrong (B3 consumed
`cli_path` introduced by B2) and is withdrawn along with that decomposition.

## Deterministic Gates and Rollback Points

> **Gate correction (P0).** The previous draft asserted `go build && go test`
> green "after **every unit**". That was mechanically unachievable: deleting a
> struct field in one unit and its consumers in another guarantees a red tree
> in between. The gate is restated at the boundary where it is actually true.

**G1 is evaluated at each ROLLBACK POINT, not after every unit.** Within a
rollback point, intermediate commits may be red; the point itself must be green.

> **Rollback-point correction (P0, attempt 3).** The previous draft placed a
> green boundary after U-B2. That is still false: `default_test.go` asserts
> `want 17` (fixed in U-D1) and `example_test.go` reflects over every toml tag
> and requires `copilot.cli_path` to be present in the example (added in
> U-D2), while `docs_test.go`'s completeness test needs
> `docs/config-reference.md` rewritten (U-D3). **The config remediation is
> irreducibly atomic**: schema, validation, tests, defaults, example, and
> reference doc are mutually coupled by two mechanical drift tests, so the
> first green tree after U-B0 begins is at **U-D3**. Stating anything else
> reproduces the very "gate asserted where it cannot hold" defect this plan
> has now hit twice.

| Gate | Command | Pass condition |
|---|---|---|
| G1 | `go build ./... && go test ./...` | green at every rollback point |
| G2 | `go vet ./...`, `golangci-lint run` (lint job); `go mod tidy` dirty-check (test job) | clean |
| G3 | `go test ./internal/config -run TestExampleConfig` | example loads, **0** unknown keys; `copilot.cli_path` present as a written key |
| G4 | `go test ./internal/config -run <renamed reference test>` | doc completeness against U-D1's recomputed set |
| G5 | validation-contract test | exactly **7** rules **enumerated by name**, fixed order, multi-violation fixture. Rule 5 is the single three-branch `cli_path` rule (empty \| bare \| absolute-exists; relative rejected) — the relative-path control must **not** appear as an eighth rule |
| G6 | `scripts/check-retired-architecture.sh --self-test` | non-zero on fixture, zero on clean tree |
| G7 | CI `test`/`lint`/`security` + cross-compile matrix | green; 4/4 legs |
| G8 | `go test -race ./...` | clean (baseline hygiene before phases C4–C7) |

**Rollback points** (each independently revertible; no forward-only steps):

1. After **U-A1** — toolchain floor alone.
2. After **U-D3** — the **whole** config remediation (U-B0 → U-D3): package,
   tests, defaults, example, and reference doc all coherent and green.
   *(U-B0…U-D2 are intermediate commits within this point, not boundaries.)*
3. After **U-E1a** — gate script present and self-testing, not yet wired.
4. After **U-E1b** — gate reporting in CI, non-blocking.

No destructive data actions, no backfill, no irreversible steps.

## Decisions and Rationale

* **R1 — Do not add the SDK dependency in this slice.** `go mod tidy` prunes
  unused modules, so a dependency with no importer is mechanically unstable in
  CI. The dependency lands in C2 with its first real import. This keeps C1
  single-purpose: *remove the wrong architecture*.
* **R2 — Raise the Go floor now, not in C2.** It is a one-line, independently
  revertible change, and it surfaces any toolchain/cross-compile/govulncheck
  fallout **before** SDK work depends on it. Cheap early signal.
* **R3 — Migrate `host_cli` rather than delete it.** The naive reading is
  "the SDK owns the CLI, so delete the knob." That is wrong: the Go SDK does
  **not** bundle the CLI and resolves it from `PATH`/`COPILOT_CLI_PATH`, and
  `StdioConnection.Path` is public. The knob survives with re-founded
  semantics. Deleting it would force operators to rely on ambient `PATH` with
  no override.
* **R4 — Eliminate rule 6 rather than keep it required.** Under SDK
  resolution, empty legitimately means "find it on `PATH`". Keeping it required
  would reproduce P2-COMPAT-1 in a new key. Note this is *elimination*, not
  weakening: no check, no message, no contract entry survives, which is why the
  contract goes 10 → 7 and not 10 → 8.
* **R4a — Compensate the loosening with a relative-path rejection.**
  `cli_path` selects a binary that is spawned as a child process with agent
  capabilities. Empty (PATH) and absolute (existence-checked) are permitted;
  **relative is rejected** (design §6.2), because a relative path is neither
  existence-checked nor containment-checked and resolves against the process
  working directory. Ambient-`PATH` resolution for a bare name remains an
  **accepted residual risk** — it is the SDK's own behaviour and is not made
  worse here — recorded rather than silently validated.
* **R5 — Recompute the defaults count from code.** Deriving it arithmetically
  in a document re-introduces exactly the error class this package already
  recorded twice ("16 vs 17 defaults", "9 vs 10 rules") — and which the first
  draft of this plan reproduced a **third** time in its "10 → 7" arithmetic.
  The test must enumerate names, not assert a count, and count-bearing test
  **names** are renamed so the count is not re-encoded in an identifier.
* **R6 — Preserve the fixed rule ordering property.** The shipped contract's
  determinism (which rule fires first) is a real behavioural guarantee that
  survives the correction; only membership changes.
* **R7 — Enforce retirement mechanically (U-E1a), with honest limits.**
  Documentation and review diligence already failed once here — the retired
  architecture reached `main`. A gate is the **strongest available** control,
  but it is explicitly **not** "the only durable control": `ci.yml` is
  autoharness-generated and a `pull_request` workflow runs from the PR head, so
  the gate is an anti-accident control (D6b). The claim is downgraded rather
  than dropped.
* **R7a — Narrow the gate's scope to what can actually pass.** `internal/**`
  was wrong: `internal/apperr` legitimately declares `KindSlack`/`KindIPC`/
  `KindACP`, and I5 freezes that package, so the gate could never go green.
  Scope is `internal/config/**` (excluding `*_test.go` and `testdata/`) + `config.toml.example` + `cmd/**`. `*_test.go` and `testdata/` are excluded because U-B2's mandatory legacy migration fixtures live there — a gate that reddens on its own required fixtures can never pass. The apperr
  taxonomy contamination is tracked separately (D6a), not ignored.
* **R8 — Leave `retention_days` and `default.go`'s `database` value
  untouched.** D7 defers persistence; pre-empting it is beyond this slice's
  authority, and keeping them costs nothing because they are inert. P2-NAME-1
  is resolved on the **example side only** (U-D2), so I8 holds.
* **R9 — Land schema and validation edits atomically (U-B1).** Splitting them
  guarantees a non-compiling tree, because `validate.go` and `workspace.go`
  consume the very fields being deleted. Width isolation is preserved by
  keeping *tests* (U-B2), *config* (U-D2), *docs* (U-D3), *script* (U-E1a) and
  *CI* (U-E1b) in separate units — the boundary that actually matters here is
  compile-integrity, not file count.

## Risks and Caveats

| Risk | Likelihood | Mitigation |
|---|---|---|
| Go 1.24 floor breaks a CI leg (cross-compile / govulncheck / `go mod tidy` dirty-check) | low | U-A1 is independently revertible; toolchain is already 1.26.5 so the language-floor move is conservative; G2 adds the tidy dirty-check explicitly |
| Defaults/rule count drift across the **five** pin sites | **medium-high** — this exact error class occurred twice in P2 history, and the first draft of this plan reproduced it a third time (the "10 → 7" arithmetic) | U-D1 recomputes from code and owns all five sites; count-bearing test **names** are renamed so counts are not re-encoded in identifiers; G4/G5 fail loudly |
| Rule renumbering silently changes which error an operator sees first | medium | U-B2 re-pins the multi-violation determinism fixture (I2); `TestValidateRule6WinsOverRule8` is **deleted**, not renumbered |
| Gate produces false positives, blocking unrelated PRs | medium | Scope narrowed to `internal/config/**` (excluding `*_test.go` and `testdata/`) + `config.toml.example` + `cmd/**`; `internal/apperr` **and the package's own legacy migration fixtures** deliberately excluded (D6, D6a); Go comments out of scope; committed-fixture `--self-test` proves both directions |
| Gate is not the durable control R7 claims | **medium — confirmed** | `ci.yml` is autoharness-generated, and a `pull_request` workflow runs from the PR head, so the gate is an **anti-accident**, not anti-adversary, control (D6b). Recorded in the script header and the closure artifact; upstream template change noted as follow-up |
| Removing `SlackConfig` turns a legacy `[slack.markdown_upload_extensions]` case-collision into a **fatal** error, violating I4 | medium | Resolved as an owned decision, not a fork: U-B2 asserts the fatal collision as a **documented, tested I4 exception** and U-D3 surfaces it in the Migration section. Restoring the exemption is rejected — it would hard-code `slack.markdown_upload_extensions` into `load.go`, which the gate forbids |
| Tolerant decode accepts retired keys instead of erroring | low | By design — retired keys surface via `Report.UnknownKeys` (I4), asserted in U-B2. Operators get a signal, not a wall, during migration |
| Reviewers assume the Rust oracle still governs config | medium | Superseded banners on all four historical decision artifacts; the governing decision demotes the oracle (D8) |
| Executor confuses unit IDs with roadmap phase IDs | low | Units namespaced `U-*`; phases remain `C1…C11` |

## Plan Hardening Signals (REQUIRED)

| Signal | Present | Justification |
|---|---|---|
| Public API, schema, or contract change | **YES** | Breaking rewrite of a shipped operator-facing TOML schema, of the 10-rule validation contract, **and** of `internal/config`'s exported Go surface (three channel-routing resolvers removed/re-founded in U-B0) |
| Security, auth, permission, or compliance-sensitive behavior | **YES** | Removal of `--dangerously-skip-permissions` (P2-SEC-1); eliminating rule 6's presence check is a deliberate loosening of an input-validation control on a path that **spawns an agent-capable child process**, partly compensated by a new relative-path rejection (design §6.2) |
| Migration, backfill, destructive/irreversible action | **YES** | Operator config files written for the shipped schema will not load unchanged. Migration guidance is mandatory (U-D3). No data backfill, no irreversible step |
| External integration, operator checkpoint, external dependency | **YES** | Go 1.24 floor driven by an external SDK; CLI resolution depends on an external, unbundled Copilot CLI (Q1 unresolved); U-E1b is gated on operator checkpoint H1 |
| High runtime, rollout, or rollback risk | **NO** | Both binaries return `not implemented`; no running service. Every step is text and independently revertible |

**Requires plan hardening: yes**

## Runtime Verification and Closure

| Unit | Runtime surface changed? | Verification before considered absorbed | Closure artifact |
|---|---|---|---|
| U-A1 | Build/CI toolchain | CI green across `test`, `lint`, `security`, 4/4 `cross-compile`; `go mod tidy` clean | CI run ID recorded in closure |
| U-B0 | Exported Go API of `internal/config` (zero external callers) | `go build ./...`; `resolve_test.go` green against `workspace_id` lookup | API-change note in closure |
| U-B1 | Config load + validation path (not yet reachable at runtime) | V1–V5 + relative-path rejection; corrected-architecture config loads clean; shipped-shape file reports retired keys as unknown, non-fatal | Migration note in `docs/config-reference.md` |
| U-B2 | None (tests) | `go test ./internal/config/...` green; G5 contract test | — |
| U-D1 | None (tests) | G4; five pin sites agree; no count re-encoded in a test name | — |
| U-D2 | Operator-facing example | G3 — example loads with 0 unknown keys | Example is itself the artifact |
| U-D3 | Operator-facing docs | G4; every rule-number cross-reference updated | Migration section |
| U-E1a | None (script only) | G6 `--self-test`: non-zero on fixture, zero on clean tree | Script header records D6b threat-model limits |
| U-E1b | CI reporting (non-blocking) | Step runs and reports on a PR without blocking | Closure records the autoharness-regeneration caveat + H1 as open |

**Ownership:** Ship, under the governing decision. **Validation window:**
through phase **C2** (SDK proving spike), the first real consumer of the
corrected schema and the first genuine signal that the correction holds.

---

# Phased Roadmap (post-correction)

Replaces the retired P0–P15 Rust-parity roadmap. Strictly sequential, one
shipment per phase, **no parallel implementation worktrees (P-016)**. Each
phase declares a machine-verifiable exit contract so the sequence is safe for
future dark-factory (unattended) execution.

| Phase | Scope | Depends on | Machine-verifiable exit contract |
|---|---|---|---|
| **C1** *(this plan)* | Architecture correction: config remediation, Go 1.24 floor, anti-regression gate | `003-F` (shipped) | G1–G7 above; retired-architecture gate green |
| **C2** | **SDK proving spike.** Add pinned `copilot-sdk/go v1.0.11`. Prove S1 permission round-trip, S2 event-union unknown-variant observability, S3 `Abort`/`Stop`/`ForceStop`. Record the validated Copilot CLI version (**Q1**). **Plus four mandatory concurrency-discovery questions:** (a) is the `Session.On` callback serialised or concurrently re-entrant? (b) is permission dispatch independent of event dispatch, or is there head-of-line blocking? (c) does `Abort` unblock a blocked `PermissionHandlerFunc`? (d) does `SessionEvent` carry a stable identity usable for resume de-duplication? | C1 | Spike findings artifact exists; `go.sum` pins `a550258d…`; S1/S2/S3 each have an executed result; CLI version recorded; all four (a)-(d) answered empirically **AND the design amended with the branch each answer selects** before C2 closes -- "answered empirically" alone is insufficient, since an adverse (a) or (d) invalidates the ordered fold in §5.1. (d) is additionally a blocking gate on C4 |
| **C3** | **Agent adapter (ACL).** `internal/agent`: SDK client/session lifecycle behind an intercom-go-owned interface; permission decisions converted in exactly one file; every event type switch has a logging `default:` | C2 | No package outside `internal/agent` imports `copilot-sdk/go/rpc` (mechanical gate); exhaustive-`default:` test; **`goleak` clean on adapter teardown**, and the `Session.On` unsubscribe func is called |
| **C4** | **Event envelope + canonical session state.** Envelope schema v1 with `(gen, seq)` cursor, hub-only `seq` assignment, bounded replay ring, event-fold projection, D4a session-pointer SET (cardinality-neutral per Q6) written by the adapter off-hub | C3 | Envelope round-trip/golden tests; `gen` proven never to repeat across a simulated restart; pointer lifecycle tests (atomic replace, resume-failure falls back to a new session with an `error` envelope, protocol-version mismatch refuses resume); declared snapshot cost cap and concurrent-hydration cap asserted; **per-client delivery-order assertions under a concurrent producer** (not bare `seq` monotonicity, which passes trivially); `gen`-mismatch forces full hydrate; `-race` clean |
| **C5** | **Hub + first-responder arbitration.** Register-before-broadcast, CAS resolution, hub-owned timeout, capacity-1 reply channel, `permission.resolved` with winning actor. **Authorization model decided and TESTED (design §6.1, H5)** | C4 | An action from a surface **not holding the approver role is rejected with an explicit reason**, asserted by test -- a blocking hold with no downstream test is prose, not a gate; concurrent-responder test proves exactly one winner; `permission.aborted` emitted when the reply hand-off fails; **timeout-vs-answer, shutdown-with-pending, disconnect-with-pending, and last-client-gone cases all covered**; per-test hang timeouts; `goleak`; `-race` clean |
| **C6** | **HTTP/WS transport.** Loopback listener, upgrade auth + `Origin` validation, per-client write pump, drop-on-full backpressure, ping inside the write pump | C5 | Slow-client test proves drop-not-block; auth-required and CSWSH/`Origin` tests; **privileged actions re-validate authorization per message, not only at upgrade**; **no double-close under concurrent drop paths**; ping proven to share the write pump; `-race` clean |
| **C7** | **Hydration and reconnect.** Snapshot + high-water + registration as one hub step; `(gen, seq)` cursor replay; fallback to full hydrate | C6 | No-gap/no-duplicate property test **running with a concurrent event producer** (otherwise it passes trivially); client-side gap detection proven; idempotent replay test |
| **C8** | **Bubble Tea TUI.** Viewport transcript, streaming deltas, permission modal, arbitration-aware clearing | C7 | Golden-frame render tests; modal clears on remote resolution |
| **C9** | **Process lifecycle + autoharness integration.** Single-port allocation, startup readiness at `Start`, ordered shutdown (§3.1) | C8 | Shutdown-ordering test; no stranded CLI child process after `Stop`/`ForceStop` |
| **C10** | **Dev Tunnel boundary.** Backend-side auth independent of edge gating; loopback-only bind assertion | C9 | Tunnel-boundary auth test; bind-address assertion test |
| **C11** | **Persistence decision + implementation** *(only if D7 deliberation concludes it is needed)* | C10 | Deferred — contract defined by that deliberation |

**Dark-factory readiness notes.** Every phase above has (a) a single
predecessor, (b) a mechanically checkable exit contract rather than a
subjective one, and (c) a revert-to-previous-phase rollback point. No phase
requires a parallel branch or worktree. Phases C3–C10 each carry an
anti-corruption or concurrency invariant that is expressed as a **test**, not
as prose, so an unattended executor can verify absorption without human
judgement.

---

# Plan Hardening

**Hardening required: YES.** Four of five signals are present (schema/contract
change, security-sensitive behaviour, migration, external dependency). Hardened
2026-09-04.

## Context consulted

* `docs/compound/2026-05-07-backlogit-shipment-status-constraints.md` — only
  entry in the learnings library; **not applicable** to this plan (it concerns
  shipment closure mechanics, a Ship concern). No prior config-migration or
  rollout learnings exist. Recorded so the gap is visible rather than implied.
* `.github/instructions/strict-safety.instructions.md` — `ProposedAction` /
  `ActionRisk` / `ActionResult` vocabulary applied below.
* `.github/instructions/constitution.instructions.md` — Principles II
  (Test-First, NON-NEGOTIABLE), III (Workspace Isolation), V (Structured
  Observability), VII (Destructive Command Approval, NON-NEGOTIABLE).
* Shipped source read directly: `internal/config/{config,default,validate}.go`,
  `config.toml.example`, `docs/closure/003-S-003-F-post-merge-closure.md`.

**Prior-history signal that shaped this hardening:** the P2 record contains
**two** counting errors that reached artifacts before being caught ("16 vs 17
non-zero defaults" and "9 vs 10 validation rules"). This plan changes **both**
of those counts at once. That is the single highest-probability defect in this
slice, and I1/V1 below exist specifically to contain it.

## Protected invariants

| # | Invariant | Why it is load-bearing | Enforcing gate |
|---|---|---|---|
| **I1** | The non-zero default set is **enumerated by name** in a test, never asserted as a bare count | Two prior counting errors in this exact package | U-D1 test enumerates names; G3/G4 fail on drift |
| **I2** | Validation rules fire in a **fixed, deterministic order**; a multi-violation document always reports the lowest-numbered violation | A shipped behavioural guarantee that survives the correction; membership changes, ordering does not | C3 ordering test with multi-violation fixture |
| **I3** | `OperatorDetailLevel`'s default (`Standard`) is **not** the iota zero value | Deliberate P2 design so a struct-literal `Config` cannot silently become "minimal" | U-B1 renames; U-D1 re-asserts the property |
| **I4** | Tolerant decode reports retired keys via `Report.UnknownKeys` rather than hard-failing | Operators migrating a shipped-shape `config.toml` get a signal, not a wall | U-B2 asserts retired keys appear as unknown |
| **I5** | `internal/apperr` and `internal/pathsafe` are **not modified** | Explicit scope fence; stash `7028FBE7` remains correctly deferred. **Consequence:** `apperr`'s `KindSlack`/`KindIPC`/`KindACP` survive this slice, so the U-E1a gate MUST exclude `internal/apperr` or it can never pass (D6a) | Gate scope excludes `internal/apperr`; diff review |
| **I6** | `.github/workflows/ci.yml` receives **exactly one additive step inside the existing `lint` job** — not a new job (a new job would additionally require editing `ci-gate`'s `needs:` array) | Prior plans fenced this file as approval-required | Diff review against merge-base |
| **I7** | The `toolchain go1.26.5` pin (GO-2025-3750 remediation) is preserved byte-identical | Removing it re-opens a known CVE that govulncheck will fail on | U-A1 asserts the line is untouched |
| **I8** | `retention_days` and **`default.go`'s `database` value** are left byte-identical | D7 defers persistence; this slice has no authority to pre-empt it. P2-NAME-1 is resolved on the **example side only** (U-D2), so the fence holds | Diff review |
| **I9** | No unit touches `cmd/**` | `DefaultConfigPath` is unchanged; the first draft's "at most one flag-default line" invited unnecessary `cmd/` edits | Diff review |

## Risky actions (ProposedAction / ActionRisk)

### PA-1 — Eliminate validation rule 6 (`cli_path` may be empty)

* **summary:** an input-validation control that currently *requires* a value is
  deliberately eliminated so empty is accepted.
* **targets:** `internal/config/validate.go`, `internal/config/copilotcli_test.go`
* **change_kind:** contract change (security-relevant loosening)
* **ActionRisk:** `high`
* **rollback:** revert U-B1's validation hunk; the rule is a single guard clause.
* **approval_required:** **NO** — but the *justification* must be explicit in
  review: empty is not "unvalidated", it is a **defined value** meaning
  "resolve from `PATH`/`COPILOT_CLI_PATH`", which is the SDK's documented
  resolution path. The absolute-path existence check (former rule 7) is
  **retained**, so an absolute path to a non-existent binary still fails closed.
* **Compensating control (added post-review):** a **relative** `cli_path` is
  now **rejected**. A relative path is neither existence-checked nor
  containment-checked and resolves against the process working directory —
  `./bin/copilot` or `..\x\copilot` would otherwise be spawned unchecked with
  agent capabilities. Permitted forms are **empty** (PATH resolution) and **absolute** (existence-checked); a **bare name** is permitted with a non-fatal advisory (accepted residual risk, below). Only the *relative* form is rejected. See design §6.2 for the normative predicate that separates a bare name from a relative path.
* **Residual risk accepted and RECORDED (not validated):** a bare name resolves
  through ambient `PATH`, which is a genuine PATH-hijacking surface on both
  Windows and POSIX for a binary that is spawned as a child process with agent
  capabilities. This is the SDK's own behaviour and is not made worse here. It
  remains a non-fatal advisory in `Report.Warnings`. The first draft understated
  this as merely "not made worse"; it is stated here as an accepted risk.
* **ActionResult:** `planned`

### PA-2 — Delete `host_cli_args = ["--dangerously-skip-permissions"]` from the shipped example

* **summary:** removes an example line that instructs an agent host to bypass
  permission prompts.
* **targets:** `config.toml.example`, `internal/config/config.go`
* **change_kind:** config change (security improvement)
* **ActionRisk:** `moderate`
* **rollback:** revert U-D2/U-B1 — but rollback is **discouraged**; this is a
  deliberate security correction (P2-SEC-1).
* **approval_required:** no
* **ActionResult:** `planned`

### PA-3 — Breaking rewrite of the operator-facing `config.toml` schema

* **summary:** removes/renames seven operator-visible keys and two validation
  rules. Existing operator config files stop validating cleanly.
* **targets:** `internal/config/**`, `config.toml.example`, `docs/config-reference.md`
* **change_kind:** migration (contract break)
* **ActionRisk:** `high`
* **rollback:** revert the U-B/U-D sub-epics; no data migration to undo, no state
  outside source control.
* **approval_required:** no — **and the blast radius is genuinely bounded, not
  merely asserted:** `internal/config` has exactly one non-test importer
  (`cmd/intercom/main.go`) consuming only `DefaultConfigPath`; both binaries
  return `not implemented`; there is no deployed instance and no operator
  config in production.
* **Mandatory mitigation:** U-D3's Migration section must enumerate **every**
  removed/renamed key with its replacement. A migration that lists only "what
  changed" without "what to write instead" does not satisfy this.
* **ActionResult:** `planned`

### PA-4 — Add a CI gate that could block all future PRs

* **summary:** `scripts/check-retired-architecture.sh` wired into CI. A
  false-positive would block every PR until fixed.
* **targets:** `.github/workflows/ci.yml` (one additive step inside the
  existing `lint` job), `scripts/check-retired-architecture.sh` (new)
* **change_kind:** rollout (CI enforcement)
* **ActionRisk:** `high`
* **rollback:** remove the single CI step; the script is inert on its own.
* **approval_required:** **YES for making it a *required* check (H1).** It
  lands **non-blocking** (U-E1b) and may be flipped to required only after
  (a) operator approval, (b) it has run green on at least one subsequent PR,
  and (c) the CODEOWNERS protection tracked in stash `BEDD2E70` covers
  `.github/workflows/**` **and** `scripts/check-retired-architecture.sh`.
  Without (c), "required check" conveys assurance it does not have — a
  `pull_request` workflow runs the file **from the PR head**, so the same PR
  that reintroduces a retired surface can edit the gate. In an
  agent-authored-PR repository this is not hypothetical. This mirrors the
  staged `PIPELINE_TOPOLOGY_GATE_REQUIRED` pattern already used here.
* **Honest characterisation (required in the script header and the closure):**
  this is an **anti-accident control, not an anti-adversary control** (D6b).
  Additionally, `ci.yml` is autoharness-generated, so the step can be reverted
  by the next harness render; the durable fix is an upstream template change,
  recorded as follow-up.
* **Scope discipline:** `internal/config/**` + `config.toml.example` +
  `cmd/**`. **`internal/apperr` is excluded** — it legitimately declares
  `KindSlack`/`KindIPC`/`KindACP` and is frozen by I5, so an `internal/**`
  scope could never pass. `docs/**`, Migration sections, and Go comments are
  exempt. Match on key/identifier context, not bare substring — a
  word-boundary match on `acp` is especially false-positive-prone.
* **Robustness:** `set -euo pipefail`; scan **tracked paths only** so an
  untracked local artifact cannot alter the verdict; `--self-test` against a
  **committed fixture** so the positive control is repeatable rather than a
  one-shot manual procedure.
* **ActionResult:** `planned`

### Not risky — explicitly classified

* **U-A1 (Go floor bump):** `ActionRisk: low`. Toolchain already 1.26.5; language
  floor move is conservative and single-line.
* **U-B0/U-B1/U-B2:** `ActionRisk: moderate` (shared-code/contract edits, no
  security or migration dimension beyond PA-3).
* **No `ActionRisk: destructive` action exists in this plan.** No deletes of
  data, no force operations, no irreversible steps. Constitution Principle VII
  (Destructive Command Approval) is therefore **not triggered** — stated
  explicitly so its absence is a finding, not an omission.

## Deepened runtime verification

**Environment prechecks (before any unit):**

1. `go version` — confirm the active toolchain satisfies the raised floor.
2. `git --no-pager diff --stat main` — confirm a clean starting delta.
3. Confirm `.gitignore`, `.claude/`, `.backlogit/hooks_queue.jsonl` are
   untouched and unstaged (operator-preserved local artifacts).

**Target scenarios (must all be executed, not merely asserted):**

| ID | Scenario | Expected |
|---|---|---|
| V1 | Load a config with **no** `cli_path`, **no** `channel_id` | **Loads clean, zero warnings.** The corrected-architecture happy path and the direct proof P2-COMPAT-1 is fixed. Empty `cli_path` must return early — it must NOT fall through to `exec.LookPath("")`, which would emit a spurious advisory |
| V2 | Load the **shipped-shape** `config.toml` (with `[slack]`, `[acp]`, `host_cli`, `ipc_name`, workspace `channel_id`) | Loads, and **every** retired key appears in `Report.UnknownKeys` — none silently bound, none fatal (I4). **Pin two verified edge cases:** (a) once the map-path exemption is gone, a legacy `[slack.markdown_upload_extensions]` table with keys differing only by case **does** become a fatal collision — assert that behaviour as the documented, owned I4 exception (U-B2), do not assert the opposite; (b) two workspace entries with `channel_id` yield the **duplicate** path `workspace.channel_id` twice — assertions must expect the duplicate |
| V3 | Absolute `cli_path` pointing at a non-existent file | **Fails** — the retained existence check (PA-1 residual control) |
| V4 | Bare `cli_path` name not on `PATH` | Advisory in `Report.Warnings`, **not** an error |
| **V4b** | **Relative `cli_path`** (`./bin/copilot`, `..\x\copilot`) | **Fails** — new compensating control (R4a, design §6.2) |
| V5 | Document violating several rules at once | Deterministically reports the lowest-numbered violation (I2), against the **7-rule** contract. `TestValidateRule6WinsOverRule8` is deleted, not renumbered |
| V6 | `config.toml.example` round-trip | 0 unknown keys (G3); `copilot.cli_path` present as a written key. **No `startup_timeout_seconds`** — it is deferred to C2, so writing it here would add a key with no struct field and redden G3 |
| V7 | Defaults enumeration | Test lists default field **names**; all **five** pin sites agree; `docs/config-reference.md` contains each verbatim (I1, G4) |
| V8 | Gate self-test | `--self-test` exits non-zero on the committed fixture and zero on the clean tree (both directions, repeatable) |
| V9 | `go test -race ./...` | Clean (baseline hygiene before concurrency phases C4–C7) |
| **V10** | `go mod tidy` after U-A1 | Tree stays clean (CI runs a hard dirty-check) |

**Blocked-path handling:** if any of V1–V8 cannot be executed (tooling
unavailable, CI degraded), the affected unit is reported `blocked` — **not**
`applied` — and the shipment does not close. Per P-012, do not substitute an ad
hoc `grep` for a gate that was specified as a test.

## Deepened operational closure

* **Monitoring signals (healthy):** CI `ci gate` aggregation green on the next
  PR; 4/4 cross-compile legs; `security` job green (staticcheck, govulncheck,
  gitleaks); retired-architecture gate green.
* **Failure signals:** `security` job regressing after the Go floor bump
  (most likely fallout from A1); retired-architecture gate producing a
  false positive on an unrelated PR (fallout from PA-4); example/doc drift
  tests failing (fallout from the I1 counting risk).
* **Rollback triggers:**
  * A1 → any cross-compile leg or `govulncheck` failing attributable to the
    language-floor change.
  * PA-4 → two or more false positives on unrelated PRs.
  * PA-3 → discovery of an unlisted consumer of the removed schema fields
    (none exists today; this is a tripwire, not an expectation).
* **Rollback procedure:** `git revert` the offending unit's commit. Units are
  sequenced so each is independently revertible (see *Rollback Points*). No
  forward-only migration exists, so no compensating action is ever required.
* **Owner:** Ship, under the governing architecture decision.
* **Validation window:** through the completion of phase C2 (SDK proving
  spike), which is the first real consumer of the corrected schema and the
  first genuine signal that the correction holds.

## Human checkpoints

1. **before flipping the gate to a required check** (PA-4) — operator confirmation
   required.
2. **Q1 (Copilot CLI minimum version) remains unresolved.** It is *not* a
   blocker for C1 because C1 adds no SDK dependency, but it **must** be
   resolved by C2. Carried forward explicitly so it is not lost at the phase
   boundary.
3. **Q3 (Dev Tunnel auth model) is an unconfirmed assumption.** Not a C1
   blocker; must be operator-confirmed before C10.

## Review-gate capability risks (carry into plan-review)

* This plan must be reviewed by the **Architecture**, **Scope**,
  **Go/Correctness**, **Security**, **Concurrency**, and
  **Schema-CLI-Docs Coupling** personas. Security is non-optional given PA-1
  and PA-2; Schema-CLI-Docs is non-optional given the three-way
  schema↔example↔reference coupling enforced by two mechanical tests.
* **Concurrency** has a genuinely thin surface in C1 (no concurrent code is
  added). Its finding is expected to be "no concurrency surface in this slice"
  — that is a **valid** outcome and must be recorded as such rather than
  padded. Its substantive input belongs to C4–C7.
* Plan review MUST emit literal `dispatch_mode:` and `decision:` markers.
* **P-012 degraded-review condition:** if any persona cannot be dispatched, the
  review must declare the fallback explicitly and record which persona was
  unavailable. A missing persona is a **declared degradation**, never a silent
  pass.

## Unresolved operator decisions

| # | Decision | Blocks |
|---|---|---|
| H1 | Flip the retired-architecture gate to a **required** CI check. Requires operator approval **and** the CODEOWNERS protection in stash `BEDD2E70` covering `.github/workflows/**` and the gate script (D6b) | Not C1 completion; U-E1b lands non-blocking |
| H2 | Copilot CLI minimum version / protocol-3 floor (**Q1**) | **C2** |
| H3 | Dev Tunnel authentication model (**Q3**) | **C10** |
| H4 | Whether `max_concurrent_sessions` and the `[[workspace]]` list survive the one-process-per-workspace topology (**Q6**) | **C9**. Provisionally retained in C1 with re-founded semantics; **no unit in C1 depends on the answer** |
| **H5** | **Authorization model for permission approval** — who may approve what, and whether a *remote* surface may approve destructive shell commands or only the local TUI (design §6.1). Authentication alone is insufficient because any connected surface can approve | **C5** |
| **H6** | Durable store for the D4a session pointer (file, location, and whether it is workspace-local). Deliberately minimal and explicitly **not** a reopening of D7 | **C4** |

## Post-Review Remediation Record (2026-09-04)

This plan **FAILED** its first adversarial review (six personas: Architecture,
Scope, Go, Correctness, Security, Concurrency; four returned FAIL). The
following defects were found and remediated. They are recorded because several
were factual errors that would have produced a non-compiling tree.

| # | Defect (first draft) | Remediation |
|---|---|---|
| 1 | `internal/config/workspace.go` — three exported channel-routing resolvers built on `ChannelID` — was owned by **no unit**. Deleting the field would break the build | New unit **U-B0**, sequenced first |
| 2 | Eight further contaminated files unowned (`load.go`, `resolve_test.go`, `workspace_test.go`, `paths_test.go`, `decode_semantics_test.go`, `load_test.go`, `docs_test.go`, `hostcli_test.go`) | Ground-truth file inventory added; every file assigned an owning unit |
| 3 | Gate G1 ("green after **every** unit") and rollback point 2 were **mechanically unachievable** — schema and validation edits were split across units that consume each other's fields | Schema + validation merged into atomic **U-B1**; G1 restated at rollback points; rollback points redefined |
| 4 | "10 rules → 7, with rule 6 **relaxed** rather than removed" is self-contradictory arithmetic (10 − 2 = 8) — a **third** instance of this package's counting-error class | Decision §B rewritten: rule 6 is *eliminated*, not relaxed; post-change contract enumerated **by name** |
| 5 | Invariant I8 (`database` byte-identical) contradicted U-D1's claim to resolve P2-NAME-1 | P2-NAME-1 resolved on the **example side only** (U-D2); I8 narrowed to `default.go`'s value |
| 6 | The gate scoped to `internal/**` **could never pass** — `internal/apperr` declares `KindSlack`/`KindIPC`/`KindACP` and is frozen by I5 | Scope narrowed to `internal/config/**` + `config.toml.example` + `cmd/**`; apperr contamination tracked separately (D6a) |
| 7 | *(SUPERSEDED BY ROW 31 — the key was subsequently removed from C1 entirely.)* `copilot.startup_timeout_seconds` had no specified default, so it would silently be 0 (an SDK `Client.Start` deadline of zero) | Default **30** specified explicitly in U-B1 |
| 8 | Unit IDs `C1/C2/C3` collided with roadmap phase IDs `C1/C2/C3` in the same document | Units namespaced `U-*` |
| 9 | `ci.yml` is autoharness-generated and `pull_request` workflows run from the PR head — the gate is not "the only durable control" | R7 downgraded; D6b caveat added; CODEOWNERS made a precondition for H1 |
| 10 | Dependency graph claimed B1–B5 independent, contradicted by the plan's own text | Graph rewritten as a linear chain |
| 11 | D9 (Option C = one shipment) was silently narrowed by the plan to C1+C2 | Recorded as explicit amendment **D9a** on the decision artifact |
| 12 | Empty `cli_path` would fall through to `exec.LookPath("")`, emitting a spurious advisory | U-B1 specifies early return with no advisory; V1 asserts zero warnings |
| 13 | Relative `cli_path` was neither existence- nor containment-checked, yet selects a binary spawned with agent capabilities | New compensating control **R4a** / design §6.2: relative paths rejected; V4b added |
| 14 | Design: `seq` had no generation anchor — a restart silently corrupted reconnecting clients' transcripts | `(gen, seq)` cursor; `gen` mismatch forces full hydrate (design §5.2/5.3) |
| 15 | Design: D4 durability was **unbridged** — `ResumeSession` needs a `sessionID` that lived only in memory, so crash-resume was impossible | **D4a** minimal durable session pointer, explicitly distinguished from D7 |
| 16 | Design: permission timeout had two owners (handler deadline + hub CAS), producing silent authorization divergence | Hub owns the timeout; handler returns only what the hub sends (§5.4) |
| 17 | Design: broadcast-before-registration could reject a legitimately granted permission | Register-before-broadcast, one hub iteration (§5.4) |
| 18 | Design: unbuffered reply channel could block the hub forever — one slow operator wedges the process | Capacity-1 reply channel + non-blocking hub send; "hub never blocks" invariant (§4.1) |
| 19 | Design: hydrate-then-register chose the **unrecoverable** failure mode (permanent silent gaps) | Snapshot + high-water + registration as one indivisible hub step (§5.3) |
| 20 | Design: shutdown **deadlocked** on a pending permission and stranded the CLI child | Six-step ordering with "fail pending permissions first" and "stop the hub last" (§3.1) |
| 21 | Design: "exactly one socket writer" was unachievable — ping/pong are writes | Ping fires inside the write pump's own `select` (§5.5) |
| 22 | Design: client drop had no close protocol (double-close / send-on-closed panics) | `done` channel + `sync.Once`; hub never closes `send`; write pump alone closes the socket (§5.5) |
| 23 | Design: SDK callback "must not block" vs "must push to ingress" was self-contradictory; re-entrancy unspecified | §4.2 resolves normatively (buffered + blocking send + hub-never-blocks); re-entrancy made a C2 spike question |
| 24 | Design: authorization was conflated with authentication — any connected surface can approve a shell command | §6.1 records required controls; **H5** blocks C5 |
| 25 | C4–C7 exit contracts relied on `-race`, which detects **no** deadlock, lost wakeup, or ordering violation | Roadmap exit contracts rewritten with the specific failure cases, `goleak`, and hang timeouts |
| | **— Attempt 3 (second adversarial round) —** | |
| 26 | Decision §A/§B **tables** still said rule 6 "RELAX to optional" though the prose said "eliminated" — the contradiction survived in the normative tables | Both table rows rewritten to **ELIMINATED**; §A ACP row rewritten to "all four fields deleted" |
| 27 | The relative-`cli_path` control added as remediation #13 was an **unenumerated 8th check**, re-breaking the 10−2−1=7 arithmetic it was meant to protect | Folded into **one** rule slot: rule 5 = "`cli_path` well-formed (empty \| bare \| existing absolute; relative rejected)". G5 asserts it must not appear as an 8th rule |
| 28 | §6.2's "bare name permitted" + "only empty or absolute permitted" + "relative MUST be rejected" was **unimplementable** — `filepath.IsAbs` cannot separate a bare name from a relative path, and V4/V4b were mutually exclusive | Explicit classification predicate added (empty \| IsAbs \| contains separator → reject \| else bare name), both separators tested regardless of `GOOS` per the `containsDotDotSegment` precedent; Windows drive-relative and UNC cases stated |
| 29 | Gate scope `internal/config/**` **still could not pass** — U-B2's mandatory legacy migration fixtures contain the retired tokens as string literals in `internal/config/*_test.go` (same "gate that can never pass" class as #6, relocated) | `*_test.go` and `testdata/` excluded; self-test fixture stored outside scan scope; scope mirrored normatively in D6 so the plan does not silently broaden its governing artifact |
| 30 | Rollback point "after U-B2" was **still red** (`default_test.go` `want 17`; `example_test.go` needs `copilot.cli_path`; `docs_test.go` needs the rewritten reference doc) | Config remediation declared **irreducibly atomic**: first green boundary after U-B0 begins is **U-D3**. Rollback points reduced to U-A1 \| U-D3 \| U-E1a \| U-E1b |
| 31 | `copilot.startup_timeout_seconds` was an **unconsumed key with no validation rule**, so an explicit `0` would ship a zero-`Client.Start`-deadline hazard; it also added a 6th defaults pin site | **Removed from C1**; lands in C2 with its first consumer and its own rule. Defaults recount is now 17 − 4 − 1 = **12** |
| 32 | The U-* rename left **stale unit IDs** (B4, C3, D1, B1/B2/B5, D2/B3) in Protected Invariants and the risky-action sections, re-creating the very ID collision #8 fixed | Swept; enforcing gates now name real units |
| 33 | Blast-radius prose said "9 non-test + 10 test" — a **fourth** instance of this document's own counting-error class | Corrected to the verified 5 non-test + 12 test = 17, matching the inventory table |
| 34 | The case-fold exemption fork ("preserve the exemption OR document the exception") had **no owning unit** and its first branch was unimplementable (would hard-code `slack.markdown_upload_extensions` into `load.go`, which the gate forbids) | Decided in-place: accepted as a **documented, tested I4 exception** owned by U-B2 and surfaced in U-D3's Migration section |
| 35 | `example_test.go` is **key-coverage only** and never compares values — the blind spot that let P2-NAME-1 live on `main`. U-D2 fixed the instance, not the cause | U-D2 now adds a **value-agreement assertion** with a documented exemption list |
| 36 | `TestValidateRule6FailureLeavesDefaultWorkspaceRootUnmodified` was the **only** test proving no receiver mutation after rule 2b canonicalization; rule 6's removal silently deleted that coverage | U-B2 must **re-found** it on a surviving post-2b rule |
| 37 | `decode_test.go`'s numeric couplings (136→137, `len != 1`→2, want-slice, case-fold fixture) were unnamed | Enumerated explicitly in U-B2 |
| 38 | U-D1 claimed five pin sites but listed two files; `config.go`/`default.go` doc comments were double-owned with U-B1 | Ownership split stated: U-B1 owns doc comments, U-D1 owns test tables, U-D3 owns the reference doc. Third count-bearing test name (`TestDefaultHasAll17NonZeroValues`) added to the rename list |
| 39 | U-B0's "value-not-pointer return re-asserted" was **unsatisfiable** (`WorkspaceRoot` returns a string) | Test deleted rather than re-founded; capability removal recorded explicitly |
| 40 | Design: `gen` "changes every start" was **weaker than the invariant needs** — a reset counter satisfies the letter and reproduces the corruption | `gen` MUST never repeat across process lifetimes; ULID/UUIDv7, captured once at hub construction |
| 41 | Design: §5.1's dedup rule rested on an **unverified** SDK property with no fallback, while §4.2 already records that `SessionEvent` has no sequence number | Both branches specified; the no-identity branch buffers-then-reconciles; C2(d) made a blocking gate on C4; SDK identity normalised inside the ACL so no SDK field reaches the hub |
| 42 | Design: shutdown never said **which goroutine** runs it — on the hub it is a hard deadlock | Dedicated shutdown goroutine; hub keeps draining ingress through steps 0–5, stops at 6 |
| 43 | Design: §4.2's failure path was **circular** — it required emitting a sequenced envelope through the ingress that had just failed | Out-of-band capacity-1 `degraded` channel; the **hub** produces the envelope. "Blocking send" vs "bounded wait" ambiguity removed |
| 44 | Design: bare `WaitGroup` quiesce **races** `Add`/`Wait` given unspecified re-entrancy | One-way "no new entrants" latch required |
| 45 | Design: handler had **no exit path** if the hub never sent | Handler selects on reply **and** hub-`done`, defaulting to `UserNotAvailable` |
| 46 | Design: `permission.resolved` was broadcast as final **before** the decision reached the SDK — the dual-owner defect in miniature | Finality moved to the hand-off; `permission.aborted` emitted when the hand-off fails or a third party abandons the request |
| 47 | Design: TUI "coalesce/drop-oldest" contradicted §5.2's frozen-envelope rule | **drop-oldest only**, never coalesce |
| 48 | Design: "last surface gone" did not define whether the TUI counts | TUI is a **permanent member**; condition means last remote surface gone **and** TUI exited |
| 49 | Design: §5.3's "cost budget" was asserted but never stated; unbounded projection, uncapped concurrent hydrations | Entry/byte cap declared in the C4 exit contract; concurrent hydrations capped per iteration; incremental pre-marshalling preferred |
| 50 | Design: D4a pointer had content but **no lifecycle, cardinality, or writer** — and it is the sole key to the only durable record | Set keyed by `session_id` (Q6-neutral); atomic write-after-create; pointer is a **hint** with new-session fallback; protocol-mismatch refuses resume; **adapter** owns the write, off-hub |
| 51 | H5 "blocks C5" but C5's exit contract had **no authorization assertion** — a hold with no test is prose | C5 exit now requires a rejected-non-approver test; C6 requires per-message re-validation |






<!-- plan-review-attempt: 3 -->




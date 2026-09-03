---
title: "intercom-go Foundation — Module Skeleton and Hardened CI Gates"
description: "Implementation plan for the first shippable, fork-independent slice of the agent-intercom Go port"
source: "docs/decisions/2026-09-03-intercom-go-architecture-reconciliation-deliberation.md"
origin_documents:
  - path: "docs/decisions/2026-07-06-go-port-reference-brief.md"
    role: "governing source"
  - path: "docs/design-docs/intercom-go-backend-architecture.md"
    role: "deferred, non-governing — listed for provenance only (stash A92E3FA0)"
stash_ids:
  - "4989A42D"
date: 2026-09-03
revision: 2
status: reviewed
tags:
  - "go"
  - "foundation"
  - "ci"
---

# Implementation Plan — intercom-go Foundation (rev 2)

> **Revision note.** Revision 1 of this plan failed the plan-review gate with 2 P0 and
> 12 P1 findings. Revision 2 narrows the slice to the genuinely fork-independent portion
> (module skeleton + CI gates), corrects two factually wrong technical claims, and aligns
> with workspace instruction files that revision 1 never consulted. The config, credential,
> and error-taxonomy work formerly in sub-epics C and D is **deferred to stash `037B1552`**
> with all review findings attached. See `## Plan Review` for the full record.

## Problem Frame

`intercom-go` is a greenfield repository: harness tooling and documentation scaffolding
exist, but there is no `go.mod`, no Go source, no build, and no CI workflow. The
[architecture reconciliation deliberation](../decisions/2026-09-03-intercom-go-architecture-reconciliation-deliberation.md)
selected the [Go port reference brief](../decisions/2026-07-06-go-port-reference-brief.md)
as the governing direction and deferred the
[backend architecture document](../design-docs/intercom-go-backend-architecture.md)
pending an architecture-reconciliation decision.

This plan stages the **first slice that carries no exposure to that unresolved fork**: a
Go module, two binary entrypoints, a canonical cross-compilation path, and a hardened CI
pipeline. Every one of these is required verbatim under either candidate architecture.

Revision 1 additionally staged the `config.toml` schema and the `AppError` taxonomy. Review
established that both are **Group-A-specific**, not fork-independent (`slack_detail_level`,
`SLACK_*` credentials, `[[workspace]].channel_id`, `[database].path`, and the `Slack`/`Db`/
`Mcp` error variants all presuppose the port architecture), and that freezing an
operator-facing config contract before the Rust behavioral oracle is mounted risks
codifying accidental semantics. Both are therefore deferred.

Target module map for this slice (brief §4):

* `cmd/intercom` — server binary, stub
* `cmd/intercom-ctl` — companion CLI binary, stub

No `internal/` package is created by this slice. Per review, packages are created by the
unit that gives them behavior, not speculatively.

## Scope Boundaries

**In scope:** Go module initialization, two `cmd/` entrypoints with flag surface and
structured logging, canonical 4-target CGO-free cross-compilation, and CI gates for build,
vet, race tests, formatting, import ordering, linting, static analysis, vulnerability
scanning, secret scanning, and cross-compilation.

**Out of scope (deferred to stash `037B1552`):** `config.toml` schema, config loading and
validation, credential resolution, the `AppError` taxonomy.

**Out of scope (deferred, per the deliberation):** SQLite store and the 7 tables,
`AgentDriver` seam, ACP subprocess, Slack surface, approval/diff/policy, stall detection,
steering, IPC/ctl protocol, config hot-reload, the three risk slices, and everything in
the deferred architecture document.

## Requirements Trace

| # | Requirement (brief §) | Units |
|---|---|---|
| R1 | Go module, `cmd/` layout (§4) | A1, A2, A3 |
| R2 | Two binaries: server + companion CLI (§4, §7) | A2, A3 |
| R3 | `CGO_ENABLED=0` build; cross-compiles to 4 targets from one runner (§13) | A4, B4 |
| R4 | CI runs build, `go vet`, `go test -race` (§12.1) | B1 |
| R5 | CI runs `staticcheck` and `govulncheck` (§12.1, §10) | B3 |
| R6 | Structured logging via `log/slog`, stable field names (§3) | A2, A3 |
| R7 | Workspace standard: `golangci-lint`, `gofmt`, `goimports` clean | B2 |
| R8 | Workspace standard: CI supply-chain hardening | B1, B3 |

## Constitution Check

Required by `.github/instructions/constitution.instructions.md` §Governance
("Every implementation plan MUST include a 'Constitution Check' section").

| Principle | Status | Units | Justification / rejected alternative |
|---|---|---|---|
| I — Quality gates (`gofmt`, `golangci-lint`, `go vet`, tests) | Satisfied | B1, B2, B3 | All four gates present. Rev 1 omitted `gofmt` and `golangci-lint`; restored. |
| II — Test-first (NON-NEGOTIABLE) | Satisfied | A2, A3 | Each entrypoint unit lists its `_test.go` in `Files` and declares an explicit red state. Rev 1's post-hoc test unit was removed. |
| III — Workspace isolation and security boundaries (MUST) | Satisfied | A4, B1, B3 | No filesystem-root handling exists in this slice (config deferred). Build script is constrained to the repo root (I5); secret scanning added (B3); no fixtures derived from operator files. |
| IV — Least surprise / no out-of-tree writes (NON-NEGOTIABLE) | Satisfied | A4 | Build script refuses to write outside the repo root and never recursively deletes. |
| V — Structural safety over convention | Satisfied | A4, B1 | Supply-chain safety enforced by SHA pinning and explicit `permissions:`, not by reviewer vigilance. |
| VI — Observability | Satisfied | A2, A3 | `log/slog` JSON handler from the first commit. |
| VII — Destructive actions require approval (NON-NEGOTIABLE) | **Deviation, documented** | A1 | `.gitignore` is dirty in the worktree from unrelated operator work. A1 appends only, never rewrites. A **Forbidden commands** register is added to Plan Hardening because the repo's torn submodule state makes several routine git commands destructive. Rejected alternative: creating a fresh `.gitignore` — would destroy operator work. |
| VIII — Dependency minimalism | Satisfied | A2, A3 | One direct dependency (`spf13/cobra`). `BurntSushi/toml`, `zalando/go-keyring`, and `go-playground/validator` deferred with the config slice. |
| IX — Reproducible builds | Satisfied | A1, A4, B4 | `go.sum` committed, `-mod=readonly`, pinned toolchain, single canonical target manifest. |
| X — Documented decisions | Satisfied | — | Decisions and Rationale table below; deliberation linked in frontmatter. |
| XI — Closure and verification | Satisfied | — | Runtime Verification and Closure section below. |

**Documented violations register:** one — Principle VII deviation at A1 (`.gitignore`
append into a dirty file), mitigated by append-only constraint and the Forbidden
commands register.

## Implementation Units

Every unit satisfies the 2-hour rule (< 3 files, < 5 functions, < 4 test scenarios),
width isolation (one domain), and produces an atomic verifiable milestone. File lists are
literal — no glob notation — so the file budget is auditable.

### Sub-epic A — Module skeleton and CGO-free build

#### A1. Initialize Go module and build hygiene

* **Changes:** create `go.mod` with module path `github.com/softwaresalt/intercom-go` and
  `go 1.22`; **append** build-output and local-artifact ignores to the existing `.gitignore`
  (`dist/`, `*.exe`, `.env`, `.env.*`, `config.toml`).
* **Files (2):** `go.mod`, `.gitignore`
* **Tests:** `go build ./...` exits 0; `git diff .gitignore` shows appended lines only.
* **Posture:** migration-first (scaffolding precedes behavior).
* **Domain:** config.
* **Size:** XS · **Complexity:** trivial
* **Blocking operator checkpoint:** confirm the module path and the Go 1.22 floor before starting.
* **Constraint:** `.gitignore` is modified in the working tree by unrelated operator work.
  Append only. Never rewrite, reorder, or remove existing lines.

#### A2. Add `cmd/intercom` entrypoint

* **Changes:** `cmd/intercom/main.go` with a `spf13/cobra` root command exposing
  `--config` and `--log-level`, and a `log/slog` JSON handler writing to stderr. `RunE`
  returns a not-implemented sentinel.
* **Files (2):** `cmd/intercom/main.go`, `cmd/intercom/main_test.go`
* **Red state:** `main_test.go` asserts `--help` exits 0 and that `--log-level=debug`
  produces a JSON log line on stderr — fails to compile before `main.go` exists.
* **Tests:** `--help` exits 0; unknown flag exits non-zero; slog emits parseable JSON.
* **Posture:** test-first.
* **Domain:** code.
* **Size:** S · **Complexity:** low
* **Depends on:** A1.

#### A3. Add `cmd/intercom-ctl` entrypoint

* **Changes:** `cmd/intercom-ctl/main.go`, same cobra + slog bootstrap, with `--socket`
  in place of `--config`. `RunE` returns a not-implemented sentinel.
* **Files (2):** `cmd/intercom-ctl/main.go`, `cmd/intercom-ctl/main_test.go`
* **Red state:** `main_test.go` asserts `--help` exits 0 — fails to compile before `main.go` exists.
* **Tests:** `--help` exits 0; unknown flag exits non-zero.
* **Posture:** test-first.
* **Domain:** code.
* **Size:** S · **Complexity:** low
* **Depends on:** A1.

#### A4. Add canonical cross-compile build script

* **Changes:** a single canonical target manifest and build script,
  `scripts/build.ps1` (PowerShell — matches the repo's existing `start.ps1`/`start.sh`
  convention where PowerShell is primary), building both binaries for linux-amd64,
  windows-amd64, darwin-amd64, darwin-arm64 with `CGO_ENABLED=0` into `dist/`.
* **Files (2):** `scripts/build.ps1`, `scripts/targets.json`
* **Tests:** produces 8 artifacts (2 binaries × 4 targets); `go version -m` on a produced
  artifact reports `CGO_ENABLED=0`; script exits non-zero if the resolved output path is
  not a descendant of the repo root.
* **Posture:** test-first.
* **Domain:** config/tooling.
* **Size:** M · **Complexity:** medium
* **Depends on:** A2, A3.
* **Constraint (invariant I5):** `$ErrorActionPreference = 'Stop'`; resolve `dist/` from
  the script's own location; refuse to operate if the resolved path escapes the repo root;
  create/overwrite only — **no recursive delete**. Cleanup, if needed, goes through
  `go clean` scoped to the module.
* **Note:** `scripts/targets.json` is the single source of truth for the target matrix.
  B4 reads the same manifest so local and CI builds cannot drift.

### Sub-epic B — Hardened CI gates

All units in this sub-epic extend `.github/workflows/ci.yml` and must comply with
`.github/instructions/ci-security.instructions.md`.

#### B1. Add hardened core CI workflow

* **Changes:** create `.github/workflows/ci.yml` with:
  * triggers `push` (main) and `pull_request` only — **never** `pull_request_target`
  * top-level `permissions: contents: read`
  * a `concurrency` group cancelling superseded runs
  * `actions/checkout` and `actions/setup-go` pinned to **full commit SHAs** with the
    semver as a trailing comment; `persist-credentials: false` on checkout
  * Go pinned to 1.22.x, matching `go.mod`
  * steps: `go mod verify`, `go build -mod=readonly ./...`, `go vet ./...`,
    `go test -race -mod=readonly ./...`
  * a step failing the job if `go mod tidy` dirties tracked files
* **Files (1):** `.github/workflows/ci.yml`
* **Tests:** workflow green on the PR; every `uses:` is a 40-character SHA; no secrets
  referenced by any job.
* **Posture:** test-first.
* **Domain:** config/CI.
* **Size:** M · **Complexity:** medium
* **Depends on:** A1.
* **Note on `-race` (corrected in rev 2):** `go test -race` requires cgo, and with
  `CGO_ENABLED=0` the toolchain **fails loudly** — `go: -race requires cgo; enable cgo by
  setting CGO_ENABLED=1`, exit status 2, no tests run. It does **not** silently disable the
  detector. Revision 1 asserted the opposite and built a sentinel-package guard on that
  false premise; the guard is removed. The correct handling is simply to leave cgo enabled
  for the race step (the runner default) and set `CGO_ENABLED=0` only on build and
  cross-compile steps. The failure mode is fail-closed and needs no extra machinery.

#### B2. Add format and lint gates

* **Changes:** add a `lint` job running `gofmt -l .` (fails on any output),
  `goimports -l .` (fails on any output, enforcing the workspace 3-group import order),
  and `golangci-lint run` at a pinned version; add a committed `.golangci.yml`.
* **Files (2):** `.github/workflows/ci.yml`, `.golangci.yml`
* **Tests:** all three tools report clean; job fails if a deliberately misformatted file
  is introduced.
* **Posture:** test-first.
* **Domain:** config/CI.
* **Size:** S · **Complexity:** low
* **Depends on:** B1.
* **Rationale:** required by `.github/instructions/technology-go.instructions.md`
  ("passes `golangci-lint run`, `go vet ./...`, and `gofmt -l .` with zero findings").
  Revision 1 omitted all three. Cheapest to add on an empty tree; retrofitting later
  produces a repo-wide diff.

#### B3. Add static-analysis, vulnerability, and secret-scanning gates

* **Changes:** add a `security` job running `staticcheck ./...`, `govulncheck ./...`, and
  a secret scanner (`gitleaks`), all at **pinned versions** — never `@latest`.
* **Files (1):** `.github/workflows/ci.yml`
* **Tests:** all three run and report clean; a planted synthetic secret is detected.
* **Posture:** test-first.
* **Domain:** config/CI.
* **Size:** S · **Complexity:** medium
* **Depends on:** B1.

#### B4. Add cross-compile verification matrix

* **Changes:** add a matrix job over the 4 targets from `scripts/targets.json` with
  `CGO_ENABLED=0`, asserting each build succeeds from a single ubuntu runner.
* **Files (1):** `.github/workflows/ci.yml`
* **Tests:** 4/4 matrix legs green.
* **Posture:** test-first.
* **Domain:** config/CI.
* **Size:** S · **Complexity:** low
* **Depends on:** A4, B1.

## Dependency Graph

```text
A1 (go.mod + .gitignore)
├── B1 (hardened core CI) ──┬── B2 (format/lint gates)
│                           ├── B3 (security gates)
│                           └── B4 (cross-compile matrix) ◄── A4
├── A2 (cmd/intercom) ───────┐
└── A3 (cmd/intercom-ctl) ───┴──► A4 (build script + target manifest)
```

No cycles. Suggested execution order:

`A1 → B1 → A2 → A3 → B2 → B3 → A4 → B4`

CI lands **second**, immediately after the module exists, so every subsequent unit is
gated at authoring time rather than retroactively. Revision 1 sequenced CI sixth; review
established that `go build ./...` on a bare skeleton is a valid green, so there was no
reason to delay the gates.

## Decisions and Rationale

| Decision | Rationale |
|---|---|
| `go 1.22`, not `go 1.21` | `.github/instructions/technology-go.instructions.md` mandates "Target Go 1.22 or later". The brief's "Go 1.21+" is satisfied by 1.22; the stricter workspace floor governs. Go 1.21 is also outside its security-support window as of this date. |
| Full-SHA action pinning + explicit `permissions:` | `.github/instructions/ci-security.instructions.md` requires immutable refs and least-privilege tokens. Version tags are explicitly forbidden there. |
| No `internal/` packages created in this slice | Empty placeholder packages prematurely ratify boundaries that the unresolved architecture fork may invalidate, and `internal/mode`/`internal/driver/mcp` may be retired outright. Packages are created by the unit that gives them behavior. |
| Config and error taxonomy deferred | Both are Group-A-specific rather than fork-independent, and both should be validated against the Rust behavioral oracle — which is not yet mounted — before being frozen as contracts. |
| No race-detector sentinel package | `go test -race` with cgo disabled fails loudly, so the guard solved a problem that does not exist. |
| Single canonical target manifest (`scripts/targets.json`) | Prevents drift between the local build script and the CI matrix, which would otherwise define the target set twice. |
| One build script, not two | The repo's `start.ps1`/`start.sh` pair suggests PowerShell is primary. A second script is duplicated surface with no current consumer; add it when a platform needs it. |
| `spf13/cobra` over `urfave/cli` | Brief §3 lists both; `intercom-ctl` needs a subcommand tree. |
| Stub `RunE` rather than omitting it | Gives CI a real build target and makes `--help` assertable without implying unimplemented behavior works. |

## Risks and Caveats

| Risk | Severity | Mitigation |
|---|---|---|
| Torn `references` submodule tempts a Ship executor into a destructive git command | **High** | Forbidden commands register in Plan Hardening. Explicitly out of bounds. |
| `.gitignore` rewritten instead of appended, destroying operator work | **High** | A1 constraint; `git diff` review before commit. |
| Module path wrong, forcing a repo-wide import rewrite | Medium | Blocking operator checkpoint before A1 — a one-line change now. |
| Go version floor wrong, forcing a toolchain migration | Medium | Folded into the same A1 checkpoint. |
| CI supply-chain compromise via mutable action refs | Medium | Full-SHA pinning (B1), pinned tools (B3), `go.sum` + `-mod=readonly` + `go mod verify` (B1). |
| Local build script and CI matrix drift apart | Low | Single `scripts/targets.json` manifest read by both. |
| Deferred config/error work is lost | Medium | Captured in stash `037B1552` with all 12 review findings attached. |
| CGO-free gate is currently vacuous | Low | All deps in this slice are pure Go. Recorded as invariant I1 with an explicit note that it is a **ratchet preventing future CGO ingress**, not proof the constraint holds — that comes when `modernc.org/sqlite` lands. |

## Plan Hardening Signals

| Signal | Present | Justification |
|---|---|---|
| Public API, schema, or contract change | No | Config contract deferred. Only two stub CLI flag surfaces, changeable at will. |
| Security, auth, permission, or compliance-sensitive behavior | **Yes** | CI supply-chain surface: `GITHUB_TOKEN` permissions, action pinning, credential persistence, secret scanning. No application credential code remains in this slice. |
| Migration, backfill, destructive data/config action, or irreversible step | **Yes** | Not by design, but the repo's pre-existing torn submodule and dirty `.gitignore` make several routine git commands destructive in this working tree. |
| External integration, operator checkpoint, or external dependency | **Yes** | GitHub Actions, pinned analysis toolchain, cross-compilation, and a blocking operator checkpoint at A1. |
| High runtime, rollout, or rollback risk | No | No runtime surface ships; both binaries are stubs. Rollback is `git revert` of net-new files. |

**Requires plan hardening: yes**

## Runtime Verification and Closure

| Unit | Runtime surface? | Verification | Closure artifact |
|---|---|---|---|
| A1 | No | `go build ./...` exits 0; `.gitignore` diff is append-only | Module path + Go version recorded |
| A2 | CLI (stub) | `--help` exits 0; slog emits parseable JSON | `intercom` flag surface |
| A3 | CLI (stub) | `--help` exits 0 | `intercom-ctl` flag surface |
| A4 | No (tooling) | 8 artifacts; `go version -m` confirms `CGO_ENABLED=0`; path-escape guard fires | Target manifest |
| B1 | No (CI) | Green; all `uses:` are SHAs; `-race` genuinely runs | CI gate list |
| B2 | No (CI) | `gofmt`/`goimports`/`golangci-lint` clean | Pinned lint versions + `.golangci.yml` |
| B3 | No (CI) | `staticcheck`/`govulncheck`/`gitleaks` clean; planted secret detected | Pinned tool versions |
| B4 | No (CI) | 4/4 legs green | Release target matrix |

**Slice-level closure:** absorbed when every CI job is green, `scripts/build.ps1` produces
8 artifacts locally, and `go version -m` confirms CGO-free linkage. Owner: Ship executor;
operator owns the A1 checkpoint. Validation window: the shipment PR's CI run.

## Plan Hardening

**Hardening required: yes.** Triggered by the CI supply-chain security surface, the
external-dependency surface, a blocking operator checkpoint, and — most importantly — the
repository's pre-existing dirty and torn working-tree state.

**Learnings consulted:** `docs/compound/`, `docs/closure/`, and `docs/memory/` are all empty
(`.gitkeep` only). No institutional memory exists. `.github/instructions/` **was** consulted
this revision (it was not in revision 1, which caused both P0 findings):
`constitution.instructions.md`, `technology-go.instructions.md`,
`ci-security.instructions.md`, `concurrency.instructions.md`.

### Protected invariants

| # | Invariant | Why | Enforced by |
|---|---|---|---|
| I1 | `CGO_ENABLED=0` build succeeds for all 4 targets | CGO-free static binary is the point of the port (§13). **Ratchet only** — prevents future CGO ingress; does not yet prove the constraint, since every current dep is pure Go. | A4, B4 |
| I2 | `go test -race` genuinely executes | Go's only data-race defense (§10). Fail-closed by toolchain design; no extra guard needed. | B1 |
| I3 | Every CI `uses:` is a full commit SHA; `permissions` is least-privilege; `persist-credentials: false` | Workspace CI-security standard; mutable refs are a supply-chain vector | B1 |
| I4 | Analysis and lint tool versions are pinned | An upstream release must not be able to redden unrelated PRs | B2, B3 |
| I5 | The build script never writes or deletes outside the repo root, and never recursively deletes | Principles IV and VII | A4 |
| I6 | Pre-existing operator files are appended to, never rewritten | `.gitignore` is dirty from unrelated operator work | A1 |
| I7 | No secret, real or operator-derived, enters the repository | Principle III; secret scanning is the mechanical backstop | A1, B3 |

### Risky actions

| ProposedAction | ActionRisk | Approval | Rollback |
|---|---|---|---|
| Append to the operator-modified `.gitignore` (A1) | **Medium** | No, but append-only is mandatory | Targeted revert of appended lines |
| Set module path used by every future import (A1) | **Medium** | **Yes — blocking operator checkpoint** | Repo-wide rewrite (expensive; confirm up front) |
| Set the Go language floor (A1) | **Medium** | **Yes — same checkpoint** | Toolchain migration (expensive; confirm up front) |
| Add third-party CI actions and pinned tooling (B1-B3) | Low | No | Revert workflow file |
| Write build artifacts to `dist/` (A4) | Low | No | Delete `dist/`; it is gitignored |

No destructive, migration, or irreversible action is **planned**. However — see below.

### Forbidden commands (require explicit operator approval)

The working tree carries two pre-existing dirty states unrelated to this work: a
worktree-modified `.gitignore`, and a **torn `references` submodule** (`.gitmodules` and
the submodule pointer staged while the directory is deleted in the worktree). In this
state the following routine commands are destructive and would silently consume operator
work or further damage the submodule pointer. They MUST NOT be run:

* `git clean` (any flags)
* `git checkout -- .` / `git restore .` / `git reset --hard`
* any `git submodule` subcommand, including `update --force` and `deinit`
* any recursive delete outside `dist/`

Ship must also not stage or commit `.gitignore`, `.gitmodules`, `references`, or `.claude/`
beyond the single append-only `.gitignore` change A1 requires.

### Deepened verification

* **I3 must be verified mechanically, not by eye.** B1's acceptance includes a check that
  every `uses:` value matches a 40-hex-character pattern. A human reviewer will not
  reliably catch a `@v4` that slips in later.
* **I5 must be verified by a negative test.** A4 asserts the script exits non-zero when
  the resolved output path is forced outside the repo root. Absence of out-of-tree writes
  is not proven by a passing happy-path build.
* **I7 must be verified by a planted secret.** B3 asserts the scanner detects a synthetic
  credential, proving the gate is wired rather than merely present.
* **I2 needs no special verification.** Corrected in this revision: the toolchain
  fail-closes on `-race` without cgo, so a green race job is itself the proof.

### Deepened closure

| Aspect | Detail |
|---|---|
| Monitoring signals | Per-gate CI status: build, vet, race, gofmt, goimports, golangci-lint, staticcheck, govulncheck, gitleaks, 4-target matrix. |
| Rollback trigger | Any gate red on the shipment PR; any secret detected; any unintended change to `.gitignore`, `.gitmodules`, `references`, or `.claude/`. |
| Rollback procedure | Revert the shipment PR. All files are net-new except the `.gitignore` append, which needs a targeted line revert. No data or config migration to unwind. |
| Owner | Ship executor for the PR; operator for the A1 checkpoint. |
| Validation window | The shipment PR's CI run plus one clean local `scripts/build.ps1` producing 8 artifacts. |

### Operator checkpoints

1. **BLOCKING, before A1 — confirm module path and Go version.** The module path
   `github.com/softwaresalt/intercom-go` is inferred from the source repo owner, not stated
   in either document. The Go floor is 1.22 per workspace standard, versus "1.21+" in the
   brief. Both are one-line changes now and expensive migrations later.
2. **Non-blocking — the torn `references` submodule.** Brief §1.2 wants the Rust repo
   mounted read-only as the behavioral oracle. This blocks the deferred config work and all
   future parity work, but not this slice. Requires operator action; out of bounds for both
   Stage and Ship this cycle.
3. **Non-blocking — MCP retirement.** If confirmed permanently out of scope, the deferred
   error taxonomy drops its `Mcp` variant and `internal/mode` is never created.

### Review-gate capability risks

Reviewer subagent dispatch and cross-model routing were both available in the revision-1
gate (`model_routing.anchor_review` → `openai` / `gpt-5.6-sol` / `high`). Plan review MUST
emit literal `dispatch_mode:` and `decision:` markers. Security Lens coverage is no longer
strictly required now that credential code is deferred, but supply-chain security in B1/B3
keeps it valuable.

### Unresolved operator decisions

* Module path and Go version (checkpoint 1) — **blocks A1**.
* Rust oracle mount (checkpoint 2) — blocks the deferred config slice, not this one.
* MCP retirement (checkpoint 3) — blocks nothing here.

## Plan Review

dispatch_mode: multi-agent
decision: PASS

### Attempt history

<!-- plan-review-attempt: 2 -->

| Attempt | Revision | Verdict | Summary |
|---|---|---|---|
| 1 | rev 1 | **FAIL** | 2 P0, 12 P1 across six personas. Plan anchored on the port brief without cross-checking `.github/instructions/`; two technical claims factually wrong; slice over-scoped into fork-exposed territory. |
| 2 | rev 2 | **PASS** | All P0 and P1 findings closed or deferred with traceability. Residual P2/P3 recorded below. |

### Persona coverage

| Persona | Mode | Model | Attempt 1 verdict |
|---|---|---|---|
| Constitution Reviewer | subagent | caller model | 0 P0, 5 P1, 11 P2, 6 P3 |
| Go Reviewer | subagent | caller model | 0 P0, 4 P1 |
| Scope Boundary Auditor | subagent | caller model | 0 P0, 0 P1, 7 P2, 6 P3 |
| Learnings Researcher | subagent | caller model | **2 P0**, 1 P1, 2 P2, 1 P3 |
| Architecture Strategist | subagent (**anchor route**) | `gpt-5.6-sol` (high) | 0 P0, 2 P1, 5 P2, 1 P3 |
| Security Lens Reviewer | subagent (cross-model) | `gpt-5.6-sol` (high) | 0 P0, 1 P1, 3 P2, 3 P3 |
| Agent-Native Parity Reviewer | not triggered | — | Slice exposes no MCP tools or agent-facing actions |

### P0 findings — both closed

| ID | Finding | Resolution |
|---|---|---|
| P0-1 | CI plan pinned `actions/setup-go` by version tag. `ci-security.instructions.md` MUST-requires full-SHA pinning and explicitly forbids version tags; no `permissions:` block, no `persist-credentials: false`, no concurrency control. | **Closed.** B1 rewritten: full-SHA pinning, `permissions: contents: read`, `persist-credentials: false`, concurrency group. Verified as invariant I3 with a mechanical 40-hex check. |
| P0-2 | `go.mod` set `go 1.21`. `technology-go.instructions.md` requires "Target Go 1.22 or later". | **Closed.** A1 sets `go 1.22`; B1 pins setup-go to 1.22.x; folded into the blocking A1 checkpoint. |

Both were verified independently against the instruction files before acceptance.

### P1 findings — all closed or deferred with traceability

| ID | Persona | Finding | Resolution |
|---|---|---|---|
| P1-a | Learnings | `golangci-lint`, `gofmt`, `goimports` gates missing | **Closed** — B2 added |
| P1-b | Constitution | U1 created ~21 files behind glob notation, breaching the 2-hour rule | **Closed** — A1 creates 2 literal files; no `internal/` packages created |
| P1-c | Constitution | No `## Constitution Check` section (required by Governance) | **Closed** — added, with a documented-violation register |
| P1-d | Constitution | U13 tested U12's code *after* it — inverts test-first | **Closed** — post-hoc test unit removed; A2/A3 declare explicit red states |
| P1-e | Constitution | Config carried workspace roots with no canonicalization; `PathViolation` defined but never produced | **Deferred** to stash `037B1552` with the canonicalization requirement attached |
| P1-f | Constitution | Risks table told implementers to copy a real operator `config.toml` — would commit live channel IDs and operator paths | **Closed** — instruction removed; deferred slice carries a synthetic-fixtures-only rule; B3 adds secret scanning as backstop |
| P1-g | Go, Architecture | U9 validated `host_cli` "when ACP mode is active", but no mode concept existed anywhere in the slice — unimplementable predicate | **Deferred** to `037B1552`; resolution recorded (make unconditional, per the brief's ACP-only direction) |
| P1-h | Go | Defaults applied *after* decode silently overwrite an operator's explicit `false`/`0` | **Deferred** to `037B1552`; resolution recorded (decode into a pre-populated `Default()` struct) |
| P1-i | Go | `[slack]` section omitted while strict decode rejects unknown keys — the first real `config.toml` would be rejected as a typo | **Deferred** to `037B1552`; resolution recorded (add `SlackConfig`; diff struct fields against §9 before enabling strict mode) |
| P1-j | Go | A single `Unwrap()` cannot serve both the kind sentinel and the wrapped cause; the specified conformance tests would not catch either failure mode | **Deferred** to `037B1552`; resolution recorded (implement `Is()` for the kind, reserve `Unwrap()` for the cause) |
| P1-k | Architecture | "Fork-independent" claim materially incorrect — config and error taxonomy are Group-A-specific | **Closed** — slice narrowed to the genuinely fork-independent portion; Problem Frame restated |
| P1-l | Security | Invariant I3 under-specified: sentinel-in-logs testing misses `%v`/`%+v`/`%#v`, `json.Marshal`, `slog.Any`, and wrapped provider errors | **Deferred** to `037B1552`; resolution recorded (`type Secret string` with redacting `String`/`GoString`/`MarshalJSON`/`LogValue`, making I3 hold by type rather than by convention) |

### Technical corrections adopted

The Go Reviewer refuted revision 1's single highest-rated risk. `go test -race` with
`CGO_ENABLED=0` does **not** silently disable the detector — `cmd/go` pre-flight-checks and
aborts with `go: -race requires cgo; enable cgo by setting CGO_ENABLED=1` (exit 2, no tests
run). The failure mode is fail-closed, not fail-open. Revision 1's sentinel-package guard
and "assert the detector is active" requirement were over-engineering on a false premise,
and the Constitution Reviewer independently flagged the same guard as a width-isolation
breach. Both are removed.

Two revision-1 claims were confirmed correct and retained: `BurntSushi/toml` does expose
`MetaData.Undecoded()` for strict-key rejection, and CGO-free `GOOS=darwin` cross-compilation
from a linux runner needs no macOS SDK.

### Residual P2 / P3 — accepted, recorded as follow-ups

* Shipment split (Scope Boundary P2, Architecture P2) — **adopted**; this is the split.
* `scripts/build.ps1` + `.sh` duplication (Scope Boundary P3) — **adopted**; one script plus
  a shared `targets.json` manifest.
* CGO-free gate is currently vacuous (Scope Boundary P3) — **adopted**; I1 annotated as a ratchet.
* Deliberation's "nine axes" count inflated; three are load-bearing (Scope Boundary P3) —
  accepted as a calibration note. The split decision itself was independently upheld by
  both the Scope Boundary Auditor and the Architecture Strategist.
* `AgentDriver` is agent-facing, so the deferred multiplexer would need a separate
  `OperatorGateway` seam rather than being an `AgentDriver` (Architecture P2) — **valuable
  correction to the deliberation's unresolved option (c)**; carried into stash `A92E3FA0`.
* Centralized `apperr` couples all future packages to one taxonomy (Architecture P2) —
  carried into `037B1552`.
* Credential resolution should live outside `internal/config` (Architecture P2) — carried
  into `037B1552`.
* Brief cites `docs/adrs/`, a spike decision, and `docs/ARCHITECTURE.md` that do not exist
  in this workspace (Learnings P2) — folded into operator checkpoint 2.
* `testify` and `tests/contract/` + `tests/integration/` conventions (Learnings P3) —
  carried into `037B1552`; no test directories needed while only `cmd/` exists.
* Env-vs-keychain trust assumptions, keychain account-name contract, rotation semantics
  (Security P3 ×2) — carried into `037B1552`.
* Path isolation must be a mandatory gate on the first unit doing filesystem work
  (Security P3) — carried into `037B1552`.

### Gate rationale

Both P0 findings are closed and independently verified against the instruction files. Of
the twelve P1 findings, five are closed in this revision and seven are deferred **with the
work itself** to stash `037B1552`, each with its diagnosis and recommended resolution
attached — none is silently dropped. The narrowed slice contains no unimplementable
predicate, no unresolved architectural exposure, and no secrets-handling code. Hardening is
present and materially specific, with a Forbidden commands register addressing the repo's
torn working-tree state. Runtime verification and closure are specified per unit.

**decision: PASS** — cleared for harvest.

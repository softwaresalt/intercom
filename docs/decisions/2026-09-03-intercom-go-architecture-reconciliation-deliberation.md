---
title: "intercom-go Architecture Reconciliation and First Shippable Slice"
description: "Reconcile two divergent Go backend design sources and select the first coherent, shippable staging group for intercom-go"
topic: "Which architecture governs intercom-go, and what is the first coherent slice to stage?"
depth: "deep"
decision_status: "decided"
promoted_to: "plan"
linked_artifacts:
  - "docs/plans/2026-09-03-intercom-go-foundation-plan.md"
  - "docs/decisions/2026-07-06-go-port-reference-brief.md"
  - "docs/design-docs/intercom-go-backend-architecture.md"
source_documents:
  - path: "docs/decisions/2026-07-06-go-port-reference-brief.md"
    stash_id: "4989A42D"
    role: "authoritative port spec (Group A)"
  - path: "docs/design-docs/intercom-go-backend-architecture.md"
    stash_id: "A92E3FA0"
    role: "divergent design exploration (Group B, deferred)"
tags:
  - "go"
  - "architecture"
  - "port"
  - "staging"
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


## Problem Frame

The operator supplied two design documents in the `intercom-go` workspace and asked
that they be run through the standard Stage process as, preferentially, a single
thematically coherent "Go backend port" planning source:

1. `docs/decisions/2026-07-06-go-port-reference-brief.md` — self-described
   **authoritative spec** for a feature-complete Go port of `agent-intercom` (Rust).
2. `docs/design-docs/intercom-go-backend-architecture.md` — an architecture design
   titled **"Copilot Remote Multiplexer (Backend)"**.

The staging question is twofold:

* **Coherence:** do these two documents describe one system, such that they can be
  planned together as one covering feature?
* **Scope:** whatever the coherent unit turns out to be, it must be reduced to a
  first slice that fits a single shipment under the 2-hour-per-task rule.

`intercom-go` is currently a **greenfield repository** — no `go.mod`, no Go source,
no CI workflow, no build. Whatever is staged first must therefore stand up the
foundation, because every downstream unit depends on it.

### Constraints

* CGO-free static binary is a hard, explicitly stated constraint ("that is the whole point").
* Go 1.21+ (both documents agree on this).
* The Rust repository is the behavioral oracle for the port; its test corpus defines
  acceptance criteria.
* Stage may not write product code — output is planning artifacts and backlog only.

### Success Criteria

* A single covering feature whose premises are internally consistent (no contradictory
  architecture baked into the plan).
* Every task ≤ 2 hours of human-equivalent effort.
* Provenance preserved to both source documents.
* Anything not staged is preserved in the stash, not lost.

## Research Findings

### Prior learnings

`docs/compound/` is empty (only `.gitkeep`). No prior institutional learnings apply.
Recorded as `confidence: none` — this is a first-of-its-kind decision for this workspace.

### Codebase investigation

The repository contains harness tooling (`.github/`, `.autoharness/`, `.backlogit/`,
`.engram/`, `.graphtor/`), documentation scaffolding, `README.md`, `LICENSE`, and
workspace/start scripts. There is **no Go code, no `go.mod`, and no CI workflow**.
The backlog, queue, shipment list, and checkpoint store are all empty.

A `references` git submodule is registered in the index (`.gitmodules` staged, submodule
pointer staged) but is **deleted in the working tree** — a torn, pre-existing state. The
port brief's Section 1 explicitly calls for the Rust repo to be vendored as a
submodule/subtree for read-only oracle access, so this torn submodule is almost certainly
an abandoned first attempt at exactly that step. It is out of scope for this session
(see Risks) but is a real prerequisite for later parity work.

### Document comparison

The decisive finding is that the two documents are **not two views of one system**. They
conflict on every major architectural axis:

| Axis | Port Reference Brief | Backend Architecture Doc | Compatible? |
|---|---|---|---|
| Product identity | `agent-intercom` Go port | "Copilot Remote Multiplexer" | No — different products |
| Operator surface | Slack (Socket Mode, Block Kit, modals, slash commands) | Local Bubble Tea TUI + iOS app | **Direct conflict** |
| Remote transport | Slack Socket Mode (outbound WS to Slack) | gorilla/websocket + Microsoft Dev Tunnels | **Direct conflict** |
| Persistence | SQLite as source of truth, 7 tables, retention, migrations | In-memory `SessionState` struct only | **Direct conflict** |
| ACP implementation | Hand-ported `internal/acp` codec (spawner/handshake/reader/writer) | `coder/acp-go-sdk` third-party SDK | **Direct conflict** |
| Process topology | One server, many workspaces via `[[workspace]]` channel routing, `max_concurrent_sessions` | One isolated process per workspace, dynamically allocated ports | **Direct conflict** |
| Process manager | Standalone server + `intercom-ctl` over IPC | Spawned/port-allocated by Python `autoharness run` | **Direct conflict** |
| Auth model | Slack tokens via env/OS keychain | GitHub OAuth via Dev Tunnel identity gating | **Direct conflict** |
| Status/provenance | Frontmatter: `status: active`, dated, cites a spike decision + ADRs | No frontmatter, undated, no decision lineage | Asymmetric maturity |

Genuine overlap is limited to: the Go language, Go 1.21+, `log/slog`, ACP as the
agent protocol, the general concept of mediating agent approvals to a remote human,
and per-workspace isolation. That overlap is real but shallow — it is shared
*foundation*, not shared *architecture*.

## Options Evaluated

### Option A: Stage both documents as one covering feature

Treat them as the operator's preferred single "Go backend port" source and synthesize
one covering feature spanning both.

* **Pros:** Matches the operator's stated preference; single shipment; no deferral.
* **Cons:** Requires the plan to simultaneously assert that persistence is SQLite *and*
  in-memory; that the operator surface is Slack *and* iOS/TUI; that ACP is hand-ported
  *and* SDK-delegated. Any plan built on this either silently picks winners without a
  decision record, or produces contradictory tasks that Ship cannot execute. It would
  bake an unmade product decision into the backlog.
* **Effort:** medium
* **Fit:** Poor — violates the plan-review Scope Boundary and Architecture rubrics on its face.

### Option B: Stage the port brief only; defer the multiplexer pending reconciliation

Split the group. Recognize the port brief as the anchored, decided direction (it declares
itself authoritative, carries `status: active`, and cites a spike decision plus ADRs).
Stage its first slice. Preserve the multiplexer document in the stash with an explicit
architecture-reconciliation deliberation required before it can be planned.

* **Pros:** Every staged premise is internally consistent; respects the documented decision
  lineage; nothing is lost; the unmade product decision is surfaced to the operator as an
  explicit open question rather than silently resolved by an agent.
* **Cons:** Defers roughly half the supplied material; requires an operator decision later.
* **Effort:** low
* **Fit:** Strong.

### Option C: Stage only the strict intersection (shared Go foundation), defer both architectures

Stage only what both designs provably share — module skeleton, CI, CGO-free build — and
defer *both* architecture-specific bodies of work.

* **Pros:** Zero architectural commitment; maximum optionality.
* **Cons:** The intersection is thin. It excludes the config schema and error taxonomy,
  which are fully specified, low-risk, and unblock the port's Phase 1 — deferring them
  buys optionality the operator has not asked for and leaves the first shipment
  anaemic. Also defers work the port brief has already decided.
* **Effort:** low
* **Fit:** Moderate — over-corrects.

### Option D: Deliberate the architecture fork first, stage nothing this session

Halt staging and return the fork to the operator as a blocking question.

* **Pros:** Cleanest separation of decision from execution.
* **Cons:** Produces no backlog and no shipment this session. The foundation slice
  (skeleton, CI, config, errors) is genuinely independent of how the fork resolves —
  a Go repo needs `go.mod`, a CI pipeline, and a CGO-free build gate under *either*
  architecture. Blocking on the fork would stall work that is not actually blocked.
* **Effort:** none
* **Fit:** Poor — unnecessarily conservative.

## Trade-off Comparison

| Criterion | Option A | Option B | Option C | Option D |
|---|---|---|---|---|
| Internal consistency of staged plan | Fails | Strong | Strong | N/A |
| Respects documented decision lineage | No | Yes | Partial | Yes |
| Unblocks greenfield foundation now | Yes | Yes | Yes | No |
| Surfaces the real open decision | No (hides it) | Yes | Yes | Yes |
| Nothing lost from operator input | Yes | Yes | Yes | Yes |
| Shipment coherent as one PR | No | Yes | Yes | N/A |
| Risk of rework if fork resolves the other way | High | **Low** | Lowest | None |

The rework column is the decisive one. Under Option B the staged slice is
`go.mod` + package skeleton + CI (build/vet/staticcheck/`-race`/govulncheck/cross-compile)
+ `config.toml` loading + an error taxonomy. Of these, only the config schema shape and the
error-variant list are port-specific, and both are additive and cheap to extend. Everything
else is required verbatim by the multiplexer architecture too. So Option B's rework
exposure is close to Option C's, while delivering a materially more useful first shipment.

## Decision

**Option B.** Split the two documents into two groups and stage only Group A's first slice.

* **Group A (staged this session)** — `docs/decisions/2026-07-06-go-port-reference-brief.md`,
  stash `4989A42D`. The authoritative Go port of `agent-intercom`. This session stages its
  **first slice only**: the CGO-free foundation.
* **Group B (deferred to stash)** — `docs/design-docs/intercom-go-backend-architecture.md`,
  stash `A92E3FA0`. Retained active in the stash, annotated with the reconciliation
  requirement. It must not be planned until the operator resolves the fork.

### Rationale

1. **The documents are not one source.** Nine architectural axes conflict outright.
   Planning them together would require an agent to silently pick winners on an unmade
   product decision — precisely the failure mode the plan-review gate exists to catch.
2. **The port brief is the anchored direction.** It declares itself authoritative,
   carries `status: active`, is dated, and cites both a spike decision
   (`2026-07-06-rust-vs-go-intercom-spike.md`) and a set of ADRs. The architecture doc
   has no frontmatter, no date, no status, and no decision lineage; it reads as design
   exploration for a differently-named product.
3. **The foundation slice is fork-independent.** A CGO-free Go repo with CI gates is
   required under either architecture, so staging it now is safe regardless of how the
   fork resolves.
4. **The fork is a product decision, not an agent decision.** It is surfaced as an
   explicit unresolved question below rather than resolved by inference.

### Scope of the staged slice (Group A, slice 1)

Derived from the port brief's Section 12 Phase 1 ("Skeleton + risk slices"), reduced to
the portion that carries no architectural fork exposure and fits one coherent pull request:

* Go module and `internal/` + `cmd/` package skeleton (brief §4)
* Two binary entrypoints: `cmd/intercom`, `cmd/intercom-ctl` (brief §4)
* CGO-free build verification and 4-target cross-compilation (brief §13)
* CI gates: build, `go vet`, `staticcheck`, `go test -race`, `govulncheck` (brief §12.1)
* `config.toml` schema, loading, and validation (brief §9)
* Credential resolution — env with `*_ACP` precedence, OS keychain fallback (brief §9)
* `AppError` → Go error taxonomy, 14 variants (brief §10)

### Explicitly OUT of scope for this slice

* The three risk slices (Slack Socket Mode, ACP subprocess, `modernc.org/sqlite`) — brief §11.
  Each is a substantial spike-grade unit deserving its own shipment.
* SQLite schema, the 7 tables, and the 8 repositories — brief §6.
* `AgentDriver` interface and ACP driver — brief §5.
* Slack surface, approval/diff/policy, stall detection, steering, IPC/ctl — brief §7.
* Config hot-reload / `fsnotify` watcher — deferred to the config-watcher unit.
* Everything in Group B.

## Rejected Alternatives

* **Option A (one covering feature)** — rejected because it forces contradictory premises
  into a single plan and would silently resolve an unmade product decision. This directly
  addresses the operator's instruction to treat both as one source *unless* a concrete
  reason to split exists; nine direct architectural conflicts are that concrete reason.
* **Option C (strict intersection)** — rejected as over-correction. Excluding the config
  schema and error taxonomy defers fully-specified, low-risk, already-decided work and
  yields an anaemic first shipment, for optionality that was not requested.
* **Option D (defer everything)** — rejected because the foundation slice is genuinely
  independent of the fork; blocking it would stall unblocked work and produce no shipment.

## Unresolved Questions

1. **[BLOCKING for Group B]** Which architecture governs `intercom-go`? Options appear to be:
   (a) the Slack-mediated `agent-intercom` port only; (b) the Copilot Remote Multiplexer only;
   (c) the port first, with the multiplexer as an additional operator-surface driver behind
   the existing `AgentDriver` seam; (d) two separate products in separate repositories.
   Option (c) is architecturally plausible — the port brief's `AgentDriver` interface is
   explicitly designed as a pluggable seam — but it still requires reconciling the
   persistence model (SQLite vs in-memory) and the ACP implementation
   (hand-ported codec vs `coder/acp-go-sdk`).
2. **[Non-blocking]** Should the ACP layer use `coder/acp-go-sdk` (architecture doc) instead
   of a hand-ported codec (port brief §8)? The brief requires preserving raw JSON-RPC `id`
   types for correlation; whether the SDK preserves that must be verified before adoption.
   This should be settled during the ACP risk slice, not now.
3. **[Non-blocking]** The torn `references` submodule — is it the intended Rust oracle mount
   (brief §1.2)? Resolving it is a prerequisite for the parity-test work, not for this slice.
4. **[Non-blocking]** MCP legacy mode (brief §8) — the brief says it is being retired and
   the strategic direction is ACP-only. Confirm MCP is out of scope permanently so
   `internal/mode` and `internal/driver/mcp` can be dropped from the module map.

## Risks and Mitigations

| Risk | Impact | Mitigation |
|---|---|---|
| Fork resolves toward the multiplexer, invalidating the config schema and error taxonomy | Medium | Confined to 2 of 4 sub-epics; both are additive and cheap to extend. Skeleton and CI are unaffected. |
| `modernc.org/sqlite` fails the CGO-free or performance bar (brief §11 risk slice 3) | High for the port overall | Out of scope here; must be proven in its own shipment before the store work is planned. |
| Credential handling (keychain + env tokens) mishandles secrets | High | Triggers plan hardening; secrets must never be logged and must not enter error strings. Carried into the plan's hardening section. |
| Torn `references` submodule is disturbed | Medium | Explicitly out of bounds this session; recorded as a deferred item requiring operator action. |
| Foundation shipment grows into a large PR | Low | All files are net-new in a greenfield repo; zero regression surface. Ship advised to commit incrementally per sub-epic. |
| Group B stash entry is lost or silently planned | Medium | Entry `A92E3FA0` remains **active** (not archived) and is annotated with the blocking reconciliation requirement. |

## Provenance

| Artifact | Path | Stash |
|---|---|---|
| Port reference brief (Group A source) | `docs/decisions/2026-07-06-go-port-reference-brief.md` | `4989A42D` |
| Backend architecture doc (Group B source) | `docs/design-docs/intercom-go-backend-architecture.md` | `A92E3FA0` |
| Implementation plan (this decision → plan) | `docs/plans/2026-09-03-intercom-go-foundation-plan.md` | — |

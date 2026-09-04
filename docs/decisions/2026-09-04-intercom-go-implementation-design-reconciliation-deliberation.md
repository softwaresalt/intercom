---
title: "Reconciliation: operator implementation-design doc vs. the governing corrected architecture"
description: "Section-by-section reconciliation of docs/design-docs/intercom-architecture-implementation-design.md against governing decision D0-D9a, design revision 2, and stash A92E3FA0; determines the first coherent implementation round (C2)"
topic: "Reconciling the untracked operator implementation-design document with the 2026-09-04 corrected product direction"
depth: "deep"
decision_status: "decided"
promoted_to: "plan"
linked_artifacts:
  - "docs/design-docs/intercom-architecture-implementation-design.md"
  - "docs/design-docs/intercom-go-backend-architecture.md"
  - "docs/decisions/2026-09-04-intercom-go-architecture-correction-copilot-sdk-deliberation.md"
  - "docs/plans/2026-09-04-intercom-go-architecture-correction-plan.md"
stash_refs:
  - "A92E3FA0"
  - "4989A42D"
  - "8C2D578D"
  - "8ACF7110"
tags:
  - "architecture"
  - "reconciliation"
  - "copilot-sdk"
  - "governing-decision"
  - "stage"
---

# Reconciliation: operator implementation-design doc vs. governing corrected architecture

## Framing

On 2026-09-04 the operator placed an untracked document at
`docs/design-docs/intercom-architecture-implementation-design.md`
("Intercom: Agent Control Plane (ACP) Architecture Design", hereafter
**IMPL-DESIGN**) and directed Stage to treat it as a governing
architecture/design input for the first staging round.

IMPL-DESIGN is **not** the design document the current governing stash entry
cites. Stash `A92E3FA0` (high, governing product direction) names three
governing artifacts, and IMPL-DESIGN is none of them:

| Governing artifact per `A92E3FA0` | Status |
|---|---|
| `docs/decisions/2026-09-04-...-copilot-sdk-deliberation.md` | tracked, `decision_status: decided` |
| `docs/design-docs/intercom-go-backend-architecture.md` (rev 2) | tracked |
| `docs/plans/2026-09-04-...-architecture-correction-plan.md` (C1–C11) | tracked, C1 shipped |

So there are now **two design documents** describing the same product, written
from different premises. The question this deliberation answers is not "which
document do we like" but:

> **Which parts of IMPL-DESIGN are adoptable under the corrected direction,
> which conflict with it, and what is the first coherent implementation round?**

`A92E3FA0` is explicit that the corrected direction is **governing**, and the
operator directive repeats it: *"Treat the corrected direction as governing
where explicitly stated; surface any unresolved architectural conflict rather
than silently choosing."* That instruction is decisive for every conflict below
where the corrected direction speaks explicitly, and it is equally decisive in
the other direction: where the corrected direction is **silent**, IMPL-DESIGN
is not overridden — it is *new material* requiring its own decision, not
automatic rejection.

### Verified ground truth (checked, not assumed)

* `go.mod`: module at `go 1.24`, toolchain `go1.26.5`. `copilot-sdk/go` is
  **not yet a dependency** — only `BurntSushi/toml` and `spf13/cobra`.
* Shipped Go surface is small: `cmd/intercom`, `cmd/intercom-ctl`,
  `internal/apperr`, `internal/config`, `internal/pathsafe`. Both `cmd`
  binaries still return a `not implemented` sentinel.
* Backlog is a clean slate: 0 active/queued shipments, 0 checkpoints; all
  001-F–005-F / 001-S–004-S artifacts archived.
* `scripts/check-retired-architecture.sh` scans **only** `internal/config/**`
  (excluding `_test.go` and `testdata/`), `config.toml.example`, and
  `cmd/**/*.go`. Forbidden tokens include `acp`, `ipc_name`, `slack`,
  `host_cli`, `channel_id`, `team_id`, `socketmode`. **Docs are not scanned.**

---

## Section-by-section reconciliation

Every IMPL-DESIGN section receives one of four dispositions:
**ADOPT** / **ADOPT-WITH-AMENDMENT** / **DEFER** / **REJECT**.

### Aligned — ADOPT

| IMPL-DESIGN | Disposition | Basis |
|---|---|---|
| Tenet 1 — Native SDK hosting ("imports and executes the Copilot SDK in Go", does not wrap a CLI) | **ADOPT** | Exactly restates `A92E3FA0` correction (2) and design rev 2 §7. |
| Tenet 3 — State projection, not mutation; UIs are read-only projections of a central event bus | **ADOPT** | Restates D4 ("the projection is a **cache, not the system of record**") and design rev 2 §5. |
| §6.1 Bubble Tea TUI — Elm architecture (`tea.Model`/`Update`/`View`), `lipgloss` 2D layout, intervention modals | **ADOPT** | Consistent with design rev 2 §2 stack (`bubbletea` + `lipgloss` + `bubbles/viewport`). IMPL-DESIGN adds genuinely useful layout detail rev 2 lacks. |
| §7 step 8 — Central event bus: unidirectional fan-out, strict payload typing, buffered channels, `context.Context` cancellation, timeout thresholds so a dropped mobile connection never stalls the SDK loop | **ADOPT** | This is the strongest contribution in the document. It is materially consistent with design rev 2 §4 (concurrency), §5.2 (event envelope), §5.5 (WebSocket backpressure), and sharpens them. |

**Finding:** IMPL-DESIGN's event-bus and TUI material is additive and should be
carried forward into design rev 3 when C5–C11 are planned. It is *not* first-round
scope, but it should not be lost.

### Conflicts — REJECT or ADOPT-WITH-AMENDMENT

#### RC-1 — "Agent Control Plane (ACP)" naming — **REJECT the term** (high severity)

IMPL-DESIGN titles the system *"Agent Control Plane (**ACP**)"*. In this
repository `ACP` is a **retired** term meaning *Agent Client Protocol* — the
custom broker retired by D0.1/D0.2. `A92E3FA0` states intercom-go "implements
no ACP broker, no ACP wire protocol". Design rev 2 §9 lists "Any ACP protocol
implementation" as a Non-Goal, and `acp` is a **forbidden token** in the
anti-regression gate.

* The doc itself does **not** trip CI (docs are out of gate scope — verified).
* But adopting "ACP" as the product term guarantees a future collision: the
  moment a Go identifier or TOML key in `cmd/**` or `internal/config/**` is
  named for the control plane, `check-retired-architecture.sh` hard-fails, and
  a reviewer cannot distinguish "retired protocol leaked back" from "new
  control-plane name".
* Stash `8C2D578D` already tracks `KindACP` as retired-architecture
  contamination that must be *removed*. Re-minting `ACP` with a new expansion
  while simultaneously deleting it as contamination is incoherent.

**Decision:** reject the acronym. The concept ("control plane") is fine; the
token is burned. Do not rename anything in shipped code as part of this
decision — that is `8C2D578D`'s slice.

#### RC-2 — Sidecar process supervision over Named Pipes — **REJECT for now / DEFER concept** (high severity)

IMPL-DESIGN §2 and Phase 1 make intercom the **master process manager**: it
spawns `agent-engram` (Rust), `graphtor-docs` (Rust), and `backlogit` (Go) as
long-lived children, talks to them over Windows Named Pipes, health-checks,
buffers, respawns, and multiplexes JSON-RPC.

This conflicts with the governing topology on three independent axes:

1. **Supervision direction is inverted.** Design rev 2 §1.2/§3 and D0.6: *one
   isolated Go process per workspace, **spawned by autoharness***, with
   autoharness owning port allocation and tunnel registration. IMPL-DESIGN has
   intercom supervising a process tree instead. Both cannot be true.
2. **Scope explosion beyond this repository.** Phase 1 step 1 requires
   *"Refactor Rust/Go Sidecars: update agent-engram, graphtor-docs, and
   backlogit CLI entry points"* — modifying **three external repositories**,
   two of them Rust. Nothing in the governing set authorizes cross-repo work,
   and `4989A42D` records that Rust-side work is retired.
3. **Platform regression.** Windows Named Pipes (`\\.\pipe\...`, `go-winio`)
   are Windows-only. The repository ships `start.sh`, `scripts/*.sh`, and a
   bash CI gate — a cross-platform posture. IMPL-DESIGN would silently make the
   IPC layer Windows-bound.

Additionally `ipc_name` is a forbidden gate token, so an IPC config surface
would trip CI in `internal/config/**`.

**Decision:** do not adopt sidecar supervision in the first round. The
*underlying need* (intercom reaching engram/backlogit) is real but **unscoped**
— the governing set never decided how intercom talks to workspace tooling. Log
as open question **Q7** for a future deliberation; do not silently choose a
transport now.

#### RC-3 — React / iOS / VAPID re-introduction — **REJECT** (high severity, clearest case)

IMPL-DESIGN §6.2 specifies a **React** PWA, **iOS** VoiceOver semantics, **iOS
lock-screen** notifications, VAPID keys via `webpush-go`, and a Service Worker.

`A92E3FA0` correction (1) is verbatim on this point:

> *"the remote client is an SPA, **NOT iOS** — the iOS wording is stale; SPA
> framework choice is deliberately **OUT OF SCOPE**"*

Design rev 2 §9 Non-Goals: *"SPA framework selection or SPA implementation."*
Q2 in the governing deliberation: *"SPA framework — deliberately out of scope
per the operator."*

IMPL-DESIGN reintroduces **precisely the stale wording the correction
retired**, plus a framework choice the operator explicitly withheld. This is
the single case where the operator directive ("treat the corrected direction as
governing where explicitly stated") applies with no ambiguity.

**Decision:** reject. Remote surface remains "an SPA over a Dev Tunnel",
framework unspecified. Web Push / VAPID is a *notification transport* question
that may return later on its own merits — but not as an iOS requirement, and
not in the first round.

#### RC-4 — "Managing raw token streams" — **ADOPT-WITH-AMENDMENT** (medium)

IMPL-DESIGN §1 tenet 1 says intercom manages *"raw token streams and tool
execution loops"*, and §1 frames intercom as a **MITM sidecar** that *"replaces
the ephemeral execution model of the standard copilot.exe CLI"*.

The *conclusion* (native SDK hosting) is right, but the *mechanism* is not
available: finding F2 in the governing deliberation established that the SDK
seals its protocol layer under `github.com/github/copilot-sdk/go/internal/…`
(`internal/jsonrpc2`), which is **unimportable by hosts**. A host cannot
intercept raw tokens or man-in-the-middle the SDK's own transport. `A92E3FA0`
records that the SDK internally managing a Copilot CLI server *"is an SDK
IMPLEMENTATION DETAIL"* and that intercom-go has *"no direct headless-Copilot-CLI
lifecycle"*.

**Decision:** adopt the intent, amend the language. intercom-go consumes the
SDK's **public** event/permission/session API and owns an anti-corruption layer
(D3). Drop "MITM" and "raw token streams". This is exactly what spike criteria
S1–S3 exist to prove.

#### RC-5 — Clean-room ephemeral `_stage`/`_ship` delegation engine — **DEFER** (high severity)

IMPL-DESIGN §3 has a long-running `_orchestrator` session that delegates to
*fresh headless SDK sessions loaded with `_stage.agent.md`*, pausing the
orchestrator's channel and calling `deleteSession()`.

This makes intercom-go an **agent-orchestration product** — which is
autoharness's role, not the brokering backend's. Governing scope in `A92E3FA0`
is narrower and specific: *"a per-workspace Go backend brokering a GitHub
Copilot agent session between two co-equal operator surfaces."*

It also **presumes an unresolved decision**. Q6 ("whether one process hosts >1
concurrent Copilot session") is explicitly open and *not scheduled until C9*;
D4a deliberately made the session-pointer store *"a set keyed by `session_id`"*
precisely so the Q6 answer could not invalidate C4. IMPL-DESIGN assumes
multi-session as a *foundation*.

**Decision:** defer entirely. Do not treat as rejected — it is a coherent
long-range idea — but it cannot be designed before Q6 is answered at C9.

#### RC-6 — Fail-closed circuit breaker with `git reset` + stash writes — **DEFER with a flagged risk** (high severity)

IMPL-DESIGN §4.2 has intercom **reset the worktree to the last stable git
checkpoint** and, on a 15-minute TTL, **move the item to the backlogit stash**.

Two distinct concerns:

1. **Destructive repository mutation.** An automated `git reset` of an operator's
   worktree is the highest-blast-radius action in the entire document. The
   working tree is *currently dirty with unrelated operator work* — this
   session verified modified `.gitignore`, `start.ps1`, `.backlogit/stash.jsonl`
   and untracked `.claude/`, `.github/copilot/`, and IMPL-DESIGN itself. A
   naive checkpoint-reset would destroy exactly this kind of work.
2. **Backlog write authority.** Automated stash mutation collides with the
   Stage/Ship policy model (P-010 role boundaries, P-021 capture rules), where
   stash writes are agent-role-scoped.

The *convergence-vector* analysis (§4.1: error topology, blast radius,
oscillation detection) is intellectually valuable and worth keeping.

**Decision:** defer. If revived, the reset mechanism must be re-founded on
something non-destructive (a dedicated worktree or branch), never on the
operator's checkout. Record as risk **R8**.

#### RC-7 — `.intercom/` telemetry and session state — **ADOPT-WITH-AMENDMENT** (medium)

IMPL-DESIGN §5 proposes `.intercom/session.json` and an append-only
`.intercom/events.jsonl`, with a correlation schema of
`workspace_id` + `task_id` + `session_id` + `turn_id`.

* The **sparse event log and the correlation schema are good** and compatible
  with D4 (SDK runtime is the durable event source; local state is a cache).
  Adopt in principle.
* `.intercom/session.json` as described is **over-scoped**. D4a fixed the
  durable pointer's content exactly — *"the current `session_id`, its workspace
  path, and the SDK/CLI protocol version… **Nothing else**"* — and explicitly
  held D7 (real persistence) deferred. IMPL-DESIGN adds "UI session bindings
  and iOS VAPID push subscriptions", which both re-opens D7 and re-imports
  RC-3.
* IMPL-DESIGN's §5.1 statement that `.copilot/session-state/<session_id>/` is
  SDK-managed **agrees** with D4 and is a useful corroboration.

**Decision:** adopt the event log + correlation schema as input to C3/C4
planning; reject the expanded `session.json` payload. D4a's pointer stands
unchanged. The atomic-write discipline D4a already mandates
(write-temp-then-rename + fsync) governs.

---

## Conflict summary

| ID | Conflict | Severity | Disposition |
|---|---|---|---|
| RC-1 | "ACP" acronym re-minted against a retired, gate-forbidden token | high | **REJECT term** |
| RC-2 | Sidecar supervision + Named Pipes inverts topology, spans 3 repos, Windows-only | high | **REJECT now / DEFER as Q7** |
| RC-3 | React + iOS + VAPID contradicts the explicit correction | high | **REJECT** |
| RC-4 | "MITM"/"raw token streams" not achievable (SDK `internal/`) | medium | **AMEND** |
| RC-5 | Clean-room `_stage`/`_ship` engine presumes unresolved Q6 | high | **DEFER to ≥ C9** |
| RC-6 | `git reset` of operator worktree + automated stash writes | high | **DEFER, risk R8** |
| RC-7 | `.intercom/session.json` over-scopes D4a and re-opens D7 | medium | **AMEND** |

**Net assessment.** IMPL-DESIGN is a *forward-looking whole-product vision*
written largely independently of the 2026-09-04 correction. Roughly a third of
it (native SDK hosting, projection semantics, the event bus, the TUI) is
directly adoptable and genuinely improves on design rev 2. The remainder either
contradicts explicit corrections (RC-1/RC-3), inverts the decided topology
(RC-2), or front-runs decisions deliberately deferred (RC-5/RC-6/RC-7).

It is therefore **not** a drop-in replacement for design rev 2, and it must not
be silently promoted to governing status. It is best classified as a
**candidate vision document** whose adoptable content is merged into design
rev 3 incrementally, phase by phase.

---

## Internal inconsistency found in the governing set (must be resolved)

`A92E3FA0` lists as an open blocker:

> *"Q1/H2 Copilot CLI version floor (**blocks C2**)"*

But the governing deliberation's Q1 disposition says the opposite:

> *"Unknown — not documented… **The spike (D2) must record the CLI version it
> validated against, and that becomes the provisional floor.**"*

These cannot both hold: Q1 is an **output** of the C2 spike, not an input to
it. If Q1 blocked C2, Q1 could never be answered — the only mechanism the
governing set defines for answering it is the C2 spike itself. D9a reinforces
this by defining C2 as precisely *"SDK proving spike: adds the pinned
dependency with its first real import, and proves S1/S2/S3."*

**Resolution (recommended, needs operator confirmation):** treat the
`A92E3FA0` annotation as a stale transcription. Q1 is **spike-resolved**, not
spike-blocking. C2 proceeds; the spike records the validated Copilot CLI
version and that becomes the provisional floor.

This is recorded as **decision D10** below and flagged as the round's one
genuine operator decision.

---

## Options for the first implementation round

### Option A — C2 SDK proving spike only

Follows D9a literally. **Rejected**: fails the operator's explicit directive
that IMPL-DESIGN be *"in the first-round deliberation/planning scope, not
merely a reference appended after planning."*

### Option B — Design-doc reconciliation only

Governance-only round; defers C2. **Rejected**: leaves the repository with zero
forward implementation motion and no SDK validation, while the reconciliation
work alone is too thin to justify a shipment cycle.

### Option C — Reconciliation governance **then** C2 SDK proving spike (RECOMMENDED)

One bounded shipment, two sequential sub-epics:

* **Sub-epic A (docs/governance)** — persist this reconciliation, mark
  IMPL-DESIGN's status in-place without altering operator content, amend design
  rev 2 with the terminology retirement and the deferred-item register.
* **Sub-epic B (Go spike)** — C2 exactly as D9a scopes it: pinned dependency +
  first real import, prove S1/S2/S3, emit a findings artifact recording the
  validated CLI version.

**Why A gates B:** reconciliation determines whether C2's scope is still
correct. Had IMPL-DESIGN been adopted wholesale, the first round would have
been an IPC supervisor, not an SDK spike. Settling direction before spending
the spike is the cheaper ordering, and it gives one coherent review context
("settle direction, then de-risk the SDK") rather than two.

**Selected.**

### Option D — Adopt IMPL-DESIGN and start its Phase 1 (IPC supervisor)

**Rejected** on four independent grounds: requires modifying three external
repositories (two Rust) with no authorization; inverts the decided supervision
topology (RC-2); is Windows-only against a cross-platform posture; and would
introduce `ipc_name`-shaped surfaces that trip the anti-regression gate. It also
skips SDK validation entirely, leaving R1/R3/R5 unmitigated.

---

## Decisions

* **D10 — Q1 is spike-resolved, not spike-blocking.** C2 is unblocked. The
  spike records the validated Copilot CLI version as the provisional floor.
  *(Operator confirmation requested; this is the round's only blocking
  decision.)*
* **D11 — IMPL-DESIGN is a candidate vision document, not governing.** Design
  rev 2 + the correction deliberation remain governing. IMPL-DESIGN's adoptable
  content merges into rev 3 incrementally. The file is **preserved unmodified
  in content**; only a status/provenance header is added.
* **D12 — "ACP" is retired as a product term in all senses.** Neither *Agent
  Client Protocol* nor *Agent Control Plane*. Use "control plane" in prose if
  needed; never the acronym, and never as a code or config identifier.
* **D13 — The first round is Option C**: reconciliation governance gating a
  C2 SDK proving spike, as one shipment.
* **D14 — Deferred-item register.** RC-2 (→ new **Q7**: how intercom reaches
  workspace tooling), RC-5 (→ Q6 at C9), RC-6 (→ risk **R8**), and the adoptable
  event-bus/TUI material are recorded in design rev 2 so they are retrievable
  rather than lost with an untracked file.

## Open questions carried forward

| # | Question | Blocks |
|---|---|---|
| Q1 | Copilot CLI version floor | **Resolved by C2 per D10** |
| Q3/H3 | Dev Tunnel authentication model | C10 |
| Q6/H4 | Sessions per process | C9 (and RC-5) |
| H5 | Permission authorization model | C5 |
| H6 | Session-pointer store semantics | C4 |
| **Q7 (new)** | How intercom reaches workspace tooling (engram/backlogit) — transport undecided | future; raised by RC-2 |

## Risks

| # | Risk | Severity | Mitigation |
|---|---|---|---|
| R1 | SDK experimental permission API changes | high | D3 anti-corruption layer; spike S1 |
| R3 | Silent event-variant drop | high | mandatory `default:` branch; spike S2 |
| R4 | SDK/CLI version skew | high | spike records validated CLI version (D10) |
| R5 | Cancellation/shutdown deadlock | high | spike S3 (`Abort` → `Stop` → `ForceStop`) |
| **R8 (new)** | Automated `git reset` would destroy operator work | high | RC-6 deferred; if revived, must use a dedicated worktree, never the operator checkout |
| R9 (new) | IMPL-DESIGN silently treated as governing by a later session | medium | D11 + in-file status header (task A2) |

## Context consulted

* `docs/design-docs/intercom-architecture-implementation-design.md` (full)
* `docs/design-docs/intercom-go-backend-architecture.md` §§1.2, 2, 3, 9
* `docs/decisions/2026-09-04-...-copilot-sdk-deliberation.md` §§D2, D3, D4, D4a, D8, D9, D9a, Q1–Q6, risks
* `docs/plans/2026-09-04-...-architecture-correction-plan.md` (headings, non-goals)
* `docs/closure/004-S-005-F-post-merge-closure.md` (PR #15 provenance)
* Stash `A92E3FA0`, `4989A42D`, `8C2D578D`, `8ACF7110`, `EF9352FB`
* `go.mod`, `scripts/check-retired-architecture.sh`, repository Go inventory

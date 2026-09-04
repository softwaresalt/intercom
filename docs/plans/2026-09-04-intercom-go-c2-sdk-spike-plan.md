---
title: "First round: implementation-design reconciliation + C2 Copilot SDK proving spike"
description: "Implementation plan for the first staging round — persists the IMPL-DESIGN reconciliation as governance, then executes the D9a phase-C2 SDK proving spike (S1/S2/S3 plus the four design-mandated C2 spike questions) and the §4.2 design amendment that C2 closure requires"
status: "reviewed"
source_document: "docs/decisions/2026-09-04-intercom-go-implementation-design-reconciliation-deliberation.md"
linked_artifacts:
  - "docs/decisions/2026-09-04-intercom-go-architecture-correction-copilot-sdk-deliberation.md"
  - "docs/design-docs/intercom-go-backend-architecture.md"
  - "docs/design-docs/intercom-architecture-implementation-design.md"
stash_refs:
  - "A92E3FA0"
  - "4989A42D"
tags:
  - "copilot-sdk"
  - "spike"
  - "governance"
  - "phase-c2"
---

# Plan: reconciliation governance + C2 Copilot SDK proving spike

**Source document:**
`docs/decisions/2026-09-04-intercom-go-implementation-design-reconciliation-deliberation.md`
(decision D13 selects Option C).

**Governing artifacts:**
`docs/decisions/2026-09-04-intercom-go-architecture-correction-copilot-sdk-deliberation.md`
(D2, D3, D4a, D9a, Q1, F1–F5) and
`docs/design-docs/intercom-go-backend-architecture.md` (revision 2 — §§3.1,
4.2, 4.3, 5.1, 5.4, 7.1–7.5, 8, 9). Design input reconciled this round:
`docs/design-docs/intercom-architecture-implementation-design.md` (IMPL-DESIGN).

> **Revision note (post-review remediation).** Plan-review attempt 1 returned
> FAIL. The decisive finding: design rev 2 defines **four required C2 spike
> questions** (§3.1, §4.2, §5.1(d), §5.4), one of which carries an explicit
> closure condition — *"C2 may not close until the design has been amended with
> the chosen fallback"* (§4.2). The original decomposition covered only
> S1/S2/S3 and could therefore not close C2. Tasks B5 and B7 were added, the
> B2→B4 edge was made explicit, and twelve further findings were remediated.
> Attempt 2 returned FAIL on two under-delivered remediations (an inert G6
> regex, and an UNPROVEN fallback applied to B6 only) plus one new MD041
> defect; all three are fixed. See `## Post-Review Remediation Record`.

## Problem Frame

Two problems are settled together because the first gates the second.

1. **Direction is ambiguous.** An untracked operator design document
   (IMPL-DESIGN) describes the same product from premises that conflict with
   the 2026-09-04 correction on seven points (RC-1…RC-7). Until its status is
   recorded, any later session may treat it as governing and build the wrong
   thing. Its adoptable content is also at risk of being lost, because the file
   is untracked.
2. **The SDK is unproven.** `github.com/github/copilot-sdk/go` is still not a
   dependency (`go.mod` has only `BurntSushi/toml` and `spf13/cobra`), and
   D2 forbids production wiring until three surfaces are proven against the pin
   (S1 permission round-trip, S2 event-union handling, S3 cancellation/
   shutdown). Design rev 2 adds four further **required C2 spike questions**.
   Risks R1/R3/R4/R5 are all unmitigated.

### The four design-mandated C2 spike questions (verified quotations)

| ID | Question | Source | Consequence if unanswered |
|---|---|---|---|
| **SQ-a** | Does `Session.Abort` unblock an already-blocked `PermissionHandlerFunc`? *"Until answered, assume it does not."* | §3.1 L175 | Shutdown step 0 cannot be simplified |
| **SQ-b** | Is `Session.On` callback invocation **serialised**? *"Re-entrancy is UNSPECIFIED… a **required C2 spike question**, and C2 may not close until the design has been **amended with the chosen fallback**"* | §4.2 L246–254 | **C2 cannot close.** An adverse answer invalidates the ordered fold in §5.1 |
| **SQ-c** | Is permission dispatch independent of event dispatch (head-of-line blocking behind a modal)? *"a **required C2 spike question**"* | §5.4 L442 | UI streaming behaviour behind a modal is undesignable |
| **SQ-d** | Does `SessionEvent` carry a stable identity usable as a de-duplication key? *"a **blocking gate on C4**, alongside H6"* | §5.1 L297, L306 | C4 inherits an unanswered blocking gate |

SQ-b is the strictest: it requires a **design amendment**, not merely a
recorded answer. *"'Answered empirically' is not sufficient."* That is why
task **B7** exists.

### Blast radius (verified, not assumed)

* **Go source touched:** none of the existing packages. The spike lands in a
  **new** package `internal/copilotprobe`. `internal/apperr`,
  `internal/config`, `internal/pathsafe`, `cmd/intercom`, and `cmd/intercom-ctl`
  are **not modified**.
* **`go.mod` / `go.sum`:** modified — this is the whole point of C2 per D9a.
  Language floor is already `go 1.24` (raised in 004-S/U-A1 specifically to
  satisfy the SDK's declared floor), so **no deliberate toolchain change is
  required**; B1 asserts the floor did not drift (see I4).
* **Docs touched:** `docs/design-docs/intercom-architecture-implementation-design.md`
  (header only, content preserved), `docs/design-docs/intercom-go-backend-architecture.md`
  (amendments in A2 and B7), plus two new files under `docs/decisions/` and
  this plan.
* **Anti-regression gate:** `scripts/check-retired-architecture.sh` scans only
  `internal/config/**`, `config.toml.example`, and `cmd/**/*.go` (verified by
  reading `should_scan_repo_path`). This plan touches **none** of those, so the
  gate's result cannot change — which is also why I5 needs its own check (G6).

## Requirements Trace

| Req | Source | Satisfied by |
|---|---|---|
| Record IMPL-DESIGN's status without altering operator content | D11; operator directive ("preserve it, do not overwrite or discard") | A1 |
| Retire "ACP" in both expansions | D12 / RC-1 | A2 |
| Preserve deferred + adoptable IMPL-DESIGN material against file loss | D14 / RC-2, RC-5, RC-6 | A2 |
| Add the pinned SDK dependency with a real importer | D9a | B1 |
| Prove permission round-trip | D2 S1, risk R1 | B2 |
| Prove event-union unknown-variant handling | D2 S2, risk R3 | B3 |
| Prove cancellation/shutdown ordering | D2 S3 | B4 |
| **SQ-a** Abort vs. blocked handler | design §3.1 | B4 |
| **SQ-b** callback re-entrancy | design §4.2 | B5 (probe) + **B7 (amendment)** |
| **SQ-c** dispatch independence | design §5.4 | B5 |
| **SQ-d** stable event identity | design §5.1 | B3 |
| Confirm R5 documentation drift (`SendAndWait` timeout) | F5 / design §7.5 | B3 |
| Record validated Copilot CLI version as provisional floor | Q1 / D10, risk R4 | B6 |
| Produce a findings artifact, not production wiring | D2 | B6 + Non-Goals |
| **C2 closure**: design amended with the chosen re-entrancy fallback | design §4.2 | **B7** |

## Non-Goals (explicit scope fence)

* **No production SDK wiring.** No `cmd/**` behaviour change; both binaries
  keep returning their `not implemented` sentinel.
* **No anti-corruption layer implementation.** D3's adapter is a C3+ concern.
  The probe package is explicitly disposable.
* **No SDK surfaces that design §7.4 fences off until de-flagged:**
  `InProcessConnection`, `Providers`/`Models` BYOK, `EnableCitations`,
  `SessionLimits`, `EnableMCPApps`, `ExpAssignments`. (`InProcessConnection` is
  additionally build-tag gated and would conflict with B1's importer
  constraint.)
* **No `[copilot].startup_timeout_seconds` re-introduction.** Design §8 assigns
  this to C2 *"with its first real consumer (the SDK `Client.Start` deadline)"*.
  This shipment produces **no production consumer** — the probe is disposable —
  so re-introducing the field now would ship a config surface with no reader.
  **Explicitly deferred to C3**, recorded here rather than silently dropped
  (see Decisions).
* **No IPC / sidecar supervision** (RC-2, deferred as Q7).
* **No SPA, React, iOS, VAPID, or Web Push work** (RC-3, rejected).
* **No clean-room delegation engine** (RC-5, deferred to ≥ C9).
* **No convergence circuit breaker or any automated `git` mutation** (RC-6, R8).
* **No persistence schema.** D4a's pointer is not implemented this round; D7
  stays deferred (RC-7).
* **No changes to `internal/apperr` Kind taxonomy** — stash `8C2D578D`'s slice.
* **No cross-repository work** on `agent-engram`, `graphtor-docs`, `backlogit`.

## Implementation Units

### Sub-epic A — Reconciliation governance (docs only)

Gates sub-epic B informationally (direction before spend). **A is independently
closeable** — see the A-only closure path under Human Checkpoints.

**A1 — Add a status/provenance header to IMPL-DESIGN.**
Add YAML frontmatter to
`docs/design-docs/intercom-architecture-implementation-design.md` plus a short
status banner marking it `status: candidate-vision`, `governing: false`, and
cross-linking the reconciliation decision and design rev 2.

*Banner placement (normative).* The repo's `.markdownlint.json` enables MD041
(first line must be a heading, with `MD025 front_matter_title: ""`). Inserting
a prose banner between the frontmatter and the file's existing H1
(`# **Intercom: Agent Control Plane (ACP) Architecture Design**`) would make the
first non-frontmatter line a non-heading and **fail MD041** on commit 2. The
banner MUST therefore be placed **immediately after the existing H1**, not
before it.

*Verifiable preservation procedure (the file is untracked today, so a plain
`git diff` cannot compare prose).* Two commits:
1. commit the operator's file **byte-for-byte unmodified** (`git add` the exact
   path; no reformatting, no lint autofix);
2. commit the frontmatter + banner as a second, additive-only change.

`Size: XS | Complexity: trivial`
**Acceptance:** commit 1 content hashes equal the pre-existing working-tree
file (`git hash-object` of the working file recorded before staging matches the
blob in commit 1); `git diff <c1> <c2> -- <path>` shows **additions only**,
zero deletions or modifications; banner sits after the existing H1;
markdownlint passes.

**A2 — Amend design rev 2 with terminology retirement and the deferred register.**
In `docs/design-docs/intercom-go-backend-architecture.md`: extend §9 Non-Goals
to name the rejected items (RC-2 sidecar supervision/IPC, RC-3 React/iOS/VAPID,
RC-5 delegation engine, RC-6 automated worktree reset); add a terminology note
recording D12 (the acronym "ACP" is retired in **both** expansions and must
never appear as a Go identifier or TOML key); add a "Deferred and adoptable
material" register capturing Q7, R8, and the adoptable IMPL-DESIGN event-bus /
TUI content so it survives independently of the untracked file; and **amend §8
to move `startup_timeout_seconds` re-introduction from C2 to C3**, since this
shipment produces no production consumer for it (a plan is transient — left
unamended, §8 would direct a future session to re-introduce the field in a
phase that has already passed).
`Size: S | Complexity: low`
**Acceptance:** §9 lists all four rejected items with RC references; a
terminology subsection states D12; the register names Q7, R8, and the adoptable
event-bus/TUI material with pointers to IMPL-DESIGN sections; §8 states the
C3 deferral with its rationale; markdownlint passes.

### Sub-epic B — C2 Copilot SDK proving spike

### No-CLI disposition (normative, applies to B2–B5)

Two distinct no-CLI scenarios exist and must not be conflated:

* **Determined *before* B1 starts** (no Copilot CLI is obtainable at all):
  sub-epic B is **held**, and the shipment closes via the **A-only closure
  path** under Human Checkpoints.
* **Discovered *during* B2–B5** (tests run and skip): each affected task
  **completes** with its criteria recorded **UNPROVEN plus the skip reason**;
  B6 then applies its zero-execution fallback and B7 declares C2 **not closed**.

Accordingly, every B2–B5 acceptance criterion below reads as *"…answered, **or**
recorded UNPROVEN with the skip reason"*. A criterion may never be reported as
proven on a skipped test.

**B1 — Add the pinned SDK dependency with its first real import.**
`go get github.com/github/copilot-sdk/go@v1.0.11` (tag `go/v1.0.11`, commit
`a550258d5c37bd662197536992a23d633bfe5804`), and create package
`internal/copilotprobe` whose non-test file constructs a client via the SDK's
public API so the module has a genuine importer. B1 also creates the **shared
turn/streaming fixture** (a helper that drives one turn and exposes the event
callback) consumed by B3 and B5, and creates the **findings artifact skeleton**
`docs/decisions/2026-09-04-intercom-go-c2-sdk-spike-findings.md` with one
placeholder section per criterion (S1–S3, SQ-a–SQ-d), so B2 and B3 — which run
in parallel — write into a file that already exists rather than racing to
create it.

*Design constraint:* use an **ordinary, non-build-tagged** file. Rationale is
D9a's actual one — `go mod tidy` prunes a module with **no importer at all**.
(Note: tidy *does* evaluate ordinary build tags, so a tagged importer would not
in fact be pruned; the untagged file is chosen for simplicity and because
`InProcessConnection`-style tag gating is fenced out by Non-Goals, not because
tidy would prune it.)
`Size: M | Complexity: medium`
**Acceptance:** `go.mod` requires the SDK at exactly `v1.0.11`; the `go`
directive is **still `1.24`** after `go get` (assert explicitly — `go get` can
raise it from a transitive floor, I4); the shared turn/streaming fixture and
the findings-artifact skeleton both exist; `go build ./...` succeeds;
`go vet ./...` passes; G2 tidy check passes as defined below.

**B2 — Prove S1: permission round-trip, and build the shared permission harness.**
In `internal/copilotprobe`, drive a real `PermissionRequestShell` to a
`PermissionHandlerFunc` and exercise **each** `rpc.PermissionDecision` variant,
recording which are accepted. This task **owns the reusable permission harness**
that B4 and B5 consume. Written as a Go test that `t.Skip`s cleanly when the
Copilot CLI or credentials are absent, so CI stays green (CI has neither).
`Size: M | Complexity: high — de-risking is the task's purpose (D2 spike);
a negative result is a valid, informative outcome`
**Acceptance:** every `rpc.PermissionDecision` variant present at the pin is
exercised and its acceptance/rejection recorded — **or recorded UNPROVEN with
the skip reason** (the variant set is enumerated from the pinned source as the
task's first step, and if it exceeds 4 the task is split before proceeding);
the harness is exported within the package for B4/B5 reuse; observed API shape
written to the findings artifact created in B1; skips with a clear reason when
the CLI is unavailable.

**B3 — Prove S2 (event-union handling), SQ-d (stable identity), and R5 drift.**
Enumerate the `SessionEvent.Data` variants **observed during one probe
session** (not the full generated union — cardinality is bounded deliberately)
and prove unknown-variant observability via a mandatory `default:` branch
logging `event.Type()`.

*Mechanism note (F3 / design §4.3):* `SessionEvent.Data` is a **sealed**
discriminated union with unexported methods, so a host package **cannot**
construct a synthetic variant. The achievable proof is therefore: a variant
**deliberately omitted** from the switch demonstrably reaches `default:`, plus
zero-valued/nil `Data` handling. Record which mechanism was used.

Also record **SQ-d**: whether `SessionEvent` exposes a stable identity usable
as a de-duplication key (design §5.1 predicts it does not), and confirm **R5**
(design §7.5 / F5): `Session.SendAndWait`'s doc comment describes a `timeout`
parameter that does not exist in the signature.
`Size: M | Complexity: medium`
**Acceptance:** observed variants enumerated for one session; an omitted
variant demonstrably reaches `default:` and is logged with its type; SQ-d
answered yes/no with the evidence (field or absence); R5 drift confirmed or
refuted against the pinned signature — **each of these recorded UNPROVEN with
the skip reason if the CLI is unavailable**; the instrumented event-callback
hook is exported within the package for B5's SQ-b probe; skips cleanly without
a CLI.

**B4 — Prove S3 (cancellation/shutdown) and SQ-a.**
*Depends on B2* — reuses B2's permission harness to create the blocked-handler
condition. Exercise `Abort` mid-turn, then `Stop`, then `ForceStop` as the
escape hatch, all under a deadline-bearing `context.Context`. Probe the §3.1
deadlock ordering and answer **SQ-a**: does `Session.Abort` unblock an
**already-blocked** `PermissionHandlerFunc`?
`Size: M | Complexity: high — de-risking is the task's purpose (D2 spike);
discovering a deadlock here is a successful outcome`
**Acceptance:** all three shutdown calls exercised under a deadline context;
SQ-a answered yes/no with evidence; observed ordering/deadlock behaviour
recorded — **each recorded UNPROVEN with the skip reason if the CLI is
unavailable**; no test hangs unbounded (hard deadline enforced); skips cleanly
without a CLI.

**B5 — Answer SQ-b (callback re-entrancy) and SQ-c (dispatch independence).**
*Depends on B2 and B3.* Two concurrency probes:
* **SQ-b:** instrument the `Session.On` callback with a concurrency detector
  (atomic in-flight counter plus a recorded max) across a token-streaming turn
  to determine empirically whether invocation is serialised.
* **SQ-c:** hold a `PermissionHandlerFunc` blocked (B2's harness) while a turn
  streams, and observe whether session events continue to arrive — i.e. whether
  permission dispatch is independent of event dispatch, or head-of-line blocks.

`Size: M | Complexity: high — these are the two questions gating C2 closure and
C4; empirical determination is the only available method (design §4.2)`
**Acceptance:** SQ-b answered with the observed maximum concurrent callback
count and the number of turns sampled; SQ-c answered with whether events
continued during a blocked handler — **each recorded UNPROVEN with the skip
reason if the CLI is unavailable**; both recorded in the findings artifact;
skips cleanly without a CLI; hard deadline enforced.

**B6 — Complete the spike findings artifact and record the CLI version floor.**
Complete `docs/decisions/2026-09-04-intercom-go-c2-sdk-spike-findings.md`
(created in B2) recording S1/S2/S3 and SQ-a…SQ-d outcomes, the **exact Copilot
CLI version validated against** (provisional floor, resolving Q1 per D10), the
environment that ran the proofs, documentation drift found, and an explicit
go/no-go recommendation for C3.

*Zero-execution fallback (mandatory).* If no Copilot CLI was available and
criteria were skipped, each such criterion MUST be recorded as **UNPROVEN**,
never as passed; the CLI-version field records `not validated — no CLI
available`; Q1 remains **open** rather than resolved; and the go/no-go
recommendation must be **no-go for C3** on that basis. A skipped criterion
reported as proven is a defect.
`Size: S | Complexity: low`
**Acceptance:** every criterion (S1–S3, SQ-a–SQ-d) carries an explicit
PROVEN / REFUTED / UNPROVEN status plus the executing environment; CLI version
stated exactly or the fallback text used; Q1 marked resolved-provisionally with
its value, or explicitly left open; go/no-go for C3 stated and consistent with
the UNPROVEN set; artifact cites this plan and the governing deliberation.

**B7 — Amend design rev 2 with the chosen re-entrancy fallback (C2 closure condition).**
Design §4.2 states C2 *"may not close until the design has been **amended with
the chosen fallback**"* for SQ-b. Using B5's answer, amend
`docs/design-docs/intercom-go-backend-architecture.md` §4.2 (and §5.1 if SQ-d's
answer changes the specified de-duplication branch) to record the determined
behaviour and the **normative** adapter rule that follows from it.

If SQ-b is UNPROVEN (no CLI), the amendment instead records that the question
remains open and that the §4.2 conservative default stands — *"the adapter must
be safe under concurrent invocation and must not rely on callback ordering"* —
and C2 is explicitly declared **not closed**.
`Size: S | Complexity: low`
**Acceptance:** §4.2 states the determined re-entrancy behaviour (or records it
as open) and the resulting normative adapter rule; if SQ-d was answered, §5.1's
two-branch de-duplication spec is collapsed to the applicable branch; the
amendment cites the findings artifact; C2 closure status stated explicitly;
markdownlint passes.

## Dependency Graph

```text
A1 ─┐
    ├─(informational)─> B1 ─┬─> B2 ─┬─> B4 ─┐
A2 ─┘                       │       │       │
                            └─> B3 ─┴─> B5 ─┴─> B6 ─> B7
```

* A1, A2 are independent of each other; both informationally gate B1.
* **B2 → B4** (B4 reuses B2's permission harness to block a handler).
* **B2, B3 → B5** (SQ-c needs the permission harness; SQ-b needs event streaming).
* B6 consumes all proofs; **B7 consumes B5's SQ-b answer** and is the C2
  closure condition.

## Deterministic Gates and Rollback Points

| Gate | Command | Rollback point |
|---|---|---|
| G1 — docs lint | `scripts/pre-commit-markdownlint.sh` | after A2; revert docs only |
| G2 — tidy-clean **after** the dependency commit | see `G2` in the gate-commands block below (must run *after* the commit — run before it, the diff **is** the intended change and would always fail) | after B1; `git revert` of the dep commit restores a clean two-dependency module |
| G3 — build/vet/test | `go build ./... && go vet ./... && go test ./...` | after B5 |
| G4 — retired-architecture gate | `scripts/check-retired-architecture.sh --self-test` | unchanged by this plan; must still pass |
| **G6 — I5 scoped token scan** | see `G6` in the gate-commands block below (the CI gate does **not** cover this path — verified) | after B5 |
| G7 — `go` directive unchanged | see `G7` in the gate-commands block below | after B1 |
| G5 — findings complete | manual review of B6 against S1–S3 + SQ-a–SQ-d, including UNPROVEN handling | after B6 |
| G8 — C2 closure | B7 amendment present and closure status stated | after B7 |

Four independent rollback points: docs (A), dependency (B1), proofs (B2–B5),
findings/amendment (B6–B7).

### Gate commands (authoritative — pipes are unambiguous here)

Markdown tables require `|` to be escaped, which previously made G6 read as an
ERE *escaped* pipe (a literal, never-matching string). The pipe-bearing gates
are therefore defined here, not in the table above. **Copy from this block.**

```bash
# G2 — tidy-clean, run with B1 already committed as HEAD
go mod tidy && git diff --exit-code HEAD -- go.mod go.sum

# G6 — I5 scoped token scan over the new probe package.
# Exits 0 only when NO retired-architecture token is present.
#   -r  covers untracked files (git grep would not)
#   -i with _? catches camelCase spellings (hostCLI, channelID, ipcName),
#      mirroring the CI scanner's component-splitting behaviour
! grep -rEiq '(\bacp\b|ipc_?name|\bslack\b|host_?cli|channel_?id|team_?id|socket_?mode)' internal/copilotprobe

# G7 — Go language floor did not drift
test "$(go list -m -f '{{.GoVersion}}')" = "1.24"
```

Verified during planning: the G6 expression exits 0 on a clean fixture and
exits 1 on a fixture containing `hostCLI` and `channelID`.

## Decisions and Rationale

* **Probe package, not adapter package.** D2 says the spike "does not produce
  production wiring". Naming it `internal/copilotprobe` (not `copilotadapter`)
  keeps it obviously disposable and prevents it from silently becoming D3's
  anti-corruption layer without review.
* **Untagged importer.** D9a's real rationale is that tidy prunes modules with
  no importer at all. The earlier claim that tidy prunes *build-tagged*
  importers was **incorrect** and has been removed (tidy evaluates all tags
  except `ignore`).
* **`startup_timeout_seconds` deferred to C3.** Design §8 assigns it to C2
  "with its first real consumer". The probe is disposable and is not that
  consumer; shipping the field now would create a config surface no code reads,
  which is exactly the coupling C1 removed. Recorded as an explicit,
  reviewable deviation rather than a silent omission.
* **Tests that skip, not fail, without a CLI.** CI has no Copilot CLI and no
  credentials. A spike that fails CI would block on an environmental condition
  rather than a code defect — but B6's zero-execution fallback ensures skipping
  can never be mistaken for proof.
* **A gates B informationally, and A can close alone.** Reconciliation is cheap
  and determines whether C2 is still the right slice. Because HC-1 is a
  blocking operator decision sitting between the sub-epics, an A-only closure
  path is defined so the shipment cannot strand.

## Risks and Caveats

| Risk | Severity | Mitigation |
|---|---|---|
| R1 — experimental permission API | high | B2 proves it before commitment; D3 deferred to C3 |
| R3 — silent event-variant drop | high | B3 requires a demonstrable `default:` branch |
| R4 — SDK/CLI version skew; module pin does not pin the CLI | high | B6 records the exact validated CLI version (D10), or `not validated` + no-go |
| **R5 — documentation drift** (`SendAndWait` doc describes a nonexistent `timeout`; per F5 and design §7.5 this is R5's actual definition — **this supersedes the source deliberation's risk table, which still labels R5 "cancellation/shutdown deadlock"**) | medium | B3 confirms against the pinned signature |
| **R5b — cancellation/shutdown deadlock** (distinct from R5; was mislabelled R5 in attempt 1) | high | B4 under a hard deadline; no unbounded hang permitted |
| SQ-b adverse answer invalidates the §5.1 ordered fold | high | B5 answers it; B7 amends the design before C2 closes |
| Spike cannot run in CI (no CLI/credentials) | medium | Tests skip cleanly; B6 zero-execution fallback forces UNPROVEN + no-go |
| `go get` silently raises the `go` directive from a transitive floor | medium | B1 acceptance + G7 assert `1.24` |
| I5 unenforced by the CI gate over the new package | medium | **G6** adds a scoped scan the CI gate does not cover |
| Operator content in IMPL-DESIGN altered by A1 | medium | Two-commit procedure + hash comparison |
| Q1 assumed spike-resolved but operator disagrees | medium | HC-1, confirm-by-exception (see below) |

## Plan Hardening Signals (REQUIRED)

**Requires plan hardening: yes.**

Signals: a **new external dependency** entering a two-dependency module; **three
tasks assessed `complexity: high`**; an operator decision (D10) affecting
sequencing; **modification of an untracked operator file**; and a **governing
closure condition** (§4.2) that gates the phase.

## Plan Hardening

### PA-1 — Adding `copilot-sdk/go` to a minimal module (ActionRisk: medium)

The module has 2 direct + 2 indirect dependencies. The SDK pulls a transitive
tree including `coder/websocket`.

*Controls:* pin to the exact commit `a550258d5c37bd662197536992a23d633bfe5804`
via tag `go/v1.0.11`; G2 enforces tidy-cleanliness **after** the commit; G7
asserts the `go` directive did not drift; the dependency lands in its own
commit so `git revert` is a single-commit rollback; no existing package imports
it, so removal cannot break shipped code.

### PA-2 — Modifying an untracked operator file (ActionRisk: medium)

A1 edits a file the operator created and has not committed. Overwriting or
reformatting it would destroy unversioned operator work.

*Controls:* the two-commit procedure makes preservation **mechanically
verifiable** (hash of the working file recorded before staging must equal the
blob in commit 1). **The working tree is currently dirty with unrelated
operator changes** (`.gitignore`, `start.ps1`, `.backlogit/stash.jsonl`,
`.claude/`, `.github/copilot/`) — Ship MUST stage **explicit paths only** and
MUST NOT use `git add -A`, `git add .`, `git checkout -- .`, `git stash`, or
`git reset --hard` at any point in this shipment.

### PA-3 — Spike tests that skip in CI (ActionRisk: medium)

A skipping test can silently never run, producing false confidence.

*Controls:* B6's zero-execution fallback is a hard acceptance criterion — each
criterion carries PROVEN / REFUTED / **UNPROVEN** plus the executing
environment, and an all-UNPROVEN run forces a **no-go for C3** and leaves Q1
open. G5 reviews this explicitly.

### PA-4 — Three `complexity: high` tasks (ActionRisk: accepted, justified)

The workspace rule forces a split or de-risking step for `complexity: high`.
B2, B4, and B5 **are** the de-risking step mandated by D2 and design §4.2 —
their purpose is to convert unspecified SDK behaviour into recorded fact, and a
negative result is a successful outcome. Splitting an API round-trip further
would produce meaningless fragments. The 2-hour effort budget still holds
(`Size: M`), and B2 carries an explicit split trigger if the decision-variant
set exceeds four.
*Control:* all three carry hard deadline contexts and clean-skip paths, so none
can consume unbounded time.

### PA-5 — Design amendment as a closure condition (ActionRisk: low-medium)

B7 mutates a governing document. An incorrect amendment would propagate a wrong
normative rule into C3–C11.

*Controls:* B7 is strictly downstream of B5's empirical answer and must cite the
findings artifact; if SQ-b is UNPROVEN the amendment records the question as
open and preserves §4.2's conservative default rather than inventing a rule.

### Not risky — explicitly classified

* A2 — ordinary docs change to a tracked file, fully reviewable.
* B3 — bounded enumeration plus a logging assertion; no destructive action.
* B6 — completes a docs artifact; touches no existing content.

## Protected invariants

* **I1** — Both `cmd` binaries continue to return `not implemented`; no
  behaviour change ships this round.
* **I2** — `scripts/check-retired-architecture.sh` continues to pass, including
  `--self-test` (G4).
* **I3** — `internal/apperr`, `internal/config`, `internal/pathsafe` are
  unmodified. *(Recorded deviation: this defers design §8's C2-assigned
  `startup_timeout_seconds` re-introduction to C3 — see Non-Goals.)*
* **I4** — `go` directive stays `1.24`; toolchain pin `go1.26.5` unchanged.
  Enforced by G7, not assumed.
* **I5** — No retired-architecture token (`acp`, `ipc_name`, `slack`,
  `host_cli`, `channel_id`, `team_id`, `socketmode`) appears as a Go identifier
  in `internal/copilotprobe`. Enforced by **G6** (the CI gate does not scan
  this path).
* **I6** — Unrelated operator working-tree changes are preserved untouched.

## Runtime Verification and Closure

1. `go build ./...`, `go vet ./...`, `go test ./...` green (G3).
2. `go mod tidy` produces no diff against the B1 commit (G2).
3. `go list -m -f '{{.GoVersion}}'` reports `1.24` (G7).
4. `scripts/check-retired-architecture.sh --self-test` passes (G4, I2).
5. G6 scoped token scan over `internal/copilotprobe` returns no matches (I5).
6. A1's two-commit hash comparison shows the operator file preserved
   byte-for-byte in commit 1, additions only in commit 2.
7. B6 findings artifact carries an explicit status for all seven criteria
   (S1–S3, SQ-a–SQ-d) and a consistent go/no-go (G5).
8. B7 amendment present; C2 closure status stated explicitly (G8).
9. `git status` still shows the operator's unrelated modifications intact (I6).

## Human checkpoints

* **HC-1 (confirm-by-exception, before B1):** operator confirms **D10** — that
  Q1 is resolved *by* the C2 spike rather than blocking it, contradicting the
  stale "blocks C2" annotation in stash `A92E3FA0`. The governing deliberation's
  precedence clause already subordinates a conflicting stash annotation to the
  governing artifact, so this is treated as **confirmed unless the operator
  objects** — it does not hold the round.
* **HC-2 (before B6 closure):** operator confirms the recorded Copilot CLI
  version is the intended provisional floor, **or** acknowledges the
  zero-execution fallback and the resulting no-go for C3.
* **A-only closure path:** if sub-epic B is held (operator objects to HC-1, or
  no Copilot CLI is obtainable), A1+A2 close as a **governance-only shipment**.
  The reconciliation, terminology retirement, and deferred register are all
  independently valuable and carry no dependency on the SDK. B-tasks then return
  to the backlog as a C2 shipment of their own.

## Post-Review Remediation Record (2026-09-04)

Plan-review attempt 1: **FAIL** (Scope Boundary Auditor: ADVISORY;
Correctness Reviewer: FAIL). All findings remediated:

| # | Finding | Remediation |
|---|---|---|
| P1 | Four required C2 spike questions unscoped; §4.2 closure condition absent — C2 could not close | Added SQ-a…SQ-d table, tasks **B5** and **B7**, requirements-trace rows, gate G8 |
| P2 | `go mod tidy` prunes build-tagged importers — **factually wrong** | Claim removed; D9a's actual rationale (no importer at all) restored in B1 and Decisions |
| P2 | Missing B2→B4 dependency edge | Edge added; B2 now explicitly owns the shared permission harness; B5 depends on B2+B3 |
| P2 | G2 fails by construction pre-commit | Redefined as a post-commit snapshot diff against the B1 commit |
| P2 | B3 acceptance unsatisfiable — `SessionEvent.Data` is a sealed union | Restated: omitted-variant-reaches-`default:` plus nil/zero handling; mechanism must be recorded |
| P2 | I5 unenforced over the new package | Added gate **G6** (scoped `git grep`) |
| P2 | A1's `git diff` check not executable on an untracked file | Replaced with a two-commit + `git hash-object` procedure |
| P2 | B5/B6 unsatisfiable in a zero-execution run | Added the mandatory UNPROVEN fallback, `not validated` CLI text, Q1-stays-open, and forced no-go |
| P2 | §8 `startup_timeout_seconds` C2 commitment silently dropped | Explicit deferral to C3 recorded in Non-Goals, Decisions, and I3 |
| P2 | Sub-epic bundling could strand with A landed, B held | Added the **A-only closure path**; HC-1 downgraded to confirm-by-exception |
| P3 | R5 mislabelled (it is documentation drift, per F5/§7.5) | R5 restored to its governing definition; deadlock risk renamed **R5b**; R5 traced to B3 |
| P3 | §7.4 fenced surfaces not in Non-Goals | Added all six, with the `InProcessConnection` tag-gating note |
| P3 | I4 verified only against the SDK's own floor | Added G7 and a B1 acceptance assertion |
| P3 | B2 variant cardinality unknown at plan time | Added an enumerate-first step and an explicit >4 split trigger |

### Attempt 2 (FAIL → remediated)

| # | Finding | Remediation |
|---|---|---|
| P1 | **G6 inert** — `\|` (markdown table escaping) read as an ERE *escaped* pipe, so the pattern matched one literal never-occurring string; the gate could never fail and Runtime Verification recorded a vacuous pass | Root cause was the markdown table itself. **All pipe-bearing gate commands moved into a fenced `bash` block** where pipes are unambiguous. G6 rewritten with real alternation, an inverted predicate (`grep` exits 1 on no-match), `-r` to cover **untracked** files, and `_?` + `-i` to catch camelCase (`hostCLI`, `channelID`) as the CI scanner's component-splitter would. **Empirically verified during planning:** exits 0 on a clean fixture, exits 1 on a fixture containing `hostCLI` and `channelID` |
| P2 | UNPROVEN fallback added to B6 only; B2–B5 acceptance unsatisfiable when tests skip; two conflicting no-CLI narratives | Added the normative **No-CLI disposition** block distinguishing *held before B1* (A-only closure) from *discovered during B2–B5* (complete as UNPROVEN); "or recorded UNPROVEN with the skip reason" added to B2, B3, B4, B5 |
| P2 | **NEW defect** — A1 banner placement would fail MD041 (`.markdownlint.json` enables it; banner between frontmatter and H1 makes the first line a non-heading) | Banner placement made normative: **immediately after the existing H1**; added to A1 acceptance |
| P2 | `startup_timeout_seconds` deferral recorded only plan-locally, leaving §8 still directing C2 | **A2 now amends §8 itself** to move it to C3 |
| P3 | G2 command compared worktree-to-index, not to the B1 commit; unused snapshot step | Rewritten as `go mod tidy && git diff --exit-code HEAD -- go.mod go.sum` with B1 as HEAD, in the fenced gate-commands block |
| P3 | B3→B5 edge asserted but B3 produced nothing B5 consumes; findings-artifact creation raced between parallel B2/B3 | B3 now exports the instrumented event-callback hook for B5's SQ-b probe; **artifact skeleton + shared turn/streaming fixture creation moved to B1**, which both depend on |
| P3 | Revision-note arithmetic; stale attempt marker | Corrected to twelve; marker advanced to 2 |
| P3 | R5 relabel conflicts with the source deliberation's own risk table | Added an explicit "supersedes the source deliberation's R5 label, per governing F5" note |

## Context consulted

* Source deliberation (this round's reconciliation), full text
* Governing correction deliberation §§D2, D3, D4a, D9, D9a, Q1–Q6, F1–F5,
  Shipped P2 Audit, risk table
* Design rev 2 §§1.2, 2, 3, **3.1**, **4.1, 4.2, 4.3**, **5.1**, 5.2, 5.3,
  **5.4**, 7.1–**7.4**, **7.5**, **8**, 9
* IMPL-DESIGN, full text
* `go.mod`, `scripts/check-retired-architecture.sh`, `cmd/*/main.go`,
  repository Go inventory
* `docs/compound/2026-09-04-backlogit-sizing-is-wit-gated.md` (harvest constraints)
* Stash `A92E3FA0`, `4989A42D`

<!-- plan-review-attempt: 2 -->

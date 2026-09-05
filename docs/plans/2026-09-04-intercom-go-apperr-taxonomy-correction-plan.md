---
title: "Implementation Plan — apperr taxonomy correction (009-F)"
date: 2026-09-04
status: reviewed
phase: residual-hardening
feature: 009-F
source_deliberation: "docs/decisions/2026-09-04-intercom-go-residual-hardening-triage-deliberation.md"
resolves_stash: [8C2D578D, 9A4C8749]
partially_resolves_stash: [7774C9CA]
depends_on: 008-F
---

# Implementation Plan — apperr taxonomy correction (009-F)

## Problem Frame

`internal/apperr` is the error-construction primitive for the whole product.
Three deferred findings converge on it, and the governing architecture
correction makes fixing them **now** rather than after C3 the decisive call.

1. **`8C2D578D` — retired-architecture contamination.** `KindSlack`, `KindIPC`
   and `KindACP` (string prefixes `slack:`, `ipc:`, `acp:`) survive the
   2026-09-04 architecture correction. Slack, IPC and ACP are all **retired**:
   design revision 2 has no Slack surface, no ACP broker, and no intercom-ctl
   IPC channel. These are dead taxonomy members that the U-E1a retired-
   architecture gate could not be scoped to cover (governing decision **D6a**),
   which is precisely why the gate's scope was narrowed to
   `internal/config/** + config.toml.example + cmd/**`.
2. **`7774C9CA` (apperr parts) — no exported `Stringer`; no sentinel drift
   guard.** `Kind` has only an unexported `prefix()` (`apperr.go:50`), so
   `%v`/`%d` on a `Kind` prints an opaque integer. Separately,
   `kindSentinel` has no bounds validation and nothing asserts that
   `len(allKinds)` tracks the exported `Err*` sentinel count.
3. **`9A4C8749` — `Newf` is not printf-checked.** `.golangci.yml` (verified: 152
   bytes, no govet settings) registers no `printf.funcs` entry, so `go vet`'s
   printf analyzer does not check `Newf`'s format string. With ~470 AppError
   construction sites projected across C3–C11, this is a compounding gap.

**Verified blast radius.** A repo-wide search found **zero** uses of
`KindSlack`, `KindIPC`, `KindACP`, `ErrSlack`, `ErrIPC` or `ErrACP` outside
`internal/apperr` itself. The taxonomy change is therefore fully contained
within one package plus its tests — materially smaller than the stash entry
assumed ("consumed across the codebase"). The real coupling is not call sites;
it is the **pinned count test** at `apperr_test.go:29` (`len(allKinds) != 14`)
and the exhaustive switches in `apperr.go:50` and `sentinel_test.go`.

## Decision — remove outright, do not rename

`8C2D578D` offers three options: remove, rename to architecture-neutral names,
or replace with SDK/transport/operator-surface Kinds. **Remove outright.**

* *Renaming* preserves three taxonomy slots for concepts the architecture no
  longer has, and invites future misuse of a "neutral" bucket. Rule 5:
  simplicity supersedes complexity.
* *Replacing* with SDK/transport/operator Kinds would invent taxonomy for
  surfaces **C3–C6 have not built yet**. Their real requirements are unknown;
  guessing them now is speculative design (YAGNI), and a wrong guess is harder
  to remove later than an absent Kind is to add.
* *Removing* is safe precisely because blast radius is zero, and the correct
  Kinds can be added by the phase that actually needs them, with real
  requirements in hand.

The taxonomy therefore goes from **14 → 11** members.

## Non-Goals

* Does **not** add SDK/transport/operator-surface Kinds. C3+ adds those when it
  has requirements.
* Does not change `Error`, `New`, `Wrap`, or `Newf` semantics (beyond `Newf`
  becoming vet-checked).
* Does not touch `internal/pathsafe` — 010-F owns it. Note `pathsafe` imports
  `apperr`, so 010-F is correctly sequenced *after* this shipment.
* Does not widen `.golangci.yml` beyond the single `printf.funcs` registration.

## Constitution Check

| Principle | Status | Units | Justification |
|---|---|---|---|
| I. Safety-First Go | Satisfied | U1a–U5 | Removals are compile-checked; exhaustive switches force completeness. |
| II. Test-First Development (NON-NEGOTIABLE) | Satisfied | U1a, U2, U3, U5 | U1a moves tests/docs to the target state **before** U1b removes the Kinds. `String()` (U2), the drift guard (U3), and `Wrapf` (U5) each carry tests written first. |
| III. Workspace Isolation and Security Boundaries | N/A | — | No boundary surface touched. |
| IV. CLI Workspace Containment (NON-NEGOTIABLE) | N/A | — | `KindPathViolation` is retained unchanged. |
| V. Structured Observability | Satisfied | U2, U5 | Exported `String()` and cause preservation are both observability fixes. |
| VI. Single Responsibility | Satisfied | all | Taxonomy removal, Stringer, drift guard, constructor, and lint config are separate units. |
| VII. Destructive Command Approval (NON-NEGOTIABLE) | N/A | — | No destructive *command* is executed. Removing Go API is an ordinary compile-checked refactor, not a destructive operation over operator data — it is governed by VIII below, not VII. |
| VIII. Explicit Safety Modes for Elevated Risk | Satisfied | U1b | PA-1 (iota renumbering) carries silent-corruption risk and is therefore gated by an explicit blocking precondition (U1b AC-5) with a human checkpoint, rather than proceeding on assumption. |
| IX. Git-Friendly Persistence | Satisfied | all | Source and YAML. |
| X. Agent Context Efficiency | Satisfied | all | Small, reviewable units. |
| XI. Merge Commit History Preservation (NON-NEGOTIABLE) | Satisfied | — | No history operations. |

## Implementation Units

### U1a — Retarget `apperr` tests and docs to the 11-Kind shape
*Resolves `8C2D578D` (part 1 of 2). Artifact class: tests + docs. Size S, complexity low.*

Test-first half of the removal (Principle II). Update `allKinds` to the 11
retained members and the pinned count `14 → 11`; delete or retarget
`TestACPErrorDisplay` (`apperr_test.go:47`, which references `KindACP` and
asserts `"acp: stream closed"`); update the exhaustive switch in
`sentinel_test.go` and its "14 sentinels" comment; update the package doc
comment that describes a "14-variant" taxonomy.

**Acceptance criteria**
1. `allKinds` contains exactly the 11 retained Kinds; the pinned assertion reads
   11.
2. No test or doc comment references `KindSlack`/`KindIPC`/`KindACP` or the
   number 14.
3. This unit **fails to compile** until U1b lands — that failure is the expected
   intermediate state and is why U1a and U1b are landed as one commit pair.

### U1b — Remove the retired `Slack`/`IPC`/`ACP` taxonomy members
*Resolves `8C2D578D` (part 2 of 2). Artifact class: Go code. Size S, complexity medium. Depends on U1a.*

Remove `KindSlack`, `KindIPC`, `KindACP` from the `Kind` const block; remove
`ErrSlack`, `ErrIPC`, `ErrACP` from `sentinel.go`; remove their `prefix()`
cases.

Because `Kind` is an iota const block, removing middle members **renumbers the
remaining Kinds**. This is safe only because the numeric values are not
persisted anywhere — AC-5 makes verifying that a **blocking precondition**, not
an assumption.

**Acceptance criteria**
1. `KindSlack`, `KindIPC`, `KindACP`, `ErrSlack`, `ErrIPC`, `ErrACP` no longer
   exist anywhere in the package.
2. `len(allKinds) == 11` and the pinned count assertion passes.
3. Every remaining Kind still has a distinct, unchanged string prefix; a test
   asserts prefix uniqueness across the reduced set.
4. The retained taxonomy is exactly: `Config`, `DB`, `MCP`, `Diff`, `Policy`,
   `PathViolation`, `PatchConflict`, `NotFound`, `Unauthorized`,
   `AlreadyConsumed`, `IO`.
5. **BLOCKING precondition, verified not assumed:** no `Kind` numeric value is
   persisted to disk, serialised over a wire format, or stored in any DB/backlog
   schema. Mechanical check: `Kind` carries no struct tags, no
   `MarshalJSON`/`MarshalText`/`GobEncode`, and no `int(k)` conversion escapes
   the package. If **any** is found, **halt and return the unit blocked** — the
   removal then requires a migration plan not covered here.
6. `go build ./...` and `go test ./... -race` pass; no code outside
   `internal/apperr` required modification.

### U2 — Add an exported `String()` for `Kind`
*Partially resolves `7774C9CA`. Artifact class: Go code + tests. Size S, complexity low.*

Implement `func (k Kind) String() string` so `Kind` satisfies `fmt.Stringer`,
returning the Kind's name. Keep the unexported `prefix()` as the
**display-contract** accessor used by `Error.Error()`.

Keeping the two separate is deliberate: `prefix()` participates in a pinned
error-rendering contract (e.g. `"path violation: …"` / `"io: …"` style), while
`String()` is a debugging/observability affordance. Collapsing them would couple
log formatting to the pinned wire-visible error text.

Verified Go semantics: `Error()` (`apperr.go:118`) calls `k.prefix()`
explicitly and never routes through `%v`/`String()`, and `kindSentinel` is a
distinct named type (`type kindSentinel Kind`) that does **not** inherit
`Kind`'s methods — so adding `String()` cannot alter any existing rendering.

**Acceptance criteria**
1. `Kind` satisfies `fmt.Stringer`; `%v` on a `Kind` prints its name, not an
   integer.
2. `Error.Error()` output is **byte-identical** to current behaviour for every
   retained Kind (a golden test asserts this — the display contract must not
   shift as a side effect).
3. `String()` covers all 11 Kinds exhaustively, with a defined fallback for an
   out-of-range value.

### U3 — Add a taxonomy drift test
*Partially resolves `7774C9CA`. Artifact class: tests. Size S, complexity low. Depends on U1b.*

Add a test asserting `len(allKinds)` equals the number of exported `Err*`
sentinels, so adding a Kind without its sentinel (or vice versa) fails CI. This
is the guard that makes future taxonomy changes safe, which is why it lands in
the same shipment that performs one.

The original "bounds-validate `kindSentinel`" idea from `7774C9CA` is
**deliberately dropped**: `kindSentinel` is unexported and its values are
package-level `var`s constructed only from valid `Kind` constants, so there is
no reachable API surface on which out-of-range validation could fire. Adding a
check for an unconstructable state would be dead code (rule 5, simplicity). The
drift test below covers the actual risk the finding cared about — taxonomy and
sentinel list silently diverging.

**Acceptance criteria**
1. A test asserts `len(allKinds)` equals the exported `Err*` sentinel count and
   fails if either side changes alone.
2. Proven by construction: adding a Kind without its sentinel makes the test
   fail (demonstrate, then revert).
3. `errors.Is` behaviour for all 11 retained sentinels is unchanged, asserted by
   the existing round-trip tests.

### U4 — Register `Newf` with the govet printf analyzer
*Resolves `9A4C8749`. Artifact class: config. Size S, complexity low.*

**Verified schema:** `.golangci.yml` declares `version: "2"`, and CI installs
`golangci-lint/v2`. In v2 the top-level `linters-settings` key **does not
exist** — settings live under `linters.settings`. Additionally, `printf.funcs`
uses `<import/path>.Func` for a package-level function; the parenthesised
`(<import/path>.Type).Method` form is for methods only. `Newf` is a
package-level function, so it takes **no parentheses**.

Add exactly this:

```yaml
linters:
  settings:
    govet:
      settings:
        printf:
          funcs:
            - github.com/softwaresalt/intercom-go/internal/apperr.Newf
```

**Acceptance criteria**
1. `.golangci.yml` registers `apperr.Newf` under `linters.settings.govet` using
   the unparenthesised fully-qualified function path shown above.
2. **Proven effective, not merely configured:** a committed `testdata`
   fixture or a deliberate temporary mismatch (e.g.
   `Newf(KindConfig, "%d", "str")`) is reported by `golangci-lint run`. A
   v1-schema key or a parenthesised path is silently ignored by v2, so this AC
   is the unit's real test — configuration alone proves nothing.
3. The existing lint suite still passes on the unmodified tree (no new findings
   across current call sites).
4. No other linter or setting is added or altered.

### U5 — Add a cause-preserving constructor (`Wrapf`)
*Enables `010-F` U1. Artifact class: Go code + tests. Size S, complexity low.*

**This unit exists because adversarial review found a blocker.** The existing
`Wrap(kind Kind, cause error)` (`apperr.go:142`) derives `msg` from
`cause.Error()` and accepts **no message parameter**. `New(kind, msg)` accepts a
message but **drops the cause**. Therefore no current constructor can produce an
error that both preserves the cause (for `errors.Is(err, os.ErrNotExist)`) and
retains human-readable context — which is exactly what `010-F` U1 requires and
what the existing `TestNewRootOnMissingDirReturnsPathViolation` asserts.

Add `func Wrapf(kind Kind, cause error, format string, args ...any) *Error`,
setting the formatted message **and** retaining `cause` for `Unwrap()`.

**Acceptance criteria**
1. `Wrapf` sets the formatted message and preserves `cause` such that
   `errors.Is`/`errors.As` reach the cause through `Unwrap()`.
2. `errors.Is(err, <kind sentinel>)` still matches for a `Wrapf`-constructed
   error (Kind discrimination must not regress).
3. `Wrapf` is registered in `.golangci.yml`'s `printf.funcs` alongside `Newf`
   (same unit, same mechanism — a new format-string constructor must not ship
   unchecked).
4. A nil-cause call is guarded and does not panic, matching `Wrap`'s existing
   nil handling.
5. Existing `Wrap` and `New` behaviour is unchanged.

## Dependency Graph

```
U1a ──► U1b        (tests/docs retargeted first; landed as one commit pair)
U1b ──► U3         (drift guard asserts the post-removal shape)
U1b ──► U2         (String() enumerates the reduced taxonomy)
U5                 (independent of the removal; REQUIRED BY 010-F U1)
U4                 (config; must also register U5's Wrapf)
```

**Downstream contract:** `010-F` U1 consumes `Wrapf` from U5. This is the real,
hard 009-F → 010-F dependency. (Adversarial review correctly noted that the
originally stated justification — "`pathsafe` imports `apperr`" — was spurious,
since `pathsafe` already imports `apperr` and 009-F introduces no symbol it
otherwise needs. U5 makes the dependency genuine.)

## Post-Review Remediation Record (2026-09-04)

Adversarial multi-persona review (Correctness, Maintainability, Constitution,
Scope) returned **FAIL** on the pre-remediation draft. Changes applied:

| Finding | Severity | Remediation |
|---|---|---|
| `Wrap` cannot carry a message; 010-F U1's ACs were mutually unsatisfiable | **blocker** | Added **U5 (`Wrapf`)**. This is now the real cross-shipment dependency. |
| U4's YAML used golangci-lint **v1** `linters-settings` against a **v2** config, and a malformed parenthesised func path — a silent no-op | **blocker** | Rewrote U4 with the verified v2 `linters.settings` nesting and unparenthesised `pkg.Func` form; AC-2 hardened to a committed fixture. |
| U1 exceeded the 2-hour rule (4 files, >5 functions) | major | Split into **U1a** (tests/docs) + **U1b** (code removal). |
| U1 omitted `TestACPErrorDisplay` (`apperr_test.go:47`) and the "14-variant" package doc | major | Both named explicitly in U1a AC-1/AC-2. |
| U3 AC-2 targeted a non-existent API (`kindSentinel` is unexported, values unconstructable out-of-range) | major | Dropped as dead code; replaced with a demonstrable drift test. |
| Principle VII/VIII misclassified; Principle II omitted U2; principle titles truncated | major | Constitution table rewritten with exact ratified titles and corrected classifications. |
| U2's display-contract example referenced `"acp: stream closed"`, deleted by U1b | minor | Example retargeted; Go method-set semantics recorded explicitly. |

**Confirmed correct by review, unchanged:** the zero-external-use blast-radius
claim; the exact 11-Kind retained list; the safety of iota renumbering (no
persistence); the byte-identical `Error()` claim; and the choice to remove
rather than rename.

## Risks and Caveats

**R1 — iota renumbering (HIGHEST RISK).** Removing middle `Kind` members shifts
the numeric value of every later Kind. If any `Kind` value were persisted or
serialised, this silently remaps stored errors. U1 AC-5 requires verifying no
such persistence exists before landing. Current evidence says none does
(`apperr` has no serialisation surface and no DB schema exists yet), but this
must be **confirmed**, because the failure is silent and data-corrupting rather
than a build break. This risk shrinks to zero if confirmed now and grows with
every phase that ships afterwards — another reason to do it before C3.

**R2 — Removing exported API.** `ErrSlack`/`ErrIPC`/`ErrACP` are exported. Blast
radius is verified zero *inside this repo*, and `internal/` packages are
unimportable outside the module by Go's own rules, so there is no external
consumer by construction.

**R3 — Display-contract drift.** U2 could accidentally alter `Error()` output.
U2 AC-2 pins it with a byte-identical golden assertion.

**R4 — `printf.funcs` silently not matching.** An incorrect qualified path is
accepted by golangci-lint and simply never matches, producing a config that
looks correct and checks nothing. U4 AC-2's deliberate-mismatch probe is the
mitigation.

**R5 — Divergence from the historical oracle.** The retired Kinds mirrored the
Rust oracle's taxonomy. Under decision **D8** that oracle is demoted to
historical reference and parity is explicitly no longer a definition of done, so
divergence here is intended, not a regression.

## Plan Hardening Signals

| Signal | Present | Note |
|---|---|---|
| Schema/contract change | **Yes** | Error taxonomy is a pinned contract; iota values shift. |
| Security-sensitive behaviour | No | `KindPathViolation` retained unchanged. |
| Migration | **Yes** | 14 → 11 taxonomy members. |
| External dependency | No | — |
| Silent-corruption potential | **Yes** | R1. |

**Requires plan hardening: yes.**

---

# Plan Hardening

**Hardening required: YES** (pinned contract change; migration 14 to 11; silent-corruption potential). Hardened 2026-09-04.

## Context consulted

`internal/apperr/apperr.go` (Kind const block L21-44, `prefix()` L50), `sentinel.go` (L20-33 sentinels, `Is` L42), `apperr_test.go:29` (pinned count 14), `sentinel_test.go` (exhaustive switch), `.golangci.yml` (152 bytes, no govet settings), governing decision D6a, and a verified repo-wide search proving zero external users of the retired Kinds.

## Protected invariants

* **`Error.Error()` output is byte-identical for every retained Kind.** The display contract is pinned.
* **`errors.Is` sentinel matching is unchanged for all 11 retained Kinds.**
* **`KindPathViolation` and its sentinel are untouched** (Constitution IV surface).
* **No `Kind` numeric value may be persisted anywhere.** This is the precondition that makes iota renumbering safe.
* **`internal/config` and `internal/pathsafe` must compile unchanged.**

## Risky actions (ProposedAction / ActionRisk)

### PA-1 — Remove three middle members of an iota const block (U1)
**Risk: HIGH (silent data corruption if the precondition is false).** Removing middle members renumbers every subsequent `Kind`. If any numeric value were persisted or serialised, stored errors would silently remap to different Kinds — no build break, no test failure, wrong data.
**Control:** U1 AC-5 requires **verifying** — not assuming — that no `Kind` value is persisted to disk, serialised over a wire format, or stored in a DB/backlog schema. Current evidence indicates none is (no serialisation surface, no DB schema yet), which is exactly why this must land **before** C3 begins adding surfaces that could persist one. If the verification finds any persistence, **stop and return the unit blocked**; the removal would need a migration.

### PA-2 — Remove exported API (`ErrSlack`/`ErrIPC`/`ErrACP`) (U1)
**Risk: LOW.** Verified zero in-repo users; `internal/` is unimportable outside the module by Go's own rules, so no external consumer can exist.

### PA-3 — Modify shared lint configuration (U4)
**Risk: MEDIUM (silent no-op).** An incorrect fully-qualified path in `printf.funcs` is accepted by golangci-lint and simply never matches — a control that appears configured and checks nothing. A secondary risk is newly-surfaced findings across existing call sites blocking CI.
**Control:** U4 AC-2 requires a deliberate-mismatch probe proving the analyzer actually fires, then reverting the probe. U4 AC-3 requires the existing suite to stay clean. U4 AC-4 forbids touching any other linter setting.

### Not risky — explicitly classified
U2's additive `String()`; U3's additional test assertions.

## Rollback points

U4 is config-only and reverts independently. U1 is the atomic contract change; U2 and U3 build on its post-removal shape, so a rollback of U1 requires rolling back U2 and U3 together. Ship should land U1 (with its updated count test) as one coherent commit.

## Deepened runtime verification

1. `go build ./...` and `go test ./... -race` green with **no changes required outside `internal/apperr`** — this is itself the blast-radius assertion.
2. Golden assertion that `Error()` strings for all 11 retained Kinds are byte-identical to pre-change output.
3. Prefix-uniqueness test across the reduced set.
4. Drift test proven by deliberately adding a Kind without its sentinel and confirming failure, then reverting.
5. `printf.funcs` proven by the deliberate-mismatch probe (PA-3).
6. Retired-architecture gate (U-E1a) re-run; confirm `apperr` no longer contributes contamination.

## Human checkpoints

* **Blocking:** if U1 AC-5's persistence verification finds any persisted `Kind` value, halt and escalate — the removal then requires a migration plan not covered here.
* Otherwise none; the taxonomy decision (remove, not rename) is settled in the plan body.

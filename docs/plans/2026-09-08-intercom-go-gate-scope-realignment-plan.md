# Implementation Plan — U-E1a Gate Scope Realignment

- **Date**: 2026-09-08
- **Source deliberation**: `docs/decisions/2026-09-08-intercom-go-gate-scope-realignment-deliberation.md`
- **Stash entries covered**: `A92E3FA0`, `4989A42D` (both **retained**, not archived)
- **Requires plan hardening**: **yes** — modifies a **mechanical CI control** whose blast
  radius is every future pull request in the repository. A false positive blocks unrelated
  work; a false negative silently defeats the retirement enforcement that D6/D6a/D6b exist
  to provide.
- **Constitution**: Principle II (Test-First, NON-NEGOTIABLE — red phase before green),
  Principle VIII (safety mode: **`careful` + `freeze-scope`**), Principle XI (merge commit
  required — see R7), Principle VI (Single Responsibility / one skill domain per task)
- **Revision 2** incorporates the resolution of five P0 and six P1 findings from a
  seven-persona review panel (Architecture, Correctness, Security Lens, Scope Boundary,
  Constitution, Schema-CLI-Docs Coupling). See §10 for the disposition of every finding.
- **Plan-review gate**: **PASS** (cycle 2 of max 3). Re-review by the Correctness and
  Scope Boundary personas confirmed all rev-1 P0/P1 findings resolved and returned
  `SCOPE VERDICT: COMPLIANT`. Rev 2.1 additionally applied the cycle-2 P2/P3 advisories
  (named dispatch observable, independently-derived assertion universe, per-assertion
  output, symbol-keyed edits, manifest ownership, single dispatch source, corrected I-B
  rationale). **Zero open P0/P1.**
  <!-- plan-review-attempt: 2 -->
- **Merge requirement**: all six units MUST ship in a **single PR** with a **merge commit**
  (Principle XI, NON-NEGOTIABLE). `--self-test` is deliberately RED from U1 until both U4
  and U5 land; a squash or rebase would destroy the red→green evidence (R7).

---

## 1. Objective

Retire the vestigial **D6a narrowing** of the U-E1a retired-architecture gate by
broadening `scripts/check-retired-architecture.sh` from `internal/config/**` to
`internal/**`, pin the new scope against silent re-narrowing, give the gate's previously
untested Go scanner its first fixture coverage, and reconcile the two governing trackers
with merged reality.

This closes `4989A42D` completion condition (b) once the shipment merges, and is the
single artifact where `A92E3FA0` and `4989A42D` intersect.

## 2. Anti-Goals (hard boundaries — `freeze-scope`)

- **No modification to `.github/workflows/ci.yml`.** The gate is already wired
  (`ci.yml` L262–L266 run the script and its `--self-test`); broadening the script needs
  no workflow edit. This file is the shared surface of **excluded** entries `F4F4A959`
  and `F47DB9A9` and must not be opened.
- **No modification to `scripts/check-unignore-regression.sh`** — shared surface of
  **excluded** entries `2787DA56` and `4C5BEC23`.
- **No modification to `scripts/check-write-path-precondition.sh`** — `4989A42D`
  explicitly records that `011.004-T` kept this gate as a *separate* script so condition
  (b) stays cleanly measurable. The two gates must remain independent.
- **No change to the `forbidden_parts` token taxonomy.** Adding/removing retired tokens is
  a separate decision; this slice changes *scope*, not *vocabulary*.
- **No change to any non-test `.go` file under `internal/**` or `cmd/**`.** This slice adds
  no production code. (`internal/apperr`, `internal/pathsafe`, `internal/config` are the
  surfaces of shipments `008-S`/`010-S`/`011-S` and are settled.)
- **No archiving of `A92E3FA0` or `4989A42D`.** See deliberation D-4.
- **No C3 (agent adapter / ACL) work**, and no new stash entry for C3 — `A92E3FA0`
  already tracks it.
- **No re-narrowing escape hatch by default.** If a future false positive appears, the
  remedy is a targeted exclusion with a recorded reason, never a silent scope rollback.

## 3. Units of Work (rev 2)

Ordering places **Go-scanner coverage before the broadening** (U1 → U4 → U5 → U2), so the
scanner is proven before it inherits 8 more files — the sequencing the deliberation §2.5
argument actually implies, and which the rev 1 order inverted.

### U1 — RED: structural scan-scope and dispatch assertions

- **Domain**: script/test harness. **Size**: S. **Complexity**: medium.
- **File**: `scripts/check-retired-architecture.sh`. **Note**: this unit must extract a
  shared `select_repo_paths()` helper out of `run_repo_scan()` so the assertion runs
  against the **real** selection logic (including the `git ls-files` call), not a
  re-implemented copy. The declared surface therefore includes the repo-scan path, not
  only the self-test path.
- Add assertions to `--self-test`:
  1. **Inclusion (structural, not enumerated)** — *every* tracked non-test,
     non-`testdata` `.go` file under `internal/` is selected. **The expected universe must
     be derived independently** (its own `git ls-files -- internal/` query plus an
     independently written test/testdata rule) and compared against `select_repo_paths()`.
     Deriving *expected* from `select_repo_paths()` itself would be tautological: it would
     pass on arrival (defeating the mandatory red proof), and a future re-narrowing of
     *both* pathspec and predicate would shrink expected and actual together, silently
     defeating R5. Define "non-testdata" identically on both sides — the predicate uses the
     substring test `'/testdata/' in path`, not a path-segment test.
  2. **Exclusion, scoped to `internal/`** — no `internal/**/*_test.go` and no
     `internal/**/testdata/**` path is selected.
  3. **Self-scan guard (I-A)** — no path under `scripts/` is selected, **and** a direct
     predicate probe asserts `should_scan_repo_path('scripts/testdata/retiredgo/<name>.go')`
     is `False`. R2 claims two independent guards, but the selection-set check alone
     exercises only the pathspec, because `scripts/` paths never enter `git ls-files`
     output. The probe is what pins the predicate's fail-closed behaviour.
  4. **Non-empty** — the selection count is `> 0` (closes the silent fail-open where a
     pathspec that matches nothing exits 0).
- **Output requirement**: the self-test must emit a **per-assertion PASS/FAIL line**.
  Assertion 1 closes in U4 and the dispatch assertion closes in U5, so a bare exit code
  cannot demonstrate a given unit's acceptance while a sibling assertion is still red.
- **Deliberately NOT asserted**: that no `cmd/**/*_test.go` is selected. Verified: the
  `cmd/` branch early-returns before the `_test.go`/`testdata` filters, so
  `cmd/intercom/main_test.go`, `cmd/intercom/config_flag_test.go` and
  `cmd/intercom-ctl/main_test.go` **are** selected today. Asserting otherwise would be
  permanently red. This asymmetry is recorded as a known inconsistency (§8), not fixed.
- **Red proof (required)**: `--self-test` must **FAIL** on assertion 1. A U1 that passes on
  arrival is a defective assertion and must be rejected.
- **Acceptance**: fails before U4, naming the uncovered `internal/` packages; contains no
  hardcoded package-name list; emits per-assertion PASS/FAIL lines.

### U2 — Go fixture harness: per-kind discovery and engine dispatch

- **Domain**: script/test harness. **Size**: S. **Complexity**: medium.
- **File**: `scripts/check-retired-architecture.sh`.
- The existing harness **cannot** accept Go fixtures, and adding them naively produces a
  **false green** in a merge-blocking control. Three concrete blockers, all verified:
  1. `fixture_glob = "retired-*.toml"` globbed non-recursively over `scripts/testdata`
     never sees a subdirectory.
  2. Manifest keys are bare `path.name`, so Go entries fail the
     `extra_manifest = set(manifest) - set(discovered)` reconciliation.
  3. **Worst**: the `engines` list is TOML-only and is applied to *every* discovered
     fixture. A `.go` fixture would raise in `tomllib.loads()`, so an ACCEPT fixture fails
     and a REJECT fixture "passes" **for the wrong reason** — recorded as rejected while
     proving nothing about `scan_go`.
- Required changes, stated normatively:
  - a second fixture suite descriptor (`glob`, `manifest_path`, `engines`) for Go fixtures
    under `scripts/testdata/retiredgo/`;
  - per-suite discovered-vs-manifest reconciliation (keep the bidirectional drift guard);
  - **suffix-based engine dispatch** so `.go` fixtures run `scan_go` and `.toml` fixtures
    run the two TOML engines.
- **Acceptance**: a Go REJECT fixture's recorded failure text names the **retired token and
  Go identifier** (the `scan_go` message format) — *not* a TOML parse error. This criterion
  mechanically excludes the false-green path.
- **Manifest decision (pinned, no executor discretion)**: a **new sibling manifest**,
  `scripts/testdata/retiredgo-manifest.json`, with keys sorted for stable diffs
  (Principle IX). The existing `retired-manifest.json` is **not** modified. **This unit
  creates the manifest file** (U3 only adds entries to it), so U2 = 2 files and U3 = 2
  files, keeping both inside the granularity bound.
- **Single dispatch source (simplicity over complexity)**: the harness must route fixtures
  through the **same** `engine_for_path()` helper that U5 extracts for `scan_path()`, not a
  second parallel dispatch table. Two dispatch tables in one file would drift, and — more
  importantly — fixtures adjudicated by a private harness table would **not** exercise the
  real dispatch path that U5's assertion pins. If U5 has not yet landed, U2 introduces
  `engine_for_path()` and U5 consumes it; either way there is exactly one dispatch source.
- **Correction of record**: rev 1 said "mirror the `writepath` pattern" while also saying
  "reuse the existing manifest shape." These are incompatible — verified,
  `scripts/testdata/writepath/` has **no manifest** and uses an `accept-`/`reject-`
  filename convention. This plan adopts the **manifest** contract (it keeps the
  "discovered but unlisted" drift guard) and does **not** adopt the filename convention.
- Update the script header usage block (L16–L19), which currently documents the self-test
  as `retired-*.toml`-only and will otherwise ship stale.
- **Depends on**: U1.

### U3 — Go fixture corpus (minimal, false-green-proof)

- **Domain**: fixtures. **Size**: S. **Complexity**: low.
- Add exactly **two** fixtures under `scripts/testdata/retiredgo/`:
  - **1 REJECT** — a retired token as a bare Go identifier in a const or type declaration,
    proving `scan_go` detects contamination in a `.go` file.
  - **1 ACCEPT** — a clean file plus a retired token appearing **only** in a line comment,
    proving the harness can produce a clean verdict and that masking is applied. Without an
    ACCEPT fixture the harness cannot be shown to distinguish detection from always-failing.
- **Fixture content constraints (mandatory — closes the gitleaks finding)**: retired tokens
  may appear **only** as bare identifiers, type names, or struct field names. **Never** as
  `<token>Token` / `<token>Secret` / `<token>Key` / `<token>URL` assignments, never with an
  `xox[baprs]-` or `hooks.slack.com` prefix, and never with a high-entropy literal. Rationale:
  `ci.yml` runs a **blocking** `gitleaks detect --source . --no-git` with no `.gitleaksignore`
  in the repo, and `secret-scan-history.yml` scans **full history** — a merged credential-shaped
  fixture is **not** remediable by a later delete.
- Every fixture must be **gofmt-clean, goimports-clean, and parseable Go** with a
  `package retiredgo` clause — `ci.yml` runs `gofmt -l .` and `goimports -l .` as filesystem
  walks that **do** descend into `testdata/` (unlike `golangci-lint run ./...`).
- Carry the `writepath` header comment explaining the Go toolchain ignores `testdata/`.
- **Depends on**: U2.

### U4 — GREEN: broaden the scan scope to `internal/**`

- **Domain**: script. **Size**: XS. **Complexity**: low.
- **File**: `scripts/check-retired-architecture.sh` — exactly three edits, **keyed to
  symbols, not line numbers** (U1 and U2 insert code ahead of these sites, so every
  absolute line number below is stale by the time this unit executes; the numbers are
  the plan-time anchors only):
  1. `should_scan_repo_path()`'s `internal/` prefix test (plan-time L75)
     `if not path.startswith('internal/config/'):` → `if not path.startswith('internal/'):`
  2. `run_repo_scan()`'s `git ls-files` pathspec (plan-time L391)
     `'internal/config/**'` → `'internal/**'`
  3. The header **Usage** block's scope sentence (plan-time L13–L15) → `internal/**`.
     **Ownership note**: the header's "…as a Go identifier or TOML key" claim on the same
     sentence belongs to **U5**, not this unit. Split the edit so the two parallel DAG
     branches do not both rewrite one line.
- Preserve the `/testdata/`, `_test.go` and `.go`-suffix filters unchanged (invariant I-B).
- **Acceptance**: U1 assertion 1 reports **PASS** on its per-assertion line;
  `check-retired-architecture.sh` (no args) exits 0.
- **Depends on**: U1, U3 (coverage lands before the scanner inherits new files).
- **Evidence**: a patched copy of the real engine run against the real tree exits **0**
  (deliberation §2.3).

### U5 — RED→GREEN: route `config.toml.example` to the TOML engine (P0)

- **Domain**: script. **Size**: XS. **Complexity**: low.
- **File**: `scripts/check-retired-architecture.sh`, `scan_path()` (plan-time L374–379).
- This unit owns **both** halves of its own red/green pair — the dispatch assertion moved
  here from U1 so each red proof sits with its green fix (and U1 stays within the
  four-scenario granularity bound).
- **RED — named observable (required).** Asserting "the file produces findings" is
  **unsatisfiable**: `config.toml.example` is clean, so `scan_path()` returns `[]` both
  before *and* after the fix, which would leave the assertion permanently red and invite an
  executor to weaken or delete it. Instead extract a pure dispatch helper —
  `engine_for_path(path) -> 'go' | 'toml' | None` — which `scan_path()` then consumes, and
  assert `engine_for_path('config.toml.example') == 'toml'`. That assertion is genuinely
  red before the fix and genuinely green after it.
- **GREEN**: dispatch on `path.name == 'config.toml.example' or path.suffix == '.toml'`.
- Also correct the header claim ("…as a Go identifier or TOML key"), which is false until
  this lands, and the header's stale "for the corrected C1 config surface" characterization.
- **Verified preconditions** — both are required for the fix to land green, and
  token-freeness alone is *not* sufficient:
  1. `config.toml.example` contains **no retired token** (verified); and
  2. `config.toml.example` is **valid TOML** (verified) — `scan_toml_with_tomllib` fails
     **closed** on any parse exception, so an unparseable example file would redden the
     gate at this unit's own commit boundary and breach invariant I-E.
- **Acceptance**: the dispatch assertion reports **PASS**; the gate still exits 0.
- **Depends on**: U1.

> **Sequencing consequence (explicit).** `--self-test` is expected **RED from U1 until
> both U4 and U5 have landed** — they are parallel DAG branches, each closing only one
> assertion. All units MUST therefore ship in a **single PR**, since `ci.yml`'s
> `--self-test` step is the blocking one. The per-assertion PASS/FAIL output required by U1
> is what makes each unit's acceptance observable while sibling assertions are still red.


### U6 — Reconcile governing artifacts (bounded, enumerated)

- **Domain**: docs. **Size**: S. **Complexity**: low.
- **Exactly two files** — the acceptance criterion is an **enumerated list**, not a
  repository-wide grep. A repo-wide grep would reach `.github/workflows/ci.yml`
  (**excluded** `F4F4A959`/`F47DB9A9`) and the 2026-09-07 pathsafe plan (**excluded**
  `1C6C3B46`), making a faithful executor breach the exclusion contract.
  1. `docs/plans/2026-09-04-intercom-go-architecture-correction-plan.md` — annotate as
     historical at **L361–364** (the U-E1a unit spec), **L365–376** (scope corrections),
     **L515–519** (R7a), **L538** (false-positive risk row), **L642** (I5 invariant row),
     **L860** (defect-table row 6). Rev 1 named only the last three; the first three are the
     more explicit normative statements. Discharge the **L377–378** "this scope is mirrored
     normatively in D6" coupling invariant by doing (2) in the same change.
  2. `docs/decisions/2026-09-04-intercom-go-architecture-correction-copilot-sdk-deliberation.md`
     — add a dated amendment **D6c** recording that D6a's narrowing is retired, with the
     reason (apperr Kinds removed by `009.002-T`; I5 discharged) and the §2.4 "born
     narrowed" correction. **Annotate, do not rewrite** — the original rationale stays
     readable. This artifact self-declares itself GOVERNING and wins over any plan, so it
     must not be left contradicting the shipped script.
- Record the **residual coverage statement**: enforced scope is *non-test* Go files under
  `internal/**`; `internal/**/*_test.go` and `internal/**/testdata/**` remain accepted,
  unenforced surface. Also record that the gate's **anti-accident-only** threat model (D6b)
  is **unchanged** by this broadening — gate coverage is a hygiene precondition for C3, not
  an authorization or adversarial control, and does not partially discharge H5.
- **Acceptance**: both files updated; markdownlint passes; no other file in the diff.
- **Depends on**: U4, U5.

## 4. Dependency Order

```text
U1 (RED assertions)
  ├─> U2 (Go fixture harness)
  │     └─> U3 (Go fixture corpus)
  │           └─> U4 (broaden to internal/**)
  │                 └─> U6 (docs + D6c amendment)
  └─> U5 (config.toml.example dispatch fix) ──> U6
```

`U1 → {U4, U5}` is non-negotiable red/green ordering (Principle II). `U2 → U3 → U4` places
Go-scanner coverage **before** the broadening.

## 5. Risk Register

| # | Risk | Sev | Mitigation | Owner unit |
|---|---|---|---|---|
| R1 | Broadened gate false-positives on future legitimate code (a C3 identifier splitting to `host`+`cli`, or a component `acp`) and blocks unrelated PRs | medium | Verified clean today across all 4 packages. Remedy is a **targeted, commented exclusion**, never a silent re-narrowing. **Named downstream consumers**: `internal/pathsafe/{lexical,pathsafe,root}.go` enters scan scope for the first time and is the surface of excluded entries `700B41CE` (critical), `F133AB7E`, `2362BBB5`, `BF5DE670` — a future collision there would block a critical security entry | U4 |
| R2 | Gate self-matches its own new Go fixtures | medium | **Two independent guards**: the pathspec never includes `scripts/`, and `should_scan_repo_path` fails closed for any non-`cmd/`, non-`internal/`, non-`config.toml.example` path. U1 assertion 3 pins it | U1 |
| R3 | New `.go` fixtures redden `gofmt -l .` / `goimports -l .` | medium | These are **filesystem walks that do descend into `testdata/`** — the "Go ignores testdata" argument does **not** cover them. U3 requires gofmt/goimports-clean, parseable fixtures; both commands added to §6 | U3 |
| R4 | New `.go` fixtures trip the blocking `gitleaks detect` and are then **permanent in history** | high | U3's fixture-content constraints (bare identifiers only, no token-shaped literals). §6 runs gitleaks **before first commit**, while renaming is still cheap | U3 |
| R5 | Scope silently re-narrows later (D6a recurring), or a pathspec/predicate half-edit selects nothing | medium | U1 assertions 1 and 4 are the durable structural guards; assertion 4 closes the fail-open | U1 |
| R6 | Go fixtures produce a **false green** by routing through the TOML engines | high | U2's acceptance requires the REJECT fixture's failure text to name the token and Go identifier, not a TOML parse error | U2 |
| R7 | Squash/rebase merge collapses the U1 red commit into U4/U5, destroying the Principle II evidence | medium | Principle XI (NON-NEGOTIABLE) forbids squash/rebase. Ship must verify merge strategy and halt with P-009 on detection | Ship |
| R8 | `011-S` recorded "do not modify `check-retired-architecture.sh`" | low | That anti-goal was scoped to `011-S`, to keep condition (b) cleanly measurable. `4989A42D` names a future cycle as the authorized venue; this is that cycle | — |

## 6. Verification

- `bash scripts/check-retired-architecture.sh` → exit 0
- `bash scripts/check-retired-architecture.sh --self-test` → exit 0
- `gofmt -l .` → **no output** (Quality Gate #1); `goimports -l .` → no output
- `golangci-lint run ./...` → exit 0
- `go build ./...`, `go vet ./...`, `go test ./...` → exit 0
- `gitleaks detect --source . --no-git -v` (pinned 8.30.1) → exit 0, run **before the first
  commit containing the fixtures**
- `bash scripts/check-write-path-precondition.sh --self-test` → exit 0 (unchanged; proves
  gate independence, invariant I-C). Read-only execution of an excluded entry's file is not
  triage, mutation, or inclusion
- `git diff --name-only` limited **exactly** to: `scripts/check-retired-architecture.sh`,
  `scripts/testdata/retiredgo/` (2 files), `scripts/testdata/retiredgo-manifest.json`,
  `docs/plans/2026-09-04-intercom-go-architecture-correction-plan.md`,
  `docs/decisions/2026-09-04-intercom-go-architecture-correction-copilot-sdk-deliberation.md`

## 7. Constitution Check

Mapped against the **authoritative** `constitution.instructions.md` numbering (rev 1 used
AGENTS.md's condensed numbering and mislabelled two rows while omitting five principles).

| Principle | Status |
|---|---|
| I. Safety-First Go | **Partially applicable** — no production Go added, and `golangci-lint run ./...` is package-pattern based so it ignores `testdata/`. But `gofmt -l .` and `goimports -l .` **do** walk `testdata/`; U3 fixtures must satisfy both |
| II. Test-First (NON-NEGOTIABLE) | **Satisfied** — U1 red phase precedes U4/U5 green; U2/U3 land Go-scanner coverage *before* U4 broadens scope. Note: this work is **not** reachable by `go test ./...`; the bash `--self-test` is the test surface, by construction |
| III. Workspace isolation | Satisfied — all paths repo-relative |
| IV. CLI Workspace Containment (NON-NEGOTIABLE) | **Satisfied going forward, with one recorded deviation** — the §2.3 probe wrote to an out-of-tree temp path; disclosed in the deliberation, artifact deleted, future probes constrained to the workspace tree |
| V. Structured Observability | **Satisfied** — U1's red-proof output is captured in the backlogit task record; units are registered as backlogit tasks, not a markdown-only list |
| VI. Single Responsibility | **Satisfied** — rev 2 splits rev 1's oversized U3 into harness (U2), fixtures (U3), and the two green edits (U4, U5), so each task is one domain and within the 2-hour rule. No new dependencies (`git`, `python3` already required) |
| VII. Destructive Command Approval (NON-NEGOTIABLE) | **N/A for the forward path** — U1–U6 edit or add tracked files only. Scoped exception: ProposedAction A2's *rollback* deletes a directory and requires operator approval at execution time |
| VIII. Explicit Safety Modes | **`careful` + `freeze-scope`** (rev 1 declared freeze-scope alone). Modes are additive; the self-declared large blast radius is a literal `careful` trigger |
| IX. Git-Friendly Persistence | **Satisfied** — the new manifest uses sorted keys for stable diffs; the file allowlist in §6 is exact, with no "or a sibling manifest" discretion |
| X. Agent Context Efficiency | Satisfied — one script, two fixtures, two docs |
| XI. Merge Commit History (NON-NEGOTIABLE) | **Load-bearing here** — squash or rebase would collapse the U1 red commit into its green successors and destroy the Principle II evidence for this shipment. Ship MUST use a merge commit and halt with P-009 if the strategy is squash/rebase (R7) |

## 8. Known inconsistencies recorded, not fixed (out of authorized surface)

- **`cmd/` filter asymmetry** — the `cmd/` branch applies no `_test.go` or `/testdata/`
  exclusion, so `cmd/**` tests are scanned while `internal/**` tests are not. Recorded so a
  future author does not read U1's assertions as evidence the branches behave alike.
- **`ci.yml` gate wiring** — the `Run retired-architecture gate` step carries
  `continue-on-error: true` (verified, L263) and can never fail CI; enforcement runs solely
  through the blocking `--self-test` step, which internally re-runs the repo scan. The
  "merge-blocking" framing holds only via that path. `ci.yml` is an **excluded** surface
  (`F4F4A959`/`F47DB9A9`) and must not be opened here.
- **`mask_go_non_code` clone** — the same ~88-line lexer exists verbatim in
  `check-write-path-precondition.sh`; U3's fixtures prove only this copy. Deferred to stash.
## 9. Plan Hardening

Invoked because §front-matter declares `Requires plan hardening: yes` (P-006).

### 9.1 Risk triggers present

| Trigger | Present | Detail |
|---|---|---|
| Contract / gate-surface change | **yes** | The scan-scope predicate is the contract of a merge-blocking CI control |
| Compliance-sensitive behavior | **yes** | U-E1a is the mechanical enforcement of the D6 architecture-retirement decision |
| Blast radius beyond the slice | **yes** | A false positive blocks **every** future PR, including ones owned by excluded stash entries |
| Irreversible / destructive action | **partly** | Gate behavior is fully revertible. **But** committed fixture content is permanent in git history, which `secret-scan-history.yml` scans — so a credential-shaped fixture is not remediable by a later delete (§9.3 A2) |
| External integration | no | Local script, `git ls-files` + Python only |
| Governing-contract change | **yes** | D6 self-declares GOVERNING and normatively fixes the gate scope; it must be amended (D6c) in the same change (U6) |

### 9.2 Protected invariants (must hold after every unit)

1. **I-A — the gate never scans itself.** No path under `scripts/` is ever selected.
   Violating this makes the fixtures self-match and the gate permanently red.
2. **I-B — test/fixture exclusions survive the broadening.** `/testdata/` and `_test.go`
   remain excluded *inside* `internal/**`. Dropping these while widening the prefix would
   flag `internal/config`'s own legacy migration fixtures — the second reason D6a existed.
3. **I-C — gate independence.** `check-write-path-precondition.sh` is not touched, so
   `4989A42D` condition (b) stays a single-script, unambiguous measurement.
4. **I-D — vocabulary frozen.** `forbidden_parts` is unchanged; only scope and dispatch move.
5. **I-E — the real tree stays green.** `check-retired-architecture.sh` (no args) exits 0
   at every commit boundary from U4 onward.
6. **I-F — no false green.** A Go fixture must never be adjudicated by a TOML engine. A
   REJECT fixture that "passes" via a TOML parse error proves nothing and is a defect.

> **I-B guards a real failure mode, but is forward-looking today.** The deliberation's
> empirical probe (§2.3) preserved the existing filters. A broadening that widened the
> prefix *and* loosened the filters would scan `internal/config/testdata/**`, which would be
> known-dirty by design. U1 must assert the exclusion edge, not only the inclusion edge.
>
> **I-B scoping and rationale correction (review).** I-B holds for `internal/**` only. The
> `cmd/` branch early-returns *before* those filters, so `cmd/**` tests are scanned today
> (verified). I-B must not be asserted globally or U1 is permanently red (§8).
>
> **Rationale corrected**: rev 1 justified I-B by claiming a broadening that dropped the
> filters "would flag `internal/config`'s own legacy migration fixtures." Verified false —
> **no `testdata/` directory exists anywhere under `internal/`** today. I-B is therefore
> **forward-looking**, not currently load-bearing, and the `testdata` half of U1 assertion 2
> asserts over an empty set. Only the `_test.go` half is non-vacuous today. Keeping both is
> correct as future-proofing, but I-B must not be ranked above the assertions that do
> real work now (assertions 1 and 4).

### 9.3 Risky actions (strict-safety vocabulary)

**ProposedAction A1** — widen the scan predicate and pathspec (U4).
- **ActionRisk**: *medium*. Failure mode is a red CI gate on unrelated PRs, detected
  immediately at PR time, not silently.
- **Pre-conditions**: U1 red proof captured; U2/U3 Go-scanner coverage landed;
  `internal/apperr` verified free of retired Kinds; broadened probe verified exit 0.
- **ActionResult (expected)**: `check-retired-architecture.sh` exit 0; `--self-test` exit 0.
- **Rollback**: revert the three-line edit. No coupled state, no migration, no data change.
  Rollback restores the exact pre-shipment behavior and re-opens `4989A42D` condition (b).

**ProposedAction A2** — add `.go` fixtures containing retired tokens (U3).
- **ActionRisk**: *medium* (raised from *low–medium* in rev 1).
- **Guards**: they live under `scripts/testdata/`, which two independent mechanisms keep
  outside the scan (pathspec + fail-closed predicate); the Go toolchain ignores `testdata/`
  for build/vet/lint.
- **Guards the rev 1 assessment missed**: `gofmt -l .` and `goimports -l .` are filesystem
  walks that **do** descend into `testdata/`, and `gitleaks detect` is a **blocking** CI
  step with no `.gitleaksignore` in the repo. Fixture content is constrained accordingly
  (U3).
- **ActionResult (expected)**: `go build`, `go vet`, `gofmt -l .`, `goimports -l .`,
  `gitleaks detect`, `--self-test` all clean.
- **Rollback — CORRECTED**: deleting the fixture directory rolls back *gate behavior* but
  **not history exposure**. `secret-scan-history.yml` scans full history, so a merged
  credential-shaped fixture is **not** remediable by a later delete; it would require a
  history rewrite or a `.gitleaksignore` fingerprint. Hence gitleaks runs **before the
  first commit** (§6), while renaming is still cheap. Directory deletion is itself a
  destructive action requiring operator approval at execution time (Principle VII).

**ProposedAction A3** — amend governing artifacts (U6).
- **ActionRisk**: *low–medium*. D6 is self-declared **GOVERNING** and wins over any plan,
  so amending it changes the authoritative contract — but leaving it unamended would put
  the shipped script in direct contradiction with it.
- **Guard**: dated amendment **D6c**; annotate, do not rewrite — the D6a rationale must
  remain readable as history. Enumerated file list, never a repo-wide grep (which would
  reach excluded surfaces).
- **Rollback**: revert the docs commit.

### 9.4 Added verification depth

Beyond §6, each unit carries an explicit falsification step:

- **U1** — must be observed **FAILING** before U4/U5 land, and the failure text must name
  both the uncovered `internal/` packages and the unscanned `config.toml.example`. A U1
  that passes on arrival is a defective assertion (it asserts something already true) and
  must be rejected, not accepted as convenient. Capture the red output in the backlogit
  task record (Principle V).
- **U2** — confirm a Go REJECT fixture fails with the `scan_go` message naming the retired
  token and Go identifier, **not** a TOML parse error (invariant I-F).
- **U3** — invert one fixture's manifest expectation locally and confirm `--self-test`
  fails. A fixture harness that cannot fail proves nothing. Do not commit the inversion.
  Run `gitleaks detect --source . --no-git` **before** the first commit containing fixtures.
- **U4** — after the edit, re-run the scan and confirm the selected path count **increases**
  by exactly the 8 known non-test Go files. A broadening that changes the predicate but not
  the `git ls-files` pathspec would silently select nothing new; U1 assertion 4 (non-empty)
  plus this count check catches that.
- **U5** — confirm `scan_toml` is actually invoked on `config.toml.example` (e.g. it now
  appears in a scanned-path trace), not merely that the gate still exits 0 — exit 0 was
  already true when the file was never read.
- **U6** — verify the **enumerated** sites in the two named files are updated. Do **not**
  run an unbounded repo-wide grep: it reaches `.github/workflows/ci.yml`
  (excluded `F4F4A959`/`F47DB9A9`) and the 2026-09-07 pathsafe plan (excluded `1C6C3B46`).


### 9.5 Monitoring and operator checkpoints

- **No operator checkpoint is required.** This slice depends on none of the five open
  operator decisions (Q3/H3, Q6/H4, H5, H6, Q7) and introduces no new one.
- **Post-merge signal to watch**: the first unrelated PR after merge. If U-E1a fails on it,
  that is R1 materializing. Remedy is a targeted, commented exclusion — **not** a scope
  rollback (§2 anti-goal).

### 9.6 Review-gate capability risks (P-012)

- `agent-intercom` is **not installed** in this workspace; operator visibility is
  Copilot-CLI-only. Review dispatch and verdicts are recorded in-artifact rather than
  broadcast. This is the intended configuration, **not** a degraded-review condition.
- The backlogit **MCP surface is unavailable** in this session; all backlog operations use
  the registry-declared **CLI fallback** (`DEGRADED_MODE`). This affects backlog writes
  only, not review quality, and every CLI fallback used is declared in
  `.autoharness/backlog-registry.yaml`.
- No P-012 condition blocks the plan-review gate.

### 9.7 Learnings consulted

`docs/compound/` searched (see deliberation §6). Two process-level analogues applied:
mechanical-gate self-matching (→ **I-A**, R2) and empirical-verification/scope
classification (→ deliberation §2.3 probe and the §2.4 git-history correction). No
topic-specific precedent for U-E1a exists; this cycle should produce one.

---

## 10. Review Finding Disposition (rev 1 → rev 2)

Seven-persona panel: Architecture Strategist, Correctness Reviewer, Security Lens
Reviewer, Scope Boundary Auditor, Constitution Reviewer, Schema-CLI-Docs Coupling
Reviewer, plus the Learnings Researcher. Every high-confidence finding is dispositioned.

### Resolved in rev 2

| # | Sev | Finding | Resolution |
|---|---|---|---|
| 1 | **P0** | `scan_path` dispatches on `Path.suffix`; `config.toml.example`'s suffix is `.example`, so the **TOML engine never runs on the real tree**. Found independently by Correctness and Coupling | **New unit U5** fixes dispatch; **U1 assertion 5** pins it red-first. Verified the file is token-free, so the fix lands green |
| 2 | **P0** | U6 (rev 1 U5) missed the **governing** D6 artifact, which normatively fixes gate scope and "wins over any plan" | **U6 extended** to add dated amendment **D6c**; discharges the L377–378 plan-mirrors-D6 coupling invariant |
| 3 | **P0** | rev 1's docs enumeration was incomplete — L361-364, L365-376, L515-519 are more explicit normative scope statements than the three lines named | **U6 enumerates all six sites** in the plan file |
| 4 | **P0** | Go fixtures routed through the TOML-only `engines` list ⇒ **false green**: REJECT fixtures "pass" via TOML parse errors, proving nothing | **U2 rewritten** to require per-kind discovery + suffix-based engine dispatch; acceptance requires the `scan_go` message naming token and identifier. New invariant **I-F** |
| 5 | **P0** | Non-recursive `fixture_glob` + bare-`path.name` manifest keys cannot see or reconcile `retiredgo/*.go` | **U2** adds a second suite descriptor and a pinned sibling manifest `retiredgo-manifest.json` |
| 6 | **P1** | U1's "no `_test.go` selected" assertion is **false today** — `cmd/` early-returns before the filters, so 3 `cmd` test files are selected. Would be permanently red | **U1 exclusion assertions scoped to `internal/`**; asymmetry recorded in §8 as known-not-fixed |
| 7 | **P1** | U1 hardcoded `internal/copilotprobe/`, a self-documented **DISPOSABLE spike** ⇒ future permanent red; also contradicted U1's own "data-driven" acceptance | **U1 assertion 1 made structural** — every tracked non-test `.go` under `internal/`, derived from `git ls-files`, no package allowlist |
| 8 | **P1** | `gitleaks detect` is a **blocking** CI step, no `.gitleaksignore`; fixtures could trip it and are **permanent in history** | **U3 fixture-content constraints**; gitleaks added to §6 and run **pre-first-commit**; A2 rollback corrected |
| 9 | **P1** | `gofmt -l .` / `goimports -l .` are filesystem walks that **do** descend into `testdata/` — the "Go ignores testdata" argument doesn't cover them | **U3** requires gofmt/goimports-clean parseable fixtures; both added to §6; Principle I row corrected |
| 10 | **P1** | U6 (rev 1 U5) acceptance was an **unbounded repo-wide grep**, reachable to `ci.yml` (excluded `F4F4A959`/`F47DB9A9`) and the 2026-09-07 pathsafe plan (excluded `1C6C3B46`) — a route to an exclusion breach | **Replaced with an enumerated two-file list**; §9.4 U6 explicitly forbids the repo-wide grep |
| 11 | **P1** | rev 1 U3 breached the 2-hour rule (4 files, ~5 functions, 4 scenarios) and mixed two domains | **Split** into U2 (harness) and U3 (fixtures); rev 1 U4's characterization suite deferred |
| 12 | **P1** | "Mirror the writepath pattern" + "reuse the existing manifest shape" are **incompatible** — writepath has no manifest, it uses filename conventions | **U2 pins the manifest contract** and records the correction |
| 13 | **P2** | §7 Constitution table used AGENTS.md numbering, mislabelled two rows, omitted V, VI, IX, X, XI | **§7 rewritten** against `constitution.instructions.md`; all 11 principles mapped |
| 14 | **P2** | Squash/rebase would collapse the U1 red commit and destroy the Principle II evidence | **R7 + Principle XI row**; Ship must use a merge commit and halt P-009 on squash/rebase |
| 15 | **P2** | `freeze-scope` alone under-selected for the declared blast radius | **`careful` + `freeze-scope`** declared (modes are additive) |
| 16 | **P2** | Coverage sequenced **after** the broadening, inverting the deliberation's own §2.5 argument | **Reordered** U1 → U2 → U3 → U4 |
| 17 | **P2** | `run_repo_scan` never asserts a non-empty selection ⇒ silent fail-open | **U1 assertion 4** |
| 18 | **P2** | rev 1 U1 declared "self-test path only" but must touch `run_repo_scan` to be data-driven | **U1 surface corrected**; `select_repo_paths()` extraction named |
| 19 | **P2** | Principle IV: the §2.3 probe wrote outside the workspace tree | **Disclosed** in deliberation §2.3, artifact deleted, future probes constrained |
| 20 | **P2** | R1 never named the newly-scanned `internal/pathsafe/*` as the surface of 4 excluded entries incl. critical `700B41CE` | **R1 extended** with named downstream consumers |
| 21 | **P2** | `ci.yml`'s gate step is `continue-on-error: true`; only `--self-test` blocks | **Recorded in §8** as known-not-fixed (excluded surface) |
| 22 | **P3** | §6 allowlist hedged "or a sibling manifest" — non-deterministic under freeze-scope | **Pinned** to `retiredgo-manifest.json` |
| 23 | **P3** | Script header L16-19 self-test contract would ship stale | **Assigned to U2**; L4/L15 stale claims assigned to U5 |
| 24 | **P3** | D6b anti-accident-only threat model risked being read as discharged by D-5 | **U6** records the threat model is unchanged; H5 not partially discharged |

### Deferred to the stash — real findings, outside this shipment's authorized surface

Per the run contract, these are captured for a **future** Stage cycle and do **not**
expand this shipment:

| Sev | Finding | Why deferred |
|---|---|---|
| P1 | `mask_go_non_code` (~88 lines) is a **verbatim clone** in `check-write-path-precondition.sh`; U3 proves only one copy | That script is excluded surface (`BF5DE670`) and `4989A42D` requires the gates stay independent. **Record in §8 only — do NOT create a stash entry**: a new capture against an excluded entry's own surface would duplicate it and read as de-facto triage |
| P2 | Full `mask_go_non_code` characterization suite (line/block comment, escaped quote, raw backtick) | Gold-plating for this slice (Scope Boundary P2); ≥3 more files. Stash capture permitted — not an excluded surface |
| P2 | `split_identifier` misses plural/acronym-tail forms — `ChannelIDs` → `[channel, i, ds]`, also `TeamIDs`, `IPCNames`, `SocketModes` | Changes detection **vocabulary**, not scope (anti-goal I-D). Stash capture permitted |
| P2 | Struct-tag masking false negative — `` `toml:"channel_id"` `` is invisible to both engines | Needs its own deliberation and exceeds this slice's granularity. Stash capture permitted |
| P2 | `SCAN_SCOPE` single-source-of-truth refactor (pathspec and predicate encode scope twice) | Behavior-preserving refactor beyond condition (b). Stash capture permitted |
| P2 | Extract the Python engine to `scripts/lib/` so it is importable/unit-testable | Structural refactor; would enlarge blast radius here. Stash capture permitted |
| P2 | `ci.yml` redundant `continue-on-error: true` gate step | **Excluded surface** (`F4F4A959`/`F47DB9A9`). **Record in §8 only — do NOT create a stash entry** |
| P3 | `cmd/` branch filter asymmetry (no `_test.go`/`testdata` exclusion) | Changes `cmd/**` behavior; not required by condition (b). Stash capture permitted |

### Open advisories carried to Ship (P2/P3, none blocking)

- Three excluded entries — `EF9352FB`, `AD0D9D1F`, `9D45E62E` — have no file surface
  declared in either artifact, so U6's non-intersection with them is asserted by silence
  rather than verified. Ship should state their surfaces and confirm non-intersection
  before U6 executes.
- §9.4's U4 falsification step says "increases by exactly the 8 known non-test Go files."
  Re-derive that count at execution time rather than trusting it: `internal/copilotprobe`
  self-documents as a **disposable** spike, so a correct broadening could legitimately
  produce a different delta if the package is removed or another lands.




---
title: "Implementation Plan — Residual Hardening Round 2 (011-F)"
date: 2026-09-06
status: reviewed
revision: 5
phase: residual-hardening
feature: 011-F
source_deliberation: "docs/decisions/2026-09-06-intercom-go-residual-hardening-round2-deliberation.md"
resolves_stash: [A0A2D049, 8D953C4B, 798002CB, 8472E0A1, E89E5D00, DA945722, 707FE72B, 11ECB954, AFE0E95C, 309FBF5A, 6E701953]
partially_resolves_stash: [9F9A3CB2, BF5DE670]
depends_on: 010-F
plan_review_attempt: 5
revision_note: "Revision 5 restricts the un-ignore denylist to HEAD-guaranteed paths (it must never depend on droppable stowaway content) and gives the full 52-line .gitignore stowaway hunk an owning unit."
---

# Implementation Plan — Residual Hardening Round 2 (011-F)

> **Revision 3.** Revision 1 was returned **FAIL** (4×P0, ~20×P1). Revision 2
> cleared all P0s but was returned **FAIL** on two P1s. Revision 3 remediates
> both plus supporting findings; see the "Review Remediation Ledger".

## Problem Frame

Thirteen implementable deferred-scope-expansion findings accumulated across
shipments `005-S`–`009-S`, plus one operator-authorized working-tree
carryover. All are residue against *shipped* code; none is new product
capability. Governing decisions D-A…D-I are in the source deliberation, as
amended by its Addendum (D-I revised, D-E re-scoped).

Blast radius: `internal/copilotprobe`, `internal/pathsafe`, `internal/config`,
`.github/workflows/ci.yml`, `.golangci.yml`, `scripts/`, `.gitignore`,
`start.ps1`, `docs/`.

## Requires plan hardening

**yes** — four signals: security-sensitive behaviour, contract change,
external dependency, CI topology change. Hardened 2026-09-06 (see below).

## Declared Safety Mode (Constitution VIII)

Four VIII triggers are present: real shell-command execution (`011.002-T`),
a change to a **required** CI gate (`011.009-T`/`011.010-T`), supply-chain
installer repinning (`011.013-T`), and manipulation of uncommitted operator
work (sub-epic F stowaway protocol).

**Operating posture: `freeze-scope` + `careful`.**

* `freeze-scope` — the unit list below is closed. No unit may grow beyond its
  stated ACs; anything discovered mid-execution is captured to the stash
  under P-021 C2, never absorbed.
* `careful` — every unit touching a containment control
  (`011.002-T`, `011.004-T`…`011.008-T`, `011.011-T`) is test-first with a
  **red phase observed before implementation**, and no containment
  comparison may be relaxed without a paired negative test.

## Constitution Check

Required by Governance §Compliance review. Maps every unit against
`.github/instructions/constitution.instructions.md`.

| Principle | Status | Notes |
|---|---|---|
| I. Safety-First Go | PASS | No `unsafe`, no new goroutines, no panics introduced |
| **II. Test-First (NON-NEGOTIABLE)** | PASS with 3 documented deviations | Every production-Go-changing unit is test-first with an observed red phase. **Documented deviations:** `011.007-T` is a zero-verdict-change refactor where a red phase is impossible by construction — it is **characterization-first** instead; `011.001-T`, `011.009-T`, `011.013-T` are `*Approach:* configuration` units that change no production Go and are **fixture/verification-first**. Docs-only units (`011.003-T`, `011.005-T`, `011.019-T`, `011.020-T`) change no code. **PowerShell units `011.011-T`/`011.016-T`–`011.018-T` are test-first via the existing Go integration harness that shells out to `pwsh`** (`tests/integration/output_path_guard_test.go`) — **not** Pester; no Pester dependency exists or is added |
| **III. Workspace Isolation** | EXCEPTION (bounded, expiring) | `database.path` rule-7 divergence — see `011.003-T`. Exception is dated, bounded, and **expires at the first persistence write path** |
| **IV. CLI Workspace Containment (NON-NEGOTIABLE)** | PASS — no current breach; one named residual | **Measured:** zero filesystem-write call sites exist under non-test `internal/**` or `cmd/**`, so no path is created/modified/deleted outside the tree today. `011.004-T` converts that measurement into a mechanical CI gate so the finding cannot silently expire. **Named residual (recorded per `011.008-T`):** `pathsafe` treats an unexpanded `$VAR`/`%VAR%` segment **literally**, which resolves *inside* the root and is therefore IV-compliant; the actual expansion vector is a **caller expanding before calling pathsafe**, which pathsafe cannot control. This residual is accepted and tracked — `011.008-T` adds a doc note and tests pinning literal treatment, and deliberately adds **no refusal class** (a refusal would falsely reject real directories such as `$Recycle.Bin`) |
| V. Structured Observability | PASS | No logging surface changed; `011.017-T` explicitly forbids echoing loaded values |
| VI. Single Responsibility | PASS | 20 single-purpose units; revision 1's multi-item bundles are decomposed |
| **VII. Destructive Command Approval (NON-NEGOTIABLE)** | PASS with controls | `011.002-T` gates real shell execution: default-deny, exact-match `"1"`, CI asserts unset, blast radius confined to `t.TempDir()`. `011.015-T`'s stash return requires a verified backup first |
| **VIII. Explicit Safety Modes** | PASS | Declared above |
| IX. Git-Friendly Persistence | PASS | Append-only `.gitignore`; no binary artifacts |
| X. Context Efficiency | PASS | — |
| XI. Merge Commit History | PASS | Ship-owned; unaffected |

**Justified violations on record:** exactly one — the Principle III
`database.path` exception in `011.003-T`. It is bounded, dated, carries the
rejected simpler alternative, and expires on a concrete trigger.

## Execution Units

Ordering is normative. Sizes/complexities are two independent axes.

---

### Sub-epic A — Foundation and live safety exposure

#### 011.001-T · Close `.gitignore` preserve-only gaps (lands first)

Partially resolves `9F9A3CB2`. Four preserve-only paths are confirmed
**NOT ignored**: `.claude/instructions.md`,
`.github/copilot/settings.local.json`,
`.autoharness/gates/pipeline-topology-force-audit.log`,
`.backlogit/hooks_queue.jsonl`. Two carry machine-local command/environment
payloads.

These are **pure additions** — they cannot violate I6's ordered-subsequence
predicate and cannot trip any un-ignore rule. They are sequenced **first**,
not last, so machine-local state cannot be staged during the other 19 units.
This unit is **shipment work and is NOT droppable** by the stowaway protocol.

*Approach:* configuration. Width: ignore policy. Size `S`, complexity `low`.

**Acceptance criteria**

* All four paths report ignored via `git check-ignore -v` using **bare paths
  with no trailing slash** (a trailing slash yields a spurious blank-line match).
* `scripts/check-gitignore-append-only.sh` passes.
* `git ls-files -i -c --exclude-standard` is **empty** — proving no tracked
  file became ignored. (Revision 1 cited `git ls-files`, which cannot detect this.)
* Committed as its own commit, separate from any stowaway line.

#### 011.002-T · Gate live credential-consuming probe tests behind explicit opt-in

Resolves `A0A2D049`. `internal/copilotprobe/testhelpers_test.go:77` skips only
when `Start()` *fails*, so ambient credentials silently authorize live SDK
calls **and real shell execution** (`permission_test.go`).

*Approach:* test-first. Width: tests only. Size `S`, complexity `low`.

**Acceptance criteria**

* Opt-in is `INTERCOM_LIVE_SDK_TESTS` with **default-deny value semantics**:
  enabled only on exact `"1"` (or `strconv.ParseBool` true with **deny on
  parse error**). `=0`, `=false`, `=off` and any unparseable value **disable**.
* A package-level `TestMain` **fail-closed backstop**: when the var is unset it
  **sanitizes credential environment variables and sets a package-level denial
  flag consulted by the _test-side_ helpers** (`newProbeClient` /
  `newProbeClientManualLifecycle` in `testhelpers_test.go`), while **still
  calling `m.Run()`**. **The flag and its checks live entirely in `_test.go`
  files** — this unit is `Width: tests only` and must not add denial logic to
  `client.go`, which is production code in a package designated disposable at
  C3 (J10).
  — so the gate-helper unit test executes and the package continues to
  contribute compile/behaviour coverage to `go test ./...`. It must **not**
  skip or exit the whole package (revision 2 left the mechanism unspecified;
  a skipping backstop would have blocked its own verification AC and masked
  regressions from the new Windows job).
* Verified by a **direct unit test of the gate helper's early return**,
  executed with the variable **unset** (a test that has already called
  `t.Skip` cannot assert about itself).
* **Blast radius bounded (Principle VII/IV):** the probe's permission handler
  is **deny-by-default** — it rejects any invocation that does not exactly
  match an allowlisted command string. Revision 2 required commands to be
  "enumerated", which is unsatisfiable: `permission_test.go` prompts a **live
  model**, so the executed command is model-determined, and the current harness
  approves unconditionally. A mechanical allowlist is the only enforceable form.
* Execution is confined to `t.TempDir()`. Recorded explicitly: the probe
  deliberately runs **outside** the workspace tree, which is safer than rooting
  a live shell session in the repo checkout.
* CI asserts `INTERCOM_LIVE_SDK_TESTS` is **unset in every workflow job**, and
  it appears in no committed config file.
* With the var unset, `go test ./...` passes with zero network or credential access.

---

### Sub-epic B — Containment correctness

#### 011.003-T · Record a bounded, expiring Constitution Check exception for `database.path`

Resolves `8D953C4B`. **Revision 2 replaces revision 1's branch (a)/(b) choice
with branch (c)** — see the deliberation Addendum (D-I revised).

Branch (a) ("route through `pathsafe.NewRoot`/`Resolve`") was found
**infeasible**: `pathsafe.normalize` rejects *every* absolute candidate, and
`NewRoot` requires an **already-existing directory**, while `database.path` is
a not-yet-existing file whose rule 7 deliberately permits absolute paths as a
visible operator privilege. Implementing (a) would reject legitimate in-root
absolute paths and force a new `pathsafe` capability — colliding with J2.

Branch (b) alone was found unsafe: an unbounded exception could outlive its basis.

**Branch (c):** record a **dated, bounded exception that expires at the first
persistence write path**, and register the containment requirement as a
mandatory precondition on the D7/C4 persistence work.

Principle IV is **not** breached today: measured, nothing opens or writes
`database.path`. The divergence is **latent**, and this unit ensures it cannot
become live silently.

*Approach:* documentation. Width: docs. Size `S`, complexity `low`.

**Acceptance criteria**

* A dated Constitution Check exception records: the principle (III), the
  divergence, **the rejected simpler alternative (branch (a)) and why it is
  infeasible**, and its bound.
* The exception's expiry trigger is **the same trigger** as `011.004-T`'s
  mechanical gate, so it cannot outlive the arrival of a real write path.
* The D7/C4 persistence work carries a precondition that `database.path` gains
  containment **before** any code opens it.
* No production code changes; `internal/config` behaviour is untouched.

#### 011.004-T · Make the write-path precondition a mechanical CI gate

**Partially resolves** `BF5DE670` as rescoped by D-B (the mitigation itself is
deferred to the C4–C6 trigger, so this is not a full resolution).

Add a **new, dedicated** script (e.g. `scripts/check-write-path-precondition.sh`)
wired into CI alongside the existing gates.

**Revision 3 — it must NOT be bolted onto `scripts/check-retired-architecture.sh`.**
That script's narrowing to `internal/config/**` is the measured, still-**UNMET**
completion condition (b) of retained tracker `4989A42D`, and broadening it is an
explicit anti-goal of this plan. Adding an `internal/**`-wide scan to that file
would leave it scanning `internal/**` for one purpose while nominally narrowed
for another, making `4989A42D`'s completion condition no longer cleanly
measurable next cycle.

*Approach:* test-first, with its own fixture/self-test harness. Width: script
+ fixtures. Size `M`, complexity `medium`.

**Acceptance criteria**

* **The consolidated risk register is produced.** This is `BF5DE670`'s primary
  ask and revision 2 referenced it without any AC creating it. One register
  unifies GO-14 (write-through-dangling-symlink), the Resolve→use TOCTOU
  window, and SEC-5 (`EvalSymlinks` ignores hardlinks), each naming its current
  mitigation status and the condition that would force mitigation.
* The three acceptances stay in `internal/pathsafe` **godoc** (where the author
  of the first write path will read them) and cross-reference the register —
  detail is **not** relocated out of godoc.
* The detector uses **qualified selectors** (`os.WriteFile`, `os.Create`,
  `os.OpenFile`, `os.Remove`, `os.RemoveAll`, `os.Rename`, `os.Mkdir`,
  `os.MkdirAll`, `os.Symlink`, `os.Chmod`, `os.Truncate`, `(*os.File).Write*`,
  `io.Copy`) plus embedded-store constructors (`sql.Open`, `bbolt.Open`).
  Revision 2's bare alternation would have matched any `Create`/`Remove`/`Rename`
  identifier — including the SDK's `CreateSession` — producing false positives
  that block legitimate C3 work.
* **Scope: all non-test Go files in the module — `internal/**` AND `cmd/**`.**
  `cmd/intercom` is the only `main` package and already imports
  `internal/config`, so it is the most likely site for persistence wiring; a
  gate scoped to `internal/**` alone would let the `011.003-T` exception
  silently outlive its basis.
* **Dropped in revision 3:** the three "indirect classes" (`os/exec` with a
  working directory, `database/sql` open *against `database.path`*, SDK
  `WorkingDirectory` assignment). None is a filesystem write primitive, none is
  covered by `BF5DE670`, and the first and third would pre-constrain C3 adapter
  work that this plan's anti-goals exclude. (`sql.Open` is retained above as a
  *store constructor*, which is a genuine write primitive and decidable without
  argument provenance.)
* A fixture containing a write primitive **fails** the gate; the current tree
  **passes**. Both asserted via `--self-test`.
* The failure message names the register and `011.003-T`'s Constitution
  exception, and states the **retirement procedure** for the legitimate arrival
  of a write path at C4–C6.
* Zero behavioural change to `internal/pathsafe`.
* **Anti-goal:** no TOCTOU/hardlink mitigation mechanism is added.

#### 011.005-T · Document the `ContainsDotDotSegment` export-surface limits

Resolves `798002CB`. **Docs-only** — revision 1 declared a combined
`code/docs` width, a NON-NEGOTIABLE Width Isolation violation, and permitted
an unbounded optional rename.

*Approach:* documentation. Width: docs. Size `XS`, complexity `trivial`.

**Acceptance criteria**

* The doc comment explicitly enumerates the three absent guarantees (no
  canonicalization, no absolute-path rejection, no symlink resolution) and
  names `Root.Resolve` as the containment API.
* **No rename, no un-export, no signature change** in this unit — under
  branch (c) the rule-7 call site remains, so the function is not dead API.
* Zero verdict change; no test outcome flips.

#### 011.006-T · Add pathsafe regression tests for reachable uncovered branches

Resolves `8472E0A1` items (b), (c), (e).

**Revision 2 correction:** revision 1 required tests for two branches that are
**unreachable without a test seam** — `NewRoot`'s `os.Stat` failure (EvalSymlinks
already stats every component immediately prior) and `checkSymlinkEscape`'s
filesystem-root fallback (the code's own comment declares it unreachable in
practice). Requiring "fail before, pass after" on those was unsatisfiable.

*Approach:* test-first. Width: tests only. Size `S`, complexity `low`.

**Acceptance criteria**

* Item (c) — a config-layer regression test pins that a workspace path pointing
  at a **regular file** is rejected, framed explicitly as an integration wiring
  pin over `pathsafe.NewRoot` (not a duplicate of `TestNewRootRejectsFilePathAsWorkspaceRoot`).
* Items (b) and (e) — both branches are **recorded as documented-unreachable
  with a coverage exclusion** and a one-line rationale. **Revision 3:** the
  "authorize a testability seam" alternative is **removed** — adding an
  injectable stat to `internal/pathsafe` would be a production-code change
  inside a declared tests-only unit (Width Isolation, NON-NEGOTIABLE) and
  would contradict this unit's own "no production behaviour changes" AC.
* No production behaviour changes.

#### 011.007-T · Collapse `NewRoot`'s repeated error branches

Resolves `8472E0A1` item (a).

**Revision 2 sequencing fix:** revision 1 put this in the same subtask as a
test pinning discriminability of the very branches it collapses. Refactor is
now decided **before** tests are written against the post-refactor shape.

*Approach:* test-first (characterization first). Width: production code.
Size `S`, complexity `low`.

**Acceptance criteria**

* The three identical `apperr.Wrapf` branches collapse into one unexported helper.
* **Zero verdict change**: every existing `internal/pathsafe` test passes
  unmodified, and error `Kind` plus `errors.Is`/`errors.As` discriminability
  is preserved and asserted.
* No exported surface change.

#### 011.008-T · Pin that path candidates must be raw, unexpanded strings

Resolves `8472E0A1` item (g).

**Revision 3 — over-reach corrected.** Revision 2 escalated this to a
*production refusal* of `$VAR`/`${VAR}`/`%VAR%` in the containment API. Three
reviewers independently rejected that:

* **Not authorized.** `8472E0A1`(g) asks for "an explicit doc note or test in
  pathsafe confirming candidates must be raw, unexpanded strings" — a P3
  advisory. No in-scope entry authorizes a verdict change in `pathsafe`.
* **Principle IV is not engaged.** IV refuses paths that *resolve above or
  outside* the root via expansion. A literal, unexpanded `$VAR` segment
  treated literally resolves **inside** the root — it is a directory named
  `$VAR` — so literal treatment is already IV-compliant. The real vector is a
  **caller** expanding *before* calling pathsafe, which pathsafe cannot control.
* **It would cause false rejections.** `$Recycle.Bin`, `$WinREAgent` and
  `$SysReset` are real Windows directories, and `%` is a legal POSIX filename
  character. A refusal predicate becomes a hard availability defect with no waiver.
* It contradicted J2 while its own AC re-asserted J2.

The revision-1 concern (a NON-NEGOTIABLE principle must not be closed by a
trivial doc note) is answered instead by **recording the residual caller-side
gap explicitly in the Constitution Check**, which is what IV actually requires
— rather than silently.

*Approach:* test-first. Width: tests + docs. Size `XS`, complexity `trivial`.

**Acceptance criteria**

* A doc note in `pathsafe` records that candidates MUST be raw, unexpanded
  strings, and that expansion is the **caller's** responsibility.
* A test pins that an unexpanded `$VAR` segment is treated **literally**
  (a directory named `$VAR`), proving no accidental expansion occurs.
* A positive test proves legitimate literal `$`/`%` filename segments
  (e.g. `$Recycle.Bin`) still validate.
* **Zero verdict change** (J2). No refusal class is added.
* The residual caller-side expansion gap is recorded in the Constitution Check
  as an explicit, named residual — not left implicit.

> **Dropped from revision 1:** item (f) ("GOOS-gate `hasWindowsDrivePrefix`
> consistently with `pathsafe.isRooted`"). Three reviewers independently found
> this is **not** a no-op refactor but a **security verdict loosening**:
> `hasWindowsDrivePrefix` lives in `internal/config/validate.go` and guards
> portable config text, so GOOS-gating it would make `cli_path = "C:evil"`
> *accepted* on Linux. The originating stash entry itself calls it "an
> inconsistency not a defect". Replaced by a doc note in `011.020-T` recording
> the divergence as tracked-and-deliberate. Item (d) (`os.Stat` argument order)
> and item (h) (file organization) are likewise deferred to `011.020-T` as
> documented notes — revision 1's "fix it **or** document it" wording was an
> unfalsifiable escape hatch, and (h) had no definition of done.

---

### Sub-epic C — Platform verification coverage

#### 011.009-T · Add an advisory `windows-latest` job

Jointly resolves `E89E5D00` + `DA945722` (one root cause: all jobs run
`ubuntu-latest`, so `isRooted` leading-backslash rejection,
`hasPathPrefix`/`pathEqual` case-folding, `stripUNCPrefix`, and the
junction-escape test **never execute**).

**Revision 2 risk fix:** revision 1 promoted this to a blocking gate in the
same unit that created it, while rating first-run failure **High** — putting a
high-probability red gate in front of all 19 other units. This repo already has
the right precedent: the `PIPELINE_TOPOLOGY_GATE_REQUIRED` advisory→required
toggle (`ci.yml:340-388`).

*Approach:* configuration. Width: CI config. Size `M`, complexity `medium`.

**Acceptance criteria**

* New `windows-latest` job runs `go test ./...`; **`-race` is omitted** with an
  inline comment citing
  `docs/compound/build-errors/go-race-requires-cgo-fails-closed-2026-09-03.md`
  (`windows-latest` has no cgo; `-race` fails closed with exit 2).
* All added actions pinned to **full 40-hex SHAs**; `persist-credentials: false`
  on checkout and job-level `permissions: contents: read`, matching every
  existing job.
* Job is in `ci-gate`'s `needs:` **and** its `results` expression from day one
  (J8), with its verdict wrapped in
  `continue-on-error: ${{ vars.WINDOWS_GATE_REQUIRED != 'true' }}` — present and
  wired, but **advisory** until proven green.
* Gated on `needs.changes.outputs.code == 'true'` like every other expensive job.
* **`011.002-T` is a hard precondition** — `go test ./...` on Windows would
  otherwise execute the live SDK probe tests.
* **Generated-file drift:** `ci.yml` is autoharness-generated
  (`# Generated by autoharness | Template: ci/ci.yml.tmpl`). Use the
  template-sanctioned OS-matrix escape hatch documented at `ci.yml:17-20`, or
  add an explicit `LOCAL DIVERGENCE — re-apply after regeneration` marker.

#### 011.010-T · Prove the Windows-only branches execute, then flip to required

*Approach:* test-first. Width: tests + CI config. Size `S`, complexity `low`.

**Acceptance criteria**

* A **durable assertion** replacing revision 1's one-time log inspection: on
  Windows, the junction-escape test and at least one `isRooted`/`stripUNCPrefix`
  branch **fail rather than skip** when their precondition is unmet, *except*
  for the single enumerated privilege case (junction creation requiring
  elevation), which skips with an explicit reason. **Revision 3:** revision 2's
  "unreachable for any non-privilege reason" had no operational definition;
  the skip conditions are now a closed, enumerated set.
* After a green run, `WINDOWS_GATE_REQUIRED` is set to `'true'`, making the job
  blocking. **This is a repository-configuration change under Constitution VII
  and the declared `careful` mode: it requires explicit operator confirmation**,
  and the flip is recorded in the shipment closure record.
* Rollback: unsetting the variable restores advisory mode with no code change.

---

### Sub-epic D — Script guard correctness

#### 011.011-T · Fix `OutputPathGuard.ps1` filesystem-root false rejection

Resolves `707FE72B`. `scripts/lib/OutputPathGuard.ps1:118` builds
`$repoRootWithSep = $resolvedRepoRoot + [Path]::DirectorySeparatorChar`
unconditionally. Traced: `Normalize-GuardPath` trims trailing separators only
when `fullPath.Length -gt pathRoot.Length`, so `C:\` and `/` keep theirs,
yielding `C:\\` / `//` at line 118 — `StartsWith` never matches and a
legitimate in-root path is **falsely rejected**.

**Revision 2 — CRITICAL safety addition.** Two reviewers independently found
that revision 1's only safety AC ("existing suite passes unchanged") is
inadequate: the existing suite has **no sibling-prefix case**. The separator
suffix is the *sole* mechanism preventing `C:\repo-evil` from matching root
`C:\repo`, so the two most natural fixes convert a fail-safe over-rejection
into a **genuine containment bypass that the existing suite passes cleanly**.

*Approach:* test-first, red phase mandatory. Width: PowerShell + tests.
Size `M`, complexity `medium`.

**Acceptance criteria**

* **MANDATORY negative test:** a sibling-prefix candidate (`<root>-evil`) is
  **rejected** — exercised at a **deeply nested root** and at a **top-level
  directory root** (e.g. root `C:\repo` vs candidate `C:\repo-evil`).
  **Revision 3 correction:** revision 2 required this at a *drive root*, which
  is unsatisfiable — with root `C:\`, the candidate `C:\-evil` is a legitimate
  **descendant** of `C:\` and must be *accepted*; the only way to make it
  reject is to reinstate the very bug this unit fixes. The sibling-prefix
  hazard class exists only for non-root roots.
* Positive regression: a legitimate in-root path at a **drive root (`C:\`)** and
  at a **POSIX root (`/`)** is accepted; both fail before and pass after.
  These roots get the **positive** case only.
* The test helper must **decouple `GuardPath` from `RepoRoot`** before the
  root-level fixtures can exist — today `runOutputPathGuard` derives the guard
  path from the root, which would require placing the script at the filesystem
  root (privileged).
* **UNC is excluded** from the failing-fixture requirement: `GetPathRoot(\\server\share\x)`
  excludes the trailing separator, so it is already trimmed and produces a
  correct single separator — a UNC fixture would pass *before* the fix,
  violating "fails before, passes after". A real UNC share is also not
  creatable on hosted CI.
* `..`-style escapes and all existing rejection behaviour preserved; the full
  existing guard suite passes.

---

### Sub-epic E — Supply-chain and boundary gates

#### 011.012-T · Add an un-ignore regression check via behavioural differential

Resolves `11ECB954`.

**Revision 2 — rule definition corrected.** Revision 1 defined the rule as
"a negation that re-includes a path already matched by an earlier ignore
pattern". That is the definition of **every functional gitignore negation** —
git only gives a `!` line effect when a preceding pattern already ignores the
path. Implementing it as stated would ban all effective negations, the exact
outcome D-E forbids, and reduce the paired AC to accepting only no-ops.

**Corrected rule — old-vs-new behavioural differential, using git's own
matcher rather than a gitignore engine reimplemented in bash** (which was also
the plan's clearest YAGNI violation):

> No path that is **actually ignored at base-ref** may become **un-ignored at
> head-ref**.

This is decidable, cannot be fooled by last-match-wins or the
parent-directory-exclusion rule, and correctly handles wildcard negations that
a per-line textual comparison cannot enumerate.

*Approach:* test-first, extending the existing self-test fixture harness.
Width: script + fixtures. Size `M`, complexity `medium`.

**Acceptance criteria**

* Implemented as a **two-part, provably non-vacuous** check:
  1. **Canonical sensitive-path probe (primary).** A fixed, committed denylist
     of paths that MUST remain ignored, asserted at head-ref with
     `git check-ignore --no-index`, **which does not require the path to
     exist**.

     **Revision 5 — the denylist contains ONLY paths guaranteed ignored by
     `HEAD` + `011.001-T`, so it can never depend on droppable content:**

     | Entry | Guaranteed by |
     |---|---|
     | `.env` | HEAD (`.env`) |
     | `.env.local` | HEAD (`.env.*`) |
     | `config.toml` | HEAD (`config.toml`) |
     | `.claude/instructions.md` | `011.001-T` |
     | `.github/copilot/settings.local.json` | `011.001-T` |
     | `.autoharness/gates/pipeline-topology-force-audit.log` | `011.001-T` |
     | `.backlogit/hooks_queue.jsonl` | `011.001-T` |

     Seven entries — non-vacuous, and every one survives a stowaway drop.

     **Revision 4's error, corrected here:** it added `.mcp.json`,
     `.copilot/x.json`, `.engram/x`, `backlogit.db`,
     `.backlogit/backlogit.db`, `.autoharness/staging/x` and
     `.vscode/settings.json`, described as "verified against the head-ref".
     They were in fact verified against the **uncommitted working tree** —
     measured directly, all seven are **NOT ignored at `HEAD`** and are
     ignored only by lines inside the **droppable** 52-line stowaway hunk
     (`HEAD`'s `.gitignore` ends at `config.toml` and has `.vscode/`
     commented out). That made a security gate depend on content the
     stowaway protocol may legitimately drop — red on arrival at
     `011.012-T`'s own landing commit under the E→F ordering, and
     permanently red on `main` if `011.015-T` is dropped. Same
     red-on-arrival class as revision 3, inverted.
     **Rule going forward: the denylist may never reference a pattern
     introduced by droppable content.**

     Still excluded, with reasons: `*.pem` and `id_*` match **no** pattern in
     this repo; `.backlogit/*.db` as a glob is not matched; and
     `plugin/.mcp.json` is **deliberately un-ignored** by an
     operator-authorized negation `011.015-T` exists to keep.
  2. **Existing-file differential (secondary).** Any path actually ignored at
     base-ref must still be ignored at head-ref. Base-ref ignore status is
     evaluated in a **detached worktree** (`git worktree add --detach`), since
     `git check-ignore` reads the working tree, not a ref.
* **Landing precondition (mandatory):** `--self-test` asserts **every denylist
  entry is ignored at HEAD**. This is the guard that prevents a mis-specified
  denylist from ever shipping again — it converts the revision-3 defect class
  into a mechanically detected one.
* **Per-path evaluation, literally quoted.** `git check-ignore` exits 0 if
  **any** argument is ignored, so a single batched invocation would mask an
  entry that silently became un-ignored. Evaluate each path individually (or
  `--stdin --verbose --non-matching` with per-line parsing), and quote
  glob-shaped entries so the runner's local files cannot expand them.
* **Non-vacuity assertion:** the checker reports the two counts **separately**
  (denylist entries evaluated; differential paths evaluated) and fails if the
  **denylist** count is zero. Revision 2 specified only
  `git ls-files --others --ignored --exclude-standard` at base-ref, which
  returns **empty on a fresh CI checkout**, so the gate would have reported
  green on every real PR while its fixtures passed.
* `git ls-files -i -c --exclude-standard` is empty (no tracked file ignored).
* A fixture that un-ignores a **previously ignored, existing** file is
  **rejected**, naming the path and the negation line responsible.
* A fixture adding a negation for a path that **does not currently exist** and
  is **not on the denylist** is **accepted** — the stowaway's actual case.
* **No waiver mechanism is built** (dropped as YAGNI — HEAD has zero negations
  and no concrete case exists). A deliberate un-ignore that trips the gate is
  **recorded and escalated to the operator**, not self-approved through an
  unowned bypass in a security gate.
* The checker header records its known boundaries: it inspects only the root
  `.gitignore`, the CI job is `pull_request`-gated, and **part 2 is inert on a
  clean CI checkout** (part 1 carries the real coverage there).
* `--self-test` passes and stays wired into CI.

#### 011.013-T · Pin the installer toolchain, not just the installed closure

Resolves `AFE0E95C`. `.github/constraints/autoharness-lock.txt` pins autoharness
and its closure, but pip/setuptools/wheel and the Python patch version are
ambient on the runner.

Prior learning `docs/compound/2026-09-05-pip-index-proxy-staleness-vs-real-ci.md`
is load-bearing.

*Approach:* configuration. Width: CI config. Size `M`, complexity `medium`.

**Acceptance criteria**

* Installer versions hash-pinned from the **public PyPI JSON API**, never a
  local index/proxy (the recorded prior failure was a proxy reporting a stale
  version and silently downgrading CI).
* Install uses `--require-hashes` **with** `--only-binary=:all:` to foreclose
  the unhashed-sdist PEP 517 fallback.
* **Bootstrap ordering:** the pip/setuptools/wheel install itself runs with
  `--require-hashes --only-binary=:all: --no-deps` from its own lock file and
  executes **before** the autoharness install — otherwise the hardened installer
  is not the one performing the hardened install.
* **Python pin coupling recorded:** `autoharness-lock.txt` was resolved for
  `cp312`/`manylinux_2_17_x86_64`; the `setup-python` patch pin must stay within
  3.12 or every hash becomes unsatisfiable. The coupling is noted next to both pins.
* The trust root is stated explicitly: pinning pip using the runner's
  preinstalled pip is irreducibly trust-on-first-use.
* Verified via a `workflow_dispatch` or draft-PR run **on the implementation
  branch** before the pin is relied upon. **Revision 3:** revision 2 said
  "throwaway branch", which conflicts with the single-active-branch rule
  (P-016) — no parallel implementation branch is created.
* Existing autoharness install continues to succeed.

#### 011.014-T · Add a depguard boundary rule for the Copilot SDK

Resolves `309FBF5A`(ii).

**Revision 2 — target corrected. Four reviewers independently found revision 1's
rule denied the wrong package.** It denied only
`github.com/github/copilot-sdk/go/rpc`, but the credential-consuming,
shell-executing surface — `NewClient`, `Client.Start`, `CreateSession`,
`SessionConfig.OnPermissionRequest` — lives in the **root** package
`github.com/github/copilot-sdk/go`, imported by `client.go`, `fixture.go`,
`permission.go` and every probe test. Only `permission.go` imports `/rpc`, and
it sits inside the allowlisted directory — so the revision 1 rule matched
**zero files** and would have shipped as false assurance for the boundary C3's
exit contract depends on.

*Approach:* configuration + durable negative fixture. Width: lint config.
Size `S`, complexity `low`.

**Acceptance criteria**

* The rule denies the **root prefix** `github.com/github/copilot-sdk/go`
  (which subsumes `/rpc`), exempting `internal/copilotprobe` via depguard's
  **negated, absolute-path-anchored `files` form** — approximately
  `files: ["$all", "!**/internal/copilotprobe/*.go"]`. A bare
  `internal/copilotprobe/*.go` pattern will not match, since depguard evaluates
  `files` against absolute paths.
* **`depguard` is added to `linters.enable`** — `.golangci.yml` uses
  `default: standard`, which excludes it; a rule under `settings` alone
  silently does nothing and would still satisfy a naive "lint passes" AC.
* **Two durable committed fixtures**, both asserted in CI:
  (i) a file outside the allowlist importing the SDK root package makes
  golangci-lint **exit non-zero** (proves the deny direction);
  (ii) a file under `internal/copilotprobe2/` also **fails** (proves the
  allowlist is anchored and has not silently widened to a sibling).
  A one-shot manual demonstration is insufficient — it cannot detect config drift.
* `golangci-lint run ./...` passes on the current tree.
* The allowlist carries a comment naming C3 as the removal trigger, and states
  that **depguard owns Go package-import boundaries** while
  `check-retired-architecture.sh` owns retired-architecture/path-surface
  contamination.

---

### Sub-epic F — Harness bootstrap carryover (disclosed stowaways)

> Units `011.015-T`–`011.018-T` carry **stowaway** content under the
> operator-authorized disposition in `9F9A3CB2`. Ship must disclose them under
> their own PR heading, validate each independently, and confirm no
> preserve-only path is staged. If a stowaway fails validation: **drop it from
> the PR and return it to the stash — do not block the feature, do not discard
> the change.** Note `011.001-T` is deliberately **not** in this set: it is
> non-droppable shipment work.
>
> **Drop-protocol control (Constitution VII), applying to ALL of
> `011.015-T`–`011.018-T`:** before any drop-and-return, the working-tree
> content is **backed up, the backup verified**, and the drop **requires
> explicit operator approval** — it is a scope-reducing decision over
> uncommitted operator work, not an agent-local call. The drop event is
> broadcast/recorded, not left implicit in the diff.

#### 011.015-T · Land and validate the full `.gitignore` stowaway hunk

Partially resolves `9F9A3CB2` (include-list item 1). `HEAD`'s `.gitignore`
contains **zero** negations; the stowaway adds **52 lines**, six of them the
first negations ever introduced into this repo's ignore policy.

**Revision 5 — scope corrected.** Revisions 1–4 scoped this unit to the six
negation lines only, leaving the other ~46 stowaway lines
(`bin/`, `logs/`, `.vscode/`, `.copilot/`, `.mcp.json`, `.env` +
`!.env.example`, `backlogit.db`, `.backlogit/*`, `.autoharness/staging|backups/`,
`.*.lock`, `.engram/`, `.graphtor/`, `prompt.md`, `diff.txt`, …) with **no
owning unit**. `9F9A3CB2` item 1 would have been recorded as covered while
~46 operator-preserved lines were silently dropped — the same omission class
`011.019-T` exists to prevent. The negations are also **inert or meaningless**
unless their preceding ignore lines land in the same commit
(`!plugin/.mcp.json` needs `.mcp.json`; the four `.engram/*` negations need
`.engram/`).

**Depends on `011.012-T`** (the hardened checker is the validation instrument).

*Approach:* configuration + verification. Width: ignore policy. Size `S`,
complexity `low`.

**Acceptance criteria**

* The **complete 52-line stowaway hunk** lands as one commit, strictly
  append-only, **after** `011.001-T`'s four lines so that unit's non-droppable
  content is never entangled with droppable content.
* All six negations are evaluated against `011.012-T`'s differential and the
  verdict **recorded per line**.
* **Effect asserted, not assumed:** four negations (`!.engram/.version`,
  `!.engram/.workspace-id`, `!.engram/config.toml`, `!.engram/registry.yaml`)
  are **inert** under git semantics — the preceding `.engram/` excludes the
  directory, and git cannot re-include a file whose parent directory is
  excluded. Assert their actual state via `git check-ignore -v` and record it
  rather than shipping a silently-wrong ignore policy as "validated".
* **`!plugin/.mcp.json` credential check:** MCP configs conventionally embed
  tokens in `env` blocks. Record that if ever tracked it must reference
  credentials by `${VAR}` indirection only, and run `gitleaks` explicitly over
  the newly un-ignored paths.
* `scripts/check-gitignore-append-only.sh` passes, including the new
  un-ignore rule, and `011.012-T`'s denylist still passes — **guaranteed by
  construction**, since revision 5's denylist references no pattern this hunk
  introduces.
* Committed as a **separate commit** appended **last in file order**, so a drop
  is a clean `git revert` / tail truncation that trivially preserves I6 —
  never a manual line-level edit.
* **Before any stash return** (Constitution VII): the working-tree content is
  backed up and the backup verified, operator approval obtained, then the
  return is recorded.

#### 011.016-T · Restore `start.ps1` capability-gating and tooling integration

Partially resolves `9F9A3CB2`. Revision 1 bundled **eight** functional
restorations into one `M`-sized unit — not credibly 2 hours. Decomposed by
capability cluster.

Restores: the graphtor-docs sync block including the
`.graphtor/bin/graphtor-docs.exe` fallback (its removal directly contradicts
the same change set's `.gitignore` addition of `.graphtor/`); the
`enabledSidecars` capability gating (so sidecars run on installed capability
packs, not mere command presence); and `ai_tools.copilot_cli.exe_path` forwarding.

*Approach:* test-first via the script integration harness. Width: PowerShell.
Size `M`, complexity `medium`.

**Acceptance criteria**

* All three capabilities restored and each **behaviourally asserted**, not
  merely parsed — a syntax check proves none of them.
* `enabledSidecars` gating is proven to suppress a sidecar whose command exists
  but whose capability pack is absent.
* Script parses cleanly and launches without error.

#### 011.017-T · Restore `start.ps1` secret and exit-code handling (no-echo enforced)

Partially resolves `9F9A3CB2`. Restores `GITHUB_PERSONAL_ACCESS_TOKEN`
resolution, the trailing exit propagating the Copilot exit code, and widens the
`.env.local` key regex from `[A-Z_][A-Z0-9_]*` back to `[A-Za-z_][A-Za-z0-9_]*`.

**Revision 2 — credential controls added.** This unit broadens what secret
material is loaded into the process environment **in the same cycle in which
`EF9352FB` documents a live plaintext Tavily key in `.env.local` as an open,
time-sensitive exposure.**

*Approach:* test-first. Width: PowerShell. Size `M`, complexity `medium`.

**Acceptance criteria**

* **No loaded value is ever echoed, logged, or written to a transcript** —
  asserted, covering the restored `GITHUB_PERSONAL_ACCESS_TOKEN` path.
* The regex widening is **justified against a concrete key set**; if no
  lowercase/mixed-case key is actually required, the narrower regex is retained.
* Exit-code propagation is asserted concretely (a non-zero Copilot exit
  produces a non-zero launcher exit).
* Verification runs against a **fixture script root** or with `.env.local`
  absent; the step must not capture or echo the process environment.
* `.env.local` is verified ignored and unstaged. **No secret value is read,
  printed, or committed.**

#### 011.018-T · Preserve the two improvements and add a divergence banner

Partially resolves `9F9A3CB2`.

**Revision 2 — rot control added.** Restoring the `Generated by autoharness`
marker while keeping local-only improvements is precisely the state that gets
clobbered on the next install. Revision 1's only mitigation was a paragraph in
a plan document that will be archived and never read at regeneration time.

*Approach:* test-first. Width: PowerShell. Size `S`, complexity `low`.

**Acceptance criteria**

* `Invoke-EngramCommandWithProgress` retains its bounded shared wall-clock
  budget (15000 ms default, `ENGRAM_PREWARM_TIMEOUT_MS` clamped to 1..30000),
  timeout kill, bounded cleanup wait, and the do-not-kill-a-reusable-daemon note.
* `Push-Location`/`Pop-Location` remain in `try`/`finally`.
* The `Generated by autoharness` marker is restored **adjacent to an in-file
  `LOCAL DIVERGENCE — re-apply after regeneration` block** enumerating the two
  improvements, so regeneration cannot silently erase them.
* A tracked in-repo follow-up records the upstream push (see boundary note).
* All eight revision-1 items are accounted for across `011.016-T`–`011.018-T`
  as restored, consciously kept, or explicitly deferred.

#### 011.019-T · Account for and disclose the full stowaway include list

Partially resolves `9F9A3CB2`.

**Revision 2 — closes a P1 scope omission.** `9F9A3CB2`'s operator-authorized
include list names **eight** paths; revision 1 covered only items 1–2, so
archiving the entry would have silently dropped six operator-preserved paths.

Items 3–8: `docs/bugs/2026-09-06-closure-evidence-producer-consumer-contract-mismatch.md`
(retained solely as a durable cross-reference; upstream-owned, non-actionable
locally, **no fix obligation attaches**), three `docs/memory/*.md` session
records (an established tracked convention — 11 already tracked), plus
`.backlogit/stash.jsonl` and `.backlogit/archive/stash.jsonl`
(**backlog-state carryover, explicitly NOT feature content**).

*Approach:* documentation. Width: docs. Size `S`, complexity `low`.

**Acceptance criteria**

* All eight include-list paths are enumerated with their disposition.
* **Items 3–6 are actually committed** (the `docs/bugs/` cross-reference and
  the three `docs/memory/` session records) — `9F9A3CB2` defines the include
  list as "exact paths intended for eventual commit with the next feature
  round", so enumeration alone would again close the entry on paper.
  `docs/bugs/...` ships as documentation only: **no fix obligation attaches**.
* The two `.backlogit/*.jsonl` paths are labelled **backlog-state carryover,
  not feature content**, exactly as the entry requires, and a
  `backlogit sync` reindex is run after they are committed so the index is not
  left stale.
* The four preserve-only paths are confirmed **not staged**
  (`git status` before commit — J6).
* `9F9A3CB2` is recorded as **partially resolved** and remains active until
  Ship's PR disclosure completes.

#### Boundary note (not harvested)

Pushing the two `start.ps1` improvements upstream into the autoharness template
belongs to `softwaresalt/autoharness`, a **different repository**, outside this
repo's boundary.

---

### Sub-epic G — Governance documentation

#### 011.020-T · Retire the `acp` scope, record C3 criteria and tracked divergences

Jointly resolves `6E701953` and `309FBF5A`(i).

*Approach:* documentation. Width: docs. Size `S`, complexity `low`.

**Acceptance criteria**

* `acp` removed from the Scopes enum at
  `.github/instructions/commit-message.instructions.md:31`; other scopes unchanged.
* The C3 phase entry names **copilotprobe deletion** and **depguard allowlist
  tightening** as explicit criteria.
* The **live-SDK opt-in convention** (`INTERCOM_LIVE_SDK_TESTS`) is recorded in
  a durable, package-independent location, so the policy survives
  `internal/copilotprobe`'s deletion at C3.
* Tracked divergences recorded (from `011.008-T`): `hasWindowsDrivePrefix` is
  **deliberately not GOOS-gated** because config text is portable and gating it
  would loosen a verdict on Linux; plus the `os.Stat` argument-order and
  `lexical.go` organization notes.
* No code or CI configuration touched.

---

# Plan Hardening

**Hardening required: YES.** Hardened 2026-09-06 (revision 4).

## Context consulted

* `docs/compound/build-errors/go-race-requires-cgo-fails-closed-2026-09-03.md`
  — load-bearing for `011.009-T`; `windows-latest` has no cgo, `-race` fails
  closed. *(Revision 1 cited a wrong path.)*
* `docs/compound/2026-09-05-pip-index-proxy-staleness-vs-real-ci.md` —
  load-bearing for `011.013-T`.
* `docs/compound/best-practices/external-spec-yields-to-workspace-instructions-2026-09-03.md`
  — 40-hex-SHA pinning is a MUST. *(Revision 1 cited a wrong path.)*
* `docs/compound/2026-09-04-backlogit-sizing-is-wit-gated.md` — Stage-side
  harvest hygiene.
* Source read directly: `internal/copilotprobe/testhelpers_test.go`,
  `internal/pathsafe/{pathsafe,root,lexical}.go`, `internal/config/validate.go`,
  `scripts/check-retired-architecture.sh`, `scripts/lib/OutputPathGuard.ps1`,
  `.github/workflows/ci.yml`, `.golangci.yml`, `.gitignore` (HEAD + working copy).
* **Coverage gap recorded:** the compound library has **zero** entries for
  pathsafe/containment design, the I6 append-only invariant,
  depguard/import-boundary rules, or live-SDK test gating. Sub-epics A, B, and E
  are **first-instance work with no institutional memory** — and indeed the
  review gate found its heaviest defects exactly there.

## Protected invariants

| # | Invariant | Why load-bearing | Enforcing gate |
|---|---|---|---|
| **J1** | No destructive filesystem write path is added to `internal/**` | D-B rests on the measured absence of write paths | **`011.004-T` mechanical CI gate** (revision 1 named a gate that did not exist) |
| **J2** | **No `internal/pathsafe` verdict becomes MORE PERMISSIVE.** Tightening is permitted only with a paired negative test | It is the Constitution IV control | Zero-verdict-change ACs in `011.005-T`–`011.008-T`; J12 for any tightening |
| **J3** | `internal/config` rule-**ordering** determinism (inherited I2) survives | Shipped behavioural guarantee | `011.003-T` changes no code |
| **J11** | **No `internal/config` `Validate()` verdict becomes more permissive on any GOOS** | Item (f) would have loosened `cli_path` acceptance on Linux | Item (f) dropped; `011.020-T` records the divergence |
| **J4** | `toolchain go1.26.5` preserved byte-identical | Removing it re-opens GO-2025-3750 | Diff review; `govulncheck` |
| **J5** | `.gitignore` changes remain strictly append-only | I6 is CI-enforced | `011.001-T`, `011.012-T`, `011.015-T` ACs |
| **J6** | No preserve-only path is ever staged | Four machine-local runtime paths | `011.001-T` lands first; `011.019-T` pre-commit check |
| **J7** | Every new CI action pinned to a full 40-hex SHA | `ci-security.instructions.md` MUST | `011.009-T` AC |
| **J8** | A new required job reaches `ci-gate`'s `needs:` **and** `results` | `needs:`-only leaves the gate decorative | `011.009-T` AC |
| **J9** | `011.012-T` and `011.015-T` are present in the same head state, with all six negation verdicts recorded | The hardened rule is the stowaway's validation instrument | `011.015-T` AC. *(Revision 1 framed this as commit ordering; the CI checker compares base→head, so intra-PR order is irrelevant — head state is what matters.)* |
| **J10** | No unit deletes `internal/copilotprobe` | C3-owned; deleting discards `011.002-T` | Diff review |
| **J12** | **No containment comparison is relaxed without a paired negative test** | The sibling-prefix bypass class | `011.011-T` mandatory `<root>-evil` test |

## Risk register

| Risk | Likelihood | Impact | Containment |
|---|---|---|---|
| Windows job red on first run | High | **Low** (was Medium) | Advisory via `WINDOWS_GATE_REQUIRED`; cannot block the other 19 units |
| `011.011-T` fix introduces a sibling-prefix bypass | Medium | **High** | J12 mandatory negative test at a nested root **and a top-level directory root** (not a drive root — `C:\-evil` is a legitimate descendant of `C:\`) |
| Un-ignore rule false-positives the stowaway | **Low** | Low | Behavioural differential ignores non-existent targets; denylist built from **measured** ignore status and excludes the deliberately un-ignored `plugin/.mcp.json`; trips are escalated to the operator (no waiver bypass) |
| pip hash-pin resolves against a stale proxy | Medium | High | Public PyPI JSON API; `--only-binary=:all:`; `workflow_dispatch`/draft-PR verification on the implementation branch (no second branch — P-016) |
| depguard ships as false assurance | **Closed** | — | Root-prefix deny + `linters.enable` + durable failing fixture |
| `start.ps1` restoration clobbered by regeneration | High | Medium | In-file `LOCAL DIVERGENCE` banner + tracked follow-up |
| Shipment breadth causes partial completion | Medium | Medium | Per-sub-epic exit contracts (A–G independently attestable); A–D land first |

## Feature exit contract

Attested **per sub-epic**, not as one feature-level "done", so partial
completion degrades into "A–D landed, E–G deferred" rather than an ambiguous
half-closed feature.

## Rollback

Every unit is independently revertible. The two highest-risk units
(`011.009-T`, `011.016-T`–`011.018-T`) revert without disturbing others.
`011.015-T` is a tail-truncation revert by construction. `011.010-T`'s gate
flip is reversible by unsetting a repository variable with no code change.

## Anti-goals (explicit)

* No TOCTOU/hardlink mitigation mechanism (D-B).
* No deletion of `internal/copilotprobe` (C3-owned).
* No broadening of the U-E1a gate to `internal/**` — not requested by any
  in-scope entry.
* No C3 adapter implementation.
* No outright ban on `.gitignore` negations.
* No external credential rotation (`EF9352FB` — operator-only).
* **No GOOS-gating of `hasWindowsDrivePrefix`** — it would loosen a verdict.
* No reimplementation of a gitignore matcher in bash — use git's own matcher.

## Review Remediation Ledger (revision 1 → 2)

| Finding | Severity | Remediation |
|---|---|---|
| No Constitution Check section | **P0** | Added, mapping I–XI with one justified violation |
| `011.002-T` Principle IV / branch (b) may pre-authorize a NON-NEGOTIABLE breach | **P0** | Branch (c): bounded, expiring exception; Principle IV shown not breached (no write path) |
| Branch (a) infeasible vs `pathsafe` API | P1 | Rejected with reason and recorded |
| Refactor unit had no test-first approach; contained containment-semantic changes | **P0** | Decomposed into `011.006-T`–`011.008-T`, all test-first; item (f) dropped |
| `011.008-T` now **refuses** | P0/P1 | **SUPERSEDED in revision 3** — refusal reverted as unauthorized over-reach; see the rev 3→4 ledger |
| depguard denied the wrong package; not in `linters.enable`; no durable proof | P1 ×4 | Root-prefix deny, `enable` required, committed failing fixture |
| Un-ignore rule banned every functional negation | P1 ×2 | Replaced with old-vs-new behavioural differential using git's matcher |
| `011.011-T` could introduce sibling-prefix bypass | P1 ×2 | Mandatory `<root>-evil` negative test (J12); UNC excluded with reason |
| Windows job promoted to required in its creating unit | P1 ×2 | Split; advisory via `WINDOWS_GATE_REQUIRED`, flipped in `011.010-T` |
| `ci.yml` generated-file drift unflagged | P1 | Escape hatch or divergence marker required |
| J1 named a gate its ACs did not provide | P1 | `011.004-T` is now a mechanical CI gate with an expanded write set |
| `9F9A3CB2`: 6 of 8 include-list paths unaddressed | P1 | `011.019-T` added; entry moved to `partially_resolves_stash` |
| `011.012.001-ST` bundled 8 restorations (2-hour breach) | P1 | Split into `011.016-T`–`011.018-T` |
| `.env.local` regex widened while a live key is exposed | P1 | No-echo ACs; widening must be justified or reverted |
| No safety mode declared (VIII) | P1 | `freeze-scope` + `careful` declared |
| Env-var gate could fail open; shell blast radius unbounded (VII/IV) | P1 ×2 | Default-deny semantics, `TestMain` backstop, `t.TempDir()` confinement, CI asserts unset |
| Stowaway stash return is destructive-adjacent | P1 | Verified backup required before return |
| `011.004-T` width violation (code+docs) | P1 | Now docs-only; rename branch removed |
| Unreachable branches required to "fail before, pass after" | P1 | `011.006-T` allows documented-unreachable outcome |
| Test pinned branches a sibling subtask collapses | P2 | Refactor (`011.007-T`) decided before tests |
| Droppable stowaway commingled with non-droppable work | P2 ×3 | `011.001-T` (non-droppable, first) split from `011.015-T` (droppable, last) |
| Preserve-only paths exposed for the whole shipment | P2 | `011.001-T` lands **first** |
| Over-fragmentation; verification-only subtasks | P2 ×4 | Flattened to 20 single-purpose tasks, no subtasks |
| Missing `persist-credentials: false` on Windows checkout | P3 | Added to AC |
| pip bootstrap ordering / cp312 coupling / TOFU unstated | P3 | All three added to `011.013-T` |
| Wrong `git ls-files` verification command | P2 | Corrected to `git ls-files -i -c --exclude-standard` |
| Four `.engram/*` negations inert but claimed effective | P2 | `011.015-T` asserts actual state |
| Two broken compound citation paths | P3 ×2 | Corrected |
| Boundary-enforcement ownership undeclared | P3 | Recorded in `011.014-T` |
| Live-test convention dies with copilotprobe | P3 | Recorded durably in `011.020-T` |

## Review Remediation Ledger (revision 2 → 3)

Attempt 2 cleared all four P0s. Two P1s remained, both remediated here.

| Finding | Severity | Remediation |
|---|---|---|
| **`011.012-T` differential vacuous in CI** — `git ls-files --others --ignored` is empty on a fresh checkout, so the gate could never fire on a real PR | **P1** (Correctness FAIL; Security P2 concurred) | Two-part check: canonical sensitive-path denylist via `git check-ignore --no-index` (works on non-existent paths) + existing-file differential, plus a **mandatory non-vacuity assertion** (fails if zero paths evaluated) |
| **`011.008-T` over-reach** — escalated a P3 doc-note ask into a production refusal in the containment API | **P1** (Scope; Constitution P1 and Correctness P2/P3 concurred) | Reverted to authorized scope: doc note + literal-treatment test + positive test that `$Recycle.Bin` still validates. Principle IV shown not engaged (a literal `$VAR` resolves *inside* the root); residual caller-side gap recorded explicitly in the Constitution Check |
| `011.004-T` bolted onto `check-retired-architecture.sh`, colliding with `4989A42D` condition (b) and the anti-goal | P2 | Moved to a new dedicated script; the U-E1a gate is left narrowed and measurable |
| `011.004-T` never created the consolidated register (`BF5DE670`'s primary ask) | P2 | Register creation is now an explicit AC; `BF5DE670` moved to `partially_resolves_stash` |
| `011.004-T` "indirect classes" pre-constrain C3; bare alternation false-positives on `CreateSession` | P2 ×2 | Indirect classes dropped; qualified selectors specified; scope extended to `cmd/**` (Security P2) |
| `011.011-T` drive-root sibling AC unsatisfiable (`C:\-evil` is a legitimate descendant) | P2 | Sibling test moved to nested + top-level dir root; drive/POSIX roots get the positive case only; `GuardPath`/`RepoRoot` decoupling required |
| `011.002-T` TestMain mechanism unspecified and self-conflicting; enumeration AC unsatisfiable against a live model | P1/P2 | TestMain sanitizes + sets a denial flag and still calls `m.Run()`; enumeration replaced by a **deny-by-default allowlisted** permission handler |
| `011.012-T` waiver mechanism was YAGNI and an unowned bypass | P2 ×2 | Waiver dropped; trips are recorded and escalated to the operator |
| `J2` contradicted `011.008-T` | P2 ×2 | Restated: no verdict becomes **more permissive**; tightening only with a paired negative test |
| Constitution II row claimed blanket red-phase compliance | P1 | Three deviations documented (characterization-first refactor; config-only units) |
| "Pester-style harness" misnomer — no Pester exists | P2 ×2 | Corrected to the Go integration harness that shells out to `pwsh` |
| VII drop-protocol control covered only `011.015-T` | P1 | Extended to all four droppable units, with operator approval + verified backup |
| `011.013-T` "throwaway branch" conflicts with P-016 | P2 | Changed to `workflow_dispatch`/draft-PR on the implementation branch |
| `011.010-T` "unreachable for any non-privilege reason" undefined; gate flip is a VII config change | P2 ×2 | Skip conditions enumerated as a closed set; flip requires explicit operator confirmation |
| `011.014-T` glob form would not match; allowlist widening undetectable | P3 ×2 | Negated absolute-anchored `files` form specified; second `copilotprobe2` fixture added |
| `011.006-T` testability-seam branch was production code in a tests-only unit | P2 | Seam alternative removed; documented-unreachable is the sole outcome |
| `011.019-T` had no commit obligation for include-list items 3–6 | P3 | Commit obligation added, plus a post-commit `backlogit sync` reindex |
| D-G cited a non-existent compound path | P3 | Corrected in the deliberation |

## Review Remediation Ledger (revision 3 → 4)

Attempt 3 cleared the `011.008-T` over-reach P1. One **new** P1 was introduced
by revision 3's own remediation and is fixed here, along with stale
cross-references.

| Finding | Severity | Remediation |
|---|---|---|
| **`011.012-T` denylist unsatisfiable and self-contradictory** — `*.pem`, `id_*`, `.backlogit/*.db` match **no** pattern in this repo, and `plugin/.mcp.json` is **deliberately un-ignored** by the stowaway negation `011.015-T` exists to keep. The gate would be **red on arrival** with no sanctioned exit under `freeze-scope`, forcing a choice between dropping operator-authorized content and editing the gate | **P1** (Correctness FAIL + Scope P1) | Denylist rebuilt from **measured** `git check-ignore` results; the four bogus entries removed with reasons recorded; the four paths `011.001-T` newly ignores added, so the gate protects exactly what lands first. **New landing-precondition AC:** `--self-test` asserts every denylist entry is ignored at HEAD, converting this defect class into a mechanically detected one |
| `git check-ignore` exits 0 if **any** argument matches; glob entries shell-expand | P2 | Per-path evaluation with literal quoting mandated |
| Non-vacuity counter was a tautology; part 2 still inert in CI | P3 | Two counts reported separately; part 2's CI inertness recorded in the checker header |
| Base-ref ignore status has no stated mechanism | P3 | Detached worktree (`git worktree add --detach`) specified |
| Constitution row IV still asserted `011.008-T` "refuses" | P2 ×2 | Row rewritten; the residual caller-side vector is now **named and recorded**, satisfying `011.008-T`'s own AC |
| Risk register re-asserted drive-root sibling test, the dropped waiver, and throwaway-branch verification | P2/P3 ×3 | All three rows corrected to match their units |
| `011.002-T` "client constructor" implied production code in a tests-only unit | P2 ×2 | Disambiguated to the `_test.go` helpers; explicit prohibition on touching `client.go` |
| `011.004-T` heading said "Resolves" `BF5DE670` | P3 ×2 | Changed to "Partially resolves" |
| Frontmatter `revision: 2` and "Hardened (revision 2)" stale | P3 ×2 | Corrected to revision 4 |
| Rev 1→2 ledger row for `011.008-T` not marked superseded | P2 | Marked SUPERSEDED |
| D-G citation still broken in the deliberation (rev 3 ledger claimed otherwise) | P3 | Actually corrected this time at the D-G line itself |

### Accepted residuals (tracked, not blocking)

## Review Remediation Ledger (revision 4 → 5)

Attempt 4 confirmed every revision-3 P2/P3 remediated, and found **two P1s
sharing one root cause** — both fixed here. I verified the root cause
empirically before accepting it (rebuilt `HEAD`'s `.gitignore` in a scratch
repo and re-ran `git check-ignore` per path), rather than taking the review
on faith.

| Finding | Severity | Remediation |
|---|---|---|
| **Denylist depended on droppable stowaway content.** Revision 4 claimed its entries were "verified against the head-ref"; they were verified against the **uncommitted working tree**. **Measured at `HEAD`:** `.env`, `.env.local`, `config.toml` are ignored; `.mcp.json`, `.copilot/x.json`, `.engram/x`, `backlogit.db`, `.backlogit/backlogit.db`, `.autoharness/staging/x`, `.vscode/settings.json` are **NOT** — all seven come from the droppable 52-line hunk. Gate red on arrival at `011.012-T`'s landing commit, and permanently red on `main` if `011.015-T` is dropped | **P1** | Denylist reduced to **7 entries guaranteed by `HEAD` + `011.001-T`** (both non-droppable). New standing rule recorded: *the denylist may never reference a pattern introduced by droppable content.* Non-vacuity preserved |
| **No unit owned the ~46 non-negation stowaway lines.** `011.001-T` covers 4 pure additions; `011.015-T` covered only the 6 negations. `9F9A3CB2` item 1 ("52 added lines") would have been recorded as covered while ~46 operator-preserved lines were silently dropped — the omission class `011.019-T` exists to prevent. The negations are also inert unless their preceding ignore lines land in the same commit | **P1** | `011.015-T` re-scoped to land the **complete 52-line hunk**, appended after `011.001-T`'s lines, keeping non-droppable and droppable content unentangled |

### Accepted residuals (tracked, not blocking)
* `(*os.File).Write*` is not textually decidable by a grep detector —
  `011.004-T` may drop the method selector and rely on package-qualified
  constructors.
* `011.011-T`'s top-level-directory-root fixture needs privileges on Linux CI;
  the nested fixture exercises the same code path and carries the hazard class.
* `011.016-T`/`011.017-T` still require new launcher-level fixtures; accepted
  as the cost of behaviourally validating restored capabilities.
* `011.004-T` bundles register + godoc + script; accepted as one unit because
  the register and the gate's failure message are mutually referential.

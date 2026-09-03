---
title: "Implementation Plan — intercom-go P1: Error Taxonomy and Workspace Path Containment"
description: "Port the AppError 14-variant taxonomy and the workspace path-containment security control from the agent-intercom Rust oracle into intercom-go"
date: 2026-09-03
status: reviewed
phase: P1
source_deliberation: "docs/decisions/2026-09-03-intercom-go-port-continuation-deliberation.md"
source_brief: "docs/decisions/2026-07-06-go-port-reference-brief.md"
behavioral_oracle:
  repo: "softwaresalt/agent-intercom"
  commit: "41df772"
  access: "read-only, out-of-tree"
stash_entries: ["4989A42D", "037B1552"]
merge_base: "87b800d5d3a6457d86fecd12fb7a299f170af718"
prior_closure: "docs/closure/2026-09-03-intercom-go-foundation-closure.md"
tags: ["go", "port", "errors", "path-safety", "security"]
---

# Implementation Plan — intercom-go P1: Error Taxonomy and Workspace Path Containment

## Problem Frame

Shipment `001-S` delivered the CGO-free Go module skeleton, two stub entrypoints, a
4-target cross-compile build script, and a hardened CI pipeline. It deliberately deferred
the `AppError` taxonomy, the config package, and credential resolution to stash
`037B1552`, because those are operator-facing contracts that plan review required be
validated against the Rust behavioral oracle before being frozen.

That validation has now been performed against the correct oracle
(`softwaresalt/agent-intercom` @ `41df772`) — see the source deliberation, findings F1–F6.
The deferral condition is satisfied.

This plan implements **phase P1** of the roadmap: the two Tier-0 leaf packages that every
later phase depends on.

* `internal/apperr` — the 14-variant `AppError` taxonomy with its exact display contract.
* `internal/pathsafe` — the workspace path-containment control that produces `PathViolation`.

These are the dependency root of the port. `AppError` has ~470 construction sites in the
oracle's `src/`; every config validation rule in phase P2 is expressed as an
`AppError::Config`; and `pathsafe` is required by P2 (workspace-root canonicalization)
and P10 (diff apply).

## Scope Boundaries

**In scope:** the `internal/apperr` package (14 error kinds, exact `Error()` display
strings, kind-based `errors.Is`, cause-preserving `errors.Unwrap`, typed `errors.As`);
the `internal/pathsafe` package (lexical normalization, workspace containment, symlink
escape detection); the 16 oracle tests re-expressed as Go tests; and a
`docs/oracle-pin.md` provenance record.

**Out of scope (deferred to P2, stash successor `037B1552`):** the `config.toml` schema,
non-zero default population, the nine config validation rules, workspace-to-channel
routing, and config hot-reload.

**Out of scope (deferred to P3, stash successor):** credential resolution, the keychain
provider, the `*_ACP` precedence chain, and the redacting `Secret` type. **No
credential-handling code ships in this slice.**

**Out of scope (deferred, per the deliberation roadmap):** SQLite and the 7 tables (P4/P5),
domain models (P5), ACP subprocess (P6), `AgentDriver` (P7), Slack (P8/P9), diff apply
and approval workflow (P10), the retained MCP HTTP tool endpoint (P11), resilience (P12),
IPC and `intercom-ctl` (P13), and the parity sweep (P14).

**Out of scope (explicitly excluded by the operator):** `A92E3FA0` (multiplexer backend),
the six foundation hardening follow-ups (`EFFAA358`, `90350C9A`, `F7C6420D`, `BEDD2E70`,
`35D76D5E`, `90EE7758`), and `EF9352FB` (credential rotation — an operator action).

**Not wired into `cmd/`.** Both packages are net-new leaves under `internal/` with no
importers in this slice. This is the rollback boundary: either package reverts by deleting
its directory, with zero downstream breakage.

## Requirements Trace

| Req | Source | Units |
|---|---|---|
| R1 | Brief §10 — `AppError` taxonomy, 14 variants, lowercase prefixed messages | A1 |
| R2 | Oracle `src/errors.rs` — exact display format, no trailing period, variant distinctness | A1 |
| R3 | Stash `037B1552` **P1-j** — `Is()` serves the error kind; `Unwrap()` is reserved for the wrapped cause | A2 |
| R4 | Brief §10 — workspace isolation is NON-NEGOTIABLE; reject `..`, symlink, and absolute escapes | B1, B2 |
| R5 | Oracle `src/diff/path_safety.rs` — the 7-step containment algorithm, ported faithfully | B1, B2 |
| R6 | Stash `037B1552` **P1-e** — canonicalize + `EvalSymlinks` roots, emit `PathViolation`, absolute-path invariant | B1, B2 |
| R7 | Oracle `tests/unit/error_tests.rs` (7) + `tests/unit/path_validation_tests.rs` (9) — parity corpus | A1, A2, B1, B2 |
| R8 | Deliberation D2 — oracle provenance must be pinned in-repo | C1 |
| R9 | Workspace standard — CGO-free build, `-race` tests, lint clean | all |

## Constitution Check

| Principle | Status | Units | Justification / rejected alternative |
|---|---|---|---|
| I — Quality gates (`gofmt`, `golangci-lint`, `go vet`, tests) | Satisfied | all | Existing `001-S` CI gates apply unchanged; no workflow edits in this slice. |
| II — Test-first (NON-NEGOTIABLE) | Satisfied | A1, A2, B1, B2 | Every code unit carries its `_test.go` in the same unit with tests written first. Ported oracle tests are the specification. |
| III — Workspace isolation and security boundaries (MUST) | Satisfied | B1, B2 | This slice *implements* the control. Symlink escape is covered by a dedicated test. |
| IV — Least surprise / no out-of-tree writes (NON-NEGOTIABLE) | Satisfied | all | Both packages are pure; `pathsafe` performs read-only FS stats and never writes. |
| V — Structural safety over convention | Satisfied | A2, B2 | `Secret`-by-type is deferred to P3; here, safety is structural via a single containment entry point that all callers must route through. |
| VI — Observability | Satisfied | A1 | Error kinds are enumerable and stable, enabling structured logging by kind in later phases. |
| VII — Destructive actions require approval (NON-NEGOTIABLE) | **Deviation, documented** | — | `.gitignore` carries a pre-existing unrelated local modification and `.claude/` is untracked. Both MUST be preserved byte-exact and never staged. Registered below. |
| VIII — Dependency minimalism | Satisfied | all | **Zero new module dependencies.** Standard library only (`errors`, `fmt`, `path/filepath`, `os`, `strings`, `testing`). |
| IX — Reproducible builds | Satisfied | all | No `go.mod` change, so `go.sum` is untouched and `-mod=readonly` holds. |
| X — Documented decisions | Satisfied | C1 | Deliberation linked in frontmatter; `docs/oracle-pin.md` records the oracle revision. |
| XI — Closure and verification | Satisfied | — | Runtime Verification and Closure section below. |

**Documented violations register:** one — Principle VII deviation on pre-existing dirty
state (`.gitignore`, `.claude/`, `.backlogit/hooks_queue.jsonl`, `references/herdr`).
Mitigation: these paths are forbidden to every unit; see Plan Hardening → Protected invariants.

## Implementation Units

Sizing follows the 2-hour rule: **< 3 files, < 5 functions, < 4 test scenarios** per unit,
single skill domain per unit. Unit boundaries were resized during plan review (findings
SCOPE-1…SCOPE-4) and implementation mechanisms were made explicit (findings GO-1…GO-7,
SEC-1) so no mechanism choice is left to the executor.

### Sub-epic A — `internal/apperr`: the error taxonomy

#### A1. Define the 14 error kinds and the exact display contract

* **Files:** `internal/apperr/apperr.go`, `internal/apperr/apperr_test.go`
* **Functions:** `Kind.prefix()`, `(*Error).Error()` — 2.
* **Approach:** test-first. Define `type Kind int` with 14 constants and `prefix()`
  returning the oracle's exact lowercase prefixes: `config`, `db`, `slack`, `mcp`,
  `diff`, `policy`, `ipc`, `path violation`, `patch conflict`, `not found`,
  `unauthorized`, `already consumed`, `io`, `acp`. Define
  `type Error struct { kind Kind; msg string; cause error }` with **unexported fields**
  and accessors `Kind()`/`Message()`, so an error's category cannot be mutated after
  construction (finding GO-8; this is a security-relevant taxonomy). **All methods use
  pointer receivers; all constructors return `*Error`** (finding GO-9).
  `Error() string` returns exactly `"{prefix}: {msg}"` — the cause is **never** appended,
  because the oracle's display format is the pinned contract.
* **Acceptance criteria:**
  1. A table test over all 14 kinds asserts each rendered prefix is unique and no
     rendered message ends in `.`.
  2. `(&Error{kind: KindACP, msg: "stream closed"}).Error() == "acp: stream closed"`.
  3. `go test -race ./internal/apperr/...` passes; `gofmt -l` empty; `go vet` clean.
* **Posture:** test-first. **Size:** S. **Complexity:** low.

#### A2. Kind sentinels and `errors.Is` support

* **Files:** `internal/apperr/sentinel.go`, `internal/apperr/sentinel_test.go`
* **Functions:** `kindSentinel.Error()`, `(*Error).Is()` — 2.
* **Approach:** define an **unexported** `type kindSentinel Kind` implementing `error`,
  and export 14 sentinels as `var ErrConfig error = kindSentinel(KindConfig)` … `ErrACP`.
  Sentinels are immutable value types, which removes the mutable-global hazard of
  `*Error` sentinels (finding GO-1 — **P0**; the concrete sentinel type is now specified
  and is not left to the executor). Implement
  `func (e *Error) Is(target error) bool { k, ok := target.(kindSentinel); return ok && e.kind == Kind(k) }`.
  `Is` performs a direct type assertion and never calls `errors.Is` internally, so no
  recursion is possible.
* **Acceptance criteria:**
  1. `errors.Is(&Error{kind: KindNotFound, msg: "x"}, ErrNotFound)` is true and
     `errors.Is(&Error{kind: KindNotFound, msg: "x"}, ErrDB)` is false. **Note:** A2's
     tests construct `*Error` with a package-internal struct literal because the
     constructors do not exist until A3 (finding CORR-1); this is legal since the test
     file is in the same package.
  2. A table test asserts each of the 14 sentinels matches its own kind and no other.
  3. `godoc` on `Is` documents that matching is by **kind**, and that because `Unwrap`
     is also implemented, `errors.Is` matches a kind appearing **anywhere in the cause
     chain**, not only at the top level (finding GO-2).
* **Posture:** test-first. **Size:** S. **Complexity:** medium.

#### A3. Constructors and cause-preserving `Unwrap`

* **Files:** `internal/apperr/apperr.go` (extend), `internal/apperr/wrap_test.go`
* **Functions:** `New()`, `Newf()`, `Wrap()`, `(*Error).Unwrap()` — 4.
* **Approach:** `New(kind, msg)`, `Newf(kind, format, args...)`, and `Wrap(kind, cause)`
  where `Wrap` sets `msg = cause.Error()`, reproducing the oracle's three `From` impls
  (`toml::de::Error`, `sqlx::Error`, `std::io::Error`) which stringify the source.
  **`Wrap` must guard `cause == nil`** and fall back to an empty message rather than
  panicking on a nil-interface method call (finding GO-3). `Unwrap()` returns **only**
  the cause — `Is` (kind) and `Unwrap` (cause) are deliberately separate mechanisms
  because a single `Unwrap` cannot serve both roles (stash **P1-j**). Unlike the oracle,
  the cause is retained so `errors.Is`/`errors.As` can traverse it; this is a recorded
  improvement that does not alter the display contract. Document in godoc that `Newf`
  format strings are **not** checked by `go vet`'s printf analyzer, since custom `…f`
  functions outside `fmt` are not recognized (finding GO-11).
* **Acceptance criteria:**
  1. `Wrap(KindIO, os.ErrNotExist)`: `errors.Is(err, os.ErrNotExist)` is true **and**
     `err.Error() == "io: " + os.ErrNotExist.Error()` — the cause is reachable but
     invisible in the display.
  2. `Wrap(KindIO, nil)` does not panic and renders `"io: "`.
  3. `errors.As` recovers a `*Error` and its `Kind()` both when the `*Error` is outermost
     and when it is wrapped by `fmt.Errorf("…: %w", …)` (finding GO-10).
* **Posture:** test-first. **Size:** S. **Complexity:** medium.

### Sub-epic B — `internal/pathsafe`: workspace containment

#### B1. The validated `Root` type

* **Files:** `internal/pathsafe/root.go`, `internal/pathsafe/root_test.go`
* **Functions:** `NewRoot()`, `stripUNCPrefix()`, `(Root).Path()` — 3.
* **Approach:** `type Root struct { path string }` with
  `func NewRoot(dir string) (Root, error)`, which canonicalizes **once**:
  `filepath.Abs` → `filepath.EvalSymlinks` → `stripUNCPrefix`. This amortizes what would
  otherwise be an `EvalSymlinks` syscall walk on **every** `Resolve` call — P10's
  diff-apply may validate 10–100 paths against one root (finding GO-4). Failure returns
  `apperr.New(apperr.KindPathViolation, "workspace root invalid: …")`.
  `stripUNCPrefix` removes a leading `\\?\`, which Go's `EvalSymlinks` emits on Windows
  via `GetFinalPathNameByHandle`. **Provenance note:** this mirrors
  `src/config.rs:24` `strip_unc_prefix`, which the oracle applies to
  `default_workspace_root` at `src/config.rs:546` — it is **not** called from
  `src/diff/path_safety.rs`, which relies on Rust's component-aware `Path::starts_with`
  instead. The stripping is therefore a **Go-specific adaptation** required because Go's
  containment comparison is string-based, not an oracle-faithful port of `path_safety.rs`
  (findings SEC-4 / ARCH-4 — the earlier draft misattributed it).
* **Acceptance criteria:**
  1. `NewRoot` on an existing directory returns an absolute, symlink-resolved path with
     no `\\?\` prefix.
  2. `NewRoot` on a non-existent directory returns a `PathViolation` whose message begins
     `"path violation: workspace root invalid"`.
  3. `NewRoot` is idempotent: `NewRoot(r.Path())` yields the same path.
* **Posture:** test-first. **Size:** S. **Complexity:** medium.

#### B2. Lexical rejection and component normalization

* **Files:** `internal/pathsafe/pathsafe.go`, `internal/pathsafe/lexical_test.go`
* **Functions:** `normalize()` — 1 (plus the `Resolve` entry point stubbed for B3).
* **Approach:** implement oracle steps 3's rejection arms. Go has no
  `Path::components()` typed enum, so the mechanism is specified explicitly
  (finding GO-5):
  1. **Before** any cleaning, reject if `filepath.VolumeName(candidate) != ""`. This
     catches Windows **drive-relative** forms such as `C:foo` that `filepath.IsAbs`
     reports as *false* and that would otherwise be accepted as a normal component —
     on NTFS `C:foo` is an alternate-data-stream reference. The oracle rejects these via
     `Component::Prefix` (finding SEC-1 — **P1**).
  2. Reject if `filepath.IsAbs(candidate)`.
     Both rejections emit `"absolute paths are not allowed; use workspace-relative paths"`.
  3. `filepath.Clean(candidate)`, then split on `filepath.Separator`.
  4. Walk components: `".."` pops the accumulated stack and errors
     `"path attempts to escape workspace"` if the stack is already empty; `"."` and `""`
     are skipped; all else is pushed.
  Note `filepath.Clean` already collapses interior `a/b/../c` → `a/c` and preserves only
  *leading* `..`. It also collapses redundant separators (`a//b` → `a/b`) and strips
  trailing ones (`a/b/` → `a/b`) — exactly the cases Rust's component iterator elides by
  yielding only `Normal` components. The post-`Clean` walk therefore yields the same final
  stack as the oracle's pre-`Clean` component walk, and for the same reason rather than by
  coincidence (finding CORR-3). Criterion 3 asserts this empirically. Verified behavior:
  `Clean("")` → `"."`; `Clean("../secret.txt")` → `"../secret.txt"` (leading `..`
  preserved, so rejection still fires); `Clean("src/../../secret.txt")` →
  `"../secret.txt"`; `Clean("a/b/../c.txt")` → `"a/c.txt"`.
* **Acceptance criteria:**
  1. Escapes rejected with `"path attempts to escape workspace"`: `../secret.txt`,
     `src/../../secret.txt`, `subdir/../../escape.txt`.
  2. Absolute and volume-relative rejected with the absolute-path message: `/etc/passwd`;
     and on Windows `C:\Windows\system32` and `C:foo` (finding SEC-2).
  3. Safe interior traversal is preserved: `a/b/../c.txt` normalizes to `a/c.txt`, and
     `./src/main.go` normalizes to `src/main.go` (finding GO-13, proving `Clean`
     equivalence).
* **Posture:** test-first. **Size:** M. **Complexity:** medium.

#### B3. Containment enforcement and `Resolve`

* **Files:** `internal/pathsafe/pathsafe.go` (extend), `internal/pathsafe/contain_test.go`
* **Functions:** `(Root).Resolve()`, `hasPathPrefix()` — 2.
* **Approach:** join the normalized components onto `Root.path`, then assert containment
  using an explicitly **component-aware** comparison (finding GO-6):

  ```go
  func hasPathPrefix(path, prefix string) bool {
      if pathEqual(path, prefix) { return true }
      return pathHasPrefix(path, prefix+string(filepath.Separator))
  }
  ```

  where `pathEqual`/`pathHasPrefix` fold case **only** when
  `runtime.GOOS == "windows"` (finding GO-7 — a case mismatch between a config-supplied
  root and an `EvalSymlinks`-canonicalized candidate would otherwise cause a false
  rejection). A raw `strings.HasPrefix(path, root)` is **forbidden** — it admits a
  sibling directory named `<root>-evil`. Violations emit `"path outside workspace"`.
  **Empty-candidate guard (required):** if `normalize()` yields **zero** components,
  `Resolve` MUST return `"path outside workspace"` rather than falling through. Without
  this guard the algorithm silently resolves to the workspace root itself, because
  `filepath.Clean("")` returns `"."`, the walk skips `"."`, and
  `filepath.Join(root, "")` returns `root` — empirically confirmed. The oracle accepts
  this input; rejecting it is a deliberate, recorded Go-side tightening that prevents a
  caller from being handed the root directory as a writable target (finding CORR-2).
  Godoc must state the postcondition: *the returned path is absolute, cleaned, and
  strictly contained within the root* (finding GO-15).
  **Deliberate omission:** the oracle's fast path (`src/diff/path_safety.rs` — an already
  absolute candidate that `starts_with` root skips the walk) is **not** ported, because
  B2 rejects all absolute candidates. P10 must therefore pass workspace-relative paths
  through the diff-apply pipeline, or revisit this decision then (finding SEC-3).
* **Acceptance criteria:**
  1. `src/main.go` and `./src/main.go` resolve to an absolute path inside the root.
  2. A sibling directory `<root>-evil` is rejected with `"path outside workspace"`
     (the component-aware containment proof).
  3. An empty-string candidate is rejected with `"path outside workspace"` rather than
     silently resolving to the workspace root itself (finding SEC-2).
* **Posture:** test-first. **Size:** M. **Complexity:** medium.

#### B4. Symlink escape detection

* **Files:** `internal/pathsafe/pathsafe.go` (extend), `internal/pathsafe/symlink_test.go`
* **Functions:** `checkSymlinkEscape()` — 1.
* **Approach:** implement oracle steps 6–7. If the resolved path **exists**, run
  `filepath.EvalSymlinks` + `stripUNCPrefix` and re-assert containment via
  `hasPathPrefix`; on failure emit `"symlink target escapes workspace"`. Use `os.Stat`
  (not `os.Lstat`) for the existence probe, matching the oracle's `Path::exists()` which
  follows symlinks — a **broken** symlink is therefore treated as non-existent and takes
  the lexical-only branch (finding GO-14; documented, not silently inherited).
  Non-existent paths are accepted after lexical checks only (oracle step 7).
  The package doc comment must record two known limitations, both oracle-parity:
  the **TOCTOU** window between validation and use, and the fact that
  `EvalSymlinks` does **not** resolve **hardlinks**, so a pre-existing in-workspace
  hardlink to an external file on the same volume passes validation (finding SEC-5).
  Tests gate on a **runtime symlink-privilege probe**, not `runtime.GOOS`, so elevated
  Windows environments are still covered (finding GO-12); the probe attempts
  `os.Symlink` in `t.TempDir()` and calls `t.Skip` with an explicit reason on failure —
  it must never pass vacuously.
* **Acceptance criteria:**
  1. A symlink inside the root whose target is outside the root is rejected with
     `"symlink target escapes workspace"`.
  2. A symlink inside the root whose target is also inside the root resolves successfully.
  3. A non-existent relative path under the root resolves successfully (oracle step 7).
* **Posture:** test-first. **Size:** M. **Complexity:** high.

### Sub-epic C — Provenance

#### C1. Record the behavioral-oracle pin and the parity trace

* **Files:** `docs/oracle-pin.md`
* **Approach:** documentation only. Record the oracle repository
  (`softwaresalt/agent-intercom`), the pinned commit (`41df772`), the local read-only
  path, the verification method (remote URL + module map + `src/` LOC + the exact
  56/22/46/6 test-corpus counts), and the explicit rejection of `references/herdr`
  (`ogulcancelik/herdr` @ `b07ba9ce`) as a mis-mount. State that the oracle is
  out-of-tree, so the pin is documentary and must be re-verified before being trusted.
  Include the oracle-test → acceptance-criterion traceability table (finding SCOPE-5)
  reproduced from "Deepened verification" below.
* **Acceptance criteria:**
  1. `Select-String` finds all of these literal tokens in `docs/oracle-pin.md`:
     `41df772`, `softwaresalt/agent-intercom`, `ogulcancelik/herdr`, `b07ba9ce`
     (finding SCOPE-6 — deterministic rather than judgment-based).
  2. The traceability table maps all 16 oracle tests to a covering unit and criterion.
  3. `markdownlint` passes (pre-commit hook is installed in this workspace).
* **Posture:** docs-only. **Size:** XS. **Complexity:** trivial.

## Dependency Graph

```text
A1 (kinds + display contract)
 ├──> A2 (kindSentinel + Is)
 │     └──> A3 (New/Newf/Wrap + Unwrap)
 └──> B1 (Root: canonicalize + UNC strip)   [needs apperr.KindPathViolation]
        └──> B2 (lexical rejection + normalize)
               └──> B3 (containment + Resolve)
                      └──> B4 (symlink escape)
                             └──> C1 (oracle pin + parity trace)  [docs; last, records shipped state]
```

Strictly sequential. `A1` is the only unit with no predecessor; `A3` and `B1` both depend
on `A1` but `A3` is not on `B`'s critical path, so execution order is
`A1 → A2 → A3 → B1 → B2 → B3 → B4 → C1`. There is no parallel path and no unit may be
executed in a separate worktree (P-016).

## Decisions and Rationale

| Decision | Rationale | Rejected alternative |
|---|---|---|
| Retain all 14 kinds including `PatchConflict` | Brief §10 pins the variant list; the taxonomy is a contract even where a kind is currently unused | Drop `PatchConflict` (vestigial in the oracle — never constructed). Rejected: it desynchronises the taxonomy from the spec for no benefit; P10 will decide whether to make it load-bearing. |
| Retain the `Mcp` kind | Per deliberation D1 the MCP **HTTP tool endpoint** is retained as P11 and needs it | Drop `Mcp` along with MCP-as-a-mode. Rejected: conflates three distinct surfaces. |
| `Error()` excludes the cause | The oracle's display strings are pinned by tests and are an operator-visible contract | Append `": " + cause` (idiomatic Go). Rejected: breaks the pinned format. |
| Retain the cause for `Unwrap` anyway | Enables `errors.Is(err, os.ErrNotExist)` — a strict improvement over the oracle, invisible at the display boundary | Discard the cause exactly as the oracle does. Rejected: loses diagnostic power for zero contract benefit. |
| Separate `Is` (kind) from `Unwrap` (cause) | Stash **P1-j**: one `Unwrap` cannot serve both roles | A single mechanism. Rejected: ambiguous semantics. |
| Component-aware containment, not string prefix | A raw prefix check admits `/ws-evil` against root `/ws` | `strings.HasPrefix`. Rejected: a real bypass; the oracle's `Path::starts_with` is component-aware, so a naive Go port would be **less** safe than the oracle. |
| Port the oracle's TOCTOU behavior, documented | Fidelity to the oracle plus an explicit record beats a silent divergence | Add locking/`openat` hardening. Rejected: out of scope for a leaf slice and diverges from the parity corpus; recorded as a known limitation instead. |
| No `cmd/` wiring in this slice | Keeps the revert boundary at "delete the package directory" | Wire into `cmd/intercom` immediately. Rejected: creates downstream breakage on revert for no delivered behavior. |
| Sentinels are an unexported `kindSentinel` value type | Immutable; a `*Error` sentinel is a mutable global whose `Kind` any caller could corrupt | `var ErrX = &Error{kind: …}`. Rejected: mutable global hazard on a security-relevant taxonomy (finding GO-1). |
| `Error` fields are unexported with accessors | Prevents post-construction re-categorization of an error | Exported `Kind`/`Msg` as in `url.Error`. Rejected: the security context tips the balance (finding GO-8). |
| A `Root` type canonicalizes once | `Resolve` would otherwise re-run an `EvalSymlinks` syscall walk per call; P10 validates many paths per patch | `Resolve(root, candidate string)`. Rejected on performance and API-safety grounds (finding GO-4). |
| UNC stripping is a **Go-specific adaptation**, not oracle parity | Go's containment comparison is string-based; Rust's `Path::starts_with` is component-aware and needs no stripping. `strip_unc_prefix` lives in `src/config.rs:24`, not `path_safety.rs` | Claim oracle parity (as the first draft did). Rejected: factually wrong (findings SEC-4 / ARCH-4). |
| Case-folded containment on Windows only | A config-supplied root and an `EvalSymlinks` result can differ in case, causing false rejections | Case-sensitive everywhere. Rejected: silently rejects valid Windows paths (finding GO-7). |
| The oracle's absolute-path fast path is not ported | B2 rejects all absolute candidates, so the fast path is unreachable | Port it. Deferred to P10, which must pass relative paths or revisit (finding SEC-3). |
| `os.Stat` (not `os.Lstat`) gates the symlink check | Matches the oracle's `Path::exists()`, which follows symlinks | `os.Lstat`. Rejected for oracle fidelity; the broken-symlink consequence is documented (finding GO-14). |

## Risks and Caveats

| Risk | Severity | Mitigation |
|---|---|---|
| A naive `strings.HasPrefix` containment check admits sibling-directory escapes | **High** | Explicit acceptance criterion B1.4 with a dedicated `<root>-evil` negative test. |
| Windows path semantics (volume roots, `\\?\` UNC, case-insensitivity) diverge from the Unix oracle | High | B2 strips the UNC prefix mirroring the oracle's `strip_unc_prefix`; B1.3 includes a Windows absolute-path case. CI already runs the 4-target cross-compile matrix. |
| Symlink test is Unix-only and silently skips on Windows, hiding a regression | Medium | Test must call `t.Skip` with an explicit reason (never pass vacuously), matching the oracle's Unix gating; the skip is visible in CI output. |
| Display-format drift from the oracle | Medium | A1.2 asserts prefix uniqueness and absence of a trailing period across all 14 kinds in one table test. |
| TOCTOU gap inherited silently | Medium | Documented in the package doc comment and in this plan as accepted; not left implicit. |
| Dirty `.gitignore` / untracked `.claude/` staged accidentally | **High** | Forbidden-path invariant; every commit uses explicit pathspecs. See Plan Hardening. |
| `references/herdr` disturbed | Medium | Forbidden to all units; read-only and out of scope. |
| Oracle working tree (27 dirty files) mutated during parity checks | **High** | Oracle is strictly read-only; no `git` write commands may be issued in `C:\Source\GitHub\intercom`. |
| Windows drive-relative `C:foo` accepted as a normal component (NTFS ADS reference) | **High** | B2 rejects on `filepath.VolumeName != ""` **before** cleaning, with an explicit test (finding SEC-1). |
| `Wrap(kind, nil)` panics on a nil-interface method call | High | A3 guards `cause == nil`; pinned by acceptance criterion A3.2 (finding GO-3). |
| `errors.Is` matches a kind found deeper in the cause chain, surprising a caller | Medium | Documented in `Is` godoc and asserted by A2.3; this is intended `errors.Is` semantics, not a defect (finding GO-2). |
| Hardlink aliasing an external file passes validation | Medium | `EvalSymlinks` does not resolve hardlinks — recorded in the package doc comment alongside TOCTOU as an accepted oracle-parity limitation (finding SEC-5). |
| `Newf` format strings unchecked by `go vet`'s printf analyzer | Low | Documented in godoc; call sites are few in this slice and reviewed manually (finding GO-11). |

## Plan Hardening Signals

* Implements a **security control** (workspace isolation / path containment) that brief
  §10 marks NON-NEGOTIABLE — yes.
* Handles **symlink resolution and filesystem boundary enforcement** — yes.
* Freezes an **operator-facing contract** (error display strings) that later phases and
  the parity corpus depend on — yes.
* Touches credentials or secrets — **no** (deliberately deferred to P3).
* Modifies CI, build, or release topology — no.
* Introduces new dependencies — no.

**Requires plan hardening: yes**

## Runtime Verification and Closure

**Per-unit gate** (deterministic, binary outcome, runnable without any later phase):

```powershell
gofmt -l internal/ ; go vet ./internal/... ; go test -race ./internal/... ; go build ./...
$env:CGO_ENABLED=0; go build ./...
```

**Slice-level closure:** absorbed when

1. `go test -race ./internal/apperr/... ./internal/pathsafe/...` is green with all 16
   ported oracle behaviors covered per the traceability table (7 error + 9 path-validation),
   plus the 6 Go-port-specific tests added by review findings;
2. `CGO_ENABLED=0 go build ./...` succeeds (CGO-free mandate preserved);
3. `golangci-lint run` and `staticcheck` are clean;
4. the existing `001-S` CI gates (`test`, `lint`, `security`, 4-leg `cross-compile`,
   `pipeline-topology`, `ci gate`) are all green on the PR;
5. `docs/oracle-pin.md` exists and passes `markdownlint`;
6. `git status --porcelain` shows `.gitignore` still modified-but-unstaged, `.claude/`
   still untracked, and `references/herdr` absent from the output.

## Plan Hardening

**Hardening required: yes.** Triggered by the workspace-isolation security control, the
symlink/boundary enforcement surface, and the freezing of an operator-facing error
contract.

**Learnings consulted:** `docs/compound/` contains two durable learnings.
`go-race-requires-cgo-fails-closed-2026-09-03.md` is directly applicable — `-race`
requires CGO, so the `-race` gate runs as a CGO-enabled CI job while release builds stay
`CGO_ENABLED=0`; this slice inherits that topology unchanged and must not "fix" the
apparent contradiction. `external-spec-yields-to-workspace-instructions-2026-09-03.md`
is applied in the deliberation's MCP decision.

### Protected invariants

* **I1 — Dirty-state preservation.** `.gitignore` is modified in the working tree by an
  unrelated local change and MUST remain modified-but-unstaged. `.claude/` MUST remain
  untracked. `.backlogit/hooks_queue.jsonl` is machine-local runtime state and MUST NOT
  be committed. `references/herdr` MUST NOT be staged, committed, modified, or removed.
* **I2 — Oracle immutability.** `C:\Source\GitHub\intercom` is read-only. No write, no
  `git add/commit/checkout/stash/clean/restore`, no branch operations. Its 27 pre-existing
  uncommitted changes MUST remain exactly as found.
* **I3 — Containment is component-aware.** Never compare paths with raw string prefixes.
* **I4 — Display contract.** The 14 rendered prefixes are frozen; changing one is a
  breaking change requiring a new deliberation.
* **I5 — CGO-free.** `CGO_ENABLED=0 go build ./...` must succeed at every commit.
* **I6 — `.gitignore` is append-only** (inherited invariant from `001-S`).
* **I7 — No new module dependencies.** `go.mod`/`go.sum` are not modified by this slice.

### Risky actions

| Action | Risk | Control |
|---|---|---|
| `git add` during commit | Staging protected dirty state | Explicit pathspecs only (`git add internal/apperr internal/pathsafe docs/oracle-pin.md`). **Never** `git add -A`, `git add .`, or `git commit -a`. |
| Creating symlinks in tests | Test pollution / privilege failure on Windows | Create only under `t.TempDir()`; probe for symlink privilege and `t.Skip` with an explicit reason on failure. |
| `filepath.EvalSymlinks` on a non-existent path | Returns an error, not a violation | Existence is checked first; non-existent paths take the lexical-only branch (oracle step 7). |
| Reading the oracle for parity | Accidental mutation | Read-only commands only (`Get-Content`, `Select-String`, `view`). |

### Forbidden commands (require explicit operator approval)

* `git add -A` / `git add .` / `git commit -a` / `git stash` / `git clean` /
  `git checkout -- .` / `git restore .` anywhere in `intercom-go`.
* Any mutating `git` command with `-C C:\Source\GitHub\intercom` or executed from
  within the oracle or `references/herdr`.
* `go mod tidy` / `go get` (would violate I7 and touch `go.sum`).
* Any modification to `.github/workflows/`, `scripts/`, `.gitignore`, or `go.mod`.
* Deleting or moving `references/herdr`.

### Deepened verification

* Negative tests are mandatory, not optional: sibling-directory escape (`<root>-evil`),
  deep traversal (`src/../../secret.txt`), boundary traversal
  (`subdir/../../escape.txt`), absolute path, Windows drive-relative (`C:foo`),
  empty candidate, and symlink escape.
* The symlink test MUST NOT pass vacuously — it either exercises the escape or calls
  `t.Skip` with a stated reason, gated on a runtime `os.Symlink` probe rather than
  `runtime.GOOS`.
* Prefix uniqueness is asserted programmatically across all 14 kinds, not by inspection.
* After the final commit, `git status --porcelain` output is captured and compared
  against the pre-session fingerprint recorded in the session memory checkpoint.

#### Oracle parity traceability (16 tests → covering unit)

Closure requires 16-of-16 coverage; this table makes that check mechanical rather than a
hand-count (finding SCOPE-5).

| # | Oracle test (`tests/unit/…`) | Pins | Covered by |
|---|---|---|---|
| 1 | `error_tests::acp_error_display_starts_with_acp_prefix` | prefix `acp:` | A1.1, A1.2 |
| 2 | `error_tests::acp_error_display_includes_message` | `"acp: stream closed"` | A1.2 |
| 3 | `error_tests::acp_error_message_no_trailing_period` | no trailing `.` | A1.1 |
| 4 | `error_tests::acp_error_is_distinct_from_io_error` | prefix distinctness | A1.1 |
| 5 | `error_tests::acp_error_is_distinct_from_mcp_error` | prefix distinctness | A1.1 |
| 6 | `error_tests::acp_error_implements_std_error_trait` | satisfies `error` | A1.2 |
| 7 | `error_tests::acp_error_debug_representation` | kind recoverable | A2.1, A3.3 |
| 8 | `path_validation::allows_path_inside_workspace` | relative resolves in-root | B3.1 |
| 9 | `path_validation::rejects_traversal` | `../secret.txt` | B2.1 |
| 10 | `path_validation::rejects_deep_traversal` | `src/../../secret.txt` | B2.1 |
| 11 | `path_validation::allows_relative_subdirectory` | nested relative | B3.1 |
| 12 | `path_validation::allows_dot_segment` | `./src/main.rs` | B2.3, B3.1 |
| 13 | `path_validation::rejects_workspace_root_boundary` | `subdir/../../escape.txt` | B2.1 |
| 14 | `path_validation::rejects_symlink_escape` | symlink out of root | B4.1 |
| 15 | `path_validation::path_safety_allows_non_existent_file` | non-existent accepted | B4.3 |
| 16 | `path_validation::path_safety_rejects_invalid_workspace` | bad root | B1.2 |

Go-port-specific tests **beyond** oracle parity (added by review findings): B2.2
(`C:foo` drive-relative, SEC-1), B3.2 (`<root>-evil` sibling, GO-6), B3.3 (empty
candidate, SEC-2), B2.3 (`a/b/../c.txt` `Clean`-equivalence, GO-13), A3.2 (`Wrap` nil
guard, GO-3), B1.3 (`NewRoot` idempotence).

### Deepened closure

Closure additionally requires the dirty-state evidence in gate 6 above to match the
pre-session fingerprint byte-for-byte for `.gitignore` (SHA-256
`E4D7071C607C953D9BF330336DD189CD76C94B922D6FCCE17F0105E6D031E291`).

### Operator checkpoints

* Before merge: confirm the staging PR is not auto-merged unless the controlling
  workflow authorizes it.
* `EF9352FB` (Tavily key rotation in `.env.local`) remains an open operator action,
  unrelated to this slice.

### Review-gate capability risks

No credential handling, no CI topology change, no new dependencies, and no network
surface in this slice. The concentrated risk is entirely in `pathsafe` correctness,
which is why B1 and B2 carry the heaviest negative-test requirements.

### Unresolved operator decisions

None blocking this slice. Roadmap-level open questions (SQLite latency, ACP SDK choice,
the oracle hot-reload defect, `PatchConflict` semantics) are recorded in the deliberation
and are scoped to P2/P4/P6/P10.

## Plan Review

**Verdict: PASS** (attempt 3, after remediation of all P0 and P1 findings across two rounds).

### Attempt history

| Attempt | Verdict | Outcome |
|---|---|---|
| 1 | FAIL — 1×P0, 12×P1 | Findings remediated in place; units resized and mechanisms specified. |
| 2 | ADVISORY — 2×P1 (new, introduced by the resize) | Both remediated; claims independently verified by empirical Go probe. |
| 3 | **PASS** | All P0/P1 closed. Residual P2/P3 accepted or deferred with traceability. |

### Persona coverage

Five independent reviewers ran against the plan, the deliberation, and the Rust oracle at
`41df772`: **Security Lens**, **Go Reviewer**, **Architecture Strategist**, **Scope
Boundary Auditor** (attempt 1, in parallel), and **Correctness Reviewer** (attempt 2,
which also re-verified every attempt-1 closure against the plan body rather than trusting
the review record). The schema/CLI/docs coupling dimension was folded into the
architecture pass — this slice ships no schema, CLI, or operator-docs surface beyond
`docs/oracle-pin.md`.

### P0 findings — closed

| ID | Finding | Resolution |
|---|---|---|
| GO-1 | Sentinel concrete type unspecified; `errors.Is` correctness depends on it | A2 now specifies an unexported immutable `kindSentinel` value type with the exact `Is` implementation. |

### P1 findings — all closed

| ID | Finding | Resolution |
|---|---|---|
| SEC-1 | Windows drive-relative `C:foo` bypasses `filepath.IsAbs`; NTFS ADS exposure | B2 rejects on `filepath.VolumeName != ""` before cleaning; criterion B2.2. |
| GO-2 | `Is`+`Unwrap` together make `errors.Is` match a kind anywhere in the chain | Documented in godoc; asserted by A2.3. Intended semantics, now explicit. |
| GO-3 | `Wrap(kind, nil)` panics | A3 guards nil; criterion A3.2. |
| GO-4 | `Resolve` re-resolves the root on every call | Introduced the `Root` type (B1); `Resolve` is now a method. |
| GO-5 | Go has no `Path::components()`; walk mechanics unspecified | B2 specifies `VolumeName` → `IsAbs` → `Clean` → split → walk, incl. the `Clean` equivalence argument and criterion B2.3. |
| GO-6 | Containment mechanism unspecified | B3 mandates `hasPathPrefix` with exact-equality or separator-terminated prefix; raw `strings.HasPrefix` forbidden. |
| GO-7 | Windows case-insensitivity unhandled → false rejections | B3 folds case only when `GOOS == "windows"`. |
| ARCH-1 | `acp/reader.rs:48` imports `slack::client`, inverting P6/P8 | Deliberation now introduces a consumer-side `Notifier` interface at P6, implemented by Slack at P8. |
| SCOPE-1 | A2 declared 5 functions (limit 4) | Split into A2 (sentinels + `Is`, 2 funcs) and A3 (constructors + `Unwrap`, 4 funcs). |
| SCOPE-2 | A2 declared 4 test scenarios (limit 3) | Split; both A2 and A3 now have 3. |
| SCOPE-3 | B1 declared 4 test scenarios | Old B1 split into B1/B2/B3, each with 3. |
| SCOPE-4 | B2 declared 4 test scenarios at complexity high | Old B2 split into B1 (root) and B4 (symlink), each with 3. |

### Attempt-2 findings (introduced by the resize) — closed

| ID | Finding | Resolution |
|---|---|---|
| CORR-1 | A2 criterion 1 called `New()`, which is not defined until A3 — the criterion was uncompilable at A2 time, breaking independent verifiability | A2.1 rewritten to use a package-internal struct literal, with a note explaining why. |
| CORR-2 | B3 criterion 3 required rejecting an empty candidate, but no step in the specified algorithm rejected it — `Clean("")`→`"."`, walk skips `"."`, `Join(root,"")`→`root`, containment passes | B3 now mandates an explicit zero-component guard in `Resolve`, and records that this is a deliberate tightening beyond the oracle. |
| CORR-3 | The `Clean`-equivalence claim was asserted rather than argued | B2 now states *why* the equivalence holds (redundant/trailing separators are elided identically by `Clean` and by Rust's `Normal`-only component iteration) and lists the verified `Clean` outputs. |

**Empirical verification.** Rather than reasoning about Go path semantics from
documentation, the load-bearing claims were confirmed by executing a throwaway probe
(outside the repository, deleted afterwards):

| Expression | Result | Consequence |
|---|---|---|
| `filepath.Clean("")` | `"."` | Confirms CORR-2 — empty input needs an explicit guard. |
| `filepath.Join(root, "")` | `root` | Confirms the empty candidate resolves to the root. |
| `filepath.VolumeName("C:foo")` | `"C:"` | Confirms the SEC-1 pre-check fires. |
| `filepath.IsAbs("C:foo")` | `false` | Confirms `IsAbs` alone is insufficient — SEC-1 is a real bypass. |
| `filepath.Clean("../secret.txt")` | `"../secret.txt"` | Leading `..` preserved, so B2's rejection fires. |
| `filepath.Clean("src/../../secret.txt")` | `"../secret.txt"` | Deep traversal still rejected post-`Clean`. |
| `filepath.Clean("a/b/../c.txt")` | `"a/c.txt"` | Safe interior traversal preserved. |

### Technical corrections adopted

* **SEC-4 / ARCH-4 — factual provenance error.** The first draft claimed UNC stripping
  matched "the oracle's `strip_unc_prefix`" *in `path_safety.rs`*. Verified false:
  `strip_unc_prefix` is defined at `src/config.rs:24` (and separately at
  `src/acp/handshake.rs:402`) and is **never called from `src/diff/path_safety.rs`**,
  which relies on Rust's component-aware `Path::starts_with`. Corrected in B1 and in the
  Decisions table to describe it as a Go-specific adaptation.
* **ARCH-2 — circular phase dependency.** `src/persistence/session_repo.rs:8-11` imports
  `models::{progress,session}`, so P4's `session_repo` gate could not pass with models in
  P5. `internal/models` moved into P4.
* **ARCH-3 — cycle over-stated.** `driver/mod.rs` and `driver/acp_driver.rs` import
  neither `state` nor `slack`; only the *retired* `mcp_driver.rs:18` imports `state`.
  The real cycle is `state ↔ slack`, and D1's retirement of `McpDriver` removes the
  `driver → state` edge by construction. Corrected in the deliberation.

### Residual P2 / P3 — accepted, recorded as follow-ups

* **ARCH-5 (P2)** — the oracle pin is documentary only; no automated drift check.
  Deferred to a successor stash entry rather than expanding this slice with a new script
  and CI wiring.
* **SEC-3 (P2)** — the oracle's absolute-path fast path is not ported; P10 must pass
  workspace-relative paths or revisit. Recorded in B3 and the Decisions table.
* **SEC-5 (P3)** — hardlinks are not resolved by `EvalSymlinks`. Documented as an
  accepted oracle-parity limitation in the B4 package doc requirement.
* **GO-11 (P2)** — `Newf` is invisible to `go vet`'s printf analyzer. Documented; becomes
  more material at P2 where config validation adds many call sites.
* **GO-14 (P2)** — `os.Stat` treats a broken symlink as non-existent. Chosen for oracle
  fidelity; consequence documented.
* **SCOPE-6 (P3)** — C1's content criterion tightened from "names …" to literal token
  matching.

### Gate rationale

**PASS.** All P0 and P1 findings from both rounds are closed in the artifacts, and
attempt 2 independently re-verified every attempt-1 closure against the plan body rather
than trusting the review record. The plan now specifies every load-bearing implementation
mechanism rather than leaving it to the executor; all eight units satisfy the
< 3 files / < 5 functions / < 4 test scenarios bound; every acceptance criterion is
deterministic and machine-checkable; the 16 oracle behaviors are mapped to covering
criteria; and the three claims most likely to be wrong were confirmed empirically rather
than assumed. Hardening is present and proportionate to the security control being
ported. Residual findings are P2/P3, each either recorded in-plan or deferred with an
explicit successor.

<!-- plan-review-attempt: 3 -->
<!-- plan-review-verdict: PASS -->
<!-- plan-hardening: applied -->
<!-- p0-open: 0 -->
<!-- p1-open: 0 -->


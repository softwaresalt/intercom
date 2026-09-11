---
title: "Deliberation — Pathsafe Error-Surface Parity and Windows Long-Path Precision"
date: 2026-09-10
status: accepted
agent: Stage
mode: DARK_MODE_ACTIVE
governs: 017-F / shipment 016-S
---

# Deliberation — Pathsafe Error-Surface Parity and Windows Long-Path Precision

**Date**: 2026-09-10
**Agent**: Stage (dark-factory cycle, operator AFK)
**Mode**: DARK_MODE_ACTIVE — bounded stash scope, exactly one shipment
**Stash entries in scope**: BF5DE670, 6B751D8B, 7ADAC481
**Excluded by operator**: A92E3FA0, 37FAB8C2, 4A01C53E, 1C6C3B46, 2787DA56,
4C5BEC23, 6C24E2E4, 9D45E62E, EF9352FB, F47DB9A9
**Governing prior artifacts**:

- `docs/closure/2026-09-10-015-s-016-f-pathsafe-windows-canonicalization-symmetry-adversarial-review.md`
  (the Ship-owned residual-risk record that captured `7ADAC481` and `6B751D8B`)
- `docs/decisions/2026-09-10-intercom-go-pathsafe-windows-canonicalization-symmetry-deliberation.md`
- `docs/plans/2026-09-10-intercom-go-pathsafe-windows-canonicalization-symmetry-plan.md`
- `docs/compound/2026-09-08-verify-go-stdlib-internals-claims-against-goroot-source.md`
- `docs/compound/2026-09-08-adversarial-review-empirical-verification-and-scope-check.md`
- `docs/compound/build-errors/go-race-requires-cgo-fails-closed-2026-09-03.md`
- `docs/compound/2026-09-04-backlogit-sizing-is-wit-gated.md`

---

## 1. Why this deliberation exists (P-021 C6)

Two of the three scoped entries (`BF5DE670`, `7ADAC481`) carry the literal
`DEFERRED SCOPE EXPANSION` marker written by Ship's P-021 C2 capture. Under the
Stage triage precedence rule this **forces** the `deliberate` route regardless
of apparent shape, size, priority, or triviality, and neither may reach
implementation planning without this artifact.

`6B751D8B` carries no marker (it opens `Follow-up:`) and is an advisory-only
consolidation from the same 015-S adversarial review round. It is deliberated
here because it shares a code surface and a review provenance with `7ADAC481`,
not because P-021 compels it.

## 2. P-021 C5 intake obligations — both discharged

### (A) Duplicate detection — UNCONDITIONAL

Performed on all three scoped entries regardless of source-ref population and
regardless of the absence of any `DISCOVERY-STATUS` token. All **13** active
entries enumerated via `backlogit stash list` and compared pairwise against the
three scoped entries.

| Scoped entry | Nearest neighbour considered | Verdict |
|---|---|---|
| `BF5DE670` | `37FAB8C2` (pathsafe runtime verification, also gated on "a live caller exists") | **Not a duplicate** — same gating *shape*, different deliverable: black-box integration test vs. TOCTOU/hardlink *mitigation*. |
| `BF5DE670` | `1C6C3B46` (pathsafe decision/plan doc artifacts) | Not a duplicate — documentation bookkeeping. |
| `7ADAC481` | `6B751D8B` (same review round, same package) | **Not a duplicate** — the closure record enumerates them as separate dispositions: `7ADAC481` is the P-021 out-of-scope finding, `6B751D8B` is the advisory-only bucket. Disjoint subject matter. |
| `6B751D8B` | `7ADAC481` | As above. |

**DUPLICATE SCAN: CLEAN.** No duplicate found; nothing merged, archived, or
destroyed. Recorded explicitly because an unrecorded clean scan is
indistinguishable from a scan that never ran.

### (B) Late-identifier reconciliation — MANDATORY where any source ref is `N/A`

**`7ADAC481`** — captured with `PR=N/A (pre-PR local review finding)` and
`review-thread=N/A (threadless path)`.

- **PR reconciled `N/A` → `#48`.** Recovered from the Ship-owned residual-risk
  record (`docs/closure/2026-09-10-015-s-016-f-...-adversarial-review.md`)
  joined on shipment `015-S`. The 015-S release unit merged as **PR #48**
  (merge `fbd8d8d0ab679e254b2b817af69b8e8e673626c5`), closure **PR #49**
  (merge `0ea6c87b9d86c25c1c0326fcfed77ab5fcf27ff6`), Copilot-review
  remediation **PR #50** (merge `1ff2b426a6c95533ec58768894c439931a110b5c`).
  At capture time no PR existed; one exists now. Reconciled under Stage's own
  stash authority — **no Ship write**, single-write capture invariant intact.
- **review-thread REMAINS `N/A`** — a truthful terminal record, not a
  shortfall. The finding was raised at the 015-S *local* review gate before any
  PR existed, so no thread was ever created (threadless discharge per C3).
- `task=016.007-T`, `feature=016-F`, `shipment=015-S` were already populated.

**`BF5DE670`** — `PR` was already reconciled to `#27` / closure `#28` in the
015-S cycle. `task` remains `N/A` (genuinely cross-cutting; no single
originating task) and `review-thread` remains `N/A` (raised at the 009-S local
review gate, pre-PR). Re-running reconciliation is an **idempotent no-op**: no
concrete identifier was overwritten and no `N/A` was rewritten over a concrete
value.

**`6B751D8B`** — not a deferred-expansion entry; carries no six-field source-ref
payload. Reconciliation not triggered. Its provenance (015-S adversarial review,
commit `6997700`) is recorded in the closure artifact.

Neither reconciliation gated deliberation.

## 3. Severity correction discovered during this deliberation

`7ADAC481` carries `provisional_priority: low`. The Ship-owned closure record
states the Architecture Strategist **rated it P1 in isolation**, and that it was
deferred purely on P-021 C1 *scope* grounds — not on severity grounds:

> Architecture Strategist (rated P1 in isolation): `checkSymlinkEscape`'s
> `canonicalizeReparse` error path is not wrapped via `wrapPathError` the way
> `NewRoot`'s is, so its error surface is inconsistent with `NewRoot`'s.

Under the operator's prioritization framework (reliability and security
supersede feature work), `7ADAC481` is therefore the **highest-priority item in
this cycle**, not the lowest. Re-prioritized `low` → `high` for harvest.

## 4. Problem frame — evidence gathered this cycle (not inherited)

### 4.1 `7ADAC481` — error-surface asymmetry (VERIFIED)

`internal/pathsafe/root.go` `NewRoot` (line ~346):

```go
resolved, err := canonicalizeReparse(abs)
if err != nil {
    return Root{}, wrapRootInvalid(wrapPathError("open", abs, err))
}
```

`internal/pathsafe/pathsafe.go` `checkSymlinkEscape` (line ~359):

```go
real, err := canonicalizeReparse(ancestor)
if err != nil {
    return "", apperr.Wrapf(apperr.KindPathViolation, err, symlinkUnverifiableMsg)
}
```

On Windows `canonicalizeReparse` returns a raw `syscall.Errno`. `NewRoot`
re-forms it into `*fs.PathError`; `checkSymlinkEscape` does not. So
`errors.As(err, &*fs.PathError)` matches a `NewRoot` failure but **not** the
functionally-equivalent `Resolve` → `checkSymlinkEscape` failure. Confirmed by
reading both call sites at `1ff2b42`.

**Fix feasibility verified**: `wrapPathError` lives in `root.go`, which carries
**no build tag** — same package, platform-agnostic, directly callable from
`pathsafe.go`. No new dependency, no build-tag gymnastics.

**Chain-preservation verified**: `apperr.Wrapf` retains its `cause`
(`&Error{..., cause: cause}`) and `*apperr.Error` implements `Unwrap()`, so
`errors.As` traverses the `apperr` layer into the injected `*fs.PathError`.
Read directly from `internal/apperr/apperr.go`.

### 4.2 `6B751D8B` item (1) — the stated premise is FALSE (MATERIAL CORRECTION)

The entry asserts the `\\.\` device-namespace branch is *"currently unreachable
given `filepath.IsAbs`'s own device-path handling."*

**Empirically falsified this session** with an isolated `go run` probe
(compound-learning discipline: never write a reachability claim without tracing
or reproducing it):

| Input | `filepath.IsAbs` | `filepath.VolumeName` |
|---|---|---|
| `\\.\C:\foo` | **true** | `\\.\C:` |
| `\\.\PhysicalDrive0` | **true** | `\\.\PhysicalDrive0` |
| `\\.\UNC\srv\sh\x` | **true** | `\\.\UNC\srv\sh` |
| `\\.\C:\` + 300×`a` (len 307) | **true** | — |

`filepath.IsAbs` returns `true` for `\\.\` device paths, including beyond the
threshold. The `!filepath.IsAbs` early return therefore does **not** shadow the
branch. **The branch is reachable.**

What is actually true is narrower: within `addLongPathPrefix`'s guard order the
`\\.\` branch is **reachable but behaviourally redundant**, because the very
next guard (`strings.HasPrefix(path, "\\\\")`) returns the identical value for
any `\\.\` input. Deleting the branch would change no observable output today —
but only because of that adjacent guard.

This is precisely the failure mode the compound library warns about twice:
*"do not build defensive scaffolding on an unverified premise"* and *"narrow the
claim, twice — qualify which caller, which precondition, which namespace."*

### 4.3 `6B751D8B` item (2) — UTF-8 byte length as a UTF-16 proxy (VERIFIED, conservative)

`addLongPathPrefix` gates on `len(path) < longPathThreshold` where
`longPathThreshold = syscall.MAX_PATH` (260). `len()` is UTF-8 **bytes**;
Windows measures UTF-16 **code units**.

Measured this session:

| Path content | UTF-8 bytes | UTF-16 units | runes |
|---|---|---|---|
| 200 × `é` (U+00E9) | 400 | **200** | 200 |
| 100 × 😀 (U+1F600) | 400 | **200** | 100 |

UTF-8 byte length is **always ≥** UTF-16 code-unit count (1→1, 2→1, 3→1, 4→2).
The proxy therefore fires **at or before** the true limit — it can produce a
*false positive* (prefixing a path that did not need it) but **never a false
negative**. Applying `\\?\` unnecessarily is benign here because both callers
pass already-`Abs`/`Clean`'d input (documented precondition).

**Conclusion: this is a precision improvement, not a correctness or security
defect.** Any framing of it as a vulnerability fix would be an overclaim and is
an explicit anti-goal below.

### 4.4 `BF5DE670` — trigger RE-MEASURED, HAS NOT FIRED

This entry's own archival condition is *"archive when the mitigation lands or
the gate fires."* Measured directly at `1ff2b42` this session:

- `bash scripts/check-write-path-precondition.sh` → **exit 0** (clean).
- `--self-test` → **5/5 fixtures pass** (`accept-clean`,
  `accept-mentions-in-comment`, `reject-createtemp`, `reject-link`,
  `reject-writefile`) — so the gate is genuinely functional, not vacuously green.
- **Independent** PowerShell grep across all **17** non-test `.go` files under
  `internal/**` and `cmd/**` for `os.WriteFile|Create|CreateTemp|OpenFile|
  Remove|RemoveAll|Rename|Mkdir|MkdirAll|Symlink|Link|Chmod|Truncate` →
  **zero matches**.

`pathsafe` still gates **no** destructive write path. The entry's PRIMARY ask
(the consolidated risk register) remains DISCHARGED by `011.004-T`. The residual
**mitigation** (Lstat-before-write, O_EXCL revalidation, hardlink-count checks)
must attach to a filesystem write call site, and none exists.

## 5. Options considered

### 5.1 Grouping options

**Option G1 — all three entries produce tasks (reject).** Would require
inventing a write API to hang the `BF5DE670` mitigation on. That is *scope
expansion*, a declared stop condition, and is the speculative engineering that
this entry's own three prior Stage dispositions already rejected on the same
measured grounds.

**Option G2 — `7ADAC481` alone (reject).** Leaves `6B751D8B`'s two named items
stranded despite being same-surface, already-reviewed, and cheap. The compound
library's own precedent (005-S) is that small same-surface correctness /
portability findings get fixed in-shipment.

**Option G3 — `7ADAC481` + `6B751D8B` items (1)(2); `BF5DE670` preserved as an
explicit deferred acceptance criterion; `6B751D8B` items (3)(4) preserved as
explicit follow-ups (CHOSEN).** Smallest coherent shipment that safely resolves
the implementable portion of the scope while preserving — never silently
dropping — the portion that is not yet implementable.

### 5.2 `6B751D8B` item (1) fix options

- **O1-a — delete the redundant branch.** Rejected: removes self-documenting
  intent and the defence-in-depth margin, and silently couples correctness to
  the adjacent generic `\\` guard's continued existence.
- **O1-b — add a `panic`/assertion that the branch is unreachable.** Rejected
  outright: the premise is **false** (§4.2), and the compound library records a
  prior incident (`go-race-requires-cgo`) of exactly this — scaffolding built on
  an assumed failure mode, which became review-finding surface for no safety
  benefit.
- **O1-c — keep the branch, add a locking test that independently verifies
  reachability, and correct the stale prose (CHOSEN).** Converts an unverified
  claim into a mechanically enforced, order-independent fact.

### 5.3 `6B751D8B` item (2) fix options

- **O2-a — rune count (`len([]rune(path))`).** Rejected: wrong unit. A
  surrogate pair is 1 rune but **2** UTF-16 code units, so a rune count can
  *under*-count and introduce the false negative the current code never has.
- **O2-b — `len(utf16.Encode([]rune(path)))` (CHOSEN, post-review).** Correct
  and stdlib-only (`unicode/utf16`), no new module dependency — consistent with
  decision D-3 (014.003-T).
- **O2-c — a bespoke non-allocating UTF-16 code-unit counter. WITHDRAWN at
  plan review** (Scope Boundary Auditor, P3 gold-plating). `addLongPathPrefix`'s
  only caller immediately calls `syscall.UTF16PtrFromString` (which allocates)
  and then two Win32 syscalls, so saving two small allocations buys nothing
  while adding bespoke infrastructure to a NON-NEGOTIABLE control. A plain
  `for _, r := range path` accumulator is an acceptable equivalent.

## 6. Decisions

| ID | Decision |
|---|---|
| **D-1** | Ship `7ADAC481` + `6B751D8B` items (1)(2) as feature **017-F** under shipment **016-S**. Option G3. |
| **D-2** | Re-prioritize `7ADAC481` `low` → **high**; it was deferred on scope grounds, not severity (Architecture Strategist P1-in-isolation). It is the lead task. |
| **D-3** | Fix `7ADAC481` by routing `checkSymlinkEscape`'s `canonicalizeReparse` error through the **existing** `wrapPathError` with `Op: "open"` — matching `NewRoot`'s corrected label. Do **not** move normalization into `canonicalizeReparse` itself: that would change a Windows-only internal contract relied on by two callers and enlarge blast radius on a NON-NEGOTIABLE control. |
| **D-4** | Do **not** delete the `\\.\` branch. Keep it, lock its reachability with a test, and correct the prose to "reachable but intentionally redundant", recording the verification method inline. |
| **D-5** | Make the MAX_PATH comparison UTF-16-code-unit-aware. Frame it explicitly as **precision, not a security fix** — the prior proxy was conservative in the safe direction. |
| **D-6** | `BF5DE670` is **RETAINED, NOT ARCHIVED**, and contributes **no task** this cycle. Its trigger was re-measured and has not fired. It is preserved as an explicit acceptance criterion on 017-F (the register and the gate must survive this change intact and accurate). |
| **D-7** | `6B751D8B` items **(3) buffer pooling** and **(4) CI lint scoping** are **out of scope** and preserved in the entry. Both are performance/CI-time optimizations with zero correctness or security content; the operator's scope names only items (1) and (2). Simplicity supersedes complexity. |
| **D-8** | `6B751D8B` is **RETAINED, NOT ARCHIVED** (items 3–4 survive). `7ADAC481` **is** archived on harvest — fully consumed. |
| **D-9** | Sequence `017.002-T` → `017.003-T`. *(Amended at plan review: this is a **same-function merge-coordination** constraint, not a behavioural regression-net dependency — 003 changes only the length guard and cannot regress 002's device-path property.)* |
| **D-10** | *(Added at plan review, Security Lens Reviewer P2 + Correctness Reviewer P3.)* **Accept** that the UTF-16 precision gain re-enables ordinary Win32 path normalization (trailing dot/space stripping, reserved device names) for the narrow band of paths that are ≥260 UTF-8 bytes but <260 UTF-16 code units. Those paths thereby behave exactly as every sub-MAX_PATH path already does today. The prior "verdict-neutral by construction" claim is **withdrawn** and replaced with a scoped, accurate statement. Recorded as an accepted residual, consistent with the package's existing oracle-parity risk acceptances. |
| **D-11** | *(Added at plan review, Constitution Reviewer P1.)* Declare safety modes **`freeze-scope`** and **`investigate-first`** by name for this work, as Principle VIII requires for elevated-blast-radius work on a NON-NEGOTIABLE control during an unattended cycle. |

## 7. Scope boundary (anti-goals)

This feature explicitly does **NOT**:

1. Add any filesystem write primitive, write API, or store constructor — doing
   so would trip `scripts/check-write-path-precondition.sh` and constitutes
   scope expansion (stop condition).
2. Implement Lstat-before-write, O_EXCL revalidation, or hardlink-count checks
   (`BF5DE670`'s deferred mitigation).
3. Remove, weaken, or retire the write-path precondition gate or the
   consolidated risk register in `root.go`.
4. Change `canonicalizeReparse`'s own error contract, or alter `NewRoot`'s
   behaviour in any way.
5. Change `apperr.Kind` values, the `symlinkUnverifiableMsg` text, or any
   `errors.Is` relationship that callers rely on.
6. Introduce buffer pooling (`6B751D8B` item 3) or rescope the GOOS=windows
   lint job (item 4).
7. Add a new module dependency.
8. Claim the UTF-16 change fixes a vulnerability.

## 8. Open questions

None blocking. One noted for Ship: the `\\?\UNC\...` long-path residual
(U5/AC5) remains an accepted, register-documented limitation and is **not**
touched here.

## 9. Dispositions

| Entry | Disposition |
|---|---|
| `7ADAC481` | **HARVESTED** into `017.001-T` (017-F / 016-S). Re-prioritized to high. PR ref reconciled `N/A` → `#48`. **ARCHIVE on harvest** — fully consumed. |
| `6B751D8B` | **PARTIALLY HARVESTED** — items (1)(2) → `017.002-T`, `017.003-T`. Items (3)(4) retained. **NOT ARCHIVED.** |
| `BF5DE670` | **RETAINED (deferred)** — trigger re-measured at `1ff2b42`, has not fired. No task. Preserved as an explicit AC on 017-F. **NOT ARCHIVED.** |
| All 10 excluded IDs | Untouched — not triaged, edited, harvested, archived, or included. |

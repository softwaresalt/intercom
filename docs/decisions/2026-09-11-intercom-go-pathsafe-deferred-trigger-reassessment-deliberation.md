# Deliberation — Pathsafe Deferred-Trigger Reassessment (BF5DE670 / 6B751D8B / 475E76D2)

- **Date:** 2026-09-11
- **Cycle:** dark-factory (`DARK_MODE_ACTIVE`, P-017), operator AFK
- **Stage route:** `claude-opus-5` (tier3)
- **Chartered scope:** EXACTLY three stash IDs, in operator priority order —
  `BF5DE670`, `6B751D8B`, `475E76D2`. No other stash item may be triaged,
  edited, harvested, archived, or included.
- **Measured at:** clean worktree, `main`, `e7948d59998fd8f54f72b6cbf5f3c414a630aaad`
- **Safety mode:** freeze-scope. Destructive operations NOT authorized.
- **Prior-cycle binding artifact:**
  `docs/decisions/2026-09-10-intercom-go-pathsafe-error-surface-parity-deliberation.md`
  (the immediately-preceding cycle over the same package and the same entry shapes)

---

## 1. Why this deliberation exists (P-021 C6)

Two of the three chartered entries — `BF5DE670` and `475E76D2` — carry the
literal `DEFERRED SCOPE EXPANSION` marker as their first field. Under the Stage
Step 1 **precedence** rule, that marker FORCES the `deliberate` route regardless
of the entry's apparent shape, size, priority, or triviality, and neither entry
may reach planning without a deliberation artifact.

This explicitly overrides `475E76D2`'s own `Requires deliberation: no` field.
The marker is a precedence rule, not a shape category; the entry's self-assessment
does not get to waive it.

`6B751D8B` carries no marker and is an ordinary task-shaped entry. It is
deliberated here because it is chartered and because its residual items share the
exact code surface (`internal/pathsafe/reparse_windows.go`) with `475E76D2`.

---

## 2. P-021 C5 intake obligations — both discharged

### (A) Duplicate detection — UNCONDITIONAL

Run over every chartered deferred-scope-expansion entry regardless of field
population and regardless of any `DISCOVERY-STATUS` token. Neither `BF5DE670`
nor `475E76D2` carries such a token; the scan ran anyway, because a duplicate
produced by a lookup that silently returned a false absence carries no token at all.

All **13 active stash entries** were compared pairwise. Nearest neighbours were
examined explicitly rather than dismissed on ID:

| Pair | Shared surface | Verdict |
|---|---|---|
| `BF5DE670` ↔ `37FAB8C2` | `internal/pathsafe`, both gated on a missing caller | **NOT a duplicate.** Same gating *shape*, different deliverable: `37FAB8C2` wants a black-box integration test of `Root.Resolve`; `BF5DE670` wants TOCTOU/hardlink *mitigation*. |
| `BF5DE670` ↔ `1C6C3B46` | pathsafe decision-doc bookkeeping | **NOT a duplicate.** Documentation bookkeeping vs. a security mitigation. |
| `475E76D2` ↔ `6B751D8B` | `addLongPathPrefix` in `reparse_windows.go`, shared 015-S/016-S adversarial-review provenance | **NOT a duplicate.** Disjoint residuals: `6B751D8B`'s live content is items (3) buffer pooling + (4) CI lint scoping; `475E76D2` is a fast-path guard ahead of `utf16Len`. `6B751D8B` items (1) and (2) already landed as `017.002-T` / `017.003-T`. |

**DUPLICATE SCAN: CLEAN.** Nothing merged, nothing archived, nothing destroyed.
Recorded explicitly because an unrecorded clean scan is indistinguishable from a
scan that never ran.

### (B) Late-identifier reconciliation — MANDATORY where any source ref is `N/A`

- **`BF5DE670`** — source refs `task=N/A`, `feature=010-F`, `shipment=009-S`,
  `PR=#27` (reconciled in the 015-S cycle, merge `d8ca53f5`; closure `#28`,
  merge `3d9515c2`), `thread=N/A`.
  **Outcome: idempotent NO-OP this cycle.** No concrete identifier was
  overwritten; no `N/A` was written over a concrete value. `task` remains `N/A`
  (genuinely cross-cutting — no single originating task) and `thread` remains
  `N/A` (raised at the 009-S *local* review gate before any PR existed, so no
  thread was ever created). Both stand as **truthful terminal records**, not
  shortfalls, and neither gated this deliberation.

- **`475E76D2`** — source refs `task=017.003-T`, `feature=017-F`,
  `shipment=016-S`, `PR=N/A (pre-PR local review finding)`,
  `review-thread=N/A (no thread; local Go Reviewer pass)`.
  **Outcome: RECONCILED — a late identifier surfaced.** Shipment `016-S` has
  since merged. Joining on the shipment ID against Ship-owned merged-PR history
  recovers:
  - **PR #51** — `fix(pathsafe): error-surface parity and Windows long-path precision (016-S)`, merge commit `75cfe3b1b9`
  - **closure PR #52** — merge commit `b7c8678e03`

  `review-thread` **REMAINS `N/A`** — the finding came from a local Go Reviewer
  pass, so no thread was ever created. Truthful terminal record, not a shortfall.

  Reconciliation was applied **in place** to the existing entry under Stage's own
  stash authority. No second entry was created; no Ship write was requested; the
  C5 capture-only carve-out and the single-write capture invariant are preserved
  unweakened.

- **`6B751D8B`** — no marker, no `N/A` source refs. No C5 obligation.

---

## 3. Problem frame — evidence gathered THIS cycle, not inherited

Every trigger below was re-measured directly at
`e7948d59998fd8f54f72b6cbf5f3c414a630aaad`. Prior-cycle dispositions were used
only to locate the triggers, never as evidence that they remain unfired.

### 3.1 `BF5DE670` — trigger RE-MEASURED, HAS NOT FIRED (6th consecutive cycle)

The entry's PRIMARY ask — a single consolidated risk register unifying GO-14, the
`Resolve`→use TOCTOU window, and SEC-5 — was **DISCHARGED** by `011.004-T` and
**verified still intact and accurate this cycle**: the register is present in
`internal/pathsafe/root.go`'s package doc with per-entry `STATUS:` and `TRIGGER:`
fields (GO-14 recorded as RESOLVED/FLIPPED by 013-S; the TOCTOU window and SEC-5
recorded as `accepted, oracle-parity` with the trigger *"forced the moment a real
write path"*).

The RESIDUAL — the **mitigation** (Lstat-before-write at call sites, `O_EXCL`
revalidation, hardlink-count checks) — is gated on *"before pathsafe gates any
live destructive file-write path."* Measured this cycle:

1. `scripts/check-write-path-precondition.sh` exits **0** (clean), and
   `--self-test` passes **5/5** fixtures (`accept-clean`,
   `accept-mentions-in-comment`, `reject-createtemp`, `reject-link`,
   `reject-writefile`) — so the gate is genuinely functional, not vacuously green.
2. An **independent** scan (not relying on the gate) across all **17** non-test
   `.go` files under `internal/**` and `cmd/**` for
   `os.WriteFile|Create|CreateTemp|OpenFile|Remove|RemoveAll|Rename|Mkdir|MkdirAll|Symlink|Link|Chmod|Truncate`,
   `io.Copy`, `sql.Open`, `bbolt.Open` returns **ZERO** matches. The single
   textual hit is a *comment* in `reparse_windows.go:27`, correctly excluded.
3. An independent scan for `.Resolve(` across all non-test `.go` files under
   `internal/**` and `cmd/**` returns **ZERO** production callers.
   `pathsafe.Root.Resolve` — the exact function the mitigation must attach to —
   **has no caller at all.**

**pathsafe gates NO destructive write path.** The trigger has not fired.

### 3.2 `6B751D8B` — items (1) and (2) LANDED; items (3) and (4) are the live residual

Verified in source this cycle:

- **Item (1)** (device-namespace branch reachability) landed as `017.002-T`. The
  `\\.\` guard is present in `addLongPathPrefix`, and the doc comment now records
  decision D-4 correctly — the branch IS reachable, but is *behaviourally
  redundant* given the following bare `\\` guard — with a named characterization
  lock (`TestAddLongPathPrefixDeviceNamespaceBranchReachability`) and an explicit
  "do NOT replace with a panic/unreachability assertion" instruction.
- **Item (2)** (UTF-16-aware MAX_PATH threshold) landed as `017.003-T`.
  `utf16Len` exists and the threshold comparison reads
  `if utf16Len(path) < longPathThreshold`.

**Live residual is items (3) and (4) only.**

### 3.3 `6B751D8B` item (3) — buffer pooling — trigger NOT fired

`getFinalPathNameByHandle` allocates `make([]uint16, syscall.MAX_PATH)` per call.
The entry's own trigger is *"consider pooling **if Resolve call volume grows
materially beyond** the documented 10-100-paths-per-root case."*

**CORRECTED during adversarial review (Scope Boundary Auditor, P2).** An earlier
draft of this deliberation asserted call volume is *zero* because `Root.Resolve`
has zero callers. That reasoning was **wrong**, and the correction is recorded
here rather than silently patched. `getFinalPathNameByHandle` is reached by a
second, independent path that does not involve `Resolve` at all:

```
config.Validate  (internal/config/validate.go:42, :67)
  -> pathsafe.NewRoot
       -> canonicalizeReparse            (internal/pathsafe/root.go:346)
            -> addLongPathPrefix
            -> getFinalPathNameByHandle
```

So the accurate measurement is:

- `Root.Resolve` specifically has **zero** production callers (§3.1.3) — this
  remains true and is what blocks `BF5DE670`'s mitigation, which must attach to a
  *write* call site.
- `getFinalPathNameByHandle` **is** reachable in the non-test call graph, at a
  volume of **one call per `NewRoot`** — i.e. one for `DefaultWorkspaceRoot` plus
  one per configured mapping, evaluated once at config-validation time.
- That is a handful of calls per process start, **far below** the documented
  10-100-paths-per-root case, and nowhere near *"materially beyond"* it.
- Additionally, no wired entrypoint reaches config validation today: both
  `cmd/intercom` and `cmd/intercom-ctl` return `errNotImplemented`.

The trigger has **not** fired — but on the correct grounds (volume far below the
documented threshold), not on the false grounds of zero reachability.

### 3.4 `6B751D8B` item (4) — CI lint scoping — trigger NOT fired, and the change is coverage-NEGATIVE

The entry's trigger is *"consider scoping it to windows-tagged packages only **if
CI time becomes a concern**."* Measured against real code-bearing CI runs:

| Run | Total wall clock |
|---|---|
| `34557397128` (016-S) | **155s** |
| `34556894501` | 150s |
| `34556094063` | 139s |

Per-job breakdown of run `34557397128`:

| Job | Duration |
|---|---|
| `test (windows, advisory)` | **104s** ← critical path |
| `lint` | **102s** |
| `test` | 101s |
| `security` | 52s |
| `ci gate` | 39s |

Two findings, both decisive:

1. **CI time is not a concern.** Total wall clock is ~2.5 minutes.
2. **`lint` is not even on the critical path.** `test (windows, advisory)` (104s)
   is longer than `lint` (102s) and runs in parallel. Shaving the second
   `golangci-lint run ./...` invocation would move total CI time by
   approximately **zero**. Most of the `lint` job is
   `go install golangci-lint@v2.13.2` (compiling the linter); the GOOS=windows
   re-run reuses that binary and a warm build cache, so its *marginal* cost is a
   fraction of the 102s.

Additionally, the proposed change is **coverage-negative**. The step exists
precisely because (per its own `016.001-T` comment) *"the repository's only lint
gate is GOOS-scoped via Go file discovery, so the `//go:build windows`-tagged
production surface under `internal/pathsafe` ... was never analyzed."* Narrowing
it to "windows-tagged packages only" would stop analyzing GOOS-conditional
behaviour in every other package under the Windows tag set — reducing static
analysis coverage over a **NON-NEGOTIABLE Constitution III control** to save
~0s of wall clock. This inverts the operator's own priority ordering
(*reliability/security supersede feature work*).

### 3.5 `475E76D2` — the identical optimization was already REJECTED BY NAME at plan review

The ask: add a byte-length fast-path guard before `utf16Len`'s rune scan so short
ASCII paths skip the O(n) decode.

The guard would be *sound* (UTF-8 byte length is always ≥ UTF-16 code-unit count,
so `len(path) < 260` implies `utf16Len(path) < 260`). Soundness is not the
question. The question is whether it should be built. Evidence against:

1. **`utf16Len` is already non-allocating.** It is a plain
   `for _, r := range path` accumulator — no `[]rune`, no `[]uint16`. The
   "expensive decode" the fast-path would skip is a bounded loop over a path
   string.
2. **The caller dwarfs it.** `canonicalizeReparse` does a lazy-DLL `Find()`, a
   `syscall.UTF16PtrFromString` **allocation**, a `syscall.CreateFile` syscall,
   and a `GetFinalPathNameByHandleW` syscall. A ≤260-iteration loop is noise
   against two Win32 syscalls.
3. **The code path is reachable, but carries no measurable production load.**
   **CORRECTED during adversarial review (Scope Boundary Auditor, P2).** An
   earlier draft claimed this code path "never executes in production" because
   `Root.Resolve` has no callers. That was **wrong**: `addLongPathPrefix` is
   reached via `config.Validate` → `pathsafe.NewRoot` → `canonicalizeReparse`
   (§3.3). The accurate statement is narrower — the function runs a handful of
   times per config validation, and no wired entrypoint reaches even that today
   (both CLI entrypoints return `errNotImplemented`). So the fast path would save
   a bounded loop, a handful of times, on a path that currently never runs at
   runtime.
4. **The same *category* of optimization was withdrawn at the 016-S plan review —
   though not this exact option.** **NARROWED during adversarial review (Scope
   Boundary Auditor, P2).** Plan option **O2-c** — "a bespoke non-allocating
   UTF-16 code-unit counter" — was **WITHDRAWN (Scope Boundary Auditor, P3
   gold-plating)** and recorded again as plan-review finding **R11**, with the
   reasoning: *"`addLongPathPrefix`'s only caller immediately calls
   `syscall.UTF16PtrFromString` (which allocates) and then two Win32 syscalls, so
   saving two small allocations buys nothing while adding bespoke infrastructure
   to a NON-NEGOTIABLE control."*
   It is **not** honest to call `475E76D2` "identical to O2-c" or "rejected by
   name": O2-c was a *bespoke counter*, withdrawn in favour of the plain range
   accumulator that actually shipped; `475E76D2` is a *preliminary byte-length
   early return* in front of that accumulator. They are different diffs.
   The precedent is therefore **directional, not dispositive** — it establishes
   the repo's standing disposition toward micro-optimizing this exact function on
   this NON-NEGOTIABLE control, but it does not by itself decide `475E76D2`.
   **The deferral rests primarily on the absence of measurable value (points 1-3),
   with the precedent as corroboration.**
5. **Institutional precedent.** `docs/compound/build-errors/go-race-requires-cgo-fails-closed-2026-09-03.md`:
   *"A guard built on a false premise is wasted implementation surface and can
   itself introduce review findings (P1/P2) for no safety benefit."* The repo
   convention (per `docs/decisions/2026-09-10-...-plan-review.md` §4) is that a
   discharged/withdrawn option **must not be re-litigated** without citing and
   rebutting the original rationale. **No new evidence exists to rebut it** — if
   anything the evidence is now *stronger* against, since the caller count is
   confirmed zero.

---

## 4. Options considered

### 4.1 Grouping options

- **G-1 — All three entries into one covering feature.** Rejected: `BF5DE670`'s
  deliverable cannot be built without inventing a write API (§5, D-1), so the
  group would be internally incoherent.
- **G-2 — `6B751D8B`(3,4) + `475E76D2` as a "pathsafe performance & CI" feature.**
  Considered seriously; see §4.2. Rejected on the merits, not on cohesion.
- **G-3 — Implement nothing; retain all three with re-measured dispositions.**
  **SELECTED.** See §5.

### 4.2 Options for the performance/CI remainder (`6B751D8B`(3,4) + `475E76D2`)

- **O-a — Implement all three optimizations now.** Rejected: all three triggers
  are measurably unfired; item (4) is coverage-negative; `475E76D2` re-opens a
  by-name-rejected option. Adds complexity to a NON-NEGOTIABLE control for zero
  measured benefit.
- **O-b — Implement only `475E76D2`** (the operator flagged it as "include only
  if the plan proves it remains simple and cohesive"). Rejected: the plan does
  **not** prove that. It is simple in *diff size* but it is a direct re-opening
  of withdrawn option O2-c with no rebutting evidence.
- **O-c — Build a new CI gate that fires when `Root.Resolve` gains its first
  production caller**, mechanically arming items (3), `475E76D2`, and
  `BF5DE670`'s mitigation the way `011.004-T` armed the write-path trigger.
  Rejected as **scope expansion**: no chartered entry asks for it, freeze-scope
  is active, and the precedent is explicit — "inventing a write API to hang the
  mitigation on" was rejected as scope expansion in the prior cycle for the same
  reason. Recorded as a candidate for a future *chartered* cycle.
- **O-d — Retain all, implement none, record re-measured evidence.** **SELECTED.**

---

## 5. Decisions

- **D-1 — `BF5DE670`: RETAINED (deferred). Contributes NO task. NOT archived.**
  Trigger re-measured directly this cycle and has not fired (§3.1). The
  deliverable is a mitigation that must attach to a filesystem write call site;
  none exists, and `Root.Resolve` has no caller at all. Building it would require
  either inventing a write API — which trips `017-F`'s AC-F3 (adding a write
  primitive is a **named stop condition**) and is **scope expansion**, a declared
  stop condition for this cycle — or building machinery against an API that does
  not exist, i.e. exactly the speculative engineering that this entry's five
  prior dispositions rejected on the same measured grounds.
  **NOT archived** because it remains the ONLY tracking item for a deferred
  security mitigation whose underlying risk acceptance is still live and
  unmitigated. Its primary ask (the consolidated register) remains DISCHARGED and
  was verified intact this cycle.
  **Unchanged remaining trigger:** archive when the mitigation lands, OR when
  `scripts/check-write-path-precondition.sh` FIRES and the forced mitigation is
  designed and landed.

- **D-2 — `6B751D8B`: RETAINED (deferred). Contributes NO task. NOT archived.**
  Items (1) and (2) are landed and verified in source (§3.2). Items (3) and (4)
  are the live residual; both triggers measurably unfired (§3.3, §3.4), and item
  (4) is additionally **coverage-negative** on a NON-NEGOTIABLE control for ~0s
  of wall-clock gain. This entry remains the only tracker for items (3) and (4).
  Archive when those two are implemented or explicitly retired.

- **D-3 — `475E76D2`: ACCEPT-AS-IS (do nothing), matching plan precedent.
  RETAINED. Contributes NO task. NOT archived.**
  The entry itself offered Stage the choice of "accept-as-is (do nothing,
  matching plan precedent) or re-open with fresh justification."
  The guard is **mathematically sound** (§3.5; independently re-proved by the
  Correctness Reviewer across ASCII, 2-, 3-, and 4-byte sequences and invalid
  UTF-8) and mechanically trivial. It is deferred **not because it is risky but
  because it is not worth a shipment**: the saved work is a bounded loop over a
  ≤260-unit path, dominated by the `syscall.UTF16PtrFromString` allocation and
  two Win32 syscalls that immediately follow it, on a path that no wired
  entrypoint currently reaches. The withdrawn option O2-c is **corroborating,
  directional precedent — not a by-name rejection of this diff** (narrowed at
  adversarial review, §3.5 point 4).
  Retained rather than archived because the operator did not authorize retiring
  it, and freeze-scope plus "destructive operations not authorized" means Stage
  records the disposition rather than closing the entry unilaterally.

- **D-4 — NO SHIPMENT IS CREATED THIS CYCLE.**
  The safe cohesive remainder is **EMPTY**. Every chartered entry's actionable
  content is either blocked behind a stop condition (D-1) or a
  measurably-premature optimization whose trigger has not fired (D-2, D-3).
  Manufacturing a shipment from D-2/D-3 content would mean shipping changes that
  a prior plan review already rejected by name, adding complexity to a
  NON-NEGOTIABLE security control for zero measured benefit — a direct inversion
  of the operator's stated priority ordering (*reliability/security supersede
  feature work; simplicity supersedes complexity*) and of YAGNI.
  Per the operator's own contract — *"Halt only if no safe shipment can be
  formed"* — this is the halt condition, reached on measured evidence rather than
  on inability to proceed.

- **D-5 — P-021 C5 obligations discharged and recorded** for all outcomes:
  duplicate scan CLEAN (recorded explicitly, §2A); `BF5DE670` reconciliation a
  recorded idempotent no-op; `475E76D2` reconciliation **successful** — PR
  `N/A` → **#51** (merge `75cfe3b1b9`), closure **#52** (merge `b7c8678e03`),
  recovered from Ship-owned merged-PR history joined on shipment `016-S`;
  review-thread remains `N/A` as a truthful terminal record (§2B).

---

## 6. Scope boundary (anti-goals)

- Do **NOT** add any filesystem write primitive under `internal/**` or `cmd/**`
  (named stop condition; would trip `check-write-path-precondition.sh`).
- Do **NOT** invent a caller for `Root.Resolve` to unblock `BF5DE670`,
  `37FAB8C2`, or item (3).
- Do **NOT** pull `37FAB8C2` into scope — explicitly excluded by the operator,
  and its own text defers to a live caller.
- Do **NOT** modify the consolidated risk register in `root.go`.
- Do **NOT** narrow the GOOS=windows lint step.
- Do **NOT** archive any of the three chartered entries.
- Do **NOT** touch any of the other 10 active stash entries.

---

## 7. Open questions

None blocking; zero P0/P1 at the adversarial decision-review gate
(`docs/decisions/2026-09-11-intercom-go-pathsafe-deferred-trigger-reassessment-decision-review.md`,
verdict **PASS**).

Three candidates for a **future chartered cycle**, surfaced by that review and
deliberately **not** actioned here because no chartered entry requests them and
freeze-scope is active:

1. **(Highest value — security, P2/R3) Harden
   `scripts/check-write-path-precondition.sh` against selector-text bypass.**
   The gate matches local identifier text against a fixed selector list, so a
   future write path could land while the gate stays green via import aliasing
   (`import fs "os"; fs.WriteFile`), `syscall.CreateFile`/`syscall.Write` with
   write access, Go 1.24 `os.Root` methods (`root.Create/OpenFile/Mkdir`), or an
   unlisted/aliased embedded-DB driver. This does **not** make the current
   deferral unsafe — that rests independently on the measured zero-`Resolve`-caller
   signal — but the gate must not be presented as a complete tripwire for a
   NON-NEGOTIABLE control.
2. **(R5/O-c) A mechanical tripwire firing when `Root.Resolve` gains its first
   production caller**, analogous to the write-path gate. Library-agnostic, and
   it would retire the hand re-measurement Stage has now repeated for six
   consecutive cycles. Both the Architecture and Security lenses converged on
   this independently.
3. **(R6) Generalize the SEC-5 register wording** from "EvalSymlinks does not
   resolve hardlinks" to cover `GetFinalPathNameByHandleW` on Windows — to be
   done by the next cycle that legitimately edits the register, since touching it
   is an anti-goal here (§6).

---

## 8. Dispositions

| Stash ID | Priority | Decision | Archived? | Task contributed |
|---|---|---|---|---|
| `BF5DE670` | medium | **RETAINED (deferred)** — trigger re-measured, has not fired; mitigation blocked behind a named stop condition | **No** | None |
| `6B751D8B` | low | **RETAINED (deferred)** — items (1)/(2) landed; items (3)/(4) triggers unfired, (4) coverage-negative | **No** | None |
| `475E76D2` | low | **ACCEPT-AS-IS / RETAINED** — re-opens by-name-withdrawn option O2-c with no rebutting evidence; C5 PR ref reconciled to #51/#52 | **No** | None |
| `37FAB8C2` | medium | **OUT OF CHARTER** — explicitly excluded by operator; not triaged, not edited | **No** | None |
| all other 10 | — | **OUT OF CHARTER** — scanned for duplicates only (read-only); not triaged, not edited | **No** | None |

**Shipment outcome:** none created. Safe cohesive remainder is empty (D-4).

# Decision Review — Pathsafe Deferred-Trigger Reassessment (adversarial, multi-persona)

- **Date:** 2026-09-11
- **Subject:** `docs/decisions/2026-09-11-intercom-go-pathsafe-deferred-trigger-reassessment-deliberation.md`
- **Gate type:** decision review (adversarial, multi-persona, multi-model)
- **Attempt:** 1
- **VERDICT: PASS** — zero P0/P1 findings across all four lenses; all four
  reviewers independently concurred that decision **D-4 (create no shipment)** is
  correct.

---

## 1. Why a decision review and not a plan review

Decision **D-4** is that **no implementation work is authorized this cycle**.
With an empty implementation scope there is no implementation plan to generate,
no hardening signals to harden, and nothing for a plan-review gate to act on — an
`impl-plan` artifact here would be vacuous by construction.

The operator's requirement was *"adversarial review of decisions before any PR"*
using *"the strongest applicable multi-persona decision/plan review route,
including architecture, correctness, scope, security, and maintainability
lenses."* That review route was therefore applied to the **decision artifact**,
which is the actual deliverable of this cycle. All five mandated lenses were
exercised.

`impl-plan` / `plan-harden` are recorded **N/A — empty implementation scope**,
not skipped. No `skip_plan` / `skip_review` / `force_harvest_no_gates` flag was
set or relied upon, so the Step 3.0 gate-bypass guard was never engaged.

## 2. Reviewer panel (independent, model-diverse)

| Lens | Persona | Model | Verdict | P0/P1 |
|---|---|---|---|---|
| Scope / YAGNI | Scope Boundary Auditor | `gpt-5.6-sol` | ADVISORY (D-4 correct) | 0 |
| Correctness / evidence | Correctness Reviewer | `gemini-3.8-flash` | **PASS** | 0 |
| Security | Security Lens Reviewer | `claude-opus-4.8` | ADVISORY (deferral safe) | 0 |
| Architecture + Maintainability | Architecture Strategist | `grok-4.6` | ADVISORY (no shipment correct) | 0 |

Model diversity was deliberate: the zero-shipment conclusion contradicts the
operator's stated expectation of one shipment, so it was stress-tested by four
different model families rather than confirmed by one.

The Scope reviewer was explicitly tasked to attack from **both** directions —
under-delivery/scope-avoidance *and* scope creep — to prevent a one-sided review
that only checks for over-reach.

## 3. Findings and dispositions

### R1 — P2, Scope — "identical to O2-c / rejected by name" was overstated
`475E76D2` (a preliminary byte-length early return) is **not** the same diff as
withdrawn plan option O2-c (a bespoke non-allocating UTF-16 counter). O2-c was
withdrawn in favour of the plain range accumulator that shipped as `utf16Len`.
The prior rejection is directional precedent, not a by-name rejection of this
change, and cannot by itself justify deferral.
**REMEDIATED.** Deliberation §3.5 point 4 narrowed; D-3 rewritten to rest
primarily on absence of measurable value, with precedent as corroboration only.

### R2 — P2, Scope — "never executes in production" was factually wrong
`addLongPathPrefix` and `getFinalPathNameByHandle` are reachable in the non-test
call graph via `config.Validate` → `pathsafe.NewRoot` → `canonicalizeReparse`
(`validate.go:42`, `:67`; `root.go:346`), independent of `Root.Resolve`.
**REMEDIATED.** Verified independently by Stage before accepting the finding.
Deliberation §3.3 rewritten and §3.5 point 3 corrected; both corrections are
recorded in-place as visible corrections rather than silently patched. The
trigger conclusion is unchanged but now rests on correct grounds: call volume is
a handful per config validation, far below the documented 10-100-paths-per-root
threshold, not zero.

### R3 — P2, Security — the write-path gate has real bypass vectors **(DEFERRED — needs an operator charter)**
`scripts/check-write-path-precondition.sh` matches local identifier text against
a fixed selector list. It is genuinely functional (5/5 self-test, comment/string
masking) and **not** vacuously green, but a future write path could land while
the gate stays green via:
1. **import aliasing** — `import fs "os"; fs.WriteFile(...)` emits no `os.WriteFile` token;
2. **`syscall.CreateFile`/`syscall.Write` with write access** — not in the selector list, and the package already uses `syscall.CreateFile`;
3. **Go 1.24 `os.Root` methods** — `root.Create/OpenFile/Mkdir/Remove` carry no `os.` qualifier (the module floor is Go 1.24);
4. **unlisted or aliased embedded-DB drivers** — only `sql.Open`/`bbolt.Open` are covered, while a DB layer is clearly future work (`Database.Path`, `KindDB`).

**Assessed NOT blocking for this cycle**, for a reason that is load-bearing: the
`BF5DE670` deferral does **not** rest on the gate alone. It rests independently
on the directly-measured **zero `Root.Resolve` callers** signal, which is
library-agnostic and which three reviewers confirmed separately. A write that
evades both `pathsafe` and the gate is an *ungated-write* problem, not a
TOCTOU/hardlink-*mitigation* problem.

**Not actioned this cycle: hardening the gate is not requested by any chartered
entry, and freeze-scope is active — acting on it would be scope expansion, a
declared stop condition.** Surfaced to the operator as the top candidate for the
next chartered cycle. Explicitly **not dropped**: recorded here, in the
deliberation's §7 open questions, and in the session report.

### R4 — P2, Architecture — a G-2 shipment would have violated width isolation
Combining `6B751D8B` items (3)+(4) with `475E76D2` mixes Go production code with
a CI/YAML change in one feature — a single-skill-domain violation — and item (4)
is coverage-negative on a Constitution III control. Confirms G-2's rejection on
independent grounds (cohesion) beyond the merits-based rejection already recorded.
**ACCEPTED, no change required** — strengthens D-4.

### R5 — P2, Maintainability — the re-measurement treadmill is accumulating
Six consecutive cycles of hand re-measurement guarantee a seventh identical
P-021 C6 deliberation. Option **O-c** (a mechanical tripwire on the first
production `Root.Resolve` caller) is the right long-term answer, and both the
Architecture and Security lenses independently converged on it.
**ACCEPTED.** Correctly deferred as scope expansion this cycle (inventing a gate
to unblock work is the same class of move as inventing a write API, which prior
precedent rejected). Surfaced with R3 as a chartered-cycle candidate.

### R6 — P3, Security — SEC-5 register wording is POSIX-centric
The register says "EvalSymlinks does not resolve hardlinks," but on Windows the
canonicalization path is now `GetFinalPathNameByHandleW`, which equally fails to
resolve hardlinks. The risk is accurate on both platforms; only the citation is
narrow. **ACCEPTED, not actioned** — touching the register is an explicit
anti-goal this cycle (§6). Recorded for the next cycle that legitimately edits it.

### R7 — P3, Architecture — `pathsafe`'s zero-caller state is *pre-wiring*, not dead code
Recorded as an observation to pre-empt a future agent proposing retirement or
further optimization of the package. **ACCEPTED, observation only.**

### R8 — P3, Maintainability — this deliberation is long for a halt
**ACCEPTED.** Retained as-is: the length is P-021 C5/C6 compliance surface
(unconditional duplicate scan, reconciliation outcomes for all four cases), which
is mandatory to record. Noted for future cycles to reference this document's
measurements rather than cloning §1-2.

## 4. Verified preconditions — recorded so they are not re-litigated

Independently confirmed by at least one reviewer by direct measurement:

- **17** non-test `.go` files under `internal/**` and `cmd/**`; **zero**
  filesystem write primitives (sole textual hit is a comment at
  `reparse_windows.go:27`).
- `scripts/check-write-path-precondition.sh` exits 0; `--self-test` passes 5/5.
- `pathsafe.Root.Resolve` has **zero** production callers — direct, indirect,
  via interface, or via method value.
- `6B751D8B` items (1) and (2) are landed (`017.002-T`, `017.003-T`).
- `utf16Len` is allocation-free (plain range loop, no `[]rune` / `[]uint16`).
- **`len(path) < 260` as a fast-path guard is SOUND** — `len(path) >=
  utf16Len(path)` holds for ASCII, 2-, 3-, and 4-byte sequences and for invalid
  UTF-8 (replacement rune). No false negatives are possible. *The guard was
  deferred on value, never on safety.*
- `syscall.CreateFile` in `reparse_windows.go` is metadata-only: desired access
  `0`, `OPEN_EXISTING`. The `READ|WRITE|DELETE` flags are a **share** mode, not
  an access mode.
- `lint` (102s) is structurally **not** on the CI critical path — `lint`,
  `expensive`/`test`, and `windows` all share `needs: changes` with no
  inter-dependency and run concurrently; `test (windows, advisory)` is 104s.

## 5. Cycle accounting

One review round. Zero P0/P1 findings, so no re-entry cycle was consumed
(bounded limit: 2). Two P2 findings (R1, R2) were remediated in-place before this
verdict was recorded; R3-R8 are accepted or deferred with explicit rationale.

<!-- decision-review-attempt: 1 -->

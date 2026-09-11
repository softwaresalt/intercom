---
title: "Plan Review — Pathsafe Error-Surface Parity and Windows Long-Path Precision"
date: 2026-09-10
status: PASS
verdict: PASS
agent: Stage
mode: DARK_MODE_ACTIVE
governs: 017-F / shipment 016-S
plan: docs/plans/2026-09-10-intercom-go-pathsafe-error-surface-parity-plan.md
---

# Plan Review — Pathsafe Error-Surface Parity and Windows Long-Path Precision

**Plan under review**:
`docs/plans/2026-09-10-intercom-go-pathsafe-error-surface-parity-plan.md`
**Attempt**: 1 (no re-entry cycles required)
**Verdict**: **PASS** (after in-cycle remediation of all P1 and P2 findings)

---

## 1. Method — multi-persona adversarial review

Five independent reviewers were run **in parallel**, each on a **different
model** to maximize independence and reduce shared blind spots, each given the
plan, the deliberation, and direct access to the source under change.

| Persona | Model | Lens |
|---|---|---|
| Go Reviewer | `gpt-5.6-sol` | Go safety, idiom, build tags, error semantics |
| Correctness Reviewer | `claude-opus-4.8` | Premise falsification, edge cases, off-by-one |
| Scope Boundary Auditor | `gemini-3.8-flash` | Scope creep, YAGNI, AC falsifiability |
| Constitution Reviewer | `claude-sonnet-5` | Per-principle constitutional mapping |
| Security Lens Reviewer | `grok-4.6` | Threat model, containment, data exposure |

## 2. Raw outcome

**2 × P1, 6 × P2, 7 × P3. Zero P0.**

Two findings were raised **independently by two reviewers each**, which is the
strongest available confidence signal:

- The POSIX `Op`/`Path` assertion defect (Go Reviewer P1 **and** Correctness
  Reviewer P2).
- The Win32-normalization overclaim (Security Lens P2 **and** Correctness
  Reviewer P3).

## 3. Independent verification performed by Stage

Stage did not accept reviewer claims on assertion. Two material claims were
re-verified directly against the repository before remediation:

| Claim | Verification | Result |
|---|---|---|
| The Windows `go test` CI job is advisory-only | Read `.github/workflows/ci.yml` L251–L252 | **CONFIRMED** — `continue-on-error: ${{ vars.WINDOWS_GATE_REQUIRED != 'true' }}` |
| `requireSymlinkOrFailClosed` can skip on Windows | Read the helper in `internal/pathsafe/symlink_test.go` | **CONFIRMED** — `t.Skipf` on `ERROR_PRIVILEGE_NOT_HELD` |

Both confirmed the reviewers were correct, and both changed the plan materially.

## 4. Findings and dispositions

### P1 — both REMEDIATED

| ID | Finding | Remediation |
|---|---|---|
| **R1** | `017.001-T` AC3 required `Op == "open"` / `Path == ancestor`, but `wrapPathError` is **inert on POSIX** — it returns `EvalSymlinks`' own `*fs.PathError` whose `Op` is `"lstat"`/`"EvalSymlinks"` and whose `Path` is a component, not `ancestor`. Placed in the cross-platform AC1 test, this would **fail on POSIX** or force a POSIX error-surface change, contradicting H4. | AC3 split and scoped **Windows-only** with an explicit prohibition on placing it in the cross-platform test; POSIX inertness documented as the reason. |
| **R2** | No named safety mode declared (Principle VIII) despite a NON-NEGOTIABLE Constitution III control in an unattended `DARK_MODE_ACTIVE` cycle. The plan *performed* freeze-scope and investigate-first work without *declaring* either, so the gate could not be confirmed. | New **H0** declares **`freeze-scope`** and **`investigate-first`** by name, with bounds and rationale; `careful` explicitly not declared, with reason. Recorded as **D-11**. |

### P2 — all six REMEDIATED

| ID | Finding | Remediation |
|---|---|---|
| **R3** | A symlink-based Windows test can **skip** without symlink privilege, leaving the Windows `Errno → *fs.PathError` path unexercised **while the suite reports green**. | New **AC3b** requires a privilege-independent **dangling-junction** fixture in `junction_windows_test.go`. |
| **R4** | H1's *"verdict-neutral by construction"* **overclaimed**. `\\?\` also disables Win32 normalization; dropping it in the precision band re-enables trailing-dot/space stripping and reserved device names (`CON`/`NUL`/`AUX`). | **D-10** added: side effect analyzed and **explicitly accepted** (those paths now behave like every sub-MAX_PATH path already does). H1 corrected, AC5 rescoped, and H2 explicitly barred from being cited as covering normalization. |
| **R5** | Missing the Governance-mandated **Constitution Check** section. | New **§9a** maps all principles with per-principle verdicts. |
| **R6** | H3 cited "CI runs a GOOS=windows lint step" as mitigation, omitting that the Windows **runtime test** job is advisory and the blocking lint step is compile-only — while H2's top risk is a Windows-only **runtime** regression. | H3 rewritten with measured line references and three required mitigations, including mandatory pasted Windows test output in the PR body when `WINDOWS_GATE_REQUIRED != 'true'`. |
| **R7** | `017.002-T`'s characterization-first posture deviates from Principle II's letter without invoking the Governance conflict-resolution clause. | Deviation, justification, and rejected simpler alternative documented inline at the task. |
| **R14** | `PathError.Path` becomes a new Windows information channel carrying an absolute ancestor path. | New **AC3d** documents that it is operator/diagnostic-only and MUST NOT be interpolated into session-facing agent payloads. Display string verified unchanged (`apperr.Error.Error()` never appends cause). |

### P3 — all seven REMEDIATED

| ID | Finding | Remediation |
|---|---|---|
| **R8** | 002 → 003 "hard dependency" rationale overstated. | Restated as **merge coordination**; D-9 amended. |
| **R9** | Sibling `os.Lstat` error branch left unchanged without explanation. | **AC3c** requires an inline note (it already returns `*fs.PathError`). |
| **R10** | `017.002-T` AC2 partly duplicated existing `TestAddLongPathPrefixVerdicts`. | AC2 rewritten to specify only the **delta** (below/at-threshold cases). |
| **R11** | Bespoke "non-allocating UTF-16 counter" is gold-plating. | **O2-c withdrawn**; use `utf16.Encode` or a plain range loop. |
| **R12** | `017.002-T` AC1 mislabelled "(test-first)" — it tests stdlib and passes at pre-change HEAD. | Relabelled; its real value (guarding against a future `filepath.IsAbs` change) stated. |
| **R13** | AC3/AC5 of `017.002-T` were unfalsifiable restatements of intent. | Both rewritten in falsifiable form (guard-reorder check; `git diff` shows no executable-statement change). |
| — | Verdict-neutrality contingent on the unstated `Abs`/`Clean` precondition. | Precondition now stated explicitly in H1. |

## 5. Claims independently verified as TRUE by reviewers

The Correctness Reviewer rigorously confirmed, and Stage had measured
independently:

1. UTF-8 bytes ≥ UTF-16 code units across **all** Unicode ranges (1/1, 2/1,
   3/1, 4/2) — the existing proxy can only over-trigger, never under-trigger.
2. `filepath.IsAbs` is **true** for `\\.\` device paths — the stash entry's
   "unreachable" premise is **false**; the branch is reachable.
3. Guard order is `\\?\` → `\\.\` → `\\`, so the `\\.\` branch is reachable but
   **behaviourally redundant**.
4. `apperr.Wrapf` retains `cause`; `*apperr.Error` implements `Unwrap()`;
   `errors.As` traverses into an injected `*fs.PathError`.
5. `wrapPathError` is **inert on POSIX**.
6. **No off-by-one**: `utf16len < longPathThreshold` preserves the existing
   strict-less-than boundary against the same constant; ASCII verdicts are
   byte-identical.

## 6. Security verdict

**No P0/P1 security findings.** The Security Lens Reviewer independently
re-counted the 17 production `.go` files and confirmed **zero** destructive
write primitives, that `syscall.CreateFile` is `OPEN_EXISTING` with desired
access `0`, and that `Root.Resolve` has **no production caller** (both binaries
still return `not implemented`). Verdict on `BF5DE670`'s deferral: **correct —
the TOCTOU/hardlink mitigations have nothing to attach to, and are not
exploitable now.**

## 7. Scope verdict

**PASS — strictly compliant.** Zero P0/P1 from the Scope Boundary Auditor. No
excluded stash ID was triaged, edited, harvested, archived, or included;
references to excluded IDs appear **only** inside the mandatory read-only
P-021 C5(A) duplicate-detection scan. `BF5DE670`'s reduction to feature-level
acceptance criteria was explicitly adjudicated **acceptable preservation, not
under-delivery**, and deferring `6B751D8B` items (3)(4) **correct**. All three
tasks confirmed within the 2-hour granularity limit.

## 8. Gate outcome

**VERDICT: PASS.**

Zero P0 at any point. All 2 P1 and all 6 P2 findings were remediated **in
cycle** (plan Rev 3); all 7 P3 findings were also remediated rather than
merely acknowledged. No re-entry cycle was consumed — the plan-review attempt
counter remains at **1**.

The plan carries a `## Plan Hardening` section (§9, H0–H7) plus a
Constitution Check (§9a), satisfying the P-006 hardening gate for a plan whose
`Requires plan hardening` field is `yes`.

**Cleared to harvest.**

<!-- plan-review-attempt: 1 -->

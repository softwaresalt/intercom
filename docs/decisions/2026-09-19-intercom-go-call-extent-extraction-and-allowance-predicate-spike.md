---
title: "Call-extent extraction and access-mode allowance predicate (U4-T1b): spike findings"
description: "Prototype-backed design decision for argument-aware call-extent extraction and the access-mode allowance predicate the write-path precondition gate needs to satisfy AC-4.2/AC-4.3 (shipment 031-S, feature 034-F)"
status: "complete"
source_document: "docs/plans/2026-09-18-intercom-go-gate-reliability-plan.md"
linked_artifacts:
  - "docs/decisions/2026-09-18-intercom-go-ship-contract-and-gate-reliability-deliberation.md"
  - "scripts/check-write-path-precondition.sh"
  - "internal/pathsafe/reparse_windows.go"
tags:
  - "spike"
  - "write-path-gate"
  - "034-F"
---

# Call-extent extraction and access-mode allowance predicate (U4-T1b) — Spike findings

- **Date**: 2026-09-19
- **Agent**: Stage
- **Scope**: `034-F` — Harden write-path precondition gate and add Resolve tripwire (S-7)
- **Session branch**: `chore/stage-ship-pipeline-contract-repair`
- **Governing plan**: `docs/plans/2026-09-18-intercom-go-gate-reliability-plan.md` (U4-T1b)
- **Supersedes task**: `034.001-T` (was a Ship-executed spike; the spike is executed here instead)
- **Trigger**: PR #66 Copilot cycle-3 finding on `.backlogit/queue/034.001-T.md:18`
  (thread `PRRT_kwDOTPuhps6kCdf5`) — the queued task delegated **planning authority** to Ship
  (choose the design, re-size follow-on tasks, possibly narrow AC-4.2). Ship may not update task
  planning fields or create/modify spike and plan artifacts (`_ship.agent.md:38,43`), so `031-S`
  would have halted on this item.
- **Status**: DECIDED — call-extent extraction is **FEASIBLE**; AC-4.2 is **NOT narrowed**

---

## 1. Question

`AC-4.2` requires that the existing metadata-only `syscall.CreateFile` in
`internal/pathsafe/reparse_windows.go` not trip the write-path gate, with the allowance keyed on
the **call's access mode**, not on the file. Plan review found this unsatisfiable under the current
scanner and sized the naive fix `L`/`high`, which the 2-hour rule forbids shipping directly.

Two questions had to be answered by prototype, not by argument:

1. Can an argument-aware **call extent** be recovered from `scan_file`'s masked text without a Go
   parser — i.e. is a grep-shaped gate able to evaluate a multi-line call as one unit?
2. If so, what is the **allowance predicate shape**, and does it survive `AC-4.3`'s positive
   control (a *writing* `CreateFile` in the *same* file must still be rejected)?

## 2. Measured starting state (bound snapshot `14d44e3c`)

`scripts/check-write-path-precondition.sh`:

* `scan_file` (`:186–193`) masks the file, then iterates `masked.splitlines()` and matches each of
  the 20 `SELECTOR_RES` patterns **per line**. Findings are reported as `path:line:`.
* `mask_go_non_code` (`:92–183`) blanks comment/string/rune contents **while preserving total
  length and line structure**. This property is load-bearing for the design below.
* No allowance mechanism exists anywhere in the script.

`internal/pathsafe/reparse_windows.go`:

* `syscall.CreateFile` appears **five** times: four inside comments (`:24`, `:28`, `:33`, `:145`)
  and **one real call site at `:164`**, spanning `:164–172` with the desired-access literal `0` on
  `:166` — the line *after* the selector token. Confirmed by prototype against the real file.

## 3. Prototype and result

Prototyped against the **real** `mask_go_non_code` (loaded verbatim out of the shell script) and
the **real** `reparse_windows.go`. Run in a scratch directory; no repository source was modified.

Extraction over the whole masked text (not per line) yielded exactly one code-level hit — the four
comment occurrences were correctly masked away — and decomposed it as:

```
line 164: CALL spanning lines 164..172, argc=7 (+1 trailing-comma artifact)
  arg[0] = 'pathPtr'
  arg[1] = '0'                     <-- dwDesiredAccess
  arg[2] = 'syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE'
  arg[3] = 'nil'
  arg[4] = 'syscall.OPEN_EXISTING'
  arg[5] = 'syscall.FILE_FLAG_BACKUP_SEMANTICS'
  arg[6] = '0'
```

Predicate behaviour across the adversarial case set:

| Case | Result |
|---|---|
| Live metadata-only call (multi-line) | **ALLOW** |
| **Positive control** — writing `GENERIC_WRITE` call in the **same file** as an allowed one | **REJECT** (and the allowed one still ALLOWs) |
| Access arg wrapped in nested parens: `uint32(syscall.GENERIC_WRITE\|syscall.GENERIC_READ)` | **REJECT** |
| Selector appearing in a comment only | **CLEAN** (no hit) |
| Selector as a non-call reference: `var fn = syscall.CreateFile` | **REJECT** (fail-closed) |
| Unbalanced / truncated call extent | **REJECT** (fail-closed) |

`AC-4.3` therefore passes: the allowance is genuinely keyed on the call's access mode, and a
file-scoped exemption is *not* what is being built.

## 4. Decision

**D-1 — Extraction approach: offset-based, paren-depth call-extent extraction over the masked
text.** `scan_file` stops iterating `splitlines()` and instead matches `SELECTOR_RES` against the
**whole masked string**. Because `mask_go_non_code` preserves length and line structure, the
existing `path:line:` reporting format is retained by computing
`masked.count("\n", 0, match.start()) + 1`. For each hit: skip whitespace after the selector;
require `(`; walk forward maintaining paren depth to the balanced `)`; split the interior at
**depth-1** commas to obtain the argument list. No Go parser, no new dependency.

**D-2 — Allowance predicate shape: an access-mode literal test, applied only to
`syscall.CreateFile`, fail-closed everywhere else.** The call is allowed **only** when all hold:

1. the selector is exactly `syscall.CreateFile`;
2. the extracted extent is a real call with a **balanced** argument list;
3. after discarding a trailing-comma artifact, the list has **exactly 7** arguments (the Win32
   `CreateFileW` arity) — any other arity is undecidable and rejected;
4. `arg[1]` (`dwDesiredAccess`) is the **literal `0`**.

Anything else — a non-call reference, an unbalanced extent, a wrong arity, or any non-`0` access
expression including a parenthesised or symbolic one — is **rejected**. The predicate never widens
the gate; its only failure direction is a false positive.

**D-3 — `AC-4.2` is NOT narrowed.** The spike's conditional branch ("if call-extent extraction
proves infeasible, AC-4.2 is narrowed to a line-anchored allowance") **does not fire**. `AC-4.2`
stands verbatim.

**D-4 — The other 19 selectors keep their current semantics.** They remain presence-matched, now by
offset instead of per line; the reported line number is unchanged. This keeps the change bounded
and keeps every existing `--self-test` fixture passing unmodified.

## 5. Residual (folded into U4-T4 / `034.007-T`, AC-4.7)

**Undecidable call extents are rejected, not allowed.** A `syscall.CreateFile` reached through a
wrapper, through a function value, with a non-literal access variable, or with an argument list the
extractor cannot balance is **flagged**, not exempted. This is deliberately conservative: the
residual is a *false-positive* surface (a future legitimate metadata-only call written in one of
those shapes will trip the gate and need an explicit, reviewed widening), never a silent hole. It
must be documented alongside the existing `os.Root`-method / dot-import / shadowing residuals
rather than left unstated.

## 6. Re-sizing of the follow-on tasks (`AC` of the retired spike)

The naive `L`/`high` task is replaced by two `S` tasks — the spike's whole purpose. Both are
comfortably inside the 2-hour rule: one file, one function each, bounded prototype-proven diffs.

| Task | Was | Now | Rationale |
|---|---|---|---|
| `034.002-T` (U4-T1b-1) call-extent extraction | `M` / medium | **`S`** / medium | Prototype is ~25 lines in one function (`scan_file`) in one file; line mapping is a one-liner; no new dependency |
| `034.003-T` (U4-T1b-2) allowance predicate | `M` / medium | **`S`** / medium | Predicate is ~10 lines with four explicit conditions, all enumerated in D-2; one file |

`034.004-T`, `034.005-T`, `034.006-T`, `034.007-T`, `034.008-T` are unaffected and keep their
existing sizes; none exceeds `M`. The retired spike's acceptance criterion "each follow-on
implementation task re-sized to at most M" is therefore satisfied.

## 7. What Ship must NOT do

Ship applies D-1 and D-2 as **specified implementation**. It exercises **no design discretion**: it
does not choose an alternative extraction approach, does not re-size any task, and does not narrow
`AC-4.2`. If the implementation contradicts this decision, Ship **HALTs and returns to Stage**
rather than re-deciding — re-deciding here is a P-010 role-boundary violation.

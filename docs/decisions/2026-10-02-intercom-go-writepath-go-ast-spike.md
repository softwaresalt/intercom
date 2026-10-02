---
title: "Write-path gate go/ast migration spike: findings (049.001-T, E-T1)"
description: "G-1..G-7 empirical findings for migrating the write-path gate from masked-text detection to go/parser + go/ast, measured against the full Unit D corpus on main at 961b652"
status: "complete"
source_document: "docs/plans/2026-10-01-intercom-go-post-m4-go-replan-and-containment-hardening-plan.md"
linked_artifacts:
  - "docs/decisions/2026-10-01-intercom-go-post-m4-stash-triage-and-go-replan-deliberation.md"
  - "tools/gatecheck/internal/writepath/writepath.go"
  - "tools/gatecheck/internal/writepath/writepath_oracle_test.go"
  - "tools/gatecheck/internal/retiredarch/writepath_mask_test.go"
  - "tools/gatecheck/internal/writepath/testdata/writepath_golden.json"
tags:
  - "spike"
  - "write-path-gate"
  - "go-ast"
  - "unit-e"
  - "049-F"
  - "findings"
---

# Findings: write-path gate `go/ast` migration spike

* **Governing plan:**
  `docs/plans/2026-10-01-intercom-go-post-m4-go-replan-and-containment-hardening-plan.md`,
  Unit E, task E-T1 (049.001-T).
* **Governing deliberation:**
  `docs/decisions/2026-10-01-intercom-go-post-m4-stash-triage-and-go-replan-deliberation.md`
  §4.7 (operator decision D-049-1).
* **Trigger:** stash entry 1EEBECA5. Shipment 031-S (feature 034-F) merged to
  `main` in PR #97, with closure in PR #98. 034.009-T, the last task of the
  Unit D corpus chain, is therefore on `main`.
* **Measured at:** `main` @ `961b652`, `go1.26.5 windows/amd64`.
* **Executed by:** Stage, 2026-10-02, under the P-016 explicit, time-boxed
  spike/research worktree exception. The worktree is described in
  [Method](#method).

## Verdict summary

| Question | Verdict | One-line answer |
|---|---|---|
| G-1 Feasibility and boundary | **FEASIBLE** | `go/parser` + `go/ast` reproduce every finding's text, line and order on the whole corpus; the boundary is `scanSource(rel, src string) ([]string, error)`, with `retiredarch` re-anchored on `scanSource` → `gomask.MaskGoNonCode`. |
| G-2 Fail-closed on unparseable input | **FEASIBLE** | Any parser error, even with a partial AST, is ED-2 exit 1; `PositionFor(pos, false)` ignores `//line`; `filebased` verdicts are unchanged. |
| G-3 Verdict parity | **EXACT PARITY, no HALT** | 0 divergences across 19 fixtures, 3 `filebased` inputs, 17 tracked in-scope files and all 155 module `.go` files; AC-E2.6 items 1–3 suffice. |
| G-4 D-2′ on `*ast.CallExpr` | **FEASIBLE, representation only** | Every D-2′ rejection is preserved; the only behavioural deltas are stricter (8-arg trailing string, ellipsis, split selector, import-path check). |
| G-5 Import aliases | **FEASIBLE, additive** | Named aliases resolve via `*ast.ImportSpec`; a dot import of a write-capable package is a finding; blank imports are inert; zero live write-capable aliases. |
| G-6 `NewRoot` receiver binding | **Syntactic intra-function tracking** | 3 `NewRoot(` sites, 0 `.Resolve(` calls; `go/types` rejected on cost and dependency grounds. |
| G-7 Sizing | **All within the 2-hour rule** | E-T2 M/medium (de-risked from high), E-T3 M/medium (S before plan-review cycle 2 added the extent-test port), E-T4 M/medium, E-T5 M/medium, E-T6 XS/trivial, new E-T7 M/medium. |

The migration is **feasible**. No question returned a HALT condition, so the
re-plan of E-T2..E-T6 may proceed.

## Method

1. **Worktree.** Stage created a detached spike worktree at
   `logs/spike-049-wt` (an ignored path) from `origin/main` @ `961b652`.
   * It was used only for a throwaway in-package test file. That file was
     never committed, and nothing was pushed from the worktree.
   * The worktree performed no shipment claim and no Ship execution.
   * It was removed with `git worktree remove --force` followed by
     `git worktree prune`. Afterwards `git worktree list` shows only the main
     worktree.
2. **Spike engine.** The throwaway file `zz_spike_test.go` lived in package
   `writepath`, so it could call the unexported `scanText` and
   `selectorOccurrences`.
   * It implemented `astScan(rel, src string) ([]string, error)`, a complete
     candidate engine of about 120 lines.
   * Every test compared `astScan(rel, src)` against
     `scanText(rel, gomask.MaskGoNonCode(src))`, the live production engine.
   * A copy is committed, non-normatively, as
     `assets/2026-10-02-writepath-go-ast-spike/zz_spike_test.go.txt`, with the
     run output in `spike-output.txt` alongside it. The `.txt` suffix keeps it
     out of the build; this document is the record.
3. **Corpora:**
   * **FIXTURES** — every `scripts/testdata/writepath/*.go` (19 files).
     Findings were also checked against the golden `fixture_findings` rows.
   * **FILEBASED** — the golden `filebased` inputs (CRLF, lone CR, BOM, invalid
     UTF-8).
   * **TRACKED** — every tracked file that `shouldScan` admits (17 files).
   * **ALLGO** — every `.go` file in the module except the write-path fixture
     tree (155 files). This is a superset stress corpus well beyond the gate's
     scope.
   * **HAZARDS** — about 40 hand-built inputs, each targeting a named G-3/G-4/G-5
     hazard.
4. **GOROOT verification** follows compound lesson
   `2026-09-08-verify-go-stdlib-internals-claims-against-goroot-source.md`.
   Every stdlib claim below cites `$(go env GOROOT)/src` at `go1.26.5`.

### Spike engine algorithm (proven)

1. Parse with `parser.ParseFile(fset, rel, src, parser.ParseComments|parser.SkipObjectResolution)`.
   **Any** returned error fails the whole file (G-2).
2. Compute `masked := gomask.MaskGoNonCode(src)` once for the whole file.
   `gomask` preserves rune count, so a rune-index map from source byte offsets
   lines the two texts up exactly.
3. Map each byte offset to a line number using `pysem.SplitLines` boundaries
   **of the masked text** (G-3 line hazard).
4. **Selector hits.** Walk every `*ast.SelectorExpr` whose `X` is an
   `*ast.Ident` and whose `X.Name + "." + Sel.Name` is in `Selectors`.
   * The hit line is the line of the `X` identifier.
   * `syscall.CreateFile` is skipped only when it is the direct `Fun` of a
     `*ast.CallExpr` that satisfies the D-2′ predicate (G-4).
5. **Visible raw-string hits** (struct tags and tag-shaped raw strings). For
   every raw-string `*ast.BasicLit`, compare the source interior
   (`Pos+1 .. End-1`) with the same rune range of the whole-file masked text.
   * The interior is **visible** exactly when the two are equal. Only the
     interior is compared, because `gomask` blanks the two backticks
     themselves even when it leaves a tag visible.
   * If the interior is visible, run the existing `selectorOccurrences` over it
     and record a hit for each occurrence.
6. **Output.** De-duplicate by `(line, selector index)`, sort by line and then
   by `Selectors` order, and format exactly as `scanText` does.

## G-1 — Feasibility and the source-level boundary

**Verdict: FEASIBLE.**

**Evidence.** Across all four corpora, `astScan` produced byte-identical output
to `scanText` (see the G-3 table), with every finding's text, line and order
matched.

**Boundary decision.** E-T2 introduces an unexported
`scanSource(relPath, src string) ([]string, error)`.

* It takes the decoded, **unmasked** `pysem.ReadText` output and returns
  findings plus an error. This error is the only path for a G-2 parse failure.
* `scanFile` becomes `ReadText` → `scanSource`.
* Inside `scanSource`, `gomask.MaskGoNonCode(src)` remains the **single
  canonical masker**, used for two purposes:
  * struct-tag and raw-string visibility (step 5 above);
  * the line map (step 3 above).

  No second masker and no standalone re-masking of literals is introduced. A
  literal masked on its own loses its struct-tag context, and the spike's
  first iteration diverged on exactly that point before switching to
  whole-file interior comparison.

**`retiredarch` re-anchor decision.** `writepath_mask_test.go`'s writepath
half (`writePathMaskFlow`) currently pins two things:

* `scanFile` passes a direct `gomask.MaskGoNonCode(...)` call as the second
  argument to `scanText`;
* `scanText` ranges over `pysem.SplitLines` of that parameter.

After E-T2, the same invariant is re-anchored on `scanSource`:

* `scanFile` calls `scanSource` with the `ReadText` result.
* `scanSource` binds a direct `gomask.MaskGoNonCode(src)` call (qualifier
  bound to the gomask import path) and derives its line table from
  `pysem.SplitLines` of that bound result.
* No other reference to `scanSource` exists (the same alias, method-value and
  shadow rejections as today).

The two other halves of `retiredarch` are untouched:

* the canonical-mask half, `TestMaskerDefinedExactlyOnce` with
  `canonicalMaskerDef = "internal/gomask/gomask.go:MaskGoNonCode"`;
* the single-definition half.

The re-anchor is a **rename and shape update of one helper**, not a weakening:
the protected property is still "production detection consumes the one
canonical masker". It belongs in the same E-T2 commit, because the pin and the
engine must change atomically.

## G-2 — Fail-closed on unparseable input

**Verdict: FEASIBLE.**

| Input | `go/parser` result | Disposition |
|---|---|---|
| `oracle-code-position-boundaries` (code-position `\f`) | `illegal character U+000C` | ED-2 exit 1 |
| Code-position U+0085 | illegal character | ED-2 exit 1 |
| Mid-file BOM in code | `illegal byte order mark` | ED-2 exit 1 |
| Mid-file BOM inside a comment | `illegal byte order mark` | ED-2 exit 1 (stricter; the masker would erase it) |
| Missing closing brace (partial AST returned) | error + partial `*ast.File` | ED-2 exit 1; the partial AST is never scanned |
| Bad declaration (partial AST returned) | error + partial `*ast.File` | ED-2 exit 1 |
| No package clause | error | ED-2 exit 1 |
| BOM at offset 0 | accepted (skipped) | parity |
| CRLF, lone CR (`filebased`) | accepted | golden-equal |
| Invalid UTF-8 (`filebased`) | not reached: `pysem.ReadText` returns `pysem.ErrInvalidUTF8` first | unchanged (AC-D1a.1) |

**GOROOT evidence (`go1.26.5`):**

* `go/scanner/scanner.go:59` — `bom = 0xFEFF // byte order mark, only permitted as very first character`.
* `go/scanner/scanner.go:94-95` — `r == bom && s.offset > 0` → `"illegal byte order mark"`.
* `go/scanner/scanner.go:161-162` — a BOM at offset 0 is skipped.
* `go/scanner/scanner.go:974` — `illegal character %#U` for any rune not
  starting a token, including `\f`, `\v`, U+0085 and U+2028 in code position.

**Design rule for E-T2.** `scanSource` returns the `go/parser` error whenever
`err != nil`, regardless of whether `*ast.File` is non-nil. It never uses
`parser.AllErrors` and never scans a partial tree. `runRepoScan` and
`runFixtureSelfTest` already turn `scanFile` errors into `::error::` lines with
exit 1 (the ED-2 path), so no new exit path is needed.

**`//line` directives.** In a hand-built fixture, `//line other.go:900` precedes
an `os.WriteFile` call.

* With `fset.Position(pos)`, the hit would report `other.go:900`.
* With `fset.PositionFor(pos, false)` it reports the physical line, matching
  `scanText`.

E-T2 must use `PositionFor(pos, false)` everywhere and pin it with a committed
`//line` test (AC-E2.3).

## G-3 — Verdict parity against the full Unit D corpus

**Verdict: EXACT PARITY. AC-E2.6 items 1–3 suffice; no HALT.**

| Corpus | Files | Equal | Diverge | Fail-closed parse errors | Notes |
|---|---|---|---|---|---|
| FIXTURES (`scripts/testdata/writepath/*.go`) | 19 | 19 | 0 | 0 | 0 mismatches against the 19 golden `fixture_findings` rows |
| FILEBASED (golden `filebased`) | 3 | 3 | 0 | 0 | CRLF, lone CR and BOM are golden-equal; invalid UTF-8 fails at `ReadText` before either engine |
| TRACKED (`shouldScan` scope) | 17 | 17 | 0 | 0 | — |
| ALLGO (every module `.go` except write-path fixtures) | 155 | 155 | 0 | 0 | stress superset |

### Named hazards

* **Struct tags (INV-2).** `reject-struct-tag-selector.go` and
  `reject-tag-shaped-raw-string-expr.go` stay **rejected**, with identical
  findings.
  * A naive `SelectorExpr`-only walk would flip both to accept. The step 5
    interior comparison against the whole-file mask reproduces the masker's
    visibility decision exactly.
  * Further hand-built hazards are EQUAL:
    * `tag-with-boundary-in-value` (hits on lines 5 and 6);
    * `tag-shaped-raw-expr` (line 3);
    * `non-tag-raw-expr` (blanked, so no hit);
    * `accept-non-tag-raw-string-selector.go`.
* **Line numbering.** `go/token` counts lines on `\n` only
  (`go/token/position.go:219`, `if b == '\n'`), while the golden uses
  `pysem.SplitLines` boundaries. Mapping offsets through `SplitLines`
  boundaries **of the masked text** reproduces the legacy masked-line
  numbering.
  * `oracle-comment-and-string-boundaries` is EQUAL: its boundary runes sit
    inside comments and strings, the masker erases them, and later hits keep
    their masked line numbers (AC-D1a.2).
  * `multibyte-before-hit` and `bom-at-0` are EQUAL; the rune-index map handles
    multi-byte prefixes.
* **Code-position boundary runes.** `\f`, `\v`, U+0085 and U+2028 in code
  position are not Go whitespace. They are `go/parser` errors and become G-2
  fail-closed. This is the only corpus-level verdict change. It is exactly
  AC-E2.6 item 3 (AC-D1a.2's code-position input moves from a parity assertion
  to a fail-closed assertion) and is stricter, never weaker.

### Oracle adaptation (AC-E2.6) confirmed sufficient

1. **Production side → `scanSource(rel, src)`.** The legacy side stays fed the
   masked text. The spike ran exactly this comparison shape.
2. **Fixture inputs frozen to an explicit name list.**
   `TestOracle_FixtureCorpus_Parity` currently globs
   `scripts/testdata/writepath/*.go`. Item 2 replaces the glob with the
   explicit list of the **19** fixture names on `main` at the 031-S merge:
   * accept: `accept-clean.go`, `accept-mentions-in-comment.go`,
     `accept-non-tag-raw-string-selector.go`,
     `accept-syscall-createfile-metadata.go`;
   * reject: `reject-createtemp.go`, `reject-io-copyn-copybuffer.go`,
     `reject-link.go`, `reject-os-chown.go`, `reject-os-chtimes.go`,
     `reject-os-lchown.go`, `reject-os-mkdirtemp.go`,
     `reject-os-openroot.go`, `reject-os-root-type.go`,
     `reject-struct-tag-selector.go`,
     `reject-syscall-createfile-evasion.go`,
     `reject-syscall-createfile-write.go`, `reject-syscall-write.go`,
     `reject-tag-shaped-raw-string-expr.go`, `reject-writefile.go`.

   The 3 `filebased` inputs are already enumerated through the golden.
3. **Code-position boundary input → fail-closed assertion.** Confirmed: the
   spike's only parse error among the oracle's hand-built inputs is
   `oracle-code-position-boundaries`.

No other oracle expectation, legacy copy, or the frozen 20-selector list needs
to change. H-3 is not triggered.

### Divergences found (all stricter, none in any corpus)

These hand-built inputs diverge, and in each case the AST engine **reports
where the masked engine allowed or missed**. That is acceptable under INV-2 (a
stricter verdict). None occurs in the fixtures, the golden, or the live tree.

| Hazard | Masked engine | AST engine | Why |
|---|---|---|---|
| Selector split by spaces, newline or comment (`os . WriteFile`, `os.\nWriteFile`, `os./**/WriteFile`) | no finding (residual item 7) | finding | AST is immune to split selectors; closes item 7 |
| `syscall.CreateFile` with 8 args, last a string | allowed (empty trailing segment trimmed, then 7) | reported | AST counts 8 `Args` |
| `syscall.CreateFile(args...)` | allowed when the segments look bare | reported | `call.Ellipsis != token.NoPos` |
| Visible raw string containing an allowance-shaped `syscall.CreateFile(...)` (analytical, not run) | allowed by the extent predicate if the masker leaves it visible | reported | raw-string hits are never allowance-checked |

The split-selector change is the intended closure of residual item 7. It is
pinned by E-T4's `reject-split-selector.go`, because E-T2/E-T3's golden must
stay byte-identical.

### Hazards with both engines equal (not regressions)

* `line-directive` (with `PositionFor(pos, false)`), `bom-at-0`,
  `multibyte-before-hit`.
* `unimported-os-ident` (both report): matching is by spelling until E-T4.
* `nested-selector` (`x.os.WriteFile`): both ignore it.
* `os-root-type`, `createfile-func-value` (reported).
* Every allowed `CreateFile` variant.
* `split-dispo` and the paren/binary flag spellings (rejected).
* Parenthesized callee `(syscall.CreateFile)(...)` (reported).
* Same-line allowed plus writing call (reported).
* Access `0x0` (rejected).
* Aliased imports, dot imports and blank imports: both engines miss
  `o.WriteFile` and dot-imported `WriteFile`. This is residual item 1/3,
  closed by E-T4, not by the engine swap.
* `syscall` aliased to another package: both allow, because matching is by
  name. E-T3's import-path check makes this stricter.

## G-4 — D-2′ mapped onto `*ast.CallExpr`

**Verdict: FEASIBLE as a representation change only. No widening.**

| D-2′ rule (masked engine, `occurrenceAllowed`) | AST representation (proven in spike) |
|---|---|
| Balanced call extent directly after the selector | `syscall.CreateFile` is the direct `Fun` of an `*ast.CallExpr`; a parenthesized callee or a function value is reported |
| No segment contains `()[]{}`, a quote or a backtick | every `Arg` is a **bare operand**: `*ast.Ident`, an INT/FLOAT/IMAG `*ast.BasicLit`, or `*ast.SelectorExpr`/`*ast.UnaryExpr`/`*ast.StarExpr`/`*ast.BinaryExpr` built only from bare operands; calls, composite literals, index expressions, parens, string and rune literals are rejected |
| Exactly 7 non-empty segments | `len(call.Args) == 7` **and** `call.Ellipsis == token.NoPos` |
| `args[1]` trimmed text is `0` | `Args[1]` is an `*ast.BasicLit` with `Kind == token.INT` and `Value == "0"` (`0x0` and `00` stay rejected) |
| `args[4]` is exactly `syscall.OPEN_EXISTING` | `Args[4]` is a `*ast.SelectorExpr` `syscall.OPEN_EXISTING` whose source text is exactly that token (no interior space or comment) |
| `args[5]` is exactly `syscall.FILE_FLAG_BACKUP_SEMANTICS` | the same rule for `syscall.FILE_FLAG_BACKUP_SEMANTICS` |

**Production addition required by the plan (E-T3):** the callee's `X`,
`Args[4].X` and `Args[5].X` must each resolve to the import path `"syscall"`
through the file's `*ast.ImportSpec` set. The spike matched by name, as the
masked engine does today. Path resolution is **strictly stricter**: a file
that binds the name `syscall` to `golang.org/x/sys/windows` loses the
allowance. On the live tree the only allowance user,
`internal/pathsafe/reparse_windows.go` (the call at line 164), imports
`"syscall"` unaliased (line 8), so the verdict there is unchanged.

**Live-tree check.** The production allowance user,
`internal/pathsafe/reparse_windows.go`, and
`accept-syscall-createfile-metadata.go` keep their accept verdicts. `reject-syscall-createfile-write.go` and
`reject-syscall-createfile-evasion.go` keep their rejections.

## G-5 — Import-alias resolution, dot and blank imports

**Verdict: FEASIBLE and additive.**

* **Live import inventory** at `961b652`, all module `.go` files: the only
  named import is `copilot` → `github.com/github/copilot-sdk/go`, in
  `internal/copilotprobe/client.go`, `fixture.go` and `permission.go`. That
  package is not write-capable.
  * There are zero dot imports, zero blank imports of write-capable packages,
    zero `io/ioutil` imports and zero `golang.org/x/sys` imports.
  * The repo has no `golang.org/x/sys` module dependency.
* **Named alias.** Build `alias → import path` from `f.Imports`. A
  `SelectorExpr` whose `X` names an alias bound to a write-capable path is
  matched as the canonical selector (for example, `o.WriteFile` with
  `o "os"` is reported as `'os.WriteFile'`).
* **Dot import.** A dot import of any write-capable package is reported as a
  finding at the import spec's line, using the package's canonical selector
  prefix. It fails closed: the import itself is the finding, so no
  per-identifier analysis is needed.
  * Per-identifier detection is possible but **not required**. A probe showed
    that `ast.File.Unresolved` (with object resolution enabled) lists exactly
    the dot-imported identifiers, excluding local shadows, composite-literal
    keys and field selectors.
* **Blank import.** `_ "os"` is inert: no identifier is bound.
* **Additivity rule (binding for E-T4).** Alias resolution **adds** matches; it
  must not remove any.
  * A `SelectorExpr` whose `X` is spelled exactly as a selector's package
    qualifier (`os`, `io`, `sql`, `bbolt`, `syscall`) stays a finding even when
    that name is not bound to the canonical import path, as today
    (`unimported-os-ident` is reported by both engines).
  * Keying presence matching **only** on import paths would silently drop
    those findings, a fail-open. Shadowing remains residual item 3.

## G-6 — `NewRoot` receiver binding

**Verdict: syntactic intra-function tracking. `go/types` rejected.**

* **Live inventory** at `961b652`:
  * `NewRoot(` occurs 3 times in code: `internal/config/validate.go` twice
    (lines 42 and 67, both `x, err := pathsafe.NewRoot(...)`), and
    `internal/pathsafe/root.go` once (the definition at line 340).
  * There are **zero** `.Resolve(` call sites in production code.
* **Syntactic tracking.** Within one `*ast.FuncDecl` or `*ast.FuncLit`:
  * record identifiers bound by `x := pathsafe.NewRoot(...)`,
    `x, err := pathsafe.NewRoot(...)` or `var x = pathsafe.NewRoot(...)`,
    where the qualifier resolves through E-T4's alias map to
    `github.com/softwaresalt/intercom-go/internal/pathsafe`;
  * fire on `x.Resolve(...)` for a recorded `x`.

  The cost is roughly 60–80 lines plus fixtures and needs no package loading.
* **False-negative profile.** The tracking misses cross-function flows (a root
  passed as a parameter or returned), struct-field storage, package-level
  variables, and method values (`f := x.Resolve`). These remain a documented
  residual.
  * The tripwire's purpose (AC-4.4) is to fire on the **first** production
    caller, and the expected first caller shape (construct and resolve in one
    function) is covered.
  * A conservative extension, firing on any `.Resolve(` in a file that imports
    `pathsafe` when the receiver is unknown, is an option. It is **not** chosen
    as the default because it fires on unrelated `.Resolve(` receivers
    (AC-E5.2).
* **`go/types` rejected.** It is precise across flows, but needs
  `golang.org/x/tools/go/packages` or a hand-built importer, which brings:
  * whole-module type checking on every gate run;
  * a new dependency (Constitution Principle VI);
  * a build-tag and `GOOS` matrix (`reparse_windows.go`);
  * a materially larger change than the 2-hour rule allows.

  The receiver-tracking benefit does not justify that cost at zero production
  callers.

## G-7 — Sizing against the 2-hour rule

| Task | Size | Complexity | Files | Test scenarios | Basis |
|---|---|---|---|---|---|
| E-T2 049.002-T engine swap | M | **medium** (was high) | `writepath.go`, `writepath_test.go`, `writepath_oracle_test.go`, `retiredarch/writepath_mask_test.go` | 3 (parse fail-closed incl. partial AST, `//line`, golden byte-identity re-route) | the engine is proven at about 120 lines; the parity risk is retired by this spike |
| E-T3 049.003-T predicate port | M (rev 7 cycle 2; was S) | medium | `writepath.go`, `writepath_test.go`, `writepath_extent_test.go` | 3 (8-arg-string, ellipsis, import-path `syscall`) plus the AC-D1.1–D1.3 extent-test port | predicate proven; delete `extractExtent`/`occurrenceAllowed`/`scanText`/`advanceCursor` and port their tests to AST spans |
| E-T4 049.004-T alias resolution | M | medium | `writepath.go`, `writepath_test.go`, fixtures + golden | 4 fixtures (AC-E4.1) | additivity rule (G-5) |
| E-T5 049.005-T `Resolve` tripwire | M | medium | `writepath.go`, `writepath_test.go`, fixtures + golden | 3 (fires, unrelated receiver, tracked tree clean) | G-6 syntactic strategy |
| E-T6 049.006-T residual retirement | XS | trivial | `writepath.go` header + `scripts/check-write-path-precondition.sh` header (the two residual-record locations) | 1 (AC-D4.2 scope check) | docs only |
| **E-T7 (new) selector-set widening, stash 458F9385** | M | medium | `writepath.go`, `writepath_test.go`, fixtures + golden | 3 (ioutil, syscall, x/sys path-keyed) | import-path keying needs E-T4 |

E-T2 touches four files and E-T3 three, which is at or past the upper edge of
the 2-hour heuristic (fewer than 3 files). In both, the extra files are
test-only, and the production change is confined to `writepath.go`. They are
retained as single tasks because the production change and its tests must
change **atomically**: in E-T2 the engine, the oracle and the `retiredarch` pin
(AC-E2.6, G-1); in E-T3 the extractor deletion and the
`writepath_extent_test.go` port, without which the package does not compile.
Splitting them would leave `main` red between commits. E-T4's physical count
is higher still; see the counting convention below.

**Counting convention.** The Files column counts *artifacts*, following the
D-T6 convention: a fixture set plus its golden rows counts as one artifact.
Plan revision 7 also states the physical file counts (each fixture `.go` file
and the golden JSON counted separately): E-T2 4, E-T4 7, E-T5 5, E-T7 6.
The physical counts exceed the heuristic for E-T4, E-T5 and E-T7. Every extra
file is a test fixture or golden row for the same contract, so the deviation
is recorded in the plan's Constitution Check (row VI) rather than split.

## Residual-item disposition after Unit E (forecast)

| D-T4 residual item | Closed or narrowed by |
|---|---|
| 1 Named import aliases | E-T4 (G-5) |
| 2 `NewRoot` / `Root.Resolve` tracking | E-T5 (G-6), with the cross-function residual recorded |
| 3 Dot imports, blank imports, shadowing | dot-import clause closed by E-T4; blank imports are inert; shadowing **remains** |
| 4 `os.Root` methods, `(*os.File).Write*` | **remains** |
| 5 Undecidable call shape → rejected | **remains** (a false-positive surface, not a hole) |
| 6 Primitives outside the selector set | **narrowed** by E-T7 (stash 458F9385, 50 selectors); the uncovered same-family remainder is deferred as stash **C0D28448** |
| 7 Split selectors | E-T2 engine, pinned by E-T4's `reject-split-selector.go` |
| 8 Dynamic proc invocation (new) | **remains**; `syscall.NewLazyDLL` is live at `internal/pathsafe/reparse_windows.go:19`, deferred as stash **FE2F02FF** |

Plan revision 7 adopts this forecast with the item-6 correction above
(plan-review rev 7 PR7-1). Items 2 and 6 are narrowed, not closed.

## Reproduction

The spike source and its recorded output are committed as evidence under
[`assets/2026-10-02-writepath-go-ast-spike/`](assets/2026-10-02-writepath-go-ast-spike/)
(`zz_spike_test.go.txt` and `spike-output.txt`). The `.txt` suffix keeps the
source out of the Go build. To reproduce:

1. Copy `assets/2026-10-02-writepath-go-ast-spike/zz_spike_test.go.txt` to
   `tools/gatecheck/internal/writepath/zz_spike_test.go` in a scratch worktree
   at `961b652`.
2. Run `go test ./tools/gatecheck/internal/writepath -run Spike -v` from the
   repository root (the gate is part of the root module
   `github.com/softwaresalt/intercom-go`).

Do not commit the file. The normative tests are written red-first by E-T2..E-T7.

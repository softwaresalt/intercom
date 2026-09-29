---
title: "Implementation Plan — Migrate CI gate engines from Python to Go (parity-preserving port) and retire gate-engine Python"
date: 2026-09-28
status: reviewed — plan-review attempt 3 ADVISORY; all advisory findings remediated in revision 3.1
agent: Stage
source_document: docs/decisions/2026-09-28-intercom-go-gate-engine-python-to-go-migration-deliberation.md
source_stash: C44E2C1F
folded_stash: [8387758F, 150364D2]
governs: four release units — M1 scaffold+masker+write-path, M2 retired-architecture, M3 unignore-regression+merge-strategy, M4 Python retirement
bound_snapshot: 1651936 (origin/main)
stage_branch: chore/stage-stash-python-to-go-migration
requires_plan_hardening: yes
revision: 3.1
---

# Implementation Plan — Gate-engine Python → Go migration

**Requires plan hardening: yes.** Every unit either modifies or re-hosts a merge-blocking control that
`ci.yml`'s `ci-gate` job reaches: retired-architecture, write-path precondition, or unignore-regression. The
advisory merge-strategy gate is also re-hosted. M4 deletes the fallback engines. See §11 and §13.

Revision 2 answered plan-review attempt 1, and revision 3 answers attempt 2 (§15). The finding-to-revision maps are
in §15.3 and §15.4.

## 1. Problem Frame

Five gate engines are Python today. CI steps are identified by **job name and step name**, not by line
number, so the references stay valid after any ci.yml edit.

| Gate (wrapper) | Engine today | CI job → step(s) | Blocking? | Kill switch |
|---|---|---|---|---|
| `scripts/check-write-path-precondition.sh` | inline heredoc importing `scripts/lib/gomask.py` | `lint` → "Run write-path-precondition gate" / "Run write-path-precondition self-test" | verdict is blocking by default; the integrity step is unconditional | `WRITE_PATH_GATE_ADVISORY` (verdict only) |
| `scripts/check-retired-architecture.sh` | `scripts/lib/retired_arch.py` + `gomask.py` | `lint` → "Run retired-architecture gate" / "Run retired-architecture self-test" | same shape as write-path | `RETIRED_ARCH_GATE_ADVISORY` (verdict only) |
| `scripts/check-unignore-regression.sh` | inline heredoc | `gitignore-append-only` → "Self-test the un-ignore regression checker's own logic" / "Check un-ignore regression (behavioural differential)" | blocking | **none** |
| `scripts/check-merge-strategy.sh` | inline `evaluate_json` heredoc (bash does the orchestration) | `merge-strategy` → "Self-test the checker's own logic (regression guard)" / "Run merge-strategy gate" | the self-test is blocking; the verdict is advisory unless `MERGE_STRATEGY_GATE_REQUIRED` is set | toggle |
| Python unit tests | `scripts/lib/tests/*.py` (5 files) | `lint` → "Set up Python (gate engines)" / "Python gate-engine lint and unit tests" | blocking | none |

The deliberation (D-1…D-8) chose a **parity-preserving port first**:

* The target is a Go tool in the root module, `tools/gatecheck`.
* Wrappers switch to the Go tool one at a time. Each switch carries goldens and evidence captured from the
  Python engine at the parent commit of that change.
* The Python code and the lint-job Python steps are deleted only after every wrapper has switched.
* Held shipments 030-S, 031-S and 032-S are re-planned onto Go after M4.

## 2. Requirements Trace

| Requirement (C44E2C1F or deliberation) | Implementing tasks |
|---|---|
| R-1 Reimplement the engines in Go with `go test` coverage | M1-T4…T8, M2-T2…T10, M3-T3…T6, M3-T9 |
| R-2 Put the tool somewhere the gates do not scan (D-2) | M1-T4 (`tools/gatecheck`, engines under `tools/gatecheck/internal/`) |
| R-3 Wrappers preserve exit codes and never use `go run` (D-3, F-1) | M1-T9 (runner), M1-T10, M2-T11, M3-T7, M3-T11 |
| R-4 Byte-identical masker (D-4) | M1-T2, M1-T7 |
| R-5 Identical verdicts and identical self-test / self-test-integrity modes (D-7) | goldens M1-T2/T3, M2-T1, M3-T1/T2; evidence in each switch task |
| R-6 Pathspec pin stays anchored to source text (D-6) | M2-T8 |
| R-7 TOML dual-engine agreement retained (D-5) | M2-T5, M2-T6 |
| R-8 Before/after evidence, as 029-S did | every switch task, plus M4-T9 |
| R-9 Remove the lint-job Python steps and keep the autoharness pin | M4-T5, M4-T6 |
| R-10 Update docs, the LOCAL DIVERGENCE header and CI wiring | M3-T8, M3-T10, M3-T13, M4-T5, M4-T7, M4-T8 |
| R-11 Fold in 8387758F (struct-tag write-path fixtures) | M1-T1 |
| R-12 Fold in 150364D2 (merge-strategy trap, exit table, rev-parse guard) | M3-T12, M3-T13 |
| R-13 No Python-test assertion is lost | M4-T1 coverage map, M4-T2, M4-T3 |
| R-14 Security-review fail-closed hardening, permitted only as enumerated deltas (ED-7, and the trailing-data and BOM parity rules) | M1-T8, M2-T5, M3-T9 |

## 3. Cross-cutting design contracts (bind every task)

### C-1 Package layout and import isolation

* `tools/gatecheck/main.go` is `package main`. It exposes a testable
  `run(args []string, stdin io.Reader, stdout, stderr io.Writer) int` and a sub-command table.
* Engines live in `tools/gatecheck/internal/{pysem,gomask,writepath,retiredarch,unignore,mergestrategy}`.
  * **Go's `internal` visibility rule is the isolation mechanism (INV-7).** Only packages rooted at
    `tools/gatecheck/` can import these packages. The compiler therefore rejects any import from the product's
    `cmd/**` or `internal/**`. `tools/gatecheck` itself is `package main`, so it cannot be imported at all.
    `.golangci.yml` and depguard need no change.
  * **Gate selection stays clear of the tool.** Every gate's `git ls-files` pathspec (`internal/**`, `cmd/**`,
    `config.toml.example`) runs from the repository root, so it is root-anchored. The prefix predicates test
    `startswith('internal/')` and `startswith('cmd/')` on repo-relative paths. Neither mechanism matches
    `tools/gatecheck/internal/…`. M2-T9 asserts this directly: no selected path starts with `tools/`.
* **Imports.**
  * Standard library only, plus `github.com/BurntSushi/toml` inside `internal/retiredarch` (D-5).
  * `go.mod` and `go.sum` must not change. The module is already required at v1.6.0.
  * Any need to add a module, or to bump `BurntSushi/toml` or the `go`/`toolchain` directive, is a **stop
    condition** (§13).

### C-2 Exit-code contract

Exit codes are **pinned per gate and mode by the goldens**. Nothing is asserted from memory. The known classes
today:

| Source | Behaviour | Exit |
|---|---|---|
| Python engine | pass or SKIP | 0 |
| Python engine | findings or self-test failure (`raise SystemExit(1)`) | 1 |
| Python engine | usage or infrastructure error (`raise SystemExit("msg")`) | 1, with `msg` on stderr |
| Python engine | uncaught exception (traceback) | 1 |
| Bash wrapper | usage error, or interpreter missing | **2** (only the bash wrappers emit 2 today) |

* Go engines return a typed `*exitError{code int; msg string}` or a plain `int`. `main` is the only `os.Exit`
  call site.
* The engine-level usage and infrastructure classes reproduce Python's code **1** and print `msg + "\n"` to
  stderr, exactly as `SystemExit(str)` does.
* The Go tool's own `run()` usage errors (a missing or unknown sub-command) return 2. The wrappers never call
  it that way, so no gate observes this code.
* `run()` has a top-level `recover`. A panic prints `::error::gatecheck internal error: <value>` to stderr and
  returns **1**, matching Python's uncaught traceback (exit 1). It never produces Go's default panic exit code 2.

### C-3 Wrapper runner (`scripts/lib/gatecheck-run.sh`, sourced)

* **Source anchoring.** The tool source is located **relative to the runner file**, not the working directory:
  `GATECHECK_SRC="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"`. This matches how the Python engines
  anchor on `__file__` or `SCRIPT_DIR`.
  * The gated tree's root is passed explicitly as `--root "$ROOT"`. **Each wrapper derives `$ROOT` exactly as
    it does today:**
    * write-path, unignore and merge-strategy use the working directory's `git rev-parse --show-toplevel`
    * **retired-arch** anchors on its engine's own location (Python `resolve_repo_root` uses the module
      directory), so its wrapper passes `git -C "$GATECHECK_SRC" rev-parse --show-toplevel`
  * The binary lives in `mktemp`, so it never tries to find the root from its own location.
  * M1-T3 and M2-T1 capture a run from a **subdirectory** working directory. M2-T1 also captures a run of the
    retired-arch wrapper, by absolute path, from **inside another repository**. Together these prove the root
    derivation is identical.
* **`gatecheck_build`** is called once per wrapper process. It runs after any wrapper precondition that must
  exit first; for example, the ED-1 `rev-parse` guard runs **before** the build.
  * It creates `GATECHECK_TMP="$(mktemp -d)"` and runs:
    `env -u GH_TOKEN -u GITHUB_TOKEN GOTOOLCHAIN=local GOFLAGS=-mod=readonly go -C "$GATECHECK_SRC" build -trimpath -o "$GATECHECK_TMP/gatecheck$(go env GOEXE)" ./tools/gatecheck`
    (`go -C` requires Go 1.20 or later, and `-C` must be the first flag. The `GOEXE` suffix lets a Windows Git
    Bash run execute the binary.)
  * It sets `GATECHECK_BIN`.
  * If the build fails, it prints `::error::gatecheck build failed (exit N)` to stderr and **returns 2**. The
    wrapper then exits 2 (ED-3 message class).
* **`gatecheck_invoke <subcommand> [args…]`**
  * Runs `env -u GH_TOKEN -u GITHUB_TOKEN "$GATECHECK_BIN" <subcommand> --root "$ROOT" [args…]`. **No engine
    ever sees a token**, because none needs one (PD-3).
  * stdout and stderr are **streamed**, never captured or buffered by the runner.
  * It **returns** the binary's status and **never** calls `exit`. That makes it safe inside `if`, `|| rc=$?`
    and `$(…)` command substitution.
* **`gatecheck_cleanup`** removes `$GATECHECK_TMP`. It is **idempotent**: safe to call twice, and safe when the
  directory was never created. The runner **never installs a trap**. Each wrapper installs exactly three:
  * `trap cleanup EXIT`
  * `trap 'exit 130' INT`
  * `trap 'exit 143' TERM`

  The signal traps convert the signal into an exit, so the script **does not resume** after it, and the EXIT
  trap runs the single composed `cleanup()`. That function calls `gatecheck_cleanup` together with the
  wrapper's own cleanup (for example, 150364D2's temp file). Traps are therefore never overwritten.
* **Temp-dir containment.** The build output goes to the OS temp directory (`mktemp -d`, which is
  `$RUNNER_TEMP`-backed in CI). It is never written inside the repository tree or above the working directory,
  and it is removed on every exit path.
  * This is the same class of scratch write as today's engines, which already use Python `tempfile` (the
    unignore scenario repos).
  * Principle IV governs agent file operations. This is a CI tool's scratch space. It is recorded here as the
    containment justification.
* Wrappers keep their existing bash argument parsing and usage messages verbatim.

### C-4 Python semantics (`internal/pysem`: tasks M1-T5a, M1-T5b, M1-T6a, M1-T6b)

Ported code calls `pysem` instead of Go stdlib equivalents wherever the Python engine relies on these
semantics. Each helper has a named table test with Python-derived expected values; the values are produced by
M1-T2's transient generator and stored as JSON goldens.

| Helper | Python semantics reproduced |
|---|---|
| `IsSpace` | `str.isspace()`, including `\x1c`–`\x1f`, `\x85`, `\xa0` and Unicode Z / `White_Space` |
| `IsWord` | `str.isalnum()` or `_` (the re Unicode `\w`) |
| `IsDigit` / `IsDecimal` | `str.isdigit()` (Numeric_Type=Digit or Decimal) and `str.isdecimal()` / re `\d` (`Nd`). Go has no Numeric_Type table, so `IsDigit` uses a **committed static table** `digit_table.go`. A transient Python command generates it (`unicodedata.digit` presence minus `Nd`). The command sits in the file's header comment, and no `.py` file is committed |
| `IsUpper` / `IsLower` | Python cased-character rules |
| `Lower` | `str.lower()` **full** case mapping. That needs an **explicit multi-rune table** from SpecialCasing (for example `İ` → `i̇`, 2 runes), plus Python's context-dependent **Final_Sigma** rule for `Σ`. `unicode.SpecialCase` cannot express either, because it holds only 1:1 deltas |
| `Strip` | `str.strip()` with no arguments (`IsSpace` set) |
| `SplitLines` | `str.splitlines()` boundaries: `\n`, `\r`, `\r\n`, `\v`, `\f`, `\x1c`, `\x1d`, `\x1e`, `\x85`, U+2028, U+2029 |
| `Repr` | `repr(str)` exactly as used by `!r` in findings: quote-choice rule, `\\`, `\n`, `\r`, `\t`, `\xNN`, `\uNNNN`, `\UNNNNNNNN`, and printable non-ASCII kept verbatim (`str.isprintable`) |
| `WordBoundary` | the re `\b` over `IsWord` |
| `PrecededByWordOrDot`, `FollowedByWord` | explicit replacements for the lookbehind `(?<![\w.])` and the lookahead `(?![\w])`, which RE2 cannot express |
| `ReadText(path)` | `Path.read_text(encoding='utf-8')`: `utf8.Valid`, otherwise the ED-2 error; then universal-newline translation `\r\n`→`\n`, lone `\r`→`\n` |
| `GitText(out)` | `subprocess.run(text=True)` output: UTF-8 decode plus the same universal-newline translation |

Further rules:

* All iteration is by code point (`utf8.DecodeRuneInString`). Byte-index iteration over text is forbidden in
  engine code.
* Regular expressions are used only where RE2 and Python agree for the actual input class. Any `\s`, `\w`,
  `\b`, `\d`, lookaround or case-insensitive flag in a Python pattern is replaced by `pysem` code.
* **No `bufio.Scanner` with its default buffer.** Engines read whole files or whole command outputs and split
  with `pysem.SplitLines` or `strings.Split`. That removes the default 64 KiB token-limit truncation class.
* **Unicode version skew.** Go 1.26 carries Unicode **15.0.0** (`unicode.Version`). The local Python 3.14.3
  carries **16.0.0** (`unicodedata.unidata_version`); both were checked during staging. Rules:
  * **Golden inputs** use only runes assigned in Unicode 15.0 or earlier. The generator records both versions.
  * **Classification of runes first assigned after 15.0** follows Go's tables (ED-9).
  * **Live-corpus hash and finding parity** at each switch is the backstop. It compares Python and Go over
    every scanned tracked file, so any real divergence is a C-8 stop condition.

### C-5 Goldens are fixture-scoped; the live corpus is evidence only

* **Committed goldens** (`tools/gatecheck/internal/<pkg>/testdata/*_golden.json`) cover **frozen inputs
  only**:
  * the `scripts/testdata/**` fixtures
  * edge cases, stored inline as JSON strings
  * **file-based** goldens for the newline and encoding cases (CRLF, lone CR, BOM, invalid UTF-8)
* Goldens store **names, outcomes, findings and exit codes**. Self-test lines that embed live corpus counts
  (for example the selection self-test's counts of `internal/**` files) are matched as **name + outcome**.
  Their count text is compared only in evidence.
* **Stream comparison is ordered bytes** (stdout and stderr compared separately). It is never a line set.
* The **live corpus is compared once, in each switch task**, in its evidence file:
  * masked-output SHA-256 over every tracked `.go` file in `cmd/**`, `internal/**` and `scripts/testdata/**`
  * the repo-mode verdict and exit code
  * the selected path set

  Each of these is taken at the change's parent commit and at its head.
* **Goldens are generated from the Python engine at the change's parent commit** by a transient command. The
  command is recorded verbatim in the evidence file. **No `.py` file is committed.** The Python version used is
  recorded.

### C-6 Deterministic ordering

No output may depend on Go map iteration order.

* TOML (M2-T5): walk the **decoded tree recursively**, including `map[string]any`, `[]map[string]any` (arrays
  of tables) and `[]any`. Arrays keep index order. `toml.MetaData.Keys()` is used **only to order** sibling
  keys in document order, with one complication:
  * **`Keys()` does not index array elements.** BurntSushi v1.6.0 `parse.go` records `[[t]]` elements, and
    inline tables inside arrays, without an element index; the same key path repeats for every element.
  * So order is recovered by **consuming `Keys()` sequentially, in lockstep with the decoded tree**. The walker
    holds a cursor into the ordered key list and descends the decoded value at the same time. Each **array
    element consumes exactly the entries that belong to it**, counted recursively from the decoded element:
    one entry per key, plus the entries of any nested tables and arrays. This works identically for `[[t]]`
    elements and for inline tables inside arrays. Inline tables have no per-element header, so header counting
    alone would not work.
  * **Nested arrays of tables** (`[[t]]` … `[[t.u]]` … `[[t]]` … `[[t.u]]`) restart the `t.u` element count for
    each parent `t` element. A `[t.sub]` table belongs to the most recent `t` element.
  * Goldens cover elements that list their keys in **different orders** and **different key sets**, inline
    tables inside arrays, and nested arrays of tables. If a shape's Python order cannot be reproduced exactly,
    that is a **stop condition** (halt → Stage). The implementer must not sort as a fallback.
* **UTF-8 BOM.** BurntSushi strips a leading BOM in `parse()`, and strips UTF-16 BOMs too. `tomllib` receives
  `\ufeff` from `read_text` and fails closed. To keep **parity**, the Go primary engine rejects a leading U+FEFF
  **before** decoding, producing the same fail-closed parse-reject finding. UTF-16 input never reaches the
  decoder, because `pysem.ReadText` rejects it first as invalid UTF-8 (ED-2). This is not an ED; it is
  golden-pinned in M2-T1.
* Every other output is sorted exactly where the Python code sorts, and nowhere else.
* Each engine test also runs once with `-count=3 -shuffle=on` in evidence.

### C-7 Evidence files

There is one evidence file per unit, so M2 and M3 can proceed in parallel without editing a shared file:
`docs/plans/evidence/2026-09-28-gate-engine-go-migration/{m1,m2,m3,m4}.md`. Each task appends its own section.
An evidence section is the **verification output of the task that owns it**, not a separate docs-domain
deliverable. Task domains in §4–§7 name the domain of the task's *change*.

### C-8 Enumerated deltas (ED)

Only these deltas are permitted. Any other non-identity seen in evidence is a **stop condition**: halt and
return to Stage (P-010).

| ED | Delta | Direction | Justification |
|---|---|---|---|
| ED-1 | Merge-strategy run outside a checkout: today git's `rev-parse` failure code under `set -e`; after, **2** with a message | fail-closed | 150364D2 item 3 |
| ED-2 | Invalid-UTF-8 input, or a git/read error: today a Python traceback with exit 1; after, a one-line `::error::` message with exit 1 (a synthetic finding where the engine has a finding channel) | same exit and same verdict; only the message changes | the traceback is not a contract |
| ED-3 | Missing interpreter becomes a build failure; the message text changes | exit stays 2 | the runner (C-3) |
| ED-4 | TOML parse-error prose now comes from BurntSushi instead of `tomllib` | the fail-closed reject finding and exit code are identical | D-5 |
| ED-5 | Self-test output gains the rows for fixtures and table rows added *by this migration* (M1-T1, M3-T12) | additive only; every pre-existing row is identical | evidence compares **per switch parent** (the parent already contains the earlier additions); M4-T9 subtracts ED-5 rows explicitly |
| ED-6 | Documents that are valid TOML 1.1 but invalid TOML 1.0: today a fail-closed parse reject; after, parsed and scanned. This is **unconditional**: the v1.6.0 README says "Compatible with TOML v1.1.0", and the module has no environment toggle (no `os.Getenv`; checked during staging) | content is still fully scanned, so nothing is hidden | **stop condition** if any **tracked** TOML file uses 1.1-only syntax (a live verdict would change) |
| ED-7 | Write-path repo mode with an empty selection: today exit 0 and silent; after, exit 1 with `::error::` | fail-closed | cannot happen on a healthy checkout; requested by the security review (traced as R-14) |
| ED-8 | Merge-strategy JSON. The parse-error SKIP reason now comes from `encoding/json` instead of `JSONDecodeError`. Python-only literals (`NaN`, `Infinity`) now SKIP via the parse path instead of the not-boolean path. **Trailing data after the top-level value stays a SKIP**: the Go evaluator requires a second `Decode` to return `io.EOF`, matching Python's "Extra data" rejection. Invalid UTF-8 input is the ED-2 class (exit 1) | verdict (SKIP) and exit (0) are identical; nothing moves toward PASS | the reason is prose, not a contract |
| ED-9 | Runes first assigned after Unicode 15.0 classify per Go's tables (15.0), not Python's 16.0 tables. `IsDigit` is the exception: it uses the generated `digit_table.go`, whose source Unicode version is recorded in its header | none on the live corpus (checked by live parity at each switch) | a toolchain constraint (C-4) |

## 4. Unit M1 — Scaffold `tools/gatecheck`, `pysem`, the masker and the write-path engine (feature, shipment; blocked by 029-S)

**Scope and bridge.**

* M1 stands up the tool and `pysem`, ports `gomask`, ports the write-path engine, and switches
  `check-write-path-precondition.sh`.
* The retired-architecture wrapper keeps using `scripts/lib/gomask.py` until M2. This is a **temporary parity
  bridge**: the M1 masker goldens are the same bytes both engines must produce, so the two maskers cannot
  drift during the bridge. M1-T7 adds no re-plumbing of the retired-arch wrapper.

Domain column: `code` (Go or bash), `test` (fixtures and goldens), `config` (`ci.yml`), `docs`.

| ID | Task | Domain | Files | Posture | Size | Cx |
|---|---|---|---|---|---|---|
| M1-T1 | Add write-path struct-tag characterization fixtures (8387758F) | test | `scripts/testdata/writepath/` (+1 reject, +1 accept) | characterization-first | S | low |
| M1-T2 | Capture masker and `pysem` goldens from Python | test | `internal/gomask/testdata/masker_golden.json`, `internal/pysem/testdata/pysem_golden.json`, `m1.md` | characterization-first | S | low |
| M1-T3 | Capture write-path engine goldens from Python | test | `internal/writepath/testdata/writepath_golden.json`, file-based fixtures, `m1.md` | characterization-first | S | low |
| M1-T4 | Scaffold the `gatecheck` main, exit contract and sub-command stubs | code | `tools/gatecheck/main.go`, `main_test.go` (+ one `register_<name>.go` per sub-command) | test-first | S | low |
| M1-T5a | `pysem` character classes: `IsSpace`, `IsWord`, `IsDigit` (+ generated table), `IsDecimal` | code | `internal/pysem/classes.go`, `digit_table.go`, `classes_test.go` | test-first (goldens) | M | medium |
| M1-T5b | `pysem` case mapping: `IsUpper`, `IsLower`, `Lower` (multi-rune table + Final_Sigma) | code | `internal/pysem/case.go`, `case_test.go` | test-first (goldens) | M | medium |
| M1-T6a | `pysem` string helpers: `SplitLines`, `Strip`, `WordBoundary`, `Repr` | code | `internal/pysem/text.go`, `text_test.go` | test-first (goldens) | M | medium |
| M1-T6b | `pysem` I/O and lookaround: `ReadText`, `GitText`, `PrecededByWordOrDot`, `FollowedByWord` | code | `internal/pysem/io.go`, `io_test.go` | test-first (goldens) | S | medium |
| M1-T7 | Port `gomask` as a rune-level masker on `pysem` | code | `internal/gomask/gomask.go`, `gomask_test.go` | test-first (goldens) | M | medium |
| M1-T8 | Port the write-path engine (selectors, lookaround, modes) | code | `internal/writepath/writepath.go`, `writepath_test.go` | test-first (goldens) | M | medium |
| M1-T9 | Add the shared bash runner with Go-driven red/green tests | code | `scripts/lib/gatecheck-run.sh`, `tools/gatecheck/runner_test.go` | test-first | M | medium |
| M1-T10 | Switch the write-path wrapper to Go (and wire the `write-path` registration), with parity evidence | code | `scripts/check-write-path-precondition.sh`, `tools/gatecheck/register_write_path.go`, `m1.md` | migration-first | S | medium |

**Task detail and acceptance criteria (ACs).**

* **M1-T1.** Add fixtures in the existing naming convention:
  * a **reject** fixture with a qualified write selector inside a real struct tag (a field tagged with key `x`
    and value `os.Remove`), which is 029-S decision delta D-1
  * an **accept** fixture with the same text inside a non-tag raw string

  ACs:
  * `bash scripts/check-write-path-precondition.sh --self-test` passes on the **current Python engine**, both
    fixtures are discovered, and the reject is reported at its expected line.
  * Both fixtures are gofmt- and goimports-clean.
  * Engine code is unchanged.
* **M1-T2.** Generate from the Python engine at the parent commit:
  * **masker goldens** `{name, input, masked}`. These cover every `.go` file under `scripts/testdata/**`, plus
    these named inline edge cases:
    * `multibyte-rune-in-comment`
    * `multibyte-in-string`
    * `x1c-in-struct-tag`
    * `nbsp-between-tag-pairs`
    * `nel-in-tag`
    * `digit-leading-tag-key`
    * `adjacent-tag-pairs-no-space`
    * `multiline-tag-shaped-raw-string` (the C312BD4C characterization, pinned **as-is**)
    * `unterminated-raw-string-eof`
    * `unterminated-string-eof`
    * `escaped-quote`
    * `rune-literal-quote`
    * `crlf-line-endings`
    * `lone-cr`
    * `block-comment-with-newlines`
  * **`pysem` goldens**: for every C-4 helper, input → Python output over an input set that includes every
    rune class that helper distinguishes:
    * ASCII controls
    * `\x1c`–`\x1f`, `\x85`, `\xa0`, U+1680, U+2000–U+200A, U+2028/U+2029, U+3000
    * Latin-1 letters
    * `İ`, `ß`, `ǅ`
    * Greek final sigma
    * Arabic-Indic and full-width digits
    * superscript digits (`isdigit` vs `isdecimal`)
    * `repr` quote-choice cases

  ACs:
  * Every golden comes from Python at this task's parent, and the generator command is recorded verbatim.
  * Every golden input rune is assigned in Unicode 15.0 or earlier. Python's `unidata_version` and Go's
    `unicode.Version` are both recorded (C-4).
  * No `.py` file is added.
  * A re-run reproduces the goldens byte for byte.
* **M1-T3.** Write-path goldens:
  * per-fixture findings
  * **ordered stdout and stderr bytes**, plus the exit code, for `--self-test`, `--self-test-integrity` and an
    unknown mode
  * **file-based** fixtures (JSON-described, materialized by the test into `t.TempDir()`) for CRLF, lone CR,
    BOM and invalid UTF-8 inputs, with Python's result for each (the invalid-UTF-8 case records the traceback
    exit 1 that ED-2 replaces)

  AC: same generation discipline as M1-T2. The repo-mode run is also captured from a **subdirectory** working
  directory (C-3 root anchoring).
* **M1-T4.**
  * `run()` dispatches `write-path`, `retired-arch`, `unignore` and `merge-strategy-evaluate`. **All four
    sub-commands are pre-registered as stubs, each in its own file** (`register_<name>.go`, registering via
    `init()` into the table). M2 and M3 therefore replace different files and never touch adjacent lines. Each
    stub returns `exitError{1, "<name>: not yet ported"}` until its engine lands.
  * A missing or unknown sub-command prints usage to stderr and returns 2 (C-2).
  * `run()` parses the common `--root <dir>` flag (C-3). A missing or non-directory root → exit 1 with
    `::error::`.

  ACs:
  * A **red** table test is committed first. It covers no args → 2, unknown → 2, each stub → 1 with its
    message, an `exitError` round-trip, and a **panicking stub → 1** (the C-2 `recover`).
  * `main` is the only `os.Exit` call site. `go vet` and `golangci-lint run ./...` are clean, including under
    `GOOS=windows`.
* **M1-T5a.** Implement `IsSpace`, `IsWord`, `IsDigit` (with the generated `digit_table.go`) and `IsDecimal`.
  * AC: 100% agreement with `pysem_golden.json` for these helpers.
  * AC: named sub-tests for every class listed in M1-T2.
* **M1-T5b.** Implement `IsUpper`, `IsLower` and `Lower`: the explicit multi-rune SpecialCasing table plus
  Final_Sigma.
  * AC: 100% golden agreement, including `İ`, `ß`, `ǅ` and final vs non-final `Σ`.
* **M1-T6a.** Implement `SplitLines`, `Strip`, `WordBoundary` and `Repr`.
  * AC: 100% golden agreement.
  * AC: `Repr` matches Python for every golden string, including the quote-choice rule.
* **M1-T6b.** Implement `ReadText`, `GitText`, `PrecededByWordOrDot` and `FollowedByWord`.
  * AC: 100% golden agreement.
  * AC: `ReadText` returns a typed `ErrInvalidUTF8` on invalid bytes and translates CRLF and lone CR.
  * AC: file-based tests run in `t.TempDir()`.
* **M1-T7.** Port `mask_go_non_code` and the struct-tag matcher onto `pysem`. The RE2 `\s` and `\w` are not
  used.
  * AC: `MaskGoNonCode` output is **byte-identical** to every `masker_golden.json` entry.
  * AC: rune length and line structure are preserved.
  * AC: the C312BD4C characterization (a multiline tag-shaped raw string stays **unmasked**) passes unchanged.
  * AC: the retired-arch wrapper is untouched (the bridge).
* **M1-T8.** Port:
  * the 20 qualified selectors, verbatim
  * the lookbehind `(?<![\w.])` → `PrecededByWordOrDot`, and any lookahead → `FollowedByWord`
  * the repo mode: `git ls-files -- internal/** cmd/**` through an **injectable `gitRunner`**, then
    `GitText`, `should_scan`, and `ReadText` per file
  * `--self-test` and `--self-test-integrity`
  * `!r`-formatted findings through `pysem.Repr`

  ACs:
  * Per-fixture findings and ordered stream bytes and exit codes match `writepath_golden.json`.
  * Lookaround table tests: `os.WriteFile` → hit; `xos.WriteFile` → miss; `.os.WriteFile` → miss; a non-ASCII
    letter before the selector → miss; a selector followed by a word rune (when the Python pattern has a
    lookahead) → miss.
  * **Fail-closed ACs**, each with an injected runner or file and a test:
    * git exits non-zero → exit 1 plus `::error::` (ED-2)
    * a read error → exit 1 plus `::error::`
    * invalid UTF-8 → exit 1 (ED-2)
    * an **empty selection → exit 1 (ED-7)**
    * a missing fixture directory → exit 1 with the Python message text
  * No error is ignored (errcheck is clean).
* **M1-T9.** Implement C-3.
  * `runner_test.go` runs `bash` (and `t.Skip`s only when `exec.LookPath("bash")` fails) against a temp copy
    of the runner with **PATH-shim fakes**: a fake `go` that fails, and a fake `go` that writes a fake binary
    exiting N.
  * The red tests are committed first:
    * build failure → wrapper exit 2 plus the `::error::` line
    * binary exit 0, 1, 2 and 3 → passed through unchanged, including via `gatecheck_invoke` inside `$(…)` and
      `|| rc=$?`
    * stdout and stderr are streamed, not buffered (the ordered-bytes check with an interleaved fake binary)
    * `GH_TOKEN` and `GITHUB_TOKEN` are absent from the environment of **both** the fake `go` and the fake
      binary
    * `GOTOOLCHAIN=local` and `GOFLAGS=-mod=readonly` are present
    * running with the working directory set to a **different** directory still builds from `GATECHECK_SRC`
      and passes `--root` through
    * `$GATECHECK_TMP` is removed on normal exit, on a non-zero exit, on `INT` (exit **130**) and on `TERM`
      (exit **143**); the script does **not** resume after the signal
    * calling `gatecheck_cleanup` twice is harmless
    * a wrapper's own cleanup still runs (composition)
  * AC: `shellcheck`-clean if `shellcheck` is available; otherwise its absence is recorded.
* **M1-T10.** Replace the `write-path` stub registration with `writepath.Run`. Reduce the write-path wrapper to
  argument parsing plus `gatecheck_build` and `gatecheck_invoke write-path <mode>`, with the C-3 trap set.
  * AC (evidence, C-5): ordered stream bytes and exit codes are identical at parent and head for the repo
    mode, `--self-test`, `--self-test-integrity` and a bogus flag, except ED-3/ED-5.
  * AC: live-corpus masked-output SHA-256 parity is 100%.
  * AC: a deliberately broken `tools/gatecheck` exits 2, never 0.
  * AC: `git grep -n -E '\bpython3?\b' -- scripts/check-write-path-precondition.sh` is empty.
  * AC: `ci.yml` is unchanged. The `lint` job's "Set up Go" step already precedes the gate steps.
  * AC: the environment precheck (§13) is recorded.

## 5. Unit M2 — Port the retired-architecture engine (feature, shipment; blocked by M1)

**Scope.**

* Port `scripts/lib/retired_arch.py` in full onto `internal/gomask` and `internal/pysem`, then switch
  `check-retired-architecture.sh`.
* The port is **faithful**: the two scope expressions (the `select_repo_paths` pathspec and the
  `should_scan_repo_path` prefix test) are ported *as two expressions*.
* `SCAN_SCOPE` unification stays 033-F's contract (D-1, PD-4).

| ID | Task | Domain | Files | Posture | Size | Cx |
|---|---|---|---|---|---|---|
| M2-T1 | Capture retired-arch goldens from Python | test | `internal/retiredarch/testdata/*_golden.json`, `m2.md` | characterization-first | M | low |
| M2-T2 | Port identifier splitting and vocabulary segmentation | code | `retiredarch/ident.go`, `ident_test.go` | test-first (goldens) | M | medium |
| M2-T3 | Port the forbidden-part sequence and concat matchers | code | `retiredarch/match.go`, `match_test.go` | test-first | S | medium |
| M2-T4 | Port the Go-file scan over the Go masker | code | `retiredarch/scango.go`, `scango_test.go` | test-first (goldens) | S | medium |
| M2-T5 | Port the primary TOML engine (recursive walk on BurntSushi) | code | `retiredarch/tomlprimary.go`, `tomlprimary_test.go` | test-first (goldens) | M | medium |
| M2-T6 | Port the fallback TOML lexer and dual-engine agreement | code | `retiredarch/tomlfallback.go`, `tomlfallback_test.go` | test-first (goldens) | M | medium |
| M2-T7 | Port path selection, engine routing and the synthetic fail-closed finding | code | `retiredarch/select.go`, `select_test.go` | test-first | S | medium |
| M2-T8 | Port the source-text-anchored pathspec pin with `go/parser` | code | `retiredarch/pin.go`, `pin_test.go` | test-first | S | medium |
| M2-T9 | Port the repo-selection self-test | code | `retiredarch/selftest_selection.go`, `selftest_selection_test.go` | test-first (goldens) | S | medium |
| M2-T10 | Port the fixture self-test, integrity mode and `Run` | code | `retiredarch/retiredarch.go`, `retiredarch_test.go` | test-first (goldens) | M | medium |
| M2-T11 | Switch the retired-arch wrapper to Go (and wire the `retired-arch` registration), with parity evidence | code | `scripts/check-retired-architecture.sh`, `tools/gatecheck/register_retired_arch.go`, `m2.md` | migration-first | S | medium |

**Acceptance criteria.**

* **M2-T1.** Capture:
  * per-fixture findings for `scripts/testdata/retiredgo/**`, `retiredgo-differential/**` and every
    `scripts/testdata/retired-*.toml`, **under both engines** (`tomllib` and the fallback)
  * `split_identifier` and `segment_whole` outputs over every fixture identifier, plus the `forbidden_parts`
    vocabulary
  * the ordered PASS/FAIL **assertion names and outcomes** of `--self-test` and `--self-test-integrity`, the
    ordered stream bytes and exit codes (live-count lines as name + outcome only, per C-5), and the unknown-mode
    exit code
  * **new inline TOML goldens**:
    * a reopened table (`[a]` … `[a.b]` … `[a]` extension where legal)
    * a 3-segment dotted key (`a.b.c = …`)
    * an array of tables (`[[t]]` ×2 with nested keys) **whose elements list their keys in different orders**
    * inline tables inside an array, again with differing key orders **and differing key sets**
    * **nested arrays of tables** (`[[t]]` / `[[t.u]]` repeated) plus a `[t.sub]` sub-table
    * a quoted key containing `.`
    * a **leading UTF-8 BOM** (Python: fail-closed reject; C-6 BOM parity)

  ACs: the M1-T2 generation discipline; the repo-mode verdict and selected set are recorded as **evidence
  only**. The repo mode is also captured from a **subdirectory** working directory, and by absolute path from
  **inside another repository** (C-3 root anchoring).
* **M2-T2.** Port `split_camel_acronym`, `split_identifier`, `segment_whole` and `_segment_whole_exact`, with
  the `_VOCAB_WORDS` order exactly, on `pysem` (`Lower`, `IsUpper`, `IsWord` and `WordBoundary`).
  * AC: 100% golden agreement.
  * AC: Unicode identifier tests (`İ`-containing and full-width) match Python.
* **M2-T3.** Port `window_matches` and `matches_forbidden_{sequence,concat,parts}`, with `forbidden_parts`
  verbatim.
  * AC: a table test gives each model a hit and a near-miss.
* **M2-T4.** Port `scan_go`, including the `mask=False` self-test path, reading through `pysem.ReadText`.
  * AC: per-fixture findings are identical.
  * AC: line numbering follows the Python splitting actually used (`splitlines` vs `split('\n')` is decided per
    call site by reading the source, and recorded in a code comment).
* **M2-T5.** Port `decompose_toml_key`, `compose_toml_parts`, `walk_toml_value` and the primary scan.
  * Walk the decoded tree **recursively** (C-6). Use `MetaData.Keys()` only for sibling order, with
    per-element order recovered by counting header occurrences.
  * Reject a leading U+FEFF before decoding (BOM parity).
  * ACs:
    * Findings match the `tomllib` goldens **in order** for every fixture and every new inline case,
      including the differing-key-order array elements. Parse-error prose is exempt (ED-4).
    * `retired-malformed-unparseable.toml` and the BOM case are fail-closed rejects with exit 1.
    * **ED-6 is unconditional.** Add ED-6 goldens (for example a newline inside an inline table, or the `\e`
      escape), recorded as Go-accepts / Python-rejects.
    * A live check confirms that no **tracked** `.toml` or `config.toml.example` uses 1.1-only syntax. A hit
      is a stop condition.
    * Record in `m2.md` the module-source citation showing that 1.1 syntax is accepted, and that no
      environment variable toggles it.
* **M2-T6.** Port `strip_toml_comment`, `scan_toml_with_fallback` (including the 015.009-T pre-line multiline
  state and the unterminated-at-EOF fail-closed finding) and `self_test_engines_for_name`.
  * AC: fallback findings are identical to the fallback goldens.
  * AC: the dual-engine agreement check passes on every fixture and every new inline case (a disagreement on a
    new case is recorded as a finding and is a stop condition, not a silent skip).
* **M2-T7.** Port:
  * `should_scan_repo_path` and `engine_for_path`
  * `scan_path`: an unmapped extension gives a synthetic finding, never a panic or silent return
  * `select_repo_paths`: literal pathspec arguments stay in the function body, and git runs through the
    injectable `gitRunner`

  ACs:
  * The `cmd/` includes-tests and `internal/` excludes-tests asymmetry is asserted.
  * Fail-closed: git error → exit 1; read error → a synthetic finding.
  * The selected set equals the parent-commit Python set (evidence).
* **M2-T8.** Pin (D-6):
  * Resolve the repo root with `git rev-parse --show-toplevel`. Parse
    `tools/gatecheck/internal/retiredarch/select.go` from disk with `go/parser`.
  * Locate the `selectRepoPaths` and `shouldScanRepoPath` `*ast.FuncDecl` bodies.
  * Require each pathspec literal (`config.toml.example`, `cmd/**`, `internal/**`) and each prefix literal
    (`cmd/`, `internal/`) to appear as a `*ast.BasicLit` whose **`strconv.Unquote`d value** is exact, located
    **inside** the corresponding body.
  * A read error, parse error or missing function **fails closed**.

  ACs:
  * **Mutation tests** run over `t.TempDir()` copies of `select.go`, **not** the tree. Each is red:
    * `"cmd/**"` removed
    * `"cmd/**"` moved to a package-level constant
    * the literal rewritten as a raw string with a different value
    * the function renamed
    * the file missing
  * The unmodified copy is green.
  * No self-comparison design (H-11).
* **M2-T9.** Port `expected_internal_repo_paths` and `run_repo_selection_self_test`: every assertion, in
  order, with its name.
  * AC: assertion names, order and outcomes match the goldens (the counts are evidence only).
  * AC: an **added** assertion (ED-5-exempt, as test-only): no selected path starts with `tools/`. This lives
    in `selftest_selection_test.go`, **not** in the self-test output, so the ordered bytes are unchanged.
* **M2-T10.** Port `load_fixture_manifest`, `run_fixture_self_test` and the `repo`, `self-test` and
  `self-test-integrity` modes as `retiredarch.Run`, tested directly.
  * AC: full ordered stream bytes and exit codes match the goldens (live-count lines as name + outcome only).
* **M2-T11.** Replace the `retired-arch` stub registration with `retiredarch.Run`. Switch the wrapper to the
  runner with the C-3 trap set.
  * AC (evidence): parent vs head identity for repo, `--self-test`, `--self-test-integrity` and a bogus mode,
    except ED-3/ED-5.
  * AC: the live selected-set diff is empty and live finding parity holds.
  * AC: the `RETIRED_ARCH_GATE_ADVISORY` step is untouched.
  * AC: the wrapper contains no `python`.
  * AC: `scripts/lib/gomask.py` now has **no remaining importer** (`git grep -n 'gomask' -- scripts/*.sh`
    shows only comments). That ends the M1 bridge.

## 6. Unit M3 — Port the unignore-regression and merge-strategy engines (feature, shipment; blocked by M1)

**Scope.**

* Port the unignore heredoc and the merge-strategy `evaluate_json` heredoc.
* Merge-strategy orchestration (`gh api`, `evaluate_response`, `verdict_exit_code`, the self-test table)
  **stays in bash**, so no network or credential code enters Go (PD-3).
* Add SHA-pinned `setup-go` to the two jobs that lack it (F-6), as **config tasks that precede** the wrapper
  switches.
* Fold in 150364D2.

**The 035-F defect is ported faithfully.** Today `root_gitignore_text_at` degrades an invalid ref to an empty
baseline. The port keeps that behaviour and the re-planned 035-F fixes it after M4.

**Git test isolation (every M3 test that runs `git`):**

* environment `GIT_CONFIG_GLOBAL=<os.DevNull>`, `GIT_CONFIG_NOSYSTEM=1`, `GIT_TERMINAL_PROMPT=0`, with
  `GIT_DIR`, `GIT_WORK_TREE`, `GIT_INDEX_FILE`, `GIT_COMMON_DIR` and `GIT_OBJECT_DIRECTORY` **removed** from the
  child environment. A hook-invoked `go test` therefore cannot touch the real repository.
* arguments `-c user.name=gatecheck -c user.email=gatecheck@invalid -c core.autocrlf=false -c init.defaultBranch=main`
* repositories only under `t.TempDir()`
* `t.Skip` only when `exec.LookPath("git")` fails

| ID | Task | Domain | Files | Posture | Size | Cx |
|---|---|---|---|---|---|---|
| M3-T1 | Capture unignore goldens from Python | test | `internal/unignore/testdata/unignore_golden.json`, `m3.md` | characterization-first | S | low |
| M3-T2 | Capture merge-strategy evaluator goldens from Python | test | `internal/mergestrategy/testdata/mergestrategy_golden.json`, `m3.md` | characterization-first | S | low |
| M3-T3 | Port the unignore git, baseline and scratch helpers | code | `unignore/git.go`, `git_test.go` | test-first | M | medium |
| M3-T4 | Port unignore check-ignore single, batch and verbose parsing | code | `unignore/checkignore.go`, `checkignore_test.go` | test-first | S | medium |
| M3-T5 | Port the unignore denylist, differential and ls-files checks | code | `unignore/checks.go`, `checks_test.go` | test-first (goldens) | M | medium |
| M3-T6 | Port the unignore self-test scenarios and `Run` | code | `unignore/unignore.go`, `unignore_test.go` | test-first (goldens) | M | medium |
| M3-T7 | Switch the unignore wrapper to Go (and wire the `unignore` registration), with parity evidence | code | `scripts/check-unignore-regression.sh`, `tools/gatecheck/register_unignore.go`, `m3.md` | migration-first | S | medium |
| M3-T8 | Add pinned `setup-go` to the `gitignore-append-only` job | config | `.github/workflows/ci.yml` | config | S | low |
| M3-T9 | Port the merge-strategy JSON evaluator | code | `mergestrategy/evaluate.go`, `evaluate_test.go` | test-first (goldens) | S | medium |
| M3-T10 | Add pinned `setup-go` to the `merge-strategy` job | config | `.github/workflows/ci.yml` | config | S | low |
| M3-T11 | Switch the merge-strategy evaluator to Go (and wire its registration), with parity evidence | code | `scripts/check-merge-strategy.sh`, `tools/gatecheck/register_merge_strategy_evaluate.go`, `m3.md` | migration-first | S | medium |
| M3-T12 | Fold 150364D2: trap, exit table and rev-parse guard | code | `scripts/check-merge-strategy.sh`, `m3.md` | test-first | S | low |
| M3-T13 | Update the merge-strategy gate doc for the Go evaluator and ED-1 | docs | `docs/merge-strategy-gate.md` | docs | S | low |

**Acceptance criteria.**

* **M3-T1.** For every `do_self_test_scenarios` scenario, capture its name, verdict, exit code and ordered
  stream bytes. Also capture the landing-precondition result and the exit code for an unknown flag.
  * AC: the M1-T2 generation discipline.
* **M3-T2.** Capture the evaluator output and exit code for every `scripts/testdata/mergestrategy/*` fixture,
  **by path and by stdin transport** (`evaluate_response`). Add inline cases:
  * empty input and whitespace-only input (including `\xa0` / `\u3000`, which Python's `strip()` treats as
    whitespace)
  * a non-object top level
  * a missing key
  * a **case-variant key** (`Allow_Squash_Merge`)
  * **duplicate keys** (last one wins)
  * `null`, a string, an integer, a float, a list or an object as the value
  * the `NaN` literal (ED-8)
  * **trailing data** after a valid object (e.g. `{…false…false} x`, and two concatenated objects). Python
    gives SKIP ("Extra data"), and it must **never** become PASS
  * **invalid UTF-8** bytes (Python: `UnicodeDecodeError`, exit 1)
  * both-true, one-true, both-false
* **M3-T3.** Port `git()` (through an injectable runner, `pysem.GitText`), `root_gitignore_text_at` and
  `make_scratch_gitignore`.
  * `root_gitignore_text_at` is ported **faithfully**, including the invalid-ref degradation, which is
    annotated `// 035-F: fixed after M4`.
  * Scratch directories go under `os.MkdirTemp("", …)` (tests: `t.TempDir()`), never inside the repo.

  ACs:
  * Tests cover `HEAD` read from disk, a valid ref via `git show`, and an **invalid ref**. The invalid-ref test
    pins the degradation as characterization.
  * Fail-closed: any git error other than the characterized degradation → exit 1.
* **M3-T4.** Port `is_ignored`, `is_ignored_batch` (stdin batching; a line-count mismatch is a hard error) and
  `parse_check_ignore_verbose_line`.
  * AC: the verbose-parser table covers `!` negation and quoted paths.
  * AC: a batch mismatch → exit 1 with the Python message.
  * AC: no default-buffer `bufio.Scanner` (C-4).
* **M3-T5.** Port `run_denylist_check`, `run_differential_check` (including tracked-and-un-ignored via
  `git diff --name-only`) and `run_ls_files_check`.
  * AC: scenario goldens match, and `::error::` lines are byte-identical.
* **M3-T6.** Port `do_self_test_landing_precondition`, `make_scenario_repo`, `run_scenario` and
  `do_self_test_scenarios` as `unignore.Run`, tested directly.
  * AC: every scenario's name, verdict and exit code, and the ordered stream bytes, match.
  * AC: scenario repos exist only under temp directories.
* **M3-T7.** Replace the `unignore` stub registration. Switch the wrapper to the runner with the C-3 trap set.
  * AC (evidence): parent vs head identity for `--self-test` and for the differential run against the PR base,
    except ED-3.
  * AC: the wrapper contains no `python`.
  * **Precondition:** M3-T8 is on the same branch **before** this commit.
* **M3-T8.** In job `gitignore-append-only`, insert a step **named exactly `Set up Go`** (the name the M4-T2
  test looks for):
  `uses: actions/setup-go@d35c59abb061a4a6fb18e82ac0862c26744d6ab5 # v5.5.0` with `go-version: '1.26.x'` (the
  same pin as `lint` → "Set up Go"). Put it **before** step "Self-test the un-ignore regression checker's own
  logic", and add one LOCAL DIVERGENCE header line.
  * ACs:
    * `actionlint` is clean.
    * The YAML parse check is `actionlint` itself, which parses the workflow, plus a reviewed `git diff`
      hunk. No Python YAML check is introduced. Both outputs are recorded.
    * `ci-gate.needs` is unchanged.
    * The job's step order is asserted in `m3.md`.
* **M3-T9.** Port `evaluate_json` as `gatecheck merge-strategy-evaluate <path|->`.
  * Read input through `pysem.ReadText`: invalid UTF-8 → exit 1 (ED-2), matching Python's traceback exit 1.
  * Decode into an **exact-key `map[string]any`** with `json.Decoder.UseNumber()`. Never decode into a struct,
    because struct decoding matches keys case-insensitively.
  * **After the first value, a second `Decode` must return `io.EOF`.** Anything else (trailing data or a
    second value) takes the parse-error SKIP path, matching Python's "Extra data".
  * Empty-input detection uses `pysem.Strip`.
  * Python type names in the not-boolean reason (`NoneType`, `str`, `int`, `float`, `list`, `dict`) are
    derived per value; a `json.Number` containing `.`, `e` or `E` → `float`, otherwise `int`.
  * The path argument is used as given (the file the bash caller wrote). `-` reads stdin. A read error → exit
    1 with `::error::` (ED-2).
  * ACs:
    * Every M3-T2 golden matches, except the ED-8 reason text.
    * The case-variant key → SKIP (absent). Duplicate keys → last wins.
    * **Trailing data or a second value → SKIP, never PASS**. This is a named red test, committed first.
    * `true` for either key → FAIL; both `false` → PASS; any non-bool → SKIP, never PASS.
* **M3-T10.** Same as M3-T8, including the step name `Set up Go`, for job `merge-strategy`, before step
  "Self-test the checker's own logic (regression guard)".
* **M3-T11.** Replace the `merge-strategy-evaluate` stub registration. Replace the heredoc in `evaluate_json()`
  with `gatecheck_invoke merge-strategy-evaluate "$src"` and drop the Python probe. Call `gatecheck_build`
  **once** at script start; `evaluate_json` stays callable inside `$(… || true)`.
  * AC (evidence): `--self-test` passes; verdict parity on every fixture **by path and by stdin**; a live run
    under `GITHUB_TOKEN` still reports SKIP (the documented behaviour that 124AE9DE and C8914513 rely on).
  * **Precondition:** M3-T10 is on the branch first.
* **M3-T12.** Fold 150364D2:
  * (1) the `evaluate_response` temp file is removed by the script's single composed `cleanup()` trap
  * (2) the self-test table asserts PASS→0, SKIP→0 and FAIL→1, alongside the existing unrecognized→2
  * (3) guard `git rev-parse --show-toplevel`, so a run outside a checkout exits 2 with a message (ED-1). The
    guard runs **before** `gatecheck_build`, so no build is attempted outside a checkout.

  ACs:
  * The new self-test rows are committed red-first (the verdict_exit_code mapping is broken in a scratch copy
    to show them failing; the proof is recorded).
  * ED-1 is demonstrated from a non-git `mktemp -d` directory.
  * The evidence section is separate from M3-T11's.
* **M3-T13.** Update `docs/merge-strategy-gate.md`:
  * the exit-2 wording (ED-1)
  * the evaluator description (Go)
  * the manual-probe snippet that uses `python -c` becomes a `jq` equivalent, or is explicitly marked as
    historical evidence

  AC: `git grep -n -i python -- docs/merge-strategy-gate.md` is empty or shows only the marked historical
  line.

## 7. Unit M4 — Retire gate-engine Python, CI steps and docs (feature, shipment; blocked by M2 and M3)

**Safety mode: careful mode** for M4-T5 and M4-T6, which remove CI steps and delete files (Principle VIII).

**Operator approval checkpoint (Principle VII).** Before any ci.yml step removal or `git rm`, Ship presents the
operator with:

* the M4-T1 coverage map
* the M4-T2 green run
* the M4-T3 red run
* the §13 rollback procedure

Ship records the operator's explicit approval in `m4.md` (M4-T4). Without that approval, M4-T5 and later tasks
do not start.

| ID | Task | Domain | Files | Posture | Size | Cx |
|---|---|---|---|---|---|---|
| M4-T1 | Map every Python unittest assertion to its Go equivalent | docs | `m4.md` (coverage map) | characterization-first | S | low |
| M4-T2 | Port the retained CI-wiring assertions to a Go test | test | `tools/gatecheck/ciwiring_test.go` | test-first (mutated inputs) | M | medium |
| M4-T3 | Add the Python-retirement assertions (live red, before retirement) | test | `tools/gatecheck/ciwiring_retire_test.go` | test-first (live red) | S | low |
| M4-T4 | Obtain and record operator approval for the retirement | docs | `m4.md` | operator checkpoint | XS | trivial |
| M4-T5 | Remove the lint-job Python steps and rewrite the divergence header | config | `.github/workflows/ci.yml` | migration-first | S | medium |
| M4-T6 | Delete the gate-engine Python modules and Python tests | code | `scripts/lib/gomask.py`, `scripts/lib/retired_arch.py`, `scripts/lib/tests/` (deletions; one skill domain: retiring one Python component and its own tests) | migration-first | S | low |
| M4-T7 | Update the write-path and retired-arch wrapper headers | docs | `scripts/check-write-path-precondition.sh`, `scripts/check-retired-architecture.sh` (comments only) | docs | XS | trivial |
| M4-T8 | Update the unignore and merge-strategy wrapper headers; check for dangling references | docs | `scripts/check-unignore-regression.sh`, `scripts/check-merge-strategy.sh` (comments only) | docs | XS | low |
| M4-T9 | Record the consolidated before/after parity evidence | docs | `m4.md` (final section) | verification | S | low |

* **M4-T1.** A table maps each `def test_*` in the 5 Python files to either:
  * (a) the Go test that covers it, or
  * (b) a retirement rationale. The only valid rationale is "asserts a Python-only property".

  AC: 0 rows unmapped. Any (a) row whose Go test is missing blocks M4-T4 (the approval) and therefore every
  removal.
* **M4-T2.** Port the `test_ci_wiring.py` assertions that **stay true** after retirement. The test reads `ci.yml`
  as text (no YAML library) and locates steps by their `- name:` text, **never by line number**. It asserts:
  * each gate step runs its wrapper
  * the advisory `continue-on-error` expressions are intact
  * the self-test and self-test-integrity steps have no `continue-on-error`
  * a step named `Set up Go` precedes the gate steps in `lint`, `gitignore-append-only` and `merge-strategy`
  * the `topology-check` job's SHA-pinned `setup-python` and its two `--require-hashes` pip steps are present

  Each assertion is a function over the text. ACs:
  * Unit cases feed **mutated copies** (a step removed, steps reordered, `continue-on-error` added, the pin
    changed), and each one is red.
  * The live `ci.yml` is green.
* **M4-T3.** Add three assertions:
  * (i) outside job `topology-check`, no line matches `^\s*(-\s*)?uses:\s*actions/setup-python@`
  * (ii) outside `topology-check`, no `run:` block line matches `(^|[\s;&|(])python3?(\s|$)`
  * (iii) `git ls-files -- 'scripts/*.py' 'tools/*.py'` is empty

  Pattern rules:
  * The patterns are compiled from **concatenated fragments** (for example `"setup-" + "python"`), so the test
    file cannot match itself (compound 2026-09-06).
  * The pathspec deliberately has **no `:(glob)` magic**. Without it, git's `*` matches across `/`, so
    `scripts/*.py` covers both `scripts/x.py` and `scripts/lib/tests/y.py`. The form `scripts/**/*.py` would
    miss the top-level file. A comment in the test records this.
  * Every `git` call, both the live check and the temp-repo cases, uses the §6 git-isolation environment. It
    also **removes** `GIT_LITERAL_PATHSPECS`, `GIT_GLOB_PATHSPECS`, `GIT_NOGLOB_PATHSPECS` and
    `GIT_ICASE_PATHSPECS`, because any of these could make (iii) pass while seeing nothing.

  ACs:
  * Mutated-input cases are red: a setup-python step in `lint`, a `run: python3 x`, and a fake **top-level**
    `scripts/x.py` plus a nested `scripts/a/b.py` in a temp-repo `git ls-files`.
  * **Principle II red:** on the live tree before M4-T5 and M4-T6, (i), (ii) and (iii) fail. The failing
    output is recorded in `m4.md`.
  * The branch is **not pushed** while these assertions are expected-red (M4-T3 through M4-T6), so the
    pre-push quality gates are not bypassed.
* **M4-T4.** Ship broadcasts the approval request (agent-intercom) with the four artifacts listed above. The
  branch is unpushed at this point, so the broadcast also **attaches the review material inline**:
  * `git --no-pager diff --stat <M4-parent>..HEAD`
  * the planned ci.yml step-removal hunk
  * the exact list of paths to delete

  Ship then records the approver, the timestamp and the approved commit SHA.
  * AC: `m4.md` records the approval before any removal commit exists.
* **M4-T5.** Remove the `lint` steps "Set up Python (gate engines)" and "Python gate-engine lint and unit tests".
  Rewrite the ci.yml header's LOCAL DIVERGENCE paragraph that describes them. This task runs **before** the
  file deletions, so no intermediate commit runs deleted tests.
  * AC: assertions (i) and (ii) of M4-T3 turn green.
  * AC: the `topology-check` job's lines are **byte-identical** (checked on the diff hunks).
  * AC: `actionlint` is clean.
  * AC: `ci-gate.needs` is unchanged.
* **M4-T6.** `git rm` the three paths.
  * AC: M4-T3 assertion (iii) turns green, and the whole CI-wiring test is green.
  * AC: `go build ./...`, `go test ./...` and all four wrappers' self-tests are green.
  * AC: `scripts/deploy-harness.sh` is unchanged.
* **M4-T7.** Rewrite the header comments of the write-path and retired-arch wrappers that describe Python
  engines or a Python prerequisite.
  * AC: `git grep -n -i -E '\bpython3?\b|heredoc' -- scripts/check-write-path-precondition.sh scripts/check-retired-architecture.sh`
    is empty.
* **M4-T8.** Rewrite the header comments of the unignore and merge-strategy wrappers the same way.
  * `docs/merge-strategy-gate.md` was already updated in M3-T13.
  * `scripts/deploy-harness.sh` (the autoharness prerequisite) and `scripts/pre-push-quality-gates.sh` (a
    generic comment) are **out of scope** and stay unchanged.
  * AC: `git grep -n -i -E '\bpython3?\b|heredoc' -- 'scripts/check-*.sh'` is empty.
  * AC: a **dangling-reference check** returns nothing:
    `git grep -n -E 'gomask\.py|retired_arch\.py|scripts/lib/tests' -- . ':!docs/archive' ':!docs/memory' ':!docs/closure' ':!docs/decisions' ':!docs/plans' ':!docs/compound' ':!docs/bugs' ':!.backlogit'`
* **M4-T9.** A consolidated evidence table covering every gate and mode:
  * **(a) per-switch pairs:** each switch task's parent vs its head
  * **(b) end to end:** the **M1 parent vs the M4 head**, with ED-5 rows subtracted explicitly

  ACs:
  * Every non-identity maps to ED-1…ED-9.
  * No non-identity in (b) is missing from the union of (a).
  * The CI run URLs of each unit's PR, and of the first post-merge `main` runs, are recorded.

## 8. Dependency Graph

```text
029-S (closure PR #78) ──blocks──► M1 ─┬─blocks─► M2 ─┐
                                       └─blocks─► M3 ─┴─blocks─► M4 ──blocks──► 030-S, 031-S, 032-S (re-plan, then claim)
033-S (docs) — independent, unchanged
```

Within each unit, an edge "A→B" means A blocks B.

**M1:**

| From | To |
|---|---|
| T1 | T2, T3 |
| T2 | T5a, T5b, T7 |
| T5a | T6a, T6b |
| T6a | T7 |
| T6b | T8 |
| T3 | T8 |
| T4 | T9 |
| T7 | T8 |
| T8 | T10 |
| T9 | T10 |
| T5b | T10 |

**M2:**

| From | To |
|---|---|
| T1 | T2, T5 |
| T2 | T3 |
| T3 | T4 |
| T4 | T7 |
| T5 | T6 |
| T6 | T7 |
| T7 | T8 |
| T8 | T9 |
| T9 | T10 |
| T10 | T11 |

**M3:**

| From | To |
|---|---|
| T1 | T3 |
| T3 | T4 |
| T4 | T5 |
| T5 | T6 |
| T6 | T7 |
| T8 | T7 |
| T2 | T9 |
| T9 | T11 |
| T10 | T11 |
| T11 | T12 |
| T12 | T13 |

**M4:** a single chain, T1 → T2 → T3 → T4 → T5 → T6 → T7 → T8 → T9. M4-T8's repository-wide greps also cover
the files that M4-T7 edits, so T8 must follow T7.

**Held-shipment gating decision (PD-5).** 030-S, 031-S and 032-S are gated on **M4**, not on per-gate
predecessors (032-S on M3, 030-S on M2, 031-S on M1). There are three reasons:

1. **Evidence integrity.** M4-T9 (b) compares the M1 parent with the M4 head end to end. If a held unit's intentional verdict
   change (for example 035-F's fail-closed baseline) landed between M1 and M4, it would appear as an
   un-enumerated delta and trip the C-8 stop condition.
2. **Rollback stays verdict-neutral.** Until M4 merges, reverting a wrapper switch restores a Python engine
   that is still in the tree. If a held unit's Go-side fix were already present, that revert would silently
   drop the fix.
3. **One re-plan session.** Re-planning 033-F, 034-F and 035-F together against the finished tool happens in a
   single Stage pass, instead of three passes against a partial tool.

The cost is a delay of 032-S by the time M2 and M4 take. That cost is accepted. Operator question Q-4 offers
per-gate gating as the alternative.

## 9. Decisions and Rationale

See deliberation D-1…D-8. The plan-level decisions are:

* **PD-1: Goldens are JSON in the package's `testdata/`.** They are neither `.go` nor gofmt-scanned. Edge
  cases stored inline avoid malformed `.go` files.
* **PD-2: One shared runner split into build and invoke** (C-3). This concentrates the F-1 hazard in one
  place with tests, and makes command-substitution call sites safe.
* **PD-3: Merge-strategy orchestration stays in bash.** No token or network code moves into Go.
* **PD-4: Both scope expressions are ported unchanged.** 033-F owns the unification.
* **PD-5: Held shipments are gated on M4** (see §8).
* **PD-6: Engines go under `tools/gatecheck/internal/`.** The compiler enforces import isolation, and gate
  selection is root-anchored (C-1).
* **PD-7: Merge-commit-only.** Each unit ships as one PR merged with a merge commit (P-009 / Constitution XI).
  Ship engages Copilot code review before any merge (P-018).

## 10. Risks and Caveats

| Risk | Mitigation |
|---|---|
| Python/Go semantic drift (Unicode classes, case mapping, `repr`, newline translation, lookaround) | the `pysem` package with Python-generated goldens; byte-identical engine goldens; live hash parity at each switch |
| `go run` collapsing exit codes; trap clobbering; buffered output | the C-3 runner split and the M1-T9 red/green tests |
| Map-order nondeterminism | C-6 recursive document-order walk; `-shuffle=on` runs |
| Golden volatility from product edits | fixture-scoped goldens; the live corpus is evidence only |
| Windows advisory `go test ./...` with git- or bash-dependent tests | temp-directory scratch only; skip only when `git` or `bash` is absent; `filepath` everywhere; the advisory job stays non-blocking |
| Setup-go time added to two jobs | accepted; the same pin as `lint`; the module cache comes from setup-go |
| Autoharness regeneration drops divergences | LOCAL DIVERGENCE lines in the same diff (M3-T8, M3-T10, M4-T5); M4-T2/T3 pin them |
| TOML 1.0 vs 1.1 acceptance | the ED-6 verification plus a stop condition on tracked files |

## 11. Hardening Signals

| Signal | Present? | Justification |
|---|---|---|
| Public API, schema or contract change | **present** | gate exit and verdict contracts; CI step topology in an autoharness-generated file |
| Security, auth, permission, compliance | **present** | the write-path gate enforces the NON-NEGOTIABLE containment boundary; the merge-strategy gate verifies P-009 |
| Migration, destructive or irreversible | **present** | engine migration; M4 deletes the fallback engines and CI steps |
| External integration, operator checkpoint, external dependency | **present** | `setup-go` Actions pin; `gh api` transport; BurntSushi newly used by a gate; the M4 approval checkpoint |
| High runtime, rollout or rollback risk | **present** | a merge-blocking false positive or negative affects every PR; the unignore gate has no kill switch |

Requires plan hardening: yes

## 12. Runtime Verification and Closure

| Unit | Runtime surface | Verification before absorption | Closure artifact |
|---|---|---|---|
| M1 | `lint` write-path steps | `m1.md`; the PR's `lint` job is green with the Go engine; the first 3 `main` runs are green | `m1.md` plus the Ship post-merge closure (rollback trigger: any write-path verdict change on `main`) |
| M2 | `lint` retired-arch steps | as M1, plus the selected-set diff is empty | same |
| M3 | `gitignore-append-only` and `merge-strategy` jobs | both jobs green; merge-strategy still SKIP under `GITHUB_TOKEN`; the unignore differential passes | same |
| M4 | `lint` without Python | `lint` green without Python; `topology-check` byte-identical and behaving as before | `m4.md` consolidated plus closure |

## 13. Plan Hardening

**Hardening required: yes** (all five signals are present).

**Instructions and learnings consulted.**

* Instructions:
  * `.github/instructions/constitution.instructions.md` (II, III, IV, VI, VII, VIII, XI; Task Granularity;
    Stop Conditions)
  * `ci-security.instructions.md`
  * `technology-go.instructions.md`
  * `coding-discipline.instructions.md`
  * `git-merge.instructions.md`
  * `copilot-code-review.instructions.md`
  * `strict-safety.instructions.md`
  * `backlogit.instructions.md`
  * `backlog-integration.instructions.md`
* Compound learnings:
  * `docs/compound/2026-09-28-bash-python-heredoc-stdin-and-github-token-admin-scope.md`
  * `docs/compound/workflow-issues/staging-premise-verification-before-planning-2026-09-18.md`
  * `docs/compound/workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md`
  * `docs/compound/2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md`
  * `docs/compound/2026-09-05-pip-index-proxy-staleness-vs-real-ci.md`
  * `docs/compound/2026-09-08-verify-go-stdlib-internals-claims-against-goroot-source.md`
  * `docs/compound/2026-09-08-adversarial-review-empirical-verification-and-scope-check.md`
* Gate-reliability plan §8, H-1…H-11 (particularly H-4 on invisible assertions and H-11 on self-comparison
  pins).

**Premises verified against the tree at 1651936 during staging:**

* `go run` exit collapse is confirmed in the GOROOT go1.26.5 source (`cmd/go/internal/run/run.go:56`).
* BurntSushi/toml v1.6.0 is in `go.sum` and present in `GOMODCACHE`.
* `.golangci.yml` has only the copilot-sdk depguard rule, so no rule forbids `os/exec` or `os.Exit` in
  `tools/`.
* The gate pathspecs run from the repository root, and the prefix predicates test repo-relative paths.
* `actionlint` is available locally.
* The Python `evaluate_json` embeds `JSONDecodeError` text (ED-8).
* The write-path repo mode passes on an empty selection (ED-7).

**Protected invariants.**

* **INV-1:** No merge-blocking verdict or exit code changes, except ED-1…ED-9.
* **INV-2:** The self-test and self-test-integrity steps stay unconditionally blocking.
* **INV-3:** The advisory toggles keep their names and semantics.
* **INV-4:** `ci-gate.needs` is unchanged.
* **INV-5:** The `topology-check` job and `scripts/deploy-harness.sh` are byte-identical.
* **INV-6:** The pathspec pin stays source-text-anchored.
* **INV-7:** Engines live under `tools/gatecheck/internal/`. Nothing in `cmd/**` or `internal/**` imports
  `tools/…` (the compiler enforces this).
* **INV-8:** No write primitive is added under `cmd/**` or `internal/**`.
* **INV-9:** `go.mod` and `go.sum` are unchanged.

**Surface matrix (compound 2026-09-13).**

| Gate | Engine | Wrapper | ci.yml job → step | ci-gate.needs | Fixtures | Docs | Pin / toggle |
|---|---|---|---|---|---|---|---|
| write-path | M1-T8 | M1-T10 | `lint` → "Run write-path-precondition gate/self-test" (unchanged) | lint | `scripts/testdata/writepath` (+M1-T1) | M4-T7 | `WRITE_PATH_GATE_ADVISORY` |
| retired-arch | M2-T2…T10 | M2-T11 | `lint` → "Run retired-architecture gate/self-test" (unchanged) | lint | `retiredgo*`, `retired-*.toml`, M2-T1 inline | M4-T7 | pin M2-T8; `RETIRED_ARCH_GATE_ADVISORY` |
| unignore | M3-T3…T6 | M3-T7 | `gitignore-append-only` → the un-ignore steps (+setup-go, M3-T8) | gitignore-append-only | self-test scenarios | M4-T8 | none (a revert is the rollback) |
| merge-strategy | M3-T9 | M3-T11/T12 | `merge-strategy` → self-test and gate steps (+setup-go, M3-T10) | merge-strategy | `scripts/testdata/mergestrategy`, M3-T2 inline | M3-T13 | `MERGE_STRATEGY_GATE_REQUIRED` |
| Python unittests | M4-T1/T2/T3 | — | `lint` → the two Python steps (removed in M4-T5) | lint | — | M4-T7/T8 | — |
| **probe:** runner | M1-T9 | all | — | — | fake `go` shims | C-3 | build → 2 |
| **probe:** import isolation | C-1 | — | `lint` → golangci-lint (unchanged) | lint | M2-T9 test | — | Go `internal` rule |
| **probe:** held-shipment lock | §8 PD-5 | — | — | — | — | this plan | shipment `blocks` + feature `blocked` |

**Safety modes (Principle VIII).**

* **Freeze-scope** for M1–M3. Paths are limited to `tools/gatecheck/**`, `scripts/lib/gatecheck-run.sh`, the
  four `scripts/check-*.sh`, `scripts/testdata/writepath/**` (M1-T1 only), `.github/workflows/ci.yml` (M3-T8
  and M3-T10 only), `docs/merge-strategy-gate.md` (M3-T13 only) and `docs/plans/evidence/2026-09-28-…/**`.
* **Careful mode** for M4-T5 and M4-T6, behind the M4-T4 approval.

**ProposedAction / ActionRisk.**

| # | ProposedAction | ActionRisk | Approval | Expected result / rollback |
|---|---|---|---|---|
| PA-1 | Switch write-path to Go (M1-T10) | high | PR merge approval (P-014) | Evidence identity. **Revert the wrapper commit**; the Python engine is still present. |
| PA-2 | Switch retired-arch to Go (M2-T11) | high | P-014 | Same; `RETIRED_ARCH_GATE_ADVISORY` for the verdict step only. |
| PA-3 | Switch unignore plus setup-go (M3-T7/T8) | high (no kill switch) | P-014; the PR body **states explicitly** that there is no toggle | Revert. A PR runs its own wrapper, so a revert PR is not blocked by a regressed engine. |
| PA-4 | Switch the merge-strategy evaluator, fold 150364D2, setup-go (M3-T10…T12) | medium | P-014 | Revert. |
| PA-5 | Delete the Python and remove the lint Python steps (M4-T5/T6) | high | P-014 **plus** the M4-T4 operator approval; precondition: M4-T1 has 0 unmapped rows and M1–M3 closures are green | See the rollback window below. |
| PA-6 | ci.yml edits to an autoharness-generated file (M3-T8, M3-T10, M4-T5) | medium | P-014 | LOCAL DIVERGENCE lines in the same diff; `actionlint`; M4-T2/T3 pin them. |

**Rollback window and procedure.** A gate's rollback reverts its **switch commit**, never a whole unit merge:
M1-T10, M2-T11, M3-T7, or M3-T11 together with M3-T12. Use `git revert <switch-commit>` in a revert PR.

* **Why not the unit merge.** A unit merge also carries shared infrastructure, such as M1's runner, `main.go`,
  `pysem` and `gomask`. M2 and M3 build on that infrastructure, so reverting M1's merge after them would break
  their wrappers. The unconditionally blocking integrity steps would then block the revert PR itself.
* **Until M4 merges:** reverting a switch commit is verdict-neutral, because the Python engine is still in the
  tree. The same holds for M3's two gates, which revert independently via their own switch commits.
* **After M4 merges and before the first held-shipment merge:** restore Python in **two reverts, in order**:
  1. Revert the M4 merge (`git revert -m 1 <M4-merge>`). This restores the Python files and the CI steps. The
     wrappers still use Go, so no verdict changes.
  2. Revert the affected gate's switch commit. That wrapper returns to Python.
* **The rollback window closes** at the first post-M4 merge of 030-S, 031-S or 032-S. From then on, fix forward
  only. Ship's closure for M4 records the date.

**Added verification.**

* **Environment precheck** in every switch and capture task. Record `go version` (must be go1.26.x),
  `git --version`, `bash --version` and the Python version used for the goldens.
* **Blocked-path handling.** A golden that does not reproduce byte for byte halts the task. Nondeterminism in
  the *Python* engine is itself a finding: capture it as a P-021 deferral. Never paper over it.
* **`actionlint` for every ci.yml-touching task.** It parses the YAML and lints the workflow. Both outputs are
  recorded.
* **Stop conditions (halt → Stage):**
  * a non-enumerated delta
  * any change to a fixture's expected verdict
  * code under `cmd/**` or `internal/**`
  * any change to `go.mod` or `go.sum`, including a BurntSushi or toolchain bump
  * any change to `topology-check` job lines
  * a tracked TOML file using 1.1-only syntax (ED-6)
  * a missing M4-T4 approval

**Monitoring and validation window.** After each unit merges, Ship's closure observes the first 3 `main` runs
and the next 3 PR runs of the affected jobs. The rollback trigger is any verdict on `main` that differs from the
pre-merge verdict at the same tree. Ship owns this, with escalation to the operator.

**Observability.** Ship broadcasts (agent-intercom) at each switch task's evidence completion, at the M4-T4
approval request, and at each unit's merge-readiness.

**Harvest checklist (Stage, Step 5):**

* Every task carries a parent, ACs, size and complexity.
* Intra-unit `blocks` edges are recorded per §8.
* Shipment dependencies: M1 on 029-S; M2 and M3 on M1; M4 on M2 and M3.
* 030-S, 031-S and 032-S each depend (`blocks`) on the M4 shipment.
* 033-F, 034-F and 035-F are set to `blocked` with the reason recorded.
* **A read-back query check** confirms the lock on the three held shipments and **no change** to 033-S.
  It runs `backlogit shipment get` for 030-S, 031-S, 032-S and 033-S, and `backlogit get` for 033-F, 034-F
  and 035-F to confirm `status: blocked`.

**Unresolved operator decisions (non-blocking; the defaults are implemented).**

* Q-1: `go/ast` scope. Default: a lexical port now, with AST adoption in the re-planned 034-F.
* Q-2: TOML dependency. Default: BurntSushi v1.6.0 (already required).
* Q-3: Post-M4 held-shipment order. Default: 032-S, then 030-S, then 031-S.
* Q-4: Held-shipment gating. Default: gated on M4 (PD-5). The alternative is per-gate predecessors.

## 14. Constitution Check

| Principle | How the plan complies |
|---|---|
| I Safety-first Go | Errors are never ignored (errcheck); fail-closed ACs; no panics on untrusted input; `main` is the only `os.Exit`. |
| II Test-first | Every engine code task is test-first against Python-generated goldens, or has committed red tests (M1-T4, M1-T9, M3-T9 trailing data, M3-T12, M4-T2, M4-T3 live red). **Recorded exception:** the five wrapper switch tasks (M1-T10, M2-T11, M3-T7, M3-T11, and M3-T12's wrapper part) are `migration-first`. Their red/green checks are the existing wrapper `--self-test`/`--self-test-integrity` modes, the M1-T9 runner tests, and parent-vs-head parity evidence. A switch adds no new behaviour for a new test to specify. |
| III Workspace isolation | Engines read only paths under `--root` (the git toplevel). The pin resolves its path through `rev-parse`. |
| IV CLI containment | **Conflict record.** Agents running wrappers, `go test` (`t.TempDir()`) and the transient golden generators write ephemeral scratch data to the OS temp directory, which is outside the working tree.<br>**Justification:** the scratch is OS-managed and removed on every exit path (C-3, `t.TempDir` cleanup). It is the same class as today's Python `tempfile` use and Go's own `GOCACHE`.<br>**Rejected simpler alternative:** `TMPDIR=$ROOT/.tmp` (gitignored). It would nest scratch git repositories inside the gated working tree. There, the root `.gitignore` and the outer repository's `git` would interact with the unignore gate's scenario repos, and `./...` tooling could traverse the scratch Go files.<br>No persistent file is written outside the working tree. |
| V Observability | Evidence files per unit; broadcasts; CI URLs recorded. |
| VI Single responsibility / dependencies | No new module; BurntSushi is already required (INV-9). |
| VII Destructive approval | M4-T4 operator checkpoint before the ci.yml step removal and the `git rm`. |
| VIII Safety modes | Freeze-scope for M1–M3; careful mode for M4-T5/T6. |
| IX Git-friendly persistence | JSON goldens and markdown evidence. |
| X Context efficiency | One evidence file per unit; fixture-scoped goldens. |
| XI Merge commits | PD-7. |
| Task granularity | 45 tasks (M1 12, M2 11, M3 13, M4 9). Each has one domain, fewer than 3 changed files (evidence sections are verification output, per C-7) and at most 4 functions per `pysem` task. Sizes range XS–M, and no task is `complexity: high`. **Recorded file-count exceptions:**

* **M1-T4** creates `main.go`, `main_test.go` and four ~5-line `register_<name>.go` stub files. Putting each
  stub in its own file is what stops M2 and M3 from conflicting.
* **M1-T5a** adds the generated `digit_table.go` (data only) next to `classes.go` and `classes_test.go`.

In both cases the review surface stays bounded. |

## 15. Plan Review

### 15.1 Attempt 1 (revision 1)

<!-- plan-review-attempt: 1 -->

```text
dispatch_mode: multi-agent
decision: FAIL
personas: Constitution Reviewer, Go Reviewer, Scope Boundary Auditor, Architecture Strategist, Security Lens Reviewer, Learnings Researcher
degradations: none (all six personas returned)
```

**P1 findings (blocking):**

* **Runner design.** Split build from invoke. Chain traps instead of overwriting them. Do not buffer output.
  Pin `GOTOOLCHAIN`/`GOFLAGS`. Strip `GH_TOKEN`.
* **Width and 2-hour violations.** M1-T2, M2-T8, M3-T8 and M4-T3 each bundled domains or verification
  artifacts.
* **TOML walk.** `MetaData.Keys()` is flat and misses arrays of tables. It needs a recursive decoded walk,
  goldens for reopened tables, dotted keys and `[[arrays]]`, and a TOML 1.1-vs-1.0 delta.
* **File reads.** `read_text` newline translation was unmodelled. This needs a shared read helper plus
  file-based goldens.
* **Missing Python semantics.** `repr`, `isdigit`, `isupper`, `islower`, `lower` (full case mapping),
  `isspace` and lookahead needed a dedicated `pysem` task.
* **Import isolation.** It was unenforced. Use `internal/` or depguard, plus a test.
* **Fail-closed behaviour.** No ACs covered git or read errors, empty selection, or `bufio.Scanner` limits.
* **Test-first.** No committed red tests for the runner and wrappers.
* **Operator approval.** None before the M4 deletions (Principle VII).
* **Self-matching greps.** The M4 grep assertions would match themselves.

**P2 findings:**

* Engine exit-code semantics: SystemExit → 1, and only bash emits 2.
* Streams must be compared as ordered bytes, not line sets.
* Live-count self-test lines need handling.
* Pin mutations must run on temp copies, with `strconv.Unquote` and git-toplevel resolution.
* M4-T2 needs red mutated inputs.
* Self-test rows and fixtures that the migration adds (the ED-5 class).
* Git test isolation.
* JSON exact-key decoding, `UseNumber` and Python type names.
* Held-shipment gating decision (per-gate vs M4).
* Harvest read-back check.
* Rollback window.
* Missing Constitution Check.
* Temp-dir containment justification.
* Safety modes.
* Limit the `*.py` assertion scope.
* Anchor by step name.
* Probe rows.
* The list of consulted instructions.

**P3 findings:**

* Pre-register the stubs.
* Split the evidence files.
* State P-009 merge commits.
* Broadcasts.
* A minimal M1 bridge.
* A stop condition on dependency bumps.
* P-018 Copilot review.
* The merge-strategy path argument.
* actionlint and YAML parsing everywhere.
* An explicit M4 docs file list.

### 15.2 Attempt 2 (revision 2)

<!-- plan-review-attempt: 2 -->

```text
dispatch_mode: multi-agent
decision: FAIL
personas: Constitution Reviewer (ADVISORY), Go Reviewer (FAIL), Security Lens Reviewer (ADVISORY), Architecture Strategist + scope (ADVISORY)
degradations: Scope Boundary Auditor and Learnings Researcher were folded into the Architecture Strategist dispatch for the delta review (declared, P-012)
attempt-1 findings: all CLOSED except IV conflict record (partial), pathspec fix (misapplied), rollback procedure (inconsistent)
```

**P1 findings (Go Reviewer):**

* The pathspec `'scripts/**/*.py'` without `:(glob)` misses top-level files (M4-T4/T6).
* `json.Decoder.Decode` ignores trailing data, so input Python treats as SKIP becomes PASS: a fail-open verdict
  change (M3-T9).

**P2 findings:**

* `MetaData.Keys()` does not index array-of-tables elements, so per-element order was unspecified.
* BurntSushi strips a UTF-8 BOM, while `tomllib` fails closed. This was unlisted.
* ED-6 is certain: v1.6.0 is TOML 1.1.
* `unicode.SpecialCase` cannot express multi-rune lowering or Final_Sigma.
* Unicode 15.0 (Go) vs 16.0 (Python) version skew, and `isdigit` needs a Numeric_Type table.
* The binary cannot anchor its root on its own location: build from the runner's source directory and pass
  `--root`.
* `trap … INT TERM` resumes after the signal.
* Merge-strategy invalid UTF-8 handling.
* The rollback reverted unit merges that carry shared infrastructure.
* M4-T6 ordering left no live red (Principle II).
* The switch tasks' test-first exception was unrecorded.
* M1-T5 had 9 functions.
* A missing edge, M1-T1→M1-T2.
* PD-5 reason 1 was inconsistent with M4-T8.
* The `Set up Go` step name contract was missing.
* Principle IV conflict record.

**P3 findings:**

* Recover panic → 1.
* Unset `GIT_DIR` and related variables in tests.
* Adjacent stub lines conflict between M2 and M3.
* The ED-1 guard must run before the build.
* The invoke should also strip tokens.
* Swap the M4 ci.yml and delete order.
* M4-T7 over-serialized.
* The harvest read-back should include features.
* ED-7 trace.
* Dangling-reference check.
* The 3-file limit (M2-T10, M3-T6).
* Evidence-as-domain labelling.
* M4-T4 single-domain statement.

### 15.3 Revision-2 response map (task IDs use revision-2 numbering; see §15.4 for revision 3)

| Finding | Where addressed |
|---|---|
| Runner design | C-3; M1-T9 red/green tests |
| Width / 2-hour splits | M1-T2/T3, M1-T5/T6 (`pysem`), M2-T8/T9, M3-T7/T8, M3-T10/T11/T12/T13, M4-T3/T4/T5/T6/T7 |
| TOML walk and 1.1 delta | C-6; M2-T1 inline goldens; M2-T5; ED-6 |
| Read normalization | C-4 `ReadText`/`GitText`; M1-T3 file-based goldens |
| `pysem` | C-4; M1-T5/T6 |
| Import isolation | C-1 (`tools/gatecheck/internal/`, Go `internal` rule); M2-T9 `tools/` exclusion test; INV-7 |
| Fail-closed | M1-T8, M2-T7, M3-T3, M3-T4, M3-T9 ACs; ED-2, ED-7; C-4 no default Scanner |
| Test-first reds | M1-T4, M1-T9, M3-T12, M4-T2, M4-T6 |
| Operator approval | M4-T3; §7 header; PA-5 |
| Self-matching greps | M4-T6 fragment-concatenated patterns; `docs/compound/**` is never in scope |
| Exit semantics | C-2 |
| Ordered bytes; live-count lines | C-5 |
| Pin temp-copy mutations | M2-T8 |
| ED-5 / ED-6 / ED-7 / ED-8 | C-8 |
| Git isolation | §6 header |
| JSON evaluator | M3-T2, M3-T9 |
| Gating decision | §8 PD-5; Q-4 |
| Harvest read-back | §13 harvest checklist |
| Rollback window | §13 |
| Constitution Check | §14 |
| Temp-dir containment | C-3 |
| Safety modes | §13; §7 |
| `*.py` scope | M4-T4/T6 (`scripts/**`, `tools/**`) |
| Step-name anchoring | §1; §13 surface matrix; M4-T2 |
| Probe rows | §13 surface matrix |
| Instructions consulted | §13 |
| P3 items | M1-T4 stubs; C-7 per-unit files; PD-7 (P-009, P-018); §13 broadcasts; §4 bridge; §13 stop conditions; M3-T9 path; §13 actionlint; M4-T7 file list |

### 15.4 Revision-3 response map (attempt-2 findings)

| Finding | Where addressed |
|---|---|
| P1 pathspec | M4-T3 (iii): `'scripts/*.py' 'tools/*.py'` with no `:(glob)`; red cases cover a top-level file and a nested file |
| P1 JSON trailing data | M3-T9: a second `Decode` must return `io.EOF`, with a named red test; M3-T2 goldens; ED-8 |
| Per-element TOML order | C-6: header-occurrence counting, differing-order goldens, stop condition instead of sorting |
| BOM | C-6: BOM parity rule (reject before decoding); M2-T1 golden; M2-T5 AC |
| ED-6 certain | C-8 ED-6 is unconditional; no environment toggle (verified: no `Getenv` in v1.6.0) |
| Case mapping | C-4 `Lower`: explicit multi-rune table plus Final_Sigma; M1-T5b |
| Unicode skew and `isdigit` | C-4: version rule, generated `digit_table.go`, ED-9; M1-T2 AC |
| Root anchoring | C-3: source anchoring through `GATECHECK_SRC`, `go -C`, `--root`; M1-T3 subdirectory capture; M1-T4 `--root` parsing |
| Signal traps | C-3: `exit 130` / `exit 143` traps, idempotent cleanup; M1-T9 |
| Merge-strategy invalid UTF-8 | M3-T9 (`ReadText`); M3-T2 golden; ED-8 |
| Rollback | §13: revert switch commits, never unit merges |
| Live red (Principle II) | M4-T3 runs before M4-T5/T6 and records a live red |
| Switch-task exception | §14 II |
| `pysem` function count | M1-T5a, M1-T5b, M1-T6a, M1-T6b (at most 4 functions each) |
| Edge M1-T1→T2 | §8 M1 table |
| PD-5 reason 1 | M4-T9 (b), end-to-end comparison; §8 |
| `Set up Go` name contract | M3-T8, M3-T10, M4-T2 |
| Principle IV conflict record | §14 IV |
| P3: recover | C-2; M1-T4 |
| P3: git environment | §6 git isolation |
| P3: stub conflicts | M1-T4 per-file registration; each switch task wires its registration; §14 file-count exception |
| P3: guard before build | C-3; M3-T12 |
| P3: invoke token strip | C-3 `gatecheck_invoke`; M1-T9 |
| P3: M4 order and serialization | §7: T5 (ci.yml) before T6 (delete); §8: T7 and T8 each depend only on T6 |
| P3: features in the read-back | §13 harvest checklist |
| P3: ED-7 trace | §2 R-14 |
| P3: dangling references | M4-T8 AC |
| P3: 3-file limit | M2-T10, M3-T6 (2 files each) |
| P3: evidence labelling | C-7 |
| P3: M4 single domain | M4-T6 files column |

### 15.5 Attempt 3 (revision 3) and revision 3.1 remediation

<!-- plan-review-attempt: 3 -->

```text
dispatch_mode: multi-agent
decision: ADVISORY
personas: Go Reviewer (ADVISORY; all attempt-2 P1s CLOSED), Architecture Strategist (ADVISORY; cross-reference sweep clean, graph acyclic)
degradations: focused delta review (the personas that returned FAIL or had open items). Constitution Reviewer and Security Lens Reviewer were ADVISORY in attempt 2 and had no open P1s; their attempt-2 items are closed in §15.4 (declared, P-012)
```

**Advisory findings.** Every one was remediated in revision 3.1. There was no re-dispatch, because the re-entry
cap has been reached; each remediation is a textual precision fix.

| Finding | Severity | Remediation (revision 3.1) |
|---|---|---|
| The TOML order algorithm fails for inline-table arrays and nested `[[t.u]]` | P2 | C-6: a lockstep cursor where each element consumes its own recursively counted entries; nested counts restart per parent; goldens for differing key sets and nested arrays of tables (M2-T1) |
| The retired-arch root is module-anchored today, not cwd-anchored | P2 | C-3: per-wrapper root derivation (`git -C "$GATECHECK_SRC"` for retired-arch); M2-T1 captures a run from inside another repository |
| M4-T8 greps cover files that M4-T7 edits, while T7 and T8 were parallel | P2 | §8: M4 is one chain, T7 → T8 (this supersedes the §15.4 serialization row) |
| Pathspec environment variables could blind (iii) | P3 | M4-T3: §6 isolation, and the `GIT_*_PATHSPECS` variables are removed |
| The UTF-16 BOM wording was wrong | P3 | C-6 corrected: BurntSushi strips UTF-16 BOMs, and `ReadText` rejects that input first |
| The binary had no `GOEXE` suffix | P3 | C-3 |
| `digit_table.go` Unicode version vs ED-9 | P3 | ED-9 notes the `IsDigit` exception |
| The 3-file exception for M1-T5a | P3 | §14 records the exception |
| M2-T1 subdirectory capture | P3 | M2-T1 AC |
| The M1 T4→T8 edge was over-serialized | P3 | Removed; T4 → T9 → T10 remains |
| The operator could not inspect an unpushed commit | P3 | M4-T4 broadcast attaches the diff stat, the ci.yml hunk and the deletion list |

**Gate outcome:** ADVISORY, with every finding remediated. Stage proceeds to harvest under the Orchestrator's
standing instruction ("If the decision is clear enough: … harvest"). The operator keeps a veto at the staging-PR
review. Ship cannot claim M1 until that PR merges and 029-S closes.
# Implementation Plan — Retired-Architecture Gate Detection-Quality Hardening

- **Date**: 2026-09-08
- **Source deliberation**: `docs/decisions/2026-09-08-intercom-go-retired-arch-gate-detection-quality-deliberation.md`
- **Feature / shipment**: **015-F** / **014-S**
- **Stash entries covered (8)**: `24B63533`, `8988120D`, `6285779C`, `8D603E29`,
  `9A14D3B7`, `E403E3C5`, `09CD6ACF`, `F4F4A959`
- **Requires plan hardening**: **yes** (see §7)
- **Constitution**: Principle II (Test-First), Principle VI (one skill domain per unit),
  Principle VIII (`careful` + `freeze-scope`), Principle XI (merge commit — §8)
- **Mode**: `DARK_MODE_ACTIVE`, operator AFK, exactly **one** shipment
- **Revision 3** resolves the 5 P1 findings from the cycle-2 re-review (Correctness ×3,
  Scope ×2) plus the material P2s. See §11.
  <!-- plan-review-attempt: 3 -->
<!-- rev 4 — corrections ESC-P1-1, ESC-P2-a/b/c/d applied under P-013.6 escalation
     adjudication (route: gpt-5.6-sol/openai/high; Stage route claude-opus-5 → not
     same-route degraded). Scope Boundary returned COMPLIANT at rev 3. -->

> **Rev-2 headline correction (retained).** The gate is **already hard-blocking today**:
> `ci.yml:265-266` runs `--self-test` with no `continue-on-error`, and script L639-641 shows
> `--self-test` runs `scan_with_mode self-test` **then** `scan_with_mode repo`.
>
> **Rev-3 correction.** Rev 2 fixed that discovery but introduced a **net enforcement
> downgrade** — it made the verdict advisory and never restored it, while the operator is
> AFK. Rev 3 makes the downgrade **temporary and self-restoring**: `015.001-T` opens the
> advisory window, `015.013-T` closes it **fail-closed**. The gate ends this shipment
> blocking, exactly as it began.

---

## 1. Objective

Make `scripts/check-retired-architecture.sh` **detect what it claims to detect**, **prove**
it, and **end the shipment with enforcement no weaker than it started**.

### Anti-goals

| ID | Anti-goal | Enforced by |
|---|---|---|
| **AG-1** | No change to scan **scope / pathspec** | **AC-6** (owned by `015.011-T`) |
| **AG-2** | No structural refactor of the heredoc engine (`6C24E2E4` item 2) | **AC-7** |
| **AG-3** | **`check-write-path-precondition.sh` EXCLUDED (`BF5DE670`) — untouched** | **AC-4** |
| **AG-4** | **No net change to enforcement posture.** The advisory window is bounded by `015.001-T`…`015.013-T`; the gate ends blocking | **AC-5**, **AC-9** |
| **AG-5** | No reduction of existing coverage | **AC-8** |

> **AG-4 was rewritten in rev 3.** Rev 1/2 stated "the gate is not promoted to
> merge-blocking." That was written under the falsified premise that it was advisory. It is
> blocking today, so *preserving* enforcement — not promoting it — is the correct anti-goal.
> Deliberation **D2** is superseded accordingly (see the deliberation's rev-3 addendum).

---

## 2. Reference — verified defects at `HEAD`

| # | Defect | Location |
|---|---|---|
| V1 | Acronym backtracking: `ChannelIDs` → `['channel','i','ds']` | `component_re` L85 |
| V2 | Suffix on final token: `SocketModes` → `['socket','modes']` | `matches_forbidden_parts` L117 |
| V3 | No-separator concat: `socketmode` → `['socketmode']` | `split_identifier` L110 |
| V4 | Go struct tags masked as raw strings → invisible | `mask_go_non_code` L127 |
| V5 | Dotted / hyphenated / concatenated TOML keys evade at **three** sites | L324, L368, L377 |
| V6 | Fallback lexer skips LHS on multiline-open line | `scan_toml_with_fallback` L346 |
| V7 | Fail-open dispatch on unmapped extension | L100 → L414 |
| V8 | `cmd/` early return skips `_test.go`/`testdata` filters | L88 |
| V9 | CI: repo scan blocking via un-toggled `--self-test` step | `ci.yml:265-266` |
| V10 | `mask_go_non_code` has no EOF state check → silent fail-open *(disclosed, not fixed)* | L127-210 |

---

## 3. Decomposition — 13 units in 5 sub-epics

### 015-A — CI enforcement posture *(config + gate entrypoint)*

**015.001-T** — **Open the bounded advisory window.**
- Add an **additive** integrity-only mode (e.g. `--self-test-integrity`) that runs fixture +
  selection self-tests **without** the repo scan. **`--self-test` semantics are UNCHANGED**
  (it still runs the repo scan), so AC-1 keeps subsuming AC-2.
- **Exact ci.yml wiring** (rev-2 got this wrong):
  - Step `Run retired-architecture gate` (262-264) — the **verdict** step — carries
    `continue-on-error: ${{ vars.RETIRED_ARCH_GATE_REQUIRED != 'true' }}`.
  - Step `Run retired-architecture self-test` (265-266) — the **integrity** step — switches
    to the new integrity-only mode and carries **NO `continue-on-error`**. It stays
    **blocking**. *Applying the toggle here would make integrity failures non-blocking by
    default and violate AC-5.*
- Emit the resolved mode every run (`::notice::retired-arch gate mode=…`).
- Inline comment: match is **exact, case-sensitive `'true'`**.
- Register both steps + the toggle in ci.yml's **LOCAL DIVERGENCE re-apply list (L17-30)**.
- **Inline comments only — no references to non-existent doc paths** (`F47DB9A9` is deferred
  and must not be annexed).
- Covers `F4F4A959` + V9.

**015.013-T** — **Close the window, fail-closed.** *(last unit)*
- Invert the toggle to `continue-on-error: ${{ vars.RETIRED_ARCH_GATE_ADVISORY == 'true' }}`
  so the verdict step **defaults to BLOCKING** and any misconfiguration
  (`True`, `1`, typo, wrong scope) fails **closed**, not open — this also discharges H8.
- Update the inline rationale + the `::notice::` mapping.
- **Also update the CI "LOCAL DIVERGENCE — re-apply after regeneration" list**
  (`.github/workflows/ci.yml` L17-30) to name the retired-architecture verdict and integrity
  steps and the toggle variable, so a workflow regeneration does not silently drop them.
  *(Rev-4, ESC-P2: `015.001-T` renames the toggle variable; without this the divergence list
  and the live steps drift.)*
- **Also refresh the stale `--self-test-integrity` references** in the script header usage
  block and the bash `usage:` string.
- **Precondition**: a full repo scan at `HEAD` is clean (AC-2) after every preceding unit.
- Covers `F4F4A959` (enforcement half) + AG-4.

### 015-B — Lexer / visibility layer *(code + tests)*

**015.002-T** — Masking **bypass seam**: add a `mask=True` parameter to `scan_go`, plus a
**separate differential fixture suite** (`retiredgo-differential/`) whose **only engine is
the unmasked one** (`scan_go(mask=False)`) and whose fixtures are **all `reject`**. The
manifest stays **scalar** — no per-engine dict, no manifest schema change. Suite-to-engine
selection MUST be **suite-scoped** (keyed on the suite the fixture belongs to, not on
`engine_for_path`, which returns `'go'` for both suites). Include one fixture proving the
differential is not false-green.
- **Suite declaration (P3-1)**: the differential suite MUST still declare `engines: ['go']`
  (the pre-existing `if engine_name not in suite['engines']` guard at ~L598 is keyed on that
  list, and `engine_for_path` returns `'go'`); only the **engine function** is overridden to
  `scan_go(mask=False)`.
- **Same-input differential (P3-2)**: the differential fixture MUST carry content
  **equivalent to `retiredgo/retired-accept-comment-only.go`**, so the masked-clean /
  unmasked-reject pair is a **true same-input differential** rather than two independent
  assertions.
> **Rev-3 fix (C-P1-2).** Rev 2 said "add an unmasked engine entry to
> `self_test_engines_for_name`". `run_fixture_self_test` applies **every** engine for a suite
> to **every** fixture against **one** expectation, so an unmasked engine on the shared `go`
> suite would immediately fail the existing `retired-accept-comment-only.go` ACCEPT fixture
> (it contains `slack` in a comment) and every `015.003-T` fixture. **The existing `go` suite
> and its engine list are left untouched.**
> **Rev-4 fix (ESC-P2-a/b, escalation-adjudicated).** Rev 3's per-engine dict expectations
> required both a manifest schema change and an engine-label-keyed lookup in
> `run_fixture_self_test` (which at L593-612 understands only scalar `accept`/`reject`), and
> left the suite→engine seam unspecified — an executor would reconstitute the cycle-2 P1-B
> edit. The **unmasked-only, reject-only** suite yields the same differential evidence
> (masked path is already covered by the untouched `go` suite) with **no schema change**.

**015.003-T** — `mask_go_non_code` characterization fixtures: block comments, interpreted
strings with escaped quotes, raw backtick strings with newlines, rune literals. **Exactly 4
classes — no headroom; no class may be added during execution.** No fixture may place a
retired token in a **struct-tag position**. Covers `6285779C`.

**015.007-T** — Make Go **struct tags** visible. **Candidate tag literals MUST be captured
from `mask_go_non_code`'s `raw_string` state transitions (inside the lexer), NOT by regex
over unlexed text** — a raw-text regex also matches backticks inside comments (Go doc
comments routinely quote tags), which would re-break the `retired-accept-comment-only.go`
ACCEPT contract. Unmask only literals whose **entire** content matches
`^(\s*\w+:"[^"]*")+\s*$`; scan key/value name portions only; preserve `path:line`.
- **REJECT fixture (V4, red-first)**: field `Channel` with `toml:"channel_id"`.
- **ACCEPT fixtures**: benign `json:"team_name"` tag; a **non-tag** raw backtick literal
  (SQL/template) mentioning `channel_id`.
- Covers `24B63533`(2).

**015.009-T** — Fallback lexer: capture the multiline flags **before** `strip_toml_comment`
mutates them, and use that **pre-line** state for suppression. A line that starts *outside* a
multiline string remains eligible for LHS inspection even when it **opens** one; a line that
starts *inside* remains suppressed even when it **closes** one.
- **Fixture (valid TOML, REJECT)**: opening assignment `host_cli = """`, then content, then a
  **delimiter-only** closing line. MUST be REJECT under **both** the `tomllib` and `fallback`
  engine labels after the change, and CLEAN under the **pre-change** fallback (red-first).
- **MUST NOT** add malformed close-then-assignment TOML to the shared dual-engine suite.
> **Rev-4 fix (ESC-P1-1, escalation-adjudicated).** Rev 3 mandated a fixture where a line
> closes a multiline string *and then carries an assignment*. That construct is **invalid
> TOML** (independently reproduced: `TOMLDecodeError: Expected newline or end of document`).
> Any `scripts/testdata/retired-*.toml` is auto-discovered and run under **both** engine
> labels against **one** manifest expectation, so `tomllib` (which converts parse errors into
> findings at L335-340) yields `reject` while the corrected fallback yields `clean` — either
> expectation is **permanently red** against the unconditionally-blocking integrity step. The
> only green workaround (unterminated string at EOF, as `retired-malformed-unparseable.toml`
> does) passes for a reason unrelated to the suppression it claims to pin. The unfixtureable
> malformed-input boundary is relocated to `015.012-T` as disclosure.

Covers `8988120D`(3).

### 015-C — Token vocabulary layer *(code + tests)*

**015.005-T** — Rewrite the decomposition/matching seam for V1–V3.
- Two **separately named** models behind one dispatch point —
  `matches_forbidden_sequence` and `matches_forbidden_concat` — findings report **which
  model fired**.
- **Suffix rule (rev-3 disambiguation)**: plural `s`/`es` tolerated on the **final component
  of the MATCHED WINDOW**, not the final component of the identifier — so `TeamIDsCache`
  and `SocketModesRegistry` are still caught.
- Concat model: **whole-part decomposition, not substring search**.
- **Carries exactly ONE red-first REJECT probe** (`ChannelIDs`); the full corpus lands in
  `015.006-T`. *(Rev-3 fix S-P1-2: without this sentence an executor either pulls the whole
  corpus in — reconstituting the rev-1 2-hour breach — or lands the rewrite un-red-first.)*
- Covers `24B63533`(1) + `8988120D`(1) — **D1 merge**.

**015.006-T** — Token fixture corpus. **Every REJECT case gets its own independent fixture
file; ACCEPT cases may remain bundled in one file.**
- **REJECT (8 separate files, one case each)**: `ChannelIDs`, `TeamIDs`, `IPCNames`,
  `SocketModes`, `HostCLIs`, `socketmode`, `Socketmode`, **`TeamIDsCache`** *(non-terminal
  suffix — pins the rev-3 suffix rule)*.
- **ACCEPT (closed set, 1 bundled file)**: `HostClient`, `ChannelIdentifier`, `TeamIdentity`,
  `channelIdx`, `IPCNamespace`.
- All fixture files MUST be **`gofmt`-clean** (`gofmt -l .` walks `scripts/testdata`).
> **Rev-4 fix (ESC-P2-d, escalation-adjudicated).** Rev 3 bundled all REJECT cases into one
> file. `run_fixture_self_test` produces a **binary per-file verdict**, so a single surviving
> finding keeps the file REJECT while **seven other cases silently regress**. Bundling is
> safe only for ACCEPT (which asserts *zero* findings, so any regression flips the file).
> Splitting achieves per-case observability with **no manifest schema change** — the
> alternative (asserting a complete expected finding set) would reintroduce the schema change
> that ESC-P2-a/b just removed.

**015.008-T** — TOML key matching. One `decompose_toml_key()` used by **all three** sites.
- **Composed-path matching at ALL THREE sites.** tomllib site matches over `prefix + [key]`
  (L323). **The fallback must also match over the composed path** —
  `current_table + bare_key_re.findall(lhs)` — at **both** L368 and L377. *(Rev-3 fix
  C-P1-3: rev 2 specified composition only for tomllib; the fallback iterates keys
  individually, so a `[channel]` + `id` fixture would be REJECT under tomllib and CLEAN
  under fallback → dual-engine self-test permanently red.)*
- Hyphen normalisation lives at the **TOML boundary**. *(P3-5: in the **fallback** path
  `bare_key_re` (`[A-Za-z_][A-Za-z0-9_]*`) already splits `channel-id` into two segments, so
  hyphens never reach the decomposer there — that fixture is caught by **composed-path
  matching**, not by normalisation. Do not add dead hyphen handling to the fallback.)*
- **FP boundary**: `[channel]` + `id = …` **does** become a finding. Verified safe at HEAD
  (`[copilot].cli_path`, `[[workspace]].workspace_id`, `[database].path`,
  `[stall].default_nudge_message` all miss).
- **Enumerated fixtures — each REJECT case is its own valid TOML document (4 files); ACCEPT
  bundled (1 file)**: REJECT — `channel-id`; `channelid`; `[channel]`+`id`; `channel.id`.
  ACCEPT — `[copilot].cli_path`, `[[workspace]].workspace_id`.
  > **Rev-4 fix (ESC-P2-d).** Rev 3 bundled these into 2 files. Placing `channel.id` and a
  > later `[channel]` header in the **same** document is a **duplicate-table parse error** —
  > it would pass REJECT vacuously via the fail-closed parse-error finding rather than via
  > key matching. Separate documents also restore per-case observability under the binary
  > per-file verdict.
- **AC**: every TOML REJECT fixture is caught under **both** the `tomllib` and `fallback`
  engine labels.
- Covers `8988120D`(2).

### 015-D — Dispatch & selection guards *(code)*

**015.010-T** — Fail-closed dispatch via **synthetic finding, not `raise`** (a raise inside
`run_repo_scan` aborts mid-iteration, discards findings, prints a traceback, escapes the
toggle). Keep a `raise` only in the self-test assertion. `select_repo_paths` returns **`str`**
→ assertion must use `engine_for_path(root / p)`. Covers `9A14D3B7`.

**015.011-T** — **Selection-guard assertions** in `run_repo_selection_self_test` (one seam,
one domain). *(Rev-3 fix S-P1-1: AC-6 previously had no owning unit.)*
- **cmd/ (AG-5, D4)**: `should_scan_repo_path('cmd/x/y_test.go') is True`,
  `should_scan_repo_path('cmd/x/testdata/z.go') is True`, plus ≥1 tracked
  `cmd/**/*_test.go` in the real `select_repo_paths()`. Failure message names AG-5/D4 and
  states the assertion guards a deliberate coverage decision.
- **AC-6 pathspec pin (AG-1)**: assert the literal `git ls-files` pathspec and the
  `should_scan_repo_path` prefix set **by reading `scripts/check-retired-architecture.sh`
  from disk**, then **extracting the `select_repo_paths` and `should_scan_repo_path` function
  regions and matching the literal only inside those regions**. *A module-constant
  self-comparison would be self-referential and always green — the heredoc is fed on stdin so
  `__file__` and `inspect.getsource()` are unavailable. **Rev-4 (ESC-P2-c):** a plain
  whole-file containment check is likewise vacuous, because the asserted literal appears in
  the very file being read — the assertion would match its own source text. Region anchoring,
  not occurrence counting, is required; a bare count is brittle against unrelated edits.*
  **Region extraction MUST anchor on the FIRST occurrence of each `def <name>(` (P3-3)** —
  the anchor strings recur later in the file inside the assertion itself, so a
  last-match/`rfind` implementation would extract the assertion's own region and reintroduce
  the very vacuity this closes. Extraction failure MUST fail **closed**.
  *Known partial (P3-4): region containment pins **presence** of the pathspec literal, not
  **absence of additions** — an additive pathspec entry still passes AC-6. AG-1 coverage is
  therefore partial, with AC-8's manual before/after comparison as the compensating control.*
- Covers `8D603E29` + AG-1.

### 015-E — Test framework and disclosure *(tests / docs)*

**015.004-T** — Self-test framework invariants: each suite **non-empty**; **ordinary** suites
have **≥1 accept and ≥1 reject**. The **designated differential suite**
(`retiredgo-differential/`) is **explicitly exempt and declared reject-only** — the exemption
is a named allow-list entry, not a general escape hatch, so a future all-reject suite still
trips the invariant. Manifest expectations remain **scalar**. Sequenced 4th so it constrains
later fixture-producing units. **Assertion-additive, not red-first — exempt from AC-3.**
Covers `E403E3C5`(1),(2).

**015.012-T** *(docs)* — Disclosure, verified claims only:
- The header usage block (L14-16) attaches the `_test.go`/`testdata` exclusion only to
  `internal/**` while listing `cmd/**` unqualified — state the asymmetry. *(Rev-1's broader
  claim was wrong: L18-26 already names the third check class.)*
- Residual blind spots: aliased imports; `go.mod`/`go.sum` outside pathspec; non-`.go`/
  `.toml` under `internal/**`; **V10**; interpreted-string struct tags (H2 residual).
- **Unfixtureable malformed-input boundary (relocated from `015.009-T` by ESC-P1-1)**: a
  line that *closes* a multiline string and then carries an assignment has its LHS suppressed
  by the fallback lexer. This is **not fixtureable in the shared dual-engine suite** because
  the construct is invalid TOML, so `tomllib` fail-closes to a parse-error finding while the
  fallback reports clean — no single manifest expectation can be green under both engine
  labels. Record the behaviour and the reason it is untested rather than committing a
  permanently-red fixture.
- The `mask_go_non_code` clone in the excluded `check-write-path-precondition.sh` diverges —
  record as **"current and deliberate for this cycle"**, *not* as permanent intent, so
  deferred `6C24E2E4` keeps its decision.
- **Forward-looking guard rail (attributed as such)**: any future scope expansion must keep
  the tracked-only `git ls-files` enumeration and must not switch to `Path.rglob`/`os.walk`,
  which would pull ignored-but-present `.env.*` into a scanner whose findings echo to public
  CI logs.
- **Disclosure-only — no pathspec edits** (AG-1). Covers `09CD6ACF`, `E403E3C5`(3).

> **`E403E3C5` closes only when BOTH `015.004-T` AND `015.012-T` land.**

---

## 4. Dependencies and execution order

```text
015.001-T  (opens advisory window — MUST be first; H1/H5 unmitigated until it lands)
   └─> all units

015.002-T (bypass seam + differential suite) ─> 015.003-T ─> 015.007-T
015.002-T ─> 015.004-T

015.005-T (token seam) ─> 015.006-T (corpus)
015.005-T ─> 015.007-T   (tag fixtures are judged by the rewritten seam)
015.005-T ─> 015.008-T ─> 015.009-T   (both edit scan_toml_with_fallback)

015.010-T, 015.011-T  co-located in run_repo_selection_self_test — sequential

015.007-T, 015.010-T, 015.011-T ─> 015.012-T
ALL units ─> 015.013-T  (closes the window; requires a clean HEAD scan)
```

**Authoritative order**: `015.001-T` → `015.002-T` → `015.003-T` → `015.004-T` →
`015.005-T` → `015.006-T` → `015.007-T` → `015.008-T` → `015.009-T` → `015.010-T` →
`015.011-T` → `015.012-T` → `015.013-T`.

> "Independent" means **logically** independent, **not edit-disjoint**. Where two units share
> a function, the order above is binding.

---

## 5. Execution posture per unit

| Unit | Posture |
|---|---|
| `015.005-T` (one probe), `015.006-T`, `015.007-T`, `015.008-T`, `015.009-T` | **test-first** |
| `015.003-T` | **characterization-first** |
| `015.002-T` | seam-first (differential proven not false-green) |
| `015.001-T`, `015.004-T`, `015.010-T`, `015.011-T`, `015.012-T`, `015.013-T` | assertion-additive / documentation-verified — **exempt from AC-3** |

---

## 6. Acceptance criteria

- **AC-1**: `--self-test` (unchanged semantics, including the repo scan) passes at the end.
- **AC-2**: Full repo scan at `HEAD` reports **zero findings after every unit**.
- **AC-3**: Each of **V1–V6** has a fixture that **fails against pre-change code**.
  **V7–V10 explicitly exempt.**
- **AC-4**: No file modified outside `scripts/check-retired-architecture.sh`,
  `scripts/testdata/**`, `.github/workflows/ci.yml`, `docs/plans|decisions|closure/**`.
  **`check-write-path-precondition.sh` untouched.**
- **AC-5**: The **integrity** step fails the `lint` job unconditionally. The **verdict** step
  is governed solely by the toggle.
- **AC-6** *(AG-1)*: pathspec + prefix set pinned by a **disk-read** self-test assertion,
  owned by `015.011-T`.
- **AC-7** *(AG-2)*: no new file under `scripts/lib/`; heredoc surface stays `sys.argv[1]`.
- **AC-8** *(AG-5)*: post-change `select_repo_paths()` is a **superset** of the HEAD set,
  verified by **manual before/after comparison** (same mechanism as AC-2 — no committed
  snapshot oracle is introduced).
- **AC-9** *(AG-4)*: at shipment end the verdict step is **blocking by default** and fails
  **closed** on toggle misconfiguration.

---

## 7. Plan Hardening

**H1 — False-positive blast radius (dominant).** Mitigated by: `015.001-T` first (genuine
advisory window); AC-2 per-unit gate; enumerated ACCEPT corpora targeting exactly the
near-misses a naive implementation breaks (`HostClient`, `ChannelIdentifier`, `channelIdx`,
`IPCNamespace`); separately-named matching models so the concat model can be tuned alone.
*Residual*: AC-2 proves cleanliness at HEAD only; `015.013-T` deliberately re-arms
enforcement only after a clean scan.

**H2 — Struct-tag over-exposure.** Mitigated by lexer-sourced candidate capture (not raw
regex), the exact tag grammar, and the mandatory non-tag raw-literal ACCEPT fixture.
*Residual*: interpreted-string tags (`"json:\"channel_id\""`) are missed — disclosed by
`015.012-T`. Findings print only `match.group(0)`, so no payload leaks to CI logs.

**H3 — Fail-closed dispatch.** Verified unreachable for ordinary additions (every path
`should_scan_repo_path` admits is mapped by `engine_for_path`) → **no CI-DoS surface**.
Synthetic-finding form keeps the finding set complete and toggle-governed.

**H4 — Characterization ordering is load-bearing.** `015.002-T` → `015.003-T` → `015.007-T`
all sit **inside sub-epic 015-B**; edges recorded in the backlog, not prose. `015.003-T` is
barred from tag positions, and `015.002-T` uses a **separate** suite, so no existing ACCEPT
fixture is disturbed.

**H5 — Self-test red window.** Bounded by `015.001-T` … `015.013-T`; single PR with a merge
commit preserves red→green evidence.

**H6 — Exclusion-boundary risk (P-021).** `015.002-T`/`015.003-T`/`015.007-T` all touch
`mask_go_non_code`, whose clone lives in the excluded `check-write-path-precondition.sh`.
AG-3 absolute; AC-4 checkable; divergence recorded as deliberate-for-this-cycle.

**H7 — Scope-expansion pressure.** `015.012-T` disclosure-only; `015.001-T`
inline-comment-only; AC-6/7/8 make AG-1/2/5 checkable.

**H8 — Toggle fail-open.** Discharged by `015.013-T`'s inverted, fail-closed form plus the
per-run `::notice::` so enforcement posture is falsifiable from run history.

**H9 — Extraction readiness (protects deferred `6C24E2E4` item 2).** New heredoc code must
add **no** bash-side dependency — no extra heredoc arguments, no shell interpolation, no root
assumptions beyond `Path.cwd()` (AC-7). Exception: `015.011-T`'s AC-6 assertion reads the
script from disk via a `Path.cwd()`-relative path, which introduces no *bash-side* coupling.

**H10 — Net enforcement regression (rev-3, was cycle-2 P2).** The advisory window is a real,
if temporary, downgrade of a live merge-blocking control taken while the operator is AFK.
Mitigated by making restoration **an in-shipment unit** (`015.013-T`), not an operator
follow-up: the shipment cannot be considered complete with enforcement still downgraded, and
AC-9 makes that checkable. Restoration is strictly **stronger** than the status quo because
the new form fails closed.

---

## 8. Merge requirement

All thirteen units MUST ship in a **single PR** with a **merge commit** (Principle XI).
Intermediate commits are deliberately RED; squash/rebase would destroy red→green evidence.

---

## 9. Rollback

Revert the single merge commit. Urgency depends on position: between `015.001-T` and
`015.013-T` the verdict is advisory, so a false positive is not merge-blocking; after
`015.013-T` normal enforcement urgency resumes.

---

## 10. Known-not-fixed

| Item | Why |
|---|---|
| `6C24E2E4` item 2 — `scripts/lib` extraction | Exclusion-boundary payoff (`BF5DE670`); plus `scripts/lib/` has no Python module convention (only `OutputPathGuard.ps1`), and extraction breaks single-file self-containment for a control that must run identically on dev machines, hooks and CI. Now also covers the harness added by `015.002-T`/`015.004-T`. |
| `6C24E2E4` item 1 — SCAN_SCOPE unification | Deferred on its **own** rationale (no exclusion risk): not among the 8 selected entries; admitting it is scope expansion against the recorded one-shipment selection. Open question 4. |
| Scope expressed **4×** | `select_repo_paths`, `should_scan_repo_path`, `expected_internal_repo_paths`, + `015.011-T`'s assertions. Accepted under AG-2. Future unification must cover **only the two production expressions** — `expected_internal_repo_paths` must remain an **independent test oracle**, or the inclusion assertion becomes vacuously self-referential. |
| `mask_go_non_code` clone in excluded script | AG-3 |
| **V10** — `mask_go_non_code` EOF fail-open | Same class as V7; disclosed by `015.012-T`, not fixed |
| Interpreted-string struct tags | H2 residual; disclosed |
| `internal/**/*_test.go` unscanned | Open question 1 |
| `go.mod` / `go.sum` outside pathspec | AG-1 / D3; disclosed only |

---

## 11. Cycle-2 finding disposition (rev 2 → rev 3)

| Sev | Persona | Finding | Disposition |
|---|---|---|---|
| **P1** | Correctness | `015.001-T` applied the toggle to the **integrity** step → violates its own AC-5, re-creating a fail-open | **FIXED** — exact per-step wiring specified; integrity step untoggled/blocking |
| **P1** | Correctness | `015.002-T` added an unmasked engine to the **shared** suite → existing `retired-accept-comment-only.go` + all `015.003-T` fixtures go red | **FIXED** — separate `retiredgo-differential/` suite with per-engine expectations; shared `go` suite untouched |
| **P1** | Correctness | `015.008-T` specified key-path composition only for tomllib → fallback divergent → dual-engine self-test permanently red | **FIXED** — composed-path matching mandated at **all three** sites incl. L368/L377 |
| **P1** | Scope | **AC-6 had no owning unit** → AG-1 checkability nominal | **FIXED** — assigned to `015.011-T`, with the disk-read form specified to avoid vacuity |
| **P1** | Scope | `015.005-T` test-first posture contradicted the corpus split | **FIXED** — carries exactly one red-first probe; corpus in `015.006-T` |
| P2 | Correctness | **Net enforcement regression, no restoration path** | **FIXED** — `015.013-T` restores, fail-closed; AG-4 rewritten; AC-9; H10 |
| P2 | Correctness | Suffix rule ambiguous — `TeamIDsCache` would evade | **FIXED** — bound to final component of the **matched window**; `TeamIDsCache` added to the REJECT corpus |
| P2 | Correctness | `015.007-T` never said how tag literals are located → raw regex re-breaks comment ACCEPT contract | **FIXED** — must capture from lexer `raw_string` state |
| P2 | Correctness | AC-6 at risk of vacuity (`__file__` unavailable in a stdin heredoc) | **FIXED** — disk-read form mandated |
| P2 | Scope | `015.008-T` largest unit, fixture set unbounded | **FIXED** — fixtures enumerated and bounded to 2 files |
| P2 | Scope | AC-8 mechanism undefined | **FIXED** — manual before/after, same as AC-2 |
| P3 | Correctness | `--self-test` semantics might silently narrow | **FIXED** — stated unchanged; new mode additive |
| P3 | Correctness | Missing edge `015.005-T → 015.007-T` | **FIXED** — added in §4 |
| P3 | Correctness | No V4 REJECT fixture named; cmd/ assertion couples to file existence | **FIXED** — `toml:"channel_id"` REJECT fixture named; failure message made self-describing |
| P3 | Scope | `015.006-T` file granularity unstated | **FIXED** — exactly 2 files |
| P3 | Scope | `015.003-T` at the 4-scenario ceiling | **ACCEPTED** — recorded as "no headroom" |
| P3 | Scope | `015-A` mislabelled `(config)` | **FIXED** — `(config + gate entrypoint)` |
| P3 | Scope | `015.012-T` enumeration invariant motivated by `EF9352FB` | **FIXED** — attributed as a forward-looking guard rail |
| P3 | Scope | `015.012-T` pre-judged `6C24E2E4`'s payoff | **FIXED** — "current and deliberate for this cycle" |

**Zero open P0/P1.**

---

## 12. Rev-4 disposition — P-013.6 escalation adjudication

Plan-review attempt 3 opened the consecutive-failure circuit breaker. Escalation route
`gpt-5.6-sol` / `openai` / `high` (Stage route `claude-opus-5` → **not** same-route
degraded) independently verified the finding and specified the corrections below.

| ID | Sev | Finding | Disposition |
|---|---|---|---|
| ESC-P1-1 | **P1** | `015.009-T` mandated a boundary fixture that is **invalid TOML**; dual-engine suite gives `reject` (tomllib parse-error finding) vs `clean` (corrected fallback) → permanently red against the unconditionally-blocking integrity step | **FIXED** — fixture replaced with a **valid-TOML** `host_cli = """` / delimiter-only-close REJECT fixture that agrees under both engines; malformed boundary relocated to `015.012-T` disclosure |
| ESC-P2-a | P2 | `015.002-T` per-engine expectations left the suite→engine seam unspecified (`engine_for_path` returns `'go'` for both suites) → executor reconstitutes the cycle-2 P1-B edit | **FIXED** — differential suite is **unmasked-only, reject-only**, selection is **suite-scoped** |
| ESC-P2-b | P2 | `015.004-T` invariant not shape-aware for dict expectations | **DISSOLVED** — manifest stays **scalar**; differential suite named as an explicit reject-only exemption |
| ESC-P2-c | P2 | AC-6 disk-read pin vacuous under plain containment (literal appears in the file being read) | **FIXED** — **region-anchored** extraction of `select_repo_paths` / `should_scan_repo_path` |
| ESC-P2-d | P2 | Binary per-file verdict lets bundled REJECT cases mask regressions; `channel.id` + `[channel]` in one document is a duplicate-table parse error passing REJECT vacuously | **FIXED** — every REJECT case is its own valid document; ACCEPT may stay bundled (asserts zero findings) |
| ESC-P3 | P3 | CI LOCAL DIVERGENCE list and stale `--self-test-integrity` usage strings | **FIXED** — folded into `015.013-T` |

Verified rev-3 verdicts carried forward: **Scope Boundary Auditor — COMPLIANT, zero P0/P1.**

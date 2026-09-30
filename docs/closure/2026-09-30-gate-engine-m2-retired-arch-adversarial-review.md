---
title: "Adversarial Review — M2 retired-architecture engine Python→Go port (commit 8ff0c39)"
date: 2026-09-30
mode: report-only
status: complete
reviewers: 3
commit: 8ff0c39
branch: feat/gate-engine-go-migration-m2-retired-architecture-engine
diff_scope: "git diff main HEAD"
plan: docs/plans/2026-09-28-intercom-go-gate-engine-go-migration-plan.md
evidence_reviewed: docs/plans/evidence/2026-09-28-gate-engine-go-migration/m2.md
---

# Adversarial Review — Unit M2 retired-architecture engine port

**Mode: report-only.** No fixes have been applied. Phase 7 (post-remediation re-review) is
**skipped** per the explicit report-only instruction — there is nothing to re-review because
nothing was auto-applied.

## 0. Reviewer configuration

Anchor route (`openai` / `gpt-5.6-sol`) was not configured for this invocation (no anchor
provider/family/effort was supplied and no anchor dispatch was attempted). Per the protocol's
count-specific mapping for `reviewers: 3` without a dispatchable anchor, the pool used the
non-anchor 3-reviewer mapping: **Reviewer-A (Tier 1) + Reviewer-B (Tier 2) + Reviewer-C (Tier 3)**.
This is recorded as a **declared fallback**: the anchor route was skipped by configuration, not by
a dispatch failure, and the remaining 3-reviewer pool satisfies the consensus minimum (≥2).

No alternate model provider (`` / ``) was requested for this invocation, so
no reviewer slot was routed to an alternate provider; Reviewer-B used the Tier 2 standard route.

| Reviewer | Route | Model | Result |
|---|---|---|---|
| Reviewer-A | Tier 1 (fast/cheap) | `gpt-5.4-mini` | completed, 1 finding |
| Reviewer-B | Tier 2 (standard) | `claude-sonnet-5` | completed, 2 findings |
| Reviewer-C | Tier 3 (frontier) | `claude-opus-5` | completed, 6 findings |

All 3 reviewer instances completed independently, each with full `view`-tool read access to the
Python ground truth (`scripts/lib/retired_arch.py`), every new Go file under
`tools/gatecheck/internal/retiredarch/`, the M1 `writepath` precedent, the wrapper/registration
wiring, and the `m2.md` evidence file, plus the extracted C-1..C-8 / ED-1..ED-9 / H-11 contract
text. Each was independently instructed not to re-litigate the `retiredarch.Run` bogus-mode
exit-2 design (pre-verified by the orchestrator as consistent with the M1 `writepath` precedent);
all three complied and did not raise it as a fresh defect.

The orchestrator (this agent) additionally performed an independent manual trace of the
`walkTable`/`peekChild` cursor logic in `tomlprimary.go` against the actual shipped
`reopened-table.toml` fixture and the actual `retiredarch_golden.json` entry, to confirm or refute
the two reviewers who converged on the same finding below. **The trace confirms the bug exactly
as both reviewers describe it** (see §1.1).

---

## 1. Consensus findings (confidence: HIGH — flagged by all 3 reviewers)

**None.** No single finding was independently flagged by all three reviewer instances using
strict `file + line±2 + rule` matching. This is expected: Reviewer-A (Tier 1) is the weakest
model in the pool and did not surface the deep TOML-cursor defect the other two found (see
§0 configuration note and §7 "reviewer capability signal").

---

## 1.1. Majority findings (confidence: MEDIUM by strict protocol; orchestrator-verified as
effectively confirmed — see below)

### F-1 — `walkTable` silently skips keys added by reopening a plain (non-array) TOML table

- **Severity:** CRITICAL
- **Confidence:** MEDIUM by strict reviewer-count rule (2 of 3: Reviewer-B, Reviewer-C
  independently converged on file `tomlprimary.go`, lines 150–153, same root cause). **The
  orchestrator independently re-traced the algorithm by hand against the actual shipped fixture
  and golden entry and confirmed the defect is real and reproducible** — treat this as
  effectively HIGH-confidence for remediation-planning purposes, reported at its formal MEDIUM
  tier only for protocol consistency.
- **File:** `tools/gatecheck/internal/retiredarch/tomlprimary.go:150-160` (`walkTable`)
- **In scope:** yes — this is a same-contract-surface completion of M2-T5 (C-6's TOML ordering
  algorithm), not a scope expansion.

**Issue.** `walkTable`'s per-table loop is bounded by `for len(opened) < len(table)` and asks
`cursor.peekChild(prefix)` for the next child. `peekChild` explicitly **rejects** (`ok=false`) any
cursor entry whose key length is `<= len(prefix)` — i.e., an *exact self-entry* for the table
currently being walked. When a plain (non-array) TOML table is **reopened** later in the document
via a second `[header]` line (e.g. `[a.b]` … `[a]` continuing to add siblings of `b` under `a`),
BurntSushi's `MetaData.Keys()` emits precisely such a self-entry (`["a"]`) for the second `[a]`
header, in addition to entries for the fields declared under it (`["a","y"]`). `walkTable`'s
`if !ok { return nil }` branch treats this self-entry identically to legitimate end-of-table and
**returns a clean result immediately**, even though `len(opened) < len(table)` still holds (the
field `y` under the reopened `[a]` has not been visited).

Manually traced by the orchestrator against the actual committed fixture
`tools/gatecheck/internal/retiredarch/testdata/_inline_toml_materialized/reopened-table.toml`
(content: `[a.b]\nx = 1\n\n[a]\ny = 2\n`):

1. Root `walkTable(prefix=nil, {a: …})` opens child `"a"` non-exactly (cursor not consumed),
   recurses into `walkTable(prefix=["a"], {b: …, y: 2}, …)`.
2. That call consumes `["a","b"]` and `["a","b","x"]` while descending into `b`'s own subtable,
   then loops again because `len(opened)=1 < len(table)=2` (field `y` still unvisited).
3. `peekChild(["a"])` now sees cursor entry `["a"]` (the reopening self-entry). Since
   `len(k)=1 <= len(prefix)=1`, `peekChild` returns `ok=false`.
4. `walkTable` hits `if !ok { return nil }` and returns a **clean** result, with the cursor
   left pointing at the unconsumed `["a"]` entry and `["a","y"]` **never visited or checked**
   against `matchesForbiddenParts`.
5. The root loop's own `opened` map is already `{a: true}` with `len(table)=1`, so the root
   loop also terminates cleanly. Nothing surfaces the skip.

Python's `walk_toml_value` (`scripts/lib/retired_arch.py`, the recursive `dict`-walk near
line ~460) has no equivalent "self-entry" concept — it recurses over the merged, decoded
`dict` directly, so `a.y` is unconditionally visited and checked.

**Consequence:** if the reopened section of a table (the `[a]` … `y = …` part in this example)
contained a retired token — e.g. `channel_id = 2` instead of `y = 2` — Python's engine would
correctly emit a finding (`retired token 'channel_id' in TOML key path 'a.channel_id' …`), while
this Go port's primary TOML engine would **silently report clean**. This is exactly the class of
regression the gate exists to prevent: a **false-clean verdict from the primary engine on a
common, entirely legal TOML idiom** (reopening a table to add siblings via a later header). It is
not covered by any of ED-1..ED-9 and is therefore, per C-6, a genuine stop-condition-class defect,
not a stylistic nit.

**Why the shipped fixture didn't catch it:** the committed `reopened-table.toml` golden uses the
inert key `y = 2`, which matches no `forbidden_parts` vocabulary word.
`retiredarch_golden.json`'s `inline_toml_goldens.reopened-table` entry has
`tomllib_findings: []` and `fallback_findings: []` — both engines report clean, so
`TestScanTomlPrimary_MatchesGolden_Inline`'s equality assertion passes **vacuously** whether or
not `y` was actually visited. The generator created forbidden-token variants for three other new
shapes (array-of-tables, nested array-of-tables, quoted dotted key) but never for the
reopened-table shape — the one shape where the bug actually lives.

**Fix.** In `walkTable`, when `cursor.peekChild(prefix)` returns `ok=false` but
`len(opened) < len(table)` still holds, do not return immediately. Check
`cursor.peekSelf(prefix)` (the same primitive `walkArrayOfTables` already uses correctly for its
own analogous "reopened element" case); if true, `cursor.consume()` and `continue` the loop so
the remaining fields are still discovered and checked. Only return a clean result when
`len(opened) == len(table)`. As a belt-and-suspenders hardening (recommended, not required): a
genuine "no `ok` and no `peekSelf`, but unvisited fields remain" state should return the same kind
of `fmt.Errorf("retiredarch: TOML cursor desync: …")` the rest of this file already uses for
invariant violations, rather than a silent `nil` — this converts the class of bug into an
automatic fail-closed reject if the fix above is ever incomplete, rather than a repeat of a
silent false-clean.

**Required companion fix (test coverage):** add a `reopened-table-forbidden-token` inline TOML
case (e.g. `[a.b]\nx = 1\n\n[a]\nchannel_id = 2\n`), regenerate the golden from the Python engine
at the recorded parent commit, materialize the fixture, and confirm `scanTomlPrimary` reproduces
Python's finding and that the primary/fallback engines agree. Without this, the fix above cannot
be golden-verified and the gap could silently reopen.

---

### F-2 — `loadFixtureManifest`/manifest-entry-value narrowing to `map[string]string` diverges from Python's looser `isinstance(data, dict)` check

- **Severity:** MINOR (behavioral stream divergence, not a verdict-flipping bug — both sides
  still exit 1 on a malformed manifest)
- **Confidence:** MEDIUM — flagged independently by **2 of 3** reviewers (Reviewer-A at
  `retiredarch.go:96`, Reviewer-C at `retiredarch.go:108`), both describing the same root cause
  in the same function at nearby lines. This is a genuine majority finding, not two unrelated
  observations.
- **File:** `tools/gatecheck/internal/retiredarch/retiredarch.go` (`loadFixtureManifest`, ~96-112)
- **In scope:** yes

**Issue.** `loadFixtureManifest` unmarshals directly into `map[string]string` and converts *any*
`json.Unmarshal` failure (including a syntactically valid JSON object whose values are not all
strings, e.g. `{"retired-x.toml": 1}`) into the same generic
`"invalid fixture manifest shape: <path>"` error. Python's `load_fixture_manifest`
(`scripts/lib/retired_arch.py` ~626-630) only requires `isinstance(data, dict)` at the top level;
a non-string value is **accepted** at load time and only surfaces later, inside
`run_fixture_self_test`'s per-fixture loop, as an `unknown expectation {value!r} in <manifest>`
failure line (still exit 1, but a materially different stdout/stderr byte stream — which C-5
requires to compare as ordered bytes for the parts that are not evidence-only).

Relatedly (per Reviewer-C, folded into this same finding since it shares the root cause): Go's
`manifest[name]` lookup models Python's `manifest.get(name) -> None` as a bare `hasExpectation`
boolean plus the empty string as a sentinel in `sortedNonEmptyKeys`; a manifest value that is
literally the empty string `""` becomes indistinguishable from an absent key, whereas Python's
`sorted(e for e in expectations if e is not None)` would still include `''` in a failure listing.

**Fix.** Decode into a shape that only rejects a non-object top level (e.g.
`map[string]interface{}` or by validating the raw `json.RawMessage` set is object-shaped without
constraining value types), then carry each fixture's expectation as a `(value string, present
bool)` pair (or `*string`) through the rest of `runFixtureSelfTest` so a present-but-non-string
value reaches the same `unknown expectation %s in %s` failure branch Python reaches, and an
absent key stays distinguishable from an explicit empty string. Reserve the hard
`"invalid fixture manifest shape"` reject for the genuinely non-object case (or a JSON syntax
error), matching Python's actual failure surface at that boundary.

---

## 2. Plurality findings (confidence: MEDIUM — flagged by more than one reviewer, not a strict
majority)

**None.** With `reviewers = 3`, "plurality but not majority" would mean exactly 2 of 3 in a
3-reviewer pool where majority is also 2 — so for this reviewer count, 2-of-3 is definitionally a
strict majority (already covered in §1.1) and there is no separate plurality tier to populate
(this is expected and correct for an odd reviewer count of 3; the tier exists formally in the
protocol for reviewer counts where 2 flags meets neither a strict majority nor a unique count,
e.g. 4 reviewers).

---

## 3. Unique findings (confidence: LOW — flagged by exactly one reviewer)

These are preserved as observations. Three of the four below are **directly corroborated** by the
now-confirmed F-1 root cause even though only one reviewer instance independently phrased each —
practical confidence for those three should be read as higher than the nominal LOW tier, because
they are downstream consequences of a bug this report independently verified, not free-floating
single-source speculation.

### F-3 — `m2.md` §7.2 overclaims order-verification strength for the reopened-table shape

- **Severity:** MAJOR
- **Confidence:** LOW (unique to Reviewer-B) — **but directly corroborated** by F-1: the
  evidence file's claim that every new inline shape's order was verified "in order,
  element-for-element" cannot be true for the reopened-table shape, because the only assertion
  exercised for it is an empty-findings-list equality, which is incapable of detecting the
  silent-skip bug F-1 describes.
- **File:** `docs/plans/evidence/2026-09-28-gate-engine-go-migration/m2.md` (§7.2, list including
  "reopened tables via dotted-header continuation")
- **In scope:** yes (evidence-file correctness is part of C-7's contract)

**Fix.** Once F-1's companion forbidden-token fixture is added and passes, this claim becomes
actually true and needs no further correction. Until then, the evidence file should be corrected
to not claim element-for-element order verification for a shape whose only test assertion is a
vacuous empty-set equality.

### F-4 — `reopened-table.toml` golden fixture is tautological (cannot fail even if the algorithm regresses further)

- **Severity:** MAJOR
- **Confidence:** LOW (unique to Reviewer-C) — same corroboration note as F-3; this is the test-
  quality framing of the same underlying gap (missing forbidden-token variant) that F-1's
  "required companion fix" and F-3 both already describe from different angles. Tracked
  separately here only so the remediation queue has an explicit test-quality line item distinct
  from the code fix.
- **File:**
  `tools/gatecheck/internal/retiredarch/testdata/_inline_toml_materialized/reopened-table.toml`
- **In scope:** yes

### F-5 — `walkTable`'s clean `return nil` on cursor/table mismatch should fail closed instead

- **Severity:** MAJOR
- **Confidence:** LOW (unique to Reviewer-C) — a defense-in-depth recommendation, not a
  independently-distinct bug from F-1; recorded separately because it is a general hardening
  principle (every other invariant break in this file already fails closed via
  `"retiredarch: TOML cursor desync: …"`; this one path is the sole silent exception) that
  outlives the specific F-1 fix.
- **File:** `tools/gatecheck/internal/retiredarch/tomlprimary.go:153`
- **In scope:** yes — already folded into F-1's recommended fix as the "belt-and-suspenders"
  hardening; listed here only for remediation-queue completeness.

### F-6 — Manifest-shape error uses `::error::` prefix instead of Python's bare `SystemExit(msg)` text

- **Severity:** MINOR
- **Confidence:** LOW (unique to Reviewer-C)
- **File:** `tools/gatecheck/internal/retiredarch/retiredarch.go:158` (manifest-load failure path)
- **In scope:** yes

**Issue.** Python's `raise SystemExit(f"invalid fixture manifest shape: {manifest_path.as_posix()}")`
prints the bare message plus a newline (C-2's "usage or infrastructure error" class, exit 1,
`msg + "\n"` on stderr — no `::error::` prefix). This Go port instead writes
`"::error::%v\n"`. ED-2 sanctions the `::error::` prefix specifically for the
traceback-to-one-line conversion class (git failure, file read failure, invalid UTF-8) — it does
not sanction reformatting an *existing* `SystemExit(msg)`-style message that Python already
prints bare. Exit code is unaffected (both sides exit 1); only the stderr bytes diverge, which
matters under C-5's ordered-byte comparison discipline for anything not evidence-only.

**Fix.** For this specific manifest-shape failure (and any other direct
`SystemExit(f"...")`-equivalent port), emit the bare message (`fmt.Sprintf("%v\n", err)`),
reserving `::error::` for the actual ED-2 traceback-replacement paths.

### F-7 — C-2 contract prose is stale re: "only bash wrappers emit 2" (documentation-only, no code defect)

- **Severity:** MINOR / advisory
- **Confidence:** LOW (unique to Reviewer-C, and pre-flagged by the orchestrator as an
  acceptable, non-bug observation before dispatch)
- **File:** `docs/plans/2026-09-28-intercom-go-gate-engine-go-migration-plan.md` (§3, C-2)
- **In scope:** **no** — this is a plan-prose nit, not a code change, and fixing the plan text is
  outside M2's task list (it would be a plan-maintenance edit, not an M2 code/test change).

**Verified as correct behavior, not a defect.** `retiredarch.Run`'s own `default:` case
returning 2 directly from inside the Go engine (rather than from the bash wrapper, which no
longer has its own case statement post-M2-T11) is **byte-identical to the pre-switch behavior**
(m2.md's parent-vs-head table: exit 2 on both sides for `--bogus`) and is **structurally
identical to `writepath.Run`'s own M1 precedent** (same default-case-returns-2 pattern, already
shipped and reviewed in M1). All three reviewers were instructed not to inflate this into a fresh
P0/P1 finding without new evidence; none did. This item is recorded purely so the plan's own C-2
prose ("only the bash wrappers emit 2 today") can eventually be updated to reflect that, post-
switch, the ported engine's own default case is the mechanism that preserves the pre-migration
observed exit code — a documentation-freshness note, not a remediation item.

---

## 4. Areas independently reviewed and found clean (no findings, all 3 reviewers)

Recorded explicitly per the quality bar that a clean area must not be silently omitted:

- **H-11 self-comparison pin (`pin.go`):** all three reviewers independently confirmed
  `pathspecPinLiterals`/`prefixPinLiterals` are hand-authored constants in `pin.go`, never
  derived from `select.go`'s text at runtime, and that `pin_test.go`'s mutation tests operate on
  a `t.TempDir()` copy of the *real* `select.go`, asserting each of the 5 enumerated mutations
  (cmd/** removed, moved to a package constant, rewritten as a different-value raw string,
  function renamed, file missing) is rejected while an unmodified copy is accepted. No reviewer
  found this vacuous or self-comparing. This satisfies M2-T8's AC and H-11.
- **Determinism (map iteration):** no reviewer found any `range` over a `map[string]...` feeding
  directly into ordered output without an intervening sort; `selectRepoPaths`, `discoverFixtures`,
  `runFixtureSelfTest`'s missing/extra-manifest diffs, and `runRepoSelectionSelfTest`'s
  expected/actual diffs all sort before use.
- **Concurrency/package-level state:** `fixtureSuites`, `vocabWords`, `forbiddenParts`,
  `pathspecPinLiterals`/`prefixPinLiterals` are all read-only after package initialization; no
  reviewer found runtime mutation of any package-level table.
- **Scope discipline (P-021):** no reviewer found any edit outside the M2 task list's authorized
  file set (no unrelated M1-package refactors, no drive-by edits).
- **`go.mod` / `go.sum`:** confirmed untouched by all three reviewers.
- **Exit-code/mode-dispatch parity (bogus mode, `--self-test` vs `--self-test-integrity`):**
  confirmed consistent with the M1 `writepath` precedent (see F-7 discussion above) by all three
  reviewers; no reviewer treated it as a fresh defect.
- **BOM fail-closed rejection (`tomlprimary.go`):** confirmed by all reviewers that
  `scanTomlPrimary` explicitly checks `strings.HasPrefix(text, "\ufeff")` and rejects **before**
  calling `toml.Decode`, correctly overriding BurntSushi's own silent-accept behavior to match
  Python's fail-closed parse rejection. Not dead code — exercised directly by
  `TestScanTomlPrimary_BOMRawBytes` and the materialized `bom-case.toml` golden.
- **Fallback lexer multiline-string state machine and unterminated-at-EOF handling
  (`tomlfallback.go`):** no reviewer found a divergence from `strip_toml_comment`/
  `scan_toml_with_fallback`; the pre-line-state capture (before the current line mutates it) and
  the final `in_basic`/`in_literal`/`in_multiline_*` fail-closed check at EOF were both confirmed
  ported verbatim.
- **Unicode/case-folding in `ident.go`** (İ, Final_Sigma, full-width Latin): no reviewer found a
  divergence from `pysem.Lower`'s documented multi-rune/Final_Sigma handling as used by
  `splitCamelAcronym`/`splitIdentifier`.
- **`exec.Command` argument construction (`pin.go`, `select.go`):** no reviewer found an
  injection or path-traversal risk; both `gitShowToplevel` and `DefaultGitRunner` pass arguments
  as a Go `[]string` argv (no shell interpolation), and `pin.go`'s file path is joined via
  `filepath.Join` from a `git rev-parse --show-toplevel`-derived root, not from any untrusted
  input.

---

## 5. Remediation queue (ordered by priority = confidence_weight × severity_weight)

| # | Finding | Confidence | Severity | Score | Action class | In scope |
|---|---|---|---|---|---|---|
| 1 | F-1 `walkTable` reopened-table silent skip | MEDIUM (2×) | CRITICAL (4×) | **8** | `gated_auto` — deterministic fix (mirror `walkArrayOfTables`'s `peekSelf` pattern), but confirm against the new golden before merging | yes |
| 2 | F-2 manifest `map[string]string` narrowing | MEDIUM (2×) | MINOR (2×) | 4 | `gated_auto` — widen the manifest value type, re-verify against a crafted non-string-value manifest fixture | yes |
| 3 | F-3 `m2.md` §7.2 overclaim | LOW (1×) | MAJOR (3×) | 3 | `manual` — correct evidence prose once F-1/F-4 land, or immediately if F-1 is deferred | yes |
| 4 | F-4 tautological `reopened-table.toml` golden | LOW (1×) | MAJOR (3×) | 3 | `manual` — add forbidden-token variant fixture + regenerate golden | yes |
| 5 | F-5 `walkTable` should fail closed, not `return nil`, on cursor/table mismatch | LOW (1×) | MAJOR (3×) | 3 | `gated_auto` (bundled with fix #1) | yes |
| 6 | F-6 manifest-shape error message prefix | LOW (1×) | MINOR (2×) | 2 | `advisory` | yes |
| 7 | F-7 C-2 prose staleness | LOW (1×) | MINOR (2×) | 2 | `advisory` (plan-doc only) | **no** — plan-maintenance edit, outside M2's code/test scope; defer to a plan-hygiene follow-up, not this shipment |

**No `safe_auto` action was identified.** Every in-scope finding here either changes observable
gate behavior (F-1, F-2, F-5) or evidence-file/fixture content (F-3, F-4) and therefore requires
human confirmation against a regenerated Python-engine golden before merge, per this repository's
own C-5 discipline ("goldens are generated from the Python engine at the change's parent commit
by a transient command… no `.py` file is committed"). Because this review is **report-only**, no
fixes were applied and Phase 7 (post-remediation re-review) did not run.

---

## 6. Backlog work-item entries (P0/P1 findings)

```yaml
type: bug
title: "F-1 tomlprimary.go walkTable: reopened plain TOML table silently skips sibling keys"
description: >
  walkTable's peekChild-driven loop returns a clean result when the cursor's next entry is an
  exact self-entry for a reopened plain (non-array) table, even though the table still has
  unvisited fields. A forbidden token placed in the reopened section of a table (e.g. `[a.b]` ...
  `[a]` `channel_id = 2`) is never checked, producing a false-clean verdict from the primary
  TOML engine that Python's recursive dict-walk would correctly reject. Confirmed by manual
  trace against the shipped reopened-table.toml fixture and its vacuous (empty-findings) golden
  entry.
file: tools/gatecheck/internal/retiredarch/tomlprimary.go
line: 150
severity: CRITICAL
confidence: MEDIUM (2 of 3 independent reviewers; orchestrator-verified by manual trace)
fix: >
  In walkTable, before returning on peekChild ok=false, check cursor.peekSelf(prefix); if true,
  consume it and continue the loop instead of returning, mirroring walkArrayOfTables' existing
  reopened-element handling. Only return cleanly once len(opened) == len(table). Add a hard
  fail-closed error (not a silent nil) for any other unvisited-fields-remain exit path. Add a
  reopened-table-forbidden-token golden fixture and regenerate retiredarch_golden.json from the
  Python engine at the recorded parent commit before merging the fix.
linked_review: docs/closure/2026-09-30-gate-engine-m2-retired-arch-adversarial-review.md
```

```yaml
type: bug
title: "F-2 retiredarch.go loadFixtureManifest: map[string]string narrowing diverges from Python's isinstance(dict) check"
description: >
  loadFixtureManifest unmarshals directly into map[string]string, rejecting any manifest with a
  non-string JSON value as "invalid fixture manifest shape" at load time. Python's
  load_fixture_manifest only requires a dict at the top level and defers non-string-value
  handling to the per-fixture "unknown expectation" failure branch inside
  run_fixture_self_test — same exit code (1), materially different stdout/stderr byte content.
  A manifest value of the literal empty string is also indistinguishable from an absent key in
  the Go port's sortedNonEmptyKeys handling, unlike Python's None-vs-'' distinction.
file: tools/gatecheck/internal/retiredarch/retiredarch.go
line: 96
severity: MINOR
confidence: MEDIUM (2 of 3 independent reviewers)
fix: >
  Decode into a shape that only rejects a non-object top level (do not constrain value types at
  load time). Carry each fixture's expectation as a (value string, present bool) pair through
  runFixtureSelfTest so a present-but-non-string value reaches the same "unknown expectation"
  failure text Python produces, and an absent key stays distinguishable from an explicit empty
  string.
linked_review: docs/closure/2026-09-30-gate-engine-m2-retired-arch-adversarial-review.md
```

---

## 7. Reviewer capability signal (observation, not a finding)

Reviewer-A (Tier 1 / `gpt-5.4-mini`) did not surface F-1 despite having identical file access,
identical contract text, and identical focus-area instructions to Reviewers B and C. This is
consistent with the adversarial-review design rationale (different models have different blind
spots for deep, multi-hop control-flow tracing like the TOML lockstep-cursor algorithm) rather
than a sign that F-1 is a false positive: two independently-instanced, higher-capability models
converged on the identical file, near-identical line number, and identical root cause without
being shown each other's output, and the orchestrator's own from-scratch manual trace against the
real fixture and golden data reproduces their conclusion exactly.

---

## 8. Post-remediation

```yaml
post_remediation:
  cycles_run: 0
  cap_reached: false
  residual_findings: 7
  status: "skipped"
```

Phase 7 is skipped: this invocation is report-only, no `safe_auto` fixes exist in the remediation
plan (see §5), and no changes were applied to any reviewed file. All 7 findings above remain open
pending a maintainer decision.

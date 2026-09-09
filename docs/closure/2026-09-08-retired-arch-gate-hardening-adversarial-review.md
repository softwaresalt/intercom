# Adversarial Review — feat/retired-architecture-gate-detection-quality-hardening

* **Scope**: `scripts/check-retired-architecture.sh`, `.github/workflows/ci.yml`, `scripts/testdata/**`
* **Base**: `origin/main` (merge-base `3a6bfff407c86815976f56b77ff6224f1948ad61`)
* **Shipment**: 014-S / feature 015-F, tasks 015.001-T through 015.013-T
* **Mode**: report-only, reviewers: 3
* **Date**: 2026-09-08

## Model route assignment table

| Reviewer | Route | Model | Result |
|---|---|---|---|
| Anchor Reviewer | Anchor route | `gpt-5.6-sol` (effort: high) | 4 findings |
| Reviewer-A | Tier 1 (fast/cheap) | `gpt-5.4-mini` | 0 findings |
| Reviewer-C | Tier 3 (frontier) | `claude-opus-5` | 2 findings |

Reviewer count 3, anchor route dispatchable → mapping used: Anchor + Reviewer-A (Tier 1) + Reviewer-C
(Tier 3), per the reviewer-count-3 anchor-dispatchable slot map. No `alt_provider`/`alt_family`
was configured for this invocation, so no Reviewer-B/Gemini substitution applied. No declared
fallback was needed — the anchor route dispatched successfully.

Reviewer-A returned an empty finding set. This is treated as a valid "no issues found" result
(not a tool failure — the agent completed and returned well-formed output), consistent with a
Tier 1 reviewer's expected lower depth on multi-step DP/lexer trace analysis. 2 of 3 reviewers
is still ≥ the 2-reviewer minimum for valid consensus aggregation.

**Note on methodology**: after initial aggregation, I (the orchestrating reviewer) independently
verified the two reviewer-agreement findings by hand-tracing the actual algorithm code and by
checking documented GitHub Actions expression-language semantics, and independently verified the
two Anchor-unique findings the same way. All four findings below are **confirmed as real**, not
merely asserted by a model. This is noted per-finding as "Independently verified: yes/no".

I was unable to run `git diff origin/main...HEAD --stat` directly (no shell/CLI execution tool
was available to me or to any dispatched sub-agent in this session — all attempts to delegate
`git diff --stat` / `git log --oneline` for a mechanical scope check failed with "no
shell/CLI execution tool available"). Scope conformance below is therefore assessed by direct
inspection of the final file contents and the `scripts/testdata/` directory listing, cross-checked
against the plan's own task/file inventory, rather than by a mechanical diff-stat scope check.
No reviewer (including the two that did deep multi-file reads) flagged any file outside the
declared scope.

---

## Consensus findings (confidence: HIGH — flagged by all 3 reviewers)

None. No finding was flagged by all three reviewers.

## Majority findings (confidence: MEDIUM — flagged by 2 of 3 reviewers)

### M-1. `split_camel_acronym` non-termination on non-alnum/non-lower/non-upper characters

* **Severity**: CRITICAL (most conservative of Anchor's MAJOR and Reviewer-C's MAJOR — both
  independently rated MAJOR; I am escalating to CRITICAL below, see rationale)
* **Flagged by**: Anchor Reviewer, Reviewer-C (2 of 3)
* **File**: `scripts/check-retired-architecture.sh`, `split_camel_acronym()` (~lines 224–265),
  reached via `decompose_toml_key()` (~line 614) → `split_identifier()` → `split_camel_acronym()`
* **Independently verified**: **yes**, by hand-trace.
* **Issue**: The final `else` branch of `split_camel_acronym()`'s per-character dispatch loop
  handles any character that is not `.isdigit()`, not `.isupper()`, and not part of an uppercase
  run:
  ```python
  j = i
  while j < n and chunk[j].islower():
      j += 1
  tokens.append(chunk[i:j])
  i = j
  ```
  If `chunk[i]` is itself not `.islower()` either (e.g. `.`, ` `, `:`, `/`, or any Unicode
  character Python does not classify as cased, such as most punctuation or many symbol code
  points), the `while` loop body never executes, so `j` stays equal to `i`. The function appends
  an empty string (`chunk[i:i]`) and sets `i = j = i` — **the loop index never advances**, so the
  `while i < n` loop spins forever.

  This is reachable through the new `015.008-T` (V5) TOML code path: `decompose_toml_key()` only
  normalizes `-` → `_` before calling `split_identifier()` on the **raw tomllib dict key**.
  TOML quoted keys legally contain arbitrary characters per the TOML spec (dots, spaces, colons,
  Unicode, etc.) — e.g. a single line `"channel.id" = 1` is valid TOML and parses via `tomllib`
  to one **flat key literally named `channel.id`** (containing a literal `.` character, distinct
  from a bare dotted key `channel.id = 1`, which tomllib instead expands into nested tables of
  plain bare-key segments and is safe).

  I hand-traced `split_camel_acronym("channel.id")` character by character: it correctly emits
  `"channel"` for the first 7 characters (all lowercase, consumed in one run), then at `i=7`
  (the `.` character) enters the final `else` branch with `chunk[7] == '.'`, which is not
  `.islower()`, so `j` stays at `7`, an empty token is appended, and `i` stays at `7` — confirmed
  non-terminating. A minimal reduction (`"a.b"`) reproduces the same non-termination at `i=1`.

  The Go path (`go_identifier_re` = `\b[A-Za-z_][A-Za-z0-9_]*\b`) cannot reach this branch because
  the regex itself never yields a non-alnum/underscore character. The TOML **fallback lexer**
  path is also safe because `bare_key_re` (`[A-Za-z_][A-Za-z0-9_]*`) pre-splits on non-word
  characters before any segment reaches `split_identifier`. **The `tomllib` call site
  (`decompose_toml_key` → `scan_toml_with_tomllib` → `walk_toml_value`) is the only unguarded
  entry point, and it is new/changed in this diff (`015.008-T`)** — the pre-`015.005-T`
  regex-based splitter (`component_re`, now removed) could not loop this way.

  No existing fixture in `scripts/testdata/` triggers this (I inspected
  `retired-reject-dotted-key.toml`, which uses a **bare** dotted key `channel.id = "value"`, not
  a quoted key with a literal dot — confirmed safe/inert for this bug). This means the bug is
  **live and untested**, not merely a theoretical corner case caught by the self-test suite.

  **Impact**: any TOML file in the scan pathspec (including `config.toml.example`, the exact file
  this gate exists to police) containing a single quoted key with a character outside
  `[A-Za-z0-9_-]` (a dot, space, colon, non-ASCII symbol, etc.) will hang the `--self-test`,
  `--self-test-integrity`, or plain repo-scan invocation of this script **forever**. The `lint`
  job in `.github/workflows/ci.yml` has **no `timeout-minutes` set** (confirmed by direct
  inspection — no matches for `timeout-minutes` anywhere in the file), so the job would run to
  GitHub's default 360-minute (6-hour) hosted-runner ceiling before being force-cancelled, blocking
  the PR/branch merge queue for the duration and burning CI compute. This is a genuine
  availability regression introduced by this exact shipment's new TOML composed-key-path code
  (`015.008-T`), not a pre-existing/disclosed issue — I confirmed via a targeted search of the
  plan document that no threat-model entry (H1–H10) or disclosed blind spot (`015.012-T`)
  mentions quoted keys, non-progress, `islower`, or infinite loops.

  **Escalation rationale (CRITICAL vs. reviewers' MAJOR)**: I am escalating this from the two
  reviewers' independent MAJOR rating to CRITICAL because (a) it is a confirmed, deterministic,
  unbounded hang reachable via a single line of syntactically valid TOML — not a false
  positive/negative in detection quality, but a correctness/availability defect in the gate
  itself; (b) it directly threatens the `lint` job, which also gates `gofmt`/`go vet`/other
  unrelated checks bundled in the same job, so one malformed-looking (but valid) TOML key can
  stall the entire lint gate for every PR that touches such a key; and (c) there is no
  `timeout-minutes` safety net.
* **Fix**: Make the final `else` branch guarantee progress: if `j == i` after the `.islower()`
  scan (i.e., the character matched none of digit/upper/lower), consume exactly one character as
  its own token (or as an explicit separator) and advance `i` by 1 before continuing, e.g.:
  ```python
  j = i
  while j < n and chunk[j].islower():
      j += 1
  if j == i:
      i += 1
      continue
  tokens.append(chunk[i:j])
  i = j
  ```
  Additionally (defense in depth), sanitize `decompose_toml_key()`'s input to characters in
  `[A-Za-z0-9_]` (mapping any other character to `_`) before calling `split_identifier()`, so the
  `tomllib` site behaves consistently with the fallback lexer's implicit non-word-character
  splitting. Add a regression fixture: a `.toml` file with a quoted key containing a literal `.`
  or space (e.g. `"a.b" = 1`) exercised by both `--self-test` and `--self-test-integrity`.
* **Action class**: `manual` — this is a live hang/DoS bug with a deterministic trigger; it must
  be fixed and re-fixtured, not merely gated behind confirmation. **P0, blocking.**

### M-2. Fused, single-case, plural identifiers/keys evade both matching models

* **Severity**: MAJOR (most conservative of Anchor's MAJOR and Reviewer-C's MINOR rating)
* **Flagged by**: Anchor Reviewer, Reviewer-C (2 of 3)
* **File**: `scripts/check-retired-architecture.sh`, `segment_whole()` (~line 297) /
  `matches_forbidden_concat()` (~line 341), composed with `split_camel_acronym()`
* **Independently verified**: **yes**, by hand-executing the `segment_whole` DP.
* **Issue**: `segment_whole()` requires an **exact, complete** segmentation of a component into
  `_VOCAB_WORDS` (union of all `forbidden_parts` values); it has no plural tolerance of its own.
  `window_matches()`'s plural-suffix tolerance (`'s'`/`'es'`) is only applied *after*
  `segment_whole()` (or `split_identifier`'s case-boundary split) has already produced a
  multi-token `parts` list — it never gets a chance to run against `segment_whole()`'s output when
  `segment_whole()` itself returns `[]` because the trailing plural suffix broke the exact-match
  requirement.

  I hand-traced `segment_whole("socketmodes")` (vocab includes `"socket"` (6), `"mode"` (4) from
  `forbidden_parts["socketmode"]`): the DP correctly reaches `dp[10] == [["socket", "mode"]]`
  (the singular prefix "socketmode" segments cleanly), but at `dp[11]` (the full 11-character
  word) every candidate vocabulary token fails to match the last few characters exactly (e.g.
  `"modes"` is not `"mode"`; no length-5 vocab word matches `word[6:11] == "modes"`), so
  `dp[11] == []`. `segment_whole("socketmodes")` returns `[]` entirely — **the whole word is
  unsegmentable**, so `matches_forbidden_concat` finds nothing.

  Separately, `split_identifier("socketmodes")` (all-lowercase, no case boundaries, no
  underscore) yields a **single one-element token list** `["socketmodes"]` via
  `split_camel_acronym`'s final `else` branch (which just consumes the whole lowercase run as one
  token, since there is no uppercase-run trigger to split on). `matches_forbidden_sequence` then
  bails immediately because `len(parts) (1) < len(want) (2)` for every 2-part forbidden sequence.

  **Net result**: an all-lowercase, unseparated, plural Go identifier or TOML key such as
  `socketmodes`, `channelids`, `teamids`, `hostclis`, or `ipcnames` evades **both** matching
  models simultaneously, despite `015.002-T`/`015.005-T`'s explicit stated goal of closing
  plural-evasion (V1) and compound/concatenation-evasion (V3) — the composition of the two
  evasion classes (plural **and** fused/unseparated) is not covered by either fix individually.
  This is a realistic naming pattern (a developer might plausibly write
  `var socketmodes []string`), not a purely theoretical corner case. I confirmed the singular
  fused form (`socketmode`) and the case-boundary plural form (`SocketModes`) **are** both
  correctly caught by the existing fixture corpus (`retired-reject-socketmode-lower.go`,
  `retired-reject-socket-modes.go` per the reviewers' description) — only the fused *and* plural
  combination is missed, and it is not listed among `015.012-T`'s disclosed residual blind spots
  (confirmed via targeted search of the plan: no mention of fused-plural forms in the blind-spot
  list, which only names aliased imports, `go.mod`/`go.sum`, non-`.go`/`.toml` files, V10, and
  interpreted-string struct tags).
* **Fix**: Extend `segment_whole()` (or a thin wrapper around it) to also attempt segmentation
  with a trailing `'s'`/`'es'` suffix stripped from the whole word before falling back to `[]`
  (mirroring `window_matches()`'s existing suffix-tolerance rule, but applied at the
  whole-component level rather than only at the final-window-element level). Add REJECT fixtures
  for at least one fused-plural Go identifier (`socketmodes`) and one fused-plural TOML key
  (`channelids`) to close the gap and prevent regression.
* **Action class**: `gated_auto` — confirm the exact suffix-handling semantics before applying
  (interacts with the shared `window_matches` plural rule and could have knock-on effects on
  other forbidden-sequence pairs; needs a human decision on where the tolerance is layered).
  **P1, should be fixed before merge or explicitly deferred with documented rationale and a
  tracked follow-up** — this directly undercuts the security claim that `015.005-T` "fixed
  plural/compound-token evasion."

## Plurality findings

None — with only 3 reviewers, "majority" (2 of 3) and "plurality" collapse to the same threshold,
so both findings above are reported under Majority.

## Unique findings (confidence: LOW — flagged by exactly 1 reviewer)

### U-1. CI toggle relies on a false "case-sensitive" claim; GitHub Actions `==` is case-insensitive

* **Severity**: MAJOR (Anchor's rating)
* **Flagged by**: Anchor Reviewer only (1 of 3)
* **File**: `.github/workflows/ci.yml`, line ~281 (`continue-on-error:` expression) and its
  preceding inline comment
* **Independently verified**: **yes** — this is documented GitHub Actions behavior, not
  speculation: GitHub Actions expression-language `==`/`!=` comparisons perform **case-insensitive**
  string comparison. The inline comment directly above the expression explicitly claims otherwise:
  > "Exact, case-sensitive `'true'` match — unset, any other value (`'True'`, `'1'`), or a typo is
  > NOT `'true'` and this step stays BLOCKING"

  This claim is **incorrect** for casing variants specifically: `${{ vars.RETIRED_ARCH_GATE_ADVISORY
  == 'true' }}` will evaluate to `true` (enabling advisory/non-blocking mode) for `'True'`,
  `'TRUE'`, `'tRue'`, etc., not only for the exact lowercase string, contradicting the comment's
  explicit example (`'True'`) which the comment claims stays blocking but which actually flips to
  advisory. The **unset-variable case is unaffected and still correctly resolves to blocking**
  (an unset `vars.*` reference resolves to an empty string, and `'' == 'true'` is false under any
  casing rule), so the core `H8` "toggle fail-open" threat — the scenario the plan's threat model
  actually cites (misconfiguration/absence of the variable) — remains discharged. The gap is
  narrower than a full bypass: it is that the **set of values which unexpectedly enable advisory
  mode is larger than documented**, which weakens (but does not void) the "misconfiguration now
  fails closed" claim in the code comment and in the plan's H8 resolution note ("Discharged by
  `015.013-T`'s inverted, fail-closed form..."). A repo operator who sets the variable to `True`
  (a very natural human capitalization choice) intending strict/blocking behavior, or copy-pastes
  a value from another tool's boolean convention, would silently get advisory/non-blocking mode
  instead — the opposite of what the comment promises for that exact example value.
* **Fix**: Either (a) correct the comment to accurately describe GitHub Actions' case-insensitive
  `==` semantics (weakest fix — doesn't change behavior, just stops mis-describing it), or
  (b) enforce true case-sensitivity by moving the comparison into a shell step
  (`[ "${{ vars.RETIRED_ARCH_GATE_ADVISORY }}" = "true" ]` in bash, which *is* byte-exact) that
  emits a boolean output consumed by `continue-on-error`, so only the literal lowercase `true`
  opts into advisory and every other casing/value (including `True`) stays blocking as the comment
  claims.
* **Action class**: `advisory` for the comment-accuracy issue alone; `gated_auto` if the stricter
  case-sensitive enforcement (fix (b)) is adopted, since it changes step semantics and needs
  confirmation. **P1** for tracking, given it is a direct, verified overclaim about a security
  control's precision, and P2 if only the documentation-only fix (a) is adopted.

### U-2. Enforcement-posture `::notice::` claim is not delivered by the notice text itself

* **Severity**: MINOR (Anchor's rating)
* **Flagged by**: Anchor Reviewer only (1 of 3)
* **File**: `.github/workflows/ci.yml` (no notice emitted here) and
  `scripts/check-retired-architecture.sh` lines ~42–45 (header claim) and ~1114–1128 (actual
  notice emission)
* **Independently verified**: **yes**, by direct inspection.
* **Issue**: The script's header comment states: "Every invocation emits a GitHub Actions
  `::notice::` line naming the resolved mode (repo / self-test / self-test-integrity) so
  enforcement posture is falsifiable from run history (H8)." In practice, the emitted notice text
  (`::notice::retired-arch gate mode=repo`, `mode=self-test`, or `mode=self-test-integrity`) names
  only the **script invocation mode**, which is fixed by which `ci.yml` step calls the script
  (verdict step always passes no flag → `mode=repo`; integrity step always passes
  `--self-test-integrity`) — it is **identical regardless of whether
  `RETIRED_ARCH_GATE_ADVISORY` is `true` or unset/false**, because the script has no visibility
  into that variable at all (it is only read by the GitHub Actions YAML's
  `continue-on-error:` expression, never passed into the script as an argument or environment
  variable). So the notice text itself cannot be used to distinguish a blocking run from an
  advisory run from run history, contrary to the literal claim. (In practice, GitHub Actions does
  visually distinguish `continue-on-error: true` steps in its UI — a failed step under
  `continue-on-error` renders with a warning/neutral indicator rather than a hard failure — so
  enforcement posture is *not* entirely unfalsifiable from run history in general, just not via
  the mechanism this specific comment describes.)
* **Fix**: Either correct the header comment to stop attributing posture-falsifiability to the
  notice line specifically (attribute it instead to GitHub's own continue-on-error step
  rendering), or actually thread the toggle value into the script (e.g. pass
  `RETIRED_ARCH_GATE_ADVISORY` as an env var to the verdict step and have the script emit
  `::notice::retired-arch gate mode=repo enforcement=blocking|advisory`) so the claim becomes
  literally true.
* **Action class**: `advisory`. **P2**, documentation-accuracy / observability nice-to-have, not
  a security-relevant gap given the underlying GitHub UI behavior.

---

## Scope / constitution check

No reviewer flagged any change outside `scripts/check-retired-architecture.sh`,
`.github/workflows/ci.yml`, or `scripts/testdata/**`. Direct inspection of the
`scripts/testdata/` directory listing shows only fixtures consistent with the declared task
inventory (`retired-*.toml` variants for TOML composed-key/multiline-state coverage, `retiredgo/`
and `retiredgo-differential/` Go fixture suites with matching manifests, plus the pre-existing
`depguard`, `gitignore`, `writepath` fixture families untouched). The CI header's own "LOCAL
DIVERGENCE" list (read in full) explicitly and correctly enumerates the `015.001-T`…`015.013-T`
verdict/integrity step wiring as an intentional, disclosed, non-upstream-template divergence
requiring re-application on regeneration — this is consistent with, not evidence against, proper
scope containment.

I was unable to run a mechanical `git diff --stat`/`git log --oneline` scope check (no
shell/CLI execution tool was available in this session to me or to any dispatched sub-agent,
including `general-purpose` and `task` agent types, which reported the same limitation). This is
a **methodology gap in this review**, not a positive scope finding — I recommend a human or a
tooling-enabled session run `git diff origin/main...HEAD --stat` before merge as a final
mechanical confirmation that no file outside the three declared paths was touched.

**No unauthorized scope expansion was identified** within the limits of file-content inspection
performed.

## Remediation plan (ordered by priority = confidence_weight × severity_weight)

| # | Finding | Confidence | Severity | Priority | Action class | Verified |
|---|---|---|---|---|---|---|
| 1 | M-1 `split_camel_acronym` non-termination (quoted TOML keys hang the gate) | MEDIUM (2/3) | CRITICAL (escalated) | 8 | `manual` | yes |
| 2 | M-2 Fused single-case plural evades both matching models | MEDIUM (2/3) | MAJOR | 6 | `gated_auto` | yes |
| 3 | U-1 CI toggle case-insensitivity vs. documented case-sensitive claim | LOW (1/3) | MAJOR | 3 | `advisory`/`gated_auto` | yes |
| 4 | U-2 Enforcement-posture notice claim not literally true | LOW (1/3) | MINOR | 2 | `advisory` | yes |

Priority scoring uses the raw reviewer-agreement confidence tier (M-1/M-2 = MEDIUM despite my own
independent confirmation, since only 2 of 3 reviewer *agents* — not all 3 — flagged them; U-1/U-2
remain LOW under the same rule despite being independently confirmed true). I recommend the
operator treat M-1 as an effective P0 regardless of its formal MEDIUM confidence tier, given the
independent hand-verification above.

## Bug/issue queue entries (P0/P1 findings)

```yaml
type: bug
title: "M-1: split_camel_acronym non-termination on quoted TOML keys with non-alnum characters"
description: >
  decompose_toml_key() -> split_identifier() -> split_camel_acronym() enters an infinite loop
  when a TOML key (reachable only via quoted keys, which tomllib preserves verbatim including
  characters outside [A-Za-z0-9_-]) contains a character that is neither a digit, uppercase, nor
  lowercase letter. The final dispatch branch never advances the loop index in this case,
  hanging the script forever. Reachable via config.toml.example or any scanned .toml file
  containing a quoted key with e.g. a literal '.' or space. The `lint` job has no
  timeout-minutes, so this stalls CI up to GitHub's 360-minute default ceiling.
file: "scripts/check-retired-architecture.sh"
line: 260
severity: "CRITICAL"
confidence: "MEDIUM (escalate to P0 given independent hand-verification)"
fix: >
  Guarantee forward progress in split_camel_acronym's final else branch (advance i by 1 when
  the lowercase-run scan makes no progress), and/or sanitize decompose_toml_key's input to
  [A-Za-z0-9_] before calling split_identifier. Add a quoted-key regression fixture exercised by
  both --self-test and --self-test-integrity.
linked_review: "docs/closure/2026-09-08-retired-arch-gate-hardening-adversarial-review.md"
```

```yaml
type: bug
title: "M-2: fused single-case plural identifiers/keys (e.g. 'socketmodes', 'channelids') evade both matching models"
description: >
  segment_whole() requires an exact, complete vocabulary segmentation with no plural tolerance
  of its own, so an all-lowercase, unseparated, plural form of a forbidden compound (e.g.
  'socketmodes') fails to segment and evades matches_forbidden_concat; it also fails
  matches_forbidden_sequence because split_identifier produces only a single un-decomposed
  token for an all-lowercase run. This undercuts 015.002-T/015.005-T's stated goal of closing
  plural and compound-token evasion for the composition of both evasion classes together.
file: "scripts/check-retired-architecture.sh"
line: 297
severity: "MAJOR"
confidence: "MEDIUM"
fix: >
  Extend segment_whole (or a wrapper) to also attempt segmentation with a trailing 's'/'es'
  stripped from the whole word, mirroring window_matches' existing suffix tolerance but applied
  at the whole-component level. Add REJECT fixtures for a fused-plural Go identifier and a
  fused-plural TOML key.
linked_review: "docs/closure/2026-09-08-retired-arch-gate-hardening-adversarial-review.md"
```

```yaml
type: bug
title: "U-1: CI toggle comment claims case-sensitive 'true' match; GitHub Actions == is case-insensitive"
description: >
  The inline comment above `continue-on-error: ${{ vars.RETIRED_ARCH_GATE_ADVISORY == 'true' }}`
  explicitly claims exact case-sensitive matching and gives 'True' as an example value that
  "stays BLOCKING" -- this is factually wrong: GitHub Actions expression-language string
  comparison is case-insensitive, so 'True'/'TRUE'/etc. all satisfy == 'true' and flip the step
  to advisory/non-blocking. The unset-variable case (the scenario H8 actually targets) is
  unaffected and still correctly defaults to blocking.
file: ".github/workflows/ci.yml"
line: 281
severity: "MAJOR"
confidence: "LOW (independently verified true)"
fix: >
  Correct the comment to describe actual case-insensitive semantics, or replace the expression
  with a shell-evaluated byte-exact comparison ([ "$VAR" = "true" ]) whose boolean output feeds
  continue-on-error, so only the literal lowercase 'true' opts into advisory as the comment
  currently (incorrectly) claims already happens.
linked_review: "docs/closure/2026-09-08-retired-arch-gate-hardening-adversarial-review.md"
```

---

## Answers to the required review questions

1. **P0/P1 blocking findings that must be fixed before PR creation**:
   * **P0**: M-1 — `split_camel_acronym` non-termination via quoted TOML keys. This is a
     confirmed, deterministic infinite loop reachable through a single line of valid TOML in any
     scanned file, with no `timeout-minutes` safety net on the `lint` job. **Must be fixed before
     merge.**
   * **P1**: M-2 — fused single-case plural evasion (`socketmodes`-style identifiers/keys) evades
     both matching models, undercutting the explicit stated goal of `015.002-T`/`015.005-T`.
     Should be fixed before merge, or explicitly deferred with a documented rationale and a
     tracked follow-up unit (it is not currently listed as a disclosed residual in `015.012-T`).
   * **P1** (documentation/precision, recommend fixing alongside M-1/M-2 since it's low-cost):
     U-1 — the CI toggle's case-sensitivity claim is factually incorrect for the GitHub Actions
     expression language, though the unset-variable fail-closed guarantee itself is intact.

2. **P2/P3 advisory findings for tracking**:
   * U-2 — the enforcement-posture `::notice::` claim is not literally delivered by the notice
     text (posture is only inferable via GitHub's own continue-on-error UI rendering, not from
     the script's own output). Documentation-accuracy issue, not a security gap.

3. **Is the toggle inversion (`015.013-T`) verified correct and truly fail-closed by default?**
   **Yes, for the specific scenario the plan's H8 threat model targets** (the variable being
   unset/absent, e.g. due to misconfiguration or a fresh repo fork): `vars.RETIRED_ARCH_GATE_ADVISORY`
   resolves to an empty string when unset, `'' == 'true'` is false under any casing rule, so
   `continue-on-error` evaluates to `false` and the step stays blocking. **However**, the
   companion inline-comment claim that the match is "exact, case-sensitive" is **not true** per
   documented GitHub Actions expression semantics (`==` is case-insensitive for strings) — see
   U-1. This does not reopen the H8 fail-open scenario (unset variable), but it does mean the
   set of *explicitly-set* values that inadvertently enable advisory mode is wider than the
   comment/plan documentation asserts. Recommend either fixing the comment or hardening to a
   true byte-exact shell comparison.

4. **Any out-of-scope changes relative to the 13 tasks / 8 stash IDs?**
   None identified within the limits of this review (no shell tool was available to run a
   mechanical `git diff --stat`/`git log` scope check in this session — see the "Scope /
   constitution check" section above for the recommended manual follow-up). File-content
   inspection of the final script, CI workflow, and `scripts/testdata/` directory listing shows
   content fully consistent with the declared 13-task narrative and the CI header's own
   self-documented LOCAL DIVERGENCE list. The known, disclosed `check-write-path-precondition.sh`
   `mask_go_non_code` clone divergence was correctly *not* re-flagged as a fresh finding by any
   reviewer, per instruction.

5. **Overall verdict: BLOCKED**

   This diff should **not** proceed to PR creation in its current state. M-1 is a confirmed,
   reachable, unbounded CI hang introduced by this shipment's own new code (`015.008-T`'s TOML
   composed-key-path change), with no timeout safety net — this is a hard blocker independent of
   its formal MEDIUM aggregation-confidence label (2 of 3 reviewer models found it independently,
   and I confirmed it by hand-tracing the algorithm to a concrete non-terminating input). M-2
   should also be resolved or explicitly and consciously deferred before merge, since it directly
   undercuts a headline claim of this shipment ("fixed plural/compound-token evasion"). Once M-1
   is fixed (and ideally M-2), and U-1's comment/behavior mismatch is reconciled, this becomes
   **READY_WITH_FOLLOWUPS** (U-2 remaining as a P2 advisory item). No consensus (3-of-3) findings
   were produced, and no scope violations were identified.

---

## Post-remediation review

`post_remediation_review` was not exercised in this invocation: mode was report-only, and per
the review request scope, no auto-fixes were applied in this pass (this report enumerates
findings and a remediation plan only; it does not itself modify `scripts/check-retired-architecture.sh`
or `.github/workflows/ci.yml`).

```yaml
post_remediation:
  cycles_run: 0
  cap_reached: false
  residual_findings: 0
  status: "skipped"
```

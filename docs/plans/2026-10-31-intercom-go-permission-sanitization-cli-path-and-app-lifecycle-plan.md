---
title: "Permission Sanitization, Confusable URL Defense, CLI Path Trust, and Process Lifecycle Hardening Implementation Plan"
description: "Implementation plan for Option 2 (Capability Set 1 + Capability Set 2) across two chained single-feature shipments (078-S -> 079-S): permission feedback sanitization, confusable URL defense, emoji/ZWJ summary policy, and Copilot cli_path resolution hardening (085-F); and process lifecycle, credential hygiene, localUI supervision, and config.Unconsumed architecture (086-F)."
source: "docs/decisions/2026-10-31-intercom-go-permission-sanitization-cli-path-and-app-lifecycle-deliberation.md"
status: "ready"
date: 2026-10-31
tags:
  - "permissions"
  - "cli-path"
  - "credentials"
  - "lifecycle"
  - "tui"
  - "config"
  - "security"
---

## Problem Frame

With `076-S..077-S` merged and closed on `main` (`f00ef564f49ff38a0ad9c3d9f756eb3a11edb59d`, code merge `3b1135530989eb8fe60a55fcab768ca61fd3cd74`), Option 2 (Capability Set 1 + Capability Set 2) resolves 13 active `DEFERRED SCOPE EXPANSION` stash entries (`2EE9F3ED`, `D3D0E9D5`, `F24EC587`, `2C45EFAE`, `2F1B6D4B`, `24CE7E1D`, `5E4834CF`, `AE317BD3`, `4B7308F2`, `7BFA4182`, `26E15024`, `864FDF7A`, and `D953F24D`) spanning `internal/hub`, `internal/agent/copilot`, `internal/pathsafe`, `internal/config`, `internal/transport`, `internal/tui`, `internal/app`, `cmd/intercom`, and `docs/`.

To satisfy Ship Step 6.1(a1) `S2` (`n == 1` covering feature per shipment) and maintain strict architectural layering and width isolation, the work is organized into two sequential single-feature shipments (`078-S -> 079-S`):

* **Shipment `078-S` / Covering Feature `085-F`**: Permission feedback sanitization, confusable URL defense, emoji/ZWJ summary policy, and Copilot `cli_path` resolution hardening (`2EE9F3ED`, `D3D0E9D5`, `F24EC587`, `2C45EFAE`, `2F1B6D4B`). Declared a DAG root (`labels: ["dag-root"]`, `predecessor_source: "declared_root"`) under the operator's explicit P-001 policy exception over `062-S`, with `077-S` (`f00ef564f49ff38a0ad9c3d9f756eb3a11edb59d`) documented as the code baseline.
* **Shipment `079-S` / Covering Feature `086-F`**: Process lifecycle, credential hygiene, `localUI` supervision, and `config.Unconsumed` architecture (`24CE7E1D`, `5E4834CF`, `AE317BD3`, `4B7308F2`, `7BFA4182`, `26E15024`, `864FDF7A`, `D953F24D`), blocked by `078-S` (`backlogit dep add 079-S 078-S --type blocks`).

## Requirements Trace

| Source Requirement | Stash Origin | Target Shipment & Feature | Implementation Unit(s) |
|---|---|---|---|
| Guard `truncUTF8(s, n)` against `n <= 0` (`if n <= 0 { return "" }`) so negative bounds never panic; sanitize operator deny feedback (`Decision.Feedback`) in `Hub.resolvePermission` via `sanitizeFeedback(s, MaxDecisionFeedbackBytes)` so invalid UTF-8 bytes, control runes (including `\n`, `\r`, `\t`, C0, DEL, C1), line/paragraph separators (`U+2028`, `U+2029`), format/bidi runes (`unicode.Cf`), and visually blank filler runes are replaced with `'?'` and the result is bounded on a UTF-8 rune boundary to `MaxDecisionFeedbackBytes = 1024` | `2EE9F3ED` | `078-S` / `085-F` | Unit 1.1 (`085.001-T`), Unit 1.5 (`085.005-T`) |
| Sanitize `feedback` in `copilot.rejection(feedback)` as defence in depth; replace the hand-written `urls:` homoglyph switch in `internal/agent/copilot/permission.go` with pure-stdlib NFKC compatibility decomposition (`nfkcFoldURLsRune`: fullwidth ASCII, mathematical alphanumeric alphabets/digits `U+1D400..U+1D7FF`, parenthesized/circled/enclosed Latin, superscript/subscript/modifier letters, Roman numerals, compatibility colons) plus UTS #39 confusable-skeleton folding, and skip interior invisible/blank/format/replacement/combining runes in `hasURLsLookalikePrefix`; lock the fail-closed policy keeping `Variation_Selector` (`U+FE0E`/`U+FE0F`/`U+E0100..U+E01EF`), `Cf` (`ZWJ`/`ZWNJ`), and `Co` (Private Use Area) replaced with `U+FFFD` / `<U+XXXX>` across `copilot` and `tui` | `2EE9F3ED`, `D3D0E9D5`, `F24EC587` | `078-S` / `085-F` | Unit 1.2 (`085.002-T`), Unit 1.5 (`085.005-T`) |
| Extract shared `pathsafe.ResolveCLIPath(p string) (string, CLIPathErrorKind, error)` and `pathsafe.SanitizeCLIPathForMessage(p string) string` in `internal/pathsafe/lexical.go`, consolidating `ClassifyCLIPath`, `exec.LookPath`, `filepath.Abs`, `os.Stat` + `IsRegular()`, symlink/junction/mapped-drive canonicalization via `canonicalizeReparse` (rejecting targets that resolve to UNC or device paths), NTFS ADS / extra-colon rejection (`hasADSOrExtraColon`), `.bat`/`.cmd` batch-script rejection, and Unix execute-bit validation (`cinfo.Mode().Perm()&0o111 != 0`) | `2C45EFAE`, `2F1B6D4B` | `078-S` / `085-F` | Unit 1.3 (`085.003-T`) |
| Wire `pathsafe.ResolveCLIPath` and `pathsafe.SanitizeCLIPathForMessage` into `internal/config/validate.go` (`validateCLIPath`) and `internal/agent/copilot/adapter.go` (`resolveCLIPath`), keeping caller-specific warning vs. error policy and `%w` error wrapping; document `pathsafe.ResolveCLIPath` hardening and verified empty-`cli_path` SDK behavior in design §6.2/§8 and `docs/config-reference.md` | `2C45EFAE`, `2F1B6D4B` | `078-S` / `085-F` | Unit 1.4 (`085.004-T`), Unit 1.5 (`085.005-T`) |
| Hold `transport.Token` plaintext in a wipeable `[]byte` (`type Token struct{ b []byte }`) with `(*Token).Wipe()` (`clear(t.b); t.b = nil`) and `(Token).Clone() Token` | `AE317BD3` | `079-S` / `086-F` | Unit 2.1 (`086.001-T`) |
| Render metadata-only `tui.Credential` entries (`Secret == ""`) as `(already shown)` in `tui.Model.credentialRows` so a partial `Reissue` failure after `zeroPlaintext` never omits a still-live unreplaced credential; add `Options.UnconsumedKeys []string` to `tui.Options` so dismissing the initial credential panel with `Esc` (or starting without credentials) surfaces a one-time `"unconsumed config keys: ..."` notice on the TUI status line | `5E4834CF`, `D953F24D` | `079-S` / `086-F` | Unit 2.2 (`086.002-T`), Unit 2.9 (`086.009-T`) |
| Move the accepted-but-unconsumed config key list to `config.Unconsumed(cfg *Config) []string` and `config.UnconsumedKeyNames() []string` in `internal/config/unconsumed.go`, compare `cfg.Database.Path != def.Database.Path`, add a reflect-based test that every `Config` schema key is either consumed or listed in `UnconsumedKeyNames()`, and derive `cmd/intercom/runbook_test.go` from `config.UnconsumedKeyNames()` | `864FDF7A` | `079-S` / `086-F` | Unit 2.3 (`086.003-T`), Unit 2.9 (`086.009-T`) |
| Extract `zeroTokensLocked()` and `clearLocked()` helpers in `internal/app/credentials.go`, make `mint()` early-out on `credsCleared` under `a.mu`, emit metadata-only `tui.Credential` entries in `credentialsMsg` when `Secret == ""` and `Fingerprint != ""`, and return cloned `Token`s in `App.Credentials()` | `4B7308F2`, `5E4834CF`, `AE317BD3` | `079-S` / `086-F` | Unit 2.4 (`086.004-T`) |
| Unexport `TUIOptions.RetainPlaintext` (`retainPlaintext bool`) in `internal/app/app.go` and lock with a reflection test | `AE317BD3` | `079-S` / `086-F` | Unit 2.5 (`086.005-T`) |
| Group `App` credential fields (`mintMu`, `live`, `cleared`, `zeroAfterTUI`, `afterMint`) into `credentialState` (`creds credentialState`) mirroring `shutdownState` | `26E15024` | `079-S` / `086-F` | Unit 2.6 (`086.006-T`) |
| Move `teardown` and `defaultTeardownPermissions` from `assemble.go` into `shutdown.go` beside `reap`, derive `maxApproval` from `config.MaxApprovalTimeoutSeconds`, delegate `unconsumedKeys` to `config.Unconsumed`, call `a.clearCreds()` immediately after `<-a.shut.req` in `supervise()`, and test both immediate credential clearing on shutdown start and the `supervise` -> `teardown` fallback when `hub.NewShutdown` fails; document failed-start `teardown` as a reduced shutdown sequence in design §3.1 | `24CE7E1D`, `26E15024`, `864FDF7A`, `D953F24D` | `079-S` / `086-F` | Unit 2.7 (`086.007-T`), Unit 2.9 (`086.009-T`) |
| Move `watchUI`, `pumpRejects`, `pumpRejectsTo`, `defaultRejectNoticeInterval`, and `onReject` from `assemble.go` into `ui.go`, create `watchDone` and start the watcher inside `startUI` (`startWatchedUI`), drop redundant `quit`/`kill` nil defaults in `(*localUI).stop()`, wire `UnconsumedKeys` into `startUI`, and expose `(*App).RejectedUpgrades() uint64` | `7BFA4182`, `26E15024`, `D953F24D` | `079-S` / `086-F` | Unit 2.8 (`086.008-T`) |

## Constitution Check

| Principle | Compliance | Notes |
|---|---|---|
| I. Safety-First Go | PASS | Targets Go 1.25.0 (`GOTOOLCHAIN=go1.26.10`), uses explicit sentinel/wrapped `apperr` errors, bounds all waits, and adds zero `panic` calls in library code. |
| II. Test-First Development (NON-NEGOTIABLE) | PASS | Every `code` unit (`085.001-T`..`085.004-T` and `086.001-T`..`086.008-T`) specifies failing red-phase test assertions in its colocated `_test.go` file before production edits. |
| III. Workspace Isolation and Security Boundaries | PASS | Zero new filesystem write calls and zero new `pathsafe.Root.Resolve` call sites (`go run ./tools/gatecheck write-path --self-test --root .` stays 100% green). `pathsafe.ResolveCLIPath` uses `pathsafe`'s internal `canonicalizeReparse` directly. |
| IV. CLI Workspace Containment (NON-NEGOTIABLE) | PASS | All edits and test fixtures remain strictly inside the repository or `t.TempDir()`. |
| V. Structured Observability | PASS | Unconsumed config keys are surfaced both in the structured `WARN` log and as a one-time TUI status-line notice (key names only, never values). |
| VI. Single Responsibility | PASS | Zero new external module dependencies in `go.mod`; all `.golangci.yml` `depguard` boundaries are preserved. |
| VII. Destructive Command Approval (NON-NEGOTIABLE) | PASS | No destructive commands are used; consumed stash entries are archived via non-destructive `backlogit stash archive`. |
| VIII. Explicit Safety Modes | PASS | Freeze-scope mode active on the 13 selected stash entries. |
| IX. Git-Friendly Persistence | PASS | Deliberation, plan, session memory, and backlog artifacts use Markdown with YAML frontmatter. |
| X. Agent Context Efficiency | PASS | Targeted code and backlog reads used throughout staging. |
| XI. Merge Commit History Preservation | PASS | All shipments merge via `--merge` commits (P-009). |

## Implementation Units

### Sub-Epic 1: Shipment `078-S` — Covering Feature `085-F` (Permission feedback sanitization, confusable URL defense, emoji/ZWJ summary policy, and Copilot `cli_path` resolution hardening)

#### Unit 1.1 (`085.001-T`): Guard `truncUTF8` against non-positive `n` and sanitize operator deny feedback in `internal/hub` (`2EE9F3ED`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/hub/fold.go`
  * `internal/hub/permission.go`
  * `internal/hub/permission_test.go`
* **Changes needed**:
  * In `internal/hub/fold.go`:
    * In `truncUTF8(s string, n int) string`: add `if n <= 0 { return "" }` as the first statement before `if len(s) <= n`, so `n == 0` and negative `n` always return `""` without slicing `s[:n]` with a negative index.
  * In `internal/hub/permission.go`:
    * Add `func sanitizeFeedback(s string, limit int) string`: if `limit <= 0`, return `""`; iterate over `s` by byte index using `r, size := utf8.DecodeRuneInString(s[i:])`, replacing invalid UTF-8 (`r == utf8.RuneError && size == 1`), control runes (`unicode.IsControl(r)`, which covers `\n`, `\r`, `\t`, C0, DEL, C1), Unicode line/paragraph separators (`r == '\u2028' || r == '\u2029'`), format/bidi runes (`unicode.Is(unicode.Cf, r)`), private-use (`unicode.Is(unicode.Co, r)`), variation selectors (`unicode.Is(unicode.Variation_Selector, r)`), and visually blank filler runes (`\u2800`, `\u3164`, `\u115f`, `\u1160`, `\uffa0`, `\u034f`, `\u17b4`, `\u17b5`) with `'?'`, stopping before `n+utf8.RuneLen(r) > limit` so the output is single-line, valid UTF-8, and at most `limit` bytes.
    * In `(*Hub).resolvePermission(r *actionReq)`: replace `dec.Feedback = truncUTF8(dec.Feedback, MaxDecisionFeedbackBytes)` with `dec.Feedback = sanitizeFeedback(dec.Feedback, MaxDecisionFeedbackBytes)`.
  * In `internal/hub/permission_test.go`:
    * Add `TestTruncUTF8HandlesNonPositiveBounds`: assert `truncUTF8("hello", 0) == ""` and `truncUTF8("hello", -5) == ""` without panicking.
    * Add `TestDenyFeedbackIsSanitizedBeforeDeliveryToAgent`: verify through a running `Hub` that a `DecisionDeny` carrying C0/C1 control characters (`\x00`, `\x1b`, `\u009b`), newlines/tabs (`\n`, `\r`, `\t`), line separators (`\u2028`, `\u2029`), bidi overrides (`\u202e`, `\u2066`), visually blank filler (`\u2800`, `\u3164`), and invalid UTF-8 bytes (`"\xff\xfe"`) delivers a sanitized, valid-UTF-8, single-line `Feedback` (`<= MaxDecisionFeedbackBytes`) with all forbidden runes replaced by `'?'`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/hub -run "TestTruncUTF8|TestDenyFeedbackIsSanitized|TestHubPackageHasNoBareChannelSends"`

#### Unit 1.2 (`085.002-T`): Sanitize adapter deny feedback, replace `urls:` denylist with NFKC + confusable-skeleton folding and interior-invisible tolerance, and lock fail-closed emoji/ZWJ/PUA sanitization policy (`2EE9F3ED`, `D3D0E9D5`, `F24EC587`)

* **Domain**: `code` (`size: S`, `complexity: medium`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/agent/copilot/permission.go`
  * `internal/agent/copilot/permission_cycle2_test.go`
  * `internal/tui/sanitize_test.go`
* **Changes needed**:
  * In `internal/agent/copilot/permission.go`:
    * Define `const maxDecisionFeedbackBytes = 1024` and `func sanitizeFeedback(s string, maxBytes int) string` (decoding via `utf8.DecodeRuneInString`, replacing invalid UTF-8 `r == utf8.RuneError && size == 1` and `isInvisibleOrBlankRune(r)` — including `\n`, `\r`, `\t` — with `utf8.RuneError` `U+FFFD`, and bounding to `maxBytes` bytes on a rune boundary). Update `rejection(feedback string)` to call `feedback = sanitizeFeedback(feedback, maxDecisionFeedbackBytes)` before setting `r.Feedback = &feedback`.
    * Add `func nfkcFoldURLsRune(r rune) rune`: a bounded, targeted single-rune NFKC compatibility folder (Unicode 16.0 compatibility mappings for the 5 target ASCII characters `'u'`, `'r'`, `'l'`, `'s'`, `':'`, not a general arbitrary-text NFKC engine):
      * Fullwidth ASCII (`0xFF01..0xFF5E` -> `r - 0xFEE0`);
      * Mathematical Alphanumeric Symbols (`U+1D400..U+1D6A3` across all 13 uppercase/lowercase 26-letter Latin alphabets and mathematical digits `U+1D7CE..U+1D7FF` across all 5 10-digit runs);
      * Parenthesized Latin small (`U+249C..U+24B5`), circled Latin capital (`U+24B6..U+24CF`) and small (`U+24D0..U+24E9`), and parenthesized Latin capital (`U+1F110..U+1F129`);
      * Roman numerals (`U+2160..U+216F` and `U+2170..U+217F` for `I`/`i`/`L`/`l`/`V`/`v`), modifier/superscript/subscript letters (`ᵘ` `U+1D58`, `ᵁ` `U+1D41`, `ʳ` `U+02B3`, `ᴿ` `U+1D3F`, `ˡ` `U+02E1`, `ᴸ` `U+1D38`, `ˢ` `U+02E2`, `ᵢ` `U+1D62`, `ᵣ` `U+1D63`, `ᵤ` `U+1D64`, `ₗ` `U+2097`, `ₛ` `U+209B`), and compatibility colons (`\uFE30`, `\uFE13`, `\uFE55`, `\u205A`, `\u2982`, `\u2A74`).
    * Update `foldURLsLookalikeRune(r rune) (byte, bool)` to apply `r = nfkcFoldURLsRune(r)` first, then fold through the bounded UTS #39 single-codepoint confusable skeleton table for `:`, `u`, `r`, `l`, `s` (adding Greek/Cyrillic/Latin-extended/IPA/Armenian/Cherokee confusables such as `\u028b` `ʋ`, `\u03bd` `ν`, `\u054d`, `\u017f` `ſ`, `\u026d` `ɭ`, `\u026b` `ɫ`, `\u2223` `∣`, `\u00a6` `¦`, `\u0282` `ʂ`, `\u01a7`/`\u01a8`, `\u02f8` `˸`, `\u0703`, `\u0704`, `\u1803`, `\u1809`).
    * Update `hasURLsLookalikePrefix(s string) bool` so that after `lead` transitions to `false`, it continues to skip `isInvisibleOrBlankRune(r) || r == utf8.RuneError || isCombiningMark(r)` in interior positions (tolerating interior invisible/blank/format/replacement runes such as `u\u200br\u3164l\ufe0fs:` while still rejecting interior ASCII spaces `"u r l s:"`).
  * In `internal/agent/copilot/permission_cycle2_test.go`:
    * Add `TestRejectionSanitizesControlBidiNewlineAndInvalidUTF8Feedback`: assert `toSDKDecision(agent.PermissionDecision{Kind: agent.DecisionDeny, Feedback: "no\nway\x1b[31m\u202e\xff"})` produces sanitized single-line `Feedback` with `\n`, `\x1b`, `\u202e`, and `\xff` replaced by `U+FFFD` and bounded to `1024` bytes.
    * Add `TestHasURLsLookalikePrefixFoldsNFKCAndToleratesInteriorInvisibleRunes`: verify `shellSummary` / `indentURLsFraming` indents mathematical bold/italic/fraktur/double-struck/monospace `𝐔𝐫𝐥𝐬:`, `𝑢𝑟𝑙𝑠:`, `𝕦𝕣𝕝𝕤:`, parenthesized/circled `⒰⒭⒧⒮:`, Roman numeral `uRⅼs:`, and interior-invisible `u\u200br\u200cl\u3164s\ufe0f:` by `"  "`.
    * Add `TestSummarySanitizationReplacesEmojiVariationSelectorsZWJAndPrivateUseRunes`: verify `sanitizeBound` replaces `U+FE0E`, `U+FE0F`, `U+E0100`, `U+200C` (ZWNJ), `U+200D` (ZWJ), and `U+E000`/`U+F8FF` (PUA) with `U+FFFD`.
  * In `internal/tui/sanitize_test.go`:
    * Add `TestSanitizeReplacesEmojiVariationSelectorsZWJAndPrivateUseWithVisibleMarkers`: verify `Sanitize` replaces `U+FE0E`, `U+FE0F`, `U+E0100`, `U+200C`, `U+200D`, and `U+E000`/`U+F8FF` with `<U+XXXX>` markers and marks a string composed solely of them as blank via `isBlankSummary`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/agent/copilot ./internal/tui -run "TestRejectionSanitizes|TestHasURLsLookalike|TestSummarySanitization|TestSanitize"`

#### Unit 1.3 (`085.003-T`): Extract shared `pathsafe.ResolveCLIPath` and `pathsafe.SanitizeCLIPathForMessage` with symlink/junction canonicalization, ADS/batch-script rejection, and Unix execute-bit validation (`2C45EFAE`, `2F1B6D4B`)

* **Domain**: `code` (`size: S`, `complexity: medium`)
* **Posture**: `test-first`
* **Files affected** (2 files):
  * `internal/pathsafe/lexical.go`
  * `internal/pathsafe/lexical_test.go`
* **Changes needed**:
  * In `internal/pathsafe/lexical.go`:
    * Define `type CLIPathErrorKind uint8` with constants `CLIPathErrNone`, `CLIPathErrUNCOrDevice`, `CLIPathErrDriveRelative`, `CLIPathErrRelative`, `CLIPathErrNotFound`, `CLIPathErrNotRegular`, `CLIPathErrNotExecutable`, `CLIPathErrBatchScript`, `CLIPathErrLookPath`, `CLIPathErrAbsResolve`, and `CLIPathErrResolvesToUNCOrDevice`.
    * Add `func ResolveCLIPath(p string) (resolved string, errKind CLIPathErrorKind, cause error)`:
      * Switches on `ClassifyCLIPath(p)`:
        * `CLIPathEmpty`: returns `"", CLIPathErrNone, nil`.
        * `CLIPathUNCOrDevice`: returns `"", CLIPathErrUNCOrDevice, nil`.
        * `CLIPathDriveRelative`: returns `"", CLIPathErrDriveRelative, nil`.
        * `CLIPathRelative`: returns `"", CLIPathErrRelative, nil`.
        * `CLIPathAbs`: if `hasADSOrExtraColon(p)`, returns `"", CLIPathErrNotRegular, nil`; otherwise calls `validateResolvedCLIFile(p)`.
        * `CLIPathBare`: calls `found, err := exec.LookPath(p)` (on error returns `"", CLIPathErrLookPath, err`); `abs, err := filepath.Abs(found)` (on error returns `"", CLIPathErrAbsResolve, err`); if `ClassifyCLIPath(abs) != CLIPathAbs`, returns `"", CLIPathErrResolvesToUNCOrDevice, nil`; if `hasADSOrExtraColon(abs)`, returns `"", CLIPathErrNotRegular, nil`; calls `validateResolvedCLIFile(abs)`.
      * `validateResolvedCLIFile(abs string) (string, CLIPathErrorKind, error)`:
        * `info, err := os.Stat(abs)`; if `err != nil`, returns `"", CLIPathErrNotFound, err`; if `!info.Mode().IsRegular()`, returns `"", CLIPathErrNotRegular, nil`.
        * Canonicalizes symlinks/junctions/mapped drives via `canonical, err := canonicalizeReparse(abs)`; if `err != nil`, returns `"", CLIPathErrNotFound, wrapPathError("open", abs, err)`.
        * If `ClassifyCLIPath(canonical) != CLIPathAbs`, returns `"", CLIPathErrResolvesToUNCOrDevice, nil`.
        * If `hasADSOrExtraColon(canonical)`, returns `"", CLIPathErrNotRegular, nil`.
        * `cinfo, err := os.Stat(canonical)`; if `err != nil`, returns `"", CLIPathErrNotFound, err`; if `!cinfo.Mode().IsRegular()`, returns `"", CLIPathErrNotRegular, nil`.
        * If `isBatchScriptExt(abs) || isBatchScriptExt(canonical)` (`.bat` or `.cmd`, case-insensitive), returns `"", CLIPathErrBatchScript, nil`.
        * If `runtime.GOOS != "windows" && cinfo.Mode().Perm()&0o111 == 0`, returns `"", CLIPathErrNotExecutable, nil`.
        * Returns `canonical, CLIPathErrNone, nil`.
    * Add `func SanitizeCLIPathForMessage(p string) string`: returns `p` unchanged when `utf8.ValidString(p)` and every rune in `p` satisfies `unicode.IsPrint(r) && !isInvisibleRuneForMsg(r) && r != '\''`; otherwise escapes non-printable, control, bidi/format, invalid UTF-8, and `'` runes so single-quoted error messages (`copilot.cli_path '%s'`) never emit raw control characters or unescaped quotes.
  * In `internal/pathsafe/lexical_test.go`:
    * Add `TestResolveCLIPathConsolidatesAndHardensCases`: test empty, UNC/device, drive-relative, relative, missing absolute, directory, valid regular executable file, symlink to regular file (returns canonical target), NTFS ADS suffix (`C:\dir\copilot.exe:stream`), `.bat`/`.cmd` batch scripts (both explicit and via `PATH`), Unix `0o644` non-executable file (on `runtime.GOOS != "windows"`), and `SanitizeCLIPathForMessage` on printable vs. control/newline/quote-containing paths.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/pathsafe -run "TestClassifyCLIPath|TestResolveCLIPath"`
  * `GOTOOLCHAIN=go1.26.10 go run ./tools/gatecheck write-path --self-test --root .`

#### Unit 1.4 (`085.004-T`): Wire `pathsafe.ResolveCLIPath` and `pathsafe.SanitizeCLIPathForMessage` into `internal/config` and `internal/agent/copilot` (`2C45EFAE`, `2F1B6D4B`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/config/validate.go`
  * `internal/config/copilotcli_test.go`
  * `internal/agent/copilot/adapter.go`
* **Changes needed**:
  * In `internal/config/validate.go`:
    * Rewrite `validateCLIPath(cliPath string, report *Report) (string, error)` to call `resolved, errKind, _ := pathsafe.ResolveCLIPath(cliPath)` and format `safe := pathsafe.SanitizeCLIPathForMessage(cliPath)` into the single-quoted `apperr.KindConfig` messages:
      * `CLIPathErrNone`: `return resolved, nil`
      * `CLIPathErrLookPath`: append `fmt.Sprintf("copilot.cli_path '%s' was not resolvable via PATH", safe)` to `report.Warnings` and `return cliPath, nil`
      * `CLIPathErrUNCOrDevice`: `copilot.cli_path '%s' must not be a UNC or device-namespace path; use an absolute path, a bare name, or empty`
      * `CLIPathErrDriveRelative`: `copilot.cli_path '%s' must not be drive-relative; use an absolute path, a bare name, or empty`
      * `CLIPathErrRelative`: `copilot.cli_path '%s' must not be relative; use an absolute path, a bare name, or empty`
      * `CLIPathErrNotFound`: `copilot.cli_path '%s' does not exist`
      * `CLIPathErrNotRegular`: `copilot.cli_path '%s' must be a regular file`
      * `CLIPathErrNotExecutable`: `copilot.cli_path '%s' must be executable`
      * `CLIPathErrBatchScript`: `copilot.cli_path '%s' must not be a .bat or .cmd batch script`
      * `CLIPathErrAbsResolve`: `copilot.cli_path '%s' could not be resolved to an absolute path`
      * `CLIPathErrResolvesToUNCOrDevice`: `copilot.cli_path '%s' must not resolve to a UNC or device-namespace path`
  * In `internal/agent/copilot/adapter.go`:
    * Rewrite `resolveCLIPath(p string) (string, error)` to call `resolved, errKind, cause := pathsafe.ResolveCLIPath(p)` and map each `CLIPathErrorKind` to its `agent: ...` error (wrapping `cause` with `%w` on `CLIPathErrNotFound`, `CLIPathErrLookPath`, and `CLIPathErrAbsResolve`).
  * In `internal/config/copilotcli_test.go`:
    * Update `TestValidateRule5RejectsDirectoryCLIPathAndAcceptsRegularFile` to create `file` inside `cfg.DefaultWorkspaceRoot` (which is canonicalized via `pathsafe.NewRoot`) so `cfg.Copilot.CLIPath` equals the canonicalized `file` path on every OS (including Windows 8.3 temp paths and macOS `/private/var`).
    * Add `TestValidateRule5RejectsBatchScriptsADSAndSanitizesErrorPaths`: verify `.bat`/`.cmd` rejection (both direct and via `PATH`), ADS `:stream` rejection, and control-character escaping in `validateCLIPath` error messages.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/config ./internal/agent/copilot -run "TestValidateRule5|TestNew|TestConfigReference"`

#### Unit 1.5 (`085.005-T`): Update backend architecture design doc and config reference for permission feedback sanitization, confusable `urls:` folding, emoji/ZWJ/PUA policy, and `cli_path` hardening (`2EE9F3ED`, `D3D0E9D5`, `F24EC587`, `2C45EFAE`, `2F1B6D4B`)

* **Domain**: `docs` (`size: S`, `complexity: low`)
* **Posture**: `migration-first`
* **Files affected** (2 files):
  * `docs/design-docs/intercom-go-backend-architecture.md`
  * `docs/config-reference.md`
* **Changes needed**:
  * In `docs/design-docs/intercom-go-backend-architecture.md`:
    * In §5.2: document that `indentURLsFraming` applies NFKC compatibility decomposition plus UTS #39 confusable-skeleton folding and skips interior invisible/blank/combining runes when matching `urls:`, and record the ratified `F24EC587` policy that emoji variation selectors (`U+FE0E`/`U+FE0F`, `U+E0100..U+E01EF`), ZWJ/ZWNJ (`U+200D`/`U+200C`), and private-use glyphs (`Co`) are intentionally replaced with `U+FFFD` / `<U+XXXX>` markers rather than preserved.
    * In §6.1: document that operator deny feedback (`Decision.Feedback`) is sanitized against control, newline, line/paragraph separator, format/bidi, invisible filler, and invalid UTF-8 runes in addition to the 1024-byte bound.
    * In §6.2 and §8: document the shared `pathsafe.ResolveCLIPath` helper (canonicalizing symlinks/junctions/mapped drives via `canonicalizeReparse`, rejecting UNC/device targets, NTFS ADS suffixes, `.bat`/`.cmd` batch scripts, and non-executable files on Unix) and cross-reference §7.2's verified empty-`cli_path` SDK behavior (`COPILOT_CLI_PATH` / embedded CLI / `PATH` gated by `Adapter.Start`'s `GetStatus` + `assertRuntime` version/protocol check).
  * In `docs/config-reference.md`:
    * Update the `[copilot]` table and Rule 5 validation section to document `pathsafe.ResolveCLIPath` canonicalization, `.bat`/`.cmd` and ADS rejection, Unix execute-bit validation, control-safe path formatting in error messages, and the empty-`cli_path` SDK resolution behavior (keeping all existing message substrings required by `TestConfigReferenceCoversValidationMessages`).
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/config -run TestConfigReference`

---

### Sub-Epic 2: Shipment `079-S` — Covering Feature `086-F` (Process lifecycle, credential hygiene, `localUI` supervision, and `config.Unconsumed` architecture)

#### Unit 2.1 (`086.001-T`): Convert `transport.Token` to a wipeable `[]byte` with `Wipe()` and `Clone()` (`AE317BD3`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (2 files):
  * `internal/transport/credential.go`
  * `internal/transport/credential_test.go`
* **Changes needed**:
  * In `internal/transport/credential.go`:
    * Change `type Token struct{ s string }` to `type Token struct{ b []byte }`.
    * Update `(t Token) Reveal() string` to `return string(t.b)`.
    * Add `func (t *Token) Wipe()` (`clear(t.b); t.b = nil`) to scrub the underlying byte slice in place and drop the slice reference.
    * Add `func (t Token) Clone() Token` that returns `Token{}` when `len(t.b) == 0` and otherwise returns a `Token` backed by a freshly allocated copy of `t.b` (`append([]byte(nil), t.b...)`), so callers holding a cloned snapshot cannot data-race with an in-place `Wipe()`.
    * In `(*CredentialStore).Mint`: encode into `tok := make([]byte, base64.RawURLEncoding.EncodedLen(len(raw)))` via `base64.RawURLEncoding.Encode(tok, raw[:])`, compute `digest := sha256.Sum256(tok)`, and construct `Issued{..., Token: Token{b: tok}, ...}`.
  * In `internal/transport/credential_test.go`:
    * Add `TestTokenWipeScrubsBackingBytesAndCloneIsIndependent`: mint a credential, capture `backing := iss.Token.b` and `cloned := iss.Token.Clone()`, call `iss.Token.Wipe()`, and assert that `iss.Token.Reveal() == ""`, every byte in `backing` is `0`, and `cloned.Reveal()` remains intact until `cloned.Wipe()` is called.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/transport -run "TestToken|TestCredential"`

#### Unit 2.2 (`086.002-T`): Render metadata-only partial `Reissue` credentials as `(already shown)` and surface one-time unconsumed-config-keys TUI notice (`5E4834CF`, `D953F24D`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/tui/credentials.go`
  * `internal/tui/model.go`
  * `internal/tui/credentials_test.go`
* **Changes needed**:
  * In `internal/tui/credentials.go`:
    * In `(m Model) credentialRows() []string`: when `c.Secret == ""` (a metadata-only entry whose plaintext was already shown and zeroed before a partial `Reissue` failure), render `SanitizeLine(c.Role) + ": (already shown)"` instead of an empty secret, while still rendering the `  fingerprint ..., expires ...` line below it.
    * In `(m Model) dismissCredentials() Model`: instead of unconditionally setting `m.notice = ""`, set `m.notice = m.unconsumedNotice` and `m.unconsumedNotice = ""` so a pending one-time unconsumed-keys notice is surfaced on the status line as soon as the operator dismisses the initial credential panel with `Esc`.
  * In `internal/tui/model.go`:
    * Add `UnconsumedKeys []string` to `Options` and `unconsumedNotice string` to `Model`.
    * In `New(o Options)`: when `len(o.UnconsumedKeys) > 0`, build `m.unconsumedNotice = "unconsumed config keys: " + strings.Join(sanitizedKeys, ", ")` (sanitizing each key name via `SanitizeLine`). If `len(o.Credentials.Credentials) == 0` (no initial credential panel open), set `m.notice = m.unconsumedNotice` and `m.unconsumedNotice = ""`.
  * In `internal/tui/credentials_test.go`:
    * Add `TestPartialReissueWithEmptySecretRendersAlreadyShownAndMetadata`: verify that a `ReissueFailedMsg` whose `Live` carries a replaced `approver` (with `Secret`) and an unreplaced `observer` (with `Secret: ""` and `Fingerprint: "obs-fp"`) renders `observer: (already shown)` and `fingerprint obs-fp` in `View()`.
    * Add `TestUnconsumedKeysNoticeSurfacesOnceAfterCredentialPanelDismissed`: verify that `New(Options{Credentials: initialCreds, UnconsumedKeys: []string{"commands", "retention_days"}})` shows the credential notice while the panel is open, surfaces `"unconsumed config keys: commands, retention_days"` on the status line when `Esc` dismisses the panel, and does not re-show it on subsequent updates.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/tui -run "TestPartialReissue|TestUnconsumedKeys|TestCredential"`

#### Unit 2.3 (`086.003-T`): Add `config.Unconsumed` with reflect-based schema completeness test and derive `cmd/intercom` runbook test from it (`864FDF7A`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/config/unconsumed.go`
  * `internal/config/unconsumed_test.go`
  * `cmd/intercom/runbook_test.go`
* **Changes needed**:
  * Create `internal/config/unconsumed.go`:
    * Define `func Unconsumed(cfg *Config) []string`: returns `nil` if `cfg == nil`; compares each accepted-but-unconsumed field of `cfg` against `def := Default()` — using `cfg.Database.Path != def.Database.Path` rather than `cfg.Database != def.Database` — and returns the sorted list of differing dotted TOML key names (`commands`, `database.path`, `max_concurrent_sessions`, `operator_detail_level`, `retention_days`, `stall.default_nudge_message`, `stall.enabled`, `stall.escalation_threshold_seconds`, `stall.inactivity_threshold_seconds`, `stall.max_retries`, `timeouts.prompt_seconds`, `timeouts.wait_seconds`).
    * Define `func UnconsumedKeyNames() []string`: returns the complete sorted slice of all 12 accepted-but-unconsumed dotted TOML key names (derived from `Unconsumed` on an all-non-default `Config`).
  * Create `internal/config/unconsumed_test.go`:
    * Add `TestEveryConfigSchemaKeyIsConsumedOrInUnconsumed`: walk `reflect.TypeOf(Config{})` using `collectTOMLKeyPaths` to collect all leaf and map TOML key paths on `Config`, verify that `UnconsumedKeyNames()` and the declared `consumedSchemaKeys` set are disjoint and their union equals the exact set of schema key paths on `Config`, and verify that mutating each field in `UnconsumedKeyNames()` (including `cfg.Database.Path`) causes `Unconsumed(cfg)` to report that exact key.
  * In `cmd/intercom/runbook_test.go`:
    * In `TestRunbookDocumentsLifecycleAndObservability`: replace the hand-maintained 12-element literal slice of unconsumed keys with `for _, key := range config.UnconsumedKeyNames()`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/config ./cmd/intercom -run "TestEveryConfigSchemaKey|TestUnconsumed|TestRunbook"`

#### Unit 2.4 (`086.004-T`): Harden `internal/app/credentials.go` with `zeroTokensLocked`/`clearLocked`, `mint` early-out on `credsCleared`, metadata-only partial `Reissue` entries, and disjoint `Token.Clone()` ownership across `Mint`, publication, `Reissue`, and `tuiReissue` (`4B7308F2`, `5E4834CF`, `AE317BD3`)

* **Domain**: `code` (`size: S`, `complexity: medium`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/app/credentials.go`
  * `internal/app/e2e_test.go`
  * `internal/app/review1_test.go`
* **Changes needed**:
  * In `internal/app/credentials.go`:
    * Extract `func (a *App) zeroTokensLocked()` (`for i := range a.creds { a.creds[i].Token.Wipe() }`), `func (a *App) clearCredsLocked()` (`a.credsCleared = true; a.zeroTokensLocked(); a.creds = nil`), `func cloneIssuedSlice(src []transport.Issued) []transport.Issued` (deep-cloning each `Issued.Token` via `Token.Clone()`), and `func wipeIssuedSlice(s []transport.Issued)` (wiping each `s[i].Token.Wipe()` and clearing `s`). Use `zeroTokensLocked()` and `clearCredsLocked()` in `clearCreds()`, `zeroPlaintext()`, and `mint()`'s in-flight shutdown/revocation branch (`4B7308F2`).
    * Enforce strict disjoint `Token` byte-slice ownership across `mint`, `a.creds`, `Reissue`, `mintAll`, and `tuiReissue`:
      * In `(*App).mint(ctx context.Context)`: check `a.mu.Lock(); cleared := a.credsCleared; a.mu.Unlock(); if cleared || a.shuttingDown() { return nil, a.revokedForShutdown(nil) }` at the start of `mint()` under `mintMu` (`4B7308F2`). When publishing each minted role into `a.creds` under `a.mu`, assign `a.creds = cloneIssuedSlice(next)` so `a.creds`'s stored `Token` backing arrays are strictly disjoint from the `next` slice returned to `mint`'s caller (`Reissue` / `tuiReissue` / `mintAll`). On error/shutdown abort inside `mint`, call `wipeIssuedSlice(next)` before returning.
      * In `(*App).mintAll() error`: `live, err := a.mint(context.Background()); wipeIssuedSlice(live); return err` (wiping the temporary returned slice immediately).
      * In `(*App).tuiReissue(ctx context.Context)`: `live, err := a.Reissue(ctx); defer wipeIssuedSlice(live)` so `live`'s temporary `Token` byte slices are deterministically wiped after `a.credentialsMsg(live)` extracts the `Secret` strings, and a concurrent `Shutdown` (`clearCreds()`) wiping `a.creds` can never race with `Reissue` callers or `tuiReissue` message construction.
    * In `(*App).Credentials() []transport.Issued`: return `cloneIssuedSlice(a.creds)` under `a.mu` so concurrent callers never share backing `[]byte` storage with `a.creds` (`AE317BD3`).
    * In `(*App).credentialsMsg(issued []transport.Issued) tui.CredentialsMsg`: skip an entry only when `secret == "" && iss.Fingerprint == ""`; when `secret == ""` and `iss.Fingerprint != ""` (an unreplaced role whose plaintext was already zeroed by `zeroPlaintext`), emit `tui.Credential{Role: roleName(iss.Role), Secret: "", Fingerprint: iss.Fingerprint, ExpiresAt: iss.ExpiresAt}` (`5E4834CF`).
    * Add package-internal test helpers `func (a *App) setAfterMint(fn func(hub.Role))` and `func tuiOptionsWithPlaintext(in io.Reader, out io.Writer, w, h int) TUIOptions`.
  * In `internal/app/e2e_test.go`:
    * Replace `o.TUI = TUIOptions{Input: in, Output: r.out, Width: 100, Height: 30, RetainPlaintext: true}` with `o.TUI = tuiOptionsWithPlaintext(in, r.out, 100, 30)`.
  * In `internal/app/review1_test.go`:
    * Update `a.afterMint = ...` in `TestReissuePausedAfterMintCannotRepublishAfterShutdown` to `a.setAfterMint(...)`.
    * Add `TestPartialTUIReissueAfterZeroPlaintextEmitsMetadataOnlyForUnreplacedRole`: verify that with default `zeroAfterTUI` (plaintext zeroed after initial TUI delivery), when `tuiReissue` fails on the second role (`observer`), `msg.Credentials` has length 2: `approver` with its new `Secret` and `observer` with `Secret == ""` and its live `Fingerprint` and `ExpiresAt`.
    * Add `TestMintEarlyOutsWhenCredsClearedWithoutShutReqClosed`: verify that after `a.clearCreds()`, `a.Reissue(context.Background())` immediately returns `ErrCredentialRevoked` without invoking `remote.Mint`.
    * Add `TestConcurrentReissueAndShutdownNeverRaceOnTokenBytes`: run concurrent `a.Reissue(ctx)`, `a.tuiReissue(ctx)`, and `a.Credentials()` readers while `a.Shutdown("race")` triggers `clearCreds()` under `-race`, verifying zero data races and empty post-shutdown `a.Credentials()`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/app -run "TestPartial|TestTUIReissue|TestMintEarlyOuts|TestConcurrentReissue|TestReissue|TestTUIDelivery"`

#### Unit 2.5 (`086.005-T`): Unexport `TUIOptions.retainPlaintext` (`AE317BD3`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/app/app.go`
  * `internal/app/assemble.go`
  * `internal/app/credentials_test.go`
* **Changes needed**:
  * In `internal/app/app.go`:
    * Rename `TUIOptions.RetainPlaintext` to unexported `retainPlaintext bool` so external callers cannot disable plaintext zeroing on the production API surface.
  * In `internal/app/assemble.go`:
    * Update `a.zeroAfterTUI = !o.TUI.RetainPlaintext` to `a.zeroAfterTUI = !o.TUI.retainPlaintext`.
  * In `internal/app/credentials_test.go`:
    * Update `newUIRig` to use `retainPlaintext: true` (and `tuiOptionsWithPlaintext` in `credentials.go` uses `retainPlaintext: true`).
    * Add `TestTUIOptionsHasNoExportedRetainPlaintextField`: inspect `reflect.TypeOf(TUIOptions{})` and assert no exported struct field is named `RetainPlaintext`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/app -run "TestTUIOptionsHasNoExportedRetainPlaintextField|TestCredentialsShownOnLocalTUIOnly|TestReissueThroughTUI"`

#### Unit 2.6 (`086.006-T`): Group `App` credential fields into `credentialState` mirroring `shutdownState` (`26E15024`)

* **Domain**: `code` (`size: S`, `complexity: low`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/app/assemble.go`
  * `internal/app/credentials.go`
  * `internal/app/reap_test.go`
* **Changes needed**:
  * In `internal/app/credentials.go`:
    * Define `type credentialState struct` holding `mintMu sync.Mutex`, `live []transport.Issued`, `cleared bool`, `zeroAfterTUI bool`, and `afterMint func(hub.Role)`, with methods `(*credentialState).zeroTokensLocked()` and `(*credentialState).clearLocked()`.
    * Update `mint`, `clearCreds`, `zeroPlaintext`, `tuiReissue`, `Credentials`, and `setAfterMint` on `*App` to operate on `a.creds` (`credentialState`).
  * In `internal/app/assemble.go`:
    * Replace the five loose credential fields on `App` (`mintMu`, `creds`, `credsCleared`, `zeroAfterTUI`, `afterMint`) with `creds credentialState`, and update `assemble` to set `a.creds.zeroAfterTUI = !o.TUI.retainPlaintext`.
  * In `internal/app/reap_test.go`:
    * Update `TestReapOnAPartialAppIsNilSafeAndRevokesCredentials` to initialize `creds: credentialState{live: []transport.Issued{{Fingerprint: "fp"}}}` and assert `a.creds.cleared` is true after `reap`, plus assert that `a.creds` is a `credentialState` struct grouping the credential lifecycle fields.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/app -run "TestReap|TestPartial|TestCredentials|TestShutdownClearsCredentials"`

#### Unit 2.7 (`086.007-T`): Move `teardown` beside `reap` in `shutdown.go`, call `clearCreds()` immediately on shutdown start in `supervise`, derive `maxApproval` from `config.MaxApprovalTimeoutSeconds`, and delegate `unconsumedKeys` to `config.Unconsumed` (`24CE7E1D`, `26E15024`, `864FDF7A`, `D953F24D`)

* **Domain**: `code` (`size: S`, `complexity: medium`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/app/assemble.go`
  * `internal/app/shutdown.go`
  * `internal/app/shutdown_test.go`
* **Changes needed**:
  * In `internal/app/assemble.go`:
    * Derive `const maxApproval = time.Duration(config.MaxApprovalTimeoutSeconds) * time.Second` (`D953F24D`).
    * Delegate `unconsumedKeys(cfg *config.Config) []string` to `return config.Unconsumed(cfg)` (`864FDF7A`).
    * Move `defaultTeardownPermissions` and `(*App).teardown(ctx context.Context) error` out of `assemble.go` into `internal/app/shutdown.go` beside `(*App).reap` (`26E15024`).
  * In `internal/app/shutdown.go`:
    * Place `defaultTeardownPermissions` and `(*App).teardown(ctx context.Context) error` directly beside `(*App).reap`.
    * In `(*App).supervise()`: call `a.clearCreds()` immediately after `<-a.shut.req` (in addition to `defer a.clearCreds()`) so `a.Credentials()` returns `nil` throughout the entire shutdown window (`24CE7E1D`).
  * In `internal/app/shutdown_test.go`:
    * Add `TestSuperviseClearsCredentialsImmediatelyOnShutdownStart`: block `fakeAgent.Stop` on a gate, call `a.Shutdown("test")`, wait until `agent.stop` is entered (while shutdown is still in progress and `a.Done()` is still open), and assert `len(a.Credentials()) == 0` and `a.creds.cleared == true`.
    * Add `TestSuperviseFallsBackToTeardownWhenNewShutdownFails`: configure an agent exposing `StopTimeout() time.Duration` (`5 * time.Second`) and `o.Timeouts = hub.ShutdownTimeouts{Force: 100 * time.Millisecond}` so `hub.NewShutdown` fails inside `supervise()`, and assert that `supervise()` falls back to `a.teardown`, stops the agent and hub, closes `a.Done()`, and surfaces `"shutdown executor:"` in `a.Err()`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/app -run "TestSupervise|TestSIGINT|TestTeardown|TestUnconsumed"`

#### Unit 2.8 (`086.008-T`): Move `watchUI`, `pumpRejects`, and `onReject` into `ui.go`, create `watchDone` inside `startUI`, drop redundant `quit`/`kill` nil defaults, wire `UnconsumedKeys` into `startUI`, and expose `App.RejectedUpgrades` (`7BFA4182`, `26E15024`, `D953F24D`)

* **Domain**: `code` (`size: S`, `complexity: medium`)
* **Posture**: `test-first`
* **Files affected** (3 files):
  * `internal/app/assemble.go`
  * `internal/app/ui.go`
  * `internal/app/tui_test.go`
* **Changes needed**:
  * In `internal/app/ui.go`:
    * Move `onReject`, `watchUI`, `defaultRejectNoticeInterval`, `pumpRejects`, and `pumpRejectsTo` from `assemble.go` into `ui.go` (`26E15024`).
    * Add `stopCh chan struct{}` alongside `watchDone chan struct{}` on `localUI`, both initialized at construction inside `startUI` / `startWatchedUI` (`ui := &localUI{done: make(chan struct{}), stopCh: make(chan struct{}), watchDone: make(chan struct{})}`).
    * On any early failure inside `startWatchedUI` before the watcher goroutine is spawned, close `ui.stopCh` and `ui.watchDone` before calling `_ = ui.stop()`. Once startup succeeds, spawn `go ui.watch(pump, onExit)` inside `startWatchedUI` before returning `ui, nil`.
    * Inside `(*localUI).watch(pump func(*localUI), onExit func(error))`: `defer close(u.watchDone)`; run `pump(u)` in a goroutine closing `pumpDone` (where `pumpRejectsTo` selects on `u.done` and `u.stopCh` for both mailbox reads and cooldown timer waits); wait on `select { case <-u.done: <-pumpDone; onExit(u.failure()) case <-u.stopCh: <-pumpDone }`. Because `tea.Program.Send` can only block forever if `prog.Run()` is wedged AND `kill()` failed to cancel `prog`'s internal context, on the normal path (`u.done` closed) `prog.Send` unblocks immediately, `<-pumpDone` and `onExit(u.failure())` complete deterministically, `u.watchDone` closes, and `(*localUI).stop()` unconditionally joins `<-u.watchDone` (with no `if u.watchDone != nil` branch). On the wedged anomaly path (`!waitClosed(u.done, killWait)`), `(*localUI).stop()` sets `u.stopErr`, closes `u.stopCh` (releasing `watch` and any idle/cooldown `pumpRejectsTo` wait), closes `u.feed`, and joins `u.watchDone` when `pumpDone` is not pinned inside a wedged `prog.Send` (`waitClosed(u.watchDone, killWait)`), preserving the hard upper bound on `stop()` while eliminating the post-construction nil-able `watchDone` assignment in `assemble` (`7BFA4182`).
    * In `(*localUI).stop()`: drop the redundant `if quit == nil` and `if kill == nil` fallback branches since `u.quit` and `u.kill` are always initialized at construction (`D953F24D`).
    * Pass `UnconsumedKeys` through `startWatchedUI` into `tui.Options{..., UnconsumedKeys: unconsumed}` (`D953F24D`).
  * In `internal/app/assemble.go`:
    * Call `startWatchedUI(ctx, h, o.TUI, creds, a.tuiReissue, config.Unconsumed(cfg), a.pumpRejects, a.onUIExit)` from `assemble` (removing the post-construction `ui.watchDone = make(chan struct{})` and `go a.watchUI(ui)` lines).
    * Add `func (a *App) RejectedUpgrades() uint64` returning `a.remote.RejectedUpgrades()` (0 when `a.remote == nil`), giving `(*transport.Remote).RejectedUpgrades()` an application-level caller (`D953F24D`).
  * In `internal/app/tui_test.go`:
    * Initialize `stopCh` and `watchDone` (closed by default) in `fakeUI` so `TestStopWaitsForRegisteredWatcher` can replace `u.watchDone = make(chan struct{})` to test waiting, replace `TestStopWithoutWatcherDoesNotWaitForIt` with `TestStartUICreatesWatchDoneAndStopAlwaysJoinsWatcher`, and verify in `TestWedgedTUIShutdownIsBoundedAndSurfacesInAppErr` that `<-a.ui.watchDone` is closed when `stop()` returns even when `a.ui.done` never closes.
    * Add `TestAppSurfacesUnconsumedKeysNoticeOnTUIAfterCredentialDismissalAndTracksRejectedUpgrades`: start an `App` with an unconsumed config key (`retention_days = 7`) and TUI enabled, dismiss the credential panel with `Esc`, verify `"unconsumed config keys: retention_days"` appears on the TUI output, and verify `a.RejectedUpgrades()` increments when a plain HTTP `GET` hits `a.WSURL()`.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./internal/app -run "TestStartUI|TestStop|TestRuntimeTUI|TestWedgedTUI|TestPumpRejects|TestOnReject|TestAppSurfacesUnconsumedKeys"`

#### Unit 2.9 (`086.009-T`): Update backend architecture design doc (§3.1) and operator runbook (§4, §7) (`5E4834CF`, `864FDF7A`, `D953F24D`)

* **Domain**: `docs` (`size: S`, `complexity: low`)
* **Posture**: `migration-first`
* **Files affected** (2 files):
  * `docs/design-docs/intercom-go-backend-architecture.md`
  * `docs/runbooks/intercom-operator-runbook.md`
* **Changes needed**:
  * In `docs/design-docs/intercom-go-backend-architecture.md` (§3.1 `Shutdown ordering (normative)`):
    * Add a dedicated paragraph documenting that failed-start `App.teardown` (and the `supervise` fallback when `hub.NewShutdown` errors) executes a reduced shutdown sequence tailored to a partially assembled process where no remote session was served or the executor could not be built: it clears credentials, runs the hub's bounded step-0 `BeginShutdown` to release any parked permission asks, stops the local UI and remote listener, stops the agent, and stops/awaits the hub last.
  * In `docs/runbooks/intercom-operator-runbook.md` (§4 `Credentials` and §7 `Unconsumed config keys`):
    * In §4: note that credential plaintext is held in a wipeable byte slice scrubbed in place after TUI delivery, and that if a `Ctrl+R` re-issue fails part-way, the panel shows the newly minted role's secret alongside `(already shown)` (with fingerprint and expiry) for any unreplaced role that remains live.
    * In §7: note that in addition to the startup `WARN` log (`config has accepted-but-unconsumed keys`), the local TUI surfaces a one-time `unconsumed config keys: ...` notice on the status line once the initial credential panel is dismissed.
* **Verification**:
  * `GOTOOLCHAIN=go1.26.10 go test -race -v ./cmd/intercom -run TestRunbook`

## Dependency Graph

### Shipment-Level DAG

* `078-S` (`labels: ["dag-root"]`, `predecessor_source: "declared_root"`, under explicit operator P-001 policy exception over `062-S`, code baseline `077-S` `f00ef564f49ff38a0ad9c3d9f756eb3a11edb59d`; Covering Feature `085-F`: `085.001-T -> 085.002-T -> 085.003-T -> 085.004-T -> 085.005-T`)
  * `<-blocks-- 079-S` (Covering Feature `086-F`: `086.001-T -> 086.002-T -> 086.003-T -> 086.004-T -> 086.005-T -> 086.006-T -> 086.007-T -> 086.008-T -> 086.009-T`)

### Intra-Shipment Task Dependencies

* **Within `078-S` (`085-F`)**: `085.001-T -> 085.002-T -> 085.003-T -> 085.004-T -> 085.005-T` (`085.004-T` consumes `pathsafe.ResolveCLIPath` and `pathsafe.SanitizeCLIPathForMessage` from `085.003-T`, and `085.005-T` updates the design doc and config reference after all code units land).
* **Within `079-S` (`086-F`)**: `086.001-T -> 086.002-T -> 086.003-T -> 086.004-T -> 086.005-T -> 086.006-T -> 086.007-T -> 086.008-T -> 086.009-T` (serialized so `086.004-T` consumes `Token.Wipe()`/`Token.Clone()` from `086.001-T`, `086.004-T`..`086.008-T` sequence edits across `internal/app` without file conflicts, `086.007-T`/`086.008-T` consume `config.Unconsumed` from `086.003-T` and `tui.Options.UnconsumedKeys` from `086.002-T`, and `086.009-T` updates docs last).

## Decisions and Rationale

1. **Two single-feature shipments (`078-S..079-S`)**: Complies with Ship Step 6.1(a1) `S2` (`n == 1` covering feature per shipment) while grouping permission/URL/summary sanitization and `cli_path` hardening into `078-S` (`085-F`) and process lifecycle, credential hygiene, `localUI` supervision, and `config.Unconsumed` into `079-S` (`086-F`).
2. **Pure-stdlib NFKC compatibility decomposition + UTS #39 confusable-skeleton folding (`D3D0E9D5`)**: Because `"urls:"` is a fixed 5-character ASCII prefix, `nfkcFoldURLsRune` + `foldURLsLookalikeRune` in `internal/agent/copilot/permission.go` folds all NFKC compatibility decompositions (fullwidth, mathematical alphanumeric alphabets/digits, enclosed/parenthesized/circled forms, superscripts/subscripts, Roman numerals, compatibility colons) and UTS #39 homoglyphs while skipping interior invisible/blank/combining runes, with zero new direct dependencies in `go.mod`.
3. **Fail-closed replacement of emoji variation selectors, ZWJ/ZWNJ, and PUA runes (`F24EC587`)**: Keeps `unicode.Variation_Selector`, `unicode.Cf`, and `unicode.Co` classified as hidden/invisible in both `internal/agent/copilot` and `internal/tui`, blocking Plane-14 variation-selector smuggling and zero-width/blank-glyph padding on permission modals.
4. **Shared `pathsafe.ResolveCLIPath` calling `canonicalizeReparse` (`2C45EFAE`, `2F1B6D4B`)**: Reuses `pathsafe`'s existing `canonicalizeReparse` (`GetFinalPathNameByHandleW` on Windows, `filepath.EvalSymlinks` on POSIX) without adding any `pathsafe.Root.Resolve` call site or write primitive, keeping `tools/gatecheck/internal/writepath` 100% green.
5. **Wipeable `[]byte` `transport.Token` with `Clone()` in `App.Credentials()` (`AE317BD3`)**: Scrubs credential plaintext in place via `clear(t.b)` when `zeroTokensLocked()` or `clearLocked()` runs, while `Token.Clone()` ensures concurrent readers of `App.Credentials()` never race with in-place zeroing under `go test -race`.

## Risks and Caveats

* **`TestHubPackageHasNoBareChannelSends` AST guard (`internal/hub/degraded_test.go`)**: Every channel send in `internal/hub/*.go` production files must remain inside a `select` statement (or `trySend`).
* **`TestConfigReferenceCoversValidationMessages` (`internal/config/docs_test.go`)**: `validateCLIPath` retains `'...'` framing around `pathsafe.SanitizeCLIPathForMessage(cliPath)` so all existing message substring checks in `docs_test.go` and `copilotcli_test.go` pass unchanged.
* **Windows 8.3 short-name expansion in `TestValidateRule5RejectsDirectoryCLIPathAndAcceptsRegularFile` (`internal/config/copilotcli_test.go`)**: Creating the test executable inside `cfg.DefaultWorkspaceRoot` (already canonicalized by `pathsafe.NewRoot`) ensures `cfg.Copilot.CLIPath == file` holds on Windows hosts whose `%TEMP%` uses 8.3 short names.
* **Retired-architecture identifier gate (`tools/gatecheck/internal/retiredarch`)**: Never use any retired token (such as `slack`) in Go identifiers; note `testSlack` already exists only in `_test.go` files (which `retiredarch` excludes under `internal/`), and no new identifier may use retired tokens.

## Plan Hardening Signals

* **public API, schema, or contract change**: **PRESENT** — Adds `pathsafe.CLIPathErrorKind`, `pathsafe.ResolveCLIPath`, `pathsafe.SanitizeCLIPathForMessage`, `(*transport.Token).Wipe`, `(transport.Token).Clone`, `tui.Options.UnconsumedKeys`, `config.Unconsumed`, `config.UnconsumedKeyNames`, `(*app.App).RejectedUpgrades`, and unexports `app.TUIOptions.RetainPlaintext` (`retainPlaintext`).
* **security, auth, permission, or compliance-sensitive behavior**: **PRESENT** — Hardens operator deny feedback sanitization (`2EE9F3ED`), `urls:` NFKC/confusable-skeleton anti-spoofing and emoji/ZWJ/PUA summary sanitization (`D3D0E9D5`, `F24EC587`), `copilot.cli_path` canonicalization and batch/ADS/execute-bit validation (`2C45EFAE`, `2F1B6D4B`), and in-memory credential plaintext scrubbing and shutdown clearing (`24CE7E1D`, `5E4834CF`, `AE317BD3`, `4B7308F2`).
* **migration, backfill, destructive data/config action, or irreversible step**: **ABSENT** — No data migration or destructive operation.
* **external integration, operator checkpoint, or external dependency**: **ABSENT** — Zero new external dependencies in `go.mod`.
* **high runtime, rollout, or rollback risk**: **PRESENT** — Restructures `internal/app` TUI watcher supervision (`watchDone` inside `startUI`), failed-start `teardown`, and `supervise` credential clearing.

Requires plan hardening: yes

## Runtime Verification and Closure

* **Shipment `078-S` (`085-F`)**: Permission feedback sanitization, confusable URL defense, emoji/ZWJ summary policy, and Copilot `cli_path` resolution hardening. Verified by `GOTOOLCHAIN=go1.26.10 go test -race ./internal/hub/... ./internal/agent/... ./internal/pathsafe/... ./internal/config/... ./internal/tui/... ./internal/transport/...`, `GOTOOLCHAIN=go1.26.10 go vet ./...`, `GOTOOLCHAIN=go1.26.10 golangci-lint run`, `gofmt -l .`, `GOTOOLCHAIN=go1.26.10 go run ./tools/gatecheck write-path --self-test --root .`, and `GOTOOLCHAIN=go1.26.10 go run ./tools/gatecheck retired-arch --self-test --root .`.
* **Shipment `079-S` (`086-F`)**: Process lifecycle, credential hygiene, `localUI` supervision, and `config.Unconsumed` architecture. Verified by `GOTOOLCHAIN=go1.26.10 go test -race ./...`, `GOTOOLCHAIN=go1.26.10 go vet ./...`, `GOTOOLCHAIN=go1.26.10 golangci-lint run`, `gofmt -l .`, `GOTOOLCHAIN=go1.26.10 go run ./tools/gatecheck write-path --self-test --root .`, `GOTOOLCHAIN=go1.26.10 go run ./tools/gatecheck retired-arch --self-test --root .`, and `go run ./cmd/intercom --help`. Rollback for each shipment is a clean `git revert` of its merge commit.

## Plan Hardening

### 1. Why Hardening Was Required and Context Consulted

Hardening was triggered by three signals: (1) cross-package API additions and tightening across `internal/pathsafe`, `internal/config`, `internal/transport`, `internal/tui`, and `internal/app` (including unexporting `TUIOptions.RetainPlaintext`), (2) security-sensitive sanitization, executable path validation, and in-memory credential scrubbing (`2EE9F3ED`, `D3D0E9D5`, `F24EC587`, `2C45EFAE`, `2F1B6D4B`, `24CE7E1D`, `5E4834CF`, `AE317BD3`, `4B7308F2`), and (3) `internal/app` supervision and teardown lifecycle changes (`7BFA4182`, `26E15024`, `D953F24D`).

**Reinforced instructions and compound learnings consulted**:

* `.github/instructions/constitution.instructions.md` (Principles I, II, III, IV, V, VI, VII, VIII, XI)
* `.github/instructions/strict-safety.instructions.md` (`ProposedAction` / `ActionRisk` / `ActionResult` classification)
* `docs/compound/2026-10-20-process-wiring-credential-panel-and-approval-disarm.md`
* `docs/compound/2026-10-20-tui-permission-modal-safety-and-linux-test-clock.md`
* `docs/compound/2026-10-20-backlogit-cli-pitfalls-and-permission-outcome-lessons.md`
* `docs/compound/2026-09-30-go-msys-bash-bridging-and-unicode-case-mapping-gap.md`
* `docs/compound/2026-09-08-verify-go-stdlib-internals-claims-against-goroot-source.md`
* `docs/compound/concurrency-issues/prefer-atomic-typed-fields-2026-09-04.md`
* `docs/compound/workflow-issues/cross-artifact-contract-closure-requires-every-surface-2026-09-13.md`

### 2. Protected Invariants

1. **Zero new production `pathsafe.Root.Resolve` call sites and zero filesystem write primitives (`tools/gatecheck/internal/writepath`)**: `pathsafe.ResolveCLIPath` calls `canonicalizeReparse` directly inside `internal/pathsafe` and never calls `Root.Resolve` (`reviewedResolveCallers` in `tools/gatecheck/internal/writepath/writepath.go` remains pinned to `internal/app/workspace.go` with `calls: 1`). `GOTOOLCHAIN=go1.26.10 go run ./tools/gatecheck write-path --self-test --root .` must exit 0 after every task.
2. **Retired-architecture gate (`tools/gatecheck/internal/retiredarch`) and `.golangci.yml` `depguard` boundaries**: No new identifier in `cmd/**` or `internal/**` may contain a retired token; `internal/pathsafe` imports only stdlib and `internal/apperr`; `internal/tui` never imports `internal/agent/copilot` or `copilot-sdk`.
3. **Config validation receiver non-mutation on failure (Protected Invariant I6)**: `(*Config).Validate()` commits the canonicalized `resolvedCLIPath` to `c.Copilot.CLIPath` only in the deferred-commit block after all validation rules pass.
4. **Race-free credential byte scrubbing (`go test -race`)**: `transport.Token` stores its secret in `b []byte` wiped in place via `(*Token).Wipe()` (`clear(t.b); t.b = nil`), while `cloneIssuedSlice` (using `Token.Clone()`) ensures `a.creds.live`, `(*App).Credentials()`, and `(*App).mint` / `(*App).Reissue` return values hold strictly disjoint `[]byte` backing slices, and `mintAll()` / `tuiReissue()` deterministically wipe their temporary slices (`wipeIssuedSlice`).
5. **Construction-owned `watchDone` join and bounded wedged-TUI stop in `(*localUI).stop()`**: `startWatchedUI` initializes `ui.stopCh` and `ui.watchDone` at `localUI` construction, closes both before `_ = ui.stop()` if startup fails before the watcher goroutine launches, and spawns `go ui.watch(pump, onExit)` before returning `ui, nil`. On the normal path (`u.done` closed), `(*localUI).stop()` closes `u.feed` and unconditionally joins `<-u.watchDone` (with no nil check). On the wedged anomaly path (`!waitClosed(u.done, killWait)`), `(*localUI).stop()` sets `u.stopErr`, closes `u.stopCh` (releasing `watch` and `pumpRejectsTo`), closes `u.feed`, and bounds its wait on `u.watchDone` so a wedged `tea.Program` can never hang shutdown.

### 3. Strict-Safety Action Classifications (`ProposedAction` / `ActionRisk` / `ActionResult`)

| Action ID | `ProposedAction` Summary & Targets | `change_kind` | `ActionRisk` | `approval_required` | Rollback | `ActionResult` |
|---|---|---|---|---|---|---|
| `PA-01` | Guard `truncUTF8` against `n <= 0` and sanitize operator deny feedback in `internal/hub` and `internal/agent/copilot`; replace `urls:` lookalike denylist with NFKC + UTS #39 confusable-skeleton folding and interior-invisible tolerance; lock fail-closed emoji/ZWJ/PUA sanitization symmetry; extract `pathsafe.ResolveCLIPath` and `pathsafe.SanitizeCLIPathForMessage` with `canonicalizeReparse`, ADS/`.bat`/`.cmd` rejection, and Unix execute-bit check; wire into `internal/config` and `internal/agent/copilot`; and update design doc and config reference (`078-S` / `085-F`). | `local edit` (`internal/hub`, `internal/agent/copilot`, `internal/pathsafe`, `internal/config`, `internal/tui`, `docs/`) | `high` | `false` (pre-approved under `DARK_MODE_ACTIVE` Option 2 scope) | `git revert` of `078-S` merge commit | `planned` |
| `PA-02` | Convert `transport.Token` to wipeable `[]byte` with `Wipe()`/`Clone()`; render metadata-only partial `Reissue` credentials as `(already shown)` and surface one-time `UnconsumedKeys` notice in `internal/tui`; add `config.Unconsumed` with reflect schema completeness test and derive `cmd/intercom/runbook_test.go`; unexport `TUIOptions.retainPlaintext`, group credentials into `credentialState`, extract `zeroTokensLocked`/`clearLocked`, early-out `mint` on `creds.cleared`, call `clearCreds()` immediately on shutdown start in `supervise`, move `teardown` to `shutdown.go` and `watchUI`/`pumpRejects`/`onReject` to `ui.go` with `watchDone` inside `startUI`, and update design §3.1 and operator runbook (`079-S` / `086-F`). | `local edit` (`internal/transport`, `internal/tui`, `internal/config`, `internal/app`, `cmd/intercom`, `docs/`) | `high` | `false` (pre-approved under `DARK_MODE_ACTIVE` Option 2 scope) | `git revert` of `079-S` merge commit | `planned` |

### 4. Deepened Verification, Rollback, and Review-Gate Requirements

* **Per-task verification gates**:
  * Run `GOTOOLCHAIN=go1.26.10 go test -race ./...` and `GOTOOLCHAIN=go1.26.10 go vet ./...` after each code unit.
  * Run `GOTOOLCHAIN=go1.26.10 golangci-lint run`, `GOTOOLCHAIN=go1.26.10 go run ./tools/gatecheck write-path --self-test --root .`, and `GOTOOLCHAIN=go1.26.10 go run ./tools/gatecheck retired-arch --self-test --root .` after each shipment.
* **Review-gate capability requirement (P-012)**: Plan review (Step 4) MUST dispatch the required review personas (`Constitution Reviewer`, `Scope Boundary Auditor`, `Architecture Strategist`, `Security Lens Reviewer`, `Go Reviewer`, `Learnings Researcher`) and append a `## Plan Review` section with literal `dispatch_mode:` and `decision:` markers before harvest.
* **Unresolved operator decisions**: None.

<!-- plan-review-attempt: 1 -->

## Plan Review

```text
dispatch_mode: multi-agent
decision: PASS
```

### Gate Decision and Rationale

* **Verdict**: **PASS** (after incorporating all P1, P2, and P3 findings from the multi-agent review pass — including the Anchor Reviewer `Architecture Strategist` on `gpt-6.1-sol` / `high` — directly into the plan and deliberation artifacts).
* **Plan Hardening Requirement**: `Requires plan hardening: yes` — **SATISFIED** (`## Plan Hardening` section present with `ProposedAction` / `ActionRisk` / `ActionResult` classifications `PA-01`..`PA-02` and protected invariants).

### Persona Coverage Table

| Persona | Mode | Model Route | Initial Verdict | Final Verdict (Post-Remediation) |
|---|---|---|---|---|
| **Constitution Reviewer** | Subagent | Default (`gemini-4-argon`) | PASS (0 P0, 0 P1, 0 P2, 1 P3) | **PASS** |
| **Go Reviewer** | Inline Rubric Pass (`TOOL_DEGRADED: go-reviewer-subagent-empty-response`) | Caller (`gemini-4-argon`) | PASS (0 P0, 0 P1, 0 P2, 0 P3) | **PASS** |
| **Scope Boundary Auditor** | Subagent | Default (`gemini-4-argon`) | ADVISORY (0 P0, 0 P1, 0 P2, 2 P3) | **PASS** |
| **Learnings Researcher** | Subagent | Default (`claude-haiku-5.5`) | PASS (12 high-confidence solutions verified) | **PASS** |
| **Architecture Strategist** | Anchor Subagent | Anchor (`gpt-6.1-sol` / `high`) | FAIL -> Remediated (2 P1, 2 P2) | **PASS** |
| **Security Lens Reviewer** | Subagent | Default (`gemini-4-argon`) | PASS (0 P0, 0 P1, 0 P2, 0 P3) | **PASS** |

### Findings and Applied Remediations

1. **[P1 — Remediated] Unit 2.1 / Unit 2.4 (`086.001-T`, `086.004-T`) Disjoint `Token` Byte-Slice Ownership Across `Mint`, Publication, `Reissue`, and `tuiReissue` (`Architecture Strategist`)**:
   * *Finding*: Shallow-copying `[]transport.Issued` when publishing into `a.creds` inside `mint()` shares `Token.b` backing arrays with the `next` slice returned to `Reissue` / `tuiReissue` / `mintAll`, allowing a concurrent `Shutdown` (`clearCreds()`) to race with `Reissue` consumers or `tuiReissue` message construction.
   * *Remediation*: Updated Unit 2.4 (`086.004-T`) and Protected Invariant 4 so `mint()` publishes `a.creds = cloneIssuedSlice(next)` (giving `a.creds` and the returned `next` slice strictly disjoint `Token.b` backing arrays), `mintAll()` and `tuiReissue()` deterministically wipe their temporary returned slices via `wipeIssuedSlice(live)`, and `TestConcurrentReissueAndShutdownNeverRaceOnTokenBytes` verifies concurrent `Reissue`, `tuiReissue`, `Credentials`, and `Shutdown` under `-race`.
2. **[P1 — Remediated] Unit 2.8 (`086.008-T`) Bounded Supervision Cancellation (`stopCh`) for Unconditional `watchDone` Join on Wedged TUI (`Architecture Strategist`)**:
   * *Finding*: Joining `<-u.watchDone` unconditionally in `(*localUI).stop()` would hang when `tea.Program` is wedged (`u.done` never closes after `Quit` and `Kill`) unless `watch` and `pumpRejects` have an independent stop signal closed when `stop()`'s `stopBudget + killWait` expires.
   * *Remediation*: Updated Unit 2.8 (`086.008-T`) and Protected Invariant 5 to add `stopCh chan struct{}` on `localUI` (closed by `stop()` after the `u.done` wait finishes, whether normal or wedged), with `(*localUI).watch` closing `pumpStop` when either `u.done` or `u.stopCh` closes and `(*localUI).stop()` unconditionally joining `<-u.watchDone` on both normal and wedged paths.
3. **[P2 — Remediated] Unit 1.2 (`085.002-T`) Bounded Single-Rune `"urls:"` Prefix Detector Contract (`Architecture Strategist`)**:
   * *Finding*: A rune-to-rune helper (`nfkcFoldURLsRune`) is a targeted single-codepoint compatibility/skeleton folder for the 5 ASCII bytes `'u'`, `'r'`, `'l'`, `'s'`, `':'`, not a general arbitrary-string NFKC normalizer.
   * *Remediation*: Clarified Unit 1.2 (`085.002-T`) to specify `nfkcFoldURLsRune` + `foldURLsLookalikeRune` as a bounded Unicode 16.0 compatibility and UTS #39 single-codepoint skeleton detector for the 5-character `"urls:"` prefix.
4. **[P2 — Remediated] Explicit `golangci-lint` (`depguard`) Verification (`Architecture Strategist`)**:
   * *Finding*: `go vet` and `gatecheck` do not execute `.golangci.yml` `depguard` rules.
   * *Remediation*: Added `GOTOOLCHAIN=go1.26.10 golangci-lint run` (plus `internal/transport` conformance test coverage) to the shipment verification commands and per-task verification gates.
5. **[P3 — Remediated] Deliberation Task-Count Consistency (`Constitution Reviewer` & `Scope Boundary Auditor`)**:
   * *Finding*: Two summary table cells in the deliberation referenced `11 atomic tasks` / `086.006-T` instead of `14 atomic tasks` / `086.009-T`.
   * *Remediation*: Updated both cells in `docs/decisions/2026-10-31-intercom-go-permission-sanitization-cli-path-and-app-lifecycle-deliberation.md` to match the 14-unit breakdown (`085.001-T`..`085.005-T` and `086.001-T`..`086.009-T`).


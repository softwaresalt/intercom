---
title: "Permission Sanitization, Confusable URL Defense, CLI Path Trust, and Process Lifecycle Hardening"
description: "Deliberation over 13 deferred scope expansion entries across Capability Set 1 (2EE9F3ED, D3D0E9D5, F24EC587, 2C45EFAE, 2F1B6D4B) and Capability Set 2 (24CE7E1D, 5E4834CF, AE317BD3, 4B7308F2, 7BFA4182, 26E15024, 864FDF7A, D953F24D)"
topic: "Permission feedback sanitization, confusable URL defense, emoji/ZWJ summary policy, Copilot cli_path resolution consolidation and residual hardening, process lifecycle credential hygiene, localUI supervision, and config.Unconsumed schema unification"
depth: "deep"
decision_status: "decided"
promoted_to: "plan"
linked_artifacts:
  - "docs/plans/2026-10-31-intercom-go-permission-sanitization-cli-path-and-app-lifecycle-plan.md"
tags:
  - "security"
  - "permissions"
  - "cli-path"
  - "credentials"
  - "lifecycle"
  - "tui"
  - "config"
---

## Problem Frame

This deliberation resolves 13 `DEFERRED SCOPE EXPANSION` stash entries selected under Option 2 (Capability Set 1 + Capability Set 2) during `DARK_MODE_ACTIVE` (P-017) staging. Per P-021 C6, every entry carrying the `DEFERRED SCOPE EXPANSION` marker is routed through formal deliberation before implementation planning, and undergoes both Trigger A (unconditional duplicate detection) and Trigger B (late-identifier reconciliation).

Because Ship Step 6.1(a1) S2 enforces a single covering feature per shipment (`n == 1`), the 13 entries are structured into two cohesive covering features and two sequential shipments:

* **Capability Set 1 — Covering Feature `085-F` / Shipment `078-S`**:
  * `2EE9F3ED`: Sanitize operator deny feedback (control, bidi, newline, invalid UTF-8 runes) before delivery to the agent, and add a non-positive-`n` guard to `truncUTF8`.
  * `D3D0E9D5`: Replace the hand-written `urls:` homoglyph denylist with NFKC plus confusable-skeleton folding, and tolerate interior invisible runes.
  * `F24EC587`: Decide whether summary sanitization should preserve legitimate emoji variation selectors (`U+FE0F`/`U+FE0E`), ZWJ/ZWNJ sequences, and private-use icon glyphs instead of replacing them with `U+FFFD`.
  * `2C45EFAE`: Consolidate the duplicated bare-name and absolute-path resolution steps (`LookPath`, `filepath.Abs`, re-classify as `CLIPathAbs`, `os.Stat`, `IsRegular`) of `config` `validateCLIPath` and `copilot` `resolveCLIPath` into shared `pathsafe` helpers, keeping only policy differences in each caller.
  * `2F1B6D4B`: `cli_path` residual hardening beyond lexical classification: canonicalize and resolve symlinks/junctions/mapped drives (rejecting UNC/device targets and NTFS ADS suffixes), reject `.bat`/`.cmd` batch scripts on Windows `PATHEXT` or explicit paths, enforce execute-bit check on Unix, prevent raw control/quote injection in `config` `cli_path` error messages, and document the empty `cli_path` SDK resolution contract.

* **Capability Set 2 — Covering Feature `086-F` / Shipment `079-S` (blocked by `078-S`)**:
  * `24CE7E1D`: Call `a.clearCreds()` immediately after `shutReq` closes in `App.supervise` (not only via `defer` at the end) so revoked credentials are not returned by `Credentials()` during the shutdown window.
  * `5E4834CF`: Emit metadata-only credential entries (role, fingerprint, expiry, empty `Secret`) when `zeroPlaintext` has already cleared an unreplaced role's secret prior to a partial `Reissue` failure, and render empty-secret entries in the TUI credential panel as `(already shown)`.
  * `AE317BD3`: Unexport `TUIOptions.RetainPlaintext` (`retainPlaintext`) so it is not on the production surface, and hold `transport.Token` plaintext as a wipeable `[]byte` with an explicit `Wipe()` method.
  * `4B7308F2`: Extract shared locked zeroing helpers (`zeroTokensLocked` / `clearLocked`) on `credentialState` for the thrice-duplicated zero loop (`clearCreds`, `mint` in-flight branch, `zeroPlaintext`) and make `mint()` early-out on `creds.cleared` (not only `shuttingDown`) under `mu`.
  * `7BFA4182`: Move `pumpRejects` and the TUI exit watcher from `App.assemble` into `localUI` (`internal/app/ui.go`) and create `watchDone` inside `startUI` so `stop()` always joins everything it owns without a post-construction `watchDone` assignment.
  * `26E15024`: Reorganize `internal/app` files — move `watchUI`/`pumpRejects`/`onReject` from `assemble.go` into `ui.go`, move `teardown` beside `reap` in `shutdown.go`, and group credential fields into a `credentialState` struct mirroring `shutdownState`.
  * `864FDF7A`: Move the accepted-but-unconsumed config key list to `config.Unconsumed(cfg)` with a reflect-based test that every `Config` field is consumed or listed, compare `cfg.Database.Path` rather than the whole struct, and derive the runbook test from `config.Unconsumed`.
  * `D953F24D`: Document in design §3.1 that failed-start `teardown` is a reduced shutdown sequence, add a test for the `supervise` to `teardown` fallback when `NewShutdown` errors, surface the unconsumed-config-keys warning as a one-time TUI notice on credential-panel dismissal, give `Remote.RejectedUpgrades()` an `App.RejectedUpgrades()` caller, drop the redundant `quit`/`kill` nil defaults in `localUI.stop`, and derive `maxApproval` from `config.MaxApprovalTimeoutSeconds`.

### P-021 C5/C6 Triage Audit (Trigger A & Trigger B)

All 13 entries were audited under P-021 C5/C6:

* **Trigger A (Unconditional Duplicate Scan)**: Evaluated across all 138 active entries in `.backlogit/stash.jsonl`, all archived entries in `.backlogit/archive/stash.jsonl`, and `.backlogit/queue/` (0 items). None of the 13 entries carry a `DISCOVERY-STATUS: AMBIGUOUS` or `DISCOVERY-STATUS: LOOKUP-UNAVAILABLE` token. Eleven entries (`2EE9F3ED`, `D3D0E9D5`, `F24EC587`, `2C45EFAE`, `2F1B6D4B`, `24CE7E1D`, `5E4834CF`, `AE317BD3`, `4B7308F2`, `864FDF7A`, `D953F24D`) are `CLEAN` with no overlapping entry. Entries `7BFA4182` (captured in `065-S`) and `26E15024` (captured in `066-S`) have a `COMPLEMENTARY OVERLAP` on moving `watchUI`/`pumpRejects`/`onReject` into `internal/app/ui.go`, while each also specifies distinct scope (`7BFA4182` specifies creating `watchDone` and starting the watcher inside `startUI`; `26E15024` specifies moving `teardown` to `shutdown.go` and introducing `credentialState`). Both are reconciled into unified tasks in `086-F` and both will be archived at Step 5.6.
* **Trigger B (Late-Identifier Reconciliation)**: Executed in place on `.backlogit/stash.jsonl` after the Step 1.9 branch gate passed on `chore/stage-permission-cli-path-and-app-lifecycle`.

| Stash ID | Originating Shipment / Feature / Tasks | Closure Record | Recovered PR & Merge Commit | Review-Thread ID | Trigger A Outcome |
|---|---|---|---|---|---|
| `2EE9F3ED` | `076-S` / `083-F` / `083.001-T`..`083.004-T` | `docs/closure/076-S-083-F-post-merge-closure.md` | PR `#228`, merge `e15937714e489f6c55485651029db7cee031b91f` (HEAD `cf7aa865b96b5ab02ada869c9c95926f5c89d3c3`) | `N/A` (terminal: pre-PR local review) | `CLEAN` |
| `D3D0E9D5` | `076-S` / `083-F` / `083.001-T`..`083.004-T` | `docs/closure/076-S-083-F-post-merge-closure.md` | PR `#228`, merge `e15937714e489f6c55485651029db7cee031b91f` (HEAD `cf7aa865b96b5ab02ada869c9c95926f5c89d3c3`) | `N/A` (terminal: pre-PR local review) | `CLEAN` |
| `F24EC587` | `076-S` / `083-F` / `083.001-T`..`083.004-T` | `docs/closure/076-S-083-F-post-merge-closure.md` | PR `#228`, merge `e15937714e489f6c55485651029db7cee031b91f` (HEAD `cf7aa865b96b5ab02ada869c9c95926f5c89d3c3`) | `N/A` (terminal: pre-PR local review) | `CLEAN` |
| `2C45EFAE` | `075-S` / `082-F` / `082.002-T`, `082.003-T` | `docs/closure/075-S-082-F-post-merge-closure.md` | PR `#225`, merge `a2898942a12c70fe23be050d698ae4622153d97f` (HEAD `c019210324abcadbc1750dd3a1f552413aaf6f2b`) | `N/A` (terminal: pre-PR local review) | `CLEAN` |
| `2F1B6D4B` | `075-S` / `082-F` / `082.002-T`, `082.003-T` | `docs/closure/075-S-082-F-post-merge-closure.md` | PR `#225`, merge `a2898942a12c70fe23be050d698ae4622153d97f` (HEAD `c019210324abcadbc1750dd3a1f552413aaf6f2b`) | `N/A` (terminal: pre-PR local review) | `CLEAN` |
| `24CE7E1D` | `065-S` / `072-F` / `072.003-T` | `docs/closure/065-S-072-F-post-merge-closure.md` | PR `#195`, merge `427bfbd5a3fcb22e117914c869a7faf2ab3d06c5` (HEAD `17471b7743d6dcfc05dd655e77d67a93165b2f73`) | `N/A` (terminal: pre-PR local review) | `CLEAN` |
| `5E4834CF` | `065-S` / `072-F` / `072.003-T` | `docs/closure/065-S-072-F-post-merge-closure.md` | PR `#195`, merge `427bfbd5a3fcb22e117914c869a7faf2ab3d06c5` (HEAD `17471b7743d6dcfc05dd655e77d67a93165b2f73`) | `N/A` (terminal: pre-PR local review) | `CLEAN` |
| `AE317BD3` | `065-S` / `072-F` / `072.003-T` | `docs/closure/065-S-072-F-post-merge-closure.md` | PR `#195`, merge `427bfbd5a3fcb22e117914c869a7faf2ab3d06c5` (HEAD `17471b7743d6dcfc05dd655e77d67a93165b2f73`) | `N/A` (terminal: pre-PR local review) | `CLEAN` |
| `4B7308F2` | `066-S` / `073-F` / `073.001-T`..`073.003-T` | `docs/closure/066-S-073-F-post-merge-closure.md` | PR `#198`, merge `65780da0d83a53182a56460093b80cc132b5d945` (HEAD `d588c7e4a558393632a72111ba252b8bef1d1520`) | `N/A` (terminal: pre-PR local review) | `CLEAN` |
| `7BFA4182` | `065-S` / `072-F` / `072.001-T`, `072.005-T` | `docs/closure/065-S-072-F-post-merge-closure.md` | PR `#195`, merge `427bfbd5a3fcb22e117914c869a7faf2ab3d06c5` (HEAD `17471b7743d6dcfc05dd655e77d67a93165b2f73`) | `N/A` (terminal: pre-PR local review) | `COMPLEMENTARY OVERLAP` with `26E15024` |
| `26E15024` | `066-S` / `073-F` / `073.001-T`..`073.003-T` | `docs/closure/066-S-073-F-post-merge-closure.md` | PR `#198`, merge `65780da0d83a53182a56460093b80cc132b5d945` (HEAD `d588c7e4a558393632a72111ba252b8bef1d1520`) | `N/A` (terminal: pre-PR local review) | `COMPLEMENTARY OVERLAP` with `7BFA4182` |
| `864FDF7A` | `065-S` / `072-F` / `072.002-T` | `docs/closure/065-S-072-F-post-merge-closure.md` | PR `#195`, merge `427bfbd5a3fcb22e117914c869a7faf2ab3d06c5` (HEAD `17471b7743d6dcfc05dd655e77d67a93165b2f73`) | `N/A` (terminal: pre-PR local review) | `CLEAN` |
| `D953F24D` | `065-S` / `072-F` / `072.001-T`..`072.006-T` | `docs/closure/065-S-072-F-post-merge-closure.md` | PR `#195`, merge `427bfbd5a3fcb22e117914c869a7faf2ab3d06c5` (HEAD `17471b7743d6dcfc05dd655e77d67a93165b2f73`) | `N/A` (terminal: pre-PR local review) | `CLEAN` |

## Research Findings

### Institutional Learnings (`docs/compound/`)

* `docs/compound/2026-10-20-process-wiring-credential-panel-and-approval-disarm.md`: Initial credentials are passed through `tui.Options.Credentials` at construction; re-issue updates use `reissuedMsg` and `ReissueFailedMsg` so the in-flight guard and cleared-plaintext invariants are never overwritten by stale messages.
* `docs/compound/2026-10-20-tui-permission-modal-safety-and-linux-test-clock.md`: Sanitization applies strictly to display and permission feedback boundaries, never to raw user prompt input.
* `docs/compound/2026-09-30-go-msys-bash-bridging-and-unicode-case-mapping-gap.md` and `docs/compound/2026-09-08-verify-go-stdlib-internals-claims-against-goroot-source.md`: On Windows, `t.TempDir()` can return 8.3 short names (`DEWILL~1`) whereas `GetFinalPathNameByHandleW` (`canonicalizeReparse`) returns expanded long names; `filepath.EvalSymlinks` does not resolve Windows directory junctions (`IO_REPARSE_TAG_MOUNT_POINT`), whereas `internal/pathsafe`'s existing `canonicalizeReparse` does.
* `docs/compound/concurrency-issues/prefer-atomic-typed-fields-2026-09-04.md`: When changing `transport.Token` to hold a wipeable `[]byte`, `App.Credentials()` must return a deep-cloned `Token` byte slice (`Token.Clone()`) so `zeroTokensLocked()` / `clearLocked()` wiping `a.creds.live` in place under `a.mu` cannot data-race with concurrent callers reading an earlier `a.Credentials()` snapshot under `go test -race`.
* `docs/compound/2026-10-19-hub-test-oracles-and-toolchain-pitfalls.md`: Never use the retired substring `slack` in any Go identifier, and keep all gatecheck `writepath` and `retiredarch` invariants intact.

### Codebase Verification

1. **Deny feedback sanitization and `truncUTF8` (`internal/hub/fold.go`, `internal/hub/permission.go`, `internal/hub/envelope.go`, `internal/agent/copilot/permission.go`, `internal/transport/actions.go`)**:
   * `truncUTF8(s string, n int)` in `internal/hub/fold.go:363` and `truncUTF8Bytes(s string, maxBytes int)` in `internal/transport/actions.go:196` lack an `if n <= 0 { return "" }` guard; a negative `n` causes `s[:n]` to panic.
   * `Hub.resolvePermission` in `internal/hub/permission.go:404` currently calls `dec.Feedback = truncUTF8(dec.Feedback, MaxDecisionFeedbackBytes)` without stripping control, bidi/format (`Cf`), line/paragraph separator (`U+2028`/`U+2029`), newline (`\n`/`\r`/`\t`), or invalid UTF-8 bytes. Likewise, `rejection(feedback string)` in `internal/agent/copilot/permission.go:38` forwards `feedback` raw into `rpc.PermissionDecisionReject`.
2. **Confusable URL defense and emoji/ZWJ summary policy (`internal/agent/copilot/permission.go`, `internal/tui/sanitize.go`)**:
   * `hasURLsLookalikePrefix` in `internal/agent/copilot/permission.go:167` only skips `isInvisibleOrBlankRune(r)` and `utf8.RuneError` while `lead == true`; an interior invisible rune (such as `u\u200brls:` or `u\u3164rls:`) causes `foldURLsLookalikeRune` to return `false` and evade the two-space indent.
   * `golang.org/x/text v0.41.0` is in `go.mod` as an indirect dependency; however, because the target string `"urls:"` consists of only five fixed ASCII characters (`'u'`, `'r'`, `'l'`, `'s'`, `':'`), NFKC compatibility decomposition (fullwidth `U+FF01..U+FF5E`, mathematical alphanumeric alphabets `U+1D400..U+1D7FF` and digits `U+1D7CE..U+1D7FF`, parenthesized/circled/enclosed Latin `U+249C..U+24E9` and `U+1F110..U+1F129`, superscript/subscript/modifier letters, Roman numerals `U+2160..U+217F`, and compatibility colons) plus UTS #39 confusable-skeleton folding can be implemented directly in pure standard-library Go (`unicode`) inside `internal/agent/copilot/permission.go`, preserving Constitution Principle VI (zero new direct module dependencies and zero `go.mod` edits).
   * Both `copilot.isInvisibleOrBlankRune` (`internal/agent/copilot/permission.go:136`) and `tui.isHiddenRune` (`internal/tui/sanitize.go:46`) currently classify `unicode.Variation_Selector`, `unicode.Cf` (including `U+200C` ZWNJ and `U+200D` ZWJ), and `unicode.Co` (Private Use Area) as hidden/invisible.
3. **`cli_path` resolution consolidation and residual hardening (`internal/pathsafe/lexical.go`, `internal/config/validate.go`, `internal/agent/copilot/adapter.go`, `tools/gatecheck/internal/writepath/writepath.go`)**:
   * `internal/pathsafe` already provides `canonicalizeReparse(path string) (string, error)` (`reparse_windows.go` via `GetFinalPathNameByHandleW`, `reparse_other.go` via `filepath.EvalSymlinks`) and `wrapPathError`, and its single `syscall.CreateFile` call is already allowlisted by `tools/gatecheck/internal/writepath/writepath.go`.
   * Neither `os.Stat`, `exec.LookPath`, `filepath.Abs`, nor calling `canonicalizeReparse` inside `internal/pathsafe` triggers `tools/gatecheck/internal/writepath` (which only flags the 76 write selectors and `pathsafe.Root.Resolve`).
   * `internal/config/docs_test.go` (`TestConfigReferenceCoversValidationMessages`) pins single-quoted `'...'` framing in `validateCLIPath` messages (`copilot.cli_path '{path}' does not exist`, etc.). Sanitizing/escaping non-printable, control, newline, invalid-UTF-8, and single-quote runes inside the echoed `{path}` prevents terminal/log injection while keeping every printable path assertion in `docs_test.go` and `copilotcli_test.go` intact.
4. **Process lifecycle, credential hygiene, `localUI` supervision, and `config.Unconsumed` (`internal/app`, `internal/transport/credential.go`, `internal/tui`, `internal/config`, `cmd/intercom`)**:
   * `TUIOptions.RetainPlaintext` in `internal/app/app.go:63` has zero callers outside `package app`; unexporting it to `retainPlaintext` removes the production surface switch while keeping `credentials_test.go` and `e2e_test.go` working in-package.
   * `transport.Token` in `internal/transport/credential.go:50` currently wraps `s string`. Changing it to `b []byte` with `(*Token).Wipe()` (`clear(t.b); t.b = nil`) and `(Token).Clone()` allows `credentialState` in `internal/app/credentials.go` to scrub its in-memory plaintext bytes on `zeroPlaintext()` and `clearCreds()` without data-racing concurrent callers of `App.Credentials()`.

## Options Evaluated

### Axis 1: Summary Sanitization Policy for Emoji Variation Selectors, ZWJ/ZWNJ, and Private-Use Glyphs (`F24EC587`) and Confusable `urls:` Defense (`D3D0E9D5`)

#### Option 1A: Strict Fail-Closed Sanitization Symmetry + NFKC Skeleton Folding with Interior Invisible Tolerance (Recommended)

* **Description**:
  * Retain `unicode.Variation_Selector` (`U+FE00..U+FE0F`, `U+E0100..U+E01EF`), `unicode.Cf` (including `U+200C` ZWNJ and `U+200D` ZWJ), and `unicode.Co` (Private Use Area) in both `copilot.isInvisibleOrBlankRune` and `tui.isHiddenRune`. Lock their agreement with unit tests and ratify the decision in `docs/design-docs/intercom-go-backend-architecture.md` §5.2.
  * In `internal/agent/copilot/permission.go`, upgrade `hasURLsLookalikePrefix` to skip interior invisible/blank/format/replacement/combining runes (`isInvisibleOrBlankRune(r) || r == utf8.RuneError || isCombiningMark(r)`) and fold each visible rune by applying pure-stdlib NFKC compatibility decomposition (`nfkcFoldURLsRune`: fullwidth, mathematical alphanumeric `U+1D400..U+1D7FF`, parenthesized/circled/enclosed, superscript/subscript, Roman numerals, compatibility colons) followed by a UTS #39 confusable-skeleton mapping for `u`, `r`, `l`, `s`, `:`.
* **Pros**:
  * Completely blocks Plane-14 variation-selector ("Glassworm") smuggling, standalone ZWJ/ZWNJ zero-width padding, and PUA blank-glyph spoofing on terminals without Nerd Fonts.
  * Preserves a stateless, single-pass rune predicate across both `internal/agent/copilot` and `internal/tui`.
  * `norm.NFKC` + confusable-skeleton folding + interior-invisible skipping defeats mathematical alphanumeric (`𝐔𝐫𝐥𝐬:`), fullwidth, compatibility, and zero-width-interleaved (`u\u200brls:`) spoofing of the column-0 `urls:` line.
* **Cons**:
  * Multi-codepoint ZWJ emoji sequences and Nerd Font PUA icons in permission summaries are rendered with `<U+XXXX>` / `U+FFFD` markers rather than decorative glyphs.
* **Effort estimate**: Low.
* **Fit**: High — security-first posture appropriate for an operator permission approval surface.

#### Option 1B: Context-Sensitive Grapheme Parser Allowing Paired `U+FE0F`/`U+200D` Emoji Sequences

* **Description**: Parse grapheme clusters in `sanitizeBound` and `tui.Sanitize` so `U+FE0E`/`U+FE0F` and `U+200D` are preserved when attached to an emoji base character, while still replacing standalone selectors, `U+E0100..U+E01EF`, and `Co`.
* **Pros**: Preserves decorative multi-codepoint emoji ligatures in permission summaries.
* **Cons**: Introduces stateful grapheme-cluster parsing across two packages (`copilot` and `tui`), risks terminal width miscalculation across different terminal emulators (`conhost` vs `xterm`), and adds complexity for zero security or operational benefit on a permission modal.
* **Effort estimate**: Medium.
* **Fit**: Low — violates simplicity for decorative rendering on a security gate.

### Axis 2: `cli_path` Shared Resolution Helpers and Residual Hardening (`2C45EFAE`, `2F1B6D4B`)

#### Option 2A: Shared `pathsafe.ResolveCLIPath` with Reparse Canonicalization, Batch-Script Rejection, Unix Execute-Bit Check, and Control-Safe Path Quoting (Recommended)

* **Description**:
  * Add `pathsafe.ResolveCLIPath(p string) (string, CLIPathErrorKind, error)` and `pathsafe.SanitizeCLIPathForMessage(p string) string` in `internal/pathsafe`.
  * For both `CLIPathAbs` and `CLIPathBare` (after `exec.LookPath` + `filepath.Abs`), reject paths with an NTFS Alternate Data Stream or extra colon segment (`hasADSOrExtraColon`), verify `os.Stat` + `IsRegular()`, canonicalize symlinks/junctions/mapped drives via `canonicalizeReparse(abs)`, verify the canonical target classifies as `CLIPathAbs` (rejecting symlinks/junctions/mapped drives that resolve to a UNC share `\\server\share\...` or device path) and has no ADS suffix, re-stat the canonical target for `IsRegular()`, reject `.bat` and `.cmd` extensions (case-insensitive) on both the input and canonical path, and on Unix (`runtime.GOOS != "windows"`) require at least one execute bit (`cinfo.Mode().Perm()&0o111 != 0`).
  * Both `internal/config.validateCLIPath` and `internal/agent/copilot.resolveCLIPath` delegate to `pathsafe.ResolveCLIPath`, keeping only their caller-specific policy (non-fatal `Report.Warnings` advisory on `CLIPathErrLookPath` in `config` vs. fatal wrapped `%w` error in `copilot`, plus package-specific error prefixes).
  * `config.validateCLIPath` formats echoed paths through `pathsafe.SanitizeCLIPathForMessage(cliPath)` inside `'...'` so printable paths preserve the exact `'...'` substrings tested by `internal/config/docs_test.go`, while control, newline, bidi/format, invalid UTF-8, and `'` characters are escaped.
  * Document the verified empty-`cli_path` SDK behavior (delegating to `copilot-sdk` with `Adapter.Start` enforcing `GetStatus` + `assertRuntime` version/protocol floor) in `docs/design-docs/intercom-go-backend-architecture.md` §6.2 and `docs/config-reference.md`.
* **Pros**:
  * Eliminates the duplicated resolution logic between `internal/config` and `internal/agent/copilot`.
  * Closes every residual in `2F1B6D4B` (symlink/junction/mapped-drive to UNC, ADS suffix, Windows `.bat`/`.cmd` via `PATHEXT`, Unix execute bit, log/error control-rune injection, and empty-`cli_path` documentation) without tripping `tools/gatecheck/internal/writepath` or `retiredarch`.
* **Cons**:
  * Requires updating `TestValidateRule5RejectsDirectoryCLIPathAndAcceptsRegularFile` in `internal/config/copilotcli_test.go` to compare against the canonicalized path (because `t.TempDir()` on Windows can contain an 8.3 short name expanded by `canonicalizeReparse`).
* **Effort estimate**: Medium.
* **Fit**: High.

#### Option 2B: Lexical-Only Consolidation Without `canonicalizeReparse`

* **Description**: Share `exec.LookPath` + `filepath.Abs` + `os.Stat` in `internal/pathsafe` without calling `canonicalizeReparse` or checking `.bat`/`.cmd` and Unix execute bits.
* **Pros**: Smaller diff in `internal/pathsafe`.
* **Cons**: Leaves symlinks/junctions pointing to UNC shares, Windows `.bat`/`.cmd` shims on `PATH`, and non-executable Unix files unaddressed, failing to close `2F1B6D4B`.
* **Effort estimate**: Low.
* **Fit**: Low.

### Axis 3: Process Lifecycle, Credential Hygiene, `localUI` Supervision, and `config.Unconsumed` (`24CE7E1D`, `5E4834CF`, `AE317BD3`, `4B7308F2`, `7BFA4182`, `26E15024`, `864FDF7A`, `D953F24D`)

#### Option 3A: Full In-Memory `[]byte` Token Scrubbing, `credentialState` Struct, Self-Supervised `localUI`, and Schema-Adjacent `config.Unconsumed` with One-Time TUI Notice (Recommended)

* **Description**:
  * Change `transport.Token` to hold `b []byte` with `(*Token).Wipe()` (`clear(t.b); t.b = nil`) and `(Token).Clone() Token`, and unexport `TUIOptions.retainPlaintext` (`AE317BD3`).
  * Group credential fields on `App` into `credentialState` (`26E15024`), extract `zeroTokensLocked()` and `clearLocked()` (`4B7308F2`), make `mint()` early-out on `a.creds.cleared || a.shuttingDown()` under `a.mu` (`4B7308F2`), and call `a.clearCreds()` immediately after `<-a.shut.req` in `App.supervise()` (`24CE7E1D`).
  * Update `App.credentialsMsg` to include metadata-only live credentials (`Secret == ""`, `Fingerprint != ""`) when a partial `Reissue` fails after `zeroPlaintext`, and render empty-secret entries in `tui.Model.credentialRows` as `(already shown)` (`5E4834CF`).
  * Reorganize `internal/app`: move `watchUI`, `pumpRejects`, `pumpRejectsTo`, `defaultRejectNoticeInterval`, and `onReject` into `internal/app/ui.go`, create `watchDone` and start the watcher inside `startUI` (`7BFA4182`, `26E15024`), drop redundant `quit`/`kill` nil fallbacks in `(*localUI).stop()`, move `teardown` and `defaultTeardownPermissions` into `internal/app/shutdown.go` beside `reap`, derive `maxApproval` from `config.MaxApprovalTimeoutSeconds`, add `(*App).RejectedUpgrades() uint64`, and test the `supervise` -> `teardown` fallback when `hub.NewShutdown` fails (`D953F24D`).
  * Move `unconsumedKeys` to `config.Unconsumed(cfg *Config) []string` and `config.UnconsumedKeyNames() []string` in `internal/config`, compare `cfg.Database.Path != def.Database.Path`, add a reflect-based test verifying every `Config` leaf/map key is either consumed or listed in `UnconsumedKeyNames()`, derive `cmd/intercom/runbook_test.go`'s unconsumed-key check from `config.UnconsumedKeyNames()`, and pass `UnconsumedKeys: keys` into `tui.Options` so dismissing the initial credential panel surfaces a one-time `"unconsumed config keys: ..."` notice on the TUI status line (`864FDF7A`, `D953F24D`).
* **Pros**:
  * Scrubs credential token bytes in place via `clear(t.b)` while keeping `App.Credentials()` 100% race-free via `Token.Clone()`.
  * Eliminates the four-way duplication of the unconsumed config key list and enforces schema completeness via reflection.
  * Makes `localUI` lifecycle self-contained so `startUI` owns `watchDone` and `pumpRejects` creation.
* **Cons**:
  * Touches multiple packages (`internal/transport`, `internal/config`, `internal/tui`, `internal/app`, `cmd/intercom`, `docs/`), requiring strict dependency ordering across tasks.
* **Effort estimate**: Medium.
* **Fit**: High.

#### Option 3B: Log-Only Unconsumed Keys Warning and String-Backed `transport.Token`

* **Description**: Keep `transport.Token` as an immutable `string` (documenting the limitation) and document why the unconsumed-keys warning is log-only instead of surfacing a one-time TUI notice.
* **Pros**: Slightly fewer lines changed in `internal/transport` and `internal/tui`.
* **Cons**: Leaves credential plaintext in un-wipeable Go heap strings and leaves interactive TUI operators blind to unconsumed config keys when `stderr` is not redirected.
* **Effort estimate**: Low.
* **Fit**: Medium — misses concrete security and observability improvements requested in `AE317BD3` and `D953F24D`.

## Trade-off Comparison

| Criterion | Chosen Direction (Options 1A + 2A + 3A) | Alternative (Options 1B + 2B + 3B) |
|---|---|---|
| Security & Anti-Spoofing | High — NFKC + confusable skeleton + interior invisible tolerance; fail-closed emoji/ZWJ/PUA markers; canonicalized `cli_path`; wipeable `[]byte` tokens | Low-Medium — grapheme complexity in TUI; uncanonicalized `cli_path` symlinks; immutable `string` tokens |
| Maintainability | High — single `pathsafe.ResolveCLIPath`, single `config.Unconsumed` with reflect lock, `credentialState` + `zeroTokensLocked`, self-supervised `localUI` | Low — retains partial duplication or weaker invariants |
| Gate & Test Compatibility | High — zero new `pathsafe.Root.Resolve` or write primitives; preserves `docs_test.go` and `runbook_test.go` contracts | Medium |
| Effort & Blast Radius | 2 covering features (`085-F`, `086-F`), 14 atomic tasks (`<= 2h` each, `085.001-T`..`085.005-T` and `086.001-T`..`086.009-T`), strict width isolation | Fewer tasks but incomplete closure of stash items |

## Decision

We adopt **Option 1A**, **Option 2A**, and **Option 3A**, decomposed into two covering features and two sequential shipments:

1. **Feature `085-F` / Shipment `078-S`** (*Permission Feedback Sanitization, Confusable URL Defense, Emoji/ZWJ Summary Policy, and Copilot `cli_path` Resolution Hardening*):
   * **Unit 1.1 (`085.001-T`, `code`)**: Add `if n <= 0 { return "" }` to `truncUTF8` (`internal/hub/fold.go`), and sanitize operator deny feedback via `sanitizeFeedback` in `internal/hub/permission.go` (`resolvePermission`) so control, newline (`\n`, `\r`, `\t`), line/paragraph separator (`U+2028`, `U+2029`), format/bidi (`Cf`), visually blank filler, and invalid UTF-8 runes are replaced before bounding to `MaxDecisionFeedbackBytes` (`2EE9F3ED`).
   * **Unit 1.2 (`085.002-T`, `code`)**: In `internal/agent/copilot/permission.go` and `internal/tui/sanitize_test.go`, sanitize deny feedback in `rejection(feedback)` as defence in depth (`2EE9F3ED`), upgrade `hasURLsLookalikePrefix` / `foldURLsLookalikeRune` to use pure-stdlib NFKC compatibility decomposition (`nfkcFoldURLsRune`) plus UTS #39 confusable-skeleton folding while skipping interior invisible/blank/format/replacement/combining runes (`D3D0E9D5`), and lock the fail-closed policy keeping `Variation_Selector` (`U+FE0E`/`U+FE0F`/`U+E0100..U+E01EF`), `Cf` (`ZWJ`/`ZWNJ`), and `Co` (Private Use Area) replaced with visible markers / `U+FFFD` across both `copilot` and `tui` (`F24EC587`).
   * **Unit 1.3 (`085.003-T`, `code`)**: Extract shared `pathsafe.ResolveCLIPath`, `pathsafe.CLIPathErrorKind`, and `pathsafe.SanitizeCLIPathForMessage` in `internal/pathsafe/lexical.go`, adding symlink/junction/mapped-drive canonicalization via `canonicalizeReparse` (rejecting UNC/device targets), NTFS ADS / extra-colon rejection, `.bat`/`.cmd` batch-script rejection, and Unix execute-bit validation (`2C45EFAE`, `2F1B6D4B`).
   * **Unit 1.4 (`085.004-T`, `code`)**: Wire `pathsafe.ResolveCLIPath` and `pathsafe.SanitizeCLIPathForMessage` into `internal/config/validate.go` (`validateCLIPath`) and `internal/agent/copilot/adapter.go` (`resolveCLIPath`), keeping caller-specific warning/error policies (`2C45EFAE`, `2F1B6D4B`).
   * **Unit 1.5 (`085.005-T`, `docs`)**: Update `docs/design-docs/intercom-go-backend-architecture.md` (§5.2, §6.1, §6.2, §8) and `docs/config-reference.md` for deny feedback sanitization, NFKC `urls:` confusable folding, fail-closed emoji/ZWJ/PUA summary sanitization policy, and `pathsafe.ResolveCLIPath` canonicalization/hardening + empty-`cli_path` SDK behavior (`2EE9F3ED`, `D3D0E9D5`, `F24EC587`, `2C45EFAE`, `2F1B6D4B`).

2. **Feature `086-F` / Shipment `079-S`** (*Process Lifecycle, Credential Hygiene, `localUI` Supervision, and `config.Unconsumed` Architecture*, blocked by `078-S`):
   * **Unit 2.1 (`086.001-T`, `code`)**: Convert `transport.Token` in `internal/transport/credential.go` from `string` to a wipeable `[]byte` with `(*Token).Wipe()` and `(Token).Clone()` (`AE317BD3`).
   * **Unit 2.2 (`086.002-T`, `code`)**: In `internal/tui/credentials.go` and `internal/tui/model.go`, render metadata-only `Credential` entries (`Secret == ""`) as `(already shown)` in the credential panel (`5E4834CF`) and add `Options.UnconsumedKeys` to surface a one-time `"unconsumed config keys: ..."` status-line notice when the initial credential panel is dismissed (`D953F24D`).
   * **Unit 2.3 (`086.003-T`, `code`)**: Add `config.Unconsumed(cfg *Config) []string` and `config.UnconsumedKeyNames() []string` in `internal/config/unconsumed.go` (comparing `cfg.Database.Path != def.Database.Path`) with a reflect-based completeness test in `internal/config/unconsumed_test.go`, and derive `cmd/intercom/runbook_test.go`'s unconsumed-key list from `config.UnconsumedKeyNames()` (`864FDF7A`).
   * **Unit 2.4 (`086.004-T`, `code`)**: In `internal/app/credentials.go`, `internal/app/e2e_test.go`, and `internal/app/review1_test.go`, extract `zeroTokensLocked()` / `clearLocked()`, make `mint()` early-out on `credsCleared` under `a.mu`, emit metadata-only entries in `credentialsMsg` when `Secret == ""` and `Fingerprint != ""`, clone `Token` in `Credentials()`, and add `setAfterMint` / `tuiOptionsWithPlaintext` helpers (`4B7308F2`, `5E4834CF`, `AE317BD3`).
   * **Unit 2.5 (`086.005-T`, `code`)**: In `internal/app/app.go`, `internal/app/assemble.go`, and `internal/app/credentials_test.go`, unexport `TUIOptions.RetainPlaintext` to `retainPlaintext` and lock it with a reflection test (`AE317BD3`).
   * **Unit 2.6 (`086.006-T`, `code`)**: In `internal/app/assemble.go`, `internal/app/credentials.go`, and `internal/app/reap_test.go`, group credential fields on `App` into `credentialState` (`creds credentialState`) mirroring `shutdownState` (`26E15024`).
   * **Unit 2.7 (`086.007-T`, `code`)**: In `internal/app/assemble.go`, `internal/app/shutdown.go`, and `internal/app/shutdown_test.go`, move `teardown` and `defaultTeardownPermissions` beside `reap` in `shutdown.go`, derive `maxApproval` from `config.MaxApprovalTimeoutSeconds` and delegate `unconsumedKeys` to `config.Unconsumed`, call `a.clearCreds()` immediately after `<-a.shut.req` in `supervise()`, and test both immediate credential clearing on shutdown start and the `supervise` -> `teardown` fallback when `hub.NewShutdown` fails (`24CE7E1D`, `26E15024`, `864FDF7A`, `D953F24D`).
   * **Unit 2.8 (`086.008-T`, `code`)**: In `internal/app/assemble.go`, `internal/app/ui.go`, and `internal/app/tui_test.go`, move `watchUI`, `pumpRejects`, `pumpRejectsTo`, `defaultRejectNoticeInterval`, and `onReject` into `ui.go`, create `watchDone` and start the watcher inside `startUI` (`startWatchedUI`), drop redundant `quit`/`kill` nil defaults in `(*localUI).stop()`, wire `UnconsumedKeys` into `startUI`, and expose `(*App).RejectedUpgrades() uint64` (`7BFA4182`, `26E15024`, `D953F24D`).
   * **Unit 2.9 (`086.009-T`, `docs`)**: Update `docs/design-docs/intercom-go-backend-architecture.md` (§3.1 failed-start `teardown` as a reduced shutdown sequence) and `docs/runbooks/intercom-operator-runbook.md` (§4 metadata-only partial re-issue display and §7 one-time TUI unconsumed-keys notice) (`5E4834CF`, `864FDF7A`, `D953F24D`).

## Rejected Alternatives

* **Allowing contextual emoji variation selectors / ZWJ in permission summaries (Option 1B)**: Rejected because decorative emoji ligatures add no value to a security approval modal, while stateful grapheme parsing risks terminal width divergence and Plane-14 variation-selector smuggling.
* **Calling `pathsafe.Root.Resolve` for `cli_path` (under Option 2)**: Rejected for two independent reasons: (1) `copilot.cli_path` is an operator-configured system binary outside the workspace root (whereas `Root.Resolve` enforces containment inside a workspace root), and (2) `reviewedResolveCallers` in `tools/gatecheck/internal/writepath/writepath.go` pins `pathsafe.Root.Resolve` to `internal/app/workspace.go` (`calls: 1`). Calling `canonicalizeReparse` directly inside `internal/pathsafe` canonicalizes symlinks and Windows junctions without invoking `Root.Resolve`.
* **Keeping `transport.Token` as a `string` (Option 3B)**: Rejected because converting `Token` to a `[]byte` with `Wipe()` and `Clone()` achieves true in-place zeroing of credential bytes in process memory without introducing data races.

## Unresolved Questions

None. All 13 stash entries have concrete, verified design resolutions backed by codebase inspection.

## Risks and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| `go test -race` data race when `zeroTokensLocked()` wipes `Token` `[]byte` while a caller reads an earlier `App.Credentials()`, `Reissue()`, or `tuiReissue()` slice | Medium | High | `cloneIssuedSlice` (using `(Token).Clone()`) allocates an independent `[]byte` copy for `a.creds.live`, `App.Credentials()`, and `Reissue()` return values, while `mintAll()` and `tuiReissue()` deterministically wipe their temporary slices (`wipeIssuedSlice`) and `zeroTokensLocked()` wipes `a.creds.live` in place under `a.mu`. |
| Windows `t.TempDir()` 8.3 short-name expansion by `canonicalizeReparse` breaking `TestValidateRule5RejectsDirectoryCLIPathAndAcceptsRegularFile` | High | Medium | Compare `cfg.Copilot.CLIPath` against the canonicalized path (or create the test binary inside the already-canonicalized `cfg.DefaultWorkspaceRoot`). |
| `TestConfigReferenceCoversValidationMessages` in `internal/config/docs_test.go` failing if `validateCLIPath` switches from `'...'` to `"..."` | High | Medium | Keep `'...'` framing in `validateCLIPath` and sanitize the echoed path via `pathsafe.SanitizeCLIPathForMessage`, preserving 100% compatibility with `docs_test.go` while preventing control/quote injection. |
| Width-isolation violation if code and `docs/**` changes are mixed in a single task | Low | High | Isolate all `docs/**` updates into dedicated docs tasks (`085.005-T` and `086.009-T`). |

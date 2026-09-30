# Memory checkpoint: 034-S gate-engine Go migration M1 (session 2)

- Timestamp: 2026-09-30T02:02:03Z
- Shipment: 034-S (feature 044-F), dark-mode run, merge NOT pre-authorized
- Branch: `feat/034-s-gate-engine-go-migration-m1-gatecheck-scaffold-pysem-masker-write-path`

## Completed since last checkpoint

- 044.006-T (pysem case mapping: IsUpper/IsLower/Lower + Final_Sigma) — commit `cfe0ec9`.
  Required adding `case_table.go` (otherLowerTable/otherUpperTable, Other_Lowercase /
  Other_Uppercase derived-property supplements) after discovering `unicode.IsUpper`/
  `IsLower` are Lu/Ll-category-only and under-match CPython's islower()/isupper()
  (caught by golden case U+2071 SUPERSCRIPT LATIN SMALL LETTER I). All 14 Lower goldens
  passed on first run; IsUpper/IsLower 100% after the table fix.
- 044.007-T (pysem string helpers: Strip/SplitLines/Repr/WordBoundary) — commit `2d9562f`.
  100% golden agreement on first run for all four (strip 8/8, split_lines 10/10, repr
  20/20, word_boundary 12/12). `PrecededByWordOrDot`/`FollowedByWord` deliberately
  excluded from this commit — plan's M1-T6a/M1-T6b split assigns them to io.go.
- 044.008-T (pysem I/O and lookaround: ReadText/GitText/PrecededByWordOrDot/FollowedByWord)
  — commit `6e130f7`. 100% golden agreement (5/5 read_text via both ReadText and GitText,
  12/12 preceded, 7/7 followed) on first run. `internal/pysem` package is now COMPLETE
  for M1 — all 14 exported helpers + ErrInvalidUTF8 sentinel ported and golden-verified.

## Backlog state (ground truth via backlogit)

- Shipment 034-S: `active`
- Feature 044-F: `active`
- Tasks 044.001-T through 044.008-T: `done`, archived, commits tracked
- Tasks 044.009-T through 044.012-T: `active` (backlogit sibling-flip anomaly, non-blocking;
  continuing single-threaded dependency-ordered execution per plan §8)

## Git state

- Commits this session so far (in order): `65655f7`, `86922b6`, `3b13012`, `2e079c5`,
  `637f924`, `cfe0ec9`, `2d9562f`, `6e130f7`
- Working tree clean at this checkpoint (all pysem work committed)
- go.mod/go.sum unchanged (stdlib-only imports throughout, confirmed again)

## Next steps

1. 044.009-T: gomask rune-level masker port (`internal/gomask/gomask.go`,
   `gomask_test.go`) — read `scripts/lib/gomask.py` and the committed
   `masker_golden.json` (42 entries, already captured in commit `83d99c9`/`3b13012`).
   Remember D-4: pin `multiline-tag-shaped-raw-string` AS-IS (stash C312BD4C is OUT of
   scope — do not "fix" that behavior).
2. 044.010-T: write-path engine port (`internal/writepath/writepath.go`,
   `writepath_test.go`) — uses `writepath_golden.json` (also already captured).
3. 044.011-T: shared bash runner (`scripts/lib/gatecheck-run.sh` +
   `tools/gatecheck/runner_test.go`) — remember to invoke Git-Bash explicitly
   (`C:\Program Files\Git\bin\bash.exe`), exclude WindowsApps from PATH.
4. 044.012-T: switch write-path wrapper to Go, capture parity evidence.
5. After all 12 tasks: full quality gates, local review gate (P-021 C1 classification,
   defer-capture for out-of-scope findings — stash entries 8E9F8E55/B72E9715/8E18CCF5/
   56B16321 and C312BD4C are explicitly out of scope per operator instructions), PR
   creation, CI/copilot-review handling, P-014 §1.9 readiness gate.
6. STOP at merge readiness — do NOT merge (dark mode, merge_approval_pre_authorized:
   false). Report to operator per the required return-to-Orchestrator structure.

## Environment quirks (carried forward, still relevant)

1. Bare `bash` on PATH → WSL2 (no Go toolchain); use `C:\Program Files\Git\bin\bash.exe`
   explicitly for 044.011-T/044.012-T script testing.
2. Git-Bash's `python3` may resolve to a broken Windows Store stub; exclude
   `WindowsApps` from PATH when invoking bash for anything needing Python.
3. Windows-native `python.exe` applies `\n`→`\r\n` universal-newline translation to
   subprocess stdout/stderr; normalize captured streams for goldens.
4. Python console output crashes on non-ASCII with default `cp1252` codec; set
   `$env:PYTHONIOENCODING = "utf-8"` before invoking Python for any non-ASCII output.
5. gofmt gives false positives on this Windows CRLF checkout; always verify against
   LF-normalized copies (established pattern: copy to `$env:TEMP\gofmt-check`, replace
   `` `r`n `` with `` `n ``, run `gofmt -l` there).
6. Never `git add -A` (a `.backlogit/.locks/itemlog/.identity-version` file is not
   covered by gitignore); always stage explicitly by path.
7. backlogit commit tracking: `backlogit update <id> --commit <sha>` (no separate
   "commit" subcommand).

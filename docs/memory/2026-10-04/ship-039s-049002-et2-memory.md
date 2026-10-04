---
title: "039-S E-T2 implementation checkpoint"
date: 2026-10-04
agent: ship
feature: 049-F
shipment: 039-S
task: 049.002-T
mode: dark (P-017)
status: verified-pending-commit
---

## Outcome

The E-T2 AST scanner implementation is complete and locally verified. The
current branch remains `feat/migrate-write-path-scanner-to-go-ast-post-m4-unit-e`
at `f5a9b8e` before the task commit. Shipment 039-S and task 049.002-T are
active; 049.001-T is done, with its queue-to-archive move awaiting commit.

## Implementation and acceptance evidence

* `scanSource` parses unmasked Go source with `go/parser` and fails closed on
  parse errors, including partial ASTs. Selector positions map through the
  owning `token.File`, source-byte-to-rune offsets, and masked-byte offsets.
  `PositionFor(pos, false)` keeps `//line` directives from changing physical
  finding lines.
* `scanFile` reads through `pysem.ReadText` and routes the unmasked result to
  `scanSource`. The canonical whole-file `gomask.MaskGoNonCode` behavior,
  finding text/order, raw-string struct-tag rule, and interim
  `syscall.CreateFile` text predicate remain intact. `gomask` is unchanged.
* The authorized oracle adaptation changes only the production boundary and
  moves code-position form-feed, vertical-tab, and U+2028 inputs to
  fail-closed parser assertions. The frozen legacy implementation and original
  20-selector list remain unchanged.
* R4-3's single `writepath_extent_test.go` `"x.go"` to `"y.go"` change is
  already in the committed harness scaffold `f5a9b8e`; the one test and both
  expected findings are present there and were not edited again. The E-T2
  boundary is five files when that harness change is counted.
* R4-5 is implemented with `tf.Offset(selector.End())`, then source byte
  offset to rune index to masked byte offset; invalid mappings fail closed.
* R6-2's migrated failure prefixes name `scanSource`. The retiredarch pin
  anchors `scanSource` masking its source parameter and feeding the result to
  `pysem.SplitLines`; its pin and mutation tests have no positive `scanText`
  dependency. The canonical-mask and single-definition checks remain.
* AC-E2.5 was verified against the baseline executable built from committed
  `f5a9b8e`. For repo scan, `--self-test`, and `--self-test-integrity`, stdout
  and stderr were compared byte-for-byte and exit codes matched. Captured
  before/after streams are in the session files directory as
  `039-s-repo-*`, `039-s-selftest-*`, and `039-s-integrity-*`.

## Verification

* TDD red phase was confirmed before implementation: the task-specific harness
  failures contained the expected `not implemented: 049.002-T` marker.
* Green harness: `WRITEPATH_HARNESS_TASK=049.002-T go test ./tools/gatecheck/internal/writepath -run '^TestHarness_049002_' -count=1` passed.
* `gofmt -l .` returned no files.
* `go vet ./...` passed.
* `go test ./...` passed.
* `go build ./...` passed.
* `golangci-lint run ./tools/gatecheck/internal/writepath ./tools/gatecheck/internal/retiredarch` reported 0 issues.
* `git diff --check` passed.

## Files and next step

The implementation changes are in `writepath.go`, `writepath_test.go`,
`writepath_oracle_test.go`, and `retiredarch/writepath_mask_test.go`; the
committed harness supplies the fifth R4-3 file. The pending backlog move is
`.backlogit/queue/049.001-T.md` to `.backlogit/archive/049.001-T.md`.

Next, commit the E-T2 implementation with the 049.001-T archive move, mark
049.002-T done and associate the commit through backlogit, then continue with
049.003-T. Do not rerun `pre_claim` for the already-active shipment. Engram
remains degraded and must not be retried; no PR, merge, or post-merge closure
has occurred.

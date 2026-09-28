---
title: "Behaviour-preservation evidence: 029-S / 032-F — gate engine extraction (plan unit 2)"
description: "Pre/post capture-set evidence for shipment 029-S task 032.007-T (AC-2.2, AC-2.3, AC-2.4)"
status: "complete"
tags:
  - "evidence"
  - "029-S"
  - "032-F"
  - "032.007-T"
date: 2026-09-28
shipment: 029-S
feature: 032-F
task: 032.007-T
bound_snapshot: 14d44e3c
pre_extraction_base: beeb84f
post_extraction_head: 3042ae5
---

# Behaviour-preservation evidence: 029-S / 032-F — gate engine extraction

## Scope

Evidence task 032.007-T (plan U2-T6), no production change. The capture set is the four NAMED runs:
`check-retired-architecture.sh` bare repo scan, `--self-test`, `--self-test-integrity`; and
`check-write-path-precondition.sh --self-test`. Every run was executed under WSL bash
(Python 3.14.4) from the repository root with `PYTHONDONTWRITEBYTECODE=1`; stdout and stderr are
merged in the verbatim blocks below. CI re-runs the same scripts on ubuntu-latest (Python 3.12).

## Sides captured

| Side | Scripts | Tree scanned |
|---|---|---|
| Bound snapshot | `14d44e3c` blobs of both scripts | pre-extraction base `beeb84f` |
| Pre-extraction | `beeb84f` scripts (heredoc engines) | `beeb84f` |
| Post-extraction | `3042ae5` wrapper + `scripts/lib/` modules | `3042ae5` |

Binding note: between `14d44e3c` and `beeb84f` the retired-architecture script, `cmd/**`,
`internal/**`, `config.toml.example` and every `retired*` / `writepath` fixture are unchanged; the only
deltas are the additive `--self-test-integrity` mode in the write-path script (030.001-T) and new
`scripts/testdata/mergestrategy/` fixtures that no capture-set run reads. The snapshot scripts were run
directly (extracted with `git show 14d44e3c:<path>`) rather than in a second worktree (P-016).

## Result

| Run | Snapshot exit | Pre exit | Post exit | Snapshot = pre (bytes) | Verdicts pre = post | Output pre = post |
|---|---|---|---|---|---|---|
| `scripts/check-retired-architecture.sh` | 0 | 0 | 0 | yes | yes | yes |
| `scripts/check-retired-architecture.sh --self-test` | 0 | 0 | 0 | yes | yes | one line (see below) |
| `scripts/check-retired-architecture.sh --self-test-integrity` | 0 | 0 | 0 | yes | yes | one line (see below) |
| `scripts/check-write-path-precondition.sh --self-test` | 0 | 0 | 0 | yes | yes | yes |

* **AC-2.2** — all four named runs pass post-extraction (exit 0).
* **AC-2.3** — retired-architecture verdicts are identical across all three modes: same exit codes, and the
  same ordered list of `PASS`/`FAIL <assertion>` verdict labels (text before the first `:`). The single textual
  difference is the success message of `selection pathspec pin (AC-6/AG-1)`, which names the mechanism and
  changed with 032.004-T's re-derivation (disk-read region extraction -> `inspect.getsource()`); the verdict
  label and outcome are unchanged:

```text
< PASS selection pathspec pin (AC-6/AG-1): select_repo_paths pathspec and should_scan_repo_path prefix set pinned via disk-read, region-anchored extraction
> PASS selection pathspec pin (AC-6/AG-1): select_repo_paths pathspec and should_scan_repo_path prefix set pinned via inspect.getsource() source-text extraction
```

* **AC-2.4** — write-path observed verdict delta set is EMPTY, exactly matching the EMPTY expected set
  enumerated by 032.001-T (class D-1 only, 0 instances). `--self-test` output is byte-identical, including the
  tracked-tree repo scan it runs, so no tracked `cmd/**` or `internal/**` verdict moved (AC-2.2 precedence).
  Supporting measurement (ad hoc, 032.001-T): across all 92 tracked `.go` files the two maskers differ only on 5
  files containing struct-tag raw strings (one inside the write-path selection, `internal/config/config.go`) and
  produce zero selector-finding differences; the canonical masker in `scripts/lib/gomask.py` is byte-identical to
  the former retired-architecture heredoc masker on all 92 files.

## Verbatim output — Pre-extraction (beeb84f; output byte-identical to the bound-snapshot run)

### `scripts/check-retired-architecture.sh` — exit 0

```text
::notice::retired-arch gate mode=repo
```

### `scripts/check-retired-architecture.sh --self-test` — exit 0

```text
::notice::retired-arch gate mode=self-test
PASS suite toml non-empty: suite discovered 14 fixture(s)
PASS suite toml has accept and reject coverage: suite has at least one 'accept' and one 'reject' fixture
PASS retired-accept-composed-key-corpus.toml [tomllib]: clean as expected
PASS retired-accept-composed-key-corpus.toml [fallback]: clean as expected
PASS retired-architecture-fixture.toml [tomllib]: rejected as expected
PASS retired-architecture-fixture.toml [fallback]: rejected as expected
PASS retired-malformed-unparseable.toml [tomllib]: rejected as expected
PASS retired-malformed-unparseable.toml [fallback]: rejected as expected
PASS retired-negative-inside-multiline-literal.toml [tomllib]: clean as expected
PASS retired-negative-inside-multiline-literal.toml [fallback]: clean as expected
PASS retired-negative-nested-inert-delimiters.toml [tomllib]: clean as expected
PASS retired-negative-nested-inert-delimiters.toml [fallback]: clean as expected
PASS retired-positive-after-multiline-close.toml [tomllib]: rejected as expected
PASS retired-positive-after-multiline-close.toml [fallback]: rejected as expected
PASS retired-positive-escaped-delimiter-close.toml [tomllib]: rejected as expected
PASS retired-positive-escaped-delimiter-close.toml [fallback]: rejected as expected
PASS retired-reject-concatenated-key.toml [tomllib]: rejected as expected
PASS retired-reject-concatenated-key.toml [fallback]: rejected as expected
PASS retired-reject-dotted-key.toml [tomllib]: rejected as expected
PASS retired-reject-dotted-key.toml [fallback]: rejected as expected
PASS retired-reject-fused-plural-key.toml [tomllib]: rejected as expected
PASS retired-reject-fused-plural-key.toml [fallback]: rejected as expected
PASS retired-reject-hyphenated-key.toml [tomllib]: rejected as expected
PASS retired-reject-hyphenated-key.toml [fallback]: rejected as expected
PASS retired-reject-multiline-open-key.toml [tomllib]: rejected as expected
PASS retired-reject-multiline-open-key.toml [fallback]: rejected as expected
PASS retired-reject-quoted-dotted-key.toml [tomllib]: rejected as expected
PASS retired-reject-quoted-dotted-key.toml [fallback]: rejected as expected
PASS retired-reject-table-plus-key.toml [tomllib]: rejected as expected
PASS retired-reject-table-plus-key.toml [fallback]: rejected as expected
PASS suite go non-empty: suite discovered 19 fixture(s)
PASS suite go has accept and reject coverage: suite has at least one 'accept' and one 'reject' fixture
PASS retired-accept-block-comment.go [go]: clean as expected
PASS retired-accept-comment-only.go [go]: clean as expected
PASS retired-accept-interpreted-string-escaped-quote.go [go]: clean as expected
PASS retired-accept-raw-string-multiline.go [go]: clean as expected
PASS retired-accept-struct-tag-benign.go [go]: clean as expected
PASS retired-accept-struct-tag-non-tag-raw-string.go [go]: clean as expected
PASS retired-accept-token-corpus.go [go]: clean as expected
PASS retired-reject-after-rune-literal.go [go]: rejected as expected
PASS retired-reject-channel-ids.go [go]: rejected as expected
PASS retired-reject-host-clis.go [go]: rejected as expected
PASS retired-reject-ipc-names.go [go]: rejected as expected
PASS retired-reject-slack-team-id.go [go]: rejected as expected
PASS retired-reject-socket-modes.go [go]: rejected as expected
PASS retired-reject-socketmode-lower.go [go]: rejected as expected
PASS retired-reject-socketmode-mixed.go [go]: rejected as expected
PASS retired-reject-socketmode-plural-fused.go [go]: rejected as expected
PASS retired-reject-struct-tag.go [go]: rejected as expected
PASS retired-reject-team-ids-cache.go [go]: rejected as expected
PASS retired-reject-team-ids.go [go]: rejected as expected
PASS suite go-differential non-empty: suite discovered 1 fixture(s)
PASS suite go-differential reject-only exemption: suite is a named reject-only exemption and all 1 fixture(s) are 'reject'
PASS retired-differential-comment-only.go [go-unmasked]: rejected as expected
PASS selection internal non-empty: independently-derived expected internal/** set is non-empty (15 paths)
PASS selection structural inclusion: selected every tracked internal non-test, non-testdata Go file (15 paths)
PASS selection internal exclusions: excluded internal test files and internal testdata paths
PASS selection self-scan guard: did not select scripts/ paths and predicate rejects scripts/testdata/retiredgo/x.go
PASS selection non-empty: selected 21 tracked repo paths
PASS dispatch config.toml.example: engine_for_path routes config.toml.example to the TOML engine
PASS selection includes config.toml.example: real selection set includes config.toml.example end-to-end
PASS selection dispatch coverage: every selected repo path resolves to a known scan engine
PASS selection cmd/ coverage (AG-5/D4): cmd/ deliberately includes *_test.go and testdata/ paths (3 tracked cmd/**/*_test.go file(s) selected)
PASS selection pathspec pin (AC-6/AG-1): select_repo_paths pathspec and should_scan_repo_path prefix set pinned via disk-read, region-anchored extraction
self-test passed: fixtures matched expectations and the tracked tree is clean
```

### `scripts/check-retired-architecture.sh --self-test-integrity` — exit 0

```text
::notice::retired-arch gate mode=self-test-integrity
PASS suite toml non-empty: suite discovered 14 fixture(s)
PASS suite toml has accept and reject coverage: suite has at least one 'accept' and one 'reject' fixture
PASS retired-accept-composed-key-corpus.toml [tomllib]: clean as expected
PASS retired-accept-composed-key-corpus.toml [fallback]: clean as expected
PASS retired-architecture-fixture.toml [tomllib]: rejected as expected
PASS retired-architecture-fixture.toml [fallback]: rejected as expected
PASS retired-malformed-unparseable.toml [tomllib]: rejected as expected
PASS retired-malformed-unparseable.toml [fallback]: rejected as expected
PASS retired-negative-inside-multiline-literal.toml [tomllib]: clean as expected
PASS retired-negative-inside-multiline-literal.toml [fallback]: clean as expected
PASS retired-negative-nested-inert-delimiters.toml [tomllib]: clean as expected
PASS retired-negative-nested-inert-delimiters.toml [fallback]: clean as expected
PASS retired-positive-after-multiline-close.toml [tomllib]: rejected as expected
PASS retired-positive-after-multiline-close.toml [fallback]: rejected as expected
PASS retired-positive-escaped-delimiter-close.toml [tomllib]: rejected as expected
PASS retired-positive-escaped-delimiter-close.toml [fallback]: rejected as expected
PASS retired-reject-concatenated-key.toml [tomllib]: rejected as expected
PASS retired-reject-concatenated-key.toml [fallback]: rejected as expected
PASS retired-reject-dotted-key.toml [tomllib]: rejected as expected
PASS retired-reject-dotted-key.toml [fallback]: rejected as expected
PASS retired-reject-fused-plural-key.toml [tomllib]: rejected as expected
PASS retired-reject-fused-plural-key.toml [fallback]: rejected as expected
PASS retired-reject-hyphenated-key.toml [tomllib]: rejected as expected
PASS retired-reject-hyphenated-key.toml [fallback]: rejected as expected
PASS retired-reject-multiline-open-key.toml [tomllib]: rejected as expected
PASS retired-reject-multiline-open-key.toml [fallback]: rejected as expected
PASS retired-reject-quoted-dotted-key.toml [tomllib]: rejected as expected
PASS retired-reject-quoted-dotted-key.toml [fallback]: rejected as expected
PASS retired-reject-table-plus-key.toml [tomllib]: rejected as expected
PASS retired-reject-table-plus-key.toml [fallback]: rejected as expected
PASS suite go non-empty: suite discovered 19 fixture(s)
PASS suite go has accept and reject coverage: suite has at least one 'accept' and one 'reject' fixture
PASS retired-accept-block-comment.go [go]: clean as expected
PASS retired-accept-comment-only.go [go]: clean as expected
PASS retired-accept-interpreted-string-escaped-quote.go [go]: clean as expected
PASS retired-accept-raw-string-multiline.go [go]: clean as expected
PASS retired-accept-struct-tag-benign.go [go]: clean as expected
PASS retired-accept-struct-tag-non-tag-raw-string.go [go]: clean as expected
PASS retired-accept-token-corpus.go [go]: clean as expected
PASS retired-reject-after-rune-literal.go [go]: rejected as expected
PASS retired-reject-channel-ids.go [go]: rejected as expected
PASS retired-reject-host-clis.go [go]: rejected as expected
PASS retired-reject-ipc-names.go [go]: rejected as expected
PASS retired-reject-slack-team-id.go [go]: rejected as expected
PASS retired-reject-socket-modes.go [go]: rejected as expected
PASS retired-reject-socketmode-lower.go [go]: rejected as expected
PASS retired-reject-socketmode-mixed.go [go]: rejected as expected
PASS retired-reject-socketmode-plural-fused.go [go]: rejected as expected
PASS retired-reject-struct-tag.go [go]: rejected as expected
PASS retired-reject-team-ids-cache.go [go]: rejected as expected
PASS retired-reject-team-ids.go [go]: rejected as expected
PASS suite go-differential non-empty: suite discovered 1 fixture(s)
PASS suite go-differential reject-only exemption: suite is a named reject-only exemption and all 1 fixture(s) are 'reject'
PASS retired-differential-comment-only.go [go-unmasked]: rejected as expected
PASS selection internal non-empty: independently-derived expected internal/** set is non-empty (15 paths)
PASS selection structural inclusion: selected every tracked internal non-test, non-testdata Go file (15 paths)
PASS selection internal exclusions: excluded internal test files and internal testdata paths
PASS selection self-scan guard: did not select scripts/ paths and predicate rejects scripts/testdata/retiredgo/x.go
PASS selection non-empty: selected 21 tracked repo paths
PASS dispatch config.toml.example: engine_for_path routes config.toml.example to the TOML engine
PASS selection includes config.toml.example: real selection set includes config.toml.example end-to-end
PASS selection dispatch coverage: every selected repo path resolves to a known scan engine
PASS selection cmd/ coverage (AG-5/D4): cmd/ deliberately includes *_test.go and testdata/ paths (3 tracked cmd/**/*_test.go file(s) selected)
PASS selection pathspec pin (AC-6/AG-1): select_repo_paths pathspec and should_scan_repo_path prefix set pinned via disk-read, region-anchored extraction
self-test-integrity passed: fixtures matched expectations (repo scan skipped, 015.001-T)
```

### `scripts/check-write-path-precondition.sh --self-test` — exit 0

```text
PASS accept-clean.go: clean as expected
PASS accept-mentions-in-comment.go: clean as expected
PASS reject-createtemp.go: rejected as expected
PASS reject-link.go: rejected as expected
PASS reject-writefile.go: rejected as expected
self-test passed: fixtures matched expectations and the tracked tree is clean
```

## Verbatim output — Post-extraction (3042ae5)

### `scripts/check-retired-architecture.sh` — exit 0

```text
::notice::retired-arch gate mode=repo
```

### `scripts/check-retired-architecture.sh --self-test` — exit 0

```text
::notice::retired-arch gate mode=self-test
PASS suite toml non-empty: suite discovered 14 fixture(s)
PASS suite toml has accept and reject coverage: suite has at least one 'accept' and one 'reject' fixture
PASS retired-accept-composed-key-corpus.toml [tomllib]: clean as expected
PASS retired-accept-composed-key-corpus.toml [fallback]: clean as expected
PASS retired-architecture-fixture.toml [tomllib]: rejected as expected
PASS retired-architecture-fixture.toml [fallback]: rejected as expected
PASS retired-malformed-unparseable.toml [tomllib]: rejected as expected
PASS retired-malformed-unparseable.toml [fallback]: rejected as expected
PASS retired-negative-inside-multiline-literal.toml [tomllib]: clean as expected
PASS retired-negative-inside-multiline-literal.toml [fallback]: clean as expected
PASS retired-negative-nested-inert-delimiters.toml [tomllib]: clean as expected
PASS retired-negative-nested-inert-delimiters.toml [fallback]: clean as expected
PASS retired-positive-after-multiline-close.toml [tomllib]: rejected as expected
PASS retired-positive-after-multiline-close.toml [fallback]: rejected as expected
PASS retired-positive-escaped-delimiter-close.toml [tomllib]: rejected as expected
PASS retired-positive-escaped-delimiter-close.toml [fallback]: rejected as expected
PASS retired-reject-concatenated-key.toml [tomllib]: rejected as expected
PASS retired-reject-concatenated-key.toml [fallback]: rejected as expected
PASS retired-reject-dotted-key.toml [tomllib]: rejected as expected
PASS retired-reject-dotted-key.toml [fallback]: rejected as expected
PASS retired-reject-fused-plural-key.toml [tomllib]: rejected as expected
PASS retired-reject-fused-plural-key.toml [fallback]: rejected as expected
PASS retired-reject-hyphenated-key.toml [tomllib]: rejected as expected
PASS retired-reject-hyphenated-key.toml [fallback]: rejected as expected
PASS retired-reject-multiline-open-key.toml [tomllib]: rejected as expected
PASS retired-reject-multiline-open-key.toml [fallback]: rejected as expected
PASS retired-reject-quoted-dotted-key.toml [tomllib]: rejected as expected
PASS retired-reject-quoted-dotted-key.toml [fallback]: rejected as expected
PASS retired-reject-table-plus-key.toml [tomllib]: rejected as expected
PASS retired-reject-table-plus-key.toml [fallback]: rejected as expected
PASS suite go non-empty: suite discovered 19 fixture(s)
PASS suite go has accept and reject coverage: suite has at least one 'accept' and one 'reject' fixture
PASS retired-accept-block-comment.go [go]: clean as expected
PASS retired-accept-comment-only.go [go]: clean as expected
PASS retired-accept-interpreted-string-escaped-quote.go [go]: clean as expected
PASS retired-accept-raw-string-multiline.go [go]: clean as expected
PASS retired-accept-struct-tag-benign.go [go]: clean as expected
PASS retired-accept-struct-tag-non-tag-raw-string.go [go]: clean as expected
PASS retired-accept-token-corpus.go [go]: clean as expected
PASS retired-reject-after-rune-literal.go [go]: rejected as expected
PASS retired-reject-channel-ids.go [go]: rejected as expected
PASS retired-reject-host-clis.go [go]: rejected as expected
PASS retired-reject-ipc-names.go [go]: rejected as expected
PASS retired-reject-slack-team-id.go [go]: rejected as expected
PASS retired-reject-socket-modes.go [go]: rejected as expected
PASS retired-reject-socketmode-lower.go [go]: rejected as expected
PASS retired-reject-socketmode-mixed.go [go]: rejected as expected
PASS retired-reject-socketmode-plural-fused.go [go]: rejected as expected
PASS retired-reject-struct-tag.go [go]: rejected as expected
PASS retired-reject-team-ids-cache.go [go]: rejected as expected
PASS retired-reject-team-ids.go [go]: rejected as expected
PASS suite go-differential non-empty: suite discovered 1 fixture(s)
PASS suite go-differential reject-only exemption: suite is a named reject-only exemption and all 1 fixture(s) are 'reject'
PASS retired-differential-comment-only.go [go-unmasked]: rejected as expected
PASS selection internal non-empty: independently-derived expected internal/** set is non-empty (15 paths)
PASS selection structural inclusion: selected every tracked internal non-test, non-testdata Go file (15 paths)
PASS selection internal exclusions: excluded internal test files and internal testdata paths
PASS selection self-scan guard: did not select scripts/ paths and predicate rejects scripts/testdata/retiredgo/x.go
PASS selection non-empty: selected 21 tracked repo paths
PASS dispatch config.toml.example: engine_for_path routes config.toml.example to the TOML engine
PASS selection includes config.toml.example: real selection set includes config.toml.example end-to-end
PASS selection dispatch coverage: every selected repo path resolves to a known scan engine
PASS selection cmd/ coverage (AG-5/D4): cmd/ deliberately includes *_test.go and testdata/ paths (3 tracked cmd/**/*_test.go file(s) selected)
PASS selection pathspec pin (AC-6/AG-1): select_repo_paths pathspec and should_scan_repo_path prefix set pinned via inspect.getsource() source-text extraction
self-test passed: fixtures matched expectations and the tracked tree is clean
```

### `scripts/check-retired-architecture.sh --self-test-integrity` — exit 0

```text
::notice::retired-arch gate mode=self-test-integrity
PASS suite toml non-empty: suite discovered 14 fixture(s)
PASS suite toml has accept and reject coverage: suite has at least one 'accept' and one 'reject' fixture
PASS retired-accept-composed-key-corpus.toml [tomllib]: clean as expected
PASS retired-accept-composed-key-corpus.toml [fallback]: clean as expected
PASS retired-architecture-fixture.toml [tomllib]: rejected as expected
PASS retired-architecture-fixture.toml [fallback]: rejected as expected
PASS retired-malformed-unparseable.toml [tomllib]: rejected as expected
PASS retired-malformed-unparseable.toml [fallback]: rejected as expected
PASS retired-negative-inside-multiline-literal.toml [tomllib]: clean as expected
PASS retired-negative-inside-multiline-literal.toml [fallback]: clean as expected
PASS retired-negative-nested-inert-delimiters.toml [tomllib]: clean as expected
PASS retired-negative-nested-inert-delimiters.toml [fallback]: clean as expected
PASS retired-positive-after-multiline-close.toml [tomllib]: rejected as expected
PASS retired-positive-after-multiline-close.toml [fallback]: rejected as expected
PASS retired-positive-escaped-delimiter-close.toml [tomllib]: rejected as expected
PASS retired-positive-escaped-delimiter-close.toml [fallback]: rejected as expected
PASS retired-reject-concatenated-key.toml [tomllib]: rejected as expected
PASS retired-reject-concatenated-key.toml [fallback]: rejected as expected
PASS retired-reject-dotted-key.toml [tomllib]: rejected as expected
PASS retired-reject-dotted-key.toml [fallback]: rejected as expected
PASS retired-reject-fused-plural-key.toml [tomllib]: rejected as expected
PASS retired-reject-fused-plural-key.toml [fallback]: rejected as expected
PASS retired-reject-hyphenated-key.toml [tomllib]: rejected as expected
PASS retired-reject-hyphenated-key.toml [fallback]: rejected as expected
PASS retired-reject-multiline-open-key.toml [tomllib]: rejected as expected
PASS retired-reject-multiline-open-key.toml [fallback]: rejected as expected
PASS retired-reject-quoted-dotted-key.toml [tomllib]: rejected as expected
PASS retired-reject-quoted-dotted-key.toml [fallback]: rejected as expected
PASS retired-reject-table-plus-key.toml [tomllib]: rejected as expected
PASS retired-reject-table-plus-key.toml [fallback]: rejected as expected
PASS suite go non-empty: suite discovered 19 fixture(s)
PASS suite go has accept and reject coverage: suite has at least one 'accept' and one 'reject' fixture
PASS retired-accept-block-comment.go [go]: clean as expected
PASS retired-accept-comment-only.go [go]: clean as expected
PASS retired-accept-interpreted-string-escaped-quote.go [go]: clean as expected
PASS retired-accept-raw-string-multiline.go [go]: clean as expected
PASS retired-accept-struct-tag-benign.go [go]: clean as expected
PASS retired-accept-struct-tag-non-tag-raw-string.go [go]: clean as expected
PASS retired-accept-token-corpus.go [go]: clean as expected
PASS retired-reject-after-rune-literal.go [go]: rejected as expected
PASS retired-reject-channel-ids.go [go]: rejected as expected
PASS retired-reject-host-clis.go [go]: rejected as expected
PASS retired-reject-ipc-names.go [go]: rejected as expected
PASS retired-reject-slack-team-id.go [go]: rejected as expected
PASS retired-reject-socket-modes.go [go]: rejected as expected
PASS retired-reject-socketmode-lower.go [go]: rejected as expected
PASS retired-reject-socketmode-mixed.go [go]: rejected as expected
PASS retired-reject-socketmode-plural-fused.go [go]: rejected as expected
PASS retired-reject-struct-tag.go [go]: rejected as expected
PASS retired-reject-team-ids-cache.go [go]: rejected as expected
PASS retired-reject-team-ids.go [go]: rejected as expected
PASS suite go-differential non-empty: suite discovered 1 fixture(s)
PASS suite go-differential reject-only exemption: suite is a named reject-only exemption and all 1 fixture(s) are 'reject'
PASS retired-differential-comment-only.go [go-unmasked]: rejected as expected
PASS selection internal non-empty: independently-derived expected internal/** set is non-empty (15 paths)
PASS selection structural inclusion: selected every tracked internal non-test, non-testdata Go file (15 paths)
PASS selection internal exclusions: excluded internal test files and internal testdata paths
PASS selection self-scan guard: did not select scripts/ paths and predicate rejects scripts/testdata/retiredgo/x.go
PASS selection non-empty: selected 21 tracked repo paths
PASS dispatch config.toml.example: engine_for_path routes config.toml.example to the TOML engine
PASS selection includes config.toml.example: real selection set includes config.toml.example end-to-end
PASS selection dispatch coverage: every selected repo path resolves to a known scan engine
PASS selection cmd/ coverage (AG-5/D4): cmd/ deliberately includes *_test.go and testdata/ paths (3 tracked cmd/**/*_test.go file(s) selected)
PASS selection pathspec pin (AC-6/AG-1): select_repo_paths pathspec and should_scan_repo_path prefix set pinned via inspect.getsource() source-text extraction
self-test-integrity passed: fixtures matched expectations (repo scan skipped, 015.001-T)
```

### `scripts/check-write-path-precondition.sh --self-test` — exit 0

```text
PASS accept-clean.go: clean as expected
PASS accept-mentions-in-comment.go: clean as expected
PASS reject-createtemp.go: rejected as expected
PASS reject-link.go: rejected as expected
PASS reject-writefile.go: rejected as expected
self-test passed: fixtures matched expectations and the tracked tree is clean
```

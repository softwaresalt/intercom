---
title: "Runtime verification: 016-S / 017-F — pathsafe error-surface parity and Windows long-path precision"
description: "Runtime verification report for shipment 016-S / feature 017-F"
status: "complete"
tags:
  - "runtime-verification"
  - "016-S"
  - "017-F"
date: 2026-09-11
shipment: 016-S
feature: 017-F
pr: 51
merge_commit_sha: 75cfe3b1b954c49d149f531464ccd97f1770dd0b
verdict: READY
---

# Runtime verification: 016-S / 017-F

## Scope of change

`internal/pathsafe` error-surface parity (`wrapPathError` routing for
`checkSymlinkEscape`'s reparse error branch) and Windows long-path precision
(`addLongPathPrefix`'s MAX_PATH threshold now UTF-16-code-unit-aware via
`utf16Len`), plus a corrected characterization test/doc comment for the
`\\.\` device-namespace exclusion branch. No containment-verdict change, no
new filesystem write primitive, no new CLI flag, config schema, or
user-facing behavior surface of its own.

## Runtime surface analysis

`internal/pathsafe` has no direct runtime entrypoint of its own. Its only
wired caller in this repository is `internal/config/validate.go`
(`default_workspace_root` / `[[workspace]].path` validation), which is
exercised at `cmd/intercom` / `cmd/intercom-ctl` startup whenever a config
file with workspace paths is loaded. This is the same indirect `cli`-surface
touch documented by the immediately preceding shipment
(`docs/closure/2026-09-10-015-s-016-f-pathsafe-windows-canonicalization-symmetry-runtime-verification.md`).

## Evidence

- `go build ./...`, `go vet ./...`, `gofmt -l .` — all clean on the
  post-merge closure branch (`post-merge/017-f-pathsafe-error-surface-parity`,
  built on `main` @ `75cfe3b1`).
- `go test ./...` — all packages green, including `internal/pathsafe` and
  `internal/config`, which directly exercise the changed
  `checkSymlinkEscape` / `addLongPathPrefix` code paths.
- `go test ./internal/pathsafe/... -race` — clean (re-verified pre-merge on
  the feature branch; the change set introduces no new goroutines or shared
  mutable state, so this was not re-run again post-merge as no code changed
  between the final race-clean run and the merge commit).
- CLI smoke: `go run ./cmd/intercom --help` and
  `go run ./cmd/intercom-ctl --help` both exit 0 with usage text printed, no
  panic, on the post-merge closure branch.
- Windows-tagged tests (`symlink_windows_test.go`,
  `reparse_windows_test.go`) ran as part of the full `go test ./...` pass on
  this Windows host, directly exercising both changed functions.

## TUI / WebSocket / Dev-Tunnel manual checkpoints

Not applicable — this shipment's diff does not reach any of those surfaces.

## Verdict

**READY.** The change is confined to `internal/pathsafe`'s error-wrapping
and long-path-threshold internals, reached only indirectly through
`internal/config` validation at process startup. Full green build/vet/test
evidence plus CLI smoke checks are sufficient for this narrow, indirect
touch. No conditions.

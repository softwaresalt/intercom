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
(`default_workspace_root` / `[[workspace]].path` validation), invoked from
`internal/config.Load`. **Correction (Copilot review, PR #52):** as of this
shipment, `cmd/intercom`'s `RunE` does not call `config.Load`/`Validate` at
all — it returns `errNotImplemented` immediately after logging, and
`cmd/intercom-ctl` was independently checked and shows the same
not-yet-wired state. Verified by repo-wide search: no non-test caller of
`config.Load` or `(*Config).Validate` exists outside `internal/config`
itself. The prior closure artifact for 015-S/016-F asserted the same
"reached at cmd startup" framing; that framing was aspirational, not a
verified current runtime path, and is corrected here rather than repeated.
The actual current evidence for the changed code paths is the Go test
suite (`internal/pathsafe` and `internal/config`, which call the changed
functions directly), not process startup.

## Evidence

- `go build ./...`, `go vet ./...`, `gofmt -l .` — all clean on the
  post-merge closure branch (`post-merge/017-f-pathsafe-error-surface-parity`,
  built on `main` @ `75cfe3b1`).
- `go test ./...` — all packages green, including `internal/pathsafe` and
  `internal/config`, which directly exercise the changed
  `checkSymlinkEscape` / `addLongPathPrefix` code paths. **This is the
  primary evidence for the changed code**, since neither `cmd/intercom` nor
  `cmd/intercom-ctl` currently invokes `config.Load`/`Validate` at process
  startup (see correction above).
- `go test ./internal/pathsafe/... -race` — clean (re-verified pre-merge on
  the feature branch; the change set introduces no new goroutines or shared
  mutable state, so this was not re-run again post-merge as no code changed
  between the final race-clean run and the merge commit).
- CLI smoke (basic process-startup sanity check only — does **not**
  exercise the changed `internal/pathsafe` code, since it never reaches
  `config.Load`/`Validate`): `go run ./cmd/intercom --help` and
  `go run ./cmd/intercom-ctl --help` both exit 0 with usage text printed, no
  panic, on the post-merge closure branch.
- Windows-tagged tests (`symlink_windows_test.go`,
  `reparse_windows_test.go`) ran as part of the full `go test ./...` pass on
  this Windows host, directly exercising both changed functions.

## TUI / WebSocket / Dev-Tunnel manual checkpoints

Not applicable — this shipment's diff does not reach any of those surfaces.

## Verdict

**READY.** The change is confined to `internal/pathsafe`'s error-wrapping
and long-path-threshold internals. `internal/pathsafe` has no runtime
entrypoint of its own and, as of this shipment, no `cmd/` entrypoint
actually calls `config.Load`/`Validate` at process startup (verified by
repo-wide search — see correction above), so the changed code is not
reached at runtime today. Full green build/vet/test evidence — the Go test
suite directly exercising both changed functions — is the substantive
evidence. CLI smoke checks (`--help`) confirm basic process-startup health
only and are not evidence for the changed paths. No conditions.

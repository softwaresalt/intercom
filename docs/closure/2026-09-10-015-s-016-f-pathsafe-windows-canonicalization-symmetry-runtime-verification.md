---
title: "Runtime verification: 015-S / 016-F — pathsafe Windows canonicalization symmetry and Windows lint coverage"
description: "Runtime/black-box verification report for shipment 015-S / feature 016-F"
status: "complete"
date: 2026-09-10
mode: post-merge
shipment: 015-S
feature: 016-F
pr: 48
merge_commit_sha: fbd8d8d0ab679e254b2b817af69b8e8e673626c5
verdict: READY
---

# Runtime verification: 015-S / 016-F — pathsafe Windows canonicalization symmetry and Windows lint coverage

## Surface under change

`internal/pathsafe` (root.go, pathsafe.go, reparse_windows.go,
reparse_other.go) is an internal library package with no direct CLI/API
entrypoint of its own. Its only wired runtime caller in this repository is
`internal/config/validate.go`, which routes `default_workspace_root` and
every `[[workspace]].path` config field through `pathsafe.NewRoot` /
`Root.Resolve` during config validation — a step both `cmd/intercom` (the
server entrypoint) and `cmd/intercom-ctl` (the companion CLI) execute at
startup when a config file is loaded.

This is a narrow, indirect runtime-surface touch (the `cli` surface via
config validation), not a change to the WebSocket/API surface, the TUI, or
Dev Tunnel — none of which this shipment's diff touches.

## Verdict: READY

- **CLI smoke evidence** (per `runtime_validation.validator_manifest`,
  probe `cli-smoke`): both entrypoints start and print usage without panic
  at merged HEAD `fbd8d8d0ab679e254b2b817af69b8e8e673626c5`:
  - `go run ./cmd/intercom --help` — exit 0, usage printed.
  - `go run ./cmd/intercom-ctl --help` — exit 0, usage printed.
- **Regression evidence for the actual wired call path**
  (`internal/config/validate.go` → `pathsafe.NewRoot`): `go test ./...`
  (including `internal/config`'s own validate tests and
  `internal/pathsafe`'s full suite, both GOOS variants) is green at merged
  HEAD — re-verified on the post-merge closure branch, built on `main` @
  `fbd8d8d0`.
- **Full CI check suite** on PR #48 at merge HEAD
  (`323e8520cbd5cf15c6b3954018c0d23ef1ff57a3`): all 13 checks pass
  (cross-compile ×4, lint, security, test, test (windows, advisory),
  pipeline-topology, ci gate, gitignore regression, detect-changes,
  load-cross-compile-targets).
- **Manual checkpoints** (`tui-render`, `devtunnel-connect`,
  `tool-approval-routing`): **N/A for this shipment.** None of these
  checkpoints exercise `internal/pathsafe` or `internal/config/validate.go`;
  this shipment touches neither the Bubble Tea TUI, the WebSocket/Dev
  Tunnel surface, nor the tool-approval routing logic. No manual checkpoint
  evidence is required or applicable.
- **Public-API surface** (`runtime_surfaces.public_api: true`,
  WebSocket): not exercised by this change — no code under this shipment's
  diff is reachable from the WebSocket hub or JSON-RPC decoding path.

## Compensating/prior evidence

- `go build ./...`, `go vet ./...`, `gofmt -l .` — all clean at merged HEAD
  and re-verified on the post-merge closure branch.
- The two new C2 tests (`TestNewRootErrorSurfaceIsPathError`,
  `TestNewRootResolvesEndToEndUnderVolumeGUIDRoot`) and the two C1 junction
  RED-lock tests now GREEN
  (`TestNewRootOnJunctionRootedWorkspaceResolvesExistingDescendant`,
  `TestNewRootOnJunctionRootedWorkspaceResolvesNonExistentLeaf`) all pass,
  directly exercising the exact `NewRoot`/`canonicalizeReparse` code path
  that `internal/config/validate.go` calls at startup.

## Follow-up

None required for runtime verification. The narrow `cli`-surface touch
(config validation at startup) is fully covered by existing and new
`internal/pathsafe` / `internal/config` test coverage plus the direct
`--help` smoke probes above; no additional manual verification is
outstanding.

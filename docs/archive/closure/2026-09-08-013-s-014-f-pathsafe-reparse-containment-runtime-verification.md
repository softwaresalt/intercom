# Runtime verification: 013-S / 014-F — pathsafe reparse-point containment hardening

- Surface: `cli` (target adapter), `auto` mode
- Context: PR #40, merge commit `17daffb5614f2f8ed797d7647aeadcc1dc29b1d2`
- Date: 2026-09-08

## Step 1-2: Validator contract and environment prechecks

The shipment modifies `internal/pathsafe.checkSymlinkEscape` /
`canonicalizeReparse`, reached exclusively through `pathsafe.Root.Resolve()`.
Environment precheck traced every current caller of the `internal/pathsafe`
package to determine whether a live, runnable entrypoint exists to exercise
this exact code path black-box:

- `cmd/intercom/main.go`: `newRootCmd().RunE` logs the config path and
  immediately returns `errNotImplemented` — it does not call
  `internal/config.Load` or `Validate` at all yet.
- `cmd/intercom-ctl/main.go`: reviewed; no `internal/config` or
  `internal/pathsafe` import.
- `internal/config/validate.go` (the only non-test, non-pathsafe-package
  caller of `internal/pathsafe`): calls `pathsafe.NewRoot(...)` for
  `default_workspace_root` and each `[[workspace]].path` entry. It does
  **not** call `Root.Resolve()` anywhere — `NewRoot`'s own canonicalization
  (`filepath.EvalSymlinks`) is a *different* code path from the one this
  shipment changed (`checkSymlinkEscape`/`canonicalizeReparse`, reached only
  via `Root.Resolve()`).

**Missing prerequisite**: there is currently no wired, runnable command,
server loop, or exported entrypoint anywhere in `cmd/**` that invokes
`pathsafe.Root.Resolve()` — the exact function this shipment hardens. This
is a pre-existing project-maturity gap (the server's file-operation surface,
which will eventually call `Resolve()` to check a session-relative path
against its workspace root, has not landed yet — confirmed by the README's
own "Config loading is not yet wired into the running server" note) and is
not introduced, worsened, or masked by 013-S.

## Step 3-6: Adapter selection and verdict

- `command` adapter: **not available** — no command exists that reaches
  `Root.Resolve()`.
- `api` / `browser` / `job` adapters: not applicable — no HTTP server,
  browser UI, or background job surface exists in this repository yet.
- `manual` adapter: not attempted — there is no operator-facing surface to
  manually exercise either.

**Verdict: `BLOCKED`** — meaningful black-box/live verification cannot
proceed because the runtime surface this fix protects (`Root.Resolve()`
invoked from an actual running process) has no wiring yet anywhere in the
codebase. This is recorded as a closure condition, not treated as a pass.

## Compensating evidence (does not substitute for the BLOCKED verdict)

The function-level behavior of the modified code is exhaustively verified
by the existing automated test suite, run and passing on Windows as part of
this shipment's own quality gates (see PR #40 and
`docs/closure/2026-09-08-013-s-014-f-pathsafe-reparse-containment-adversarial-review.md`):

- 45/45 tests in `internal/pathsafe` (including 3 new `Resolve()`-level
  regression locks added during adversarial review: in-root live-junction
  acceptance, out-of-root live-junction+non-existent-leaf rejection,
  dangling-junction-as-final-component rejection).
- Full `go test ./...`, including `tests/integration`, green.
- Cross-platform build proof (`linux/amd64`, `darwin/arm64`, `windows/amd64`).

This is genuine function-level/library verification, not a fabricated
runtime pass — it is explicitly not claimed to be a substitute for exercising
`Root.Resolve()` through a live process, which remains impossible until a
caller exists.

## Follow-up recommendation

Stashed as a follow-up (see closure artifact's "Stash follow-up items"
section): once the server's file-operation / session-path-resolution surface
lands and calls `pathsafe.Root.Resolve()` for the first time, add a
black-box/integration-level runtime verification (real process, real
directory junction, real request) proving the containment fix holds through
the actual wired call path — not just at the unit-test level. This is a
distinct, separately-timed follow-up because the caller does not exist yet;
it is not a residual risk in the code shipped today.

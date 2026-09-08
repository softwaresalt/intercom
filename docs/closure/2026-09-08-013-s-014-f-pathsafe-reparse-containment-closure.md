# Post-merge closure: 013-S / 014-F — pathsafe reparse-point containment hardening

- Mode: `post-merge`
- PR: #40 (`fix(pathsafe): close Windows directory-junction containment bypass (013-S/014-F)`)
- Merge commit: `17daffb5614f2f8ed797d7647aeadcc1dc29b1d2` (merge-commit strategy, P-009)
- Date: 2026-09-08
- Compaction status (P-020): `pending` (finalized by Ship after `compact-context` runs — see Step 8 of the Ship pipeline)

## Summary of the change

Closed a Windows directory-junction path-containment bypass in
`internal/pathsafe` (`checkSymlinkEscape` used `filepath.EvalSymlinks`, which
never resolves `IO_REPARSE_TAG_MOUNT_POINT`). Introduced `canonicalizeReparse`
(direct `GetFinalPathNameByHandleW` call), wired it into the containment
check's ancestor-climb loop, flipped a fail-open terminal branch to
fail-closed, and reconciled the package's risk register and doc comments.
Strict test-first development across all 8 manifest tasks (014.001-T…
014.008-T). Full detail in PR #40's body and the task-by-task table there.

## CI status and unresolved review items

- All CI checks green at merge HEAD (`3b929ad1325dc01018e58f4d039af67b9587b96a`):
  cross-compile ×4, lint, security, test, test (windows, advisory),
  pipeline-topology, ci gate.
- Copilot review: `SATISFIED` (P-018 gate) — 1 round, 1 thread (a false-positive
  `make()`/`uintptr` compile-error claim), replied with a verified rationale
  (Go spec + empirical repro) and resolved. No unresolved Copilot threads at
  merge.
- Standard code review + 3-reviewer adversarial review: both
  `READY_WITH_FOLLOWUPS`. All in-scope findings (P-021 C1) fixed directly; 3
  out-of-scope findings deferred via P-021 capture (stash `4104AF54`,
  `E428AB46`, `E4C5413F` — see PR #40 body "Review remediation" section for
  full disposition).
- No unresolved P0/P1 findings at merge.

## Runtime verification report

See `docs/closure/2026-09-08-013-s-014-f-pathsafe-reparse-containment-runtime-verification.md`.

**Verdict: `BLOCKED`** — the exact function this shipment hardens
(`pathsafe.Root.Resolve()`) has no wired live entrypoint anywhere in
`cmd/**` yet (`cmd/intercom`'s `RunE` is unimplemented;
`internal/config.Validate()` only calls the unrelated `pathsafe.NewRoot`).
This is a pre-existing project-maturity gap, not introduced or masked by
013-S. Compensating function-level evidence (45/45 pathsafe tests including
3 new `Resolve()`-level regression locks, full `go test ./...`, cross-platform
build) is documented in the runtime-verification report and is not treated
as a substitute pass.

## Invariants to preserve

- `pathsafe.Root.Resolve()` must reject any candidate path whose real
  (reparse-resolved) target lies outside the root, regardless of whether the
  reparse point (`IO_REPARSE_TAG_MOUNT_POINT` or `IO_REPARSE_TAG_SYMLINK`) is
  an intermediate or the terminal path component.
- `pathsafe.Root.Resolve()` must accept a benign in-root reparse-point shape
  (a live junction/symlink whose real target still resolves inside the root)
  — this shipment's C1 fix and its incidental false-rejection fix (F3-adjacent,
  see PR #40 body) both depend on this remaining true.
- The Windows-only test job (`test (windows, advisory)`) remains
  `continue-on-error` per `WINDOWS_GATE_REQUIRED` being unset (plan §7,
  deliberately not flipped by this shipment) — a future change to this CI
  variable is out of this shipment's scope.

## Pre-deploy audits

Not applicable — this is a merge-only library change to an internal Go
package with no currently-wired runtime entrypoint (see runtime-verification
report). No migration, flag, config-schema, or access-control change
accompanies this shipment.

## Deployment / rollout path

**Merge-only.** No deployment, canary, or phased rollout exists for this
repository at its current maturity (no running service; `cmd/intercom`'s
`RunE` is unimplemented). The change lands in `main` and becomes part of the
next `go build`/`go install` of this module; it has no independent release
or rollout event of its own.

## Post-deploy checks

Not applicable in the traditional sense (no live deployment). The concrete
next check is functional, not operational: when a future feature wires
`internal/config.Load`/`Validate` (or any new caller) into an actual running
command, that change's own runtime verification must exercise
`pathsafe.Root.Resolve()` against a real directory-junction escape attempt
end-to-end (see Follow-up recommendation below and the stashed follow-up
item).

## Risky action record

None. No destructive, migration, or irreversible action was taken as part of
this shipment or its closure. The shipment safe-close used the P-015 verified
fully-covered-root cascade path (`backlogit shipment ship`), which was
independently verified via the deterministic classifier
(`classify_shipment_close_path`) before invocation and fully verified
post-invocation (empty `returned_ids`, exact two-set `allowed_ids`/
`required_ids` match, `parent_id` preservation on all 8 tasks) — see
`.backlogit/reconcile/013-S-safe-close-2026-09-08-1615.md`.

## Healthy signals

- `go build ./...`, `go vet ./...`, `gofmt -l .`, `golangci-lint run ./...`,
  and `go test ./...` remain green on `main` after merge (re-verified below).
- No new golangci-lint findings on the Windows-tagged file introduced by this
  shipment slip through CI's ubuntu-only lint job undetected beyond the
  already-stashed gap (`4104AF54`).

## Failure signals

- A future regression in `internal/pathsafe`'s containment logic that
  reintroduces the C1 bypass, or any of the 3 new `Resolve()`-level
  regression-lock tests (in-root live-junction acceptance, out-of-root
  live-junction+non-existent-leaf rejection, dangling-junction-as-final-
  component rejection) failing.
- Any future PR that wires `pathsafe.Root.Resolve()` into a live command
  without first re-running black-box runtime verification against a real
  junction escape attempt.

## Monitoring plan

Not applicable at runtime (no live service). The durable "monitoring" for
this shipment is the regression-lock test suite itself
(`internal/pathsafe/junction_windows_test.go`,
`internal/pathsafe/reparse_windows_test.go`), which runs on every CI build
of `main` going forward (`test (windows, advisory)` job).

## Rollback trigger / procedure

Not applicable — merge-only, no live rollout to roll back. If the change is
later found to be defective, the standard remedy is a follow-up PR reverting
or fixing the specific regression, gated through the same review/CI process,
not an operational rollback.

## Validation window

Not applicable — no live deployment window. The validation window that does
apply is: the next time any code wires `pathsafe.Root.Resolve()` into a live
entrypoint, that PR's own runtime verification must close the BLOCKED gap
recorded above.

## Owner

Repository maintainer (`softwaresalt`) for both the merged change and the
stashed follow-up recommending live-surface runtime verification once a
caller exists.

## Releasability evidence

**READY_WITH_CONDITIONS.**

- Condition 1 (informational, not blocking): runtime verification for
  `pathsafe.Root.Resolve()` is `BLOCKED` due to no wired live entrypoint —
  this is a pre-existing, unrelated project-maturity gap, not a defect in
  this shipment, and does not block merge or this closure. It is recorded so
  it is not silently lost, and a follow-up item is stashed to close the gap
  when the relevant caller lands.
- Condition 2: the three P-021 stash entries (`4104AF54`, `E428AB46`,
  `E4C5413F`) remain open follow-up work, none P0/P1, all already recorded
  in the PR body's Local Review Readiness follow-up field.

No condition blocks the merge already completed; both are forward-looking
follow-up tracking.

## Post-merge re-verification (on `main` after merge)

Re-ran the full quality-gate sequence on `main` @ `17daffb5` (post-merge
closure branch `post-merge/013-s-pathsafe-reparse-containment`, prior to any
closure-only commits): `go build ./...`, `go vet ./...`, `gofmt -l .`, and
`go test ./...` — all green, confirming the merge itself introduced no
regression. Detailed results recorded in the session memory checkpoint.

## Stash follow-up items

- stash `37FAB8C2`: add black-box/integration-level runtime verification for
  `pathsafe.Root.Resolve()` once a live caller exists (closes the `BLOCKED`
  runtime-verification gap recorded above). Provisional priority: medium.

## Source artifact cleanup

Checked `custom_fields.source_stash_id` and `custom_fields.source_deliberation_id`
on both shipped top-level items (`014-F`, `013-S`) after archival. Neither
field is present on either record — this feature/shipment was assembled
directly from `docs/decisions/2026-09-08-intercom-go-pathsafe-reparse-containment-deliberation.md`
and `docs/plans/2026-09-08-intercom-go-pathsafe-reparse-containment-plan.md`
via `references`, not via a `custom_fields`-linked stash or deliberation
artifact ID. No stash or deliberation archival action was applicable or
taken.

## Feed Back Into the Harness

No durable runtime, deployment, or workflow learning surfaced beyond what is
already captured in the review/adversarial-review artifacts and the stashed
follow-up items. No documentation beyond this closure artifact and the
already-updated `internal/pathsafe/root.go` risk register requires further
graduation for this shipment.

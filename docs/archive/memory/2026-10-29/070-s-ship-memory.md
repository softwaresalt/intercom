# 070-S (077-F) ship session memory

Status: complete through post-merge closure PR preparation.

## Items completed

* 077.001-T `87e5c1e`: Zl/Zp sanitization, `sanitizeIdent` (128-rune cap), pre-send `gone()` re-check.
* 077.002-T `52fb8bf`: latch quiesce on failed Start, informational `session.shutdown`, late terminal idle delivery.
* 077.003-T `cfeb75e`: `SDKStopLockTimeout`, `forceStopWait`, `ForceStopBudget`, hub shutdown budget folding.
* 077.004-T `855bacf`: Hub.Run error recorded in `App.Err`.
* Follow-up commits: `3f89a09` (review test hardening), `69bbf9f` (Copilot thread: hub-level cutoff regression).

## Branch and PR state

* Feature branch merged via PR #210 (merge commit af87a21, merge strategy: merge commit, P-009).
* Closure branch `post-merge/077-f-copilot-adapter-stop-forcestop-resilience-latch-permission-hardening-and-app-err-stuck-lane-reporting`.

## Decisions

See the compacted summary. Escalations: none. Circuit breakers: none tripped.

## Errors and resolutions

* Relayed cascade-close flags `--approved-by`/`--approval-mode` unsupported by the installed CLI (exit 2, no mutation); reran without them.
* CI `tools/gatecheck` wrapper timing flake; passed on rerun.

## Next steps

Closure PR review gates and merge, then 071-S and 072-S.

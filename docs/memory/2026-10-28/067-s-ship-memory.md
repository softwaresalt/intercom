---
title: "067-S Ship session memory"
date: 2026-10-28
shipment: 067-S
feature: 074-F
---

## Status

* Branch: `feat/067-s-hub-fold-publish-and-lifecycle-hardening`
* Dark mode (P-017) active; explicit operator P-001 exception over `062-S`
* Tasks `074.001-T` through `074.005-T` implemented test-first and archived as done

## Commits

| Task | Code commit | Backlog commit |
|---|---|---|
| 074.001-T | 664690d | e256c08 |
| 074.002-T | 8a0175c | b68cc84 |
| 074.003-T | b992e05 | 769cae3 |
| 074.004-T | 38efce4 | 5064387 |
| 074.005-T | c0f2df5 | eae78ab |

## Decisions

* `apperr.KindIO` used for the bounded lane-join error because `apperr.KindInternal`
  does not exist; adding a kind is outside the task file list
* `internal/hub/permission.go` also touched (`intakeAsk`, `permReasonTurnAborted`)
  because `handleAsk` lives there
* `gofmt -l` reports only the pre-existing `logs/drive/main.go`; out of scope

## Local review

* Concurrency and correctness persona reviews: no P0-P2 findings
* P3 advisories recorded as residual risk (abortedTurnIDs cap eviction,
  empty-TurnID ask matching, dual non-blocking sends on action failure)

## Next steps

* Open the feature PR, pass CI and P-018 Copilot gate, merge with a merge commit
* Run post-merge closure on `post-merge/074-f-hub-fold-publish-and-lifecycle-hardening`

# 067-S (074-F) compacted memory

* Scope: hub fold, publish and lifecycle hardening (stash 2C74C5FC, 550437DB, 8769C9A0). Tasks 074.001-T (`664690d`), 074.002-T (`8a0175c`), 074.003-T (`b992e05`), 074.004-T (`38efce4`) and 074.005-T (`c0f2df5`) done, implemented test-first. Branch `feat/067-s-hub-fold-publish-and-lifecycle-hardening`, reviewed HEAD `b7a2934658d4d97af08e1997ec6ae09dabc737cb`. PR #202 merged as a merge commit a2cacf32c29b5e3a0ee596f01f14686ceeb3775d.
* Decisions: `apperr.KindIO` used for the bounded lane-join error because `apperr.KindInternal` does not exist (adding a kind is outside the task file list). `internal/hub/permission.go` also touched (`intakeAsk`, `permReasonTurnAborted`) because `handleAsk` lives there. `gofmt -l` reports only the pre-existing `logs/drive/main.go`; out of scope. Explicit operator P-001 exception over 062-S.
* Review: concurrency and correctness persona reviews found no P0-P2. P3 advisories kept as residual risk: abortedTurnIDs cap eviction, empty-TurnID ask matching, dual non-blocking sends on action failure.
* Closure: cascade close of 067-S exited 0 (archived 067-S shipped, 074-F, 074.001-T..074.005-T); evidence `docs/closure/evidence/067-S-074-F-close-evidence.json`.
* Next: closure PR, CI, Copilot gate, merge.
* Original: docs/archive/memory/2026-10-28/067-s-ship-memory.md.

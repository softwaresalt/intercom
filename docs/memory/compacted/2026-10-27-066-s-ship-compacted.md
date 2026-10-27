# 066-S (073-F) compacted memory

* Scope: zero-behavior-change maintainability refactors of `internal/app`. Tasks 073.001-T (`b9b0ffe`), 073.002-T (`db79f5b`) and 073.003-T (`e8236a9`) done, plus local-review remediation `3e61cc7`. Branch `feat/066-s-maintainability-refactors-of-internal-app` from main `fed1f84`; claim commit `3537d88` after pre-reconcile PROCEED and topology pre_claim/post_claim exit 0. PR #198 merged as a merge commit 65780da0d83a53182a56460093b80cc132b5d945.
* Harness: characterization-first. The existing `go test -race ./internal/app/...` suite was confirmed green before and after each refactor; two small characterization tests added (`late_agent_test.go`, `reap_test.go`). harness-architect was not run as a separate skill invocation for these refactors and Ship applied no `harness-ready` label.
* Decisions: `reap(rctx)` is clearCreds, stopRemote, awaitHub, stopUI (normal supervise ordering). Failed-start `teardown` composes the same helpers in its own original order (hub BeginShutdown, UI, remote, agent, hub) because the agent stops between the UI/remote steps and the hub wait. `hubCancel` now precedes `<-hubRunDone` in `teardown` (hub.Stop is idempotent; behavior-neutral).
* Review: Adversarial Review (3 reviewers) plus Correctness, Go, Concurrency, Maintainability, Architecture and Security personas; no P0/P1; in-scope P3 fixed.
* Deferred scope (stash, P-021 C2): 4B7308F2, 2F9C75F8, 26E15024.
* Gates: gofmt, vet, build, `go test -race ./...`, integration, gatecheck, write-path self-test, golangci-lint (0 issues), govulncheck (clean).
* Next: post-merge closure for 066-S (closure PR, CI, Copilot loop).
* Original: docs/archive/memory/2026-10-27-066-s-ship-memory.md.

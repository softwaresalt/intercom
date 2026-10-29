# Ship session memory: 068-S (075-F)

* shipment: 068-S, covering feature 075-F (shutdown broadcast/drain fidelity, in-place hub Client.Rehydrate)
* branch: feat/068-s-shutdown-broadcast-drain-fidelity-and-in-place-hub-client-rehydrate
* mode: DARK_MODE_ACTIVE (P-017); merge_approval_pre_authorized, admin fallback NOT authorized

## Completed

| task | commit | summary |
|---|---|---|
| 075.001-T | 09bfe4e | intake sweep, hasPendingConnect reap guard, terminal server.shutdown delivery (inflight/sendFrame/deliverTerminal) |
| 075.002-T | 46d5f0a | write-ack Drained tracking, ctlRehydrate, in-place Client.Rehydrate |
| 075.003-T | 7f0984c | session.writePump EnableWriteAck/AckWrite, unconditional adapterStopTimeout |
| 075.004-T | bf8890a | mint(ctx) honours cancelled ctx; shutdownScenario drains remote concurrently |

Each task followed red-then-green: failing tests confirmed (vet green, test red) before the production change. For 075.002-T and 075.003-T/075.004-T the `harness-ready` label was applied immediately after red was confirmed but after the production edit had already been written (before running it): a process-order slip recorded here.

## Decisions

* Terminal clients (server.shutdown queued) still receive later frames while the buffer has room and silently skip them when full. Never drained, never marked stale. The plan text said "skip subsequent broadcasts", but existing tests (agent.idle after shutdown, teardown events) require post-shutdown frames.
* Added `reapUnlessConnecting` as the single guarded reap used at both Run call sites.
* handleRehydrate refuses terminal clients and does not double-queue stale entries (review P3 fixes).

## Review

Local review: concurrency + correctness reviewers, no P0/P1. P3 follow-ups:
* Deferred (P-021 C2) stash entry 80996C79: Cursor/wsclient gap handling may discard the server.shutdown frame.
* Advisory, not actioned: `mint` takes mintMu without selecting on ctx; Drained assumes write-ack is enabled before any frame is dequeued (true in newSession).

## Next

Full gates green (gofmt, vet, build, go test -race ./..., gatecheck write-path self-test). Push, PR, Copilot review (P-018), merge-commit merge, then post-merge closure.

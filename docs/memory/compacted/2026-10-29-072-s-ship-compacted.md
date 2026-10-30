# 072-S (079-F) compacted memory

* Scope: `agent.Event` rune-safe bounds (`boundUTF8`, `Truncated`), `tool.ended` and reasoning enrichment (`agent.reasoning`, `agent.reasoning.delta`), write-diff and shell `possibleUrls` surfacing in `PermissionRequest.Summary`, and rate-limited unmapped-event warns (adapter `unmappedLogGate` and hub). Tasks 079.001-T to 079.006-T done; design sections 3.1, 4.2, 4.3 and 5.2 amended. PR #214 merged as a merge commit d0d863fae5ec4bdc1d197b35800223a786ed87c5 (reviewed HEAD 54e4988).
* Decisions: the clipped write preview keeps `out.Truncated=false` per the plan, and failing it closed is out of scope (P-021 C1), captured as stash 773B2CB2 (bug, high, requires deliberation). `boundUTF8` look-back is bounded, fold flags turn truncation, and clipped text and the hub unmapped-type key are cloned so oversized SDK strings are not pinned.
* Closure: a1 moved 079-F active to done; cascade close of 072-S exited 0 (archived 072-S shipped, 079-F, 079.001-T to 079.006-T); evidence `docs/closure/evidence/072-S-079-F-close-evidence.json`. Live verification (ConPTY, live Copilot, loopback 21 and Dev Tunnel 17 assertions, all pass) in `docs/closure/evidence/072-S-live-verification-evidence.md`; closure `READY`.
* Process notes: the lock helper is `.github/skills/file-lock/scripts/release_lock.ps1` (not `scripts/`). The ConPTY harness needs `STARTF_USESTDHANDLES`; the throwaway verifier is not committed.
* Next: closure PR, CI, Copilot gate, merge.
* Original: `docs/archive/memory/2026-10-29/072-s-ship-memory.md`.

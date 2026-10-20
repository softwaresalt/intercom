# 062-S (070-F) compacted memory

- Scope: MVP S7 process wiring and end-to-end verification (`internal/app`, `cmd/intercom`, TUI credential panel, headless e2e, opt-in live SDK test, operator runbook). Tasks 070.001-T..070.006-T done. PR #185 merged as a merge commit 6290151 (head reviewed 6930bde); dark mode (P-017) active, operator AFK, admin fallback not authorised or used.
- Review: seven personas plus Adversarial Review (2 of 3 reviewers); no P0/P1; one fix cycle (828dc6c) with Security, Concurrency and Correctness deltas. Copilot: three rounds with findings (live probe CLI path override; stale re-issue results and clipped credentials at small terminal sizes; startup credential delivery racing Ctrl+R; armed approval not disarmed by shortcut keys) all fixed, replied, resolved; fourth round clean.
- Decisions: composition root in `internal/app`; hub context never cancelled before step 6; exact one-call writepath allowance for the first `pathsafe.Root.Resolve` caller; TUI-only credentials with terminal-only stdout and discarded logs when the TUI owns the terminal; single disarm choke point; initial credentials at model construction.
- Deferred scope (stash): FAE4134F (lifecycle hardening), AD8472DD (internal/app refactors), BDC0ADDF (verification-surface and contract follow-ups). Left open and not absorbed: 1DCB6108, E9223340, 7EF4DA98 leftovers, 550437DB, 6738A8E5, 4AFE4DA4, 026CF750.
- Lessons: docs/compound/2026-10-20-process-wiring-credential-panel-and-approval-disarm.md.
- Next: closure PR through CI and Copilot loop. MVP DAG complete; live SDK run and real tunnel run are operator actions.
- Original: docs/memory/2026-10-20-062-s-ship-memory.md.

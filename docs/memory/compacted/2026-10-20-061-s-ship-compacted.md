# 061-S (069-F) compacted memory

- Scope: MVP S6 local Bubble Tea TUI in `internal/tui` (transcript viewport, permission modal, prompt input, feed via own buffered channel). Tasks 069.001-T..069.004-T done. PR #183 merged as a merge commit 8bbd57f (head reviewed 49b7452); dark mode (P-017) active, operator AFK, admin fallback not authorised.
- Review: seven personas plus single-reviewer Adversarial Review (degradation; no three-model dispatch); no P0/P1. One fix cycle (b41b760) with Correctness and Security deltas. b0ce34e pinned hub `Config.Now` to remove a Linux-only flake found under WSL. Copilot round 1: two real threads (modal not synchronized after dropped head request; approve allowed when modal not rendered at a small terminal), fixed in 49b7452; re-review clean.
- Decisions: two-key approve; approve gated on fully-viewed; hydrate authoritative; forced re-hydrate is close plus redial with a documented fail-closed window; one shared "modal rendered" predicate used by View and approvable; syncModal after every pending-set mutation. Dependencies bubbletea v1.3.10, lipgloss v1.1.0, bubbles v1.0.0; x/sys and x/text bumps raised go directive to 1.25.0 (toolchain pin go1.26.9 unchanged).
- Deferred scope (stash): 1DCB6108 (hub in-place rehydrate/redial retry), E9223340 (cmd wiring, transcript cap and render cache, invisible-char sanitizing, P3 hardening).
- Lessons: docs/compound/2026-10-20-tui-permission-modal-safety-and-linux-test-clock.md. Pin every injected clock in tests; run Linux binaries via WSL.
- Next: closure PR through CI and Copilot loop; then 062-S composition root.
- Original: docs/memory/2026-10-20-061-s-ship-memory.md (left in place; session open for closure PR).

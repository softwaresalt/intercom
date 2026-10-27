# 064-S (071-F) compacted memory

* Scope: verification-surface and contract follow-ups. Tasks 071.001-T through 071.006-T done: `already_resolved` vs `request_unknown` GoDoc (hub, wsclient); deterministic late TUI approve in `TestE2EFirstResponder`; `config.MaxApprovalTimeoutSeconds = 86400` upper bound; native Windows Toolhelp32 + PEB command-line process snapshot with a POSIX split and orphan regression test; hardened `readInsideWorkspace` with credential snapshot and lookalike regression matrix; config reference, live SDK runbook and operator runbook updates. PR #193 merged as a merge commit ee9eb55 (head d2b25b2; code last reviewed at 7ed1dcb).
* Review: Correctness, Go, Security, Concurrency, Maintainability, Schema-CLI-Docs and Architecture personas plus Adversarial Review (3 of 3 reviewers); no P0/P1; one fix cycle (7ed1dcb) re-reviewed READY. Copilot reviewed once with no findings (P-018 SATISFIED before and immediately before merge); CI all green.
* Decisions: followed the plan and task specs (`MaxApprovalTimeoutSeconds = 86400`, `docs/runbooks/live-sdk-tests.md`) over the invocation's `MaxApprovalSeconds`/31536000 and `docs/runbook.md` wording, because Ship may not edit planning fields and the one-year clamp already lives in `approvalTimeout`; in-scope review findings fixed under P-021 C1; cascade close ran on the Orchestrator-relayed closure-specific approval.
* Deferred scope (stash, P-021 C2): 987B9A82, 85139DD1, CCB9709E, 0111BCEC, D9292468, BB17B793 and EBA7AE1E.
* Closure: cascade close exited 0 with postcondition pass; evidence `docs/closure/evidence/064-S-071-F-close-evidence.json`.
* Next: closure PR through CI and the Copilot loop, then 065-S.
* Original: docs/archive/memory/2026-10-26-064-s-ship-memory.md.

# Ship session memory: shipment 050-S (feature 020-F), 2026-10-17

* Mode: DARK_MODE_ACTIVE (P-017), scope 050-S only; merge pre-authorized; no admin fallback.
* Branch `feat/050-s-mechanical-enforcement-of-the-stage-default-branch-contract` from main a40347d; pre_claim and post_claim topology gates passed; claim verified active.
* Tasks (JIT per-task harness under the repaired Ship contract, red then green): 020.001-T red 8a46672 / green fe4e9ee; 020.002-T 6a52918 / 4380c82; 020.003-T 6c76ffa / 9ce4c8a; 020.006-T 2a0e584 / 4dc3ea0; 020.004-T 0ada8be / 8bb4fe4; 020.005-T 23d626f / 02a620c. Review remediation: 16e6ad7 (red tests) / 07e8f7e (fixes).
* Deviation recorded: the mutation proof mutates the policies copy (detectP3 reads the P3 grant from workflow-policies.md), not _stage.agent.md as the plan text said; the plan allowed either with an exactly-one-occurrence assertion.
* Review: Correctness, Security, Scope reviewers: no P0/P1; P2 findings fixed (P6 historical, lazy continuation, P5 missing rows, P4 scope, P3 assertion). Deferred P3 items captured as stash 6333B903 (P-021 C2).
* Done-task archive moves withheld from the PR (plan C-1 convention) and land in the closure PR.
* Operator-owned follow-up (not performed): required code-owner review in branch protection / base-ref required workflow.
* Telemetry begin/record not exercised (delegated subagents).

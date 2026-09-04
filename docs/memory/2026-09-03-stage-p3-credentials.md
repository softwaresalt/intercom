---
title: "Stage session memory — intercom-go P3 credential resolution"
date: 2026-09-03
agent: stage
session_id: stage-p3-credentials-2026-09-03
phase: gate-blocked
stash_entries: ["037B1552"]
roadmap_parent: "4989A42D"
predecessor: "003-F / 003-S (P2), merge 40c7d289, closure 2a547913"
outcome: "deliberation accepted; plan FAILed the adversarial gate at attempt 3; circuit breaker opened; no harvest, no shipment"
---

# Stage Session Memory — intercom-go P3 Credential Resolution

## Entry point

Operator-targeted single feature-shaped stash entry `037B1552` (P3 credentials, residual scope of
the original config+errors slice), under roadmap `4989A42D`. Feature-shaped, so Step 1.5 contextual
grouping was correctly skipped. No `DEFERRED SCOPE EXPANSION` marker → no P-021 obligations.

## Environment state at start

* `ALL_TOOLS_OK` — backlogit CLI 1.10.1 (registry: `sizing`, `shipments`, `checkpoints`,
  `dependencies`, `semantic_links` all true). Manual/CLI mode; no MCP probe failures.
* `INDEX_SYNC_OK` — 62 artifacts.
* Zero checkpoints → ZERO-CANDIDATE NORMAL STARTUP, not a failure, no recovery performed.
* Zero shipments, no open PR, main at `2a54791`, exactly one worktree (P-016 satisfied).

## Oracle research — six stash corrections

Oracle `softwaresalt/agent-intercom` @ `41df772`, read-only, never modified. `.env.local` never
opened in either repository.

1. **F2 (largest)** — `authorized_user_ids` does **not** use the keychain chain. It is a separate,
   synchronous, env-only two-tier path with its own error message and CSV parsing. A single
   generic resolver over all four credentials would have been wrong.
2. **F3** — tiers 3–4 are hard-gated by `if mode != ServerMode::Mcp`; mode comes only from the
   `--mode` CLI flag. Since the port retires MCP as a server mode and is ACP-first, the surviving
   chain is the four-tier ACP chain, as a mode-free constant.
3. **F6** — the oracle collapses every keychain error into `Err(())`; a locked keychain is
   indistinguishable from an unset credential.
4. **F7** — `keyring = "3"` with no features may select the mock keystore, so the oracle's
   keychain tiers may be inert in production. Carried to `8ACF7110`.
5. **F8** — the oracle never tests its own precedence; its tests *assume* the keychain is empty and
   would invert on any provisioned machine. This is the empirical justification for the seam.
6. **F9** — the credential test corpus is 14, not 8; the extra 6 include the only redaction test.

Also found: `ipc_auth_token` is an additional unredacted secret (P13 scope), and
`SLACK_TEAM_ID_ACP` is real but undocumented in the oracle (verified across four operator surfaces).

## Decisions (deliberation, accepted)

P3-D1(a) static four-tier ACP chain · P3-D2(b) provider seam + real adapter + fake ·
P3-D3(a) no bespoke platform adapters · P3-D4(c) classify `Miss` vs `Unavailable`, both continue ·
P3-D5(b) total `fmt.Formatter`, not just the four named surfaces · P3-D6 tokens are `Secret`,
`TeamID` plain, ACL count-redacted · P3-D7(c) standalone package, caller composes ·
P3-D8 startup integration · P3-D9 fake-first hermetic tests.

Dependency verified empirically, not assumed: `zalando/go-keyring` v0.2.8 builds `CGO_ENABLED=0`
on all four CI targets, `govulncheck` clean, MIT/MIT/BSD-2-Clause, absolute `/usr/bin/security`
on darwin, distinguishes `ErrNotFound` from `ErrUnsupportedPlatform`.

## Gate outcome

Three adversarial rounds, five personas, `dispatch_mode: full` throughout:

* rev 1 → 5×FAIL, 20 P1
* rev 2 → Architecture PASS, Schema-CLI-Docs PASS; 6 P1
* rev 3 → 5×FAIL, 7 P1 — but **no P0 at any revision**, and the residual findings are
  plan-specification defects about test scaffolding rather than design or security defects.

Circuit breaker opened at attempt 3. Escalation route resolved to `gpt-5.6-sol`/`openai`/`high`
(not degraded — Stage's own route is `claude-opus-5`). Payload at
`docs/memory/2026-09-03-stage-p3-credentials-escalation.md`.

## Four empirically verified findings

Verified by running Go against the pinned `go1.26.5`, not by argument:

1. A value-held `Secret` leaks plaintext via `%w` on a non-error, `%p`, and unexported-field
   traversal; pointer boxing fixes all three, including `slog` `TextHandler`.
2. The same mechanism leaks `Resolved.authorizedUserIDs`; unexported-ness *causes* it.
3. Rev 1's dual-`errors.Is` criterion was unimplementable against `apperr`; an `Unwrap() []error`
   composite satisfies all three targets and preserves the message.
4. The keychain dependency meets every workspace floor (CGO, vulnerability, licence, platform).

## What was NOT done, and why

* **No harvest.** Step 5 requires a passing (or operator-accepted) review gate. Creating a
  feature/task hierarchy from a FAILed plan would be a P-005 violation.
* **No shipment.** Consequently no `shipment_id` and no predecessor link to `003-S`.
* **Stash `037B1552` NOT archived.** Residual scope remains — the entry was not consumed, because
  nothing was promoted to the backlog. It was updated in place with session progress and the
  corrections, so the next session does not repeat the oracle research.
* **Roadmap `4989A42D` untouched.** P4+ ordering preserved exactly.

## Resumption

Resume at **Step 4 (plan review gating)** with the 17-item remediation queue recorded in the
plan's *Gate Status* section. The deliberation is accepted and needs no rework; do not re-run
triage, oracle research, or deliberation.

## Preservation

Locally modified `.gitignore`, untracked `.claude/`, and `.backlogit/hooks_queue.jsonl` were left
unstaged and uncommitted throughout. Neither reference repository was modified. `.env.local` was
never read in either repository.

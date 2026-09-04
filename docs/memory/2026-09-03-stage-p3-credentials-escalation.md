---
title: "Stage escalation payload — intercom-go P3 credentials plan-review circuit breaker"
date: 2026-09-03
agent: stage
session_id: stage-p3-credentials-2026-09-03
threshold_kind: consecutive-plan-review-failures
threshold_count: 3
resolved_escalation_route: "gpt-5.6-sol / openai / high"
escalation_degraded: false
stash_entries: ["037B1552"]
roadmap_parent: "4989A42D"
---

# Escalation Payload — P3 Credential Resolution Plan Review

## Threshold

`plan-review-attempt` counter reached **3** with three consecutive `decision: FAIL` outcomes
across a five-persona adversarial gate. Circuit opened per P-013.6. Stage halted the review loop
and did **not** harvest.

## Route resolution

Read fresh from `.autoharness/config.yaml` at session start, per the Session-Start Dynamic Reload
contract (E8B5B3C5/H6/H7):

* `model_routing.stage.model_family` = `claude-opus-5` (Stage's own role route)
* `model_routing.stage.escalation` = `gpt-5.6-sol` / `openai` / `high`
* `model_routing.escalation` (legacy flat) = empty → no F02FD596 both-present ambiguity
* Same-route guard: `gpt-5.6-sol` ≠ `claude-opus-5` → **not `ESCALATION_DEGRADED`**

## Failure summary

| Attempt | Revision | Security | Go | Architecture | Scope | Schema-CLI-Docs | P1 count |
|---|---|---|---|---|---|---|---|
| 1 | rev 1 | FAIL | FAIL | FAIL | FAIL | FAIL | 20 |
| 2 | rev 2 | FAIL | FAIL | PASS | FAIL | PASS | 6 |
| 3 | rev 3 | FAIL | FAIL | FAIL | FAIL | FAIL | 7 |

`dispatch_mode: full` in every round; no persona was ever unavailable, so no P-012 degradation
applies. **No P0 was raised at any revision.**

## Diagnosis

Rounds 1→2 converged strongly on substance: every rev-1 P1 was a real design or security defect,
and three were empirically disproved/confirmed by running Go against the pinned `go1.26.5`
toolchain rather than argued. Round 3's residual P1s are almost entirely **plan-specification
defects about test scaffolding** — Go test-package placement, `go:build` tag reachability from the
gate command, a canary literal colliding with a scan allowlist. Each remediation round added
detail that became the next round's defect surface.

The plan is over-specified for its artifact class: it is litigating implementation-time decisions
at plan granularity. That is the failure mode the circuit breaker is designed to detect.

## Recommended disposition for the escalated reviewer

1. Judge whether the 17-item remediation queue in
   `docs/plans/2026-09-03-intercom-go-p3-credentials-plan.md` § *Gate Status* is better applied
   as a plan rev 4, or **absorbed into the implementation slice** as unit-level acceptance detail
   with the plan trimmed back to design contracts.
2. Adjudicate the standing Security dissent on **SEC-5** (keychain-first precedence poisoning),
   which rev 3 records as an explicitly accepted residual risk (Decision R15). Security accepted
   the acceptance at attempt 3; the item is closed unless the escalated reviewer disagrees.
3. Confirm the four empirically verified facts below are load-bearing and correctly applied.

## Empirically verified facts (do not re-litigate without re-running)

Run against `go1.26.5`, the toolchain pinned in `go.mod`:

1. A `Secret` holding a plain `string` **leaks the plaintext** via `fmt.Errorf("%w", s)` on a
   non-error, via `%p`, and via traversal of an unexported field of a containing struct — because
   `fmt`'s `badVerb` sets `erroring` before `handleMethods`, and `printValue` skips `Formatter`
   when `CanInterface()` is false. Boxing the value behind `*string` renders an address in all
   three cases. `slog`'s `TextHandler` inherits the same leak and the same fix.
2. The same applies to `Resolved.authorizedUserIDs []string`: unexported-ness **causes** the leak
   rather than preventing it. An opaque wrapper over `*[]string` fixes it.
3. `errors.Is(err, credentialsSentinel)` is **false** for an `*apperr.Error` built with
   `apperr.New`, and `apperr.Wrap` destroys the descriptive message — so rev 1's dual-`errors.Is`
   acceptance criterion was unimplementable. A composite with `Unwrap() []error` satisfies the
   sentinel match, `apperr.ErrConfig`, and `errors.As(*apperr.Error)` while preserving the message.
4. `github.com/zalando/go-keyring` v0.2.8 cross-compiles under `CGO_ENABLED=0` for
   `linux/amd64`, `windows/amd64`, `darwin/amd64`, `darwin/arm64`; `govulncheck` v1.7.0 reports
   zero affecting vulnerabilities; licences are MIT / MIT / BSD-2-Clause; macOS uses the absolute
   path `/usr/bin/security` (no `PATH`-hijack surface); and the library distinguishes
   `ErrNotFound` from `ErrUnsupportedPlatform`.

## Artifact references

* Deliberation: `docs/decisions/2026-09-03-intercom-go-p3-credentials-deliberation.md`
* Plan (rev 3, FAILed): `docs/plans/2026-09-03-intercom-go-p3-credentials-plan.md`
* Session memory: `docs/memory/2026-09-03-stage-p3-credentials.md`
* Backlog checkpoint: `.backlogit/checkpoints/checkpoint-20260904-023022.json`

## Resumption checkpoint

Resume at **Step 4 (plan review gating)**. The deliberation is accepted and requires no rework.
Do **not** re-run stash triage, oracle research, or deliberation. No backlog items and no shipment
were created, so there is nothing to reconcile or roll back on the backlog side.

## Escalation Outcome

The configured route (`gpt-5.6-sol` / `openai` / `high`) produced revisions 4–6 and ran three
full five-persona reviews. The final bounded gate returned:

```text
dispatch_mode: multi-agent
decision: FAIL
P0: 1
P1: 3
P2: 6
P3: 2
```

The P0 is an oracle-contract conflict: revision 6 specifies first-present ACL env semantics, while
the pinned oracle removes empty ACP values before shared fallback. The operator must choose exact
oracle parity or authorize a deliberated divergence. The three remaining P1s are C6/D1b compile
ordering, cancel-during-lookup orchestration, and incomplete total-fmt test enumeration.

No harvest, backlog hierarchy, shipment, dependency edge, or stash archival was performed. Stash
`037B1552` remains active and `4989A42D` remains unchanged for P4+.

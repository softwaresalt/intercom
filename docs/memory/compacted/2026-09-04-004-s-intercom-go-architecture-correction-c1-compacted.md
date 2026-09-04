---
title: "Compacted memory — 004-S intercom-go architecture correction C1"
date: 2026-09-04
shipment: 004-S
feature: 005-F
pr: 15
merge_commit_sha: 2588f2e91e4d572d8e408538cba9e894c6fb3a55
sources:
  - docs/archive/memory/2026-09-04-stage-intercom-go-architecture-correction.md
  - docs/archive/memory/2026-09-04-ship-004-S-pre-merge-checkpoint.md
---

# Compacted memory — 004-S / 005-F (architecture correction C1)

## Trigger and governing artifacts

Corrective Stage+Ship lifecycle after an explicit operator architecture
correction (2026-09-04) that retires Slack / custom-ACP broker /
headless-Copilot-CLI lifecycle and adopts the official
`github.com/github/copilot-sdk/go` (pinned v1.0.11,
`a550258d5c37bd662197536992a23d633bfe5804`, landing phase C2) as the agent
boundary. Local UX Bubble Tea; remote UX SPA; Dev Tunnels transport.

Governing: `docs/decisions/2026-09-04-intercom-go-architecture-correction-copilot-sdk-deliberation.md`,
`docs/design-docs/intercom-go-backend-architecture.md` (rev 2),
`docs/plans/2026-09-04-intercom-go-architecture-correction-plan.md` (hardened,
adversarial review PASS at attempt 3 of 3; six personas; 51 remediations
recorded in the plan's Post-Review Remediation Record).

## Stage phase (assembly)

Feature `005-F` + 4 tasks + 10 subtasks assembled into shipment `004-S`
(queued, 15 items, predecessor `003-S` shipped). 13 pre-existing
deferred-scope-expansion stash entries reconciled under P-021 C6 (late PR
identifiers recovered, all duplicate scans CLEAN). Stash `037B1552`
(Slack P3 credentials) archived with a RETIRED banner. New stash `8C2D578D`
captured for `internal/apperr`'s `KindSlack`/`KindIPC`/`KindACP` taxonomy
contamination (D6a, explicitly deferred, out of C1 scope).

## Ship phase (execution)

Implemented all 10 plan units (U-A1 Go 1.24 floor; U-B0 remove
channel-routing API; U-B1a/b delete retired schema+validation, add
`[copilot].cli_path` rule 5; U-B2 re-found test surface; U-D1 re-pin
defaults 17→12; U-D2 rewrite example; U-D3 rewrite config-reference.md;
U-E1a retired-architecture gate script; U-E1b wire into CI non-blocking).
Full gate suite green (build/vet/fmt/mod-tidy/test/-race/golangci-lint/
staticcheck/govulncheck/gate-self-test). Six-persona structured adversarial
review: 0 P0/P1; 2 P2 test-coverage gaps and 3 P3 nits fixed directly
(same-contract-surface completions, P-021 C1); 2 P3 script-robustness
advisories deferred as stash `B6203CCA` and `78D13775`.

PR #15 merged via merge-commit strategy at
`2588f2e91e4d572d8e408538cba9e894c6fb3a55`; CI 10/10 green (cross-compile
4/4, lint, test, security, ci gate, topology, detect-changes); P-018
copilot-review gate `NOT_APPLICABLE`; P-009 merge-commit strategy verified.

## Shipment closure

Shipment `004-S` closed via the P-015 verified fully-covered-root cascade
path (`005-F` is a root feature, fully covered by the manifest with no
unshipped siblings or extra descendants; the generic non-cascading
`backlogit move --status shipped` path is unavailable in this backlogit
build, confirming the cascade path was required, not merely permitted).
`archived_ids` = all 16 (shipment + feature + 4 tasks + 10 subtasks + the
shipment record double-counted feature); `returned_ids: []`.

## Carried-forward open operator decisions (unchanged from Stage)

| # | Decision | Blocks |
|---|---|---|
| H1 | Flip gate to required (needs CODEOWNERS per `BEDD2E70`) | — |
| H2 / Q1 | Copilot CLI minimum version / protocol-3 floor | C2 |
| H3 / Q3 | Dev Tunnel authentication model | C10 |
| H4 / Q6 | Sessions per workspace process | C9 |
| H5 | Permission authorization model | C5 |
| H6 | D4a session-pointer store semantics | C4 |

## Next step

Phase C2 (SDK proving spike) is next in the phased roadmap, depends on
`004-S`/C1 (shipped).

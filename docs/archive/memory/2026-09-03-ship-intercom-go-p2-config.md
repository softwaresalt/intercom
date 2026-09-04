---
title: "Ship session — intercom-go P2 configuration package (003-S)"
date: 2026-09-03
status: awaiting-merge-approval
---

# Ship session — intercom-go P2 configuration package (003-S)

- **Agent:** Ship (`claude-sonnet-5`, high; escalation route `gpt-5.6-sol`/openai/high, unused — no consecutive failures occurred)
- **Mode:** standard sequential (not dark factory)
- **Branch:** `feat/003-s-intercom-go-p2-configuration-package` from `main` @ `10745b1`
- **PR:** https://github.com/softwaresalt/intercom/pull/11
- **Outcome so far:** implementation, review, and CI complete; awaiting explicit operator merge approval (P-014)

## Scope delivered

Shipment `003-S` — covering feature `003-F`, 5 tasks, 13 subtasks, 19 manifest
items. All complete. Ports the `agent-intercom` (Rust) `config.toml` schema
into `internal/config`: schema types, 17 non-zero defaults, tolerant TOML
decode (new pinned dependency `github.com/BurntSushi/toml v1.6.0`), one
unified 10-rule fail-fast validation pass, workspace/channel resolvers,
operator-facing example/reference docs, and a one-line `--config` flag
default alignment in `cmd/intercom/main.go`.

## Step trace

| Step | Result |
|---|---|
| Pre-flight | topology gate `pre_claim`/`post_claim` PASS; `002-S` predecessor closure confirmed READY |
| Claim | shipment `003-S` claimed, `CLAIM_VERIFY_OK` |
| Branch | `feat/003-s-intercom-go-p2-configuration-package` created from clean `main` @ `10745b1` |
| Implementation | 13 units (A1→E3) implemented strictly in dependency order, one commit per unit, test-first per unit's acceptance criteria; 15 implementation commits total |
| Quality gates (initial) | `go build/vet` clean, `gofmt`/`goimports` clean, `go test -race ./...` all green, `golangci-lint` 0 issues, `staticcheck` clean, `govulncheck` 0 vulnerabilities in this code, `go mod tidy` no diff, 4/4 cross-compile targets (`CGO_ENABLED=0`) |
| Review gate | 5 parallel structured reviews (Go, Correctness, Security, Schema-CLI-Docs Coupling, Architecture Strategist), report-only mode: converged on 5×P1 findings, all remediated (commit `b13d33d`) plus a follow-up test-coverage commit (`6b90c62`); re-review verdict **READY**, 0 residual P0/P1 |
| Quality gates (post-remediation) | Re-ran full suite — all clean |
| PR | Created PR #11 with `## Local Review Readiness` block (Reviewed HEAD `8978bfa2c4a9ab2799b99bdf5ce2b24aa3a5d093`, outcome `READY`) |
| CI | 11/11 checks green (`ci gate`, `detect code changes`, `pipeline-topology (ambient)`, `test`, `lint`, `security`, `load cross-compile targets`, 4/4 `cross-compile` legs) |
| P-018 gate | `NOT_APPLICABLE` — Copilot review not engaged, enforcement `auto` |
| P-009 gate | Repo allows merge-commit strategy (`allow_merge_commit: true`); merge will be executed with explicit `--merge` |
| P-014 gate | **AWAITING explicit operator approval** — standard mode, not dark mode; not merged |

## Operator-approved checkpoints (pre-resolved per task brief, used as-is)

1. `github.com/BurntSushi/toml v1.6.0` dependency — approved, pinned, govulncheck clean, zero transitive additions.
2. Divergences V7/V8 (stricter-than-oracle containment: workspace/database path checks) — approved, documented in `docs/config-reference.md`.

## Review remediations (commit `b13d33d`, follow-up `6b90c62`)

- `findCaseFoldCollision` no longer false-positives on case-variant entries inside map-typed schema fields (`Commands`, `Slack.MarkdownUploadExtensions`).
- `filterLeafKeys` rewritten O(n²) → O(n log n) (CPU-exhaustion hardening on a hostile ~1 MiB config).
- `sanitizeKeyPath` truncates on a UTF-8 rune boundary (no more invalid-UTF-8 risk at the 128-byte cap).
- `Load` reads via a capped `io.LimitReader` instead of stat-then-read (TOCTOU closed).
- README/`config.toml.example`/`docs/config-reference.md` wording corrected for accuracy (no present-tense overclaim; correct CWD/PATH-timing statements).
- `workspace.go` resolvers document the `Validate()`-passed precondition for their containment guarantees.
- `validate.go` rule 2b/9 messages use `apperr.Error.Message()` instead of `.Error()` to remove a redundant nested "invalid:" phrase.

## Deferred out of scope (P-021)

- Stash `7028FBE7` (threadless capture, low priority, requires deliberation): rule 10's `database.path` `..`-segment check duplicates logic already present in `internal/pathsafe` with more rigor. Sharing it would require exporting new `internal/pathsafe` surface, which this shipment's plan scope guard does not authorize.

## Dirty-state boundaries preserved

`.gitignore` (local modification), `.claude/`, and `.backlogit/hooks_queue.jsonl`
were never staged or committed throughout the session — confirmed clean at
every commit checkpoint. Neither `C:\Source\GitHub\intercom` (oracle,
read-only) nor `references/herdr` were touched.

## Next action

Awaiting operator's explicit merge approval for PR #11 (branch
`feat/003-s-intercom-go-p2-configuration-package`, HEAD
`8978bfa2c4a9ab2799b99bdf5ce2b24aa3a5d093`). On approval: re-run the P-018/P-009
last-mile gates, merge with `gh pr merge 11 --merge`, then proceed to Step 6
post-merge closure (shipment archival, `docs/closure/003-S-003-F-post-merge-closure.md`,
P-020 compact-context, closure PR).

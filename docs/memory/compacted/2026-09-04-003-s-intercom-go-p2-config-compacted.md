---
title: "Compacted memory — intercom-go P2: configuration package (003-S)"
date: 2026-09-04
shipment: 003-S
feature: 003-F
status: shipped
pr_main: 11
pr_closure: TBD
merge_commit_sha: 40c7d289a60548e689ce714dd705411d9002a214
compacted_from:
  - docs/archive/memory/2026-09-03-stage-intercom-go-p2-config.md
  - docs/archive/memory/2026-09-03-ship-intercom-go-p2-config.md
---

# Compacted Memory — intercom-go P2 (003-S)

## Outcome

Shipment `003-S` shipped. PR #11 merged (merge commit
`40c7d289a60548e689ce714dd705411d9002a214`, merge-commit strategy, P-009).
Covering feature `003-F`, 5 tasks, 13 subtasks — all `done`, archived via
the P-015 verified fully-covered-root cascade close
(`backlogit shipment ship 003-S`). Closure artifact:
`docs/closure/003-S-003-F-post-merge-closure.md`.

## Key decisions (with rationale)

* Ports the `agent-intercom` (Rust oracle, `softwaresalt/agent-intercom` @
  `41df772`) `config.toml` schema into `internal/config`: schema types, 17
  non-zero defaults, tolerant TOML decode over a pre-populated `Default()`
  struct, one unconditional 10-rule fail-fast validation pass,
  workspace-to-channel routing, and operator-facing example/reference docs.
* New pinned dependency `github.com/BurntSushi/toml v1.6.0` — operator
  pre-approved, `govulncheck` clean, zero transitive additions.
* Divergences V7/V8 (stricter-than-oracle containment on workspace/database
  path checks) — operator pre-approved, documented in
  `docs/config-reference.md`.
* 13 implementation units (A1→E3) built strictly in dependency order, one
  commit per unit, test-first per unit's acceptance criteria (15
  implementation commits total).
* One-line `--config` flag default alignment with `config.DefaultConfigPath`
  in `cmd/intercom/main.go` (unit E3).

## Failed approaches

None — implementation proceeded without rework beyond the review-remediation
cycle noted below.

## Validation evidence

`go build`/`go vet` clean; `gofmt`/`goimports` clean; `go test -race ./...`
all green; `golangci-lint` 0 issues; `staticcheck` clean; `govulncheck` 0
vulnerabilities in this code; `go mod tidy` no diff; 4/4 cross-compile
targets (`CGO_ENABLED=0`) green.

## Review gate

Five parallel structured reviewers (Go, Correctness, Security,
Schema-CLI-Docs Coupling, Architecture Strategist), report-only: converged
on 5×P1 findings, all remediated (commit `b13d33d`) plus a follow-up
test-coverage commit (`6b90c62`); re-review verdict **READY**, 0 residual
P0/P1. Remediations: map-collision false-positive fix in
`findCaseFoldCollision`, `filterLeafKeys` O(n²)→O(n log n) hardening,
UTF-8-safe truncation in `sanitizeKeyPath`, TOCTOU closure in `Load` via
capped `io.LimitReader`, doc-wording corrections, resolver precondition
documentation, and error-message redundancy cleanup.

## Follow-ups (stash-captured, non-blocking, P-021 C2)

`7028FBE7` (low, threadless capture, requires deliberation): rule 10's
`database.path` `..`-segment check duplicates logic already present in
`internal/pathsafe` with more rigor; sharing it would require exporting new
`internal/pathsafe` surface, not authorized by this shipment's plan scope
guard.

## PR / merge / closure

* PR #11 merged with `--merge --delete-branch` (merge-commit strategy,
  P-009; verified 2-parent merge commit). Reviewed HEAD
  `8978bfa2c4a9ab2799b99bdf5ce2b24aa3a5d093`; HEAD advanced by one
  docs-only memory-checkpoint commit (`a94e5f2`) before merge — revalidated
  via `go build`/`go vet`/`go test ./...` (all green) since no source
  changed; local review remained fully valid. 11/11 CI checks green, P-018
  `NOT_APPLICABLE`, 0 open review threads/comments, topology `lifecycle`
  gate passed pre-claim/pre-merge.
* Shipment closed via P-015 verified fully-covered-root cascade:
  `returned_ids: []`; `archived_ids` (20: shipment + feature + 5 tasks + 13
  subtasks) exactly matches computed `allowed_ids`/`required_ids`; every
  `parent_id` preserved; shipment record `status: archived`,
  `archived_status: shipped`,
  `commit: 40c7d289a60548e689ce714dd705411d9002a214`.
* Backlog archival committed on
  `post-merge/003-s-intercom-go-p2-configuration-package`, not `main`
  (P-020/branch-per-release-unit).
* Source artifact cleanup: `003-F` carries no `source_stash_id` /
  `source_deliberation_id` custom fields — none applicable.

## Open items surfaced during staging (still relevant post-ship)

* Successor P3 (credential resolution) preserved in stash `037B1552`, per
  the Stage deliberation addendum
  (`docs/decisions/2026-09-03-intercom-go-p2-config-deliberation-addendum.md`).
* Deferred stash `7028FBE7` (see Follow-ups above) awaits Stage
  deliberation on whether/how to extract a shared traversal-segment helper
  from `internal/pathsafe`.

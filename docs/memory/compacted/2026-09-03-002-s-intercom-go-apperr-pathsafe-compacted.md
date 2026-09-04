---
title: "Compacted memory — intercom-go P1: error taxonomy and workspace path containment (002-S)"
date: 2026-09-03
shipment: 002-S
feature: 002-F
status: shipped
pr_main: 8
pr_closure: TBD
merge_commit_sha: 9efb156839e679fd7c29484fb99acd35fb9acd8b
compacted_from:
  - docs/archive/memory/2026-09-03-stage-intercom-go-apperr-pathsafe.md
  - docs/archive/memory/2026-09-03-ship-intercom-go-apperr-pathsafe.md
---

# Compacted Memory — intercom-go P1 (002-S)

## Outcome

Shipment `002-S` shipped. PR #8 merged (merge commit
`9efb156839e679fd7c29484fb99acd35fb9acd8b`). Covering feature `002-F`, 3 tasks,
8 subtasks — all `done`, archived via the P-015 verified fully-covered-root
cascade close (`backlogit shipment ship 002-S`). Closure artifact:
`docs/closure/002-S-002-F-post-merge-closure.md`.

## Key decisions (with rationale)

* **Oracle correction (P0, Stage).** `references/herdr` is **not** the
  behavioral oracle — it is `ogulcancelik/herdr` (Rust terminal multiplexer,
  0 hits for oracle-specific symbols). The real oracle is
  `softwaresalt/agent-intercom` at `C:\Source\GitHub\intercom` @ `41df772`.
  `references/herdr` left untouched throughout.
* **MCP-retirement blocker closed (Stage D1)**: retire MCP as a server mode;
  retain the MCP HTTP tool endpoint (phase P11) and the `protocol_mode`
  column/wire tool names for HITL clearance flows.
* **Group harvested as one bounded unit**: only the dependency-root phase P1
  (`002-F`: `internal/apperr` + `internal/pathsafe`) was harvested from the
  combined `4989A42D`/`037B1552` deliberation; later phases deferred with
  explicit successor ordering.
* **Plan review PASS at attempt 3** (1×P0 + 12×P1 at attempt 1, 2 new P1s at
  attempt 2's re-verification, all closed). Three load-bearing Go path
  claims verified empirically rather than assumed.
* **Two convergent in-implementation findings closed same-contract-surface**:
  (1) Windows rooted-vs-absolute gap (`filepath.IsAbs` under-rejects
  `/etc/passwd` on Windows) — closed via `isRooted()`. (2) P1,
  cross-reviewer-convergent: symlinked-intermediate-directory bypass in
  `checkSymlinkEscape` (only stat'd the full path, missing a symlinked
  ancestor combined with a non-existent leaf) — closed by walking up to the
  nearest existing ancestor and re-asserting containment. Both locked in by
  new regression tests.

## Failed approaches

None at the implementation stage beyond the plan-review iteration cycle
already noted above (rev 1 → rev 2 → rev 3 convergence).

## Validation evidence

`gofmt`/`go vet`/`golangci-lint`/`staticcheck` clean; `govulncheck` 0 new
findings; `go test -race ./...` full pass (symlink-privilege-gated suite
exercised, not skipped); `CGO_ENABLED=0 go build ./...` succeeds; zero new
dependencies (stdlib only).

## Review gate

Five parallel structured reviewers (Go, Correctness, Security, Architecture,
Constitution), report-only: converged on 1 P1 + 1 P2, both remediated same
session (commit `f3a7fda`); full suite re-confirmed green.

## Follow-ups (stash-captured, non-blocking, P-021 C2)

`4A91CA81` (low, `NewRoot` cause-dropping — plan-mandated mechanism,
deliberation required), `7774C9CA` (low, Go-idiom P3 advisories),
`3D61B5A8` (low, filesystem-root edge case + directory-type check +
GO-14 regression gap), `9A4C8749` (medium, `Newf` missing from
`.golangci.yml` printf allowlist — deliberation required), `2130906D`
(medium, plan's Constitution Check table principle-number mismatch —
Stage-owned artifact correction, deliberation required), `8ACF7110` (medium,
machine-checkable oracle-drift gate — deferred from plan review, depends on
`002.003.001-ST`).

## PR / merge / closure

* PR #8 merged with `--merge` (merge-commit strategy, P-009). Reviewed HEAD
  `093945b66b1bf8c4e07027110db36ec1d91ef553` — tree-identical to `a95e7c5`
  (last content commit before a memory-only docs commit). 11/11 CI checks
  green, P-018 `NOT_APPLICABLE`, 0 open review threads/comments.
* Shipment closed via P-015 verified fully-covered-root cascade:
  `returned_ids: []`; `archived_ids` (13: shipment + feature + 3 tasks + 8
  subtasks) exactly matches computed `allowed_ids`/`required_ids`; every
  `parent_id` preserved; shipment record `status: archived`,
  `archived_status: shipped`, `commit: 9efb156839e679fd7c29484fb99acd35fb9acd8b`.
* Backlog archival committed on
  `post-merge/002-s-intercom-go-p1-error-taxonomy-and-workspace-path-containment`,
  not `main` (P-020/branch-per-release-unit).
* Source artifact cleanup: `002-F` carries no `source_stash_id` /
  `source_deliberation_id` custom fields — none applicable.

## Open items surfaced during staging (still relevant post-ship)

* Architecture fork unresolved (multiplexer architecture doc, stash
  `A92E3FA0`) — pending operator decision.
* Torn `references` submodule blocks the Rust oracle mount for future
  parity work involving stash `037B1552`/`4989A42D` successors.
* Sizing degradation noted: this backend defines `size` on `task` only and
  `complexity` on no type — recorded as labeled prose on feature/subtasks.

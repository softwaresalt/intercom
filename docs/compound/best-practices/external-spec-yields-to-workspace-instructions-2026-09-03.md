---
title: "External specs never override stricter workspace instruction floors"
description: "A plan built directly from an external port brief violated the workspace's Go-version and CI action-pinning instructions; the stricter workspace standard always governs."
source: "docs/compound/best-practices/external-spec-yields-to-workspace-instructions-2026-09-03.md"
doc_type: "learning"
problem_type: "planning-process"
category: "best-practices"
component: "plan-review / instruction compliance"
root_cause: "Plan rev 1 anchored on an external Rust-port reference brief (go 1.21+, unpinned setup-go version tags) without cross-checking .github/instructions/ for stricter workspace-local floors."
resolution_type: "design_change"
severity: "high"
message: "plan-review P0: Go version floor / CI action pinning below workspace instruction requirement"
file_path: ".github/instructions/technology-go.instructions.md"
citations:
  - "docs/plans/2026-09-03-intercom-go-foundation-plan.md"
  - "docs/decisions/2026-09-03-intercom-go-architecture-reconciliation-deliberation.md"
  - "docs/closure/2026-09-03-intercom-go-foundation-closure.md"
tags:
  - "planning"
  - "instructions"
  - "ci-security"
  - "go"
---

## Problem

An implementation plan for `001-F` (intercom-go foundation) was built to satisfy
an external reference document (`docs/decisions/2026-07-06-go-port-reference-brief.md`,
a Rust-to-Go port brief) that specified `go 1.21+` and did not require SHA-pinned
GitHub Actions. Plan review (attempt 1) rejected the plan with 2 P0 findings:

* `.github/instructions/technology-go.instructions.md` requires "Target Go 1.22
  or later" — the plan set `go 1.21`.
* `.github/instructions/ci-security.instructions.md` MUST-requires full-SHA
  action pinning and explicitly forbids version tags — the plan pinned
  `actions/setup-go` by version tag.

## Root Cause

The plan author treated the external port brief as the authoritative source of
truth for technology floors, without cross-checking it against this
workspace's own `.github/instructions/` files. The external document and the
workspace instructions disagreed, and the external document was followed.

## Resolution

Plan rev 2 corrected both floors to the workspace's stricter values (`go 1.22`
language floor, later further hardened by a `go1.26.5` toolchain pin during CI
remediation for GO-2025-3750; full 40-hex-SHA action pinning throughout
`.github/workflows/ci.yml`). Rev 2 passed plan review with 0 P0.

**General rule**: when planning or implementing against an external spec,
reference document, or brief, always cross-check technology floors, security
requirements, and process constraints against this workspace's
`.github/instructions/` files. The workspace's own instructions are the
stricter, authoritative local standard and must never be silently
superseded by an external document's older or looser requirement — even when
the external document is explicitly the task's source-of-truth for domain
behavior.

## Prevention

Before finalizing a plan, an implementation plan, or a deliberation that
draws on an external spec/brief, explicitly enumerate the relevant
`.github/instructions/*.instructions.md` files for the technology stack and
confirm the plan's stated floors/constraints meet or exceed each one.
`plan-review`'s Learnings Researcher / Architecture Strategist personas are a
good backstop for this, but relying on review to catch it costs a full
plan-rewrite cycle — check it during initial plan authoring instead.

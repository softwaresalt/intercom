---
title: "go test -race fails closed (not silently disabled) when CGO_ENABLED=0"
description: "cmd/go pre-flight-checks -race and aborts with exit 2 if CGO is unavailable/disabled; it never silently skips the race detector."
source: "docs/compound/build-errors/go-race-requires-cgo-fails-closed-2026-09-03.md"
doc_type: "learning"
problem_type: "false-premise-design"
category: "build-errors"
component: "CI / go test invocation"
root_cause: "Plan rev 1 assumed `go test -race` with CGO_ENABLED=0 would silently disable the race detector, and added a sentinel-package guard to defend against that assumed silent-disable failure mode."
resolution_type: "design_change"
severity: "medium"
message: "go: -race requires cgo"
file_path: ".github/workflows/ci.yml"
citations:
  - "docs/plans/2026-09-03-intercom-go-foundation-plan.md"
  - "docs/closure/2026-09-03-intercom-go-foundation-closure.md"
tags:
  - "go"
  - "testing"
  - "ci"
  - "cgo"
---

## Problem

Plan rev 1 included a "sentinel package" guard intended to catch the case
where `go test -race` would silently disable the race detector under
`CGO_ENABLED=0` (the project's CGO-free build policy), producing a false
sense of race-safety in CI.

## Root Cause

The premise was false. `cmd/go`'s own pre-flight check refuses to run `-race`
at all when cgo is unavailable/disabled: it aborts immediately with
`go: -race requires cgo` and a non-zero (exit 2) status. There is no code
path where `-race` silently no-ops — the failure mode is fail-**closed**
(the build/test step itself fails loudly), not fail-open (silently skipping
detection).

The Go Reviewer persona identified and refuted this during plan review,
citing `cmd/go`'s actual pre-flight behavior rather than the plan's assumed
behavior.

## Resolution

The sentinel-package guard was removed from the plan as over-engineering
against a non-existent failure mode. CI simply runs
`go test -race -mod=readonly ./...` with `CGO_ENABLED` left at its Go
toolchain default (cgo-capable) for the test job specifically, distinct from
the CGO-free **build** requirement enforced separately by
`scripts/build.ps1` and its cross-compile matrix job. `go vet`/`go build` and
the actual shipped binaries remain `CGO_ENABLED=0`; only the `-race` test
invocation needs cgo available, and `cmd/go` itself guarantees that
requirement is enforced (fail-closed) rather than needing bespoke guarding.

## Prevention

Before adding defensive scaffolding (sentinel packages, guard tests, extra CI
steps) against an assumed tool failure mode, verify the tool's actual
documented/source behavior first (here: `cmd/go`'s own `-race`/cgo
precondition check). A guard built on a false premise is wasted
implementation surface and can itself introduce review findings (P1/P2) for
no safety benefit.

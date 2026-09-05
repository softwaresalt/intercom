---
title: "Go toolchain-pin maintenance expectation"
date: 2026-09-04
status: accepted
agent: Stage
resolves_stash: 90EE7758
---

# Go toolchain-pin maintenance expectation

## Context

`go.mod` currently declares two independent version floors:

```gomod
go 1.24
toolchain go1.26.5
```

These exist for **different reasons**, and conflating them is the failure mode
this note prevents.

* **`go 1.24` — the language/API floor.** Raised in shipment `004-S` (unit U-A1)
  to satisfy the declared floor of `github.com/github/copilot-sdk/go` ahead of
  its phase-C2 adoption. It is driven by *dependency requirements*.
* **`toolchain go1.26.5` — the build-toolchain pin.** Raised mid-shipment to
  remediate **GO-2025-3750** (inconsistent handling of `O_CREATE|O_EXCL` on Unix
  vs. Windows in `os`/`syscall`). The 1.22.x line was never patched for this
  CVE, so `govulncheck` cannot pass on any 1.22.x toolchain. It is driven by
  *stdlib vulnerability remediation*.

Until now this rationale existed only as an inline `go.mod` comment. Stash entry
`90EE7758` flagged that the ongoing *maintenance expectation* was captured
nowhere.

## Decision

**The toolchain pin is a maintenance obligation, not a one-time fix.**

1. `toolchain` MUST remain greater than or equal to the `go` language floor.
   It is expected to sit strictly above it whenever an unpatched stdlib CVE
   affects the floor release line.
2. When `govulncheck` reports a new stdlib advisory, the **first** remedy is to
   raise `toolchain`, not the `go` language floor. Raising the language floor is
   a breaking consumer-facing change and is reserved for genuine dependency or
   language-feature requirements.
3. Raising the `go` language floor requires its own justification recorded in a
   plan or decision artifact (as `004-S` did for the Copilot SDK). It is never a
   side effect of CVE remediation.
4. Any change to either line MUST update the inline `go.mod` comment so the two
   rationales stay separable at the point of use.

## Rationale

The two floors drift for unrelated reasons on unrelated schedules. A future
maintainer seeing `toolchain` above `go` could plausibly "tidy" them into
agreement — either by dropping the toolchain pin (reintroducing GO-2025-3750
and breaking `govulncheck`) or by raising the language floor (an unnecessary
breaking change). Recording the asymmetry as deliberate prevents both.

This is a documentation/process decision with **no code change**: `go.mod` is
already in the desired state. No CI enforcement is added, because
`govulncheck` already runs in the quality gates and will surface the next
advisory on its own — adding a second mechanism to assert what `govulncheck`
already asserts would be redundant (rule 5, simplicity).

## Scope

Documentation only. No source, test, or CI change accompanies this note.

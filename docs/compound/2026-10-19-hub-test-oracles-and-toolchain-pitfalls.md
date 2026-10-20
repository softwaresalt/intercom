---
title: "Hub test oracles must honour the documented TUI overflow-recovery contract; lint and gate toolchain pitfalls"
date: 2026-10-19
tags: ["go", "hub", "testing", "concurrency", "golangci-lint", "toolchain", "retired-arch-gate"]
shipment: 058-S
severity: medium
---

# Hub test oracles and toolchain pitfalls (058-S)

## 1. Oracles for a drop-oldest consumer must model the recovery contract

The hub drains a stale TUI client's queue (drop-oldest) and then sends a `session.hydrate`. A consumer reading the raw
`Frames()` channel may therefore see a seq gap before the hydrate. A strict "contiguous seq" oracle on that stream fails on a
valid interleaving, and the failure depends on scheduling. Model an awaiting-hydrate state in the oracle (discard live frames
until a hydrate, then require contiguity again) and keep strict contiguity for remote streams, which never recover in place.
Also: when a helper such as `connect()` already consumed the initial hydrate, every hydrate observed later is a forced one, so
do not subtract it again. Copilot review caught both mistakes in a test that engineers had called "non-vacuous".

## 2. Worker deadlines must be reported, not swallowed

A worker loop that returns silently on deadline expiry lets a stalled stream pass the property test with an empty result.
Separate the stop-triggered exit from deadline expiry and record the latter as a liveness failure.

## 3. Mutation checks must require a safety violation

A mutation check that accepts "worker timed out" as detection proves liveness, not the property. Require GAP or DUPLICATE.

## 4. Toolchain and gate gotchas

* The system Go (1.27) breaks the installed `golangci-lint` (built with go1.26.5); run tests and lint with
  `GOTOOLCHAIN=go1.26.9`, which matches the `go.mod` pin.
* The retired-architecture gate flags the token `slack` in identifiers (for example a constant named `entrySlack`); use a
  neutral name such as `entryOverhead`.
* Files created through some editor tools may carry CRLF; `.gitattributes` and `gofmt` want LF, so run `gofmt -w` on new files.

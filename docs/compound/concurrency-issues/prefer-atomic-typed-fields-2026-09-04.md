---
title: "Prefer atomic.IntN struct fields over raw intN + sync/atomic functions"
description: "A raw int64 struct field manipulated via atomic.AddInt64(&field, ...) risks 32-bit alignment bugs; use atomic.Int64/atomic.Int32 (Go >= 1.19) instead, caught by GitHub Copilot PR review"
source: "docs/compound/concurrency-issues/prefer-atomic-typed-fields-2026-09-04.md"
doc_type: "learning"
problem_type: "concurrency-portability-risk"
category: "concurrency-issues"
component: "internal/copilotprobe (PermissionHarness)"
root_cause: "sync/atomic's function-based API (atomic.AddInt64, atomic.LoadInt64, ...) requires the target int64/uint64 value to be 8-byte aligned, which the Go runtime only guarantees for the first word of an allocated struct/slice/array on 32-bit platforms; a plain int64 field elsewhere in a struct can be misaligned there even though it is never an issue on amd64/arm64"
resolution_type: "code_fix"
severity: "low"
message: "sync/atomic operations on 64-bit values require 64-bit alignment; using an int64 field inside a struct can be misaligned on 32-bit architectures"
file_path: "internal/copilotprobe/permission.go"
citations:
  - "shipment 005-S, PR #17, Copilot review comment (discussion_r3939243615)"
tags:
  - "go"
  - "sync-atomic"
  - "code-review"
  - "portability"
---

## Problem

`PermissionHarness` used a plain `nextN int64` struct field, incremented via
`atomic.AddInt64(&h.nextN, 1)` to atomically reserve an index without
holding the harness's mutex. GitHub Copilot's automated PR review flagged
this as an alignment risk.

## Root Cause

`sync/atomic`'s legacy function-based API (`atomic.AddInt64`,
`atomic.LoadInt64`, etc.) requires its target address to be 8-byte aligned.
The Go compiler/runtime guarantees this for the *first* word of a struct,
slice, or allocated value, but a field placed anywhere else in a struct can
end up on a 4-byte boundary on 32-bit architectures (386, arm), where the
struct's other fields shift the offset. amd64/arm64 (the overwhelmingly
common deployment target, and the only one exercised in this repo's CI
matrix) are unaffected, which is why this class of bug is easy to miss
entirely in day-to-day development and testing.

## Resolution

Replace the raw field + function-API pair with the typed atomic wrapper
introduced in Go 1.19: `atomic.Int64` (or `atomic.Int32`, `atomic.Bool`,
`atomic.Pointer[T]`, etc. as appropriate). The typed wrappers manage their
own internal alignment guarantees regardless of struct field placement, and
their method-based API (`h.nextN.Add(1)`, `h.nextN.Load()`) is also more
idiomatic and harder to misuse (no risk of forgetting `&` or mixing
`atomic.AddInt64` with a non-atomic read elsewhere).

```go
// Before
type PermissionHarness struct {
    mu     sync.Mutex
    nextN  int64 // atomic
}
n := int(atomic.AddInt64(&h.nextN, 1) - 1)

// After
type PermissionHarness struct {
    mu    sync.Mutex
    nextN atomic.Int64
}
n := int(h.nextN.Add(1) - 1)
```

## Prevention

Default to `atomic.Int32`/`atomic.Int64`/`atomic.Bool`/`atomic.Pointer[T]`
for any new atomically-manipulated struct field in Go code targeting >= 1.19
(this repo's floor is 1.24). Only reach for the raw `atomic.AddInt64`-style
function API when the target is a local variable or the first field of a
struct with no 32-bit deployment target ambiguity, and even then prefer the
typed wrapper for consistency. This class of bug is exactly the kind
automated PR review (GitHub Copilot, or an equivalent Go linter such as
`fieldalignment`/`structcheck`) is well-suited to catch — treat a flagged
instance as a quick, safe, in-scope fix rather than deferring it.

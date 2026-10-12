---
title: "Decision note: hermetic and containable start-script integration tests (Batch 3)"
description: "Short Stage decision note (no full deliberation) for stash C5E50EFD: scratch-copy start.ps1, sidecar-free PATH, cmd.WaitDelay; reconciles stash 26EDD633"
date: 2026-10-12
status: accepted
stash: "C5E50EFD (consumed), 26EDD633 (absorbed)"
tags:
  - "tests"
  - "start-script"
  - "hermetic"
---

# Decision note: hermetic and containable start-script integration tests

Stage routing: stash `C5E50EFD` carries the `DEFERRED SCOPE EXPANSION` marker (P-021 C6 forces the
`deliberate` route) **and** the entry itself records `requires deliberation: no`, an accepted
Orchestrator root-cause analysis, and a verified fix shape. The operator directive for this batch
confirms skipping the full `deliberate` skill. This note is the durable decision artifact that
satisfies the C6 "no un-deliberated deferred entry reaches planning" rule. It records only the
genuine design choices (D-HS-1..D-HS-6); there is no option-space beyond them.

## Triage record (Step 1, P-021 C5/C6)

* **Shape:** task-shaped, solo group (no peer task shares the `tests/integration` start-script
  surface). Implicit covering feature: "Hermetic and containable start-script integration tests".
* **(A) Unconditional duplicate scan:** performed over the active stash (85 entries) by text
  search for `WaitDelay|hermetic|start.ps1|start-script|start_script|Get-EnabledSidecars`; only
  `26EDD633` and `C5E50EFD` match. `96B308D1` (different test) is related only by name in the
  entry's text.
  Result: **no duplicate DEFERRED SCOPE EXPANSION entry**; `26EDD633` is a pre-marker follow-up
  (not a duplicate capture of the same expansion) and is handled as a related-entry reconciliation
  under D-HS-6.
* **(B) Late-identifier reconciliation:** the entry's `review-thread ID` and `task ID` are `N/A`.
  Searched the Ship-owned residual-risk records that cite `C5E50EFD`
  (`docs/closure/042-S-052-F-post-merge-closure.md`,
  `docs/closure/evidence/042-S-verification-evidence.md`, the 042-S resume memory): they carry
  the same refs (052-F, 042-S, PR #123/#124) and **no late review-thread ID or task ID was
  found**; the recorded `N/A` stands as a truthful terminal record (non-blocking).
* **Source refs carried:** feature 052-F, shipment 042-S, PR #123 (merge
  `591f37e2ba877d81b1304d31905369e79096657f`), PR #124 (closure), operator disposition
  2026-10-11T21:47 -07:00.

## Decisions

### D-HS-1: Hermetic scratch copy of start.ps1 (byte copy, not symlink)

The two launcher tests run a **byte-for-byte copy** of the repository `start.ps1` placed in a
`t.TempDir()` directory, so `$PSScriptRoot`, `.env.local`, `.copilot`, `.engram`, and
`.graphtor\bin\graphtor-docs.exe` all resolve inside the scratch directory and the tests can no
longer read or write real workspace state (root cause iii) or pick up the repo-local graphtor
binary. A symlink is rejected: `$PSScriptRoot` would resolve through the link to the real
repository on some hosts, and Windows symlink creation needs privilege (the suite already has
`symlink_privilege_*` tests for that reason). The copy is made at test time from the real file, so
it cannot drift (the unrelated
test `TestGeneratedByAutoharnessMarkerAndLocalDivergenceBannerPresent` keeps reading the real file).

### D-HS-2: PATH scrub by removing sidecar-bearing directories, proven positively

The child environment's PATH is rebuilt by dropping every PATH entry whose directory contains an
executable named `backlogit`, `engram`, or `graphtor-docs` (candidate names: the bare name,
`.ps1` on every OS, and every `PATHEXT` extension on Windows). All other entries are kept, so
`pwsh` and system tools keep resolving. Rejected: an allowlist of only the pwsh directory (cannot
be proven safe for every Windows host's pwsh startup needs, and the failure mode is an obscure
launcher crash rather than a clear skip). Because a denylist over an open grammar can silently miss
a spelling (see `docs/compound/2026-10-01-textual-yaml-guard-denylist-bypass-loop.md`), the guard
is **positive**: inside the scrubbed environment, a real `pwsh` runs
`Get-Command backlogit,engram,graphtor-docs -ErrorAction SilentlyContinue` and the test asserts it
returns nothing. A fake sidecar stub directory is injected into the base environment first so the
guard is red on a pass-through implementation on **every** host, including CI hosts with no real
sidecars (otherwise the guard would be vacuous there).

* Environment surface: PATH plus the caller's overrides only. `start.ps1` does not read `GIT_*`;
  ambient `COPILOT_EXE` is shadowed by the test's `COPILOT_EXE_PATH`. The read-only
  `gh auth token` call and ambient `GITHUB_*` tokens in `Resolve-GitHubTokens` are a noted,
  accepted residual (not a sidecar, no workspace writes). The scrub is intentionally coarse: a
  directory such as `C:\Tools` is dropped whole.
* The env key is matched case-insensitively (`Path` vs `PATH`) via the existing `filterEnvKeys`
  helper, then re-appended; the `os.PathListSeparator` and `filepath.SplitList` are used so the
  same code is correct on Windows and Linux.
* `pwsh` is pinned by absolute path from `exec.LookPath` against the **parent** PATH before
  scrubbing; the second test's helper script invokes the same absolute path rather than bare
  `pwsh`.
* Edge: if `pwsh`'s own directory also holds a sidecar, the scrub would remove `pwsh`; the helper
  calls `t.Skipf` with an explicit reason (honest skip, not a faked pass). This is not reachable on
  the dev box (`C:\Program Files\PowerShell\7` vs `C:\Tools`) or on CI runners.

### D-HS-3: cmd.WaitDelay via one shared bounded runner

A single helper (`startScriptCombinedOutputBounded`) sets `cmd.WaitDelay = startScriptWaitDelay` (5 s) and
then calls `CombinedOutput`, used by both named tests. `WaitDelay` bounds the wait on the
inherited stdout pipe after the context kills the launched process (root cause ii), so a timed-out
`start.ps1` can no longer block until an index worker exits. The existing
`errors.Is(ctx.Err(), context.DeadlineExceeded)` timeout branch is kept first, so the failure
message stays the same. `WaitDelay` is containment only; it does not kill orphans. Removing the
workers at the source is D-HS-1/D-HS-2's job. The claim is proven, not assumed: a helper-process
guard (a child that spawns a grandchild holding the inherited stdout pipe, then outlives the
context) must return within a bound well under the grandchild lifetime; it is red without
`WaitDelay`. Per `docs/compound/2026-09-08-verify-go-stdlib-internals-claims-against-goroot-source.md`
the implementer reads `GOROOT/src/os/exec/exec.go` (`go1.26.9`) before stating `WaitDelay`
semantics in comments.

### D-HS-4: Cross-platform / skip semantics

No new Windows-only skip and no build tags. Both tests keep the existing
`exec.LookPath("pwsh")` skip. CI runs them on `ubuntu-latest` (no sidecars; the scrub is a no-op
there but the guards still prove it through the injected fake stub) and on `windows-latest`.
`GOOS=linux go vet ./tests/integration/...` proves the new code compiles for POSIX from the
Windows dev host.

### D-HS-5: Scope held to the two named tests

`runStartScriptSnippet` (dot-source-only snippet runner) has the same unbounded `CombinedOutput`
shape but never launches a sidecar sync; changing it is outside the operator-accepted statement
and is captured as a P-021 deferred-scope entry (Stage own-authority intake) instead of being
absorbed. The stale comment inside `TestInvokeStartScriptMainPropagatesNonZeroExitCode` ("no real
backlogit/engram/graphtor-docs on PATH ... is fine") is corrected as part of the rewire because it
is the claim this change makes true.

### D-HS-6: Reconciliation of related stash entries

* **`26EDD633` (low, same test, "investigate a budget or host-load guard"): ABSORB and archive.**
  Same surface (`TestInvokeStartScriptMainPropagatesNonZeroExitCode`, same 30 s full-suite
  timeout symptom). Its workspace-size hypothesis is **not confirmed** by the accepted RCA, which
  attributes the cost to constant-enabled sidecars running against the real repository. Making
  the test hermetic removes the sidecar workload, so a separate budget guard is unnecessary. Its
  symptom is covered by an explicit acceptance criterion in the final task (full-suite runs on the
  Windows dev host, each of the two tests well under the 30 s budget). Archive with a promotion
  note naming the new feature and this note.
* **`96B308D1` (`TestGatecheckCleanup_*` Windows path bug, `tools/gatecheck`): NOT absorbed**
  (different test, different package). Left active, untouched.

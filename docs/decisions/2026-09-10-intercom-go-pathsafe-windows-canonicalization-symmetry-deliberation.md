---
title: "Deliberation — Pathsafe Windows Canonicalization Symmetry and Windows Lint Coverage"
date: 2026-09-10
status: accepted
agent: Stage
mode: DARK_MODE_ACTIVE
governs: 016-F / shipment 015-S
---

# Deliberation — Pathsafe Windows Canonicalization Symmetry and Windows Lint Coverage

**Date**: 2026-09-10
**Agent**: Stage (dark-factory cycle, operator AFK)
**Mode**: DARK_MODE_ACTIVE — bounded stash scope, exactly one shipment
**Stash entries in scope**: BF5DE670, 4104AF54, E428AB46, E4C5413F
**Excluded by operator**: A92E3FA0, 4A01C53E, 37FAB8C2, EF9352FB, F47DB9A9,
2787DA56, 4C5BEC23, 9D45E62E, 1C6C3B46, 6C24E2E4
**Governing prior artifacts**:
- `docs/closure/2026-09-08-013-s-014-f-pathsafe-reparse-containment-adversarial-review.md` (findings U1, U5, U6, U7)
- `docs/decisions/2026-09-08-intercom-go-pathsafe-reparse-containment-deliberation.md`
- `docs/plans/2026-09-08-intercom-go-pathsafe-reparse-containment-plan.md`
- `docs/compound/2026-09-08-verify-go-stdlib-internals-claims-against-goroot-source.md`
- `docs/compound/2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md`

---

## 1. Why this deliberation exists (P-021 C6)

All four scoped stash entries carry the literal `DEFERRED SCOPE EXPANSION`
marker written by Ship's P-021 C2 capture. Under the Stage triage precedence
rule this **forces** the `deliberate` route regardless of each entry's apparent
shape, size, priority, or triviality, and none may reach implementation
planning without this artifact. This document discharges that obligation for
all four.

## 2. P-021 C5 intake obligations — both discharged

### (A) Duplicate detection — UNCONDITIONAL, performed on all four

Scanned all 14 active entries in `.backlogit/stash.jsonl` pairwise
(`backlogit stash get` per scoped ID plus a full-file read of every active
record). Three of the four scoped entries carry a
`DISCOVERY-STATUS: LOOKUP-UNAVAILABLE` token, making them declared known-risk
entries whose candidate IDs seed the scan — but per the unconditional rule the
scan covered BF5DE670 as well, which carries no such token.

**Result: CLEAN.** No duplicates found. Nothing merged, nothing archived,
nothing destroyed. The four entries cover four genuinely distinct expansions:

| Entry | Distinct expansion |
|---|---|
| BF5DE670 | TOCTOU / hardlink / dangling-symlink **mitigation** (register already delivered) |
| 4104AF54 | CI lint GOOS coverage for windows-tagged files |
| E428AB46 | `NewRoot` root-construction canonicalization asymmetry |
| E4C5413F | U5/U6/U7 robustness of `reparse_windows.go` |

Adjacent non-scoped entries were checked for overlap and are **not** duplicates:
`37FAB8C2` (black-box runtime verification once a live caller exists) is a
different ask gated on a different trigger; `1C6C3B46` (missing frontmatter on
pathsafe docs) is a documentation defect; `6C24E2E4` (gate scan_scope refactor)
is a different surface. None were triaged, altered, archived, or planned.

**The `LOOKUP-UNAVAILABLE` token's premise is now falsified and should not be
re-emitted.** The token's stated cause was "no backlogit stash query/search
command available in this CLI". backlogit 1.10.1 exposes `stash get`,
`stash list`, and `search`; the scan that capture could not perform has now
been performed. Recorded here so a future capture does not repeat the claim.

### (B) Late-identifier reconciliation — MANDATORY, all four had `N/A` fields

Recovered from Ship-owned residual-risk records (merged PR history) joined on
the deferred entry IDs. Stage performs this under its own pre-existing stash
authority; **no Ship write is required and the C5 capture-only carve-out and
single-write capture invariant are preserved unweakened.**

| Entry | Field | Captured | Reconciled | Source |
|---|---|---|---|---|
| BF5DE670 | PR | `N/A` (pre-PR) | **#27** (merge `d8ca53f5`), closure **#28** (`3d9515c2`) | 009-S release-unit PR |
| BF5DE670 | task | `N/A` (cross-cutting) | **remains `N/A`** — genuinely cross-cutting, no single task | — |
| BF5DE670 | thread | `N/A` | **remains `N/A`** — local pre-PR review gate, no thread ever created | — |
| 4104AF54 | PR | `N/A` (pre-PR) | **#40** (merge `17daffb5`), closure **#41** (`0f85d405`) | 013-S release-unit PR |
| 4104AF54 | thread | `N/A` | **remains `N/A`** — PR #40 has 3 reviews, 0 comment threads | — |
| E428AB46 | PR | `N/A` (pre-PR) | **#40** (merge `17daffb5`), closure **#41** | 013-S release-unit PR |
| E428AB46 | thread | `N/A` | **remains `N/A`** — local adversarial review, no thread | — |
| E4C5413F | PR | `N/A` (pre-PR) | **#40** (merge `17daffb5`), closure **#41** | 013-S release-unit PR |
| E4C5413F | thread | `N/A` | **remains `N/A`** — local adversarial review, no thread | — |

Per the non-blocking rule, the `N/A` values that could not be reconciled
**stand as truthful terminal records** — those findings were raised at local
review gates that never produced a PR review thread. This is not a C3 or C6
shortfall and does not gate anything below.

## 3. Problem frame

`internal/pathsafe` is a NON-NEGOTIABLE containment control. Shipment 013-S
closed a real, empirically reproduced Windows directory-junction containment
bypass by introducing `canonicalizeReparse`
(`GetFinalPathNameByHandleW`-based, junction-aware) on the **candidate** side
of the containment comparison. It did not touch the **root** side.

That left the control comparing two paths produced by two different
canonicalization mechanisms, and left the new Windows-only code structurally
invisible to the repository's only lint gate.

### Verified evidence gathered this cycle (not inherited)

Every claim below was measured directly at `db4edf0` this session:

1. **`lint` job is `runs-on: ubuntu-latest` and runs `golangci-lint run ./...`.**
   Go file discovery is GOOS-scoped, so `internal/pathsafe/reparse_windows.go`
   (`//go:build windows`) is never analyzed. Confirmed by reading
   `.github/workflows/ci.yml`. A `test (windows, advisory)` job already exists
   on `windows-latest`, so a Windows runner is already an established,
   budgeted shape in this pipeline. **4104AF54 confirmed real.**
2. **`NewRoot` (root.go:265) uses `filepath.EvalSymlinks(abs)` then
   `stripUNCPrefix`; `checkSymlinkEscape` (pathsafe.go:359) uses
   `canonicalizeReparse(ancestor)` then `stripUNCPrefix`, compared at
   pathsafe.go:371 via `hasPathPrefix(real, root.path)`.** The asymmetry is
   exactly as described. **E428AB46 confirmed real.**
3. **U5/U6/U7 all confirmed by reading `reparse_windows.go`**: `syscall.CreateFile`
   is called on the raw path with no `\\?\` prefixing; `canonicalizeReparse`
   returns the raw `\\?\`-prefixed form and relies on its sole caller to strip;
   `procGetFinalPathNameByHandleW.Call` via `syscall.LazyProc` panics rather
   than returning an error if the export cannot be resolved.
   **E4C5413F confirmed real.**
4. **BF5DE670's own archival trigger has NOT fired.** Ran
   `scripts/check-write-path-precondition.sh` (exit 0, clean) and
   `--self-test` (all 5 accept/reject fixtures pass, so the gate is genuinely
   functional and not vacuously green). An independent grep for
   `os.WriteFile|Create|OpenFile|Remove|RemoveAll|Rename|Mkdir|MkdirAll|Symlink|Chmod|Truncate`
   across non-test `internal/**` and `cmd/**` returns **zero** matches.
   pathsafe still gates **no** destructive write path.

### Severity correction discovered during this deliberation

E428AB46's stash text self-assesses as "plausible-but-unconfirmed; no
reproduction was attempted". The 013-S adversarial review contains an
**independently traced** analysis that supersedes that provisional wording:

> "this specific configuration worked correctly pre-013-S and is broken (in the
> fail-closed direction) by this shipment."

This is therefore a **regression introduced by 013-S**, not a pre-existing
theoretical gap. Under a junctioned workspace root, `Lstat` on the root reports
the junction's own `ModeIrregular` and trips the unconditional ancestor
rejection, or the candidate resolves to its real target while `root.path`
remains the unresolved junction path — so `hasPathPrefix` compares two
namespaces and **every** access under such a root fails, in the worse case as a
false-positive `symlinkEscapeMsg` asserting a "GENUINE, VERIFIED escape" for a
workspace that never escaped. Redirected project directories and
corporate-policy-redirected folders are plausible real deployment shapes.

This raises E428AB46 from `medium` to the highest-priority item in scope and is
recorded as decision **D-2** below.

## 4. Options considered

### Grouping options

**Option G1 — all four entries in one shipment** (the Orchestrator's candidate).
Rejected. BF5DE670's deliverable is a *mitigation* (Lstat-before-write, O_EXCL
revalidation, hardlink-count checks) that must attach to a filesystem write
call site. Evidence item 4 above proves **zero such call sites exist**.
Implementing it would require first *inventing* a write API — which is neither
in any scoped entry nor authorized, i.e. scope expansion — or writing
machinery against an API that does not exist, i.e. speculative engineering
that the entry's own two prior Stage dispositions already rejected on the same
measured grounds.

**Option G2 — three entries (4104AF54, E428AB46, E4C5413F), defer BF5DE670.**
**CHOSEN.** These three form a genuinely causal cluster, not merely a
co-located one (see D-1).

**Option G3 — E428AB46 alone, defer the rest.** Rejected. It would ship a
change to `reparse_windows.go`'s caller graph while leaving that file
unlintable (4104AF54) and would make U6 load-bearing without fixing it (see
D-1's dependency argument).

### E428AB46 fix options

**Option A — `NewRoot` canonicalizes through `canonicalizeReparse` + `stripUNCPrefix`.**
**CHOSEN (D-2).** Puts both sides of the containment comparison in one
namespace *by construction* rather than by coincidence. Notably this is a
**provable no-op on non-Windows**: `reparse_other.go`'s `canonicalizeReparse`
is literally `filepath.EvalSymlinks`, so POSIX behavior is bit-identical and
the blast radius is confined to Windows. This is also the fix prescribed by the
013-S adversarial review.

**Option B — reject junction/GUID-volume roots outright.** Rejected. Hostile to
legitimate deployments (redirected project folders, dev drives, volumes mounted
without a drive letter) and converts an availability bug into a hard refusal.

**Option C — document + regression test only, no behavior change.** Rejected as
the primary remedy: it leaves a known fail-closed regression live. Its
regression test is nonetheless **retained as required evidence** under Option A.

### 4104AF54 gate-posture options

**Option L1 — add the Windows lint job as advisory.** Rejected. Learning from
014-S: a plan that made a blocking gate advisory without restoring it was a net
enforcement downgrade. Any advisory window must be self-restoring within the
same shipment; simply avoiding one is cleaner.

**Option L2 — add the Windows lint job blocking, discovering and fixing any
findings it surfaces in the same shipment.** **CHOSEN (D-3).** The
windows-tagged production surface is exactly one file
(`reparse_windows.go` — 4104AF54 itself states it is the repo's first), which
is already in scope, so the remediation surface is bounded and in-scope by
construction.

## 5. Decisions

**D-1 — Grouping: three entries, one shipment.** The cluster
{4104AF54, E428AB46, E4C5413F} is coherent on a stronger basis than shared
directory. It is **causally ordered**:

- **AMENDED 2026-09-10 after adversarial plan review.** This decision
  originally rested on E4C5413F's **U6** — the claim that D-2 creates the
  "second caller" U6 warns about, making U6 a hard prerequisite. **Review
  falsified that claim and it is retracted here rather than preserved.** D-2
  retains `NewRoot`'s own caller-side `stripUNCPrefix`, `checkSymlinkEscape`
  already strips at `pathsafe.go:369`, and `stripUNCPrefix` was read and
  confirmed idempotent — so both callers stay normalized regardless of
  ordering. U6 is a real encapsulation improvement, not a correctness gate.
- **The dependency that IS load-bearing is U5 → D-2**, discovered by the same
  review. `NewRoot` currently gets Go's internal long-path handling for free
  via `filepath.EvalSymlinks`' `os.Lstat` walk. Routing it through
  `canonicalizeReparse`'s raw `syscall.CreateFile` removes that handling, so
  landing D-2 without U5 would introduce a **new** fail-closed regression for
  long workspace roots. U5 and D-2 must land together.
- **4104AF54** is the *verification surface* for both: every change above lands
  in windows-tagged files that the current gate structurally cannot lint.
  Without it, this shipment's own code ships unlinted.

The grouping conclusion is unchanged; only its supporting dependency is
corrected. The three entries remain one coherent, causally ordered cluster.

**D-2 — `NewRoot` adopts `canonicalizeReparse`** (Option A), preserving
verbatim: the 009-S `apperr.Wrapf` dual-discriminability guarantees on
`NewRoot`'s failure branches, and the existing non-directory root rejection.
The `DOCUMENTED-UNREACHABLE` coverage-excluded reasoning on the subsequent
`os.Stat` must be **re-derived, not merely re-pasted**, because its current
justification is explicitly grounded in `EvalSymlinks`' Lstat walk having
already succeeded.

**D-3 — Windows lint job lands blocking, not advisory** (Option L2), and its
workflow edit is validated by **both** `yaml.safe_load` and `actionlint` before
push — the 010-S compound learning records that neither alone suffices and that
multi-persona review did not catch a structural Actions error.

**D-4 — BF5DE670 is DEFERRED and REMAINS ACTIVE, not archived.** Its trigger
was re-measured this cycle and has not fired. It is the only tracker for a
deferred security mitigation; archiving it would retire the tracker while the
risk acceptance is live. Its archival condition is unchanged: archive when the
mitigation lands, or when
`scripts/check-write-path-precondition.sh` fires. This is an
evidence-backed disposition inside scope, **not** a silent drop.

**D-5 — No claim that `filepath.EvalSymlinks` internally calls
`GetFinalPathNameByHandleW` may enter any artifact or doc comment.** This is a
twice-corrected known falsehood in this repository
(`docs/compound/2026-09-08-verify-go-stdlib-internals-claims-against-goroot-source.md`).
Doc text must state only the verified narrow claim.

**D-6 — `root_windows_test.go`'s `"volume guid"` case pins non-stripping of
`\\?\Volume{GUID}` as INTENDED behavior and must not be "fixed".** U6's
internalization of `stripUNCPrefix` must preserve it.

## 6. Scope boundary (anti-goals)

- **No** TOCTOU, hardlink, or O_EXCL mitigation machinery (D-4).
- **No** filesystem write API is introduced.
- **No** relaxation of the F3 `ModeIrregular` strict-ancestor rejection —
  the 013-S review's U2 records that relaxing it without making the branch
  reparse-aware silently reintroduces the bypass with no test turning red.
- **No** changes to the ten operator-excluded stash entries.
- **No** second shipment.

## 7. Open questions

None blocking. The one judgment call — whether a junctioned workspace root is a
supported deployment shape — is answered affirmatively by D-2 on the evidence
that it **worked before 013-S**, so preserving it is regression repair rather
than new feature scope.

## 8. Dispositions

| Entry | Disposition | Rationale |
|---|---|---|
| 4104AF54 | **HARVEST** into 016-F | Verification prerequisite; confirmed real |
| E428AB46 | **HARVEST** into 016-F | 013-S regression, fail-closed availability break |
| E4C5413F | **HARVEST** into 016-F | U6 is a hard prerequisite of D-2; U5/U7 same file |
| BF5DE670 | **DEFER, remain active** | Trigger measured, has not fired (D-4) |

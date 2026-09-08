# Deliberation — pathsafe Containment Correctness and Cross-Platform Coverage

- **Date**: 2026-09-07
- **Agent**: Stage (P-017 dark-factory cycle, post-010-S)
- **Mode**: DARK_MODE_ACTIVE, operator AFK, hard limit of exactly ONE shipment
- **Scope under consideration**: 13 stash entries
  (A92E3FA0, EF9352FB, 4989A42D, BF5DE670, ECE3DAB7, F47DB9A9,
  F4F4A959, 5FE4A7BE, 5158769F, 2787DA56, CA469B3D, 4C5BEC23, 805248F7)
- **Governing predecessors**:
  - `docs/decisions/2026-09-06-intercom-go-residual-hardening-round2-deliberation.md`
  - `docs/decisions/2026-09-04-intercom-go-residual-hardening-triage-deliberation.md`
  - `docs/plans/2026-09-04-intercom-go-architecture-correction-plan.md` (phases C1–C11)

---

## 1. Problem Frame

Thirteen stash entries were in scope. Every one is either a deferred scope expansion
captured under P-021 C1 during shipments 001-S / 009-S / 010-S, or a standing tracker.
None had been harvested. The task was to group them by contextual/code similarity,
select the single highest-value coherent group that is SAFE to execute autonomously,
and harvest only that group.

The distinguishing question was not "what is largest" but "which group contains the
highest product-outcome defect that an agent may safely fix without operator input".

---

## 2. Entry Classification (all 13, verified against merged HEAD)

Every claim below was verified directly against source at merged HEAD — not accepted
from the stash text. Four of the low-confidence entries flagged
"unverified by aggregator" were independently confirmed TRUE during this triage.

### Group A — `internal/pathsafe` containment correctness and cross-platform coverage

| Stash | Pri | Kind | Claim | Verification |
|---|---|---|---|---|
| ECE3DAB7 | medium | bug | `checkSymlinkEscape` ancestor walk uses `os.Stat` (follows symlinks); a **dangling intermediate** symlink is walked PAST without re-resolving its target | **CONFIRMED REAL** — `pathsafe.go` `checkSymlinkEscape`, `os.Stat(ancestor)` errors on a dangling link → `ancestor = parent` → escape |
| CA469B3D | low | bug | `stripUNCPrefix` mishandles `\\?\UNC\server\share` → yields relative `UNC\server\share` | **CONFIRMED REAL** — `root.go` is a bare `strings.TrimPrefix(p, uncPrefix)`; breaks `Root.Path()`'s absolute-path contract |
| 5FE4A7BE | low | bug | Case-folding gated on `GOOS=="windows"` only, but darwin is a shipped target with zero CI coverage | **CONFIRMED REAL** — `root.go`/`pathsafe.go` `pathEqual`/`pathHasPrefix` fold on Windows only; `scripts/targets.json` **does** list `darwin/amd64` and `darwin/arm64` |
| 5158769F | low | bug | `createDirectorySymlink` skips on ANY `os.Symlink` error, unlike `requireSymlinkOrFailClosed` | **CONFIRMED REAL** — `tests/integration/output_path_guard_test.go:78-89` is a bare `t.Skipf`; `internal/pathsafe/symlink_privilege_windows_test.go` has the fail-closed-except-1314 predicate |
| 805248F7 | low | bug | Two test doc comments begin mid-sentence, truncation artifacts | **CONFIRMED REAL, exact lines located** — `lexical_test.go:69-70` and `symlink_test.go:171` |

### Group B — CI workflow and shell-script gate hygiene

| Stash | Pri | Kind | Surface |
|---|---|---|---|
| F4F4A959 | low | bug | `.github/workflows/ci.yml:263` bare undocumented `continue-on-error: true` on retired-architecture gate |
| F47DB9A9 | low | bug | `ci.yml` references two docs files that do not exist |
| 2787DA56 | low | bug | `scripts/check-unignore-regression.sh` treats any `git show` failure as empty baseline |
| 4C5BEC23 | low | bug | `check-unignore-regression.sh` header comment describes a `git worktree` design the implementation does not use |

### Group C — Non-implementable trackers / externally-authorized

| Stash | Pri | Kind | Disposition |
|---|---|---|---|
| A92E3FA0 | high | feature | Standing roadmap tracker + the ONLY index of 5 unresolved **operator** decisions (Q3/H3, Q6/H4, H5, H6, Q7). Not agent-executable. |
| 4989A42D | medium | feature | Migration-remediation tracker; completion condition (b) re-measured UNMET (`check-retired-architecture.sh:75,391` still `internal/config/**`). |
| EF9352FB | high | task | Tavily credential rotation — **operator-only external action**. |
| BF5DE670 | medium | task | pathsafe TOCTOU/hardlink **mitigation** — trigger has not fired. |

---

## 3. Options Considered

### Option 1 — Ship Group A (pathsafe containment correctness) — **CHOSEN**

Five all-bug entries on one tight code surface (`internal/pathsafe/**` plus that
package's own integration test). Contains the single confirmed **security containment
bypass** in the whole 13-entry scope.

- **Pro**: fixes a real escape-from-workspace primitive in the control that Brief §10
  marks NON-NEGOTIABLE; tightest code cohesion of any group; every item is a bug
  (highest position in the operator's type order); no operator decision required.
- **Con**: touches a security-critical file, so demands red-phase-first discipline
  (Constitution VIII) and raises per-task review cost.

### Option 2 — Ship Group B (CI/script gate hygiene)

- **Pro**: lower blast radius; entirely meta-tooling.
- **Con**: no product-runtime impact. Critically, **two of its four items require
  operator policy decisions** that cannot be made AFK without scope expansion:
  F4F4A959 asks whether to flip a security-adjacent gate from permanently-advisory to
  operator-toggleable (a CI-blocking policy change), and F47DB9A9 asks whether to
  *author* two missing documents or *redirect* the references (an authoring decision
  with no correct autonomous answer). Selecting Group B would have forced either an
  operator wait or a scope-expansion stop.

### Option 3 — Mixed "all low-hanging fruit" group

- **Rejected**: violates the operator's rule 1 (group by contextual/code similarity).
  It would fuse YAML, bash, and Go security code into one PR with no shared review
  context and a diffuse rollback story.

### Option 4 — Ship nothing (closure-only)

- **Rejected**: a confirmed containment bypass is present and is safely fixable. Doing
  nothing would leave an exploitable defect live for another cycle with no
  justification.

---

## 4. Decision: Group A selected

### Why Group A outranks Group B (explicit ranking rationale)

1. **Operator rule 3 — reliability and security supersede everything.** Group A
   contains ECE3DAB7, a *confirmed logic defect* in the containment control: a
   dangling intermediate symlink lets a `Resolve()`-blessed path escape the workspace.
   (Adversarial review correctly noted this is not *uniquely* exploitable — GO-14 is the
   same class, and both await a real write caller — but ECE3DAB7 is the instance that is
   both in-scope and safely repairable now.) Group B contains zero runtime-security
   defects; it is build-pipeline self-documentation.
2. **Operator rule 2 — type order (bug first).** Both groups are 100% bugs, so this
   is a tie; it is broken by product outcome. Group A repairs the product's security
   boundary; Group B repairs the CI system's description of itself.
3. **Operator rule 1 — contextual/code similarity.** Group A is a single-package
   cluster: all five entries land in `internal/pathsafe/**` plus that package's own
   integration test. Group B spans GitHub Actions YAML, Bash, and absent Markdown —
   materially looser cohesion.
4. **Safety under AFK.** Every Group A item has exactly one defensible technical
   answer. Two Group B items require operator *policy* judgment and would have
   triggered a declared stop condition.
5. **Group C is not implementable at all** — three trackers plus an
   externally-authorized credential action.

### Critical scope boundary: ECE3DAB7 is NOT the deferred BF5DE670 mitigation

These are easy to conflate and must not be. **BF5DE670** is the deferred
TOCTOU/hardlink *mitigation* (Lstat-before-write, O_EXCL revalidation, hardlink
counts) whose trigger — "before pathsafe gates any live destructive file-write path"
— has **not fired** (re-measured: zero write primitives under `internal/**`/`cmd/**`).
Building it now is speculative engineering against an API that does not exist.

**ECE3DAB7 is different in scope**: it is a *logic defect in containment code that
already exists and already runs*. `checkSymlinkEscape` was explicitly written to
close the symlinked-intermediate-directory hole (its own doc comment says so, and
`TestResolveRejectsSymlinkedIntermediateDirEscapingRoot` pins it) — it simply fails to
do so when that intermediate symlink is *dangling*. Fixing an existing control that
does not do what it documents is repair, not speculative mitigation. **BF5DE670
therefore remains excluded and unconsumed.**

**Correction after adversarial review (ADV-04)**: the original wording of this section
called ECE3DAB7 "the only confirmed exploitable defect" and framed the two as
"different in kind". That was overstated and is withdrawn. Both currently depend on the
same absent write caller, so neither is exploitable through production code at merged
HEAD; and GO-14 — retained under BF5DE670 — is itself a write-through escape of the
*same class*, differing only in whether the offending link sits at the final component
or a strict ancestor. The honest boundary is **scope**, not kind: this shipment repairs
a narrow strict-ancestor defect in existing logic, while BF5DE670 retains final-component
GO-14, the general TOCTOU window, and hardlink mitigation. See plan §7.1 RESIDUAL, which
states plainly that this shipment *narrows* the escape class rather than closing it.

---

## 5. Design Decisions

### D-1 — ECE3DAB7: scope the existence probe by walk depth

The exploit path, traced concretely: with `root/danglinglink -> /outside/not-yet-there`,
`Resolve("danglinglink/new.txt")` passes the lexical `hasPathPrefix` check; then
`os.Stat(root/danglinglink/new.txt)` fails, `os.Stat(root/danglinglink)` **also** fails
(Stat follows the link to a non-existent target), so the walk steps past the symlink to
`root`, which exists and is trivially contained. `Resolve` returns a path that the OS
will later resolve *through* the symlink to a location outside the workspace.

| Option | Assessment |
|---|---|
| **1. Swap `os.Stat` → `os.Lstat` wholesale** | **Rejected.** `Lstat` sees a dangling *final* component as existing, so `EvalSymlinks` fails and the path is rejected — reversing the documented GO-14 oracle-parity acceptance and breaking its explicit regression lock `TestResolveAllowsDanglingSymlinkAtFinalComponent`. Too blunt. |
| **2. `readlink` + manual target re-resolution during the walk** | **Rejected.** Reimplements symlink resolution by hand inside a security control, adds relative/absolute target join logic and a chase loop. Violates "simplicity supersedes complexity". |
| **3. Depth-scoped probe — CHOSEN** | Probe the **first** iteration (`ancestor == resolved`, the final component) with `os.Stat`; probe every **strict ancestor above it** with `os.Lstat`. |

**Why Option 3 is exactly right**: GO-14 is *defined* as the final-component case, and
the first loop iteration *is* the final component — so the documented acceptance is
preserved by construction, not by coincidence. For every strict ancestor, `Lstat` and
`Stat` differ **only** when that ancestor is a dangling symlink, which is precisely the
defect. A real directory, an in-root symlink, and an escaping symlink all behave
identically to today. The blast radius is therefore surgically limited to the buggy case.

**Accepted, deliberate tightening**: an intermediate dangling symlink is now rejected
*regardless of where its target would point*, because `EvalSymlinks` fails on it. This
is correct fail-closed behavior — pathsafe cannot verify the containment of a target
that does not exist yet, since it may be created as anything before the write lands.
Accepting it would simply reintroduce the same escape. This tightening must be
documented in the risk register (T3) and pinned by a positive test.

### D-2 — CA469B3D: re-form the extended-UNC prefix

Fix `stripUNCPrefix` so a remainder beginning `UNC\` is re-formed as `\\server\share`
rather than left as the relative string `UNC\server\share`. Red-phase Windows-only
regression test first, per Constitution VIII for containment code.

### D-3 — 5FE4A7BE: register + coverage, do NOT extend case-folding

The stash entry's own suggested remedy ("extend case-folding to darwin") is the
**unsafe** option and is explicitly rejected. Direction of error matters:

- **Under-folding (today)**: on case-insensitive APFS a legitimate path may be
  *rejected*. Fail-closed — an availability bug, not a security hole.
- **Over-folding (extending to darwin)**: macOS volumes **can** be formatted
  case-sensitive. On such a volume `/a/B/evil` is genuinely outside root `/a/b`, and
  folding would **accept** it. That converts a fail-closed bug into a **fail-open
  containment bypass** — precisely backwards for a security control under operator
  rule 3.

**Chosen**: add a consolidated-risk-register entry with explicit STATUS/TRIGGER
(matching the established 011.004-T convention) recording **both** directions of the
case-folding risk — the darwin under-fold *and* the fact that the already-enabled
**Windows** fold is itself a potential fail-open on per-directory case-sensitive NTFS
(WSL-created trees). The darwin availability defect is labelled
**plausible/unconfirmed** pending a reproduction. A runtime case-sensitivity probe in
`NewRoot` was rejected as speculative complexity in a hot constructor.

**Revised after review**: an advisory `macos-latest` CI job was originally proposed and
has been **dropped**. Four reviewers independently established that GitHub's macOS
runners use case-**insensitive** APFS by default, so the job could not exercise the
case-sensitive hazard it was meant to cover; it was permanently advisory, and it was the
sole reason this shipment would open `.github/workflows/ci.yml` — the shared surface of
excluded entries F4F4A959 and F47DB9A9. 5FE4A7BE's own text authorizes the remedies with
"and/or", so the register entry discharges it honestly on its own.

### D-4 — 5158769F: mirror the fail-closed predicate

Give `tests/integration` a build-tagged privilege predicate equivalent to
`isSymlinkPrivilegeError` and make `createDirectorySymlink` fail-closed except for
ERROR_PRIVILEGE_NOT_HELD (1314). `internal/pathsafe`'s predicate is unexported and
cannot be imported across the package boundary, so a local equivalent is required.

### D-5 — 805248F7: safe_auto docs repair

Exact lines located: `lexical_test.go:69-70`, `symlink_test.go:171`. Restore the
function name and opening clause. No deliberation needed.

---

## 6. Excluded Entries — preserved unchanged and unconsumed

| Stash | Reason for exclusion |
|---|---|
| A92E3FA0 | Standing roadmap tracker; sole index of 5 open **operator** decisions. Not agent-executable. |
| 4989A42D | Migration tracker; completion condition (b) re-measured and UNMET. Broadening the gate is unbidden scope expansion. |
| EF9352FB | **HARD BLOCKER, operator-only.** Rotating a live Tavily credential requires a third-party console. Repo-side containment re-verified CLOSED (no `.env*` ever committed; ignored via `.gitignore:36`). Deliberately isolated into NO shipment so it cannot stall safe work. Verified at metadata level only — no key material read. |
| BF5DE670 | Deferred mitigation; trigger has not fired (zero write primitives). See §4 boundary note. |
| F47DB9A9, F4F4A959, 2787DA56, 4C5BEC23 | Group B — coherent, real, and worth a future cycle, but excluded by the one-shipment limit and outranked per §4. Two require operator policy decisions. |

---

## 7. Open Questions

None blocking. Every Group A item has a single defensible technical answer recorded
above. No P0/P1 finding remains unresolved.

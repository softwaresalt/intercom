---
title: "Deliberation — Retired-Architecture Gate Detection-Quality Hardening"
date: 2026-09-08
status: accepted
agent: Stage
mode: DARK_MODE_ACTIVE
governs: 015-F / shipment 014-S
---

# Deliberation — Retired-Architecture Gate Detection-Quality Hardening

- **Date**: 2026-09-08
- **Agent**: Stage (P-017 dark-factory cycle, post-013-S)
- **Mode**: `DARK_MODE_ACTIVE`, operator AFK, hard limit of exactly **ONE** shipment
- **Scope under consideration**: all 22 active stash entries, of which **8 are selected**:
  `24B63533`, `8988120D`, `6285779C`, `8D603E29`, `9A14D3B7`, `E403E3C5`, `09CD6ACF`,
  `F4F4A959`
- **Governing predecessors**:
  - `docs/closure/2026-09-08-012-s-retire-d6a-gate-narrowing-adversarial-review.md` (U-1…U-9)
  - `docs/closure/012-S-013-F-post-merge-closure.md` (PR #36, merge `9d8937ae`)
  - `docs/closure/010-S-011-F-post-merge-closure.md` (PR #29)
  - `docs/decisions/2026-09-08-intercom-go-pathsafe-reparse-containment-deliberation.md` §5
    (the dispositions table that named this cluster as the next cycle's shipment)

---

## 1. Framing — what problem are we actually solving?

`scripts/check-retired-architecture.sh` is the **anti-accident hygiene control** that stops
the retired Slack / ACP / IPC / socket-mode architecture from being silently reintroduced.
Decision **D6b** explicitly classifies it as an anti-accident gate, **not** a security
boundary — that classification is preserved unchanged by this cycle and is the reason this
cluster did not outrank the critical containment bypass in the previous cycle.

Shipment 012-S (feature 013-F, PR #36) broadened the gate's **scan scope**
(`internal/config/**` → `internal/**`) and added a Go fixture harness. It deliberately
froze the gate's **detection vocabulary** as anti-goal **I-D**. Adversarial, correctness,
and architecture review of 013-F consequently produced a cluster of findings that are all
about the thing 012-S refused to touch: *the gate scans more files, but it still cannot
reliably recognise a retired token when it sees one.*

The problem this cycle solves is therefore precise:

> The gate's **coverage** was widened without widening its **detection accuracy**, so the
> broadened-coverage claim currently overstates the protection actually delivered.

A gate that scans everything and detects the wrong things is worse than a narrow gate,
because it manufactures false confidence. That is a **reliability** defect in a
merge-blocking control, which is why it ranks first under operator rule 1.

---

## 2. Empirical verification performed during this deliberation

Stage did **not** take the stash entries on trust. Read-only verification against `HEAD`:

**(a) Token decomposition — verified by executing the shipped regex in isolation.**
`component_re = [A-Z]+(?=[A-Z][a-z]|$)|[A-Z]?[a-z]+|[0-9]+` at line 85.

| Identifier | `split_identifier` output | Result |
|---|---|---|
| `ChannelID` | `['channel','id']` | caught |
| `ChannelIDs` | `['channel','i','ds']` | **MISSED** |
| `TeamIDs` | `['team','i','ds']` | **MISSED** |
| `IPCNames` | `['ipc','names']` | **MISSED** |
| `SocketModes` | `['socket','modes']` | **MISSED** |
| `HostCLIs` | `['host','cl','is']` | **MISSED** |
| `socketmode` / `Socketmode` | `['socketmode']` | **MISSED** |

This **confirms** `24B63533` item (1) and `8988120D` item (1), and surfaces a **third,
previously unrecorded failure mode**: a plain lowercase suffix on the *final* component
(`SocketModes` → `modes` ≠ `mode`) defeats the exact-sequence match in
`matches_forbidden_parts` even when decomposition is otherwise correct. The practical
consequence is broader than either entry stated: **every plural form of every multi-part
forbidden token escapes the gate today.** Only the single-part tokens (`slack`, `acp`) are
immune.

This third mode is **not** new scope — it is the same defect class, in the same two
functions (`split_identifier` / `matches_forbidden_parts`), that the two selected entries
already authorise hardening for. It is recorded here as refined evidence, not as a new
stash entry.

**(b) Fail-open dispatch — verified by reading `scan_path` (line 414) and
`engine_for_path` (line 100).** `engine_for_path` returns `None` for any unmapped
extension and `scan_path` then returns `[]`. A selected path with no engine branch is
silently reported clean. **Confirms `9A14D3B7`.**

**(c) `cmd/` filter asymmetry — verified by reading `should_scan_repo_path` (line 88).**
The `cmd/` branch is an early `return path.endswith('.go')` placed *above* the
`_test.go` / `/testdata/` exclusions that the `internal/` branch reaches. **Confirms
`8D603E29`**: `cmd/**/*_test.go` is scanned today, `internal/**/*_test.go` is not.

**(d) Dual scan-scope source — verified.** `select_repo_paths` (line 423) passes the
pathspec `config.toml.example cmd/** internal/**` to `git ls-files`, and
`should_scan_repo_path` re-expresses the same scope as prefix tests ~335 lines away.
**Confirms `6C24E2E4` item (1)** — which is nonetheless *deferred*, see §4.

---

## 3. Options considered

### Option A — Ship the whole of "Cluster B" (all 9 entries) in one shipment

The previous cycle's disposition table named all 9 entries as one coherent cluster.

- **Pro**: discharges the entire cluster; no follow-up cycle needed.
- **Con**: includes `6C24E2E4`, a ~435-line structural extraction of the embedded Python
  engine into `scripts/lib/`. That refactor rewrites the same file every other task edits,
  producing a large, conflict-dense diff against a **merge-blocking** control.
- **Con**: `6C24E2E4`'s own stated payoff is de-duplicating the `mask_go_non_code` clone
  shared with `scripts/check-write-path-precondition.sh` — and that file is an
  **explicitly EXCLUDED surface** under stash `BF5DE670`. Pursuing the refactor's actual
  value invites a P-021 exclusion-boundary violation.

### Option B — Ship only the two `medium` bugs (`24B63533`, `8988120D`)

- **Pro**: smallest, safest diff.
- **Con**: leaves the fail-open dispatch and the uncharacterised masking state machine in
  place, so the corrected vocabulary would land **unproven**. `6285779C` exists precisely
  because a masking regression can currently land fully green.

### Option C (**chosen**) — Ship the detection-correctness core; defer the structural refactor

Select the 8 entries that together make the gate **detect correctly and prove that it
does**, and defer `6C24E2E4` alone.

- **Pro**: every selected entry changes detection behaviour or the evidence for it — one
  coherent purpose.
- **Pro**: fixes land **with** the fixtures and self-test assertions that prove them, so
  no correction is unproven.
- **Pro**: avoids the exclusion-boundary risk and the large behaviour-preserving diff.
- **Pro**: honours operator rule 2 (simplicity/composability) without letting a
  behaviour-preserving refactor displace live reliability defects.
- **Con**: `6C24E2E4` remains open, so the dual-scope drift risk it names persists one
  more cycle. Accepted: 013-F already mitigated it with assertions (`013.001-T`), so the
  residual is drift *detection*, not drift *exposure*.

---

## 4. Decision

**Option C is adopted.**

**Selected (8):** `24B63533`, `8988120D`, `6285779C`, `8D603E29`, `9A14D3B7`, `E403E3C5`,
`09CD6ACF`, `F4F4A959` → feature **015-F**, shipment **014-S**.

**Deferred from the cluster (1):** `6C24E2E4` — behaviour-preserving structural refactor
whose payoff crosses into the `BF5DE670` excluded surface. It should lead a subsequent
cycle, **after** this cycle's vocabulary changes have settled, so that the extraction is a
pure move over already-correct code.

### Scope decisions taken during deliberation

- **D1 — Merge the recorded overlap.** `24B63533` item (1) and `8988120D` item (1)
  describe two *distinct* failure modes (acronym backtracking vs. no-separator
  concatenation) in the *same* two functions. The prior cycle recorded this as an overlap
  that "MUST be merged into a single unit of work when Cluster B is harvested, or the fix
  will be implemented twice." Honoured: both, plus the newly found suffix mode, become the
  single task `015.001-T`.
- **D2 — `F4F4A959` stays default-advisory.** Converting the CI step to a documented,
  operator-controlled `RETIRED_ARCH_GATE_REQUIRED` toggle matching the established
  `PIPELINE_TOPOLOGY_GATE_REQUIRED` / `WINDOWS_GATE_REQUIRED` convention. The toggle
  **defaults to advisory**, so this cycle makes the decision *visible and flippable*
  without turning the gate red on `main` in the same shipment that changes its detection
  vocabulary. Flipping it is a separate, explicit operator decision.
- **D3 — `09CD6ACF` is disclosure-only.** The entry names three blind spots
  (aliased imports masked with string literals, `go.mod`/`go.sum` outside the pathspec,
  non-`.go`/`.toml` files under `internal/**`). Closing them means **expanding scan
  scope**, which is a fresh CI-red risk and a separate decision. This cycle only
  **documents** them so the D6c broadened-coverage claim stops overstating protection.
  Scope expansion is explicitly *not* authorised here.
- **D4 — `8D603E29` resolves toward coverage, not away from it.** Of the entry's two
  offered resolutions, option (a) *reduces* coverage. For an anti-accident gate that is the
  wrong direction; the asymmetry is resolved by documenting `cmd/` test scanning as
  intentional and recording the `internal/**/*_test.go` gap, with **no coverage reduction**.
- **D5 — False-positive blast radius is the dominant risk.** Every vocabulary widening in
  `015.001-T`–`015.003-T` can turn `main` red for a legitimate identifier. Each such task
  must run the full repo scan at `HEAD` and prove zero new findings before it is
  acceptable. This is the single most important acceptance constraint in the plan.

---

### Addendum (rev 3, same day) — decision **D2 is SUPERSEDED**

Plan review discovered that **D2 rested on a false premise**. D2 assumed the
retired-architecture gate was permanently advisory in CI. It is **not**:
`.github/workflows/ci.yml:265-266` runs `--self-test` with **no** `continue-on-error`, and
`scripts/check-retired-architecture.sh:639-641` shows `--self-test` runs
`scan_with_mode self-test` **followed by `scan_with_mode repo`** — the full repo scan.
**The gate has been hard-blocking all along.**

Consequences:

- **D2 as written would have caused a net downgrade** of a live merge-blocking control while
  the operator is AFK — the opposite of operator rule 1.
- **Replaced by D2′**: the shipment opens a **bounded advisory window** (plan unit
  `015.001-T`) so that deliberately-RED test-first commits do not wedge `main`, and
  **closes it inside the same shipment** (`015.013-T`) in a **fail-closed** form
  (`RETIRED_ARCH_GATE_ADVISORY == 'true'`, defaulting to blocking). Enforcement ends the
  shipment **no weaker than it began, and strictly more robust** to toggle misconfiguration.
- **Open question 2 is withdrawn** — there is no "flip to required" decision left for the
  operator, because the gate is not being downgraded. What remains is only the optional
  future choice to *relax* it, which is out of scope.

`F4F4A959`'s intent is fully preserved: the invisible, undocumented `continue-on-error`
default becomes an explicit, documented, operator-controllable, per-run-logged toggle.

---

## 5. P-021 intake compliance for this cycle

**Duplicate detection — UNCONDITIONAL, performed over all 22 active entries: CLEAN.**
No non-Cluster-B entry references `check-retired-architecture.sh`. The only overlap is the
`24B63533` ↔ `8988120D` item-(1) overlap, which is **not** a duplicate (neither entry is a
subset of the other) and is resolved by decision **D1**, not by archival. **Nothing was
merged, archived, or destroyed by the duplicate scan.**

**Late-identifier reconciliation — MANDATORY where source refs were `N/A`, performed under
Stage's own stash authority (no Ship write, single-write capture invariant preserved):**

| Entry | Captured | Recovered this cycle | Source record |
|---|---|---|---|
| `8988120D` | PR `N/A`, thread `N/A` | **PR #36** | `docs/closure/012-S-013-F-post-merge-closure.md` |
| `9A14D3B7` | PR `N/A`, thread `N/A` | **PR #36** | same |
| `E403E3C5` | PR `N/A`, thread `N/A` | **PR #36** | same |
| `09CD6ACF` | PR `N/A`, thread `N/A` | **PR #36** | same |
| `F4F4A959` | PR `N/A` (pre-dates 010-S) | **PR #29** | `docs/closure/010-S-011-F-post-merge-closure.md` |
| `24B63533`, `6285779C`, `8D603E29` | Stage-side deferrals from cycle 012-S | n/a — no Ship capture refs to reconcile | — |

**Review-thread IDs remain a terminal, truthful `N/A`** for all of the above: the 012-S
closure record documents `0` unresolved threads with all findings raised **pre-PR** by
local persona reviewers, so no thread ever existed to recover. This is recorded, not
silently left blank. No concrete identifier was overwritten; reconciliation is idempotent.

**Non-blocking**: no missing identifier gated deliberation, planning, or harvest.

---

## 6. Re-evaluation of `EF9352FB` (operator rule 6)

Independently re-verified this cycle by **metadata-only** checks. **No key material was
read, printed, copied, or committed.**

- `git log --all -- .env.local` → **no commits**; the file was never tracked.
- `git check-ignore -v .env.local` → ignored via `.gitignore:36`, pattern `.env.*`.
- `git ls-files -- .env*` → **empty** at merged `HEAD`.

**Repo-side containment is CLOSED. There is no in-repo remediation to harvest and no code
change to make.** The operator's belief that this is very low priority *for the repository*
is **confirmed by concrete evidence**. Per operator rule 6 it is **de-prioritised
`high` → `low` and deferred**; it is deliberately placed in **no** shipment.

The sole residual action — rotating a live Tavily credential in an external provider
console — is **operator-only**. Stage does not hold that authority and must not simulate
it. **Stage did not and will not rotate the credential.** The untracked plaintext file
remains on operator disk (existence checked only), so that exposure persists until the
operator rotates it. Reported as a standing operator advisory, **not** a blocker on this
cycle, because it lies entirely outside the selected shipment boundary and has zero
in-repo surface.

---

## 7. Deferred candidates and rationale

| Entry | Pri | Deferred because |
|---|---|---|
| `A92E3FA0` | high | Governing product direction — a multi-shipment epic (TUI + SPA + Dev Tunnel broker). Cannot fit one shipment; under rule 1 reliability work precedes feature build-out. |
| `EF9352FB` | high→**low** | Repo containment closed; operator-only credential rotation. See §6. |
| `6C24E2E4` | low | Cluster-B structural refactor; exclusion-boundary risk. See §4. |
| `BF5DE670`, `E428AB46`, `37FAB8C2`, `E4C5413F`, `1C6C3B46` | med/low | `internal/pathsafe` cluster — a separate coherent group and the natural next cycle. |
| `2787DA56`, `4C5BEC23` | low | `check-unignore-regression.sh` cluster; both self-marked LOW confidence / unverified. |
| `F47DB9A9` | low | `ci.yml` topology-check comment accuracy — docs-grade, different job from `F4F4A959`. |
| `4104AF54` | medium | golangci-lint cannot lint windows-tagged Go — real reliability gap but a distinct CI-matrix concern, not gate detection. Strong candidate for a following cycle. |
| `4A01C53E`, `9D45E62E` | med/low | Closure/stash bookkeeping hygiene — rule 3 places documentation below features. |

None of the deferred entries is blocked by this cycle, and this cycle's diff does not touch
any of their surfaces.

---

## 8. Open questions (non-blocking)

1. Should `internal/**/*_test.go` be brought into scan scope? Raised by `8D603E29`,
   deliberately **not** decided here — it is a coverage expansion with its own
   false-positive budget. Recorded for a future cycle.
[WITHDRAWN by rev-3 addendum] Should the `RETIRED_ARCH_GATE_REQUIRED` toggle be flipped to required once the
   vocabulary fixes have soaked? Explicit operator decision, out of dark-mode authority.
3. Should `go.mod` / `go.sum` enter the pathspec (`09CD6ACF` item 2)? Deferred with D3.

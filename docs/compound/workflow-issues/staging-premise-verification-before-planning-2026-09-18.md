---
title: "Verify every premise against the artifact before planning on it"
date: 2026-09-18
category: workflow-issues
agent: Stage
cycle: chore/stage-ship-pipeline-contract-repair
base_commit: 14d44e3c1f28b321e8db24ab8cafcba346996133
triggers:
  - "a plan-review gate returns FAIL with a finding that refutes a measurement, not a design"
  - "two consecutive gates produce a finding worded 'recurs' / 'same class as' / 'relocated' / 'partially closed'"
  - "an exploration subagent's findings are carried into a plan without independent re-measurement"
  - "an acceptance criterion would pass at the parent commit, before the change it certifies"
discharges:
  - "O-3 (cross-artifact contract-surface matrix discipline), recorded overdue against 026-F"
  - "the STOP RULE of cross-artifact-contract-closure-requires-every-surface-2026-09-13.md"
---

# Verify every premise against the artifact before planning on it

## What happened

A 22-entry autonomous staging cycle produced three implementation plans. Plan review **FAILED all
three on attempt 1**, and the findings were overwhelmingly not design disagreements — they were
**refuted factual premises**. Attempt 2 returned ADVISORY / ADVISORY / FAIL. Two of the three plans
needed four revisions to become shippable, and one unit had to be withdrawn from the cycle entirely.

The cost was not the re-planning. The cost was that **the original decomposition was ranked,
ordered, and split using measurements that were wrong**, so the ranking argument itself had to be
re-derived after the fact.

## The five failure modes, each with its live instance

### 1. Inherited measurement — a subagent's finding used as a premise

An exploration subagent reported a scan-scope drift in `check-retired-architecture.sh`: the repo
scan supposedly globbed `internal/**` only, silently excluding `cmd/**`. An entire shipment unit was
justified by it.

**Refuted.** `select_repo_paths` (`:823–831`) *does* enumerate `cmd/**`. The `internal/**`-only glob
at `:836–837` belongs to `expected_internal_repo_paths` (`:834–852`) — a **self-test oracle** that
never runs on the repo-scan path. The subagent had matched the right text in the wrong function.

The unit survived, but its character changed completely: from *fixing a live coverage hole* to
*removing a maintenance hazard*. Those rank differently, and the original ranking was therefore
unsound.

> **Rule.** A subagent's measurement is a **lead**, never a premise. Re-open the file and read the
> enclosing function before any plan depends on it. Line numbers without enclosing scope are the
> single most reliable source of this error.

### 2. Slot confusion — reading a location, not a role

`_ship.agent.md:440` was described as an empty *evidence format contract*. It is in fact the
**command slot of quality gate #2** (`:439` = `go vet ./...`, `:441` = `go test ./...`). The stash
entry's ask was real; its characterization was not.

> **Rule.** Before asserting a line is empty/missing/wrong, read its **two neighbours**. Structure is
> usually legible from adjacency alone.

### 3. Severity inflation — "fails open" asserted without tracing control flow

`check-unignore-regression.sh` was described as failing open on an unresolvable base ref.

**Refuted, in both directions.** It does *not* fail open — `run_differential_check` raises
`SystemExit` (~`:293–301`) *before* the baseline read (~`:316`), and `run_denylist_check` only ever
passes `ref="HEAD"`. The defect is **latent defence-in-depth**. But the opposite error also appeared:
the checker was separately assumed advisory when it is in fact **merge-blocking** (`ci.yml:420`, no
`continue-on-error`, inside `ci-gate.needs`).

> **Rule.** "Fails open" and "is advisory" are both **control-flow claims**. Trace guard-to-use
> ordering, and trace the CI wiring to the required check. Never infer either from a script's prose.

### 4. Green-on-arrival acceptance criteria

Several criteria would have passed **at the parent commit**, certifying nothing:

* asserting two derivations agree, when they already agreed;
* a pinned assertion degraded into `assert SCAN_SCOPE contains "cmd/**"` — a **tautological
  self-comparison** replacing a check that deliberately read *source text* precisely so something
  outside the process could disagree;
* an entire shipment (`S-10`) whose every task was already physically satisfied;
* a test needle already satisfied by an unrelated existing step.

> **Rule.** For each acceptance criterion ask: *does this fail at the parent commit?* If not, it is
> characterization — **label it as such**. A green-on-arrival criterion presented as verification is
> worse than no criterion, because it manufactures false confidence. Pin against the change's **own
> parent commit**, never a fixed snapshot, which fails spuriously on unrelated drift.

### 5. Frozen-vs-forbidden contradiction

One plan froze a clause range (`:832–835`) *and* required rerouting the behaviour that clause
governs. Attempt 2's reviewer called this *"the same defect class as attempt-1's P0-2, now
under-constrained"* — the wording that fires the STOP RULE.

> **Rule.** A freeze and a behaviour change over the same surface must be reconciled **in the plan**,
> with the exact reconciled wording written out. Deferring it to implementation guarantees the
> defect relocates rather than closes.

## The generalisable rule

**Build the contract-surface matrix before the first edit, not after the first FAIL.**

| Surface | Current evidence (file:line, *in its enclosing function*) | Intended change | Executable verification | Depends on |

Any row that cannot be filled from a **direct read** is not a premise — it is a question. Send it to
a spike, or state it as an open question. Do not let it become a plan's justification.

## Corollaries earned this cycle

* **Two consecutive gates naming a recurrence ⇒ stop gating.** The budget is two re-entry cycles;
  a third attempt is a P-013.6 escalation. When the STOP RULE fires, **adopt the reviewer's own
  disposition** — here, *"ship the sound units, hold the rest"* — rather than re-planning wholesale.
  Attempt 2's reviewer produced a better scope decision than a third attempt would have.
* **A unit whose every task is green-on-arrival is not a release unit.** Fold it into a real one.
* **Prefer withdrawal to shipping a wrong premise.** `3289EB69` was held, not shipped, because its
  gap list pointed at the wrong file. A held item with a written unblocking condition is honest;
  a shipped item full of green assertions is not.
* **Probe the tool, not the prose.** `--complexity` is documented and accepted by the CLI, but
  `artifact type "task" does not define a complexity field` — the error only surfaced because the
  field was **read back** after writing. A prior memory note claiming it worked was simply wrong.

> **Meta-rule: read back every mutation.** Three separate silent failures this cycle — the
> `--complexity` no-op, a `--section` flag silently dropped for tasks, and a PowerShell
> case-insensitive `-match 'Error'` matching the word *"error-surface"* inside an entry's own text
> and reporting two successful writes as failures — were all invisible until the artifact was
> re-read. Write, then read back, then believe.

---
title: "Plan — TASK_ONLY_FINALIZE classifier, guards and recovery in shipment-reconcile (Skill implementation)"
date: 2026-09-12
status: planned
agent: Stage
revision: 2
feature: 024-F
task: 024.001-T
shipment: 023-S
depends_on_shipment: 022-S
deliberation: docs/decisions/2026-09-12-intercom-go-three-shipment-bootstrap-deliberation.md
umbrella_evidence: docs/plans/2026-09-12-intercom-go-task-only-shipment-finalization-plan.md
evidence_dir: docs/plans/evidence/2026-09-12-task-only-shipment-finalization/
---

# Plan — TASK_ONLY_FINALIZE classifier, guards and recovery in shipment-reconcile

> **Requires plan hardening**: **yes — applied in rev 2** (see `## Plan Hardening`).

- **Shipment**: `023-S` (C, Skill implementation) — third of three
- **Depends on**: `022-S` (B), which depends on `021-S` (A)
- **Closes by**: the **existing** `CASCADE` exception — **rev 2, forced by measurement**
- **Manifest**: **fully-covered root** `[024.001-T, 024-F]`
- **Activates**: `TASK_ONLY_FINALIZE`, whose **first real exercise is `017-S`**

> **Rev 2 — why the manifest shape changed.** Rev 1 kept `023-S` task-only so C would
> close by the path it ships. Probes 22/23/24 falsified that design:
> **Probe 22** measured that claiming a task-only shipment transitions its
> **out-of-manifest** covering feature to `active`, and that the feature **stays `active`**
> after the task-only close — so `024-F` would become a live top-level release unit and
> trip Ship Step 1's P-001 `Active` gate when `017-S` is routed.
> **Probe 23** measured that the natural remedy — a `blocks` edge from `017-S` to `024-F` —
> is **not constructible**: the engine refuses it (`both endpoints must be shipments`).
> **No dependency-graph barrier against a feature exists at all.**
> **Probe 24** measured the remedy adopted here: with `024-F` **inside** the manifest the
> engine's own cascade archives it (`archived_ids: [task, feature, shipment]`), leaving
> **zero** other active top-level units and `017-S` eligible with **no** barrier and **no**
> Stage disposition step. The hazard is removed at its source rather than gated around.
> The self-proof is **relocated, not lost** — see §7.3 Probe 21 and §8.4.

## 1. Objective

Implement, in `shipment-reconcile/SKILL.md`, the procedure that B's dormant policy branch
authorizes:

1. the **capability/schema token** (activating B's gate);
2. the **narrow classifier/selector** for `TASK_ONLY_FINALIZE`;
3. the **`G0–G13` pre-call guards** and **`P1–P10` post-call guards**;
4. **CLI digest-bound invocation**;
5. the **two-path failure split** and **bounded recovery**;
6. the **self-host ordering**.

On merge, A's generic delegation + B's dormant authorization **activate together**. The
activated path's **first real exercise is `017-S`**, the 12-member task-only shipment this
whole chain exists to make closable. C itself closes by the **pre-existing `CASCADE`**
exception (§8), exactly as A and B do.

## 2. Surface — exactly 2 files

| # | File | Change |
|---|---|---|
| 1 | `.github/skills/shipment-reconcile/SKILL.md` | 16 clause sites + `B11` + `B18` + token in frontmatter |
| 2 | `tests/integration/taskonly_finalize_contract_test.go` | one table-driven test fn + ≤4 helpers |

**New production code files: 0.**

**Test-file decision (explicit).** A **separate** test file is used rather than extending
an existing integration file. Rationale: task isolation stays unambiguous, the H0 red
count is trivially `1 of 1`, and the file is the natural home for the token-parity
assertion against P-015. Still exactly two files.

**Evidence (§7.3) is not a third surface.** `docs/plans/evidence/…` is an evidence
directory, as rev 6 §6.1-E established. A and B may **cite** shared observations from this
directory; **C owns it.**

**Rev 2 — no committed `clause-inventory.md` is required.** Rev 1 charged C with
regenerating and committing an inventory file. That requirement is **removed**: the exact
inventory already lives in **§3 of this committed plan** and in `024.001-T`'s acceptance
criteria, both of which are on the branch and under review. A second copy would be a
duplicate system of record, and duplication is precisely the defect class A exists to
remove. Any runtime **Probe 21** transcript is **closure/runtime evidence** and is budgeted
under **Ship closure**, not under task coding (§12).

### 2.1 Token feasibility — falsified in advance

Plan B hardening 6 flagged "the skill may be unable to advertise a machine-readable
token" as C's discovery risk. **Resolved during planning**: `SKILL.md` already carries
YAML frontmatter (`name`, `description`) at lines 1–4, so a `capabilities:` key is
directly addable. Risk closes **low** before C starts.

### 2.2 Execution boundary — the skill markdown IS the classifier here (verified)

`SKILL.md` names `src/autoharness/gates/shipment_closure.py` as *"this self-hosting
repository's own `classify_shipment_close_path` implementation"* at **line 413** (inside
`B9`) and **line 1075** (`B17`). That is template text inherited from the autoharness
source repository.

**Verified at `e7e981d`: `intercom-go` has NO `src/` directory and no executable close-path
classifier.** `.autoharness/gates` holds only a log; Ship Step 6 invokes only
`autoharness gate pipeline-topology`. Classification in this workspace is therefore
**agent-driven from the skill markdown**, which is what makes the "0 source changes" claim
true and what makes C's markdown edits actually activate the path.

Both stale references fall **inside C's existing inventory** (`B9` at 413, `B17` at 1075)
and are reconciled there — scoping them to *"workspaces with a Python implementation
installed; in this workspace the skill text is authoritative"*. **No new clause site** (the
count stays 16); `B17`'s budget is corrected from ~1.5 to ~3 min in §12, because it is two
lines rather than one (§3).

Had the executable classifier actually been present, C's markdown-only edits would **not**
activate the path and §8's runtime proof would fail. Recording the verification is what
distinguishes a checked assumption from a lucky one. Asserted by **AC-16**.

## 3. Exact inventory — 16 clause sites + 2 additive

Line anchors from rev 6 Inventory Table §6.1-B; frontmatter and lines 10–32 spot-verified
against the installed file at `e7e981d`.

| # | Line | Clause | Class |
|---|---|---|---|
| **B1** | 3 | frontmatter *"**except for** the narrow, machine-verified P-015 fully-covered-root case"* | **MUST** |
| **B2** | 10–11 | *"run `mode: safe-close` **in place of** the destructive cascade"* | CONSISTENCY |
| **B3** | 14–15 | *"runs the P-015 … classification and, **only when** every precondition holds, delegates to the Cascade Close Sub-Procedure instead"* | **MUST** |
| **B4** | 28 | *"**instead of** the cascade … call, then `mode: post`"* | CONSISTENCY |
| **B5** | 30–32 | *"never calls the cascade op directly **unless** that classification confirms the narrow verified fully-covered-root exception"* | **MUST** |
| **B6** | 232–233 | *"It **never calls** the cascade `backlogit_ship_shipment`"* | CONSISTENCY |
| **B7** | 324 | pre-mode `RECONCILE_FAIL`: *"Do NOT call `backlogit_ship_shipment`"* | CONSISTENCY |
| **B8** | 356–358 | Safe-Close preamble: *"… Step 0 below, where the cascade op is **the** permitted close path"* | **MUST** |
| **B9** | 408–415 | **Step 0(c) "Classify the close path"** — classifier body | **MUST** |
| **B10** | 505–512 | **Step 0 close-path SELECTOR** — binary → **ternary** | **MUST** |
| **B12** | 618–620 | Cascade Sub-Procedure heading *"(… exception **ONLY**)"* | CONSISTENCY |
| **B13** | 698–709 | No-substitution rule — must also bind the **third** verdict | **MUST** |
| **B14** | 998 | lock-failure: *"Do NOT call `backlogit_ship_shipment`"* | CONSISTENCY |
| **B15** | 1015–1016 | Safety invariants — must enumerate the new path | **MUST** |
| **B16** | 1022 | *"… **never via the cascade op**"* | CONSISTENCY |
| **B17** | 1074–1075 | **two adjacent reference-list lines**: (a) 1074 names P-015 as *"cascade prohibition"* — must also name the amended authorization; (b) 1075 names `src/autoharness/gates/shipment_closure.py` as *"this self-hosting repository's own `classify_shipment_close_path` implementation"* — a stale inherited reference (§2.2) that must be scoped | **MUST** |

| # | Site | Content |
|---|---|---|
| **B11** | 591–605 | Safe-Close **step 8** — fail-closed cross-reference; step 8's own text otherwise unchanged |
| **B18** | new section | **NEW** `TASK_ONLY_FINALIZE` section: token, classifier, `G0–G13`, `P1–P10`, failure split, bounded recovery, self-host ordering |

**16 clause sites (9 MUST + 7 CONSISTENCY) + `B11` + `B18`.** This is C's entire share.

**`B17` classification.** Rev 6 classes `B17` CONSISTENCY in Table §6.1-B but budgets it
as MUST-mechanical in §10.1. This plan follows **§10.1 (MUST)** — the stricter reading —
consistent with the deliberation §6 resolution.

**Rev 2 — `B17` is two lines, not one.** Rev 1 anchored `B17` at line 1074 only, while
§2.2 separately asserted that the stale `src/autoharness/...` reference at line **1075**
would be "reconciled inside `B17`". Those two statements were inconsistent: 1075 is a
**distinct line requiring its own edit**, and an inventory that does not name it cannot be
used to verify completion. `B17` is therefore re-anchored to **1074–1075** as an explicit
two-part site. The site **count** is unchanged (16) — this adds no new clause — but the
budget is corrected honestly in §12 (`B17` ~1.5 → ~3 min).

### 3.1 Partition check

`A` takes `C1–C6` (6). `B` takes `A1–A6` (6). `C` takes `B1–B10`, `B12–B17` (16).
**6 + 6 + 16 = 28.** Exact. No duplicated obligation; no missing site.

## 4. What B18 contains

| Part | Content |
|---|---|
| **Token** | `capabilities:` frontmatter key advertising the exact-version token P-015 A7 requires (**CO-1**) |
| **Classifier** | Evaluates after `CASCADE`, before `SAFE_CLOSE`. Any failure/ambiguity/query error → not classified → falls through |
| **Selector** | Step 0 selector becomes ternary (`B10`) |
| **`G0–G13`** | Conjunctive, fail-closed pre-call guards |
| **`P1–P10`** | Post-call guards on the **RAW post-call state, BEFORE any archive restoration** (P-007 ordering, umbrella §8.2) |
| **Failure split** | Two separate paths (umbrella §6.5) |
| **Bounded recovery** | Enumerated-set-only restore; approval before any destructive step; byte-prefix append-only preservation; out-of-bounds paths reported-not-moved; sync-only rehydration (Probe 19) |
| **Digest binding** | §5 |
| **Self-host ordering** | §8 — including the **token-activation rule** of §8.1: a session MUST NOT select a verdict whose authorizing token it merged in that same session (Go test row 13) |

T1–T5 are **not** restated — they are B's surface. B18 **cites** P-015.

## 5. Engine digest binding (unchanged from rev 6)

Engine identity is **binding, not advisory**:

```text
backlogit 1.10.1-0.20260823032255-b07729386a31+dirty
SHA-256 1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98
```

`+dirty` is not a unique build identity, so **the digest binds**. **Mismatch → HALT +
full probe refresh.** The rev-4.1 advisory fallback stays **withdrawn**.

Invocation is **CLI-only**. No MCP equivalence is claimed; the registry advertises MCP
tools, and that claim is explicitly **not** made.

## 6. Ordering (test-first; `main` stays green)

1. **H0 — red.** One table-driven test function. Marker:
   `not implemented: task-only finalize contract`. Red count **1 of 1**; `go vet` clean;
   `go test ./...` non-zero.
2. **H1 — green.** Apply `B1–B18`. Re-run: `go vet` clean, `go test ./...` green,
   `markdownlint` clean.

Exactly **one** generated test function — the `021.001-T` constraint.

## 7. Verification obligations

### 7.1 What the Go test proves — text, parity, tokens

**13 rows** over `SKILL.md` (+ P-015 for parity):

| # | Assertion | Kind |
|---|---|---|
| 1 | Frontmatter advertises the capability token | positive |
| 2 | **Token parity**: the token string in `SKILL.md` **exactly equals** the one P-015 A7 requires | **parity (CO-1)** |
| 3 | Step 0 selector is **ternary**, ordered `CASCADE` → `TASK_ONLY_FINALIZE` → `SAFE_CLOSE` | **ordering** |
| 4 | `B18` states `G0`–`G13` | positive |
| 5 | `B18` states `P1`–`P10` | positive |
| 6 | `P1`–`P10` are stated to evaluate on the **raw post-call state before archive restoration** | **ordering** |
| 7 | `B18` states the two-path failure split | positive |
| 8 | `B18` states bounded recovery (enumerated-set-only; approval-before-destructive) | positive |
| 9 | `B18` states the digest binding and CLI-only invocation | positive |
| 10 | `B11` step-8 fail-closed cross-reference present; step 8's own text otherwise unchanged | **negative** |
| 11 | The **existing** fully-covered-root exception's skill-side handling is unmodified | **negative** |
| 12 | `SKILL.md` does **not** restate T1–T5 as its own authority (cites P-015 instead) | **negative** |
| 13 | `B18` states the **self-host ordering rule**: a session MUST NOT select a verdict whose authorizing capability/schema token was merged by that same session (§8.1) | **ordering/consistency** |

Rows 10–12 are the widening guards. Row 2 discharges **CO-1** from C's side. **Row 13 is
the bootstrap guard** — it fails if C ships without the rule that forces its own closure
into a second session.

### 7.2 What the Go test does **not** prove

It proves **contract text, parity and tokens**. It does **not** prove runtime behavior:
not that the classifier selects correctly, not that guards hold against the live engine,
not that recovery restores correctly. Those are **§7.3 durable runtime probes**. The
separation is stated exactly so no acceptance criterion conflates them.

### 7.3 Durable runtime probes — behavior

Retained from rev 6 and extended in rev 2, committed and reproducible under
`docs/plans/evidence/2026-09-12-task-only-shipment-finalization/`:

| Probe | Proves | Status |
|---|---|---|
| **14** | dependency lifecycle end-to-end (shipment→shipment `blocks`) | committed, reused |
| **15** | EXACT `017-S` 12-member fixture, `INVARIANCE_FAILURES=0` | committed, reused |
| **16** | injected partial failure + bounded recovery | **WITHDRAWN as evidence** — see below |
| **18** | live-config inbound-edge byte-invariance; `MOVE_SHIPPED_EXIT=9` | committed, reused |
| **19** | **bounded recovery v2, both approval branches — AUTHORITATIVE** | committed, reused |
| **20** | post-archive closure route (O-6) + **post-archive lifecycle-gate tripwire** | committed, reused |
| **21** | activation probe — **relocated to closure/runtime evidence** (§8.4) | Ship closure |
| **22** | **NEW** — covering-feature status at claim and after task-only close | **executed at `7bc0318`** |
| **23** | **NEW** — shipment→feature dependency barrier is **not constructible** | **executed at `7bc0318`** |
| **24** | **NEW** — fully-covered-root re-shape disposes the feature mechanically | **executed at `7bc0318`** |

Rev 4.1's `%TEMP%`-and-deleted probes 5/5b/10/12/13 remain **WITHDRAWN** as evidence.

**Probe 16 is withdrawn; Probe 19 is the authoritative recovery evidence.** Rev 1 listed
Probe 16 as "committed, reused" and Hardening 4 claimed "Probes 16 **and** 19 executed
against both recovery branches". Both statements are corrected. The umbrella plan already
records (§ Probe table, and the rev-6 probe-obligation note) that Probe 16 **diverged from
the specified procedure in five ways** and that the claim it proved "the exact procedure"
is **withdrawn**. Probe 16 is retained **for history only** and MUST NOT be cited as
evidence for the bounded-recovery procedure. **Probe 19** implements the procedure as
specified and executes **both** the approval-granted and approval-withheld branches; it is
the sole authority for `B18`'s recovery content.

**Probe 20's post-archive lifecycle-gate tripwire is load-bearing for §8.** Probe 20
measured that a `--phase lifecycle` topology-gate invocation fails closed once the shipment
is archived: on the two routes that are genuinely post-archive lifecycle invocations
(`--mode agent` and `--mode manual`, each with `--shipment <archived> --phase lifecycle`)
the gate exits **1** with `LIFECYCLE_NO_ACTIVE_SHIPMENT`. *(Stated precisely: the probe
tabulates four routes and records `POST_ARCHIVE_ROUTES_NONZERO_EXIT=4`, but the other two
exit **2** on argument validation — `agent mode requires --shipment` and `--phase lifecycle
requires --shipment` — and Probe 17 records the identical exit-2 tokens against a queued,
unclaimed shipment. Those two are archival-independent and one is not a lifecycle-phase
route at all. The tripwire rests on the two valid routes; the "all four routes" framing
does not, and is not relied on here.)* The consequence for every closure sequence in this
chain: the `a0` lifecycle gate runs **while the shipment is still active**, immediately
before the close; **no** `--phase lifecycle` invocation may occur after the archive. Probe
20 step 5 further cites the installed contract directly: Ship Step 6.0 invokes
`pr-lifecycle` for the closure PR **without** a lifecycle gate, so the closure PR is
reachable post-archive. This tripwire is carried into §8 and into the deliberation's
coupling matrix as **CO-12**.

**Probe 21 (activation) — relocated, and strengthened.** Rev 1 discharged Probe 21's
post-half by C's own task-only self-closure. Rev 2's re-shape removes that self-closure, so
Probe 21 is **relocated to closure/runtime evidence** (§8.4) and is budgeted under **Ship
closure**, not task coding (§12). This is a strengthening, not a loss: the relocated probe
exercises the classifier against the **exact `017-S` 12-member fixture** (Probe 15's
fixture) rather than C's trivial one-task shape, so it tests the shape the path actually
has to serve.

**Rev 2 probes 22/23/24 — what they measured, exactly.**

| Probe | Measurement | Result |
|---|---|---|
| **22** | ARM C: task-only manifest, feature outside. Feature status pre-claim / after claim / after task-only close | `queued` → **`active`** → **`active`** |
| **22** | ARM AB: fully-covered-root manifest, feature is a member. Feature status after claim | **`active`** (discharges A's PO-1a and B's §8 step 1) |
| **23** | `backlogit dep add 002-S 001-F --type blocks` | **REFUSED**: *"prerequisite 001-F has type `feature`; both endpoints must be shipments"* (`Q0_FEATURE_ENDPOINT_EDGE_CONSTRUCTIBLE=False`). With the predecessor shipment archived and the feature still `active`, `ready_set=[002-S]` — the successor is **already eligible**, with nothing suppressing it |
| **24** | ARM Q: cascade over `[feature, task]` | `archived_ids: [task, feature, shipment]`; **0** other active top-level units; successor eligible with **no** barrier |
| **24** | ARM QR: **the exact live construction** — task-only manifest, then `shipment add` the feature | `MANIFEST_IS_TASK_FIRST=True`; same result as ARM Q; `ORDER_INDEPENDENCE_MEASURED=True` |
| **24** | ARM T (control): identical topology, task-only | feature left **`active`**, **not** archived — difference attributable to manifest shape alone |

All three ran CLI-only under `DIGEST_GATE=PASS`, seeded from the **live** `.backlogit`
control files and the **live** `.autoharness` config, in a disposable gitignored
workspace-contained scratch directory that was removed after the transcript was captured.
The live backlog was never mutated by a probe.

## 8. Self-hosting closure sequence — two sessions, no mid-session authority gain

`023-S` manifest is the **fully-covered root** `[024.001-T, 024-F]` (rev 2). `024-F` is a
root (no `parent_id`), its complete descendant set at every depth is exactly
`{024.001-T}`, and the manifest contains nothing else. C therefore closes by the
**existing, already-reviewed** `CASCADE` exception — the same path A and B use.

**C's merge ACTIVATES the `TASK_ONLY_FINALIZE` token. The session that performs that merge
MUST NOT use it.** This is the same property A establishes for Role Boundary changes,
applied to **authorization-token activation**. C is therefore a **two-session** close.

*Manifest item order note:* the live record is `[024.001-T, 024-F]` (the feature was
appended to an existing task-only manifest; the engine exposes no manifest re-ordering
operation). Order is **immaterial, and this was measured rather than assumed** — **Probe 24
ARM QR** reproduces the live construction exactly (create task-only, then
`backlogit shipment add`), confirms `MANIFEST_IS_TASK_FIRST=True`, and then closes it:
`archived_ids: [task, feature, shipment]`, `QR_FEATURE_ARCHIVED_BY_CASCADE=True`,
`QR_SUCCESSOR_ELIGIBLE=True`, `ORDER_INDEPENDENCE_MEASURED=True`. ARM Q had built the
manifest feature-first, so ARM QR exists specifically because task-first was the one order
the remedy had **not** been probed in — and it is the order the live re-shape produced.
P-015 classification is over set membership, and Ship Step 0.5.1a filters `items` by
artifact type, so neither depends on order either.

| Step | Session | Action |
|---|---|---|
| 1 | S1 | Orchestrator routes `023-S` **only after** `022-S` is archived `shipped`; Step 0.5 pre-claim topology gate; **claim** (`024.001-T -> active`, and — measured, Probe 22 — `024-F -> active` as an **observed claim effect**) |
| 2 | S1 | **Probe 21 (pre)**: with the token still absent, the classifier does **not** select `TASK_ONLY_FINALIZE`. **Then remove the scratch workspace and re-run `backlogit sync`, confirming zero `duplicate source id` warnings, before any backlog mutation** (§8.4) |
| 3 | S1 | H0 red → implement `B1–B18` → H1 green |
| 4 | S1 | **Step 4.3 quality gates; Step 4.4 review gate** |
| 5 | S1 | **Step 4.5 Complete Task — commit, then `024.001-T -> done`.** *(Installed order: Step 4.5 precedes Step 5. The task reaches `done` BEFORE the implementation PR merges.)* |
| 6 | S1 | **Step 5 PR lifecycle** — full gate sequence, `--phase lifecycle` topology gate, build, push, PR, operator approval, **merge** |
| 7 | S1 | **Step 6 Merge Confirmation Gate** — `gh pr view` state `MERGED`; `git fetch origin main`; `git merge-base --is-ancestor {merge_sha} origin/main` |
| 8 | S1 | **Mandatory pre-self-close reload** of merged `main`: Ship instructions, `shipment-reconcile`, **and P-015** (A's `C1`). Positively verify the merged **tokens** and the **merge commit** |
| 9 | S1 | **STOP. The reload just made this session's own merge the source of a new authorization token.** Per §8.1, S1 MUST NOT classify or close. Write a checkpoint and **END** |
| 10 | **S2** | **Fresh session.** Full checkpoint-recovery protocol (§8.2) — enumeration, anomaly gate, explicit owner selection, operator confirmation, resolve-after-resume |
| 11 | S2 | Re-verify: current `main`; the merge SHA is an ancestor of `origin/main`; the **merged P-015 token**, the **merged skill token** and the **merged Ship delegation** are all present and mutually consistent |
| 12 | S2 | **Step 6.0 Post-Merge Branch Protocol** — `git checkout main`, `git pull`, `git checkout -b post-merge/024-taskonly-finalize`. **This branch is created BEFORE any post-merge backlog mutation**, because Step 6.1(e) commits `.backlogit/` and those commits must not land on `main` |
| 13 | S2 | **Step 6.1(a0)** `--phase lifecycle` topology gate — run **while `023-S` is still `active`**. *(Probe 20 tripwire: after the archive this gate fails closed on every route. It must never be re-run post-archive.)* |
| 14 | S2 | **Step 6.1(a1)** covering-feature completion — `024-F` **is** a manifest member, so a1 **applies**: verify A §5 conditions 1–5, then `024-F active -> done` |
| 15 | S2 | **Step 6.1(a)** pre-mode, `expected_status: done` — both members `done` → `PROCEED` |
| 16 | S2 | Classification → **`CASCADE`** (fully-covered root). `TASK_ONLY_FINALIZE` is **not** selected here (T3: the manifest has a live feature member). Digest check (§5); mismatch ⇒ **HALT** |
| 17 | S2 | **Step 6.1(b)** cascade close; **6.1(c)** P-007 archive-integrity verify; **6.1(d)** post-mode; **6.1(e)** commit `.backlogit/` **on the closure branch** |
| 18 | S2 | **Step 6 item 2** `operational-closure mode=post-merge` → `docs/closure/`; **P-020** compact-context finalizes the compaction status — **all on the closure branch, before the closure PR is pushed** |
| 19 | S2 | **Probe 21 (post)** — §8.4 activation evidence, against the exact `017-S` 12-member fixture. **Then remove the scratch workspace and re-run `backlogit sync`, confirming zero `duplicate source id` warnings, BEFORE step 20's closure sync** (§8.4) |
| 20 | S2 | **Step 6 item 9** `backlogit sync` (closure index resync), then push the closure branch, closure PR, P-014 local review, operator approval, merge |
| 21 | S2 | **Step 6 item 10** return to `main`; `git pull` |
| 22 | — | Orchestrator routes **`017-S`**. **No Stage disposition step is required** — the cascade archived `024-F` (Probe 24) |

**Why the order in steps 5–6 and 12–20 is exactly this.** It is read from the **installed**
`_ship.agent.md`, not inferred:

| Installed anchor | Consequence for this sequence |
|---|---|
| Step 4.5 (L517) precedes Step 5 (L554) | task `-> done` **before** the implementation PR merges (step 5) |
| Step 6.0 (L726) item 2–3: branch created from fresh `main`; *"All subsequent Step 6 work happens on this branch"* | closure branch exists **before** the 6.1(e) backlog commit (step 12) |
| Step 6.1(a0) *"before the pre-archive reconciliation gate … the shipment-scoped check immediately preceding the safe-close mutation itself"* | lifecycle gate runs while `023-S` is still `active` (step 13) |
| Step 6.1(a)→(b)→(c)→(d)→(e) | pre-mode → close → P-007 → post-mode → commit (step 17) |
| Step 6 item 2 (`operational-closure`) and item 8 (P-020) precede item 9 (`sync`) and the 6.0 item 4 push | closure artifacts + P-020 land on the closure branch **before** the closure PR is pushed (steps 18–20) |

Rev 1 placed `024.001-T -> done` **after** the merge and placed operational closure + P-020
**after** the close without establishing the branch first. Both contradicted the installed
contract. Corrected above.

**One deliberate deviation from the installed text's own order, stated rather than
glossed.** In the installed file the *"Mandatory pre-self-close context reload"* paragraph
sits at **L756–764**, i.e. **after** Step 6.0 items 1–5 including the branch creation at
L734–737. This plan places the reload at **step 8**, before Step 6.0 (step 12). That is
deliberate and necessary: §8.1 requires S1 to **end** at the reload, so S1 must not have
created a closure branch it will never use, and S2 — a different session — is the one that
runs Step 6.0. The installed ordering assumes a single session performs both; C (like A)
does not. No installed requirement is skipped — the reload still happens after the merge
is confirmed and before any close action, which is the property L756–764 exists to
guarantee.

### 8.1 Why C requires TWO sessions (rev 2 — this replaces "C may close in one session")

Rev 1 argued C could close in one session because C changes **no Role Boundary**, and A's
authority-escalation rule binds **authority** changes only. That reasoning was **wrong in
its premise**: it treated "authority" as meaning only the Role Boundary table. C's merge
introduces the **capability/schema token** that flips B's dormant P-015 gate from
not-selectable to selectable. That is an **authorization envelope change** — the session
that merges it would be selecting a verdict that its own merge authorized.

A's §4.2 pinned wording deliberately scopes A's carve-out to **Role Boundary** changes, and
A's Go test row 11 asserts that scoping. Rev 2 does **not** widen A's carve-out and does
**not** disturb row 11. Instead the rule lands where it belongs — in **C's own `B18`
self-host ordering** — as a *procedure* rule, which A's §4.2 explicitly says takes effect
on reload exactly as before:

> A session MUST NOT select a close-path verdict whose authorizing capability/schema token
> was merged **by that same session**. On detecting that condition at the mandatory
> pre-self-close reload, the session writes a checkpoint and ends; a fresh session performs
> the classification and closure.

This is self-bootstrapping in the same way A is: S1 reloads merged `main` at step 8, reads
the rule it just merged, and obeys it by ending. Asserted by **AC-14** and Go test row 13.

**C gains no authority mid-session, and neither did A.** A proved the property for Role
Boundary; C proves it for authorization tokens. The two together are why `017-S` can trust
the path.

### 8.2 S2's checkpoint recovery is the FULL protocol, not a resume shortcut

S2 MUST execute the complete owner-selected, operator-confirmed recovery protocol. Naming
it explicitly because "restore the checkpoint and continue" is the failure mode:

1. **Enumeration** — `backlogit_list_checkpoints` with `consumer_id: "ship"` and **no**
   `status`/`agent` filter, so a quarantined or schema-invalid record cannot be silently
   excluded.
2. **Anomaly gate FIRST** — any validation error, quarantine flag, or missing/malformed
   required field in **any** enumerated summary ⇒ **FAIL CLOSED** to operator handoff
   before the zero-candidate check is even evaluated.
3. **Explicit owner selection** — never auto-pick, **even if exactly one candidate is
   returned**. The operator selects a single checkpoint by filename; ambiguity fails closed.
4. **Owner validation** — the CheckpointV1 `agent` field MUST be exactly `ship`. A
   `stage`-owned checkpoint is never selectable here (P-001 role separation).
5. **Operator confirmation before restore** — present `resume_hint` and recorded state;
   no automatic resume. Checkpoint schema V1 carries no heartbeat/lease, so age can never
   prove S1 dead.
6. **Same cursor** — resume the **same** single active shipment `023-S`; no parallel
   resume, no new worktree (P-001/P-016).
7. **Resolve AFTER resume** — `backlogit_resolve_checkpoint` is called **only after** S2
   confirms a successful resume, and **only** for that one selected checkpoint. No bulk
   sweep.
8. **Fail closed, no fresh-start fallback** — an invalid, torn, or unreadable checkpoint
   halts to the operator. S2 MUST NOT discard it and start fresh.

S1's checkpoint MUST be written through `backlogit_create_checkpoint` with
`schema_version: 1`, `agent: ship`, `phase: awaiting-fresh-session-activation-close`, and a
`resume_hint` naming `023-S`, `024-F`, `024.001-T` and the **merge SHA**. Domain data goes
under `context`, never at the top level.

### 8.3 No `024-F` disposition step is required (rev 2 — this replaces the Stage follow-up)

Rev 1 carried a mandatory Stage-owned disposition of `024-F` between C's close and `017-S`
routing, because a task-only close leaves the out-of-manifest root parent live. Probes
22/23/24 replaced that prose step with a mechanical property:

| Rev 1 claim | Rev 2 measurement |
|---|---|
| `024-F` is left live after a task-only close | **Confirmed** — Probe 22 ARM C and Probe 24 ARM T: feature stays **`active`** |
| A live `024-F` can trip Ship Step 1's P-001 `Active` gate and stall `017-S` | **Confirmed** — P-001 (L308) tests for status `Active` |
| A Stage disposition step before `017-S` discharges it | **Insufficient as a gate** — it is **prose-only**. Probe 23 proved no dependency-graph barrier against a feature is constructible (`both endpoints must be shipments`), so nothing mechanically prevents `017-S` from being routed first |
| — | **Remedy (Probe 24):** with `024-F` **inside** the manifest, the cascade archives it (`archived_ids: [024.001-T-analogue, 024-F-analogue, 023-S-analogue]`), leaving **zero** other active top-level units and `017-S` eligible with no barrier and no follow-up |

The P-001 hazard is therefore **eliminated at its source**, by the engine, inside C's own
closure — not gated around by an instruction someone has to remember. Asserted by
**AC-17**.

**Residual, stated honestly.** `017-S` itself is a task-only shipment whose covering
feature `018-F` sits outside its manifest, so `017-S`'s own close will leave `018-F`
`active` by the same mechanism. That is **not** a routing blocker: `017-S` is terminal in
this chain, so there is no downstream shipment for `018-F` to block. It is recorded as a
**Stage hygiene disposition** after `017-S` closes, with no gate attached to it. This is
the honest scope of the fix — C's residue is removed because it has a consumer; `017-S`'s
residue is recorded because it does not.

### 8.4 Probe 21 — activation evidence, budgeted under Ship closure

| Half | When | What it shows |
|---|---|---|
| **pre** | S1, before merge (step 2) | token absent ⇒ classifier does **not** select `TASK_ONLY_FINALIZE` |
| **post** | S2, at closure (step 19) | token present ⇒ classifier **selects** `TASK_ONLY_FINALIZE` for the **exact `017-S` 12-member task-only fixture** (Probe 15's fixture), and **does not** select it for C's own fully-covered-root manifest (T3) |

The post-half runs in a disposable, live-config-seeded, digest-bound scratch workspace —
never against the live backlog.

**Mandatory scratch discipline (NON-NEGOTIABLE).** Both halves MUST, immediately after the
transcript is captured: (1) remove the scratch directory under
`.backlogit/runtime/stage-probe-scratch/`, and (2) re-run `backlogit sync` and confirm
**zero** `duplicate source id` warnings — **before** any live backlog state is read,
mutated, or reported. This is the evidence directory's own binding rule (evidence
`README.md` §3), and it has already fired once for real: residual probe scratch was
rehydrated into the live index and surfaced a phantom `002-S "Successor blocked shipment"`
in `backlogit shipment list`. A **12-member `017-S` analogue fixture** is the single worst
artifact to leave rehydratable, and the post-half sits immediately before the closure
`backlogit sync` at §8 step 20. Asserted by **AC-23**.

**Budget (charged to Ship closure, not to `024.001-T`).**

| Item | Session | Estimate |
|---|---|---|
| Probe 21 **pre**-half — seed, run, capture, scratch-remove + sync-verify | S1 | **~8 min** |
| Probe 21 **post**-half — seed the 12-member fixture, run, capture, scratch-remove + sync-verify | S2 | **~12 min** |
| S1 checkpoint write + S2 full recovery protocol (§8.2) | S1 + S2 | **~10 min** |
| **Total closure-evidence budget** | — | **~30 min** |

This is **separate from** the ~105 min implementation budget in §12 and is **not** subject
to the 2-hour task rule, which governs `024.001-T`'s coding scope. It is stated here so the
work is priced rather than silently absorbed — the rev-1 draft made the pre-half mandatory
via AC-13 while giving it no budget line anywhere, inside the tightest and now valve-free
budget of the three.

*Note on §7.3's status column:* Probe 21 is tagged "Ship closure" there to indicate which
**budget** owns it. Its **pre**-half still executes in S1 before implementation begins
(§8 step 2, AC-13); only the **post**-half runs at closure.

### 8.5 If C's closure fails after the implementation merged

The implementation is on `main`; only the **closure** failed. Do **not** revert. Follow
the umbrella plan §8.3 contingency and §6.5 bounded recovery: `P1–P10` evaluate on the raw
post-call state, recovery is enumerated-set-only with approval before any destructive step,
and out-of-bounds paths are **reported, never moved**. If recovery cannot complete within
its bound, **HALT** and return to Stage with the raw state preserved. Because the failure
is in S2, S1's checkpoint remains the durable record of where the chain stood.

## 9. Intermediate and final state

| After | P-015 | Skill | Ship | Selectable verdicts |
|---|---|---|---|---|
| A | unchanged | unchanged | delegates | `CASCADE`, `SAFE_CLOSE`, `HALT` |
| B | authorizes, token-gated | unchanged (no token) | delegates | `CASCADE`, `SAFE_CLOSE`, `HALT` |
| **C** | authorizes, token-gated | **advertises token** | delegates | `CASCADE`, **`TASK_ONLY_FINALIZE`**, `SAFE_CLOSE`, `HALT` |

Every intermediate row is coherent and fail-closed. Activation happens **once**, at C.

**Rev 2 — C's own closure does not use the row it creates.** C closes by `CASCADE` (its
manifest has a live feature member, so T3 excludes `TASK_ONLY_FINALIZE`). The activated
verdict's first selection in the live workspace is at **`017-S`**. The final row therefore
describes what is *available* after C, not what C itself selected.

## 10. Risks, residuals, rollback

| ID | Risk | Disposition |
|---|---|---|
| **R-1** | **Budget is tight (~105 min, ~15 min margin — still smaller than `B18`'s own estimating variance)** | **No valve** (§12.1, rev 2). An overrun is an explicit **HALT to Stage**, never a partial `Done`. Risk is real, unhedged, and bounded to one shipment |
| **R-2** | `B18` sprawls beyond its allotment | `B18` **cites** T1–T5 rather than restating them (test row 12); the failure split and recovery are **carried from rev 6**, not re-derived |
| **R-3** | Token string drifts from P-015 (**CO-1**) | Test row 2 asserts **exact parity** against P-015. Any change routes through **Stage** into B's surface — never adjusted unilaterally here |
| **R-4** | C's closure fails, leaving the chain half-activated | §8.5. Implementation stays on `main`; only closure retries. S1's checkpoint is the durable record |
| **R-5** | **(rev 2 — resolved)** `024-F` left live after closure blocks `017-S` via Ship's P-001 gate | **Eliminated at source.** `024-F` is a manifest member, so the cascade archives it (Probe 24: `archived_ids` contains the feature; **0** other active top-level units). No barrier, no follow-up. Asserted by **AC-17**. *(Rev 1's Stage-disposition step was prose-only and, per Probe 23, un-enforceable — no feature-endpoint dependency edge exists.)* |
| **R-6** | Engine digest mismatch at closure | **HALT + full probe refresh** (§5). Not advisory |
| **R-7** | Guard drift between `B18` and P-015 over time | Residual **accepted**; a CI coupling gate is out of scope. Test row 2 covers the token, not the full guard set |
| **R-8** | **(rev 2)** `TASK_ONLY_FINALIZE` is never exercised by a **live** closure before `017-S` depends on it | Residual **accepted and named**. Mitigated by §8.4's relocated Probe 21 post-half against the **exact `017-S` 12-member fixture** (Probe 15's), plus Probe 15's own invariance run. Not equivalent to a live closure, and the plan does not claim it is |
| **R-9** | **(rev 2)** S1 ignores §8.1 and closes in the same session that merged the token | `B18` contract text + **AC-14** + **Go test row 13**. Same three-defense structure as A's §6, and the same honest residual: the checkpoint is evidence, not enforcement |
| **R-10** | **(rev 2)** `017-S`'s own task-only close leaves `018-F` `active` | Accepted, **not** gated. `017-S` is terminal in this chain, so no downstream shipment can be blocked. Recorded as a Stage hygiene disposition (§8.3) |

### 10.1 Rollback — reverse dependency order is MANDATORY after activation

An earlier draft described each merge as independently revertible. **That is false once
the chain is activated**, and the correction matters:

| Revert | While installed | Result |
|---|---|---|
| **C** | — | Token disappears ⇒ B's gate returns dormant ⇒ `CASCADE`/`SAFE_CLOSE`/`HALT`. **Clean.** Does **not** undo an already-completed closure |
| **B** | C still installed | Skill advertises **and implements** `TASK_ONLY_FINALIZE` with **no P-015 authorization**. **PROHIBITED** |
| **A** | B and/or C installed | Restores Ship's stale binary restatement and the `children`-only drift, contradicting both newer surfaces. **PROHIBITED** |

**Rule: after activation, roll back in `C → B → A` order only.** Reverting B requires C
already reverted; reverting A requires B and C already reverted. Reverting A alone is
additionally undesirable on its own merits — it reintroduces the §2.2 coverage drift that
A fixed.

**Downstream shipments MUST be held before any upstream contract is reverted (rev 2).**
Reverse-order reverting is necessary but **not sufficient**: a revert changes the contract
that a *queued* downstream shipment was planned against. Before reverting any of A/B/C:

1. **Confirm no downstream shipment is `active`.** If one is, it must reach a terminal
   state or be returned to Stage **first** — never revert the contract under a running
   Ship session.
2. **Hold `017-S` explicitly.** `017-S` is planned against the *activated* contract; with
   C reverted it is unclosable by design (that is what the whole chain exists to fix).
   Reverting C without holding `017-S` leaves a queued shipment whose close path has
   silently disappeared. The hold is a **Stage** action, recorded in the backlog, taken
   **before** the revert — not discovered afterwards by Ship.
3. **Re-plan before re-routing.** After any revert, the affected shipments return to Stage
   for re-planning. They are not re-routed on the assumption the old plan still applies.

The ordering rule protects the *installed contracts* from becoming mutually incoherent.
This hold rule protects the *queued work* from being routed against a contract that no
longer exists. Both are required; neither substitutes for the other.

Before activation (i.e. between merges) each shipment is independently revertible, because
the downstream surfaces do not yet exist. The claim is **phase-relative**, and this table
states which phase it holds in.

### 10.2 Failure paths

| Failure | Response |
|---|---|
| Any §3 site (MUST **or** CONSISTENCY) incomplete at 2 h | **HALT, return to Stage.** Deferring **any** site is **not** authorized (§12.1, rev 2) |
| Token cannot be advertised machine-readably | HALT — falsifies B's A7 gate. Return to Stage; B's surface is re-planned |
| Digest mismatch | HALT + full probe refresh |
| Classifier selects `TASK_ONLY_FINALIZE` for a non-conforming manifest during Probe 21 | HALT — guards are insufficient. Return to Stage |
| Classification returns `TASK_ONLY_FINALIZE` at **C's own** closure | **HALT** — the manifest is not the fully-covered root rev 2 asserts (T3 should exclude it). Return to Stage |
| Classification returns `SAFE_CLOSE` at C's closure | **HALT** — indicates `024-F` is not a qualifying root or the manifest carries something else. Return to Stage |
| S1 reaches the classification step at all | **HALT** — §8.1 required S1 to checkpoint and end after the reload. Return to Stage; the two-session property has been violated |
| S2's checkpoint recovery hits a validation/quarantine anomaly, ambiguity, or a `stage`-owned record | **FAIL CLOSED** to operator handoff (§8.2). No restore, no resume, no prune, no resolve |
| `--phase lifecycle` gate invoked after the archive | **HALT** — Probe 20 tripwire; it fails closed with zero active shipments and there is nothing left to re-claim |
| `P1–P10` detect an out-of-bounds artifact post-call | Bounded recovery; approval before any destructive step; HALT if unbounded |

## 11. Acceptance criteria

- **AC-1** `B1–B18` all applied: 16 clause sites + `B11` + `B18`.
- **AC-2** Frontmatter advertises the exact-version token; **parity with P-015 A7 asserted**.
- **AC-3** Step 0 selector is ternary and correctly ordered.
- **AC-4** `B18` states `G0–G13`, `P1–P10`, the failure split, bounded recovery, digest binding, CLI-only invocation, self-host ordering.
- **AC-5** `P1–P10` evaluate on the **raw post-call state before** archive restoration.
- **AC-6** `B18` **cites** T1–T5 from P-015 and does **not** restate them as its own authority.
- **AC-7** `B11` cross-reference present; step 8's own text otherwise unchanged.
- **AC-8** The existing fully-covered-root exception's skill-side handling is unmodified.
- **AC-9** Exactly **one** Go test function; H0 red 1 of 1; H1 green; `go vet` clean.
- **AC-10** `markdownlint` clean.
- **AC-11** `workflow-policies.md` and `_ship.agent.md` are **unchanged** by this shipment.
- **AC-12** **(rev 2)** No `clause-inventory.md` is produced. The §3 inventory in this committed plan is the single system of record. If implementation discovers a site absent from §3, it is **added to §3 by Stage, never skipped**; a newly discovered **MUST** site invalidates the budget and returns the unit to Stage.
- **AC-13** Probe 21 **pre**-half executed and recorded before merge (S1 step 2).
- **AC-14** **(rev 2)** Closure classifies **`CASCADE`** and completes as a fully-covered root. `TASK_ONLY_FINALIZE` is **not** selected at C's own closure (T3). **S1 does not classify or close**: it writes a checkpoint after the pre-self-close reload and ends; a **fresh S2** performs the full recovery protocol (§8.2) and the closure. AC-14 is **not** satisfied if one session performs both.
- **AC-15** Exactly 2 files changed (+ the evidence directory). 0 new production files.
- **AC-16** The absence of an executable close-path classifier in this workspace is **verified and recorded**; the stale `src/autoharness/...` references at `B9` (413) and `B17` (1075) are reconciled (§2.2).
- **AC-17** **(rev 2)** `024-F` is archived **by C's own cascade close**, as a manifest member — not by a follow-up step. After `023-S` archives, **zero** other top-level release units are `active`, so Ship Step 1's P-001 gate passes for `017-S` with no Stage disposition and no barrier.
- **AC-18** **(rev 2)** **Every** MUST and CONSISTENCY site in §3 is implemented. There is **no** overflow valve and **no** deferrable subset. An overrun is **HALT to Stage** — never a partial `Done`.
- **AC-19** **(rev 2)** `B18` states the §8.1 token-activation self-host ordering rule; Go test row 13 passes.
- **AC-20** **(rev 2)** S1's checkpoint is written via `backlogit_create_checkpoint` with `schema_version: 1`, `agent: ship`, a `phase`, and a `resume_hint` naming `023-S`, `024-F`, `024.001-T` and the merge SHA; domain data under `context`.
- **AC-21** **(rev 2)** The Step 6.1(a0) `--phase lifecycle` gate is run **while `023-S` is still active**, and **no** `--phase lifecycle` invocation occurs after the archive (Probe 20 tripwire).
- **AC-22** **(rev 2)** The post-merge closure branch is created from fresh `main` **before** any post-merge backlog mutation; the 6.1(e) backlog commit, the `operational-closure` artifact and the P-020 compaction status all land on that branch **before** the closure PR is pushed.
- **AC-23** **(rev 2)** Both Probe 21 halves remove their scratch workspace and re-run `backlogit sync` with **zero** `duplicate source id` warnings **before** any live backlog read, mutation, or closure sync (§8.4).

## 12. Sizing

**Size `M` · Complexity `high`.** Complexity carried as **enum-validated prose** —
`header-def.yaml` defines `size` on tasks but no `complexity` field.

**Rev 2 — implementation scope is exactly the skill + the Go test.** Evidence generation is
**not** task coding and is budgeted separately in **§8.4** (~30 min, charged to Ship
closure and not subject to the 2-hour task rule).

| Work item (implementation — charged to `024.001-T`) | Estimate |
|---|---|
| 8 MUST clause edits (`B1 B3 B5 B8 B13 B15` + selectors `B9 B10`) | ~28.5 min |
| `B17` mechanical MUST — **two lines** (1074 + 1075), rev 2 | ~3 min |
| 7 CONSISTENCY edits (`B2 B4 B6 B7 B12 B14 B16`) | ~10.5 min |
| `B11` step-8 cross-reference | ~3 min |
| **`B18` new classification section** (incl. the §8.1 token-activation rule) | **~29 min** |
| Go table-driven test (**13** rows) + ≤4 helpers | ~21 min |
| H0→H1, `go vet`, `go test`, `markdownlint` | ~10 min |
| **Total** | **~105 min ≈ 1.75 h** |

Under the 2-hour rule with **~15 min** margin. This is still the **tightest** of the
three and the plan says so rather than rounding it away.

**Rev 2 delta, stated honestly rather than netted to zero.** `B17` **+1.5** (it is two
lines, not one — §3); `B18` **+1** (the §8.1 token-activation rule); Go test **+1** (row
13); committed inventory re-scan **−5** (removed, AC-12). Net **−1.5 min**, so margin
moves ~13 → ~15 min. That margin is **still thinner than `B18`'s own estimating
variance**, and rev 2 no longer offers a valve to hide behind.

`B18` is ~29 min rather than rev 6's ~30 because **T1–T5 moved to B**. The rest of rev
6's `B18` content is carried, not re-derived.

### 12.1 No overflow valve (rev 2 — the valve is REMOVED)

Rev 1 declared a pre-authorized overflow valve that deferred a 5-clause CONSISTENCY
subset at the 100-minute mark. **That valve is removed in its entirety.**

**Why.** The valve permitted a `024.001-T` marked `Done` while sites named in its own
completion acceptance criteria were unimplemented. A completion criterion that the
completion path is pre-authorized to skip is not a criterion. The five "deferrable"
clauses are all statements about which close paths exist or are permitted; leaving any
of them describing a two-verdict world after a third verdict ships leaves the contract
**incorrect**, not merely uneven — the same defect class rev 1 already accepted for
`B7` and `B12`. The distinction between the two groups did not survive review.

**The rule now.** **Every** MUST and CONSISTENCY site in §3 is implemented. There is no
deferrable subset, no trigger, and no partial-completion state.

| Condition | Response |
|---|---|
| All §3 sites complete within 2 h | Proceed to closure |
| Any §3 site incomplete at 2 h | **HALT and return to Stage.** Never mark `024.001-T` `Done`. Never defer a site to a follow-up stash entry |
| `B18` materially incomplete at any point | **HALT and return to Stage** — it was never sheddable |

**Execution order is still normative**, for a different reason than rev 1's. The 7
CONSISTENCY edits run **after** `B18` and the Go test — not so they can be shed, but
because `B18` is the highest-uncertainty item and an early overrun should be discovered
while there is still time to HALT cleanly rather than at the 2-hour wall.

**If the MUST set alone cannot close in 2 h, HALT and return to Stage.** Deferring any
site — MUST or CONSISTENCY — is **not** authorized.

## Plan Hardening

**Blast radius.** C is the **only** shipment of the three that changes runtime behavior.
A removes duplication; B lands dormant; **C activates**. All three shipments' risk
concentrates here — which is exactly why it is last, why it is smallest in surface (2
files), and why its activation is observable in one place (the token).

**Hardening 1 — the tight budget is the top risk, and there is no longer a valve.**
~15 min of margin is still thinner than the estimating variance of `B18` (~29 min)
alone. Rev 1 answered that with a pre-authorized overflow valve; review established the
valve was itself the defect — it authorized marking `024.001-T` `Done` with sites named
in its own completion criteria unimplemented. §12.1 **removes it entirely**. The only
response to an overrun is **HALT to Stage**, never a partial `Done`. This makes the
budget risk **more** visible, not less: there is no slack left to absorb a bad estimate,
and the plan says so.

**Hardening 2 — C no longer closes by the path it ships, and that is a measured
downgrade, not a quiet one.** Rev 1's strongest property was that C self-proved
`TASK_ONLY_FINALIZE` by closing with it. Probes 22/23/24 made that untenable: the
task-only shape leaves `024-F` `active` (Probe 22), no barrier against a feature is
constructible (Probe 23), and only manifest membership disposes it (Probe 24). C now
closes by `CASCADE`. **What is lost:** an executed self-closure through the new path.
**What replaces it:** §8.4's relocated Probe 21 post-half, which exercises the
classifier against the **exact `017-S` 12-member fixture** rather than C's trivial
one-task shape — a closer analogue of the real consumer. **What remains unproven until
`017-S`:** an end-to-end *live* task-only closure. That is stated as a residual (R-8),
not papered over.

**Hardening 3 — the token-activation boundary is the highest-risk step.** C's merge
flips B's dormant gate live. A session that merged the token and then used it would be
authorizing itself — the precise failure A's §6 exists to prevent, in a form A's
Role-Boundary-scoped carve-out does **not** cover. §8.1 puts the rule in C's own
`B18` as a *procedure* rule (which A §4.2 says takes effect on reload), forcing S1 to
checkpoint and end. Three independent defenses, mirroring A's: normative contract text
in `B18`; **AC-14** fails if one session performs both; **Go test row 13** asserts the
rule is present in the merged file. A's row 11 scoping is **not** disturbed.

**Hardening 4 — CO-1 is asserted from both sides, but it is not the only coupling.**
B's row 3 asserts the token requirement exists; C's row 2 asserts **exact parity**. A
mismatch fails C's harness at H1, **before** any closure attempt. Unilateral adjustment
inside C is forbidden: a token change is a **policy** change and routes through Stage
into B's surface. The full coupling set is the deliberation's §7.4 matrix (CO-1 …
CO-14); CO-1 is only the one requiring a **literal string match**.

**Hardening 5 — the guard set is inherited, not invented, and its evidence is Probe 19
alone.** `G0–G13`, `P1–P10`, the failure split and bounded recovery are **carried from
the umbrella plan** (rev 6 §6.2–§6.5), which survived six revisions and multiple
adversarial rounds. **Correction (rev 2):** rev 1 cited "Probes 16 **and** 19". Probe 16
is **withdrawn as evidence** — it diverged from the specified procedure in five ways —
and is retained for history only. **Probe 19** implements the procedure as specified and
executes **both** approval branches; it is the sole authority. C's job is
**transcription into the skill**, not re-derivation. That is what makes ~29 min for
`B18` credible where inventing it would not be.

**Hardening 6 — the P-001 residue is eliminated mechanically, and the one that remains
has no consumer.** Rev 1 carried T4's live orphan as a Stage follow-up. A follow-up is
prose: nothing stops `017-S` being routed before someone remembers it, and Probe 23
proved no dependency-graph barrier against a feature can be written. Rev 2 removes the
hazard at its source — `024-F` is a manifest member, so the engine's own cascade
archives it (Probe 24), leaving **zero** other active top-level units. The symmetric
residue at `017-S` (`018-F` left `active`) is recorded rather than fixed **because it
has no downstream consumer** — `017-S` is terminal. Fixing one and recording the other
is a deliberate scope line, stated in §8.3.

**Hardening 7 — what could make this plan wrong.** (a) If the token cannot be advertised
machine-readably, B's A7 gate is unimplementable and the chain re-plans — **falsified in
advance** at §2.1, risk **low**. (b) If implementation finds a **MUST** site absent from
§3, the budget is invalidated and the unit returns to Stage (**AC-12**) — the same
tripwire rev 6 set, retained deliberately, now enforced against the committed §3
inventory rather than a regenerated file. (c) If the engine digest has moved, §5 halts
before any call. (d) If an **executable** classifier existed in this workspace,
markdown-only edits would not activate the path — **verified false** at §2.2 (no `src/`,
no classifier), which is why "0 source changes" holds here. (e) If the cascade does
**not** archive a manifest feature member, §8.3's remedy fails and `024-F` survives —
**measured false** by Probe 24 (`archived_ids` contains the feature). Each failure mode
has a **detector** and a **named owner**; none degrades silently.

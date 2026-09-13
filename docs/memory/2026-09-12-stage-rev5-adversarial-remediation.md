# Stage session — rev-5 adversarial remediation (task-only shipment finalization)

- **Date:** 2026-09-12
- **Agent:** Stage
- **Branch / base HEAD:** `chore/stage-pipeline-policy-gap` @ `e06b4d8`
- **Scope:** ONE bounded remediation cycle after the adversarial `MUST_REPLAN` on plan rev 4.1
- **Mode:** docs/backlog only — no production, test, skill, agent or policy implementation;
  no push, no PR, no Copilot invocation, no Ship handoff, no shipment claim, no merge

## 1. Checkpoint lifecycle (owner-exclusive)

| Phase | Result |
|---|---|
| Enumeration | `backlogit checkpoint list`, **no** `status`/`agent` filter → **8** checkpoints |
| Anomaly scan (fail-closed, runs first) | **0** anomalies — no quarantined, malformed, or empty-`agent`/`status` records |
| Candidate partition | exactly **1** active, `agent: stage` → `checkpoint-20260913-014814.json` |
| Operator selection | explicit + unique (operator directed autonomous continuation and named the file) |
| Owner validation | `agent: stage` ✓ — no `ship`-owned checkpoint touched |
| Prune gate (engram) | **GREEN** — see §1.1 |
| Restore | loaded; phase `harvest-complete-shipment-assembled`, `021-S`/`022-F` context restored |
| Resolve | `backlogit checkpoint resolve` → `ok: true`; **active checkpoints now 0** |

### 1.1 Engram prune gate — recovered rather than failed closed

The gate initially failed: `daemon unavailable: Daemon failed to reach Ready state`.
The documented `ENGRAM_DIRECT=1` fallback **also** failed. Diagnosis found a **stuck
daemon (PID 16536, ~2h old, never Ready)** holding the workspace lock — exactly the
condition the tool's own error text names ("stop that engram process if it appears
stuck"). Stopping it and retrying returned a healthy index (86 code files, 297 edges,
100% embedding coverage), so the gate passed **green**.

This mattered: `engram unreachable` is a **session-ending fail-closed verdict**. It
would have been wrong to declare it without exhausting the documented remedy, and
equally wrong to skip the gate. Engram data is gitignored, rebuildable cache, so the
bounded risk of stopping the daemon was justified.

## 2. Adjudications applied

### 2.1 Preserved, not re-derived

`022-F`, `022.001-T`, task-only `021-S` (manifest exactly `[022.001-T]`), and
`017-S -> 021-S (blocks)` were **preserved unchanged**. Verified post-edit.

### 2.2 Step 5.5 item 5 compliance rationale (recorded precisely, no invented exception)

Stage Step 5.5 item 5 says *verify the manifest and **report** any discrepancies*. It
does **not** say halt, and it does **not** state manifest/hierarchy equality as a final
invariant. What Stage actually did:

1. assembled **parent-first** (feature first, per item 4a);
2. excluded `022-F` using the **registry-declared** `return_blocked` operation — inside
   Stage's Role Boundary, not an improvised mutation;
3. **restored** `022-F` to `queued` afterwards (it is live, root, and protected);
4. **verified** the resulting manifest is exactly `[022.001-T]`;
5. **reported** the intentional discrepancy (1 manifest member vs 2 harvested items).

That is item 5 **satisfied as written**. No unwritten exception is claimed, and the
task-only shape is required by the plan's own T2/T4 topology.

### 2.3 Actor reconciliation — **tested, and it returned a BLOCKER**

Directed to make the artifacts consistently permit **Ship OR operator** as PR actor,
*unless the installed Ship mandatory sequence makes it impossible*. Stage tested it.

Ship's Role Boundary does list Git push and PR create/update/merge as **Allowed** — so
the claim was plausible. But the Role Boundary is **necessary, not sufficient**: Ship
Step 5 items 1a/5a mandate the `pipeline-topology` gate before `pr-lifecycle`, with
*"exit 1/2 halts immediately … never fail-open"*. **Probe 17** exercised all four routes;
**all fail closed** without an **active** shipment, and `active` requires a **claim** —
forbidden here. `--force` is *"Operator-only … Never reachable from an agent surface."*

**Disposition: BLOCKER returned, not feasibility asserted.** The OPERATOR is the PR
actor for PR #54. The pre-existing `018-F`/`018.008-T` operator-only language is
therefore **confirmed and unweakened** — the apparent contradiction dissolves because
this is a **sequence** limitation, not a Role Boundary limitation. Both artifacts gained
an `ACTOR RECONCILIATION` paragraph recording the evidence. Independently, **no
non-Stage actor may modify planning content in any mode**, so review comments requiring
planning edits route back to **Stage** regardless of who carries the PR.

## 3. Probes executed (durable, committed, reproducible)

Evidence: `docs/plans/evidence/2026-09-12-task-only-shipment-finalization/`
(scripts + transcripts). Scratch: `.backlogit/runtime/stage-probe-scratch/` (gitignored).
Recovery: `.autoharness/backups/stage-recovery/` (gitignored). No `.db`/`-wal`/`-shm`
committed.

| Probe | Verdict |
|---|---|
| **14** dependency lifecycle | **PASS** — successor suppressed while predecessor `queued` **and** `active`; predecessor → `archived` + `archived_status: shipped` + `commit`; after `sync`, successor eligible. All verdicts `status: ok` |
| **15** EXACT `017-S` fixture (12 members: 1 live + 3 archived tasks + 8 archived subtasks) | **PASS — `INVARIANCE_FAILURES=0`**; `returned_ids=[]`; `archived_ids=[<live task>, <shipment>]` |
| **16** injected partial failure + bounded recovery | **PASS** — real torn state; `MUTABLE_EQUIVALENCE_FAILURES=0`, `RESIDUAL_UNEXPECTED_PATHS=0`, `APPEND_ONLY_REWIND_VIOLATIONS=0` |
| **17** Ship PR-lifecycle gate | **BLOCKED** — all four routes fail closed |

**Method correction worth recording.** Probe 14's first run returned `status: degraded`
(`required backlog directory is unavailable: …\archive`) for the two baseline checks,
because a freshly-`init`-ed workspace has no `archive/` until its first archival. A
`degraded` reading is **not a verdict** — reporting it as suppression would have been a
vacuous claim. The script now pre-creates `archive/`; the committed transcript shows
`degraded_reason: null` throughout.

**Engine identity (binding):** `backlogit 1.10.1-0.20260823032255-b07729386a31+dirty`,
**SHA-256 `1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98`**
(`C:\Tools\backlogit.exe`, 25,293,824 bytes). The `+dirty` string is not unique, so the
**digest** binds. Mismatch = **HALT + probe refresh**; the advisory fallback is withdrawn.

**Withdrawn as evidence:** rev 4.1's `%TEMP%`-and-deleted probes 5, 5b, 10, 12, 13.

## 4. Artifacts changed

| Artifact | Change |
|---|---|
| `docs/plans/…-task-only-shipment-finalization-plan.md` | **rev 4.1 → rev 5** (all 10 findings + §11 blocker) |
| `docs/plans/evidence/2026-09-12-task-only-shipment-finalization/` | **NEW** — README + 3 scripts + 4 transcripts |
| `.backlogit/queue/022-F.md` | 4-file surface, T1–T5, dependency proof, sequencing gate, blocker |
| `.backlogit/queue/022.001-T.md` | exact 4-file inventory, AC-1..AC-19, itemized edit budget, engine digest |
| `.backlogit/queue/018-F.md` | `ACTOR RECONCILIATION` paragraph |
| `.backlogit/queue/018.008-T.md` | `ACTOR RECONCILIATION` paragraph |

## 5. Validation

| Check | Result |
|---|---|
| `backlogit sync` | `Indexed 232 artifacts` |
| `backlogit doctor` | **No issues found** |
| `021-S` manifest | exactly `[022.001-T]` — task-only preserved |
| `017-S` dependencies | `017-S → 021-S (blocks)` preserved |
| `022-F` children | exactly one (`022.001-T`) |
| DAG | `ready_set: [021-S]`, `next_eligible: 021-S`, `cycle_detected: false`, `status: ok` |
| `go vet ./...` | exit **0** |
| `go test ./...` | exit **0** — suite green, so H0-red/H1-green is feasible |
| markdownlint | exit **0** |
| Worktrees | **1** |
| Diff scope | artifact-only (`docs/**`, `.backlogit/**`) |
| Frontmatter integrity | all IDs/types/statuses/parents/size fields intact after description edits |
| Internal plan review | **ADVISORY** — see §5.1 |

### 5.1 Internal plan review — it found real defects, and that changed the output

An independent internal reviewer examined all 14 sections, Plan Hardening, and the
four transcripts, and **independently re-verified all 14 §6.1 line citations against
the installed contract files** (all accurate). It returned **15 findings**.

**Closed in place: 1×P0, 4×P1, 3×P2, 3×P3.** The four most consequential were
defects that would have shipped:

1. **P0 — a self-contradiction.** The plan simultaneously required the P-015
   exception to be "byte-for-byte unchanged" *and* mandated edits to two lines
   **inside** that exception. Resolved by defining a precise **preserved region**
   (§4.1): the safety-bearing preconditions stay byte-identical; only the two
   sentences asserting *exclusivity* are amended.
2. **P1 — P4 would have failed on every successful close.** Append-only logs were
   inside the baseline mutation-comparison scope, yet Probe 16's own transcript
   shows those logs grow on every operation. The happy path would have routed
   straight to recovery + HALT, and §6.5.2 forbids restoring them, so recovery
   could never have re-established equivalence. The path would have been
   permanently unusable.
3. **P1 — unbounded quarantine.** §6.5.3 said to quarantine "any path present now
   but absent from the inventory". Against the live `.backlogit/queue/` that would
   have relocated **unrelated live backlog records** out of the queue during
   recovery — worse than the failure being recovered. Now bounded to the baseline
   scope, with out-of-scope paths **reported, never moved**.
4. **P1 — an evidence overclaim of exactly the kind rev 5 exists to remove.** The
   draft asserted "No conclusion in this plan rests on a deleted sandbox." That was
   **false**: G11/G12/G13/P1/P5 and the whole T5 rationale rest on retained-but-
   uncommitted `%TEMP%` probes. The sentence is withdrawn and replaced by an
   explicit **evidence-degraded table** (§3).

One reviewer premise was checked and **partly rejected on evidence**: the claim that
Probe 16 violated T2 because its sole task sat in `archive/` pre-call. Verified
against the transcripts — `backlogit move --status done` **relocates** a record to
`archive/` while it still *declares* `status: done`, which is the documented
behavior the plan's own G7 describes. Under the plan's frontmatter-based liveness
rule the probe **is** T2-conforming, so §3 row 16 stands. The reviewer's adjacent
G9 finding was, however, correct and is fixed (**G9 conjunct 2**: destination path
must be absent and not a directory).

**Left OPEN and recorded, not silently closed: 8 items (O-1…O-8, all P2/P3).** None
is reachable before implementation begins, so none blocks the artifact-only PR. The
highest-risk is **O-6**: the post-archival closure route (§8 steps 9–12) has never
been probed against the same gate Probe 17 falsified — after the shipment archives
there are **zero active shipments**, so a deadlock there would land *after* the
destructive call. They are enumerated in plan §12.2 and carried into the re-review
scope at §13 item 9.

## 6. Open blockers

1. **PR #54 has no agent-executable route** (§11). Operator decision: carry PR #54
   directly, **or** commission a separate, separately-reviewed unit to amend Ship's
   Step 5 gate for artifact-only no-shipment PRs. Stage takes neither.
2. **`017-S` stays blocked** until `021-S` ships and closes — enforced by the edge,
   confirmed by Probe 14.

## 7. Next actor

**The OPERATOR.** Not Ship, not Orchestrator.

1. Independent multi-model adversarial re-review — scope: plan **§13**, which now
   has **9 items**: the original 8 plus item 9 covering the eight OPEN internal-review
   findings (O-1…O-8) in §12.2.
2. If READY, operator carries PR #54 under the §14 gates.
3. **No source implementation or shipment claim** until PR #54 merges **and** `021-S`
   is present on `main`; only then does the Orchestrator route `021-S`, and only `021-S`.

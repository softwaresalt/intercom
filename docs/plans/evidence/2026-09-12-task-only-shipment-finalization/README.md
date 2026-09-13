# Probe evidence — verified task-only shipment finalization

Durable, workspace-contained evidence for
`docs/plans/2026-09-12-intercom-go-task-only-shipment-finalization-plan.md` (rev 5).

Superseded by this directory: the rev-4/4.1 claim that probes ran "in throwaway
backlogit workspaces under `%TEMP%`" with the sandboxes deleted afterwards. Those
runs left **no reproducible artifact**. Every probe recorded here was re-executed
inside the workspace, and both the command script and the full transcript are
committed.

## 1. Engine identity of record (binding, not advisory)

| Field | Value |
|---|---|
| Version string | `backlogit version 1.10.1-0.20260823032255-b07729386a31+dirty` |
| Executable path | `C:\Tools\backlogit.exe` |
| **SHA-256** | `1E106F5FD1E2D82E4F632AFEE40FD70B95DC7416886C3EBF266361F406959A98` |
| Size (bytes) | `25293824` |
| Last write (UTC) | `2026-08-23T03:55:36.6658095Z` |
| Captured | 2026-09-12 (this session) |

Reproduce:

```powershell
$cmd = Get-Command backlogit
backlogit --version
(Get-FileHash -Algorithm SHA256 $cmd.Source).Hash
```

**Binding rule (normative).** The `+dirty` suffix is *not* a unique build
identity — any build from an uncommitted tree at commit `b07729386a31` reports
the identical string. The **SHA-256 above is therefore the binding identity.**
Before the authorized call, Ship recomputes the digest of the executable it is
about to invoke and compares it to this value. **A mismatch is a HALT and
requires a probe refresh** (re-run every probe in this directory against the new
binary and re-record the digest). There is no advisory fallback: the previous
"documented as advisory rather than as a guarantee" escape is **withdrawn**.

## 2. Probe index

| Probe | Question | Script | Transcript | Verdict |
|---|---|---|---|---|
| **14** | Does a shipment-blocks edge suppress a successor, and does closing the predecessor release it? | `probe14-dep-lifecycle.ps1` | `probe14-dependency-lifecycle.txt` | **PASS** *(stock `init` seed)* |
| **15** | Do `017-S`'s **exact** 11 pre-archived members survive `ShipShipment` byte/path/parent unchanged? | `probe15-exact-017s-fixture.ps1` | `probe15-exact-017s-fixture.txt` | **PASS — 0 failures** *(stock `init` seed)* |
| **16** | Does a real partial failure recover to byte equivalence without rewinding append-only state? | `probe16-bounded-recovery.ps1` | `probe16-bounded-recovery.txt` | **SUPERSEDED by 19 — withdrawn as evidence** |
| **17** | Can Ship satisfy its Step 5 topology gate in a PR-lifecycle-only session? | (inline) | `probe17-ship-pr-lifecycle-gate.txt` | **BLOCKED — route not executable** |
| **18** | Under the **live** configuration, does the inbound `017-S -> 021-S` edge holder survive the prerequisite close **byte-identical**? | `probe18-liveconfig-inbound-edge.ps1` | `probe18-liveconfig-inbound-edge.txt` | **PASS** |
| **19** | Does the **§6.5.3 procedure as specified** recover a real errored invocation — on **both** approval branches? | `probe19-bounded-recovery-v2.ps1` | `probe19-bounded-recovery-v2.txt` | **PASS — both branches** |
| **20** | **O-6:** post-archive, does the closure route hit a topology-gate deadlock? | `probe20-post-archive-closure-route.ps1` | `probe20-post-archive-closure-route.txt` | **Gate blocks on all 4 routes; closure-PR path does not invoke it** |

### Rev-6 evidence rules (binding on probes 18, 19, 20)

Probes 18–20 were written to close rev-5 open item **O-7** and satisfy three
rules that probes 14–17 do not:

1. **In-script digest gate.** Each script resolves the backlogit executable path,
   computes its SHA-256, compares it to the identity of record, records
   `--version`, prints `DIGEST_GATE=PASS`/`FAIL`, and **`exit 1` before any
   operation** on mismatch. All four fields appear in the transcripts.
2. **Live-configuration seeding.** The disposable workspaces are seeded from the
   **live** `.backlogit` control files — `config.yaml`, `header-def.yaml`,
   `hooks.yaml`, `registry.yaml`, `migration.yaml` and `templates/` — with each
   seed file's SHA-256 recorded in the transcript. They are **not** stock
   `backlogit init` workspaces. Probe 20 additionally seeds the live
   `.autoharness` config so the topology gate resolves.
3. **Robust root resolution.** Repo root is resolved via
   `git rev-parse --show-toplevel`, not by counting `..` segments.

**The configuration delta is material, not cosmetic.** Probe 18 records:

- `SEED_STATUS_ENUM_HAS_SHIPPED=False` — the live status enum has **no**
  `shipped` value, which is the mechanism behind the §2 root cause;
- `SEED_HOOKS_VALIDATE_TRANSITION=True` and `SEED_HOOKS_PRE_TASK_GATE=True` —
  the live workspace enforces lifecycle transition validation and a
  pre-task-completion gate that a stock workspace does not have;
- the live `registry.yaml` routes `done` artifacts to `archive/`, so a **live**
  `done` task is physically located in `archive/` — which is why liveness must
  be read from frontmatter `status`, never from directory location.

### Authorization surface — CLI only

Every runtime authorization in this evidence set was established **through the
`backlogit` CLI**. The **MCP surface is neither probed nor authorized**: no
script invokes `backlogit mcp`, no transcript records MCP behavior, and **no
MCP/CLI equivalence is claimed**. Authorizing closure through MCP would require
its own digest binding and its own probes.

### Probe 18 — live-config inbound-edge invariance

Fixture is the **exact prerequisite shape**, not a generic one: a root covering
feature live and outside the manifest (`022-F` analog), one live `done` task
(`022.001-T`), a task-only shipment whose manifest is exactly that task
(`021-S`), plus a second shipment holding `depends_on <prerequisite> --type
blocks` (`017-S`).

| Assertion | Result |
|---|---|
| `DIGEST_GATE` | **PASS** |
| root cause under live config (`move --status shipped`) | **`exit 9`** — *"shipment must be shipped via ShipShipment, not a direct status update"* |
| `INBOUND_EDGE_BYTE_IDENTICAL` | **True** |
| `INBOUND_RECORD_PATH_CHANGED` | **False** |
| `PARENT_FEATURE_BYTE_IDENTICAL` | **True** |
| `returned_ids` | `[]` |
| archived shipment record | `status: archived`, `archived_status: shipped`, `commit: <sha>` |
| edge after archival | `002-S → 001-S (blocks)` — persists |

This discharges the requirement that the inbound `017-S -> 021-S` record be in
the baseline/lock scope **and** proven byte-identical through the prerequisite
close.

### Probe 19 — bounded recovery, rewritten and run on both approval branches

**Why probe 16 was superseded.** Probe 16 was committed, but it did not implement
the procedure it was cited for. Five divergences:

1. it enumerated **whole `queue/` + `archive/` directories** and moved everything
   absent from its inventory — an unbounded quarantine that, against the live
   queue, would relocate unrelated live records;
2. it compared append-only streams by **length only**, so an equal-length or
   prefix mutation was invisible;
3. it never executed the **approval-withheld** branch, so the approval gate was
   asserted but not demonstrated;
4. it resolved its repo root by `..\..\..` from a four-deep directory, landing on
   `docs\` — the committed script **does not reproduce as written**;
5. it ran on stock `backlogit init` defaults.

**Probe 19 results.** Failure injection is identical (archive destination
occupied by a directory); `SHIP_EXIT=1` after partial mutation, torn state
observed (shipment live in `queue/` declaring `status: shipped`).

| Assertion | Withheld | Granted |
|---|---|---|
| approval gate evaluated before any destructive step | **yes** | **yes** |
| quarantined (moved, never deleted) | **0** | **1** |
| restored (enumerated inventory only) | **0** | **4** |
| out-of-bounds path `.001.001-T.md.lock` | **reported, not moved** | **reported, not moved** |
| torn state preserved as evidence | **yes** | n/a |
| `MUTABLE_BYTE_EQUIVALENCE_FAILURES` | n/a | **0** |
| `MUTABLE_PATH_EQUIVALENCE_FAILURES` | n/a | **0** |
| `DEPENDENCY_EQUIVALENCE_OK` | n/a | **True** |
| `RESIDUAL_UNEXPECTED_PATHS_IN_BOUNDED_SET` | n/a | **0** |
| `APPEND_ONLY_PREFIX_VIOLATIONS` (full byte prefix) | **0** | **0** |
| `APPEND_ONLY_DELETIONS` | **0** | **0** |
| cache rehydration | n/a | official `backlogit sync`; `DB_BYTE_RESTORED=False` |
| terminal disposition | **HALT** | **HALT**, quarantine preserved |

The residual assertion is deliberately scoped to the **enumerated bounded set**,
not to whole directories: out-of-scope paths are legitimately present and are
deliberately left untouched.

This is **bounded recovery, not atomic rollback.** The engine offers no
transaction. What is established is that the §6.5.3 procedure, as specified,
executes end-to-end against a real errored invocation under live configuration
with both approval outcomes demonstrated — **not** that it is correct for all
failure modes.

### Probe 20 — O-6, the post-archive closure route

Driven to the exact post-archive state (shipment archived, **zero** active
shipments) in a live-config- and live-`.autoharness`-seeded disposable workspace.

**Q1 — does a lifecycle gate block post-archive?** Yes, on every route:

| Route | Exit | Token |
|---|---|---|
| `--mode agent --shipment <archived> --phase lifecycle` | **1** | `LIFECYCLE_NO_ACTIVE_SHIPMENT` |
| `--mode agent --phase ambient` | **2** | `agent mode requires --shipment` |
| `--mode manual --shipment <archived> --phase lifecycle` | **1** | `LIFECYCLE_NO_ACTIVE_SHIPMENT` |
| `--mode ci --phase lifecycle` | **2** | `--phase lifecycle requires --shipment` |

`POST_ARCHIVE_ROUTES_NONZERO_EXIT=4`. No active shipment remains to satisfy the
gate, and re-claiming an archived shipment is not a supported transition, so this
deadlock — if entered — is unrecoverable in band.

**Q2 — does the installed Ship contract run that gate before the closure PR?**
No. Cited from `.github/agents/_ship.agent.md`: the lifecycle topology gate is
mandated at **L559** (before build), **L571** (before *implementation* PR
creation) and **L767** (before closure/safe-close, while the shipment is still
active). The closure-PR step at **L741–744** invokes `pr-lifecycle` without a
topology-gate precondition, and **L723** requires fresh operator approval
("the prior main PR approval does not transfer").

**Disposition:** not a deadlock in the plan's §8 sequence, but a one-clause-deep
hazard. Plan §8.1.1 therefore **prohibits** running a lifecycle-phase topology
gate during the closure-PR steps and records a tripwire.

### Probe 14 — dependency lifecycle, end to end

Fixture mirrors the live `017-S -> 021-S` topology: two task-only shipments,
successor `depends_on` predecessor `--type blocks`.

| Stage | Predecessor | `ready_set` | Successor eligible? |
|---|---|---|---|
| baseline | `queued` | `["001-S"]` | **no — suppressed** |
| after claim | `active` | `[]` | **no — suppressed** |
| after close + `sync` | `archived` / `archived_status: shipped` | `["002-S"]` | **yes** |

All three verdicts are `status: ok` (**not** `degraded`).

> **Method note, load-bearing.** The first run of this probe returned
> `status: degraded`, `degraded_reason: "required backlog directory is
> unavailable: ...\.backlog\archive"` for the two baseline checks, because a
> freshly-`init`-ed workspace has no `archive/` directory until its first
> archival. A `degraded` reading is **not a verdict** and would have made the
> suppression claim vacuous. The script now pre-creates `archive/` and the
> committed transcript shows `degraded_reason: null` throughout.

**Archived-predecessor handling (clarified).** After the predecessor archives,
the successor **still carries the edge** —
`dependencies: [{id: 001-S, type: blocks}]` is present in `002-S` — yet the
successor is eligible. Eligibility is therefore derived from the predecessor's
**status**, not from edge removal: an `archived` predecessor **satisfies** a
`blocks` edge. The edge persists as a durable lineage record. This is the same
rule the Orchestrator consumes (`autoharness gate dag-readiness`).

Merge metadata survives closure: the archived shipment record carries
`status: archived`, `archived_status: shipped`, and `commit: <sha>`.

### Probe 15 — exact `017-S` fixture

Rev 4.1 supported pre-archived members on a **three-member representative**
fixture. This probe builds the **exact** live shape instead:

- one root covering feature, live in `queue/`, excluded from the manifest;
- one live task, `done` at closure;
- **3 archived tasks + 8 archived subtasks (3/3/2)** = 11 pre-archived members;
- **12-member task-only manifest**.

That is byte-for-byte the topology inventoried from the live `017-S`.

Result: `archived_ids: ["001.004-T", "001-S"]`, `returned_ids: []`,
and **all 11 pre-archived members plus the root parent are `byte=SAME`,
`path=SAME`, `parent=SAME` — `INVARIANCE_FAILURES=0`.**

This discharges the adversarial precondition that pre-archived descendants may
be supported **only** on an exact executed fixture. Support is granted; the
contingent post-prerequisite manifest normalization before `017-S` claim is
**not** required.

### Probe 16 — injected partial failure and bounded recovery (SUPERSEDED)

> **Superseded by Probe 19 and withdrawn as evidence for the §6.5.3 procedure.**
> Retained for history. Its torn-state observation remains a valid
> characterization of engine behavior; its *recovery* run is not evidence that
> the specified procedure works, because it implemented a different procedure
> (unbounded directory sweep, length-only append-only comparison, no
> approval-withheld branch) and its committed script does not reproduce as
> written (repo root resolves to `docs\`). See Probe 19 above.

Failure injection: the shipment record's archive destination is occupied by a
directory named `001-S.md`.

- `SHIP_EXIT=1` — a real non-zero exit **after** partial mutation.
- **Torn state observed:** `001-S` present in `queue/` declaring
  `status: shipped` *and* present in `archive/` as a directory, while
  `001.001-T` was already archived. The engine refuses to create a live
  `status: shipped` shipment via `move` (exit 9), so this state is
  self-inconsistent by construction.
- **Approval gate evaluated BEFORE any destructive move.** Enumeration is
  non-destructive and runs first; quarantine and restore are gated.
- Recovery path: **`.autoharness/backups/stage-recovery/`** — workspace-contained,
  `RECOVERY_PATH_GITIGNORED=True`, and outside the compared
  `.backlog` queue/archive/log inventory.
- Unexpected paths **quarantined (moved), never deleted**; quarantine preserved.
- Only enumerated **mutable markdown** restored from snapshot.
- **Append-only streams never rewound:** `APPEND_ONLY_REWIND_VIOLATIONS=0`
  (per-artifact logs grew 341→1301 and 614→1031 bytes across the failure and
  recovery). A **recovery event was appended**, not a rewind.
- **DB/WAL/SHM never byte-restored**; the official `backlogit sync` rehydrated
  the index.
- `MUTABLE_EQUIVALENCE_FAILURES=0`, `RESIDUAL_UNEXPECTED_PATHS=0`, engine view
  restored, terminal disposition **HALT**.

This is **bounded recovery, not atomic rollback.** The engine offers no
transaction; what is proven is evidence-based restoration of an enumerated file
set.

### Probe 17 — Ship PR-lifecycle-only route: NOT EXECUTABLE

The `pipeline-topology` gate is installed and live in this workspace
(`.autoharness/gates/pipeline-topology-force-audit.log` records real operator
`--force` use on 2026-09-05). Ship Step 5 items 1a/5a mandate it before
`pr-lifecycle`, with "exit 1/2 halts immediately … never fail-open".

| Invocation | Exit | Token / message |
|---|---|---|
| `--mode agent --phase lifecycle` (no shipment) | **2** | `agent mode requires --shipment <shipment_id>` |
| `--mode agent --shipment 021-S --phase lifecycle` (queued, unclaimed) | **1** | `LIFECYCLE_NO_ACTIVE_SHIPMENT: expected exactly one active shipment` |
| `--mode manual --phase lifecycle` (no shipment) | **2** | `--phase lifecycle requires --shipment <shipment_id> in any mode` |
| `--mode agent --phase ambient` (no shipment) | **2** | `agent mode requires --shipment <shipment_id>` |

Every route fails closed. The gate can only pass with **exactly one active
shipment**, which requires a **claim** — precisely what a PR-lifecycle-only
session must not do. `--force` is documented "Operator-only … **Never reachable
from an agent surface**."

**Conclusion:** Ship's Role Boundary permits Git push and PR create/update/merge
*as categories*, but the installed Step 5 sequence makes a **PR-lifecycle-only
Ship session unexecutable** on PR #54. The operator is the only actor who can
carry PR #54. Recorded as a blocker rather than asserted away — see plan §11.

## 3. Scratch and cleanup policy

Probes execute in `.backlogit/runtime/stage-probe-scratch/` (gitignored,
workspace-contained). Recovery/quarantine artifacts live in
`.autoharness/backups/stage-recovery/` (gitignored, workspace-contained).
Neither is under `.backlog`/`.backlogit` queue, archive, or log inventory.
Scratch workspaces are removed after transcripts are captured; **the durable
evidence in this directory is never deleted.** No raw `.db`, `-wal`, or `-shm`
file is committed.

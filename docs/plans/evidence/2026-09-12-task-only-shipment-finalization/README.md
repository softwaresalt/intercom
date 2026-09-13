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
| **14** | Does a shipment-blocks edge suppress a successor, and does closing the predecessor release it? | `probe14-dep-lifecycle.ps1` | `probe14-dependency-lifecycle.txt` | **PASS** |
| **15** | Do `017-S`'s **exact** 11 pre-archived members survive `ShipShipment` byte/path/parent unchanged? | `probe15-exact-017s-fixture.ps1` | `probe15-exact-017s-fixture.txt` | **PASS — 0 failures** |
| **16** | Does a real partial failure recover to byte equivalence without rewinding append-only state? | `probe16-bounded-recovery.ps1` | `probe16-bounded-recovery.txt` | **PASS** |
| **17** | Can Ship satisfy its Step 5 topology gate in a PR-lifecycle-only session? | (inline) | `probe17-ship-pr-lifecycle-gate.txt` | **BLOCKED — route not executable** |

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

### Probe 16 — injected partial failure and bounded recovery

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

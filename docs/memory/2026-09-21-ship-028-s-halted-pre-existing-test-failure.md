# Ship session checkpoint — shipment 028-S (merge-strategy structural verification gate)

**Date**: 2026-09-21
**Branch**: `feat/028-s-merge-strategy-structural-verification-gate`
**Shipment**: 028-S (status: `active`, claimed this session)
**Covering feature**: 031-F — "Merge-strategy structural verification gate"

## Status: HALTED before PR creation — pre-existing unrelated test failure blocks required CI

All 5 tasks in the shipment manifest are **done** and committed:

| Task | Status | Commit |
|---|---|---|
| 031.001-T — `scripts/check-merge-strategy.sh` + self-test fixtures | done | `d47fc50` |
| 031.002-T — GITHUB_TOKEN feasibility finding | done | `b2b7a83` |
| 031.003-T — wire job into CI + `ci-gate` needs | done | `ce576c0` |
| 031.004-T — LOCAL DIVERGENCE header update | done | `ce576c0` |
| 031.005-T — operator-trigger/promotion doc | done | `322b6e5` |

Housekeeping commit `ceb6f9b` preserves a pre-existing session memory file from
the PR #74 handoff (untracked at session start; content preserved, no data
lost). `.backlogit/stash.jsonl`'s reported modification at handoff was
verified byte-for-byte identical after CRLF normalization (zero content
diff) and was restored to HEAD non-destructively before branch creation.

All quality gates pass for the actual shipment 028-S change:
* `go vet ./...` — PASS
* `gofmt -l .` — no output (PASS)
* `go build ./...` — PASS
* `bash scripts/check-merge-strategy.sh --self-test` — PASS (6/6 fixtures)
* `.github/workflows/ci.yml` — YAML-valid, `ci-gate` needs + results string
  wired correctly (verified via PyYAML parse)

**Blocking finding**: `go test ./...` fails on **both** `main` and this
branch (confirmed identical failure on stock `main` via
`git checkout main && go test ./tests/integration/... -run
TestOperationalClosurePostMergeFilenameConformance`):

```text
--- FAIL: TestOperationalClosurePostMergeFilenameConformance (0.00s)
    operational_closure_post_merge_filename_test.go:173: artifact
    "027-S-030-F-closure-evidence-repair-post-merge-closure.md" does not
    conform to documented Output form
    "docs/closure/{shipment_id}-{feature_id}-post-merge-closure.md"
```

Root cause: the prior 027-S/030-F closure-evidence-repair chore (PR
#73/#74, an ad hoc non-shipment-tracked chore per that session's own
memory note) named its closure artifact with an extra descriptive segment
(`-closure-evidence-repair-`) that the strict
`{shipment_id}-{feature_id}-post-merge-closure.md` filename-conformance
test does not tolerate. This is **completely unrelated** to shipment
028-S / feature 031-F (merge-strategy gate) — no code, file, or contract
surface overlap.

**P-021 disposition**: classified as out-of-scope per C1 (different
contract surface, requires a design decision between renaming the
artifact or relaxing the regex — flagged `requires_deliberation: true`).
Captured as a deferred scope-expansion stash entry rather than fixed:

* **Deferred entry ID: `DCD67C30`** (`backlogit stash add`, kind=bug,
  priority=medium/provisional)
* Discovery lookup performed first (active stash search for
  "filename conformance" / "027-S-030-F" / "PostMergeFilename") — zero
  matches, so this is a fresh capture, not a duplicate.
* No PR/review-thread existed at capture time (threadless path) — the
  entry ID is recorded here and must be cited in the eventual PR body /
  closure residual-risk record when this shipment reaches that stage.

**Why halted here instead of opening the PR**: `go test -race ./...`
runs unconditionally (not toggle-gated) in the CI `expensive`/`test` job,
which is a required member of `ci-gate`'s `needs:` array. A real `FAIL`
there fails `ci-gate` for real and is **not** an admin-fallback-eligible
block (Ship Step 5 item 17 explicitly excludes failed/missing required
checks from admin fallback). Opening the PR now would produce a red,
unmergeable required check caused entirely by out-of-scope work Ship has
no authority to fix (P-021 C1/C4) or to triage/re-prioritize (Role
Boundary — stash triage is Stage's job). Proceeding to PR creation would
either consume fix-ci cycles on an unfixable-by-Ship issue or present a
structurally blocked PR without a clear resolution path.

## Exact next operator action needed

One of:

1. **Run Stage** to triage/deliberate stash entry `DCD67C30` and land a
   tiny, separately-authorized fix on `main` (rename the artifact file to
   match the strict pattern, or adjust the conformance regex/test to
   tolerate ad hoc chore naming) — then Ship rebases this branch and
   resumes to PR creation.
2. **Explicitly authorize** Ship to perform a minimal, narrowly-scoped fix
   for entry `DCD67C30` as a distinct pre-authorized exception (per P-021
   C4, this must be a separate authorization, not a same-cycle
   expansion of shipment 028-S).
3. If the operator judges the existing regex/test itself is wrong (not
   the artifact name), direct Stage/Ship accordingly.

Once `go test ./...` is green on `main` (or the operator explicitly
accepts the residual risk and instructs Ship to proceed with a
documented `READY_WITH_FOLLOWUPS`-style exception), Ship will: rebase
`feat/028-s-merge-strategy-structural-verification-gate` if needed, run
the local review gate, run the full local build/test evidence capture,
prepare the PR body's Local Review Readiness block, and open the PR
citing deferred entry `DCD67C30` in the residual-risk section.

## Workspace state at halt

* Branch: `feat/028-s-merge-strategy-structural-verification-gate` (8
  commits ahead of `main`, working tree clean)
* Shipment 028-S: `active` (claimed, not yet shipped/closed)
* No PR opened yet for 028-S
* No merge approval requested yet — none needed at this state

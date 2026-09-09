---
title: "Post-merge closure: 014-S / 015-F — retired-architecture gate detection-quality hardening"
description: "Operational closure artifact for shipment 014-S / feature 015-F"
status: "complete"
tags:
  - "closure"
  - "014-S"
  - "015-F"
  - "post-merge"
date: 2026-09-08
mode: post-merge
shipment: 014-S
feature: 015-F
pr: 46
merge_commit_sha: d7be7e883f1d309a22037aae97ec769cb647b020
compaction_status: done
closure_status: READY
releasability: READY
conditions: []
---

# Post-merge closure: 014-S / 015-F — retired-architecture gate detection-quality hardening

- Mode: `post-merge`
- PR: #46 (`feat(014-S): retired-architecture gate detection-quality hardening (#46)`)
- Merge commit: `d7be7e883f1d309a22037aae97ec769cb647b020` (merge-commit strategy, P-009)
- Date: 2026-09-08
- Compaction status (P-020): `done` — `compact-context` invoked (target:
  all) during this closure session; bounded Tier-1 consolidation of this
  release unit's own memory (Stage's `2026-09-09-stage-014-s-session.md` +
  Ship's `2026-09-09-ship-014-s-session.md`) into
  `docs/memory/compacted/2026-09-09-014-s-015-f-retired-arch-gate-hardening-compacted.md`,
  verbose originals moved to `docs/archive/memory/`. No other memory/plan/
  closure artifacts in the repository qualified for compaction under this
  invocation's candidate rules (this release unit's own closure records are
  0 days old, below the 14-day threshold; other shipments' memory files are
  out of this bounded invocation's scope).

## Summary of the change

Hardened `scripts/check-retired-architecture.sh` (the retired-architecture
detection gate wired into CI's `lint` job) across 13 tasks
(`015.001-T`–`015.013-T`), covering stash provenance `24B63533`, `8988120D`,
`6285779C`, `8D603E29`, `9A14D3B7`, `E403E3C5`, `09CD6ACF`, `F4F4A959`:

- Tightened TOML-key and Go struct-tag/string/comment tokenization to close
  false-negative evasions (dotted/hyphenated/quoted/multiline/table-plus
  keys; raw/interpreted Go strings, block comments, rune literals, struct
  tags) and false-positive traps (benign struct tags, non-tag raw strings,
  token corpora).
  - `struct_tag_re` was additionally tightened in Copilot review round 1
    (post-merge-of-PR-review, pre-merge-of-PR) to reject digit-leading keys
    and no-whitespace-separated pairs, matching Go's real struct-tag
    grammar exactly instead of a permissive approximation.
- Added a differential Go/TOML dual-engine self-test suite and a
  `--self-test-integrity` mode, both wired into CI as unconditionally
  blocking steps.
- Flipped the CI advisory toggle default from advisory to fail-closed
  (`015.013-T`), removing the last "silently non-blocking" window for this
  gate.
- Fixed 4 findings during adversarial multi-model review before merge:
  M-1 (P0, hand-rolled TOML tokenizer infinite loop on a specific quoted-key
  edge case), M-2 (P1, fused-plural/compound-token evasion still possible
  after `015.005-T`), U-1 (P1, CI-toggle case-sensitivity doc claim
  factually wrong for GitHub Actions `vars` context), U-2 (P2, overclaim in
  a doc comment). Full detail:
  `docs/closure/2026-09-08-retired-arch-gate-hardening-adversarial-review.md`.
- Fixed 3 additional findings during GitHub Copilot PR review (round 2, all
  P-021 C1 same-contract-surface, all fixed directly, no deferrals): the
  `struct_tag_re` grammar gap above, missing YAML frontmatter on the
  adversarial-review closure doc, and a misleading fixture comment in
  `retired-differential-comment-only.go`.

## CI status and unresolved review items

- All CI checks green at final merge HEAD
  (`eb6a816c742475120580df61fcbb5efe81a2d7b2`): cross-compile ×4, lint,
  security, test, test (windows, advisory), pipeline-topology, ci gate,
  gitignore regression, detect-changes, load-cross-compile-targets.
- Copilot review (P-018 gate): `SATISFIED` — 2 rounds.
  - Round 1 (HEAD `7c00203`): 3 comment threads (struct-tag regex grammar
    gap, missing closure-doc frontmatter, misleading fixture comment). All
    classified P-021 C1 in-scope, fixed directly, gates re-run green,
    committed (`eb6a816`) and pushed, replied to each thread with the
    fixing commit, resolved each thread via `gh api graphql`
    (`resolveReviewThread`).
  - Round 2 (HEAD `eb6a816`, after explicit reviewer re-request — Copilot
    review requests are consumed after each submitted review and do not
    auto-re-trigger on push in this repo's configuration): 0 new comments.
    Gate returned `SATISFIED`.
  - No unresolved Copilot-authored threads at merge.
- Standard adversarial multi-model review (3-persona: correctness,
  architecture/maintainability, security/scope): `READY_WITH_FOLLOWUPS` →
  `READY` after M-1/M-2/U-1/U-2 were all fixed directly (all passed P-021
  C1; no deferrals were needed for this shipment).
- No unresolved P0/P1 findings at merge. No P-021 deferred-scope entries
  were created for this shipment (every finding across both review rounds
  was in-scope and fixed directly).
- Mechanical scope check (recommended as a manual follow-up in the
  adversarial review, since that review's own invocation lacked shell tool
  access): `git diff --stat` between merge-base `3a6bfff407c86815976f56b77ff6224f1948ad61`
  and merge commit `d7be7e883f1d309a22037aae97ec769cb647b020` (excluding
  `.backlogit/`) shows exactly 32 files changed — `scripts/check-retired-architecture.sh`,
  `.github/workflows/ci.yml`, the adversarial-review closure doc, and
  `scripts/testdata/**` fixtures/manifests — fully consistent with the
  declared 13-task/8-stash-ID scope. No out-of-scope files found. This
  discharges that review's residual recommendation.

## Runtime verification report

See `docs/closure/2026-09-08-014-s-015-f-retired-arch-gate-hardening-runtime-verification.md`.

**Verdict: `READY`** — unlike a library change with no wired caller, this
shipment's own surface (`scripts/check-retired-architecture.sh`) is its own
live runtime entrypoint, executed as a real GitHub Actions step
(`.github/workflows/ci.yml`'s `lint` job, verdict + `--self-test-integrity`
steps) on every PR and push to `main`. PR #46's own CI run at merge HEAD
already exercised this exact hardened surface, live, against the real
repository tree, and passed — there is no outstanding "first activation"
event, and the newly fail-closed advisory toggle (`015.013-T`) is already
armed and has already been exercised once cleanly.

## Invariants to preserve

- `scripts/check-retired-architecture.sh`'s TOML/Go tokenizers must never
  enter an unbounded loop on any well-formed or malformed input (the M-1
  regression class) — enforced going forward by the 60-second hang-detection
  guard already present in `--self-test`.
- The differential Go/TOML dual-engine self-test suite and
  `--self-test-integrity` mode must remain wired as unconditionally blocking
  CI steps (not `continue-on-error`) — this is what makes the gate's own
  correctness self-verifying on every future change to the script.
- The CI advisory toggle (`RETIRED_ARCH_GATE_ADVISORY`) defaults fail-closed
  as of `015.013-T`; flipping it back to advisory-by-default would reopen
  the exact gap this shipment closed and requires its own deliberate,
  reviewed change.
- `struct_tag_re` must continue to reject digit-leading struct-tag keys and
  keys not separated by whitespace, matching Go's real struct-tag grammar —
  regressing this reopens the narrow evasion Copilot's round-1 review found.

## Pre-deploy audits

Not applicable — this is a CI tooling/script change with no runtime service
surface, migration, flag, or config-schema change. The gate itself already
runs on every PR (including this shipment's own PR #46), so there is no
separate "pre-deploy" activation step distinct from the CI run already
evidenced above.

## Deployment / rollout path

**Merge-only, immediately active.** The change lands in `main` and takes
effect on the very next CI run of any PR or push (no phased rollout, canary,
or separate release event applies to a CI script). The fail-closed advisory
toggle flip (`015.013-T`) means this is already the enforced behavior for
every subsequent PR, starting immediately after this merge.

## Post-deploy checks

The concrete "post-deploy" check is CI itself: every subsequent PR's `lint`
job re-runs the hardened gate and its self-test suite. No additional manual
post-deploy verification step is required beyond the standard CI green-check
already gating merges repository-wide.

## Risky action record

None during implementation or review — all fixes were additive/corrective
to the gate's own detection logic and test fixtures, with no destructive or
irreversible action.

One process near-miss during post-merge closure, self-caught and corrected
before any push: the backlog-archival cascade-close commit
(`8027589`, "close shipment via cascade") was initially committed directly
on `main` instead of first creating the required
`post-merge/retired-architecture-gate-detection-quality-hardening` branch,
violating Step 6.0's NON-NEGOTIABLE branch protocol. Caught immediately by
checking `git rev-list --count origin/main..main` (=1, confirming nothing
had been pushed yet) before any push occurred. Corrected via `git branch
post-merge/retired-architecture-gate-detection-quality-hardening 8027589`
(preserving the commit), `git reset --hard origin/main` (restoring local
`main` to match `origin/main` exactly), and `git checkout
post-merge/retired-architecture-gate-detection-quality-hardening` (moving
work onto the correct branch). No data loss, no rewrite of shared/pushed
history, no policy violation actually occurred (caught pre-push) — disclosed
here for full transparency per the operator's reporting requirement.

The shipment safe-close itself used the P-015 verified fully-covered-root
cascade path (`backlogit shipment ship`), independently classified before
invocation (015-F: root, fully covered by exactly its 13 manifest tasks, no
deeper descendants, terminal, no linked deliberation) and fully verified
post-invocation (empty `returned_ids`; exact two-set `allowed_ids`/
`required_ids` match; `parent_id` preserved on all 13 tasks) — see
`.backlogit/reconcile/014-S-safe-close-2026-09-08.md`.

## Healthy signals

- `go build ./...`, `go vet ./...`, `gofmt -l .`, and `go test ./...`
  (including `tests/integration`) remain green on the post-merge closure
  branch (built on top of merge commit `d7be7e8`) — re-verified during this
  closure session.
- CI's `lint` job (which runs the hardened gate + its own self-test) passed
  cleanly on PR #46's final CI run.

## Failure signals

- A future regression that reintroduces an unbounded loop in the TOML/Go
  tokenizer (the M-1 class), caught by the 60-second hang-detection guard in
  `--self-test`.
- A future change that narrows `--self-test-integrity` coverage or flips it
  back to `continue-on-error`, silently reopening the "gate could regress
  undetected" risk this shipment closed.
- A future PR that reverts the CI advisory-toggle default back to advisory,
  silently reopening the original detection-quality gap.

## Monitoring plan

The durable monitoring for this shipment is the CI pipeline itself: every
future PR's `lint` job re-runs both the verdict scan and the
`--self-test-integrity` self-test suite unconditionally (not advisory), so
any future regression in gate correctness surfaces immediately as a CI
failure rather than silently.

## Rollback trigger / procedure

Not applicable in the operational sense (no live service to roll back). If
a future regression is found in the hardened detection logic, the remedy is
a standard follow-up PR fixing or reverting the specific change, gated
through the same review/CI process — not an operational rollback.

## Validation window

Not applicable — no live deployment window. The gate is already
continuously validated on every future PR via CI, starting immediately after
this merge.

## Owner

Repository maintainer (`softwaresalt`) for the merged change, the CI
wiring, and the shipment closure.

## Releasability evidence

**READY.** No conditions. Both the review chain (adversarial + Copilot, all
findings resolved in-scope, zero P-021 deferrals) and runtime verification
(`READY` — the gate's own live CI execution on PR #46 is direct evidence of
successful production exercise) are clean with no open caveats.

## Post-merge re-verification (on the post-merge closure branch, built on `main` after merge)

Re-ran the full quality-gate sequence on the post-merge closure branch
(`post-merge/retired-architecture-gate-detection-quality-hardening`, built
on `main` @ `d7be7e8`/`eb6a816`, prior to any closure-only source changes):
`go build ./...`, `go vet ./...`, `gofmt -l .` — all clean; `go test ./...`
(including `tests/integration`, 142s) — all green. Confirms the merge itself
introduced no regression.

## Stash follow-up items

None. No residual follow-up work was identified during implementation,
review (adversarial or Copilot), or this closure session. The mechanical
scope check recommended by the adversarial review was performed directly in
this closure session (see "CI status and unresolved review items" above)
rather than deferred as a stash item, since the tooling to perform it was
already available to Ship.

## Source artifact cleanup

Checked `custom_fields.source_stash_id` and
`custom_fields.source_deliberation_id` on both shipped top-level items
(`015-F`, `014-S`) before archival. Neither field is present on either
record. The 8 stash provenance IDs cited in the operator's directive
(`24B63533`, `8988120D`, `6285779C`, `8D603E29`, `9A14D3B7`, `E403E3C5`,
`09CD6ACF`, `F4F4A959`) were searched for as standalone backlogit artifacts
(direct `backlogit get` lookup and full-text search across `.backlogit/`) —
none resolve as separately-tracked artifacts. They exist only as prose
provenance citations inside each task's own description (e.g., "Covers
24B63533(1) + 8988120D(1)"), consistent with Stage's harvest workflow
consolidating multiple stash entries into task/feature descriptions rather
than retaining them as independently linked artifacts. This exact pattern
was also confirmed present in the immediately prior shipment 013-S's
closure (see `docs/closure/013-S-014-F-post-merge-closure.md`, "Source
artifact cleanup" section) — two independent data points now confirm this
is the norm for this repository's Stage-harvest workflow, not a defect. No
stash or deliberation archival action was applicable or taken.

## Feed Back Into the Harness

- `docs/compound/2026-09-06-requesting-copilot-review-non-collaborator-repo.md`
  was reviewed for staleness (compound-refresh assessment) and found fully
  accurate and directly confirmed by this shipment's own Copilot-review
  round-trip (the documented REST-endpoint-with-`[bot]`-suffix resolution,
  and the "requests are consumed after each review, must re-request on
  every new HEAD" caveat, both matched this session's live experience
  exactly). Classification: **keep**, no update needed.
- No other `docs/compound/` entries reference this shipment's surface
  (`check-retired-architecture.sh` or retired-architecture detection); none
  required update, consolidation, or replacement.
- No new compound entry was created for the M-1/M-2 review findings — both
  are already captured in full detail with root cause and remediation in
  `docs/closure/2026-09-08-retired-arch-gate-hardening-adversarial-review.md`,
  and neither represents a reusable pattern beyond this specific script's
  own tokenizer logic.

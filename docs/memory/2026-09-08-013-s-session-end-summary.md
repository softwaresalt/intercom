---
date: 2026-09-08
session: ship-dark-013-s-e2e
shipment: 013-S
feature: 014-F
status: complete
---

# Session-end memory — 013-S / 014-F end-to-end (DARK FACTORY MODE)

## Outcome: COMPLETE. No stop conditions fired. `main` is clean, green, and fully closed.

## PRs and merge SHAs

| PR | Title | Merge strategy | Merge commit SHA | Merged at |
|---|---|---|---|---|
| #40 | fix(pathsafe): close Windows directory-junction containment bypass (013-S/014-F) | merge-commit (P-009) | `17daffb5614f2f8ed797d7647aeadcc1dc29b1d2` | 2026-09-08 |
| #41 | chore: post-merge closure for 013-S — Pathsafe reparse-point containment hardening (014-F) | merge-commit (P-009) | `0f85d405cb72198658bd5dee07293af1a3e18de0` | 2026-09-08 23:33:18Z |
| #42 | chore: continuous-learning capture from 013-S closure session | merge-commit (P-009) | `5799acb473eb9eb892c006d71d4a0d8d7fc3eee1` | 2026-09-08 23:39:29Z |

All 3 PRs confirmed `MERGED` via `gh pr view` + `git merge-base --is-ancestor <sha> origin/main` (Merge Confirmation Gate passed each time). `main` is currently at `5799acb`, fast-forwarded, worktree clean.

## Implementation (PR #40) — strict test-first, 8/8 manifest tasks

014.001-T…014.008-T, all RED→GREEN, per `docs/plans/2026-09-08-intercom-go-pathsafe-reparse-containment-plan.md`. Closed a Windows directory-junction path-containment bypass in `internal/pathsafe`: `checkSymlinkEscape` used `filepath.EvalSymlinks`, which never resolves `IO_REPARSE_TAG_MOUNT_POINT` (junctions). Introduced `canonicalizeReparse` (direct `GetFinalPathNameByHandleW`), wired into the ancestor-climb containment loop, flipped a fail-open terminal branch to fail-closed, reconciled risk register/doc comments. No scope expansion into deferred case-folding or retired-architecture work.

## Quality gates

- `go build ./...`, `go vet ./...`, `gofmt -l .`, `go test ./...` (incl. Windows junction tests) — all green, multiple times (post-implementation, post-review-fixes, post-merge on `main`, and again after the closure/learning PRs).
- Cross-compile ×4 (darwin/amd64, darwin/arm64, linux/amd64, windows/amd64), lint, security scan, test, test-windows-advisory, pipeline-topology, ci gate — all green on every PR.

## Review

- **Standard report-only review**: 3 fixes applied pre-PR.
- **3-reviewer adversarial review**: 5 fixes applied pre-PR; 3 findings deferred under P-021 (stash `4104AF54`, `E428AB46`, `E4C5413F` — CI lint job Windows-file blind spot, `NewRoot`/`canonicalizeReparse` asymmetry on a junction-rooted workspace, long-path/UNC-prefix + `LazyProc.Call` panic-path bundle). All out-of-scope per C1, captured before PR, cited in PR body.
- **Copilot review, PR #40**: 1 round after HEAD-advance rework, 1 thread — a false-positive claim that `make([]uint16, n+1)` with `n uintptr` "will not compile." Verified false two ways (empirical: code already built/vetted/tested green; isolated `go run` repro proving `make` accepts `uintptr` size args per the Go spec). Declined with rationale, thread resolved. P-018 gate: `SATISFIED`.
- **Copilot review, PR #41 (closure)**: 1 round, 1 legitimate thread — closure artifact was missing the repo's established YAML frontmatter block. In-scope, fixed immediately (commit `3653ff0`), replied with fixing-commit citation, thread resolved. P-018 gate: `SATISFIED` on re-poll after one `REVIEW_TIMEOUT` (not a failure — Copilot simply hadn't finished yet).
- **Copilot review, PR #42 (learning capture)**: 1 round, 0 findings. P-018 gate: `SATISFIED` immediately.

## Runtime verification

**Verdict: `BLOCKED`.** Traced all callers of `internal/pathsafe` — no wired live entrypoint anywhere in `cmd/**` invokes `pathsafe.Root.Resolve()` (the exact function hardened by 013-S); `cmd/intercom`'s `RunE` is unimplemented, and `internal/config.Validate()` only calls the unrelated `pathsafe.NewRoot`. Pre-existing project-maturity gap, not introduced by this shipment. Compensating function-level evidence documented (45/45 pathsafe tests, full `go test ./...`, cross-platform build) as non-substitute evidence. Follow-up stashed: `37FAB8C2` (black-box runtime verification once a live caller exists).

## Shipment / archive state

- `013-S`: **shipped and archived** (`archived_status: shipped`), via the **P-015 verified fully-covered-root cascade path** (`classify_shipment_close_path` → `CASCADE`; 014-F is a root feature fully covered by exactly its 8 manifest tasks, no linked deliberations). Two-set gate (`allowed_ids`/`required_ids`) verified exact match; `parent_id` preserved on all 8 tasks; P-007 deleted-file guard clean. Full trail: `.backlogit/reconcile/013-S-safe-close-2026-09-08-1615.md`.
- `014-F`: **archived** (`archived_status: done`).
- All 8 tasks (014.001-T…014.008-T): **archived** (`status: archived`, `parent_id: 014-F` preserved, `commit: 17daffb5...`).
- Backlog index resynced twice (`backlogit sync`) — `CLOSURE_INDEX_SYNC_OK` both times (178 artifacts indexed after 013-S closure; not re-checked after the learning PR since it touched no `.backlogit/` queue/archive items).

## Closure / compaction state

- Post-merge closure artifact: `docs/closure/2026-09-08-013-s-014-f-pathsafe-reparse-containment-closure.md` — releasability `READY_WITH_CONDITIONS` (frontmatter added per Copilot review in PR #41).
- Compaction status (P-020): **`done`**. `compact-context` invoked (bounded, this-release-unit scope): consolidated Stage's origin/triage memory + Ship's pre-PR checkpoint into `docs/memory/compacted/2026-09-08-013-s-014-f-pathsafe-reparse-containment-compacted.md`; verbose originals moved to `docs/archive/memory/`.
- Knowledge graduation: new compound learning `docs/compound/2026-09-08-verify-go-stdlib-internals-claims-against-goroot-source.md` (verify stdlib/tool claims against actual GOROOT source or an isolated repro, not plausibility alone). No `docs/ARCHITECTURE.md` exists in this repo; `AGENTS.md` has no pathsafe references — no graduation needed there.
- Continuous-learning (this session, PR #42): 1 new observation (closure-frontmatter-omission recurring pattern, 3rd occurrence); 1 instinct file with 2 instincts, 1 **promoted** (reached the 3-occurrence threshold); 1 `evolve --mode propose` artifact — `.github/instructions/learned-closure-frontmatter-seed.instructions.md` (status: `proposed`, **not yet applied/binding** — awaits operator review or a later `evolve --mode apply`).

## Follow-up stash entries produced this shipment (all P-021, all non-blocking)

- `4104AF54` — CI `golangci-lint` job is ubuntu-only; structurally can't lint Windows-tagged production files.
- `E428AB46` — `NewRoot` vs. `canonicalizeReparse` canonicalization-mechanism asymmetry if a workspace root is itself ever a junction/GUID volume.
- `E4C5413F` — bundled: long-path `CreateFile` prefixing asymmetry, caller-side `stripUNCPrefix` reliance, theoretical `LazyProc.Call` panic path.
- `37FAB8C2` — add black-box runtime verification for `Root.Resolve()` once a live caller exists.

## Final state

- `main` @ `5799acb473eb9eb892c006d71d4a0d8d7fc3eee1`, worktree clean, `go build`/`go vet`/`gofmt`/`go test ./...` all green.
- No active shipment (P-001 satisfied — 013-S was the only shipment in this dark-run's ordered sequence, `shipment_limit: 1` honored; no other shipment claimed or executed).
- No active/unresolved backlogit checkpoints (the one `ship`-owned checkpoint on file is from a prior session and already `status: resolved`).
- No P0/P1 findings outstanding anywhere. No admin-fallback used (not needed — every merge succeeded via the normal path). No P-021 scope expansion. No named stop condition fired.

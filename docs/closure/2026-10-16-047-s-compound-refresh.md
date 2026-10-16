---
title: "Compound refresh: 047-S / 057-F"
description: "Evidence-based review of compound entries touched by the gatecheck Go hardening and test hygiene shipment."
date: 2026-10-16
mode: apply
shipment: 047-S
feature: 057-F
pr: 143
---

# Compound refresh: 047-S / 057-F

## Scope and evidence

Searched `docs/compound/` for gatecheck, retiredarch, unignore, writepath, `GIT_CONFIG_GLOBAL`,
junction and `git show` topics (20 matching files; 9 read closely). Evidence: the PR #143 merge diff
(`179b2ed`), including `unignore/git.go`, `writepath/writepath.go`, `scripts/lib/gatecheck-run.sh`
and `scripts/check-write-path-precondition.sh`, and the 047-S ship session memory. Engram and
Graphtor were not consulted; a direct targeted search was used.

## Classifications

| Entry | Classification | Evidence |
|---|---|---|
| `2026-10-02-git-show-ref-failure-ambiguity-resolve-before-show.md` | **update** | `rootGitignoreTextAt` now runs `git cat-file blob <tree>:.gitignore`, not `git show`. The core guidance is intact. An update note was added in place. The closure of the `48EA04C1` residual is marked unverified. |
| `workflow-issues/git-pathspec-glob-misses-top-level-link-named-like-the-prefix-2026-10-11.md` | **keep** | The PR changed no pathspec or scan-set selection; only containment comments, log escaping and test code changed. |
| `workflow-issues/staging-verification-commands-and-ambient-git-env-2026-10-10.md` | **keep** | The `--root` requirement and ambient-git-env lesson still hold. The `gatecheck-run.sh` change snapshots `compgen -e` before the scrub loop and fails closed; it still scrubs `GIT_*` and sets `GIT_CONFIG_GLOBAL=/dev/null`, so it adds to the lesson. |
| `2026-10-07-pin-canonical-text-maintenance-and-unbounded-ast-rule-review.md` | **keep** | This PR touched retiredarch tests and small source edits, but the pin and `canonicalDecls` maintenance guidance is not contradicted. |
| `2026-09-08-adversarial-review-empirical-verification-and-scope-check.md` | **keep** | Historical junction case. The new Windows junction test adds coverage and does not supersede it. |
| `2026-09-08-verify-go-stdlib-internals-claims-against-goroot-source.md` | **keep** | `EvalSymlinks` junction behavior is unchanged. |
| `2026-09-30-go-msys-bash-bridging-and-unicode-case-mapping-gap.md` | **keep** | The 047-S session reproduced the WSL-bash-shadowing warning. |
| `2026-09-28-extracting-bash-heredoc-python-engines-into-importable-modules.md`, `2026-09-30-shipment-reconcile-skill-bypass-manual-substitution-risk.md`, `2026-10-01-textual-yaml-guard-denylist-bypass-loop.md`, `2026-10-02`/`2026-10-08` cascade and staging entries, `workflow-issues/staging-premise-verification-before-planning-2026-09-18.md` | **keep** | Matched the search terms only incidentally; not affected by this shipment. |

## Changes and follow-up

* Updated: `docs/compound/2026-10-02-git-show-ref-failure-ambiguity-resolve-before-show.md`.
* No entry was consolidated, replaced, archived, deleted or marked stale.
* Follow-up: confirm by test that a directory named `.gitignore` now errors under `cat-file blob`, then
  drop the hedge in the update note.
* No stash or backlog item was changed.

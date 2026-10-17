---
title: "Compound refresh: 048-S / 058-F"
description: "Evidence-based review of compound entries touched by the scripts, CI and Windows test fixes shipment."
date: 2026-10-16
mode: apply
shipment: 048-S
feature: 058-F
pr: 145
---

# Compound refresh: 048-S / 058-F

## Scope and evidence

Searched `docs/compound/` for WSL, bash launcher, `toBashPath`, MSYS, write-path, closure filename,
toolchain pin and Windows path topics. Evidence: the PR #145 merge diff (`60415750`), including
`tools/gatecheck/runner_test.go` (`requireWindowsCompatibleBash`, `toBashPath`),
`scripts/check-write-path-precondition.sh`, `tools/gatecheck/internal/writepath`, the closure-filename
conformance test, `.github/workflows/ci.yml` and the toolchain-pin decision note annotation, plus the
048-S ship session memory. Engram and Graphtor were not consulted; a direct targeted search was used.

## Classifications

| Entry | Classification | Evidence |
|---|---|---|
| `2026-09-30-go-msys-bash-bridging-and-unicode-case-mapping-gap.md` | **update** | Sections 1 and 2 are partly superseded: `toBashPath` emits `/c/...`, and `requireWindowsCompatibleBash` skips with a reason under the WSL launcher. An update note was added in place; the PATH workaround and prevention checklist remain valid. |
| `2026-10-08-powershell-backtick-control-chars-and-stdlib-toolchain-pin.md` | **keep** | The toolchain-pin rule still holds. The shipped change only annotated the decision note's stale `go1.26.5` example (now `go1.26.9`); the entry cites no pin value. |
| `2026-10-07-pin-canonical-text-maintenance-and-unbounded-ast-rule-review.md` | **keep** | Pin maintenance guidance is not contradicted. |
| `2026-10-02-git-show-ref-failure-ambiguity-resolve-before-show.md` | **keep** | The 047-S update note stands; 048-S did not touch `unignore`. |
| `workflow-issues/staging-verification-commands-and-ambient-git-env-2026-10-10.md` | **keep** | Ambient git env and `--root` lessons unchanged. The new `::notice::write-path gate mode=X` line only adds output. |
| `2026-09-08-adversarial-review-empirical-verification-and-scope-check.md`, `2026-09-08-verify-go-stdlib-internals-claims-against-goroot-source.md` | **keep** | Matched incidentally; unaffected. |
| Other entries (cascade, staging, textual-YAML guard, Copilot review) | **keep** | No overlap with this shipment. No compound entry covers closure-filename conformance. |

## Changes and follow-up

* Updated: `docs/compound/2026-09-30-go-msys-bash-bridging-and-unicode-case-mapping-gap.md`.
* No entry was consolidated, replaced, archived, deleted or marked stale.
* Follow-up (optional): the 048-S review loop may merit a new compound entry on WSL-launcher skip gating;
  the existing update note covers it for now.
* No stash or backlog item was changed.

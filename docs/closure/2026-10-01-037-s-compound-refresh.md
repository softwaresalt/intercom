---
title: "compound-refresh: 037-S / 047-F post-merge closure"
date: 2026-10-01
mode: apply
scope: recent
shipment: 037-S
---

# Compound Refresh — 037-S / 047-F post-merge closure

## Entries reviewed

| Entry | Classification | Evidence / Action |
|---|---|---|
| `2026-09-30-shipment-reconcile-skill-bypass-manual-substitution-risk.md` | **update** | 037-S's closure performed the fresh recompute-and-compare before the cascade. The binding `0f9be8a3…92d9` matched at 18:20:12Z, and the mutation ran only after that match. This is the first closure to apply item 2 of the guidance. A new section, "Partially applied in 037-S", records this along with the unchanged residual: no invokable `safe-close`, and the recompute was agent-executed rather than tool-enforced. Stash `D10D3AFC` stays relevant. The core guidance is unchanged. |
| `2026-09-30-copilot-review-body-prose-signal-vs-gate-verdict.md` | **keep** | Confirmed again. Several PR #87 rounds raised findings only as "previously missed" notes in the review overview, with no inline thread. Treating the overview text as a signal, and the `autoharness gate copilot-review` verdict as the gate, handled them correctly. |
| `2026-09-11-copilot-suppressed-comments-vs-review-threads.md` | **keep** | Still accurate. It is a separate pattern, and PR #87 did not contradict it. |
| `2026-09-30-go-msys-bash-bridging-and-unicode-case-mapping-gap.md` | **keep** | Confirmed. The M4 tests and parity runs again needed Git-Bash on `PATH`, because plain `bash` on the Windows host is WSL (m4.md §6 and §8.1). |
| `2026-09-28-extracting-bash-heredoc-python-engines-into-importable-modules.md` | **keep** | Historical. M4 deleted the modules this entry describes extracting (`scripts/lib/*.py`). The entry still accurately records how 029-S worked, and nothing supersedes its guidance for future heredoc extractions. No edit. |
| `2026-09-28-bash-python-heredoc-stdin-and-github-token-admin-scope.md` | **keep** | Historical, for the same reason. The Python heredocs are gone, but the GitHub-token admin-scope part still applies to `check-merge-strategy.sh`. |

## New compound capture

* `docs/compound/2026-10-01-textual-yaml-guard-denylist-bypass-loop.md` (new). M4's textual
  CI-YAML guards were written as denylists of bad spellings, and that drove 22 Copilot rounds of
  equivalent-form bypasses. The lesson: assert a canonical-shape allowlist first and pin required
  wiring positively. If a guard's review loop keeps finding new spellings past the cycle limit,
  change the design.

## Files updated

* `docs/compound/2026-09-30-shipment-reconcile-skill-bypass-manual-substitution-risk.md`:
  appended the "Partially applied in 037-S" section.
* `docs/compound/2026-10-01-textual-yaml-guard-denylist-bypass-loop.md`: created.

## Follow-up items requiring manual review

None. No entry needed consolidation, replacement or deletion. Stash `D10D3AFC` remains with
Stage as the structural fix for the reconcile entry's residual.

---
title: "compound-refresh: 036-S / 046-F post-merge closure"
date: 2026-10-01
mode: apply
scope: recent
shipment: 036-S
---

# Compound Refresh — 036-S / 046-F post-merge closure

## Entries reviewed

| Entry | Classification | Evidence / Action |
|---|---|---|
| `2026-09-30-shipment-reconcile-skill-bypass-manual-substitution-risk.md` | **update** | **Corrected 2026-10-01, post-Copilot-review on PR #86.** An earlier draft of this refresh (and the entry itself) incorrectly classified 036-S's closure as a successful application of the entry's guidance. It is not: 036-S's closure computed a `CLASSIFICATION_BINDING` but then invoked `backlogit shipment ship` directly, with no intervening `mode: safe-close` call to recompute/compare it — the same root-cause category of gap as 035-S (missing live safe-close revalidation), not a fix. The entry now has a "Recurred in 036-S" section replacing the inaccurate "Applied successfully" section, and the closure artifact's Risky Action Record / Releasability evidence now accurately disclose the gap instead of claiming it closed. A new stash entry (`D10D3AFC`) was captured proposing the structural remediation (an actually-invokable `safe-close` CLI/gate) for Stage's deliberation queue. Core guidance (the "Fix / Prevention" section) remains accurate and unchanged — it correctly predicted this exact failure mode in its item 2; only the classification of 036-S's own execution was wrong and has been corrected. |
| `2026-09-30-copilot-review-body-prose-signal-vs-gate-verdict.md` | **keep** | Still accurate. 036-S's own Copilot loop reused the same pattern (programmatic `autoharness gate copilot-review` verdict, plus the `<!-- ccr-overview-v2 -->` "Findings: None" structured summary, as the authoritative convergence signal) with no contradicting evidence. No update needed. |
| `2026-09-30-go-msys-bash-bridging-and-unicode-case-mapping-gap.md` | **keep** | Not directly exercised this shipment (M3 did not hit the WSL-bash-vs-Git-Bash `PATH` ambiguity or Unicode case-mapping gap in session). Remains accurate for M1/M2's scope; no contradicting or superseding evidence surfaced. |
| `2026-09-28-extracting-bash-heredoc-python-engines-into-importable-modules.md` | **keep** | Describes 029-S's Python-heredoc-to-module extraction pattern specifically. M3 is a different migration shape (Python-to-Go full port, not heredoc-to-module), so this entry's scope and guidance remain distinct and unaffected. No overlap requiring consolidation. |
| `2026-09-28-bash-python-heredoc-stdin-and-github-token-admin-scope.md` | **keep** | Same rationale as above — distinct scope (heredoc stdin capture / token admin-scope), not touched by M3's work. |
| `2026-09-11-copilot-suppressed-comments-vs-review-threads.md` | **keep** | Distinct pattern (labeled "Suppressed comments" section vs. unlabeled narrative prose, per the 2026-09-30 prose-vs-verdict entry's own cross-reference). Not touched by this shipment's review cycle (036-S's Copilot findings all surfaced as normal inline review-thread comments, not suppressed-comment or narrative-only findings). |

## New compound capture considered, not created

This shipment's review cycle (persona + adversarial + Copilot) surfaced several genuine
findings, but all were either (a) faithful Python-parity confirmations (not novel
problem/solution pairs worth a new compound entry — they are captured as P-021 stash follow-ups
instead, which is the correct venue for deferred-scope findings) or (b) already covered by
existing compound entries (Copilot review-loop mechanics). No new compound entry was warranted
for this shipment's own work beyond the update above.

## Files updated

* `docs/compound/2026-09-30-shipment-reconcile-skill-bypass-manual-substitution-risk.md` —
  replaced the inaccurate "Applied successfully in 036-S" section with an accurate "Recurred in
  036-S" section (see table above and the self-correction note below).

## Self-correction (2026-10-01, post-Copilot-review on PR #86)

This refresh report, the compound entry it updated, the closure artifact
(`docs/closure/036-S-046-F-post-merge-closure.md`), the cascade-close reconcile report
(`.backlogit/reconcile/036-S-cascade-close-2026-10-01T02-25-49Z.md`), the compacted memory
(`docs/memory/compacted/2026-10-01-036-s-046-f-compacted.md`), and the archived session-memory
checkpoint (`docs/archive/memory/2026-10-01/036-s-046-f-post-merge-closure-session-memory.md`)
all originally claimed that 036-S's shipment closure followed a "full bound sequence" and closed
the 035-S-disclosed process-compliance gap. Copilot review on PR #86 (comments
`4151321043`/`4151321103`/`4151321159`/`4151321194`/`4151321224`/`4151321255`) correctly
identified that this claim was inaccurate: the actual sequence computed a binding during
classification but then called `backlogit shipment ship` directly, without the mandated
`mode: safe-close` fresh-recompute-and-compare revalidation step in between. All six affected
documents were corrected in place (not silently rewritten — each correction is annotated and
traceable) to disclose the gap accurately instead of claiming it closed. A new stash entry
(`D10D3AFC`) captures the structural remediation recommendation for Stage's deliberation queue.
The resulting backlog archive state's data integrity (two-set gate, parent-ID preservation,
archive provenance) is independently verified and unaffected by this correction — only the
narrative characterization of the *procedure* that produced it was wrong and has been fixed.

## Follow-up items requiring manual review

All reviewed compound entries are either confirmed accurate (`keep`) or updated with
evidence-backed corrections (`update`); no entry required consolidation, replacement, or
deletion. One structural follow-up is flagged for Stage's deliberation queue: stash `D10D3AFC`
(high priority, requires deliberation) proposes implementing an actually-invokable
`shipment-reconcile mode: safe-close` CLI/gate subcommand, since this is now the second
consecutive shipment closure (035-S, 036-S) to disclose the same category of process-compliance
gap from manual approximation of that step.

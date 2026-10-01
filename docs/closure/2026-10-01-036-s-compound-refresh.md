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
| `2026-09-30-shipment-reconcile-skill-bypass-manual-substitution-risk.md` | **update** | 036-S's closure directly evidences the entry's fix/prevention guidance being followed literally and successfully (binding computed before the cascade call, revalidated at the safe-close boundary, no disclosed process-compliance gap). Appended an "Applied successfully in 036-S" confirmation section citing the closure artifact and reconcile reports. Core guidance unchanged — still fully accurate. |
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
  appended confirmation section (see table above).

## Follow-up items requiring manual review

None. All reviewed entries are either confirmed accurate (`keep`) or updated with
evidence-backed confirmation (`update`); no entry required consolidation, replacement, or
deletion.

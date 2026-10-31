---
title: "075-S / 082-F compacted Ship handoff"
description: "Consolidated Ship and Stage record for Copilot cli_path classification and resolution hardening and [remote] config validation"
date: 2026-10-31
status: shipped
tags: [memory, compacted, cli-path, remote-config, 075-S]
source:
  - docs/archive/memory/2026-10-31/075-s-ship-memory.md
  - docs/archive/memory/2026-10-30/stage-write-preview-and-cli-path-safety-memory.md
---

## Staging

* Stash 773B2CB2 and 70315264 fed 074-S / 081-F. Stash 56CE7077, FD093EFC and FC76C5D8 fed 075-S / 082-F. All five were archived by Stage
* Plan: `docs/plans/2026-10-30-intercom-go-write-preview-and-cli-path-safety-plan.md` (left in place; archived backlog items reference it)
* Deliberation: `docs/decisions/2026-10-30-intercom-go-write-preview-and-cli-path-safety-deliberation.md`

## Execution

* 082.001-T `a416b85`: shared `pathsafe.ClassifyCLIPath`
* 082.002-T `f12cfa4`: `internal/config` `validateCLIPath` hardening
* 082.003-T `e9c3f65`: `internal/agent/copilot` `resolveCLIPath`
* 082.004-T `4069cd4`: `[remote]` unknown-key rejection, hex IPv4 host rejection, `DefaultRemoteMaxConnections` locked to `hub.DefaultMaxRemote`
* 082.005-T `f1ee9c5`: design doc and config reference
* Review follow-ups `c019210`. PR 225 merged by merge commit `a2898942a12c70fe23be050d698ae4622153d97f`; all 14 CI checks green, P-018 not applicable

## Decisions

* Hex-host origins were already rejected by `url.Parse`; the test asserts rejection without a specific message
* Adversarial Review could not dispatch sub-reviewers. The gap was disclosed in the PR body instead of claiming consensus
* Deferred scope captured under P-021 C2: stash 2C45EFAE (shared resolution helpers) and 2F1B6D4B (residual cli_path hardening, requires deliberation)

## Closure

* Step a1 moved 082-F to `done`; classify-only selected CASCADE (engine VERIFIED)
* Operator approved the mutating cascade-close at 2026-10-31T02:50:44Z. The run exited 0 with `postcondition_verdict: pass`, `shipment_record_archived_status: shipped`, seven records archived and `returned_ids` empty
* Post-mode reconcile: `.backlogit/reconcile/075-S-post-done.md`, `PROCEED`. Evidence: `docs/closure/evidence/075-S-082-F-close-evidence.json`

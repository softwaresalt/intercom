---
title: "Stage session memory: intercom MVP primary scenario DAG (056-S to 062-S)"
date: 2026-10-19
branch: chore/stage-intercom-mvp-primary-scenario
---

* Base: origin/main 28b4850. Branch gate passed: single worktree, clean tree (a CRLF-only
  phantom change on `.backlogit/stash.jsonl` was cleared with `git add`, no content diff),
  branch created from `refs/remotes/origin/main`.
* Request: operator reprioritized an MVP of the full primary scenario ahead of queued
  deliberations and the gate/tooling residual stash. Normal pipeline, not dark mode.
* Traced stash: A92E3FA0 (annotated, not consumed). No other stash entry consumed,
  re-prioritized, or archived. Duplicate scan: clean.
* Deliberation: docs/decisions/2026-10-19-intercom-go-mvp-primary-scenario-deliberation.md
  (MVP-D1 to D10; operator-blocking: D3 auth, D4 authorization, D6 tunnel anonymous-off).
* Plan: docs/plans/2026-10-19-intercom-go-mvp-primary-scenario-plan.md. Hardening: yes.
  Review: attempt 1 Architecture ADVISORY, Security FAIL; revised; attempt 2 Architecture
  PASS, Security ADVISORY; advisories applied. Gate PASS.
* Harvest: features 064-F..070-F; tasks 064.001-T..070.006-T (46 tasks, 89 task blocks
  edges, size structured, complexity recorded as prose). Shipments 056-S..062-S with blocks
  edges 057<-056, 058<-057, 059<-058, 060<-059, 061<-059, 062<-060, 062<-061.
  043-F stays blocked as umbrella; `informs` links 043-F to each new feature.
* Task map: S1 064.001-064.004; S2 065.001-065.007; S3 066.001-066.009;
  S4 067.001-067.008; S5 068.001-068.008; S6 069.001-069.004; S7 070.001-070.006.
  Plan key to ID: 5.8 = 068.004, 5.4 = 068.005, 5.5 = 068.006, 5.7 = 068.007, 5.6 = 068.008;
  7.6 = 070.006; other keys follow numeric order within the feature.
* Degradations: the task type exposes no `complexity` field in backlogit 1.11.0 (CLI
  error), so complexity is prose in the description. No queue_position control exists
  for shipments; queue order is shipment ID order plus blocks edges.
  `dep add` took about 35 s per edge (89 edges); creation was scripted from files in
  `C:\Windows\Temp` (outside the repo).
* Eligible first: 056-S (057-S to 062-S are blocked by predecessors).
* Not claimed. Operator must push `chore/stage-intercom-mvp-primary-scenario` and merge its
  staging PR (merge commit) before Ship runs; Ship reads the backlog from main.
* Deferred: all other stash entries, 003-DL to 005-DL (no MVP dependency).

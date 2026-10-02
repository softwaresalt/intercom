---
title: "Compound refresh: 038-S / 048-F"
description: "Evidence-based review of compound entries relevant to gatecheck CLI containment hardening."
date: 2026-10-02
mode: propose
shipment: 038-S
feature: 048-F
pr: 91
---

# Compound refresh: 038-S / 048-F

## Scope and evidence

Reviewed compound entries matching the shipped gatecheck, root-containment,
wrapper, argument, and Windows Bash surfaces. The Engram daemon was unavailable
during this post-merge retrieval attempt, so indexed retrieval was degraded.
Graphtor's registered sources did not include the repository's local
`docs/compound/`; a direct targeted search of that directory was used instead.
The shipped Unit C plan and `scripts/check-write-path-precondition.sh` were read
to confirm the current contract and implementation.

## Classifications

| Entry | Classification | Evidence |
|---|---|---|
| `docs/compound/2026-09-30-go-msys-bash-bridging-and-unicode-case-mapping-gap.md` | **keep** | Its Windows shell-selection warning remains accurate and was reproduced locally: bare `bash` resolved to WSL Bash, while prepending Git for Windows Bash made wrapper tests pass. This shipment did not change shell resolution, path translation, or the Unicode compatibility code. |
| `docs/compound/2026-10-01-textual-yaml-guard-denylist-bypass-loop.md` | **keep** | The note addresses textual CI-YAML shape guards and their review-loop behavior. Unit C changes `parseRoot` and the write-path wrapper; it does not change CI YAML or those structural guard tests. Its advice and evidence remain distinct and current. |

## Changes and follow-up

No compound entry was updated, consolidated, replaced, archived, or deleted.
The shipment neither supersedes nor contradicts these learnings. No additional
compound entry was found to cover the exact trusted-root/duplicate-root contract
in a way that this work would duplicate. No Stage-owned stash or deferred-scope
entry was changed.

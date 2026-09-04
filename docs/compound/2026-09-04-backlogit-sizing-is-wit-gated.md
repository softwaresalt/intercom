---
title: "backlogit sizing is WIT-gated: features.sizing:true does not mean size/complexity are settable"
date: 2026-09-04
category: tooling-constraints
tags: [backlogit, sizing, harvest, stage, degradation]
---

# backlogit sizing is WIT-gated

## Symptom

`.autoharness/backlog-registry.yaml` advertises `features.sizing: true` and
declares `size`, `complexity`, `size_source`, and `size_ruleset_version` as
`update_task` params. Following the harness contract's structured-emission path
therefore looks correct — but the calls fail, and **`2>$null` hides it**:

```text
$ backlogit update 005.002-T --complexity high
Error: set complexity: validate complexity for artifact 005.002-T:
       artifact type "task" does not define a complexity field: backlogit: validation error
exit 1

$ backlogit update 005.002.004-ST --size M --size-source agent --size-ruleset-version v1
Error: artifact type "subtask" does not define a size field: backlogit: validation error
exit 1
```

Silent partial success is the dangerous shape: on **tasks**, `size` /
`size_source` / `size_ruleset_version` *do* persist into `custom_fields`, so
`backlogit queue view` shows a populated SIZE column and the run looks healthy
— while `complexity` was rejected on every item and `size` was rejected on
every subtask.

## Root cause

Sizing support is **per-artifact-type**, resolved from the workspace's WIT
(work-item-type) metadata — not from the registry's `features` block. The
registry flag advertises that the *tool build* supports sizing; it says nothing
about whether *this workspace's* artifact types define the fields. In the
`intercom-go` workspace:

| Artifact type | `size` | `complexity` |
|---|---|---|
| task | defined | **not defined** |
| subtask | **not defined** | **not defined** |

So `features.sizing: true` is necessary but not sufficient.

## Fix

1. **Probe before bulk-applying.** Run one `update` per artifact type with
   stderr visible and check `$LASTEXITCODE`. Do not wrap sizing calls in
   `2>$null` during a bulk loop — you will record a success count that is
   entirely fictional.
2. Inspect the actual WIT metadata rather than trusting the registry flag:
   `backlogit_get_wit_metadata` (MCP) or `backlogit list-types`.
3. When a field is undefined for a type, fall back to the harness's
   non-structured path: keep the value **enum-validated**, write it as labeled
   prose in the artifact body (`Size: M | Complexity: medium (source: agent;
   ruleset: ...)`), and **flag the degradation explicitly** in the Stage report.
   Do not skip assigning the values, and do not halt.
4. Verify by reading the artifact file, not the list view — `queue view` will
   happily show a size that came from a *different* call than the one you think
   populated it.

## Related gotchas found in the same session

* **Dependency type.** `backlogit dep add --type blocked-by` is rejected
  (`invalid dependency type`). The valid default is `blocks`, so express the
  edge in the `blocks` direction: `dep add <dependent> <prerequisite>` yields
  `dependent → prerequisite (blocks)`.
* **`stash edit --text` replaces the whole body.** There is no append mode. To
  annotate an entry, read it back with `stash get`, concatenate, and write the
  full text — and make the operation idempotent by checking for your own marker
  first, or a re-run will duplicate the annotation.
* **`backlogit add` fails silently on a malformed invocation.** A call whose
  flags don't parse can still create a bare artifact (title only, no
  description) and print nothing, at the next free ID. Check the printed
  `Created <id>` line; if it is absent, list the queue directory before
  retrying, or you will orphan an artifact (this session orphaned `004-F` that
  way and had to archive it).
* **`git add` aborts the whole invocation on one nonexistent pathspec.** Adding
  a list that contains a directory which does not exist stages *nothing*, and
  the error is easy to swallow. Stage explicit paths and verify with
  `git diff --cached --name-status` before committing.

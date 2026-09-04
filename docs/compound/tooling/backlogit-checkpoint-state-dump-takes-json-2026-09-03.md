---
title: "backlogit checkpoint create --state-dump takes JSON content, not a file path"
description: "Passing a path to --state-dump writes the path string into the checkpoint, producing a corrupt record that fail-closes the next session's recovery scan."
source: "docs/compound/tooling/backlogit-checkpoint-state-dump-takes-json-2026-09-03.md"
doc_type: "learning"
problem_type: "tooling-misuse"
category: "tooling"
component: "backlogit CLI / Stage crash-resumption protocol"
root_cause: "Stage invoked `backlogit checkpoint create --state-dump <tempfile.json>` assuming the flag accepted a path. The CLI treats the argument as the state-dump value itself, so the literal Windows path string was persisted where a JSON object was expected."
resolution_type: "process_change"
severity: "medium"
message: "resolve checkpoint: backlogit: checkpoint file corrupt or unparseable: invalid character 'C' looking for beginning of value"
file_path: ".backlogit/checkpoints/"
citations:
  - "docs/memory/2026-09-03-stage-p3-credentials.md"
  - "docs/plans/2026-09-03-intercom-go-p3-credentials-plan.md"
tags:
  - "backlogit"
  - "checkpoints"
  - "stage"
  - "crash-resumption"
  - "windows"
---

# `backlogit checkpoint create --state-dump` takes JSON content, not a file path

## Problem

During the P3 credentials Stage session, a mid-session checkpoint was written with:

```powershell
backlogit checkpoint create --state-dump $tmp   # $tmp = C:\Users\...\Temp\ckpt1.json
```

The command reported success and returned a checkpoint path, so the failure was silent. The
written file was 47 bytes containing the literal path string rather than the intended payload.

The defect only surfaced at session end, when `backlogit checkpoint resolve` failed:

```text
resolve checkpoint: backlogit: checkpoint file corrupt or unparseable:
invalid character 'C' looking for beginning of value
```

(`'C'` is the drive letter of the temp path.)

## Why it matters

This is worse than a lost checkpoint. Stage's crash-resumption protocol is **fail-closed**: at
session start it enumerates *all* checkpoint summaries with no `status`/`agent` filter, precisely
so that a parse-failed record — which is returned with an empty `agent` and empty `status` — cannot
be silently filtered out. `backlogit checkpoint list` correctly reported:

```json
{ "filename": "checkpoint-20260904-023022.json", "agent": "", "status": "",
  "validation_error": "...corrupt or unparseable...", "needs_quarantine": true }
```

A record in that state **halts the next session to operator handoff before any work begins**. So a
silently-malformed checkpoint turns a routine next-session startup into a blocked one, and the
operator has no obvious link back to the session that wrote it.

## Resolution

Pass the JSON **content**, not a path. On PowerShell, build the object and convert inline:

```powershell
$dump = @{ schema_version = 1; agent = 'stage'; session_id = '...'; phase = '...';
           resume_hint = '...'; context = @{ } } | ConvertTo-Json -Depth 6 -Compress
backlogit checkpoint create --state-dump $dump
```

Then **verify immediately** rather than trusting the success message:

```powershell
backlogit checkpoint list   # needs_quarantine must be 0
```

If a corrupt checkpoint already exists, `resolve` cannot clear it (resolution parses first). Either
run `backlogit checkpoint quarantine '<file>' --reason '<reason>'` — the remediation command the
tool itself returns in `remediation_command` — or, for a local-only artifact under the gitignored
`.backlogit/checkpoints/`, delete the file. Re-run `checkpoint list` and confirm `total: 0` and
`needs_quarantine: 0` before ending the session.

## Prevention

- Treat "the command printed a path" as **not** evidence of a valid checkpoint. The create call
  succeeds regardless of payload validity.
- Add a `checkpoint list` validation read after every `checkpoint create`, and again at session
  end alongside the existing index sync.
- Never end a Stage or Ship session leaving `needs_quarantine > 0`; that state is a startup
  blocker for the next session, not a cosmetic warning.
- Generally: for any CLI flag accepting structured data, confirm from `--help` whether it takes a
  literal value, a path, or `@file` syntax before wiring it into a protocol step. Here `--help`
  describes `--state-dump` as the state dump itself.

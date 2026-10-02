---
title: "A stale backlogit binary earlier on PATH can write an index the current engine cannot read"
date: 2026-10-02
category: tooling
tags: [backlogit, path, index, environment, shipment-reconcile, 030-s]
---

# A stale backlogit binary earlier on PATH can write an index the current engine cannot read

## Symptom

During 030-S post-merge closure, `backlogit shipment get 030-S` (engine 1.11.0) failed with
an `unmarshal dependencies` error before `shipment-reconcile` pre-mode could start.

## Root cause

Two backlogit binaries were installed:

* `C:\Tools\backlogit.exe`: 1.11.0, the intended engine.
* `%USERPROFILE%\go\bin\backlogit.exe`: a stale 1.5.0.

A shell whose `PATH` resolved the stale 1.5.0 first had rewritten the SQLite index cache. In
that schema, `dependencies` is serialized differently from 1.11.0's list of `{id, type}`
objects. The Markdown records were never affected, because the index is a disposable cache.

## Fix

1. Put the intended engine first: `$env:PATH = "C:\Tools;C:\Program Files\Git\bin;$env:PATH"`.
2. Confirm it with `backlogit --version`. It must report the engine you expect.
3. Rebuild the cache with `backlogit sync`, run by that engine.

Do not hand-edit `.backlogit/backlogit.db`.

## Prevention

* Before any safe-close, record the engine line (`backlogit --version`) next to the
  `CLASSIFICATION_BINDING`. The binding embeds the engine line, so a binary switch between
  classify and safe-close shows up as drift instead of passing silently.
* Code that parses `dependencies` from CLI JSON should read `.id` from each element, not
  assume plain strings.
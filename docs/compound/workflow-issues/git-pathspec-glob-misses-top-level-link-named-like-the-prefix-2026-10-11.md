---
title: "Git pathspec `internal/**` does not list a tracked link named exactly `internal`"
date: 2026-10-11
category: workflow-issues
problem_type: gate false-clean (scan-set selection)
tags: [git, pathspec, symlink, gitlink, gatecheck, writepath, retiredarch]
---

# Git pathspec `dir/**` does not list a tracked link named exactly `dir`

## Problem

A gate that selects its scan set with `git ls-files -- internal/** cmd/**` and then judges each
returned record would be expected to see every tracked entry under those directories. Two gaps were
found while staging write-path Batch 2 (stash EEE39D23, plan-review finding SEC-1):

1. A tracked symlink or gitlink whose name is **exactly** `internal` or `cmd` (mode 120000 or
   160000) is **not** returned by the glob pathspecs `internal/**` and `cmd/**`. The literal
   pathspecs `internal` and `cmd` return it. A top-level `cmd -> ../pkg` link is therefore never
   judged by a containment check that only runs on returned records.
2. Records returned for a non-`.go` name (for example a symlink `internal/fx` or a gitlink
   `internal/sub`) were dropped silently by `shouldScan` before any containment check ran.

## Probe (reproduce before scoping)

Create the index entries without needing OS symlink privilege:

```text
blob=$(printf 'pkg' | git hash-object -w --stdin)
git update-index --add --cacheinfo 120000,$blob,cmd
git update-index --add --cacheinfo 160000,<40-hex>,cmd2
git ls-files -- 'internal/**' 'cmd/**'     # lists neither `cmd` nor `cmd2`
git ls-files -- internal cmd                # lists `cmd`
```

A link named `internal` cannot coexist with `internal/ok.go` in one index (directory/file
conflict), so a test needs one repository per top-level row.

## Resolution

List the literal directories (`-- internal cmd`; identical bytes on a normal tree) and apply the
containment check to every listed record, not only the selected `.go` ones. Plan review also found
that a `GIT_LITERAL_PATHSPECS` env-isolation vector becomes vacuous once the pathspecs are literal,
so a vector that still changes the literal listing (`GIT_ICASE_PATHSPECS`) is needed.

## Prevention

* Probe the exact listing shape (`ls-files -s`) before trusting a glob pathspec as "everything under".
* Judge the shape of every record the runner returns; select by shape, not by a named spelling.
* The retired-architecture engine has the same gap (stash `284270E8`); its selection is frozen by
  `pin.go`, so the fix needs an atomic refreeze of `select.go` and `pin.go` `canonicalDecls`.

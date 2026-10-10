---
title: "Probe the verification commands a plan prescribes, and scrub ambient git env before probing git vectors"
description: "Two staging lessons from gatecheck Batch B1: a plan listed gatecheck commands without the mandatory --root, and an agent host's ambient GIT_* made a config vector non-live while a partial-index vector bypassed an empty-selection guard"
source: "docs/compound/workflow-issues/staging-verification-commands-and-ambient-git-env-2026-10-10.md"
doc_type: "learning"
problem_type: "plan verification commands unprobed; ambient environment skews a probe"
category: "workflow-issues"
component: "tools/gatecheck (write-path gate), Stage planning"
root_cause: "the plan copied verification commands from a shipped precedent without checking the CLI contract; probes of git environment variables inherited the agent host's own GIT_* variables"
resolution_type: "workaround"
severity: "medium"
message: "--root is required"
file_path: "tools/gatecheck/main.go"
citations:
  - "docs/plans/2026-10-10-intercom-go-gatecheck-batch-b1-write-path-input-hardening-plan.md"
  - "docs/decisions/2026-10-10-intercom-go-gatecheck-batch-b1-write-path-input-hardening-deliberation.md"
  - "docs/plans/2026-10-08-intercom-go-gatecheck-batch-a-correctness-plan.md"
tags:
  - "stage"
  - "gatecheck"
  - "plan-review"
  - "git-env"
  - "verification-commands"
---

## Problem

1. **Unprobed verification commands.** The Batch B1 plan (attempt 1) listed
   `go run ./tools/gatecheck write-path` and its `--self-test` variants as the proof that the live tree
   still passes. `tools/gatecheck/main.go` (`parseRoot`) rejects a missing `--root`
   (`--root is required`), so none of those commands could validate anything. The same commands had
   been copied from the shipped Batch A plan, where nobody had run them literally. Only the
   Architecture Strategist (anchor route) caught it.
2. **Ambient git environment skews probes.** The agent host exports `GIT_CONFIG_COUNT=3` and
   `GIT_CONFIG_KEY_0..2`. A staging probe of the "`GIT_CONFIG_COUNT=1` with no key" vector therefore
   exited 0: the key was not missing, so the vector looked dead. Separately, a `GIT_INDEX_FILE` that
   points at a partial index made `git ls-files` print a **non-empty, partial** listing, which the
   gate's empty-selection guard (ED-7) cannot detect.

## Root Cause

1. A precedent's command block was treated as a contract. A CLI's required flags are part of its
   contract and live in its source, not in a neighbouring plan.
2. An empty-selection guard is not a completeness guard. And a probe that does not strip the host
   environment measures the host as much as the vector.

## Resolution

1. Read the CLI's argument parser (`parseRoot`) and added `--root .` to every direct invocation; added
   the wrapper (`scripts/check-write-path-precondition.sh`) as the end-to-end check because it derives
   the root itself.
2. Probed each vector in a scratch repo under the git-ignored `logs/`, with exit codes captured outside
   a pipeline, and wrote the observed outcome per vector into the deliberation. Tests that follow must
   call a hermetic-env helper first (strip every `GIT_*`, case-folded, then add only the vector), and
   compare listing bytes with a scrubbed baseline rather than comparing exit codes.

## Prevention

* Before a plan prescribes a CLI command as verification, read the entry point's flag parsing (Stage may
  not run builds, but it can read `main.go`) and copy the usage line, not a precedent's command block.
* When a plan's own wrapper or script computes an argument (here `ROOT`), list it in the surface matrix:
  isolating the child process does not isolate how the script chose its inputs.
* Treat any "selection must be non-empty" guard as a non-vacuity check only; name what it does not
  prove (partial or wrong-tree listings) in the plan and capture the gap as a stash entry.
* Probe git environment vectors from a scrubbed environment first, and record both the scrubbed and the
  unscrubbed result when the host exports `GIT_*`.

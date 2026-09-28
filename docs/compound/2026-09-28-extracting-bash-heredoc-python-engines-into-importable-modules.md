---
title: "Moving a Python engine out of a bash heredoc into an importable module"
date: 2026-09-28
tags: ["ci", "bash", "python", "heredoc", "refactor", "gates", "pycache", "import-side-effects"]
shipment: 029-S
severity: medium
---

# Moving a Python engine out of a bash heredoc into an importable module

Shipment 029-S (feature 032-F, PR #77) moved the Python engine of `scripts/check-retired-architecture.sh` into
`scripts/lib/retired_arch.py`. The engine was about 1050 lines inside a `python - <<'PY'` heredoc. The same shipment
moved the Go masker into `scripts/lib/gomask.py` and shared it with `scripts/check-write-path-precondition.sh`.

The engine logic moved verbatim. Local review still found three wrapper-level regressions that the heredoc form had
never had to deal with. The same risks apply to any other heredoc-embedded gate engine.

## 1. Resolve the engine from the script's location, never from `$PWD`

A heredoc travels with its script, but a module file does not. Locate it relative to the wrapper:

```bash
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENGINE="$SCRIPT_DIR/lib/retired_arch.py"
```

The engine then derives the repository root from its own `__file__` (`resolve_repo_root`), not from the current
directory. This keeps the gate correct when CI or a developer invokes it from a subdirectory.

## 2. Importing a module writes `__pycache__` into the working tree

A heredoc program is never cached. An imported module is, so every gate run left
`scripts/lib/__pycache__/*.pyc` in the working tree. That noise can trip gitignore and unignore regression gates.

Fix it in two places:

* Export `PYTHONDONTWRITEBYTECODE=1` in every wrapper that imports from `scripts/lib/`.
* Add `__pycache__/` and `*.pyc` to `.gitignore` as a backstop, in append-only form.

## 3. Module import must have no side effects

Heredoc code runs top to bottom, so reading `sys.argv` or the current directory at top level was harmless there.
In a module, the same reads run at import time, which breaks unit tests and any other importer.

Move all argv, cwd and environment reads into `main()`.

Pin this with a test that imports the module with an empty `sys.argv` from a foreign cwd
(`test_import_with_empty_argv_and_foreign_cwd`). Also add an AST guard. It has to cover decorators and default
argument expressions, because both evaluate at import time and the first version of the guard missed them.

## Prevention checklist

* Produce behaviour-preservation evidence: capture gate output before and after the extraction and diff it
  byte-for-byte. See `docs/closure/029-S-032-F-behaviour-preservation-evidence.md`, which was re-captured at
  the review-fix HEAD.
* Run the new stdlib unit tests in CI on a pinned interpreter (`actions/setup-python` 3.12). Pinning the gate
  steps' own interpreter is deferred to `40C421EF`.
* Keep one canonical copy of shared helpers such as the masker. Delete the local clones, and add a test that
  asserts the wrappers import the shared copy.

## References

* `docs/closure/029-S-032-F-post-merge-closure.md`
* `docs/closure/029-S-032-F-behaviour-preservation-evidence.md`
* Related: `docs/compound/2026-09-28-bash-python-heredoc-stdin-and-github-token-admin-scope.md` (heredoc stdin
  capture)

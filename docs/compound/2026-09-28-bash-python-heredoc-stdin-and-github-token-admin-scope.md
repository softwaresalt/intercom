---
title: "A python heredoc takes over stdin, and GITHUB_TOKEN has no administration scope"
date: 2026-09-28
tags: ["ci", "github-actions", "bash", "python", "heredoc", "stdin", "github-token", "permissions"]
shipment: 028-S
severity: high
---

# A python heredoc takes over stdin, and GITHUB_TOKEN has no administration scope

Shipment 028-S (feature 031-F, PR #75) added `scripts/check-merge-strategy.sh`. Local adversarial review took three
cycles to clear, and two of the defects it found are general enough to reuse.

## 1. `python - <<'PY'` consumes stdin, so data piped to it never arrives

### What happened

The checker fed the live `gh api` response to an inline Python evaluator in two ways at once: a here-string
(`<<< "$json"`) or pipe for the data, and `python - <<'PY' ... PY` for the program. The program heredoc wins,
because both redirections target file descriptor 0. Python therefore read its program from stdin and saw no data.
The evaluator read an empty document and classified it as SKIP.

This made every live repository scan report SKIP. The fixture self-test still passed because it reads fixture files
by path, so the failure only showed up on the live path. Cycle 1 of the review missed it; cycle 2 caught it.

### Fix

When a program heredoc occupies stdin, pass data through a file:

```bash
tmp="$(mktemp)"
printf '%s' "$response" >"$tmp"
"$PYTHON_BIN" - "$tmp" <<'PY'
import json, sys
data = json.load(open(sys.argv[1]))
PY
```

Alternatively, pass the program with `python -c` and keep stdin free for the data.

### Prevention

* Give the self-test a case that uses the same transport as the live path, not only the fixture-by-path path. The fix
  commit `32519f5` added transport coverage plus exit-code assertions.
* Treat "the live path always returns the degraded verdict" as a bug signal, not as expected advisory behavior.

## 2. `permissions: administration: read` is not a valid GITHUB_TOKEN scope

### What happened

The first draft of the CI job declared `permissions: administration: read` so that it could read
`allow_squash_merge` and `allow_rebase_merge`. The workflow `permissions:` key has no `administration` entry:
actionlint rejects it, and GitHub would refuse to run the whole workflow, not only this job.

### Facts

* `GET /repos/{owner}/{repo}` returns `allow_squash_merge`, `allow_rebase_merge` and `allow_merge_commit` only when
  the caller has admin rights on the repository. With the default `GITHUB_TOKEN`, those fields are absent, and they
  are not `false`.
* Any explicit `permissions:` block sets every unlisted scope to `none`.
* The only way to read these fields in CI is a PAT or a GitHub App token with repository administration read access.
  That work is tracked as deferred stash entry `124AE9DE`.
* A checker must classify "field absent" and "field not a boolean" as SKIP, never as PASS. See
  `scripts/testdata/mergestrategy/skip-*.json`.

### Prevention

* Run `actionlint` on every workflow edit before review. It flags invalid permission scopes immediately.
* Check the current list of `permissions:` keys in the GitHub docs; do not infer scope names from REST API
  categories.

## References

* `docs/merge-strategy-gate.md`: GITHUB_TOKEN feasibility finding (031.002-T)
* `docs/closure/028-S-031-F-post-merge-closure.md`
* Related: `docs/compound/2026-09-06-ci-self-matching-grep-and-actionlint-verification-gap.md`

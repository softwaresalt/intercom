# Merge-strategy structural verification gate

Tracks the standing verification for Constitution Principle XI / P-009
(merge-commit-only PR merges), implemented by
`scripts/check-merge-strategy.sh` (031.001-T, shipment 028-S).

## GITHUB_TOKEN feasibility finding (031.002-T)

**Question**: can the workflow's default `GITHUB_TOKEN` read
`allow_squash_merge` / `allow_rebase_merge` from
`GET /repos/{owner}/{repo}`?

**Method**: compared the API response for the same public repository
(`softwaresalt/intercom`) under two credential postures — an
authenticated `gh` CLI session holding a classic PAT with `repo` scope
(admin/push/maintain/pull/triage on this repository), and a fully
anonymous, unauthenticated `curl` request.

**Command 1 — authenticated (repo-scope PAT, admin/push access)**:

```console
$ gh api repos/softwaresalt/intercom --jq "{allow_squash_merge, allow_rebase_merge, allow_merge_commit, permissions}"
{"allow_merge_commit":true,"allow_rebase_merge":false,"allow_squash_merge":false,"permissions":{"admin":true,"maintain":true,"pull":true,"push":true,"triage":true}}
```

Exit code: `0`. All three merge-strategy fields, plus the `permissions`
block, are present.

**Command 2 — anonymous, no token**:

```console
$ curl -s -o response_anon.json -w "HTTP_STATUS:%{http_code}\n" https://api.github.com/repos/softwaresalt/intercom
HTTP_STATUS:200
$ python -c "import json; d=json.load(open('response_anon.json')); print({k: d.get(k, '<ABSENT>') for k in ['allow_squash_merge','allow_rebase_merge','allow_merge_commit','permissions']})"
{'allow_squash_merge': '<ABSENT>', 'allow_rebase_merge': '<ABSENT>', 'allow_merge_commit': '<ABSENT>', 'permissions': '<ABSENT>'}
```

HTTP status: `200` (the read itself succeeds — this is a public
repository — but the response body omits all four fields entirely
rather than returning `null` or `false`).

**Finding**: `allow_squash_merge`, `allow_rebase_merge`,
`allow_merge_commit`, and `permissions` are returned by GitHub's REST API
only to a credential holding at least push-level access to the
repository; an anonymous or pull-only read gets a normal `200` response
with those fields silently absent. `scripts/check-merge-strategy.sh`
treats this absence as `SKIP`, never a `PASS` (AC-1.3).

This repository's CI workflow (`.github/workflows/ci.yml`) declares an
explicit top-level `permissions: contents: read` block (and each job that
sets its own `permissions:` restates the same narrow scope). Declaring
any explicit `permissions:` block switches the default `GITHUB_TOKEN`
from GitHub's broad repository-level default down to an allow-list where
every unlisted scope is `none` — there is no `administration` scope
granted anywhere in this workflow today. A `contents: read`-scoped token
carries no elevated repository-administration permission, so it is
expected to behave like the anonymous case above: the merge-strategy
fields are absent from the API response, and the checker correctly
reports `SKIP` rather than a false `PASS` or a job failure.

**Credential required for promotion to `MERGE_STRATEGY_GATE_REQUIRED`**:
GitHub Actions supports an `administration: read` permission scope for
the workflow-scoped `GITHUB_TOKEN` (see "Permissions for the
`GITHUB_TOKEN`" in GitHub's Actions security documentation). Granting
`permissions: administration: read` to the merge-strategy job — no new
secret or PAT needed — is sufficient to let `GITHUB_TOKEN` read these
fields going forward. See "Advisory-to-required promotion" below for the
exact steps and condition.

## Operator trigger and advisory→required promotion (031.005-T)

**Scope note**: this section documents the operator trigger only. Flipping
the repository settings themselves, and flipping the
`MERGE_STRATEGY_GATE_REQUIRED` repository variable, are both GitHub
repository administration actions outside both Stage's and Ship's role
boundary (deliberation D-5). No repository setting is changed by this
unit or by this document (AC-1.7).

### The two settings to disable

| Setting | Required value | Where |
|---|---|---|
| `allow_squash_merge` | `false` | GitHub → repo Settings → General → Pull Requests |
| `allow_rebase_merge` | `false` | GitHub → repo Settings → General → Pull Requests |

In the GitHub UI these correspond to unchecking **"Allow squash merging"**
and **"Allow rebase merging"** (leaving **"Allow merge commits"** checked),
per Constitution Principle XI.

### How to verify

Advisory (default), local, no CI required:

```console
$ bash scripts/check-merge-strategy.sh
```

Prints `PASS`/`FAIL`/`SKIP` plus a reason and exits `0` for `PASS`/`SKIP`,
`1` for `FAIL`. In CI, the same command runs in the `merge-strategy` job
of `.github/workflows/ci.yml` on every push/PR, and is advisory
(non-blocking) by default (AC-1.2).

To confirm the current live repository state directly:

```console
$ gh api repos/softwaresalt/intercom --jq "{allow_squash_merge, allow_rebase_merge}"
```

This requires a credential with at least push-level repository access
(see the GITHUB_TOKEN feasibility finding above) — an unauthorized or
insufficiently-scoped read reports the fields absent, which the checker
treats as `SKIP`, never a `PASS` (AC-1.3).

### Required credential (from the 031.002-T finding above)

Reading `allow_squash_merge`/`allow_rebase_merge` in CI requires the
`merge-strategy` job's `GITHUB_TOKEN` to carry the `administration: read`
permission scope. This repository's `ci.yml` already grants that scope to
the job (031.003-T), so no additional secret or PAT is needed to move from
`SKIP` to a real `PASS`/`FAIL` verdict.

### Exact promotion condition (advisory → required)

Promote `MERGE_STRATEGY_GATE_REQUIRED` from advisory to blocking only when
**all** of the following hold:

1. The `merge-strategy` CI job has run on `main` at least once (any recent
   push or PR) and returned a real `PASS` verdict — not `SKIP` — confirming
   the `administration: read` grant actually lets `GITHUB_TOKEN` read the
   fields in this repository's live CI environment.
2. The live repository settings already satisfy the gate
   (`allow_squash_merge == false` and `allow_rebase_merge == false`), so
   flipping the toggle to blocking does not immediately fail every open PR.
3. An operator with repository administration access has explicitly
   confirmed the flip via the repository's `MERGE_STRATEGY_GATE_REQUIRED`
   variable (GitHub → repo Settings → Secrets and variables → Actions →
   Variables), analogous to the documented
   `PIPELINE_TOPOLOGY_GATE_REQUIRED` / `WINDOWS_GATE_REQUIRED` rollout
   precedent in this repository.

Until all three hold, `MERGE_STRATEGY_GATE_REQUIRED` stays unset
(advisory) and the job never blocks a merge (AC-1.2).

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

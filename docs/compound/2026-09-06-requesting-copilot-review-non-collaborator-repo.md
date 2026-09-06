---
title: "Requesting GitHub Copilot PR review on a repo where it isn't a listed collaborator"
date: 2026-09-06
category: build-errors
tags: [github, copilot, pr-review, gh-cli, graphql]
---

# Requesting GitHub Copilot PR review on a repo where it isn't a listed collaborator

## Symptom

`gh pr edit <pr> --add-reviewer copilot-pull-request-reviewer` fails:

```text
'' not found
```

And `gh api repos/<owner>/<repo>/pulls/<pr>/requested_reviewers -X POST -f
"reviewers[]=copilot-pull-request-reviewer"` (no `[bot]` suffix) fails:

```text
Reviews may only be requested from collaborators. One or more of the users
or teams you specified is not a collaborator of the <owner>/<repo>
repository. (HTTP 422)
```

Both give the impression that Copilot code review is entirely unavailable
on the repository, and `autoharness gate copilot-review <pr> --enforcement
auto` correctly reports `NOT_APPLICABLE: PASS` at this point (no engagement
signal, enforcement not `required`) — a safe default, but easy to mistake
for "Copilot review cannot be engaged here" when it actually can be.

## Root cause

The REST `requested_reviewers` endpoint's collaborator check is satisfied by
the **exact bot login string with the `[bot]` suffix**:
`copilot-pull-request-reviewer[bot]`, not the unsuffixed
`copilot-pull-request-reviewer` (which is how the login appears in GraphQL
`Bot.login` fields and in review/thread `author.login` responses — see the
"Advisory Bot Identity" table in
`.github/instructions/github-pr-automation.instructions.md` §1.9.3). The gh
CLI's `--add-reviewer` reviewer-name resolution does not reliably support
Copilot's special reviewer identity either, independent of suffix.

## Resolution

Request Copilot review with the REST endpoint directly, using the
`[bot]`-suffixed login:

```powershell
gh api repos/<owner>/<repo>/pulls/<pr>/requested_reviewers -X POST `
  -f "reviewers[]=copilot-pull-request-reviewer[bot]"
```

This succeeds (returns the PR object) even on a repository where Copilot is
not listed as a normal collaborator and where `gh pr edit --add-reviewer`
and the unsuffixed-login REST call both fail. Verify the request landed via
GraphQL (which reports the unsuffixed login, per §1.9.3):

```graphql
query($owner:String!, $name:String!, $number:Int!) {
  repository(owner:$owner, name:$name) {
    pullRequest(number:$number) {
      reviewRequests(first:10) {
        nodes { requestedReviewer { __typename ... on Bot { login } } }
      }
    }
  }
}
```

Copilot does **not** automatically re-review on every subsequent push in
this configuration — re-run the same REST request after each HEAD advance
before polling for the next review, exactly as the instructions already
describe for the general case, but this confirms it also applies to a
repository not shown as having Copilot enabled by default.

## Verification

`autoharness gate copilot-review <pr> --enforcement auto --max-wait 0`
transitions from `NOT_APPLICABLE: PASS` (before the request) to
`WAITING_FOR_REVIEW: BLOCK` (immediately after, review pending) to
`SATISFIED: PASS` (once Copilot submits a review for the current HEAD and
all Copilot-authored threads are resolved) — this transition sequence is
itself confirmation the request worked, independent of the GraphQL check
above.

## Compounding value

Do not treat a failed `gh pr edit --add-reviewer` or a 422 from the
unsuffixed-login REST call as proof Copilot review is unavailable on a
repository. Retry with the REST endpoint and the `[bot]`-suffixed login
before concluding shadow review must fall back to "unavailable/advisory"
per §1.1's documented fallback.

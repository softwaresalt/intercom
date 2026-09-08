---
title: "Replying to a PR review comment via REST requires the pull_number segment in the path"
date: 2026-09-08
category: build-errors
tags: [github, copilot, pr-review, gh-cli, rest-api]
---

# Replying to a PR review comment via REST requires the `pull_number` segment in the path

## Symptom

Posting a reply to a Copilot (or any) review comment with:

```powershell
gh api repos/<owner>/<repo>/pulls/comments/<comment_id>/replies -X POST `
  --field body=@reply-body.md
```

fails with:

```json
{"message":"Not Found","documentation_url":"https://docs.github.com/rest","status":"404"}
```

even though `gh api repos/<owner>/<repo>/pulls/comments/<comment_id>` (GET,
no `/replies`) succeeds and confirms the comment exists.

## Root cause

The GitHub REST "reply to review comment" endpoint is scoped under the pull
request, not the bare repo-level comments collection:
`POST /repos/{owner}/{repo}/pulls/{pull_number}/comments/{comment_id}/replies`.
The bare `pulls/comments/{comment_id}/replies` path (no `{pull_number}`
segment) is not a valid endpoint at all — it 404s regardless of whether the
comment ID is correct, which is easy to misdiagnose as a wrong/stale
`comment_id` when the comment lookup by the same bare `pulls/comments/{id}`
path (GET, no trailing `/replies`) succeeds fine.

## Resolution

Always include the PR number in the path when posting a reply:

```powershell
$bodyText = 'Fixed in <sha>. <description>'
[System.IO.File]::WriteAllText((Join-Path $PWD 'reply-body.md'), $bodyText, (New-Object System.Text.UTF8Encoding $false))
gh api repos/<owner>/<repo>/pulls/<pr_number>/comments/<comment_id>/replies -X POST `
  --field body=@reply-body.md
Remove-Item reply-body.md
```

This returns the newly created reply comment object (with `in_reply_to_id`
set to the original comment's ID) on success. Continue to use the
file-backed `--field body=@file` pattern from
`.github/instructions/github-pr-automation.instructions.md` §1.5's "Shell-Safe
Comment Body Construction" section — this endpoint gotcha is independent of
that shell-escaping guidance and both apply together.

## Verification

`gh api repos/<owner>/<repo>/pulls/comments/<new_reply_id>` returns the reply
with `in_reply_to_id` matching the original comment, and the GraphQL
`reviewThreads` query shows the reply as an additional comment on the same
thread node.

## Compounding value

A 404 from the bare `pulls/comments/{id}/replies` path is a path-shape bug in
the calling command, not evidence the comment ID is wrong or the comment was
deleted/inaccessible — confirm the comment still resolves via
`pulls/comments/{id}` (no `/replies`) first, then retry the reply with the
`pulls/{pr_number}/comments/{id}/replies` shape before assuming a deeper
problem.

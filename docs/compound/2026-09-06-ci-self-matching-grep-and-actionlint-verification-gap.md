---
title: "A bare-substring CI assertion for 'this string must not appear' self-matches its own describing step"
date: 2026-09-06
tags: ["ci", "github-actions", "self-reference", "bash", "grep", "workflow-authoring"]
shipment: 010-S
severity: critical
---

# A bare-substring CI assertion self-matches its own describing step

## What happened

Shipment 010-S (011.002-T) added a CI step asserting that
`INTERCOM_LIVE_SDK_TESTS` must never be committed into any workflow file:

```bash
if grep -R --include='*.yml' --include='*.yaml' -n 'INTERCOM_LIVE_SDK_TESTS' .github/workflows/; then
  echo "::error::INTERCOM_LIVE_SDK_TESTS must not appear in any committed workflow file"
  exit 1
fi
```

This step's own **name** ("Assert INTERCOM_LIVE_SDK_TESTS is not enabled
in CI") and its own **script body** (which necessarily mentions the
variable name to describe what it's checking) both live inside
`.github/workflows/ci.yml` — the exact directory the grep scans. The
assertion therefore matched itself on every single run, making the check
**permanently self-failing**. This was never caught by:

* Local review (9 personas, including a Constitution reviewer and a
  Schema-CLI-Docs Coupling reviewer) — none of them executed the actual
  grep locally against the real file.
* `python -c "import yaml; yaml.safe_load(...)"` — this validates YAML
  *syntax*, not the *semantic* correctness of a shell one-liner embedded
  in a `run:` block.

It WAS caught by a multi-model adversarial review, specifically because
one reviewer (the lowest-confidence, 1-of-4 "unique finding" bucket)
actually reasoned through what the grep pattern would match against the
file containing it, rather than trusting the step's own stated intent.
The finding was independently verified by re-running the exact grep
command against the real file before the fix, confirming the match, then
again after the fix, confirming no match — this before/after empirical
check is what elevated it from "one reviewer's opinion" to "confirmed
blocking."

## The generalizable lesson

**Any CI assertion of the shape "grep for the presence of literal text
`X` and fail if found" is at risk of self-matching if the assertion step
itself must mention `X` by name to be comprehensible** (in its step name,
inline comments, or error message). This is a distinct failure mode from
a normal false positive — it's not that the DETECTOR is too broad, it's
that the **detector's own describing metadata is indistinguishable from a
genuine violation** to a bare substring match.

## Correct pattern

Anchor the pattern to the SHAPE of an actual violation (an assignment),
never a bare mention:

```bash
# WRONG: matches the step's own name/comments/error message.
grep -R 'INTERCOM_LIVE_SDK_TESTS' .github/workflows/

# RIGHT: matches only a genuine YAML env-key or shell assignment.
grep -RE '^[[:space:]]*(export[[:space:]]+)?INTERCOM_LIVE_SDK_TESTS[[:space:]]*[:=]' .github/workflows/
```

## Verification discipline this incident reinforces

1. **Any embedded shell command inside a CI `run:` block that greps the
   workflow directory for a literal string must be manually executed
   against the real, already-committed file** before trusting it —
   YAML-syntax validation and even a full multi-persona local review are
   insufficient; only actually running the exact command (or an
   equivalent, e.g. `actionlint`-style structural check) catches this
   class of self-reference bug.
2. More broadly for THIS shipment: a second, unrelated critical bug in
   the same PR (a deleted YAML step-boundary line that produced a
   workflow GitHub Actions could not parse at all — "This run likely
   failed because of a workflow file issue", zero jobs created) was
   caught only by the FIRST REAL CI RUN, not by any local check,
   including `python -c "import yaml; yaml.safe_load(...)"` (valid
   generic YAML) and a full local review round. **`actionlint` (a
   GitHub-Actions-schema-aware linter, not a generic YAML parser) is the
   tool that actually catches CI-specific structural errors** (e.g. a
   step whose `uses:`/`with:` keys got merged into a preceding step's
   mapping) that a plain YAML parser cannot, because both are valid YAML
   syntax — they are just semantically wrong for GitHub Actions'
   step-shape schema.

## Recommendation for future `ci.yml` edits in this repository

Before pushing any `.github/workflows/*.yml` edit, run BOTH:

```bash
python -c "import yaml; yaml.safe_load(open('.github/workflows/ci.yml', encoding='utf-8'))"
actionlint .github/workflows/ci.yml
```

Neither alone is sufficient; both are cheap (sub-second) and catch two
different, non-overlapping classes of error (generic YAML syntax vs.
GitHub-Actions-specific schema/step-shape). For any embedded shell
command that asserts "text X must not appear in this directory," manually
run the exact grep command against the real file first.

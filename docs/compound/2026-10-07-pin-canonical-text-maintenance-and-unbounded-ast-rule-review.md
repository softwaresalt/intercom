---
title: "Frozen-pin canonical texts: refreeze atomically, stage the freeze for a red, and stop unbounded AST-rule review"
date: 2026-10-07
category: process
tags: [gatecheck, retiredarch, pin, canonicalDecls, plan-review, threat-model, tdd, backlogit, 040-s]
---

# Frozen-pin canonical texts: refreeze atomically, stage the freeze for a red, and stop unbounded AST-rule review

## Symptom

The 14-unit retiredarch gate-integrity plan, `docs/plans/2026-10-07-intercom-go-retiredarch-gate-integrity-plan.md` (now superseded), failed multi-persona plan-review three times in a row.

- Each revision added new AST closure rules: CLI registration, declaration index, an in-checkout git location guard, and env-mutator denylists.
- Each review round then found a new bypass of the rules just added, or a red test that was already green at the parent.

The batch was split by operator disposition (D-RA-7, Option B). The narrowed correctness plan, `docs/plans/2026-10-07-intercom-go-retiredarch-correctness-plan.md`, reached ADVISORY with no P1s on its second attempt.

## Root causes

1. **Unbounded rule surface.**
   - A denylist or closure rule over Go AST shapes has no natural stopping point. Every reviewer can name another construct that evades it.
   - Without a written, bounded threat model (who the adversary is, and what they can already do in the same job), plan-review can't decide "enough", so it keeps failing.
   - The `gitPathInside` guard alone drew three P1s. Yet an attacker able to plant `git` on PATH already runs code in the job, which makes them the same trust class as the accepted residual R-A2b.
2. **Canonical texts are a second copy of the code.**
   - `pin.go` `canonicalDecls` must match `select.go` token for token.
   - Adding a top-level declaration also requires a `closedWorldDecls` entry, plus a frozen-set entry where applicable.
   - The `"import"` decl sits in BOTH `pathspecFrozenDecls` and `prefixFrozenDecls`, so a single import change refreezes both.
   - `writeMutatedCopy` anchors (about 14) must each stay unique in `select.go`.
3. **Fake reds.**
   - A reject row added in the same step as the freeze is "red" only because the pin is broken mid-change.
   - A fake `git` that prints nothing is "red" only because of an existing "returned nothing" error.
   - Both look like test-first, but neither proves the new rule.

## Fix / practice

- **Bound the threat model before adding pin rules.** State the adversary and the same-job trust class. Anything that already executes in the job is a residual, not a rule target.
  - Prefer an organizational control: required CODEOWNERS review on `tools/gatecheck/**` (needs operator branch protection) over ever-stricter AST rules.
  - The deferred entries DC921AF6, 3750C37C and 0ECC1895 return to deliberation on that basis.
- **Refreeze atomically (ALP pattern).** The code change plus `canonicalDecls`/`closedWorldDecls`/frozen-set updates go in ONE commit.
  - Verify with `git show --stat` that it lists exactly the select.go and pin.go pairs.
  - Roll back in dependency order: never revert the refreeze alone while a later unit still uses the new declaration.
- **Stage the freeze to get a real red:**
  1. Refreeze the existing texts and add the closed-world entry, so the live tree is `OK()`.
  2. Write the reject row and observe it pass the pin (red).
  3. Freeze the new declaration and observe the row go green.

  Assert the full field pattern (`SelectFound && GuardFound && PrefixOK && !PathspecOK`), so a parse-error all-false result can't masquerade as a rejection.
- **Fake executables must be natively launchable and must succeed at the parent.**
  - Copy `os.Executable()` (the test binary) to `dir/git(.exe)` with mode `0o755`. Use a `TestMain` branch gated on an env var AND git-style argv, and have it exit non-zero if the env var is set with test-style argv.
  - The fake prints a valid answer and writes a marker file. The red is that the parent accepts it; the fix refuses it and the marker is absent.

## Tooling notes (backlogit 1.8 CLI on Windows)

- `backlogit add --section` keeps only `description`. Set `acceptance-criteria` and `implementation-notes` with a follow-up `update --section`.
- When a PowerShell helper function calls `Write-Output` before `return $id`, its return value becomes an array. Passing that to `--dependencies` produced a "mutation partial (not-applied)" error. backlogit rolled the item back cleanly. Use `Write-Host` for progress output inside such helpers.
- Non-ASCII text passed to native exes from a BOM-less `.ps1` via `pwsh -File` was mangled to U+FFFD. Keep CLI text ASCII, or pass args through Python `subprocess` (CreateProcessW).

---
title: "Textual CI-YAML guards written as bypass denylists invite an unbounded review loop"
date: 2026-10-01
category: test-design
tags: [ci, yaml, guards, copilot-review, gatecheck, 037-s]
---

# Textual CI-YAML guards written as bypass denylists invite an unbounded review loop

## Symptom

In 037-S (M4 of the gate-engine Go migration), new Go tests guarded `.github/workflows/ci.yml`
against three things coming back: a Python interpreter, a Python CI step, or a weakened gate
step. The guards scanned the YAML as text with line patterns, because the module has no YAML
parser dependency.

PR #87 went through 23 Copilot reviews. Rounds 1–22 each found another YAML form that runs or
weakens a step without matching the patterns. Each was a valid finding:

* flow-style steps and flow-sequence aliases
* YAML aliases and anchors
* explicit `? key` mappings
* whitespace before a key's colon (`run : python -V`)
* escapes in double-quoted keys (`"r\u0075n":`, `"i\u0066": false`)
* job-level or reformatted `continue-on-error`
* Windows and versioned interpreter forms (`python.exe`, `py.exe {0}`, `pythonw`, `python3.14`)
* shells reached through `shell:` or through environment values
* in the Go init-time guards, function values and closures

Each fix closed one form and the next round found another.

## Root cause

The guards were denylists: "reject these known-bad spellings". YAML has many equivalent
spellings for the same mapping, and an interpreter can be named in many ways. A denylist over
an open grammar never converges. Each review round can always produce one more equivalent form.

## Fix / prevention

1. **Write a shape allowlist first.** Before any content check, assert that every node outside a
   small, explicitly exempted region (here `topology-check`) uses one canonical block-style form.
   Fail closed on anything else: flow collections, aliases, anchors, tags, explicit keys, quoted
   keys, and multi-document streams. The content patterns then need to match only that one
   form. 037-S reached this design incrementally (the shared "YAML shape check"). Starting there
   would have removed most rounds.
2. **Pin required wiring positively, not by absence.** Check that a protected step carries its
   exact `uses:` SHA line, has no `if:` and has no `continue-on-error`, instead of searching for
   bad edits.
3. **Prefer structural parsing when a dependency is acceptable.** A real YAML decoder removes the
   spelling problem, though it is a new dependency (Principle VI). If one is added, keep the shape
   allowlist: a decoder accepts aliases and merge keys that a guard should still reject.
4. **Budget review rounds explicitly.** When a guard's review loop passes the 3-cycle limit, ask
   whether it is converging. If each round finds a new equivalent spelling, change the design to
   an allowlist rather than adding the next pattern.

## Evidence

* `docs/plans/evidence/2026-09-28-gate-engine-go-migration/m4.md` §10 (rounds 1–22).
* `docs/closure/037-S-047-F-post-merge-closure.md` (risky action record).
* `tools/gatecheck/ciwiring_retire_test.go` and the related CI-wiring tests.

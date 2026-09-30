---
title: "Go/MSYS bash-bridging pitfalls and a Go Unicode case-mapping gap vs CPython"
date: 2026-09-30
tags: ["go", "bash", "msys", "windows", "unicode", "case-mapping", "gates", "gatecheck"]
shipment: 034-S
severity: medium
---

# Go/MSYS bash-bridging pitfalls and a Go Unicode case-mapping gap vs CPython

Shipment 034-S (feature 044-F, PR #81) ported the gate-engine's Python semantics layer, masker, and
write-path engine to Go, plus a shared bash runner (`scripts/lib/gatecheck-run.sh`) that a Go test
suite drives directly. Porting surfaced two categories of hard-won, reusable lessons.

## 1. MSYS/Git-Bash Windows-path translation is not simple string substitution

A helper that converts a Windows drive-letter path to the MSYS POSIX form for a PATH-shim
directory (`toBashPath()`) had an off-by-slice bug:

```go
// wrong: silently produces "/c:/Users/..." instead of "/c/Users/..."
return p[len(m[0])-2:]

// right
return "/" + p[len(m[0]):]
```

The bug did not fail loudly. It defeated a fake-`go` PATH shim used by a test, so the test silently
invoked the real system `go` instead — and one assertion coincidentally still passed, because a real
`go` failing on a missing `go.mod` produces the same exit-2 signature the test asserted for a
deliberately-failing fake `go`. **A PATH-shim test that "passes" is not proof the shim resolved** —
assert on a shim-specific side effect (a marker file, a distinguishing stderr string), not merely on
an exit code that a fallback path could also produce.

Separately, deriving a path via `cd ... && pwd` inside Git-Bash is not round-trippable through a
naive drive-letter regex, for two reasons that both showed up on the same dev box:

* Under the Windows TEMP directory, Git-Bash's `pwd` uses its special `/tmp` mount alias rather than
  a plain drive-letter conversion.
* The current user's profile directory had a short-name (8.3, e.g. `DEWILL~1`) vs long-name
  (`dewilliams`) alias mismatch, which defeats string-equality path comparison entirely.

Fix: round-trip through bash's own `pwd -W` (a Windows-only branch, gated on `runtime.GOOS ==
"windows"`) and compare paths by filesystem identity (`os.SameFile`), not string equality. POSIX
platforms compare `os.SameFile` directly, with no `pwd -W` branch needed.

## 2. Bare `bash` on PATH is ambiguous on Windows dev boxes

On at least one dev machine, bare `bash` resolved to the Windows-builtin WSL2 shim
(`C:\Windows\system32\bash.exe`) ahead of Git for Windows' own bash
(`C:\Program Files\Git\bin\bash.exe`). The WSL shim has no Go toolchain and does not perform the
Windows-path argument translation a Git-Bash-targeted script depends on — every wrapper-invocation
test failed with exit 127 and a mangled, backslash-stripped path. CI (`ubuntu-latest`) has only one
`bash` and is unaffected. Local dev-loop fix: prepend `C:\Program Files\Git\bin` to `PATH` before
invoking `bash`/`go test` for anything that shells out to a bash script — this is an environment
workaround, not a code or CI change.

Related, and independently confirmed: `syscall.SysProcAttr.Setpgid` does not exist on Windows, so
process-group signal tests need `!windows`/`windows`-tagged files, with the Windows side honestly
reporting an unsupported/SKIP condition rather than faking a PASS. Where a Windows dev box cannot
exercise POSIX signal-trap logic (`SIGINT`/`SIGTERM` + cleanup) at all, hand-verify it independently
via WSL2 (`set -m` + `kill -INT/-TERM -- -$PGID`) before relying solely on the (untestable-here) CI
run.

## 3. `unicode.IsUpper`/`unicode.IsLower` under-match CPython's `str.isupper()`/`islower()`

Go's `unicode.IsUpper`/`IsLower` are Lu/Ll Unicode-category-only. CPython's `str.isupper()` /
`str.islower()` additionally honor the `Other_Uppercase` / `Other_Lowercase` **derived** Unicode
properties. This under-matching was caught by a golden case: U+2071 SUPERSCRIPT LATIN SMALL LETTER
I is `Other_Lowercase` but not Ll, so CPython's `islower()` returns true while a naive Go port
returned false.

Fix: build a supplementary derived-property table (`case_table.go`,
`otherLowerTable`/`otherUpperTable`) and OR it into the category check. When porting any Python
`str.is*()` semantics to Go, do not assume the stdlib `unicode` package's category-based predicates
are equivalent — check whether CPython additionally consults a *derived* property, and add a golden
covering at least one derived-property-only code point.

## Prevention checklist

* When bridging Go tests to a bash script, assert on a shim/mock-specific side effect, not merely a
  reproducible exit code.
* Prefer filesystem-identity (`os.SameFile`) over string equality when comparing paths that may have
  round-tripped through a shell's own path resolution (drive-letter form, mount aliases, 8.3 names).
* Pin the shell interpreter explicitly (absolute path) in any Windows dev-loop instructions or
  scripts that assume Git-Bash semantics; do not rely on bare `bash` resolving predictably.
* When porting a Python string/character predicate to Go, check the Unicode Character Database for
  a *derived* property (`Other_Lowercase`, `Other_Uppercase`, etc.) that CPython may consult beyond
  the raw General_Category, and add a golden for it.

## References

* `docs/plans/evidence/2026-09-28-gate-engine-go-migration/m1.md` (M1-T5b, M1-T9, M1-T10 sections)
* `docs/closure/034-S-044-F-post-merge-closure.md`
* `docs/plans/2026-09-28-intercom-go-gate-engine-go-migration-plan.md`

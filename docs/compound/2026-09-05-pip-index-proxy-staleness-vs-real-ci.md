---
title: "A locally configured pip index/proxy can silently diverge from what real CI actually resolves"
date: 2026-09-05
category: build-errors
tags: [pip, hash-pinning, ci, supply-chain, package-index, 006-S]
---

# A locally configured pip index/proxy can silently diverge from what real CI actually resolves

## Symptom

Built a hash-pinned pip constraints lock file for `.github/workflows/ci.yml`'s
`autoharness` install (007.001-T / U1, shipment 006-S). `pip index versions
autoharness` in this sandbox reported `LATEST: 1.4.11`, so the lock file was
built for `autoharness==1.4.11`, fully verified end-to-end locally
(`pip install --require-hashes --only-binary=:all: ...` succeeded against a
cross-platform target matching the real CI runner). Pushed. Real GitHub
Actions CI (public `ubuntu-latest`, unrestricted network) immediately failed
with:

```
Unknown gate subcommand: pipeline-topology
##[error][CI Topology Check] INVALID (exit 2)
```

The `topology-check` job's own entrypoint script calls `autoharness gate
pipeline-topology` — a subcommand that simply does not exist in `1.4.11`; it
was added in a later release.

## Root cause

`pip config list` revealed `global.index-url=https://packagefeedproxy.microsoft.io/pypi/simple/`
— this sandbox's pip is pinned to an **internal corporate package-feed proxy**,
not the real public PyPI. That proxy's mirror of `autoharness` was stale:
its "latest" was `1.4.11` while the real `https://pypi.org/pypi/autoharness/json`
(and this repo's own last-known-good `main`-branch CI job log) showed the
floating `pip install autoharness` had already been resolving `1.5.0` on the
actual GitHub-hosted runner (which has no such proxy and talks to public
PyPI directly) for some time. Pinning to the proxy's stale "latest" silently
**downgraded** the runtime CI would get, rather than reproducing it — the
exact opposite of what a hash-pin is supposed to do (U1's own AC-3 explicitly
requires the pin to reproduce, not change, CI's current resolution).

## Fix

1. **Never trust a locally configured index/proxy as ground truth for what a
   different, unrestricted-network CI runner will resolve.** Cross-check
   against the *real* public index directly:
   `curl -s https://pypi.org/pypi/<name>/json | python -c "import json,sys; print(json.load(sys.stdin)['info']['version'])"`
   — this JSON metadata endpoint also publishes the exact SHA-256 digest per
   uploaded file (`urls[].digests.sha256`), which is authoritative even when
   `files.pythonhosted.org` itself is unreachable (see below).
2. **Even better: read the actual CI runner's own most recent job log** for
   the pre-existing floating install, if one exists. It is ground truth for
   "what does this exact runner/network path currently resolve" in a way no
   local reasoning about "latest" can be.
3. **A locked-down sandbox may block the real package CDN even when the
   metadata API is reachable.** Here, `files.pythonhosted.org` (the actual
   wheel-download CDN) returned `SSLV3_ALERT_HANDSHAKE_FAILURE` (network
   policy blocking direct public internet for large binary downloads) while
   `pypi.org`'s JSON API (metadata only) worked fine. When a full local
   `pip download`-and-hash verification isn't possible, PyPI's own published
   `digests.sha256` in the JSON metadata is a legitimate, authoritative
   substitute — it is the same digest pip's own `--require-hashes` will
   check against, computed by PyPI itself at upload time.
4. **The rest of a resolved transitive closure often survives a
   single-package version correction unchanged.** Here, `autoharness`
   1.4.11 and 1.5.0 declared byte-identical `requires_dist`
   (`jsonschema>=4.23.0`, `pyyaml>=6.0.2`), so only the one `autoharness`
   line in the lock file needed correcting — the other 7 pinned
   hashes (already CI-verified in the failing run's own install step, which
   succeeded right up until the topology-check script itself ran) needed no
   change.
5. **Let real CI be the final verification when local verification is
   incomplete.** State that limitation explicitly in the commit/PR rather
   than overclaiming a local check that didn't actually cover the corrected
   version.

## Related gotchas from the same shipment

* **Cross-platform `pip download --python-version X --platform Y` resolution
  can silently omit a real, marker-conditional transitive dependency.**
  `referencing==0.37.0` declares `typing-extensions>=4.4.0; python_version <
  "3.13"` — true under the target Python 3.12 — but a `pip download` cross-
  platform/cross-version resolve omitted it entirely from its output. A
  **real** interpreter at the target Python version, run with `pip install
  --dry-run` and no platform override, correctly included it. When
  `--require-hashes` needs the *exact* resolved closure, prefer verifying
  against a real interpreter at the target Python version over a
  cross-platform simulated resolve, especially for packages with
  environment-marker-conditional dependencies.
* **`--require-hashes` alone does not close the unhashed-sdist-build-
  isolation fallback.** If a hash-pinned wheel is ever unavailable for the
  exact platform/Python combination, pip can fall back to building an
  unhashed sdist via PEP 517 build isolation — a live code-execution path
  the hash-pin was supposed to close. Add `--only-binary=:all:` alongside
  `--require-hashes` to foreclose that fallback entirely.
* **`grep -Ev` exits 1 on zero matches — a legitimate empty result, not an
  error.** Under `set -euo pipefail`, a naive caller can misread that as
  failure. Worse: **naively toggling `set +e`/`set -e` inside a helper
  function to work around this is itself unsafe**, because `errexit` is
  GLOBAL shell state, not scoped per function call — a nested `set -e` can
  clobber an outer caller's own `set +e`, causing a silent mid-script abort
  with no error message at all (this happened during this exact shipment's
  own remediation). The safe pattern is `if var="$(...)"; then ... else
  status=$?; fi`, which captures exit status via the `if` condition's own
  errexit exemption without ever touching global shell state.
* **`sed | grep` inside a single `$(...)` runs the whole pipe in a subshell**
  whose internal `PIPESTATUS` does not propagate to the outer shell, and
  `pipefail` alone can report the rightmost command's exit code even when an
  earlier stage is the one that actually failed. Run each pipeline stage as
  a separate command-substitution assignment when you need to distinguish
  which stage failed.

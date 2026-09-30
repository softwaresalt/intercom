#!/usr/bin/env bash
set -euo pipefail

# Retired-architecture regression gate for the corrected C1 config surface.
#
# Threat-model limits (D6b): this is an anti-accident control, not an
# anti-adversary control. CI runs workflow files from the PR head, so a PR
# can edit the workflow or this script. The CI workflow is also
# autoharness-generated and may overwrite the wiring step on a future render.
#
# Usage:
#   scripts/check-retired-architecture.sh
#     Scans tracked files in internal/** (excluding *_test.go and any
#     testdata/ directory), config.toml.example, and cmd/**. Exits 0 when no
#     retired-architecture token is found as a Go identifier or TOML key.
#     Disclosure (015.012-T): this sentence attaches the *_test.go/testdata
#     exclusion only to internal/** and lists cmd/** unqualified -- that is
#     not an oversight in the description. cmd/**'s selection predicate
#     deliberately does NOT exclude *_test.go or testdata/ paths (unlike
#     internal/**'s), a decision guarded by the
#     'selection cmd/ coverage (AG-5/D4)' self-test assertion (015.011-T).
#   scripts/check-retired-architecture.sh --self-test
#     Runs three checks: (1) verifies the committed TOML fixture suite in
#     scripts/testdata/retired-*.toml against scripts/testdata/retired-manifest.json
#     and the committed Go fixture suite in scripts/testdata/retiredgo/*.go against
#     scripts/testdata/retiredgo-manifest.json; (2) verifies the real selection
#     logic structurally (every tracked internal/** non-test, non-testdata .go
#     file is selected, internal/** test/testdata paths and scripts/ itself are
#     excluded, selection is non-empty, and config.toml.example dispatches to the
#     TOML engine); (3) verifies the real tracked tree passes a repo scan. Exits 0
#     only when all three checks succeed. Semantics UNCHANGED by 015.001-T.
#   scripts/check-retired-architecture.sh --self-test-integrity
#     Integrity-only mode (added 015.001-T, permanent as of 015.013-T's
#     window closure): runs ONLY checks (1) and (2) above (the fixture
#     suites and the selection-logic self-test). Does NOT run the repo
#     scan (3). Used by ci.yml's integrity step so that step stays
#     unconditionally blocking while the separately-toggled verdict step
#     (governed by the `RETIRED_ARCH_GATE_ADVISORY` repository variable,
#     fail-closed by default since 015.013-T) owns the real repo-scan
#     enforcement decision.
#
# Every invocation emits a GitHub Actions ::notice:: line naming the
# resolved SCRIPT INVOCATION MODE (repo / self-test / self-test-integrity)
# so which code path ran is falsifiable from run history. Correction
# (post-review, U-2): the notice text itself does NOT vary with the
# RETIRED_ARCH_GATE_ADVISORY toggle's resolved value (the verdict step
# always invokes this script with no flag, so it always prints
# mode=repo regardless of advisory/blocking posture) -- H8's actual
# falsifiability guarantee for the TOGGLE's own outcome comes from the
# step's pass/fail/continue-on-error result in the run's job summary and
# the "::notice::retired-arch gate mode=repo" line's mere PRESENCE (proof
# the verdict step actually executed the real repo scan), not from a
# toggle-value-specific notice payload.
#
# Disclosure (015.012-T) -- verified claims only, no pathspec edits:
#
# Residual blind spots (accepted risk, not fixed by this shipment):
#   - Aliased imports (e.g. `sm "github.com/x/socketmode"`) are not
#     specially tracked; the alias identifier itself is still scanned like
#     any other Go identifier, but an import whose PACKAGE PATH names a
#     retired component while every local alias/reference does not would
#     not be caught.
#   - go.mod / go.sum are outside the scan pathspec entirely.
#   - Non-.go / non-.toml files under internal/** are not scanned (no
#     engine is registered for other extensions there).
#   - V10: mask_go_non_code() has no explicit EOF state check for the
#     string/rune states (only raw_string gained one, in 015.007-T, because
#     that state's unmask decision specifically depends on knowing the
#     whole buffered content). An unterminated string or rune literal at
#     EOF in the `string`/`rune` states masks to the end of the file
#     silently rather than emitting a finding -- a silent fail-open,
#     disclosed and not fixed.
#   - Interpreted-string struct tags (e.g. `"json:\"channel_id\""` as a
#     Go string literal rather than a raw backtick literal) are missed by
#     015.007-T's lexer-sourced, raw_string-only capture (H2 residual).
#     Findings only ever print `match.group(0)` (the matched identifier),
#     so this residual gap cannot leak any additional payload to CI logs.
#
# Unfixtureable malformed-input boundary (relocated from 015.009-T by
# escalation-adjudicated review): a line that CLOSES a multiline string
# and then carries an assignment on the same line has its LHS suppressed
# by the fallback lexer's pre-line-state check (015.009-T). This is NOT
# fixtureable in the shared dual-engine suite: that construct is invalid
# TOML, so the tomllib engine fail-closes to a parse-error finding while
# the fallback engine would report clean -- no single manifest expectation
# can be green under both engine labels for the same fixture. The
# behaviour is recorded here, with the reason it is untested, rather than
# committing a fixture that would be permanently red under one engine.
#
# check-write-path-precondition.sh divergence (SUPERSEDED by 032.001-T,
# retained verbatim as the historical record): that script (outside this
# shipment's pathspec, per AG-3) carries its own independent
# mask_go_non_code() clone that has NOT been updated with this shipment's
# raw_string buffering / struct-tag-visibility change (015.007-T). This
# divergence is CURRENT AND DELIBERATE FOR THIS CYCLE ONLY -- it is not a
# statement of permanent intent, so the deferred follow-up (stash
# provenance 6C24E2E4, unification of the two clones) keeps its own
# decision about whether and when to converge them.
#
# CANONICAL GO MASKER DECISION (032.001-T) -- shipment 029-S, plan unit 2
# / U2-T1 of docs/plans/2026-09-18-intercom-go-gate-reliability-plan.md:
# the convergence direction left undecided above is now DECIDED. The
# canonical Go masker is the retired-architecture SUPERSET -- raw_string
# content buffering, struct_tag_re unmasking of a raw-string literal whose
# ENTIRE content is a well-formed struct tag, and the
# unterminated-raw-string EOF fail-closed branch. Rationale: AC-2.3
# (retired-arch verdicts identical pre/post extraction) forecloses the
# alternative -- adopting the write-path clone would drop struct_tag_re
# unmasking and flip the 015.007-T struct-tag fixture
# retired-reject-struct-tag.go from reject to clean (measured: the only
# retiredgo*/ fixture whose verdict changes under the clone; the
# go-differential suite scans unmasked and is unaffected), so only the
# superset can be canonical. The masker
# moves to scripts/lib/gomask.py (032.002-T) and both gates import it
# (032.003-T / 032.005-T).
#
# EXPECTED WRITE-PATH VERDICT DELTAS (enumerated in advance, AC-2.4):
#   Masker output differs from the write-path clone ONLY for a raw-string
#   literal whose whole content matches struct_tag_re: the superset leaves
#   that content visible instead of blanking it. The unterminated-raw EOF
#   branch masks the buffered content, which is byte-identical to the
#   clone's char-by-char masking, so it is NOT a delta.
#   D-1 (class): a write-path finding can newly appear iff such a struct
#       tag's content itself contains a qualified selector token (e.g.
#       `x:"os.Remove"`) with valid selector boundaries. Direction: fail
#       closed (a new finding, never a lost one) -- nothing the clone
#       exposed is masked by the superset.
#   D-1 measured instances at the pre-extraction base (beeb84f, verdict-
#       equivalent to bound snapshot 14d44e3c for every gate input):
#       write-path fixtures (scripts/testdata/writepath/*.go): 0;
#       tracked cmd/** and internal/** non-test .go files: 0. One
#       write-path-selected file (internal/config/config.go) has
#       struct-tag content newly exposed, and none of it contains a
#       selector token, so its verdict is unchanged.
#   Therefore the expected write-path verdict delta set is EMPTY; any
#   observed delta is unenumerated and a defect (AC-2.4), and any delta on
#   a tracked cmd/** or internal/** file must be remediated in the same
#   change (AC-2.2 precedence).
#
# Forward-looking guard rail (attributed as such, not yet load-bearing):
# any future scope expansion of this gate MUST keep the tracked-only
# `git ls-files` enumeration used by select_repo_paths() and MUST NOT
# switch to Path.rglob()/os.walk() or any other filesystem-walking
# enumeration. A filesystem walk would also pick up ignored-but-present
# files (e.g. a local, gitignored `.env.*`), and this gate's findings are
# printed to stdout/stderr, which GitHub Actions echoes to a public CI
# log -- so a filesystem-walk enumeration could turn an ignored secret
# file into a public-log disclosure vector.
#
# M2-T11 (docs/plans/2026-09-28-intercom-go-gate-engine-go-migration-plan.md,
# plan section C-3/M2-T11): the engine that used to run under an embedded
# interpreter now runs entirely in Go, at
# tools/gatecheck/internal/retiredarch. This wrapper is reduced to the
# shared build/invoke/cleanup runner (scripts/lib/gatecheck-run.sh, C-3): it
# builds the gatecheck binary once per invocation and forwards this
# script's own first argument straight through as `gatecheck retired-arch`'s
# mode flag. Exit codes and ordered stdout/stderr bytes are preserved
# (ED-3/ED-5 are the only permitted deltas; see m2.md for the parent-vs-head
# parity evidence). The `::notice::` line and the --self-test/
# --self-test-integrity success banners moved INTO the Go engine itself
# (internal/retiredarch.Run), so this wrapper no longer prints them.
#
# Root anchoring (C-3 exception): retired-arch anchors its root on ITS OWN
# ENGINE LOCATION, not the caller's working directory -- this MATCHES the
# pre-switch Python engine's own resolve_repo_root(), which anchored on
# __file__ rather than cwd. This is why ROOT below is derived from
# GATECHECK_SRC (the gatecheck tool source location, set by
# gatecheck-run.sh) via `git -C`, unlike write-path/unignore/
# merge-strategy-evaluate's wrappers, which use a plain cwd-based
# `git rev-parse --show-toplevel`.
#
# Caller-cwd precondition guard (Copilot round 13, PR #83): the
# pre-switch wrapper's OWN `ROOT="$(git rev-parse --show-toplevel)"` call
# ran in the CALLER's cwd, so under `set -e` the whole invocation failed
# before the Python engine ever ran when invoked from outside any Git work
# tree -- independent of, and prior to, Python's own module-anchored
# resolve_repo_root() call. Collapsing straight to the GATECHECK_SRC-
# anchored `-C` call above preserves the ENGINE's root-COMPUTATION
# mechanism (still module-anchored, matching Python exactly) but silently
# dropped this separate caller-cwd PRECONDITION: GATECHECK_SRC is always
# inside a git work tree (it is this repo's own checkout), so the `-C`
# call can never itself fail for this reason, and the wrapper began
# succeeding from arbitrary non-Git directories -- an observable CLI
# behavior change the PR's "exit codes and ordered stdout/stderr bytes are
# preserved" promise does not license. Restore the precondition as its own
# statement, decoupled from the actual (still engine-anchored) $ROOT
# value below: same command, same caller cwd, same fatal message and exit
# 128 on failure, output discarded since only the precondition matters.
git rev-parse --show-toplevel >/dev/null

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=lib/gatecheck-run.sh
source "${SCRIPT_DIR}/lib/gatecheck-run.sh"

ROOT="$(git -C "${GATECHECK_SRC}" rev-parse --show-toplevel)"

cleanup() {
  gatecheck_cleanup
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

# Caller-argument allowlist (post-review remediation, PR #83 Copilot
# round 4/5 findings): gatecheck_invoke appends the trusted "--root
# ${ROOT}" ahead of any args this wrapper forwards, but main.go's
# parseRoot scans the WHOLE arg list for "--root"/"--root=" and the LAST
# occurrence wins. The pre-M2-T11 wrapper never had this exposure because
# its own `case "${1:-}"` dispatch inspected ONLY the first argument and
# silently ignored $2 onward -- it never forwarded them anywhere, so a
# stray --root past the first argument was inert. Round 4 closed the
# security gap but overshot the CLI-surface-preservation promise (the PR
# summary explicitly commits to "no CLI surface changes beyond swapping
# the interpreter") by rejecting >1 argument outright instead of matching
# that same silent-ignore behavior.
#
# This restores BOTH properties at once: MODE captures only the first
# argument (validated against the same three-way allowlist the old
# wrapper used, still rejecting an unrecognized first argument with exit
# 2, unchanged from round 4), and only that single validated value -- not
# "$@" -- is ever forwarded to gatecheck_invoke below. Any second or later
# argument (a duplicate --root included) is therefore never seen by
# parseRoot at all, exactly reproducing the old wrapper's "$2 onward is
# ignored" contract while still closing the root-override gap. This does
# not touch the shared --root contract in main.go, which is used by every
# gate engine (write-path, unignore, merge-strategy-evaluate) outside
# this shipment's scope.
mode="${1:-}"
case "${mode}" in
"" | --self-test | --self-test-integrity) ;;
*)
  echo "usage: scripts/check-retired-architecture.sh [--self-test|--self-test-integrity]" >&2
  exit 2
  ;;
esac

gatecheck_build
build_rc=$?
if [ "${build_rc}" -ne 0 ]; then
  exit "${build_rc}"
fi

if [ -n "${mode}" ]; then
  gatecheck_invoke retired-arch "${mode}"
else
  gatecheck_invoke retired-arch
fi
exit $?
#!/usr/bin/env bash
set -euo pipefail

# check-depguard-fixtures.sh
#
# Durable proof for the depguard Copilot-SDK-boundary rule (011.014-T,
# resolves 309FBF5A(ii), .golangci.yml `copilot-sdk-boundary`). A one-shot
# manual demonstration cannot detect future config drift (e.g. depguard
# silently falling out of `linters.enable`, or the `files` pattern silently
# widening) -- this script re-proves all three directions on every run:
#
#   (i)   a file OUTSIDE the internal/copilotprobe allowlist that imports
#         the Copilot SDK root package must make golangci-lint FAIL WITH
#         THE SPECIFIC depguard copilot-sdk-boundary DIAGNOSTIC (not merely
#         a nonzero exit, which could come from an unrelated typecheck or
#         config error and would be indistinguishable from proof that
#         depguard itself is what rejected the fixture -- adversarial
#         review finding, remediated).
#   (ii)  a file under a SIBLING directory name (internal/copilotprobe2/)
#         must ALSO be rejected with the same specific diagnostic, proving
#         the allowlist pattern is anchored to the exact directory and has
#         not silently widened to match any "copilotprobe*"-prefixed name.
#   (iii) the real internal/copilotprobe package itself (the allowlisted
#         directory, which genuinely imports the SDK in client.go) must
#         PASS lint -- proving the allowlist is not so narrow it rejects
#         its own legitimate exemption (positive-direction assertion,
#         adversarial review finding, remediated).
#
# Fixtures are committed as .go.tmpl templates under
# scripts/testdata/depguard/ (the Go toolchain ignores non-.go extensions)
# and copied at runtime into real, git-ignored locations so golangci-lint's
# actual import-resolution and depguard path-matching run for real, then
# removed on exit regardless of outcome.
#
# Usage:
#   scripts/check-depguard-fixtures.sh
#     Runs all three checks. Exits 0 only if both negative fixtures are
#     rejected specifically by depguard's copilot-sdk-boundary rule AND
#     the real internal/copilotprobe package passes lint.

ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"

if ! command -v golangci-lint >/dev/null 2>&1; then
  echo "golangci-lint is required on PATH" >&2
  exit 2
fi

OUTSIDE_DIR="internal/depguardfixture"
SIBLING_DIR="internal/copilotprobe2"
DEPGUARD_DIAGNOSTIC_MARKER="copilot-sdk-boundary"

# Refuse to proceed if either fixture directory already exists (adversarial
# review finding, verified): os.MkdirAll/`mkdir -p` succeeds silently on an
# already-existing directory, and the unconditional `rm -rf` cleanup below
# would then delete any PRE-EXISTING, unrelated content at these
# deterministic, guessable paths. This script may only create and later
# remove directories it is certain it itself created.
if [ -e "$OUTSIDE_DIR" ]; then
  echo "::error::refusing to proceed: $OUTSIDE_DIR already exists (would risk deleting unrelated content on cleanup)" >&2
  exit 2
fi
if [ -e "$SIBLING_DIR" ]; then
  echo "::error::refusing to proceed: $SIBLING_DIR already exists (would risk deleting unrelated content on cleanup)" >&2
  exit 2
fi

# Scratch log directory via mktemp (adversarial review finding, verified):
# fixed, predictable /tmp paths are a classic symlink-race / plant hazard
# on a shared runner, unlike every other script in this shipment, which
# already uses mktemp/tempfile.TemporaryDirectory for scratch state.
LOG_DIR="$(mktemp -d)"

cleanup() {
  rm -rf -- "$OUTSIDE_DIR" "$SIBLING_DIR" "$LOG_DIR"
}
trap cleanup EXIT

mkdir -p "$OUTSIDE_DIR" "$SIBLING_DIR"
cp scripts/testdata/depguard/violation-outside-allowlist.go.tmpl "$OUTSIDE_DIR/violation.go"
cp scripts/testdata/depguard/violation-sibling-directory.go.tmpl "$SIBLING_DIR/violation.go"

overall_status=0

# assert_depguard_rejection DIR LOG_FILE LABEL
#
# Runs golangci-lint against DIR and requires BOTH a nonzero exit AND the
# specific depguard copilot-sdk-boundary diagnostic identity in the
# captured log -- an unrelated failure (typecheck error, missing
# dependency, config error) must not be silently accepted as proof this
# rule specifically fired.
assert_depguard_rejection() {
  local dir="$1" log_file="$2" label="$3"

  if golangci-lint run "./${dir}/..." >"$log_file" 2>&1; then
    echo "FAIL: golangci-lint accepted ${label} (deny direction NOT proven -- lint exited 0)"
    cat "$log_file"
    return 1
  fi

  if ! grep -q "$DEPGUARD_DIAGNOSTIC_MARKER" "$log_file"; then
    echo "FAIL: golangci-lint rejected ${label}, but NOT with the depguard '${DEPGUARD_DIAGNOSTIC_MARKER}' diagnostic -- the rejection may be an unrelated lint/typecheck/config failure, which does not prove this rule specifically fired:"
    cat "$log_file"
    return 1
  fi

  echo "PASS: golangci-lint rejected ${label} specifically via depguard's ${DEPGUARD_DIAGNOSTIC_MARKER} diagnostic, as expected"
  return 0
}

if ! assert_depguard_rejection "$OUTSIDE_DIR" "$LOG_DIR/outside.log" "a Copilot-SDK import OUTSIDE the internal/copilotprobe allowlist"; then
  overall_status=1
fi

if ! assert_depguard_rejection "$SIBLING_DIR" "$LOG_DIR/sibling.log" "a Copilot-SDK import under a SIBLING directory name"; then
  overall_status=1
fi

# Positive-direction assertion: the real, allowlisted internal/copilotprobe
# package (client.go genuinely imports the SDK) must PASS lint -- proving
# the allowlist is not so narrow it rejects its own legitimate exemption.
if golangci-lint run "./internal/copilotprobe/..." >"$LOG_DIR/positive.log" 2>&1; then
  echo "PASS: golangci-lint accepts internal/copilotprobe's own genuine Copilot-SDK import, as expected"
else
  echo "FAIL: golangci-lint rejected the real, allowlisted internal/copilotprobe package -- the allowlist itself is broken:"
  cat "$LOG_DIR/positive.log"
  overall_status=1
fi

if [ "$overall_status" -eq 0 ]; then
  echo "check-depguard-fixtures: PASS (both deny directions and the positive allowlist direction proven)"
else
  echo "::error::check-depguard-fixtures: FAIL"
fi
exit "$overall_status"

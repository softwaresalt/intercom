#!/usr/bin/env bash
set -euo pipefail

# check-depguard-fixtures.sh
#
# Durable proof for the depguard Copilot-SDK-boundary rule (011.014-T,
# resolves 309FBF5A(ii), .golangci.yml `copilot-sdk-boundary`). A one-shot
# manual demonstration cannot detect future config drift (e.g. depguard
# silently falling out of `linters.enable`, or the `files` pattern silently
# widening) -- this script re-proves both DENY directions on every run:
#
#   (i)  a file OUTSIDE the internal/copilotprobe allowlist that imports the
#        Copilot SDK root package must make golangci-lint exit non-zero.
#   (ii) a file under a SIBLING directory name (internal/copilotprobe2/)
#        must ALSO make golangci-lint exit non-zero, proving the allowlist
#        pattern is anchored to the exact directory and has not silently
#        widened to match any "copilotprobe*"-prefixed name.
#
# Fixtures are committed as .go.tmpl templates under
# scripts/testdata/depguard/ (the Go toolchain ignores non-.go extensions)
# and copied at runtime into real, temporary, git-ignored locations so
# golangci-lint's actual import-resolution and depguard path-matching run
# for real, then removed on exit regardless of outcome.
#
# Usage:
#   scripts/check-depguard-fixtures.sh
#     Runs both fixture checks. Exits 0 only if BOTH fixtures are
#     REJECTED by golangci-lint (deny direction proven both ways).

ROOT="$(git rev-parse --show-toplevel)"
cd "$ROOT"

if ! command -v golangci-lint >/dev/null 2>&1; then
  echo "golangci-lint is required on PATH" >&2
  exit 2
fi

OUTSIDE_DIR="internal/depguardfixture"
SIBLING_DIR="internal/copilotprobe2"

cleanup() {
  rm -rf -- "$OUTSIDE_DIR" "$SIBLING_DIR"
}
trap cleanup EXIT

mkdir -p "$OUTSIDE_DIR" "$SIBLING_DIR"
cp scripts/testdata/depguard/violation-outside-allowlist.go.tmpl "$OUTSIDE_DIR/violation.go"
cp scripts/testdata/depguard/violation-sibling-directory.go.tmpl "$SIBLING_DIR/violation.go"

overall_status=0

if golangci-lint run "./${OUTSIDE_DIR}/..." >/tmp/depguard-outside.log 2>&1; then
  echo "FAIL: golangci-lint accepted a Copilot-SDK import OUTSIDE the internal/copilotprobe allowlist (deny direction NOT proven)"
  cat /tmp/depguard-outside.log
  overall_status=1
else
  echo "PASS: golangci-lint rejected the outside-allowlist fixture as expected"
fi

if golangci-lint run "./${SIBLING_DIR}/..." >/tmp/depguard-sibling.log 2>&1; then
  echo "FAIL: golangci-lint accepted a Copilot-SDK import under a SIBLING directory name (allowlist anchoring NOT proven)"
  cat /tmp/depguard-sibling.log
  overall_status=1
else
  echo "PASS: golangci-lint rejected the sibling-directory fixture as expected"
fi

rm -f /tmp/depguard-outside.log /tmp/depguard-sibling.log

if [ "$overall_status" -eq 0 ]; then
  echo "check-depguard-fixtures: PASS (both deny directions proven)"
else
  echo "::error::check-depguard-fixtures: FAIL"
fi
exit "$overall_status"

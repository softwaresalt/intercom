// Package writepath is a fixture used only by
// scripts/check-write-path-precondition.sh --self-test. It lives under a
// "testdata" directory, which the Go toolchain always ignores for
// build/vet/test/list purposes, so it is never compiled as part of
// `go build ./...` despite this scan input existing purely to be read as
// text by the detector, not imported by anything.
package writepath

func CleanHelper(a, b int) int {
	return a + b
}

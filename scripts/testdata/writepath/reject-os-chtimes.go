// Package writepath is a fixture used only by
// scripts/check-write-path-precondition.sh --self-test. It lives under a
// "testdata" directory, which the Go toolchain always ignores for
// build/vet/test/list purposes, so it is never compiled as part of
// `go build ./...` despite containing a package-qualified write primitive.
//
// 034.009-T (D-T6, AC-D6.1, AC-D6.3): this fixture pins regression detection
// for os.Chtimes in the 049-F go/ast verdict-parity corpus (stash 8E9F8E55).
package writepath

import (
	"os"
	"time"
)

func touch(name string, t time.Time) error {
	return os.Chtimes(name, t, t)
}

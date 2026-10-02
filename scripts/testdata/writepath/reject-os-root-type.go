// Package writepath is a fixture used only by
// scripts/check-write-path-precondition.sh --self-test. It lives under a
// "testdata" directory, which the Go toolchain always ignores for
// build/vet/test/list purposes, so it is never compiled as part of
// `go build ./...` despite containing a package-qualified write primitive.
//
// 034.011-T (D-T5b, AC-D5b.1, AC-D5b.3): this fixture is part of the
// verdict-parity corpus for the 049-F go/ast migration. It pins presence
// detection for the os.Root type selector on its own: an os.Root-typed
// parameter with no os.OpenRoot call anywhere in the file's code
// (stash B72E9715).
package writepath

import "os"

func rootName(r *os.Root) string {
	return r.Name()
}

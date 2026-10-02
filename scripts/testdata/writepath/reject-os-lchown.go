// Package writepath is a fixture used only by
// scripts/check-write-path-precondition.sh --self-test. It lives under a
// "testdata" directory, which the Go toolchain always ignores for
// build/vet/test/list purposes, so it is never compiled as part of
// `go build ./...` despite containing a package-qualified write primitive.
//
// 034.009-T (D-T6, AC-D6.1, AC-D6.3): this fixture is part of the
// verdict-parity corpus for the 049-F go/ast migration. It pins regression
// detection for the pre-existing os.Lchown selector (stash 8E9F8E55).
package writepath

import "os"

func setLinkOwner(name string, uid, gid int) error {
	return os.Lchown(name, uid, gid)
}

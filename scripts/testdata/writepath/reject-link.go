// Package writepath is a fixture used only by
// scripts/check-write-path-precondition.sh --self-test. It lives under a
// "testdata" directory, which the Go toolchain always ignores for
// build/vet/test/list purposes, so it is never compiled as part of
// `go build ./...` despite containing a package-qualified write primitive.
package writepath

import "os"

func LinkSomething(oldname, newname string) error {
	return os.Link(oldname, newname)
}

// Package writepath is a fixture used only by
// scripts/check-write-path-precondition.sh --self-test. It lives under a
// "testdata" directory, which the Go toolchain always ignores for
// build/vet/test/list purposes, so it is never compiled as part of
// `go build ./...` despite containing a package-qualified write primitive.
//
// 034.011-T (D-T5b, AC-D5b.1, AC-D5b.3): this fixture is part of the
// verdict-parity corpus for the 049-F go/ast migration. It pins presence
// detection for the io.CopyN and io.CopyBuffer selectors, one per line, so
// each is reported independently and neither is mistaken for io.Copy
// (stash B72E9715).
package writepath

import "io"

func copies(dst io.Writer, src io.Reader, buf []byte) error {
	if _, err := io.CopyN(dst, src, 8); err != nil {
		return err
	}
	_, err := io.CopyBuffer(dst, src, buf)
	return err
}

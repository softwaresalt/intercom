// Package writepath is a fixture used only by
// scripts/check-write-path-precondition.sh --self-test. It lives under a
// "testdata" directory, which the Go toolchain always ignores for
// build/vet/test/list purposes, so it is never compiled as part of
// `go build ./...` despite containing a package-qualified write primitive.
//
// 034.008-T (D-T5a, AC-D5.1, AC-D5.6): this fixture is part of the
// verdict-parity corpus for the 049-F go/ast migration. It is the positive
// control proving the D-2' allowance is per-occurrence, not file-level:
// the metadata-only call in metadataOnly must produce no finding, the
// GENERIC_WRITE call in writeHandle must be reported at its own line, and
// the single line in mixedLine holding one allowed call and one writing
// call must be reported because its writing occurrence is not exempt.
package writepath

import "syscall"

func metadataOnly(pathPtr *uint16) (syscall.Handle, error) {
	return syscall.CreateFile(
		pathPtr,
		0,
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
}

func writeHandle(pathPtr *uint16) (syscall.Handle, error) {
	return syscall.CreateFile(
		pathPtr,
		syscall.GENERIC_WRITE,
		0,
		nil,
		syscall.CREATE_ALWAYS,
		syscall.FILE_ATTRIBUTE_NORMAL,
		0,
	)
}

func keep(_ syscall.Handle, err error) error { return err }

func mixedLine(pathPtr *uint16) (error, error) {
	return keep(syscall.CreateFile(pathPtr, 0, 0, nil, syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS, 0)), keep(syscall.CreateFile(pathPtr, syscall.GENERIC_WRITE, 0, nil, syscall.CREATE_ALWAYS, 0, 0))
}

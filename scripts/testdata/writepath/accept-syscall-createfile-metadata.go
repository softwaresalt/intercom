// Package writepath is a fixture used only by
// scripts/check-write-path-precondition.sh --self-test. It lives under a
// "testdata" directory, which the Go toolchain always ignores for
// build/vet/test/list purposes, so it is never compiled as part of
// `go build ./...`.
//
// 034.008-T (D-T5a, AC-D5.1, AC-D5.6): this fixture is part of the
// verdict-parity corpus for the 049-F go/ast migration. It holds the D-2'
// access-mode allowance's single admitted shape on its own: the
// metadata-only syscall.CreateFile call that
// internal/pathsafe/reparse_windows.go makes (access 0, OPEN_EXISTING and
// only FILE_FLAG_BACKUP_SEMANTICS). The gate must leave it clean, so
// disabling the predicate flips this fixture to rejected.
package writepath

import "syscall"

func metadataOnlyHandle(pathPtr *uint16) (syscall.Handle, error) {
	return syscall.CreateFile(
		pathPtr,
		0, // no data access requested; this is a metadata-only query
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
}

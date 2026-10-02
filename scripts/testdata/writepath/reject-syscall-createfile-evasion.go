// Package writepath is a fixture used only by
// scripts/check-write-path-precondition.sh --self-test. It lives under a
// "testdata" directory, which the Go toolchain always ignores for
// build/vet/test/list purposes, so it is never compiled as part of
// `go build ./...` despite containing a package-qualified write primitive.
//
// 034.008-T (D-T5a, AC-D5.1, AC-D5.6): this fixture is part of the
// verdict-parity corpus for the 049-F go/ast migration. Each line in
// evasions holds one D-2' rejection shape that sits next to the admitted
// metadata-only shape but is not exempt from the allowance: a CREATE_NEW
// disposition, a FILE_FLAG_DELETE_ON_CLOSE flag, an OR-ed flag
// combination, a brace-composite multi-value argument, a tag-shaped
// raw-string argument, and a function-value reference. Widening the
// predicate to admit any of them drops that line's golden finding.
package writepath

import "syscall"

// fileFlagDeleteOnClose is FILE_FLAG_DELETE_ON_CLOSE, which the syscall
// package does not export.
const fileFlagDeleteOnClose = 0x04000000

func evasions(p *uint16) {
	_, _ = syscall.CreateFile(p, 0, 0, nil, syscall.CREATE_NEW, syscall.FILE_FLAG_BACKUP_SEMANTICS, 0)
	_, _ = syscall.CreateFile(p, 0, 0, nil, syscall.OPEN_EXISTING, fileFlagDeleteOnClose, 0)
	_, _ = syscall.CreateFile(p, 0, 0, nil, syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS|syscall.FILE_FLAG_OPEN_REPARSE_POINT, 0)
	_, _ = syscall.CreateFile(p, 0, 0, &syscall.SecurityAttributes{Length: 0, InheritHandle: 1}, syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS, 0)
	_, _ = syscall.CreateFile(`x:"y"`, 0, 0, nil, syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS, 0)
	open := syscall.CreateFile
	_ = open
}

//go:build windows

package pathsafe

import (
	"syscall"
	"unsafe"

	"github.com/softwaresalt/intercom-go/internal/apperr"
)

// modkernel32 / procGetFinalPathNameByHandleW resolve GetFinalPathNameByHandleW
// via syscall.NewLazyDLL rather than golang.org/x/sys/windows (014.003-T AC3,
// D-3): this package must not add a new module dependency for a single Win32
// call.
var (
	modkernel32                   = syscall.NewLazyDLL("kernel32.dll")
	procGetFinalPathNameByHandleW = modkernel32.NewProc("GetFinalPathNameByHandleW")
)

// canonicalizeReparse resolves path to its true filesystem target using
// GetFinalPathNameByHandleW semantics (014.003-T). Unlike
// filepath.EvalSymlinks, opening a handle with CreateFileW and
// FILE_FLAG_BACKUP_SEMANTICS (required to obtain a traversable handle for a
// directory) and then querying GetFinalPathNameByHandleW resolves
// IO_REPARSE_TAG_MOUNT_POINT (directory junctions) in addition to
// IO_REPARSE_TAG_SYMLINK, because CreateFileW's default (non
// FILE_FLAG_OPEN_REPARSE_POINT) open mode transparently follows any reparse
// point encountered while resolving the path, and the returned handle then
// refers to the real, final target -- regardless of whether the reparse
// point was an intermediate or the terminal path component.
//
// path must name an existing filesystem entry; a non-existent path returns
// the OS error (typically ERROR_FILE_NOT_FOUND / ERROR_PATH_NOT_FOUND)
// wrapped for errors.Is/errors.As inspection by the caller.
//
// The handle opened here is closed on every return path, including error
// paths (014.003-T AC1 / plan H11), via a single deferred CloseHandle.
func canonicalizeReparse(path string) (string, error) {
	// U7 (016.002-T): resolve procGetFinalPathNameByHandleW's lazy export
	// explicitly and up front, before syscall.CreateFile ever opens a
	// handle, so a missing export surfaces as a returned
	// apperr.KindPathViolation instead of panicking through
	// LazyProc.Call's internal mustFind. DOCUMENTED-UNREACHABLE on every
	// supported target: GetFinalPathNameByHandleW has shipped in
	// kernel32.dll since Windows Vista / Server 2008, and this module's Go
	// 1.24 floor requires Windows 10 / Server 2016+, so this Find() cannot
	// fail in practice. Find() transitively covers both the kernel32.dll
	// load and the export lookup via LazyDLL.Load()'s internal sync.Once,
	// so no separate modkernel32 guard is required.
	if err := procGetFinalPathNameByHandleW.Find(); err != nil {
		return "", apperr.Wrapf(apperr.KindPathViolation, err, "GetFinalPathNameByHandleW unavailable: %s", err.Error())
	}

	pathPtr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}

	handle, err := syscall.CreateFile(
		pathPtr,
		0, // no data access requested; this is a metadata-only query
		syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE,
		nil,
		syscall.OPEN_EXISTING,
		syscall.FILE_FLAG_BACKUP_SEMANTICS,
		0,
	)
	if err != nil {
		return "", err
	}
	// The close error is deliberately, explicitly discarded (not merely
	// unchecked): this is a metadata-only, read-only handle opened solely
	// to query GetFinalPathNameByHandleW below; by the time this deferred
	// close runs, that query has already completed (successfully or not)
	// and its result is already being returned to the caller. A close
	// failure here cannot retroactively invalidate data already read, and
	// there is no caller-actionable response to a failed handle close on
	// this path -- silently leaking the report through the function's own
	// return value would incorrectly overwrite (or mask) the real
	// getFinalPathNameByHandle result/error with an unrelated close
	// failure. errcheck (2026-09-08 review finding) flagged the bare
	// `defer syscall.CloseHandle(handle)` form; this explicit discard
	// keeps identical runtime behavior while satisfying the linter.
	defer func() { _ = syscall.CloseHandle(handle) }()

	return getFinalPathNameByHandle(handle)
}

// getFinalPathNameByHandle calls GetFinalPathNameByHandleW, growing the
// output buffer if the initial MAX_PATH-sized buffer is insufficient (the
// API reports the required buffer length, including the NUL terminator, as
// its return value when the supplied buffer is too small).
func getFinalPathNameByHandle(handle syscall.Handle) (string, error) {
	buf := make([]uint16, syscall.MAX_PATH)
	for {
		n, _, callErr := procGetFinalPathNameByHandleW.Call(
			uintptr(handle),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(len(buf)),
			0,
		)
		if n == 0 {
			return "", callErr
		}
		if int(n) >= len(buf) {
			buf = make([]uint16, n+1)
			continue
		}
		return syscall.UTF16ToString(buf[:n]), nil
	}
}

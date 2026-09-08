//go:build windows

package pathsafe

import (
	"syscall"
	"unsafe"
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
	defer syscall.CloseHandle(handle)

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

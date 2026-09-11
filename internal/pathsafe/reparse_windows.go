//go:build windows

package pathsafe

import (
	"path/filepath"
	"strings"
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

// longPathThreshold is the classic MAX_PATH limit (260, including the NUL
// terminator) that syscall.CreateFile does not itself handle by extending
// with a `\\?\` prefix -- unlike Go's own os package, which applies
// equivalent long-path handling internally for its own high-level API
// (os.Open, os.Mkdir, etc. via the unexported os.fixLongPath). A raw
// syscall.CreateFile call, as canonicalizeReparse makes below, gets none of
// that help.
const longPathThreshold = syscall.MAX_PATH

// addLongPathPrefix prepends the `\\?\` extended-length-path prefix before
// canonicalizeReparse calls syscall.CreateFile (016.004-T, U5), which --
// unlike Go's os package -- performs no long-path handling of its own. The
// prefix is applied only when path is ALL of:
//   - at or beyond the longPathThreshold (260-char MAX_PATH) -- below that,
//     ordinary paths need no adjustment;
//   - absolute (filepath.IsAbs) -- a relative path cannot be meaningfully
//     extended-prefixed;
//   - not already `\\?\`-prefixed -- covers both an ordinary extended-length
//     path and a non-strippable `\\?\Volume{GUID}\...` / `\\?\GLOBALROOT\...`
//     form; double-prefixing corrupts the path;
//   - not a `\\.\` device-namespace path -- these already bypass Win32 path
//     parsing and must never be extended-prefixed;
//   - not a bare UNC path (`\\server\share\...`) -- converting a UNC path to
//     its extended form requires the distinct `\\?\UNC\server\share\...`
//     form, which this shipment deliberately does NOT implement. This is an
//     explicitly recorded residual (U5/AC5): a long UNC-rooted workspace
//     path remains subject to the pre-existing MAX_PATH limitation, tracked
//     in the package risk register.
//
// D-4 (017.002-T): the `\\.\` guard IS reachable -- filepath.IsAbs reports
// true for every measured `\\.\` form (`\\.\C:\foo`, `\\.\PhysicalDrive0`,
// `\\.\UNC\srv\sh\x`, and a `\\.\C:`-rooted path past longPathThreshold), so
// the preceding `!filepath.IsAbs(path)` guard never shadows it. It is,
// however, currently BEHAVIORALLY REDUNDANT: the immediately-following bare
// `\\`-prefix guard returns the identical value for any `\\.\` input, so
// deleting this branch today would be a no-op. It is kept anyway as
// self-documenting, order-independent defense-in-depth against a future
// change to the generic `\\` guard -- see
// TestAddLongPathPrefixDeviceNamespaceBranchReachability
// (reparse_windows_test.go) for the characterization lock and the measured
// reachability evidence. Do NOT delete this branch, and do NOT replace it
// with a panic/unreachability assertion (that option was considered and
// rejected: the "unreachable" premise it would encode is false).
//
// Applying `\\?\` disables Win32 path normalization, so this must only run
// on an already Abs/Clean'd path -- true at canonicalizeReparse's callers
// (checkSymlinkEscape in pathsafe.go, and NewRoot as of 016.007-T), both of
// which pass filepath.Abs'd input.
//
// 017.003-T: the threshold comparison uses utf16Len (UTF-16 code-unit
// count), not len(path) (UTF-8 byte count), because that is what Windows
// itself measures against MAX_PATH. A UTF-8 byte count is a conservative
// proxy -- it is always >= the true UTF-16 code-unit count (1-byte ASCII ->
// 1 unit; 2/3-byte BMP characters -> 1 unit; 4-byte characters -> a 2-unit
// surrogate pair) -- so the pre-017.003-T proxy could only ever produce a
// false positive (prefixing a path that did not need it), never a false
// negative. This is a precision improvement, not a correctness or security
// fix (plan D-5): framing it as a vulnerability fix is an explicit
// anti-goal.
func addLongPathPrefix(path string) string {
	if utf16Len(path) < longPathThreshold {
		return path
	}
	if !filepath.IsAbs(path) {
		return path
	}
	if strings.HasPrefix(path, uncPrefix) {
		return path
	}
	if strings.HasPrefix(path, `\\.\`) {
		return path
	}
	if strings.HasPrefix(path, `\\`) {
		return path
	}
	return uncPrefix + path
}

// utf16Len returns the number of UTF-16 code units path would occupy when
// encoded for a Win32 call -- the same unit Windows measures MAX_PATH in --
// without allocating the intermediate []rune / []uint16 slices that
// len(utf16.Encode([]rune(path))) would (017.003-T, plan O2-b). Every
// codepoint above U+FFFF requires a 2-unit UTF-16 surrogate pair; every
// other valid codepoint requires exactly 1 unit. A plain rune count
// (rejected as plan option O2-a) would UNDER-count any character above
// U+FFFF and so could introduce a false negative the byte-length proxy
// this replaces never had -- see
// TestUTF16LenMatchesStdlibEncoding (reparse_windows_test.go), which pins
// this function's agreement with the stdlib utf16.Encode reference
// computation.
func utf16Len(path string) int {
	n := 0
	for _, r := range path {
		if r > 0xFFFF {
			n += 2
		} else {
			n++
		}
	}
	return n
}

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

	pathPtr, err := syscall.UTF16PtrFromString(addLongPathPrefix(path))
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

	resolved, err := getFinalPathNameByHandle(handle)
	if err != nil {
		return "", err
	}
	// U6 (016.003-T): internalize the stripUNCPrefix postcondition here so
	// it holds for ANY caller, not just ones that remember to strip it
	// themselves. GetFinalPathNameByHandleW's raw result always carries a
	// leading `\\?\` extended-path prefix; stripUNCPrefix is idempotent
	// (verified: an already-unprefixed input returns unchanged) and
	// preserves a non-strippable `\\?\Volume{GUID}\...` / `\\?\GLOBALROOT\...`
	// form unchanged (D-6) -- so this call cannot introduce a new failure
	// mode for any input this function can produce.
	return stripUNCPrefix(resolved), nil
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

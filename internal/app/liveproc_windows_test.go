//go:build windows

package app

// Windows process-table snapshot for the live-SDK test's "no stranded CLI
// child" assertion (see liveproc_test.go). It uses only the standard library
// syscall package (no golang.org/x/sys, no go.mod change) and replaces a
// powershell.exe Get-CimInstance subprocess:
//
//  1. CreateToolhelp32Snapshot / Process32First / Process32Next yield pid,
//     parent pid, and executable name for every process.
//  2. For each process the FULL COMMAND LINE is read from its own memory:
//     OpenProcess -> ntdll!NtQueryInformationProcess(ProcessBasicInformation)
//     -> PEB.ProcessParameters -> RTL_USER_PROCESS_PARAMETERS.CommandLine (a
//     UNICODE_STRING) -> kernel32!ReadProcessMemory. The command line is what
//     lets newChildren's needle rule find a child whose launcher already
//     exited (a reparented node.exe ...\copilot\index.js), whose executable
//     name alone never contains the needle.
//  3. Only when the process already exited, access is denied, or the target's
//     layout cannot be read (a process of different bitness than this test
//     binary, or a failed/implausible read) does the entry fall back to the
//     executable name.
//
// The PEB offsets are those of the CURRENT architecture (this test binary's
// pointer size); see commandLineOffsets.

import (
	"context"
	"errors"
	"fmt"
	"syscall"
	"unsafe"
)

const (
	processVMRead                  = 0x0010
	processQueryLimitedInformation = 0x1000

	// processBasicInformationClass is PROCESSINFOCLASS ProcessBasicInformation.
	processBasicInformationClass = 0

	// maxCommandLineBytes bounds the UNICODE_STRING we are willing to read:
	// CreateProcess limits a command line to 32767 UTF-16 code units.
	maxCommandLineBytes = 32767 * 2

	ptrSize = unsafe.Sizeof(uintptr(0))
)

var (
	ntdll                         = syscall.NewLazyDLL("ntdll.dll")
	kernel32                      = syscall.NewLazyDLL("kernel32.dll")
	procNtQueryInformationProcess = ntdll.NewProc("NtQueryInformationProcess")
	procReadProcessMemory         = kernel32.NewProc("ReadProcessMemory")
)

// processBasicInformation is PROCESS_BASIC_INFORMATION (winternl.h):
// Reserved1 holds the NTSTATUS ExitStatus, Reserved2 is two reserved
// pointers, and the layout is pointer-size aligned on both 32- and 64-bit.
type processBasicInformation struct {
	Reserved1       uintptr
	PebBaseAddress  uintptr
	Reserved2       [2]uintptr
	UniqueProcessID uintptr
	Reserved3       uintptr
}

// unicodeString is UNICODE_STRING; Go's natural alignment reproduces the
// 4-byte padding before Buffer on 64-bit.
type unicodeString struct {
	Length        uint16 // bytes, excluding any terminator
	MaximumLength uint16 // bytes
	Buffer        uintptr
}

// commandLineOffsets returns, for this binary's architecture, the offset of
// PEB.ProcessParameters and of RTL_USER_PROCESS_PARAMETERS.CommandLine.
func commandLineOffsets() (processParameters, commandLine uintptr) {
	if ptrSize == 8 {
		return 0x20, 0x70
	}
	return 0x10, 0x40
}

// listProcesses returns the process table with each entry's cmd set to the
// full command line when readable, else the executable name.
func listProcesses(ctx context.Context) ([]procEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	snap, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return nil, fmt.Errorf("process snapshot: %w", err)
	}
	defer func() { _ = syscall.CloseHandle(snap) }()

	var entry syscall.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))
	var out []procEntry
	err = syscall.Process32First(snap, &entry)
	for err == nil {
		if cerr := ctx.Err(); cerr != nil {
			return nil, cerr
		}
		cmd, ok := processCommandLine(entry.ProcessID)
		if !ok {
			cmd = syscall.UTF16ToString(entry.ExeFile[:])
		}
		out = append(out, procEntry{pid: int(entry.ProcessID), ppid: int(entry.ParentProcessID), cmd: cmd})
		err = syscall.Process32Next(snap, &entry)
	}
	if !errors.Is(err, syscall.ERROR_NO_MORE_FILES) {
		return nil, fmt.Errorf("enumerate processes: %w", err)
	}
	return out, nil
}

// processCommandLine reads pid's full command line from its PEB. ok is false
// when the process is gone, access is denied, the target's PEB layout is not
// this binary's (different bitness), or the data fails validation.
func processCommandLine(pid uint32) (cmd string, ok bool) {
	if pid == 0 { // System Idle Process has no PEB
		return "", false
	}
	h, err := syscall.OpenProcess(processQueryLimitedInformation|processVMRead, false, pid)
	if err != nil {
		return "", false
	}
	defer func() { _ = syscall.CloseHandle(h) }()

	var pbi processBasicInformation
	var retLen uint32
	status, _, _ := procNtQueryInformationProcess.Call(
		uintptr(h), processBasicInformationClass,
		uintptr(unsafe.Pointer(&pbi)), unsafe.Sizeof(pbi), uintptr(unsafe.Pointer(&retLen)))
	if uint32(status) != 0 || pbi.PebBaseAddress == 0 { // NTSTATUS != STATUS_SUCCESS
		return "", false
	}

	paramsOff, cmdOff := commandLineOffsets()
	var params uintptr
	if !readProcessMemory(h, pbi.PebBaseAddress+paramsOff, unsafe.Pointer(&params), ptrSize) || params == 0 {
		return "", false
	}
	var us unicodeString
	if !readProcessMemory(h, params+cmdOff, unsafe.Pointer(&us), unsafe.Sizeof(us)) {
		return "", false
	}
	if us.Buffer == 0 || us.Length == 0 || us.Length%2 != 0 ||
		us.Length > us.MaximumLength || us.Length > maxCommandLineBytes {
		return "", false
	}
	buf := make([]uint16, us.Length/2)
	if !readProcessMemory(h, us.Buffer, unsafe.Pointer(&buf[0]), uintptr(us.Length)) {
		return "", false
	}
	cmd = syscall.UTF16ToString(buf)
	return cmd, cmd != ""
}

// readProcessMemory copies exactly size bytes at addr in process h into dst.
func readProcessMemory(h syscall.Handle, addr uintptr, dst unsafe.Pointer, size uintptr) bool {
	var n uintptr
	r, _, _ := procReadProcessMemory.Call(uintptr(h), addr, uintptr(dst), size, uintptr(unsafe.Pointer(&n)))
	return r != 0 && n == size
}

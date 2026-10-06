package p

import (
	"golang.org/x/sys/unix"
	win "golang.org/x/sys/windows"
)

func windowsPrimitives() {
	win.WriteFile()
	win.CreateFile()
	win.DeleteFile()
	win.MoveFile()
	win.MoveFileEx()
	win.CreateDirectory()
	win.RemoveDirectory()
	win.CreateHardLink()
	win.CreateSymbolicLink()
	win.SetEndOfFile()
	win.SetFileInformationByHandle()
}

func unixPrimitives() {
	unix.Open()
	unix.Openat()
	unix.Openat2()
	unix.Creat()
	unix.Write()
	unix.Pwrite()
	unix.Unlink()
	unix.Unlinkat()
	unix.Rename()
	unix.Renameat()
	unix.Renameat2()
	unix.Mkdir()
	unix.Mkdirat()
	unix.Rmdir()
	unix.Link()
	unix.Linkat()
	unix.Symlink()
	unix.Symlinkat()
	unix.Truncate()
	unix.Ftruncate()
	unix.Chmod()
	unix.Fchmodat()
}

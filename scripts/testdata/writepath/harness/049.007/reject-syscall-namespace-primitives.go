package p

import "syscall"

func syscallPrimitives() {
	syscall.WriteFile()
	syscall.Open()
	syscall.Unlink()
	syscall.Rename()
	syscall.Mkdir()
	syscall.Rmdir()
	syscall.CreateHardLink()
	syscall.DeleteFile()
	syscall.MoveFile()
	syscall.RemoveDirectory()
	syscall.CreateDirectory()
	syscall.CreateSymbolicLink()
	syscall.Truncate()
	syscall.Creat()
}

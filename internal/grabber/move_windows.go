//go:build windows
// +build windows

package grabber

import (
	"golang.org/x/sys/windows"
)

func moveFile(oldpath, newpath string) error {
	from, err := windows.UTF16PtrFromString(oldpath)
	if err != nil {
		return err
	}
	to, err := windows.UTF16PtrFromString(newpath)
	if err != nil {
		return err
	}
	return windows.MoveFile(from, to)
}

// +build !windows

package grabber

import (
	"os"
)

func moveFile(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

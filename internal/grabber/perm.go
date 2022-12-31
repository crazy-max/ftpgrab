//go:build !windows
// +build !windows

package grabber

import (
	"os"
)

func (c *Client) fixPerms(filepath string) error {
	fileinfo, err := os.Stat(filepath)
	if err != nil {
		return err
	}
	chmod := c.config.ChmodFile
	if fileinfo.IsDir() {
		chmod = c.config.ChmodDir
	}
	if err = os.Chmod(filepath, chmod); err != nil {
		return err
	}
	return os.Chown(filepath, c.config.UID, c.config.GID)
}

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

	chmod := os.FileMode(c.config.ChmodFile)
	if fileinfo.IsDir() {
		chmod = os.FileMode(c.config.ChmodDir)
	}

	if err := os.Chmod(filepath, chmod); err != nil {
		return err
	}

	if err := os.Chown(filepath, c.config.UID, c.config.GID); err != nil {
		return err
	}

	return nil
}

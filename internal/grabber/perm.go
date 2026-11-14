//go:build !windows

package grabber

import (
	stderrors "errors"
	"os"
)

func (c *Client) fixDirPerms(filepath string) error {
	if err := c.root.Chmod(filepath, c.config.ChmodDir); err != nil {
		return err
	}
	return c.root.Chown(filepath, c.config.UID, c.config.GID)
}

func (c *Client) fixFilePerms(file *os.File) error {
	return stderrors.Join(file.Chown(c.config.UID, c.config.GID), file.Chmod(c.config.ChmodFile))
}

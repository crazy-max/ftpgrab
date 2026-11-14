//go:build !windows

package grabber

func (c *Client) fixPerms(filepath string) error {
	fileinfo, err := c.root.Stat(filepath)
	if err != nil {
		return err
	}
	chmod := c.config.ChmodFile
	if fileinfo.IsDir() {
		chmod = c.config.ChmodDir
	}
	if err = c.root.Chmod(filepath, chmod); err != nil {
		return err
	}
	return c.root.Chown(filepath, c.config.UID, c.config.GID)
}

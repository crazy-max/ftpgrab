//go:build windows

package grabber

import "os"

func (c *Client) fixDirPerms(_ string) error {
	return nil
}

func (c *Client) fixFilePerms(_ *os.File) error {
	return nil
}

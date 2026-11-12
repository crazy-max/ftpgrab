//go:build windows

package grabber

func (c *Client) fixPerms(_ string) error {
	return nil
}

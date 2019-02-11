package journal

import (
	"github.com/ftpgrab/ftpgrab/internal/model"
)

// Client represents an active journal object
type Client struct {
	*model.Journal
}

// New creates new journal instance
func New() *Client {
	return &Client{&model.Journal{}}
}

// AddEntry adds an entry in the journal
func (c *Client) AddEntry(entry model.Entry) {
	c.Entries = append(c.Entries, entry)
	if entry.StatusType == "error" {
		c.Count.Error++
	} else if entry.StatusType == "skip" {
		c.Count.Skip++
	} else if entry.StatusType == "success" {
		c.Count.Success++
	}
}

// IsEmpty verifies if journal is empty
func (c *Client) IsEmpty() bool {
	return c.Count.Error == 0 && c.Count.Success == 0
}

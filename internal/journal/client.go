package journal

// Client represents an active journal object
type Client struct {
	Journal
}

// New creates new journal instance
func New() *Client {
	return &Client{Journal{}}
}

// Add adds an entry in the journal
func (c *Client) Add(entry Entry) {
	c.Entries = append(c.Entries, entry)
	switch entry.Level {
	case EntryLevelError:
		c.Count.Error++
	case EntryLevelSkip:
		c.Count.Skip++
	case EntryLevelSuccess:
		c.Count.Success++
	}
}

// IsEmpty checks if journal is empty
func (c *Client) IsEmpty() bool {
	return len(c.Entries) == 0
}

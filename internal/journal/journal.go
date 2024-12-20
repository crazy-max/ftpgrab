package journal

import (
	"encoding/json"
	"time"

	"github.com/hako/durafmt"
)

// Journal holds journal entries
type Journal struct {
	ServerHost string  `json:"-"`
	Entries    []Entry `json:"entries,omitempty"`
	Count      struct {
		Success int `json:"success,omitempty"`
		Error   int `json:"error,omitempty"`
		Skip    int `json:"skip,omitempty"`
	} `json:"count,omitempty"`
	Status   string        `json:"status,omitempty"`
	Duration time.Duration `json:"duration,omitempty"`
}

func (j Journal) MarshalJSON() ([]byte, error) {
	type Alias Journal
	return json.Marshal(&struct {
		Alias
		Duration string `json:"duration,omitempty"`
	}{
		Alias:    (Alias)(j),
		Duration: durafmt.ParseShort(j.Duration).String(),
	})
}

// IsEmpty checks if journal is empty
func (j Journal) IsEmpty() bool {
	return len(j.Entries) == 0
}

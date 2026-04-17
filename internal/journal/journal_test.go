package journal

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	client := New()

	client.Add(Entry{File: "/ok.txt", Level: EntryLevelSuccess})
	client.Add(Entry{File: "/skip.txt", Level: EntryLevelSkip})
	client.Add(Entry{File: "/err.txt", Level: EntryLevelError})
	client.Add(Entry{File: "/warn.txt", Level: EntryLevelWarning})

	require.Len(t, client.Entries, 4)
	assert.Equal(t, 1, client.Count.Success)
	assert.Equal(t, 1, client.Count.Skip)
	assert.Equal(t, 1, client.Count.Error)
}

func TestMarshalJSON(t *testing.T) {
	jnl := Journal{
		ServerHost: "ftp.example.com",
		Entries: []Entry{
			{File: "/file.txt", Level: EntryLevelSuccess},
		},
		Status:   "success",
		Duration: 90 * time.Second,
	}
	jnl.Count.Success = 1

	payload, err := jnl.MarshalJSON()
	require.NoError(t, err)

	var got map[string]any
	require.NoError(t, json.Unmarshal(payload, &got))
	assert.Equal(t, "1 minute", got["duration"])
	assert.Equal(t, "success", got["status"])
	assert.NotContains(t, string(payload), "ftp.example.com")
}

func TestIsSkipped(t *testing.T) {
	cases := []struct {
		name   string
		status EntryStatus
		want   bool
	}{
		{name: "already downloaded", status: EntryStatusAlreadyDl, want: true},
		{name: "hash exists", status: EntryStatusHashExists, want: true},
		{name: "outdated", status: EntryStatusOutdated, want: true},
		{name: "not included", status: EntryStatusNotIncluded, want: true},
		{name: "excluded", status: EntryStatusExcluded, want: true},
		{name: "never downloaded", status: EntryStatusNeverDl, want: false},
		{name: "size different", status: EntryStatusSizeDiff, want: false},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.IsSkipped())
		})
	}
}

func TestIsEmpty(t *testing.T) {
	client := New()
	assert.True(t, client.IsEmpty())
	assert.True(t, client.Journal.IsEmpty())

	client.Add(Entry{File: "/file.txt", Level: EntryLevelSuccess})

	assert.False(t, client.IsEmpty())
	assert.False(t, client.Journal.IsEmpty())
}

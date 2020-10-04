package journal

// Entry represents a journal entry
type Entry struct {
	File   string      `json:"file,omitempty"`
	Status EntryStatus `json:"status,omitempty"`
	Level  EntryLevel  `json:"level,omitempty"`
	Text   string      `json:"text,omitempty"`
}

//  EntryLevel represents an entry kevek
type EntryLevel string

const (
	EntryLevelError   = EntryLevel("error")
	EntryLevelWarning = EntryLevel("warning")
	EntryLevelSkip    = EntryLevel("skip")
	EntryLevelSuccess = EntryLevel("success")
)

// EntryStatus represents entry status
type EntryStatus string

const (
	EntryStatusOutdated    = EntryStatus("Outdated file")
	EntryStatusNotIncluded = EntryStatus("Not included")
	EntryStatusExcluded    = EntryStatus("Excluded")
	EntryStatusNeverDl     = EntryStatus("Never downloaded")
	EntryStatusAlreadyDl   = EntryStatus("Already downloaded")
	EntryStatusSizeDiff    = EntryStatus("Exists but size is different")
	EntryStatusHashExists  = EntryStatus("Hash sum exists")
)

func (es *EntryStatus) IsSkipped() bool {
	return *es == EntryStatusAlreadyDl ||
		*es == EntryStatusHashExists ||
		*es == EntryStatusOutdated ||
		*es == EntryStatusNotIncluded ||
		*es == EntryStatusExcluded
}

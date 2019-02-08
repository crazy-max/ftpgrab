package model

import "time"

// Journal holds ftpgrab entries and status
type Journal struct {
	Entries []Entry
	Count   struct {
		Success int
		Error   int
		Skip    int
	}
	Status   string
	Duration time.Duration
}

// Entry represents a journal entry
type Entry struct {
	File       string
	StatusType string
	StatusText string
}

// EntryStatus represents entry status
type EntryStatus string

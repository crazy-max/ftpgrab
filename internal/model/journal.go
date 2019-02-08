package model

import "time"

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

type Entry struct {
	File       string
	StatusType string
	StatusText string
}

type EntryStatus string

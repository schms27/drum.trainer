package models

import "time"

// Track is the metadata for an uploaded MIDI file stored in the database.
type Track struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Filename   string    `json:"filename"`
	FilePath   string    `json:"-"` // not exposed to API clients
	FileSize   int64     `json:"file_size"`
	DurationMs float64   `json:"duration_ms"`
	NoteCount  int       `json:"note_count"`
	PPQ        int       `json:"ppq"`
	CreatedAt  time.Time `json:"created_at"`
}

// TrackWithNotes is the full track response including parsed note events.
type TrackWithNotes struct {
	Track
	Notes []NoteEvent `json:"notes"`
}

// NoteMapping associates a MIDI note number with an instrument group name.
type NoteMapping struct {
	NoteNumber int    `json:"note_number"`
	Instrument string `json:"instrument"`
}

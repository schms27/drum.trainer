package models

import "time"

// Track is the metadata for an uploaded MIDI file stored in the database.
type Track struct {
	ID               int64     `json:"id"`
	Name             string    `json:"name"`
	Filename         string    `json:"filename"`
	FilePath         string    `json:"-"` // not exposed to API clients
	FileSize         int64     `json:"file_size"`
	DurationMs       float64   `json:"duration_ms"`
	NoteCount        int       `json:"note_count"`
	PPQ              int       `json:"ppq"`
	BPM              float64   `json:"bpm"`
	TimeSigNumerator int       `json:"time_sig_numerator"` // beats per bar (e.g. 4 for 4/4, 3 for 3/4)
	PracticeBPM      *float64  `json:"practice_bpm"`       // nil = use native BPM
	CreatedAt        time.Time `json:"created_at"`
	// Stats is only filled by endpoints that report practice progress.
	Stats *TrackStats `json:"stats,omitempty"`
}

// TrackStats summarises the user's practice on one track.
type TrackStats struct {
	Sessions        int        `json:"sessions"`
	BestHitRate     *float64   `json:"best_hit_rate"`     // nil until the first take
	LastHitRate     *float64   `json:"last_hit_rate"`     // most recent take
	LastPracticedAt *time.Time `json:"last_practiced_at"` // nil until the first take
	// RecentHitRates holds the hit rates of the latest takes, oldest first
	// (at most RecentTakes entries) — enough for a small trend line.
	RecentHitRates []float64 `json:"recent_hit_rates"`
}

// RecentTakes is how many takes TrackStats.RecentHitRates covers.
const RecentTakes = 10

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

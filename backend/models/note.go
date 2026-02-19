package models

// NoteEvent represents a single parsed drum hit from a MIDI file (channel 10).
type NoteEvent struct {
	NoteNumber int     `json:"note_number"`
	TimeMs     float64 `json:"time_ms"`
	Velocity   int     `json:"velocity"`
	Instrument string  `json:"instrument"`
}

// RecordedHit is a single drum hit captured by the browser via Web MIDI API.
type RecordedHit struct {
	NoteNumber int     `json:"note_number"`
	TimeMs     float64 `json:"time_ms"`
	Velocity   int     `json:"velocity"`
}

package models

import "time"

// PilotTrack is an audio file attached to a drum track for playing along to.
type PilotTrack struct {
	ID        int64     `json:"id"`
	TrackID   int64     `json:"track_id"`
	Name      string    `json:"name"`
	Filename  string    `json:"filename"`
	FilePath  string    `json:"-"` // not exposed to API clients
	FileSize  int64     `json:"file_size"`
	MimeType  string    `json:"mime_type"`
	CreatedAt time.Time `json:"created_at"`
}

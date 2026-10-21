package models

import "time"

// KitProfile describes how the pads of a user's e-drum kit map to
// instrument groups. It applies to recorded hits.
type KitProfile struct {
	ID        int64         `json:"id"`
	Name      string        `json:"name"`
	Mappings  []NoteMapping `json:"mappings"`
	Active    bool          `json:"active"`
	// Default playback levels (0–1) for tracks without their own mix; nil = full level.
	MetronomeVolume *float64 `json:"metronome_volume"`
	PilotVolume     *float64 `json:"pilot_volume"`
	CreatedAt time.Time     `json:"created_at"`
}

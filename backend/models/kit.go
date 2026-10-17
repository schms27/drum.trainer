package models

import "time"

// KitProfile describes how the pads of a user's e-drum kit map to
// instrument groups. It applies to recorded hits.
type KitProfile struct {
	ID        int64         `json:"id"`
	Name      string        `json:"name"`
	Mappings  []NoteMapping `json:"mappings"`
	Active    bool          `json:"active"`
	CreatedAt time.Time     `json:"created_at"`
}

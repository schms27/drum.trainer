package models

// Tag is one of a user's own labels for organising their library.
type Tag struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Color      string `json:"color"`       // one of TagColors
	TrackCount int    `json:"track_count"` // tracks carrying the tag (in tag listings)
}

// TagRef is a tag as it appears on a track.
type TagRef struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// TagColors are the colours a tag can have. The frontend maps each key to a
// chip style that is readable in light and dark mode, so free colours are
// not accepted.
var TagColors = []string{"gray", "red", "orange", "amber", "green", "teal", "sky", "blue", "violet", "pink"}

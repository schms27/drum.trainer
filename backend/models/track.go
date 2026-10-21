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
	MetronomeVolume  *float64  `json:"metronome_volume"`   // 0–1; nil = use the kit profile's default
	NotesRevision    int       `json:"notes_revision"`     // bumped on every edit of the notes
	Origin           string    `json:"origin"`             // TrackOriginUpload or TrackOriginEditor
	CreatedAt        time.Time `json:"created_at"`
	// Sharing. Only the owner ever receives a Track, so the token is safe here.
	Visibility    string  `json:"visibility"`      // "private" | "unlisted"
	ShareToken    *string `json:"share_token"`     // nil while private
	RightsBasis   *string `json:"rights_basis"`    // owner's declaration when shared
	SourceTrackID *int64  `json:"source_track_id"` // set on copies of a shared track, cleared if the original is deleted
	Copied        bool    `json:"copied"`          // true for every copy, even once the original is gone
	// Disabled is set when moderation removed the track (or the track it was
	// copied from). It stays in the library, but can no longer be practised.
	Disabled       bool    `json:"disabled"`
	DisabledReason *string `json:"disabled_reason"`
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

// Origin values for Track.Origin.
const (
	TrackOriginUpload = "upload" // parsed from an uploaded MIDI file
	TrackOriginEditor = "editor" // created from scratch; has no uploaded file
)

// Visibility values for Track.Visibility.
const (
	VisibilityPrivate  = "private"
	VisibilityUnlisted = "unlisted"
	VisibilityPublic   = "public" // shared by link and listed in Explore
)

// Rights bases an owner can declare when sharing a track.
var RightsBases = []string{"original", "licensed", "public_domain", "transcription"}

// ListableRightsBases are the rights bases a track needs to be listed
// publicly. Transcriptions of other people's songs can only be shared by link.
var ListableRightsBases = []string{"original", "licensed", "public_domain"}

// PublicTrack is a track as the Explore gallery lists it.
type PublicTrack struct {
	ShareToken       string    `json:"share_token"`
	Name             string    `json:"name"`
	DurationMs       float64   `json:"duration_ms"`
	NoteCount        int       `json:"note_count"`
	BPM              float64   `json:"bpm"`
	TimeSigNumerator int       `json:"time_sig_numerator"`
	RightsBasis      string    `json:"rights_basis"`
	ListedAt         time.Time `json:"listed_at"`
	Copies           int       `json:"copies"` // how many people added it to their library
}

// SharedTrack is the view of a shared track that anyone holding the link sees.
// It deliberately carries no file names, owner identity or full note list.
type SharedTrack struct {
	Name             string         `json:"name"`
	DurationMs       float64        `json:"duration_ms"`
	NoteCount        int            `json:"note_count"`
	BPM              float64        `json:"bpm"`
	TimeSigNumerator int            `json:"time_sig_numerator"`
	RightsBasis      string         `json:"rights_basis"`
	Visibility       string         `json:"visibility"` // "unlisted" | "public"
	SharedAt         time.Time      `json:"shared_at"`
	Instruments      map[string]int `json:"instruments"`   // note count per instrument group
	PreviewNotes     []NoteEvent    `json:"preview_notes"` // the opening bars only
	PreviewMs        float64        `json:"preview_ms"`    // length the preview covers
	// Viewer-specific; only filled when the viewer is signed in.
	IsOwner       bool   `json:"is_owner"`
	CopiedTrackID *int64 `json:"copied_track_id"` // the viewer's existing copy, if any
}

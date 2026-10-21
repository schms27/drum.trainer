package models

import (
	"math"
	"time"
)

// Session stores metadata about a single play-along recording.
type Session struct {
	ID         int64     `json:"id"`
	TrackID    int64     `json:"track_id"`
	StartedAt  time.Time `json:"started_at"`
	EndedAt    time.Time `json:"ended_at"`
	DurationMs float64   `json:"duration_ms"`
	HitCount   int       `json:"hit_count"`
	CreatedAt  time.Time `json:"created_at"`
	// NotesRevision is the track's notes revision the take was played
	// against; nil for takes recorded before revisions existed.
	NotesRevision *int `json:"notes_revision"`
	SessionTempo
}

// SessionTempo is the tempo a take was played at. All fields are nil for
// takes recorded before the tempo was stored.
type SessionTempo struct {
	PracticeBPM  *float64 `json:"practice_bpm"`  // tempo the player chose
	TrackBPM     *float64 `json:"track_bpm"`     // the track's native tempo at the time
	TempoPercent *int     `json:"tempo_percent"` // practice_bpm / track_bpm, in whole percent
}

// NewSessionTempo fills TempoPercent from the two tempos. Either may be nil.
func NewSessionTempo(practiceBPM, trackBPM *float64) SessionTempo {
	t := SessionTempo{PracticeBPM: practiceBPM, TrackBPM: trackBPM}
	if practiceBPM != nil && trackBPM != nil && *trackBPM > 0 {
		// floor(x + 0.5) rather than math.Round so it matches the SQL in
		// db.ListSessions exactly (PostgreSQL's ROUND rounds halves to even).
		pct := int(math.Floor(*practiceBPM / *trackBPM * 100 + 0.5))
		t.TempoPercent = &pct
	}
	return t
}

// InstrumentBreakdown holds per-instrument scoring stats.
type InstrumentBreakdown struct {
	Instrument   string   `json:"instrument"`
	TargetCount  int      `json:"target_count"`
	MatchedCount int      `json:"matched_count"`
	MissedCount  int      `json:"missed_count"`
	GhostCount   int      `json:"ghost_count"`
	HitRate      float64  `json:"hit_rate"`
	AvgOffsetMs  *float64 `json:"avg_offset_ms"` // nil if no matches
	StdOffsetMs  *float64 `json:"std_offset_ms"` // nil if no matches
}

// GhostHit is a recorded hit that did not match any target note.
type GhostHit struct {
	RecordedTimeMs float64 `json:"recorded_time_ms"`
	NoteNumber     int     `json:"note_number"`
	Instrument     string  `json:"instrument"`
}

// MatchedHit is a target note paired with the closest recorded hit.
type MatchedHit struct {
	TargetTimeMs   float64 `json:"target_time_ms"`
	RecordedTimeMs float64 `json:"recorded_time_ms"`
	OffsetMs       float64 `json:"offset_ms"` // recorded - target; negative=early, positive=late
	NoteNumber     int     `json:"note_number"`
	Instrument     string  `json:"instrument"`
}

// ScoreResult is the full computed result for a session.
type ScoreResult struct {
	HitRate      float64               `json:"hit_rate"`
	MissRate     float64               `json:"miss_rate"`
	GhostRate    float64               `json:"ghost_rate"`
	AvgOffsetMs  float64               `json:"avg_offset_ms"`
	StdOffsetMs  float64               `json:"std_offset_ms"`
	ByInstrument []InstrumentBreakdown `json:"by_instrument"`
	MatchedHits  []MatchedHit          `json:"matched_hits"`
	GhostHits    []GhostHit            `json:"ghost_hits"`
}

// SessionSummary is a lightweight session row used in list responses.
type SessionSummary struct {
	ID            int64     `json:"id"`
	TrackID       int64     `json:"track_id"`
	HitCount      int       `json:"hit_count"`
	MatchedCount  int       `json:"matched_count"`
	TargetCount   int       `json:"target_count"`
	HitRate       float64   `json:"hit_rate"`
	AvgOffsetMs   float64   `json:"avg_offset_ms"`
	CreatedAt     time.Time `json:"created_at"`
	NotesRevision *int      `json:"notes_revision"` // see Session.NotesRevision
	SessionTempo
}

// SessionFilter narrows a session listing. Zero values mean "no filter".
type SessionFilter struct {
	TrackID int64
	// TempoMin and TempoMax bound the tempo in whole percent of the track's
	// native tempo, inclusive. Takes without a recorded tempo never match a
	// tempo bound.
	TempoMin *int
	TempoMax *int
}

// SessionWithScore is the full session + computed score, used in detail responses.
type SessionWithScore struct {
	Session Session     `json:"session"`
	Score   ScoreResult `json:"score"`
}

// SubmitSessionRequest is the payload sent by the browser after a recording.
type SubmitSessionRequest struct {
	TrackID   int64         `json:"track_id"`
	StartedAt string        `json:"started_at"` // ISO 8601 UTC
	EndedAt   string        `json:"ended_at"`   // ISO 8601 UTC
	Hits      []RecordedHit `json:"hits"`
	// PracticeBPM is the tempo the take was played at. Optional; omitted by
	// older clients, in which case the tempo is unknown.
	PracticeBPM *float64 `json:"practice_bpm"`
}

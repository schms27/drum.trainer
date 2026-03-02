package models

import "time"

// Session stores metadata about a single play-along recording.
type Session struct {
	ID         int64     `json:"id"`
	TrackID    int64     `json:"track_id"`
	StartedAt  time.Time `json:"started_at"`
	EndedAt    time.Time `json:"ended_at"`
	DurationMs float64   `json:"duration_ms"`
	HitCount   int       `json:"hit_count"`
	CreatedAt  time.Time `json:"created_at"`
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
	ID          int64     `json:"id"`
	TrackID     int64     `json:"track_id"`
	HitCount    int       `json:"hit_count"`
	HitRate     float64   `json:"hit_rate"`
	AvgOffsetMs float64   `json:"avg_offset_ms"`
	CreatedAt   time.Time `json:"created_at"`
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
}

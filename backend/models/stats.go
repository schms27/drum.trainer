package models

// WeekStat holds aggregated practice data for one ISO calendar week.
type WeekStat struct {
	WeekStart       string  `json:"week_start"`       // Monday date "2025-01-06"
	SessionCount    int     `json:"session_count"`
	MinutesPlayed   float64 `json:"minutes_played"`
	TracksPracticed int     `json:"tracks_practiced"` // distinct tracks in that week
}

// StatTotals holds aggregated totals across the full trailing year.
type StatTotals struct {
	Sessions        int     `json:"sessions"`
	MinutesPlayed   float64 `json:"minutes_played"`
	TracksPracticed int     `json:"tracks_practiced"`
}

// StatsResponse is the full payload returned by GET /api/v1/stats.
type StatsResponse struct {
	Weekly []WeekStat `json:"weekly"`
	Totals StatTotals `json:"totals"`
	// StreakDays is the number of consecutive days with at least one take,
	// ending today — or yesterday, so the streak is not lost before today's
	// practice. Days are in the requested time zone.
	StreakDays int `json:"streak_days"`
	// PracticedToday is true when there is a take today.
	PracticedToday bool `json:"practiced_today"`
}

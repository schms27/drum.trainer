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
}

package db

import (
	"database/sql"
	"time"

	"github.com/simu/drum-trainer/backend/models"
)

// GetWeeklyStats returns 52 weekly buckets (oldest → newest, ending with the
// current week) covering the trailing year for the given user.
// Weeks with no activity have zero values.
func GetWeeklyStats(db *sql.DB, userID int64) (*models.StatsResponse, error) {
	since := time.Now().UTC().AddDate(-1, 0, 0)

	rows, err := db.Query(
		Rebind(`SELECT track_id, duration_ms, created_at
		 FROM sessions WHERE user_id = ? AND created_at >= ?`),
		userID, since.Format(time.RFC3339),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type rawEntry struct {
		trackID   int64
		durationMs float64
		createdAt  time.Time
	}

	var entries []rawEntry
	for rows.Next() {
		var e rawEntry
		var ct scanTime
		if err := rows.Scan(&e.trackID, &e.durationMs, &ct); err != nil {
			return nil, err
		}
		e.createdAt = ct.T
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Aggregate by week (keyed by Monday date string).
	type weekData struct {
		count  int
		ms     float64
		tracks map[int64]struct{}
	}
	weekMap := make(map[string]*weekData)

	totalTracks := make(map[int64]struct{})
	totalMs := 0.0

	for _, e := range entries {
		key := mondayOf(e.createdAt).Format("2006-01-02")
		wd := weekMap[key]
		if wd == nil {
			wd = &weekData{tracks: make(map[int64]struct{})}
			weekMap[key] = wd
		}
		wd.count++
		wd.ms += e.durationMs
		wd.tracks[e.trackID] = struct{}{}

		totalTracks[e.trackID] = struct{}{}
		totalMs += e.durationMs
	}

	// Build 52 ordered slots ending with the current week.
	thisMonday := mondayOf(time.Now().UTC())
	weekly := make([]models.WeekStat, 52)
	for i := 0; i < 52; i++ {
		monday := thisMonday.AddDate(0, 0, -(51-i)*7)
		key := monday.Format("2006-01-02")
		stat := models.WeekStat{WeekStart: key}
		if wd, ok := weekMap[key]; ok {
			stat.SessionCount = wd.count
			stat.MinutesPlayed = wd.ms / 60_000
			stat.TracksPracticed = len(wd.tracks)
		}
		weekly[i] = stat
	}

	totals := models.StatTotals{
		Sessions:        len(entries),
		MinutesPlayed:   totalMs / 60_000,
		TracksPracticed: len(totalTracks),
	}

	return &models.StatsResponse{Weekly: weekly, Totals: totals}, nil
}

// mondayOf returns the Monday (00:00 UTC) of the week containing t.
func mondayOf(t time.Time) time.Time {
	wd := int(t.Weekday())
	if wd == 0 {
		wd = 7 // Sunday → 7 so Monday is always offset 1
	}
	return time.Date(t.Year(), t.Month(), t.Day()-wd+1, 0, 0, 0, 0, time.UTC)
}

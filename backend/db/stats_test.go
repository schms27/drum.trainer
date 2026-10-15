package db

import (
	"testing"
	"time"
)

func TestWeeklyStatsUsesUserTimeZone(t *testing.T) {
	db := openTestDB(t)

	user, err := UpsertUser(db, "github", "1", "a@example.com", "A", "")
	if err != nil {
		t.Fatal(err)
	}
	var trackID int64
	if err := db.QueryRow(
		`INSERT INTO tracks (name, filename, file_path, file_size, duration_ms, note_count, ppq, user_id)
		 VALUES ('t', 't.mid', 't.mid', 1, 1, 1, 480, ?) RETURNING id`, user.ID,
	).Scan(&trackID); err != nil {
		t.Fatal(err)
	}
	// Sunday 2026-10-11 23:30 UTC is already Monday 2026-10-12 01:30 in Zurich.
	played := time.Date(2026, 10, 11, 23, 30, 0, 0, time.UTC).Format("2006-01-02 15:04:05")
	if _, err := db.Exec(
		`INSERT INTO sessions (track_id, started_at, ended_at, duration_ms, hit_count, user_id, created_at)
		 VALUES (?, ?, ?, 60000, 1, ?, ?)`, trackID, played, played, user.ID, played,
	); err != nil {
		t.Fatal(err)
	}

	zurich, err := time.LoadLocation("Europe/Zurich")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 10, 14, 12, 0, 0, 0, zurich)

	weekOf := func(loc *time.Location) string {
		resp, err := weeklyStats(db, user.ID, now.In(loc))
		if err != nil {
			t.Fatal(err)
		}
		for _, w := range resp.Weekly {
			if w.SessionCount > 0 {
				return w.WeekStart
			}
		}
		t.Fatal("session not counted in any week")
		return ""
	}

	if got := weekOf(time.UTC); got != "2026-10-05" {
		t.Errorf("UTC: week = %s, want 2026-10-05", got)
	}
	if got := weekOf(zurich); got != "2026-10-12" {
		t.Errorf("Zurich: week = %s, want 2026-10-12", got)
	}
}

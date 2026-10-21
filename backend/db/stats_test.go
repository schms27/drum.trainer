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

	// The same take lands on Sunday in UTC but Monday in Zurich.
	for loc, want := range map[*time.Location]string{time.UTC: "2026-10-11", zurich: "2026-10-12"} {
		resp, err := weeklyStats(db, user.ID, now.In(loc))
		if err != nil {
			t.Fatal(err)
		}
		if len(resp.Daily) != 1 || resp.Daily[0].Date != want || resp.Daily[0].SessionCount != 1 || resp.Daily[0].MinutesPlayed != 1 {
			t.Errorf("%s: daily = %+v, want one minute on %s", loc, resp.Daily, want)
		}
	}
}

func TestPracticeStreak(t *testing.T) {
	now := time.Date(2026, 10, 16, 20, 0, 0, 0, time.UTC)
	set := func(ds ...string) map[string]bool {
		m := map[string]bool{}
		for _, d := range ds {
			m[d] = true
		}
		return m
	}
	cases := []struct {
		name   string
		days   map[string]bool
		streak int
		today  bool
	}{
		{"none", set(), 0, false},
		{"today only", set("2026-10-16"), 1, true},
		{"three days to today", set("2026-10-14", "2026-10-15", "2026-10-16"), 3, true},
		{"ends yesterday", set("2026-10-14", "2026-10-15"), 2, false},
		{"gap breaks streak", set("2026-10-13", "2026-10-15", "2026-10-16"), 2, true},
		{"last take two days ago", set("2026-10-14"), 0, false},
	}
	for _, c := range cases {
		streak, today := practiceStreak(c.days, now)
		if streak != c.streak || today != c.today {
			t.Errorf("%s: got (%d, %v), want (%d, %v)", c.name, streak, today, c.streak, c.today)
		}
	}
}

package db

import "testing"

func TestListTrackStats(t *testing.T) {
	db := openTestDB(t)

	user, err := UpsertUser(db, "github", "1", "a@example.com", "A", "")
	if err != nil {
		t.Fatal(err)
	}
	newTrack := func() int64 {
		var id int64
		if err := db.QueryRow(
			`INSERT INTO tracks (name, filename, file_path, file_size, duration_ms, note_count, ppq, user_id)
			 VALUES ('t', 't.mid', 't.mid', 1, 1, 1, 480, ?) RETURNING id`, user.ID,
		).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}
	addTake := func(trackID int64, hitRate float64) {
		var sid int64
		if err := db.QueryRow(
			`INSERT INTO sessions (track_id, started_at, ended_at, duration_ms, hit_count, user_id)
			 VALUES (?, '2026-10-16T10:00:00Z', '2026-10-16T10:01:00Z', 60000, 1, ?) RETURNING id`, trackID, user.ID,
		).Scan(&sid); err != nil {
			t.Fatal(err)
		}
		if _, err := db.Exec(
			`INSERT INTO session_scores (session_id, hit_rate, miss_rate, ghost_rate, avg_offset_ms, std_offset_ms, score_json)
			 VALUES (?, ?, 0, 0, 0, 0, '{}')`, sid, hitRate,
		); err != nil {
			t.Fatal(err)
		}
	}

	practiced, untouched := newTrack(), newTrack()
	for i := 0; i < 12; i++ {
		addTake(practiced, 0.5+float64(i)*0.01)
	}
	addTake(practiced, 0.4) // latest take is worse than the best

	stats, err := ListTrackStats(db, user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := stats[untouched]; ok {
		t.Error("track without takes should have no stats entry")
	}
	st := stats[practiced]
	if st == nil {
		t.Fatal("missing stats for practised track")
	}
	if st.Sessions != 13 {
		t.Errorf("sessions = %d, want 13", st.Sessions)
	}
	if *st.BestHitRate < 0.6099 || *st.BestHitRate > 0.6101 {
		t.Errorf("best = %v, want 0.61", *st.BestHitRate)
	}
	if *st.LastHitRate != 0.4 {
		t.Errorf("last = %v, want 0.4", *st.LastHitRate)
	}
	if len(st.RecentHitRates) != 10 || st.RecentHitRates[9] != 0.4 {
		t.Errorf("recent = %v, want 10 entries ending with 0.4", st.RecentHitRates)
	}
	if st.LastPracticedAt == nil {
		t.Error("last practiced time missing")
	}
}

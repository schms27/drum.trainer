package db

import (
	"github.com/simu/drum-trainer/backend/models"
)

// ListTrackStats returns practice statistics for every track of a user that
// has at least one take, keyed by track ID.
func ListTrackStats(db *DB, userID int64) (map[int64]*models.TrackStats, error) {
	rows, err := db.Query(
		db.Rebind(`SELECT s.track_id, ss.hit_rate, s.created_at
		 FROM sessions s
		 JOIN session_scores ss ON ss.session_id = s.id
		 WHERE s.user_id = ?
		 ORDER BY s.id ASC`),
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats := make(map[int64]*models.TrackStats)
	for rows.Next() {
		var trackID int64
		var hitRate float64
		var ct scanTime
		if err := rows.Scan(&trackID, &hitRate, &ct); err != nil {
			return nil, err
		}
		st := stats[trackID]
		if st == nil {
			st = &models.TrackStats{RecentHitRates: []float64{}}
			stats[trackID] = st
		}
		st.Sessions++
		if st.BestHitRate == nil || hitRate > *st.BestHitRate {
			best := hitRate
			st.BestHitRate = &best
		}
		last, at := hitRate, ct.T
		st.LastHitRate = &last
		st.LastPracticedAt = &at
		st.RecentHitRates = append(st.RecentHitRates, hitRate)
		if len(st.RecentHitRates) > models.RecentTakes {
			st.RecentHitRates = st.RecentHitRates[1:]
		}
	}
	return stats, rows.Err()
}

// EmptyTrackStats is what a track without any takes reports.
func EmptyTrackStats() *models.TrackStats {
	return &models.TrackStats{RecentHitRates: []float64{}}
}

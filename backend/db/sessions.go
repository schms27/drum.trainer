package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/simu/drum-trainer/backend/models"
)

// InsertSession persists a session, its raw hits, and the computed score in one transaction.
func InsertSession(db *sql.DB, session *models.Session, hits []models.RecordedHit, score *models.ScoreResult, instrumentGroups map[string]string, userID int64) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var sessionID int64
	err = tx.QueryRow(
		Rebind(`INSERT INTO sessions (track_id, started_at, ended_at, duration_ms, hit_count, user_id)
		 VALUES (?, ?, ?, ?, ?, ?) RETURNING id`),
		session.TrackID,
		session.StartedAt.UTC().Format(time.RFC3339),
		session.EndedAt.UTC().Format(time.RFC3339),
		session.DurationMs,
		session.HitCount,
		userID,
	).Scan(&sessionID)
	if err != nil {
		return 0, fmt.Errorf("insert session: %w", err)
	}

	// Insert raw hits.
	hitStmt, err := tx.Prepare(
		Rebind(`INSERT INTO session_hits (session_id, note_number, time_ms, velocity, instrument)
		 VALUES (?, ?, ?, ?, ?)`),
	)
	if err != nil {
		return 0, err
	}
	defer hitStmt.Close()

	for _, h := range hits {
		instr := instrumentGroups[fmt.Sprintf("%d", h.NoteNumber)]
		if instr == "" {
			instr = "Other"
		}
		if _, err := hitStmt.Exec(sessionID, h.NoteNumber, h.TimeMs, h.Velocity, instr); err != nil {
			return 0, fmt.Errorf("insert hit: %w", err)
		}
	}

	// Serialize score to JSON blob.
	scoreJSON, err := json.Marshal(score)
	if err != nil {
		return 0, err
	}

	_, err = tx.Exec(
		Rebind(`INSERT INTO session_scores (session_id, hit_rate, miss_rate, ghost_rate, avg_offset_ms, std_offset_ms, score_json)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`),
		sessionID, score.HitRate, score.MissRate, score.GhostRate,
		score.AvgOffsetMs, score.StdOffsetMs, string(scoreJSON),
	)
	if err != nil {
		return 0, fmt.Errorf("insert score: %w", err)
	}

	// Insert per-instrument breakdown rows.
	instrStmt, err := tx.Prepare(
		Rebind(`INSERT INTO session_instrument_scores
		 (session_id, instrument, target_count, matched_count, missed_count, ghost_count, hit_rate, avg_offset_ms, std_offset_ms)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`),
	)
	if err != nil {
		return 0, err
	}
	defer instrStmt.Close()

	for _, bd := range score.ByInstrument {
		if _, err := instrStmt.Exec(
			sessionID, bd.Instrument,
			bd.TargetCount, bd.MatchedCount, bd.MissedCount, bd.GhostCount,
			bd.HitRate, bd.AvgOffsetMs, bd.StdOffsetMs,
		); err != nil {
			return 0, fmt.Errorf("insert instrument score: %w", err)
		}
	}

	return sessionID, tx.Commit()
}

// GetSession returns a session's metadata by ID scoped to the given user.
func GetSession(db *sql.DB, id, userID int64) (*models.Session, error) {
	row := db.QueryRow(
		Rebind(`SELECT id, track_id, started_at, ended_at, duration_ms, hit_count, created_at
		 FROM sessions WHERE id = ? AND user_id = ?`), id, userID,
	)
	return scanSession(row)
}

// GetSessionScore returns the ScoreResult JSON for a session.
func GetSessionScore(db *sql.DB, sessionID int64) (*models.ScoreResult, error) {
	var blob string
	err := db.QueryRow(
		Rebind(`SELECT score_json FROM session_scores WHERE session_id = ?`), sessionID,
	).Scan(&blob)
	if err != nil {
		return nil, err
	}
	var score models.ScoreResult
	if err := json.Unmarshal([]byte(blob), &score); err != nil {
		return nil, err
	}
	return &score, nil
}

// ListSessions returns a paginated list of sessions for a user, optionally filtered by track.
// limit controls the page size; before is a cursor (session ID) — only sessions with id < before
// are returned. Pass before=0 to start from the most recent session.
// Returns (sessions, hasMore, error). hasMore is true when there are additional rows beyond limit.
func ListSessions(db *sql.DB, userID, trackID int64, limit int, before int64) ([]models.SessionSummary, bool, error) {
	query := `
		SELECT s.id, s.track_id, s.hit_count,
		       COALESCE(SUM(si.matched_count), 0), COALESCE(SUM(si.target_count), 0),
		       ss.hit_rate, ss.avg_offset_ms, s.created_at
		FROM sessions s
		JOIN session_scores ss ON ss.session_id = s.id
		LEFT JOIN session_instrument_scores si ON si.session_id = s.id
		WHERE s.user_id = ?`
	args := []any{userID}

	if trackID > 0 {
		query += ` AND s.track_id = ?`
		args = append(args, trackID)
	}
	if before > 0 {
		query += ` AND s.id < ?`
		args = append(args, before)
	}
	query += ` GROUP BY s.id, s.track_id, s.hit_count, ss.hit_rate, ss.avg_offset_ms, s.created_at`
	query += ` ORDER BY s.id DESC`
	// Fetch one extra row to determine whether more pages exist.
	query += ` LIMIT ?`
	args = append(args, limit+1)

	rows, err := db.Query(Rebind(query), args...)
	if err != nil {
		return nil, false, err
	}
	defer rows.Close()

	var list []models.SessionSummary
	for rows.Next() {
		var s models.SessionSummary
		var ct scanTime
		if err := rows.Scan(&s.ID, &s.TrackID, &s.HitCount, &s.MatchedCount, &s.TargetCount, &s.HitRate, &s.AvgOffsetMs, &ct); err != nil {
			return nil, false, err
		}
		s.CreatedAt = ct.T
		list = append(list, s)
	}
	if err := rows.Err(); err != nil {
		return nil, false, err
	}

	hasMore := len(list) > limit
	if hasMore {
		list = list[:limit]
	}
	return list, hasMore, nil
}

func scanSession(s scanner) (*models.Session, error) {
	var sess models.Session
	var startedAt, endedAt scanTime
	var ct scanTime
	err := s.Scan(&sess.ID, &sess.TrackID, &startedAt, &endedAt, &sess.DurationMs, &sess.HitCount, &ct)
	if err != nil {
		return nil, err
	}
	sess.StartedAt = startedAt.T
	sess.EndedAt = endedAt.T
	sess.CreatedAt = ct.T
	return &sess, nil
}

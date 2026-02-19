package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/simu/drum-trainer/backend/models"
)

// InsertSession persists a session, its raw hits, and the computed score in one transaction.
func InsertSession(db *sql.DB, session *models.Session, hits []models.RecordedHit, score *models.ScoreResult, instrumentGroups map[string]string) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		`INSERT INTO sessions (track_id, started_at, ended_at, duration_ms, hit_count)
		 VALUES (?, ?, ?, ?, ?)`,
		session.TrackID,
		session.StartedAt.UTC().Format(time.RFC3339),
		session.EndedAt.UTC().Format(time.RFC3339),
		session.DurationMs,
		session.HitCount,
	)
	if err != nil {
		return 0, fmt.Errorf("insert session: %w", err)
	}

	sessionID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Insert raw hits.
	hitStmt, err := tx.Prepare(
		`INSERT INTO session_hits (session_id, note_number, time_ms, velocity, instrument)
		 VALUES (?, ?, ?, ?, ?)`,
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
		`INSERT INTO session_scores (session_id, hit_rate, miss_rate, ghost_rate, avg_offset_ms, std_offset_ms, score_json)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		sessionID, score.HitRate, score.MissRate, score.GhostRate,
		score.AvgOffsetMs, score.StdOffsetMs, string(scoreJSON),
	)
	if err != nil {
		return 0, fmt.Errorf("insert score: %w", err)
	}

	// Insert per-instrument breakdown rows.
	instrStmt, err := tx.Prepare(
		`INSERT INTO session_instrument_scores
		 (session_id, instrument, target_count, matched_count, missed_count, ghost_count, hit_rate, avg_offset_ms, std_offset_ms)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
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

// GetSession returns a session's metadata by ID.
func GetSession(db *sql.DB, id int64) (*models.Session, error) {
	row := db.QueryRow(
		`SELECT id, track_id, started_at, ended_at, duration_ms, hit_count, created_at
		 FROM sessions WHERE id = ?`, id,
	)
	return scanSession(row)
}

// GetSessionScore returns the ScoreResult JSON for a session.
func GetSessionScore(db *sql.DB, sessionID int64) (*models.ScoreResult, error) {
	var blob string
	err := db.QueryRow(
		`SELECT score_json FROM session_scores WHERE session_id = ?`, sessionID,
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

// ListSessions returns a lightweight list of sessions, optionally filtered by track.
func ListSessions(db *sql.DB, trackID int64) ([]models.SessionSummary, error) {
	query := `
		SELECT s.id, s.track_id, s.hit_count, ss.hit_rate, ss.avg_offset_ms, s.created_at
		FROM sessions s
		JOIN session_scores ss ON ss.session_id = s.id`
	args := []any{}

	if trackID > 0 {
		query += ` WHERE s.track_id = ?`
		args = append(args, trackID)
	}
	query += ` ORDER BY s.id DESC`

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.SessionSummary
	for rows.Next() {
		var s models.SessionSummary
		var createdAt string
		if err := rows.Scan(&s.ID, &s.TrackID, &s.HitCount, &s.HitRate, &s.AvgOffsetMs, &createdAt); err != nil {
			return nil, err
		}
		s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		list = append(list, s)
	}
	return list, rows.Err()
}

func scanSession(s scanner) (*models.Session, error) {
	var sess models.Session
	var startedAt, endedAt, createdAt string
	err := s.Scan(&sess.ID, &sess.TrackID, &startedAt, &endedAt, &sess.DurationMs, &sess.HitCount, &createdAt)
	if err != nil {
		return nil, err
	}
	sess.StartedAt, _ = time.Parse(time.RFC3339, startedAt)
	sess.EndedAt, _ = time.Parse(time.RFC3339, endedAt)
	sess.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	return &sess, nil
}

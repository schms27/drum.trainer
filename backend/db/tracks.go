package db

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/simu/drum-trainer/backend/models"
)

// InsertTrack inserts a track row and all its notes inside a single transaction.
// It returns the new track ID.
func InsertTrack(db *sql.DB, track *models.Track, notes []models.NoteEvent) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(
		`INSERT INTO tracks (name, filename, file_path, file_size, duration_ms, note_count, ppq)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		track.Name, track.Filename, track.FilePath,
		track.FileSize, track.DurationMs, track.NoteCount, track.PPQ,
	)
	if err != nil {
		return 0, fmt.Errorf("insert track: %w", err)
	}

	trackID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Bulk-insert notes using a prepared statement for performance.
	stmt, err := tx.Prepare(
		`INSERT INTO track_notes (track_id, note_number, time_ms, velocity, instrument)
		 VALUES (?, ?, ?, ?, ?)`,
	)
	if err != nil {
		return 0, fmt.Errorf("prepare note insert: %w", err)
	}
	defer stmt.Close()

	for _, n := range notes {
		if _, err := stmt.Exec(trackID, n.NoteNumber, n.TimeMs, n.Velocity, n.Instrument); err != nil {
			return 0, fmt.Errorf("insert note: %w", err)
		}
	}

	return trackID, tx.Commit()
}

// GetTrack returns a track by ID (without notes).
func GetTrack(db *sql.DB, id int64) (*models.Track, error) {
	row := db.QueryRow(
		`SELECT id, name, filename, file_path, file_size, duration_ms, note_count, ppq, created_at
		 FROM tracks WHERE id = ?`, id,
	)
	return scanTrack(row)
}

// GetTrackNotes returns all parsed notes for a track, sorted by time.
func GetTrackNotes(db *sql.DB, trackID int64) ([]models.NoteEvent, error) {
	rows, err := db.Query(
		`SELECT note_number, time_ms, velocity, instrument
		 FROM track_notes WHERE track_id = ? ORDER BY time_ms ASC`, trackID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []models.NoteEvent
	for rows.Next() {
		var n models.NoteEvent
		if err := rows.Scan(&n.NoteNumber, &n.TimeMs, &n.Velocity, &n.Instrument); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

// ListTracks returns all tracks without notes, newest first.
func ListTracks(db *sql.DB) ([]models.Track, error) {
	rows, err := db.Query(
		`SELECT id, name, filename, file_path, file_size, duration_ms, note_count, ppq, created_at
		 FROM tracks ORDER BY id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tracks []models.Track
	for rows.Next() {
		t, err := scanTrack(rows)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, *t)
	}
	return tracks, rows.Err()
}

// DeleteTrack removes a track and all related data (sessions/notes cascade).
func DeleteTrack(db *sql.DB, id int64) error {
	res, err := db.Exec(`DELETE FROM tracks WHERE id = ?`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// scanner is satisfied by both *sql.Row and *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func scanTrack(s scanner) (*models.Track, error) {
	var t models.Track
	var createdAt string
	err := s.Scan(
		&t.ID, &t.Name, &t.Filename, &t.FilePath,
		&t.FileSize, &t.DurationMs, &t.NoteCount, &t.PPQ, &createdAt,
	)
	if err != nil {
		return nil, err
	}
	t.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
	return &t, nil
}

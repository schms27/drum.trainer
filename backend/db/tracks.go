package db

import (
	"database/sql"
	"fmt"

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

	var trackID int64
	err = tx.QueryRow(
		Rebind(`INSERT INTO tracks (name, filename, file_path, file_size, duration_ms, note_count, ppq, bpm)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?) RETURNING id`),
		track.Name, track.Filename, track.FilePath,
		track.FileSize, track.DurationMs, track.NoteCount, track.PPQ, track.BPM,
	).Scan(&trackID)
	if err != nil {
		return 0, fmt.Errorf("insert track: %w", err)
	}

	// Bulk-insert notes using a prepared statement for performance.
	stmt, err := tx.Prepare(
		Rebind(`INSERT INTO track_notes (track_id, note_number, time_ms, velocity, instrument)
		 VALUES (?, ?, ?, ?, ?)`),
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
		Rebind(`SELECT id, name, filename, file_path, file_size, duration_ms, note_count, ppq, bpm, practice_bpm, created_at
		 FROM tracks WHERE id = ?`), id,
	)
	return scanTrack(row)
}

// GetTrackNotes returns all parsed notes for a track, sorted by time.
func GetTrackNotes(db *sql.DB, trackID int64) ([]models.NoteEvent, error) {
	rows, err := db.Query(
		Rebind(`SELECT note_number, time_ms, velocity, instrument
		 FROM track_notes WHERE track_id = ? ORDER BY time_ms ASC`), trackID,
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
		`SELECT id, name, filename, file_path, file_size, duration_ms, note_count, ppq, bpm, practice_bpm, created_at
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
	res, err := db.Exec(Rebind(`DELETE FROM tracks WHERE id = ?`), id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// GetTrackNoteNumbers returns the distinct MIDI note numbers used in a track.
func GetTrackNoteNumbers(db *sql.DB, trackID int64) ([]int, error) {
	rows, err := db.Query(
		Rebind(`SELECT DISTINCT note_number FROM track_notes WHERE track_id = ? ORDER BY note_number ASC`),
		trackID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []int
	for rows.Next() {
		var n int
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}
	return notes, rows.Err()
}

// GetTrackMapping returns the custom note-to-instrument overrides for a track.
// The map is empty if no overrides have been set.
func GetTrackMapping(db *sql.DB, trackID int64) (map[int]string, error) {
	rows, err := db.Query(
		Rebind(`SELECT note_number, instrument FROM track_note_mappings WHERE track_id = ?`),
		trackID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	m := make(map[int]string)
	for rows.Next() {
		var note int
		var instr string
		if err := rows.Scan(&note, &instr); err != nil {
			return nil, err
		}
		m[note] = instr
	}
	return m, rows.Err()
}

// SetTrackMapping replaces all custom note mappings for a track in one transaction.
func SetTrackMapping(db *sql.DB, trackID int64, mappings []models.NoteMapping) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(Rebind(`DELETE FROM track_note_mappings WHERE track_id = ?`), trackID); err != nil {
		return fmt.Errorf("delete old mappings: %w", err)
	}

	stmt, err := tx.Prepare(
		Rebind(`INSERT INTO track_note_mappings (track_id, note_number, instrument) VALUES (?, ?, ?)`),
	)
	if err != nil {
		return fmt.Errorf("prepare mapping insert: %w", err)
	}
	defer stmt.Close()

	for _, m := range mappings {
		if _, err := stmt.Exec(trackID, m.NoteNumber, m.Instrument); err != nil {
			return fmt.Errorf("insert mapping: %w", err)
		}
	}
	return tx.Commit()
}

// scanner is satisfied by both *sql.Row and *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

// SetPracticeBPM persists the user's preferred practice tempo for a track.
// Pass nil to revert to the track's native BPM.
func SetPracticeBPM(db *sql.DB, trackID int64, bpm *float64) error {
	res, err := db.Exec(Rebind(`UPDATE tracks SET practice_bpm = ? WHERE id = ?`), bpm, trackID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func scanTrack(s scanner) (*models.Track, error) {
	var t models.Track
	var practiceBPM sql.NullFloat64
	var ct scanTime
	err := s.Scan(
		&t.ID, &t.Name, &t.Filename, &t.FilePath,
		&t.FileSize, &t.DurationMs, &t.NoteCount, &t.PPQ, &t.BPM, &practiceBPM, &ct,
	)
	if err != nil {
		return nil, err
	}
	if practiceBPM.Valid {
		t.PracticeBPM = &practiceBPM.Float64
	}
	t.CreatedAt = ct.T
	return &t, nil
}

package db

import (
	"database/sql"
	"fmt"

	"github.com/simu/drum-trainer/backend/models"
)

// InsertTrack inserts a track row and all its notes inside a single transaction,
// setting the track's ID and creation time.
func InsertTrack(db *DB, track *models.Track, notes []models.NoteEvent, userID int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var trackID int64
	var ct scanTime
	err = tx.QueryRow(
		db.Rebind(`INSERT INTO tracks (name, filename, file_path, file_size, duration_ms, note_count, ppq, bpm, time_sig_numerator, user_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?) RETURNING id, created_at`),
		track.Name, track.Filename, track.FilePath,
		track.FileSize, track.DurationMs, track.NoteCount, track.PPQ, track.BPM, track.TimeSigNumerator, userID,
	).Scan(&trackID, &ct)
	if err != nil {
		return fmt.Errorf("insert track: %w", err)
	}

	// Bulk-insert notes using a prepared statement for performance.
	stmt, err := tx.Prepare(
		db.Rebind(`INSERT INTO track_notes (track_id, note_number, time_ms, velocity, instrument)
		 VALUES (?, ?, ?, ?, ?)`),
	)
	if err != nil {
		return fmt.Errorf("prepare note insert: %w", err)
	}
	defer stmt.Close()

	for _, n := range notes {
		if _, err := stmt.Exec(trackID, n.NoteNumber, n.TimeMs, n.Velocity, n.Instrument); err != nil {
			return fmt.Errorf("insert note: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	track.ID = trackID
	track.CreatedAt = ct.T
	track.Visibility = models.VisibilityPrivate
	return nil
}

// GetTrack returns a track by ID scoped to the given user.
func GetTrack(db *DB, id, userID int64) (*models.Track, error) {
	row := db.QueryRow(
		db.Rebind(`SELECT `+trackColumns+`
		 FROM tracks WHERE id = ? AND user_id = ?`), id, userID,
	)
	return scanTrack(row)
}

// GetTrackNotes returns all parsed notes for a track, sorted by time.
func GetTrackNotes(db *DB, trackID int64) ([]models.NoteEvent, error) {
	rows, err := db.Query(
		db.Rebind(`SELECT note_number, time_ms, velocity, instrument
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

// ListTracks returns all tracks for a user without notes, newest first.
func ListTracks(db *DB, userID int64) ([]models.Track, error) {
	rows, err := db.Query(
		db.Rebind(`SELECT `+trackColumns+`
		 FROM tracks WHERE user_id = ? ORDER BY id DESC`),
		userID,
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
func DeleteTrack(db *DB, id, userID int64) error {
	res, err := db.Exec(db.Rebind(`DELETE FROM tracks WHERE id = ? AND user_id = ?`), id, userID)
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
func GetTrackNoteNumbers(db *DB, trackID int64) ([]int, error) {
	rows, err := db.Query(
		db.Rebind(`SELECT DISTINCT note_number FROM track_notes WHERE track_id = ? ORDER BY note_number ASC`),
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
func GetTrackMapping(db *DB, trackID int64) (map[int]string, error) {
	rows, err := db.Query(
		db.Rebind(`SELECT note_number, instrument FROM track_note_mappings WHERE track_id = ?`),
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
func SetTrackMapping(db *DB, trackID int64, mappings []models.NoteMapping) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(db.Rebind(`DELETE FROM track_note_mappings WHERE track_id = ?`), trackID); err != nil {
		return fmt.Errorf("delete old mappings: %w", err)
	}

	stmt, err := tx.Prepare(
		db.Rebind(`INSERT INTO track_note_mappings (track_id, note_number, instrument) VALUES (?, ?, ?)`),
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
func SetPracticeBPM(db *DB, trackID, userID int64, bpm *float64) error {
	res, err := db.Exec(db.Rebind(`UPDATE tracks SET practice_bpm = ? WHERE id = ? AND user_id = ?`), bpm, trackID, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// trackColumns is the column list scanTrack expects, in order.
const trackColumns = `id, name, filename, file_path, file_size, duration_ms, note_count, ppq, bpm, time_sig_numerator, practice_bpm, created_at,
	visibility, share_token, rights_basis, source_track_id, origin_track_id IS NOT NULL, disabled_at IS NOT NULL, disabled_reason`

func scanTrack(s scanner) (*models.Track, error) {
	var t models.Track
	var practiceBPM sql.NullFloat64
	var ct scanTime
	var shareToken, rightsBasis, disabledReason sql.NullString
	var sourceID sql.NullInt64
	err := s.Scan(
		&t.ID, &t.Name, &t.Filename, &t.FilePath,
		&t.FileSize, &t.DurationMs, &t.NoteCount, &t.PPQ, &t.BPM, &t.TimeSigNumerator, &practiceBPM, &ct,
		&t.Visibility, &shareToken, &rightsBasis, &sourceID, &t.Copied, &t.Disabled, &disabledReason,
	)
	if err != nil {
		return nil, err
	}
	if practiceBPM.Valid {
		t.PracticeBPM = &practiceBPM.Float64
	}
	if shareToken.Valid {
		t.ShareToken = &shareToken.String
	}
	if rightsBasis.Valid {
		t.RightsBasis = &rightsBasis.String
	}
	if sourceID.Valid {
		t.SourceTrackID = &sourceID.Int64
	}
	if disabledReason.Valid {
		t.DisabledReason = &disabledReason.String
	}
	t.CreatedAt = ct.T
	return &t, nil
}

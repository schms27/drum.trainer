package db

import (
	"database/sql"
	"fmt"

	"github.com/simu/drum-trainer/backend/models"
)

// InsertPilotTrack persists a new pilot track row, setting its ID and creation time.
func InsertPilotTrack(db *DB, pilot *models.PilotTrack, userID int64) error {
	var ct scanTime
	err := db.QueryRow(
		db.Rebind(`INSERT INTO pilot_tracks (track_id, user_id, name, filename, file_path, file_size, mime_type)
		 VALUES (?, ?, ?, ?, ?, ?, ?) RETURNING id, created_at`),
		pilot.TrackID, userID, pilot.Name, pilot.Filename, pilot.FilePath, pilot.FileSize, pilot.MimeType,
	).Scan(&pilot.ID, &ct)
	if err != nil {
		return fmt.Errorf("insert pilot track: %w", err)
	}
	pilot.CreatedAt = ct.T
	return nil
}

// ListPilotTracks returns all pilot tracks for a drum track owned by the user.
func ListPilotTracks(db *DB, trackID, userID int64) ([]models.PilotTrack, error) {
	rows, err := db.Query(
		db.Rebind(`SELECT id, track_id, name, filename, file_path, file_size, mime_type, created_at
		 FROM pilot_tracks WHERE track_id = ? AND user_id = ? ORDER BY id ASC`),
		trackID, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pilots []models.PilotTrack
	for rows.Next() {
		p, err := scanPilot(rows)
		if err != nil {
			return nil, err
		}
		pilots = append(pilots, *p)
	}
	return pilots, rows.Err()
}

// GetPilotTrack returns a single pilot track, scoped to the owning user.
func GetPilotTrack(db *DB, id, trackID, userID int64) (*models.PilotTrack, error) {
	row := db.QueryRow(
		db.Rebind(`SELECT id, track_id, name, filename, file_path, file_size, mime_type, created_at
		 FROM pilot_tracks WHERE id = ? AND track_id = ? AND user_id = ?`),
		id, trackID, userID,
	)
	return scanPilot(row)
}

// ListPilotFilePaths returns only the file_path values for all pilot tracks
// belonging to a track. Used to clean up files on disk before a track is deleted.
func ListPilotFilePaths(db *DB, trackID int64) ([]string, error) {
	rows, err := db.Query(
		db.Rebind(`SELECT file_path FROM pilot_tracks WHERE track_id = ?`),
		trackID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return nil, err
		}
		paths = append(paths, p)
	}
	return paths, rows.Err()
}

// DeletePilotTrack removes the pilot track row. Returns sql.ErrNoRows if not found
// or not owned by the user.
func DeletePilotTrack(db *DB, id, trackID, userID int64) error {
	res, err := db.Exec(
		db.Rebind(`DELETE FROM pilot_tracks WHERE id = ? AND track_id = ? AND user_id = ?`),
		id, trackID, userID,
	)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func scanPilot(s scanner) (*models.PilotTrack, error) {
	var p models.PilotTrack
	var ct scanTime
	err := s.Scan(
		&p.ID, &p.TrackID, &p.Name, &p.Filename,
		&p.FilePath, &p.FileSize, &p.MimeType, &ct,
	)
	if err != nil {
		return nil, err
	}
	p.CreatedAt = ct.T
	return &p, nil
}

package db

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/simu/drum-trainer/backend/models"
)

// ErrCopyNotShareable is returned when the owner tries to share a track that is
// itself a copy of someone else's shared track. Only the original uploader can
// share, so there is exactly one accountable person per shared track.
var ErrCopyNotShareable = errors.New("copies of shared tracks cannot be shared")

// ErrNotShared is returned when rotating the link of a track that is private.
var ErrNotShared = errors.New("track is not shared")

// ShareEvent records who changed a track's sharing, and what they attested to.
type ShareEvent struct {
	Action      string // "share" | "rotate" | "unshare"
	RightsBasis string
	TermsVer    string
	IPHash      string
}

// newShareToken returns a URL-safe token with 128 bits of entropy.
func newShareToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(fmt.Sprintf("crypto/rand failed: %v", err))
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// ShareTrack makes a track reachable through an unlisted link and returns the
// link token. An already-shared track keeps its token; only the rights
// declaration is updated.
func ShareTrack(db *DB, trackID, userID int64, ev ShareEvent) (string, error) {
	tx, err := db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var token sql.NullString
	var sourceID sql.NullInt64
	err = tx.QueryRow(
		db.Rebind(`SELECT share_token, source_track_id FROM tracks WHERE id = ? AND user_id = ?`), trackID, userID,
	).Scan(&token, &sourceID)
	if err != nil {
		return "", err
	}
	if sourceID.Valid {
		return "", ErrCopyNotShareable
	}

	t := token.String
	if !token.Valid {
		t = newShareToken()
	}
	_, err = tx.Exec(db.Rebind(`UPDATE tracks
		SET visibility = ?, share_token = ?, rights_basis = ?, rights_terms_ver = ?,
		    shared_at = COALESCE(shared_at, CURRENT_TIMESTAMP)
		WHERE id = ?`),
		models.VisibilityUnlisted, t, ev.RightsBasis, ev.TermsVer, trackID,
	)
	if err != nil {
		return "", fmt.Errorf("share track: %w", err)
	}
	ev.Action = "share"
	if err := insertShareEvent(db, tx, trackID, userID, models.VisibilityUnlisted, ev); err != nil {
		return "", err
	}
	return t, tx.Commit()
}

// RotateShareToken replaces a shared track's link token, so the old link stops
// working. Existing copies are not affected.
func RotateShareToken(db *DB, trackID, userID int64, ev ShareEvent) (string, error) {
	tx, err := db.Begin()
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	var visibility string
	var rights sql.NullString
	err = tx.QueryRow(
		db.Rebind(`SELECT visibility, rights_basis FROM tracks WHERE id = ? AND user_id = ?`), trackID, userID,
	).Scan(&visibility, &rights)
	if err != nil {
		return "", err
	}
	if visibility == models.VisibilityPrivate {
		return "", ErrNotShared
	}

	t := newShareToken()
	if _, err := tx.Exec(db.Rebind(`UPDATE tracks SET share_token = ? WHERE id = ?`), t, trackID); err != nil {
		return "", fmt.Errorf("rotate token: %w", err)
	}
	ev.Action = "rotate"
	ev.RightsBasis = rights.String
	if err := insertShareEvent(db, tx, trackID, userID, visibility, ev); err != nil {
		return "", err
	}
	return t, tx.Commit()
}

// UnshareTrack makes a track private again and invalidates its link.
func UnshareTrack(db *DB, trackID, userID int64, ev ShareEvent) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(db.Rebind(`UPDATE tracks
		SET visibility = ?, share_token = NULL, shared_at = NULL
		WHERE id = ? AND user_id = ?`),
		models.VisibilityPrivate, trackID, userID,
	)
	if err != nil {
		return fmt.Errorf("unshare track: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	ev.Action = "unshare"
	if err := insertShareEvent(db, tx, trackID, userID, models.VisibilityPrivate, ev); err != nil {
		return err
	}
	return tx.Commit()
}

func insertShareEvent(db *DB, tx *sql.Tx, trackID, userID int64, visibility string, ev ShareEvent) error {
	_, err := tx.Exec(db.Rebind(`INSERT INTO track_share_events
		(track_id, user_id, action, visibility, rights_basis, terms_ver, ip_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?)`),
		trackID, userID, ev.Action, visibility, nullIfEmpty(ev.RightsBasis), nullIfEmpty(ev.TermsVer), nullIfEmpty(ev.IPHash),
	)
	if err != nil {
		return fmt.Errorf("insert share event: %w", err)
	}
	return nil
}

func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// SharedTrackRef identifies the track behind a share token.
type SharedTrackRef struct {
	TrackID  int64
	OwnerID  int64
	SharedAt scanTime
}

// GetSharedTrack resolves a share token to its track. Private tracks and
// unknown tokens both return sql.ErrNoRows.
func GetSharedTrack(db *DB, token string) (*models.Track, *SharedTrackRef, error) {
	if token == "" {
		return nil, nil, sql.ErrNoRows
	}
	var ref SharedTrackRef
	err := db.QueryRow(db.Rebind(`SELECT id, user_id, shared_at FROM tracks
		WHERE share_token = ? AND visibility <> ?`), token, models.VisibilityPrivate,
	).Scan(&ref.TrackID, &ref.OwnerID, &ref.SharedAt)
	if err != nil {
		return nil, nil, err
	}
	t, err := GetTrack(db, ref.TrackID, ref.OwnerID)
	if err != nil {
		return nil, nil, err
	}
	return t, &ref, nil
}

// FindCopy returns the ID of the user's copy of sourceID, or sql.ErrNoRows.
func FindCopy(db *DB, sourceID, userID int64) (int64, error) {
	var id int64
	err := db.QueryRow(db.Rebind(`SELECT id FROM tracks WHERE source_track_id = ? AND user_id = ? ORDER BY id LIMIT 1`),
		sourceID, userID,
	).Scan(&id)
	return id, err
}

// CopyTrack clones a track into userID's library: metadata, notes and note
// mapping, but not the original MIDI file, pilot audio, practice tempo or any
// practice history. Copying the same source twice returns the existing copy.
func CopyTrack(db *DB, sourceID, userID int64) (int64, error) {
	if id, err := FindCopy(db, sourceID, userID); err == nil {
		return id, nil
	} else if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var newID int64
	err = tx.QueryRow(db.Rebind(`INSERT INTO tracks
		(name, filename, file_path, file_size, duration_ms, note_count, ppq, bpm, time_sig_numerator, user_id, source_track_id)
		SELECT name, filename, '', file_size, duration_ms, note_count, ppq, bpm, time_sig_numerator, CAST(? AS INTEGER), id
		FROM tracks WHERE id = ?
		RETURNING id`), userID, sourceID,
	).Scan(&newID)
	if err != nil {
		return 0, fmt.Errorf("copy track: %w", err)
	}
	if _, err := tx.Exec(db.Rebind(`INSERT INTO track_notes (track_id, note_number, time_ms, velocity, instrument)
		SELECT CAST(? AS INTEGER), note_number, time_ms, velocity, instrument FROM track_notes WHERE track_id = ? ORDER BY id`),
		newID, sourceID,
	); err != nil {
		return 0, fmt.Errorf("copy notes: %w", err)
	}
	if _, err := tx.Exec(db.Rebind(`INSERT INTO track_note_mappings (track_id, note_number, instrument)
		SELECT CAST(? AS INTEGER), note_number, instrument FROM track_note_mappings WHERE track_id = ?`),
		newID, sourceID,
	); err != nil {
		return 0, fmt.Errorf("copy mapping: %w", err)
	}
	return newID, tx.Commit()
}

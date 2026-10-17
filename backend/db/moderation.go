package db

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/simu/drum-trainer/backend/models"
)

// nullScanTime is scanTime for nullable timestamp columns.
type nullScanTime struct{ T *time.Time }

func (n *nullScanTime) Scan(v any) error {
	if v == nil {
		n.T = nil
		return nil
	}
	var st scanTime
	if err := st.Scan(v); err != nil {
		return err
	}
	n.T = &st.T
	return nil
}

// TrackFingerprint hashes a track's notes (instrument note numbers and their
// timing relative to the first note, to 10 ms), so the same drum part can be
// recognised after it is uploaded again. Empty tracks have no fingerprint.
func TrackFingerprint(db *DB, trackID int64) (string, error) {
	notes, err := GetTrackNotes(db, trackID)
	if err != nil {
		return "", err
	}
	if len(notes) == 0 {
		return "", nil
	}
	var b strings.Builder
	start := notes[0].TimeMs
	for _, n := range notes {
		b.WriteString(strconv.Itoa(n.NoteNumber))
		b.WriteByte(':')
		b.WriteString(strconv.FormatInt(int64(math.Round((n.TimeMs-start)/10)), 10))
		b.WriteByte(';')
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:]), nil
}

// IsFingerprintBlocked reports whether fp belongs to a removed track.
func IsFingerprintBlocked(db *DB, fp string) (bool, error) {
	if fp == "" {
		return false, nil
	}
	var n int
	err := db.QueryRow(db.Rebind(`SELECT COUNT(*) FROM blocked_fingerprints WHERE fingerprint = ?`), fp).Scan(&n)
	return n > 0, err
}

// CreateReport stores a report about a shared track.
func CreateReport(db *DB, r *models.Report, ipHash string) error {
	var ct scanTime
	err := db.QueryRow(db.Rebind(`INSERT INTO track_reports
		(track_id, track_name, owner_user_id, share_token, reporter_user_id, reporter_name, reporter_email,
		 kind, work_identified, details, good_faith, ip_hash)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		RETURNING id, status, created_at`),
		r.TrackID, r.TrackName, r.OwnerUserID, r.ShareToken, r.ReporterUserID, r.ReporterName, r.ReporterEmail,
		r.Kind, r.WorkIdentified, r.Details, r.GoodFaith, nullIfEmpty(ipHash),
	).Scan(&r.ID, &r.Status, &ct)
	if err != nil {
		return fmt.Errorf("insert report: %w", err)
	}
	r.CreatedAt = ct.T
	return nil
}

// ListReports returns reports with the given status ("" for all), open ones
// oldest first so the queue is worked in order, resolved ones newest first.
func ListReports(db *DB, status string) ([]models.ReportWithContext, error) {
	order := "r.id DESC"
	if status == models.ReportOpen {
		order = "r.id ASC"
	}
	rows, err := db.Query(db.Rebind(`SELECT
		r.id, r.track_id, r.track_name, r.owner_user_id, r.share_token, r.reporter_user_id, r.reporter_name,
		r.reporter_email, r.kind, r.work_identified, r.details, r.good_faith, r.status, r.resolution_note,
		r.created_at, r.resolved_at,
		t.id IS NOT NULL,
		EXISTS (SELECT 1 FROM tracks d WHERE (d.id = r.track_id OR d.origin_track_id = r.track_id) AND d.disabled_at IS NOT NULL),
		t.rights_basis,
		(SELECT COUNT(*) FROM tracks c WHERE c.origin_track_id = r.track_id),
		u.name, u.email, COALESCE(u.strikes, 0), u.sharing_banned_at IS NOT NULL
		FROM track_reports r
		LEFT JOIN tracks t ON t.id = r.track_id
		LEFT JOIN users u ON u.id = r.owner_user_id
		WHERE ? = '' OR r.status = ?
		ORDER BY `+order+` LIMIT 200`), status, status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.ReportWithContext{}
	for rows.Next() {
		var rc models.ReportWithContext
		var ownerID, reporterID sql.NullInt64
		var note, rights, ownerName, ownerEmail sql.NullString
		var ct scanTime
		var rt nullScanTime
		err := rows.Scan(
			&rc.ID, &rc.TrackID, &rc.TrackName, &ownerID, &rc.ShareToken, &reporterID, &rc.ReporterName,
			&rc.ReporterEmail, &rc.Kind, &rc.WorkIdentified, &rc.Details, &rc.GoodFaith, &rc.Status, &note,
			&ct, &rt,
			&rc.TrackExists, &rc.TrackDisabled, &rights, &rc.Copies,
			&ownerName, &ownerEmail, &rc.OwnerStrikes, &rc.OwnerBanned,
		)
		if err != nil {
			return nil, err
		}
		rc.OwnerUserID = nullInt(ownerID)
		rc.ReporterUserID = nullInt(reporterID)
		rc.ResolutionNote = nullString(note)
		rc.RightsBasis = nullString(rights)
		rc.OwnerName = nullString(ownerName)
		rc.OwnerEmail = nullString(ownerEmail)
		rc.CreatedAt = ct.T
		rc.ResolvedAt = rt.T
		out = append(out, rc)
	}
	return out, rows.Err()
}

func nullInt(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	return &v.Int64
}

func nullString(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	return &v.String
}

// RejectReport closes an open report without action.
func RejectReport(db *DB, reportID int64, note string) error {
	res, err := db.Exec(db.Rebind(`UPDATE track_reports
		SET status = ?, resolution_note = ?, resolved_at = CURRENT_TIMESTAMP
		WHERE id = ? AND status = ?`),
		models.ReportRejected, nullIfEmpty(note), reportID, models.ReportOpen,
	)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ModerationResult describes what a takedown or restore changed.
type ModerationResult struct {
	TracksAffected int  `json:"tracks_affected"` // the track plus copies
	OwnerStrikes   int  `json:"owner_strikes"`
	OwnerBanned    bool `json:"owner_banned"`
}

// DisableTrackLineage removes a shared track and every copy made from it:
// they become private and unplayable, the share link dies, the notes'
// fingerprint is blocked from being shared again and open reports about the
// track are closed as actioned. reason is shown to the owner and to everyone
// holding a copy. With strike, the owner gets a strike (once per takedown);
// reaching models.StrikeLimit bans them from sharing and makes all of their
// tracks private.
//
// It works even after the owner deleted the original, as long as a copy or a
// report still names it.
func DisableTrackLineage(db *DB, trackID int64, reason string, strike bool) (*ModerationResult, error) {
	// Reads before the transaction: SQLite runs with a single connection.
	var anyID int64
	var alreadyDisabled bool
	err := db.QueryRow(db.Rebind(`SELECT id, disabled_at IS NOT NULL FROM tracks
		WHERE id = ? OR origin_track_id = ?
		ORDER BY CASE WHEN id = ? THEN 0 ELSE 1 END, id LIMIT 1`), trackID, trackID, trackID,
	).Scan(&anyID, &alreadyDisabled)
	if err != nil {
		return nil, err
	}
	fp, err := TrackFingerprint(db, anyID)
	if err != nil {
		return nil, err
	}
	ownerID, err := lineageOwner(db, trackID)
	if err != nil {
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	if fp != "" {
		if _, err := tx.Exec(db.Rebind(`INSERT INTO blocked_fingerprints (fingerprint, track_id, reason)
			VALUES (?, ?, ?) ON CONFLICT (fingerprint) DO NOTHING`), fp, trackID, reason); err != nil {
			return nil, fmt.Errorf("block fingerprint: %w", err)
		}
	}
	res, err := tx.Exec(db.Rebind(`UPDATE tracks
		SET disabled_at = COALESCE(disabled_at, CURRENT_TIMESTAMP), disabled_reason = ?,
		    visibility = ?, share_token = NULL, shared_at = NULL, listed_at = NULL
		WHERE id = ? OR origin_track_id = ?`),
		reason, models.VisibilityPrivate, trackID, trackID,
	)
	if err != nil {
		return nil, fmt.Errorf("disable tracks: %w", err)
	}
	affected, _ := res.RowsAffected()
	if _, err := tx.Exec(db.Rebind(`UPDATE track_reports
		SET status = ?, resolution_note = ?, resolved_at = CURRENT_TIMESTAMP
		WHERE track_id = ? AND status = ?`),
		models.ReportActioned, reason, trackID, models.ReportOpen,
	); err != nil {
		return nil, fmt.Errorf("close reports: %w", err)
	}

	result := &ModerationResult{TracksAffected: int(affected)}
	if ownerID != nil {
		if strike && !alreadyDisabled {
			if _, err := tx.Exec(db.Rebind(`UPDATE users SET strikes = strikes + 1 WHERE id = ?`), *ownerID); err != nil {
				return nil, fmt.Errorf("add strike: %w", err)
			}
		}
		if err := applyStrikeLimit(db, tx, *ownerID, result); err != nil {
			return nil, err
		}
	}
	return result, tx.Commit()
}

// RestoreTrackLineage undoes DisableTrackLineage, e.g. after a counter-notice:
// the track and its copies become playable again (but stay private), the
// fingerprint is unblocked and the reports that led to the takedown are
// marked rejected. With revokeStrike the owner loses one strike, which lifts
// a sharing ban once they are below the limit again.
func RestoreTrackLineage(db *DB, trackID int64, note string, revokeStrike bool) (*ModerationResult, error) {
	ownerID, err := lineageOwner(db, trackID)
	if err != nil {
		return nil, err
	}

	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	res, err := tx.Exec(db.Rebind(`UPDATE tracks SET disabled_at = NULL, disabled_reason = NULL
		WHERE (id = ? OR origin_track_id = ?) AND disabled_at IS NOT NULL`), trackID, trackID)
	if err != nil {
		return nil, fmt.Errorf("restore tracks: %w", err)
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return nil, sql.ErrNoRows
	}
	if _, err := tx.Exec(db.Rebind(`DELETE FROM blocked_fingerprints WHERE track_id = ?`), trackID); err != nil {
		return nil, fmt.Errorf("unblock fingerprint: %w", err)
	}
	if _, err := tx.Exec(db.Rebind(`UPDATE track_reports
		SET status = ?, resolution_note = ?, resolved_at = CURRENT_TIMESTAMP
		WHERE track_id = ? AND status = ?`),
		models.ReportRejected, nullIfEmpty(note), trackID, models.ReportActioned,
	); err != nil {
		return nil, fmt.Errorf("reopen reports: %w", err)
	}

	result := &ModerationResult{TracksAffected: int(affected)}
	if ownerID != nil {
		if revokeStrike {
			if _, err := tx.Exec(db.Rebind(`UPDATE users SET strikes = strikes - 1 WHERE id = ? AND strikes > 0`), *ownerID); err != nil {
				return nil, fmt.Errorf("revoke strike: %w", err)
			}
		}
		if err := applyStrikeLimit(db, tx, *ownerID, result); err != nil {
			return nil, err
		}
	}
	return result, tx.Commit()
}

// applyStrikeLimit bans a user from sharing (and unshares all their tracks)
// once they reach the strike limit, lifts the ban below it, and fills the
// owner fields of result.
func applyStrikeLimit(db *DB, tx *sql.Tx, userID int64, result *ModerationResult) error {
	var banned bool
	if err := tx.QueryRow(db.Rebind(`SELECT strikes, sharing_banned_at IS NOT NULL FROM users WHERE id = ?`), userID).
		Scan(&result.OwnerStrikes, &banned); err != nil {
		return fmt.Errorf("read strikes: %w", err)
	}
	switch {
	case result.OwnerStrikes >= models.StrikeLimit && !banned:
		if _, err := tx.Exec(db.Rebind(`UPDATE users SET sharing_banned_at = CURRENT_TIMESTAMP WHERE id = ?`), userID); err != nil {
			return fmt.Errorf("ban user: %w", err)
		}
		if _, err := tx.Exec(db.Rebind(`UPDATE tracks SET visibility = ?, share_token = NULL, shared_at = NULL, listed_at = NULL WHERE user_id = ?`),
			models.VisibilityPrivate, userID); err != nil {
			return fmt.Errorf("unshare banned user's tracks: %w", err)
		}
		banned = true
	case result.OwnerStrikes < models.StrikeLimit && banned:
		if _, err := tx.Exec(db.Rebind(`UPDATE users SET sharing_banned_at = NULL WHERE id = ?`), userID); err != nil {
			return fmt.Errorf("lift ban: %w", err)
		}
		banned = false
	}
	result.OwnerBanned = banned
	return nil
}

// lineageOwner returns the uploader of an original track, falling back to the
// owner recorded on its reports once the track itself has been deleted.
func lineageOwner(db *DB, trackID int64) (*int64, error) {
	var id sql.NullInt64
	err := db.QueryRow(db.Rebind(`SELECT user_id FROM tracks WHERE id = ? AND origin_track_id IS NULL`), trackID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		err = db.QueryRow(db.Rebind(`SELECT owner_user_id FROM track_reports
			WHERE track_id = ? AND owner_user_id IS NOT NULL ORDER BY id DESC LIMIT 1`), trackID).Scan(&id)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return nullInt(id), nil
}

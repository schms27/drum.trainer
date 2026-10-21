package db

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/simu/drum-trainer/backend/models"
)

// ErrTagExists is returned when a user already has a tag with that name
// (ignoring case).
var ErrTagExists = errors.New("tag exists")

// ListTags returns the user's tags sorted by name, with how many tracks use each.
func ListTags(db *DB, userID int64) ([]models.Tag, error) {
	rows, err := db.Query(db.Rebind(`SELECT t.id, t.name, t.color, COUNT(tt.track_id)
		FROM tags t LEFT JOIN track_tags tt ON tt.tag_id = t.id
		WHERE t.user_id = ?
		GROUP BY t.id, t.name, t.color
		ORDER BY lower(t.name)`), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tags := []models.Tag{}
	for rows.Next() {
		var t models.Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Color, &t.TrackCount); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

// tagNameTaken reports whether the user has another tag with this name.
func tagNameTaken(db *DB, userID int64, name string, exceptID int64) (bool, error) {
	var n int
	err := db.QueryRow(db.Rebind(`SELECT COUNT(*) FROM tags WHERE user_id = ? AND lower(name) = lower(?) AND id <> ?`),
		userID, name, exceptID).Scan(&n)
	return n > 0, err
}

// CreateTag adds a tag for the user. Returns ErrTagExists for a duplicate name.
func CreateTag(db *DB, userID int64, name, color string) (*models.Tag, error) {
	if taken, err := tagNameTaken(db, userID, name, 0); err != nil {
		return nil, err
	} else if taken {
		return nil, ErrTagExists
	}
	t := &models.Tag{Name: name, Color: color}
	err := db.QueryRow(db.Rebind(`INSERT INTO tags (user_id, name, color) VALUES (?, ?, ?) RETURNING id`),
		userID, name, color).Scan(&t.ID)
	if err != nil {
		return nil, fmt.Errorf("insert tag: %w", err)
	}
	return t, nil
}

// UpdateTag renames and/or recolours one of the user's tags (nil leaves a
// field as it is). Returns sql.ErrNoRows if the tag isn't the user's.
func UpdateTag(db *DB, id, userID int64, name, color *string) error {
	if name != nil {
		if taken, err := tagNameTaken(db, userID, *name, id); err != nil {
			return err
		} else if taken {
			return ErrTagExists
		}
	}
	res, err := db.Exec(db.Rebind(`UPDATE tags SET name = COALESCE(?, name), color = COALESCE(?, color)
		WHERE id = ? AND user_id = ?`), name, color, id, userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// DeleteTag removes one of the user's tags from every track and deletes it.
func DeleteTag(db *DB, id, userID int64) error {
	res, err := db.Exec(db.Rebind(`DELETE FROM tags WHERE id = ? AND user_id = ?`), id, userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// SetTrackTags replaces the tags on one of the user's tracks. Every tag must
// be the user's own; otherwise nothing changes and sql.ErrNoRows is returned,
// like for a track that isn't theirs.
func SetTrackTags(db *DB, trackID, userID int64, tagIDs []int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var n int
	if err := tx.QueryRow(db.Rebind(`SELECT COUNT(*) FROM tracks WHERE id = ? AND user_id = ?`), trackID, userID).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	unique := make(map[int64]bool, len(tagIDs))
	for _, id := range tagIDs {
		unique[id] = true
	}
	if len(unique) > 0 {
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(unique)), ",")
		args := []any{userID}
		for id := range unique {
			args = append(args, id)
		}
		var owned int
		if err := tx.QueryRow(db.Rebind(`SELECT COUNT(*) FROM tags WHERE user_id = ? AND id IN (`+placeholders+`)`), args...).Scan(&owned); err != nil {
			return err
		}
		if owned != len(unique) {
			return sql.ErrNoRows
		}
	}
	if _, err := tx.Exec(db.Rebind(`DELETE FROM track_tags WHERE track_id = ?`), trackID); err != nil {
		return err
	}
	for id := range unique {
		if _, err := tx.Exec(db.Rebind(`INSERT INTO track_tags (track_id, tag_id) VALUES (?, ?)`), trackID, id); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// TrackTags returns the tags on the user's tracks, keyed by track ID and
// sorted by name. Pass trackID > 0 to load a single track's tags.
func TrackTags(db *DB, userID, trackID int64) (map[int64][]models.TagRef, error) {
	query := `SELECT tt.track_id, t.id, t.name, t.color
		FROM track_tags tt JOIN tags t ON t.id = tt.tag_id
		WHERE t.user_id = ?`
	args := []any{userID}
	if trackID > 0 {
		query += ` AND tt.track_id = ?`
		args = append(args, trackID)
	}
	rows, err := db.Query(db.Rebind(query), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make(map[int64][]models.TagRef)
	for rows.Next() {
		var track int64
		var t models.TagRef
		if err := rows.Scan(&track, &t.ID, &t.Name, &t.Color); err != nil {
			return nil, err
		}
		out[track] = append(out[track], t)
	}
	for _, list := range out {
		sort.Slice(list, func(i, j int) bool { return strings.ToLower(list[i].Name) < strings.ToLower(list[j].Name) })
	}
	return out, rows.Err()
}

// RenameTrack changes the name of one of the user's tracks. Renaming a shared
// track is logged as a share event, since the name shows on the share link.
func RenameTrack(db *DB, trackID, userID int64, name string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var visibility string
	err = tx.QueryRow(db.Rebind(`UPDATE tracks SET name = ? WHERE id = ? AND user_id = ? RETURNING visibility`),
		name, trackID, userID).Scan(&visibility)
	if err != nil {
		return err
	}
	if visibility != models.VisibilityPrivate {
		if err := insertShareEvent(db, tx, trackID, userID, visibility, ShareEvent{Action: "rename"}); err != nil {
			return err
		}
	}
	return tx.Commit()
}

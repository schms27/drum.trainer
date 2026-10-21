package db

import (
	"database/sql"
	"fmt"

	"github.com/simu/drum-trainer/backend/models"
)

// ListKitProfiles returns the user's kit profiles with their mappings,
// oldest first, marking the active one.
func ListKitProfiles(db *DB, userID int64) ([]models.KitProfile, error) {
	var active sql.NullInt64
	if err := db.QueryRow(db.Rebind(`SELECT active_kit_profile_id FROM users WHERE id = ?`), userID).Scan(&active); err != nil {
		return nil, err
	}

	rows, err := db.Query(
		db.Rebind(`SELECT id, name, created_at, metronome_volume, pilot_volume FROM kit_profiles WHERE user_id = ? ORDER BY id ASC`),
		userID,
	)
	if err != nil {
		return nil, err
	}
	var kits []models.KitProfile
	for rows.Next() {
		var k models.KitProfile
		var ct scanTime
		if err := rows.Scan(&k.ID, &k.Name, &ct, &k.MetronomeVolume, &k.PilotVolume); err != nil {
			rows.Close()
			return nil, err
		}
		k.CreatedAt = ct.T
		k.Active = active.Valid && active.Int64 == k.ID
		kits = append(kits, k)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range kits {
		m, err := kitMappings(db, kits[i].ID)
		if err != nil {
			return nil, err
		}
		kits[i].Mappings = m
	}
	return kits, nil
}

// GetKitProfile returns one of the user's kit profiles, or sql.ErrNoRows.
func GetKitProfile(db *DB, id, userID int64) (*models.KitProfile, error) {
	kits, err := ListKitProfiles(db, userID)
	if err != nil {
		return nil, err
	}
	for _, k := range kits {
		if k.ID == id {
			return &k, nil
		}
	}
	return nil, sql.ErrNoRows
}

func kitMappings(db *DB, profileID int64) ([]models.NoteMapping, error) {
	rows, err := db.Query(
		db.Rebind(`SELECT note_number, instrument FROM kit_profile_mappings WHERE profile_id = ? ORDER BY note_number`),
		profileID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	mappings := []models.NoteMapping{}
	for rows.Next() {
		var m models.NoteMapping
		if err := rows.Scan(&m.NoteNumber, &m.Instrument); err != nil {
			return nil, err
		}
		mappings = append(mappings, m)
	}
	return mappings, rows.Err()
}

// CreateKitProfile stores a new kit profile. The user's first profile
// becomes their active one.
func CreateKitProfile(db *DB, userID int64, name string, mappings []models.NoteMapping) (*models.KitProfile, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var id int64
	if err := tx.QueryRow(
		db.Rebind(`INSERT INTO kit_profiles (user_id, name) VALUES (?, ?) RETURNING id`), userID, name,
	).Scan(&id); err != nil {
		return nil, fmt.Errorf("insert kit profile: %w", err)
	}
	if err := writeKitMappings(db, tx, id, mappings); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(
		db.Rebind(`UPDATE users SET active_kit_profile_id = ? WHERE id = ? AND active_kit_profile_id IS NULL`), id, userID,
	); err != nil {
		return nil, fmt.Errorf("activate kit profile: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return GetKitProfile(db, id, userID)
}

// UpdateKitProfile replaces a kit profile's name and mappings.
// Returns sql.ErrNoRows if the profile does not belong to the user.
func UpdateKitProfile(db *DB, id, userID int64, name string, mappings []models.NoteMapping) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(db.Rebind(`UPDATE kit_profiles SET name = ? WHERE id = ? AND user_id = ?`), name, id, userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	if _, err := tx.Exec(db.Rebind(`DELETE FROM kit_profile_mappings WHERE profile_id = ?`), id); err != nil {
		return err
	}
	if err := writeKitMappings(db, tx, id, mappings); err != nil {
		return err
	}
	return tx.Commit()
}

func writeKitMappings(db *DB, tx *sql.Tx, profileID int64, mappings []models.NoteMapping) error {
	stmt, err := tx.Prepare(db.Rebind(`INSERT INTO kit_profile_mappings (profile_id, note_number, instrument) VALUES (?, ?, ?)`))
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, m := range mappings {
		if _, err := stmt.Exec(profileID, m.NoteNumber, m.Instrument); err != nil {
			return fmt.Errorf("insert kit mapping: %w", err)
		}
	}
	return nil
}

// DeleteKitProfile removes a kit profile (and clears it as active).
// Returns sql.ErrNoRows if the profile does not belong to the user.
func DeleteKitProfile(db *DB, id, userID int64) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// Clear explicitly rather than relying on ON DELETE SET NULL, which SQLite
	// only enforces with foreign keys enabled.
	if _, err := tx.Exec(
		db.Rebind(`UPDATE users SET active_kit_profile_id = NULL WHERE id = ? AND active_kit_profile_id = ?`), userID, id,
	); err != nil {
		return err
	}
	res, err := tx.Exec(db.Rebind(`DELETE FROM kit_profiles WHERE id = ? AND user_id = ?`), id, userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return tx.Commit()
}

// SetActiveKitProfile selects the profile used for scoring the user's hits;
// nil means none (standard General MIDI numbering). Returns sql.ErrNoRows if
// the profile does not belong to the user.
func SetActiveKitProfile(db *DB, userID int64, id *int64) error {
	if id != nil {
		var n int
		if err := db.QueryRow(
			db.Rebind(`SELECT COUNT(*) FROM kit_profiles WHERE id = ? AND user_id = ?`), *id, userID,
		).Scan(&n); err != nil {
			return err
		}
		if n == 0 {
			return sql.ErrNoRows
		}
	}
	_, err := db.Exec(db.Rebind(`UPDATE users SET active_kit_profile_id = ? WHERE id = ?`), id, userID)
	return err
}

// SetKitMix saves a kit profile's default playback levels (nil clears one).
func SetKitMix(db *DB, id, userID int64, metronome, pilot *float64) error {
	res, err := db.Exec(db.Rebind(`UPDATE kit_profiles SET metronome_volume = ?, pilot_volume = ? WHERE id = ? AND user_id = ?`),
		metronome, pilot, id, userID)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// ActiveKitMapping returns the note→instrument overrides of the user's active
// kit profile; ok is false when no profile is active.
func ActiveKitMapping(db *DB, userID int64) (mapping map[int]string, ok bool, err error) {
	var id sql.NullInt64
	if err := db.QueryRow(db.Rebind(`SELECT active_kit_profile_id FROM users WHERE id = ?`), userID).Scan(&id); err != nil {
		return nil, false, err
	}
	if !id.Valid {
		return nil, false, nil
	}
	list, err := kitMappings(db, id.Int64)
	if err != nil {
		return nil, false, err
	}
	mapping = make(map[int]string, len(list))
	for _, m := range list {
		mapping[m.NoteNumber] = m.Instrument
	}
	return mapping, true, nil
}

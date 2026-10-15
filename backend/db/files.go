package db

import (
	"fmt"
	"strings"
)

// NormalizeFilePaths rewrites file_path values stored by older versions (full
// paths including UPLOAD_DIR) to bare file names, which is what the storage
// layer now expects. It is idempotent and cheap to run on every startup.
func NormalizeFilePaths(db *DB) error {
	for _, table := range []string{"tracks", "pilot_tracks"} {
		if err := normalizeTable(db, table); err != nil {
			return fmt.Errorf("normalize %s.file_path: %w", table, err)
		}
	}
	return nil
}

func normalizeTable(db *DB, table string) error {
	// Filter in Go rather than with LIKE: backslash is an escape character in
	// PostgreSQL LIKE patterns but not in SQLite.
	rows, err := db.Query(`SELECT id, file_path FROM ` + table)
	if err != nil {
		return err
	}
	type row struct {
		id   int64
		path string
	}
	var pending []row
	for rows.Next() {
		var r row
		if err := rows.Scan(&r.id, &r.path); err != nil {
			rows.Close()
			return err
		}
		if baseName(r.path) != r.path {
			pending = append(pending, r)
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}

	for _, r := range pending {
		if _, err := db.Exec(db.Rebind(`UPDATE `+table+` SET file_path = ? WHERE id = ?`), baseName(r.path), r.id); err != nil {
			return err
		}
	}
	return nil
}

// baseName returns the last element of p, treating both / and \ as
// separators so paths written on Windows normalize correctly on Linux.
func baseName(p string) string {
	if i := strings.LastIndexAny(p, `/\`); i >= 0 {
		return p[i+1:]
	}
	return p
}

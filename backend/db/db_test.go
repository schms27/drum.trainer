package db

import (
	"path/filepath"
	"testing"

	"github.com/simu/drum-trainer/backend/config"
)

// openTestDB opens a fresh, fully migrated SQLite database in a temp dir.
func openTestDB(t *testing.T) *DB {
	t.Helper()
	db, err := Open(config.Config{DBPath: filepath.Join(t.TempDir(), "test.db")})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestOpenAppliesAllMigrations(t *testing.T) {
	db := openTestDB(t)

	entries, err := migrationsFS.ReadDir("migrations/sqlite")
	if err != nil {
		t.Fatal(err)
	}
	var applied int
	if err := db.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&applied); err != nil {
		t.Fatal(err)
	}
	if applied != len(entries) {
		t.Fatalf("applied %d migrations, want %d", applied, len(entries))
	}
}

func TestSQLitePragmas(t *testing.T) {
	db := openTestDB(t)

	var fk int
	if err := db.QueryRow(`PRAGMA foreign_keys`).Scan(&fk); err != nil {
		t.Fatal(err)
	}
	if fk != 1 {
		t.Fatalf("foreign_keys = %d, want 1", fk)
	}
	var mode string
	if err := db.QueryRow(`PRAGMA journal_mode`).Scan(&mode); err != nil {
		t.Fatal(err)
	}
	if mode != "wal" {
		t.Fatalf("journal_mode = %q, want wal", mode)
	}
}

func TestFailedMigrationIsRolledBack(t *testing.T) {
	db := openTestDB(t)

	err := db.applyMigration("999_broken.sql", `CREATE TABLE half_done (id INTEGER); THIS IS NOT SQL;`)
	if err == nil {
		t.Fatal("expected error")
	}
	var n int
	db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE name = 'half_done'`).Scan(&n)
	if n != 0 {
		t.Fatal("table from failed migration was not rolled back")
	}
	db.QueryRow(`SELECT COUNT(*) FROM schema_migrations WHERE version = '999_broken.sql'`).Scan(&n)
	if n != 0 {
		t.Fatal("failed migration was recorded")
	}
}

func TestRebind(t *testing.T) {
	q := `SELECT * FROM t WHERE a = ? AND b = ?`
	if got := (&DB{}).Rebind(q); got != q {
		t.Errorf("sqlite: got %q", got)
	}
	if got, want := (&DB{postgres: true}).Rebind(q), `SELECT * FROM t WHERE a = $1 AND b = $2`; got != want {
		t.Errorf("postgres: got %q, want %q", got, want)
	}
}

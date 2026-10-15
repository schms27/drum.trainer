package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/simu/drum-trainer/backend/config"
	_ "modernc.org/sqlite"
)

//go:embed migrations
var migrationsFS embed.FS

// DB wraps *sql.DB with knowledge of the SQL dialect in use, so queries can be
// written once with ? placeholders and rebound for PostgreSQL.
type DB struct {
	*sql.DB
	postgres bool
}

// Open opens (or creates) the database and runs any pending migrations.
// When cfg.DatabaseURL starts with "postgres://" or "postgresql://", PostgreSQL
// is used via the pgx driver; otherwise SQLite is used (zero-config default).
func Open(cfg config.Config) (*DB, error) {
	var db *DB

	if strings.HasPrefix(cfg.DatabaseURL, "postgres://") || strings.HasPrefix(cfg.DatabaseURL, "postgresql://") {
		sqlDB, err := sql.Open("pgx", cfg.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("open postgres: %w", err)
		}
		sqlDB.SetMaxOpenConns(25)
		db = &DB{DB: sqlDB, postgres: true}
	} else {
		// Pragmas go in the DSN so they apply to every connection the pool
		// opens: foreign_keys is per-connection (and needed for ON DELETE
		// CASCADE), and journal_mode cannot be changed inside a transaction.
		dsn := cfg.DBPath + "?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
		sqlDB, err := sql.Open("sqlite", dsn)
		if err != nil {
			return nil, fmt.Errorf("open sqlite: %w", err)
		}
		// SQLite performs best with a single writer connection.
		sqlDB.SetMaxOpenConns(1)
		db = &DB{DB: sqlDB}
	}

	if err := db.runMigrations(); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return db, nil
}

// Rebind replaces ? placeholders with $1, $2, … for PostgreSQL.
// It is a no-op for SQLite.
func (db *DB) Rebind(query string) string {
	if !db.postgres {
		return query
	}
	n := 0
	var b strings.Builder
	for _, ch := range query {
		if ch == '?' {
			n++
			fmt.Fprintf(&b, "$%d", n)
		} else {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func (db *DB) runMigrations() error {
	// Tracking table — safe to create on every startup.
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	subdir := "migrations/sqlite"
	if db.postgres {
		subdir = "migrations/postgres"
	}

	entries, err := fs.ReadDir(migrationsFS, subdir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	for _, entry := range entries {
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		// Skip already-applied migrations.
		var n int
		if err := db.QueryRow(db.Rebind(`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`), name).Scan(&n); err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if n > 0 {
			continue
		}

		data, err := migrationsFS.ReadFile(subdir + "/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if err := db.applyMigration(name, string(data)); err != nil {
			return err
		}
	}
	return nil
}

// applyMigration runs one migration and records it in a single transaction,
// so a failing migration leaves neither partial schema changes nor a
// schema_migrations row behind. (Both PostgreSQL and SQLite support
// transactional DDL.)
func (db *DB) applyMigration(name, sqlText string) error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("begin migration %s: %w", name, err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(sqlText); err != nil {
		return fmt.Errorf("apply migration %s: %w", name, err)
	}
	if _, err := tx.Exec(db.Rebind(`INSERT INTO schema_migrations (version) VALUES (?)`), name); err != nil {
		return fmt.Errorf("record migration %s: %w", name, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", name, err)
	}
	return nil
}

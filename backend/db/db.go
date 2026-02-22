package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/simu/drum-trainer/backend/config"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"
)

//go:embed migrations
var migrationsFS embed.FS

var driverName string

// Open opens (or creates) the database and runs any pending migrations.
// When cfg.DatabaseURL starts with "postgres://" or "postgresql://", PostgreSQL
// is used via the pgx driver; otherwise SQLite is used (zero-config default).
func Open(cfg config.Config) (*sql.DB, error) {
	var (
		db  *sql.DB
		err error
	)

	if strings.HasPrefix(cfg.DatabaseURL, "postgres://") || strings.HasPrefix(cfg.DatabaseURL, "postgresql://") {
		db, err = sql.Open("pgx", cfg.DatabaseURL)
		if err != nil {
			return nil, fmt.Errorf("open postgres: %w", err)
		}
		db.SetMaxOpenConns(25)
		driverName = "pgx"
	} else {
		db, err = sql.Open("sqlite", cfg.DBPath)
		if err != nil {
			return nil, fmt.Errorf("open sqlite: %w", err)
		}
		// SQLite performs best with a single writer connection.
		db.SetMaxOpenConns(1)
		driverName = "sqlite"
	}

	if err := runMigrations(db, driverName); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return db, nil
}

// Rebind replaces ? placeholders with $1, $2, … for PostgreSQL.
// It is a no-op for SQLite.
func Rebind(query string) string {
	if driverName != "pgx" {
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

func runMigrations(db *sql.DB, driver string) error {
	// Tracking table — safe to create on every startup.
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY)`); err != nil {
		return fmt.Errorf("create schema_migrations: %w", err)
	}

	subdir := "migrations/sqlite"
	if driver == "pgx" {
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
		if err := db.QueryRow(Rebind(`SELECT COUNT(*) FROM schema_migrations WHERE version = ?`), name).Scan(&n); err != nil {
			return fmt.Errorf("check migration %s: %w", name, err)
		}
		if n > 0 {
			continue
		}

		data, err := migrationsFS.ReadFile(subdir + "/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if _, err := db.Exec(string(data)); err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
		if _, err := db.Exec(Rebind(`INSERT INTO schema_migrations (version) VALUES (?)`), name); err != nil {
			return fmt.Errorf("record migration %s: %w", name, err)
		}
	}
	return nil
}

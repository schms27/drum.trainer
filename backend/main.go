package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
	_ "time/tzdata" // embed the zone database so ?tz= works without OS tzdata

	"github.com/simu/drum-trainer/backend/api"
	"github.com/simu/drum-trainer/backend/config"
	"github.com/simu/drum-trainer/backend/db"
	"github.com/simu/drum-trainer/backend/storage"
)

func main() {
	cfg := config.Load()
	if cfg.SessionSecret == config.DefaultSessionSecret && !cfg.DevMode {
		log.Printf("WARNING: SESSION_SECRET is not set; OAuth state cookies are signed with a public default. Generate one with: openssl rand -hex 32")
	}

	// Ensure local data directory exists (SQLite only).
	if cfg.DatabaseURL == "" {
		if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0755); err != nil {
			log.Fatalf("create data dir: %v", err)
		}
	}

	database, err := db.Open(cfg)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer database.Close()

	if err := db.NormalizeFilePaths(database); err != nil {
		log.Fatalf("normalize stored file paths: %v", err)
	}

	store, err := storage.NewStore(cfg.UploadDir)
	if err != nil {
		log.Fatalf("init storage: %v", err)
	}

	maxBytes      := cfg.MaxUploadSizeMB << 20 // MB → bytes
	maxPilotBytes := cfg.MaxPilotSizeMB  << 20
	router := api.NewRouter(database, store, maxBytes, cfg, maxPilotBytes)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.Port),
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
		// Generous read/write limits: pilot uploads and audio streaming can be
		// up to MAX_PILOT_SIZE_MB over slow connections.
		ReadTimeout:  5 * time.Minute,
		WriteTimeout: 5 * time.Minute,
		IdleTimeout:  2 * time.Minute,
	}

	// Shut down gracefully on SIGINT/SIGTERM (docker stop), letting in-flight
	// requests finish before the database is closed.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go purgeExpiredSessions(ctx, database, time.Hour)

	serverErr := make(chan error, 1)
	go func() {
		log.Printf("drum-trainer backend listening on http://localhost%s", srv.Addr)
		serverErr <- srv.ListenAndServe()
	}()

	select {
	case err := <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	case <-ctx.Done():
		log.Printf("shutting down…")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
		}
	}
}

// purgeExpiredSessions deletes expired login sessions at startup and then on
// every tick until ctx is cancelled.
func purgeExpiredSessions(ctx context.Context, database *db.DB, every time.Duration) {
	ticker := time.NewTicker(every)
	defer ticker.Stop()
	for {
		if n, err := db.DeleteExpiredSessions(database); err != nil {
			log.Printf("purge expired sessions: %v", err)
		} else if n > 0 {
			log.Printf("purged %d expired sessions", n)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

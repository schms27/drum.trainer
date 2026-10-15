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

	"github.com/simu/drum-trainer/backend/api"
	"github.com/simu/drum-trainer/backend/config"
	"github.com/simu/drum-trainer/backend/db"
	"github.com/simu/drum-trainer/backend/storage"
)

func main() {
	cfg := config.Load()

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

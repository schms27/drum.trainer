package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/simu/drum-trainer/backend/api"
	"github.com/simu/drum-trainer/backend/config"
	"github.com/simu/drum-trainer/backend/db"
	"github.com/simu/drum-trainer/backend/storage"
)

func main() {
	cfg := config.Load()

	// Ensure data directory exists.
	if err := os.MkdirAll(filepath.Dir(cfg.DBPath), 0755); err != nil {
		log.Fatalf("create data dir: %v", err)
	}

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer database.Close()

	store, err := storage.NewStore(cfg.UploadDir)
	if err != nil {
		log.Fatalf("init storage: %v", err)
	}

	maxBytes := cfg.MaxUploadSizeMB << 20 // MB → bytes
	router := api.NewRouter(database, store, maxBytes)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("drum-trainer backend listening on http://localhost%s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

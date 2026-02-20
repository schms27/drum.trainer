package api

import (
	"database/sql"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/simu/drum-trainer/backend/storage"
)

// NewRouter builds and returns the fully configured chi router.
func NewRouter(db *sql.DB, store *storage.Store, maxUploadBytes int64) http.Handler {
	r := chi.NewRouter()

	// Middleware stack.
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:4173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	th := &tracksHandler{db: db, store: store, maxBytes: maxUploadBytes}
	sh := &sessionsHandler{db: db}

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", healthHandler)

		// Tracks.
		r.Post("/tracks", th.create)
		r.Get("/tracks", th.list)
		r.Get("/tracks/{id}", th.get)
		r.Delete("/tracks/{id}", th.delete)
		r.Get("/tracks/{id}/mapping", th.getMapping)
		r.Put("/tracks/{id}/mapping", th.updateMapping)

		// Sessions.
		r.Post("/sessions", sh.create)
		r.Get("/sessions", sh.list)
		r.Get("/sessions/{id}", sh.get)
	})

	return r
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": "1.0.0"})
}

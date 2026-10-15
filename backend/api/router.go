package api

import (
	"database/sql"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/simu/drum-trainer/backend/config"
	"github.com/simu/drum-trainer/backend/storage"
)

// NewRouter builds and returns the fully configured chi router.
func NewRouter(db *sql.DB, store *storage.Store, maxUploadBytes int64, cfg config.Config, maxPilotBytes int64) http.Handler {
	r := chi.NewRouter()

	// Middleware stack.
	// RealIP rewrites RemoteAddr from X-Real-IP / X-Forwarded-For so that the
	// per-IP rate limiters see the client, not the reverse proxy. The backend
	// must therefore only be reachable through the trusted proxy.
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.SetHeader("X-Content-Type-Options", "nosniff"))
	r.Use(httprate.LimitByIP(200, time.Minute))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:4173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	ah := &authHandler{db: db, cfg: cfg}
	th := &tracksHandler{db: db, store: store, maxBytes: maxUploadBytes}
	sh := &sessionsHandler{db: db, maxHits: cfg.MaxSessionHits}
	sth := &statsHandler{db: db}
	ph := &pilotsHandler{db: db, store: store, maxBytes: maxPilotBytes}

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", healthHandler)

		// Auth — public (no session required).
		r.With(httprate.LimitByIP(10, time.Minute)).Get("/auth/{provider}/login", ah.login)
		r.Get("/auth/{provider}/callback", ah.callback)
		r.Post("/auth/logout", ah.logout)
		r.Get("/auth/me", ah.me)

		// Protected resources — session cookie required.
		r.Group(func(r chi.Router) {
			r.Use(RequireAuth(db))

			// Tracks.
			r.With(httprate.LimitByIP(20, time.Minute)).Post("/tracks", th.create)
			r.Get("/tracks", th.list)
			r.Get("/tracks/{id}", th.get)
			r.Delete("/tracks/{id}", th.delete)
			r.Patch("/tracks/{id}/practice-bpm", th.setPracticeBPM)
			r.Get("/tracks/{id}/mapping", th.getMapping)
			r.Put("/tracks/{id}/mapping", th.updateMapping)

			// Sessions.
			r.With(httprate.LimitByIP(30, time.Minute)).Post("/sessions", sh.create)
			r.Get("/sessions", sh.list)
			r.Get("/sessions/{id}", sh.get)

			// Stats.
			r.Get("/stats", sth.get)

			// Pilot tracks.
			r.With(httprate.LimitByIP(20, time.Minute)).Post("/tracks/{id}/pilots", ph.create)
			r.Get("/tracks/{id}/pilots", ph.list)
			r.Delete("/tracks/{id}/pilots/{pilotId}", ph.delete)
			r.Get("/tracks/{id}/pilots/{pilotId}/audio", ph.audio)
		})
	})

	return r
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": "1.0.0"})
}

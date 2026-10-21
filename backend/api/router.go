package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/simu/drum-trainer/backend/config"
	dbpkg "github.com/simu/drum-trainer/backend/db"
	"github.com/simu/drum-trainer/backend/storage"
)

// Version is the build version reported by /health. It is set at build time:
//
//	go build -ldflags "-X github.com/simu/drum-trainer/backend/api.Version=1.2.3"
var Version = "dev"

// NewRouter builds and returns the fully configured chi router.
func NewRouter(db *dbpkg.DB, store *storage.Store, maxUploadBytes int64, cfg config.Config, maxPilotBytes int64) http.Handler {
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
	kh := &kitsHandler{db: db}
	shh := &sharingHandler{db: db, secret: cfg.SessionSecret}
	rh := &reportsHandler{db: db, secret: cfg.SessionSecret}
	adh := &adminHandler{db: db}
	eh := &exploreHandler{db: db}

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/health", healthHandler)
		r.Get("/site", siteHandler(cfg))

		// Auth — public (no session required).
		r.With(httprate.LimitByIP(10, time.Minute)).Get("/auth/{provider}/login", ah.login)
		r.Get("/auth/{provider}/callback", ah.callback)
		r.Post("/auth/logout", ah.logout)
		r.Get("/auth/me", ah.me)

		// Shared track landing page — public, the link token is the credential.
		r.With(httprate.LimitByIP(60, time.Minute)).Get("/shared/{token}", shh.get)

		// Explore gallery of public tracks — public.
		r.With(httprate.LimitByIP(60, time.Minute)).Get("/explore", eh.list)

		// Copyright / abuse reports — public so rights holders need no account.
		r.With(httprate.LimitByIP(5, time.Hour)).Post("/reports", rh.create)

		// Protected resources — session cookie required.
		r.Group(func(r chi.Router) {
			r.Use(RequireAuth(db))

			// Tracks.
			r.With(httprate.LimitByIP(20, time.Minute)).Post("/tracks", th.create)
			r.Get("/tracks", th.list)
			r.Get("/tracks/{id}", th.get)
			r.Delete("/tracks/{id}", th.delete)
			r.Patch("/tracks/{id}/practice-bpm", th.setPracticeBPM)
			r.Put("/tracks/{id}/notes", th.replaceNotes)
			r.Get("/tracks/{id}/mapping", th.getMapping)
			r.Put("/tracks/{id}/mapping", th.updateMapping)
			r.Put("/tracks/{id}/mix", th.setMix)
			r.Get("/tracks/{id}/file", th.download)
			r.Get("/tracks/{id}/export.mid", th.export)

			// Sharing.
			r.Put("/tracks/{id}/sharing", shh.share)
			r.Post("/tracks/{id}/sharing/rotate", shh.rotate)
			r.Delete("/tracks/{id}/sharing", shh.unshare)
			r.With(httprate.LimitByIP(20, time.Minute)).Post("/shared/{token}/copy", shh.copy)

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

			// Kit profiles (how the user's pads map to instruments).
			r.Get("/kits", kh.list)
			r.Post("/kits", kh.create)
			r.Put("/kits/active", kh.setActive)
			r.Put("/kits/{id}", kh.update)
			r.Put("/kits/{id}/mix", kh.setMix)
			r.Delete("/kits/{id}", kh.delete)

			// Moderation — ADMIN_EMAILS only.
			r.Route("/admin", func(r chi.Router) {
				r.Use(RequireAdmin(cfg))
				r.Get("/reports", adh.listReports)
				r.Post("/reports/{id}/reject", adh.rejectReport)
				r.Post("/tracks/{id}/disable", adh.disableTrack)
				r.Post("/tracks/{id}/restore", adh.restoreTrack)
			})
		})
	})

	return r
}

// siteHandler serves public, non-secret site settings to the frontend, which is
// a prebuilt static app and can't be configured at deploy time itself.
func siteHandler(cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"contact_email": cfg.ContactEmail})
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": Version})
}

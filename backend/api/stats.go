package api

import (
	"net/http"
	"time"

	dbpkg "github.com/simu/drum-trainer/backend/db"
)

type statsHandler struct {
	db *dbpkg.DB
}

// GET /api/v1/stats?tz=Europe/Zurich
// tz is an optional IANA time zone used for week boundaries (default UTC).
func (h *statsHandler) get(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())

	loc := time.UTC
	if tz := r.URL.Query().Get("tz"); tz != "" {
		l, err := time.LoadLocation(tz)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid tz")
			return
		}
		loc = l
	}

	resp, err := dbpkg.GetWeeklyStats(h.db, user.ID, loc)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch stats")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

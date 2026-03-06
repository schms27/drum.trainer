package api

import (
	"database/sql"
	"net/http"

	dbpkg "github.com/simu/drum-trainer/backend/db"
)

type statsHandler struct {
	db *sql.DB
}

// GET /api/v1/stats
func (h *statsHandler) get(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())

	resp, err := dbpkg.GetWeeklyStats(h.db, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch stats")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	dbpkg "github.com/simu/drum-trainer/backend/db"
	midipkg "github.com/simu/drum-trainer/backend/midi"
	"github.com/simu/drum-trainer/backend/models"
	"github.com/simu/drum-trainer/backend/scoring"
)

type sessionsHandler struct {
	db *sql.DB
}

// POST /api/v1/sessions
func (h *sessionsHandler) create(w http.ResponseWriter, r *http.Request) {
	var req models.SubmitSessionRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.TrackID == 0 || req.StartedAt == "" || req.EndedAt == "" {
		writeError(w, http.StatusBadRequest, "track_id, started_at, and ended_at are required")
		return
	}

	// Verify track exists and load target notes.
	_, err := dbpkg.GetTrack(h.db, req.TrackID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "track not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to fetch track")
		}
		return
	}

	targetNotes, err := dbpkg.GetTrackNotes(h.db, req.TrackID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch target notes")
		return
	}

	startedAt, err := time.Parse(time.RFC3339, req.StartedAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid started_at format (use RFC3339)")
		return
	}
	endedAt, err := time.Parse(time.RFC3339, req.EndedAt)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid ended_at format (use RFC3339)")
		return
	}

	// Run the scoring algorithm.
	score := scoring.Score(targetNotes, req.Hits, midipkg.InstrumentGroup)

	session := &models.Session{
		TrackID:    req.TrackID,
		StartedAt:  startedAt,
		EndedAt:    endedAt,
		DurationMs: float64(endedAt.Sub(startedAt).Milliseconds()),
		HitCount:   len(req.Hits),
	}

	// Build instrument group map for raw hit annotation.
	instrMap := make(map[string]string)
	for note, group := range midipkg.DrumMap {
		instrMap[strconv.Itoa(note)] = group
	}

	sessionID, err := dbpkg.InsertSession(h.db, session, req.Hits, score, instrMap)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save session")
		return
	}

	session.ID = sessionID
	writeJSON(w, http.StatusCreated, models.SessionWithScore{
		Session: *session,
		Score:   *score,
	})
}

// GET /api/v1/sessions
func (h *sessionsHandler) list(w http.ResponseWriter, r *http.Request) {
	trackID := int64(0)
	if s := r.URL.Query().Get("track_id"); s != "" {
		var err error
		trackID, err = strconv.ParseInt(s, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid track_id")
			return
		}
	}

	sessions, err := dbpkg.ListSessions(h.db, trackID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list sessions")
		return
	}
	if sessions == nil {
		sessions = []models.SessionSummary{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"sessions": sessions})
}

// GET /api/v1/sessions/{id}
func (h *sessionsHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	session, err := dbpkg.GetSession(h.db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "session not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to fetch session")
		}
		return
	}

	score, err := dbpkg.GetSessionScore(h.db, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch score")
		return
	}

	writeJSON(w, http.StatusOK, models.SessionWithScore{
		Session: *session,
		Score:   *score,
	})
}

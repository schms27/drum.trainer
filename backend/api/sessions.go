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

// minHitTimeMs is the earliest accepted hit time relative to the start of the track.
const minHitTimeMs = -1000.0

type sessionsHandler struct {
	db      *dbpkg.DB
	maxHits int
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
	if len(req.Hits) > h.maxHits {
		writeError(w, http.StatusBadRequest, "too many hits")
		return
	}
	for _, h := range req.Hits {
		if h.NoteNumber < 0 || h.NoteNumber > 127 {
			writeError(w, http.StatusBadRequest, "hit note_number out of range (0–127)")
			return
		}
		if h.Velocity < 1 || h.Velocity > 127 {
			writeError(w, http.StatusBadRequest, "hit velocity out of range (1–127)")
			return
		}
		// Slightly negative times are valid: with a count-in, time zero is the
		// first beat and a first note played early lands just before it.
		if h.TimeMs < minHitTimeMs {
			writeError(w, http.StatusBadRequest, "hit time_ms must be at least -1000")
			return
		}
	}

	user := UserFromContext(r.Context())

	// Verify track exists and belongs to this user, then load target notes.
	track, err := dbpkg.GetTrack(h.db, req.TrackID, user.ID)
	if err == nil && track.Disabled {
		writeError(w, http.StatusConflict, "this track was removed after a report and can't be practised")
		return
	}
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
	if !endedAt.After(startedAt) {
		writeError(w, http.StatusBadRequest, "ended_at must be after started_at")
		return
	}

	// The track's mapping applies to the MIDI file's notes. The user's active
	// kit profile, if any, applies to the recorded hits; without one, hits use
	// the track's mapping as well (the behaviour before kit profiles existed).
	// Errors fall back to the General MIDI defaults.
	trackMap, _ := dbpkg.GetTrackMapping(h.db, req.TrackID)
	targetGroup := func(n int) string {
		if g, ok := trackMap[n]; ok {
			return g
		}
		return midipkg.InstrumentGroup(n)
	}
	hitGroup := targetGroup
	if kitMap, ok, err := dbpkg.ActiveKitMapping(h.db, user.ID); err == nil && ok {
		hitGroup = func(n int) string {
			if g, ok := kitMap[n]; ok {
				return g
			}
			return midipkg.InstrumentGroup(n)
		}
	}

	// Run the scoring algorithm.
	score := scoring.ScoreWithGroups(targetNotes, req.Hits, targetGroup, hitGroup)

	session := &models.Session{
		TrackID:    req.TrackID,
		StartedAt:  startedAt,
		EndedAt:    endedAt,
		DurationMs: float64(endedAt.Sub(startedAt).Milliseconds()),
		HitCount:   len(req.Hits),
	}

	// Instrument of each recorded note number, for annotating the raw hits.
	instrMap := make(map[string]string)
	for _, hit := range req.Hits {
		instrMap[strconv.Itoa(hit.NoteNumber)] = hitGroup(hit.NoteNumber)
	}

	sessionID, err := dbpkg.InsertSession(h.db, session, req.Hits, score, instrMap, user.ID)
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

const defaultSessionPageSize = 25
const maxSessionPageSize = 100

// GET /api/v1/sessions
func (h *sessionsHandler) list(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())

	trackID := int64(0)
	if s := r.URL.Query().Get("track_id"); s != "" {
		var err error
		trackID, err = strconv.ParseInt(s, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid track_id")
			return
		}
	}

	limit := defaultSessionPageSize
	if s := r.URL.Query().Get("limit"); s != "" {
		n, err := strconv.Atoi(s)
		if err != nil || n < 1 || n > maxSessionPageSize {
			writeError(w, http.StatusBadRequest, "limit must be between 1 and 100")
			return
		}
		limit = n
	}

	before := int64(0)
	if s := r.URL.Query().Get("before"); s != "" {
		var err error
		before, err = strconv.ParseInt(s, 10, 64)
		if err != nil || before < 0 {
			writeError(w, http.StatusBadRequest, "invalid before cursor")
			return
		}
	}

	sessions, hasMore, err := dbpkg.ListSessions(h.db, user.ID, trackID, limit, before)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list sessions")
		return
	}
	if sessions == nil {
		sessions = []models.SessionSummary{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"sessions": sessions, "has_more": hasMore})
}

// GET /api/v1/sessions/{id}
func (h *sessionsHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	user := UserFromContext(r.Context())
	session, err := dbpkg.GetSession(h.db, id, user.ID)
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

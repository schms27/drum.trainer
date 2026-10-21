package api

import (
	"database/sql"
	"errors"
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"
	dbpkg "github.com/simu/drum-trainer/backend/db"
	midipkg "github.com/simu/drum-trainer/backend/midi"
	"github.com/simu/drum-trainer/backend/models"
)

const (
	// maxEditorNotes bounds an edited track; uploads with more notes are rare.
	maxEditorNotes = 50000
	// maxTrackDurationMs is the longest track the editor accepts (2 hours).
	maxTrackDurationMs = 2 * 60 * 60 * 1000.0
)

type editedNote struct {
	NoteNumber int     `json:"note_number"`
	TimeMs     float64 `json:"time_ms"`
	Velocity   int     `json:"velocity"`
}

// PUT /api/v1/tracks/{id}/notes
// Replaces a track's notes with an edited set. The notes' instruments follow
// the track's note mapping. duration_ms is optional: it defaults to the
// current length, extended to the last note if that is later.
func (h *tracksHandler) replaceNotes(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		Notes      []editedNote `json:"notes"`
		DurationMs *float64     `json:"duration_ms"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if msg := validateEditedNotes(body.Notes, body.DurationMs); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	user := UserFromContext(r.Context())
	track, err := dbpkg.GetTrack(h.db, id, user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "track not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to fetch track")
		}
		return
	}
	if track.Disabled {
		writeError(w, http.StatusConflict, "this track was removed after a report and can't be edited")
		return
	}

	mapping, err := dbpkg.GetTrackMapping(h.db, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load note mapping")
		return
	}
	notes := make([]models.NoteEvent, len(body.Notes))
	last := 0.0
	for i, n := range body.Notes {
		instr, ok := mapping[n.NoteNumber]
		if !ok {
			instr = midipkg.InstrumentGroup(n.NoteNumber)
		}
		notes[i] = models.NoteEvent{NoteNumber: n.NoteNumber, TimeMs: n.TimeMs, Velocity: n.Velocity, Instrument: instr}
		last = max(last, n.TimeMs)
	}
	sort.SliceStable(notes, func(a, b int) bool { return notes[a].TimeMs < notes[b].TimeMs })

	duration := max(track.DurationMs, last)
	if body.DurationMs != nil {
		duration = max(*body.DurationMs, last)
	}

	if _, err := dbpkg.ReplaceTrackNotes(h.db, id, user.ID, notes, duration); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save notes")
		return
	}
	updated, err := dbpkg.GetTrack(h.db, id, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch track")
		return
	}
	writeJSON(w, http.StatusOK, models.TrackWithNotes{Track: *updated, Notes: notes})
}

// validateEditedNotes returns a message for the client when the notes or the
// requested duration are out of range.
func validateEditedNotes(notes []editedNote, durationMs *float64) string {
	if len(notes) > maxEditorNotes {
		return "too many notes"
	}
	if durationMs != nil && (*durationMs <= 0 || *durationMs > maxTrackDurationMs) {
		return "duration_ms must be between 0 and 2 hours"
	}
	for _, n := range notes {
		switch {
		case n.NoteNumber < 0 || n.NoteNumber > 127:
			return "note_number must be between 0 and 127"
		case n.Velocity < 1 || n.Velocity > 127:
			return "velocity must be between 1 and 127"
		case n.TimeMs < 0 || n.TimeMs > maxTrackDurationMs:
			return "time_ms must be between 0 and 2 hours"
		}
	}
	return ""
}

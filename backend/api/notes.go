package api

import (
	"database/sql"
	"errors"
	"net/http"
	"sort"
	"strings"
	"unicode/utf8"

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

// maxTrackNameLength bounds the name of a track created in the editor.
const maxTrackNameLength = 100

// POST /api/v1/tracks/blank
// Creates a track from scratch for the editor: no uploaded file, just a name,
// tempo, meter, length in bars and optional starting notes. The active kit
// profile's pad mapping becomes the track's note mapping, so notes the kit
// sends are grouped the same way.
func (h *tracksHandler) createBlank(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name             string       `json:"name"`
		BPM              float64      `json:"bpm"`
		TimeSigNumerator int          `json:"time_sig_numerator"`
		Bars             int          `json:"bars"`
		Notes            []editedNote `json:"notes"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	switch {
	case body.Name == "" || utf8.RuneCountInString(body.Name) > maxTrackNameLength:
		writeError(w, http.StatusBadRequest, "name must be 1–100 characters")
		return
	case body.BPM < minPracticeBPM || body.BPM > maxPracticeBPM:
		writeError(w, http.StatusBadRequest, "bpm must be between 20 and 300")
		return
	case body.TimeSigNumerator < 1 || body.TimeSigNumerator > 16:
		writeError(w, http.StatusBadRequest, "time_sig_numerator must be between 1 and 16")
		return
	case body.Bars < 1 || body.Bars > 999:
		writeError(w, http.StatusBadRequest, "bars must be between 1 and 999")
		return
	}
	duration := float64(body.Bars*body.TimeSigNumerator) * 60000 / body.BPM
	if duration > maxTrackDurationMs {
		writeError(w, http.StatusBadRequest, "the track would be longer than 2 hours")
		return
	}
	if msg := validateEditedNotes(body.Notes, nil); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	user := UserFromContext(r.Context())
	kitMap, _, err := dbpkg.ActiveKitMapping(h.db, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load kit profile")
		return
	}
	notes := make([]models.NoteEvent, 0, len(body.Notes))
	for _, n := range body.Notes {
		if n.TimeMs >= duration {
			continue // past the end
		}
		instr, ok := kitMap[n.NoteNumber]
		if !ok {
			instr = midipkg.InstrumentGroup(n.NoteNumber)
		}
		notes = append(notes, models.NoteEvent{NoteNumber: n.NoteNumber, TimeMs: n.TimeMs, Velocity: n.Velocity, Instrument: instr})
	}
	sort.SliceStable(notes, func(a, b int) bool { return notes[a].TimeMs < notes[b].TimeMs })

	track := &models.Track{
		Name:             body.Name,
		DurationMs:       duration,
		NoteCount:        len(notes),
		PPQ:              midipkg.ExportPPQ,
		BPM:              body.BPM,
		TimeSigNumerator: body.TimeSigNumerator,
		Origin:           models.TrackOriginEditor,
	}
	if err := dbpkg.InsertTrack(h.db, track, notes, user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save track")
		return
	}
	if len(kitMap) > 0 {
		mappings := make([]models.NoteMapping, 0, len(kitMap))
		for n, instr := range kitMap {
			mappings = append(mappings, models.NoteMapping{NoteNumber: n, Instrument: instr})
		}
		if err := dbpkg.SetTrackMapping(h.db, track.ID, mappings); err != nil {
			writeError(w, http.StatusInternalServerError, "failed to save note mapping")
			return
		}
	}
	track.Stats = dbpkg.EmptyTrackStats()
	writeJSON(w, http.StatusCreated, models.TrackWithNotes{Track: *track, Notes: notes})
}

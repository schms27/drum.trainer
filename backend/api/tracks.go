package api

import (
	"database/sql"
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	dbpkg "github.com/simu/drum-trainer/backend/db"
	midipkg "github.com/simu/drum-trainer/backend/midi"
	"github.com/simu/drum-trainer/backend/models"
	"github.com/simu/drum-trainer/backend/storage"
)

type tracksHandler struct {
	db       *dbpkg.DB
	store    *storage.Store
	maxBytes int64
}

// POST /api/v1/tracks
func (h *tracksHandler) create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBytes)

	if err := r.ParseMultipartForm(multipartMemoryBytes); err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "file too large or invalid form")
		return
	}
	defer r.MultipartForm.RemoveAll()

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing 'file' field")
		return
	}
	defer file.Close()

	// Read file bytes. A single Read is not guaranteed to fill the buffer, and
	// header.Size is client-supplied, so read until EOF instead.
	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read uploaded file")
		return
	}

	// Optional channel parameter: 1–16 (display-numbered), or "auto" to detect
	// the drum part. Defaults to 10 (GM drums).
	drumChannel := midipkg.GMDrumChannel // 0-indexed
	switch ch := r.FormValue("channel"); ch {
	case "":
	case "auto":
		detected, err := midipkg.DetectDrumChannel(data)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		drumChannel = detected
	default:
		n, err := strconv.Atoi(ch)
		if err != nil || n < 1 || n > 16 {
			writeError(w, http.StatusBadRequest, "channel must be between 1 and 16, or auto")
			return
		}
		drumChannel = uint8(n - 1)
	}

	// Parse MIDI.
	parsed, err := midipkg.ParseMIDIFile(data, header.Filename, drumChannel)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Save to disk.
	storedName := storage.UniqueFilename(header.Filename)
	if err := h.store.Save(storedName, data); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save file")
		return
	}

	user := UserFromContext(r.Context())

	track := &models.Track{
		Name:             parsed.Name,
		Filename:         header.Filename,
		FilePath:         storedName,
		FileSize:         int64(len(data)),
		DurationMs:       parsed.DurationMs,
		NoteCount:        len(parsed.Notes),
		PPQ:              parsed.PPQ,
		BPM:              parsed.BPM,
		TimeSigNumerator: parsed.TimeSigNumerator,
	}

	if err := dbpkg.InsertTrack(h.db, track, parsed.Notes, user.ID); err != nil {
		// Best-effort cleanup.
		h.store.Delete(storedName)
		writeError(w, http.StatusInternalServerError, "failed to save track")
		return
	}

	track.Stats = dbpkg.EmptyTrackStats()
	writeJSON(w, http.StatusCreated, models.TrackWithNotes{
		Track: *track,
		Notes: parsed.Notes,
	})
}

// GET /api/v1/tracks
func (h *tracksHandler) list(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	tracks, err := dbpkg.ListTracks(h.db, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tracks")
		return
	}
	if tracks == nil {
		tracks = []models.Track{}
	}
	stats, err := dbpkg.ListTrackStats(h.db, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load practice stats")
		return
	}
	for i := range tracks {
		tracks[i].Stats = statsOrEmpty(stats, tracks[i].ID)
	}
	writeJSON(w, http.StatusOK, map[string]any{"tracks": tracks})
}

// GET /api/v1/tracks/{id}
func (h *tracksHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
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

	notes := []models.NoteEvent{}
	// A track removed by moderation keeps its page and history, not its notes.
	if !track.Disabled {
		notes, err = dbpkg.GetTrackNotes(h.db, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to fetch notes")
			return
		}
	}

	stats, err := dbpkg.ListTrackStats(h.db, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load practice stats")
		return
	}
	track.Stats = statsOrEmpty(stats, id)

	writeJSON(w, http.StatusOK, models.TrackWithNotes{Track: *track, Notes: notes})
}

// DELETE /api/v1/tracks/{id}
func (h *tracksHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
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

	// Collect pilot file paths before the DB delete cascades the rows away.
	pilotPaths, _ := dbpkg.ListPilotFilePaths(h.db, id)

	if err := dbpkg.DeleteTrack(h.db, id, user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete track")
		return
	}

	// Best-effort file deletion — MIDI file and all pilot audio files.
	h.store.Delete(track.FilePath)
	for _, p := range pilotPaths {
		h.store.Delete(p)
	}

	w.WriteHeader(http.StatusNoContent)
}

// PATCH /api/v1/tracks/{id}/practice-bpm
func (h *tracksHandler) setPracticeBPM(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	var body struct {
		BPM *float64 `json:"bpm"` // null clears the override
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.BPM != nil && (*body.BPM < 20 || *body.BPM > 300) {
		writeError(w, http.StatusBadRequest, "bpm must be between 20 and 300")
		return
	}

	user := UserFromContext(r.Context())
	if err := dbpkg.SetPracticeBPM(h.db, id, user.ID, body.BPM); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "track not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to save practice BPM")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /api/v1/tracks/{id}/mapping
func (h *tracksHandler) getMapping(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	user := UserFromContext(r.Context())
	if _, err := dbpkg.GetTrack(h.db, id, user.ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "track not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to fetch track")
		}
		return
	}

	mappings, err := buildMappings(h.db, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"mappings": mappings})
}

// PUT /api/v1/tracks/{id}/mapping
func (h *tracksHandler) updateMapping(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	user := UserFromContext(r.Context())
	if _, err := dbpkg.GetTrack(h.db, id, user.ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "track not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to fetch track")
		}
		return
	}

	var body struct {
		Mappings []models.NoteMapping `json:"mappings"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	for _, m := range body.Mappings {
		if !midipkg.IsValidInstrument(m.Instrument) {
			writeError(w, http.StatusBadRequest, "invalid instrument: "+m.Instrument)
			return
		}
	}

	if err := dbpkg.SetTrackMapping(h.db, id, body.Mappings); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save mapping")
		return
	}

	mappings, err := buildMappings(h.db, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"mappings": mappings})
}

// buildMappings returns the effective note→instrument list for a track,
// merging default GM groups with any custom overrides.
func buildMappings(db *dbpkg.DB, trackID int64) ([]models.NoteMapping, error) {
	noteNumbers, err := dbpkg.GetTrackNoteNumbers(db, trackID)
	if err != nil {
		return nil, err
	}
	customMap, err := dbpkg.GetTrackMapping(db, trackID)
	if err != nil {
		return nil, err
	}
	mappings := make([]models.NoteMapping, 0, len(noteNumbers))
	for _, note := range noteNumbers {
		instr := midipkg.InstrumentGroup(note)
		if override, ok := customMap[note]; ok {
			instr = override
		}
		mappings = append(mappings, models.NoteMapping{NoteNumber: note, Instrument: instr})
	}
	return mappings, nil
}

// statsOrEmpty returns the track's stats, or empty stats if it has no takes.
func statsOrEmpty(stats map[int64]*models.TrackStats, trackID int64) *models.TrackStats {
	if st, ok := stats[trackID]; ok {
		return st
	}
	return dbpkg.EmptyTrackStats()
}

func parseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

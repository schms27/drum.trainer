package api

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	dbpkg "github.com/simu/drum-trainer/backend/db"
	midipkg "github.com/simu/drum-trainer/backend/midi"
	"github.com/simu/drum-trainer/backend/models"
	"github.com/simu/drum-trainer/backend/storage"
)

type tracksHandler struct {
	db       *sql.DB
	store    *storage.Store
	maxBytes int64
}

// POST /api/v1/tracks
func (h *tracksHandler) create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBytes)

	if err := r.ParseMultipartForm(h.maxBytes); err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "file too large or invalid form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing 'file' field")
		return
	}
	defer file.Close()

	// Read file bytes.
	data := make([]byte, header.Size)
	if _, err := file.Read(data); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read uploaded file")
		return
	}

	// Parse MIDI.
	parsed, err := midipkg.ParseMIDIFile(data, header.Filename)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Save to disk.
	storedName := storage.UniqueFilename(header.Filename)
	filePath, err := h.store.Save(storedName, data)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save file")
		return
	}

	track := &models.Track{
		Name:       parsed.Name,
		Filename:   header.Filename,
		FilePath:   filePath,
		FileSize:   header.Size,
		DurationMs: parsed.DurationMs,
		NoteCount:  len(parsed.Notes),
		PPQ:        parsed.PPQ,
	}

	trackID, err := dbpkg.InsertTrack(h.db, track, parsed.Notes)
	if err != nil {
		// Best-effort cleanup.
		h.store.Delete(filePath)
		writeError(w, http.StatusInternalServerError, "failed to save track")
		return
	}

	track.ID = trackID
	writeJSON(w, http.StatusCreated, models.TrackWithNotes{
		Track: *track,
		Notes: parsed.Notes,
	})
}

// GET /api/v1/tracks
func (h *tracksHandler) list(w http.ResponseWriter, r *http.Request) {
	tracks, err := dbpkg.ListTracks(h.db)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tracks")
		return
	}
	if tracks == nil {
		tracks = []models.Track{}
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

	track, err := dbpkg.GetTrack(h.db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "track not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to fetch track")
		}
		return
	}

	notes, err := dbpkg.GetTrackNotes(h.db, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch notes")
		return
	}

	writeJSON(w, http.StatusOK, models.TrackWithNotes{Track: *track, Notes: notes})
}

// DELETE /api/v1/tracks/{id}
func (h *tracksHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	track, err := dbpkg.GetTrack(h.db, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "track not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to fetch track")
		}
		return
	}

	if err := dbpkg.DeleteTrack(h.db, id); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete track")
		return
	}

	// Best-effort file deletion.
	h.store.Delete(track.FilePath)

	w.WriteHeader(http.StatusNoContent)
}

func parseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

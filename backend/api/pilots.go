package api

import (
	"bytes"
	"database/sql"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	dbpkg "github.com/simu/drum-trainer/backend/db"
	"github.com/simu/drum-trainer/backend/models"
	"github.com/simu/drum-trainer/backend/storage"
)

type pilotsHandler struct {
	db       *dbpkg.DB
	store    *storage.Store
	maxBytes int64
}

// POST /api/v1/tracks/{id}/pilots
func (h *pilotsHandler) create(w http.ResponseWriter, r *http.Request) {
	trackID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid track id")
		return
	}

	user := UserFromContext(r.Context())

	// Verify the track exists and belongs to this user before accepting the upload.
	if _, err := dbpkg.GetTrack(h.db, trackID, user.ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "track not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to fetch track")
		}
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.maxBytes)
	if err := r.ParseMultipartForm(multipartMemoryBytes); err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "file too large or invalid form")
		return
	}
	defer r.MultipartForm.RemoveAll()

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "missing 'file' field")
		return
	}
	defer file.Close()

	// Determine the MIME type from the file content, not the client header.
	head, err := readHead(file, 512)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read uploaded file")
		return
	}
	mimeType := sniffAudioMIME(head)
	if mimeType == "" {
		writeError(w, http.StatusBadRequest, "unsupported file type — upload MP3, WAV, OGG, FLAC, AAC, or M4A")
		return
	}

	// Stream to disk rather than buffering up to MAX_PILOT_SIZE_MB in memory.
	storedName := storage.UniqueFilename(header.Filename)
	size, err := h.store.SaveStream(storedName, io.MultiReader(bytes.NewReader(head), file))
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save file")
		return
	}

	pilot := &models.PilotTrack{
		TrackID:  trackID,
		Name:     name,
		Filename: header.Filename,
		FilePath: storedName,
		FileSize: size,
		MimeType: mimeType,
	}

	id, err := dbpkg.InsertPilotTrack(h.db, pilot, user.ID)
	if err != nil {
		h.store.Delete(storedName)
		writeError(w, http.StatusInternalServerError, "failed to save pilot track")
		return
	}

	pilot.ID = id
	writeJSON(w, http.StatusCreated, pilot)
}

// GET /api/v1/tracks/{id}/pilots
func (h *pilotsHandler) list(w http.ResponseWriter, r *http.Request) {
	trackID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid track id")
		return
	}

	user := UserFromContext(r.Context())

	if _, err := dbpkg.GetTrack(h.db, trackID, user.ID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "track not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to fetch track")
		}
		return
	}

	pilots, err := dbpkg.ListPilotTracks(h.db, trackID, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list pilot tracks")
		return
	}
	if pilots == nil {
		pilots = []models.PilotTrack{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"pilots": pilots})
}

// DELETE /api/v1/tracks/{id}/pilots/{pilotId}
func (h *pilotsHandler) delete(w http.ResponseWriter, r *http.Request) {
	trackID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid track id")
		return
	}
	pilotID, err := parseID(chi.URLParam(r, "pilotId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid pilot id")
		return
	}

	user := UserFromContext(r.Context())

	pilot, err := dbpkg.GetPilotTrack(h.db, pilotID, trackID, user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "pilot track not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to fetch pilot track")
		}
		return
	}

	if err := dbpkg.DeletePilotTrack(h.db, pilotID, trackID, user.ID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete pilot track")
		return
	}

	h.store.Delete(pilot.FilePath)

	w.WriteHeader(http.StatusNoContent)
}

// GET /api/v1/tracks/{id}/pilots/{pilotId}/audio
// Streams the audio file with support for range requests (browser seeking).
func (h *pilotsHandler) audio(w http.ResponseWriter, r *http.Request) {
	trackID, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid track id")
		return
	}
	pilotID, err := parseID(chi.URLParam(r, "pilotId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid pilot id")
		return
	}

	user := UserFromContext(r.Context())

	pilot, err := dbpkg.GetPilotTrack(h.db, pilotID, trackID, user.ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "pilot track not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to fetch pilot track")
		}
		return
	}

	f, err := h.store.Open(pilot.FilePath)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to open audio file")
		return
	}
	defer f.Close()

	// Set Content-Type explicitly before ServeContent so it is not sniffed.
	w.Header().Set("Content-Type", pilot.MimeType)

	// http.ServeContent handles Range, If-Modified-Since, and Content-Length.
	http.ServeContent(w, r, pilot.Filename, time.Time{}, f)
}

package api

import (
	"database/sql"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	dbpkg "github.com/simu/drum-trainer/backend/db"
	"github.com/simu/drum-trainer/backend/models"
	"github.com/simu/drum-trainer/backend/storage"
)

// allowedAudioMIME is the set of MIME types accepted for pilot track uploads.
var allowedAudioMIME = map[string]bool{
	"audio/mpeg":  true, // .mp3
	"audio/wav":   true, // .wav
	"audio/wave":  true, // .wav (alternate)
	"audio/ogg":   true, // .ogg
	"audio/flac":  true, // .flac
	"audio/aac":   true, // .aac
	"audio/x-m4a": true, // .m4a
	"audio/mp4":   true, // .m4a / .mp4 audio
}

type pilotsHandler struct {
	db       *sql.DB
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
	if err := r.ParseMultipartForm(h.maxBytes); err != nil {
		writeError(w, http.StatusRequestEntityTooLarge, "file too large or invalid form")
		return
	}

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

	// Validate MIME type from the multipart Content-Type header.
	mimeType := header.Header.Get("Content-Type")
	// Strip any parameters (e.g. "audio/mpeg; charset=binary").
	if i := strings.Index(mimeType, ";"); i >= 0 {
		mimeType = strings.TrimSpace(mimeType[:i])
	}
	mimeType = strings.ToLower(mimeType)
	if !allowedAudioMIME[mimeType] {
		writeError(w, http.StatusBadRequest, "unsupported file type — upload MP3, WAV, OGG, FLAC, AAC, or M4A")
		return
	}

	data := make([]byte, header.Size)
	if _, err := file.Read(data); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to read uploaded file")
		return
	}

	storedName := storage.UniqueFilename(header.Filename)
	filePath, err := h.store.Save(storedName, data)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save file")
		return
	}

	pilot := &models.PilotTrack{
		TrackID:  trackID,
		Name:     name,
		Filename: header.Filename,
		FilePath: filePath,
		FileSize: header.Size,
		MimeType: mimeType,
	}

	id, err := dbpkg.InsertPilotTrack(h.db, pilot, user.ID)
	if err != nil {
		h.store.Delete(filePath)
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

	f, err := os.Open(pilot.FilePath)
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

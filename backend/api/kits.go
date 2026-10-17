package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	dbpkg "github.com/simu/drum-trainer/backend/db"
	midipkg "github.com/simu/drum-trainer/backend/midi"
	"github.com/simu/drum-trainer/backend/models"
)

const maxKitNameLength = 60

type kitsHandler struct {
	db *dbpkg.DB
}

type kitRequest struct {
	Name     string               `json:"name"`
	Mappings []models.NoteMapping `json:"mappings"`
}

// validate trims the name and checks the mappings; it returns a message for
// the client when something is wrong.
func (k *kitRequest) validate() string {
	k.Name = strings.TrimSpace(k.Name)
	if k.Name == "" {
		return "name is required"
	}
	if utf8.RuneCountInString(k.Name) > maxKitNameLength {
		return fmt.Sprintf("name must be at most %d characters", maxKitNameLength)
	}
	seen := make(map[int]bool, len(k.Mappings))
	for _, m := range k.Mappings {
		if m.NoteNumber < 0 || m.NoteNumber > 127 {
			return "note_number must be between 0 and 127"
		}
		if seen[m.NoteNumber] {
			return fmt.Sprintf("note %d is mapped twice", m.NoteNumber)
		}
		seen[m.NoteNumber] = true
		if !midipkg.IsValidInstrument(m.Instrument) {
			return "invalid instrument: " + m.Instrument
		}
	}
	return ""
}

// GET /api/v1/kits
func (h *kitsHandler) list(w http.ResponseWriter, r *http.Request) {
	user := UserFromContext(r.Context())
	kits, err := dbpkg.ListKitProfiles(h.db, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list kits")
		return
	}
	if kits == nil {
		kits = []models.KitProfile{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"kits": kits})
}

// POST /api/v1/kits
func (h *kitsHandler) create(w http.ResponseWriter, r *http.Request) {
	var req kitRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if msg := req.validate(); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	user := UserFromContext(r.Context())
	kit, err := dbpkg.CreateKitProfile(h.db, user.ID, req.Name, req.Mappings)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save kit")
		return
	}
	writeJSON(w, http.StatusCreated, kit)
}

// PUT /api/v1/kits/{id}
func (h *kitsHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req kitRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if msg := req.validate(); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	user := UserFromContext(r.Context())
	if err := dbpkg.UpdateKitProfile(h.db, id, user.ID, req.Name, req.Mappings); err != nil {
		writeKitError(w, err)
		return
	}
	kit, err := dbpkg.GetKitProfile(h.db, id, user.ID)
	if err != nil {
		writeKitError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, kit)
}

// DELETE /api/v1/kits/{id}
func (h *kitsHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	user := UserFromContext(r.Context())
	if err := dbpkg.DeleteKitProfile(h.db, id, user.ID); err != nil {
		writeKitError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PUT /api/v1/kits/active — body {"id": 3} or {"id": null} for none.
func (h *kitsHandler) setActive(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ID *int64 `json:"id"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	user := UserFromContext(r.Context())
	if err := dbpkg.SetActiveKitProfile(h.db, user.ID, body.ID); err != nil {
		writeKitError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeKitError(w http.ResponseWriter, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "kit not found")
		return
	}
	writeError(w, http.StatusInternalServerError, "failed to update kit")
}

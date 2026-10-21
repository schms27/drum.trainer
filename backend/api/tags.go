package api

import (
	"database/sql"
	"errors"
	"net/http"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	dbpkg "github.com/simu/drum-trainer/backend/db"
	"github.com/simu/drum-trainer/backend/models"
)

const maxTagNameLength = 30

// cleanTagName trims a tag name and returns a message for the client when it
// is unusable.
func cleanTagName(name string) (string, string) {
	name = strings.Join(strings.Fields(name), " ")
	if name == "" || utf8.RuneCountInString(name) > maxTagNameLength {
		return "", "tag name must be 1–30 characters"
	}
	return name, ""
}

func validTagColor(c string) bool {
	return slices.Contains(models.TagColors, c)
}

func writeTagError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, dbpkg.ErrTagExists):
		writeError(w, http.StatusConflict, "you already have a tag with that name")
	case errors.Is(err, sql.ErrNoRows):
		writeError(w, http.StatusNotFound, "tag not found")
	default:
		writeError(w, http.StatusInternalServerError, "failed to save tag")
	}
}

// GET /api/v1/tags
func (h *tracksHandler) listTags(w http.ResponseWriter, r *http.Request) {
	tags, err := dbpkg.ListTags(h.db, UserFromContext(r.Context()).ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tags")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tags": tags, "colors": models.TagColors})
}

// POST /api/v1/tags
func (h *tracksHandler) createTag(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	name, msg := cleanTagName(body.Name)
	if msg == "" && !validTagColor(body.Color) {
		msg = "color must be one of " + strings.Join(models.TagColors, ", ")
	}
	if msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}
	tag, err := dbpkg.CreateTag(h.db, UserFromContext(r.Context()).ID, name, body.Color)
	if err != nil {
		writeTagError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, tag)
}

// PATCH /api/v1/tags/{id}
func (h *tracksHandler) updateTag(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		Name  *string `json:"name"`
		Color *string `json:"color"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if body.Name != nil {
		name, msg := cleanTagName(*body.Name)
		if msg != "" {
			writeError(w, http.StatusBadRequest, msg)
			return
		}
		body.Name = &name
	}
	if body.Color != nil && !validTagColor(*body.Color) {
		writeError(w, http.StatusBadRequest, "color must be one of "+strings.Join(models.TagColors, ", "))
		return
	}
	if err := dbpkg.UpdateTag(h.db, id, UserFromContext(r.Context()).ID, body.Name, body.Color); err != nil {
		writeTagError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// DELETE /api/v1/tags/{id}
func (h *tracksHandler) deleteTag(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := dbpkg.DeleteTag(h.db, id, UserFromContext(r.Context()).ID); err != nil {
		writeTagError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// PUT /api/v1/tracks/{id}/tags
// Replaces the tags on a track with the given ones (all the user's own).
func (h *tracksHandler) setTrackTags(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		TagIDs []int64 `json:"tag_ids"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if len(body.TagIDs) > 50 {
		writeError(w, http.StatusBadRequest, "too many tags")
		return
	}
	user := UserFromContext(r.Context())
	if err := dbpkg.SetTrackTags(h.db, id, user.ID, body.TagIDs); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "track or tag not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to save tags")
		}
		return
	}
	tags, err := dbpkg.TrackTags(h.db, user.ID, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to load tags")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tags": orEmptyTags(tags[id])})
}

// PATCH /api/v1/tracks/{id}
// Renames a track.
func (h *tracksHandler) rename(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" || utf8.RuneCountInString(name) > maxTrackNameLength {
		writeError(w, http.StatusBadRequest, "name must be 1–100 characters")
		return
	}
	if err := dbpkg.RenameTrack(h.db, id, UserFromContext(r.Context()).ID, name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "track not found")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to rename track")
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"name": name})
}

func orEmptyTags(tags []models.TagRef) []models.TagRef {
	if tags == nil {
		return []models.TagRef{}
	}
	return tags
}

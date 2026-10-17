package api

import (
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"net"
	"net/http"
	"slices"

	"github.com/go-chi/chi/v5"
	dbpkg "github.com/simu/drum-trainer/backend/db"
	"github.com/simu/drum-trainer/backend/models"
)

// TermsVersion identifies the terms of use an owner accepts when sharing.
// Bump it whenever frontend/src/routes/terms changes materially.
const TermsVersion = "2026-10-17"

// previewBars is how many bars of notes the public share page may show.
const previewBars = 4

type sharingHandler struct {
	db     *dbpkg.DB
	secret string // keys the IP hash in the share audit log
}

// PUT /api/v1/tracks/{id}/sharing
func (h *sharingHandler) share(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		RightsBasis string `json:"rights_basis"`
		AcceptTerms bool   `json:"accept_terms"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if !slices.Contains(models.RightsBases, body.RightsBasis) {
		writeError(w, http.StatusBadRequest, "rights_basis must be one of original, licensed, public_domain, transcription")
		return
	}
	if !body.AcceptTerms {
		writeError(w, http.StatusBadRequest, "you must confirm you have the right to share this track")
		return
	}

	user := UserFromContext(r.Context())
	if user.SharingBanned {
		writeError(w, http.StatusForbidden, "sharing is turned off for your account after repeated copyright reports")
		return
	}
	token, err := dbpkg.ShareTrack(h.db, id, user.ID, dbpkg.ShareEvent{
		RightsBasis: body.RightsBasis,
		TermsVer:    TermsVersion,
		IPHash:      ipHash(h.secret, r),
	})
	if err != nil {
		writeShareError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"visibility":   models.VisibilityUnlisted,
		"share_token":  token,
		"rights_basis": body.RightsBasis,
	})
}

// POST /api/v1/tracks/{id}/sharing/rotate
func (h *sharingHandler) rotate(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	user := UserFromContext(r.Context())
	token, err := dbpkg.RotateShareToken(h.db, id, user.ID, dbpkg.ShareEvent{TermsVer: TermsVersion, IPHash: ipHash(h.secret, r)})
	if err != nil {
		writeShareError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"visibility": models.VisibilityUnlisted, "share_token": token})
}

// DELETE /api/v1/tracks/{id}/sharing
func (h *sharingHandler) unshare(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	user := UserFromContext(r.Context())
	if err := dbpkg.UnshareTrack(h.db, id, user.ID, dbpkg.ShareEvent{IPHash: ipHash(h.secret, r)}); err != nil {
		writeShareError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeShareError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		writeError(w, http.StatusNotFound, "track not found")
	case errors.Is(err, dbpkg.ErrCopyNotShareable), errors.Is(err, dbpkg.ErrNotShared),
		errors.Is(err, dbpkg.ErrTrackDisabled), errors.Is(err, dbpkg.ErrBlockedContent):
		writeError(w, http.StatusConflict, err.Error())
	default:
		writeError(w, http.StatusInternalServerError, "failed to update sharing")
	}
}

// GET /api/v1/shared/{token} — no session required.
//
// Returns metadata and the opening bars only; the full note list is only
// available by copying the track into a signed-in user's library.
func (h *sharingHandler) get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("X-Robots-Tag", "noindex, nofollow")
	w.Header().Set("Cache-Control", "private, no-store")

	track, ref, err := dbpkg.GetSharedTrack(h.db, chi.URLParam(r, "token"))
	if err != nil {
		writeSharedLookupError(w, err)
		return
	}
	notes, err := dbpkg.GetTrackNotes(h.db, track.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch notes")
		return
	}
	custom, err := dbpkg.GetTrackMapping(h.db, track.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to fetch mapping")
		return
	}

	out := models.SharedTrack{
		Name:             track.Name,
		DurationMs:       track.DurationMs,
		NoteCount:        track.NoteCount,
		BPM:              track.BPM,
		TimeSigNumerator: track.TimeSigNumerator,
		SharedAt:         ref.SharedAt.T,
		Instruments:      map[string]int{},
		PreviewNotes:     []models.NoteEvent{},
		PreviewMs:        previewMs(track),
	}
	if track.RightsBasis != nil {
		out.RightsBasis = *track.RightsBasis
	}
	for _, n := range notes {
		if instr, ok := custom[n.NoteNumber]; ok {
			n.Instrument = instr
		}
		out.Instruments[n.Instrument]++
		if n.TimeMs < out.PreviewMs {
			out.PreviewNotes = append(out.PreviewNotes, n)
		}
	}

	if viewer := optionalUser(h.db, r); viewer != nil {
		out.IsOwner = viewer.ID == ref.OwnerID
		if copyID, err := dbpkg.FindCopy(h.db, track.ID, viewer.ID); err == nil {
			out.CopiedTrackID = &copyID
		}
	}
	writeJSON(w, http.StatusOK, out)
}

// POST /api/v1/shared/{token}/copy
func (h *sharingHandler) copy(w http.ResponseWriter, r *http.Request) {
	track, ref, err := dbpkg.GetSharedTrack(h.db, chi.URLParam(r, "token"))
	if err != nil {
		writeSharedLookupError(w, err)
		return
	}
	user := UserFromContext(r.Context())
	if user.ID == ref.OwnerID {
		// Nothing to copy: the owner already has the track.
		writeJSON(w, http.StatusOK, map[string]any{"track_id": track.ID})
		return
	}
	id, err := dbpkg.CopyTrack(h.db, track.ID, user.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to copy track")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"track_id": id})
}

func writeSharedLookupError(w http.ResponseWriter, err error) {
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "shared track not found")
	} else {
		writeError(w, http.StatusInternalServerError, "failed to fetch shared track")
	}
}

// previewMs is the length of the opening bars shown on the share page.
func previewMs(t *models.Track) float64 {
	bpm, beats := t.BPM, t.TimeSigNumerator
	if bpm <= 0 {
		bpm = 120
	}
	if beats <= 0 {
		beats = 4
	}
	return min(float64(previewBars*beats)*60000/bpm, t.DurationMs)
}

// optionalUser returns the signed-in user, or nil for anonymous requests.
func optionalUser(db *dbpkg.DB, r *http.Request) *models.User {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return nil
	}
	user, err := dbpkg.ValidateSession(db, cookie.Value)
	if err != nil {
		return nil
	}
	return user
}

// ipHash returns a keyed hash of the client IP for audit logs, so they can
// link actions from one address without storing the address itself.
func ipHash(secret string, r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(host))
	return hex.EncodeToString(mac.Sum(nil))[:32]
}

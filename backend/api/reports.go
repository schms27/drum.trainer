package api

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"net/mail"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"
	dbpkg "github.com/simu/drum-trainer/backend/db"
	"github.com/simu/drum-trainer/backend/models"
)

type reportsHandler struct {
	db     *dbpkg.DB
	secret string // keys the IP hash stored with each report
}

// POST /api/v1/reports — no session required, so rights holders without an
// account can file a notice.
func (h *reportsHandler) create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token          string `json:"token"`
		Kind           string `json:"kind"`
		ReporterName   string `json:"reporter_name"`
		ReporterEmail  string `json:"reporter_email"`
		WorkIdentified string `json:"work_identified"`
		Details        string `json:"details"`
		GoodFaith      bool   `json:"good_faith"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	body.ReporterName = strings.TrimSpace(body.ReporterName)
	body.ReporterEmail = strings.TrimSpace(body.ReporterEmail)
	body.WorkIdentified = strings.TrimSpace(body.WorkIdentified)
	body.Details = strings.TrimSpace(body.Details)
	if msg := validateReport(body.Kind, body.ReporterName, body.ReporterEmail, body.WorkIdentified, body.Details, body.GoodFaith); msg != "" {
		writeError(w, http.StatusBadRequest, msg)
		return
	}

	track, ref, err := dbpkg.GetSharedTrack(h.db, extractToken(body.Token))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "no shared track found for that link — it may already have been turned off")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to look up the shared track")
		}
		return
	}

	report := &models.Report{
		TrackID:        track.ID,
		TrackName:      track.Name,
		OwnerUserID:    &ref.OwnerID,
		ShareToken:     *track.ShareToken,
		ReporterName:   body.ReporterName,
		ReporterEmail:  body.ReporterEmail,
		Kind:           body.Kind,
		WorkIdentified: body.WorkIdentified,
		Details:        body.Details,
		GoodFaith:      body.GoodFaith,
	}
	if viewer := optionalUser(h.db, r); viewer != nil {
		report.ReporterUserID = &viewer.ID
	}
	if err := dbpkg.CreateReport(h.db, report, ipHash(h.secret, r)); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to save report")
		return
	}
	log.Printf("report %d (%s) filed against track %d", report.ID, report.Kind, report.TrackID)
	writeJSON(w, http.StatusCreated, map[string]any{"id": report.ID})
}

// validateReport returns a user-facing message for an invalid report, or "".
func validateReport(kind, name, email, work, details string, goodFaith bool) string {
	switch {
	case !slices.Contains(models.ReportKinds, kind):
		return "kind must be copyright, abuse or other"
	case name == "" || utf8.RuneCountInString(name) > 200:
		return "please give your name (up to 200 characters)"
	case len(email) > 254 || !validEmail(email):
		return "please give a valid email address so we can reach you"
	case utf8.RuneCountInString(details) < 10 || utf8.RuneCountInString(details) > 5000:
		return "please describe the problem (10 to 5000 characters)"
	case utf8.RuneCountInString(work) > 500:
		return "the work description is too long (up to 500 characters)"
	case kind == models.ReportCopyright && work == "":
		return "please name the copyrighted work you believe is infringed"
	case kind == models.ReportCopyright && !goodFaith:
		return "please confirm the good-faith statement"
	}
	return ""
}

func validEmail(s string) bool {
	addr, err := mail.ParseAddress(s)
	return err == nil && addr.Address == s && strings.Contains(s[strings.LastIndexByte(s, '@')+1:], ".")
}

// extractToken accepts either a bare share token or a full share link.
func extractToken(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.LastIndex(s, "/s/"); i >= 0 {
		s = s[i+3:]
	}
	if i := strings.IndexAny(s, "?#/"); i >= 0 {
		s = s[:i]
	}
	return s
}

type adminHandler struct {
	db *dbpkg.DB
}

// GET /api/v1/admin/reports?status=open|actioned|rejected (default open; "all" for every report)
func (h *adminHandler) listReports(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	switch status {
	case "":
		status = models.ReportOpen
	case "all":
		status = ""
	case models.ReportOpen, models.ReportActioned, models.ReportRejected:
	default:
		writeError(w, http.StatusBadRequest, "invalid status")
		return
	}
	reports, err := dbpkg.ListReports(h.db, status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list reports")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"reports": reports})
}

// POST /api/v1/admin/reports/{id}/reject
func (h *adminHandler) rejectReport(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		Note string `json:"note"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if err := dbpkg.RejectReport(h.db, id, strings.TrimSpace(body.Note)); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "no open report with that id")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to reject report")
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// POST /api/v1/admin/tracks/{id}/disable
func (h *adminHandler) disableTrack(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		Reason string `json:"reason"` // shown to the owner and to holders of copies
		Strike bool   `json:"strike"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	reason := strings.TrimSpace(body.Reason)
	if reason == "" || utf8.RuneCountInString(reason) > 500 {
		writeError(w, http.StatusBadRequest, "a reason (up to 500 characters) is required — the owner will see it")
		return
	}
	res, err := dbpkg.DisableTrackLineage(h.db, id, reason, body.Strike)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "neither the track nor any copy of it exists any more")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to remove track")
		}
		return
	}
	admin := UserFromContext(r.Context())
	log.Printf("moderation: %s removed track %d and its copies (%d tracks, strike=%v, owner strikes=%d, banned=%v)",
		admin.Email, id, res.TracksAffected, body.Strike, res.OwnerStrikes, res.OwnerBanned)
	writeJSON(w, http.StatusOK, res)
}

// POST /api/v1/admin/tracks/{id}/restore
func (h *adminHandler) restoreTrack(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		Note         string `json:"note"`
		RevokeStrike bool   `json:"revoke_strike"`
	}
	if err := decodeJSON(r, &body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	res, err := dbpkg.RestoreTrackLineage(h.db, id, strings.TrimSpace(body.Note), body.RevokeStrike)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeError(w, http.StatusNotFound, "no removed track with that id")
		} else {
			writeError(w, http.StatusInternalServerError, "failed to restore track")
		}
		return
	}
	admin := UserFromContext(r.Context())
	log.Printf("moderation: %s restored track %d (%d tracks, revoke_strike=%v)", admin.Email, id, res.TracksAffected, body.RevokeStrike)
	writeJSON(w, http.StatusOK, res)
}

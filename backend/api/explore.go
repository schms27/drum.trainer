package api

import (
	"net/http"
	"strconv"
	"unicode/utf8"

	dbpkg "github.com/simu/drum-trainer/backend/db"
)

// explorePageSize is how many tracks one Explore page holds.
const explorePageSize = 24

type exploreHandler struct {
	db *dbpkg.DB
}

// GET /api/v1/explore?q=&sort=newest|popular&page=1 — no session required.
func (h *exploreHandler) list(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	query := q.Get("q")
	if utf8.RuneCountInString(query) > 100 {
		writeError(w, http.StatusBadRequest, "search is limited to 100 characters")
		return
	}
	sort := q.Get("sort")
	switch sort {
	case "":
		sort = dbpkg.SortNewest
	case dbpkg.SortNewest, dbpkg.SortPopular:
	default:
		writeError(w, http.StatusBadRequest, "sort must be newest or popular")
		return
	}
	page := 1
	if p := q.Get("page"); p != "" {
		n, err := strconv.Atoi(p)
		if err != nil || n < 1 || n > 1000 {
			writeError(w, http.StatusBadRequest, "invalid page")
			return
		}
		page = n
	}

	tracks, more, err := dbpkg.ListPublicTracks(h.db, query, sort, explorePageSize, (page-1)*explorePageSize)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list tracks")
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=30")
	writeJSON(w, http.StatusOK, map[string]any{"tracks": tracks, "page": page, "has_more": more})
}

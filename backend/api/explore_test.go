package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/simu/drum-trainer/backend/models"
)

type explorePage struct {
	Tracks  []models.PublicTrack `json:"tracks"`
	Page    int                  `json:"page"`
	HasMore bool                 `json:"has_more"`
}

func TestPublicSharingAndExplore(t *testing.T) {
	e := newSharingEnv(t)
	sharing := "/api/v1/tracks/" + itoa(e.trackID) + "/sharing"

	if code := e.do("PUT", sharing, "owner", map[string]any{"visibility": "public", "rights_basis": "transcription", "accept_terms": true}, nil); code != http.StatusBadRequest {
		t.Fatalf("public transcription: got %d, want 400", code)
	}
	if code := e.do("PUT", sharing, "owner", map[string]any{"visibility": "everyone", "rights_basis": "original", "accept_terms": true}, nil); code != http.StatusBadRequest {
		t.Fatalf("bad visibility: got %d, want 400", code)
	}

	var shared struct {
		Visibility string `json:"visibility"`
		ShareToken string `json:"share_token"`
	}
	if code := e.do("PUT", sharing, "owner", map[string]any{"visibility": "public", "rights_basis": "original", "accept_terms": true}, &shared); code != http.StatusOK || shared.Visibility != "public" {
		t.Fatalf("public share: got %d, %+v", code, shared)
	}

	// Explore needs no session and shows the track.
	var page explorePage
	if code := e.do("GET", "/api/v1/explore", "", nil, &page); code != http.StatusOK {
		t.Fatalf("explore: got %d", code)
	}
	if len(page.Tracks) != 1 || page.Tracks[0].ShareToken != shared.ShareToken || page.HasMore {
		t.Fatalf("unexpected explore page: %+v", page)
	}
	e.do("GET", "/api/v1/explore?q=nothing-matches", "", nil, &page)
	if len(page.Tracks) != 0 {
		t.Fatalf("search should match nothing: %+v", page.Tracks)
	}
	for _, bad := range []string{"?sort=trending", "?page=0", "?page=x"} {
		if code := e.do("GET", "/api/v1/explore"+bad, "", nil, nil); code != http.StatusBadRequest {
			t.Errorf("explore%s: got %d, want 400", bad, code)
		}
	}

	// Public share pages may be indexed; unlisted ones may not.
	robots := func() string {
		req := httptest.NewRequest("GET", "/api/v1/shared/"+shared.ShareToken, nil)
		rec := httptest.NewRecorder()
		e.router.ServeHTTP(rec, req)
		return rec.Header().Get("X-Robots-Tag")
	}
	if got := robots(); got != "" {
		t.Errorf("public track has X-Robots-Tag %q", got)
	}

	// Rotating keeps it public; switching to unlisted drops it from Explore.
	var rotated struct {
		Visibility string `json:"visibility"`
	}
	e.do("POST", sharing+"/rotate", "owner", nil, &rotated)
	if rotated.Visibility != "public" {
		t.Fatalf("rotate reported visibility %q", rotated.Visibility)
	}
	e.do("PUT", sharing, "owner", map[string]any{"visibility": "unlisted", "rights_basis": "original", "accept_terms": true}, &shared)
	e.do("GET", "/api/v1/explore", "", nil, &page)
	if len(page.Tracks) != 0 {
		t.Fatalf("unlisted track still in explore: %+v", page.Tracks)
	}
	if got := robots(); got != "noindex, nofollow" {
		t.Errorf("unlisted track X-Robots-Tag = %q", got)
	}
}

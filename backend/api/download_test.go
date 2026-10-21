package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	midipkg "github.com/simu/drum-trainer/backend/midi"
)

// get sends a GET as the given user and returns the raw response.
func (e *sharingEnv) get(path, user string) *httptest.ResponseRecorder {
	e.t.Helper()
	req := httptest.NewRequest("GET", path, nil)
	if user != "" {
		req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: user})
	}
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, req)
	return rec
}

func TestDownloadOwnTrack(t *testing.T) {
	e := newSharingEnv(t)
	// The env's track points at "x.mid"; put the uploaded bytes there.
	original := []byte("MThd original upload")
	if err := e.store.Save("x.mid", original); err != nil {
		t.Fatal(err)
	}
	base := "/api/v1/tracks/" + itoa(e.trackID)

	rec := e.get(base+"/file", "owner")
	if rec.Code != http.StatusOK || rec.Body.String() != string(original) {
		t.Fatalf("owner download: got %d %q", rec.Code, rec.Body.String())
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, `filename=groove.mid`) {
		t.Fatalf("Content-Disposition = %q", cd)
	}

	rec = e.get(base+"/export.mid", "owner")
	if rec.Code != http.StatusOK {
		t.Fatalf("owner export: got %d", rec.Code)
	}
	parsed, err := midipkg.ParseMIDIFile(rec.Body.Bytes(), "export.mid", midipkg.GMDrumChannel)
	if err != nil || len(parsed.Notes) != 32 || parsed.BPM != 120 || parsed.Name != "groove" {
		t.Fatalf("export: err=%v %+v", err, parsed)
	}
	if cd := rec.Header().Get("Content-Disposition"); !strings.Contains(cd, "groove (drums).mid") {
		t.Fatalf("export Content-Disposition = %q", cd)
	}

	// Nobody else gets it, signed in or not.
	for _, p := range []string{"/file", "/export.mid"} {
		if code := e.get(base+p, "other").Code; code != http.StatusNotFound {
			t.Errorf("other user %s: got %d, want 404", p, code)
		}
		if code := e.get(base+p, "").Code; code != http.StatusUnauthorized {
			t.Errorf("anonymous %s: got %d, want 401", p, code)
		}
	}
}

func TestDownloadRefusedForCopiesAndRemovedTracks(t *testing.T) {
	e := newSharingEnv(t)
	if err := e.store.Save("x.mid", []byte("MThd")); err != nil {
		t.Fatal(err)
	}
	var shared struct {
		ShareToken string `json:"share_token"`
	}
	e.do("PUT", "/api/v1/tracks/"+itoa(e.trackID)+"/sharing", "owner", map[string]any{"rights_basis": "original", "accept_terms": true}, &shared)
	var copied struct {
		TrackID int64 `json:"track_id"`
	}
	if code := e.do("POST", "/api/v1/shared/"+shared.ShareToken+"/copy", "other", nil, &copied); code != http.StatusCreated {
		t.Fatalf("copy: got %d", code)
	}
	for _, p := range []string{"/file", "/export.mid"} {
		if code := e.get("/api/v1/tracks/"+itoa(copied.TrackID)+p, "other").Code; code != http.StatusNotFound {
			t.Errorf("copy %s: got %d, want 404", p, code)
		}
	}

	if code := e.do("POST", "/api/v1/admin/tracks/"+itoa(e.trackID)+"/disable", "admin", map[string]any{"reason": "notice"}, nil); code != http.StatusOK {
		t.Fatalf("disable: got %d", code)
	}
	for _, p := range []string{"/file", "/export.mid"} {
		if code := e.get("/api/v1/tracks/"+itoa(e.trackID)+p, "owner").Code; code != http.StatusNotFound {
			t.Errorf("removed track %s: got %d, want 404", p, code)
		}
	}
}

func TestExportFilename(t *testing.T) {
	for in, want := range map[string]string{
		"Groove":         "Groove (drums).mid",
		"song.mid":       "song (drums).mid",
		"Mr. Brightside": "Mr. Brightside (drums).mid",
		`AC/DC: "Back"?`: "AC_DC_ _Back__ (drums).mid",
		"   ":            "track (drums).mid",
	} {
		if got := exportFilename(in); got != want {
			t.Errorf("exportFilename(%q) = %q, want %q", in, got, want)
		}
	}
}

package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/simu/drum-trainer/backend/config"
	dbpkg "github.com/simu/drum-trainer/backend/db"
	"github.com/simu/drum-trainer/backend/models"
	"github.com/simu/drum-trainer/backend/storage"
)

type sharingEnv struct {
	t       *testing.T
	router  http.Handler
	trackID int64
}

// newSharingEnv starts a router on a fresh database with two signed-in users
// ("owner" and "other") and one 8-bar track owned by "owner".
func newSharingEnv(t *testing.T) *sharingEnv {
	t.Helper()
	dir := t.TempDir()
	cfg := config.Config{DBPath: filepath.Join(dir, "test.db"), SessionSecret: "test"}
	db, err := dbpkg.Open(cfg)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	store, err := storage.NewStore(filepath.Join(dir, "uploads"))
	if err != nil {
		t.Fatal(err)
	}

	var trackID int64
	for _, name := range []string{"owner", "other"} {
		u, err := dbpkg.UpsertUser(db, "github", name, name+"@example.com", name, "")
		if err != nil {
			t.Fatal(err)
		}
		if err := dbpkg.CreateSession(db, name, u.ID); err != nil {
			t.Fatal(err)
		}
		if name != "owner" {
			continue
		}
		// 120 BPM 4/4: a bar is 2000 ms, so 8 bars of one note per beat.
		var notes []models.NoteEvent
		for beat := range 32 {
			notes = append(notes, models.NoteEvent{NoteNumber: 36, TimeMs: float64(beat) * 500, Velocity: 100, Instrument: "kick"})
		}
		track := &models.Track{Name: "groove", Filename: "groove.mid", FilePath: "x.mid", FileSize: 1,
			DurationMs: 16000, NoteCount: len(notes), PPQ: 480, BPM: 120, TimeSigNumerator: 4}
		if err := dbpkg.InsertTrack(db, track, notes, u.ID); err != nil {
			t.Fatal(err)
		}
		trackID = track.ID
	}
	return &sharingEnv{t: t, router: NewRouter(db, store, 1<<20, cfg, 1<<20), trackID: trackID}
}

// do sends a request as the given user ("" for anonymous) and decodes a JSON
// response into out when out is non-nil.
func (e *sharingEnv) do(method, path, user string, body any, out any) int {
	e.t.Helper()
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	if user != "" {
		req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: user})
	}
	rec := httptest.NewRecorder()
	e.router.ServeHTTP(rec, req)
	if out != nil && rec.Code < 300 {
		if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
			e.t.Fatalf("%s %s: decode: %v (%s)", method, path, err, rec.Body.String())
		}
	}
	return rec.Code
}

func TestShareRequiresAttestation(t *testing.T) {
	e := newSharingEnv(t)
	if code := e.do("PUT", "/api/v1/tracks/"+itoa(e.trackID)+"/sharing", "owner", map[string]any{"rights_basis": "original"}, nil); code != http.StatusBadRequest {
		t.Fatalf("without accept_terms: got %d, want 400", code)
	}
	if code := e.do("PUT", "/api/v1/tracks/"+itoa(e.trackID)+"/sharing", "owner", map[string]any{"rights_basis": "stolen", "accept_terms": true}, nil); code != http.StatusBadRequest {
		t.Fatalf("bad rights_basis: got %d, want 400", code)
	}
	if code := e.do("PUT", "/api/v1/tracks/"+itoa(e.trackID)+"/sharing", "other", map[string]any{"rights_basis": "original", "accept_terms": true}, nil); code != http.StatusNotFound {
		t.Fatalf("non-owner: got %d, want 404", code)
	}
}

func TestSharedTrackFlow(t *testing.T) {
	e := newSharingEnv(t)

	var shared struct {
		ShareToken string `json:"share_token"`
	}
	if code := e.do("PUT", "/api/v1/tracks/"+itoa(e.trackID)+"/sharing", "owner", map[string]any{"rights_basis": "original", "accept_terms": true}, &shared); code != http.StatusOK {
		t.Fatalf("share: got %d", code)
	}
	path := "/api/v1/shared/" + shared.ShareToken

	// Anonymous viewers see metadata and only the opening bars.
	var view models.SharedTrack
	if code := e.do("GET", path, "", nil, &view); code != http.StatusOK {
		t.Fatalf("anonymous view: got %d", code)
	}
	if view.NoteCount != 32 || len(view.PreviewNotes) != 16 || view.Instruments["kick"] != 32 {
		t.Fatalf("unexpected view: notes=%d preview=%d instruments=%v", view.NoteCount, len(view.PreviewNotes), view.Instruments)
	}
	if view.IsOwner || view.CopiedTrackID != nil {
		t.Fatal("anonymous viewer should not be owner or have a copy")
	}

	// Copying needs a session.
	if code := e.do("POST", path+"/copy", "", nil, nil); code != http.StatusUnauthorized {
		t.Fatalf("anonymous copy: got %d, want 401", code)
	}
	var copied struct {
		TrackID int64 `json:"track_id"`
	}
	if code := e.do("POST", path+"/copy", "other", nil, &copied); code != http.StatusCreated {
		t.Fatalf("copy: got %d", code)
	}
	var full models.TrackWithNotes
	if code := e.do("GET", "/api/v1/tracks/"+itoa(copied.TrackID), "other", nil, &full); code != http.StatusOK {
		t.Fatalf("get copy: got %d", code)
	}
	if len(full.Notes) != 32 || full.SourceTrackID == nil {
		t.Fatalf("copy incomplete: %d notes, source %v", len(full.Notes), full.SourceTrackID)
	}
	e.do("GET", path, "other", nil, &view)
	if view.CopiedTrackID == nil || *view.CopiedTrackID != copied.TrackID {
		t.Fatalf("copied_track_id = %v, want %d", view.CopiedTrackID, copied.TrackID)
	}

	// The copy cannot be shared onwards.
	if code := e.do("PUT", "/api/v1/tracks/"+itoa(copied.TrackID)+"/sharing", "other", map[string]any{"rights_basis": "original", "accept_terms": true}, nil); code != http.StatusConflict {
		t.Fatalf("re-share copy: got %d, want 409", code)
	}

	// Unsharing kills the link with a 404, indistinguishable from a bad token.
	if code := e.do("DELETE", "/api/v1/tracks/"+itoa(e.trackID)+"/sharing", "owner", nil, nil); code != http.StatusNoContent {
		t.Fatalf("unshare: got %d", code)
	}
	if code := e.do("GET", path, "", nil, nil); code != http.StatusNotFound {
		t.Fatalf("after unshare: got %d, want 404", code)
	}
	if code := e.do("GET", "/api/v1/shared/nonsense", "", nil, nil); code != http.StatusNotFound {
		t.Fatalf("unknown token: got %d, want 404", code)
	}
}

func TestPreviewMs(t *testing.T) {
	if got := previewMs(&models.Track{BPM: 120, TimeSigNumerator: 4, DurationMs: 60000}); got != 8000 {
		t.Errorf("4 bars of 4/4 at 120: got %v, want 8000", got)
	}
	if got := previewMs(&models.Track{BPM: 120, TimeSigNumerator: 4, DurationMs: 3000}); got != 3000 {
		t.Errorf("short track: got %v, want 3000", got)
	}
}

func itoa(n int64) string { return strconv.FormatInt(n, 10) }

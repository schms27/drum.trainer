package api

import (
	"net/http"
	"testing"

	"github.com/simu/drum-trainer/backend/models"
)

func TestReplaceNotes(t *testing.T) {
	e := newSharingEnv(t)
	path := "/api/v1/tracks/" + itoa(e.trackID) + "/notes"
	take := map[string]any{"track_id": e.trackID, "started_at": "2026-10-17T10:00:00Z", "ended_at": "2026-10-17T10:01:00Z", "hits": []any{}}

	var before models.SessionWithScore
	e.do("POST", "/api/v1/sessions", "owner", take, &before)
	if r := before.Session.NotesRevision; r == nil || *r != 0 {
		t.Fatalf("take before the edit: revision %v, want 0", r)
	}

	edited := map[string]any{"notes": []any{
		map[string]any{"note_number": 38, "time_ms": 500, "velocity": 90},
		map[string]any{"note_number": 36, "time_ms": 0, "velocity": 100},
		map[string]any{"note_number": 30, "time_ms": 20000, "velocity": 80}, // past the 16 s track
	}}
	var got models.TrackWithNotes
	if code := e.do("PUT", path, "owner", edited, &got); code != http.StatusOK {
		t.Fatalf("replace notes: got %d", code)
	}
	if got.NoteCount != 3 || got.NotesRevision != 1 || got.DurationMs != 20000 {
		t.Fatalf("track after edit: count=%d revision=%d duration=%v", got.NoteCount, got.NotesRevision, got.DurationMs)
	}
	if got.Notes[0].NoteNumber != 36 || got.Notes[1].Instrument == "" {
		t.Fatalf("notes not sorted or instruments missing: %+v", got.Notes)
	}

	var after models.SessionWithScore
	e.do("POST", "/api/v1/sessions", "owner", take, &after)
	if r := after.Session.NotesRevision; r == nil || *r != 1 {
		t.Fatalf("take after the edit: revision %v, want 1", r)
	}

	bad := []map[string]any{
		{"notes": []any{map[string]any{"note_number": 200, "time_ms": 0, "velocity": 100}}},
		{"notes": []any{map[string]any{"note_number": 36, "time_ms": -5, "velocity": 100}}},
		{"notes": []any{map[string]any{"note_number": 36, "time_ms": 0, "velocity": 0}}},
		{"notes": []any{}, "duration_ms": 0},
	}
	for _, b := range bad {
		if code := e.do("PUT", path, "owner", b, nil); code != http.StatusBadRequest {
			t.Errorf("invalid %v: got %d, want 400", b, code)
		}
	}
	if code := e.do("PUT", path, "other", edited, nil); code != http.StatusNotFound {
		t.Fatalf("someone else's track: got %d, want 404", code)
	}
}

func TestReplaceNotesOfSharedAndRemovedTracks(t *testing.T) {
	e := newSharingEnv(t)
	path := "/api/v1/tracks/" + itoa(e.trackID) + "/notes"
	edited := map[string]any{"notes": []any{map[string]any{"note_number": 36, "time_ms": 0, "velocity": 100}}}

	e.do("PUT", "/api/v1/tracks/"+itoa(e.trackID)+"/sharing", "owner", map[string]any{"rights_basis": "original", "accept_terms": true}, nil)
	if code := e.do("PUT", path, "owner", edited, nil); code != http.StatusOK {
		t.Fatalf("edit shared track: got %d", code)
	}
	var n int
	if err := e.db.QueryRow(`SELECT COUNT(*) FROM track_share_events WHERE track_id = ? AND action = 'edit_notes'`, e.trackID).Scan(&n); err != nil || n != 1 {
		t.Fatalf("edit of a shared track should be logged once: n=%d err=%v", n, err)
	}

	e.do("POST", "/api/v1/admin/tracks/"+itoa(e.trackID)+"/disable", "admin", map[string]any{"reason": "notice"}, nil)
	if code := e.do("PUT", path, "owner", edited, nil); code != http.StatusConflict {
		t.Fatalf("edit removed track: got %d, want 409", code)
	}
}

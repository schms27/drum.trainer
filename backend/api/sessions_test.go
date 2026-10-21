package api

import (
	"net/http"
	"testing"

	"github.com/simu/drum-trainer/backend/models"
)

func TestSessionTempo(t *testing.T) {
	e := newSharingEnv(t)
	take := func(bpm any) map[string]any {
		body := map[string]any{"track_id": e.trackID, "started_at": "2026-10-17T10:00:00Z", "ended_at": "2026-10-17T10:01:00Z", "hits": []any{}}
		if bpm != nil {
			body["practice_bpm"] = bpm
		}
		return body
	}

	// The track is 120 BPM, so 96 BPM is 80 %.
	var slow models.SessionWithScore
	if code := e.do("POST", "/api/v1/sessions", "owner", take(96), &slow); code != http.StatusCreated {
		t.Fatalf("submit at 96 BPM: got %d", code)
	}
	if p := slow.Session.TempoPercent; p == nil || *p != 80 || *slow.Session.TrackBPM != 120 {
		t.Fatalf("tempo of slow take: %+v", slow.Session.SessionTempo)
	}
	var unknown models.SessionWithScore
	if code := e.do("POST", "/api/v1/sessions", "owner", take(nil), &unknown); code != http.StatusCreated {
		t.Fatalf("submit without tempo: got %d", code)
	}
	if unknown.Session.TempoPercent != nil || unknown.Session.PracticeBPM != nil {
		t.Fatalf("take without tempo should have none: %+v", unknown.Session.SessionTempo)
	}
	if code := e.do("POST", "/api/v1/sessions", "owner", take(120), nil); code != http.StatusCreated {
		t.Fatalf("submit at 120 BPM: got %d", code)
	}
	if code := e.do("POST", "/api/v1/sessions", "owner", take(400), nil); code != http.StatusBadRequest {
		t.Fatalf("out of range tempo: got %d, want 400", code)
	}

	var got models.SessionWithScore
	e.do("GET", "/api/v1/sessions/"+itoa(slow.Session.ID), "owner", nil, &got)
	if p := got.Session.TempoPercent; p == nil || *p != 80 {
		t.Fatalf("GET session tempo: %+v", got.Session.SessionTempo)
	}

	list := func(query string) []models.SessionSummary {
		t.Helper()
		var page struct {
			Sessions []models.SessionSummary `json:"sessions"`
		}
		if code := e.do("GET", "/api/v1/sessions?track_id="+itoa(e.trackID)+query, "owner", nil, &page); code != http.StatusOK {
			t.Fatalf("list %q: got %d", query, code)
		}
		return page.Sessions
	}
	if n := len(list("")); n != 3 {
		t.Fatalf("all takes: got %d, want 3", n)
	}
	if s := list("&tempo_max=99"); len(s) != 1 || *s[0].TempoPercent != 80 {
		t.Fatalf("slower takes: %+v", s)
	}
	if s := list("&tempo_min=100&tempo_max=100"); len(s) != 1 || *s[0].TempoPercent != 100 {
		t.Fatalf("original tempo takes: %+v", s)
	}
	if s := list("&tempo_min=101"); len(s) != 0 {
		t.Fatalf("faster takes: %+v", s)
	}
	if code := e.do("GET", "/api/v1/sessions?tempo_min=fast", "owner", nil, nil); code != http.StatusBadRequest {
		t.Fatalf("bad tempo_min: got %d, want 400", code)
	}
}

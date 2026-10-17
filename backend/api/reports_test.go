package api

import (
	"net/http"
	"testing"

	dbpkg "github.com/simu/drum-trainer/backend/db"
	"github.com/simu/drum-trainer/backend/models"
)

// shareFixtureTrack shares the env's track and returns its token.
func shareFixtureTrack(t *testing.T, e *sharingEnv) string {
	t.Helper()
	var shared struct {
		ShareToken string `json:"share_token"`
	}
	if code := e.do("PUT", "/api/v1/tracks/"+itoa(e.trackID)+"/sharing", "owner", map[string]any{"rights_basis": "transcription", "accept_terms": true}, &shared); code != http.StatusOK {
		t.Fatalf("share: got %d", code)
	}
	return shared.ShareToken
}

func validReport(token string) map[string]any {
	return map[string]any{
		"token":           "http://localhost:5173/s/" + token,
		"kind":            "copyright",
		"reporter_name":   "Rights Holder",
		"reporter_email":  "rights@example.com",
		"work_identified": "Song X by Band Y",
		"details":         "This is a note-for-note copy of our drum part.",
		"good_faith":      true,
	}
}

func TestReportValidation(t *testing.T) {
	e := newSharingEnv(t)
	token := shareFixtureTrack(t, e)

	for name, mutate := range map[string]func(map[string]any){
		"bad kind":      func(b map[string]any) { b["kind"] = "spam" },
		"no name":       func(b map[string]any) { b["reporter_name"] = " " },
		"bad email":     func(b map[string]any) { b["reporter_email"] = "nope" },
		"short details": func(b map[string]any) { b["details"] = "bad" },
		"no work":       func(b map[string]any) { b["work_identified"] = "" },
		"no good faith": func(b map[string]any) { b["good_faith"] = false },
		"unknown token": func(b map[string]any) { b["token"] = "missing" },
	} {
		e.ip = "198.51.100." + itoa(int64(len(name))) // stay under the per-IP report limit
		body := validReport(token)
		mutate(body)
		want := http.StatusBadRequest
		if name == "unknown token" {
			want = http.StatusNotFound
		}
		if code := e.do("POST", "/api/v1/reports", "", body, nil); code != want {
			t.Errorf("%s: got %d, want %d", name, code, want)
		}
	}

	// Abuse reports don't need a work or the good-faith statement.
	abuse := validReport(token)
	abuse["kind"], abuse["work_identified"], abuse["good_faith"] = "abuse", "", false
	e.ip = "203.0.113.1"
	if code := e.do("POST", "/api/v1/reports", "", abuse, nil); code != http.StatusCreated {
		t.Errorf("abuse report: got %d, want 201", code)
	}
}

func TestAdminRoutesHiddenFromUsers(t *testing.T) {
	e := newSharingEnv(t)
	for _, user := range []string{"", "owner"} {
		code := e.do("GET", "/api/v1/admin/reports", user, nil, nil)
		if (user == "" && code != http.StatusUnauthorized) || (user != "" && code != http.StatusNotFound) {
			t.Errorf("user %q: got %d", user, code)
		}
	}
	var me models.User
	e.do("GET", "/api/v1/auth/me", "admin", nil, &me)
	if !me.IsAdmin {
		t.Error("/auth/me should flag the admin")
	}
	e.do("GET", "/api/v1/auth/me", "owner", nil, &me)
	if me.IsAdmin {
		t.Error("/auth/me flags a non-admin")
	}
}

func TestTakedownFlow(t *testing.T) {
	e := newSharingEnv(t)
	token := shareFixtureTrack(t, e)

	var copied struct {
		TrackID int64 `json:"track_id"`
	}
	e.do("POST", "/api/v1/shared/"+token+"/copy", "other", nil, &copied)

	if code := e.do("POST", "/api/v1/reports", "", validReport(token), nil); code != http.StatusCreated {
		t.Fatalf("report: got %d", code)
	}
	var queue struct {
		Reports []models.ReportWithContext `json:"reports"`
	}
	e.do("GET", "/api/v1/admin/reports", "admin", nil, &queue)
	if len(queue.Reports) != 1 || queue.Reports[0].TrackID != e.trackID || queue.Reports[0].Copies != 1 ||
		queue.Reports[0].OwnerEmail == nil || *queue.Reports[0].OwnerEmail != "owner@example.com" {
		t.Fatalf("unexpected queue: %+v", queue.Reports)
	}

	if code := e.do("POST", "/api/v1/admin/tracks/"+itoa(e.trackID)+"/disable", "admin", map[string]any{"reason": ""}, nil); code != http.StatusBadRequest {
		t.Fatalf("disable without reason: got %d, want 400", code)
	}
	var res dbpkg.ModerationResult
	if code := e.do("POST", "/api/v1/admin/tracks/"+itoa(e.trackID)+"/disable", "admin",
		map[string]any{"reason": "Removed after a copyright notice for Song X.", "strike": true}, &res); code != http.StatusOK {
		t.Fatalf("disable: got %d", code)
	}
	if res.TracksAffected != 2 || res.OwnerStrikes != 1 {
		t.Fatalf("unexpected result: %+v", res)
	}

	// The link is dead, the copy keeps its page but loses its notes and can't be practised.
	if code := e.do("GET", "/api/v1/shared/"+token, "", nil, nil); code != http.StatusNotFound {
		t.Fatalf("shared link after takedown: got %d", code)
	}
	var tr models.TrackWithNotes
	e.do("GET", "/api/v1/tracks/"+itoa(copied.TrackID), "other", nil, &tr)
	if !tr.Disabled || len(tr.Notes) != 0 || tr.DisabledReason == nil {
		t.Fatalf("copy not disabled: disabled=%v notes=%d", tr.Disabled, len(tr.Notes))
	}
	take := map[string]any{"track_id": copied.TrackID, "started_at": "2026-10-17T10:00:00Z", "ended_at": "2026-10-17T10:01:00Z", "duration_ms": 1000, "hits": []any{}}
	if code := e.do("POST", "/api/v1/sessions", "other", take, nil); code != http.StatusConflict {
		t.Fatalf("practising a removed track: got %d, want 409", code)
	}
	if code := e.do("PUT", "/api/v1/tracks/"+itoa(e.trackID)+"/sharing", "owner", map[string]any{"rights_basis": "original", "accept_terms": true}, nil); code != http.StatusConflict {
		t.Fatalf("re-sharing a removed track: got %d, want 409", code)
	}

	if code := e.do("POST", "/api/v1/admin/tracks/"+itoa(e.trackID)+"/restore", "admin", map[string]any{"note": "Counter-notice", "revoke_strike": true}, &res); code != http.StatusOK {
		t.Fatalf("restore: got %d", code)
	}
	e.do("GET", "/api/v1/tracks/"+itoa(copied.TrackID), "other", nil, &tr)
	if tr.Disabled || len(tr.Notes) == 0 {
		t.Fatal("copy not restored")
	}
}

func TestBannedUserCannotShare(t *testing.T) {
	e := newSharingEnv(t)
	shareFixtureTrack(t, e)
	for range models.StrikeLimit {
		// Each takedown needs the track enabled again first; strikes stay.
		e.do("POST", "/api/v1/admin/tracks/"+itoa(e.trackID)+"/disable", "admin", map[string]any{"reason": "notice", "strike": true}, nil)
		e.do("POST", "/api/v1/admin/tracks/"+itoa(e.trackID)+"/restore", "admin", map[string]any{}, nil)
	}
	var me models.User
	e.do("GET", "/api/v1/auth/me", "owner", nil, &me)
	if !me.SharingBanned {
		t.Fatal("owner should be banned after reaching the strike limit")
	}
	if code := e.do("PUT", "/api/v1/tracks/"+itoa(e.trackID)+"/sharing", "owner", map[string]any{"rights_basis": "original", "accept_terms": true}, nil); code != http.StatusForbidden {
		t.Fatalf("banned share: got %d, want 403", code)
	}
}

func TestExtractToken(t *testing.T) {
	for in, want := range map[string]string{
		"abc":                             "abc",
		" https://x.example/s/abc ":       "abc",
		"https://x.example/s/abc?utm=1#x": "abc",
		"/s/abc/":                         "abc",
	} {
		if got := extractToken(in); got != want {
			t.Errorf("extractToken(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestReportRateLimit(t *testing.T) {
	e := newSharingEnv(t)
	token := shareFixtureTrack(t, e)
	e.ip = "192.0.2.50"
	for i := range 6 {
		code := e.do("POST", "/api/v1/reports", "", validReport(token), nil)
		if want := http.StatusCreated; i < 5 && code != want {
			t.Fatalf("report %d: got %d, want %d", i+1, code, want)
		}
		if i == 5 && code != http.StatusTooManyRequests {
			t.Fatalf("6th report in an hour: got %d, want 429", code)
		}
	}
}

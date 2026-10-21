package api

import (
	"net/http"
	"testing"

	"github.com/simu/drum-trainer/backend/models"
)

func TestTags(t *testing.T) {
	e := newSharingEnv(t)
	var warm, gig models.Tag
	if code := e.do("POST", "/api/v1/tags", "owner", map[string]any{"name": "  Warm   up ", "color": "teal"}, &warm); code != http.StatusCreated || warm.Name != "Warm up" {
		t.Fatalf("create tag: got %d %+v", code, warm)
	}
	e.do("POST", "/api/v1/tags", "owner", map[string]any{"name": "Gig set", "color": "pink"}, &gig)
	for _, bad := range []map[string]any{
		{"name": "warm UP", "color": "red"}, // duplicate, ignoring case → 409
		{"name": "x", "color": "#ff0000"},   // not a palette colour
		{"name": "", "color": "red"},        // empty
		{"name": string(make([]byte, 31)), "color": "red"},
	} {
		code := e.do("POST", "/api/v1/tags", "owner", bad, nil)
		if code != http.StatusConflict && code != http.StatusBadRequest {
			t.Errorf("invalid tag %v: got %d", bad, code)
		}
	}
	// Someone else may use the same name.
	if code := e.do("POST", "/api/v1/tags", "other", map[string]any{"name": "Warm up", "color": "red"}, nil); code != http.StatusCreated {
		t.Fatalf("other user's tag with the same name: got %d", code)
	}

	tagsPath := "/api/v1/tracks/" + itoa(e.trackID) + "/tags"
	var set struct {
		Tags []models.TagRef `json:"tags"`
	}
	if code := e.do("PUT", tagsPath, "owner", map[string]any{"tag_ids": []int64{warm.ID, gig.ID}}, &set); code != http.StatusOK || len(set.Tags) != 2 || set.Tags[0].Name != "Gig set" {
		t.Fatalf("set tags: got %d %+v", code, set.Tags)
	}
	var tr models.TrackWithNotes
	e.do("GET", "/api/v1/tracks/"+itoa(e.trackID), "owner", nil, &tr)
	if len(tr.Tags) != 2 {
		t.Fatalf("track tags: %+v", tr.Tags)
	}

	// Another user's tag or track is a 404, and nothing changes.
	var foreign models.Tag
	e.do("POST", "/api/v1/tags", "other", map[string]any{"name": "Mine", "color": "red"}, &foreign)
	if code := e.do("PUT", tagsPath, "owner", map[string]any{"tag_ids": []int64{foreign.ID}}, nil); code != http.StatusNotFound {
		t.Fatalf("foreign tag: got %d, want 404", code)
	}
	if code := e.do("PUT", tagsPath, "other", map[string]any{"tag_ids": []int64{}}, nil); code != http.StatusNotFound {
		t.Fatalf("foreign track: got %d, want 404", code)
	}

	// Rename and recolour, then the listing shows counts.
	if code := e.do("PATCH", "/api/v1/tags/"+itoa(gig.ID), "owner", map[string]any{"name": "Gig", "color": "violet"}, nil); code != http.StatusNoContent {
		t.Fatalf("update tag: got %d", code)
	}
	if code := e.do("PATCH", "/api/v1/tags/"+itoa(gig.ID), "owner", map[string]any{"name": "WARM UP"}, nil); code != http.StatusConflict {
		t.Fatalf("rename to an existing name: got %d, want 409", code)
	}
	var list struct {
		Tags   []models.Tag `json:"tags"`
		Colors []string     `json:"colors"`
	}
	e.do("GET", "/api/v1/tags", "owner", nil, &list)
	if len(list.Tags) != 2 || list.Tags[0].Name != "Gig" || list.Tags[0].Color != "violet" || list.Tags[0].TrackCount != 1 || len(list.Colors) != len(models.TagColors) {
		t.Fatalf("tag list: %+v", list)
	}

	// Copies of shared tracks don't get the sharer's tags.
	var shared struct {
		ShareToken string `json:"share_token"`
	}
	e.do("PUT", "/api/v1/tracks/"+itoa(e.trackID)+"/sharing", "owner", map[string]any{"rights_basis": "original", "accept_terms": true}, &shared)
	var copied struct {
		TrackID int64 `json:"track_id"`
	}
	e.do("POST", "/api/v1/shared/"+shared.ShareToken+"/copy", "other", nil, &copied)
	var copy models.TrackWithNotes
	e.do("GET", "/api/v1/tracks/"+itoa(copied.TrackID), "other", nil, &copy)
	if len(copy.Tags) != 0 {
		t.Fatalf("copy should have no tags: %+v", copy.Tags)
	}

	// Deleting a tag removes it from tracks.
	if code := e.do("DELETE", "/api/v1/tags/"+itoa(warm.ID), "owner", nil, nil); code != http.StatusNoContent {
		t.Fatalf("delete tag: got %d", code)
	}
	tr = models.TrackWithNotes{}
	e.do("GET", "/api/v1/tracks/"+itoa(e.trackID), "owner", nil, &tr)
	if len(tr.Tags) != 1 || tr.Tags[0].Name != "Gig" {
		t.Fatalf("after delete: %+v", tr.Tags)
	}
}

func TestRenameTrack(t *testing.T) {
	e := newSharingEnv(t)
	path := "/api/v1/tracks/" + itoa(e.trackID)
	if code := e.do("PATCH", path, "owner", map[string]any{"name": "  My groove  "}, nil); code != http.StatusOK {
		t.Fatalf("rename: got %d", code)
	}
	var tr models.TrackWithNotes
	e.do("GET", path, "owner", nil, &tr)
	if tr.Name != "My groove" {
		t.Fatalf("name = %q", tr.Name)
	}
	if code := e.do("PATCH", path, "owner", map[string]any{"name": "   "}, nil); code != http.StatusBadRequest {
		t.Fatalf("empty name: got %d, want 400", code)
	}
	if code := e.do("PATCH", path, "other", map[string]any{"name": "Stolen"}, nil); code != http.StatusNotFound {
		t.Fatalf("someone else's track: got %d, want 404", code)
	}

	// Renaming a shared track is logged and shows on the share link.
	var shared struct {
		ShareToken string `json:"share_token"`
	}
	e.do("PUT", path+"/sharing", "owner", map[string]any{"rights_basis": "original", "accept_terms": true}, &shared)
	e.do("PATCH", path, "owner", map[string]any{"name": "Shared groove"}, nil)
	var view models.SharedTrack
	e.do("GET", "/api/v1/shared/"+shared.ShareToken, "", nil, &view)
	if view.Name != "Shared groove" {
		t.Fatalf("share link name = %q", view.Name)
	}
	var n int
	if err := e.db.QueryRow(`SELECT COUNT(*) FROM track_share_events WHERE track_id = ? AND action = 'rename'`, e.trackID).Scan(&n); err != nil || n != 1 {
		t.Fatalf("rename events: n=%d err=%v", n, err)
	}
}
